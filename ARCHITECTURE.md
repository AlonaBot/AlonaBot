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
| `sdk/`（含 `sdk/message`、`sdk/gormdb` 等子包） | 插件侧 API：根插件、Context/Send、注册、生命周期 Hook、runtime.lock、按需 GORM | Core 状态库、管理凭证、其他插件 | 部分实现：`sdk`（Context、注册器、Collect）、`sdk/plugin`、`sdk/message`、`sdk/config`、`sdk/gormdb`（OpenSQLite/OpenPostgres、RunVersioned、迁移接口互斥检查）；`Run` 完成取锁、握手、准备、激活与停止，并把 slog 默认 Logger 换成写 stderr 的 JSON 行；业务分发与 `Send` 仍返回 UNSUPPORTED |
| `internal/core/` | 状态库与 core.lock、进程、RPC 编排、路由、权限、媒体、服务、调度、HTTP、配置、管理 | 平台协议细节（交 adapters）、插件进程内逻辑 | 部分实现：`internal/core/state`（core.lock、state.sqlite、提交原语）、`internal/core/process`（精简环境、日志、进程组/Job Object、runtime.lock 临时持有）、`internal/core/host`（握手认证、Bootstrap、REGISTER 原子校验、PREPARED、提交与激活门禁、Degraded、停止/排空/并行 Shutdown、有限重启策略）、`internal/core/supervisor`（启动/停止一个插件的完整装配）、`internal/core/logging`（控制台与 JSON 格式、插件 stderr 解析、按来源级别、按日期轮转的日志文件） |
| 根包 `alonabot`（模块根目录的 Go 文件） | 库模式入口 `alonabot.Run`（D-32）与 `alonabot serve` 共用的 Core 装配 | 业务逻辑（只组装 `internal/core` 与 SDK） | 拟建 |
| `plugins/agent/`、`plugins/memory/`、`plugins/help/`、`plugins/welcome/`、`plugins/rss/` | 官方 Agent 插件、官方记忆插件与预装插件（D-27、D-28、D-32），与第三方插件一样只经 SDK 与 Core 通信 | Core 的工具策略、确认与会话存储（归 Core） | 已实现：`plugins/agent/provider`（Provider 流式客户端）；其余拟建 |
| `adapters/telegram/` | 官方 Telegram Adapter 程序：接收、标准化、发送计划、能力与平台映射（D-26） | Core 策略、命令路由、游标持久化与按钮 token（归 Core） | 拟建 |
| `cmd/alonabot/` | Core 启停、开发者与管理 CLI 的装配入口 | 业务逻辑（只组装） | 部分实现：开发用 `alonabot run`（单插件启停、有限重启）；尚无正式管理 CLI |
| `examples/plugins/` | 随 SDK 编译的示例与集成样例 | 生产代码依赖 | 部分实现：`minimal`（签名组合示例）、`lifecycle`（可真实运行的生命周期示例） |
| `web/admin/` | Vite + TypeScript + React 19 管理界面 | 管理逻辑本身（调用 Core 管理 API） | 拟建 |
| `Dockerfile`、`deploy/` | 官方容器镜像与 compose 示例（CONTRACTS §15.3） | Core 行为（镜像只打包发行归档的内容） | 拟建 |
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
| `cmd/alonabot -> internal/core/...`、`-> contracts`、`-> 根包` | 装配；`serve` 复用根包的 Core 装配 |
| `根包 -> internal/core/...`、`-> sdk/...`、`-> contracts` | 库模式在同一进程中装配 Core 与开发中的组件 |
| `cmd/alonabot -> sdk/generate` | 构建检查复用 SDK 的路由生成器（拟建） |
| `examples/plugins/... -> sdk/...`、`-> contracts` | 示例只走公共 SDK |
| `examples/devbot -> 根包`、`-> sdk/...` | 库模式示例 |
| `plugins/... -> sdk/...`、`-> contracts` | 官方插件只走公共 SDK；Agent 插件另用已批准的 MCP go-sdk |

明确禁止（重点）：`internal/oslock` 与 `internal/tune` 不 import 本 module 任何包；`contracts ->` 任何本 module 包；`internal/wire -> sdk | internal/core`；`sdk -> internal/core | adapters | plugins | cmd`；`internal/core -> sdk | adapters | plugins | cmd`；`adapters -> internal/core`；`plugins -> internal/core | adapters | cmd`；`internal/core | sdk | adapters | plugins -> 根包`；任何包 `-> examples`。依赖图不允许环。

## 状态归属

| 状态 | 唯一所有者 |
| --- | --- |
| `<data_dir>/core.lock`、`state.sqlite`（安装、代次、Journal、策略、secret 引用、工具策略、Agent 会话） | `internal/core` |
| `<plugins>/<id>/runtime.lock`、插件受管 DB 连接 | 插件进程内的 `sdk` |
| 插件业务数据（SQLite 候选文件 / PG schema） | 插件；Core 只负责候选复制与准入 |

## 文件组织

模块内部按职责分文件，类型、方法和专属常量就近；独立共享数据契约按主题组织，实现状态仍归自己的所有者。不要求统一 DTO/常量文件或一个类型一个文件。文件内容、注释和拆分标准以 [编码规范](docs/coding-standards.md) 为准；这不改变上面的依赖表。

## 检查

依赖方向由代码审查对照上表把关；import 环由 Go 编译器拒绝。
