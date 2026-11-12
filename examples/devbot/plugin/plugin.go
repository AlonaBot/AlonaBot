// Package plugin 提供库模式示例的消息回声插件。
package plugin

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

//go:generate go run ../../../cmd/alonabot plugin generate

// New 创建独立插件，绑定生成资源并登记回声路由。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()
	sdk.UseGenerated(p, alonabotGenerated)
	p.On("echo", sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error {
		_, err := c.Send(e.Chain)
		return err
	})
	return p
}
