package media

import (
	"context"
	"errors"
	"fmt"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

func allowsTarget(targets []*pb.Target, target *pb.Target) bool {
	return slices.ContainsFunc(targets, func(scoped *pb.Target) bool { return proto.Equal(scoped, target) })
}

func targetSubset(targets, original []*pb.Target) bool {
	for _, target := range targets {
		if !allowsTarget(original, target) {
			return false
		}
	}
	return true
}

// grantSource 在持有 Store.mu 并清理过期 token 后检查原令牌的权利与范围。
// 上传对象由 owner 直接委托；Adapter 消费租约与上传委托本身只授予读取权。
func (s *Store) grantSource(c router.Caller, request *pb.GrantRequest) (*incomingAccess, *entry) {
	if access, err := s.incomingAccess(request.Token, c.SessionKey); err == nil {
		if targetSubset(request.Targets, access.targets) {
			return access, nil
		}
		return nil, nil
	}
	upload := s.refs[request.Token]
	if upload != nil && upload.owner == c.SessionKey && upload.lease == nil && upload.grant == nil && targetSubset(request.Targets, upload.targets) {
		return nil, upload
	}
	return nil, nil
}

func (s *Store) grant(ctx context.Context, c router.Caller, in *wire.Inbound, dispatch *router.Router, allowed func(context.Context, string, string) (bool, error)) wire.Response {
	request := &pb.GrantRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err)
	}
	var defaultTarget *pb.Target
	if invocation, ok := in.ParentAttach.(*router.Invocation); ok {
		defaultTarget = invocation.DefaultTarget
	}
	for i, target := range request.Targets {
		completed, completionError := s.readers.Complete(target, defaultTarget)
		if completionError != nil {
			return wire.Response{Err: completionError}
		}
		request.Targets[i] = completed
	}
	denied := failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, errors.New("media token does not allow grant"))
	// 1. 先核对原 token，查询其他所有者的状态时释放媒体锁。
	s.mu.Lock()
	s.sweep()
	incoming, upload := s.grantSource(c, request)
	s.mu.Unlock()
	if incoming == nil && upload == nil {
		return denied
	}
	grantee, ok := dispatch.Caller(request.GranteePluginId)
	if !ok {
		return denied
	}
	if incoming != nil && (incoming.event == nil || !dispatch.HasInFlight(grantee, incoming.event)) {
		permitted, err := allowed(ctx, grantee.PluginID, incoming.object.source.instance.InstanceID())
		if err != nil {
			return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, fmt.Errorf("read allowed adapter instances: %w", err))
		}
		if !permitted {
			return denied
		}
	}
	// 2. 查询期间可能发生替换、撤销或到期；签发前重核接收 Session 与原 token。
	if current, ok := dispatch.Caller(request.GranteePluginId); !ok || current != grantee {
		return denied
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	incoming, upload = s.grantSource(c, request)
	if incoming == nil && upload == nil {
		return denied
	}
	var targets []*pb.Target
	for _, target := range request.Targets {
		targets = append(targets, proto.CloneOf(target))
	}
	id := token()
	result := &pb.GrantResponse{Token: id}
	if incoming != nil {
		s.incomingTokens[id] = &incomingAccess{object: incoming.object, sessionKey: grantee.SessionKey, expires: incoming.expires, targets: targets, event: incoming.event, unsupported: incoming.unsupported}
		result.ExpiresAtUnixMs = incoming.expires.UnixMilli()
	} else {
		upload.grants++
		s.refs[id] = &entry{file: upload.file, owner: grantee.SessionKey, grant: upload, targets: targets}
		result.ExpiresAtUnixMs = upload.file.ExpiresAt.UnixMilli()
	}
	return response(result)
}
