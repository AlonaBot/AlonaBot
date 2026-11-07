package configschema

import (
	"encoding/json"
	"fmt"
)

const zeroNumber = json.Number("0")

// valueProblem 同时用于 default 和 condition 的合法值检查，错误不包含配置取值。
func valueProblem(f *field, value any, path string) *Problem {
	valid := false
	empty := false
	switch f.typ {
	case typeString, typeText:
		text, ok := value.(string)
		valid, empty = ok, text == ""
		if options, ok := f.attributes.get("options").([]any); ok && valid {
			valid = false
			for _, option := range options {
				if text == option {
					valid = true
					break
				}
			}
		}
	case typeInt, typeFloat:
		valid = validNumber(value, f.typ)
		if slider, ok := f.attributes.get("slider").(orderedObject); ok && valid {
			valid = compareNumbers(value, slider.get("min")) >= 0 && compareNumbers(value, slider.get("max")) <= 0
		}
	case typeBool:
		_, valid = value.(bool)
	case typeObject, typeDict:
		object, ok := value.(map[string]any)
		valid, empty = ok, len(object) == 0
		if ok && f.typ == typeObject {
			if problem := objectProblem(f.items, object, path); problem != nil {
				return problem
			}
		}
	case typeList, typeTemplateList:
		values, ok := value.([]any)
		valid, empty = ok, len(values) == 0
		if ok {
			for i, child := range values {
				childPath := fmt.Sprintf("%s[%d]", path, i)
				if f.typ == typeTemplateList {
					if problem := templateProblem(f, child, childPath); problem != nil {
						return problem
					}
				} else if f.attributes.has("items") {
					object, ok := child.(map[string]any)
					if !ok {
						return &Problem{childPath, "expected an object"}
					}
					if problem := objectProblem(f.items, object, childPath); problem != nil {
						return problem
					}
				} else {
					text, ok := child.(string)
					if !ok {
						return &Problem{childPath, "expected a string"}
					}
					if options, ok := f.attributes.get("options").([]any); ok {
						found := false
						for _, option := range options {
							if text == option {
								found = true
								break
							}
						}
						if !found {
							return &Problem{childPath, "value is outside options"}
						}
					}
				}
			}
		}
	case typeTemplate:
		if problem := templateProblem(f, value, path); problem != nil {
			return problem
		}
		valid = true
	}
	if !valid {
		return &Problem{path, "value does not match the declared type or constraints"}
	}
	if required, _ := f.attributes.get("required").(bool); required && empty {
		return &Problem{path, "required value must not be empty"}
	}
	return nil
}

func templateProblem(f *field, value any, path string) *Problem {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return &Problem{path, "expected exactly one declared template key"}
	}
	for key, child := range object {
		for _, template := range f.templates {
			if template.name != key {
				continue
			}
			fields, ok := child.(map[string]any)
			if !ok {
				return &Problem{joinPath(path, key), "expected a template object"}
			}
			return objectProblem(template.items, fields, joinPath(path, key))
		}
		return &Problem{joinPath(path, key), "undeclared template key"}
	}
	return nil
}

// objectProblem 按准备后的缺失值语义检查组合默认值，避免误拒绝可由子字段默认值补齐的对象。
func objectProblem(fields []*field, object map[string]any, path string) *Problem {
	for _, f := range fields {
		value, present := object[f.name]
		if !present {
			if f.attributes.has("default") {
				value = plainValue(f.attributes.get("default"))
			} else {
				switch f.typ {
				case typeString, typeText:
					value = ""
				case typeInt, typeFloat:
					value = zeroNumber
				case typeBool:
					value = false
				case typeObject, typeDict:
					value = map[string]any{}
				case typeList, typeTemplateList:
					value = []any{}
				case typeTemplate:
					if required, _ := f.attributes.get("required").(bool); required {
						return &Problem{joinPath(path, f.name), "required template must be selected"}
					}
					continue
				}
			}
		}
		if problem := valueProblem(f, value, joinPath(path, f.name)); problem != nil {
			return problem
		}
	}
	return nil
}
