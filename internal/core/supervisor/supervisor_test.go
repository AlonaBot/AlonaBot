//go:build unix

package supervisor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
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
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// faultEnv 选择辅助进程注入的故障。
const faultEnv = "SUPERVISOR_TEST_FAULT"

const (
	scheduleActivationGate = "SUPERVISOR_TEST_ACTIVATION_GATE"
	scheduleResponseGate   = "SUPERVISOR_TEST_RESPONSE_GATE"
	scheduleInterval       = time.Second
	lifecycleWait          = 5 * time.Second
	lifecycleFailure       = 8 // 原始帧插件的调度或 HTTP 故障。
)

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
	helloMessage := &pb.Hello{PluginId: id, Secret: secret, SupportedMajor: 1, BuildDigest: digest}
	scheduleHTTP := strings.HasPrefix(fault, "schedule-http-")
	if scheduleHTTP {
		helloMessage.SupportedCapabilities = []*pb.ProtocolCapability{{Name: wire.CapabilityHTTPListener, Version: 1}}
	}
	hello, _ := wire.ControlEnvelope(nil, 0, id, pb.FrameKind_FRAME_KIND_HELLO, helloMessage)
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
	registration := &pb.Register{CandidateConfigJson: "{}"}
	registration.Migration = &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}
	if scheduleHTTP {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return lifecycleFailure
		}
		token := hex.EncodeToString(ack.GetBootstrap().GetHttpProxyAuth().GetToken())
		server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			if request.Header.Get(wire.HeaderProxyAuth) != token {
				http.Error(w, "invalid proxy token", http.StatusForbidden)
				return
			}
			if request.URL.Query().Get("hold") == "1" {
				fmt.Fprintln(os.Stderr, "HTTP request held")
				<-request.Context().Done() // 保留在途请求，直到 STOPPING 关闭 Listener。
				return
			}
			if _, err := fmt.Fprint(w, hex.EncodeToString(sid)); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		})}
		defer server.Close()
		go func() {
			if err := server.Serve(listener); err != http.ErrServerClosed {
				fmt.Fprintln(os.Stderr, err)
			}
		}()
		registration.HttpListener = &pb.HttpListener{Host: "127.0.0.1", Port: uint32(listener.Addr().(*net.TCPAddr).Port)}
		registration.Routes = []*pb.RouteDescriptor{
			{RouteId: "job.probe", Kind: pb.RouteKind_ROUTE_KIND_SCHEDULE,
				Spec: &pb.RouteDescriptor_Schedule{Schedule: &pb.ScheduleSpec{JobId: "probe",
					Trigger: &pb.ScheduleSpec_IntervalMs{IntervalMs: uint64(scheduleInterval.Milliseconds())}}}},
			{RouteId: "http.probe", Kind: pb.RouteKind_ROUTE_KIND_HTTP,
				Spec: &pb.RouteDescriptor_Http{Http: &pb.HttpRouteSpec{Method: http.MethodGet, Path: "/", IsPublic: true,
					AuthMode: pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_SIGNATURE}}},
		}
	}
	reg, _ := wire.ControlEnvelope(sid, 0, id, pb.FrameKind_FRAME_KIND_REGISTER, registration)
	_ = conn.WriteFrame(reg)
	recv() // PREPARED
	_, msg = recv()
	act := msg.(*pb.Activate)
	if gate := os.Getenv(scheduleActivationGate); scheduleHTTP && gate != "" {
		fmt.Fprintln(os.Stderr, "activation gate reached")
		if err := waitLifecycleFile(gate); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return lifecycleFailure
		}
	}
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
	var ticks int
	var pending *pb.Envelope // 第三次调用等超时及闸门后才响应。
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
		if !scheduleHTTP {
			continue
		}
		if env.GetKind() == pb.FrameKind_FRAME_KIND_REQUEST {
			request := &pb.HandlerRequest{}
			if err := proto.Unmarshal(env.GetPayload(), request); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return lifecycleFailure
			}
			ticks++
			fmt.Fprintf(os.Stderr, "schedule tick %d %d\n", request.GetEvent().GetEventPayload().GetScheduled().GetScheduledAtUnixMs(), time.Now().UnixMilli())
			if fault == "schedule-http-drain" {
				continue // 模拟一直执行的 callback，由停止阶段取消并回收进程。
			}
			if ticks == 3 && fault == "schedule-http-late" {
				pending = env
				continue
			}
		} else if env.GetKind() == pb.FrameKind_FRAME_KIND_CANCEL && pending != nil && env.GetCallId() == pending.GetCallId() {
			fmt.Fprintln(os.Stderr, "schedule cancellation received")
			if err := waitLifecycleFile(os.Getenv(scheduleResponseGate)); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return lifecycleFailure
			}
			env, pending = pending, nil
		} else {
			continue
		}
		response, _ := wire.ControlEnvelope(sid, env.GetProtocolMinor(), id, pb.FrameKind_FRAME_KIND_RESPONSE,
			&pb.HandlerOutcome{Matched: true, Propagation: pb.Propagation_PROPAGATION_CONTINUE})
		response.InitiatorSide, response.CallId = env.GetInitiatorSide(), env.GetCallId()
		if err := conn.WriteFrame(response); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return lifecycleFailure
		}
	}
}

