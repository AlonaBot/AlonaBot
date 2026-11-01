package media

import (
	"context"
	"errors"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

type platformFile struct{ instanceID, fileID string }

// incomingObject 的状态由 Store.mu 保护；只有一个取件者在锁外写文件。
// fetching 关闭后发布取件结果，其他读取者共享同一文件或失败结果。
type incomingObject struct {
	source   platformFile
	size     *uint64
	mime     *string
	file     File
	expires  time.Time
	issued   bool
	fetching chan struct{}
	fetched  bool
	err      error
	reserved int64
	charges  map[string]int64 // 每个消费 Session 的引用预算，Core 磁盘只计一份
}

type incomingAccess struct {
	object     *incomingObject
	sessionKey string
	expires    time.Time
	targets    []*pb.Target
}

// RegisterIncoming 为可信实例的平台文件登记对象，同一实例与 fileID 复用对象。
// size 与 mime 保留未知状态；登记时不下载文件。
func (s *Store) RegisterIncoming(instanceID, fileID string, size *uint64, mime *string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	if s.incoming == nil {
		s.incoming = make(map[string]*incomingObject)
		s.bySource = make(map[platformFile]string)
		s.incomingTokens = make(map[string]*incomingAccess)
	}
	source := platformFile{instanceID, fileID}
	if id, ok := s.bySource[source]; ok {
		return id
	}
	id := token()
	object := &incomingObject{source: source, expires: s.now().Add(s.lim.TTL), charges: make(map[string]int64)}
	if size != nil {
		object.size = proto.Uint64(*size)
	}
	if mime != nil {
		object.mime = proto.String(*mime)
	}
	s.incoming[id] = object
	s.bySource[source] = id
	return id
}

// IssueIncoming 为一个接收 Session 签发独立、目标受限的读取令牌。
// 对象已被回收时返回 nil，调用方据此处理过期事件。
func (s *Store) IssueIncoming(objectID string, c router.Caller, targets []*pb.Target) *pb.IncomingMedia {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	object := s.incoming[objectID]
	if object == nil {
		return nil
	}
	expires := s.now().Add(s.lim.TTL)
	id := token()
	access := &incomingAccess{object: object, sessionKey: c.SessionKey, expires: expires}
	for _, target := range targets {
		access.targets = append(access.targets, proto.CloneOf(target))
	}
	s.incomingTokens[id] = access
	object.issued = true
	result := &pb.IncomingMedia{Token: id, ExpiresAtUnixMs: expires.UnixMilli(), Readability: pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED}
	if object.size != nil {
		result.Size = proto.Uint64(*object.size)
	}
	if object.mime != nil {
		result.Mime = proto.String(*object.mime)
	}
	if instance, ok := s.readers.Get(object.source.instanceID); ok {
		if _, ok := instance.(adapter.MediaReader); ok {
			result.Readability = pb.MediaReadability_MEDIA_READABILITY_READABLE
		}
	}
	return result
}

func mediaCallError(code pb.ErrorCode, message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func incomingFailure(err error) error {
	var call *wire.CallError
	if errors.As(err, &call) {
		code := pb.ErrorCode_ERROR_CODE_UNAVAILABLE
		if call.ErrorInfo.Code == pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
			code = pb.ErrorCode_ERROR_CODE_UNSUPPORTED
		}
		return mediaCallError(code, call.ErrorInfo.SafeMessage)
	}
	return mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, err.Error())
}

func (s *Store) incomingAccess(id, sessionKey string) (*incomingAccess, error) {
	access := s.incomingTokens[id]
	if access == nil || access.sessionKey != sessionKey || !s.now().Before(access.expires) {
		return nil, mediaCallError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "media token not found")
	}
	return access, nil
}

func (s *Store) uncharge(object *incomingObject, sessionKey string) {
	s.used[sessionKey] -= object.charges[sessionKey]
	delete(object.charges, sessionKey)
	if s.used[sessionKey] == 0 {
		delete(s.used, sessionKey)
	}
}

func (s *Store) releaseIncomingSession(sessionKey string) {
	for id, access := range s.incomingTokens {
		if access.sessionKey == sessionKey {
			delete(s.incomingTokens, id)
		}
	}
	for _, object := range s.incoming {
		if _, ok := object.charges[sessionKey]; ok {
			s.uncharge(object, sessionKey)
		}
	}
	s.sweepIncoming()
}

