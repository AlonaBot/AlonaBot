package configschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

const (
	correctionConverted = "converted"
	correctionReset     = "reset"
	correctionRemoved   = "removed"
)

// Correction 记录一次配置转换、恢复或模板元素移除；原值只供管理读取。
type Correction struct {
	Path   string          `json:"path"`
	Reason string          `json:"reason"`             // converted、reset 或 removed。
	Old    json.RawMessage `json:"original,omitempty"` // 涉及 secret 时为空。
}

// Correct 补齐并校正完整配置，保留未知键，返回独立 JSON 与校正记录。
// 校正后只检查 required；恢复到类型零值的字段可供管理员随后修订。
// JSON 语法、根对象或 required 错误返回 INVALID_ARGUMENT，消息省略取值。
func (s *Schema) Correct(raw json.RawMessage) (json.RawMessage, []Correction, error) {
	// 1. 保留数字文本，在独立对象上处理，避免校正改写调用方的原始配置。
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, nil, correctionError("$", "expected a JSON object")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, nil, correctionError("$", "expected exactly one JSON document")
	}
	// 2. 逐层校正后检查必填项；合法默认值与类型零值不再触发第二轮恢复。
	corrections := []Correction{}
	correctFields(s.fields, object, "$", &corrections)
	corrected, err := json.Marshal(object)
	if err != nil {
		return nil, corrections, fmt.Errorf("encode corrected config: %w", err)
	}
	if problem := requiredProblem(s.fields, object, "$"); problem != nil {
		return corrected, corrections, correctionError(problem.Path, problem.Message)
	}
	return corrected, corrections, nil
}

func correctionError(path, message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		Message: "config validation failed at " + path + ": " + message}
}

func correctFields(fields []*Field, object map[string]any, path string, corrections *[]Correction) {
	for _, f := range fields {
		value, present := object[f.name]
		if !present || value == nil {
			value = missingValue(f)
		}
		value = correctValue(f, value, joinPath(path, f.name), corrections)
		if f.typ == typeTemplate && value == nil {
			delete(object, f.name)
		} else {
			object[f.name] = value
		}
	}
}

func recordCorrection(f *Field, path, reason string, old any, corrections *[]Correction) {
	correction := Correction{Path: path, Reason: reason}
	// 容器原值也可能包含秘密，省略整份原值以保护嵌套字段。
	if !containsSecret(f) {
		correction.Old, _ = json.Marshal(old) // 输入只含 JSON 解码器产生的值。
	}
	*corrections = append(*corrections, correction)
}

func correctValue(f *Field, value any, path string, corrections *[]Correction) any {
	switch f.typ {
	case typeObject, typeDict:
		object, ok := value.(map[string]any)
		if !ok {
			recordCorrection(f, path, correctionReset, value, corrections)
			object = missingValue(f).(map[string]any)
		}
		if f.typ == typeObject {
			correctFields(f.items, object, path, corrections)
		}
		return object
	case typeTemplate:
		if value == nil {
			return nil
		}
		template, object := selectedTemplate(f, value)
		if template == nil || object == nil {
			recordCorrection(f, path, correctionReset, value, corrections)
			value = missingValue(f)
			if value == nil {
				return nil
			}
			template, object = selectedTemplate(f, value)
		}
		correctFields(template.items, object, joinPath(path, template.name), corrections)
		return value
	case typeList, typeTemplateList:
		values, ok := value.([]any)
		if !ok {
			if listElementConvertible(f, value) {
				recordCorrection(f, path, correctionConverted, value, corrections)
				values = []any{value}
			} else {
				recordCorrection(f, path, correctionReset, value, corrections)
				values = missingValue(f).([]any)
			}
		}
		result := make([]any, 0, len(values))
		for i, child := range values {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			if f.typ == typeTemplateList {
				template, object := selectedTemplate(f, child)
				if template == nil || object == nil {
					recordCorrection(f, childPath, correctionRemoved, child, corrections)
					continue
				}
				correctFields(template.items, object, joinPath(childPath, template.name), corrections)
			} else if f.attributes.has("items") {
				element := &Field{typ: typeObject, items: f.items}
				child = correctValue(element, child, childPath, corrections)
			} else {
				element := &Field{typ: typeString, attributes: orderedObject{}}
				if f.attributes.has("options") {
					element.attributes = append(element.attributes, property{"options", f.attributes.get("options")})
				}
				child = correctValue(element, child, childPath, corrections)
			}
			result = append(result, child)
		}
		return result
	default:
		converted, ok := convertScalar(f.typ, value)
		if !ok || scalarProblem(f, converted) {
			recordCorrection(f, path, correctionReset, value, corrections)
			// 默认值已通过声明校验；同样规范化整数表示，保持重复校正幂等。
			converted, _ = convertScalar(f.typ, missingValue(f))
			return converted
		}
		before, _ := json.Marshal(value)
		after, _ := json.Marshal(converted)
		if !bytes.Equal(before, after) {
			recordCorrection(f, path, correctionConverted, value, corrections)
		}
		return converted
	}
}

