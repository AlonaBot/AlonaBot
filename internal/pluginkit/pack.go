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
	"runtime"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/generate"
)

// Pack 构建单个平台归档；assetURL 为完整资产 URL，保留直接调用的约定。
func Pack(ctx context.Context, options generate.Options, assetURL, keyFile, outDir string) (string, error) {
	archives, err := packTargets(ctx, options, []manifest.Platform{{OS: options.GOOS, Arch: options.GOARCH}}, assetURL, keyFile, outDir, false)
	if err != nil {
		return "", err
	}
	return archives[0], nil
}

// PackTargets 逐目标构建归档，合并到一份发布清单并签名；assetURL 为地址前缀，空值从 GitHub 来源推导。
// targets 为空时使用本机平台；outDir 为空时写入工程 bin/。
func PackTargets(ctx context.Context, options generate.Options, targets []manifest.Platform, assetURL, keyFile, outDir string) (archives []string, err error) {
	if len(targets) == 0 {
		targets = []manifest.Platform{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
	}
	seen := map[manifest.Platform]bool{}
	for _, target := range targets {
		if !manifest.SupportedPlatform(target.OS, target.Arch) {
			return nil, fmt.Errorf("unsupported target %s/%s; supported targets: %v", target.OS, target.Arch, manifest.Platforms)
		}
		if seen[target] {
			return nil, fmt.Errorf("duplicate target %s/%s", target.OS, target.Arch)
		}
		seen[target] = true
	}
	return packTargets(ctx, options, targets, assetURL, keyFile, outDir, true)
}

func packTargets(ctx context.Context, options generate.Options, targets []manifest.Platform, assetURL, keyFile, outDir string, assetBase bool) (archives []string, err error) {
	options.CGOEnabled = "0"
	options.RequireAdapterConstructor = true
	// 1. 发布地址与密钥先于构建检查；旧的单平台入口仍接收完整 URL。
	if assetBase && assetURL == "" {
		raw, err := os.ReadFile(filepath.Join(options.Dir, projectFile))
		if err != nil {
			return nil, err
		}
		project, err := manifest.ParseProject(raw)
		if err != nil {
			return nil, err
		}
		repo, github := strings.CutPrefix(project.Manifest.Source, "github.com/")
		parts := strings.Split(repo, "/")
		if !github || len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(repo, "?#\\ \t\n") {
			return nil, fmt.Errorf("cannot infer asset base from source %q; specify --asset-base", project.Manifest.Source)
		}
		assetURL = "https://github.com/" + repo + "/releases/download/v" + project.Manifest.Version
	}
	u, err := url.Parse(assetURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, errors.New("asset URL must be an absolute https URL")
	}
	if assetBase && (u.RawQuery != "" || u.Fragment != "") {
		return nil, errors.New("--asset-base must not contain a query or fragment")
	}
	var key ed25519.PrivateKey
	var keyInfo os.FileInfo
	if keyFile != "" {
		raw, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("read signing key: %w", err)
		}
		decoded, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(raw)))
		if err != nil {
			return nil, fmt.Errorf("decode signing key: %w", err)
		}
		if len(decoded) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("private key length %d, want %d", len(decoded), ed25519.PrivateKeySize)
		}
		key = ed25519.PrivateKey(decoded)
		keyInfo, err = os.Stat(keyFile)
		if err != nil {
			return nil, err
		}
	}
	if outDir == "" {
		outDir = filepath.Join(options.Dir, "bin")
	}
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return nil, err
	}
	staging, err := os.MkdirTemp("", "alonabot-pack-*")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()

	// 2. 每个平台分别扫描、构建与归档，成功后才发布清单与全部归档。
	var publication release.Release
	outputs := map[string][]byte{}
	for i, target := range targets {
		options.GOOS, options.GOARCH = target.OS, target.Arch
		snapshot, err := scanProject(ctx, options)
		if err != nil {
			return nil, err
		}
		name, compressed, m, err := packArchive(ctx, options, snapshot, filepath.Join(staging, fmt.Sprintf("build-%d", i)), keyInfo)
		if err != nil {
			return nil, err
		}
		if _, exists := outputs[name]; exists {
			continue
		} // 静态 WebUI 的 any 归档只发布一次。
		address := assetURL
		if assetBase {
			address = strings.TrimRight(assetURL, "/") + "/" + name
		}
		// 摘要绑定完整归档，满足 release 资产完整性契约。
		digest := sha256.Sum256(compressed)
		publication.Source, publication.PluginID, publication.Version = m.Source, m.ID, m.Version
		publication.Assets = append(publication.Assets, release.Asset{OS: m.OS, Arch: m.Arch, URL: address, Size: int64(len(compressed)), SHA256: hex.EncodeToString(digest[:])})
		outputs[name] = compressed
		archives = append(archives, filepath.Join(outDir, name))
	}
	if key != nil {
		publication.PublisherKeyID = release.KeyID(keyPublic(key))
	}
	raw, err := json.MarshalIndent(publication, "", "  ")
	if err != nil {
		return nil, err
	}
	raw = append(raw, '\n')
	outputs["release.json"] = raw
	if key != nil {
		outputs["release.json.sig"] = []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(key, raw)) + "\n")
		outputs["release.pub"] = []byte(base64.StdEncoding.EncodeToString(keyPublic(key)) + "\n")
	}
	if _, err := release.Verify(raw, outputs["release.json.sig"], keyPublic(key)); err != nil {
		return nil, err
	}
	// 3. 全部内容准备成功后发布；unsigned 重用输出目录时移除旧签名身份。
	if err := os.MkdirAll(outDir, privateMode); err != nil {
		return nil, err
	}
	for name, raw := range outputs {
		if err := os.WriteFile(filepath.Join(staging, name), raw, fileMode); err != nil {
			return nil, err
		}
	}
	if key == nil {
		for _, name := range []string{"release.json.sig", "release.pub"} {
			if err := os.Remove(filepath.Join(outDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(outputs)) {
		if err := os.Rename(filepath.Join(staging, name), filepath.Join(outDir, name)); err != nil {
			return nil, err
		}
	}
	return archives, nil
}

func packArchive(ctx context.Context, options generate.Options, snapshot *projectScan, buildDir string, keyInfo os.FileInfo) (string, []byte, *manifest.Manifest, error) {
	files := snapshot.files
	if keyInfo != nil {
		for name := range files {
			info, err := os.Stat(filepath.Join(options.Dir, filepath.FromSlash(name)))
			if err != nil {
				return "", nil, nil, err
			}
			if os.SameFile(keyInfo, info) {
				return "", nil, nil, fmt.Errorf("signing key must not be included as resource %q", name)
			}
		}
	}
	// 归档资源取同一次构建的规范副本，薄入口与临时 module 信息不进入包。
	build, err := buildComponent(ctx, options, buildDir, false, snapshot)
	if err != nil {
		return "", nil, nil, err
	}
	parsed, err := manifest.Parse(build.PackageManifest)
	if err != nil {
		return "", nil, nil, err
	}
	m := parsed.Manifest
	for _, reserved := range []string{projectFile, "build-info", m.Entrypoint} {
		if _, exists := files[reserved]; exists {
			return "", nil, nil, fmt.Errorf("resource %q conflicts with a generated package file", reserved)
		}
	}
	name := fmt.Sprintf("%s-%s-%s-%s.tar.gz", m.ID, m.Version, m.OS, m.Arch)
	if m.Role == "webui" && m.Entrypoint == "" {
		name = fmt.Sprintf("%s-%s-any.tar.gz", m.ID, m.Version)
	} else {
		files[m.Entrypoint], err = os.ReadFile(build.Executable)
		if err != nil {
			return "", nil, nil, err
		}
		info, err := buildinfo.ReadFile(build.Executable)
		if err != nil {
			return "", nil, nil, fmt.Errorf("read executable build-info: %w", err)
		}
		files["build-info"] = []byte(info.String())
	}
	files[projectFile] = build.PackageManifest
	compressed, err := archiveFiles(files, m.Entrypoint)
	if err != nil {
		return "", nil, nil, err
	}
	return name, compressed, m, nil
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
