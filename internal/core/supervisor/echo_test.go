//go:build unix

package supervisor

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// echoLog 收集异步插件输出与 Core 日志，读取时与写入互斥。
type echoLog struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (l *echoLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buffer.Write(p)
}
func (l *echoLog) String() string { l.mu.Lock(); defer l.mu.Unlock(); return l.buffer.String() }

// TestEchoBothTransports 验证真实注册和激活后的原目标回复，以及无匹配/无文本事件不发送。
func TestEchoBothTransports(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "echo")
	if out, err := exec.Command("go", "build", "-o", binary, "../../../examples/plugins/echo").CombinedOutput(); err != nil {
		t.Fatalf("build echo plugin: %v\n%s", err, out)
	}
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) {
			ctx := context.Background()
			logs := &echoLog{}
			output, err := logging.Open(logging.Options{Format: logging.FormatJSON, Level: slog.LevelDebug, Stderr: logs})
			if err != nil {
				t.Fatal(err)
			}
			setDefaultLogger(t, output.Logger())
			defer output.Close()
			defer func() {
				if t.Failed() {
					t.Logf("Core and plugin logs:\n%s", logs.String())
				}
			}()
			values := tune.Defaults()
			store, spec := setup(t, binary)
			spec.Stdio = stdio
			spec.Limits = host.NewLimits(values)
			spec.Log = output.Plugin(spec.PluginID, spec.GenerationID)
			registry := adapter.NewRegistry()
			platform := memory.New("memory", "instance", "account")
			if err := registry.Add(platform); err != nil {
				t.Fatal(err)
			}
			mux := router.NewMux()
			sender := message.New(registry, policy.New(store), nil, values.RequestTimeout)
			mux.Handle(wire.ServiceSend, sender.Send)
			caller := router.NewCaller(spec.PluginID, spec.GenerationID)
			spec.Handler = mux.For(caller)
			events := router.New(values.DispatchQueueEvents, registry, sender.Send, nil, nil)
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
					t.Errorf("echo did not stop cleanly: %+v", stopped)
				}
			}()
			events.Attach(router.FromInstance(running.Instance), caller)
			target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat", ThreadId: proto.String("topic")}
			original := &pb.MessageChain{Segments: []*pb.Segment{
				{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "hello 世界"}}},
				{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "second fragment"}}},
			}}
			event := func(chain *pb.MessageChain) *pb.Event {
				return &pb.Event{EventType: "message.created", Target: target, SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER,
					EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: chain}}}}
			}
			if err := platform.Inject(ctx, event(original)); err != nil {
				t.Fatal(err)
			}
			if err := platform.Inject(ctx, &pb.Event{EventType: "member.joined", Target: target}); err != nil {
				t.Fatal(err)
			}
			if err := platform.Inject(ctx, event(&pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Mention{Mention: &pb.MentionSegment{ExternalUserId: "user"}}}}})); err != nil {
				t.Fatal(err)
			}
			marker := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "queue barrier"}}}}}
			if err := platform.Inject(ctx, event(marker)); err != nil {
				t.Fatal(err)
			}
			// 单 worker 顺序派发：marker 的发送证明前面的无匹配、无文本事件均已处理。
			deadline := time.Now().Add(5 * time.Second)
			for {
				sent := platform.Sent()
				if len(sent) > 0 && proto.Equal(sent[len(sent)-1].Chain, marker) {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("echo barrier not sent; recorded sends: %v", sent)
				}
				time.Sleep(10 * time.Millisecond)
			}
			sent := platform.Sent()
			if len(sent) != 2 {
				t.Fatalf("send count = %d, want 2 (unmatched and non-text events must not send)", len(sent))
			}
			if !proto.Equal(sent[0].Chain, original) || !proto.Equal(sent[0].Target, target) || !proto.Equal(sent[1].Target, target) {
				t.Fatalf("echo changed text or target: %v", sent)
			}
		})
	}
}
