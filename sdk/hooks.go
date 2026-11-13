package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

type hookRoute struct {
	scope      *pluginScope
	permission Permission
	id         string
	phase      pb.HookPhase
	event      EventType
	priority   int32
	handler    any
	handles    int
}

func (r *pluginScope) addHook(id string, phase pb.HookPhase, event EventType) *hookRoute {
	r.plugin.checkRegistration()
	route := &hookRoute{scope: r, id: id, phase: phase, event: event}
	r.plugin.hooks = append(r.plugin.hooks, route)
	return route
}

// PreEvent 开始登记一种事件的业务载荷改写 Hook。
func (p *Plugin) PreEvent(routeID string, event EventType) *PreEventBuilder {
	return &PreEventBuilder{route: p.scope.addHook(routeID, pb.HookPhase_HOOK_PHASE_PRE_EVENT, event)}
}

// PreEvent 开始登记子插件的业务载荷改写 Hook。
func (s *SubPlugin) PreEvent(routeID string, event EventType) *PreEventBuilder {
	return &PreEventBuilder{route: s.scope.addHook(routeID, pb.HookPhase_HOOK_PHASE_PRE_EVENT, event)}
}

// PreEventBuilder 构建 PreEvent Hook，Handle 的输入和输出载荷类型均须与事件一致。
type PreEventBuilder struct{ route *hookRoute }

// Priority 设置 Hook 优先级，数值越小越先执行。
func (b *PreEventBuilder) Priority(p int32) *PreEventBuilder {
	b.route.scope.plugin.checkRegistration()
	b.route.priority = p
	return b
}

// Handle 设置 func(*Context, P) (P, error)，返回 nil 载荷时保留当前内容。
func (b *PreEventBuilder) Handle(h any) {
	b.route.scope.plugin.checkRegistration()
	b.route.handles++
	b.route.handler = h
}

// BeforeSend 开始注册消息链改写 Hook。
func (p *Plugin) BeforeSend(routeID string) *BeforeSendBuilder {
	return &BeforeSendBuilder{route: p.scope.addHook(routeID, pb.HookPhase_HOOK_PHASE_BEFORE_SEND, EventType{})}
}

// BeforeSend 开始登记子插件的消息链改写 Hook。
func (s *SubPlugin) BeforeSend(routeID string) *BeforeSendBuilder {
	return &BeforeSendBuilder{route: s.scope.addHook(routeID, pb.HookPhase_HOOK_PHASE_BEFORE_SEND, EventType{})}
}

// BeforeSendBuilder 构建发送前的改链或拒绝 Hook。
type BeforeSendBuilder struct{ route *hookRoute }

// Priority 设置 Hook 优先级，数值越小越先执行。
func (b *BeforeSendBuilder) Priority(p int32) *BeforeSendBuilder {
	b.route.scope.plugin.checkRegistration()
	b.route.priority = p
	return b
}

// Handle 设置改链函数；返回 nil Chain 保留当前内容，返回错误拒绝本次操作。
func (b *BeforeSendBuilder) Handle(h func(c *Context, to message.Session, m message.Chain) (message.Chain, error)) {
	b.route.scope.plugin.checkRegistration()
	b.route.handles++
	b.route.handler = h
}

// AfterSend 开始注册发送结果观察 Hook。
func (p *Plugin) AfterSend(routeID string) *AfterSendBuilder {
	return &AfterSendBuilder{route: p.scope.addHook(routeID, pb.HookPhase_HOOK_PHASE_AFTER_SEND, EventType{})}
}

// AfterSend 开始登记子插件的发送结果观察 Hook。
func (s *SubPlugin) AfterSend(routeID string) *AfterSendBuilder {
	return &AfterSendBuilder{route: s.scope.addHook(routeID, pb.HookPhase_HOOK_PHASE_AFTER_SEND, EventType{})}
}

// AfterSendBuilder 构建观察最终发送结果的 Hook。
type AfterSendBuilder struct{ route *hookRoute }

// Priority 设置 Hook 优先级，数值越小越先执行。
func (b *AfterSendBuilder) Priority(p int32) *AfterSendBuilder {
	b.route.scope.plugin.checkRegistration()
	b.route.priority = p
	return b
}

// Handle 设置观察函数；其错误由 Core 记录，不改变原发送结果。
func (b *AfterSendBuilder) Handle(h func(c *Context, to message.Session, res *message.SendResult, sendErr error) error) {
	b.route.scope.plugin.checkRegistration()
	b.route.handles++
	b.route.handler = h
}

