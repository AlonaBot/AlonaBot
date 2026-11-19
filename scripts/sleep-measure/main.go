//go:build linux

// sleep-measure 命令在独立 Core 中采集插件冷启动的原始日志与统计。
//
// 构建来源由 <alonabot>.measure.json 记录；--help 给出最小格式。
// 每组先安装和启用，再以原代次执行 100 次 restart；计时口径见 CONTRACTS §4.5。
package main

import (
	"bufio"
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"
	"time"
)

const (
	coldAttempts = 100 // 每组固定真实尝试数；失败编号保留且不补跑。
	privateDir   = 0o700
	privateFile  = 0o600
	startupPoll  = 20 * time.Millisecond // 只观察启动日志，截止由 Core 和调用 context 决定。
	echoID       = "echo"
	configdbID   = "alonabot-plugin-configdb"
)

type options struct {
	alonabot, echoPackage, configdbPackage, workDir, out, phase string
	round                                                       int
}

type artifactSource struct {
	Path      string            `json:"path"`
	BuildArgv [][]string        `json:"build_argv"` // 实际执行的完整 argv，只记录，不执行。
	BuildEnv  map[string]string `json:"build_env"`  // 构建命令显式设置的环境；无覆盖时写 {}。
}

type provenance struct {
	ProductHead             string         `json:"product_head"`
	Core                    artifactSource `json:"core"`
	Echo                    artifactSource `json:"echo"`
	ConfigDB                artifactSource `json:"configdb"`
	FileCache               string         `json:"file_cache"`
	EnvironmentLimits       string         `json:"environment_limits"`
	PreviousMachineRelation string         `json:"previous_machine_relation"`
}

type machine struct {
	ID         string            `json:"id"`
	Hostname   string            `json:"hostname"`
	OS         string            `json:"os"`
	Arch       string            `json:"arch"`
	LogicalCPU int               `json:"logical_cpus"`
	GoVersion  string            `json:"tool_go_version"`
	Kernel     string            `json:"kernel"`
	Raw        map[string]string `json:"raw"` // 保留 CPU 型号、总内存、cgroup 与实际进程限制原文。
}

type runMetadata struct {
	StartedAt  time.Time        `json:"started_at"`
	FinishedAt time.Time        `json:"finished_at"`
	Command    []string         `json:"command"`
	Phase      string           `json:"phase"`
	Round      int              `json:"round"`
	Machine    machine          `json:"machine"`
	Provenance provenance       `json:"provenance"`
	CoreBuild  *debug.BuildInfo `json:"core_build"`
	Result     string           `json:"result"`
	Error      string           `json:"error,omitempty"`
}

type logRecord struct {
	Time               time.Time `json:"time"`
	Component          string    `json:"component"`
	Message            string    `json:"msg"`
	PluginID           string    `json:"plugin_id"`
	GenerationID       string    `json:"generation_id"`
	SessionID          string    `json:"session_id"`
	LifecycleSessionID string    `json:"lifecycle_session_id"`
	PID                int       `json:"pid"`
	Listen             string    `json:"listen"`
	ActivationReason   string    `json:"activation_reason"`
	DurationNS         *int64    `json:"duration_ns"`
	ElapsedNS          *int64    `json:"elapsed_ns"`
	PrepareNS          *int64    `json:"prepare_ns"`
	ActivateNS         *int64    `json:"activate_ns"`
	Phase              string    `json:"phase"`
	Code               string    `json:"code"`
	Executed           *bool     `json:"executed"`
	Outcome            string    `json:"outcome"`
	ObservedNS         *int64    `json:"observed_ns"`
	Error              string    `json:"error"`
	CallbackError      string    `json:"err"`
	Stage              string    `json:"stage"`
}

