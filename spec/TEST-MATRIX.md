# AlonaBot v0.1 验收与故障注入矩阵

版本：v0.6-review · 2026-10-21
状态：**计划中的验收用例；目前全部未执行**。用于实施交接与风险门槛，并非自动化测试报告。
规范来源：[PLAN.md](PLAN.md)（T-01～T-10 详细工作包）、[CONTRACTS-v0.1.md](CONTRACTS-v0.1.md)、[WIRE-FORMAT.md](WIRE-FORMAT.md)、[MESSAGE-MODEL.md](MESSAGE-MODEL.md)。决策依据见 [adr/DECISIONS.md](adr/DECISIONS.md)。阶段归属更新：2026-10-22。

## 1. 逐任务验收条目

| 测试 ID | 任务 | 场景/方法 | 通过条件 |
| --- | --- | --- | --- |
| V-01 | T-01、T-04、T-09 | 解析 manifest，市场与本地 ID、缺失/重复 ID、OS/arch 不匹配 | 市场规范 ID 可用；本地非前缀安全 slug 仅警告；路径穿越与缺失 ID 拒绝；重复 ID 的先后顺序稳定 |
| V-02 | T-01、T-02、T-04 | Protocol 1.0、枚举零值、Major/Minor、未知必需能力、Schema 缺失 | Major/required 不兼容明确拒绝；UNSPECIFIED/未知kind不视为成功；Minor正确取交集，无注册残留 |
| V-03 | T-02 | TCP 和 stdio 使用同样的 Length-Prefixed Protobuf fixture | 双传输序列化语义一致；4 MiB 限额、零帧及错帧正确拒绝；没有 TCP 独占功能 |
| V-04 | T-02 | 伪造 Session Secret、跨插件 Session、旧 Generation 请求 | 校验失败不激活；旧代次不接受请求；身份错误不会泄漏 Secret |
| V-05 | T-02、T-07 | 双向相同call_id、乱序/重复/Cancel迟到Response、父调用已完成 | 三元键不串号；用户父上下文失效后子请求不能沿用Actor；迟到结果不二次完成（scheduled只释放实际运行槽） |
| V-06 | T-02 | stderr/stdout 混入日志、单插件崩溃/退出 | TCP 日志带插件 ID；stdio stdout 非协议字节隔离会话；3 次异常崩溃限重启，正常退出不重启 |
| V-07 | T-03 | Telegram 不同实例/账号/话题的回复、主动发送与实例许可撤销 | 原事件Target可回复；后台/其他Target需allowed_adapter_instances；无许可PERMISSION_DENIED；凭证不串用、无目标NO_DEFAULT_TARGET |
| V-08 | T-03 | 一条 Text/Image/Reference/Buttons Chain、两次独立 Send | 一次 Chain 的逻辑发送和两次 Send 区分明确；支持能力正常渲染，不支持能力返回 UNSUPPORTED且不静默降级 |
| V-09 | T-03 | 大媒体分块上传、摘要/大小/到期、传递其他插件 | 未授权或过期 MediaRef 被拒绝；提交前校验哈希；转发有 Core 签发的范围受限 token |
| V-10 | T-04 | 包内多个 root/Config、注解函数、实例方法、外部依赖扫描 | 单 root/至多单 Config；多 Config/重复路由报错；依赖源码不扫描；实例方法不会被错误构造 |
| V-11 | T-04、T-05 | 长/短命令、别名、跨插件同名、Regex/Raw、引号/raw/可选值 | 最长完整路径+固定排序；跨插件遮蔽可见；路径已识别后禁用/权限/解析/Rule失败不回落；非法类型返回usage或声明silent |
| V-12 | T-04、T-09 | WebUI 改写组/子命令权限、禁用命令 | 不重启插件即可生效并持久化；Core 全局安全限制始终优先 |
| V-13 | T-05 | 同事件多插件 priorities、小优先级先、同优先级反向安装次序 | 以 priority、plugin_id、route_id 稳定排序，跨进程启动顺序变化不影响语义 |
| V-14 | T-05 | 两种Handler签名、自定义Rule false/error、Stop后error、自动型调用Send | HandlerOutcome能表达matched/propagation；显式Stop不被error清掉；自动非空回复只发一次，自动型Send明确拒绝 |
| V-15 | T-05、T-06 | PreEvent改身份/按钮任务绑定、BeforeSend拒绝/改链、AfterSend失败、Hook经Service递归Send | 身份/Target及已验证路由绑定不可改；每次Send一组Hook；错误不伪造成功或改变已获平台确认事实；直接和间接递归都被拒绝 |
| V-16 | T-05、T-07 | scheduled 没有 Actor、插件伪造 is_admin、按钮/消息编辑等事件 | 无 Actor 无默认 Admin；仅可信 Adapter/Core 赋予身份；未支持事件明确返回能力缺失 |
| V-17 | T-06 | Event Bus 的插件离线/队列溢出/慢订阅者 | 离线不积压、不补发；到固定上限丢弃并计数；其他订阅者不被阻塞 |
| V-18 | T-06 | A→B→C Service、A→B→A 循环、两条独立并发调用链 | 嵌套调用正确返回；同链重复插件时 CALL_CYCLE；两独立链互不误判 |
| V-19 | T-06 | Service JSON Schema、错误 encoding/未知版本、缺授权 | 不受支持的编码/Schema、超限 payload、跨插件不授权均返回明确错误且不运行目标函数 |
| V-20 | T-07 | 长短任务、五字段cron/IANA时区/夏令时、interval、离线和重叠 | 不同job并发；同job跳过重叠；漏tick不补；触发按固定起点与真实时间；非法语法注册失败 |
| V-21 | T-02、T-07、T-08、T-09 | 配置/迁移/OnStart/注册/OnReady失败与挂起；Core多插件Shutdown | 失败控制帧合法；各阶段有预算；提交前恢复旧版、提交后Degraded；新路由不提前开放；全部插件共享10s停止预算 |
| V-22 | T-07、T-09 | 插件仅注册 Go `http.Handler`，自动启动/停止内部 loopback Listener；Core 统一反向代理、公开/私有回调与不可信页面 | 插件作者不提供端口且无 HTTP 路由插件不监听；用户/容器只暴露 Core 公开端口；内部端口仅 loopback、插件停止/Session 变更不可再代理；私有默认鉴权、公开声明需校验、Core 保留路径不可覆盖、插件 JS 不与 Admin UI 同 Origin |
| V-23 | T-08、T-09 | config.json 直接编辑、显式 reload、false/0/空串、未知键 | 编辑未激活前运行值不变；reload 创建 Pending；显式默认值不被覆盖；未知键保留 |
| V-24 | T-08、T-09 | 仅改配置，候选OnStart写SQLite后失败；Core在候选阶段中断 | 候选使用独立DB，旧数据/配置/secret不变；新旧进程不同时写入；恢复只信Core状态库Active指针 |
| V-25 | T-08、T-09 | secret修改后候选失败、旧secret字段转未知键、hidden/i18、掩码/PATCH | 新secret不覆盖旧引用；保留未知secret仍保密；掩码不成为保存值；hidden不意外删除；翻译不改键和值 |
| V-26 | T-01、T-08 | GORM SQLite/PostgreSQL CRUD/Tx、Migrate 与 Migrations 同注册 | 两后端基本访问可用，迁移声明互斥检查失败；PG 最小权限凭证不复用 Core 超级用户 |
| V-27 | T-09 | 安装包归档含 ../、绝对路径、symlink 逃逸、来源 ID 冲突 | 解压前拒绝路径越界；首次注册 ID 保留；后来源不可静默替换 |
| V-28 | T-09 | 下载损坏 digest、签名错误、错误 OS/arch、不受信来源更新 | 所有错误在停止旧版前发现并拒绝；SHA256 不当作来源身份签名 |
| V-29 | T-09 | SQLite WAL候选备份；迁移/OnStart失败；空间不足与残存进程 | 确认runtime.lock可取后备份；旧DB不污染；候选失败旧程序新Session恢复；空间/锁失败不激活 |
| V-30 | T-09 | Core崩溃于PREPARED、Active+COMMITTED提交前后、OnReady失败 | 未提交只恢复旧版；已提交只恢复新代次；旧进程退出未确认时不启动第二写入者；无双代次接单或静默回滚 |
| V-31 | T-09 | PG 或插件自管 SQL 有不可隔离变更，尝试自动升级 | 返回 NEEDS_MANUAL_MIGRATION，提示维护/备份流程；无虚假无损回退承诺 |
| V-32 | T-10 | Linux amd64/arm64、Windows amd64、macOS amd64/arm64 编译、启动/退出 | 五个组合都有真实记录；CGO 插件在不兼容平台标不支持而非虚假发布 |
| V-33 | T-10 | 真实不同二进制、GORM/HTTP 下载、长时并发与 NoneBot2 等价工作负载 | 报告 RSS/私有内存、Linux PSS、CPU、P50/P95/P99 与版本/机器/负载；不借同 worker PoC 代替 |
| V-34 | T-10 | 100 次启停、多次异常退出、普通子进程树、外部副作用重试 | 无 Core 崩溃/会话泄漏，合作型子进程按平台能力清理，业务副作用默认不自动重试 |
| V-35 | T-03 | 多物理消息、平台明确拒绝、先成功再超时、相册按钮失败、成功message_id=0 | ACCEPTED/PARTIAL/REJECTED/UNKNOWN与每part一致；已确认ID保留、0不当可编辑ID；未尝试列出；payload+error同时返回，不自动补发 |
| V-36 | T-02、T-08、T-09 | 同data_dir双Core、同插件并发enable/update、Core被kill而子进程仍写DB | 独占锁/管理串行及预期代次检查生效；SDK断链退出；runtime.lock未释放不打开DB/启动替代实例 |
| V-37 | T-01、T-02、T-08、T-09 | 旧配置需OnLoad才能符合新Schema、Bootstrap/REGISTER超4MiB、阶段错误 | 旧JSON先迁移再验证；完整配置只经私有IPC；超限/失败用合法SESSION_ERROR且无半注册；ACTIVATED严格先于业务帧 |
| V-38 | T-08、T-09 | 配置版本迁移、nested未知键、null/缺失/false、Merge Patch、陈旧base_generation | 迁移/默认/typed/Validate顺序固定；null删除与空值区分；未知键保留；陈旧提交拒绝；restart不偷改Active |
| V-39 | T-01、T-08、T-09 | GORM两后端、逐步迁移失败/重复ID、PG空schema首装、Active resume、已有数据升级与伪维护许可 | Up+成功账本同插件DB事务；失败无成功记录；PG空schema可初始化；resume只核对已提交迁移；非当前Session/代次许可无效；未许可生产写入前返回维护错误 |
| V-40 | T-09 | install首次enable、restart/reload区别、pin、uninstall重装与显式rollback | 首次创建Active；restart仅新Session；卸载保留数据/来源且仅同来源可重装；pin持久化；rollback建立新代次/DB副本且不复活旧按钮，PG不假装通用恢复 |
| V-41 | T-01、T-09 | 外部release清单签名、归档摘要、来源/公钥变化、无签名本地包 | 对确切字节验签；无自包含摘要循环；所有身份/平台错误停机前拒绝；签名失败不降级本地信任 |
| V-42 | T-02、T-06 | Request计数满、字节满、缓冲满、嵌套Send/Service与Cancel并发、Request发送前取消 | 超额明确失败；Response/控制仍有进展；无阻塞等待自己额度；Cancel不超越自己的Request；Event Bus 256条/4MiB双界限计数并drop-newest |
| V-43 | T-04、T-05、T-09 | Core Admin与chat_admin、组禁用/子项启用、子权限放宽、运行时改策略 | 群管理员不取得Core管理权；父禁用优先；最近显式权限可覆盖但不能绕过全局限制；已开始请求不伪称撤销 |
| V-44 | T-03 | 入站媒体惰性Read、多Session共享、越界offset/配额/过期、owner退出 | 不预下载所有图；不泄BotToken/路径；首次取件校验后读取；有效租约保护共享对象；错误不制造MediaRef/空成功 |
| V-45 | T-03 | getMe账号重复、webhook/401/409、allowed_updates/匿名身份、游标提交崩溃及7天空闲ID重置 | 冲突报管理状态、不擅自删webhook；逐条持久游标后非持久派发；显式订阅成员/Reaction且核对权限；匿名sender_chat不伪造User；7天空闲重建轮次不沿用旧高offset；允许丢失边界明确 |
| V-46 | T-07 | job超时后callback继续执行、迟到Response、RPC断链但进程未退出 | 超时结果先返回但running不释放；无重叠/补跑；实际callback返回或进程确认退出才释放；迟到成功不改超时 |
| V-47 | T-07、T-09 | 首次管理token/离线reset、登录/CLI、同端口两个Host、代理身份头伪造、单Host主动内容 | 插件host不可访问Admin API；admin Cookie/Authorization不传入插件；Host/CSRF/内部认证检查生效；首次能登录，reset撤销旧会话/票据；单Host不执行插件主动内容 |

