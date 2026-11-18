// Package scheduler 持久保存静态调度状态与动态任务，并通过协议触发 scheduled Handler。
//
// [Scheduler] 分别用 plugin_id/job_id 和 plugin_id/name/key 保存静态与动态实际运行槽，
// 用本次 Attach 的注册保存定时器与到期队列；[Scheduler.Services] 提供任务和静态调度的固定服务。
// 超时结束 RPC 结果后仍保留槽，直到迟到响应或 [Scheduler.Released] 确认进程退出；
// 插件恢复或定时器落后时只安排未来 tick（CONTRACTS §4.3）。
// 装配处负责激活后 [Scheduler.Attach]、排空时 [Scheduler.Drain]、撤销时 [Scheduler.Detach] 与退出后的 Released。
package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/schedule"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	invocationIDBytes    = 16
	handlerSchemaVersion = 1
	scheduledEventType   = "scheduled"
)

// Options 配置必填状态库、任务上限、调用期限与触发计算使用的时钟。
type Options struct {
	Store            *state.Store // 与生产装配共用的 Core 状态库。
	TasksPerPlugin   int
	TaskPayloadBytes int64 // 参数 JSON 字节上限。
	TaskConcurrency  int
	JobTimeout       time.Duration    // 装配处传入 tune.Values.JobTimeout
	Now              func() time.Time // nil 时使用 time.Now
}

// JobStats 汇总同一 plugin_id/job_id 在本次 Core 运行期间的计数。
type JobStats struct {
	Runs, Skipped, TimedOut, Late uint64
	Running                       bool // 包含结果已超时或取消、实际执行尚未结束的调用
}

// Scheduler 并发安全地安排定时器、占用运行槽并记录任务统计。
// mu 串行化状态库操作、timer、服务和生命周期；IPC 在锁外执行。
type Scheduler struct {
	mu           sync.Mutex // 保护 plugins、注册定时器、jobState 与 jobRun
	store        *state.Store
	maxTasks     int
	payloadBytes int64
	concurrency  int
	jobTimeout   time.Duration
	now          func() time.Time
	plugins      map[string]*pluginJobs
	closed       bool
}

type pluginJobs struct {
	current   *registration // nil 表示已 Detach；jobs 的槽和统计仍保留
	tasks     map[taskKey]*jobState
	running   int // 动态实际执行数，包含超时与协作取消之后的在途调用。
	taskStats TaskStats
	jobs      map[string]*jobState
}

type registration struct {
	plugin router.Plugin
	names  map[string]bool
	tasks  map[taskKey]*registeredTask
	queue  []*queuedTask
	jobs   []*registeredJob
}

type registeredJob struct {
	routeID     string
	jobID       string
	description string // 保留声明的触发规则，供管理列表展示。
	trigger     schedule.Trigger
	timer       *time.Timer // 由 Scheduler.mu 保护，每次只安排一个未来 tick
	saved       state.Job
	next        time.Time
	state       *jobState // Attach 替换注册时仍共享同一 plugin_id/job_id 的运行槽
}

type jobState struct {
	lastAt  time.Time
	stats   JobStats
	running *jobRun
}

type jobRun struct {
	cancel    context.CancelFunc
	dynamic   bool
	reg       *registration
	timedOut  bool // Late 可早于 Call 返回，两个路径在同一锁下只记一次超时
	key       taskKey
	completed func(string, wire.Response, error) // 显式触发的 RPC 结果；在锁外调用一次。
}

// New 创建独立的 Scheduler。
func New(opt Options) *Scheduler {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return &Scheduler{store: opt.Store, maxTasks: opt.TasksPerPlugin, payloadBytes: opt.TaskPayloadBytes, concurrency: opt.TaskConcurrency, jobTimeout: opt.JobTimeout, now: opt.Now, plugins: make(map[string]*pluginJobs)}
}

