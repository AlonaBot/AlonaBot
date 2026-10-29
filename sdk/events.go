package sdk

import (
	"fmt"
	"strings"
)

// EventRoute 保存一个手动发送事件 Handler 的静态注册。
type EventRoute struct {
	RouteID   string
	EventType string
	Handler   func(*Context) error
	Priority  int32 // 数值越小越先，缺省为 0
}

// EventRegistry 按注册顺序收集事件路由，错误由 Collect 统一返回。
// 零值可用；Collect 另将命令的 route_id 登记到 seen，检查跨注册器冲突。
type EventRegistry struct {
	events []EventRoute
	errs   []error
	seen   map[string]bool
}

// On 注册事件路由；空 ID、重复 ID、未知事件类型或 nil Handler 记为注册错误。
func (r *EventRegistry) On(routeID, eventType string, handler func(*Context) error, opts ...RouteOption) {
	var options routeOptions
	for _, option := range opts {
		option(&options)
	}
	switch {
	case routeID == "":
		r.errs = append(r.errs, fmt.Errorf("event route is missing route_id"))
		return
	case strings.HasPrefix(routeID, "cmd."):
		r.errs = append(r.errs, fmt.Errorf("event route %q uses reserved cmd. prefix", routeID))
		return
	case len(options.aliases) != 0:
		r.errs = append(r.errs, fmt.Errorf("event route %q cannot declare aliases", routeID))
		return
	case r.seen[routeID]:
		r.errs = append(r.errs, fmt.Errorf("route %q registered twice", routeID))
		return
	case !validEventType(eventType):
		r.errs = append(r.errs, fmt.Errorf("event route %q has unknown event type %q", routeID, eventType))
		return
	case handler == nil:
		r.errs = append(r.errs, fmt.Errorf("handler for event route %q is nil", routeID))
		return
	}
	if r.seen == nil {
		r.seen = make(map[string]bool)
	}
	r.seen[routeID] = true
	r.events = append(r.events, EventRoute{RouteID: routeID, EventType: eventType, Handler: handler, Priority: options.priority})
}

func validEventType(eventType string) bool {
	// message.proto 的 Event.event_type 与 EventPayload 的分支一一对应。
	switch eventType {
	case "message.created", "message.edited", "message.deleted", "button.clicked",
		"member.joined", "member.left", "reaction.added", "scheduled", "bot.online", "bot.offline":
		return true
	}
	return false
}

func commandRouteID(command Command) string { return "cmd." + strings.Join(command.Path, ".") }
