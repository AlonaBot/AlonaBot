//go:build unix

package remote_test

import (
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestPlatformAction 经真实 Adapter 进程、Core Mux、发送策略与 SQLite 验证动作准入和许可撤销。
func TestPlatformAction(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	spec.ExtraEnv[helperEnv] = "actions"
	registry := adapter.NewRegistry()
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: &eventSink{}})
	mux := router.NewMux()
	process.Services(mux)
	service := message.New(registry, policy.New(store), nil, nil, tune.Defaults().RequestTimeout, nil)
	mux.Handle(wire.ServiceAction, service.Action)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	})
	if err := process.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	if err := process.StartInstance(ctx, "one", []byte(`{"account":"a"}`)); err != nil {
		t.Fatal(err)
	}
	registry.SetDefaults(map[string]string{"loop": "one"})
	const pluginID = "action-caller"
	if err := store.AddInstall(ctx, state.Install{PluginID: pluginID, Role: state.RolePlugin, BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	handler := mux.For(router.NewCaller(pluginID, "gen"))
	target := &pb.Target{AdapterId: "loop", ChatId: "chat"}
	call := func(params string) wire.Response {
		t.Helper()
		request := &pb.ActionRequest{To: target, Name: "loop.echo", Version: 1, ParamsJson: params}
		encoded, err := proto.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		return handler(ctx, &wire.Inbound{Target: wire.ServiceAction, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
	}
	count := func(want string) {
		t.Helper()
		instance, _ := registry.Get("one")
		result, safe := instance.Send(ctx, &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat"}, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "action-count"}}}}})
		if safe != nil || len(result.GetPlatformMessageIds()) != 1 || result.PlatformMessageIds[0] != want {
			t.Fatalf("adapter action count=%v error=%v want=%s", result, safe, want)
		}
	}
	response := call(`{"value":"ok"}`)
	if response.Err != nil {
		t.Fatal(response.Err)
	}
	result := &pb.ActionResponse{}
	if response.Payload.SchemaID != validate.SchemaID(result) || response.Payload.SchemaVersion != 1 || response.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
		t.Fatalf("unexpected action result schema: %+v", response.Payload)
	}
	if err := proto.Unmarshal(response.Payload.Bytes, result); err != nil || result.ResultJson != ` {"value": "ok", "platform_extra": true} ` {
		t.Fatalf("result=%v error=%v", result, err)
	}
	count("1")
	if response := call(`{"value":7}`); response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
		t.Fatalf("invalid params error=%v", response.Err)
	}
	count("1")
	if err := store.SetAdapterInstanceAccess(ctx, pluginID, false, nil); err != nil {
		t.Fatal(err)
	}
	if response := call(`{"value":"ok"}`); response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
		t.Fatalf("revoked permission error=%v", response.Err)
	}
	count("1")
}
