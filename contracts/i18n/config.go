package i18n

import (
	"maps"
	"slices"
)

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
