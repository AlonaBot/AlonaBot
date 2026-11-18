package message

import (
	"context"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"google.golang.org/protobuf/proto"
)

type replyKey struct {
	adapter, instance, account, chat, thread, message string
	threadPresent                                     bool
}

type replyDecoration struct {
	mention *pb.MentionSegment
	prefix  string
}

type replySession struct {
	messages map[replyKey]replyDecoration // Service.replyMu 保护；Release 后在途结果不再发布。
}

// SetReply 注入按可信会话读取平台回复设置的入口；nil 关闭回复处理。
func (s *Service) SetReply(resolve func(context.Context, *pb.Target) (settings.Reply, error)) {
	s.replyMu.Lock()
	defer s.replyMu.Unlock()
	s.resolveReply = resolve
}

// Release 删除本次 Caller Session 的首条回执关联。
func (s *Service) Release(c router.Caller) {
	s.replyMu.Lock()
	defer s.replyMu.Unlock()
	delete(s.replySessions, c)
}

func messageReplyKey(target *pb.Target, id string) replyKey {
	return replyKey{target.GetAdapterId(), target.GetAdapterInstanceId(), target.GetAccountId(), target.GetChatId(), target.GetThreadId(), id, target.ThreadId != nil}
}

func replyRoot(inv *router.Invocation, target *pb.Target) *router.Invocation {
	if inv == nil || inv.Root == nil {
		return nil
	}
	root := inv.Root
	// 路由根由 Router 保存；PreEvent、后台 Service 和定时调用均没有消息路由根。
	if root.RouteID == "" || root.Event.GetEventType() != "message.created" || root.Event.GetSourceKind() != pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER || !proto.Equal(root.DefaultTarget, target) || !proto.Equal(root.Event.GetTarget(), target) {
		return nil
	}
	return root
}

func (s *Service) prepareReply(ctx context.Context, caller router.Caller, inv *router.Invocation, target *pb.Target, instance adapter.Instance, chain *pb.MessageChain) (*pb.MessageChain, replyDecoration, *replySession, error) {
	root := replyRoot(inv, target)
	if root == nil {
		return chain, replyDecoration{}, nil, nil
	}
	s.replyMu.Lock()
	resolve := s.resolveReply
	if resolve == nil {
		s.replyMu.Unlock()
		return chain, replyDecoration{}, nil, nil
	}
	// 配置读取前固定对象，Release 或新绑定都使本次准备失效。
	if s.replySessions == nil {
		s.replySessions = make(map[router.Caller]*replySession)
	}
	session := s.replySessions[caller]
	if session == nil {
		session = &replySession{messages: make(map[replyKey]replyDecoration)}
		s.replySessions[caller] = session
	}
	s.replyMu.Unlock()
	reply, err := resolve(ctx, target)
	if err != nil {
		return nil, replyDecoration{}, nil, err
	}
	// 首条占用只锁内存表；配置读取、根结束通知和平台 I/O 都在锁外。
	s.replyMu.Lock()
	if s.replySessions[caller] != session {
		s.replyMu.Unlock()
		return chain, replyDecoration{}, nil, nil
	}
	if s.replyRoots == nil {
		s.replyRoots = make(map[*router.Invocation]struct{})
	}
	_, claimed := s.replyRoots[root]
	if !claimed {
		s.replyRoots[root] = struct{}{}
	} else {
		session = nil
	}
	s.replyMu.Unlock()
	var decoration replyDecoration
	if !claimed {
		root.AfterFinished(func() {
			s.replyMu.Lock()
			delete(s.replyRoots, root)
			s.replyMu.Unlock()
		})
		decoration.prefix = reply.Prefix
		actor := root.Event.GetActor()
		if reply.MentionSender && root.Event.GetChatType() == pb.ChatType_CHAT_TYPE_GROUP && actor != nil && !mentionsUser(chain, actor.GetExternalUserId()) {
			display := actor.GetDisplayName()
			if display == "" {
				display = actor.GetExternalUserId()
			}
			decoration.mention = &pb.MentionSegment{ExternalUserId: actor.GetExternalUserId(), Display: proto.String(display)}
		}
		modified := proto.CloneOf(chain)
		applyReplyPrefix(modified, decoration.prefix)
		if decoration.mention != nil {
			modified.Segments = append([]*pb.Segment{{Data: &pb.Segment_Mention{Mention: proto.CloneOf(decoration.mention)}}}, modified.Segments...)
		}
		if replyTooLong(instance, modified) {
			decoration = replyDecoration{}
		} else {
			chain = modified
		}
		if reply.Quote && !hasReference(chain) {
			chain.Segments = append([]*pb.Segment{{Data: &pb.Segment_Reference{Reference: &pb.ReferenceSegment{OriginalPlatformMessageId: root.Event.GetEventPayload().GetMessageCreated().GetPlatformMessageId(), Target: proto.CloneOf(target)}}}}, chain.Segments...)
		}
	}
	if reply.ForwardThreshold > 0 && supportsReplyForward(instance) {
		count := 0
		for _, segment := range chain.Segments {
			count += utf8.RuneCountInString(segment.GetText().GetText())
		}
		if count > reply.ForwardThreshold {
			chain = &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Forward{Forward: &pb.ForwardSegment{Items: []*pb.ForwardItem{{Source: &pb.ForwardSource{ExternalUserId: proto.String(target.GetAccountId())}, Chain: chain}}}}}}}
		}
	}
	return chain, decoration, session, nil
}

