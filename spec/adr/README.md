# ADR：决策与历史依据

**技术基线：** v0.6-review  
**整理日期：** 2026-10-22

本目录保存选择原因、决策处置、历史审阅和技术来源。直接实施请打开根目录的 [PLAN.md](../PLAN.md)；精确行为由 [CONTRACTS-v0.1.md](../CONTRACTS-v0.1.md)、[WIRE-FORMAT.md](../WIRE-FORMAT.md) 和 [MESSAGE-MODEL.md](../MESSAGE-MODEL.md) 定义。

## 1. 决策与记录索引

| 文件 | 内容与效力 |
| --- | --- |
| [DECISIONS.md](DECISIONS.md) | D-01～D-25 原决策，以及 P-01～P-14、Q-01～Q-18 的议题、处置、契约位置和状态 |
| [history/SELF-REVIEW.md](history/SELF-REVIEW.md) | 2026-10-21 可实施性审阅的 A-01～A-20 修复、历史 R 项与当时静态检查结果 |
| [history/NOTES.md](history/NOTES.md) | 历史 Native 原型、实验局限、技术主源与当时核对记录 |
| [history/LEGACY-PRD-COMPARISON.md](history/LEGACY-PRD-COMPARISON.md) | 旧 PRD 的 FR-01～FR-30 对照及历史方案冲突 |
| [history/POST-PRD-AUDIT.md](history/POST-PRD-AUDIT.md) | PDF 后 42 项已归档讨论的覆盖记录及边界 |

历史文档保留当时版本和描述。旧记录中的“待批/需补齐”不能覆盖现行契约，也不能当作当前实施已通过；需要查当前处置时先读 DECISIONS 的 P/Q 表。真实测试状态统一记录在 [TEST-MATRIX.md](../TEST-MATRIX.md)。

## 2. 维护规则

D 编号及原已确认内容保留；后续改变某项行为时追加有日期的记录，说明背景、选择、被替代的规则、影响及对应任务/验收，并同步根目录契约。P/Q 表记录历史议题的去向，已完成讨论不再作为根目录实施待办重复出现。现阶段保留一份集中决策表，避免复制成多份相互冲突的记录。

## 3. 本次结构与计划整理

- 根目录保留 README、PLAN、DESIGN、CONTRACTS、WIRE、MESSAGE 和 TEST-MATRIX 七份直接实施资料；五份决策/历史文件迁入本目录。
- 原 CONTRACTS §11 的 P/Q 表合并到 DECISIONS；原章节号保留为实施契约的交付映射，§12 及其引用不改号。
- PLAN 展开 T-01～T-10，每项具备前置输入、有序步骤、拟建目录、交付物、V 验收和失败停止条件。
- 跨阶段用例指定最终闭合点：命令策略与 WebUI、Hook 与 Service、基础状态/候选资源与正式 Journal、HTTP 授权组件与首次管理登录分别有明确交接。
- 本次是文档组织和执行细节补充；技术基线仍为 v0.6-review，V-01～V-47 仍未运行。

## 4. 原 PLAN 修订记录

以下从重整前的主计划移入，保留当时版本与范围。

- **v0.6-review 同版修订 / 2026-10-21**：按用户授权完成剩余细节并审阅可实施性；修复候选 DB/secret 污染、崩溃双进程、Bootstrap/控制失败、Handler 传播与部分发送契约；补齐迁移、命令权限、HTTP Origin、安装清单与阶段依赖。版本编号不变，运行时测试仍未执行。

- **v0.6-review / 2026-10-21**：修复 v0.5 误插入风险表格的段落；收敛 CandidatePrepared/Active、数据库回退边界、双传输协议/SDK 与配置重启一致性；增加独立验收矩阵，仍未实施。
- **v0.5-review / 2026-10-21**：新增 v0.1 待批技术契约，历史待办均有初步取舍和测试方向。

- **v0.4-draft / 2026-10-19**：开展全 PDF + PDF 后设计追溯自审，纠正事件类型/定时任务身份、动态命令路由调用、配置包装层和生命周期阶段的覆盖缺口；新增 history/SELF-REVIEW.md。

- **v0.3-draft / 2026-10-19**：针对 PDF 之后的长时设计讨论进行了独立覆盖审计并修订 SDK/冲突行为、可选互斥 DB 迁移、配置 i18 及未来功能边界；详情见 history/POST-PRD-AUDIT.md。

- **v0.2-draft / 2026-10-19**：依据 2026-10-11 旧版 32 页 PRD 逐条核对 FR-01～FR-30，补齐消息/Adapter、配置激活、更新 Journal、量化验收与历史方案冲突说明；不重新引入 WASM。

- **v0.1-draft / 2026-10-19**：首次归档目前所有明确决策；从 WASM 转为纯 Native 子进程，明确双 IPC、Protobuf、权限、事件管道与跨插件协作；实施和真实集成测试均未开始。
