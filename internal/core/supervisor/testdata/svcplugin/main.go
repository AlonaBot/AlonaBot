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
	inputSchema   = "main.input"
	outputSchema  = "main.reply"
	schemaVersion = 1
	exitFailure   = 1
)

type input struct{ Text string }
type inputV2 input

func (inputV2) SchemaVersion() uint32 { return 2 }

type wrongInput input
type reply struct {
	Text   string
	Caller string
	Calls  int64
	Code   pb.ErrorCode
	Error  string
}
type invalidOutput struct {
	Secret string
	Values []string
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
	sdk.JSON[input, reply](r, "echo").Handle(echo)
	sdk.JSON[input, reply](r, "pub").Access(sdk.AccessPublic).Handle(echo)
	sdk.JSON[inputV2, reply](r, "v2").Access(sdk.AccessPublic).Handle(func(c *sdk.Context, in inputV2) (reply, error) { return echo(c, input(in)) })
	sdk.JSON[input, reply](r, "allow").Access(sdk.AccessAllowlist, "b").Handle(echo)
	sdk.JSON[input, reply](r, "count").Access(sdk.AccessPublic).Handle(func(*sdk.Context, input) (reply, error) { return reply{Calls: p.calls.Load()}, nil })
	sdk.JSON[input, invalidOutput](r, "invalid_output").Access(sdk.AccessPublic).Handle(func(*sdk.Context, input) (invalidOutput, error) {
		p.calls.Add(1)
		return invalidOutput{Secret: "output-payload-must-not-be-logged"}, nil
	})
	sdk.Proto[*pb.ConversationClearRequest, *pb.ConversationClearRequest](r, "proto").Access(sdk.AccessPublic).Handle(func(_ *sdk.Context, in *pb.ConversationClearRequest) (*pb.ConversationClearRequest, error) {
		p.calls.Add(1)
		return in, nil
	})
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
			out, err = sdk.CallJSON[input, reply](c.Context(), target, input{Text: "hello 世界"})
		case "v2":
			out, err = sdk.CallJSON[inputV2, reply](c.Context(), target, inputV2{Text: "hello 世界"})
		case "version":
			out, err = sdk.CallJSON[inputV2, reply](c.Context(), target, inputV2{})
		case "schema":
			out, err = sdk.CallJSON[wrongInput, reply](c.Context(), target, wrongInput{})
		case "invalid_json":
			out, err = sdk.CallJSON[json.RawMessage, reply](c.Context(), target, json.RawMessage(`{`))
		case "invalid_output":
			_, err = sdk.CallJSON[input, invalidOutput](c.Context(), target, input{Text: "hello 世界"})
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
