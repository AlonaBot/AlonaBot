// Package configschema 解析并检查组件的手写配置声明。
//
// Schema 保留 JSON 键顺序与数字原文，供生成器、SDK 和 Core 共用。
// Parse 检查声明，Canonical 与 Hash 提供构建和 REGISTER 使用的规范内容及摘要。
// Migrate 按声明版本转换旧配置，Complete 与 Validate 随后补值和校验。
// SecretPaths 与 CheckI18n 分别提供秘密字段路径与翻译核对。
// 配置格式见 spec/CONTRACTS-v0.1.md §5.1、§5.3 与 §5.5。
package configschema

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

const (
	typeString       = "string"
	typeText         = "text"
	typeInt          = "int"
	typeFloat        = "float"
	typeBool         = "bool"
	typeObject       = "object"
	typeList         = "list"
	typeDict         = "dict"
	typeTemplate     = "template"
	typeTemplateList = "template_list"
	defaultVersion   = 1
)

// Schema 保存已检查的配置声明；所有字段均由 Parse 初始化。
type Schema struct {
	root       orderedObject
	fields     []*field
	canonical  []byte
	version    int
	migrations []migration
}

// Problem 定位 schema 中的一项错误，Message 使用英文。
type Problem struct {
	Path    string // 点分隔的 schema 路径；根节点为 $。
	Message string
}

type field struct {
	name       string
	typ        string
	attributes orderedObject
	items      []*field
	templates  []schemaTemplate
}

type schemaTemplate struct {
	name       string
	attributes orderedObject
	items      []*field
}

// Parse 检查配置声明；有错误时返回 nil 和带 schema 路径的问题列表。
func Parse(raw []byte) (*Schema, []Problem) {
	// 1. 有序解码，拒绝重复键和尾随内容，避免摘要与字段解释不一致。
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, problem := readValue(decoder, "")
	if problem != nil {
		if problem.Path == "" {
			problem.Path = "$"
		}
		return nil, []Problem{*problem}
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, []Problem{{"$", "expected exactly one JSON document"}}
	}
	root, ok := value.(orderedObject)
	if !ok {
		return nil, []Problem{{"$", "expected a JSON object"}}
	}
	schema := &Schema{root: root, version: defaultVersion}
	var problems []Problem
	if version := root.get("$version"); root.has("$version") {
		number, ok := version.(json.Number)
		n, err := strconv.Atoi(string(number))
		if !ok || err != nil || n <= 0 {
			problems = append(problems, Problem{"$version", "expected a positive integer"})
		} else {
			schema.version = n
		}
	}
	// 2. 先建立同层字段，再核对条件、默认值和模板引用。
	schema.fields = parseFields(root, "", true, &problems)
	if len(problems) == 0 {
		checkProfileFields(schema.fields, "", true, false, &problems)
		checkFields(schema.fields, "", &problems)
		schema.parseMigrations(&problems)
	}
	if len(problems) != 0 {
		return nil, problems
	}
	schema.canonical = appendCanonical(nil, root)
	return schema, nil
}

// Canonical 返回保留键顺序与数字原文的紧凑 JSON 副本。
func (s *Schema) Canonical() []byte { return bytes.Clone(s.canonical) }

// Hash 返回 Canonical 的 SHA-256，供 REGISTER 的 config_schema_hash 使用。
func (s *Schema) Hash() []byte {
	sum := sha256.Sum256(s.canonical)
	return sum[:]
}

// Version 返回声明的 $version，缺省为 1。
func (s *Schema) Version() int { return s.version }

type property struct {
	name  string
	value any
}
type orderedObject []property

func (o orderedObject) get(name string) any {
	for _, p := range o {
		if p.name == name {
			return p.value
		}
	}
	return nil
}

func (o orderedObject) has(name string) bool {
	for _, p := range o {
		if p.name == name {
			return true
		}
	}
	return false
}

func readValue(decoder *json.Decoder, path string) (any, *Problem) {
	token, err := decoder.Token()
	if err != nil {
		return nil, &Problem{path, "invalid JSON"}
	}
	switch token {
	case json.Delim('{'):
		object := orderedObject{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return nil, &Problem{path, "invalid object key"}
			}
			name := key.(string) // Decoder 已检查对象的键必须是字符串。
			if object.has(name) {
				return nil, &Problem{joinPath(path, name), "duplicate key"}
			}
			value, problem := readValue(decoder, joinPath(path, name))
			if problem != nil {
				return nil, problem
			}
			object = append(object, property{name, value})
		}
		if _, err := decoder.Token(); err != nil {
			return nil, &Problem{path, "invalid object"}
		}
		return object, nil
	case json.Delim('['):
		values := []any{}
		for decoder.More() {
			value, err := readValue(decoder, fmt.Sprintf("%s[%d]", path, len(values)))
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		if _, err := decoder.Token(); err != nil {
			return nil, &Problem{path, "invalid array"}
		}
		return values, nil
	default:
		return token, nil
	}
}

func appendCanonical(dst []byte, value any) []byte {
	switch value := value.(type) {
	case orderedObject:
		dst = append(dst, '{')
		for i, p := range value {
			if i != 0 {
				dst = append(dst, ',')
			}
			dst = appendCanonical(dst, p.name)
			dst = append(dst, ':')
			dst = appendCanonical(dst, p.value)
		}
		return append(dst, '}')
	case []any:
		dst = append(dst, '[')
		for i, child := range value {
			if i != 0 {
				dst = append(dst, ',')
			}
			dst = appendCanonical(dst, child)
		}
		return append(dst, ']')
	default:
		raw, _ := json.Marshal(value) // 解析器只产生 JSON 标量，编码不会失败。
		return append(dst, raw...)
	}
}

func joinPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

func validKey(key string) bool {
	if key == "" {
		return false
	}
	for _, c := range key {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// SecretPaths 返回配置中的 secret 路径模式；* 匹配任意数组下标，模板键占一段。
func (s *Schema) SecretPaths() [][]string {
	var paths [][]string
	secretPaths(s.fields, nil, &paths)
	return paths
}

func secretPaths(fields []*field, parent []string, paths *[][]string) {
	for _, f := range fields {
		path := append(append([]string(nil), parent...), f.name)
		if secret, _ := f.attributes.get("secret").(bool); secret {
			*paths = append(*paths, path)
		}
		switch f.typ {
		case typeObject:
			secretPaths(f.items, path, paths)
		case typeList:
			secretPaths(f.items, append(path, "*"), paths)
		case typeTemplate, typeTemplateList:
			if f.typ == typeTemplateList {
				path = append(path, "*")
			}
			for _, template := range f.templates {
				secretPaths(template.items, append(path, template.name), paths)
			}
		}
	}
}
