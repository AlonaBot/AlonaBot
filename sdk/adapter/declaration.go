package adapter

import (
	"net/http"

	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// ExtensionDeclaration 保存 Extension 从 Go 载荷类型推导的声明。
type ExtensionDeclaration = convert.ExtensionDeclaration

// Extension 声明平台专有片段，T 须为结构体；推导错误在 RunAdapter 注册前报告。
func Extension[T any](name string, version uint32) ExtensionDeclaration {
	return convert.DeclareExtension[T](name, version)
}

// EventDeclaration 保存 Event 从 Go 载荷类型推导的声明。
type EventDeclaration = convert.EventDeclaration

// Event 声明完整名的扩展事件，T 与插件侧 AdapterEvent 使用同一结构体类型。
func Event[T any](name string, version uint32) EventDeclaration {
	return convert.DeclareEvent[T](name, version)
}

// PermissionDeclaration 保存平台权限、隐含权限与展示文案键。
type PermissionDeclaration = convert.PermissionDeclaration

// HTTPAuth 指定 Adapter 路由的认证方式；零值要求 Core 会话。
type HTTPAuth int

// AuthPlatformToken 声明由 Adapter 校验公开路由的平台凭证。
const AuthPlatformToken HTTPAuth = 1

// HTTPRoute 声明 SDK 托管的 HTTP 路由。
type HTTPRoute struct {
	Method  string
	Path    string
	Auth    HTTPAuth
	Handler http.Handler
}

// Descriptor 保存 Adapter 的能力、片段、权限、事件与动作声明。
type Descriptor struct {
	AdapterID    string
	Capabilities []message.Capability
	Extensions   []ExtensionDeclaration
	Permissions  []PermissionDeclaration
	Events       []EventDeclaration
	Actions      []ActionDeclaration
	HTTP         []HTTPRoute
}
