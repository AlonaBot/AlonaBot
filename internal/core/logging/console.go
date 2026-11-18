package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// traceDigits 是控制台显示的 trace_id 位数（CONTRACTS §14.2）。
const traceDigits = 8

// continuationIndent 是多行值后续行的缩进（CONTRACTS §14.2）。
const continuationIndent = "    "

// 控制台着色使用的 ANSI 序列（CONTRACTS §14.2）。
const (
	ansiReset  = "\x1b[0m"
	ansiGreen  = "\x1b[32m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
	ansiYellow = "\x1b[33m"
	ansiRed    = "\x1b[31m"
)

// consoleWriter 串行化同一目标上的写入；同一输出派生出的 Handler 共用它。
type consoleWriter struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
}

// consoleHandler 以 CONTRACTS §14.2 的控制台格式输出记录：
//
//	[时:分:秒.毫秒] [来源] [级别] [目录/文件:行号]: 消息 key=value ...
//
// 它不按级别过滤。来源标签由 component 与来源 ID 决定，这些字段不再重复输出为 key=value。
type consoleHandler struct {
	out    *consoleWriter
	attrs  []slog.Attr // WithAttrs 累积的字段，键已带分组前缀
	prefix string      // WithGroup 累积的分组前缀，例如 "a.b."
}

func newConsoleHandler(w io.Writer, color bool) *consoleHandler {
	return &consoleHandler{out: &consoleWriter{w: w, color: color}}
}

func (h *consoleHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *consoleHandler) WithAttrs(as []slog.Attr) slog.Handler {
	c := *h
	c.attrs = appendFlat(append([]slog.Attr(nil), h.attrs...), h.prefix, as)
	return &c
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	c := *h
	c.prefix = h.prefix + name + "."
	return &c
}

// appendFlat 把 as 展开成不含分组的字段追加到 dst，分组名以点号拼进键；空字段与空分组被跳过。
func appendFlat(dst []slog.Attr, prefix string, as []slog.Attr) []slog.Attr {
	for _, a := range as {
		a.Value = a.Value.Resolve()
		switch {
		case a.Equal(slog.Attr{}):
		case a.Value.Kind() == slog.KindGroup:
			p := prefix
			if a.Key != "" {
				p += a.Key + "."
			}
			dst = appendFlat(dst, p, a.Value.Group())
		default:
			a.Key = prefix + a.Key
			dst = append(dst, a)
		}
	}
	return dst
}

