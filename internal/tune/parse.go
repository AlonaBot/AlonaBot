package tune

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Parse 依次应用每次 --args 的内容，后者覆盖前者，再检查全部约束。
// 返回的覆盖每个参数只保留最后一次，并按参数表顺序排列。
func Parse(args []string) (Values, []Override, error) {
	v := Defaults()
	overrides := make(map[string]Override)
	sources := make(map[string]string)
	// 1. 解析输入并保留最后一次覆盖的值与出处。
	apply := func(text, origin, source string) error {
		name, raw, ok := strings.Cut(text, "=")
		name, raw = strings.TrimSpace(name), strings.TrimSpace(raw)
		if !ok {
			return fmt.Errorf("%s: parameter %q: missing '='", source, name)
		}
		if name == "" {
			return fmt.Errorf("%s: parameter %q: empty name", source, name)
		}
		for _, p := range parameters {
			if p.name != name {
				continue
			}
			n, err := p.parse(raw)
			if err != nil {
				return fmt.Errorf("%s: %s=%q: invalid format: %w", source, name, raw, err)
			}
			p.set(&v, n)
			overrides[name] = Override{Name: name, Value: raw, Origin: origin}
			sources[name] = source
			return nil
		}
		return fmt.Errorf("%s: %s: unknown parameter", source, name)
	}
	for i, arg := range args {
		source := fmt.Sprintf("--args item %d", i+1)
		if strings.Contains(arg, "=") {
			for j, text := range strings.Split(arg, ",") {
				if err := apply(text, "--args", fmt.Sprintf("%s entry %d", source, j+1)); err != nil {
					return Values{}, nil, err
				}
			}
			continue
		}
		content, err := os.ReadFile(arg)
		if err != nil {
			return Values{}, nil, fmt.Errorf("%s: parameter file %q: %w", source, arg, err)
		}
		for line, text := range strings.Split(string(content), "\n") {
			text = strings.TrimSpace(text)
			if text == "" || strings.HasPrefix(text, "#") {
				continue
			}
			origin := fmt.Sprintf("%s:%d", arg, line+1)
			if err := apply(text, origin, origin); err != nil {
				return Values{}, nil, err
			}
		}
	}
	// 2. 在所有覆盖应用后检查单值与相互约束，诊断使用最终值及其来源。
	describe := func(names ...string) string {
		var descriptions []string
		for _, name := range names {
			for _, p := range parameters {
				if p.name != name {
					continue
				}
				source := sources[name]
				if source == "" {
					source = "default"
				}
				value := p.format(p.get(&v))
				if override, ok := overrides[name]; ok {
					value = override.Value
				}
				descriptions = append(descriptions, fmt.Sprintf("%s=%s (%s)", name, value, source))
			}
		}
		return strings.Join(descriptions, ", ")
	}
	for _, p := range parameters {
		n := p.get(&v)
		_, overridden := overrides[p.name]
		if p.positive && n <= 0 && (p.defaultValue != 0 || overridden) {
			return Values{}, nil, fmt.Errorf("%s: must be > 0", describe(p.name))
		}
	}
	for _, name := range []string{"SESSION_MAX_PAYLOAD", "DISPATCH_QUEUE_BYTES", "BUS_QUEUE_BYTES"} {
		for _, p := range parameters {
			if p.name == name && p.get(&v) < maxFrame {
				return Values{}, nil, fmt.Errorf("%s: must be >= frame limit 4MiB", describe(name))
			}
		}
	}
	for _, check := range []struct {
		invalid bool
		names   []string
		rule    string
	}{
		{v.CoreMaxPayload < v.SessionMaxPayload, []string{"CORE_MAX_PAYLOAD", "SESSION_MAX_PAYLOAD"}, "CORE_MAX_PAYLOAD must be >= SESSION_MAX_PAYLOAD"},
		{v.ContinuationMax >= v.SessionMaxInflight, []string{"CONTINUATION_MAX", "SESSION_MAX_INFLIGHT"}, "CONTINUATION_MAX must be < SESSION_MAX_INFLIGHT"},
		{v.MediaMaxObject > v.MediaSessionBudget, []string{"MEDIA_MAX_OBJECT", "MEDIA_SESSION_BUDGET"}, "MEDIA_MAX_OBJECT must be <= MEDIA_SESSION_BUDGET"},
		{v.MediaSessionBudget > v.MediaDiskBudget, []string{"MEDIA_SESSION_BUDGET", "MEDIA_DISK_BUDGET"}, "MEDIA_SESSION_BUDGET must be <= MEDIA_DISK_BUDGET"},
		{v.InstallMaxExpanded < v.InstallMaxArchive, []string{"INSTALL_MAX_EXPANDED", "INSTALL_MAX_ARCHIVE"}, "INSTALL_MAX_EXPANDED must be >= INSTALL_MAX_ARCHIVE"},
		// 用减法比较，避免两个合法时长相加时溢出。
		{v.StopDrain > v.StopTotal || v.StopOnStop > v.StopTotal-v.StopDrain, []string{"STOP_TOTAL", "STOP_DRAIN", "STOP_ONSTOP"}, "STOP_TOTAL must be >= STOP_DRAIN + STOP_ONSTOP"},
	} {
		if check.invalid {
			return Values{}, nil, fmt.Errorf("%s: %s", describe(check.names...), check.rule)
		}
	}
	// 3. 按参数表顺序返回最终覆盖，避免 map 顺序进入日志。
	var result []Override
	for _, p := range parameters {
		if override, ok := overrides[p.name]; ok {
			result = append(result, override)
		}
	}
	return v, result, nil
}

