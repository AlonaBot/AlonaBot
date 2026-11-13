// Package scheduler 按插件声明的 interval 或 cron 触发 scheduled Handler。
//
// [Scheduler] 用 plugin_id/job_id 保存统计与实际运行槽，用本次 Attach 的注册保存定时器。
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
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	invocationIDBytes    = 16
	handlerSchemaVersion = 1
	scheduledEventType   = "scheduled"
)

// Options 配置任务期限与触发计算使用的时钟。
type Options struct {
	JobTimeout time.Duration    // 装配处传入 tune.Values.JobTimeout
	Now        func() time.Time // nil 时使用 time.Now
}

// JobStats 汇总同一 plugin_id/job_id 在本次 Core 运行期间的计数。
type JobStats struct {
	Runs, Skipped, TimedOut, Late uint64
	Running                       bool // 包含结果已超时或取消、实际执行尚未结束的调用
}

// Scheduler 并发安全地安排定时器、占用运行槽并记录任务统计。
// mu 串行化 tick 与 Attach/Detach/Released；RPC 与日志均在锁外进行。
type Scheduler struct {
	mu         sync.Mutex // 保护 plugins、注册定时器、jobState 与 jobRun
	jobTimeout time.Duration
	now        func() time.Time
	plugins    map[string]*pluginJobs
	closed     bool
}

type pluginJobs struct {
	current *registration // nil 表示已 Detach；jobs 的槽和统计仍保留
	jobs    map[string]*jobState
}

type registration struct {
	plugin router.Plugin
	anchor time.Time
	jobs   []*registeredJob
}

type registeredJob struct {
	routeID string
	jobID   string
	trigger schedule.Trigger
	timer   *time.Timer // 由 Scheduler.mu 保护，每次只安排一个未来 tick
	state   *jobState   // Attach 替换注册时仍共享同一 plugin_id/job_id 的运行槽
}

type jobState struct {
	stats   JobStats
	running *jobRun
}

type jobRun struct {
	timedOut bool // Late 可早于 Call 返回，两个路径在同一锁下只记一次超时
}

// New 创建独立的 Scheduler。
func New(opt Options) *Scheduler {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return &Scheduler{jobTimeout: opt.JobTimeout, now: opt.Now, plugins: make(map[string]*pluginJobs)}
}

// Attach 为已通过 REGISTER 校验的 SCHEDULE 路由安排下一次未来触发。
// 同插件重新接入时替换定时器，仍未结束的旧实例调用继续占用原运行槽。
func (s *Scheduler) Attach(p router.Plugin, activatedAt time.Time) {
	// 1. 从已校验的不可变注册描述构造本次触发规则。
	reg := &registration{plugin: p, anchor: activatedAt}
	for _, route := range p.Routes() {
		if route.GetKind() != pb.RouteKind_ROUTE_KIND_SCHEDULE {
			continue
		}
		trigger, err := schedule.FromSpec(route.GetSchedule())
		if err != nil {
			panic(err) // REGISTER 已校验；构造失败意味着调用方破坏了 Attach 的前置条件。
		}
		reg.jobs = append(reg.jobs, &registeredJob{routeID: route.GetRouteId(), jobID: route.GetSchedule().GetJobId(), trigger: trigger})
	}
	pluginID := p.PluginID()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	// 2. 撤销旧定时器，保留按稳定任务 ID 管理的统计与实际运行槽。
	plugin := s.plugins[pluginID]
	if plugin == nil {
		plugin = &pluginJobs{jobs: make(map[string]*jobState)}
		s.plugins[pluginID] = plugin
	}
	stopTimers(plugin.current)
	plugin.current = reg
	now := s.now()
	for _, job := range reg.jobs {
		state := plugin.jobs[job.jobID]
		if state == nil {
			state = &jobState{}
			plugin.jobs[job.jobID] = state
		}
		job.state = state
		s.arm(pluginID, reg, job, now)
	}
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
	}
}

// Released 在确认旧进程退出后释放该插件的全部运行槽。
func (s *Scheduler) Released(pluginID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if plugin := s.plugins[pluginID]; plugin != nil {
		for _, job := range plugin.jobs {
			job.running = nil
		}
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
	}
}

