package i18n

import (
	"fmt"
	"regexp"
	"strings"
)

// RFC 5646 §2.1 的 language-tag 语法；重复 variant 与 extension singleton 另行检查。
const languageTag = `^(?:` +
	`(?:[A-Za-z]{2,3}(?:-[A-Za-z]{3}){0,3}|[A-Za-z]{4}|[A-Za-z]{5,8})` +
	`(?:-[A-Za-z]{4})?(?:-(?:[A-Za-z]{2}|[0-9]{3}))?` +
	`(?:-(?:[A-Za-z0-9]{5,8}|[0-9][A-Za-z0-9]{3}))*` +
	`(?:-[0-9A-WY-Za-wy-z](?:-[A-Za-z0-9]{2,8})+)*` +
	`(?:-[xX](?:-[A-Za-z0-9]{1,8})+)?|[xX](?:-[A-Za-z0-9]{1,8})+)$`

var tagSyntax = regexp.MustCompile(languageTag)

// RFC 5646 §2.1 固定的 grandfathered 标签保留其登记大小写。
var grandfathered = map[string]string{
	"en-gb-oed": "en-GB-oed", "i-ami": "i-ami", "i-bnn": "i-bnn",
	"i-default": "i-default", "i-enochian": "i-enochian", "i-hak": "i-hak",
	"i-klingon": "i-klingon", "i-lux": "i-lux", "i-mingo": "i-mingo",
	"i-navajo": "i-navajo", "i-pwn": "i-pwn", "i-tao": "i-tao",
	"i-tay": "i-tay", "i-tsu": "i-tsu", "sgn-be-fr": "sgn-BE-FR",
	"sgn-be-nl": "sgn-BE-NL", "sgn-ch-de": "sgn-CH-DE", "art-lojban": "art-lojban",
	"cel-gaulish": "cel-gaulish", "no-bok": "no-bok", "no-nyn": "no-nyn",
	"zh-guoyu": "zh-guoyu", "zh-hakka": "zh-hakka", "zh-min": "zh-min",
	"zh-min-nan": "zh-min-nan", "zh-xiang": "zh-xiang",
}

// canonicalLocale 检查 BCP 47 语法并按 language/script/region 规范大小写；
// extension 与 private-use 保持小写，规则与 Actor.locale 的大小写约定一致。
func canonicalLocale(locale string) (string, error) {
	lower := strings.ToLower(locale)
	if tag, exists := grandfathered[lower]; exists {
		return tag, nil
	}
	if !tagSyntax.MatchString(locale) {
		return "", fmt.Errorf("invalid BCP 47 tag %q", locale)
	}
	parts := strings.Split(lower, "-")
	if parts[0] == "x" {
		return lower, nil
	}
	variants := make(map[string]bool)
	singletons := make(map[string]bool)
	extension := false
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if len(part) == 1 {
			if part == "x" {
				break // private-use 后的单字符属于私有内容。
			}
			if singletons[part] {
				return "", fmt.Errorf("duplicate extension %q in BCP 47 tag %q", part, locale)
			}
			singletons[part] = true
			extension = true
		}
		if extension {
			continue
		}
		if len(part) >= 5 || len(part) == 4 && part[0] >= '0' && part[0] <= '9' {
			if variants[part] {
				return "", fmt.Errorf("duplicate variant %q in BCP 47 tag %q", part, locale)
			}
			variants[part] = true
		} else if len(part) == 4 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		} else if len(part) == 2 {
			parts[i] = strings.ToUpper(part)
		}
	}
	return strings.Join(parts, "-"), nil
}
