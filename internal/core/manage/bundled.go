package manage

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

// BundledPackage 定位发行物中指定组件的唯一平台归档。
func BundledPackage(dir, id string) (InstallRequest, error) {
	req := InstallRequest{ID: id}
	if _, err := validate.PluginID(id); err != nil {
		return req, err
	}
	directory := filepath.Join(dir, id)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return req, fmt.Errorf("read bundled %s: %w", id, err)
	}
	publication := release.GitHubRelease{}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			publication.Assets = append(publication.Assets, release.Download{Name: entry.Name()})
		}
	}
	asset, err := publication.Select(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return req, fmt.Errorf("select bundled %s: %w", id, err)
	}
	req.Archive, err = filepath.Abs(filepath.Join(directory, asset.Name))
	return req, err
}

// bundledUpdate 只选择仍绑定官方来源的更高版本，和正式更新共用完整预检。
func (s *Service) bundledUpdate(ctx context.Context, id string) (req InstallRequest, version string, err error) {
	if s.opt.BundledPackages == "" {
		return req, "", nil
	}
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err != nil {
		return req, "", err
	}
	if in.Uninstalled || !trust.MatchesSource(id, in.Source) {
		return req, "", nil
	}
	if _, err := os.Stat(filepath.Join(s.opt.BundledPackages, id)); errors.Is(err, os.ErrNotExist) {
		return req, "", nil
	} else if err != nil {
		return req, "", err
	}
	req, err = BundledPackage(s.opt.BundledPackages, id)
	if err != nil {
		return req, "", err
	}
	candidate, err := s.prepare(ctx, req)
	if err != nil {
		return req, "", err
	}
	defer func() { err = errors.Join(err, candidate.Discard()) }()
	m := candidate.Manifest
	if m.ID != id || m.Role != string(in.Role) || m.AdapterID != in.AdapterID || m.Source != in.Source {
		return req, "", denied("bundled update identity differs from installed binding")
	}
	if err := checkWebUIVersion(m); err != nil {
		return req, "", err
	}
	if compareBundledVersions(m.Version, in.Version) <= 0 {
		return req, "", nil
	}
	return req, m.Version, nil
}

// compareBundledVersions 的输入都已通过 SemVer 校验；数字先比位数以免溢出。
func compareBundledVersions(a, b string) int {
	a, _, _ = strings.Cut(a, "+")
	b, _, _ = strings.Cut(b, "+")
	left, leftPre, _ := strings.Cut(a, "-")
	right, rightPre, _ := strings.Cut(b, "-")
	leftParts, rightParts := strings.Split(left, "."), strings.Split(right, ".")
	for i := range leftParts {
		if order := compareBundledNumber(leftParts[i], rightParts[i]); order != 0 {
			return order
		}
	}
	if leftPre == rightPre {
		return 0
	}
	if leftPre == "" {
		return 1
	}
	if rightPre == "" {
		return -1
	}
	leftParts, rightParts = strings.Split(leftPre, "."), strings.Split(rightPre, ".")
	for i := range min(len(leftParts), len(rightParts)) {
		l, r := leftParts[i], rightParts[i]
		leftNumeric, rightNumeric := strings.Trim(l, "0123456789") == "", strings.Trim(r, "0123456789") == ""
		order := strings.Compare(l, r)
		switch {
		case leftNumeric && rightNumeric:
			order = compareBundledNumber(l, r)
		case leftNumeric:
			order = -1
		case rightNumeric:
			order = 1
		}
		if order != 0 {
			return order
		}
	}
	return cmp.Compare(len(leftParts), len(rightParts))
}

func compareBundledNumber(a, b string) int {
	if order := cmp.Compare(len(a), len(b)); order != 0 {
		return order
	}
	return strings.Compare(a, b)
}
