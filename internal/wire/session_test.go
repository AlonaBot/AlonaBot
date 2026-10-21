package wire

import (
	"bytes"
	"context"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

var sid16 = bytes.Repeat([]byte{7}, 16)

// gatedConn 可以让写出阻塞，用来观察「尚未写出」的请求。
type gatedConn struct {
	net.Conn
	gated atomic.Bool
	gate  chan struct{}
}

func (g *gatedConn) Write(b []byte) (int, error) {
	if g.gated.Load() {
		<-g.gate
	}
	return g.Conn.Write(b)
}

func jsonPayload(s string) Payload {
	return Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: "test.echo.v1", SchemaVersion: 1, Data: []byte(`"` + s + `"`)}
}

func echo(prefix string) Handler {
	return func(_ context.Context, in *Inbound) Response {
		return Response{Payload: jsonPayload(prefix + strings.Trim(string(in.Payload.Data), `"`))}
	}
}

type pair struct {
	core, plugin         *Session
	coreConn, pluginConn *gatedConn
}

func newPair(t *testing.T, coreOpt, pluginOpt Options) *pair {
	t.Helper()
	a, b := net.Pipe()
	ca := &gatedConn{Conn: a, gate: make(chan struct{})}
	cb := &gatedConn{Conn: b, gate: make(chan struct{})}
	coreOpt.Local, coreOpt.SessionID = pb.InitiatorSide_INITIATOR_SIDE_CORE, sid16
	pluginOpt.Local, pluginOpt.SessionID = pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, sid16
	pluginOpt.SourcePluginID = "demo"
	p := &pair{core: NewSession(NewStreamConn(ca), coreOpt), plugin: NewSession(NewStreamConn(cb), pluginOpt), coreConn: ca, pluginConn: cb}
	t.Cleanup(func() { _ = p.core.Close(); _ = p.plugin.Close() })
	return p
}

func call(t *testing.T, s *Session, data string) Response {
	t.Helper()
	r, err := s.Call(context.Background(), Request{Target: "demo/echo", Payload: jsonPayload(data)})
	if err != nil {
		t.Fatalf("Call(%s): %v", data, err)
	}
	return r
}

func TestProductionConstants(t *testing.T) {
	if DefaultDeadline != 30*time.Second || MaxInflightPerDir != 256 || MaxBytesPerDir != 16<<20 || GlobalPayloadBytes != 64<<20 {
		t.Fatal("首发常量与 CONTRACTS §1.2 不一致")
	}
}

// 两个方向同时使用 call_id=1，结果互不串号。
func TestSameCallIDBothDirections(t *testing.T) {
	p := newPair(t, Options{Handler: echo("core:")}, Options{Handler: echo("plugin:")})
	var wg sync.WaitGroup
	var fromPlugin, fromCore Response
	wg.Add(2)
	go func() { defer wg.Done(); fromPlugin = call(t, p.core, "a") }()
	go func() { defer wg.Done(); fromCore = call(t, p.plugin, "b") }()
	wg.Wait()
	if string(fromPlugin.Payload.Data) != `"plugin:a"` || string(fromCore.Payload.Data) != `"core:b"` {
		t.Fatalf("串号：%s / %s", fromPlugin.Payload.Data, fromCore.Payload.Data)
	}
}

// 后发先回：每个调用者都拿到自己的结果。
func TestOutOfOrderResponses(t *testing.T) {
	release := map[string]chan struct{}{"1": make(chan struct{}), "2": make(chan struct{}), "3": make(chan struct{})}
	p := newPair(t, Options{}, Options{Handler: func(_ context.Context, in *Inbound) Response {
		<-release[strings.Trim(string(in.Payload.Data), `"`)]
		return Response{Payload: in.Payload}
	}})
	results := make(chan string, 3)
	for _, k := range []string{"1", "2", "3"} {
		go func(k string) { results <- string(call(t, p.core, k).Payload.Data) }(k)
	}
	for _, k := range []string{"3", "2", "1"} {
		close(release[k])
		if got := <-results; got != `"`+k+`"` {
			t.Fatalf("期望先拿到 %s 的结果，实际 %s", k, got)
		}
	}
}

