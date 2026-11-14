// echo 命令演示显式注册、参数、延续调用与类型化服务，并回送未消费的文本和图片。
package main

//go:generate go run ../../../cmd/alonabot plugin generate

import (
	"log/slog"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// New 为每个插件对象创建独立的笔记状态并绑定生成资源。
func New() *sdk.Plugin {
	r := sdk.NewPlugin()
	sdk.UseGenerated(r, alonabotGenerated)
	registerCommands(r)
	r.On("echo", sdk.MessageCreated).Rule(sdk.NotCommand()).Handle(func(c *sdk.Context, e *message.CreatedPayload) error {
		var chain message.Chain
		for _, segment := range e.Chain {
			switch segment := segment.(type) {
			case message.TextSegment:
				chain = append(chain, segment)
			case message.ImageSegment:
				chain = append(chain, segment)
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
