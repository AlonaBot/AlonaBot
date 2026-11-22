package settings

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

type configField struct {
	Type      string                    `json:"type"`
	Items     map[string]configField    `json:"items"`
	Templates map[string]configTemplate `json:"templates"`
	Secret    bool                      `json:"secret"`
}

type configTemplate struct {
	Items map[string]configField `json:"items"`
}

const adapterGroupDeclaration = `{"package":{"type":"string"},"default_instance":{"type":"string"},"instances":{"type":"dict"}}`

func coreSchema() (*configschema.Schema, map[string]configField, *node, error) {
	raw, err := defaultDocument()
	if err != nil {
		return nil, nil, nil, err
	}
	defaults, err := parseDocument(raw)
	if err != nil {
		return nil, nil, nil, err
	}
	declaration := defaultFields(defaults.root, "")
	raw, err = json.Marshal(declaration)
	if err != nil {
		return nil, nil, nil, err
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) != 0 {
		return nil, nil, nil, fmt.Errorf("invalid Core settings schema: %v", problems)
	}
	var fields map[string]configField
	if err := json.Unmarshal(schema.Canonical(), &fields); err != nil {
		return nil, nil, nil, err
	}
	return schema, fields, defaults.root, nil
}

// 默认值来自 Defaults；这里只补充既有解析器使用的类型与取值约束。
func defaultFields(object *node, path string) map[string]any {
	fields := make(map[string]any)
	for _, field := range object.members {
		if strings.HasPrefix(field.key, "$") {
			continue
		}
		name := field.key
		if path != "" {
			name = path + "." + name
		}
		declaration := map[string]any{"default": json.RawMessage(field.value.encode())}
		switch field.value.delimiter {
		case '{':
			if name == "adapters" || name == "log.sources" || name == "sleep.plugins" {
				declaration["type"] = "dict"
			} else {
				declaration["type"], declaration["items"] = "object", defaultFields(field.value, name)
				delete(declaration, "default")
			}
		case '[':
			declaration["type"] = "list"
		default:
			switch field.value.value.(type) {
			case bool:
				declaration["type"] = "bool"
			case json.Number:
				declaration["type"] = "int"
			default:
				declaration["type"] = "string"
			}
		}
		if name == "log.level" {
			declaration["options"] = []string{"debug", "info", "warn", "error", "critical"}
		}
		if name == "log.format" {
			declaration["options"] = []string{"text", "json"}
		}
		if name == "sleep.mode" {
			declaration["options"] = []string{"always_on", "component_idle"}
		}
		if name == "reply.forward_threshold" || name == "log.file.max_age_days" {
			min := 0
			if name == "log.file.max_age_days" {
				min = 1
			}
			declaration["slider"] = map[string]any{"min": min, "max": json.Number("9223372036854775807")}
		}
		if name == "github_api_base" {
			declaration["description"] = "GitHub Release API base; empty uses the official API."
		}
		if name == "github_download_prefix" {
			declaration["description"] = "Prefix prepended verbatim to the original GitHub archive URL; empty downloads directly."
		}
		if name == "proxy" {
			declaration["secret"] = true // 代理 URL 可以携带认证材料。
			delete(declaration, "default")
		}
		fields[field.key] = declaration
	}
	return fields
}

func correctConfiguration(raw []byte, schema *configschema.Schema) (json.RawMessage, []configschema.Correction, []string, error) {
	var fields map[string]configField
	if err := json.Unmarshal(schema.Canonical(), &fields); err != nil {
		return nil, nil, nil, err
	}
	return correctDeclaredConfiguration(raw, schema, fields)
}

func correctDeclaredConfiguration(raw []byte, schema *configschema.Schema, fields map[string]configField) (json.RawMessage, []configschema.Correction, []string, error) {
	doc, err := parseDocument(raw)
	if err != nil {
		return nil, nil, nil, err
	}
	records := []configschema.Correction{}
	unknown := []string{}
	expandFields(doc.root, fields, "$", &records, &unknown)
	corrected, corrections, err := schema.Correct(doc.root.encode())
	missingEnvironment := make(map[string]bool, len(records))
	for _, record := range records {
		missingEnvironment[record.Path] = true
	}
	for _, correction := range corrections {
		if !missingEnvironment[correction.Path] {
			records = append(records, correction)
		}
	}
	return corrected, records, unknown, err
}

func expandFields(object *node, fields map[string]configField, path string, records *[]configschema.Correction, unknown *[]string) {
	for _, member := range object.members {
		fieldPath := path + "." + member.key
		field, known := fields[member.key]
		if !known {
			*unknown = append(*unknown, path+"["+strconv.Quote(member.key)+"]")
			continue
		}
		expandField(member.value, field, fieldPath, records, unknown)
	}
}

