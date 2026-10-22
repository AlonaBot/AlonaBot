//go:build unix

// 本文件用真实子进程测试插件的启动、停止、崩溃恢复、启动故障、runtime.lock 与管理锁串行化。
// 依赖 Unix 信号与进程组语义。

package supervisor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// faultEnv 选择辅助进程注入的故障。
const faultEnv = "SUPERVISOR_TEST_FAULT"

// 辅助进程：用原始协议帧扮演插件，按 SUPERVISOR_TEST_FAULT 注入故障。
func TestMain(m *testing.M) {
	fault, ok := os.LookupEnv(faultEnv)
	if !ok {
		os.Exit(m.Run())
	}
	os.Exit(rawPlugin(fault))
}

// rawPlugin 用原始协议帧扮演插件走完握手、注册与激活，按 fault 在相应位置注入故障，返回进程退出码。
func rawPlugin(fault string) int {
	if fault == "exit-early" {
		return 4
	}
	var conn *wire.Conn
	if os.Getenv(wire.EnvTransport) == wire.TransportStdio {
		if fault == "pollute" {
			fmt.Fprintln(os.Stdout, "plugin started, hello world") // 违规：stdout 混入普通日志
		}
		conn = wire.NewConn(os.Stdin, os.Stdout, os.Stdout)
	} else {
		c, err := net.Dial("tcp", os.Getenv(wire.EnvCoreAddr))
		if err != nil {
			return 5
		}
		conn = wire.NewStreamConn(c)
	}
	switch fault {
	case "hang":
		time.Sleep(time.Hour)
	case "badframe":
		_ = conn.WriteFrame(&pb.Envelope{ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_REQUEST})
		time.Sleep(time.Hour)
	}
	id := os.Getenv(wire.EnvPluginID)
	secret, _ := hex.DecodeString(os.Getenv(wire.EnvSecret))
	digest, _ := hex.DecodeString(os.Getenv(wire.EnvBuildDigest))
	hello, _ := wire.ControlEnvelope(nil, 0, id, pb.FrameKind_FRAME_KIND_HELLO,
		&pb.Hello{PluginId: id, Secret: secret, SupportedMajor: 1, BuildDigest: digest})
	_ = conn.WriteFrame(hello)
	recv := func() (*pb.Envelope, any) {
		env, err := conn.ReadFrame()
		if err != nil {
			os.Exit(6)
		}
		msg, _ := validate.DecodeControl(env)
		return env, msg
	}
	_, msg := recv()
	ack := msg.(*pb.HelloAck)
	sid := ack.GetSessionId()
	reg, _ := wire.ControlEnvelope(sid, 0, id, pb.FrameKind_FRAME_KIND_REGISTER, &pb.Register{CandidateConfigJson: "{}"})
	_ = conn.WriteFrame(reg)
	recv() // PREPARED
	_, msg = recv()
	act := msg.(*pb.Activate)
	done, _ := wire.ControlEnvelope(sid, 0, id, pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: act.GetGenerationId(), Ready: true})
	_ = conn.WriteFrame(done)
	// orphan-*：先启动一个留在进程组内的子进程（pid 写到 stderr），再崩溃或在 STOPPING 后正常退出。
	if strings.HasPrefix(fault, "orphan-") {
		child := exec.Command("/bin/sleep", "600")
		if err := child.Start(); err != nil {
			return 7
		}
		fmt.Fprintf(os.Stderr, "orphan %d\n", child.Process.Pid)
	}
	if fault == "crash" || fault == "orphan-crash" {
		return 3
	}
	for {
		env, err := conn.ReadFrame()
		if err != nil {
			return 0 // 断链即退出
		}
		if env.GetKind() == pb.FrameKind_FRAME_KIND_STOPPING {
			if fault == "nostop" {
				time.Sleep(time.Hour) // 不合作：收到 STOPPING 也不退出
			}
			return 0
		}
	}
}

