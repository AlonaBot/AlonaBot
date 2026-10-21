package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

var (
	secretA = bytes.Repeat([]byte{0xA1}, 32)
	secretB = bytes.Repeat([]byte{0xB2}, 32)
	digestA = bytes.Repeat([]byte{0xD1}, 32)
)

func expectA() Expect {
	return Expect{PluginID: "plugin-a", Secret: secretA, BuildDigest: digestA,
		Offer: validate.Offer{Major: 1, Minor: 0, Supported: []*pb.ProtocolCapability{{Name: "core.send", Version: 1}}}}
}

func bootA() *pb.Bootstrap {
	return &pb.Bootstrap{GenerationId: "gen-1", CandidateConfigJson: `{"token":"CFG-SECRET-VALUE"}`,
		DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE}}
}

var fastLimits = Limits{Handshake: 500 * time.Millisecond, Config: 200 * time.Millisecond, Migration: 200 * time.Millisecond,
	OnStart: 200 * time.Millisecond, OnReady: 200 * time.Millisecond, StopTotal: time.Second}

// fake 是用原始帧扮演插件的一端。
type fake struct {
	t    *testing.T
	conn *wire.Conn
	raw  net.Conn
	sid  []byte
}

func (f *fake) send(kind pb.FrameKind, msg proto.Message, mutate func(*pb.Envelope)) {
	f.t.Helper()
	data, _ := proto.Marshal(msg)
	env := &pb.Envelope{ProtocolMajor: 1, SessionId: f.sid, Kind: kind, SourcePluginId: "plugin-a",
		PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaId: validate.SchemaID(msg), SchemaVersion: 1, Payload: data}
	if mutate != nil {
		mutate(env)
	}
	if err := f.conn.WriteFrame(env); err != nil {
		f.t.Fatalf("发送 %s: %v", kind, err)
	}
}

func (f *fake) recv() (*pb.Envelope, proto.Message, error) {
	env, err := f.conn.ReadFrame()
	if err != nil {
		return nil, nil, err
	}
	msg, err := validate.DecodeControl(env)
	return env, msg, err
}

func hello(secret []byte, major uint32, required ...*pb.ProtocolCapability) *pb.Hello {
	return &pb.Hello{PluginId: "plugin-a", Secret: secret, SupportedMajor: major, BuildDigest: digestA,
		SupportedCapabilities: []*pb.ProtocolCapability{{Name: "core.send", Version: 1}}, RequiredCapabilities: required}
}

func cmd(id string) *pb.RouteDescriptor {
	return &pb.RouteDescriptor{RouteId: id, Kind: pb.RouteKind_ROUTE_KIND_COMMAND,
		Spec: &pb.RouteDescriptor_Command{Command: &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH, Path: []string{id}}}}
}

// run 启动 Core 侧 Prepare，返回假插件与结果通道。
func run(t *testing.T, exp Expect, boot *pb.Bootstrap, lim Limits) (*fake, <-chan result) {
	t.Helper()
	a, b := net.Pipe()
	t.Cleanup(func() { _ = a.Close(); _ = b.Close() })
	out := make(chan result, 1)
	go func() {
		p, err := Prepare(context.Background(), wire.NewStreamConn(a), exp, boot, lim)
		out <- result{p, err}
	}()
	return &fake{t: t, conn: wire.NewStreamConn(b), raw: b}, out
}

type result struct {
	p   *Prepared
	err error
}

func phaseOf(t *testing.T, err error) *PhaseError {
	t.Helper()
	var pe *PhaseError
	if !errors.As(err, &pe) {
		t.Fatalf("期望 *PhaseError，实际 %v", err)
	}
	return pe
}

func TestPrepareHappyPath(t *testing.T) {
	f, out := run(t, expectA(), bootA(), DefaultLimits)
	f.send(pb.FrameKind_FRAME_KIND_HELLO, hello(secretA, 1), nil)
	env, msg, err := f.recv()
	if err != nil || env.GetKind() != pb.FrameKind_FRAME_KIND_HELLO_ACK {
		t.Fatalf("应收到 HELLO_ACK: %v %v", env.GetKind(), err)
	}
	ack := msg.(*pb.HelloAck)
	f.sid = ack.GetSessionId()
	b := ack.GetBootstrap()
	if len(f.sid) != 16 || b.GetGenerationId() != "gen-1" || b.GetCandidateConfigJson() != `{"token":"CFG-SECRET-VALUE"}` ||
		b.GetPhaseLimits().GetHandshakeMs() != 10000 || b.GetPhaseLimits().GetOnStartMs() != 60000 || b.GetPreparationDeadlineUnixMs() == 0 {
		t.Fatalf("Bootstrap 内容不符: %v", b)
	}
	f.send(pb.FrameKind_FRAME_KIND_REGISTER, &pb.Register{Routes: []*pb.RouteDescriptor{cmd("hello")}, CandidateConfigJson: `{}`}, nil)
	env, msg, err = f.recv()
	if err != nil || env.GetKind() != pb.FrameKind_FRAME_KIND_PREPARED {
		t.Fatalf("应收到 PREPARED: %v %v", env.GetKind(), err)
	}
	r := <-out
	if r.err != nil {
		t.Fatal(r.err)
	}
	prep := msg.(*pb.Prepared)
	if !bytes.Equal(prep.GetRouteDescriptorHash(), r.p.RouteHash) || len(r.p.ConfigDigest) != 32 || r.p.Register.GetRoutes()[0].GetRouteId() != "hello" {
		t.Fatalf("PREPARED 与返回结果不一致: %+v", r.p)
	}
}

