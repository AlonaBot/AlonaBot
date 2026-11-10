package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const responsesEndpoint = "responses"

func responsesBody(model string, req Request) (map[string]any, error) {
	// 1. 展开完整历史，让推理项紧邻它所属的工具调用。
	input := []any{}
	systems := []string{}
	for _, message := range req.Messages {
		if message.Role == roleSystem {
			systems = append(systems, message.Content)
			continue
		}
		for _, result := range message.ToolResults {
			input = append(input, map[string]any{"type": "function_call_output", "call_id": result.ToolCallID, "output": result.Content})
		}
		if message.Content != "" {
			input = append(input, map[string]any{"role": message.Role, "content": message.Content})
		}
		for _, call := range message.ToolCalls {
			if len(call.Replay) > 0 {
				if bytes.HasPrefix(bytes.TrimSpace(call.Replay), []byte("[")) {
					var items []json.RawMessage
					if err := json.Unmarshal(call.Replay, &items); err != nil {
						return nil, fmt.Errorf("decode Responses replay: %w", err)
					}
					for _, item := range items {
						input = append(input, item)
					}
				} else {
					input = append(input, call.Replay)
				}
			}
			input = append(input, map[string]any{
				"type": "function_call", "call_id": call.ID, "name": call.Name, "arguments": string(call.Arguments),
			})
		}
	}

	// 2. 明确禁用服务端存储和工具严格模式，保留现有工具的 Schema。
	body := map[string]any{
		"model": model, "input": input, "stream": true, "store": false,
		"max_output_tokens": req.MaxTokens, "include": []string{"reasoning.encrypted_content"},
	}
	if len(systems) > 0 {
		body["instructions"] = strings.Join(systems, "\n\n")
	}
	if len(req.Tools) > 0 {
		tools := []map[string]any{}
		for _, tool := range req.Tools {
			tools = append(tools, map[string]any{
				"type": "function", "name": tool.Name, "description": tool.Description,
				"parameters": tool.InputSchema, "strict": false,
			})
		}
		body["tools"] = tools
	}
	return body, nil
}
