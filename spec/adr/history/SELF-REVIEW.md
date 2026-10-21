# AlonaBot 规格可实施性审阅

> 历史记录：本文保存 2026-10-21 及更早的审阅、来源或实验结果；当时的文件数、链接数和“本轮”均指该次记录。2026-10-22 迁入 adr/history/，相对引用已更新；当前执行步骤见 [PLAN.md](../../PLAN.md)。

**版本：** v0.6-review · 2026-10-21  
**性质：** 对现有规格补齐实施细节、检查跨章一致性并修复缺陷。技术选择由用户本轮“自动完善剩余细节、审阅可行性并修复缺陷”的授权确定；保留版本和既有 D-01～D-25。  
**结论：** Native Go 路线已有可进入实施的契约与任务依赖。本轮发现的启动、数据恢复、权限和交付缺口已在规格中修正；运行可行性仍须通过代码、平台与故障注入验证，不以文档审阅代替。

执行入口为 [PLAN.md](../../PLAN.md)，行为基线为 [CONTRACTS-v0.1.md](../../CONTRACTS-v0.1.md) 及 Wire/消息专章。本报告记录问题和证据，不建立第二份实施计划。

## 1. 本轮材料与边界

| 材料 | 实际处理 | 证据边界 |
| --- | --- | --- |
| `/Users/flanchan/alonabot/spec` 的 12 份 Markdown | 全部读取，按运行时/协议消息/配置数据分工审阅，合并后交叉复核并同步当前状态 | 是规格检查；没有正式 Core/SDK 可做本轮编译或集成测试 |
| D-01～D-25、P-01～P-14、Q-01～Q-18 | D 原文保留，P/Q 逐项给出当前规则、后置或实测去向 | 不把本轮技术补全写成过去用户已逐项确认 |
| 旧 PDF 与后续讨论映射 | 保留 LEGACY-PRD-COMPARISON 的 FR-01～FR-30 与 POST-PRD-AUDIT 的 42 项历史记录 | 本轮没有重新取得/读取原 PDF；32 页核对属于历史审阅记录，不能声称重新核验所有历史聊天 |
| 历史 Native worker PoC | 保留 NOTES 的数据和限制 | 未重跑；同二进制 worker 的 PSS 不能证明真实多插件的总内存 |
| 官方技术主源 | 核对 Protobuf、Go HTTP、浏览器 Origin、SQLite/GORM/PostgreSQL、Telegram 等约束；链接在 NOTES 与对应专章 | 主源支持技术判断，不能证明尚未实现的 AlonaBot 集成正确 |

## 2. 本轮发现与修复

以下为具体缺陷及最小闭合措施；“已修复”仅表示规格已给出一致、可验收的规则。

