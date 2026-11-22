package scheduler

import (
	"context"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type triggerResult struct {
	runID    string // 空串表示尚未派发的定位、存储或取消错误。
	response wire.Response
	err      error
}

func triggerError(code pb.ErrorCode, message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func (q *queuedTask) reject(err error) {
	if q.result != nil {
		q.result <- triggerResult{err: err}
	}
}

func (s *Scheduler) triggerPlugin(pluginID string) (*pluginJobs, error) {
	plugin := s.plugins[pluginID]
	if s.closed || plugin == nil || plugin.current == nil || !plugin.current.plugin.Activated() {
		return nil, triggerError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "plugin is not activated")
	}
	return plugin, nil
}

// Trigger 立即派发已激活插件的静态任务，在锁外通知本次 RPC 结果后返回。
func (s *Scheduler) Trigger(ctx context.Context, pluginID, jobID string, completed func(runID string, response wire.Response, err error)) error {
	// 1. 与自然 tick 共用暂停状态与实际运行槽，拒绝重叠显式触发。
	s.mu.Lock()
	plugin, err := s.triggerPlugin(pluginID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	reg := plugin.current
	var job *registeredJob
	for _, candidate := range reg.jobs {
		if candidate.jobID == jobID {
			job = candidate
			break
		}
	}
	if job == nil {
		s.mu.Unlock()
		return triggerError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "static job is not registered")
	}
	if job.saved.Paused || job.state.running != nil {
		s.mu.Unlock()
		return triggerError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "static job is paused or running")
	}
	runContext, cancel := context.WithCancel(ctx)
	run := &jobRun{cancel: cancel, completed: completed}
	job.state.running = run
	job.state.stats.Runs++
	now := s.now()
	s.mu.Unlock()
	// 2. 结果与 Late 槽释放继续交给统一调用路径；保留自然周期起点。
	s.call(runContext, pluginID, reg, job, run, now, now, nil)
	return nil
}

// TriggerTask 将保存的实例立即放入正常到期队列，通知本次 RPC 结果后返回。
func (s *Scheduler) TriggerTask(ctx context.Context, pluginID, taskName, key string, completed func(runID string, response wire.Response, err error)) error {
	// 1. 接收者先登记，再进入并发队列，快速完成也能保留结果。
	s.mu.Lock()
	plugin, err := s.triggerPlugin(pluginID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	reg := plugin.current
	task := reg.tasks[taskKey{taskName, key}]
	if running := plugin.tasks[taskKey{taskName, key}]; running != nil && running.running != nil {
		s.mu.Unlock()
		return triggerError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "task instance is running")
	}
	if task == nil {
		s.mu.Unlock()
		return triggerError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "task instance not found")
	}
	if task.queued {
		s.mu.Unlock()
		return triggerError(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "task instance is queued")
	}
	queued := &queuedTask{task: task, scheduledAt: s.now(), ctx: ctx, result: make(chan triggerResult, 1)}
	s.enqueue(reg, queued)
	s.pump(pluginID, plugin)
	s.mu.Unlock()
	// 2. 排队取消只撤回本轮触发；已经派发的调用由 Session 产生取消结果。
	var result triggerResult
	select {
	case result = <-queued.result:
	case <-ctx.Done():
		s.mu.Lock()
		index := slices.Index(reg.queue, queued)
		if index >= 0 {
			reg.queue = slices.Delete(reg.queue, index, index+1)
			task.queued = false
		}
		s.mu.Unlock()
		if index >= 0 {
			return ctx.Err()
		}
		result = <-queued.result
	}
	if result.runID == "" {
		return result.err
	}
	completed(result.runID, result.response, result.err)
	return nil
}
