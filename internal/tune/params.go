package tune

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

// Values 保存一次运行的全部生效调优值。
type Values struct {
	SessionMaxInflight    int
	SessionMaxPayload     int64 // 字节
	CoreMaxPayload        int64 // 字节
	RequestTimeout        time.Duration
	CallChainMaxDepth     int
	DispatchConcurrency   int
	DispatchQueueEvents   int
	DispatchQueueBytes    int64 // 字节
	ContinuationMax       int
	ContinuationTimeout   time.Duration
	BusQueueEvents        int
	BusQueueBytes         int64 // 字节
	MediaMaxObject        int64 // 字节
	MediaMaxUploads       int
	MediaSessionBudget    int64 // 字节
	MediaDiskBudget       int64 // 字节
	MediaTTL              time.Duration
	PhaseHello            time.Duration
	PhaseConfig           time.Duration
	PhaseMigration        time.Duration
	PhaseOnStart          time.Duration
	PhaseOnReady          time.Duration
	StopTotal             time.Duration
	StopDrain             time.Duration
	StopOnStop            time.Duration
	RestartWindow         time.Duration
	RestartMaxCrashes     int
	TasksPerPlugin        int
	TaskPayloadBytes      int64 // 字节
	TaskConcurrency       int
	JobTimeout            time.Duration
	ButtonTokenTTL        time.Duration
	ToolConfirmTimeout    time.Duration
	LogBufferRecords      int
	InstallMaxArchive     int64 // 字节
	InstallMaxExpanded    int64 // 字节
	InstallMaxEntries     int
	DevReloadPoll         time.Duration
	PluginMemorySoftLimit int64 // 字节；0 表示不设置
	CoreMemorySoftLimit   int64 // 字节；0 表示不设置
	OptionsPage           int
	WakeQueueRequests     int
	WakeQueueBytes        int64 // 每插件等待唤醒的载荷字节数
	SleepMinUptime        time.Duration
	SleepColdStartWindow  time.Duration
	SleepColdStartMax     int
}

const maxOptionsPage = 200 // 配置选项单页的固定上限。

// Override 记录一项最终生效的覆盖及其来源。
type Override struct {
	Name   string
	Value  string
	Origin string // 直接参数为 --args，文件参数为路径:行号
}

type unit string

const (
	integerUnit  unit = "integer"
	durationUnit unit = "duration"
	bytesUnit    unit = "bytes"
	// maxFrame 与 wire.MaxFrame 相同；此包独立于协议运行时。
	maxFrame int64 = 4 << 20
)

type parameter struct {
	name         string
	unit         unit
	defaultValue int64
	positive     bool // 覆盖值须大于零；缺省0表示未设置的内存软上限除外
	constraint   string
	location     string
	get          func(*Values) int64
	set          func(*Values, int64)
}

func param[T ~int | ~int64](name string, unit unit, value T, positive bool, constraint, location string, field func(*Values) *T) parameter {
	return parameter{name: name, unit: unit, defaultValue: int64(value), positive: positive, constraint: constraint, location: location,
		get: func(v *Values) int64 { return int64(*field(v)) },
		set: func(v *Values, n int64) { *field(v) = T(n) },
	}
}

