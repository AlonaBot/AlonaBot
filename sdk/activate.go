package sdk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/alonabot/contracts/validate"
	"github.com/AlonaBot/alonabot/internal/wire"
	"github.com/AlonaBot/alonabot/sdk/plugin"
)

// activation 跟踪 ACTIVATE → OnReady → ACTIVATED 与本地 ready gate（CONTRACTS §4.1）。
type activation struct {
	l    launch
	ack  *pb.HelloAck
	root plugin.Root
	sess *wire.Session

	ready    atomic.Bool   // ready gate：ACTIVATED 完整写出后才为 true
	readyCh  chan struct{} // ready gate 打开时关闭
	started  atomic.Bool   // 已收到 ACTIVATE
	failOnce sync.Once
	failErr  chan error
}

// admit 是读循环中的同步门禁：ready gate 打开之前拒绝 Core 的业务请求。
func (a *activation) admit(*wire.Inbound) *pb.ErrorInfo {
	if !a.ready.Load() {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "插件尚未就绪"}
	}
	return nil
}

func (a *activation) fail(err error) {
	a.failOnce.Do(func() { a.failErr <- err })
}

// control 在读循环中同步处理控制帧；OnReady 在独立 goroutine 中运行，不阻塞读循环。
func (a *activation) control(env *pb.Envelope) error {
	msg, err := validate.DecodeControl(env)
	if err != nil {
		return err
	}
	switch m := msg.(type) {
	case *pb.Activate:
		if a.started.Swap(true) {
			return errors.New("重复的 ACTIVATE")
		}
		if m.GetGenerationId() != a.ack.GetBootstrap().GetGenerationId() {
			return fmt.Errorf("ACTIVATE 的代次 %q 与 Bootstrap 不符", m.GetGenerationId())
		}
		go a.runOnReady(m)
		return nil
	case *pb.Heartbeat:
		return nil
	}
	return fmt.Errorf("激活阶段不应收到 %s", env.GetKind())
}

// runOnReady 在 ACTIVATE 给出的期限内执行 OnReady；用户函数不响应 ctx 时由 watchdog 判定失败。
func (a *activation) runOnReady(m *pb.Activate) {
	deadline := time.UnixMilli(m.GetOnReadyDeadlineUnixMs())
	if r, ok := a.root.(Readier); ok {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		errc := make(chan error, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					errc <- fmt.Errorf("OnReady panic: %v", p)
				}
			}()
			errc <- r.OnReady(ctx)
		}()
		timer := time.NewTimer(time.Until(deadline))
		defer timer.Stop()
		select {
		case err := <-errc:
			if err != nil {
				fmt.Fprintf(os.Stderr, "OnReady 失败: %v\n", err)
				a.reportFailure(pb.ErrorCode_ERROR_CODE_PLUGIN_NOT_READY, "OnReady 返回错误", err)
				return
			}
		case <-timer.C:
			a.reportFailure(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "OnReady 超过预算", nil)
			return
		}
	}
	// 唯一 writer 完整写出 ACTIVATED 之后才打开 ready gate。
	if err := a.sess.SendControlSync(pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: m.GetGenerationId(), Ready: true}); err != nil {
		a.fail(fmt.Errorf("写出 ACTIVATED 失败: %w", err))
		return
	}
	a.ready.Store(true)
	close(a.readyCh)
}

// reportFailure 发送 SESSION_ERROR(ON_READY)（只含安全消息）后关闭 Session。
func (a *activation) reportFailure(code pb.ErrorCode, safe string, cause error) {
	env, err := wire.ControlEnvelope(a.ack.GetSessionId(), a.ack.GetChosenMinor(), a.l.pluginID, pb.FrameKind_FRAME_KIND_SESSION_ERROR,
		&pb.SessionError{FailedPhase: pb.SessionPhase_SESSION_PHASE_ON_READY})
	if err == nil {
		env.Error = &pb.ErrorInfo{Code: code, SafeMessage: safe}
		_ = a.sess.SendEnvelopeSync(env)
	}
	_ = a.sess.Close()
	if cause != nil {
		a.fail(fmt.Errorf("%s: %w", safe, cause))
	} else {
		a.fail(errors.New(safe))
	}
}

// businessHandler 处理 ready 之后的 Core 请求。
// shortcut: B1 尚无消息分发（T-03/T-04 接入），一律回 UNSUPPORTED。
func businessHandler(context.Context, *wire.Inbound) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "插件尚未实现业务分发"}}
}

// activate 在 PREPARED 之后启动 Session，等待 ACTIVATE、执行 OnReady 并写出 ACTIVATED。
// 返回时 ready gate 已打开；失败时已向 Core 报告并关闭 Session。
func activate(conn *wire.Conn, l launch, ack *pb.HelloAck, root plugin.Root) (*wire.Session, error) {
	a := &activation{l: l, ack: ack, root: root, readyCh: make(chan struct{}), failErr: make(chan error, 1)}
	a.sess = wire.New(conn, wire.Options{
		Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: ack.GetSessionId(), ProtocolMinor: ack.GetChosenMinor(),
		SourcePluginID: l.pluginID, Handler: businessHandler, Admit: a.admit, Control: a.control,
	})
	a.sess.Start() // a.sess 赋值之后再启动，回调中使用的指针一定有效
	select {
	case <-a.readyCh:
		return a.sess, nil
	case err := <-a.failErr:
		return nil, err
	case <-a.sess.Done():
		select {
		case err := <-a.failErr:
			return nil, err
		default:
		}
		return nil, fmt.Errorf("激活前 Session 关闭: %w", a.sess.Err())
	}
}
