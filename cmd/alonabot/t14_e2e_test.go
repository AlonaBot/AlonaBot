//go:build unix

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/adapters/telegram/telegramtest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// TestReleaseEndToEnd 从真实发行资产联验首次登录、账号配置、流式对话和停止。
func TestReleaseEndToEnd(t *testing.T) {
	for _, name := range []string{"archive"} {
		t.Run(name, func(t *testing.T) {
			// 1. 正式构建入口使用隔离的临时密钥，运行时只消费解压后的资产。
			public, private, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			keyFile := filepath.Join(dir, "signing.key")
			if err := os.WriteFile(keyFile, []byte(base64.StdEncoding.EncodeToString(private)), 0o600); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, "dist")
			build := exec.CommandContext(t.Context(), "go", "run", "./scripts/dist", "--version", "0.1.0", "--key", keyFile, "--out", out, "--platforms", runtime.GOOS+"/"+runtime.GOARCH)
			build.Dir = "../.."
			output, err := build.CombinedOutput()
			t.Logf("distribution build:\n%s", output)
			if err != nil {
				t.Fatalf("build distribution: %v", err)
			}
			root := fmt.Sprintf("alonabot-0.1.0-%s-%s", runtime.GOOS, runtime.GOARCH)
			extracted := filepath.Join(dir, "extracted")
			if err := packages.Extract(filepath.Join(out, root+".tar.gz"), extracted, packages.NewLimits(tune.Defaults())); err != nil {
				t.Fatal(err)
			}
			asset := filepath.Join(extracted, root)
			t.Logf("release assets: %s", asset)
			for _, id := range []string{"alonabot-plugin-help-official", "alonabot-webui-admin", "alonabot-agent-official", "alonabot-adapter-telegram"} {
				metadata := releaseRead(t, filepath.Join(asset, "packages", id, "release.json"))
				signature := releaseRead(t, filepath.Join(asset, "packages", id, "release.json.sig"))
				if _, err := release.Verify(metadata, signature, public); err != nil {
					t.Fatalf("verify %s: %v", id, err)
				}
			}
			// 2. 独立目录与非终端 stdin 保持发行场景，环境只提供合成的账号凭证。
			t.Chdir(asset)
			t.Setenv("T14_TELEGRAM_TOKEN", "synthetic-telegram-token")
			binary, data := filepath.Join(asset, "alonabot"), filepath.Join(dir, "data")
			c := startRecoveryCore(t, binary, data, "")
			releaseEndToEnd(t, c, binary, data, release.KeyID(public))
		})
	}
}

