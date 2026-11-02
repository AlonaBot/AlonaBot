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
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"google.golang.org/protobuf/proto"
)

// defaultPriorityFrames 限定优先通道（Response/Cancel/控制帧）的容量。
const defaultPriorityFrames = 1024

// Payload 携带一次调用或响应的业务载荷。
type Payload struct {
	Encoding      pb.PayloadEncoding // Bytes 的编码
	SchemaID      string             // 载荷契约 ID；为空表示没有载荷
	SchemaVersion uint32             // 载荷契约版本
	Bytes         []byte             // 编码后的载荷，其长度计入准入预算
}

// Request 描述本端发起的一次调用。
type Request struct {
	Target  string // 被调用的 route/service 或 Core 服务名
	Payload Payload
	TraceID []byte // 16 字节业务追踪 ID；没有业务上下文时为空
	// Deadline 为零时使用 Options.DefaultDeadline；最终期限还会受 ctx 与父调用约束。
	Deadline time.Time
	// ParentCallID 只能由插件侧设置，引用同 Session 中仍在途的 Core→插件 请求。
	ParentCallID uint64
	// Attach 只保存在本端；对端以本调用为父发起子请求时，交给子请求的 Handler（Core 借此派生可信上下文）。
	Attach any
}

// Response 携带一次调用的结果；Payload 与 Err 可以同时存在（例如部分发送成功）。
type Response struct {
	Payload Payload       // 对端返回的业务结果；出错时也可能携带有效的部分结果
	Err     *pb.ErrorInfo // 对调用者公开的错误；对端发来的未知错误码已映射为 INTERNAL
	// 对端发来的原始错误码，只供内部诊断；与 Err.Code 不同说明发生了映射。
	RawErrorCode pb.ErrorCode
}

// Inbound 描述对端发来的一次调用。
type Inbound struct {
	CallID   uint64 // 对端分配的调用编号
	Target   string // 被调用的 route/service 名
	Payload  Payload
	TraceID  []byte    // 对端帧上的追踪 ID
	Deadline time.Time // 有效期限，已按父调用期限收紧
	// 父调用编号；Admit 看到的是帧上声明的值，Handler 看到的是已校验在途的值，0 表示没有。
	ParentCallID uint64
	// 父调用（本端发起）的 Request.Attach；无父调用时为 nil。
	ParentAttach any
}

// Handler 处理对端调用；ctx 在本端期限到达、对端取消、父调用结束或 Session 关闭时取消。
// 返回值即响应：没有业务结果时必须返回 Err。
type Handler func(ctx context.Context, in *Inbound) Response

// CallError 把对端或本端的 ErrorInfo 包装成 error。
type CallError struct {
	ErrorInfo *pb.ErrorInfo // 调用者可见的错误
}

// Error 返回「错误码: 安全消息」。
func (e *CallError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorInfo.GetCode(), e.ErrorInfo.GetSafeMessage())
}

// Code 返回错误码。
func (e *CallError) Code() pb.ErrorCode { return e.ErrorInfo.GetCode() }

// callErr 构造本地产生的调用错误。
func callErr(code pb.ErrorCode, msg string) *CallError {
	return &CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: msg}}
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
	Local         pb.InitiatorSide // 本端身份（CORE 或 PLUGIN）
	SessionID     []byte           // 握手分配的 16 字节 Session ID
	ProtocolMinor uint32           // 协商出的协议次版本，写入发出的帧并核对收到的帧
	// 插件的认证 ID：Local 为 PLUGIN 时写入发出的帧，Local 为 CORE 时用来核对对端帧。
	SourcePluginID string
	// Handler 处理对端的业务 REQUEST；为 nil 时回复 UNSUPPORTED。
	Handler Handler
	// Control 在读循环中同步处理非调用帧（控制帧与 EVENT），保证状态变化按帧顺序生效；返回错误关闭 Session。
	Control func(*pb.Envelope) error
	// Event 在读循环中接纳 EVENT，必须立即返回；处理结束后调用 release 归还载荷预算。
	// 为 nil 时 EVENT 交给 Control；release 可重复调用。
	Event func(env *pb.Envelope, release func())
	// Admit 在读循环中同步决定是否接受对端的业务 REQUEST（阶段门禁）；返回非 nil 时直接回该错误，不执行 Handler。
	// 必须在读循环中同步判断：状态变化（如 ACTIVATED）也在读循环中同步生效，帧顺序因此得到保证。
	Admit  func(*Inbound) *pb.ErrorInfo
	Global *Budget // 跨 Session 共享的全局预算；为 nil 时不限额（插件侧）

	// 以下可调值为零时使用 tune.Defaults()；正常运行由 Core 与 SDK 显式设置。

	// MaxInflight 限定每方向在途业务请求数。
	MaxInflight int
	// MaxBytes 限定每方向在途与排队载荷字节数。
	MaxBytes int64
	// PriorityFrames 覆盖优先通道容量。
	PriorityFrames int
	// DefaultDeadline 覆盖默认请求期限。
	DefaultDeadline time.Duration
}

