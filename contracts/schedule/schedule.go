// Package schedule validates and calculates deterministic interval and cron triggers.
package schedule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

type kind uint8

const (
	intervalKind kind = iota + 1
	cronKind
)

type field struct {
	allowed  []bool
	wildcard bool
}
type cronSpec struct {
	minute, hour, day, month, weekday field
	dayRestricted, weekdayRestricted  bool
}
type Trigger struct {
	kind     kind
	interval time.Duration
	cron     cronSpec
	location *time.Location
}

func Interval(d time.Duration) (Trigger, error) {
	if d < time.Second {
		return Trigger{}, errors.New("interval must be at least 1s")
	}
	return Trigger{kind: intervalKind, interval: d}, nil
}
func Cron(expr, timezone string) (Trigger, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	if timezone == "Local" {
		return Trigger{}, errors.New("timezone must be an IANA name, not Local")
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return Trigger{}, fmt.Errorf("invalid IANA timezone %q", timezone)
	}
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return Trigger{}, errors.New("cron must contain exactly five fields")
	}
	minute, err := parseField(parts[0], 0, 59, false)
	if err != nil {
		return Trigger{}, fieldError("minute", err)
	}
	hour, err := parseField(parts[1], 0, 23, false)
	if err != nil {
		return Trigger{}, fieldError("hour", err)
	}
	day, err := parseField(parts[2], 1, 31, false)
	if err != nil {
		return Trigger{}, fieldError("day", err)
	}
	month, err := parseField(parts[3], 1, 12, false)
	if err != nil {
		return Trigger{}, fieldError("month", err)
	}
	weekday, err := parseField(parts[4], 0, 7, true)
	if err != nil {
		return Trigger{}, fieldError("weekday", err)
	}
	spec := cronSpec{minute: minute, hour: hour, day: day, month: month, weekday: weekday, dayRestricted: parts[2] != "*", weekdayRestricted: parts[4] != "*"}
	t := Trigger{kind: cronKind, cron: spec, location: loc}
	// An expression that cannot fire in an eight-year calendar window is rejected.
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if t.nextCron(start, start.Add(-time.Minute)).IsZero() {
		return Trigger{}, errors.New("cron has no occurrence within eight years")
	}
	return t, nil
}
func FromSpec(s *pb.ScheduleSpec) (Trigger, error) {
	if s == nil {
		return Trigger{}, errors.New("schedule spec is required")
	}
	switch x := s.GetTrigger().(type) {
	case *pb.ScheduleSpec_IntervalMs:
		if x.IntervalMs < 1000 {
			return Trigger{}, errors.New("interval must be at least 1000ms")
		}
		return Interval(time.Duration(x.IntervalMs) * time.Millisecond)
	case *pb.ScheduleSpec_Cron:
		return Cron(x.Cron, s.GetTimezone())
	default:
		return Trigger{}, errors.New("exactly one schedule trigger is required")
	}
}
func (t Trigger) Next(anchor, now time.Time) time.Time {
	switch t.kind {
	case intervalKind:
		if t.interval < time.Second || anchor.IsZero() {
			return time.Time{}
		}
		if !now.Before(anchor.Add(t.interval)) {
			elapsed := now.Sub(anchor)
			n := elapsed/t.interval + 1
			return anchor.Add(n * t.interval)
		}
		return anchor.Add(t.interval)
	case cronKind:
		return t.nextCron(anchor, now)
	default:
		return time.Time{}
	}
}
func (t Trigger) nextCron(_ time.Time, now time.Time) time.Time {
	if t.kind != cronKind || t.location == nil {
		return time.Time{}
	}
	candidate := now.UTC().Truncate(time.Minute).Add(time.Minute)
	limit := candidate.AddDate(8, 0, 0)
	for candidate.Before(limit) {
		local := candidate.In(t.location)
		if t.matches(local) {
			return candidate
		}
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}
}
func (t Trigger) matches(local time.Time) bool {
	c := t.cron
	if !c.month.has(int(local.Month())) || !c.hour.has(local.Hour()) || !c.minute.has(local.Minute()) {
		return false
	}
	dayMatch := c.day.has(local.Day())
	weekday := int(local.Weekday())
	weekdayMatch := c.weekday.has(weekday)
	if c.dayRestricted && c.weekdayRestricted {
		if !dayMatch && !weekdayMatch {
			return false
		}
	} else if c.dayRestricted {
		if !dayMatch {
			return false
		}
	} else if c.weekdayRestricted {
		if !weekdayMatch {
			return false
		}
	}
	return true
}
func fieldError(name string, err error) error {
	return fmt.Errorf("invalid cron %s field: %w", name, err)
}
func (f field) has(v int) bool { return v >= 0 && v < len(f.allowed) && f.allowed[v] }
func parseField(raw string, min, max int, weekday bool) (field, error) {
	f := field{allowed: make([]bool, max+1), wildcard: raw == "*"}
	if raw == "" {
		return field{}, errors.New("empty field")
	}
	for _, part := range strings.Split(raw, ",") {
		if part == "" {
			return field{}, errors.New("empty list item")
		}
		base, stepText, hasStep := strings.Cut(part, "/")
		step := 1
		if hasStep {
			if stepText == "" {
				return field{}, errors.New("empty step")
			}
			n, err := strconv.Atoi(stepText)
			if err != nil || n < 1 {
				return field{}, errors.New("step must be a positive integer")
			}
			step = n
		}
		start, end := min, max
		if base != "*" {
			if strings.Contains(base, "-") {
				a, b, ok := strings.Cut(base, "-")
				if !ok || a == "" || b == "" || strings.Contains(b, "-") {
					return field{}, errors.New("invalid range")
				}
				var err error
				start, err = strconv.Atoi(a)
				if err != nil {
					return field{}, errors.New("names and macros are unsupported")
				}
				end, err = strconv.Atoi(b)
				if err != nil {
					return field{}, errors.New("names and macros are unsupported")
				}
				if start < min || start > max || end < min || end > max || start > end {
					return field{}, errors.New("range is out of bounds or reversed")
				}
			} else {
				if hasStep {
					return field{}, errors.New("steps require * or a-b")
				}
				n, err := strconv.Atoi(base)
				if err != nil {
					return field{}, errors.New("names and macros are unsupported")
				}
				if n < min || n > max {
					return field{}, errors.New("value is out of range")
				}
				start, end = n, n
			}
		} else if !hasStep && part != "*" {
			return field{}, errors.New("invalid wildcard")
		}
		for n := start; n <= end; n += step {
			if weekday && n == 7 {
				f.allowed[0] = true
			} else {
				f.allowed[n] = true
			}
		}
	}
	for n := min; n <= max; n++ {
		if f.allowed[n] {
			return f, nil
		}
	}
	return field{}, errors.New("field selects no values")
}
