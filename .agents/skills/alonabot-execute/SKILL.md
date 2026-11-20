---
name: alonabot-execute
description: 从 dev 的维护 goal 推进 AlonaBot 产品任务；在 main 派生的独立 worktree 实现，维护回写只留 dev，按明确授权交付。
---

# 执行 AlonaBot 维护任务

先读根 [AGENTS.md](../../../AGENTS.md) 和 [goal 共同约束](../../../docs/goals/README.md#所有-goal-的共同约束)。历史切片语句再次使用前由审查者按当前 main 核对，完整任务通过派发消息或本地 `goal-N/input.md` 提供。

## 开始

1. 从最新 `origin/main` 创建独立 `goal-N` 或 `codex/<主题>` worktree，保留 dev worktree 作为个人规则与历史依据入口。
2. 在产品 worktree 读取当前架构、编码规范、相关契约、实现、调用者与已有 diff；维护资料在 dev 阅读。任务卡位于产品 worktree 的 ignored `goal-N/`。
3. 大型任务使用 goal-mode，`input.md` 保留完整目标与基线，`plan.md` 写依据、接口、验证及恢复，`tasks.md` 逐项列范围、行为、测试、完成条件与回写。小任务直接维护简短清单。

## 实现与验证

每轮完成一个连贯任务，按[开发技能](../alonabot-develop/SKILL.md)、[测试技能](../alonabot-test/SKILL.md)和[注释技能](../alonabot-code-commenting/SKILL.md)修改产品与公共文档。覆盖所有授权行为与必要相邻调用方；契约冲突、未批准依赖或影响目标的范围变化交由维护者决定，继续独立部分。

测试按实际风险选择，提交范围以编码规范为准。fixture 与交叉构建分别记录；未运行项写原因。每个任务回写实际变更、命令、平台、结果、剩余问题与下一步，任务卡留在本机。

## 交付

用户未明确授权时，交付可审阅的本地 diff。获准创建 PR 后提交、推送产品分支，按[PR 模板](../../../.github/PULL_REQUEST_TEMPLATE.md)向 main 开 PR，写明验证与公共文档同步点。每 task 单独 commit 仅在已有提交授权时执行。执行者按审查意见修正，回归写在本地；合并按 AGENTS 的现行授权规则处理。

PLAN、TEST-MATRIX、goal 状态和 ADR 在 dev 更新，单独记录产品提交/PR 及证据来源。产品 PR 排除这些路径与 Agent 鞍具；同步 main 到 dev 时保留维护覆盖层。
