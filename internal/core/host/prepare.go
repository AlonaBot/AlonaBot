// Package host 是 Core 侧单个插件 Session 的生命周期编排（CONTRACTS §1.1、§4、§12.1）：
// 握手与认证、Bootstrap 下发、REGISTER 原子校验、PREPARED；提交与激活见 activate.go（G1-17）。
package host

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Limits 是阶段预算（CONTRACTS §4.2 的首发常量）；测试可注入更短的值。
type Limits struct {
	// Handshake 是从连接建立到 HELLO 认证完成的预算。
	Handshake time.Duration
	// Config 是配置转换与校验的预算。
	Config time.Duration
	// Migration 是数据库迁移的预算。
	Migration time.Duration
	// OnStart 是 OnStart 的预算。
	OnStart time.Duration
	// OnReady 是 OnReady 的预算。
	OnReady time.Duration
	// StopTotal 是一次停用的总预算。
	StopTotal time.Duration
}

// DefaultLimits 与 CONTRACTS §4.2 一致。
var DefaultLimits = Limits{
	Handshake: 10 * time.Second, Config: 60 * time.Second, Migration: 15 * time.Minute,
	OnStart: 60 * time.Second, OnReady: 30 * time.Second, StopTotal: 10 * time.Second,
}

// proto 把预算转为 Bootstrap 中的毫秒值。
func (l Limits) proto() *pb.PhaseLimits {
	ms := func(d time.Duration) uint32 { return uint32(d / time.Millisecond) }
	return &pb.PhaseLimits{HandshakeMs: ms(l.Handshake), ConfigMs: ms(l.Config), MigrationMs: ms(l.Migration),
		OnStartMs: ms(l.OnStart), OnReadyMs: ms(l.OnReady), StopTotalMs: ms(l.StopTotal)}
}

// failNotifyTimeout 限制失败路径上发送 SESSION_ERROR 的时间：会话已经判定失败，只是尽力通知。
const failNotifyTimeout = time.Second

// Expect 是 Core 启动记录中与本次进程绑定的身份信息。
type Expect struct {
	// PluginID 是 Core 启动该进程时的插件 ID。
	PluginID string
	// Secret 是本次启动的 32 字节 Session Secret。
	Secret []byte
	// BuildDigest 是安装时已验证的归档摘要。
	BuildDigest []byte
	// Offer 是 Core 一侧支持/要求的协议版本与能力。
	Offer validate.Offer
}

// PhaseError 是准备阶段的失败：哪个阶段、什么错误码。消息不含 Secret 或配置内容。
type PhaseError struct {
	// Phase 是失败的阶段。
	Phase pb.SessionPhase
	// Code 是公开的错误码。
	Code pb.ErrorCode
	// Msg 是不含敏感信息的说明。
	Msg string
}

// Error 返回「阶段: 错误码: 说明」。
func (e *PhaseError) Error() string { return fmt.Sprintf("%s: %s: %s", e.Phase, e.Code, e.Msg) }

// phaseErr 用格式化说明构造 *PhaseError。
func phaseErr(phase pb.SessionPhase, code pb.ErrorCode, format string, args ...any) *PhaseError {
	return &PhaseError{Phase: phase, Code: code, Msg: fmt.Sprintf(format, args...)}
}

// Prepared 是 REGISTER 校验通过后的结果；在 Core 提交 ActiveGeneration 之前不对外公开任何路由。
type Prepared struct {
	// SessionID 是本 Session 的 ID。
	SessionID []byte
	// Minor 是协商出的协议次版本。
	Minor uint32
	// Capabilities 是双方都支持的协议能力。
	Capabilities []*pb.ProtocolCapability
	// Register 是已通过校验的注册载荷。
	Register *pb.Register
	// RouteHash 是 PREPARED 中的 route_descriptor_hash。
	RouteHash []byte
	// ConfigDigest 是 PREPARED 中的 candidate_config_digest。
	ConfigDigest []byte
}

