// Package plugin 提供根插件类型必须嵌入的 Base。插件 ID 只来自 plugin.yaml，根类型无需实现 ID()。
package plugin

// Base 由根插件以值的形式嵌入：
//
//	type RSS struct{ plugin.Base }
//
// 运行时状态（日志、Session 信息等）由 SDK 在启动时注入，T-02 起实现。
type Base struct {
	_ struct{} // 保留给 SDK 注入的运行时状态，禁止插件依赖其字段
}

// Root 是嵌入了 Base 的根插件；未导出的方法保证只有嵌入 Base 的类型才能满足它。
type Root interface {
	alonabotRoot()
}

func (*Base) alonabotRoot() {}
