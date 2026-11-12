package adapter

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// CommandSynchronizer 替换实例的命令菜单，空列表清空菜单。
type CommandSynchronizer interface {
	SyncCommands(ctx context.Context, commands []*pb.CommandMenuItem) *pb.ErrorInfo
}
