## [0.3.0] - 2026-11-23

### Added

- 从 Docker 宿主机首次打开 WebUI 时，可以在浏览器领取管理令牌。登录时也可以选择在当前设备保持会话 ([#312](https://github.com/AlonaBot/AlonaBot/pull/312))。
- 管理员可以直接从 GitHub 源码构建并安装组件，构建任务会显示进度和日志。下载设置支持 GitHub 镜像、API 地址、Go 下载地址、`GOPROXY` 和网络代理；经镜像下载的发行资产仍按 GitHub 原始摘要校验 ([#312](https://github.com/AlonaBot/AlonaBot/pull/312))。
- 管理员可以为每份配置文件选择允许全部聊天、仅允许所选聊天，或屏蔽所选聊天。该设置适用于受管组件 ([#315](https://github.com/AlonaBot/AlonaBot/pull/315))。
- Adapter 作者可以用 `page_purpose = instance_config` 声明账号配置页面。WebUI 会在添加或编辑账号时打开页面；页面可以读取配置上下文、显示二维码，并把用户的修改交回 WebUI ([#317](https://github.com/AlonaBot/AlonaBot/pull/317))。
- 开发者可以用 `alonabot adapter instance list/actions/call` 或 `sdk/debug` 查看并调用运行中 Adapter 实例的动作。联调时，`plugintest` 支持多个 Adapter 和实例、自定义平台返回值，以及异步提交和等待事件 ([#310](https://github.com/AlonaBot/AlonaBot/pull/310))。
- `alonabot --version` 会显示版本、构建日期和发行页地址 ([#309](https://github.com/AlonaBot/AlonaBot/pull/309))。
- 组件作者可以用 `OnSleep` 处理休眠，用 `agent.Run` 启动多轮 Agent 任务，并通过 AI 钩子和 Goal 接口参与任务执行。服务协议支持流式响应 ([#308](https://github.com/AlonaBot/AlonaBot/pull/308))。

### Changed

- 插件和 Agent 作者应将配置 Schema 文件改名为 `plugin_config_schema.json` 后重新构建；Adapter 继续使用 `adapter_config_schema.json`。旧版 Agent 在本版 Core 上无法运行；使用独立部署的管理员应更新到 Agent v0.3.0 后再启用。现有配置和受管数据会保留。Compose 镜像 `ghcr.io/alonabot/alonabot:0.3.0` 已包含 Agent 和 WebUI v0.3.0。
- SDK 的单次模型调用改用接收结构化请求的 `llm.Generate`。使用 `llm.Ask` 或旧 `agent.Generate` 的组件应改用新入口。只有路径命令的 Builder 提供 `Menu`、参数和别名 ([#308](https://github.com/AlonaBot/AlonaBot/pull/308))。
- 源码构建改为串行编译。可用内存不足或构建长时间没有输出时，日志会给出提示；如果编译进程被终止，任务会报告可能的内存不足并给出替代做法。Linux 上内存耗尽时会优先结束编译进程，降低 Core 同时退出的风险 ([#316](https://github.com/AlonaBot/AlonaBot/pull/316))。
- 管理 API 在组件列表中提供名称和简介，并在账号实例中提供所属组件的状态 ([#312](https://github.com/AlonaBot/AlonaBot/pull/312))。

### Fixed

- 更新处于 Degraded 状态的组件后，Core 会保留已成功启用的新版本，修复更新后组件立即被停用的问题 ([#316](https://github.com/AlonaBot/AlonaBot/pull/316))。

[0.3.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.0
