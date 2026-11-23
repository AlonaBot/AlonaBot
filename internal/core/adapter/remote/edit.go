package remote

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

var _ adapter.Editor = (*instance)(nil)

// Edit 经创建实例时绑定的 Session 编辑原消息，不自动重试。
func (entry *instance) Edit(ctx context.Context, target *pb.Target, platformMessageID string, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
	// 1. 检查引用的实例仍在运行，旧引用沿用创建时绑定的 Session。
	p := entry.process
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if !running {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	// 2. 编码请求并提交唯一一次编辑调用。
	request, err := payload(&pb.AdapterEditRequest{InstanceId: entry.id, InstanceGeneration: entry.generation, Target: target, PlatformMessageId: platformMessageID, Chain: chain})
	if err != nil {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	response, err := entry.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterEdit, Payload: request})
	// 3. 先解码有效回执，再完整保留同时返回的错误。
	var result *pb.SendResult
	if response.Payload.SchemaID != "" {
		result = &pb.SendResult{}
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
