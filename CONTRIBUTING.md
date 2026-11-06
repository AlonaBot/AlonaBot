# 参与 AlonaBot 开发

本文件负责贡献流程、Git 授权与 PR 交付；代码和注释规则见 [编码规范](docs/coding-standards.md)，环境和检查命令见 [开发指南](docs/development.md)。流程或自动化变化时同步维护本文件。

## 从一个明确的问题开始

Bug Issue 写清发生了什么和复现步骤，能提供版本、系统和日志更好。功能 Issue 说明要解决的问题，期望的效果可选。日志与配置先脱敏，不提交 Bot Token、Session Secret、数据库凭证或真实用户数据。

行为以 [spec](spec/README.md) 为准，结构以 [ARCHITECTURE.md](ARCHITECTURE.md) 为准。实现前读取对应 T 工作包及失败停止条件；契约有冲突时先按 spec 的维护流程解决，再实现。

## 在独立 worktree 中修改

- 先检查分支、基线和已有修改，保留其他人的工作。Git 写操作在独立 worktree 中完成。
- 普通任务使用 `codex/<主题>` 分支；切片使用 `goal-N` 和 `.worktrees/goal-N`，每个 task 一个 commit。
- 普通任务中 Agent 默认不 commit、不 push；用户明确要求时才执行。要求创建 PR 授权该 PR 必需的 commit、push 和创建操作，不授权 merge、发布、改 tag 或仓库保护规则。切片中的提交、推送和合并授权见下文「角色与合并」。
- commit message 用简体中文，说明改了什么以及必要原因，不要求固定类型前缀。
- 一个 PR 解决一个连贯问题。规范/工具建设与产品修复分开；不把本地临时测试、个人配置或验收草稿放进 PR。

## 角色与合并

实施按 [PLAN §3.1](spec/PLAN.md#31-执行切片) 划分的章节与各章节实施计划中的切片推进，一个切片对应一个 `goal-N` 分支和一个 PR。

| 角色 | 负责 | Git 授权 |
| --- | --- | --- |
| 维护者 | 范围与契约取舍、启动执行者、合并审查者自己的改动 | 全部 |
| 审查者 | 维护 [docs/goals/](docs/goals/README.md) 的 goal 语句与 [执行顺序](docs/execution-order.md)；审查执行者的 PR；修订计划与流程 | 向执行者的 PR 分支推送普通 commit（切片状态、文档同步）；审查通过后 squash 合并执行者的切片 PR；合并后在 main 上直接更新 docs/goals/ 中的状态与接口段。其他改动走普通 PR，由维护者合并 |
| 执行者 | 按 goal 语句和 [执行技能](.agents/skills/alonabot-execute/SKILL.md) 完成 task、推送并开 PR、处理审查意见 | 在自己的 `goal-N` 分支上 commit 和推送，创建和更新该 PR；不合并、不推送 `main`、不 force-push |

审查者合并切片 PR 前，必须同时满足以下条件：

- 按 [审查技能](.agents/skills/alonabot-review/SKILL.md) 完成契约与规范两条轴的审查，并按 goal 语句和 PLAN 失败停止条件自己写临时测试实际运行过，没有未解决的 P0/P1。
- PR 正文列全了每个 task 的结果与证据；阻塞项要么已解决，要么经维护者同意延后。
- 所属章节实施计划的切片表中该切片的状态已在本 PR 内更新为「已完成」，PR 正文「文档同步」列出的内容已补上。
- `Quality gate` 与 `PR metadata` 在当前 head 上通过，合并时用 `gh pr merge --squash --match-head-commit <完整 SHA>` 锁定审查过的 commit。

仓库目前没有启用分支保护，以上条件靠审查者执行，不是平台强制的规则。合并后确认 `main` 上的 `Quality gate` 通过。

## 用证据交付

先选择能覆盖实际风险的最小检查。小而可撤销的改动不添加照搬实现的测试；已有测试通过后，没有相关新改动、失败或未解决问题就不重跑或扩大范围。仓库只提交集成测试与协议编号检查，单元测试和临时复现写在被忽略的 `*_local_test.go` 或 `tmp/test/`，规则见 [编码规范「测试提交规则」](docs/coding-standards.md#测试提交规则)。

新增的非平凡逻辑至少用一项测试验证，但只有集成测试进入仓库；信任边界、数据持久化和公开协议兼容性靠集成测试覆盖，`sdk/` 与 `internal/core/supervisor/` 的例外见编码规范（D-86）。跨模块大改动须进行独立代码审查：使用可用的 `code-review-expert`，没有该技能时按仓库 [审查技能](.agents/skills/alonabot-review/SKILL.md) 完成同等检查并说明工具差异。

PR 按 [模板](.github/PULL_REQUEST_TEMPLATE.md) 写清变更原因、实际命令及结果、兼容性与风险。未运行的检查说明原因，不把交叉编译、fixture 或跳过的测试写成真实环境验证；V 编号证据遵循 [TEST-MATRIX §3](spec/TEST-MATRIX.md#3-证据和失败门槛)。交付说明写在 PR 或回复，不另建报告文件。

## GitHub 自动化的范围

- Issue 表单将问题指派给维护者；PR 自动指派维护者、同步路径标签，非维护者提交的非草稿 PR 自动请求审阅。
- `Quality gate` 运行 `git diff --check`，并在存在产品 Go 源码时执行格式、生成一致性、vet、race 测试和五平台构建；依赖允许清单由审查核对。具体触发与命令以 [工作流](.github/workflows/ci.yml) 为准。
- `PR metadata` 检查非草稿 PR 的说明段落及必填确认项；这些声明仍需要人工审查，不代表自动证明了代码质量。
- 带写权限的 triage 只操作 GitHub 元数据，不 checkout 或执行 PR 代码；代码验证使用只读权限且不接触生产 secret。
- 工作流文件不等于已启用合并保护；维护者确认远端保护配置后才可称为强制合并门禁。

这些流程参考 [pixiv-cli 的仓库配置](https://github.com/FlanChanXwO/pixiv-cli/tree/ef4accb9853dc422d8670613e9112401465788cd/.github)，仅保留 AlonaBot 当前使用的部分。发布、容器、Homebrew、外部服务 `/test` 和自动关闭 PR 不在本仓库的基础流程中。
