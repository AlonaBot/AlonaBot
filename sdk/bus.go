package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// SubscriptionRegistry 收集静态 Event Bus 订阅；零值可用，Collect 统一返回描述错误。
type SubscriptionRegistry struct {
	subscriptions []subscriptionRoute
	errs          []error
}

type subscriptionRoute struct {
	descriptor *pb.RouteDescriptor
	invoke     func(*Context, []byte) error
}

// Subscribe 注册 eventName（plugin_id/event_name）的类型化 JSON 订阅，以 routeID 标识回调。
// 回调按 Session 串行执行，Context 无 Actor 与默认 Target；错误和 panic 只写日志，无业务回执。
func Subscribe[T any](r *SubscriptionRegistry, routeID, eventName, schemaID string, version uint32, h func(*Context, T) error) {
	if h == nil {
		r.errs = append(r.errs, fmt.Errorf("handler for subscription %q is nil", routeID))
		return
	}
	r.subscriptions = append(r.subscriptions, subscriptionRoute{descriptor: &pb.RouteDescriptor{
		RouteId: routeID, Kind: pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION, DefaultEnabled: true, RequiredPermission: pb.Permission_PERMISSION_ANY,
		Spec: &pb.RouteDescriptor_BusSubscription{BusSubscription: &pb.BusSubscriptionSpec{EventName: eventName, SchemaId: schemaID, SchemaVersion: version, CallbackRouteId: routeID}},
	}, invoke: func(c *Context, payload []byte) error {
		var input T
		if err := json.Unmarshal(payload, &input); err != nil {
			c.Logger().Warn("decode bus event payload failed", "error", err)
			return nil
		}
		return h(c, input)
	}})
}

func collectSubscriptions(root Root, reg *Registration) error {
	r := &SubscriptionRegistry{}
	root.Subscriptions(r)
	reg.subscriptions = r.subscriptions
	if len(r.subscriptions) == 0 {
		return errors.Join(r.errs...)
	}
	registration := &pb.Register{CandidateConfigJson: "{}"}
	appendSubscriptionRoutes(reg, registration)
	r.errs = append(r.errs, validate.Register(registration))
	// 复用实际 REGISTER 的 route_id 生成规则，核对与其他注册器的冲突；HTTP Listener 此时尚未创建。
	seen := make(map[string]bool)
	for _, route := range buildRegister(reg, "{}", 0).GetRoutes() {
		if route.GetKind() != pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION {
			seen[route.GetRouteId()] = true
		}
	}
	for _, sub := range r.subscriptions {
		if seen[sub.descriptor.GetRouteId()] {
			r.errs = append(r.errs, fmt.Errorf("subscription route %q conflicts with another route", sub.descriptor.GetRouteId()))
		}
	}
	return errors.Join(r.errs...)
}

func appendSubscriptionRoutes(reg *Registration, registration *pb.Register) {
	for _, sub := range reg.subscriptions {
		registration.Routes = append(registration.Routes, proto.Clone(sub.descriptor).(*pb.RouteDescriptor))
	}
}

// Publish 将 JSON 载荷发布到本插件的 eventName，SDK 补上 plugin_id/ 前缀。
// 返回值只确认 Core 已完成发布准入。Handler 内外均可调用；ctx 只传递取消和期限，不继承用户权限。
func Publish(ctx context.Context, eventName, schemaID string, version uint32, payload any) (eventID string, err error) {
	// 1. 复用 Session ready gate，以当前插件身份编码 JSON 和发布请求。
	session, err := currentSession()
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode bus event payload: " + err.Error()}
	}
	request := &pb.PublishRequest{EventName: activeSession.Load().l.pluginID + "/" + eventName, SchemaId: schemaID, SchemaVersion: version, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Payload: encoded}
	encoded, err = proto.Marshal(request)
	if err != nil {
		return "", &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode publish request: " + err.Error()}
	}
	// 2. 独立请求不设置 parent_call_id、trace_id 或可信调用附件。
	response, err := session.Call(ctx, wire.Request{Target: wire.ServiceBusPublish,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	if err != nil {
		info := err.(*wire.CallError).ErrorInfo
		return "", &Error{Code: validate.PublicCode(info.GetCode()), Message: info.GetSafeMessage()}
	}
	result := &pb.PublishResponse{}
	if err := decodeInvocation(response.Payload, result); err != nil {
		return "", &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: err.Error()}
	}
	return result.GetEventId(), nil
}

