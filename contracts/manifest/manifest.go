// Package manifest 解析工程与包内的 plugin.yaml、生成包内声明并检查组件依赖。
//
// 解析是严格的：未知字段、多个 YAML 文档、缺失必填项或非法取值都会失败。
// [ParseProject] 读取作者信息，[Generate] 补入构建信息，[Parse] 校验包内声明。
// 三个入口共用 [Manifest] 的角色、ID、版本与相对路径规则，见 spec/CONTRACTS-v0.1.md §9.1。
// [CheckDependencies] 根据安装快照报告缺失或版本不足的依赖；本包不读取归档或安装记录。
package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
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

var (
	// SemVer 2.0 官方正则。
	semver                = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	marketComponentSuffix = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	localAdapterID        = regexp.MustCompile(fmt.Sprintf(`^[a-z][a-z0-9_-]{0,%d}$`, validate.MaxPluginIDLen-1))
	reservedAdapterID     = regexp.MustCompile(`^(con|prn|aux|nul|com[1-9]|lpt[1-9])$`)
)

const (
	rolePlugin  = "plugin"
	roleAdapter = "adapter"
	roleAgent   = "agent"
	roleWebUI   = "webui"
)

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
	Name                 string              `yaml:"name"`        // 英文源显示名，必填
	Description          string              `yaml:"description"` // 英文源描述，可选
	Author               string              `yaml:"author"`      // 作者，可选
	Role                 string              `yaml:"role"`        // 缺省为 plugin
	AdapterID            string              `yaml:"adapter_id"`  // adapter 角色的平台类型
	Icon                 string              `yaml:"icon"`        // 包内图标的相对路径，可选
	Assets               string              `yaml:"assets"`      // webui 的包内静态根目录
	AdminAPIMinMajor     uint32              `yaml:"admin_api_min_major"`
	AdminAPIMaxMajor     uint32              `yaml:"admin_api_max_major"`
	AdapterDependencies  []AdapterDependency `yaml:"adapter_dependencies"`
	PluginDependencies   []PluginDependency  `yaml:"plugin_dependencies"`
	ManifestVersion      *int                `yaml:"manifest_version"`      // manifest 格式版本，目前必须为 1
	ID                   string              `yaml:"id"`                    // 组件 ID，市场前缀与 Role 一致
	Version              string              `yaml:"version"`               // 组件版本，SemVer 2.0
	ProtocolMinMajor     *uint32             `yaml:"protocol_min_major"`    // 支持的最低协议主版本，不能为 0
	ProtocolMaxMajor     *uint32             `yaml:"protocol_max_major"`    // 支持的最高协议主版本，不小于 ProtocolMinMajor
	RequiredCapabilities *[]Capability       `yaml:"required_capabilities"` // 必需的协议能力；没有时必须写空列表
	OS                   string              `yaml:"os"`                    // 包的目标 GOOS
	Arch                 string              `yaml:"arch"`                  // 包的目标 GOARCH
	Entrypoint           string              `yaml:"entrypoint"`            // 包内可执行文件的相对路径，规则见 ValidateRelPath
	ConfigVersion        *uint32             `yaml:"config_version"`        // 包内配置 Schema 的版本
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

// Parse 严格解析包内 plugin.yaml，要求完整构建信息及非空 name。
func Parse(data []byte) (*Result, error) {
	return parse(data, false)
}

func parse(data []byte, project bool) (*Result, error) {
	// 1. 按字段出现检查工程声明，避免零值、空串与 null 隐去构建字段。
	var declared map[string]yaml.Node
	if err := yaml.Unmarshal(data, &declared); err != nil {
		return nil, fmt.Errorf("parse plugin.yaml fields: %w", err)
	}
	if project {
		if err := checkProjectFields(declared); err != nil {
			return nil, err
		}
	}
	// 2. 严格解码单份 YAML 文档，拒绝未知字段和额外文档。
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("parse plugin.yaml: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("plugin.yaml must contain exactly one YAML document")
	}
	// 3. 两种声明共用作者字段与角色校验。
	warnings, err := m.validate(declared, project)
	if err != nil {
		return nil, err
	}
	return &Result{Manifest: &m, Warnings: warnings}, nil
}

