package configschema

import (
	"fmt"
	"strconv"
	"strings"
)

func (s *Schema) fieldAt(path []string) *field {
	if len(path) == 0 {
		return nil
	}
	current := findField(s.fields, path[0])
	for _, part := range path[1:] {
		if current == nil {
			return nil
		}
		switch current.typ {
		case typeObject:
			current = findField(current.items, part)
		case typeList:
			if part != "*" {
				return nil
			}
			if current.attributes.has("items") {
				current = &field{typ: typeObject, items: current.items}
			} else {
				attributes := orderedObject{}
				if current.attributes.has("options") {
					attributes = append(attributes, property{"options", current.attributes.get("options")})
				}
				current = &field{typ: typeString, attributes: attributes}
			}
		case typeTemplateList:
			if part != "*" {
				return nil
			}
			current = &field{typ: typeTemplate, templates: current.templates}
		case typeTemplate:
			var selected *field
			for _, template := range current.templates {
				if template.name == part {
					selected = &field{typ: typeObject, items: template.items}
					break
				}
			}
			current = selected
		default:
			return nil
		}
	}
	return current
}

// transformPath 只访问实际存在的路径；数组先完成所有下标的变换，再发布删除后的列表。
// visit 返回 false 表示删除该节点，避免遍历中删除元素使后续源下标偏移。
func transformPath(value any, path, actual []string, indexes []int, visit func(any, []string, []int) (any, bool, error)) (any, bool, error) {
	if len(path) == 0 {
		return visit(value, actual, indexes)
	}
	part := path[0]
	if part == "*" {
		values, ok := value.([]any)
		if !ok {
			return value, true, nil
		}
		result := make([]any, 0, len(values))
		for i, child := range values {
			next, keep, err := transformPath(child, path[1:], append(actual, strconv.Itoa(i)), append(indexes, i), visit)
			if err != nil {
				return nil, false, err
			}
			if keep {
				result = append(result, next)
			}
		}
		return result, true, nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return value, true, nil
	}
	child, present := object[part]
	if !present {
		return value, true, nil
	}
	next, keep, err := transformPath(child, path[1:], append(actual, part), indexes, visit)
	if err != nil {
		return nil, false, err
	}
	if keep {
		object[part] = next
	} else {
		delete(object, part)
	}
	return object, true, nil
}

// bindIndexes 将目标中第 n 个 * 绑定到源路径第 n 个数组下标；Parse 已检查数量。
func bindIndexes(path []string, indexes []int) []string {
	bound := append([]string(nil), path...)
	index := 0
	for i, part := range bound {
		if part == "*" {
			bound[i] = strconv.Itoa(indexes[index])
			index++
		}
	}
	return bound
}

func putDestination(root map[string]any, pattern, path []string, value any) error {
	var current any = root
	for i, part := range path {
		if pattern[i] == "*" {
			values, ok := current.([]any)
			index, _ := strconv.Atoi(part) // bindIndexes 只产生十进制的非负数组下标。
			if !ok || index >= len(values) {
				return fmt.Errorf("rename destination array is missing at %s", configPath(path[:i+1]))
			}
			if i == len(path)-1 {
				return fmt.Errorf("rename destination already exists at %s", configPath(path))
			}
			current = values[index]
			continue
		}
		object, ok := current.(map[string]any)
		if !ok {
			return fmt.Errorf("rename destination parent is not an object at %s", configPath(path[:i]))
		}
		child, present := object[part]
		if i == len(path)-1 {
			if present {
				return fmt.Errorf("rename destination already exists at %s", configPath(path))
			}
			object[part] = value
			return nil
		}
		if !present {
			if pattern[i+1] == "*" {
				return fmt.Errorf("rename destination array is missing at %s", configPath(path[:i+1]))
			}
			child = map[string]any{}
			object[part] = child
		}
		current = child
	}
	return nil
}

func configPath(path []string) string {
	if len(path) == 0 {
		return "$"
	}
	return "$." + strings.Join(path, ".")
}
