package wire

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/alonabot/contracts/validate"
	"google.golang.org/protobuf/proto"
)

// 首发固定常量（CONTRACTS §1.2）。Options 中的同名字段只供测试缩小。
const (
	DefaultDeadline       = 30 * time.Second
	MaxInflightPerDir     = 256
	MaxBytesPerDir        = 16 << 20
	GlobalPayloadBytes    = 64 << 20
	defaultPriorityFrames = 1024
)

// Budget 是跨 Session 共享的 payload 字节预算（Core 合计 64 MiB）。
type Budget struct {
	mu          sync.Mutex
	used, limit int64
}

// NewBudget 创建上限为 limit 字节的预算。
func NewBudget(limit int64) *Budget { return &Budget{limit: limit} }

func (b *Budget) acquire(n int64) bool {
	if b == nil {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.used+n > b.limit {
		return false
	}
	b.used += n
	return true
}

func (b *Budget) release(n int64) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.used -= n
	b.mu.Unlock()
}

// Used 返回当前占用的字节数。
func (b *Budget) Used() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.used
}

// Payload 是一次调用或响应的业务载荷。
type Payload struct {
	Encoding      pb.PayloadEncoding
	SchemaID      string
	SchemaVersion uint32
	Data          []byte
}

// Request 是本端发起的一次调用。
type Request struct {
	Target  string
	Payload Payload
	TraceID []byte
	// Deadline 为零时使用 DefaultDeadline；最终期限还会受 ctx 与父调用约束。
	Deadline time.Time
	// ParentCallID 只能由插件侧设置，引用同 Session 中仍在途的 Core→插件 请求。
	ParentCallID uint64
	// Attach 只保存在本端；对端以本调用为父发起子请求时，交给子请求的 Handler（Core 借此派生可信上下文）。
	Attach any
}

// Response 是一次调用的结果；Payload 与 Err 可以同时存在（例如部分发送成功）。
type Response struct {
	Payload Payload
	Err     *pb.ErrorInfo
}

// Inbound 是对端发来的一次调用。
type Inbound struct {
	CallID       uint64
	Target       string
	Payload      Payload
	TraceID      []byte
	Deadline     time.Time
	ParentCallID uint64
	// ParentAttach 是父调用（本端发起）的 Request.Attach；无父调用时为 nil。
	ParentAttach any
}

// Handler 处理对端调用；ctx 在本端期限到达、对端取消、父调用结束或 Session 关闭时取消。
// 返回值即响应：没有业务结果时必须返回 Err。
type Handler func(ctx context.Context, in *Inbound) Response

// CallError 是携带 ErrorInfo 的调用错误。
type CallError struct{ Info *pb.ErrorInfo }

func (e *CallError) Error() string {
	return fmt.Sprintf("%s: %s", e.Info.GetCode(), e.Info.GetSafeMessage())
}

// Code 返回错误码。
func (e *CallError) Code() pb.ErrorCode { return e.Info.GetCode() }

func callErr(code pb.ErrorCode, msg string) *CallError {
	return &CallError{Info: &pb.ErrorInfo{Code: code, SafeMessage: msg}}
}

// CodeOf 返回 err 中的协议错误码；不是 *CallError 时返回 UNSPECIFIED。
func CodeOf(err error) pb.ErrorCode {
	var ce *CallError
	if errors.As(err, &ce) {
		return ce.Code()
	}
	return pb.ErrorCode_ERROR_CODE_UNSPECIFIED
}

// Options 配置一个 Session。
type Options struct {
	Local          pb.InitiatorSide
	SessionID      []byte
	ProtocolMinor  uint32
	SourcePluginID string // Local 为 PLUGIN 时写入发出的帧
	Handler        Handler
	// Control 在读循环中同步处理非调用帧（控制帧与 EVENT），保证状态变化按帧顺序生效；返回错误关闭 Session。
	Control func(*pb.Envelope) error
	Global  *Budget

	// 以下为零时使用首发常量，只供测试缩小。
	MaxInflight     int
	MaxBytes        int64
	PriorityFrames  int
	DefaultDeadline time.Duration
}

// Stats 是可观测的计数。
type Stats struct {
	LateResponses  int64 // 迟到、重复或未知的 Response
	UnknownCancels int64
}

type callState int

const (
	queued callState = iota
	writing
	written
	withdrawn
)

