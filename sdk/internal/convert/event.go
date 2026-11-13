package convert

import (
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// EventFromProto 为 Handler 创建与可信协议值分离的事件副本。
func EventFromProto(e *pb.Event) *message.Event {
	if e == nil {
		return nil
	}
	out := &message.Event{ID: e.EventId, Type: e.EventType, ChatType: message.ChatType(e.ChatType), Source: message.SourceKind(e.SourceKind), OccurredAt: timestamp(e.OccurredAtUnixMs), AdapterID: e.GetAdapterId(), AdapterInstanceID: e.GetAdapterInstanceId(), AccountID: e.GetAccountId(), Session: SessionFromProto(e.Target), Capabilities: CapabilitiesFromProto(e.Capabilities), PlatformPayload: slices.Clone(e.PlatformPayload), SourceEventID: e.GetSourceEventId()}
	out.ChatTitle = e.GetChatTitle()
	if a := e.Actor; a != nil {
		out.Actor = &message.Actor{ExternalUserID: a.ExternalUserId, AdapterInstanceID: a.AdapterInstanceId, DisplayName: a.GetDisplayName(), Locale: a.GetLocale(), Permissions: slices.Clone(a.Permissions)}
	}
	if m := e.Message; m != nil {
		out.Message = &message.PlatformMessage{PlatformMessageID: m.PlatformMessageId, Chain: ChainFromProto(m.Chain, e.Target), MediaGroupID: m.GetMediaGroupId()}
	}
	switch p := e.EventPayload.GetKind().(type) {
	case *pb.EventPayload_MessageCreated:
		out.Payload = &message.CreatedPayload{PlatformMessageID: p.MessageCreated.PlatformMessageId, Chain: ChainFromProto(p.MessageCreated.Chain, e.Target), ToMe: p.MessageCreated.ToMe}
	case *pb.EventPayload_MessageEdited:
		out.Payload = &message.EditedPayload{PlatformMessageID: p.MessageEdited.PlatformMessageId, Chain: ChainFromProto(p.MessageEdited.Chain, e.Target), OldChain: ChainFromProto(p.MessageEdited.OldChain, e.Target)}
	case *pb.EventPayload_MessageDeleted:
		out.Payload = &message.DeletedPayload{PlatformMessageIDs: slices.Clone(p.MessageDeleted.PlatformMessageIds)}
	case *pb.EventPayload_ButtonClicked:
		v := p.ButtonClicked
		out.Payload = &message.ButtonClickedPayload{ActionID: v.ActionId, ActionSchemaVersion: v.ActionSchemaVersion, Value: v.Value, CallbackQueryID: v.CallbackQueryId, OriginPlatformMessageID: v.GetOriginPlatformMessageId()}
	case *pb.EventPayload_MemberJoined:
		out.Payload = &message.MemberPayload{MemberExternalUserID: p.MemberJoined.MemberExternalUserId}
	case *pb.EventPayload_MemberLeft:
		out.Payload = &message.MemberPayload{MemberExternalUserID: p.MemberLeft.MemberExternalUserId}
	case *pb.EventPayload_ReactionAdded:
		out.Payload = &message.ReactionAddedPayload{PlatformMessageID: p.ReactionAdded.PlatformMessageId, Reaction: p.ReactionAdded.Reaction}
	case *pb.EventPayload_Scheduled:
		out.Payload = &message.ScheduledPayload{JobID: p.Scheduled.JobId, RunID: p.Scheduled.RunId, ScheduledAt: timestamp(p.Scheduled.ScheduledAtUnixMs)}
	case *pb.EventPayload_BotOnline:
		v := p.BotOnline
		out.Payload = &message.BotOnlinePayload{AdapterID: v.AdapterId, AdapterInstanceID: v.AdapterInstanceId, AccountID: v.AccountId, AccountDisplayName: v.GetAccountDisplayName()}
	case *pb.EventPayload_BotOffline:
		v := p.BotOffline
		out.Payload = &message.BotOfflinePayload{AdapterID: v.AdapterId, AdapterInstanceID: v.AdapterInstanceId, AccountID: v.AccountId, Reason: message.OfflineReason(v.Reason), SafeError: ErrorFromProto(v.SafeError)}
	case *pb.EventPayload_AdapterEvent:
		out.Payload = &message.AdapterEventPayload{Version: p.AdapterEvent.Version, PayloadJSON: p.AdapterEvent.PayloadJson}
	}
	return out
}

// EventToProto 编码 Adapter 事件，目标由会话与可信实例、账号组成。
func EventToProto(e *message.Event) (*pb.Event, error) {
	if e == nil {
		return nil, nil
	}
	out := &pb.Event{EventId: e.ID, EventType: e.Type, ChatType: pb.ChatType(e.ChatType), SourceKind: pb.EventSourceKind(e.Source), OccurredAtUnixMs: milliseconds(e.OccurredAt), AdapterId: optional(e.AdapterID), AdapterInstanceId: optional(e.AdapterInstanceID), AccountId: optional(e.AccountID), Capabilities: CapabilitiesToProto(e.Capabilities), PlatformPayload: slices.Clone(e.PlatformPayload), SourceEventId: optional(e.SourceEventID)}
	out.ChatTitle = optional(e.ChatTitle)
	if e.Session != "" {
		target, err := TargetToProto(e.Session, e.AdapterInstanceID)
		if err != nil {
			return nil, err
		}
		target.AccountId = e.AccountID
		out.Target = target
	}
	if a := e.Actor; a != nil {
		out.Actor = &pb.Actor{ExternalUserId: a.ExternalUserID, AdapterInstanceId: a.AdapterInstanceID, DisplayName: optional(a.DisplayName), Locale: optional(a.Locale), Permissions: slices.Clone(a.Permissions)}
	}
	if m := e.Message; m != nil {
		chain, err := ChainToProto(m.Chain, out.Target)
		if err != nil {
			return nil, err
		}
		out.Message = &pb.PlatformMessage{PlatformMessageId: m.PlatformMessageID, Chain: chain, MediaGroupId: optional(m.MediaGroupID)}
	}
	payload := &pb.EventPayload{}
	switch p := e.Payload.(type) {
	case *message.CreatedPayload:
		chain, err := ChainToProto(p.Chain, out.Target)
		if err != nil {
			return nil, err
		}
		payload.Kind = &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{PlatformMessageId: p.PlatformMessageID, Chain: chain, ToMe: p.ToMe}}
	case *message.EditedPayload:
		chain, err := ChainToProto(p.Chain, out.Target)
		if err != nil {
			return nil, err
		}
		old, err := ChainToProto(p.OldChain, out.Target)
		if err != nil {
			return nil, err
		}
		payload.Kind = &pb.EventPayload_MessageEdited{MessageEdited: &pb.MessageEditedPayload{PlatformMessageId: p.PlatformMessageID, Chain: chain, OldChain: old}}
	case *message.DeletedPayload:
		payload.Kind = &pb.EventPayload_MessageDeleted{MessageDeleted: &pb.MessageDeletedPayload{PlatformMessageIds: slices.Clone(p.PlatformMessageIDs)}}
	case *message.ButtonClickedPayload:
		payload.Kind = &pb.EventPayload_ButtonClicked{ButtonClicked: &pb.ButtonClickedPayload{ActionId: p.ActionID, ActionSchemaVersion: p.ActionSchemaVersion, Value: p.Value, CallbackQueryId: p.CallbackQueryID, OriginPlatformMessageId: optional(p.OriginPlatformMessageID)}}
	case *message.MemberPayload:
		member := &pb.MemberPayload{MemberExternalUserId: p.MemberExternalUserID}
		if e.Type == "member.joined" {
			payload.Kind = &pb.EventPayload_MemberJoined{MemberJoined: member}
		} else {
			payload.Kind = &pb.EventPayload_MemberLeft{MemberLeft: member}
		}
	case *message.ReactionAddedPayload:
		payload.Kind = &pb.EventPayload_ReactionAdded{ReactionAdded: &pb.ReactionAddedPayload{PlatformMessageId: p.PlatformMessageID, Reaction: p.Reaction}}
	case *message.ScheduledPayload:
		payload.Kind = &pb.EventPayload_Scheduled{Scheduled: &pb.ScheduledPayload{JobId: p.JobID, RunId: p.RunID, ScheduledAtUnixMs: milliseconds(p.ScheduledAt)}}
	case *message.BotOnlinePayload:
		payload.Kind = &pb.EventPayload_BotOnline{BotOnline: &pb.BotOnlinePayload{AdapterId: p.AdapterID, AdapterInstanceId: p.AdapterInstanceID, AccountId: p.AccountID, AccountDisplayName: optional(p.AccountDisplayName)}}
	case *message.BotOfflinePayload:
		payload.Kind = &pb.EventPayload_BotOffline{BotOffline: &pb.BotOfflinePayload{AdapterId: p.AdapterID, AdapterInstanceId: p.AdapterInstanceID, AccountId: p.AccountID, Reason: pb.BotOfflineReason(p.Reason), SafeError: ErrorToProto(p.SafeError)}}
	case *message.AdapterEventPayload:
		payload.Kind = &pb.EventPayload_AdapterEvent{AdapterEvent: &pb.AdapterEventPayload{Version: p.Version, PayloadJson: p.PayloadJSON}}
	}
	if e.Payload != nil {
		out.EventPayload = payload
	}
	return out, nil
}