func releaseEndToEnd(t *testing.T, c *core, binary, data, publisher string) {
	t.Helper()
	t.Cleanup(func() { t.Logf("Core log:\n%s", c.log()) })
	// 1. 首次预装必须启动真实 help 与 WebUI 后台，并通过文件交付凭证。
	for _, id := range []string{"alonabot-plugin-help-official", "alonabot-webui-admin"} {
		c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"` + id + `"`)
	}
	tokenFile := filepath.Join(data, adminauth.FirstTokenFile)
	info, err := os.Stat(tokenFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("first token permissions: %v", info.Mode())
	}
	output, err := exec.CommandContext(t.Context(), binary, "admin-token", "show", "--data-dir", data).CombinedOutput()
	if err != nil {
		t.Fatalf("show token: %v\n%s", err, output)
	}
	token := strings.TrimSpace(string(output))
	if token == "" || strings.Contains(c.log(), token) {
		t.Fatal("missing token or token leaked into logs")
	}
	var document struct {
		BaseURL string `json:"base_url"`
	}
	releaseDecode(t, releaseRead(t, filepath.Join(data, settings.FileName)), &document)
	base, err := url.Parse(document.BaseURL)
	if err != nil {
		t.Fatal(err)
	}
	var cookies []*http.Cookie
	var csrf string
	call := func(method, path string, input any) []byte {
		t.Helper()
		var body []byte
		if input != nil {
			var err error
			body, err = json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequestWithContext(t.Context(), method, "http://"+c.address()+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Host = base.Host
		req.Header.Set("Origin", base.String())
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Alonabot-CSRF", csrf)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		raw, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s %s: %d %s\n%s", method, path, response.StatusCode, raw, c.log())
		}
		if path == "/api/auth/login" {
			cookies = response.Cookies()
		}
		return raw
	}
	var login struct {
		CSRFToken string `json:"csrf_token"`
	}
	releaseDecode(t, call(http.MethodPost, "/api/auth/login", map[string]string{"token": token}), &login)
	csrf = login.CSRFToken
	if len(cookies) != 1 || csrf == "" {
		t.Fatal("login did not deliver a session and CSRF token")
	}
	if _, err := os.Stat(tokenFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("token file after login: %v", err)
	}
	// 2. 浏览器会话与真实发行 CLI 对照同一初始状态；CLI 转发只改受控 Host。
	target, err := url.Parse("http://" + c.address())
	if err != nil {
		t.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(r *http.Request) { director(r); r.Host = base.Host }
	server := httptest.NewServer(proxy)
	defer server.Close()
	command := exec.CommandContext(t.Context(), binary, "list", "--admin-url", server.URL)
	command.Env = append(os.Environ(), "ALONABOT_ADMIN_TOKEN="+token)
	output, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI list: %v\n%s", err, output)
	}
	var api, listed []manage.Package
	releaseDecode(t, call(http.MethodGet, "/api/v1/packages", nil), &api)
	releaseDecode(t, output, &listed)
	if len(api) != 4 || !reflect.DeepEqual(api, listed) {
		t.Fatalf("package states: API=%+v CLI=%+v", api, listed)
	}
	want := map[string]bool{"alonabot-plugin-help-official": true, "alonabot-webui-admin": true, "alonabot-agent-official": false, "alonabot-adapter-telegram": false}
	for _, item := range api {
		enabled, exists := want[item.ID]
		if !exists || item.Enabled != enabled || item.Uninstalled || !item.Signed || item.PublisherKeyID != publisher || item.Version != "0.1.0" || item.AvailableUpdate != "" {
			t.Fatalf("preinstalled package: %+v", item)
		}
		if enabled && (item.State != "Activated" || item.Active == "") || !enabled && (item.State != "Installed" || item.Active != "") {
			t.Fatalf("initial activation: %+v", item)
		}
	}
	if strings.Count(c.log(), `"msg":"plugin activated"`) != 2 {
		t.Fatalf("unexpected initial component processes:\n%s", c.log())
	}
	t.Logf("initial packages: %+v; core: %s", api, call(http.MethodGet, "/api/v1/core", nil))
	if raw := call(http.MethodGet, "/", nil); !bytes.Contains(raw, []byte("<html")) {
		t.Fatalf("static WebUI: %s", raw)
	}
	// 3. 实例配置保留环境引用，运行副本展开秘密，管理读取只返回设置状态。
	bot := telegramtest.NewServer(t, "synthetic-telegram-token")
	bot.SetMe(101, "release_bot")
	call(http.MethodPost, "/api/v1/adapter/instances/create", map[string]any{
		"package_id": "alonabot-adapter-telegram", "instance_id": "tg_e2e", "confirm": false,
		"config": map[string]any{"token": "${T14_TELEGRAM_TOKEN}", "api_base_url": bot.URL(), "proxy_mode": "none"},
	})
	telegramEventually(t, c, func() bool { return bot.Calls("getUpdates") > 0 })
	var disk struct {
		Adapters map[string]struct {
			Instances map[string]struct{ Token string }
		}
		Admins []string
	}
	releaseDecode(t, releaseRead(t, filepath.Join(data, settings.FileName)), &disk)
	if disk.Adapters["telegram"].Instances["tg_e2e"].Token != "${T14_TELEGRAM_TOKEN}" {
		t.Fatal("instance token environment reference was not preserved")
	}
	var instance struct {
		Config struct {
			Token map[string]bool
		}
	}
	raw := call(http.MethodGet, "/api/v1/adapters/telegram/instances/tg_e2e", nil)
	releaseDecode(t, raw, &instance)
	if !instance.Config.Token["$secret_set"] || bytes.Contains(raw, []byte("synthetic-telegram-token")) || bytes.Contains(raw, []byte("T14_TELEGRAM_TOKEN")) {
		t.Fatalf("instance secret disclosure: %s", raw)
	}
	releaseDecode(t, call(http.MethodGet, "/api/v1/packages", nil), &api)
	for _, item := range api {
		if item.ID == "alonabot-adapter-telegram" && (!item.Enabled || item.State != "Activated") {
			t.Fatalf("adapter was not automatically enabled: %+v", item)
		}
	}
	// 4. Provider 的第二个增量等待首条 Telegram 发送，避免瞬时完成掩盖编辑。
	type providerRequest struct {
		Method, Path, Authorization string
		Body                        []byte
		Err                         error
	}
	requests := make(chan providerRequest, 1)
	continueStream := make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		// 主测试持有断言；HTTP goroutine 仅交付完整请求与读取错误。
		select {
		case requests <- providerRequest{r.Method, r.URL.Path, r.Header.Get("Authorization"), body, err}:
		case <-t.Context().Done():
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if _, err := io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n"); err != nil {
			t.Errorf("write first Provider delta: %v", err)
			return
		}
		w.(http.Flusher).Flush()
		// 请求取消也解除等待，让失败路径可以正常回收模拟服务。
		select {
		case <-continueStream:
		case <-r.Context().Done():
			return
		case <-t.Context().Done():
			return
		}
		if _, err := io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"); err != nil {
			t.Errorf("write final Provider delta: %v", err)
		}
	}))
	t.Cleanup(provider.Close)
	call(http.MethodPost, "/api/v1/packages/enable", map[string]string{"id": "alonabot-agent-official"})
	const configPath = "/api/v1/plugins/alonabot-agent-official/config"
	var before, after struct {
		ActiveGeneration string          `json:"active_generation"`
		Active           json.RawMessage `json:"active"`
		Pending          json.RawMessage `json:"pending"`
	}
	releaseDecode(t, call(http.MethodGet, configPath, nil), &before)
	if before.ActiveGeneration == "" {
		t.Fatal("empty Agent configuration did not activate")
	}
	configuration := map[string]any{
		"providers": []any{map[string]any{"custom": map[string]any{"name": "Local", "protocol": "openai-compatible", "base_url": provider.URL + "/v1", "api_key": "synthetic-provider-key"}}},
		"models":    []any{map[string]any{"name": "Chat", "provider": "Local", "model": "local-model", "type": "chat", "max_tokens": 128}}, "chat_model": "Chat",
	}
	call(http.MethodPut, configPath+"/pending", map[string]any{"base_generation": before.ActiveGeneration, "config": configuration})
	call(http.MethodPost, "/api/v1/plugins/alonabot-agent-official/reload", nil)
	releaseDecode(t, call(http.MethodGet, configPath, nil), &after)
	var active struct {
		ChatModel string `json:"chat_model"`
		Providers []struct {
			Custom struct {
				APIKey map[string]bool `json:"api_key"`
			}
		}
	}
	releaseDecode(t, after.Active, &active)
	if after.ActiveGeneration == "" || after.ActiveGeneration == before.ActiveGeneration || string(after.Pending) != "null" || active.ChatModel != "Chat" || len(active.Providers) != 1 || !active.Providers[0].Custom.APIKey["$secret_set"] || bytes.Contains(after.Active, []byte("synthetic-provider-key")) {
		t.Fatalf("Agent reload did not activate the redacted Pending: %+v", after)
	}
	// 5. 通过真实命令和 Provider 请求检查新代次，再核对同一消息的发送与编辑。
	telegramMessage(t, bot, 1, map[string]any{"text": "/ask hello", "chat": map[string]any{"id": 88, "type": "private"}})
	sent := telegramRequests(t, c, bot, 1)
	if sent[0].Method != "sendMessage" || sent[0].Params["text"] != "hel" || sent[0].Params["chat_id"] != float64(88) {
		t.Fatalf("first stream send: %+v", sent)
	}
	var request providerRequest
	select {
	case request = <-requests:
	default:
		t.Fatal("Telegram replied without a Provider request")
	}
	if request.Err != nil {
		t.Fatal(request.Err)
	}
	var completion struct {
		Model    string
		Stream   bool
		Messages []struct{ Role, Content string }
	}
	releaseDecode(t, request.Body, &completion)
	if request.Method != http.MethodPost || request.Path != "/v1/chat/completions" || request.Authorization != "Bearer synthetic-provider-key" || completion.Model != "local-model" || !completion.Stream || len(completion.Messages) == 0 {
		t.Fatalf("Provider request: method=%s path=%s body=%s", request.Method, request.Path, request.Body)
	}
	user := completion.Messages[len(completion.Messages)-1]
	if user.Role != "user" || user.Content != "hello" {
		t.Fatalf("Provider user message: %+v", user)
	}
	close(continueStream) // 只有确认首个发送成功后，才允许 Provider 完成流。
	sent = telegramRequests(t, c, bot, 2)
	// 模拟 Bot API 从 1 分配发送 ID，首次发送的响应固定为 1。
	if sent[1].Method != "editMessageText" || sent[1].Params["message_id"] != float64(1) || sent[1].Params["chat_id"] != sent[0].Params["chat_id"] || sent[1].Params["text"] != "hello" {
		t.Fatalf("final stream edit: %+v", sent)
	}
	t.Logf("Provider request: %s; stream requests: %+v", request.Body, sent)
	// 6. 每条命令从发送前的位置取回复，核对命令发现、身份与管理员持久化。
	for i, command := range []string{"/help", "/whoami"} {
		position := len(bot.Requests())
		telegramMessage(t, bot, i+2, map[string]any{"text": command, "chat": map[string]any{"id": 88, "type": "private"}})
		reply := telegramRequests(t, c, bot, position+1)[position]
		text, ok := reply.Params["text"].(string)
		if !ok || reply.Method != "sendMessage" || reply.Params["chat_id"] != float64(88) {
			t.Fatalf("%s reply: %+v", command, reply)
		}
		if command == "/help" && !strings.Contains(text, "ask") || command == "/whoami" && (!strings.Contains(text, "tg_e2e") || !strings.Contains(text, "88")) {
			t.Fatalf("%s reply: %s", command, text)
		}
		t.Logf("%s: %s", command, text)
	}
	call(http.MethodPost, "/api/v1/admins/add", map[string]string{"user": "telegram:88"})
	var admins []string
	releaseDecode(t, call(http.MethodGet, "/api/v1/core/admins", nil), &admins)
	releaseDecode(t, releaseRead(t, filepath.Join(data, settings.FileName)), &disk)
	if !reflect.DeepEqual(admins, []string{"telegram:88"}) || !reflect.DeepEqual(disk.Admins, admins) || disk.Adapters["telegram"].Instances["tg_e2e"].Token != "${T14_TELEGRAM_TOKEN}" {
		t.Fatalf("persisted settings: admins=%v disk=%+v", admins, disk)
	}
	// 7. 正常停止后检查所有实际启动的代次进程，包含 reload 退出的旧 Agent。
	stopTestCore(t, c)
	activated := make(map[string]bool)
	for _, line := range strings.Split(c.log(), "\n") {
		var entry struct {
			Message  string `json:"msg"`
			PluginID string `json:"plugin_id"`
			PID      int    `json:"pid"`
		}
		releaseDecode(t, []byte(line), &entry)
		if entry.Message == "plugin activated" {
			activated[entry.PluginID] = true
			if entry.PID <= 0 || !pluginExited(entry.PID, 2*time.Second) {
				t.Fatalf("component process did not exit: %+v", entry)
			}
		}
	}
	if len(activated) != 4 || !strings.Contains(c.log(), `"msg":"Core stopped"`) {
		t.Fatalf("incomplete component lifecycle: %v\n%s", activated, c.log())
	}
	if strings.Contains(c.log(), token) || strings.Contains(c.log(), "synthetic-provider-key") || strings.Contains(c.log(), "synthetic-telegram-token") {
		t.Fatal("credentials leaked into Core logs")
	}
}

func releaseRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func releaseDecode(t *testing.T, raw []byte, value any) {
	t.Helper()
	if err := json.Unmarshal(raw, value); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, raw)
	}
}
