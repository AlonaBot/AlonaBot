package profile

import (
	"cmp"
	"context"
	"slices"
	"strings"

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

// ToolPolicies 按 ADMIN-API §23 编码工具策略集合。
type ToolPolicies struct {
	Items []ToolPolicy `json:"items"`
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