| ID | 原规格缺口及失败后果 | 已落入的修复 | 条款与验收 |
| --- | --- | --- | --- |
| A-01 | 准备期禁止业务 RPC，却未提供配置/DB 输入；失败阶段没有合法回包，枚举零值不完整 | HelloAck Bootstrap 一次下发输入，REGISTER 回传规范候选；SESSION_ERROR、枚举零值及能力协商；Envelope 1～16/错误码 1～17 保号 | CONTRACTS §1、WIRE §1～§3；V-02/V-37 |
| A-02 | 配置更新可能复用旧 SQLite，候选 OnStart 写库后失败无法恢复旧数据 | reload 与 update 统一停止旧进程、Backup 独立候选 DB；提交前失败恢复旧程序新 Session，提交后 Degraded | CONTRACTS §5/§6/§12；V-24/V-29/V-30 |
| A-03 | secret 引用可以原地改值，旧 DB/配置快照仍会使用新密钥；移除字段后未知 secret 可能泄露 | Secret Store 保存不可变版本，配置只存引用；保留历史 secret 路径标记，按引用清理，restart 复用原引用 | CONTRACTS §5；V-25/V-38 |
| A-04 | Go 迁移回调执行方、成功账本事务、PG 维护许可输入及普通恢复分支不闭合 | SDK 执行迁移；Up 与账本同插件 DB 事务；Bootstrap 带启动原因/限定维护许可；PG resume 核对已提交结果且不重跑迁移 | CONTRACTS §6、WIRE §3；V-26/V-31/V-39 |
| A-05 | Journal 能选择代次，但不能阻止 Core 崩溃后旧子进程继续写入；并发管理动作可互相覆盖 | data_dir Core 锁、逐插件变更串行与预期代次检查、稳定 runtime.lock、SDK 断链退出；确认旧进程退出后才接管 | CONTRACTS §12；V-30/V-36 |
| A-06 | Ready 名称混用，OnReady 等待公开业务会自锁；多插件停机会累计超时 | Prepared/COMMITTED/Activated 分开；单 writer 写完 ACTIVATED 再开门；全局并行停止 10s，drain≤6s，STOPPING 明确剩余≤4s | CONTRACTS §4/§12、WIRE §3/§4；V-21/V-37 |
| A-07 | 只限制 Request 条数仍可堆积大载荷；Handler 内嵌套 Send/Service 可能堵住响应；Cancel 可越过未发送 Request | reader/writer/业务执行分离，数量与字节预算并用，控制优先但保留每 call 因果序；满额立即失败而非等自身槽位 | CONTRACTS §1～§3、WIRE §4；V-01/V-03/V-18/V-42 |
| A-08 | cron/时区规则不完整；任务超时即释放 running 会与仍执行的 callback 重叠 | 五字段 cron/IANA 时区、interval 起点、离线不补；调用超时与实际结束分开，实际返回或确认进程退出才释放 | CONTRACTS §4.3；V-20/V-46 |
| A-09 | PreEvent 可改按钮目标等路由字段；出站 Hook 经 Service 再 Send 可递归；AfterSend 失败可能改写发送事实 | Core 保留身份/Target/按钮任务绑定；send_forbidden 沿派生调用传播；AfterSend 仅观察回执 | CONTRACTS §2、MESSAGE §5；V-15/V-16 |
| A-10 | “实例授权”没有默认规则，后台 Send 要么全拒要么全放；媒体读取与发信权限混用 | 真实入站原 Target 回复默认允许；其他发送需 allowed_adapter_instances；Read/Grant/Send 分别核对已有对象授权与目标权限 | CONTRACTS §1/§8、MESSAGE §1；V-07/V-43/V-44 |
| A-11 | Handler 两签名无法表达停止；同名命令、Rule false/error、参数与 Raw 尾部有歧义；群管理员可能混为 Core Admin | Context.StopPropagation→HandlerOutcome；完整 token 最长路径和稳定单胜者；拒绝消费命令；参数语法明确；Core 管理角色与 chat_admin 分开 | CONTRACTS §2、MESSAGE §3；V-09～V-14/V-43 |
| A-12 | 媒体上传的回收边界不完整，入站图片只有描述却没有 SDK 读取路径 | 顺序上传/校验/配额/租约；IncomingMedia 惰性读取与共享缓存，offset/EOF 明确，不泄露平台 token/宿主路径 | CONTRACTS §1.3/§1.4、WIRE §5；V-06/V-44 |
| A-13 | Telegram 一次 Send 的多物理结果、按钮限制、入站确认、显式订阅与匿名身份不完整 | 全链预检、有序分组与逐部分结果；error 不丢成功 IDs；按钮 token/代次；显式 allowed_updates、游标/空闲轮次、真实 Actor 与冲突状态 | MESSAGE §4/§6；V-08/V-35/V-45 |
| A-14 | “统一入口+不同 Origin”没有具体部署方式；路径前缀本身无法隔离插件脚本 | 一个公开 Listener/端口配两个全局 host；清理代理认证头，独立插件会话；单 host 禁插件主动内容，声明插件间共享 Origin 边界 | CONTRACTS §7、DESIGN §I；V-22/V-47 |
| A-15 | 已有管理员会话的验证写了，但无法取得首个会话或恢复丢失凭证 | 本机 admin-token init/reset；浏览器登录、8h 会话、显式 CLI Bearer、重置撤销派生会话；复用同一个管理入口 | CONTRACTS §7.4、PLAN T-09；V-47 |
| A-16 | 新 Schema 可能先拒绝需要迁移的旧配置；typed marshal 丢未知键；null/default/PATCH 的含义模糊 | OnLoad→缺键默认→typed/Validate；保留原始 JSON；Merge Patch/完整替换区分；base_generation 防陈旧覆盖，字段类型明确 | CONTRACTS §5、DESIGN §J；V-23/V-25/V-38 |
| A-17 | 归档摘要可能自包含循环，验签失败可能降级本地信任，包解压和平台约束不足 | 外部 release.json 确切字节 Ed25519 签名；归档 SHA256 与单文件摘要分开；绑定来源，受控解压与固定限额，停机前验证 | CONTRACTS §9；V-27/V-28/V-41 |
| A-18 | restart/reload/首次 enable/卸载重装/回退语义不完整，恢复可能直接修改保留快照 | 完整 CLI 动作；同来源重装；rollback 创建新 Generation 与 SQLite 副本；普通 restart 保留 Active；PG 外部数据走维护 | CONTRACTS §6/§9/§12；V-39/V-40 |
| A-19 | 承诺五平台+原生 GORM，却没有可检验的默认 SQLite/Backup 驱动路线；静态扫描可能要求推导任意 Go 控制流 | 明确 pure-Go 默认组合，T-01 先验证并锁实际版本；只扫描自身受标记包级函数，实例方法完整集合在 REGISTER 校验 | CONTRACTS §6/§9、DESIGN §D；V-26/V-32/V-39 |
| A-20 | 实施任务缺少可执行依赖，历史待批语句与当前授权混用，审阅报告仍按旧文档数量和版本 | T-01～T-10 依赖收敛；全部 12 份当前/历史文档标明效力；V-01～V-47 写明未来验收，清理无效表格分隔 | PLAN、README、DECISIONS、TEST-MATRIX；本节静态检查 |