// Attach 为已通过 REGISTER 校验的 SCHEDULE 路由安排下一次未来触发。
// 同插件重新接入时替换定时器，仍未结束的旧实例调用继续占用原运行槽。
func (s *Scheduler) Attach(ctx context.Context, p router.Plugin, activatedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	// 1. 所有读取与必要写入成功后，才替换本 Session 的定时器。
	savedJobs, err := s.store.Jobs(ctx, p.PluginID())
	if err != nil {
		return err
	}
	tasks, err := s.store.Tasks(ctx, p.PluginID(), "")
	if err != nil {
		return err
	}
	saved := make(map[string]state.Job, len(savedJobs))
	for _, job := range savedJobs {
		saved[job.JobID] = job
	}
	reg := &registration{plugin: p, names: make(map[string]bool), tasks: make(map[taskKey]*registeredTask)}
	for _, route := range p.Routes() {
		switch route.GetKind() {
		case pb.RouteKind_ROUTE_KIND_TASK:
			reg.names[route.GetTask().GetName()] = true
		case pb.RouteKind_ROUTE_KIND_SCHEDULE:
			spec := route.GetSchedule()
			trigger, err := schedule.FromSpec(spec)
			if err != nil {
				panic(err)
			} // REGISTER 已验证。
			job, exists := saved[spec.GetJobId()]
			interval := time.Duration(spec.GetIntervalMs()) * time.Millisecond
			if !exists || job.Interval != interval || (interval != 0 && job.Anchor.IsZero()) {
				job.PluginID, job.JobID, job.Interval = p.PluginID(), spec.GetJobId(), interval
				job.Anchor = time.Time{}
				if interval != 0 {
					job.Anchor = activatedAt
				}
				if err := s.store.SaveJob(ctx, job); err != nil {
					return err
				}
			}
			description := "cron " + spec.GetCron()
			if interval != 0 {
				description = "interval " + interval.String()
			}
			reg.jobs = append(reg.jobs, &registeredJob{routeID: route.GetRouteId(), jobID: spec.GetJobId(), description: description, trigger: trigger, saved: job})
		}
	}
	// 2. 先删除过期 at，再区分已注册和暂时没有处理函数的实例。
	now := s.now()
	plugin := s.plugins[p.PluginID()]
	if plugin == nil {
		plugin = &pluginJobs{jobs: make(map[string]*jobState), tasks: make(map[taskKey]*jobState)}
		s.plugins[p.PluginID()] = plugin
	}
	var unregistered uint64
	for _, task := range tasks {
		if !task.At.IsZero() && !task.At.After(now) {
			if err := s.store.DeleteTask(ctx, p.PluginID(), task.Name, task.Key); err != nil {
				return err
			}
			plugin.taskStats.Expired++
			slog.Warn("expired one-time task discarded", "plugin_id", p.PluginID(), "task_name", task.Name, "key", task.Key, "scheduled_at", task.At)
			continue
		}
		if !reg.names[task.Name] {
			unregistered++
			continue
		}
		registered, err := restoreTask(task)
		if err != nil {
			return err
		}
		reg.tasks[taskKey{task.Name, task.Key}] = registered
	}
	// 3. 发布注册并复用稳定键对应的实际槽；旧回调按 registration 身份失效。
	stopTimers(plugin.current)
	plugin.current = reg
	plugin.taskStats.Unregistered = unregistered
	for _, job := range reg.jobs {
		js := plugin.jobs[job.jobID]
		if js == nil {
			js = &jobState{}
			plugin.jobs[job.jobID] = js
		}
		job.state = js
		s.arm(p.PluginID(), reg, job, now)
	}
	for key, task := range reg.tasks {
		task.state = s.taskState(plugin, key)
		s.armTask(p.PluginID(), reg, task, now)
	}
	return nil
}

// Drain 停止该插件的新 tick，已经触发的调用继续等待实际响应或进程退出。
func (s *Scheduler) Drain(pluginID string) {
	s.Detach(pluginID)
}

// Detach 停止该插件的新触发，保留运行槽与统计。
func (s *Scheduler) Detach(pluginID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if plugin := s.plugins[pluginID]; plugin != nil {
		stopTimers(plugin.current)
		plugin.current = nil
		for key, task := range plugin.tasks {
			if task.running == nil {
				delete(plugin.tasks, key)
			}
		}
	}
}

// Released 在确认旧进程退出后释放其运行槽，保留已接入新 Session 的调用。
func (s *Scheduler) Released(pluginID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if plugin := s.plugins[pluginID]; plugin != nil {
		// 新 Session 已接入时，退出确认只属于旧注册；当前激活注册的槽继续保留。
		keepCurrent := plugin.current != nil && plugin.current.plugin.Activated()
		for _, job := range plugin.jobs {
			if run := job.running; run != nil && (!keepCurrent || run.reg != plugin.current) {
				run.cancel()
				job.running = nil
			}
		}
		for key, task := range plugin.tasks {
			if run := task.running; run != nil && (!keepCurrent || run.reg != plugin.current) {
				run.cancel()
				task.running = nil
				plugin.running--
			}
			if task.running == nil && (plugin.current == nil || plugin.current.tasks[key] == nil) {
				delete(plugin.tasks, key)
			}
		}
		s.pump(pluginID, plugin)
	}
}

