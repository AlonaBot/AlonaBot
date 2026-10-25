// alonabot 命令启动 AlonaBot Core。
//
// 目前只提供开发用的 run 子命令：在给定的 data_dir 中登记并启动一个本地插件，
// 收到 SIGINT 或 SIGTERM 时停止它；插件异常退出时按有限重启策略重启或禁用。
// 正式的安装与管理命令尚未提供。
package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
)

// 进程退出码。
const (
	exitFailure = 1 // 运行失败
	exitUsage   = 2 // 参数错误
)

// main 解析 run 子命令的参数并运行；参数错误以 exitUsage 退出，运行失败以 exitFailure 退出。
func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Fprintln(os.Stderr, "usage: alonabot run --data-dir DIR --plugin-id ID --entry PATH [--stdio]")
		os.Exit(exitUsage)
	}
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	dataDir := fs.String("data-dir", "", "Core data directory (only one Core per directory)")
	pluginID := fs.String("plugin-id", "", "plugin ID")
	entry := fs.String("entry", "", "plugin executable path (development mode: a local build)")
	stdio := fs.Bool("stdio", false, "use stdio transport (default: TCP loopback)")
	_ = fs.Parse(os.Args[2:])
	if *dataDir == "" || *pluginID == "" || *entry == "" {
		fs.Usage()
		os.Exit(exitUsage)
	}
	if err := run(*dataDir, *pluginID, *entry, *stdio); err != nil {
		fmt.Fprintln(os.Stderr, "alonabot:", err)
		os.Exit(exitFailure)
	}
}

// run 打开 data_dir、登记开发安装记录，然后启动插件；收到 SIGINT/SIGTERM 时停止它，
// 异常退出按有限重启策略重启或禁用。
func run(dataDir, pluginID, entry string, stdio bool) error {
	// 1. 打开状态库（取得 core.lock），定位入口并计算摘要。
	ctx := context.Background()
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		return err
	}
	defer store.Close()

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
	policy := host.NewRestartPolicy()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	for {
		r, err := launch(ctx, store, pluginID, pluginDir, exe, digest, stdio)
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
			fmt.Fprintf(os.Stderr, "alonabot: plugin %s exited (%s)\n", pluginID, st)
			exitc <- exit{abnormal, err}
		}()
		select {
		case <-sigc:
			res := r.Stop(time.Now(), pb.StopReason_STOP_REASON_CORE_SHUTDOWN)
			fmt.Fprintf(os.Stderr, "alonabot: stopped plugin %s (drained=%v killed=%v)\n", pluginID, res.Drained, res.Killed)
			if res.Err != nil {
				return fmt.Errorf("process tree of plugin %s not confirmed empty: %w", pluginID, res.Err)
			}
			return nil
		case e := <-exitc:
			if e.err != nil {
				return fmt.Errorf("process tree of plugin %s not confirmed empty, not starting a replacement: %w", pluginID, e.err)
			}
			switch policy.OnExit(pluginID, e.abnormal, time.Now()) {
			case host.DecisionRestart:
				fmt.Fprintf(os.Stderr, "alonabot: plugin %s exited abnormally, restarting\n", pluginID)
				continue
			case host.DecisionDisable:
				_ = store.SetEnabled(ctx, pluginID, false)
				return fmt.Errorf("plugin %s exited abnormally for the 3rd time within 5 minutes and has been disabled", pluginID)
			default:
				return nil
			}
		}
	}
}

// launch 首次运行时提交新代次；之后复用已提交的 Active 代次（普通 restart，不产生新 Generation）。
func launch(ctx context.Context, store *state.Store, pluginID, pluginDir, exe string, digest []byte, stdio bool) (*supervisor.Running, error) {
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
			DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE}},
		NewCandidate: isNew, Stdio: stdio, Limits: host.DefaultLimits, Stop: host.DefaultStopTiming,
	})
	if err != nil {
		return nil, err
	}
	mode := "restart"
	if isNew {
		mode = "new"
	}
	fmt.Fprintf(os.Stderr, "alonabot: plugin %s activated generation=%s mode=%s pid=%d\n", pluginID, gen, mode, r.Process.Pid())
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
