package remote

import (
	"context"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// ReadNodes 查询绑定账号的节点，递归登记返回媒体后保留部分结果和错误。
func (i *instance) ReadNodes(ctx context.Context, target *pb.Target, segment *pb.ForwardSegment) ([]*pb.ForwardItem, *pb.ErrorInfo) {
	p := i.process
	p.mu.Lock()
	running := i.running
	p.mu.Unlock()
	if !running {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	request, err := payload(&pb.AdapterReadNodesRequest{InstanceId: i.id, InstanceGeneration: i.generation, Target: target, Segment: segment})
	if err != nil {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
	}
	response, callError := i.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterReadNodes, Payload: request})
	result := &pb.ReadNodesResponse{}
	if response.Payload.SchemaID != "" {
		if err := decode(response.Payload, result); err != nil {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}
		}
		for index, item := range result.Items {
			var validation error
			if item.GetChain() == nil {
				validation = fmt.Errorf("node response requires chain")
			} else {
				validation = validate.IncomingChain(item.GetChain())
			}
			if validation != nil {
				detail := fmt.Sprintf("items[%d]: %s", index, validation)
				if response.Err == nil {
					response.Err = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: detail}
				} else {
					response.Err = proto.CloneOf(response.Err)
					response.Err.SafeMessage += "; " + detail
				}
			} else {
				p.registerIncomingChains(i, item.GetChain())
			}
		}
	}
	if response.Err != nil {
		return result.Items, response.Err
	}
	if callError != nil {
		return result.Items, &pb.ErrorInfo{Code: wire.CodeOf(callError), SafeMessage: callError.Error()}
	}
	return result.Items, nil
}
