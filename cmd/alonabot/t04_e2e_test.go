//go:build unix

package main

import (
	"fmt"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/adapters/telegram/telegramtest"
)

// TestSDKExamples 通过真实 CLI、Core、SDK 进程与本地 Bot API 验证示例的新增命令。
func TestSDKExamples(t *testing.T) {
	server := telegramtest.NewServer(t, "sdk-examples-token")
	server.SetMe(101, "examples_bot")
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{
		"listen":"127.0.0.1:0","adapters":{"telegram":{
		"package":"alonabot-adapter-telegram","instances":{"examples":{
		"token":"sdk-examples-token","api_base_url":%q,"proxy_mode":"none"}}}}}`, server.URL()))
	c := startEmptyCore(t, dir, "../../adapters/telegram", "../../examples/plugins/echo", "../../examples/plugins/media")
	for _, id := range []string{"alonabot-adapter-telegram", "echo", "media"} {
		developmentPID(t, c, id)
	}
	telegramEventually(t, c, func() bool { return server.Calls("getUpdates") > 0 })
	defer func() { t.Logf("Core log:\n%s", c.log()) }()
	for i, scenario := range []struct{ name, input, method, output string }{
		{"explicit", "/echo hello world", "sendMessage", "hello world"},
		{"optional", "/echo", "sendMessage", "echo"},
		{"method_raw", "/note add first  note", "sendMessage", "Note added."},
		{"alias_raw", "/n add second note", "sendMessage", "Note added."},
		{"group", "/note list", "sendMessage", "first  note\nsecond note"},
		{"alias", "/n list", "sendMessage", "first  note\nsecond note"},
		{"default", "/calc 4", "sendMessage", "5"},
		{"explicit_default", "/calc 4 3", "sendMessage", "7"},
		{"captures", "roll 2d6", "sendMessage", "Rolling 2d6."},
		{"keyword", "ECHO EXAMPLE KEYWORD", "sendMessage", "Keyword matched."},
		{"raw", "echo example raw", "sendMessage", "Raw matched."},
		{"unconsumed", "ordinary text", "sendMessage", "ordinary text"},
		{"struct_subplugin", "/caption a  custom caption", "sendPhoto", "a  custom caption"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			telegramMessage(t, server, i+1, map[string]any{"text": scenario.input, "chat": map[string]any{"id": 9, "type": "private"}})
			requests := telegramRequests(t, c, server, i+1)
			request := requests[i]
			field := "text"
			if scenario.method == "sendPhoto" {
				field = "caption"
			}
			if request.Method != scenario.method || request.Params[field] != scenario.output || fmt.Sprint(request.Params["chat_id"]) != "9" {
				t.Fatalf("input=%q method=%s params=%+v, want %s %q", scenario.input, request.Method, request.Params, scenario.method, scenario.output)
			}
		})
	}
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit=%d\n%s", code, c.log())
	}
}
