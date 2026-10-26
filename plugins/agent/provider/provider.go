// Package provider 以流式方式调用 Agent 的模型服务（Provider），
// 支持 OpenAI 兼容的 Chat Completions 与 Anthropic Messages 两种协议。
//
// [Client.Stream] 逐个解析 SSE 事件，按到达顺序把文本增量交给回调；
// 工具调用的参数分段到达，组装完整且是 JSON 对象后才交出。
// 只有收到协议的终止标记才算成功，请求失败不自动重试。接口与流语义见 spec/PLAN.md T-12 步骤 1。
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// Protocol 是 Provider 的 HTTP API 协议。
type Protocol string

// 支持的 Provider 协议。
const (
	OpenAICompatible  Protocol = "openai-compatible"
	AnthropicMessages Protocol = "anthropic-messages"
)

const (
	openAIEndpoint    = "chat/completions"
	anthropicEndpoint = "messages"
	// anthropicVersion 是 Anthropic Messages API 要求的 anthropic-version 请求头。
	anthropicVersion = "2023-06-01"
	// openAIStreamDone 是 OpenAI 兼容流的终止标记。
	openAIStreamDone = "[DONE]"
)

// Message 的角色。
const (
	roleSystem    = "system"
	roleUser      = "user"
	roleAssistant = "assistant"
)

// Config 是一个 Provider 的连接设置。
type Config struct {
	Protocol Protocol
	BaseURL  string // API 根路径，例如 https://api.openai.com/v1；请求路径追加在其后
	APIKey   string
	Model    string
}

// Message 是对话中的一条消息。
type Message struct {
	Role        string // system、user 或 assistant
	Content     string
	ToolCalls   []ToolCall   // 只出现在 assistant 消息中
	ToolResults []ToolResult // 只出现在 user 消息中
}

// Tool 是提供给模型的一个工具。
type Tool struct {
	Name        string // 同一 Request 中唯一
	Description string
	InputSchema json.RawMessage // JSON Schema，必须是 JSON 对象
}

// Request 是一次流式请求。
type Request struct {
	Messages  []Message
	Tools     []Tool
	MaxTokens int // 必须大于 0；由调用方显式给出，不猜缺省值
}

// ToolCall 是模型发起的一次工具调用。
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage // 完整的 JSON 对象
}

// ToolResult 是一次工具调用的结果，回传给模型。
type ToolResult struct {
	ToolCallID string
	Content    string
}

// Event 是流中的一个事件：ToolCall 非 nil 时是一次组装完整的工具调用，否则 Text 是一段文本增量。
type Event struct {
	Text     string
	ToolCall *ToolCall
}

// Client 向一个 Provider 发起流式请求。
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    *url.URL
}

// New 校验 config 并返回 Client。httpClient 不能为 nil，由调用方注入以便复用代理设置。
// BaseURL 必须是不含用户信息、查询串与片段的 HTTP(S) 地址。
func New(config Config, httpClient *http.Client) (*Client, error) {
	if config.Protocol != OpenAICompatible && config.Protocol != AnthropicMessages {
		return nil, fmt.Errorf("unsupported provider protocol %q", config.Protocol)
	}
	if config.Model == "" || config.APIKey == "" {
		return nil, errors.New("provider model and API key are required")
	}
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil || baseURL.Host == "" || (baseURL.Scheme != "http" && baseURL.Scheme != "https") ||
		baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		// 不回显原始地址，避免把其中的凭证写进错误。
		return nil, errors.New("provider base URL must be HTTP(S) without credentials, query, or fragment")
	}
	return &Client{config: config, httpClient: httpClient, baseURL: baseURL}, nil
}

// Stream 发出 req，并把流中的事件依次交给 emit。
//
// emit 返回错误时立即停止读取并返回该错误。提前结束、无效事件、非 2xx 状态与 Provider
// 报告的错误都返回错误；已交给 emit 的文本增量由调用方保留，Stream 不重发请求。
// 错误文本不含 API key 与认证头。
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) error {
	// 1. 校验请求：消息与工具来自对话记录和模型输出，在发出前检查。
	if err := validateRequest(req); err != nil {
		return err
	}

	// 2. 按协议编码请求体，拼出端点并设置认证头。
	endpoint, body := anthropicEndpoint, anthropicBody(c.config.Model, req)
	if c.config.Protocol == OpenAICompatible {
		endpoint, body = openAIEndpoint, openAIBody(c.config.Model, req)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode provider request: %w", err)
	}
	target := c.baseURL.JoinPath(endpoint)
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build provider request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "text/event-stream")
	if c.config.Protocol == OpenAICompatible {
		httpRequest.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	} else {
		httpRequest.Header.Set("x-api-key", c.config.APIKey)
		httpRequest.Header.Set("anthropic-version", anthropicVersion)
	}

	// 3. 发出请求；响应体可能含 Provider 回显的凭证，非 2xx 时只报告状态码。
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("send provider request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("provider returned HTTP status %d", response.StatusCode)
	}

	// 4. 逐事件解析响应流。
	if c.config.Protocol == OpenAICompatible {
		return streamOpenAI(response.Body, emit)
	}
	return streamAnthropic(response.Body, emit)
}

