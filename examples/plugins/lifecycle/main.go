// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：可以真实运行的最小插件：由 Core 启动后走完 HELLO、REGISTER、ACTIVATE、STOPPING 的完整生命周期。日志只写 stderr，stdio 传输时 stdout 专用于协议帧。

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// Lifecycle 在每个生命周期 Hook 里向 stderr 打印一行。
type Lifecycle struct {
	// Base 提供根插件的默认实现。
	plugin.Base
}

// OnStart 打印 lifecycle: OnStart。
func (*Lifecycle) OnStart(context.Context) error {
	fmt.Fprintln(os.Stderr, "lifecycle: OnStart")
	return nil
}

// OnReady 打印 lifecycle: OnReady。
func (*Lifecycle) OnReady(context.Context) error {
	fmt.Fprintln(os.Stderr, "lifecycle: OnReady")
	return nil
}

// OnStop 打印 lifecycle: OnStop。
func (*Lifecycle) OnStop(context.Context) error {
	fmt.Fprintln(os.Stderr, "lifecycle: OnStop")
	return nil
}

// Commands 注册回复 pong 的 ping 命令。
func (*Lifecycle) Commands(r *sdk.CommandRegistry) {
	r.Command("ping", func(*sdk.Context) (message.Chain, error) {
		return message.Chain{message.Text("pong")}, nil
	})
}

// main 运行插件；正常结束时打印 lifecycle: 正常退出，失败时以 1 退出。
func main() {
	if err := sdk.Run(&Lifecycle{}); err != nil {
		fmt.Fprintln(os.Stderr, "lifecycle:", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "lifecycle: 正常退出")
}
