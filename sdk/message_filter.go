package sdk

import (
	"errors"
	"fmt"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// MessageFilter 声明 Core 在交付消息事件前按 OR 匹配的条件。
type MessageFilter struct {
	ToMe            bool               // 匹配发给本机器人的消息。
	ChatTypes       []message.ChatType // 匹配任一列出的会话类型。
	WakeWordsConfig string             // 字符串列表字段的点分路径。
	TriggersConfig  string             // 触发方式多选字段的点分路径；替代 ToMe 和 ChatTypes，并要求 WakeWordsConfig。
}

// Filter 在 Handle 前声明一次 Core 消息过滤；它复制 ChatTypes，保留独立执行的 SDK Rule。
func (b *EventBuilder) Filter(filter MessageFilter) *EventBuilder {
	b.plugin.checkRegistration()
	if b.handles != 0 {
		b.afterHandle = true
	}
	if b.route.filter != nil {
		b.filterAgain = true
		return b
	}
	filter.ChatTypes = slices.Clone(filter.ChatTypes)
	b.route.filter = &filter
	return b
}

func (f *MessageFilter) check() error {
	if f == nil {
		return nil
	}
	if !f.ToMe && len(f.ChatTypes) == 0 && f.WakeWordsConfig == "" && f.TriggersConfig == "" {
		return errors.New("empty message filter")
	}
	for _, chatType := range f.ChatTypes {
		switch chatType {
		case message.ChatPrivate, message.ChatGroup, message.ChatChannel:
		default:
			return fmt.Errorf("invalid message filter chat type %d", chatType)
		}
	}
	return nil
}

func (f *MessageFilter) proto() *pb.MessageFilter {
	if f == nil {
		return nil
	}
	out := &pb.MessageFilter{ToMe: f.ToMe, WakeWordsConfig: f.WakeWordsConfig, TriggersConfig: f.TriggersConfig}
	for _, chatType := range f.ChatTypes {
		out.ChatTypes = append(out.ChatTypes, pb.ChatType(chatType))
	}
	return out
}
