// Package packages 在安装前检查组件安装包：校验发布清单及可选签名，受控解压归档，
// 核对包内 manifest，并把通过检查的候选移入不可变的包目录。
//
// 包不下载归档、不登记安装、不持久化来源信任。检查顺序与上限见
// spec/CONTRACTS-v0.1.md §9.1、§9.2 与 §9.3 末段。
//
// 包的组成：
//   - [Preflight] 依次验签、核对归档与 manifest，返回暂存目录中的 [Candidate]。
//   - [Extract] 按 [Limits] 受控解压 tar.gz，只接受普通文件和目录。
//   - [Candidate.MoveTo] 与 [Candidate.Discard] 结束暂存目录的生命周期。
package packages

import (
	"archive/tar"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

const (
	manifestFileName    = "plugin.yaml"
	stagingPattern      = "alonabot-candidate-"
	buildInfoFileName   = "build-info"
	missingBuildInfo    = "archive has no build-info"
	invalidBuildInfo    = "parse build-info: %v"
	mismatchedBuildInfo = "build-info does not match the entrypoint"
	mismatchedBuildOS   = "build-info GOOS does not match the manifest os"
	mismatchedBuildArch = "build-info GOARCH does not match the manifest arch"
	// digestPrefixLength 是包目录名中归档 SHA256 的十六进制位数。
	digestPrefixLength = 12

	dirMode        os.FileMode = 0o700
	fileMode       os.FileMode = 0o600
	executableMode os.FileMode = 0o700 // 归档中带任一执行位的文件
)

// Limits 是一次安装的资源上限，超出时返回 RESOURCE_EXHAUSTED。
type Limits struct {
	MaxArchiveBytes  int64 // 压缩归档的字节数
	MaxExpandedBytes int64 // 解压后普通文件的字节总数
	MaxEntries       int   // 归档条目数，含目录
}

// NewLimits 从生效调优值构造安装包的资源上限（CONTRACTS §16）。
func NewLimits(v tune.Values) Limits {
	return Limits{MaxArchiveBytes: v.InstallMaxArchive, MaxExpandedBytes: v.InstallMaxExpanded, MaxEntries: v.InstallMaxEntries}
}

// Input 是 [Preflight] 的输入。
type Input struct {
	ReleaseJSON []byte            // release.json 的原始字节
	Signature   []byte            // release.json.sig 的原始字节
	PublicKey   ed25519.PublicKey // 已信任的发布者公钥
	Source      string            // 已绑定的来源，例如 github.com/owner/repo
	OS, Arch    string            // 本机平台
	ArchivePath string            // 已下载到本地的归档
}

// Candidate 是通过预检、仍在暂存目录中的安装包。
type Candidate struct {
	Dir      string // 暂存目录；MoveTo 成功或 Discard 后为空
	Manifest *manifest.Manifest
	Release  *release.Release
	Asset    release.Asset // 本机平台对应的资产
	Warnings []string      // manifest.Parse 返回的全部警告，保持原顺序。
}

// Preflight 校验发布清单、可选签名与归档，并解压到 stagingRoot 下新建的唯一目录。
//
// 任何一步失败都会删除已建的暂存目录。校验失败返回带协议错误码的 *validate.Error：
// 签名或来源不符为 PERMISSION_DENIED，缺少本平台资产为 UNSUPPORTED，超出上限为
// RESOURCE_EXHAUSTED，归档与清单或 manifest 不一致为 INVALID_ARGUMENT。
func Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error) {
	// 1. 先验签再解析清单，核对来源与本机平台。
	r, err := release.Verify(in.ReleaseJSON, in.Signature, in.PublicKey)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "verify release: "+err.Error())
	}
	if r.Source != in.Source {
		return nil, failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "release source does not match the bound source")
	}
	asset, ok := r.Asset("any", "any")
	if !ok {
		asset, ok = r.Asset(in.OS, in.Arch)
	}
	if !ok {
		return nil, failure(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, fmt.Sprintf("release has no asset for %s/%s", in.OS, in.Arch))
	}

	// 2. 核对归档的大小与整个文件的 SHA256，之后才解压。
	archive, err := os.Open(in.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer archive.Close()
	info, err := archive.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat archive: %w", err)
	}
	if info.Size() > lim.MaxArchiveBytes {
		return nil, failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive exceeds the size limit")
	}
	if info.Size() != asset.Size {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive size does not match the release")
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, archive); err != nil {
		return nil, fmt.Errorf("hash archive: %w", err)
	}
	if hex.EncodeToString(hash.Sum(nil)) != asset.SHA256 {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive SHA256 does not match the release")
	}
	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind archive: %w", err)
	}

	// 3. 解压到新的唯一暂存目录；之后的失败都删除整个目录。
	if err := os.MkdirAll(stagingRoot, dirMode); err != nil {
		return nil, fmt.Errorf("create staging root: %w", err)
	}
	dir, err := os.MkdirTemp(stagingRoot, stagingPattern)
	if err != nil {
		return nil, fmt.Errorf("create staging directory: %w", err)
	}
	keep := false
	defer func() {
		if !keep {
			os.RemoveAll(dir)
		}
	}()
	if err := extract(archive, dir, lim); err != nil {
		return nil, err
	}

	// 4. 包内 manifest 必须与签名清单和本机平台一致，声明的入口与图标必须是普通文件。
	data, err := os.ReadFile(filepath.Join(dir, manifestFileName))
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive has no readable "+manifestFileName)
	}
	parsed, err := manifest.Parse(data)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	m := parsed.Manifest
	switch {
	case m.ID != r.PluginID:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest plugin ID does not match the release")
	case m.Version != r.Version:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest version does not match the release")
	case m.Source != r.Source:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest source does not match the release")
	case m.OS != asset.OS || m.Arch != asset.Arch:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest platform does not match this host")
	case m.Role != "webui" && (asset.OS != in.OS || asset.Arch != in.Arch):
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "only webui can use the any platform")
	}
	// 解压已拒绝所有链接；这里核对声明的文件存在且类型正确。
	files := []string{m.Entrypoint}
	if m.Role == "webui" {
		files = []string{m.Assets + "/index.html"}
	}
	if m.Icon != "" {
		files = append(files, m.Icon)
	}
	for _, name := range files {
		entry, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil || !entry.Mode().IsRegular() {
			return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("manifest file %q is not a regular file in the archive", name))
		}
	}
	// 5. 进程包的外置构建信息须来自同一入口，并与声明的平台一致。
	if m.Role != "webui" {
		if err := checkBuildInfo(dir, m); err != nil {
			return nil, err
		}
	}
	keep = true
	return &Candidate{Dir: dir, Manifest: m, Release: r, Asset: asset, Warnings: parsed.Warnings}, nil
}

