// 本文件测试插件侧激活与停止：ready gate、OnReady 失败与超时、代次不符、STOPPING 与断链后的 OnStop。
package sdk

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// readyRoot 是 OnReady 行为可注入的根插件。
type readyRoot struct {
	// Base 提供根插件的默认实现。
	plugin.Base
	// onReady 是 OnReady 的实际行为。
	onReady func(context.Context) error
}

// OnReady 调用注入的 onReady。
func (r *readyRoot) OnReady(ctx context.Context) error { return r.onReady(ctx) }

// activateAsync 在后台执行 activate，返回其错误；成功时等到 Session 关闭再结束。
func activateAsync(conn *wire.Conn, a *pb.HelloAck, root plugin.Root) <-chan error {
	errc := make(chan error, 1)
	go func() {
		s, _, err := activate(conn, testLaunch, a, root)
		if s != nil {
			defer s.Close()
		}
		errc <- err
		if s != nil {
			<-s.Done()
		}
	}()
	return errc
}

// coreRequest 构造 Core 发出的合法业务 REQUEST。
func coreRequest(id uint64) *pb.Envelope {
	ms := time.Now().Add(time.Minute).UnixMilli()
	return &pb.Envelope{ProtocolMajor: 1, SessionId: testSID, Kind: pb.FrameKind_FRAME_KIND_REQUEST,
		InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_CORE, CallId: id, DeadlineUnixMs: &ms}
}

// ready gate 打开之前拒绝业务请求；ACTIVATED 被 Core 读走（完整写出）之后 activate 才返回。
func TestActivateGateOpensAfterActivatedWritten(t *testing.T) {
	core, conn := newFakeCore(t)
	ran := make(chan struct{})
	errc := activateAsync(conn, ack(1000), &readyRoot{onReady: func(context.Context) error { close(ran); return nil }})

	// ACTIVATE 之前的业务请求：拒绝。
	_ = core.conn.WriteFrame(coreRequest(1))
	resp, err := core.conn.ReadFrame()
	if err != nil || resp.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
		t.Fatalf("就绪前的请求应被拒绝: %v %v", resp, err)
	}

	core.send(pb.FrameKind_FRAME_KIND_ACTIVATE, &pb.Activate{GenerationId: "g1", OnReadyDeadlineUnixMs: time.Now().Add(time.Second).UnixMilli()}, nil)
	<-ran
	// net.Pipe 是同步的：Core 读走 ACTIVATED 之前，写出未完成，ready gate 不应打开。
	select {
	case err := <-errc:
		t.Fatalf("ACTIVATED 尚未写出时 activate 不应返回: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	env, msg := core.recv()
	if env.GetKind() != pb.FrameKind_FRAME_KIND_ACTIVATED || msg.(*pb.Activated).GetGenerationId() != "g1" || !msg.(*pb.Activated).GetReady() {
		t.Fatalf("应收到 ACTIVATED(g1): %v", msg)
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	// 就绪之后的请求通过门禁（业务分发尚未实现，回 UNSUPPORTED 而不是 FAILED_PRECONDITION）。
	_ = core.conn.WriteFrame(coreRequest(2))
	resp, err = core.conn.ReadFrame()
	if err != nil || resp.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("就绪后的请求应通过门禁: %v %v", resp, err)
	}
}

// OnReady 返回错误时报告 SESSION_ERROR(ON_READY, PLUGIN_NOT_READY)，安全消息不含用户错误原文。
func TestActivateOnReadyErrorReported(t *testing.T) {
	core, conn := newFakeCore(t)
	errc := activateAsync(conn, ack(1000), &readyRoot{onReady: func(context.Context) error { return errors.New("token=SECRET-R") }})
	core.send(pb.FrameKind_FRAME_KIND_ACTIVATE, &pb.Activate{GenerationId: "g1", OnReadyDeadlineUnixMs: time.Now().Add(time.Second).UnixMilli()}, nil)
	env, msg := core.recv()
	if msg.(*pb.SessionError).GetFailedPhase() != pb.SessionPhase_SESSION_PHASE_ON_READY || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_PLUGIN_NOT_READY {
		t.Fatalf("应报告 SESSION_ERROR(ON_READY, PLUGIN_NOT_READY): %v", env)
	}
	if env.GetError().GetSafeMessage() == "" || strings.Contains(env.GetError().GetSafeMessage(), "SECRET-R") {
		t.Fatalf("安全消息不应含用户错误原文: %q", env.GetError().GetSafeMessage())
	}
	if err := <-errc; err == nil {
		t.Fatal("OnReady 失败时 activate 应返回错误")
	}
}

// OnReady 超过 ACTIVATE 给出的截止时间时报告 DEADLINE_EXCEEDED。
func TestActivateOnReadyHangsBeyondDeadline(t *testing.T) {
	core, conn := newFakeCore(t)
	block := make(chan struct{})
	defer close(block)
	errc := activateAsync(conn, ack(1000), &readyRoot{onReady: func(context.Context) error { <-block; return nil }})
	core.send(pb.FrameKind_FRAME_KIND_ACTIVATE, &pb.Activate{GenerationId: "g1", OnReadyDeadlineUnixMs: time.Now().Add(100 * time.Millisecond).UnixMilli()}, nil)
	env, _ := core.recv()
	if env.GetKind() != pb.FrameKind_FRAME_KIND_SESSION_ERROR || env.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		t.Fatalf("OnReady 挂起应报告 DEADLINE_EXCEEDED: %v", env)
	}
	if err := <-errc; err == nil {
		t.Fatal("应返回错误")
	}
}

