package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"sync"
	"time"
)

// Record 保留一条已输出记录的完整实时身份、属性和追踪值。
type Record struct {
	Time    time.Time      `json:"time"`
	Level   slog.Level     `json:"level"`
	Source  string         `json:"source"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs"`
	TraceID string         `json:"trace_id"`
}

// MarshalJSON 将最近记录与流式记录中的自定义级别写作 CRITICAL。
func (r Record) MarshalJSON() ([]byte, error) {
	type record Record
	level := r.Level.String()
	if r.Level == LevelCritical {
		level = "CRITICAL"
	}
	return json.Marshal(struct {
		record
		Level string `json:"level"`
	}{record(r), level})
}

// UnmarshalJSON 读取管理 API 的级别文本，同时保留 slog 支持的其他级别。
func (r *Record) UnmarshalJSON(data []byte) error {
	type record Record
	var value struct {
		record
		Level string `json:"level"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if value.Level == "CRITICAL" {
		value.record.Level = LevelCritical
	} else if err := value.record.Level.UnmarshalText([]byte(value.Level)); err != nil {
		return err
	}
	*r = Record(value.record)
	return nil
}

type recordRing struct {
	records []Record
	next    int
}

type subscription struct {
	queue    chan Record
	overflow chan struct{} // 输出侧关闭，解除仍在发送初始快照的 goroutine。
}

type recordBuffer struct {
	mu          sync.Mutex // 同一把锁连接订阅建立、初始快照与后续追加。
	limit       int
	sources     map[string]*recordRing
	subscribers map[*subscription]struct{}
}

func (b *recordBuffer) append(record Record) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ring := b.sources[record.Source]
	if ring == nil {
		ring = &recordRing{}
		b.sources[record.Source] = ring
	}
	if len(ring.records) < b.limit {
		ring.records = append(ring.records, record)
	} else {
		ring.records[ring.next] = record
		ring.next = (ring.next + 1) % b.limit
	}
	for sub := range b.subscribers {
		select { // 慢订阅只关闭自身，日志输出从不等待网络消费者。
		case sub.queue <- record:
		default:
			delete(b.subscribers, sub)
			close(sub.overflow)
		}
	}
}

func cloneRecord(record Record) Record {
	record.Attrs = cloneAttrs(record.Attrs)
	return record
}

func cloneAttrs(attrs map[string]any) map[string]any {
	copy := make(map[string]any, len(attrs))
	for key, value := range attrs {
		copy[key] = cloneValue(value)
	}
	return copy
}
func cloneValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneAttrs(v)
	case []any:
		copy := make([]any, len(v))
		for i, value := range v {
			copy[i] = cloneValue(value)
		}
		return copy
	default:
		return value
	}
}

func (b *recordBuffer) recent(source string) []Record {
	result := make([]Record, 0)
	for id, ring := range b.sources {
		if source != "" && source != id {
			continue
		}
		for i := 0; i < len(ring.records); i++ {
			result = append(result, cloneRecord(ring.records[(ring.next+i)%len(ring.records)]))
		}
	}
	slices.SortStableFunc(result, func(a, b Record) int { return a.Time.Compare(b.Time) })
	return result
}

// Recent 返回来源的最近记录；空来源合并全部来源，按时间升序返回。
func (o *Output) Recent(source string) []Record {
	o.buffer.mu.Lock()
	defer o.buffer.mu.Unlock()
	return o.buffer.recent(source)
}

// Subscribe 原子取得初始快照并订阅后续记录；ctx 取消或发送队列溢出时关闭返回的通道。
// ctx 尚未取消时关闭表示 overflow，消费者应记录最后交付时间并重新读取 recent。
func (o *Output) Subscribe(ctx context.Context) <-chan Record {
	b := o.buffer
	sub := &subscription{queue: make(chan Record, b.limit), overflow: make(chan struct{})}
	b.mu.Lock()
	initial := b.recent("")
	b.subscribers[sub] = struct{}{}
	b.mu.Unlock()
	out := make(chan Record)
	go func() {
		defer close(out) // 唯一发送者关闭通道，消费者无需与日志写入共用锁。
		defer func() { b.mu.Lock(); delete(b.subscribers, sub); b.mu.Unlock() }()
		send := func(record Record) bool {
			select {
			case <-sub.overflow:
				return false
			default:
			}
			select {
			case <-ctx.Done():
				return false
			case <-sub.overflow:
				return false
			case out <- cloneRecord(record):
				return true
			}
		}
		for _, record := range initial {
			if !send(record) {
				return
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-sub.overflow:
				return
			case record := <-sub.queue:
				if !send(record) {
					return
				}
			}
		}
	}()
	return out
}

// bufferedHandler 在共用输出后保存记录；来源身份由 Core 预绑定属性覆盖。
type bufferedHandler struct {
	output     *Output
	next       slog.Handler
	stderrOnly bool // 文件故障报告留在 stderr 和缓冲，避免重入文件锁。
	attrs      []slog.Attr
	groups     []string
}

func (h *bufferedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}
func (h *bufferedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := *h
	bound := map[string]bool{}
	for _, a := range h.attrs {
		if sourceKeys[a.Key] {
			bound[a.Key] = true
		}
	}
	accepted := make([]slog.Attr, 0, len(attrs))
	for _, a := range inlineAttrs(attrs) {
		if len(h.groups) == 0 && bound[a.Key] {
			continue
		}
		accepted = append(accepted, a)
	}
	c.attrs = appendGrouped(h.attrs, h.groups, accepted)
	return &c
}
func (h *bufferedHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	c := *h
	c.groups = append(slices.Clone(h.groups), name)
	return &c
}

// appendGrouped 合并 WithGroup 累积的同一对象，避免 JSON 重复组名覆盖先前属性。
func appendGrouped(existing []slog.Attr, groups []string, attrs []slog.Attr) []slog.Attr {
	result := slices.Clone(existing)
	if len(groups) == 0 {
		return append(result, attrs...)
	}
	for i := len(result) - 1; i >= 0; i-- {
		if result[i].Key == groups[0] && result[i].Value.Kind() == slog.KindGroup {
			result[i].Value = slog.GroupValue(appendGrouped(result[i].Value.Group(), groups[1:], attrs)...)
			return result
		}
	}
	return append(result, slog.Attr{Key: groups[0], Value: slog.GroupValue(appendGrouped(nil, groups[1:], attrs)...)})
}

// inlineAttrs 按 slog 语义展开无名组，再检查身份字段，避免无名组绕过来源保护。
func inlineAttrs(attrs []slog.Attr) []slog.Attr {
	result := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		attr = resolveAttr(attr)
		if attr.Key == "" && attr.Value.Kind() == slog.KindGroup {
			result = append(result, inlineAttrs(attr.Value.Group())...)
		} else {
			result = append(result, attr)
		}
	}
	return result
}

func resolveAttr(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	if attr.Value.Kind() == slog.KindGroup {
		group := slices.Clone(attr.Value.Group())
		for i := range group {
			group[i] = resolveAttr(group[i])
		}
		attr.Value = slog.GroupValue(group...)
	}
	return attr
}
func (h *bufferedHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	trusted := map[string]slog.Attr{}
	for _, a := range h.attrs {
		if sourceKeys[a.Key] {
			trusted[a.Key] = a
		} else {
			attrs = append(attrs, a)
		}
	}
	// 同进程 SDK 的发送阈值更新前，Core 仍按当前来源设置复核这条记录。
	id := trusted[KeyPluginID].Value.String()
	if trusted[KeyComponent].Value.String() == ComponentCore {
		id = ""
	}
	if r.Level < h.output.levels.Source(id) {
		return nil
	}
	for _, a := range trusted {
		attrs = append(attrs, a)
	}
	live := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		for _, a := range inlineAttrs([]slog.Attr{a}) {
			if len(h.groups) == 0 {
				if a.Key == KeyInstanceID && trusted[KeyComponent].Value.String() != ComponentAdapter && trusted[KeyComponent].Value.String() != ComponentCore {
					continue
				}
				if _, bound := trusted[a.Key]; bound {
					continue
				}
			}
			live = append(live, a)
		}
		return true
	})
	attrs = appendGrouped(attrs, h.groups, live)

	normalized := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	normalized.AddAttrs(attrs...)
	if err := h.next.Handle(ctx, normalized); err != nil {
		return err
	}
	// 标准 JSON Handler 负责 LogValuer、error、分组及源码位置的同一转换规则。
	var encoded bytes.Buffer
	handler := slog.NewJSONHandler(&encoded, &slog.HandlerOptions{AddSource: true})
	if err := handler.Handle(ctx, normalized); err != nil {
		return err
	}
	var values map[string]any
	decoder := json.NewDecoder(&encoded)
	decoder.UseNumber()
	if err := decoder.Decode(&values); err != nil {
		return err
	}
	for _, key := range []string{slog.TimeKey, slog.LevelKey, slog.MessageKey} {
		delete(values, key)
	}
	source := ComponentCore
	if trusted[KeyComponent].Value.String() != ComponentCore {
		if identity, ok := trusted[KeyPluginID]; ok {
			source = identity.Value.String()
		}
	}
	trace, _ := values[KeyTraceID].(string)
	record := Record{Time: r.Time, Level: r.Level, Source: source, Message: r.Message, Attrs: values, TraceID: trace}
	h.output.buffer.append(record)
	if h.stderrOnly {
		return nil
	}
	encodedRecord, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = h.output.file.Write(append(encodedRecord, '\n'))
	return err
}
