# 适配器实施计划

本章节覆盖各平台的官方 Adapter 实现，每个平台一个子章节。总目标、章节划分与跨章节依赖见 [PLAN](../../PLAN.md)；goal 语句见 [Telegram](../../../docs/goals/adapters/telegram.md) 与 [OneBot 11](../../../docs/goals/adapters/onebot11.md)。

| 子章节 | 程序 | 计划 |
| --- | --- | --- |
| Telegram | `adapters/telegram/` | [telegram.md](telegram.md) |
| OneBot 11 | `adapters/onebot11/` | [onebot11.md](onebot11.md) |

## 1. 边界

- 本章节只写平台实现：平台 API 客户端、接收与标准化、发送计划、平台取件、平台权限、本地模拟服务。Adapter 的接口层面在[核心实施计划](../core.md)：`sdk/adapter`、`internal/core/adapter` 与 `remote`、`adapter.*` 与 `core.adapter.*` 服务、能力与 extension 声明的校验、实例启停与游标。
- 适配器切片只依赖核心切片与本平台的切片，不依赖其他平台；前置一律写切片编号。核心功能需要真实平台验收时，用含平台部分的核心切片完成（见各子章节的「核心切片中的平台部分」），不在适配器章节重复实现 Core 功能。
- 跨 Adapter 一致性验收（T-15.3）在 OneBot 11 子章节，对照两个平台的同一组 Chain 与入站消息。

## 2. v0.1 的范围与扩展约定

v0.1 按各子章节已列的切片交付。2026-11-09 维护者要求补齐与 AstrBot 适配器的功能差距（D-87），对标的切片已追加到核心与两个子章节；除此之外不扩充平台功能，CONTRACTS §0 仍把 Telegram webhook 接收与其他平台排除在 v0.1 外。v0.1 只要求两个 Adapter 具备继续开发的结构：

- 实现放在可导入的包中（`telegram.New()`、`onebot11.New()`），程序入口只调用 `sdk.RunAdapter`，库模式可以直接引用；实例配置的 schema 写在包内的 `instance_schema.json`（D-85）。
- 每个平台有可控故障的本地模拟服务（`telegramtest`、`onebottest`），新功能的测试在其上追加。
- 平台权限与扩展事件放在常量包（`adapters/<id>/perm`、`adapters/<id>/event`），能力与 extension 在 Adapter 描述中声明，新增时只追加。
- 真实平台的验证单独记录；没有测试账号时记为未运行，不用模拟服务的结果代替。

v0.1 之后的平台功能在对应子章节追加切片，编号沿用该平台的工作包（Telegram 为 T-03.N，OneBot 11 为 T-15.N），由维护者确认范围后写入；需要改协议或 Core 时，先在核心章节追加切片。新增其他平台的官方 Adapter 时新开子章节。
