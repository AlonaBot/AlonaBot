// Package options 按组件 Active 配置声明查询内置或私有 Service 提供的选项。
//
// [Service.Query] 从 schema 读取来源和 filter，经 [Options] 注入 Router、Broker、实例快照与聊天查询。
// 配置值的补齐、保存与校验继续由 configschema 负责（CONTRACTS §5.1）。
package options

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/google/jsonschema-go/jsonschema"
)

const optionsSchemaVersion = 1

// Query 指定搜索或存储值；filter 固定取自字段的 options_from。
type Query struct {
	Search string
	Locale string
	Limit  int // 0 使用 OPTIONS_PAGE。
	Values []string
}

// Page 返回选项与是否还有更多结果。
type Page struct {
	Options []Option `json:"options"`
	More    bool     `json:"more"`
}

// Option 提供选项存储值、显示名与说明。
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// SchemaKind 选择组件配置或 Adapter 实例配置的声明。
type SchemaKind int

const (
	// ConfigSchema 选择组件配置声明。
	ConfigSchema SchemaKind = iota
	// AdapterInstanceSchema 选择 Adapter 实例配置声明。
	AdapterInstanceSchema
)

// Options 注入当前代次的声明、路由、实例快照、聊天查询与分页上限。
type Options struct {
	Schema    func(componentID string, kind SchemaKind) (*configschema.Schema, bool)
	Router    *router.Router
	Broker    *broker.Broker
	Instances func(context.Context) ([]AdapterInstance, error)
	Chats     func(ctx context.Context, q state.ChatQuery) ([]state.Chat, bool, error) // nil 时聊天来源不可用。
	Page      int                                                                      // 传入 tune.Values.OptionsPage。
	Agent     func(ctx context.Context) (string, error)                                // 仅 @agent 来源需要，空 ID 表示未安装。
}

// AdapterInstance 标识一个已配置的平台接入实例。
type AdapterInstance struct{ InstanceID, AdapterID string }

// Service 保存查询所需的 Core 依赖，由 New 构造。
type Service struct{ opt Options }

// New 创建按 schema 字段查询选项的入口。
func New(opt Options) *Service { return &Service{opt: opt} }

// PageSize 返回本次 Core 运行中生效的 OPTIONS_PAGE。
func (s *Service) PageSize() int { return s.opt.Page }

// Query 从 Active schema 读取来源；字段未声明来源返回 INVALID_ARGUMENT，来源失败返回 UNAVAILABLE。
func (s *Service) Query(ctx context.Context, componentID string, kind SchemaKind, fieldPath string, q Query) (Page, error) {
	// 1. 来源与 filter 只由配置声明给出，调用方只控制查询条件。
	schema, ok := s.opt.Schema(componentID, kind)
	if !ok {
		return Page{}, unavailable("component schema is unavailable")
	}
	from, ok := schema.OptionsFrom(fieldPath)
	if !ok {
		return Page{}, invalid("field has no options source")
	}
	return s.query(ctx, from, q)
}

// QueryBuiltin 查询登记的内置来源，只接受该来源声明的 filter 键。
func (s *Service) QueryBuiltin(ctx context.Context, source string, filter map[string]string, q Query) (Page, error) {
	keys, ok := configschema.BuiltinOptionSources[source]
	if !ok {
		return Page{}, invalid("unknown builtin options source: " + source)
	}
	for key := range filter {
		if !slices.Contains(keys, key) {
			return Page{}, invalid("unknown options filter: " + key)
		}
	}
	return s.query(ctx, configschema.OptionsFrom{Source: source, Filter: filter}, q)
}

func (s *Service) query(ctx context.Context, from configschema.OptionsFrom, q Query) (Page, error) {
	// 1. 两种入口共用运行时的分页上限。
	if q.Limit < 0 {
		return Page{}, invalid("options limit must be non-negative")
	}
	if q.Limit == 0 || q.Limit > s.opt.Page {
		q.Limit = s.opt.Page
	}
	// 2. 内置来源在进程内读取；组件来源沿用 Broker 的访问和 Session 校验。
	var page Page
	var err error
	if _, builtin := configschema.BuiltinOptionSources[from.Source]; builtin {
		switch from.Source {
		case "core.adapter_instances":
			page, err = s.instances(ctx, from, q)
		case "core.chats":
			page, err = s.chats(ctx, from, q)
		default:
			return Page{}, unavailable("options source is unavailable: " + from.Source)
		}
	} else {
		page, err = s.service(ctx, from, q)
	}
	if err != nil {
		return Page{}, err
	}
	// 3. 按值查询只返回请求的存储值，统一执行分页上限。
	if len(q.Values) > 0 {
		selected := make([]Option, 0, len(page.Options))
		for _, option := range page.Options {
			if slices.Contains(q.Values, option.Value) {
				selected = append(selected, option)
			}
		}
		page.Options = selected
	}
	if len(page.Options) > q.Limit {
		page.Options = page.Options[:q.Limit]
		page.More = true
	}
	return page, nil
}

