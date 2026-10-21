# AlonaBot 架构地图

负责范围：模块职责、允许的 import 方向、状态归属。行为细节以 `spec/` 为准（见 [spec/README.md](spec/README.md)）。
更新时机：新增顶层包、改变包之间的依赖或状态归属时，同步修改本文件和 [docs/development.md](docs/development.md) 的检查表。

**当前状态（2026-10-21，goal-1 进行中）：** `contracts/` 已有 proto 源、生成类型、纯函数校验（validate/manifest/release）与测试；`sdk/` 有公共类型与注册收集，`examples/plugins/minimal` 可编译；其余包仍是 PLAN §2 的拟建结构，标「拟建」的尚未实现。

## 模块与职责

单 Go module `github.com/AlonaBot/alonabot`；内部按职责分包，不拆 module。

| 模块（路径） | 负责 | 不负责 | 首次交付 | 状态 |
| --- | --- | --- | --- | --- |
| `contracts/` | `.proto` 源（`proto/`）与生成类型（`pb/alonabotv1`）、manifest/release/消息结构校验、兼容 fixture；生成工具 `contracts/internal/protocgen`（只由 go:generate 运行） | 任何运行时 I/O、会话状态、平台/GORM 依赖 | T-01 | 已实现：proto/生成、`validate`（Envelope/协商/Chain）、`manifest`、`release` |
| `internal/wire/` | 帧编解码（4 MiB 上限）、TCP/stdio 传输、Session 调用表、准入预算、优先通道、取消 | 业务路由、权限、插件生命周期决策 | T-02 | 拟建 |
| `sdk/`（含 `sdk/message`、`sdk/gormdb` 等子包） | 插件侧 API：根插件、Context/Send、注册、生命周期 Hook、runtime.lock、按需 GORM | Core 状态库、管理凭证、其他插件 | T-01 类型，T-02 起实现 | 部分实现：`sdk`（Context、注册器、Collect）、`sdk/plugin`、`sdk/message`、`sdk/config`、`sdk/gormdb`（OpenSQLite/OpenPostgres、RunVersioned、迁移接口互斥检查）；Run/Send 返回 UNSUPPORTED |
| `internal/core/` | 状态库与 core.lock、进程、RPC 编排、路由、权限、媒体、服务、调度、HTTP、配置、管理 | 平台协议细节（交 adapters）、插件进程内逻辑 | T-02 起 | 拟建 |
| `internal/adapters/telegram/` | Telegram 接收、发送计划、能力与平台映射 | Core 策略、命令路由 | T-03 | 拟建 |
| `cmd/alonabot/` | Core 启停、开发者与管理 CLI 的装配入口 | 业务逻辑（只组装） | T-02 起 | 拟建 |
| `examples/plugins/` | 随 SDK 编译的示例与集成样例 | 生产代码依赖 | T-01 起 | 部分实现：`minimal`（签名组合示例） |
| `web/admin/` | Vite + TypeScript + React 19 管理界面 | 管理逻辑本身（调用 Core 管理 API） | T-08/T-09 | 拟建 |
| `experiments/ogre-dna/` | 历史实验，独立嵌套 module | 不参与产品构建与验收 | 已存在 | 隔离 |
| `spec/` | 现行规格与 ADR | — | 已存在 | 现行 |

## 依赖方向

`A -> B` 表示 **A 的 Go 源码 import B**（静态依赖），不表示运行时调用或数据流。下表未列出的模块内 import 边一律禁止；标准库和已批准的第三方依赖不受此表限制。

| 允许的边 | 理由 |
| --- | --- |
| `internal/wire -> contracts` | 帧内承载生成的 Envelope 类型 |
| `sdk/... -> contracts`、`sdk/... -> internal/wire` | 插件侧复用同一协议实现（Go 允许本 module 的 `sdk` import `internal`） |
| `internal/core/... -> contracts`、`internal/core/... -> internal/wire` | Core 侧协议 |
| `internal/adapters/... -> internal/core/<adapter 接口包>`、`-> contracts` | 实现 Core 定义的 Adapter 接口；具体接口包名在 T-03 确定后写入此处 |
| `cmd/alonabot -> internal/core/...`、`-> internal/adapters/...`、`-> contracts` | 装配 |
| `cmd/alonabot -> sdk/generate` | T-04 构建检查复用同一生成器 |
| `examples/plugins/... -> sdk/...`、`-> contracts` | 示例只走公共 SDK |

明确禁止（重点）：`contracts ->` 任何本 module 包；`internal/wire -> sdk | internal/core`；`sdk -> internal/core | internal/adapters | cmd`；`internal/core -> sdk | internal/adapters | cmd`；任何包 `-> examples | experiments`。依赖图不允许环。

## 状态归属

| 状态 | 唯一所有者 |
| --- | --- |
| `<data_dir>/core.lock`、`state.sqlite`（安装、代次、Journal、策略、secret 引用） | `internal/core` |
| `<plugins>/<id>/runtime.lock`、插件受管 DB 连接 | 插件进程内的 `sdk` |
| 插件业务数据（SQLite 候选文件 / PG schema） | 插件；Core 只负责候选复制与准入 |

## 检查

import 边规则由 goal-1 中的架构守卫测试执行（见 [docs/development.md](docs/development.md)「规则与检查」）；守卫落地之前，这些规则只是审查规则。
