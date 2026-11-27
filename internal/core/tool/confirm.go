package tool

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	corei18n "github.com/AlonaBot/AlonaBot/internal/core/i18n"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// 摘要和实际参数展示上限，以 UTF-8 字节计数。
const (
	maxSummaryBytes          = 1024
	maxArgumentsDisplayBytes = 2048
	buttonsCapability        = "message.buttons"
	editCapability           = "message.edit"
)

type invoker struct {
	catalog *Catalog
	options InvokeOptions
	mu      sync.Mutex // 保护 pending 和首次点击；不在锁内执行数据库或平台操作
	pending map[string]*confirmation
}

type confirmation struct {
	ctx      context.Context // 父调用结束、Session 失效或确认超时均结束本记录
	actor    *pb.Actor
	caller   router.Caller
	decision chan bool // 缓冲一个结果，点击处理无需等待平台 Send 返回
}

func (i *invoker) confirm(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	request := &pb.ToolConfirmRequest{}
	if err := decodeRequest(in, request); err != nil {
		return wire.Response{Err: err}
	}
	if request.GetOperation() == "" || len(request.GetSummary()) > maxSummaryBytes {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "operation is required and summary must not exceed 1024 bytes")
	}
	if _, ok := pb.ToolEffect_name[int32(request.GetEffect())]; !ok {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "unknown tool effect")
	}
	i.catalog.observe(caller.PluginID, request.GetOperation(), request.GetEffect())
	view, err := i.catalog.profileSnapshot(ctx, in)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read profile policies: "+err.Error())
	}
	policy, err := view.policy(caller.PluginID, request.GetOperation(), true)
	if err == nil && policy == pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED {
		policy, err = i.catalog.store.GetToolPolicy(ctx, caller.PluginID, request.GetOperation())
		if errors.Is(err, state.ErrNotFound) {
			policy, err = pb.ToolPolicy_TOOL_POLICY_ALLOW, nil
		}
	}
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read operation policy: "+err.Error())
	}
	if policy == pb.ToolPolicy_TOOL_POLICY_DENY {
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "operation policy denies confirmation")
	}
	if policy == pb.ToolPolicy_TOOL_POLICY_CONFIRM || request.GetRequireConfirmation() {
		if err := i.ask(ctx, caller, in.ParentAttach.(*router.Invocation), request.GetOperation(), request.GetSummary(), false); err != nil {
			return wire.Response{Err: err}
		}
	}
	return toolResponse(&pb.ToolConfirmResponse{})
}

