package pluginkit

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/sdk/generate"
)

// Pack 构建组件并发布可重复归档、release.json 和可选的 Ed25519 签名。
func Pack(ctx context.Context, options generate.Options, assetURL, keyFile, outDir string) (archive string, err error) {
	options.CGOEnabled = "0"
	options.RequireAdapterConstructor = true
	// 1. 发布地址、密钥与静态资源问题先于构建和产物发布拒绝。
	u, err := url.Parse(assetURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", errors.New("--asset-url must be an absolute https URL")
	}
	var key ed25519.PrivateKey
	var keyInfo os.FileInfo
	if keyFile != "" {
		raw, err := os.ReadFile(keyFile)
		if err != nil {
			return "", fmt.Errorf("read signing key: %w", err)
		}
		decoded, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(raw)))
		if err != nil {
			return "", fmt.Errorf("decode signing key: %w", err)
		}
		if len(decoded) != ed25519.PrivateKeySize {
			return "", fmt.Errorf("private key length %d, want %d", len(decoded), ed25519.PrivateKeySize)
		}
		key = ed25519.PrivateKey(decoded)
		keyInfo, err = os.Stat(keyFile)
		if err != nil {
			return "", err
		}
	}
	snapshot, err := scanProject(ctx, options)
	if err != nil {
		return "", err
	}
	files := snapshot.files
	if keyInfo != nil {
		for name := range files {
			info, err := os.Stat(filepath.Join(options.Dir, filepath.FromSlash(name)))
			if err != nil {
				return "", err
			}
			if os.SameFile(keyInfo, info) {
				return "", fmt.Errorf("signing key must not be included as resource %q", name)
			}
		}
	}
	if outDir == "" {
		outDir = filepath.Join(options.Dir, "bin")
	}
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return "", err
	}
	staging, err := os.MkdirTemp("", "alonabot-pack-*")
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()
	// 2. 归档资源取同一次构建的规范副本，薄入口与临时 module 信息不进入包。
	build, err := buildComponent(ctx, options, filepath.Join(staging, "build"), false, snapshot)
	if err != nil {
		return "", err
	}
	parsed, err := manifest.Parse(build.PackageManifest)
	if err != nil {
		return "", err
	}
	m := parsed.Manifest
	for _, reserved := range []string{projectFile, "build-info", m.Entrypoint} {
		if _, exists := files[reserved]; exists {
			return "", fmt.Errorf("resource %q conflicts with a generated package file", reserved)
		}
	}
	name := fmt.Sprintf("%s-%s-%s-%s.tar.gz", m.ID, m.Version, m.OS, m.Arch)
	if m.Role == "webui" {
		name = fmt.Sprintf("%s-%s-any.tar.gz", m.ID, m.Version)
	} else {
		files[m.Entrypoint], err = os.ReadFile(build.Executable)
		if err != nil {
			return "", err
		}
		info, err := buildinfo.ReadFile(build.Executable)
		if err != nil {
			return "", fmt.Errorf("read executable build-info: %w", err)
		}
		files["build-info"] = []byte(info.String())
	}
	files[projectFile] = build.PackageManifest
	compressed, err := archiveFiles(files, m.Entrypoint)
	if err != nil {
		return "", err
	}
	// SHA256 只绑定完整归档，满足 release 资产完整性契约。
	digest := sha256.Sum256(compressed)
	publication := release.Release{Source: m.Source, PluginID: m.ID, Version: m.Version, Assets: []release.Asset{{OS: m.OS, Arch: m.Arch, URL: assetURL, Size: int64(len(compressed)), SHA256: hex.EncodeToString(digest[:])}}}
	outputs := map[string][]byte{name: compressed}
	if key != nil {
		publication.PublisherKeyID = release.KeyID(key.Public().(ed25519.PublicKey))
	}
	raw, err := json.MarshalIndent(publication, "", "  ")
	if err != nil {
		return "", err
	}
	raw = append(raw, '\n')
	outputs["release.json"] = raw
	if key != nil {
		outputs["release.json.sig"] = []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(key, raw)) + "\n")
		outputs["release.pub"] = []byte(base64.StdEncoding.EncodeToString(key.Public().(ed25519.PublicKey)) + "\n")
	}
	if _, err := release.Verify(raw, outputs["release.json.sig"], keyPublic(key)); err != nil {
		return "", err
	}
	// 3. 全部内容准备成功后发布；unsigned 重用输出目录时移除旧签名身份。
	if err := os.MkdirAll(outDir, privateMode); err != nil {
		return "", err
	}
	for name, raw := range outputs {
		if err := os.WriteFile(filepath.Join(staging, name), raw, fileMode); err != nil {
			return "", err
		}
	}
	if key == nil {
		for _, name := range []string{"release.json.sig", "release.pub"} {
			if err := os.Remove(filepath.Join(outDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return "", err
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(outputs)) {
		if err := os.Rename(filepath.Join(staging, name), filepath.Join(outDir, name)); err != nil {
			return "", err
		}
	}
	return filepath.Join(outDir, name), nil
}

func keyPublic(key ed25519.PrivateKey) ed25519.PublicKey {
	if key == nil {
		return nil
	}
	return key.Public().(ed25519.PublicKey)
}

func archiveFiles(files map[string][]byte, executable string) ([]byte, error) {
	var output bytes.Buffer
	compressed := gzip.NewWriter(&output)
	archive := tar.NewWriter(compressed)
	seen := map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		if err := manifest.ValidateRelPath(name); err != nil {
			return nil, err
		}
		for parent := name; parent != "."; parent = filepath.ToSlash(filepath.Dir(parent)) {
			if _, exists := files[parent]; exists && parent != name {
				return nil, fmt.Errorf("archive file %q conflicts with parent directory of %q", parent, name)
			}
			folded := strings.ToLower(parent)
			if previous, ok := seen[folded]; ok && previous != parent {
				return nil, fmt.Errorf("case-conflicting archive paths %q and %q", previous, parent)
			}
			seen[folded] = parent
		}
		mode := int64(0o644)
		if name == executable {
			mode = 0o755
		}
		if err := archive.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(files[name])), Typeflag: tar.TypeReg, Format: tar.FormatPAX}); err != nil {
			return nil, err
		}
		if _, err := archive.Write(files[name]); err != nil {
			return nil, err
		}
	}
	if err := errors.Join(archive.Close(), compressed.Close()); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