// ACTIVATE 的代次与 Bootstrap 不符时激活失败。
func TestActivateWrongGenerationClosesSession(t *testing.T) {
	core, conn := newFakeCore(t)
	errc := activateAsync(conn, ack(1000), &readyRoot{onReady: func(context.Context) error { return nil }})
	core.send(pb.FrameKind_FRAME_KIND_ACTIVATE, &pb.Activate{GenerationId: "other", OnReadyDeadlineUnixMs: time.Now().Add(time.Second).UnixMilli()}, nil)
	if err := <-errc; err == nil {
		t.Fatal("代次不符的 ACTIVATE 应使激活失败")
	}
}

// stopRoot 是 OnStop 行为可注入的根插件。
type stopRoot struct {
	// Base 提供根插件的默认实现。
	plugin.Base
	// onStop 是 OnStop 的实际行为。
	onStop func(context.Context) error
}

// OnStop 调用注入的 onStop。
func (r *stopRoot) OnStop(ctx context.Context) error { return r.onStop(ctx) }

// activatedSDK 让 SDK 完成激活，返回假 Core、Session 与 activation。
func activatedSDK(t *testing.T, root plugin.Root) (*fakeCore, *wire.Session, *activation) {
	t.Helper()
	core, conn := newFakeCore(t)
	// res 是一次 activate 的返回值。
	type res struct {
		// s 是激活后的 Session。
		s *wire.Session
		// a 是激活状态。
		a *activation
		// err 是 activate 的错误。
		err error
	}
	out := make(chan res, 1)
	go func() { s, a, err := activate(conn, testLaunch, ack(1000), root); out <- res{s, a, err} }()
	core.send(pb.FrameKind_FRAME_KIND_ACTIVATE, &pb.Activate{GenerationId: "g1", OnReadyDeadlineUnixMs: time.Now().Add(time.Second).UnixMilli()}, nil)
	core.recv() // ACTIVATED
	r := <-out
	if r.err != nil {
		t.Fatal(r.err)
	}
	return core, r.s, r.a
}

