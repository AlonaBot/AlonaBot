//go:build unix

// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：用真实的 alonabot 与 SDK 插件进程做端到端测试：两种传输的生命周期、restart、双 Core 互斥、Core 被杀后的恢复和相对 data_dir。依赖 Unix 信号与进程组语义。

package main

import (
	"bufio"
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
			fmt.Fprintf(os.Stderr, "构建 %s 失败: %v\n%s", name, err, out)
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
			c.t.Fatalf("进程已退出，未出现 %q；日志:\n%s", re, c.log())
		case <-deadline:
			c.t.Fatalf("等待 %q 超时；日志:\n%s", re, c.log())
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
		c.t.Fatalf("alonabot 未在 %v 内退出；日志:\n%s", timeout, c.log())
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
const activatedRe = `已激活 generation=(\S+) mode=(\w+) pid=(\d+)`

// 完整生命周期：两种传输都能走完 OnStart → OnReady → 激活 → SIGTERM → STOPPING/OnStop → 正常退出。
func TestLifecycleOverTCPAndStdio(t *testing.T) {
	for _, transport := range []string{"tcp", "stdio"} {
		t.Run(transport, func(t *testing.T) {
			var extra []string
			if transport == "stdio" {
				extra = []string{"--stdio"}
			}
			c := startCore(t, t.TempDir(), extra...)
			m := c.waitFor(activatedRe)
			pid, _ := strconv.Atoi(m[3])
			if m[2] != "new" {
				t.Fatalf("首次运行应提交新代次，实际 mode=%s", m[2])
			}
			_ = c.cmd.Process.Signal(syscall.SIGTERM)
			if code := c.wait(15 * time.Second); code != 0 {
				t.Fatalf("alonabot 退出码 %d；日志:\n%s", code, c.log())
			}
			log := c.log()
			for _, want := range []string{"lifecycle: OnStart", "lifecycle: OnReady", "lifecycle: OnStop", "lifecycle: 正常退出", "强杀=false"} {
				if !strings.Contains(log, want) {
					t.Fatalf("日志缺少 %q:\n%s", want, log)
				}
			}
			if !pluginExited(pid, 2*time.Second) {
				t.Fatalf("插件进程 %d 未退出", pid)
			}
		})
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
		t.Fatalf("应以 restart 复用 %s，实际 generation=%s mode=%s", gen, m[1], m[2])
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
		t.Fatal("第二个 Core 不应启动")
	}
	if !strings.Contains(c2.log(), "已被另一个 Core 锁定") {
		t.Fatalf("应报告 data_dir 已锁定:\n%s", c2.log())
	}
	_ = c1.cmd.Process.Signal(syscall.SIGTERM)
	if code := c1.wait(15 * time.Second); code != 0 {
		t.Fatalf("第一个 Core 应正常停止: %d", code)
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
		t.Fatalf("Core 被 kill 后插件 %d 未在断链预算内退出", pid)
	}
	c2 := startCore(t, dir)
	m2 := c2.waitFor(activatedRe)
	if m2[2] != "restart" || m2[1] != m[1] {
		t.Fatalf("恢复后应复用原代次: %v", m2)
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
					t.Fatalf("第 %d 次运行应为 mode=%s，实际 %s", i+1, mode, m[2])
				}
				_ = c.cmd.Process.Signal(syscall.SIGTERM)
				if code := c.wait(15 * time.Second); code != 0 {
					t.Fatalf("alonabot 退出码 %d；日志:\n%s", code, c.log())
				}
			}
			if _, err := os.Stat(filepath.Join("data", "plugins", "demo", "runtime.lock")); err != nil {
				t.Fatalf("runtime.lock 应保留: %v", err)
			}
		})
	}
}