func TestPayloadAndErrorBothKept(t *testing.T) {
	p := newPair(t, Options{}, Options{Handler: func(context.Context, *Inbound) Response {
		return Response{Payload: jsonPayload("partial"), Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, SafeMessage: "第二部分超时"}}
	}})
	r, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("x")})
	if CodeOf(err) != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED || string(r.Payload.Data) != `"partial"` {
		t.Fatalf("payload 与 error 应同时保留：payload=%s err=%v", r.Payload.Data, err)
	}
}

func TestHandlerPanicBecomesInternal(t *testing.T) {
	p := newPair(t, Options{}, Options{Handler: func(context.Context, *Inbound) Response { panic("boom") }})
	if _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("x")}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_INTERNAL {
		t.Fatalf("panic 应转为 INTERNAL，实际 %v", err)
	}
}

// 256 个在途后立即拒绝；业务槽满时 Response 与 Cancel 仍能推进。
func TestAdmissionFullButResponsesAndCancelsProgress(t *testing.T) {
	block := make(chan struct{})
	var canceledSeen atomic.Int32
	// 对端入站上限放大到 1000：第 257 个请求只能被本端出站准入拒绝，而不是被对端拒绝。
	p := newPair(t, Options{}, Options{MaxInflight: 1000, Handler: func(ctx context.Context, in *Inbound) Response {
		if string(in.Payload.Data) == `"cancel-me"` {
			<-ctx.Done()
			canceledSeen.Add(1)
			return Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_CANCELLED}}
		}
		<-block
		return Response{Payload: in.Payload}
	}})
	cancelCtx, cancel := context.WithCancel(context.Background())
	canceled := make(chan error, 1)
	go func() {
		_, err := p.core.Call(cancelCtx, Request{Payload: jsonPayload("cancel-me")})
		canceled <- err
	}()
	// 先确认它已写出（call_id 1）：尚未写出的请求被取消时会撤回、不发 Cancel，就测不到 Cancel 的推进。
	waitFor(t, func() bool { return pendingState(p.core, 1) == written })
	done := make(chan error, MaxInflightPerDir)
	for i := 1; i < MaxInflightPerDir; i++ {
		go func() { _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("w")}); done <- err }()
	}
	waitFor(t, func() bool {
		p.core.mu.Lock()
		defer p.core.mu.Unlock()
		return len(p.core.pending) == MaxInflightPerDir
	})

	start := time.Now()
	if _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("over")}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("第 257 个请求应立即被拒绝，实际 %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("满额拒绝不应等待槽位")
	}
	// 满额时 Cancel 仍可送达对端。
	cancel()
	if err := <-canceled; CodeOf(err) != pb.ErrorCode_ERROR_CODE_CANCELLED {
		t.Fatalf("取消应立即返回 CANCELLED，实际 %v", err)
	}
	waitFor(t, func() bool { return canceledSeen.Load() == 1 })
	// 满额时 Response 仍可返回。
	close(block)
	for i := 1; i < MaxInflightPerDir; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	call(t, p.core, "again") // 槽位释放后可再次准入
}

func TestInboundAdmission(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	p := newPair(t, Options{}, Options{MaxInflight: 2, Handler: func(_ context.Context, in *Inbound) Response { <-block; return Response{Payload: in.Payload} }})
	for i := 0; i < 2; i++ {
		go func() { _, _ = p.core.Call(context.Background(), Request{Payload: jsonPayload("w")}) }()
	}
	waitFor(t, func() bool { p.plugin.mu.Lock(); defer p.plugin.mu.Unlock(); return len(p.plugin.inbound) == 2 })
	if _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("over")}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("对端入站满额时应收到 RESOURCE_EXHAUSTED，实际 %v", err)
	}
}

