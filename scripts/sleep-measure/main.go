//go:build linux

// sleep-measure 命令在独立 Core 中采集插件冷启动与 Linux 进程资源。
//
// 构建来源由 <alonabot>.measure.json 记录；--help 给出最小格式。
// cold 每组以原代次执行 100 次 restart；resources/core 每次采集指定轮次。
// 计时与休眠预算见 CONTRACTS §4.5。
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
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	coldAttempts     = 100 // 每组固定真实尝试数；失败编号保留且不补跑。
	privateDir       = 0o700
	privateFile      = 0o600
	startupPoll      = 20 * time.Millisecond // 只观察启动日志，截止由 Core 和调用 context 决定。
	resourceInterval = 2 * time.Minute       // 固定机器验收的 ACTIVE/SUSPENDED 空闲区间。
	kibPerMiB        = 1024
	echoID           = "echo"
	configdbID       = "alonabot-plugin-configdb"
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
	State              string    `json:"state"`
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
	if o.phase == "core" {
		return measureResources(ctx, o, metadata, "core", "", "")
	}
	// 2. 顺序固定为 echo、configdb_gorm；两组各自从新的 Core/data_dir 开始。
	var failures error
	measure := measureCold
	if o.phase == "resources" {
		measure = measureResources
	}
	for _, group := range []struct{ name, id, archive string }{{"echo", echoID, o.echoPackage}, {"configdb_gorm", configdbID, o.configdbPackage}} {
		if ctx.Err() != nil {
			return errors.Join(failures, ctx.Err())
		}
		if err := measure(ctx, o, metadata, group.name, group.id, group.archive); err != nil {
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
		if o.phase == "core" && entry.path == "" {
			continue
		}
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
	raw, err := c.api(ctx, name+"-runtime", http.MethodGet, "/api/v1/plugins/"+url.PathEscape(id)+"/runtime", nil)
	if err != nil {
		return status, err
	}
	if err := json.Unmarshal(raw, &status); err != nil {
		return status, fmt.Errorf("parse runtime: %w: %s", err, raw)
	}
	return status, nil
}

// api 保存每次请求及完整响应；管理凭证只放在 Authorization 头内。
func (c *coreProcess) api(ctx context.Context, name, method, path string, payload []byte) ([]byte, error) {
	token, err := os.ReadFile(filepath.Join(c.dataDir, "admin_token"))
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, method, c.address+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		finished := time.Now()
		return nil, errors.Join(err, appendJSON(filepath.Join(c.out, "api.jsonl"), map[string]any{"name": name, "method": request.Method, "url": request.URL.String(), "request_body": string(payload), "started_at": started.UTC(), "finished_at": finished.UTC(), "duration_ns": finished.Sub(started).Nanoseconds(), "duration_ms": float64(finished.Sub(started)) / float64(time.Millisecond), "error": err.Error()}))
	}
	raw, readErr := io.ReadAll(response.Body)
	finished := time.Now()
	responseFile := name + ".json"
	readErr = errors.Join(readErr, response.Body.Close(), os.WriteFile(filepath.Join(c.out, responseFile), raw, privateFile))
	entry := map[string]any{"name": name, "method": request.Method, "url": request.URL.String(), "request_body": string(payload), "started_at": started.UTC(), "finished_at": finished.UTC(), "duration_ns": finished.Sub(started).Nanoseconds(), "duration_ms": float64(finished.Sub(started)) / float64(time.Millisecond), "status": response.Status, "response_file": responseFile}
	if response.StatusCode != http.StatusOK {
		readErr = errors.Join(readErr, fmt.Errorf("API %s %s: %s: %s", method, path, response.Status, raw))
	}
	if readErr != nil {
		entry["error"] = readErr.Error()
	}
	if err := errors.Join(readErr, appendJSON(filepath.Join(c.out, "api.jsonl"), entry)); err != nil {
		return raw, err
	}
	return raw, nil
}

