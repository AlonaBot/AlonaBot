---
name: alonabot-pr
description: 准备、创建或更新 AlonaBot PR，或诊断其 GitHub Actions 时使用；核对授权、当前 head 的检查和模板，不执行隐含的合并、发布或生产验证。
---

# 交付 AlonaBot PR 与检查 CI

本技能负责 PR 范围、GitHub 自动化边界和当前提交的交付证据；贡献规则读取 [CONTRIBUTING](../../../CONTRIBUTING.md)，实际配置读取 [.github/workflows](../../../.github/workflows)。

## 发布前核对范围

核对仓库、base、head、独立 worktree 与最终 diff，确认只包含该任务和必要相邻调整。除 [贡献流程「角色与合并」](../../../CONTRIBUTING.md#角色与合并) 给切片角色的授权外，用户未明确要求时不 commit/push；创建 PR 的要求授权该 PR 必需的提交与推送，但不授权 merge、tag、release、branch protection 或生产 secret。审查者合并执行者切片 PR 的条件以贡献流程为准。

按目标仓库 [PR 模板](../../../.github/PULL_REQUEST_TEMPLATE.md) 填实际变更、验证与兼容性风险，保留字面标题；只勾选事实成立的确认项。新流程尚未合并时，不能称自动指派或 `pull_request_target` 检查已经生效。临时正文与探针放在忽略目录，不夹带到提交中。

## 创建并读取当前检查

有授权时推送专用分支并创建/更新 PR，不 force-push。使用可用的 GitHub 工具或已认证 `gh`，核对结果 URL 和当前 head SHA。可用只读命令：

```sh
gh pr checks PR_NUMBER --repo AlonaBot/AlonaBot
gh run view RUN_ID --repo AlonaBot/AlonaBot --json event,headSha,status,conclusion,jobs,url
gh run view RUN_ID --repo AlonaBot/AlonaBot --log-failed
```

替换已核实的编号；检查 pending、skipped、failed 与 success 的区别。CI 的 `Repository contract` 负责文档路由和技能入口；`Go quality` 仅在存在产品 Go 源码时运行。没有产品代码不等于 Go 验收通过。

## 修复 CI 的实际原因

先读取 workflow 和它执行的脚本，定位首个因果错误及 run/attempt/SHA。区分代码失败、环境缺失和基础设施故障；只有改动或明确的基础设施假设才值得重跑，不能重试掩盖确定性故障。

triage/审阅请求只操作 GitHub 元数据；保持 `pull_request_target` 不执行 PR 代码。`PR metadata` 将正文作为数据，不从标题、正文或评论构造 shell。产品检查保持只读权限、无生产凭据，不为了通过 CI 放宽依赖、停止预算或信任边界。

交付 PR URL、当前检查结果和仍影响合并的事项。只有确认了远端保护规则，才能说工作流是强制门禁；权限/API 错误与产品失败分别说明。
