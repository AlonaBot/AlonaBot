// Package agent 提供当前 Agent 的多轮模型与工具 Run 服务契约。
package agent

import (
	"context"
	"encoding/json"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
	"github.com/AlonaBot/AlonaBot/sdk/service"
	"strings"
)

// ToolSelection 指定完整工具标识；空 Names 关闭工具。
type ToolSelection struct {
	Names []string `json:"names"`
}

// RunRequest 保存本次运行的输入；nil Tools 使用当前可访问的工具目录。
type RunRequest struct {
	Input llm.Request    `json:"input"`
	Tools *ToolSelection `json:"tools,omitempty"`
}

// Failure 保留跨服务可序列化的错误码与完整消息。
type Failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ToolCall 携带经过钩子处理的完整 JSON 对象参数。
type ToolCall struct {
	Name      string          `json:"name"`
	CallID    string          `json:"call_id"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult 同时保留原始结果、模型可见文本与执行错误。
type ToolResult struct {
	Call    ToolCall        `json:"call"`
	Raw     json.RawMessage `json:"raw"`
	Text    string          `json:"text"`
	Failure *Failure        `json:"failure,omitempty"`
}

// StepKind 区分模型步骤与工具步骤。
type StepKind string

const (
	// ModelStep 保存一次模型响应。
	ModelStep StepKind = "model"
	// ToolStep 保存一次工具执行。
	ToolStep StepKind = "tool"
)

// RunStep 按 Kind 恰好填写 Model 或 Tool，RequestID 关联所属模型请求。
type RunStep struct {
	Kind      StepKind      `json:"kind"`
	RequestID string        `json:"request_id"`
	Model     *llm.Response `json:"model,omitempty"`
	Tool      *ToolResult   `json:"tool,omitempty"`
}

// RunResult 保留实际步骤顺序及所有已交付模型文本；用量仅在每个请求均提供时合计。
type RunResult struct {
	RunID        string    `json:"run_id"`
	Text         string    `json:"text"`
	Steps        []RunStep `json:"steps"`
	InputTokens  *int      `json:"input_tokens,omitempty"`
	OutputTokens *int      `json:"output_tokens,omitempty"`
}

// EventKind 指定流式事件的有效载荷。
type EventKind string

const (
	// TextDelta 交付一段文本。
	TextDelta EventKind = "text_delta"
	// ToolStarted 交付即将执行的完整工具参数。
	ToolStarted EventKind = "tool_started"
	// ToolFinished 交付工具结果与执行错误。
	ToolFinished EventKind = "tool_finished"
)

// RunEvent 的 Sequence 在整个 Run 内从零递增，按 Kind 恰好填写对应载荷。
type RunEvent struct {
	RunID     string      `json:"run_id"`
	RequestID string      `json:"request_id"`
	Sequence  uint64      `json:"sequence"`
	Kind      EventKind   `json:"kind"`
	Text      string      `json:"text,omitempty"`
	Call      *ToolCall   `json:"call,omitempty"`
	Result    *ToolResult `json:"result,omitempty"`
}

// RunService 绑定独立的 Agent 工具循环服务。
var RunService = service.JSON[RunRequest, RunResult]("@agent/agent.run")

// RunStreamService 绑定同一 Run 的事件流模式。
var RunStreamService = service.StreamJSON[RunRequest, RunEvent, RunResult]("@agent/agent.run.stream")

// Run 调用一次独立 Agent Run，沿用调用身份、取消与期限。
func Run(ctx context.Context, req RunRequest) (RunResult, error) {
	if err := llm.Validate(req.Input); err != nil {
		return RunResult{}, err
	}
	return RunService.Call(ctx, req)
}

// RunStream 顺序交付 Run 事件，回调错误取消当前运行。
func RunStream(ctx context.Context, req RunRequest, onEvent func(RunEvent) error) (RunResult, error) {
	if err := llm.Validate(req.Input); err != nil {
		return RunResult{}, err
	}
	if onEvent == nil {
		return RunResult{}, errcode.InvalidArgument.New("stream callback is nil")
	}
	var delivered strings.Builder
	confirmed := RunResult{}
	response, err := RunStreamService.Call(ctx, req, func(event RunEvent) error {
		if err := onEvent(event); err != nil {
			return err
		}
		confirmed.RunID = event.RunID
		switch event.Kind {
		case TextDelta:
			delivered.WriteString(event.Text)
			last := len(confirmed.Steps) - 1
			if last >= 0 && confirmed.Steps[last].Kind == ModelStep && confirmed.Steps[last].RequestID == event.RequestID {
				confirmed.Steps[last].Model.Text += event.Text
			} else {
				confirmed.Steps = append(confirmed.Steps, RunStep{Kind: ModelStep, RequestID: event.RequestID, Model: &llm.Response{Text: event.Text}})
			}
		case ToolFinished:
			confirmed.Steps = append(confirmed.Steps, RunStep{Kind: ToolStep, RequestID: event.RequestID, Tool: event.Result})
		}
		return nil
	})
	if err != nil {
		if response.RunID == "" && len(response.Steps) == 0 {
			response = confirmed
		}
		response.Text = delivered.String()
		return response, err
	}
	if response.Text != delivered.String() {
		return response, errcode.Internal.New("run response text differs from delivered deltas")
	}
	return response, nil
}
