// Package packages 在安装前检查组件安装包：受控解压归档，
// 核对包内 manifest，并把通过检查的候选移入不可变的包目录。
//
// 包不下载归档、不登记安装、不持久化来源信任。
//
// 包的组成：
//   - [Preflight] 核对归档与 manifest，返回暂存目录中的 [Candidate]。
//   - [Extract] 按 [Limits] 受控解压 tar.gz，只接受普通文件和目录。
//   - [Candidate.MoveTo] 与 [Candidate.Discard] 结束暂存目录的生命周期。
package packages

import (
	"archive/tar"
	"cmp"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
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

// NewLimits 从生效调优值构造安装包的资源上限。
func NewLimits(v tune.Values) Limits {
	return Limits{MaxArchiveBytes: v.InstallMaxArchive, MaxExpandedBytes: v.InstallMaxExpanded, MaxEntries: v.InstallMaxEntries}
}

// Input 是 [Preflight] 的输入。
type Input struct {
	Source      string // 在线安装时绑定的仓库来源；本地安装取包内声明。
	ID, Version string // 已知的目标身份与 Release 版本，空串表示由包内声明确定。
	OS, Arch    string
	ArchivePath string
}

// Candidate 是通过预检、仍在暂存目录中的安装包。
type Candidate struct {
	Dir      string // 暂存目录；MoveTo 成功或 Discard 后为空
	Manifest *manifest.Manifest
	Digest   []byte   // 不可变包目录和安装记录使用的归档摘要。
	Warnings []string // manifest.Parse 返回的全部警告，保持原顺序。
}

// Preflight 校验组件归档，并解压到 stagingRoot 下新建的唯一目录。
//
// 任何一步失败都会删除已建的暂存目录。校验失败返回带协议错误码的 *validate.Error：
// 来源不符为 PERMISSION_DENIED，平台不符为 UNSUPPORTED，超出上限为
// RESOURCE_EXHAUSTED，归档或 manifest 无效为 INVALID_ARGUMENT。
func Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error) {
	// 1. 检查归档大小并计算现有不可变包标识。
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
	hash := sha256.New()
	if _, err := io.Copy(hash, archive); err != nil {
		return nil, fmt.Errorf("hash archive: %w", err)
	}
	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind archive: %w", err)
	}

	// 2. 解压到新的唯一暂存目录；之后的失败都删除整个目录。
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

	// 3. 核对期望身份、本机平台以及声明的入口与图标。
	data, err := os.ReadFile(filepath.Join(dir, manifestFileName))
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive has no readable "+manifestFileName)
	}
	parsed, err := manifest.Parse(data)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	m := parsed.Manifest
	if _, err := os.Lstat(filepath.Join(dir, "config_schema.json")); err == nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("component %s: rename config_schema.json to plugin_config_schema.json and rebuild", m.ID))
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	switch {
	case in.ID != "" && m.ID != in.ID:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest plugin ID does not match the requested component")
	case in.Version != "" && m.Version != in.Version:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest version does not match the release tag")
	case in.Source != "" && m.Source != in.Source:
		return nil, failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "manifest source does not match the bound source")
	case m.OS != in.OS || m.Arch != in.Arch:
		if m.Role != "webui" || m.Entrypoint != "" || m.OS != "any" || m.Arch != "any" {
			return nil, failure(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "manifest platform does not match this host")
		}
	}
	// 解压已拒绝所有链接；这里核对声明的文件存在且类型正确。
	files := []string{m.Entrypoint}
	if m.Role == "webui" {
		files = []string{m.Assets + "/index.html"}
		if m.Entrypoint != "" {
			files = append(files, m.Entrypoint)
		}
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
	if err := checkChangelog(dir, m.Version); err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	// 4. 示例图片来自已检查的声明；解压已拒绝链接，仍需核对普通文件与图片签名。
	if err := checkExampleImages(dir); err != nil {
		return nil, err
	}
	// 5. 进程包的外置构建信息须来自同一入口，并与声明的平台一致。
	if m.Role != "webui" || m.Entrypoint != "" {
		if err := checkBuildInfo(dir, m); err != nil {
			return nil, err
		}
	}
	keep = true
	return &Candidate{Dir: dir, Manifest: m, Digest: hash.Sum(nil), Warnings: parsed.Warnings}, nil
}