func collectHooks(root *Plugin, reg *Registration) error {
	seen := make(map[string]bool)
	for _, route := range buildRegister(reg, "{}", 0, nil).Routes {
		seen[route.GetRouteId()] = true
	}
	var errs []error
	for _, route := range root.hooks {
		expected := "func(*sdk.Context, message.Session, message.Chain) (message.Chain, error)"
		if route.phase == pb.HookPhase_HOOK_PHASE_AFTER_SEND {
			expected = "func(*sdk.Context, message.Session, *message.SendResult, error) error"
		} else if route.phase == pb.HookPhase_HOOK_PHASE_PRE_EVENT {
			expected = fmt.Sprintf("func(*sdk.Context, %s) (%s, error)", route.event.payload, route.event.payload)
			if err := route.event.check(); err != nil {
				errs = append(errs, fmt.Errorf("hook route %q: expected %s: %w", route.id, expected, err))
				continue
			}
		}
		t := reflect.TypeOf(route.handler)
		valid := route.handles == 1 && t != nil && t.Kind() == reflect.Func && !reflect.ValueOf(route.handler).IsNil()
		if valid && route.phase == pb.HookPhase_HOOK_PHASE_PRE_EVENT {
			valid = !t.IsVariadic() && t.NumIn() == 2 && t.In(0) == contextType && t.In(1) == route.event.payload && t.NumOut() == 2 && t.Out(0) == route.event.payload && t.Out(1) == errorType
		}
		if !valid {
			errs = append(errs, fmt.Errorf("hook route %q: expected exactly one Handle with %s; got %T (%d calls)", route.id, expected, route.handler, route.handles))
			continue
		}
		if route.id == "" || seen[route.id] {
			errs = append(errs, fmt.Errorf("hook route %q has empty or duplicate route_id", route.id))
			continue
		}
		seen[route.id] = true
		collected := *route
		if inherited := route.scope.commands.permission; inherited != nil {
			collected.permission = *inherited
		}
		if err := checkPermission(collected.permission); err != nil {
			errs = append(errs, fmt.Errorf("hook route %q: %w", route.id, err))
			continue
		}
		reg.hooks = append(reg.hooks, collected)
		reg.subPlugins[route.id] = route.scope.subID
	}
	return errors.Join(errs...)
}

func appendHookRoutes(reg *Registration, registration *pb.Register) {
	for _, route := range reg.hooks {
		spec := &pb.HookSpec{Phase: route.phase}
		if route.phase == pb.HookPhase_HOOK_PHASE_PRE_EVENT {
			spec.EventTypeFilter = []string{route.event.name}
		}
		registration.Routes = append(registration.Routes, &pb.RouteDescriptor{RouteId: route.id, Kind: pb.RouteKind_ROUTE_KIND_HOOK, Priority: route.priority, RequiredPermissions: route.permission.Names(), Spec: &pb.RouteDescriptor_Hook{Hook: spec}})
	}
}

type editingContextKey struct{}

// Editing 返回出站 Hook 本次编辑的消息句柄；发送与其他 Handler 返回空串。
func (c *Context) Editing() message.MessageRef {
	ref, _ := c.ctx.Value(editingContextKey{}).(message.MessageRef)
	return ref
}

func hookHandler(reg *Registration) func(context.Context, *wire.Inbound) (wire.Response, bool) {
	routes := make(map[string]hookRoute, len(reg.hooks))
	for _, route := range reg.hooks {
		routes[route.id] = route
	}
	return func(ctx context.Context, in *wire.Inbound) (wire.Response, bool) {
		route, ok := routes[in.Target]
		if !ok {
			return wire.Response{}, false
		}
		return invokeHook(ctx, in, route), true
	}
}

