package supervisor

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Prepared 持有尚未激活的候选；调用方负责 Activate 或 Close。
// 准备失败且进程树尚未清空时 Registration 为 nil，资源仍由本句柄持有。
type Prepared struct {
	Registration   *host.Prepared
	store          *state.Store
	spec           Spec
	process        *process.Process
	conn           *wire.Conn
	listener       net.Listener
	messages       *i18n.Catalog
	snapshotRoutes bool
	released       bool // 进程资源已释放或已移交 Running。
	runtime        *runtimeRecord
	startedAt      time.Time     // process.Start 前的单调起点，跨准备与激活阶段保留。
	prepareElapsed time.Duration // host.Prepare 的实际调用区间。
}

// Activate 在提交回调成功后激活同一 Session，移交资源给返回的 Running。
// 提交后失败也返回 Degraded Running；返回 nil 时 Prepared 仍由调用方 Close。
// commit 为 nil 表示复用已提交的代次。
func (p *Prepared) Activate(ctx context.Context, commit func(context.Context) error) (*Running, error) {
	unlock := p.store.LockPlugin(p.spec.PluginID)
	defer unlock()
	return p.activate(ctx, commit, true)
}

func (p *Prepared) activate(ctx context.Context, commit func(context.Context) error, handoffFailure bool) (result *Running, err error) {
	if p.spec.Waker != nil {
		defer func() {
			if err != nil {
				p.spec.Waker.failed(p.spec.PluginID, err)
			}
		}()
	}
	if p.released || p.Registration == nil {
		return nil, errors.New("candidate session is not prepared")
	}
	stage := "activate"
	var activateElapsed time.Duration
	activateCalled := false
	defer func() {
		if err != nil {
			attrs := []any{"prepare_ns", p.prepareElapsed.Nanoseconds()}
			if activateCalled {
				attrs = append(attrs, "activate_ns", activateElapsed.Nanoseconds())
			}
			p.logStartupFailure(err, stage, attrs...)
		}
	}()
	if p.spec.NewCandidate && commit == nil {
		return p.activationFailed(ctx, nil, errors.New("candidate activation requires a commit callback"), handoffFailure)
	}
	spec := p.spec
	if !spec.NewCandidate && p.snapshotRoutes {
		if err := p.checkActiveRoutes(ctx); err != nil {
			return p.activationFailed(ctx, nil, err, false)
		}
	}
	opt := host.ActivateOptions{PluginID: spec.PluginID, GenerationID: spec.GenerationID, Handler: spec.Handler, Global: spec.Global, Limits: spec.Limits, Messages: p.messages}
	activity := &activitySession{}
	if spec.Waker != nil {
		opt.RequestActivity = func(in *wire.Inbound) (func(), *pb.ErrorInfo) {
			return spec.Waker.requestActivity(activity, in)
		}
	}
	opt.Commit = commit
	activateStartedAt := time.Now()
	in, err := host.Activate(ctx, p.conn, p.Registration, opt)
	completedAt := time.Now()
	activateElapsed, activateCalled = completedAt.Sub(activateStartedAt), true
	if err != nil {
		return p.activationFailed(ctx, in, err, handoffFailure)
	}
	// ACTIVATED 已核对；完整绑定及 Ready 的等待不计入冷启动耗时。
	slog.Debug("cold start completed", append(p.coldStartAttrs(), "duration_ns", completedAt.Sub(p.startedAt).Nanoseconds(),
		"prepare_ns", p.prepareElapsed.Nanoseconds(), "activate_ns", activateElapsed.Nanoseconds())...)
	stage = "binding"
	reason := "activated"
	if spec.Bootstrap.GetDatabaseDescriptor().GetActivationReason() == pb.ActivationReason_ACTIVATION_REASON_RESUME {
		reason = "resume"
	}
	if err := p.runtime.writeIfCurrent(ctx, state.RuntimeActive, reason); err != nil {
		return p.activationFailed(ctx, in, err, handoffFailure)
	}
	// 6. ACTIVATED 已应用后才开放 HTTP 与定时任务，绑定顺序保证首个 tick 可以访问 HTTP。
	activatedAt := time.Now()
	running := p.running(in)
	plugin := router.FromInstance(in)
	if spec.Waker != nil {
		spec.Waker.register(running, activity)
		plugin = router.WithActivity(plugin, func() (func(), error) { return spec.Waker.AcquireActivity(spec.PluginID, in.Session) })
	}
	if spec.HTTP != nil && in.HTTPListener != nil {
		spec.HTTP.Bind(spec.PluginID, in.GenerationID, in.HTTPListener, in.HTTPProxyToken, in.Routes, spec.Role)
	}
	if spec.Scheduler != nil {
		if err := spec.Scheduler.Attach(ctx, plugin, activatedAt); err != nil {
			if spec.Waker != nil {
				spec.Waker.released(running)
			}
			if spec.HTTP != nil && in.HTTPListener != nil {
				spec.HTTP.Unbind(spec.PluginID, in.GenerationID, in.HTTPProxyToken)
			}
			return p.activationFailed(ctx, in, err, handoffFailure)
		}
	}
	if spec.CommandsChanged != nil {
		spec.CommandsChanged()
	}
	p.released = true // 激活成功后，资源所有权移交给 Running。
	return running, nil
}