// STOPPING：关闭 ready gate、取消 lifetime，用独立 context 按 STOPPING 截止执行 OnStop，然后正常返回。
func TestStoppingRunsOnStopWithIndependentContext(t *testing.T) {
	// seen 是 OnStop 观察到的 context 状态。
	type seen struct {
		// alreadyDone 为 true 表示传入的 context 一开始就已取消。
		alreadyDone bool
		// deadline 是传入 context 的截止时间。
		deadline time.Time
	}
	got := make(chan seen, 1)
	release := make(chan struct{})
	core, sess, a := activatedSDK(t, &stopRoot{onStop: func(ctx context.Context) error {
		d, _ := ctx.Deadline()
		got <- seen{ctx.Err() != nil, d}
		<-release
		return nil
	}})
	served := make(chan error, 1)
	go func() { served <- serve(sess, a) }()

	stopAt := time.Now().Add(2 * time.Second)
	core.send(pb.FrameKind_FRAME_KIND_STOPPING, &pb.Stopping{DeadlineUnixMs: stopAt.UnixMilli(), Reason: pb.StopReason_STOP_REASON_DISABLE}, nil)
	s := <-got
	if s.alreadyDone {
		t.Fatal("OnStop 不得复用已取消的 lifetime context")
	}
	if s.deadline.Sub(stopAt).Abs() > time.Millisecond {
		t.Fatalf("OnStop 截止应为 STOPPING 截止 %v，实际 %v", stopAt, s.deadline)
	}
	if a.lifetime.Err() == nil {
		t.Fatal("STOPPING 后 lifetime 应已取消")
	}
	// STOPPING 之后到达的业务请求被拒绝。
	_ = core.conn.WriteFrame(coreRequest(9))
	resp, err := core.conn.ReadFrame()
	if err != nil || resp.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
		t.Fatalf("停止中的请求应被拒绝: %v %v", resp, err)
	}
	close(release)
	if err := <-served; err != nil {
		t.Fatalf("收到 STOPPING 后应正常返回: %v", err)
	}
}

// OnStop 不响应 ctx：watchdog 在 STOPPING 截止放弃等待，serve 照常返回。
func TestStoppingOnStopHangs(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	core, sess, a := activatedSDK(t, &stopRoot{onStop: func(context.Context) error { <-block; return nil }})
	served := make(chan error, 1)
	go func() { served <- serve(sess, a) }()
	start := time.Now()
	core.send(pb.FrameKind_FRAME_KIND_STOPPING, &pb.Stopping{DeadlineUnixMs: time.Now().Add(150 * time.Millisecond).UnixMilli()}, nil)
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OnStop 挂起时 serve 应在截止后返回")
	}
	if el := time.Since(start); el > time.Second {
		t.Fatalf("返回耗时 %v，远超 150ms 截止", el)
	}
}

// 断链（未收到 STOPPING）：立即停止新工作，执行 OnStop 后返回错误，不重连。
func TestEOFRunsOnStopAndReturnsError(t *testing.T) {
	stopped := make(chan struct{})
	core, sess, a := activatedSDK(t, &stopRoot{onStop: func(context.Context) error { close(stopped); return nil }})
	served := make(chan error, 1)
	go func() { served <- serve(sess, a) }()
	_ = core.conn.Close()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("断链后应执行 OnStop")
	}
	if err := <-served; err == nil {
		t.Fatal("断链应返回错误")
	}
	if a.lifetime.Err() == nil {
		t.Fatal("断链后 lifetime 应已取消")
	}
}

// 重复的 STOPPING 是协议错误，关闭 Session。
func TestDuplicateStoppingIsProtocolError(t *testing.T) {
	core, sess, a := activatedSDK(t, &stopRoot{onStop: func(context.Context) error { time.Sleep(200 * time.Millisecond); return nil }})
	served := make(chan error, 1)
	go func() { served <- serve(sess, a) }()
	st := &pb.Stopping{DeadlineUnixMs: time.Now().Add(time.Second).UnixMilli()}
	core.send(pb.FrameKind_FRAME_KIND_STOPPING, st, nil)
	core.send(pb.FrameKind_FRAME_KIND_STOPPING, st, nil)
	<-served
	if sess.Err() == nil {
		t.Fatal("重复 STOPPING 应关闭 Session")
	}
}

// SDK 的 OnStop 与握手预算和 CONTRACTS §4.2 一致。
func TestOnStopBudgetMatchesContract(t *testing.T) {
	if onStopBudget != 4*time.Second || handshakeTimeout != 10*time.Second {
		t.Fatalf("SDK 预算应为 OnStop 4s、握手 10s：%v %v", onStopBudget, handshakeTimeout)
	}
}
