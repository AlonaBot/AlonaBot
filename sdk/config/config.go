// Package config 准备插件的完整 JSON 配置，并提供按声明类型读取的只读视图。
//
// 根插件通过 sdk.Plugin.Config 取得 [Values]，按字段类型读取配置；[Template] 还提供选中的模板键。
// 取值对象在配置准备后就绪，同一 Session 内保持不变，返回的集合可以独立修改。
//
// 包的组成：[Prepare] 使用 contracts/configschema 执行 [Loader] 回调、补值与校正、required 校验，
// 将保留未知键的完整规范 JSON、校正记录和 Values 写入 [Prepared]。
// 回调只做纯配置变换；数据库迁移与插件生命周期由 SDK 分阶段执行。
// 配置声明、准备流程与秘密字段契约见 https://alonabot.github.io/docs/tutorial/config。
package config
