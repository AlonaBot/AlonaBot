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

// CommandMenuConfig 报告实例菜单模式，AutoSyncCommands 与 ClearCommandsOnDisable 对应平台配置。
type CommandMenuConfig = convert.CommandMenuConfig

// CommandMenuScope 定位 default、chat、chat_administrators 或 chat_member。
type CommandMenuScope = convert.CommandMenuScope

// CommandMenuEntry 保存一个命令及当前语言的说明。
type CommandMenuEntry = convert.CommandMenuEntry

// CommandMenu 保存整份 scope 与 language_code 的命令集合。
type CommandMenu = convert.CommandMenu

// CommandMenuController 支持按位置设置、读取、删除菜单；Core 串行化所有菜单操作。
// 每次成功只确认当前整份位置，错误完整返回，平台操作不自动重试。
type CommandMenuController interface {
	CommandMenuConfig() CommandMenuConfig
	SetCommandMenu(ctx context.Context, menu CommandMenu) *errcode.Error
	GetCommandMenu(ctx context.Context, scope CommandMenuScope, languageCode string) ([]CommandMenuEntry, *errcode.Error)
	DeleteCommandMenu(ctx context.Context, scope CommandMenuScope, languageCode string) *errcode.Error
}
