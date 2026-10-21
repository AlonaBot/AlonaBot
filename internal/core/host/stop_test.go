package host

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

var fastStop = StopTiming{Drain: 200 * time.Millisecond, Total: 600 * time.Millisecond, OnStopAt: 300 * time.Millisecond}

// fakeTree 模拟进程树：exit() 表示进程自行退出，KillTree 立即令其退出。
type fakeTree struct {
	exited chan struct{}
	once   atomic.Bool
	killed atomic.Bool
}

func newTree() *fakeTree { return &fakeTree{exited: make(chan struct{})} }

func (f *fakeTree) Exited() <-chan struct{} { return f.exited }
func (f *fakeTree) exit() {
	if f.once.CompareAndSwap(false, true) {
		close(f.exited)
	}
}
func (f *fakeTree) KillTree() error { f.killed.Store(true); f.exit(); return nil }

// activated 返回已进入 Activated 的实例与假插件；frames 持续接收假插件读到的帧。
func activated(t *testing.T, handler wire.Handler) (*fake, *Instance, <-chan *pb.Envelope) {
	t.Helper()
	f, conn, p := prepared(t)
	s := store(t)
	out := activateAsync(conn, p, ActivateOptions{PluginID: "plugin-a", GenerationID: "gen-1", Commit: commitGen1(s), Handler: handler, Limits: DefaultLimits})
	f.recv() // ACTIVATE
	f.send(pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: "gen-1", Ready: true}, nil)
	r := <-out
	if r.err != nil {
		t.Fatal(r.err)
	}
	frames := make(chan *pb.Envelope, 64)
	go func() {
		for {
			env, err := f.conn.ReadFrame()
			if err != nil {
				close(frames)
				return
			}
			frames <- env
		}
	}()
	return f, r.in, frames
}

func next(t *testing.T, frames <-chan *pb.Envelope, kind pb.FrameKind) *pb.Envelope {
	t.Helper()
	timeout := time.After(3 * time.Second)
	for {
		select {
		case env, ok := <-frames:
			if !ok {
				t.Fatalf("等待 %s 时连接关闭", kind)
			}
			if env.GetKind() == kind {
				return env
			}
		case <-timeout:
			t.Fatalf("等待 %s 超时", kind)
		}
	}
}

func childReq(f *fake, id, parent uint64) *pb.Envelope {
	e := req(f, id)
	if parent != 0 {
		e.ParentCallId = proto.Uint64(parent)
	}
	return e
}

func stopAsync(in *Instance, tree Tree, timing StopTiming) (time.Time, <-chan StopResult) {
	t0 := time.Now()
	out := make(chan StopResult, 1)
	go func() { out <- in.Stop(t0, pb.StopReason_STOP_REASON_DISABLE, tree, timing) }()
	return t0, out
}

func waitState(t *testing.T, in *Instance, s State) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for in.State() != s {
		if time.Now().After(deadline) {
			t.Fatalf("状态应为 %s，实际 %s", s, in.State())
		}
		time.Sleep(time.Millisecond)
	}
}

// 排空期间：新根调用被拒绝，已接纳调用的子调用仍可完成；排空完成后发送 STOPPING。
func TestStopDrainAllowsChildCallsAndRejectsNewRoots(t *testing.T) {
	f, in, frames := activated(t, echoHandler)
	tree := newTree()

	callDone := make(chan error, 1)
	go func() {
		_, err := in.Session.Call(context.Background(), wire.Request{Target: "plugin-a/x", Payload: wire.Payload{
			Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: "t.v1", SchemaVersion: 1, Data: []byte(`1`)}})
		callDone <- err
	}()
	parent := next(t, frames, pb.FrameKind_FRAME_KIND_REQUEST) // Core→插件 的在途请求

	t0, out := stopAsync(in, tree, fastStop)
	waitState(t, in, StateDraining)

	_ = f.conn.WriteFrame(childReq(f, 10, 0)) // 新根调用
	if r := next(t, frames, pb.FrameKind_FRAME_KIND_RESPONSE); r.GetCallId() != 10 || r.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("排空期间的新根调用应被拒绝: %v", r)
	}
	_ = f.conn.WriteFrame(childReq(f, 11, parent.GetCallId())) // 父链内的子调用
	if r := next(t, frames, pb.FrameKind_FRAME_KIND_RESPONSE); r.GetCallId() != 11 || r.GetError() != nil {
		t.Fatalf("父链内的子调用应能完成: %v", r)
	}
	// 插件回应父请求 → 排空完成。
	_ = f.conn.WriteFrame(&pb.Envelope{ProtocolMajor: 1, SessionId: f.sid, Kind: pb.FrameKind_FRAME_KIND_RESPONSE,
		InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_CORE, CallId: parent.GetCallId(),
		PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaId: "t.v1", SchemaVersion: 1, Payload: []byte(`2`)})
	if err := <-callDone; err != nil {
		t.Fatalf("排空期间在途调用应正常完成: %v", err)
	}

	st := next(t, frames, pb.FrameKind_FRAME_KIND_STOPPING)
	msg, _ := validate.DecodeControl(st)
	tree.exit()
	r := <-out
	if !r.Drained || r.Killed {
		t.Fatalf("应在排空期内完成且无需强杀: %+v", r)
	}
	want := r.StoppingSentAt.Add(fastStop.OnStopAt)
	if total := t0.Add(fastStop.Total); total.Before(want) {
		want = total
	}
	if got := time.UnixMilli(msg.(*pb.Stopping).GetDeadlineUnixMs()); got.Sub(want).Abs() > time.Millisecond {
		t.Fatalf("STOPPING 截止 = %v，期望 min(t0+Total, 发送时刻+OnStop) = %v", got, want)
	}
	if in.State() != StateStopped {
		t.Fatalf("停止后状态应为 Stopped，实际 %s", in.State())
	}
}

