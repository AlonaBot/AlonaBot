package sdk

import (
	"context"
	"encoding/json"

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
	config      config.Values
	configReady bool // 准备成功后写入，生命周期 Hook 与 Handler 启动前完成。
}

// Root 列出 SDK 会调用的全部根插件 Hook，供 [Run] 使用；嵌入 [Base] 即满足它。
// 进程生命周期 Hook 见 CONTRACTS §4.1，配置回调 OnConfigLoad 见 §5.2。
type Root interface {
	// OnStart 在配置准备之后、REGISTER 之前执行，不能发起业务调用；失败时本次启动失败。
	OnStart(context.Context) error
	// OnReady 在收到 ACTIVATE 之后、预算内执行；成功后才写出 ACTIVATED。
	OnReady(context.Context) error
	// OnStop 在 STOPPING 或断链后执行，最多使用 STOPPING 给出的期限关闭资源。
	OnStop(context.Context) error
	// OnConfigLoad 接收声明式迁移后的完整 JSON 和原始版本号；只做纯配置变换，此时 Config 尚未就绪。
	OnConfigLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
	// Commands 在 Collect 时调用一次，注册命令。
	Commands(*CommandRegistry)
	// Events 在 Collect 时调用一次，注册事件路由。
	Events(*EventRegistry)
	// Services 在 Collect 时调用一次，注册服务。
	Services(*ServiceRegistry)
	// Subscriptions 在 Collect 时调用一次，注册 Event Bus 订阅。
	Subscriptions(*SubscriptionRegistry)
	// HTTP 在 Collect 时调用一次，注册 HTTP 路由。
	HTTP(*HTTPRegistry)
	// Schedules 在 Collect 时调用一次，注册定时任务。
	Schedules(*ScheduleRegistry)

	setConfig(config.Values) // 只由 Base 实现，注入本 Session 的配置并约束根插件嵌入 Base。
}

// OnStart 默认什么也不做。
func (*Base) OnStart(context.Context) error { return nil }

// OnReady 默认什么也不做。
func (*Base) OnReady(context.Context) error { return nil }

// OnStop 默认什么也不做。
func (*Base) OnStop(context.Context) error { return nil }

// OnConfigLoad 默认原样返回 raw。
func (*Base) OnConfigLoad(_ context.Context, _ uint32, raw json.RawMessage) (json.RawMessage, error) {
	return raw, nil
}

// Config 返回本 Session 的只读配置；配置准备完成前调用会 panic。
func (b *Base) Config() config.Values {
	if !b.configReady {
		panic("config is not prepared")
	}
	return b.config
}

// Commands 默认不注册命令。
func (*Base) Commands(*CommandRegistry) {}

// Events 默认不注册事件路由。
func (*Base) Events(*EventRegistry) {}

// Services 默认不注册服务。
func (*Base) Services(*ServiceRegistry) {}

// Subscriptions 默认不订阅 Event Bus。
func (*Base) Subscriptions(*SubscriptionRegistry) {}

// HTTP 默认不注册 HTTP 路由。
func (*Base) HTTP(*HTTPRegistry) {}

// Schedules 默认不注册定时任务。
func (*Base) Schedules(*ScheduleRegistry) {}

func (b *Base) setConfig(values config.Values) {
	b.config = values
	b.configReady = true
}
