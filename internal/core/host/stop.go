package host

import (
	"context"
	"sync"
	"time"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
)

// StopTiming 是停止预算（CONTRACTS §4.2）；测试注入更短的值。
type StopTiming struct {
	Drain    time.Duration // 冻结后最多排空的时间
	Total    time.Duration // 从 t0 起到终止进程树的总截止
	OnStopAt time.Duration // STOPPING 发出后留给 OnStop 与资源关闭的时间
}

// DefaultStopTiming 与 CONTRACTS §4.2 一致：排空 6s、总计 10s、OnStop 4s。
var DefaultStopTiming = StopTiming{Drain: 6 * time.Second, Total: 10 * time.Second, OnStopAt: 4 * time.Second}

// Tree 是被停止插件的进程树（process.Process 满足此接口）。
type Tree interface {
	Exited() <-chan struct{}
	KillTree() error
}

// StopResult 记录一次停止的经过，便于诊断与测试。
type StopResult struct {
	Drained        bool      // 在排空期限内完成了所有在途调用
	StoppingSentAt time.Time // 发出 STOPPING 的时刻（零值表示 Session 已关闭、未发送）
	StopDeadline   time.Time // STOPPING 携带的绝对截止
	Killed         bool      // 到总截止仍未退出，终止了进程树
}

// Stop 停用一个插件实例：从 t0 起冻结新根调用 → 最多排空 Drain → 取消剩余调用并发送唯一 STOPPING
// （截止为 min(t0+Total, 发送时刻+OnStopAt)）→ 等待进程退出，到 t0+Total 仍未退出则终止进程树。
// 只在确认进程退出（tree.Exited）之后返回。
func (in *Instance) Stop(t0 time.Time, reason pb.StopReason, tree Tree, timing StopTiming) StopResult {
	var res StopResult
	if !in.state.CompareAndSwap(int32(StateActivated), int32(StateDraining)) {
		in.state.Store(int32(StateDraining)) // Activating/Degraded 等状态同样不再接单
	}

	total := t0.Add(timing.Total)
	drainCtx, cancel := context.WithDeadline(context.Background(), t0.Add(timing.Drain))
	res.Drained = in.Session.WaitIdle(drainCtx) == nil
	cancel()

	in.Session.CancelOutbound(pb.CancelReason_CANCEL_REASON_SESSION_STOPPING)
	in.state.Store(int32(StateStopping))
	now := time.Now()
	deadline := now.Add(timing.OnStopAt)
	if total.Before(deadline) {
		deadline = total
	}
	if err := in.Session.SendControl(pb.FrameKind_FRAME_KIND_STOPPING,
		&pb.Stopping{DeadlineUnixMs: deadline.UnixMilli(), Reason: reason}); err == nil {
		res.StoppingSentAt, res.StopDeadline = now, deadline
	}

	timer := time.NewTimer(time.Until(total))
	defer timer.Stop()
	select {
	case <-tree.Exited():
	case <-timer.C:
		res.Killed = true
		_ = tree.KillTree()
		<-tree.Exited() // 必须确认旧进程退出，不能用 sleep 或猜测代替
	}
	_ = in.Session.Close() // 旧 Session 立即失效
	in.state.Store(int32(StateStopped))
	return res
}

// StopAll 并行停止多个实例（Core Shutdown），共享同一个 t0 与总截止，不逐个累加预算。
func StopAll(t0 time.Time, reason pb.StopReason, targets map[*Instance]Tree, timing StopTiming) map[*Instance]StopResult {
	var mu sync.Mutex
	out := make(map[*Instance]StopResult, len(targets))
	var wg sync.WaitGroup
	for in, tree := range targets {
		wg.Add(1)
		go func(in *Instance, tree Tree) {
			defer wg.Done()
			r := in.Stop(t0, reason, tree, timing)
			mu.Lock()
			out[in] = r
			mu.Unlock()
		}(in, tree)
	}
	wg.Wait()
	return out
}

// RestartPolicy 实现有限恢复（CONTRACTS §4.2）：窗口内第 Limit 次异常退出自动禁用；
// 正常退出与用户主动停用不重启；手动启用重置窗口。失败窗口只保存在内存中，禁用状态由调用方持久化。
type RestartPolicy struct {
	Window time.Duration
	Limit  int

	mu       sync.Mutex
	failures map[string][]time.Time
}

// NewRestartPolicy 使用首发规则：5 分钟内第 3 次异常退出时禁用。
func NewRestartPolicy() *RestartPolicy {
	return &RestartPolicy{Window: 5 * time.Minute, Limit: 3, failures: map[string][]time.Time{}}
}

// Decision 是一次进程退出后的处理决定。
type Decision int

const (
	DecisionNone    Decision = iota // 正常退出或用户停用：不重启
	DecisionRestart                 // 确认旧进程退出后启动新 Session
	DecisionDisable                 // 自动禁用并保留诊断
)

// OnExit 记录一次退出并给出决定。abnormal 为 false 表示正常退出（STOPPING 之后或退出码 0 且非崩溃）。
func (p *RestartPolicy) OnExit(pluginID string, abnormal bool, at time.Time) Decision {
	if !abnormal {
		return DecisionNone
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	kept := p.failures[pluginID][:0]
	for _, t := range p.failures[pluginID] {
		if at.Sub(t) < p.Window {
			kept = append(kept, t)
		}
	}
	kept = append(kept, at)
	p.failures[pluginID] = kept
	if len(kept) >= p.Limit {
		return DecisionDisable
	}
	return DecisionRestart
}

// Reset 在手动启用时清空失败窗口。
func (p *RestartPolicy) Reset(pluginID string) {
	p.mu.Lock()
	delete(p.failures, pluginID)
	p.mu.Unlock()
}
