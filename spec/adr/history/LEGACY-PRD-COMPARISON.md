> 维护归档：此处保留 2026-11-20 的原始记录。现行分支、派发与授权规则见根 [AGENTS.md](../../../AGENTS.md)；重新派发前须核对当前 main。

# 旧版 PRD 与当前 AlonaBot 设计的逐项对比

> 历史审计记录（2026-10-19）。文中的“待定/需补齐/当前版本”描述保留当时状态；现行处置见 [DECISIONS.md](../DECISIONS.md) 的 P/Q 处置表 与 [SELF-REVIEW.md](SELF-REVIEW.md)。本轮未重新取得旧 PDF 原件。

**比较日期：** 2026-10-19\
**旧版：** `Go_WASM_Bot_Framework_Enterprise_PRD_v1.0.pdf`，2026-10-11，32 页。\
**新版：** `PLAN.md` v0.4-draft、`DESIGN.md`、`DECISIONS.md`。
**重要规则：** 旧版 PRD 本身是 Draft，历史上的“建议 / P1 / P2”并不因对比而自动升格为用户已批准的 V0.1 要求；较新的 Native-only 等明确决定优先。

## 一、FR-01～FR-30 完整差异映射

“承接”代表保留目标，“调整”代表改变实现或阶段，“替代”代表旧方案不得重新采用，“后置”代表不进入当前 V0.1 首版。

| 旧 ID | 旧版要求及页码 | 对比结论 | 在新规格中的去向 |
| --- | --- | --- | --- |
| FR-01 | Base 插件自动发现、一个包多插件类（第 9～10 页） | **调整** | 一个包只允许一个根插件类；多个功能 Handler 分文件/包组织；构建期检查 |
| FR-02 | Commands/Events 集中注册（第 10 页） | 承接 | DESIGN §D |
| FR-03 | 分散式方法注解 + 静态生成（第 10～11 页） | **调整** | 标注的包级独立函数可自动发现，结构体方法显式注册；不沿用旧版仅方法可注解限制 |
| FR-04 | Handler 支持 error / (Chain,error)（第 12 页） | **需补齐** | DESIGN §E 中补入两种签名的拟议契约；最终 SDK 签名待冻结 |
| FR-05 | MessageChain 片段（第 12 页） | **需补齐** | DESIGN §E 中补齐片段、平台渲染、媒体能力 |
| FR-06 | Send 主动/跨会话、多次 Send 与回执（第 12 页） | **需补齐** | DESIGN §E；默认目标与发送回执语义继续细化 |
| FR-07 | Event Type、Matcher、权限路由（第 11～12 页） | 承接并扩展 | DESIGN §E/F/H：补回 edited/deleted/button/member/reaction/scheduled 事件与无 Actor 不默认 Admin；融合 AstrBot Hook 与 NoneBot2 Rule |
| FR-08 | WASM describe/invoke/Host ABI（第 13～14 页） | **替代** | Native 子进程 + Length-Prefixed Protobuf，DESIGN §C |
| FR-09 | WASM 热加载/卸载（第 13～14 页） | **替代** | Native 插件可执行程序安装/重启/卸载；Process Manager |
| FR-10 | 单 WASM 实例串行调度（第 14 页） | **替代** | Native 插件内部 goroutine 并发，Core 保留内部背压；不增加逐插件 UI 并发开关 |
| FR-11 | SQLite Host SQL Proxy 支持 GORM（第 15～16 页） | **替代** | 原生数据库驱动/GORM；SQLite 与 PostgreSQL；去掉 Host SQL Proxy |
| FR-12 | JSON 配置仅在加载/重载时应用（第 17 页） | 承接 | DESIGN §J 的 Active/Pending 激活模型 |
| FR-13 | OnLoad 迁移 + 默认清理未知键（第 18、29 页） | **部分替代** | 保留 OnLoad/Validate，但**最新决策为默认保留未知键**；清理只能明确选择 |
| FR-14 | Go Struct 生成 JSON/UI Schema（第 17、29 页） | 承接 | DESIGN §J；校验 Schema 与 WebUI 展示信息分别生成的命名待定 |
| FR-15 | 更新预检查与激活前回滚（第 21～22 页） | 承接 | PLAN T-09、DESIGN §K；补状态机 |
| FR-16 | 版本状态/数据库快照恢复（第 16、21～22、30 页） | **需补齐** | Generation + Journal + SQLite 一致快照；Postgres 迁移恢复另验 |
| FR-17 | Core 调度器（第 14 页） | 承接并扩展 | DESIGN §H；不同定时任务可以并发 |
| FR-18 | Host FS 能力代理（第 18～19 页） | **替代安全假设** | Native 原生文件访问无法由普通 Core RPC 限制；可提供便捷 SDK，但非隔离；另列安全风险 |
| FR-19 | Host HTTP 出网代理（第 19 页） | **替代安全假设** | Native 可直接 net/http；不可承诺强制代理、SSRF 防护或网络隔离 |
| FR-20 | Host Exec 白名单执行（第 19 页） | **替代安全假设** | Native 可直接启动进程；Core 可管理自己的子进程，但无操作系统隔离时无法强制插件白名单 |
| FR-21 | Adapter ID/Start/Stop/Send/Capabilities（第 20 页） | **需补齐** | DESIGN §M 的 Adapter 参考接口 |
| FR-22 | Telegram、OneBot、Discord、QQ 原生接入（第 20 页） | **阶段调整** | V0.1 先 Telegram；其他 Adapter 后续 |
| FR-23 | 版本化专属能力扩展（第 20 页） | **需补齐** | DESIGN §M：capability 版本、平台 media group 等 |
| FR-24 | plugin.yaml 手写 id + CLI 打包构建（第 21、30 页） | 承接并调整 | Native 按平台预构建优先；CLI/check/schema 细节待定 |
| FR-25 | GitHub Releases + index、固定版本（第 21、30 页） | 承接为候选 | 以现有插件市场/来源模型为目标，暂不扩张为完整商业市场 |
| FR-26 | 配置/权限审计与结构化日志（第 22 页） | **需补齐** | DESIGN §O 的关联字段/审计事件；不新增复杂用户配置 |
| FR-27 | RSS/PSS/队列/时延/错误率（第 22～23 页） | **需补齐** | PLAN T-10、NOTES 的基准；区分 Windows/macOS 的可得指标 |
| FR-28 | 动态选项/依赖显隐/多级复杂表单（第 7 页） | **后置** | 维持最小配置 UI，后续再决定是否需要 |
| FR-29 | Blue-Green 升级（第 7、21 页） | **后置** | 首发采用低内存切换，明确非零停机；不强制双进程并存 |
| FR-30 | WASM 插件实例池（第 7、14 页） | **替代** | Native 进程内允许并发 Handler，不引入多进程副本池 |

