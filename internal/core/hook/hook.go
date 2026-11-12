// Package hook 执行已激活插件的出站 Hook，并将递归禁令保存在 Core 调用附件中。
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

// BeforeSend 顺序运行改写 Hook；nil Chain 表示沿用当前 Chain，错误立即拒绝本次发送。
func (h *Runner) BeforeSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation) (*pb.MessageChain, *pb.ErrorInfo) {
	for _, candidate := range h.candidates(pb.HookPhase_HOOK_PHASE_BEFORE_SEND) {
		if !h.current(candidate) {
			continue
		}
		request := &pb.BeforeSendRequest{
			RouteId: candidate.route.GetRouteId(), Context: invocationContext(parent),
			Target: target, Chain: chain, Operation: op, EditedPlatformMessageId: editedMessageID(op, edited),
		}
		result := &pb.BeforeSendResult{}
		if safe := call(ctx, candidate, request, result, parent); safe != nil {
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
			RouteId: candidate.route.GetRouteId(), Context: invocationContext(parent),
			Target: target, Chain: chain, Result: result, Error: errInfo,
			Operation: op, EditedPlatformMessageId: editedMessageID(op, edited),
		}
		if safe := call(ctx, candidate, request, &pb.AfterSendResult{}, parent); safe != nil {
			slog.Warn("after-send hook failed", "plugin_id", candidate.caller.PluginID, "route_id", request.RouteId, "code", safe.Code, "error", safe.SafeMessage)
		}
	}
}

func invocationContext(parent *router.Invocation) *pb.InvocationContext {
	trusted := &pb.InvocationContext{SendForbidden: true}
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

func call(ctx context.Context, candidate candidate, request, result proto.Message, parent *router.Invocation) *pb.ErrorInfo {
	// 1. 附件只在 Core 内保存；协议上下文供插件读取，二者均携带禁令。
	invocation := &router.Invocation{PluginID: candidate.caller.PluginID, SendForbidden: true}
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
