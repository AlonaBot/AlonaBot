package sdk

import (
	"context"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// JobState 保留静态任务的暂停状态与本次 Core 运行的派发时刻。
type JobState struct {
	Paused         bool
	LastAt, NextAt *time.Time // nil 保留协议 absent；重启后 LastAt 缺失，暂停时 NextAt 缺失。
}

// PauseJob 暂停静态任务的未来触发，已在途的函数继续执行。
func PauseJob(ctx context.Context, jobID string) error {
	_, err := Proto[*pb.JobPauseRequest, *pb.JobPauseResponse](wire.ServiceJobPause).Call(ctx, &pb.JobPauseRequest{JobId: jobID})
	return err
}

// ResumeJob 恢复静态任务，interval 沿保存的起点安排未来触发。
func ResumeJob(ctx context.Context, jobID string) error {
	_, err := Proto[*pb.JobResumeRequest, *pb.JobResumeResponse](wire.ServiceJobResume).Call(ctx, &pb.JobResumeRequest{JobId: jobID})
	return err
}

// JobStatus 查询静态任务状态，保留 Core 的 NOT_FOUND 等错误码。
func JobStatus(ctx context.Context, jobID string) (JobState, error) {
	response, err := Proto[*pb.JobStatusRequest, *pb.JobStatusResponse](wire.ServiceJobStatus).Call(ctx, &pb.JobStatusRequest{JobId: jobID})
	if err != nil {
		return JobState{}, err
	}
	return JobState{Paused: response.Paused, LastAt: optionalTaskTime(response.LastAtUnixMs), NextAt: optionalTaskTime(response.NextAtUnixMs)}, nil
}
