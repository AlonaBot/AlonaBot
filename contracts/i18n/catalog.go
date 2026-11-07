// Package i18n 解析组件的多语言文案目录，并按候选语言格式化命名占位符。
//
// [Parse] 检查 i18n/<locale>.json 的 messages、metadata 和 config 结构；
// [Catalog] 保留已编译的文案，可供 Core 与 SDK 并发只读使用。
// 语言选择与文件格式见 spec/CONTRACTS-v0.1.md §5.3、§5.4。
package i18n

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"
)

const sourceLocale = "en-US"
const catalogExtension = ".json"

// Catalog 保存一个组件的全部 locale 文案；由 Parse 构造后只读。
type Catalog struct {
	messages map[string]map[string][]fragment
	locales  []string // 按字典序排列，使主语言回落结果确定。
}

// Problem 标明输入文件中的文案或结构错误。
type Problem struct {
	File    string
	Key     string // 点分隔的 JSON 路径；文件级错误为空。
	Message string
}

// Parse 校验全部文件；有问题时返回 nil Catalog，空输入返回 nil, nil。
// files 的键为规范 BCP 47 标签加 .json，不含目录。
func Parse(files map[string][]byte) (*Catalog, []Problem) {
	if len(files) == 0 {
		return nil, nil
	}
	catalog := &Catalog{messages: make(map[string]map[string][]fragment)}
	var problems []Problem
	// 1. 编译单文件文案，结构问题保留文件名与 JSON 路径。
	for _, file := range slices.Sorted(maps.Keys(files)) {
		locale := strings.TrimSuffix(file, catalogExtension)
		canonical, err := canonicalLocale(locale)
		if !strings.HasSuffix(file, catalogExtension) {
			problems = append(problems, Problem{File: file, Message: "filename must end with .json"})
			continue
		}
		if err != nil {
			problems = append(problems, Problem{File: file, Message: err.Error()})
			continue
		}
		if canonical != locale {
			problems = append(problems, Problem{File: file, Message: fmt.Sprintf("filename must use canonical locale %q", canonical)})
			continue
		}
		// encoding/json 会替换非法 UTF-8；先拒绝它，避免悄悄改变原文。
		if !utf8.Valid(files[file]) {
			problems = append(problems, Problem{File: file, Message: "JSON must contain valid UTF-8"})
			continue
		}
		var document any
		if err := json.Unmarshal(files[file], &document); err != nil {
			problems = append(problems, Problem{File: file, Message: fmt.Sprintf("parse JSON: %v", err)})
			continue
		}
		issue := func(key, message string) {
			problems = append(problems, Problem{File: file, Key: key, Message: message})
		}
		fields, ok := document.(map[string]any)
		if !ok {
			issue("", "expected an object")
		}
		messages := make(map[string][]fragment)
		for _, key := range slices.Sorted(maps.Keys(fields)) {
			switch key {
			case "metadata":
				checkMetadata(fields[key], issue)
			case "config":
				checkFields(fields[key], key, issue)
			case "messages":
				values, ok := fields[key].(map[string]any)
				if !ok {
					issue(key, "expected an object")
					continue
				}
				for _, name := range slices.Sorted(maps.Keys(values)) {
					path := key + "." + name
					if !identifier(name) {
						issue(path, "message key must contain dot-separated ASCII identifiers")
						continue
					}
					text, ok := values[name].(string)
					if !ok {
						issue(path, "expected a string")
						continue
					}
					compiled, err := compile(text)
					if err != nil {
						issue(path, err.Error())
						continue
					}
					messages[name] = compiled
				}
			default:
				issue(key, "unknown top-level key")
			}
		}
		catalog.messages[locale] = messages
		catalog.locales = append(catalog.locales, locale)
	}
	// 2. 以 en-US 为键与占位符集合的唯一来源。
	if _, exists := files[sourceLocale+catalogExtension]; !exists {
		problems = append(problems, Problem{File: sourceLocale + catalogExtension, Message: "source catalog is missing"})
	}
	for _, locale := range catalog.locales {
		if locale == sourceLocale {
			continue
		}
		for _, key := range slices.Sorted(maps.Keys(catalog.messages[locale])) {
			source, exists := catalog.messages[sourceLocale][key]
			message := ""
			if !exists {
				message = "message key is missing from en-US"
			} else if !slices.Equal(names(source), names(catalog.messages[locale][key])) {
				message = "placeholders differ from en-US"
			}
			if message != "" {
				problems = append(problems, Problem{File: locale + catalogExtension, Key: "messages." + key, Message: message})
			}
		}
	}
	if len(problems) != 0 {
		return nil, problems
	}
	return catalog, nil
}

// Has 报告 en-US 是否包含 key。
func (c *Catalog) Has(key string) bool {
	_, exists := c.messages[sourceLocale][key]
	return exists
}

// Format 按 locales 的优先级选择文案并替换参数，最后回落到 en-US。
// 未知键或缺少占位符参数返回错误，多余参数忽略。
func (c *Catalog) Format(locales []string, key string, args map[string]string) (string, error) {
	message, exists := c.messages[sourceLocale][key]
	if !exists {
		return "", fmt.Errorf("unknown message key %q", key)
	}
	locale := c.selectLocale(locales)
	if translated, exists := c.messages[locale][key]; exists {
		message = translated
	}
	var output strings.Builder
	for _, fragment := range message {
		if fragment.name == "" {
			output.WriteString(fragment.text)
			continue
		}
		value, exists := args[fragment.name]
		if !exists {
			return "", fmt.Errorf("message %q is missing argument %q", key, fragment.name)
		}
		output.WriteString(value)
	}
	return output.String(), nil
}

func (c *Catalog) selectLocale(locales []string) string {
	for _, candidate := range locales {
		if _, exists := c.messages[candidate]; exists {
			return candidate
		}
		if candidate == "" {
			continue
		}
		primary, _, _ := strings.Cut(candidate, "-")
		for _, locale := range c.locales {
			language, _, _ := strings.Cut(locale, "-")
			if language == primary {
				return locale
			}
		}
	}
	return sourceLocale
}
