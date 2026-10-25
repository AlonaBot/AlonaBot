package host

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// State 表示 Core 侧插件实例的可见状态（CONTRACTS §12.1），这里覆盖 Activating 之后的部分。
type State int32

// 实例状态，按生命周期顺序排列。
const (
	StateActivating State = iota // 已提交（或复用已提交记录），等待 ACTIVATED；不派发业务
	StateActivated               // 唯一可以接单的状态
	StateDegraded                // 提交后失败：新代次仍是权威，不自动回到旧数据
	StateDraining                // 停止中：拒绝新根调用，已接纳调用的合法子调用仍可完成
	StateStopping                // 已发送 STOPPING，等待进程退出
	StateStopped                 // 进程已退出
)

// String 返回状态名。
func (s State) String() string {
	return [...]string{"Activating", "Activated", "Degraded", "Draining", "Stopping", "Stopped"}[s]
}

// ActivateOptions 配置一次激活。
type ActivateOptions struct {
	PluginID     string // 插件 ID，Session 用它核对对端帧的来源
	GenerationID string // 要激活的代次，ACTIVATED 必须回显它
	// Commit 在发送 ACTIVATE 之前于 Core 状态库提交唯一 ActiveGeneration+COMMITTED（新候选）。
	// 为 nil 表示普通 restart / 恢复：复用原有提交记录与 Generation，不再提交、不产生新代次。
	// 事务内只操作元数据，不得等待插件 RPC。
	Commit func(ctx context.Context) error
	// Handler 处理 Activated 之后插件发起的业务调用。
	Handler wire.Handler
	Global  *wire.Budget // Core 全局载荷预算
	// Limits 提供 OnReady 预算。
	Limits Limits
}

// Instance 跟踪一个已进入激活流程的插件 Session 及其状态。
type Instance struct {
	PluginID     string
	GenerationID string        // 本实例服务的代次
	Session      *wire.Session // 本实例的调用引擎；进入激活流程时创建

	state atomic.Int32 // State，读循环与停止流程并发访问
	// activated 在读循环应用 ACTIVATED 时关闭。
	activated chan struct{}
	// failOnce 保证只记录第一个失败原因。
	failOnce sync.Once
	failErr  error // 进入 Degraded 的原因，由 failOnce 写入一次
}

// State 返回当前状态。
func (in *Instance) State() State { return State(in.state.Load()) }

// Err 返回进入 Degraded 的原因（未失败时为 nil）。
func (in *Instance) Err() error {
	if in.State() != StateDegraded {
		return nil
	}
	return in.failErr
}

// degrade 记录第一个失败原因并进入 Degraded。
func (in *Instance) degrade(err error) {
	in.failOnce.Do(func() {
		in.failErr = err
		in.state.Store(int32(StateDegraded))
	})
}

// admit 在读循环中同步检查业务 REQUEST：Activated 之前（Preparing/Prepared/Activating）一律拒绝。
// Draining 时只放行引用 Core 在途请求的子调用（parent_call_id），让已接纳的调用链能够完成。
func (in *Instance) admit(req *wire.Inbound) *pb.ErrorInfo {
	switch in.State() {
	case StateActivated:
		return nil
	case StateDraining:
		if req.ParentCallID != 0 {
			return nil
		}
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "插件正在停止，拒绝新的根调用"}
	}
	return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "插件尚未激活，拒绝业务请求"}
}

// control 在读循环中同步处理控制帧：ACTIVATED 在这里生效，因此紧随其后的 REQUEST 一定看到 Activated。
func (in *Instance) control(env *pb.Envelope) error {
	msg, err := validate.DecodeControl(env)
	if err != nil {
		return err
	}
	switch m := msg.(type) {
	case *pb.Activated:
		if in.State() != StateActivating || m.GetGenerationId() != in.GenerationID || !m.GetReady() {
			return fmt.Errorf("意外的 ACTIVATED（状态 %s，代次 %q）", in.State(), m.GetGenerationId())
		}
		in.state.Store(int32(StateActivated))
		close(in.activated)
		return nil
	case *pb.SessionError:
		in.degrade(&PhaseError{Phase: m.GetFailedPhase(), Code: validate.PublicCode(env.GetError().GetCode()), Message: env.GetError().GetSafeMessage()})
		return errors.New("插件报告 SESSION_ERROR")
	case *pb.Heartbeat:
		return nil
	}
	return fmt.Errorf("激活后不应收到 %s", env.GetKind())
}