type outCall struct {
	id       uint64
	env      *pb.Envelope
	bytes    int64
	deadline time.Time
	attach   any
	result   chan Response // 容量 1，只写一次

	ctx    context.Context // 调用存续期间有效，结束时取消（子调用以此派生）
	cancel context.CancelFunc

	mu           sync.Mutex
	state        callState
	cancelLater  bool // 写出过程中被取消：写完后再发 Cancel
	doneComplete bool
}

type inCall struct {
	bytes    int64
	deadline time.Time
	cancel   context.CancelFunc
}

// Session 是一条已认证连接上的双向调用引擎。
type Session struct {
	conn *Conn
	opt  Options
	peer pb.InitiatorSide

	prio   chan *pb.Envelope
	bizSig chan struct{}
	done   chan struct{}

	mu       sync.Mutex
	nextID   uint64
	lastIn   uint64
	biz      []*outCall
	pending  map[uint64]*outCall // 本端发起、尚未完成
	inbound  map[uint64]*inCall  // 对端发起、尚未响应
	outBytes int64
	inBytes  int64
	err      error

	late, unknownCancel atomic.Int64
	closeOnce           sync.Once
}

// NewSession 在已认证的连接上启动读循环与写循环。
func NewSession(conn *Conn, opt Options) *Session {
	if opt.MaxInflight == 0 {
		opt.MaxInflight = MaxInflightPerDir
	}
	if opt.MaxBytes == 0 {
		opt.MaxBytes = MaxBytesPerDir
	}
	if opt.PriorityFrames == 0 {
		opt.PriorityFrames = defaultPriorityFrames
	}
	if opt.DefaultDeadline == 0 {
		opt.DefaultDeadline = DefaultDeadline
	}
	peer := pb.InitiatorSide_INITIATOR_SIDE_CORE
	if opt.Local == pb.InitiatorSide_INITIATOR_SIDE_CORE {
		peer = pb.InitiatorSide_INITIATOR_SIDE_PLUGIN
	}
	s := &Session{
		conn: conn, opt: opt, peer: peer,
		prio:    make(chan *pb.Envelope, opt.PriorityFrames),
		bizSig:  make(chan struct{}, 1),
		done:    make(chan struct{}),
		pending: map[uint64]*outCall{},
		inbound: map[uint64]*inCall{},
	}
	go s.readLoop()
	go s.writeLoop()
	return s
}

// Done 在 Session 关闭后关闭。
func (s *Session) Done() <-chan struct{} { return s.done }

// Err 返回 Session 关闭的原因（未关闭时为 nil）。
func (s *Session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Close 主动关闭 Session；未完成的调用以 UNAVAILABLE 结束。
func (s *Session) Close() error {
	s.fail(errors.New("Session 已关闭"))
	return nil
}

// Stats 返回计数快照。
func (s *Session) Stats() Stats {
	return Stats{LateResponses: s.late.Load(), UnknownCancels: s.unknownCancel.Load()}
}

func (s *Session) envelope(kind pb.FrameKind) *pb.Envelope {
	env := &pb.Envelope{
		ProtocolMajor: validate.ProtocolMajor, ProtocolMinor: s.opt.ProtocolMinor,
		SessionId: s.opt.SessionID, Kind: kind,
	}
	if s.opt.Local == pb.InitiatorSide_INITIATOR_SIDE_PLUGIN {
		env.SourcePluginId = s.opt.SourcePluginID
	}
	return env
}

func setPayload(env *pb.Envelope, p Payload) {
	env.PayloadEncoding, env.SchemaId, env.SchemaVersion, env.Payload = p.Encoding, p.SchemaID, p.SchemaVersion, p.Data
}

// ControlEnvelope 构造控制帧（以及 Core 固定载荷）的 Envelope：PROTOBUF 编码、确定性序列化、
// schema_id 为 alonabot.<snake>.v1。sourcePluginID 只在插件发出的帧上设置。Core、SDK 与 Session 共用。
func ControlEnvelope(sid []byte, minor uint32, sourcePluginID string, kind pb.FrameKind, msg proto.Message) (*pb.Envelope, error) {
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &pb.Envelope{ProtocolMajor: validate.ProtocolMajor, ProtocolMinor: minor, SessionId: sid, Kind: kind,
		SourcePluginId: sourcePluginID, PayloadEncoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
		SchemaId: validate.SchemaID(msg), SchemaVersion: validate.SchemaVersion, Payload: data}, nil
}

func (s *Session) controlEnvelope(kind pb.FrameKind, msg proto.Message) (*pb.Envelope, error) {
	src := ""
	if s.opt.Local == pb.InitiatorSide_INITIATOR_SIDE_PLUGIN {
		src = s.opt.SourcePluginID
	}
	return ControlEnvelope(s.opt.SessionID, s.opt.ProtocolMinor, src, kind, msg)
}

// SendControl 经优先通道发出控制帧。
func (s *Session) SendControl(kind pb.FrameKind, msg proto.Message) error {
	env, err := s.controlEnvelope(kind, msg)
	if err != nil {
		return err
	}
	return s.enqueuePrio(env)
}

// enqueuePrio 把帧放入有界优先通道；放不进说明连控制通道都无法推进，关闭 Session（CONTRACTS §1.2）。
func (s *Session) enqueuePrio(env *pb.Envelope) error {
	select {
	case <-s.done:
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "Session 已关闭")
	default:
	}
	select {
	case s.prio <- env:
		return nil
	default:
		err := errors.New("优先通道已满，控制/响应无法推进")
		s.fail(err)
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, err.Error())
	}
}

