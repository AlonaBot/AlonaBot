package router

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/metrics"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Options 提供事件分发限额与现有的路由依赖；限额使用 tune.Values 对应值。
type Options struct {
	Dispatched func(Dispatch)                                                                           // nil 时不记录分发结果；仅库模式测试套件装配。
	PreEvent   func(ctx context.Context, ev *pb.Event) (payload *pb.EventPayload, stop bool, err error) // nil 时跳过；返回载荷替换业务副本
	// MessageConfig 返回本次消息会话的生效配置快照；错误供诊断，须省略配置值。
	MessageConfig  func(ctx context.Context, pluginID string, event *pb.Event) (json.RawMessage, error)
	EnsureActive   func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)
	RequestTimeout time.Duration                                                                 // 现有 REQUEST_TIMEOUT，从事件接纳时刻起算。
	Concurrency    int                                                                           // DISPATCH_CONCURRENCY
	QueueEvents    int                                                                           // DISPATCH_QUEUE_EVENTS
	QueueBytes     int64                                                                         // DISPATCH_QUEUE_BYTES
	Global         *wire.Budget                                                                  // 排队事件占用的 Core 全局预算；nil 时不限额
	Adapters       *adapter.Registry                                                             // 账号用户名查询
	Send           ServiceFunc                                                                   // usage 与自动回复发送
	Prepare        func(*pb.Event, Caller) (*pb.Event, error)                                    // 按接收 Session 准备媒体；nil 时原样派发
	Prefixes       func(adapterInstanceID string) []string                                       // nil 时所有实例使用 ["/"]
	Buttons        func(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error) // nil 时丢弃按钮点击
	Intercept      func(ctx context.Context, ev *pb.Event) (handled bool)                        // 入队前在锁外消费 Core 事件；nil 时跳过
	Policy         CommandPolicy                                                                 // nil 时仅按声明权限判定且不赋予 admin
	CommandState   func(context.Context, string, *pb.RouteDescriptor) (bool, []string, error)    // 菜单读取完整策略；nil 时使用注册描述。
	CommandNames   func(pluginID string) map[string][][]string                                   // nil 时使用声明；在 Router 锁内调用，不得做 I/O。
	Continuations  Continuations                                                                 // nil 时跳过延续与等待。
	Metrics        *metrics.Counters                                                             // nil 时关闭原因计数；全局接纳使用空插件 ID。
}

// Continuations 接入 Handler 结束通知、入站等待匹配及插件释放。
type Continuations interface {
	Finished(inv *Invocation, ok bool)
	Match(ev *pb.Event) bool
	Release(c Caller)
}

// Stats 记录本次 Router 运行的准入计数与当前链、等待事件数。
type Stats struct {
	Accepted, DroppedQueueFull, DroppedBudget uint64
	Running, Queued                           int
}

type queuedEvent struct {
	event *trustedEvent
	bytes int64
}

// Stats 取得准入计数与运行状态的同一时刻快照。
func (r *Router) Stats() Stats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats
}

// Start 开放队列派发；重复启动及关闭后的启动不产生新链。
func (r *Router) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started || r.closed {
		return
	}
	r.started = true
	r.startLocked()
}

// Deliver 立即尝试接纳事件，不等待空闲链；条数、字节或全局预算不足时返回 ErrQueueFull。
func (r *Router) Deliver(ctx context.Context, ev *pb.Event) error {
	if err := ctx.Err(); err != nil {
		r.metrics.Add("", "event", "cancelled")
		return err
	}
	if r.intercept != nil && ev.GetEventType() != commandInvokedEvent {
		r.mu.Lock()
		closed := r.closed
		r.mu.Unlock()
		if closed {
			return errors.New("event router is closed")
		}
		// 确认解析会查询插件身份，必须在锁外执行且无需等待派发槽。
		if r.intercept(ctx, ev) {
			return nil
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		r.metrics.Add("", "event", "offline")
		return errors.New("event router is closed")
	}
	n := int64(proto.Size(ev))
	if len(r.queue) >= r.queueEvents || r.queuedBytes+n > r.queueBytes {
		r.stats.DroppedQueueFull++
		reason := metrics.Reason("dropped_queue")
		if len(r.queue) < r.queueEvents {
			reason = "dropped_budget"
		}
		r.metrics.Add("", "event", reason)
		return adapter.ErrQueueFull
	}
	if !r.global.Acquire(n) {
		r.stats.DroppedBudget++
		r.metrics.Add("", "event", "dropped_budget")
		return adapter.ErrQueueFull
	}
	event := retainEvent(ev)
	if r.requestTimeout > 0 {
		event.deadline = time.Now().Add(r.requestTimeout)
	}
	if deadline, ok := ctx.Deadline(); ok && (event.deadline.IsZero() || deadline.Before(event.deadline)) {
		event.deadline = deadline
	}
	r.queue = append(r.queue, queuedEvent{event: event, bytes: n})
	r.queuedBytes += n
	r.stats.Accepted++
	r.stats.Queued++
	r.startLocked()
	return nil
}

// startLocked 在 mu 下按 FIFO 交接事件；出队即释放等待预算，链在锁外执行。
func (r *Router) startLocked() {
	for r.started && !r.closed && r.stats.Running < r.concurrency && len(r.queue) > 0 {
		next := r.queue[0]
		r.queue[0] = queuedEvent{} // 释放已出队快照，避免底层切片保留事件。
		r.queue = r.queue[1:]
		r.queuedBytes -= next.bytes
		r.global.Release(next.bytes)
		r.stats.Queued--
		r.stats.Running++
		go r.run(next.event) // Session 读循环和接纳入口都不执行 Handler。
	}
}

func (r *Router) run(event *trustedEvent) {
	r.observe(event)
	r.dispatch(event)
	r.completeDispatch(event.dispatch)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Running--
	if r.closed && r.stats.Running == 0 {
		close(r.drained) // 只有最后一条运行链通知 Close。
	}
	r.startLocked()
}

// Close 停止接纳并释放排队事件，等待运行链；ctx 到期则取消调用并等链退出，返回 ctx.Err()。
func (r *Router) Close(ctx context.Context) error {
	// 1. 在接纳锁下丢弃等待事件，并记录本次调用是否需要等运行链。
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		for id, state := range r.commandSyncs {
			state.cancel()
			delete(r.commandSyncs, id)
		}
		r.global.Release(r.queuedBytes)
		for range r.queue {
			r.metrics.Add("", "event", "offline")
		}
		r.queue = nil
		r.queuedBytes = 0
		r.stats.Queued = 0
		if r.stats.Running == 0 {
			close(r.drained)
		}
	}
	drained := r.stats.Running == 0
	r.mu.Unlock()
	if drained {
		r.cancel()
		return nil
	}
	// 2. 等待期间保留在途关联；排空与取消同时就绪时仍保留 ctx 错误。
	select {
	case <-r.drained:
		r.cancel()
		return ctx.Err()
	case <-ctx.Done():
		r.cancel()
		<-r.drained // 取消后仍等链退出，归还运行槽与在途关联。
		return ctx.Err()
	}
}
