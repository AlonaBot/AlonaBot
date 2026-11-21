package validate

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

const maxInlineMediaBytes = 1 << 20 // CONTRACTS §1.3：内联来源最多 1 MiB。

// Chain 检查 MessageChain 的结构（MESSAGE-MODEL §2）。只做与平台无关的结构检查；
// 平台能力（UNSUPPORTED）、媒体租约与目标授权由 Core/Adapter 在发送前另行检查。
// 结构错误返回 INVALID_ARGUMENT，转发节点中无法表达的片段返回 UNSUPPORTED；消息保留片段位置。
func Chain(c *pb.MessageChain) error {
	if err := chain(c, "", false); err != nil {
		code := pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		var validation *Error
		if errors.As(err, &validation) {
			code = validation.Code
		}
		return &Error{Code: code, Message: err.Error()}
	}
	return nil
}

// chain 校验一条 Chain；at 是嵌套位置前缀（顶层为空），用于组成错误中的 segment 位置。
func chain(c *pb.MessageChain, at string, incoming bool) error {
	// 1. 逐个校验片段，并统计 reference/buttons 与可发送内容的数量。
	segs := c.GetSegments()
	if len(segs) == 0 {
		return fmt.Errorf("%schain is empty", at)
	}
	var refs, buttons, content int
	for i, s := range segs {
		pos := fmt.Sprintf("%ssegment[%d]", at, i)
		var err error
		switch d := s.GetData().(type) {
		case *pb.Segment_Markdown:
			content++
			if d.Markdown.GetContent() == "" || !utf8.ValidString(d.Markdown.GetContent()) {
				err = fmt.Errorf("markdown content must be nonempty UTF-8")
			}
		case *pb.Segment_Text:
			content++
			err = text(d.Text)
		case *pb.Segment_Image:
			content++
			err = media(d.Image.GetSource())
		case *pb.Segment_Video:
			content++
			err = media(d.Video.GetSource())
		case *pb.Segment_Voice:
			content++
			err = media(d.Voice.GetSource())
		case *pb.Segment_File:
			content++
			if err = media(d.File.GetSource()); err == nil && d.File.Filename != nil {
				err = filename(d.File.GetFilename())
			}
		case *pb.Segment_Mention:
			content++
			if d.Mention.GetExternalUserId() == "" {
				err = fmt.Errorf("mention is missing external_user_id")
			}
		case *pb.Segment_Reference:
			refs++
			err = reference(d.Reference)
		case *pb.Segment_Buttons:
			buttons++
			err = buttonRows(d.Buttons)
		case *pb.Segment_Forward:
			content++
			err = forward(d.Forward, pos, incoming)
		case *pb.Segment_Extension:
			content++
			err = extension(d.Extension)
		default:
			err = fmt.Errorf("segment has no type set")
		}
		if err != nil {
			if strings.HasPrefix(err.Error(), pos) {
				return err // 嵌套 Chain 已带完整位置
			}
			return fmt.Errorf("%s: %w", pos, err)
		}
	}
	// 2. 检查整条 Chain 的组合约束。
	switch {
	case refs > 1:
		return fmt.Errorf("%sa chain allows at most one reference, got %d", at, refs)
	case buttons > 1:
		return fmt.Errorf("%sa chain allows at most one buttons segment, got %d", at, buttons)
	case content == 0:
		return fmt.Errorf("%schain has only metadata such as reference/buttons and nothing to send", at)
	}
	return nil
}

func extension(e *pb.ExtensionSegment) error {
	switch {
	case e.GetAdapterId() == "":
		return fmt.Errorf("extension is missing adapter_id")
	case e.GetName() == "":
		return fmt.Errorf("extension is missing name")
	case e.GetVersion() == 0:
		return fmt.Errorf("extension is missing version")
	case !IsJSONObject([]byte(e.GetPayloadJson())):
		return fmt.Errorf("extension payload_json is not a JSON object")
	}
	return nil
}

