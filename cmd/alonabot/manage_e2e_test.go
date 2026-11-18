//go:build unix

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"

	"github.com/AlonaBot/AlonaBot/contracts/release"
)

var packageKey ed25519.PrivateKey

func prepareTestPackages() error {
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	packageKey = private
	if _, err = writeTestPackage(filepath.Join(binDir, "demo-package"), "demo", "plugin", "", "github.com/example/lifecycle", "1.0.0", filepath.Join(binDir, "lifecycle"), "", nil); err != nil {
		return err
	}
	for name, source := range map[string]string{"adapter-fixture": adapterFixtureSource, "command-fixture": commandFixtureSource} {
		file := filepath.Join(binDir, name+".go")
		if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
			return err
		}
		if output, err := exec.Command("go", fixtureBuildArgs("-o", filepath.Join(binDir, name), file)...).CombinedOutput(); err != nil {
			return fmt.Errorf("build fixture %s: %w: %s", name, err, output)
		}
	}
	return nil
}

// fixtureBuildArgs 让手工组件构建沿用外层测试的 race 依赖缓存。
func fixtureBuildArgs(args ...string) []string {
	build := []string{"build"}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "-race" && setting.Value == "true" {
				build = append(build, "-race")
			}
		}
	}
	return append(build, args...)
}

