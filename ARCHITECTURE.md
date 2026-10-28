# AlonaBot 架构地图

负责范围：模块职责、允许的 import 方向、状态归属。行为细节以 `spec/` 为准（见 [spec/README.md](spec/README.md)）。
更新时机：新增顶层包、改变包之间的依赖或状态归属时，同步修改本文件和 [docs/development.md](docs/development.md) 的检查表。

**当前状态：** 下表「状态」列给出每个模块的实际范围；标「拟建」的尚未实现，实施顺序见 [spec/PLAN.md](spec/PLAN.md)。

## 模块与职责

单 Go module `github.com/AlonaBot/AlonaBot`；内部按职责分包，不拆 module。

| 模块（路径） | 负责 | 不负责 | 状态 |
| --- | --- | --- | --- |
| `contracts/` | `.proto` 源（`proto/`）与生成类型（`pb/alonabotv1`）、manifest/release/消息结构校验、调度触发计算（`schedule`）、兼容 fixture；生成工具 `contracts/internal/protocgen`（只由 go:generate 运行） | 任何运行时 I/O、会话状态、平台/GORM 依赖 | 已实现：proto/生成、`validate`（Envelope/协商/Chain）、`manifest`、`release`、`schedule` |
| `internal/wire/` | 帧编解码（4 MiB 上限）、TCP/stdio 传输、Session 调用表、准入预算、优先通道、取消 | 业务路由、权限、插件生命周期决策 | 部分实现：帧编解码、Conn、ListenLoopback、全局预算（`budget.go`）、Session 调用引擎（配对/取消/准入/优先通道、认证后身份与版本核对 `CheckPeer`）；握手与生命周期编排在 `internal/core/host` 与 `sdk` |
| `internal/tune/` | 调优参数表：名称、缺省值、单位、约束与 `--args` 解析（CONTRACTS §16） | 参数的业务含义（由使用方实现） | 拟建 |
| `internal/oslock/` | 跨进程 OS 独占文件锁（flock / LockFileEx），供 core.lock 与 runtime.lock 共用 | 锁的业务语义（何时加锁由调用方决定） | 已实现 |
| `internal/inproc/` | 库模式同进程组件的启动参数（根包构造、SDK 读取） | 启动流程本身（归 sdk 与根包） | 拟建 |
| `internal/pluginkit/` | 开发 CLI 的打包、release.json 生成与签名（plugin build/pack） | 安装与预检（归 `internal/core/packages`） | 拟建 |
| `sdk/`（含 `sdk/message`、`sdk/sqlitedb`、`sdk/gormdb`、`sdk/llm` 等子包） | 组件侧 API：根插件、Context/Send、注册、生命周期 Hook、runtime.lock；受管 SQLite 的 `database/sql` 与 GORM 两个按需入口（D-49）；文本生成服务的类型与以 `@agent` 为目标的调用辅助 | Core 状态库、管理凭证、其他插件 | 部分实现：`sdk`（Base 与 Root、Context、注册器、Collect）、`sdk/message`、`sdk/config`、`sdk/gormdb`（OpenSQLite、RunVersioned、迁移接口互斥检查；`sdk/sqlitedb` 拟建）；`Run` 完成取锁、握手、准备、激活与停止，并把 slog 默认 Logger 换成写 stderr 的 JSON 行；业务分发与 `Send` 仍返回 UNSUPPORTED |
| `internal/core/` | 状态库与 core.lock、进程、RPC 编排、路由、权限、媒体、服务、调度、HTTP、配置、管理 | 平台协议细节（交 adapters）、插件进程内逻辑 | 部分实现：`internal/core/state`（core.lock、state.sqlite、提交原语）、`internal/core/process`（精简环境、日志、进程组/Job Object、runtime.lock 临时持有）、`internal/core/host`（握手认证、Bootstrap、REGISTER 原子校验、PREPARED、提交与激活门禁、Degraded、停止/排空/并行 Shutdown、有限重启策略）、`internal/core/supervisor`（启动/停止一个插件的完整装配）、`internal/core/logging`（控制台与 JSON 格式、插件 stderr 解析、按来源级别、按日期轮转的日志文件）、`internal/core/packages`（安装包预检、受控解压与不可变包目录）、`internal/core/config`（配置 JSON 规范化、Merge Patch 与 Pending 编辑） |
| 根包 `alonabot`（模块根目录的 Go 文件） | 库模式入口 `alonabot.Run`（D-32）与 `alonabot serve` 共用的 Core 装配 | 业务逻辑（只组装 `internal/core` 与 SDK） | 拟建 |
| `plugins/help/`、`plugins/welcome/`、`plugins/rss/` | 预装插件（D-32），与第三方插件一样只经 SDK 与 Core 通信 | Core 的命令列表与权限（归 Core） | 拟建 |
| `agent/` | 官方 Agent 组件（`role: agent`，D-27、D-53～D-55）：对话循环、Provider、工具与 MCP、记忆、技能、定时任务与 `llm.generate`；只经 SDK 与 Core 通信，可被社区 Agent 替换 | Core 的工具策略、确认与会话存储（归 Core） | 部分实现：`agent/provider`（Provider 流式客户端）；其余拟建 |
| `adapters/telegram/` | 官方 Telegram Adapter 程序：接收、标准化、发送计划、能力与平台映射（D-26） | Core 策略、命令路由、游标持久化与按钮 token（归 Core） | 拟建 |
| `adapters/onebot11/` | 官方 OneBot 11 Adapter 程序：正向 WebSocket 连接、接收、标准化、发送计划、能力与平台映射，以及本地模拟服务（D-46） | Core 策略、命令路由与按钮 token（归 Core） | 拟建 |
| `plugintest/` | 插件测试套件（D-47）：在 `go test` 中用 `alonabot.Start` 启动真实 Core、测试 Adapter 与被测插件，注入事件并取得发送 | Core 的检查与派发（照常由 Core 执行，不另建路径） | 拟建 |
| `cmd/alonabot/` | Core 启停、开发者与管理 CLI 的装配入口 | 业务逻辑（只组装） | 部分实现：开发用 `alonabot run`（单插件启停、有限重启）；尚无正式管理 CLI |
| `examples/plugins/` | 随 SDK 编译的示例与集成样例 | 生产代码依赖 | 部分实现：`showcase`（签名组合示例）、`lifecycle`（可真实运行的生命周期示例） |
| `web/admin/` | 官方 WebUI 组件（`role: webui`，D-56）的源码：Vite + TypeScript + React 19 管理界面，构建产物打成静态组件包 | 管理逻辑本身（只调用带主版本号的 Admin API，CONTRACTS §7.5） | 拟建 |
| `Dockerfile`、`deploy/` | 官方容器镜像、compose 整合包入口与 `install.sh`（CONTRACTS §15.3、§15.6，D-58） | Core 行为（镜像与脚本只安装发行归档的内容） | 拟建 |
| `spec/` | 现行规格与 ADR | — | 现行 |

