package pluginkit

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
)

const changelogFile = "CHANGELOG.md"
const changelogDirectory = "changelog"
const changelogSource = "en-US.md"

func checkChangelog(dir, version string) error {
	single, err := os.Lstat(filepath.Join(dir, changelogFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	hasSingle := err == nil
	directory, err := os.Lstat(filepath.Join(dir, changelogDirectory))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	hasDirectory := err == nil
	if hasSingle && hasDirectory {
		return fmt.Errorf("%s and %s must not both exist", changelogFile, changelogDirectory)
	}
	read := func(name string) error {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if !utf8.Valid(raw) {
			return fmt.Errorf("%s: Markdown must contain valid UTF-8", name)
		}
		return nil
	}
	if hasSingle {
		if !single.Mode().IsRegular() {
			return fmt.Errorf("%s: expected a regular file", changelogFile)
		}
		return read(changelogFile)
	}
	if !hasDirectory {
		return nil
	}
	if !directory.IsDir() {
		return fmt.Errorf("%s: expected a directory", changelogDirectory)
	}
	versions, err := os.ReadDir(filepath.Join(dir, changelogDirectory))
	if err != nil {
		return err
	}
	for _, entry := range versions {
		name := filepath.Join(changelogDirectory, entry.Name())
		if !entry.IsDir() || !manifest.ValidVersion(entry.Name()) {
			return fmt.Errorf("%s: expected a SemVer version directory", name)
		}
		if compareChangelogVersions(entry.Name(), version) > 0 {
			return fmt.Errorf("%s: version exceeds manifest version %s", name, version)
		}
		files, err := os.ReadDir(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		source := false
		for _, file := range files {
			path := filepath.Join(name, file.Name())
			locale := strings.TrimSuffix(file.Name(), ".md")
			canonical, err := canonicalChangelogLocale(locale)
			if file.Type() != 0 || !strings.HasSuffix(file.Name(), ".md") || err != nil || canonical != locale {
				return fmt.Errorf("%s: expected a canonical BCP 47 locale Markdown file", path)
			}
			if err := read(path); err != nil {
				return err
			}
			source = source || file.Name() == changelogSource
		}
		if !source {
			return fmt.Errorf("%s: source changelog is missing", filepath.Join(name, changelogSource))
		}
	}
	return nil
}

func compareChangelogVersions(a, b string) int {
	// 1. 先按数值比较主版本、次版本与修订版本。
	left, right := changelogVersionParts(a), changelogVersionParts(b)
	for i := 1; i <= 3; i++ {
		if order := compareChangelogNumber(left[i], right[i]); order != 0 {
			return order
		}
	}
	// 2. 正式版本高于预发布版本；build metadata 不参与版本排序。
	if left[4] == right[4] {
		return 0
	}
	if left[4] == "" {
		return 1
	}
	if right[4] == "" {
		return -1
	}
	// 3. 逐个比较预发布标识符，数字标识符低于文本标识符。
	leftParts, rightParts := strings.Split(left[4], "."), strings.Split(right[4], ".")
	for i := range min(len(leftParts), len(rightParts)) {
		l, r := leftParts[i], rightParts[i]
		leftNumeric, rightNumeric := strings.Trim(l, "0123456789") == "", strings.Trim(r, "0123456789") == ""
		var order int
		switch {
		case leftNumeric && rightNumeric:
			order = compareChangelogNumber(l, r)
		case leftNumeric:
			order = -1
		case rightNumeric:
			order = 1
		default:
			order = strings.Compare(l, r)
		}
		if order != 0 {
			return order
		}
	}
	return cmp.Compare(len(leftParts), len(rightParts))
}

func compareChangelogNumber(a, b string) int {
	// SemVer 数字不含前导零；先比较位数，避免大版本号溢出整数类型。
	if order := cmp.Compare(len(a), len(b)); order != 0 {
		return order
	}
	return strings.Compare(a, b)
}

func changelogVersionParts(version string) []string {
	version, _, _ = strings.Cut(version, "+")
	base, pre, _ := strings.Cut(version, "-")
	return append(append([]string{""}, strings.Split(base, ".")...), pre)
}

// RFC 5646 §2.1 的 language-tag 语法；重复 variant 与 extension singleton 另行检查。
const changelogLanguageTag = `^(?:` +
	`(?:[A-Za-z]{2,3}(?:-[A-Za-z]{3}){0,3}|[A-Za-z]{4}|[A-Za-z]{5,8})` +
	`(?:-[A-Za-z]{4})?(?:-(?:[A-Za-z]{2}|[0-9]{3}))?` +
	`(?:-(?:[A-Za-z0-9]{5,8}|[0-9][A-Za-z0-9]{3}))*` +
	`(?:-[0-9A-WY-Za-wy-z](?:-[A-Za-z0-9]{2,8})+)*` +
	`(?:-[xX](?:-[A-Za-z0-9]{1,8})+)?|[xX](?:-[A-Za-z0-9]{1,8})+)$`

var changelogTagSyntax = regexp.MustCompile(changelogLanguageTag)

// RFC 5646 §2.1 固定的 grandfathered 标签保留其登记大小写。
var changelogGrandfathered = map[string]string{
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

// canonicalChangelogLocale 检查 BCP 47 语法并按 language/script/region 规范大小写；
// extension 与 private-use 保持小写，规则与 Actor.locale 的大小写约定一致。
func canonicalChangelogLocale(locale string) (string, error) {
	lower := strings.ToLower(locale)
	if tag, exists := changelogGrandfathered[lower]; exists {
		return tag, nil
	}
	if !changelogTagSyntax.MatchString(locale) {
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
