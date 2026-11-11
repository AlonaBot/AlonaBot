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

const editCapability = "message.edit"

// Edit 按 Send 的目标准入规则校验编辑请求，再由支持编辑的实例修改原消息。
// Adapter 返回的回执和错误完整转发；失败时保持单次编辑，不发送新消息。
func (s *Service) Edit(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 取得可信默认目标，校验协议载荷。
	invocation, _ := in.ParentAttach.(*router.Invocation)
	var target *pb.Target
	if invocation != nil {
		target = invocation.DefaultTarget
	}
	reject := func(code pb.ErrorCode, message string) wire.Response {
		return editRejected(target, &pb.ErrorInfo{Code: code, SafeMessage: message})
	}
	request := &pb.EditRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected EditRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid EditRequest protobuf: "+err.Error())
	}
	// 2. 显式目标优先，随后使用与 Send 相同的许可检查。
	if request.To != nil {
		target = request.To
	}
	if target == nil {
		return reject(pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, "no default target")
	}
	if safe := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); safe != nil {
		return editRejected(target, safe)
	}
	// 3. 校验原消息 ID、片段结构以及编辑专属的组合限制。
	if request.PlatformMessageId == "" {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "platform_message_id is empty")
	}
	if err := validate.Chain(request.Chain); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	texts := 0
	for _, segment := range request.Chain.Segments {
		switch segment.GetData().(type) {
		case *pb.Segment_Text:
			texts++
		case *pb.Segment_Buttons:
			// buttons 的数量和内容已由 validate.Chain 校验。
		default:
			return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "edit chain allows only text and buttons segments")
		}
	}
	if texts != 1 {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "edit chain requires exactly one text segment")
	}
	// 4. 校验实例身份绑定与显式声明的编辑能力。
	instance, ok := s.adapters.Get(target.GetAdapterInstanceId())
	if !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "target adapter or account does not match instance")
	}
	supported := false
	for _, capability := range instance.Capabilities() {
		if capability.GetName() == editCapability && capability.GetSupported() {
			supported = true
			break
		}
	}
	editor, ok := instance.(adapter.Editor)
	if !supported || !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support message editing")
	}
	// 5. 执行唯一一次编辑，并完整返回平台回执与错误。
	// shortcut: 编辑前后尚无出站 Hook，接入时以 edit 和 platform_message_id 标明操作并沿用 Send 的调用位置。
	result, safe := editor.Edit(ctx, target, request.PlatformMessageId, request.Chain)
	return response(result, safe)
}

func editRejected(target *pb.Target, safe *pb.ErrorInfo) wire.Response {
	return response(&pb.SendResult{
		Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target,
		PartResults: []*pb.PartResult{{Status: pb.PartStatus_PART_STATUS_REJECTED, Error: safe}},
	}, safe)
}