// Stats 汇总 Session 的可观测计数。
type Stats struct {
	LateResponses  int64 // 迟到、重复或未知的 Response 数
	UnknownCancels int64 // 找不到对应在途调用的 Cancel 数
	DroppedEvents  int64 // 接收方向预算不足而丢弃的 EVENT 数
}

// callState 表示本端业务帧在写出方向上的进度，只前进不回退。
type callState int

// 请求的写出状态。
const (
	queued    callState = iota // 在业务队列中等待 writer
	writing                    // writer 正在写出
	written                    // 已完整写出
	withdrawn                  // 写出前被结束，已移出队列，永不上线
)

// outCall 跟踪本端发起的 REQUEST 或 EVENT；EVENT 只等待写出结果。
type outCall struct {
	id uint64 // 本端分配的调用编号
	// 待写出的业务帧，由 Session.mu 保护；移出队列或写出后置 nil，不再持有载荷。
	env      *pb.Envelope
	deadline time.Time // 有效期限，子调用不能超过它
	// Request.Attach，交给以本调用为父的子请求的 Handler。
	attach any
	// result 容量 1，只写一次。
	result chan Response
	// eventResult 非 nil 表示无业务回执的 EVENT；只通知写出结果，不登记到 pending。
	eventResult chan error

	// ctx 在调用存续期间有效，结束时取消（子调用以此派生）。
	ctx context.Context
	// cancel 取消 ctx。
	cancel context.CancelFunc

	// 以下字段由 Session.mu 保护。

	bytes int64 // 仍计入本方向与全局预算的载荷字节；归还后为 0，保证只归还一次
	state callState
	// finished 为 true 表示调用方已得到结果；REQUEST 同时移出 pending。
	finished bool
	// 写出过程中被本端放弃时，写完后发送 Cancel 的原因；UNSPECIFIED 表示不发。
	cancelReason pb.CancelReason
}

// priorityFrame 承载优先通道中的一帧；written 非 nil 时在写出完成后收到写结果。
type priorityFrame struct {
	env *pb.Envelope
	// written 为 nil 时不通知写结果；否则容量 1，收到一次写结果。
	written chan error
}

// inCall 跟踪对端发起、尚未响应的一次调用。
type inCall struct {
	bytes    int64     // 请求载荷字节数，响应时归还
	deadline time.Time // 有效期限，用于约束本端发起的子请求
	// cancel 取消 Handler 的 ctx（对端 Cancel 或 Session 关闭时）。
	cancel context.CancelFunc
}