func applyReplyPrefix(chain *pb.MessageChain, prefix string) {
	if prefix == "" {
		return
	}
	for i, segment := range chain.Segments {
		if text := segment.GetText(); text != nil {
			if text.GetFormat() == pb.TextFormat_TEXT_FORMAT_MARKDOWN {
				// plain 片段由平台渲染器转义，避免前缀参与 Markdown 解析。
				chain.Segments = slices.Insert(chain.Segments, i, &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: prefix}}})
				return
			}
			text.Text = prefix + text.Text
			shiftEntities(text, uint32(utf16Units(prefix)))
			return
		}
	}
	chain.Segments = append([]*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: prefix}}}}, chain.Segments...)
}

func shiftEntities(text *pb.TextSegment, units uint32) {
	for _, entity := range text.Entities {
		entity.OffsetUtf16 += units
	}
}

func utf16Units(text string) int {
	n := 0
	for _, r := range text {
		n += utf16.RuneLen(r)
	}
	return n
}

func mentionsUser(chain *pb.MessageChain, id string) bool {
	for _, segment := range chain.GetSegments() {
		if mention := segment.GetMention(); mention != nil && mention.GetExternalUserId() == id {
			return true
		}
		for _, entity := range segment.GetText().GetEntities() {
			if entity.GetType() == pb.EntityType_ENTITY_TYPE_TEXT_MENTION && entity.GetExternalUserId() == id {
				return true
			}
		}
	}
	return false
}

func hasReference(chain *pb.MessageChain) bool {
	for _, segment := range chain.Segments {
		if segment.GetReference() != nil {
			return true
		}
	}
	return false
}

func replyTooLong(instance adapter.Instance, chain *pb.MessageChain) bool {
	for _, capability := range instance.Capabilities() {
		if capability.GetName() != "message.send" || !capability.GetSupported() || capability.GetTextLimit() == nil {
			continue
		}
		limit := capability.TextLimit
		var length uint64
		for _, segment := range chain.Segments {
			text := segment.GetText().GetText()
			if mention := segment.GetMention(); mention != nil {
				text = mention.GetDisplay()
				if text == "" {
					text = mention.GetExternalUserId()
				}
			}
			switch limit.Unit {
			case pb.TextUnit_TEXT_UNIT_UTF16_CODE_UNITS:
				length += uint64(utf16Units(text))
			case pb.TextUnit_TEXT_UNIT_UNICODE_CODE_POINTS:
				length += uint64(utf8.RuneCountInString(text))
			case pb.TextUnit_TEXT_UNIT_BYTES:
				length += uint64(len(text))
			}
		}
		return length > limit.Max
	}
	return false
}

func supportsReplyForward(instance adapter.Instance) bool {
	for _, capability := range instance.Capabilities() {
		if capability.GetName() == forwardCapability && capability.GetSupported() {
			return true
		}
	}
	return false
}

