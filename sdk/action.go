package sdk

import (
	"context"
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
)

// CallAction 关联当前 Handler 的父调用与期限；opts 只接受 Via，账号动作传空会话。
func (c *Context) CallAction(to string, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	return callAction(c.ctx, c.inbound, to, name, version, params, opts...)
}

func callAction(ctx context.Context, parent *wire.Inbound, to string, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error) {
	// 1. 编码参数与目标，动作只接受既有的实例选择选项。
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	request, err := convert.ActionRequestToProto(to, name, version, params, options.via)
	if err != nil {
		return nil, err
	}
	request.UseDefaultTarget = options.actionDefaultTarget
	session, err := currentSession(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := adapterPayload(request)
	if err != nil {
		return nil, errcode.InvalidArgument.New("encode ActionRequest: " + err.Error())
	}
	// 2. Handler 保留父调用；后台只沿用任务 context 的取消和期限。
	call := wire.Request{Target: wire.ServiceAction, Payload: payload}
	if parent != nil {
		call.ParentCallID = parent.CallID
		call.TraceID = parent.TraceID
	}
	response, callError := session.Call(ctx, call)
	// 3. 先解码有效 JSON 结果，再保留同时返回的安全错误。
	var result json.RawMessage
	if response.Payload.SchemaID != "" {
		decoded := &pb.ActionResponse{}
		if err := decodeInvocation(response.Payload, decoded); err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		if decoded.ResultJson != "" {
			result = json.RawMessage(decoded.ResultJson)
		}
	}
	if response.Err != nil {
		return result, convert.ErrorFromProto(response.Err)
	}
	if callError != nil {
		return result, errcode.Code(wire.CodeOf(callError)).New(callError.Error())
	}
	if len(result) == 0 {
		return nil, errcode.Internal.New("missing action result")
	}
	return result, nil
}
