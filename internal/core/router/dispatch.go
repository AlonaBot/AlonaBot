package router

import (
	"context"
	"errors"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Options 提供事件分发限额与现有的路由依赖；限额使用 tune.Values 对应值。
type Options struct {
	Concurrency int                                                                           // DISPATCH_CONCURRENCY
	QueueEvents int                                                                           // DISPATCH_QUEUE_EVENTS
	QueueBytes  int64                                                                         // DISPATCH_QUEUE_BYTES
	Global      *wire.Budget                                                                  // 排队事件占用的 Core 全局预算；nil 时不限额
	Adapters    *adapter.Registry                                                             // 账号用户名查询
	Send        ServiceFunc                                                                   // usage 与自动回复发送
	Prepare     func(*pb.Event, Caller) (*pb.Event, error)                                    // 按接收 Session 准备媒体；nil 时原样派发
	Prefixes    func(adapterInstanceID string) []string                                       // nil 时所有实例使用 ["/"]
	Buttons     func(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error) // nil 时丢弃按钮点击
	Policy      CommandPolicy                                                                 // nil 时仅按声明权限判定且不赋予 admin
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
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return errors.New("event router is closed")
	}
	n := int64(proto.Size(ev))
	if len(r.queue) >= r.queueEvents || r.queuedBytes+n > r.queueBytes {
		r.stats.DroppedQueueFull++
		return adapter.ErrQueueFull
	}
	if !r.global.Acquire(n) {
		r.stats.DroppedBudget++
		return adapter.ErrQueueFull
	}
	r.queue = append(r.queue, queuedEvent{event: retainEvent(ev), bytes: n})
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
	r.dispatch(event)
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
		r.global.Release(r.queuedBytes)
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
