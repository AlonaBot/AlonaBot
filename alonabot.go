// Package alonabot 在当前进程装配 Core，并通过 loopback Session 运行开发组件。
//
// Start 等全部组件激活后返回，调用方取消 context 后通过 Core.Wait 等待停止。
// Run 同时处理 SIGINT 与 SIGTERM。已安装组件仍由独立进程运行。
package alonabot

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlonaBot/AlonaBot/internal/app"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

// Option 配置一次 Core 启动。
type Option func(*app.Options)

// Plugin 登记由生成资源确定身份的开发插件或唯一 Agent。
func Plugin(root *sdk.Plugin) Option {
	return func(o *app.Options) { o.Components = append(o.Components, app.PluginComponent(root)) }
}

// Adapter 登记由生成资源确定身份的开发 Adapter。
func Adapter(a adapter.Adapter) Option {
	return func(o *app.Options) { o.Components = append(o.Components, app.AdapterComponent(a)) }
}

// WebUI 以 fsys 根目录作为本次运行的开发 WebUI，覆盖已安装组件的静态来源。
// 开发文件系统没有 manifest，启动时记录跳过 Admin API 版本检查的警告。
func WebUI(fsys fs.FS) Option {
	return func(o *app.Options) { o.WebUIs = append(o.WebUIs, fsys) }
}

// DataDir 指定 Core 数据目录，缺省为当前工作目录下的 data。
func DataDir(dir string) Option { return func(o *app.Options) { o.DataDir = dir } }

// Listen 覆盖本次运行的 HTTP 监听地址，安全条件与 core_config.json 相同。
func Listen(addr string) Option { return func(o *app.Options) { o.Listen = &addr } }

// LogLevel 覆盖本次运行的全局日志级别。
func LogLevel(level slog.Level) Option { return func(o *app.Options) { o.LogLevel = &level } }

// LogFormat 覆盖本次运行的 stderr 日志格式，接受 text 或 json。
func LogFormat(format string) Option { return func(o *app.Options) { o.LogFormat = &format } }

// Args 追加调优参数或参数文件，后于 ALONABOT_ARGS 应用。
func Args(values ...string) Option {
	args := append([]string(nil), values...)
	return func(o *app.Options) { o.Args = append(o.Args, args...) }
}

// Core 持有已装配的运行资源；取消 Start 的 context 后调用 Wait 等待停止结果。
type Core struct{ core *app.Core }

// Run 启动 Core，并等待 SIGINT、SIGTERM 或开发组件异常后完成停止。
func Run(opts ...Option) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	core, err := Start(ctx, opts...)
	if err == context.Canceled {
		return nil
	}
	if err != nil {
		return err
	}
	return core.Wait()
}

// Start 完成身份检查、握手、注册与激活后返回 Core；失败时清理已启动的部分。
// 启动期间 ctx 取消且清理成功时返回 context.Canceled，清理错误随启动错误一起返回。
func Start(ctx context.Context, opts ...Option) (*Core, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o := app.Options{DataDir: "data"}
	for _, apply := range opts {
		apply(&o)
	}
	core, err := app.Start(ctx, o)
	if err != nil {
		return nil, err
	}
	return &Core{core: core}, nil
}

// Wait 等待 Core 停止；组件停止超时或开发组件异常退出时返回含组件 ID 的错误。
func (c *Core) Wait() error { return c.core.Wait() }
