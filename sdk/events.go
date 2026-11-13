package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const commandRoutePrefix = "cmd."

// EventRoute 保存一个手动发送事件 Handler 的静态注册。
type EventRoute struct {
	RouteID       string
	EventType     EventType
	Handler       any
	permission    Permission
	rules         []Rule // Collect 固定的调用顺序，含按钮路由。
	continuation  bool
	actionID      string
	actionVersion uint32
	decodeButton  func(string) (any, error)
	handleButton  func(*Context, any) error
	subID         string
	Priority      int32 // 数值越小越先，缺省为 0
}

// eventRegistry 按注册顺序收集事件路由，错误由 Collect 统一返回。
// 零值可用；Collect 另将命令的 route_id 登记到 seen，检查跨注册器冲突。
type eventRegistry struct {
	plugin   *Plugin
	builders []*EventBuilder
	events   []EventRoute
	errs     []error
	seen     map[string]bool
}

// On 开始构建事件路由。
func (r *eventRegistry) On(routeID string, eventType EventType) *EventBuilder {
	r.plugin.checkRegistration()
	b := &EventBuilder{plugin: r.plugin, route: EventRoute{RouteID: routeID, EventType: eventType}}
	r.builders = append(r.builders, b)
	return b
}

// EventBuilder 收集一个事件路由；Handle 结束构建。
type EventBuilder struct {
	plugin      *Plugin
	route       EventRoute
	handles     int
	afterHandle bool
	button      bool // 仅由 OnButton 开始的构建链可占用保留前缀
}

// Priority 设置路由优先级。
func (b *EventBuilder) Priority(p int32) *EventBuilder {
	b.plugin.checkRegistration()
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.route.Priority = p
	return b
}

// Permission 设置事件路由权限；缺省为空集合。
func (b *EventBuilder) Permission(p Permission) *EventBuilder {
	b.plugin.checkRegistration()
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.route.permission = p
	return b
}

// Handle 结束事件路由构建。
func (b *EventBuilder) Handle(h any) { b.plugin.checkRegistration(); b.handles++; b.route.Handler = h }

// Rule 按调用与参数顺序追加条件，全部匹配后执行 Handler。
func (b *EventBuilder) Rule(rules ...Rule) *EventBuilder {
	b.plugin.checkRegistration()
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.route.rules = append(b.route.rules, rules...)
	return b
}

// Continue 让事件 Handler 在独立的延续调用中执行。
func (b *EventBuilder) Continue() *EventBuilder {
	if b.handles != 0 {
		b.afterHandle = true
	}
	b.route.continuation = true
	return b
}

func (r *eventRegistry) collect() {
	r.events = nil
	r.errs = nil
	for _, b := range r.builders {
		if b.handles != 1 || b.afterHandle || (strings.HasPrefix(b.route.RouteID, buttonRoutePrefix) && !b.button) || (b.button && (b.route.actionID == "" || b.route.actionVersion == 0)) {
			r.errs = append(r.errs, fmt.Errorf("event route %q has invalid Handle sequence", b.route.RouteID))
			continue
		}
		if err := checkPermission(b.route.permission); err != nil {
			r.errs = append(r.errs, fmt.Errorf("event route %q: %w", b.route.RouteID, err))
			continue
		}
		if err := checkRules(b.route.rules); err != nil {
			r.errs = append(r.errs, fmt.Errorf("event route %q: %w", b.route.RouteID, err))
			continue
		}
		route := b.route
		route.rules = slices.Clone(route.rules)
		r.add(route)
	}
}

