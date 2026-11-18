package manage

import (
	"cmp"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"golang.org/x/text/language"
)

//go:generate go run changelog_generate.go

const changelogFile = "CHANGELOG.md"
const changelogDirectory = "changelog"
const changelogLocale = "en-US"
const changelogExtension = ".md"

var changelogHeading = regexp.MustCompile(`^## \[([^\]\r\n]+)\](?: - ([^\r\n]+))?[ \t]*$`)

type changelogEntry struct {
	Version  string `json:"version"`
	Date     string `json:"date,omitempty"`
	Markdown string `json:"markdown"`
}

type changelogResponse struct {
	Locale  string           `json:"locale"`
	Entries []changelogEntry `json:"entries"`
}

func (s *Service) packageChangelog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	unlock := s.lock(id)
	defer unlock()
	installed, err := s.opt.Store.GetInstall(r.Context(), id)
	if err != nil {
		respond(w, nil, failure(err))
		return
	}
	if installed.Uninstalled || installed.Source == "dev" {
		respond(w, nil, failure(state.ErrNotFound))
		return
	}
	root, err := os.OpenRoot(s.directory(installed))
	if err != nil {
		respond(w, nil, failure(err))
		return
	}
	defer root.Close()
	result, err := readChangelog(root.FS(), r.Header.Get("Accept-Language"))
	respond(w, result, failure(err))
}

func (s *Service) coreChangelog(w http.ResponseWriter, r *http.Request) {
	respond(w, changelogResponse{Locale: changelogLocale, Entries: splitChangelog(coreChangelogMarkdown)}, nil)
}

func readChangelog(root fs.FS, acceptLanguage string) (changelogResponse, error) {
	result := changelogResponse{Locale: changelogLocale, Entries: []changelogEntry{}}
	raw, err := fs.ReadFile(root, changelogFile)
	if err == nil {
		result.Entries = splitChangelog(string(raw))
		return result, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return result, err
	}
	versions, err := fs.ReadDir(root, changelogDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	// 1. 先确定版本顺序，响应语言取最新版本实际采用的语言。
	slices.SortStableFunc(versions, func(a, b fs.DirEntry) int { return -compareChangelogVersions(a.Name(), b.Name()) })
	tags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
	for _, version := range versions {
		dir := changelogDirectory + "/" + version.Name()
		files, err := fs.ReadDir(root, dir)
		if err != nil {
			return result, err
		}
		locales := make([]string, 0, len(files))
		for _, file := range files {
			locales = append(locales, strings.TrimSuffix(file.Name(), changelogExtension))
		}
		locale := selectChangelogLocale(locales, tags)
		raw, err := fs.ReadFile(root, dir+"/"+locale+changelogExtension)
		if err != nil {
			return result, err
		}
		if len(result.Entries) == 0 {
			result.Locale = locale
		}
		result.Entries = append(result.Entries, changelogEntry{Version: version.Name(), Markdown: string(raw)})
	}
	return result, nil
}

func selectChangelogLocale(locales []string, tags []language.Tag) string {
	for _, tag := range tags {
		locale := tag.String()
		if slices.Contains(locales, locale) {
			return locale
		}
		for _, available := range locales {
			normalized, err := language.Parse(available)
			if err == nil && normalized == tag {
				return available
			}
		}
		base, _, _ := strings.Cut(locale, "-")
		for _, available := range locales {
			primary, _, _ := strings.Cut(language.Make(available).String(), "-")
			if primary == base {
				return available
			}
		}
	}
	return changelogLocale
}

func splitChangelog(markdown string) []changelogEntry {
	entries := []changelogEntry{}
	var current *changelogEntry
	start, offset := 0, 0
	fence := ""
	// 标题只在围栏代码块之外切分；正文切片保留原有换行与空白。
	for _, line := range strings.SplitAfter(markdown, "\n") {
		text := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		trimmed := strings.TrimLeft(text, " ")
		if len(text)-len(trimmed) <= 3 && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			char := trimmed[0]
			count := len(trimmed) - len(strings.TrimLeft(trimmed, string(char)))
			if fence == "" {
				if char != '`' || !strings.Contains(trimmed[count:], "`") {
					fence = trimmed[:count]
				}
			} else if char == fence[0] && count >= len(fence) && strings.TrimSpace(trimmed[count:]) == "" {
				fence = ""
			}
			offset += len(line)
			continue
		}
		if fence != "" {
			offset += len(line)
			continue
		}
		heading := changelogHeading.FindStringSubmatch(text)
		if heading != nil && (heading[1] == "Unreleased" || manifest.ValidVersion(heading[1])) {
			if current != nil {
				current.Markdown = markdown[start:offset]
				entries = append(entries, *current)
			}
			current = nil
			if heading[1] != "Unreleased" {
				current = &changelogEntry{Version: heading[1], Date: strings.TrimSpace(heading[2])}
			}
			start = offset + len(line)
		}
		offset += len(line)
	}
	if current != nil {
		current.Markdown = markdown[start:]
		entries = append(entries, *current)
	}
	slices.SortStableFunc(entries, func(a, b changelogEntry) int { return -compareChangelogVersions(a.Version, b.Version) })
	return entries
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
