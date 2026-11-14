// Package plugin 提供库模式示例的消息回声插件，并在受管 dev.sqlite 中记录启动配置。
//
// 开发模式的 config.json 保留原文，secret 字段原样下发并跳过 Secret Store，
// 开发代次快照也保存 secret 原值；安装组件的快照使用 Secret 引用。
// dev 来源不分发，开发数据目录按原始配置保管。
package plugin

import (
	"context"
	_ "embed"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)

//go:generate go run ../../../cmd/alonabot build --out ../../../bin/generated/examples-devbot-plugin

//go:embed startup.sql
var startupDDL string

// New 创建独立插件，绑定生成资源，登记启动记录与回声路由。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()
	sdk.UseGenerated(p, alonabotGenerated)
	p.OnStart(func(ctx context.Context) error {
		db, err := sqlitedb.Open(ctx)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, startupDDL); err != nil {
			return err
		}
		var starts int64
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM startup").Scan(&starts); err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, "INSERT INTO startup (id, greeting) VALUES (?, ?)", starts+1, p.Config().Objects("greetings")[0].String("text"))
		return err
	})
	p.Command("hello").Description("Show a translated greeting").Handle(func(c *sdk.Context) error {
		return c.Reply(c.T("hello"))
	})
	p.On("echo", sdk.MessageCreated).Rule(sdk.NotCommand()).Handle(func(c *sdk.Context, e *message.CreatedPayload) error {
		_, err := c.Send(e.Chain)
		return err
	})
	return p
}
