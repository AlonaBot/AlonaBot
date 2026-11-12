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

// registerIncoming 将事件消息链中的平台文件标识改写为 Core 对象 ID。
func (p *Process) registerIncoming(instanceID string, ev *pb.Event) {
	seen := make(map[*pb.IncomingMedia]bool)
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
				incoming.Token = p.opt.RegisterIncoming(instanceID, incoming.Token, incoming.Size, incoming.Mime)
			}
			for _, item := range segment.GetForward().GetItems() {
				rewrite(item.GetChain())
			}
		}
	}
	rewrite(ev.GetMessage().GetChain())
	payload := ev.GetEventPayload()
	rewrite(payload.GetMessageCreated().GetChain())
	rewrite(payload.GetMessageEdited().GetChain())
	rewrite(payload.GetMessageEdited().GetOldChain())
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
	request, err := payload(&pb.AdapterMediaReadRequest{InstanceId: entry.id, FileId: fileID, Offset: offset, Length: length})
	if err != nil {
		return nil, 0, false, err
	}
	response, err := entry.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterMediaRead, Payload: request})
	if err != nil {
		return nil, 0, false, err
	}
	result := &pb.AdapterMediaReadResponse{}
	if err := decode(response.Payload, result); err != nil {
		return nil, 0, false, err
	}
	return result.Data, result.NextOffset, result.Eof, nil
}
