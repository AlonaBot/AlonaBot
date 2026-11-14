package scheduler

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/schedule"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// TaskStats 汇总本次 Core 的过期计数、当前缺少处理函数的实例和实际动态槽。
type TaskStats struct {
	Expired         uint64
	Unregistered    uint64
	Running, Queued int
}

type taskKey struct{ name, key string }
type registeredTask struct {
	definition state.Task
	trigger    schedule.Trigger // at 使用 definition.At；周期触发复用 contracts/schedule。
	timer      *time.Timer
	next       time.Time
	state      *jobState
	queued     bool
}
type queuedTask struct {
	task        *registeredTask
	scheduledAt time.Time
	ctx         context.Context    // nil 为自然到期，使用后台 context。
	result      chan triggerResult // nil 为自然到期；缓冲接收准入失败或 RPC 结果。
}

// TasksStats 返回该插件的动态任务诊断快照。
func (s *Scheduler) TasksStats(pluginID string) TaskStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	plugin := s.plugins[pluginID]
	if plugin == nil {
		return TaskStats{}
	}
	stats := plugin.taskStats
	stats.Running = plugin.running
	if plugin.current != nil {
		stats.Queued = len(plugin.current.queue)
	}
	return stats
}

func restoreTask(definition state.Task) (*registeredTask, error) {
	task := &registeredTask{definition: definition}
	var err error
	if definition.Interval != 0 {
		task.trigger, err = schedule.Interval(definition.Interval)
	} else if definition.Cron != "" {
		task.trigger, err = schedule.Cron(definition.Cron, definition.Timezone)
	}
	return task, err
}

func (s *Scheduler) taskState(plugin *pluginJobs, key taskKey) *jobState {
	js := plugin.tasks[key]
	if js == nil {
		js = &jobState{}
		plugin.tasks[key] = js
	}
	return js
}

// armTask、tickTask 和 pump 共用 s.mu，替换、取消与派发只有一个提交顺序。
func (s *Scheduler) armTask(pluginID string, reg *registration, task *registeredTask, now time.Time) {
	next := task.definition.At
	if next.IsZero() {
		next = task.trigger.Next(task.definition.Anchor, now)
	} else if !next.After(now) {
		next = time.Time{}
	}
	task.next = next
	task.timer = nil
	if !next.IsZero() {
		task.timer = time.AfterFunc(next.Sub(now), func() { s.tickTask(pluginID, reg, task, next) })
	}
}

func (s *Scheduler) tickTask(pluginID string, reg *registration, task *registeredTask, scheduledAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plugin := s.plugins[pluginID]
	key := taskKey{task.definition.Name, task.definition.Key}
	if s.closed || plugin.current != reg || reg.tasks[key] != task || !task.next.Equal(scheduledAt) {
		return
	}
	now := s.now()
	if now.Before(scheduledAt) {
		s.armTask(pluginID, reg, task, now)
		return
	}
	s.armTask(pluginID, reg, task, now)
	if !reg.plugin.Activated() {
		return
	}
	if task.definition.At.IsZero() {
		next := task.trigger.Next(task.definition.Anchor, scheduledAt)
		if !next.IsZero() && !now.Before(next) {
			return
		} // 落后的周期只安排未来 tick。
	}
	if task.state.running != nil || task.queued {
		return
	}
	s.enqueue(reg, &queuedTask{task: task, scheduledAt: scheduledAt})
	s.pump(pluginID, plugin)
}

func (s *Scheduler) enqueue(reg *registration, queued *queuedTask) {
	queued.task.queued = true
	reg.queue = append(reg.queue, queued)
	slices.SortFunc(reg.queue, func(a, b *queuedTask) int {
		if c := a.scheduledAt.Compare(b.scheduledAt); c != 0 {
			return c
		}
		if c := cmp.Compare(a.task.definition.Name, b.task.definition.Name); c != 0 {
			return c
		}
		return cmp.Compare(a.task.definition.Key, b.task.definition.Key)
	})
}

// pump 只持锁操作库与运行记录；真正的 IPC 由新 goroutine 执行。
func (s *Scheduler) pump(pluginID string, plugin *pluginJobs) {
	reg := plugin.current
	if s.closed || reg == nil || !reg.plugin.Activated() {
		return
	}
	for plugin.running < s.concurrency && len(reg.queue) > 0 {
		queued := reg.queue[0]
		reg.queue = slices.Delete(reg.queue, 0, 1)
		task := queued.task
		task.queued = false
		key := taskKey{task.definition.Name, task.definition.Key}
		if reg.tasks[key] != task || task.state.running != nil {
			continue
		}
		// 一次性定义先事务删除；失败保留持久定义，记录错误并放弃本轮。
		if !task.definition.At.IsZero() {
			if err := s.store.DeleteTask(context.Background(), pluginID, key.name, key.key); err != nil {
				slog.Error("delete one-time task before dispatch failed", "plugin_id", pluginID, "task_name", key.name, "key", key.key, "error", err)
				queued.reject(fmt.Errorf("delete one-time task before dispatch: %w", err))
				continue
			}
			if task.timer != nil {
				task.timer.Stop()
			}
			delete(reg.tasks, key)
		}
		parent := queued.ctx
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithCancel(parent)
		run := &jobRun{cancel: cancel, dynamic: true, reg: reg, key: key}
		if queued.result != nil {
			run.completed = func(id string, response wire.Response, err error) {
				queued.result <- triggerResult{runID: id, response: response, err: err}
			}
		}
		task.state.running = run
		plugin.running++
		job := &registeredJob{routeID: "task." + key.name, jobID: "task." + key.name, state: task.state}
		// 派发的时刻决定动态期限；排队时刻仍留在 ScheduledPayload。
		go s.call(ctx, pluginID, reg, job, run, queued.scheduledAt, s.now(), &task.definition)
	}
}

func removeTask(reg *registration, key taskKey) {
	if task := reg.tasks[key]; task != nil {
		if task.timer != nil {
			task.timer.Stop()
		}
		task.queued = false
		delete(reg.tasks, key)
	}
	reg.queue = slices.DeleteFunc(reg.queue, func(q *queuedTask) bool {
		if q.task.definition.Name == key.name && q.task.definition.Key == key.key {
			q.reject(triggerError(pb.ErrorCode_ERROR_CODE_CANCELLED, "task instance replaced or cancelled"))
			return true
		}
		return false
	})
}
