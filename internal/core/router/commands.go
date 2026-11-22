package router

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// 命令菜单变化从首次变化起合并 2 秒。
const commandSyncWindow = 2 * time.Second
const commandSyncCapability = "commands.sync"

// commandSync 由一个 worker 串行发布菜单；due 由 Router.mu 保护。
// wake 只提示重新读取 due，连续通知合并而不移动首次变化的截止时间。
type commandSync struct {
	instance   adapter.Instance
	ctx        context.Context
	cancel     context.CancelFunc
	wake       chan struct{}
	done       chan struct{}
	operations chan *menuOperation
	controller adapter.CommandMenuController
	account    state.CommandMenuAccount // 其余状态仅由 worker 读写。
	chats      map[string]state.CommandMenuChat
	published  map[string]*pb.CommandMenu // 本 Session 已发布且位置记账成功的自动菜单。
	notified   map[menuReadyKey]bool
	pending    map[string]bool // true 提示成员关系发生变化，仅触发处理与缓存失效，由 Router.mu 保护。
	observed   map[string]bool // 已见聊天，由 Router.mu 保护。
	due        time.Time
}

// StartCommandSync 为声明支持菜单同步的运行实例发布一次菜单，并管理后续变化。
func (r *Router) StartCommandSync(in adapter.Instance) {
	if !slices.ContainsFunc(in.Capabilities(), func(capability *pb.Capability) bool {
		return capability.GetName() == commandSyncCapability && capability.GetSupported()
	}) {
		return
	}
	ctx, cancel := context.WithCancel(r.ctx)
	next := &commandSync{instance: in, ctx: ctx, cancel: cancel, wake: make(chan struct{}, 1), done: make(chan struct{}), operations: make(chan *menuOperation), chats: make(map[string]state.CommandMenuChat), published: make(map[string]*pb.CommandMenu), notified: make(map[menuReadyKey]bool), pending: make(map[string]bool), observed: make(map[string]bool)}
	if controller, ok := in.(adapter.CommandMenuController); ok && controller.CommandMenuConfig() != nil {
		next.controller = controller
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		cancel()
		return
	}
	previous := r.commandSyncs[in.InstanceID()]
	if previous != nil {
		if previous.instance == in && previous.ctx.Err() == nil {
			r.mu.Unlock()
			cancel()
			return
		}
		previous.cancel()
	}
	r.commandSyncs[in.InstanceID()] = next
	r.mu.Unlock()
	go func() {
		// 实例替换也等待旧 worker 排空，避免跨 Session 的模式切换与写入重叠。
		if previous != nil {
			<-previous.done
		}
		r.runCommandSync(next)
	}()
}

// StopCommandSync 撤销本次运行的同步任务；旧实例的通知保留新实例的任务。
func (r *Router) StopCommandSync(in adapter.Instance) {
	r.mu.Lock()
	state := r.commandSyncs[in.InstanceID()]
	if state != nil && state.instance == in {
		state.cancel()
	}
	r.mu.Unlock()
}

// CommandsChanged 在路由激活、撤销或命令策略变化后重新计算运行实例的菜单。
func (r *Router) CommandsChanged() {
	defer r.scheduleCommandSync()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.rebuildCommands()
}

// scheduleCommandSync 在注册更新解锁后通知菜单任务，沿用首次变化的截止时间。
func (r *Router) scheduleCommandSync() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	for _, state := range r.commandSyncs {
		if state.due.IsZero() {
			state.due = time.Now().Add(commandSyncWindow)
		}
		select {
		case state.wake <- struct{}{}:
		default:
		} // 已有提示时保留原截止时间。
	}
}