// waitLifecycleFile 以文件闸门协调不同进程；实际阶段超时仍由 Core 负责。
func waitLifecycleFile(path string) error {
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		time.Sleep(time.Millisecond)
	}
}

// 缩短的阶段预算与停止预算。
var (
	fastLimits = host.Limits{Handshake: 2 * time.Second, Config: 2 * time.Second, Migration: 2 * time.Second,
		OnStart: 2 * time.Second, OnReady: 2 * time.Second, StopTotal: time.Second,
		SessionMaxInflight: tune.Defaults().SessionMaxInflight, SessionMaxPayload: tune.Defaults().SessionMaxPayload, RequestTimeout: tune.Defaults().RequestTimeout}
	fastStop = host.StopTiming{Drain: 200 * time.Millisecond, Total: 800 * time.Millisecond, OnStopAt: 400 * time.Millisecond}
)

// logBuf 收集插件输出与 Core 日志，供断言与失败时打印。
type logBuf struct {
	// mu 保护 b。
	mu sync.Mutex
	b  bytes.Buffer // 插件行带「插件ID|流|内容」前缀，Core 行由默认 Logger 写入
}

// Line 记录一行日志，签名与 process.LogFunc 相同。
func (l *logBuf) Line(id, stream string, line []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(&l.b, "%s|%s|%s\n", id, stream, line)
}

// Write 把默认 Logger 的输出写入同一日志缓冲区。
func (l *logBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

// SetSession 满足 PluginLog；测试不记录 session_id。
func (l *logBuf) SetSession([]byte) {}

// String 返回已收集的全部日志。
func (l *logBuf) String() string { l.mu.Lock(); defer l.mu.Unlock(); return l.b.String() }

// SetDefault 会同时修改 log 包的输出与标志，清理时一并恢复。
func setDefaultLogger(t *testing.T, logger *slog.Logger) {
	t.Helper()
	previous, writer, flags := slog.Default(), log.Writer(), log.Flags()
	slog.SetDefault(logger)
	t.Cleanup(func() {
		slog.SetDefault(previous)
		log.SetOutput(writer)
		log.SetFlags(flags)
	})
}

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
	if err := s.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: "demo", Source: "test", Entrypoint: exe, BuildDigest: digest[:]}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddGeneration(ctx, "demo", "gen-1", "{}", 1); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugins", "demo")
	_ = os.MkdirAll(pluginDir, 0o700)
	return s, Spec{Role: state.RolePlugin, PluginID: "demo", PluginDir: pluginDir, Executable: exe, BuildDigest: digest[:], GenerationID: "gen-1",
		Bootstrap: &pb.Bootstrap{GenerationId: "gen-1", CandidateConfigJson: "{}", TargetConfigVersion: 1,
			DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}},
		NewCandidate: true, Limits: fastLimits, Stop: fastStop, Log: &logBuf{}}
}