func checkExampleImages(dir string) error {
	for _, schemaName := range []string{"plugin_config_schema.json", "adapter_config_schema.json"} {
		raw, err := os.ReadFile(filepath.Join(dir, schemaName))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", schemaName, err)
		}
		schema, problems := configschema.Parse(raw)
		if len(problems) != 0 {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("%s: %v", schemaName, problems))
		}
		for _, name := range schema.ExampleImages() {
			location := schemaName + ":" + exampleImageLocation(raw, name)
			info, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(name)))
			if err != nil || !info.Mode().IsRegular() {
				return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("%s: image %q must be a regular file: %v", location, name, err))
			}
			content, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
			if err != nil {
				return fmt.Errorf("%s: %w", location, err)
			}
			if http.DetectContentType(content) != exampleContentType(name) {
				return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("%s: image %q content does not match its extension", location, name))
			}
		}
	}
	return nil
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
			// tar 结束标记早于 gzip 校验尾，继续读取以拒绝截断或损坏的归档。
			n, err := io.Copy(io.Discard, io.LimitReader(gz, lim.MaxExpandedBytes-expanded+1))
			if err != nil {
				return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("archive gzip stream is invalid: %v", err))
			}
			if n > lim.MaxExpandedBytes-expanded {
				return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive expanded size exceeds the limit")
			}
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
	final := filepath.Join(parent, c.Manifest.Version+"-"+hex.EncodeToString(c.Digest)[:digestPrefixLength])
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

const changelogFile = "CHANGELOG.md"
const changelogDirectory = "changelog"
const changelogSource = "en-US.md"

