package plugintest

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Action 执行声明过的平台动作，完整保留 JSON 结果与安全错误。
func (a *testInstance) Action(ctx context.Context, to, name string, version uint32, params json.RawMessage) (json.RawMessage, *errcode.Error) {
	call := a.call(ctx, "action", to)
	call.Name, call.Version, call.Params = name, version, slices.Clone(params)
	var result json.RawMessage
	safe := errcode.Unsupported.New("test action behavior is not configured")
	if a.config.Behavior.Action != nil {
		result, safe = a.config.Behavior.Action(ctx, call)
	}
	call.Payload, call.Err = result, safe
	a.h.recordCall(call)
	return result, safe
}

// Delete 通过 SDK 撤回入口记录每个物理消息的实际结果。
func (a *testInstance) Delete(ctx context.Context, to string, ids []string) (*message.DeleteResult, *errcode.Error) {
	call := a.call(ctx, "message.delete", to)
	call.MessageIDs = slices.Clone(ids)
	result := &message.DeleteResult{Session: to, AdapterInstanceID: a.config.InstanceID}
	for _, id := range ids {
		result.Items = append(result.Items, message.DeleteItemResult{PlatformMessageID: id, Status: message.PartAccepted})
	}
	call.DeleteResult = result
	var safe *errcode.Error
	if a.config.Behavior.Delete != nil {
		result, safe = a.config.Behavior.Delete(ctx, call)
	}
	call.DeleteResult, call.Err = result, safe
	a.h.recordCall(call)
	return result, safe
}

// ReadNodes 记录完整的节点查询并执行可配置结果。
func (a *testInstance) ReadNodes(ctx context.Context, to string, segment message.ForwardSegment) ([]message.ForwardItem, *errcode.Error) {
	call := a.call(ctx, "message.read_nodes", to)
	call.Chain = message.Chain{segment}
	nodes := segment.Items
	var safe *errcode.Error
	if a.config.Behavior.Nodes != nil {
		nodes, safe = a.config.Behavior.Nodes(ctx, call)
	}
	call.Nodes, call.Err = nodes, safe
	a.h.recordCall(call)
	return nodes, safe
}

// ReadPlatformFile 按真实媒体 RPC 从本实例 fixture 读取分块。
func (a *testInstance) ReadPlatformFile(ctx context.Context, id string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	call := a.call(ctx, "media.read", "")
	call.Name, call.Offset, call.Length = id, offset, length
	a.h.mu.Lock()
	fixture, ok := a.h.fixtures[a.config.InstanceID][id]
	a.h.mu.Unlock()
	if !ok {
		call.Err = errcode.NotFound.New("platform fixture not found")
		a.h.recordCall(call)
		return nil, 0, false, call.Err
	}
	if offset > uint64(len(fixture.Data)) {
		call.Err = errcode.InvalidArgument.New("media offset exceeds fixture size")
		a.h.recordCall(call)
		return nil, 0, false, call.Err
	}
	next := min(offset+uint64(length), uint64(len(fixture.Data)))
	data := slices.Clone(fixture.Data[offset:next])
	eof := next == uint64(len(fixture.Data))
	call.Data, call.NextOffset, call.EOF = data, next, eof
	a.h.recordCall(call)
	return data, next, eof, nil
}

// SyncCommands 记录传统命令菜单同步。
func (a *testInstance) SyncCommands(ctx context.Context, commands []adapter.CommandMenuItem) *errcode.Error {
	call := a.call(ctx, "commands.sync", "")
	call.Params, _ = json.Marshal(commands)
	var safe *errcode.Error
	if a.config.Behavior.Menu != nil {
		_, safe = a.config.Behavior.Menu(ctx, call)
	}
	call.Err = safe
	a.h.recordCall(call)
	return safe
}

// CommandMenuConfig 返回该测试实例配置的菜单模式。
func (a *testInstance) CommandMenuConfig() adapter.CommandMenuConfig { return a.config.MenuConfig }

type menuLocation struct {
	scope    adapter.CommandMenuScope
	language string
}

func (a *testInstance) menu(ctx context.Context, operation string, menu adapter.CommandMenu) ([]adapter.CommandMenuEntry, *errcode.Error) {
	call := a.call(ctx, "command.menu."+operation, "")
	call.Menu = menu
	var entries []adapter.CommandMenuEntry
	var safe *errcode.Error
	if a.config.Behavior.Menu != nil {
		entries, safe = a.config.Behavior.Menu(ctx, call)
	} else {
		location := menuLocation{menu.Scope, menu.LanguageCode}
		a.h.mu.Lock()
		switch operation {
		case "set":
			a.menus[location] = slices.Clone(menu.Commands)
			entries = slices.Clone(menu.Commands)
		case "get":
			entries = slices.Clone(a.menus[location])
		case "delete":
			delete(a.menus, location)
		}
		a.h.mu.Unlock()
	}

	call.Payload, _ = json.Marshal(entries)
	call.Err = safe
	a.h.recordCall(call)
	return entries, safe
}

// SetCommandMenu 通过 SDK 菜单串行化入口替换指定位置。
func (a *testInstance) SetCommandMenu(ctx context.Context, menu adapter.CommandMenu) *errcode.Error {
	_, safe := a.menu(ctx, "set", menu)
	return safe
}

// GetCommandMenu 查询指定位置，保留结果与错误。
func (a *testInstance) GetCommandMenu(ctx context.Context, scope adapter.CommandMenuScope, language string) ([]adapter.CommandMenuEntry, *errcode.Error) {
	return a.menu(ctx, "get", adapter.CommandMenu{Scope: scope, LanguageCode: language})
}

// DeleteCommandMenu 撤销指定位置。
func (a *testInstance) DeleteCommandMenu(ctx context.Context, scope adapter.CommandMenuScope, language string) *errcode.Error {
	_, safe := a.menu(ctx, "delete", adapter.CommandMenu{Scope: scope, LanguageCode: language})
	return safe
}