type subscriptionDelivery struct {
	env     *pb.Envelope
	release func() // 由 wire 交付的预算所有权，处理结束或丢弃时归还
}

// subscriptionEvents 为本次 Session 创建唯一的订阅投递 goroutine，接收端只做门禁与入队。
// 队列载荷仍占 wire 的接收预算，不另设容量；慢回调不会占用读循环。
func subscriptionEvents(a *activation, reg *Registration) func(*pb.Envelope, func()) {
	routes := make(map[string]subscriptionRoute, len(reg.subscriptions))
	for _, sub := range reg.subscriptions {
		routes[sub.descriptor.GetBusSubscription().GetCallbackRouteId()] = sub
	}
	var mu sync.Mutex // 保护 queue 与 closed，串行交接预算所有权
	var queue []subscriptionDelivery
	closed := false
	wake := make(chan struct{}, 1)
	go func() {
		defer func() {
			mu.Lock()
			closed = true
			pending := queue
			queue = nil
			mu.Unlock()
			for _, delivery := range pending {
				delivery.release()
			}
		}()
		for {
			mu.Lock()
			if a.lifetime.Err() != nil {
				mu.Unlock()
				return
			}
			if len(queue) == 0 {
				mu.Unlock()
				select {
				case <-wake:
				case <-a.lifetime.Done():
					return
				}
				continue
			}
			delivery := queue[0]
			queue[0] = subscriptionDelivery{}
			queue = queue[1:]
			mu.Unlock()
			invokeSubscription(a, routes, delivery)
		}
	}()
	return func(env *pb.Envelope, release func()) {
		// 与 ACTIVATED 写出及 ready 发布串行化，避免首条 EVENT 被本地门禁误丢弃。
		a.readyMu.Lock()
		ready := a.ready.Load() && !a.stopping.Load()
		a.readyMu.Unlock()
		if !ready {
			slog.Warn("bus event dropped while plugin is not ready", "route_id", env.GetTarget())
			release()
			return
		}
		mu.Lock()
		if closed {
			mu.Unlock()
			slog.Warn("bus event dropped after subscription dispatcher stopped", "route_id", env.GetTarget())
			release()
			return
		}
		queue = append(queue, subscriptionDelivery{env: env, release: release})
		mu.Unlock()
		// 唤醒信号合并，接收路径不等待用户回调。
		select {
		case wake <- struct{}{}:
		default:
		}
	}
}

func invokeSubscription(a *activation, routes map[string]subscriptionRoute, delivery subscriptionDelivery) {
	defer delivery.release()
	env := delivery.env
	if !a.ready.Load() || a.stopping.Load() || a.lifetime.Err() != nil {
		slog.Warn("bus event dropped while plugin is stopping", "route_id", env.GetTarget())
		return
	}
	// 1. 校验实际回调与事件声明，失败仅记日志，不产生 RESPONSE。
	sub, ok := routes[env.GetTarget()]
	if !ok {
		slog.Warn("bus subscription route not found", "route_id", env.GetTarget())
		return
	}
	ev := &pb.BusEvent{}
	if err := proto.Unmarshal(env.GetPayload(), ev); err != nil {
		slog.Warn("decode BusEvent failed", "route_id", env.GetTarget(), "error", err)
		return
	}
	spec := sub.descriptor.GetBusSubscription()
	if ev.GetEventName() != spec.GetEventName() || ev.GetSchemaId() != spec.GetSchemaId() || ev.GetSchemaVersion() != spec.GetSchemaVersion() || ev.GetEncoding() != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || !utf8.Valid(ev.GetPayload()) {
		slog.Warn("bus event does not match registered JSON subscription", "route_id", env.GetTarget(), "event_name", ev.GetEventName())
		return
	}
	// 2. 回调没有入站 Event/Actor、默认 Target 或父调用编号，后续调用仅带订阅插件身份。
	ctx, cancel := context.WithCancel(a.lifetime)
	c := &Context{ctx: ctx, inbound: &wire.Inbound{Target: env.GetTarget()}}
	bindServiceContext(c)
	c.active.Store(true)
	defer func() {
		c.active.Store(false)
		cancel()
		if p := recover(); p != nil {
			c.Logger().Error("subscription callback panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
		}
	}()
	if err := sub.invoke(c, ev.GetPayload()); err != nil {
		c.Logger().Error("subscription callback returned an error", "error", err)
	}
}
