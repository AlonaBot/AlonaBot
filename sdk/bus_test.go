package sdk

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// TestSubscriptionBudgetPreservesRequests 验证阻塞回调的积压受限，同 Session 的业务请求仍能准入，放行后按序处理保留事件。
func TestSubscriptionBudgetPreservesRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	entered, unblock := make(chan struct{}), make(chan struct{})
	seen := make(chan string, 20)
	r := &SubscriptionRegistry{}
	Subscribe(r, "sub", "pub/x", "test.string", 1, func(_ *Context, value string) error {
		if value[0] == 'a' {
			close(entered)
			<-unblock
		}
		seen <- value[:1]
		return nil
	})
	a := &activation{lifetime: ctx}
	a.ready.Store(true)
	x, y := net.Pipe()
	sid := make([]byte, validate.SessionIDLen)
	plugin := wire.New(wire.NewStreamConn(y), wire.Options{
		Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: sid, SourcePluginID: "sub",
		Event: subscriptionEvents(a, &Registration{subscriptions: r.subscriptions}),
		Handler: func(context.Context, *wire.Inbound) wire.Response {
			return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: "test.reply", SchemaVersion: 1, Bytes: []byte(`{}`)}}
		},
	})
	core := wire.New(wire.NewStreamConn(x), wire.Options{Local: pb.InitiatorSide_INITIATOR_SIDE_CORE, SessionID: sid, SourcePluginID: "sub"})
	released := false
	defer func() {
		if !released {
			close(unblock)
		}
		cancel()
		_ = plugin.Close()
		_ = core.Close()
	}()
	plugin.Start()
	core.Start()
	for i := range 20 {
		payload := []byte(`"` + strings.Repeat(string(rune('a'+i)), 1<<20) + `"`)
		ev := &pb.BusEvent{EventName: "pub/x", SchemaId: "test.string", SchemaVersion: 1, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Payload: payload}
		if err := core.SendEvent(ctx, "sub", ev); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
	}
	// REQUEST 在 EVENT 之后读取，2 MiB 载荷可区分 SDK 的 4 MiB 上限与原有共享 16 MiB 积压。
	if _, err := core.Call(ctx, wire.Request{Target: "request", Payload: wire.Payload{
		Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: "test.request", SchemaVersion: 1, Bytes: make([]byte, 2<<20),
	}}); err != nil {
		t.Fatalf("request while subscription is blocked: %v", err)
	}
	close(unblock)
	released = true
	// 每条略大于 1 MiB，因此 4 MiB 内恰好保留 a/b/c，包含正在处理的 a。
	for _, want := range []string{"a", "b", "c"} {
		select {
		case got := <-seen:
			if got != want {
				t.Fatalf("callback = %q, want %q", got, want)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	// 顺序屏障：后续小事件证明溢出事件已丢弃，预算也已经归还。
	if err := core.SendEvent(ctx, "sub", &pb.BusEvent{EventName: "pub/x", SchemaId: "test.string", SchemaVersion: 1, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Payload: []byte(`"z"`)}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-seen:
		if got != "z" {
			t.Fatalf("callback after overflow = %q", got)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if got := plugin.Stats().DroppedEvents; got != 0 {
		t.Fatalf("wire dropped events = %d; SDK should release excess events", got)
	}
}
