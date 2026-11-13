package policy

import (
	"context"
	"fmt"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// LoadCommandNames 从状态库整体载入命令名快照，在创建 Router 前调用。
func (p *Policy) LoadCommandNames(ctx context.Context) error {
	p.namesMu.Lock()
	defer p.namesMu.Unlock()
	rows, err := p.store.AllCommandNames(ctx)
	if err != nil {
		return err
	}
	next := make(map[string]map[string][][]string)
	for _, row := range rows {
		s, err := storedScope(row.Scope)
		if err != nil {
			return err
		}
		if s.Kind != ScopeRoute {
			return fmt.Errorf("command names stored outside route scope: %q", row.Scope)
		}
		if next[row.PluginID] == nil {
			next[row.PluginID] = make(map[string][][]string)
		}
		next[row.PluginID][s.RouteID] = row.Names
	}
	p.names = next
	return nil
}

// SetCommandNames 校验并保存一条路径命令的全部命令名，更新快照后通知 Router。
// routes 为 nil 时允许为未注册的插件保存；冲突错误带出对应 route_id。
func (p *Policy) SetCommandNames(ctx context.Context, pluginID, routeID string, names [][]string, routes []*pb.RouteDescriptor) error {
	// 1. 管理入口先校验格式与注册声明，失败时不写库或修改快照。
	if err := checkScope(Scope{Kind: ScopeRoute, RouteID: routeID}); err != nil {
		return err
	}
	if len(names) == 0 {
		return invalidArgument("command names are required")
	}
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if len(name) == 0 {
			return invalidArgument("command name is required")
		}
		for _, token := range name {
			if !validate.IsCommandToken(token) {
				return invalidArgument("invalid command token: " + token)
			}
		}
		key := strings.Join(name, " ")
		if seen[key] {
			return invalidArgument("duplicate command name: " + key)
		}
		seen[key] = true
	}
	if routes != nil {
		target := slices.IndexFunc(routes, func(route *pb.RouteDescriptor) bool { return route.GetRouteId() == routeID })
		if target < 0 || routes[target].GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND || routes[target].GetCommand().GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH {
			return invalidArgument("route is not a registered path command: " + routeID)
		}
		for _, route := range routes {
			if route.GetRouteId() == routeID {
				continue
			}
			command := route.GetCommand()
			if route.GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND || command.GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH {
				continue
			}
			if seen[strings.Join(command.GetPath(), " ")] {
				return commandNameConflict(route.GetRouteId())
			}
			for _, alias := range command.GetAliases() {
				if seen[strings.Join(alias.GetPath(), " ")] {
					return commandNameConflict(route.GetRouteId())
				}
			}
		}
	}
	// 2. 在同一快照锁内检查其他覆盖并持久化，避免并发保存同名。
	if err := p.saveCommandNames(ctx, pluginID, routeID, names); err != nil {
		return err
	}
	if p.commandsChanged != nil {
		p.commandsChanged()
	}
	return nil
}

// DeleteCommandNames 清除命令名覆盖，恢复声明的名字并通知 Router。
func (p *Policy) DeleteCommandNames(ctx context.Context, pluginID, routeID string) error {
	if err := checkScope(Scope{Kind: ScopeRoute, RouteID: routeID}); err != nil {
		return err
	}
	if err := p.saveCommandNames(ctx, pluginID, routeID, nil); err != nil {
		return err
	}
	if p.commandsChanged != nil {
		p.commandsChanged()
	}
	return nil
}

func (p *Policy) saveCommandNames(ctx context.Context, pluginID, routeID string, names [][]string) error {
	p.namesMu.Lock()
	defer p.namesMu.Unlock()
	for otherID, otherNames := range p.names[pluginID] {
		if otherID == routeID {
			continue
		}
		for _, name := range names {
			if slices.ContainsFunc(otherNames, func(other []string) bool { return slices.Equal(name, other) }) {
				return commandNameConflict(otherID)
			}
		}
	}
	key := scopeKey(Scope{Kind: ScopeRoute, RouteID: routeID})
	if err := p.store.PutCommandNames(ctx, pluginID, key, names); err != nil {
		return err
	}
	if names == nil {
		delete(p.names[pluginID], routeID)
		return nil
	}
	if p.names == nil {
		p.names = make(map[string]map[string][][]string)
	}
	if p.names[pluginID] == nil {
		p.names[pluginID] = make(map[string][][]string)
	}
	p.names[pluginID][routeID] = cloneCommandNames(names)
	return nil
}

// CommandNames 返回插件命令名快照的深副本，供 Router 持锁调用，全程不做 I/O。
func (p *Policy) CommandNames(pluginID string) map[string][][]string {
	p.namesMu.RLock()
	defer p.namesMu.RUnlock()
	names := make(map[string][][]string, len(p.names[pluginID]))
	for routeID, routeNames := range p.names[pluginID] {
		names[routeID] = cloneCommandNames(routeNames)
	}
	return names
}

func cloneCommandNames(names [][]string) [][]string {
	cloned := make([][]string, len(names))
	for i, name := range names {
		cloned[i] = slices.Clone(name)
	}
	return cloned
}

func commandNameConflict(routeID string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, SafeMessage: "command name conflicts with route_id: " + routeID}}
}
