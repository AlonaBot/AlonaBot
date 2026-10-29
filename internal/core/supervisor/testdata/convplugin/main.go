// Package main 运行经 SDK 操作会话存储的集成测试插件。
package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

type plugin struct{ sdk.Base }

type entry struct {
	Seq  uint64
	JSON string
}

type reply struct {
	Last    uint64
	Entries []entry
	Next    uint64
	More    bool
	Code    pb.ErrorCode
	Error   string
}

// Events 注册命令入口；前缀指定插件，便于验证同键隔离。
func (*plugin) Events(r *sdk.EventRegistry) {
	r.On("conversation", "message.created", func(c *sdk.Context) error {
		text := c.Event().GetEventPayload().GetMessageCreated().GetChain().GetSegments()[0].GetText().GetText()
		id, command, _ := strings.Cut(text, " ")
		if id != os.Getenv("CONV_PLUGIN_ID") {
			return nil
		}
		op, tail, _ := strings.Cut(command, " ")
		key, args, _ := strings.Cut(tail, " ")
		out := reply{}
		var err error
		switch op {
		case "append":
			var values []string
			err = json.Unmarshal([]byte(args), &values)
			if err == nil {
				entries := make([]json.RawMessage, len(values))
				for i, value := range values {
					entries[i] = json.RawMessage(value)
				}
				out.Last, err = sdk.AppendConversation(c.Context(), key, entries...)
			}
		case "list":
			var from uint64
			from, err = strconv.ParseUint(args, 10, 64)
			if err == nil {
				var entries []sdk.ConversationEntry
				entries, out.Next, out.More, err = sdk.ListConversation(c.Context(), key, from)
				for _, item := range entries {
					out.Entries = append(out.Entries, entry{Seq: item.Seq, JSON: string(item.Entry)})
				}
			}
		case "clear":
			err = sdk.ClearConversation(c.Context(), key)
		default:
			return errors.New("unknown conversation command")
		}
		if err != nil {
			var public *sdk.Error
			if !errors.As(err, &public) {
				return err
			}
			out.Code, out.Error = public.Code, public.Message
		}
		body, err := json.Marshal(out)
		if err != nil {
			return err
		}
		_, err = c.Send(message.Chain{message.Text(string(body))})
		return err
	})
}

func main() {
	if err := sdk.Run(&plugin{}); err != nil {
		slog.Error("conversation plugin stopped with an error", "error", err)
		os.Exit(1)
	}
}