## 依赖方向

`A -> B` 表示 **A 的 Go 源码 import B**（静态依赖，含测试代码），不表示运行时调用或数据流。「模块」指上表每一行的路径前缀；同一模块内部的 import（如 `contracts/release -> contracts/manifest`、`sdk -> sdk/message`）不受限制。跨模块的边只允许下表列出的；标准库和已批准的第三方依赖不受此表限制。

| 允许的边 | 理由 |
| --- | --- |
| `internal/wire -> contracts` | 帧内承载生成的 Envelope 类型 |
| `internal/wire -> internal/tune`、`sdk/... -> internal/tune`、`internal/core/... -> internal/tune` | 读取调优参数的生效值 |
| `sdk/... -> contracts`、`sdk/... -> internal/wire`、`sdk/... -> internal/oslock` | 插件侧复用同一协议实现与 runtime.lock（Go 允许本 module 的 `sdk` import `internal`） |
| `internal/core/... -> contracts`、`internal/core/... -> internal/wire`、`internal/core/... -> internal/oslock` | Core 侧协议、core.lock 与停机复制时临时持有 runtime.lock |
| `adapters/... -> sdk/...`、`-> contracts` | Adapter 程序与插件示例一样，只经公共 SDK（`sdk/adapter`）与 Core 通信 |
| `cmd/alonabot -> internal/core/...`、`-> contracts`、`-> 根包`、`-> internal/tune` | 装配；`serve` 复用根包的 Core 装配；`--args` 与 `tune list` 使用调优参数表 |
| `根包 -> internal/core/...`、`-> sdk/...`、`-> contracts` | 库模式在同一进程中装配 Core 与开发中的组件 |
| `根包 -> internal/inproc`、`sdk/... -> internal/inproc` | 库模式把同进程组件的启动参数从根包交给 SDK；`internal` 保证外部模块无法构造这些参数 |
| `cmd/alonabot -> sdk/generate` | 构建检查复用 SDK 的路由生成器（拟建） |
| `cmd/alonabot -> internal/pluginkit`、`internal/pluginkit -> contracts/...`、`-> sdk/generate` | 开发 CLI 的构建与打包复用生成器与契约包 |
| `examples/plugins/... -> sdk/...`、`-> contracts` | 示例只走公共 SDK |
| `scripts/dist -> contracts/release`、`scripts/release-check -> contracts/release` | 发行构建与发布复验只用契约包核对签名；其余步骤经 `os/exec` 调用 go 与 alonabot |
| `examples/devbot -> 根包`、`-> sdk/...` | 库模式示例 |
| `plugintest -> 根包`、`-> sdk/...`、`-> internal/inproc`、`-> contracts`；`examples/plugins/... -> plugintest`（仅测试代码） | 套件经根包的 `Start` 装配 Core，经 `internal/inproc` 接收分发结束通知；示例插件的测试使用套件 |
| `plugins/... -> sdk/...`、`-> contracts`；`agent/... -> sdk/...`、`-> contracts` | 官方插件与 Agent 组件只走公共 SDK；Agent 另用已批准的 MCP go-sdk |

