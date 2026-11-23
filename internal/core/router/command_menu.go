package router

import (
	"context"
	"errors"
	"log/slog"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	menuModeAuto     = "auto"
	menuModeManual   = "manual"
	menuScopeDefault = "default"
	menuScopeChat    = "chat"
	menuReadyEvent   = "bot.command_menu_ready"
)

type menuOperation struct {
	admin   bool // 仅管理认证后的调试入口设置。
	ctx     context.Context
	caller  Caller
	request *pb.CommandMenuRequest
	result  chan wire.Response
}

// observeMenuChatLocked 只登记处理提示，由 worker 异步登记首次聊天并按已提交关系发布。
func (r *Router) observeMenuChatLocked(ev *pb.Event) {
	worker := r.commandSyncs[ev.GetAdapterInstanceId()]
	chatID := ev.GetTarget().GetChatId()
	if worker == nil || worker.controller == nil || chatID == "" || ev.GetAccountId() != worker.instance.AccountID() || ev.GetEventType() == menuReadyEvent {
		return
	}
	membershipChanged := ev.GetEventType() == "member.joined" && ev.GetEventPayload().GetMemberJoined().GetMemberExternalUserId() == ev.GetAccountId() ||
		ev.GetEventType() == "member.left" && ev.GetEventPayload().GetMemberLeft().GetMemberExternalUserId() == ev.GetAccountId()
	if worker.observed[chatID] && !membershipChanged {
		return
	}
	worker.observed[chatID] = true
	worker.pending[chatID] = worker.pending[chatID] || membershipChanged
	if worker.due.IsZero() {
		worker.due = time.Now().Add(commandSyncWindow)
	}
	select {
	case worker.wake <- struct{}{}:
	default:
	} // 合并提示，普通消息立即继续派发。
}

func (r *Router) runManagedMenus(worker *commandSync) {
	if r.menuStore == nil {
		slog.Error("command menu store is missing", "adapter_instance_id", worker.instance.InstanceID())
		return
	}
	r.reconcileMenus(worker.ctx, worker)
	for {
		r.mu.Lock()
		due := worker.due
		r.mu.Unlock()
		var timer *time.Timer
		var deadline <-chan time.Time
		if !due.IsZero() {
			timer = time.NewTimer(time.Until(due))
			deadline = timer.C
		}
		select {
		case <-worker.ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		case operation := <-worker.operations:
			if timer != nil {
				timer.Stop()
			}
			operation.result <- r.executeMenuOperation(worker, operation)
		case <-worker.wake:
			if timer != nil {
				timer.Stop()
			}
		case <-deadline:
			r.mu.Lock()
			worker.due = time.Time{}
			r.mu.Unlock()
			r.reconcileMenus(worker.ctx, worker)
		}
	}
}