// Close 停止全部定时器，不等待在途调用；后续 Attach 不再安排触发。
func (s *Scheduler) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for _, plugin := range s.plugins {
		stopTimers(plugin.current)
		plugin.current = nil
	}
}

// Stats 返回任务的累计计数与实际运行槽状态；未知任务返回零值。
func (s *Scheduler) Stats(pluginID, jobID string) JobStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	if plugin := s.plugins[pluginID]; plugin != nil {
		if job := plugin.jobs[jobID]; job != nil {
			stats := job.stats
			stats.Running = job.running != nil
			return stats
		}
	}
	return JobStats{}
}

func stopTimers(reg *registration) {
	if reg == nil {
		return
	}
	for _, job := range reg.jobs {
		if job.timer != nil {
			job.timer.Stop()
		}
		job.next = time.Time{}
	}
	for _, task := range reg.tasks {
		if task.timer != nil {
			task.timer.Stop()
		}
		task.next = time.Time{}
		task.queued = false
	}
	for _, queued := range reg.queue {
		queued.reject(triggerError(pb.ErrorCode_ERROR_CODE_CANCELLED, "task registration stopped"))
	}
	reg.queue = nil
	reg.tasks = nil
}

// arm 只安排 Next 给出的未来 tick；调用方持有 s.mu。
func (s *Scheduler) arm(pluginID string, reg *registration, job *registeredJob, now time.Time) {
	job.next = time.Time{}
	job.timer = nil
	if job.saved.Paused {
		return
	}
	next := job.trigger.Next(job.saved.Anchor, now)
	job.next = next
	if !next.IsZero() {
		job.timer = time.AfterFunc(next.Sub(now), func() { s.tick(pluginID, reg, job, next) })
	}
}

func (s *Scheduler) tick(pluginID string, reg *registration, job *registeredJob, scheduledAt time.Time) {
	s.mu.Lock()
	if s.closed || s.plugins[pluginID].current != reg || job.saved.Paused || !job.next.Equal(scheduledAt) {
		s.mu.Unlock()
		return // Stop 与已唤醒的回调可以交错，旧注册的 tick 在此失效。
	}
	now := s.now()
	s.arm(pluginID, reg, job, now)
	// 后续轮次已到或本轮期限已过时，旧回调只安排未来 tick，避免补跑和空占槽。
	next := job.trigger.Next(job.saved.Anchor, scheduledAt)
	if (!next.IsZero() && !now.Before(next)) || !now.Before(scheduledAt.Add(s.jobTimeout)) {
		s.mu.Unlock()
		return
	}
	if !reg.plugin.Activated() {
		s.mu.Unlock()
		return
	}
	if job.state.running != nil {
		job.state.stats.Skipped++
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	run := &jobRun{cancel: cancel, reg: reg}
	job.state.running = run
	job.state.stats.Runs++
	job.state.lastAt = now
	s.mu.Unlock()
	// RPC 不占调度锁，不同 job 的用户函数可以并发执行。
	go s.call(ctx, pluginID, reg, job, run, scheduledAt, now, nil)
}

// scheduledRequest 统一派发与帧大小校验使用的完整载荷。
func scheduledRequest(job *registeredJob, scheduledAt, now time.Time, eventID, runID string, task *state.Task) *pb.HandlerRequest {
	event := &pb.Event{EventId: eventID, EventType: scheduledEventType,
		SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_SCHEDULER, OccurredAtUnixMs: now.UnixMilli(),
		EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_Scheduled{Scheduled: &pb.ScheduledPayload{
			JobId: job.jobID, RunId: runID, ScheduledAtUnixMs: scheduledAt.UnixMilli()}}}}
	if task != nil {
		payload := event.EventPayload.GetScheduled()
		payload.TaskName, payload.TaskKey, payload.TaskPayloadJson = task.Name, task.Key, task.PayloadJSON
	}
	return &pb.HandlerRequest{RouteId: job.routeID, Context: &pb.InvocationContext{Scopes: []string{"job"}}, Event: event}
}

