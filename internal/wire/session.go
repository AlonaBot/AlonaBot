// 本文件负责一条已认证连接上的双向调用引擎：调用表、准入预算、优先通道、取消因果与读写生命周期（WIRE §4）。
package wire

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"google.golang.org/protobuf/proto"
)

// 首发固定常量（CONTRACTS §1.2）。Options 中的同名字段只供测试缩小。
const (
	// DefaultDeadline 是请求未指定期限时的默认期限。
	DefaultDeadline = 30 * time.Second
	// MaxInflightPerDir 是每 Session 每方向最多的在途业务请求数。
	MaxInflightPerDir = 256
	// MaxBytesPerDir 是每 Session 每方向在途与排队载荷的字节上限。
	MaxBytesPerDir = 16 << 20
	// GlobalPayloadBytes 是 Core 所有 Session 合计的载荷字节上限。
	GlobalPayloadBytes = 64 << 20
	// defaultPriorityFrames 是优先通道（Response/Cancel/控制帧）的容量。
	defaultPriorityFrames = 1024
)

// Payload 是一次调用或响应的业务载荷。
type Payload struct {
	// Encoding 是 Data 的编码。
	Encoding pb.PayloadEncoding
	// SchemaID 是载荷契约 ID；为空表示没有载荷。
	SchemaID string
	// SchemaVersion 是载荷契约版本。
	SchemaVersion uint32
	// Data 是编码后的载荷，其长度计入准入预算。
	Data []byte
}

// Request 是本端发起的一次调用。
type Request struct {
	// Target 是被调用的 route/service 或 Core 服务名。
	Target string
	// Payload 是请求载荷。
	Payload Payload
	// TraceID 是 16 字节业务追踪 ID；没有业务上下文时为空。
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
	// Payload 是对端返回的业务结果；出错时也可能携带有效的部分结果。
	Payload Payload
	// Err 是对调用者公开的错误；对端发来的未知错误码已映射为 INTERNAL。
	Err *pb.ErrorInfo
	// RawErrorCode 是对端发来的原始错误码，只供内部诊断；与 Err.Code 不同说明发生了映射。
	RawErrorCode pb.ErrorCode
}

// Inbound 是对端发来的一次调用。
type Inbound struct {
	// CallID 是对端分配的调用编号。
	CallID uint64
	// Target 是被调用的 route/service 名。
	Target string
	// Payload 是请求载荷。
	Payload Payload
	// TraceID 是对端帧上的追踪 ID。
	TraceID []byte
	// Deadline 是有效期限，已按父调用期限收紧。
	Deadline time.Time
	// ParentCallID 是父调用编号；Admit 看到的是帧上声明的值，Handler 看到的是已校验在途的值，0 表示没有。
	ParentCallID uint64
	// ParentAttach 是父调用（本端发起）的 Request.Attach；无父调用时为 nil。
	ParentAttach any
}

// Handler 处理对端调用；ctx 在本端期限到达、对端取消、父调用结束或 Session 关闭时取消。
// 返回值即响应：没有业务结果时必须返回 Err。
type Handler func(ctx context.Context, in *Inbound) Response

// CallError 是携带 ErrorInfo 的调用错误。
type CallError struct {
	// Info 是调用者可见的错误。
	Info *pb.ErrorInfo
}

// Error 返回「错误码: 安全消息」。
func (e *CallError) Error() string {
	return fmt.Sprintf("%s: %s", e.Info.GetCode(), e.Info.GetSafeMessage())
}

// Code 返回错误码。
func (e *CallError) Code() pb.ErrorCode { return e.Info.GetCode() }

