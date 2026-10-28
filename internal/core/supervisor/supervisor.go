// Package supervisor 把状态库、进程、传输与 Session 生命周期组装成启动和停止一个插件的完整流程。
//
// Launch 启动进程并完成握手、准备、提交与激活，任何一步失败都会结束整个进程树后再返回；
// Stop 停用插件，WaitExit 处理进程的意外退出。同一插件的这些操作由状态库的逐插件锁串行执行。
//
// 尚不含正式的安装、Journal 与恢复编排。
package supervisor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// memoryLimitEnv 指定 Go 运行时读取插件内存软上限的环境变量。
const memoryLimitEnv = "GOMEMLIMIT"

// Spec 描述一次插件启动。
type Spec struct {
	Role      state.Role // 必填，来自安装记录
	AdapterID string     // 安装记录绑定的平台类型
	PluginID  string
	// <plugins>/<id>：工作目录与 runtime.lock 所在；可以是相对路径，Launch 会转为绝对路径。
	PluginDir    string
	Executable   string // 已校验的入口可执行文件
	BuildDigest  []byte // 安装时已验证的归档摘要
	GenerationID string // 要激活的代次
	// Bootstrap 携带 generation_id 与候选配置等；阶段预算由 host 填写。
	Bootstrap *pb.Bootstrap
	// NewCandidate 为 true 时在 ACTIVATE 前提交 ActiveGeneration（期望旧值 ExpectedOld）；
	// false 表示普通 restart，复用已提交的代次。
	NewCandidate bool
	ExpectedOld  string // 提交新候选时预期的旧 ActiveGeneration；首次激活为空
	// Stdio 为 true 时使用 stdio 传输，否则使用 TCP loopback。
	Stdio    bool
	Args     []string          // 传给插件进程的额外参数（测试辅助进程使用）
	ExtraEnv map[string]string // 额外环境变量（测试辅助进程使用）
	Limits   host.Limits
	Stop     host.StopTiming
	Log      PluginLog // 接收插件进程的输出，必须设置
	// Handler 处理插件激活后发起的业务调用。
	Handler         wire.Handler
	Global          *wire.Budget // Core 全局载荷预算
	MemorySoftLimit int64        // 字节；大于 0 时为插件进程设置 GOMEMLIMIT
}

// PluginLog 接收一个插件进程的输出行；Launch 在 Session 建立后调用 SetSession，之后的行带上 session_id。
type PluginLog interface {
	// Line 的签名与 process.LogFunc 相同。
	Line(pluginID, stream string, line []byte)
	SetSession(sessionID []byte)
}

// Running 持有一个已激活插件的实例与进程，负责它的停止与清理。
type Running struct {
	Instance *host.Instance   // 已激活的 Core 侧实例
	Process  *process.Process // 插件进程及其进程树
	// store 提供逐插件管理锁。
	store    *state.Store
	pluginID string
	// TCP 传输的 loopback Listener；stdio 时为 nil。
	listener net.Listener
	timing   host.StopTiming

	// mu 串行化 Stop 与 WaitExit 的清理，保护 stopping、released 与 cleanupErr。
	mu sync.Mutex
	// stopping 为 true 表示 Stop 已接管清理；此后的进程退出不算异常，也不由 WaitExit 清理。
	stopping bool
	// released 为 true 表示进程树已处理完、Listener 与进程资源已释放；只设置一次。
	released   bool
	cleanupErr error // WaitExit 终止剩余进程树时的错误，供随后的 Stop 返回
}

// Launch 启动插件进程并完成握手、准备、提交与激活。任何一步失败都会终止进程树、确认清空后返回；无法确认清空时把原因并入返回的错误。
// 同一插件的管理操作由状态库的逐插件锁串行化。
func Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error) {
	unlock := store.LockPlugin(spec.PluginID)
	defer unlock()

	// 1. 准备启动参数：工作目录与 runtime.lock 路径用同一个绝对路径：SDK 只接受绝对锁路径，相对路径也会随子进程 cwd 改变含义。
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
	if spec.MemorySoftLimit > 0 {
		env[memoryLimitEnv] = strconv.FormatInt(spec.MemorySoftLimit, 10)
	}
	// 2. TCP 传输先持有 loopback Listener，再把地址交给插件。
	var ln net.Listener
	if spec.Stdio {
		env[wire.EnvTransport] = wire.TransportStdio
	} else {
		if ln, err = wire.ListenLoopback(); err != nil {
			return nil, err
		}
		env[wire.EnvCoreAddr] = ln.Addr().String()
	}

	// 3. 启动进程；之后任何失败都由 fail 终止并确认进程树清空再返回，无法确认时错误中带上原因。
	proc, err := process.Start(process.Spec{PluginID: spec.PluginID, Path: spec.Executable, Dir: pluginDir,
		Args: spec.Args, Env: env, Stdio: spec.Stdio}, spec.Log.Line)
	if err != nil {
		closeListener(ln)
		return nil, err
	}
	// fail 终止并确认进程树清空、释放资源后返回 err；无法确认清空时把原因并入错误，调用方不能另起写入者。
	fail := func(err error) (*Running, error) {
		kerr := proc.KillTree()
		if kerr == nil {
			<-proc.Exited()
		}
		_ = proc.Release()
		closeListener(ln)
		if kerr != nil {
			err = errors.Join(err, fmt.Errorf("process tree of plugin %s not confirmed empty: %w", spec.PluginID, kerr))
		}
		return nil, err
	}

	// 4. 建立连接，完成握手与准备，再提交（新候选）并激活。
	conn, err := connect(ln, proc, spec)
	if err != nil {
		return fail(err)
	}
	secretBytes, _ := hex.DecodeString(secret)
	exp := host.Expect{Role: spec.Role, AdapterID: spec.AdapterID, PluginID: spec.PluginID, Secret: secretBytes, BuildDigest: spec.BuildDigest,
		Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}}
	prep, err := host.Prepare(ctx, conn, exp, spec.Bootstrap, spec.Limits)
	if err != nil {
		return fail(err)
	}
	spec.Log.SetSession(prep.SessionID)
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
	// accepted 记录一次 Accept 的结果。
	type accepted struct {
		c   net.Conn // 插件拨入的连接
		err error    // Accept 的错误（含握手期限到期）
	}
	ch := make(chan accepted, 1)
	if tl, ok := ln.(*net.TCPListener); ok {
		_ = tl.SetDeadline(time.Now().Add(spec.Limits.Handshake))
	}
	go func() { c, err := ln.Accept(); ch <- accepted{c, err} }()
	select {
	case a := <-ch:
		if a.err != nil {
			return nil, fmt.Errorf("wait for plugin connection (handshake deadline %v): %w", spec.Limits.Handshake, a.err)
		}
		if tl, ok := ln.(*net.TCPListener); ok {
			_ = tl.SetDeadline(time.Time{})
		}
		return wire.NewStreamConn(a.c), nil
	case <-proc.Exited():
		return nil, fmt.Errorf("plugin process exited before connecting: %v", proc.ExitState())
	}
}

// closers 把多个 Closer 合成一个，用于 stdio 传输同时关闭两条管道。
type closers []io.Closer

// Close 依次关闭全部成员并合并错误。
func (cs closers) Close() error {
	var errs []error
	for _, c := range cs {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

// closeListener 关闭 Listener；ln 为 nil 时什么也不做。
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
	// 组内已没有进程时不再发信号，避免进程组号被复用后误杀无关进程。
	if empty, err := r.Process.WaitTree(time.Now()); err != nil {
		r.cleanupErr = err
	} else if !empty {
		r.cleanupErr = r.Process.KillTree()
	}
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