// validate 逐项校验字段，返回非致命警告；任一项不合法即返回错误。
func (m *Manifest) validate(declared map[string]yaml.Node, project bool) ([]string, error) {
	// 1. 检查所有角色共用的身份、版本、来源与图标。
	if !project && (m.ManifestVersion == nil || *m.ManifestVersion != 1) {
		return nil, errors.New("manifest_version must be 1")
	}
	if _, present := declared["role"]; !present {
		m.Role = rolePlugin
	}
	switch m.Role {
	case rolePlugin, roleAdapter, roleAgent, roleWebUI:
	default:
		return nil, fmt.Errorf("unknown role %q", m.Role)
	}
	warnings, err := componentID(m.ID, m.Role)
	if err != nil {
		return nil, err
	}
	if m.Name == "" {
		return nil, errors.New("missing name")
	}
	if !ValidVersion(m.Version) {
		return nil, fmt.Errorf("version %q is not valid SemVer", m.Version)
	}
	if m.Source == "" {
		return nil, errors.New("missing source")
	}
	if m.Icon != "" {
		if err := ValidateRelPath(m.Icon); err != nil {
			return nil, fmt.Errorf("icon: %w", err)
		}
		switch path.Ext(m.Icon) {
		case ".svg", ".png", ".webp":
		default:
			return nil, fmt.Errorf("icon %q must have an .svg, .png or .webp extension", m.Icon)
		}
	}
	// 2. 静态组件单独校验，通过后返回；其余角色继续检查进程字段。
	if m.Role == roleWebUI {
		if err := forbidFields(declared, m.Role, "adapter_id", "entrypoint", "protocol_min_major", "protocol_max_major", "required_capabilities", "config_version", "migration_mode", "prepare_effects", "adapter_dependencies", "plugin_dependencies"); err != nil {
			return nil, err
		}
		if !project && (m.OS != "any" || m.Arch != "any") {
			return nil, errors.New("webui requires os and arch to be any")
		}
		if err := ValidateRelPath(m.Assets); err != nil {
			return nil, fmt.Errorf("assets: %w", err)
		}
		if m.AdminAPIMinMajor == 0 || m.AdminAPIMinMajor > m.AdminAPIMaxMajor {
			return nil, fmt.Errorf("invalid Admin API major version range [%d, %d]", m.AdminAPIMinMajor, m.AdminAPIMaxMajor)
		}
		return warnings, nil
	}
	// 3. 检查进程角色的专用字段、协议能力、平台、入口与配置版本。
	if err := forbidFields(declared, m.Role, "assets", "admin_api_min_major", "admin_api_max_major"); err != nil {
		return nil, err
	}
	if m.Role == roleAdapter {
		if err := checkAdapterID(m.AdapterID); err != nil {
			return nil, fmt.Errorf("adapter_id: %w", err)
		}
		if err := forbidFields(declared, m.Role, "adapter_dependencies", "plugin_dependencies", "migration_mode", "prepare_effects"); err != nil {
			return nil, err
		}
	} else if err := forbidFields(declared, m.Role, "adapter_id"); err != nil {
		return nil, err
	}
	if !project {
		if err := m.validateProcess(); err != nil {
			return nil, err
		}
	}
	if m.Role == roleAdapter {
		return warnings, nil
	}
	// 4. 插件与 Agent 还需校验依赖、迁移方式和准备期副作用。
	if err := m.validateDependencies(); err != nil {
		return nil, err
	}
	if project {
		return warnings, nil
	}
	switch m.MigrationMode {
	case "", MigrationNone, MigrationSelf, MigrationVersioned:
	default:
		return nil, fmt.Errorf("unknown migration_mode %q", m.MigrationMode)
	}
	switch m.PrepareEffects {
	case "", EffectsNone, EffectsManagedDB, EffectsExternal:
	default:
		return nil, fmt.Errorf("unknown prepare_effects %q", m.PrepareEffects)
	}
	if m.MigrationMode == "" {
		warnings = append(warnings, "migration_mode not declared; preparation and rollback assume no external side effects")
	}
	if m.PrepareEffects == "" {
		warnings = append(warnings, "prepare_effects not declared; preparation and rollback assume no external side effects")
	}
	return warnings, nil
}

func (m *Manifest) validateProcess() error {
	if m.ProtocolMinMajor == nil || m.ProtocolMaxMajor == nil {
		return errors.New("missing protocol_min_major or protocol_max_major")
	}
	if *m.ProtocolMinMajor == 0 || *m.ProtocolMinMajor > *m.ProtocolMaxMajor {
		return fmt.Errorf("invalid protocol major version range [%d, %d]", *m.ProtocolMinMajor, *m.ProtocolMaxMajor)
	}
	if m.RequiredCapabilities == nil {
		return errors.New("missing required_capabilities (use an empty list when there are none)")
	}
	for _, c := range *m.RequiredCapabilities {
		if c.Name == "" || c.Version == 0 {
			return fmt.Errorf("required capability %+v is missing name or version", c)
		}
	}
	if !SupportedPlatform(m.OS, m.Arch) {
		return fmt.Errorf("unsupported platform %s/%s", m.OS, m.Arch)
	}
	if err := ValidateRelPath(m.Entrypoint); err != nil {
		return fmt.Errorf("entrypoint: %w", err)
	}
	if m.ConfigVersion == nil {
		return errors.New("missing config_version")
	}
	return nil
}

func componentID(id, role string) ([]string, error) {
	warnings, err := validate.PluginID(id)
	if err != nil {
		return nil, err
	}
	for _, marketRole := range []string{rolePlugin, roleAdapter, roleAgent, roleWebUI} {
		prefix := "alonabot-" + marketRole + "-"
		if suffix, market := strings.CutPrefix(id, prefix); market {
			if marketRole != role || !marketComponentSuffix.MatchString(suffix) {
				return nil, fmt.Errorf("component ID %q must match the market prefix and ID rule for role %q", id, role)
			}
			return nil, nil
		}
	}
	if role != rolePlugin {
		warnings = []string{fmt.Sprintf("component ID %q lacks the market prefix alonabot-%s- and can only be installed locally", id, role)}
	}
	return warnings, nil
}

func checkAdapterID(id string) error {
	if !localAdapterID.MatchString(id) || reservedAdapterID.MatchString(id) {
		return fmt.Errorf("invalid local Adapter ID %q", id)
	}
	return nil
}

func forbidFields(declared map[string]yaml.Node, role string, names ...string) error {
	for _, name := range names {
		if _, present := declared[name]; present {
			return fmt.Errorf("role %q must not declare %s", role, name)
		}
	}
	return nil
}

// ValidateRelPath 校验包内相对路径：使用 / 分隔，非空、非绝对、无 ..、无反斜杠或盘符，且已是规范形式。
func ValidateRelPath(p string) error {
	switch {
	case p == "":
		return errors.New("empty path")
	case strings.ContainsAny(p, "\\:\x00"):
		return fmt.Errorf("path %q contains a backslash, drive letter or NUL", p)
	case path.IsAbs(p):
		return fmt.Errorf("path %q is absolute", p)
	case path.Clean(p) != p:
		return fmt.Errorf("path %q is not in canonical form", p)
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." || seg == "." {
			return fmt.Errorf("path %q contains %s", p, seg)
		}
	}
	return nil
}
