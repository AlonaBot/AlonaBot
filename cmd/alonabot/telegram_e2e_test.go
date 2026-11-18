//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/adapters/telegram/telegramtest"
)

// TestTelegramDevelopment 经真实目录入口验证两账号、示例媒体上传、按钮编辑与发送失败。
func TestTelegramDevelopment(t *testing.T) {
	a := telegramtest.NewServer(t, "development-token-a")
	b := telegramtest.NewServer(t, "development-token-b")
	a.SetMe(101, "bot_a")
	b.SetMe(202, "bot_b")
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{
		"listen":"127.0.0.1:0",
		"adapters":{"telegram":{"package":"alonabot-adapter-telegram","default_instance":"tg_a","instances":{
			"tg_a":{"token":"development-token-a","api_base_url":%q,"proxy_mode":"none"},
			"tg_b":{"token":"development-token-b","api_base_url":%q,"proxy_mode":"none"}
		}}}}`, a.URL(), b.URL()))
	c := startEmptyCore(t, dir, "../../adapters/telegram", "../../examples/plugins/echo", "../../examples/plugins/media")
	for _, id := range []string{"alonabot-adapter-telegram", "echo", "media"} {
		developmentPID(t, c, id)
	}
	telegramEventually(t, c, func() bool { return a.Calls("getUpdates") > 0 && b.Calls("getUpdates") > 0 })
	defer func() { t.Logf("Core log:\n%s", c.log()) }()

	t.Run("accounts_and_topic", func(t *testing.T) {
		for _, server := range []*telegramtest.Server{a, b} {
			telegramMessage(t, server, 1, map[string]any{"text": "hello", "chat": map[string]any{"id": 9, "type": "private"}})
			requests := telegramRequests(t, c, server, 1)
			if requests[0].Method != "sendMessage" || requests[0].Params["text"] != "hello" || requests[0].Params["chat_id"] != float64(9) {
				t.Fatalf("text reply: %+v", requests)
			}
		}
		telegramMessage(t, a, 2, map[string]any{"text": "topic", "chat": map[string]any{"id": -99, "type": "supergroup"}, "is_topic_message": true, "message_thread_id": 42})
		requests := telegramRequests(t, c, a, 2)
		if requests[1].Params["message_thread_id"] != float64(42) || requests[1].Params["chat_id"] != float64(-99) || b.Calls("sendMessage") != 1 {
			t.Fatalf("topic or instance isolation: a=%+v b=%+v", requests, b.Requests())
		}
	})
	t.Run("incoming_file_id", func(t *testing.T) {
		before := a.Calls("getFile")
		telegramMessage(t, a, 3, map[string]any{"caption": "original caption", "photo": []any{map[string]any{"file_id": "incoming-file-id", "width": 64, "height": 64, "file_size": 100}}, "chat": map[string]any{"id": 9, "type": "private"}})
		requests := telegramRequests(t, c, a, 3)
		photo := requests[2]
		if photo.Method != "sendPhoto" || photo.Params["photo"] != "incoming-file-id" || photo.Params["caption"] != "original caption" || len(photo.Files) != 0 || a.Calls("getFile") != before || b.Calls("sendPhoto") != 0 {
			t.Fatalf("incoming image: %+v getFile=%d", photo, a.Calls("getFile")-before)
		}
	})
	t.Run("upload_album_button_edit", func(t *testing.T) {
		telegramMessage(t, a, 4, map[string]any{"text": "/media", "chat": map[string]any{"id": 9, "type": "private"}})
		requests := telegramRequests(t, c, a, 5)
		album, tail := requests[3], requests[4]
		if album.Method != "sendMediaGroup" || tail.Method != "sendMessage" || tail.Params["text"] != "Click to edit this text." {
			t.Fatalf("album order: %+v", requests[3:])
		}
		sample := developmentRead(t, "../../examples/plugins/media/assets/sample.png")
		if len(album.Files) != 2 || !bytes.Equal(album.Files["media_0"], sample) || !bytes.Equal(album.Files["media_1"], sample) {
			t.Fatalf("uploaded bytes differ: %+v", album.Files)
		}
		media := album.Params["media"].([]any)
		if len(media) != 2 || media[0].(map[string]any)["media"] != "attach://media_0" || media[1].(map[string]any)["media"] != "attach://media_1" || album.Params["reply_markup"] != nil {
			t.Fatalf("album media: %+v", album)
		}
		keyboard := tail.Params["reply_markup"].(map[string]any)["inline_keyboard"].([]any)
		button := keyboard[0].([]any)[0].(map[string]any)
		token := button["callback_data"].(string)
		if len(token) != 43 || button["text"] != "Edit text" {
			t.Fatalf("button: %+v", button)
		}
		telegramUpdate(t, a, map[string]any{"update_id": 5, "callback_query": map[string]any{
			"id": "development-click", "from": map[string]any{"id": 88, "first_name": "User", "language_code": "en"}, "data": token,
			"message": map[string]any{"message_id": 6, "chat": map[string]any{"id": 9, "type": "private"}},
		}})
		requests = telegramRequests(t, c, a, 7)
		if !reflect.DeepEqual([]string{requests[5].Method, requests[6].Method}, []string{"answerCallbackQuery", "editMessageText"}) || requests[6].Params["message_id"] != float64(6) || requests[6].Params["text"] != "Text edited." || requests[6].Params["chat_id"] != float64(9) || b.Calls("editMessageText") != 0 {
			t.Fatalf("callback and original text edit: %+v", requests[5:])
		}
	})
	t.Run("tail_rejected_without_album_retry", func(t *testing.T) {
		before := len(b.Requests())
		groups, texts := b.Calls("sendMediaGroup"), b.Calls("sendMessage")
		b.FailMethod("sendMessage", texts+1, 400, "tail rejected")
		telegramMessage(t, b, 2, map[string]any{"text": "/media", "chat": map[string]any{"id": 9, "type": "private"}})
		requests := telegramRequests(t, c, b, before+2)
		if requests[before].Method != "sendMediaGroup" || requests[before+1].Method != "sendMessage" {
			t.Fatalf("failed tail order: %+v", requests[before:])
		}
		// 后续真实发送完成后核对累计计数，覆盖失败后的进程继续工作与无补发。
		telegramMessage(t, b, 3, map[string]any{"text": "after failure", "chat": map[string]any{"id": 9, "type": "private"}})
		requests = telegramRequests(t, c, b, before+3)
		if b.Calls("sendMediaGroup")-groups != 1 || b.Calls("sendMessage")-texts != 2 || requests[before+2].Params["text"] != "after failure" || b.Calls("editMessageText") != 0 {
			t.Fatalf("request deltas after failure: %+v", requests[before:])
		}
	})
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit=%d\n%s", code, c.log())
	}
	t.Logf("Bot API a=%+v b=%+v", a.Requests(), b.Requests())
}

func telegramEventually(t *testing.T, c *core, predicate func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for !predicate() {
		if time.Now().After(deadline) {
			t.Fatalf("Telegram condition timed out\n%s", c.log())
		}
		time.Sleep(10 * time.Millisecond) // 只等待外部进程的可观察结果。
	}
}

func telegramRequests(t *testing.T, c *core, server *telegramtest.Server, count int) []telegramtest.Request {
	t.Helper()
	telegramEventually(t, c, func() bool { return len(server.Requests()) >= count })
	requests := server.Requests()
	if len(requests) != count {
		t.Fatalf("requests=%+v, want count=%d", requests, count)
	}
	return requests
}

func telegramMessage(t *testing.T, server *telegramtest.Server, id int, message map[string]any) {
	t.Helper()
	message["message_id"], message["date"] = id, 123
	message["from"] = map[string]any{"id": 88, "first_name": "User", "language_code": "en"}
	telegramUpdate(t, server, map[string]any{"update_id": id, "message": message})
}

func telegramUpdate(t *testing.T, server *telegramtest.Server, update map[string]any) {
	t.Helper()
	raw, err := json.Marshal(update)
	if err != nil {
		t.Fatal(err)
	}
	server.Push(raw)
}
