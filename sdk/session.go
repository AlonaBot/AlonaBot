package sdk

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
)

// handshakeTimeout 与 CONTRACTS §4.2 的 Hello/HelloAck 预算一致。
var handshakeTimeout = 10 * time.Second

const scheduleRoutePrefix = "job."

// launch 保存 Core 通过环境变量下发的启动参数（见 internal/wire/env.go）。
type launch struct {
	pluginID  string
	transport string        // tcp 或 stdio
	addr      string        // TCP 传输时 Core 的 loopback 地址
	secret    []byte        // 32 字节 Session Secret
	digest    []byte        // 安装时已验证的归档摘要，在 HELLO 中原样回显
	http      *httpListener // 有HTTP路由时持有本Run的Listener，传给prepare与activation
}

// readLaunch 读取并校验 Core 下发的启动环境变量。
func readLaunch() (launch, error) {
	l := launch{pluginID: os.Getenv(wire.EnvPluginID), transport: os.Getenv(wire.EnvTransport), addr: os.Getenv(wire.EnvCoreAddr)}
	var err1, err2 error
	l.secret, err1 = hex.DecodeString(os.Getenv(wire.EnvSecret))
	l.digest, err2 = hex.DecodeString(os.Getenv(wire.EnvBuildDigest))
	switch {
	case l.pluginID == "":
		return l, fmt.Errorf("missing %s (the plugin must be started by Core)", wire.EnvPluginID)
	case err1 != nil || len(l.secret) != 32:
		return l, fmt.Errorf("%s is not 32 bytes of hex", wire.EnvSecret)
	case err2 != nil:
		return l, fmt.Errorf("%s is not hex", wire.EnvBuildDigest)
	case l.transport != wire.TransportTCP && l.transport != wire.TransportStdio:
		return l, fmt.Errorf("%s must be tcp or stdio", wire.EnvTransport)
	}
	return l, nil
}

// dial 建立到 Core 的传输。stdio 时 stdout 只传协议帧，插件日志必须写 stderr。
func dial(l launch) (*wire.Conn, error) {
	if l.transport == wire.TransportStdio {
		return wire.NewConn(os.Stdin, os.Stdout, closers{os.Stdin, os.Stdout}), nil
	}
	host, _, err := net.SplitHostPort(l.addr)
	if ip := net.ParseIP(host); err != nil || ip == nil || !ip.IsLoopback() {
		return nil, fmt.Errorf("%s must be a loopback address, got %q", wire.EnvCoreAddr, l.addr)
	}
	c, err := net.Dial("tcp", l.addr)
	if err != nil {
		return nil, err
	}
	return wire.NewStreamConn(c), nil
}

// closers 把多个 Closer 合成一个，用于 stdio 传输同时关闭 stdin 与 stdout。
type closers []io.Closer

// Close 依次关闭全部成员并合并错误。
func (cs closers) Close() error {
	var errs []error
	for _, c := range cs {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

// sdkOffer 声明 SDK 一侧的协议版本与能力。
var sdkOffer = validate.Offer{Major: validate.ProtocolMajor, Minor: validate.ProtocolMinor}

// withWatchdog 在 d 内未完成时关闭连接，使阻塞的读写返回。返回的 stop 必须调用。
func withWatchdog(conn *wire.Conn, d time.Duration) (stop func()) {
	t := time.AfterFunc(d, func() { _ = conn.Close() })
	return func() { t.Stop() }
}

// clientHandshake 发送 HELLO 并等待 HELLO_ACK。
func clientHandshake(conn *wire.Conn, l launch) (*pb.HelloAck, error) {
	stop := withWatchdog(conn, handshakeTimeout)
	defer stop()
	supported, required := httpOffer(l)
	hello, err := wire.ControlEnvelope(nil, 0, l.pluginID, pb.FrameKind_FRAME_KIND_HELLO, &pb.Hello{
		PluginId: l.pluginID, Secret: l.secret, SupportedMajor: sdkOffer.Major, SupportedMinor: sdkOffer.Minor,
		BuildDigest: l.digest, SupportedCapabilities: supported, RequiredCapabilities: required})
	if err != nil {
		return nil, err
	}
	if err := conn.WriteFrame(hello); err != nil {
		return nil, fmt.Errorf("send HELLO: %w", err)
	}
	env, err := conn.ReadFrame()
	if err != nil {
		return nil, fmt.Errorf("wait for HELLO_ACK (authentication rejected or timed out): %w", err)
	}
	msg, err := validate.DecodeControl(env)
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *pb.HelloAck:
		if len(m.GetSessionId()) != 16 {
			return nil, errors.New("HELLO_ACK has an invalid session_id")
		}
		if err := wire.CheckPeer(env, m.GetSessionId(), m.GetChosenMinor(), ""); err != nil {
			return nil, fmt.Errorf("invalid HELLO_ACK: %w", err)
		}
		return m, nil
	case *pb.SessionError:
		return nil, &Error{Code: errcode.Code(validate.PublicCode(env.GetError().GetCode())), Message: "Core rejected the handshake: " + env.GetError().GetSafeMessage()}
	}
	return nil, fmt.Errorf("unexpected %s during handshake", env.GetKind())
}

// sessionFailure 报告准备阶段失败：发送 SESSION_ERROR（只含安全消息），返回给调用方的错误保留原因。
func sessionFailure(conn *wire.Conn, sid []byte, minor uint32, pluginID string, phase pb.SessionPhase, kind pb.FrameKind, code pb.ErrorCode, safe string, cause error) error {
	env, err := wire.ControlEnvelope(sid, minor, pluginID, pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: phase, FailedKind: kind})
	if err == nil {
		env.Error = &pb.ErrorInfo{Code: code, SafeMessage: safe}
		_ = conn.WriteFrameTimeout(env, time.Second)
	}
	_ = conn.Close()
	if cause != nil {
		return fmt.Errorf("%s: %s: %w", phase, safe, cause)
	}
	return fmt.Errorf("%s: %s", phase, safe)
}

