# AlonaBot 代理工作地图

轻量级跨平台 Go 聊天机器人框架：Core + 独立进程 Native Go 插件 + Telegram Adapter。目前处于实施初期。

## 先读什么

本仓库的维护技能位于 `.agents/skills/`，不要求个人全局技能。客户端未自动发现时直接读取下表链接；只加载本次任务需要的入口。

| 任务 | 先读 | 完成前检查 |
| --- | --- | --- |
| 实现、诊断、重构或设计 Go 代码 | [alonabot-develop](.agents/skills/alonabot-develop/SKILL.md)、[编码规范](docs/coding-standards.md) | 所有者、调用者、兼容性及必要相邻修改完整 |
| 新增或修改代码注释、结构字段、函数步骤 | [alonabot-code-commenting](.agents/skills/alonabot-code-commenting/SKILL.md) | 文件职责头、逐字段说明、函数作用及真实阶段 step comment |
| 选择测试、复现失败或记录验证 | [alonabot-test](.agents/skills/alonabot-test/SKILL.md) | 实际风险有证据，临时探针不进入提交 |
| 审查分支、commit 或 PR | [alonabot-review](.agents/skills/alonabot-review/SKILL.md) | 契约与规范两轴，问题有位置、影响和修复方案 |
| 准备或更新 PR、诊断 GitHub Actions | [alonabot-pr](.agents/skills/alonabot-pr/SKILL.md)、[贡献流程](CONTRIBUTING.md) | 授权、模板、当前 head 的检查及真实未运行项 |
| 修改文档、AGENTS/CLAUDE 或仓库技能 | [alonabot-docs](.agents/skills/alonabot-docs/SKILL.md) | 权威单一、路由有效、状态有依据 |
| 任何实现工作 | [spec/PLAN.md](spec/PLAN.md) 中对应的 T 工作包，及其引用的 CONTRACTS/WIRE/MESSAGE 小节 | 工作包的「停止和失败条件」 |
| 新增包、跨包 import | [ARCHITECTURE.md](ARCHITECTURE.md) | 依赖表允许这条边 |
| 构建、测试、依赖、外部环境 | [docs/development.md](docs/development.md) | 按需运行 `go generate ./...`、`go test ./...`、`go build ./...` |
| 记录验收 | [spec/TEST-MATRIX.md](spec/TEST-MATRIX.md) §3 | 写明测试位置、命令、平台和结果；没运行的写「未运行」 |
| 契约有缺口或冲突 | [spec/README.md](spec/README.md)「执行与维护」 | 先修契约、PLAN 和 TEST-MATRIX，再在 spec/adr/DECISIONS.md 记录原因 |
| 推进批次 goal | 当前分支上的 `goal-N/input.md`、`plan.md`、`tasks.md` | 每个 task 的回写区填写完整 |

## 项目约定

- 代码注释、commit message、文档一律用简体中文；标识符和协议名保留英文。
- 依赖只用 docs/development.md 中列出的已批准清单；需要新依赖时先停下来征求维护者同意。
- 生成代码只改源文件（`.proto`、生成器），然后重新运行 `go generate ./...`，不手改生成产物。
- 不把 fixture、模拟服务或交叉编译结果记成真实平台/环境的验证结果。
- 业务副作用不加自动重试；限额、超时等常量以 CONTRACTS 为准，不擅自增减。
- `experiments/` 是独立的历史实验，不 import，也不作为验收依据。

## Git

- `main` 只接收维护者确认后的合并。Git 写操作使用独立 worktree；用户未明确要求时不 commit、不 push。
- 批次工作仍使用 `goal-N` 和 `.worktrees/goal-N`，每个 task 一个 commit，默认不 push；普通任务使用 `codex/<主题>`。创建 PR 的明确请求及其授权范围见 [贡献流程](CONTRIBUTING.md)。
