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

// TestBoundPlatformAction 验证声明作用域、可信默认目标及跨实例授权使用同一动作入口。
func TestBoundPlatformAction(t *testing.T) {
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
	for id, config := range map[string]string{"one": `{"account":"a"}`, "two": `{"account":"b"}`} {
		if err := process.StartInstance(ctx, id, []byte(config)); err != nil {
			t.Fatal(err)
		}
	}
	registry.SetDefaults(map[string]string{"loop": "two"})
	const pluginID = "bound-caller"
	if err := store.AddInstall(ctx, state.Install{PluginID: pluginID, Role: state.RolePlugin, BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	if err := store.SetAdapterInstanceAccess(ctx, pluginID, false, nil); err != nil {
		t.Fatal(err)
	}
	target := &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat"}
	parent := &router.Invocation{PluginID: pluginID, DefaultTarget: target}
	handler := mux.For(router.NewCaller(pluginID, "gen"))
	call := func(name, instance string, automatic bool, to *pb.Target, parent *router.Invocation, version uint32, params string) wire.Response {
		t.Helper()
		request := &pb.ActionRequest{To: to, AdapterInstanceId: instance, Name: name, Version: version, ParamsJson: params, UseDefaultTarget: automatic}
		encoded, err := proto.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		return handler(ctx, &wire.Inbound{Target: wire.ServiceAction, ParentAttach: parent, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
	}
	for _, tc := range []struct {
		name, action, instance string
		automatic              bool
		to                     *pb.Target
		parent                 *router.Invocation
		version                uint32
		params                 string
		code                   pb.ErrorCode
	}{
		{"chat", "loop.echo", "one", true, nil, parent, 1, `{"value":"ok"}`, 0},
		{"account", "loop.account", "one", true, nil, parent, 1, `{"value":"ok"}`, 0},
		{"account event", "loop.account", "one", true, nil, &router.Invocation{PluginID: pluginID, Event: &pb.Event{AdapterInstanceId: proto.String("one")}}, 1, `{"value":"ok"}`, 0},
		{"missing chat", "loop.echo", "one", true, nil, &router.Invocation{PluginID: pluginID, Event: &pb.Event{AdapterInstanceId: proto.String("one")}}, 1, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET},
		{"missing parent", "loop.echo", "one", true, nil, nil, 1, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{"other instance", "loop.echo", "two", true, nil, parent, 1, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{"other chat", "loop.echo", "", false, &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", ChatId: "other"}, parent, 1, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{"old empty chat", "loop.echo", "one", false, nil, parent, 1, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{"old account", "loop.account", "one", false, nil, parent, 1, `{"value":"ok"}`, 0},
		{"explicit chat", "loop.echo", "", false, target, parent, 1, `{"value":"ok"}`, 0},
		{"version", "loop.echo", "one", true, nil, parent, 2, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_UNSUPPORTED},
		{"schema", "loop.echo", "one", true, nil, parent, 1, `{"value":7}`, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{"hook", "loop.echo", "one", true, nil, &router.Invocation{PluginID: pluginID, DefaultTarget: target, SendForbidden: true}, 1, `{"value":"ok"}`, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := call(tc.action, tc.instance, tc.automatic, tc.to, tc.parent, tc.version, tc.params)
			if response.Err.GetCode() != tc.code {
				t.Fatalf("error=%v want=%v", response.Err, tc.code)
			}
			if tc.code == 0 {
				result := &pb.ActionResponse{}
				if err := proto.Unmarshal(response.Payload.Bytes, result); err != nil || result.ResultJson != ` {"value": "ok", "platform_extra": true} ` {
					t.Fatalf("result=%v error=%v", result, err)
				}
			}
		})
	}
	if err := store.SetAdapterInstanceAccess(ctx, pluginID, false, []string{"two"}); err != nil {
		t.Fatal(err)
	}
	if response := call("loop.echo", "two", true, nil, parent, 1, `{"value":"ok"}`); response.Err != nil {
		t.Fatal(response.Err)
	}
	for id, want := range map[string]string{"one": "5", "two": "1"} {
		instance, _ := registry.Get(id)
		result, safe := instance.Send(ctx, &pb.Target{AdapterId: "loop", AdapterInstanceId: id, ChatId: "chat"}, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "action-count"}}}}})
		if safe != nil || len(result.GetPlatformMessageIds()) != 1 || result.PlatformMessageIds[0] != want {
			t.Fatalf("instance=%s result=%v error=%v want=%s", id, result, safe, want)
		}
	}
}
