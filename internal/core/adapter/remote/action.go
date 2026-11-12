package remote

import (
	"context"
	"encoding/json"
	"log/slog"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

var _ adapter.ActionExecutor = (*instance)(nil)

func (i *instance) actionDeclaration(name string, version uint32) (declaredAction, bool) {
	p := i.process
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.session == nil || p.session.Session != i.session {
		return declaredAction{}, false
	}
	action, ok := p.declarations.actions[actionKey{name, version}]
	return action, ok
}

// CheckAction 用绑定 Session 的已解析声明核对动作版本与参数。
func (i *instance) CheckAction(name string, version uint32, paramsJSON string) (pb.ActionScope, *pb.ErrorInfo) {
	action, declared := i.actionDeclaration(name, version)
	if !declared {
		return pb.ActionScope_ACTION_SCOPE_UNSPECIFIED, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter action is not declared"}
	}
	var value any
	if err := json.Unmarshal([]byte(paramsJSON), &value); err != nil {
		return action.scope, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "invalid action params JSON"}
	}
	if err := action.params.Validate(value); err != nil {
		return action.scope, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "action params do not match schema"}
	}
	return action.scope, nil
}

// Action 经创建实例时的 Session 调用一次，并按注册的结果 Schema 校验 JSON 原文。
func (i *instance) Action(ctx context.Context, target *pb.Target, name string, version uint32, paramsJSON string) (string, *pb.ErrorInfo) {
	// 1. 快照实例状态和声明，旧实例引用始终保留自己的 Session。
	p := i.process
	p.mu.Lock()
	running := i.running
	p.mu.Unlock()
	if !running {
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	action, declared := i.actionDeclaration(name, version)
	if !declared {
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter action is not declared"}
	}
	// 2. 账号动作省略目标，Core 补齐的账号只参与准入检查。
	if action.scope == pb.ActionScope_ACTION_SCOPE_ACCOUNT {
		target = nil
	}
	request, err := payload(&pb.AdapterActionRequest{InstanceId: i.id, Target: target, Name: name, Version: version, ParamsJson: paramsJSON})
	if err != nil {
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	response, callError := i.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterAction, Payload: request})
	// 3. 有效结果与安全错误同时保留，诊断只记录身份与动作名。
	invalidResult := func() (string, *pb.ErrorInfo) {
		slog.Error("adapter action result schema validation failed", "plugin_id", p.opt.PluginID, "instance_id", i.id, "action", name, "version", version)
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "invalid adapter action result"}
	}
	var result string
	if response.Payload.SchemaID != "" {
		decoded := &pb.ActionResponse{}
		if err := decode(response.Payload, decoded); err != nil {
			return invalidResult()
		}
		result = decoded.ResultJson
	}
	if result != "" {
		var value any
		if err := json.Unmarshal([]byte(result), &value); err != nil {
			return invalidResult()
		}
		if err := action.result.Validate(value); err != nil {
			return invalidResult()
		}
	} else if response.Err == nil && callError == nil {
		return invalidResult()
	}
	if response.Err != nil {
		return result, response.Err
	}
	if callError != nil {
		return result, &pb.ErrorInfo{Code: wire.CodeOf(callError), SafeMessage: callError.Error()}
	}
	return result, nil
}
