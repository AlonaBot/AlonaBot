package sdk

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"google.golang.org/protobuf/proto"
)

type subscriptionRegistry = registry.Subscriptions[*Context]
type subscriptionRoute = registry.Subscription[*Context]

func collectSubscriptions(root *Plugin, reg *registration) error {
	r := &root.subscriptions
	r.Subscriptions = nil
	r.Errs = nil
	for _, finish := range r.Pending {
		r.Errs = append(r.Errs, finish())
	}
	reg.subscriptions = r.Subscriptions
	registration := &pb.Register{CandidateConfigJson: "{}"}
	appendSubscriptionRoutes(reg, registration)
	r.Errs = append(r.Errs, validate.Register(registration))
	return errors.Join(r.Errs...)
}

func appendSubscriptionRoutes(reg *registration, registration *pb.Register) {
	for _, sub := range reg.subscriptions {
		registration.Routes = append(registration.Routes, proto.Clone(sub.Descriptor).(*pb.RouteDescriptor))
	}
}

type subscriptionDelivery struct {
	env     *pb.Envelope
	release func() // 由 wire 交付的预算所有权，处理结束或丢弃时归还
}

// subscriptionEvents 为本次 Session 创建唯一的订阅投递 goroutine，接收端只做门禁与入队。
// 排队及正在处理的载荷合计最多占 BUS_QUEUE_BYTES 的缺省值，满时 drop-newest 并立即归还 wire 预算。
func subscriptionEvents(a *activation, reg *registration) func(*pb.Envelope, func()) {
	routes := make(map[string]subscriptionRoute, len(reg.subscriptions))
	for _, sub := range reg.subscriptions {
		routes[sub.Descriptor.GetBusSubscription().GetCallbackRouteId()] = sub
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
			componentLogger(a.lifetime).Warn(fmt.Sprintf("Dropped a bus event for route %s because the component is not ready", env.GetTarget()), "route_id", env.GetTarget())
			release()
			return
		}
		mu.Lock()
		if closed {
			mu.Unlock()
			componentLogger(a.lifetime).Warn(fmt.Sprintf("Dropped a bus event for route %s because the subscription dispatcher stopped", env.GetTarget()), "route_id", env.GetTarget())
			release()
			return
		}
		n := int64(len(env.GetPayload()))
		if bytes+n > maxBytes {
			mu.Unlock()
			componentLogger(a.lifetime).Warn(fmt.Sprintf("Dropped a bus event for route %s because the subscription payload budget is full", env.GetTarget()), "route_id", env.GetTarget())
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
	if !a.l.runtime.beginCallback() {
		return
	}
	defer a.l.runtime.callbacks.Done()
	env := delivery.env
	if !a.ready.Load() || a.stopping.Load() || a.lifetime.Err() != nil {
		componentLogger(a.lifetime).Warn(fmt.Sprintf("Dropped a bus event for route %s because the component is stopping", env.GetTarget()), "route_id", env.GetTarget())
		return
	}
	// 1. 校验实际回调与事件声明，失败仅记日志，不产生 RESPONSE。
	sub, ok := routes[env.GetTarget()]
	if !ok {
		componentLogger(a.lifetime).Warn(fmt.Sprintf("Discarded a bus event because subscription route %s is not registered", env.GetTarget()), "route_id", env.GetTarget())
		return
	}
	ev := &pb.BusEvent{}
	if err := proto.Unmarshal(env.GetPayload(), ev); err != nil {
		componentLogger(a.lifetime).Warn(fmt.Sprintf("Failed to decode a bus event for route %s: %v", env.GetTarget(), err), "route_id", env.GetTarget(), "error", err)
		return
	}
	spec := sub.Descriptor.GetBusSubscription()
	if ev.GetEventName() != spec.GetEventName() || ev.GetSchemaId() != spec.GetSchemaId() || ev.GetSchemaVersion() != spec.GetSchemaVersion() || ev.GetEncoding() != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || !utf8.Valid(ev.GetPayload()) {
		componentLogger(a.lifetime).Warn(fmt.Sprintf("Bus event %s does not match the registered JSON subscription for route %s", ev.GetEventName(), env.GetTarget()), "route_id", env.GetTarget(), "event_name", ev.GetEventName())
		return
	}
	// 2. 回调没有入站 Event/Actor、默认 Target 或父调用编号，后续调用仅带订阅插件身份。
	ctx, cancel := context.WithCancel(a.lifetime)
	c := &Context{plugin: sub.Plugin.(*Plugin), subID: sub.SubID, ctx: ctx, inbound: &wire.Inbound{Target: env.GetTarget()}}
	c.freezeProfile(nil)
	bindServiceContext(c)
	c.active.Store(true)
	defer func() {
		c.active.Store(false)
		cancel()
		if p := recover(); p != nil {
			c.Logger().Error("subscription callback panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
		}
	}()
	if err := sub.Invoke(c, ev.GetPayload()); err != nil {
		c.Logger().Error("subscription callback returned an error", "error", err)
	}
}