// buildRegister 把收集到的注册项转成 REGISTER。
// shortcut: route_id 暂用 "cmd.<路径以 . 连接>"，有稳定 route_id 生成器后替换。
func buildRegister(reg *Registration, configJSON string, configVersion uint32, configSchema []byte) *pb.Register {
	r := &pb.Register{CandidateConfigJson: configJSON, CandidateConfigVersion: configVersion,
		Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	if len(configSchema) != 0 {
		schema, _ := configschema.Parse(configSchema) // 配置准备已检查生成项，收集注册项时此参数为 nil。
		r.ConfigSchemaHash = schema.Hash()
	}
	for _, c := range reg.Commands {
		spec := c.spec
		spec.Rule = &pb.RuleDescriptor{HasCustomRule: len(c.rules) != 0}
		r.Routes = append(r.Routes, &pb.RouteDescriptor{RouteId: commandRouteID(c), Priority: c.Priority, Kind: pb.RouteKind_ROUTE_KIND_COMMAND,
			DefaultEnabled: true, RequiredPermissions: c.permission.Names(), Spec: &pb.RouteDescriptor_Command{Command: spec}})
	}
	for _, event := range reg.Events {
		r.Routes = append(r.Routes, &pb.RouteDescriptor{RouteId: event.RouteID, Priority: event.Priority, Kind: pb.RouteKind_ROUTE_KIND_EVENT,
			DefaultEnabled: true, RequiredPermissions: event.permission.Names(),
			Spec: &pb.RouteDescriptor_Event{Event: &pb.EventSpec{EventTypeFilter: []string{event.EventType.name}, ActionId: event.actionID, ActionSchemaVersion: event.actionVersion, HandlerMode: pb.HandlerMode_HANDLER_MODE_MANUAL, Rule: &pb.RuleDescriptor{HasCustomRule: len(event.rules) != 0}}}})
	}
	for _, job := range reg.Schedules {
		spec := &pb.ScheduleSpec{JobId: job.JobID, Timezone: job.Timezone}
		if job.Interval != 0 {
			spec.Trigger = &pb.ScheduleSpec_IntervalMs{IntervalMs: uint64(job.Interval / time.Millisecond)}
		} else {
			spec.Trigger = &pb.ScheduleSpec_Cron{Cron: job.Cron}
		}
		r.Routes = append(r.Routes, &pb.RouteDescriptor{RouteId: scheduleRoutePrefix + job.JobID, Kind: pb.RouteKind_ROUTE_KIND_SCHEDULE,
			DefaultEnabled: true, Spec: &pb.RouteDescriptor_Schedule{Schedule: spec}})
	}
	appendHTTPRoutes(r, reg.HTTP)
	appendServiceRoutes(reg, r)
	appendSubscriptionRoutes(reg, r)
	return r
}

// prepare 执行准备阶段：配置准备 → 迁移 → OnStart（各有独立预算）→ REGISTER → 等待 PREPARED。
// 准备期不能发起业务 RPC。失败时已向 Core 报告 SESSION_ERROR 并关闭连接。
func prepare(conn *wire.Conn, l launch, ack *pb.HelloAck, root Root, reg *Registration) (prepared *pb.Prepared, err error) {
	sid, boot, minor := ack.GetSessionId(), ack.GetBootstrap(), ack.GetChosenMinor()
	setI18n(reg.messages, boot.GetDefaultLocale())
	limits := boot.GetPhaseLimits()
	// 1. Core 必须下发阶段预算与准备截止时间；缺失时报协议错误，而不是把预算当作 0 误报超时。
	if limits == nil || limits.GetConfigMs() == 0 || limits.GetMigrationMs() == 0 || limits.GetOnStartMs() == 0 || boot.GetPreparationDeadlineUnixMs() == 0 {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK,
			pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "Bootstrap is missing phase limits or the preparation deadline", nil)
	}

	sessionLimits := boot.GetSessionLimits()
	if sessionLimits.GetMaxInflight() == 0 || sessionLimits.GetMaxPayloadBytes() == 0 || sessionLimits.GetRequestTimeoutMs() == 0 {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK,
			pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "Bootstrap is missing valid session limits", nil)
	}

	// 2. 在 OnStart 之前完成配置迁移、默认值与校验；restart 与恢复均使用 RESUME。
	configContext, cancelConfig := context.WithTimeout(context.Background(), time.Duration(limits.GetConfigMs())*time.Millisecond)
	preparedConfig, err := config.Prepare(configContext, root.OnConfigLoad, generatedFor(root).ConfigSchema,
		[]byte(boot.GetCandidateConfigJson()), boot.PreviousConfigVersion, boot.GetTargetConfigVersion(),
		boot.GetDatabaseDescriptor().GetActivationReason() == pb.ActivationReason_ACTIVATION_REASON_RESUME)
	cancelConfig()
	if err != nil {
		failure := err.(*validate.Error) // Prepare 的失败均携带协议错误码和安全消息。
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_UNSPECIFIED,
			failure.Code, failure.Message, err)
	}
	root.setConfig(preparedConfig.Values)

	// 3. 数据库从迁移阶段开始可用；后续任何准备失败都由这里关闭连接。
	dbhook.Begin(boot.GetDatabaseDescriptor())
	defer func() {
		if err != nil {
			err = errors.Join(err, dbhook.Close())
		}
	}()
	migrationContext, cancelMigration := context.WithTimeout(context.Background(), time.Duration(limits.GetMigrationMs())*time.Millisecond)
	migration, migrationError := prepareMigrations(migrationContext, root, boot.GetDatabaseDescriptor())
	cancelMigration()
	if migrationError != nil {
		code := pb.ErrorCode_ERROR_CODE_INTERNAL
		var failure *validate.Error
		if errors.As(migrationError, &failure) {
			code = failure.Code
		}
		if errors.Is(migrationError, context.DeadlineExceeded) {
			code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
		}
		slog.Error("database migration failed", "err", migrationError)
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_MIGRATION, pb.FrameKind_FRAME_KIND_UNSPECIFIED,
			code, "database migration failed", migrationError)
	}

	if err := l.http.start(boot); err != nil {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, pb.ErrorCode_ERROR_CODE_INTERNAL, "prepare HTTP listener failed", err)
	}

	// 4. OnStart：用户函数不响应 context 时，由 watchdog 判定失败（Run 随后退出进程）。
	budget := time.Duration(limits.GetOnStartMs()) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	errc := make(chan error, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				logPanic("OnStart", p)
				errc <- fmt.Errorf("OnStart panic: %v", p)
			}
		}()
		errc <- root.OnStart(ctx)
	}()
	timer := time.NewTimer(budget)
	select {
	case err := <-errc:
		timer.Stop()
		cancel()
		if err != nil {
			// 用户错误原文可能含敏感信息：只写日志，SESSION_ERROR 中只放安全消息。
			slog.Error("OnStart failed", "err", err)
			return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_ON_START, pb.FrameKind_FRAME_KIND_UNSPECIFIED,
				pb.ErrorCode_ERROR_CODE_INTERNAL, "OnStart returned an error", err)
		}
	case <-timer.C:
		cancel()
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_ON_START, pb.FrameKind_FRAME_KIND_UNSPECIFIED,
			pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "OnStart exceeded its budget", nil)
	}

	// 5. 发送完整 REGISTER；超过 4 MiB 时报告 RESOURCE_EXHAUSTED。
	registration := buildRegister(reg, string(preparedConfig.JSON), preparedConfig.Version, generatedFor(root).ConfigSchema)
	registration.Migration = migration
	l.http.report(registration)
	regEnv, err := wire.ControlEnvelope(sid, minor, l.pluginID, pb.FrameKind_FRAME_KIND_REGISTER, registration)
	if err != nil {
		return nil, err
	}
	err = conn.WriteFrame(regEnv)
	if errors.Is(err, wire.ErrFrameTooLarge) {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER,
			pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "registration exceeds 4 MiB", err)
	} else if err != nil {
		return nil, fmt.Errorf("send REGISTER: %w", err)
	}

	// 6. 在准备截止前等待 PREPARED 或 Core 的 SESSION_ERROR。
	deadline := time.UnixMilli(boot.GetPreparationDeadlineUnixMs())
	stop := withWatchdog(conn, time.Until(deadline))
	defer stop()
	env, err := conn.ReadFrame()
	if err != nil {
		return nil, fmt.Errorf("wait for PREPARED: %w", err)
	}
	if err := wire.CheckPeer(env, sid, minor, ""); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("invalid frame during preparation: %w", err)
	}
	msg, err := validate.DecodeControl(env)
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *pb.Prepared:
		return m, nil
	case *pb.SessionError:
		_ = conn.Close()
		return nil, &Error{Code: errcode.Code(validate.PublicCode(env.GetError().GetCode())), Message: "Core rejected the registration: " + env.GetError().GetSafeMessage()}
	}
	return nil, fmt.Errorf("unexpected %s during preparation", env.GetKind())
}