// Prepare 在一条刚建立的连接上完成 HELLO → HELLO_ACK(Bootstrap) → REGISTER → PREPARED。
// 认证失败直接关闭且不回复细节；认证之后的失败发送 SESSION_ERROR 再关闭。任何失败都不留下注册结果。
// 成功返回后连接交给 wire.Session 继续（ACTIVATE 见 G1-17）。
func Prepare(ctx context.Context, conn *wire.Conn, exp Expect, boot *pb.Bootstrap, lim Limits) (*Prepared, error) {
	// watchdog：阶段超时即关闭连接，阻塞中的读写随之返回；stage 每次送入新阶段的预算。
	stage := make(chan time.Duration, 1)
	done := make(chan struct{})
	defer close(done)
	timedOut := make(chan struct{})
	go func() {
		d := <-stage
		t := time.NewTimer(d)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				_ = conn.Close()
				return
			case d = <-stage:
				t.Reset(d)
			case <-t.C:
				close(timedOut)
				_ = conn.Close()
				return
			}
		}
	}()
	stage <- lim.Handshake

	// fail 把读写失败转为该阶段的 PhaseError：已超时报 DEADLINE_EXCEEDED，帧超限报 RESOURCE_EXHAUSTED。
	fail := func(phase pb.SessionPhase, err error) error {
		select {
		case <-timedOut:
			return phaseErr(phase, pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "阶段超过预算")
		default:
		}
		if errors.Is(err, wire.ErrFrameTooLarge) {
			return phaseErr(phase, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "帧超过 4 MiB")
		}
		var pe *PhaseError
		if errors.As(err, &pe) {
			return pe
		}
		return phaseErr(phase, pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "%v", err)
	}

	// 1. HELLO：首帧必须是 HELLO，身份与 Secret 匹配。
	env, err := conn.ReadFrame()
	if err != nil {
		_ = conn.Close()
		return nil, fail(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, err)
	}
	msg, err := validate.DecodeControl(env)
	hello, isHello := msg.(*pb.Hello)
	if err != nil || !isHello {
		_ = conn.Close()
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "首帧不是合法的 HELLO")
	}
	if !authentic(env, hello, exp) {
		_ = conn.Close() // 认证失败：不回复，不暴露哪一项不符
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_UNAUTHENTICATED, "插件身份认证失败")
	}

	sid := make([]byte, 16)
	if _, err := rand.Read(sid); err != nil {
		_ = conn.Close()
		return nil, err
	}
	// sendErr 认证之后的失败路径：限时尽力发出 SESSION_ERROR 后关闭连接。
	sendErr := func(phase pb.SessionPhase, kind pb.FrameKind, code pb.ErrorCode, safe string) {
		if e, err := wire.ControlEnvelope(sid, 0, "", pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: phase, FailedKind: kind}); err == nil {
			e.Error = &pb.ErrorInfo{Code: code, SafeMessage: safe}
			_ = conn.WriteFrameTimeout(e, failNotifyTimeout) // 尽力通知；对端不读时不阻塞失败路径
		}
		_ = conn.Close()
	}

	agree, err := validate.Negotiate(exp.Offer, validate.Offer{
		Major: hello.GetSupportedMajor(), Minor: hello.GetSupportedMinor(),
		Supported: hello.GetSupportedCapabilities(), Required: hello.GetRequiredCapabilities()})
	if err != nil {
		sendErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.FrameKind_FRAME_KIND_HELLO, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "协议版本或必需能力不兼容")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "%v", err)
	}

	// 2. HELLO_ACK + Bootstrap：Core 只检查配置是 JSON 对象，不先用新 Schema 否决旧配置。
	if !validate.IsJSONObject([]byte(boot.GetCandidateConfigJson())) {
		sendErr(pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "候选配置不是 JSON 对象")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_CONFIG, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "候选配置不是 JSON 对象")
	}
	prepBudget := lim.Config + lim.Migration + lim.OnStart // Preparing 的硬截止为各适用阶段预算之和
	boot = proto.Clone(boot).(*pb.Bootstrap)
	boot.PhaseLimits = lim.proto()
	boot.PreparationDeadlineUnixMs = time.Now().Add(prepBudget).UnixMilli()
	ack := &pb.HelloAck{SessionId: sid, ChosenMinor: agree.Minor, CapabilityIntersection: agree.Intersection, Bootstrap: boot}
	ackEnv, err := wire.ControlEnvelope(sid, agree.Minor, "", pb.FrameKind_FRAME_KIND_HELLO_ACK, ack)
	if err == nil {
		err = conn.WriteFrame(ackEnv)
	}
	if errors.Is(err, wire.ErrFrameTooLarge) {
		sendErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.FrameKind_FRAME_KIND_HELLO_ACK, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "Bootstrap 超过 4 MiB")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "Bootstrap 超过 4 MiB，拒绝启动")
	} else if err != nil {
		_ = conn.Close()
		return nil, fail(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, err)
	}

	// 3. 准备阶段：等待 REGISTER（或插件报告的 SESSION_ERROR）。准备期不接受业务帧。
	stage <- prepBudget
	env, err = conn.ReadFrame()
	if err != nil {
		_ = conn.Close()
		return nil, fail(pb.SessionPhase_SESSION_PHASE_REGISTER, err)
	}
	if err := wire.CheckPeer(env, sid, agree.Minor, exp.PluginID); err != nil {
		sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, env.GetKind(), pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "帧的会话身份或协商版本不符")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "%v", err)
	}
	msg, err = validate.DecodeControl(env)
	if err != nil {
		sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, env.GetKind(), pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "准备阶段收到非法帧")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "准备阶段收到非法帧: %v", err)
	}
	switch m := msg.(type) {
	case *pb.SessionError:
		_ = conn.Close()
		return nil, &PhaseError{Phase: m.GetFailedPhase(), Code: validate.PublicCode(env.GetError().GetCode()), Msg: env.GetError().GetSafeMessage()}
	case *pb.Register:
		if err := validate.Register(m); err != nil {
			var ve *validate.Error
			code := pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			if errors.As(err, &ve) {
				code = ve.Code
			}
			sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, code, "注册描述不合法")
			return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, code, "%v", err)
		}
		if l := m.GetHttpListener(); l != nil && boot.GetHttpProxyAuth() == nil {
			sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "未声明 HTTP 能力却上报了 Listener")
			return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "未声明 HTTP 能力却上报了 Listener")
		}
		p := &Prepared{SessionID: sid, Minor: agree.Minor, Capabilities: agree.Intersection, Register: m,
			RouteHash: routeHash(m), ConfigDigest: digest([]byte(m.GetCandidateConfigJson()))}
		prep, err := wire.ControlEnvelope(sid, agree.Minor, "", pb.FrameKind_FRAME_KIND_PREPARED, &pb.Prepared{RouteDescriptorHash: p.RouteHash, CandidateConfigDigest: p.ConfigDigest})
		if err == nil {
			err = conn.WriteFrame(prep)
		}
		if err != nil {
			_ = conn.Close()
			return nil, fail(pb.SessionPhase_SESSION_PHASE_REGISTER, err)
		}
		return p, nil
	default:
		sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, env.GetKind(), pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "准备阶段只接受 REGISTER 或 SESSION_ERROR")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "准备阶段收到 %s", env.GetKind())
	}
}

// authentic 用恒时比较核对 Secret，并核对插件 ID 与 build_digest；不区分是哪一项失败。
func authentic(env *pb.Envelope, h *pb.Hello, exp Expect) bool {
	secretOK := len(exp.Secret) == 32 && subtle.ConstantTimeCompare(h.GetSecret(), exp.Secret) == 1
	idOK := h.GetPluginId() == exp.PluginID && env.GetSourcePluginId() == exp.PluginID
	digestOK := subtle.ConstantTimeCompare(h.GetBuildDigest(), exp.BuildDigest) == 1
	return secretOK && idOK && digestOK
}

// routeHash 按顺序拼接各 RouteDescriptor 的确定性编码并计算 SHA256。
func routeHash(r *pb.Register) []byte {
	h := sha256.New()
	for _, rd := range r.GetRoutes() {
		b, _ := proto.MarshalOptions{Deterministic: true}.Marshal(rd)
		h.Write(b)
	}
	return h.Sum(nil)
}

// digest 返回 b 的 SHA256。
func digest(b []byte) []byte {
	s := sha256.Sum256(b)
	return s[:]
}
