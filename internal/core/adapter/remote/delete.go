package remote

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

var _ adapter.Deleter = (*instance)(nil)

// Delete 经创建实例时绑定的 Session 撤回消息，不自动重试。
func (entry *instance) Delete(ctx context.Context, target *pb.Target, platformMessageIDs []string) (*pb.DeleteResult, *pb.ErrorInfo) {
	// 1. 旧引用只使用创建时绑定的 Session，停止后拒绝操作。
	p := entry.process
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if !running {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	request, err := payload(&pb.AdapterDeleteRequest{InstanceId: entry.id, InstanceGeneration: entry.generation, Target: target, PlatformMessageIds: platformMessageIDs})
	if err != nil {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	response, err := entry.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterDelete, Payload: request})
	// 2. 有效回执与安全错误同时保留，超时和断链沿用 Session 的未知语义。
	var result *pb.DeleteResult
	if response.Payload.SchemaID != "" {
		result = &pb.DeleteResult{}
		if err := decode(response.Payload, result); err != nil {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
		}
	}
	if response.Err != nil {
		return result, response.Err
	}
	if err != nil {
		return result, &pb.ErrorInfo{Code: wire.CodeOf(err), SafeMessage: err.Error()}
	}
	return result, nil
}
