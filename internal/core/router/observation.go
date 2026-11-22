package router

import (
	"errors"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Outcome 记录一次实际调用的 Handler 结果，Err 为 nil 表示成功。
type Outcome struct {
	PluginID, RouteID string
	Err               *pb.ErrorInfo
	Waiting           bool
}

// Dispatch 汇总事件到结束或首次等待时的 Handler 结果。
type Dispatch struct {
	InstanceID, SourceEventID string
	Waiting                   bool
	Outcomes                  []Outcome
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

func (r *Router) observe(event *trustedEvent) {
	if r.dispatched == nil {
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
		if !call.state.done {
			call.state.result.Outcomes = append(call.state.result.Outcomes, Outcome{PluginID: inv.PluginID, RouteID: inv.RouteID, Waiting: true})
		}
		result, notify := r.finishDispatchLocked(call.state, true)
		r.mu.Unlock()
		if notify {
			r.dispatched(result)
		}
	}
	inv.OnWaitMatch = func(ev *pb.Event) {
		r.mu.Lock()
		defer r.mu.Unlock()
		call.state = r.dispatches[ev.GetEventId()]
		call.state.pending++ // 在等待结果交给 SDK 前，将恢复调用纳入新事件。
	}
}

func (r *Router) recordOutcome(inv *Invocation, response wire.Response, err error) {
	if inv.dispatch == nil {
		return
	}
	failure := response.Err
	if err != nil {
		var callError *wire.CallError
		if errors.As(err, &callError) {
			failure = callError.ErrorInfo
		} else {
			failure = &pb.ErrorInfo{Code: wire.CodeOf(err), SafeMessage: err.Error()}
		}
	} else if failure == nil {
		if _, err := decodeOutcome(response.Payload); err != nil {
			failure = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	state := inv.dispatch.state
	if !state.done {
		state.result.Outcomes = append(state.result.Outcomes, Outcome{PluginID: inv.PluginID, RouteID: inv.RouteID, Err: failure})
	}
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
		r.dispatched(result)
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
	return result, true
}
