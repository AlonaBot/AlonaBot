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
	var defaultTarget *pb.Target
	if invocation != nil {
		defaultTarget = invocation.DefaultTarget
	}
	target := defaultTarget
	reject := func(code pb.ErrorCode, message string) wire.Response {
		return editRejected(target, &pb.ErrorInfo{Code: code, SafeMessage: message})
	}
	if invocation != nil && invocation.SendForbidden {
		return reject(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "editing is forbidden in outbound hooks")
	}
	request := &pb.EditRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected EditRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid EditRequest protobuf: "+err.Error())
	}
	// 2. 显式目标先补齐实例与账号，再按与 Send 相同的规则检查许可。
	if request.To != nil {
		completed, completionError := s.adapters.Complete(request.To, defaultTarget)
		if completionError != nil {
			return editRejected(request.To, completionError)
		}
		target = completed
	}
	if target == nil {
		return reject(pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, "no default target")
	}
	if safe := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); safe != nil {
		// 准入拒绝只回传调用方已知的目标，避免带出补齐的实例身份。
		rejectedTarget := request.To
		if rejectedTarget == nil {
			rejectedTarget = defaultTarget
		}
		return editRejected(rejectedTarget, safe)
	}
	// 3. 校验原消息 ID、片段结构以及编辑专属的组合限制。
	if request.PlatformMessageId == "" {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "platform_message_id is empty")
	}
	if safe := checkEditChain(request.Chain); safe != nil {
		return editRejected(target, safe)
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
	chain, safe := s.preserveReply(caller, target, request.PlatformMessageId, request.Chain, instance)
	if safe != nil {
		return editRejected(target, safe)
	}
	request.Chain = chain
	if safe := checkEditChain(request.Chain); safe != nil {
		return editRejected(target, safe)
	}
	// 5. 改写后仍须符合编辑片段限制，拒绝时不签发按钮、不调用平台。
	if s.hooks != nil {
		chain, safe := s.hooks.BeforeSend(ctx, pb.SendOperation_SEND_OPERATION_EDIT, request.PlatformMessageId, target, request.Chain, invocation, checkEditChain)
		if safe != nil {
			return editRejected(target, safe)
		}
		if chain != nil {
			request.Chain = chain
		}
	}
	// 6. 复用发送的签发入口，为本次编辑的按钮生成新 token。
	if s.issue != nil {
		if err := s.issue(ctx, caller, target, request.Chain); err != nil {
			return editRejected(target, rejectError(target, err).Err)
		}
	}
	// 7. 执行唯一一次编辑，观察 Hook 的结果不改变平台回执与错误。
	result, safe := editor.Edit(ctx, target, request.PlatformMessageId, request.Chain)
	if s.hooks != nil {
		s.hooks.AfterSend(ctx, pb.SendOperation_SEND_OPERATION_EDIT, request.PlatformMessageId, target, request.Chain, result, safe, invocation)
	}
	return response(result, safe)
}

func checkEditChain(chain *pb.MessageChain) *pb.ErrorInfo {
	invalid := func(message string) *pb.ErrorInfo {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: message}
	}
	if err := validate.Chain(chain); err != nil {
		return invalid(err.Error())
	}
	texts := 0
	for _, segment := range chain.Segments {
		switch segment.GetData().(type) {
		case *pb.Segment_Text:
			texts++
		case *pb.Segment_Buttons:
			// buttons 的数量和内容已由 validate.Chain 校验。
		default:
			return invalid("edit chain allows only text and buttons segments")
		}
	}
	if texts != 1 {
		return invalid("edit chain requires exactly one text segment")
	}
	return nil
}

func editRejected(target *pb.Target, safe *pb.ErrorInfo) wire.Response {
	return response(&pb.SendResult{
		Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target,
		PartResults: []*pb.PartResult{{Status: pb.PartStatus_PART_STATUS_REJECTED, Error: safe}},
	}, safe)
}
