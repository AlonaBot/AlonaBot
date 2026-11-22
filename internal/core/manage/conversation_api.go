package manage

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/internal/core/conversation"
)

type conversationSummary struct {
	Key       string    `json:"key"`
	Entries   int       `json:"entries"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Service) conversationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/plugins/{id}/conversations", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := s.installed(r.Context(), id); err != nil {
			respond(w, nil, err)
			return
		}
		summaries, err := conversation.New(s.opt.Store).List(r.Context(), id)
		result := make([]conversationSummary, 0, len(summaries))
		for _, summary := range summaries {
			result = append(result, conversationSummary{summary.Key, summary.Entries, summary.UpdatedAt})
		}
		respond(w, result, err)
	})
}

// deleteConversation 在 ServeMux 清理路径前接收不透明的会话键，保留斜杠、冒号及点段。
// Require 已完成认证和 CSRF 检查；只解码一次，键值原样交给会话服务。
func (s *Service) deleteConversation(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodDelete {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.EscapedPath(), "/"), "/")
	if len(parts) != 6 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "plugins" || parts[4] != "conversations" {
		return false
	}
	id, err := url.PathUnescape(parts[3])
	if err != nil {
		respond(w, nil, invalid("invalid plugin path"))
		return true
	}
	key, err := url.PathUnescape(parts[5])
	if err != nil {
		respond(w, nil, invalid("invalid conversation path"))
		return true
	}
	if _, err := s.installed(r.Context(), id); err != nil {
		respond(w, nil, err)
		return true
	}
	err = conversation.New(s.opt.Store).Clear(r.Context(), id, key)
	if err != nil {
		// 会话服务的键校验返回普通错误；在进入状态库前的失败按请求错误呈现。
		// 会话键由服务校验，保留服务给出的完整错误文本。
		if key == "" || len(key) > conversationKeyBytes || !utf8.ValidString(key) {
			err = invalid(err.Error())
		}
	}
	respond(w, struct{}{}, err)
	return true
}

const conversationKeyBytes = 256
