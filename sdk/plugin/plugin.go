// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：根插件需要嵌入的 Base 类型。

// Package plugin 提供根插件类型必须嵌入的 Base。插件 ID 只来自 plugin.yaml，根类型不需要实现 ID()。
package plugin

// Base 由根插件以值的形式嵌入：
//
//	type RSS struct{ plugin.Base }
//
// 字段保留给 SDK 将来注入的运行时状态（日志、Session 信息等），目前为空。
type Base struct {
	_ struct{} // 保留给 SDK 注入的运行时状态，禁止插件依赖其字段
}

// Root 约束根插件必须嵌入 Base；未导出的方法保证只有嵌入 Base 的类型才能满足它。
type Root interface {
	alonabotRoot() // 标记方法，只由 Base 实现
}

// alonabotRoot 使嵌入 Base 的类型满足 Root。
func (*Base) alonabotRoot() {}
