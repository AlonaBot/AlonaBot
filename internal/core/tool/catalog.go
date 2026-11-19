// Package tool 提供工具目录、调用策略及带父调用链的工具执行与用户确认。
//
// [Catalog] 从 Router 的已 Attach 声明取得工具，并在每次查询时读取状态库覆盖。
// 固定服务只列出已激活且调用者可访问的工具；管理查询同时保留未声明工具的覆盖（CONTRACTS §13.1）。
// [Catalog.RegisterInvoke] 装配 Broker 转发与 Core 直接发送的确认消息，确认随父调用失效。
package tool

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const schemaVersion = 1

// Catalog 合并插件工具声明与状态库中的策略覆盖。
type Catalog struct {
	store      *state.Store
	router     *router.Router
	resolve    func(context.Context, *pb.Target) (*profile.Service, string, error) // 装配时设置，每次策略操作建立独立快照。
	mu         sync.Mutex                                                          // 保护本次 Core 运行中出现的操作名；持久化覆盖仍归 store
	operations map[string]ToolPolicyEntry
}

// ToolPolicyEntry 列出当前目录身份、缺省策略与已保存的覆盖。
type ToolPolicyEntry struct {
	PluginID, Service string
	Identities        []string      // 按 tool、operation 排列；仅保存覆盖时为空数组。
	Default           pb.ToolPolicy // UNSPECIFIED 表示当前没有声明与确认记录
	Override          pb.ToolPolicy // UNSPECIFIED 表示没有覆盖
	Effect            pb.ToolEffect
}

// New 创建从 r 读取工具声明、从 store 读取策略的目录。
func New(store *state.Store, r *router.Router) *Catalog {
	return &Catalog{store: store, router: r, operations: make(map[string]ToolPolicyEntry)}
}

// SetProfile 注入按可信默认 Target 选择本次操作配置文件快照的入口；在注册服务前调用。
func (c *Catalog) SetProfile(resolve func(ctx context.Context, target *pb.Target) (profiles *profile.Service, profileID string, err error)) {
	c.resolve = resolve
}

// Register 将 core.tool.list 注册到固定服务分流器。
func (c *Catalog) Register(m *router.Mux) { m.Handle(wire.ServiceToolList, c.list) }

// Effective 优先返回持久化覆盖，否则读取已 Attach 工具的缺省策略。
// 没有覆盖且缺少工具声明时返回 NOT_FOUND；激活与访问检查由调用方负责。
func (c *Catalog) Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error) {
	// 1. 覆盖独立于当前声明与激活状态，读取失败保留原始错误。
	p, err := c.store.GetToolPolicy(ctx, pluginID, service)
	if !errors.Is(err, state.ErrNotFound) {
		return p, err
	}
	// 2. 缺少覆盖时才从当前 Attach 的服务声明取得缺省策略。
	if plugin, ok := c.router.Plugin(pluginID); ok {
		for _, route := range plugin.Routes() {
			spec := route.GetService()
			if route.GetKind() == pb.RouteKind_ROUTE_KIND_SERVICE && spec.GetServiceName() == service && spec.GetTool() != nil {
				return spec.GetTool().GetDefaultPolicy(), nil
			}
		}
	}
	return pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED, policyError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, fmt.Sprintf("tool %s/%s not found", pluginID, service))
}

// SetPolicy 原子保存一组 ALLOW、CONFIRM 或 DENY 覆盖；UNSPECIFIED 删除覆盖。
// 安装记录缺失返回 NOT_FOUND，未知策略返回 INVALID_ARGUMENT；成功返回前已提交事务。
func (c *Catalog) SetPolicy(ctx context.Context, rows ...state.ToolPolicy) error {
	for index, row := range rows {
		switch row.Policy {
		case pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED, pb.ToolPolicy_TOOL_POLICY_ALLOW, pb.ToolPolicy_TOOL_POLICY_CONFIRM, pb.ToolPolicy_TOOL_POLICY_DENY:
		default:
			return policyError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("items[%d]: unknown tool policy %d", index, row.Policy))
		}
	}
	err := c.store.SetToolPolicy(ctx, rows...)
	if errors.Is(err, state.ErrNotFound) {
		return policyError(pb.ErrorCode_ERROR_CODE_NOT_FOUND, err.Error())
	}
	return err
}

