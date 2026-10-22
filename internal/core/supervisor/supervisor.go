// Package supervisor 把状态库、进程、传输与 Session 生命周期装配成「启动 / 停止一个插件」的完整流程
// （PLAN T-02 第 8 步）。正式的安装、Journal 与恢复编排由 T-09 在此基础上扩展。
package supervisor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Spec 描述一次插件启动。
type Spec struct {
	PluginID     string
	PluginDir    string // <plugins>/<id>：工作目录与 runtime.lock 所在
	Executable   string // 已校验的入口可执行文件
	BuildDigest  []byte // 安装时已验证的归档摘要
	GenerationID string
	Bootstrap    *pb.Bootstrap // generation_id 与候选配置等；阶段预算由 host 填写
	// NewCandidate 为 true 时在 ACTIVATE 前提交 ActiveGeneration（期望旧值 ExpectedOld）；
	// false 表示普通 restart，复用已提交的代次。
	NewCandidate bool
	ExpectedOld  string
	Stdio        bool
	Args         []string          // 传给插件进程的额外参数（测试辅助进程使用）
	ExtraEnv     map[string]string // 额外环境变量（测试辅助进程使用）
	Limits       host.Limits
	Stop         host.StopTiming
	Log          process.LogFunc
	Handler      wire.Handler
	Global       *wire.Budget
}

// Running 是一个已激活的插件。
type Running struct {
	Instance *host.Instance
	Process  *process.Process
	store    *state.Store
	pluginID string
	listener net.Listener
	timing   host.StopTiming
	stopping atomic.Bool // Stop 与 WaitExit 在不同 goroutine 中访问
}

// Launch 启动插件进程并完成握手、准备、提交与激活。任何一步失败都会终止进程树并确认其退出后返回。
// 同一插件的管理操作由状态库的逐插件锁串行化。
func Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error) {
	unlock := store.LockPlugin(spec.PluginID)
	defer unlock()

	// 工作目录与 runtime.lock 路径用同一个绝对路径：SDK 只接受绝对锁路径，相对路径也会随子进程 cwd 改变含义。
	pluginDir, err := filepath.Abs(spec.PluginDir)
	if err != nil {
		return nil, err
	}
	secret, err := process.NewSecret()
	if err != nil {
		return nil, err
	}
	env := map[string]string{
		wire.EnvPluginID: spec.PluginID, wire.EnvSecret: secret,
		wire.EnvBuildDigest: hex.EncodeToString(spec.BuildDigest),
		wire.EnvRuntimeLock: filepath.Join(pluginDir, wire.RuntimeLockFile),
		wire.EnvTransport:   wire.TransportTCP,
	}
	for k, v := range spec.ExtraEnv {
		env[k] = v
	}
	var ln net.Listener
	if spec.Stdio {
		env[wire.EnvTransport] = wire.TransportStdio
	} else {
		if ln, err = wire.ListenLoopback(); err != nil {
			return nil, err
		}
		env[wire.EnvCoreAddr] = ln.Addr().String()
	}

	proc, err := process.Start(process.Spec{PluginID: spec.PluginID, Path: spec.Executable, Dir: pluginDir,
		Args: spec.Args, Env: env, Stdio: spec.Stdio}, spec.Log)
	if err != nil {
		closeListener(ln)
		return nil, err
	}
	fail := func(err error) (*Running, error) {
		_ = proc.KillTree()
		<-proc.Exited() // 确认旧进程退出后才返回，不能另起写入者
		_ = proc.Release()
		closeListener(ln)
		return nil, err
	}

	conn, err := connect(ln, proc, spec)
	if err != nil {
		return fail(err)
	}
	secretBytes, _ := hex.DecodeString(secret)
	exp := host.Expect{PluginID: spec.PluginID, Secret: secretBytes, BuildDigest: spec.BuildDigest,
		Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}}
	prep, err := host.Prepare(ctx, conn, exp, spec.Bootstrap, spec.Limits)
	if err != nil {
		return fail(err)
	}
	opt := host.ActivateOptions{PluginID: spec.PluginID, GenerationID: spec.GenerationID, Handler: spec.Handler, Global: spec.Global, Limits: spec.Limits}
	if spec.NewCandidate {
		opt.Commit = func(ctx context.Context) error {
			return store.CommitActive(ctx, spec.PluginID, spec.ExpectedOld, spec.GenerationID)
		}
	}
	in, err := host.Activate(ctx, conn, prep, opt)
	if err != nil {
		return fail(err)
	}
	return &Running{Instance: in, Process: proc, store: store, pluginID: spec.PluginID, listener: ln, timing: spec.Stop}, nil
}

// connect 在握手期限内建立与插件的连接：TCP 等待插件拨入（进程提前退出则放弃），stdio 直接使用管道。
func connect(ln net.Listener, proc *process.Process, spec Spec) (*wire.Conn, error) {
	if spec.Stdio {
		return wire.NewConn(proc.Stdout, proc.Stdin, closers{proc.Stdin, proc.Stdout}), nil
	}
	type accepted struct {
		c   net.Conn
		err error
	}
	ch := make(chan accepted, 1)
	if tl, ok := ln.(*net.TCPListener); ok {
		_ = tl.SetDeadline(time.Now().Add(spec.Limits.Handshake))
	}
	go func() { c, err := ln.Accept(); ch <- accepted{c, err} }()
	select {
	case a := <-ch:
		if a.err != nil {
			return nil, fmt.Errorf("等待插件连接失败（握手期限 %v）: %w", spec.Limits.Handshake, a.err)
		}
		if tl, ok := ln.(*net.TCPListener); ok {
			_ = tl.SetDeadline(time.Time{})
		}
		return wire.NewStreamConn(a.c), nil
	case <-proc.Exited():
		return nil, fmt.Errorf("插件进程在连接前退出: %v", proc.ExitState())
	}
}

type closers []io.Closer

func (cs closers) Close() error {
	var errs []error
	for _, c := range cs {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

func closeListener(ln net.Listener) {
	if ln != nil {
		_ = ln.Close()
	}
}

// Stop 停用插件（CONTRACTS §4.2），确认进程退出后关闭 Listener、释放进程资源。
// 与 Launch 共用逐插件管理锁（CONTRACTS §12.2）：停止完成前，同一插件的下一次启动不会开始。
func (r *Running) Stop(t0 time.Time, reason pb.StopReason) host.StopResult {
	unlock := r.store.LockPlugin(r.pluginID)
	defer unlock()
	r.stopping.Store(true)
	res := r.Instance.Stop(t0, reason, r.Process, r.timing)
	closeListener(r.listener)
	_ = r.Process.Release()
	return res
}

// WaitExit 等待插件进程退出，报告是否为异常退出（未经 Stop 的退出都视为异常）。
func (r *Running) WaitExit() (abnormal bool, state string) {
	<-r.Process.Exited()
	st := r.Process.ExitState()
	if !r.stopping.Load() {
		_ = r.Instance.Session.Close() // 旧 Session 立即失效
		closeListener(r.listener)
		_ = r.Process.Release()
	}
	return !r.stopping.Load(), fmt.Sprint(st)
}
