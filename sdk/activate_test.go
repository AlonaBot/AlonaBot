package sdk

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/alonabot/internal/wire"
	"github.com/AlonaBot/alonabot/sdk/plugin"
)

type readyRoot struct {
	plugin.Base
	onReady func(context.Context) error
}

func (r *readyRoot) OnReady(ctx context.Context) error { return r.onReady(ctx) }

func activateAsync(conn *wire.Conn, a *pb.HelloAck, root plugin.Root) <-chan error {
	errc := make(chan error, 1)
	go func() {
		s, err := activate(conn, testLaunch, a, root)
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
	// 就绪之后的请求通过门禁（B1 业务分发尚未实现，回 UNSUPPORTED 而不是 FAILED_PRECONDITION）。
	_ = core.conn.WriteFrame(coreRequest(2))
	resp, err = core.conn.ReadFrame()
	if err != nil || resp.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("就绪后的请求应通过门禁: %v %v", resp, err)
	}
}

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

func TestActivateWrongGenerationClosesSession(t *testing.T) {
	core, conn := newFakeCore(t)
	errc := activateAsync(conn, ack(1000), &readyRoot{onReady: func(context.Context) error { return nil }})
	core.send(pb.FrameKind_FRAME_KIND_ACTIVATE, &pb.Activate{GenerationId: "other", OnReadyDeadlineUnixMs: time.Now().Add(time.Second).UnixMilli()}, nil)
	if err := <-errc; err == nil {
		t.Fatal("代次不符的 ACTIVATE 应使激活失败")
	}
}
