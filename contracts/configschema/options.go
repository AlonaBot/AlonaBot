package configschema

import (
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// BuiltinOptionSources 登记内置选项来源与各来源允许的 filter 键。
var BuiltinOptionSources = map[string][]string{
	"core.adapter_instances": {"adapter_id"},
	"core.chats":             {"chat_type", "adapter_id"},
}

// OptionsFrom 保存配置字段声明的来源名与固定筛选条件。
type OptionsFrom struct {
	Source string
	Filter map[string]string
}

// OptionsFrom 按点分隔的字段路径读取选项来源；list 省略下标，模板包含模板键。
func (s *Schema) OptionsFrom(fieldPath string) (OptionsFrom, bool) {
	parts := strings.Split(fieldPath, ".")
	fields := s.fields
	for i := 0; i < len(parts); i++ {
		f := findField(fields, parts[i])
		if f == nil {
			return OptionsFrom{}, false
		}
		if i == len(parts)-1 {
			from, ok := f.attributes.get("options_from").(orderedObject)
			if !ok {
				return OptionsFrom{}, false
			}
			result := OptionsFrom{Source: from.get("source").(string)}
			if filter, ok := from.get("filter").(orderedObject); ok {
				result.Filter = make(map[string]string, len(filter))
				for _, p := range filter {
					result.Filter[p.name] = p.value.(string)
				}
			}
			return result, true
		}
		fields = f.items
		if f.typ == typeTemplate || f.typ == typeTemplateList {
			i++
			fields = nil
			for _, template := range f.templates {
				if template.name == parts[i] {
					fields = template.items
					break
				}
			}
		}
	}
	return OptionsFrom{}, false
}

func checkOptionsFrom(from orderedObject, path string, problems *[]Problem) {
	source, ok := from.get("source").(string)
	keys, builtin := BuiltinOptionSources[source]
	validSource := builtin
	if componentID, name, service := strings.Cut(source, "/"); service && validate.IsServiceName(name) {
		_, err := validate.PluginID(componentID)
		validSource = componentID == "@agent" || err == nil
	}
	if !ok || !validSource {
		*problems = append(*problems, Problem{joinPath(path, "source"), "expected a known Core source or component_id/service_name"})
	}
	for _, p := range from {
		switch p.name {
		case "source":
		case "filter":
			filter, ok := p.value.(orderedObject)
			if !ok {
				*problems = append(*problems, Problem{joinPath(path, "filter"), "expected a filter object"})
				continue
			}
			for _, entry := range filter {
				filterPath := joinPath(joinPath(path, "filter"), entry.name)
				if _, ok := entry.value.(string); !ok {
					*problems = append(*problems, Problem{filterPath, "expected a string filter value"})
				}
				if builtin && !slices.Contains(keys, entry.name) {
					*problems = append(*problems, Problem{filterPath, "unknown filter key for builtin source"})
				}
			}
		default:
			*problems = append(*problems, Problem{joinPath(path, p.name), "unknown options_from attribute"})
		}
	}
}
