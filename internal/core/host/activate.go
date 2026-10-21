package host

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/alonabot/contracts/validate"
	"github.com/AlonaBot/alonabot/internal/wire"
)

// State 是 Core 侧插件实例的可见状态（CONTRACTS §12.1，本文件覆盖 Activating 之后的部分）。
type State int32

const (
	StateActivating State = iota // 已提交（或复用已提交记录），等待 ACTIVATED；不派发业务
	StateActivated               // 唯一可以接单的状态
	StateDegraded                // 提交后失败：新代次仍是权威，不自动回到旧数据
	StateDraining                // 停止中：拒绝新根调用，已接纳调用的合法子调用仍可完成
	StateStopping                // 已发送 STOPPING，等待进程退出
	StateStopped                 // 进程已退出
)

func (s State) String() string {
	return [...]string{"Activating", "Activated", "Degraded", "Draining", "Stopping", "Stopped"}[s]
}

// ActivateOptions 配置一次激活。
type ActivateOptions struct {
	PluginID     string
	GenerationID string
	// Commit 在发送 ACTIVATE 之前于 Core 状态库提交唯一 ActiveGeneration+COMMITTED（新候选）。
	// 为 nil 表示普通 restart / 恢复：复用原有提交记录与 Generation，不再提交、不产生新代次。
	// 事务内只操作元数据，不得等待插件 RPC。
	Commit  func(ctx context.Context) error
	Handler wire.Handler // Activated 之后插件发起的业务调用
	Global  *wire.Budget
	Limits  Limits
}

// Instance 是一个已进入激活流程的插件 Session。
type Instance struct {
	PluginID, GenerationID string
	Session                *wire.Session

	state     atomic.Int32
	activated chan struct{}
	failOnce  sync.Once
	failErr   error
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

func (in *Instance) degrade(err error) {
	in.failOnce.Do(func() {
		in.failErr = err
		in.state.Store(int32(StateDegraded))
	})
}

// admit 是读循环中的同步门禁：Activated 之前（Preparing/Prepared/Activating）拒绝一切业务 REQUEST。
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
		in.degrade(&PhaseError{Phase: m.GetFailedPhase(), Code: validate.PublicCode(env.GetError().GetCode()), Msg: env.GetError().GetSafeMessage()})
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

	in := &Instance{PluginID: opt.PluginID, GenerationID: opt.GenerationID, activated: make(chan struct{})}
	in.state.Store(int32(StateActivating))
	in.Session = wire.NewSession(conn, wire.Options{
		Local: pb.InitiatorSide_INITIATOR_SIDE_CORE, SessionID: prep.SessionID, ProtocolMinor: prep.Minor,
		Handler: opt.Handler, Admit: in.admit, Control: in.control, Global: opt.Global,
	})

	deadline := time.Now().Add(opt.Limits.OnReady)
	marker := sha256.Sum256(append(append([]byte(opt.GenerationID), prep.RouteHash...), prep.ConfigDigest...))
	if err := in.Session.SendControl(pb.FrameKind_FRAME_KIND_ACTIVATE,
		&pb.Activate{GenerationId: opt.GenerationID, CommitMarker: marker[:], OnReadyDeadlineUnixMs: deadline.UnixMilli()}); err != nil {
		in.degrade(err)
		_ = in.Session.Close()
		return in, err
	}

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
