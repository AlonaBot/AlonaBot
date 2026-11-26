---
name: alonabot-docs
description: 修改 AlonaBot 文档、AGENTS/CLAUDE 或仓库技能时使用；维护单一权威、准确状态和干净检出可读取的相对引用，不扩展为产品实现。
---

# 维护 AlonaBot 文档与技能

本技能负责文档归属、代理路由和说明准确性。先读受影响的实现、协议或工作流，再修改其现有权威文档。

## 选择唯一所有者

- 协议字段与消息结构：[Protobuf 定义](../../../contracts/proto/alonabot/v1)；文档只解释，不另立契约。
- 整体设计与设计理由：[设计导读](../../../docs/design.zh-CN.md)，核心机制变化时同步。
- 模块边界、依赖方向与状态归属：[ARCHITECTURE.md](../../../ARCHITECTURE.md)。
- Go 风格、注释与测试提交：[编码规范](../../../docs/coding-standards.zh-CN.md)。
- 工具、命令、验证与 CI：[源码开发](../../../docs/development.zh-CN.md)。
- 组件作者的 SDK 用法：[插件指南](../../../docs/plugin-guide.zh-CN.md)；Agent 契约：[Agent API](../../../docs/agent-api.zh-CN.md)。
- 打包与发布：[发行维护](../../../docs/releasing.md)。
- 贡献流程：[CONTRIBUTING](../../../CONTRIBUTING.zh-CN.md)；版本变化：[changelog](../../../changelog/README.zh-CN.md)。
- 代理入口：根 [AGENTS.md](../../../AGENTS.md) 只放路由与必要红线，`CLAUDE.md` 保留薄引用；重复的项目流程写成 `.agents/skills/alonabot-*/SKILL.md`。
- 安装、使用和日常运维由文档站 [AlonaBot/alonabot.github.io](https://github.com/AlonaBot/alonabot.github.io) 维护，本仓库只链接。
- ADR、spec、goal 与计划等个人维护资料在 dev 分支，CI 拒绝它们进入 main；main 的文档不把 dev 列为必读。

## 保持说明可用

内部维护说明、注释、commit message 和仓库技能用简体中文；用户与开发指南提供中英两份，路径相同、内容同步。标识符、命令和报错保留原文。技能的 `name` 与目录一致，`description` 写明触发条件；`agents/openai.yaml` 的提示引用 `$技能名`，不依赖维护者机器路径、全局技能或个人代理配置。

普通文档标题后的导语说明负责范围与更新时机。新增规则前先找现有所有者，不在多处重复。代码注释的写法见[注释技能](../alonabot-code-commenting/SKILL.md)。

状态来自当前分支与实际证据：不把未合并的实现写成 main 已有，不把写好的 workflow 写成已经执行的检查。公开声明保留事实、版本、平台与单位，不为文案整齐弱化行为或丢掉已知失败。

## 完成检查

运行 `git diff --check`，核对技能 frontmatter 与 YAML、相对链接及锚点，中英两份同步修改。缺少校验工具时先报告，不自动安装依赖。文档任务不启动产品实现、发布或整库重构。
