// alonabot 命令启动 AlonaBot Core。
//
// run 子命令用于开发：在给定的 data_dir 中登记并启动一个本地插件，
// 收到 SIGINT 或 SIGTERM 时停止它；插件异常退出时按有限重启策略重启或禁用。
// Core 与插件的日志写到 stderr，并写入 <data_dir>/logs/ 下按日期轮转的文件（CONTRACTS §14）。
// tune list 列出调优参数；全局 --args 在启动前覆盖并校验参数。
// 首次启动生成管理 token：交互终端只显示一次，非终端写入 admin_token，show 可在本机读出。
// admin-token init/reset 在 Core 停止后离线更新管理凭证，仅向本机终端显示一次。
package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/core/webhost"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// 进程退出码。
const (
	exitFailure = 1 // 运行失败
	exitUsage   = 2 // 参数错误
)

// logsDir 是 data_dir 下日志文件的目录（CONTRACTS §14.6）。
const logsDir = "logs"

// main 统一返回命令执行的退出码。
func main() { os.Exit(execute(os.Args[1:])) }

// run 打开 data_dir、登记开发安装记录，然后启动插件；收到 SIGINT/SIGTERM 时停止它，
// 异常退出按有限重启策略重启或禁用。
func run(dataDir, pluginID, entry string, stdio bool, v tune.Values, overrides []tune.Override, selected startupOverrides) (runErr error) {
	if v.CoreMemorySoftLimit > 0 {
		debug.SetMemoryLimit(v.CoreMemorySoftLimit)
	}
	// 1. 取得 core.lock，加载设置与日志，交付首次凭证并装配公开入口。
	ctx := context.Background()
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		return err
	}
	defer store.Close()
	settingsFile, err := settings.Load(dataDir)
	if err != nil {
		return &usageError{err}
	}
	configuration := settingsFile.Settings()
	baseURL, err := url.Parse(configuration.BaseURL)
	if err != nil {
		return fmt.Errorf("parse base_url: %w", err)
	}
	if selected.listen != nil {
		if err := settings.ValidateListen(*selected.listen); err != nil {
			return &usageError{fmt.Errorf("--listen: %w", err)}
		}
		host, _, _ := net.SplitHostPort(*selected.listen) // 已通过 ValidateListen。
		ip, _ := netip.ParseAddr(host)
		if !ip.IsLoopback() && !strings.EqualFold(strings.TrimSuffix(host, "."), "localhost") && baseURL.Scheme != "https" && len(configuration.TrustedProxies) == 0 {
			return &usageError{errors.New("--listen: non-loopback listen requires HTTPS base_url or trusted_proxies")}
		}
		configuration.Listen = *selected.listen
	}
	if selected.level != nil {
		configuration.LogLevel = *selected.level
	}
	if selected.format != nil {
		configuration.LogFormat = *selected.format
	}
	out, err := logging.Open(logging.Options{Format: logging.Format(configuration.LogFormat), Level: configuration.LogLevel, Stderr: os.Stderr})
	if err != nil {
		return &usageError{err}
	}
	slog.SetDefault(out.Logger())
	defer out.Close()
	defer func() {
		if runErr != nil {
			slog.Error("Core stopped", "err", runErr)
			runErr = &reportedError{runErr}
		}
	}()
	if configuration.LogFile.Enabled {
		if err := out.EnableFile(logging.FileOptions{Dir: filepath.Join(dataDir, logsDir), MaxSize: configuration.LogFile.MaxSize, MaxTotal: configuration.LogFile.MaxTotal, MaxAgeDays: configuration.LogFile.MaxAgeDays}); err != nil {
			return err
		}
	}
	interactive := terminal(os.Stdin) && terminal(os.Stderr)
	token, err := adminauth.FirstToken(ctx, store, dataDir, interactive)
	if err != nil {
		return err
	}
	if token != "" {
		if _, err := fmt.Fprintln(os.Stderr, token); err != nil {
			return fmt.Errorf("write admin token: %w", err)
		}
	} else if !interactive {
		name := filepath.Join(dataDir, adminauth.FirstTokenFile)
		if _, err := os.Stat(name); err == nil {
			slog.Info("first admin token is available", "path", name)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat first admin token file: %w", err)
		}
	}
	if configuration.Proxy != "" {
		for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY"} {
			if err := os.Setenv(name, configuration.Proxy); err != nil {
				return fmt.Errorf("set %s: %w", name, err)
			}
		}
	}
	environment := make(map[string]string)
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		if value, exists := os.LookupEnv(name); exists {
			environment[name] = value
		}
	}
	authentication := adminauth.New(store, time.Now, baseURL, dataDir)
	adminMux := http.NewServeMux()
	authentication.Routes(adminMux)
	adminMux.Handle("/", authentication.Require(http.NotFoundHandler()))
	proxy := httpproxy.New(authentication)
	publicHost := webhost.New(webhost.Options{BaseURL: baseURL, Proxy: proxy, Admin: adminMux})
	listener, err := net.Listen("tcp", configuration.Listen)
	if err != nil {
		return fmt.Errorf("listen HTTP: %w", err)
	}
	httpServer := &http.Server{Handler: publicHost, ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError)}
	httpErrors := make(chan error, 1)
	go func() { httpErrors <- httpServer.Serve(listener) }()
	defer httpServer.Close()
	slog.Info("HTTP listener started", "listen", listener.Addr().String())

	if len(overrides) > 0 {
		attrs := make([]any, 0, len(overrides))
		for _, override := range overrides {
			attrs = append(attrs, slog.Group(override.Name, "value", override.Value, "origin", override.Origin))
		}
		slog.Warn("tuning parameters overridden", attrs...)
	}
	slog.Info("Core started", "data_dir", dataDir)

	exe, err := filepath.Abs(entry)
	if err != nil {
		return err
	}
	digest, err := fileDigest(exe)
	if err != nil {
		return err
	}
	pluginDir := filepath.Join(dataDir, "plugins", pluginID)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return err
	}
	// 2. 开发模式：首次运行时登记安装记录（不做正式安装与来源校验）。
	if _, err := store.GetInstall(ctx, pluginID); errors.Is(err, state.ErrNotFound) {
		if err := store.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: pluginID, Source: "dev-local", Entrypoint: exe, BuildDigest: digest}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 3. 启动插件并等待信号或进程退出；只有旧进程树确认清空后才重启。
	policy := host.NewRestartPolicy(v)
	global := wire.NewBudget(v.CoreMaxPayload)
	jobs := scheduler.New(scheduler.Options{JobTimeout: v.JobTimeout})
	defer jobs.Close()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)
	for {
		r, err := launch(ctx, out, store, pluginID, pluginDir, exe, digest, stdio, v, global, environment, jobs, proxy)
		if err != nil {
			return err
		}
		stop := func() error {
			now := time.Now()
			stopContext, cancel := context.WithDeadline(context.Background(), now.Add(v.StopTotal))
			defer cancel()
			httpStopped := make(chan error, 1)
			go func() { httpStopped <- httpServer.Shutdown(stopContext) }()
			res := r.Stop(now, pb.StopReason_STOP_REASON_CORE_SHUTDOWN)
			slog.With(logging.KeyPluginID, pluginID).Info("plugin stopped", "drained", res.Drained, "killed", res.Killed)
			return errors.Join(res.Err, <-httpStopped)
		}

		// exit 记录一次 WaitExit 的结果。
		type exit struct {
			// abnormal 为 true 表示未经 Stop 且退出码非 0。
			abnormal bool
			// err 非 nil 表示剩余进程树未确认清空。
			err error
		}
		exitc := make(chan exit, 1)
		go func() {
			abnormal, st, err := r.WaitExit()
			if abnormal {
				slog.With(logging.KeyPluginID, pluginID).Error("plugin exited abnormally", "state", st)
			} else {
				slog.With(logging.KeyPluginID, pluginID).Info("plugin exited", "state", st)
			}
			exitc <- exit{abnormal, err}
		}()
		select {
		case <-sigc:
			if err := stop(); err != nil {
				return fmt.Errorf("stop Core: %w", err)
			}
			slog.Info("Core stopped")
			return nil
		case err := <-httpErrors:
			return errors.Join(fmt.Errorf("serve HTTP: %w", err), stop())
		case e := <-exitc:
			if e.err != nil {
				return fmt.Errorf("process tree of plugin %s not confirmed empty, not starting a replacement: %w", pluginID, e.err)
			}
			switch policy.OnExit(pluginID, e.abnormal, time.Now()) {
			case host.DecisionRestart:
				slog.With(logging.KeyPluginID, pluginID).Info("restarting plugin")
				continue
			case host.DecisionDisable:
				_ = store.SetEnabled(ctx, pluginID, false)
				return fmt.Errorf("plugin %s exited abnormally %d times within %s and has been disabled", pluginID, v.RestartMaxCrashes, v.RestartWindow)
			default:
				slog.Info("Core stopped")
				return nil
			}
		}
	}
}

