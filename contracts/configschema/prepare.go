package configschema

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Complete 只给缺失的声明字段补默认值或零值，递归补齐对象、列表和模板内的字段。
// 显式值和未知键保持原样；未选择且无默认值的 template 保持缺失。
func (s *Schema) Complete(config map[string]any) {
	completeFields(s.fields, config)
}

// Validate 检查补齐后的配置，失败返回带配置 JSON 路径的 INVALID_ARGUMENT。
// 错误不包含取值；未知键保留，不参与声明字段校验。
func (s *Schema) Validate(config map[string]any) error {
	if problem := objectProblem(s.fields, config, "$"); problem != nil {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "config validation failed at " + problem.Path + ": " + problem.Message}
	}
	return nil
}

func completeFields(fields []*Field, config map[string]any) {
	for _, f := range fields {
		value, present := config[f.name]
		if !present {
			if f.typ == typeTemplate && !f.attributes.has("default") {
				continue
			}
			value = missingValue(f)
			config[f.name] = value
		}
		completeValue(f, value)
	}
}

func missingValue(f *Field) any {
	if f.attributes.has("default") {
		return plainValue(f.attributes.get("default"))
	}
	switch f.typ {
	case typeString, typeText:
		return ""
	case typeInt, typeFloat:
		return zeroNumber
	case typeBool:
		return false
	case typeObject, typeDict:
		return map[string]any{}
	case typeList, typeTemplateList:
		return []any{}
	}
	return nil // template 没有零值，缺失时由调用方保持未选择状态。
}

// completeValue 保留类型不符的输入，交给随后 Validate 定位，避免补值掩盖原始错误。
func completeValue(f *Field, value any) {
	switch f.typ {
	case typeObject:
		if object, ok := value.(map[string]any); ok {
			completeFields(f.items, object)
		}
	case typeList:
		if values, ok := value.([]any); ok && f.attributes.has("items") {
			for _, child := range values {
				if object, ok := child.(map[string]any); ok {
					completeFields(f.items, object)
				}
			}
		}
	case typeTemplate:
		completeTemplate(f, value)
	case typeTemplateList:
		if values, ok := value.([]any); ok {
			for _, child := range values {
				completeTemplate(f, child)
			}
		}
	}
}

func completeTemplate(f *Field, value any) {
	if object, ok := value.(map[string]any); ok {
		for _, template := range f.templates {
			if fields, ok := object[template.name].(map[string]any); ok {
				completeFields(template.items, fields)
			}
		}
	}
}
