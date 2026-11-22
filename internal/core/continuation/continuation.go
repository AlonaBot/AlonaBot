// Package continuation 管理 Handler 的一次性延续调用与同用户、同会话的消息等待。
//
// Manager 将登记保留到原调用结束，再在独立 goroutine 中发出延续请求；
// 等待只保存在内存中，命中事件由 Router 在命令层之前交付。
package continuation

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/clock"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const schemaVersion = 1

// Options 提供生效的延续调用限额、期限及 Router 的调用和事件准备入口。
type Options struct {
	Clock           clock.Clock   // nil 使用真实时钟；测试与 Scheduler 共用。
	Max             int           // 每个 Session 已登记与正在运行的总数，取 CONTINUATION_MAX。
	Timeout         time.Duration // 独立调用的期限，取 CONTINUATION_TIMEOUT。
	Call            func(ctx context.Context, inv *router.Invocation, continuationOf uint64, deadline time.Time) (wire.Response, error)
	Prepare         func(*pb.Event, router.Caller) (*pb.Event, error) // nil 时原样交付。
	Caller          func(pluginID string) (router.Caller, bool)       // 由 Router.Caller 查询当前激活绑定。
	AcquireActivity func(c router.Caller) (release func(), err error)
}

// Manager 串行化登记、等待命中与释放；网络调用及事件准备在锁外进行。
type Manager struct {
	opt        Options
	mu         sync.Mutex // 保护 calls、waits 与 delivering；只有一方取得交付权。
	calls      map[*router.Invocation]*continuation
	waits      map[waitKey]*waiter
	delivering map[*waiter]struct{} // Match 摘除键后，仍允许取消与 Release 结束等待。
}

type continuation struct {
	finish   func() // 可选的分发关联释放，由登记失败或运行结束调用。
	release  func() // 从登记到交付结束的活动引用；运行中的实际 Call 另由 Plugin 包装持有。
	caller   router.Caller
	parentID uint64
	ctx      context.Context
	cancel   context.CancelFunc
	started  bool
}

type waitKey struct{ instanceID, chatID, threadID, userID string }
type waiter struct {
	inv    *router.Invocation
	caller router.Caller
	result chan wire.Response // 单次交付缓冲，超时退出后匹配仍可完成交接。
}

var _ router.Continuations = (*Manager)(nil)

// New 创建内存中的延续登记与等待表。
func New(opt Options) *Manager {
	if opt.Clock == nil {
		opt.Clock = clock.Real{}
	}
	return &Manager{opt: opt, calls: make(map[*router.Invocation]*continuation), waits: make(map[waitKey]*waiter), delivering: make(map[*waiter]struct{})}
}

// Register 注册 core.continuation.start 与 core.wait。
func (m *Manager) Register(mux *router.Mux) {
	mux.Handle(wire.ServiceContinuationStart, m.start)
	mux.Handle(wire.ServiceWait, m.wait)
}

func (m *Manager) start(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	inv, ok := in.ParentAttach.(*router.Invocation)
	if in.ParentCallID == 0 || !ok || inv.PluginID != caller.PluginID || inv.Continuation || inv.Event.GetActor() == nil || inv.DefaultTarget == nil {
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "continuation requires an inbound Handler with an Actor and default Target")
	}
	if err := decode(in, &pb.ContinuationStartRequest{}); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.current(caller) {
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session released")
	}
	if ctx.Err() != nil {
		return failure(pb.ErrorCode_ERROR_CODE_CANCELLED, "continuation request cancelled")
	}
	if _, exists := m.calls[inv]; exists {
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "continuation already requested")
	}
	count := 0
	for _, call := range m.calls {
		if call.caller == caller {
			count++
		}
	}
	if count >= m.opt.Max {
		return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "continuation limit reached")
	}
	var release func()
	if m.opt.AcquireActivity != nil {
		var err error
		release, err = m.opt.AcquireActivity(caller)
		if err != nil {
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, err.Error())
		}
	}
	// 延续与原请求的有效期独立，只在释放或自身期限结束时取消。
	lifetime, cancel := context.WithCancel(context.Background())
	call := &continuation{caller: caller, parentID: in.ParentCallID, ctx: lifetime, cancel: cancel, release: release}
	if inv.OnContinuation != nil {
		call.finish = inv.OnContinuation()
	}
	m.calls[inv] = call
	return response(&pb.ContinuationStartResponse{})
}

// Finished 在原调用成功结束后交接登记，失败则立即释放名额。
func (m *Manager) Finished(inv *router.Invocation, ok bool) {
	m.mu.Lock()
	call := m.calls[inv]
	if call == nil || call.started {
		m.mu.Unlock()
		return
	}
	if !ok {
		delete(m.calls, inv)
		call.cancel()
		m.mu.Unlock()
		if call.release != nil {
			call.release()
		}
		if call.finish != nil {
			call.finish()
		}
		return
	}
	call.started = true
	m.mu.Unlock()
	go m.run(inv, call) // 原事件链立即继续，延续不占 Router 的分发槽。
}

func (m *Manager) run(inv *router.Invocation, call *continuation) {
	defer func() {
		call.cancel()
		m.mu.Lock()
		delete(m.calls, inv)
		m.mu.Unlock()
		if call.finish != nil {
			call.finish()
		}
		if call.release != nil {
			call.release()
		}
	}()
	if call.ctx.Err() != nil {
		return
	}
	deadline := m.opt.Clock.Now().Add(m.opt.Timeout)
	ctx, cancel := clock.WithDeadline(call.ctx, m.opt.Clock, deadline)
	defer cancel()
	deadline, _ = ctx.Deadline()
	result, err := m.opt.Call(ctx, inv, call.parentID, deadline)
	// 日志只保留身份与错误码，插件参数和错误文本可能含业务数据。
	if err != nil || result.Err != nil {
		code := result.Err.GetCode()
		if code == pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
			code = wire.CodeOf(err)
		}
		slog.Warn("continuation call failed", "plugin_id", inv.PluginID, "route_id", inv.RouteID, "code", code)
	}
}

