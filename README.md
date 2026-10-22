# AlonaBot

轻量级、跨平台的 Go 聊天机器人框架。

当前处于规格完成、待实施阶段；尚未创建正式 Core、SDK 或插件实现。Module：`github.com/AlonaBot/alonabot`。

从 [spec/PLAN.md](spec/PLAN.md) 的 T-01 开始实施；规格索引见 [spec/README.md](spec/README.md)，决策与历史依据见 [spec/adr/README.md](spec/adr/README.md)。开发环境与命令见 [docs/development.md](docs/development.md)，模块边界见 [ARCHITECTURE.md](ARCHITECTURE.md)，代理工作约定见 [AGENTS.md](AGENTS.md)。

参与开发先读 [贡献流程](CONTRIBUTING.md) 和 [编码与注释规范](docs/coding-standards.md)；项目维护技能随仓库提供，不依赖个人配置。

- 首发：Native Go 插件独立子进程、Telegram、命令/事件、Service/广播、调度、配置/数据库、HTTP、管理 CLI/WebUI 和插件安装更新。
- 后续：更多正式 Adapter、动态命令 Router、Agent 与额外客户端等，按规格范围推进。
- 实施约束：保持轻量、良好的开发体验和最小必要配置；以编译、集成和目标平台测试记录完成情况。
