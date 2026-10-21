# 设计决策记录

版本：v0.6-review（2026-10-21；保留已确认 D-01～D-25，不升版本）

D 表保留历史已确认方向；P/Q 在本文件集中记录议题、处置及当前状态。2026-10-21 的规格完善任务将剩余技术默认值收敛为实施基线，不追加虚构的历史确认记录；2026-10-22 将决策和历史记录归入 adr/。

**实施入口**：[PLAN.md](../PLAN.md) 给出 T-01～T-10 详细工作包；[CONTRACTS-v0.1.md](../CONTRACTS-v0.1.md) 固定行为，表内 § 编号均指该契约；验收见 [TEST-MATRIX.md](../TEST-MATRIX.md)。P/Q 处置表已从根目录 CONTRACTS 合并到本文件。D-07/D-18 的历史“Ready”统一解释为新状态机中完成 OnReady 并被 Core 接受的 `Activated`，不是 `Prepared`。

## 已确认

| ID | 决策 | 说明 |
| --- | --- | --- |
| D-01 | 只支持 Native Go 插件 | 每个插件包独立子进程；取消 WASM 运行时/Host ABI/SQL Proxy 设计；允许普通 goroutine、net/http、GORM |
| D-02 | 预构建与源码构建并存 | **预构建优先**，市场按 OS/arch 提供匹配产物；本地源码构建可选 |
| D-03 | 两种 IPC 传输方式 | 本地 TCP RPC 默认，stdio 可选；两者共享相同数据协议，SDK 封装实现 |
| D-04 | IPC 编码 | Length-Prefixed Protobuf；CallID、Session ID；可并发/乱序响应；大媒体不直接塞入常规 RPC |
| D-05 | 插件日志可归因 | Core 捕获 stdout/stderr 并标记插件名；stdio 模式 stdout 专用 IPC，普通日志重定向 stderr；原始 FD 写入仍有风险 |
| D-06 | 进程恢复 | Core 启动、监控、终止、有限自动重启；按此前接受的 5 分钟内连续崩溃 3 次停止自动重启；业务调用不自动重试 |
| D-07 | 注册静态、策略动态 | Handler 注册表在插件启动 Ready 阶段固定；权限、启用状态由 Core 动态更新 |
| D-08 | 命令权限 | 插件/命令组/子命令的权限继承及覆盖由 Core 执行，支持运行时修改并持久化；不得越过全局安全规则 |
| D-09 | 不开放每插件并发上限配置 | Native Handler、不同定时任务可并发；goroutine 开发者自行管理；Core 仍有内部保护，不增加 WebUI 调参 |
| D-10 | 统一事件处理模型 | 融合 AstrBot 的 Hook/优先级和 NoneBot2 的 Rule/Permission/停止传播；不引入额外复杂 Matcher 系统 |
| D-11 | Event Bus | 跨插件异步事件，非持久化尽力投递，离线不补发；不同于有结果的调用 |
| D-12 | Service Call | 由 Core 路由，允许 A→B→C 嵌套，避免同步环；支持超时/取消，不自动重试有副作用的调用 |
| D-13 | Go SDK 注册风格 | 一个插件包一个根插件类（plugin.Base），最多一个 Config 类；独立函数可经注解自动发现，实例方法显式注册；命令组/子命令 |
| D-14 | 数据库与配置 | 首发 SQLite/PostgreSQL；原生 GORM；插件 JSON 配置、WebUI schema、secret/hidden/description、i18/；未知配置键默认保留 |
| D-15 | HTTP 路由 | Core 统一公共 HTTP 服务，插件路由按 ID 前缀；不可信插件 UI 不能与 Admin UI 共用浏览器 Origin |
| D-16 | 插件身份与更新来源 | `plugin.yaml` 的 ID 必须有且是 HTTP 路由前缀；市场发布需 `alonabot_plugin_` 前缀，本地命名不规范只警告；更新同时检查 ID + 原可信来源，避免同 ID 不同仓库混淆 |
| D-17 | 首发范围 | Telegram 与最小机器人主链路优先；OneBot/QQ 官方/Discord 等其他 Adapter 后续扩展 |
| D-18 | RPC 握手 | 身份、主/次版本与能力协商，主版本不兼容拒绝启动；只有 Ready 才激活路由 |
| D-19 | 调度归 Core | Core 管理定时触发；不同任务在同一个 Native 插件内可以并发，不应彼此串行阻塞 |
| D-20 | **重复冲突两类规则** | **不同插件包 ID 重复：first-wins，保留最先注册者、后者报告冲突，不可静默替换**。**同插件自动+显式 Handler / HTTP 路由重复：报错**，不再使用较早的 first-wins 路由规则 |
| D-21 | GORM 和迁移 | 提供统一 `gormdb.Open(...)` SDK 入口；可选择不迁移、自行实现 `Migrate()`、或由 Core 协助的 `Migrations()`，后两种互斥；跨数据库搬迁由开发者负责 |
| D-22 | 配置与翻译细节 | 每包最多一个 Config，多 Config 构建失败；`i18/` 下翻译 JSON 采用顶层 `config` 嵌套字段，不增加额外 i18n tag；secret/hidden、固定排序遵守最小配置原则 |
| D-23 | SDK 开发体验 | plugin.yaml 元数据 ID 是真源，根类无需重复写 `ID()`；独立函数可自动注册，实例方法显式注册；辅助 Handler 可显式拿到根类指针；优先原生 Go Modules 与少反射 |
| D-24 | 保留两条跨插件协作途径 | 公共接口注册到 Core → Service Call（推荐）；同时**保留**通过 Core 动态调用另一插件已注册命令/事件路由的需求，具体权限语义和是否进入 V0.1 尚待定，不能直接持有另一进程 Go 对象 |
| D-25 | 统一 HTTP 入口、插件端口由 SDK 托管 | **用户确认**：所有框架托管的插件 HTTP 路由只由 Core 的统一 HTTP(S) 入口对外提供；插件开发者只注册 Go `http.Handler`/路由，**不需要设置、管理或告知用户开放端口**。SDK 只在有 HTTP 路由时自动启动 loopback 随机端口的内部 Listener；Core 维护 Session 绑定的内部反向代理、路由/权限及启停清理。部署用户只需开放 Core 对外端口，插件内部 Listener 不映射公网；命令/事件/Service Call 仍使用 Protobuf RPC。此确认不代表 Native 插件绝不能自行创建框架外 Listener，也不改变插件 HTML/JS 与 Admin UI 不同浏览器 Origin 的既有安全要求。 |

