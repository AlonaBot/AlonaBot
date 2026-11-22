package command

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

// Scope 定位 chat、chat_administrators 或 chat_member；话题共用聊天菜单。
type Scope = convert.CommandMenuScope

// Item 保存一个命令及当前语言的说明。
type Item = convert.CommandMenuEntry

// Menu 保存整份菜单；LanguageCode 为空表示平台缺省语言。
type Menu = convert.CommandMenu

// Position 返回最后成功写入的位置、模式与整份菜单的插件所有者。
type Position struct {
	Scope         Scope
	LanguageCode  string
	Mode          string
	OwnerPluginID string
}

// State 返回实例模式、就绪状态、上次同步错误与已持久化位置。
type State struct {
	AutoSyncCommands       bool
	ClearCommandsOnDisable bool
	Ready                  bool
	LastError              string
	Positions              []Position
}

// Set 覆盖整份菜单，成功后当前插件成为该位置的所有者；仅手动模式允许。
func Set(ctx context.Context, instanceID string, menu Menu) error {
	_, err := menuCall(ctx, instanceID, "set", menu)
	return err
}

// Get 读取平台当前完整菜单，保留平台与传输错误。
func Get(ctx context.Context, instanceID string, scope Scope, languageCode string) ([]Item, error) {
	response, err := menuCall(ctx, instanceID, "get", Menu{Scope: scope, LanguageCode: languageCode})
	if err != nil {
		return nil, err
	}
	return convert.MenuFromProto(&pb.CommandMenu{Commands: response.Commands}).Commands, nil
}

// Delete 删除明确指定的 scope 与 language；仅手动模式允许。
func Delete(ctx context.Context, instanceID string, scope Scope, languageCode string) error {
	_, err := menuCall(ctx, instanceID, "delete", Menu{Scope: scope, LanguageCode: languageCode})
	return err
}

// Status 读取 Core 保存的状态，不查询远端或启动额外同步。
func Status(ctx context.Context, instanceID string) (State, error) {
	response, err := menuCall(ctx, instanceID, "status", Menu{})
	if err != nil {
		return State{}, err
	}
	out := State{AutoSyncCommands: response.GetConfig().GetAutoSyncCommands(), ClearCommandsOnDisable: response.GetConfig().GetClearCommandsOnDisable(), Ready: response.Ready, LastError: response.LastError}
	for _, position := range response.Positions {
		menu := convert.MenuFromProto(&pb.CommandMenu{Scope: position.Scope, LanguageCode: position.LanguageCode})
		out.Positions = append(out.Positions, Position{Scope: menu.Scope, LanguageCode: menu.LanguageCode, Mode: position.Mode, OwnerPluginID: position.OwnerPluginId})
	}
	return out, nil
}

// Sync 在自动模式全量重建活跃聊天及受管 default 菜单；手动模式重新处理清理与就绪通知。
func Sync(ctx context.Context, instanceID string) error {
	_, err := menuCall(ctx, instanceID, "sync", Menu{})
	return err
}

func menuCall(ctx context.Context, instanceID, operation string, menu Menu) (*pb.CommandMenuResponse, error) {
	return registry.Proto[*pb.CommandMenuRequest, *pb.CommandMenuResponse](wire.ServiceCommandMenu).Call(ctx, &pb.CommandMenuRequest{InstanceId: instanceID, Operation: operation, Menu: convert.MenuToProto(menu)})
}