// rawSpec 返回以测试二进制作为 rawPlugin 辅助进程、注入 fault 故障的启动描述与日志收集器。
func rawSpec(t *testing.T, fault string, stdio bool) (*state.Store, Spec, *logBuf) {
	s, spec := setup(t, os.Args[0])
	spec.Args = []string{"-test.run=^$"}
	// -race 编译的辅助进程默认在退出前 sleep 1s（GORACE atexit_sleep_ms），会被误判为不响应 STOPPING。
	spec.ExtraEnv = map[string]string{faultEnv: fault, "GORACE": "atexit_sleep_ms=0"}
	spec.Stdio = stdio
	logs := &logBuf{}
	spec.Log = logs
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
					t.Logf("plugin log:\n%s", logs.String())
				}
			}()
			r, err := Launch(context.Background(), s, spec)
			if err != nil {
				t.Fatal(err)
			}
			if r.Instance.State() != host.StateActivated {
				t.Fatalf("should be Activated, got %s", r.Instance.State())
			}
			if g, _ := s.Active(context.Background(), "demo"); g != "gen-1" {
				t.Fatalf("gen-1 should be committed, got %q", g)
			}
			var addr string
			if r.listener != nil {
				addr = r.listener.Addr().String()
			}
			res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if res.Killed || r.Instance.State() != host.StateStopped {
				t.Fatalf("cooperating plugin should exit normally: %+v %s", res, r.Instance.State())
			}
			if st := r.Process.ExitState(); st == nil || st.ExitCode() != 0 {
				t.Fatalf("plugin should exit with 0: %v", st)
			}
			if alive(r.Process.Pid()) {
				t.Fatal("process still alive after stop")
			}
			if addr != "" {
				if c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond); err == nil {
					_ = c.Close()
					t.Fatal("listener still accepting after stop")
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
		t.Fatalf("plugin ignoring STOPPING should be killed at the total deadline: %+v", res)
	}
	if el := time.Since(start); el < fastStop.Total || el > fastStop.Total+time.Second {
		t.Fatalf("termination took %v, should be close to %v", el, fastStop.Total)
	}
}

// 激活后崩溃：报告异常退出；有限重启在第三次时禁用。
func TestCrashAfterActivationAndRestartPolicy(t *testing.T) {
	s, spec, _ := rawSpec(t, "crash", false)
	policy := host.NewRestartPolicy(tune.Defaults())
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
			t.Fatalf("should report an abnormal exit (exit code 3): %v %s", abnormal, st)
		}
		decisions = append(decisions, policy.OnExit("demo", abnormal, time.Now()))
	}
	if decisions[0] != host.DecisionRestart || decisions[1] != host.DecisionRestart || decisions[2] != host.DecisionDisable {
		t.Fatalf("should restart twice and disable on the third: %v", decisions)
	}
}

// 握手挂起、非法帧、stdout 污染、连接前退出都不激活：报告对应错误，候选不提交。
func TestStartupFaultsAreKilledAndReported(t *testing.T) {
	cases := []struct {
		fault string // 注入的故障
		want  string // 错误中应包含的片段；空串表示不检查
		// stdio 为 true 时使用 stdio 传输。
		stdio bool
	}{
		{"hang", "DEADLINE_EXCEEDED", false},
		{"badframe", "PROTOCOL_ERROR", false},
		{"pollute", "", true}, // stdout 混入日志：读帧失败
		{"exit-early", "exited before connecting", false},
	}
	for _, c := range cases {
		t.Run(c.fault, func(t *testing.T) {
			s, spec, _ := rawSpec(t, c.fault, c.stdio)
			r, err := Launch(context.Background(), s, spec)
			if err == nil {
				r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				t.Fatal("faulty plugin should not activate")
			}
			if c.want != "" && !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error should contain %q: %v", c.want, err)
			}
			if g, _ := s.Active(context.Background(), "demo"); g != "" {
				t.Fatalf("failed candidate must not be committed, got Active=%q", g)
			}
		})
	}
}