func invokeHook(ctx context.Context, in *wire.Inbound, route hookRoute) (response wire.Response) {
	// 1. 按 route_id 对应的阶段解码，并建立本次 Hook 的有效调用上下文。
	var request proto.Message
	switch route.phase {
	case pb.HookPhase_HOOK_PHASE_PRE_EVENT:
		request = &pb.PreEventRequest{}
	case pb.HookPhase_HOOK_PHASE_BEFORE_SEND:
		request = &pb.BeforeSendRequest{}
	case pb.HookPhase_HOOK_PHASE_AFTER_SEND:
		request = &pb.AfterSendRequest{}
	}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if request.(interface{ GetRouteId() string }).GetRouteId() != in.Target {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "hook route_id does not match target")
	}
	ctx, cancel := context.WithCancel(ctx)
	scope, _ := ctx.Value(pluginContextKey{}).(pluginScope)
	c := &Context{plugin: scope.plugin, subID: scope.subID, ctx: ctx, inbound: in}
	bindServiceContext(c)
	c.active.Store(true)
	defer func() {
		c.propagationMu.Lock()
		c.active.Store(false)
		c.propagationMu.Unlock()
		cancel()
		if p := recover(); p != nil {
			c.Logger().Error("hook panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
			response = dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "hook panicked")
		}
	}()
	// 2. 插件只回传当前阶段允许改写的数据；出站上下文包含本次编辑句柄。
	var result proto.Message
	var err error
	switch request := request.(type) {
	case *pb.PreEventRequest:
		c.event = request.Event
		if request.GetEvent().GetEventType() != route.event.name {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "pre-event type does not match registered filter")
		}
		out := &pb.PreEventResult{Propagation: pb.Propagation_PROPAGATION_CONTINUE}
		out.EventPayload, err = invokePreEvent(c, route)
		c.propagationMu.Lock()
		if c.stopped {
			out.Propagation = pb.Propagation_PROPAGATION_STOP
		}
		c.active.Store(false)
		c.propagationMu.Unlock()
		result = out
	case *pb.BeforeSendRequest:
		bindOutboundHook(c, request.Context, request.Operation, request.Target, request.GetEditedPlatformMessageId())
		var chain message.Chain
		chain, err = route.handler.(func(*Context, message.Session, message.Chain) (message.Chain, error))(c, convert.SessionFromProto(request.Target), convert.ChainFromProto(request.Chain, request.Target))
		out := &pb.BeforeSendResult{}
		if err == nil && chain != nil {
			out.Chain, err = convert.ChainToProto(chain, request.Target)
		}
		result = out
	case *pb.AfterSendRequest:
		bindOutboundHook(c, request.Context, request.Operation, request.Target, request.GetEditedPlatformMessageId())
		var sendErr error
		if request.Error != nil {
			sendErr = convert.ErrorFromProto(request.Error)
		}
		err = route.handler.(func(*Context, message.Session, *message.SendResult, error) error)(c, convert.SessionFromProto(request.Target), convert.SendResultFromProto(request.Result), sendErr)
		result = &pb.AfterSendResult{}
	}
	if err != nil {
		var public *Error
		if errors.As(err, &public) {
			return wire.Response{Err: convert.ErrorToProto(public)}
		}
		c.Logger().Error("hook returned an error", "error", err)
		return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "hook returned an error")
	}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode hook result: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: invocationSchemaVersion, Bytes: encoded}}
}

func bindOutboundHook(c *Context, invocation *pb.InvocationContext, op pb.SendOperation, target *pb.Target, edited string) {
	c.actor = invocation.GetActor()
	c.defaultTarget = invocation.GetDefaultTarget()
	if op == pb.SendOperation_SEND_OPERATION_EDIT {
		c.ctx = context.WithValue(c.ctx, editingContextKey{}, message.NewMessageRef(convert.SessionFromProto(target), edited, target.GetAdapterInstanceId()))
	}
}

func invokePreEvent(c *Context, route hookRoute) (*pb.EventPayload, error) {
	var payload reflect.Value
	if route.event.adapter {
		raw, ok := c.Event().Payload.(*message.AdapterEventPayload)
		if !ok || raw.Version != route.event.version {
			return nil, fmt.Errorf("adapter event payload does not match registered version %d", route.event.version)
		}
		payload = reflect.New(route.event.payload.Elem())
		if err := json.Unmarshal([]byte(raw.PayloadJSON), payload.Interface()); err != nil {
			return nil, fmt.Errorf("decode adapter event payload: %w", err)
		}
	} else {
		payload = reflect.ValueOf(c.Event().Payload)
		if !payload.IsValid() || payload.Type() != route.event.payload || payload.IsNil() {
			return nil, errors.New("pre-event payload type mismatch")
		}
	}
	values := reflect.ValueOf(route.handler).Call([]reflect.Value{reflect.ValueOf(c), payload})
	if !values[1].IsNil() {
		return nil, values[1].Interface().(error)
	}
	if values[0].IsNil() {
		return nil, nil
	}
	var rewritten message.Payload
	if route.event.adapter {
		encoded, err := json.Marshal(values[0].Interface())
		if err != nil {
			return nil, fmt.Errorf("encode adapter event payload: %w", err)
		}
		rewritten = &message.AdapterEventPayload{Version: route.event.version, PayloadJSON: string(encoded)}
	} else {
		rewritten = values[0].Interface().(message.Payload)
	}
	event, err := convert.EventToProto(&message.Event{Type: route.event.name, Session: c.Session(), AdapterInstanceID: c.event.GetAdapterInstanceId(), Payload: rewritten})
	if err != nil {
		return nil, err
	}
	return event.EventPayload, nil
}
