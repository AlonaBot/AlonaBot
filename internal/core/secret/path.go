package secret

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

func decodeObject(raw []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var object map[string]any
	// JSON 来自插件或持久化快照；错误仅报告路径，不带可能含秘密的输入片段。
	if !json.Valid(raw) {
		return nil, pathError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, nil, "invalid JSON")
	}
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, pathError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, nil, "expected JSON object")
	}
	return object, nil
}

func reference(value any, path []string) (string, bool, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}
	ref, found := object[RefKey]
	if !found {
		return "", false, nil
	}
	id, ok := ref.(string)
	if !ok || id == "" || len(object) != 1 {
		return "", false, pathError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, path, "invalid secret reference")
	}
	return id, true, nil
}

// walk 在进入子节点前调用 visit；替换节点后不再遍历其内容。
func walk(value any, path []string, visit func([]string, any) (any, bool, error)) (any, error) {
	replacement, replaced, err := visit(path, value)
	if err != nil {
		return nil, err
	}
	if replaced {
		return replacement, nil
	}
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			next, err := walk(child, append(path, key), visit)
			if err != nil {
				return nil, err
			}
			current[key] = next
		}
	case []any:
		for i, child := range current {
			next, err := walk(child, append(path, strconv.Itoa(i)), visit)
			if err != nil {
				return nil, err
			}
			current[i] = next
		}
	}
	return value, nil
}

func escape(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "~", "~0"), ".", "~1")
}

func displayPath(path []string) string {
	parts := make([]string, len(path))
	for i, part := range path {
		parts[i] = escape(part)
	}
	if len(parts) == 0 {
		return "$"
	}
	return strings.Join(parts, ".")
}

func matches(pattern, path []string) bool {
	if len(pattern) != len(path) {
		return false
	}
	for i, part := range pattern {
		if part != "*" && part != escape(path[i]) {
			return false
		}
	}
	return true
}

func at(value any, path []string) any {
	for _, part := range path {
		switch current := value.(type) {
		case map[string]any:
			value = current[part]
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(current) {
				return nil
			}
			value = current[i]
		default:
			return nil
		}
	}
	return value
}
