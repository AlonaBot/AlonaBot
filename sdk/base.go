package sdk

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/config"
)

// Base 由根插件以值的形式嵌入：
//
//	type RSS struct{ sdk.Base }
//
// Base 为 [Root] 的全部 Hook 提供空的默认实现，根插件只覆盖需要的方法；SDK 原样调用，不区分是否覆盖。
// 覆盖时签名必须与默认实现一致，否则根插件不再满足 Root，调用 [Run] 处编译失败。根插件也不能有与
// Hook 同名的字段（例如 HTTP），字段会遮住默认实现，同样导致编译失败。
type Base struct {
	_ struct{} // 保留给 SDK 注入的运行时状态，禁止插件依赖其字段
}

// Root 是传给 [Run] 的根插件，列出 SDK 会调用的全部 Hook；嵌入 [Base] 即满足它。
// 进程生命周期 Hook 见 CONTRACTS §4.1，配置回调 OnLoad 由配置类型实现（见 sdk/config）。
type Root interface {
	// OnStart 在配置准备之后、REGISTER 之前执行，不能发起业务调用；失败时本次启动失败。
	OnStart(context.Context) error
	// OnReady 在收到 ACTIVATE 之后、预算内执行；成功后才写出 ACTIVATED。
	OnReady(context.Context) error
	// OnStop 在 STOPPING 或断链后执行，最多使用 STOPPING 给出的期限关闭资源。
	OnStop(context.Context) error
	// Config 返回供 SDK 解码配置的结构体指针（嵌入 config.Base）；没有配置时返回 nil。每个插件包至多一个 Config。
	Config() config.Config
	// Commands 在 Collect 时调用一次，注册命令。
	Commands(*CommandRegistry)
	// Events 在 Collect 时调用一次，注册事件路由。
	Events(*EventRegistry)
	// Services 在 Collect 时调用一次，注册服务。
	Services(*ServiceRegistry)
	// HTTP 在 Collect 时调用一次，注册 HTTP 路由。
	HTTP(*HTTPRegistry)
	// Schedules 在 Collect 时调用一次，注册定时任务。
	Schedules(*ScheduleRegistry)

	alonabotRoot() // 标记方法，只由 Base 实现
}

// OnStart 默认什么也不做。
func (*Base) OnStart(context.Context) error { return nil }

// OnReady 默认什么也不做。
func (*Base) OnReady(context.Context) error { return nil }

// OnStop 默认什么也不做。
func (*Base) OnStop(context.Context) error { return nil }

// Config 默认返回 nil，表示插件没有配置。
func (*Base) Config() config.Config { return nil }

// Commands 默认不注册命令。
func (*Base) Commands(*CommandRegistry) {}

// Events 默认不注册事件路由。
func (*Base) Events(*EventRegistry) {}

// Services 默认不注册服务。
func (*Base) Services(*ServiceRegistry) {}

// HTTP 默认不注册 HTTP 路由。
func (*Base) HTTP(*HTTPRegistry) {}

// Schedules 默认不注册定时任务。
func (*Base) Schedules(*ScheduleRegistry) {}

// alonabotRoot 使嵌入 Base 的类型满足 Root。
func (*Base) alonabotRoot() {}
