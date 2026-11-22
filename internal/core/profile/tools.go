package profile

import (
	"cmp"
	"context"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Operation 标识插件内的稳定确认操作名。
type Operation struct {
	PluginID string `json:"plugin_id"`
	Name     string `json:"name"`
}

// ToolPolicy 指定工具或确认操作的策略；输入 inherit 删除这项覆盖。
type ToolPolicy struct {
	Tool      string     `json:"tool,omitempty"`
	Operation *Operation `json:"operation,omitempty"`
	Policy    string     `json:"policy"`
}

// ToolPolicies 编码工具策略集合供管理端读取。
type ToolPolicies struct {
	Items []ToolPolicy `json:"items"`
}

// ToolPolicy 从本次快照读取工具或确认操作的覆盖及所属插件开关。
// 缺少覆盖返回 UNSPECIFIED，缺少开关表示启用。
func (s *Service) ToolPolicy(id, pluginID, name string, operation bool) (policy pb.ToolPolicy, enabled bool, err error) {
	if _, err := s.find(id); err != nil {
		return pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED, false, err
	}
	enabled = true
	for _, row := range s.snapshot.Plugins {
		if row.ProfileID == id && row.PluginID == pluginID {
			enabled = row.Enabled
			break
		}
	}
	for _, row := range s.snapshot.Tools {
		if row.ProfileID != id || row.PluginID != pluginID || row.Name != name || row.Operation != operation {
			continue
		}
		switch row.Policy {
		case "allow":
			return pb.ToolPolicy_TOOL_POLICY_ALLOW, enabled, nil
		case "confirm":
			return pb.ToolPolicy_TOOL_POLICY_CONFIRM, enabled, nil
		case "deny":
			return pb.ToolPolicy_TOOL_POLICY_DENY, enabled, nil
		default:
			return policy, enabled, invalid("invalid stored tool policy: " + row.Policy)
		}
	}
	return policy, enabled, nil
}

// Tools 列出该配置文件已保存的策略覆盖。
func (s *Service) Tools(id string) (ToolPolicies, error) {
	if _, err := s.find(id); err != nil {
		return ToolPolicies{}, err
	}
	result := ToolPolicies{Items: []ToolPolicy{}}
	for _, row := range s.snapshot.Tools {
		if row.ProfileID == id {
			item := ToolPolicy{Policy: row.Policy}
			if row.Operation {
				item.Operation = &Operation{row.PluginID, row.Name}
			} else {
				item.Tool = row.PluginID + "/" + row.Name
			}
			result.Items = append(result.Items, item)
		}
	}
	slices.SortFunc(result.Items, func(a, b ToolPolicy) int { return cmp.Compare(toolKey(a), toolKey(b)) })
	return result, nil
}
func toolKey(item ToolPolicy) string {
	if item.Operation != nil {
		return item.Operation.PluginID + "/" + item.Operation.Name + "/operation"
	}
	return item.Tool + "/tool"
}

// PutTools 校验全部条目后原子更新所列策略，inherit 删除对应覆盖。
func (s *Service) PutTools(ctx context.Context, id string, policies ToolPolicies) error {
	if _, err := s.find(id); err != nil {
		return err
	}
	if policies.Items == nil {
		return invalid("items array is required")
	}
	rows := make([]state.ProfileTool, 0, len(policies.Items))
	for _, item := range policies.Items {
		switch item.Policy {
		case "allow", "confirm", "deny", "inherit":
		default:
			return invalid("invalid tool policy")
		}
		row := state.ProfileTool{ProfileID: id, Policy: item.Policy}
		if (item.Tool == "") == (item.Operation == nil) {
			return invalid("exactly one tool or operation is required")
		}
		if item.Operation != nil {
			row.PluginID, row.Name, row.Operation = item.Operation.PluginID, item.Operation.Name, true
			if row.Name == "" {
				return invalid("operation name is required")
			}
		} else {
			var found bool
			row.PluginID, row.Name, found = strings.Cut(item.Tool, "/")
			if !found || !validate.IsServiceName(row.Name) {
				return invalid("invalid tool name")
			}
		}
		if _, err := s.plugin(row.PluginID); err != nil {
			return err
		}
		rows = append(rows, row)
	}
	if err := s.opt.Store.PutProfileTools(ctx, rows); err != nil {
		return err
	}
	return s.refresh(ctx)
}
