package message

import "encoding/json"

// SupportStatus 表达本次查询或预检的支持状态。
type SupportStatus string

const (
	// SupportUnknown 表示缺少声明或决定组合的元数据。
	SupportUnknown SupportStatus = "unknown"
	// SupportSupported 表示已声明支持或本次可确定检查通过。
	SupportSupported SupportStatus = "supported"
	// SupportUnsupported 表示明确声明或预检拒绝。
	SupportUnsupported SupportStatus = "unsupported"
)

// MessageFormat 描述标准片段或指定版本的扩展载荷。
type MessageFormat struct {
	Name             string          `json:"name"`
	Version          uint32          `json:"version"`
	Status           SupportStatus   `json:"status"`
	Description      string          `json:"description"`
	SchemaJSON       json.RawMessage `json:"schema_json,omitempty"`  // 标准格式描述 part，扩展描述该版本的 payload。
	ExampleJSON      json.RawMessage `json:"example_json,omitempty"` // 标准格式给完整 send_message 参数，扩展给 payload。
	Sources          []string        `json:"sources,omitempty"`      // 缺省表示来源支持情况未知。
	TextLimit        *TextLimit      `json:"text_limit,omitempty"`
	ButtonLabelLimit *TextLimit      `json:"button_label_limit,omitempty"`
	AdapterID        string          `json:"adapter_id,omitempty"`
}

// CombinationRule 描述平台编排规则；ExampleJSON 包含完整 send_message 参数。
type CombinationRule struct {
	Name        string          `json:"name"`
	Formats     []string        `json:"formats"`
	Description string          `json:"description"`
	ExampleJSON json.RawMessage `json:"example_json,omitempty"`
}

// OutboundDescription 汇集当前目标实例的标准格式、扩展和组合声明。
type OutboundDescription struct {
	Formats      []MessageFormat   `json:"formats"`
	Extensions   []MessageFormat   `json:"extensions"`
	Combinations []CombinationRule `json:"combinations"`
}

