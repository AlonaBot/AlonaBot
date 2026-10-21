package host

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// prepared 完成握手与准备，返回假插件、Core 侧连接与 Prepared。
func prepared(t *testing.T) (*fake, *wire.Conn, *Prepared) {
	t.Helper()
	a, b := net.Pipe()
	t.Cleanup(func() { _ = a.Close(); _ = b.Close() })
	coreConn := wire.NewStreamConn(a)
	f := &fake{t: t, conn: wire.NewStreamConn(b), raw: b}
	out := make(chan result, 1)
	go func() {
		p, err := Prepare(context.Background(), coreConn, expectA(), bootA(), DefaultLimits)
		out <- result{p, err}
	}()
	handshake(t, f)
	f.send(pb.FrameKind_FRAME_KIND_REGISTER, &pb.Register{Routes: []*pb.RouteDescriptor{cmd("hello")}, CandidateConfigJson: `{}`}, nil)
	if _, _, err := f.recv(); err != nil { // PREPARED
		t.Fatal(err)
	}
	r := <-out
	if r.err != nil {
		t.Fatal(r.err)
	}
	return f, coreConn, r.p
}

// store 准备一个已安装 plugin-a、含候选代次 gen-1 的状态库。
func store(t *testing.T) *state.Store {
	t.Helper()
	ctx := context.Background()
	s, err := state.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.AddInstall(ctx, state.Install{PluginID: "plugin-a", Source: "local", Entrypoint: "x", BuildDigest: digestA}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddGeneration(ctx, "plugin-a", "gen-1", "{}", 1); err != nil {
		t.Fatal(err)
	}
	return s
}

func commitGen1(s *state.Store) func(context.Context) error {
	return func(ctx context.Context) error { return s.CommitActive(ctx, "plugin-a", "", "gen-1") }
}

func activateAsync(conn *wire.Conn, p *Prepared, opt ActivateOptions) <-chan struct {
	in  *Instance
	err error
} {
	out := make(chan struct {
		in  *Instance
		err error
	}, 1)
	go func() {
		in, err := Activate(context.Background(), conn, p, opt)
		out <- struct {
			in  *Instance
			err error
		}{in, err}
	}()
	return out
}

func req(f *fake, id uint64) *pb.Envelope {
	ms := time.Now().Add(time.Minute).UnixMilli()
	return &pb.Envelope{ProtocolMajor: 1, SessionId: f.sid, Kind: pb.FrameKind_FRAME_KIND_REQUEST, SourcePluginId: "plugin-a",
		InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, CallId: id, DeadlineUnixMs: &ms}
}

func echoHandler(_ context.Context, in *wire.Inbound) wire.Response {
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: "t.v1", SchemaVersion: 1, Data: []byte(`"ok"`)}}
}

// 激活前的业务 REQUEST 被拒绝；ACTIVATED 与紧随其后的 REQUEST（同一次写出）按帧顺序处理。
func TestActivateGateAndOrdering(t *testing.T) {
	f, conn, p := prepared(t)
	s := store(t)
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: commitGen1(s), Handler: echoHandler, Limits: DefaultLimits})

	env, msg, err := f.recv()
	if err != nil || env.GetKind() != pb.FrameKind_FRAME_KIND_ACTIVATE || msg.(*pb.Activate).GetGenerationId() != "gen-1" {
		t.Fatalf("应收到 ACTIVATE(gen-1): %v %v", env, err)
	}
	if g, _ := s.Active(context.Background(), "plugin-a"); g != "gen-1" {
		t.Fatalf("ACTIVATE 之前必须已提交 ActiveGeneration，实际 %q", g)
	}

	// Activating 阶段的业务请求：拒绝。
	if err := f.conn.WriteFrame(req(f, 1)); err != nil {
		t.Fatal(err)
	}
	resp, err := f.conn.ReadFrame()
	if err != nil || resp.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
		t.Fatalf("激活前的 REQUEST 应被拒绝: %v %v", resp, err)
	}

	// ACTIVATED 与 REQUEST 拼在同一次写出中：Core 必须先应用 Activated 再处理 REQUEST。
	act, _ := wire.ControlEnvelope(f.sid, 0, "plugin-a", pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: "gen-1", Ready: true})
	b1, _ := wire.EncodeFrame(act)
	b2, _ := wire.EncodeFrame(req(f, 2))
	go func() { _, _ = f.raw.Write(append(b1, b2...)) }()
	resp, err = f.conn.ReadFrame()
	if err != nil || resp.GetError() != nil || resp.GetCallId() != 2 {
		t.Fatalf("紧随 ACTIVATED 的 REQUEST 应被接受: %v %v", resp, err)
	}
	r := <-out
	if r.err != nil || r.in.State() != StateActivated {
		t.Fatalf("应进入 Activated: %v %v", r.in.State(), r.err)
	}
}

