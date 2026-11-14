// Package host 编排 Core 侧单个插件 Session 的生命周期。
//
// 一个插件进程连上 Core 后，依次经过握手认证、Bootstrap 下发、REGISTER 校验与 PREPARED（Prepare），
// 提交 ActiveGeneration、ACTIVATE 与 ACTIVATED（Activate），以及冻结、排空、STOPPING 与结束进程树（Stop）。
// 有限重启策略也在这里。状态与可见性规则见 spec/CONTRACTS-v0.1.md §4 与 §12.1。
//
// 本包不启动进程，也不持久化状态；这些由 internal/core/process、internal/core/state 和
// internal/core/supervisor 负责。
//
// 包的组成：
//   - [Prepare] 完成握手与准备阶段，返回 [Prepared]；阶段预算由 [Limits] 给出，失败时返回 [PhaseError]。
//   - [Activate] 提交代次并完成激活，返回跟踪实例状态的 [Instance]。
//   - [Instance.Stop] 与 [StopAll] 按 [StopTiming] 停止实例；[RestartPolicy] 决定进程退出后是否重启。
package host

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"slices"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Limits 规定阶段预算与双方共用的 Session 限额（CONTRACTS §4.2、§16）。
type Limits struct {
	// 从连接建立到 HELLO_ACK 发出的预算（CONTRACTS §4.2 的 Hello/HelloAck）。
	Handshake          time.Duration
	Config             time.Duration // 配置转换与校验的预算
	Migration          time.Duration // 数据库迁移的预算
	OnStart            time.Duration
	OnReady            time.Duration
	StopTotal          time.Duration // 一次停用的总预算
	SessionMaxInflight int
	SessionMaxPayload  int64 // 每方向在途与排队载荷的字节预算
	RequestTimeout     time.Duration
}

// NewLimits 从本次运行的生效调优值构造阶段预算与 Session 限额。
func NewLimits(v tune.Values) Limits {
	return Limits{Handshake: v.PhaseHello, Config: v.PhaseConfig, Migration: v.PhaseMigration,
		OnStart: v.PhaseOnStart, OnReady: v.PhaseOnReady, StopTotal: v.StopTotal,
		SessionMaxInflight: v.SessionMaxInflight, SessionMaxPayload: v.SessionMaxPayload, RequestTimeout: v.RequestTimeout}
}

// proto 把预算转为 Bootstrap 中的毫秒值。
func (l Limits) proto() *pb.PhaseLimits {
	ms := func(d time.Duration) uint32 { return uint32(d / time.Millisecond) }
	return &pb.PhaseLimits{HandshakeMs: ms(l.Handshake), ConfigMs: ms(l.Config), MigrationMs: ms(l.Migration),
		OnStartMs: ms(l.OnStart), OnReadyMs: ms(l.OnReady), StopTotalMs: ms(l.StopTotal)}
}

// failNotifyTimeout 限制失败路径上发送 SESSION_ERROR 的时间：会话已经判定失败，只是尽力通知。
const failNotifyTimeout = time.Second

// HTTP 代理材料与 WIRE §3.1 的 http_listener v1 协议一致。
const (
	httpListenerVersion = 1
	httpProxyTokenBytes = 32
)

// Expect 携带本次进程的可信身份及配置快照回调。
type Expect struct {
	Role        state.Role     // 必填，由安装记录指定
	AdapterID   string         // Adapter 角色要求的平台注册身份
	PluginID    string         // Core 启动该进程时的插件 ID
	Secret      []byte         // 本次启动的 32 字节 Session Secret
	BuildDigest []byte         // 安装时已验证的归档摘要
	Offer       validate.Offer // Core 一侧支持/要求的协议版本与能力
	// Snapshot 在 REGISTER 校验后生成持久化快照；nil 时使用原候选配置。
	// 回调的错误消息须省略配置值。
	Snapshot func(candidate []byte) (snapshot []byte, err error)
}

// PhaseError 记录准备阶段在哪个阶段、以什么错误码失败；消息不含 Secret 或配置内容。
type PhaseError struct {
	Phase   pb.SessionPhase // 失败的阶段
	Code    pb.ErrorCode    // 公开的错误码
	Message string          // 不含敏感信息的说明
}

// Error 返回「阶段: 错误码: 说明」。
func (e *PhaseError) Error() string { return fmt.Sprintf("%s: %s: %s", e.Phase, e.Code, e.Message) }

// phaseErr 用格式化说明构造 *PhaseError。
func phaseErr(phase pb.SessionPhase, code pb.ErrorCode, format string, args ...any) *PhaseError {
	return &PhaseError{Phase: phase, Code: code, Message: fmt.Sprintf(format, args...)}
}