func (s *Store) sweepIncoming() {
	now := s.now()
	live := make(map[*incomingObject]map[string]bool)
	for id, access := range s.incomingTokens {
		if !now.Before(access.expires) {
			delete(s.incomingTokens, id)
			continue
		}
		if live[access.object] == nil {
			live[access.object] = make(map[string]bool)
		}
		live[access.object][access.sessionKey] = true
	}
	for id, object := range s.incoming {
		for sessionKey := range object.charges {
			if !live[object][sessionKey] {
				s.uncharge(object, sessionKey)
			}
		}
		// 取件期间保留预留和文件，由取件者统一发布结果或清理失败。
		if object.fetching != nil {
			continue
		}
		if len(live[object]) > 0 || (!object.issued && now.Before(object.expires)) {
			continue
		}
		if object.fetched {
			s.remove(object.file.Path)
		}
		s.total -= object.reserved
		delete(s.incoming, id)
		delete(s.bySource, object.source)
	}
}

func (s *Store) loadIncoming(ctx context.Context, c router.Caller, id string) error {
	for {
		// 1. 每次唤醒后重新校验令牌，只消费同一对象已发布的结果。
		s.mu.Lock()
		s.sweep()
		access, err := s.incomingAccess(id, c.SessionKey)
		if err != nil {
			s.mu.Unlock()
			return err
		}
		object := access.object
		if object.err != nil {
			s.mu.Unlock()
			return object.err
		}
		if object.fetching != nil {
			done := object.fetching
			s.mu.Unlock()
			// 共享取件结果，等待者取消只结束自己的读取。
			select {
			case <-done:
				continue
			case <-ctx.Done():
				return incomingFailure(ctx.Err())
			}
		}
		if object.fetched {
			if _, charged := object.charges[c.SessionKey]; !charged {
				size := int64(object.file.Size)
				if size > s.lim.MaxSessionBytes-s.used[c.SessionKey] {
					s.mu.Unlock()
					return mediaCallError(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "media quota exceeded")
				}
				object.charges[c.SessionKey] = size
				s.used[c.SessionKey] += size
			}
			s.mu.Unlock()
			return nil
		}
		// 2. 首个取件者在接收字节前同时预留 Core 与消费 Session 的预算。
		reservation := s.lim.MaxFileBytes
		if object.size != nil {
			if *object.size > uint64(s.lim.MaxFileBytes) {
				s.mu.Unlock()
				return mediaCallError(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "media object exceeds size limit")
			}
			reservation = int64(*object.size)
		}
		if reservation > s.lim.MaxSessionBytes-s.used[c.SessionKey] || reservation > s.lim.MaxCoreBytes-s.total {
			s.mu.Unlock()
			return mediaCallError(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "media quota exceeded")
		}
		instance, ok := s.readers.Get(object.source.instanceID)
		if !ok {
			s.mu.Unlock()
			return mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
		}
		reader, supported := instance.(adapter.MediaReader)
		if !supported {
			s.mu.Unlock()
			return mediaCallError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support media reads")
		}
		object.fetching = make(chan struct{})
		object.reserved = reservation
		object.charges[c.SessionKey] = reservation
		s.total += reservation
		s.used[c.SessionKey] += reservation
		budget := min(access.expires.Sub(s.now()), s.lim.TTL)
		s.mu.Unlock()

		// 3. 文件与平台 I/O 在锁外进行，其他对象和 Session 撤销可继续。
		fetchCtx, cancel := context.WithTimeout(ctx, budget)
		file, fetchErr := s.fetchIncoming(fetchCtx, id, object, reader, reservation)
		cancel()
		// 4. 成功后按实际大小计费，失败释放预留并保留本次失败结果。
		s.mu.Lock()
		if fetchErr != nil {
			object.err = fetchErr
			s.total -= object.reserved
			object.reserved = 0
			for sessionKey := range object.charges {
				s.uncharge(object, sessionKey)
			}
		} else {
			object.file = file
			object.fetched = true
			actual := int64(file.Size)
			s.total -= object.reserved - actual
			object.reserved = actual
			for sessionKey, charged := range object.charges {
				s.used[sessionKey] -= charged - actual
				object.charges[sessionKey] = actual
			}
		}
		// 先发布结果再唤醒等待者，下一轮重新核对令牌和 Session 是否仍有效。
		close(object.fetching)
		object.fetching = nil
		s.mu.Unlock()
	}
}
