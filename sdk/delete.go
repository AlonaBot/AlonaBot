package sdk

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// Delete 撤回同一会话和实例的消息，关联当前 Handler 的父调用与期限。
func (c *Context) Delete(refs ...message.MessageRef) (*message.DeleteResult, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	return callDelete(c.ctx, refs, c.inbound)
}

// Delete 通过已激活的 Session 独立撤回句柄所指的消息。
func Delete(ctx context.Context, refs ...message.MessageRef) (*message.DeleteResult, error) {
	return callDelete(ctx, refs, nil)
}

func callDelete(ctx context.Context, refs []message.MessageRef, parent *wire.Inbound) (*message.DeleteResult, error) {
	// 1. 在取得 Session 和发送 RPC 前校验全部句柄。
	request, err := convert.DeleteRequestToProto(refs)
	if err != nil {
		return nil, err
	}
	session, err := currentSession()
	if err != nil {
		return nil, err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, errcode.InvalidArgument.New("encode DeleteRequest: " + err.Error())
	}
	call := wire.Request{Target: wire.ServiceDelete, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}}
	if parent != nil {
		call.ParentCallID = parent.CallID
		call.TraceID = parent.TraceID
	}
	response, callError := session.Call(ctx, call)
	// 2. 先保留有效逐项结果，再返回同时出现的安全错误。
	var result *message.DeleteResult
	if response.Payload.SchemaID != "" {
		decoded := &pb.DeleteResult{}
		if err := decodeInvocation(response.Payload, decoded); err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		result = convert.DeleteResultFromProto(decoded)
	}
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		return result, &Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
	}
	return result, nil
}
