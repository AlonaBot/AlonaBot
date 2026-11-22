# 未发布

- 插件与 Agent 配置 Schema 改用 `plugin_config_schema.json`，Adapter 保留 `adapter_config_schema.json`。旧 `config_schema.json` 返回包含组件身份的改名与重建错误。先停用旧组件并升级 Core，再经管理端更新使用新 SDK 重建的组件包，最后启用；原配置值、秘密和受管数据保留。

- 新增 Adapter 实时操作目录、管理认证统一调用 API、`sdk/debug` 客户端和 `adapter instance list/actions/call` CLI。目录按真实 SDK 接口、能力及平台动作生成，调用保留部分结果、错误与 trace。使用新增目录 RPC 时，Adapter 须使用同一版本 SDK 重建。

- 扩展 plugintest：多 Adapter/实例/账号、可配置平台结果与完整调用记录、真实媒体授权取件、异步 Submit/Wait 和 trace 归属、Scheduler/continuation 共用手动时钟。新增 Connect 经管理认证注入标准事件并收集真实路由、回复与错误；关闭客户端保持 Core 运行。标准调试操作使用 `sdk/` 前缀，平台 Action 保留既有名称。Outcomes 保留实际 Handler 结果，参数解析和准入失败由带阶段与 trace 的 Diagnostics 单独呈现；同平台各实例的能力配置独立生效。
