package validate

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
)

// Chain 检查 MessageChain 的结构（MESSAGE-MODEL §2）。只做与平台无关的结构检查；
// 平台能力（UNSUPPORTED）、媒体租约与目标授权由 Core/Adapter 在发送前另行检查。
// 失败返回 INVALID_ARGUMENT，消息指明出错的 segment 位置。
func Chain(c *pb.MessageChain) error {
	if err := chain(c, ""); err != nil {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Msg: err.Error()}
	}
	return nil
}

func chain(c *pb.MessageChain, at string) error {
	segs := c.GetSegments()
	if len(segs) == 0 {
		return fmt.Errorf("%sChain 为空", at)
	}
	var refs, buttons, content int
	for i, s := range segs {
		pos := fmt.Sprintf("%ssegment[%d]", at, i)
		var err error
		switch d := s.GetData().(type) {
		case *pb.Segment_Text:
			content++
			err = text(d.Text)
		case *pb.Segment_Image:
			content++
			err = media(d.Image.GetSource())
		case *pb.Segment_Video:
			content++
			err = media(d.Video.GetSource())
		case *pb.Segment_File:
			content++
			if err = media(d.File.GetSource()); err == nil && d.File.Filename != nil {
				err = filename(d.File.GetFilename())
			}
		case *pb.Segment_Mention:
			content++
			if d.Mention.GetExternalUserId() == "" {
				err = fmt.Errorf("mention 缺少 external_user_id")
			}
		case *pb.Segment_Reference:
			refs++
			err = reference(d.Reference)
		case *pb.Segment_Buttons:
			buttons++
			err = buttonRows(d.Buttons)
		case *pb.Segment_Forward:
			content++
			err = forward(d.Forward, pos)
		default:
			err = fmt.Errorf("片段没有设置类型")
		}
		if err != nil {
			if strings.HasPrefix(err.Error(), pos) {
				return err // 嵌套 Chain 已带完整位置
			}
			return fmt.Errorf("%s: %w", pos, err)
		}
	}
	switch {
	case refs > 1:
		return fmt.Errorf("%s一条 Chain 最多一个 reference，实际 %d 个", at, refs)
	case buttons > 1:
		return fmt.Errorf("%s一条 Chain 最多一个 buttons，实际 %d 个", at, buttons)
	case content == 0:
		return fmt.Errorf("%sChain 只有 reference/buttons 等元数据，没有可发送内容", at)
	}
	return nil
}

func text(t *pb.TextSegment) error {
	s := t.GetText()
	if s == "" {
		return fmt.Errorf("text 为空")
	}
	if !utf8.ValidString(s) {
		return fmt.Errorf("text 不是合法 UTF-8")
	}
	if t.Format != nil {
		f := t.GetFormat()
		if f == pb.TextFormat_TEXT_FORMAT_UNSPECIFIED || !knownEnum(f) {
			return fmt.Errorf("format 取值非法: %d", int32(f))
		}
		if len(t.GetEntities()) > 0 {
			return fmt.Errorf("format 与 entities 不能同时指定")
		}
	}
	if len(t.GetEntities()) == 0 {
		return nil
	}
	// 只登记实体实际用到的起止偏移，扫描一遍文本判断它们是否落在字符边界上；
	// 额外内存与实体数成正比，与文本长度无关（文本最长可达帧上限 4 MiB）。
	boundary := make(map[int]bool, 2*len(t.GetEntities()))
	for _, e := range t.GetEntities() {
		start := int(e.GetOffsetUtf16())
		boundary[start] = false
		boundary[start+int(e.GetLengthUtf16())] = false
	}
	n := 0
	for _, r := range s {
		if _, ok := boundary[n]; ok {
			boundary[n] = true
		}
		n += utf16.RuneLen(r) // 已校验 UTF-8，不会返回 -1
	}
	if _, ok := boundary[n]; ok {
		boundary[n] = true
	}
	for i, e := range t.GetEntities() {
		if err := entity(e, n, boundary); err != nil {
			return fmt.Errorf("entities[%d]: %w", i, err)
		}
	}
	return nil
}

