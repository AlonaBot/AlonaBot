//go:build unix

package main

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/alonabot-adapter-telegram/telegramtest"
)

// TestReleaseEndToEnd 从真实发行资产联验首次登录、账号配置、流式对话和停止。
func TestReleaseEndToEnd(t *testing.T) {
	for _, scenario := range []struct{ name, arch string }{{"archive", ""}, {"docker-arm64", "arm64"}, {"docker-amd64", "amd64"}} {
		t.Run(scenario.name, func(t *testing.T) {
			if scenario.arch != "" && os.Getenv("ALONABOT_TEST_DOCKER") != "1" {
				t.Skip("Docker scenario requires ALONABOT_TEST_DOCKER=1")
			}
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
			target := runtime.GOOS + "/" + runtime.GOARCH
			if scenario.arch != "" {
				target = "linux/" + scenario.arch
			}
			optional := releaseOptionalPackages(t, keyFile, target)
			if scenario.arch != "" {
				asset := releaseContainer(t, scenario.arch, keyFile)
				for i, archive := range optional {
					destination := fmt.Sprintf("/optional-%d", i)
					releaseDocker(t, "cp", filepath.Dir(archive), asset.container+":"+destination)
					asset.optional = append(asset.optional, destination+"/"+filepath.Base(archive))
				}
				releaseEndToEnd(t, asset, release.KeyID(public))
				return
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
			for _, id := range []string{"alonabot-webui-admin", "alonabot-agent-official"} {
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
			releaseEndToEnd(t, releaseAsset{core: c, binary: binary, data: data, address: c.address(), optional: optional}, release.KeyID(public))
		})
	}
}

// releaseAsset 保留归档和容器各自的进程入口，业务断言共用实际 HTTP 与文件状态。
type releaseAsset struct {
	core                  *core
	binary, data, address string
	container             string   // 空串表示宿主归档进程。
	optional              []string // 按需安装的独立仓库组件包，路径属于当前宿主或容器。
}

// releaseOptionalPackages 单独构建按需组件，整合归档与镜像保持默认的两组件布局。
func releaseOptionalPackages(t *testing.T, keyFile, target string) []string {
	t.Helper()
	var archives []string
	for _, name := range []string{"help", "telegram"} {
		out := t.TempDir()
		command := exec.CommandContext(t.Context(), filepath.Join(binDir, "alonabot"), "pack", componentDirectories[name], "--target", target, "--key", keyFile, "--out", out)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("pack optional %s: %v\n%s", name, err, output)
		}
		files, err := filepath.Glob(filepath.Join(out, "*.tar.gz"))
		if err != nil || len(files) != 1 {
			t.Fatalf("optional %s archives: %v, %v", name, files, err)
		}
		// Docker 按目录原权限复制，容器内的非 root CLI 需要读取公开的测试资产。
		if err := os.Chmod(out, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{files[0], filepath.Join(out, "release.json"), filepath.Join(out, "release.json.sig"), filepath.Join(out, "release.pub")} {
			if err := os.Chmod(name, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		archives = append(archives, files[0])
	}
	return archives
}

func releaseEndToEnd(t *testing.T, asset releaseAsset, publisher string) {
	t.Helper()
	c := asset.core
	t.Cleanup(func() { t.Logf("Core log:\n%s", c.log()) })
	// 1. 首次预装启动真实 WebUI 后台，并通过文件交付凭证。
	c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"alonabot-webui-admin"`)
	tokenFile := asset.file(t, adminauth.FirstTokenFile)
	info, err := os.Stat(tokenFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("first token permissions: %v", info.Mode())
	}
	output := asset.cli(t, nil, "admin-token", "show", "--data-dir", asset.data)
	token := strings.TrimSpace(string(output))
	if token == "" || strings.Contains(c.log(), token) {
		t.Fatal("missing token or token leaked into logs")
	}
	var document struct {
		BaseURL string `json:"base_url"`
	}
	releaseDecode(t, releaseRead(t, asset.file(t, settings.FileName)), &document)
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
		req, err := http.NewRequestWithContext(t.Context(), method, "http://"+asset.address+path, bytes.NewReader(body))
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
	asset.tokenRemoved(t)
	// 2. 浏览器会话与真实发行 CLI 对照同一初始状态；CLI 转发只改受控 Host。
	cliURL := base.String()
	if asset.container == "" {
		target, err := url.Parse("http://" + asset.address)
		if err != nil {
			t.Fatal(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		director := proxy.Director
		proxy.Director = func(r *http.Request) { director(r); r.Host = base.Host }
		cliURL = asset.serve(t, proxy)
	}
	output = asset.cli(t, []string{"ALONABOT_ADMIN_TOKEN=" + token}, "list", "--admin-url", cliURL)
	var api, listed []manage.Package
	releaseDecode(t, call(http.MethodGet, "/api/v1/packages", nil), &api)
	releaseDecode(t, output, &listed)
	if len(api) != 2 || !reflect.DeepEqual(api, listed) {
		t.Fatalf("package states: API=%+v CLI=%+v", api, listed)
	}
	want := map[string]bool{"alonabot-webui-admin": true, "alonabot-agent-official": false}
	for _, item := range api {
		enabled, exists := want[item.ID]
		if !exists || item.Enabled != enabled || item.Uninstalled || !item.Signed || item.PublisherKeyID != publisher || item.Version != "0.1.0" || item.AvailableUpdate != "" {
			t.Fatalf("preinstalled package: %+v, want publisher=%s", item, publisher)
		}
		if enabled && (item.State != "Activated" || item.Active == "") || !enabled && (item.State != "Installed" || item.Active != "") {
			t.Fatalf("initial activation: %+v", item)
		}
	}
	if strings.Count(c.log(), `"msg":"plugin activated"`) != 1 {
		t.Fatalf("unexpected initial component processes:\n%s", c.log())
	}
	t.Logf("initial packages: %+v; core: %s", api, call(http.MethodGet, "/api/v1/core", nil))
	if raw := call(http.MethodGet, "/", nil); !bytes.Contains(raw, []byte("<html")) {
		t.Fatalf("static WebUI: %s", raw)
	}
	// 账号和帮助功能按独立签名包安装，覆盖整合包之外的正常 CLI 安装链。
	for _, archive := range asset.optional {
		asset.cli(t, []string{"ALONABOT_ADMIN_TOKEN=" + token}, "install", archive, "--admin-url", cliURL)
	}
	call(http.MethodPost, "/api/v1/packages/enable", map[string]string{"id": "alonabot-plugin-help-official"})
	c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"alonabot-plugin-help-official"`)
	// 3. 实例配置保留环境引用，运行副本展开秘密，管理读取只返回设置状态。
	bot := telegramtest.NewServer(t, "synthetic-telegram-token")
	bot.SetMe(101, "release_bot")
	botURL := bot.URL()
	if asset.container != "" {
		target, err := url.Parse(botURL)
		if err != nil {
			t.Fatal(err)
		}
		botURL = asset.serve(t, httputil.NewSingleHostReverseProxy(target))
	}
	call(http.MethodPost, "/api/v1/adapter/instances/create", map[string]any{
		"package_id": "alonabot-adapter-telegram", "instance_id": "tg_e2e", "confirm": false,
		"config": map[string]any{"token": "${T14_TELEGRAM_TOKEN}", "api_base_url": botURL, "proxy_mode": "none"},
	})
	telegramEventually(t, c, func() bool { return bot.Calls("getUpdates") > 0 })
	var disk struct {
		Adapters map[string]struct {
			Instances map[string]struct{ Token string }
		}
		Admins []string
	}
	releaseDecode(t, releaseRead(t, asset.file(t, settings.FileName)), &disk)
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
	providerURL := asset.serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v1/models" {
			if r.Header.Get("Authorization") != "Bearer synthetic-provider-key" {
				http.Error(w, "invalid Provider credential", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := io.WriteString(w, `{"data":[{"id":"local-model"}]}`); err != nil {
				t.Errorf("write Provider models: %v", err)
			}
			return
		}
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
		"providers": []any{map[string]any{"custom": map[string]any{"name": "Local", "protocol": "openai-compatible", "base_url": providerURL + "/v1", "api_key": "synthetic-provider-key"}}},
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
	var discovery struct{ Models []struct{ ID string } }
	releaseDecode(t, call(http.MethodPost, "/plugins/alonabot-agent-official/models/discover", map[string]string{"provider": "Local"}), &discovery)
	if len(discovery.Models) != 1 || discovery.Models[0].ID != "local-model" {
		t.Fatalf("Provider is unreachable from the Agent: %+v", discovery)
	}
	t.Logf("mock reachability: Telegram getUpdates=%d at %s; Provider discovery=%+v at %s", bot.Calls("getUpdates"), botURL, discovery, providerURL)
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
	releaseDecode(t, releaseRead(t, asset.file(t, settings.FileName)), &disk)
	if !reflect.DeepEqual(admins, []string{"telegram:88"}) || !reflect.DeepEqual(disk.Admins, admins) || disk.Adapters["telegram"].Instances["tg_e2e"].Token != "${T14_TELEGRAM_TOKEN}" {
		t.Fatalf("persisted settings: admins=%v disk=%+v", admins, disk)
	}
	// 7. 正常停止后检查所有实际启动的代次进程，包含 reload 退出的旧 Agent。
	asset.stop(t)
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
			if entry.PID <= 0 || asset.container == "" && !pluginExited(entry.PID, 2*time.Second) {
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

func releaseContainer(t *testing.T, arch, keyFile string) releaseAsset {
	t.Helper()
	// 1. 唯一 tag 与卷只属于本次测试；构建密钥经 BuildKit secret 传入。
	name := "alonabot-t14-" + strings.ToLower(rand.Text())
	image := name + ":" + arch
	t.Logf("Docker version: %s; buildx: %s; target=linux/%s", releaseDocker(t, "version", "--format", "{{.Client.Version}} {{.Server.Version}} {{.Server.Os}}/{{.Server.Arch}}"), releaseDocker(t, "buildx", "version"), arch)
	// BuildKit 不按 secret 内容失效缓存，每轮临时密钥都需要重新执行签名构建。
	build := exec.CommandContext(t.Context(), "docker", "buildx", "build", "--progress=plain", "--no-cache-filter", "build", "--platform", "linux/"+arch, "--build-arg", "VERSION=0.1.0", "--secret", "id=release_key,src="+keyFile, "--load", "-t", image, ".")
	build.Dir = "../.."
	t.Logf("building Docker image %s for linux/%s", image, arch)
	output, err := build.CombinedOutput()
	t.Logf("Docker build:\n%s", output)
	if err != nil {
		t.Fatalf("build Docker image: %v", err)
	}
	t.Cleanup(func() { releaseDockerCleanup(t, "image", "rm", image) })
	platform := strings.TrimSpace(string(releaseDocker(t, "image", "inspect", "--format", "{{.Os}}/{{.Architecture}} user={{.Config.User}}", image)))
	if platform != "linux/"+arch+" user=10001:10001" {
		t.Fatalf("image platform or user: %s", platform)
	}
	releaseDocker(t, "volume", "create", name)
	t.Cleanup(func() { releaseDockerCleanup(t, "volume", "rm", name) })
	// 2. Docker 分配空闲宿主端口，发布限 loopback；模拟服务走宿主网关。
	args := []string{"run", "--name", name, "--init", "--platform", "linux/" + arch, "-d", "-p", "127.0.0.1::7650", "--mount", "source=" + name + ",target=/data", "-e", "T14_TELEGRAM_TOKEN=synthetic-telegram-token"}
	if runtime.GOOS == "linux" {
		args = append(args, "--add-host", "host.docker.internal:host-gateway")
	}
	args = append(args, image, "run", "--data-dir", "/data", "--log-format", "json")
	container := strings.TrimSpace(string(releaseDocker(t, args...)))
	var c *core
	t.Cleanup(func() {
		releaseDockerCleanup(t, "rm", "--force", container)
		if c != nil {
			<-c.done // 容器移除后日志跟随结束，等待回收本次 Docker CLI。
		}
	})
	var ports map[string][]struct{ HostIP, HostPort string }
	releaseDecode(t, releaseDocker(t, "inspect", "--format", "{{json .NetworkSettings.Ports}}", container), &ports)
	bindings := ports["7650/tcp"]
	if len(bindings) != 1 || bindings[0].HostIP != "127.0.0.1" || bindings[0].HostPort == "" {
		t.Fatalf("container port bindings: %+v", ports)
	}
	if state := strings.TrimSpace(string(releaseDocker(t, "inspect", "--format", "{{.HostConfig.Init}} {{.Config.Tty}}", container))); state != "true false" {
		t.Fatalf("container init and TTY: %s", state)
	}
	t.Logf("Docker candidate: image=%s platform=%s container=%s volume=%s ports=%+v", image, platform, container, name, ports)
	// 3. 复用 Core 的日志等待与 Telegram 观察，日志流随容器实际退出结束。
	logs := &core{t: t, cmd: exec.Command("docker", "logs", "--follow", container), added: make(chan struct{}, 1), done: make(chan struct{})}
	stderr, err := logs.cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	logs.cmd.Stdout = logs.cmd.Stderr
	if err := logs.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	c = logs
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			c.mu.Lock()
			c.lines = append(c.lines, scanner.Text())
			c.mu.Unlock()
			select { // 容量 1 的通知合并日志批次，等待者每次读取完整日志。
			case c.added <- struct{}{}:
			default:
			}
		}
		if err := errors.Join(scanner.Err(), c.cmd.Wait()); err != nil {
			t.Errorf("collect Docker logs: %v", err)
		}
		close(c.done) // Wait 完成后再发布退出，供停止与清理读取。
	}()
	return releaseAsset{core: c, container: container, data: "/data", address: net.JoinHostPort(bindings[0].HostIP, bindings[0].HostPort)}
}

func (a releaseAsset) cli(t *testing.T, env []string, args ...string) []byte {
	t.Helper()
	var command *exec.Cmd
	if a.container == "" {
		command = exec.CommandContext(t.Context(), a.binary, args...)
		command.Env = append(os.Environ(), env...)
	} else {
		dockerArgs := []string{"exec"}
		for _, value := range env {
			dockerArgs = append(dockerArgs, "-e", value)
		}
		dockerArgs = append(dockerArgs, a.container, "/alonabot")
		command = exec.CommandContext(t.Context(), "docker", append(dockerArgs, args...)...)
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("release CLI %v: %v\n%s", args, err, output)
	}
	return output
}

// file 把容器卷中的指定文件按原权限复制到临时目录，归档直接读取真实文件。
func (a releaseAsset) file(t *testing.T, name string) string {
	t.Helper()
	if a.container == "" {
		return filepath.Join(a.data, name)
	}
	local := filepath.Join(t.TempDir(), name)
	releaseDocker(t, "cp", a.container+":"+a.data+"/"+name, local)
	return local
}

func (a releaseAsset) tokenRemoved(t *testing.T) {
	t.Helper()
	if a.container == "" {
		if _, err := os.Stat(filepath.Join(a.data, adminauth.FirstTokenFile)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("token file after login: %v", err)
		}
		return
	}
	output, err := exec.CommandContext(t.Context(), "docker", "cp", a.container+":/data/admin_token", filepath.Join(t.TempDir(), "deleted_token")).CombinedOutput()
	if err == nil || !strings.Contains(string(output), "Could not find the file /data/admin_token in container") {
		t.Fatalf("token file after login: %v\n%s", err, output)
	}
	t.Logf("deleted container token: %v\n%s", err, output)
}

// serve 仅在容器场景监听宿主可达接口，返回候选进程应使用的 URL。
func (a releaseAsset) serve(t *testing.T, handler http.Handler) string {
	t.Helper()
	server := httptest.NewUnstartedServer(handler)
	if a.container != "" {
		if err := server.Listener.Close(); err != nil {
			t.Fatal(err)
		}
		listener, err := net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			t.Fatal(err)
		}
		server.Listener = listener
	}
	server.Start()
	t.Cleanup(server.Close)
	if a.container != "" {
		return "http://" + net.JoinHostPort("host.docker.internal", fmt.Sprint(server.Listener.Addr().(*net.TCPAddr).Port))
	}
	return server.URL
}

func (a releaseAsset) stop(t *testing.T) {
	t.Helper()
	if a.container == "" {
		stopTestCore(t, a.core)
		return
	}
	t.Logf("container processes before stop:\n%s", releaseDocker(t, "top", a.container, "-eo", "pid,ppid,user,args"))
	releaseDocker(t, "stop", a.container)
	raw := releaseDocker(t, "inspect", "--format", "{{json .State}}", a.container)
	t.Logf("stopped container: %s", raw)
	var state struct {
		Running       bool
		PID, ExitCode int
	}
	releaseDecode(t, raw, &state)
	if state.Running || state.PID != 0 || state.ExitCode != 0 {
		t.Fatalf("container did not stop cleanly: %s", raw)
	}
	output, err := exec.CommandContext(t.Context(), "docker", "top", a.container).CombinedOutput()
	if err == nil || !strings.Contains(string(output), "is not running") {
		t.Fatalf("container still has processes: %v\n%s", err, output)
	}
	t.Logf("no live container processes: %v\n%s", err, output)
	if code := a.core.wait(15 * time.Second); code != 0 {
		t.Fatalf("Docker log collector exit=%d\n%s", code, a.core.log())
	}
}

func releaseDocker(t *testing.T, args ...string) []byte {
	t.Helper()
	output, err := exec.CommandContext(t.Context(), "docker", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %v: %v\n%s", args, err, output)
	}
	return output
}

func releaseDockerCleanup(t *testing.T, args ...string) {
	t.Helper()
	// 测试 Context 在 Cleanup 前取消；资源删除必须独立完成。
	if output, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		t.Errorf("clean Docker resource %v: %v\n%s", args, err, output)
	}
}
