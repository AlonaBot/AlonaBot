// Package release 校验归档外的 release.json 及其 Ed25519 签名（CONTRACTS §9.2）。
// 先对文件原始字节验签，再解析；不重新序列化 JSON 后验签。
package release

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"

	"github.com/AlonaBot/alonabot/contracts/manifest"
)

// Release 对应 release.json。
type Release struct {
	// KeyID(发布者公钥)。
	PublisherKeyID string `json:"publisher_key_id"`
	// 绑定的仓库来源，例如 github.com/owner/repo。
	Source   string  `json:"source"`
	PluginID string  `json:"plugin_id"`
	Version  string  `json:"version"`
	Assets   []Asset `json:"assets"`
}

// Asset 是某个 OS/arch 的归档；SHA256 是整个 tar.gz 的摘要（小写十六进制）。
type Asset struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	URL    string `json:"url"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// KeyID 返回公钥标识：ed25519: 加公钥 SHA256 的小写十六进制。
func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return "ed25519:" + hex.EncodeToString(sum[:])
}

// DecodeSignature 解析 release.json.sig：标准 base64 编码的 64 字节签名，允许末尾空白。
func DecodeSignature(sigFile []byte) ([]byte, error) {
	sig, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(sigFile)))
	if err != nil {
		return nil, fmt.Errorf("签名文件不是合法的 base64: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return nil, fmt.Errorf("签名长度 %d，应为 %d", len(sig), ed25519.SignatureSize)
	}
	return sig, nil
}

// Verify 用已信任的公钥对 release.json 的原始字节验签，通过后才解析并校验内容。
func Verify(raw, sigFile []byte, pub ed25519.PublicKey) (*Release, error) {
	if len(pub) != ed25519.PublicKeySize {
		return nil, errors.New("公钥长度错误")
	}
	sig, err := DecodeSignature(sigFile)
	if err != nil {
		return nil, err
	}
	if !ed25519.Verify(pub, raw, sig) {
		return nil, errors.New("release.json 签名校验失败")
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var r Release
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("release.json 解析失败: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("release.json 在 JSON 对象之后还有多余内容")
	}
	if r.PublisherKeyID != KeyID(pub) {
		return nil, fmt.Errorf("publisher_key_id %q 与验签公钥 %s 不符", r.PublisherKeyID, KeyID(pub))
	}
	if err := r.validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *Release) validate() error {
	if _, err := manifest.ValidateID(r.PluginID); err != nil {
		return err
	}
	if r.Source == "" {
		return errors.New("缺少 source")
	}
	if !manifest.ValidVersion(r.Version) {
		return fmt.Errorf("version %q 不是合法的 SemVer", r.Version)
	}
	if len(r.Assets) == 0 {
		return errors.New("assets 为空")
	}
	seen := map[manifest.Platform]bool{}
	for _, a := range r.Assets {
		p := manifest.Platform{OS: a.OS, Arch: a.Arch}
		switch {
		case !manifest.SupportedPlatform(a.OS, a.Arch):
			return fmt.Errorf("不支持的平台 %s/%s", a.OS, a.Arch)
		case seen[p]:
			return fmt.Errorf("平台 %s/%s 重复", a.OS, a.Arch)
		case a.Size <= 0:
			return fmt.Errorf("%s/%s 的 size 必须为正", a.OS, a.Arch)
		case !sha256Hex.MatchString(a.SHA256):
			return fmt.Errorf("%s/%s 的 sha256 不是 64 位小写十六进制", a.OS, a.Arch)
		}
		u, err := url.Parse(a.URL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("%s/%s 的 url 必须是 https 绝对地址", a.OS, a.Arch)
		}
		seen[p] = true
	}
	return nil
}

// Asset 返回指定平台的资产。
func (r *Release) Asset(os, arch string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.OS == os && a.Arch == arch {
			return a, true
		}
	}
	return Asset{}, false
}
