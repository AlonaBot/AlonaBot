package manage

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
)

type configDeclaration struct {
	Type      string
	Items     map[string]configDeclaration
	Templates map[string]struct{ Items map[string]configDeclaration }
}

// unknownConfig 只列出或删除声明之外的键；dict 的任意键均有效。
func unknownConfig(schema, snapshot []byte, prune bool) (json.RawMessage, []string, error) {
	var fields map[string]configDeclaration
	if len(schema) != 0 {
		if err := json.Unmarshal(schema, &fields); err != nil {
			return nil, nil, err
		}
	}
	var object map[string]any
	if err := decodeConfigObject(snapshot, &object); err != nil {
		return nil, nil, err
	}
	unknown := []string{}
	unknownFields(fields, object, "$", prune, &unknown)
	raw, err := json.Marshal(object)
	return raw, unknown, err
}

func unknownFields(fields map[string]configDeclaration, object map[string]any, path string, prune bool, unknown *[]string) {
	for _, key := range slices.Sorted(maps.Keys(object)) {
		field, exists := fields[key]
		if !exists {
			*unknown = append(*unknown, path+"["+strconv.Quote(key)+"]")
			if prune {
				delete(object, key)
			}
			continue
		}
		unknownValue(field, object[key], path+"."+key, prune, unknown)
	}
}

func unknownValue(field configDeclaration, value any, path string, prune bool, unknown *[]string) {
	switch field.Type {
	case "object":
		if object, ok := value.(map[string]any); ok {
			unknownFields(field.Items, object, path, prune, unknown)
		}
	case "template":
		if object, ok := value.(map[string]any); ok {
			for _, key := range slices.Sorted(maps.Keys(object)) {
				template, exists := field.Templates[key]
				if !exists {
					*unknown = append(*unknown, path+"["+strconv.Quote(key)+"]")
					if prune {
						delete(object, key)
					}
				} else if child, ok := object[key].(map[string]any); ok {
					unknownFields(template.Items, child, path+"."+key, prune, unknown)
				}
			}
		}
	case "list", "template_list":
		if children, ok := value.([]any); ok {
			element := configDeclaration{Type: "object", Items: field.Items}
			if field.Type == "template_list" {
				element = configDeclaration{Type: "template", Templates: field.Templates}
			} else if field.Items == nil {
				return
			}
			for i, child := range children {
				unknownValue(element, child, fmt.Sprintf("%s[%d]", path, i), prune, unknown)
			}
		}
	}
}

func (s *Service) pruneConfig(ctx context.Context, id string) (*pendingConfig, error) {
	unlock := s.lock(id)
	defer unlock()
	schema, _, err := s.configResources(ctx, id)
	if err != nil {
		return nil, err
	}
	active, snapshot, _, err := s.configSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	pruned, _, err := unknownConfig(schema, snapshot, true)
	if err != nil {
		return nil, err
	}
	if _, err := s.configurations().Replace(ctx, id, active, pruned); err != nil {
		return nil, err
	}
	redacted, err := s.redactConfig(ctx, active, schema, pruned)
	return &pendingConfig{BaseGeneration: active, Config: redacted}, err
}
