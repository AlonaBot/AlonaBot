package sdk

import "github.com/AlonaBot/AlonaBot/sdk/message"

// ChatType 优先返回本次可信事件的会话类型；没有事件时读取继承的调用上下文，
// 没有真实入站事件时返回 ChatUnspecified。
func (c *Context) ChatType() message.ChatType {
	if c.event != nil {
		return message.ChatType(c.event.GetChatType())
	}
	return message.ChatType(c.chatType)
}

// IsPrivate 报告本次调用是否来自私聊。
func (c *Context) IsPrivate() bool { return c.ChatType() == message.ChatPrivate }

// IsGroup 报告本次调用是否来自群聊。
func (c *Context) IsGroup() bool { return c.ChatType() == message.ChatGroup }

// IsChannel 报告本次调用是否来自频道。
func (c *Context) IsChannel() bool { return c.ChatType() == message.ChatChannel }