// Policies 按完整名字典序列出已 Attach 工具、本次运行出现的操作与保存覆盖的并集。
func (c *Catalog) Policies(ctx context.Context) ([]ToolPolicyEntry, error) {
	// 1. 保存的覆盖先占位，使未 Attach 或已移除的声明仍可被管理。
	entries, err := c.overrides(ctx)
	if err != nil {
		return nil, err
	}
	// 2. 声明补全缺省策略；未激活的 Attach 声明也包含在查询中。
	for _, plugin := range c.router.Plugins() {
		for _, route := range plugin.Routes() {
			spec := route.GetService()
			if route.GetKind() != pb.RouteKind_ROUTE_KIND_SERVICE || spec.GetTool() == nil {
				continue
			}
			entry := entries[plugin.PluginID()+"/"+spec.GetServiceName()]
			entry.PluginID, entry.Service = plugin.PluginID(), spec.GetServiceName()
			entry.Identities = []string{"tool"}
			entry.Default = spec.GetTool().GetDefaultPolicy()
			entry.Effect = spec.GetTool().GetEffect()
			entries[entry.PluginID+"/"+entry.Service] = entry
		}
	}
	// 3. 确认记录提供最新类别，保留声明的缺省策略与保存覆盖。
	c.mu.Lock()
	for key, operation := range c.operations {
		entry := entries[key]
		if entry.Default != pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED {
			operation.Default = entry.Default
		}
		operation.Override = entry.Override
		operation.Identities = append(entry.Identities, "operation")
		entries[key] = operation
	}
	c.mu.Unlock()
	result := make([]ToolPolicyEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}
	slices.SortFunc(result, func(a, b ToolPolicyEntry) int {
		return cmp.Compare(a.PluginID+"/"+a.Service, b.PluginID+"/"+b.Service)
	})
	return result, nil
}

func (c *Catalog) observe(pluginID, operation string, effect pb.ToolEffect) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.operations[pluginID+"/"+operation] = ToolPolicyEntry{PluginID: pluginID, Service: operation, Default: pb.ToolPolicy_TOOL_POLICY_CONFIRM, Effect: effect}
}

func (c *Catalog) overrides(ctx context.Context) (map[string]ToolPolicyEntry, error) {
	rows, err := c.store.ToolPolicies(ctx)
	if err != nil {
		return nil, err
	}
	entries := make(map[string]ToolPolicyEntry, len(rows))
	for _, row := range rows {
		entries[row.PluginID+"/"+row.Service] = ToolPolicyEntry{PluginID: row.PluginID, Service: row.Service, Identities: []string{}, Override: row.Policy}
	}
	return entries, nil
}

func (c *Catalog) list(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
	// 1. 固定服务按声明的 Protobuf Schema 校验请求，并读取本次调用的策略快照。
	request := &pb.ToolListRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != schemaVersion {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "expected ToolListRequest protobuf schema version 1")
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid tool list request protobuf: "+err.Error())
	}
	if in.ParentCallID != 0 {
		if parent, ok := in.ParentAttach.(*router.Invocation); !ok || parent == nil {
			return failure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "tool list requires an active parent call")
		}
	}
	view, err := c.profileSnapshot(ctx, in)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read profile policies: "+err.Error())
	}
	entries, err := c.overrides(ctx)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read tool policies: "+err.Error())
	}
	// 2. 只列出当前激活、访问规则允许且生效策略非 DENY 的工具，包括调用者自身。
	result := &pb.ToolListResponse{}
	for _, plugin := range c.router.Plugins() {
		if !plugin.Activated() {
			continue
		}
		for _, route := range plugin.Routes() {
			spec := route.GetService()
			if route.GetKind() != pb.RouteKind_ROUTE_KIND_SERVICE || spec.GetTool() == nil || !router.ServiceCallable(spec, caller.PluginID, plugin.PluginID()) {
				continue
			}
			name := plugin.PluginID() + "/" + spec.GetServiceName()
			p, err := view.policy(plugin.PluginID(), spec.GetServiceName(), false)
			if err != nil {
				return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "read profile tool policy: "+err.Error())
			}
			if p == pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED {
				p = entries[name].Override
			}
			if p == pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED {
				p = spec.GetTool().GetDefaultPolicy()
			}
			if p == pb.ToolPolicy_TOOL_POLICY_DENY {
				continue
			}
			result.Tools = append(result.Tools, &pb.ToolInfo{Name: name, Description: spec.GetTool().GetDescription(), InputSchemaJson: spec.GetTool().GetInputSchemaJson(), Policy: p, Effect: spec.GetTool().GetEffect()})
		}
	}
	slices.SortFunc(result.Tools, func(a, b *pb.ToolInfo) int { return cmp.Compare(a.GetName(), b.GetName()) })
	encoded, err := proto.Marshal(result)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, "encode tool list response: "+err.Error())
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: schemaVersion, Bytes: encoded}}
}

func failure(code pb.ErrorCode, message string) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}

func policyError(code pb.ErrorCode, message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: code, SafeMessage: message}}
}
