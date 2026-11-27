package media

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"google.golang.org/protobuf/proto"
)

// PrepareCheckSource 核对来源 Session、期限和完整目标授权，返回原来源的独立副本及可信 MIME。
// 只读取受管元数据；不取件、打开文件、签发令牌或增加持有及预算。
func (s *Store) PrepareCheckSource(ctx context.Context, c router.Caller, target *pb.Target, source *pb.MediaSource) (*pb.MediaSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	prepared := proto.CloneOf(source)
	switch value := source.GetSource().(type) {
	case *pb.MediaSource_MediaRef:
		upload, err := s.sendUpload(value.MediaRef, c, target)
		if err != nil {
			return nil, err
		}
		prepared.Mime = upload.file.MIME
	case *pb.MediaSource_IncomingRef:
		access, err := s.sendIncoming(value.IncomingRef.GetToken(), c, target)
		if err != nil {
			return nil, err
		}
		prepared.Mime = ""
		if access.object.mime != nil {
			prepared.Mime = *access.object.mime
		}
	case *pb.MediaSource_HttpsUrl, *pb.MediaSource_InlineBase64:
		prepared.Mime = ""
	default:
		return nil, mediaCallError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid media source")
	}
	return prepared, nil
}

// sendIncoming 要求持有 Store.mu，查询与租约共用完整目标范围授权。
func (s *Store) sendIncoming(id string, c router.Caller, target *pb.Target) (*incomingAccess, error) {
	access, err := s.incomingAccess(id, c.SessionKey)
	if err != nil {
		return nil, err
	}
	if !allowsTarget(access.targets, target) {
		return nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "media token does not allow target")
	}
	return access, nil
}

// sendUpload 要求持有 Store.mu，保留上传所有者、只读租约与 Grant 的既有权利。
func (s *Store) sendUpload(id string, c router.Caller, target *pb.Target) (*entry, error) {
	upload := s.refs[id]
	if upload == nil || upload.owner != c.SessionKey || !s.now().Before(upload.file.ExpiresAt) {
		return nil, mediaCallError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "media reference not found")
	}
	if upload.lease != nil {
		return nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "read-only media token cannot be delegated")
	}
	if upload.grant != nil {
		if !allowsTarget(upload.targets, target) {
			return nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "media token does not allow target")
		}
		upload = upload.grant
	}
	return upload, nil
}
