// commanddispatch 运行命令选择与参数绑定的集成测试插件。
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	auditEnv             = "ALONABOT_TEST_COMMAND_AUDIT"
	serviceInputSchema   = "test.command.service.input.v1"
	serviceOutputSchema  = "test.command.service.output.v1"
	serviceSchemaVersion = 1
	mediaReadBytes       = 1024
)

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

// readImage 同时检查参数与 prepared Event 的同源内容，并用当前 Session 读取实际媒体。
func readImage(c *sdk.Context, image *pb.ImageSegment) ([]byte, error) {
	matched := false
	for _, segment := range c.Event().GetEventPayload().GetMessageCreated().GetChain().GetSegments() {
		if proto.Equal(segment.GetImage(), image) {
			matched = true
			break
		}
	}
	if !matched {
		return nil, errors.New("image argument differs from prepared event")
	}
	data, next, eof, err := sdk.ReadMedia(c.Context(), image.GetSource().GetIncomingRef().GetToken(), 0, mediaReadBytes)
	if err != nil {
		return nil, err
	}
	if !eof || next != uint64(len(data)) {
		return nil, fmt.Errorf("unexpected media read: next=%d eof=%v", next, eof)
	}
	return data, nil
}

// Commands 注册有实际副作用记录的命令与文本路由。
func (*plugin) Commands(r *sdk.CommandRegistry) {
	r.Command("service", func(c *sdk.Context) (message.Chain, error) {
		caller, err := sdk.CallJSON[string](c.Context(), "demo/context", serviceInputSchema, serviceSchemaVersion, "service reply")
		if err != nil {
			return nil, err
		}
		return reply("service", caller)
	})
	r.Command("rss", func(*sdk.Context) (message.Chain, error) { return reply("rss", "rss") }, sdk.Alias("feed"))
	r.Group("rss", func(g *sdk.CommandGroup) {
		g.Command("add", func(_ *sdk.Context, url string) (message.Chain, error) { return reply("add:"+url, "added:"+url) }, sdk.Args("url"))
		g.Command("guarded", func(_ *sdk.Context, n int64) (message.Chain, error) {
			return reply("guarded", strconv.FormatInt(n, 10))
		}, sdk.Args("n"))
		g.Command("list", func(*sdk.Context) (message.Chain, error) { return reply("list", "list") })
	})
	r.Command("permissions", func(c *sdk.Context) (message.Chain, error) {
		return reply("permissions", strings.Join(c.Event().GetActor().GetPermissions(), ","))
	})
	r.Command("calc", func(_ *sdk.Context, a, b int64) (message.Chain, error) {
		return reply("calc", strconv.FormatInt(a+b, 10))
	}, sdk.Args("a", "b"))
	r.Command("kick", func(_ *sdk.Context, user sdk.UserID) (message.Chain, error) {
		return reply("kick:"+string(user), string(user))
	}, sdk.Args("user"))
	r.Command("picture", func(c *sdk.Context, image *pb.ImageSegment) (message.Chain, error) {
		data, err := readImage(c, image)
		if err != nil {
			return nil, err
		}
		return reply("picture", string(data))
	}, sdk.Args("image"))
	r.Command("relay", func(c *sdk.Context, body *pb.MessageChain) (message.Chain, error) {
		for _, segment := range body.GetSegments() {
			if image := segment.GetImage(); image != nil {
				if _, err := readImage(c, image); err != nil {
					return nil, err
				}
			}
		}
		encoded, err := protojson.Marshal(body)
		if err != nil {
			return nil, err
		}
		return reply("relay", string(encoded))
	}, sdk.Args("body"))
	r.Command("quiet", func(_ *sdk.Context, n int64) error { return record("quiet:" + strconv.FormatInt(n, 10)) }, sdk.Args("n"), sdk.Silent())
	r.Command("say", func(_ *sdk.Context, tail string) (message.Chain, error) { return reply("say:"+tail, tail) }, sdk.RawTail())
	r.Command("quote", func(_ *sdk.Context, a, b string) (message.Chain, error) { return reply("quote", a+"|"+b) }, sdk.Args("a", "b"))
	r.Command("optional", func(_ *sdk.Context, n int64) (message.Chain, error) {
		return reply("optional", strconv.FormatInt(n, 10))
	}, sdk.Optional("arg1", "9"))
	r.Command("custom", func(_ *sdk.Context, n int64) error { return record("custom") }, sdk.UsageKey("command.custom.usage"))
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

// Services 通过原 Target 回复，验证命令派生的 Service 调用保留可信上下文。
func (*plugin) Services(r *sdk.ServiceRegistry) {
	sdk.JSONService(r, "context", serviceInputSchema, serviceOutputSchema, serviceSchemaVersion, func(c *sdk.Context, text string) (string, error) {
		_, err := c.Send(message.Chain{message.Text(text)})
		return c.Caller(), err
	})
}

// Events 记录命令层继续后进入普通事件路由的实际调用。
func (*plugin) Events(r *sdk.EventRegistry) {
	r.On("event", "message.created", func(*sdk.Context) error { return record("event") })
	r.On("restricted.event", "member.joined", func(*sdk.Context) error { return record("restricted.event") }, sdk.Priority(-1))
	r.On("observer", "member.joined", func(c *sdk.Context) error {
		return record("observer:" + strings.Join(c.Event().GetActor().GetPermissions(), ","))
	})

}

func main() {
	if err := sdk.Run(&plugin{}); err != nil {
		slog.Error("command dispatch plugin failed", "error", err)
		os.Exit(1)
	}
}