func validateRequest(req Request) error {
	if req.MaxTokens <= 0 || len(req.Messages) == 0 {
		return errors.New("messages and positive max tokens are required")
	}
	for _, m := range req.Messages {
		if m.Role != roleSystem && m.Role != roleUser && m.Role != roleAssistant {
			return fmt.Errorf("invalid message role %q", m.Role)
		}
		if len(m.ToolCalls) > 0 && m.Role != roleAssistant {
			return errors.New("tool calls require assistant role")
		}
		if len(m.ToolResults) > 0 && m.Role != roleUser {
			return errors.New("tool results require user role")
		}
		for _, result := range m.ToolResults {
			if result.ToolCallID == "" {
				return errors.New("tool result call ID is required")
			}
		}
	}
	names := make(map[string]bool, len(req.Tools))
	for _, tool := range req.Tools {
		if tool.Name == "" || names[tool.Name] {
			return errors.New("tool names must be non-empty and unique")
		}
		names[tool.Name] = true
		if !isJSONObject(tool.InputSchema) {
			return fmt.Errorf("input schema of tool %q must be a JSON object", tool.Name)
		}
	}
	return nil
}

func isJSONObject(raw []byte) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(raw, &object) == nil && object != nil
}

// openAIBody 编码 Chat Completions 请求：每个工具结果单独成为一条 tool 角色消息。
func openAIBody(model string, req Request) map[string]any {
	messages := []map[string]any{}
	for _, m := range req.Messages {
		if len(m.ToolResults) > 0 {
			for _, result := range m.ToolResults {
				messages = append(messages, map[string]any{"role": "tool", "tool_call_id": result.ToolCallID, "content": result.Content})
			}
			continue
		}
		message := map[string]any{"role": m.Role, "content": m.Content}
		if len(m.ToolCalls) > 0 {
			calls := []map[string]any{}
			for _, call := range m.ToolCalls {
				calls = append(calls, map[string]any{
					"id":       call.ID,
					"type":     "function",
					"function": map[string]any{"name": call.Name, "arguments": string(call.Arguments)},
				})
			}
			message["tool_calls"] = calls
		}
		messages = append(messages, message)
	}
	body := map[string]any{"model": model, "messages": messages, "max_tokens": req.MaxTokens, "stream": true}
	if len(req.Tools) > 0 {
		tools := []map[string]any{}
		for _, tool := range req.Tools {
			tools = append(tools, map[string]any{
				"type":     "function",
				"function": map[string]any{"name": tool.Name, "description": tool.Description, "parameters": tool.InputSchema},
			})
		}
		body["tools"] = tools
	}
	return body
}

// anthropicBody 编码 Messages 请求：system 消息按顺序合并到顶层 system 字段，
// 工具调用与工具结果转成 content block。
func anthropicBody(model string, req Request) map[string]any {
	messages := []map[string]any{}
	systems := []string{}
	for _, m := range req.Messages {
		if m.Role == roleSystem {
			systems = append(systems, m.Content)
			continue
		}
		message := map[string]any{"role": m.Role, "content": m.Content}
		if len(m.ToolCalls) > 0 {
			blocks := []map[string]any{}
			for _, call := range m.ToolCalls {
				blocks = append(blocks, map[string]any{"type": "tool_use", "id": call.ID, "name": call.Name, "input": call.Arguments})
			}
			message["content"] = blocks
		}
		if len(m.ToolResults) > 0 {
			blocks := []map[string]any{}
			for _, result := range m.ToolResults {
				blocks = append(blocks, map[string]any{"type": "tool_result", "tool_use_id": result.ToolCallID, "content": result.Content})
			}
			message["content"] = blocks
		}
		messages = append(messages, message)
	}
	body := map[string]any{"model": model, "messages": messages, "max_tokens": req.MaxTokens, "stream": true}
	if len(systems) > 0 {
		body["system"] = strings.Join(systems, "\n\n")
	}
	if len(req.Tools) > 0 {
		tools := []map[string]any{}
		for _, tool := range req.Tools {
			tools = append(tools, map[string]any{"name": tool.Name, "description": tool.Description, "input_schema": tool.InputSchema})
		}
		body["tools"] = tools
	}
	return body
}

