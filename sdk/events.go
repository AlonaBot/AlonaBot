package sdk

import (
	"fmt"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

const commandRoutePrefix = "cmd."

// EventRoute 保存一个手动发送事件 Handler 的静态注册。
type EventRoute struct {
	RouteID    string
	EventType  string
	Handler    func(*Context) error
	permission Permission
	Priority   int32 // 数值越小越先，缺省为 0
}

// EventRegistry 按注册顺序收集事件路由，错误由 Collect 统一返回。
// 零值可用；Collect 另将命令的 route_id 登记到 seen，检查跨注册器冲突。
type EventRegistry struct {
	builders []*EventBuilder
	events   []EventRoute
	errs     []error
	seen     map[string]bool
}

// On 开始构建事件路由。
func (r *EventRegistry) On(routeID, eventType string) *EventBuilder {
	b := &EventBuilder{route: EventRoute{RouteID: routeID, EventType: eventType}}
	r.builders = append(r.builders, b)
	return b
}

// EventBuilder 收集一个事件路由；Handle 结束构建。
type EventBuilder struct {
	route       EventRoute
	handles     int
	afterHandle bool
}

// Priority 设置路由优先级。
func (b *EventBuilder) Priority(p int32) *EventBuilder {
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.route.Priority = p
	return b
}

// Permission 设置事件路由权限；缺省为空集合。
func (b *EventBuilder) Permission(p Permission) *EventBuilder {
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.route.permission = p
	return b
}

// Handle 结束事件路由构建。
func (b *EventBuilder) Handle(h func(*Context) error) { b.handles++; b.route.Handler = h }

func (r *EventRegistry) collect() {
	for _, b := range r.builders {
		if b.handles != 1 || b.afterHandle {
			r.errs = append(r.errs, fmt.Errorf("event route %q has invalid Handle sequence", b.route.RouteID))
			continue
		}
		if err := checkPermission(b.route.permission); err != nil {
			r.errs = append(r.errs, fmt.Errorf("event route %q: %w", b.route.RouteID, err))
			continue
		}
		r.add(b.route)
	}
}

func (r *EventRegistry) add(route EventRoute) {
	routeID, eventType, handler := route.RouteID, route.EventType, route.Handler
	switch {
	case routeID == "":
		r.errs = append(r.errs, fmt.Errorf("event route is missing route_id"))
		return
	case strings.HasPrefix(routeID, commandRoutePrefix):
		r.errs = append(r.errs, fmt.Errorf("event route %q uses reserved %s prefix", routeID, commandRoutePrefix))
		return
	case r.seen[routeID]:
		r.errs = append(r.errs, fmt.Errorf("route %q registered twice", routeID))
		return
	case !validate.IsEventType(eventType) && !validate.IsAdapterEventType(eventType):
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
	r.events = append(r.events, route)
}

func commandRouteID(command Command) string {
	if command.routeID != "" {
		return command.routeID
	}
	return commandRoutePrefix + strings.Join(command.Path, ".")
}