// text 校验文本片段：非空 UTF-8、format 与 entities 互斥，以及每个实体的范围与类型约束。
func text(t *pb.TextSegment) error {
	// 1. 文本本身与 format。
	s := t.GetText()
	if s == "" {
		return fmt.Errorf("text is empty")
	}
	if !utf8.ValidString(s) {
		return fmt.Errorf("text is not valid UTF-8")
	}
	if t.Format != nil {
		f := t.GetFormat()
		if f == pb.TextFormat_TEXT_FORMAT_UNSPECIFIED || !knownEnum(f) {
			return fmt.Errorf("invalid format value: %d", int32(f))
		}
		if len(t.GetEntities()) > 0 {
			return fmt.Errorf("format and entities cannot both be set")
		}
	}
	if len(t.GetEntities()) == 0 {
		return nil
	}
	// 2. 只登记实体实际用到的起止偏移，扫描一遍文本判断它们是否落在字符边界上；
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
	// 3. 逐个校验实体。
	for i, e := range t.GetEntities() {
		if err := entity(e, n, boundary); err != nil {
			return fmt.Errorf("entities[%d]: %w", i, err)
		}
	}
	return nil
}

// entity 校验单个实体：类型已知、范围不越界且落在字符边界（boundary 由 text 计算），
// url/external_user_id/language 只出现在对应类型上。
func entity(e *pb.TextEntity, utf16Len int, boundary map[int]bool) error {
	typ := e.GetType()
	if typ == pb.EntityType_ENTITY_TYPE_UNSPECIFIED || !knownEnum(typ) {
		return fmt.Errorf("unsupported entity type %d", int32(typ))
	}
	start, length := int(e.GetOffsetUtf16()), int(e.GetLengthUtf16())
	end := start + length
	switch {
	case length == 0:
		return fmt.Errorf("length_utf16 must not be 0")
	case end > utf16Len:
		return fmt.Errorf("range [%d,%d) exceeds text length %d (UTF-16)", start, end, utf16Len)
	case !boundary[start] || !boundary[end]:
		return fmt.Errorf("range [%d,%d) splits a character (surrogate pair)", start, end)
	}
	isLink := typ == pb.EntityType_ENTITY_TYPE_TEXT_LINK
	isMention := typ == pb.EntityType_ENTITY_TYPE_TEXT_MENTION
	isPre := typ == pb.EntityType_ENTITY_TYPE_PRE
	switch {
	case isLink != (e.Url != nil):
		return fmt.Errorf("url is allowed and required only for text_link")
	case isLink && !absoluteURL(e.GetUrl(), "https", "http"):
		return fmt.Errorf("text_link url is not an absolute http(s) URL")
	case isMention != (e.ExternalUserId != nil) || (isMention && e.GetExternalUserId() == ""):
		return fmt.Errorf("external_user_id is allowed and required only for text_mention")
	case e.Language != nil && !isPre:
		return fmt.Errorf("language is allowed only for pre")
	}
	return nil
}

// media 校验媒体来源恰好设置一种且内容非空，https_url 必须是 https 绝对地址。
func media(src *pb.MediaSource) error {
	switch s := src.GetSource().(type) {
	case *pb.MediaSource_MediaRef:
		if s.MediaRef == "" {
			return fmt.Errorf("media_ref is empty")
		}
	case *pb.MediaSource_IncomingRef:
		if s.IncomingRef.GetToken() == "" {
			return fmt.Errorf("incoming_ref is missing token")
		}
	case *pb.MediaSource_HttpsUrl:
		if !absoluteURL(s.HttpsUrl, "https") {
			return fmt.Errorf("https_url is not an absolute https URL")
		}
	case *pb.MediaSource_InlineBase64:
		data, err := base64.StdEncoding.Strict().DecodeString(s.InlineBase64)
		if err != nil || len(data) == 0 || len(data) > maxInlineMediaBytes || base64.StdEncoding.EncodeToString(data) != s.InlineBase64 {
			return fmt.Errorf("inline_base64 must be standard base64 encoding of 1 B to 1 MiB")
		}
	case *pb.MediaSource_PlatformFile:
		return fmt.Errorf("platform_file is reserved for Core")
	default:
		return fmt.Errorf("media must have exactly one source (media_ref, incoming_ref, https_url or inline_base64)")
	}
	return nil
}