**旧版优先级仅作历史参考**：PRD 原标记 P0 共 16 项、P1 共 11 项、P2 共 3 项（FR-28～30）。Native-only 决策替换了部分旧 P0 方案；不能直接把原来的优先级当成当前产品的范围承诺。

## 二、非功能目标与验收基线

旧版第 4～7、22～26 页包含成功目标、测试门槛和可观测性清单，当前新版原先只有宽泛验收。本轮恢复以下适用于 Native 的**候选验收维度**：

- 可用性：重复启动/停止、崩溃重建、临界时刻更新与 Journal 恢复。
- 兼容性：Windows/Linux/macOS 的 Native 程序与双 IPC 传输一致性；五个首发 OS/arch 组合。
- 消息行为：MessageChain 的平台能力降级、按钮、相册、跨会话主动发送及回执 ID。
- 数据与配置：SQLite/PostgreSQL CRUD/Tx/迁移、Pending 配置验证、secret 遮蔽、未知键策略。
- 性能：Core + 真实**不同插件二进制** + 外部 sidecar 的 RSS/私有内存、CPU、P95 与队列行为；对照等价 NoneBot2 工作负载。
- 注入测试：下载失败、签名/来源拒绝、插件启动失败、数据库迁移失败、Core 更新时中断、服务调用超时/崩溃。
- 运行与观测：pluginID、routeID、callID、session/generation 等关联字段；消息发送与命令权限修改保留必要审计记录。

