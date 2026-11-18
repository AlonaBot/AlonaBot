package configschema

import (
	"encoding/json"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

func parseFields(object orderedObject, path string, root bool, problems *[]Problem) []*Field {
	var fields []*Field
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
		f := &Field{name: p.name, typ: typ, attributes: attributes}
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

func checkAttributes(f *Field, path string, problems *[]Problem) {
	for _, p := range f.attributes {
		attributePath := joinPath(path, p.name)
		var allowed bool
		var valid bool
		switch p.name {
		case "type", "description", "hint":
			allowed = true
			_, valid = p.value.(string)
		case "invisible", "obvious_hint", "profile":
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
		case "options_from":
			allowed = f.typ == typeString || f.typ == typeList && !f.attributes.has("items")
			_, valid = p.value.(orderedObject)
		case "slider":
			allowed = f.typ == typeInt || f.typ == typeFloat
			_, valid = p.value.(orderedObject)
		case "example":
			allowed, valid = true, true
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
	if from, ok := f.attributes.get("options_from").(orderedObject); ok {
		checkOptionsFrom(from, joinPath(path, "options_from"), problems)
	}
	if f.attributes.has("options_from") {
		for _, key := range []string{"options", "labels", "secret"} {
			if f.attributes.has(key) {
				*problems = append(*problems, Problem{joinPath(path, key), "options_from must not coexist with options, labels or secret"})
			}
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

func checkSlider(f *Field, slider orderedObject, path string, problems *[]Problem) {
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

func checkFields(fields []*Field, path string, problems *[]Problem) {
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
				typed := conditionField(sibling)
				completeValue(typed, value)
				if problem := valueProblem(typed, value, conditionPath); problem != nil {
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
				var selected *Field
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

// conditionField 只保留类型和嵌套形状，让显示条件复用类型检查而不继承取值约束。
func conditionField(f *Field) *Field {
	typed := &Field{name: f.name, typ: f.typ}
	if f.attributes.has("items") {
		typed.attributes = orderedObject{{"items", f.attributes.get("items")}}
	}
	for _, child := range f.items {
		typed.items = append(typed.items, conditionField(child))
	}
	for _, template := range f.templates {
		projection := schemaTemplate{name: template.name}
		for _, child := range template.items {
			projection.items = append(projection.items, conditionField(child))
		}
		typed.templates = append(typed.templates, projection)
	}
	return typed
}

func findField(fields []*Field, name string) *Field {
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
	leftSign, leftDigits, leftOrder := decimalParts(aNumber)
	rightSign, rightDigits, rightOrder := decimalParts(bNumber)
	if leftSign != rightSign {
		return leftSign - rightSign
	}
	if leftSign == 0 {
		return 0
	}
	if order := leftOrder.Cmp(rightOrder); order != 0 {
		return leftSign * order
	}
	// 有效数字已去掉尾零；数量级相同时按字典序比较，相当于在短串后补零。
	return leftSign * strings.Compare(leftDigits, rightDigits)
}

// decimalParts 拆分合法 JSON 数字；order 是小数点相对有效数字首位的位置。
// 指数单独存储，避免极小数展开成巨大的十进制分母。
func decimalParts(number json.Number) (sign int, digits string, order *big.Int) {
	// 1. 分离符号与指数，指数的长度只取决于输入文本。
	text := string(number)
	sign = 1
	if strings.HasPrefix(text, "-") {
		sign, text = -1, text[1:]
	}
	order = new(big.Int)
	if exponent := strings.IndexAny(text, "eE"); exponent >= 0 {
		order.SetString(text[exponent+1:], 10) // 输入已通过 JSON 数字语法检查。
		text = text[:exponent]
	}
	// 2. 去掉小数点和无效的零，用有效数字长度修正数量级。
	fraction := 0
	if point := strings.IndexByte(text, '.'); point >= 0 {
		fraction = len(text) - point - 1
		text = text[:point] + text[point+1:]
	}
	digits = strings.TrimLeft(text, "0")
	if digits == "" {
		return 0, "", order
	}
	order.Add(order, big.NewInt(int64(len(digits)-fraction)))
	return sign, strings.TrimRight(digits, "0"), order
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