// 缩短的阶段预算与停止预算。
var (
	fastLimits = host.Limits{Handshake: 2 * time.Second, Config: 2 * time.Second, Migration: 2 * time.Second,
		OnStart: 2 * time.Second, OnReady: 2 * time.Second, StopTotal: time.Second}
	fastStop = host.StopTiming{Drain: 200 * time.Millisecond, Total: 800 * time.Millisecond, OnStopAt: 400 * time.Millisecond}
)

// logBuf 收集插件日志，供断言与失败时打印。
type logBuf struct {
	// mu 保护 b。
	mu sync.Mutex
	// b 是「插件ID|流|内容」形式的日志。
	b bytes.Buffer
}

// log 是 process.LogFunc，记录一行日志。
func (l *logBuf) log(id, stream string, line []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(&l.b, "%s|%s|%s\n", id, stream, line)
}

// String 返回已收集的全部日志。
func (l *logBuf) String() string { l.mu.Lock(); defer l.mu.Unlock(); return l.b.String() }

// setup 准备已安装 demo、含候选代次 gen-1 的状态库，并返回以 exe 为入口的启动描述。
func setup(t *testing.T, exe string) (*state.Store, Spec) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	s, err := state.Open(ctx, filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	digest := sha256.Sum256([]byte(exe))
	if err := s.AddInstall(ctx, state.Install{PluginID: "demo", Source: "test", Entrypoint: exe, BuildDigest: digest[:]}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddGeneration(ctx, "demo", "gen-1", "{}", 1); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugins", "demo")
	_ = os.MkdirAll(pluginDir, 0o700)
	return s, Spec{PluginID: "demo", PluginDir: pluginDir, Executable: exe, BuildDigest: digest[:], GenerationID: "gen-1",
		Bootstrap: &pb.Bootstrap{GenerationId: "gen-1", CandidateConfigJson: "{}", TargetConfigVersion: 1,
			DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE}},
		NewCandidate: true, Limits: fastLimits, Stop: fastStop}
}

// rawSpec 返回以测试二进制作为 rawPlugin 辅助进程、注入 fault 故障的启动描述与日志收集器。
func rawSpec(t *testing.T, fault string, stdio bool) (*state.Store, Spec, *logBuf) {
	s, spec := setup(t, os.Args[0])
	spec.Args = []string{"-test.run=^$"}
	// -race 编译的辅助进程默认在退出前 sleep 1s（GORACE atexit_sleep_ms），会被误判为不响应 STOPPING。
	spec.ExtraEnv = map[string]string{faultEnv: fault, "GORACE": "atexit_sleep_ms=0"}
	spec.Stdio = stdio
	logs := &logBuf{}
	spec.Log = logs.log
	return s, spec, logs
}

// alive 报告 pid 对应的进程是否仍存在。
func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

// 正常生命周期：两种传输都能激活，停止后进程退出、Listener 关闭、状态库提交新代次。
func TestLaunchAndStopBothTransports(t *testing.T) {
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) {
			s, spec, logs := rawSpec(t, "none", stdio)
			defer func() {
				if t.Failed() {
					t.Logf("插件日志:\n%s", logs.String())
				}
			}()
			r, err := Launch(context.Background(), s, spec)
			if err != nil {
				t.Fatal(err)
			}
			if r.Instance.State() != host.StateActivated {
				t.Fatalf("应为 Activated，实际 %s", r.Instance.State())
			}
			if g, _ := s.Active(context.Background(), "demo"); g != "gen-1" {
				t.Fatalf("应已提交 gen-1，实际 %q", g)
			}
			var addr string
			if r.listener != nil {
				addr = r.listener.Addr().String()
			}
			res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if res.Killed || r.Instance.State() != host.StateStopped {
				t.Fatalf("合作的插件应正常退出: %+v %s", res, r.Instance.State())
			}
			if st := r.Process.ExitState(); st == nil || st.ExitCode() != 0 {
				t.Fatalf("插件应以 0 退出: %v", st)
			}
			if alive(r.Process.Pid()) {
				t.Fatal("停止后进程仍存活")
			}
			if addr != "" {
				if c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond); err == nil {
					_ = c.Close()
					t.Fatal("停止后 Listener 仍在监听")
				}
			}
		})
	}
}