func prepareMigrations(ctx context.Context, root Root, descriptor *pb.DatabaseDescriptor) (*pb.MigrationState, error) {
	type result struct {
		state *pb.MigrationState
		err   error
	}
	completed := make(chan result, 1) // 超时后用户回调仍可返回，避免结果发送阻塞。
	go func() {
		defer func() {
			if p := recover(); p != nil {
				logPanic("migration", p)
				completed <- result{err: fmt.Errorf("migration panic: %v", p)}
			}
		}()
		migrators := dbhook.Recognize(root)
		if len(migrators) > 1 {
			completed <- result{err: &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "root plugin implements multiple migration interfaces"}}
			return
		}
		if descriptor.GetBackend() == pb.DatabaseBackend_DATABASE_BACKEND_NONE || len(migrators) == 0 {
			completed <- result{state: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}}
			return
		}
		state, err := migrators[0].Run(ctx)
		completed <- result{state: state, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-completed:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return result.state, result.err
	}
}

// streamTarget 保留完整 Target，话题缺省与显式值分别计入节流键。
type streamTarget struct {
	adapter, instance, account, chat, thread string
	hasThread                                bool
}

// streamThrottle 用令牌串行化同目标的编辑；令牌持有者独占 last，无需跨 RPC 持锁。
type streamThrottle struct {
	token chan struct{}
	last  time.Time
}

