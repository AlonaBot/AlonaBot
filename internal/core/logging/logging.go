// Package logging 汇集 Core 进程的全部日志：Core 自己的记录与插件进程输出的记录（CONTRACTS §14）。
//
// 每条记录写到 Core 的 stderr（控制台格式或 JSON），启用日志文件后同时以不着色的控制台格式写入
// 按日期轮转的文件。记录的来源由 component 字段区分：core、plugin 或 adapter；插件与 Adapter 的记录
// 另带 plugin_id 或 adapter_id，以及 Core 补充的 session_id 与 generation_id。插件以 JSON 行写 stderr，
// [Source] 逐行解析后转成同样的记录，插件给出的同名字段被丢弃，不能覆盖 Core 补充的来源字段。
//
// Core 自己的记录按全局级别过滤；插件的记录按该来源的生效级别过滤（[Levels]）。
// 按来源的级别目前只保存在内存中。
//
// 包的组成：
//   - [Open] 按 [Options] 建立 [Output]；[Output.Logger] 是 Core 自己的 Logger，[Output.EnableFile] 开始写日志文件。
//   - [Output.Plugin] 为一个插件进程建立 [Source]，[Source.Line] 的签名与 process.LogFunc 相同。
//   - [Levels] 保存全局与按来源的级别；[ProtoLevel] 把级别转成 Bootstrap 的 log_level。
//   - 控制台格式在 console.go，插件输出的解析在 source.go，日志文件的轮转与清理在 file.go。
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// 记录中表示来源的字段名（CONTRACTS §14.1）。
const (
	KeyComponent    = "component" // 来源类别：ComponentCore、ComponentPlugin 或 ComponentAdapter
	KeyPluginID     = "plugin_id"
	KeyAdapterID    = "adapter_id"
	KeyInstanceID   = "instance_id"
	KeySessionID    = "session_id" // 十六进制
	KeyGenerationID = "generation_id"
	KeyTraceID      = "trace_id" // 控制台只显示前 traceDigits 位
)

// 来源类别。
const (
	ComponentCore    = "core"
	ComponentPlugin  = "plugin"
	ComponentAdapter = "adapter"
)

// sourceKeys 是由 Core 补充、插件不能覆盖的字段。
var sourceKeys = map[string]bool{
	KeyComponent: true, KeyPluginID: true, KeyAdapterID: true, KeyInstanceID: true,
	KeySessionID: true, KeyGenerationID: true,
}

// LevelOff 高于任何记录的级别，用作来源级别 off：该来源的记录全部不输出。
const LevelOff = slog.Level(math.MaxInt)

// Format 选择 stderr 上的输出格式（CONTRACTS §14.2）。
type Format string

// stderr 的输出格式。
const (
	FormatText Format = "text" // 控制台格式
	FormatJSON Format = "json" // 每行一个 JSON 对象，含全部字段
)

// Options 描述 Core 日志的启动设置。
type Options struct {
	Format Format
	Level  slog.Level // 全局级别
	// Stderr 是 Core 的 stderr。它是终端且未设置 NO_COLOR 时，控制台格式才着色。
	Stderr io.Writer
}

// Output 把 Core 与插件的记录写到 stderr 与日志文件；方法可以并发调用。
type Output struct {
	levels  *Levels
	stderr  slog.Handler // 只写 stderr，不按级别过滤
	file    *fileWriter  // EnableFile 之前丢弃写入
	handler slog.Handler // stderr 与日志文件，不按级别过滤
}

// Open 按 opt 建立日志输出；opt.Format 不是 text 或 json 时返回错误。
func Open(opt Options) (*Output, error) {
	var stderr slog.Handler
	switch opt.Format {
	case FormatText:
		stderr = newConsoleHandler(opt.Stderr, colorEnabled(opt.Stderr))
	case FormatJSON:
		stderr = slog.NewJSONHandler(opt.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.Level(math.MinInt)})
	default:
		return nil, fmt.Errorf("unknown log format %q (want %s or %s)", opt.Format, FormatText, FormatJSON)
	}
	o := &Output{levels: &Levels{sources: map[string]slog.Level{}}, stderr: stderr}
	o.levels.global.Set(opt.Level)
	o.file = &fileWriter{report: o.reportFileError}
	o.handler = slog.NewMultiHandler(stderr, newConsoleHandler(o.file, false))
	return o, nil
}

