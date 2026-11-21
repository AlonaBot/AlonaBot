package options

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Core 使用统一的会话编码；SDK 的类型留在 SDK 模块内。
var chatEscape = strings.NewReplacer("%", "%25", "#", "%23")
var chatUnescape = strings.NewReplacer("%25", "%", "%23", "#")

func (s *Service) chats(ctx context.Context, from configschema.OptionsFrom, q Query) (Page, error) {
	if s.opt.Chats == nil {
		return Page{}, unavailable("options source is unavailable: core.chats")
	}
	query := state.ChatQuery{Search: q.Search, ChatType: from.Filter["chat_type"], AdapterID: from.Filter["adapter_id"], Limit: q.Limit}
	if len(q.Values) > 0 {
		query.Keys = make([]state.ChatKey, 0, len(q.Values))
		for _, value := range q.Values {
			if key, ok := chatKey(value); ok {
				query.Keys = append(query.Keys, key)
			}
		}
	}
	chats, more, err := s.opt.Chats(ctx, query)
	if err != nil {
		return Page{}, fmt.Errorf("%w: %w", unavailable("read chats"), err)
	}
	page := Page{Options: make([]Option, 0, len(chats)), More: more}
	for _, chat := range chats {
		label := chat.Title
		if label == "" {
			label = chat.ChatID
		}
		page.Options = append(page.Options, Option{Value: chat.AdapterID + ":" + chatEscape.Replace(chat.ChatID), Label: label, Description: strings.TrimSpace(chat.AdapterID + " " + chat.ChatType)})
	}
	return page, nil
}

func chatKey(value string) (state.ChatKey, bool) {
	adapterID, encoded, found := strings.Cut(value, ":")
	if !found || encoded == "" || strings.Contains(encoded, "#") {
		return state.ChatKey{}, false
	}
	name := adapterID
	if strings.HasPrefix(name, validate.MarketPrefix) {
		name = strings.Replace(name, "-", "_", 1)
	}
	if _, err := validate.PluginID(name); err != nil {
		return state.ChatKey{}, false
	}
	chatID := chatUnescape.Replace(encoded)
	if chatEscape.Replace(chatID) != encoded {
		return state.ChatKey{}, false
	}
	return state.ChatKey{AdapterID: adapterID, ChatID: chatID}, true
}
