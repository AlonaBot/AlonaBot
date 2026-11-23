package router

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/debugapi"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type debugObservation struct {
	traces  map[string]bool
	replies []debugapi.Reply
	done    chan Dispatch
}

// DebugEvent 注入标准 Adapter 事件，沿用真实实例的入站验证和路由；取消仅结束本次观察。
func (r *Router) DebugEvent(ctx context.Context, event *pb.Event) (debugapi.EventResult, error) {
	ctx, instance, ok := r.adapters.Pin(ctx, event.GetAdapterInstanceId())
	if !ok {
		return debugapi.EventResult{}, fmt.Errorf("adapter instance is unavailable")
	}
	injector, ok := instance.(interface {
		Inject(context.Context, *pb.Event) error
	})
	if !ok {
		return debugapi.EventResult{}, fmt.Errorf("adapter instance does not support event injection")
	}
	source := event.GetSourceEventId()
	observation := &debugObservation{traces: map[string]bool{}, done: make(chan Dispatch, 1)}
	r.mu.Lock()
	if r.debugEvents == nil {
		r.debugEvents = map[string]*debugObservation{}
	}
	if _, exists := r.debugEvents[source]; exists {
		r.mu.Unlock()
		return debugapi.EventResult{}, fmt.Errorf("source event is already being observed")
	}
	r.debugEvents[source] = observation
	r.mu.Unlock()
	defer func() { r.mu.Lock(); delete(r.debugEvents, source); r.mu.Unlock() }()
	if err := injector.Inject(ctx, event); err != nil {
		return debugapi.EventResult{}, err
	}
	select {
	case d := <-observation.done:
		result := debugapi.EventResult{TraceIDs: d.TraceIDs, Waiting: d.Waiting}
		r.mu.Lock()
		result.Replies = append([]debugapi.Reply(nil), observation.replies...)
		r.mu.Unlock()
		for _, o := range d.Outcomes {
			result.Outcomes = append(result.Outcomes, debugapi.Outcome{TraceID: o.TraceID, PluginID: o.PluginID, RouteID: o.RouteID, Waiting: o.Waiting, Error: debugError(o.Err)})
		}
		for _, diagnostic := range d.Diagnostics {
			result.Diagnostics = append(result.Diagnostics, debugapi.Diagnostic{Phase: diagnostic.Phase, TraceID: diagnostic.TraceID, PluginID: diagnostic.PluginID, RouteID: diagnostic.RouteID, Error: debugError(diagnostic.Err)})
		}
		return result, nil
	case <-ctx.Done():
		return debugapi.EventResult{}, ctx.Err()
	case <-r.ctx.Done():
		return debugapi.EventResult{}, r.ctx.Err()
	}
}

func debugError(err *pb.ErrorInfo) *debugapi.Error {
	if err == nil {
		return nil
	}
	raw, _ := protojson.Marshal(err)
	return &debugapi.Error{Code: strings.TrimPrefix(err.Code.String(), "ERROR_CODE_"), Message: err.SafeMessage, Details: raw}
}

func (r *Router) publishDispatch(d Dispatch) {
	r.mu.Lock()
	if observer := r.debugEvents[d.SourceEventID]; observer != nil {
		for _, trace := range d.TraceIDs {
			observer.traces[trace] = true
		}
		if !d.Started {
			observer.done <- d
		}
	}
	r.mu.Unlock()
	if r.dispatched != nil {
		r.dispatched(d)
	}
}

// ObserveReply 保存真实平台回执和同时返回的错误，归属正在观察的事件 trace。
func (r *Router) ObserveReply(trace []byte, operation string, target *pb.Target, chain *pb.MessageChain, receipt *pb.SendResult, safe *pb.ErrorInfo) {
	id := hex.EncodeToString(trace)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, observer := range r.debugEvents {
		if !observer.traces[id] {
			continue
		}
		to, _ := protojson.Marshal(target)
		content, _ := protojson.Marshal(chain)
		var result []byte
		if receipt != nil {
			result, _ = protojson.Marshal(proto.CloneOf(receipt))
		}
		observer.replies = append(observer.replies, debugapi.Reply{TraceID: id, Operation: operation, To: to, Chain: content, Result: result, Error: debugError(safe)})
	}
}
