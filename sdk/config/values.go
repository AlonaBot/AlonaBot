package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"slices"
	"strconv"

	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
)

const (
	valueStringType       = "string"
	valueTextType         = "text"
	valueIntType          = "int"
	valueFloatType        = "float"
	valueBoolType         = "bool"
	valueObjectType       = "object"
	valueListType         = "list"
	valueStringsType      = "list of strings"
	valueObjectsType      = "list of objects"
	valueDictType         = "dict"
	valueTemplateType     = "template"
	valueTemplateListType = "template_list"
)

type valueView = dbhook.ConfigView[Values]

// Values 提供已准备配置的只读取值；可并发读取，返回的集合可独立修改。
// 键未声明或取值方法与字段类型不符时 panic，消息包含点分隔的配置路径。
type Values struct {
	valueView
	fields      map[string]valueEntry
	path        string
	templateKey string // 选中模板的键；普通对象或未选择模板时为空。
}

// Template 提供选中模板的键与该模板 items 的只读配置。
// 未选择模板时 Key 返回空串，读取其中任何键都会 panic。
type Template struct{ Values }

type valueEntry struct {
	typ   string
	value any
}

type valueDeclaration struct {
	Type      string                      `json:"type"`
	Items     map[string]valueDeclaration `json:"items"`
	Templates map[string]struct {
		Items map[string]valueDeclaration `json:"items"`
	} `json:"templates"`
}

// String 读取 string 或 text 字段，secret 字段返回已解析的明文。
func (v Values) String(key string) string {
	entry := v.entry(key)
	if entry.typ != valueStringType && entry.typ != valueTextType {
		v.wrongType(key, entry.typ, valueStringType)
	}
	return entry.value.(string)
}

// Int 读取 int 字段，保留 int64 全部范围的精度。
func (v Values) Int(key string) int64 { return v.get(key, valueIntType).(int64) }

// Float 读取有限的 float 字段。
func (v Values) Float(key string) float64 { return v.get(key, valueFloatType).(float64) }

// Bool 读取 bool 字段。
func (v Values) Bool(key string) bool { return v.get(key, valueBoolType).(bool) }

// Strings 读取未声明 items 的 list，返回独立切片。
func (v Values) Strings(key string) []string {
	return slices.Clone(v.get(key, valueStringsType).([]string))
}

// Object 读取 object 字段的 items。
func (v Values) Object(key string) Values { return v.get(key, valueObjectType).(Values) }

// Objects 读取声明了 items 的 list，返回其中各对象的只读配置。
func (v Values) Objects(key string) []Values {
	return slices.Clone(v.get(key, valueObjectsType).([]Values))
}

// Dict 读取 dict 字段，返回独立 map 与原始 JSON 值的副本。
func (v Values) Dict(key string) map[string]json.RawMessage {
	source := v.get(key, valueDictType).(map[string]json.RawMessage)
	result := make(map[string]json.RawMessage, len(source))
	for name, raw := range source {
		result[name] = bytes.Clone(raw)
	}
	return result
}

// Template 读取 template 字段；未选择时返回 Key 为空串的模板。
func (v Values) Template(key string) Template { return v.get(key, valueTemplateType).(Template) }

// TemplateList 读取 template_list 字段，返回各选中模板的独立切片。
func (v Values) TemplateList(key string) []Template {
	return slices.Clone(v.get(key, valueTemplateListType).([]Template))
}

// Key 返回选中模板的键，未选择时为空串。
func (t Template) Key() string { return t.templateKey }

func (v Values) entry(key string) valueEntry {
	if resolved, ok := dbhook.ResolveConfig(&v); ok {
		return resolved.entry(key)
	}
	entry, ok := v.fields[key]
	if !ok {
		panic(fmt.Sprintf("config key %q is not declared", valuePath(v.path, key)))
	}
	return entry
}

func (v Values) get(key, typ string) any {
	entry := v.entry(key)
	if entry.typ != typ {
		v.wrongType(key, entry.typ, typ)
	}
	return entry.value
}

func (v Values) wrongType(key, actual, want string) {
	if resolved, ok := dbhook.ResolveConfig(&v); ok {
		resolved.wrongType(key, actual, want)
		return
	}
	panic(fmt.Sprintf("config key %q is %s, not %s", valuePath(v.path, key), actual, want))
}

