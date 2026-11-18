//go:build unix

package remote_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	"github.com/AlonaBot/AlonaBot/internal/core/buttons"
	coremessage "github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// TestMessageSegments 验证整链预检，以及新增片段和会话类型跨真实 Adapter、插件 Session 的保留。
func TestMessageSegments(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	registry := adapter.NewRegistry()
	mediaStore := openMediaStore(t)
	dispatch := router.New(router.Options{Concurrency: 1, QueueBytes: tune.Defaults().DispatchQueueBytes, QueueEvents: tune.Defaults().DispatchQueueEvents, Adapters: registry, Prepare: mediaStore.PrepareIncoming})
	dispatch.Start()
	defer func() {
		if err := dispatch.Close(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: dispatch, RegisterIncoming: mediaStore.RegisterIncoming})
	mux := router.NewMux()
	process.Services(mux)
	tokens := buttons.New(store, time.Now, tune.Defaults().ButtonTokenTTL, func(string) (string, []*pb.RouteDescriptor, bool) {
		t.Error("URL buttons queried action routes")
		return "", nil, false
	})
	sender := coremessage.New(registry, policy.New(store), mediaStore.LeaseForSend, tokens.Issue, tune.Defaults().RequestTimeout, nil)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	if err := process.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	for _, instance := range []struct{ id, config string }{{"supported", `{"account":"a","segments":true}`}, {"unsupported", `{"account":"b"}`}} {
		if err := process.StartInstance(ctx, instance.id, []byte(instance.config)); err != nil {
			t.Fatal(err)
		}
	}
	audit := filepath.Join(t.TempDir(), "segments.jsonl")
	if err := os.WriteFile(audit, nil, 0600); err != nil {
		t.Fatal(err)
	}
	observer := spec
	observer.PluginID = "observer"
	observer.Role = state.RolePlugin
	observer.AdapterID = ""
	observer.GenerationID = "observer-gen"
	observer.PluginDir = t.TempDir()
	observer.Handler = nil
	observer.Bootstrap = &pb.Bootstrap{GenerationId: observer.GenerationID, CandidateConfigJson: "{}", DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	observer.ExtraEnv = map[string]string{helperEnv: "plugin", observerAuditEnv: audit, "GORACE": "atexit_sleep_ms=0"}
	if err := store.AddInstall(ctx, state.Install{PluginID: observer.PluginID, Role: state.RolePlugin, BuildDigest: observer.BuildDigest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, observer.PluginID, observer.GenerationID, "{}"); err != nil {
		t.Fatal(err)
	}
	plugin, err := supervisor.Launch(ctx, store, observer)
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	dispatch.Attach(router.FromInstance(plugin.Instance), router.NewCaller(observer.PluginID, observer.GenerationID))
	chain := &pb.MessageChain{Segments: []*pb.Segment{
		{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "segments"}}},
		{Data: &pb.Segment_Voice{Voice: &pb.VoiceSegment{Source: &pb.MediaSource{Source: &pb.MediaSource_HttpsUrl{HttpsUrl: "https://example.com/voice"}}}}},
		{Data: &pb.Segment_Buttons{Buttons: &pb.ButtonsSegment{Rows: []*pb.ButtonRow{{Buttons: []*pb.Button{{Label: "web", Url: proto.String("https://example.com/page"), CoreToken: proto.String("foreign")}, {Label: "bot", Url: proto.String("tg://resolve?domain=bot"), CoreToken: proto.String("foreign")}}}}}}},
	}}
	for _, id := range []string{"unsupported", "supported"} {
		instance, _ := registry.Get(id)
		target := &pb.Target{AdapterId: "loop", AdapterInstanceId: id, AccountId: instance.AccountID(), ChatId: "chat"}
		request := &pb.SendRequest{To: target, Chain: chain}
		encoded, err := proto.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		response := sender.Send(ctx, caller, &wire.Inbound{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
		if id == "unsupported" {
			if response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
				t.Fatalf("capability rejection=%v", response.Err)
			}
			result, safe := instance.Send(ctx, target, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "send-count"}}}}})
			if safe != nil || !reflect.DeepEqual(result.GetPlatformMessageIds(), []string{"0"}) {
				t.Fatalf("unsupported adapter was called: %v %v", result, safe)
			}
		} else if response.Err != nil {
			t.Fatal(response.Err)
		}
	}
	// 以最后一条屏障记录确认所有事件 Handler 已写完。
	if err := dispatch.Deliver(ctx, &pb.Event{EventType: "scheduled", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_SCHEDULER, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_Scheduled{Scheduled: &pb.ScheduledPayload{}}}}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	var reports []observedEvent
	for {
		file, err := os.Open(audit)
		if err != nil {
			t.Fatal(err)
		}
		decoder := json.NewDecoder(file)
		reports = nil
		for decoder.More() {
			var report observedEvent
			if err := decoder.Decode(&report); err != nil {
				t.Fatal(err)
			}
			reports = append(reports, report)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		if len(reports) > 0 && reports[len(reports)-1].Route == "barrier" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("plugin barrier missing: %v", reports)
		}
		time.Sleep(10 * time.Millisecond)
	}
	var received int
	for _, report := range reports {
		if report.Route != "loop" {
			continue
		}
		received++
		if report.ChatType != message.ChatGroup || !reflect.DeepEqual(report.Segments, []string{"text", "voice", "buttons"}) || report.VoiceURL != "https://example.com/voice" || !reflect.DeepEqual(report.ButtonURLs, []string{"https://example.com/page", "tg://resolve?domain=bot"}) || !reflect.DeepEqual(report.ButtonTokens, []string{"", ""}) {
			t.Fatalf("plugin received=%+v", report)
		}
	}
	if received != 1 {
		t.Fatalf("received message events=%d", received)
	}
}