func checkChangelog(dir, version string) error {
	single, err := os.Lstat(filepath.Join(dir, changelogFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	hasSingle := err == nil
	directory, err := os.Lstat(filepath.Join(dir, changelogDirectory))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	hasDirectory := err == nil
	if hasSingle && hasDirectory {
		return fmt.Errorf("%s and %s must not both exist", changelogFile, changelogDirectory)
	}
	read := func(name string) error {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if !utf8.Valid(raw) {
			return fmt.Errorf("%s: Markdown must contain valid UTF-8", name)
		}
		return nil
	}
	if hasSingle {
		if !single.Mode().IsRegular() {
			return fmt.Errorf("%s: expected a regular file", changelogFile)
		}
		return read(changelogFile)
	}
	if !hasDirectory {
		return nil
	}
	if !directory.IsDir() {
		return fmt.Errorf("%s: expected a directory", changelogDirectory)
	}
	versions, err := os.ReadDir(filepath.Join(dir, changelogDirectory))
	if err != nil {
		return err
	}
	for _, entry := range versions {
		name := filepath.Join(changelogDirectory, entry.Name())
		if !entry.IsDir() || !manifest.ValidVersion(entry.Name()) {
			return fmt.Errorf("%s: expected a SemVer version directory", name)
		}
		if compareChangelogVersions(entry.Name(), version) > 0 {
			return fmt.Errorf("%s: version exceeds manifest version %s", name, version)
		}
		files, err := os.ReadDir(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		source := false
		for _, file := range files {
			path := filepath.Join(name, file.Name())
			locale := strings.TrimSuffix(file.Name(), ".md")
			canonical, err := canonicalChangelogLocale(locale)
			if file.Type() != 0 || !strings.HasSuffix(file.Name(), ".md") || err != nil || canonical != locale {
				return fmt.Errorf("%s: expected a canonical BCP 47 locale Markdown file", path)
			}
			if err := read(path); err != nil {
				return err
			}
			source = source || file.Name() == changelogSource
		}
		if !source {
			return fmt.Errorf("%s: source changelog is missing", filepath.Join(name, changelogSource))
		}
	}
	return nil
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

// RFC 5646 §2.1 的 language-tag 语法；重复 variant 与 extension singleton 另行检查。
const changelogLanguageTag = `^(?:` +
	`(?:[A-Za-z]{2,3}(?:-[A-Za-z]{3}){0,3}|[A-Za-z]{4}|[A-Za-z]{5,8})` +
	`(?:-[A-Za-z]{4})?(?:-(?:[A-Za-z]{2}|[0-9]{3}))?` +
	`(?:-(?:[A-Za-z0-9]{5,8}|[0-9][A-Za-z0-9]{3}))*` +
	`(?:-[0-9A-WY-Za-wy-z](?:-[A-Za-z0-9]{2,8})+)*` +
	`(?:-[xX](?:-[A-Za-z0-9]{1,8})+)?|[xX](?:-[A-Za-z0-9]{1,8})+)$`

var changelogTagSyntax = regexp.MustCompile(changelogLanguageTag)

// RFC 5646 §2.1 固定的 grandfathered 标签保留其登记大小写。
var changelogGrandfathered = map[string]string{
	"en-gb-oed": "en-GB-oed", "i-ami": "i-ami", "i-bnn": "i-bnn",
	"i-default": "i-default", "i-enochian": "i-enochian", "i-hak": "i-hak",
	"i-klingon": "i-klingon", "i-lux": "i-lux", "i-mingo": "i-mingo",
	"i-navajo": "i-navajo", "i-pwn": "i-pwn", "i-tao": "i-tao",
	"i-tay": "i-tay", "i-tsu": "i-tsu", "sgn-be-fr": "sgn-BE-FR",
	"sgn-be-nl": "sgn-BE-NL", "sgn-ch-de": "sgn-CH-DE", "art-lojban": "art-lojban",
	"cel-gaulish": "cel-gaulish", "no-bok": "no-bok", "no-nyn": "no-nyn",
	"zh-guoyu": "zh-guoyu", "zh-hakka": "zh-hakka", "zh-min": "zh-min",
	"zh-min-nan": "zh-min-nan", "zh-xiang": "zh-xiang",
}

// canonicalChangelogLocale 检查 BCP 47 语法并按 language/script/region 规范大小写；
// extension 与 private-use 保持小写，规则与 Actor.locale 的大小写约定一致。
func canonicalChangelogLocale(locale string) (string, error) {
	lower := strings.ToLower(locale)
	if tag, exists := changelogGrandfathered[lower]; exists {
		return tag, nil
	}
	if !changelogTagSyntax.MatchString(locale) {
		return "", fmt.Errorf("invalid BCP 47 tag %q", locale)
	}
	parts := strings.Split(lower, "-")
	if parts[0] == "x" {
		return lower, nil
	}
	variants := make(map[string]bool)
	singletons := make(map[string]bool)
	extension := false
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if len(part) == 1 {
			if part == "x" {
				break // private-use 后的单字符属于私有内容。
			}
			if singletons[part] {
				return "", fmt.Errorf("duplicate extension %q in BCP 47 tag %q", part, locale)
			}
			singletons[part] = true
			extension = true
		}
		if extension {
			continue
		}
		if len(part) >= 5 || len(part) == 4 && part[0] >= '0' && part[0] <= '9' {
			if variants[part] {
				return "", fmt.Errorf("duplicate variant %q in BCP 47 tag %q", part, locale)
			}
			variants[part] = true
		} else if len(part) == 4 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		} else if len(part) == 2 {
			parts[i] = strings.ToUpper(part)
		}
	}
	return strings.Join(parts, "-"), nil
}

func exampleContentType(name string) string {
	switch path.Ext(name) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return ""
}

// exampleImageLocation 仅处理已通过 schema 检查的声明，把文件错误定位回示例片段。
func exampleImageLocation(raw []byte, image string) string {
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	var find func(map[string]json.RawMessage, string) string
	find = func(fields map[string]json.RawMessage, parent string) string {
		for _, key := range slices.Sorted(maps.Keys(fields)) {
			var field struct {
				Example          json.RawMessage
				Items, Templates map[string]json.RawMessage
			}
			_ = json.Unmarshal(fields[key], &field)
			location := parent + key
			if len(field.Example) > 0 {
				examples := []json.RawMessage{field.Example}
				array := field.Example[0] == '['
				if array {
					_ = json.Unmarshal(field.Example, &examples)
				}
				for i, raw := range examples {
					var example struct{ Reply json.RawMessage }
					_ = json.Unmarshal(raw, &example)
					if example.Reply[0] != '[' {
						continue
					}
					var segments []struct{ Image string }
					_ = json.Unmarshal(example.Reply, &segments)
					for j, segment := range segments {
						if segment.Image == image {
							suffix := ""
							if array {
								suffix = fmt.Sprintf("[%d]", i)
							}
							return fmt.Sprintf("%s.example%s.reply[%d].image", location, suffix, j)
						}
					}
				}
			}
			if result := find(field.Items, location+".items."); result != "" {
				return result
			}
			if result := find(field.Templates, location+".templates."); result != "" {
				return result
			}
		}
		return ""
	}
	return find(fields, "")
}
