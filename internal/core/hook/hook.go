// Package hook 执行已激活插件的三阶段 Hook，并将出站递归禁令保存在 Core 调用附件中。
//
// [Runner] 每次从 Router 读取候选，按 priority、plugin_id、route_id 顺序调用。
// BeforeSend 只接受 Chain 改写或错误；AfterSend 只观察回执，失败记录 WARN（CONTRACTS §2.3）。
package hook

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const schemaVersion = 1

// Runner 通过 Router 的当前插件绑定执行 Hook；各次执行独立收集候选。
type Runner struct {
	router *router.Router
}

// New 创建使用指定 Router 的 Hook 执行器。
func New(r *router.Router) *Runner { return &Runner{router: r} }

type candidate struct {
	plugin router.Plugin
	caller router.Caller // 与候选插件的 Session 同时固定，替换后丢弃旧候选
	route  *pb.RouteDescriptor
}

func (h *Runner) candidates(phase pb.HookPhase) []candidate {
	var candidates []candidate
	for _, plugin := range h.router.Plugins() {
		caller, ok := h.router.Caller(plugin.PluginID())
		if !ok {
			continue
		}
		plugin, ok = h.router.Plugin(caller.PluginID)
		if !ok || !plugin.Activated() {
			continue
		}
		for _, route := range plugin.Routes() {
			if route.GetKind() == pb.RouteKind_ROUTE_KIND_HOOK && route.GetHook().GetPhase() == phase {
				candidates = append(candidates, candidate{plugin: plugin, caller: caller, route: route})
			}
		}
	}
	slices.SortFunc(candidates, func(a, b candidate) int {
		if order := cmp.Compare(a.route.GetPriority(), b.route.GetPriority()); order != 0 {
			return order
		}
		if order := cmp.Compare(a.plugin.PluginID(), b.plugin.PluginID()); order != 0 {
			return order
		}
		return cmp.Compare(a.route.GetRouteId(), b.route.GetRouteId())
	})
	return candidates
}

func (h *Runner) current(candidate candidate) bool {
	caller, ok := h.router.Caller(candidate.caller.PluginID)
	return ok && caller == candidate.caller && candidate.plugin.Activated()
}

// PreEvent 顺序改写业务载荷；错误或 STOP 结束本次入站链，可信来源绑定始终沿用原事件。
func (h *Runner) PreEvent(ctx context.Context, ev *pb.Event) (*pb.EventPayload, bool, error) {
	current := proto.CloneOf(ev)
	for _, candidate := range h.candidates(pb.HookPhase_HOOK_PHASE_PRE_EVENT) {
		if !h.current(candidate) || !slices.Contains(candidate.route.GetHook().GetEventTypeFilter(), ev.GetEventType()) {
			continue
		}
		parent := &router.Invocation{Event: current, DefaultTarget: current.Target}
		request := &pb.PreEventRequest{RouteId: candidate.route.GetRouteId(), Context: invocationContext(parent, false), Event: current}
		result := &pb.PreEventResult{}
		safe := call(ctx, candidate, request, result, parent, false)
		if safe == nil && result.EventPayload != nil {
			if !businessPayloadMatches(ev.GetEventType(), result.EventPayload) {
				safe = failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "pre-event payload does not match event type "+ev.GetEventType())
			} else {
				current.EventPayload = result.EventPayload
				// 消息 ID 属于来源绑定；每个后继 Hook 都读取原来的 ID。
				switch ev.GetEventType() {
				case "message.created":
					current.EventPayload.GetMessageCreated().PlatformMessageId = ev.EventPayload.GetMessageCreated().GetPlatformMessageId()
				case "message.edited":
					current.EventPayload.GetMessageEdited().PlatformMessageId = ev.EventPayload.GetMessageEdited().GetPlatformMessageId()
				case "reaction.added":
					current.EventPayload.GetReactionAdded().PlatformMessageId = ev.EventPayload.GetReactionAdded().GetPlatformMessageId()
				}
			}
		}
		if safe != nil {
			slog.Warn("pre-event hook failed", "plugin_id", candidate.caller.PluginID, "route_id", request.RouteId, "code", safe.Code, "error", safe.SafeMessage)
			return current.EventPayload, false, fmt.Errorf("pre-event hook %s/%s: %s", candidate.caller.PluginID, request.RouteId, safe.SafeMessage)
		}
		if result.Propagation == pb.Propagation_PROPAGATION_STOP {
			return current.EventPayload, true, nil
		}
	}
	return current.EventPayload, false, nil
}