// Session 在一条已认证连接上收发双向调用。
type Session struct {
	conn *Conn
	opt  Options          // 已补齐默认值的配置，创建后不变
	peer pb.InitiatorSide // 对端身份，由 opt.Local 推出

	// 优先通道：Response、Cancel 与控制帧，先于业务 REQUEST/EVENT 写出。
	priority chan priorityFrame
	// bizSig 容量 1，有新业务帧入队时唤醒 writer。
	bizSig chan struct{}
	// done 在 Session 关闭时关闭。
	done chan struct{}

	// mu 保护以下字段以及各 outCall 中标明由它保护的字段。
	mu     sync.Mutex
	nextID uint64 // 本端最近分配的调用编号
	lastIn uint64 // 对端最近的调用编号，用于检查单调递增
	// 等待写出的业务帧，按入队顺序排列，只含 queued 状态的发送。
	biz      []*outCall
	events   map[*outCall]struct{} // 尚未写完的 EVENT，由 mu 保护
	pending  map[uint64]*outCall   // 本端发起、尚未完成的调用
	inbound  map[uint64]*inCall    // 对端发起、尚未响应的调用
	outBytes int64                 // 本方向仍计入预算的载荷字节
	inBytes  int64                 // 对端方向在途的载荷字节
	err      error                 // 关闭原因；非 nil 表示 Session 已关闭
	// changed 在在途调用数变化时关闭并替换，唤醒 WaitIdle。
	changed chan struct{}

	// late 计数迟到、重复或未知的 Response。
	late atomic.Int64
	// unknownCancel 计数找不到在途调用的 Cancel。
	unknownCancel atomic.Int64
	droppedEvents atomic.Int64
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
	defaults := tune.Defaults()
	if opt.MaxInflight == 0 {
		opt.MaxInflight = defaults.SessionMaxInflight
	}
	if opt.MaxBytes == 0 {
		opt.MaxBytes = defaults.SessionMaxPayload
	}
	if opt.PriorityFrames == 0 {
		opt.PriorityFrames = defaultPriorityFrames
	}
	if opt.DefaultDeadline == 0 {
		opt.DefaultDeadline = defaults.RequestTimeout
	}
	peer := pb.InitiatorSide_INITIATOR_SIDE_CORE
	if opt.Local == pb.InitiatorSide_INITIATOR_SIDE_CORE {
		peer = pb.InitiatorSide_INITIATOR_SIDE_PLUGIN
	}
	s := &Session{
		conn: conn, opt: opt, peer: peer,
		priority: make(chan priorityFrame, opt.PriorityFrames),
		bizSig:   make(chan struct{}, 1),
		done:     make(chan struct{}),
		pending:  map[uint64]*outCall{},
		events:   map[*outCall]struct{}{},
		inbound:  map[uint64]*inCall{},
		changed:  make(chan struct{}),
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
	s.fail(errors.New("session is closed"))
	return nil
}

// Stats 返回计数快照。
func (s *Session) Stats() Stats {
	return Stats{LateResponses: s.late.Load(), UnknownCancels: s.unknownCancel.Load(), DroppedEvents: s.droppedEvents.Load()}
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
	env.PayloadEncoding, env.SchemaId, env.SchemaVersion, env.Payload = p.Encoding, p.SchemaID, p.SchemaVersion, p.Bytes
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
	return s.enqueuePriority(env)
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
	if err := s.enqueuePriorityFrame(priorityFrame{env: env, written: written}); err != nil {
		return err
	}
	select {
	case err := <-written:
		return err
	case <-s.done:
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "session is closed; the control frame may not have been written")
	}
}

// enqueuePriority 把帧放入有界优先通道；放不进说明连控制通道都无法推进，关闭 Session（CONTRACTS §1.2）。
func (s *Session) enqueuePriority(env *pb.Envelope) error {
	return s.enqueuePriorityFrame(priorityFrame{env: env})
}

// enqueuePriorityFrame 把一帧放入优先通道，可携带写结果通知。
func (s *Session) enqueuePriorityFrame(f priorityFrame) error {
	select {
	case <-s.done:
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "session is closed")
	default:
	}
	select {
	case s.priority <- f:
		return nil
	default:
		err := errors.New("priority channel is full; control frames and responses cannot proceed")
		s.fail(err)
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, err.Error())
	}
}

// Call 发起一次调用并等待结果。准入不足时立即返回 RESOURCE_EXHAUSTED，不等待槽位。
// 本地期限或取消先结束调用并返回，再尽力通知对端；已发出的副作用不会撤销，也不自动重试。
func (s *Session) Call(ctx context.Context, req Request) (Response, error) {
	// 1. 有效期限取请求期限（或默认期限）与 ctx 期限中较早者。
	n := int64(len(req.Payload.Bytes))
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
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "session is closed")
	}
	if len(s.pending) >= s.opt.MaxInflight || s.outBytes+n > s.opt.MaxBytes {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "in-flight request count or byte budget for this direction is full")
	}
	var parent *inCall
	if req.ParentCallID != 0 {
		parent = s.inbound[req.ParentCallID]
		if s.opt.Local != pb.InitiatorSide_INITIATOR_SIDE_PLUGIN || parent == nil {
			s.mu.Unlock()
			return Response{}, callErr(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "parent call does not exist or has finished")
		}
		if parent.deadline.Before(deadline) {
			deadline = parent.deadline // 子请求不得延长父期限
		}
	}
	if !deadline.After(time.Now()) {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "deadline already passed; request not sent")
	}
	if !s.opt.Global.acquire(n) {
		s.mu.Unlock()
		return Response{}, callErr(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "Core global payload budget is full")
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
		return s.abandon(c, pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, pb.CancelReason_CANCEL_REASON_DEADLINE_EXCEEDED, "call exceeded its deadline")
	case <-ctx.Done():
		code := pb.ErrorCode_ERROR_CODE_CANCELLED
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
		}
		return s.abandon(c, code, pb.CancelReason_CANCEL_REASON_CALLER_CANCELLED, "caller cancelled")
	}
}