func (p *Prepared) coldStartAttrs() []any {
	reason := p.spec.Bootstrap.GetDatabaseDescriptor().GetActivationReason().String()
	attrs := []any{"plugin_id", p.spec.PluginID, "generation_id", p.spec.GenerationID,
		"activation_reason", strings.ToLower(strings.TrimPrefix(reason, "ACTIVATION_REASON_"))}
	if p.process != nil {
		attrs = append(attrs, "pid", p.process.Pid())
	}
	if p.Registration != nil {
		attrs = append(attrs, "session_id", hex.EncodeToString(p.Registration.SessionID))
	}
	return attrs
}

func (p *Prepared) logStartupFailure(err error, stage string, attrs ...any) {
	fields := append(p.coldStartAttrs(), "stage", stage, "elapsed_ns", time.Since(p.startedAt).Nanoseconds(), "error", err)
	var phase *host.PhaseError
	var validation *validate.Error
	var call *wire.CallError
	switch {
	case errors.As(err, &phase):
		fields = append(fields, "phase", phase.Phase.String(), "code", phase.Code.String())
	case errors.As(err, &validation):
		fields = append(fields, "code", validation.Code.String())
	case errors.As(err, &call):
		fields = append(fields, "code", call.ErrorInfo.GetCode().String())
	}
	slog.Debug("cold start failed", append(fields, attrs...)...)
}

func (p *Prepared) checkActiveRoutes(ctx context.Context) error {
	snapshot, err := p.store.ActiveRouteSnapshot(ctx, p.spec.PluginID)
	if err != nil {
		return fmt.Errorf("read active route snapshot: %w", err)
	}
	if snapshot.GenerationID != p.spec.GenerationID || !bytes.Equal(snapshot.BuildDigest, p.spec.BuildDigest) || !bytes.Equal(snapshot.RouteHash, p.Registration.RouteHash) {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "registered routes or program differ from the active route snapshot"}
	}
	return nil
}

func (p *Prepared) activationFailed(ctx context.Context, in *host.Instance, cause error, handoff bool) (*Running, error) {
	err := errors.Join(cause, p.runtime.writeIfCurrent(context.WithoutCancel(ctx), state.RuntimeFailed, cause.Error()))
	if in != nil {
		in.Degrade(cause)
		err = errors.Join(err, in.Session.Close())
	}
	if handoff {
		if in == nil {
			return nil, err
		}
		p.released = true
		return p.running(in), err
	}
	p.runtime.release()
	return nil, errors.Join(err, p.close())
}

func (p *Prepared) running(in *host.Instance) *Running {
	spec := p.spec
	return &Running{Instance: in, Process: p.process, store: p.store, pluginID: spec.PluginID, listener: p.listener, timing: spec.Stop,
		scheduler: spec.Scheduler, http: spec.HTTP, commandsChanged: spec.CommandsChanged, onDetach: spec.Detach, runtime: p.runtime, waker: spec.Waker}
}

// Close 关闭准备连接，按停止总预算等待合作退出，到期终止整棵树；确认清空后释放句柄。
// 树清空失败时保留所有权，允许调用方再次清理。
func (p *Prepared) Close() error {
	unlock := p.store.LockPlugin(p.spec.PluginID)
	defer unlock()
	if p.released {
		return nil
	}
	defer p.runtime.release()
	return errors.Join(p.close(), p.runtime.writeIfCurrent(context.Background(), state.RuntimeFailed, "candidate preparation aborted"))
}

func (p *Prepared) close() error {
	if p.released {
		return nil
	}
	deadline := time.Now().Add(p.spec.Stop.Total)
	var closeErr error
	if p.conn != nil {
		closeErr = p.conn.Close()
		if errors.Is(closeErr, net.ErrClosed) || errors.Is(closeErr, os.ErrClosed) {
			closeErr = nil
		}
	}
	empty, treeErr := p.process.WaitTree(deadline)
	if treeErr == nil && !empty {
		treeErr = p.process.KillTree()
	}
	if treeErr != nil {
		return errors.Join(closeErr, treeErr)
	}
	<-p.process.Exited() // 树已清空，等待 cmd.Wait 发布真实退出状态。
	drainLogs(p.process, deadline)
	if p.listener != nil {
		err := p.listener.Close()
		if !errors.Is(err, net.ErrClosed) {
			closeErr = errors.Join(closeErr, err)
		}
	}
	releaseErr := p.process.Release()
	if releaseErr == nil {
		p.released = true
	}
	return errors.Join(closeErr, releaseErr)
}
