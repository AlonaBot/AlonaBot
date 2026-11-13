package router

import (
	"cmp"
	"log/slog"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// CommandEntry 保存一条规范路径的命令候选。
type CommandEntry struct {
	PluginID string
	RouteID  string
	Priority int32
	Path     []string   // 命中的生效完整路径
	Alias    bool       // Path 不是主名
	Names    [][]string // 生效的全部命令名，主名在前；只读
}

// InvalidName 记录覆盖因插件当前声明而失效的命令名。
type InvalidName struct {
	PluginID, RouteID string
	Name              []string // 整条覆盖指向未注册或非路径命令时为 nil
	Reason            string   // "declared" 表示声明名冲突；"unregistered" 表示路由不可用
}

const (
	invalidNameDeclared     = "declared"
	invalidNameUnregistered = "unregistered"
)

// Shadow 保存同一路径的胜出者及按胜出顺序排列的被遮蔽者。
type Shadow struct {
	Path     []string
	Winner   CommandEntry
	Shadowed []CommandEntry
}

// CommandIndex 保存不可变的静态命令快照；零值为空索引。
// 路径在建立索引时复制；返回的 Path 及 Shadows 切片内容只读。
type CommandIndex struct {
	paths        map[string][]CommandEntry        // 候选按 priority、plugin_id、route_id 升序排列
	names        map[string]map[string][][]string // 每条路径命令的生效名字，供菜单与匹配共用
	invalidNames []InvalidName
}

// InvalidNames 按 plugin_id、route_id 与名字返回失效覆盖诊断；返回的切片只读。
func (x *CommandIndex) InvalidNames() []InvalidName { return x.invalidNames }

// Lookup 按最长完整 token 路径查找命令，大小写敏感；consumed 为匹配的 token 数。
// 返回的 entry.Path 只读。
func (x *CommandIndex) Lookup(tokens []string) (entry CommandEntry, consumed int, ok bool) {
	for n := len(tokens); n > 0; n-- {
		candidates := x.paths[strings.Join(tokens[:n], " ")]
		if len(candidates) > 0 && slices.Equal(candidates[0].Path, tokens[:n]) {
			return candidates[0], n, true
		}
	}
	return CommandEntry{}, 0, false
}

// Shadows 按规范路径字典序返回遮蔽关系；返回的 Path 和 Shadowed 切片只读。
func (x *CommandIndex) Shadows() []Shadow {
	var shadows []Shadow
	for _, entries := range x.paths {
		if len(entries) < 2 {
			continue
		}
		shadow := Shadow{Path: entries[0].Path, Winner: entries[0], Shadowed: entries[1:]}
		shadows = append(shadows, shadow)
	}
	slices.SortFunc(shadows, func(a, b Shadow) int { return strings.Compare(strings.Join(a.Path, " "), strings.Join(b.Path, " ")) })
	return shadows
}

// Commands 返回最近一次注册或策略变化建立的索引快照，可在后续重建期间继续读取。
func (r *Router) Commands() *CommandIndex {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.commands
}

// rebuildCommands 在持有 mu 时建立并发布快照，沿用注册边界已经校验过的路由。
func (r *Router) rebuildCommands() {
	// 1. 只收集已激活插件；声明名始终优先于同插件其他路由的覆盖。
	next := &CommandIndex{paths: make(map[string][]CommandEntry), names: make(map[string]map[string][][]string)}
	for _, session := range r.plugins {
		plugin := session.plugin
		if !plugin.Activated() {
			continue
		}
		pluginID := plugin.PluginID()
		routes := plugin.Routes()
		declared := make(map[string]string)
		registered := make(map[string]bool)
		for _, route := range routes {
			if route.GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND || route.GetCommand().GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH {
				continue
			}
			registered[route.GetRouteId()] = true
			for _, name := range declaredCommandNames(route.GetCommand()) {
				declared[strings.Join(name, " ")] = route.GetRouteId()
			}
		}
		var overrides map[string][][]string
		if r.commandNames != nil {
			overrides = r.commandNames(pluginID)
		}
		for routeID := range overrides {
			if !registered[routeID] {
				next.invalidNames = append(next.invalidNames, InvalidName{PluginID: pluginID, RouteID: routeID, Reason: invalidNameUnregistered})
			}
		}
		next.names[pluginID] = make(map[string][][]string)
		for _, route := range routes {
			routeID := route.GetRouteId()
			if !registered[routeID] {
				continue
			}
			var names [][]string
			for _, name := range overrides[routeID] {
				if owner, ok := declared[strings.Join(name, " ")]; ok && owner != routeID {
					next.invalidNames = append(next.invalidNames, InvalidName{PluginID: pluginID, RouteID: routeID, Name: slices.Clone(name), Reason: invalidNameDeclared})
					continue
				}
				names = append(names, slices.Clone(name))
			}
			if len(names) == 0 {
				names = declaredCommandNames(route.GetCommand())
			}
			next.names[pluginID][routeID] = names
			entry := CommandEntry{PluginID: pluginID, RouteID: routeID, Priority: route.GetPriority(), Names: names}
			for i, name := range names {
				entry.Alias = i != 0
				next.add(entry, name)
			}
		}
	}
	slices.SortFunc(next.invalidNames, func(a, b InvalidName) int {
		if order := cmp.Compare(a.PluginID, b.PluginID); order != 0 {
			return order
		}
		if order := cmp.Compare(a.RouteID, b.RouteID); order != 0 {
			return order
		}
		return slices.Compare(a.Name, b.Name)
	})
	for _, entries := range next.paths {
		slices.SortFunc(entries, func(a, b CommandEntry) int {
			if order := cmp.Compare(a.Priority, b.Priority); order != 0 {
				return order
			}
			if order := cmp.Compare(a.PluginID, b.PluginID); order != 0 {
				return order
			}
			return cmp.Compare(a.RouteID, b.RouteID)
		})
	}
	// 2. 同一胜出者与被遮蔽者的关系只在新出现时记录，胜出者改变则产生新关系。
	for _, shadow := range next.Shadows() {
		previous := r.commands.paths[strings.Join(shadow.Path, " ")]
		for _, entry := range shadow.Shadowed {
			if len(previous) > 1 && sameRoute(previous[0], shadow.Winner) && slices.ContainsFunc(previous[1:], func(old CommandEntry) bool { return sameRoute(old, entry) }) {
				continue
			}
			slog.Info("command route shadowed", "path", strings.Join(shadow.Path, " "), "winner_plugin_id", shadow.Winner.PluginID, "winner_route_id", shadow.Winner.RouteID, "shadowed_plugin_id", entry.PluginID, "shadowed_route_id", entry.RouteID)
		}
	}
	r.commands = next
}

func (x *CommandIndex) add(entry CommandEntry, path []string) {
	entry.Path = slices.Clone(path)
	key := strings.Join(path, " ")
	x.paths[key] = append(x.paths[key], entry)
}

func sameRoute(a, b CommandEntry) bool { return a.PluginID == b.PluginID && a.RouteID == b.RouteID }

func declaredCommandNames(command *pb.CommandSpec) [][]string {
	names := [][]string{slices.Clone(command.GetPath())}
	for _, alias := range command.GetAliases() {
		names = append(names, slices.Clone(alias.GetPath()))
	}
	return names
}
