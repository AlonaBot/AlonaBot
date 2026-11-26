---
name: alonabot-develop
description: 实现、诊断、重构或设计 AlonaBot 的 Go 代码时使用；先确定所有者、调用者与兼容性，再做最小而完整的修改，不夹带无关重构。
---

# 开发 AlonaBot

本技能把需求落实到现有 Core、SDK 或协议的所有者。规则以仓库公共文档为准：[AGENTS.md](../../../AGENTS.md)、[架构地图](../../../ARCHITECTURE.md)、[编码规范](../../../docs/coding-standards.zh-CN.md)、[源码开发](../../../docs/development.zh-CN.md)与[贡献流程](../../../CONTRIBUTING.zh-CN.md)。

## 确定责任和完成条件

读取需求、受影响实现、调用者、已有测试与当前 diff，确认分支和基线；Git 写操作使用独立 worktree。用[架构地图](../../../ARCHITECTURE.md)确定负责的模块和允许的 import 边，用 [Protobuf 定义](../../../contracts/proto/alonabot/v1)与各指南确认现行行为。

对 Bug 先追踪输入、责任层和可复现失败；对新行为明确使用者、可观察结果和兼容性影响。现行行为与需求冲突时先向维护者说明，不改测试或文档迁就未接受的行为。

## 在一个所有者内完成修改

代码风格遵循 Google Go Style Guide（可用时加载 `go-google-style-guide`、`go-google-style-decisions`、`go-google-best-practices`），命名使用[编码规范](../../../docs/coding-standards.zh-CN.md#命名与术语)中的术语；只在信任边界做校验，不写防御性的重复检查。优先已有机制、具体类型和标准库。类型、方法与常量按职责就近放置，连接、锁和状态机留在实现所有者。新增跨包边时先核对架构表，并同步修改它。

修改协议只编辑 `.proto` 或生成器，再运行 `go generate ./...`。注意 Session 身份、就绪门禁、父调用权限、payload 与 error、停止预算、进程树与锁的顺序；这些机制变化时核查双方调用者及真实失败路径。首次发布前改签名或语义时同步改全部调用方，不写兼容层（[编码规范](../../../docs/coding-standards.zh-CN.md#首次发布前不写兼容层)）。

本次触及的包文档、类型字段和函数阶段按[注释技能](../alonabot-code-commenting/SKILL.md)补齐。新依赖须先说明必要性与替代方案并取得维护者批准；新增限额、重试或降级须有现行行为或可复现故障依据，可调参数登记在 `internal/tune`。

## 验证并交付

按[测试技能](../alonabot-test/SKILL.md)选择最小有效证据，同步受影响的调用者、fixture、示例和文档。跨模块大改动独立审查，按[审查技能](../alonabot-review/SKILL.md)检查最终 diff。

交付说明位置、实际行为、命令结果和未运行项。实现完成不等于已获 commit、push、PR 或 merge 授权；创建 PR 时再使用 [PR 技能](../alonabot-pr/SKILL.md)。
