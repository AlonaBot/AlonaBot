// platform 命令以指定的 Core 与组件产物执行平台验收并保存 JSON 证据。
//
// 验收通过管理 CLI 与 HTTP 驱动真实进程，进程工作目录和数据目录由本命令独立创建。
// 平台与进程所有权遵循 CONTRACTS §7.1、§12.2。
//
// 包的组成：[run] 创建证据文件与独立工作区，[report.business] 驱动管理认证和业务，
// [botServer] 拥有本地 Bot API 服务与请求等待。[coreProcess] 拥有 Core 进程和日志，
// [report.lifecycle] 检查启停、异常恢复及合作型进程树，平台观测由 inspect 文件实现。
package main

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	privateDirectoryMode = 0o700
	privateFileMode      = 0o600
	exitFailure          = 1
	statusPassed         = "passed"
	statusFailed         = "failed"
	statusIncomplete     = "incomplete"
)

type options struct {
	core        string
	packages    string
	scenarios   string
	out         string
	environment string
	host        string // 操作者提供的宿主或虚拟机说明，单独保留声明来源。
}

type candidate struct {
	Path      string               `json:"path"`
	GoVersion string               `json:"go_version"`
	Program   string               `json:"program"`
	Module    debug.Module         `json:"module"`
	Settings  []debug.BuildSetting `json:"settings"`
	OS        string               `json:"os"`
	Arch      string               `json:"arch"`
}

type packageAsset struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

type stepResult struct {
	Name           string           `json:"name"`
	CorePID        int              `json:"core_pid,omitempty"`
	Started        time.Time        `json:"started"`
	DurationMS     float64          `json:"duration_ms"`
	Status         string           `json:"status"`
	Error          string           `json:"error,omitempty"`
	Command        []string         `json:"command,omitempty"`
	Directory      string           `json:"directory,omitempty"`
	Output         string           `json:"output,omitempty"`    // stdout/stderr 合并，完整保留。
	ExitCode       *int             `json:"exit_code,omitempty"` // 启动失败时没有退出码。
	HTTP           *httpObservation `json:"http,omitempty"`
	OutputRedacted bool             `json:"output_redacted,omitempty"` // 仅凭证输出。
}

