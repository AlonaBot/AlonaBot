// Package llm 提供 Agent 文本生成服务的共享声明与调用辅助。
//
// [Generate] 完成一次结构化文本生成，[GenerateStream] 接收交付前改写的增量。
// 服务以 @agent 定位当前安装的 Agent，调用许可与期限见 https://alonabot.github.io/docs/advanced/llm。
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
	StopReason   string `json:"stop_reason,omitempty"`
	Model        string `json:"model"`                   // 实际使用的模型配置名。
	InputTokens  *int   `json:"input_tokens,omitempty"`  // nil 表示 Provider 未返回用量。
	OutputTokens *int   `json:"output_tokens,omitempty"` // 有用量时保留 0。
}

// GenerateService 共享提供方与调用方的目标和输入输出类型。
var GenerateService = service.JSON[Request, Response](AgentTarget + "/" + ServiceName)

// Delta 携带单次请求从零开始的文本交付序号。
type Delta struct {
	RequestID string `json:"request_id"`
	Sequence  uint64 `json:"sequence"`
	Text      string `json:"text"`
}

// GenerateStreamService 声明同一生成服务的流式模式。
var GenerateStreamService = service.StreamJSON[Request, Delta, Response](AgentTarget + "/" + ServiceName + ".stream")

// Validate 检查外部生成输入，保留调用者的完整文本。
func Validate(req Request) error {
	if len(req.Messages) == 0 {
		return errcode.InvalidArgument.New("messages must not be empty")
	}
	if req.MaxOutputTokens < 0 {
		return errcode.InvalidArgument.New("max_output_tokens must not be negative")
	}
	for _, msg := range req.Messages {
		if msg.Role != "user" && msg.Role != "assistant" {
			return errcode.InvalidArgument.Errorf("invalid message role %q", msg.Role)
		}
	}
	return nil
}

// Generate 沿用调用身份与期限，返回完整结果或带错误的部分结果。
func Generate(ctx context.Context, req Request) (Response, error) {
	if err := Validate(req); err != nil {
		return Response{}, err
	}
	return GenerateService.Call(ctx, req)
}

// GenerateStream 顺序处理增量，回调错误取消调用并随部分结果返回。
func GenerateStream(ctx context.Context, req Request, onDelta func(Delta) error) (Response, error) {
	if err := Validate(req); err != nil {
		return Response{}, err
	}
	if onDelta == nil {
		return Response{}, errcode.InvalidArgument.New("stream callback is nil")
	}
	var delivered strings.Builder
	response, err := GenerateStreamService.Call(ctx, req, func(delta Delta) error {
		if err := onDelta(delta); err != nil {
			return err
		}
		delivered.WriteString(delta.Text)
		return nil
	})
	if err != nil {
		response.Text = delivered.String()
		return response, err
	}
	if response.Text != delivered.String() {
		return response, errcode.Internal.New("stream response text differs from delivered deltas")
	}
	return response, nil
}
