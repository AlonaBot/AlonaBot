package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Outcome 记录一次实际调用的 Handler 结果，Err 为 nil 表示成功。
type Outcome struct {
	TraceID           string
	PluginID, RouteID string
	Err               *pb.ErrorInfo
	Waiting           bool
}

// Diagnostic 记录 Handler 调用前的失败，Phase 区分准入和参数解析。
type Diagnostic struct {
	Phase, TraceID, PluginID, RouteID string
	Err                               *pb.ErrorInfo
}

// Dispatch 汇总事件到结束或首次等待时的 Handler 结果与调用前诊断。
type Dispatch struct {
	Started                   bool
	TraceIDs                  []string
	InstanceID, SourceEventID string
	Waiting                   bool
	Outcomes                  []Outcome
	Diagnostics               []Diagnostic
}

// dispatchState 的 pending 包含普通链和已登记的延续；全部字段由 Router.mu 保护。
// 等待立即发布一次结果；命中等待时，延续的关联转交给新事件。
type dispatchState struct {
	eventID string
	pending int
	done    bool
	result  Dispatch
}

type dispatchCall struct{ state *dispatchState }

func (r *Router) recordDiagnostic(state *dispatchState, c candidate, trace []byte, phase string, failure *pb.ErrorInfo) {
	if state == nil {
		return
	}
	if len(trace) == 0 {
		trace = make([]byte, traceIDBytes)
		rand.Read(trace)
	}
	id := hex.EncodeToString(trace)
	r.mu.Lock()
	state.result.TraceIDs = append(state.result.TraceIDs, id)
	state.result.Diagnostics = append(state.result.Diagnostics, Diagnostic{Phase: phase, TraceID: id, PluginID: c.pluginID, RouteID: c.route.GetRouteId(), Err: failure})
	r.mu.Unlock()
	r.publishDispatch(Dispatch{Started: true, InstanceID: state.result.InstanceID, SourceEventID: state.result.SourceEventID, TraceIDs: []string{id}})
}

func (r *Router) observe(event *trustedEvent) {
	r.mu.Lock()
	observing := r.dispatched != nil || len(r.debugEvents) > 0
	r.mu.Unlock()
	if !observing {
		return
	}
	ev := event.retained
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.dispatches == nil {
		r.dispatches = make(map[string]*dispatchState)
	}
	event.dispatch = &dispatchState{eventID: ev.GetEventId(), pending: 1, result: Dispatch{InstanceID: ev.GetAdapterInstanceId(), SourceEventID: ev.GetSourceEventId()}}
	r.dispatches[ev.GetEventId()] = event.dispatch
}

func (r *Router) observeInvocation(inv *Invocation, state *dispatchState) {
	if state == nil {
		return
	}
	call := &dispatchCall{state: state}
	inv.dispatch = call
	trace := hex.EncodeToString(inv.TraceID)
	r.mu.Lock()
	state.result.TraceIDs = append(state.result.TraceIDs, trace)
	r.mu.Unlock()
	r.publishDispatch(Dispatch{Started: true, InstanceID: state.result.InstanceID, SourceEventID: state.result.SourceEventID, TraceIDs: []string{trace}})

	inv.OnContinuation = func() func() {
		r.mu.Lock()
		call.state.pending++ // 登记与原调用结束交接前保留，防止普通链提前完成。
		r.mu.Unlock()
		return func() {
			r.mu.Lock()
			state := call.state
			r.mu.Unlock()
			r.completeDispatch(state)
		}
	}
	inv.OnWait = func() {
		r.mu.Lock()
		if call.state == nil {
			r.mu.Unlock()
			return
		}
		if !call.state.done {
			call.state.result.Outcomes = append(call.state.result.Outcomes, Outcome{TraceID: trace, PluginID: inv.PluginID, RouteID: inv.RouteID, Waiting: true})
		}
		result, notify := r.finishDispatchLocked(call.state, true)
		r.mu.Unlock()
		if notify {
			r.publishDispatch(result)
		}
	}
	inv.OnWaitMatch = func(ev *pb.Event) {
		r.mu.Lock()

		call.state = r.dispatches[ev.GetEventId()]
		if call.state == nil {
			r.mu.Unlock()
			return
		}
		call.state.result.TraceIDs = append(call.state.result.TraceIDs, trace)
		call.state.pending++ // 在交付等待结果前纳入新事件。
		started := Dispatch{Started: true, InstanceID: call.state.result.InstanceID, SourceEventID: call.state.result.SourceEventID, TraceIDs: []string{trace}}
		r.mu.Unlock()
		r.publishDispatch(started)
	}
}

func (r *Router) recordOutcome(inv *Invocation, response wire.Response, err error) {
	if inv.dispatch == nil {
		return
	}
	failure := response.Err
	if err != nil {
		failure = invocationError(err)
	} else if failure == nil {
		if _, err := decodeOutcome(response.Payload); err != nil {
			failure = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	state := inv.dispatch.state
	if state != nil && !state.done {
		state.result.Outcomes = append(state.result.Outcomes, Outcome{TraceID: hex.EncodeToString(inv.TraceID), PluginID: inv.PluginID, RouteID: inv.RouteID, Err: failure})
	}
}

func invocationError(err error) *pb.ErrorInfo {
	var callError *wire.CallError
	if errors.As(err, &callError) {
		return callError.ErrorInfo
	}
	code := wire.CodeOf(err)
	if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	} else if errors.Is(err, context.Canceled) {
		code = pb.ErrorCode_ERROR_CODE_CANCELLED
	}
	return &pb.ErrorInfo{Code: code, SafeMessage: err.Error()}
}

func (r *Router) completeDispatch(state *dispatchState) {
	if state == nil {
		return
	}
	r.mu.Lock()
	state.pending--
	var result Dispatch
	var notify bool
	if state.pending == 0 {
		result, notify = r.finishDispatchLocked(state, false)
	}
	r.mu.Unlock()
	if notify {
		r.publishDispatch(result)
	}
}

// finishDispatchLocked 只发布一次快照；自动回复由普通链持有的 pending 覆盖。
func (r *Router) finishDispatchLocked(state *dispatchState, waiting bool) (Dispatch, bool) {
	if state.done {
		return Dispatch{}, false
	}
	state.done = true
	delete(r.dispatches, state.eventID)
	result := state.result
	result.Waiting = waiting
	result.Outcomes = slices.Clone(result.Outcomes)
	result.Diagnostics = slices.Clone(result.Diagnostics)
	result.TraceIDs = slices.Clone(result.TraceIDs)
	return result, true
}
