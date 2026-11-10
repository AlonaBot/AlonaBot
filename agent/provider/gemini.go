package provider

import "strings"

func geminiBody(req Request) (map[string]any, error) {
	// 1. 保持历史顺序，将工具结果关联到此前的调用。
	contents := []map[string]any{}
	systems := []string{}
	for i, message := range req.Messages {
		if message.Role == roleSystem {
			systems = append(systems, message.Content)
			continue
		}
		parts := []map[string]any{}
		for _, result := range message.ToolResults {
			name, err := toolCallName(req.Messages[:i], result.ToolCallID)
			if err != nil {
				return nil, err
			}
			parts = append(parts, map[string]any{"functionResponse": map[string]any{
				"id": result.ToolCallID, "name": name, "response": map[string]any{"output": result.Content},
			}})
		}
		if message.Content != "" {
			parts = append(parts, map[string]any{"text": message.Content})
		}
		for _, call := range message.ToolCalls {
			part := map[string]any{"functionCall": map[string]any{"id": call.ID, "name": call.Name, "args": call.Arguments}}
			if len(call.Replay) > 0 {
				part["thoughtSignature"] = call.Replay
			}
			parts = append(parts, part)
		}
		role := message.Role
		if role == roleAssistant {
			role = "model"
		}
		contents = append(contents, map[string]any{"role": role, "parts": parts})
	}

	// 2. 加入独立的 system 指令、工具声明与调用方给出的输出上限。
	body := map[string]any{"contents": contents, "generationConfig": map[string]any{"maxOutputTokens": req.MaxTokens}}
	if len(systems) > 0 {
		body["systemInstruction"] = map[string]any{"parts": []map[string]any{{"text": strings.Join(systems, "\n\n")}}}
	}
	if len(req.Tools) > 0 {
		declarations := []map[string]any{}
		for _, tool := range req.Tools {
			declarations = append(declarations, map[string]any{
				"name": tool.Name, "description": tool.Description, "parametersJsonSchema": tool.InputSchema,
			})
		}
		body["tools"] = []map[string]any{{"functionDeclarations": declarations}}
	}
	return body, nil
}
