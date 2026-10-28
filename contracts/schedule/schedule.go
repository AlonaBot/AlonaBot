// Package schedule 校验 interval 与五字段 cron 触发规则，并计算下一次触发时刻。
//
// 包只做纯计算：不读当前时间、不持有状态，同样的输入总是得到同样的结果。
// Core Scheduler 与官方 Agent 组件的定时任务共用它。触发语义见 spec/CONTRACTS-v0.1.md §4.3。
//
// interval 的触发时刻是 anchor + k×d（k ≥ 1），错过的直接跳过。
// cron 按真实时刻逐分钟推进，用任务时区的本地时间判断是否匹配：
// 夏令时跳过的本地时刻不触发，回拨重复的本地时刻对应两个 UTC 时刻，各触发一次。
package schedule

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // 没有系统时区库的平台也能解析 IANA 名称

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

const (
	// minInterval 是 interval 触发的最短周期（CONTRACTS §4.3）。
	minInterval = time.Second

	// defaultTimezone 是省略 timezone 时使用的时区（CONTRACTS §4.3）。
	defaultTimezone = "UTC"

	// searchYears 是判断 cron 能否触发时向后搜索的年数：
	// 八年覆盖闰年的 2 月 29 日与星期的全部组合，搜不到即视为永不触发。
	searchYears = 8
)

// cronFields 按 cron 表达式中的顺序列出五个字段及其取值范围。
// 星期的 0 与 7 都表示周日。
var cronFields = [...]struct {
	name     string
	min, max int
}{
	{"minute", 0, 59},
	{"hour", 0, 23},
	{"day", 1, 31},
	{"month", 1, 12},
	{"weekday", 0, 7},
}

const (
	fieldMinute = iota
	fieldHour
	fieldDay
	fieldMonth
	fieldWeekday
)

// Trigger 是校验过的触发规则，由 [Interval]、[Cron] 或 [FromSpec] 构造；零值不可用。
type Trigger struct {
	interval time.Duration  // interval 触发的周期；cron 触发时为 0
	cron     cronSpec       // interval 触发时为零值
	location *time.Location // cron 匹配所用的时区
}

// cronSpec 保存解析后的五个字段。
type cronSpec struct {
	fields [len(cronFields)][]bool // fields[i][v] 表示第 i 个字段允许取值 v
	// 日与星期都受限时按 OR 匹配，否则按 AND；以字段是否为 "*" 判断是否受限。
	dayRestricted, weekdayRestricted bool
}

// Interval 返回周期为 d 的 interval 触发规则；d 小于 1s 时返回错误。
func Interval(d time.Duration) (Trigger, error) {
	if d < minInterval {
		return Trigger{}, fmt.Errorf("interval must be at least %s", minInterval)
	}
	return Trigger{interval: d}, nil
}

