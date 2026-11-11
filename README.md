# AlonaBot

轻量级、跨平台的 Go 聊天机器人框架。

Module：`github.com/AlonaBot/AlonaBot`。

**当前进度：** 已实现协议契约、生成工具链、manifest/release 校验、SDK 公共类型、受管 SQLite 验证，以及 Core 状态库与锁、Native 插件进程、TCP/stdio 双 IPC、握手/激活/停止生命周期、有限重启；验收证据与未运行项见 [spec/TEST-MATRIX.md §3.1](spec/TEST-MATRIX.md#31-已运行的分项证据t-01t-02)：只在 darwin/arm64 上运行过，Linux/Windows 只做了交叉构建。已接通命令最长路径/别名选择、文本匹配与 Raw、参数绑定、usage/silent 与自动回复、按调用者语言选择的 en-US/zh-CN usage 与插件 `Context.T` 文案、显式事件注册、原目标手动回复、实例许可准入、后台 `sdk.Send`、媒体分块上传 `sdk.Upload`、入站媒体按需读取与转授（`sdk.ReadMedia`/`sdk.GrantMedia`）及发送期间的只读媒体租约、Core 发送服务、插件 HTTP 路由经 Core 公开入口的代理（私有路由尚未接入授权，一律 401）、Core 设置文件 `core_config.json`（保序写回与首次管理 token 交付）、经 Core Broker 准入的跨插件 Service 调用、非持久 Event Bus 广播（`sdk.Publish`/`sdk.Subscribe`）、插件 interval/cron 定时任务的注册校验与 Core 调度（运行槽、超期与迟到响应），以及构建期检查并嵌入配置 Schema、Service Schema 与 `i18n/` 文案的 `alonabot plugin generate`/`plugin check`（命令与事件只经链式注册）；Telegram 接收、标准化、入站媒体取件、整链发送（物理计划、媒体上传、引用与按钮渲染、逐步回执）与文本消息编辑已实现，按钮回调、命令策略与 Rule、调度的生产装配和管理等业务能力继续实施；目前只有开发用入口 `alonabot run` 与 `alonabot plugin generate`/`check`，用法见 [docs/development.md](docs/development.md)。

最小事件回复示例见 [echo](examples/plugins/echo/main.go)：注册 `message.created`，在 Handler 中调用 `Context.Send`。真实 TCP/stdio 装配和验证见 [echo 集成测试](internal/core/supervisor/echo_test.go)；原 Target 回复无需额外实例许可；显式其他目标和后台发送按 Core 持久化的实例许可逐次准入，撤销在下一次 Send 时生效。许可读写入口当前供 Core 内部使用，CLI 装配与管理入口留给后继切片。许可分项证据见 [TEST-MATRIX §3.3](spec/TEST-MATRIX.md#33-target-与发信许可分项证据)。

Telegram Adapter 接收入口见 [alonabot-adapter-telegram](adapters/telegram/cmd/alonabot-adapter-telegram/main.go)，本地 Bot API 模拟服务见 [telegramtest](adapters/telegram/telegramtest/server.go)。OneBot 11 Adapter 已实现正向 WebSocket 连接与接收（发送与媒体随后继切片），入口见 [alonabot-adapter-onebot11](adapters/onebot11/cmd/alonabot-adapter-onebot11/main.go)，本地模拟服务见 [onebottest](adapters/onebot11/onebottest/server.go)。已在 darwin/arm64 验证真实私聊消息接收；完整 Core 装配与其他真实平台场景由后继切片联验。

按 [spec/PLAN.md](spec/PLAN.md) 的工作包继续实施；规格索引见 [spec/README.md](spec/README.md)，决策与历史依据见 [spec/adr/README.md](spec/adr/README.md)。开发环境与命令见 [docs/development.md](docs/development.md)，插件作者的规则见 [docs/plugin-guide.md](docs/plugin-guide.md)，模块边界见 [ARCHITECTURE.md](ARCHITECTURE.md)，代理工作约定见 [AGENTS.md](AGENTS.md)。

参与开发先读 [贡献流程](CONTRIBUTING.md) 和 [编码与注释规范](docs/coding-standards.md)；项目维护技能随仓库提供，不依赖个人配置。

- 首发：可安装的组件分为插件、Adapter、Agent 与 WebUI 四种，各自是独立子进程（WebUI 只是静态文件）；官方 Telegram 与 OneBot 11 Adapter、官方 Agent（含记忆，工具调用需用户确认）与官方 WebUI 都可以换成社区组件；命令/文本匹配/事件、多轮等待、Service/广播、调度、配置/数据库、HTTP、管理 CLI、组件安装更新、英文与简体中文的用户文案；面向用户的官方整合包可用发行归档、`install.sh`（Linux systemd、macOS launchd）或 `docker compose` 安装，开箱即用，开发者也可以在自己的 `main.go` 中把框架当作 Go 库运行并热部署，并用插件测试套件在 `go test` 中以真实 Core 测试插件。
- 后续：更多正式 Adapter、动态命令 Router、持久化的多轮会话、跨会话记忆检索与额外客户端等，按规格范围推进。
- 实施约束：保持轻量、良好的开发体验和最小必要配置；以编译、集成和目标平台测试记录完成情况。