func (s *Scheduler) call(ctx context.Context, pluginID string, reg *registration, job *registeredJob, run *jobRun, scheduledAt, now time.Time, task *state.Task) {
	defer run.cancel()
	// 1. 由 Core 构造独立运行 ID 与无 Actor、默认目标的可信调用上下文。
	trace := make([]byte, invocationIDBytes)
	runBytes := make([]byte, invocationIDBytes)
	rand.Read(trace)
	rand.Read(runBytes)
	runID := hex.EncodeToString(runBytes)
	request := scheduledRequest(job, scheduledAt, now, hex.EncodeToString(trace), runID, task)
	event := request.Event
	encoded, err := proto.Marshal(request)
	var response wire.Response
	// 2. 编码成功后才提交 RPC；编码失败也走下面的槽释放和错误记录路径。
	if err == nil {
		deadline := scheduledAt.Add(s.jobTimeout)
		if task != nil {
			deadline = now.Add(s.jobTimeout)
		}
		response, err = reg.plugin.Call(ctx, wire.Request{Target: job.routeID, TraceID: trace,
			Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
				SchemaID: validate.SchemaID(request), SchemaVersion: handlerSchemaVersion, Bytes: encoded},
			Deadline: deadline, Attach: &router.Invocation{Depth: 1, PluginID: pluginID, Event: event, TraceID: trace, Scopes: []string{"job"}, RouteID: job.routeID},
			Late: func(response wire.Response) {
				s.mu.Lock()
				defer s.mu.Unlock()
				if response.LocalErrorCode == pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
					markTimeout(job.state, run)
				}
				// 本地排队撤回只有公开错误，成功或带原始错误码的实际响应才计迟到。
				if response.Err == nil || response.RawErrorCode != pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
					job.state.stats.Late++
				}
				s.releaseRun(pluginID, job.state, run)
			}})
	}
	// 3. 对端已响应的错误也结束实际执行；本地超时、停止取消与断链继续保留槽。
	// 本地超时或取消由 Late 释放实际槽；断链由 Released 在确认进程树清空后释放。
	s.mu.Lock()
	code := wire.CodeOf(err)
	if response.RawErrorCode == pb.ErrorCode_ERROR_CODE_UNSPECIFIED && code == pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		markTimeout(job.state, run)
	} else if response.RawErrorCode != pb.ErrorCode_ERROR_CODE_UNSPECIFIED || (code != pb.ErrorCode_ERROR_CODE_UNAVAILABLE && code != pb.ErrorCode_ERROR_CODE_CANCELLED) {
		s.releaseRun(pluginID, job.state, run)
	}
	s.mu.Unlock()
	logResponse(pluginID, job.routeID, runID, event.GetEventId(), response, err)
	if run.completed != nil {
		run.completed(runID, response, err)
	}
}

// markTimeout 与 releaseRun 都在 Scheduler.mu 下操作该次 run。
func markTimeout(job *jobState, run *jobRun) {
	if !run.timedOut {
		run.timedOut = true
		job.stats.TimedOut++
	}
}

func (s *Scheduler) releaseRun(pluginID string, job *jobState, run *jobRun) {
	if job.running != run {
		return
	} // 旧回调只释放自身，Released 后不能清除新调用的槽。
	job.running = nil
	run.cancel()
	if run.dynamic {
		plugin := s.plugins[pluginID]
		plugin.running--
		if plugin.current == nil || plugin.current.tasks[run.key] == nil {
			delete(plugin.tasks, run.key)
		}
		s.pump(pluginID, plugin)
	}
}

func logResponse(pluginID, routeID, runID, traceID string, response wire.Response, err error) {
	logger := slog.With("plugin_id", pluginID, "route_id", routeID, "run_id", runID, "trace_id", traceID)
	if err != nil {
		logger.Error("scheduled handler failed", "error", err, "raw_error_code", response.RawErrorCode)
	}
	if response.Payload.SchemaID == "" {
		return
	}
	outcome := &pb.HandlerOutcome{}
	if response.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || response.Payload.SchemaID != validate.SchemaID(outcome) || response.Payload.SchemaVersion != handlerSchemaVersion {
		logger.Error("invalid scheduled HandlerOutcome schema", "schema_id", response.Payload.SchemaID, "schema_version", response.Payload.SchemaVersion)
		return
	}
	if err := proto.Unmarshal(response.Payload.Bytes, outcome); err != nil {
		logger.Error("decode scheduled HandlerOutcome failed", "error", err)
		return
	}
	logger.Debug("scheduled handler completed", "outcome", outcome)
}
