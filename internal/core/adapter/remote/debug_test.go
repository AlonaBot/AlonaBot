//go:build unix

package remote_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	coremessage "github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/debug"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TestDebugAdapter 经真实 SDK 子进程、目录 RPC、管理认证和公共客户端验证调用及实例替换。
func TestDebugAdapter(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	spec.Stdio = true
	spec.ExtraEnv[helperEnv] = "actions"
	registry := adapter.NewRegistry()
	media := openMediaStore(t)
	sink := &eventSink{}
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: sink, RegisterIncoming: media.RegisterIncoming})
	mux := router.NewMux()
	media.AddReadServices(mux, registry, nil, store.AdapterInstanceAllowed)
	process.Services(mux)
	messages := coremessage.New(registry, policy.New(store), nil, nil, tune.Defaults().RequestTimeout, nil)
	mux.Handle(wire.ServiceSend, messages.Send)
	mux.Handle(wire.ServiceDelete, messages.Delete)
	mux.Handle(wire.ServiceAction, messages.Action)
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
	audit := filepath.Join(t.TempDir(), "delete.json")
	if err := os.WriteFile(audit, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	config := []byte(`{"account":"a","delete":true,"delete_audit":` + quoteJSON(audit) + `}`)
	if err := process.StartInstance(ctx, "one", config); err != nil {
		t.Fatal(err)
	}
	credentials := t.TempDir()
	token, err := adminauth.InitToken(ctx, store, credentials)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse("http://localhost")
	manager := manage.New(manage.Options{Store: store, Adapters: registry, DebugCall: func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response { return mux.For(c)(ctx, in) }, DebugRelease: func(c router.Caller) {
		if err := messages.Release(c); err != nil {
			t.Error(err)
		}
	}, DebugMedia: media.ReadDebugIncoming})
	client := &debug.Client{URL: base.String(), Token: token, HTTP: &http.Client{Transport: debugTransport{manager.Handler(adminauth.New(store, time.Now, base, credentials))}}}
	operations, err := client.Operations(ctx, "one", "")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, op := range operations {
		names[op.Name] = true
		if !json.Valid(op.ParamsSchema) || !json.Valid(op.ResultSchema) {
			t.Fatalf("invalid operation schema: %+v", op)
		}
	}
	for _, name := range []string{"sdk/message.send", "sdk/message.delete", "loop.echo", "loop.account"} {
		if !names[name] {
			t.Fatalf("missing operation %q in %v", name, names)
		}
	}
	if names["sdk/message.edit"] || names["sdk/media.read"] || names["sdk/command.menu"] {
		t.Fatalf("wrapper exposed unsupported operations: %v", names)
	}
	unauthorized := &debug.Client{URL: client.URL, Token: "invalid", HTTP: client.HTTP}
	if _, err := unauthorized.Operations(ctx, "one", ""); err == nil {
		t.Fatal("invalid admin token was accepted")
	}
	result, err := client.Invoke(ctx, debug.Call{Instance: "one", Name: "loop.echo", Version: 1, To: json.RawMessage(`{"chatId":"chat"}`), Params: json.RawMessage(`{"value":"ok"}`)})
	if err != nil || len(result.TraceID) != 32 || string(result.Payload) != `{"value":"ok","platform_extra":true}` {
		t.Fatalf("action result=%+v error=%v", result, err)
	}
	sent, err := client.Invoke(ctx, debug.Call{Instance: "one", Name: "sdk/message.send", Version: 1, To: json.RawMessage(`{"chatId":"chat"}`), Params: json.RawMessage(`{"chain":{"segments":[{"text":{"text":"send-count"}}]}}`)})
	if err != nil || len(sent.Payload) == 0 {
		t.Fatalf("send result=%+v error=%v", sent, err)
	}

	partial, err := client.Invoke(ctx, debug.Call{Instance: "one", Name: "sdk/message.delete", Version: 1, To: json.RawMessage(`{"chatId":"chat"}`), Params: json.RawMessage(`{"platformMessageIds":["first","reject","last"]}`)})
	if err == nil || partial.Error == nil || len(partial.Payload) == 0 || partial.TraceID == "" {
		t.Fatalf("partial result=%+v error=%v", partial, err)
	}
	var receipt struct{ Items []struct{ Status string } }
	if err := json.Unmarshal(partial.Payload, &receipt); err != nil || len(receipt.Items) != 3 {
		t.Fatalf("delete receipt=%s error=%v", partial.Payload, err)
	}
	previous, _ := registry.Get("one")
	if err := process.StopInstance(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Operations(ctx, "one", ""); err == nil {
		t.Fatal("stopped instance returned a catalog")
	}
	if err := process.StartInstance(ctx, "one", []byte(`{"account":"b","media":true}`)); err != nil {
		t.Fatal(err)
	}
	operations, err = client.Operations(ctx, "one", "sdk/message.delete")
	if err != nil || len(operations) != 0 {
		t.Fatalf("old catalog survived replacement: %+v %v", operations, err)
	}
	target := &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "b", ChatId: "chat"}
	chain := &pb.MessageChain{Segments: []*pb.Segment{
		{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "media-created"}}},
		{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: &pb.IncomingMedia{Token: "platform-file"}}}}}},
	}}
	// 在真实 Session 里重放旧代次，模拟该调用在 SDK 查表前遇上重启。
	stale := &pb.AdapterSendRequest{InstanceId: "one", InstanceGeneration: 1, Target: &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat"}, Chain: chain}
	encoded, err := proto.Marshal(stale)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := running.Instance.Session.Call(ctx, wire.Request{Target: wire.ServiceAdapterSend, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(stale), SchemaVersion: 1, Bytes: encoded}}); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("stale send error = %v", err)
	}
	read := debug.Call{Instance: "one", Name: "sdk/media.read", Version: 1, Params: json.RawMessage(`{"fileId":"platform-file","length":64}`)}
	if result, err := client.Invoke(ctx, read); err == nil || result.Error == nil || result.Error.Code != "NOT_FOUND" {
		t.Fatalf("unregistered file read result=%+v error=%v", result, err)
	}
	current, _ := registry.Get("one")
	injected := &pb.Event{EventType: "message.created", AdapterId: proto.String("loop"), AdapterInstanceId: proto.String("one"), AccountId: proto.String("b"), Target: target, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: chain}}}}
	if err := current.(interface {
		Inject(context.Context, *pb.Event) error
	}).Inject(ctx, injected); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
		t.Fatalf("injected platform media error = %v", err)
	}
	if _, safe := current.Send(ctx, target, chain); safe != nil {
		t.Fatal(safe)
	}
	events := sink.snapshot()
	object := events[len(events)-1].GetMessage().GetChain().GetSegments()[1].GetImage().GetSource().GetIncomingRef().GetToken()
	read.Params = json.RawMessage(`{"fileId":` + quoteJSON(object) + `,"length":64}`)
	mediaResult, err := client.Invoke(ctx, read)
	if err != nil {
		t.Fatalf("media read result=%+v error=%v", mediaResult, err)
	}
	var chunk pb.ReadResponse
	if err := protojson.Unmarshal(mediaResult.Payload, &chunk); err != nil || string(chunk.Data) != "image bytes" || !chunk.Eof {
		t.Fatalf("media read payload=%s error=%v", mediaResult.Payload, err)
	}

	if _, err := previous.(adapter.OperationCatalog).Operations(ctx); err == nil {
		t.Fatal("old instance returned a catalog")
	}
	if _, err := client.Invoke(ctx, debug.Call{Instance: "one", Name: "loop.account", Version: 1, Params: json.RawMessage(`{"value":"still running"}`)}); err != nil {
		t.Fatal(err)
	}
}

func quoteJSON(value string) string { raw, _ := json.Marshal(value); return string(raw) }

type debugTransport struct{ handler http.Handler }

func (d debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	d.handler.ServeHTTP(w, req)
	return w.Result(), nil
}