func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	// 1. 汇总字段；插件记录的源码位置以 *slog.Source 字段传入（见 source.go）。
	attrs := append([]slog.Attr(nil), h.attrs...)
	src := r.Source()
	r.Attrs(func(a slog.Attr) bool {
		if s, ok := a.Value.Any().(*slog.Source); ok && a.Key == slog.SourceKey && h.prefix == "" {
			src = s
			return true
		}
		attrs = appendFlat(attrs, h.prefix, []slog.Attr{a})
		return true
	})
	fields := map[string]string{}
	for _, a := range attrs {
		if sourceKeys[a.Key] {
			fields[a.Key] = a.Value.String()
		}
	}

	// 2. 时间、来源标签、级别与源码位置。
	var b bytes.Buffer
	color := h.out.color
	b.WriteByte('[')
	paint(&b, color, ansiGreen, r.Time.Local().Format("15:04:05.000"))
	b.WriteString("] [")
	b.WriteString(sourceTag(fields))
	b.WriteString("] [")
	abbr, code := levelLabel(r.Level)
	paint(&b, color, code, abbr)
	b.WriteString("]")
	if src != nil && src.File != "" {
		file := path.Clean(strings.ReplaceAll(src.File, `\`, "/"))
		b.WriteString(" [")
		b.WriteString(printable(path.Base(path.Dir(file)) + "/" + path.Base(file) + ":" + strconv.Itoa(src.Line)))
		b.WriteString("]")
	}
	b.WriteString(": ")
	b.WriteString(printable(r.Message))

	// 3. key=value 按记录顺序输出在同一行；多行的值放到最后，后续行缩进输出。
	hidden := hiddenKeys(fields[KeyComponent])
	var multiline []slog.Attr
	for _, a := range attrs {
		if hidden[a.Key] {
			continue
		}
		v := valueText(a.Value)
		if a.Key == KeyTraceID {
			b.WriteString(" trace=")
			b.WriteString(printable(v[:min(len(v), traceDigits)]))
			continue
		}
		if strings.Contains(strings.TrimRight(v, "\n"), "\n") {
			multiline = append(multiline, a)
			continue
		}
		b.WriteByte(' ')
		b.WriteString(printable(a.Key))
		b.WriteByte('=')
		b.WriteString(quoteValue(v))
	}
	for _, a := range multiline {
		b.WriteByte(' ')
		b.WriteString(printable(a.Key))
		b.WriteByte('=')
		for _, line := range strings.Split(strings.TrimRight(valueText(a.Value), "\n"), "\n") {
			b.WriteByte('\n')
			b.WriteString(continuationIndent)
			b.WriteString(printable(line))
		}
	}
	b.WriteByte('\n')

	h.out.mu.Lock()
	defer h.out.mu.Unlock()
	_, err := h.out.w.Write(b.Bytes())
	return err
}

// sourceTag 返回来源标签：Core、Plug:<id> 或 Adpt:<adapter_id>[/<instance_id>]；插件 ID 去掉市场前缀。
func sourceTag(fields map[string]string) string {
	switch fields[KeyComponent] {
	case ComponentPlugin:
		return "Plug:" + printable(strings.TrimPrefix(fields[KeyPluginID], validate.MarketPrefix))
	case ComponentAdapter:
		tag := "Adpt:" + printable(fields[KeyAdapterID])
		if id := fields[KeyInstanceID]; id != "" {
			tag += "/" + printable(id)
		}
		return tag
	}
	return "Core"
}

// hiddenKeys 返回控制台不输出的字段：来源类别总是体现在标签中；插件与 Adapter 的记录
// 由 Core 补充的来源字段也不重复输出。Core 关于某个插件的记录照常输出 plugin_id 等字段。
func hiddenKeys(component string) map[string]bool {
	switch component {
	case ComponentPlugin, ComponentAdapter:
		return sourceKeys
	}
	return map[string]bool{KeyComponent: true}
}

// levelLabel 返回级别缩写与颜色；介于标准级别之间的值归入较低的标准级别。
func levelLabel(lv slog.Level) (string, string) {
	switch {
	case lv >= LevelCritical:
		return "CRITICAL", ansiRed
	case lv >= slog.LevelError:
		return "ERRO", ansiRed
	case lv >= slog.LevelWarn:
		return "WARN", ansiYellow
	case lv >= slog.LevelInfo:
		return "INFO", ansiCyan
	}
	return "DBUG", ansiBlue
}

// paint 写入 s，color 为 true 时用 code 着色。
func paint(b *bytes.Buffer, color bool, code, s string) {
	if !color {
		b.WriteString(s)
		return
	}
	b.WriteString(code)
	b.WriteString(s)
	b.WriteString(ansiReset)
}

// valueText 返回字段值的文本；插件记录中的对象与数组保持原始 JSON 文本。
func valueText(v slog.Value) string {
	if raw, ok := v.Any().(json.RawMessage); ok && v.Kind() == slog.KindAny {
		return string(raw)
	}
	return v.String()
}

// quoteValue 在值为空、含空白、引号或不可打印字符时按 Go 的带引号字符串输出。
func quoteValue(s string) string {
	if s == "" || strings.ContainsFunc(s, func(r rune) bool { return unicode.IsSpace(r) || r == '"' || !unicode.IsPrint(r) }) {
		return strconv.Quote(s)
	}
	return s
}

// printable 原样返回只含可打印字符的 s，否则返回带引号的形式。插件提供的消息、键与位置经过它，
// 换行或终端控制序列不能伪造出另一条记录或改写终端。
func printable(s string) string {
	if strings.ContainsFunc(s, func(r rune) bool { return !unicode.IsPrint(r) }) {
		return strconv.Quote(s)
	}
	return s
}