type commandRecord struct {
	Name       string            `json:"name"`
	Argv       []string          `json:"argv"`
	Env        map[string]string `json:"env_overrides"`
	PID        int               `json:"pid"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt time.Time         `json:"finished_at"`
	DurationNS int64             `json:"duration_ns"`
	DurationMS float64           `json:"duration_ms"`
	ExitCode   int               `json:"exit_code"`
	Error      string            `json:"error,omitempty"`
}

type runtimeStatus struct {
	Enabled      bool   `json:"enabled"`
	GenerationID string `json:"generation_id"`
	State        string `json:"state"`
	Reason       string `json:"reason"`
}

type attempt struct {
	Group      string        `json:"group"`
	Number     int           `json:"number"`
	StartedAt  time.Time     `json:"command_started_at"`
	FinishedAt time.Time     `json:"command_finished_at"`
	DurationNS int64         `json:"command_duration_ns"`
	DurationMS float64       `json:"command_duration_ms"`
	Runtime    runtimeStatus `json:"runtime"`
	Error      string        `json:"error,omitempty"`
}

type callback struct {
	Executed   *bool    `json:"executed"`
	Outcome    string   `json:"outcome"`
	DurationNS *int64   `json:"duration_ns"`
	DurationMS *float64 `json:"duration_ms"`
	ObservedNS *int64   `json:"observed_ns"`
	ObservedMS *float64 `json:"observed_ms"`
	Missing    string   `json:"missing,omitempty"`
	Error      string   `json:"error,omitempty"`
	Invalid    string   `json:"invalid,omitempty"`
}

type coldSample struct {
	Attempt          attempt             `json:"attempt"`
	MachineID        string              `json:"machine_id"`
	ProductHead      string              `json:"product_head"`
	PluginID         string              `json:"plugin_id"`
	GenerationID     string              `json:"generation_id"`
	SessionID        string              `json:"session_id"`
	PID              int                 `json:"pid"`
	StartedAt        *time.Time          `json:"started_at"`
	FinishedAt       *time.Time          `json:"finished_at"`
	TimeBasis        string              `json:"time_basis"`
	ActivationReason string              `json:"activation_reason"`
	ColdStartNS      *int64              `json:"cold_start_ns"`
	ColdStartMS      *float64            `json:"cold_start_ms"`
	PrepareNS        *int64              `json:"prepare_ns"`
	PrepareMS        *float64            `json:"prepare_ms"`
	ActivateNS       *int64              `json:"activate_ns"`
	ActivateMS       *float64            `json:"activate_ms"`
	FailureElapsedNS *int64              `json:"failure_elapsed_ns"`
	FailureElapsedMS *float64            `json:"failure_elapsed_ms"`
	FailureStage     string              `json:"failure_stage,omitempty"`
	FailurePhase     string              `json:"failure_phase,omitempty"`
	FailureCode      string              `json:"failure_code,omitempty"`
	CoreRecords      []logRecord         `json:"core_records"`
	Callbacks        map[string]callback `json:"callbacks"`
	Result           string              `json:"result"`
	Errors           []string            `json:"errors,omitempty"`
	Missing          map[string]string   `json:"missing,omitempty"`
}

type distribution struct {
	N       int      `json:"n"`
	Partial bool     `json:"partial"`
	P50NS   *int64   `json:"p50_ns"`
	P95NS   *int64   `json:"p95_ns"`
	P99NS   *int64   `json:"p99_ns"`
	P50MS   *float64 `json:"p50_ms"`
	P95MS   *float64 `json:"p95_ms"`
	P99MS   *float64 `json:"p99_ms"`
}

type coldSummary struct {
	Group          string                  `json:"group"`
	Attempts       int                     `json:"attempts"`
	Successes      int                     `json:"successes"`
	Failures       int                     `json:"failures"`
	FailureRate    float64                 `json:"failure_rate"`
	Complete       bool                    `json:"complete"`
	Distributions  map[string]distribution `json:"distributions"`
	EchoP95Over200 *bool                   `json:"echo_p95_over_200ms,omitempty"`
	Error          string                  `json:"error,omitempty"`
}

// coreProcess 持有本次启动的 Core；done 关闭后发布 waitErr，日志随后可完整离线关联。
type coreProcess struct {
	command *exec.Cmd
	done    chan struct{}
	waitErr error
	stderr  *os.File
	stdout  *os.File
	dataDir string
	address string
	cli     string
	out     string
	record  commandRecord
	started time.Time // 保留单调时钟，JSON 时间戳另取 UTC。
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "sleep-measure:", err)
		os.Exit(1)
	}
}

func parseOptions(args []string) (options, error) {
	var o options
	flags := flag.NewFlagSet("sleep-measure", flag.ContinueOnError)
	flags.StringVar(&o.alonabot, "alonabot", "", "absolute path to the ordinary AlonaBot build")
	flags.StringVar(&o.echoPackage, "echo-package", "", "echo archive with adjacent release.json")
	flags.StringVar(&o.configdbPackage, "configdb-package", "", "configdb_gorm archive with adjacent release.json")
	flags.StringVar(&o.workDir, "work-dir", "", "new work directory for isolated Core data")
	flags.StringVar(&o.out, "out", "", "new directory for raw evidence")
	flags.StringVar(&o.phase, "phase", "", "cold, resources or core")
	flags.IntVar(&o.round, "round", 0, "resource or Core round: 1, 2 or 3")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: sleep-measure --alonabot <absolute CLI> --echo-package <archive> --configdb-package <archive> --work-dir <new directory> --out <new directory> --phase <cold|resources|core> [--round <1|2|3>]")
		flags.PrintDefaults()
		fmt.Fprintln(flags.Output(), `
Required sidecar: <absolute CLI path>.measure.json
All artifacts share product_head. Record actual argv and explicit environment
overrides, including GOOS/GOARCH/CGO_ENABLED and configdb_gorm. Commands run from
the product checkout. These declarations are retained separately from observed
Go build information; the tool never executes sidecar commands.
Example (replace paths, head, environment and observations with actual values):
{
  "product_head": "<fixed product commit>",
  "core": {"path": "/build/alonabot", "build_argv": [["go", "build", "-o", "/build/alonabot", "./cmd/alonabot"]], "build_env": {"GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0"}},
  "echo": {"path": "/build/echo/echo-0.1.0-linux-amd64.tar.gz", "build_argv": [["/build/alonabot", "pack", "./examples/plugins/echo", "--target", "linux/amd64", "--out", "/build/echo"]], "build_env": {"CGO_ENABLED": "0"}},
  "configdb": {"path": "/build/configdb/alonabot-plugin-configdb-0.1.0-linux-amd64.tar.gz", "build_argv": [["/build/alonabot", "pack", "./examples/plugins/configdb", "--tags", "configdb_gorm", "--target", "linux/amd64", "--out", "/build/configdb"]], "build_env": {"CGO_ENABLED": "0"}},
  "file_cache": "<actual cache conditions; echo then configdb_gorm>",
  "environment_limits": "<actual machine/container limits>",
  "previous_machine_relation": "<relation to V-33/V-74 machines>"
}`)
	}
	if err := flags.Parse(args); err != nil {
		return o, err
	}
	if flags.NArg() != 0 || !filepath.IsAbs(o.alonabot) || o.workDir == "" || o.out == "" {
		return o, errors.New("--alonabot must be absolute; --work-dir and --out are required; positional arguments are unsupported")
	}
	if o.phase != "cold" && o.phase != "resources" && o.phase != "core" {
		return o, errors.New("--phase must be cold, resources or core")
	}
	if o.phase == "cold" && o.round != 0 || o.phase != "cold" && (o.round < 1 || o.round > 3) {
		return o, errors.New("--round is required only for resources/core and must be 1, 2 or 3")
	}
	if o.phase != "core" && (o.echoPackage == "" || o.configdbPackage == "") {
		return o, errors.New("--echo-package and --configdb-package are required")
	}
	for _, path := range []*string{&o.alonabot, &o.echoPackage, &o.configdbPackage, &o.workDir, &o.out} {
		if *path != "" {
			absolute, err := filepath.Abs(*path)
			if err != nil {
				return o, err
			}
			*path = absolute
		}
	}
	return o, nil
}

func run(ctx context.Context, args []string) (err error) {
	o, err := parseOptions(args)
	if err != nil {
		return err
	}
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		return errors.New("measurement requires linux/amd64")
	}
	if o.phase != "cold" {
		return fmt.Errorf("phase %s is not implemented", o.phase)
	}
	// 1. 新目录由本次运行独占；失败时保留原始输入、日志与已完成的尝试。
	if err := os.Mkdir(o.out, privateDir); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	metadata := runMetadata{StartedAt: time.Now().UTC(), Command: append([]string{os.Args[0]}, args...), Phase: o.phase, Round: o.round, Result: "failed"}
	defer func() {
		metadata.FinishedAt = time.Now().UTC()
		if err != nil {
			metadata.Error = err.Error()
		} else {
			metadata.Result = "success"
		}
		err = errors.Join(err, writeJSON(filepath.Join(o.out, "run.json"), metadata))
	}()
	if err := os.Mkdir(o.workDir, privateDir); err != nil {
		return fmt.Errorf("create work directory: %w", err)
	}
	metadata.Machine, err = readMachine(ctx)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(o.alonabot + ".measure.json")
	if err != nil {
		return fmt.Errorf("read build provenance: %w", err)
	}
	if err := os.WriteFile(filepath.Join(o.out, "build-provenance.json"), raw, privateFile); err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &metadata.Provenance); err != nil {
		return fmt.Errorf("parse build provenance: %w", err)
	}
	if err := checkProvenance(o, metadata.Provenance); err != nil {
		return err
	}
	metadata.CoreBuild, err = readBuild(o.alonabot, "")
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(o.out, "run.json"), metadata); err != nil {
		return err
	}
	// 2. 顺序固定为 echo、configdb_gorm；两组各自从新的 Core/data_dir 开始。
	var failures error
	for _, group := range []struct{ name, id, archive string }{{"echo", echoID, o.echoPackage}, {"configdb_gorm", configdbID, o.configdbPackage}} {
		if ctx.Err() != nil {
			return errors.Join(failures, ctx.Err())
		}
		if err := measureCold(ctx, o, metadata, group.name, group.id, group.archive); err != nil {
			failures = errors.Join(failures, fmt.Errorf("measure %s: %w", group.name, err))
		}
	}
	return failures
}

func checkProvenance(o options, p provenance) error {
	if p.ProductHead == "" || p.FileCache == "" || p.EnvironmentLimits == "" || p.PreviousMachineRelation == "" {
		return errors.New("build provenance requires product_head, file_cache, environment_limits and previous_machine_relation")
	}
	for _, entry := range []struct {
		path   string
		source artifactSource
	}{{o.alonabot, p.Core}, {o.echoPackage, p.Echo}, {o.configdbPackage, p.ConfigDB}} {
		actual, err := filepath.Abs(entry.path)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(entry.source.Path) || actual != filepath.Clean(entry.source.Path) || len(entry.source.BuildArgv) == 0 || entry.source.BuildEnv == nil {
			return fmt.Errorf("build provenance for %s requires matching absolute path, build_argv and build_env", entry.path)
		}
		for _, argv := range entry.source.BuildArgv {
			if len(argv) == 0 || argv[0] == "" {
				return fmt.Errorf("build provenance for %s has empty build argv", entry.path)
			}
		}
	}
	return nil
}

func readMachine(ctx context.Context) (machine, error) {
	m := machine{OS: runtime.GOOS, Arch: runtime.GOARCH, LogicalCPU: runtime.NumCPU(), GoVersion: runtime.Version(), Raw: make(map[string]string)}
	var err error
	m.Hostname, err = os.Hostname()
	if err != nil {
		return m, err
	}
	for _, path := range []string{"/etc/machine-id", "/proc/cpuinfo", "/proc/meminfo", "/proc/self/limits", "/proc/self/cgroup"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			return m, fmt.Errorf("read machine %s: %w", path, err)
		}
		m.Raw[path] = string(raw)
	}
	m.ID = strings.TrimSpace(m.Raw["/etc/machine-id"])
	kernel, err := exec.CommandContext(ctx, "uname", "-srvm").CombinedOutput()
	if err != nil {
		return m, fmt.Errorf("uname -srvm: %w: %s", err, kernel)
	}
	m.Kernel = strings.TrimSpace(string(kernel))
	return m, nil
}

func readBuild(path, tags string) (*debug.BuildInfo, error) {
	build, err := buildinfo.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read build %s: %w", path, err)
	}
	settings := make(map[string]string)
	for _, setting := range build.Settings {
		settings[setting.Key] = setting.Value
	}
	if settings["GOOS"] != "linux" || settings["GOARCH"] != "amd64" || settings["-race"] == "true" {
		return build, fmt.Errorf("%s must be an ordinary linux/amd64 build", path)
	}
	if tags != "" && !slices.Contains(strings.Split(settings["-tags"], ","), tags) {
		return build, fmt.Errorf("%s is missing build tag %s", path, tags)
	}
	return build, nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), privateFile)
}

func command(ctx context.Context, cli, out, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, cli, args...)
	// 独立测量使用缺省 tune；管理凭证经受保护的 token-file 传递。
	cmd.Env = append(os.Environ(), "ALONABOT_ARGS=", "ALONABOT_ADMIN_TOKEN=")
	started := time.Now()
	output, runErr := cmd.CombinedOutput()
	finished := time.Now()
	writeErr := os.WriteFile(filepath.Join(out, name+".log"), output, privateFile)
	if runErr != nil {
		runErr = fmt.Errorf("%s %s: %w: %s", cli, strings.Join(args, " "), runErr, output)
	}
	record := commandRecord{Name: name, Argv: cmd.Args, Env: map[string]string{"ALONABOT_ARGS": "", "ALONABOT_ADMIN_TOKEN": ""}, StartedAt: started.UTC(), FinishedAt: finished.UTC(), DurationNS: finished.Sub(started).Nanoseconds(), ExitCode: -1}
	record.DurationMS = *milliseconds(&record.DurationNS)
	if cmd.Process != nil {
		record.PID = cmd.Process.Pid
	}
	if cmd.ProcessState != nil {
		record.ExitCode = cmd.ProcessState.ExitCode()
	}
	if runErr != nil {
		record.Error = runErr.Error()
	}
	return errors.Join(runErr, writeErr, appendJSON(filepath.Join(out, "commands.jsonl"), record))
}

func appendJSON(path string, value any) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, privateFile)
	if err != nil {
		return err
	}
	return errors.Join(json.NewEncoder(file).Encode(value), file.Close())
}

func startCore(ctx context.Context, o options, group string) (_ *coreProcess, err error) {
	c := &coreProcess{cli: o.alonabot, dataDir: filepath.Join(o.workDir, group), out: filepath.Join(o.out, group), done: make(chan struct{})}
	if err := os.Mkdir(c.out, privateDir); err != nil {
		return nil, err
	}
	if err := os.Mkdir(c.dataDir, privateDir); err != nil {
		return nil, err
	}
	settings := map[string]any{"listen": "127.0.0.1:0", "base_url": "http://127.0.0.1", "sleep": map[string]any{"mode": "always_on", "idle_timeout": "15m"}, "log": map[string]any{"level": "debug", "format": "json", "file": map[string]any{"enabled": true}}}
	if err := writeJSON(filepath.Join(c.dataDir, "core_config.json"), settings); err != nil {
		return nil, err
	}
	if err := writeJSON(filepath.Join(c.out, "core-config-input.json"), settings); err != nil {
		return nil, err
	}
	c.stderr, err = os.OpenFile(filepath.Join(c.out, "core-stderr.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, privateFile)
	if err != nil {
		return nil, err
	}
	c.stdout, err = os.OpenFile(filepath.Join(c.out, "core-stdout.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, privateFile)
	if err != nil {
		return nil, errors.Join(err, c.stderr.Close())
	}
	c.command = exec.Command(o.alonabot, "run", "--data-dir", c.dataDir)
	c.command.Stdout, c.command.Stderr = c.stdout, c.stderr
	c.command.Env = append(os.Environ(), "ALONABOT_ARGS=", "ALONABOT_ADMIN_TOKEN=")
	c.started = time.Now()
	c.record = commandRecord{Name: "core", Argv: c.command.Args, Env: map[string]string{"ALONABOT_ARGS": "", "ALONABOT_ADMIN_TOKEN": ""}, StartedAt: c.started.UTC(), ExitCode: -1}
	if err := c.command.Start(); err != nil {
		finished := time.Now()
		c.record.FinishedAt = finished.UTC()
		c.record.DurationNS = finished.Sub(c.started).Nanoseconds()
		c.record.DurationMS = *milliseconds(&c.record.DurationNS)
		c.record.Error = err.Error()
		return nil, errors.Join(err, writeJSON(filepath.Join(c.out, "core-process.json"), c.record), c.stderr.Close(), c.stdout.Close())
	}
	c.record.PID = c.command.Process.Pid
	go func() { c.waitErr = c.command.Wait(); close(c.done) }() // 关闭发布 Wait 的结果，保证离线读取前所有输出已写完。
	defer func() {
		if err != nil {
			err = errors.Join(err, c.stop())
		}
	}()
	if err := writeJSON(filepath.Join(c.out, "core-process.json"), c.record); err != nil {
		return nil, err
	}
	if err := c.awaitStarted(ctx); err != nil {
		return nil, err
	}
	if err := command(ctx, o.alonabot, c.out, "effective-tune", "tune", "list"); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *coreProcess) awaitStarted(ctx context.Context) error {
	_, err := c.awaitLog(ctx, func(record logRecord) bool {
		if record.Component != "core" {
			return false
		}
		if record.Message == "HTTP listener started" {
			c.address = "http://" + record.Listen
		}
		return record.Message == "Core started"
	})
	if err == nil && c.address == "" {
		err = errors.New("Core started without HTTP listener address")
	}
	return err
}

// awaitLog 只解析已换行的记录，避免运行中的 stderr 尾行尚未写完。
func (c *coreProcess) awaitLog(ctx context.Context, match func(logRecord) bool) (logRecord, error) {
	file, err := os.Open(c.stderr.Name())
	if err != nil {
		return logRecord{}, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var pending []byte
	for {
		line, readErr := reader.ReadBytes('\n')
		pending = append(pending, line...)
		if len(line) > 0 && line[len(line)-1] == '\n' {
			var record logRecord
			if err := json.Unmarshal(pending, &record); err != nil {
				return logRecord{}, fmt.Errorf("parse Core log: %w: %s", err, pending)
			}
			pending = nil
			if match(record) {
				return record, nil
			}
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return logRecord{}, readErr
		}
		if readErr == nil {
			continue
		}
		select { // 等待下一条启动记录或实际进程结束，不重试任何业务操作。
		case <-ctx.Done():
			return logRecord{}, ctx.Err()
		case <-c.done:
			return logRecord{}, fmt.Errorf("Core exited before expected log: %w; partial log: %s", errors.Join(c.waitErr, errors.New("expected log absent")), pending)
		case <-time.After(startupPoll):
		}
	}
}

func (c *coreProcess) stop() error {
	var signalErr error
	select {
	case <-c.done:
	default:
		signalErr = c.command.Process.Signal(syscall.SIGTERM)
	}
	if errors.Is(signalErr, os.ErrProcessDone) {
		signalErr = nil
	}
	<-c.done // Core 自己执行既有停止预算并回收插件树。
	finished := time.Now()
	c.record.FinishedAt = finished.UTC()
	c.record.DurationNS = finished.Sub(c.started).Nanoseconds()
	c.record.DurationMS = *milliseconds(&c.record.DurationNS)
	c.record.ExitCode = c.command.ProcessState.ExitCode()
	if err := errors.Join(signalErr, c.waitErr); err != nil {
		c.record.Error = err.Error()
	}
	config, configErr := os.ReadFile(filepath.Join(c.dataDir, "core_config.json"))
	if configErr == nil {
		configErr = os.WriteFile(filepath.Join(c.out, "core-config-final.json"), config, privateFile)
	}
	return errors.Join(signalErr, c.waitErr, c.stderr.Close(), c.stdout.Close(), configErr, writeJSON(filepath.Join(c.out, "core-process.json"), c.record))
}

func (c *coreProcess) cliCall(ctx context.Context, name, action string, args ...string) error {
	argv := append([]string{action}, args...)
	argv = append(argv, "--data-dir", c.dataDir, "--admin-url", c.address, "--token-file", filepath.Join(c.dataDir, "admin_token"))
	return command(ctx, c.cli, c.out, name, argv...)
}

func (c *coreProcess) runtime(ctx context.Context, id, name string) (runtimeStatus, error) {
	var status runtimeStatus
	token, err := os.ReadFile(filepath.Join(c.dataDir, "admin_token"))
	if err != nil {
		return status, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/api/v1/plugins/"+url.PathEscape(id)+"/runtime", nil)
	if err != nil {
		return status, err
	}
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		finished := time.Now()
		return status, errors.Join(err, appendJSON(filepath.Join(c.out, "api.jsonl"), map[string]any{"name": name, "method": request.Method, "url": request.URL.String(), "started_at": started.UTC(), "finished_at": finished.UTC(), "duration_ns": finished.Sub(started).Nanoseconds(), "duration_ms": float64(finished.Sub(started)) / float64(time.Millisecond), "error": err.Error()}))
	}
	raw, readErr := io.ReadAll(response.Body)
	finished := time.Now()
	responseFile := name + "-runtime.json"
	readErr = errors.Join(readErr, response.Body.Close(), os.WriteFile(filepath.Join(c.out, responseFile), raw, privateFile))
	entry := map[string]any{"name": name, "method": request.Method, "url": request.URL.String(), "started_at": started.UTC(), "finished_at": finished.UTC(), "duration_ns": finished.Sub(started).Nanoseconds(), "duration_ms": float64(finished.Sub(started)) / float64(time.Millisecond), "status": response.Status, "response_file": responseFile}
	if readErr != nil {
		entry["error"] = readErr.Error()
	}
	if err := errors.Join(readErr, appendJSON(filepath.Join(c.out, "api.jsonl"), entry)); err != nil {
		return status, err
	}
	if response.StatusCode != http.StatusOK {
		return status, fmt.Errorf("runtime API %s: %s", response.Status, raw)
	}
	if err := json.Unmarshal(raw, &status); err != nil {
		return status, fmt.Errorf("parse runtime: %w: %s", err, raw)
	}
	return status, nil
}

func measureCold(ctx context.Context, o options, metadata runMetadata, group, id, archive string) (err error) {
	c, err := startCore(ctx, o, group)
	if err != nil {
		return err
	}
	stopped := false
	defer func() {
		if !stopped {
			err = errors.Join(err, c.stop())
		}
	}()
	// 1. 正式安装与首次迁移在全部编号样本之前完成。
	release, err := os.ReadFile(filepath.Join(filepath.Dir(archive), "release.json"))
	if err != nil {
		return fmt.Errorf("read package release metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(c.out, "release.json"), release, privateFile); err != nil {
		return err
	}
	var publication struct {
		PluginID string `json:"plugin_id"`
	}
	if err := json.Unmarshal(release, &publication); err != nil {
		return fmt.Errorf("parse package release metadata: %w", err)
	}
	if publication.PluginID != id {
		return fmt.Errorf("package release ID %q does not match %q", publication.PluginID, id)
	}
	if err := c.cliCall(ctx, "install", "install", archive, "--yes"); err != nil {
		return err
	}
	if err := c.cliCall(ctx, "enable", "enable", id, "--yes"); err != nil {
		return err
	}
	initial, err := c.runtime(ctx, id, "initial")
	if err != nil {
		return err
	}
	if !initial.Enabled || initial.State != "ACTIVE" || initial.GenerationID == "" {
		return fmt.Errorf("initial runtime is not ACTIVE: %+v", initial)
	}
	initialLog, err := c.awaitLog(ctx, func(record logRecord) bool {
		return record.Component == "core" && record.Message == "cold start completed" && record.PluginID == id && record.GenerationID == initial.GenerationID
	})
	if err != nil {
		return err
	}
	initialPID := initialLog.PID
	if initialPID <= 0 {
		return errors.New("initial activation has no process identity")
	}
	tags := ""
	if group == "configdb_gorm" {
		tags = "configdb_gorm"
	}
	build, err := readBuild(fmt.Sprintf("/proc/%d/exe", initialPID), tags)
	if err := errors.Join(err, writeJSON(filepath.Join(c.out, "plugin-build.json"), build)); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(c.out, "initial-runtime.json"), initial); err != nil {
		return err
	}
	attemptFile, err := os.OpenFile(filepath.Join(c.out, "attempts.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, privateFile)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, attemptFile.Close()) }()
	encoder := json.NewEncoder(attemptFile)
	var attempts []attempt
	// 2. 失败仍占用原编号，后续 restart 为独立样本，不补跑失败编号。
	for number := 1; number <= coldAttempts; number++ {
		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}
		started := time.Now()
		callErr := c.cliCall(ctx, fmt.Sprintf("restart-%03d", number), "restart", id, "--yes")
		finished := time.Now()
		status, statusErr := c.runtime(ctx, id, fmt.Sprintf("restart-%03d", number))
		item := attempt{Group: group, Number: number, StartedAt: started.UTC(), FinishedAt: finished.UTC(), DurationNS: finished.Sub(started).Nanoseconds(), DurationMS: float64(finished.Sub(started)) / float64(time.Millisecond), Runtime: status}
		if statusErr == nil && (!status.Enabled || status.State != "ACTIVE" || status.GenerationID != initial.GenerationID) {
			statusErr = fmt.Errorf("restart runtime differs from original enabled ACTIVE generation: %+v", status)
		}
		if failure := errors.Join(callErr, statusErr); failure != nil {
			item.Error = failure.Error()
		}
		attempts = append(attempts, item)
		if writeErr := encoder.Encode(item); writeErr != nil {
			err = errors.Join(err, writeErr)
			break
		}
	}
	// 3. 正常停止 Core 后读完整日志，避免插件日志跨管道稍晚到达而误记缺项。
	stopErr := c.stop()
	stopped = true
	records, readErr := readRecords(c.stderr.Name())
	err = errors.Join(err, stopErr, readErr)
	samples := correlateCold(attempts, records, metadata.Machine.ID, metadata.Provenance.ProductHead, id, initial.GenerationID)
	file, writeErr := os.OpenFile(filepath.Join(c.out, "cold.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, privateFile)
	if writeErr != nil {
		return errors.Join(err, writeErr)
	}
	for _, sample := range samples {
		if writeErr := json.NewEncoder(file).Encode(sample); writeErr != nil {
			err = errors.Join(err, writeErr)
			break
		}
	}
	err = errors.Join(err, file.Close())
	summary := summarizeCold(group, samples)
	summary.Complete = summary.Complete && err == nil
	if !summary.Complete {
		err = errors.Join(err, fmt.Errorf("incomplete cold batch: attempts=%d successes=%d failures=%d", summary.Attempts, summary.Successes, summary.Failures))
	}
	if err != nil {
		summary.Error = err.Error()
	}
	return errors.Join(err, writeJSON(filepath.Join(c.out, "summary.json"), summary))
}

func readRecords(path string) (records []logRecord, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	reader := bufio.NewReader(file)
	var parseErr error
	lineNumber := 0
	for {
		line, readErr := reader.ReadBytes('\n')
		lineNumber++
		if len(bytes.TrimSpace(line)) != 0 {
			var record logRecord
			if err := json.Unmarshal(line, &record); err != nil {
				parseErr = errors.Join(parseErr, fmt.Errorf("parse log %s line %d: %w: %s", path, lineNumber, err, line))
			} else {
				records = append(records, record)
			}
		}
		if errors.Is(readErr, io.EOF) {
			return records, parseErr
		}
		if readErr != nil {
			return records, errors.Join(parseErr, readErr)
		}
	}
}

func milliseconds(value *int64) *float64 {
	if value == nil {
		return nil
	}
	converted := float64(*value) / float64(time.Millisecond)
	return &converted
}

func correlateCold(attempts []attempt, records []logRecord, machineID, head, id, generation string) []coldSample {
	samples := make([]coldSample, 0, len(attempts))
	seen := make(map[string]bool)
	for _, item := range attempts {
		sample := coldSample{Attempt: item, MachineID: machineID, ProductHead: head, PluginID: id, GenerationID: generation, Result: "failed", Callbacks: make(map[string]callback), Missing: make(map[string]string)}
		if item.Error != "" {
			sample.Errors = append(sample.Errors, item.Error)
		}
		var matches []logRecord
		for _, record := range records {
			if record.Component == "core" && record.PluginID == id && (record.Message == "cold start completed" || record.Message == "cold start failed") && !record.Time.Before(item.StartedAt) && !record.Time.After(item.FinishedAt) {
				matches = append(matches, record)
			}
		}
		if len(matches) != 1 {
			sample.Missing["cold_start"] = fmt.Sprintf("expected one cold-start record in restart interval, got %d", len(matches))
		} else {
			record := matches[0]
			sample.GenerationID, sample.SessionID, sample.PID = record.GenerationID, record.SessionID, record.PID
			sample.ActivationReason = record.ActivationReason
			sample.FinishedAt = &record.Time
			sample.TimeBasis = "log timestamp; start derived by subtracting Core monotonic elapsed duration"
			elapsed := record.ElapsedNS
			if record.Message == "cold start completed" {
				sample.ColdStartNS = record.DurationNS
				elapsed = record.DurationNS
			}
			if elapsed != nil {
				started := record.Time.Add(-time.Duration(*elapsed))
				sample.StartedAt = &started
			}
			sample.PrepareNS, sample.ActivateNS = record.PrepareNS, record.ActivateNS
			sample.ColdStartMS, sample.PrepareMS, sample.ActivateMS = milliseconds(sample.ColdStartNS), milliseconds(sample.PrepareNS), milliseconds(sample.ActivateNS)
			if record.GenerationID != generation || record.ActivationReason != "resume" || record.SessionID == "" || record.PID <= 0 || seen[record.SessionID] {
				sample.Errors = append(sample.Errors, "cold-start identity/reason does not match one new Session of the original generation")
			}
			seen[record.SessionID] = true
			for key, value := range map[string]*int64{"cold_start_ns": sample.ColdStartNS, "prepare_ns": sample.PrepareNS, "activate_ns": sample.ActivateNS} {
				if value == nil {
					sample.Missing[key] = "phase did not complete or its measurement was not logged"
				} else if *value < 0 {
					sample.Errors = append(sample.Errors, "negative "+key)
				}
			}
		}
		sample.CoreRecords = matches
		for _, record := range matches {
			if record.Message == "cold start failed" {
				sample.FailureElapsedNS, sample.FailureElapsedMS = record.ElapsedNS, milliseconds(record.ElapsedNS)
				sample.FailureStage, sample.FailurePhase, sample.FailureCode = record.Stage, record.Phase, record.Code
				if record.Error != "" {
					sample.Errors = append(sample.Errors, record.Error)
				}
			}
		}
		for _, phase := range []string{"OnStart", "OnReady"} {
			observation := readCallback(records, id, generation, sample.SessionID, phase)
			if observation.Executed == nil || (*observation.Executed && observation.DurationNS == nil) {
				sample.Missing[phase] = observation.Missing
			}
			if observation.Invalid != "" {
				sample.Errors = append(sample.Errors, phase+": "+observation.Invalid)
			}
			if observation.Error != "" {
				sample.Errors = append(sample.Errors, phase+": "+observation.Error)
			}
			sample.Callbacks[phase] = observation
		}
		if len(sample.Errors) == 0 && len(sample.Missing) == 0 && item.Runtime.Enabled && item.Runtime.State == "ACTIVE" && item.Runtime.GenerationID == generation {
			sample.Result = "success"
		}
		samples = append(samples, sample)
	}
	return samples
}

func readCallback(records []logRecord, id, generation, session, phase string) callback {
	observation := callback{Missing: "callback log absent for this Session; process may have exited before callback completion"}
	completed := 0
	for _, record := range records {
		if session == "" || record.Component != "plugin" || record.PluginID != id || record.GenerationID != generation || record.LifecycleSessionID != session || record.Phase != phase {
			continue
		}
		switch record.Message {
		case "lifecycle callback deadline observed":
			observation.ObservedNS = record.ObservedNS
		case "lifecycle callback completed":
			completed++
			observation.Executed, observation.Outcome, observation.DurationNS, observation.Error = record.Executed, record.Outcome, record.DurationNS, record.CallbackError
		}
	}
	// 截止观测与真实返回相互独立，日志到达顺序不会覆盖实际返回的结果。
	observation.ObservedMS = milliseconds(observation.ObservedNS)
	if completed == 0 {
		if observation.ObservedNS != nil {
			observation.Outcome = "timeout"
			observation.Missing = "deadline observed; actual callback return was not logged before Core exit"
		}
		return observation
	}
	if completed != 1 || observation.Executed == nil {
		observation.Invalid = fmt.Sprintf("expected one completed callback with executed field, got %d", completed)
		return observation
	}
	if !*observation.Executed {
		observation.Missing = "callback not configured; excluded from callback duration distribution"
		if observation.Outcome != "not_executed" || observation.DurationNS != nil {
			observation.Invalid = "unexecuted callback has an invalid outcome or duration"
		}
		return observation
	}
	if observation.DurationNS == nil {
		observation.Missing = "executed callback has no actual-return duration"
		return observation
	}
	observation.Missing = ""
	observation.DurationMS = milliseconds(observation.DurationNS)
	if *observation.DurationNS < 0 || observation.Outcome != "success" {
		observation.Invalid = "callback outcome=" + observation.Outcome + " or duration is negative"
	}
	return observation
}

func nearestRank(values []int64, q float64) *int64 {
	if len(values) == 0 {
		return nil
	}
	ordered := slices.Clone(values)
	slices.Sort(ordered)
	value := ordered[int(math.Ceil(q*float64(len(ordered))))-1]
	return &value
}

func summarizeCold(group string, samples []coldSample) coldSummary {
	summary := coldSummary{Group: group, Attempts: len(samples), Distributions: make(map[string]distribution)}
	values := map[string][]int64{"cold_start": {}, "prepare": {}, "activate": {}, "OnStart": {}, "OnReady": {}}
	for _, sample := range samples {
		if sample.Result != "success" {
			summary.Failures++
			continue
		}
		summary.Successes++
		for key, value := range map[string]*int64{"cold_start": sample.ColdStartNS, "prepare": sample.PrepareNS, "activate": sample.ActivateNS, "OnStart": sample.Callbacks["OnStart"].DurationNS, "OnReady": sample.Callbacks["OnReady"].DurationNS} {
			if value != nil {
				values[key] = append(values[key], *value)
			}
		}
	}
	if summary.Attempts != 0 {
		summary.FailureRate = float64(summary.Failures) / float64(summary.Attempts)
	}
	summary.Complete = summary.Attempts == coldAttempts && summary.Successes == coldAttempts
	for key, series := range values {
		d := distribution{N: len(series), Partial: len(series) != coldAttempts, P50NS: nearestRank(series, .50), P95NS: nearestRank(series, .95), P99NS: nearestRank(series, .99)}
		d.P50MS, d.P95MS, d.P99MS = milliseconds(d.P50NS), milliseconds(d.P95NS), milliseconds(d.P99NS)
		summary.Distributions[key] = d
	}
	if group == "echo" && summary.Distributions["cold_start"].P95NS != nil {
		over := *summary.Distributions["cold_start"].P95NS > int64(200*time.Millisecond)
		summary.EchoP95Over200 = &over
	}
	return summary
}
