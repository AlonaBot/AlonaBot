// Package message 实现 Core 消息发送与编辑服务的目标准入和回执转发。
//
// [Service] 根据可信父调用的原目标或持久化的实例许可准入，先校验载荷、目标和消息链，
// 再准备媒体消费租约并由注册的 Adapter 实例执行发送；结果与错误同时保留（CONTRACTS §8）。
// 编辑仅接受文本及可选按钮，实例须声明能力并实现 adapter.Editor。
package message

import (
	"context"
	"errors"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const schemaVersion = 1

// Service 按插件发信许可准入目标，再由注册的 Adapter 实例执行发送或编辑。
type Service struct {
	requestTimeout time.Duration
	adapters       *adapter.Registry
	policy         *policy.Policy
	lease          func(context.Context, router.Caller, string, *pb.Target, *pb.MediaSource, time.Time) (string, func(), error)
}

// New 创建消息服务；装配处将 Send、Edit 分别注册为 wire.ServiceSend、wire.ServiceEdit。
// lease 由媒体 Store 提供，签发绑定目标 Adapter Session 的只读令牌；
// requestTimeout 使用生效配置，在调用没有期限时提供媒体请求预算。
func New(adapters *adapter.Registry, p *policy.Policy, lease func(ctx context.Context, c router.Caller, readerSessionKey string, target *pb.Target, source *pb.MediaSource, deadline time.Time) (string, func(), error), requestTimeout time.Duration) *Service {
	return &Service{adapters: adapters, policy: p, lease: lease, requestTimeout: requestTimeout}
}

// Send 按载荷、可信目标、消息链和实例绑定依次准入，不自动重试平台发送。
// 拒绝时同时返回 REJECTED 回执与错误，发送后保留 Adapter 返回的二者。
func (s *Service) Send(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 取得已知的可信目标供拒绝回执使用，再校验协议载荷。
	invocation, _ := in.ParentAttach.(*router.Invocation)
	var defaultTarget *pb.Target
	if invocation != nil {
		defaultTarget = invocation.DefaultTarget
	}
	request := &pb.SendRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(defaultTarget, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected SendRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(defaultTarget, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid SendRequest protobuf: "+err.Error())
	}
	// 2. 显式目标先补齐实例与账号，否则沿用可信父调用的默认目标，再检查发送许可。
	target := request.To
	if target == nil {
		target = defaultTarget
	} else {
		completed, completionError := s.adapters.Complete(target, defaultTarget)
		if completionError != nil {
			return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, completionError)
		}
		target = completed
	}
	if target == nil {
		return reject(nil, pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, "no default target")
	}
	if sendError := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); sendError != nil {
		// 准入拒绝只回传调用方已知的目标，避免带出补齐的实例身份。
		rejectedTarget := request.To
		if rejectedTarget == nil {
			rejectedTarget = defaultTarget
		}
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: rejectedTarget}, sendError)
	}
	// 3. 复用结构校验，空链和仅元数据均在产生副作用前拒绝。
	if completionError := s.completeReferences(request.Chain, target); completionError != nil {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, completionError)
	}
	if err := validate.Chain(request.Chain); err != nil {
		return reject(target, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	// 4. 校验最终实例绑定，媒体来源准备也使用这一实例的 Session。
	instance, ok := s.adapters.Get(target.GetAdapterInstanceId())
	if !ok {
		return reject(target, pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return reject(target, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "target adapter or account does not match instance")
	}
	// 5. 全部来源先取得只读租约；任一失败或发送结束均释放本次已得租约。
	sources := mediaSources(request.Chain)
	if len(sources) > 0 {
		reader, ok := instance.(adapter.MediaSession)
		if !ok {
			return reject(target, pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support media tokens")
		}
		deadline := in.Deadline
		if parent, ok := ctx.Deadline(); ok && (deadline.IsZero() || parent.Before(deadline)) {
			deadline = parent
		}
		if deadline.IsZero() {
			// Core 自动回复直接调用 Send；没有 RPC 期限时使用契约的普通请求预算。
			deadline = time.Now().Add(s.requestTimeout)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
		for _, source := range sources {
			token, release, err := s.lease(ctx, caller, reader.MediaSessionKey(), target, source, deadline)
			if err != nil {
				return rejectError(target, err)
			}
			defer release()
			source.Source = &pb.MediaSource_MediaRef{MediaRef: token}
		}
		if err := ctx.Err(); err != nil {
			return rejectError(target, err)
		}
	}
	// 6. 所有媒体已就绪后执行唯一一次发送，完整转发回执与错误。
	result, sendError := instance.Send(ctx, target, request.Chain)
	return response(result, sendError)
}

func (s *Service) completeReferences(chain *pb.MessageChain, target *pb.Target) *pb.ErrorInfo {
	for _, segment := range chain.GetSegments() {
		if reference := segment.GetReference(); reference.GetTarget() != nil {
			completed, completionError := s.adapters.Complete(reference.Target, target)
			if completionError != nil {
				return completionError
			}
			reference.Target = completed
		}
		for _, item := range segment.GetForward().GetItems() {
			if completionError := s.completeReferences(item.GetChain(), target); completionError != nil {
				return completionError
			}
		}
	}
	return nil
}

// mediaSources 按消息顺序收集所有需租约的来源，HTTPS 保留给 Adapter 处理。
func mediaSources(chain *pb.MessageChain) []*pb.MediaSource {
	var sources []*pb.MediaSource
	for _, segment := range chain.GetSegments() {
		var source *pb.MediaSource
		switch value := segment.GetData().(type) {
		case *pb.Segment_Image:
			source = value.Image.GetSource()
		case *pb.Segment_Video:
			source = value.Video.GetSource()
		case *pb.Segment_File:
			source = value.File.GetSource()
		}
		switch source.GetSource().(type) {
		case *pb.MediaSource_MediaRef, *pb.MediaSource_IncomingRef:
			sources = append(sources, source)
		}
		for _, item := range segment.GetForward().GetItems() {
			sources = append(sources, mediaSources(item.GetChain())...)
		}
	}
	return sources
}

func rejectError(target *pb.Target, err error) wire.Response {
	var call *wire.CallError
	if errors.As(err, &call) {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, call.ErrorInfo)
	}
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	if errors.Is(err, context.Canceled) {
		code = pb.ErrorCode_ERROR_CODE_CANCELLED
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return reject(target, code, err.Error())
}

func reject(target *pb.Target, code pb.ErrorCode, message string) wire.Response {
	return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, &pb.ErrorInfo{Code: code, SafeMessage: message})
}

func response(result *pb.SendResult, sendError *pb.ErrorInfo) wire.Response {
	out := wire.Response{Err: sendError}
	if result == nil {
		return out
	}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "encode SendResult: " + err.Error()}}
	}
	out.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}
	return out
}