旧文中的「100 虚拟用户」「100 次启停」「20 次失败升级」为**参考测试场景**，不表示已经测量或被冻结为绝对发布指标。

## 三、旧版保留的行为细节

1. **Chain ≠ 多条 Send**（第 12 页）：Chain 表示一条逻辑消息的有序片段；只有多次 Send 才表示多条独立发信。没有默认目标时需要明确错误。
2. **跨平台事件字段**（第 11～12 页）：event_id、event_type、adapter_instance、session、actor、timestamp、payload、capabilities，并区分缺失字段。
3. **Send 回执**（第 12 页）：平台 API 的成功接收与最终用户送达是两回事；建议提供 SendResult/平台消息 ID，不将仅入队误记为成功。
4. **Adapter 扩展**（第 20 页）：接口相对稳定，平台特有能力通过 versioned capability 扩展，不强行通用化所有功能。
5. **配置生效**（第 17～18 页）：Draft/Pending 不直接驱动运行中的插件；只有验证通过的 Candidate 才能成为 Active。
6. **更新提交**（第 21～22、30 页）：安装预检 → 停止新派发 → 候选状态/一致快照 → 启动新程序/健康检查 → Journal 提交 → 原子切换；Core 崩溃可以根据 Journal 决定恢复旧版本或新版本。该流程需验证，不能假设数据库支持任何情况的事务式回滚。
7. **Builder 单一真源**（第 17、21 页）：Go Config Struct/Handler 定义生成对应 schema、路由描述和检查产物，减少手写重复字段。

## 四、必须明确废弃或改写的旧假设

- **Standard Go WASM/wazero Reactor、单实例串行 Handler、WASM Host ABI、Host SQL Driver**：全部由 Native 进程+Protobuf IPC+原生 GORM 取代。
- **单个插件包多个根类、只允许在实例方法上注解**：被“单根插件类、可分拆功能 Handler、独立函数自动注册、实例方法显式注册”取代。
- **plugin.yaml 的 community.* 示例 ID**：不满足当前市场 alonabot_plugin_ 命名约束（可作为不规范本地旧例子），不能用于正式示例。
- **默认未知键清理**：与最新讨论冲突；采用默认保留未知键，不自动删除管理员/旧版本留下的数据。
- **强制 Host 文件/HTTP/Exec 权限审查**：Native 进程可绕过普通 SDK。除非以后加入真正的 OS 级隔离，否则这只是应用层约定，不能做安全承诺。
- **“重试未送达事件 + Outbox”作为通用事件行为**：与已确认 Event Bus 尽力投递、业务 Service Call 不自动重试冲突。只可在未来单独定义受控消息发送去重/可靠出站，不能暗中改变 Event Bus 语义。
- **不同平台只运行同一个 .wasm**：Native 模式要分别构建 OS/arch 产物；同版本不等于同一可执行文件。
- **更新失败一律无损回滚**：只能对明确的激活前候选状态和已确认快照做恢复保证；激活后有副作用的操作和某些 DDL 不可保证。

## 五、现阶段仍未完成的设计

这次比对**不会自动决定**事件相同优先级排序、复杂 Matcher、同一调度任务重复触发、完整 proto envelope、日志级别配置、新版 CLI 语法、Adapter 外置插件方式以及 Native 执行权限模型。

后续继续讨论这些点；不要把旧版未批准的建议移入 DECISIONS.md 的“已确认”表。

## 六、版本沿革与资料位置
- 本次自审见 [SELF-REVIEW.md](SELF-REVIEW.md)：确认旧 PRD 的 30 个 FR 编号均已进入映射，但旧 PDF 的内容不自动等价于已接受的新版接口或安全实现。

- 旧版 PDF：`Go_WASM_Bot_Framework_Enterprise_PRD_v1.0.pdf`（2026-10-11；32 页；之前的聊天附件，**未复制进此 Mac 工作区目录**）。
- 当前主计划：`spec/PLAN.md` v0.2-draft。
- 设计正文：`spec/DESIGN.md`；决策与候选：`spec/adr/DECISIONS.md`；测试记录：`spec/adr/history/NOTES.md`。

本文件是**需求迁移与差异记录**，不是证明所有旧 SDK 示例已在 Native Go 中编译通过。
