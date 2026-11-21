// Package release 校验发布在归档之外的 release.json 及其 Ed25519 签名。
//
// 验签针对 release.json 文件的原始字节，先验签后解析，不会重新序列化 JSON 再验签。
// 签名只证明发布来源，不证明插件行为安全。
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

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Release 对应 release.json。
type Release struct {
	PublisherKeyID string  `json:"publisher_key_id"` // 发布者公钥的 KeyID，必须与验签公钥一致
	Source         string  `json:"source"`           // 绑定的仓库来源，例如 github.com/owner/repo
	PluginID       string  `json:"plugin_id"`        // 插件 ID，规则同 manifest
	Version        string  `json:"version"`          // 发布版本，SemVer 2.0
	Assets         []Asset `json:"assets"`           // 各平台的归档，每个平台最多一个
}

// Asset 描述某个 OS/arch 的归档；SHA256 为整个 tar.gz 的摘要（小写十六进制）。
type Asset struct {
	OS     string `json:"os"`     // 目标 GOOS
	Arch   string `json:"arch"`   // 目标 GOARCH
	URL    string `json:"url"`    // 归档的 https 绝对地址
	Size   int64  `json:"size"`   // 归档字节数，必须为正
	SHA256 string `json:"sha256"` // 整个归档的摘要，64 位小写十六进制
}

// sha256Hex 匹配 64 位小写十六进制的 SHA256。
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
		return nil, fmt.Errorf("signature file is not valid base64: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return nil, fmt.Errorf("signature length %d, want %d", len(sig), ed25519.SignatureSize)
	}
	return sig, nil
}

// DecodePublicKey 解析 release.pub：标准 base64 的 32 字节 Ed25519 公钥，允许末尾空白。
func DecodePublicKey(raw []byte) (ed25519.PublicKey, error) {
	key, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(raw)))
	if err != nil {
		return nil, fmt.Errorf("public key file is not valid base64: %w", err)
	}
	if len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key length %d, want %d", len(key), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(key), nil
}

// Verify 校验 release.json；有签名时先验证原始字节及发布者指纹，无签名时只检查清单。
func Verify(raw, sigFile []byte, pub ed25519.PublicKey) (*Release, error) {
	// 1. 对原始字节验签，签名不对时不解析内容。
	if len(sigFile) != 0 {
		if len(pub) != ed25519.PublicKeySize {
			return nil, errors.New("invalid public key length")
		}
		sig, err := DecodeSignature(sigFile)
		if err != nil {
			return nil, err
		}
		if !ed25519.Verify(pub, raw, sig) {
			return nil, errors.New("release.json signature verification failed")
		}
	}

	// 2. 严格解析：未知字段或对象之后的多余内容都拒绝。
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var r Release
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("parse release.json: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("release.json has trailing content after the JSON object")
	}
	// 3. 声明的发布者必须就是验签公钥，再校验各字段。
	if len(sigFile) != 0 && r.PublisherKeyID != KeyID(pub) {
		return nil, fmt.Errorf("publisher_key_id %q does not match verification key %s", r.PublisherKeyID, KeyID(pub))
	}
	if err := r.validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

// validate 校验插件 ID、来源、版本，以及每个资产的平台、大小、摘要和 URL。
func (r *Release) validate() error {
	if _, err := validate.PluginID(r.PluginID); err != nil {
		return err
	}
	if r.Source == "" {
		return errors.New("missing source")
	}
	if !manifest.ValidVersion(r.Version) {
		return fmt.Errorf("version %q is not valid SemVer", r.Version)
	}
	if len(r.Assets) == 0 {
		return errors.New("assets is empty")
	}
	seen := map[manifest.Platform]bool{}
	for _, a := range r.Assets {
		p := manifest.Platform{OS: a.OS, Arch: a.Arch}
		switch {
		case !manifest.SupportedPlatform(a.OS, a.Arch) && !(a.OS == "any" && a.Arch == "any"):
			return fmt.Errorf("unsupported platform %s/%s", a.OS, a.Arch)
		case seen[p]:
			return fmt.Errorf("duplicate platform %s/%s", a.OS, a.Arch)
		case a.Size <= 0:
			return fmt.Errorf("size of %s/%s must be positive", a.OS, a.Arch)
		case !sha256Hex.MatchString(a.SHA256):
			return fmt.Errorf("sha256 of %s/%s is not 64 lowercase hex characters", a.OS, a.Arch)
		}
		u, err := url.Parse(a.URL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("url of %s/%s must be an absolute https URL", a.OS, a.Arch)
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