// callErr 构造本地产生的调用错误。
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
	// Local 是本端身份（CORE 或 PLUGIN）。
	Local pb.InitiatorSide
	// SessionID 是握手分配的 16 字节 Session ID。
	SessionID []byte
	// ProtocolMinor 是协商出的协议次版本，写入发出的帧并核对收到的帧。
	ProtocolMinor uint32
	// SourcePluginID 是插件的认证 ID：Local 为 PLUGIN 时写入发出的帧，Local 为 CORE 时用来核对对端帧。
	SourcePluginID string
	// Handler 处理对端的业务 REQUEST；为 nil 时回复 UNSUPPORTED。
	Handler Handler
	// Control 在读循环中同步处理非调用帧（控制帧与 EVENT），保证状态变化按帧顺序生效；返回错误关闭 Session。
	Control func(*pb.Envelope) error
	// Admit 在读循环中同步决定是否接受对端的业务 REQUEST（阶段门禁）；返回非 nil 时直接回该错误，不执行 Handler。
	// 必须在读循环中同步判断：状态变化（如 ACTIVATED）也在读循环中同步生效，帧顺序因此得到保证。
	Admit func(*Inbound) *pb.ErrorInfo
	// Global 是跨 Session 共享的全局预算；为 nil 时不限额（插件侧）。
	Global *Budget

	// 以下为零时使用首发常量，只供测试缩小。

	// MaxInflight 覆盖 MaxInflightPerDir。
	MaxInflight int
	// MaxBytes 覆盖 MaxBytesPerDir。
	MaxBytes int64
	// PriorityFrames 覆盖优先通道容量。
	PriorityFrames int
	// DefaultDeadline 覆盖默认请求期限。
	DefaultDeadline time.Duration
}

// Stats 是可观测的计数。
type Stats struct {
	// LateResponses 是迟到、重复或未知的 Response 数。
	LateResponses int64
	// UnknownCancels 是找不到对应在途调用的 Cancel 数。
	UnknownCancels int64
}

// callState 是本端请求在写出方向上的状态，只前进不回退。
type callState int

// 请求的写出状态。
const (
	queued    callState = iota // 在业务队列中等待 writer
	writing                    // writer 正在写出
	written                    // 已完整写出
	withdrawn                  // 写出前被结束，已移出队列，永不上线
)

// outCall 是本端发起、尚未完成的一次调用。
type outCall struct {
	// id 是本端分配的调用编号。
	id uint64
	// env 是待写出的 REQUEST；移出队列或写出后置 nil，不再持有载荷。
	env *pb.Envelope
	// deadline 是有效期限，子调用不能超过它。
	deadline time.Time
	// attach 是 Request.Attach，交给以本调用为父的子请求的 Handler。
	attach any
	// result 容量 1，只写一次。
	result chan Response

	// ctx 在调用存续期间有效，结束时取消（子调用以此派生）。
	ctx context.Context
	// cancel 取消 ctx。
	cancel context.CancelFunc

	// 以下字段由 Session.mu 保护。

	// bytes 是仍计入本方向与全局预算的载荷字节；归还后为 0，保证只归还一次。
	bytes int64
	// state 是写出状态。
	state callState
	// finished 为 true 表示调用方已得到结果（已移出 pending）。
	finished bool
	// cancelReason 是写出过程中被本端放弃时，写完后发送 Cancel 的原因；UNSPECIFIED 表示不发。
	cancelReason pb.CancelReason
}

// prioFrame 是优先通道中的一帧；written 非 nil 时在写出完成后收到写结果。
type prioFrame struct {
	// env 是要写出的帧。
	env *pb.Envelope
	// written 为 nil 时不通知写结果；否则容量 1，收到一次写结果。
	written chan error
}

// inCall 是对端发起、尚未响应的一次调用。
type inCall struct {
	// bytes 是请求载荷字节数，响应时归还。
	bytes int64
	// deadline 是有效期限，用于约束本端发起的子请求。
	deadline time.Time
	// cancel 取消 Handler 的 ctx（对端 Cancel 或 Session 关闭时）。
	cancel context.CancelFunc
}