// 不响应 STOPPING 的插件在总截止被终止。
func TestNonCooperativePluginKilledAtDeadline(t *testing.T) {
	s, spec, _ := rawSpec(t, "nostop", false)
	r, err := Launch(context.Background(), s, spec)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	res := r.Stop(start, pb.StopReason_STOP_REASON_DISABLE)
	if !res.Killed || alive(r.Process.Pid()) {
		t.Fatalf("不响应 STOPPING 的插件应在总截止被终止: %+v", res)
	}
	if el := time.Since(start); el < fastStop.Total || el > fastStop.Total+time.Second {
		t.Fatalf("终止耗时 %v，应接近 %v", el, fastStop.Total)
	}
}

// 激活后崩溃：报告异常退出；有限重启在第三次时禁用。
func TestCrashAfterActivationAndRestartPolicy(t *testing.T) {
	s, spec, _ := rawSpec(t, "crash", false)
	policy := host.NewRestartPolicy()
	var decisions []host.Decision
	for i := 0; i < 3; i++ {
		r, err := Launch(context.Background(), s, spec)
		if err != nil {
			t.Fatal(err)
		}
		spec.NewCandidate = false // 之后的启动复用已提交代次
		abnormal, st, err := r.WaitExit()
		if err != nil {
			t.Fatal(err)
		}
		if !abnormal || !strings.Contains(st, "exit status 3") {
			t.Fatalf("应报告异常退出（退出码 3）: %v %s", abnormal, st)
		}
		decisions = append(decisions, policy.OnExit("demo", abnormal, time.Now()))
	}
	if decisions[0] != host.DecisionRestart || decisions[1] != host.DecisionRestart || decisions[2] != host.DecisionDisable {
		t.Fatalf("前两次应重启、第三次禁用: %v", decisions)
	}
}

// 握手挂起、非法帧、stdout 污染、连接前退出都不激活：报告对应错误，候选不提交。
func TestStartupFaultsAreKilledAndReported(t *testing.T) {
	cases := []struct {
		// fault 是注入的故障。
		fault string
		// want 是错误中应包含的片段；空串表示不检查。
		want string
		// stdio 为 true 时使用 stdio 传输。
		stdio bool
	}{
		{"hang", "DEADLINE_EXCEEDED", false},
		{"badframe", "PROTOCOL_ERROR", false},
		{"pollute", "", true}, // stdout 混入日志：读帧失败
		{"exit-early", "连接前退出", false},
	}
	for _, c := range cases {
		t.Run(c.fault, func(t *testing.T) {
			s, spec, _ := rawSpec(t, c.fault, c.stdio)
			r, err := Launch(context.Background(), s, spec)
			if err == nil {
				r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				t.Fatal("故障插件不应激活")
			}
			if c.want != "" && !strings.Contains(err.Error(), c.want) {
				t.Fatalf("错误应包含 %q: %v", c.want, err)
			}
			if g, _ := s.Active(context.Background(), "demo"); g != "" {
				t.Fatalf("失败的候选不得提交，实际 Active=%q", g)
			}
		})
	}
}

