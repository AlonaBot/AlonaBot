package sdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
	"google.golang.org/protobuf/proto"
)

var (
	testSecret = bytes.Repeat([]byte{0x5A}, 32)
	testDigest = bytes.Repeat([]byte{0x3C}, 32)
	testSID    = bytes.Repeat([]byte{0x11}, 16)
	testLaunch = launch{pluginID: "demo", transport: wire.TransportTCP, secret: testSecret, digest: testDigest}
)

// fakeCore 用原始帧扮演 Core 一端。
type fakeCore struct {
	t    *testing.T
	conn *wire.Conn
}

func (c *fakeCore) recv() (*pb.Envelope, proto.Message) {
	c.t.Helper()
	env, err := c.conn.ReadFrame()
	if err != nil {
		c.t.Fatalf("假 Core 读帧: %v", err)
	}
	msg, err := validate.DecodeControl(env)
	if err != nil {
		c.t.Fatalf("假 Core 解码: %v", err)
	}
	return env, msg
}

func (c *fakeCore) send(kind pb.FrameKind, msg proto.Message, mutate func(*pb.Envelope)) {
	data, _ := proto.Marshal(msg)
	env := &pb.Envelope{ProtocolMajor: 1, SessionId: testSID, Kind: kind, PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
		SchemaId: validate.SchemaID(msg), SchemaVersion: 1, Payload: data}
	if mutate != nil {
		mutate(env)
	}
	_ = c.conn.WriteFrame(env)
}

func newFakeCore(t *testing.T) (*fakeCore, *wire.Conn) {
	a, b := net.Pipe()
	t.Cleanup(func() { _ = a.Close(); _ = b.Close() })
	return &fakeCore{t: t, conn: wire.NewStreamConn(a)}, wire.NewStreamConn(b)
}

func ack(onStartMs uint32) *pb.HelloAck {
	return &pb.HelloAck{SessionId: testSID, Bootstrap: &pb.Bootstrap{GenerationId: "g1", CandidateConfigJson: `{"k":1}`,
		TargetConfigVersion: 1, PhaseLimits: &pb.PhaseLimits{OnStartMs: onStartMs},
		PreparationDeadlineUnixMs: time.Now().Add(5 * time.Second).UnixMilli()}}
}

type lifecycleRoot struct {
	plugin.Base
	onStart func(context.Context) error
}

func (r *lifecycleRoot) OnStart(ctx context.Context) error { return r.onStart(ctx) }
func (r *lifecycleRoot) Commands(c *CommandRegistry) {
	c.Command("ping", func(*Context) (message.Chain, error) { return nil, nil })
}

func TestClientHandshakeSendsIdentity(t *testing.T) {
	core, conn := newFakeCore(t)
	got := make(chan *pb.HelloAck, 1)
	go func() {
		a, err := clientHandshake(conn, testLaunch)
		if err != nil {
			t.Error(err)
		}
		got <- a
	}()
	env, msg := core.recv()
	h := msg.(*pb.Hello)
	if env.GetKind() != pb.FrameKind_FRAME_KIND_HELLO || len(env.GetSessionId()) != 0 || env.GetSourcePluginId() != "demo" ||
		h.GetPluginId() != "demo" || !bytes.Equal(h.GetSecret(), testSecret) || !bytes.Equal(h.GetBuildDigest(), testDigest) || h.GetSupportedMajor() != 1 {
		t.Fatalf("HELLO 内容不符: %v %v", env, h)
	}
	core.send(pb.FrameKind_FRAME_KIND_HELLO_ACK, ack(1000), nil)
	if a := <-got; a == nil || string(a.GetSessionId()) != string(testSID) {
		t.Fatalf("HELLO_ACK 解析失败: %v", a)
	}
}

func TestClientHandshakeRejectedByCore(t *testing.T) {
	core, conn := newFakeCore(t)
	errc := make(chan error, 1)
	go func() { _, err := clientHandshake(conn, testLaunch); errc <- err }()
	core.recv()
	_ = core.conn.Close() // 认证失败：Core 直接关闭
	if err := <-errc; err == nil {
		t.Fatal("Core 关闭连接时握手应失败")
	}
}

func prepareAsync(t *testing.T, conn *wire.Conn, a *pb.HelloAck, root plugin.Root) <-chan error {
	reg, err := Collect(root)
	if err != nil {
		t.Fatal(err)
	}
	errc := make(chan error, 1)
	go func() { _, err := prepare(conn, testLaunch, a, root, reg); errc <- err }()
	return errc
}

func TestPrepareHappyPath(t *testing.T) {
	core, conn := newFakeCore(t)
	started := false
	root := &lifecycleRoot{onStart: func(context.Context) error { started = true; return nil }}
	errc := prepareAsync(t, conn, ack(1000), root)
	env, msg := core.recv()
	reg := msg.(*pb.Register)
	if env.GetKind() != pb.FrameKind_FRAME_KIND_REGISTER || !started || reg.GetCandidateConfigJson() != `{"k":1}` ||
		len(reg.GetRoutes()) != 1 || reg.GetRoutes()[0].GetRouteId() != "cmd.ping" {
		t.Fatalf("REGISTER 不符（OnStart 必须先于 REGISTER）: started=%v %v", started, reg)
	}
	core.send(pb.FrameKind_FRAME_KIND_PREPARED, &pb.Prepared{RouteDescriptorHash: []byte{1}}, nil)
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
}

