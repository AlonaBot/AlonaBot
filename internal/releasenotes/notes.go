// Package releasenotes 从双语 changelog 源文件生成 GitHub Release 正文。
package releasenotes

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
)

var comments = regexp.MustCompile(`(?s)<!--.*?-->`)

// Render 按英文、简体中文分区合并指定版本，去掉版本标题与 HTML 编辑注释。
func Render(root, version string) ([]byte, error) {
	if !manifest.ValidVersion(strings.TrimPrefix(version, "v")) {
		return nil, fmt.Errorf("invalid release version %q", version)
	}
	var sections []string
	for _, locale := range []string{"en", "zh-CN"} {
		filename := filepath.Join(root, version, locale+".md")
		raw, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read release notes: %w", err)
		}
		if !utf8.Valid(raw) {
			return nil, fmt.Errorf("release notes must be UTF-8: %s", filename)
		}
		text := strings.TrimSpace(comments.ReplaceAllString(string(raw), ""))
		heading, body, _ := strings.Cut(text, "\n")
		if strings.HasPrefix(heading, "## ["+strings.TrimPrefix(version, "v")+"] - ") {
			text = strings.TrimSpace(body)
		}
		if text == "" {
			return nil, fmt.Errorf("release notes are empty: %s", filename)
		}
		language := "English"
		if locale == "zh-CN" {
			language = "简体中文"
		}
		sections = append(sections, "## "+language+"\n\n"+text)
	}
	return []byte(strings.Join(sections, "\n\n---\n\n") + "\n"), nil
}
