package router

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	maxMenuCommands    = 100
	maxMenuName        = 32
	maxMenuDescription = 256
)

func checkMenu(menu *pb.CommandMenu, writing, allowDefault bool) error {
	invalid := func(message string) error { return menuError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, message) }
	scope := menu.GetScope()
	switch scope.GetType() {
	case menuScopeDefault:
		if !allowDefault || scope.GetChatId() != "" || scope.GetUserId() != "" {
			return invalid("default menu scope is reserved for Core")
		}
	case menuScopeChat, "chat_administrators":
		if scope.GetChatId() == "" || scope.GetUserId() != "" {
			return invalid("chat menu scope requires chat_id and no user_id")
		}
	case "chat_member":
		if scope.GetChatId() == "" || scope.GetUserId() == "" {
			return invalid("chat_member menu scope requires chat_id and user_id")
		}
	default:
		return invalid("invalid command menu scope")
	}
	language := menu.GetLanguageCode()
	if language != "" && (len(language) != 2 || language[0] < 'a' || language[0] > 'z' || language[1] < 'a' || language[1] > 'z') {
		return invalid("menu language_code must be empty or two lowercase ASCII letters")
	}
	if !writing {
		return nil
	}
	if len(menu.Commands) > maxMenuCommands {
		return invalid(fmt.Sprintf("command menu has %d commands; maximum is %d", len(menu.Commands), maxMenuCommands))
	}
	names := make(map[string]bool)
	for _, command := range menu.Commands {
		name := command.GetName()
		if name == "" || len(name) > maxMenuName {
			return invalid(fmt.Sprintf("invalid menu command name %q", name))
		}
		for _, c := range name {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
				return invalid(fmt.Sprintf("invalid menu command name %q", name))
			}
		}
		if names[name] {
			return invalid(fmt.Sprintf("duplicate menu command %q", name))
		}
		names[name] = true
		description := command.GetDescription()
		if !utf8.ValidString(description) || utf8.RuneCountInString(description) == 0 || utf8.RuneCountInString(description) > maxMenuDescription {
			return invalid(fmt.Sprintf("invalid menu description for %q", name))
		}
	}
	return nil
}

// localizedMenus 在平台写入前完成 locale 映射、冲突与整份菜单检查。
func localizedMenus(scope *pb.CommandMenuScope, commands []*pb.CommandMenuItem) ([]*pb.CommandMenu, error) {
	languages := map[string]bool{"": true}
	translations := make([]map[string]string, len(commands))
	for i, command := range commands {
		translations[i] = make(map[string]string)
		locales := make([]string, 0, len(command.Descriptions))
		for locale := range command.Descriptions {
			locales = append(locales, locale)
		}
		slices.Sort(locales)
		for _, locale := range locales {
			language, _, _ := strings.Cut(locale, "-")
			language = strings.ToLower(language)
			description := command.Descriptions[locale]
			if previous, exists := translations[i][language]; exists && previous != description {
				return nil, menuError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("conflicting menu descriptions for %q in language %q", command.Name, language))
			}
			translations[i][language] = description
			languages[language] = true
		}
	}
	ordered := make([]string, 0, len(languages))
	for language := range languages {
		ordered = append(ordered, language)
	}
	slices.Sort(ordered)
	var menus []*pb.CommandMenu
	for _, language := range ordered {
		menu := &pb.CommandMenu{Scope: proto.CloneOf(scope), LanguageCode: language}
		for i, command := range commands {
			description := translations[i][language]
			if description == "" {
				description = command.Descriptions["en-US"]
			}
			if description == "" {
				description = command.Name
			}
			menu.Commands = append(menu.Commands, &pb.CommandMenuEntry{Name: command.Name, Description: description})
		}
		if err := checkMenu(menu, true, false); err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}
	return menus, nil
}

func (r *Router) publishAutomaticMenus(ctx context.Context, worker *commandSync, positions []state.CommandMenuPosition) error {
	// 1. 完整计算并检查所有聊天；非法命令或语言冲突保留已有远端菜单。
	var desired []*pb.CommandMenu
	languages := map[string]bool{"": true}
	chatIDs := make([]string, 0, len(worker.chats))
	for id, chat := range worker.chats {
		if chat.Active {
			chatIDs = append(chatIDs, id)
		}
	}
	slices.Sort(chatIDs)
	for _, chatID := range chatIDs {
		commands, err := r.commandMenuForTarget(ctx, worker.instance, menuTarget(worker, chatID))
		if err != nil {
			return err
		}
		menus, err := localizedMenus(&pb.CommandMenuScope{Type: menuScopeChat, ChatId: chatID}, commands)
		if err != nil {
			return err
		}
		for _, menu := range menus {
			languages[menu.LanguageCode] = true
		}
		desired = append(desired, menus...)
	}
	// 受管全局菜单始终为空，包含已知历史 default 语言。
	for _, position := range positions {
		if position.Scope == menuScopeDefault {
			languages[position.LanguageCode] = true
		}
	}
	ordered := make([]string, 0, len(languages))
	for language := range languages {
		ordered = append(ordered, language)
	}
	slices.Sort(ordered)
	for _, language := range ordered {
		desired = append(desired, &pb.CommandMenu{Scope: &pb.CommandMenuScope{Type: menuScopeDefault}, LanguageCode: language})
	}
	// 2. 跳过本 Session 的成功内容快照；单个成功回执立即记账再保存快照。
	written := make(map[string]bool)
	for _, menu := range desired {
		if err := ctx.Err(); err != nil {
			return err
		}
		position := menuPosition(worker, menu, menuModeAuto, "")
		key := state.CommandMenuKey(position.Scope, position.ChatID, position.UserID, position.LanguageCode)
		if chatID := menu.GetScope().GetChatId(); chatID != "" {
			chat, exists, err := r.menuStore.MenuChat(ctx, worker.instance.AdapterID(), worker.instance.AccountID(), chatID)
			if err != nil {
				return err
			}
			if exists && !chat.Active {
				delete(worker.published, key)
				continue
			}
			commands, err := r.commandMenuForTarget(ctx, worker.instance, menuTarget(worker, chatID))
			if err != nil {
				return err
			}
			current, err := localizedMenus(menu.Scope, commands)
			if err != nil {
				return err
			}
			menu.Commands = nil
			for _, candidate := range current {
				if candidate.LanguageCode == menu.LanguageCode {
					menu.Commands = candidate.Commands
					break
				}
			}
		}
		if proto.Equal(worker.published[key], menu) {
			written[key] = true
			continue
		}
		// 请求失败或结果未知、记账失败都使旧快照失效，下一次变化可重新发布。
		delete(worker.published, key)
		if _, failure := worker.controller.CommandMenu(ctx, "set", menu); failure != nil {
			return &wire.CallError{ErrorInfo: failure}
		}
		if err := r.menuStore.SaveMenuPosition(context.WithoutCancel(ctx), position); err != nil {
			return err
		}
		worker.published[key] = proto.CloneOf(menu)
		written[key] = true
	}
	// 3. 清理失效的 scope 与语言；离群的目标保留记录，等待重入恢复。
	for _, position := range positions {
		key := state.CommandMenuKey(position.Scope, position.ChatID, position.UserID, position.LanguageCode)
		if written[key] {
			continue
		}
		if position.ChatID != "" && !worker.chats[position.ChatID].Active {
			continue
		}
		if err := r.deleteMenuPosition(ctx, worker, position); err != nil {
			return err
		}
	}
	return nil
}
