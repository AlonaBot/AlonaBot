package convert

import pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"

// CommandMenuScope 定位聊天、管理员或指定成员的菜单；default 仅由 Core 管理。
type CommandMenuScope struct {
	Type   string
	ChatID string
	UserID string
}

// CommandMenuEntry 保存一个命令及当前语言的说明。
type CommandMenuEntry struct {
	Name        string
	Description string
}

// CommandMenu 保存一个 scope 与 language 的完整命令集合。
type CommandMenu struct {
	Scope        CommandMenuScope
	LanguageCode string
	Commands     []CommandMenuEntry
}

// CommandMenuConfig 保存 Adapter 报告的自动同步与退出自动模式时的清理选项。
type CommandMenuConfig struct {
	AutoSyncCommands       bool
	ClearCommandsOnDisable bool
}

// MenuToProto 复制菜单及位置，避免共享可变切片。
func MenuToProto(menu CommandMenu) *pb.CommandMenu {
	out := &pb.CommandMenu{Scope: &pb.CommandMenuScope{Type: menu.Scope.Type, ChatId: menu.Scope.ChatID, UserId: menu.Scope.UserID}, LanguageCode: menu.LanguageCode}
	for _, entry := range menu.Commands {
		out.Commands = append(out.Commands, &pb.CommandMenuEntry{Name: entry.Name, Description: entry.Description})
	}
	return out
}

// MenuFromProto 复制远端菜单及位置。
func MenuFromProto(menu *pb.CommandMenu) CommandMenu {
	out := CommandMenu{Scope: CommandMenuScope{Type: menu.GetScope().GetType(), ChatID: menu.GetScope().GetChatId(), UserID: menu.GetScope().GetUserId()}, LanguageCode: menu.GetLanguageCode()}
	for _, entry := range menu.GetCommands() {
		out.Commands = append(out.Commands, CommandMenuEntry{Name: entry.GetName(), Description: entry.GetDescription()})
	}
	return out
}
