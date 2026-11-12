package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/google/jsonschema-go/jsonschema"
)

// OptionsQuery 传递搜索条件；Values 非空时按存储值查询，忽略 Search、Filter 与 Locale。
type OptionsQuery struct {
	Search string            `json:"search"`
	Filter map[string]string `json:"filter,omitempty"`
	Locale string            `json:"locale"`
	Limit  int               `json:"limit"`
	Values []string          `json:"values,omitempty"`
}

// OptionsPage 返回选项与是否还有更多结果。
type OptionsPage struct {
	Options []Option `json:"options"`
	More    bool     `json:"more"`
}

// Option 提供存储值、显示名与可省略的说明。
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// ServeOptions 登记固定 Schema 的私有 JSON 选项来源，Collect 检查重复名称与 nil Handler。
func ServeOptions(r Registrar, serviceName string, fn func(ctx context.Context, q OptionsQuery) (OptionsPage, error)) {
	scope := r.registrar()
	scope.plugin.checkRegistration()
	b := &serviceBuilder{plugin: scope.plugin, subID: scope.subID, name: serviceName, access: AccessPrivate, handles: 1}
	b.prepare = func() (serviceRoute, error) {
		if fn == nil {
			return serviceRoute{}, fmt.Errorf("options handler is nil")
		}
		// 1. 固定类型的 Schema 由 SDK 提供，无需插件生成资源。
		input, err := jsonschema.For[OptionsQuery](nil)
		if err != nil {
			return serviceRoute{}, fmt.Errorf("build options query schema: %w", err)
		}
		output, err := jsonschema.For[OptionsPage](nil)
		if err != nil {
			return serviceRoute{}, fmt.Errorf("build options page schema: %w", err)
		}
		inputSchema, err := input.Resolve(nil)
		if err != nil {
			return serviceRoute{}, err
		}
		outputSchema, err := output.Resolve(nil)
		if err != nil {
			return serviceRoute{}, err
		}
		// 2. 复用 Service 调用封装，在用户函数两端校验固定 JSON 契约。
		route := b.route(validate.OptionsQuerySchemaID, validate.OptionsPageSchemaID, invocationSchemaVersion, invocationSchemaVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, func(c *Context, data []byte) ([]byte, error) {
			if err := validateServiceJSON(inputSchema, data); err != nil {
				return nil, errcode.InvalidArgument.New("invalid options query: " + err.Error())
			}
			var q OptionsQuery
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
			if err := validateServiceJSON(outputSchema, data); err != nil {
				return nil, fmt.Errorf("invalid options page: %w", err)
			}
			return data, nil
		})
		route.descriptor.GetService().OptionsSource = true
		return route, nil
	}
	scope.plugin.services.builders = append(scope.plugin.services.builders, b)
}