// 认证失败：Core 不回复任何帧，直接关闭（不暴露是哪一项不符）。
func TestPrepareRejectsBadIdentity(t *testing.T) {
	cases := map[string]func(*pb.Envelope){
		"伪造 Secret":       func(e *pb.Envelope) { e.Payload, _ = proto.Marshal(hello(bytes.Repeat([]byte{9}, 32), 1)) },
		"其他插件的 Secret":    func(e *pb.Envelope) { e.Payload, _ = proto.Marshal(hello(secretB, 1)) },
		"插件 ID 不符":        func(e *pb.Envelope) { e.SourcePluginId = "plugin-b" },
		"build digest 不符": func(e *pb.Envelope) { h := hello(secretA, 1); h.BuildDigest = secretB; e.Payload, _ = proto.Marshal(h) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f, out := run(t, expectA(), bootA(), DefaultLimits)
			f.send(pb.FrameKind_FRAME_KIND_HELLO, hello(secretA, 1), mutate)
			if _, _, err := f.recv(); !errors.Is(err, io.EOF) && !errors.Is(err, wire.ErrTruncated) {
				t.Fatalf("认证失败后插件不应收到任何帧，实际 %v", err)
			}
			r := <-out
			if pe := phaseOf(t, r.err); pe.Code != pb.ErrorCode_ERROR_CODE_UNAUTHENTICATED || r.p != nil {
				t.Fatalf("应为 UNAUTHENTICATED 且无结果: %v", r.err)
			}
			if strings.Contains(r.err.Error(), hex.EncodeToString(secretA)) {
				t.Fatal("错误信息泄漏了 Secret")
			}
		})
	}
}

// 认证通过但协议不兼容：回 SESSION_ERROR(HANDSHAKE) 后关闭。
func TestPrepareProtocolMismatch(t *testing.T) {
	cases := map[string]*pb.Hello{
		"Major 不一致": hello(secretA, 2),
		"未知必需能力":    hello(secretA, 1, &pb.ProtocolCapability{Name: "agent.v2", Version: 1}),
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			f, out := run(t, expectA(), bootA(), DefaultLimits)
			f.send(pb.FrameKind_FRAME_KIND_HELLO, h, nil)
			env, msg, err := f.recv()
			if err != nil || env.GetKind() != pb.FrameKind_FRAME_KIND_SESSION_ERROR ||
				msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_HANDSHAKE ||
				env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
				t.Fatalf("应收到 SESSION_ERROR(HANDSHAKE, FAILED_PRECONDITION): %v %v", env, err)
			}
			if pe := phaseOf(t, (<-out).err); pe.Code != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
				t.Fatalf("Prepare 错误 %v", pe)
			}
		})
	}
}

// handshake 后送出 HELLO 并读取 HELLO_ACK，返回已就绪的假插件。
func handshake(t *testing.T, f *fake) {
	t.Helper()
	f.send(pb.FrameKind_FRAME_KIND_HELLO, hello(secretA, 1), nil)
	_, msg, err := f.recv()
	if err != nil {
		t.Fatal(err)
	}
	f.sid = msg.(*pb.HelloAck).GetSessionId()
}

func TestPrepareRegisterRejectedAtomically(t *testing.T) {
	f, out := run(t, expectA(), bootA(), DefaultLimits)
	handshake(t, f)
	f.send(pb.FrameKind_FRAME_KIND_REGISTER, &pb.Register{Routes: []*pb.RouteDescriptor{cmd("a"), cmd("b"), cmd("a")}, CandidateConfigJson: "{}"}, nil)
	env, msg, err := f.recv()
	if err != nil || msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_REGISTER || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
		t.Fatalf("重复 route 应收到 SESSION_ERROR(REGISTER, ALREADY_EXISTS): %v %v", env, err)
	}
	r := <-out
	if r.p != nil || phaseOf(t, r.err).Code != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
		t.Fatalf("注册失败不得返回部分结果: %+v %v", r.p, r.err)
	}
}

