// Package options 注册配置字段的私有 JSON 选项来源，保留固定查询与分页 Schema（CONTRACTS §3、§5.1）。
package options

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"github.com/google/jsonschema-go/jsonschema"
)

// Query 传递搜索条件；Values 非空时按存储值查询，忽略 Search、Filter 与 Locale。
type Query struct {
	Search string            `json:"search"`
	Filter map[string]string `json:"filter,omitempty"`
	Locale string            `json:"locale"`
	Limit  int               `json:"limit"`
	Values []string          `json:"values,omitempty"`
}

// Page 返回选项与是否还有更多结果。
type Page struct {
	Options []Option `json:"options"`
	More    bool     `json:"more"`
}

// Option 提供存储值、显示名与可省略的说明。
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Serve 登记固定 Schema 的私有 JSON 选项来源，Run 收集注册时检查重复名称与 nil Handler。
func Serve(r sdk.Registrar, serviceName string, fn func(ctx context.Context, q Query) (Page, error)) {
	scope := registry.For[*sdk.Context, sdk.Permission](r)
	scope.Check()
	b := &registry.ServiceBuilder[*sdk.Context, sdk.Permission]{Scope: scope, Name: serviceName, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Access: registry.AccessPrivate, Handles: 1}
	b.Prepare = func() (registry.Service[*sdk.Context], error) {
		if fn == nil {
			return registry.Service[*sdk.Context]{}, fmt.Errorf("options handler is nil")
		}
		// 1. 固定类型的 Schema 由 SDK 提供，无需插件生成资源。
		input, err := jsonschema.For[Query](nil)
		if err != nil {
			return registry.Service[*sdk.Context]{}, fmt.Errorf("build options query schema: %w", err)
		}
		output, err := jsonschema.For[Page](nil)
		if err != nil {
			return registry.Service[*sdk.Context]{}, fmt.Errorf("build options page schema: %w", err)
		}
		inputSchema, err := input.Resolve(nil)
		if err != nil {
			return registry.Service[*sdk.Context]{}, err
		}
		outputSchema, err := output.Resolve(nil)
		if err != nil {
			return registry.Service[*sdk.Context]{}, err
		}
		// 2. 复用 Service 调用封装，在用户函数两端校验固定 JSON 契约。
		route := b.Route(validate.OptionsQuerySchemaID, validate.OptionsPageSchemaID, registry.SchemaVersion, registry.SchemaVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, func(c *sdk.Context, data []byte) ([]byte, error) {
			if err := registry.ValidateJSON(inputSchema, data); err != nil {
				return nil, errcode.InvalidArgument.New("invalid options query: " + err.Error())
			}
			var q Query
			if err := json.Unmarshal(data, &q); err != nil {
				return nil, errcode.InvalidArgument.New("decode options query: " + err.Error())
			}
			page, err := fn(c.Context(), q)
			if err != nil {
				return nil, err
			}
			if page.Options == nil {
				page.Options = []Option{} // 空结果按固定 Schema 编码为数组。
			}
			data, err = json.Marshal(page)
			if err != nil {
				return nil, fmt.Errorf("encode options page: %w", err)
			}
			if err := registry.ValidateJSON(outputSchema, data); err != nil {
				return nil, fmt.Errorf("invalid options page: %w", err)
			}
			return data, nil
		})
		route.Descriptor.GetService().OptionsSource = true
		return route, nil
	}
	scope.Services.Builders = append(scope.Services.Builders, b)
}
