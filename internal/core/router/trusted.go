package router

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

// trustedEvent 分离来源绑定与可改写载荷；每条分发链独占其副本。
// retained 只在链开始时更新权限，此后不交给 Prepare 或插件。
type trustedEvent struct {
	retained        *pb.Event
	payload         *pb.EventPayload
	commandConsumed bool // 当前分发链的命令层消费结果，不接受插件回写。
}

func retainEvent(ev *pb.Event) *trustedEvent {
	retained := proto.CloneOf(ev)
	payload := retained.EventPayload
	bindings := &pb.EventPayload{}
	switch p := payload.GetKind().(type) {
	case *pb.EventPayload_MessageCreated:
		bindings.Kind = &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{PlatformMessageId: p.MessageCreated.GetPlatformMessageId()}}
		p.MessageCreated.PlatformMessageId = ""
	case *pb.EventPayload_MessageEdited:
		bindings.Kind = &pb.EventPayload_MessageEdited{MessageEdited: &pb.MessageEditedPayload{PlatformMessageId: p.MessageEdited.GetPlatformMessageId()}}
		p.MessageEdited.PlatformMessageId = ""
	case *pb.EventPayload_ReactionAdded:
		bindings.Kind = &pb.EventPayload_ReactionAdded{ReactionAdded: &pb.ReactionAddedPayload{PlatformMessageId: p.ReactionAdded.GetPlatformMessageId()}}
		p.ReactionAdded.PlatformMessageId = ""
	case *pb.EventPayload_MessageDeleted, *pb.EventPayload_ButtonClicked, *pb.EventPayload_Scheduled:
		bindings = payload
		payload = nil
	}
	retained.EventPayload = bindings
	return &trustedEvent{retained: retained, payload: payload}
}

func (e *trustedEvent) rebuild() *pb.Event {
	ev := proto.CloneOf(e.retained)
	bindings := ev.EventPayload
	ev.EventPayload = proto.CloneOf(e.payload)
	// 来源绑定按原分支恢复，零值也覆盖业务副本提供的同名字段。
	switch p := bindings.GetKind().(type) {
	case *pb.EventPayload_MessageCreated:
		body := &pb.MessageCreatedPayload{}
		if current := ev.EventPayload.GetMessageCreated(); current != nil {
			body = current
		}
		body.PlatformMessageId = p.MessageCreated.GetPlatformMessageId()
		ev.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: body}}
	case *pb.EventPayload_MessageEdited:
		body := &pb.MessageEditedPayload{}
		if current := ev.EventPayload.GetMessageEdited(); current != nil {
			body = current
		}
		body.PlatformMessageId = p.MessageEdited.GetPlatformMessageId()
		ev.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_MessageEdited{MessageEdited: body}}
	case *pb.EventPayload_ReactionAdded:
		body := &pb.ReactionAddedPayload{}
		if current := ev.EventPayload.GetReactionAdded(); current != nil {
			body = current
		}
		body.PlatformMessageId = p.ReactionAdded.GetPlatformMessageId()
		ev.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_ReactionAdded{ReactionAdded: body}}
	case *pb.EventPayload_MessageDeleted, *pb.EventPayload_ButtonClicked, *pb.EventPayload_Scheduled:
		ev.EventPayload = bindings
	}
	return ev
}