// colorEnabled 报告 w 是否是终端且未设置 NO_COLOR（https://no-color.org）。
func colorEnabled(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// Logger 返回 Core 自己的 Logger：记录带 component=core，按全局级别过滤。
func (o *Output) Logger() *slog.Logger {
	return slog.New(&leveled{h: o.handler.WithAttrs([]slog.Attr{slog.String(KeyComponent, ComponentCore)}), level: &o.levels.global})
}

// Levels 返回全局与按来源的级别。
func (o *Output) Levels() *Levels { return o.levels }

// EnableFile 开始按 opt 写日志文件，并按上限清理旧文件。只能调用一次，调用方须已持有 core.lock，
// 避免两个 Core 写同一个文件。上限不合法时返回错误；目录或文件无法打开时只报告一次并继续运行。
func (o *Output) EnableFile(opt FileOptions) error {
	return o.file.enable(opt)
}

// Close 关闭当前日志文件；之后的记录只写 stderr。
func (o *Output) Close() error { return o.file.close() }

// reportFileError 把日志文件的错误只写到 stderr，避免再次写入出错的文件。
func (o *Output) reportFileError(err error) {
	slog.New(o.stderr).Error("log file unavailable", KeyComponent, ComponentCore, "err", err)
}

// Plugin 为插件进程的一次启动建立 [Source]，记录带 component=plugin、plugin_id 与 generation_id。
func (o *Output) Plugin(pluginID, generationID string) *Source {
	h := o.handler.WithAttrs([]slog.Attr{
		slog.String(KeyComponent, ComponentPlugin), slog.String(KeyPluginID, pluginID), slog.String(KeyGenerationID, generationID),
	})
	return &Source{levels: o.levels, id: pluginID, h: h}
}

// Levels 保存全局级别与按来源单独设置的级别；方法可以并发调用。
type Levels struct {
	global slog.LevelVar
	// mu 保护 sources。
	mu      sync.RWMutex
	sources map[string]slog.Level // 只含单独设置过的来源，其余跟随全局
}

// SetGlobal 修改全局级别，立即生效。
func (l *Levels) SetGlobal(lv slog.Level) { l.global.Set(lv) }

// SetSource 为插件或 Adapter 包 id 单独设置级别（可以是 [LevelOff]），立即生效。
func (l *Levels) SetSource(id string, lv slog.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sources[id] = lv
}

// InheritSource 取消 id 的单独设置，恢复跟随全局。
func (l *Levels) InheritSource(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.sources, id)
}

// Source 返回 id 的生效级别。
func (l *Levels) Source(id string) slog.Level {
	l.mu.RLock()
	lv, ok := l.sources[id]
	l.mu.RUnlock()
	if ok {
		return lv
	}
	return l.global.Level()
}

// ProtoLevel 把生效级别转成 Bootstrap 的 log_level；介于标准级别之间的值取较低的标准级别。
func ProtoLevel(lv slog.Level) pb.LogLevel {
	switch {
	case lv >= LevelOff:
		return pb.LogLevel_LOG_LEVEL_OFF
	case lv >= slog.LevelError:
		return pb.LogLevel_LOG_LEVEL_ERROR
	case lv >= slog.LevelWarn:
		return pb.LogLevel_LOG_LEVEL_WARN
	case lv >= slog.LevelInfo:
		return pb.LogLevel_LOG_LEVEL_INFO
	}
	return pb.LogLevel_LOG_LEVEL_DEBUG
}

// leveled 在 h 之前按 level 过滤记录。
type leveled struct {
	h     slog.Handler
	level slog.Leveler
}

func (l *leveled) Enabled(ctx context.Context, lv slog.Level) bool {
	return lv >= l.level.Level() && l.h.Enabled(ctx, lv)
}

func (l *leveled) Handle(ctx context.Context, r slog.Record) error { return l.h.Handle(ctx, r) }

func (l *leveled) WithAttrs(as []slog.Attr) slog.Handler {
	return &leveled{h: l.h.WithAttrs(as), level: l.level}
}

func (l *leveled) WithGroup(name string) slog.Handler {
	return &leveled{h: l.h.WithGroup(name), level: l.level}
}