// parameters 是名称、单位、默认值与约束的源表，顺序同时决定 tune list 的输出。
var parameters = []parameter{
	param("SESSION_MAX_INFLIGHT", integerUnit, int(256), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.SessionMaxInflight }),
	param("SESSION_MAX_PAYLOAD", bytesUnit, int64(16<<20), true, ">= frame limit 4MiB", "internal/tune/params.go", func(v *Values) *int64 { return &v.SessionMaxPayload }),
	param("CORE_MAX_PAYLOAD", bytesUnit, int64(64<<20), true, ">= SESSION_MAX_PAYLOAD", "internal/tune/params.go", func(v *Values) *int64 { return &v.CoreMaxPayload }),
	param("REQUEST_TIMEOUT", durationUnit, time.Duration(30*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.RequestTimeout }),
	param("CALL_CHAIN_MAX_DEPTH", integerUnit, int(16), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.CallChainMaxDepth }),
	param("DISPATCH_CONCURRENCY", integerUnit, int(64), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.DispatchConcurrency }),
	param("DISPATCH_QUEUE_EVENTS", integerUnit, int(256), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.DispatchQueueEvents }),
	param("DISPATCH_QUEUE_BYTES", bytesUnit, int64(8<<20), true, ">= frame limit 4MiB", "internal/tune/params.go", func(v *Values) *int64 { return &v.DispatchQueueBytes }),
	param("CONTINUATION_MAX", integerUnit, int(128), true, ">= 1 and < SESSION_MAX_INFLIGHT", "internal/tune/params.go", func(v *Values) *int { return &v.ContinuationMax }),
	param("CONTINUATION_TIMEOUT", durationUnit, time.Duration(15*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.ContinuationTimeout }),
	param("BUS_QUEUE_EVENTS", integerUnit, int(256), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.BusQueueEvents }),
	param("BUS_QUEUE_BYTES", bytesUnit, int64(4<<20), true, ">= frame limit 4MiB", "internal/tune/params.go", func(v *Values) *int64 { return &v.BusQueueBytes }),
	param("MEDIA_MAX_OBJECT", bytesUnit, int64(100<<20), false, "<= MEDIA_SESSION_BUDGET", "internal/tune/params.go", func(v *Values) *int64 { return &v.MediaMaxObject }),
	param("MEDIA_MAX_UPLOADS", integerUnit, int(4), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.MediaMaxUploads }),
	param("MEDIA_SESSION_BUDGET", bytesUnit, int64(256<<20), false, "<= MEDIA_DISK_BUDGET", "internal/tune/params.go", func(v *Values) *int64 { return &v.MediaSessionBudget }),
	param("MEDIA_DISK_BUDGET", bytesUnit, int64(2<<30), true, "> 0", "internal/tune/params.go", func(v *Values) *int64 { return &v.MediaDiskBudget }),
	param("MEDIA_TTL", durationUnit, time.Duration(15*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.MediaTTL }),
	param("PHASE_HELLO", durationUnit, time.Duration(10*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.PhaseHello }),
	param("PHASE_CONFIG", durationUnit, time.Duration(60*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.PhaseConfig }),
	param("PHASE_MIGRATION", durationUnit, time.Duration(15*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.PhaseMigration }),
	param("PHASE_ONSTART", durationUnit, time.Duration(60*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.PhaseOnStart }),
	param("PHASE_ONREADY", durationUnit, time.Duration(30*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.PhaseOnReady }),
	param("STOP_TOTAL", durationUnit, time.Duration(10*time.Second), true, ">= STOP_DRAIN + STOP_ONSTOP", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.StopTotal }),
	param("STOP_DRAIN", durationUnit, time.Duration(6*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.StopDrain }),
	param("STOP_ONSTOP", durationUnit, time.Duration(4*time.Second), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.StopOnStop }),
	param("RESTART_WINDOW", durationUnit, time.Duration(5*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.RestartWindow }),
	param("RESTART_MAX_CRASHES", integerUnit, int(3), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.RestartMaxCrashes }),
	param("SLEEP_MIN_UPTIME", durationUnit, time.Duration(5*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.SleepMinUptime }),
	param("SLEEP_COLD_START_WINDOW", durationUnit, time.Duration(time.Hour), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.SleepColdStartWindow }),
	param("SLEEP_COLD_START_MAX", integerUnit, int(3), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.SleepColdStartMax }),
	param("JOB_TIMEOUT", durationUnit, time.Duration(15*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.JobTimeout }),
	param("TASKS_PER_PLUGIN", integerUnit, int(10000), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.TasksPerPlugin }),
	param("TASK_PAYLOAD_BYTES", bytesUnit, int64(64<<10), true, "1..4MiB", "internal/tune/params.go", func(v *Values) *int64 { return &v.TaskPayloadBytes }),
	param("TASK_CONCURRENCY", integerUnit, int(16), true, "1..SESSION_MAX_INFLIGHT", "internal/tune/params.go", func(v *Values) *int { return &v.TaskConcurrency }),
	param("BUTTON_TOKEN_TTL", durationUnit, time.Duration(24*time.Hour), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.ButtonTokenTTL }),
	param("TOOL_CONFIRM_TIMEOUT", durationUnit, time.Duration(5*time.Minute), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.ToolConfirmTimeout }),
	param("LOG_BUFFER_RECORDS", integerUnit, int(1000), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.LogBufferRecords }),
	param("INSTALL_MAX_ARCHIVE", bytesUnit, int64(256<<20), true, "> 0", "internal/tune/params.go", func(v *Values) *int64 { return &v.InstallMaxArchive }),
	param("INSTALL_MAX_EXPANDED", bytesUnit, int64(1<<30), true, ">= INSTALL_MAX_ARCHIVE", "internal/tune/params.go", func(v *Values) *int64 { return &v.InstallMaxExpanded }),
	param("INSTALL_MAX_ENTRIES", integerUnit, int(10000), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.InstallMaxEntries }),
	param("DEV_RELOAD_POLL", durationUnit, time.Duration(500*time.Millisecond), true, "> 0", "internal/tune/params.go", func(v *Values) *time.Duration { return &v.DevReloadPoll }),
	param("PLUGIN_MEMORY_SOFT_LIMIT", bytesUnit, int64(0), true, "> 0", "internal/tune/params.go", func(v *Values) *int64 { return &v.PluginMemorySoftLimit }),
	param("CORE_MEMORY_SOFT_LIMIT", bytesUnit, int64(0), true, "> 0", "internal/tune/params.go", func(v *Values) *int64 { return &v.CoreMemorySoftLimit }),
	param("OPTIONS_PAGE", integerUnit, int(50), true, ">= 1 and <= 200", "internal/tune/params.go", func(v *Values) *int { return &v.OptionsPage }),
	param("WAKE_QUEUE_REQUESTS", integerUnit, int(256), true, ">= 1", "internal/tune/params.go", func(v *Values) *int { return &v.WakeQueueRequests }),
	param("WAKE_QUEUE_BYTES", bytesUnit, int64(16<<20), true, ">= frame limit 4MiB and <= CORE_MAX_PAYLOAD", "internal/tune/params.go", func(v *Values) *int64 { return &v.WakeQueueBytes }),
}

