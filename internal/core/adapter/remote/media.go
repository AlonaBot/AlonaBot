package remote

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const mediaReadCapability = "media.read"

// mediaInstance 只包装明确声明读取能力的实例，其他可选接口沿用 instance。
type mediaInstance struct{ *instance }

var (
	_ adapter.MediaReader  = (*mediaInstance)(nil)
	_ adapter.MediaSession = (*instance)(nil)
)

func eventChains(ev *pb.Event) []*pb.MessageChain {
	payload := ev.GetEventPayload()
	return []*pb.MessageChain{ev.GetMessage().GetChain(), payload.GetMessageCreated().GetChain(), payload.GetMessageEdited().GetChain(), payload.GetMessageEdited().GetOldChain()}
}

func incomingMedia(chains ...*pb.MessageChain) []*pb.IncomingMedia {
	seen := make(map[*pb.IncomingMedia]bool)
	var result []*pb.IncomingMedia
	var rewrite func(*pb.MessageChain)
	rewrite = func(chain *pb.MessageChain) {
		for _, segment := range chain.GetSegments() {
			var source *pb.MediaSource
			switch value := segment.GetData().(type) {
			case *pb.Segment_Image:
				source = value.Image.GetSource()
			case *pb.Segment_Video:
				source = value.Video.GetSource()
			case *pb.Segment_Voice:
				source = value.Voice.GetSource()
			case *pb.Segment_File:
				source = value.File.GetSource()
			}
			if incoming := source.GetIncomingRef(); incoming != nil && !seen[incoming] {
				// 同一描述符可出现在事件的多条消息链中，避免重复改写已登记的对象 ID。
				seen[incoming] = true
				result = append(result, incoming)
			}
			for _, item := range segment.GetForward().GetItems() {
				rewrite(item.GetChain())
			}
		}
	}
	for _, chain := range chains {
		rewrite(chain)
	}
	return result
}

// registerIncomingChains 仅从已验证的 Adapter 事件或节点响应登记平台文件。
func (p *Process) registerIncomingChains(entry *instance, chains ...*pb.MessageChain) {
	for _, incoming := range incomingMedia(chains...) {
		incoming.Token = p.opt.RegisterIncoming(entry.registered, incoming)
	}
}

// MediaSessionKey 返回实例创建时绑定的 Adapter Session 身份。
func (entry *instance) MediaSessionKey() string { return entry.sessionKey }

// ReadMedia 经实例创建时的 Session 按需读取平台文件，不自动重试。
func (entry *mediaInstance) ReadMedia(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	p := entry.process
	p.mu.Lock()
	running := entry.running
	p.mu.Unlock()
	if !running {
		return nil, 0, false, failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	request, err := payload(&pb.AdapterMediaReadRequest{InstanceId: entry.id, InstanceGeneration: entry.generation, FileId: fileID, Offset: offset, Length: length})
	if err != nil {
		return nil, 0, false, err
	}
	response, callErr := entry.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterMediaRead, Payload: request})
	result := &pb.AdapterMediaReadResponse{}
	if response.Payload.SchemaID != "" {
		if err := decode(response.Payload, result); err != nil {
			return nil, 0, false, err
		}
	}
	return result.Data, result.NextOffset, result.Eof, callErr
}
