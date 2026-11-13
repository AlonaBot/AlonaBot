# 实施计划：插件

总目标、章节划分与跨章节依赖见 [PLAN](../PLAN.md)；本章节的 goal 语句见 [plugins.md](../../docs/goals/plugins.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本章节覆盖预装插件 `plugins/help/`、`plugins/welcome/` 与 `plugins/rss/`（T-14 步骤 1、2）。`examples/` 下的示例插件用于验收 SDK，随对应的核心切片交付，不在本章节。

- v0.1 交付一个成熟插件：rss。它用到调度、受管 SQLite、配置、文案与后台发送，能在真实使用中检验 SDK；其余两个预装插件按 T-14 的要求交付。
- 插件的打包、预装与首次运行在[集成与发布实施计划](release.md)。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-14.1 | goal-71 | help 插件与 `core.command.list`，复用 T-03.17 按 D-138 交付的说明字段、路径覆盖键与翻译选择 | T-14 步骤 1 | T-04.4、T-05.2、T-08.8 | V-67 的 help 分项 | 未开始 |
| T-14.2 | goal-72 | welcome 与 rss 插件 | T-14 步骤 2 | T-05.2、T-07.2、T-08.5、T-08.8、T-08.11、T-09.4 | V-67 | 未开始 |

## 3. 工作包

### T-14 的预装插件步骤

以下步骤原属 T-14，编号保持不变。T-14 的前置输入、拟目录、验收（V-67）与失败停止条件见[集成与发布实施计划的 T-14](release.md#t-14官方整合包与预装组件)。

1. **help 插件与 `core.command.list`。** Core 按调用者 Actor 与默认 Target 返回可用的命令：已启用、权限满足、未被遮蔽；说明复用 T-03.17 按 D-138 交付的 CommandSpec.description 英文原文与 commands 规范路径覆盖，按 Actor locale、插件默认 locale、英文原文选择；本切片不再新增说明字段或 SDK Description。`plugins/help/` 注册 `/help [命令]`，列出全部命令或单个命令的用法；`/whoami` 回复调用者的用户字符串 `<adapter_id>:<external_user_id>` 与收到消息的实例名，用户字符串可以直接写进 `admins`，用于添加第一个 Core Admin（CONTRACTS §15.2，D-105）。

2. **welcome 与 rss。**
