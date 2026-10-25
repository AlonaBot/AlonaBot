package sdk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// activation 跟踪 ACTIVATE → OnReady → ACTIVATED 与本地 ready gate（CONTRACTS §4.1）。
type activation struct {
	l    launch
	ack  *pb.HelloAck  // 握手结果，提供 Session ID、协商版本与 Bootstrap
	root plugin.Root   // 用户的根插件
	sess *wire.Session // 激活后使用的 Session

	ready   atomic.Bool   // ready gate：ACTIVATED 完整写出后才为 true
	readyCh chan struct{} // ready gate 打开时关闭
	started atomic.Bool   // 已收到 ACTIVATE
	// failOnce 保证只报告第一个激活失败。
	failOnce sync.Once
	// failErr 容量 1，接收激活失败的原因。
	failErr chan error

	// lifetime 在 STOPPING、断链或进程停止时取消；后台工作应监听它。
	lifetime context.Context
	// stopLifetime 取消 lifetime。
	stopLifetime context.CancelFunc
	stopping     atomic.Bool // 已收到 STOPPING
	// shutdownOnce 保证 OnStop 与关闭流程只执行一次。
	shutdownOnce sync.Once
	finished     chan struct{} // OnStop 结束（或超时）、Session 关闭后关闭
}

// admit 在读循环中同步检查 Core 的业务请求：ready gate 打开之前一律拒绝。
func (a *activation) admit(*wire.Inbound) *pb.ErrorInfo {
	if !a.ready.Load() {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "插件尚未就绪"}
	}
	return nil
}

// fail 报告第一个激活失败，之后的调用被忽略。
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
	case *pb.Stopping:
		if a.stopping.Swap(true) {
			return errors.New("重复的 STOPPING")
		}
		// 在读循环中同步关闭 ready gate：此后到达的业务请求一律被拒绝。
		a.ready.Store(false)
		go a.shutdown(time.UnixMilli(m.GetDeadlineUnixMs()))
		return nil
	case *pb.Heartbeat:
		return nil
	}
	return fmt.Errorf("激活阶段不应收到 %s", env.GetKind())
}

// runOnReady 在 ACTIVATE 给出的期限内执行 OnReady；用户函数不响应 ctx 时由 watchdog 判定失败。
func (a *activation) runOnReady(m *pb.Activate) {
	// 1. 在 ACTIVATE 给出的截止前执行 OnReady（如有）；失败或超时报告 SESSION_ERROR。
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
	// 2. 唯一 writer 完整写出 ACTIVATED 之后才打开 ready gate。
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

// onStopBudget 限制断链（未收到 STOPPING）时留给 OnStop 与资源关闭的时间（CONTRACTS §4.2）。
var onStopBudget = 4 * time.Second

// shutdown 取消 lifetime、关闭 ready gate，然后用独立 context 在 deadline 前执行 OnStop（不复用已取消的
// lifetime）；用户函数不响应 ctx 时 watchdog 到期即放弃等待。最后关闭 Session。只执行一次。
func (a *activation) shutdown(deadline time.Time) {
	a.shutdownOnce.Do(func() {
		// 1. 关闭 ready gate 并取消 lifetime，不再接受新业务。
		a.ready.Store(false)
		a.stopLifetime()
		// 2. 用独立的截止 context 执行 OnStop，最多等到 deadline。
		if s, ok := a.root.(Stopper); ok {
			ctx, cancel := context.WithDeadline(context.Background(), deadline)
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer func() {
					if p := recover(); p != nil {
						fmt.Fprintf(os.Stderr, "OnStop panic: %v\n", p)
					}
				}()
				if err := s.OnStop(ctx); err != nil {
					fmt.Fprintf(os.Stderr, "OnStop 失败: %v\n", err)
				}
			}()
			timer := time.NewTimer(time.Until(deadline))
			select {
			case <-done:
			case <-timer.C:
				fmt.Fprintln(os.Stderr, "OnStop 超过停止截止时间，放弃等待")
			}
			timer.Stop()
			cancel()
		}
		// 3. 关闭 Session 并通知 serve 返回。
		_ = a.sess.Close()
		close(a.finished)
	})
}

// serve 在激活之后等待停止：收到 STOPPING 时按其截止执行 OnStop 后正常返回；
// 断链或 Core 消失时立即停止新工作，在 onStopBudget 内执行 OnStop 后返回错误（不重连）。
func serve(sess *wire.Session, a *activation) error {
	select {
	case <-a.finished:
		return nil
	case <-sess.Done():
	}
	if a.stopping.Load() {
		<-a.finished
		return nil
	}
	a.shutdown(time.Now().Add(onStopBudget))
	return fmt.Errorf("与 Core 的连接已断开: %w", sess.Err())
}

// businessHandler 处理 ready 之后的 Core 请求。
// shortcut: 尚无业务消息分发，一律回 UNSUPPORTED；接入命令与事件分发时替换。
func businessHandler(context.Context, *wire.Inbound) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "插件尚未实现业务分发"}}
}

// activate 在 PREPARED 之后启动 Session，等待 ACTIVATE、执行 OnReady 并写出 ACTIVATED。
// 返回时 ready gate 已打开；失败时已向 Core 报告并关闭 Session。之后用 serve 等待停止。
func activate(conn *wire.Conn, l launch, ack *pb.HelloAck, root plugin.Root) (*wire.Session, *activation, error) {
	a := &activation{l: l, ack: ack, root: root, readyCh: make(chan struct{}), failErr: make(chan error, 1), finished: make(chan struct{})}
	a.lifetime, a.stopLifetime = context.WithCancel(context.Background())
	a.sess = wire.New(conn, wire.Options{
		Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: ack.GetSessionId(), ProtocolMinor: ack.GetChosenMinor(),
		SourcePluginID: l.pluginID, Handler: businessHandler, Admit: a.admit, Control: a.control,
	})
	a.sess.Start() // a.sess 赋值之后再启动，回调中使用的指针一定有效
	select {
	case <-a.readyCh:
		return a.sess, a, nil
	case err := <-a.failErr:
		a.stopLifetime()
		return nil, nil, err
	case <-a.sess.Done():
		a.stopLifetime()
		select {
		case err := <-a.failErr:
			return nil, nil, err
		default:
		}
		return nil, nil, fmt.Errorf("激活前 Session 关闭: %w", a.sess.Err())
	}
}