// arm 只安排 Next 给出的未来 tick；调用方持有 s.mu。
func (s *Scheduler) arm(pluginID string, reg *registration, job *registeredJob, now time.Time) {
	next := job.trigger.Next(reg.anchor, now)
	job.timer = nil
	if !next.IsZero() {
		job.timer = time.AfterFunc(next.Sub(now), func() { s.tick(pluginID, reg, job, next) })
	}
}

func (s *Scheduler) tick(pluginID string, reg *registration, job *registeredJob, scheduledAt time.Time) {
	s.mu.Lock()
	if s.closed || s.plugins[pluginID].current != reg {
		s.mu.Unlock()
		return // Stop 与已唤醒的回调可以交错，旧注册的 tick 在此失效。
	}
	now := s.now()
	s.arm(pluginID, reg, job, now)
	// 后续轮次已到或本轮期限已过时，旧回调只安排未来 tick，避免补跑和空占槽。
	next := job.trigger.Next(reg.anchor, scheduledAt)
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
	run := &jobRun{}
	job.state.running = run
	job.state.stats.Runs++
	s.mu.Unlock()
	// RPC 不占调度锁，不同 job 的用户函数可以并发执行。
	go s.call(pluginID, reg, job, run, scheduledAt, now)
}

func (s *Scheduler) call(pluginID string, reg *registration, job *registeredJob, run *jobRun, scheduledAt, now time.Time) {
	// 1. 由 Core 构造独立运行 ID 与无 Actor、默认目标的可信调用上下文。
	trace := make([]byte, invocationIDBytes)
	runBytes := make([]byte, invocationIDBytes)
	rand.Read(trace)
	rand.Read(runBytes)
	runID := hex.EncodeToString(runBytes)
	event := &pb.Event{EventId: hex.EncodeToString(trace), EventType: scheduledEventType,
		SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_SCHEDULER, OccurredAtUnixMs: now.UnixMilli(),
		EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_Scheduled{Scheduled: &pb.ScheduledPayload{
			JobId: job.jobID, RunId: runID, ScheduledAtUnixMs: scheduledAt.UnixMilli()}}}}
	request := &pb.HandlerRequest{RouteId: job.routeID, Context: &pb.InvocationContext{}, Event: event}
	encoded, err := proto.Marshal(request)
	var response wire.Response
	// 2. 编码成功后才提交 RPC；编码失败也走下面的槽释放和错误记录路径。
	if err == nil {
		response, err = reg.plugin.Call(context.Background(), wire.Request{Target: job.routeID, TraceID: trace,
			Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
				SchemaID: validate.SchemaID(request), SchemaVersion: handlerSchemaVersion, Bytes: encoded},
			Deadline: scheduledAt.Add(s.jobTimeout), Attach: &router.Invocation{PluginID: pluginID, Event: event, TraceID: trace},
			Late: func(response wire.Response) {
				s.mu.Lock()
				defer s.mu.Unlock()
				markTimeout(job.state, run)
				// 本地排队撤回只有公开错误，成功或带原始错误码的实际响应才计迟到。
				if response.Err == nil || response.RawErrorCode != pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
					job.state.stats.Late++
				}
				releaseRun(job.state, run)
			}})
	}
	// 3. 对端已响应的错误也结束实际执行；本地超时、停止取消与断链继续保留槽。
	// 停止取消没有 Late 回调，由 Released 在确认进程树清空后释放槽。
	s.mu.Lock()
	code := wire.CodeOf(err)
	if response.RawErrorCode == pb.ErrorCode_ERROR_CODE_UNSPECIFIED && code == pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED {
		markTimeout(job.state, run)
	} else if response.RawErrorCode != pb.ErrorCode_ERROR_CODE_UNSPECIFIED || (code != pb.ErrorCode_ERROR_CODE_UNAVAILABLE && code != pb.ErrorCode_ERROR_CODE_CANCELLED) {
		releaseRun(job.state, run)
	}
	s.mu.Unlock()
	logResponse(pluginID, job.routeID, runID, event.GetEventId(), response, err)
}

// markTimeout 与 releaseRun 都在 Scheduler.mu 下操作该次 run。
func markTimeout(job *jobState, run *jobRun) {
	if !run.timedOut {
		run.timedOut = true
		job.stats.TimedOut++
	}
}

func releaseRun(job *jobState, run *jobRun) {
	if job.running == run {
		job.running = nil // Released 后旧响应可晚于新调用，按指针身份避免释放新槽。
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