// SendEvent 在 Core 侧发送 EVENT，等待帧写出；事件与 REQUEST 共用业务 FIFO 和载荷预算。
// ctx 在写出前结束时撤回帧；写出中结束时先返回，预算保留到 writer 完成。
func (s *Session) SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.opt.Local != pb.InitiatorSide_INITIATOR_SIDE_CORE {
		return callErr(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "only Core can send events")
	}
	env, err := s.controlEnvelope(pb.FrameKind_FRAME_KIND_EVENT, ev)
	if err != nil {
		return fmt.Errorf("encode BusEvent: %w", err)
	}
	env.Target = target
	n := int64(len(env.Payload))
	// 1. 事件只占字节预算，不占在途 Request 条数。
	s.mu.Lock()
	if s.err != nil {
		s.mu.Unlock()
		return callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "session is closed")
	}
	if s.outBytes+n > s.opt.MaxBytes || !s.opt.Global.acquire(n) {
		s.mu.Unlock()
		return callErr(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "event payload budget is full")
	}
	c := &outCall{env: env, bytes: n, eventResult: make(chan error, 1)}
	s.events[c] = struct{}{}
	s.outBytes += n
	s.biz = append(s.biz, c)
	s.mu.Unlock()
	s.signalBiz()
	// 2. 只等唯一 writer 的结果；取消通知不产生 CANCEL 帧。
	select {
	case err := <-c.eventResult:
		return err
	case <-ctx.Done():
		s.finishEvent(c, ctx.Err())
		return <-c.eventResult
	}
}

func (s *Session) finishEvent(c *outCall, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.finished {
		return
	}
	c.finished = true
	if c.state == queued {
		s.unqueueLocked(c)
		c.state = withdrawn
		s.releaseLocked(c)
		delete(s.events, c)
	}
	c.eventResult <- err // 容量为 1，结果仅由第一个结束者交付。
}

func (s *Session) writeEvent(c *outCall, env *pb.Envelope) {
	err := s.write(env)
	s.mu.Lock()
	defer s.mu.Unlock()
	c.state, c.env = written, nil
	s.releaseLocked(c)
	delete(s.events, c)
	if !c.finished {
		c.finished = true
		c.eventResult <- err
	}
}

// receiveEvent 在读循环中只占预算并交接，业务处理和释放由 Event 回调负责。
func (s *Session) receiveEvent(env *pb.Envelope) {
	n := int64(len(env.Payload))
	s.mu.Lock()
	if s.inBytes+n > s.opt.MaxBytes || !s.opt.Global.acquire(n) {
		s.mu.Unlock()
		s.droppedEvents.Add(1)
		return
	}
	s.inBytes += n
	s.mu.Unlock()
	var once sync.Once // 重复 release 只归还一次预算。
	s.opt.Event(env, func() {
		once.Do(func() {
			s.mu.Lock()
			s.inBytes -= n
			s.mu.Unlock()
			s.opt.Global.release(n)
		})
	})
}

// errorOf 把结果中的错误转为 *CallError；没有错误时返回 nil。
func errorOf(r Response) error {
	if r.Err != nil {
		return &CallError{ErrorInfo: r.Err}
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
		s.complete(c, Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_CANCELLED, SafeMessage: "plugin is stopping"}}, reason)
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
	_ = s.enqueuePriority(env)
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

// popBiz 取出队首 REQUEST 或 EVENT 并标记为正在写出；队列中只有尚未结束的发送。
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

