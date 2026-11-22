package scheduler

import (
	"context"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

func (s *Scheduler) job(ctx context.Context, pluginID string, plugin *pluginJobs, in *wire.Inbound) wire.Response {
	var request proto.Message
	var id string
	switch in.Target {
	case wire.ServiceJobPause:
		request = &pb.JobPauseRequest{}
	case wire.ServiceJobResume:
		request = &pb.JobResumeRequest{}
	case wire.ServiceJobStatus:
		request = &pb.JobStatusRequest{}
	}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	switch r := request.(type) {
	case *pb.JobPauseRequest:
		id = r.JobId
	case *pb.JobResumeRequest:
		id = r.JobId
	case *pb.JobStatusRequest:
		id = r.JobId
	}
	var job *registeredJob
	for _, j := range plugin.current.jobs {
		if j.jobID == id {
			job = j
			break
		}
	}
	if job == nil {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "static job is not registered")
	}
	if in.Target == wire.ServiceJobStatus {
		result := &pb.JobStatusResponse{Paused: job.saved.Paused}
		if !job.state.lastAt.IsZero() {
			result.LastAtUnixMs = proto.Int64(job.state.lastAt.UnixMilli())
		}
		if !job.next.IsZero() {
			result.NextAtUnixMs = proto.Int64(job.next.UnixMilli())
		}
		return response(result)
	}
	paused := in.Target == wire.ServiceJobPause
	if err := s.setJobPaused(ctx, pluginID, plugin.current, job, paused); err != nil {
		return storageFailure(pluginID, err)
	}
	if paused {
		return response(&pb.JobPauseResponse{})
	}
	return response(&pb.JobResumeResponse{})
}

// setJobPaused 由持锁调用方共用持久化与计时器切换，提交失败时保留原状态。
func (s *Scheduler) setJobPaused(ctx context.Context, pluginID string, reg *registration, job *registeredJob, paused bool) error {
	if job.saved.Paused != paused {
		saved := job.saved
		saved.Paused = paused
		if err := s.store.SaveJob(ctx, saved); err != nil {
			return err
		}
		job.saved = saved
		if job.timer != nil {
			job.timer.Stop()
		}
		job.timer = nil
		job.next = time.Time{}
		if !paused {
			s.arm(pluginID, reg, job, s.now())
		}
	}
	return nil
}

// StaticJob 返回本次注册的静态任务及运行状态，零时刻表示尚未运行或没有未来触发。
type StaticJob struct {
	ID, Trigger          string
	Paused               bool
	LastRunAt, NextRunAt time.Time
}

// Jobs 列出保留计划中的静态任务；已撤销注册时返回空列表。
func (s *Scheduler) Jobs(pluginID string) []StaticJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]StaticJob, 0)
	if plugin := s.plugins[pluginID]; plugin != nil && plugin.current != nil {
		for _, job := range plugin.current.jobs {
			jobs = append(jobs, StaticJob{job.jobID, job.description, job.saved.Paused, job.state.lastAt, job.next})
		}
	}
	return jobs
}

// SetJobPaused 通过与插件 RPC 相同的状态切换暂停或恢复静态任务。
func (s *Scheduler) SetJobPaused(ctx context.Context, pluginID, jobID string, paused bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plugin := s.plugins[pluginID]
	if plugin != nil && plugin.current != nil {
		for _, job := range plugin.current.jobs {
			if job.jobID == jobID {
				if _, err := s.triggerPlugin(pluginID); err != nil {
					return err
				}
				return s.setJobPaused(ctx, pluginID, plugin.current, job, paused)
			}
		}
	}
	return triggerError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "static job is not registered")
}
