package plugintest

import (
	"time"

	"github.com/AlonaBot/AlonaBot/internal/clock"
)

// Clock 同时控制 Scheduler 与 continuation 的当前时间、到期和取消。
type Clock = clock.Manual

// NewClock 创建从 start 开始的可推进测试时钟。
func NewClock(start time.Time) *Clock { return clock.New(start) }

// WithClock 将同一测试时钟注入 Core 的调度器与会话等待。
func WithClock(c *Clock) Option { return func(o *options) { o.clock = c } }