// absoluteURL 报告 raw 是否为带主机名、协议属于 schemes 之一的绝对 URL。
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
		return fmt.Errorf("filename %q is not a bare file name", name)
	}
	return nil
}

// reference 校验回复引用：必须有原消息 ID，给出 target 时四项标识齐全。
func reference(r *pb.ReferenceSegment) error {
	if r.GetOriginalPlatformMessageId() == "" {
		return fmt.Errorf("reference is missing original_platform_message_id")
	}
	if t := r.GetTarget(); t != nil {
		if t.GetAdapterId() == "" || t.GetAdapterInstanceId() == "" || t.GetAccountId() == "" || t.GetChatId() == "" {
			return fmt.Errorf("reference.target is missing adapter/instance/account/chat identifiers")
		}
	}
	return nil
}

// buttonRows 校验键盘结构、动作与 URL 互斥，以及 JSON 值。
func buttonRows(b *pb.ButtonsSegment) error {
	if len(b.GetRows()) == 0 {
		return fmt.Errorf("buttons has no rows")
	}
	for i, row := range b.GetRows() {
		if len(row.GetButtons()) == 0 {
			return fmt.Errorf("buttons row %d is empty", i)
		}
		for j, btn := range row.GetButtons() {
			switch {
			case btn == nil:
				return fmt.Errorf("button [%d][%d] is missing", i, j)
			case (btn.GetActionId() != "") == (btn.Url != nil):
				return fmt.Errorf("button [%d][%d] must have exactly one action_id or url", i, j)
			case btn.GetLabel() == "":
				return fmt.Errorf("button [%d][%d] is missing label", i, j)
			case btn.Url != nil:
				u, err := url.Parse(btn.GetUrl())
				if err != nil || !u.IsAbs() || u.Host == "" || (u.Scheme != "https" && u.Scheme != "tg") {
					return fmt.Errorf("button [%d][%d] url must be an absolute https or tg link", i, j)
				}
				if btn.Value != nil {
					return fmt.Errorf("URL button [%d][%d] cannot have value", i, j)
				}
			case btn.Value != nil && !json.Valid([]byte(btn.GetValue())):
				return fmt.Errorf("value of button [%d][%d] is not valid JSON", i, j)
			}
		}
	}
	return nil
}

// forward 校验出站节点；平台转发标识只供入站读取。
func forward(f *pb.ForwardSegment, pos string, incoming bool) error {
	if !incoming && f != nil && f.PlatformForwardId != nil {
		return fmt.Errorf("platform_forward_id is only allowed inbound")
	}
	if incoming && f.GetPlatformForwardId() != "" && len(f.GetItems()) == 0 {
		return nil
	}
	if len(f.GetItems()) == 0 {
		return fmt.Errorf("forward has no items")
	}
	for i, item := range f.GetItems() {
		at := fmt.Sprintf("%s.forward.items[%d].", pos, i)
		if item == nil {
			return fmt.Errorf("%snode is missing", at)
		}
		if (item.GetChain() != nil) == (item.PlatformMessageId != nil) {
			return fmt.Errorf("%snode must have exactly one chain or platform_message_id", at)
		}
		if item.PlatformMessageId != nil {
			if item.GetPlatformMessageId() == "" {
				return fmt.Errorf("%splatform_message_id is empty", at)
			}
			continue
		}
		if err := chain(item.GetChain(), at, incoming); err != nil {
			return err
		}
	}
	return nil
}

// IncomingChain 校验平台返回的递归链，允许平台合并转发资源标识。
func IncomingChain(c *pb.MessageChain) error { return chain(c, "", true) }

// ReadNodes 仅接受平台查询标识，防止将调用者构造的媒体来源重新签发为平台授权。
func ReadNodes(f *pb.ForwardSegment) error {
	if f != nil && f.PlatformForwardId != nil {
		if f.GetPlatformForwardId() == "" || len(f.Items) != 0 {
			return fmt.Errorf("node query requires a nonempty resource ID without inline items")
		}
		return nil
	}
	for index, item := range f.GetItems() {
		if item.GetChain() != nil {
			return fmt.Errorf("segment.forward.items[%d]: node query requires platform_message_id; traverse expanded chains directly", index)
		}
	}
	return forward(f, "segment", true)
}