func checkBuildInfo(dir string, m *manifest.Manifest) error {
	raw, err := os.ReadFile(filepath.Join(dir, buildInfoFileName))
	if errors.Is(err, os.ErrNotExist) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, missingBuildInfo)
	}
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf(invalidBuildInfo, err))
	}
	info, err := debug.ParseBuildInfo(string(raw))
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf(invalidBuildInfo, err))
	}
	entry, err := buildinfo.ReadFile(filepath.Join(dir, filepath.FromSlash(m.Entrypoint)))
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, mismatchedBuildInfo+": "+err.Error())
	}
	if entry.String() != string(raw) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, mismatchedBuildInfo)
	}
	var goos, goarch string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "GOOS":
			goos = setting.Value
		case "GOARCH":
			goarch = setting.Value
		}
	}
	if goos != m.OS {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, mismatchedBuildOS)
	}
	if goarch != m.Arch {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, mismatchedBuildArch)
	}
	return nil
}

// Extract 把 archivePath 处的 tar.gz 受控解压到已存在的目录 dest。
//
// 只接受普通文件和目录；绝对路径、.. 组件、反斜杠、盘符、空名字、链接、设备、FIFO、
// 重复路径以及只有大小写不同的路径都返回 INVALID_ARGUMENT。条目数或写出的字节总数超过
// lim 时返回 RESOURCE_EXHAUSTED。目录权限为 0700，普通文件为 0600，带执行位的为 0700，
// 不保留 setuid、setgid 与 sticky。失败时 dest 中可能留有部分文件，由调用方删除。
func Extract(archivePath, dest string, lim Limits) error {
	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer archive.Close()
	return extract(archive, dest, lim)
}

