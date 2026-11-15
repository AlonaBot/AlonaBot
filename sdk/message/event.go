package message

import (
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

// Actor 保存可信触发者及 Core 计算的权限集合。
type Actor struct {
	ExternalUserID, AdapterInstanceID, DisplayName, Locale string
	Permissions                                            []string
}

// Capability 保存平台公布的能力与版本。
type Capability struct {
	Name      string
	Version   uint32
	Supported bool
	TextLimit *TextLimit // nil 表示未声明文本上限。
}

// TextUnit 指定文本限额的计量单位。
type TextUnit int

const (
	TextUTF16CodeUnits    TextUnit = iota + 1 // TextUTF16CodeUnits 按 UTF-16 码元计数。
	TextUnicodeCodePoints                     // TextUnicodeCodePoints 按 Unicode 码点计数。
	TextBytes                                 // TextBytes 按 UTF-8 字节计数。
)

// TextLimit 声明单条文本的正整数上限与计量单位。
type TextLimit struct {
	Max  uint64
	Unit TextUnit
}

// SourceKind 指定事件的来源类别。
type SourceKind int

const (
	SourceAdapter   SourceKind = iota + 1 // 平台 Adapter 来源。
	SourceScheduler                       // 调度器来源。
)

// ChatType 描述事件所在的会话类别。
type ChatType int

const (
	ChatUnspecified ChatType = iota // 平台未提供会话类型。
	ChatPrivate                     // 私聊。
	ChatGroup                       // 群聊。
	ChatChannel                     // 频道。
)

// PlatformMessage 保存与事件关联的平台消息。
type PlatformMessage struct {
	PlatformMessageID string
	Chain             Chain
	MediaGroupID      string
}

// Payload 限定标准事件和 Adapter 事件的载荷类型。
type Payload interface{ payload() }

// Event 保存一次入站事件；Handler 只读取本次调用的独立副本。
type Event struct {
	ID, Type                                string
	Source                                  SourceKind
	OccurredAt                              time.Time
	AdapterID, AdapterInstanceID, AccountID string
	Session                                 string
	ChatType                                ChatType
	ChatTitle                               string // 平台给出的会话显示名；空串表示未提供。
	Actor                                   *Actor
	Message                                 *PlatformMessage
	Payload                                 Payload
	Capabilities                            []Capability
	PlatformPayload                         []byte
	SourceEventID                           string
}

// Ref 返回创建、编辑、回应或按钮所在消息的句柄；其他事件返回空串。
func (e *Event) Ref() string {
	var id string
	switch p := e.Payload.(type) {
	case *CreatedPayload:
		id = p.PlatformMessageID
	case *EditedPayload:
		id = p.PlatformMessageID
	case *ReactionAddedPayload:
		id = p.PlatformMessageID
	case *ButtonClickedPayload:
		id = p.OriginPlatformMessageID
	}
	if id == "" || e.Session == "" {
		return ""
	}
	return Ref(e.Session, id, e.AdapterInstanceID)
}

// CreatedPayload 保存新消息内容及是否指向本机器人。
type CreatedPayload struct {
	PlatformMessageID string
	Chain             Chain
	ToMe              bool
}

// EditedPayload 保存新旧消息内容，OldChain 为 nil 表示平台未提供。
type EditedPayload struct {
	PlatformMessageID string
	Chain, OldChain   Chain
}

// DeletedPayload 保存已删除的平台消息 ID。
type DeletedPayload struct{ PlatformMessageIDs []string }

// ButtonClickedPayload 保存 Core 验证后的按钮绑定与平台回调信息。
type ButtonClickedPayload struct {
	ActionID                                        string
	ActionSchemaVersion                             uint32
	Value, CallbackQueryID, OriginPlatformMessageID string
}

// MemberPayload 保存加入或离开的成员，通过 Event.Type 区分动作。
type MemberPayload struct{ MemberExternalUserID string }

// ReactionAddedPayload 保存被回应的消息及回应内容。
type ReactionAddedPayload struct{ PlatformMessageID, Reaction string }

// ScheduledPayload 保存任务及本次运行的计划时间。
type ScheduledPayload struct {
	JobID, RunID string
	ScheduledAt  time.Time
	TaskKey      string // 动态任务实例的 key；静态 job 为空串。
}

// BotOnlinePayload 保存上线实例的账号信息。
type BotOnlinePayload struct{ AdapterID, AdapterInstanceID, AccountID, AccountDisplayName string }

// OfflineReason 指定实例离线的原因。
type OfflineReason int

const (
	OfflineStopped        OfflineReason = iota + 1 // 主动停止。
	OfflineAdapterExited                           // Adapter 进程退出。
	OfflineInstanceFailed                          // 实例报告致命故障。
)

// BotOfflinePayload 保存离线原因及可选安全错误。
type BotOfflinePayload struct {
	AdapterID, AdapterInstanceID, AccountID string
	Reason                                  OfflineReason
	SafeError                               *errcode.Error
}

// AdapterEventPayload 保存 Adapter 扩展事件的版本及 JSON 原文。
type AdapterEventPayload struct {
	Version     uint32
	PayloadJSON string
}

func (*CreatedPayload) payload() {}

func (*EditedPayload) payload() {}

func (*DeletedPayload) payload() {}

func (*ButtonClickedPayload) payload() {}

func (*MemberPayload) payload() {}

func (*ReactionAddedPayload) payload() {}

func (*ScheduledPayload) payload() {}

func (*BotOnlinePayload) payload() {}

func (*BotOfflinePayload) payload() {}

func (*AdapterEventPayload) payload() {}