func (s *Service) instances(ctx context.Context, from configschema.OptionsFrom, q Query) (Page, error) {
	instances, err := s.opt.Instances(ctx)
	if err != nil {
		return Page{}, unavailable(fmt.Sprintf("read adapter instances: %v", err))
	}
	page := Page{Options: []Option{}}
	search := strings.ToLower(q.Search)
	for _, instance := range instances {
		if len(q.Values) > 0 {
			if !slices.Contains(q.Values, instance.InstanceID) {
				continue
			}
		} else {
			if adapterID, filtered := from.Filter["adapter_id"]; filtered && instance.AdapterID != adapterID {
				continue
			}
			if !strings.Contains(strings.ToLower(instance.InstanceID), search) && !strings.Contains(strings.ToLower(instance.AdapterID), search) {
				continue
			}
		}
		page.Options = append(page.Options, Option{Value: instance.InstanceID, Label: instance.InstanceID, Description: instance.AdapterID})
	}
	slices.SortFunc(page.Options, func(a, b Option) int { return cmp.Compare(a.Value, b.Value) })
	return page, nil
}

func (s *Service) service(ctx context.Context, from configschema.OptionsFrom, q Query) (Page, error) {
	componentID, name, _ := strings.Cut(from.Source, "/")
	if componentID == "@agent" {
		if s.opt.Agent == nil {
			return Page{}, unavailable("options source agent is unavailable")
		}
		var err error
		componentID, err = s.opt.Agent(ctx)
		if err != nil {
			return Page{}, fmt.Errorf("resolve options source agent: %w: %w", unavailable("options source agent is unavailable"), err)
		}
		if componentID == "" {
			return Page{}, unavailable("options source agent is unavailable")
		}
	}
	target, ok := s.opt.Router.Caller(componentID)
	if !ok {
		return Page{}, unavailable("options source component is not active: " + componentID)
	}
	plugin, ok := s.opt.Router.PluginFor(target)
	if !ok || !plugin.Activated() {
		return Page{}, unavailable("options source component is not active: " + componentID)
	}
	marked := false
	for _, route := range plugin.Routes() {
		if route.GetKind() == pb.RouteKind_ROUTE_KIND_SERVICE && route.GetService().GetServiceName() == name {
			marked = route.GetService().GetOptionsSource()
			break
		}
	}
	if !marked {
		return Page{}, unavailable("service is not an options source: " + from.Source)
	}
	input := struct {
		Search string            `json:"search"`
		Filter map[string]string `json:"filter,omitempty"`
		Locale string            `json:"locale"`
		Limit  int               `json:"limit"`
		Values []string          `json:"values,omitempty"`
	}{q.Search, from.Filter, q.Locale, q.Limit, q.Values}
	data, err := json.Marshal(input)
	if err != nil {
		return Page{}, unavailable(fmt.Sprintf("encode options query: %v", err))
	}
	response := s.opt.Broker.CallSession(ctx, router.Caller{}, target, &wire.Inbound{Target: componentID + "/" + name, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: validate.OptionsQuerySchemaID, SchemaVersion: optionsSchemaVersion, Bytes: data}})
	if response.Err != nil {
		return Page{}, fmt.Errorf("query options source %s: %w: %w", from.Source, unavailable(response.Err.GetSafeMessage()), &wire.CallError{ErrorInfo: response.Err})
	}
	out := response.Payload
	if out.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || out.SchemaID != validate.OptionsPageSchemaID || out.SchemaVersion != optionsSchemaVersion {
		return Page{}, unavailable("invalid options page encoding or schema")
	}
	// 插件响应属于信任边界，固定类型的 Schema 同时检查必需键和字段类型。
	schema, err := jsonschema.For[Page](nil)
	if err != nil {
		return Page{}, unavailable(fmt.Sprintf("build options page schema: %v", err))
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return Page{}, unavailable(fmt.Sprintf("resolve options page schema: %v", err))
	}
	var value any
	if err := json.Unmarshal(out.Bytes, &value); err != nil {
		return Page{}, unavailable(fmt.Sprintf("decode options page: %v", err))
	}
	if err := resolved.Validate(value); err != nil {
		return Page{}, unavailable(fmt.Sprintf("invalid options page: %v", err))
	}
	var page Page
	if err := json.Unmarshal(out.Bytes, &page); err != nil {
		return Page{}, unavailable(fmt.Sprintf("decode options page: %v", err))
	}
	return page, nil
}

func invalid(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: message}
}
func unavailable(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, Message: message}
}
