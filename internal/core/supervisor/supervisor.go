// Package supervisor 把状态库、进程、传输与 Session 生命周期组装成启动和停止一个插件的完整流程。
//
// Launch 启动进程并完成握手、准备、提交与激活，任何一步失败都会结束整个进程树后再返回；
// Stop 排空并停用插件，WaitExit 在断链或退出时撤销业务入口，确认旧进程树清空后释放调度运行槽。
// Launch 与 Stop 由状态库的逐插件锁串行执行，Stop 与 WaitExit 由 Running 的锁交接清理所有权。
//
// 尚不含正式的安装、Journal 与恢复编排。
package supervisor

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/process"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// memoryLimitEnv 指定 Go 运行时读取插件内存软上限的环境变量。
const memoryLimitEnv = "GOMEMLIMIT"

// Core 全局默认语言的缺省值见 CONTRACTS §5.4。
const defaultLocale = "en-US"

// 受管 SQLite 的代次路径与权限见 CONTRACTS §6.1。
const (
	generationsDirectory    = "generations"
	generationDatabaseFile  = "db.sqlite"
	generationDirectoryMode = 0o700
)

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
	Bootstrap       *pb.Bootstrap
	ConfigSchema    []byte            // 当前安装包 config_schema.json 的规范 JSON；没有配置时为 nil
	I18n            map[string][]byte // 当前包 i18n/*.json 原文；无资源时为 nil
	DefaultLocale   string            // 本次 Session 的 Core 默认语言；空值使用 en-US
	Secrets         *secret.Store     // nil 时绑定本次 Launch 使用的状态库
	KeptSecretPaths []string          // 调用方额外提供的旧 secret 路径模式
	// NewCandidate 为 true 时在 ACTIVATE 前提交 ActiveGeneration（期望旧值 ExpectedOld）；
	// false 表示普通 restart，复用已提交的代次。
	NewCandidate bool
	Development  bool   // 开发代次每次读取原始配置，快照以明文直接覆盖。
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
	Global          *wire.Budget         // Core 全局载荷预算
	MemorySoftLimit int64                // 字节；大于 0 时为插件进程设置 GOMEMLIMIT
	Scheduler       *scheduler.Scheduler // 可选；激活后接入，排空时冻结，进程树清空后释放槽
	HTTP            *httpproxy.Server    // 可选；只绑定已激活 Session 的端点
	CommandsChanged func()               // 可选；激活与撤销后在状态锁外通知命令菜单。
	Detach          func()               // 可选；STOPPING 或断链时撤销装配层的本次 Session 绑定。
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
	listener        net.Listener
	timing          host.StopTiming
	scheduler       *scheduler.Scheduler
	http            *httpproxy.Server
	commandsChanged func()
	onDetach        func()

	// mu 串行化 Stop 与 WaitExit 的清理，保护 stopping、released 与 cleanupErr。
	mu sync.Mutex
	// stopping 为 true 表示 Stop 已接管清理；此后的进程退出不算异常，也不由 WaitExit 清理。
	stopping bool
	// released 为 true 表示进程树已处理完、Listener 与进程资源已释放；只设置一次。
	released   bool
	cleanupErr error // Stop 或 WaitExit 清理进程树时的错误，供随后的 Stop 返回
}

