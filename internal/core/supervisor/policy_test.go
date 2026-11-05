//go:build unix

package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestCommandPolicyBothTransports 验证真实插件、SQLite 与命令/发送策略的动态组合。
func TestCommandPolicyBothTransports(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "commanddispatch")
	if out, err := exec.Command("go", "build", "-o", binary, "./testdata/commanddispatch").CombinedOutput(); err != nil {
		t.Fatalf("build policy plugin: %v\n%s", err, out)
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
			p := policy.New(store)
			registry := adapter.NewRegistry()
			platform := memory.New("memory", "instance", "account")
			if err := registry.Add(platform); err != nil {
				t.Fatal(err)
			}
			sender := message.New(registry, p, nil, values.RequestTimeout)
			target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat", ThreadId: proto.String("topic")}
			actor := &pb.Actor{ExternalUserId: "user", AdapterInstanceId: "instance"}
			mux := router.NewMux()
			sdkFailures := make(chan *pb.ErrorInfo, 1)
			mux.Handle(wire.ServiceSend, func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
				response := sender.Send(ctx, c, in)
				if response.Err != nil {
					sdkFailures <- proto.CloneOf(response.Err)
				}
				return response
			})
			caller := router.NewCaller(spec.PluginID, spec.GenerationID)
			spec.Handler = mux.For(caller)
			events := router.New(values.DispatchQueueEvents, registry, sender.Send, nil, nil, p)
			events.Start()
			defer events.Close()
			services := broker.New(events, values.RequestTimeout)
			var denyDuringService atomic.Bool
			mux.Forward(func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
				// 在已接纳命令派生的真实 Service 调用前禁止原会话，验证 SDK 回复也重新准入。
				if denyDuringService.Swap(false) {
					if err := p.Deny(ctx, policy.DenyChat, target.AdapterInstanceId, target.ChatId); err != nil {
						return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
					}
				}
				return services.Call(ctx, c, in)
			})
			if err := platform.Start(ctx, events); err != nil {
				t.Fatal(err)
			}
			defer platform.Stop(ctx)
			running, err := Launch(ctx, store, spec)
			if err != nil {
				t.Fatal(err)
			}
			stopped := false
			stop := func() {
				if stopped {
					return
				}
				stopped = true
				result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				if result.Err != nil || result.Killed {
					t.Errorf("policy plugin stop: %+v", result)
				}
			}
			defer stop()
			for _, route := range running.Instance.Routes {
				if route.GetRouteId() == "cmd.rss.list" && !reflect.DeepEqual(route.GetRequiredPermissions(), []string{"user"}) {
					t.Fatalf("list declaration permissions: %v", route.GetRequiredPermissions())
				}
			}
			events.Attach(router.FromInstance(running.Instance), caller)
			pluginScope := policy.Scope{Kind: policy.ScopePlugin}
			groupScope := policy.Scope{Kind: policy.ScopeGroup, Group: []string{"rss"}}
			listScope := policy.Scope{Kind: policy.ScopeRoute, RouteID: "cmd.rss.list"}
			set := func(scope policy.Scope, o policy.Override) {
				t.Helper()
				if err := p.SetOverride(ctx, spec.PluginID, scope, o); err != nil {
					t.Fatal(err)
				}
			}
			// 独立会话中的 barrier 始终获准，其回执证明单 worker 已结束此前消息的全部派发。
			set(policy.Scope{Kind: policy.ScopeRoute, RouteID: "cmd.barrier"}, policy.Override{Permissions: []string{}})
			syncTarget := proto.CloneOf(target)
			syncTarget.ChatId = "sync"
			inject := func(text string, as *pb.Actor, to *pb.Target) {
				t.Helper()
				ev := &pb.Event{EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, AdapterId: proto.String(to.AdapterId), AdapterInstanceId: proto.String(to.AdapterInstanceId), AccountId: proto.String(to.AccountId), Target: to, Actor: as, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}}}}}}
				if err := platform.Inject(ctx, ev); err != nil {
					t.Fatal(err)
				}
			}
			check := func(name, text string, as *pb.Actor, calls, replies []string) {
				t.Helper()
				t.Run(name, func(t *testing.T) {
					beforeAudit, beforeSend := len(commandAudit(t, audit)), len(platform.Sent())
					inject(text, as, target)
					inject("/barrier", &pb.Actor{ExternalUserId: "sync", AdapterInstanceId: "instance"}, syncTarget)
					deadline := time.Now().Add(5 * time.Second)
					for {
						sent := platform.Sent()
						if len(sent) > beforeSend && sent[len(sent)-1].Target.ChatId == "sync" && sent[len(sent)-1].Chain.GetSegments()[0].GetText().GetText() == "barrier" {
							break
						}
						if time.Now().After(deadline) {
							t.Fatalf("policy barrier not reached: sent=%v", sent)
						}
						time.Sleep(10 * time.Millisecond)
					}
					if got := commandAudit(t, audit)[beforeAudit:]; !reflect.DeepEqual(got, calls) {
						t.Fatalf("calls=%v, want %v", got, calls)
					}
					sent := platform.Sent()[beforeSend:]
					if len(sent) != len(replies)+1 {
						t.Fatalf("sends=%v, want %d replies and barrier", sent, len(replies))
					}
					for i, want := range replies {
						if got := sent[i].Chain.GetSegments()[0].GetText().GetText(); got != want {
							t.Fatalf("reply=%q, want %q", got, want)
						}
						if !proto.Equal(sent[i].Target, target) {
							t.Fatalf("reply target changed: %v", sent[i].Target)
						}
					}
				})
			}
			check("default_user", "/rss list", actor, []string{"list", "event"}, []string{"list"})
			set(pluginScope, policy.Override{Permissions: []string{"admin"}})
			check("plugin_admin_rejects_user", "/rss list hello", actor, []string{"event"}, nil)
			p.SetAdmins([]policy.Admin{{InstanceID: "instance", UserID: "user"}})
			check("Core_admin_allowed", "/rss list", actor, []string{"list", "event"}, []string{"list"})
			check("plugin_received_admin", "/permissions", actor, []string{"permissions", "event"}, []string{"admin,user"})
			p.SetAdmins(nil)
			set(pluginScope, policy.Override{})
			forged := proto.CloneOf(actor)
			forged.Permissions = []string{"admin", "user"}
			check("adapter_admin_discarded", "/permissions", forged, []string{"permissions", "event"}, []string{"user"})
			set(groupScope, policy.Override{Permissions: []string{"admin"}})
			set(listScope, policy.Override{Permissions: []string{}})
			check("nearest_empty_list", "/rss list", actor, []string{"list", "event"}, []string{"list"})
			check("sibling_keeps_admin", "/rss add hello", actor, []string{"event"}, nil)
			set(listScope, policy.Override{Permissions: []string{"loop.owner"}})
			check("unknown_permission_user", "/rss list", actor, []string{"event"}, nil)
			p.SetAdmins([]policy.Admin{{InstanceID: "instance", UserID: "user"}})
			check("unknown_permission_admin", "/rss list", actor, []string{"event"}, nil)
			found := false
			rows, err := p.Overrides(ctx, spec.PluginID)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row.Scope.RouteID == listScope.RouteID {
					found = reflect.DeepEqual(row.Override.Permissions, []string{"loop.owner"})
				}
			}
			if !found {
				t.Fatal("undeclared permission override was not retained")
			}
			p.SetAdmins(nil)
			set(groupScope, policy.Override{Enabled: proto.Bool(false)})
			set(listScope, policy.Override{Permissions: []string{}, Enabled: proto.Bool(true)})
			session := running.Instance.Session
			process := running.Process
			check("parent_disabled", "/rss list hello", actor, []string{"event"}, nil)
			set(groupScope, policy.Override{})
			check("reenabled_without_restart", "/rss list", actor, []string{"list", "event"}, []string{"list"})
			if running.Instance.Session != session || running.Process != process || running.Instance.State() != host.StateActivated {
				t.Fatal("policy update replaced or deactivated the plugin")
			}
			p.SetAdmins([]policy.Admin{{InstanceID: "instance", UserID: "user"}})
			if err := p.Deny(ctx, policy.DenyUser, "instance", "user"); err != nil {
				t.Fatal(err)
			}
			check("user_deny_beats_admin", "/rss list hello", actor, []string{"event"}, nil)
			if err := p.Undeny(ctx, policy.DenyUser, "instance", "user"); err != nil {
				t.Fatal(err)
			}
			check("user_undeny_immediate", "/rss list", actor, []string{"list", "event"}, []string{"list"})
			denyDuringService.Store(true)
			check("sdk_reply_original_target_denied", "/service", actor, []string{"event"}, nil)
			select {
			case failure := <-sdkFailures:
				if failure.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
					t.Fatalf("SDK Send error=%v", failure)
				}
			default:
				t.Fatal("SDK Send did not return a failure")
			}
			check("chat_deny_command", "/rss list hello", actor, []string{"event"}, nil)
			// 即使已获实例许可，显式目标发送也受同一会话禁止约束。
			if err := store.SetAdapterInstanceAccess(ctx, spec.PluginID, false, []string{"instance"}); err != nil {
				t.Fatal(err)
			}
			request := &pb.SendRequest{To: target, Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "explicit"}}}}}}
			encoded, err := proto.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			beforeSend := len(platform.Sent())
			response := sender.Send(ctx, caller, &wire.Inbound{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
			if response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || len(platform.Sent()) != beforeSend {
				t.Fatalf("explicit banned target send=%v", response.Err)
			}
			beforeOverrides, err := p.Overrides(ctx, spec.PluginID)
			if err != nil {
				t.Fatal(err)
			}
			beforeDenies, err := p.Denies(ctx)
			if err != nil {
				t.Fatal(err)
			}
			events.Close()
			stop()
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := state.Open(ctx, filepath.Join(filepath.Dir(filepath.Dir(spec.PluginDir)), "data"))
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			restored := policy.New(reopened)
			if got, err := restored.Overrides(ctx, spec.PluginID); err != nil || !reflect.DeepEqual(got, beforeOverrides) {
				t.Fatalf("reopened overrides=%v, %v; want %v", got, err, beforeOverrides)
			}
			if got, err := restored.Denies(ctx); err != nil || !reflect.DeepEqual(got, beforeDenies) {
				t.Fatalf("reopened denies=%v, %v; want %v", got, err, beforeDenies)
			}
			if len(restored.Admins()) != 0 {
				t.Fatal("Core Admin leaked into state database")
			}
		})
	}
}
