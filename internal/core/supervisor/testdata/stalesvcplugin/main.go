// stalesvcplugin 命令用过期的输出 Schema 验证 SDK 的运行期校验。
package main

import (
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/AlonaBot/AlonaBot/sdk"
)

const exitFailure = 1

type input struct{ Text string }
type reply struct{ Calls int64 }
type invalidOutput struct{ Secret string }

type plugin struct {
	sdk.Base
	calls atomic.Int64
}

func init() {
	// 模拟类型已更新、生成文件仍声明旧字段类型的插件。
	sdk.UseGenerated((*plugin)(nil), sdk.Generated{ServiceSchemas: map[string][]byte{
		"main.input":         []byte(`{"type":"object","properties":{"Text":{"type":"string"}},"required":["Text"]}`),
		"main.reply":         []byte(`{"type":"object","properties":{"Calls":{"type":"integer"}},"required":["Calls"]}`),
		"main.invalidOutput": []byte(`{"type":"object","properties":{"Secret":{"type":"integer"}},"required":["Secret"]}`),
	}})
}

// Services 注册过期 Schema 的输出服务与调用计数服务。
func (p *plugin) Services(r *sdk.ServiceRegistry) {
	sdk.JSON[input, reply](r, "count").Access(sdk.AccessPublic).Handle(func(*sdk.Context, input) (reply, error) {
		return reply{Calls: p.calls.Load()}, nil
	})
	sdk.JSON[input, invalidOutput](r, "invalid_output").Access(sdk.AccessPublic).Handle(func(*sdk.Context, input) (invalidOutput, error) {
		p.calls.Add(1)
		return invalidOutput{Secret: "output-payload-must-not-be-logged"}, nil
	})
}

func main() {
	if err := sdk.Run(&plugin{}); err != nil {
		slog.Error("stale service plugin failed", "error", err)
		os.Exit(exitFailure)
	}
}
