package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

const (
	testBotToken  = "platform-test-token"
	testBotID     = 101
	testUserID    = 88
	testMessageID = 1
	testBotName   = "platform_bot"
	echoText      = "platform echo roundtrip"
)

type botCall struct {
	Method string          `json:"method"`
	Body   json.RawMessage `json:"body,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type botServer struct {
	server  *httptest.Server
	updates chan json.RawMessage
	replies chan map[string]any
	mu      sync.Mutex // HTTP 并发处理程序共同追加调用证据。
	calls   []botCall
}

func newBotServer() *botServer {
	bot := &botServer{updates: make(chan json.RawMessage), replies: make(chan map[string]any)}
	bot.server = httptest.NewServer(http.HandlerFunc(bot.serveHTTP))
	return bot
}

func (b *botServer) record(call botCall) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls = append(b.calls, call)
}

func (b *botServer) snapshot() []botCall {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]botCall(nil), b.calls...)
}

func (b *botServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := "/bot" + testBotToken + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.Error(w, "unexpected Bot API path", http.StatusNotFound)
		b.record(botCall{Method: r.URL.Path, Error: "unexpected Bot API path"})
		return
	}
	method := strings.TrimPrefix(r.URL.Path, prefix)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		b.record(botCall{Method: method, Error: err.Error()})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(raw) != 0 && !json.Valid(raw) {
		b.record(botCall{Method: method, Error: fmt.Sprintf("invalid JSON body: %s", raw)})
		http.Error(w, "invalid Bot API JSON", http.StatusBadRequest)
		return
	}
	b.record(botCall{Method: method, Body: raw})
	var result any
	switch method {
	case "getMe":
		result = map[string]any{"id": testBotID, "is_bot": true, "first_name": testBotName, "username": testBotName}
	case "getWebhookInfo":
		result = map[string]any{"url": "", "has_custom_certificate": false, "pending_update_count": 0}
	case "getUpdates":
		var request struct {
			Timeout int `json:"timeout"`
		}
		if err := json.Unmarshal(raw, &request); err != nil || request.Timeout < 0 {
			b.record(botCall{Method: method, Error: fmt.Sprintf("invalid polling request: %s", raw)})
			http.Error(w, "invalid polling request", http.StatusBadRequest)
			return
		}
		timer := time.NewTimer(time.Duration(request.Timeout) * time.Second)
		defer timer.Stop()
		select { // 按 Adapter 请求的 long poll 期限、实际更新或请求取消结束。
		case update := <-b.updates:
			result = []json.RawMessage{update}
		case <-timer.C:
			result = []json.RawMessage{}
		case <-r.Context().Done():
			return
		}
	case "sendMessage":
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			b.record(botCall{Method: method, Error: err.Error()})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		select { // 收到的文本由验收方核对，Core 退出时请求取消解除等待。
		case b.replies <- params:
		case <-r.Context().Done():
			return
		}
		result = map[string]any{"message_id": testMessageID, "date": time.Now().Unix(), "chat": map[string]any{"id": testUserID, "type": "private"}, "text": params["text"]}
	default:
		b.record(botCall{Method: method, Error: "unexpected Bot API method"})
		http.Error(w, "unexpected Bot API method", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result}); err != nil && r.Context().Err() == nil {
		b.record(botCall{Method: method, Error: err.Error()})
	}
}
