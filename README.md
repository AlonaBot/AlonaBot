# AlonaBot

轻量级、跨平台的 Go 聊天机器人框架。

Module：`github.com/AlonaBot/AlonaBot`。

**当前进度：** 已完成 T-01（协议契约、生成工具链、manifest/release 校验、SDK 公共类型、SQLite/PostgreSQL 验证）与 T-02（Core 状态库与锁、Native 插件进程、TCP/stdio 双 IPC、握手/激活/停止生命周期、有限重启）。业务能力（Telegram、命令、事件、服务、调度、HTTP、配置、管理）从 T-03 起实施；目前只有开发用入口 `alonabot run`，用法见 [docs/development.md](docs/development.md)。

按 [spec/PLAN.md](spec/PLAN.md) 的工作包继续实施；规格索引见 [spec/README.md](spec/README.md)，决策与历史依据见 [spec/adr/README.md](spec/adr/README.md)。开发环境与命令见 [docs/development.md](docs/development.md)，模块边界见 [ARCHITECTURE.md](ARCHITECTURE.md)，代理工作约定见 [AGENTS.md](AGENTS.md)。

- 首发：Native Go 插件独立子进程、Telegram、命令/事件、Service/广播、调度、配置/数据库、HTTP、管理 CLI/WebUI 和插件安装更新。
- 后续：更多正式 Adapter、动态命令 Router、Agent 与额外客户端等，按规格范围推进。
- 实施约束：保持轻量、良好的开发体验和最小必要配置；以编译、集成和目标平台测试记录完成情况。
