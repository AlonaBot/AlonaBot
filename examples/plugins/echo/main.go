// echo 命令运行一个回复 message.created 文本片段的 Native 插件。
package main

//go:generate go run ../../../cmd/alonabot plugin generate

import (
	"log/slog"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// New 构造手动回复事件的插件对象。
func New() *sdk.Plugin {
	r := sdk.NewPlugin()
	sdk.UseGenerated(r, alonabotGenerated)
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
	return r
}

func main() {
	if err := sdk.Run(New()); err != nil {
		slog.Error("echo plugin stopped with an error", "error", err)
		os.Exit(1)
	}
}
