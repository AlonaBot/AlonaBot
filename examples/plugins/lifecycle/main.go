// lifecycle 是可以真实运行的最小插件示例。
//
// 它由 Core 启动，走完 HELLO、REGISTER、ACTIVATE、STOPPING 的完整生命周期，
// 在每个生命周期 Hook 中用 log/slog 记录一条 INFO。SDK 把记录以 JSON 行写到 stderr，由 Core 输出。
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// Lifecycle 在每个生命周期 Hook 里记录一条日志。
type Lifecycle struct {
	// Base 提供根插件的默认实现。
	plugin.Base
}

// OnStart 记录 OnStart。
func (*Lifecycle) OnStart(context.Context) error {
	slog.Info("OnStart")
	return nil
}

// OnReady 记录 OnReady。
func (*Lifecycle) OnReady(context.Context) error {
	slog.Info("OnReady")
	return nil
}

// OnStop 记录 OnStop。
func (*Lifecycle) OnStop(context.Context) error {
	slog.Info("OnStop")
	return nil
}

// Commands 注册回复 pong 的 ping 命令。
func (*Lifecycle) Commands(r *sdk.CommandRegistry) {
	r.Command("ping", func(*sdk.Context) (message.Chain, error) {
		return message.Chain{message.Text("pong")}, nil
	})
}

// main 运行插件；正常结束时记录 exiting normally，失败时记录错误并以 1 退出。
func main() {
	if err := sdk.Run(&Lifecycle{}); err != nil {
		slog.Error("plugin stopped with an error", "err", err)
		os.Exit(1)
	}
	slog.Info("exiting normally")
}
