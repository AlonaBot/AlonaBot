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
}

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
	constraint   string
	location     string
	get          func(*Values) int64
	set          func(*Values, int64)
}

func param[T ~int | ~int64](name string, unit unit, value T, constraint, location string, field func(*Values) *T) parameter {
	return parameter{name: name, unit: unit, defaultValue: int64(value), constraint: constraint, location: location,
		get: func(v *Values) int64 { return int64(*field(v)) },
		set: func(v *Values, n int64) { *field(v) = T(n) },
	}
}

// parameters 的顺序与 CONTRACTS §16 一致。
var parameters = []parameter{
	param("SESSION_MAX_INFLIGHT", integerUnit, int(256), ">= 1", "§1.2", func(v *Values) *int { return &v.SessionMaxInflight }),
	param("SESSION_MAX_PAYLOAD", bytesUnit, int64(16<<20), ">= frame limit 4MiB", "§1.2", func(v *Values) *int64 { return &v.SessionMaxPayload }),
	param("CORE_MAX_PAYLOAD", bytesUnit, int64(64<<20), ">= SESSION_MAX_PAYLOAD", "§1.2", func(v *Values) *int64 { return &v.CoreMaxPayload }),
	param("REQUEST_TIMEOUT", durationUnit, time.Duration(30*time.Second), "> 0", "§1.2", func(v *Values) *time.Duration { return &v.RequestTimeout }),
	param("CALL_CHAIN_MAX_DEPTH", integerUnit, int(16), ">= 1", "§3", func(v *Values) *int { return &v.CallChainMaxDepth }),
	param("DISPATCH_CONCURRENCY", integerUnit, int(64), ">= 1", "§2.1", func(v *Values) *int { return &v.DispatchConcurrency }),
	param("DISPATCH_QUEUE_EVENTS", integerUnit, int(256), ">= 1", "§2.1", func(v *Values) *int { return &v.DispatchQueueEvents }),
	param("DISPATCH_QUEUE_BYTES", bytesUnit, int64(8<<20), ">= frame limit 4MiB", "§2.1", func(v *Values) *int64 { return &v.DispatchQueueBytes }),
	param("CONTINUATION_MAX", integerUnit, int(128), ">= 1 and < SESSION_MAX_INFLIGHT", "§2.4", func(v *Values) *int { return &v.ContinuationMax }),
	param("CONTINUATION_TIMEOUT", durationUnit, time.Duration(15*time.Minute), "> 0", "§2.4", func(v *Values) *time.Duration { return &v.ContinuationTimeout }),
	param("BUS_QUEUE_EVENTS", integerUnit, int(256), ">= 1", "§1.2, §3", func(v *Values) *int { return &v.BusQueueEvents }),
	param("BUS_QUEUE_BYTES", bytesUnit, int64(4<<20), ">= frame limit 4MiB", "§1.2, §3", func(v *Values) *int64 { return &v.BusQueueBytes }),
	param("MEDIA_MAX_OBJECT", bytesUnit, int64(100<<20), "<= MEDIA_SESSION_BUDGET", "§1.3", func(v *Values) *int64 { return &v.MediaMaxObject }),
	param("MEDIA_MAX_UPLOADS", integerUnit, int(4), ">= 1", "§1.3", func(v *Values) *int { return &v.MediaMaxUploads }),
	param("MEDIA_SESSION_BUDGET", bytesUnit, int64(256<<20), "<= MEDIA_DISK_BUDGET", "§1.3", func(v *Values) *int64 { return &v.MediaSessionBudget }),
	param("MEDIA_DISK_BUDGET", bytesUnit, int64(2<<30), "> 0", "§1.3", func(v *Values) *int64 { return &v.MediaDiskBudget }),
	param("MEDIA_TTL", durationUnit, time.Duration(15*time.Minute), "> 0", "§1.3, §1.4", func(v *Values) *time.Duration { return &v.MediaTTL }),
	param("PHASE_HELLO", durationUnit, time.Duration(10*time.Second), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.PhaseHello }),
	param("PHASE_CONFIG", durationUnit, time.Duration(60*time.Second), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.PhaseConfig }),
	param("PHASE_MIGRATION", durationUnit, time.Duration(15*time.Minute), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.PhaseMigration }),
	param("PHASE_ONSTART", durationUnit, time.Duration(60*time.Second), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.PhaseOnStart }),
	param("PHASE_ONREADY", durationUnit, time.Duration(30*time.Second), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.PhaseOnReady }),
	param("STOP_TOTAL", durationUnit, time.Duration(10*time.Second), ">= STOP_DRAIN + STOP_ONSTOP", "§4.2", func(v *Values) *time.Duration { return &v.StopTotal }),
	param("STOP_DRAIN", durationUnit, time.Duration(6*time.Second), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.StopDrain }),
	param("STOP_ONSTOP", durationUnit, time.Duration(4*time.Second), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.StopOnStop }),
	param("RESTART_WINDOW", durationUnit, time.Duration(5*time.Minute), "> 0", "§4.2", func(v *Values) *time.Duration { return &v.RestartWindow }),
	param("RESTART_MAX_CRASHES", integerUnit, int(3), ">= 1", "§4.2", func(v *Values) *int { return &v.RestartMaxCrashes }),
	param("JOB_TIMEOUT", durationUnit, time.Duration(15*time.Minute), "> 0", "§4.3", func(v *Values) *time.Duration { return &v.JobTimeout }),
	param("BUTTON_TOKEN_TTL", durationUnit, time.Duration(24*time.Hour), "> 0", "MESSAGE-MODEL §6.3", func(v *Values) *time.Duration { return &v.ButtonTokenTTL }),
	param("TOOL_CONFIRM_TIMEOUT", durationUnit, time.Duration(5*time.Minute), "> 0", "§13.2", func(v *Values) *time.Duration { return &v.ToolConfirmTimeout }),
	param("LOG_BUFFER_RECORDS", integerUnit, int(1000), ">= 1", "§14.4", func(v *Values) *int { return &v.LogBufferRecords }),
	param("INSTALL_MAX_ARCHIVE", bytesUnit, int64(256<<20), "> 0", "§9.3", func(v *Values) *int64 { return &v.InstallMaxArchive }),
	param("INSTALL_MAX_EXPANDED", bytesUnit, int64(1<<30), ">= INSTALL_MAX_ARCHIVE", "§9.3", func(v *Values) *int64 { return &v.InstallMaxExpanded }),
	param("INSTALL_MAX_ENTRIES", integerUnit, int(10000), ">= 1", "§9.3", func(v *Values) *int { return &v.InstallMaxEntries }),
	param("DEV_RELOAD_POLL", durationUnit, time.Duration(500*time.Millisecond), "> 0", "§15.5", func(v *Values) *time.Duration { return &v.DevReloadPoll }),
	param("PLUGIN_MEMORY_SOFT_LIMIT", bytesUnit, int64(0), "> 0", "§16", func(v *Values) *int64 { return &v.PluginMemorySoftLimit }),
	param("CORE_MEMORY_SOFT_LIMIT", bytesUnit, int64(0), "> 0", "§16", func(v *Values) *int64 { return &v.CoreMemorySoftLimit }),
}

// Defaults 返回 CONTRACTS §16 规定的全部缺省值。
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
			return "unset"
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