## P 类历史议题与当前处置

| ID | 议题 | 当前处置与实施规则 | 契约位置 | 状态 |
| --- | --- | --- | --- | --- |
| P-01 | 调度重叠/漏触发 | 同任务跳过重叠，离线漏触发不补。 | §4 | 已收敛 |
| P-02 | 资源/超时常量 | Core 固定 Session、队列的数量/字节预算与分阶段时限，无逐插件调参 UI。 | §1/§2/§4 | 已收敛，数值需测试 |
| P-03 | 代次与恢复 | 独占锁与 Journal 唯一提交点约束恢复；旧 Session 失效，不自动重放业务调用。 | §1/§4/§6/§12 | 已收敛 |
| P-04 | 多轮会话 | v0.1 仅调用上下文；持久会话状态与暂停 DSL 后置 v0.2。 | §0/§9 | 明确后置 |
| P-05 | Handler 双形式 | 两种 Handler 签名共用 HandlerOutcome；自动回复与主动 Send 分明，一次 Chain 可产生多物理消息，手动模式可多次 Send。 | §8 | 已收敛 |
| P-06 | Adapter 形态 | Telegram 静态集成、能力版本化，实例可启停；类型热加载后置。 | §8 | 已收敛 |
| P-07 | 更新恢复 | Generation/Journal 管理受管 SQLite 候选副本；提交后显式恢复，PG/不可隔离迁移进入维护。 | §6 | 已收敛 |
| P-08 | 配置激活 | Pending→Candidate→Prepared→ActiveGeneration+COMMITTED→OnReady/Activated；secret 引用不可变，提交前失败启动旧版的新 Session。 | §5/§12 | 已收敛 |
| P-09 | 匹配/传播 | 普通命令最长路径优先，识别后拒绝不回落；同事件 priority 小先、同级稳定 ID 排序，Regex/Raw 与 Stop 语义固定。 | §2/§3 | 已收敛 |
| P-10 | Adapter 动态装载 | 已有 Adapter 实例开关首发；新类型热加载后置。 | §8 | 已收敛 |
| P-11 | JSON 包装层 | 运行配置为普通 JSON，无 config 包装；i18 翻译使用 config 包装。 | §5 | 已收敛 |
| P-12 | 事件与 Actor | 最低事件清单采用类型化 payload；scheduled 等系统任务没有用户角色或默认 Admin。 | §2/§4 | 已收敛 |
| P-13 | 生命周期 | OnStart/OnReady/OnStop 首发，配置回调独立 OnLoad/Validate；OnConfigChange 后置。 | §4/§5 | 已收敛 |
| P-14 | 多账号 | Target 明确 adapter/instance/account/chat/thread，凭证按实例隔离；不内置 X/Twitter 动态凭证。 | §8 | 已收敛 |