## 2. 跨任务验收的最终闭合点

“任务”列列出参与实现与验证的阶段，不要求最早任务提前完成尚不存在的下游能力。单项证据可以分阶段记录，整项 V 只有全部分项完成后才可通过；T-10 核对所有证据及真实平台结果。

| 验收编号 | 前阶段要交付的证据 | 最终集成部分 |
| --- | --- | --- |
| V-01、V-02 | T-01 Schema/字段/能力 fixture；T-02 握手拒绝；T-04 注册冲突 | T-04 完成 V-02 注册原子性，T-09 完成 V-01 安装来源/ID 占用 |
| V-05、V-16 | T-02 请求关联/迟到结果；T-05 无 Actor 的 payload 与权限 fixture | T-07 真实 scheduled callback、迟到响应和身份行为 |
| V-11、V-43 | T-04 命令/策略服务，T-05 Rule 与事件链 | V-11 在 T-05 完成；V-43 的真实管理入口变更在 T-09 复验 |
| V-12 | T-04 持久策略服务与下一次准入生效 | T-09 鉴权后通过真实 WebUI 修改并验证 |
| V-15 | T-05 三阶段 Hook、直接 Send 禁令与可信标记继承 | T-06 真实 Hook→Service→Send 间接递归拒绝 |
| V-21、V-37 | T-01 类型/上限，T-02 基础阶段/激活门；T-07 停止集成，T-08 配置/DB 回调 | T-09 与正式候选提交前后失败恢复组合 |
| V-22、V-47 | T-07 Listener/代理、Host/凭证边界、票据与测试会话组件 | T-09 真实 init/reset、login/CLI、私有插件页面与授权撤销 |
| V-23～V-25、V-38 | T-08 raw JSON、Secret、候选 DB、表单/i18 组件 | T-09 真实 reload/WebUI/API 和 Journal 崩溃恢复 |
| V-26 | T-01 驱动 CRUD/事务验证 | T-08 实际 SDK 连接、迁移声明互斥与插件凭证边界 |
| V-36 | T-02 双 Core/操作串行/进程锁，T-08 DB 锁与候选资源 | T-09 正式 enable/update、Core 突退且旧进程仍写入 |
| V-39 | T-01 驱动/事务小验证，T-08 迁移及许可检查 | T-09 首装/resume/更新及显式维护的真实管理入口 |
| V-41 | T-01 外部清单与签名对象 fixture | T-09 真实包验签/来源/更新预检 |
| V-42 | T-02 传输额度、控制进展与取消顺序 | T-06 真实 Service 嵌套与 Bus 双预算 |
| V-32～V-34 | 各任务提供真实示例、构建与故障夹具 | T-10 五目标运行、100 次启停及等价资源比较 |