// Call 发起一次调用并等待结果。准入不足时立即返回 RESOURCE_EXHAUSTED，不等待槽位。
// 本地期限或取消先结束调用并返回，再尽力通知对端；已发出的副作用不会撤销，也不自动重试。
func (s *Session) Call(ctx context.Context, req Request) (Response, error) {
	n := int64(len(req.Payload.Data))
	deadline := req.Deadline
	if deadline.IsZero() {
		deadline = time.Now().Add(s.opt.DefaultDeadline)
	}
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	s.mu.Lock()
	if s.err != nil {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "Session 已关闭")
	}
	if len(s.pending) >= s.opt.MaxInflight || s.outBytes+n > s.opt.MaxBytes {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "本方向在途请求数或字节预算已满")
	}
	var parent *inCall
	if req.ParentCallID != 0 {
		parent = s.inbound[req.ParentCallID]
		if s.opt.Local != pb.InitiatorSide_INITIATOR_SIDE_PLUGIN || parent == nil {
			s.mu.Unlock()
			return Response{}, callErr(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "父调用不存在或已结束")
		}
		if parent.deadline.Before(deadline) {
			deadline = parent.deadline // 子请求不得延长父期限
		}
	}
	if !deadline.After(time.Now()) {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "期限已过，请求未发送")
	}
	if !s.opt.Global.acquire(n) {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "Core 全局 payload 预算已满")
	}
	s.nextID++
	id := s.nextID
	env := s.envelope(pb.FrameKind_FRAME_KIND_REQUEST)
	env.InitiatorSide, env.CallId, env.Target, env.TraceId = s.opt.Local, id, req.Target, req.TraceID
	env.DeadlineUnixMs = proto.Int64(deadline.UnixMilli())
	if req.ParentCallID != 0 {
		env.ParentCallId = proto.Uint64(req.ParentCallID)
	}
	setPayload(env, req.Payload)
	cctx, cancel := context.WithCancel(context.Background())
	c := &outCall{id: id, env: env, bytes: n, deadline: deadline, attach: req.Attach, result: make(chan Response, 1), ctx: cctx, cancel: cancel}
	s.pending[id] = c
	s.outBytes += n
	s.biz = append(s.biz, c)
	s.mu.Unlock()
	s.signalBiz()

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case r := <-c.result:
		return r, errorOf(r)
	case <-timer.C:
		return s.abandon(c, pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, pb.CancelReason_CANCEL_REASON_DEADLINE_EXCEEDED, "调用超过期限")
	case <-ctx.Done():
		code := pb.ErrorCode_ERROR_CODE_CANCELLED
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
		}
		return s.abandon(c, code, pb.CancelReason_CANCEL_REASON_CALLER_CANCELLED, "调用方已取消")
	}
}

func errorOf(r Response) error {
	if r.Err != nil {
		return &CallError{Info: r.Err}
	}
	return nil
}

// abandon 在本地先结束调用，再按写出状态撤回请求或随后发送 Cancel。
func (s *Session) abandon(c *outCall, code pb.ErrorCode, reason pb.CancelReason, msg string) (Response, error) {
	r := Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: msg}}
	if !s.complete(c, r) {
		// 结果已经先一步到达。
		r = <-c.result
		return r, errorOf(r)
	}
	<-c.result
	c.mu.Lock()
	sendCancel := false
	switch c.state {
	case queued:
		c.state = withdrawn // 写循环会跳过它，请求从未上线
	case writing:
		c.cancelLater = true
	case written:
		sendCancel = true
	}
	c.mu.Unlock()
	if sendCancel {
		s.sendCancel(c.id, reason)
	}
	return r, &CallError{Info: r.Err}
}