## Q 类历史问题与当前处置

| ID | 议题 | 当前处置与实施规则 | 契约位置 | 状态 |
| --- | --- | --- | --- | --- |
| Q-01 | RPC 精确格式 | WIRE-FORMAT 固定 Envelope、控制失败、Bootstrap、状态、错误与 deadline/Cancel。 | §1 | 已收敛 |
| Q-02 | Native 信任边界 | Native 插件按可信本地代码运行，不提供跨平台恶意代码沙箱。 | §7 | 已收敛；安全测试仍需做 |
| Q-03 | 真实多插件内存 | T-10 测量多个真实不同 Native 可执行文件的总资源；不能以相同 worker 的 PSS 预设结论。 | §10 | 性能数值只能实施期验证 |
| Q-04 | 五平台/CGO | T-01 锁定驱动/工具链，T-10 完成五平台真实构建与启动；CGO=0 优先，CGO 插件单独发布并记录条件。 | §7/§9 | 已收敛；需真实构建 |
| Q-05 | 事件 Hook/排序 | CONTRACTS 与 MESSAGE-MODEL 固定 priority、Rule 错误、StopPropagation、Hook 阶段、结果及允许修改范围。 | §2 | 已收敛 |
| Q-06 | 媒体 | Core 顺序分块上传，MediaRef 使用受限 token、TTL、配额和消费租约。 | §1.3 | 已收敛 |
| Q-07 | 版本/签名 | Core/protocol major.minor 与 SDK 编译版本区分；manifest、归档摘要、来源身份和外部 release 清单签名分别校验。 | §7/§9 | 已收敛 |
| Q-08 | Config/i18 | Go Schema/JSON tag/i18 嵌套对应原始 JSON；固定类型、迁移、默认、PATCH、未知键保留与翻译规则。 | §5 | 已收敛 |
| Q-09 | PostgreSQL | 管理员预置每插件专用 PG role/schema；SDK 执行迁移，遵守维护门槛，失败按人工恢复边界处理。 | §6 | 已收敛；需 PG 验证 |
| Q-10 | 第三方 Adapter | 首发官方静态 Telegram；第三方 Adapter 类型及装载机制后置。 | §8 | 明确后置 |
| Q-11 | 回执/片段 | MESSAGE-MODEL 定义两种 Handler 输出、物理发送计划、部分成功回执与不支持片段的明确失败。 | §8 | 已收敛 |
| Q-12 | CLI/分发 | 交付已承诺的开发/管理 CLI、tar.gz、manifest、外部 release 清单与 GitHub Release 来源规则。 | §9 | 已收敛 |
| Q-13 | Journal/快照 | Journal/Generation、SQLite Backup 候选、进程锁和持久化次序共同保证提交恢复；PG 不可逆变更拦截，恢复需故障注入。 | §6/§12 | 已收敛；恢复需故障注入 |
| Q-14 | SDK 生成 | 在插件目标构建范围内扫描本模块包级函数；单 root、自动/显式冲突报错、实例显式绑定、Schema 生成由 T-01/T-04 编译验证。 | §9 | 已收敛；编译器必须验证 |
| Q-15 | 动态命令调用 | v0.1 交付 Service Call/Event Bus；动态跨插件命令 Router 后置 v0.2，保留稳定 route ID 和需求。 | §3 | 明确后置（需求保留） |
| Q-16 | Reload/Ready | 配置变更通过新 Session 重启式应用；OnReady 完成并 Activated 后才接业务，OnStop 负责停止。 | §4 | 已收敛 |
| Q-17 | 配置包装 | 运行时 Config 直接使用 JSON 字段，无顶层包装。 | §5 | 已收敛 |
| Q-18 | 进程/密钥/账号 | 合作型进程锁/子进程树清理与实例凭证作用域有明确规则，需跨平台验证；Native 仍不是沙箱。 | §7/§8/§12 | 已收敛；安全测试仍需做 |

