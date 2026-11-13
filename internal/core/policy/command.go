package policy

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
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

// SetOverride 保存权限与启用覆盖，成功返回前已完成持久化；两项都为 nil 时恢复继承。
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
	// 2. 只保存权限与启用，保留命令名以及 nil 与显式空集合的区别。
	key := scopeKey(s)
	err := p.store.PutCommandPolicy(ctx, state.CommandPolicy{PluginID: pluginID, Scope: key, Permissions: permissions, Enabled: o.Enabled})
	if err == nil && p.commandsChanged != nil {
		p.commandsChanged()
	}
	return err
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
		// 合法 token 不含空白，空格分隔可保留原始字节与 token 边界。
		return groupScopePrefix + strings.Join(s.Group, " ")
	default:
		return routeScopePrefix + s.RouteID
	}
}

func storedScope(key string) (Scope, error) {
	switch {
	case key == pluginScopeKey:
		return Scope{Kind: ScopePlugin}, nil
	case strings.HasPrefix(key, groupScopePrefix):
		return Scope{Kind: ScopeGroup, Group: strings.Split(strings.TrimPrefix(key, groupScopePrefix), " ")}, nil
	case strings.HasPrefix(key, routeScopePrefix):
		return Scope{Kind: ScopeRoute, RouteID: strings.TrimPrefix(key, routeScopePrefix)}, nil
	default:
		return Scope{}, fmt.Errorf("unknown stored command scope %q", key)
	}
}

// CheckCommand 按全局禁止、逐层启用与最近显式权限判定当前命令。
// 每次检查读取同一状态库快照；返回 nil 放行，读取失败返回 INTERNAL。
func (p *Policy) CheckCommand(ctx context.Context, pluginID string, route *pb.RouteDescriptor, actor *pb.Actor, target *pb.Target) *pb.ErrorInfo {
	// 1. 全局禁止优先，管理员身份与空权限集合都不能绕过。
	var subjects []state.GlobalDeny
	if actor != nil {
		subjects = append(subjects, state.GlobalDeny{Kind: string(DenyUser), AdapterID: target.GetAdapterId(), SubjectID: actor.GetExternalUserId()})
	}
	if target != nil {
		subjects = append(subjects, state.GlobalDeny{Kind: string(DenyChat), AdapterID: target.GetAdapterId(), SubjectID: target.GetChatId()})
	}
	rows, denied, err := p.store.CommandPolicySnapshot(ctx, pluginID, subjects)
	if err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "read command policy: " + err.Error()}
	}
	if denied {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, SafeMessage: "command is globally denied"}
	}

	// 2. 与菜单共用同一份继承计算，命令派发保留全局禁止快照。
	enabled, permissions := commandState(rows, route)
	if !enabled {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "command is disabled"}
	}

	// 3. 最近的显式权限集合按交集判定；空集合也允许无 Actor 的命令。
	if len(permissions) == 0 {
		return nil
	}
	for _, permission := range permissions {
		if slices.Contains(actor.GetPermissions(), permission) {
			return nil
		}
	}
	return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, SafeMessage: "command permission is required"}
}

// SetCommandsChanged 在开始处理命令策略前装配持久化成功后的变化通知。
func (p *Policy) SetCommandsChanged(fn func()) { p.commandsChanged = fn }

// CommandState 返回逐层启用状态与最近显式权限，读取失败保留原错误。
func (p *Policy) CommandState(ctx context.Context, pluginID string, route *pb.RouteDescriptor) (enabled bool, permissions []string, err error) {
	rows, _, err := p.store.CommandPolicySnapshot(ctx, pluginID, nil)
	if err != nil {
		return false, nil, err
	}
	enabled, permissions = commandState(rows, route)
	return enabled, slices.Clone(permissions), nil
}

func commandState(rows []state.CommandPolicy, route *pb.RouteDescriptor) (bool, []string) {
	overrides := make(map[string]state.CommandPolicy, len(rows))
	for _, row := range rows {
		overrides[row.Scope] = row
	}
	keys := []string{pluginScopeKey}
	if command := route.GetCommand(); command.GetMatch() == pb.CommandMatch_COMMAND_MATCH_PATH {
		path := command.GetPath()
		for length := 1; length < len(path); length++ {
			keys = append(keys, scopeKey(Scope{Kind: ScopeGroup, Group: path[:length]}))
		}
	}
	routeKey := scopeKey(Scope{Kind: ScopeRoute, RouteID: route.GetRouteId()})
	keys = append(keys, routeKey)
	permissions := route.GetRequiredPermissions()
	enabled := true
	for _, key := range keys {
		override := overrides[key]
		if override.Enabled != nil && !*override.Enabled {
			enabled = false
		}
		if override.Permissions != nil {
			permissions = override.Permissions
		}
	}
	if overrides[routeKey].Enabled == nil && !route.GetDefaultEnabled() {
		enabled = false
	}

	return enabled, permissions
}