func entity(e *pb.TextEntity, utf16Len int, boundary map[int]bool) error {
	typ := e.GetType()
	if typ == pb.EntityType_ENTITY_TYPE_UNSPECIFIED || !knownEnum(typ) {
		return fmt.Errorf("不支持的实体类型 %d", int32(typ))
	}
	start, length := int(e.GetOffsetUtf16()), int(e.GetLengthUtf16())
	end := start + length
	switch {
	case length == 0:
		return fmt.Errorf("length_utf16 不能为 0")
	case end > utf16Len:
		return fmt.Errorf("范围 [%d,%d) 超出文本长度 %d（UTF-16）", start, end, utf16Len)
	case !boundary[start] || !boundary[end]:
		return fmt.Errorf("范围 [%d,%d) 截断了字符（代理对）", start, end)
	}
	isLink := typ == pb.EntityType_ENTITY_TYPE_TEXT_LINK
	isMention := typ == pb.EntityType_ENTITY_TYPE_TEXT_MENTION
	isPre := typ == pb.EntityType_ENTITY_TYPE_PRE
	switch {
	case isLink != (e.Url != nil):
		return fmt.Errorf("url 只能且必须用于 text_link")
	case isLink && !absoluteURL(e.GetUrl(), "https", "http"):
		return fmt.Errorf("text_link 的 url 不是 http(s) 绝对地址")
	case isMention != (e.ExternalUserId != nil) || (isMention && e.GetExternalUserId() == ""):
		return fmt.Errorf("external_user_id 只能且必须用于 text_mention")
	case e.Language != nil && !isPre:
		return fmt.Errorf("language 只能用于 pre")
	}
	return nil
}

func media(src *pb.MediaSource) error {
	switch s := src.GetSource().(type) {
	case *pb.MediaSource_MediaRef:
		if s.MediaRef == "" {
			return fmt.Errorf("media_ref 为空")
		}
	case *pb.MediaSource_IncomingRef:
		if s.IncomingRef.GetToken() == "" {
			return fmt.Errorf("incoming_ref 缺少 token")
		}
	case *pb.MediaSource_HttpsUrl:
		if !absoluteURL(s.HttpsUrl, "https") {
			return fmt.Errorf("https_url 不是 https 绝对地址")
		}
	default:
		return fmt.Errorf("媒体必须恰好有一个来源（media_ref、incoming_ref 或 https_url）")
	}
	return nil
}

func absoluteURL(raw string, schemes ...string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	for _, s := range schemes {
		if u.Scheme == s {
			return true
		}
	}
	return false
}

// filename 只是展示名称，不能带路径。
func filename(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return fmt.Errorf("filename %q 不是无路径的文件名", name)
	}
	return nil
}

func reference(r *pb.ReferenceSegment) error {
	if r.GetOriginalPlatformMessageId() == "" {
		return fmt.Errorf("reference 缺少 original_platform_message_id")
	}
	if t := r.GetTarget(); t != nil {
		if t.GetAdapterId() == "" || t.GetAdapterInstanceId() == "" || t.GetAccountId() == "" || t.GetChatId() == "" {
			return fmt.Errorf("reference.target 缺少 adapter/instance/account/chat 标识")
		}
	}
	return nil
}

func buttonRows(b *pb.ButtonsSegment) error {
	if len(b.GetRows()) == 0 {
		return fmt.Errorf("buttons 没有任何行")
	}
	for i, row := range b.GetRows() {
		if len(row.GetButtons()) == 0 {
			return fmt.Errorf("buttons 第 %d 行为空", i)
		}
		for j, btn := range row.GetButtons() {
			switch {
			case btn.GetActionId() == "":
				return fmt.Errorf("按钮 [%d][%d] 缺少 action_id", i, j)
			case btn.GetLabel() == "":
				return fmt.Errorf("按钮 [%d][%d] 缺少 label", i, j)
			case btn.Value != nil && !json.Valid([]byte(btn.GetValue())):
				return fmt.Errorf("按钮 [%d][%d] 的 value 不是合法 JSON", i, j)
			}
		}
	}
	return nil
}

func forward(f *pb.ForwardSegment, pos string) error {
	if len(f.GetItems()) == 0 {
		return fmt.Errorf("forward 没有任何条目")
	}
	for i, item := range f.GetItems() {
		if err := chain(item.GetChain(), fmt.Sprintf("%s.forward.items[%d].", pos, i)); err != nil {
			return err
		}
	}
	return nil
}