// Defaults 从 parameters 构造全部参数的默认值。
func Defaults() Values {
	var v Values
	for _, p := range parameters {
		p.set(&v, p.defaultValue)
	}
	return v
}

// WriteList 按参数表顺序写出名称、单位、缺省值、生效值、约束与契约位置。
func WriteList(w io.Writer, v Values) error {
	if _, err := fmt.Fprintln(w, "NAME\tUNIT\tDEFAULT\tEFFECTIVE\tCONSTRAINT\tCONTRACT"); err != nil {
		return err
	}
	for _, p := range parameters {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", p.name, p.unit, p.format(p.defaultValue), p.format(p.get(&v)), p.constraint, p.location); err != nil {
			return err
		}
	}
	return nil
}

func (p parameter) format(n int64) string {
	switch p.unit {
	case durationUnit:
		return time.Duration(n).String()
	case bytesUnit:
		if n == 0 {
			if p.defaultValue == 0 {
				return "unset"
			}
			return "0B"
		}
		for _, suffix := range []struct {
			name string
			size int64
		}{{"GiB", 1 << 30}, {"MiB", 1 << 20}, {"KiB", 1 << 10}} {
			if n%suffix.size == 0 {
				return strconv.FormatInt(n/suffix.size, 10) + suffix.name
			}
		}
		return strconv.FormatInt(n, 10) + "B"
	default:
		return strconv.FormatInt(n, 10)
	}
}