// Cron 解析五字段 cron 表达式。timezone 必须是 IANA 名称，为空时使用 UTC，不接受 "Local"。
//
// 支持 *、数字、a-b、逗号列表、*/n 与 a-b/n；名称、@ 宏、L/W/#、秒字段一律拒绝。
// 永远不会触发的表达式（例如 2 月 30 日）也返回错误。
func Cron(expr, timezone string) (Trigger, error) {
	// 1. 解析时区：Local 依赖宿主机设置，结果不确定，因此拒绝。
	if timezone == "" {
		timezone = defaultTimezone
	}
	if timezone == "Local" {
		return Trigger{}, errors.New("timezone must be an IANA name, not Local")
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return Trigger{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}

	// 2. 逐个字段解析。
	parts := strings.Fields(expr)
	if len(parts) != len(cronFields) {
		return Trigger{}, fmt.Errorf("cron must contain exactly %d fields", len(cronFields))
	}
	var spec cronSpec
	for i, f := range cronFields {
		allowed, err := parseField(parts[i], f.min, f.max)
		if err != nil {
			return Trigger{}, fmt.Errorf("invalid cron %s field: %w", f.name, err)
		}
		spec.fields[i] = allowed
	}
	// 星期 7 与 0 都是周日，统一记到 0 上，匹配时只需查 time.Weekday 的取值。
	spec.fields[fieldWeekday][0] = spec.fields[fieldWeekday][0] || spec.fields[fieldWeekday][7]
	spec.dayRestricted = parts[fieldDay] != "*"
	spec.weekdayRestricted = parts[fieldWeekday] != "*"

	// 3. 拒绝永不触发的组合：各字段单独合法，但组合起来可能没有对应的日期。
	t := Trigger{cron: spec, location: location}
	start := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	if t.nextCron(start.Add(-time.Minute)).IsZero() {
		return Trigger{}, fmt.Errorf("cron has no occurrence within %d years", searchYears)
	}
	return t, nil
}

// FromSpec 从 REGISTER 中的 ScheduleSpec 构造触发规则；interval_ms 与 cron 必须恰有一个。
// 不检查 job_deadline_ms。
func FromSpec(s *pb.ScheduleSpec) (Trigger, error) {
	switch trigger := s.GetTrigger().(type) {
	case *pb.ScheduleSpec_IntervalMs:
		// interval_ms 是 uint64，超出 time.Duration 范围时换算会溢出。
		if trigger.IntervalMs > math.MaxInt64/uint64(time.Millisecond) {
			return Trigger{}, errors.New("interval is too large")
		}
		return Interval(time.Duration(trigger.IntervalMs) * time.Millisecond)
	case *pb.ScheduleSpec_Cron:
		return Cron(trigger.Cron, s.GetTimezone())
	default:
		return Trigger{}, errors.New("exactly one schedule trigger is required")
	}
}

// Next 返回严格晚于 now 的下一次触发时刻。anchor 是激活时刻，只用于 interval 触发。
// cron 在 searchYears 年内没有触发时刻时返回零值。
func (t Trigger) Next(anchor, now time.Time) time.Time {
	if t.interval == 0 {
		return t.nextCron(now)
	}
	if now.Before(anchor) {
		return anchor.Add(t.interval)
	}
	// 跳过已错过的 tick，不补跑：取 anchor + k×interval 中严格晚于 now 的最小值。
	k := now.Sub(anchor)/t.interval + 1
	return anchor.Add(k * t.interval)
}

// nextCron 从 now 之后的第一个整分钟起逐分钟查找匹配的时刻。
// 以 UTC 分钟推进、用本地时间匹配，夏令时的跳过与重复因此自然成立。
func (t Trigger) nextCron(now time.Time) time.Time {
	candidate := now.UTC().Truncate(time.Minute).Add(time.Minute)
	limit := candidate.AddDate(searchYears, 0, 0)
	for ; candidate.Before(limit); candidate = candidate.Add(time.Minute) {
		if t.cron.matches(candidate.In(t.location)) {
			return candidate
		}
	}
	return time.Time{}
}

func (c *cronSpec) matches(local time.Time) bool {
	if !c.fields[fieldMinute][local.Minute()] ||
		!c.fields[fieldHour][local.Hour()] ||
		!c.fields[fieldMonth][local.Month()] {
		return false
	}
	dayMatch := c.fields[fieldDay][local.Day()]
	weekdayMatch := c.fields[fieldWeekday][local.Weekday()]
	if c.dayRestricted && c.weekdayRestricted {
		return dayMatch || weekdayMatch
	}
	return dayMatch && weekdayMatch
}

// parseField 解析一个字段，返回长度为 max+1 的取值表。
func parseField(raw string, min, max int) ([]bool, error) {
	allowed := make([]bool, max+1)
	selected := false
	for _, part := range strings.Split(raw, ",") {
		start, end, step, err := parseRange(part, min, max)
		if err != nil {
			return nil, err
		}
		for v := start; v <= end; v += step {
			allowed[v] = true
			selected = true
		}
	}
	if !selected {
		return nil, errors.New("field selects no values")
	}
	return allowed, nil
}

// parseRange 解析逗号列表中的一项：*、n、a-b、*/n 或 a-b/n。
func parseRange(part string, min, max int) (start, end, step int, err error) {
	base, stepText, hasStep := strings.Cut(part, "/")
	step = 1
	if hasStep {
		step, err = strconv.Atoi(stepText)
		if err != nil || step < 1 {
			return 0, 0, 0, errors.New("step must be a positive integer")
		}
	}
	if base == "*" {
		return min, max, step, nil
	}
	low, high, isRange := strings.Cut(base, "-")
	if !isRange {
		// 单个数字后不能带步长，例如 5/2。
		if hasStep {
			return 0, 0, 0, errors.New("steps require * or a-b")
		}
		high = low
	}
	if start, err = strconv.Atoi(low); err != nil {
		return 0, 0, 0, fmt.Errorf("unsupported value %q: names and macros are not supported", low)
	}
	if end, err = strconv.Atoi(high); err != nil {
		return 0, 0, 0, fmt.Errorf("unsupported value %q: names and macros are not supported", high)
	}
	if start < min || end > max || start > end {
		return 0, 0, 0, fmt.Errorf("range %q is out of bounds or reversed", base)
	}
	return start, end, step, nil
}
