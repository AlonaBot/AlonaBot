//go:build unix

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDevelopmentExample 验证真实示例的翻译热部署与开发、安装组件混合运行的连续性。
func TestDevelopmentExample(t *testing.T) {
	plugin, adapter, dir := t13Example(t)
	c := startEmptyCore(t, dir, plugin, adapter, "--reload", "--log-format", "json", "--args", "DEV_RELOAD_POLL=50ms")
	pluginPID := t13PID(t, c, "alonabot-plugin-devbot")
	adapterPID := t13PID(t, c, "alonabot-adapter-devbot")
	corePID := c.cmd.Process.Pid
	adapterSession := t13AdapterSession(t, c)
	t13Reply(t, c, dir, "/hello", "Hello from devbot!")

	// 1. 通过管理 CLI 启用真实签名包，保留其进程与 Session 的观察依据。
	c.management(t, dir, "", "install", filepath.Join(binDir, "demo-package", "package.tar.gz"), "--yes")
	c.management(t, dir, "", "enable", "demo")
	installedPID := t13PID(t, c, "demo")
	installedSession := lastSession(t, c)
	pageState := t13Page(t, c, dir)
	t13Reply(t, c, dir, "/ping", "pong")

	// 2. 只改翻译资源，回复与开发插件 PID 必须变化，其他组件保持连续。
	path := filepath.Join(plugin, "i18n", "en-US.json")
	developmentWrite(t, path, strings.ReplaceAll(string(developmentRead(t, path)), "Hello from devbot!", "Reloaded translation!"))
	t13Reply(t, c, dir, "/hello", "Reloaded translation!")
	if t13PID(t, c, "alonabot-plugin-devbot") == pluginPID {
		t.Fatal("translation change did not replace plugin process")
	}
	if c.cmd.Process.Pid != corePID || t13PID(t, c, "alonabot-adapter-devbot") != adapterPID || t13PID(t, c, "demo") != installedPID {
		t.Fatalf("translation change restarted an unrelated process:\n%s", c.log())
	}
	if t13AdapterSession(t, c) != adapterSession || strings.Count(c.log(), "loopback instance started") != 1 {
		t.Fatalf("Adapter Session or instance restarted:\n%s", c.log())
	}
	if got := t13Page(t, c, dir); got != pageState || lastSession(t, c) != installedSession {
		t.Fatal("installed plugin Session changed")
	}
	t13Reply(t, c, dir, "/ping", "pong")
	t.Logf("Core PID=%d; Adapter PID=%d Session=%s; installed PID=%d Session=%s; translation plugin PID=%d -> %d", corePID, adapterPID, adapterSession, installedPID, installedSession, pluginPID, t13PID(t, c, "alonabot-plugin-devbot"))
	stopTestCore(t, c)
}

func t13Example(t *testing.T) (string, string, string) {
	t.Helper()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	project, dir := t.TempDir(), t.TempDir()
	source := filepath.Join(repository, "examples", "devbot")
	if err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(project, relative), 0o700)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(project, relative), raw, 0o600)
	}); err != nil {
		t.Fatal(err)
	}
	developmentWrite(t, filepath.Join(project, "go.mod"), fmt.Sprintf("module example.com/devbot\n\ngo 1.27.1\n\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => %s\n", filepath.ToSlash(repository)))
	// 薄入口从副本构建；依赖仍使用仓库已锁定的版本。
	developmentWrite(t, filepath.Join(project, "main.go"), strings.ReplaceAll(string(developmentRead(t, filepath.Join(project, "main.go"))), "github.com/AlonaBot/AlonaBot/examples/devbot/", "example.com/devbot/"))
	command := exec.CommandContext(t.Context(), "go", "mod", "tidy")
	command.Dir = project
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("tidy example: %v\n%s", err, output)
	}
	output := filepath.Join(dir, "sent-one")
	t.Setenv("DEVBOT_OUTPUT", output)
	settings := string(developmentRead(t, filepath.Join(project, "core_config.example.json")))
	developmentWrite(t, filepath.Join(dir, "core_config.json"), strings.ReplaceAll(settings, `"127.0.0.1:7650"`, `"127.0.0.1:0"`))
	return filepath.Join(project, "plugin"), filepath.Join(project, "adapter"), dir
}

func t13Request(t *testing.T, c *core, dir, method, path string) (http.Header, int, []byte) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), method, "http://"+c.address()+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "127.0.0.1:7650"
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(developmentRead(t, filepath.Join(dir, "admin_token")))))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Fatalf("read example response: %v %v", readErr, closeErr)
	}
	return response.Header, response.StatusCode, body
}

func t13Reply(t *testing.T, c *core, dir, text, want string) {
	t.Helper()
	path := filepath.Join(dir, "sent-one")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for {
		_, status, body := t13Request(t, c, dir, http.MethodPost, "/plugins/alonabot-adapter-devbot/inject?instance=one&text="+url.QueryEscape(text))
		if status != http.StatusNoContent {
			t.Fatalf("inject: %d %s\n%s", status, body, c.log())
		}
		time.Sleep(20 * time.Millisecond)
		raw, err := os.ReadFile(path)
		if err == nil && string(raw) == want {
			return
		}
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatalf("reply = %q, want %q\n%s", raw, want, c.log())
		}
	}
}

func t13PID(t *testing.T, c *core, id string) int {
	t.Helper()
	c.waitFor(`"msg":"plugin activated"[^\n]*"plugin_id":"` + id + `"`)
	var pid int
	for _, line := range strings.Split(c.log(), "\n") {
		var record struct {
			Message  string `json:"msg"`
			PluginID string `json:"plugin_id"`
			PID      int    `json:"pid"`
		}
		if json.Unmarshal([]byte(line), &record) == nil && record.Message == "plugin activated" && record.PluginID == id {
			pid = record.PID
		}
	}
	if pid == 0 {
		t.Fatalf("plugin %s has no PID", id)
	}
	return pid
}

func t13AdapterSession(t *testing.T, c *core) string {
	t.Helper()
	c.waitFor(`"msg":"loopback instance started"`)
	var session string
	for _, line := range strings.Split(c.log(), "\n") {
		var record struct {
			Message string `json:"msg"`
			Session string `json:"session_id"`
		}
		if json.Unmarshal([]byte(line), &record) != nil || record.Message != "loopback instance started" {
			continue
		}
		if record.Session == "" || (session != "" && session != record.Session) {
			t.Fatalf("Adapter Session missing or changed: %s", line)
		}
		session = record.Session
	}
	return session
}

func t13Page(t *testing.T, c *core, dir string) string {
	t.Helper()
	header, status, body := t13Request(t, c, dir, http.MethodGet, "/plugins/demo/page")
	state := header.Get("X-Lifecycle-Callback-State")
	if status != http.StatusOK || !strings.Contains(string(body), "Lifecycle") || state == "" {
		t.Fatalf("installed page: %d %s %v", status, body, header)
	}
	return state
}
