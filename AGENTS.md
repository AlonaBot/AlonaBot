# AlonaBot 代理工作地图

轻量级跨平台 Go 聊天机器人框架：Core + 独立进程 Native Go 插件 + Telegram Adapter。目前处于实施初期。

## 先读什么

本仓库的维护技能位于 `.agents/skills/`，不要求个人全局技能。客户端未自动发现时直接读取下表链接；只加载本次任务需要的入口。

| 任务 | 先读 | 完成前检查 |
| --- | --- | --- |
| 实现、诊断、重构或设计 Go 代码 | [alonabot-develop](.agents/skills/alonabot-develop/SKILL.md)、[编码规范](docs/coding-standards.md)（Google Go Style 与项目术语） | 所有者、调用者、兼容性及必要相邻修改完整 |
| 新增或修改代码注释、结构字段、函数步骤 | [alonabot-code-commenting](.agents/skills/alonabot-code-commenting/SKILL.md) | 标准库风格的包文档、文档注释、必要的字段注释与 step 注释 |
| 选择测试、复现失败或记录验证 | [alonabot-test](.agents/skills/alonabot-test/SKILL.md) | 只提交集成测试与协议编号检查；本地测试写在 `*_local_test.go` |
| 审查分支、commit 或 PR | [alonabot-review](.agents/skills/alonabot-review/SKILL.md) | 契约与规范两轴，问题有位置、影响和修复方案 |
| 准备或更新 PR、诊断 GitHub Actions | [alonabot-pr](.agents/skills/alonabot-pr/SKILL.md)、[贡献流程](CONTRIBUTING.md) | 授权、模板、当前 head 的检查及真实未运行项 |
| 修改文档、AGENTS/CLAUDE 或仓库技能 | [alonabot-docs](.agents/skills/alonabot-docs/SKILL.md) | 权威单一、路由有效、状态有依据 |
| 任何实现工作 | [spec/PLAN.md](spec/PLAN.md) 中对应的 T 工作包，及其引用的 CONTRACTS/WIRE/MESSAGE 小节 | 工作包的「停止和失败条件」 |
| 新增包、跨包 import | [ARCHITECTURE.md](ARCHITECTURE.md) | 依赖表允许这条边 |
| 构建、测试、依赖、外部环境 | [docs/development.md](docs/development.md) | 按需运行 `go generate ./...`、`go test ./...`、`go build ./...` |
| 记录验收 | [spec/TEST-MATRIX.md](spec/TEST-MATRIX.md) §3 | 写明测试位置、命令、平台和结果；没运行的写「未运行」 |
| 契约有缺口或冲突 | [spec/README.md](spec/README.md)「执行与维护」 | 先修契约、PLAN 和 TEST-MATRIX，再在 spec/adr/DECISIONS.md 记录原因 |
| 执行切片 goal（执行者） | [goals.md](goals.md) 中该 goal 的语句、[alonabot-execute](.agents/skills/alonabot-execute/SKILL.md)、[执行顺序](docs/execution-order.md)、本地 `goal-N/`（被 git 忽略） | 每个 task 的回写区填写完整；完成后推送并开 PR，不合并 |
| 编写 goal 语句、审查并合并执行者 PR（审查者） | [goals.md](goals.md)、[执行顺序](docs/execution-order.md)、[alonabot-review](.agents/skills/alonabot-review/SKILL.md)「审查执行者的切片 PR」 | 语句的接口与测试足够具体；合并前核对 head SHA、CI 与自己的复现结果；合并后更新后继 goal |

## 项目约定

- Go 代码遵循 [Google Go Style Guide](https://google.github.io/styleguide/go/)（Guide、Decisions、Best Practices；可用时加载 `go-google-style-guide`、`go-google-style-decisions`、`go-google-best-practices` 技能），项目自己的命名、文件头、注释和测试提交规则见 [编码规范](docs/coding-standards.md)，两者冲突时以编码规范为准。
- 代码注释、commit message、文档一律用简体中文；标识符和协议名保留英文。
- 依赖只用 docs/development.md 中列出的已批准清单；需要新依赖时先停下来征求维护者同意。
- 生成代码只改源文件（`.proto`、生成器），然后重新运行 `go generate ./...`，不手改生成产物。
- 不把 fixture、模拟服务或交叉编译结果记成真实平台/环境的验证结果。
- 业务副作用不加自动重试；限额、超时等常量以 CONTRACTS 为准，不擅自增减。

## Git

- `main` 只接收两类合并：执行者的切片 PR 经审查者审查通过后，由审查者 squash 合并；其他改动由维护者确认后合并。角色与授权见 [贡献流程「角色与合并」](CONTRIBUTING.md#角色与合并)。
- Git 写操作使用独立 worktree。除上述角色授权外，用户未明确要求时不 commit、不 push。
- 切片使用 `goal-N` 分支和 `.worktrees/goal-N`，每个 task 一个 commit，完成后由执行者推送并开 PR；普通任务使用 `codex/<主题>`。创建 PR 的明确请求及其授权范围见 [贡献流程](CONTRIBUTING.md)。
