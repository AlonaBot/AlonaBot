package adapter

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// CommandSynchronizer 替换实例的命令菜单，空列表清空菜单。
type CommandSynchronizer interface {
	SyncCommands(ctx context.Context, commands []*pb.CommandMenuItem) *pb.ErrorInfo
}

// CommandMenuController 提供按 scope 与 language 管理菜单的运行实例接口。
type CommandMenuController interface {
	CommandMenuConfig() *pb.CommandMenuConfig
	CommandMenu(ctx context.Context, operation string, menu *pb.CommandMenu) (*pb.CommandMenuResponse, *pb.ErrorInfo)
}
