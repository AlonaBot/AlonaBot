package process

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// helperEnv 选择测试二进制作为辅助进程时的行为。
const helperEnv = "PROCESS_TEST_HELPER"

// TestMain 在设置 helperEnv 时作为辅助进程运行：由测试二进制重入执行，模拟插件进程。
func TestMain(m *testing.M) {
	switch os.Getenv(helperEnv) {
	case "":
		os.Exit(m.Run())
	case "env":
		for _, kv := range os.Environ() {
			fmt.Println(kv)
		}
	case "log":
		fmt.Fprintln(os.Stdout, "hello from stdout")
		fmt.Fprintln(os.Stderr, "hello from stderr")
		fmt.Fprint(os.Stderr, strings.Repeat("x", 100<<10)+"\n") // 超过 64 KiB 的长行
	case "exit3":
		os.Exit(3)
	case "spawn":
		// 启动一个留在同一进程组的孙进程，报告其 pid 后挂起。
		child := exec.Command(os.Args[0], "-test.run=^$")
		child.Env = append(os.Environ(), helperEnv+"=sleep")
		if err := child.Start(); err != nil {
			fmt.Println("error", err)
			os.Exit(2)
		}
		fmt.Println("grandchild", child.Process.Pid)
		time.Sleep(time.Hour) // 不能用 select{}：没有其他 goroutine 时运行时会判定死锁并退出
	case "sleep":
		time.Sleep(time.Hour)
	case "stdio-echo":
		r := bufio.NewReader(os.Stdin)
		line, _ := r.ReadString('\n')
		fmt.Print("echo:" + line)
		fmt.Fprintln(os.Stderr, "stdio log")
	}
	os.Exit(0)
}

// logSink 收集 LogFunc 输出的日志行。
type logSink struct {
	// mu 保护 lines。
	mu    sync.Mutex
	lines []string // 「插件ID|流|内容」形式的日志行
}

// log 记录一行日志，签名与 LogFunc 相同。
func (l *logSink) log(pluginID, stream string, line []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, pluginID+"|"+stream+"|"+string(line))
}

// all 返回已收集日志的副本。
func (l *logSink) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.lines...)
}

// spec 构造以 mode 运行辅助进程的启动描述，extra 叠加到会话变量中。
func spec(mode string, extra map[string]string) Spec {
	env := map[string]string{helperEnv: mode}
	for k, v := range extra {
		env[k] = v
	}
	return Spec{PluginID: "demo", Path: os.Args[0], Args: []string{"-test.run=^$"}, Env: env}
}

// 插件只收到白名单变量与会话变量，Core 的管理凭证与其他插件的凭证不会泄漏。
func TestMinimalEnvWithoutAdminCredentials(t *testing.T) {
	t.Setenv("ALONABOT_ADMIN_TOKEN", "must-not-leak")
	t.Setenv("OTHER_PLUGIN_SECRET", "must-not-leak")
	secret, err := NewSecret()
	if err != nil || len(secret) != 64 {
		t.Fatalf("Secret 应为 32 字节十六进制: %q %v", secret, err)
	}
	logs := &logSink{}
	p, err := Start(spec("env", map[string]string{wire.EnvPluginID: "demo", wire.EnvSecret: secret}), logs.log)
	if err != nil {
		t.Fatal(err)
	}
	<-p.Exited()
	p.WaitLogs()
	env := map[string]string{}
	for _, l := range logs.all() {
		parts := strings.SplitN(l, "|", 3)
		if kv := strings.SplitN(parts[2], "=", 2); len(kv) == 2 {
			env[kv[0]] = kv[1]
		}
	}
	if _, ok := env["ALONABOT_ADMIN_TOKEN"]; ok {
		t.Fatal("Core 管理凭证进入了插件环境")
	}
	if _, ok := env["OTHER_PLUGIN_SECRET"]; ok {
		t.Fatal("白名单之外的变量进入了插件环境")
	}
	if env[wire.EnvPluginID] != "demo" || env[wire.EnvSecret] != secret {
		t.Fatalf("会话变量缺失: %v", env)
	}
	for k := range env {
		allowed := k == helperEnv || strings.HasPrefix(k, "ALONABOT_")
		for _, w := range inheritedEnv {
			allowed = allowed || k == w
		}
		// 部分平台的运行时会自动补充极少数变量（如 Windows 的 SYSTEMROOT），这里只在 Unix 上严格检查。
		if !allowed && runtime.GOOS != "windows" {
			t.Errorf("意外的环境变量 %s", k)
		}
	}
}

