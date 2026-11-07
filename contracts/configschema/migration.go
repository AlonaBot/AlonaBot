package configschema

import (
	"encoding/json"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

const firstMigrationVersion = defaultVersion + 1

type migration struct {
	to       int
	renames  []rename
	mappings []valueMapping
	wraps    [][]string
	removes  [][]string
}

type rename struct{ from, to []string }
type valueMapping struct {
	path         []string
	replacements orderedObject
}

func (s *Schema) parseMigrations(problems *[]Problem) {
	if !s.root.has("$migrations") {
		return
	}
	steps, ok := s.root.get("$migrations").([]any)
	if !ok {
		*problems = append(*problems, Problem{"$migrations", "expected a migration array"})
		return
	}
	for i, step := range steps {
		path := fmt.Sprintf("$migrations[%d]", i)
		attributes, ok := step.(orderedObject)
		if !ok {
			*problems = append(*problems, Problem{path, "expected a migration object"})
			continue
		}
		number, ok := attributes.get("to").(json.Number)
		to, err := strconv.Atoi(string(number))
		if !ok || err != nil || to != i+firstMigrationVersion {
			*problems = append(*problems, Problem{joinPath(path, "to"), "migration versions must be consecutive starting at 2"})
		}
		m := migration{to: to}
		for _, attribute := range attributes {
			attributePath := joinPath(path, attribute.name)
			switch attribute.name {
			case "to":
			case "rename":
				operations, ok := attribute.value.(orderedObject)
				if !ok {
					*problems = append(*problems, Problem{attributePath, "expected an object of source and destination paths"})
					continue
				}
				for _, operation := range operations {
					operationPath := joinPath(attributePath, operation.name)
					from := migrationPath(operation.name, operationPath, problems)
					destination, ok := operation.value.(string)
					if !ok {
						*problems = append(*problems, Problem{operationPath, "expected a destination path string"})
						continue
					}
					to := migrationPath(destination, operationPath, problems)
					if len(to) == 0 || s.fieldAt(to) == nil {
						*problems = append(*problems, Problem{operationPath, "rename destination is not declared in the current schema"})
					}
					if wildcardCount(to) > wildcardCount(from) {
						*problems = append(*problems, Problem{operationPath, "destination wildcard has no source array index"})
					}
					m.renames = append(m.renames, rename{from, to})
				}
			case "map":
				operations, ok := attribute.value.(orderedObject)
				if !ok {
					*problems = append(*problems, Problem{attributePath, "expected an object of value mappings"})
					continue
				}
				for _, operation := range operations {
					operationPath := joinPath(attributePath, operation.name)
					targetPath := migrationPath(operation.name, operationPath, problems)
					target := s.fieldAt(targetPath)
					replacements, ok := operation.value.(orderedObject)
					if !ok {
						*problems = append(*problems, Problem{operationPath, "expected an object of old and new values"})
						continue
					}
					if target == nil {
						*problems = append(*problems, Problem{operationPath, "map destination is not declared in the current schema"})
						continue
					}
					for _, replacement := range replacements {
						_, isString := replacement.value.(string)
						_, isNumber := replacement.value.(json.Number)
						if !isString && !isNumber {
							*problems = append(*problems, Problem{operationPath, "map replacements must be strings or numbers"})
							continue
						}
						if problem := valueProblem(target, replacement.value, operationPath); problem != nil {
							*problems = append(*problems, *problem)
						}
					}
					m.mappings = append(m.mappings, valueMapping{targetPath, replacements})
				}
			case "wrap", "remove":
				operations, ok := attribute.value.([]any)
				if !ok {
					*problems = append(*problems, Problem{attributePath, "expected an array of paths"})
					continue
				}
				for j, operation := range operations {
					operationPath := fmt.Sprintf("%s[%d]", attributePath, j)
					text, ok := operation.(string)
					if !ok {
						*problems = append(*problems, Problem{operationPath, "expected a path string"})
						continue
					}
					targetPath := migrationPath(text, operationPath, problems)
					if attribute.name == "remove" {
						m.removes = append(m.removes, targetPath)
						continue
					}
					target := s.fieldAt(targetPath)
					if target == nil || target.typ != typeList && target.typ != typeTemplateList {
						*problems = append(*problems, Problem{operationPath, "wrap destination must be a declared list or template_list"})
					}
					m.wraps = append(m.wraps, targetPath)
				}
			default:
				*problems = append(*problems, Problem{attributePath, "unknown migration attribute"})
			}
		}
		s.migrations = append(s.migrations, m)
	}
	if len(steps) > 0 && (len(s.migrations) == 0 || s.migrations[len(s.migrations)-1].to != s.version) {
		*problems = append(*problems, Problem{"$migrations", "last migration version must equal $version"})
	}
}

func migrationPath(text, path string, problems *[]Problem) []string {
	parts := strings.Split(text, ".")
	for _, part := range parts {
		if part != "*" && !validKey(part) {
			*problems = append(*problems, Problem{path, "expected a dot-separated field path with optional array wildcards"})
			return nil
		}
	}
	return parts
}

func wildcardCount(path []string) int {
	count := 0
	for _, part := range path {
		if part == "*" {
			count++
		}
	}
	return count
}

// Migrate 在独立副本中按版本执行 rename、map、wrap、remove，成功后才更新 config。
// from 高于 Version 时返回写明两个版本号的 FAILED_PRECONDITION；错误不包含配置取值。
func (s *Schema) Migrate(config map[string]any, from int) error {
	if from > s.version {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: fmt.Sprintf("config version %d is newer than schema version %d", from, s.version)}
	}
	if from == s.version || len(s.migrations) == 0 {
		return nil
	}
	candidate := plainValue(config).(map[string]any)
	for _, step := range s.migrations {
		if step.to <= from {
			continue
		}
		if err := step.apply(candidate); err != nil {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: fmt.Sprintf("config migration from version %d to %d: %s", step.to-1, step.to, err)}
		}
	}
	clear(config)
	for key, value := range candidate {
		config[key] = value
	}
	return nil
}