// 排空超时：取消剩余调用（对端收到 Cancel），随后发送 STOPPING；STOPPING 截止受 t0+Total 约束。
func TestStopDrainTimeoutCancelsRemaining(t *testing.T) {
	_, in, frames := activated(t, echoHandler)
	tree := newTree()
	callDone := make(chan error, 1)
	go func() {
		_, err := in.Session.Call(context.Background(), wire.Request{Target: "plugin-a/x"})
		callDone <- err
	}()
	parent := next(t, frames, pb.FrameKind_FRAME_KIND_REQUEST) // 插件永不响应
	timing := StopTiming{Drain: 200 * time.Millisecond, Total: 300 * time.Millisecond, OnStopAt: 300 * time.Millisecond}
	t0, out := stopAsync(in, tree, timing)
	if c := next(t, frames, pb.FrameKind_FRAME_KIND_CANCEL); c.GetCallId() != parent.GetCallId() {
		t.Fatalf("应取消在途请求 %d: %v", parent.GetCallId(), c)
	}
	if err := <-callDone; wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_CANCELLED {
		t.Fatalf("被取消的调用应得到 CANCELLED: %v", err)
	}
	st := next(t, frames, pb.FrameKind_FRAME_KIND_STOPPING)
	msg, _ := validate.DecodeControl(st)
	got := time.UnixMilli(msg.(*pb.Stopping).GetDeadlineUnixMs())
	if want := t0.Add(timing.Total); got.Sub(want).Abs() > time.Millisecond {
		t.Fatalf("发送时刻+OnStop 晚于 t0+Total 时，截止应为 t0+Total=%v，实际 %v", want, got)
	}
	tree.exit()
	if r := <-out; r.Drained {
		t.Fatal("排空未完成却报告 Drained")
	}
}

// OnStop 挂起、进程不退出：在 t0+Total 终止进程树。
func TestStopKillsTreeAtTotalDeadline(t *testing.T) {
	_, in, frames := activated(t, echoHandler)
	tree := newTree()
	t0, out := stopAsync(in, tree, fastStop)
	next(t, frames, pb.FrameKind_FRAME_KIND_STOPPING) // 插件收到但不退出
	r := <-out
	elapsed := time.Since(t0)
	if !r.Killed || !tree.killed.Load() {
		t.Fatal("到总截止仍未退出时应终止进程树")
	}
	if elapsed < fastStop.Total || elapsed > fastStop.Total+300*time.Millisecond {
		t.Fatalf("终止时刻 %v 应在总截止 %v 附近", elapsed, fastStop.Total)
	}
}

// Core Shutdown：多个插件并行停止，共享同一个总截止，总耗时不超过一份预算。
func TestStopAllSharesOneBudget(t *testing.T) {
	targets := map[*Instance]Tree{}
	for i := 0; i < 3; i++ {
		_, in, frames := activated(t, echoHandler)
		go func() {
			for range frames {
			}
		}()
		targets[in] = newTree() // 都不会自行退出
	}
	t0 := time.Now()
	res := StopAll(t0, pb.StopReason_STOP_REASON_CORE_SHUTDOWN, targets, fastStop)
	elapsed := time.Since(t0)
	if len(res) != 3 {
		t.Fatalf("结果数 %d", len(res))
	}
	for _, r := range res {
		if !r.Killed {
			t.Fatal("每个插件都应在共享截止时被终止")
		}
	}
	if elapsed > fastStop.Total+300*time.Millisecond {
		t.Fatalf("并行停止耗时 %v，超过一份预算 %v（不应串行累加）", elapsed, fastStop.Total)
	}
}

func TestRestartPolicy(t *testing.T) {
	p := NewRestartPolicy()
	base := time.Unix(1_000_000, 0)
	if d := p.OnExit("a", false, base); d != DecisionNone {
		t.Fatalf("正常退出不应重启，实际 %v", d)
	}
	if p.OnExit("a", true, base) != DecisionRestart || p.OnExit("a", true, base.Add(time.Minute)) != DecisionRestart {
		t.Fatal("前两次异常退出应重启")
	}
	if d := p.OnExit("a", true, base.Add(4*time.Minute)); d != DecisionDisable {
		t.Fatalf("5 分钟内第三次异常退出应自动禁用，实际 %v", d)
	}
	// 窗口滑动：间隔超过 5 分钟的失败不计入。
	q := NewRestartPolicy()
	q.OnExit("b", true, base)
	q.OnExit("b", true, base.Add(3*time.Minute))
	if d := q.OnExit("b", true, base.Add(6*time.Minute)); d != DecisionRestart {
		t.Fatalf("第一次失败已移出窗口，应重启，实际 %v", d)
	}
	// 手动启用重置窗口；不同插件互不影响。
	p.Reset("a")
	if d := p.OnExit("a", true, base.Add(5*time.Minute)); d != DecisionRestart {
		t.Fatalf("Reset 后应重新计数，实际 %v", d)
	}
	if d := p.OnExit("c", true, base); d != DecisionRestart {
		t.Fatalf("其他插件不受影响，实际 %v", d)
	}
}

func TestStopDefaultsMatchContract(t *testing.T) {
	if DefaultStopTiming != (StopTiming{6 * time.Second, 10 * time.Second, 4 * time.Second}) {
		t.Fatalf("DefaultStopTiming = %+v，应为排空 6s、总计 10s、OnStop 4s", DefaultStopTiming)
	}
	if p := NewRestartPolicy(); p.Window != 5*time.Minute || p.Limit != 3 {
		t.Fatalf("重启策略应为 5 分钟内 3 次: %+v", p)
	}
}