T-03 的本地 Bot API 故障服务与真实 Telegram 结果分别记录；前者可支持后续开发，真实平台部分仍由 T-10 查证。测试会话、无配置/无 DB 最小插件和 payload fixture 的范围必须写明，不可据此关闭生产认证或数据恢复用例。

## 3. 证据和失败门槛

每项 V-xx 在实施后才填写：测试代码位置、运行命令、实际平台与版本、运行日志/报告路径、通过或失败、所对应的变更。**V-01～V-47 当前全部为「未运行」**，且 V-35 明确纳入多物理消息的部分成功场景。不能把“规格写明如何测”描述成测试通过。

发布阻塞：协议/Session 鉴权、Core 全局权限、数据代次唯一性、回滚/维护的真实性、安装包越界/来源安全及目标五平台构建失败，均不能通过文案豁免。非阻塞项的阈值（例如内存量与吞吐目标）须等基线测量之后再由产品方批准；**不杜撰 100 MB 等硬性能指标**。

## 4. 实施交接

1. 采用 v0.6-review 技术基线，按 PLAN 的 T-01 开始；任务完成条件及部分验收的后续承接以详细工作包为准。
2. 执行方确认 Go 编译器与目标 OS/arch CI 可用；源码、数据库、Telegram 凭证以及真实测试环境由授权范围决定。
3. T-01→T-02→T-03→T-04→T-05 完成主链路，T-08可在T-04后并行，T-06/T-07在T-05后并行，T-09待相关依赖完成再集成。
4. 执行中发现契约缺口，先修对应契约、工作包和用例，在 adr/DECISIONS.md 记录原因；不能在代码中默默更改已接受的协议。

