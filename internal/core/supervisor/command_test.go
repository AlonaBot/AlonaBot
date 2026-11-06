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
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type commandAdapter struct{ *incomingMediaAdapter }

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
			setDefaultLogger(t, output.Logger())
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
			mediaStore, err := media.Open(t.TempDir(), media.NewLimits(values), time.Now)
			if err != nil {
				t.Fatal(err)
			}
			registry := adapter.NewRegistry()
			mediaContent := []byte("command media 世界")
			platform := &commandAdapter{&incomingMediaAdapter{Adapter: memory.New("memory", "instance", "account"), data: mediaContent}}
			if err := registry.Add(platform); err != nil {
				t.Fatal(err)
			}
			customPlatform := memory.New("memory", "custom", "custom-account")
			if err := registry.Add(customPlatform); err != nil {
				t.Fatal(err)
			}
			sender := message.New(registry, policy.New(store), mediaStore.LeaseForSend, values.RequestTimeout)
			target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat", ThreadId: proto.String("topic")}
			actor := &pb.Actor{ExternalUserId: "user", AdapterInstanceId: "instance"}
			mux := router.NewMux()
			mux.Handle(wire.ServiceSend, sender.Send)
			caller := router.NewCaller(spec.PluginID, spec.GenerationID)
			defer mediaStore.ReleaseSession(caller.SessionKey)
			spec.Handler = mux.For(caller)
			events := router.New(values.DispatchQueueEvents, registry, sender.Send, mediaStore.PrepareIncoming, func(instanceID string) []string {
				if instanceID == "custom" {
					return []string{"!", "!!", ""}
				}
				return []string{"/"}
			}, nil)
			mediaStore.AddReadServices(mux, registry, events, store.AdapterInstanceAllowed)
			services := broker.New(events, values.RequestTimeout)
			serviceParents := make(chan uint64, 1)
			mux.Forward(func(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
				parent, ok := in.ParentAttach.(*router.Invocation)
				expectedActor := proto.CloneOf(actor)
				expectedActor.Permissions = []string{"user"}
				if in.ParentCallID == 0 || !ok || !proto.Equal(parent.Event.GetActor(), expectedActor) || !proto.Equal(parent.DefaultTarget, target) {
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
			if err := customPlatform.Start(ctx, events); err != nil {
				t.Fatal(err)
			}
			defer customPlatform.Stop(ctx)
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
			uninstalledRoute := false
			for _, route := range running.Instance.Routes {
				var want []string
				if route.GetKind() == pb.RouteKind_ROUTE_KIND_COMMAND {
					want = []string{"user"}
				}
				if !reflect.DeepEqual(route.GetRequiredPermissions(), want) {
					t.Errorf("route %s permissions=%v, want %v", route.GetRouteId(), route.GetRequiredPermissions(), want)
				}
				// SDK 的按路由权限 API 尚待接入；这里为 Core 提供对应的静态描述。
				switch route.GetRouteId() {
				case "cmd.rss.guarded", "restricted.event", "restricted.text", "restricted.raw":
					route.RequiredPermissions = []string{"memory.admin"}
				case "allowed.text":
					route.RequiredPermissions = []string{"memory.support", "user"}
				case "uninstalled.poke":
					uninstalledRoute = reflect.DeepEqual(route.GetEvent().GetEventTypeFilter(), []string{"not-installed.poke"})
				}
			}
			if !uninstalledRoute {
				t.Fatal("uninstalled adapter event route was not registered")
			}
			events.Attach(router.FromInstance(running.Instance), caller)
			injectAs := func(as *pb.Actor, parts ...string) {
				t.Helper()
				chain := &pb.MessageChain{}
				for _, part := range parts {
					chain.Segments = append(chain.Segments, &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: part}}})
				}
				if err := platform.Inject(ctx, &pb.Event{EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, AdapterId: proto.String("memory"), AdapterInstanceId: proto.String("instance"), AccountId: proto.String("account"), Target: target, Actor: as, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: chain}}}}); err != nil {
					t.Fatal(err)
				}
			}
			inject := func(parts ...string) { injectAs(actor, parts...) }
			permissionActors := map[string]*pb.Actor{
				"permissions_denied":     actor,
				"permissions_allowed":    {ExternalUserId: "privileged", AdapterInstanceId: "instance", Permissions: []string{"memory.admin"}},
				"no_actor":               nil,
				"no_actor_text":          nil,
				"permissions_normalized": {ExternalUserId: "privileged", AdapterInstanceId: "instance", Permissions: []string{"user", "admin", "memory.admin", "chat_admin", "memory.admin"}},
			}
			permissionActors["text_raw_allowed"] = permissionActors["permissions_allowed"]
			inputEvents := map[string]*pb.Event{
				"event_permissions_denied":   {EventType: "member.joined", Actor: actor},
				"event_permissions_allowed":  {EventType: "member.joined", Actor: permissionActors["permissions_allowed"]},
				"event_permissions_no_actor": {EventType: "member.joined"},
			}
			customTarget := &pb.Target{AdapterId: "memory", AdapterInstanceId: "custom", AccountId: "custom-account", ChatId: "chat"}
			textSegment := func(text string) *pb.Segment {
				return &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}
			}
			commandEvent := func(to *pb.Target, segments ...*pb.Segment) *pb.Event {
				return &pb.Event{EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER,
					AdapterId: proto.String(to.AdapterId), AdapterInstanceId: proto.String(to.AdapterInstanceId), AccountId: proto.String(to.AccountId), Target: to,
					Actor:        &pb.Actor{ExternalUserId: "user", AdapterInstanceId: to.AdapterInstanceId},
					EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: segments}}}}}
			}
			inputEvents["adapter_event"] = &pb.Event{EventType: "memory.poke", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER,
				AdapterId: proto.String("memory"), AdapterInstanceId: proto.String("instance"), AccountId: proto.String("account"), Target: target, Actor: actor,
				EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_AdapterEvent{AdapterEvent: &pb.AdapterEventPayload{Version: 1, PayloadJson: `{"command":"/rss hello"}`}}}}
			for name, text := range map[string]string{"longest_prefix": "!!rss", "empty_prefix": "rss", "custom_usage_prefix": "!!calc x", "empty_usage_prefix": "calc x", "custom_slash_not_implicit": "/rss"} {
				inputEvents[name] = commandEvent(customTarget, textSegment(text))
			}
			inputEvents["instance_prefix_isolated"] = commandEvent(target, textSegment("!!rss"))
			inputEvents["leading_self"] = commandEvent(target,
				textSegment("  "), &pb.Segment{Data: &pb.Segment_Reference{Reference: &pb.ReferenceSegment{OriginalPlatformMessageId: "original"}}},
				&pb.Segment{Data: &pb.Segment_Mention{Mention: &pb.MentionSegment{ExternalUserId: "account"}}}, textSegment(" /rss"))
			inputEvents["leading_other"] = commandEvent(target, &pb.Segment{Data: &pb.Segment_Mention{Mention: &pb.MentionSegment{ExternalUserId: "other"}}}, textSegment(" /rss"))
			mentioned := &pb.Segment{Data: &pb.Segment_Mention{Mention: &pb.MentionSegment{ExternalUserId: "external-7"}}}
			inputEvents["user_mention"] = commandEvent(target, textSegment("/kick "), mentioned)
			inputEvents["user_text_mention"] = commandEvent(target, &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "/kick @某 人", Entities: []*pb.TextEntity{{Type: pb.EntityType_ENTITY_TYPE_TEXT_MENTION, OffsetUtf16: 6, LengthUtf16: 4, ExternalUserId: proto.String("external-7")}}}}})
			inputEvents["leading_kick"] = commandEvent(target,
				&pb.Segment{Data: &pb.Segment_Reference{Reference: &pb.ReferenceSegment{OriginalPlatformMessageId: "original"}}},
				&pb.Segment{Data: &pb.Segment_Mention{Mention: &pb.MentionSegment{ExternalUserId: "account"}}}, textSegment(" /kick "), mentioned)
			inputEvents["int_mention"] = commandEvent(target, textSegment("/calc "), mentioned, textSegment(" 2"))
			objectID := mediaStore.RegisterIncoming(platform.InstanceID(), incomingMediaFileID, proto.Uint64(uint64(len(mediaContent))), proto.String("image/png"))
			image := &pb.Segment{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Alt: proto.String("photo"), Source: &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: &pb.IncomingMedia{Token: objectID, Size: proto.Uint64(uint64(len(mediaContent))), Mime: proto.String("image/png")}}}}}}
			inputEvents["image"] = commandEvent(target, textSegment("/picture "), image)
			body := &pb.MessageChain{Segments: []*pb.Segment{textSegment("body "), mentioned}}
			inputEvents["message"] = commandEvent(target, textSegment("/relay body "), mentioned)
			mediaBody := proto.CloneOf(body)
			mediaBody.Segments = append(mediaBody.Segments, image)
			inputEvents["message_media"] = commandEvent(target, textSegment("/relay body "), mentioned, image)
			encodedBody, err := protojson.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			encodedMediaBody, err := protojson.Marshal(mediaBody)
			if err != nil {
				t.Fatal(err)
			}
			for _, tc := range []struct {
				name    string
				parts   []string
				calls   []string
				replies []string
			}{
				{"adapter_event", nil, []string{`adapter.poke:1:{"command":"/rss hello"}`}, nil},
				{"text_raw_denied", []string{"protected"}, []string{"allowed.text", "raw", "event"}, nil},
				{"text_raw_allowed", []string{"protected"}, []string{"restricted.text", "allowed.text", "restricted.raw", "raw", "event"}, nil},
				{"user_mention", nil, []string{"kick:external-7", "event"}, []string{"external-7"}},
				{"user_text_mention", nil, []string{"kick:external-7", "event"}, []string{"external-7"}},
				{"leading_kick", nil, []string{"kick:external-7", "event"}, []string{"external-7"}},
				{"image", nil, []string{"picture", "event"}, []string{string(mediaContent)}},
				{"message", nil, []string{"relay", "event"}, []string{string(encodedBody)}},
				{"message_media", nil, []string{"relay", "event"}, []string{string(encodedMediaBody)}},
				{"int_mention", nil, []string{"event"}, []string{"Usage: /calc <a:int> <b:int>"}},
				{"missing_user", []string{"/kick"}, []string{"event"}, []string{"Usage: /kick <user:user>"}},
				{"missing_image", []string{"/picture"}, []string{"event"}, []string{"Usage: /picture <image:image>"}},
				{"missing_message", []string{"/relay"}, []string{"event"}, []string{"Usage: /relay <body:message>"}},
				{"longest_prefix", nil, []string{"rss", "event"}, []string{"rss"}},
				{"empty_prefix", nil, []string{"rss", "event"}, []string{"rss"}},
				{"custom_usage_prefix", nil, []string{"event"}, []string{"Usage: !!calc <a:int> <b:int>"}},
				{"empty_usage_prefix", nil, []string{"event"}, []string{"Usage: calc <a:int> <b:int>"}},
				{"custom_slash_not_implicit", nil, []string{"raw", "event"}, nil},
				{"instance_prefix_isolated", nil, []string{"raw", "event"}, nil},
				{"leading_self", nil, []string{"rss", "event"}, []string{"rss"}},
				{"leading_other", nil, []string{"raw", "event"}, nil},
				{"event_permissions_denied", nil, []string{"observer:user"}, nil},
				{"event_permissions_allowed", nil, []string{"restricted.event", "observer:memory.admin,user"}, nil},
				{"event_permissions_no_actor", nil, []string{"observer:"}, nil},
				{"permissions_denied", []string{"/rss guarded bad"}, []string{"event"}, nil},
				{"permissions_allowed", []string{"/rss guarded 7"}, []string{"guarded", "event"}, []string{"7"}},
				{"no_actor", []string{"/rss"}, []string{"event"}, nil},
				{"no_actor_text", []string{"hello"}, []string{"event"}, nil},
				{"permissions_normalized", []string{"/permissions"}, []string{"permissions", "event"}, []string{"chat_admin,memory.admin,user"}},
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
				{"custom_usage_key", []string{"/custom x"}, []string{"event"}, []string{"Usage: /custom <arg1:int>"}},
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
					beforeCustomSend := len(customPlatform.Sent())
					as, selected := permissionActors[tc.name]
					if !selected {
						as = actor
					}
					if event, selected := inputEvents[tc.name]; selected {
						selectedPlatform := platform.Adapter
						if event.GetAdapterInstanceId() == "custom" {
							selectedPlatform = customPlatform
						}
						if err := selectedPlatform.Inject(ctx, event); err != nil {
							t.Fatal(err)
						}
					} else {
						injectAs(as, tc.parts...)
					}
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
					sent := append(customPlatform.Sent()[beforeCustomSend:], platform.Sent()[beforeSend:]...)
					if len(sent) != len(tc.replies)+1 {
						t.Fatalf("send count=%d, want %d: %v", len(sent), len(tc.replies)+1, sent)
					}
					for i, reply := range tc.replies {
						got := sent[i].Chain.GetSegments()[0].GetText().GetText()
						if tc.name == "message" || tc.name == "message_media" {
							var received pb.MessageChain
							if err := protojson.Unmarshal([]byte(got), &received); err != nil {
								t.Fatal(err)
							}
							var want pb.MessageChain
							if err := protojson.Unmarshal([]byte(reply), &want); err != nil {
								t.Fatal(err)
							}
							if tc.name == "message_media" {
								if len(received.Segments) != 3 || received.Segments[2].GetImage() == nil {
									t.Fatalf("message media lost segments: %v", &received)
								}
								incoming := received.Segments[2].GetImage().GetSource().GetIncomingRef()
								if incoming.GetToken() == "" || incoming.GetToken() == objectID || incoming.GetReadability() != pb.MediaReadability_MEDIA_READABILITY_READABLE || incoming.GetExpiresAtUnixMs() == 0 || incoming.GetSize() != uint64(len(mediaContent)) || incoming.GetMime() != "image/png" {
									t.Fatalf("invalid prepared incoming_ref: %v", incoming)
								}
								// 随机令牌已在插件进程实际读取；只还原该字段以精确比较其余片段。
								received.Segments[2].GetImage().Source = proto.CloneOf(image.GetImage().Source)
							}
							if !proto.Equal(&received, &want) {
								t.Errorf("message=%v, want %v", &received, &want)
							}
							continue
						}
						if got != reply {
							t.Errorf("reply %d=%q, want %q", i, got, reply)
						}
					}
					for i, send := range sent {
						wantTarget := target
						if event := inputEvents[tc.name]; event != nil && event.GetTarget() != nil && i < len(sent)-1 {
							wantTarget = event.GetTarget()
						}
						if !proto.Equal(send.Target, wantTarget) {
							t.Errorf("target changed: %v", send.Target)
						}
					}
					if tc.name == "permissions_denied" && !strings.Contains(logs.String(), `"msg":"route permission denied"`) {
						t.Error("missing route permission rejection diagnostic")
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
			if got := platform.reads.Load(); got != 1 {
				t.Fatalf("platform reads=%d, want shared single fetch", got)
			}
			if image.GetImage().GetSource().GetIncomingRef().GetToken() != objectID {
				t.Fatal("dispatch mutated original incoming media object ID")
			}
			// 直接经真实 Session 检查错误的 ArgValue 分支；正常片段由上面的端到端场景覆盖。
			for _, tc := range []struct {
				name     string
				route    string
				arg      *pb.ArgValue
				want     string
				call     string
				rejected bool
			}{
				{"user", "kick", &pb.ArgValue{Name: "user", Value: &pb.ArgValue_UserId{UserId: "external-7"}}, "external-7", "kick:external-7", false},
				{"wrong_user_branch", "kick", &pb.ArgValue{Name: "user", Value: &pb.ArgValue_StringValue{StringValue: "external-7"}}, "", "", true},
			} {
				t.Run("sdk_binding/"+tc.name, func(t *testing.T) {
					before := len(commandAudit(t, audit))
					request := &pb.HandlerRequest{RouteId: "cmd." + tc.route, Command: &pb.CommandInvocation{Args: []*pb.ArgValue{tc.arg}}}
					encoded, err := proto.Marshal(request)
					if err != nil {
						t.Fatal(err)
					}
					response, err := running.Instance.Session.Call(ctx, wire.Request{Target: request.RouteId, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
					if tc.rejected {
						if wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT || response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT || len(commandAudit(t, audit)) != before {
							t.Fatalf("invalid argument reached handler: response=%+v", response)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					outcome := &pb.HandlerOutcome{}
					if response.Err != nil {
						t.Fatal(response.Err)
					}
					if err := proto.Unmarshal(response.Payload.Bytes, outcome); err != nil {
						t.Fatal(err)
					}
					if !outcome.GetMatched() || len(outcome.GetChain().GetSegments()) != 1 || outcome.Chain.Segments[0].GetText().GetText() != tc.want || !reflect.DeepEqual(commandAudit(t, audit)[before:], []string{tc.call}) {
						t.Fatalf("binding outcome=%v audit=%v, want %q and %q", outcome, commandAudit(t, audit)[before:], tc.want, tc.call)
					}
				})
			}

		})
	}
}