type report struct {
	Arguments   []string       `json:"arguments"`
	Started     time.Time      `json:"started"`
	DurationMS  float64        `json:"duration_ms"`
	Status      string         `json:"status"`
	Error       string         `json:"error,omitempty"`
	OS          string         `json:"os"`
	Arch        string         `json:"arch"`
	GoVersion   string         `json:"go_version"`
	OSVersion   string         `json:"os_version"`
	Environment string         `json:"environment"`
	Host        string         `json:"host_description"`
	Core        candidate      `json:"core"`
	Packages    string         `json:"packages"`
	Assets      []packageAsset `json:"assets"`
	Scenarios   string         `json:"scenarios"`
	Workspace   string         `json:"workspace"`
	Directory   string         `json:"directory"`
	DataDir     string         `json:"data_dir"`
	Steps       []stepResult   `json:"steps"`
	Runs        []*coreRun     `json:"runs"`
	Components  []candidate    `json:"components"`
	BotCalls    []botCall      `json:"bot_calls"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := run(ctx, os.Args[1:], os.Stdout)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "platform:", err)
		os.Exit(exitFailure)
	}
}

func run(ctx context.Context, args []string, output io.Writer) (err error) {
	var opt options
	flags := flag.NewFlagSet("platform", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&opt.core, "alonabot", "", "candidate Core executable (required)")
	flags.StringVar(&opt.packages, "packages", "", "current-platform example package directory (required)")
	flags.StringVar(&opt.scenarios, "scenarios", "", "scripts/scenarios directory (required)")
	flags.StringVar(&opt.out, "out", "", "new JSON result file; parent directory must exist (required)")
	flags.StringVar(&opt.environment, "environment", "", "execution environment: native, vm, container or emulation (required)")
	flags.StringVar(&opt.host, "host", "", "operator-provided host or VM description (required)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if opt.out == "" {
		return errors.New("--out is required")
	}
	// 1. 先独占证据文件；参数或产物校验失败也留下结果，已有证据始终保留。
	file, err := os.OpenFile(opt.out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, privateFileMode)
	if err != nil {
		return fmt.Errorf("create result: %w", err)
	}
	r := report{
		Arguments: args,
		Started:   time.Now().UTC(), Status: statusIncomplete,
		OS: runtime.GOOS, Arch: runtime.GOARCH, GoVersion: runtime.Version(),
		Environment: opt.environment, Host: opt.host,
	}
	defer func() {
		r.DurationMS = time.Since(r.Started).Seconds() * 1000
		if err != nil {
			r.Status, r.Error = statusFailed, err.Error()
		}
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = errors.Join(err, encoder.Encode(r), file.Close())
	}()
	if err := r.step("inputs", func(_ *stepResult) error {
		if flags.NArg() != 0 || opt.core == "" || opt.packages == "" || opt.scenarios == "" || strings.TrimSpace(opt.host) == "" {
			return errors.New("--alonabot, --packages, --scenarios and --host are required; positional arguments are not accepted")
		}
		switch opt.environment {
		case "native", "vm", "container", "emulation":
		default:
			return errors.New("--environment must be native, vm, container or emulation")
		}
		for _, path := range []*string{&opt.core, &opt.packages, &opt.scenarios} {
			absolute, err := filepath.Abs(*path)
			if err != nil {
				return err
			}
			*path = absolute
		}
		r.Core.Path, r.Packages, r.Scenarios = opt.core, opt.packages, opt.scenarios
		for _, directory := range []string{opt.packages, opt.scenarios} {
			info, err := os.Stat(directory)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("expected a directory: %s", directory)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// 2. 所有待运行命令从独立目录启动，避免被 Core 识别为源码开发工程。
	if err := r.step("workspace", func(_ *stepResult) error {
		workspace, err := os.MkdirTemp("", "alonabot-platform-*")
		if err != nil {
			return err
		}
		r.Workspace = workspace
		r.Directory, r.DataDir = filepath.Join(workspace, "cwd"), filepath.Join(workspace, "data")
		for _, directory := range []string{r.Directory, r.DataDir} {
			if err := os.Mkdir(directory, privateDirectoryMode); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := r.step("candidate", func(_ *stepResult) error {
		build, err := buildinfo.ReadFile(opt.core)
		if err != nil {
			return fmt.Errorf("read candidate build metadata: %w", err)
		}
		r.Core.GoVersion, r.Core.Program, r.Core.Module, r.Core.Settings = build.GoVersion, build.Path, build.Main, build.Settings
		for _, setting := range build.Settings {
			switch setting.Key {
			case "GOOS":
				r.Core.OS = setting.Value
			case "GOARCH":
				r.Core.Arch = setting.Value
			}
		}
		if r.Core.OS != runtime.GOOS || r.Core.Arch != runtime.GOARCH {
			return fmt.Errorf("candidate target %s/%s differs from verifier %s/%s", r.Core.OS, r.Core.Arch, runtime.GOOS, runtime.GOARCH)
		}
		return filepath.WalkDir(opt.packages, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Type().IsRegular() && (strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".zip")) {
				info, err := entry.Info()
				if err != nil {
					return err
				}
				r.Assets = append(r.Assets, packageAsset{Path: path, Bytes: info.Size()})
			}
			return nil
		})
	}); err != nil {
		return err
	}
	// 3. 同时记录系统原始版本和候选的实际执行结果；构建成功单独不足以证明可执行。
	system := []string{"uname", "-a"}
	if runtime.GOOS == "windows" {
		system = []string{"cmd.exe", "/c", "ver"}
	}
	if err := r.command(ctx, "os_version", system...); err != nil {
		return err
	}
	r.OSVersion = strings.TrimSpace(r.Steps[len(r.Steps)-1].Output)
	if err := r.command(ctx, "core_help", opt.core, "--help"); err != nil {
		return err
	}
	return r.business(ctx, opt)
}

func (r *report) step(name string, action func(*stepResult) error) error {
	step := stepResult{Name: name, Started: time.Now().UTC(), Status: statusPassed}
	err := action(&step)
	if len(r.Runs) > 0 {
		step.CorePID = r.Runs[len(r.Runs)-1].PID
	}
	step.DurationMS = time.Since(step.Started).Seconds() * 1000
	if err != nil {
		step.Status, step.Error = statusFailed, err.Error()
	}
	r.Steps = append(r.Steps, step)
	return err
}

func (r *report) command(ctx context.Context, name string, argv ...string) error {
	return r.step(name, func(step *stepResult) error {
		command := exec.CommandContext(ctx, argv[0], argv[1:]...)
		command.Dir = r.Directory
		step.Command, step.Directory = argv, command.Dir
		var output bytes.Buffer
		command.Stdout, command.Stderr = &output, &output
		err := command.Run()
		step.Output = output.String()
		if command.ProcessState != nil {
			code := command.ProcessState.ExitCode()
			step.ExitCode = &code
		}
		if err != nil {
			return fmt.Errorf("run %s: %w", name, err)
		}
		return nil
	})
}
