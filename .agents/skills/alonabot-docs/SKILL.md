---
name: alonabot-docs
description: 修改 AlonaBot 文档、AGENTS/CLAUDE 或仓库技能时使用；维护单一权威、准确状态和干净 checkout 可读取的相对引用，不扩展为产品实现。
---

# 维护 AlonaBot 文档与技能

本技能负责文档归属、agent 路由和说明准确性。先读受影响的实现、规格或工作流，再修改其现有权威文档。

## 选择唯一所有者

- 行为契约及 T/V 编号：`spec/`；契约修改按 [spec 维护流程](../../../spec/README.md) 同步 PLAN、TEST-MATRIX 与 ADR。
- 模块边界：[ARCHITECTURE.md](../../../ARCHITECTURE.md)。
- 文件组织、Go 与注释：[coding-standards.md](../../../docs/coding-standards.md)。
- 工具、命令、CI 与环境证据：[development.md](../../../docs/development.md)。
- 贡献与 Git 授权：[CONTRIBUTING.md](../../../CONTRIBUTING.md)。
- 任务入口：根 [AGENTS.md](../../../AGENTS.md)；`CLAUDE.md` 保留薄引用。
- 重复的项目流程：`.agents/skills/alonabot-*/SKILL.md`；临时进度使用已有 goal 回写区或 PR/回复。

## 保持说明可用

文档、注释、commit message 和仓库技能用简体中文，标识符/命令/报错保留原文。技能有与目录一致的 `name` 和明确触发条件的 `description`；元数据提示引用 `$技能名`，不依赖维护者机器路径、全局技能或个人代理配置。

技能使用必需的 YAML frontmatter，其后正文说明用途；普通文档标题后的导语说明权威范围与更新时机。新增重复规则前先寻找现有所有者，AGENTS 只放路由与必要红线。代码的包文档、注释与 step 注释的写法读取 [注释技能](../alonabot-code-commenting/SKILL.md)。

状态来自当前分支与实际证据，不能把未合入分支的实现写成 main 已实现，也不能把写好的 workflow 写成已经执行的检查。公开声明保留事实、版本、平台与单位；不为了文案整齐弱化契约或丢失已知失败。

## 完成检查

运行 `node scripts/check-repository.mjs` 和 `git diff --check`，检查技能 frontmatter/UI YAML、相对链接及相关锚点；检查器不联网、不验证锚点或注释语义，必要时人工核对。技能可用现有校验器验证，缺工具先报告，不自动安装依赖。普通文档任务不启动产品实现、发布或整库重构。
