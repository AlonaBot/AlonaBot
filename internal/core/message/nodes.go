package message

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// SetIncomingChain 绑定媒体 Store 的查询结果签发入口，复用入站链处理。
func (s *Service) SetIncomingChain(prepare func(*pb.MessageChain, router.Caller, *pb.Target) (*pb.MessageChain, error)) {
	s.prepareIncoming = prepare
}

// ReadNodes 按可信调用者、会话目标与实例账号准入查询，再为接收 Session 签发媒体。
func (s *Service) ReadNodes(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	reject := func(code pb.ErrorCode, message string) wire.Response {
		return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
	}
	request := &pb.ReadNodesRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected ReadNodesRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if err := validate.ReadNodes(request.Segment); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	invocation, _ := in.ParentAttach.(*router.Invocation)
	var defaultTarget *pb.Target
	if invocation != nil {
		defaultTarget = invocation.DefaultTarget
	}
	target := request.To
	if target == nil {
		target = defaultTarget
	} else {
		completed, safe := s.adapters.Complete(target, defaultTarget)
		if safe != nil {
			return wire.Response{Err: safe}
		}
		target = completed
	}
	if target == nil {
		return reject(pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, "no default target")
	}
	if safe := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); safe != nil {
		return wire.Response{Err: safe}
	}
	instance, ok := s.adapters.Resolve(ctx, target.AdapterInstanceId)
	if !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if target.AdapterId != instance.AdapterID() || target.AccountId != instance.AccountID() {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "node target does not match instance")
	}
	reader, ok := instance.(adapter.NodesReader)
	if !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support node reads")
	}
	if !in.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, in.Deadline)
		defer cancel()
	}
	items, safe := reader.ReadNodes(ctx, target, request.Segment)
	chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Forward{Forward: &pb.ForwardSegment{Items: items}}}}}
	// 管理调试返回已登记的对象 ID，后续独立请求按来源实例签发读取令牌。
	if len(items) > 0 && caller.PluginID != router.DebugCallerID {
		prepared, err := s.prepareIncoming(chain, caller, target)
		if err != nil {
			return rejectError(target, err)
		}
		items = prepared.Segments[0].GetForward().Items
	}
	result := &pb.ReadNodesResponse{Items: items}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}, Err: safe}
}
