// Package message 实现 Core 消息发送服务的目标准入与回执转发。
//
// [Service] 根据可信父调用的原目标或持久化的实例许可准入，先校验载荷、目标和消息链，
// 再由注册的 Adapter 实例执行发送；结果与错误同时保留（CONTRACTS §8）。
package message

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const schemaVersion = 1

// Service 按插件发信许可准入目标，再由注册的 Adapter 实例执行发送。
type Service struct {
	adapters *adapter.Registry
	policy   *policy.Policy
}

// New 创建消息发送服务；装配处将 Send 注册为 wire.ServiceSend。
func New(adapters *adapter.Registry, p *policy.Policy) *Service {
	return &Service{adapters: adapters, policy: p}
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
	// 2. 显式目标优先，否则使用可信父调用的默认目标，再检查发送许可。
	target := request.To
	if target == nil {
		target = defaultTarget
	}
	if target == nil {
		return reject(nil, pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET, "no default target")
	}
	if sendError := s.policy.CheckSend(ctx, caller.PluginID, invocation, target); sendError != nil {
		return response(&pb.SendResult{Status: pb.SendStatus_SEND_STATUS_REJECTED, Target: target}, sendError)
	}
	// 3. 复用结构校验，空链和仅元数据均在产生副作用前拒绝。
	if err := validate.Chain(request.Chain); err != nil {
		return reject(target, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	// 4. 校验实例绑定后执行唯一一次发送，完整转发回执与错误。
	instance, ok := s.adapters.Get(target.GetAdapterInstanceId())
	if !ok {
		return reject(target, pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return reject(target, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "target adapter or account does not match instance")
	}
	result, sendError := instance.Send(ctx, target, request.Chain)
	return response(result, sendError)
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
