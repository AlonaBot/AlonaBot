//go:build unix

package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
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
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
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
	for name, source := range map[string]string{"adapter-fixture": adapterFixtureSource, "command-fixture": commandFixtureSource, "job-fixture": jobFixtureSource} {
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
	t.Run("plugin-data", testManagementPluginData)
	t.Run("sleep-capable-runtime", func(t *testing.T) {
		const id, source = "demo", "github.com/example/lifecycle"
		archive := testPackage(t, id, "plugin", "", source, "sleep_capable: true\n", nil)
		dir := t.TempDir()
		start := func() *core { return startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json") }
		c := start()
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", id)
		active := findPackage(t, c, dir, id)
		if !active.Enabled || active.Active == "" || active.State != "Activated" || active.RuntimeState != string(state.RuntimeActive) {
			t.Fatalf("enabled package: %+v", active)
		}
		c.management(t, dir, "", "disable", id)
		checkStopped := func(c *core) {
			t.Helper()
			item := findPackage(t, c, dir, id)
			if item.Enabled || item.Uninstalled || item.Active != active.Active || item.Version != active.Version || item.Source != source || item.State != "Stopped" || item.RuntimeState != string(state.RuntimeStopped) {
				t.Fatalf("disabled package: %+v", item)
			}
		}
		checkStopped(c)
		stopTestCore(t, c)
		c = start()
		checkStopped(c)
		stopTestCore(t, c)
	})
	t.Run("static-job-and-token-rotation", func(t *testing.T) {
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
		address := c.address()
		raw, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		token := strings.TrimSpace(string(raw))
		base, err := url.Parse(settings.Defaults().BaseURL)
		if err != nil {
			t.Fatal(err)
		}
		call := func(method, path, token, csrf string, cookie *http.Cookie) (int, []byte, []*http.Cookie) {
			t.Helper()
			body, err := json.Marshal(map[string]string{"token": token})
			if err != nil {
				t.Fatal(err)
			}
			request, err := http.NewRequestWithContext(t.Context(), method, "http://"+address+path, bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			request.Host = base.Host
			request.Header.Set("Origin", base.String())
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
			return response.StatusCode, raw, response.Cookies()
		}
		status, raw, cookies := call(http.MethodPost, "/api/auth/login", token, "", nil)
		var login struct {
			CSRFToken string `json:"csrf_token"`
		}
		if status != http.StatusOK || json.Unmarshal(raw, &login) != nil || login.CSRFToken == "" || len(cookies) != 1 {
			t.Fatalf("login status=%d", status)
		}
		cookie := cookies[0]
		archive, err := writeTestPackage(t.TempDir(), "jobs", "plugin", "", "github.com/example/jobs", "1.0.0", filepath.Join(binDir, "job-fixture"), "", nil)
		if err != nil {
			t.Fatal(err)
		}
		c.management(t, dir, token, "install", archive, "--yes")
		c.management(t, dir, token, "enable", "jobs")
		if status, raw, _ := call(http.MethodPost, "/api/v1/plugins/jobs/jobs/pulse/pause", "", login.CSRFToken, cookie); status != http.StatusOK {
			t.Fatalf("pause status=%d: %s", status, raw)
		}
		lastRun := func() int64 {
			t.Helper()
			status, raw, _ := call(http.MethodGet, "/plugins/jobs/runs", "", "", cookie)
			var stamp int64
			if status != http.StatusOK || json.Unmarshal(raw, &stamp) != nil {
				t.Fatalf("job runs status=%d: %s", status, raw)
			}
			return stamp
		}
		// 暂停前已派发的 Handler 可能仍在途，先以调度状态确认它完成交付。
		status, raw, _ = call(http.MethodGet, "/api/v1/plugins/jobs/jobs", "", "", cookie)
		var jobs struct {
			Jobs []struct {
				Paused    bool       `json:"paused"`
				LastRunAt *time.Time `json:"last_run_at"`
				NextRunAt *time.Time `json:"next_run_at"`
			}
		}
		if status != http.StatusOK || json.Unmarshal(raw, &jobs) != nil || len(jobs.Jobs) != 1 || !jobs.Jobs[0].Paused || jobs.Jobs[0].NextRunAt != nil {
			t.Fatalf("paused jobs status=%d: %s", status, raw)
		}
		if jobs.Jobs[0].LastRunAt != nil {
			deadline := time.Now().Add(5 * time.Second)
			for lastRun() < jobs.Jobs[0].LastRunAt.UnixMilli() {
				if time.Now().After(deadline) {
					t.Fatal("pending job did not reach the plugin")
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		pausedRun := lastRun()
		time.Sleep(2 * time.Second)
		if got := lastRun(); got != pausedRun {
			t.Fatalf("paused job fired: before=%d after=%d", pausedRun, got)
		}
		if status, raw, _ := call(http.MethodPost, "/api/v1/plugins/jobs/jobs/pulse/resume", "", login.CSRFToken, cookie); status != http.StatusOK {
			t.Fatalf("resume status=%d: %s", status, raw)
		}
		deadline := time.Now().Add(5 * time.Second)
		for lastRun() == pausedRun {
			if time.Now().After(deadline) {
				t.Fatal("resumed job did not fire")
			}
			time.Sleep(10 * time.Millisecond)
		}
		status, raw, _ = call(http.MethodPost, "/api/auth/token/rotate", token, login.CSRFToken, cookie)
		var rotated struct {
			Token string `json:"token"`
		}
		if status != http.StatusOK || json.Unmarshal(raw, &rotated) != nil || rotated.Token == "" || rotated.Token == token {
			t.Fatalf("token rotation status=%d", status)
		}
		if status, _, _ := call(http.MethodGet, "/api/auth/session", "", "", cookie); status != http.StatusUnauthorized {
			t.Fatalf("old session status=%d", status)
		}
		status, raw, cookies = call(http.MethodPost, "/api/auth/login", rotated.Token, "", nil)
		if status != http.StatusOK || json.Unmarshal(raw, &login) != nil || login.CSRFToken == "" || len(cookies) != 1 {
			t.Fatalf("new token login status=%d", status)
		}
		stopTestCore(t, c)
	})
	t.Run("tool-policies", func(t *testing.T) {
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
		archive, err := writeTestPackage(t.TempDir(), "commands", "plugin", "", "github.com/example/commands", "1.0.0", filepath.Join(binDir, "command-fixture"), "", nil)
		if err != nil {
			t.Fatal(err)
		}
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", "commands")
		check := func(overridden bool) {
			t.Helper()
			status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/tools/policies")
			var directory struct {
				Policies []struct{ Service, Effect, Default, Override string }
			}
			if status != http.StatusOK || json.Unmarshal(body, &directory) != nil {
				t.Fatalf("policies: %d %s", status, body)
			}
			want := map[string]string{"read": "read_only", "write": "write", "remove": "destructive", "unknown": "unspecified"}
			if len(directory.Policies) != len(want) {
				t.Fatalf("policies = %s", body)
			}
			for _, entry := range directory.Policies {
				override := "inherit"
				if overridden && entry.Service == "read" {
					override = "deny"
				}
				if overridden && entry.Service == "write" {
					override = "allow"
				}
				if entry.Effect != want[entry.Service] || entry.Default != "confirm" || entry.Override != override {
					t.Errorf("policy = %+v, want effect=%s confirm/%s", entry, want[entry.Service], override)
				}
			}
		}
		check(false)
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(http.MethodPut, "http://"+c.address()+"/api/v1/tools/policies", strings.NewReader(`{"items":[{"plugin_id":"commands","service":"read","policy":"deny"},{"plugin_id":"commands","service":"write","policy":"allow"}]}`))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
		request.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		body, readError := io.ReadAll(response.Body)
		closeError := response.Body.Close()
		if response.StatusCode != http.StatusOK || readError != nil || closeError != nil || strings.TrimSpace(string(body)) != "{}" {
			t.Fatalf("batch PUT: status=%d body=%s read=%v close=%v", response.StatusCode, body, readError, closeError)
		}
		check(true)
		stopTestCore(t, c)
	})
	t.Run("changelog", func(t *testing.T) {
		archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: "../../examples/plugins/showcase"}, "https://example.invalid/showcase.tar.gz", "", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
		c.management(t, dir, "", "install", archive, "--yes")
		status, raw := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/packages/showcase/changelog")
		var got struct {
			Locale  string `json:"locale"`
			Entries []struct {
				Version  string `json:"version"`
				Date     string `json:"date"`
				Markdown string `json:"markdown"`
			} `json:"entries"`
		}
		if status != http.StatusOK || json.Unmarshal(raw, &got) != nil {
			t.Fatalf("changelog: %d %s", status, raw)
		}
		source, err := os.ReadFile("../../examples/plugins/showcase/CHANGELOG.md")
		if err != nil {
			t.Fatal(err)
		}
		_, body, ok := strings.Cut(string(source), "## [0.1.0] - 2026-11-17\n")
		if !ok || got.Locale != "en-US" || len(got.Entries) != 1 || got.Entries[0].Version != "0.1.0" || got.Entries[0].Date != "2026-11-17" || got.Entries[0].Markdown != body || bytes.Contains(raw, []byte("[Unreleased]")) {
			t.Fatalf("changelog entries: %s", raw)
		}
		stopTestCore(t, c)
	})
	t.Run("configuration-corrections-and-prune", func(t *testing.T) {
		archive := weatherProfilePackage(t, `{"count":{"type":"int","default":7}}`, "1.0.0")
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
		c.management(t, dir, "", "install", archive, "--yes")
		name := filepath.Join(dir, "plugins", "weather", "config.json")
		if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(`{"count":{},"obsolete":"kept"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		c.management(t, dir, "", "enable", "weather")
		initial := findPackage(t, c, dir, "weather")
		if initial.State != "Activated" {
			t.Fatalf("corrected plugin did not start: %+v", initial)
		}
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		call := func(method, path string) []byte {
			t.Helper()
			request, err := http.NewRequestWithContext(t.Context(), method, "http://"+c.address()+path, nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			raw, readErr := io.ReadAll(response.Body)
			if err := errors.Join(readErr, response.Body.Close()); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK {
				t.Fatalf("%s %s: %d %s", method, path, response.StatusCode, raw)
			}
			return raw
		}
		var snapshot struct {
			Active      map[string]json.RawMessage `json:"active"`
			Corrections []struct {
				Path, Reason string
				Original     json.RawMessage
			} `json:"corrections"`
			Unknown []string `json:"unknown"`
		}
		raw := call(http.MethodGet, "/api/v1/plugins/weather/config")
		if err := json.Unmarshal(raw, &snapshot); err != nil || string(snapshot.Active["count"]) != "7" ||
			len(snapshot.Corrections) != 1 || snapshot.Corrections[0].Path != "$.count" || snapshot.Corrections[0].Reason != "reset" ||
			string(snapshot.Corrections[0].Original) != "{}" || !slices.Equal(snapshot.Unknown, []string{`$["obsolete"]`}) {
			t.Fatalf("configuration records = %s, error = %v", raw, err)
		}
		var pending struct {
			BaseGeneration string                     `json:"base_generation"`
			Config         map[string]json.RawMessage `json:"config"`
		}
		raw = call(http.MethodPost, "/api/v1/plugins/weather/config/prune")
		if err := json.Unmarshal(raw, &pending); err != nil || pending.BaseGeneration != initial.Active ||
			string(pending.Config["count"]) != "7" || pending.Config["obsolete"] != nil {
			t.Fatalf("pruned pending = %s, error = %v", raw, err)
		}
		call(http.MethodPost, "/api/v1/plugins/weather/reload")
		raw = call(http.MethodGet, "/api/v1/plugins/weather/config")
		snapshot.Active = nil
		if err := json.Unmarshal(raw, &snapshot); err != nil || snapshot.Active["obsolete"] != nil || len(snapshot.Unknown) != 0 || len(snapshot.Corrections) != 0 {
			t.Fatalf("reloaded pruned configuration = %s, error = %v", raw, err)
		}
		stopTestCore(t, c)
	})
	t.Run("configuration-examples", func(t *testing.T) {
		archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: "../../examples/plugins/showcase"}, "https://example.invalid/showcase.tar.gz", "", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", "showcase")
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		get := func(path string) (http.Header, int, []byte) {
			t.Helper()
			request, err := http.NewRequest(http.MethodGet, "http://"+c.address()+path, nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			body, readErr := io.ReadAll(response.Body)
			if err := errors.Join(readErr, response.Body.Close()); err != nil {
				t.Fatal(err)
			}
			return response.Header, response.StatusCode, body
		}
		icon, status, _ := get("/api/v1/packages/showcase/icon")
		if status != http.StatusOK {
			t.Fatalf("icon status = %d", status)
		}
		headers, status, body := get("/api/v1/packages/showcase/examples/greeting.png")
		image, err := os.ReadFile("../../examples/plugins/showcase/examples/greeting.png")
		if err != nil {
			t.Fatal(err)
		}
		if status != http.StatusOK || !bytes.Equal(body, image) || headers.Get("Content-Type") != "image/png" {
			t.Fatalf("example: status = %d, headers = %v, bytes = %d", status, headers, len(body))
		}
		for header, want := range map[string]string{"Content-Security-Policy": "sandbox; default-src 'none'; style-src 'unsafe-inline'", "X-Content-Type-Options": "nosniff"} {
			if headers.Get(header) != want || headers.Get(header) != icon.Get(header) {
				t.Fatalf("%s = %q, icon = %q, want %q", header, headers.Get(header), icon.Get(header), want)
			}
		}
		if _, status, _ := get("/api/v1/packages/showcase/examples/unreferenced.png"); status != http.StatusNotFound {
			t.Fatalf("unreferenced example status = %d", status)
		}
		stopTestCore(t, c)
	})
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
		adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
		c.management(t, dir, "", "install", adapter, "--yes")
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", "one")
		c.management(t, dir, "", "enable", "alonabot-adapter-fixture")
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
			status, body := fixtureHTTP(t, c, dir, http.MethodPost, "/plugins/alonabot-adapter-fixture/deliver?instance=one&user=alice&text=%2Fping")
			if status != http.StatusNoContent {
				t.Fatalf("ping: %d %s", status, body)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				status, body = fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot-adapter-fixture/sent")
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
		adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
		commands, err := writeTestPackage(t.TempDir(), "commands", "plugin", "", "github.com/example/commands", "1.0.0", filepath.Join(binDir, "command-fixture"), "", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{archive, adapter, commands} {
			c.management(t, dir, "", "install", path, "--yes")
		}
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", "one")
		for _, id := range []string{"alonabot-webui-fixture", "alonabot-adapter-fixture", "commands"} {
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
		send(http.MethodGet, "/plugins/alonabot-webui-fixture/status", "", http.StatusUnauthorized)
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
		response, _ = send(http.MethodGet, "/plugins/alonabot-webui-fixture/status", "", http.StatusOK)
		if strings.Contains(response.Header.Get("Content-Security-Policy"), "sandbox") {
			t.Fatal("WebUI backend was isolated from its management origin")
		}
		// 3. 经 Adapter 送入命令，再从 WebUI 进程读取其事件路由收到的结果。
		send(http.MethodPost, "/plugins/alonabot-adapter-fixture/deliver?instance=one&user=alice&text=%2Finspect", "", http.StatusNoContent)
		deadline := time.Now().Add(5 * time.Second)
		for {
			_, raw := send(http.MethodGet, "/plugins/alonabot-webui-fixture/status", "", http.StatusOK)
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
	t.Run("profile-storage-and-assignments", func(t *testing.T) {
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
		archive := weatherProfilePackage(t, `{"city":{"type":"string","profile":true},"plain":{"type":"string"}}`, "1.0.0")
		c.management(t, dir, "", "install", archive, "--yes")
		c.management(t, dir, "", "enable", "weather")
		adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
		c.management(t, dir, "", "install", adapter, "--yes")
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", "one")
		var created struct {
			ID string `json:"id"`
		}
		output := c.management(t, dir, "", "profile", "create", "天气群")
		if err := json.Unmarshal([]byte(output), &created); err != nil || created.ID == "" {
			t.Fatalf("create: %s: %v", output, err)
		}
		c.management(t, dir, "", "profile", "set", created.ID, "weather", "city", `"北京"`)
		c.management(t, dir, "", "profile", "assign", created.ID, "one")
		c.management(t, dir, "", "profile", "assign", created.ID, "one", "devloop:chat")
		status, raw := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/profiles/assignments")
		var assignments profile.Assignments
		if status != http.StatusOK || json.Unmarshal(raw, &assignments) != nil || assignments.Instances["one"] != created.ID || len(assignments.Sessions) != 1 || assignments.Sessions[0] != (profile.Assignment{Instance: "one", Session: "devloop:chat", Profile: created.ID}) {
			t.Fatalf("assignments: %d %s", status, raw)
		}
		status, raw = fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/profiles/"+created.ID+"/plugins")
		var plugins []profile.Plugin
		if status != http.StatusOK || json.Unmarshal(raw, &plugins) != nil || len(plugins) != 1 || string(plugins[0].Overrides["city"]) != `"北京"` {
			t.Fatalf("overrides: %d %s", status, raw)
		}
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(http.MethodPatch, "http://"+c.address()+"/api/v1/profiles/"+created.ID+"/plugins/weather", strings.NewReader(`{"overrides":{"plain":"rejected"}}`))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
		request.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		raw, err = io.ReadAll(response.Body)
		if closeErr := response.Body.Close(); err != nil || closeErr != nil {
			t.Fatalf("read invalid override: %v %v", err, closeErr)
		}
		if response.StatusCode != http.StatusBadRequest || !bytes.Contains(raw, []byte("plain")) {
			t.Fatalf("unmarked field: %d %s", response.StatusCode, raw)
		}
		c.management(t, dir, "", "profile", "delete", created.ID)
		assignments = profile.Assignments{}
		status, raw = fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/profiles/assignments")
		if status != http.StatusOK || json.Unmarshal(raw, &assignments) != nil || len(assignments.Instances) != 0 || len(assignments.Sessions) != 0 {
			t.Fatalf("deleted assignments: %d %s", status, raw)
		}
		var remaining []profile.Summary
		output = c.management(t, dir, "", "profile", "list")
		if json.Unmarshal([]byte(output), &remaining) != nil || len(remaining) != 1 || remaining[0].ID != "default" {
			t.Fatalf("fallback: %s", output)
		}
		stopTestCore(t, c)
	})
	t.Run("profile-dispatch", func(t *testing.T) {
		// 1. 真实 Core、SDK 插件、回环 Adapter 与既有事件观察组件。
		dir := t.TempDir()
		c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0", "--log-format", "json")
		adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
		weather := weatherProfilePackage(t, `{"city":{"type":"string","profile":true,"default":"base"}}`, "1.0.0")
		observer := packWebUIFixture(t, "1.0.0", webUIFixtureSource, nil)
		for _, archive := range []string{adapter, observer, weather} {
			c.management(t, dir, "", "install", archive, "--yes")
		}
		for _, instance := range []string{"one", "two"} {
			c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", instance)
		}
		// weather 最后启动，lastSession 记录它自己的真实 Session。
		for _, id := range []string{"alonabot-adapter-fixture", "alonabot-webui-fixture", "weather"} {
			c.management(t, dir, "", "enable", id)
		}
		c.waitFor("fixture instance two started")
		session := lastSession(t, c)
		var created struct {
			ID string `json:"id"`
		}
		output := c.management(t, dir, "", "profile", "create", "会话一")
		if err := json.Unmarshal([]byte(output), &created); err != nil || created.ID == "" {
			t.Fatalf("create: %s: %v", output, err)
		}
		c.management(t, dir, "", "profile", "set", created.ID, "weather", "city", `"first"`)
		c.management(t, dir, "", "profile", "assign", created.ID, "one", "devloop:chat")
		var wantReplies []string
		calls := 0
		invoke := func(instance, want, result string) {
			t.Helper()
			calls++
			if want != "" {
				wantReplies = append(wantReplies, want)
			}
			status, raw := fixtureHTTP(t, c, dir, http.MethodPost, "/plugins/alonabot-adapter-fixture/deliver?instance="+instance+"&user=alice&text=%2Fcity")
			if status != http.StatusNoContent {
				t.Fatalf("deliver: %d %s", status, raw)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				status, raw = fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot-webui-fixture/status")
				var outcomes []struct{ PluginID, Result string }
				if status != http.StatusOK || json.Unmarshal(raw, &outcomes) != nil {
					t.Fatalf("outcomes: %d %s", status, raw)
				}
				status, sent := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot-adapter-fixture/sent")
				var replies []string
				if status != http.StatusOK || json.Unmarshal(sent, &replies) != nil {
					t.Fatalf("replies: %d %s", status, sent)
				}
				if len(outcomes) == calls && slices.Equal(replies, wantReplies) {
					if outcomes[calls-1].PluginID != "weather" || outcomes[calls-1].Result != result {
						t.Fatalf("outcomes: %s", raw)
					}
					return
				}
				if time.Now().After(deadline) {
					t.Fatalf("profile dispatch: outcomes=%s replies=%s want=%v\n%s", raw, sent, wantReplies, c.log())
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		// 2. 相同命令按会话读取各自视图，在线更新无需更换 Session。
		invoke("one", "first", "ok")
		invoke("two", "base", "ok")
		c.management(t, dir, "", "profile", "set", created.ID, "weather", "city", `"updated"`)
		invoke("one", "updated", "ok")
		if lastSession(t, c) != session {
			t.Fatal("override update restarted the plugin Session")
		}
		// 3. 经管理 API 关闭会话一插件；观察真实 denied 结果，再确认另一会话继续工作。
		token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(http.MethodPatch, "http://"+c.address()+"/api/v1/profiles/"+created.ID+"/plugins/weather", strings.NewReader(`{"enabled":false}`))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
		request.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		raw, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("disable: %d %s", response.StatusCode, raw)
		}
		invoke("one", "", "denied")
		invoke("two", "base", "ok")
		stopTestCore(t, c)
	})

}

const webUIFixtureManifest = `id: alonabot-webui-fixture
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
	adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{"token":{"type":"string","required":true}}`)})
	c.management(t, dir, "", "install", adapter, "--yes")
	c.management(t, dir, "", "enable", "alonabot-adapter-fixture")
	c.management(t, dir, "", "enable", "dependent")
	if output := managementFailure(t, c, dir, "disable", "alonabot-adapter-fixture"); !strings.Contains(output, "dependent") || !strings.Contains(output, "--yes") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "disable", "alonabot-adapter-fixture", "--yes")
	if findPackage(t, c, dir, "dependent").State != "Activated" || findPackage(t, c, dir, "alonabot-adapter-fixture").Enabled {
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
	first := testPackage(t, "alonabot-agent-first", "agent", "", "github.com/example/agent1", "", nil)
	second := testPackage(t, "alonabot-agent-second", "agent", "", "github.com/example/agent2", "", nil)
	c.management(t, dir, "", "install", first, "--yes")
	c.management(t, dir, "", "enable", "alonabot-agent-first")
	if output := managementFailure(t, c, dir, "install", second, "--yes"); !strings.Contains(output, "ALREADY_EXISTS") || !strings.Contains(output, "alonabot-agent-first") {
		t.Fatal(output)
	}
	if output := managementFailure(t, c, dir, "install", second, "--replace"); !strings.Contains(output, "alonabot-agent-first") {
		t.Fatal(output)
	}
	c.management(t, dir, "", "install", second, "--replace", "--yes")
	if !findPackage(t, c, dir, "alonabot-agent-first").Uninstalled || !findPackage(t, c, dir, "alonabot-agent-second").Enabled {
		t.Fatal("replacement lost enabled state")
	}
	for _, id := range []string{"devloop", "another"} {
		archive := testPackage(t, "alonabot-adapter-"+id, "adapter", id, "github.com/example/"+id, "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
		c.management(t, dir, "", "install", archive, "--yes")
		if findPackage(t, c, dir, "alonabot-adapter-"+id).Uninstalled {
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
	adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{"token":{"type":"string","required":true,"secret":true}}`)})
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
	for _, id := range []string{"one", "two"} {
		c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", id, "--config", configFile)
	}
	if output := managementFailure(t, c, dir, "adapter", "instance", "create", "alonabot-adapter-fixture", "fail", "--config", configFile); !strings.Contains(output, "fixture instance rejected") {
		t.Fatal(output)
	}
	raw, err := os.ReadFile(filepath.Join(dir, settings.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("${MANAGE_E2E_TOKEN}")) || bytes.Contains(raw, []byte("fixture-credential-value")) {
		t.Fatal("credential reference was changed")
	}
	if adapter := findPackage(t, c, dir, "alonabot-adapter-fixture"); !adapter.Enabled || adapter.State != "Activated" {
		t.Fatalf("adapter was not automatically enabled: %+v", adapter)
	}
	c.management(t, dir, "", "enable", "commands")
	for _, id := range []string{"one", "two"} {
		c.waitFor("fixture instance " + id + " started")
	}
	status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/api/v1/adapter/instances")
	var instances struct {
		Instances []struct {
			ID        string `json:"instance_id"`
			State     string `json:"state"`
			SafeError string `json:"safe_error"`
		} `json:"instances"`
	}
	if status != http.StatusOK || json.Unmarshal(body, &instances) != nil {
		t.Fatalf("instances: %d %s", status, body)
	}
	if len(instances.Instances) != 3 || instances.Instances[0].ID != "fail" || instances.Instances[0].State != "failed" || !strings.Contains(instances.Instances[0].SafeError, "fixture instance rejected") {
		t.Fatalf("failed instance was not retained: %s", body)
	}
	check := func(instance, user, text, want string, count int) {
		t.Helper()
		path := "/plugins/alonabot-adapter-fixture/deliver?instance=" + instance + "&user=" + user + "&text=" + url.QueryEscape(text)
		if status, body := fixtureHTTP(t, c, dir, http.MethodPost, path); status != http.StatusNoContent {
			t.Fatalf("deliver: %d %s", status, body)
		}
		deadline := time.Now().Add(5 * time.Second)
		for {
			status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot-adapter-fixture/sent")
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
	c.waitFor(`adapter instance start failed plugin_id=alonabot-adapter-fixture instance_id=fail`)
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
 "github.com/AlonaBot/AlonaBot/sdk/service"
 "github.com/AlonaBot/AlonaBot/sdk/tool"
)
type toolArguments struct{}
func init(){genimpl.Register("main",genimpl.Resources{ServiceSchemas:map[string][]byte{"main.toolArguments":[]byte("{\"type\":\"object\"}")}})}
func main() {
 p:=sdk.NewPlugin()
 p.Command("inspect").Handle(func(c *sdk.Context)(message.Content,error){return message.Chain{message.Text(fmt.Sprintf("admin=%t",c.Has(sdk.Admin)))},nil})
 p.Command("number").Args("n").Handle(func(*sdk.Context,int64)error{return nil})
 for _,entry:=range []struct{name string; effect tool.Effect}{{"read",tool.ReadOnly},{"write",tool.Write},{"remove",tool.Destructive},{"unknown",tool.Unspecified}}{
  builder:=service.Serve(p,service.JSON[toolArguments,toolArguments]("commands/"+entry.name)).Tool("Fixture tool",tool.DefaultConfirm)
  if entry.effect!=tool.Unspecified {builder.Effect(entry.effect)}
  builder.Handle(func(*sdk.Context,toolArguments)(toolArguments,error){return toolArguments{},nil})
 }
 if err:=sdk.Run(p);err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}
}
`

const jobFixtureSource = `package main
import (
 "encoding/json"
 "log/slog"
 "net/http"
 "os"
 "sync/atomic"
 "time"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
)
func init(){genimpl.Register("main",genimpl.Resources{})}
func main(){
 p:=sdk.NewPlugin()
 var lastRun atomic.Int64
 p.Every("pulse",time.Second).Handle(func(c *sdk.Context)error{lastRun.Store(c.Event().OccurredAt.UnixMilli());return nil})
 p.Route(http.MethodGet,"/runs").HandleFunc(func(w http.ResponseWriter,_ *http.Request){if err:=json.NewEncoder(w).Encode(lastRun.Load());err!=nil{slog.Error("write runs","error",err)}})
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
		manifest += "protocol_min_major: 1\nprotocol_max_major: 1\nrequired_capabilities: []\nentrypoint: bin/component\n"
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

// weatherProfilePackage 打包带真实 schema 资源的天气插件，供配置文件管理路径使用。
func weatherProfilePackage(t *testing.T, schema, version string) string {
	t.Helper()
	dir := t.TempDir()
	source := fmt.Sprintf(`package main
import (
 "context"
 "log/slog"
 "os"
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/genimpl"
 "github.com/AlonaBot/AlonaBot/sdk/message"
)
func init(){genimpl.Register("main",genimpl.Resources{ConfigSchema:[]byte(%q)})}
func main(){
 p:=sdk.NewPlugin()
 p.OnReady(func(context.Context)error{slog.Info("OnReady");return nil})
 p.Command("city").Handle(func(c *sdk.Context)(message.Content,error){return message.Chain{message.Text(c.Config().String("city"))},nil})
 if err:=sdk.Run(p);err!=nil{slog.Error("weather fixture failed","error",err);os.Exit(1)}}
`, schema)
	file, binary := filepath.Join(dir, "main.go"), filepath.Join(dir, "weather")
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("go", fixtureBuildArgs("-o", binary, file)...).CombinedOutput(); err != nil {
		t.Fatalf("build weather fixture: %v: %s", err, output)
	}
	archive, err := writeTestPackage(dir, "weather", "plugin", "", "github.com/example/weather", version, binary, "", map[string][]byte{"config_schema.json": []byte(schema)})
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

// testManagementPluginData 经认证 API 修改运行中的受管数据，再由真实插件读取。
func testManagementPluginData(t *testing.T) {
	// 1. 复用动态构建和安装包工程，启动实际 Core、Adapter 与 SQLite 插件。
	archive, err := pluginkit.Pack(t.Context(), generate.Options{Dir: "../../examples/plugins/configdb"}, "https://example.invalid/configdb.tar.gz", "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	c := startEmptyCore(t, dir, "--listen", "127.0.0.1:0")
	adapter := testPackage(t, "alonabot-adapter-fixture", "adapter", "devloop", "github.com/example/adapter", "", map[string][]byte{"adapter_config_schema.json": []byte(`{}`)})
	c.management(t, dir, "", "install", adapter, "--yes")
	c.management(t, dir, "", "adapter", "instance", "create", "alonabot-adapter-fixture", "one")
	c.management(t, dir, "", "enable", "alonabot-adapter-fixture")
	c.waitFor("fixture instance one started")
	const id = "alonabot-plugin-configdb"
	c.management(t, dir, "", "install", archive, "--yes")
	c.management(t, dir, "", "enable", id)
	token, err := os.ReadFile(filepath.Join(dir, "admin_token"))
	if err != nil {
		t.Fatal(err)
	}
	api := func(method, path, body string) []byte {
		t.Helper()
		request, err := http.NewRequestWithContext(t.Context(), method, "http://"+c.address()+"/api/v1/plugins/"+id+"/data/"+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		raw, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s %s: %d %s", method, path, response.StatusCode, raw)
		}
		return raw
	}
	replies := 0
	pluginRead := func(command, want string) {
		t.Helper()
		path := "/plugins/alonabot-adapter-fixture/deliver?instance=one&user=alice&text=" + url.QueryEscape(command)
		if status, body := fixtureHTTP(t, c, dir, http.MethodPost, path); status != http.StatusNoContent {
			t.Fatalf("deliver: %d %s", status, body)
		}
		replies++
		deadline := time.Now().Add(5 * time.Second)
		for {
			status, body := fixtureHTTP(t, c, dir, http.MethodGet, "/plugins/alonabot-adapter-fixture/sent")
			var sent []string
			if status != http.StatusOK || json.Unmarshal(body, &sent) != nil {
				t.Fatalf("sent: %d %s", status, body)
			}
			if len(sent) >= replies {
				if len(sent) != replies || sent[replies-1] != want {
					t.Fatalf("plugin reply %v, want %s", sent, want)
				}
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("missing reply %s: %s\n%s", want, body, c.log())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	pluginRead("/note set seed plugin-seed", "Saved seed")
	if raw := api("GET", "tables", ""); !bytes.Contains(raw, []byte(`"name":"configdb_note"`)) {
		t.Fatal(string(raw))
	}
	if raw := api("GET", "tables/configdb_note/rows", ""); !bytes.Contains(raw, []byte(`"value":"plugin-seed"`)) {
		t.Fatal(string(raw))
	}
	// 2. 行、KV 与共用文件均在运行中改写；插件下一次调用读取同一数据。
	raw := api("POST", "tables/configdb_note/rows", `{"values":{"key":"api","value":"initial"}}`)
	var inserted struct {
		RowID int64 `json:"rowid"`
	}
	if err := json.Unmarshal(raw, &inserted); err != nil {
		t.Fatal(err)
	}
	api("PATCH", "tables/configdb_note/rows/"+strconv.FormatInt(inserted.RowID, 10), `{"values":{"value":"api-updated"}}`)
	api("PUT", "kv/topic", `{"text":"kv-api"}`)
	api("PUT", "files/content?path=note.txt", "file-api")
	pluginRead("/note get api", "api: api-updated")
	pluginRead("/note kvget topic", "topic: kv-api")
	pluginRead("/note fileget", "note.txt: file-api")
	raw = api("GET", "tables/configdb_note/rows?where.key=seed", "")
	var seeded struct {
		Rows []struct {
			RowID int64 `json:"rowid"`
		}
	}
	if err := json.Unmarshal(raw, &seeded); err != nil || len(seeded.Rows) != 1 {
		t.Fatalf("seed rows: %s %v", raw, err)
	}
	api("DELETE", "tables/configdb_note/rows/"+strconv.FormatInt(seeded.Rows[0].RowID, 10), "")
	// 3. 打开导出快照并核对文件，随后验证清空保留结构和真实迁移账本。
	raw = api("GET", "export", "")
	exported, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	foundDatabase, foundFile := false, false
	for _, entry := range exported.File {
		if entry.Name != "db.sqlite" && entry.Name != "plugin_data/note.txt" {
			continue
		}
		file, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, readErr := io.ReadAll(file)
		if err := errors.Join(readErr, file.Close()); err != nil {
			t.Fatal(err)
		}
		if entry.Name == "plugin_data/note.txt" {
			foundFile = true
			if string(content) != "file-api" {
				t.Fatal(string(content))
			}
			continue
		}
		foundDatabase = true
		snapshot := filepath.Join(t.TempDir(), "db.sqlite")
		if err := os.WriteFile(snapshot, content, 0o600); err != nil {
			t.Fatal(err)
		}
		db, err := sql.Open("sqlite", snapshot)
		if err != nil {
			t.Fatal(err)
		}
		var value string
		if err := db.QueryRow("SELECT value FROM configdb_note WHERE key='api'").Scan(&value); err != nil || value != "api-updated" {
			t.Fatalf("exported row: %s %v", value, err)
		}
		if err := db.QueryRow("SELECT CAST(value AS TEXT) FROM _alonabot_kv WHERE key='topic'").Scan(&value); err != nil || value != "kv-api" {
			t.Fatalf("exported KV: %s %v", value, err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if !foundDatabase || !foundFile {
		t.Fatal("export omitted database or plugin file")
	}
	api("POST", "clear", `{"confirm":true}`)
	for _, table := range []string{"configdb_note", "configdb_startup"} {
		if raw := api("GET", "tables/"+table+"/rows", ""); !bytes.Contains(raw, []byte(`"total":0`)) {
			t.Fatal(string(raw))
		}
	}
	if raw := api("GET", "kv", ""); !bytes.Contains(raw, []byte(`"total":0`)) {
		t.Fatal(string(raw))
	}
	if raw := api("GET", "files", ""); !bytes.Contains(raw, []byte(`"entries":[]`)) {
		t.Fatal(string(raw))
	}
	if raw := api("GET", "tables/_alonabot_migrations/rows", ""); !bytes.Contains(raw, []byte(`"total":2`)) {
		t.Fatal(string(raw))
	}
	if findPackage(t, c, dir, id).State != "Activated" {
		t.Fatal("data management stopped the plugin")
	}
	stopTestCore(t, c)
}
