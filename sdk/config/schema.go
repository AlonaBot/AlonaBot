package config

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/google/jsonschema-go/jsonschema"
)

const (
	// KeywordSecret 标记 string 或 *string 配置字段中的秘密值。
	KeywordSecret = "x-alonabot-secret"
	// KeywordHidden 标记表单中隐藏的配置字段。
	KeywordHidden = "x-alonabot-hidden"
	// KeywordForm 承载 label、widget、options、option_icons、min、max 表单元数据。
	KeywordForm = "x-alonabot-form"
	// KeywordI18n 保存字段的 config.<json 路径> 翻译路径。
	KeywordI18n = "x-alonabot-i18n"
)

// MarshalSchema 输出紧凑的规范 JSON：properties 遵循 PropertyOrder，其余对象键排序。
func MarshalSchema(s *jsonschema.Schema) ([]byte, error) {
	// 1. 复用库的 Schema 编码与 PropertyOrder，保留其类型和关键字检查。
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshal config schema: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber() // 默认值与扩展关键字中的大整数保持原始精度。
	// 2. 排序 Schema 关键字及普通 JSON 对象，保留 properties 的声明顺序。
	canonical, err := canonicalSchemaJSON(decoder, schemaValue)
	if err != nil {
		return nil, fmt.Errorf("canonicalize config schema: %w", err)
	}
	return canonical, nil
}

// SchemaHash 计算规范 Schema 字节的 SHA256，供 REGISTER 回传。
func SchemaHash(canonical []byte) []byte {
	digest := sha256.Sum256(canonical)
	return digest[:]
}

// schemaPosition 区分 Schema 结构与 default、enum、扩展关键字中的普通 JSON；
// 后者即使含有名为 properties 的键，也按普通对象排序。
type schemaPosition int

const (
	jsonValue schemaPosition = iota
	schemaValue
	schemaMap
	propertyMap
)

func (p schemaPosition) child(key string) schemaPosition {
	switch p {
	case schemaMap, propertyMap:
		return schemaValue
	case schemaValue:
		switch key {
		case "properties":
			return propertyMap
		case "$defs", "definitions", "patternProperties", "dependentSchemas", "dependencies":
			return schemaMap
		case "items", "prefixItems", "additionalItems", "contains", "unevaluatedItems",
			"additionalProperties", "propertyNames", "unevaluatedProperties",
			"allOf", "anyOf", "oneOf", "not", "if", "then", "else", "contentSchema":
			return schemaValue
		}
	}
	return jsonValue
}

func canonicalSchemaJSON(decoder *json.Decoder, position schemaPosition) ([]byte, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	switch token {
	case json.Delim('{'):
		type member struct {
			key   string
			value []byte
		}
		var members []member
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			name := key.(string) // 输入来自 json.Marshal，对象键始终为字符串。
			value, err := canonicalSchemaJSON(decoder, position.child(name))
			if err != nil {
				return nil, err
			}
			members = append(members, member{name, value})
		}
		if position != propertyMap {
			slices.SortStableFunc(members, func(a, b member) int { return cmp.Compare(a.key, b.key) })
		}
		out.WriteByte('{')
		for i, member := range members {
			if i > 0 {
				out.WriteByte(',')
			}
			key, _ := json.Marshal(member.key) // 字符串编码没有失败分支。
			out.Write(key)
			out.WriteByte(':')
			out.Write(member.value)
		}
		out.WriteByte('}')
	case json.Delim('['):
		out.WriteByte('[')
		for decoder.More() {
			value, err := canonicalSchemaJSON(decoder, position)
			if err != nil {
				return nil, err
			}
			if out.Len() > 1 {
				out.WriteByte(',')
			}
			out.Write(value)
		}
		out.WriteByte(']')
	default:
		return json.Marshal(token)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
