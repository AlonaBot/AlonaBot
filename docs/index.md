# 文档索引

| 文档 | 用途 | 权威性 |
| --- | --- | --- |
| [spec/README.md](../spec/README.md) | 规格入口：PLAN、CONTRACTS、WIRE、MESSAGE、DESIGN、TEST-MATRIX | 行为的唯一权威 |
| [spec/adr/README.md](../spec/adr/README.md) | D/P/Q 决策与历史审阅 | 解释「为什么」，不覆盖现行契约 |
| [ARCHITECTURE.md](../ARCHITECTURE.md) | 模块职责、import 方向、状态归属 | 结构规则 |
| [development.md](development.md) | 环境、命令、外部测试环境、规则与检查 | 验证入口 |
| [coding-standards.md](coding-standards.md) | 命名与术语、按职责组织文件、包文档与注释、测试提交规则 | 编码审查规则 |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Issue、PR、Git 授权和 GitHub 自动化 | 贡献流程 |
| [AGENTS.md](../AGENTS.md) | 项目维护技能与规格的按任务路由 | Agent 工作入口 |
| [goals.md](../goals.md) | 可派发的切片 goal 语句与状态 | 派发入口 |
| [execution-order.md](execution-order.md) | 并行派发、共享文件与合并顺序 | 并行约束 |

按任务找文档：

- 修改协议、消息或生命周期行为：先读对应的 spec 文件；改动契约时同步 PLAN、TEST-MATRIX，并在 ADR 中记录原因。
- 新增包或新的跨包 import：读 ARCHITECTURE.md，必要时先更新依赖表。
- 记录验收证据：按 TEST-MATRIX §3 的字段，写进当前 goal 本地的 tasks.md；`goal-*/` 不入库，最终证据落在 PR 正文。
