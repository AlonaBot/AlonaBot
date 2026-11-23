// Package debugapi 定义经管理认证的 Adapter 调试请求与完整响应。
package debugapi

import "encoding/json"

// Operation 描述当前实例可调用的一项操作。
type Operation struct {
	Name         string          `json:"name"`
	Version      uint32          `json:"version"`
	Scope        string          `json:"scope"`
	ParamsSchema json.RawMessage `json:"params_schema"`
	ResultSchema json.RawMessage `json:"result_schema"`
}

// Call 指定实例、操作版本、会话与原始 JSON 参数；To 使用 Target 的 protobuf JSON。
type Call struct {
	Instance string          `json:"instance"`
	Name     string          `json:"name"`
	Version  uint32          `json:"version"`
	To       json.RawMessage `json:"to,omitempty"`
	Params   json.RawMessage `json:"params"`
}

// Error 保留业务错误的编号、安全消息和原始详情。
type Error struct {
	Code    string          `json:"code"`
	Message string          `json:"safe_message"`
	Details json.RawMessage `json:"details,omitempty"`
}

// Error 返回业务错误，供客户端以非零状态报告失败。
func (e *Error) Error() string { return e.Code + ": " + e.Message }

// Result 同时保存业务结果、业务错误及本次 Core 调用的 trace。
type Result struct {
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	TraceID string          `json:"trace_id"`
}

// EventRequest 携带标准 Event 的 protobuf JSON；实例身份由 Core 验证。
type EventRequest struct {
	Event json.RawMessage `json:"event"`
}

// Reply 保留实际发送的消息链、回执、错误与 trace。
type Reply struct {
	TraceID   string          `json:"trace_id"`
	Operation string          `json:"operation"`
	To        json.RawMessage `json:"to"`
	Chain     json.RawMessage `json:"chain"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *Error          `json:"error,omitempty"`
}

// Outcome 保存 Core 实际调用的 Handler 结果，包括返回的业务错误。
type Outcome struct {
	TraceID  string `json:"trace_id"`
	PluginID string `json:"plugin_id"`
	RouteID  string `json:"route_id"`
	Waiting  bool   `json:"waiting"`
	Error    *Error `json:"error,omitempty"`
}

const (
	// PhaseAdmission 标识 Handler 调用前的权限、生命周期和路由准入检查。
	PhaseAdmission = "admission"
	// PhaseArguments 标识 Handler 调用前的命令参数解析。
	PhaseArguments = "arguments"
)

// Diagnostic 保留 Handler 调用前的失败及其阶段、路由和 trace。
type Diagnostic struct {
	Phase    string `json:"phase"`
	TraceID  string `json:"trace_id"`
	PluginID string `json:"plugin_id"`
	RouteID  string `json:"route_id"`
	Error    *Error `json:"error"`
}

// EventResult 汇总事件正常分发到结束或会话等待时的结果。
type EventResult struct {
	TraceIDs    []string     `json:"trace_ids"`
	Waiting     bool         `json:"waiting"`
	Outcomes    []Outcome    `json:"outcomes"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	Replies     []Reply      `json:"replies"`
}
