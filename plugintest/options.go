package plugintest

import (
	"slices"
	"testing"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

// Option 配置一次插件测试套件启动。
type Option func(*options)

type options struct {
	t            testing.TB
	plugins      []*sdk.Plugin
	configs      map[*sdk.Plugin]string
	admins       []string
	allowed      map[*sdk.Plugin][]string
	capabilities []string
	descriptor   adapter.Descriptor
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