// 用真实 SDK 插件验证：同一插件并发启动时，runtime.lock 保证只有一个受管写入者。
func TestConcurrentLaunchSamePluginOnlyOneRuns(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "lifecycle")
	if out, err := exec.Command("go", "build", "-o", bin, "../../../examples/plugins/lifecycle").CombinedOutput(); err != nil {
		t.Fatalf("build example plugin: %v\n%s", err, out)
	}
	s, spec := setup(t, bin)
	spec.Limits = host.NewLimits(tune.Defaults())
	first, err := Launch(context.Background(), s, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	spec.NewCandidate = false
	logs := &logBuf{}
	spec.Log = logs
	second, err := Launch(context.Background(), s, spec)
	if err == nil {
		second.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
		t.Fatal("second instance should not start while the old instance is running")
	}
	// 插件 stderr 由后台 goroutine 异步转发，Launch 返回时最后几行可能还没送达。
	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(logs.String(), "runtime.lock") {
		if time.Now().After(deadline) {
			t.Fatalf("second instance should be rejected by runtime.lock; Launch error: %v; log: %s", err, logs.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if first.Instance.State() != host.StateActivated {
		t.Fatal("first instance should be unaffected")
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
			t.Fatal("Stop did not start")
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
		t.Fatalf("Launch for the same plugin returned (%[2]v) before Stop finished (%[1]v); management operations are not serialized", stopped.Format("15:04:05.000"), launched.Format("15:04:05.000"))
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
		t.Fatalf("exit code 0 should be a normal exit: abnormal=%v state=%s err=%v", abnormal, st, err)
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
			t.Fatalf("did not get the child pid, log:\n%s", logs.String())
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
			t.Fatalf("leader exited normally but the child is still alive at the total deadline; the tree should be killed: %+v", res)
		}
		if alive(pid) {
			t.Fatalf("child %d still alive when Stop returned", pid)
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
			t.Fatalf("crash should be reported as abnormal and cleaned up: abnormal=%v err=%v", abnormal, err)
		}
		if alive(pid) {
			t.Fatalf("child %d still alive when WaitExit returned; a replacement must not start", pid)
		}
		if res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); res.Err != nil {
			t.Fatalf("Stop after cleanup should not touch the process tree: %+v", res)
		}
	})
}

// scheduledTicks 读取插件实际收到的调度帧时间，供计数与固定起点间隔交叉核对。
func scheduledTicks(logs *logBuf) [][2]int64 {
	var ticks [][2]int64
	for _, line := range strings.Split(logs.String(), "\n") {
		var scheduled, received int64
		_, body, _ := strings.Cut(line, "|stderr|")
		if _, err := fmt.Sscanf(body, "schedule tick %d %d", &scheduled, &received); err == nil {
			ticks = append(ticks, [2]int64{scheduled, received})
		}
	}
	return ticks
}

func waitLifecycle(t *testing.T, logs *logBuf, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(lifecycleWait)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("lifecycle condition timed out; log:\n%s", logs.String())
		}
		time.Sleep(time.Millisecond)
	}
}

func lifecycleHTTP(proxy *httpproxy.Server, pluginID, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/plugins/"+pluginID+"/"+path, nil))
	return response
}

