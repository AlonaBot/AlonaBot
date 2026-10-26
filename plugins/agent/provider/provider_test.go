package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testAPIKey = "test-secret"

// 模拟服务返回的 SSE 流。
const (
	openAITextAndTool = "data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"c1\",\"function\":{\"name\":\"lookup\",\"arguments\":\"{\\\"q\\\":\"}}]}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"\\\"x\\\"}\"}}]}}]}\n\n" +
		"data: [DONE]\n\n"
	anthropicTextAndTool = "event: message_start\ndata: {\"type\":\"message_start\"}\n\n" +
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"hel\"}}\n\n" +
		": keep-alive\n\n" +
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"lo\"}}\n\n" +
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"c1\",\"name\":\"lookup\"}}\n\n" +
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"q\\\":\"}}\n\n" +
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"\\\"x\\\"}\"}}\n\n" +
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":1}\n\n" +
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
)

// TestStream 用本地模拟 Provider 驱动两种协议的完整请求与流解析（PLAN T-12 步骤 1，V-57 的 Provider 分项）。
func TestStream(t *testing.T) {
	stopErr := errors.New("stop")
	for _, tc := range []struct {
		name      string
		protocol  Protocol
		status    int
		stream    string
		emitErr   error // emit 返回的错误
		cancel    bool  // 发请求前取消 context
		wantText  string
		wantCalls int
		wantErr   error // 非 nil 时要求 errors.Is 匹配；为 nil 且 failed 时只要求出错
		failed    bool
	}{
		{name: "openai text and tool", protocol: OpenAICompatible, stream: openAITextAndTool, wantText: "hello", wantCalls: 1},
		{name: "anthropic text and tool", protocol: AnthropicMessages, stream: anthropicTextAndTool, wantText: "hello", wantCalls: 1},
		{name: "openai HTTP status", protocol: OpenAICompatible, status: http.StatusUnauthorized, stream: testAPIKey, failed: true},
		{name: "anthropic HTTP status", protocol: AnthropicMessages, status: http.StatusTooManyRequests, stream: testAPIKey, failed: true},
		{name: "openai missing DONE", protocol: OpenAICompatible, stream: strings.TrimSuffix(openAITextAndTool, "data: [DONE]\n\n"), wantText: "hello", failed: true},
		{name: "anthropic missing message_stop", protocol: AnthropicMessages, stream: strings.Split(anthropicTextAndTool, "event: content_block_start")[0], wantText: "hello", failed: true},
		{name: "openai cut mid-event", protocol: OpenAICompatible, stream: "data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}", failed: true},
		{name: "openai malformed JSON", protocol: OpenAICompatible, stream: "data: {not json}\n\n", failed: true},
		{name: "openai error event", protocol: OpenAICompatible, stream: "data: {\"error\":{\"message\":\"overloaded\"}}\n\n", failed: true},
		{name: "anthropic error event", protocol: AnthropicMessages, stream: "event: error\ndata: {\"type\":\"error\"}\n\n", failed: true},
		{name: "anthropic tool arguments not object", protocol: AnthropicMessages, stream: strings.ReplaceAll(anthropicTextAndTool, `\"x\"}`, `\"x\"]`), wantText: "hello", failed: true},
		{name: "emit error stops stream", protocol: OpenAICompatible, stream: openAITextAndTool, emitErr: stopErr, wantErr: stopErr, failed: true},
		{name: "canceled context", protocol: AnthropicMessages, stream: anthropicTextAndTool, cancel: true, wantErr: context.Canceled, failed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(mockProvider(t, tc.protocol, tc.status, tc.stream))
			t.Cleanup(server.Close)
			client, err := New(Config{Protocol: tc.protocol, BaseURL: server.URL + "/v1", APIKey: testAPIKey, Model: "test-model"}, server.Client())
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancel {
				cancel()
			}

			var text strings.Builder
			var calls []ToolCall
			err = client.Stream(ctx, testRequest(), func(e Event) error {
				if e.ToolCall != nil {
					calls = append(calls, *e.ToolCall)
				}
				text.WriteString(e.Text)
				return tc.emitErr
			})

			if tc.failed != (err != nil) || (tc.wantErr != nil && !errors.Is(err, tc.wantErr)) {
				t.Fatalf("Stream() error = %v, want failed=%v (%v)", err, tc.failed, tc.wantErr)
			}
			if err != nil && strings.Contains(err.Error(), testAPIKey) {
				t.Fatalf("error leaks API key: %v", err)
			}
			if tc.emitErr == nil && text.String() != tc.wantText {
				t.Fatalf("text = %q, want %q", text.String(), tc.wantText)
			}
			if len(calls) != tc.wantCalls {
				t.Fatalf("tool calls = %+v, want %d", calls, tc.wantCalls)
			}
			if tc.wantCalls > 0 && (calls[0].ID != "c1" || calls[0].Name != "lookup" || string(calls[0].Arguments) != `{"q":"x"}`) {
				t.Fatalf("tool call = %+v", calls[0])
			}
		})
	}
}

// mockProvider 返回模拟 Provider：校验路径、认证头与请求体编码后，按 status 返回 stream。
func mockProvider(t *testing.T, protocol Protocol, status int, stream string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []map[string]any `json:"messages"`
			System   string           `json:"system"`
			Tools    []map[string]any `json:"tools"`
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		switch protocol {
		case OpenAICompatible:
			if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer "+testAPIKey {
				t.Errorf("path %s, Authorization %q", r.URL.Path, r.Header.Get("Authorization"))
			}
			// system、user、assistant 与每个工具结果各一条 tool 消息。
			if len(body.Messages) != 6 || body.Messages[4]["role"] != "tool" || body.Messages[5]["tool_call_id"] != "c00" {
				t.Errorf("messages = %s", raw)
			}
		case AnthropicMessages:
			if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != testAPIKey || r.Header.Get("anthropic-version") != anthropicVersion {
				t.Errorf("path %s, headers %v", r.URL.Path, r.Header)
			}
			// system 合并到顶层，工具结果放进同一条 user 消息的 content block。
			if body.System != "one\n\ntwo" || len(body.Messages) != 3 {
				t.Errorf("request = %s", raw)
			}
		}
		if len(body.Tools) != 1 {
			t.Errorf("tools = %s", raw)
		}
		if status != 0 {
			w.WriteHeader(status)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, stream)
	}
}

func testRequest() Request {
	return Request{
		Messages: []Message{
			{Role: "system", Content: "one"},
			{Role: "system", Content: "two"},
			{Role: "user", Content: "hi"},
			{Role: "assistant", ToolCalls: []ToolCall{{ID: "c0", Name: "lookup", Arguments: json.RawMessage(`{"q":"a"}`)}, {ID: "c00", Name: "lookup", Arguments: json.RawMessage(`{"q":"b"}`)}}},
			{Role: "user", ToolResults: []ToolResult{{ToolCallID: "c0", Content: "A"}, {ToolCallID: "c00", Content: "B"}}},
		},
		Tools:     []Tool{{Name: "lookup", Description: "look up", InputSchema: json.RawMessage(`{"type":"object"}`)}},
		MaxTokens: 20,
	}
}
