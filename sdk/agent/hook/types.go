// Package hook 注册 Agent 八阶段的类型化服务；调度、注入和改写由 Agent 实现。
package hook

import (
	"encoding/json"
	"github.com/AlonaBot/AlonaBot/sdk/agent"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
)

// Meta 关联运行与请求；独立生成的 RunID 为空，Iteration 在 Run 中从 1 开始。
type Meta struct {
	RunID     string `json:"run_id"`
	RequestID string `json:"request_id"`
	Iteration int    `json:"iteration"`
	GoalID    string `json:"goal_id"`
}

// Scope 限定注入的生命周期。
type Scope string

const (
	// Request 注入仅持续到对应模型请求结束。
	Request Scope = "request"
	// Run 注入持续到本次 Run 结束。
	Run Scope = "run"
)

// Operation 指定注入更新或显式删除。
type Operation string

const (
	// Upsert 设置或更新文本。
	Upsert Operation = "upsert"
	// Delete 删除指定键。
	Delete Operation = "delete"
)

// InjectionChange 的所有者由 Agent 根据可信服务身份设置。
type InjectionChange struct {
	Op    Operation `json:"op"`
	Key   string    `json:"key"`
	Scope Scope     `json:"scope"`
	Text  string    `json:"text"`
}

// Injection 保存当前有效的注入及其可信所有者。
type Injection struct {
	PluginID string `json:"plugin_id"`
	RouteID  string `json:"route_id"`
	Key      string `json:"key"`
	Scope    Scope  `json:"scope"`
	Text     string `json:"text"`
}

// Message 表达普通消息、assistant 工具调用或对应的工具结果；Replay 由 Provider 原样保存且只读。
type Message struct {
	Role       string            `json:"role"`
	Text       string            `json:"text"`
	ToolCalls  []agent.ToolCall  `json:"tool_calls,omitempty"`
	ToolResult *agent.ToolResult `json:"tool_result,omitempty"`
	Replay     json.RawMessage   `json:"replay,omitempty"`
}

// Tool 描述当前可用工具及其 JSON 对象参数 Schema。
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// RequestView 分开保存基础系统提示词、消息与目录，Agent 统一拼装有效注入。
type RequestView struct {
	System          string    `json:"system"`
	Messages        []Message `json:"messages"`
	Model           string    `json:"model"`
	MaxOutputTokens int       `json:"max_output_tokens"`
	Tools           []Tool    `json:"tools"`
}

// RunInput 提供初始运行输入与有效注入。
type RunInput struct {
	Meta       Meta             `json:"meta"`
	Request    agent.RunRequest `json:"request"`
	Injections []Injection      `json:"injections"`
}

// RunPatch 的 nil 指针保持原值，非 nil 指针允许改为空值。
type RunPatch struct {
	Request    *agent.RunRequest `json:"request,omitempty"`
	Injections []InjectionChange `json:"injections"`
}

// RequestInput 提供当前基础模型请求和有效注入。
type RequestInput struct {
	Meta       Meta        `json:"meta"`
	Request    RequestView `json:"request"`
	Injections []Injection `json:"injections"`
}

// RequestPatch 仅调整可写请求字段；Provider replay 保持原样。
type RequestPatch struct {
	System          *string           `json:"system,omitempty"`
	Messages        *[]Message        `json:"messages,omitempty"`
	Model           *string           `json:"model,omitempty"`
	MaxOutputTokens *int              `json:"max_output_tokens,omitempty"`
	Tools           *[]Tool           `json:"tools,omitempty"`
	Injections      []InjectionChange `json:"injections"`
}

// ResponseInput 提供完整模式的实际响应。
type ResponseInput struct {
	Meta     Meta         `json:"meta"`
	Response llm.Response `json:"response"`
}

// ResponsePatch 只改写文本，实际模型与用量保持记录。
type ResponsePatch struct {
	Text *string `json:"text,omitempty"`
}

// StreamPhase 指定改写状态机事件。
type StreamPhase string

const (
	// Start 初始化单路由、单请求的私有状态。
	Start StreamPhase = "start"
	// Delta 处理交付前的文本块。
	Delta StreamPhase = "delta"
	// End 刷出剩余缓冲。
	End StreamPhase = "end"
	// Abort 清理失败请求的私有状态。
	Abort StreamPhase = "abort"
)

// StreamInput 的 State 仅属于当前钩子服务和 RequestID；Failure 在 Abort 时提供。
type StreamInput struct {
	Meta    Meta            `json:"meta"`
	Phase   StreamPhase     `json:"phase"`
	Text    string          `json:"text"`
	State   json.RawMessage `json:"state"`
	Failure *agent.Failure  `json:"failure,omitempty"`
}

// StreamPatch 返回零到多段输出；nil State 清空私有状态。
type StreamPatch struct {
	Texts []string        `json:"texts"`
	State json.RawMessage `json:"state"`
}

// ToolCallInput 提供即将执行的完整工具调用。
type ToolCallInput struct {
	Meta Meta           `json:"meta"`
	Call agent.ToolCall `json:"call"`
}

// ToolCallPatch 只修改参数；Agent 按最终参数重做 Schema 与工具策略检查。
type ToolCallPatch struct {
	Arguments *json.RawMessage `json:"arguments,omitempty"`
}

// ToolResultInput 保留真实执行结果及模型可见文本。
type ToolResultInput struct {
	Meta   Meta             `json:"meta"`
	Result agent.ToolResult `json:"result"`
}

// ToolResultPatch 只修改提供给模型的文本。
type ToolResultPatch struct {
	Text *string `json:"text,omitempty"`
}

// ResponseOutcome 观察改写后的完整响应或部分响应及错误。
type ResponseOutcome struct {
	Meta     Meta           `json:"meta"`
	Response llm.Response   `json:"response"`
	Failure  *agent.Failure `json:"failure,omitempty"`
}

// RunOutcome 观察已确认运行步骤及错误。
type RunOutcome struct {
	Meta    Meta            `json:"meta"`
	Result  agent.RunResult `json:"result"`
	Failure *agent.Failure  `json:"failure,omitempty"`
}

// Observed 确认观察者已完成处理。
type Observed struct{}