// 用真实 SDK 插件验证：同一插件并发启动时，runtime.lock 保证只有一个受管写入者。
func TestConcurrentLaunchSamePluginOnlyOneRuns(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "lifecycle")
	if out, err := exec.Command("go", "build", "-o", bin, "../../../examples/plugins/lifecycle").CombinedOutput(); err != nil {
		t.Fatalf("构建示例插件失败: %v\n%s", err, out)
	}
	s, spec := setup(t, bin)
	spec.Limits = host.DefaultLimits
	first, err := Launch(context.Background(), s, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	spec.NewCandidate = false
	logs := &logBuf{}
	spec.Log = logs.log
	second, err := Launch(context.Background(), s, spec)
	if err == nil {
		second.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		t.Fatal("旧实例仍在运行时，第二个实例不应启动")
	}
	// 插件 stderr 由后台 goroutine 异步转发，Launch 返回时最后几行可能还没送达。
	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(logs.String(), "runtime.lock") {
		if time.Now().After(deadline) {
			t.Fatalf("第二个实例应因 runtime.lock 拒绝启动；Launch 错误: %v；日志: %s", err, logs.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if first.Instance.State() != host.StateActivated {
		t.Fatal("第一个实例不应受影响")
	}
}

// 停止与启动共用逐插件管理锁：Stop 进行中发起的同一插件 Launch，要等 Stop 完成后才开始。
func TestStopAndLaunchSerializedPerPlugin(t *testing.T) {
	s, spec, _ := rawSpec(t, "nostop", false) // 不响应 STOPPING，Stop 会持续到总截止
	r, err := Launch(context.Background(), s, spec)
	if err != nil {
		t.Fatal(err)
	}
	stopDone := make(chan time.Time, 1)
	go func() { r.Stop(time.Now(), pb.StopReason_STOP_REASON_RESTART); stopDone <- time.Now() }()
	deadline := time.Now().Add(2 * time.Second)
	for r.Instance.State() == host.StateActivated {
		if time.Now().After(deadline) {
			t.Fatal("Stop 没有开始")
		}
		time.Sleep(time.Millisecond)
	}
	next := spec
	next.NewCandidate = false
	next.ExtraEnv = map[string]string{faultEnv: "none", "GORACE": "atexit_sleep_ms=0"}
	r2, err := Launch(context.Background(), s, next)
	launched := time.Now()
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	stopped := <-stopDone
	if launched.Before(stopped) {
		t.Fatalf("同一插件的 Launch 在 Stop 完成（%v）之前就返回（%v），管理操作未串行", stopped.Format("15:04:05.000"), launched.Format("15:04:05.000"))
	}
}

// Core 断开连接后插件以退出码 0 自行退出：正常退出，不进入崩溃计数。
func TestNormalExitIsNotAbnormal(t *testing.T) {
	s, spec, _ := rawSpec(t, "none", false)
	r, err := Launch(context.Background(), s, spec)
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Instance.Session.Close() // rawPlugin 读到 EOF 后 return 0
	abnormal, st, err := r.WaitExit()
	if err != nil || abnormal {
		t.Fatalf("退出码 0 应为正常退出: abnormal=%v state=%s err=%v", abnormal, st, err)
	}
}

// orphanPid 从插件日志中取出 orphan-* 辅助进程启动的子进程 pid。
func orphanPid(t *testing.T, logs *logBuf) int {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		for _, line := range strings.Split(logs.String(), "\n") {
			var pid int
			if _, err := fmt.Sscanf(line, "demo|stderr|orphan %d", &pid); err == nil {
				t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) }) // 断言失败时不遗留进程
				return pid
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("没有拿到子进程 pid，日志:\n%s", logs.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// 领头进程先退出、子进程仍留在进程组内：停止与崩溃恢复两条路径都必须在子进程消失后才返回。
func TestLeftoverChildReapedBeforeReturn(t *testing.T) {
	t.Run("stop", func(t *testing.T) {
		s, spec, logs := rawSpec(t, "orphan-stop", false)
		r, err := Launch(context.Background(), s, spec)
		if err != nil {
			t.Fatal(err)
		}
		pid := orphanPid(t, logs)
		res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		if res.Err != nil || !res.Killed {
			t.Fatalf("领头进程正常退出但子进程到总截止仍在，应终止进程树: %+v", res)
		}
		if alive(pid) {
			t.Fatalf("Stop 返回时子进程 %d 仍存活", pid)
		}
	})
	t.Run("crash", func(t *testing.T) {
		s, spec, logs := rawSpec(t, "orphan-crash", false)
		r, err := Launch(context.Background(), s, spec)
		if err != nil {
			t.Fatal(err)
		}
		pid := orphanPid(t, logs)
		abnormal, _, err := r.WaitExit()
		if err != nil || !abnormal {
			t.Fatalf("崩溃应报告异常退出且清理成功: abnormal=%v err=%v", abnormal, err)
		}
		if alive(pid) {
			t.Fatalf("WaitExit 返回时子进程 %d 仍存活，不能启动替代实例", pid)
		}
		if res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); res.Err != nil {
			t.Fatalf("清理后的 Stop 不应再触碰进程树: %+v", res)
		}
	})
}