func TestPrepareOnStartErrorReportsSafeMessage(t *testing.T) {
	core, conn := newFakeCore(t)
	root := &lifecycleRoot{onStart: func(context.Context) error { return errors.New("db password=SECRET-XYZ rejected") }}
	errc := prepareAsync(t, conn, ack(1000), root)
	env, msg := core.recv()
	if msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_ON_START || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_INTERNAL {
		t.Fatalf("应报告 SESSION_ERROR(ON_START, INTERNAL): %v", env)
	}
	if strings.Contains(env.GetError().GetSafeMessage(), "SECRET-XYZ") {
		t.Fatal("SESSION_ERROR 的安全消息泄漏了用户错误原文")
	}
	if err := <-errc; err == nil || !strings.Contains(err.Error(), "SECRET-XYZ") {
		t.Fatalf("本地错误应保留原因用于 stderr 诊断: %v", err)
	}
}

// OnStart 不响应 context 时，由 watchdog 在预算到达时判定失败，而不是一直等下去。
func TestPrepareOnStartHangsBeyondBudget(t *testing.T) {
	core, conn := newFakeCore(t)
	block := make(chan struct{})
	defer close(block)
	root := &lifecycleRoot{onStart: func(context.Context) error { <-block; return nil }} // 忽略 ctx
	start := time.Now()
	errc := prepareAsync(t, conn, ack(100), root)
	env, msg := core.recv()
	if msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_ON_START || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		t.Fatalf("应报告 SESSION_ERROR(ON_START, DEADLINE_EXCEEDED): %v", env)
	}
	if err := <-errc; err == nil {
		t.Fatal("超时应返回错误")
	}
	if el := time.Since(start); el > 2*time.Second {
		t.Fatalf("watchdog 判定耗时 %v，远超 100ms 预算", el)
	}
}

type manyCommands struct{ plugin.Base }

func (manyCommands) Commands(r *CommandRegistry) {
	long := strings.Repeat("c", 1000)
	for i := 0; i < 5000; i++ {
		r.Command(fmt.Sprintf("%s%d", long, i), func(*Context) error { return nil })
	}
}

func TestPrepareRegisterTooLarge(t *testing.T) {
	core, conn := newFakeCore(t)
	errc := prepareAsync(t, conn, ack(1000), &manyCommands{})
	env, msg := core.recv()
	if msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_REGISTER || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("超限 REGISTER 应报告 SESSION_ERROR(REGISTER, RESOURCE_EXHAUSTED): %v", env)
	}
	if err := <-errc; !errors.Is(err, wire.ErrFrameTooLarge) {
		t.Fatalf("应返回帧超限错误，实际 %v", err)
	}
}

func TestPrepareRegisterRejectedByCore(t *testing.T) {
	core, conn := newFakeCore(t)
	errc := prepareAsync(t, conn, ack(1000), &lifecycleRoot{onStart: func(context.Context) error { return nil }})
	core.recv()
	core.send(pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: pb.SessionPhase_SESSION_PHASE_REGISTER},
		func(e *pb.Envelope) {
			e.Error = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, SafeMessage: "重复"}
		})
	var e *Error
	if err := <-errc; !errors.As(err, &e) || e.Code != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
		t.Fatalf("应返回 Core 的拒绝原因，实际 %v", err)
	}
}

// 声明了尚未支持的 HTTP/定时任务时，Run 在连接 Core 之前就明确失败，不静默丢弃声明。
func TestRunRejectsUnsupportedDeclarations(t *testing.T) {
	withSchedule := &root{schedules: func(s *ScheduleRegistry) { s.Every("t", time.Minute, func(*Context) error { return nil }) }}
	var e *Error
	if err := Run(withSchedule); !errors.As(err, &e) || e.Code != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("声明定时任务应返回 UNSUPPORTED，实际 %v", err)
	}
}

func TestReadLaunchValidation(t *testing.T) {
	t.Setenv(wire.EnvPluginID, "demo")
	t.Setenv(wire.EnvSecret, "abcd")
	t.Setenv(wire.EnvTransport, wire.TransportTCP)
	if _, err := readLaunch(); err == nil {
		t.Fatal("过短的 Secret 应被拒绝")
	}
	t.Setenv(wire.EnvSecret, strings.Repeat("ab", 32))
	t.Setenv(wire.EnvTransport, "udp")
	if _, err := readLaunch(); err == nil {
		t.Fatal("未知传输应被拒绝")
	}
	t.Setenv(wire.EnvTransport, wire.TransportTCP)
	if _, err := readLaunch(); err != nil {
		t.Fatal(err)
	}
	if _, err := dial(launch{transport: wire.TransportTCP, addr: "10.0.0.1:80"}); err == nil {
		t.Fatal("非 loopback 的 Core 地址应被拒绝")
	}
}

// Bootstrap 缺少阶段预算时报协议错误，而不是把预算当作 0 误报 OnStart 超时。
func TestPrepareRejectsMissingPhaseLimits(t *testing.T) {
	core, conn := newFakeCore(t)
	a := ack(1000)
	a.Bootstrap.PhaseLimits = nil
	called := false
	errc := prepareAsync(t, conn, a, &lifecycleRoot{onStart: func(context.Context) error { called = true; return nil }})
	env, msg := core.recv()
	if msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_CONFIG || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR {
		t.Fatalf("应报告 SESSION_ERROR(CONFIG, PROTOCOL_ERROR): %v", env)
	}
	if err := <-errc; err == nil || called {
		t.Fatalf("缺少预算时不应执行 OnStart：err=%v called=%v", err, called)
	}
}
