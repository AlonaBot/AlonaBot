// commanddispatch 运行命令选择与参数绑定的集成测试插件。
package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const auditEnv = "ALONABOT_TEST_COMMAND_AUDIT"

type plugin struct{ sdk.Base }

func record(entry string) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(os.Getenv(auditEnv), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	_, writeError := file.Write(append(data, '\n'))
	return errors.Join(writeError, file.Close())
}

func reply(entry, text string) (message.Chain, error) {
	if err := record(entry); err != nil {
		return nil, err
	}
	return message.Chain{message.Text(text)}, nil
}

// Commands 注册有实际副作用记录的命令与文本路由。
func (*plugin) Commands(r *sdk.CommandRegistry) {
	r.Command("rss", func(*sdk.Context) (message.Chain, error) { return reply("rss", "rss") }, sdk.Alias("feed"))
	r.Group("rss", func(g *sdk.CommandGroup) {
		g.Command("add", func(_ *sdk.Context, url string) (message.Chain, error) { return reply("add:"+url, "added:"+url) }, sdk.Args("url"))
		g.Command("list", func(*sdk.Context) (message.Chain, error) { return reply("list", "list") })
	})
	r.Command("calc", func(_ *sdk.Context, a, b int64) (message.Chain, error) {
		return reply("calc", strconv.FormatInt(a+b, 10))
	}, sdk.Args("a", "b"))
	r.Command("quiet", func(_ *sdk.Context, n int64) error { return record("quiet:" + strconv.FormatInt(n, 10)) }, sdk.Args("n"), sdk.Silent())
	r.Command("say", func(_ *sdk.Context, tail string) (message.Chain, error) { return reply("say:"+tail, tail) }, sdk.RawTail())
	r.Command("quote", func(_ *sdk.Context, a, b string) (message.Chain, error) { return reply("quote", a+"|"+b) }, sdk.Args("a", "b"))
	r.Command("optional", func(_ *sdk.Context, n int64) (message.Chain, error) {
		return reply("optional", strconv.FormatInt(n, 10))
	}, sdk.Optional("arg1", "9"))
	r.Command("custom", func(_ *sdk.Context, n int64) error { return record("custom") }, sdk.Usage("custom usage"))
	r.Command("sendguard", func(c *sdk.Context) (message.Chain, error) {
		if err := record("sendguard"); err != nil {
			return nil, err
		}
		_, err := c.Send(message.Chain{message.Text("forbidden")})
		return nil, err
	})
	r.Command("barrier", func(c *sdk.Context) (message.Chain, error) {
		c.StopPropagation()
		return message.Chain{message.Text("barrier")}, nil
	})
	r.Regex("regex", "(hello)", func(c *sdk.Context) error { return record("regex:" + strings.Join(c.Match(), "|")) }, sdk.Priority(-10))
	r.Keyword("keyword", []string{"hello"}, func(c *sdk.Context) error {
		if err := record("keyword"); err != nil {
			return err
		}
		var text strings.Builder
		for _, segment := range c.Event().GetEventPayload().GetMessageCreated().GetChain().GetSegments() {
			text.WriteString(segment.GetText().GetText())
		}
		if strings.Contains(text.String(), "stop") {
			c.StopPropagation()
		}
		return nil
	}, sdk.Priority(10), sdk.IgnoreCase())
	r.StartsWith("starts", []string{"BEGIN"}, func(*sdk.Context) error { return record("starts") })
	r.EndsWith("ends", []string{"END"}, func(*sdk.Context) error { return record("ends") })
	r.FullMatch("full", []string{"hello", "WHOLE"}, func(*sdk.Context) error { return record("full") })
	r.Raw("raw", func(*sdk.Context) error { return record("raw") }, sdk.Priority(-100))
}

// Events 记录命令层继续后进入普通事件路由的实际调用。
func (*plugin) Events(r *sdk.EventRegistry) {
	r.On("event", "message.created", func(*sdk.Context) error { return record("event") })
}

func main() {
	if err := sdk.Run(&plugin{}); err != nil {
		slog.Error("command dispatch plugin failed", "error", err)
		os.Exit(1)
	}
}