// StandardFormats 返回各标准 part 的字段 Schema 和完整 send_message 例子。
// 状态初始为 unknown；来源与限制由目标实例的声明提供。
// Schema 使用消息工具的字段名，来源的 incoming 在工具中写作 attachment。
func StandardFormats() []MessageFormat {
	// 1. 建立标准字段、来源、实体与按钮的公共形状。
	text := map[string]any{"type": "string"}
	object := func(properties map[string]any, required ...string) map[string]any {
		return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
	}
	array := func(items any) map[string]any { return map[string]any{"type": "array", "items": items, "minItems": 1} }
	var sources []any
	for _, name := range []string{"https_url", "media_ref", "attachment", "local_path"} {
		sources = append(sources, object(map[string]any{name: text}, name))
	}
	source := map[string]any{"oneOf": sources}
	entity := object(map[string]any{
		"type":         map[string]any{"type": "integer", "enum": []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, "description": "1 bold, 2 italic, 3 underline, 4 strikethrough, 5 spoiler, 6 code, 7 pre, 8 text_link, 9 text_mention. url is required only for text_link; language is optional only for pre; external_user_id is required only for text_mention."},
		"offset_utf16": map[string]any{"type": "integer", "minimum": 0},
		"length_utf16": map[string]any{"type": "integer", "minimum": 1},
		"url":          text, "language": text, "external_user_id": text,
	}, "type", "offset_utf16", "length_utf16")
	button := map[string]any{"oneOf": []any{object(map[string]any{"label": text, "url": text}, "label", "url"), object(map[string]any{"label": text}, "label")}}
	parts := []struct {
		name, description string
		fields            map[string]any
		required          []string
		example           string
	}{
		{"text", "Text with optional entities measured in UTF-16 code units; compose with other formats.", map[string]any{"text": text, "entities": map[string]any{"type": "array", "items": entity}}, []string{"text"}, `{"type":"text","text":"Details","entities":[{"type":1,"offset_utf16":0,"length_utf16":7}]}`},
		{"markdown", "CommonMark with strikethrough; compose with other formats.", map[string]any{"content": text}, []string{"content"}, `{"type":"markdown","content":"**Details**"}`},
		{"image", "Image with an optional caption.", map[string]any{"source": source, "caption": text}, []string{"source"}, `{"type":"image","source":{"https_url":"https://example.com/image.jpg"},"caption":"Image"}`},
		{"video", "Video with an optional caption.", map[string]any{"source": source, "caption": text}, []string{"source"}, `{"type":"video","source":{"media_ref":"<uploaded-media-ref>"},"caption":"Video"}`},
		{"voice", "Voice recording with an optional caption.", map[string]any{"source": source, "caption": text}, []string{"source"}, `{"type":"voice","source":{"media_ref":"<uploaded-media-ref>"},"caption":"Voice"}`},
		{"file", "File with an optional display filename, MIME type and caption.", map[string]any{"source": source, "caption": text, "filename": text, "mime": text}, []string{"source"}, `{"type":"file","source":{"media_ref":"<uploaded-media-ref>"},"filename":"report.pdf","mime":"application/pdf"}`},
		{"mention", "Mention a platform user in the target session.", map[string]any{"user_id": text, "display": text}, []string{"user_id"}, `{"type":"mention","user_id":"12345","display":"Alex"}`},
		{"quote", "Reply to a confirmed message reference from the same session and instance.", map[string]any{"message_ref": text}, []string{"message_ref"}, `{"type":"quote","message_ref":"<confirmed-message-ref>"}`},
		{"forward", "Forward bot-authored nodes or confirmed messages from the same session and instance; nodes may be mixed.", map[string]any{"nodes": array(map[string]any{"oneOf": []any{object(map[string]any{"parts": array(map[string]any{"$ref": "#/$defs/part"})}, "parts"), object(map[string]any{"message_ref": text}, "message_ref")}})}, []string{"nodes"}, `{"type":"forward","nodes":[{"parts":[{"type":"text","text":"Summary"}]},{"message_ref":"<confirmed-message-ref>"}]}`},
		{"buttons", "Rows of link buttons or visible choices; a choice submits its label as the clicking user's input.", map[string]any{"rows": array(array(button))}, []string{"rows"}, `{"type":"buttons","rows":[[{"label":"Continue"},{"label":"Documentation","url":"https://example.com/docs"}]]}`},
	}
	// 2. 递归节点引用同一组片段 Schema，扩展保持通用载荷形状。
	var alternatives []any
	for _, part := range parts {
		part.fields["type"] = map[string]any{"const": part.name}
		alternatives = append(alternatives, object(part.fields, append([]string{"type"}, part.required...)...))
	}
	// 扩展名称与版本由本轮目标声明校验，递归节点仅描述公共载荷形状。
	alternatives = append(alternatives, object(map[string]any{
		"type":    map[string]any{"type": "string", "description": "A declared extension name from get_message_format"},
		"version": map[string]any{"type": "integer", "minimum": 1},
		"payload": map[string]any{"type": "object"},
	}, "type", "version", "payload"))
	// 3. 每个格式返回独立 Schema 和完整发送例子，平台声明只覆盖支持情况与限制。
	out := make([]MessageFormat, 0, len(parts))
	for i, part := range parts {
		schema := alternatives[i].(map[string]any)
		if part.name == "forward" {
			// 根与递归定义分开，避免在序列化时形成 Go map 环。
			schema = object(part.fields, "type", "nodes")
			schema["$defs"] = map[string]any{"part": map[string]any{"oneOf": alternatives}}
		}
		example := `{"parts":[` + part.example + `]}`
		switch part.name {
		case "text", "markdown", "mention", "quote":
			example = `{"parts":[` + part.example + `,{"type":"image","source":{"https_url":"https://example.com/image.jpg"}}]}`
		case "buttons":
			example = `{"parts":[{"type":"text","text":"Choose an option"},` + part.example + `]}`
		}
		raw, err := json.Marshal(schema)
		if err != nil {
			panic(err)
		} // Schema 仅由固定 JSON 值构造。
		out = append(out, MessageFormat{Name: part.name, Version: 1, Status: SupportUnknown, Description: part.description, SchemaJSON: raw, ExampleJSON: json.RawMessage(example)})
	}
	return out
}
