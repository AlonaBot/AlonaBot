package convert

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const maxDeleteMessages = 100

// DeleteRequestToProto 校验同一会话与实例的句柄，保留输入顺序、重复 ID 及省略的实例。
func DeleteRequestToProto(refs []string) (*pb.DeleteRequest, error) {
	if len(refs) == 0 || len(refs) > maxDeleteMessages {
		return nil, errcode.InvalidArgument.New("delete requires 1 to 100 message references")
	}
	request := &pb.DeleteRequest{}
	var session string
	var instance string
	for i, ref := range refs {
		target, id, err := RefToProto(ref)
		if err != nil {
			return nil, err
		}
		current := SessionFromProto(target)
		if i == 0 {
			request.To, session, instance = target, current, target.AdapterInstanceId
		} else if current != session || target.AdapterInstanceId != instance {
			return nil, errcode.InvalidArgument.New("delete references must share the same session and adapter instance")
		}
		request.PlatformMessageIds = append(request.PlatformMessageIds, id)
	}
	return request, nil
}

// DeleteResultFromProto 转换逐项状态与错误，保留顺序和实际目标。
func DeleteResultFromProto(in *pb.DeleteResult) *message.DeleteResult {
	out := &message.DeleteResult{Session: SessionFromProto(in.Target), AdapterInstanceID: in.GetTarget().GetAdapterInstanceId()}
	for _, item := range in.Items {
		out.Items = append(out.Items, message.DeleteItemResult{PlatformMessageID: item.GetPlatformMessageId(), Status: message.PartStatus(item.GetStatus()), Error: ErrorFromProto(item.GetError())})
	}
	return out
}

// DeleteResultToProto 转换 Adapter 回执，目标身份取自 Core 已校验的请求。
func DeleteResultToProto(in *message.DeleteResult, target *pb.Target) *pb.DeleteResult {
	if in == nil {
		return nil
	}
	out := &pb.DeleteResult{Target: target}
	for _, item := range in.Items {
		out.Items = append(out.Items, &pb.DeleteItemResult{PlatformMessageId: item.PlatformMessageID, Status: pb.PartStatus(item.Status), Error: ErrorToProto(item.Error)})
	}
	return out
}
