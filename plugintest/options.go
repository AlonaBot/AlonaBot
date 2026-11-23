package plugintest

import (
	"context"
	"slices"
	"testing"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
)

// Option 配置一次插件测试套件启动。
type Option func(*options)

type options struct {
	clock          *Clock
	instances      []InstanceConfig
	defaults       map[string]string
	behavior       Behavior
	t              testing.TB
	plugins        []*sdk.Plugin
	configs        map[*sdk.Plugin]string
	admins         []string
	platformAdmins []string
	allowed        map[*sdk.Plugin][]string
	capabilities   []string
	descriptor     adapter.Descriptor
	llmHandler     func(context.Context, llm.Request) (llm.Response, error)
}

// LLM 登记经真实 Agent Service 调用的生成替身；多次设置时最后一个生效。
// handler 为 nil 时 Start 在启动组件前使测试失败。
func LLM(handler func(ctx context.Context, req llm.Request) (llm.Response, error)) Option {
	return func(o *options) {
		if handler == nil {
			o.t.Fatal("LLM handler must not be nil")
		}
		o.llmHandler = handler
	}
}

// Plugin 登记被测插件，可多次使用。
func Plugin(p *sdk.Plugin) Option {
	return func(o *options) { o.plugins = append(o.plugins, p) }
}

// Config 写入插件开发目录中的 config.json。
func Config(p *sdk.Plugin, configJSON string) Option {
	return func(o *options) { o.configs[p] = configJSON }
}

// Admin 将用户登记为当前测试 Adapter 的 Core Admin。
func Admin(userID string) Option {
	return func(o *options) { o.admins = append(o.admins, userID) }
}

// AllowedInstances 收紧插件的发信许可；空列表拒绝全部实例。
func AllowedInstances(p *sdk.Plugin, instanceIDs ...string) Option {
	ids := slices.Clone(instanceIDs)
	return func(o *options) { o.allowed[p] = ids }
}

// Capabilities 替换测试实例的缺省能力。
func Capabilities(names ...string) Option {
	names = slices.Clone(names)
	return func(o *options) { o.capabilities = names }
}

// Adapter 设置 Adapter ID 与平台声明；其他声明类型使测试失败。
func Adapter(id string, declarations ...any) Option {
	return func(o *options) {
		o.descriptor = adapter.Descriptor{AdapterID: id}
		for _, declaration := range declarations {
			switch d := declaration.(type) {
			case adapter.PermissionDeclaration:
				o.descriptor.Permissions = append(o.descriptor.Permissions, d)
			case adapter.EventDeclaration:
				o.descriptor.Events = append(o.descriptor.Events, d)
			case adapter.ExtensionDeclaration:
				o.descriptor.Extensions = append(o.descriptor.Extensions, d)
			case adapter.ActionDeclaration:
				o.descriptor.Actions = append(o.descriptor.Actions, d)
			default:
				o.t.Fatalf("unsupported Adapter declaration %T", declaration)
			}
		}
	}
}

// AdminOn 将用户登记为指定平台的 Core Admin。
func AdminOn(adapterID, userID string) Option {
	return func(o *options) { o.platformAdmins = append(o.platformAdmins, adapterID+":"+userID) }
}
