package configschema

import (
	"bytes"
	"encoding/json"
	"io"
)

// CheckI18n 核对翻译文件的 config 根节点与声明同形，问题路径以 config 开头。
// 空 raw 表示未提供配置翻译；检查覆盖键与字段、模板及选项的对应关系。
func (s *Schema) CheckI18n(config json.RawMessage) []Problem {
	if len(config) == 0 {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(config))
	decoder.UseNumber()
	value, problem := readValue(decoder, "config")
	if problem != nil {
		return []Problem{*problem}
	}
	if _, err := decoder.Token(); err != io.EOF {
		return []Problem{{"config", "expected exactly one JSON document"}}
	}
	object, ok := value.(orderedObject)
	if !ok {
		return []Problem{{"config", "expected a config translation object"}}
	}
	var problems []Problem
	checkFieldTranslations(s.fields, object, "config", &problems)
	return problems
}

func checkFieldTranslations(fields []*field, object orderedObject, path string, problems *[]Problem) {
	for _, p := range object {
		fieldPath := joinPath(path, p.name)
		f := findField(fields, p.name)
		if f == nil {
			*problems = append(*problems, Problem{fieldPath, "translation references an undeclared field"})
			continue
		}
		attributes, ok := p.value.(orderedObject)
		if !ok {
			*problems = append(*problems, Problem{fieldPath, "expected a field translation object"})
			continue
		}
		for _, attribute := range attributes {
			attributePath := joinPath(fieldPath, attribute.name)
			switch attribute.name {
			case "description", "hint":
				// 文案类型由翻译目录的结构检查负责，此处只核对与 schema 的对应关系。
			case "labels":
				options, ok := f.attributes.get("options").([]any)
				labels, isArray := attribute.value.([]any)
				if !ok || !isArray || len(labels) > len(options) {
					*problems = append(*problems, Problem{attributePath, "labels must correspond to declared options"})
				}
			case "items":
				items, ok := attribute.value.(orderedObject)
				if !ok || !f.attributes.has("items") {
					*problems = append(*problems, Problem{attributePath, "items must correspond to declared child fields"})
					continue
				}
				checkFieldTranslations(f.items, items, attributePath, problems)
			case "templates":
				templates, ok := attribute.value.(orderedObject)
				if !ok || f.typ != typeTemplate && f.typ != typeTemplateList {
					*problems = append(*problems, Problem{attributePath, "templates must correspond to declared templates"})
					continue
				}
				checkTemplateTranslations(f.templates, templates, attributePath, problems)
			default:
				*problems = append(*problems, Problem{attributePath, "unknown field translation attribute"})
			}
		}
	}
}

func checkTemplateTranslations(templates []schemaTemplate, object orderedObject, path string, problems *[]Problem) {
	for _, p := range object {
		templatePath := joinPath(path, p.name)
		var selected *schemaTemplate
		for i := range templates {
			if templates[i].name == p.name {
				selected = &templates[i]
				break
			}
		}
		if selected == nil {
			*problems = append(*problems, Problem{templatePath, "translation references an undeclared template"})
			continue
		}
		attributes, ok := p.value.(orderedObject)
		if !ok {
			*problems = append(*problems, Problem{templatePath, "expected a template translation object"})
			continue
		}
		for _, attribute := range attributes {
			attributePath := joinPath(templatePath, attribute.name)
			switch attribute.name {
			case "name", "hint":
			case "items":
				items, ok := attribute.value.(orderedObject)
				if !ok {
					*problems = append(*problems, Problem{attributePath, "expected child field translations"})
					continue
				}
				checkFieldTranslations(selected.items, items, attributePath, problems)
			default:
				*problems = append(*problems, Problem{attributePath, "unknown template translation attribute"})
			}
		}
	}
}
