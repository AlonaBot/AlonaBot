package router

import (
	"cmp"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// CommandEntry 保存一条规范路径的命令候选。
type CommandEntry struct {
	PluginID string
	RouteID  string
	Priority int32
	Path     []string // 命中的完整路径，来自 path 或某个 alias
	Alias    bool     // Path 来自 alias
}

// Shadow 保存同一路径的胜出者及按胜出顺序排列的被遮蔽者。
type Shadow struct {
	Path     []string
	Winner   CommandEntry
	Shadowed []CommandEntry
}

// CommandIndex 保存不可变的静态命令快照；零值为空索引。
// 路径在建立索引时复制；返回的 Path 及 Shadows 切片内容只读。
type CommandIndex struct {
	paths map[string][]CommandEntry // 候选按 priority、plugin_id、route_id 升序排列
}

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

// Commands 返回最近一次 Attach/Detach 建立的索引快照，可在后续重建期间继续读取。
func (r *Router) Commands() *CommandIndex {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.commands
}

// rebuildCommands 在持有 mu 时建立并发布快照，沿用注册边界已经校验过的路由。
func (r *Router) rebuildCommands() {
	// 1. 只收集已激活插件，命令 path 与各 alias 分别参与选择。
	next := &CommandIndex{paths: make(map[string][]CommandEntry)}
	for _, session := range r.plugins {
		plugin := session.plugin
		if !plugin.Activated() {
			continue
		}
		for _, route := range plugin.Routes() {
			if route.GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND {
				continue
			}
			command := route.GetCommand()
			entry := CommandEntry{PluginID: plugin.PluginID(), RouteID: route.GetRouteId(), Priority: route.GetPriority()}
			if len(command.GetPath()) > 0 {
				next.add(entry, command.GetPath())
			}
			entry.Alias = true
			for _, alias := range command.GetAliases() {
				next.add(entry, alias.GetPath())
			}
		}
	}
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
			r.log.Info("command route shadowed", "path", strings.Join(shadow.Path, " "), "winner_plugin_id", shadow.Winner.PluginID, "winner_route_id", shadow.Winner.RouteID, "shadowed_plugin_id", entry.PluginID, "shadowed_route_id", entry.RouteID)
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
