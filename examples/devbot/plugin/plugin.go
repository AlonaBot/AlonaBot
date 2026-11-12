// Package plugin 提供库模式示例的消息回声插件。
package plugin

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

//go:generate go run ../../../cmd/alonabot plugin generate

type echo struct{ sdk.Base }

// New 创建独立的回声插件对象。
func New() sdk.Root { return &echo{} }

func (*echo) Events(r *sdk.EventRegistry) {
	r.On("echo", sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error {
		_, err := c.Send(e.Chain)
		return err
	})
}
