package sdk

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// ReadNodes 沿当前 Handler 的可信父调用查询当前会话中的节点。
func (c *Context) ReadNodes(segment message.ForwardSegment) ([]message.ForwardItem, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	if c.target() == nil {
		return nil, errcode.NoDefaultTarget.New("no default target")
	}
	return callReadNodes(c.ctx, nil, c.target(), segment, c.inbound)
}

// ReadNodes 经已激活 Session 查询指定会话，Via 选择实例。
func ReadNodes(ctx context.Context, sid string, segment message.ForwardSegment, opts ...SendOption) ([]message.ForwardItem, error) {
	target, err := messageTarget(sid, opts)
	if err != nil {
		return nil, err
	}
	return callReadNodes(ctx, target, target, segment, nil)
}
func callReadNodes(ctx context.Context, to, target *pb.Target, segment message.ForwardSegment, parent *wire.Inbound) ([]message.ForwardItem, error) {
	if segment.PlatformForwardID != "" {
		segment.Items = nil // 已展开结果再次查询时只提交平台资源标识。
	}
	chain, err := convert.ChainToProto(message.Chain{segment}, target)
	if err != nil {
		return nil, err
	}
	request := &pb.ReadNodesRequest{To: to, Segment: chain.Segments[0].GetForward()}
	if err := validate.ReadNodes(request.Segment); err != nil {
		return nil, errcode.InvalidArgument.New(err.Error())
	}
	session, err := currentSession(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := adapterPayload(request)
	if err != nil {
		return nil, err
	}
	call := wire.Request{Target: wire.ServiceReadNodes, Payload: payload}
	if parent != nil {
		call.ParentCallID = parent.CallID
		call.TraceID = parent.TraceID
	}
	response, callError := session.Call(ctx, call)
	var nodes []message.ForwardItem
	if response.Payload.SchemaID != "" {
		decoded := &pb.ReadNodesResponse{}
		if err := decodeInvocation(response.Payload, decoded); err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		nodes = convert.ChainFromProto(&pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Forward{Forward: &pb.ForwardSegment{Items: decoded.Items}}}}}, target)[0].(message.ForwardSegment).Items
	}
	if response.Err != nil {
		return nodes, convert.ErrorFromProto(response.Err)
	}
	if callError != nil {
		return nodes, errcode.Code(wire.CodeOf(callError)).New(callError.Error())
	}
	return nodes, nil
}
