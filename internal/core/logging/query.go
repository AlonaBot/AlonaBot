package logging

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var historyHeader = regexp.MustCompile(`^\[([0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3})\] \[(Core|Plug:[^\]]+|Adpt:[^\]]+)\]( \[(DBUG|INFO|WARN|ERRO|CRITICAL)\](?: \[([^\]]+)\])?: )(.*)$`)
var historyMultilineField = regexp.MustCompile(` [^\s=]+=\n`)
var historyField = regexp.MustCompile(`(?:^| )([^\s=]+)=`)

// Query 按 Core 本地日期、来源、最低级别及 trace 前缀读取 JSON 或旧版文本日志。
// 旧版 Message 保留来源标签之后的原文；文件中省略的身份与精度不会被补造。
func Query(dir string, date time.Time, source string, level slog.Level, traceID string) ([]Record, error) {
	writer := &fileWriter{opt: FileOptions{Dir: dir}}
	files, err := writer.list()
	if errors.Is(err, os.ErrNotExist) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list log files: %w", err)
	}
	records := make([]Record, 0)
	day := date.In(time.Local).Format(dateLayout)
	for _, file := range files {
		if file.day != day {
			continue
		}
		name := filepath.Join(dir, file.name)
		f, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		rows, readErr := readHistory(f, date)
		if err := errors.Join(readErr, f.Close()); err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		for _, record := range rows {
			trace := traceID
			if len(record.TraceID) == traceDigits {
				trace = traceID[:min(len(traceID), traceDigits)]
			}
			if (source == "" || source == record.Source) && record.Level >= level && (trace == "" || strings.HasPrefix(record.TraceID, trace)) {
				records = append(records, record)
			}
		}
	}
	return records, nil
}

