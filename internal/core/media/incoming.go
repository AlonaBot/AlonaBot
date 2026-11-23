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

type platformFile struct {
	instance    adapter.Instance
	fileID      string
	unsupported bool
}

// incomingObject 的状态由 Store.mu 保护；只有一个取件者在锁外写文件。
// 每轮取件单独发布完成信号与错误，失败后对象可由新的读取调用重新取件。
type incomingObject struct {
	source   platformFile
	size     *uint64
	mime     *string
	file     File
	expires  time.Time
	issued   bool
	fetching *incomingFetch
	fetched  bool
	reserved int64
	charges  map[string]int64 // 每个消费 Session 的引用预算，Core 磁盘只计一份
	leases   int              // 发送消费租约数，由 Store.mu 保护
}

// incomingFetch 在 done 关闭后只读，等待者持有本轮结果，避免消费后来取件的结果。
type incomingFetch struct {
	done chan struct{}
	err  error
}

type incomingAccess struct {
	object      *incomingObject
	sessionKey  string
	expires     time.Time
	targets     []*pb.Target
	event       *pb.Event // 原事件的 event_id 快照；直接 IssueIncoming 时为 nil
	unsupported bool      // 保留签发时的来源可读性，Grant 继承该限制
}

// RegisterIncoming 为可信实例的平台文件登记对象，绑定该次运行和来源的可读性。
// size 与 mime 保留未知状态；登记时不下载文件。
func (s *Store) RegisterIncoming(instance adapter.Instance, media *pb.IncomingMedia) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	if s.incoming == nil {
		s.incoming = make(map[string]*incomingObject)
		s.bySource = make(map[platformFile]string)
		s.incomingTokens = make(map[string]*incomingAccess)
	}
	source := platformFile{instance, media.Token, media.Readability == pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED}
	if id, ok := s.bySource[source]; ok {
		return id
	}
	id := token()
	object := &incomingObject{source: source, expires: s.now().Add(s.lim.TTL), charges: make(map[string]int64)}
	if media.Size != nil {
		object.size = proto.Uint64(*media.Size)
	}
	if media.Mime != nil {
		object.mime = proto.String(*media.Mime)
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
	object.issued = true
	return s.issueIncoming(object, c, targets)
}

func (s *Store) issueIncoming(object *incomingObject, c router.Caller, targets []*pb.Target) *pb.IncomingMedia {
	expires := s.now().Add(s.lim.TTL)
	id := token()
	access := &incomingAccess{object: object, sessionKey: c.SessionKey, expires: expires}
	for _, target := range targets {
		access.targets = append(access.targets, proto.CloneOf(target))
	}
	s.incomingTokens[id] = access
	result := &pb.IncomingMedia{Token: id, ExpiresAtUnixMs: expires.UnixMilli(), Readability: pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED}
	if object.size != nil {
		result.Size = proto.Uint64(*object.size)
	}
	if object.mime != nil {
		result.Mime = proto.String(*object.mime)
	}
	if !object.source.unsupported {
		if _, ok := object.source.instance.(adapter.MediaReader); ok {
			result.Readability = pb.MediaReadability_MEDIA_READABILITY_READABLE
		}
	}
	access.unsupported = result.Readability == pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED
	return result
}

// PrepareIncoming 克隆事件，为接收 Session 改写所有入站来源，并记录原事件关联。
// Adapter 显式标记的 UNSUPPORTED 与每个来源的可选元数据保持原样。
func (s *Store) PrepareIncoming(ev *pb.Event, c router.Caller) (prepared *pb.Event, err error) {
	// 1. 保存独立事件与原事件标识，失败时撤销本次未派发的令牌。
	prepared = proto.CloneOf(ev)
	payload := prepared.GetEventPayload()
	chains := []*pb.MessageChain{prepared.GetMessage().GetChain(), payload.GetMessageCreated().GetChain(), payload.GetMessageEdited().GetChain(), payload.GetMessageEdited().GetOldChain()}
	err = s.prepareIncomingChains(chains, c, ev.Target, &pb.Event{EventId: ev.GetEventId()})
	return prepared, err
}

// PrepareIncomingChain 克隆查询结果并为接收 Session 签发目标受限令牌。
func (s *Store) PrepareIncomingChain(chain *pb.MessageChain, c router.Caller, target *pb.Target) (*pb.MessageChain, error) {
	prepared := proto.CloneOf(chain)
	err := s.prepareIncomingChains([]*pb.MessageChain{prepared}, c, target, nil)
	return prepared, err
}

func (s *Store) prepareIncomingChains(chains []*pb.MessageChain, c router.Caller, target *pb.Target, event *pb.Event) (err error) {
	var targets []*pb.Target
	if target != nil {
		targets = []*pb.Target{target}
	}
	var issued []string
	defer func() {
		if err != nil {
			// 某个来源过期或 Session 被撤销时，本次未派发的令牌一并回收。
			s.mu.Lock()
			defer s.mu.Unlock()
			for _, id := range issued {
				delete(s.incomingTokens, id)
			}
			s.sweepIncoming()
		}
	}()
	// 2. 递归改写消息链中的来源，保留可选元数据与来源的不可读限制。
	seen := make(map[*pb.IncomingMedia]bool)
	var rewrite func(*pb.MessageChain) error
	rewrite = func(chain *pb.MessageChain) error {
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
				seen[incoming] = true
				access := s.IssueIncoming(incoming.Token, c, targets)
				if access == nil {
					return mediaCallError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, "incoming media object not found")
				}
				issued = append(issued, access.Token)
				s.mu.Lock()
				bound, err := s.incomingAccess(access.Token, c.SessionKey)
				if err == nil {
					bound.event = event
					bound.unsupported = bound.unsupported || incoming.Readability == pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED
				}
				s.mu.Unlock()
				if err != nil {
					return err
				}
				incoming.Token = access.Token
				incoming.ExpiresAtUnixMs = access.ExpiresAtUnixMs
				if incoming.Readability != pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED {
					incoming.Readability = access.Readability
				}
			}
			for _, item := range segment.GetForward().GetItems() {
				if err := rewrite(item.GetChain()); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, chain := range chains {
		if err := rewrite(chain); err != nil {
			return err
		}
	}
	return nil
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
		// 取件和发送期间保留对象，由对应的完成路径发布结果或释放租约。
		if object.fetching != nil || object.leases > 0 {
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
		if access.unsupported {
			s.mu.Unlock()
			return mediaCallError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "media source does not support reads")
		}
		if object.fetching != nil {
			fetch := object.fetching
			s.mu.Unlock()
			// 共享取件结果，等待者取消只结束自己的读取。
			select {
			case <-fetch.done:
				if fetch.err != nil {
					return fetch.err
				}
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
		instance, ok := s.readers.Resolve(ctx, object.source.instance.InstanceID())
		if !ok || instance != object.source.instance {
			s.mu.Unlock()
			return mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "media source adapter instance is no longer running")
		}
		reader, supported := instance.(adapter.MediaReader)
		if !supported {
			s.mu.Unlock()
			return mediaCallError(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "adapter instance does not support media reads")
		}
		fetch := &incomingFetch{done: make(chan struct{})}
		object.fetching = fetch
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
			fetch.err = fetchErr
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
		// 关闭信号前写入本轮结果；后续新调用可重新取件，本轮调用直接返回失败。
		close(fetch.done)
		object.fetching = nil
		s.mu.Unlock()
		if fetchErr != nil {
			return fetchErr
		}
	}
}
