// minimal 是 T-01 的签名组合示例：一个根插件、两类 Handler、StopPropagation、
// 配置回调、HTTP Handler 与一个定时任务。运行行为从 T-02 起由 SDK 提供。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/AlonaBot/alonabot/sdk"
	"github.com/AlonaBot/alonabot/sdk/message"
	"github.com/AlonaBot/alonabot/sdk/plugin"
)

// Config 是本插件唯一的配置类型。
type Config struct {
	Greeting string `json:"greeting"`
	Interval int64  `json:"interval_seconds"`
}

// OnLoad 把版本 1 的 {"hello": "..."} 迁移为当前格式；其他版本原样返回。
func (c *Config) OnLoad(_ context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error) {
	if fromVersion != 1 {
		return raw, nil
	}
	var old map[string]any
	if err := json.Unmarshal(raw, &old); err != nil {
		return nil, err
	}
	if v, ok := old["hello"]; ok {
		old["greeting"] = v
		delete(old, "hello")
	}
	return json.Marshal(old)
}

func (c *Config) Validate(context.Context) error {
	if c.Interval < 0 {
		return errors.New("interval_seconds 不能为负数")
	}
	return nil
}

// Demo 是根插件；插件 ID 来自 plugin.yaml。
type Demo struct {
	plugin.Base
	cfg Config
}

func (p *Demo) Config() any { return &p.cfg }

func (p *Demo) Commands(r *sdk.CommandRegistry) {
	r.Group("demo", func(g *sdk.CommandGroup) {
		g.Command("hello", p.Hello)
		g.Command("echo", p.Echo)
	})
}

// Hello 是自动回复型 Handler：成功返回非空 Chain 时由 Core 发送一次。
func (p *Demo) Hello(c *sdk.Context) (message.Chain, error) {
	c.StopPropagation()
	return message.Chain{message.Text(p.cfg.Greeting)}, nil
}

// Echo 是主动发送型 Handler：可以多次 Send。
func (p *Demo) Echo(c *sdk.Context, text string, times int64) error {
	for i := int64(0); i < times; i++ {
		if _, err := c.Send(message.Chain{message.Text(text)}); err != nil {
			return err
		}
	}
	return nil
}

func (p *Demo) HTTP(r *sdk.HTTPRegistry) {
	r.Handle(http.MethodGet, "/status", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
}

func (p *Demo) Schedules(r *sdk.ScheduleRegistry) {
	r.Every("tick", time.Minute, func(*sdk.Context) error { return nil })
}

func (p *Demo) OnStart(context.Context) error { return nil }

func main() {
	if err := sdk.Run(&Demo{}); err != nil {
		log.Fatal(err)
	}
}