func (i *invoker) ask(ctx context.Context, caller router.Caller, parent *router.Invocation, name, detail string, arguments bool) (result *pb.ErrorInfo) {
	// 1. 可信身份和目标来自父附件；能力检查在签发及发送之前完成。
	actor, target := parent.Event.GetActor(), parent.DefaultTarget
	if actor.GetExternalUserId() == "" || target.GetChatId() == "" || actor.GetAdapterInstanceId() != target.GetAdapterInstanceId() {
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "confirmation requires a trusted actor and default target").Err
	}
	if i.options.CheckSessionAccess != nil {
		if safe := i.options.CheckSessionAccess(ctx, target); safe != nil {
			return safe
		}
	}
	instance, ok := i.options.Adapters.Get(target.GetAdapterInstanceId())
	if !ok {
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "confirmation adapter instance is unavailable").Err
	}
	if !supports(instance, buttonsCapability) {
		return failure(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support buttons").Err
	}
	locales := []string{actor.GetLocale(), i.options.DefaultLocale()}
	text := func(key string, args map[string]string) (string, error) {
		return corei18n.Messages().Format(locales, key, args)
	}
	if arguments {
		detail = displayArguments(detail)
		if len(detail) > maxArgumentsDisplayBytes {
			detail = detail[:maxArgumentsDisplayBytes]
			for !utf8.ValidString(detail) {
				detail = detail[:len(detail)-1]
			}
			note, err := text("core.tool.truncated", nil)
			if err != nil {
				return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error()).Err
			}
			detail += "\n" + note
		}
	}
	body, err := text("core.tool.confirm", map[string]string{"name": name, "detail": detail})
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error()).Err
	}
	allowLabel, err := text("core.tool.allow", nil)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error()).Err
	}
	denyLabel, err := text("core.tool.deny", nil)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error()).Err
	}
	// 2. 记录先于 Send 可见，平台即时点击也由同一个等待者收束与编辑。
	ctx, cancel := context.WithTimeout(ctx, i.options.ConfirmTimeout)
	defer cancel()
	id := rand.Text()
	record := &confirmation{ctx: ctx, actor: actor, caller: caller, decision: make(chan bool, 1)}
	i.mu.Lock()
	i.pending[id] = record
	i.mu.Unlock()
	defer func() {
		i.mu.Lock()
		delete(i.pending, id)
		i.mu.Unlock()
		// 即使父调用已结束，持久化的 token 仍须实际删除。
		if err := i.options.Buttons.RevokeConfirm(context.WithoutCancel(ctx), id); err != nil {
			slog.Error(fmt.Sprintf("Failed to revoke confirmation tokens for component %s: %v", caller.PluginID, err), "plugin_id", caller.PluginID, "error", err)
			if result == nil {
				result = failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "revoke confirmation tokens: "+err.Error()).Err
			}
		}
	}()
	allow, deny, err := i.options.Buttons.IssueConfirm(ctx, caller, target, id)
	if err != nil {
		return failure(wire.CodeOf(err), err.Error()).Err
	}
	chain := textChain(body)
	chain.Segments = append(chain.Segments, &pb.Segment{Data: &pb.Segment_Buttons{Buttons: &pb.ButtonsSegment{Rows: []*pb.ButtonRow{{Buttons: []*pb.Button{
		{ActionId: "allow", Label: allowLabel, CoreToken: proto.String(allow)},
		{ActionId: "deny", Label: denyLabel, CoreToken: proto.String(deny)},
	}}}}}})
	sent, sendError := instance.Send(ctx, target, chain)
	if sendError != nil {
		return sendError
	}
	// 3. 取消与点击可能同时就绪，继续执行前仍以父调用的存活状态为准。
	var allowed bool
	select {
	case allowed = <-record.decision:
	case <-ctx.Done():
		return contextFailure(ctx).Err
	}
	if ctx.Err() != nil {
		return contextFailure(ctx).Err
	}
	key := "core.tool.denied"
	if allowed {
		key = "core.tool.allowed"
	}
	conclusion, err := text(key, map[string]string{"name": name})
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error()).Err
	}
	if editor, ok := instance.(adapter.Editor); ok && supports(instance, editCapability) {
		if _, err := editor.Edit(ctx, target, sent.GetPrimaryPlatformMessageId(), textChain(conclusion)); err != nil {
			slog.Warn(fmt.Sprintf("Failed to edit the confirmation result for component %s: %s", caller.PluginID, err.GetSafeMessage()), "plugin_id", caller.PluginID, "error", err.GetSafeMessage())
		}
	}
	if !allowed {
		return failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "user denied confirmation").Err
	}
	return nil
}

func (i *invoker) clicked(ctx context.Context, id string, allow bool, ev *pb.Event) {
	// Core 确认点击在 Router 入队前处理，沿用同一会话策略。
	if i.options.CheckSessionAccess != nil {
		if safe := i.options.CheckSessionAccess(ctx, ev.GetTarget()); safe != nil {
			if safe.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
				slog.Warn(fmt.Sprintf("Failed to admit confirmation event %s to its Session: %s", ev.GetEventId(), safe.GetSafeMessage()), "event_id", ev.GetEventId(), "error", safe.GetSafeMessage())
			}
			return
		}
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	record := i.pending[id]
	if record == nil || record.ctx.Err() != nil {
		return
	}
	if record.actor.GetExternalUserId() != ev.GetActor().GetExternalUserId() || record.actor.GetAdapterInstanceId() != ev.GetActor().GetAdapterInstanceId() {
		slog.Info(fmt.Sprintf("Discarded a confirmation click for component %s because it came from another actor", record.caller.PluginID), "plugin_id", record.caller.PluginID, "adapter_instance_id", ev.GetActor().GetAdapterInstanceId(), "external_user_id", ev.GetActor().GetExternalUserId())
		return
	}
	delete(i.pending, id)
	// 从表中移除即选定唯一结果；另一枚 token 也无法再次完成该记录。
	record.decision <- allow
}

func displayArguments(arguments string) string {
	var formatted bytes.Buffer
	// invoke 已校验 JSON，缩进只改变展示，不影响转发的原始参数。
	_ = json.Indent(&formatted, []byte(arguments), "", "  ")
	return formatted.String()
}

func textChain(text string) *pb.MessageChain {
	return &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}}
}

func supports(instance adapter.Instance, name string) bool {
	return slices.ContainsFunc(instance.Capabilities(), func(c *pb.Capability) bool { return c.GetName() == name && c.GetSupported() })
}