func TestByteBudgets(t *testing.T) {
	big := Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: "test.bytes.v1", SchemaVersion: 1, Data: make([]byte, 600)}
	block := make(chan struct{})
	h := func(_ context.Context, in *Inbound) Response { <-block; return Response{Payload: jsonPayload("ok")} }

	// 每方向字节预算。
	p := newPair(t, Options{MaxBytes: 1000}, Options{Handler: h})
	go func() { _, _ = p.core.Call(context.Background(), Request{Payload: big}) }()
	waitFor(t, func() bool { p.core.mu.Lock(); defer p.core.mu.Unlock(); return p.core.outBytes == 600 })
	if _, err := p.core.Call(context.Background(), Request{Payload: big}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("超过每方向字节预算应被拒绝，实际 %v", err)
	}

	// 跨 Session 共享的全局预算。
	global := NewBudget(1000)
	p1 := newPair(t, Options{Global: global}, Options{Handler: h})
	p2 := newPair(t, Options{Global: global}, Options{Handler: h})
	go func() { _, _ = p1.core.Call(context.Background(), Request{Payload: big}) }()
	waitFor(t, func() bool { return global.Used() == 600 })
	if _, err := p2.core.Call(context.Background(), Request{Payload: big}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
		t.Fatalf("另一 Session 应受全局预算限制，实际 %v", err)
	}
	close(block)
	waitFor(t, func() bool { return global.Used() == 0 })
}

// 尚未写出的请求被取消：从队列撤出，对端既收不到它也收不到它的 Cancel。
func TestCancelBeforeSendIsWithdrawn(t *testing.T) {
	var mu sync.Mutex
	var seen []uint64
	p := newPair(t, Options{}, Options{Handler: func(_ context.Context, in *Inbound) Response {
		mu.Lock()
		seen = append(seen, in.CallID)
		mu.Unlock()
		return Response{Payload: in.Payload}
	}})
	p.coreConn.gated.Store(true) // Core 的写出卡住

	first := make(chan error, 1)
	go func() { _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("A")}); first <- err }()
	waitFor(t, func() bool { return pendingState(p.core, 1) == writing })

	ctx, cancel := context.WithCancel(context.Background())
	second := make(chan error, 1)
	go func() { _, err := p.core.Call(ctx, Request{Payload: jsonPayload("B")}); second <- err }()
	waitFor(t, func() bool { return pendingState(p.core, 2) == queued })
	cancel()
	if err := <-second; CodeOf(err) != pb.ErrorCode_ERROR_CODE_CANCELLED {
		t.Fatalf("发送前取消应返回 CANCELLED，实际 %v", err)
	}

	p.coreConn.gated.Store(false)
	close(p.coreConn.gate)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
	call(t, p.core, "C") // call_id 3
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 3 {
		t.Fatalf("对端收到的 call_id = %v，被取消的 2 不应上线", seen)
	}
	if got := p.plugin.Stats().UnknownCancels; got != 0 {
		t.Fatalf("未上线请求不应发送 Cancel，对端收到 %d 个未知 Cancel", got)
	}
}

func pendingState(s *Session, id uint64) callState {
	s.mu.Lock()
	c := s.pending[id]
	s.mu.Unlock()
	if c == nil {
		return -1
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// 本地超时：调用者先得到 DEADLINE_EXCEEDED，对端收到 Cancel；迟到的 Response 只计数。
func TestLateResponseCountedNotCompleted(t *testing.T) {
	sawCancel := make(chan struct{})
	finish := make(chan struct{})
	p := newPair(t, Options{}, Options{Handler: func(ctx context.Context, in *Inbound) Response {
		<-ctx.Done()
		close(sawCancel)
		<-finish // 不合作：取消后仍继续运行
		return Response{Payload: in.Payload}
	}})
	_, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("slow"), Deadline: time.Now().Add(50 * time.Millisecond)})
	if CodeOf(err) != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		t.Fatalf("应返回 DEADLINE_EXCEEDED，实际 %v", err)
	}
	<-sawCancel
	close(finish)
	waitFor(t, func() bool { return p.core.Stats().LateResponses == 1 })
}