func (r *Router) runCommandSync(state *commandSync) {
	defer close(state.done)
	if state.controller != nil {
		r.runManagedMenus(state)
		return
	}
	r.syncCommands(state)
	for {
		r.mu.Lock()
		due := state.due
		r.mu.Unlock()
		if due.IsZero() {
			select {
			case <-state.ctx.Done():
				return
			case <-state.wake:
			}
			continue
		}
		timer := time.NewTimer(time.Until(due))
		select {
		case <-state.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		r.mu.Lock()
		state.due = time.Time{}
		r.mu.Unlock()
		r.syncCommands(state)
	}
}

func (r *Router) syncCommands(state *commandSync) {
	if state.ctx.Err() != nil {
		return
	}
	commands, err := r.commandMenu(state.ctx, state.instance)
	if err != nil {
		slog.ErrorContext(state.ctx, "compute adapter command menu failed", "adapter_instance_id", state.instance.InstanceID(), "error", err)
		return
	}
	if state.ctx.Err() != nil {
		return
	}
	synchronizer, ok := state.instance.(adapter.CommandSynchronizer)
	var failure *pb.ErrorInfo
	if !ok {
		failure = &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter instance does not support command synchronization"}
	} else {
		failure = synchronizer.SyncCommands(state.ctx, commands)
	}
	if failure != nil {
		slog.ErrorContext(state.ctx, "adapter command synchronization failed", "adapter_instance_id", state.instance.InstanceID(), "error", failure)
	}
}

func (r *Router) commandMenu(ctx context.Context, instance adapter.Instance) ([]*pb.CommandMenuItem, error) {
	target := &pb.Target{AdapterId: instance.AdapterID(), AdapterInstanceId: instance.InstanceID(), AccountId: instance.AccountID()}
	return r.commandMenuForTarget(ctx, instance, target)
}

func (r *Router) commandMenuForTarget(ctx context.Context, instance adapter.Instance, target *pb.Target) ([]*pb.CommandMenuItem, error) {
	prefixes, err := r.prefixes(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("read command menu prefixes for %s: %w", instance.InstanceID(), err)
	}
	if !slices.Contains(prefixes, "/") {
		return nil, nil
	}
	// 1. 路由和胜出索引取同一快照；策略读取与插件访问均在状态锁外执行。
	r.mu.Lock()
	index := r.commands
	var routes []commandRoute
	for _, cached := range r.matches {
		routes = append(routes, cached...)
	}
	r.mu.Unlock()
	var menu []*pb.CommandMenuItem
	for _, route := range routes {
		command := route.route.GetCommand()
		if !command.GetMenu() || command.GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH || !r.available(route.candidate) {
			continue
		}
		names := index.names[route.pluginID][route.route.GetRouteId()]
		if len(names) == 0 || len(names[0]) != 1 {
			continue
		}
		path := names[0]
		winner, _, ok := index.Lookup(path)
		if !ok || winner.Alias || winner.PluginID != route.pluginID || winner.RouteID != route.route.GetRouteId() {
			continue
		}
		if r.menuStore != nil {
			allowed, err := r.menuStore.AdapterInstanceAllowed(ctx, route.pluginID, instance.InstanceID())
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
		}
		_, profileEnabled, err := r.Profile(ctx, route.pluginID, target)
		if err != nil {
			return nil, err
		}
		if !profileEnabled {
			continue
		}
		// 2. 使用完整生效权限集合，父层禁用的胜出者仍占据路径。
		enabled, permissions := route.route.GetDefaultEnabled(), route.route.GetRequiredPermissions()
		if r.commandState != nil {
			var err error
			enabled, permissions, err = r.commandState(ctx, route.pluginID, route.route)
			if err != nil {
				return nil, fmt.Errorf("read command state for %s/%s: %w", route.pluginID, route.route.GetRouteId(), err)
			}
		}
		if !enabled || len(permissions) != 0 && !slices.Contains(permissions, "user") {
			continue
		}
		if r.policy != nil {
			actor := &pb.Actor{AdapterInstanceId: instance.InstanceID(), Permissions: []string{"user"}}
			if failure := r.policy.CheckCommand(ctx, route.pluginID, route.route, actor, target); failure != nil {
				if failure.Code == pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || failure.Code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
					continue
				}
				return nil, &wire.CallError{ErrorInfo: failure}
			}
		}
		// 3. 英文来自代码，目录中的每个语言逐键回落；未登记说明时使用命令名。
		descriptions := CommandDescriptions(route.messages(), route.defaultLocale(), command.GetPath(), command.GetDescription())
		if descriptions == nil {
			descriptions = map[string]string{"en-US": path[0]}
			if catalog := route.messages(); catalog != nil {
				for _, locale := range catalog.Locales() {
					descriptions[locale] = path[0]
				}
			}
		}
		menu = append(menu, &pb.CommandMenuItem{Name: path[0], Descriptions: descriptions})
	}
	slices.SortFunc(menu, func(a, b *pb.CommandMenuItem) int { return cmp.Compare(a.Name, b.Name) })
	return menu, nil
}

// CommandDescriptions 按包内语言及 Session 默认语言选取说明；未登记英文说明时返回 nil。
func CommandDescriptions(catalog *i18n.Catalog, defaultLocale string, path []string, source string) map[string]string {
	if source == "" {
		return nil
	}
	descriptions := map[string]string{"en-US": source}
	if catalog != nil {
		for _, locale := range catalog.Locales() {
			descriptions[locale] = catalog.CommandDescription([]string{locale, defaultLocale}, path, source)
		}
	}
	return descriptions
}
