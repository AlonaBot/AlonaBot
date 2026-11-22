package pluginkit

import (
	"encoding/json"
	"fmt"
	"maps"
	"path"
	"slices"
)

func exampleContentType(name string) string {
	switch path.Ext(name) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return ""
}

// exampleImageLocation 仅处理已通过 schema 检查的声明，把文件错误定位回示例片段。
func exampleImageLocation(raw []byte, image string) string {
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	var find func(map[string]json.RawMessage, string) string
	find = func(fields map[string]json.RawMessage, parent string) string {
		for _, key := range slices.Sorted(maps.Keys(fields)) {
			var field struct {
				Example          json.RawMessage
				Items, Templates map[string]json.RawMessage
			}
			_ = json.Unmarshal(fields[key], &field)
			location := parent + key
			if len(field.Example) > 0 {
				examples := []json.RawMessage{field.Example}
				array := field.Example[0] == '['
				if array {
					_ = json.Unmarshal(field.Example, &examples)
				}
				for i, raw := range examples {
					var example struct{ Reply json.RawMessage }
					_ = json.Unmarshal(raw, &example)
					if example.Reply[0] != '[' {
						continue
					}
					var segments []struct{ Image string }
					_ = json.Unmarshal(example.Reply, &segments)
					for j, segment := range segments {
						if segment.Image == image {
							suffix := ""
							if array {
								suffix = fmt.Sprintf("[%d]", i)
							}
							return fmt.Sprintf("%s.example%s.reply[%d].image", location, suffix, j)
						}
					}
				}
			}
			if result := find(field.Items, location+".items."); result != "" {
				return result
			}
			if result := find(field.Templates, location+".templates."); result != "" {
				return result
			}
		}
		return ""
	}
	return find(fields, "")
}
