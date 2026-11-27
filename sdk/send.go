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

// Send 通过已激活的 Session 独立发送到指定会话，opts 接受 Via 和 RequireOriginalSegmentIndexes。
func Send(ctx context.Context, sid string, m message.Content, opts ...SendOption) (*message.SendResult, error) {
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	target, err := convert.TargetToProto(sid, options.via)
	if err != nil {
		return nil, err
	}
	return callSend(ctx, target, m, nil, nil, options.requireOriginalSegmentIndexes)
}

// SendText 独立发送一段文本，并丢弃发送回执。
func SendText(ctx context.Context, sid, text string, opts ...SendOption) error {
	_, err := Send(ctx, sid, message.Text(text), opts...)
	return err
}
func callSend(ctx context.Context, to *pb.Target, m message.Content, parent *wire.Inbound, defaultTarget *pb.Target, requireOriginalSegmentIndexes bool) (*message.SendResult, error) {
	target := to
	if target == nil {
		target = defaultTarget
	}
	chain, err := convert.ContentToProto(m, target)
	if err != nil {
		return nil, err
	}
	request := &pb.SendRequest{To: to, Chain: chain, RequireOriginalSegmentIndexes: requireOriginalSegmentIndexes}
	return callMessage(ctx, wire.ServiceSend, request, parent)
}
func callMessage(ctx context.Context, target string, request proto.Message, parent *wire.Inbound) (*message.SendResult, error) {
	// 1. 取得已就绪的 Session，再发出已经校验过的消息请求。
	session, err := currentSession(ctx)
	if err != nil {
		return nil, err
	}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, errcode.InvalidArgument.New("encode " + string(request.ProtoReflect().Descriptor().Name()) + ": " + err.Error())
	}
	call := wire.Request{Target: target, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}}
	if parent != nil {
		call.ParentCallID = parent.CallID
		call.TraceID = parent.TraceID
	}
	response, callError := session.Call(ctx, call)
	// 2. 错误与部分成功可以同时出现，先保留有效回执。
	var result *message.SendResult
	if response.Payload.SchemaID != "" {
		decoded := &pb.SendResult{}
		if err := decodeInvocation(response.Payload, decoded); err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		result = convert.SendResultFromProto(decoded)
	}
	if callError != nil {
		info := callError.(*wire.CallError).ErrorInfo
		return result, &Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
	}
	return result, nil
}
