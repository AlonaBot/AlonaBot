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
	r.Command("service").Handle(func(c *sdk.Context) (message.Chain, error) {
		caller, err := sdk.CallJSON[string](c.Context(), "demo/context", serviceInputSchema, serviceSchemaVersion, "service reply")
		if err != nil {
			return nil, err
		}
		return reply("service", caller)
	})
	r.Command("rss").Alias("feed").Handle(func(*sdk.Context) (message.Chain, error) { return reply("rss", "rss") })
	{
		g := r.Group("rss")
		g.Command("add").Args("url").Handle(func(_ *sdk.Context, url string) (message.Chain, error) { return reply("add:"+url, "added:"+url) })
		g.Command("guarded").Args("n").Handle(func(_ *sdk.Context, n int64) (message.Chain, error) {
			return reply("guarded", strconv.FormatInt(n, 10))
		})
		g.Command("list").Handle(func(*sdk.Context) (message.Chain, error) { return reply("list", "list") })
	}
	r.Command("permissions").Handle(func(c *sdk.Context) (message.Chain, error) {
		return reply("permissions", strings.Join(c.Event().GetActor().GetPermissions(), ","))
	})
	r.Command("calc").Args("a", "b").Handle(func(_ *sdk.Context, a, b int64) (message.Chain, error) {
		return reply("calc", strconv.FormatInt(a+b, 10))
	})
	r.Command("kick").Args("user").Handle(func(_ *sdk.Context, user sdk.UserID) (message.Chain, error) {
		return reply("kick:"+string(user), string(user))
	})
	r.Command("picture").Args("image").Handle(func(c *sdk.Context, image *pb.ImageSegment) (message.Chain, error) {
		data, err := readImage(c, image)
		if err != nil {
			return nil, err
		}
		return reply("picture", string(data))
	})
	r.Command("relay").Args("body").Handle(func(c *sdk.Context, body *pb.MessageChain) (message.Chain, error) {
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
	})
	r.Command("quiet").Args("n").Silent().Handle(func(_ *sdk.Context, n int64) error { return record("quiet:" + strconv.FormatInt(n, 10)) })
	r.Command("say").RawTail().Handle(func(_ *sdk.Context, tail string) (message.Chain, error) { return reply("say:"+tail, tail) })
	r.Command("quote").Args("a", "b").Handle(func(_ *sdk.Context, a, b string) (message.Chain, error) { return reply("quote", a+"|"+b) })
	r.Command("optional").Optional("arg1", "9").Handle(func(_ *sdk.Context, n int64) (message.Chain, error) {
		return reply("optional", strconv.FormatInt(n, 10))
	})
	r.Command("custom").UsageKey("command.custom.usage").Handle(func(_ *sdk.Context, n int64) error { return record("custom") })
	r.Command("sendguard").Handle(func(c *sdk.Context) (message.Chain, error) {
		if err := record("sendguard"); err != nil {
			return nil, err
		}
		_, err := c.Send(message.Chain{message.Text("explicit")})
		return message.Chain{message.Text("automatic")}, err
	})
	r.Command("barrier").Handle(func(c *sdk.Context) (message.Chain, error) {
		c.StopPropagation()
		return message.Chain{message.Text("barrier")}, nil
	})
	r.Regex("regex", "(hello)").Priority(-10).Handle(func(c *sdk.Context) error { return record("regex:" + strings.Join(c.Match(), "|")) })
	r.Keyword("keyword", "hello").Priority(10).IgnoreCase().Handle(func(c *sdk.Context) error {
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
	})
	r.StartsWith("starts", "BEGIN").Handle(func(*sdk.Context) error { return record("starts") })
	r.EndsWith("ends", "END").Handle(func(*sdk.Context) error { return record("ends") })
	r.FullMatch("full", "hello", "WHOLE").Handle(func(*sdk.Context) error { return record("full") })
	r.Keyword("restricted.text", "protected").Priority(-20).Handle(func(*sdk.Context) error { return record("restricted.text") })
	r.Keyword("allowed.text", "protected").Priority(-19).Handle(func(*sdk.Context) error { return record("allowed.text") })
	r.Raw("restricted.raw").Priority(-101).Handle(func(*sdk.Context) error { return record("restricted.raw") })
	r.Raw("raw").Priority(-100).Handle(func(*sdk.Context) error { return record("raw") })
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
	r.On("event", "message.created").Handle(func(*sdk.Context) error { return record("event") })
	r.On("restricted.event", "member.joined").Priority(-1).Handle(func(*sdk.Context) error { return record("restricted.event") })
	r.On("observer", "member.joined").Handle(func(c *sdk.Context) error {
		return record("observer:" + strings.Join(c.Event().GetActor().GetPermissions(), ","))
	})
	r.On("adapter.poke", "memory.poke").Handle(func(c *sdk.Context) error {
		payload := c.Event().GetEventPayload().GetAdapterEvent()
		return record("adapter.poke:" + strconv.FormatUint(uint64(payload.GetVersion()), 10) + ":" + payload.GetPayloadJson())
	})
	r.On("uninstalled.poke", "not-installed.poke").Handle(func(*sdk.Context) error { return record("uninstalled.poke") })
}

func main() {
	if err := sdk.Run(&plugin{}); err != nil {
		slog.Error("command dispatch plugin failed", "error", err)
		os.Exit(1)
	}
}
