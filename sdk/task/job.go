package task

import (
	"context"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

// State 保留静态任务的暂停状态与本次 Core 运行的派发时刻。
type State struct {
	Paused         bool
	LastAt, NextAt *time.Time // nil 保留协议 absent；重启后 LastAt 缺失，暂停时 NextAt 缺失。
}

// Pause 暂停静态任务的未来触发，已在途的函数继续执行。
func Pause(ctx context.Context, jobID string) error {
	_, err := registry.Proto[*pb.JobPauseRequest, *pb.JobPauseResponse](wire.ServiceJobPause).Call(ctx, &pb.JobPauseRequest{JobId: jobID})
	return err
}

// Resume 恢复静态任务，interval 沿保存的起点安排未来触发。
func Resume(ctx context.Context, jobID string) error {
	_, err := registry.Proto[*pb.JobResumeRequest, *pb.JobResumeResponse](wire.ServiceJobResume).Call(ctx, &pb.JobResumeRequest{JobId: jobID})
	return err
}

// Status 查询静态任务状态，保留 Core 的 NOT_FOUND 等错误码。
func Status(ctx context.Context, jobID string) (State, error) {
	response, err := registry.Proto[*pb.JobStatusRequest, *pb.JobStatusResponse](wire.ServiceJobStatus).Call(ctx, &pb.JobStatusRequest{JobId: jobID})
	if err != nil {
		return State{}, err
	}
	return State{Paused: response.Paused, LastAt: optionalTaskTime(response.LastAtUnixMs), NextAt: optionalTaskTime(response.NextAtUnixMs)}, nil
}
