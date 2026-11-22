// Package message 提供 SDK 的消息、会话、事件与发送结果类型。
//
// [Chain] 按顺序组成一次逻辑发送，[Segment] 与 Chain 都实现 [Content]。
// 会话、消息句柄与用户身份使用 string，由 [SID]、[Ref]、[UID] 拼接，Parse 函数校验并拆分。
// [Event] 与 [SendResult] 提供关联消息的句柄。
// 对应的协议类型见 contracts/proto/alonabot/v1/message.proto 与 identity.proto。
package message

import (
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/markdown"
	"github.com/AlonaBot/AlonaBot/sdk/internal/messagecodec"
)

// Content 允许消息链或单个片段作为发送与自动回复内容。
type Content interface{ content() Chain }

// Segment 限定消息链中由本包提供的片段。
type Segment interface {
	Content
	segment()
}

// Chain 按顺序组成一次逻辑发送，空链表示没有内容。
type Chain []Segment

func (c Chain) content() Chain { return c }

// Text 依次拼接全部文本片段，跳过其他片段。
func (c Chain) Text() string {
	var out strings.Builder
	for _, s := range c {
		switch text := s.(type) {
		case TextSegment:
			out.WriteString(text.Text)
		case MarkdownSegment:
			out.WriteString(text.Content)
		}
	}
	return out.String()
}

// TextFormat 指定文本的标记格式，零值表示未指定。
type TextFormat int

const (
	FormatPlain    TextFormat = iota + 1 // 普通文本。
	FormatMarkdown                       // Markdown 文本。
)

// EntityType 指定富文本实体的语义。
type EntityType int

const (
	EntityBold          EntityType = iota + 1 // 粗体。
	EntityItalic                              // 斜体。
	EntityUnderline                           // 下划线。
	EntityStrikethrough                       // 删除线。
	EntitySpoiler                             // 剧透遮盖。
	EntityCode                                // 行内代码。
	EntityPre                                 // 代码块。
	EntityTextLink                            // 文本链接。
	EntityTextMention                         // 文本提及。
)

// TextEntity 指定 UTF-16 码元区间及该实体需要的可选属性。
type TextEntity struct {
	Type                          EntityType
	OffsetUTF16, LengthUTF16      uint32
	URL, Language, ExternalUserID string
}

// TextSegment 保存普通文本或带实体、标记格式的文本。
type TextSegment struct {
	Text     string
	Format   TextFormat
	Entities []TextEntity
}

// ImageSegment 保存图片来源及可选说明。
type ImageSegment struct {
	Source       MediaSource
	Alt, Caption string
}

// VideoSegment 保存视频来源，Duration 为零表示未知。
type VideoSegment struct {
	Source   MediaSource
	Duration time.Duration
	Caption  string
}

// VoiceSegment 保存语音来源，Duration 为零表示未知。
type VoiceSegment struct {
	Source   MediaSource
	Duration time.Duration
	Caption  string
}

// FileSegment 保存文件来源和展示元数据，Filename 不包含路径。
type FileSegment struct {
	Source                  MediaSource
	Filename, MIME, Caption string
}

// MentionSegment 指定平台用户，Display 为空时由平台选择展示文本。
type MentionSegment struct{ ExternalUserID, Display string }

type referenceOrigin = messagecodec.ReferenceOrigin

// ReferenceSegment 通过句柄引用消息，实例省略时由 Core 补齐。
type ReferenceSegment struct {
	referenceOrigin
	Ref string
}

// ForwardSource 保存转发条目的可选来源展示信息。
type ForwardSource struct{ DisplayName, ExternalUserID string }

// ForwardItem 保存一条转发的来源与消息内容。
type ForwardItem struct {
	OccurredAt        time.Time // 零值表示平台未提供时间。
	Source            *ForwardSource
	Chain             Chain
	PlatformMessageID string // 与 Chain 恰有一个，引用同一会话的已有消息。
}

// ForwardSegment 保存有序转发条目，或入站的平台转发标识。
type ForwardSegment struct {
	Items             []ForwardItem
	PlatformForwardID string // 只在入站读取时出现。
}

// Text 构造普通文本片段。
func Text(s string) Segment { return TextSegment{Text: s} }

// ImageURL 构造 HTTPS 图片片段。
func ImageURL(httpsURL string) Segment { return Image(MediaSource{HTTPSURL: httpsURL}) }

// VideoURL 构造 HTTPS 视频片段。
func VideoURL(httpsURL string) Segment { return Video(MediaSource{HTTPSURL: httpsURL}) }

// FileURL 构造 HTTPS 文件片段。
func FileURL(httpsURL, filename string) Segment {
	return File(MediaSource{HTTPSURL: httpsURL}, filename)
}

// Image 构造指定来源的图片片段。
func Image(src MediaSource) Segment { return ImageSegment{Source: src} }

// Video 构造指定来源的视频片段。
func Video(src MediaSource) Segment { return VideoSegment{Source: src} }

// File 构造指定来源和展示名称的文件片段。
func File(src MediaSource, filename string) Segment {
	return FileSegment{Source: src, Filename: filename}
}

// Mention 构造用户提及片段。
func Mention(externalUserID string) Segment { return MentionSegment{ExternalUserID: externalUserID} }

// Quote 构造引用指定消息的片段。
func Quote(ref string) Segment { return ReferenceSegment{Ref: ref} }

func (s TextSegment) segment()       {}
func (s TextSegment) content() Chain { return Chain{s} }

func (s ImageSegment) segment()       {}
func (s ImageSegment) content() Chain { return Chain{s} }

func (s VideoSegment) segment()       {}
func (s VideoSegment) content() Chain { return Chain{s} }

func (s FileSegment) segment()       {}
func (s FileSegment) content() Chain { return Chain{s} }

func (s MentionSegment) segment()       {}
func (s MentionSegment) content() Chain { return Chain{s} }

func (s ReferenceSegment) segment()       {}
func (s ReferenceSegment) content() Chain { return Chain{s} }

func (s ForwardSegment) segment()       {}
func (s ForwardSegment) content() Chain { return Chain{s} }

func (s ButtonsSegment) segment()       {}
func (s ButtonsSegment) content() Chain { return Chain{s} }

func (s ExtensionSegment) segment()       {}
func (s ExtensionSegment) content() Chain { return Chain{s} }

// Voice 构造指定来源的语音片段。
func Voice(src MediaSource) Segment { return VoiceSegment{Source: src} }

// VoiceURL 构造 HTTPS 语音片段。
func VoiceURL(httpsURL string) Segment { return Voice(MediaSource{HTTPSURL: httpsURL}) }

// Forward 按顺序构造合并转发片段。
func Forward(items ...ForwardItem) Segment { return ForwardSegment{Items: items} }

func (s VoiceSegment) segment()       {}
func (s VoiceSegment) content() Chain { return Chain{s} }

// MarkdownSegment 保存通用 CommonMark 与删除线正文。
type MarkdownSegment struct{ Content string }

// Markdown 构造通用 Markdown 片段。
func Markdown(content string) MarkdownSegment { return MarkdownSegment{Content: content} }

// PlainText 显式转换为保留段落、列表、引用和链接的普通文本。
func (m MarkdownSegment) PlainText() TextSegment {
	return TextSegment{Text: markdown.PlainText(m.Content)}
}
func (m MarkdownSegment) segment()       {}
func (m MarkdownSegment) content() Chain { return Chain{m} }
