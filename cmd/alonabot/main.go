// alonabot 命令启动 AlonaBot Core。
//
// run 子命令用于开发：在给定的 data_dir 中登记并启动一个本地插件，
// 收到 SIGINT 或 SIGTERM 时停止它；插件异常退出时按有限重启策略重启或禁用。
// Core 与插件的日志写到 stderr，并写入 <data_dir>/logs/ 下按日期轮转的文件（CONTRACTS §14）。
// tune list 列出调优参数；全局 --args 在启动前覆盖并校验参数。
// 正式的安装与管理命令尚未提供。
package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
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
func run(out *logging.Output, dataDir, pluginID, entry string, stdio bool, v tune.Values) error {
	if v.CoreMemorySoftLimit > 0 {
		debug.SetMemoryLimit(v.CoreMemorySoftLimit)
	}
	// 1. 打开状态库（取得 core.lock）后才开始写日志文件，再定位入口并计算摘要。
	ctx := context.Background()
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := out.EnableFile(logging.FileOptions{Dir: filepath.Join(dataDir, logsDir), MaxSize: logging.DefaultMaxFileSize,
		MaxTotal: logging.DefaultMaxTotalSize, MaxAgeDays: logging.DefaultMaxAgeDays}); err != nil {
		return err
	}
	log := out.Logger()
	log.Info("Core started", "data_dir", dataDir)
	plog := log.With(logging.KeyPluginID, pluginID)

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
		if err := store.AddInstall(ctx, state.Install{PluginID: pluginID, Source: "dev-local", Entrypoint: exe, BuildDigest: digest}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 3. 启动插件并等待信号或进程退出；只有旧进程树确认清空后才重启。
	policy := host.NewRestartPolicy(v)
	global := wire.NewBudget(v.CoreMaxPayload)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	for {
		r, err := launch(ctx, out, plog, store, pluginID, pluginDir, exe, digest, stdio, v, global)
		if err != nil {
			return err
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
				plog.Error("plugin exited abnormally", "state", st)
			} else {
				plog.Info("plugin exited", "state", st)
			}
			exitc <- exit{abnormal, err}
		}()
		select {
		case <-sigc:
			res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_CORE_SHUTDOWN)
			plog.Info("plugin stopped", "drained", res.Drained, "killed", res.Killed)
			if res.Err != nil {
				return fmt.Errorf("process tree of plugin %s not confirmed empty: %w", pluginID, res.Err)
			}
			log.Info("Core stopped")
			return nil
		case e := <-exitc:
			if e.err != nil {
				return fmt.Errorf("process tree of plugin %s not confirmed empty, not starting a replacement: %w", pluginID, e.err)
			}
			switch policy.OnExit(pluginID, e.abnormal, time.Now()) {
			case host.DecisionRestart:
				plog.Info("restarting plugin")
				continue
			case host.DecisionDisable:
				_ = store.SetEnabled(ctx, pluginID, false)
				return fmt.Errorf("plugin %s exited abnormally %d times within %s and has been disabled", pluginID, v.RestartMaxCrashes, v.RestartWindow)
			default:
				log.Info("Core stopped")
				return nil
			}
		}
	}
}

// launch 首次运行时提交新代次；之后复用已提交的 Active 代次（普通 restart，不产生新 Generation）。
func launch(ctx context.Context, out *logging.Output, plog *slog.Logger, store *state.Store, pluginID, pluginDir, exe string, digest []byte, stdio bool, v tune.Values, global *wire.Budget) (*supervisor.Running, error) {
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
	r, err := supervisor.Launch(ctx, store, supervisor.Spec{
		PluginID: pluginID, PluginDir: pluginDir, Executable: exe, BuildDigest: digest, GenerationID: gen,
		Bootstrap: &pb.Bootstrap{GenerationId: gen, CandidateConfigJson: "{}", TargetConfigVersion: 1,
			DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE},
			LogLevel:           logging.ProtoLevel(out.Levels().Source(pluginID))},
		NewCandidate: isNew, Stdio: stdio, Limits: host.NewLimits(v), Stop: host.NewStopTiming(v), Log: out.Plugin(pluginID, gen),
		Global: global, MemorySoftLimit: v.PluginMemorySoftLimit,
	})
	if err != nil {
		return nil, err
	}
	mode := "restart"
	if isNew {
		mode = "new"
	}
	plog.Info("plugin activated", logging.KeyGenerationID, gen, "mode", mode, "pid", r.Process.Pid())
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
