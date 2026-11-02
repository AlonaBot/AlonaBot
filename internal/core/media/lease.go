package media

import (
	"context"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"google.golang.org/protobuf/proto"
)

// mediaLease 在 Store.mu 下管理只读令牌、文件所有权和结束回调。
// 调用者释放、context 取消、绝对期限与 Sweep 可以竞争结束；只有第一次释放生效。
type mediaLease struct {
	token       string
	upload      *entry
	incoming    *incomingObject
	released    bool
	timer       *time.Timer
	stopContext func() bool
}

// LeaseForSend 为目标 Adapter Session 签发只读令牌，保护对象到发送结束或 deadline。
// 入站对象先检查原令牌的目标范围再取件；release 可重复调用，取消 ctx 也会释放租约。
func (s *Store) LeaseForSend(ctx context.Context, c router.Caller, readerSessionKey string, target *pb.Target, source *pb.MediaSource, deadline time.Time) (readToken string, release func(), err error) {
	if err := ctx.Err(); err != nil {
		return "", nil, err
	}
	if parent, ok := ctx.Deadline(); ok && parent.Before(deadline) {
		deadline = parent
	}
	if !deadline.After(s.now()) {
		return "", nil, mediaCallError(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "media send deadline has expired")
	}
	// 1. 区分来源，入站对象先核对完整目标范围，再按原读取者的预算取件。
	var original string
	incoming := false
	switch value := source.GetSource().(type) {
	case *pb.MediaSource_MediaRef:
		original = value.MediaRef
	case *pb.MediaSource_IncomingRef:
		original = value.IncomingRef.GetToken()
		incoming = true
	default:
		return "", nil, mediaCallError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected media_ref or incoming_ref source")
	}
	if incoming {
		s.mu.Lock()
		s.sweep()
		access, err := s.incomingAccess(original, c.SessionKey)
		if err != nil {
			s.mu.Unlock()
			return "", nil, err
		}
		allowed := false
		for _, scoped := range access.targets {
			if proto.Equal(scoped, target) {
				allowed = true
				break
			}
		}
		s.mu.Unlock()
		if !allowed {
			return "", nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "media token does not allow target")
		}
		fetchCtx, cancel := context.WithTimeout(ctx, deadline.Sub(s.now()))
		err = s.loadIncoming(fetchCtx, c, original)
		cancel()
		if err != nil {
			return "", nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return "", nil, err
	}
	if !deadline.After(s.now()) {
		return "", nil, mediaCallError(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "media send deadline has expired")
	}
	// 2. 在同一回收锁下重新核对原 token，并向目标 Session 绑定新的只读引用。
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	lease := &mediaLease{token: token()}
	var file File
	if incoming {
		access, err := s.incomingAccess(original, c.SessionKey)
		if err != nil {
			return "", nil, err
		}
		lease.incoming = access.object
		lease.incoming.leases++
		file = access.object.file
	} else {
		upload := s.refs[original]
		if upload == nil || upload.owner != c.SessionKey {
			return "", nil, mediaCallError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "media reference not found")
		}
		if upload.lease != nil {
			return "", nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "read-only media token cannot be delegated")
		}
		lease.upload = upload
		upload.leases++
		file = upload.file
	}
	file.ExpiresAt = deadline
	s.refs[lease.token] = &entry{file: file, owner: readerSessionKey, lease: lease}
	// 3. 回调只持有此租约；结束路径停止另一个回调，避免回收后的长期引用。
	release = func() { s.mu.Lock(); defer s.mu.Unlock(); s.releaseLease(lease) }
	lease.timer = time.AfterFunc(deadline.Sub(s.now()), release)
	lease.stopContext = context.AfterFunc(ctx, release)
	return lease.token, release, nil
}

// releaseLease 要求持有 Store.mu；只读引用不再占一份磁盘或 Session 预算。
func (s *Store) releaseLease(lease *mediaLease) {
	if lease.released {
		return
	}
	lease.released = true
	delete(s.refs, lease.token)
	lease.timer.Stop()
	lease.stopContext()
	if lease.upload != nil {
		lease.upload.leases--
		if lease.upload.retired && lease.upload.leases == 0 {
			s.release(nil, "", lease.upload)
		}
	} else {
		lease.incoming.leases--
		s.sweepIncoming()
	}
}
