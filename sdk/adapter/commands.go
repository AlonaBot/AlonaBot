package adapter

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
)

// CommandMenuItem 提供一个根命令及按 locale 选择的普通文本说明。
type CommandMenuItem = convert.CommandMenuItem

// CommandSynchronizer 替换实例的命令菜单，空列表清空菜单。
type CommandSynchronizer interface {
	SyncCommands(ctx context.Context, commands []CommandMenuItem) *errcode.Error
}
