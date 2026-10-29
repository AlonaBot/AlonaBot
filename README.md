# AlonaBot

轻量级、跨平台的 Go 聊天机器人框架。

Module：`github.com/AlonaBot/AlonaBot`。

**当前进度：** 已实现协议契约、生成工具链、manifest/release 校验、SDK 公共类型、受管 SQLite 验证，以及 Core 状态库与锁、Native 插件进程、TCP/stdio 双 IPC、握手/激活/停止生命周期、有限重启；验收证据与未运行项见 [spec/TEST-MATRIX.md §3.1](spec/TEST-MATRIX.md#31-已运行的分项证据t-01t-02)：只在 darwin/arm64 上运行过，Linux/Windows 只做了交叉构建。已接通显式事件注册、原目标手动回复、实例许可准入、后台 `sdk.Send`、媒体分块上传 `sdk.Upload` 与 Core 发送服务；Telegram、命令分发、跨插件服务、调度、HTTP 和管理等业务能力继续实施；目前只有开发用入口 `alonabot run`，用法见 [docs/development.md](docs/development.md)。

最小事件回复示例见 [echo](examples/plugins/echo/main.go)：注册 `message.created`，在 Handler 中调用 `Context.Send`。真实 TCP/stdio 装配和验证见 [echo 集成测试](internal/core/supervisor/echo_test.go)；原 Target 回复无需额外实例许可；显式其他目标和后台发送按 Core 持久化的实例许可逐次准入，撤销在下一次 Send 时生效。许可读写入口当前供 Core 内部使用，CLI 装配与管理入口留给后继切片。许可分项证据见 [TEST-MATRIX §3.3](spec/TEST-MATRIX.md#33-target-与发信许可分项证据)。

按 [spec/PLAN.md](spec/PLAN.md) 的工作包继续实施；规格索引见 [spec/README.md](spec/README.md)，决策与历史依据见 [spec/adr/README.md](spec/adr/README.md)。开发环境与命令见 [docs/development.md](docs/development.md)，插件作者的规则见 [docs/plugin-guide.md](docs/plugin-guide.md)，模块边界见 [ARCHITECTURE.md](ARCHITECTURE.md)，代理工作约定见 [AGENTS.md](AGENTS.md)。

参与开发先读 [贡献流程](CONTRIBUTING.md) 和 [编码与注释规范](docs/coding-standards.md)；项目维护技能随仓库提供，不依赖个人配置。

- 首发：可安装的组件分为插件、Adapter、Agent 与 WebUI 四种，各自是独立子进程（WebUI 只是静态文件）；官方 Telegram 与 OneBot 11 Adapter、官方 Agent（含记忆，工具调用需用户确认）与官方 WebUI 都可以换成社区组件；命令/文本匹配/事件、多轮等待、Service/广播、调度、配置/数据库、HTTP、管理 CLI、组件安装更新、英文与简体中文的用户文案；面向用户的官方整合包可用发行归档、`install.sh`（Linux systemd、macOS launchd）或 `docker compose` 安装，开箱即用，开发者也可以在自己的 `main.go` 中把框架当作 Go 库运行并热部署，并用插件测试套件在 `go test` 中以真实 Core 测试插件。
- 后续：更多正式 Adapter、动态命令 Router、持久化的多轮会话、跨会话记忆检索与额外客户端等，按规格范围推进。
- 实施约束：保持轻量、良好的开发体验和最小必要配置；以编译、集成和目标平台测试记录完成情况。