func extract(r io.Reader, dest string, lim Limits) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive is not valid gzip data")
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{} // 已出现的路径，按小写比较以拒绝大小写冲突
	var entries int
	var expanded int64
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive tar stream is invalid")
		}
		entries++
		if entries > lim.MaxEntries {
			return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive has too many entries")
		}
		name, err := entryName(h)
		if err != nil {
			return err
		}
		if seen[strings.ToLower(name)] {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive has duplicate or case-colliding paths")
		}
		seen[strings.ToLower(name)] = true

		target := filepath.Join(dest, filepath.FromSlash(name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, dirMode); err != nil {
				return fmt.Errorf("create directory %s: %w", name, err)
			}
			// MkdirAll 不改动已隐式创建的目录，也受 umask 影响，所以显式设置权限。
			if err := os.Chmod(target, dirMode); err != nil {
				return fmt.Errorf("chmod directory %s: %w", name, err)
			}
		case tar.TypeReg:
			if h.Size > lim.MaxExpandedBytes-expanded {
				return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive expanded size exceeds the limit")
			}
			if err := writeFile(target, tr, h); err != nil {
				return err
			}
			expanded += h.Size
		default:
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("archive entry %q has a disallowed type", name))
		}
	}
}

// entryName 返回条目的规范相对路径，不合法时返回 INVALID_ARGUMENT。
func entryName(h *tar.Header) (string, error) {
	name := h.Name
	if h.Typeflag == tar.TypeDir {
		name = strings.TrimSuffix(name, "/")
	}
	// 反斜杠与冒号在 Windows 上是分隔符和盘符，任何平台都拒绝，保证同一归档在五个平台上结果一致。
	if strings.ContainsAny(name, "\\:\x00") || !filepath.IsLocal(filepath.FromSlash(name)) || path.Clean(name) != name {
		return "", failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("archive path %q is not a canonical relative path", h.Name))
	}
	return name, nil
}

// writeFile 新建 target 并写入当前条目的内容；target 已存在时视为重复路径。
func writeFile(target string, content io.Reader, h *tar.Header) error {
	if err := os.MkdirAll(filepath.Dir(target), dirMode); err != nil {
		return fmt.Errorf("create parent of %s: %w", h.Name, err)
	}
	mode := fileMode
	if h.Mode&0o111 != 0 {
		mode = executableMode
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if errors.Is(err, os.ErrExist) {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive has duplicate or case-colliding paths")
	}
	if err != nil {
		return fmt.Errorf("create %s: %w", h.Name, err)
	}
	_, copyErr := io.CopyN(out, content, h.Size)
	if err := errors.Join(copyErr, out.Close()); err != nil {
		return fmt.Errorf("write %s: %w", h.Name, err)
	}
	// OpenFile 的权限受 umask 影响，写完后显式设置。
	if err := os.Chmod(target, mode); err != nil {
		return fmt.Errorf("chmod %s: %w", h.Name, err)
	}
	return nil
}

// MoveTo 把暂存目录移动为 <packagesRoot>/<plugin_id>/<version>-<归档 SHA256 前 12 位>，返回最终目录。
//
// 目标已存在时返回 ALREADY_EXISTS，绝不覆盖。暂存目录与 packagesRoot 须在同一文件系统，
// 由调用方保证；os.Rename 失败时原样返回错误。
func (c *Candidate) MoveTo(packagesRoot string) (string, error) {
	parent := filepath.Join(packagesRoot, c.Manifest.ID)
	if err := os.MkdirAll(parent, dirMode); err != nil {
		return "", fmt.Errorf("create package parent: %w", err)
	}
	final := filepath.Join(parent, c.Manifest.Version+"-"+c.Asset.SHA256[:digestPrefixLength])
	if _, err := os.Lstat(final); err == nil {
		return "", failure(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "package directory already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("check package directory: %w", err)
	}
	if err := os.Rename(c.Dir, final); err != nil {
		return "", err
	}
	c.Dir = ""
	return final, nil
}

// Discard 删除暂存目录；候选已移动或已丢弃时什么也不做。
func (c *Candidate) Discard() error {
	if c.Dir == "" {
		return nil
	}
	if err := os.RemoveAll(c.Dir); err != nil {
		return fmt.Errorf("remove staging directory: %w", err)
	}
	c.Dir = ""
	return nil
}

func failure(code pb.ErrorCode, message string) error {
	return &validate.Error{Code: code, Message: message}
}
