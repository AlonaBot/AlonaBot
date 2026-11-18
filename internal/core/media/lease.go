package media

import (
	"context"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
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
	urls        []string // 与此消费租约一起撤销的临时 URL，由 Store.mu 保护
	activity    func()
	caller      router.Caller // 媒体来源 Session，临时 URL 读者继续为它保留引用。
}

// LeaseForSend 为目标 Adapter Session 准备媒体来源，保护对象到发送结束或 deadline。
// 入站对象先检查完整目标范围；同实例复用平台标识，跨实例按需取件。
// release 可重复调用，取消 ctx 也会释放租约。
func (s *Store) LeaseForSend(ctx context.Context, c router.Caller, readerSessionKey string, target *pb.Target, source *pb.MediaSource, deadline time.Time) (prepared *pb.MediaSource, release func(), err error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if parent, ok := ctx.Deadline(); ok && parent.Before(deadline) {
		deadline = parent
	}
	if !deadline.After(s.now()) {
		return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "media send deadline has expired")
	}
	// 1. 区分来源，入站对象先核对完整目标范围，再决定是否取件。
	var original string
	incoming := false
	switch value := source.GetSource().(type) {
	case *pb.MediaSource_MediaRef:
		original = value.MediaRef
	case *pb.MediaSource_IncomingRef:
		original = value.IncomingRef.GetToken()
		incoming = true
	case *pb.MediaSource_HttpsUrl, *pb.MediaSource_InlineBase64:
		return source, func() {}, nil
	default:
		return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid media source")
	}
	activity, err := s.acquire(c)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			activity()
		}
	}()
	if incoming {
		s.mu.Lock()
		s.sweep()
		access, err := s.incomingAccess(original, c.SessionKey)
		if err != nil {
			s.mu.Unlock()
			return nil, nil, err
		}
		if !allowsTarget(access.targets, target) {
			s.mu.Unlock()
			return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "media token does not allow target")
		}
		if access.object.source.instanceID == target.GetAdapterInstanceId() {
			lease := &mediaLease{token: token(), incoming: access.object, activity: activity, caller: c}
			access.object.leases++
			release := s.bindLease(ctx, lease, readerSessionKey, File{ExpiresAt: deadline})
			prepared := &pb.MediaSource{Source: &pb.MediaSource_PlatformFile{PlatformFile: access.object.source.fileID}}
			s.mu.Unlock()
			return prepared, release, nil
		}
		s.mu.Unlock()
		fetchCtx, cancel := context.WithTimeout(ctx, deadline.Sub(s.now()))
		err = s.loadIncoming(fetchCtx, c, original)
		cancel()
		if err != nil {
			return nil, nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if !deadline.After(s.now()) {
		return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED, "media send deadline has expired")
	}
	// 2. 在同一回收锁下重新核对原 token，并向目标 Session 绑定新的只读引用。
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	lease := &mediaLease{token: token(), activity: activity, caller: c}
	var file File
	if incoming {
		access, err := s.incomingAccess(original, c.SessionKey)
		if err != nil {
			return nil, nil, err
		}
		lease.incoming = access.object
		lease.incoming.leases++
		file = access.object.file
	} else {
		upload := s.refs[original]
		if upload == nil || upload.owner != c.SessionKey {
			return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "media reference not found")
		}
		if upload.lease != nil {
			return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "read-only media token cannot be delegated")
		}
		if upload.grant != nil {
			if !allowsTarget(upload.targets, target) {
				return nil, nil, mediaCallError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "media token does not allow target")
			}
			upload = upload.grant
		}
		lease.upload = upload
		upload.leases++
		file = upload.file
	}
	file.ExpiresAt = deadline
	release = s.bindLease(ctx, lease, readerSessionKey, file)
	return &pb.MediaSource{Source: &pb.MediaSource_MediaRef{MediaRef: lease.token}}, release, nil
}

// bindLease 要求持有 Store.mu；先发布 Session 索引，再安装结束回调。
func (s *Store) bindLease(ctx context.Context, lease *mediaLease, readerSessionKey string, file File) func() {
	s.refs[lease.token] = &entry{file: file, owner: readerSessionKey, lease: lease}
	// 回调争用同一回收锁，确保发布完成后才释放对象。
	release := func() { s.mu.Lock(); defer s.mu.Unlock(); s.releaseLease(lease) }
	lease.timer = time.AfterFunc(file.ExpiresAt.Sub(s.now()), release)
	lease.stopContext = context.AfterFunc(ctx, release)
	return release
}

// releaseLease 要求持有 Store.mu；只读引用不再占一份磁盘或 Session 预算。
func (s *Store) releaseLease(lease *mediaLease) {
	if lease.released {
		return
	}
	lease.released = true
	delete(s.refs, lease.token)
	for _, id := range lease.urls {
		delete(s.urls, id)
	}
	lease.timer.Stop()
	lease.stopContext()
	if lease.activity != nil {
		lease.activity()
	}
	s.releaseObject(lease)
}

// releaseObject 要求持有 Store.mu；HTTP 读者和发送租约共用对象回收计数。
func (s *Store) releaseObject(lease *mediaLease) {
	if lease.upload != nil {
		lease.upload.leases--
		if lease.upload.retired && lease.upload.leases == 0 && lease.upload.grants == 0 {
			s.release(nil, "", lease.upload)
		}
	} else {
		lease.incoming.leases--
		s.sweepIncoming()
	}
}