## 3. 历史自审 R-01～R-12 的承接

2026-10-19 自审曾记录下列问题。这里保留编号和当前去向，不把当时的“待定”当作本轮仍未处理的空白。

| 历史 ID | 原问题 | 当前去向 |
| --- | --- | --- |
| R-01 | 标准事件类型缺漏 | MESSAGE §1 类型化事件；平台未支持时不伪造 |
| R-02 | scheduled 无 Actor 却可能默认 Admin | CONTRACTS §2/§4/§8 的身份、job scope 与发送许可 |
| R-03 | 动态调用其他插件命令需求丢失 | 保留 D-24，CONTRACTS §3 明确后置 v0.2 |
| R-04 | 生命周期阶段缺少 Ready/重载/卸载 | CONTRACTS §4/§12 与 Wire 状态机 |
| R-05 | 运行 JSON 与 i18 的 config 包装层混淆 | CONTRACTS §5：运行无包装，i18 有包装 |
| R-06 | 翻译显示与 options 实际值未分离 | CONTRACTS §5.3、DESIGN §J |
| R-07 | Secret Store 被误解为同账号 Native 沙箱 | CONTRACTS §7 明确信任边界 |
| R-08 | 子进程树资源可能残留 | CONTRACTS §7/§12 与 V-34/V-36 平台验证 |
| R-09 | 多 Adapter 实例/账号 Target 与凭证作用域 | CONTRACTS §8、MESSAGE §1 与 V-07 |
| R-10 | FR-26 的设计章节引用错误 | 历史比较指向 DESIGN §O，保留章节索引 |
| R-11 | 插件 ID first-wins 与包内路由重复报错混淆 | D-20、CONTRACTS §2/§7/§9 |
| R-12 | GORM/互斥迁移、未知键与 i18 过于笼统 | CONTRACTS §5/§6 与 V-23～V-26/V-39 |