// readSSE 按 SSE 格式读取 r，每遇到一个空行就把累积的 event 名和 data 交给 handle。
// 流在两个事件之间结束时返回 nil，在事件中途结束时返回错误。
func readSSE(r io.Reader, handle func(event, data string) error) error {
	reader := bufio.NewReader(r)
	event := ""
	var data strings.Builder
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch {
		case line == "" && readErr == nil:
			// 空行结束一个事件。
			if data.Len() > 0 {
				if err := handle(event, strings.TrimSuffix(data.String(), "\n")); err != nil {
					return err
				}
			}
			event = ""
			data.Reset()
		case field == "event":
			event = value
		case field == "data":
			data.WriteString(value)
			data.WriteByte('\n')
		}
		// 以 ":" 开头的注释行与未知字段按 SSE 规范忽略。
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) {
				return fmt.Errorf("read provider stream: %w", readErr)
			}
			if data.Len() > 0 {
				return errors.New("provider stream ended in the middle of an event")
			}
			return nil
		}
	}
}

// openAIChunk 是 OpenAI 兼容流中一个 data 事件的载荷。
type openAIChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"` // 同一工具调用的各段共用一个 index
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
	Error json.RawMessage `json:"error"`
}

// streamOpenAI 解析 OpenAI 兼容流。工具调用的分段按 index 累积，
// 在收到 [DONE] 后按 index 顺序逐个校验并交出。
func streamOpenAI(r io.Reader, emit func(Event) error) error {
	calls := map[int]*ToolCall{}
	done := false
	err := readSSE(r, func(_, data string) error {
		if data == openAIStreamDone {
			done = true
			return nil
		}
		var chunk openAIChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("decode OpenAI-compatible stream event: %w", err)
		}
		if len(chunk.Error) > 0 {
			return errors.New("OpenAI-compatible provider reported an error")
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				if err := emit(Event{Text: choice.Delta.Content}); err != nil {
					return err
				}
			}
			for _, part := range choice.Delta.ToolCalls {
				call := calls[part.Index]
				if call == nil {
					call = &ToolCall{}
					calls[part.Index] = call
				}
				call.ID += part.ID
				call.Name += part.Function.Name
				call.Arguments = append(call.Arguments, part.Function.Arguments...)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !done {
		return errors.New("OpenAI-compatible stream ended without [DONE]")
	}
	for _, index := range slices.Sorted(maps.Keys(calls)) {
		call := calls[index]
		if call.ID == "" || call.Name == "" || !isJSONObject(call.Arguments) {
			return errors.New("invalid tool call in OpenAI-compatible stream")
		}
		if err := emit(Event{ToolCall: call}); err != nil {
			return err
		}
	}
	return nil
}

// anthropicEvent 是 Anthropic 流中一个 data 事件的载荷；不同 type 只使用其中部分字段。
type anthropicEvent struct {
	Type         string `json:"type"`
	Index        int    `json:"index"` // content block 的序号
	ContentBlock struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"content_block"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
	} `json:"delta"`
}

// streamAnthropic 解析 Anthropic Messages 流。tool_use block 在 content_block_stop 时交出。
func streamAnthropic(r io.Reader, emit func(Event) error) error {
	toolUses := map[int]*ToolCall{} // 未结束的 tool_use block，按 block 序号索引
	stopped := false
	err := readSSE(r, func(name, data string) error {
		var event anthropicEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("decode Anthropic stream event: %w", err)
		}
		if name == "error" || event.Type == "error" {
			return errors.New("Anthropic provider reported an error")
		}
		switch event.Type {
		case "content_block_start":
			if event.ContentBlock.Type == "tool_use" {
				toolUses[event.Index] = &ToolCall{ID: event.ContentBlock.ID, Name: event.ContentBlock.Name}
			}
		case "content_block_delta":
			switch event.Delta.Type {
			case "text_delta":
				if event.Delta.Text != "" {
					return emit(Event{Text: event.Delta.Text})
				}
			case "input_json_delta":
				call := toolUses[event.Index]
				if call == nil {
					return errors.New("Anthropic tool input delta has no matching tool_use block")
				}
				call.Arguments = append(call.Arguments, event.Delta.PartialJSON...)
			}
		case "content_block_stop":
			call := toolUses[event.Index]
			if call == nil {
				return nil
			}
			delete(toolUses, event.Index)
			if call.ID == "" || call.Name == "" || !isJSONObject(call.Arguments) {
				return errors.New("invalid tool call in Anthropic stream")
			}
			return emit(Event{ToolCall: call})
		case "message_stop":
			stopped = true
		case "message_start", "message_delta", "ping":
		default:
			return fmt.Errorf("unsupported Anthropic stream event %q", event.Type)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !stopped {
		return errors.New("Anthropic stream ended without message_stop")
	}
	if len(toolUses) > 0 {
		return errors.New("Anthropic stream has unfinished tool_use blocks")
	}
	return nil
}
