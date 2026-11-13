// showcase 演示 SDK 全部公共签名的组合方式；可以真实运行的最小插件见 examples/plugins/lifecycle。
//
// 示例包括一个根插件、两类 Handler、StopPropagation、配置回调、HTTP Handler 与静态及动态任务。
// New 显式绑定生成资源、路由与生命周期回调。
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
	plugin   *sdk.Plugin
	reminder sdk.TaskType[ReminderArgs]
}

// ReminderArgs 保存提醒发送所需的会话与完整文本。
type ReminderArgs struct {
	Session message.Session `json:"session"`
	Text    string          `json:"text"`
}

// New 每次创建独立的插件和示例状态。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()
	sdk.UseGenerated(p, alonabotGenerated)
	demo := &Demo{plugin: p, reminder: sdk.Task[ReminderArgs]("remind")}
	sdk.OnTask(p, demo.reminder).Handle(demo.Remind)
	p.OnConfigLoad(demo.OnConfigLoad)
	p.OnStart(demo.OnStart)
	demo.Commands(p)
	demo.HTTP(p)
	demo.Schedules(p)
	return p
}

// Commands 在 demo 分组下注册问候、回显和提醒命令。
func (p *Demo) Commands(r *sdk.Plugin) {
	{
		g := r.Group("demo")
		g.Command("hello").Handle(p.Hello)           // /demo hello
		g.Command("echo").Handle(p.Echo)             // /demo echo <text> <times>，例如 /demo echo hi 3
		g.Command("remind").Handle(p.CreateReminder) // /demo remind <key> <seconds> <text>
		g.Command("cancel").Handle(p.CancelReminder) // /demo cancel <key>
	}
}

// CreateReminder 以作者指定的 key 创建提醒，回复实际使用的取消键。
func (p *Demo) CreateReminder(c *sdk.Context, key string, seconds int64, text string) (message.Content, error) {
	key, err := p.reminder.At(c.Context(), time.Now().Add(time.Duration(seconds)*time.Second),
		ReminderArgs{Session: c.Session(), Text: text}, sdk.Key(key))
	if err != nil {
		return nil, err
	}
	return message.Text(key), nil
}

// CancelReminder 按 key 取消提醒，成功后回显取消键。
func (p *Demo) CancelReminder(c *sdk.Context, key string) (message.Content, error) {
	if err := p.reminder.Cancel(c.Context(), key); err != nil {
		return nil, err
	}
	return message.Text(key), nil
}

// Remind 使用参数保存的会话显式发送提醒。
func (*Demo) Remind(c *sdk.Context, args ReminderArgs) error {
	_, err := c.Send(message.Text(args.Text), sdk.To(args.Session))
	return err
}

// Hello 响应 /demo hello，返回要自动回复的 Chain；成功且非空时由 Core 发送一次。
func (p *Demo) Hello(c *sdk.Context) (message.Content, error) {
	c.StopPropagation()
	return message.Chain{message.Text(p.plugin.Config().String("greeting"))}, nil
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
func (p *Demo) HTTP(r *sdk.Plugin) {
	r.Route(http.MethodGet, "/status").HandleFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
}

// Schedules 注册每分钟运行一次的 tick 任务。
func (p *Demo) Schedules(r *sdk.Plugin) {
	r.Every("tick", time.Minute).Handle(func(*sdk.Context) error { return nil })
}

// OnStart 演示生命周期 Hook 的签名，不做任何事。
func (p *Demo) OnStart(context.Context) error { return nil }

// main 把根插件交给 SDK 运行，失败时打印原因并退出。
func main() {
	if err := sdk.Run(New()); err != nil {
		log.Fatal(err)
	}
}