// reconcileMenus 将模式切换、自动发布与 owner 回收放在同一个执行入口。
func (r *Router) reconcileMenus(ctx context.Context, worker *commandSync) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	account, err := r.menuStore.MenuAccount(ctx, worker.instance.AdapterID(), worker.instance.AccountID())
	if err != nil {
		return r.menuFailure(worker, err)
	}
	worker.account = account
	r.mu.Lock()
	pending := worker.pending
	worker.pending = make(map[string]bool)
	r.mu.Unlock()
	for chatID, membershipChanged := range pending {
		if membershipChanged {
			// 同一合并窗口内离群后重入，也必须重建原内容。
			worker.invalidateChatMenus(chatID)
		}
		// 首次发现路径只插入；成员关系以游标事务已提交的值为准。
		chat := state.CommandMenuChat{AdapterID: account.AdapterID, AccountID: account.AccountID, ChatID: chatID, Active: true}
		if err := r.menuStore.ObserveMenuChat(ctx, chat); err != nil {
			// 保存失败保留完整待处理关系，显式 Sync 可重新处理。
			r.mu.Lock()
			for id, changed := range pending {
				worker.pending[id] = worker.pending[id] || changed
			}
			r.mu.Unlock()
			return r.menuFailure(worker, err)
		}
	}
	rows, err := r.menuStore.MenuChats(ctx, account.AdapterID, account.AccountID)
	if err != nil {
		return r.menuFailure(worker, err)
	}
	for _, row := range rows {
		if !row.Active || !worker.chats[row.ChatID].Active {
			worker.invalidateChatMenus(row.ChatID)
		}
		worker.chats[row.ChatID] = row
	}
	r.mu.Lock()
	for id := range worker.chats {
		worker.observed[id] = true
	}
	r.mu.Unlock()
	positions, err := r.menuStore.MenuPositions(ctx, account.AdapterID, account.AccountID)
	if err != nil {
		return r.menuFailure(worker, err)
	}
	config := worker.controller.CommandMenuConfig()
	mode := menuModeManual
	if config.GetAutoSyncCommands() {
		mode = menuModeAuto
	}
	if worker.account.Mode != mode {
		clear(worker.published)
		worker.account.ClearAutomatic = account.Mode == menuModeAuto && config.GetClearCommandsOnDisable()
		worker.account.Mode, worker.account.Ready = mode, false
		// 先保存切换意图，部分自动发布后再转手动也能识别并清理已写入的位置。
		if err := r.menuStore.SaveMenuAccount(ctx, worker.account); err != nil {
			return r.menuFailure(worker, err)
		}
	}
	if mode == menuModeManual {
		for _, position := range positions {
			chat := worker.chats[position.ChatID]
			if position.ChatID != "" && !chat.Active {
				continue
			}
			remove := worker.account.ClearAutomatic && position.Mode == menuModeAuto
			if !remove && position.Mode == menuModeManual {
				allowed, err := r.menuOwnerAllowed(ctx, worker, position.OwnerPluginID, position.ChatID)
				if err != nil {
					return r.menuFailure(worker, err)
				}
				remove = !allowed
			}
			if remove {
				if err := r.deleteMenuPosition(ctx, worker, position); err != nil {
					return r.menuFailure(worker, err)
				}
			}
		}
	} else if err := r.publishAutomaticMenus(ctx, worker, positions); err != nil {
		return r.menuFailure(worker, err)
	}
	worker.account.Mode, worker.account.Ready, worker.account.LastError = mode, true, ""
	if err := r.menuStore.SaveMenuAccount(ctx, worker.account); err != nil {
		return r.menuFailure(worker, err)
	}
	if mode == menuModeManual {
		return r.notifyMenuReady(ctx, worker)
	}
	return nil
}

func (r *Router) menuFailure(worker *commandSync, err error) error {
	worker.account.AdapterID, worker.account.AccountID = worker.instance.AdapterID(), worker.instance.AccountID()
	worker.account.Ready, worker.account.LastError = false, err.Error()
	// 平台结果已发生时，请求取消仍需记录；数据库自身有现行 busy_timeout。
	saveErr := r.menuStore.SaveMenuAccount(context.WithoutCancel(worker.ctx), worker.account)
	err = errors.Join(err, saveErr)
	slog.Error("command menu synchronization failed", "adapter_instance_id", worker.instance.InstanceID(), "account_id", worker.instance.AccountID(), "error", err)
	return err
}