历史 v0.6 已修正的风险表格误插入、Prepared/Active 区分及 V-01～V-35 仍在现行文档中承接；本轮补充 V-36～V-47，并收紧原有用例。

## 4. 文档验证结果

本节仅填写对本轮成稿实际执行的静态检查；各检查通过后再写回源文件。

| 检查 | 实际结果 |
| --- | --- |
| 文件集合与当前版本 | 12 份原有 Markdown 齐全；现行主文档维持 v0.6-review，历史文件明确标记效力 |
| 既有决定与议题 | D-01～D-25 的 25 行原文与修订前一致；P 14 项、Q 18 项在决定表与契约处置表均连续无重号 |
| 任务与验收 | T-01～T-10 共 10 项、依赖图无环；V-01～V-47 共 47 项连续，验收表保持完整表格 |
| 历史追溯与本轮审阅 | FR 30 项、历史讨论 42 项、历史 R 12 项和本轮 A 20 项均连续有唯一表项 |
| 文档结构 | 59 个相对 Markdown 链接全部指向现存文件；7 个围栏块闭合，1 个 JSON 示例可解析；标题无重复，表格列数一致 |
| Wire 保号 | Envelope 1～16 的编号/名称/类型保持；错误码 1～17 保持，Kind 1～12 保持并追加 SESSION_ERROR=13；零值明确保留 |
| 语义交叉审阅 | 启动/就绪/停止、候选提交/回退、PG 恢复/维护、配置/secret、权限、媒体、HTTP、安装与任务依赖已逐项核对；发现的不一致已回写对应章节 |
| 运行时验证 | 未执行；V-01～V-47 明确为未运行，历史实验未重跑 |

静态检查不验证任意 Go 示例能编译，也不能证明 SQLite/PG 故障恢复、Telegram 发送或跨平台锁已正确实现。没有将规格审阅、官方文档检索或历史 PoC 记成新测试通过。

## 5. 可实施性结论与实施门槛

现有方案可以按纵向主链落地：T-01 固定可编译 contracts 和必要依赖，T-02 建立状态/进程/IPC，T-03～T-05 完成真实消息和插件处理；配置/数据库及 Service/Scheduler/HTTP 按明确依赖展开，T-09 再整合完整安装更新与 WebUI。协议、候选数据与恢复的关键假设已提前到 T-01/T-02 验证，避免等整体框架完成才发现基础路线不能成立。

以下仍属于实施期任务，不能在文档阶段关闭：

- 选定 Go/Protobuf/GORM/SQLite/PG 版本的真实编译、迁移与 Backup API 行为；五 OS/arch 的构建、启动、进程锁与普通子进程清理。
- SQLite WAL、副本、提交中断、旧 SDK 未退出、PG 维护分支和 secret 回退的故障注入结果。
- RPC 满队列、嵌套调用、Cancel 顺序、OnReady 门禁、job 超时仍执行等并发边界。
- Telegram 真实权限/媒体/按钮/多结果与 HTTP 两 host 的浏览器行为；首个管理凭证及重置路径。
- 不同真实插件程序及依赖的 RSS/私有内存、适用平台 PSS、CPU 和延迟，与 NoneBot2 等价负载比较。

**TEST-MATRIX 的 V-01～V-47 当前全部未运行。** 当某个假设在实现验证中失败，应修正对应实现或有依据地修订该条契约；不得宣称目前已达到生产可用、低于某个内存上限或具备跨平台强沙箱。

## 6. 维护规则

新增需求或修订行为时，同步 CONTRACTS/Wire/消息专章、DESIGN、P/Q 当前处置、PLAN 依赖与 TEST-MATRIX。历史 D 表保留原文；新技术决定写明来源和日期。状态必须区分“有规格”“已实现”“已实测”，不能只因各文件提到了同一个名词就认定行为闭合。
