package manifest

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// BuildInfo 提供生成包内 manifest 的构建值；指针字段保留缺失与零值的区别。
type BuildInfo struct {
	ManifestVersion      int           // manifest_version，必须为 1
	ProtocolMinMajor     *uint32       // 进程角色必填
	ProtocolMaxMajor     *uint32       // 进程角色必填
	RequiredCapabilities *[]Capability // 进程角色必填，无能力时指向空列表
	OS, Arch             string        // 纯静态 WebUI 必须为 any/any
	Entrypoint           string        // 进程角色的包内相对路径
	MigrationMode        string        // 非 Adapter 进程必须为 none、self 或 versioned
	PrepareEffects       string        // 非 Adapter 进程必须为 none 或 managed_db
}

// Generate 将已解析的工程 manifest 与构建信息合成为稳定的包内 YAML。
// 作者字段保持原值，external 优先于构建副作用；返回前通过 Parse 校验。
func Generate(project *Manifest, build BuildInfo) ([]byte, error) {
	// 1. 在副本中合入构建值，保留作者显式声明的 external。
	m := *project
	m.ManifestVersion = &build.ManifestVersion
	m.ProtocolMinMajor = build.ProtocolMinMajor
	m.ProtocolMaxMajor = build.ProtocolMaxMajor
	m.RequiredCapabilities = build.RequiredCapabilities
	m.OS, m.Arch = build.OS, build.Arch
	m.Entrypoint = build.Entrypoint
	m.MigrationMode = build.MigrationMode
	staticWebUI := m.Role == roleWebUI && m.Entrypoint == ""
	if staticWebUI && m.webUIProcessField != "" {
		return nil, fmt.Errorf("static webui must not declare %s", m.webUIProcessField)
	}
	if staticWebUI && (len(m.AdapterDependencies) != 0 || len(m.PluginDependencies) != 0 || m.PrepareEffects != "") {
		return nil, fmt.Errorf("static webui must not declare dependencies or prepare_effects")
	}
	if m.Role != roleAdapter && !staticWebUI {
		switch build.MigrationMode {
		case MigrationNone, MigrationSelf, MigrationVersioned:
		default:
			return nil, fmt.Errorf("invalid build migration_mode %q", build.MigrationMode)
		}
		switch build.PrepareEffects {
		case EffectsNone, EffectsManagedDB:
		default:
			return nil, fmt.Errorf("invalid build prepare_effects %q", build.PrepareEffects)
		}
		if m.PrepareEffects != EffectsExternal {
			m.PrepareEffects = build.PrepareEffects
		}
	}
	// 2. 沿用结构体字段顺序，按角色移除不允许输出的字段。
	var node yaml.Node
	if err := node.Encode(&m); err != nil {
		return nil, fmt.Errorf("encode package manifest: %w", err)
	}
	fields := node.Content[:0]
	for i := 0; i < len(node.Content); i += 2 {
		name := node.Content[i].Value
		switch name {
		case "adapter_id":
			if m.Role != roleAdapter {
				continue
			}
		case "assets", "admin_api_min_major", "admin_api_max_major":
			if m.Role != roleWebUI {
				continue
			}
		case "sleep_capable":
			if !m.SleepCapable && m.Role != rolePlugin && m.Role != roleAgent {
				continue
			}
		case "adapter_dependencies", "plugin_dependencies", "migration_mode", "prepare_effects":
			if m.Role == roleAdapter || staticWebUI {
				continue
			}
		case "protocol_min_major", "protocol_max_major", "required_capabilities", "entrypoint":
			if staticWebUI {
				continue
			}
		}
		fields = append(fields, node.Content[i], node.Content[i+1])
	}
	node.Content = fields
	data, err := yaml.Marshal(&node)
	if err != nil {
		return nil, fmt.Errorf("marshal package manifest: %w", err)
	}
	// 3. 输出与安装入口使用相同的严格校验，保留完整错误。
	if _, err := Parse(data); err != nil {
		return nil, err
	}
	return data, nil
}