// Launch 启动插件进程并完成握手、准备、提交与激活。任何一步失败都会终止进程树、确认清空后返回；无法确认清空时把原因并入返回的错误。
// 同一插件的管理操作由状态库的逐插件锁串行化。
func Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error) {
	unlock := store.LockPlugin(spec.PluginID)
	defer unlock()

	// 启动前检查包内文案，失败时尚未读取候选资源或创建进程。
	messages, problems := i18n.Parse(spec.I18n)
	if len(problems) != 0 {
		details := make([]string, 0, len(problems))
		for _, problem := range problems {
			details = append(details, fmt.Sprintf("%s:%s: %s", problem.File, problem.Key, problem.Message))
		}
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid message catalog: " + strings.Join(details, "; ")}
	}

	// 1. 读取上一快照和路径；restart 的配置以已提交代次为准，新候选使用调用方提供的 Pending 快照。
	secrets := spec.Secrets
	if secrets == nil {
		secrets = secret.New(store)
	}
	previousID := spec.ExpectedOld
	if !spec.NewCandidate {
		previousID = spec.GenerationID
	}
	var previous []byte
	keptPaths := append([]string(nil), spec.KeptSecretPaths...)
	if previousID != "" {
		snapshot, err := store.GenerationConfig(ctx, previousID)
		if err != nil {
			return nil, err
		}
		previous = []byte(snapshot)
		paths, err := store.GenerationSecretPaths(ctx, previousID)
		if err != nil {
			return nil, err
		}
		keptPaths = append(keptPaths, paths...)
	}
	boot := proto.Clone(spec.Bootstrap).(*pb.Bootstrap)
	boot.DefaultLocale = spec.DefaultLocale
	if boot.DefaultLocale == "" {
		boot.DefaultLocale = defaultLocale
	}
	input := []byte(boot.GetCandidateConfigJson())
	if !spec.NewCandidate && !spec.Development {
		input = previous
	}
	plaintext := input
	if !spec.Development {
		var err error
		plaintext, err = secrets.Resolve(ctx, spec.PluginID, input)
		if err != nil {
			return nil, err
		}
	}
	boot.CandidateConfigJson = string(plaintext)

	// 2. 准备启动参数：工作目录与 runtime.lock 路径用同一个绝对路径：SDK 只接受绝对锁路径，相对路径也会随子进程 cwd 改变含义。
	pluginDir, err := filepath.Abs(spec.PluginDir)
	if err != nil {
		return nil, err
	}
	if database := boot.GetDatabaseDescriptor(); database.GetBackend() == pb.DatabaseBackend_DATABASE_BACKEND_SQLITE && database.GetSqlitePath() == "" {
		directory := filepath.Join(pluginDir, generationsDirectory, spec.GenerationID)
		database.SqlitePath = filepath.Join(directory, generationDatabaseFile)
		if spec.NewCandidate {
			if err := os.MkdirAll(directory, generationDirectoryMode); err != nil {
				return nil, err
			}
		}
	}
	sessionSecret, err := process.NewSecret()
	if err != nil {
		return nil, err
	}
	env := map[string]string{
		wire.EnvPluginID: spec.PluginID, wire.EnvSecret: sessionSecret,
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
	// 3. TCP 传输先持有 loopback Listener，再把地址交给插件。
	var ln net.Listener
	if spec.Stdio {
		env[wire.EnvTransport] = wire.TransportStdio
	} else {
		if ln, err = wire.ListenLoopback(); err != nil {
			return nil, err
		}
		env[wire.EnvCoreAddr] = ln.Addr().String()
	}

	// 4. 启动进程；之后任何失败都由 fail 终止并确认进程树清空再返回，无法确认时错误中带上原因。
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
			drainLogs(proc, time.Now().Add(spec.Stop.Total))
		}
		_ = proc.Release()
		closeListener(ln)
		if kerr != nil {
			err = errors.Join(err, fmt.Errorf("process tree of plugin %s not confirmed empty: %w", spec.PluginID, kerr))
		}
		return nil, err
	}

	// 5. 建立连接，注册后抽取快照；新候选先持久化，restart 只核对且保持原代次不变。
	conn, err := connect(ln, proc, spec)
	if err != nil {
		return fail(err)
	}
	secretBytes, _ := hex.DecodeString(sessionSecret)
	exp := host.Expect{Role: spec.Role, AdapterID: spec.AdapterID, PluginID: spec.PluginID, Secret: secretBytes, BuildDigest: spec.BuildDigest,
		Offer: validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}}
	exp.Snapshot = func(candidate []byte) ([]byte, error) {
		if spec.Development {
			var paths []string
			if len(spec.ConfigSchema) != 0 {
				schema, problems := configschema.Parse(spec.ConfigSchema)
				if len(problems) != 0 {
					return nil, fmt.Errorf("config schema: %v", problems)
				}
				for _, path := range schema.SecretPaths() {
					paths = append(paths, strings.Join(path, "."))
				}
			}
			if err := store.SetDevGenerationConfig(ctx, spec.GenerationID, string(candidate), boot.TargetConfigVersion, paths); err != nil {
				return nil, err
			}
			return candidate, nil
		}
		snapshot, paths, err := secrets.Extract(ctx, spec.PluginID, spec.ConfigSchema, candidate, previous, keptPaths)
		if err != nil {
			return nil, err
		}
		if spec.NewCandidate {
			if err := store.SetGenerationConfig(ctx, spec.GenerationID, string(snapshot), paths); err != nil {
				return nil, err
			}
		} else if !bytes.Equal(snapshot, previous) {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "config snapshot changed during restart"}
		}
		return snapshot, nil
	}
	prep, err := host.Prepare(ctx, conn, exp, boot, spec.Limits)
	if err != nil {
		return fail(err)
	}
	spec.Log.SetSession(prep.SessionID)
	opt := host.ActivateOptions{PluginID: spec.PluginID, GenerationID: spec.GenerationID, Handler: spec.Handler, Global: spec.Global, Limits: spec.Limits, Messages: messages}
	if spec.NewCandidate {
		opt.Commit = func(ctx context.Context) error {
			return store.CommitActive(ctx, spec.PluginID, spec.ExpectedOld, spec.GenerationID)
		}
	}
	in, err := host.Activate(ctx, conn, prep, opt)
	if err != nil {
		return fail(err)
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
			return fail(errors.Join(err, in.Session.Close()))
		}
	}
	if spec.CommandsChanged != nil {
		spec.CommandsChanged()
	}
	return &Running{Instance: in, Process: proc, store: store, pluginID: spec.PluginID, listener: ln, timing: spec.Stop,
		scheduler: spec.Scheduler, http: spec.HTTP, commandsChanged: spec.CommandsChanged, onDetach: spec.Detach}, nil
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
	res := r.Instance.Stop(t0, reason, r.Process, r.timing, host.StopHooks{
		Frozen: func(deadline time.Time) {
			if r.scheduler != nil {
				r.scheduler.Drain(r.pluginID)
			}
			if r.http != nil {
				r.http.Drain(r.pluginID, deadline)
			}
		},
		Stopping: func() {
			r.detach()
			if r.commandsChanged != nil {
				r.commandsChanged()
			}
		},
	})

	// 3. 进程树处理完后读完插件的最后输出，再释放 Listener 与进程资源；读取受同一总截止约束。
	drainLogs(r.Process, t0.Add(r.timing.Total))
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupErr = res.Err
	r.release()
	return res
}