func (m migration) apply(config map[string]any) error {
	// 1. rename 移动值，目标存在或路径冲突时保留整个迁移前的配置。
	for _, operation := range m.renames {
		_, _, err := transformPath(config, operation.from, nil, nil, func(value any, path []string, indexes []int) (any, bool, error) {
			target := bindIndexes(operation.to, indexes)
			if len(target) > len(path) && slices.Equal(target[:len(path)], path) {
				return nil, false, fmt.Errorf("rename destination at %s is inside its source", configPath(target))
			}
			if err := putDestination(config, operation.to, target, value); err != nil {
				return nil, false, err
			}
			return nil, false, nil
		})
		if err != nil {
			return err
		}
	}
	// 2. map 以实际旧值的类型匹配，数字比较保留精度，替换值已由 Parse 检查。
	for _, operation := range m.mappings {
		_, _, err := transformPath(config, operation.path, nil, nil, func(value any, _ []string, _ []int) (any, bool, error) {
			for _, replacement := range operation.replacements {
				if matchesOldValue(value, replacement.name) {
					return replacement.value, true, nil
				}
			}
			return value, true, nil
		})
		if err != nil {
			return err
		}
	}
	// 3. wrap 保留已有数组，remove 只删除明确给出的路径。
	for _, path := range m.wraps {
		_, _, err := transformPath(config, path, nil, nil, func(value any, _ []string, _ []int) (any, bool, error) {
			if _, ok := value.([]any); ok {
				return value, true, nil
			}
			return []any{value}, true, nil
		})
		if err != nil {
			return err
		}
	}
	for _, path := range m.removes {
		_, _, err := transformPath(config, path, nil, nil, func(any, []string, []int) (any, bool, error) { return nil, false, nil })
		if err != nil {
			return err
		}
	}
	return nil
}

func matchesOldValue(value any, key string) bool {
	if text, ok := value.(string); ok {
		return text == key
	}
	number, ok := numberValue(value)
	if !ok || !json.Valid([]byte(key)) {
		return false
	}
	current, currentOK := new(big.Rat).SetString(string(number))
	old, oldOK := new(big.Rat).SetString(key)
	return currentOK && oldOK && current.Cmp(old) == 0
}
