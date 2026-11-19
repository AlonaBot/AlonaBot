package scheduler

import (
	"context"
	"errors"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"google.golang.org/protobuf/proto"
)

// Suspend 撤销正常空闲停止的交付绑定，保留计划、到期队列与未来定时器。
// 调用方已经取得 IDLE 停止权；运行状态仍由 Supervisor 写入。
func (s *Scheduler) Suspend(pluginID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if plugin := s.plugins[pluginID]; plugin != nil && plugin.current != nil {
		plugin.current.suspended = true
	}
}

func equalRoutes(a, b router.Plugin) bool {
	x, y := a.Routes(), b.Routes()
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if !proto.Equal(x[i], y[i]) {
			return false
		}
	}
	return true
}

// canSchedule 只允许活 Session 或本次正常休眠保留的计划接纳自然到期。
func (s *Scheduler) canSchedule(reg *registration) bool {
	return reg.suspended && s.ensureActive != nil || !reg.suspended && reg.plugin.Activated()
}

// cancelWaiting 在撤销计划时打断尚未交付的等待；已发出的请求由正常排空收尾。
// 调用方持有 s.mu；等待者随后自行归还稳定槽和 Waker 预算。
func (s *Scheduler) cancelWaiting(plugin *pluginJobs) {
	for _, job := range plugin.jobs {
		if run := job.running; run != nil && run.delivery == nil {
			run.cancel()
		}
	}
	for _, task := range plugin.tasks {
		if run := task.running; run != nil && run.delivery == nil {
			run.cancel()
		}
	}
}

// destination 在锁外等待 Waker；Ready 返回后再在锁内核对计划和槽的所有权。
func (s *Scheduler) destination(ctx context.Context, pluginID string, reg *registration, job *registeredJob, run *jobRun, payloadBytes int64) (router.Plugin, func(), error) {
	release := func() {}
	if s.ensureActive != nil {
		var err error
		release, err = s.ensureActive(ctx, pluginID, payloadBytes)
		if err != nil {
			return nil, nil, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := ctx.Err()
	if err != nil {
		code := pb.ErrorCode_ERROR_CODE_CANCELLED
		if errors.Is(err, context.DeadlineExceeded) {
			code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
		}
		err = errors.Join(triggerError(code, err.Error()), err)
	} else if s.closed || s.plugins[pluginID].current != reg || reg.suspended || job.state.running != run || !reg.plugin.Activated() {
		err = triggerError(pb.ErrorCode_ERROR_CODE_CANCELLED, "scheduled registration is no longer active")
	}
	if err != nil {
		release()
		return nil, nil, err
	}
	run.delivery = reg.plugin
	job.state.lastAt = s.now()
	return run.delivery, release, nil
}
