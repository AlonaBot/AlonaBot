// echo 命令运行一个回复 message.created 文本片段的 Native 插件。
package main

import (
	"log/slog"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

type echo struct{ sdk.Base }

// Events 注册手动回复事件 Handler；无文本片段时不发送。
func (*echo) Events(r *sdk.EventRegistry) {
	r.On("echo", "message.created").Handle(func(c *sdk.Context) error {
		var chain message.Chain
		for _, segment := range c.Event().GetEventPayload().GetMessageCreated().GetChain().GetSegments() {
			if text := segment.GetText(); text != nil {
				chain = append(chain, message.Text(text.GetText()))
			}
		}
		if len(chain) == 0 {
			return nil
		}
		_, err := c.Send(chain)
		return err
	})
}

func main() {
	if err := sdk.Run(&echo{}); err != nil {
		slog.Error("echo plugin stopped with an error", "error", err)
		os.Exit(1)
	}
}
