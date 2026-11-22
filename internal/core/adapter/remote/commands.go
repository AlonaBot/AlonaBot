package remote

import (
	"context"
	"errors"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

var _ adapter.CommandSynchronizer = (*instance)(nil)

func (p *Process) startCommandSync(entry *instance) {
	entry.syncMu.Lock()
	defer entry.syncMu.Unlock()
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if running && p.opt.StartCommandSync != nil {
		p.opt.StartCommandSync(entry)
	}
}

// SyncCommands 经运行实例绑定的 Session 替换菜单，不自动重试。
func (entry *instance) SyncCommands(ctx context.Context, commands []*pb.CommandMenuItem) *pb.ErrorInfo {
	p := entry.process
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if !running {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	err := call(ctx, entry.session, wire.ServiceAdapterCommandsSync, &pb.AdapterCommandsSyncRequest{InstanceId: entry.id, Commands: commands}, &pb.AdapterCommandsSyncResponse{})
	if err != nil {
		var failure *wire.CallError
		if errors.As(err, &failure) {
			return failure.ErrorInfo
		}
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	return nil
}

// CommandMenuConfig 返回本次运行由 Adapter 报告的菜单配置，nil 表示传统同步。
func (entry *instance) CommandMenuConfig() *pb.CommandMenuConfig { return entry.commandMenu }

// CommandMenu 在当前运行 Session 上执行单个菜单位置操作，保留完整错误。
func (entry *instance) CommandMenu(ctx context.Context, operation string, menu *pb.CommandMenu) (*pb.CommandMenuResponse, *pb.ErrorInfo) {
	p := entry.process
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if !running {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	response := &pb.CommandMenuResponse{}
	err := call(ctx, entry.session, wire.ServiceAdapterCommandMenu, &pb.CommandMenuRequest{InstanceId: entry.id, Operation: operation, Menu: menu}, response)
	if err != nil {
		var failure *wire.CallError
		if errors.As(err, &failure) {
			return response, failure.ErrorInfo
		}
		return response, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	return response, nil
}