func TestParentChildLinkage(t *testing.T) {
	parentDeadline := time.Now().Add(2 * time.Second)
	type seen struct {
		attach   any
		deadline time.Time
	}
	childSeen := make(chan seen, 1)
	childCanceled := make(chan struct{})
	var p *pair
	p = newPair(t,
		Options{Handler: func(ctx context.Context, in *Inbound) Response {
			childSeen <- seen{in.ParentAttach, in.Deadline}
			if strings.Contains(string(in.Payload.Data), "hold") {
				<-ctx.Done()
				close(childCanceled)
			}
			return Response{Payload: jsonPayload("child-ok")}
		}},
		Options{Handler: func(ctx context.Context, in *Inbound) Response {
			r, err := p.plugin.Call(ctx, Request{Payload: in.Payload, ParentCallID: in.CallID, Deadline: time.Now().Add(time.Hour)})
			if err != nil {
				return Response{Err: &pb.ErrorInfo{Code: CodeOf(err), SafeMessage: err.Error()}}
			}
			return r
		}})

	r, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("x"), Deadline: parentDeadline, Attach: "actor-42"})
	if err != nil || string(r.Payload.Data) != `"child-ok"` {
		t.Fatalf("A→B→A 子调用失败: %s %v", r.Payload.Data, err)
	}
	got := <-childSeen
	if got.attach != "actor-42" {
		t.Fatalf("子调用应继承父调用的 Core 记录，实际 %v", got.attach)
	}
	if got.deadline.After(parentDeadline) {
		t.Fatalf("子调用期限 %v 晚于父调用 %v", got.deadline, parentDeadline)
	}

	// 父调用结束（这里是本地超时）后，正在执行的子调用被取消。
	_, err = p.core.Call(context.Background(), Request{Payload: jsonPayload("hold"), Deadline: time.Now().Add(100 * time.Millisecond)})
	if CodeOf(err) != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		t.Fatalf("父调用应超时，实际 %v", err)
	}
	<-childSeen
	select {
	case <-childCanceled:
	case <-time.After(2 * time.Second):
		t.Fatal("父调用结束后子调用没有被取消")
	}

	// 插件侧：父调用不在途时本地直接拒绝。
	if _, err := p.plugin.Call(context.Background(), Request{Payload: jsonPayload("x"), ParentCallID: 999}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
		t.Fatalf("引用不存在的父调用应失败，实际 %v", err)
	}
}

// 期限已过的调用不入队、不上线。
func TestExpiredDeadlineNotSent(t *testing.T) {
	var handled atomic.Int32
	p := newPair(t, Options{}, Options{Handler: func(_ context.Context, in *Inbound) Response { handled.Add(1); return Response{Payload: in.Payload} }})
	if _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("x"), Deadline: time.Now().Add(-time.Second)}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		t.Fatalf("期限已过应立即返回 DEADLINE_EXCEEDED，实际 %v", err)
	}
	call(t, p.core, "y")
	if handled.Load() != 1 {
		t.Fatalf("期限已过的请求不应上线，对端执行了 %d 次", handled.Load())
	}
}

// Core 侧：插件引用的父调用在 Core 已结束时，返回 FAILED_PRECONDITION 而不执行。
func TestCoreRejectsChildOfFinishedParent(t *testing.T) {
	coreHandled := atomic.Int32{}
	childErr := make(chan error, 1)
	parentDone := make(chan struct{})
	var p *pair
	p = newPair(t,
		Options{Handler: func(context.Context, *Inbound) Response {
			coreHandled.Add(1)
			return Response{Payload: jsonPayload("no")}
		}},
		Options{Handler: func(_ context.Context, in *Inbound) Response {
			<-parentDone // Core 侧父调用已超时结束，但插件侧仍在执行
			_, err := p.plugin.Call(context.Background(), Request{Payload: jsonPayload("late-child"), ParentCallID: in.CallID})
			childErr <- err
			return Response{Payload: jsonPayload("done")}
		}})
	// 父调用因 Core 侧取消而结束（期限未到），插件侧的父调用记录仍在途。
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(50 * time.Millisecond); cancel() }()
	_, _ = p.core.Call(ctx, Request{Payload: jsonPayload("x"), Deadline: time.Now().Add(time.Minute)})
	close(parentDone)
	if err := <-childErr; CodeOf(err) != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
		t.Fatalf("父调用已结束的子请求应被 Core 拒绝，实际 %v", err)
	}
	if coreHandled.Load() != 0 {
		t.Fatal("被拒绝的子请求不应执行 Core Handler")
	}
}

