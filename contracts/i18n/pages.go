package i18n

import (
	"maps"
	"slices"
)

func parsePages(value any, issue func(string, string)) map[string]string {
	pages, ok := value.(map[string]any)
	if !ok {
		issue("pages", "expected an object")
		return nil
	}
	titles := make(map[string]string, len(pages))
	for _, path := range slices.Sorted(maps.Keys(pages)) {
		text, ok := pages[path].(string)
		if !ok {
			issue("pages."+path, "expected a string title")
			continue
		}
		titles[path] = text
	}
	return titles
}

// PageTitle 按候选语言逐键选择页面标题，缺失覆盖时返回英文源文案。
func (c *Catalog) PageTitle(locales []string, path, source string) string {
	for _, candidate := range locales {
		if candidate == sourceLocale {
			return source
		}
		locale := c.matchLocale(candidate)
		if locale == sourceLocale {
			return source
		}
		if text, exists := c.pages[locale][path]; exists {
			return text
		}
	}
	return source
}
