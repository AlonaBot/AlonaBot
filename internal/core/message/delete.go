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

const (
	deleteCapability  = "message.delete"
	maxDeleteMessages = 100
)

// Delete 补齐目标并按发信许可准入，调用一次 Adapter 后保留有序逐项回执与安全错误。
func (s *Service) Delete(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	invocation, _ := in.ParentAttach.(*router.Invocation)
	if invocation != nil && invocation.SendForbidden {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "sending is forbidden in outbound hooks"}}
	}
	reject := func(code pb.ErrorCode, message string) wire.Response {
		return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
	}
	// 1. 在准入和平台操作前校验载荷及全部输入 ID。
	request := &pb.DeleteRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected DeleteRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid DeleteRequest protobuf: "+err.Error())
	}
	if len(request.PlatformMessageIds) == 0 || len(request.PlatformMessageIds) > maxDeleteMessages {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "delete requires 1 to 100 platform message IDs")
	}
	for _, id := range request.PlatformMessageIds {
		if id == "" {
			return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "platform_message_id is empty")
		}
	}
	// 2. 沿用 Send 的目标补齐、原目标豁免及当前实例许可。
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
	instance, ok := s.adapters.Resolve(ctx, target.GetAdapterInstanceId())
	if !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "target adapter or account does not match instance")
	}
	// 3. 声明和实现均支持撤回时才调用 Adapter，不执行出站 Hook。
	supported := false
	for _, capability := range instance.Capabilities() {
		if capability.GetName() == deleteCapability && capability.GetSupported() {
			supported = true
			break
		}
	}
	deleter, ok := instance.(adapter.Deleter)
	if !supported || !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support message deletion")
	}
	result, safe := deleter.Delete(ctx, target, request.PlatformMessageIds)
	// 4. 在 Adapter 信任边界校验回执身份与状态，完整转发有效结果和错误。
	if result == nil {
		if safe == nil {
			return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, "adapter returned no delete result")
		}
		return wire.Response{Err: safe}
	}
	if len(result.Items) != len(request.PlatformMessageIds) {
		return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, "adapter delete result count does not match input")
	}
	for i, item := range result.Items {
		if item.GetPlatformMessageId() != request.PlatformMessageIds[i] {
			return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, "adapter delete result ID does not match input")
		}
		switch item.GetStatus() {
		case pb.PartStatus_PART_STATUS_ACCEPTED, pb.PartStatus_PART_STATUS_REJECTED, pb.PartStatus_PART_STATUS_UNKNOWN, pb.PartStatus_PART_STATUS_NOT_ATTEMPTED:
		default:
			return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, "adapter delete result has invalid status")
		}
		if item.Error != nil {
			item.Error.Code = validate.PublicCode(item.Error.Code)
		}
	}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode DeleteResult: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}, Err: safe}
}
