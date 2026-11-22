package message

import (
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

// SendStatus 指定一次逻辑发送的整体结果。
type SendStatus int

const (
	SendAccepted SendStatus = iota + 1 // 全部已接受。
	SendPartial                        // 部分已接受。
	SendRejected                       // 明确拒绝。
	SendUnknown                        // 平台结果未知。
)

// PartStatus 指定一次物理平台操作的结果。
type PartStatus int

const (
	PartAccepted     PartStatus = iota + 1 // 已接受。
	PartRejected                           // 明确拒绝。
	PartUnknown                            // 结果未知。
	PartNotAttempted                       // 前序失败而未尝试。
)

// PartResult 保存物理操作与原片段的对应关系和完整回执。
type PartResult struct {
	PartIndex             uint32
	SegmentIndexes        []uint32
	Status                PartStatus
	PlatformMessageIDs    []string
	MessageIDsUnavailable bool
	Error                 *errcode.Error
}

// SendResult 保存一次逻辑发送的已确认消息及各操作结果。
type SendResult struct {
	Status                   SendStatus
	Session                  string
	AdapterInstanceID        string
	PlatformMessageIDs       []string
	PrimaryPlatformMessageID string
	ProviderAcceptedAt       time.Time
	CapabilitiesUsed         []Capability
	PartResults              []PartResult
}

// Ref 返回首个已确认消息的句柄，缺少消息 ID 时返回空串。
func (r *SendResult) Ref() string {
	if r.PrimaryPlatformMessageID == "" || r.Session == "" {
		return ""
	}
	return Ref(r.Session, r.PrimaryPlatformMessageID, r.AdapterInstanceID)
}

// Refs 按平台回执顺序返回全部已确认消息的句柄。
func (r *SendResult) Refs() []string {
	var refs []string
	for _, id := range r.PlatformMessageIDs {
		refs = append(refs, Ref(r.Session, id, r.AdapterInstanceID))
	}
	return refs
}