func (p parameter) parse(raw string) (int64, error) {
	var n int64
	var err error
	switch p.unit {
	case integerUnit:
		if strings.HasPrefix(raw, "+") {
			return 0, fmt.Errorf("expected a decimal integer without a leading plus sign")
		}
		n, err = strconv.ParseInt(raw, 10, strconv.IntSize)
	case durationUnit:
		var value time.Duration
		value, err = time.ParseDuration(raw)
		n = int64(value)
	case bytesUnit:
		for _, suffix := range []string{"KB", "MB", "GB"} {
			if strings.HasSuffix(raw, suffix) {
				return 0, fmt.Errorf("use B, KiB, MiB or GiB instead of %s", suffix)
			}
		}
		multiplier := int64(0)
		digits := ""
		for _, suffix := range []struct {
			name string
			size int64
		}{{"GiB", 1 << 30}, {"MiB", 1 << 20}, {"KiB", 1 << 10}, {"B", 1}} {
			if strings.HasSuffix(raw, suffix.name) {
				multiplier = suffix.size
				digits = strings.TrimSuffix(raw, suffix.name)
				break
			}
		}
		if multiplier == 0 {
			return 0, fmt.Errorf("expected a decimal integer with B, KiB, MiB or GiB suffix")
		}
		n, err = strconv.ParseInt(digits, 10, 64)
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("size must be non-negative")
			}
			if n > math.MaxInt64/multiplier {
				return 0, fmt.Errorf("size exceeds int64 range")
			}
			n *= multiplier
		}
	}
	if err != nil {
		return 0, err
	}
	switch p.name {
	case "SESSION_MAX_INFLIGHT", "SESSION_MAX_PAYLOAD":
		if n > math.MaxUint32 {
			return 0, fmt.Errorf("value exceeds uint32 wire range")
		}
	case "REQUEST_TIMEOUT", "PHASE_HELLO", "PHASE_CONFIG", "PHASE_MIGRATION", "PHASE_ONSTART", "PHASE_ONREADY", "STOP_TOTAL":
		// Bootstrap 以 uint32 毫秒表示这些期限；禁止下发时截断或溢出。
		if n > 0 && (n%int64(time.Millisecond) != 0 || n/int64(time.Millisecond) > math.MaxUint32) {
			return 0, fmt.Errorf("duration must fit uint32 whole milliseconds")
		}
	}
	return n, nil
}