func businessPayloadMatches(eventType string, payload *pb.EventPayload) bool {
	switch eventType {
	case "message.deleted", "button.clicked", "scheduled":
		return false
	}
	name := strings.ReplaceAll(eventType, ".", "_")
	if !validate.IsEventType(eventType) && validate.IsAdapterEventType(eventType) {
		name = "adapter_event"
	}
	message := payload.ProtoReflect()
	field := message.WhichOneof(message.Descriptor().Oneofs().ByName("kind"))
	return field != nil && string(field.Name()) == name && message.Get(field).Message().IsValid()
}

// BeforeSend 顺序运行改写 Hook；nil Chain 表示沿用当前 Chain，错误立即拒绝本次发送。
func (h *Runner) BeforeSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation) (*pb.MessageChain, *pb.ErrorInfo) {
	for _, candidate := range h.candidates(pb.HookPhase_HOOK_PHASE_BEFORE_SEND) {
		if !h.current(candidate) {
			continue
		}
		request := &pb.BeforeSendRequest{
			RouteId: candidate.route.GetRouteId(), Context: invocationContext(parent, true),
			Target: target, Chain: chain, Operation: op, EditedPlatformMessageId: editedMessageID(op, edited),
		}
		result := &pb.BeforeSendResult{}
		if safe := call(ctx, candidate, request, result, parent, true); safe != nil {
			return nil, safe
		}
		if result.Chain != nil {
			chain = result.Chain
		}
	}
	return chain, nil
}

// AfterSend 将最终回执和错误逐个交给观察 Hook；其失败只记录日志。
func (h *Runner) AfterSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, result *pb.SendResult, errInfo *pb.ErrorInfo, parent *router.Invocation) {
	for _, candidate := range h.candidates(pb.HookPhase_HOOK_PHASE_AFTER_SEND) {
		if !h.current(candidate) {
			continue
		}
		request := &pb.AfterSendRequest{
			RouteId: candidate.route.GetRouteId(), Context: invocationContext(parent, true),
			Target: target, Chain: chain, Result: result, Error: errInfo,
			Operation: op, EditedPlatformMessageId: editedMessageID(op, edited),
		}
		if safe := call(ctx, candidate, request, &pb.AfterSendResult{}, parent, true); safe != nil {
			slog.Warn("after-send hook failed", "plugin_id", candidate.caller.PluginID, "route_id", request.RouteId, "code", safe.Code, "error", safe.SafeMessage)
		}
	}
}

func invocationContext(parent *router.Invocation, forbidden bool) *pb.InvocationContext {
	trusted := &pb.InvocationContext{SendForbidden: forbidden}
	if parent != nil {
		trusted.Actor = parent.Event.GetActor()
		trusted.DefaultTarget = parent.DefaultTarget
	}
	return trusted
}

func editedMessageID(op pb.SendOperation, edited string) *string {
	if op == pb.SendOperation_SEND_OPERATION_EDIT {
		return proto.String(edited)
	}
	return nil
}

func call(ctx context.Context, candidate candidate, request, result proto.Message, parent *router.Invocation, forbidden bool) *pb.ErrorInfo {
	// 1. 附件只在 Core 内保存；出站调用的附件和协议上下文均携带禁令。
	invocation := &router.Invocation{PluginID: candidate.caller.PluginID, SendForbidden: forbidden}
	if parent != nil {
		invocation.Event = parent.Event
		invocation.DefaultTarget = parent.DefaultTarget
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode hook request: "+err.Error())
	}
	response, err := candidate.plugin.Call(ctx, wire.Request{
		Target: candidate.route.GetRouteId(), Attach: invocation,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: schemaVersion, Bytes: encoded},
	})
	if response.Err != nil {
		return response.Err
	}
	if err != nil {
		var callError *wire.CallError
		if errors.As(err, &callError) {
			return callError.ErrorInfo
		}
		code := pb.ErrorCode_ERROR_CODE_INTERNAL
		if errors.Is(err, context.Canceled) {
			code = pb.ErrorCode_ERROR_CODE_CANCELLED
		} else if errors.Is(err, context.DeadlineExceeded) {
			code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
		}
		return failure(code, err.Error())
	}
	// 2. 插件返回值是信任边界，须匹配本阶段的响应类型。
	payload := response.Payload
	if payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || payload.SchemaID != validate.SchemaID(result) || payload.SchemaVersion != schemaVersion {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, fmt.Sprintf("expected %s protobuf schema version 1", result.ProtoReflect().Descriptor().Name()))
	}
	if err := proto.Unmarshal(payload.Bytes, result); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "decode hook result: "+err.Error())
	}
	return nil
}

func failure(code pb.ErrorCode, message string) *pb.ErrorInfo {
	return &pb.ErrorInfo{Code: code, SafeMessage: message}
}
