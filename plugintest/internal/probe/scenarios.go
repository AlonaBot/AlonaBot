package probe

import (
	"context"
	"strconv"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/button"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	taskapi "github.com/AlonaBot/AlonaBot/sdk/task"
)

const scheduleInterval = 24 * time.Hour

// TaskArgs 保存 Core 持久化并在触发时恢复的发送参数。
type TaskArgs struct {
	Session string `json:"session"`
	Text    string `json:"text"`
}

func registerScenarios(p *sdk.Plugin) {
	first, second, admin := button.Define[string]("first"), button.Define[string]("second"), button.Define[string]("admin")
	click := func(c *sdk.Context, value string) error {
		return c.Reply(c.T("choice", "value", value, "user", c.User()))
	}
	button.On(p, first).Handle(click)
	button.On(p, second).Handle(click)
	button.On(p, admin).Permission(sdk.Admin).Handle(click)
	p.Command("menu").Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Chain{message.Text(c.T("buttons")), message.Buttons([]message.Button{
			first.Button(c.T("first"), "one"), second.Button(c.T("second"), "two"),
		})}, nil
	})
	p.Command("adminmenu").Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Chain{message.Text(c.T("buttons")), message.Buttons([]message.Button{admin.Button(c.T("admin"), "admin")})}, nil
	})
	p.Command("rule").Rule(sdk.ToMe()).Handle(func(c *sdk.Context) error { return c.Reply(c.T("pong")) })
	p.Command("album").Handle(func(c *sdk.Context) error {
		_, err := c.Send(message.Chain{message.ImageURL("https://example.com/one.png"), message.ImageURL("https://example.com/two.png")})
		return err
	})
	p.Every("tick", scheduleInterval).Handle(func(c *sdk.Context) error {
		_, err := c.SendTo(message.SID("plugintest", "private:alice", ""), message.Text(c.T("tick")))
		return err
	})
	task := taskapi.Define[TaskArgs]("remind")
	taskapi.On(p, task).Handle(func(c *sdk.Context, args TaskArgs) error {
		_, err := c.SendTo(args.Session, message.Text(args.Text))
		return err
	})
	p.Command("remind").Args("key", "text").Handle(func(c *sdk.Context, key, text string) error {
		_, err := task.At(c.Context(), time.Now().Add(scheduleInterval), TaskArgs{Session: c.Session(), Text: text}, taskapi.Key(key))
		return err
	})
	p.Command("tasks").Handle(func(c *sdk.Context) error {
		tasks, err := task.List(c.Context())
		if err != nil {
			return err
		}
		return c.Reply(strconv.Itoa(len(tasks)))
	})
}

// Background 让测试在 Start 返回后控制一次插件身份的真实后台发送。
type Background struct {
	Send   <-chan struct{}
	Done   chan<- error
	To     string
	cancel context.CancelFunc
}

// Ready 只登记受插件 context 管理的 goroutine，由测试 channel 决定发送时刻。
func (b *Background) Ready(ctx context.Context) error {
	// OnReady 的调用期限在返回时结束；后台归属插件，通过 OnStop 取消。
	ctx, b.cancel = context.WithCancel(context.WithoutCancel(ctx))
	go func() {
		select { // 停止插件时退出，测试无需为清理再发送信号。
		case <-ctx.Done():
			return
		case <-b.Send:
		}
		_, err := sdk.Send(ctx, b.To, message.Text(sdk.T(ctx, "background")))
		select {
		case b.Done <- err:
		case <-ctx.Done():
		}
	}()
	return nil
}

// Stop 结束尚未收到测试信号的后台工作。
func (b *Background) Stop(context.Context) error {
	b.cancel()
	return nil
}
