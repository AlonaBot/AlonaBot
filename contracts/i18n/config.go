package i18n

import (
	"fmt"
	"maps"
	"slices"
)

// Metadata 按候选语言逐键选择插件名或说明，缺少覆盖时返回 manifest 原文。
func (c *Catalog) Metadata(locales []string, key, source string) string {
	for _, candidate := range locales {
		locale := c.matchLocale(candidate)
		if text, exists := c.metadata[locale][key]; exists {
			return text
		}
		if locale == sourceLocale {
			return source
		}
	}
	return source
}

func checkMetadata(value any, issue func(string, string)) {
	fields, ok := value.(map[string]any)
	if !ok {
		issue("metadata", "expected an object")
		return
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		path := "metadata." + key
		if key != "name" && key != "description" {
			issue(path, "unknown metadata key")
		} else if _, ok := fields[key].(string); !ok {
			issue(path, "expected a string")
		}
	}
}

func checkFields(value any, path string, issue func(string, string)) {
	fields, ok := value.(map[string]any)
	if !ok {
		issue(path, "expected an object")
		return
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		checkOverride(fields[key], path+"."+key, false, issue)
	}
}

// 模板和字段各自允许的覆盖键不同；items 都进入同一字段结构。
func checkOverride(value any, path string, template bool, issue func(string, string)) {
	fields, ok := value.(map[string]any)
	if !ok {
		issue(path, "expected an object")
		return
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		child := path + "." + key
		switch {
		case key == "hint", key == "description" && !template, key == "name" && template:
			if _, ok := fields[key].(string); !ok {
				issue(child, "expected a string")
			}
		case key == "items":
			checkFields(fields[key], child, issue)
		case key == "example" && !template:
			if examples, ok := fields[key].([]any); ok {
				for i, example := range examples {
					checkExample(example, fmt.Sprintf("%s[%d]", child, i), issue)
				}
			} else {
				checkExample(fields[key], child, issue)
			}
		case key == "labels" && !template:
			labels, ok := fields[key].([]any)
			if !ok {
				issue(child, "expected an array of strings")
				continue
			}
			for _, label := range labels {
				if _, ok := label.(string); !ok {
					issue(child, "expected an array of strings")
					break
				}
			}
		case key == "templates" && !template:
			templates, ok := fields[key].(map[string]any)
			if !ok {
				issue(child, "expected an object")
				continue
			}
			for _, name := range slices.Sorted(maps.Keys(templates)) {
				checkOverride(templates[name], child+"."+name, true, issue)
			}
		default:
			issue(child, "unknown override key")
		}
	}
}

// checkExample 检查可缺项的示例文案；源形状、片段位置与占位符由 Schema.CheckI18n 核对。
func checkExample(value any, path string, issue func(string, string)) {
	fields, ok := value.(map[string]any)
	if !ok {
		issue(path, "expected an example translation object")
		return
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		child := path + "." + key
		switch key {
		case "input":
			if _, ok := fields[key].(string); !ok {
				issue(child, "expected example text")
			}
		case "reply":
			if _, ok := fields[key].(string); ok {
				continue
			}
			segments, ok := fields[key].([]any)
			if !ok {
				issue(child, "expected a reply string or segment translation array")
				continue
			}
			for i, segment := range segments {
				location := fmt.Sprintf("%s[%d]", child, i)
				texts, ok := segment.(map[string]any)
				if !ok {
					issue(location, "expected a segment translation object")
					continue
				}
				for _, name := range slices.Sorted(maps.Keys(texts)) {
					if name != "text" && name != "alt" {
						issue(location+"."+name, "only source text or image alt may be translated")
					} else if _, ok := texts[name].(string); !ok {
						issue(location+"."+name, "expected example text")
					}
				}
			}
		default:
			issue(child, "only example input and reply text may be translated")
		}
	}
}
