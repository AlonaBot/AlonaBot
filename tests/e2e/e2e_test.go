//go:build unix

package main

import (
	"bufio"
	"bytes"
	"context"
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
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
)

// binDir 存放 TestMain 构建的 Core 与真实插件二进制。
var binDir string

// TestMain 先构建 Core 与集成插件，再运行测试，结束后删除构建目录。
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "alonabot-e2e-")
	if err != nil {
		panic(err)
	}
	binDir = dir
	loadComponentDirectories()
	buildCLI := exec.Command("go", "build", "-o", filepath.Join(dir, "alonabot"), "./cmd/alonabot")
	buildCLI.Dir = "../.."
	if out, err := buildCLI.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build CLI: %v\n%s", err, out)
		os.Exit(1)
	}
	build, err := pluginkit.BuildDev(context.Background(), "../../examples/plugins/lifecycle", dir)
	if err != nil {
		panic(err)
	}
	if err := os.Rename(build.Executable, filepath.Join(dir, "lifecycle")); err != nil {
		panic(err)
	}

	if err := prepareTestPackages(); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// core 包装一个运行中的 alonabot run 进程，供测试读取日志和发送信号。
type core struct {
	t   *testing.T
	cmd *exec.Cmd
	// mu 保护 lines。
	mu    sync.Mutex
	lines []string // 已读到的 stderr 行
	// added 容量 1，有新行时通知等待者。
	added chan struct{}
	// done 在 stderr 读完且进程被回收后关闭。
	done chan struct{}
}

// startCore 以 dataDir 运行 alonabot run 并收集 stderr；测试结束时杀掉进程并等待回收。
func startCore(t *testing.T, dataDir string, extra ...string) *core {
	return startCoreWithToken(t, dataDir, "", extra...)
}

func startCoreWithToken(t *testing.T, dataDir, token string, extra ...string) *core {
	t.Helper()
	c := startEmptyCore(t, dataDir, extra...)
	if _, err := os.Stat(filepath.Join(dataDir, "packages", "demo")); errors.Is(err, os.ErrNotExist) {
		c.management(t, dataDir, token, "install", filepath.Join(binDir, "demo-package", "package.tar.gz"), "--yes")
		c.management(t, dataDir, token, "enable", "demo")
	}
	return c
}

func startEmptyCore(t *testing.T, dataDir string, extra ...string) *core {
	t.Helper()
	args := append([]string{"run", "--data-dir", dataDir}, extra...)
	c := &core{t: t, cmd: exec.Command(filepath.Join(binDir, "alonabot"), args...), added: make(chan struct{}, 1), done: make(chan struct{})}
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			c.mu.Lock()
			c.lines = append(c.lines, sc.Text())
			c.mu.Unlock()
			select {
			case c.added <- struct{}{}:
			default:
			}
		}
		_ = c.cmd.Wait()
		close(c.done)
	}()
	t.Cleanup(func() {
		_ = c.cmd.Process.Kill()
		<-c.done
	})
	return c
}