func (c *coreProcess) installSample(ctx context.Context, group, id, archive string) (runtimeStatus, logRecord, error) {
	// 1. 正式安装与首次迁移在采样之前完成。
	release, err := os.ReadFile(filepath.Join(filepath.Dir(archive), "release.json"))
	if err != nil {
		return runtimeStatus{}, logRecord{}, fmt.Errorf("read package release metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(c.out, "release.json"), release, privateFile); err != nil {
		return runtimeStatus{}, logRecord{}, err
	}
	var publication struct {
		PluginID string `json:"plugin_id"`
	}
	if err := json.Unmarshal(release, &publication); err != nil {
		return runtimeStatus{}, logRecord{}, fmt.Errorf("parse package release metadata: %w", err)
	}
	if publication.PluginID != id {
		return runtimeStatus{}, logRecord{}, fmt.Errorf("package release ID %q does not match %q", publication.PluginID, id)
	}
	if err := c.cliCall(ctx, "install", "install", archive, "--yes"); err != nil {
		return runtimeStatus{}, logRecord{}, err
	}
	if err := c.cliCall(ctx, "enable", "enable", id, "--yes"); err != nil {
		return runtimeStatus{}, logRecord{}, err
	}
	// 2. 运行状态与激活日志共同确定原代次和首次进程。
	initial, err := c.runtime(ctx, id, "initial")
	if err != nil {
		return runtimeStatus{}, logRecord{}, err
	}
	if !initial.Enabled || initial.State != "ACTIVE" || initial.GenerationID == "" {
		return runtimeStatus{}, logRecord{}, fmt.Errorf("initial runtime is not ACTIVE: %+v", initial)
	}
	initialLog, err := c.awaitLog(ctx, func(record logRecord) bool {
		return record.Component == "core" && record.Message == "cold start completed" && record.PluginID == id && record.GenerationID == initial.GenerationID
	})
	if err != nil {
		return runtimeStatus{}, logRecord{}, err
	}
	initialPID := initialLog.PID
	if initialPID <= 0 {
		return runtimeStatus{}, logRecord{}, errors.New("initial activation has no process identity")
	}
	// 3. 实际执行文件的构建信息与输入 sidecar 声明分别保留。
	tags := ""
	if group == "configdb_gorm" {
		tags = "configdb_gorm"
	}
	build, err := readBuild(fmt.Sprintf("/proc/%d/exe", initialPID), tags)
	if err := errors.Join(err, writeJSON(filepath.Join(c.out, "plugin-build.json"), build)); err != nil {
		return runtimeStatus{}, logRecord{}, err
	}
	return initial, initialLog, nil
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
	initial, _, err := c.installSample(ctx, group, id, archive)
	if err != nil {
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

type procStat struct {
	PID         int       `json:"pid"`
	ParentPID   int       `json:"ppid"`
	GroupPID    int       `json:"pgrp"`
	Command     string    `json:"comm"`
	State       string    `json:"state"`
	StartTicks  uint64    `json:"starttime_ticks"`
	UserTicks   uint64    `json:"utime_ticks"`
	SystemTicks uint64    `json:"stime_ticks"`
	ObservedAt  time.Time `json:"observed_at"` // 内存中的值保留单调时钟，CPU 差值按各 PID 实际读时计算。
	Raw         string    `json:"raw"`
	Error       string    `json:"error,omitempty"`
}

type procMemory struct {
	PSSKiB          *uint64 `json:"pss_kib"`
	RSSKiB          *uint64 `json:"rss_kib"`
	PrivateCleanKiB *uint64 `json:"private_clean_kib"`
	PrivateDirtyKiB *uint64 `json:"private_dirty_kib"`
}

type procReading struct {
	Stat          procStat   `json:"stat"`
	Memory        procMemory `json:"memory"`
	SmapsRaw      string     `json:"smaps_rollup_raw"`
	VerifiedStat  procStat   `json:"verified_stat"`
	IdentityError string     `json:"identity_error,omitempty"`
	Error         string     `json:"error,omitempty"`
}

type procSnapshot struct {
	StartedAt  time.Time           `json:"started_at"`
	FinishedAt time.Time           `json:"finished_at"`
	Census     map[int]procStat    `json:"census"` // PPID 树的完整取证；无关进程在枚举期间退出也保留原错误。
	Processes  map[int]procReading `json:"processes"`
	Error      string              `json:"error,omitempty"`
}

type resourceUsage struct {
	PSSMiB     *float64 `json:"pss_mib"`
	RSSMiB     *float64 `json:"rss_mib"`
	PrivateMiB *float64 `json:"private_mib"`
	CPUPercent *float64 `json:"cpu_percent"` // 100% 对应一个逻辑核。
}

type processUsage struct {
	PID       int           `json:"pid"`
	Role      string        `json:"role"`
	Ticks     *uint64       `json:"cpu_ticks"`
	ElapsedNS int64         `json:"elapsed_ns"`
	Usage     resourceUsage `json:"usage"`
	Error     string        `json:"error,omitempty"`
}

type resourceSample struct {
	MachineID   string                   `json:"machine_id"`
	ProductHead string                   `json:"product_head"`
	Group       string                   `json:"group"`
	Round       int                      `json:"round"`
	State       string                   `json:"state"`
	Runtime     *runtimeStatus           `json:"runtime,omitempty"`
	ClockTicks  uint64                   `json:"clock_ticks_per_second"`
	RequestedNS int64                    `json:"requested_interval_ns"`
	ElapsedNS   int64                    `json:"elapsed_ns"`
	Begin       procSnapshot             `json:"begin"`
	End         procSnapshot             `json:"end"`
	Processes   []processUsage           `json:"processes"`
	Totals      map[string]resourceUsage `json:"totals"`
	Result      string                   `json:"result"`
	Error       string                   `json:"error,omitempty"`
}

func parseProcStat(raw string) (procStat, error) {
	s := procStat{Raw: raw}
	open, close := strings.IndexByte(raw, '('), strings.LastIndexByte(raw, ')')
	if open < 0 || close <= open {
		return s, fmt.Errorf("invalid proc stat command: %q", raw)
	}
	var err error
	s.PID, err = strconv.Atoi(strings.TrimSpace(raw[:open]))
	if err != nil || s.PID <= 0 {
		return s, fmt.Errorf("invalid proc stat pid: %q: %w", raw, errors.Join(err, errors.New("positive pid required")))
	}
	s.Command = raw[open+1 : close]
	fields := strings.Fields(raw[close+1:]) // comm 可含空格和括号，余下从第 3 字段 state 开始。
	if len(fields) < 20 {
		return s, fmt.Errorf("proc stat has %d fields after comm; need 20: %q", len(fields), raw)
	}
	s.State = fields[0]
	s.ParentPID, err = strconv.Atoi(fields[1])
	if err != nil || s.ParentPID < 0 {
		return s, fmt.Errorf("invalid proc stat ppid %q: %w", fields[1], errors.Join(err, errors.New("nonnegative ppid required")))
	}
	s.GroupPID, err = strconv.Atoi(fields[2])
	if err != nil || s.GroupPID < 0 {
		return s, fmt.Errorf("invalid proc stat pgrp %q: %w", fields[2], errors.Join(err, errors.New("nonnegative pgrp required")))
	}
	for _, field := range []struct {
		index  int
		target *uint64
	}{{11, &s.UserTicks}, {12, &s.SystemTicks}, {19, &s.StartTicks}} {
		*field.target, err = strconv.ParseUint(fields[field.index], 10, 64)
		if err != nil {
			return s, fmt.Errorf("parse proc stat field %d %q: %w", field.index+3, fields[field.index], err)
		}
	}
	return s, nil
}

func readProcStat(root string, pid int) (procStat, error) {
	path := filepath.Join(root, strconv.Itoa(pid), "stat")
	raw, readErr := os.ReadFile(path)
	observed := time.Now()
	s, parseErr := parseProcStat(string(raw))
	s.ObservedAt = observed
	if readErr != nil {
		parseErr = readErr
	}
	if parseErr == nil && s.PID != pid {
		parseErr = fmt.Errorf("proc stat pid %d differs from path pid %d", s.PID, pid)
	}
	if parseErr != nil {
		parseErr = fmt.Errorf("read %s: %w", path, parseErr)
		s.Error = parseErr.Error()
	}
	return s, parseErr
}

func readProcCensus(root string) (map[int]procStat, error) {
	table := make(map[int]procStat)
	entries, err := os.ReadDir(root)
	if err != nil {
		return table, err
	}
	for _, entry := range entries {
		pid, parseErr := strconv.Atoi(entry.Name())
		if parseErr != nil || pid <= 0 || !entry.IsDir() {
			continue
		}
		s, readErr := readProcStat(root, pid)
		table[pid] = s
		// 无关进程的自然退出保留在 census；其余错误使树成员关系未获确认。
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			err = errors.Join(err, readErr)
		}
	}
	return table, err
}

func procTree(table map[int]procStat, root int) map[int]bool {
	tree := map[int]bool{root: true}
	for changed := true; changed; {
		changed = false
		for pid, s := range table {
			if s.Error == "" && tree[s.ParentPID] && !tree[pid] {
				tree[pid], changed = true, true
			}
		}
	}
	return tree
}

func sameProcess(before, after procStat) error {
	if before.Error != "" || after.Error != "" {
		return fmt.Errorf("process identity unavailable: before=%s; after=%s", before.Error, after.Error)
	}
	if before.PID != after.PID || before.StartTicks != after.StartTicks {
		return fmt.Errorf("process identity changed: pid/starttime %d/%d -> %d/%d", before.PID, before.StartTicks, after.PID, after.StartTicks)
	}
	return nil
}

func parseProcMemory(raw string) (procMemory, error) {
	var memory procMemory
	wanted := map[string]**uint64{"Pss:": &memory.PSSKiB, "Rss:": &memory.RSSKiB, "Private_Clean:": &memory.PrivateCleanKiB, "Private_Dirty:": &memory.PrivateDirtyKiB}
	var failures error
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		target, ok := wanted[fields[0]]
		if !ok {
			continue
		}
		if len(fields) != 3 || fields[2] != "kB" || *target != nil {
			failures = errors.Join(failures, fmt.Errorf("invalid or duplicate smaps_rollup field: %q", line))
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			failures = errors.Join(failures, fmt.Errorf("parse smaps_rollup %q: %w", line, err))
			continue
		}
		*target = &value
	}
	for key, target := range wanted {
		if *target == nil {
			failures = errors.Join(failures, fmt.Errorf("smaps_rollup missing %s", key))
		}
	}
	return memory, failures
}

func readProcSnapshot(root string, core procStat) (procSnapshot, error) {
	// 1. 先保存 PPID 关系并核对本轮 Core 身份。
	snapshot := procSnapshot{StartedAt: time.Now(), Processes: make(map[int]procReading)}
	var err error
	snapshot.Census, err = readProcCensus(root)
	if observed, ok := snapshot.Census[core.PID]; !ok {
		err = errors.Join(err, fmt.Errorf("Core pid %d missing from proc census", core.PID))
	} else {
		err = errors.Join(err, sameProcess(core, observed))
	}
	// 2. 每个 PID 在 smaps 读取后再次核对身份，原文和部分数据一起留下。
	for pid := range procTree(snapshot.Census, core.PID) {
		reading := procReading{Stat: snapshot.Census[pid]}
		raw, readErr := os.ReadFile(filepath.Join(root, strconv.Itoa(pid), "smaps_rollup"))
		reading.SmapsRaw = string(raw)
		memory, parseErr := parseProcMemory(string(raw))
		reading.Memory = memory
		after, identityErr := readProcStat(root, pid)
		reading.VerifiedStat = after
		identityErr = errors.Join(identityErr, sameProcess(reading.Stat, after))
		if identityErr != nil {
			reading.IdentityError = identityErr.Error()
		}
		if failure := errors.Join(readErr, parseErr, identityErr); failure != nil {
			failure = fmt.Errorf("sample pid %d: %w", pid, failure)
			reading.Error = failure.Error()
			err = errors.Join(err, failure)
		}
		snapshot.Processes[pid] = reading
	}
	snapshot.FinishedAt = time.Now()
	if err != nil {
		snapshot.Error = err.Error()
	}
	return snapshot, err
}

func processCPU(before, after procStat, clockTicks uint64) (*uint64, int64, *float64, error) {
	elapsed := after.ObservedAt.Sub(before.ObservedAt).Nanoseconds()
	if err := sameProcess(before, after); err != nil {
		return nil, elapsed, nil, err
	}
	if clockTicks == 0 || elapsed <= 0 || after.UserTicks < before.UserTicks || after.SystemTicks < before.SystemTicks {
		return nil, elapsed, nil, fmt.Errorf("invalid CPU interval: clock_ticks=%d elapsed_ns=%d utime=%d->%d stime=%d->%d", clockTicks, elapsed, before.UserTicks, after.UserTicks, before.SystemTicks, after.SystemTicks)
	}
	ticks := after.UserTicks - before.UserTicks + after.SystemTicks - before.SystemTicks
	percent := float64(ticks) / float64(clockTicks) / (float64(elapsed) / float64(time.Second)) * 100
	return &ticks, elapsed, &percent, nil
}

func memoryUsage(memory procMemory) resourceUsage {
	var usage resourceUsage
	if memory.PSSKiB != nil {
		value := float64(*memory.PSSKiB) / kibPerMiB
		usage.PSSMiB = &value
	}
	if memory.RSSKiB != nil {
		value := float64(*memory.RSSKiB) / kibPerMiB
		usage.RSSMiB = &value
	}
	if memory.PrivateCleanKiB != nil && memory.PrivateDirtyKiB != nil {
		value := (float64(*memory.PrivateCleanKiB) + float64(*memory.PrivateDirtyKiB)) / kibPerMiB
		usage.PrivateMiB = &value
	}
	return usage
}

// summarizeResources 只汇总完整的对应读数；某 PID 缺项时该列合计保持 null。
func summarizeResources(sample *resourceSample, corePID, pluginPID int) error {
	roles := map[string][]resourceUsage{"core": {}, "auxiliary": {}, "core_with_auxiliary": {}, "total": {}}
	if sample.Group != "core" {
		roles["plugin_tree"] = nil
	}
	pluginTree := procTree(sample.Begin.Census, pluginPID)
	pids := make(map[int]bool)
	for pid := range sample.Begin.Processes {
		pids[pid] = true
	}
	for pid := range sample.End.Processes {
		pids[pid] = true
	}
	ordered := make([]int, 0, len(pids))
	for pid := range pids {
		ordered = append(ordered, pid)
	}
	slices.Sort(ordered)
	var failures error
	for _, pid := range ordered {
		before, beforeOK := sample.Begin.Processes[pid]
		after, afterOK := sample.End.Processes[pid]
		process := processUsage{PID: pid, Role: "auxiliary"}
		if pid == corePID {
			process.Role = "core"
		} else if pluginPID != 0 && pluginTree[pid] {
			process.Role = "plugin_tree"
		}
		var err error
		if !beforeOK || !afterOK {
			err = fmt.Errorf("pid %d missing at interval boundary: begin=%t end=%t", pid, beforeOK, afterOK)
		} else {
			process.Ticks, process.ElapsedNS, process.Usage.CPUPercent, err = processCPU(before.Stat, after.Stat, sample.ClockTicks)
			if identityErr := errors.Join(sameProcess(before.Stat, after.Stat), errorText(before.IdentityError), errorText(after.IdentityError)); identityErr != nil {
				err = errors.Join(err, identityErr)
				process.Usage = resourceUsage{}
			} else {
				memory := memoryUsage(after.Memory)
				process.Usage.PSSMiB, process.Usage.RSSMiB, process.Usage.PrivateMiB = memory.PSSMiB, memory.RSSMiB, memory.PrivateMiB
			}
			err = errors.Join(err, errorText(before.Error), errorText(after.Error))
		}
		if err != nil {
			process.Error = err.Error()
			failures = errors.Join(failures, fmt.Errorf("pid %d: %w", pid, err))
		}
		sample.Processes = append(sample.Processes, process)
		roles[process.Role] = append(roles[process.Role], process.Usage)
		roles["total"] = append(roles["total"], process.Usage)
		if process.Role != "plugin_tree" {
			roles["core_with_auxiliary"] = append(roles["core_with_auxiliary"], process.Usage)
		}
	}
	sample.Totals = make(map[string]resourceUsage)
	for role, usages := range roles {
		sample.Totals[role] = sumUsage(usages)
	}
	return failures
}

func errorText(value string) error {
	if value == "" {
		return nil
	}
	return errors.New(value)
}

func sumUsage(usages []resourceUsage) resourceUsage {
	pss, rss, private, cpu := 0.0, 0.0, 0.0, 0.0
	total := resourceUsage{&pss, &rss, &private, &cpu}
	for _, usage := range usages {
		for _, pair := range []struct {
			target **float64
			value  *float64
		}{{&total.PSSMiB, usage.PSSMiB}, {&total.RSSMiB, usage.RSSMiB}, {&total.PrivateMiB, usage.PrivateMiB}, {&total.CPUPercent, usage.CPUPercent}} {
			if pair.value == nil {
				*pair.target = nil
			} else if *pair.target != nil {
				**pair.target += *pair.value
			}
		}
	}
	return total
}

// exitedTree 同时核对已见 PID 身份和原进程组，覆盖停止期间重新挂到其他父进程的子进程。
func exitedTree(table map[int]procStat, old []procStat, groupPID int) error {
	var failures error
	for _, previous := range old {
		if current, ok := table[previous.PID]; ok {
			if current.Error != "" {
				failures = errors.Join(failures, fmt.Errorf("old pid %d exit identity unavailable: %s", previous.PID, current.Error))
			} else if err := sameProcess(previous, current); err == nil {
				failures = errors.Join(failures, fmt.Errorf("old plugin pid %d starttime %d still exists (state %s)", previous.PID, previous.StartTicks, current.State))
			}
		}
	}
	for pid, current := range table {
		if current.Error == "" && current.GroupPID == groupPID {
			failures = errors.Join(failures, fmt.Errorf("plugin process group %d still contains pid %d", groupPID, pid))
		}
	}
	return failures
}

func (c *coreProcess) resourceInterval(ctx context.Context, base resourceSample, id, generation string, core, plugin procStat, old []procStat) (sample resourceSample, err error) {
	sample = base
	sample.RequestedNS, sample.Result = resourceInterval.Nanoseconds(), "failed"
	name := strings.ToLower(sample.State)
	defer func() {
		if err != nil {
			sample.Error = err.Error()
		} else {
			sample.Result = "success"
		}
		err = errors.Join(err, writeJSON(filepath.Join(c.out, name+".json"), sample))
	}()
	// 1. 区间两端保留整棵树和逐 PID 原文；中途取消仍留下起点。
	started := time.Now()
	sample.Begin, err = readProcSnapshot("/proc", core)
	err = errors.Join(err, writeJSON(filepath.Join(c.out, name+"-begin.json"), sample.Begin))
	if sample.State == "SUSPENDED" {
		err = errors.Join(err, exitedTree(sample.Begin.Census, old, plugin.GroupPID))
	}
	if sample.State == "ACTIVE" && id != "" {
		current, ok := sample.Begin.Processes[plugin.PID]
		if !ok {
			err = errors.Join(err, errors.New("ACTIVE plugin missing from Core tree"))
		} else {
			err = errors.Join(err, sameProcess(plugin, current.Stat))
		}
	}
	if err != nil {
		return sample, err
	}
	timer := time.NewTimer(time.Until(started.Add(resourceInterval)))
	defer timer.Stop()
	select { // 只等待规定的观测区间；Core 提前退出或调用取消会保留未完成状态。
	case <-ctx.Done():
		err = ctx.Err()
	case <-c.done:
		err = fmt.Errorf("Core exited during resource interval: %w", errors.Join(c.waitErr, errors.New("interval incomplete")))
	case <-timer.C:
	}
	// 2. 终点内存与同区间 CPU 分别计算，缺项沿对应列传播。
	var readErr error
	sample.End, readErr = readProcSnapshot("/proc", core)
	sample.ElapsedNS = time.Since(started).Nanoseconds()
	err = errors.Join(err, readErr)
	pluginPID := plugin.PID
	var treeErr error
	if sample.State == "SUSPENDED" {
		pluginPID = 0
		treeErr = exitedTree(sample.End.Census, old, plugin.GroupPID)
		err = errors.Join(err, treeErr)
	}
	err = errors.Join(err, summarizeResources(&sample, core.PID, pluginPID))
	if id != "" {
		status, statusErr := c.runtime(ctx, id, name+"-end")
		sample.Runtime = &status
		if statusErr == nil && (!status.Enabled || status.GenerationID != generation || status.State != sample.State) {
			statusErr = fmt.Errorf("resource runtime changed: expected %s generation %s, got %+v", sample.State, generation, status)
		}
		err = errors.Join(err, statusErr)
		if sample.State == "SUSPENDED" {
			treeErr = errors.Join(treeErr, statusErr)
		}
	}
	if treeErr != nil {
		// 退出或状态取证失败时整树零值失去依据，保留已测 Core/附属值。
		sample.Totals["plugin_tree"] = resourceUsage{}
		sample.Totals["total"] = sumUsage([]resourceUsage{sample.Totals["core_with_auxiliary"], sample.Totals["plugin_tree"]})
	}
	return sample, err
}

func measureResources(ctx context.Context, o options, metadata runMetadata, group, id, archive string) (err error) {
	c, err := startCore(ctx, o, group)
	if err != nil {
		return err
	}
	completed := make(map[string]bool)
	defer func() {
		err = errors.Join(err, c.stop())
		summary := map[string]any{"group": group, "round": o.round, "machine_id": metadata.Machine.ID, "product_head": metadata.Provenance.ProductHead, "completed_intervals": completed, "result": "success"}
		if err != nil {
			summary["result"], summary["error"] = "failed", err.Error()
		}
		err = errors.Join(err, writeJSON(filepath.Join(c.out, "summary.json"), summary))
	}()
	// 1. 只使用机器实际 CLK_TCK，Core 与插件均保持缺省 tune。
	if err := command(ctx, "getconf", c.out, "clock-ticks", "CLK_TCK"); err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(c.out, "clock-ticks.log"))
	if err != nil {
		return err
	}
	clockTicks, err := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil || clockTicks == 0 {
		return fmt.Errorf("invalid getconf CLK_TCK %q: %w", raw, errors.Join(err, errors.New("positive clock frequency required")))
	}
	core, identityErr := readProcStat("/proc", c.command.Process.Pid)
	if err := errors.Join(identityErr, writeJSON(filepath.Join(c.out, "core-identity.json"), core)); err != nil {
		return err
	}
	if _, err := c.api(ctx, "initial-sleep", http.MethodGet, "/api/v1/core/sleep", nil); err != nil {
		return err
	}
	base := resourceSample{MachineID: metadata.Machine.ID, ProductHead: metadata.Provenance.ProductHead, Group: group, Round: o.round, ClockTicks: clockTicks, State: "ACTIVE"}
	if id == "" {
		_, err = c.resourceInterval(ctx, base, "", "", core, procStat{}, nil)
		completed["ACTIVE"] = err == nil
		return err
	}
	initial, activation, err := c.installSample(ctx, group, id, archive)
	if err != nil {
		return err
	}
	plugin, identityErr := readProcStat("/proc", activation.PID)
	if identityErr == nil && plugin.GroupPID != plugin.PID {
		identityErr = fmt.Errorf("plugin pid %d does not own expected process group: %d", plugin.PID, plugin.GroupPID)
	}
	if err := errors.Join(identityErr, writeJSON(filepath.Join(c.out, "plugin-identity.json"), plugin)); err != nil {
		return err
	}
	policyPath := "/api/v1/plugins/" + url.PathEscape(id) + "/sleep-policy"
	if _, err := c.api(ctx, "initial-policy", http.MethodGet, policyPath, nil); err != nil {
		return err
	}
	active, err := c.resourceInterval(ctx, base, id, initial.GenerationID, core, plugin, nil)
	completed["ACTIVE"] = err == nil
	if err != nil {
		return err
	}
	// 2. ACTIVE 测完再开启 component_idle；策略写回不重置已有活动时间。
	waitingSince := time.Now()
	var policyErr error
	if id == echoID {
		_, policyErr = c.api(ctx, "set-policy", http.MethodPut, policyPath, []byte(`{"policy":"15m"}`))
	}
	_, modeErr := c.api(ctx, "set-sleep", http.MethodPut, "/api/v1/core/sleep", []byte(`{"mode":"component_idle","idle_timeout":"15m"}`))
	_, sleepErr := c.api(ctx, "effective-sleep", http.MethodGet, "/api/v1/core/sleep", nil)
	policyRaw, effectiveErr := c.api(ctx, "effective-policy", http.MethodGet, policyPath, nil)
	if err := errors.Join(policyErr, modeErr, sleepErr, effectiveErr); err != nil {
		return err
	}
	var policy struct {
		EffectivePolicy string `json:"effective_policy"`
		Eligible        bool   `json:"eligible"`
	}
	if err := json.Unmarshal(policyRaw, &policy); err != nil {
		return fmt.Errorf("parse effective sleep policy: %w: %s", err, policyRaw)
	}
	duration, err := time.ParseDuration(policy.EffectivePolicy)
	if err != nil || !policy.Eligible || duration != 15*time.Minute {
		return fmt.Errorf("unexpected effective sleep policy: %s: %w", policyRaw, errors.Join(err, errors.New("eligible 15m policy required")))
	}
	transition, err := c.awaitLog(ctx, func(record logRecord) bool {
		return record.Component == "core" && record.Message == "plugin runtime state changed" && record.PluginID == id && !record.Time.Before(waitingSince) && (record.State == "SUSPENDED" || record.State == "FAILED" || record.State == "STOPPED")
	})
	if err := errors.Join(err, writeJSON(filepath.Join(c.out, "suspend-transition.json"), transition)); err != nil {
		return err
	}
	status, err := c.runtime(ctx, id, "suspended")
	if err != nil {
		return err
	}
	if !status.Enabled || status.GenerationID != initial.GenerationID || status.State != "SUSPENDED" {
		return fmt.Errorf("automatic suspend did not preserve enabled generation: %+v", status)
	}
	// 3. 旧树全退的独立证据先落盘，随后才允许记休眠树为零。
	var old []procStat
	for pid := range procTree(active.End.Census, plugin.PID) {
		old = append(old, active.End.Census[pid])
	}
	slices.SortFunc(old, func(a, b procStat) int { return a.PID - b.PID })
	census, censusErr := readProcCensus("/proc")
	exitErr := errors.Join(censusErr, exitedTree(census, old, plugin.GroupPID))
	exitRecord := map[string]any{"checked_at": time.Now().UTC(), "old_tree": old, "old_process_group": plugin.GroupPID, "census": census, "exited": exitErr == nil}
	if exitErr != nil {
		exitRecord["error"] = exitErr.Error()
	}
	if err := errors.Join(exitErr, writeJSON(filepath.Join(c.out, "old-tree-exit.json"), exitRecord)); err != nil {
		return err
	}
	base.State = "SUSPENDED"
	_, err = c.resourceInterval(ctx, base, id, initial.GenerationID, core, plugin, old)
	completed["SUSPENDED"] = err == nil
	return err
}
