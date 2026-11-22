// Package convert 在 SDK 的协议边界转换消息、事件、身份、回执与 Adapter 声明。
//
// 公共 SDK 使用 message 与 errcode 类型；本包保留 protobuf 编解码及可选字段语义。
package convert

import (
	"errors"
	"fmt"
	"slices"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/messagecodec"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func optional[T comparable](value T) *T {
	var zero T
	if value == zero {
		return nil
	}
	return &value
}
func timestamp(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}
func milliseconds(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

// SessionFromProto 提取不含账号和实例的会话字符串。
func SessionFromProto(t *pb.Target) string {
	if t == nil {
		return ""
	}
	return message.SID(t.AdapterId, t.ChatId, t.GetThreadId())
}

// TargetToProto 拆分会话与可选实例，账号由 Core 补齐。
func TargetToProto(s string, instance string) (*pb.Target, error) {
	a, c, t, err := message.ParseSID(s)
	if err != nil {
		return nil, err
	}
	return &pb.Target{AdapterId: a, ChatId: c, ThreadId: optional(t), AdapterInstanceId: instance}, nil
}

// RefToProto 拆分消息句柄中的目标、消息 ID 与实例。
func RefToProto(ref string) (*pb.Target, string, error) {
	s, id, instance, err := message.ParseRef(ref)
	if err != nil {
		return nil, "", err
	}
	target, err := TargetToProto(s, instance)
	return target, id, err
}

// ErrorFromProto 保留安全错误的数值和文案。
func ErrorFromProto(e *pb.ErrorInfo) *errcode.Error {
	if e == nil {
		return nil
	}
	return &errcode.Error{Code: errcode.Code(e.Code), Message: e.SafeMessage}
}

// ErrorToProto 转换 SDK 安全错误。
func ErrorToProto(e *errcode.Error) *pb.ErrorInfo {
	if e == nil {
		return nil
	}
	return &pb.ErrorInfo{Code: pb.ErrorCode(e.Code), SafeMessage: e.Message}
}

// CapabilitiesFromProto 转换平台能力集合。
func CapabilitiesFromProto(in []*pb.Capability) []message.Capability {
	var out []message.Capability
	for _, c := range in {
		capability := message.Capability{Name: c.Name, Version: c.Version, Supported: c.Supported}
		if c.TextLimit != nil {
			capability.TextLimit = &message.TextLimit{Max: c.TextLimit.Max, Unit: message.TextUnit(c.TextLimit.Unit)}
		}
		out = append(out, capability)
	}
	return out
}

// CapabilitiesToProto 转换 SDK 的平台能力集合。
func CapabilitiesToProto(in []message.Capability) []*pb.Capability {
	var out []*pb.Capability
	for _, c := range in {
		capability := &pb.Capability{Name: c.Name, Version: c.Version, Supported: c.Supported}
		if c.TextLimit != nil {
			capability.TextLimit = &pb.TextLimit{Max: c.TextLimit.Max, Unit: pb.TextUnit(c.TextLimit.Unit)}
		}
		out = append(out, capability)
	}
	return out
}

// ContentChain 将单片段归一为链；nil 表示没有内容。
func ContentChain(m message.Content) message.Chain {
	switch c := m.(type) {
	case message.Chain:
		return c
	case message.Segment:
		return message.Chain{c}
	default:
		return nil
	}
}

// ContentToProto 在显式发送边界拒绝空内容。
func ContentToProto(m message.Content, target *pb.Target) (*pb.MessageChain, error) {
	c := ContentChain(m)
	if len(c) == 0 {
		return nil, errcode.InvalidArgument.New("message chain must not be empty")
	}
	return chainToProto(c, target, true)
}

// ChainToProto 保留空链与 absent 的区别，递归转换全部片段。
func ChainToProto(c message.Chain, target *pb.Target) (*pb.MessageChain, error) {
	return chainToProto(c, target, false)
}

func chainToProto(c message.Chain, target *pb.Target, outbound bool) (*pb.MessageChain, error) {
	if c == nil {
		return nil, nil
	}
	out := &pb.MessageChain{}
	for _, s := range c {
		encoded, err := segmentToProto(s, target, outbound)
		if err != nil {
			return nil, err
		}
		out.Segments = append(out.Segments, encoded)
	}
	return out, nil
}
func segmentToProto(s message.Segment, enclosingTarget *pb.Target, outbound bool) (*pb.Segment, error) {
	if err := messagecodec.Error(s); err != nil {
		return nil, errcode.InvalidArgument.New("encode message: " + err.Error())
	}
	out := &pb.Segment{}
	switch v := s.(type) {
	case message.MarkdownSegment:
		out.Data = &pb.Segment_Markdown{Markdown: &pb.MarkdownSegment{Content: v.Content}}
	case message.TextSegment:
		text := &pb.TextSegment{Text: v.Text, Format: optional(pb.TextFormat(v.Format))}
		for _, e := range v.Entities {
			text.Entities = append(text.Entities, &pb.TextEntity{Type: pb.EntityType(e.Type), OffsetUtf16: e.OffsetUTF16, LengthUtf16: e.LengthUTF16, Url: optional(e.URL), Language: optional(e.Language), ExternalUserId: optional(e.ExternalUserID)})
		}
		out.Data = &pb.Segment_Text{Text: text}
	case message.ImageSegment:
		src, err := sourceToProto(v.Source)
		if err != nil {
			return nil, err
		}
		out.Data = &pb.Segment_Image{Image: &pb.ImageSegment{Source: src, Alt: optional(v.Alt), Caption: optional(v.Caption)}}
	case message.VideoSegment:
		if v.Duration < 0 {
			return nil, errcode.InvalidArgument.New("video duration must not be negative")
		}
		src, err := sourceToProto(v.Source)
		if err != nil {
			return nil, err
		}
		out.Data = &pb.Segment_Video{Video: &pb.VideoSegment{Source: src, DurationMs: optional(uint64(v.Duration.Milliseconds())), Caption: optional(v.Caption)}}
	case message.VoiceSegment:
		if v.Duration < 0 {
			return nil, errcode.InvalidArgument.New("voice duration must not be negative")
		}
		src, err := sourceToProto(v.Source)
		if err != nil {
			return nil, err
		}
		out.Data = &pb.Segment_Voice{Voice: &pb.VoiceSegment{Source: src, DurationMs: optional(uint64(v.Duration.Milliseconds())), Caption: optional(v.Caption)}}
	case message.FileSegment:
		src, err := sourceToProto(v.Source)
		if err != nil {
			return nil, err
		}
		out.Data = &pb.Segment_File{File: &pb.FileSegment{Source: src, Filename: optional(v.Filename), Mime: optional(v.MIME), Caption: optional(v.Caption)}}
	case message.MentionSegment:
		out.Data = &pb.Segment_Mention{Mention: &pb.MentionSegment{ExternalUserId: v.ExternalUserID, Display: optional(v.Display)}}
	case message.ReferenceSegment:
		target, id, err := RefToProto(v.Ref)
		if err != nil {
			return nil, err
		}
		// 仅对原先省略目标且句柄未改动的引用恢复省略，跨目标发送仍保留原会话和实例。
		if messagecodec.ImplicitReference(v) == v.Ref && SessionFromProto(target) == SessionFromProto(enclosingTarget) && target.AdapterInstanceId == enclosingTarget.GetAdapterInstanceId() {
			target = nil
		}
		out.Data = &pb.Segment_Reference{Reference: &pb.ReferenceSegment{Target: target, OriginalPlatformMessageId: id}}
	case message.ButtonsSegment:
		b := &pb.ButtonsSegment{}
		for _, row := range v.Rows {
			r := &pb.ButtonRow{}
			for _, button := range row {
				if err := messagecodec.ButtonError(button); err != nil {
					return nil, errcode.InvalidArgument.New("encode message: " + err.Error())
				}
				r.Buttons = append(r.Buttons, &pb.Button{ActionId: button.ActionID, Label: button.Label, Value: optional(button.Value), CoreToken: optional(button.CoreToken), Url: optional(button.URL)})
			}
			b.Rows = append(b.Rows, r)
		}
		out.Data = &pb.Segment_Buttons{Buttons: b}
	case message.ExtensionSegment:
		out.Data = &pb.Segment_Extension{Extension: &pb.ExtensionSegment{AdapterId: v.AdapterID, Name: v.Name, Version: v.Version, PayloadJson: v.PayloadJSON}}
	case message.ForwardSegment:
		f := &pb.ForwardSegment{PlatformForwardId: optional(v.PlatformForwardID)}
		for _, item := range v.Items {
			chain, err := chainToProto(item.Chain, enclosingTarget, outbound)
			if err != nil {
				return nil, err
			}
			next := &pb.ForwardItem{Chain: chain, PlatformMessageId: optional(item.PlatformMessageID), OccurredAtUnixMs: milliseconds(item.OccurredAt)}
			if item.Source != nil {
				next.Source = &pb.ForwardSource{DisplayName: optional(item.Source.DisplayName), ExternalUserId: optional(item.Source.ExternalUserID)}
			}
			f.Items = append(f.Items, next)
		}
		out.Data = &pb.Segment_Forward{Forward: f}
		if outbound {
			if err := validate.Chain(&pb.MessageChain{Segments: []*pb.Segment{out}}); err != nil {
				var validation *validate.Error
				if errors.As(err, &validation) {
					return nil, errcode.Code(validation.Code).New(validation.Message)
				}
				return nil, err
			}
		}
	default:
		return nil, errcode.InvalidArgument.New("missing message segment")
	}
	return out, nil
}
func sourceToProto(s message.MediaSource) (*pb.MediaSource, error) {
	if s.PlatformFile != "" {
		return nil, errcode.InvalidArgument.New("platform_file is reserved for Core")
	}
	count := 0
	out := &pb.MediaSource{Mime: s.MIME}
	if s.MediaRef != "" {
		count++
		out.Source = &pb.MediaSource_MediaRef{MediaRef: s.MediaRef}
	}
	if s.Incoming != nil {
		count++
		out.Source = &pb.MediaSource_IncomingRef{IncomingRef: incomingToProto(s.Incoming)}
	}
	if s.HTTPSURL != "" {
		count++
		out.Source = &pb.MediaSource_HttpsUrl{HttpsUrl: s.HTTPSURL}
	}
	if s.InlineBase64 != "" {
		count++
		out.Source = &pb.MediaSource_InlineBase64{InlineBase64: s.InlineBase64}
	}
	if count != 1 {
		return nil, errcode.InvalidArgument.New("media source must have exactly one source")
	}
	return out, nil
}
func sourceFromProto(s *pb.MediaSource) message.MediaSource {
	if s == nil {
		return message.MediaSource{}
	}
	return message.MediaSource{MIME: s.GetMime(), MediaRef: s.GetMediaRef(), Incoming: incomingFromProto(s.GetIncomingRef()), HTTPSURL: s.GetHttpsUrl(), InlineBase64: s.GetInlineBase64(), PlatformFile: s.GetPlatformFile()}
}
func incomingFromProto(m *pb.IncomingMedia) *message.IncomingMedia {
	if m == nil {
		return nil
	}
	return &message.IncomingMedia{Token: m.Token, Size: m.GetSize(), MIME: m.GetMime(), Readability: message.Readability(m.Readability), ExpiresAt: timestamp(m.ExpiresAtUnixMs)}
}
func incomingToProto(m *message.IncomingMedia) *pb.IncomingMedia {
	return &pb.IncomingMedia{Token: m.Token, Size: optional(m.Size), Mime: optional(m.MIME), Readability: pb.MediaReadability(m.Readability), ExpiresAtUnixMs: milliseconds(m.ExpiresAt)}
}

// ImageFromProto 转换命令中的 image 参数。
func ImageFromProto(p *pb.ImageSegment) *message.ImageSegment {
	if p == nil {
		return nil
	}
	return &message.ImageSegment{Source: sourceFromProto(p.Source), Alt: p.GetAlt(), Caption: p.GetCaption()}
}

// ChainFromProto 将全部片段转为 SDK 值类型，省略目标的引用继承所属目标。
func ChainFromProto(c *pb.MessageChain, enclosingTarget *pb.Target) message.Chain {
	if c == nil {
		return nil
	}
	out := make(message.Chain, 0, len(c.Segments))
	for _, s := range c.Segments {
		switch v := s.Data.(type) {
		case *pb.Segment_Markdown:
			out = append(out, message.Markdown(v.Markdown.GetContent()))
		case *pb.Segment_Text:
			text := message.TextSegment{Text: v.Text.Text, Format: message.TextFormat(v.Text.GetFormat())}
			for _, e := range v.Text.Entities {
				text.Entities = append(text.Entities, message.TextEntity{Type: message.EntityType(e.Type), OffsetUTF16: e.OffsetUtf16, LengthUTF16: e.LengthUtf16, URL: e.GetUrl(), Language: e.GetLanguage(), ExternalUserID: e.GetExternalUserId()})
			}
			out = append(out, text)
		case *pb.Segment_Image:
			out = append(out, *ImageFromProto(v.Image))
		case *pb.Segment_Video:
			out = append(out, message.VideoSegment{Source: sourceFromProto(v.Video.Source), Duration: time.Duration(v.Video.GetDurationMs()) * time.Millisecond, Caption: v.Video.GetCaption()})
		case *pb.Segment_Voice:
			out = append(out, message.VoiceSegment{Source: sourceFromProto(v.Voice.Source), Duration: time.Duration(v.Voice.GetDurationMs()) * time.Millisecond, Caption: v.Voice.GetCaption()})
		case *pb.Segment_File:
			out = append(out, message.FileSegment{Source: sourceFromProto(v.File.Source), Filename: v.File.GetFilename(), MIME: v.File.GetMime(), Caption: v.File.GetCaption()})
		case *pb.Segment_Mention:
			out = append(out, message.MentionSegment{ExternalUserID: v.Mention.ExternalUserId, Display: v.Mention.GetDisplay()})
		case *pb.Segment_Reference:
			target := v.Reference.Target
			if target == nil {
				target = enclosingTarget
			}
			reference := message.ReferenceSegment{Ref: message.Ref(SessionFromProto(target), v.Reference.OriginalPlatformMessageId, target.GetAdapterInstanceId())}
			if v.Reference.Target == nil {
				messagecodec.SetImplicitReference(&reference, reference.Ref)
			}
			out = append(out, reference)
		case *pb.Segment_Buttons:
			buttons := message.ButtonsSegment{}
			for _, row := range v.Buttons.Rows {
				var r []message.Button
				for _, b := range row.Buttons {
					r = append(r, message.Button{ActionID: b.ActionId, Label: b.Label, Value: b.GetValue(), CoreToken: b.GetCoreToken(), URL: b.GetUrl()})
				}
				buttons.Rows = append(buttons.Rows, r)
			}
			out = append(out, buttons)
		case *pb.Segment_Extension:
			out = append(out, message.ExtensionSegment{AdapterID: v.Extension.AdapterId, Name: v.Extension.Name, Version: v.Extension.Version, PayloadJSON: v.Extension.PayloadJson})
		case *pb.Segment_Forward:
			f := message.ForwardSegment{PlatformForwardID: v.Forward.GetPlatformForwardId()}
			for _, item := range v.Forward.Items {
				next := message.ForwardItem{Chain: ChainFromProto(item.GetChain(), enclosingTarget), PlatformMessageID: item.GetPlatformMessageId(), OccurredAt: timestamp(item.GetOccurredAtUnixMs())}
				if item.GetSource() != nil {
					next.Source = &message.ForwardSource{DisplayName: item.Source.GetDisplayName(), ExternalUserID: item.Source.GetExternalUserId()}
				}
				f.Items = append(f.Items, next)
			}
			out = append(out, f)
		}
	}
	return out
}

// MediaRefFromProto 转换上传回执与范围。
func MediaRefFromProto(m *pb.MediaRef) *message.MediaRef {
	if m == nil {
		return nil
	}
	out := &message.MediaRef{Token: m.Token, Size: m.Size, MIME: m.Mime, SHA256: slices.Clone(m.Sha256), ExpiresAt: timestamp(m.ExpiresAtUnixMs), ConsumerPluginIDs: slices.Clone(m.GetScope().GetConsumerPluginIds())}
	for _, t := range m.GetScope().GetTargets() {
		out.Sessions = append(out.Sessions, SessionFromProto(t))
	}
	return out
}

// MediaRefToProto 转换媒体范围，目标账号与实例由 Core 负责补齐。
func MediaRefToProto(m *message.MediaRef) (*pb.MediaRef, error) {
	if m == nil {
		return nil, nil
	}
	out := &pb.MediaRef{Token: m.Token, Size: m.Size, Mime: m.MIME, Sha256: slices.Clone(m.SHA256), ExpiresAtUnixMs: milliseconds(m.ExpiresAt)}
	if len(m.ConsumerPluginIDs) > 0 || len(m.Sessions) > 0 {
		out.Scope = &pb.MediaScope{ConsumerPluginIds: slices.Clone(m.ConsumerPluginIDs)}
	}
	for _, s := range m.Sessions {
		t, err := TargetToProto(s, "")
		if err != nil {
			return nil, fmt.Errorf("media scope: %w", err)
		}
		out.Scope.Targets = append(out.Scope.Targets, t)
	}
	return out, nil
}
