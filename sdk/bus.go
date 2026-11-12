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
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"google.golang.org/protobuf/proto"
)

// SubscriptionRegistry 收集静态 Event Bus 订阅；零值可用，Collect 统一返回描述错误。
type SubscriptionRegistry struct {
	pending       []func() error
	subscriptions []subscriptionRoute
	errs          []error
}

type subscriptionRoute struct {
	descriptor *pb.RouteDescriptor
	invoke     func(*Context, []byte) error
}

// Subscription 构建一条类型化订阅；Handle 结束声明。
type Subscription[T any] struct {
	routeID   string
	eventName string
	handles   int
	handler   func(*Context, T) error
}

// Subscribe 开始构建 eventName（plugin_id/event_name）的 JSON 订阅。
// 回调按 Session 串行执行，Context 无 Actor 与默认 Target。
func Subscribe[T any](r *SubscriptionRegistry, routeID, eventName string) *Subscription[T] {
	s := &Subscription[T]{routeID: routeID, eventName: eventName}
	r.pending = append(r.pending, func() error {
		if s.handles != 1 {
			return fmt.Errorf("subscription route %q has invalid Handle sequence", s.routeID)
		}
		if s.handler == nil {
			return fmt.Errorf("handler for subscription route %q is nil", s.routeID)
		}
		handler := s.handler // 收集后冻结 Handler，运行时注册描述保持固定。
		id, version, err := jsonType[T]()
		if err != nil {
			return fmt.Errorf("subscription route %q: %w", s.routeID, err)
		}
		r.subscriptions = append(r.subscriptions, subscriptionRoute{descriptor: &pb.RouteDescriptor{
			RouteId: s.routeID, Kind: pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION, DefaultEnabled: true,
			Spec: &pb.RouteDescriptor_BusSubscription{BusSubscription: &pb.BusSubscriptionSpec{EventName: s.eventName, SchemaId: id, SchemaVersion: version, CallbackRouteId: s.routeID}},
		}, invoke: func(c *Context, payload []byte) error {
			var input T
			if err := json.Unmarshal(payload, &input); err != nil {
				c.Logger().Warn("decode bus event payload failed", "error", err)
				return nil
			}
			return handler(c, input)
		}})
		return nil
	})
	return s
}

// Handle 结束订阅声明，签名由类型参数约束。
func (s *Subscription[T]) Handle(h func(*Context, T) error) { s.handles++; s.handler = h }

func collectSubscriptions(root Root, reg *Registration) error {
	r := &SubscriptionRegistry{}
	root.Subscriptions(r)
	for _, finish := range r.pending {
		r.errs = append(r.errs, finish())
	}
	reg.subscriptions = r.subscriptions
	if len(r.subscriptions) == 0 {
		return errors.Join(r.errs...)
	}
	registration := &pb.Register{CandidateConfigJson: "{}"}
	appendSubscriptionRoutes(reg, registration)
	r.errs = append(r.errs, validate.Register(registration))
	// 复用实际 REGISTER 的 route_id 生成规则，核对与其他注册器的冲突；HTTP Listener 此时尚未创建。
	seen := make(map[string]bool)
	for _, route := range buildRegister(reg, "{}", 0, nil).GetRoutes() {
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

// Publish 从具名载荷类型推导 Schema 标识与版本，再发布到本插件的 eventName。
// SDK 补上 plugin_id/ 前缀。
// 返回值只确认 Core 已完成发布准入。Handler 内外均可调用；ctx 只传递取消和期限，不继承用户权限。
func Publish[T any](ctx context.Context, eventName string, payload T) (eventID string, err error) {
	schemaID, version, err := jsonType[T]()
	if err != nil {
		return "", err
	}
	// 1. 复用 Session ready gate，以当前插件身份编码 JSON 和发布请求。
	session, err := currentSession()
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", &Error{Code: errcode.InvalidArgument, Message: "encode bus event payload: " + err.Error()}
	}
	request := &pb.PublishRequest{EventName: activeSession.Load().l.pluginID + "/" + eventName, SchemaId: schemaID, SchemaVersion: version, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Payload: encoded}
	encoded, err = proto.Marshal(request)
	if err != nil {
		return "", &Error{Code: errcode.InvalidArgument, Message: "encode publish request: " + err.Error()}
	}
	// 2. 独立请求不设置 parent_call_id、trace_id 或可信调用附件。
	response, err := session.Call(ctx, wire.Request{Target: wire.ServiceBusPublish,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	if err != nil {
		message := err.Error()
		var callError *wire.CallError
		if errors.As(err, &callError) {
			message = callError.ErrorInfo.GetSafeMessage()
		}
		return "", &Error{Code: errcode.Code(validate.PublicCode(wire.CodeOf(err))), Message: message}
	}
	result := &pb.PublishResponse{}
	if err := decodeInvocation(response.Payload, result); err != nil {
		return "", &Error{Code: errcode.Internal, Message: err.Error()}
	}
	return result.GetEventId(), nil
}

type subscriptionDelivery struct {
	env     *pb.Envelope
	release func() // 由 wire 交付的预算所有权，处理结束或丢弃时归还
}

// subscriptionEvents 为本次 Session 创建唯一的订阅投递 goroutine，接收端只做门禁与入队。
// 排队及正在处理的载荷合计最多占 BUS_QUEUE_BYTES 的缺省值，满时 drop-newest 并立即归还 wire 预算。
func subscriptionEvents(a *activation, reg *Registration) func(*pb.Envelope, func()) {
	routes := make(map[string]subscriptionRoute, len(reg.subscriptions))
	for _, sub := range reg.subscriptions {
		routes[sub.descriptor.GetBusSubscription().GetCallbackRouteId()] = sub
	}
	var mu sync.Mutex // 保护 queue、bytes 与 closed，串行交接预算所有权
	var bytes int64   // 包含正在执行回调的载荷，防止慢回调占满共享接收预算
	maxBytes := tune.Defaults().BusQueueBytes
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
		n := int64(len(env.GetPayload()))
		if bytes+n > maxBytes {
			mu.Unlock()
			slog.Warn("bus event dropped because subscription payload budget is full", "route_id", env.GetTarget())
			release()
			return
		}
		bytes += n
		queue = append(queue, subscriptionDelivery{env: env, release: func() {
			mu.Lock()
			bytes -= n
			mu.Unlock()
			release()
		}})
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
