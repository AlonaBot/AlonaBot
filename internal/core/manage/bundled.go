package manage

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

const officialRepository = "github.com/AlonaBot/AlonaBot"

// BundledPackage 读取发行物中指定组件的本机安装输入，保留清单、签名和公钥的原始字节。
// 验签、摘要与包内身份校验由 Install 或 Update 完成；缺失文件保留 os.ErrNotExist。
func BundledPackage(dir, id string) (InstallRequest, error) {
	var req InstallRequest
	if _, err := validate.PluginID(id); err != nil {
		return req, err
	}
	directory := filepath.Join(dir, id)
	for _, file := range []struct {
		name   string
		target *string
	}{
		{"release.json", &req.ReleaseJSON},
		{"release.json.sig", &req.Signature},
		{"release.pub", &req.PublicKey},
	} {
		raw, err := os.ReadFile(filepath.Join(directory, file.name))
		if err != nil {
			return req, fmt.Errorf("read bundled %s: %w", id, err)
		}
		*file.target = string(raw)
	}
	var publication release.Release
	if err := json.Unmarshal([]byte(req.ReleaseJSON), &publication); err != nil {
		return req, fmt.Errorf("parse bundled %s release: %w", id, err)
	}
	if publication.PluginID != id {
		return req, fmt.Errorf("bundled component %s has release identity %q", id, publication.PluginID)
	}
	asset, ok := publication.Asset("any", "any")
	if !ok {
		asset, ok = publication.Asset(runtime.GOOS, runtime.GOARCH)
	}
	if !ok {
		return req, fmt.Errorf("bundled component %s has no asset for %s/%s", id, runtime.GOOS, runtime.GOARCH)
	}
	address, err := url.Parse(asset.URL)
	if err != nil {
		return req, fmt.Errorf("parse bundled %s asset URL: %w", id, err)
	}
	name := path.Base(address.Path)
	if !filepath.IsLocal(name) || name == "." {
		return req, fmt.Errorf("bundled component %s has an invalid archive name", id)
	}
	req.Archive, err = filepath.Abs(filepath.Join(directory, name))
	if err != nil {
		return req, err
	}
	if _, err := os.Stat(req.Archive); err != nil {
		return req, fmt.Errorf("stat bundled %s archive: %w", id, err)
	}
	return req, nil
}

// bundledUpdate 只选择仍绑定官方来源与发布者的更高版本，和正式更新共用完整预检。
func (s *Service) bundledUpdate(ctx context.Context, id string) (req InstallRequest, version string, err error) {
	if s.opt.BundledPackages == "" {
		return req, "", nil
	}
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err != nil {
		return req, "", err
	}
	key, official := trust.Official()
	if !official || in.Uninstalled || strings.TrimPrefix(in.Source, "https://") != officialRepository || in.PublisherKeyID != release.KeyID(key) {
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
	if req.Signature == "" {
		return req, "", denied("bundled update requires an official signature")
	}
	publication, err := release.Verify([]byte(req.ReleaseJSON), []byte(req.Signature), key)
	if err != nil {
		return req, "", err
	}
	if publication.Source != in.Source {
		return req, "", denied("bundled update source differs from installed binding")
	}
	if compareBundledVersions(publication.Version, in.Version) <= 0 {
		return req, "", nil
	}
	candidate, err := s.prepare(ctx, req)
	if err != nil {
		return req, "", err
	}
	defer func() { err = errors.Join(err, candidate.Discard()) }()
	m := candidate.Manifest
	if m.ID != id || m.Role != string(in.Role) || m.AdapterID != in.AdapterID || candidate.Release.PublisherKeyID != in.PublisherKeyID {
		return req, "", denied("bundled update identity differs from installed binding")
	}
	if err := checkWebUIVersion(m); err != nil {
		return req, "", err
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
