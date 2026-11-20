//go:build unix

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/AlonaBot/AlonaBot/adapters/telegram/telegramtest"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestCrossComponentEndToEnd 通过正式包和管理入口联验真实组件进程。
func TestCrossComponentEndToEnd(t *testing.T) {
	for _, transport := range []string{"tcp", "stdio"} {
		t.Run(transport, func(t *testing.T) {
			t.Run("components", func(t *testing.T) { testCrossComponentPackages(t, transport) })
			t.Run("capabilities", func(t *testing.T) { testCrossCapabilityPackages(t, transport) })
			t.Run("boundaries", func(t *testing.T) { testCrossPermissionPackages(t, transport) })
		})
	}
}

func testCrossCapabilityPackages(t *testing.T, transport string) {
	projects := []struct{ directory, id string }{
		{"../../adapters/telegram", "alonabot-adapter-telegram"},
		{"../../examples/plugins/lifecycle", "demo"},
		{"../../examples/plugins/echo", "echo"},
		{"../../examples/plugins/media", "media"},
	}
	archives := make([]string, len(projects))
	for i, project := range projects {
		archives[i] = packAgentComponent(t, project.directory, project.id)
	}
	server := telegramtest.NewServer(t, "capability-token")
	server.SetMe(101, "capability_bot")
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{"adapters":{"telegram":{"package":"alonabot-adapter-telegram","default_instance":"capability_bot","instances":{"capability_bot":{"token":"capability-token","api_base_url":%q,"proxy_mode":"none"}}}}}`, server.URL()))
	args := []string{"--listen", "127.0.0.1:0", "--log-format", "json"}
	if transport == "stdio" {
		args = append(args, "--stdio")
	}
	c := startEmptyCore(t, dir, args...)
	for i, project := range projects {
		c.management(t, dir, "", "install", archives[i], "--yes")
		c.management(t, dir, "", "enable", project.id)
	}
	telegramEventually(t, c, func() bool { return server.Calls("getUpdates") > 0 })
	// 无配置、无数据库的命令插件与文本匹配插件分别产生真实回复。
	for i, item := range []struct{ input, want string }{{"/ping", "pong"}, {"roll 3d8", "Rolling 3d8."}} {
		before := len(server.Requests())
		telegramMessage(t, server, i+1, map[string]any{"text": item.input, "chat": map[string]any{"id": 9, "type": "private"}})
		telegramEventually(t, c, func() bool {
			for _, request := range server.Requests()[before:] {
				if request.Method == "sendMessage" && request.Params["text"] == item.want {
					return true
				}
			}
			return false
		})
	}
	// 媒体由另一个实际上传并写库的插件实现，核对 Adapter 收到的完整字节。
	telegramMessage(t, server, 3, map[string]any{"text": "/media", "chat": map[string]any{"id": 9, "type": "private"}})
	sample := developmentRead(t, "../../examples/plugins/media/assets/sample.png")
	telegramEventually(t, c, func() bool { return server.Calls("sendMediaGroup") == 1 })
	for _, request := range server.Requests() {
		if request.Method == "sendMediaGroup" {
			if len(request.Files) != 2 || !reflect.DeepEqual(request.Files["media_0"], sample) || !reflect.DeepEqual(request.Files["media_1"], sample) {
				t.Fatal("media plugin upload differs from its packaged asset")
			}
		}
	}
	telegramEventually(t, c, func() bool {
		for _, request := range server.Requests() {
			if request.Method == "sendMessage" && request.Params["text"] == "Click to edit this text." {
				return true
			}
		}
		return false
	})
	stopTestCore(t, c)
	for _, tag := range []string{"sqlite", "configdb_gorm"} {
		t.Run(tag, func(t *testing.T) { testCrossDatabasePackages(t, transport, tag) })
	}
}

func testCrossDatabasePackages(t *testing.T, transport, tag string) {
	out := t.TempDir()
	arguments := []string{"pack", "../../examples/plugins/configdb", "--asset-base", "https://example.invalid", "--out", out}
	if tag == "configdb_gorm" {
		arguments = append(arguments, "--tags", tag)
	}
	if output, err := exec.CommandContext(t.Context(), filepath.Join(binDir, "alonabot"), arguments...).CombinedOutput(); err != nil {
		t.Fatalf("pack %s: %v\n%s", tag, err, output)
	}
	const id = "alonabot-plugin-configdb"
	archive := filepath.Join(out, id+"-0.1.0-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz")
	adapter := packAgentComponent(t, "../../adapters/telegram", "alonabot-adapter-telegram")
	server := telegramtest.NewServer(t, "database-token")
	server.SetMe(101, "database_bot")
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{"adapters":{"telegram":{"package":"alonabot-adapter-telegram","default_instance":"database_bot","instances":{"database_bot":{"token":"database-token","api_base_url":%q,"proxy_mode":"none"}}}}}`, server.URL()))
	configFile := filepath.Join(dir, "plugins", id, "config.json")
	developmentWrite(t, configFile, `{"greeting":["First","value"],"count":"invalid","enabled":false,"unknown":41}`)
	args := []string{"--listen", "127.0.0.1:0", "--log-format", "json"}
	if transport == "stdio" {
		args = append(args, "--stdio")
	}
	c := startEmptyCore(t, dir, args...)
	for _, item := range []struct{ archive, id string }{{adapter, "alonabot-adapter-telegram"}, {archive, id}} {
		c.management(t, dir, "", "install", item.archive, "--yes")
		c.management(t, dir, "", "enable", item.id)
	}
	first := findPackage(t, c, dir, id).Active
	status, raw := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/plugins/"+id+"/config")
	var snapshot struct {
		Active      map[string]any `json:"active"`
		Unknown     []string       `json:"unknown"`
		Corrections []any          `json:"corrections"`
	}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK || snapshot.Active["greeting"] != "First value" || snapshot.Active["count"] != float64(7) || snapshot.Active["enabled"] != false || len(snapshot.Corrections) == 0 || !reflect.DeepEqual(snapshot.Unknown, []string{`$["unknown"]`}) {
		t.Fatalf("normalized configuration: %d %s", status, raw)
	}
	telegramEventually(t, c, func() bool { return server.Calls("getUpdates") > 0 })
	reply := func(sequence int, text, want string) {
		t.Helper()
		before := len(server.Requests())
		telegramMessage(t, server, sequence, map[string]any{"text": text, "chat": map[string]any{"id": 9, "type": "private"}})
		telegramEventually(t, c, func() bool {
			for _, request := range server.Requests()[before:] {
				if request.Method == "sendMessage" && request.Params["text"] == want {
					return true
				}
			}
			return false
		})
	}
	reply(1, "/note set durable persisted", "Saved durable")
	// 候选先写自己的启动表再失败；原代次与真实笔记继续可用。
	developmentWrite(t, configFile, `{"greeting":"Candidate only","fail_start":true}`)
	output, err := recoveryCommand(t, c, dir, "reload", id).CombinedOutput()
	if err == nil || !strings.Contains(c.log(), "requested failure after startup database write") {
		t.Fatalf("failed candidate: %v %s\n%s", err, output, c.log())
	}
	if current := findPackage(t, c, dir, id); current.Active != first || current.State != "Activated" {
		t.Fatalf("failed reload changed Active: %+v", current)
	}
	reply(2, "/note get durable", "durable: persisted")
	developmentWrite(t, configFile, `{"greeting":"","count":0,"enabled":false}`)
	c.management(t, dir, "", "reload", id)
	last := findPackage(t, c, dir, id).Active
	if last == first || last == "" {
		t.Fatal("successful reload did not create a generation")
	}
	status, raw = fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/plugins/"+id+"/config")
	var reloaded struct {
		Active  map[string]any `json:"active"`
		Pending any            `json:"pending"`
	}
	if err := json.Unmarshal(raw, &reloaded); err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK || reloaded.Active["greeting"] != "" || reloaded.Active["count"] != float64(0) || reloaded.Active["enabled"] != false || reloaded.Pending != nil {
		t.Fatalf("explicit empty configuration after reload: %d %s", status, raw)
	}
	reply(3, "/note get durable", "durable: persisted")
	stopTestCore(t, c)
	// Core 退出后读取权威状态库及两个真实业务库，避免日志代替数据证据。
	store := recoveryStore(t, dir)
	active, err := store.Active(t.Context(), id)
	if err != nil || active != last {
		t.Fatalf("persisted Active: %s %v", active, err)
	}
	records, err := store.Journals(t.Context(), id)
	if err != nil || len(records) != 3 {
		t.Fatalf("database journals: %+v %v", records, err)
	}
	for _, record := range records {
		wantPhase := "committed"
		if record.NewGeneration != first && record.NewGeneration != last {
			wantPhase = "cleaned"
			candidate := filepath.Join(dir, "plugins", id, "generations", record.NewGeneration)
			if _, err := os.Stat(candidate); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("failed candidate directory retained: %s %v", candidate, err)
			}
		}
		if record.Phase != wantPhase {
			t.Fatalf("database journal: %+v", record)
		}
	}
	var previous os.FileInfo
	for _, item := range []struct{ generation, startup string }{{first, "First value"}, {last, ""}} {
		file := filepath.Join(dir, "plugins", id, "generations", item.generation, "db.sqlite")
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}
		if previous != nil && os.SameFile(previous, info) {
			t.Fatal("reload reused the old database file")
		}
		previous = info
		db, err := sql.Open("sqlite", file)
		if err != nil {
			t.Fatal(err)
		}
		var startup, note string
		var migrations int
		errs := []error{
			db.QueryRow("SELECT value FROM configdb_startup WHERE key='startup'").Scan(&startup),
			db.QueryRow("SELECT value FROM configdb_note WHERE key='durable'").Scan(&note),
			db.QueryRow("SELECT count(*) FROM _alonabot_migrations").Scan(&migrations),
			db.Close(),
		}
		if err := errors.Join(errs...); err != nil || startup != item.startup || note != "persisted" || migrations != 2 {
			t.Fatalf("%s %s database: startup=%q note=%q migrations=%d err=%v", tag, item.generation, startup, note, migrations, err)
		}
	}
	t.Logf("%s/%s Active=%s journals=%+v", transport, tag, active, records)
}

