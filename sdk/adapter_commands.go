package sdk

import (
	"context"
	"errors"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
)

func (r *adapterRuntime) commandMenu(ctx context.Context, in *wire.Inbound) wire.Response {
	request := &pb.CommandMenuRequest{}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	r.mu.Lock()
	entry, safe := r.operationEntry(request.InstanceId, request.InstanceGeneration)
	if safe != nil {
		r.mu.Unlock()
		return wire.Response{Err: safe}
	}
	core := entry.core
	r.mu.Unlock()
	// 与传统同步共用实际执行锁，RPC 到期后仍等待平台回调结束。
	core.commandSync.Lock()
	defer core.commandSync.Unlock()
	if err := ctx.Err(); err != nil {
		code := pb.ErrorCode_ERROR_CODE_CANCELLED
		if errors.Is(err, context.DeadlineExceeded) {
			code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
		}
		return dispatchError(code, err.Error())
	}
	r.mu.Lock()
	if entry.core != core || entry.busy || entry.instance == nil || core.stopped.Load() {
		r.mu.Unlock()
		return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is not running")
	}
	controller, supported := entry.instance.(adapter.CommandMenuController)
	r.mu.Unlock()
	if !supported {
		return dispatchError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support scoped command menus")
	}
	menu := convert.MenuFromProto(request.Menu)
	response := &pb.CommandMenuResponse{}
	var failure *errcode.Error
	switch request.Operation {
	case "set":
		failure = controller.SetCommandMenu(ctx, menu)
	case "get":
		var commands []adapter.CommandMenuEntry
		commands, failure = controller.GetCommandMenu(ctx, menu.Scope, menu.LanguageCode)
		response.Commands = convert.MenuToProto(adapter.CommandMenu{Commands: commands}).Commands
	case "delete":
		failure = controller.DeleteCommandMenu(ctx, menu.Scope, menu.LanguageCode)
	default:
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter command menu operation")
	}
	return adapterResponse(response, convert.ErrorToProto(failure))
}