func TestSessionCloseFailsPending(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	p := newPair(t, Options{}, Options{Handler: func(context.Context, *Inbound) Response { <-block; return Response{} }})
	errc := make(chan error, 1)
	go func() { _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("x")}); errc <- err }()
	waitFor(t, func() bool { p.plugin.mu.Lock(); defer p.plugin.mu.Unlock(); return len(p.plugin.inbound) == 1 })
	_ = p.plugin.Close()
	if err := <-errc; CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("断链后在途调用应为 UNAVAILABLE，实际 %v", err)
	}
	if _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("y")}); CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("关闭后的 Session 应拒绝新调用，实际 %v", err)
	}
}

func TestNonMonotonicCallIDClosesSession(t *testing.T) {
	a, b := net.Pipe()
	core := NewSession(NewStreamConn(a), Options{Local: pb.InitiatorSide_INITIATOR_SIDE_CORE, SessionID: sid16,
		Handler: func(_ context.Context, in *Inbound) Response { return Response{Payload: in.Payload} }})
	defer core.Close()
	raw := NewStreamConn(b)
	go func() { // 吸收 Core 的响应
		for {
			if _, err := raw.ReadFrame(); err != nil {
				return
			}
		}
	}()
	req := func(id uint64) *pb.Envelope {
		ms := time.Now().Add(time.Minute).UnixMilli()
		return &pb.Envelope{ProtocolMajor: 1, SessionId: sid16, Kind: pb.FrameKind_FRAME_KIND_REQUEST,
			InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, CallId: id, DeadlineUnixMs: &ms, SourcePluginId: "demo"}
	}
	_ = raw.WriteFrame(req(5))
	_ = raw.WriteFrame(req(5))
	select {
	case <-core.Done():
		if !strings.Contains(core.Err().Error(), "单调递增") {
			t.Fatalf("关闭原因 %v", core.Err())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("重复 call_id 应关闭 Session")
	}
}

func TestWrongSessionIDClosesSession(t *testing.T) {
	a, b := net.Pipe()
	core := NewSession(NewStreamConn(a), Options{Local: pb.InitiatorSide_INITIATOR_SIDE_CORE, SessionID: sid16})
	defer core.Close()
	other := bytes.Repeat([]byte{9}, 16)
	_ = NewStreamConn(b).WriteFrame(&pb.Envelope{ProtocolMajor: 1, SessionId: other, Kind: pb.FrameKind_FRAME_KIND_RESPONSE,
		InitiatorSide: pb.InitiatorSide_INITIATOR_SIDE_CORE, CallId: 1, Error: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL}})
	select {
	case <-core.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("其他 Session 的帧应关闭当前 Session")
	}
}

// 写出卡住、优先通道也写满时关闭 Session，未完成调用以 UNAVAILABLE 结束，而不是死锁。
func TestPriorityFullClosesSession(t *testing.T) {
	p := newPair(t, Options{PriorityFrames: 1}, Options{})
	p.coreConn.gated.Store(true)
	errc := make(chan error, 1)
	go func() { _, err := p.core.Call(context.Background(), Request{Payload: jsonPayload("x")}); errc <- err }()
	waitFor(t, func() bool { return pendingState(p.core, 1) == writing })
	hb := &pb.Heartbeat{Seq: 1}
	var err error
	for i := 0; i < 3 && err == nil; i++ {
		err = p.core.SendControl(pb.FrameKind_FRAME_KIND_HEARTBEAT, hb)
	}
	if CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE || !strings.Contains(p.core.Err().Error(), "优先通道已满") {
		t.Fatalf("优先通道写满应关闭 Session：err=%v sessionErr=%v", err, p.core.Err())
	}
	if err := <-errc; CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("在途调用应以 UNAVAILABLE 结束，实际 %v", err)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("等待条件超时")
		}
		time.Sleep(time.Millisecond)
	}
}
