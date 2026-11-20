> 维护归档：此处保留 2026-11-20 的原始记录。现行分支、派发与授权规则见根 [AGENTS.md](../../../AGENTS.md)；重新派发前须核对当前 main。

# 研究记录与验证备忘

> 历史记录：本文保存 2026-10-21 及更早的审阅、来源或实验结果；当时的文件数、链接数和“本轮”均指该次记录。2026-10-22 迁入 adr/history/，相对引用已更新；当前执行步骤见 [PLAN.md](../../PLAN.md)。

更新日期：2026-10-21。此文档保存**历史证据、局限和后续实验**，不把估算当作正式性能承诺。

## E-02：旧版 PRD 逐页复核（2026-10-19 历史记录）

- 原始文件：`Go_WASM_Bot_Framework_Enterprise_PRD_v1.0.pdf`（2026-10-11，32 页，旧聊天附件）。
- 2026-10-19 的记录称已读取 PDF 中全部 17 个章节及 FR-01～FR-30 表格、接口伪代码、验收条款，并与当时规格对照；历史映射见 [LEGACY-PRD-COMPARISON.md](LEGACY-PRD-COMPARISON.md)。本轮未重新读取 PDF 原件。
- **继承**：MessageChain/Send 语义、跨平台事件/Adapter 能力、配置 Active/Pending、安全迁移、更新 Generation/Journal、性能与生命周期测试。
- **废弃/改写**：WASM Reactor/ABI、Host SQL/FS/HTTP/Exec 的强制隔离前提、单实例串行、旧版同包多根插件类、默认清理未知键和通用自动补发。
- **原件存放注意**：历史记录指向原聊天附件，原件尚未放入 Mac 的 alonabot/spec；本轮没有验证原聊天附件目前仍可取得。本目录保存差异记录。
- 旧版 PRD 是 Draft，其测试结论也主要属于可行性推断，不能当作 Native 方案的验证报告。


## E-01：Native Go 多进程轻量性初步原型（历史记录）

此前曾在 Linux/amd64、Go 1.23.2 的容器环境运行两个简化 Go 子进程原型，其中 HTTP worker 可执行本地 HTTP 请求、并发 goroutine、JSON-over-stdio 测试 IPC。以下为前次对话的测试记录，**本次文档导出未重新运行实验**：

| 情形 | RSS 总和 | PSS 总和 |
| --- | ---: | ---: |
| 1 个相同 worker 进程 | 5.25 MiB | 4.67 MiB |
| 5 个相同 worker 进程 | 26.14 MiB | 8.35 MiB |
| 10 个相同 worker 进程 | 52.36 MiB | 12.99 MiB |
| 20 个相同 worker 进程 | 104.55 MiB | 21.97 MiB |
| 20 个进程分别执行一次本地 HTTP 下载后 | 131.97 MiB | 29.43 MiB |

补充：单进程执行 8 个并发、每个约 150 ms 延迟的本地 HTTP 请求，约 0.154 s 完成。测试服务端增大 TCP accept backlog 后才得到该结果；先前默认服务器 backlog 曾造成误判。

**局限：**

1. PSS 反映共享映射的摊销，比各进程 RSS 直接求和更适合粗略观察多个相同可执行文件的共享效果；它不是独立插件真实总内存承诺。
2. 此测试的 20 个 worker 使用**相同编译产物**，共享二进制代码页；真实插件分别编译、依赖不同，PSS 可能显著增加。
3. 未包括 Core、真实 Telegram Adapter、GORM/SQLite/PostgreSQL、真实 RSS 下载、长时内存增长和 macOS/Windows。
4. 没有测 IPC 高负载、崩溃循环、热更新或进程退出后的清理。
5. 原型仅支持“Native 值得继续验证”的方向性判断，**不能作为生产容量规划**。

## 历史实验清单（E-01 后续；下列 V 编号不是现行 TEST-MATRIX 编号）

- V-01：构建 3–5 个依赖不同的独立插件程序，记录总 RSS、PSS（Linux）与进程私有内存。
- V-02：加载 GORM + SQLite，完成真实 RSS 抓取/图片下载与并发定时任务。
- V-03：验证 TCP 与 stdio 两套传输在一致 Protobuf framing 下的功能、延迟、退出和日志处理。
- V-04：在 Linux AMD64/ARM64、Windows AMD64、macOS AMD64/ARM64 实测启动、优雅停止、崩溃恢复和文件权限。
- V-05：注入 TCP 中断、Session 过期、RPC 取消、循环 Service Call 和重复事件。
- V-06：验证数据库迁移中断、更新回滚、配置 secret 遮蔽和插件来源匹配。

