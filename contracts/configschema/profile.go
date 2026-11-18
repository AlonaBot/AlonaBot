package configschema

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

func checkProfileFields(fields []*Field, path string, allowed, ancestor bool, problems *[]Problem) {
	for _, f := range fields {
		fieldPath := joinPath(path, f.name)
		declared := f.attributes.has("profile")
		if declared {
			secret, _ := f.attributes.get("secret").(bool)
			if !allowed || secret {
				*problems = append(*problems, Problem{joinPath(fieldPath, "profile"), "profile is only allowed on non-secret top-level fields and object children"})
			}
			if ancestor {
				*problems = append(*problems, Problem{joinPath(fieldPath, "profile"), "profile must not occur on both an ancestor and its descendant"})
			}
		}
		checkProfileFields(f.items, joinPath(fieldPath, "items"), allowed && f.typ == typeObject, ancestor || declared, problems)
		for _, template := range f.templates {
			checkProfileFields(template.items, joinPath(joinPath(joinPath(fieldPath, "templates"), template.name), "items"), false, ancestor || declared, problems)
		}
	}
}

func containsSecret(f *Field) bool {
	var paths [][]string
	secretPaths([]*Field{f}, nil, &paths)
	return len(paths) != 0
}

// ProfilePaths 按声明顺序返回允许按配置文件覆盖的点分字段路径。
func (s *Schema) ProfilePaths() []string {
	paths := []string{}
	var visit func([]*Field, string)
	visit = func(fields []*Field, parent string) {
		for _, f := range fields {
			path := joinPath(parent, f.name)
			if marked, _ := f.attributes.get("profile").(bool); marked {
				paths = append(paths, path)
			}
			if f.typ == typeObject {
				visit(f.items, path)
			}
		}
	}
	visit(s.fields, "")
	return paths
}

// ValidateProfileValue 校验已标记字段的一个完整 JSON 值，不补缺失值。
// 错误包含字段路径，不包含取值（CONTRACTS §5.6）。
func (s *Schema) ValidateProfileValue(path string, raw json.RawMessage) error {
	f := s.fieldAt(strings.Split(path, "."))
	if f == nil {
		return profileValueError(path, "field is not marked for profiles")
	}
	if marked, _ := f.attributes.get("profile").(bool); !marked {
		return profileValueError(path, "field is not marked for profiles")
	}
	// 在值校验处拒绝秘密子树，保证已有复合标记仍可读取并列为无效。
	if containsSecret(f) {
		return profileValueError(path, "field contains secrets")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, problem := readValue(decoder, path)
	if problem != nil {
		return profileValueError(problem.Path, problem.Message)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return profileValueError(path, "expected exactly one JSON value")
	}
	if problem := valueProblem(f, plainValue(value), path); problem != nil {
		return profileValueError(problem.Path, problem.Message)
	}
	return nil
}

func profileValueError(path, message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "profile validation failed at " + path + ": " + message}
}

// ApplyProfile 在独立基础对象中替换已标记字段，保留其余内容，不补默认值。
func (s *Schema) ApplyProfile(raw json.RawMessage, overrides map[string]json.RawMessage) (json.RawMessage, error) {
	var replace func(map[string]json.RawMessage, []string, json.RawMessage) error
	replace = func(object map[string]json.RawMessage, path []string, value json.RawMessage) error {
		if len(path) == 1 {
			object[path[0]] = value
			return nil
		}
		child := map[string]json.RawMessage{}
		if previous, ok := object[path[0]]; ok {
			if err := json.Unmarshal(previous, &child); err != nil {
				return err
			}
		}
		if err := replace(child, path[1:], value); err != nil {
			return err
		}
		encoded, err := json.Marshal(child)
		if err != nil {
			return err
		}
		object[path[0]] = encoded
		return nil
	}
	object := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	for path, value := range overrides {
		if err := s.ValidateProfileValue(path, value); err != nil {
			return nil, err
		}
		if err := replace(object, strings.Split(path, "."), value); err != nil {
			return nil, err
		}
	}
	return json.Marshal(object)
}