func (c *core) address() string {
	m := c.waitFor(`HTTP listener started listen=(\S+)|"msg":"HTTP listener started"[^\n]*"listen":"([^"]+)"`)
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

func (c *core) management(t *testing.T, dir, token string, args ...string) string {
	t.Helper()
	address := c.address()
	if token == "" {
		raw, err := os.ReadFile(filepath.Join(dir, adminauth.FirstTokenFile))
		if err != nil {
			t.Fatal(err)
		}
		token = strings.TrimSpace(string(raw))
	}
	// 特殊 Host 的认证测试仍走真实 CLI；本地转发器只改 Host，不参与认证或管理行为。
	target, _ := url.Parse("http://" + address)
	proxy := httputil.NewSingleHostReverseProxy(target)
	base := settings.Defaults().BaseURL
	if raw, err := os.ReadFile(filepath.Join(dir, settings.FileName)); err == nil {
		var doc struct {
			BaseURL string `json:"base_url"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		if doc.BaseURL != "" {
			base = doc.BaseURL
		}
	}
	baseURL, _ := url.Parse(base)
	director := proxy.Director
	proxy.Director = func(r *http.Request) { director(r); r.Host = baseURL.Host }
	server := httptest.NewServer(proxy)
	defer server.Close()
	command := exec.Command(filepath.Join(binDir, "alonabot"), append(args, "--admin-url", server.URL)...)
	command.Env = append(os.Environ(), "ALONABOT_ADMIN_TOKEN="+token)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("management %v: %v\n%s\nCore:\n%s", args, err, output, c.log())
	}
	return string(output)
}

// log 返回已收集的全部 stderr。
func (c *core) log() string { c.mu.Lock(); defer c.mu.Unlock(); return strings.Join(c.lines, "\n") }

// waitFor 等待日志出现匹配 re 的行，返回第一组子匹配。
func (c *core) waitFor(re string) []string {
	c.t.Helper()
	rx := regexp.MustCompile(re)
	deadline := time.After(20 * time.Second)
	for {
		if m := rx.FindStringSubmatch(c.log()); m != nil {
			return m
		}
		select {
		case <-c.added:
		case <-c.done:
			if m := rx.FindStringSubmatch(c.log()); m != nil {
				return m
			}
			c.t.Fatalf("process exited before %q appeared; log:\n%s", re, c.log())
		case <-deadline:
			c.t.Fatalf("timed out waiting for %q; log:\n%s", re, c.log())
		}
	}
}

// wait 等待 alonabot 退出并返回退出码，超时即终止测试。
func (c *core) wait(timeout time.Duration) int {
	c.t.Helper()
	select {
	case <-c.done:
		return c.cmd.ProcessState.ExitCode()
	case <-time.After(timeout):
		c.t.Fatalf("alonabot did not exit within %v; log:\n%s", timeout, c.log())
		return -1
	}
}

// pluginExited 报告插件进程是否在 within 内消失。
func pluginExited(pid int, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for syscall.Kill(pid, 0) == nil {
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond) // 只用于观察外部进程是否消失，不作为业务判断
	}
	return true
}

// activatedRe 匹配激活日志，子匹配依次为代次、mode（new/restart）与插件 pid。
const activatedRe = `plugin activated plugin_id=demo generation_id=(\S+) mode=(\w+) pid=(\d+)`

// lifecycleLogRe 是一次完整生命周期在控制台格式（CONTRACTS §14.2）中必须出现的行，覆盖 Core 与插件两类来源。
var lifecycleLogRe = []string{
	`(?m)^\[\d{2}:\d{2}:\d{2}\.\d{3}\] \[Core\] \[INFO\] \[[^\]]+/alonabot\.go:\d+\]: Core started data_dir=`,
	`(?m)^\[\d{2}:\d{2}:\d{2}\.\d{3}\] \[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: OnStart$`,
	`(?m)\[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: OnReady$`,
	`(?m)\[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: OnStop$`,
	`(?m)\[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: exiting normally$`,
	`(?m)\[Core\] \[INFO\] \[[^\]]+/assembly\.go:\d+\]: plugin stopped plugin_id=demo drained=true killed=false$`,
}

// 完整生命周期：TCP 走完 OnStart → OnReady → 激活 → SIGTERM → STOPPING/OnStop → 正常退出；
// 日志按控制台格式输出到 stderr（不是终端，不着色），并写入当天的日志文件。
func TestLifecycleOverTCPAndStdio(t *testing.T) {
	for _, transport := range []string{"tcp", "stdio"} {
		t.Run(transport, func(t *testing.T) {
			dir := t.TempDir()
			var c *core
			if transport == "stdio" {
				c = startEmptyCore(t, dir, "../../examples/plugins/lifecycle", "--stdio")
			} else {
				c = startCore(t, dir)
			}
			m := c.waitFor(activatedRe)
			pid, _ := strconv.Atoi(m[3])
			if m[2] != "new" {
				t.Fatalf("first run should commit a new generation, got mode=%s", m[2])
			}
			_ = c.cmd.Process.Signal(syscall.SIGTERM)
			if code := c.wait(15 * time.Second); code != 0 {
				t.Fatalf("alonabot exit code %d; log:\n%s", code, c.log())
			}
			// 运行可能跨过零点，读取全部日志文件。
			names, err := filepath.Glob(filepath.Join(dir, "logs", "alonabot-*.log"))
			if err != nil || len(names) == 0 {
				t.Fatalf("no log file under %s/logs: %v", dir, err)
			}
			var file []byte
			for _, n := range names {
				b, err := os.ReadFile(n)
				if err != nil {
					t.Fatal(err)
				}
				file = append(file, b...)
			}
			for name, log := range map[string]string{"stderr": c.log(), "log file": string(file)} {
				for _, re := range lifecycleLogRe {
					if !regexp.MustCompile(re).MatchString(log) {
						t.Fatalf("%s does not match %s:\n%s", name, re, log)
					}
				}
				if strings.Contains(log, "\x1b[") {
					t.Fatalf("%s should not be colored:\n%s", name, log)
				}
			}
			if !pluginExited(pid, 2*time.Second) {
				t.Fatalf("plugin process %d did not exit", pid)
			}
		})
	}
}

// --log-format json：插件的记录带上 Core 补充的来源字段与 SDK 给出的源码位置（CONTRACTS §14.1）。
func TestJSONLogFormat(t *testing.T) {
	c := startCore(t, t.TempDir(), "--log-format", "json")
	line := c.waitFor(`(?m)^(\{.*"msg":"OnReady".*\})$`)[1]
	var rec struct {
		Level        string
		Msg          string
		Component    string
		PluginID     string `json:"plugin_id"`
		SessionID    string `json:"session_id"`
		GenerationID string `json:"generation_id"`
		Source       struct{ File string }
	}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("%v: %s", err, line)
	}
	if rec.Level != "INFO" || rec.Component != "plugin" || rec.PluginID != "demo" || rec.SessionID == "" || rec.GenerationID == "" ||
		!strings.HasSuffix(rec.Source.File, "lifecycle/main.go") {
		t.Fatalf("unexpected plugin record: %s", line)
	}
	_ = c.cmd.Process.Signal(syscall.SIGTERM)
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("alonabot exit code %d; log:\n%s", code, c.log())
	}
}

// 同一 data_dir 再次运行：复用已提交代次（restart），不产生新 Generation。
func TestRestartReusesGeneration(t *testing.T) {
	dir := t.TempDir()
	c1 := startCore(t, dir)
	gen := c1.waitFor(activatedRe)[1]
	_ = c1.cmd.Process.Signal(syscall.SIGTERM)
	c1.wait(15 * time.Second)

	c2 := startCore(t, dir)
	m := c2.waitFor(activatedRe)
	if m[1] != gen || m[2] != "restart" {
		t.Fatalf("restart should reuse %s, got generation=%s mode=%s", gen, m[1], m[2])
	}
	_ = c2.cmd.Process.Signal(syscall.SIGTERM)
	c2.wait(15 * time.Second)
}

// 同一 data_dir 上第二个 Core 报告已锁定并以非零退出，第一个 Core 不受影响。
func TestSecondCoreSameDataDirRejected(t *testing.T) {
	dir := t.TempDir()
	c1 := startCore(t, dir)
	c1.waitFor(activatedRe)
	c2 := startCore(t, dir)
	if code := c2.wait(10 * time.Second); code == 0 {
		t.Fatal("second Core should not start")
	}
	if !strings.Contains(c2.log(), "locked by another Core") {
		t.Fatalf("expected a data_dir locked report:\n%s", c2.log())
	}
	_ = c1.cmd.Process.Signal(syscall.SIGTERM)
	if code := c1.wait(15 * time.Second); code != 0 {
		t.Fatalf("first Core should stop normally: %d", code)
	}
}

// Core 被 SIGKILL：插件收到断链后执行 OnStop 并退出；随后新的 Core 能取得锁并恢复运行。
func TestCoreKilledPluginExitsAndRecovers(t *testing.T) {
	dir := t.TempDir()
	c1 := startCore(t, dir)
	m := c1.waitFor(activatedRe)
	pid, _ := strconv.Atoi(m[3])
	_ = c1.cmd.Process.Signal(syscall.SIGKILL)
	c1.wait(5 * time.Second)
	if !pluginExited(pid, 6*time.Second) {
		t.Fatalf("plugin %d did not exit within the disconnect budget after Core was killed", pid)
	}
	c2 := startCore(t, dir)
	m2 := c2.waitFor(activatedRe)
	if m2[2] != "restart" || m2[1] != m[1] {
		t.Fatalf("recovery should reuse the original generation: %v", m2)
	}
	_ = c2.cmd.Process.Signal(syscall.SIGTERM)
	c2.wait(15 * time.Second)
}

// 相对 --data-dir：TCP 能启动、停止，并以 restart 再次打开同一目录；runtime.lock 文件保留。
func TestRelativeDataDir(t *testing.T) {
	for _, transport := range []string{"tcp"} {
		t.Run(transport, func(t *testing.T) {
			t.Chdir(t.TempDir())
			for i, mode := range []string{"new", "restart"} {
				c := startCore(t, "data")
				if m := c.waitFor(activatedRe); m[2] != mode {
					t.Fatalf("run %d should be mode=%s, got %s", i+1, mode, m[2])
				}
				_ = c.cmd.Process.Signal(syscall.SIGTERM)
				if code := c.wait(15 * time.Second); code != 0 {
					t.Fatalf("alonabot exit code %d; log:\n%s", code, c.log())
				}
			}
			if _, err := os.Stat(filepath.Join("data", "plugins", "demo", "runtime.lock")); err != nil {
				t.Fatalf("runtime.lock should be kept: %v", err)
			}
		})
	}
}

// 首次设置文件保持缺省值，显式监听覆盖只影响当前进程。
func TestSettingsAndListenOverride(t *testing.T) {
	dir := t.TempDir()
	c := startCore(t, dir, "--listen", "127.0.0.1:0")
	c.waitFor(activatedRe)
	address := c.waitFor(`HTTP listener started listen=(\S+)`)[1]
	request, err := http.NewRequest(http.MethodGet, "http://"+address+"/plugins/absent/unknown", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "127.0.0.1"
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unmapped plugin status = %d", response.StatusCode)
	}
	name := filepath.Join(dir, settings.FileName)
	original, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	mode, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	if mode.Mode().Perm() != 0o600 {
		t.Fatalf("settings mode = %v", mode.Mode())
	}
	loaded, err := settings.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	defaults, err := settings.Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Settings(), defaults.Settings()) {
		t.Fatalf("default settings changed by override: %+v", loaded)
	}
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("exit %d: %s", code, c.log())
	}
	second := startCore(t, dir, "--listen", "127.0.0.1:0")
	second.waitFor(activatedRe)
	if err := second.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := second.wait(15 * time.Second); code != 0 {
		t.Fatalf("exit %d: %s", code, second.log())
	}
	current, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, current) {
		t.Fatal("command line override rewrote settings")
	}
}

// 设置文件中的日志设置可由显式参数覆盖，日志文件开关仍按文件生效。
func TestConfiguredLogOverrides(t *testing.T) {
	dir := t.TempDir()
	original := []byte(`{"listen":"127.0.0.1:0","log":{"level":"warn","format":"json","file":{"enabled":false}}}`)
	name := filepath.Join(dir, settings.FileName)
	if err := os.WriteFile(name, original, 0o600); err != nil {
		t.Fatal(err)
	}
	c := startCore(t, dir, "--log-level", "info", "--log-format", "text")
	c.waitFor(activatedRe)
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("exit %d: %s", code, c.log())
	}
	if _, err := os.Stat(filepath.Join(dir, "logs")); !os.IsNotExist(err) {
		t.Fatalf("disabled file logging created logs: %v", err)
	}
	current, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, current) {
		t.Fatal("log overrides rewrote settings")
	}
	// 未显式给参数时，日志格式取自文件，而非 Cobra 旗标的缺省值。
	configured := bytes.Replace(original, []byte(`"level":"warn"`), []byte(`"level":"info"`), 1)
	if err := os.WriteFile(name, configured, 0o600); err != nil {
		t.Fatal(err)
	}
	fromFile := startCore(t, dir)
	fromFile.waitFor(`(?m)^(\{.*"msg":"OnReady".*\})$`)
	if err := fromFile.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := fromFile.wait(15 * time.Second); code != 0 {
		t.Fatalf("exit %d: %s", code, fromFile.log())
	}

}

// TestUnknownSettingIgnored 验证未知设置保留在原文中，Core 仍可正常启动和停止。
func TestUnknownSettingIgnored(t *testing.T) {
	dir := t.TempDir()
	original := []byte("{\n  \"listen\": \"127.0.0.1:0\",\n  \"log\": {\"unknown\": true}\n}\n")
	name := filepath.Join(dir, settings.FileName)
	if err := os.WriteFile(name, original, 0o600); err != nil {
		t.Fatal(err)
	}
	c := startEmptyCore(t, dir)
	_ = c.address()
	stopTestCore(t, c)
	actual, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, original) {
		t.Fatalf("unknown setting was rewritten: %s", actual)
	}
}

// TestAdminTokenRequiresTerminal 验证非终端命令以 usage 退出且没有初始化凭证。
func TestAdminTokenRequiresTerminal(t *testing.T) {
	for _, action := range []string{"init", "reset"} {
		t.Run(action, func(t *testing.T) {
			dir := t.TempDir()
			command := exec.Command(filepath.Join(binDir, "alonabot"), "admin-token", action, "--data-dir", dir)
			output, err := command.CombinedOutput()
			if err == nil || command.ProcessState.ExitCode() != exitUsage || !strings.Contains(string(output), "stdin and stderr to be terminals") {
				t.Fatalf("non-terminal %s: err=%v, output=%s", action, err, output)
			}
			store, err := state.Open(context.Background(), dir)
			if err != nil {
				t.Fatal(err)
			}
			_, readErr := store.AdminCredential(context.Background())
			closeErr := store.Close()
			if !errors.Is(readErr, state.ErrNotFound) || closeErr != nil {
				t.Fatalf("credential after non-terminal %s: read=%v, close=%v", action, readErr, closeErr)
			}
		})
	}
}

// TestAdminTokenResetWhileCoreRunning 通过真实伪终端越过终端门禁，验证 core.lock 拒绝离线 reset。
func TestAdminTokenResetWhileCoreRunning(t *testing.T) {
	dir := t.TempDir()
	c := startCore(t, dir, "--listen", "127.0.0.1:0")
	c.waitFor(activatedRe)
	binary := filepath.Join(binDir, "alonabot")
	var command *exec.Cmd
	if runtime.GOOS == "linux" {
		// util-linux script 接受单个 shell 命令，用环境传参以保留路径原文。
		command = exec.Command("script", "-q", "-e", "-c", `exec "$ALONABOT_TEST_BINARY" admin-token reset --data-dir "$ALONABOT_TEST_DATA_DIR"`, os.DevNull)
		command.Env = append(os.Environ(), "ALONABOT_TEST_BINARY="+binary, "ALONABOT_TEST_DATA_DIR="+dir)
	} else {
		command = exec.Command("script", "-q", os.DevNull, binary, "admin-token", "reset", "--data-dir", dir)
	}
	output, err := command.CombinedOutput()
	if err == nil || command.ProcessState.ExitCode() != exitFailure || !strings.Contains(string(output), state.ErrDataDirLocked.Error()) || !strings.Contains(string(output), "stop Core") {
		t.Fatalf("reset with Core holding core.lock: err=%v, output=%s", err, output)
	}
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit %d: %s", code, c.log())
	}
}

// TestAdminAuthRun 通过真实 Core 公开入口验证管理认证装配与 Host 隔离。
func TestAdminAuthRun(t *testing.T) {
	for _, initialized := range []bool{false, true} {
		t.Run(fmt.Sprintf("initialized=%t", initialized), func(t *testing.T) {
			dir := t.TempDir()
			var token string
			var credentials []string
			if initialized {
				store, err := state.Open(context.Background(), dir)
				if err != nil {
					t.Fatal(err)
				}
				token, err = adminauth.InitToken(context.Background(), store)
				closeErr := store.Close()
				if err != nil || closeErr != nil {
					t.Fatalf("prepare credential: %v; close: %v", err, closeErr)
				}
			}
			const origin = "http://admin.test:7650"
			if err := os.WriteFile(filepath.Join(dir, settings.FileName), []byte(`{"listen":"127.0.0.1:0","base_url":"`+origin+`"}`), 0o600); err != nil {
				t.Fatal(err)
			}
			c := startCoreWithToken(t, dir, token)
			c.waitFor(activatedRe)
			address := c.waitFor(`HTTP listener started listen=(\S+)`)[1]
			send := func(method, path, host, source, body string, cookie *http.Cookie, csrf, bearer string, want int) (*http.Response, []byte) {
				t.Helper()
				req, err := http.NewRequest(method, "http://"+address+path, strings.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				req.Host = host
				req.Header.Set("Content-Type", "application/json")
				if source != "" {
					req.Header.Set("Origin", source)
				}
				if cookie != nil {
					req.AddCookie(cookie)
				}
				if csrf != "" {
					req.Header.Set("X-Alonabot-CSRF", csrf)
				}
				if bearer != "" {
					req.Header.Set("Authorization", "Bearer "+bearer)
				}
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				content, readErr := io.ReadAll(response.Body)
				closeErr := response.Body.Close()
				if readErr != nil || closeErr != nil {
					t.Fatalf("read HTTP response: %v; close: %v", readErr, closeErr)
				}
				if response.StatusCode != want {
					t.Fatalf("%s %s Host=%s: got %d want %d", method, path, host, response.StatusCode, want)
				}
				return response, content
			}
			send("POST", "/api/auth/login", "plugin.test", origin, `{"token":"wrong"}`, nil, "", "", 421)
			send("POST", "/api/auth/login", "unknown.test", origin, `{"token":"wrong"}`, nil, "", "", 421)
			if !initialized {
				_, content := send("POST", "/api/auth/login", "admin.test", origin, `{"token":"wrong"}`, nil, "", "", 401)
				if strings.TrimSpace(string(content)) != "unauthorized" {
					t.Fatal("automatically initialized login did not reject an invalid token")
				}
				send("GET", "/api/probe", "admin.test", "", "", nil, "", "", 401)
			} else {
				send("POST", "/api/auth/login", "admin.test", origin, `{"token":"wrong"}`, nil, "", "", 401)
				body, err := json.Marshal(map[string]string{"token": token})
				if err != nil {
					t.Fatal(err)
				}
				send("POST", "/api/auth/login", "admin.test", "http://admin.test:7651", string(body), nil, "", "", 403)
				response, content := send("POST", "/api/auth/login", "admin.test", origin, string(body), nil, "", "", 200)
				var reply struct {
					CSRFToken string `json:"csrf_token"`
				}
				if err := json.Unmarshal(content, &reply); err != nil {
					t.Fatal(err)
				}
				cookies := response.Cookies()
				if len(cookies) != 1 || reply.CSRFToken == "" {
					t.Fatal("login did not return one cookie and CSRF token")
				}
				cookie := cookies[0]
				// 业务路由尚未注册；通过认证后到达404，未认证时被入口拒绝。
				send("GET", "/api/probe", "admin.test", "", "", nil, "", "", 401)
				send("GET", "/api/probe", "admin.test", "", "", cookie, "", "", 404)
				send("GET", "/api/probe", "admin.test", "", "", nil, "", token, 404)
				send("POST", "/api/probe", "admin.test", origin, "", cookie, "", "", 401)
				send("POST", "/api/probe", "admin.test", origin, "", cookie, reply.CSRFToken, "", 404)
				send("POST", "/api/auth/logout", "admin.test", origin, "", cookie, reply.CSRFToken, "", 204)
				send("GET", "/api/probe", "admin.test", "", "", cookie, "", "", 401)
				credentials = []string{token, cookie.Value, reply.CSRFToken}
			}
			if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			if code := c.wait(15 * time.Second); code != 0 {
				t.Fatalf("Core exit %d: %s", code, c.log())
			}
			for _, credential := range credentials {
				if strings.Contains(c.log(), credential) {
					t.Fatal("Core log contains management credentials")
				}
			}
		})
	}
}

// TestPluginPageSandbox 通过真实 Core 与 lifecycle 子进程验证插件页面隔离和管理授权。
func TestPluginPageSandbox(t *testing.T) {
	// 1. 在单一受控主机上启动 Core 与 lifecycle 插件。
	dir := t.TempDir()
	const origin = "http://core.test:7650"
	if err := os.WriteFile(filepath.Join(dir, settings.FileName), []byte(`{"listen":"127.0.0.1:0","base_url":"`+origin+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c := startCore(t, dir)
	c.waitFor(activatedRe)
	address := c.waitFor(`HTTP listener started listen=(\S+)`)[1]
	send := func(method, path, host string, headers http.Header, cookie *http.Cookie, body string, want int) (*http.Response, []byte) {
		t.Helper()
		r, err := http.NewRequest(method, "http://"+address+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		r.Host = host
		r.Header = headers.Clone()
		if cookie != nil {
			r.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		content, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read response: %v; close: %v", readErr, closeErr)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s Host=%s: status=%d, want %d", method, path, host, response.StatusCode, want)
		}
		if strings.HasPrefix(path, "/plugins/") {
			if got := response.Header.Values("Content-Security-Policy"); !reflect.DeepEqual(got, []string{"sandbox allow-scripts allow-forms allow-popups"}) {
				t.Errorf("sandbox CSP = %v", got)
			}
			if got := response.Header.Values("X-Content-Type-Options"); !reflect.DeepEqual(got, []string{"nosniff"}) {
				t.Errorf("nosniff = %v", got)
			}
			if got := response.Header.Values("Set-Cookie"); len(got) != 0 {
				t.Error("plugin response contains Set-Cookie")
			}
		}
		return response, content
	}
	pageHeaders := http.Header{"Accept": {"text/html"}}
	send(http.MethodGet, "/plugins/demo/page", "core.test", pageHeaders, nil, "", http.StatusUnauthorized)
	// 2. 经真实登录取得 Cookie 和 CSRF token，核对页面的响应头被 Core 覆盖。
	raw, err := os.ReadFile(filepath.Join(dir, adminauth.FirstTokenFile))
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimSpace(string(raw))
	input, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		t.Fatal(err)
	}
	response, content := send(http.MethodPost, "/api/auth/login", "core.test", http.Header{"Content-Type": {"application/json"}, "Origin": {origin}}, nil, string(input), http.StatusOK)
	var login struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(content, &login); err != nil {
		t.Fatal(err)
	}
	if len(response.Cookies()) != 1 || login.CSRFToken == "" {
		t.Fatal("login did not return one cookie and a CSRF token")
	}
	cookie := response.Cookies()[0]
	page, body := send(http.MethodGet, "/plugins/demo/page", "core.test", pageHeaders, cookie, "", http.StatusOK)
	if !strings.HasPrefix(page.Header.Get("Content-Type"), "text/html") || !bytes.Contains(body, []byte("<!doctype html>")) {
		t.Fatal("plugin HTML response was replaced")
	}
	callbackState := page.Header.Get("X-Lifecycle-Callback-State")
	if callbackState == "" {
		t.Fatal("private page did not provide callback state")
	}
	pageHeaders.Set("Authorization", "Bearer "+token)
	send(http.MethodGet, "/plugins/demo/page", "core.test", pageHeaders, nil, "", http.StatusOK)
	// 3. Cookie 写请求验证 CSRF，回显证明管理凭证未进入插件；Bearer 独立授权。
	headers := http.Header{"Origin": {origin}, "X-Sandbox-Probe": {"preserved"}}
	send(http.MethodPost, "/plugins/demo/echo", "core.test", headers, cookie, "", http.StatusUnauthorized)
	headers.Set("X-Alonabot-CSRF", login.CSRFToken)
	for _, bearer := range []bool{false, true} {
		if bearer {
			headers.Set("Authorization", "Bearer "+token)
		}
		_, content := send(http.MethodPost, "/plugins/demo/echo", "core.test", headers, cookie, "", http.StatusOK)
		var received http.Header
		if err := json.Unmarshal(content, &received); err != nil {
			t.Fatal(err)
		}
		if received.Get("X-Sandbox-Probe") != "preserved" {
			t.Fatal("plugin did not echo the forwarded request")
		}
		for _, name := range []string{"Cookie", "Authorization", "X-Alonabot-CSRF"} {
			if received.Get(name) != "" {
				t.Errorf("plugin received management header %s", name)
			}
		}
	}
	send(http.MethodPost, "/plugins/demo/callback", "core.test", http.Header{"X-Lifecycle-Callback-State": {callbackState}}, nil, "", http.StatusNoContent)
	send(http.MethodGet, "/plugins/demo/page", "elsewhere.test", pageHeaders, nil, "", http.StatusMisdirectedRequest)
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit %d: %s", code, c.log())
	}
}

// TestFirstAdminToken 联验非终端首次启动、show 读取与登录后删除凭证文件（V-47、V-88）。
func TestFirstAdminToken(t *testing.T) {
	dir := t.TempDir()
	c := startCore(t, dir, "--listen", "127.0.0.1:0")
	c.waitFor(activatedRe)
	address := c.waitFor(`HTTP listener started listen=(\S+)`)[1]
	name := filepath.Join(dir, adminauth.FirstTokenFile)
	stat, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm() != 0o600 {
		t.Fatalf("first token file mode = %v", stat.Mode())
	}

	binary := filepath.Join(binDir, "alonabot")
	command := exec.Command(binary, "admin-token", "show", "--data-dir", dir)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		t.Fatalf("show first token: %v; stderr=%s", err, stderr.String())
	}
	token := strings.TrimSpace(string(output))
	if token == "" {
		t.Fatal("show returned an empty token")
	}
	input, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, "http://"+address+"/api/auth/login", bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "127.0.0.1"
	request.Header.Set("Origin", "http://127.0.0.1:7650")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	content, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Fatalf("read login response: %v; close: %v", readErr, closeErr)
	}
	if response.StatusCode != http.StatusOK || len(response.Cookies()) != 1 {
		t.Fatalf("login: status=%d body=%s", response.StatusCode, content)
	}
	if _, err := os.Stat(name); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("first token file after login: %v", err)
	}

	command = exec.Command(binary, "admin-token", "show", "--data-dir", dir)
	output, err = command.CombinedOutput()
	if err == nil || command.ProcessState.ExitCode() != exitFailure || !strings.Contains(string(output), "read first admin token file") {
		t.Fatalf("show after login: err=%v output=%s", err, output)
	}
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit %d: %s", code, c.log())
	}
}
