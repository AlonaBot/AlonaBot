package sdk

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func (r *adapterRuntime) checkMessage(ctx context.Context, in *wire.Inbound) wire.Response {
	request := &pb.AdapterSendRequest{}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	r.mu.Lock()
	entry, safe := r.operationEntry(request.InstanceId, request.InstanceGeneration)
	if safe != nil {
		r.mu.Unlock()
		return wire.Response{Err: safe}
	}
	checker, supported := entry.instance.(adapter.MessageChecker)
	r.mu.Unlock()
	result := &pb.CheckMessageResponse{}
	if supported {
		status, safe := checker.CheckMessage(ctx, convert.SessionFromProto(request.Target), convert.ChainFromProto(request.Chain, request.Target))
		if status != message.SupportUnknown && status != message.SupportSupported && status != message.SupportUnsupported {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter check status")
		}
		result.Status = convert.SupportStatusToProto(status)
		return adapterResponse(result, convert.ErrorToProto(safe))
	}
	return adapterResponse(result, nil)
}
