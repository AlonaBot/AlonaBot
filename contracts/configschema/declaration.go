package configschema

import (
	"encoding/json"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

func parseFields(object orderedObject, path string, root bool, problems *[]Problem) []*field {
	var fields []*field
	for _, p := range object {
		if root && (p.name == "$version" || p.name == "$migrations") {
			continue
		}
		fieldPath := joinPath(path, p.name)
		if !validKey(p.name) {
			*problems = append(*problems, Problem{fieldPath, "expected a nonempty ASCII field key using letters, digits, '_' or '-'"})
			continue
		}
		attributes, ok := p.value.(orderedObject)
		if !ok {
			*problems = append(*problems, Problem{fieldPath, "expected a field description object"})
			continue
		}
		typ, _ := attributes.get("type").(string)
		if !slices.Contains([]string{typeString, typeText, typeInt, typeFloat, typeBool, typeObject, typeList, typeDict, typeTemplate, typeTemplateList}, typ) {
			*problems = append(*problems, Problem{joinPath(fieldPath, "type"), "expected a supported field type"})
			continue
		}
		f := &field{name: p.name, typ: typ, attributes: attributes}
		checkAttributes(f, fieldPath, problems)
		if attributes.has("items") {
			if items, ok := attributes.get("items").(orderedObject); ok && (typ == typeObject || typ == typeList) {
				f.items = parseFields(items, joinPath(fieldPath, "items"), false, problems)
			}
		}
		if templates, ok := attributes.get("templates").(orderedObject); ok && (typ == typeTemplate || typ == typeTemplateList) {
			for _, t := range templates {
				templatePath := joinPath(joinPath(fieldPath, "templates"), t.name)
				if !validKey(t.name) {
					*problems = append(*problems, Problem{templatePath, "expected a nonempty ASCII template key using letters, digits, '_' or '-'"})
					continue
				}
				attributes, ok := t.value.(orderedObject)
				if !ok {
					*problems = append(*problems, Problem{templatePath, "expected a template description object"})
					continue
				}
				checkTemplate(attributes, templatePath, problems)
				template := schemaTemplate{name: t.name, attributes: attributes}
				if items, ok := attributes.get("items").(orderedObject); ok {
					template.items = parseFields(items, joinPath(templatePath, "items"), false, problems)
				}
				f.templates = append(f.templates, template)
			}
		}
		fields = append(fields, f)
	}
	return fields
}

func checkAttributes(f *field, path string, problems *[]Problem) {
	for _, p := range f.attributes {
		attributePath := joinPath(path, p.name)
		var allowed bool
		var valid bool
		switch p.name {
		case "type", "description", "hint":
			allowed = true
			_, valid = p.value.(string)
		case "invisible", "obvious_hint":
			allowed = true
			_, valid = p.value.(bool)
		case "editor_mode", "secret":
			allowed = f.typ == typeString || f.typ == typeText
			_, valid = p.value.(bool)
		case "editor_language":
			allowed = f.typ == typeString || f.typ == typeText
			_, valid = p.value.(string)
		case "required":
			allowed = slices.Contains([]string{typeString, typeText, typeList, typeDict, typeTemplate, typeTemplateList}, f.typ)
			_, valid = p.value.(bool)
		case "default":
			allowed, valid = f.typ != typeObject, true
		case "items":
			allowed = f.typ == typeObject || f.typ == typeList
			_, valid = p.value.(orderedObject)
		case "templates":
			allowed = f.typ == typeTemplate || f.typ == typeTemplateList
			_, valid = p.value.(orderedObject)
		case "options", "labels":
			allowed = f.typ == typeString || f.typ == typeList && !f.attributes.has("items")
			valid = stringArray(p.value)
		case "slider":
			allowed = f.typ == typeInt || f.typ == typeFloat
			_, valid = p.value.(orderedObject)
		case "condition":
			allowed = true
			_, valid = p.value.(orderedObject)
		default:
			*problems = append(*problems, Problem{attributePath, "unknown field attribute"})
			continue
		}
		if !allowed {
			*problems = append(*problems, Problem{attributePath, "attribute does not apply to this field type"})
		} else if !valid {
			*problems = append(*problems, Problem{attributePath, "invalid attribute value type"})
		}
	}
	for _, required := range []struct{ typ, key string }{{typeObject, "items"}, {typeTemplate, "templates"}, {typeTemplateList, "templates"}} {
		if f.typ == required.typ && !f.attributes.has(required.key) {
			*problems = append(*problems, Problem{joinPath(path, required.key), "missing required attribute"})
		}
	}
	if secret, _ := f.attributes.get("secret").(bool); secret {
		for _, key := range []string{"default", "options"} {
			if f.attributes.has(key) {
				*problems = append(*problems, Problem{joinPath(path, key), "secret fields must not declare default or options"})
			}
		}
	}
	if labels, ok := f.attributes.get("labels").([]any); ok {
		options, hasOptions := f.attributes.get("options").([]any)
		if !hasOptions || len(labels) > len(options) {
			*problems = append(*problems, Problem{joinPath(path, "labels"), "labels must not outnumber options"})
		}
	}
	if slider, ok := f.attributes.get("slider").(orderedObject); ok {
		checkSlider(f, slider, joinPath(path, "slider"), problems)
	}
}

func stringArray(value any) bool {
	values, ok := value.([]any)
	if !ok {
		return false
	}
	for _, v := range values {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}

func checkTemplate(attributes orderedObject, path string, problems *[]Problem) {
	for _, p := range attributes {
		valid := false
		switch p.name {
		case "name", "hint", "display_item":
			_, valid = p.value.(string)
		case "hide_hint_in_list":
			_, valid = p.value.(bool)
		case "items":
			_, valid = p.value.(orderedObject)
		default:
			*problems = append(*problems, Problem{joinPath(path, p.name), "unknown template attribute"})
			continue
		}
		if !valid {
			*problems = append(*problems, Problem{joinPath(path, p.name), "invalid template attribute value type"})
		}
	}
	if !attributes.has("items") {
		*problems = append(*problems, Problem{joinPath(path, "items"), "missing required attribute"})
	}
}

func checkSlider(f *field, slider orderedObject, path string, problems *[]Problem) {
	for _, p := range slider {
		if p.name != "min" && p.name != "max" && p.name != "step" {
			*problems = append(*problems, Problem{joinPath(path, p.name), "unknown slider attribute"})
			continue
		}
		if !validNumber(p.value, f.typ) {
			*problems = append(*problems, Problem{joinPath(path, p.name), "expected a number matching the field type"})
		}
	}
	for _, key := range []string{"min", "max"} {
		if !slider.has(key) {
			*problems = append(*problems, Problem{joinPath(path, key), "missing required slider bound"})
		}
	}
	if validNumber(slider.get("min"), f.typ) && validNumber(slider.get("max"), f.typ) && compareNumbers(slider.get("min"), slider.get("max")) >= 0 {
		*problems = append(*problems, Problem{path, "slider min must be less than max"})
	}
}

func checkFields(fields []*field, path string, problems *[]Problem) {
	for _, f := range fields {
		fieldPath := joinPath(path, f.name)
		if f.attributes.has("default") {
			value := plainValue(f.attributes.get("default"))
			completeValue(f, value)
			if problem := valueProblem(f, value, joinPath(fieldPath, "default")); problem != nil {
				*problems = append(*problems, *problem)
			}
		}
		if condition, ok := f.attributes.get("condition").(orderedObject); ok {
			for _, expected := range condition {
				conditionPath := joinPath(joinPath(fieldPath, "condition"), expected.name)
				sibling := findField(fields, expected.name)
				if sibling == nil {
					*problems = append(*problems, Problem{conditionPath, "condition references an undeclared sibling field"})
					continue
				}
				value := plainValue(expected.value)
				completeValue(sibling, value)
				if problem := valueProblem(sibling, value, conditionPath); problem != nil {
					*problems = append(*problems, *problem)
				}
			}
		}
		checkFields(f.items, joinPath(fieldPath, "items"), problems)
		for _, template := range f.templates {
			templatePath := joinPath(joinPath(fieldPath, "templates"), template.name)
			checkFields(template.items, joinPath(templatePath, "items"), problems)
			if display, ok := template.attributes.get("display_item").(string); ok {
				current := template.items
				var selected *field
				parts := strings.Split(display, ".")
				for i, key := range parts {
					selected = findField(current, key)
					if selected == nil {
						break
					}
					if i < len(parts)-1 && selected.typ != typeObject {
						selected = nil
						break
					}
					current = selected.items
				}
				if selected == nil || selected.typ != typeString {
					*problems = append(*problems, Problem{joinPath(templatePath, "display_item"), "display_item must reference a string field through objects"})
				}
			}
		}
	}
}

func findField(fields []*field, name string) *field {
	for _, f := range fields {
		if f.name == name {
			return f
		}
	}
	return nil
}

// plainValue 复制 JSON 集合，并将声明中的有序对象转成配置值。
func plainValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		object := make(map[string]any, len(value))
		for key, child := range value {
			object[key] = plainValue(child)
		}
		return object
	case orderedObject:
		object := make(map[string]any, len(value))
		for _, p := range value {
			object[p.name] = plainValue(p.value)
		}
		return object
	case []any:
		values := make([]any, len(value))
		for i, child := range value {
			values[i] = plainValue(child)
		}
		return values
	default:
		return value
	}
}

func validNumber(value any, typ string) bool {
	number, ok := numberValue(value)
	if !ok {
		return false
	}
	if typ == typeInt {
		integer, ok := new(big.Rat).SetString(string(number))
		return ok && integer.IsInt() && integer.Num().IsInt64()
	}
	decimal, err := number.Float64()
	return err == nil && !math.IsInf(decimal, 0) && !math.IsNaN(decimal)
}

func compareNumbers(a, b any) int {
	aNumber, _ := numberValue(a)
	bNumber, _ := numberValue(b)
	left, _ := new(big.Rat).SetString(string(aNumber))
	right, _ := new(big.Rat).SetString(string(bNumber))
	return left.Cmp(right)
}

// numberValue 接受 JSON Decoder.UseNumber 与 json.Unmarshal 的两种数字表示。
func numberValue(value any) (json.Number, bool) {
	switch value := value.(type) {
	case json.Number:
		return value, true
	case float64:
		return json.Number(strconv.FormatFloat(value, 'g', -1, 64)), true
	default:
		return "", false
	}
}
