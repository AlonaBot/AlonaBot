# AlonaBot 实施规格

**技术基线：** v0.6-review  
**目录更新：** 2026-10-22  
**当前状态：** T-01（契约与工具链）与 T-02（Core 状态、Native 进程、双 IPC、基础生命周期）以及日志基线（T-02.1）已实现，分项验收证据见 [TEST-MATRIX §3.1](TEST-MATRIX.md#31-已运行的分项证据t-01t-02)；调优参数（T-02.2）、T-03～T-14 的能力与全部跨任务验收尚未实施。

按 [PLAN.md §3.1](PLAN.md#31-执行切片) 的章节、各章节实施计划（[plan/](plan/core.md)）与 [docs/goals/](../docs/goals/README.md) 推进。依赖图中最早可推进的未实现切片是 T-02.2；T-02.1、T-07.1、T-08.2、T-09.2 与 T-12.1 已合并。实际派发仍须满足 docs/goals/ 的状态规则。T-03.1 需等待 T-02.2 合并。PLAN 给出任务依赖、切片划分、输入、操作步骤、交付物和完成条件。

## 根目录：当前实施所需资料

| 文件 | 执行时的用途 |
| --- | --- |
| [PLAN.md](PLAN.md) | 实施计划总纲：最终交付、工作包依赖与交付次序、章节划分（§3.1）、验收规则；§7 为 v0.1 之后的休眠扩展 S-01～S-06 |
| [plan/](plan/core.md) | 各章节的实施计划：核心、Agent、适配器（[总述](plan/adapters/README.md)、Telegram、OneBot 11）、插件、WebUI、集成与发布；每章有切片表与工作包 |
| [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) | API、权限、生命周期、配置、数据与安装的精确行为 |
| [WIRE-FORMAT.md](WIRE-FORMAT.md) | Protobuf 字段、枚举、控制帧、Bootstrap 与载荷定义 |
| [MESSAGE-MODEL.md](MESSAGE-MODEL.md) | Event/Target/Chain、发送回执、媒体和 Telegram 映射 |
| [DESIGN.md](DESIGN.md) | Core/SDK/Adapter 职责、模块关系与生命周期说明 |
| [TEST-MATRIX.md](TEST-MATRIX.md) | V-01～V-114 的场景、通过条件、阶段分工与证据要求 |

## 决策与历史依据

[adr/README.md](adr/README.md) 索引 D/P/Q 决策、历史 PRD 对照、审阅记录和技术来源。根目录维护当前要执行的内容；历史讨论及当时的检查结果保留在 ADR 中，并标明适用时间。

## 执行与维护

按 PLAN 的依赖实施，按 CONTRACTS/WIRE/MESSAGE 确认接口，再用 TEST-MATRIX 记录证据。跨任务用例须等最终集成部分也通过后才能关闭。本文所列代码目录是实施目标；不能把实验、接口草图或文档检查记成运行验证。

变更行为时同步契约、相关工作包和验收条目，并在 ADR 记录原因及影响；已有 T/V 与历史 D/P/Q 编号保持稳定。已确定的技术细节直接按基线实施。
