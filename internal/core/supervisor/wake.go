package supervisor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// WakeOptions 绑定当前 Core 的状态、共享预算与完整启动入口。
type WakeOptions struct {
	Store         *state.Store
	Global        *wire.Budget
	QueueRequests int
	QueueBytes    int64
	Start         func(context.Context, string) (*Running, error) // 使用 Core 生命周期，返回前完成全部业务绑定。
}

// Waker 合并逐插件的启动，并唯一持有等待预算与当前 Session 的活动记录。
// mu 同时保护接纳与停止门禁；启动、数据库访问和等待均在锁外进行。
// 旧 Session 的释放闭包始终指向自己的记录，替换绑定后也不操作新记录。
type Waker struct {
	opt     WakeOptions
	mu      sync.Mutex
	plugins map[string]*wakePlugin
}

type wakePlugin struct {
	changed           chan struct{} // 状态改变后关闭再替换，广播给全部等待者。
	current           *activitySession
	waking            bool
	failure           error
	coldStarts        []time.Time
	coldStartSequence uint64 // 实际启动总序号，历史回收后保持单调。
	count             int
	bytes             int64
}

// NewWaker 为一个 Core 创建唯一的唤醒与活动所有者。
func NewWaker(opt WakeOptions) *Waker {
	return &Waker{opt: opt, plugins: make(map[string]*wakePlugin)}
}

func (w *Waker) plugin(id string) *wakePlugin {
	p := w.plugins[id]
	if p == nil {
		p = &wakePlugin{changed: make(chan struct{})}
		w.plugins[id] = p
	}
	return p
}

func (p *wakePlugin) notify() {
	close(p.changed)
	p.changed = make(chan struct{})
}

// EnsureActive 等待原代次完整绑定并取得一次活动引用；取消只撤回本等待者。
func (w *Waker) EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error) {
	queued := false
	defer func() {
		if queued {
			w.mu.Lock()
			p := w.plugin(pluginID)
			p.count--
			p.bytes -= payloadBytes
			w.opt.Global.Release(payloadBytes)
			w.mu.Unlock()
		}
	}()
	for {
		w.mu.Lock()
		observed := w.plugin(pluginID).changed
		w.mu.Unlock()
		// 1. 在等待之前及每次唤醒后复核管理意图和已提交代次。
		if err := ctx.Err(); err != nil {
			return nil, wakeContextError(err)
		}
		in, err := w.opt.Store.GetInstall(ctx, pluginID)
		if err != nil {
			return nil, err
		}
		if in.Uninstalled || in.Source == "" || in.Source == "dev" || in.Role != state.RolePlugin && in.Role != state.RoleAgent {
			return nil, wakeError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin has no installed independent runtime")
		}
		enabled, err := w.opt.Store.Enabled(ctx, pluginID)
		if err != nil {
			return nil, err
		}
		if !enabled {
			return nil, wakeError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin is disabled")
		}
		generation, err := w.opt.Store.Active(ctx, pluginID)
		if err != nil {
			return nil, err
		}
		if generation == "" {
			return nil, wakeError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin has no committed generation")
		}
		status, err := w.opt.Store.RuntimeStatus(ctx, pluginID)
		if err != nil {
			return nil, err
		}

		// 2. 活动与停止共用同一门禁；本次引用只归当前 Session。
		w.mu.Lock()
		p := w.plugin(pluginID)
		if observed != p.changed {
			w.mu.Unlock()
			continue // 数据库读取期间发生了生命周期交接，重新取得一致的状态。
		}
		if err := ctx.Err(); err != nil {
			w.mu.Unlock()
			return nil, wakeContextError(err)
		}
		if current := p.current; status.State == state.RuntimeActive && current != nil && current.ready && !current.stopping && current.running.Instance.GenerationID == generation && current.active() {
			release := w.acquire(current)
			w.mu.Unlock()
			return release, nil
		}
		if p.failure != nil {
			err := p.failure
			w.mu.Unlock()
			return nil, err
		}
		// ACTIVE 可能先于 Session 登记可见；共享启动尚未交接时继续使用同一等待队列。
		waitable := status.State == state.RuntimeSuspended || status.State == state.RuntimeWaking ||
			status.State == state.RuntimeActive && (p.current != nil || p.waking) ||
			status.State == state.RuntimeQuiescing && p.current != nil && p.current.idleStop
		if !waitable {
			err := p.failure
			if err == nil {
				err = wakeError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, fmt.Sprintf("plugin runtime is %s: %s", status.State, status.Reason))
			}
			w.mu.Unlock()
			return nil, err
		}
		// 3. 每份等待载荷只预留一次；取消、失败和就绪走同一个归还出口。
		if !queued {
			reason := ""
			switch {
			case p.count >= w.opt.QueueRequests:
				reason = "wake request queue is full"
			case payloadBytes > w.opt.QueueBytes-p.bytes:
				reason = "wake payload queue is full"
			case !w.opt.Global.Acquire(payloadBytes):
				reason = "Core global payload budget is full"
			}
			if reason != "" {
				w.mu.Unlock()
				slog.Warn("wake request rejected", "plugin_id", pluginID, "reason", reason)
				return nil, wakeError(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, reason)
			}
			p.count++
			p.bytes += payloadBytes
			queued = true
		}
		if status.State == state.RuntimeSuspended && p.current == nil && !p.waking {
			p.waking, p.failure = true, nil
			// 启动不继承首个等待者的取消与期限；Start 装配方绑定 Core 生命周期。
			go w.start(context.WithoutCancel(ctx), pluginID, p)
		}
		changed := p.changed
		w.mu.Unlock()
		select { // 每个等待者独立退出，共享启动继续服务其余等待者。
		case <-ctx.Done():
			return nil, wakeContextError(ctx.Err())
		case <-changed:
		}
	}
}

func (w *Waker) start(ctx context.Context, id string, p *wakePlugin) {
	running, err := w.opt.Start(ctx, id)
	w.mu.Lock()
	if err == nil && (p.current == nil || p.current.running != running || !p.current.ready) {
		err = wakeError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin startup returned before complete binding")
	}
	w.mu.Unlock()
	if err != nil {
		err = wakeFailure(err)
	}
	w.mu.Lock()
	p.waking, p.failure = false, err
	p.notify()
	w.mu.Unlock()
}

// preparing/failed 也覆盖管理启动与候选激活，使等待者观察到共享启动以外的失败。
func (w *Waker) preparing(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(id)
	if p.current == nil {
		p.failure = nil
		p.notify()
	}
}

func (w *Waker) failed(id string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(id)
	if p.current == nil || p.current.stopping {
		p.failure = wakeFailure(err)
		p.notify()
	}
}

func wakeFailure(err error) error {
	var validation *validate.Error
	var phase *host.PhaseError
	if errors.As(err, &validation) {
		return errors.Join(wakeError(validation.Code, validation.Message), err)
	}
	if errors.As(err, &phase) {
		return errors.Join(wakeError(phase.Code, err.Error()), err)
	}
	return err
}

func wakeError(code pb.ErrorCode, message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func wakeContextError(err error) error {
	code := pb.ErrorCode_ERROR_CODE_CANCELLED
	if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return errors.Join(wakeError(code, err.Error()), err)
}