func readHistory(reader io.Reader, date time.Time) ([]Record, error) {
	input := bufio.NewReader(reader) // ReadString 保留任意长度的普通记录和多行字段。
	records := make([]Record, 0)
	var current string
	flush := func() error {
		if current == "" {
			return nil
		}
		record, err := parseHistory(current, date)
		if err != nil {
			return err
		}
		records = append(records, record)
		return nil
	}
	for lineNumber := 1; ; lineNumber++ {
		line, err := input.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		line = strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(line, continuationIndent) {
			if current == "" {
				return nil, fmt.Errorf("line %d: continuation without record", lineNumber)
			}
			current += "\n" + line
		} else if line != "" {
			if e := flush(); e != nil {
				return nil, fmt.Errorf("before line %d: %w", lineNumber, e)
			}
			current = line
		} else if err == nil {
			return nil, fmt.Errorf("line %d: empty log line", lineNumber)
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return records, nil
}

func parseHistory(raw string, date time.Time) (Record, error) {
	if strings.HasPrefix(raw, "{") {
		var record Record
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&record); err != nil {
			return Record{}, fmt.Errorf("decode structured log record: %w", err)
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			return Record{}, fmt.Errorf("unexpected content after structured log record")
		}
		return record, nil
	}
	first, _, _ := strings.Cut(raw, "\n")
	m := historyHeader.FindStringSubmatch(first)
	if m == nil {
		return Record{}, fmt.Errorf("invalid log record %q", first)
	}
	stamp, err := time.ParseInLocation(dateLayout+" 15:04:05.000", date.In(time.Local).Format(dateLayout)+" "+m[1], time.Local)
	if err != nil {
		return Record{}, fmt.Errorf("invalid log time: %w", err)
	}
	level := map[string]slog.Level{"DBUG": slog.LevelDebug, "INFO": slog.LevelInfo, "WARN": slog.LevelWarn, "ERRO": slog.LevelError, "CRITICAL": LevelCritical}[m[4]]
	source := strings.TrimPrefix(strings.TrimPrefix(m[2], "Plug:"), "Adpt:")
	attrs := map[string]any{}
	switch {
	case m[2] == "Core":
		source = ComponentCore
		attrs[KeyComponent] = ComponentCore
	case strings.HasPrefix(m[2], "Plug:"):
		attrs[KeyComponent] = ComponentPlugin
		attrs[KeyPluginID] = source
	default:
		attrs[KeyComponent] = ComponentAdapter
		id, instance, ok := strings.Cut(source, "/")
		attrs[KeyAdapterID] = id
		if ok {
			attrs[KeyInstanceID] = instance
		}
	}
	if m[5] != "" {
		position := m[5]
		if strings.HasPrefix(position, "\"") {
			var err error
			position, err = strconv.Unquote(position)
			if err != nil {
				return Record{}, fmt.Errorf("invalid source position: %w", err)
			}
		}
		separator := strings.LastIndexByte(position, ':')
		if separator < 0 {
			return Record{}, fmt.Errorf("invalid source position %q", position)
		}
		file, line := position[:separator], position[separator+1:]
		number, e := strconv.Atoi(line)
		if e != nil || number <= 0 {
			return Record{}, fmt.Errorf("invalid source position %q", position)
		}
		attrs[slog.SourceKey] = map[string]any{"file": file, "line": number}
	}
	suffix := strings.TrimPrefix(raw, first) // 多行字段继续属于同一条原始记录。
	body := m[6] + suffix
	fields := historyAttrs(body)
	for key, value := range fields {
		if !sourceKeys[key] || source == ComponentCore {
			attrs[key] = value
		}
	}
	trace, _ := fields["trace"].(string)
	if trace != "" {
		attrs[KeyTraceID] = trace
		delete(attrs, "trace")
	}
	start := strings.Index(raw, "] ["+m[2]+"]") + len("] ["+m[2]+"]")
	return Record{Time: stamp, Level: level, Source: source, Message: raw[start:], Attrs: attrs, TraceID: trace}, nil
}

func historyAttrs(body string) map[string]any {
	attrs := map[string]any{}
	// 跳过消息和字段中的引号内容，避免把其中的 key=value 误判为另一个属性。
	for offset := 0; offset < len(body); {
		if body[offset] == '"' {
			end, _, _ := quotedHistory(body[offset:])
			offset += end
			continue
		}
		m := historyField.FindStringSubmatchIndex(body[offset:])
		if m == nil {
			break
		}
		begin := offset + m[0]
		if quote := strings.IndexByte(body[offset:begin], '"'); quote >= 0 {
			offset += quote
			continue
		}
		key := body[offset+m[2] : offset+m[3]]
		offset += m[1]
		switch {
		case offset < len(body) && body[offset] == '"':
			end, value, ok := quotedHistory(body[offset:])
			if ok {
				attrs[key] = value
			}
			offset += end
		case strings.HasPrefix(body[offset:], "\n"+continuationIndent):
			value := body[offset+1:]
			end := len(value)
			if next := historyMultilineField.FindStringIndex(value); next != nil {
				end = next[0]
			}
			lines := strings.Split(value[:end], "\n")
			for i := range lines {
				lines[i] = strings.TrimPrefix(lines[i], continuationIndent)
			}
			attrs[key] = strings.Join(lines, "\n")
			offset += 1 + end
		default:
			end := strings.IndexAny(body[offset:], " \n")
			if end < 0 {
				end = len(body) - offset
			}
			attrs[key] = body[offset : offset+end]
			offset += end
		}
	}
	return attrs
}

// quotedHistory 尝试读取完整的带引号字段；结束边界有歧义时只跳过开头引号，继续寻找可靠字段。
func quotedHistory(text string) (int, string, bool) {
	escaped := false
	for i := 1; i < len(text); i++ {
		if escaped {
			escaped = false
			continue
		}
		if text[i] == '\\' {
			escaped = true
			continue
		}
		if text[i] == '"' {
			if i+1 < len(text) && text[i+1] != ' ' && text[i+1] != '\n' {
				return 1, "", false
			}
			value, err := strconv.Unquote(text[:i+1])
			return i + 1, value, err == nil
		}
	}
	return 1, "", false
}