func (r *Router) menuOwnerAllowed(ctx context.Context, worker *commandSync, pluginID, chatID string) (bool, error) {
	r.mu.Lock()
	_, available := r.routeSourcesLocked()[pluginID]
	r.mu.Unlock()
	if !available {
		installed, err := r.menuStore.GetInstall(ctx, pluginID)
		if errors.Is(err, state.ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if installed.Uninstalled {
			return false, nil
		}
		enabled, err := r.menuStore.Enabled(ctx, pluginID)
		if err != nil || !enabled {
			return false, err
		}
	}
	allowed, err := r.menuStore.AdapterInstanceAllowed(ctx, pluginID, worker.instance.InstanceID())
	if err != nil || !allowed || chatID == "" {
		return allowed, err
	}
	target := menuTarget(worker, chatID)
	_, enabled, err := r.Profile(ctx, pluginID, target)
	if err != nil || !enabled {
		return false, err
	}
	denied, err := r.menuStore.HasGlobalDeny(ctx, state.GlobalDeny{Kind: "chat", AdapterID: target.AdapterId, SubjectID: chatID})
	return !denied, err
}

func menuTarget(worker *commandSync, chatID string) *pb.Target {
	return &pb.Target{AdapterId: worker.instance.AdapterID(), AdapterInstanceId: worker.instance.InstanceID(), AccountId: worker.instance.AccountID(), ChatId: chatID}
}

func menuPosition(worker *commandSync, menu *pb.CommandMenu, mode, owner string) state.CommandMenuPosition {
	return state.CommandMenuPosition{AdapterID: worker.instance.AdapterID(), AccountID: worker.instance.AccountID(), Scope: menu.GetScope().GetType(), ChatID: menu.GetScope().GetChatId(), UserID: menu.GetScope().GetUserId(), LanguageCode: menu.GetLanguageCode(), Mode: mode, OwnerPluginID: owner}
}

func positionMenu(position state.CommandMenuPosition) *pb.CommandMenu {
	return &pb.CommandMenu{Scope: &pb.CommandMenuScope{Type: position.Scope, ChatId: position.ChatID, UserId: position.UserID}, LanguageCode: position.LanguageCode}
}

func (worker *commandSync) invalidateChatMenus(chatID string) {
	for key, menu := range worker.published {
		if menu.GetScope().GetChatId() == chatID {
			delete(worker.published, key)
		}
	}
	for key := range worker.notified {
		if key.chatID == chatID {
			delete(worker.notified, key)
		}
	}
}

func (r *Router) deleteMenuPosition(ctx context.Context, worker *commandSync, position state.CommandMenuPosition) error {
	delete(worker.published, state.CommandMenuKey(position.Scope, position.ChatID, position.UserID, position.LanguageCode))
	if position.ChatID != "" {
		chat, exists, err := r.menuStore.MenuChat(ctx, position.AdapterID, position.AccountID, position.ChatID)
		if err != nil {
			return err
		}
		if exists && !chat.Active {
			return nil
		}
	}
	if _, failure := worker.controller.CommandMenu(ctx, "delete", positionMenu(position)); failure != nil {
		return &wire.CallError{ErrorInfo: failure}
	}
	return r.menuStore.DeleteMenuPosition(context.WithoutCancel(ctx), position)
}

func menuResponse(message *pb.CommandMenuResponse, err error) wire.Response {
	out := wire.Response{}
	if message != nil {
		encoded, encodeErr := proto.Marshal(message)
		if encodeErr != nil {
			return menuResponse(nil, encodeErr)
		}
		out.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: handlerSchemaVersion, Bytes: encoded}
	}
	if err == nil {
		return out
	}
	var call *wire.CallError
	if errors.As(err, &call) {
		out.Err = proto.CloneOf(call.ErrorInfo)
		if err.Error() != call.Error() {
			out.Err.SafeMessage = err.Error()
		}
		return out
	}
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	if errors.Is(err, context.Canceled) {
		code = pb.ErrorCode_ERROR_CODE_CANCELLED
	}
	if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	out.Err = &pb.ErrorInfo{Code: code, SafeMessage: err.Error()}
	return out
}

func menuError(code pb.ErrorCode, message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func (r *Router) commandMenuService(ctx context.Context, caller Caller, in *wire.Inbound) wire.Response {
	request := &pb.CommandMenuRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != handlerSchemaVersion {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected CommandMenuRequest protobuf schema version 1"))
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "decode command menu request: "+err.Error()))
	}
	parent, _ := in.ParentAttach.(*Invocation)
	if request.InstanceId == "" && parent != nil && in.ParentCallID != 0 {
		request.InstanceId = parent.Event.GetAdapterInstanceId()
	}
	r.mu.Lock()
	worker := r.commandSyncs[request.InstanceId]
	r.mu.Unlock()
	if worker == nil || worker.controller == nil {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support scoped command menus"))
	}
	operation := &menuOperation{ctx: ctx, caller: caller, request: request, result: make(chan wire.Response, 1)}
	select {
	case worker.operations <- operation:
	case <-ctx.Done():
		return menuResponse(nil, ctx.Err())
	case <-worker.ctx.Done():
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable"))
	}
	select {
	case response := <-operation.result:
		return response
	case <-ctx.Done():
		return menuResponse(nil, ctx.Err())
	case <-worker.ctx.Done():
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable"))
	}
}