// WaitExit 在 Session 断开或进程退出时撤销业务入口，再等待并清理旧进程树。
// 未经 Stop 且退出码非 0 才算异常；err 非 nil 表示旧进程树未确认清空，调用方不能启动替代实例。
func (r *Running) WaitExit() (abnormal bool, state string, err error) {
	// 1. 断链立即撤销入口，保留尚存进程的运行槽，不等待进程先自行退出。
	select {
	case <-r.Instance.Session.Done():
	case <-r.Process.Exited():
	}
	total := time.Now().Add(r.timing.Total)
	_ = r.Instance.Session.Close() // 进程退出也立即撤销 Session，菜单与派发共用同一就绪状态。
	if r.commandsChanged != nil {
		r.commandsChanged()
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.released {
		return false, fmt.Sprint(r.Process.ExitState()), r.cleanupErr
	}
	if r.stopping {
		// Stop 持有清理权；释放锁等待领头进程退出，让 Stop 可以完成资源释放。
		r.mu.Unlock()
		<-r.Process.Exited()
		r.mu.Lock()
		return false, fmt.Sprint(r.Process.ExitState()), nil
	}
	r.detach()
	// 2. 合作退出与日志收尾共用同一个总截止；到期仍存活才终止树内余留进程。
	// 组内已没有进程时不再发信号，避免进程组号被复用后误杀无关进程。
	if empty, err := r.Process.WaitTree(total); err != nil {
		r.cleanupErr = err
	} else if !empty {
		r.cleanupErr = r.Process.KillTree()
	}
	if r.cleanupErr == nil {
		<-r.Process.Exited() // 树清空后仍要取得 cmd.Wait 发布的真实退出状态。
	}
	drainLogs(r.Process, total)
	r.release()
	st := r.Process.ExitState()
	if r.cleanupErr != nil {
		return true, fmt.Sprint(st), r.cleanupErr
	}
	return !st.Success(), fmt.Sprint(st), r.cleanupErr
}

// detach 撤销本实例的注册；由 Stop 或 WaitExit 中取得清理权的一方调用。
func (r *Running) detach() {
	if r.onDetach != nil {
		r.onDetach()
	}
	if r.http != nil {
		r.http.Unbind(r.pluginID, r.Instance.GenerationID, r.Instance.HTTPProxyToken)
	}
	if r.scheduler != nil {
		r.scheduler.Detach(r.pluginID)
	}
}

// drainLogs 等待进程的日志管道读完，插件最后的输出才不会因 Core 随后退出而丢失（CONTRACTS §14.1）。
func drainLogs(proc *process.Process, deadline time.Time) {
	if !proc.WaitLogsUntil(deadline) {
		slog.Warn("plugin output was not fully read before the deadline", "plugin_id", proc.PluginID)
	}
}

// release 关闭 Listener 并释放进程资源；调用方持有 r.mu，重复调用无副作用。
func (r *Running) release() {
	if r.released {
		return
	}
	r.released = true
	if r.cleanupErr == nil && r.scheduler != nil {
		r.scheduler.Released(r.pluginID)
	}
	closeListener(r.listener)
	_ = r.Process.Release()
}