## 借鉴对象（概念参考，非直接移植）

- [AstrBot](https://github.com/AstrBotDevs/AstrBot)：消息处理管道、事件阶段 Hook、命令分组与插件生命周期。
- [NoneBot2](https://github.com/nonebot/nonebot2)：Rule / Permission 组合、Matcher 优先级和事件停止传播。
- [Go 官方文档](https://go.dev/doc/)：跨平台构建、goroutine、context、os/exec。
- [Protocol Buffers](https://protobuf.dev/)：传输协议的数据类型定义。
- [GORM](https://gorm.io/docs/)：原生数据库访问与迁移选项。

参考链接仅用作继续设计和验证的入口，本次没有对对应版本源码进行逐项复核。之前的 WASM 方案文档如有留存，不适用于当前 Native-only 架构。

## 风险摘要

- **Native 插件运行权限**：子进程不是沙箱，不能依赖“独立进程”实现文件/网络访问隔离。
- **取消语义**：context 取消只会通知协作式任务，不能保证撤销已经发生的外部副作用。
- **日志**：stdio stdout 必须严格保留给 Protobuf；第三方库直接写原始 fd 可能破坏协议。
- **重复效果**：业务调用崩溃/超时后可能已对外发消息、写 DB，Core 禁止自动重试。
- **兼容与分发**：预构建跨平台需确保二进制架构、CGO、目标 OS 支持和可信来源。


## E-03：本轮技术依据核对（2026-10-21）

本轮读取现行 12 份 spec，并检查以下官方主源；来源说明技术约束，具体默认值与接口为本项目的设计选择。没有重新运行 E-01，也未把文档审阅记为 Core/SDK 测试通过。

| 主源 | 已核对的约束 | 对规格的影响 |
| --- | --- | --- |
| [Protocol Buffers proto3](https://protobuf.dev/programming-guides/proto3/) | 首个枚举项为 0；字段号与 presence 影响兼容 | Wire 添加 UNSPECIFIED，固定可选字段与控制失败，禁止重用已发布编号 |
| [SQLite Backup API](https://sqlite.org/backup.html) | 完成备份可得到一致数据库副本 | 受管 SQLite 候选初始化只写独立副本，不能用普通文件复制忽略 WAL |
| [PostgreSQL CREATE INDEX](https://www.postgresql.org/docs/current/sql-createindex.html) | CONCURRENTLY 不能在事务块内执行 | 不承诺任意 PostgreSQL DDL 都可自动回滚；保留维护门槛 |
| [PostgreSQL Schemas](https://www.postgresql.org/docs/current/ddl-schemas.html) | schema/search_path 与 role 权限共同决定可信访问 | 管理员预置插件 role/schema，固定受信 search_path，不复用 Core 管理凭证 |
| [Go ReverseProxy](https://pkg.go.dev/net/http/httputil) | Rewrite 支持重建代理目标和可信转发头 | 使用标准反向代理，清理外部身份头并注入当前 Session 内部认证 |
| [WHATWG Origin](https://html.spec.whatwg.org/multipage/browsers.html#concept-origin) | Origin 由 scheme/host/port 区分，路径不是隔离 | 同一公开端口通过两个全局主机名分离 Admin 与插件页面 |
| [Telegram Bot API](https://core.telegram.org/bots/api) | 媒体组 2～10 项、callback_data 1～64 bytes、UTF-16 实体、暂缺可用消息 ID；显式 update 订阅、空闲后 ID 可能重置、匿名 sender_chat | 有序物理计划与真实回执；按钮 token；明确入站游标/订阅与可信 Actor，平台限制在首次发送前检查 |

依赖精确版本在 T-01 验证后写入 go.mod/go.sum 与构建工具锁定记录；当前文档不宣称任意最新版均已编译兼容。

本轮补充选型主源：[GORM 数据库连接](https://gorm.io/docs/connecting_to_the_database.html)、[libtnb/sqlite](https://github.com/libtnb/sqlite)、[modernc SQLite Backup](https://pkg.go.dev/modernc.org/sqlite#Backup)。默认复用同一 modernc 底层支持 Core 元数据/备份与插件 GORM；T-01 锁定实际验证通过的版本。配置 PATCH 采用 [RFC 7396](https://www.rfc-editor.org/rfc/rfc7396) 定义的对象合并/空值删除语义。