func (s *Session) sendCancel(id uint64, reason pb.CancelReason) {
	env, err := s.controlEnvelope(pb.FrameKind_FRAME_KIND_CANCEL, &pb.Cancel{Reason: reason})
	if err != nil {
		return
	}
	env.InitiatorSide, env.CallId = s.opt.Local, id
	_ = s.enqueuePrio(env)
}

// complete 只完成一次调用：移出在途表、释放预算、取消其 ctx（连带取消以它为父的子调用）。
func (s *Session) complete(c *outCall, r Response) bool {
	s.mu.Lock()
	if _, ok := s.pending[c.id]; !ok {
		s.mu.Unlock()
		return false
	}
	delete(s.pending, c.id)
	s.outBytes -= c.bytes
	s.mu.Unlock()
	s.opt.Global.release(c.bytes)
	c.cancel()
	c.result <- r
	return true
}

func (s *Session) signalBiz() {
	select {
	case s.bizSig <- struct{}{}:
	default:
	}
}

func (s *Session) popBiz() *outCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.biz) == 0 {
		return nil
	}
	c := s.biz[0]
	s.biz[0] = nil
	s.biz = s.biz[1:]
	return c
}

// writeLoop 是唯一的写者：优先通道（Response/Cancel/控制）总是先于业务 Request。
func (s *Session) writeLoop() {
	for {
		select {
		case env := <-s.prio:
			s.write(env)
			continue
		case <-s.done:
			return
		default:
		}
		if c := s.popBiz(); c != nil {
			s.writeRequest(c)
			continue
		}
		select {
		case env := <-s.prio:
			s.write(env)
		case <-s.bizSig:
		case <-s.done:
			return
		}
	}
}

func (s *Session) write(env *pb.Envelope) {
	if err := s.conn.WriteFrame(env); err != nil {
		s.fail(fmt.Errorf("写帧失败: %w", err))
	}
}

func (s *Session) writeRequest(c *outCall) {
	c.mu.Lock()
	if c.state == withdrawn {
		c.mu.Unlock()
		return
	}
	c.state = writing
	c.mu.Unlock()
	s.write(c.env)
	c.mu.Lock()
	c.state = written
	later := c.cancelLater
	c.mu.Unlock()
	if later {
		s.sendCancel(c.id, pb.CancelReason_CANCEL_REASON_CALLER_CANCELLED)
	}
}

// readLoop 是唯一的读者：调用帧与 Handler 执行分离，控制帧同步交给 Options.Control。
func (s *Session) readLoop() {
	for {
		env, err := s.conn.ReadFrame()
		if err != nil {
			s.fail(fmt.Errorf("读帧失败: %w", err))
			return
		}
		if err := s.dispatch(env); err != nil {
			s.fail(err)
			return
		}
	}
}

func (s *Session) dispatch(env *pb.Envelope) error {
	if err := validate.Envelope(env); err != nil {
		return err
	}
	if string(env.GetSessionId()) != string(s.opt.SessionID) {
		return errors.New("帧的 session_id 与本 Session 不符")
	}
	switch env.GetKind() {
	case pb.FrameKind_FRAME_KIND_REQUEST:
		return s.onRequest(env)
	case pb.FrameKind_FRAME_KIND_RESPONSE:
		return s.onResponse(env)
	case pb.FrameKind_FRAME_KIND_CANCEL:
		return s.onCancel(env)
	}
	if s.opt.Control == nil {
		return fmt.Errorf("未处理的帧种类 %s", env.GetKind())
	}
	return s.opt.Control(env)
}

func payloadOf(env *pb.Envelope) Payload {
	return Payload{env.GetPayloadEncoding(), env.GetSchemaId(), env.GetSchemaVersion(), env.GetPayload()}
}

