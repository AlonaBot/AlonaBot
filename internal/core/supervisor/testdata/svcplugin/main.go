// svcplugin 命令运行 Service 集成测试中的提供者与调用者。
package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const (
	pluginIDEnv   = "ALONABOT_TEST_SERVICE_ID"
	badOutputEnv  = "ALONABOT_TEST_BAD_SERVICE_OUTPUT"
	inputSchema   = "test.service.input.v1"
	outputSchema  = "test.service.output.v1"
	schemaVersion = 1
	exitFailure   = 1
)

type input struct{ Text string }
type reply struct {
	Text   string
	Caller string
	Calls  int64
	Code   pb.ErrorCode
	Error  string
}
type plugin struct {
	sdk.Base
	calls atomic.Int64
}

// Services 为 a 注册访问规则不同的 JSON 服务及 Protobuf 服务。
func (p *plugin) Services(r *sdk.ServiceRegistry) {
	if os.Getenv(pluginIDEnv) != "a" {
		return
	}
	echo := func(c *sdk.Context, in input) (reply, error) {
		return reply{Text: in.Text, Caller: c.Caller(), Calls: p.calls.Add(1)}, nil
	}
	sdk.JSONService(r, "echo", inputSchema, outputSchema, schemaVersion, echo)
	sdk.JSONService(r, "pub", inputSchema, outputSchema, schemaVersion, echo, sdk.WithAccess(sdk.AccessPublic))
	sdk.JSONService(r, "v2", inputSchema, outputSchema, schemaVersion+1, echo, sdk.WithAccess(sdk.AccessPublic))
	sdk.JSONService(r, "allow", inputSchema, outputSchema, schemaVersion, echo, sdk.WithAccess(sdk.AccessAllowlist, "b"))
	sdk.JSONService(r, "count", inputSchema, outputSchema, schemaVersion, func(*sdk.Context, input) (reply, error) { return reply{Calls: p.calls.Load()}, nil }, sdk.WithAccess(sdk.AccessPublic))
	sdk.ProtoService(r, "proto", func(_ *sdk.Context, in *pb.ConversationClearRequest) (*pb.ConversationClearRequest, error) {
		p.calls.Add(1)
		return in, nil
	}, sdk.WithAccess(sdk.AccessPublic))
}

// Events 在 b/c 按消息文本发起指定的 SDK Service 调用并回复结果。
func (*plugin) Events(r *sdk.EventRegistry) {
	if os.Getenv(pluginIDEnv) == "a" {
		return
	}
	r.On("call", "message.created").Handle(func(c *sdk.Context) error {
		text := c.Event().GetEventPayload().GetMessageCreated().GetChain().GetSegments()[0].GetText().GetText()
		parts := strings.Fields(text)
		if parts[0] != os.Getenv(pluginIDEnv) {
			return nil
		}
		operation, target := parts[1], parts[2]
		out := reply{}
		var err error
		switch operation {
		case "json", "count":
			out, err = sdk.CallJSON[reply](c.Context(), target, inputSchema, schemaVersion, input{Text: "hello 世界"})
		case "v2":
			out, err = sdk.CallJSON[reply](c.Context(), target, inputSchema, schemaVersion+1, input{Text: "hello 世界"})
		case "version":
			out, err = sdk.CallJSON[reply](c.Context(), target, inputSchema, schemaVersion+1, input{})
		case "schema":
			out, err = sdk.CallJSON[reply](c.Context(), target, "wrong.schema", schemaVersion, input{})
		case "invalid_json":
			out, err = sdk.CallJSON[reply](c.Context(), target, inputSchema, schemaVersion, json.RawMessage(`{`))
		case "encoding", "proto":
			var result *pb.ConversationClearRequest
			result, err = sdk.CallProto[*pb.ConversationClearRequest](c.Context(), target, &pb.ConversationClearRequest{ConversationKey: "hello 世界"})
			if err == nil {
				out.Text = result.ConversationKey
			}
		default:
			return errors.New("unknown service test operation")
		}
		if err != nil {
			var public *sdk.Error
			if !errors.As(err, &public) {
				return err
			}
			out.Code, out.Error = public.Code, public.Message
		}
		encoded, err := json.Marshal(out)
		if err != nil {
			return err
		}
		_, err = c.Send(message.Chain{message.Text(string(encoded))})
		return err
	})
}

func main() {
	var err error
	if os.Getenv(badOutputEnv) == "1" {
		err = runBadOutput()
	} else {
		err = sdk.Run(&plugin{})
	}
	if err != nil {
		slog.Error("service plugin failed", "error", err)
		os.Exit(exitFailure)
	}
}
