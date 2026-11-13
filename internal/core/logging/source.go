package logging

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"
)

// unstructuredMessage 是非结构化输出行的记录消息（CONTRACTS §14.1）。
const unstructuredMessage = "unstructured output"

// streamStderr 是插件进程中写 JSON 记录的流；stdout 上的文本一律是非结构化输出。
const streamStderr = "stderr"

// Source 把一个插件进程的输出行转成 Core 的记录；方法可以并发调用。
type Source struct {
	levels *Levels
	id     string // 按来源取生效级别的 ID
	// mu 保护 h。
	mu sync.Mutex
	h  slog.Handler // 已带来源字段，不按级别过滤
}

// SetSession 在 Session 建立后为之后的记录补上 session_id；每个 Source 只调用一次。
func (s *Source) SetSession(sessionID []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.h = s.h.WithAttrs([]slog.Attr{slog.String(KeySessionID, hex.EncodeToString(sessionID))})
}

// Line 记录插件进程输出的一行（不含换行符）。stderr 上合法的 JSON 记录按原级别、时间、位置与字段记录，
// 其余的行作为 WARN 原样保留；然后按该来源的生效级别过滤。pluginID 不使用，来源以建立 Source 时为准。
func (s *Source) Line(_, stream string, line []byte) {
	r, ok := slog.Record{}, false
	if stream == streamStderr {
		r, ok = parseRecord(line)
	}
	if !ok {
		r = slog.NewRecord(time.Now(), slog.LevelWarn, unstructuredMessage, 0)
		r.AddAttrs(slog.String("stream", stream), slog.String("output", string(line)))
	}
	if r.Level < s.levels.Source(s.id) {
		return
	}
	s.mu.Lock()
	h := s.h
	s.mu.Unlock()
	_ = h.Handle(context.Background(), r)
}

// parseRecord 把 SDK 写出的一行 JSON 记录转成 slog.Record：time、level、msg 必须存在且合法，
// source 转成 *slog.Source 字段放在最前；其余字段按出现顺序保留，Core 补充的来源字段被丢弃。
func parseRecord(line []byte) (slog.Record, bool) {
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.UseNumber()
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return slog.Record{}, false
	}
	var (
		t                   time.Time
		lv                  slog.Level
		msg                 string
		src                 *slog.Source
		attrs               []slog.Attr
		hasT, hasLv, hasMsg bool
	)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return slog.Record{}, false
		}
		key, _ := tok.(string) // 对象内的键一定是字符串
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return slog.Record{}, false
		}
		switch key {
		case slog.TimeKey:
			var s string
			if json.Unmarshal(raw, &s) != nil {
				return slog.Record{}, false
			}
			if t, err = time.Parse(time.RFC3339Nano, s); err != nil {
				return slog.Record{}, false
			}
			hasT = true
		case slog.LevelKey:
			var s string
			if json.Unmarshal(raw, &s) != nil || lv.UnmarshalText([]byte(s)) != nil {
				return slog.Record{}, false
			}
			hasLv = true
		case slog.MessageKey:
			if json.Unmarshal(raw, &msg) != nil {
				return slog.Record{}, false
			}
			hasMsg = true
		case slog.SourceKey:
			src = &slog.Source{}
			if json.Unmarshal(raw, src) != nil {
				return slog.Record{}, false
			}
		default:
			if !sourceKeys[key] || key == KeyInstanceID {
				attrs = append(attrs, jsonAttr(key, raw))
			}
		}
	}
	// 对象之后不能再有内容。
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return slog.Record{}, false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) || !hasT || !hasLv || !hasMsg {
		return slog.Record{}, false
	}
	r := slog.NewRecord(t, lv, msg, 0)
	if src != nil {
		r.AddAttrs(slog.Any(slog.SourceKey, src))
	}
	r.AddAttrs(attrs...)
	return r, true
}

// jsonAttr 把一个 JSON 值转成字段：字符串与布尔值用对应类型，数字保留原文，对象、数组与 null 保留原始 JSON。
func jsonAttr(key string, raw json.RawMessage) slog.Attr {
	switch raw[0] {
	case '"':
		var s string
		_ = json.Unmarshal(raw, &s) // Decode 已确认是合法 JSON
		return slog.String(key, s)
	case 't', 'f':
		return slog.Bool(key, raw[0] == 't')
	case '{', '[', 'n':
		return slog.Any(key, raw)
	}
	return slog.Any(key, json.Number(raw))
}