// Prepared 保存 REGISTER 校验通过后的结果；Core 提交 ActiveGeneration 之前不对外公开任何路由。
type Prepared struct {
	SessionID      []byte
	Minor          uint32                   // 协商出的协议次版本
	Capabilities   []*pb.ProtocolCapability // 双方都支持的协议能力
	Register       *pb.Register             // 已通过校验的注册载荷
	RouteHash      []byte                   // PREPARED 中的 route_descriptor_hash
	ConfigSnapshot []byte                   // 准备持久化的配置，secret 字段只包含引用
	// PREPARED 中的 candidate_config_digest，对 ConfigSnapshot 计算。
	ConfigDigest   []byte
	httpProxyToken []byte // 本 Session 的内部代理凭证；仅交给 Instance，不进入日志
	defaultLocale  string // 本 Session 下发的 Bootstrap 默认语言快照
}

// Prepare 在一条刚建立的连接上完成 HELLO → HELLO_ACK(Bootstrap) → REGISTER → PREPARED。
// 认证失败直接关闭且不回复细节；认证之后的失败发送 SESSION_ERROR 再关闭。任何失败都不留下注册结果。
// 成功返回后连接交给 wire.Session 继续（ACTIVATE 见 activate.go）。
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
			return phaseErr(phase, pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "phase exceeded its budget")
		default:
		}
		if errors.Is(err, wire.ErrFrameTooLarge) {
			return phaseErr(phase, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "frame exceeds 4 MiB")
		}
		var pe *PhaseError
		if errors.As(err, &pe) {
			return pe
		}
		if err := ctx.Err(); err != nil {
			return err
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
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "first frame is not a valid HELLO")
	}
	if !authentic(env, hello, exp) {
		_ = conn.Close() // 认证失败：不回复，不暴露哪一项不符
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_UNAUTHENTICATED, "plugin authentication failed")
	}

	sid := make([]byte, validate.SessionIDLen)
	if _, err := rand.Read(sid); err != nil {
		_ = conn.Close()
		return nil, err
	}
	// minor 在协商成功后更新为协商值，之后的 SESSION_ERROR 都使用它（WIRE §1）。
	var minor uint32
	// sendErr 认证之后的失败路径：限时尽力发出 SESSION_ERROR 后关闭连接。
	sendErr := func(phase pb.SessionPhase, kind pb.FrameKind, code pb.ErrorCode, safe string) {
		if e, err := wire.ControlEnvelope(sid, minor, "", pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: phase, FailedKind: kind}); err == nil {
			e.Error = &pb.ErrorInfo{Code: code, SafeMessage: safe}
			_ = conn.WriteFrameTimeout(e, failNotifyTimeout) // 尽力通知；对端不读时不阻塞失败路径
		}
		_ = conn.Close()
	}

	offer := exp.Offer
	offer.Supported = append(slices.Clone(offer.Supported),
		&pb.ProtocolCapability{Name: wire.CapabilityHTTPListener, Version: httpListenerVersion})
	agree, err := validate.Negotiate(offer, validate.Offer{
		Major: hello.GetSupportedMajor(), Minor: hello.GetSupportedMinor(),
		Supported: hello.GetSupportedCapabilities(), Required: hello.GetRequiredCapabilities()})
	if err != nil {
		sendErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.FrameKind_FRAME_KIND_HELLO, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "incompatible protocol version or required capabilities")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "%v", err)
	}
	minor = agree.Minor

	// 2. HELLO_ACK + Bootstrap：Core 只检查配置是 JSON 对象，不先用新 Schema 否决旧配置。
	if !validate.IsJSONObject([]byte(boot.GetCandidateConfigJson())) {
		sendErr(pb.SessionPhase_SESSION_PHASE_CONFIG, pb.FrameKind_FRAME_KIND_HELLO_ACK, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "candidate config is not a JSON object")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_CONFIG, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "candidate config is not a JSON object")
	}
	prepBudget := lim.Config + lim.Migration + lim.OnStart // Preparing 的硬截止为各适用阶段预算之和
	boot = proto.Clone(boot).(*pb.Bootstrap)
	boot.HttpProxyAuth = nil
	for _, capability := range agree.Intersection {
		if capability.GetName() == wire.CapabilityHTTPListener && capability.GetVersion() == httpListenerVersion {
			token := make([]byte, httpProxyTokenBytes)
			if _, err := rand.Read(token); err != nil {
				sendErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.FrameKind_FRAME_KIND_HELLO_ACK, pb.ErrorCode_ERROR_CODE_INTERNAL, "generate HTTP proxy credential failed")
				return nil, fmt.Errorf("generate HTTP proxy credential: %w", err)
			}
			boot.HttpProxyAuth = &pb.HttpProxyAuth{Token: token}
			break
		}
	}
	boot.PhaseLimits = lim.proto()
	boot.SessionLimits = &pb.SessionLimits{MaxInflight: uint32(lim.SessionMaxInflight), MaxPayloadBytes: uint32(lim.SessionMaxPayload), RequestTimeoutMs: uint32(lim.RequestTimeout / time.Millisecond)}
	boot.PreparationDeadlineUnixMs = time.Now().Add(prepBudget).UnixMilli()
	ack := &pb.HelloAck{SessionId: sid, ChosenMinor: agree.Minor, CapabilityIntersection: agree.Intersection, Bootstrap: boot}
	ackEnv, err := wire.ControlEnvelope(sid, agree.Minor, "", pb.FrameKind_FRAME_KIND_HELLO_ACK, ack)
	if err == nil {
		err = conn.WriteFrame(ackEnv)
	}
	if errors.Is(err, wire.ErrFrameTooLarge) {
		sendErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.FrameKind_FRAME_KIND_HELLO_ACK, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "Bootstrap exceeds 4 MiB")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_HANDSHAKE, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "Bootstrap exceeds 4 MiB, refusing to start")
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
		sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, env.GetKind(), pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "frame session identity or negotiated version mismatch")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "%v", err)
	}
	msg, err = validate.DecodeControl(env)
	if err != nil {
		sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, env.GetKind(), pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "invalid frame during preparation")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "invalid frame during preparation: %v", err)
	}
	switch m := msg.(type) {
	case *pb.SessionError:
		_ = conn.Close()
		return nil, &PhaseError{Phase: m.GetFailedPhase(), Code: validate.PublicCode(env.GetError().GetCode()), Message: env.GetError().GetSafeMessage()}
	case *pb.Register:
		var registerErr error
		if exp.Role == state.RoleAdapter {
			registerErr = validate.AdapterRegister(m, exp.AdapterID)
			if registerErr == nil && boot.GetDatabaseDescriptor().GetBackend() != pb.DatabaseBackend_DATABASE_BACKEND_NONE {
				registerErr = &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "adapter registration contains a business database"}
			}
		} else if exp.Role == state.RolePlugin || exp.Role == state.RoleAgent {
			registerErr = validate.Register(m)
		} else {
			registerErr = fmt.Errorf("invalid expected role %q", exp.Role)
		}
		if registerErr == nil && len(boot.GetConfigSchemaHash()) != 0 && !bytes.Equal(m.GetConfigSchemaHash(), boot.GetConfigSchemaHash()) {
			registerErr = &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "registered config schema hash differs from Bootstrap"}
		}
		if registerErr == nil && m.GetCandidateConfigVersion() != boot.GetTargetConfigVersion() {
			registerErr = &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "registered config version differs from Bootstrap"}
		}
		if registerErr == nil {
			registerErr = checkMigration(m.GetMigration(), boot.GetDatabaseDescriptor())
		}
		if err := registerErr; err != nil {
			var ve *validate.Error
			code := pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			if errors.As(err, &ve) {
				code = ve.Code
			}
			sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, code, "invalid registration")
			return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, code, "%v", err)
		}
		if l := m.GetHttpListener(); l != nil && boot.GetHttpProxyAuth() == nil {
			sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "listener reported without declaring HTTP capability")
			return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "listener reported without declaring HTTP capability")
		}
		// 4. 注册校验通过后形成持久化快照，PREPARED 只确认该快照的摘要。
		snapshot := []byte(m.GetCandidateConfigJson())
		if exp.Snapshot != nil {
			snapshot, err = exp.Snapshot(snapshot)
			if err != nil {
				code := pb.ErrorCode_ERROR_CODE_INTERNAL
				safe := "create config snapshot failed"
				var ve *validate.Error
				if errors.As(err, &ve) {
					code, safe = validate.PublicCode(ve.Code), ve.Message
				}
				sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.FrameKind_FRAME_KIND_REGISTER, code, safe)
				return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, code, "%v", err)
			}
		}
		p := &Prepared{SessionID: sid, Minor: agree.Minor, Capabilities: agree.Intersection, Register: m,
			RouteHash: routeHash(m), ConfigSnapshot: snapshot, ConfigDigest: digest(snapshot), httpProxyToken: boot.GetHttpProxyAuth().GetToken(), defaultLocale: boot.GetDefaultLocale()}
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
		sendErr(pb.SessionPhase_SESSION_PHASE_REGISTER, env.GetKind(), pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "only REGISTER or SESSION_ERROR is accepted during preparation")
		return nil, phaseErr(pb.SessionPhase_SESSION_PHASE_REGISTER, pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, "received %s during preparation", env.GetKind())
	}
}

func checkMigration(migration *pb.MigrationState, database *pb.DatabaseDescriptor) error {
	switch migration.GetMode() {
	case pb.MigrationMode_MIGRATION_MODE_NONE, pb.MigrationMode_MIGRATION_MODE_VERSIONED, pb.MigrationMode_MIGRATION_MODE_SELF:
	default:
		return fmt.Errorf("invalid registered migration mode %s", migration.GetMode())
	}
	if database.GetMigrationMode() != pb.MigrationMode_MIGRATION_MODE_UNSPECIFIED && migration.GetMode() != database.GetMigrationMode() {
		return fmt.Errorf("registered migration mode %s does not match Bootstrap mode %s", migration.GetMode(), database.GetMigrationMode())
	}
	if migration.GetMode() == pb.MigrationMode_MIGRATION_MODE_VERSIONED {
		completed := make(map[uint32]bool, len(migration.GetCompletedMigrationIds()))
		for _, id := range migration.GetCompletedMigrationIds() {
			completed[id] = true
		}
		for _, id := range database.GetAppliedMigrationIds() {
			if !completed[id] {
				return fmt.Errorf("registered migrations omit previously applied ID %d", id)
			}
		}
	}
	return nil
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