func expandField(value *node, field configField, path string, records *[]configschema.Correction, unknown *[]string) {
	if text, ok := value.value.(string); ok && environmentReference.MatchString(text) {
		resolved, exists := os.LookupEnv(text[2 : len(text)-1])
		if exists {
			value.value = resolved
		} else {
			record := configschema.Correction{Path: path, Reason: "reset"}
			if !field.Secret && !fieldContainsSecret(field) {
				record.Old = value.encode()
			}
			*records = append(*records, record)
			value.value = nil // 交给共享校正器按缺失值补齐 default。
		}
	}
	if field.Type == "object" && value.delimiter == '{' {
		expandFields(value, field.Items, path, records, unknown)
	}
	if field.Type == "template" && value.delimiter == '{' {
		for _, member := range value.members {
			template, known := field.Templates[member.key]
			if !known {
				*unknown = append(*unknown, path+"["+strconv.Quote(member.key)+"]")
			} else if member.value.delimiter == '{' {
				expandFields(member.value, template.Items, path+"."+member.key, records, unknown)
			}
		}
	}
	if field.Type == "dict" {
		expandDictionary(value, path, records)
	}
	if field.Type == "list" || field.Type == "template_list" {
		element := configField{Type: "string"}
		if field.Type == "template_list" {
			element = configField{Type: "template", Templates: field.Templates}
		} else if field.Items != nil {
			element = configField{Type: "object", Items: field.Items}
		}
		if value.delimiter != '[' && (element.Type == "object" || element.Type == "template") {
			expandField(value, element, path, records, unknown)
		}
		for i, child := range value.elements {
			expandField(child, element, fmt.Sprintf("%s[%d]", path, i), records, unknown)
		}
	}
}

func expandDictionary(value *node, path string, records *[]configschema.Correction) {
	if text, ok := value.value.(string); ok && environmentReference.MatchString(text) {
		resolved, exists := os.LookupEnv(text[2 : len(text)-1])
		if exists {
			value.value = resolved
		} else {
			*records = append(*records, configschema.Correction{Path: path, Reason: "reset", Old: value.encode()})
			value.value = nil
		}
	}
	for _, member := range value.members {
		expandDictionary(member.value, path+"["+strconv.Quote(member.key)+"]", records)
	}
	for i, child := range value.elements {
		expandDictionary(child, fmt.Sprintf("%s[%d]", path, i), records)
	}
}

func fieldContainsSecret(field configField) bool {
	if field.Secret {
		return true
	}
	for _, child := range field.Items {
		if fieldContainsSecret(child) {
			return true
		}
	}
	for _, template := range field.Templates {
		for _, child := range template.Items {
			if fieldContainsSecret(child) {
				return true
			}
		}
	}
	return false
}

func logCorrections(records []configschema.Correction) {
	for _, record := range records {
		slog.Warn("configuration corrected", "path", record.Path, "reason", record.Reason)
	}
}

// Prune 删除 Core 声明之外的键，保留其余值与环境变量原文；由 File.Update 提交。
func (d *Document) Prune() error {
	_, fields, _, err := coreSchema()
	if err != nil {
		return err
	}
	pruneFields(d.root, fields)
	if index := d.root.index("adapters"); index >= 0 {
		var groupFields map[string]configField
		if err := json.Unmarshal([]byte(adapterGroupDeclaration), &groupFields); err != nil {
			return err
		}
		for _, group := range d.root.members[index].value.members {
			pruneFields(group.value, groupFields)
		}
	}
	return nil
}

func pruneFields(object *node, fields map[string]configField) {
	object.members = slices.DeleteFunc(object.members, func(member member) bool {
		field, known := fields[member.key]
		if known {
			pruneField(member.value, field)
		}
		return !known
	})
}

func pruneField(value *node, field configField) {
	if field.Type == "object" && value.delimiter == '{' {
		pruneFields(value, field.Items)
	}
	if field.Type == "template" && value.delimiter == '{' {
		value.members = slices.DeleteFunc(value.members, func(member member) bool {
			template, known := field.Templates[member.key]
			if known && member.value.delimiter == '{' {
				pruneFields(member.value, template.Items)
			}
			return !known
		})
	}
	if field.Type == "list" || field.Type == "template_list" {
		if value.delimiter != '[' {
			if field.Type == "template_list" {
				pruneField(value, configField{Type: "template", Templates: field.Templates})
			} else if field.Items != nil {
				pruneField(value, configField{Type: "object", Items: field.Items})
			}
		}
		for _, child := range value.elements {
			element := configField{Type: "object", Items: field.Items}
			if field.Type == "template_list" {
				element = configField{Type: "template", Templates: field.Templates}
			}
			if field.Type == "template_list" || field.Items != nil {
				pruneField(child, element)
			}
		}
	}
}