// 插件发来超过 4 MiB 的 REGISTER：Core 在读取载荷前拒绝，按 RESOURCE_EXHAUSTED 失败。
func TestPrepareRegisterTooLarge(t *testing.T) {
	f, out := run(t, expectA(), bootA(), DefaultLimits)
	handshake(t, f)
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], wire.MaxFrame+1)
	go func() { _, _ = f.raw.Write(hdr[:]) }()
	if pe := phaseOf(t, (<-out).err); pe.Code != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED || pe.Phase != pb.SessionPhase_SESSION_PHASE_REGISTER {
		t.Fatalf("超限 REGISTER 应为 RESOURCE_EXHAUSTED，实际 %v", pe)
	}
}

func TestPrepareBootstrapTooLarge(t *testing.T) {
	boot := bootA()
	boot.CandidateConfigJson = `{"x":"` + strings.Repeat("a", wire.MaxFrame) + `"}`
	f, out := run(t, expectA(), boot, DefaultLimits)
	f.send(pb.FrameKind_FRAME_KIND_HELLO, hello(secretA, 1), nil)
	env, _, err := f.recv()
	if err != nil || env.GetKind() != pb.FrameKind_FRAME_KIND_SESSION_ERROR || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("超限 Bootstrap 应收到 SESSION_ERROR(RESOURCE_EXHAUSTED): %v %v", env, err)
	}
	if pe := phaseOf(t, (<-out).err); pe.Code != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("Prepare 错误 %v", pe)
	}
}

// 插件在 OnStart 失败时报告 SESSION_ERROR：Core 按插件报告的阶段返回。
func TestPreparePluginReportsOnStartError(t *testing.T) {
	f, out := run(t, expectA(), bootA(), DefaultLimits)
	handshake(t, f)
	f.send(pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: pb.SessionPhase_SESSION_PHASE_ON_START},
		func(e *pb.Envelope) {
			e.Error = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "OnStart 返回错误"}
		})
	pe := phaseOf(t, (<-out).err)
	if pe.Phase != pb.SessionPhase_SESSION_PHASE_ON_START || pe.Code != pb.ErrorCode_ERROR_CODE_INTERNAL {
		t.Fatalf("应返回 ON_START/INTERNAL，实际 %v", pe)
	}
}

func TestPrepareTimeouts(t *testing.T) {
	t.Run("握手超时", func(t *testing.T) {
		start := time.Now()
		_, out := run(t, expectA(), bootA(), fastLimits) // 插件一直不发 HELLO
		pe := phaseOf(t, (<-out).err)
		if pe.Code != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED || pe.Phase != pb.SessionPhase_SESSION_PHASE_HANDSHAKE {
			t.Fatalf("应为握手超时，实际 %v", pe)
		}
		if el := time.Since(start); el > 2*time.Second {
			t.Fatalf("超时判定耗时 %v，远超预算", el)
		}
	})
	t.Run("准备阶段超时（OnStart 挂起）", func(t *testing.T) {
		f, out := run(t, expectA(), bootA(), fastLimits)
		handshake(t, f) // 之后插件一直不发 REGISTER
		pe := phaseOf(t, (<-out).err)
		if pe.Code != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
			t.Fatalf("应为准备阶段超时，实际 %v", pe)
		}
	})
}

func TestPrepareRejectsBusinessFrameDuringPreparing(t *testing.T) {
	f, out := run(t, expectA(), bootA(), DefaultLimits)
	handshake(t, f)
	ms := time.Now().Add(time.Minute).UnixMilli()
	_ = f.conn.WriteFrame(&pb.Envelope{ProtocolMajor: 1, SessionId: f.sid, Kind: pb.FrameKind_FRAME_KIND_REQUEST, SourcePluginId: "plugin-a",
		InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, CallId: 1, DeadlineUnixMs: &ms})
	if pe := phaseOf(t, (<-out).err); pe.Code != pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR {
		t.Fatalf("准备期的业务 REQUEST 应被拒绝，实际 %v", pe)
	}
}

// 任何失败的错误信息都不包含 Secret 或配置内容（Bootstrap 不进入日志的前提）。
func TestErrorsDoNotLeakSecrets(t *testing.T) {
	f, out := run(t, expectA(), bootA(), DefaultLimits)
	handshake(t, f)
	f.send(pb.FrameKind_FRAME_KIND_REGISTER, &pb.Register{Routes: []*pb.RouteDescriptor{cmd("a"), cmd("a")}, CandidateConfigJson: `{"token":"CFG-SECRET-VALUE"}`}, nil)
	err := (<-out).err
	for _, s := range []string{"CFG-SECRET-VALUE", hex.EncodeToString(secretA), string(secretA)} {
		if strings.Contains(err.Error(), s) {
			t.Fatalf("错误信息包含敏感内容 %q: %v", s, err)
		}
	}
}

func TestDefaultLimitsMatchContract(t *testing.T) {
	want := Limits{10 * time.Second, 60 * time.Second, 15 * time.Minute, 60 * time.Second, 30 * time.Second, 10 * time.Second}
	if DefaultLimits != want {
		t.Fatalf("DefaultLimits = %+v，应与 CONTRACTS §4.2 一致", DefaultLimits)
	}
}