func key(target *pb.Target, actor *pb.Actor) waitKey {
	return waitKey{target.GetAdapterInstanceId(), target.GetChatId(), target.GetThreadId(), actor.GetExternalUserId()}
}

func (m *Manager) wait(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 只接受延续父调用，等待身份取 Core 保留的原目标与 Actor。
	inv, ok := in.ParentAttach.(*router.Invocation)
	if in.ParentCallID == 0 || !ok || inv.PluginID != caller.PluginID || !inv.Continuation || inv.Event.GetActor() == nil || inv.DefaultTarget == nil {
		return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "wait requires a continuation Handler")
	}
	request := &pb.WaitRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	timeout := time.Duration(request.TimeoutMs) * time.Millisecond
	if timeout < time.Millisecond || timeout > m.opt.Timeout {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "wait timeout is out of range")
	}
	k := key(inv.DefaultTarget, inv.Event.Actor)
	w := &waiter{inv: inv, caller: caller, result: make(chan wire.Response, 1)}
	m.mu.Lock()
	if !m.current(caller) {
		m.mu.Unlock()
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session released")
	}
	if ctx.Err() != nil {
		m.mu.Unlock()
		return failure(pb.ErrorCode_ERROR_CODE_CANCELLED, "wait request cancelled")
	}
	if _, exists := m.waits[k]; exists {
		m.mu.Unlock()
		return failure(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "wait already exists for this Actor and Target")
	}
	if m.opt.AcquireActivity != nil {
		release, err := m.opt.AcquireActivity(caller)
		if err != nil {
			m.mu.Unlock()
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, err.Error())
		}
		defer release()
	}
	timer := m.opt.Clock.NewTimer(timeout)
	defer timer.Stop()
	m.waits[k] = w
	if inv.OnWait != nil {
		inv.OnWait()
	}
	m.mu.Unlock()
	// 2. 超时、父调用取消和交付竞争；取得表项的一方负责交付唯一结果。

	var expired wire.Response
	select {
	case result := <-w.result:
		return result
	case <-ctx.Done():
		expired = failure(pb.ErrorCode_ERROR_CODE_CANCELLED, "wait request cancelled")
	case <-timer.C():
		expired = failure(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "wait timed out")
	}
	m.mu.Lock()
	if m.waits[k] == w {
		delete(m.waits, k)
		m.mu.Unlock()
		return expired
	}
	if _, preparing := m.delivering[w]; preparing {
		delete(m.delivering, w)
		m.mu.Unlock()
		return expired
	}
	m.mu.Unlock()
	return <-w.result // Match 或 Release 已在锁内发布结果，保留先取得交付权的一方。
}

// Match 消费同一 Target 与 Actor 的下一条 message.created，并为等待 Session 准备事件。
func (m *Manager) Match(ev *pb.Event) bool {
	if ev.GetEventType() != "message.created" || ev.GetActor() == nil || ev.GetTarget() == nil {
		return false
	}
	k := key(ev.Target, ev.Actor)
	m.mu.Lock()
	w := m.waits[k]
	if w != nil {
		delete(m.waits, k)
		if !m.current(w.caller) {
			w.result <- failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session released")
			m.mu.Unlock()
			return false
		}
		m.delivering[w] = struct{}{}
	}
	m.mu.Unlock()
	if w == nil {
		return false
	}
	prepared := ev
	var prepareError error
	if m.opt.Prepare != nil {
		prepared, prepareError = m.opt.Prepare(ev, w.caller)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.current(w.caller) // Prepare 在锁外执行，交付前再次复核替换与 Release 的窗口。
	if _, waiting := m.delivering[w]; waiting {
		delete(m.delivering, w)
		if current && w.inv.OnWaitMatch != nil {
			w.inv.OnWaitMatch(ev)
		}
		if !current {
			w.result <- failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session released")
		} else if prepareError != nil {
			w.result <- failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "prepare waiting event: "+prepareError.Error())
		} else {
			w.result <- response(&pb.WaitResponse{Event: prepared})
		}
	}
	return current
}

func (m *Manager) current(c router.Caller) bool {
	current, ok := m.opt.Caller(c.PluginID)
	return ok && current == c
}

// Release 取消指定 Session 的延续调用，以 UNAVAILABLE 结束等待并丢弃待发登记。
func (m *Manager) Release(c router.Caller) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for inv, call := range m.calls {
		if call.caller == c {
			call.cancel()
			delete(m.calls, inv)
			if !call.started && call.finish != nil {
				call.finish()
			}
			if !call.started && call.release != nil {
				call.release()
			}
		}
	}
	for k, w := range m.waits {
		if w.caller == c {
			delete(m.waits, k)
			w.result <- failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session released")
		}
	}
	for w := range m.delivering {
		if w.caller == c {
			delete(m.delivering, w)
			w.result <- failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session released")
		}
	}
}

func decode(in *wire.Inbound, request proto.Message) error {
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return fmt.Errorf("expected %s protobuf schema version %d", validate.SchemaID(request), schemaVersion)
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return fmt.Errorf("decode continuation request: %w", err)
	}
	return nil
}
func failure(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}
func response(result proto.Message) wire.Response {
	encoded, err := proto.Marshal(result)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode continuation response: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}}
}
