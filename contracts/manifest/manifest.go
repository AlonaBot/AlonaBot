// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：解析并校验 plugin.yaml、插件 ID 与包内相对路径。

// Package manifest 解析并校验插件包内的 plugin.yaml（CONTRACTS §9.1）与插件 ID 规则（§7.1）。
package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Platform 描述一个首发支持的 OS/arch 组合。
type Platform struct {
	OS   string // GOOS 取值
	Arch string // GOARCH 取值
}

// Platforms 列出首发的五个目标组合（CONTRACTS §7.1）。
var Platforms = []Platform{
	{"linux", "amd64"}, {"linux", "arm64"}, {"windows", "amd64"}, {"darwin", "amd64"}, {"darwin", "arm64"},
}

// SupportedPlatform 报告 os/arch 是否属于首发目标。
func SupportedPlatform(os, arch string) bool {
	for _, p := range Platforms {
		if p.OS == os && p.Arch == arch {
			return true
		}
	}
	return false
}

// 插件 ID 规则（CONTRACTS §7.1）。
const (
	MaxIDLen     = 64                 // 插件 ID 的最大字节数
	MarketPrefix = "alonabot_plugin_" // 市场插件 ID 的固定前缀
)

var (
	marketID = regexp.MustCompile(`^alonabot_plugin_[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
	localID  = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	// SemVer 2.0 官方正则。
	semver = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
)

// reservedNames 收录 Windows 保留设备名，它们作为路径组件在 Windows 上不可用。
var reservedNames = func() map[string]bool {
	m := map[string]bool{"con": true, "prn": true, "aux": true, "nul": true}
	for i := 1; i <= 9; i++ {
		m[fmt.Sprintf("com%d", i)] = true
		m[fmt.Sprintf("lpt%d", i)] = true
	}
	return m
}()

// ValidateID 校验插件 ID。市场 ID 直接通过；不带市场前缀的安全本地 ID 通过但返回警告。
// 大写字母被拒绝：ID 以规范小写保存，不能靠大小写区分两个安装。
func ValidateID(id string) (warnings []string, err error) {
	switch {
	case id == "":
		return nil, errors.New("缺少插件 ID")
	case len(id) > MaxIDLen:
		return nil, fmt.Errorf("插件 ID 长度 %d 超过 %d", len(id), MaxIDLen)
	case marketID.MatchString(id):
		return nil, nil
	case strings.HasPrefix(id, MarketPrefix):
		return nil, fmt.Errorf("插件 ID %q 带市场前缀但不符合市场 ID 规则", id)
	case strings.ToLower(id) != id:
		return nil, fmt.Errorf("插件 ID %q 含大写字母，必须使用规范小写", id)
	case !localID.MatchString(id):
		return nil, fmt.Errorf("插件 ID %q 不是安全 slug（只允许小写字母开头，后跟小写字母、数字、_ 或 -）", id)
	case reservedNames[id]:
		return nil, fmt.Errorf("插件 ID %q 是 Windows 保留设备名", id)
	}
	return []string{fmt.Sprintf("插件 ID %q 不带市场前缀 %s，只能作为本地插件安装", id, MarketPrefix)}, nil
}

// ValidVersion 报告 v 是否为合法 SemVer 2.0。
func ValidVersion(v string) bool { return semver.MatchString(v) }

// Capability 描述 manifest 声明的一项协议必需能力。
type Capability struct {
	Name    string `yaml:"name"`    // 能力名，不能为空
	Version uint32 `yaml:"version"` // 能力版本，不能为 0
}

// 迁移模式与准备期副作用声明（§9.1）。空串表示未声明：manifest 仍可解析，但不允许自动升级。
const (
	MigrationNone      = "none"
	MigrationSelf      = "self"
	MigrationVersioned = "versioned"

	EffectsNone      = "none"
	EffectsManagedDB = "managed_db"
	EffectsExternal  = "external"
)

// Manifest 对应 plugin.yaml。必填字段使用指针以区分「缺失」与零值。
type Manifest struct {
	ManifestVersion      *int          `yaml:"manifest_version"`      // manifest 格式版本，目前必须为 1
	ID                   string        `yaml:"id"`                    // 插件 ID，规则见 ValidateID
	Version              string        `yaml:"version"`               // 插件版本，SemVer 2.0
	ProtocolMinMajor     *uint32       `yaml:"protocol_min_major"`    // 支持的最低协议主版本，不能为 0
	ProtocolMaxMajor     *uint32       `yaml:"protocol_max_major"`    // 支持的最高协议主版本，不小于 ProtocolMinMajor
	RequiredCapabilities *[]Capability `yaml:"required_capabilities"` // 必需的协议能力；没有时必须写空列表
	OS                   string        `yaml:"os"`                    // 包的目标 GOOS
	Arch                 string        `yaml:"arch"`                  // 包的目标 GOARCH
	Entrypoint           string        `yaml:"entrypoint"`            // 包内可执行文件的相对路径，规则见 ValidateRelPath
	ConfigVersion        *uint32       `yaml:"config_version"`        // 包内配置 Schema 的版本
	// 包的来源（例如 github.com/owner/repo），不能为空。
	Source         string `yaml:"source"`
	MigrationMode  string `yaml:"migration_mode"`  // 数据库迁移方式（Migration*）；空串表示未声明
	PrepareEffects string `yaml:"prepare_effects"` // 准备期副作用声明（Effects*）；空串表示未声明
}

// Result 保存解析成功的 manifest 及其非致命警告。
type Result struct {
	Manifest *Manifest // 已通过校验的 manifest
	Warnings []string  // 不阻止安装、但需要告知用户的问题
}

// Parse 严格解析 plugin.yaml：未知字段、多个 YAML 文档、缺失必填项或非法值都会失败。
func Parse(data []byte) (*Result, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("plugin.yaml 解析失败: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("plugin.yaml 只能包含一个 YAML 文档")
	}
	warnings, err := m.validate()
	if err != nil {
		return nil, err
	}
	return &Result{Manifest: &m, Warnings: warnings}, nil
}

// validate 逐项校验字段，返回非致命警告；任一项不合法即返回错误。
func (m *Manifest) validate() ([]string, error) {
	if m.ManifestVersion == nil || *m.ManifestVersion != 1 {
		return nil, errors.New("manifest_version 必须为 1")
	}
	warnings, err := ValidateID(m.ID)
	if err != nil {
		return nil, err
	}
	if !ValidVersion(m.Version) {
		return nil, fmt.Errorf("version %q 不是合法的 SemVer", m.Version)
	}
	if m.ProtocolMinMajor == nil || m.ProtocolMaxMajor == nil {
		return nil, errors.New("缺少 protocol_min_major 或 protocol_max_major")
	}
	if *m.ProtocolMinMajor == 0 || *m.ProtocolMinMajor > *m.ProtocolMaxMajor {
		return nil, fmt.Errorf("协议主版本范围 [%d, %d] 不合法", *m.ProtocolMinMajor, *m.ProtocolMaxMajor)
	}
	if m.RequiredCapabilities == nil {
		return nil, errors.New("缺少 required_capabilities（没有必需能力时写空列表）")
	}
	for _, c := range *m.RequiredCapabilities {
		if c.Name == "" || c.Version == 0 {
			return nil, fmt.Errorf("必需能力 %+v 缺少 name 或 version", c)
		}
	}
	if !SupportedPlatform(m.OS, m.Arch) {
		return nil, fmt.Errorf("不支持的平台 %s/%s", m.OS, m.Arch)
	}
	if err := ValidateRelPath(m.Entrypoint); err != nil {
		return nil, fmt.Errorf("entrypoint: %w", err)
	}
	if m.ConfigVersion == nil {
		return nil, errors.New("缺少 config_version")
	}
	if m.Source == "" {
		return nil, errors.New("缺少 source")
	}
	switch m.MigrationMode {
	case "", MigrationNone, MigrationSelf, MigrationVersioned:
	default:
		return nil, fmt.Errorf("未知的 migration_mode %q", m.MigrationMode)
	}
	switch m.PrepareEffects {
	case "", EffectsNone, EffectsManagedDB, EffectsExternal:
	default:
		return nil, fmt.Errorf("未知的 prepare_effects %q", m.PrepareEffects)
	}
	if m.MigrationMode == "" || m.PrepareEffects == "" {
		warnings = append(warnings, "未声明 migration_mode 或 prepare_effects，该包不允许自动升级")
	}
	return warnings, nil
}

// ValidateRelPath 校验包内相对路径：使用 / 分隔，非空、非绝对、无 ..、无反斜杠或盘符，且已是规范形式。
func ValidateRelPath(p string) error {
	switch {
	case p == "":
		return errors.New("路径为空")
	case strings.ContainsAny(p, "\\:\x00"):
		return fmt.Errorf("路径 %q 含反斜杠、盘符或 NUL", p)
	case path.IsAbs(p):
		return fmt.Errorf("路径 %q 是绝对路径", p)
	case path.Clean(p) != p:
		return fmt.Errorf("路径 %q 不是规范形式", p)
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." || seg == "." {
			return fmt.Errorf("路径 %q 含 %s", p, seg)
		}
	}
	return nil
}

// PostgresIdentifier 把插件 ID 映射为合法的 PostgreSQL 标识符（role/schema 名，CONTRACTS §6.3）：
// prefix + 清洗后的 ID（非 [a-z0-9_] 字符替换为 _）+ "_" + 原 ID 的 SHA256 前 8 位十六进制。
// 哈希后缀避免 "my-tool" 与 "my_tool" 撞名；总长度不超过 PostgreSQL 的 63 字节上限。
func PostgresIdentifier(prefix, id string) string {
	sum := sha256.Sum256([]byte(id))
	suffix := "_" + hex.EncodeToString(sum[:4])
	var b strings.Builder
	for _, r := range strings.ToLower(id) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	body := b.String()
	if max := 63 - len(prefix) - len(suffix); len(body) > max {
		body = body[:max]
	}
	return prefix + body + suffix
}