func testCrossComponentPackages(t *testing.T, transport string) {
	// 1. pack 读取真实工程，安装时只消费正式产物。
	projects := []struct{ directory, id string }{
		{"echo", "echo"},
		{"configdb", "alonabot-plugin-configdb"},
		{"showcase", "showcase"},
		{"../../adapters/telegram", "alonabot-adapter-telegram"},
		{"../../plugins/help", "alonabot-plugin-help-official"},
		{"../../agent", "alonabot-agent-official"},
		{"webui", "alonabot-webui-admin"},
	}
	archives := make([]string, len(projects))
	for i, project := range projects {
		if project.directory == "webui" {
			archives[i] = packCrossComponentWebUI(t)
			continue
		}
		directory := filepath.Join("../../examples/plugins", project.directory)
		if strings.HasPrefix(project.directory, "../../") {
			directory = project.directory
		}
		archives[i] = packAgentComponent(t, directory, project.id)
	}
	server := telegramtest.NewServer(t, "cross-component-token")
	server.SetMe(101, "cross_component_bot")
	var providerMu sync.Mutex
	var providerRequests []agentPackageRequest
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request agentPackageRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		providerMu.Lock()
		providerRequests = append(providerRequests, request)
		providerMu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		writeAgentPackageText(t, w, "cross component answer")
	}))
	defer provider.Close()
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{"adapters":{"telegram":{"package":"alonabot-adapter-telegram","default_instance":"cross_bot","instances":{"cross_bot":{"token":"cross-component-token","api_base_url":%q,"proxy_mode":"none"}}}}}`, server.URL()))
	developmentWrite(t, filepath.Join(dir, "plugins", agentPackageID, "config.json"), fmt.Sprintf(`{"providers":[{"custom":{"name":"cross","protocol":"openai-compatible","base_url":%q,"api_key":"cross-test-key"}}],"models":[{"name":"cross","provider":"cross","model":"cross-model","type":"chat","max_tokens":128}],"chat_model":"cross"}`, provider.URL))
	args := []string{"--listen", "127.0.0.1:0", "--log-format", "json"}
	if transport == "stdio" {
		args = append(args, "--stdio")
	}
	c := startEmptyCore(t, dir, args...)
	address := c.address()
	tokenFile := filepath.Join(dir, adminauth.FirstTokenFile)
	info, err := os.Stat(tokenFile)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("first token permissions: %v %v", info, err)
	}
	output, err := exec.CommandContext(t.Context(), filepath.Join(binDir, "alonabot"), "admin-token", "show", "--data-dir", dir).CombinedOutput()
	if err != nil {
		t.Fatalf("show first token: %v", err)
	}
	token := strings.TrimSpace(string(output))
	raw, err := os.ReadFile(tokenFile)
	if err != nil || token == "" || token != strings.TrimSpace(string(raw)) {
		t.Fatalf("show did not preserve the first token file: %v", err)
	}
	base, err := url.Parse(settings.Defaults().BaseURL)
	if err != nil {
		t.Fatal(err)
	}
	var cookie *http.Cookie
	var csrf string
	call := func(method, path, body, origin, csrfToken string, authenticated bool, want int) ([]byte, []*http.Cookie) {
		t.Helper()
		request, err := http.NewRequestWithContext(t.Context(), method, "http://"+address+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Host = base.Host
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", origin)
		request.Header.Set("X-Alonabot-CSRF", csrfToken)
		if authenticated {
			request.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		content, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: got %d want %d: %s\nCore:\n%s", method, path, response.StatusCode, want, content, c.log())
		}
		return content, response.Cookies()
	}
	// 2. 首次认证领取 Cookie；来源和 CSRF 拒绝均发生在组件动作之前。
	call(http.MethodGet, "/api/v1/packages", "", "", "", false, http.StatusUnauthorized)
	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		t.Fatal(err)
	}
	call(http.MethodPost, "/api/auth/login", string(body), "http://untrusted.invalid", "", false, http.StatusForbidden)
	if _, err := os.Stat(tokenFile); err != nil {
		t.Fatalf("rejected login consumed first token: %v", err)
	}
	content, cookies := call(http.MethodPost, "/api/auth/login", string(body), base.String(), "", false, http.StatusOK)
	var login struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(content, &login); err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || login.CSRFToken == "" {
		t.Fatal("login did not return one cookie and CSRF token")
	}
	cookie, csrf = cookies[0], login.CSRFToken
	if cookie.Name != "alonabot_admin" || cookie.Path != "/" || cookie.Domain != "" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Secure {
		t.Fatal("management cookie attributes do not match the HTTP contract")
	}
	if _, err := os.Stat(tokenFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("first token file after successful login: %v", err)
	}
	packages := func() []manage.Package {
		t.Helper()
		content, _ := call(http.MethodGet, "/api/v1/packages", "", "", "", true, http.StatusOK)
		var items []manage.Package
		if err := json.Unmarshal(content, &items); err != nil {
			t.Fatal(err)
		}
		return items
	}
	// 3. CLI Bearer 与浏览器 Cookie 观察同一安装与激活状态。
	for i, project := range projects {
		c.management(t, dir, token, "install", archives[i], "--yes")
		items := packages()
		if len(items) != i+1 || items[i].ID != project.id || items[i].Enabled || items[i].Active != "" || items[i].State != "Installed" {
			t.Fatalf("installed packages: %+v", items)
		}
		action, err := json.Marshal(map[string]string{"id": project.id})
		if err != nil {
			t.Fatal(err)
		}
		call(http.MethodPost, "/api/v1/packages/enable", string(action), base.String(), "", true, http.StatusUnauthorized)
		call(http.MethodPost, "/api/v1/packages/enable", string(action), "http://untrusted.invalid", csrf, true, http.StatusUnauthorized)
		if after := packages(); !reflect.DeepEqual(items, after) {
			t.Fatalf("rejected enable changed packages: before=%+v after=%+v", items, after)
		}
		call(http.MethodPost, "/api/v1/packages/enable", string(action), base.String(), csrf, true, http.StatusOK)
		items = packages()
		if !items[i].Enabled || items[i].Active == "" || items[i].State != "Activated" {
			t.Fatalf("enabled package: %+v", items[i])
		}
	}
	telegramEventually(t, c, func() bool { return server.Calls("getUpdates") > 0 })
	sequence := 0
	send := func(text, want string) {
		t.Helper()
		sequence++
		telegramMessage(t, server, sequence, map[string]any{"text": text, "chat": map[string]any{"id": 9, "type": "private"}})
		telegramEventually(t, c, func() bool {
			for _, request := range server.Requests() {
				if (request.Method == "sendMessage" || request.Method == "editMessageText") && request.Params["text"] == want {
					return true
				}
			}
			return false
		})
	}
	c.management(t, dir, token, "admin", "add", "telegram:88")
	call(http.MethodPut, "/api/v1/plugins/echo/policies", `{"scope":{"kind":"plugin"},"permissions":["admin"]}`, base.String(), csrf, true, http.StatusOK)
	send("/echo installed chain", "installed chain")
	send("/note set cross persisted", "Saved cross")
	send("/note get cross", "cross: persisted")
	// 配置通过 Pending 编辑，激活后再观察实际消息。
	items := packages()
	var old string
	for _, item := range items {
		if item.ID == "alonabot-plugin-configdb" {
			old = item.Active
		}
	}
	configBody, err := json.Marshal(map[string]any{"base_generation": old, "patch": map[string]any{"greeting": "Cross"}})
	if err != nil {
		t.Fatal(err)
	}
	call(http.MethodPatch, "/api/v1/plugins/alonabot-plugin-configdb/config/pending", string(configBody), base.String(), csrf, true, http.StatusOK)
	call(http.MethodPost, "/api/v1/plugins/alonabot-plugin-configdb/reload", "", base.String(), csrf, true, http.StatusOK)
	configRaw, _ := call(http.MethodGet, "/api/v1/plugins/alonabot-plugin-configdb/config", "", "", "", true, http.StatusOK)
	var configuration struct {
		ActiveGeneration string         `json:"active_generation"`
		Active           map[string]any `json:"active"`
		Pending          any            `json:"pending"`
	}
	if err := json.Unmarshal(configRaw, &configuration); err != nil {
		t.Fatal(err)
	}
	if configuration.ActiveGeneration == old || configuration.Active["greeting"] != "Cross" || configuration.Pending != nil {
		t.Fatalf("activated configuration: %s", configRaw)
	}
	send("/note hello Reader", "Hello, Reader")
	c.management(t, dir, token, "restart", "echo")
	send("/echo after restart", "after restart")
	call(http.MethodGet, "/plugins/showcase/status", "", "", "", true, http.StatusOK)
	for _, path := range []string{"/status", "/page"} {
		call(http.MethodGet, "/plugins/showcase"+path, "", "", "", false, http.StatusUnauthorized)
	}
	page, _ := call(http.MethodGet, "/plugins/showcase/page", "", "", "", true, http.StatusOK)
	for _, text := range []string{`href="kit.css"`, `src="kit.js"`, "window.AlonaBot.request", "window.AlonaBot.t"} {
		if !strings.Contains(string(page), text) {
			t.Fatalf("showcase page is missing %q", text)
		}
	}
	// 页面保持管理认证，套件由真实插件按静态资源公开提供。
	for _, resource := range []struct{ file, mime string }{{"kit.js", "text/javascript"}, {"kit.css", "text/css"}} {
		var etag string
		for _, conditional := range []bool{false, true} {
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address+"/plugins/showcase/"+resource.file, nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Host = base.Host
			if conditional {
				request.Header.Set("If-None-Match", etag)
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			data, readErr := io.ReadAll(response.Body)
			if err := errors.Join(readErr, response.Body.Close()); err != nil {
				t.Fatal(err)
			}
			if conditional {
				if response.StatusCode != http.StatusNotModified || len(data) != 0 {
					t.Fatalf("conditional resource %s: %d %s", resource.file, response.StatusCode, data)
				}
			} else {
				etag = response.Header.Get("ETag")
				if response.StatusCode != http.StatusOK || len(data) == 0 || etag == "" || !strings.HasPrefix(response.Header.Get("Content-Type"), resource.mime) {
					t.Fatalf("public resource %s: %d %v", resource.file, response.StatusCode, response.Header)
				}
			}
		}
	}
	c.management(t, dir, token, "pin", "echo", "0.1.0")
	c.management(t, dir, token, "pin", "echo", "")
	c.management(t, dir, token, "disable", "echo")
	c.management(t, dir, token, "enable", "echo")
	send("/echo after enable", "after enable")
	// Agent 命令使用管理覆盖名，避开 echo 的同名 ask 声明。
	policiesRaw, _ := call(http.MethodGet, "/api/v1/plugins/"+agentPackageID+"/policies", "", "", "", true, http.StatusOK)
	var policies struct {
		Commands []struct {
			RouteID string     `json:"route_id"`
			Names   [][]string `json:"declared_names"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(policiesRaw, &policies); err != nil {
		t.Fatal(err)
	}
	foundAsk := false
	for _, command := range policies.Commands {
		for _, name := range command.Names {
			if len(name) == 1 && name[0] == "ask" {
				call(http.MethodPut, "/api/v1/plugins/"+agentPackageID+"/commands/"+url.PathEscape(command.RouteID)+"/names", `{"names":[["crossask"]]}`, base.String(), csrf, true, http.StatusOK)
				foundAsk = true
			}
		}
	}
	if !foundAsk {
		t.Fatal("Agent ask command not registered")
	}
	send("/crossask verify chain", "cross component answer")
	providerMu.Lock()
	captured := append([]agentPackageRequest(nil), providerRequests...)
	providerMu.Unlock()
	if len(captured) != 1 || captured[0].Model != "cross-model" || captured[0].MaxTokens != 128 || !captured[0].Stream {
		t.Fatalf("Provider requests: %+v", captured)
	}
	// 升级与回退都经候选编排；回退产生新代次，继续实际消息验证。
	echoPackages := packages()
	var originalGeneration string
	for _, item := range echoPackages {
		if item.ID == "echo" {
			originalGeneration = item.Active
		}
	}
	next := validationPackage(t, archives[0], "0.2.0", nil, "")
	c.management(t, dir, token, "update", "echo", "--archive", next, "--yes")
	send("/echo after update", "after update")
	c.management(t, dir, token, "rollback", "echo", originalGeneration, "--yes")
	var reverted manage.Package
	for _, item := range packages() {
		if item.ID == "echo" {
			reverted = item
		}
	}
	if reverted.Version != "0.1.0" || reverted.Active == originalGeneration || !reverted.Enabled {
		t.Fatalf("rollback package: %+v", reverted)
	}
	send("/echo after rollback", "after rollback")
	call(http.MethodGet, "/", "", "", "", false, http.StatusOK)
	c.management(t, dir, token, "uninstall", "echo")
	c.management(t, dir, token, "install", archives[0], "--yes")
	c.management(t, dir, token, "enable", "echo")
	send("/echo after reinstall", "after reinstall")
	var listed []manage.Package
	if err := json.Unmarshal([]byte(c.management(t, dir, token, "list")), &listed); err != nil {
		t.Fatal(err)
	}
	if browser := packages(); !reflect.DeepEqual(browser, listed) {
		t.Fatalf("package states differ: browser=%+v CLI=%+v", browser, listed)
	}
	stopTestCore(t, c)
	for _, credential := range []string{token, cookie.Value, csrf} {
		if strings.Contains(c.log(), credential) {
			t.Fatal("Core logs contain management credentials")
		}
	}
}