// TestScheduleAndHTTPLifecycle 验证真实 TCP 插件的调度、代理与统一停止流程。
func TestScheduleAndHTTPLifecycle(t *testing.T) {
	// 留出一次完整 interval，确保观察期间实际跨过下一个定时器触发点。
	const observation = scheduleInterval + scheduleInterval/10
	t.Run("activation_interval_and_late_response", func(t *testing.T) {
		store, spec, logs := rawSpec(t, "schedule-http-late", false)
		setDefaultLogger(t, slog.New(slog.NewTextHandler(logs, nil)))
		spec.HTTP = httpproxy.New(nil)
		spec.Scheduler = scheduler.New(scheduler.Options{JobTimeout: 250 * time.Millisecond, Store: store, TasksPerPlugin: tune.Defaults().TasksPerPlugin, TaskPayloadBytes: tune.Defaults().TaskPayloadBytes, TaskConcurrency: tune.Defaults().TaskConcurrency})
		t.Cleanup(spec.Scheduler.Close)
		spec.Limits.OnReady = lifecycleWait
		activationGate := filepath.Join(t.TempDir(), "activate")
		responseGate := filepath.Join(t.TempDir(), "respond")
		spec.ExtraEnv[scheduleActivationGate] = activationGate
		spec.ExtraEnv[scheduleResponseGate] = responseGate

		// 1. 让 Launch 留在 ACTIVATED 之前，跨过一个 interval 后仍无业务入口。
		var running *Running
		var launchErr error
		launched := make(chan struct{}) // 关闭后发布 Launch 结果，避免轮询共享指针。
		go func() {
			running, launchErr = Launch(context.Background(), store, spec)
			close(launched)
		}()
		t.Cleanup(func() {
			for _, gate := range []string{activationGate, responseGate} {
				if err := os.WriteFile(gate, nil, 0o600); err != nil {
					t.Error(err)
				}
			}
			<-launched
			if running != nil {
				running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			}
		})
		waitLifecycle(t, logs, func() bool { return strings.Contains(logs.String(), "activation gate reached") })
		select {
		case <-launched:
			t.Fatalf("Launch finished before ACTIVATED: %v", launchErr)
		case <-time.After(observation):
		}
		if got := spec.Scheduler.Stats("demo", "probe"); got != (scheduler.JobStats{}) || len(scheduledTicks(logs)) != 0 {
			t.Fatalf("scheduler ran before ACTIVATED: %+v; log:\n%s", got, logs.String())
		}
		if got := lifecycleHTTP(spec.HTTP, "demo", "").Code; got != http.StatusServiceUnavailable {
			t.Fatalf("HTTP before ACTIVATED = %d, want 503", got)
		}
		activatedAfter := time.Now()
		if err := os.WriteFile(activationGate, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		select {
		case <-launched:
		case <-time.After(lifecycleWait):
			t.Fatal("Launch did not finish after activation gate opened")
		}
		activatedBefore := time.Now()
		if launchErr != nil {
			t.Fatal(launchErr)
		}
		if got := lifecycleHTTP(spec.HTTP, "demo", "").Code; got != http.StatusOK {
			t.Fatalf("HTTP after ACTIVATED = %d, want 200", got)
		}

		// 2. 前两次正常响应，第三次超时后跨过下一 tick，槽仍占用且不重叠。
		waitLifecycle(t, logs, func() bool { return spec.Scheduler.Stats("demo", "probe").Skipped != 0 })
		spec.Scheduler.Drain("demo") // 固定统计快照，迟到响应只负责结束原调用。
		stats := spec.Scheduler.Stats("demo", "probe")
		ticks := scheduledTicks(logs)
		if stats.Runs != 3 || stats.TimedOut != 1 || stats.Late != 0 || !stats.Running || len(ticks) != int(stats.Runs) {
			t.Fatalf("timed-out job stats = %+v, received ticks = %v", stats, ticks)
		}
		first := time.UnixMilli(ticks[0][0])
		if first.Before(activatedAfter.Add(scheduleInterval).Truncate(time.Millisecond)) || first.After(activatedBefore.Add(scheduleInterval)) {
			t.Fatalf("first tick = %v, activation window = [%v, %v]", first, activatedAfter, activatedBefore)
		}
		for i, tick := range ticks {
			if tick[1] < tick[0] || (i != 0 && tick[0]-ticks[i-1][0] != scheduleInterval.Milliseconds()) {
				t.Fatalf("tick timing disagrees with the fixed interval: %v", ticks)
			}
		}
		waitLifecycle(t, logs, func() bool { return strings.Contains(logs.String(), "schedule cancellation received") })
		if err := os.WriteFile(responseGate, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		waitLifecycle(t, logs, func() bool { return !spec.Scheduler.Stats("demo", "probe").Running })
		if got := spec.Scheduler.Stats("demo", "probe"); got.Runs != stats.Runs || got.Skipped != stats.Skipped || got.TimedOut != 1 || got.Late != 1 {
			t.Fatalf("late response stats = %+v, before response = %+v", got, stats)
		}
		t.Logf("received %d fixed-interval ticks; timeout retained the slot until the late response", len(ticks))
	})

	t.Run("stop_freezes_and_closes_listeners", func(t *testing.T) {
		store, spec, logs := rawSpec(t, "schedule-http-drain", false)
		setDefaultLogger(t, slog.New(slog.NewTextHandler(logs, nil)))
		spec.HTTP = httpproxy.New(nil)
		spec.Scheduler = scheduler.New(scheduler.Options{JobTimeout: time.Minute, Store: store, TasksPerPlugin: tune.Defaults().TasksPerPlugin, TaskPayloadBytes: tune.Defaults().TaskPayloadBytes, TaskConcurrency: tune.Defaults().TaskConcurrency})
		t.Cleanup(spec.Scheduler.Close)
		spec.Stop = host.StopTiming{Drain: 1500 * time.Millisecond, Total: 2500 * time.Millisecond, OnStopAt: time.Second}
		running, err := Launch(context.Background(), store, spec)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) })
		waitLifecycle(t, logs, func() bool { return len(scheduledTicks(logs)) == 1 })

		// 1. 同时保留 RPC 与 HTTP 在途请求，让两个排空步骤共享绝对截止。
		heldHTTP := make(chan int, 1)
		go func() { heldHTTP <- lifecycleHTTP(spec.HTTP, "demo", "?hold=1").Code }()
		waitLifecycle(t, logs, func() bool { return strings.Contains(logs.String(), "HTTP request held") })
		t0 := time.Now()
		type stopped struct {
			result host.StopResult
			at     time.Time
		}
		finished := make(chan stopped, 1)
		go func() {
			result := running.Stop(t0, pb.StopReason_STOP_REASON_DISABLE)
			finished <- stopped{result, time.Now()}
		}()
		waitLifecycle(t, logs, func() bool { return lifecycleHTTP(spec.HTTP, "demo", "").Code == http.StatusServiceUnavailable })
		frozen := spec.Scheduler.Stats("demo", "probe")
		if running.Instance.State() != host.StateDraining || !frozen.Running || frozen.Runs != 1 {
			t.Fatalf("expected a draining live callback: state=%s stats=%+v", running.Instance.State(), frozen)
		}
		select {
		case done := <-finished:
			t.Fatalf("Stop returned before the shared drain deadline: %+v", done.result)
		case <-time.After(observation):
		}
		if got := spec.Scheduler.Stats("demo", "probe"); got != frozen || len(scheduledTicks(logs)) != 1 {
			t.Fatalf("scheduler advanced after Drain: before=%+v after=%+v; log:\n%s", frozen, got, logs.String())
		}
		if got := lifecycleHTTP(spec.HTTP, "demo", "").Code; got != http.StatusServiceUnavailable {
			t.Fatalf("HTTP during Drain = %d, want 503", got)
		}

		// 2. STOPPING 关闭 Listener、取消余留调用，Stop 返回前确认进程树消失。
		var done stopped
		select {
		case done = <-finished:
		case <-time.After(lifecycleWait):
			t.Fatal("Stop did not finish within its total budget")
		}
		result := done.result
		if result.Err != nil || result.Killed || result.Drained || running.Instance.State() != host.StateStopped || alive(running.Process.Pid()) {
			t.Fatalf("stop result = %+v, state = %s", result, running.Instance.State())
		}
		if result.StoppingSentAt.Before(t0.Add(spec.Stop.Drain)) || result.StopDeadline.After(t0.Add(spec.Stop.Total)) || done.at.Sub(t0) > spec.Stop.Total {
			t.Fatalf("stop exceeded the shared budget: elapsed=%v result=%+v", done.at.Sub(t0), result)
		}
		endpoint := running.Instance.HTTPListener
		for _, address := range []string{running.listener.Addr().String(), net.JoinHostPort(endpoint.GetHost(), fmt.Sprint(endpoint.GetPort()))} {
			if conn, err := net.DialTimeout("tcp", address, time.Second); err == nil {
				conn.Close()
				t.Fatalf("listener %s still accepts connections after Stop", address)
			}
		}
		select {
		case status := <-heldHTTP:
			if status != http.StatusBadGateway {
				t.Fatalf("inflight HTTP after Listener closure = %d, want 502", status)
			}
		case <-time.After(lifecycleWait):
			t.Fatal("inflight HTTP survived Listener closure")
		}
		if got := spec.Scheduler.Stats("demo", "probe"); got.Running || got.Runs != frozen.Runs || got.Skipped != frozen.Skipped {
			t.Fatalf("job stats after process exit = %+v", got)
		}
		if got := lifecycleHTTP(spec.HTTP, "demo", "").Code; got != http.StatusServiceUnavailable {
			t.Fatalf("HTTP after Stop = %d, want 503", got)
		}
		t.Logf("RPC and HTTP drained against one deadline; Stop finished in %v of %v", done.at.Sub(t0), spec.Stop.Total)
	})

	t.Run("crash_and_new_session", func(t *testing.T) {
		store, spec, logs := rawSpec(t, "schedule-http-drain", false)
		setDefaultLogger(t, slog.New(slog.NewTextHandler(logs, nil)))
		spec.HTTP = httpproxy.New(nil)
		spec.Scheduler = scheduler.New(scheduler.Options{JobTimeout: 250 * time.Millisecond, Store: store, TasksPerPlugin: tune.Defaults().TasksPerPlugin, TaskPayloadBytes: tune.Defaults().TaskPayloadBytes, TaskConcurrency: tune.Defaults().TaskConcurrency})
		t.Cleanup(spec.Scheduler.Close)
		old, err := Launch(context.Background(), store, spec)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { old.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) })
		oldHTTP := lifecycleHTTP(spec.HTTP, "demo", "")
		if oldHTTP.Code != http.StatusOK {
			t.Fatalf("HTTP before crash = %d, want 200", oldHTTP.Code)
		}
		waitLifecycle(t, logs, func() bool { return spec.Scheduler.Stats("demo", "probe").TimedOut == 1 })
		if got := spec.Scheduler.Stats("demo", "probe"); !got.Running {
			t.Fatalf("callback slot released before process crash: %+v", got)
		}

		// 1. 杀掉真实进程模拟崩溃，WaitExit 撤销映射并确认旧槽释放。
		if err := old.Process.KillTree(); err != nil {
			t.Fatal(err)
		}
		if abnormal, state, err := old.WaitExit(); err != nil || !abnormal {
			t.Fatalf("crash cleanup: abnormal=%v state=%s err=%v", abnormal, state, err)
		}
		if got := lifecycleHTTP(spec.HTTP, "demo", "").Code; got != http.StatusServiceUnavailable {
			t.Fatalf("HTTP after crash = %d, want 503", got)
		}
		endpoint := old.Instance.HTTPListener
		address := net.JoinHostPort(endpoint.GetHost(), fmt.Sprint(endpoint.GetPort()))
		if conn, err := net.DialTimeout("tcp", address, time.Second); err == nil {
			conn.Close()
			t.Fatal("old HTTP endpoint still accepts connections after crash")
		}
		select {
		case <-old.Instance.Session.Done():
		default:
			t.Fatal("old RPC Session survived crash cleanup")
		}
		beforeRestart := spec.Scheduler.Stats("demo", "probe")
		if beforeRestart.Running {
			t.Fatalf("old running slot survived process cleanup: %+v", beforeRestart)
		}
		<-time.After(observation)
		if got := spec.Scheduler.Stats("demo", "probe"); got != beforeRestart {
			t.Fatalf("old registration advanced after crash: before=%+v after=%+v", beforeRestart, got)
		}

		// 2. 同代次重启使用新 Session/token，代理成功而旧凭证被新端点拒绝。
		spec.NewCandidate = false
		spec.ExtraEnv[faultEnv] = "schedule-http-live"
		restartLogs := &logBuf{}
		spec.Log = restartLogs
		savedJobs, err := store.Jobs(context.Background(), "demo")
		if err != nil || len(savedJobs) != 1 {
			t.Fatalf("saved jobs before restart = %v, err = %v", savedJobs, err)
		}
		restartedAt := time.Now()
		current, err := Launch(context.Background(), store, spec)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { current.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) })
		response := lifecycleHTTP(spec.HTTP, "demo", "")
		if response.Code != http.StatusOK || response.Body.String() == oldHTTP.Body.String() {
			t.Fatalf("new Session route unavailable or stale: status=%d", response.Code)
		}
		if bytes.Equal(current.Instance.HTTPProxyToken, old.Instance.HTTPProxyToken) {
			t.Fatal("HTTP proxy token was reused across Sessions")
		}
		endpoint = current.Instance.HTTPListener
		request, err := http.NewRequest(http.MethodGet, "http://"+net.JoinHostPort(endpoint.GetHost(), fmt.Sprint(endpoint.GetPort()))+"/", nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set(wire.HeaderProxyAuth, hex.EncodeToString(old.Instance.HTTPProxyToken))
		client := &http.Client{Timeout: lifecycleWait}
		rejected, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		if err := rejected.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if rejected.StatusCode != http.StatusForbidden {
			t.Fatalf("old token at new endpoint = %d, want 403", rejected.StatusCode)
		}
		waitLifecycle(t, restartLogs, func() bool {
			stats := spec.Scheduler.Stats("demo", "probe")
			return stats.Runs == beforeRestart.Runs+1 && !stats.Running && len(scheduledTicks(restartLogs)) == 1
		})
		// 重启沿保存的起点取下一未来 tick；握手跨过的 tick 直接作废。
		anchor := savedJobs[0].Anchor
		next := anchor.Add((restartedAt.Sub(anchor)/scheduleInterval + 1) * scheduleInterval)
		if tick := scheduledTicks(restartLogs)[0][0]; tick < next.UnixMilli() || (tick-anchor.UnixMilli())%scheduleInterval.Milliseconds() != 0 {
			t.Fatalf("restart tick is off the saved schedule: scheduled_at=%d next_at=%d anchor=%s", tick, next.UnixMilli(), anchor)
		}
		if got := spec.Scheduler.Stats("demo", "probe"); got.TimedOut != beforeRestart.TimedOut || got.Late != beforeRestart.Late {
			t.Fatalf("old execution affected the new Session: before=%+v after=%+v", beforeRestart, got)
		}
		t.Log("crash removed the old endpoint and slot; restart used a new Session token and future tick")
	})

	t.Run("parallel_stop_shared_budget", func(t *testing.T) {
		store, spec, logs := rawSpec(t, "schedule-http-drain", false)
		setDefaultLogger(t, slog.New(slog.NewTextHandler(logs, nil)))
		spec.HTTP = httpproxy.New(nil)
		spec.Scheduler = scheduler.New(scheduler.Options{JobTimeout: time.Minute, Store: store, TasksPerPlugin: tune.Defaults().TasksPerPlugin, TaskPayloadBytes: tune.Defaults().TaskPayloadBytes, TaskConcurrency: tune.Defaults().TaskConcurrency})
		t.Cleanup(spec.Scheduler.Close)
		spec.Stop = host.StopTiming{Drain: 1500 * time.Millisecond, Total: 2500 * time.Millisecond, OnStopAt: time.Second}
		next := spec
		next.PluginID, next.GenerationID = "other", "gen-2"
		next.PluginDir = filepath.Join(filepath.Dir(spec.PluginDir), next.PluginID)
		next.Bootstrap = proto.Clone(spec.Bootstrap).(*pb.Bootstrap)
		next.Bootstrap.GenerationId = next.GenerationID
		if err := os.MkdirAll(next.PluginDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := store.AddInstall(context.Background(), state.Install{Role: state.RolePlugin, PluginID: next.PluginID,
			Source: "test", Entrypoint: next.Executable, BuildDigest: next.BuildDigest}); err != nil {
			t.Fatal(err)
		}
		if err := store.AddGeneration(context.Background(), next.PluginID, next.GenerationID, "{}", 1); err != nil {
			t.Fatal(err)
		}

		// 1. 两个插件共用状态库、Scheduler 与代理，各保留一个真实 RPC 和 HTTP 请求。
		var targets []*Running
		heldHTTP := make(chan int, 2)
		for _, description := range []Spec{spec, next} {
			running, err := Launch(context.Background(), store, description)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) })
			targets = append(targets, running)
		}
		for _, running := range targets {
			id := running.pluginID
			waitLifecycle(t, logs, func() bool { return spec.Scheduler.Stats(id, "probe").Runs == 1 })
			go func() { heldHTTP <- lifecycleHTTP(spec.HTTP, id, "?hold=1").Code }()
			waitLifecycle(t, logs, func() bool { return strings.Contains(logs.String(), id+"|stderr|HTTP request held") })
		}

		// 2. 同一 t0 并行停止；完整排空耗时若逐插件累加就会超出共同总预算。
		type stopped struct {
			id      string
			result  host.StopResult
			elapsed time.Duration
		}
		finished := make(chan stopped, len(targets))
		t0 := time.Now()
		for _, running := range targets {
			go func() {
				result := running.Stop(t0, pb.StopReason_STOP_REASON_DISABLE)
				finished <- stopped{running.pluginID, result, time.Since(t0)}
			}()
		}
		for range targets {
			select {
			case done := <-finished:
				if done.result.Err != nil || done.result.Killed || done.result.Drained || done.elapsed > spec.Stop.Total ||
					done.result.StoppingSentAt.Before(t0.Add(spec.Stop.Drain)) || done.result.StopDeadline.After(t0.Add(spec.Stop.Total)) {
					t.Fatalf("plugin %s exceeded shared stop budget: elapsed=%v result=%+v", done.id, done.elapsed, done.result)
				}
				t.Logf("plugin %s stopped %v after shared t0 (total budget %v)", done.id, done.elapsed, spec.Stop.Total)
			case <-time.After(lifecycleWait):
				t.Fatal("parallel Stop did not finish")
			}
		}
		for _, running := range targets {
			if alive(running.Process.Pid()) || spec.Scheduler.Stats(running.pluginID, "probe").Running {
				t.Fatalf("plugin %s retained its process or running slot", running.pluginID)
			}
			if got := lifecycleHTTP(spec.HTTP, running.pluginID, "").Code; got != http.StatusServiceUnavailable {
				t.Fatalf("plugin %s HTTP after Stop = %d, want 503", running.pluginID, got)
			}
			select {
			case status := <-heldHTTP:
				if status != http.StatusBadGateway {
					t.Fatalf("inflight HTTP after Stop = %d, want 502", status)
				}
			case <-time.After(lifecycleWait):
				t.Fatal("inflight HTTP survived parallel Stop")
			}
		}
	})
}
