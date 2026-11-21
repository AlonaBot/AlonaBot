// Package llm 提供 Agent 文本生成服务的共享声明与调用辅助。
//
// [Ask] 使用对话模型完成一问一答，完整请求通过 sdk/agent.Generate 发送。
// 服务以 @agent 定位当前安装的 Agent，调用许可与期限见 docs/plugin-guide.md。
package llm

import (
	"context"
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/service"
)

const (
	// ServiceName 指定 Agent 提供的文本生成服务。
	ServiceName = "llm.generate"
	// AgentTarget 由 Core 解析为当前安装的 Agent 组件 ID。
	AgentTarget = "@agent"
	// SchemaIn 标识共享请求的 JSON Schema。
	SchemaIn = "github.com/AlonaBot/AlonaBot/sdk/llm.Request"
	// SchemaOut 标识共享响应的 JSON Schema。
	SchemaOut = "github.com/AlonaBot/AlonaBot/sdk/llm.Response"
	// SchemaVersion 使用 SDK 类型声明的缺省版本。
	SchemaVersion uint32 = 1
)

// Message 按角色携带一条输入文本，Role 只接受 user 或 assistant。
type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// Request 描述一次独立生成，Messages 至少包含一条消息。
type Request struct {
	Messages        []Message `json:"messages"`
	System          string    `json:"system,omitempty"`
	Model           string    `json:"model,omitempty"`             // 空串使用 Agent 的 chat_model。
	MaxOutputTokens int       `json:"max_output_tokens,omitempty"` // 0 使用所选模型的配置值。
}

// Response 返回生成文本、实际使用的模型与可选的 token 用量。
type Response struct {
	Text         string `json:"text"`
	Model        string `json:"model"`                   // 实际使用的模型配置名。
	InputTokens  *int   `json:"input_tokens,omitempty"`  // nil 表示 Provider 未返回用量。
	OutputTokens *int   `json:"output_tokens,omitempty"` // 有用量时保留 0。
}

// GenerateService 共享提供方与调用方的目标和输入输出类型。
var GenerateService = service.JSON[Request, Response](AgentTarget + "/" + ServiceName)

// Ask 将非空白 prompt 原样作为唯一 user 消息，返回对话模型的生成文本。
func Ask(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", errcode.InvalidArgument.New("prompt must not be blank")
	}
	response, err := GenerateService.Call(ctx, Request{Messages: []Message{{Role: "user", Text: prompt}}})
	return response.Text, err
}
