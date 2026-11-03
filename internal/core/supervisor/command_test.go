//go:build unix

package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

type commandAdapter struct{ *memory.Adapter }

func (*commandAdapter) AccountName() string { return "bot" }

func commandAudit(t *testing.T, path string) []string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var entries []string
	for {
		var entry string
		err := decoder.Decode(&entry)
		if err == io.EOF {
			return entries
		}
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
}

// TestCommandDispatchBothTransports 验证真实 SDK、Session、Router 与发送服务的命令链。
func TestCommandDispatchBothTransports(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "commanddispatch")
	if out, err := exec.Command("go", "build", "-o", binary, "./testdata/commanddispatch").CombinedOutput(); err != nil {
		t.Fatalf("build command dispatch plugin: %v\n%s", err, out)
	}
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) {
			ctx := context.Background()
			logs := &echoLog{}
			output, err := logging.Open(logging.Options{Format: logging.FormatJSON, Level: slog.LevelDebug, Stderr: logs})
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			defer func() {
				if t.Failed() {
					t.Logf("Core and plugin logs:\n%s", logs.String())
				}
			}()
			store, spec := setup(t, binary)
			values := tune.Defaults()
			spec.Stdio = stdio
			spec.Limits = host.NewLimits(values)
			spec.Log = output.Plugin(spec.PluginID, spec.GenerationID)
			audit := filepath.Join(t.TempDir(), "audit.jsonl")
			if err := os.WriteFile(audit, nil, 0600); err != nil {
				t.Fatal(err)
			}
			spec.ExtraEnv = map[string]string{"ALONABOT_TEST_COMMAND_AUDIT": audit}
			registry := adapter.NewRegistry()
			platform := &commandAdapter{memory.New("memory", "instance", "account")}
			if err := registry.Add(platform); err != nil {
				t.Fatal(err)
			}
			sender := message.New(registry, policy.New(store), nil, values.RequestTimeout)
			target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat", ThreadId: proto.String("topic")}
			actor := &pb.Actor{ExternalUserId: "user", AdapterInstanceId: "instance"}
			mux := router.NewMux()
			mux.Handle(wire.ServiceSend, sender.Send)
			caller := router.NewCaller(spec.PluginID, spec.GenerationID)
			spec.Handler = mux.For(caller)
			events := router.New(output.Logger(), values.DispatchQueueEvents, registry, sender.Send, nil, nil)
			services := broker.New(events, output.Logger(), values.RequestTimeout)
			serviceParents := make(chan uint64, 1)
			mux.Forward(func(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
				parent, ok := in.ParentAttach.(*router.Invocation)
				if in.ParentCallID == 0 || !ok || !proto.Equal(parent.Event.GetActor(), actor) || !proto.Equal(parent.DefaultTarget, target) {
					t.Errorf("service lost command parent: call_id=%d attach=%+v", in.ParentCallID, in.ParentAttach)
				}
				serviceParents <- in.ParentCallID
				return services.Call(ctx, caller, in)
			})
			events.Start()
			defer events.Close()
			if err := platform.Start(ctx, events); err != nil {
				t.Fatal(err)
			}
			defer platform.Stop(ctx)
			running, err := Launch(ctx, store, spec)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				stopped := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				if stopped.Err != nil || stopped.Killed {
					t.Errorf("plugin did not stop cleanly: %+v", stopped)
				}
			}()
			events.Attach(router.FromInstance(running.Instance), caller)
			inject := func(parts ...string) {
				t.Helper()
				chain := &pb.MessageChain{}
				for _, part := range parts {
					chain.Segments = append(chain.Segments, &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: part}}})
				}
				if err := platform.Inject(ctx, &pb.Event{EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, Target: target, Actor: actor, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: chain}}}}); err != nil {
					t.Fatal(err)
				}
			}
			for _, tc := range []struct {
				name    string
				parts   []string
				calls   []string
				replies []string
			}{
				{"longest", []string{"/rss add x"}, []string{"add:x", "event"}, []string{"added:x"}},
				{"list", []string{"/rss list"}, []string{"list", "event"}, []string{"list"}},
				{"alias", []string{"/feed"}, []string{"rss", "event"}, []string{"rss"}},
				{"full_token", []string{"/rssx"}, []string{"raw", "event"}, nil},
				{"username", []string{"/rss@BoT"}, []string{"rss", "event"}, []string{"rss"}},
				{"wrong_username", []string{"/rss@other"}, []string{"raw", "event"}, nil},
				{"sum", []string{"/calc 1 2"}, []string{"calc", "event"}, []string{"3"}},
				{"service", []string{"/service"}, []string{"service", "event"}, []string{"service reply", "demo"}},
				{"bad_number", []string{"/calc 1 x"}, []string{"event"}, []string{"Usage: /calc <a:int> <b:int>"}},
				{"missing", []string{"/calc 1"}, []string{"event"}, []string{"Usage: /calc <a:int> <b:int>"}},
				{"extra", []string{"/calc 1 2 3"}, []string{"event"}, []string{"Usage: /calc <a:int> <b:int>"}},
				{"silent", []string{"/quiet x"}, []string{"event"}, nil},
				{"raw_tail", []string{`/say  "a b"  c`}, []string{`say:"a b"  c`, "event"}, []string{`"a b"  c`}},
				{"single_literal", []string{`/quote 'a\b' c`}, []string{"quote", "event"}, []string{`a\b|c`}},
				{"double_escape", []string{`/quote "a\"b" c\ d`}, []string{"quote", "event"}, []string{`a"b|c d`}},
				{"empty_string", []string{`/quote "" c`}, []string{"quote", "event"}, []string{"|c"}},
				{"unclosed", []string{`/quote "a b c`}, []string{"event"}, []string{"Usage: /quote <a> <b>"}},
				{"trailing_escape", []string{`/quote a b\`}, []string{"event"}, []string{"Usage: /quote <a> <b>"}},
				{"default", []string{"/optional"}, []string{"optional", "event"}, []string{"9"}},
				{"custom_usage", []string{"/custom x"}, []string{"event"}, []string{"custom usage"}},
				{"auto_send_guard", []string{"/sendguard"}, []string{"sendguard", "event"}, nil},
				{"consume_text_matches", []string{"/say hello"}, []string{"say:hello", "event"}, []string{"hello"}},
				{"mixed_text_order", []string{"hel", "lo"}, []string{"regex:hello|hello", "full", "keyword", "raw", "event"}, nil},
				{"stop", []string{"hello stop"}, []string{"regex:hello|hello", "keyword"}, nil},
				{"ignore_case", []string{"HELLO"}, []string{"keyword", "raw", "event"}, nil},
				{"starts", []string{"BEGIN text"}, []string{"starts", "raw", "event"}, nil},
				{"ends", []string{"text END"}, []string{"ends", "raw", "event"}, nil},
				{"full", []string{"WHOLE"}, []string{"full", "raw", "event"}, nil},
			} {
				t.Run(tc.name, func(t *testing.T) {
					beforeAudit := len(commandAudit(t, audit))
					beforeSend := len(platform.Sent())
					inject(tc.parts...)
					inject("/barrier")
					// barrier 的 STOP 消除后继调用；其回执证明单 worker 已完成前面消息的完整派发。
					deadline := time.Now().Add(5 * time.Second)
					for {
						sent := platform.Sent()
						if len(sent) > beforeSend && sent[len(sent)-1].Chain.GetSegments()[0].GetText().GetText() == "barrier" {
							break
						}
						if time.Now().After(deadline) {
							t.Fatalf("barrier not reached: %v", sent)
						}
						time.Sleep(10 * time.Millisecond)
					}
					gotCalls := commandAudit(t, audit)[beforeAudit:]
					if !reflect.DeepEqual(gotCalls, tc.calls) {
						t.Fatalf("calls=%v, want %v", gotCalls, tc.calls)
					}
					sent := platform.Sent()[beforeSend:]
					if len(sent) != len(tc.replies)+1 {
						t.Fatalf("send count=%d, want %d: %v", len(sent), len(tc.replies)+1, sent)
					}
					for i, reply := range tc.replies {
						if got := sent[i].Chain.GetSegments()[0].GetText().GetText(); got != reply {
							t.Errorf("reply %d=%q, want %q", i, got, reply)
						}
					}
					for _, send := range sent {
						if !proto.Equal(send.Target, target) {
							t.Errorf("target changed: %v", send.Target)
						}
					}
					if tc.name == "service" {
						select {
						case parent := <-serviceParents:
							if parent == 0 {
								t.Error("service request has no parent_call_id")
							}
						default:
							t.Error("command did not call Service Broker")
						}
					}
					if tc.name == "auto_send_guard" && !strings.Contains(logs.String(), "ERROR_CODE_FAILED_PRECONDITION") {
						t.Error("missing auto-reply Send rejection diagnostic")
					}
				})
			}
		})
	}
}