func TestActivateOnReadyFailureIsDegradedAndKeepsNewGeneration(t *testing.T) {
	f, conn, p := prepared(t)
	s := store(t)
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: commitGen1(s), Limits: DefaultLimits})
	f.recv() // ACTIVATE
	f.send(pb.FrameKind_FRAME_KIND_SESSION_ERROR, &pb.SessionError{FailedPhase: pb.SessionPhase_SESSION_PHASE_ON_READY},
		func(e *pb.Envelope) {
			e.Error = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PLUGIN_NOT_READY, SafeMessage: "依赖未就绪"}
		})
	r := <-out
	var pe *PhaseError
	if r.in == nil || r.in.State() != StateDegraded || !errors.As(r.err, &pe) || pe.Phase != pb.SessionPhase_SESSION_PHASE_ON_READY {
		t.Fatalf("OnReady 失败应进入 Degraded: %v %v", r.in, r.err)
	}
	if g, _ := s.Active(context.Background(), "plugin-a"); g != "gen-1" {
		t.Fatalf("提交后失败不得回到旧代次，状态库中应仍为 gen-1，实际 %q", g)
	}
}

func TestActivateOnReadyTimeoutIsDegraded(t *testing.T) {
	f, conn, p := prepared(t)
	s := store(t)
	lim := DefaultLimits
	lim.OnReady = 150 * time.Millisecond
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: commitGen1(s), Limits: lim})
	f.recv() // ACTIVATE，之后插件一直不回 ACTIVATED
	go func() {
		for {
			if _, err := f.conn.ReadFrame(); err != nil {
				return
			}
		}
	}()
	r := <-out
	var pe *PhaseError
	if r.in.State() != StateDegraded || !errors.As(r.err, &pe) || pe.Code != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		t.Fatalf("OnReady 超时应进入 Degraded: %v %v", r.in.State(), r.err)
	}
	if g, _ := s.Active(context.Background(), "plugin-a"); g != "gen-1" {
		t.Fatalf("超时后状态库应仍为 gen-1，实际 %q", g)
	}
}

// 普通 restart：复用已提交的代次，不再提交、不产生新 Generation。
func TestRestartReusesCommittedGeneration(t *testing.T) {
	s := store(t)
	if err := commitGen1(s)(context.Background()); err != nil {
		t.Fatal(err)
	}
	f, conn, p := prepared(t)
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: nil, Limits: DefaultLimits})
	env, msg, _ := f.recv()
	if env.GetKind() != pb.FrameKind_FRAME_KIND_ACTIVATE || msg.(*pb.Activate).GetGenerationId() != "gen-1" {
		t.Fatalf("restart 应使用原代次 gen-1: %v", msg)
	}
	f.send(pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: "gen-1", Ready: true}, nil)
	if r := <-out; r.err != nil || r.in.State() != StateActivated {
		t.Fatalf("restart 激活失败: %v", r.err)
	}
	if g, _ := s.Active(context.Background(), "plugin-a"); g != "gen-1" {
		t.Fatalf("restart 不应改变 Active，实际 %q", g)
	}
	// 原代次不能再次提交，证明 restart 路径没有走提交。
	if err := commitGen1(s)(context.Background()); err == nil {
		t.Fatal("gen-1 已提交，不应能再次提交")
	}
}

// 提交失败（陈旧候选）：不发送 ACTIVATE，Active 不变。
func TestCommitFailureDoesNotActivate(t *testing.T) {
	s := store(t)
	ctx := context.Background()
	_ = s.AddGeneration(ctx, "plugin-a", "gen-0", "{}", 1)
	_ = s.CommitActive(ctx, "plugin-a", "", "gen-0") // 另一操作抢先提交
	f, conn, p := prepared(t)
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: commitGen1(s), Limits: DefaultLimits})
	env, _, err := f.recv()
	if err != nil || env.GetKind() != pb.FrameKind_FRAME_KIND_SESSION_ERROR {
		t.Fatalf("提交失败应收到 SESSION_ERROR 而不是 ACTIVATE: %v %v", env, err)
	}
	r := <-out
	if r.err == nil || !errors.Is(r.err, state.ErrStaleGeneration) || r.in != nil {
		t.Fatalf("应返回陈旧候选错误: %v", r.err)
	}
	if g, _ := s.Active(ctx, "plugin-a"); g != "gen-0" {
		t.Fatalf("Active 应保持 gen-0，实际 %q", g)
	}
}

// commit_marker 是由代次与准备结果计算的 32 字节摘要。
func TestCommitMarkerIsDigest(t *testing.T) {
	f, conn, p := prepared(t)
	s := store(t)
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: commitGen1(s), Limits: DefaultLimits})
	_, msg, _ := f.recv()
	m1 := msg.(*pb.Activate).GetCommitMarker()
	f.send(pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: "gen-1", Ready: true}, nil)
	<-out
	if len(m1) != 32 || bytes.Equal(m1, make([]byte, 32)) {
		t.Fatalf("commit_marker 应为 32 字节摘要: %x", m1)
	}
}