// Session 是一条已认证连接上的双向调用引擎。
type Session struct {
	// conn 是底层帧连接。
	conn *Conn
	// opt 是已补齐默认值的配置，创建后不变。
	opt Options
	// peer 是对端身份，由 opt.Local 推出。
	peer pb.InitiatorSide

	// prio 是优先通道：Response、Cancel 与控制帧，总是先于业务 Request 写出。
	prio chan prioFrame
	// bizSig 容量 1，有新业务 Request 入队时唤醒 writer。
	bizSig chan struct{}
	// done 在 Session 关闭时关闭。
	done chan struct{}

	// mu 保护以下字段以及各 outCall 中标明由它保护的字段。
	mu sync.Mutex
	// nextID 是本端最近分配的调用编号。
	nextID uint64
	// lastIn 是对端最近的调用编号，用于检查单调递增。
	lastIn uint64
	// biz 是等待写出的业务 Request，按入队顺序排列，只含 queued 状态的调用。
	biz []*outCall
	// pending 是本端发起、尚未完成的调用。
	pending map[uint64]*outCall
	// inbound 是对端发起、尚未响应的调用。
	inbound map[uint64]*inCall
	// outBytes 是本方向仍计入预算的载荷字节。
	outBytes int64
	// inBytes 是对端方向在途的载荷字节。
	inBytes int64
	// err 是关闭原因；非 nil 表示 Session 已关闭。
	err error
	// changed 在在途调用数变化时关闭并替换，唤醒 WaitIdle。
	changed chan struct{}

	// late 计数迟到、重复或未知的 Response。
	late atomic.Int64
	// unknownCancel 计数找不到在途调用的 Cancel。
	unknownCancel atomic.Int64
	// closeOnce 保证关闭流程只执行一次。
	closeOnce sync.Once
}

// NewSession 在已认证的连接上创建 Session 并立即启动读写循环。
func NewSession(conn *Conn, opt Options) *Session {
	s := New(conn, opt)
	s.Start()
	return s
}

// New 创建 Session 但不启动读写循环；回调需要引用该 Session 时，先保存指针再调用 Start，
// 避免第一帧在赋值之前就被处理。
func New(conn *Conn, opt Options) *Session {
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
		prio:    make(chan prioFrame, opt.PriorityFrames),
		bizSig:  make(chan struct{}, 1),
		done:    make(chan struct{}),
		pending: map[uint64]*outCall{},
		inbound: map[uint64]*inCall{},
		changed: make(chan struct{}),
	}
	return s
}

// Start 启动读循环与写循环；只能调用一次。
func (s *Session) Start() {
	go s.readLoop()
	go s.writeLoop()
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

// envelope 构造本 Session 发出的帧外壳：协议版本、session_id，插件侧另填 source_plugin_id。
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

// setPayload 把业务载荷写入帧。
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

// controlEnvelope 以本 Session 的身份与协商版本构造控制帧。
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

// SendControlSync 发出控制帧并等待唯一 writer 把它完整写出（例如 ACTIVATED 写完后才打开 ready gate）。
func (s *Session) SendControlSync(kind pb.FrameKind, msg proto.Message) error {
	env, err := s.controlEnvelope(kind, msg)
	if err != nil {
		return err
	}
	return s.SendEnvelopeSync(env)
}

// SendEnvelopeSync 经优先通道发出调用方构造好的控制帧（例如带 ErrorInfo 的 SESSION_ERROR）并等待写出。
func (s *Session) SendEnvelopeSync(env *pb.Envelope) error {
	written := make(chan error, 1)
	if err := s.enqueuePrioFrame(prioFrame{env: env, written: written}); err != nil {
		return err
	}
	select {
	case err := <-written:
		return err
	case <-s.done:
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "Session 已关闭，控制帧可能未写出")
	}
}

// enqueuePrio 把帧放入有界优先通道；放不进说明连控制通道都无法推进，关闭 Session（CONTRACTS §1.2）。
func (s *Session) enqueuePrio(env *pb.Envelope) error { return s.enqueuePrioFrame(prioFrame{env: env}) }

// enqueuePrioFrame 是 enqueuePrio 的实现，可携带写结果通知。
func (s *Session) enqueuePrioFrame(f prioFrame) error {
	select {
	case <-s.done:
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "Session 已关闭")
	default:
	}
	select {
	case s.prio <- f:
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
	// 1. 有效期限取请求期限（或默认期限）与 ctx 期限中较早者。
	n := int64(len(req.Payload.Data))
	deadline := req.Deadline
	if deadline.IsZero() {
		deadline = time.Now().Add(s.opt.DefaultDeadline)
	}
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	// 2. 在 s.mu 下做准入：Session 未关闭、本方向未满、父调用在途（子请求不延长父期限）、期限未过、全局预算可占用。
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
	// 3. 分配编号、构造 REQUEST，登记为在途并放入业务队列。
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

	// 4. 等待结果；本地期限或 ctx 先到时放弃调用（见 abandon）。
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

// errorOf 把结果中的错误转为 *CallError；没有错误时返回 nil。
func errorOf(r Response) error {
	if r.Err != nil {
		return &CallError{Info: r.Err}
	}
	return nil
}

// abandon 在本地先结束调用，再按写出状态撤回请求或随后发送 Cancel（见 complete）。
func (s *Session) abandon(c *outCall, code pb.ErrorCode, reason pb.CancelReason, msg string) (Response, error) {
	s.complete(c, Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: msg}}, reason)
	r := <-c.result // 本地结束的结果，或先一步到达的真实结果
	return r, errorOf(r)
}

