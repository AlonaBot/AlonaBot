# 文档索引

| 文档 | 用途 | 权威性 |
| --- | --- | --- |
| [spec/README.md](../spec/README.md) | 规格入口：PLAN、CONTRACTS、WIRE、MESSAGE、DESIGN、TEST-MATRIX | 行为的唯一权威 |
| [spec/adr/README.md](../spec/adr/README.md) | D/P/Q 决策与历史审阅 | 解释「为什么」，不覆盖现行契约 |
| [ARCHITECTURE.md](../ARCHITECTURE.md) | 模块职责、import 方向、状态归属 | 结构规则 |
| [development.md](development.md) | 环境、命令、外部测试环境、规则与检查 | 验证入口 |
| [coding-standards.md](coding-standards.md) | 按职责组织文件、Go 约定、文件头/字段/步骤注释 | 编码审查规则 |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Issue、PR、Git 授权和 GitHub 自动化 | 贡献流程 |
| [AGENTS.md](../AGENTS.md) | 项目维护技能与规格的按任务路由 | Agent 工作入口 |
| `goal-N/`（在 `goal-N` 分支上） | 批次执行计划：input/plan/tasks 和每个 task 的证据 | 执行进度 |

按任务找文档：

- 修改协议、消息或生命周期行为：先读对应的 spec 文件；改动契约时同步 PLAN、TEST-MATRIX，并在 ADR 中记录原因。
- 新增包或新的跨包 import：读 ARCHITECTURE.md，必要时先更新依赖表。
- 记录验收证据：按 TEST-MATRIX §3 的字段，写进当前 goal 的 tasks.md。
