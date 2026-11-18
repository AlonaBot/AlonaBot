# AlonaBot

轻量级、跨平台的 Go 聊天机器人框架。

Module：`github.com/AlonaBot/AlonaBot`。

**当前进度：** 已实现协议契约、生成工具链、manifest/release 校验、SDK 公共类型、受管 SQLite 验证，以及 Core 状态库与锁、Native 插件进程、TCP/stdio 双 IPC、握手/激活/停止生命周期、有限重启；验收证据与未运行项见 [spec/TEST-MATRIX.md §3.1](spec/TEST-MATRIX.md#31-已运行的分项证据t-01t-02)：只在 darwin/arm64 上运行过，Linux/Windows 只做了交叉构建。已接通命令最长路径/别名选择、文本匹配与 Raw、参数绑定、usage/silent 与自动回复、按调用者语言选择的 en-US/zh-CN usage 与插件 `Context.T` 文案、显式事件注册、原目标手动回复、实例许可准入、后台 `sdk.Send`、媒体分块上传 `media.Upload`、入站媒体按需读取与转授（`media.Open`/`media.Grant`）及发送期间的只读媒体租约、Core 发送服务、插件 HTTP 路由经 Core 公开入口的代理（与管理入口共用 `base_url` 主机名，私有路由复用管理认证与 CSRF 校验，响应统一加 sandbox CSP 与 nosniff 并清除插件的 Set-Cookie）、Core 设置文件 `core_config.json`（保序写回、首次管理 token 交付，以及按内置 schema 和 Adapter 的 `adapter_config_schema.json` 校正设置与实例配置，运行校正保留原文件）、经 Core Broker 准入的跨插件 Service 调用、非持久 Event Bus 广播（`bus.Event.Publish`/`bus.Subscribe`）、插件 interval/cron 定时任务的注册校验与 Core 调度（运行槽、超期与迟到响应；`alonabot run` 中随插件激活接入，停止时与插件 HTTP 一起排空），以及构建期检查并嵌入配置 Schema、Service Schema 与 `i18n/` 文案的 `alonabot build`/`check`（命令与事件只经链式注册）；Telegram 接收、标准化、入站媒体取件、整链发送（物理计划、媒体上传、引用与按钮渲染、逐步回执）与文本消息编辑已实现，按钮回调、命令策略与 Rule、调度的生产装配和管理等业务能力继续实施；开发用入口为 `alonabot run` 与顶层的 `init`、`check`、`build`、`test`、`pack`，用法见 [docs/development.md](docs/development.md)。

Core 的[管理 API](spec/ADMIN-API.md#20-webui-需要的补充接口)已支持查询保留代次、列出并暂停或恢复静态任务、查看 Adapter 实例状态、筛选和分页查看聊天目录、在线轮换管理令牌，以及列出由 SDK `.Page` 声明的插件页面及其本地化标题。对应的官方 WebUI 页面随各自切片交付。

最小事件回复示例见 [echo](examples/plugins/echo/main.go)：注册 `message.created`，在 Handler 中调用 `Context.Send`。真实 TCP/stdio 装配和验证见 [echo 集成测试](internal/core/supervisor/echo_test.go)；原 Target 回复无需额外实例许可；显式其他目标和后台发送按 Core 持久化的实例许可逐次准入，撤销在下一次 Send 时生效。许可读写入口当前供 Core 内部使用，CLI 装配与管理入口留给后继切片。许可分项证据见 [TEST-MATRIX §3.3](spec/TEST-MATRIX.md#33-target-与发信许可分项证据)。

Telegram Adapter 接收入口见 [alonabot-adapter-telegram](adapters/telegram/cmd/alonabot-adapter-telegram/main.go)，本地 Bot API 模拟服务见 [telegramtest](adapters/telegram/telegramtest/server.go)。OneBot 11 Adapter 已实现正向 WebSocket 连接与接收（发送与媒体随后继切片），入口见 [alonabot-adapter-onebot11](adapters/onebot11/cmd/alonabot-adapter-onebot11/main.go)，本地模拟服务见 [onebottest](adapters/onebot11/onebottest/server.go)。已在 darwin/arm64 验证真实私聊消息接收；完整 Core 装配与其他真实平台场景由后继切片联验。

插件配置按 `OnConfigLoad(ctx, raw)` → 补值与校正 → required 校验载入。管理 API 已提供插件、Adapter 实例与 Core 设置的校正记录、无效配置路径和清理入口；插件清理写入 Pending 后经 reload 生效。用法见[配置与 secret](docs/plugin-guide.md#7-配置与-secret)及[Core 设置](docs/development.md#core-设置与-http-入口)，WebUI 的提示与清理按钮由 T-09.9 接入。

配置文件可通过 `alonabot profile` 列出、新建、指定到 Adapter 实例或会话，以及编辑插件字段。用法见[开发指南](docs/development.md#运行-core-与管理组件)；平台设置与插件、工具策略由[管理 API](spec/ADMIN-API.md#23-配置文件)保存，运行时应用由后续切片接入。
官方 WebUI 已提供桌面与手机外壳、全局搜索、浅深色视觉令牌及本地品牌资源。运行 `npm --prefix web/admin run dev` 后，从开发服务器首页进入 `/dev/forms`，可预览配置表单和生成的 JSON Merge Patch；真实登录与管理页面由后继切片接入。安装依赖与验证命令见[官方 WebUI 开发指南](docs/development.md#官方-webui-开发)。

[showcase 的 `greeting` 字段](examples/plugins/showcase/config_schema.json) 展示了带文字和图片的[配置回复示例](docs/plugin-guide.md#配置字段的回复示例)。作者工具校验示例与翻译，Core 经管理 API 提供 Active 包引用的图片；WebUI 表单展示由后继切片实现。

Core、SDK 与管理 CLI 已支持 CRITICAL 日志；用 `alonabot logs --level critical` 查看严重故障。验收见 [TEST-MATRIX §3.2](spec/TEST-MATRIX.md#t-025-critical-日志级别)。

排查 Core 问题时，可用 `alonabot diagnose` 导出默认脱敏的诊断包。在线导出使用管理认证；停机后可从数据目录只读收集。参数与停机条件见[导出诊断包](docs/development.md#导出诊断包)。

组件可随安装包提供更新日志，作者写法见[插件发布指南](docs/plugin-guide.md#9-发布)。管理员可通过[管理 API](spec/ADMIN-API.md#22-更新内容changelog)读取组件及 Core 的更新日志。

按 [spec/PLAN.md](spec/PLAN.md) 的工作包继续实施；规格索引见 [spec/README.md](spec/README.md)，决策与历史依据见 [spec/adr/README.md](spec/adr/README.md)。开发环境与命令见 [docs/development.md](docs/development.md)，插件作者的规则见 [docs/plugin-guide.md](docs/plugin-guide.md)，模块边界见 [ARCHITECTURE.md](ARCHITECTURE.md)，代理工作约定见 [AGENTS.md](AGENTS.md)。

组件市场 ID 统一使用 `alonabot-plugin-`、`alonabot-adapter-`、`alonabot-agent-`、`alonabot-webui-` 前缀与连字符分词，本地 ID 规则保持不变（[D-172](spec/adr/DECISIONS.md)，代码实施见 [goal-177](docs/goals/core.md#goal-177--t-0927-组件市场-id-改用连字符)）。官方包 ID 与首次发布前的数据重建要求见 [CONTRACTS §15.2](spec/CONTRACTS-v0.1.md#152-发行包官方整合包与预装组件)。

官方 [help 插件](plugins/help/help.go) 已实现。启用后，在聊天中发送 `/help` 查看当前会话可用的命令和说明；发送 `/whoami` 查看收到消息的实例名与用户字符串。将该用户字符串传给 `alonabot admin add`，即可登记 Core Admin。

插件可用 [`sdk/kv` 保存少量状态](docs/plugin-guide.md#保存少量状态)，用 [`sdk.PluginDataDir(ctx)` 取得文件目录](docs/plugin-guide.md#保存插件文件)；[configdb](examples/plugins/configdb/main.go) 演示两种用法。

参与开发先读 [贡献流程](CONTRIBUTING.md) 和 [编码与注释规范](docs/coding-standards.md)；项目维护技能随仓库提供，不依赖个人配置。

- 首发：可安装的组件分为插件、Adapter、Agent 与 WebUI 四种，各自是独立子进程（WebUI 只是静态文件）；官方 Telegram 与 OneBot 11 Adapter、官方 Agent（含记忆，工具调用需用户确认）与官方 WebUI 都可以换成社区组件；命令/文本匹配/事件、多轮等待、Service/广播、调度、配置/数据库、HTTP、管理 CLI、组件安装更新、英文与简体中文的用户文案；面向用户的官方整合包可用发行归档、`install.sh`（Linux systemd、macOS launchd）或 `docker compose` 安装，开箱即用，开发者可以通过 `alonabot run --reload` 热部署组件，或在自己的 `main.go` 中用 `alonabot.Run`、`alonabot.Start` 与 `Core.Wait` 将框架作为 Go 库运行，并用插件测试套件在 `go test` 中以真实 Core 测试插件。
- 后续：更多正式 Adapter、动态命令 Router、持久化的多轮会话、跨会话记忆检索与额外客户端等，按规格范围推进。
- 实施约束：保持轻量、良好的开发体验和最小必要配置；以编译、集成和目标平台测试记录完成情况。

为插件添加管理页面可使用 `sdk/page` 的嵌入式样式与桥客户端，挂载方式见[插件开发指南](docs/plugin-guide.md#插件页面)。