// packCrossComponentWebUI 在独立副本构建正式页面，保留工作区的依赖与构建目录。
func packCrossComponentWebUI(t *testing.T) string {
	t.Helper()
	root := filepath.Join("../..", "tmp", "test")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	project, err := os.MkdirTemp(root, "t10-webui-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(project); err != nil {
			t.Error(err)
		}
	})
	if err := os.CopyFS(project, os.DirFS("../../web/admin")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"ci"}, {"run", "build"}} {
		command := exec.CommandContext(t.Context(), "npm", args...)
		command.Dir = project
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build official WebUI %v: %v\n%s", args, err, output)
		}
	}
	return packAgentComponent(t, project, "alonabot-webui-admin")
}

func testCrossPermissionPackages(t *testing.T, transport string) {
	projects := []struct{ directory, id string }{
		{"../../adapters/telegram", "alonabot-adapter-telegram"},
		{"../../examples/plugins/echo", "echo"},
		{"../../examples/plugins/lifecycle", "demo"},
	}
	archives := make([]string, len(projects))
	for i, project := range projects {
		archives[i] = packAgentComponent(t, project.directory, project.id)
	}
	server := telegramtest.NewServer(t, "boundary-token")
	server.SetMe(101, "boundary_bot")
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{"adapters":{"telegram":{"package":"alonabot-adapter-telegram","default_instance":"boundary_bot","instances":{"boundary_bot":{"token":"boundary-token","api_base_url":%q,"proxy_mode":"none"}}}}}`, server.URL()))
	args := []string{"--listen", "127.0.0.1:0", "--log-format", "json", "--log-level", "debug"}
	if transport == "stdio" {
		args = append(args, "--stdio")
	}
	c := startEmptyCore(t, dir, args...)
	for i, project := range projects {
		c.management(t, dir, "", "install", archives[i], "--yes")
		c.management(t, dir, "", "enable", project.id)
	}
	put := func(path, body string) {
		t.Helper()
		if status, raw := fixtureHTTP(t, c, dir, http.MethodPut, path, body); status != http.StatusOK {
			t.Fatalf("boundary policy %s: %d %s", path, status, raw)
		}
	}
	// demo 的稳定名字排序胜出；echo 保留自己的真实 Handler 作为被遮蔽候选。
	put("/api/v1/plugins/echo/commands/cmd.echo/names", `{"names":[["ping"]]}`)
	status, raw := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/plugins/demo/policies")
	var policies struct {
		Shadows []struct {
			Winner struct {
				PluginID string `json:"plugin_id"`
			} `json:"winner"`
		} `json:"shadows"`
	}
	if err := json.Unmarshal(raw, &policies); err != nil || status != http.StatusOK || len(policies.Shadows) != 1 || policies.Shadows[0].Winner.PluginID != "demo" {
		t.Fatalf("command shadow: %d %s %v", status, raw, err)
	}
	telegramEventually(t, c, func() bool { return server.Calls("getUpdates") > 0 })
	sequence := 0
	send := func() {
		sequence++
		telegramMessage(t, server, sequence, map[string]any{"text": "/ping", "chat": map[string]any{"id": 9, "type": "private"}})
	}
	awaitReplies := func(want int) {
		t.Helper()
		telegramEventually(t, c, func() bool { return server.Calls("sendMessage") >= want })
		if got := server.Calls("sendMessage"); got != want {
			t.Fatalf("unexpected fallback reply count: %d want %d", got, want)
		}
		for _, request := range server.Requests() {
			if request.Method == "sendMessage" && request.Params["text"] != "pong" {
				t.Fatalf("shadowed Handler ran: %+v", request)
			}
		}
	}
	send()
	awaitReplies(1)
	for _, body := range []string{
		`{"scope":{"kind":"route","route_id":"cmd.ping"},"enabled":false}`,
		`{"scope":{"kind":"route","route_id":"cmd.ping"},"permissions":["admin"]}`,
	} {
		put("/api/v1/plugins/demo/policies", body)
		before := strings.Count(c.log(), "command policy denied")
		send()
		telegramEventually(t, c, func() bool { return strings.Count(c.log(), "command policy denied") > before })
		if got := server.Calls("sendMessage"); got != 1 {
			t.Fatalf("rejected winner fell through: %d", got)
		}
	}
	// 管理凭证仅修改策略；聊天 Actor 通过独立的 Core Admin 列表取得权限。
	c.management(t, dir, "", "admin", "add", "telegram:88")
	send()
	awaitReplies(2)
	c.management(t, dir, "", "admin", "remove", "telegram:88")
	put("/api/v1/plugins/demo/policies", `{"scope":{"kind":"route","route_id":"cmd.ping"}}`)
	send()
	awaitReplies(3)
	stopTestCore(t, c)
}
