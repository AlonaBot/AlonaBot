package manifest

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// AdapterDependency 声明所依赖的平台类型、extension 与动作版本。
type AdapterDependency struct {
	AdapterID  string                 `yaml:"adapter_id"`
	Extensions []ExtensionRequirement `yaml:"extensions"`
	Actions    []ActionRequirement    `yaml:"actions"`
	Required   bool                   `yaml:"required"` // false 表示可选依赖
}

// ExtensionRequirement 指定一个 extension 的名称与精确版本。
type ExtensionRequirement struct {
	Name    string `yaml:"name"`
	Version uint32 `yaml:"version"`
}

// ActionRequirement 指定平台动作的完整名称与精确版本。
type ActionRequirement struct {
	Name    string `yaml:"name"`
	Version uint32 `yaml:"version"`
}

// PluginDependency 声明所依赖的插件与最低版本。
type PluginDependency struct {
	PluginID   string `yaml:"plugin_id"`
	MinVersion string `yaml:"min_version"` // 空串表示不限版本
	Required   bool   `yaml:"required"`
}

// Installed 提供依赖检查所需的安装与启用快照。
type Installed struct {
	Plugins  map[string]string                 // 只含已启用插件的 ID 与版本
	Adapters map[string][]ExtensionRequirement // 已安装的平台类型与已声明的 extension
	Actions  map[string][]ActionRequirement    // 按 adapter_id 索引已声明的动作
}

// Problem 报告一项依赖缺失或版本不足。
type Problem struct {
	Dependency string
	Required   bool // 调用方据此拒绝启用或仅记录诊断
	Reason     string
}

func (m *Manifest) validateDependencies() error {
	// 1. 校验 Adapter ID、extension 与动作的名称和版本，并拒绝重复声明。
	adapters := make(map[string]bool)
	for _, dependency := range m.AdapterDependencies {
		if err := checkAdapterID(dependency.AdapterID); err != nil {
			return fmt.Errorf("adapter_dependencies: %w", err)
		}
		if adapters[dependency.AdapterID] {
			return fmt.Errorf("duplicate Adapter dependency %q", dependency.AdapterID)
		}
		adapters[dependency.AdapterID] = true
		extensions := make(map[ExtensionRequirement]bool)
		for _, extension := range dependency.Extensions {
			if extension.Name == "" || extension.Version == 0 {
				return fmt.Errorf("adapter dependency %q extension is missing name or version", dependency.AdapterID)
			}
			if extensions[extension] {
				return fmt.Errorf("adapter dependency %q repeats extension %q version %d", dependency.AdapterID, extension.Name, extension.Version)
			}
			extensions[extension] = true
		}
		actions := make(map[ActionRequirement]bool)
		for _, action := range dependency.Actions {
			if action.Name == "" || action.Version == 0 {
				return fmt.Errorf("adapter dependency %q action is missing name or version", dependency.AdapterID)
			}
			if actions[action] {
				return fmt.Errorf("adapter dependency %q repeats action %q version %d", dependency.AdapterID, action.Name, action.Version)
			}
			actions[action] = true
		}
	}
	// 2. 校验插件 ID 与最低版本，并拒绝重复依赖和必需的直接自依赖。
	plugins := make(map[string]bool)
	for _, dependency := range m.PluginDependencies {
		if _, err := componentID(dependency.PluginID, rolePlugin); err != nil {
			return fmt.Errorf("plugin_dependencies: %w", err)
		}
		if plugins[dependency.PluginID] {
			return fmt.Errorf("duplicate plugin dependency %q", dependency.PluginID)
		}
		plugins[dependency.PluginID] = true
		if dependency.MinVersion != "" && !ValidVersion(dependency.MinVersion) {
			return fmt.Errorf("plugin dependency %q min_version %q is not valid SemVer", dependency.PluginID, dependency.MinVersion)
		}
		if dependency.Required && dependency.PluginID == m.ID {
			return fmt.Errorf("required plugin dependency %q refers to itself", dependency.PluginID)
		}
	}
	return nil
}

// CheckDependencies 根据已校验的 manifest 与安装快照报告全部依赖问题，不修改启用状态。
// 插件版本使用 SemVer 比较；可选依赖的问题同样返回，由调用方记录诊断。
func CheckDependencies(m *Manifest, in Installed) []Problem {
	var problems []Problem
	// 1. 检查 Adapter 已安装且声明了所需 extension 与动作的精确版本。
	for _, dependency := range m.AdapterDependencies {
		extensions, installed := in.Adapters[dependency.AdapterID]
		if !installed {
			problems = append(problems, Problem{dependency.AdapterID, dependency.Required, "adapter is not installed"})
			continue
		}
		for _, extension := range dependency.Extensions {
			if !slices.Contains(extensions, extension) {
				problems = append(problems, Problem{dependency.AdapterID, dependency.Required, fmt.Sprintf("extension %q version %d is not declared", extension.Name, extension.Version)})
			}
		}
		for _, action := range dependency.Actions {
			if !slices.Contains(in.Actions[dependency.AdapterID], action) {
				problems = append(problems, Problem{dependency.AdapterID, dependency.Required, fmt.Sprintf("action %q version %d is not declared", action.Name, action.Version)})
			}
		}
	}
	// 2. 检查插件已启用且达到最低版本。
	for _, dependency := range m.PluginDependencies {
		version, enabled := in.Plugins[dependency.PluginID]
		if !enabled {
			problems = append(problems, Problem{dependency.PluginID, dependency.Required, "plugin is not installed or enabled"})
		} else if dependency.MinVersion != "" && compareVersions(version, dependency.MinVersion) < 0 {
			problems = append(problems, Problem{dependency.PluginID, dependency.Required, fmt.Sprintf("version %q is lower than min_version %q", version, dependency.MinVersion)})
		}
	}
	return problems
}

func compareVersions(a, b string) int {
	// 1. 先按数值比较主版本、次版本与修订版本。
	left, right := semver.FindStringSubmatch(a), semver.FindStringSubmatch(b)
	for i := 1; i <= 3; i++ {
		if order := compareNumericIdentifier(left[i], right[i]); order != 0 {
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
			order = compareNumericIdentifier(l, r)
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

func compareNumericIdentifier(a, b string) int {
	// SemVer 数字不含前导零；先比较位数，避免大版本号溢出整数类型。
	if order := cmp.Compare(len(a), len(b)); order != 0 {
		return order
	}
	return strings.Compare(a, b)
}