“已收敛”表示已有实施规则，后置需求仍保留；性能、驱动、生成器及故障恢复的结果须在 TEST-MATRIX 中取得实际证据。

## 旧版 PRD 冲突解决（现行方案优先）

- WASM-only、WASI Reactor、Host ABI/SQL Proxy、单实例串行：由 D-01/D-03/D-04/D-09 取代。
- 同包多个根插件类、方法注解必需：由 D-13 的单根类/独立函数注解/显式方法注册取代。
- 旧版默认 PruneUnknown=true：被 D-14 **默认保留未知键**取代，OnLoad/Validate 保留。
- 强制 Host FS/HTTP/Exec 权限拦截：Native 可绕开普通 SDK，无 OS 级隔离时不得宣称安全隔离。
- 通用事件自动补发/Outbox：Event Bus 按 D-11 尽力投递，业务 Service Call 不自动重试；可靠发信应单独定义。
- 同一个 .wasm 跨平台：改为同版本下按 OS/arch 分发不同 Native 可执行文件。

完整 FR-01～FR-30 迁移表见 [LEGACY-PRD-COMPARISON.md](history/LEGACY-PRD-COMPARISON.md)。

PDF **之后**的长时间设计讨论覆盖情况、仍未完整恢复的议题及冲突区分见 [POST-PRD-AUDIT.md](history/POST-PRD-AUDIT.md)。

## 决策变更记录

- 2026-10-19：建立正式设计记录；Native 子进程取代 WASM-only。
- 2026-10-19：确认预构建优先、源码构建可选；双传输 Protobuf、日志归属、动态命令权限与最小配置。
- 2026-10-19：确定 Event Bus 尽力投递、允许嵌套 Service Call、融合 AstrBot 与 NoneBot2 事件处理。
- 2026-10-19：v0.2 对照旧版 32 页 PRD，保留可迁移需求，显式标注过时的 WASM/Host 安全假设及仍待批准内容。
- 2026-10-19：v0.3 对 PDF 后的设计约束再次核查：纠正插件 ID first-wins 与同插件 Handler 冲突报错混淆；增补可选迁移接口、i18/ 文案结构、根插件与 Go 模块协作、未来方向。
- 2026-10-19：v0.4 自审补入旧 PDF 事件身份/权限基线、动态 Router 调用、关键生命周期阶段、配置文件包装层候选及多账号作用域；区分安全必需条件与未定实现细节。
- 2026-10-21：v0.5 新增 CONTRACTS-v0.1.md：整理**待批提案**中的配置重启、事件排序、RPC、服务权限、媒体、备份/更新、原生信任和跨平台验收。原 P/Q 状态保留，避免假定审批。
- 2026-10-21：**D-25 已确认**：框架托管的插件 HTTP 只注册 Handler，不管理端口；统一由 Core 公网入口反向代理到 SDK 自动管理的 loopback Listener。只更新这一条决定，不代表批准其他 v0.6 提案。


- 2026-10-21：按用户本轮“自动完善剩余细节、审阅可行性并修复缺陷”的授权，同版收敛 P/Q 与执行契约；D-01～D-25 历史内容保持，新增技术细节的来源就是本轮规格完善任务。
- 2026-10-22：按用户要求迁入 adr/；合并原 CONTRACTS §11 的 P/Q 处置表，根目录改为直接实施工作包与契约。D-01～D-25 原文和既有编号保持。