func (r *eventRegistry) add(route EventRoute) {
	routeID, handler := route.RouteID, route.Handler
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
	case handler == nil:
		r.errs = append(r.errs, fmt.Errorf("handler for event route %q is nil", routeID))
		return
	}
	if err := checkEventHandler(route); err != nil {
		r.errs = append(r.errs, err)
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

// EventType 绑定事件名及 Handler 的载荷类型，零值由 Collect 拒绝。
type EventType struct {
	name    string
	payload reflect.Type
	version uint32
	adapter bool
}

var (
	MessageCreated = EventType{name: "message.created", payload: reflect.TypeFor[*message.CreatedPayload]()}      // 新消息。
	MessageEdited  = EventType{name: "message.edited", payload: reflect.TypeFor[*message.EditedPayload]()}        // 消息编辑。
	MessageDeleted = EventType{name: "message.deleted", payload: reflect.TypeFor[*message.DeletedPayload]()}      // 消息删除。
	ButtonClicked  = EventType{name: "button.clicked", payload: reflect.TypeFor[*message.ButtonClickedPayload]()} // 按钮点击。
	MemberJoined   = EventType{name: "member.joined", payload: reflect.TypeFor[*message.MemberPayload]()}         // 成员加入。
	MemberLeft     = EventType{name: "member.left", payload: reflect.TypeFor[*message.MemberPayload]()}           // 成员离开。
	ReactionAdded  = EventType{name: "reaction.added", payload: reflect.TypeFor[*message.ReactionAddedPayload]()} // 消息回应。
	Scheduled      = EventType{name: "scheduled", payload: reflect.TypeFor[*message.ScheduledPayload]()}          // 定时任务。
	BotOnline      = EventType{name: "bot.online", payload: reflect.TypeFor[*message.BotOnlinePayload]()}         // 机器人上线。
	BotOffline     = EventType{name: "bot.offline", payload: reflect.TypeFor[*message.BotOfflinePayload]()}       // 机器人离线。
)

// AdapterEvent 声明 Adapter 扩展事件，T 须为结构体，Handler 接收 *T。
func AdapterEvent[T any](name string, version uint32) EventType {
	return EventType{name: name, payload: reflect.TypeFor[*T](), version: version, adapter: true}
}
func (e EventType) check() error {
	if e.payload == nil {
		return fmt.Errorf("zero EventType")
	}
	if e.adapter {
		if !validate.IsAdapterEventType(e.name) || e.version == 0 || e.payload.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("invalid AdapterEvent %q: expected a valid name, nonzero version and struct payload", e.name)
		}
	} else if !validate.IsEventType(e.name) {
		return fmt.Errorf("unknown event type %q", e.name)
	}
	return nil
}
func checkEventHandler(route EventRoute) error {
	event := route.EventType
	if err := event.check(); err != nil {
		return fmt.Errorf("event route %q: %w", route.RouteID, err)
	}
	t := reflect.TypeOf(route.Handler)
	expected := fmt.Sprintf("func(*sdk.Context) error or func(*sdk.Context, %s) error", event.payload)
	if t != nil && t.Kind() == reflect.Func && !reflect.ValueOf(route.Handler).IsNil() && !t.IsVariadic() && t.NumIn() >= 1 && t.In(0) == contextType && t.NumOut() == 1 && t.Out(0) == errorType && (t.NumIn() == 1 || (t.NumIn() == 2 && t.In(1) == event.payload)) {
		return nil
	}
	return fmt.Errorf("event route %q: expected %s; got %T", route.RouteID, expected, route.Handler)
}
func invokeEvent(c *Context, route EventRoute) error {
	handler := reflect.ValueOf(route.Handler)
	event := route.EventType
	payload := c.Event().Payload
	var value reflect.Value
	if event.adapter {
		if handler.Type().NumIn() == 1 {
			return eventResult(handler.Call([]reflect.Value{reflect.ValueOf(c)}))
		}
		raw, ok := payload.(*message.AdapterEventPayload)
		if !ok {
			c.Logger().Error("event payload type mismatch")
			return errcode.Internal.New("event payload type mismatch")
		}
		if raw.Version != event.version {
			c.Logger().Warn("adapter event version mismatch", "route_id", route.RouteID, "version", raw.Version, "expected_version", event.version)
			return nil
		}
		value = reflect.New(event.payload.Elem())
		if err := json.Unmarshal([]byte(raw.PayloadJSON), value.Interface()); err != nil {
			c.Logger().Warn("decode adapter event", "route_id", route.RouteID, "event_type", event.name, "error_type", fmt.Sprintf("%T", err))
			return nil
		}
	} else {
		value = reflect.ValueOf(payload)
		if !value.IsValid() || value.Type() != event.payload || value.IsNil() {
			c.Logger().Error("event payload type mismatch")
			return errcode.Internal.New("event payload type mismatch")
		}
	}
	if route.decodeButton != nil {
		button := payload.(*message.ButtonClickedPayload)
		decoded, err := route.decodeButton(button.Value)
		if err != nil {
			c.Logger().Warn("decode button value", "route_id", route.RouteID, "error", err)
			return nil
		}
		return route.handleButton(c, decoded)
	}
	if handler.Type().NumIn() == 1 {
		return eventResult(handler.Call([]reflect.Value{reflect.ValueOf(c)}))
	}
	return eventResult(handler.Call([]reflect.Value{reflect.ValueOf(c), value}))
}
func eventResult(values []reflect.Value) error {
	if values[0].IsNil() {
		return nil
	}
	return values[0].Interface().(error)
}
