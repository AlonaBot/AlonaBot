# 参与 AlonaBot 开发

本文件负责贡献流程、Git 授权与 PR 交付；代码和注释规则见 [编码规范](docs/coding-standards.md)，环境和检查命令见 [开发指南](docs/development.md)。流程或自动化变化时同步维护本文件。

## 从一个明确的问题开始

Bug Issue 提供版本或 commit、OS/arch、最小复现、预期与实际结果。功能 Issue 说明要解决的问题、期望行为和兼容性影响。日志与配置先脱敏，不提交 Bot Token、Session Secret、数据库凭证或真实用户数据。

行为以 [spec](spec/README.md) 为准，结构以 [ARCHITECTURE.md](ARCHITECTURE.md) 为准。实现前读取对应 T 工作包及失败停止条件；契约有冲突时先按 spec 的维护流程解决，再实现。

## 在独立 worktree 中修改

- 先检查分支、基线和已有修改，保留其他人的工作。Git 写操作在独立 worktree 中完成。
- 普通任务使用 `codex/<主题>` 分支；批次继续使用 `goal-N` 和 `.worktrees/goal-N`，每个 task 一个 commit。
- Agent 默认不 commit、不 push；用户明确要求时才执行。要求创建 PR 授权该 PR 必需的 commit、push 和创建操作，不授权 merge、发布、改 tag 或仓库保护规则。用户为某批次明确规定的「不 push」仍适用于该批次。
- commit message 用简体中文，说明改了什么以及必要原因，不要求固定类型前缀。
- 一个 PR 解决一个连贯问题。规范/工具建设与产品修复分开；不把临时探针、个人配置或验收草稿放进 PR。

## 用证据交付

先选择能覆盖实际风险的最小检查。小而可撤销的改动不添加照搬实现的测试；已有测试通过后，没有相关新改动、失败或未解决问题就不重跑或扩大范围。临时复现与探索放在忽略的 `tmp/test/`；Go 探针使用 `tmp/test/_probe/` 并显式运行。

新增非平凡逻辑至少用一项有效测试或临时自检验证；信任边界、数据持久化和公开协议兼容性缺口保留最小回归。跨模块大改动须进行独立代码审查：使用可用的 `code-review-expert`，没有该技能时按仓库 [审查技能](.agents/skills/alonabot-review/SKILL.md) 完成同等检查并说明工具差异。

PR 按 [模板](.github/PULL_REQUEST_TEMPLATE.md) 写清变更原因、实际命令及结果、兼容性与风险。未运行的检查说明原因，不把交叉编译、fixture 或跳过的测试写成真实环境验证；V 编号证据遵循 [TEST-MATRIX §3](spec/TEST-MATRIX.md#3-证据和失败门槛)。交付说明写在 PR 或回复，不另建报告文件。

## GitHub 自动化的范围

- Issue 表单将问题指派给维护者；PR 自动指派维护者、同步路径标签，非维护者提交的非草稿 PR 自动请求审阅。
- `Quality gate` 校验仓库文档入口，并在存在产品 Go 源码时执行格式、生成一致性、vet、race 测试、临时 PostgreSQL 验证和五平台构建。具体触发与命令以 [工作流](.github/workflows/ci.yml) 为准。
- `PR metadata` 检查非草稿 PR 的说明段落及必填确认项；这些声明仍需要人工审查，不代表自动证明了代码质量。
- 带写权限的 triage 只操作 GitHub 元数据，不 checkout 或执行 PR 代码；代码验证使用只读权限且不接触生产 secret。
- 工作流文件不等于已启用合并保护；维护者确认远端保护配置后才可称为强制合并门禁。

这些流程参考 [pixiv-cli 的仓库配置](https://github.com/FlanChanXwO/pixiv-cli/tree/ef4accb9853dc422d8670613e9112401465788cd/.github)，仅保留 AlonaBot 当前使用的部分。发布、容器、Homebrew、外部服务 `/test` 和自动关闭 PR 不在本仓库的基础流程中。
