package configschema

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	s.checkFieldTranslations(s.fields, object, "config", &problems)
	return problems
}

func (s *Schema) checkExampleTranslation(source, translated any, path string, problems *[]Problem) {
	if examples, ok := source.([]any); ok {
		translations, ok := translated.([]any)
		if !ok || len(translations) > len(examples) {
			*problems = append(*problems, Problem{path, "example translations must match the source example array"})
			return
		}
		for i, translation := range translations {
			s.checkExampleTranslation(examples[i], translation, fmt.Sprintf("%s[%d]", path, i), problems)
		}
		return
	}
	example, sourceOK := source.(orderedObject)
	translation, ok := translated.(orderedObject)
	if !sourceOK || !ok {
		*problems = append(*problems, Problem{path, "example translation must match a declared example object"})
		return
	}
	for _, attribute := range translation {
		attributePath := joinPath(path, attribute.name)
		switch attribute.name {
		case "input":
			s.exampleText(attribute.value, attributePath, problems)
		case "reply":
			if _, textReply := example.get("reply").(string); textReply {
				s.exampleText(attribute.value, attributePath, problems)
				continue
			}
			segments := example.get("reply").([]any)
			overrides, ok := attribute.value.([]any)
			if !ok || len(overrides) > len(segments) {
				*problems = append(*problems, Problem{attributePath, "reply translation must match the source segment array"})
				continue
			}
			for i, override := range overrides {
				segmentPath := fmt.Sprintf("%s[%d]", attributePath, i)
				segment, ok := override.(orderedObject)
				if !ok {
					*problems = append(*problems, Problem{segmentPath, "expected a segment translation object"})
					continue
				}
				sourceSegment := segments[i].(orderedObject)
				for _, text := range segment {
					location := joinPath(segmentPath, text.name)
					switch {
					case text.name == "text" && sourceSegment.has("text"):
						s.exampleText(text.value, location, problems)
					case text.name == "alt" && sourceSegment.has("image"):
						if _, ok := text.value.(string); !ok {
							*problems = append(*problems, Problem{location, "expected image alternative text"})
						}
					default:
						*problems = append(*problems, Problem{location, "only source text or image alt may be translated"})
					}
				}
			}
		default:
			*problems = append(*problems, Problem{attributePath, "only example input and reply text may be translated"})
		}
	}
}

func (s *Schema) checkFieldTranslations(fields []*Field, object orderedObject, path string, problems *[]Problem) {
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
			case "example":
				s.checkExampleTranslation(f.attributes.get("example"), attribute.value, attributePath, problems)
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
				s.checkFieldTranslations(f.items, items, attributePath, problems)
			case "templates":
				templates, ok := attribute.value.(orderedObject)
				if !ok || f.typ != typeTemplate && f.typ != typeTemplateList {
					*problems = append(*problems, Problem{attributePath, "templates must correspond to declared templates"})
					continue
				}
				s.checkTemplateTranslations(f.templates, templates, attributePath, problems)
			default:
				*problems = append(*problems, Problem{attributePath, "unknown field translation attribute"})
			}
		}
	}
}

func (s *Schema) checkTemplateTranslations(templates []schemaTemplate, object orderedObject, path string, problems *[]Problem) {
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
				s.checkFieldTranslations(selected.items, items, attributePath, problems)
			default:
				*problems = append(*problems, Problem{attributePath, "unknown template translation attribute"})
			}
		}
	}
}
