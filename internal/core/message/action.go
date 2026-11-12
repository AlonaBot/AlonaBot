package message

import (
	"context"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Action 先按发信规则准入目标，再核对动作声明、范围和参数并执行一次。
// 装配处将其注册为 wire.ServiceAction；JSON 结果与安全错误同时保留。
func (s *Service) Action(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 在协议边界核对载荷与互斥目标字段，动作完整名决定所属 Adapter。
	reject := func(code pb.ErrorCode, message string) wire.Response {
		return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
	}
	request := &pb.ActionRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected ActionRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid ActionRequest protobuf: "+err.Error())
	}
	if !validate.IsAdapterEventType(request.Name) || request.Version == 0 {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid action name or version")
	}
	if request.To != nil && request.AdapterInstanceId != "" {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "to and adapter_instance_id are mutually exclusive")
	}
	adapterID := request.Name[:strings.LastIndexByte(request.Name, '.')]
	invocation, _ := in.ParentAttach.(*router.Invocation)
	var defaultTarget *pb.Target
	if invocation != nil {
		defaultTarget = invocation.DefaultTarget
	}
	// 2. 会话动作沿用完整发信准入，账号动作只检查实例许可。
	var target *pb.Target
	var safe *pb.ErrorInfo
	if request.To != nil {
		target, safe = s.adapters.Complete(request.To, defaultTarget)
		if safe == nil {
			safe = s.policy.CheckSend(ctx, caller.PluginID, invocation, target)
		}
	} else {
		target, safe = s.adapters.CompleteAccount(adapterID, request.AdapterInstanceId)
		if safe == nil {
			safe = s.policy.CheckInstance(ctx, caller.PluginID, invocation, target.GetAdapterInstanceId())
		}
	}
	if safe != nil {
		return wire.Response{Err: safe}
	}
	// 3. 实例启用与账号、Adapter 绑定先于声明检查和任何平台调用。
	instance, ok := s.adapters.Get(target.GetAdapterInstanceId())
	if !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	if adapterID != instance.AdapterID() || target.GetAdapterId() != instance.AdapterID() || target.GetAccountId() != instance.AccountID() {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "action adapter or target account does not match instance")
	}
	executor, ok := instance.(adapter.ActionExecutor)
	if !ok {
		return reject(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support platform actions")
	}
	scope, safe := executor.CheckAction(request.Name, request.Version, request.ParamsJson)
	if safe != nil {
		return wire.Response{Err: safe}
	}
	if (scope == pb.ActionScope_ACTION_SCOPE_CHAT && request.To == nil) || (scope == pb.ActionScope_ACTION_SCOPE_ACCOUNT && request.To != nil) {
		return reject(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "action scope does not match target")
	}
	// 4. 通过全部准入后提交一次，保留结果 JSON 原文及同时返回的错误。
	if scope == pb.ActionScope_ACTION_SCOPE_ACCOUNT {
		target = nil
	}
	result, safe := executor.Action(ctx, target, request.Name, request.Version, request.ParamsJson)
	out := wire.Response{Err: safe}
	if result != "" {
		response := &pb.ActionResponse{ResultJson: result}
		encoded, err := proto.Marshal(response)
		if err != nil {
			return reject(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode ActionResponse: "+err.Error())
		}
		out.Payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(response), SchemaVersion: schemaVersion, Bytes: encoded}
	}
	return out
}