// SendResultFromProto 保留整体状态、已确认消息与各物理操作的完整结果。
func SendResultFromProto(r *pb.SendResult) *message.SendResult {
	if r == nil {
		return nil
	}
	out := &message.SendResult{Status: message.SendStatus(r.Status), Session: SessionFromProto(r.Target), AdapterInstanceID: r.Target.GetAdapterInstanceId(), PlatformMessageIDs: slices.Clone(r.PlatformMessageIds), PrimaryPlatformMessageID: r.GetPrimaryPlatformMessageId(), ProviderAcceptedAt: timestamp(r.GetProviderAcceptedTimeUnixMs()), CapabilitiesUsed: CapabilitiesFromProto(r.CapabilitiesUsed)}
	for _, p := range r.PartResults {
		out.PartResults = append(out.PartResults, message.PartResult{PartIndex: p.PartIndex, SegmentIndexes: slices.Clone(p.SegmentIndexes), Status: message.PartStatus(p.Status), PlatformMessageIDs: slices.Clone(p.PlatformMessageIds), MessageIDsUnavailable: p.MessageIdsUnavailable, Error: ErrorFromProto(p.Error)})
	}
	return out
}

// SendResultToProto 编码 Adapter 回执；requestTarget 提供会话字符串省略的账号绑定。
func SendResultToProto(r *message.SendResult, requestTarget *pb.Target) (*pb.SendResult, error) {
	if r == nil {
		return nil, nil
	}
	out := &pb.SendResult{Status: pb.SendStatus(r.Status), PlatformMessageIds: slices.Clone(r.PlatformMessageIDs), PrimaryPlatformMessageId: optional(r.PrimaryPlatformMessageID), ProviderAcceptedTimeUnixMs: optional(milliseconds(r.ProviderAcceptedAt)), CapabilitiesUsed: CapabilitiesToProto(r.CapabilitiesUsed)}
	if r.Session != "" {
		target, err := TargetToProto(r.Session, r.AdapterInstanceID)
		if err != nil {
			return nil, err
		}
		if requestTarget != nil {
			target.AccountId = requestTarget.AccountId
		}
		out.Target = target
	}
	for _, p := range r.PartResults {
		out.PartResults = append(out.PartResults, &pb.PartResult{PartIndex: p.PartIndex, SegmentIndexes: slices.Clone(p.SegmentIndexes), Status: pb.PartStatus(p.Status), PlatformMessageIds: slices.Clone(p.PlatformMessageIDs), MessageIdsUnavailable: p.MessageIDsUnavailable, Error: ErrorToProto(p.Error)})
	}
	return out, nil
}
