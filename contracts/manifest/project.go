package manifest

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// ParseProject 严格解析工程 plugin.yaml，仅接受作者字段与 external 副作用声明。
// 构建字段按是否出现拒绝，成功结果的构建字段保持未设置。
func ParseProject(data []byte) (*Result, error) {
	return parse(data, true)
}

func checkProjectFields(declared map[string]yaml.Node) error {
	if err := forbidFields(declared, "project", "manifest_version", "protocol_min_major", "protocol_max_major", "required_capabilities", "os", "arch", "entrypoint", "config_version", "migration_mode"); err != nil {
		return err
	}
	if node, present := declared["prepare_effects"]; present {
		var effects string
		if err := node.Decode(&effects); err != nil {
			return fmt.Errorf("project prepare_effects: %w", err)
		}
		if effects != EffectsExternal {
			return fmt.Errorf("project prepare_effects must be external, got %q", effects)
		}
	}
	return nil
}
