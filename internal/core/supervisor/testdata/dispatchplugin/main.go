// dispatchplugin 运行分发排序、并发与可信事件的集成测试插件。
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/encoding/protojson"
)

const auditEnv = "ALONABOT_TEST_DISPATCH_AUDIT"
const pluginIDEnv = "ALONABOT_PLUGIN_ID"
const priorityEnv = "ALONABOT_TEST_DISPATCH_PRIORITY"

type plugin struct {
	sdk.Base
	priority int32
}

func record(route, phase string, event *pb.Event) error {
	body, err := protojson.Marshal(event)
	if err != nil {
		return err
	}
	entry := struct {
		Plugin, Route, Phase string
		Event                json.RawMessage
	}{os.Getenv(pluginIDEnv), route, phase, body}
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(os.Getenv(auditEnv), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	_, writeError := file.Write(append(encoded, '\n'))
	return errors.Join(writeError, file.Close())
}

func handle(route string, c *sdk.Context) error {
	ev := c.Event()
	if err := record(route, "entered", ev); err != nil {
		return err
	}
	var text strings.Builder
	for _, s := range ev.GetEventPayload().GetMessageCreated().GetChain().GetSegments() {
		text.WriteString(s.GetText().GetText())
	}
	words := strings.Fields(text.String())
	duration, err := time.ParseDuration(words[0])
	if err != nil {
		return err
	}
	// 独立 Send 与慢 Handler 共用同一 Session；收到 Response 后才记录完成。
	var probe chan error
	if len(words) > 1 && words[1] == "probe" && route == "a" {
		probe = make(chan error, 1)
		go func() {
			result, err := sdk.Send(c.Context(), ev.GetTarget(), message.Chain{message.Text("probe:" + ev.GetEventId())})
			if err == nil && result.GetStatus() != pb.SendStatus_SEND_STATUS_ACCEPTED {
				err = fmt.Errorf("probe send status: %s", result.GetStatus())
			}
			if err == nil {
				err = record(route, "send.response", ev)
			}
			probe <- err
		}()
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-c.Context().Done():
		return c.Context().Err()
	}
	if probe != nil {
		if err := <-probe; err != nil {
			return err
		}
	}
	_, err = c.Send(message.Chain{message.Text(os.Getenv(pluginIDEnv) + ":" + route + ":" + ev.GetEventId())})
	if err != nil {
		return err
	}
	return record(route, "finished", ev)
}

// Events 逆注册同优先级路由，排序由 Core 决定；任务事件只记录文件，不假设默认目标。
func (p *plugin) Events(r *sdk.EventRegistry) {
	for _, route := range []string{"z", "a"} {
		r.On(route, "message.created").Priority(p.priority).Handle(func(c *sdk.Context) error { return handle(route, c) })
	}
	for _, route := range []string{"button", "scheduled.open", "scheduled.user"} {
		eventType := "scheduled"
		if route == "button" {
			eventType = "button.clicked"
		}
		r.On(route, eventType).Handle(func(c *sdk.Context) error { return record(route, "entered", c.Event()) })
	}
}

func main() {
	priority, err := strconv.ParseInt(os.Getenv(priorityEnv), 10, 32)
	if err == nil {
		err = sdk.Run(&plugin{priority: int32(priority)})
	}
	if err != nil {
		slog.Error("dispatch plugin failed", "error", err)
		os.Exit(1)
	}
}
