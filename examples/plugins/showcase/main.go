// showcase 演示 SDK 全部公共签名的组合方式；可以真实运行的最小插件见 examples/plugins/lifecycle。
//
// 示例包括一个根插件、两类 Handler、StopPropagation、配置回调、HTTP Handler 与一个定时任务。
// 根插件的 Hook 都由 sdk.Base 提供默认实现，这里只覆盖用到的。
package main

//go:generate go run ../../../cmd/alonabot plugin generate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// OnConfigLoad 检查 interval_seconds 非负，原样返回完整配置。
func (*Demo) OnConfigLoad(_ context.Context, _ uint32, raw json.RawMessage) (json.RawMessage, error) {
	var c struct {
		Interval int64 `json:"interval_seconds"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if c.Interval < 0 {
		return nil, errors.New("interval_seconds must not be negative")
	}
	return raw, nil
}

// Demo 作为根插件注册命令、HTTP 路由与定时任务；插件 ID 来自 plugin.yaml。
type Demo struct {
	// Base 提供根插件的默认实现。
	sdk.Base
}

// Commands 在 demo 分组下注册两个命令，行尾注释是各自的触发路径。
func (p *Demo) Commands(r *sdk.CommandRegistry) {
	{
		g := r.Group("demo")
		g.Command("hello").Handle(p.Hello) // /demo hello
		g.Command("echo").Handle(p.Echo)   // /demo echo <text> <times>，例如 /demo echo hi 3
	}
}

// Hello 响应 /demo hello，返回要自动回复的 Chain；成功且非空时由 Core 发送一次。
func (p *Demo) Hello(c *sdk.Context) (message.Chain, error) {
	c.StopPropagation()
	return message.Chain{message.Text(p.Config().String("greeting"))}, nil
}

// Echo 响应 /demo echo <text> <times>，在 Handler 内主动发送 times 次 text。
func (p *Demo) Echo(c *sdk.Context, text string, times int64) error {
	for i := int64(0); i < times; i++ {
		if _, err := c.Send(message.Chain{message.Text(text)}); err != nil {
			return err
		}
	}
	return nil
}

// HTTP 注册一个返回 ok 的 GET /status。
func (p *Demo) HTTP(r *sdk.HTTPRegistry) {
	r.Handle(http.MethodGet, "/status", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
}

// Schedules 注册每分钟运行一次的 tick 任务。
func (p *Demo) Schedules(r *sdk.ScheduleRegistry) {
	r.Every("tick", time.Minute, func(*sdk.Context) error { return nil })
}

// OnStart 演示生命周期 Hook 的签名，不做任何事。
func (p *Demo) OnStart(context.Context) error { return nil }

// main 把根插件交给 SDK 运行，失败时打印原因并退出。
func main() {
	if err := sdk.Run(&Demo{}); err != nil {
		log.Fatal(err)
	}
}