// rememberReply 在 Hook 和平台发送后只记录实际保留的修饰；部分成功仍可提供可编辑 ID。
func (s *Service) rememberReply(caller router.Caller, session *replySession, target *pb.Target, chain *pb.MessageChain, result *pb.SendResult, decoration replyDecoration) {
	if session == nil {
		return
	}
	id := result.GetPrimaryPlatformMessageId()
	if id == "" && len(result.GetPlatformMessageIds()) > 0 {
		id = result.PlatformMessageIds[0]
	}
	if id == "" {
		return
	}
	// 只把该平台 ID 已确认接受的片段关联到它，后续失败或未尝试部分不计入。
	confirmed := make(map[uint32]bool)
	for _, part := range result.GetPartResults() {
		if part.GetStatus() == pb.PartStatus_PART_STATUS_ACCEPTED && slices.Contains(part.GetPlatformMessageIds(), id) {
			for _, index := range part.GetSegmentIndexes() {
				confirmed[index] = true
			}
		}
	}
	if len(chain.Segments) == 1 && chain.Segments[0].GetForward() != nil {
		accepted := confirmed[0]
		chain = chain.Segments[0].GetForward().GetItems()[0].GetChain()
		clear(confirmed)
		if accepted {
			for i := range chain.Segments {
				confirmed[uint32(i)] = true
			}
		}
	}
	if decoration.mention != nil {
		id := decoration.mention.GetExternalUserId()
		decoration.mention = nil
		for i, segment := range chain.GetSegments() {
			if !confirmed[uint32(i)] {
				continue
			}
			if mention := segment.GetMention(); mention != nil && mention.GetExternalUserId() == id {
				decoration.mention = proto.CloneOf(mention)
				if decoration.mention.GetDisplay() == "" {
					decoration.mention.Display = proto.String(id)
				}
				break
			}
			for _, entity := range segment.GetText().GetEntities() {
				if entity.GetType() == pb.EntityType_ENTITY_TYPE_TEXT_MENTION && entity.GetExternalUserId() == id {
					units := utf16.Encode([]rune(segment.GetText().GetText()))
					display := string(utf16.Decode(units[entity.OffsetUtf16 : entity.OffsetUtf16+entity.LengthUtf16]))
					decoration.mention = &pb.MentionSegment{ExternalUserId: id, Display: proto.String(display)}
					break
				}
			}
		}
	}
	prefix := decoration.prefix
	decoration.prefix = ""
	for i, segment := range chain.GetSegments() {
		if text := segment.GetText(); text != nil {
			leadingBytes, _ := replyMentionEnd(text, decoration.mention)
			if confirmed[uint32(i)] && strings.HasPrefix(text.Text[leadingBytes:], prefix) {
				decoration.prefix = prefix
			}
			break
		}
	}
	s.replyMu.Lock()
	defer s.replyMu.Unlock()
	if s.replySessions[caller] == session {
		session.messages[messageReplyKey(target, id)] = decoration
	}
}

// replyMentionEnd 定位合并进 text 开头的自动提及，分别返回 UTF-8 字节与 UTF-16 码元边界。
func replyMentionEnd(text *pb.TextSegment, mention *pb.MentionSegment) (int, uint32) {
	if mention != nil && strings.HasPrefix(text.Text, mention.GetDisplay()) {
		for _, entity := range text.Entities {
			if entity.GetType() == pb.EntityType_ENTITY_TYPE_TEXT_MENTION && entity.GetExternalUserId() == mention.GetExternalUserId() && entity.OffsetUtf16 == 0 && entity.LengthUtf16 == uint32(utf16Units(mention.GetDisplay())) {
				return len(mention.GetDisplay()), entity.LengthUtf16
			}
		}
	}
	return 0, 0
}

func (s *Service) preserveReply(caller router.Caller, target *pb.Target, id string, chain *pb.MessageChain, instance adapter.Instance) (*pb.MessageChain, *pb.ErrorInfo) {
	s.replyMu.Lock()
	var decoration replyDecoration
	if session := s.replySessions[caller]; session != nil {
		decoration = session.messages[messageReplyKey(target, id)]
	}
	s.replyMu.Unlock()
	if decoration.mention == nil && decoration.prefix == "" {
		return chain, nil
	}
	modified := proto.CloneOf(chain)
	var text *pb.TextSegment
	for _, segment := range modified.Segments {
		if segment.GetText() != nil {
			text = segment.GetText()
			break
		}
	}
	// 编辑只接受一个 text；已有开头的自动提及单独定位，避免把前缀叠到它之前。
	leadingBytes, leadingUnits := replyMentionEnd(text, decoration.mention)
	if decoration.prefix != "" && !strings.HasPrefix(text.Text[leadingBytes:], decoration.prefix) {
		text.Text = text.Text[:leadingBytes] + decoration.prefix + text.Text[leadingBytes:]
		for _, entity := range text.Entities {
			if entity.OffsetUtf16 >= leadingUnits {
				entity.OffsetUtf16 += uint32(utf16Units(decoration.prefix))
			} else if entity.OffsetUtf16+entity.LengthUtf16 > leadingUnits {
				entity.LengthUtf16 += uint32(utf16Units(decoration.prefix))
			}
		}
	}
	if mention := decoration.mention; mention != nil && !mentionsUser(modified, mention.GetExternalUserId()) {
		if text.GetFormat() == pb.TextFormat_TEXT_FORMAT_MARKDOWN {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "automatic reply mention cannot be combined with markdown editing"}
		}
		text.Format = nil // plain 与 entities 表达相同文本，格式字段和实体按契约互斥。
		units := uint32(utf16Units(mention.GetDisplay()))
		text.Text = mention.GetDisplay() + text.Text
		shiftEntities(text, units)
		text.Entities = append([]*pb.TextEntity{{Type: pb.EntityType_ENTITY_TYPE_TEXT_MENTION, LengthUtf16: units, ExternalUserId: proto.String(mention.GetExternalUserId())}}, text.Entities...)
	}
	if replyTooLong(instance, modified) {
		return chain, nil
	}
	if decoration.prefix != "" && text.GetFormat() == pb.TextFormat_TEXT_FORMAT_MARKDOWN {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "automatic reply prefix cannot be combined with markdown editing"}
	}
	return modified, nil
}
