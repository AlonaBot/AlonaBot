// Package config 维护插件配置的 JSON 文档与 Pending 编辑。
//
// 配置以完整 JSON 对象为真源，未知键随原对象保留；Canonical 给出键有序、数字保留原文的确定性形式，
// MergePatch 实现 RFC 7396。Service 把编辑写入状态库的 Pending，并只在 base generation 仍是插件当前
// Active 时认为它有效。默认值、Schema 校验、secret 与激活不在这里处理。
package config
