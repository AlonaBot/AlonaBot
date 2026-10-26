//go:build unix

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// binDir 存放 TestMain 构建的 alonabot 与 lifecycle 二进制。
var binDir string

// TestMain 先构建 alonabot 与 examples/plugins/lifecycle，再运行测试，结束后删除构建目录。
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "alonabot-e2e-")
	if err != nil {
		panic(err)
	}
	binDir = dir
	for name, pkg := range map[string]string{"alonabot": ".", "lifecycle": "../../examples/plugins/lifecycle"} {
		if out, err := exec.Command("go", "build", "-o", filepath.Join(dir, name), pkg).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "build %s failed: %v\n%s", name, err, out)
			os.Exit(1)
		}
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
	t.Helper()
	args := append([]string{"run", "--data-dir", dataDir, "--plugin-id", "demo", "--entry", filepath.Join(binDir, "lifecycle")}, extra...)
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
	`(?m)^\[\d{2}:\d{2}:\d{2}\.\d{3}\] \[Core\] \[INFO\] \[alonabot/main\.go:\d+\]: Core started data_dir=`,
	`(?m)^\[\d{2}:\d{2}:\d{2}\.\d{3}\] \[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: OnStart$`,
	`(?m)\[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: OnReady$`,
	`(?m)\[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: OnStop$`,
	`(?m)\[Plug:demo\] \[INFO\] \[lifecycle/main\.go:\d+\]: exiting normally$`,
	`(?m)\[Core\] \[INFO\] \[alonabot/main\.go:\d+\]: plugin stopped plugin_id=demo drained=true killed=false$`,
}

// 完整生命周期：两种传输都能走完 OnStart → OnReady → 激活 → SIGTERM → STOPPING/OnStop → 正常退出；
// 日志按控制台格式输出到 stderr（不是终端，不着色），并写入当天的日志文件。
func TestLifecycleOverTCPAndStdio(t *testing.T) {
	for _, transport := range []string{"tcp", "stdio"} {
		t.Run(transport, func(t *testing.T) {
			var extra []string
			if transport == "stdio" {
				extra = []string{"--stdio"}
			}
			dir := t.TempDir()
			c := startCore(t, dir, extra...)
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

// 相对 --data-dir：两种传输下都能启动、停止，并以 restart 再次打开同一目录；runtime.lock 文件保留。
func TestRelativeDataDir(t *testing.T) {
	for _, transport := range []string{"tcp", "stdio"} {
		t.Run(transport, func(t *testing.T) {
			t.Chdir(t.TempDir())
			var extra []string
			if transport == "stdio" {
				extra = []string{"--stdio"}
			}
			for i, mode := range []string{"new", "restart"} {
				c := startCore(t, "data", extra...)
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