var streamSessions = struct {
	sync.Mutex
	targets map[*wire.Session]map[streamTarget]*streamThrottle
}{targets: make(map[*wire.Session]map[streamTarget]*streamThrottle)}

func sessionStreamThrottle(session *wire.Session, target *pb.Target) *streamThrottle {
	key := streamTarget{target.AdapterId, target.AdapterInstanceId, target.AccountId, target.ChatId, target.GetThreadId(), target.ThreadId != nil}
	streamSessions.Lock()
	defer streamSessions.Unlock()
	targets := streamSessions.targets[session]
	if targets == nil {
		targets = make(map[streamTarget]*streamThrottle)
		streamSessions.targets[session] = targets
		// 每个使用流的 Session 只有一个清理者，连接结束后释放全部目标状态。
		go func() {
			<-session.Done()
			streamSessions.Lock()
			delete(streamSessions.targets, session)
			streamSessions.Unlock()
		}()
	}
	throttle := targets[key]
	if throttle == nil {
		throttle = &streamThrottle{token: make(chan struct{}, 1)}
		throttle.token <- struct{}{}
		targets[key] = throttle
	}
	return throttle
}

func (s *streamThrottle) acquire(ctx context.Context, done <-chan struct{}, wait bool) (bool, error) {
	// 排队与间隔等待都监听取消；未到刷新时机的增量留在调用者缓存中。
	if wait {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-done:
			return false, errcode.FailedPrecondition.New("plugin session has stopped")
		case <-s.token:
		}
	} else {
		select {
		case <-s.token:
		default:
			return false, nil
		}
	}
	release := func() { s.token <- struct{}{} }
	if err := ctx.Err(); err != nil {
		release()
		return false, err
	}
	select {
	case <-done:
		release()
		return false, errcode.FailedPrecondition.New("plugin session has stopped")
	default:
	}
	if remaining := time.Until(s.last.Add(streamEditInterval)); remaining > 0 {
		if !wait {
			release()
			return false, nil
		}
		timer := time.NewTimer(remaining)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			release()
			return false, ctx.Err()
		case <-done:
			release()
			return false, errcode.FailedPrecondition.New("plugin session has stopped")
		case <-timer.C:
		}
	}
	return true, nil
}

func (s *streamThrottle) release() {
	// 从 RPC 完成时计时，确保阻塞或调度延迟也不会缩短实际 Edit 发起间隔。
	s.last = time.Now()
	s.token <- struct{}{}
}