func valuePath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// newValues 只消费已通过配置准备的 schema 与完整 JSON；对象解码使取值状态独立于输入字节。
func newValues(schema, raw []byte) (Values, error) {
	var declarations map[string]json.RawMessage
	if len(schema) != 0 {
		if err := json.Unmarshal(schema, &declarations); err != nil {
			return Values{}, fmt.Errorf("decode config declarations: %w", err)
		}
	}
	fields := make(map[string]valueDeclaration, len(declarations))
	for name, raw := range declarations {
		if name == "$version" || name == "$migrations" {
			continue
		}
		var field valueDeclaration
		if err := json.Unmarshal(raw, &field); err != nil {
			return Values{}, fmt.Errorf("decode config declaration %q: %w", name, err)
		}
		fields[name] = field
	}
	return decodeValues(fields, raw, "")
}

func decodeValues(declarations map[string]valueDeclaration, raw []byte, path string) (Values, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return Values{}, fmt.Errorf("decode config object at %q: %w", path, err)
	}
	values := Values{fields: make(map[string]valueEntry, len(declarations)), path: path}
	for key, declaration := range declarations {
		entry, err := decodeValue(declaration, object[key], valuePath(path, key))
		if err != nil {
			return Values{}, fmt.Errorf("decode config key %q: %w", valuePath(path, key), err)
		}
		values.fields[key] = entry
	}
	return values, nil
}

func decodeValue(declaration valueDeclaration, raw json.RawMessage, path string) (valueEntry, error) {
	entry := valueEntry{typ: declaration.Type}
	var err error
	switch declaration.Type {
	case valueIntType:
		number, parseErr := prepareJSON(raw)
		if parseErr != nil {
			return valueEntry{}, parseErr
		}
		integer, _ := new(big.Rat).SetString(string(number.(json.Number))) // 准备阶段已验证整数范围。
		entry.value = integer.Num().Int64()
	case valueFloatType:
		var value float64
		err = json.Unmarshal(raw, &value)
		entry.value = value
	case valueObjectType:
		entry.value, err = decodeValues(declaration.Items, raw, path)
	case valueListType:
		if declaration.Items == nil {
			var values []string
			err = json.Unmarshal(raw, &values)
			entry.typ, entry.value = valueStringsType, values
		} else {
			var objects []json.RawMessage
			if err = json.Unmarshal(raw, &objects); err != nil {
				return valueEntry{}, err
			}
			values := make([]Values, len(objects))
			for i, object := range objects {
				values[i], err = decodeValues(declaration.Items, object, valuePath(path, strconv.Itoa(i)))
				if err != nil {
					return valueEntry{}, err
				}
			}
			entry.typ, entry.value = valueObjectsType, values
		}
	case valueDictType:
		var values map[string]json.RawMessage
		err = json.Unmarshal(raw, &values)
		entry.value = values
	case valueTemplateType:
		entry.value, err = decodeTemplate(declaration, raw, path)
	case valueTemplateListType:
		var objects []json.RawMessage
		if err = json.Unmarshal(raw, &objects); err != nil {
			return valueEntry{}, err
		}
		values := make([]Template, len(objects))
		for i, object := range objects {
			values[i], err = decodeTemplate(declaration, object, valuePath(path, strconv.Itoa(i)))
			if err != nil {
				return valueEntry{}, err
			}
		}
		entry.value = values
	default:
		entry.value, err = prepareJSON(raw)
	}
	return entry, err
}

func decodeTemplate(declaration valueDeclaration, raw json.RawMessage, path string) (Template, error) {
	if len(raw) == 0 {
		return Template{Values: Values{path: path}}, nil
	}
	var selection map[string]json.RawMessage
	if err := json.Unmarshal(raw, &selection); err != nil {
		return Template{}, err
	}
	for key, raw := range selection {
		values, err := decodeValues(declaration.Templates[key].Items, raw, valuePath(path, key))
		if err != nil {
			return Template{}, err
		}
		values.templateKey = key
		return Template{Values: values}, nil
	}
	return Template{Values: Values{path: path}}, nil
}