// 本切片首次接通管理 CLI、依赖、替换和设置写回，以下场景需要多个端到端入口。
func TestManagementLifecycle(t *testing.T) {
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
	c.management(t, dir, "", "install", filepath.Join(binDir, "demo-package", "package.tar.gz"), "--yes")
	item := findPackage(t, c, dir, "demo")
	if item.Enabled || item.Active != "" || item.State != "Installed" || !item.Signed {
		t.Fatalf("installed: %+v", item)
	}
	c.management(t, dir, "", "enable", "demo")
	first := findPackage(t, c, dir, "demo")
	session := lastSession(t, c)
	if !first.Enabled || first.Active == "" || first.State != "Activated" {
		t.Fatalf("enabled: %+v", first)
	}
	c.management(t, dir, "", "restart", "demo")
	second := findPackage(t, c, dir, "demo")
	if second.Active != first.Active || lastSession(t, c) == session {
		t.Fatalf("restart changed generation or reused session: %+v", second)
	}
	stopTestCore(t, c)
	t.Run("package-warnings-and-build-info", func(t *testing.T) {
		archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: "../../examples/plugins/lifecycle"}, "https://example.invalid/package.tar.gz", "", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
		first := validationPackage(t, archive, "1.0.0", []string{"prepare_effects"}, "")
		var installed manage.Package
		output := c.management(t, dir, "", "install", first, "--yes")
		if err := json.Unmarshal([]byte(output), &installed); err != nil {
			t.Fatal(err)
		}
		if installed.Version != "1.0.0" || len(installed.Warnings) != 2 || installed.Warnings[1] != "prepare_effects not declared; preparation and rollback assume no external side effects" {
			t.Fatalf("install: %s", output)
		}
		c.management(t, dir, "", "enable", "demo")
		adapter := testPackage(t, "alonabot_adapter_fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{"$version":1}`)})
		c.management(t, dir, "", "install", adapter, "--yes")
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot_adapter_fixture", "one")
		c.management(t, dir, "", "enable", "alonabot_adapter_fixture")
		c.waitFor("fixture instance one started")
		next := validationPackage(t, archive, "1.1.0", []string{"migration_mode", "prepare_effects"}, "")
		output = c.management(t, dir, "", "update", "demo", "--archive", next, "--yes")
		var updated manage.UpdateResult
		if err := json.Unmarshal([]byte(output), &updated); err != nil {
			t.Fatal(err)
		}
		want := []string{installed.Warnings[0], "migration_mode not declared; preparation and rollback assume no external side effects", installed.Warnings[1]}
		if !slices.Equal(updated.Warnings, want) {
			t.Fatalf("update: %s; want %v", output, want)
		}
		counts := map[string]int{}
		for _, line := range strings.Split(c.log(), "\n") {
			var record struct {
				Level    string `json:"level"`
				Message  string `json:"msg"`
				PluginID string `json:"plugin_id"`
				Warning  string `json:"warning"`
			}
			if json.Unmarshal([]byte(line), &record) != nil {
				continue
			}
			if record.Level == "WARN" && record.PluginID == "demo" && record.Message == "package manifest warning" {
				counts[record.Warning]++
			}
		}
		for i, warning := range want {
			count := 2
			if i == 1 {
				count = 1
			}
			if counts[warning] != count {
				t.Fatalf("WARN %q count = %d, want %d\n%s", warning, counts[warning], count, c.log())
			}
		}
		active := findPackage(t, c, dir, "demo")
		session := lastSession(t, c)
		if active.Version != "1.1.0" || active.State != "Activated" {
			t.Fatalf("updated package: %+v", active)
		}
		for i, tc := range []struct{ change, message string }{{"change", "build-info does not match the entrypoint"}, {"remove", "archive has no build-info"}} {
			bad := validationPackage(t, archive, "1.2.0", nil, tc.change)
			output := managementFailure(t, c, dir, "update", "demo", "--archive", bad, "--yes")
			if !strings.Contains(output, "INVALID_ARGUMENT") || !strings.Contains(output, tc.message) {
				t.Fatal(output)
			}
			after := findPackage(t, c, dir, "demo")
			if after.Active != active.Active || after.Version != active.Version || after.State != "Activated" || lastSession(t, c) != session {
				t.Fatalf("rejected update changed running package: %+v", after)
			}
			status, body := fixtureHTTP(t, c, dir, http.MethodPost, "/plugins/alonabot_adapter_fixture/deliver?instance=one&user=alice&text=%2Fping")
			if status != http.StatusNoContent {
				t.Fatalf("ping: %d %s", status, body)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				status, body = fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot_adapter_fixture/sent")
				var replies []string
				if status != http.StatusOK || json.Unmarshal(body, &replies) != nil {
					t.Fatalf("sent: %d %s", status, body)
				}
				if len(replies) == i+1 && replies[i] == "pong" {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("old plugin did not reply: %s\n%s", body, c.log())
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		stopTestCore(t, c)
	})
	t.Run("webui-process-command-invoked", func(t *testing.T) {
		// 1. 使用真实构建产物安装 WebUI、Adapter 与命令组件。
		archive := packWebUIFixture(t, "1.0.0", webUIFixtureSource, nil)
		dir := t.TempDir()
		const origin = "http://webui.test:7650"
		if err := os.WriteFile(filepath.Join(dir, settings.FileName), []byte(`{"listen":"127.0.0.1:0","base_url":"`+origin+`"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		c := startEmptyCore(t, dir)
		adapter := testPackage(t, "alonabot_adapter_fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{"$version":1}`)})
		commands, err := writeTestPackage(t.TempDir(), "commands", "plugin", "", "github.com/example/commands", "1.0.0", filepath.Join(binDir, "command-fixture"), "", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{archive, adapter, commands} {
			c.management(t, dir, "", "install", path, "--yes")
		}
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot_adapter_fixture", "one")
		for _, id := range []string{"alonabot_webui_fixture", "alonabot_adapter_fixture", "commands"} {
			c.management(t, dir, "", "enable", id)
		}
		c.waitFor("fixture instance one started")
		var cookie *http.Cookie
		var csrf string
		send := func(method, path, body string, want int) (*http.Response, []byte) {
			t.Helper()
			request, err := http.NewRequest(method, "http://"+c.address()+path, strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			request.Host = "webui.test"
			request.Header.Set("Origin", origin)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Alonabot-CSRF", csrf)
			if cookie != nil {
				request.AddCookie(cookie)
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			raw, readErr := io.ReadAll(response.Body)
			if err := errors.Join(readErr, response.Body.Close()); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != want {
				t.Fatalf("%s %s: %d, want %d: %s\n%s", method, path, response.StatusCode, want, raw, c.log())
			}
			return response, raw
		}
		// 2. 同一管理源的后台路由先拒绝匿名请求，再接受真实登录 Cookie。
		send(http.MethodGet, "/plugins/alonabot_webui_fixture/status", "", http.StatusUnauthorized)
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		input, err := json.Marshal(map[string]string{"token": strings.TrimSpace(string(token))})
		if err != nil {
			t.Fatal(err)
		}
		response, raw := send(http.MethodPost, "/api/auth/login", string(input), http.StatusOK)
		var login struct {
			CSRFToken string `json:"csrf_token"`
		}
		if err := json.Unmarshal(raw, &login); err != nil {
			t.Fatal(err)
		}
		if len(response.Cookies()) != 1 || login.CSRFToken == "" {
			t.Fatal("login did not return a cookie and CSRF token")
		}
		cookie, csrf = response.Cookies()[0], login.CSRFToken
		response, _ = send(http.MethodGet, "/plugins/alonabot_webui_fixture/status", "", http.StatusOK)
		if strings.Contains(response.Header.Get("Content-Security-Policy"), "sandbox") {
			t.Fatal("WebUI backend was isolated from its management origin")
		}
		// 3. 经 Adapter 送入命令，再从 WebUI 进程读取其事件路由收到的结果。
		send(http.MethodPost, "/plugins/alonabot_adapter_fixture/deliver?instance=one&user=alice&text=%2Finspect", "", http.StatusNoContent)
		deadline := time.Now().Add(5 * time.Second)
		for {
			_, raw := send(http.MethodGet, "/plugins/alonabot_webui_fixture/status", "", http.StatusOK)
			var received []struct {
				PluginID, RouteID, CommandName, Result string
				DurationMS                             uint64
			}
			if err := json.Unmarshal(raw, &received); err != nil {
				t.Fatal(err)
			}
			if len(received) != 0 {
				if len(received) != 1 || received[0].PluginID != "commands" || received[0].RouteID != "cmd.inspect" || received[0].CommandName != "inspect" || received[0].Result != "ok" {
					t.Fatalf("command invocation results: %s", raw)
				}
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("WebUI received no command invocation: %s\n%s", raw, c.log())
			}
			time.Sleep(10 * time.Millisecond)
		}
		stopTestCore(t, c)
	})
}

const webUIFixtureManifest = `id: alonabot_webui_fixture
name: WebUI Fixture
version: %s
role: webui
source: github.com/example/webui_fixture
assets: assets
admin_api_min_major: 1
admin_api_max_major: 1
`

const webUIFixtureSource = `package fixture

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func New() *sdk.Plugin {
	p := sdk.NewPlugin()
	var mu sync.Mutex
	var received []message.CommandInvokedPayload
	p.On("command.invoked", sdk.CommandInvoked).Handle(func(_ *sdk.Context, payload *message.CommandInvokedPayload) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, *payload)
		return nil
	})
	p.Route(http.MethodGet, "/status").Public(sdk.AuthCallbackState).HandleFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(received); err != nil {
			slog.Error("write command invocation results failed", "error", err)
		}
	})
	return p
}
`

func packWebUIFixture(t *testing.T, version, source string, files map[string]string) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if files == nil {
		files = make(map[string]string)
	}
	if _, ok := files["plugin.yaml"]; !ok {
		files["plugin.yaml"] = fmt.Sprintf(webUIFixtureManifest, version)
	}
	files["assets/index.html"] = "static-original"
	if source != "" {
		files["main.go"] = source
		files["go.mod"] = "module example.com/webui\ngo 1.27.1\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => " + strconv.Quote(root) + "\n"
		// 只让临时 module 填充 SDK 的传递依赖。
		t.Setenv("GOFLAGS", os.Getenv("GOFLAGS")+" -mod=mod")
	}
	for name, raw := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: dir}, "https://example.invalid/webui.tar.gz", "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

// validationPackage 只改真实 pack 产物的声明与构建信息，并同步归档发布清单。
func validationPackage(t *testing.T, original, version string, missing []string, buildInfoChange string) string {
	t.Helper()
	file, err := os.Open(original)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	var archive bytes.Buffer
	compressed := gzip.NewWriter(&archive)
	writer := tar.NewWriter(compressed)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		switch header.Name {
		case "plugin.yaml":
			parsed, err := manifest.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			raw = bytes.Replace(raw, []byte("version: "+parsed.Manifest.Version+"\n"), []byte("version: "+version+"\n"), 1)
			for _, field := range missing {
				raw = bytes.Replace(raw, []byte(field+": none\n"), nil, 1)
			}
		case "build-info":
			if buildInfoChange == "remove" {
				continue
			}
			if buildInfoChange == "change" {
				index := bytes.Index(raw, []byte("go1."))
				if index < 0 {
					t.Fatal("build-info has no Go version")
				}
				raw[index+2] = '2'
			}
		}
		header.Size = int64(len(raw))
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(raw); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(original), "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	var released release.Release
	if err := json.Unmarshal(raw, &released); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(archive.Bytes())
	released.Version, released.Assets[0].Size, released.Assets[0].SHA256 = version, int64(archive.Len()), hex.EncodeToString(digest[:])
	raw, err = json.Marshal(released)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for name, content := range map[string][]byte{"package.tar.gz": archive.Bytes(), "release.json": raw} {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "package.tar.gz")
}

func TestManagementDependencies(t *testing.T) {
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
	plugin := testPackage(t, "dependent", "plugin", "", "github.com/example/dependent", "adapter_dependencies:\n  - adapter_id: devloop\n    required: true\n", nil)
	c.management(t, dir, "", "install", plugin, "--yes")
	if output := managementFailure(t, c, dir, "enable", "dependent"); !strings.Contains(output, "FAILED_PRECONDITION") || !strings.Contains(output, "devloop") {
		t.Fatal(output)
	}
	adapter := testPackage(t, "alonabot_adapter_fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{"$version":1,"token":{"type":"string","required":true}}`)})
	c.management(t, dir, "", "install", adapter, "--yes")
	c.management(t, dir, "", "enable", "alonabot_adapter_fixture")
	c.management(t, dir, "", "enable", "dependent")
	if output := managementFailure(t, c, dir, "disable", "alonabot_adapter_fixture"); !strings.Contains(output, "dependent") || !strings.Contains(output, "--yes") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "disable", "alonabot_adapter_fixture", "--yes")
	if findPackage(t, c, dir, "dependent").State != "Activated" || findPackage(t, c, dir, "alonabot_adapter_fixture").Enabled {
		t.Fatal("dependency confirmation changed dependents")
	}
	stopTestCore(t, c)
}

func TestManagementUninstallScopes(t *testing.T) {
	dir := t.TempDir()
	c := startCore(t, dir, "--listen", "127.0.0.1:0")
	c.waitFor(activatedRe)
	active := findPackage(t, c, dir, "demo").Active
	stopTestCore(t, c)
	store, err := state.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddSecret(context.Background(), state.Secret{PluginID: "demo", RefID: "fixture-secret", Value: []byte("fixture-value")}); err != nil {
		t.Fatal(err)
	}
	// 保留配置引用该 secret，验证卸载范围时也适用正常引用回收规则。
	if err := store.PutPendingConfig(context.Background(), state.PendingConfig{PluginID: "demo", BaseGeneration: active, ConfigJSON: `{"key":{"$secret_ref":"fixture-secret"}}`}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutCommandPolicy(context.Background(), state.CommandPolicy{PluginID: "demo", Scope: "command:ping", Permissions: []string{"user"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "plugins", "demo", "generations", active, "fixture.data")
	if err := os.MkdirAll(filepath.Dir(marker), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("retained"), 0o600); err != nil {
		t.Fatal(err)
	}
	configuration := filepath.Join(dir, "plugins", "demo", "config.json")
	if err := os.WriteFile(configuration, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c = startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
	c.waitFor(activatedRe)
	c.management(t, dir, "", "uninstall", "demo")
	if !findPackage(t, c, dir, "demo").Uninstalled {
		t.Fatal("installation was not retained")
	}
	other := testPackage(t, "demo", "plugin", "", "github.com/other/lifecycle", "", nil)
	if output := managementFailure(t, c, dir, "install", other, "--yes"); !strings.Contains(output, "PERMISSION_DENIED") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "install", filepath.Join(binDir, "demo-package", "package.tar.gz"), "--yes")
	if raw, err := os.ReadFile(marker); err != nil || string(raw) != "retained" {
		t.Fatalf("reinstall lost data: %s %v", raw, err)
	}
	c.management(t, dir, "", "enable", "demo")
	if findPackage(t, c, dir, "demo").Active != active {
		t.Fatal("reinstall did not reuse active data")
	}
	if output := managementFailure(t, c, dir, "uninstall", "demo", "--purge-data"); !strings.Contains(output, "generation") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "uninstall", "demo", "--purge-data", "--yes")
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("data remains: %v", err)
	}
	if _, err := os.Stat(configuration); err != nil {
		t.Fatalf("editing configuration was removed: %v", err)
	}
	stopTestCore(t, c)
	store, err = state.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := store.ListSecrets(context.Background(), "demo")
	if err != nil || len(secrets) != 1 {
		t.Fatalf("secret removed: %v %v", secrets, err)
	}
	policies, err := store.CommandPolicies(context.Background(), "demo")
	if err != nil || len(policies) != 1 {
		t.Fatalf("policy removed: %v %v", policies, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	c = startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
	if output := managementFailure(t, c, dir, "uninstall", "demo", "--purge"); !strings.Contains(output, "secret") || !strings.Contains(output, "command_policy") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "uninstall", "demo", "--purge", "--yes")
	if output := c.management(t, dir, "", "list"); strings.TrimSpace(output) != "[]" {
		t.Fatal(output)
	}
	if _, err := os.Stat(configuration); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("configuration remains: %v", err)
	}
	stopTestCore(t, c)
}

func TestManagementReplacement(t *testing.T) {
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
	first := testPackage(t, "alonabot_agent_first", "agent", "", "github.com/example/agent1", "", nil)
	second := testPackage(t, "alonabot_agent_second", "agent", "", "github.com/example/agent2", "", nil)
	c.management(t, dir, "", "install", first, "--yes")
	c.management(t, dir, "", "enable", "alonabot_agent_first")
	if output := managementFailure(t, c, dir, "install", second, "--yes"); !strings.Contains(output, "ALREADY_EXISTS") || !strings.Contains(output, "alonabot_agent_first") {
		t.Fatal(output)
	}
	if output := managementFailure(t, c, dir, "install", second, "--replace"); !strings.Contains(output, "alonabot_agent_first") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "install", second, "--replace", "--yes")
	if !findPackage(t, c, dir, "alonabot_agent_first").Uninstalled || !findPackage(t, c, dir, "alonabot_agent_second").Enabled {
		t.Fatal("replacement lost enabled state")
	}
	for _, id := range []string{"devloop", "another"} {
		archive := testPackage(t, "alonabot_adapter_"+id, "adapter", id, "github.com/example/"+id, "", map[string][]byte{"adapter_config_schema.json": []byte(`{"$version":1}`)})
		c.management(t, dir, "", "install", archive, "--yes")
		if findPackage(t, c, dir, "alonabot_adapter_"+id).Uninstalled {
			t.Fatal("different adapter IDs conflict")
		}
	}
	stopTestCore(t, c)
}

func TestManagementInstancesAndCommands(t *testing.T) {
	t.Setenv("MANAGE_E2E_TOKEN", "fixture-credential-value")
	dir := t.TempDir()
	configuration := `{"listen":"127.0.0.1:0","command_prefixes":["!"],"command_prefixes_by_instance":{"one":["?"]},"admins":["devloop:initial"],"default_locale":"zh-CN"}`
	if err := os.WriteFile(filepath.Join(dir, settings.FileName), []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	c := startEmptyCore(t, dir)
	adapter := testPackage(t, "alonabot_adapter_fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{"$version":1,"token":{"type":"string","required":true,"secret":true}}`)})
	commands, err := writeTestPackage(t.TempDir(), "commands", "plugin", "", "github.com/example/commands", "1.0.0", filepath.Join(binDir, "command-fixture"), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.management(t, dir, "", "install", adapter, "--yes")
	c.management(t, dir, "", "install", commands, "--yes")
	configFile := filepath.Join(t.TempDir(), "instance.json")
	if err := os.WriteFile(configFile, []byte(`{"token":"${MANAGE_E2E_TOKEN}"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"one", "two", "fail"} {
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot_adapter_fixture", id, "--config", configFile)
	}
	raw, err := os.ReadFile(filepath.Join(dir, settings.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("${MANAGE_E2E_TOKEN}")) || bytes.Contains(raw, []byte("fixture-credential-value")) {
		t.Fatal("credential reference was changed")
	}
	c.management(t, dir, "", "enable", "alonabot_adapter_fixture")
	c.management(t, dir, "", "enable", "commands")
	for _, id := range []string{"one", "two"} {
		c.waitFor("fixture instance " + id + " started")
	}
	c.waitFor(`adapter instance start failed plugin_id=alonabot_adapter_fixture instance_id=fail`)
	check := func(instance, user, text, want string, count int) {
		t.Helper()
		path := "/plugins/alonabot_adapter_fixture/deliver?instance=" + instance + "&user=" + user + "&text=" + url.QueryEscape(text)
		if status, body := fixtureHTTP(t, c, dir, http.MethodPost, path); status != http.StatusNoContent {
			t.Fatalf("deliver: %d %s", status, body)
		}
		deadline := time.Now().Add(5 * time.Second)
		for {
			status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot_adapter_fixture/sent")
			var sent []string
			if status != http.StatusOK || json.Unmarshal(body, &sent) != nil {
				t.Fatalf("sent: %d %s", status, body)
			}
			if len(sent) >= count {
				if len(sent) != count || !strings.HasPrefix(sent[count-1], want) {
					t.Fatalf("reply: %v, want %s", sent, want)
				}
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("missing reply %s: %s\n%s", want, body, c.log())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	check("one", "initial", "?inspect", "admin=true", 1)
	check("two", "new", "!inspect", "admin=false", 2)
	c.management(t, dir, "", "admin", "add", "devloop:new")
	check("two", "new", "!inspect", "admin=true", 3)
	check("one", "initial", "?number invalid", "用法：?number", 4)
	c.management(t, dir, "", "adapter", "instance", "stop", "two")
	c.management(t, dir, "", "adapter", "instance", "start", "two")
	stopTestCore(t, c)
	c = startEmptyCore(t, dir)
	for _, id := range []string{"one", "two"} {
		c.waitFor("fixture instance " + id + " started")
	}
	c.waitFor(`adapter instance start failed plugin_id=alonabot_adapter_fixture instance_id=fail`)
	listed := c.management(t, dir, "", "list")
	if strings.Contains(listed, "fixture-credential-value") || strings.Contains(listed, "MANAGE_E2E_TOKEN") || strings.Contains(c.log(), "fixture-credential-value") {
		t.Fatal("credentials escaped through management or logs")
	}
	c.management(t, dir, "", "adapter", "instance", "delete", "two")
	raw, err = os.ReadFile(filepath.Join(dir, settings.FileName))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Adapters map[string]struct {
			Instances map[string]json.RawMessage `json:"instances"`
		} `json:"adapters"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if _, exists := document.Adapters["devloop"].Instances["two"]; exists {
		t.Fatal("deleted instance remains configured")
	}
	stopTestCore(t, c)
}

func fixtureHTTP(t *testing.T, c *core, dir, method, path string) (int, []byte) {
	t.Helper()
	token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(method, "http://"+c.address()+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(response.Body)
	if closeErr := response.Body.Close(); err != nil || closeErr != nil {
		t.Fatalf("read fixture HTTP: %v %v", err, closeErr)
	}
	return response.StatusCode, raw
}

func testPackage(t *testing.T, id, role, adapter, source, extra string, resources map[string][]byte) string {
	t.Helper()
	binary := filepath.Join(binDir, "lifecycle")
	if role == "adapter" {
		binary = filepath.Join(binDir, "adapter-fixture")
	}
	archive, err := writeTestPackage(t.TempDir(), id, role, adapter, source, "1.0.0", binary, extra, resources)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

func findPackage(t *testing.T, c *core, dir, id string) manage.Package {
	t.Helper()
	var packages []manage.Package
	if err := json.Unmarshal([]byte(c.management(t, dir, "", "list")), &packages); err != nil {
		t.Fatal(err)
	}
	for _, item := range packages {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("package %s not listed", id)
	return manage.Package{}
}

func stopTestCore(t *testing.T, c *core) {
	t.Helper()
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit %d: %s", code, c.log())
	}
}

func lastSession(t *testing.T, c *core) string {
	t.Helper()
	c.waitFor(`"msg":"OnReady"`)
	var session string
	for _, line := range strings.Split(c.log(), "\n") {
		var item struct {
			Message string `json:"msg"`
			Session string `json:"session_id"`
		}
		if json.Unmarshal([]byte(line), &item) == nil && item.Message == "OnReady" {
			session = item.Session
		}
	}
	if session == "" {
		t.Fatal("plugin session was not logged")
	}
	return session
}

func managementFailure(t *testing.T, c *core, dir string, args ...string) string {
	t.Helper()
	token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(filepath.Join(binDir, "alonabot"), append(args, "--admin-url", "http://"+c.address())...)
	command.Env = append(os.Environ(), "ALONABOT_ADMIN_TOKEN="+strings.TrimSpace(string(token)))
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("command unexpectedly succeeded: %v: %s", args, output)
	}
	return string(output)
}

const commandFixtureSource = `package main
import (
 "fmt"
 "log/slog"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
 "github.com/AlonaBot/AlonaBot/sdk/message"
)
func init(){genimpl.Register("main",genimpl.Resources{})}
func main() {
 p:=sdk.NewPlugin()
 p.Command("inspect").Handle(func(c *sdk.Context)(message.Content,error){return message.Chain{message.Text(fmt.Sprintf("admin=%t",c.Has(sdk.Admin)))},nil})
 p.Command("number").Args("n").Handle(func(*sdk.Context,int64)error{return nil})
 if err:=sdk.Run(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}
}
`

const adapterFixtureSource = `package main
import (
 "context"
 "encoding/json"
 "errors"
 "log/slog"
 "net/http"
 "os"
 "sync"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
 "github.com/AlonaBot/AlonaBot/sdk/adapter"
 "github.com/AlonaBot/AlonaBot/sdk/message"
 "github.com/AlonaBot/AlonaBot/sdk/errcode"
 example "github.com/AlonaBot/AlonaBot/examples/devbot/adapter"
)
type platform struct {adapter.Adapter; mu sync.Mutex; cores map[string]adapter.Core; sent []string}
type instance struct {adapter.Instance; platform *platform; id string}
func (p *platform) Start(ctx context.Context,req adapter.StartRequest,core adapter.Core)(adapter.Instance,error){
 if req.InstanceID=="fail" {return nil,errors.New("fixture instance rejected")}
 base,err:=p.Adapter.Start(ctx,req,core);if err!=nil{return nil,err}
 p.mu.Lock();p.cores[req.InstanceID]=core;p.mu.Unlock()
 slog.Info("fixture instance "+req.InstanceID+" started")
 return &instance{Instance:base,platform:p,id:req.InstanceID},nil
}
func(i *instance) Send(ctx context.Context,to string,chain message.Chain)(*message.SendResult,*errcode.Error){
 i.platform.mu.Lock();i.platform.sent=append(i.platform.sent,chain.Text());i.platform.mu.Unlock()
 return i.Instance.Send(ctx,to,chain)
}
func(i *instance) Stop(ctx context.Context)error{
 i.platform.mu.Lock();delete(i.platform.cores,i.id);i.platform.mu.Unlock()
 slog.Info("fixture instance "+i.id+" stopped");return i.Instance.Stop(ctx)
}
func(p *platform) Descriptor()*adapter.Descriptor{
 d:=p.Adapter.Descriptor()
 d.HTTP=[]adapter.HTTPRoute{{Method:"POST",Path:"/deliver",Handler:http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  id:=r.URL.Query().Get("instance");p.mu.Lock();core:=p.cores[id];p.mu.Unlock()
  if core==nil{http.Error(w,"instance missing",404);return}
  _,err:=core.Deliver(r.Context(),&message.Event{Type:"message.created",AdapterID:"devloop",AdapterInstanceID:id,AccountID:id,Session:message.SID("devloop","chat",""),Actor:&message.Actor{ExternalUserID:r.URL.Query().Get("user"),AdapterInstanceID:id},Payload:&message.CreatedPayload{PlatformMessageID:"fixture",Chain:message.Chain{message.Text(r.URL.Query().Get("text"))}}},nil)
  if err!=nil{http.Error(w,err.Error(),500);return};w.WriteHeader(204)
 })}, {Method:"GET",Path:"/sent",Handler:http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){p.mu.Lock();defer p.mu.Unlock();if err:=json.NewEncoder(w).Encode(p.sent);err!=nil{slog.Error("write sent failed","error",err)}})}}
 return d
}
func init(){genimpl.Register("main",genimpl.Resources{})}
func main(){p:=&platform{Adapter:example.New(),cores:map[string]adapter.Core{}};if err:=sdk.RunAdapter(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}}
`

// writeTestPackage 把真实二进制与内层 manifest 打成签名安装包，密钥仅存在本次测试中。
func writeTestPackage(dir, id, role, adapterID, source, version, executable, extra string, resources map[string][]byte) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	platformOS, arch := runtime.GOOS, runtime.GOARCH
	manifest := fmt.Sprintf("manifest_version: 1\nid: %s\nname: Integration Fixture\nversion: %s\nsource: %s\nrole: %s\n", id, version, source, role)
	if role == "webui" {
		platformOS, arch = "any", "any"
		manifest += "assets: public\nadmin_api_min_major: 1\nadmin_api_max_major: 1\n"
	} else {
		manifest += "protocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: bin/component\nconfig_version: 1\n"
		if role != "adapter" {
			manifest += "migration_mode: none\nprepare_effects: none\n"
		}
	}
	if adapterID != "" {
		manifest += "adapter_id: " + adapterID + "\n"
	}
	manifest += fmt.Sprintf("os: %s\narch: %s\n", platformOS, arch) + extra
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	files := map[string][]byte{"plugin.yaml": []byte(manifest)}
	if role != "webui" {
		raw, err := os.ReadFile(executable)
		if err != nil {
			return "", err
		}
		files["bin/component"] = raw
		info, err := buildinfo.ReadFile(executable)
		if err != nil {
			return "", err
		}
		files["build-info"] = []byte(info.String())
	} else {
		files["public/index.html"] = []byte("<!doctype html><title>Fixture</title>")
	}
	for name, raw := range resources {
		files[name] = raw
	}
	for name, raw := range files {
		mode := int64(0o600)
		if name == "bin/component" {
			mode = 0o700
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(raw))}); err != nil {
			return "", err
		}
		if _, err := tw.Write(raw); err != nil {
			return "", err
		}
	}
	if err := tw.Close(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	digest := sha256.Sum256(archive.Bytes())
	pub := packageKey.Public().(ed25519.PublicKey)
	r := release.Release{PluginID: id, Version: version, Source: source, PublisherKeyID: release.KeyID(pub), Assets: []release.Asset{{OS: platformOS, Arch: arch, URL: "https://example.invalid/package.tar.gz", Size: int64(archive.Len()), SHA256: hex.EncodeToString(digest[:])}}}
	raw, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	for name, data := range map[string][]byte{"package.tar.gz": archive.Bytes(), "release.json": raw,
		"release.json.sig": []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(packageKey, raw))), "release.pub": []byte(base64.StdEncoding.EncodeToString(pub))} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, "package.tar.gz"), nil
}
