// Package clock 为调度与会话等待提供一致的当前时间和定时器。
package clock

import (
	"context"
	"sync"
	"time"
)

// Timer 保留单次定时器的停止与到期通知。
type Timer interface {
	Stop() bool
	C() <-chan time.Time
}

// Clock 同时拥有当前时间、回调定时器与通道定时器。
type Clock interface {
	Now() time.Time
	AfterFunc(time.Duration, func()) Timer
	NewTimer(time.Duration) Timer
}

// Real 使用 Go 运行时的真实时间。
type Real struct{}

// Now 返回真实时间。
func (Real) Now() time.Time { return time.Now() }

// AfterFunc 创建真实回调定时器。
func (Real) AfterFunc(d time.Duration, f func()) Timer { return realTimer{time.AfterFunc(d, f)} }

// NewTimer 创建真实通知定时器。
func (Real) NewTimer(d time.Duration) Timer { return realTimer{time.NewTimer(d)} }

type realTimer struct{ *time.Timer }

func (t realTimer) C() <-chan time.Time { return t.Timer.C }

// Manual 持有测试时间与全部定时器；Advance 在定时器表锁外执行到期回调。
type Manual struct {
	advance sync.Mutex // 串行化含回调的完整推进，避免旧目标覆盖新时间。
	mu      sync.Mutex
	now     time.Time
	next    uint64
	timers  map[uint64]*manualTimer
	closed  bool
}
type manualTimer struct {
	clock *Manual
	id    uint64
	at    time.Time
	f     func()
	ch    chan time.Time
}

// New 创建从 start 开始的可推进时钟。
func New(start time.Time) *Manual { return &Manual{now: start, timers: map[uint64]*manualTimer{}} }

// Now 返回当前测试时间。
func (c *Manual) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }

// AfterFunc 登记一次到期回调。
func (c *Manual) AfterFunc(d time.Duration, f func()) Timer { return c.add(d, f, nil) }

// NewTimer 登记一个带缓冲的单次到期通知。
func (c *Manual) NewTimer(d time.Duration) Timer { return c.add(d, nil, make(chan time.Time, 1)) }
func (c *Manual) add(d time.Duration, f func(), ch chan time.Time) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.next++
	t := &manualTimer{clock: c, id: c.next, at: c.now.Add(d), f: f, ch: ch}
	if !c.closed {
		if d <= 0 {
			if f != nil {
				go f()
			} else {
				ch <- c.now
			}
		} else {
			c.timers[t.id] = t
		}
	}
	return t
}
func (t *manualTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	_, ok := t.clock.timers[t.id]
	delete(t.clock.timers, t.id)
	return ok
}
func (t *manualTimer) C() <-chan time.Time { return t.ch }

// Advance 推进时间，按到期时间与登记顺序执行截至目标时刻的定时器。
// 回调可以读时钟或操作定时器；正在执行的回调须返回后才能继续推进。
func (c *Manual) Advance(d time.Duration) {
	if d < 0 {
		panic("clock cannot move backwards")
	}
	c.advance.Lock()
	defer c.advance.Unlock()
	c.mu.Lock()
	end := c.now.Add(d)
	for {
		var next *manualTimer
		for _, timer := range c.timers {
			if !timer.at.After(end) && (next == nil || timer.at.Before(next.at) || timer.at.Equal(next.at) && timer.id < next.id) {
				next = timer
			}
		}
		if next == nil {
			c.now = end
			c.mu.Unlock()
			return
		}
		delete(c.timers, next.id)
		c.now = next.at
		c.mu.Unlock()
		if next.f != nil {
			next.f()
		} else {
			next.ch <- next.at
		}
		c.mu.Lock()
	}
}

// Close 注销未到期定时器；生命周期取消由各定时器的调用方负责。
func (c *Manual) Close() { c.mu.Lock(); defer c.mu.Unlock(); c.closed = true; clear(c.timers) }

type deadlineContext struct {
	context.Context
	deadline time.Time
}

func (c deadlineContext) Deadline() (time.Time, bool) { return c.deadline, true }
func (c deadlineContext) Err() error {
	if context.Cause(c.Context) == context.DeadlineExceeded {
		return context.DeadlineExceeded
	}
	return c.Context.Err()
}

// WithDeadline 使用同一时钟取消 context，保留 DeadlineExceeded 与父调用取消的区别。
func WithDeadline(parent context.Context, c Clock, deadline time.Time) (context.Context, context.CancelFunc) {
	// 协议期限使用真实绝对时间；本地到期由同一逻辑时钟取消并传播给 Session。
	duration := deadline.Sub(c.Now())
	if _, real := c.(Real); !real {
		deadline = time.Now().Add(duration)
	}
	if previous, ok := parent.Deadline(); ok && previous.Before(deadline) {
		deadline = previous
		duration = time.Until(previous)
	}
	base, cancel := context.WithCancelCause(parent)
	timer := c.AfterFunc(duration, func() { cancel(context.DeadlineExceeded) })
	stopped := context.AfterFunc(base, func() { timer.Stop() })
	return deadlineContext{base, deadline}, func() { timer.Stop(); stopped(); cancel(context.Canceled) }
}
