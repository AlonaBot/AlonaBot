package supervisor

import (
	"encoding/json"
	"strings"
)

type secretPathField struct {
	Type      string
	Items     map[string]secretPathField
	Templates map[string]struct{ Items map[string]secretPathField }
}

// correctedSecretPaths 将历史路径对齐到校正后的声明结构；未知字段的剩余路径保持原样。
func correctedSecretPaths(schema []byte, paths []string) ([]string, error) {
	if len(schema) == 0 {
		return paths, nil
	}
	var fields map[string]secretPathField
	if err := json.Unmarshal(schema, &fields); err != nil {
		return nil, err
	}
	mapped := make([]string, 0, len(paths))
	for _, path := range paths {
		mapped = append(mapped, strings.Join(correctedSecretFields(fields, strings.Split(path, ".")), "."))
	}
	return mapped, nil
}

func correctedSecretFields(fields map[string]secretPathField, path []string) []string {
	if len(path) == 0 {
		return nil
	}
	return append([]string{path[0]}, correctedSecretField(fields[path[0]], path[1:])...)
}

func correctedSecretField(field secretPathField, path []string) []string {
	switch field.Type {
	case "object":
		return correctedSecretFields(field.Items, path)
	case "list", "template_list":
		// list 校正把单元素输入包成数组；既有数组模式只消费一次通配符。
		if len(path) > 0 && path[0] == "*" {
			path = path[1:]
		}
		if field.Type == "list" {
			return append([]string{"*"}, correctedSecretFields(field.Items, path)...)
		}
		field.Type = "template"
		return append([]string{"*"}, correctedSecretField(field, path)...)
	case "template":
		if len(path) == 0 {
			return nil
		}
		return append([]string{path[0]}, correctedSecretFields(field.Templates[path[0]].Items, path[1:])...)
	default:
		return path
	}
}
