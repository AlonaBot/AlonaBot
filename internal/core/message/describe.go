package message

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// SetCheckSource 绑定媒体 Store 的只读授权与可信 MIME 查询，在开放服务前装配。
func (s *Service) SetCheckSource(prepare func(context.Context, router.Caller, *pb.Target, *pb.MediaSource) (*pb.MediaSource, error)) {
	s.checkSource = prepare
}

// resolveTarget 复用发送的目标补齐、发信许可及实例账号绑定。
func (s *Service) resolveTarget(ctx context.Context, caller router.Caller, invocation *router.Invocation, requested *pb.Target) (*pb.Target, adapter.Instance, *pb.ErrorInfo) {
	var defaultTarget *pb.Target
	if invocation != nil {
		defaultTarget = invocation.DefaultTarget
	}
	target := requested
	if target == nil {
		target = defaultTarget
	} else {
		completed, safe := s.adapters.Complete(target, defaultTarget)
		if safe != nil {
			return nil, nil, safe
		}
		target = completed
	}
	if target == nil {
		return nil, nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, SafeMessage: "no default target"}
	}
	if safe := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); safe != nil {
		return nil, nil, safe
	}
	instance, ok := s.adapters.Resolve(ctx, target.GetAdapterInstanceId())
	if !ok {
		return nil, nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	if target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return nil, nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "target adapter or account does not match instance"}
	}
	return target, instance, nil
}

// DescribeMessage 返回当前实例的实际声明；旧实例缺少声明时返回空描述。
func (s *Service) DescribeMessage(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	ctx, cancel := s.requestContext(ctx, in)
	defer cancel()
	request := &pb.DescribeMessageRequest{}
	if safe := decodeMessageQuery(in, request); safe != nil {
		return wire.Response{Err: safe}
	}
	invocation, _ := in.ParentAttach.(*router.Invocation)
	_, instance, safe := s.resolveTarget(ctx, caller, invocation, request.To)
	if safe != nil {
		return wire.Response{Err: safe}
	}
	if describer, ok := instance.(adapter.MessageDescriber); ok {
		description, safe := describer.DescribeMessage(ctx)
		return messageQueryResponse(description, safe)
	}
	return messageQueryResponse(&pb.MessageDescription{}, nil)
}

// CheckMessage 在结构与来源授权检查后调用可选预检，始终保留状态及具体错误。
func (s *Service) CheckMessage(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	ctx, cancel := s.requestContext(ctx, in)
	defer cancel()
	request := &pb.CheckMessageRequest{}
	if safe := decodeMessageQuery(in, request); safe != nil {
		return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED, safe)
	}
	invocation, _ := in.ParentAttach.(*router.Invocation)
	if invocation != nil && invocation.SendForbidden {
		return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "sending is forbidden in outbound hooks"})
	}
	// 1. 与发送共用目标准入及基础校验，调用者填写的 source.Mime 在此拒绝。
	target, instance, safe := s.resolveTarget(ctx, caller, invocation, request.To)
	if safe != nil {
		return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, safe)
	}
	if describer, ok := instance.(adapter.MessageDescriber); ok {
		description, safe := describer.DescribeMessage(ctx)
		if safe != nil {
			return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, safe)
		}
		if safe := checkDeclaredFormats(description, request.Chain); safe != nil {
			return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED, safe)
		}
	}
	if safe := s.checkChain(request.Chain, target); safe != nil {
		return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED, safe)
	}
	if safe := checkExtensions(instance, request.Chain); safe != nil {
		return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED, safe)
	}
	// 2. 来源保留原 token 与类别，Store 只在授权成功后填可信 MIME。
	for _, source := range mediaSources(request.Chain) {
		prepared, err := s.checkSource(ctx, caller, target, source)
		if err != nil {
			return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, messageQueryError(err))
		}
		source.Mime = prepared.Mime
	}
	if err := ctx.Err(); err != nil {
		return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, messageQueryError(err))
	}
	// 3. 可选接口缺失保留 unknown；查询及 RPC 的真实错误原样返回。
	if checker, ok := instance.(adapter.MessageChecker); ok {
		status, safe := checker.CheckMessage(ctx, target, request.Chain)
		return checkResponse(status, safe)
	}
	return checkResponse(pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, nil)
}
func decodeMessageQuery(in *wire.Inbound, request proto.Message) *pb.ErrorInfo {
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: fmt.Sprintf("expected %s protobuf schema version 1", request.ProtoReflect().Descriptor().Name())}
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: err.Error()}
	}
	return nil
}
func checkResponse(status pb.SupportStatus, safe *pb.ErrorInfo) wire.Response {
	return messageQueryResponse(&pb.CheckMessageResponse{Status: status}, safe)
}
func messageQueryResponse(result proto.Message, safe *pb.ErrorInfo) wire.Response {
	if result == nil || !result.ProtoReflect().IsValid() {
		return wire.Response{Err: safe}
	}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}, Err: safe}
}

func messageQueryError(err error) *pb.ErrorInfo {
	var call *wire.CallError
	if errors.As(err, &call) {
		return call.ErrorInfo
	}
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	if errors.Is(err, context.Canceled) {
		code = pb.ErrorCode_ERROR_CODE_CANCELLED
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return &pb.ErrorInfo{Code: code, SafeMessage: err.Error()}
}

func checkDeclaredFormats(description *pb.MessageDescription, chain *pb.MessageChain) *pb.ErrorInfo {
	for _, segment := range chain.GetSegments() {
		var name string
		switch segment.GetData().(type) {
		case *pb.Segment_Text:
			name = "text"
		case *pb.Segment_Markdown:
			name = "markdown"
		case *pb.Segment_Image:
			name = "image"
		case *pb.Segment_Video:
			name = "video"
		case *pb.Segment_Voice:
			name = "voice"
		case *pb.Segment_File:
			name = "file"
		case *pb.Segment_Mention:
			name = "mention"
		case *pb.Segment_Reference:
			name = "quote"
		case *pb.Segment_Forward:
			name = "forward"
		case *pb.Segment_Buttons:
			name = "buttons"
		}
		for _, format := range description.GetFormats() {
			if format.Name == name && format.Status == pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED {
				return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: fmt.Sprintf("adapter instance declares message format %q unsupported", name)}
			}
		}
		for _, item := range segment.GetForward().GetItems() {
			if safe := checkDeclaredFormats(description, item.GetChain()); safe != nil {
				return safe
			}
		}
	}
	return nil
}