func (r *Router) executeMenuOperation(worker *commandSync, operation *menuOperation) wire.Response {
	ctx, cancel := context.WithCancel(operation.ctx)
	stop := context.AfterFunc(worker.ctx, cancel)
	defer stop()
	defer cancel()
	if err := ctx.Err(); err != nil {
		return menuResponse(nil, err)
	}
	if worker.ctx.Err() != nil {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable"))
	}
	if _, active := r.PluginFor(operation.caller); !operation.admin && !active {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin session is no longer active"))
	}
	request := operation.request
	chatID := request.GetMenu().GetScope().GetChatId()
	allowed, err := true, error(nil)
	if !operation.admin {
		allowed, err = r.menuOwnerAllowed(ctx, worker, operation.caller.PluginID, chatID)
	}
	if err != nil {
		return menuResponse(nil, err)
	}
	if !allowed {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "plugin is not allowed to manage this command menu"))
	}
	switch request.Operation {
	case "status":
		positions, err := r.menuStore.MenuPositions(ctx, worker.instance.AdapterID(), worker.instance.AccountID())
		if err != nil {
			return menuResponse(nil, err)
		}
		response := &pb.CommandMenuResponse{Config: proto.CloneOf(worker.controller.CommandMenuConfig()), Ready: worker.account.Ready, LastError: worker.account.LastError}
		for _, position := range positions {
			response.Positions = append(response.Positions, &pb.CommandMenuPosition{Scope: positionMenu(position).Scope, LanguageCode: position.LanguageCode, Mode: position.Mode, OwnerPluginId: position.OwnerPluginID})
		}
		return menuResponse(response, nil)
	case "sync":
		clear(worker.published)
		return menuResponse(&pb.CommandMenuResponse{}, r.reconcileMenus(ctx, worker))
	case "set", "get", "delete":
		if err := checkMenu(request.Menu, request.Operation == "set", operation.admin); err != nil {
			return menuResponse(nil, err)
		}
	default:
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid command menu operation"))
	}
	chat, exists, err := r.menuStore.MenuChat(ctx, worker.instance.AdapterID(), worker.instance.AccountID(), chatID)
	if err != nil {
		return menuResponse(nil, err)
	}
	if exists && !chat.Active {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "bot has left this chat"))
	}
	if request.Operation != "get" && (worker.controller.CommandMenuConfig().GetAutoSyncCommands() || !worker.account.Ready) {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "manual command menus are not ready"))
	}
	// 准入在实际执行前重读；成功回执后才转移 owner，失败保留上次所有权。
	response, failure := worker.controller.CommandMenu(ctx, request.Operation, request.Menu)
	if failure != nil {
		return menuResponse(response, &wire.CallError{ErrorInfo: failure})
	}
	position := menuPosition(worker, request.Menu, menuModeManual, operation.caller.PluginID)
	if request.Operation == "set" {
		err = r.menuStore.SaveMenuPosition(context.WithoutCancel(ctx), position)
	}
	if request.Operation == "delete" {
		err = r.menuStore.DeleteMenuPosition(context.WithoutCancel(ctx), position)
	}
	if err != nil {
		return menuResponse(nil, r.menuFailure(worker, err))
	}
	if !exists && request.Operation == "set" {
		chat = state.CommandMenuChat{AdapterID: worker.instance.AdapterID(), AccountID: worker.instance.AccountID(), ChatID: chatID, Active: true}
		worker.chats[chatID] = chat
	}
	return menuResponse(response, nil)
}
