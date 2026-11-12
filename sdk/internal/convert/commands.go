package convert

import (
	"maps"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// CommandMenuItem 保存 SDK 自有的菜单项，供 adapter 包复用。
type CommandMenuItem struct {
	Name         string
	Descriptions map[string]string
}

// CommandMenuFromProto 将菜单协议快照复制为 Adapter 自有类型。
func CommandMenuFromProto(commands []*pb.CommandMenuItem) []CommandMenuItem {
	menu := make([]CommandMenuItem, len(commands))
	for i, command := range commands {
		menu[i] = CommandMenuItem{Name: command.GetName(), Descriptions: maps.Clone(command.GetDescriptions())}
	}
	return menu
}