func selectedTemplate(f *Field, value any) (*schemaTemplate, map[string]any) {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return nil, nil
	}
	for i := range f.templates {
		if child, found := object[f.templates[i].name]; found {
			fields, _ := child.(map[string]any)
			return &f.templates[i], fields
		}
	}
	return nil, nil
}

func listElementConvertible(f *Field, value any) bool {
	if f.typ == typeTemplateList {
		template, object := selectedTemplate(f, value)
		return template != nil && object != nil
	}
	if f.attributes.has("items") {
		_, ok := value.(map[string]any)
		return ok
	}
	_, ok := convertScalar(typeString, value)
	return ok
}

func scalarProblem(f *Field, value any) bool {
	field := *f
	field.attributes = nil
	for _, attribute := range f.attributes {
		if attribute.name != "required" {
			field.attributes = append(field.attributes, attribute)
		}
	}
	return valueProblem(&field, value, "$") != nil
}

func convertScalar(typ string, value any) (any, bool) {
	switch typ {
	case typeString, typeText:
		switch value := value.(type) {
		case string:
			return value, true
		case json.Number:
			return string(value), true
		case bool:
			return strconv.FormatBool(value), true
		}
	case typeBool:
		if value, ok := value.(bool); ok {
			return value, true
		}
		if value, ok := value.(string); ok && (value == "true" || value == "false") {
			return value == "true", true
		}
	case typeInt, typeFloat:
		number, ok := value.(json.Number)
		if text, isString := value.(string); isString && json.Valid([]byte(text)) {
			decoder := json.NewDecoder(strings.NewReader(text))
			decoder.UseNumber()
			var parsed any
			if err := decoder.Decode(&parsed); err == nil {
				number, ok = parsed.(json.Number)
			}
		}
		if !ok || !validNumber(number, typ) {
			return nil, false
		}
		if typ == typeInt {
			integer, ok := integerText(number)
			return json.Number(integer), ok
		}
		return number, true
	}
	return nil, false
}

// integerText 按有效数字与数量级判断 int64，指数再大也只展开至多 19 位。
func integerText(number json.Number) (string, bool) {
	sign, digits, order := decimalParts(number)
	if sign == 0 {
		return "0", true
	}
	const maxIntegerDigits = 19
	if !order.IsInt64() || order.Int64() < int64(len(digits)) || order.Int64() > maxIntegerDigits {
		return "", false
	}
	text := digits + strings.Repeat("0", int(order.Int64())-len(digits))
	if sign < 0 {
		text = "-" + text
	}
	_, err := strconv.ParseInt(text, 10, 64)
	return text, err == nil
}

func requiredProblem(fields []*Field, object map[string]any, path string) *Problem {
	for _, f := range fields {
		value := object[f.name]
		fieldPath := joinPath(path, f.name)
		empty := value == nil
		switch value := value.(type) {
		case string:
			empty = value == ""
		case []any:
			empty = len(value) == 0
		case map[string]any:
			empty = len(value) == 0
		}
		if required, _ := f.attributes.get("required").(bool); required && empty {
			return &Problem{fieldPath, "required value must not be empty"}
		}
		if f.typ == typeObject {
			if problem := requiredProblem(f.items, value.(map[string]any), fieldPath); problem != nil {
				return problem
			}
		}
		if f.typ == typeTemplate && value != nil {
			template, child := selectedTemplate(f, value)
			if problem := requiredProblem(template.items, child, joinPath(fieldPath, template.name)); problem != nil {
				return problem
			}
		}
		if f.typ == typeList || f.typ == typeTemplateList {
			for i, child := range value.([]any) {
				childPath := fmt.Sprintf("%s[%d]", fieldPath, i)
				fields := f.items
				if f.typ == typeTemplateList {
					template, object := selectedTemplate(f, child)
					fields, child = template.items, object
					childPath = joinPath(childPath, template.name)
				} else if !f.attributes.has("items") {
					continue
				}
				if problem := requiredProblem(fields, child.(map[string]any), childPath); problem != nil {
					return problem
				}
			}
		}
	}
	return nil
}
