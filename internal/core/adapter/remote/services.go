package remote

import (
	"context"
	"errors"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Services 注册只对本安装包开放的交付、游标与故障服务。
func (p *Process) Services(m *router.Mux) {
	m.Handle(wire.ServiceCoreAdapterDeliver, p.service)
	m.Handle(wire.ServiceCoreAdapterCursor, p.service)
	m.Handle(wire.ServiceCoreAdapterStatus, p.service)
}

func serviceError(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}
func serviceResponse(message proto.Message) wire.Response {
	out, err := payload(message)
	if err != nil {
		return serviceError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
	}
	return wire.Response{Payload: out}
}

func (p *Process) service(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 身份先于载荷处理；无权调用者不会得到实例信息。
	if caller.PluginID != p.opt.PluginID {
		return serviceError(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "caller does not own adapter process")
	}
	var id string
	var request proto.Message
	switch in.Target {
	case wire.ServiceCoreAdapterDeliver:
		request = &pb.AdapterDeliverRequest{}
	case wire.ServiceCoreAdapterCursor:
		request = &pb.AdapterCursorRequest{}
	case wire.ServiceCoreAdapterStatus:
		request = &pb.AdapterStatusRequest{}
	}
	if err := decode(in.Payload, request); err != nil {
		return serviceError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	switch r := request.(type) {
	case *pb.AdapterDeliverRequest:
		id = r.InstanceId
	case *pb.AdapterCursorRequest:
		id = r.InstanceId
	case *pb.AdapterStatusRequest:
		id = r.InstanceId
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	entry := p.instances[id]
	invalid := func() wire.Response {
		p.opt.Log.Warn("invalid adapter instance event", "plugin_id", p.opt.PluginID, "instance_id", id)
		return serviceError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter instance or event identity mismatch")
	}
	if ctx.Err() != nil || p.session == nil || entry == nil || !entry.running || entry.session != p.session.Session {
		return invalid()
	}
	// 2. 所有归属检查必须先于游标持久化。
	switch r := request.(type) {
	case *pb.AdapterDeliverRequest:
		ev := r.Event
		if ev == nil || !adapterEventPayload(ev) || ev.GetAdapterId() != p.opt.AdapterID || ev.GetAdapterInstanceId() != id || ev.GetAccountId() != entry.account {
			return invalid()
		}
		if target := ev.Target; target != nil && (target.AdapterId != p.opt.AdapterID || target.AdapterInstanceId != id || target.AccountId != entry.account) {
			return invalid()
		}
		if err := p.opt.Store.PutAdapterCursor(ctx, p.opt.PluginID, id, r.Cursor); err != nil {
			return serviceError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		// 3. Core 重建可信事件字段，Adapter 只保留平台群管理员事实。
		ev = proto.Clone(ev).(*pb.Event)
		ev.EventId = p.event(entry, ev.EventType).EventId
		ev.SourceKind = pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER
		ev.Capabilities = entry.capabilities
		if ev.Actor != nil {
			roles := ev.Actor.TrustedRoles
			ev.Actor.TrustedRoles = nil
			for _, role := range roles {
				if role == pb.TrustedRole_TRUSTED_ROLE_CHAT_ADMIN {
					ev.Actor.TrustedRoles = append(ev.Actor.TrustedRoles, role)
				}
			}
		}
		err := p.deliver(ctx, entry, ev)
		if err != nil && !errors.Is(err, adapter.ErrQueueFull) {
			return serviceError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		return serviceResponse(&pb.AdapterDeliverResponse{Accepted: err == nil})
	case *pb.AdapterCursorRequest:
		if err := p.opt.Store.PutAdapterCursor(ctx, p.opt.PluginID, id, r.Cursor); err != nil {
			return serviceError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		return serviceResponse(&pb.AdapterCursorResponse{})
	case *pb.AdapterStatusRequest:
		if r.State != pb.AdapterInstanceState_ADAPTER_INSTANCE_STATE_FAILED {
			return serviceError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "unsupported adapter instance state")
		}
		entry.enabled = false
		if err := p.offline(ctx, entry, pb.BotOfflineReason_BOT_OFFLINE_REASON_INSTANCE_FAILED, r.SafeError); err != nil && !errors.Is(err, adapter.ErrQueueFull) {
			return serviceError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		return serviceResponse(&pb.AdapterStatusResponse{})
	}
	panic("unreachable adapter service")
}

// adapterEventPayload 只接受平台事件，并核对类型与载荷的唯一分支。
func adapterEventPayload(ev *pb.Event) bool {
	switch ev.EventType {
	case "message.created":
		return ev.GetEventPayload().GetMessageCreated() != nil
	case "message.edited":
		return ev.GetEventPayload().GetMessageEdited() != nil
	case "message.deleted":
		return ev.GetEventPayload().GetMessageDeleted() != nil
	case "button.clicked":
		return ev.GetEventPayload().GetButtonClicked() != nil
	case "member.joined":
		return ev.GetEventPayload().GetMemberJoined() != nil
	case "member.left":
		return ev.GetEventPayload().GetMemberLeft() != nil
	case "reaction.added":
		return ev.GetEventPayload().GetReactionAdded() != nil
	default:
		return false
	}
}
