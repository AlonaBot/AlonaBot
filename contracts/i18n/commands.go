package i18n

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func parseCommands(value any, issue func(string, string)) map[string]string {
	commands, ok := value.(map[string]any)
	if !ok {
		issue("commands", "expected an object")
		return nil
	}
	descriptions := make(map[string]string)
	for _, path := range slices.Sorted(maps.Keys(commands)) {
		fields, ok := commands[path].(map[string]any)
		if !ok {
			issue("commands."+path, "expected an object")
			continue
		}
		text, ok := fields["description"].(string)
		if !ok {
			issue("commands."+path+".description", "expected a string description")
			continue
		}
		descriptions[path] = text
	}
	return descriptions
}

// Locales 返回目录内语言的字典序副本。
func (c *Catalog) Locales() []string {
	return slices.Clone(c.locales)
}

// CheckCommands 检查覆盖路径对应本次注册的规范命令，并已登记英文说明。
func (c *Catalog) CheckCommands(commands map[string]string) []Problem {
	var problems []Problem
	for _, locale := range c.locales {
		for _, path := range slices.Sorted(maps.Keys(c.commands[locale])) {
			source, exists := commands[path]
			message := ""
			if !exists {
				message = fmt.Sprintf("command path %q is not registered", path)
			} else if source == "" {
				message = fmt.Sprintf("command path %q has no English description", path)
			}
			if message != "" {
				problems = append(problems, Problem{File: locale + catalogExtension, Key: "commands." + path + ".description", Message: message})
			}
		}
	}
	return problems
}

// CommandDescription 按候选语言逐键选择普通文本说明，最后返回英文原文。
func (c *Catalog) CommandDescription(locales []string, path []string, source string) string {
	if source == "" {
		return ""
	}
	key := strings.Join(path, " ")
	for _, candidate := range locales {
		if candidate == sourceLocale {
			return source
		}
		locale := c.matchLocale(candidate)
		if locale == sourceLocale {
			return source
		}
		if text, exists := c.commands[locale][key]; exists {
			return text
		}
	}
	return source
}
