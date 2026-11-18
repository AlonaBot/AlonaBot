package idle

import (
	"slices"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// Controller 持有逐插件冷却进度；调用方在 Core 生命周期循环内串行使用。
type Controller struct {
	values  tune.Values
	plugins map[string]*coldStarts
}

type coldStarts struct {
	processed uint64 // 启动序号跨越 Waker 历史回收；冷却期间同样只消费一次。
	recent    []time.Time
	until     time.Time
}

// New 从本次 Core 生效调优值建立空闲决策器。
func New(values tune.Values) *Controller {
	return &Controller{values: values, plugins: make(map[string]*coldStarts)}
}

// CanStop 消费实际启动记录，并检查生效时长、最短运行、就绪状态和零活动引用。
// 返回 true 后，调用方仍须经 StopIdle 在活动门禁内复核本次快照。
func (c *Controller) CanStop(pluginID string, policy Policy, runtime state.RuntimeStatus, activity supervisor.ActivityStatus, now time.Time) bool {
	starts := c.plugins[pluginID]
	if starts == nil {
		starts = &coldStarts{}
		c.plugins[pluginID] = starts
	}
	// 1. 从越过阈值的真实启动时刻固定冷却截止，期间的新启动只标记为已处理。
	first := activity.ColdStartSequence - uint64(len(activity.ColdStarts))
	for i, at := range activity.ColdStarts {
		sequence := first + uint64(i) + 1
		if sequence <= starts.processed {
			continue
		}
		starts.processed = sequence
		if at.Before(starts.until) {
			continue
		}
		starts.recent = slices.DeleteFunc(starts.recent, func(previous time.Time) bool {
			return at.Sub(previous) >= c.values.SleepColdStartWindow
		})
		starts.recent = append(starts.recent, at)
		if len(starts.recent) > c.values.SleepColdStartMax {
			starts.until = at.Add(c.values.SleepColdStartWindow)
			starts.recent = nil
		}
	}
	starts.recent = slices.DeleteFunc(starts.recent, func(at time.Time) bool {
		return now.Sub(at) >= c.values.SleepColdStartWindow
	})
	// 2. 阈值变更沿用当前 Session 的实际活动结束时刻。
	if !policy.Eligible || runtime.State != state.RuntimeActive || !activity.Ready || activity.References != 0 || now.Before(starts.until) {
		return false
	}
	duration, _ := time.ParseDuration(policy.EffectivePolicy) // Resolve 只发布 settings 已校验的正时长。
	return now.Sub(runtime.LastWakeAt) >= c.values.SleepMinUptime && now.Sub(activity.LastActivityAt) >= duration
}
