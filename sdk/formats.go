package sdk

import (
	"context"
	"errors"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// DescribeMessage 经已激活 Session 查询目标实例，Via 固定实例。
func DescribeMessage(ctx context.Context, to string, opts ...SendOption) (*message.OutboundDescription, error) {
	target, err := messageTarget(to, opts)
	if err != nil {
		return nil, err
	}
	return callDescribeMessage(ctx, target, nil)
}

// DescribeMessage 沿当前 Handler 的可信父调用查询目标实例。
func (c *Context) DescribeMessage(to string, opts ...SendOption) (*message.OutboundDescription, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	target, err := messageTarget(to, opts)
	if err != nil {
		return nil, err
	}
	return callDescribeMessage(c.ctx, target, c.inbound)
}

// CheckMessage 经已激活 Session 只读预检目标消息，保留状态及错误。
func CheckMessage(ctx context.Context, to string, content message.Content, opts ...SendOption) (message.SupportStatus, error) {
	target, err := messageTarget(to, opts)
	if err != nil {
		return message.SupportUnknown, err
	}
	return callCheckMessage(ctx, target, content, nil)
}

// CheckMessage 沿当前 Handler 的可信父调用只读预检消息，保留原会话回复许可。
func (c *Context) CheckMessage(to string, content message.Content, opts ...SendOption) (message.SupportStatus, error) {
	if !c.active.Load() {
		return message.SupportUnknown, errcode.FailedPrecondition.New("handler context has expired")
	}
	target, err := messageTarget(to, opts)
	if err != nil {
		return message.SupportUnknown, err
	}
	return callCheckMessage(c.ctx, target, content, c.inbound)
}

func messageTarget(to string, opts []SendOption) (*pb.Target, error) {
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	return convert.TargetToProto(to, options.via)
}
func callDescribeMessage(ctx context.Context, target *pb.Target, parent *wire.Inbound) (*message.OutboundDescription, error) {
	result := &pb.MessageDescription{}
	present, err := callMessageQuery(ctx, wire.ServiceDescribeMessage, &pb.DescribeMessageRequest{To: target}, result, parent)
	if !present {
		return nil, err
	}
	return convert.OutboundDescriptionFromProto(result), err
}
func callCheckMessage(ctx context.Context, target *pb.Target, content message.Content, parent *wire.Inbound) (message.SupportStatus, error) {
	chain, err := convert.ContentToProto(content, target)
	if err != nil {
		return message.SupportUnknown, err
	}
	result := &pb.CheckMessageResponse{}
	_, err = callMessageQuery(ctx, wire.ServiceCheckMessage, &pb.CheckMessageRequest{To: target, Chain: chain}, result, parent)
	return convert.SupportStatusFromProto(result.Status), err
}
func callMessageQuery(ctx context.Context, name string, request, result proto.Message, parent *wire.Inbound) (bool, error) {
	session, err := currentSession(ctx)
	if err != nil {
		return false, err
	}
	payload, err := adapterPayload(request)
	if err != nil {
		return false, err
	}
	call := wire.Request{Target: name, Payload: payload}
	if parent != nil {
		call.ParentCallID = parent.CallID
		call.TraceID = parent.TraceID
	}
	response, callError := session.Call(ctx, call)
	var publicError error
	if response.Err != nil {
		publicError = convert.ErrorFromProto(response.Err)
	} else if callError != nil {
		publicError = errcode.Code(wire.CodeOf(callError)).New(callError.Error())
	}
	present := response.Payload.SchemaID != ""
	if present {
		if err := decodeInvocation(response.Payload, result); err != nil {
			return false, errors.Join(publicError, errcode.Internal.New(err.Error()))
		}
	} else if publicError == nil {
		return false, errcode.Internal.New("message query response is missing")
	}
	return present, publicError
}
