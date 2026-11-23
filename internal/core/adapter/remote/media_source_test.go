//go:build unix

package remote_test

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	coremessage "github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/core/webhost"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// inspectMediaSource 在真实 Adapter 发送在途时观察 Core 来源，并由平台端拉取临时 URL。
func (i *loopInstance) inspectMediaSource(ctx context.Context, chain message.Chain) (*message.SendResult, *errcode.Error) {
	source := chain[1].(message.ImageSegment).Source
	var observed []string
	switch {
	case source.InlineBase64 != "":
		observed = []string{"inline", source.InlineBase64}
	case source.PlatformFile != "":
		observed = []string{"platform", source.PlatformFile, source.MIME}
	case source.MediaRef != "":
		address, expires, err := i.core.MediaURL(ctx, source.MediaRef, "")
		if err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		if !expires.After(time.Now()) {
			return nil, errcode.Internal.New("URL expired before platform read")
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		data, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil {
			return nil, errcode.Internal.New(readErr.Error())
		}
		if closeErr != nil {
			return nil, errcode.Internal.New(closeErr.Error())
		}
		if response.StatusCode != http.StatusOK {
			return nil, errcode.Internal.New("platform media GET failed: " + response.Status)
		}
		observed = []string{"url", address, string(data), source.MIME}
	default:
		return nil, errcode.Internal.New("missing prepared media source")
	}
	return &message.SendResult{Status: message.SendAccepted, PlatformMessageIDs: observed}, nil
}

// TestMediaSources 通过真实回环 Adapter 验证直通、同实例复用和发送租约内的 URL。
func TestMediaSources(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	registry := adapter.NewRegistry()
	mediaStore := openMediaStore(t)
	mux := router.NewMux()
	dispatch := router.New(router.Options{Concurrency: 1, QueueBytes: tune.Defaults().DispatchQueueBytes, QueueEvents: 1, Adapters: registry})
	t.Cleanup(func() {
		if err := dispatch.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	mediaStore.AddReadServices(mux, registry, dispatch, store.AdapterInstanceAllowed)
	server := httptest.NewUnstartedServer(nil)
	baseURL, err := url.Parse("http://" + server.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	mediaStore.AddURLService(mux, baseURL)
	server.Config.Handler = webhost.New(webhost.Options{BaseURL: baseURL, Media: mediaStore})
	server.Start()
	t.Cleanup(server.Close)
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: &eventSink{}, RegisterIncoming: mediaStore.RegisterIncoming})
	process.Services(mux)
	adapterCaller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(adapterCaller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
		mediaStore.ReleaseSession(adapterCaller.SessionKey)
	})
	if err := process.Bind(ctx, running.Instance, adapterCaller); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"source", "other"} {
		if err := process.StartInstance(ctx, id, []byte(`{"account":"`+id+`","media":true}`)); err != nil {
			t.Fatal(err)
		}
	}
	consumer := router.NewCaller("sender", "generation")
	t.Cleanup(func() { mediaStore.ReleaseSession(consumer.SessionKey) })
	sourceTarget := &pb.Target{AdapterId: "loop", AdapterInstanceId: "source", AccountId: "source", ChatId: "chat"}
	otherTarget := &pb.Target{AdapterId: "loop", AdapterInstanceId: "other", AccountId: "other", ChatId: "chat"}
	service := coremessage.New(registry, policy.New(store), mediaStore.LeaseForSend, nil, tune.Defaults().RequestTimeout, nil)
	mux.Handle(wire.ServiceSend, service.Send)
	send := func(target *pb.Target, source *pb.MediaSource) *pb.SendResult {
		t.Helper()
		request := &pb.SendRequest{Chain: &pb.MessageChain{Segments: []*pb.Segment{
			{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "media-source"}}},
			{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: source}}},
		}}}
		encoded, err := proto.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		out := mux.For(consumer)(ctx, &wire.Inbound{Target: wire.ServiceSend, Deadline: time.Now().Add(tune.Defaults().RequestTimeout), ParentAttach: &router.Invocation{Depth: 1, PluginID: consumer.PluginID, DefaultTarget: target}, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
		if out.Err != nil {
			t.Fatal(out.Err)
		}
		result := &pb.SendResult{}
		if err := proto.Unmarshal(out.Payload.Bytes, result); err != nil {
			t.Fatal(err)
		}
		if result.Status != pb.SendStatus_SEND_STATUS_ACCEPTED {
			t.Fatalf("send result: %v", result)
		}
		return result
	}
	inline := base64.StdEncoding.EncodeToString([]byte("inline bytes"))
	if result := send(sourceTarget, &pb.MediaSource{Source: &pb.MediaSource_InlineBase64{InlineBase64: inline}}); !reflect.DeepEqual(result.PlatformMessageIds, []string{"inline", inline}) {
		t.Fatalf("inline changed: %v", result)
	}
	instance, _ := registry.Get("source")
	id := mediaStore.RegisterIncoming(instance, &pb.IncomingMedia{Token: "platform-file", Size: proto.Uint64(11), Mime: proto.String("application/octet-stream")})
	incoming := mediaStore.IssueIncoming(id, consumer, []*pb.Target{sourceTarget, otherTarget})
	if result := send(sourceTarget, &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: incoming}}); !reflect.DeepEqual(result.PlatformMessageIds, []string{"platform", "platform-file", "application/octet-stream"}) {
		t.Fatalf("platform source changed: %v", result)
	}
	sourceInstance, _ := registry.Get("source")
	count, safe := sourceInstance.Send(ctx, sourceTarget, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "media-read-count"}}}}})
	if safe != nil || !reflect.DeepEqual(count.PlatformMessageIds, []string{"0"}) {
		t.Fatalf("same-instance platform reads: %v %v", count, safe)
	}
	result := send(otherTarget, &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: incoming}})
	if len(result.PlatformMessageIds) != 4 || result.PlatformMessageIds[0] != "url" || result.PlatformMessageIds[2] != "image bytes" || result.PlatformMessageIds[3] != "application/octet-stream" {
		t.Fatalf("platform URL read: %v", result)
	}
	response, err := server.Client().Get(result.PlatformMessageIds[1])
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("URL after send: %s", response.Status)
	}
}
