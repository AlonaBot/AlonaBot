// Package release 校验 GitHub Release 的仓库、版本和组件归档资产。
package release

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
)

// GitHubRelease 读取 GitHub Release API 的版本与下载资产。
type GitHubRelease struct {
	Tag    string     `json:"tag_name"`
	URL    string     `json:"html_url"`
	Assets []Download `json:"assets"`
}

// Download 描述 GitHub Release 上的一个附件。
type Download struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

// Validate 核对 Release 标签及页面所属仓库。
func (r GitHubRelease) Validate(repo string) error {
	if !manifest.ValidVersion(strings.TrimPrefix(r.Tag, "v")) {
		return fmt.Errorf("invalid release tag %q", r.Tag)
	}
	u, err := url.Parse(r.URL)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !strings.EqualFold(strings.TrimSuffix(u.Path, "/releases/tag/"+r.Tag), "/"+repo) || !strings.HasSuffix(u.Path, "/releases/tag/"+r.Tag) {
		return fmt.Errorf("release URL does not match repository %s and tag %s: %q", repo, r.Tag, r.URL)
	}
	return nil
}

// Select 选择唯一的平台组件归档，优先 any；缺失或同平台歧义时返回错误。
func (r GitHubRelease) Select(goos, goarch string) (Download, error) {
	for _, suffix := range []string{"-any.tar.gz", "-" + goos + "-" + goarch + ".tar.gz"} {
		var matches []Download
		for _, asset := range r.Assets {
			if strings.HasSuffix(asset.Name, suffix) {
				matches = append(matches, asset)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			return Download{}, fmt.Errorf("multiple component archives match %s: %v", suffix, matches)
		}
	}
	return Download{}, fmt.Errorf("release has no component archive for %s/%s", goos, goarch)
}

// Validate 核对下载地址属于给定仓库、标签和附件名。
func (a Download) Validate(repo, tag string) error {
	u, err := url.Parse(a.URL)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || path.Base(u.Path) != a.Name || !strings.EqualFold(strings.TrimSuffix(u.Path, "/releases/download/"+tag+"/"+a.Name), "/"+repo) || !strings.HasSuffix(u.Path, "/releases/download/"+tag+"/"+a.Name) {
		return fmt.Errorf("asset URL does not match repository %s and tag %s: %q", repo, tag, a.URL)
	}
	return nil
}

// ArchiveName 返回组件归档的规范名称。
func ArchiveName(m *manifest.Manifest) string {
	platform := m.OS + "-" + m.Arch
	if m.OS == "any" && m.Arch == "any" {
		platform = "any"
	}
	return m.ID + "-" + m.Version + "-" + platform + ".tar.gz"
}
