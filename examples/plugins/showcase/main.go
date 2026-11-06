// showcase 演示 SDK 全部公共签名的组合方式；可以真实运行的最小插件见 examples/plugins/lifecycle。
//
// 示例包括一个根插件、两类 Handler、StopPropagation、配置回调、HTTP Handler 与一个定时任务。
// 根插件的 Hook 都由 sdk.Base 提供默认实现，这里只覆盖用到的。
// SDK 尚不支持 HTTP 路由与定时任务，直接运行时 Run 返回 UNSUPPORTED。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Config 声明本插件唯一的配置类型；config.Base 提供默认的 OnLoad。
type Config struct {
	config.Base
	Greeting string `json:"greeting"`         // /demo hello 回复的文本
	Interval int64  `json:"interval_seconds"` // 间隔秒数，不能为负
}

// OnLoad 把版本 1 的 {"hello": "..."} 迁移为当前格式，并拒绝负数的 interval_seconds。
func (*Config) OnLoad(_ context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error) {
	// 1. 版本 1 的 hello 改名为 greeting；值保持原始 JSON，不经 float64 转换。
	if fromVersion == 1 {
		var doc map[string]json.RawMessage
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("decode config: %w", err)
		}
		if v, ok := doc["hello"]; ok {
			doc["greeting"] = v
			delete(doc, "hello")
		}
		var err error
		if raw, err = json.Marshal(doc); err != nil {
			return nil, fmt.Errorf("encode config: %w", err)
		}
	}
	// 2. 解码到局部变量检查取值，raw 原样返回，保留未知键。
	var c Config
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
	config Config // SDK 解码后的配置
}

// Config 返回供 SDK 解码配置的目标。
func (p *Demo) Config() config.Config { return &p.config }

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
	return message.Chain{message.Text(p.config.Greeting)}, nil
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
