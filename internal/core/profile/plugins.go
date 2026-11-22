package profile

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Invalid 定位插件更新后失效的覆盖；原值留在状态库供清理。
type Invalid struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// Plugin 返回配置文件内的插件开关、允许覆盖的字段与可公开的已保存覆盖。
type Plugin struct {
	PluginID  string                     `json:"plugin_id"`
	Enabled   bool                       `json:"enabled"`
	Fields    []string                   `json:"fields"`
	Overrides map[string]json.RawMessage `json:"overrides"`
	Invalid   []Invalid                  `json:"invalid"`
}

// PluginPatch 同时编辑插件开关与字段覆盖；JSON null 删除覆盖。
type PluginPatch struct {
	Enabled   *bool                      `json:"enabled"`
	Overrides map[string]json.RawMessage `json:"overrides"`
}

func (s *Service) invalidReason(pluginID, path string, value json.RawMessage) string {
	schema := s.schemas[pluginID]
	if schema == nil {
		return "schema unavailable"
	}
	if !slices.Contains(schema.ProfilePaths(), path) {
		return "unmarked"
	}
	if err := schema.ValidateProfileValue(path, value); err != nil {
		return "type"
	}
	return ""
}

// Overrides 返回全部配置文件中仍符合 Active schema 的覆盖值副本。
func (s *Service) Overrides(pluginID string) map[string]map[string]json.RawMessage {
	result := map[string]map[string]json.RawMessage{}
	for _, row := range s.snapshot.Overrides {
		value := json.RawMessage(row.Value)
		if row.PluginID != pluginID || row.ProfileID == Default || s.invalidReason(pluginID, row.Path, value) != "" {
			continue
		}
		if result[row.ProfileID] == nil {
			result[row.ProfileID] = map[string]json.RawMessage{}
		}
		result[row.ProfileID][row.Path] = bytes.Clone(value)
	}
	return result
}

// Plugins 列出已安装插件与 Agent，另列失效覆盖；秘密及 schema 不可用时的原值不公开。
func (s *Service) Plugins(id string) ([]Plugin, error) {
	if _, err := s.find(id); err != nil {
		return nil, err
	}
	result := []Plugin{}
	for _, in := range s.installs {
		if in.Uninstalled || in.Role != state.RolePlugin && in.Role != state.RoleAgent {
			continue
		}
		plugin := Plugin{PluginID: in.PluginID, Enabled: true, Fields: []string{}, Overrides: map[string]json.RawMessage{}, Invalid: []Invalid{}}
		if schema := s.schemas[in.PluginID]; schema != nil {
			plugin.Fields = schema.ProfilePaths()
		}
		for _, row := range s.snapshot.Plugins {
			if row.ProfileID == id && row.PluginID == in.PluginID {
				plugin.Enabled = row.Enabled
			}
		}
		if id != Default {
			for _, row := range s.snapshot.Overrides {
				if row.ProfileID == id && row.PluginID == in.PluginID {
					value := json.RawMessage(row.Value)
					if s.publicOverride(in.PluginID, row.Path) {
						plugin.Overrides[row.Path] = value
					}
					if reason := s.invalidReason(in.PluginID, row.Path, value); reason != "" {
						plugin.Invalid = append(plugin.Invalid, Invalid{row.Path, reason})
					}
				}
			}
		}
		slices.SortFunc(plugin.Invalid, func(a, b Invalid) int { return cmp.Compare(a.Path, b.Path) })
		result = append(result, plugin)
	}
	return result, nil
}

func (s *Service) publicOverride(pluginID, path string) bool {
	schema := s.schemas[pluginID]
	if schema == nil {
		return false
	}
	parts := strings.Split(path, ".")
	for _, secret := range schema.SecretPaths() {
		n := min(len(parts), len(secret))
		matches := true
		for i := range n {
			if secret[i] != "*" && secret[i] != parts[i] {
				matches = false
				break
			}
		}
		if matches {
			return false // 整项隐藏，避免复合值或旧子路径绕过秘密保护。
		}
	}
	return true
}

// PatchPlugin 先校验全部覆盖与序列化集合大小，再原子写入开关与覆盖。
func (s *Service) PatchPlugin(ctx context.Context, id, pluginID string, patch PluginPatch) error {
	if _, err := s.find(id); err != nil {
		return err
	}
	if _, err := s.plugin(pluginID); err != nil {
		return err
	}
	if id == Default && len(patch.Overrides) > 0 {
		return invalid("default overrides must be edited through pending config")
	}
	// 1. null 允许移除已失效的旧字段；新覆盖必须按 Active schema 校验。
	all := s.Overrides(pluginID)
	for path, value := range patch.Overrides {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			delete(all[id], path)
			continue
		}
		schema := s.schemas[pluginID]
		if schema == nil {
			return invalid("profile validation failed at " + path + ": active schema is unavailable")
		}
		if err := schema.ValidateProfileValue(path, value); err != nil {
			return err
		}
		if all[id] == nil {
			all[id] = map[string]json.RawMessage{}
		}
		all[id][path] = value
	}
	// 2. 每个插件的全部配置文件共用同一 4 MiB 覆盖集合预算。
	for profile, values := range all {
		if len(values) == 0 {
			delete(all, profile)
		}
	}
	if err := checkOverridesLimit(all); err != nil {
		return err
	}
	normalized := make(map[string]json.RawMessage, len(patch.Overrides))
	for path, value := range patch.Overrides {
		var compact bytes.Buffer
		if err := json.Compact(&compact, value); err != nil {
			return err
		}
		normalized[path] = compact.Bytes()
	}
	if err := s.opt.Store.PatchProfilePlugin(ctx, id, pluginID, patch.Enabled, normalized); err != nil {
		return err
	}
	return s.refresh(ctx)
}

func checkOverridesLimit(all map[string]map[string]json.RawMessage) error {
	raw, err := json.Marshal(all)
	if err != nil {
		return err
	}
	if len(raw) > wire.MaxFrame {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "profile overrides exceed the frame limit"}
	}
	return nil
}

// Prune 只删除此快照中列为失效的覆盖，返回实际删除的路径。
func (s *Service) Prune(ctx context.Context, id, pluginID string) ([]string, error) {
	plugins, err := s.Plugins(id)
	if err != nil {
		return nil, err
	}
	if _, err := s.plugin(pluginID); err != nil {
		return nil, err
	}
	removed := []string{}
	values := map[string]json.RawMessage{}
	for _, plugin := range plugins {
		if plugin.PluginID == pluginID {
			for _, item := range plugin.Invalid {
				removed = append(removed, item.Path)
				values[item.Path] = json.RawMessage("null")
			}
		}
	}
	if err := s.opt.Store.PatchProfilePlugin(ctx, id, pluginID, nil, values); err != nil {
		return nil, err
	}
	return removed, s.refresh(ctx)
}
