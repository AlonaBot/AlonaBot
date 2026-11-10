package provider

func ollamaBody(model string, req Request) (map[string]any, error) {
	// 1. 保持消息顺序，将工具结果关联到此前的调用名称。
	messages := []map[string]any{}
	for i, message := range req.Messages {
		for _, result := range message.ToolResults {
			name, err := toolCallName(req.Messages[:i], result.ToolCallID)
			if err != nil {
				return nil, err
			}
			messages = append(messages, map[string]any{"role": "tool", "content": result.Content, "tool_name": name})
		}
		if message.Role == roleUser && message.Content == "" {
			continue
		}
		encoded := map[string]any{"role": message.Role, "content": message.Content}
		if len(message.ToolCalls) > 0 {
			calls := []map[string]any{}
			for _, call := range message.ToolCalls {
				calls = append(calls, map[string]any{"function": map[string]any{"name": call.Name, "arguments": call.Arguments}})
			}
			encoded["tool_calls"] = calls
		}
		messages = append(messages, encoded)
	}

	// 2. 加入工具声明与调用方指定的输出上限。
	body := map[string]any{"model": model, "stream": true, "options": map[string]any{"num_predict": req.MaxTokens}, "messages": messages}
	if len(req.Tools) > 0 {
		tools := []map[string]any{}
		for _, tool := range req.Tools {
			tools = append(tools, map[string]any{"type": "function", "function": map[string]any{
				"name": tool.Name, "description": tool.Description, "parameters": tool.InputSchema,
			}})
		}
		body["tools"] = tools
	}
	return body, nil
}
