//go:build unix

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/sdk"
)

const messageFilterMode = "message-filter"

const messageFilterSchema = `{
  "chat": {"type":"object", "items":{
    "triggers":{"type":"list", "options":["mention", "private", "wake_word"]},
    "words":{"type":"list"}
  }}
}`

func messageFilterPlugin() *sdk.Plugin {
	p := sdk.NewPlugin()
	genresources.Bind(p, genresources.Resources{ConfigSchema: []byte(messageFilterSchema)})
	p.Command("ping").Handle(func(c *sdk.Context) error { return pipelineObserve("handler", "cmd.ping", c) })
	p.On("filtered", sdk.MessageCreated).
		Filter(sdk.MessageFilter{TriggersConfig: "chat.triggers", WakeWordsConfig: "chat.words"}).
		Rule(sdk.NotCommand()).Handle(func(c *sdk.Context) error { return pipelineObserve("handler", "filtered", c) })
	p.On("plain", sdk.MessageCreated).Handle(func(c *sdk.Context) error { return pipelineObserve("handler", "plain", c) })
	return p
}

// TestMessageFilter 验证真实 stdio Session 的注册与 Handler REQUEST，以及注入快照的会话隔离。
func TestMessageFilter(t *testing.T) {
	const pluginID = "filtered-plugin"
	var useOverride atomic.Bool
	h := newPipelineWithMessageConfig(t, t.TempDir(), tune.Defaults(), true, func(_ context.Context, id string, ev *pb.Event) (json.RawMessage, error) {
		if id != pluginID {
			return nil, fmt.Errorf("unexpected plugin %q", id)
		}
		switch ev.GetTarget().GetChatId() {
		case "first":
			if useOverride.Load() {
				return json.RawMessage(`{"chat":{"triggers":["wake_word"],"words":["silver"]}}`), nil
			}
			return json.RawMessage(`{"chat":{"triggers":["wake_word"],"words":["amber", "ping"]}}`), nil
		case "second":
			return json.RawMessage(`{"chat":{"triggers":["mention"],"words":["amber"]}}`), nil
		default:
			return nil, fmt.Errorf("unexpected chat %q", ev.GetTarget().GetChatId())
		}
	})
	h.install(pluginID, messageFilterMode)
	c := h.launch(pluginID, messageFilterMode, 0)
	routes := c.instance.Routes
	i := slices.IndexFunc(routes, func(route *pb.RouteDescriptor) bool { return route.GetRouteId() == "filtered" })
	if i < 0 || routes[i].GetEvent().GetMessageFilter().GetTriggersConfig() != "chat.triggers" ||
		routes[i].GetEvent().GetMessageFilter().GetWakeWordsConfig() != "chat.words" || !routes[i].GetEvent().GetRule().GetHasCustomRule() {
		t.Fatal("activated route is missing filter paths or custom Rule")
	}
	for n, tc := range []struct {
		name, chat, text           string
		toMe, override             bool
		wantRequests, wantHandlers []string
	}{
		{"miss", "first", "ordinary", false, false, []string{"plain"}, []string{"plain"}},
		{"wake word", "first", "prefix AMBER", false, false, []string{"filtered", "plain"}, []string{"filtered", "plain"}},
		{"other chat does not use wake word", "second", "prefix AMBER", false, false, []string{"plain"}, []string{"plain"}},
		{"other chat selects mention", "second", "ordinary", true, false, []string{"filtered", "plain"}, []string{"filtered", "plain"}},
		{"SDK rejects consumed command", "first", "/ping", false, false, []string{"cmd.ping", "filtered", "plain"}, []string{"cmd.ping", "plain"}},
		{"override replaces old word", "first", "amber", false, true, []string{"plain"}, []string{"plain"}},
		{"override matches new word", "first", "silver", false, true, []string{"filtered", "plain"}, []string{"filtered", "plain"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			useOverride.Store(tc.override)
			eventID := fmt.Sprintf("filter-%d", n)
			chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: tc.text}}}}}
			ev := &pb.Event{EventId: eventID, EventType: "message.created", ChatType: pb.ChatType_CHAT_TYPE_GROUP,
				Target: &pb.Target{AdapterId: "loop", AdapterInstanceId: pipelineInstanceID, AccountId: pipelineAccountID, ChatId: tc.chat},
				Actor:  &pb.Actor{ExternalUserId: "user", AdapterInstanceId: pipelineInstanceID},
				EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{
					PlatformMessageId: eventID, ToMe: tc.toMe, Chain: chain}}}}
			if err := h.assembly.router.Deliver(t.Context(), ev); err != nil {
				t.Fatal(err)
			}
			h.idle()
			h.mu.Lock()
			defer h.mu.Unlock()
			var requests, handlers []string
			for _, call := range h.outcomes {
				if call.Request.GetEvent().GetEventId() != eventID {
					continue
				}
				route := call.Request.GetRouteId()
				requests = append(requests, route)
				wantConsumed := tc.text == "/ping" && route != "cmd.ping"
				if call.Code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED || call.Request.GetContext().GetCommandConsumed() != wantConsumed ||
					call.Outcome.GetMatched() != (tc.text != "/ping" || route != "filtered") {
					t.Fatalf("%s: code=%v consumed=%v matched=%v", route, call.Code, call.Request.GetContext().GetCommandConsumed(), call.Outcome.GetMatched())
				}
			}
			for _, record := range h.records {
				if record.Stage == "handler" && record.MessageID == eventID {
					handlers = append(handlers, record.Route)
					if record.Text != tc.text {
						t.Errorf("handler text = %q, want %q", record.Text, tc.text)
					}
				}
			}
			if !slices.Equal(requests, tc.wantRequests) || !slices.Equal(handlers, tc.wantHandlers) {
				t.Fatalf("requests=%v handlers=%v; want %v and %v", requests, handlers, tc.wantRequests, tc.wantHandlers)
			}
		})
	}
}
