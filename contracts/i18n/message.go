package i18n

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type fragment struct {
	text string
	name string // 非空时表示占位符，text 不参与格式化。
}

// Placeholders 返回文案中去重并按字典序排列的占位符；花括号语法错误时返回错误。
func Placeholders(message string) ([]string, error) {
	fragments, err := compile(message)
	if err != nil {
		return nil, err
	}
	return names(fragments), nil
}

func names(fragments []fragment) []string {
	set := make(map[string]struct{})
	for _, fragment := range fragments {
		if fragment.name != "" {
			set[fragment.name] = struct{}{}
		}
	}
	return slices.Sorted(maps.Keys(set))
}

func identifier(name string) bool {
	for _, part := range strings.Split(name, ".") {
		if part == "" {
			return false
		}
		for _, char := range part {
			if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_') {
				return false
			}
		}
	}
	return true
}

func compile(message string) ([]fragment, error) {
	var fragments []fragment
	var literal strings.Builder
	for i := 0; i < len(message); {
		char := message[i]
		if char != '{' && char != '}' {
			literal.WriteByte(char)
			i++
			continue
		}
		if i+1 < len(message) && message[i+1] == char {
			literal.WriteByte(char)
			i += 2
			continue
		}
		if char == '}' {
			return nil, fmt.Errorf("unmatched closing brace at byte %d", i)
		}
		end := strings.IndexByte(message[i+1:], '}')
		if end < 0 {
			return nil, fmt.Errorf("unmatched opening brace at byte %d", i)
		}
		name := message[i+1 : i+1+end]
		if !identifier(name) {
			return nil, fmt.Errorf("invalid placeholder %q at byte %d", name, i)
		}
		if literal.Len() != 0 {
			fragments = append(fragments, fragment{text: literal.String()})
			literal.Reset()
		}
		fragments = append(fragments, fragment{name: name})
		i += end + 2
	}
	if literal.Len() != 0 {
		fragments = append(fragments, fragment{text: literal.String()})
	}
	return fragments, nil
}
