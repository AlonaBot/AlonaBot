package policy

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// ScopeKind 区分命令策略的覆盖层级。
type ScopeKind int

const (
	// ScopePlugin 覆盖插件的全部命令。
	ScopePlugin ScopeKind = iota
	// ScopeGroup 覆盖规范命令路径下的组。
	ScopeGroup
	// ScopeRoute 覆盖稳定路由 ID 对应的命令。
	ScopeRoute
)

const (
	pluginScopeKey   = "plugin"
	groupScopePrefix = "group:"
	routeScopePrefix = "route:"
)

// Scope 定位插件、命令组或路由；无关层级的字段不参与定位。
type Scope struct {
	Kind    ScopeKind
	Group   []string // ScopeGroup 使用规范命令路径的前缀 token
	RouteID string   // ScopeRoute 使用稳定路由 ID
}

// Override 保存显式覆盖，nil 表示继承上层。
type Override struct {
	Permissions []string // 非 nil 的空集合表示不要求权限
	Enabled     *bool
}

// ScopedOverride 将覆盖绑定到所属作用域。
type ScopedOverride struct {
	Scope    Scope
	Override Override
}

// SetOverride 保存一项覆盖，成功返回前已完成持久化；两项都为 nil 时删除。
// 重复或非法权限名返回 INVALID_ARGUMENT；有效列表按字典序保存。
func (p *Policy) SetOverride(ctx context.Context, pluginID string, s Scope, o Override) error {
	// 1. 在管理入口校验作用域与权限，失败时不触及状态库。
	if err := checkScope(s); err != nil {
		return err
	}
	permissions := slices.Clone(o.Permissions)
	slices.Sort(permissions)
	for i, permission := range permissions {
		if !validate.IsPermissionName(permission) {
			return invalidArgument("invalid permission name: " + permission)
		}
		if i > 0 && permissions[i-1] == permission {
			return invalidArgument("duplicate permission name: " + permission)
		}
	}
	// 2. 删除或整体覆盖当前项，保留 nil 与显式空集合的区别。
	key := scopeKey(s)
	if o.Permissions == nil && o.Enabled == nil {
		return p.store.DeleteCommandPolicy(ctx, pluginID, key)
	}
	return p.store.PutCommandPolicy(ctx, state.CommandPolicy{PluginID: pluginID, Scope: key, Permissions: permissions, Enabled: o.Enabled})
}

// Overrides 按插件、组路径、路由 ID 的顺序读取覆盖，每一类内按字典序排列。
func (p *Policy) Overrides(ctx context.Context, pluginID string) ([]ScopedOverride, error) {
	rows, err := p.store.CommandPolicies(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	var overrides []ScopedOverride
	for _, row := range rows {
		s, err := storedScope(row.Scope)
		if err != nil {
			return overrides, err
		}
		overrides = append(overrides, ScopedOverride{Scope: s, Override: Override{Permissions: row.Permissions, Enabled: row.Enabled}})
	}
	slices.SortFunc(overrides, func(a, b ScopedOverride) int {
		if order := cmp.Compare(a.Scope.Kind, b.Scope.Kind); order != 0 {
			return order
		}
		if a.Scope.Kind == ScopeGroup {
			return slices.Compare(a.Scope.Group, b.Scope.Group)
		}
		return cmp.Compare(a.Scope.RouteID, b.Scope.RouteID)
	})
	return overrides, nil
}

func checkScope(s Scope) error {
	switch s.Kind {
	case ScopePlugin:
	case ScopeGroup:
		if len(s.Group) == 0 {
			return invalidArgument("command group is required")
		}
		for _, token := range s.Group {
			if !validate.IsCommandToken(token) {
				return invalidArgument("invalid command group token: " + token)
			}
		}
	case ScopeRoute:
		if s.RouteID == "" {
			return invalidArgument("route ID is required")
		}
	default:
		return invalidArgument("unknown scope kind")
	}
	return nil
}

func scopeKey(s Scope) string {
	switch s.Kind {
	case ScopePlugin:
		return pluginScopeKey
	case ScopeGroup:
		// JSON 数组保留 token 边界，即使 token 内带有分隔符也不会产生相同主键。
		encoded, _ := json.Marshal(s.Group) // []string 的 JSON 编码不会失败
		return groupScopePrefix + string(encoded)
	default:
		return routeScopePrefix + s.RouteID
	}
}

func storedScope(key string) (Scope, error) {
	switch {
	case key == pluginScopeKey:
		return Scope{Kind: ScopePlugin}, nil
	case strings.HasPrefix(key, groupScopePrefix):
		var group []string
		if err := json.Unmarshal([]byte(strings.TrimPrefix(key, groupScopePrefix)), &group); err != nil {
			return Scope{}, fmt.Errorf("decode command group scope: %w", err)
		}
		return Scope{Kind: ScopeGroup, Group: group}, nil
	case strings.HasPrefix(key, routeScopePrefix):
		return Scope{Kind: ScopeRoute, RouteID: strings.TrimPrefix(key, routeScopePrefix)}, nil
	default:
		return Scope{}, fmt.Errorf("unknown stored command scope %q", key)
	}
}
