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
	r.On("echo", sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error {
		var chain message.Chain
		for _, segment := range e.Chain {
			if text, ok := segment.(message.TextSegment); ok {
				chain = append(chain, text)
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
