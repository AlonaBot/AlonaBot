package hook

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// PrepareAfterSend 固定观察快照并等待容量；调用者必须在平台完成后调用 complete，或以 release 撤销。
// 容量包含已预留、等待执行与执行中的批次；两个回调合计只消费一次预留。
func (h *Runner) PrepareAfterSend(ctx context.Context, source router.Caller, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation) (complete func(*pb.SendResult, *pb.ErrorInfo), release func(), err error) {
	candidates := h.candidates(pb.HookPhase_HOOK_PHASE_AFTER_SEND)
	if len(candidates) == 0 {
		return nil, func() {}, nil
	}
	select {
	case h.slots <- struct{}{}:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-h.stopping:
		return nil, nil, errors.New("after-send observer is stopping")
	}
	h.mu.Lock()
	if h.closed || ctx.Err() != nil {
		h.mu.Unlock()
		<-h.slots
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		return nil, nil, errors.New("after-send observer is stopping")
	}
	h.pending.Add(1)
	h.mu.Unlock()
	var releases []func()
	finish := func() {
		for i := len(releases) - 1; i >= 0; i-- {
			releases[i]()
		}
		<-h.slots
		h.pending.Done()
	}
	if h.opt.AcquireActivity != nil {
		callers := make(map[router.Caller]bool)
		for _, c := range candidates {
			callers[c.caller] = true
		}
		if source.PluginID != "" && source.PluginID != router.DebugCallerID {
			callers[source] = true
		}
		for caller := range callers {
			done, acquireErr := h.opt.AcquireActivity(caller)
			if acquireErr != nil {
				finish()
				return nil, nil, acquireErr
			}
			releases = append(releases, done)
		}
	}
	var snapshot *router.Invocation
	if parent != nil {
		releases = append(releases, parent.Begin())
		snapshot = &router.Invocation{ProfileID: parent.ProfileID, Root: parent.Root, PluginID: parent.PluginID,
			Event: proto.CloneOf(parent.Event), DefaultTarget: proto.CloneOf(parent.DefaultTarget),
			Chain: slices.Clone(parent.Chain), Scopes: slices.Clone(parent.Scopes), TraceID: slices.Clone(parent.TraceID), Depth: parent.Depth, SendForbidden: true}
	}
	request := &pb.AfterSendRequest{Context: invocationContext(snapshot, true), Target: proto.CloneOf(target), Chain: proto.CloneOf(chain), Operation: op, EditedPlatformMessageId: editedMessageID(op, edited)}
	trace, _ := hex.DecodeString(wire.TraceID(ctx)) // 追踪值由 Wire 生成并在可信 Context 中编码。
	logger := slog.With("trace_id", hex.EncodeToString(trace))
	var once sync.Once
	release = func() { once.Do(finish) }
	stopCancellation := context.AfterFunc(h.ctx, release)
	revoke := release
	release = func() { stopCancellation(); revoke() }
	complete = func(result *pb.SendResult, safe *pb.ErrorInfo) {
		once.Do(func() {
			stopCancellation()
			request.Result, request.Error = proto.CloneOf(result), proto.CloneOf(safe)
			// 批次从平台完成起拥有独立预算，调用者返回和取消均不撤销已接纳的观察。
			observerCtx, cancel := context.WithTimeout(wire.WithTraceID(h.ctx, trace), h.opt.RequestTimeout)
			go func() {
				defer finish()
				defer cancel()
				select {
				case h.workers <- struct{}{}:
					defer func() { <-h.workers }()
				case <-observerCtx.Done():
					logger.Warn(fmt.Sprintf("After-send observation expired while waiting for an execution slot: %v", observerCtx.Err()), "error", observerCtx.Err())
					return
				}
				for _, candidate := range candidates {
					if err := observerCtx.Err(); err != nil {
						logger.Warn(fmt.Sprintf("Stopped after-send observation because its execution budget ended: %v", err), "error", err)
						return
					}
					request.RouteId = candidate.route.GetRouteId()
					if !h.current(candidate) {
						logger.Warn(fmt.Sprintf("Skipped after-send hook %s of component %s because its Session ended", request.RouteId, candidate.caller.PluginID), "plugin_id", candidate.caller.PluginID, "route_id", request.RouteId)
						continue
					}
					if safe := call(observerCtx, candidate, request, &pb.AfterSendResult{}, snapshot, true); safe != nil {
						logger.Warn(fmt.Sprintf("After-send hook %s of component %s failed: %s", request.RouteId, candidate.caller.PluginID, safe.SafeMessage), "plugin_id", candidate.caller.PluginID, "route_id", request.RouteId, "code", safe.Code, "error", safe.SafeMessage)
					}
				}
			}()
		})
	}
	return complete, release, nil
}

// Close 停止接纳并排空已预留批次；排空期限到达时取消观察，再等待引用归还。
func (h *Runner) Close(ctx context.Context) error {
	h.mu.Lock()
	if !h.closed {
		h.closed = true
		close(h.stopping)
		go func() { h.pending.Wait(); close(h.drained) }()
	}
	h.mu.Unlock()
	select {
	case <-h.drained:
		h.cancel()
		return ctx.Err()
	case <-ctx.Done():
		h.cancel()
		<-h.drained
		return ctx.Err()
	}
}
