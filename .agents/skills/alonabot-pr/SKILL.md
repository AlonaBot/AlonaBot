---
name: alonabot-pr
description: 准备、创建或更新 AlonaBot PR，或诊断其 GitHub Actions 时使用；核对授权、当前 head 的检查和模板，不执行隐含的合并、发布或生产验证。
---

# 交付 AlonaBot PR 与检查 CI

本技能负责 PR 范围、GitHub 自动化边界和当前提交的交付证据。流程以[贡献流程](../../../CONTRIBUTING.zh-CN.md#提交-pr)为准，实际检查读取 [.github/workflows](../../../.github/workflows)。

## 发布前核对范围

核对仓库、base、head、独立 worktree 与最终 diff，确认只包含该任务和必要的相邻调整。用户未明确要求时不 commit 或 push；创建 PR 的要求授权该 PR 必需的提交与推送，但不授权 merge、tag、release、分支保护或生产 secret。

标题使用 `type/scope: 中文摘要`。按 [PR 模板](../../../.github/PULL_REQUEST_TEMPLATE.md)填写实际变更、验证与兼容性风险，只勾选事实成立的确认项。临时正文与本地测试放在被忽略的位置，不夹带到提交中。CI 会拒绝个人维护资料路径（`.claude/`、`spec/`、`docs/goals/` 等），这些资料在 dev 分支维护。

## 创建并读取当前检查

有授权时推送专用分支并创建或更新 PR，不 force-push。用已认证的 `gh` 核对结果 URL 和当前 head SHA：

```sh
gh pr checks PR_NUMBER --repo AlonaBot/AlonaBot
gh run view RUN_ID --repo AlonaBot/AlonaBot --json event,headSha,status,conclusion,jobs,url
gh run view RUN_ID --repo AlonaBot/AlonaBot --log-failed
```

区分 pending、skipped、failed 与 success。`Repository` 检查空白及 main 文件范围；Go 检查只在改动 Go 源码、协议或 CI 时运行。没有运行 Go 检查不等于 Go 验证通过。main 要求必需检查通过且分支与 main 同步。

## 修复 CI 的实际原因

先读 workflow 和它执行的脚本，定位首个因果错误及 run、attempt 和 SHA。区分代码失败、环境缺失和基础设施故障；只有改动或明确的基础设施问题才值得重跑，不用重试掩盖确定性故障。

`pull_request_target` 工作流不执行 PR 代码；`PR metadata` 只把标题和正文当数据读取。产品检查保持只读权限、无生产凭据，不为了通过 CI 放宽依赖、停止预算或信任边界。

交付 PR URL、当前检查结果和仍影响合并的事项。权限或 API 错误与产品失败分开说明。