// launch 首次运行时提交新代次；之后复用已提交的 Active 代次（普通 restart，不产生新 Generation）。
// Supervisor 发布 HTTP 映射并接入调度后，再记录激活。
func launch(ctx context.Context, out *logging.Output, store *state.Store, pluginID, pluginDir, exe string, digest []byte, stdio bool, v tune.Values, global *wire.Budget, environment map[string]string, jobs *scheduler.Scheduler, proxy *httpproxy.Server) (*supervisor.Running, error) {
	active, err := store.Active(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	gen, isNew := active, active == ""
	if isNew {
		gen = fmt.Sprintf("gen-%d", time.Now().UnixNano())
		if err := store.AddGeneration(ctx, pluginID, gen, "{}", 1); err != nil {
			return nil, err
		}
	}
	r, err := supervisor.Launch(ctx, store, supervisor.Spec{Role: state.RolePlugin,
		PluginID: pluginID, PluginDir: pluginDir, Executable: exe, BuildDigest: digest, GenerationID: gen,
		Bootstrap: &pb.Bootstrap{GenerationId: gen, CandidateConfigJson: "{}", TargetConfigVersion: 1,
			DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE},
			LogLevel:           logging.ProtoLevel(out.Levels().Source(pluginID))},
		NewCandidate: isNew, Stdio: stdio, Limits: host.NewLimits(v), Stop: host.NewStopTiming(v), Log: out.Plugin(pluginID, gen),
		Global: global, MemorySoftLimit: v.PluginMemorySoftLimit, ExtraEnv: environment, Scheduler: jobs, HTTP: proxy,
	})
	if err != nil {
		return nil, err
	}
	mode := "restart"
	if isNew {
		mode = "new"
	}
	slog.With(logging.KeyPluginID, pluginID).Info("plugin activated", logging.KeyGenerationID, gen, "mode", mode, "pid", r.Process.Pid())
	return r, nil
}

// fileDigest 返回文件内容的 SHA256。
func fileDigest(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
