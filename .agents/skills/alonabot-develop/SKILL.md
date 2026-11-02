---
name: alonabot-develop
description: 实现、诊断、重构或设计 AlonaBot 的 Go 代码时使用；根据现行工作包确定所有者、兼容性与最小变更，不启动新的批次或夹带仓库重构。
---

# 开发 AlonaBot

本技能将实际需求落实到现有 Core、SDK 或协议所有者；规则以仓库文档为准，不依赖个人技能库。

## 确定责任和完成条件

读取根 [AGENTS.md](../../../AGENTS.md)、[编码规范](../../../docs/coding-standards.md)、[架构地图](../../../ARCHITECTURE.md) 和 [PLAN](../../../spec/PLAN.md) §3.1 所指章节实施计划中对应的 T 工作包及失败停止条件。检查分支、基线、已有 diff、调用者与测试；Git 写操作使用独立 worktree，遵循 [贡献流程](../../../CONTRIBUTING.md) 的授权边界。

对 Bug 先追踪输入、责任层和可复现失败；对新行为明确使用者、可观察结果和兼容性影响。若契约冲突，先按 spec 的维护流程解决，不改测试或文档来迁就未接受的行为。

## 在一个所有者内完成修改

代码风格遵循 Google Go Style Guide（可用时加载 `go-google-style-guide`、`go-google-style-decisions`、`go-google-best-practices`），命名使用 [编码规范](../../../docs/coding-standards.md#命名与术语) 中的规格术语；只在信任边界做校验，不写防御性的重复检查。优先已有机制、具体类型和标准库。按职责放置类型、方法与常量；DTO 按主题共享，连接、锁和状态机留在实现所有者，不批量搬进统一 types 文件。新增跨包边时先核对架构表。

修改协议只编辑 `.proto` 或生成器。注意 Session 身份、就绪门禁、父调用权限、payload+error、停止预算、进程树与锁的顺序；这些机制变化时核查双方调用者及真实失败路径。既有入口的行为以契约为准，不把新入口的前置条件施加给旧入口；首次发布前改签名或语义时同步改全部调用方，不写迁移旧状态库或让旧调用方免改的兼容层（[编码规范](../../../docs/coding-standards.md#首次发布前不写兼容层)）。

本次触及的包文档、类型字段和函数阶段按 [注释技能](../alonabot-code-commenting/SKILL.md) 补齐。新依赖或安装工具须先说明必要性、替代方案并取得明确批准；新增限额、重试或降级须有现行契约或可复现故障依据。

## 验证并交付

按 [测试技能](../alonabot-test/SKILL.md) 选择最小有效证据，同步受影响的契约、调用者、fixture 和文档。跨模块大改动独立审查，按 [审查技能](../alonabot-review/SKILL.md) 检查最终 diff。

交付说明位置、实际行为、命令结果和未运行项。实现完成不等于已获 commit、push、PR 或 merge 授权；创建 PR 时再使用 [PR/CI 技能](../alonabot-pr/SKILL.md)。