// Activate 完成「提交 → ACTIVATE → OnReady → ACTIVATED」。提交失败时不发送 ACTIVATE，旧代次保持权威；
// 提交之后 OnReady 失败、超时或断链进入 Degraded，状态库中仍是新代次（CONTRACTS §4.1、§12.3）。
// 返回的 Instance 在 Degraded 时也非 nil，便于调用方记录与清理。
func Activate(ctx context.Context, conn *wire.Conn, prep *Prepared, opt ActivateOptions) (*Instance, error) {
	// 1. 新候选先提交；失败时通知插件并关闭，不发送 ACTIVATE。
	if opt.Commit != nil {
		if err := opt.Commit(ctx); err != nil {
			if e, eerr := wire.ControlEnvelope(prep.SessionID, prep.Minor, "", pb.FrameKind_FRAME_KIND_SESSION_ERROR,
				&pb.SessionError{FailedPhase: pb.SessionPhase_SESSION_PHASE_ACTIVATE}); eerr == nil {
				e.Error = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "Core 提交候选代次失败"}
				_ = conn.WriteFrameTimeout(e, failNotifyTimeout)
			}
			_ = conn.Close()
			return nil, fmt.Errorf("提交代次 %s 失败（未发送 ACTIVATE，旧代次保持权威）: %w", opt.GenerationID, err)
		}
	}

	// 2. 在 Activating 状态下启动 Session，并发出带 OnReady 截止时间的 ACTIVATE。
	in := &Instance{PluginID: opt.PluginID, GenerationID: opt.GenerationID, activated: make(chan struct{})}
	in.state.Store(int32(StateActivating))
	in.Session = wire.NewSession(conn, wire.Options{
		Local: pb.InitiatorSide_INITIATOR_SIDE_CORE, SessionID: prep.SessionID, ProtocolMinor: prep.Minor,
		SourcePluginID: opt.PluginID, Handler: opt.Handler, Admit: in.admit, Control: in.control, Global: opt.Global,
	})

	deadline := time.Now().Add(opt.Limits.OnReady)
	marker := sha256.Sum256(append(append([]byte(opt.GenerationID), prep.RouteHash...), prep.ConfigDigest...))
	if err := in.Session.SendControl(pb.FrameKind_FRAME_KIND_ACTIVATE,
		&pb.Activate{GenerationId: opt.GenerationID, CommitMarker: marker[:], OnReadyDeadlineUnixMs: deadline.UnixMilli()}); err != nil {
		in.degrade(err)
		_ = in.Session.Close()
		return in, err
	}

	// 3. 等待读循环应用 ACTIVATED；超时、断链或取消则进入 Degraded。
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	var failure error
	select {
	case <-in.activated:
		return in, nil
	case <-timer.C:
		failure = phaseErr(pb.SessionPhase_SESSION_PHASE_ON_READY, pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "OnReady 超过预算")
	case <-in.Session.Done():
		failure = phaseErr(pb.SessionPhase_SESSION_PHASE_ACTIVATE, pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "激活期间 Session 断开: %v", in.Session.Err())
	case <-ctx.Done():
		failure = phaseErr(pb.SessionPhase_SESSION_PHASE_ACTIVATE, pb.ErrorCode_ERROR_CODE_CANCELLED, "激活被取消")
	}
	// ACTIVATED 可能与超时同时到达：只有仍处于 Activating 时才判定失败。
	if !in.state.CompareAndSwap(int32(StateActivating), int32(StateDegraded)) {
		if in.State() == StateActivated {
			return in, nil
		}
	} else {
		in.failOnce.Do(func() { in.failErr = failure })
	}
	_ = in.Session.Close()
	return in, in.Err()
}
