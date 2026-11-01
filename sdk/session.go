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
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// handshakeTimeout 与 CONTRACTS §4.2 的 Hello/HelloAck 预算一致。
var handshakeTimeout = 10 * time.Second

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
		return nil, &Error{Code: validate.PublicCode(env.GetError().GetCode()), Message: "Core rejected the handshake: " + env.GetError().GetSafeMessage()}
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
func buildRegister(reg *Registration, configJSON string, configVersion uint32) *pb.Register {
	r := &pb.Register{CandidateConfigJson: configJSON, CandidateConfigVersion: configVersion,
		Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	for _, c := range reg.Commands {
		spec := &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH, Path: c.Path, HandlerMode: c.Mode, UsageMode: pb.UsageMode_USAGE_MODE_DEFAULT}
		for _, path := range c.Aliases {
			spec.Aliases = append(spec.Aliases, &pb.CommandAlias{Path: path})
		}
		for i, p := range c.Params {
			spec.Params = append(spec.Params, &pb.CommandParam{Name: fmt.Sprintf("arg%d", i+1), Type: p, Required: true})
		}
		r.Routes = append(r.Routes, &pb.RouteDescriptor{RouteId: commandRouteID(c), Priority: c.Priority, Kind: pb.RouteKind_ROUTE_KIND_COMMAND,
			DefaultEnabled: true, RequiredPermission: pb.Permission_PERMISSION_ANY, Spec: &pb.RouteDescriptor_Command{Command: spec}})
	}
	for _, event := range reg.Events {
		r.Routes = append(r.Routes, &pb.RouteDescriptor{RouteId: event.RouteID, Priority: event.Priority, Kind: pb.RouteKind_ROUTE_KIND_EVENT,
			DefaultEnabled: true, RequiredPermission: pb.Permission_PERMISSION_ANY,
			Spec: &pb.RouteDescriptor_Event{Event: &pb.EventSpec{EventTypeFilter: []string{event.EventType}, HandlerMode: pb.HandlerMode_HANDLER_MODE_MANUAL}}})
	}
	appendHTTPRoutes(r, reg.HTTP)
	appendServiceRoutes(reg, r)
	return r
}

// prepare 执行准备阶段：配置准备（目前原样透传）→ OnStart（独立 watchdog）→ REGISTER → 等待 PREPARED。
// 准备期不能发起业务 RPC。失败时已向 Core 报告 SESSION_ERROR 并关闭连接。
func prepare(conn *wire.Conn, l launch, ack *pb.HelloAck, root Root, reg *Registration) (*pb.Prepared, error) {
	sid, boot, minor := ack.GetSessionId(), ack.GetBootstrap(), ack.GetChosenMinor()
	limits := boot.GetPhaseLimits()
	// 1. Core 必须下发阶段预算与准备截止时间；缺失时报协议错误，而不是把预算当作 0 误报超时。
	if limits == nil || limits.GetOnStartMs() == 0 || boot.GetPreparationDeadlineUnixMs() == 0 {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK,
			pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "Bootstrap is missing phase limits or the preparation deadline", nil)
	}

	sessionLimits := boot.GetSessionLimits()
	if sessionLimits.GetMaxInflight() == 0 || sessionLimits.GetMaxPayloadBytes() == 0 || sessionLimits.GetRequestTimeoutMs() == 0 {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK,
			pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "Bootstrap is missing valid session limits", nil)
	}

	if err := l.http.start(boot); err != nil {
		return nil, sessionFailure(conn, sid, minor, l.pluginID, pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, pb.ErrorCode_ERROR_CODE_INTERNAL, "prepare HTTP listener failed", err)
	}

	// 2. OnStart：用户函数不响应 context 时，由 watchdog 判定失败（Run 随后退出进程）。
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

	// 3. 发送完整 REGISTER；超过 4 MiB 时报告 RESOURCE_EXHAUSTED。
	registration := buildRegister(reg, boot.GetCandidateConfigJson(), boot.GetTargetConfigVersion())
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

	// 4. 在准备截止前等待 PREPARED 或 Core 的 SESSION_ERROR。
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
		return nil, &Error{Code: validate.PublicCode(env.GetError().GetCode()), Message: "Core rejected the registration: " + env.GetError().GetSafeMessage()}
	}
	return nil, fmt.Errorf("unexpected %s during preparation", env.GetKind())
}
