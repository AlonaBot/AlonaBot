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
	"sync"
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

	// mu 串行化 Stop 与 WaitExit 的清理，保护 stopping 与 released。
	mu sync.Mutex
	// stopping 为 true 表示 Stop 已接管清理；此后的进程退出不算异常，也不由 WaitExit 清理。
	stopping bool
	// released 为 true 表示进程树已处理完、Listener 与进程资源已释放；只设置一次。
	released bool
	// cleanupErr 是 WaitExit 终止剩余进程树时的错误，供随后的 Stop 返回。
	cleanupErr error
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

// Stop 停用插件（CONTRACTS §4.2），确认进程树清空后关闭 Listener、释放进程资源。
// 与 Launch 共用逐插件管理锁（CONTRACTS §12.2）：停止完成前，同一插件的下一次启动不会开始。
// 进程已先行退出并由 WaitExit 清理时，只返回该次清理的错误。
func (r *Running) Stop(t0 time.Time, reason pb.StopReason) host.StopResult {
	unlock := r.store.LockPlugin(r.pluginID)
	defer unlock()

	// 1. 接管清理；WaitExit 已经清理过时不再触碰已释放的进程树句柄。
	r.mu.Lock()
	if r.released {
		r.mu.Unlock()
		return host.StopResult{Err: r.cleanupErr}
	}
	r.stopping = true
	r.mu.Unlock()

	// 2. 冻结、排空、STOPPING，并等待整棵进程树清空或在总截止终止它。
	res := r.Instance.Stop(t0, reason, r.Process, r.timing)

	// 3. 进程树处理完后才释放 Listener 与进程资源。
	r.mu.Lock()
	defer r.mu.Unlock()
	r.release()
	return res
}

// WaitExit 等待领头进程退出，报告退出状态与是否异常：未经 Stop 且退出码非 0 才算异常。
// 未经 Stop 的退出由这里终止树内剩余进程并释放资源；err 非 nil 表示旧进程树未确认清空，
// 调用方不能启动替代实例。
func (r *Running) WaitExit() (abnormal bool, state string, err error) {
	<-r.Process.Exited()
	st := r.Process.ExitState()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopping || r.released {
		return false, fmt.Sprint(st), nil // Stop 负责清理与结果
	}
	_ = r.Instance.Session.Close() // 旧 Session 立即失效
	r.cleanupErr = r.Process.KillTree()
	r.release()
	return !st.Success(), fmt.Sprint(st), r.cleanupErr
}

// release 关闭 Listener 并释放进程资源；调用方持有 r.mu，重复调用无副作用。
func (r *Running) release() {
	if r.released {
		return
	}
	r.released = true
	closeListener(r.listener)
	_ = r.Process.Release()
}
