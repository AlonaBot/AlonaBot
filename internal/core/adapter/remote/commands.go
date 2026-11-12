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
