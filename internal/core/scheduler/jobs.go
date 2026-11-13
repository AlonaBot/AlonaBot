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
	if job.saved.Paused != paused {
		saved := job.saved
		saved.Paused = paused
		if err := s.store.SaveJob(ctx, saved); err != nil {
			return storageFailure(pluginID, err)
		}
		job.saved = saved
		if job.timer != nil {
			job.timer.Stop()
		}
		job.timer = nil
		job.next = time.Time{}
		if !paused {
			s.arm(pluginID, plugin.current, job, s.now())
		}
	}
	if paused {
		return response(&pb.JobPauseResponse{})
	}
	return response(&pb.JobResumeResponse{})
}