// writeLoop 作为唯一的写者运行：优先通道（Response/Cancel/控制）先于业务 REQUEST/EVENT 写出。
func (s *Session) writeLoop() {
	for {
		select {
		case f := <-s.priority:
			s.writePriority(f)
			continue
		case <-s.done:
			return
		default:
		}
		if c, env := s.popBiz(); c != nil {
			if c.eventResult != nil {
				s.writeEvent(c, env)
			} else {
				s.writeRequest(c, env)
			}
			continue
		}
		select {
		case f := <-s.priority:
			s.writePriority(f)
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
		s.fail(fmt.Errorf("write frame: %w", err))
	}
	return err
}

// writePriority 写出优先通道的一帧，并按需通知写结果。
func (s *Session) writePriority(f priorityFrame) {
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

// readLoop 作为唯一的读者运行：调用帧与 Handler 执行分离，控制帧同步交给 Options.Control。
func (s *Session) readLoop() {
	for {
		env, err := s.conn.ReadFrame()
		if err != nil {
			s.fail(fmt.Errorf("read frame: %w", err))
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
	var payloadError *validate.Error
	if err := validate.Envelope(env); err != nil {
		if env.GetKind() != pb.FrameKind_FRAME_KIND_REQUEST || !errors.As(err, &payloadError) || payloadError.Code != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
			return err
		}
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
		return s.onRequest(env, payloadError)
	case pb.FrameKind_FRAME_KIND_RESPONSE:
		return s.onResponse(env)
	case pb.FrameKind_FRAME_KIND_CANCEL:
		return s.onCancel(env)
	case pb.FrameKind_FRAME_KIND_EVENT:
		if s.opt.Event != nil {
			s.receiveEvent(env)
			return nil
		}
	}
	if s.opt.Control == nil {
		return fmt.Errorf("unhandled frame kind %s", env.GetKind())
	}
	return s.opt.Control(env)
}

// payloadOf 从帧中取出业务载荷。
func payloadOf(env *pb.Envelope) Payload {
	return Payload{env.GetPayloadEncoding(), env.GetSchemaId(), env.GetSchemaVersion(), env.GetPayload()}
}

// onRequest 处理对端 REQUEST：检查编号单调、阶段门禁、父调用与准入预算，通过后另起 goroutine 执行 Handler；
// 拒绝时经优先通道回复错误。
func (s *Session) onRequest(env *pb.Envelope, payloadError *validate.Error) error {
	if env.GetInitiatorSide() != s.peer {
		return errors.New("REQUEST initiator_side is not the peer")
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
		return fmt.Errorf("peer call_id %d is not increasing (previous %d)", id, s.lastIn)
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
			return s.respondErr(id, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "parent call does not exist or has finished")
		}
		in.ParentCallID, in.ParentAttach, parentCtx = pid, parent.attach, parent.ctx
		if parent.deadline.Before(deadline) {
			deadline = parent.deadline // 子请求不得延长父期限
			in.Deadline = deadline
		}
	}
	// JSON 语法错误只拒绝本次调用；身份、编号、阶段和父调用检查仍先执行。
	if payloadError != nil {
		s.mu.Unlock()
		return s.respondErr(id, payloadError.Code, payloadError.Message)
	}
	// 4. 对端方向的准入预算；通过后登记在途，另起 goroutine 执行 Handler，读循环不等待它。
	if len(s.inbound) >= s.opt.MaxInflight || s.inBytes+n > s.opt.MaxBytes || !s.opt.Global.acquire(n) {
		s.mu.Unlock()
		return s.respondErr(id, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "in-flight request count or byte budget for this direction is full")
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
		r.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler returned neither a result nor an error"}
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
	_ = s.enqueuePriority(env)
}

// handle 调用 Handler；panic 转为 INTERNAL，不当作空成功（CONTRACTS §2.1）。
func (s *Session) handle(ctx context.Context, in *Inbound) (r Response) {
	if s.opt.Handler == nil {
		return Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "this side does not accept calls"}}
	}
	defer func() {
		if p := recover(); p != nil {
			r = Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "handler panicked"}}
		}
	}()
	return s.opt.Handler(ctx, in)
}

// respondErr 经优先通道回复对端调用 id 一个错误。
func (s *Session) respondErr(id uint64, code pb.ErrorCode, msg string) error {
	env := s.envelope(pb.FrameKind_FRAME_KIND_RESPONSE)
	env.InitiatorSide, env.CallId = s.peer, id
	env.Error = &pb.ErrorInfo{Code: code, SafeMessage: msg}
	return s.enqueuePriority(env)
}

// onResponse 用 RESPONSE 完成对应的本端调用；迟到、重复或未知的响应只计数。
func (s *Session) onResponse(env *pb.Envelope) error {
	if env.GetInitiatorSide() != s.opt.Local {
		return errors.New("RESPONSE initiator_side is not this side")
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
		return errors.New("CANCEL initiator_side is not the peer")
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
		events := make([]*outCall, 0, len(s.events))
		for c := range s.events {
			events = append(events, c)
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
			s.complete(c, Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "session disconnected"}}, pb.CancelReason_CANCEL_REASON_UNSPECIFIED)
		}
		for _, c := range events {
			s.finishEvent(c, callErr(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "session disconnected"))
		}
		for _, in := range inbound {
			in.cancel()
		}
	})
}
