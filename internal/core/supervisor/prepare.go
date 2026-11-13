package supervisor

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Prepared 持有尚未激活的候选；调用方负责 Activate 或 Close。
// 准备失败且进程树尚未清空时 Registration 为 nil，资源仍由本句柄持有。
type Prepared struct {
	Registration *host.Prepared
	store        *state.Store
	spec         Spec
	process      *process.Process
	conn         *wire.Conn
	listener     net.Listener
	messages     *i18n.Catalog
	released     bool // 进程资源已释放或已移交 Running。
}

// Activate 在提交回调成功后激活同一 Session，移交资源给返回的 Running。
// 提交后失败也返回 Degraded Running；返回 nil 时 Prepared 仍由调用方 Close。
// commit 为 nil 表示复用已提交的代次。
func (p *Prepared) Activate(ctx context.Context, commit func(context.Context) error) (*Running, error) {
	unlock := p.store.LockPlugin(p.spec.PluginID)
	defer unlock()
	return p.activate(ctx, commit, true)
}

func (p *Prepared) activate(ctx context.Context, commit func(context.Context) error, handoffFailure bool) (*Running, error) {
	if p.released || p.Registration == nil {
		return nil, errors.New("candidate session is not prepared")
	}
	if p.spec.NewCandidate && commit == nil {
		return nil, errors.New("candidate activation requires a commit callback")
	}
	spec := p.spec
	opt := host.ActivateOptions{PluginID: spec.PluginID, GenerationID: spec.GenerationID, Handler: spec.Handler, Global: spec.Global, Limits: spec.Limits, Messages: p.messages}
	opt.Commit = commit
	in, err := host.Activate(ctx, p.conn, p.Registration, opt)
	if err != nil {
		if handoffFailure {
			if in == nil {
				return nil, err
			}
			p.released = true
			return p.running(in), err
		}
		return nil, errors.Join(err, p.close())
	}
	// 6. ACTIVATED 已应用后才开放 HTTP 与定时任务，绑定顺序保证首个 tick 可以访问 HTTP。
	activatedAt := time.Now()
	if spec.HTTP != nil && in.HTTPListener != nil {
		spec.HTTP.Bind(spec.PluginID, in.GenerationID, in.HTTPListener, in.HTTPProxyToken, in.Routes, spec.Role == state.RoleAdapter)
	}
	if spec.Scheduler != nil {
		if err := spec.Scheduler.Attach(ctx, router.FromInstance(in), activatedAt); err != nil {
			if spec.HTTP != nil && in.HTTPListener != nil {
				spec.HTTP.Unbind(spec.PluginID, in.GenerationID, in.HTTPProxyToken)
			}
			if handoffFailure {
				in.Degrade(err)
				p.released = true
				return p.running(in), errors.Join(err, in.Session.Close())
			}
			return nil, errors.Join(err, in.Session.Close(), p.close())
		}
	}
	if spec.CommandsChanged != nil {
		spec.CommandsChanged()
	}
	p.released = true // 激活成功后，资源所有权移交给 Running。
	return p.running(in), nil
}

func (p *Prepared) running(in *host.Instance) *Running {
	spec := p.spec
	return &Running{Instance: in, Process: p.process, store: p.store, pluginID: spec.PluginID, listener: p.listener, timing: spec.Stop,
		scheduler: spec.Scheduler, http: spec.HTTP, commandsChanged: spec.CommandsChanged, onDetach: spec.Detach}
}

// Close 关闭准备连接，按停止总预算等待合作退出，到期终止整棵树；确认清空后释放句柄。
// 树清空失败时保留所有权，允许调用方再次清理。
func (p *Prepared) Close() error {
	unlock := p.store.LockPlugin(p.spec.PluginID)
	defer unlock()
	return p.close()
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