func (s *Session) onRequest(env *pb.Envelope) error {
	if env.GetInitiatorSide() != s.peer {
		return errors.New("REQUEST 的 initiator_side 不是对端")
	}
	id := env.GetCallId()
	n := int64(len(env.GetPayload()))
	deadline := time.UnixMilli(env.GetDeadlineUnixMs())
	in := &Inbound{CallID: id, Target: env.GetTarget(), Payload: payloadOf(env), TraceID: env.GetTraceId(), Deadline: deadline}

	s.mu.Lock()
	if id <= s.lastIn {
		s.mu.Unlock()
		return fmt.Errorf("对端 call_id %d 未单调递增（上一个为 %d）", id, s.lastIn)
	}
	s.lastIn = id
	parentCtx := context.Background()
	if pid := env.GetParentCallId(); env.ParentCallId != nil {
		parent := s.pending[pid]
		if parent == nil {
			s.mu.Unlock()
			return s.respondErr(id, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "父调用不存在或已结束")
		}
		in.ParentCallID, in.ParentAttach, parentCtx = pid, parent.attach, parent.ctx
		if parent.deadline.Before(deadline) {
			deadline = parent.deadline // 子请求不得延长父期限
			in.Deadline = deadline
		}
	}
	if len(s.inbound) >= s.opt.MaxInflight || s.inBytes+n > s.opt.MaxBytes || !s.opt.Global.acquire(n) {
		s.mu.Unlock()
		return s.respondErr(id, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "本方向在途请求数或字节预算已满")
	}
	ctx, cancel := context.WithDeadline(parentCtx, deadline)
	s.inbound[id] = &inCall{bytes: n, deadline: deadline, cancel: cancel}
	s.inBytes += n
	s.mu.Unlock()

	go s.serve(ctx, cancel, in, n)
	return nil
}

func (s *Session) serve(ctx context.Context, cancel context.CancelFunc, in *Inbound, n int64) {
	defer cancel()
	r := s.handle(ctx, in)
	if r.Payload.SchemaID == "" && r.Err == nil {
		r.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "处理函数既没有结果也没有错误"}
	}
	s.mu.Lock()
	delete(s.inbound, in.CallID)
	s.inBytes -= n
	s.mu.Unlock()
	s.opt.Global.release(n)
	env := s.envelope(pb.FrameKind_FRAME_KIND_RESPONSE)
	env.InitiatorSide, env.CallId, env.Error = s.peer, in.CallID, r.Err
	setPayload(env, r.Payload)
	_ = s.enqueuePrio(env)
}

// handle 调用 Handler；panic 转为 INTERNAL，不当作空成功（CONTRACTS §2.1）。
func (s *Session) handle(ctx context.Context, in *Inbound) (r Response) {
	if s.opt.Handler == nil {
		return Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "本端不接受调用"}}
	}
	defer func() {
		if p := recover(); p != nil {
			r = Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "处理函数 panic"}}
		}
	}()
	return s.opt.Handler(ctx, in)
}

func (s *Session) respondErr(id uint64, code pb.ErrorCode, msg string) error {
	env := s.envelope(pb.FrameKind_FRAME_KIND_RESPONSE)
	env.InitiatorSide, env.CallId = s.peer, id
	env.Error = &pb.ErrorInfo{Code: code, SafeMessage: msg}
	return s.enqueuePrio(env)
}

func (s *Session) onResponse(env *pb.Envelope) error {
	if env.GetInitiatorSide() != s.opt.Local {
		return errors.New("RESPONSE 的 initiator_side 不是本端")
	}
	s.mu.Lock()
	c := s.pending[env.GetCallId()]
	s.mu.Unlock()
	if c == nil || !s.complete(c, Response{Payload: payloadOf(env), Err: env.GetError()}) {
		s.late.Add(1) // 迟到、重复或未知：不第二次完成调用
	}
	return nil
}

func (s *Session) onCancel(env *pb.Envelope) error {
	if env.GetInitiatorSide() != s.peer {
		return errors.New("CANCEL 的 initiator_side 不是对端")
	}
	s.mu.Lock()
	in := s.inbound[env.GetCallId()]
	s.mu.Unlock()
	if in == nil {
		s.unknownCancel.Add(1)
		return nil
	}
	in.cancel()
	return nil
}

// fail 关闭 Session：关闭连接，未完成的本端调用以 UNAVAILABLE 结束，取消对端调用的处理。
func (s *Session) fail(err error) {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.err = err
		pending := make([]*outCall, 0, len(s.pending))
		for _, c := range s.pending {
			pending = append(pending, c)
		}
		inbound := make([]*inCall, 0, len(s.inbound))
		for _, in := range s.inbound {
			inbound = append(inbound, in)
		}
		s.mu.Unlock()
		close(s.done)
		_ = s.conn.Close()
		for _, c := range pending {
			s.complete(c, Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "Session 已断开"}})
		}
		for _, in := range inbound {
			in.cancel()
		}
	})
}