明确禁止（重点）：`internal/oslock`、`internal/tune` 与 `internal/inproc` 不 import 本 module 任何包；`contracts ->` 任何本 module 包；`internal/wire -> sdk | internal/core`；`sdk -> internal/core | adapters | plugins | agent | cmd`；`internal/core -> sdk | adapters | plugins | agent | cmd`；`adapters -> internal/core`；`plugins | agent -> internal/core | adapters | cmd`；`plugins`、`agent` 之间不互相 import；`internal/core | sdk | adapters | plugins | agent -> 根包`；`internal/core | sdk | adapters | 根包 -> plugintest`，以及非测试代码 `-> plugintest`；任何包 `-> examples`。依赖图不允许环。

## 状态归属

| 状态 | 唯一所有者 |
| --- | --- |
| `<data_dir>/core.lock`、`state.sqlite`（安装、组件角色与来源、代次、Journal、策略、secret 引用、工具策略、Agent 会话） | `internal/core` |
| WebUI 组件与组件图标的静态文件（不可变包目录） | `internal/core` 只读提供；内容归组件发布者 |
| `<plugins>/<id>/runtime.lock`、插件受管 DB 连接 | 插件进程内的 `sdk` |
| 插件业务数据（受管 SQLite 候选文件） | 插件；Core 只负责候选复制与准入 |
| 插件自管的外部数据（外部数据库、框架外文件） | 插件；Core 不创建、不校验、不备份（D-48） |

## 文件组织

模块内部按职责分文件，类型、方法和专属常量就近；独立共享数据契约按主题组织，实现状态仍归自己的所有者。不要求统一 DTO/常量文件或一个类型一个文件。文件内容、注释和拆分标准以 [编码规范](docs/coding-standards.md) 为准；这不改变上面的依赖表。

## 检查

依赖方向由代码审查对照上表把关；import 环由 Go 编译器拒绝。Core 可执行文件与 SDK 根包能链接的第三方模块由 `node scripts/check-go-deps.mjs` 在 CI 中检查（D-52），清单见 [docs/development.md](docs/development.md)。
