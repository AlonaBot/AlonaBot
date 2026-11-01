package remote

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

var (
	_ adapter.MediaReader  = (*instance)(nil)
	_ adapter.MediaSession = (*instance)(nil)
)

// MediaSessionKey 返回实例创建时绑定的 Adapter Session 身份。
func (entry *instance) MediaSessionKey() string { return entry.sessionKey }

// ReadMedia 经实例创建时的 Session 按需读取平台文件，不自动重试。
func (entry *instance) ReadMedia(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
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