// stdout/stderr 日志带插件 ID 前缀，超过 64 KiB 的行分块完整转发。
func TestLogsTaggedWithPluginID(t *testing.T) {
	logs := &logSink{}
	p, err := Start(spec("log", nil), logs.log)
	if err != nil {
		t.Fatal(err)
	}
	<-p.Exited()
	p.WaitLogs()
	got := logs.all()
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "demo|stdout|hello from stdout") || !strings.Contains(joined, "demo|stderr|hello from stderr") {
		t.Fatalf("日志应带插件 ID 与来源流: %q", got)
	}
	total := 0
	for _, l := range got {
		if strings.HasPrefix(l, "demo|stderr|x") {
			total += len(l) - len("demo|stderr|")
		}
	}
	if total != 100<<10 {
		t.Fatalf("超长行应分块完整转发，收到 %d 字节", total)
	}
}

// 退出码如实记录在 ExitState 中。
func TestExitStatus(t *testing.T) {
	p, err := Start(spec("exit3", nil), (&logSink{}).log)
	if err != nil {
		t.Fatal(err)
	}
	<-p.Exited()
	if st := p.ExitState(); st == nil || st.ExitCode() != 3 {
		t.Fatalf("退出码应为 3，实际 %v", st)
	}
}

// stdio 模式下 stdin/stdout 是协议管道、不进入日志，stderr 进入日志；Release 关闭管道端且可重复调用。
func TestStdioTransportPipes(t *testing.T) {
	s := spec("stdio-echo", nil)
	s.Stdio = true
	logs := &logSink{}
	p, err := Start(s, logs.log)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Stdin.Write([]byte("ping\n")); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(p.Stdout).ReadString('\n')
	if err != nil || line != "echo:ping\n" {
		t.Fatalf("stdio 管道往返: %q %v", line, err)
	}
	<-p.Exited()
	p.WaitLogs()
	for _, l := range logs.all() {
		if strings.Contains(l, "echo:") {
			t.Fatal("stdio 模式下 stdout 是协议通道，不应进入日志")
		}
	}
	if !strings.Contains(strings.Join(logs.all(), "\n"), "demo|stderr|stdio log") {
		t.Fatal("stdio 模式下 stderr 应进入日志")
	}
	if err := p.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if _, err := p.Stdin.Write([]byte("x")); err == nil {
		t.Fatal("Release 应关闭 Core 持有的 stdio 管道端")
	}
	if err := p.Release(); err != nil {
		t.Fatalf("重复 Release 不应报错: %v", err)
	}
}

// 入口必须是包内存在的可执行普通文件，越界或绝对路径被拒绝。
func TestEntrypoint(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin", "demo")
	_ = os.MkdirAll(filepath.Dir(bin), 0o755)
	_ = os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755)
	if p, err := Entrypoint(dir, "bin/demo"); err != nil || p != bin {
		t.Fatalf("Entrypoint = %q, %v", p, err)
	}
	for _, bad := range []string{"../demo", "/bin/demo", "bin", "missing"} {
		if _, err := Entrypoint(dir, bad); err == nil {
			t.Errorf("%q 应被拒绝", bad)
		}
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(bin, 0o644)
		if _, err := Entrypoint(dir, "bin/demo"); err == nil {
			t.Error("不可执行的文件应被拒绝")
		}
	}
}

// Core 临时持有 runtime.lock 时，其他持有者（例如 SDK）取锁失败；释放后锁文件仍在。
func TestLockRuntime(t *testing.T) {
	dir := t.TempDir()
	l, err := LockRuntime(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := oslock.TryLock(filepath.Join(dir, wire.RuntimeLockFile)); !errors.Is(err, oslock.ErrLocked) {
		t.Fatalf("Core 持锁期间其他持有者应失败，实际 %v", err)
	}
	_ = l.Unlock()
	if _, err := os.Stat(filepath.Join(dir, wire.RuntimeLockFile)); err != nil {
		t.Fatal("锁文件必须保留")
	}
}
