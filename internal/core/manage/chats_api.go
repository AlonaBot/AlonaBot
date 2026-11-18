package manage

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

const defaultChatLimit = 50
const chatPageMultiplier = 10

var chatEscape = strings.NewReplacer("%", "%25", "#", "%23")

type chatSummary struct {
	Session      string    `json:"session"`
	Title        string    `json:"title"`
	ChatType     string    `json:"chat_type"`
	InstanceID   string    `json:"instance_id"`
	LastActiveAt time.Time `json:"last_active_at"`
}

func (s *Service) chats(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	offset, limit := 0, defaultChatLimit
	for _, parameter := range []struct {
		name  string
		value *int
	}{{"offset", &offset}, {"limit", &limit}} {
		if query.Has(parameter.name) {
			value, err := strconv.Atoi(query.Get(parameter.name))
			if err != nil || value < 0 || (parameter.name == "limit" && value == 0) {
				respond(w, nil, invalid(parameter.name+" must be a valid pagination integer"))
				return
			}
			*parameter.value = value
		}
	}
	chatType := query.Get("chat_type")
	if chatType != "" && chatType != "private" && chatType != "group" && chatType != "channel" {
		respond(w, nil, invalid("unknown chat_type"))
		return
	}
	limit = min(limit, s.opt.Options.PageSize()*chatPageMultiplier)
	// 全量匹配提供精确 total，Unicode 搜索和排序沿用状态库的同一查询。
	chats, _, err := s.opt.Store.Chats(r.Context(), state.ChatQuery{Search: query.Get("q"), AdapterID: query.Get("adapter_id"), ChatType: chatType})
	if err != nil {
		respond(w, nil, err)
		return
	}
	start := min(offset, len(chats))
	end := start + min(limit, len(chats)-start)
	result := make([]chatSummary, 0, end-start)
	for _, chat := range chats[start:end] {
		result = append(result, chatSummary{chat.AdapterID + ":" + chatEscape.Replace(chat.ChatID), chat.Title, chat.ChatType, chat.InstanceID, chat.LastSeen})
	}
	respond(w, struct {
		Chats []chatSummary `json:"chats"`
		Total int           `json:"total"`
	}{result, len(chats)}, nil)
}
