package router

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// DebugCommandMenu 在管理认证后向当前实例的菜单 worker 提交操作。
// 和自动同步、插件菜单共享队列及生命周期；平台回调仍由 SDK 的实例锁排空。
func (r *Router) DebugCommandMenu(ctx context.Context, instance adapter.Instance, request *pb.CommandMenuRequest) wire.Response {
	r.mu.Lock()
	worker := r.commandSyncs[instance.InstanceID()]
	r.mu.Unlock()
	_, bound := r.adapters.Resolve(ctx, instance.InstanceID())
	if !bound || worker == nil || worker.controller == nil {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter command menu worker is unavailable"))
	}
	if worker.instance != instance {
		return menuResponse(nil, menuError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter command menu worker is unavailable"))
	}
	operation := &menuOperation{ctx: ctx, request: request, result: make(chan wire.Response, 1), admin: true}
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
