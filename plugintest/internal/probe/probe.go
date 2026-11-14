// Package probe 提供插件测试套件使用的命令、权限、配置、数据库与多轮等待插件。
//
// New 为每次测试创建独立的插件对象；配置和计数由真实 Core 的开发组件资源管理。
package probe

import (
	"context"
	_ "embed"
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)

const waitTimeout = time.Minute

//go:embed counter.sql
var counterDDL string

// New 登记测试命令，绑定生成资源，并在启动时初始化受管计数表。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()

	p.OnStart(func(ctx context.Context) error {
		db, err := sqlitedb.Open(ctx)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, counterDDL); err != nil {
			return err
		}
		if p.Config().Bool("fail_start") {
			return errors.New("requested probe startup failure")
		}
		return nil
	})
	p.Command("ping").Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Text(c.T("pong")), nil
	})
	p.Command("admin").Permission(sdk.Admin).Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Text(c.T("admin")), nil
	})
	p.Command("mod").Permission(sdk.ChatAdmin).Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Text(c.T("mod")), nil
	})
	p.Command("buttons").Handle(buttons)
	p.Command("config").Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Text(c.Config().String("value")), nil
	})
	p.Command("count").Handle(func(c *sdk.Context) (message.Content, error) {
		value, err := nextCount(c.Context())
		if err != nil {
			return nil, err
		}
		return message.Text(strconv.FormatInt(value, 10)), nil
	})
	p.Command("edit").Handle(edit)
	p.Command("ask").Handle(ask)
	p.Command("llmgenerate").Handle(generateLLM)
	p.Command("llmask").Handle(func(c *sdk.Context) (message.Content, error) {
		text, err := llm.Ask(c.Context(), "  original prompt\n")
		if err != nil {
			return nil, err
		}
		return message.Text(text), nil
	})
	registerScenarios(p)
	return p
}

func generateLLM(c *sdk.Context, usage string) (message.Content, error) {
	response, err := llm.Generate(c.Context(), llm.Request{
		Messages: []llm.Message{{Role: "user", Text: "first"}, {Role: "assistant", Text: "second"}, {Role: "user", Text: "third"}},
		System:   "system prompt", Provider: "synthetic", MaxOutputTokens: 37,
	})
	if err != nil {
		return nil, err
	}
	zero := 0
	expected := llm.Response{Text: "generated", Provider: "synthetic", Model: "test-model", OutputTokens: &zero}
	if usage == "input" {
		expected.InputTokens, expected.OutputTokens = &zero, nil
	}
	if !reflect.DeepEqual(response, expected) {
		return nil, errors.New("unexpected generation response")
	}
	return message.Text(response.Text), nil
}

func nextCount(ctx context.Context) (int64, error) {
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return 0, err
	}
	var value int64
	err = db.QueryRowContext(ctx, "UPDATE counter SET value = value + 1 WHERE id = 1 RETURNING value").Scan(&value)
	return value, err
}

func buttons(c *sdk.Context) error {
	_, err := c.Send(message.Chain{
		message.Text(c.T("buttons")),
		message.Buttons([]message.Button{message.URLButton(c.T("button_label"), "https://example.com")}),
	})
	if errcode.Of(err) == errcode.Unsupported {
		_, err = c.Send(message.Text(c.T("plain")))
	}
	return err
}

func edit(c *sdk.Context) error {
	sent, err := c.Send(message.Text(c.T("before_edit")))
	if err != nil {
		return err
	}
	_, err = c.Edit(sent.Ref(), message.Text(c.T("after_edit")))
	return err
}

func ask(c *sdk.Context) error {
	return c.Continue(func(c *sdk.Context) error {
		if _, err := c.Send(message.Text(c.T("question"))); err != nil {
			return err
		}
		next, err := c.Wait(waitTimeout)
		if err != nil {
			return err
		}
		_, err = next.Send(message.Text(next.Text()))
		return err
	})
}