// CancelOutbound 结束本端所有在途请求（停止时 drain 到期后使用）：调用者得到 CANCELLED，对端按写出状态收到 Cancel。
func (s *Session) CancelOutbound(reason pb.CancelReason) {
	s.mu.Lock()
	calls := make([]*outCall, 0, len(s.pending))
	for _, c := range s.pending {
		calls = append(calls, c)
	}
	s.mu.Unlock()
	for _, c := range calls {
		s.complete(c, Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_CANCELLED, SafeMessage: "插件正在停止"}}, reason)
	}
}

// Inflight 返回在途的本端请求数与对端请求数。
func (s *Session) Inflight() (outbound, inbound int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending), len(s.inbound)
}

// WaitIdle 等待两个方向都没有在途调用；ctx 结束时返回其错误，Session 关闭时返回 nil。
func (s *Session) WaitIdle(ctx context.Context) error {
	for {
		s.mu.Lock()
		if len(s.pending) == 0 && len(s.inbound) == 0 {
			s.mu.Unlock()
			return nil
		}
		ch := s.changed
		s.mu.Unlock()
		select {
		case <-ch:
		case <-s.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// notifyChange 唤醒 WaitIdle；调用方必须持有 s.mu。
func (s *Session) notifyChange() {
	close(s.changed)
	s.changed = make(chan struct{})
}

// sendCancel 经优先通道发出本端调用 id 的 Cancel；只能在该 Request 已完整写出后调用。
func (s *Session) sendCancel(id uint64, reason pb.CancelReason) {
	env, err := s.controlEnvelope(pb.FrameKind_FRAME_KIND_CANCEL, &pb.Cancel{Reason: reason})
	if err != nil {
		return
	}
	env.InitiatorSide, env.CallId = s.opt.Local, id
	_ = s.enqueuePrio(env)
}

// complete 只完成一次调用：移出在途表、取消其 ctx（连带取消以它为父的子调用）并交付结果。
// reason 非 UNSPECIFIED 表示本端放弃调用，按写出状态处理线上请求：尚未写出则移出队列、永不上线，
// 正在写出则写完后再发 Cancel，已写出则立即发 Cancel（Cancel 不会越过自己的 Request）。
// 预算在调用结束且 writer 不再持有载荷时归还：正在写出的载荷由 writeRequest 写完后归还。
func (s *Session) complete(c *outCall, r Response, reason pb.CancelReason) bool {
	// 1. 在 s.mu 下结束调用，并按写出状态决定载荷与预算的去向。
	s.mu.Lock()
	if c.finished {
		s.mu.Unlock()
		return false
	}
	c.finished = true
	delete(s.pending, c.id)
	s.notifyChange()
	sendCancel := false
	switch c.state {
	case queued:
		s.unqueueLocked(c)
		c.state = withdrawn
		s.releaseLocked(c)
	case writing:
		c.cancelReason = reason
	case written:
		s.releaseLocked(c)
		sendCancel = reason != pb.CancelReason_CANCEL_REASON_UNSPECIFIED
	}
	s.mu.Unlock()

	// 2. 交付结果后再通知对端，调用者不等待 Cancel 入队。
	c.cancel()
	c.result <- r
	if sendCancel {
		s.sendCancel(c.id, reason)
	}
	return true
}

// unqueueLocked 把尚未写出的调用移出业务队列并丢弃其帧；调用方持有 s.mu。
func (s *Session) unqueueLocked(c *outCall) {
	for i, q := range s.biz {
		if q == c {
			copy(s.biz[i:], s.biz[i+1:])
			s.biz[len(s.biz)-1] = nil
			s.biz = s.biz[:len(s.biz)-1]
			break
		}
	}
	c.env = nil
}

// releaseLocked 归还调用占用的本方向与全局预算，重复调用无副作用；调用方持有 s.mu。
func (s *Session) releaseLocked(c *outCall) {
	s.outBytes -= c.bytes
	s.opt.Global.release(c.bytes)
	c.bytes = 0
}

// signalBiz 唤醒 writer 处理业务队列；已有未处理的信号时不阻塞。
func (s *Session) signalBiz() {
	select {
	case s.bizSig <- struct{}{}:
	default:
	}
}

// popBiz 取出队首请求并标记为正在写出；队列中只有尚未结束的调用。
func (s *Session) popBiz() (*outCall, *pb.Envelope) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.biz) == 0 {
		return nil, nil
	}
	c := s.biz[0]
	s.biz[0] = nil
	s.biz = s.biz[1:]
	c.state = writing
	return c, c.env
}

// writeLoop 是唯一的写者：优先通道（Response/Cancel/控制）总是先于业务 Request。
func (s *Session) writeLoop() {
	for {
		select {
		case f := <-s.prio:
			s.writePrio(f)
			continue
		case <-s.done:
			return
		default:
		}
		if c, env := s.popBiz(); c != nil {
			s.writeRequest(c, env)
			continue
		}
		select {
		case f := <-s.prio:
			s.writePrio(f)
		case <-s.bizSig:
		case <-s.done:
			return
		}
	}
}

// write 写出一帧；失败时关闭 Session。
func (s *Session) write(env *pb.Envelope) error {
	err := s.conn.WriteFrame(env)
	if err != nil {
		s.fail(fmt.Errorf("写帧失败: %w", err))
	}
	return err
}

// writePrio 写出优先通道的一帧，并按需通知写结果。
func (s *Session) writePrio(f prioFrame) {
	err := s.write(f.env)
	if f.written != nil {
		f.written <- err
	}
}

// writeRequest 写出一个 REQUEST；写出期间调用已结束时，写完才归还预算并按需发送 Cancel。
func (s *Session) writeRequest(c *outCall, env *pb.Envelope) {
	_ = s.write(env) // 写失败会关闭 Session，由 fail 结束调用
	s.mu.Lock()
	c.state, c.env = written, nil
	reason := pb.CancelReason_CANCEL_REASON_UNSPECIFIED
	if c.finished {
		s.releaseLocked(c)
		reason = c.cancelReason
	}
	s.mu.Unlock()
	if reason != pb.CancelReason_CANCEL_REASON_UNSPECIFIED {
		s.sendCancel(c.id, reason)
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

// dispatch 校验一帧并按种类分流；返回错误时读循环关闭 Session。
func (s *Session) dispatch(env *pb.Envelope) error {
	if err := validate.Envelope(env); err != nil {
		return err
	}
	peerID := "" // 对端为 Core 时帧上不带插件来源
	if s.opt.Local == pb.InitiatorSide_INITIATOR_SIDE_CORE {
		peerID = s.opt.SourcePluginID
	}
	if err := CheckPeer(env, s.opt.SessionID, s.opt.ProtocolMinor, peerID); err != nil {
		return err
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

// payloadOf 从帧中取出业务载荷。
func payloadOf(env *pb.Envelope) Payload {
	return Payload{env.GetPayloadEncoding(), env.GetSchemaId(), env.GetSchemaVersion(), env.GetPayload()}
}

// onRequest 处理对端 REQUEST：检查编号单调、阶段门禁、父调用与准入预算，通过后另起 goroutine 执行 Handler；
// 拒绝时经优先通道回复错误。
func (s *Session) onRequest(env *pb.Envelope) error {
	if env.GetInitiatorSide() != s.peer {
		return errors.New("REQUEST 的 initiator_side 不是对端")
	}
	id := env.GetCallId()
	n := int64(len(env.GetPayload()))
	deadline := time.UnixMilli(env.GetDeadlineUnixMs())
	// ParentCallID 先填入帧中声明的值，供 Admit 判断；父调用是否真在途由下面的查找校验。
	in := &Inbound{CallID: id, Target: env.GetTarget(), Payload: payloadOf(env), TraceID: env.GetTraceId(), Deadline: deadline,
		ParentCallID: env.GetParentCallId()}

	// 1. call_id 必须单调递增。
	s.mu.Lock()
	if id <= s.lastIn {
		s.mu.Unlock()
		return fmt.Errorf("对端 call_id %d 未单调递增（上一个为 %d）", id, s.lastIn)
	}
	s.lastIn = id
	s.mu.Unlock()
	// 2. 阶段门禁；不持锁调用 Admit：回调可以安全地读取调用方自己的状态。
	if s.opt.Admit != nil {
		if e := s.opt.Admit(in); e != nil {
			return s.respondErr(id, e.GetCode(), e.GetSafeMessage())
		}
	}
	// 3. 父调用必须是本端仍在途的请求，子请求继承其 ctx 且不延长其期限。
	s.mu.Lock()
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
	// 4. 对端方向的准入预算；通过后登记在途，另起 goroutine 执行 Handler，读循环不等待它。
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

// serve 执行一次对端调用并回复：Handler 返回后先移出在途表、归还预算，再经优先通道发出 RESPONSE。
func (s *Session) serve(ctx context.Context, cancel context.CancelFunc, in *Inbound, n int64) {
	defer cancel()
	// 1. 执行 Handler；没有结果也没有错误时按 INTERNAL 回复。
	r := s.handle(ctx, in)
	if r.Payload.SchemaID == "" && r.Err == nil {
		r.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "处理函数既没有结果也没有错误"}
	}
	// 2. 先结束在途登记并归还预算，再回复，避免对端收到响应后立即发起的请求被已完成的调用占额。
	s.mu.Lock()
	delete(s.inbound, in.CallID)
	s.inBytes -= n
	s.notifyChange()
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

// respondErr 经优先通道回复对端调用 id 一个错误。
func (s *Session) respondErr(id uint64, code pb.ErrorCode, msg string) error {
	env := s.envelope(pb.FrameKind_FRAME_KIND_RESPONSE)
	env.InitiatorSide, env.CallId = s.peer, id
	env.Error = &pb.ErrorInfo{Code: code, SafeMessage: msg}
	return s.enqueuePrio(env)
}

// onResponse 用 RESPONSE 完成对应的本端调用；迟到、重复或未知的响应只计数。
func (s *Session) onResponse(env *pb.Envelope) error {
	if env.GetInitiatorSide() != s.opt.Local {
		return errors.New("RESPONSE 的 initiator_side 不是本端")
	}
	s.mu.Lock()
	c := s.pending[env.GetCallId()]
	s.mu.Unlock()
	if c == nil || !s.complete(c, responseOf(env), pb.CancelReason_CANCEL_REASON_UNSPECIFIED) {
		s.late.Add(1) // 迟到、重复或未知：不第二次完成调用
	}
	return nil
}

// responseOf 把 RESPONSE 帧转为调用结果：保留载荷、安全消息与 trace_id，错误码按 PublicCode 公开，原值留作诊断。
func responseOf(env *pb.Envelope) Response {
	r := Response{Payload: payloadOf(env)}
	if e := env.GetError(); e != nil {
		r.RawErrorCode = e.GetCode()
		r.Err = &pb.ErrorInfo{Code: validate.PublicCode(e.GetCode()), SafeMessage: e.GetSafeMessage(), TraceId: e.GetTraceId()}
	}
	return r
}

// onCancel 取消对端调用的 Handler ctx；找不到在途调用时只计数。
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
		// 1. 记录原因，并在 s.mu 下取出两个方向的在途调用快照。
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
		// 2. 通知读写循环退出并关闭连接，阻塞中的读写随之返回。
		close(s.done)
		_ = s.conn.Close()
		// 3. 结束本端调用（排队中的请求随之移出队列），取消对端调用的 Handler。
		for _, c := range pending {
			s.complete(c, Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "Session 已断开"}}, pb.CancelReason_CANCEL_REASON_UNSPECIFIED)
		}
		for _, in := range inbound {
			in.cancel()
		}
	})
}
