// Package agent 提供当前 Agent 组件的调用入口。
//
// [Generate] 发送完整文本生成请求，调用许可与期限见 CONTRACTS §13.5。
package agent

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/llm"
)

// Generate 经 Core 调用当前 Agent，沿用 ctx 的调用关联与期限。
func Generate(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.GenerateService.Call(ctx, req)
}
