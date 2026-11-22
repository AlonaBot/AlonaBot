// Package command 查询可用路径命令并管理平台命令菜单。
//
// List 接收 Handler 的 Context，返回 Core 筛选并本地化的 Entry（CONTRACTS §2.2、§2.3、§5.4）。
// Set、Get、Delete、Status 与 Sync 按实例和聊天位置管理菜单（CONTRACTS §8）。
package command

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

// Entry 保存生效的命令名、所属路由及本地化说明。
type Entry struct {
	PluginID, RouteID string
	Path              []string   // 生效主名的完整 token 路径。
	Aliases           [][]string // 其余未被遮蔽的名字，保持配置顺序。
	Description       string     // 空串表示没有登记说明。
}

// List 返回当前 Actor 在默认 Target 中可用的命令，按主名字典序排列。
// 后台或已结束的调用上下文返回 FAILED_PRECONDITION。
func List(ctx context.Context) ([]Entry, error) {
	if _, active := registry.Parent(ctx); !active {
		return nil, &sdk.Error{Code: errcode.FailedPrecondition, Message: "command list requires an active handler context"}
	}
	result, err := registry.Proto[*pb.CommandListRequest, *pb.CommandListResponse](wire.ServiceCommandList).Call(ctx, &pb.CommandListRequest{})
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(result.GetCommands()))
	for _, info := range result.GetCommands() {
		entry := Entry{PluginID: info.GetPluginId(), RouteID: info.GetRouteId(), Path: info.GetPath(), Description: info.GetDescription()}
		for _, alias := range info.GetAliases() {
			entry.Aliases = append(entry.Aliases, alias.GetPath())
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
