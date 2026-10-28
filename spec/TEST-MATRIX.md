# AlonaBot v0.1 验收与故障注入矩阵

版本：v0.6-review · 2026-10-21
状态：**T-01/T-02 的分项已有运行证据（见 §3.1），其余分项与全部跨任务闭合点未执行**。用于实施交接与风险门槛，并非自动化测试报告。
规范来源：[PLAN.md](PLAN.md)（T-01～T-16 详细工作包）、[CONTRACTS-v0.1.md](CONTRACTS-v0.1.md)、[WIRE-FORMAT.md](WIRE-FORMAT.md)、[MESSAGE-MODEL.md](MESSAGE-MODEL.md)。决策依据见 [adr/DECISIONS.md](adr/DECISIONS.md)。阶段归属更新：2026-10-22。

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
| V-26 | T-01、T-08 | 受管 SQLite 经 `sqlitedb`（database/sql）与 `gormdb` 的 CRUD/Tx、同一插件同时使用两个入口、根插件实现多于一个迁移接口、只导入 `sqlitedb` 的插件 | 两个入口都可用，共用同一连接并看到同一数据；实现多个迁移接口时检查失败；只导入 `sqlitedb` 的插件不链接 GORM |
| V-27 | T-09 | 安装包归档含 ../、绝对路径、symlink 逃逸、来源 ID 冲突 | 解压前拒绝路径越界；首次注册 ID 保留；后来源不可静默替换 |
| V-28 | T-09 | 下载损坏 digest、签名错误、错误 OS/arch、不受信来源更新 | 所有错误在停止旧版前发现并拒绝；SHA256 不当作来源身份签名 |
| V-29 | T-09 | SQLite WAL候选备份；迁移/OnStart失败；空间不足与残存进程 | 确认runtime.lock可取后备份；旧DB不污染；候选失败旧程序新Session恢复；空间/锁失败不激活 |
| V-30 | T-09 | Core崩溃于PREPARED、Active+COMMITTED提交前后、OnReady失败 | 未提交只恢复旧版；已提交只恢复新代次；旧进程退出未确认时不启动第二写入者；无双代次接单或静默回滚 |
| V-31 | T-09 | 声明 `prepare_effects=external` 的插件（准备期写外部数据库），尝试自动 reload、update 与 rollback | 返回 NEEDS_MANUAL_MIGRATION，提示维护/备份流程；维护许可只对本次 Session/代次有效；普通 restart 不受影响；无虚假无损回退承诺 |
| V-32 | T-10 | Linux amd64/arm64、Windows amd64、macOS amd64/arm64 编译、启动/退出 | 五个组合都有真实记录；CGO 插件在不兼容平台标不支持而非虚假发布 |
| V-33 | T-10 | 同一台固定机器、同一个本地 Bot API 模拟服务；NoneBot2 使用其 Telegram Adapter 与功能相同的插件。四种负载：空闲（Core 与 echo、configdb、HTTP 示例三个插件运行 10 分钟）；消息（每秒 20 条文本、持续 10 分钟，echo 逐条回复）；数据库（每秒 20 条命令，每条在 SQLite 中写入并读取一行，持续 10 分钟）；HTTP（每秒 50 个请求访问插件路由，持续 10 分钟） | 每种负载报告 RSS/私有内存、Linux PSS、CPU 与回复或响应延迟的 P50/P95/P99，并记录版本、机器与负载参数；不借同 worker PoC 代替 |
| V-34 | T-10 | 100 次启停、多次异常退出、普通子进程树、外部副作用重试 | 无 Core 崩溃/会话泄漏，合作型子进程按平台能力清理，业务副作用默认不自动重试 |
| V-35 | T-03 | 多物理消息、平台明确拒绝、先成功再超时、相册按钮失败、成功message_id=0 | ACCEPTED/PARTIAL/REJECTED/UNKNOWN与每part一致；已确认ID保留、0不当可编辑ID；未尝试列出；payload+error同时返回，不自动补发 |
| V-36 | T-02、T-08、T-09 | 同data_dir双Core、同插件并发enable/update、Core被kill而子进程仍写DB | 独占锁/管理串行及预期代次检查生效；SDK断链退出；runtime.lock未释放不打开DB/启动替代实例 |
| V-37 | T-01、T-02、T-08、T-09 | 旧配置需OnLoad才能符合新Schema、Bootstrap/REGISTER超4MiB、阶段错误 | 旧JSON先迁移再验证；完整配置只经私有IPC；超限/失败用合法SESSION_ERROR且无半注册；ACTIVATED严格先于业务帧 |
| V-38 | T-08、T-09 | 配置版本迁移、nested未知键、null/缺失/false、Merge Patch、陈旧base_generation | 迁移/默认/typed顺序固定；null删除与空值区分；未知键保留；陈旧提交拒绝；restart不偷改Active |
| V-39 | T-01、T-08、T-09 | 两个入口的逐步迁移失败/重复ID、首装、Active resume、external 插件已有安装的升级与伪维护许可 | Up+成功账本同插件DB事务；失败无成功记录；resume 不建立候选；非当前Session/代次许可无效；未许可的 external 候选在准备前返回维护错误 |
| V-40 | T-09 | install首次enable、restart/reload区别、pin、uninstall重装与显式rollback | 首次创建Active；restart仅新Session；卸载保留数据/来源且仅同来源可重装；pin持久化；rollback建立新代次/DB副本且不复活旧按钮，external 数据不假装通用恢复 |
| V-41 | T-01、T-09 | 外部release清单签名、归档摘要、来源/公钥变化、无签名本地包 | 对确切字节验签；无自包含摘要循环；所有身份/平台错误停机前拒绝；签名失败不降级本地信任 |
| V-42 | T-02、T-06 | Request计数满、字节满、缓冲满、嵌套Send/Service与Cancel并发、Request发送前取消 | 超额明确失败；Response/控制仍有进展；无阻塞等待自己额度；Cancel不超越自己的Request；Event Bus 256条/4MiB双界限计数并drop-newest |
| V-43 | T-04、T-05、T-09 | Core Admin与chat_admin、组禁用/子项启用、子权限放宽、运行时改策略 | 群管理员不取得Core管理权；父禁用优先；最近显式权限可覆盖但不能绕过全局限制；已开始请求不伪称撤销 |
| V-44 | T-03 | 入站媒体惰性Read、多Session共享、越界offset/配额/过期、owner退出 | 不预下载所有图；不泄BotToken/路径；首次取件校验后读取；有效租约保护共享对象；错误不制造MediaRef/空成功 |
| V-45 | T-03 | getMe账号重复、webhook/401/409、allowed_updates/匿名身份、私聊/回复机器人/提及机器人时的 to_me、实例代理与全局代理、代理连接失败、游标提交崩溃及7天空闲ID重置 | 冲突报管理状态、不擅自删webhook；逐条持久游标后非持久派发；显式订阅成员/Reaction且核对权限；匿名sender_chat不伪造User；7天空闲重建轮次不沿用旧高offset；允许丢失边界明确 |
| V-46 | T-07 | job超时后callback继续执行、迟到Response、RPC断链但进程未退出 | 超时结果先返回但running不释放；无重叠/补跑；实际callback返回或进程确认退出才释放；迟到成功不改超时 |
| V-47 | T-07、T-09 | 首次管理token/离线reset、登录/CLI、同端口两个Host、代理身份头伪造、单Host主动内容 | 插件host不可访问Admin API；admin Cookie/Authorization不传入插件；Host/CSRF/内部认证检查生效；首次能登录，reset撤销旧会话/票据；单Host不执行插件主动内容 |
| V-48 | T-03、T-09 | 安装/卸载 Adapter 包、同包多个实例单独启停、Adapter 进程崩溃后恢复、Adapter 为非自身实例交付事件、插件依赖 Adapter 的 extension 片段 | 启停一个实例不影响同包其他实例且不重启进程；崩溃后只恢复启用的实例并从已保存游标继续；伪造实例的事件被拒；凭证不进入插件；必需依赖缺失时拒绝启用插件；extension 发往其他 Adapter 返回 UNSUPPORTED |
| V-49 | T-04、T-05 | keyword/startswith/endswith/fullmatch 与 Regex 混合声明、ignore_case、空模式、命令与文本匹配同时可命中、`ToMe` Rule | 已识别普通命令时文本匹配不运行；文本匹配层按 priority/plugin_id/route_id 统一排序；ignore_case 按 Unicode 简单折叠；空模式与非法 Regex 拒绝注册 |
| V-50 | T-05 | Handler 请求延续调用后立即返回、延续调用内多轮等待、两个用户在同一群同时等待、等待超时、等待期间插件停用、延续调用内再请求延续调用、第 129 个进行中的延续调用 | 延续调用保留原 Actor/Target 且不占分发链；原事件传播只看原 HandlerOutcome；等待只命中同一 Actor 同一 Target（含 thread）的下一条消息且该消息不进入普通分发；同键第二个等待 ALREADY_EXISTS；超时 DEADLINE_EXCEEDED；Session 失效后不补发；嵌套请求 FAILED_PRECONDITION；超过 128 个时 RESOURCE_EXHAUSTED，普通派发不受影响 |
| V-51 | T-03 | 实例启动、管理员停用实例、Adapter 进程崩溃、401/409 使实例停止、Core 重启 | 每次进入运行状态有一条 bot.online，离开运行状态有一条带正确 reason 的 bot.offline；事件无 Actor/Target；Core 重启不补发 |
| V-52 | T-03、T-09 | 必需/可选插件依赖缺失、版本低于 min_version、停用或卸载被依赖插件、依赖成环 | 必需依赖不满足时拒绝启用并列出缺失项；可选依赖缺失只记诊断；停用/卸载/降级被依赖插件须显式确认，确认后依赖方继续运行且调用返回 UNAVAILABLE；成环在安装时拒绝 |
| V-53 | T-03 | 编辑文本消息、编辑带按钮、内容未变化、编辑不可编辑的消息、目标不支持 message.edit、Edit Chain 含媒体 | 成功时 SendResult 只有一项且 ID 为原消息；平台拒绝为 REJECTED 且不发新消息；不支持时 UNSUPPORTED；Chain 不合规时 INVALID_ARGUMENT；每次 Edit 执行一组出站 Hook |
| V-54 | T-03、T-08、T-09 | Actor.locale 为 zh-CN/zh-TW/未知/absent、插件缺少翻译键、占位符不一致、Core usage 与确认消息、WebUI 切换语言 | 选择顺序为 Actor locale→Core 默认→en-US，主语言回落确定；check/pack 拒绝缺少 en-US 与占位符不一致；未知键返回错误不输出键名；日志与错误始终为英文 |
| V-55 | T-11、T-09 | 工具 allow/confirm/deny、确认允许/拒绝/超时/他人点击、确认消息展示的参数、无 Actor 的后台调用、目标不支持按钮、core.tool.confirm 的 MCP 操作、运行时改策略 | deny 与拒绝返回 PERMISSION_DENIED 且不执行工具；超时 DEADLINE_EXCEEDED；invoke 的确认消息展示实际 JSON 参数；只接受同一 Actor 的第一次点击；无 Actor 时 confirm 直接拒绝；不支持按钮时 UNSUPPORTED；策略下一次调用生效；工具仍按继承 Actor 检查 required_permission |
| V-56 | T-11、T-09 | 多条原子追加、跨帧上限分页、跨插件读取、清除、Core 重启、卸载与 purge | 序号连续且顺序稳定；分页不丢不重；插件只能访问自己的会话；重启后保留；卸载默认保留，purge 删除；不截断 |
| V-57 | T-12 | OpenAI 兼容与 Anthropic 的流式文本与工具调用（本地模拟服务）、/ask 与自由对话开关、群聊中非 to_me 的消息、`/provider` 切换、人格切换、/reset、超过单条长度的回复、Provider 返回错误或中途断开 | 消息先发送再按间隔编辑，最后一次编辑为完整内容；超长回复在段落边界分成多条；自由对话默认关闭，只有 Admin 能开启，群聊中只响应 to_me 的消息；工具调用正确解析；断开或错误时保留已显示内容并报告错误，不自动重发请求；历史写入会话存储 |
| V-58 | T-12 | 模型连续调用 Core 工具与 MCP 工具（stdio、streamable HTTP 模拟服务）、确认拒绝、工具出错、超过单轮步数上限 | 每个工具经策略与确认；拒绝与错误作为工具结果交回模型；达到上限后结束本轮并告知用户；MCP 进程随插件停止 |
| V-59 | T-12 | Agent 的记忆：`memory` 工具的 add/patch/remove、超过字符上限、重复条目、不可见字符、群聊多用户、会话中写入与 /reset、管理员把 `memory` 工具设为 confirm 与 deny | 超限与非法内容返回错误且不丢已有条目；快照在会话内不变、下一会话生效；每位用户只看到自己的档案；`memory` 工具按策略确认或拒绝，拒绝作为工具结果交回模型 |
| V-60 | T-12 | 技能目录含合法与缺少 frontmatter 的 SKILL.md、load_skill 不存在的技能 | 系统提示只含合法技能的描述；不合法文件记诊断并跳过；load_skill 返回正文或明确错误 |
| V-61 | T-12 | 非 Admin 尝试创建任务、创建 interval/cron 定时 Agent 任务、到期运行、无发信许可、需要确认的工具、插件离线错过 | 只有 Core Admin 能创建与取消任务；任务按 CONTRACTS §4.3 的触发规则运行且不补跑；无许可时 PERMISSION_DENIED；confirm 工具一律拒绝；后台运行不注入记忆 |
| V-62 | T-02、T-09 | Core/插件/Adapter 三类来源、插件写合法 JSON 行与非 JSON 行、插件伪造来源字段、Handler panic、text 与 json 格式、终端与非终端输出、含 Bot Token 的 URL 错误 | 控制台格式与来源标签符合 CONTRACTS §14.2；非结构化行原样保留为 WARN；插件不能覆盖来源字段；panic 为一条带 stack 的 ERROR；非终端或设置 NO_COLOR 时不着色；日志中没有 token、Secret 与配置明文 |
| V-63 | T-02、T-09 | 全局与按来源的级别、off、运行时修改、Core 重启、最近记录上限、SSE 订阅、`alonabot logs`、WebUI 筛选 | 修改立即生效；按来源的设置重启后保留，全局级别的运行时修改重启后恢复；off 时 SDK 不输出而 Core 关于它的记录仍在；每个来源最多保留 1000 条；未认证无法读取；WebUI 筛选不改变输出 |
| V-64 | T-13 | 库模式下同进程插件与 Adapter 的完整生命周期、从 `instances/` 读取实例配置、dev 来源尝试 update/rollback、未捕获的 panic、停止超时、已安装包与开发组件混合运行、两次启动之间的配置与 dev.sqlite | 握手、注册、权限与独立进程一致；dev 组件使用固定代次，重启后按钮仍有效；update/rollback 被拒绝；panic 使进程退出且不重启；停止超时时 Run 返回错误 |
| V-65 | T-13 | 热部署下改动插件源码、编译失败、改动 go.mod、来自其他 module 的 Adapter | 只重启开发中的组件，Adapter 连接与其他插件不中断；编译失败时旧组件继续运行；go.mod 变化输出 WARN；其他 module 的组件不重启 |
| V-66 | T-14 | 干净数据目录首次运行、交互终端与非终端、预装包签名错误、卸载预装包后重启、新版本发行包 | 预装包全部安装，只启用 help；help 与官方 WebUI 已启用；Agent 与 Telegram、OneBot 11 Adapter 都未启用，没有对应进程；终端下 token 只显示一次，非终端拒绝启动；签名错误的包不安装并记 ERROR；已卸载的不重新安装；新版本只作为可用更新列出 |
| V-67 | T-14 | help 的权限过滤与翻译、`/whoami`、welcome 的入群欢迎与开关权限、rss 的订阅权限/去重/推送/私网地址/超大响应 | help 只列出调用者可用的命令；`/whoami` 返回调用者的实例与用户 ID；welcome 只响应 member.joined，只有 chat_admin 或 Core Admin 能开关；rss 的增删只有 chat_admin 或 Core Admin 能执行，超过 5 MiB 或 30s 的抓取失败，不重复推送同一条目，拒绝 loopback、私网与链路本地地址（含重定向之后），后台推送需要 allowed_adapter_instances |
| V-68 | T-02、T-09 | 跨过零点、同日写满单文件上限、合计超过 1 GiB、超过 30 天、max_size 大于 max_total、磁盘写满、按日期查询 | 跨日换到新日期的文件；同日续写编号文件；从最旧的文件开始删除且不删正在写入的文件；配置矛盾时拒绝启动；写入失败时 Core 继续运行并只报告一次；`alonabot logs --date` 与 WebUI 能查到前一天的记录 |
| V-69 | T-02 | `--args` 直接参数与文件输入、重复出现的键、未知键、`100MB` 等十进制单位、违反约束与相互矛盾的值、`PLUGIN_MEMORY_SOFT_LIMIT` | 后出现的覆盖先出现的；错误使 Core 拒绝启动并指明参数或文件行号；有覆盖时启动 WARN 列出全部覆盖项；SDK 使用 Bootstrap 下发的值；子进程收到 GOMEMLIMIT |
| V-70 | T-14 | linux/amd64 与 linux/arm64 镜像、干净卷首次运行、无 TTY 启动、`-it admin-token init`、端口只发布到宿主本机、停止容器 | 镜像只含 `alonabot`、`packages/` 与 CA 证书并以 UID 10001 运行；无 TTY 时拒绝启动并提示初始化；初始化后正常启动并能访问 WebUI；出站 HTTPS 可用；没有 `ALONABOT_CONTAINER` 时 `0.0.0.0` 搭配 http Base URL 被拒绝；停止容器后插件进程全部退出 |
| V-71 | T-15 | 本地 OneBot 11 模拟服务：get_login_info、access_token 错误、断线重连、self_id 不符、字符串格式上报、私聊/群聊/匿名消息、撤回与成员变动、to_me 的三种情况（含 get_msg 失败）、入站图片取件、实例队列满；发送混合段、video 拆分、reference、base64 媒体、async、failed、请求写出后断线 | 401 与字符串格式使实例致命失败并给出 safe_error；重连不重放已提交的 Send；self_id 不符的事件被丢弃；匿名消息没有 Actor；无通用对应的段只在 platform_payload 中；队列满时丢弃并计数而读循环继续；PartResult 与物理消息一一对应，async 标记 message_ids_unavailable，写出后断线为 UNKNOWN 且后续 NOT_ATTEMPTED；真实 QQ 端只有维护者手测记录，没有时记为未运行 |
| V-72 | T-03、T-15 | 同一插件经 Telegram 与 OneBot 11 模拟服务的两个实例收发同一组 Chain：纯文本、mention、reference、图片加文本、buttons、forward、file、markdown、Edit、`onebot11.face` extension、发往其他 Adapter 的 extension；同一组入站消息 | 通用片段在两边语义一致：顺序、引用位置、PartResult 与 segment indexes 对应；每个实例只用自己声明的 Capability，未声明的片段或组合在首个平台请求前返回 UNSUPPORTED，模拟服务没有收到请求；extension 只在所属 Adapter 上成功；入站的 text、mention、reference、image 在两边得到相同结构的 chain；插件代码不按 adapter_id 分支 |
| V-73 | T-13、T-16 | 用 `plugintest` 测试示例插件：命令回复、Rule、普通用户/chat_admin/Core Admin 的权限、Capability 子集下的 UNSUPPORTED、按钮点击、多轮等待、schedule 触发、后台发送有无许可、配置与 `dev.sqlite`；`go test -count=2 -race`；`alonabot.Start` 的激活失败与 ctx 取消 | 每次注入在分发结束或进入多轮等待后返回，结果完整，不依赖 sleep；行为与独立进程运行相同；测试之间的数据目录互不影响；Start 失败或停止后不残留组件 goroutine、core.lock 与子进程 |
| V-74 | T-10 | linux/amd64，与 V-33 同一台机器；各进程 Activated 后空闲 2 分钟，读 `/proc/<pid>/smaps_rollup` 的 Pss，取 3 轮中位数：只运行 Core（不启用任何插件与 Adapter）；echo 示例插件进程；rss 插件进程（`sqlitedb`，有一个订阅）；configdb 示例的 `gormdb` 变体进程 | Core ≤ 24 MiB，echo ≤ 10 MiB，rss ≤ 14 MiB，configdb ≤ 16 MiB；同时记录 RSS 与私有内存。超出视为回归，按 CONTRACTS §10 处理 |
| V-75 | T-01、T-10 | `node scripts/check-go-deps.mjs`：在五个目标上列出 Core 可执行文件（`./cmd/alonabot`）与 SDK 根包（`./sdk`）链接的模块；临时加入一个清单外的 import | 只含允许清单中的模块；清单外的模块使检查失败并列出模块名 |
| V-76 | T-12 | 测试插件调用 `llm.generate`：在与不在 `llm_service_allowed_plugins` 中、指定不存在的 Provider、模拟 Provider 出错与中途断开、调用方期限短于生成耗时、Agent 配置中没有 Provider、Agent 组件未启用、调用方经 `@agent` 调用 | 不在列表中返回 PERMISSION_DENIED，模拟服务没有收到请求；不存在的 Provider 为 INVALID_ARGUMENT，没有 Provider 为 FAILED_PRECONDITION；出错与断开为 UNAVAILABLE，safe_message 不含 API Key，模拟服务只收到一次请求；期限到达时调用方得到 DEADLINE_EXCEEDED；Agent 未启用时按目标未激活失败；`@agent` 解析到已安装的 Agent 组件；成功时返回 text、provider、model 与用量，会话存储没有新增条目 |
| V-82 | T-09、T-12 | 已安装官方 Agent、WebUI 与 Telegram Adapter 时，不带与带 `--replace` 安装同类的社区组件（本地签名的测试包与本地 Release 模拟服务）、新组件预检失败、替换后 update、`@agent` 调用、按原来源重装旧组件、两个不同 `adapter_id` 的 Adapter | 不带 `--replace` 返回 ALREADY_EXISTS 并列出已安装的组件；预检失败时旧组件不受影响；替换在确认后卸载旧组件并保留其数据与来源绑定，新组件沿用旧组件的启用状态，被替换 Adapter 的实例不出现在新 Adapter 下；之后 update 只查询新来源；`@agent` 调用到达新 Agent；旧组件可按原来源重装并复用数据；不同 `adapter_id` 的 Adapter 可以并存 |
| V-83 | T-08、T-09 | WebUI 组件：已启用、未安装、Admin API 主版本不在声明范围内、前端路由路径、`/api/` 与内部路径、带扩展名的不存在路径、库模式 `alonabot.WebUI`；图标：manifest 与表单选项声明的 SVG、PNG、WebP，未声明的包内文件、其他扩展名、含 `<script>` 的 SVG、未认证请求 | WebUI 文件来自不可变包目录，前端路由回退 `index.html`，`/api/` 不被覆盖，带扩展名的不存在路径返回 404；未安装返回 503 与提示；版本不兼容时拒绝安装与启用；图标只提供声明过的路径并带 sandbox CSP 与 nosniff，其他扩展名在包校验时拒绝，未认证返回 401；直接打开含脚本的 SVG 时脚本不在 Admin 源执行 |
| V-84 | T-14 | 在真实 systemd Linux（root）与 macOS（当前用户）上：全新安装、`SHA256SUMS` 不一致、经管道运行且有与没有终端、再次运行升级、`--uninstall`、不支持的系统与没有 systemd 的 Linux | 校验失败时不改动已有安装；有终端时 token 只显示一次并启动服务，没有终端时不启动服务并输出初始化命令；升级保留数据目录、设置文件与已安装组件；卸载删除服务与程序目录并保留数据目录；不支持的系统在改动前退出；容器内的 systemd 与模拟环境不算真实结果 |

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
| V-26 | T-01 驱动 CRUD/事务验证 | T-08 两个入口的 Session 连接、迁移声明互斥与不链接 GORM 的验证 |
| V-36 | T-02 双 Core/操作串行/进程锁，T-08 DB 锁与候选资源 | T-09 正式 enable/update、Core 突退且旧进程仍写入 |
| V-39 | T-01 驱动/事务小验证，T-08 迁移及许可检查 | T-09 首装/resume/更新及显式维护的真实管理入口 |
| V-41 | T-01 外部清单与签名对象 fixture | T-09 真实包验签/来源/更新预检 |
| V-42 | T-02 传输额度、控制进展与取消顺序 | T-06 真实 Service 嵌套与 Bus 双预算 |
| V-32～V-34 | 各任务提供真实示例、构建与故障夹具 | T-10 五目标运行、100 次启停及等价资源比较 |
| V-48 | T-03 Adapter 进程、实例启停、事件归属、extension 校验与依赖解析 | T-09 真实安装/卸载、实例管理与启用插件时的依赖检查 |
| V-49 | T-04 文本匹配与注册校验 | T-05 与真实 Rule、传播组合 |
| V-52 | T-03 manifest 解析与开发入口的依赖检查 | T-09 正式启用、停用、卸载与更新时的检查和确认 |
| V-54 | T-03 locale 映射，T-08 SDK 与 Core 文案目录及 check/pack 检查 | T-09 WebUI 语言切换与真实页面 |
| V-55、V-56 | T-11 Core 服务、策略与确认、会话存储 | T-09 工具策略与会话清除的管理页 |
| V-62、V-63、V-68 | T-02.1 记录、格式、来源标签、进程内过滤与日志文件 | T-09.10 按来源级别的持久化、最近记录、按日期查询、管理 API、CLI 与日志页 |
| V-72 | T-03 Telegram 发送计划、能力与模拟服务；T-15 OneBot 11 映射与模拟服务 | T-15.3 以同一组用例运行两个 Adapter |
| V-73 | T-13.1 `alonabot.Start`；T-16.1 套件基础 | T-16.2 按钮、多轮等待、schedule 与示例插件的测试 |
| V-82 | T-03.8 manifest 角色解析；T-12.8 `@agent` 解析 | T-09.4 同类唯一、`--source` 与 `--replace` 的正式管理流程 |
| V-83 | T-08.1 表单选项图标的构建检查；T-08.7 WebUI 用 `<img>` 显示图标 | T-09.12 WebUI 文件与图标的提供、Admin API 版本检查 |

T-03 的本地 Bot API 故障服务与真实 Telegram 结果分别记录；前者可支持后续开发，真实平台部分仍由 T-10 查证。测试会话、无配置/无 DB 最小插件和 payload fixture 的范围必须写明，不可据此关闭生产认证或数据恢复用例。

## 3. 证据和失败门槛

每项 V-xx 在实施后才填写：测试代码位置、运行命令、实际平台与版本、运行日志/报告路径、通过或失败、所对应的变更。V-35 明确纳入多物理消息的部分成功场景。不能把“规格写明如何测”描述成测试通过。

### 3.1 已运行的分项证据（T-01、T-02）

下表只记录 T-01/T-02 范围内实际运行过的分项；没有任何一项 V 因此整体通过，「未完成」列由后续任务承接。表中未列出的 V 编号以及各行的未完成部分均为「未运行」。

- **运行环境：** darwin/arm64（macOS，Darwin 27.0.0），Go 1.27.1，protoc 36.0；PostgreSQL 分项使用一次性 `postgres:17` 容器（镜像 a63c1ea3f6a3，Docker 29.4.2），用完删除，未残留测试 role/schema。
- **候选提交与命令：** 在 B1 修复分支的提交 e802828 上执行 `go test -race -count=1 ./...`（设置 `ALONABOT_TEST_PG_DSN`）：15 个包、291 个测试（含子测试）全部通过，无跳过；`go test -race -count=3 ./cmd/alonabot/ ./internal/core/supervisor/` 通过；`go generate ./...` 后无差异；五目标 `CGO_ENABLED=0 go build ./...` 通过。
- **未运行：** linux/amd64、linux/arm64、windows/amd64 只做了交叉构建与 vet，**没有运行**；darwin/amd64 只构建。Windows Job Object、LockFileEx 与 Job 清空确认的代码因此没有运行证据。上述 fixture、辅助进程与本机运行都不代表真实平台结果。
- **T-02.1（日志基线）：** 在 goal-65 分支上执行 `go test -race ./...` 全部通过（PostgreSQL 测试未设置 DSN，未运行）；`go generate ./...` 后无差异；五目标 `CGO_ENABLED=0 go build ./...` 通过。只在 darwin/arm64 运行。
- **测试的去留：** 2026-10-25 起仓库只提交集成测试与协议编号检查（编码规范「测试提交规则」）。下表中 `contracts/validate`、`contracts/manifest`、`contracts/release`、`internal/core/host`、`internal/wire` 的 Session 测试（含 `TestUnknownErrorCodeMappedToInternal`、`TestCancelledRequestsReleaseQueueAndBudget`）以及 `sdk` 的单元测试已移出仓库；对应结果是它们在 e802828 上运行时的记录，之后不再由 CI 重跑。仍在仓库中的是 `cmd/alonabot`、`internal/core/supervisor`、`internal/core/process`、`internal/core/state`、`internal/oslock`、`internal/wire` 的帧测试、`sdk/gormdb`（PostgreSQL 测试已随 D-48 删除）与 `sdk` 的 runtime.lock 测试，以及 `contracts/wirespec_test.go`。

| V | 已运行的分项 | 测试位置 | 结果 | 未完成（承接任务） |
| --- | --- | --- | --- | --- |
| V-01 | 市场/本地 ID、缺失/路径穿越/保留名/大写 ID、os/arch 不支持 | `contracts/manifest` | 通过 | 重复 ID 先后顺序、安装来源（T-04、T-09） |
| V-02 | Major 不一致、Minor 取较小值、未知必需能力、UNSPECIFIED/未知 kind、字段存在性、未知错误码对调用者为 INTERNAL | `contracts/validate`、`contracts/wirespec_test.go`、`internal/wire`（`TestUnknownErrorCodeMappedToInternal`） | 通过 | 注册冲突与原子性的完整矩阵（T-04） |
| V-03 | 帧 1 字节～4 MiB 边界、零长/超限/截断/错帧、同一 golden fixture 经 TCP 与 stdio 一致、真实进程在两种传输上的完整生命周期 | `internal/wire`、`cmd/alonabot`（e2e） | 通过（darwin/arm64） | 五平台运行（T-10） |
| V-04 | 伪造/跨插件 Secret、插件 ID 或 digest 不符不回复；认证后帧的 session_id、source_plugin_id、protocol_minor 不符使当前 Session 失败（含准备期 REGISTER 与 Core 帧伪带插件来源）；代次不符的 ACTIVATED 被拒绝 | `internal/core/host`、`internal/wire`、`sdk` | 通过 | — |
| V-05 | 双向相同 call_id、乱序、迟到/重复 Response 只计数、父调用结束后子调用被拒或取消 | `internal/wire` | 通过 | Scheduler 迟到响应释放运行槽（T-07） |
| V-06 | 日志带插件 ID、stdio stdout 污染使会话失败、5 分钟内第 3 次异常退出禁用、退出码 0 不进入崩溃计数 | `internal/core/process`、`internal/core/host`、`internal/core/supervisor`、`cmd/alonabot` | 通过 | 五平台运行（T-10） |
| V-21 | OnStart/OnReady 失败与挂起、准备期硬截止、提交后 Degraded、ACTIVATED 前不开放业务、多实例共享停止预算 | `internal/core/host`、`sdk` | 通过 | 真实配置/DB 回调（T-08）、正式候选恢复（T-09）、Scheduler 停止集成（T-07） |
| V-26 | GORM SQLite/PostgreSQL 的 CRUD/Tx、Migrate 与 Migrations 互斥 | `sdk/gormdb` | 通过（PG 在临时容器上实测） | `sqlitedb` 入口与两个入口共用连接（T-08）；PG 部分随 D-48 移出框架，相应测试已删除 |
| V-32 | 五目标 `CGO_ENABLED=0` 构建 | 交叉编译命令（docs/development.md） | 构建 5/5 通过；运行只有 darwin/arm64 | 五平台运行（T-10） |
| V-34 | 可复用的真实进程夹具；领头进程退出后，留在进程组内的子进程在停止与崩溃恢复两条路径上都被终止并确认消失后才放行 | `internal/core/process`、`internal/core/supervisor`（`TestLeftoverChildReapedBeforeReturn`） | 通过（Unix 进程组） | Windows Job 运行、100 次启停（T-10） |
| V-36 | 双 Core（同进程/跨进程/真实二进制）、按插件串行（含 Stop 与 Launch）、旧实例持有 runtime.lock 时第二个实例在任何用户回调之前失败、Core 被 kill 后 SDK 断链退出并可恢复、相对 data_dir 在两种传输下可启动与重开 | `internal/core/state`、`internal/core/supervisor`、`sdk`、`cmd/alonabot` | 通过 | DB 锁与候选资源（T-08）、正式 enable/update 与仍写 DB 的子进程（T-09） |
| V-37 | Bootstrap 与 REGISTER 超 4 MiB 拒绝、SESSION_ERROR 带失败阶段、无半注册、ACTIVATED 严格先于业务帧 | `internal/core/host`、`sdk` | 通过 | 旧配置先经 OnLoad 再验证（T-08）、T-09 组合 |
| V-39 | Up 与 `_alonabot_migrations` 同一事务、失败不记录成功（SQLite 与 PG） | `sdk/gormdb` | 通过 | 首装/resume/维护许可（T-08、T-09）；PG 部分随 D-48 移出框架 |
| V-41 | release.json 对确切字节验签、重新序列化无法通过、签名有效但内容非法仍拒绝 | `contracts/release` | 通过 | 真实包验签与更新预检（T-09） |
| V-42 | 每方向 256 个/16 MiB 与全局 64 MiB 满额立即拒绝；Response/Cancel 仍可推进；发送前取消的请求移出队列并归还预算、不能借取消越过字节上限；写出中被取消的请求写完才归还预算，Cancel 不越过其 Request | `internal/wire`（含 `TestCancelledRequestsReleaseQueueAndBudget`） | 通过 | Service 嵌套与 Event Bus 双预算（T-06） |
| V-62 | Core 与插件两类来源标签（插件 ID 去掉市场前缀）、text 与 json 格式、终端着色（`script` 伪终端）、非终端与 NO_COLOR 不着色、插件合法 JSON 行（时间、级别、源码位置、字段顺序、嵌套值）、非 JSON 行与 stdout 行记为 `unstructured output` 的 WARN、伪造来源字段被丢弃、消息与键中的换行和控制字符被转义、多行值缩进、生命周期 Hook 的 panic 为一条带 stack 的 ERROR | `cmd/alonabot`（e2e：`TestLifecycleOverTCPAndStdio`、`TestJSONLogFormat`）；本地测试 `internal/core/logging`、`sdk`（`*_local_test.go`，未提交）；着色用 `script -q /dev/null alonabot run ...` 手动核对 | 通过（darwin/arm64） | Adapter 来源（T-03.3）、Handler 的调用字段与 panic（T-03.1）、Bot Token URL 脱敏（T-03.4） |
| V-63 | 全局级别过滤、按来源 debug 与 off（内存中）、Bootstrap 下发级别后 SDK 在进程内过滤（含 OFF 与 UNSPECIFIED） | 本地测试 `internal/core/logging`、`sdk`（未提交） | 通过 | 持久化、运行时修改、最近记录、SSE、CLI 与 WebUI（T-09.10） |
| V-68 | 跨过零点换文件、同日写满后续写编号文件、重启后接着写序号最大的文件、合计上限从最旧删除且不删当前文件、超过 30 天删除、max_size 大于 max_total 拒绝、目录不可写时只报告一次并在换日期时恢复 | `cmd/alonabot`（e2e 读取当天日志文件）；本地测试 `internal/core/logging`（未提交） | 通过（darwin/arm64） | 真实磁盘写满、按日期查询（T-09.10） |

### 3.2 已合并独立切片的验证证据

下列结果仅记录已实际运行的检查。PR 的质量门槛通过不代表整个 V 编号已闭合。

| 切片 | 实现与证据 | 结果 | 尚未闭合 |
| --- | --- | --- | --- |
| T-07.1 / V-20 计算分项 | `contracts/schedule/`；本地测试 `schedule_local_test.go`（未提交）：错过的 tick 被跳过、interval 小于 1s 或超出 Duration 范围被拒、cron 语法（`*`、a-b、列表、`*/n`、a-b/n、星期 0 与 7）及名称/宏/L/越界/步长 0/a>b/字段数错误被拒、日与星期同时受限按 OR、America/New_York 春季跳过 02:30 与秋季 01:30 触发两次、Local 与无效时区被拒、`0 0 30 2 *` 被拒、FromSpec 两个分支与空 trigger | `go test -race -count=1 ./contracts/schedule/` 通过（darwin/arm64） | Core Scheduler 注册/运行槽、迟到响应与停止集成仍归 T-07.2～T-07.5；V-20 整体未闭合 |
| T-12.1 / V-57 Provider 分项 | `agent/provider/`；集成测试 `TestStream`（本地模拟 Provider，`openai`、`openai-compatible` 与 Anthropic 三种协议）：多段文本增量、分段工具调用组装、兼容实现重复 id 与 name 的分段、忽略未知 Anthropic 事件、停止原因与用量（含 Anthropic 缓存 token）、两种 OpenAI 协议的输出上限参数与 `stream_options`、工具结果与文本同在的消息编码、请求编码与认证头、非 2xx 及其错误说明、缺少终止标记、事件中途断开、畸形 JSON、Provider 错误事件及其说明、工具参数不是对象、emit 返回错误、取消 context，错误不含 API key（回显的 key 替换为 `[REDACTED]`） | `go test -race -count=1 ./agent/provider/` 通过（darwin/arm64，2026-10-28 D-64 修正后重跑）；只对模拟服务运行，真实 Provider 未运行 | Agent 对话、流式显示、真实 Provider 验收归 T-12.2 与 T-10；V-57 整体未闭合 |
| T-09.2 / V-27、V-28、V-41 预检分项 | `internal/core/packages/`；集成测试 `TestPreflightAndMoveTo`、`TestPreflightRejects`、`TestPreflightFileModes`（unix）：在临时目录生成真实 tar.gz、release.json 与 Ed25519 签名，经 Preflight 与 MoveTo 驱动；覆盖合法包移入不可变目录、同一目标再次移动返回 ALREADY_EXISTS、签名/来源/平台/大小/摘要不符、manifest 的 id/version/source/平台/entrypoint 不符、`../x`、`/abs`、`a/../../x`、反斜杠、盘符、符号链接、硬链接、字符与块设备、FIFO、大小写冲突、重复路径、归档/条目数/展开字节超限，每项断言暂存目录无残留；权限 0600/0700 且去掉 setuid | `go test -race -count=1 ./internal/core/packages/` 通过（darwin/arm64）；Windows 未运行 | 下载、首次来源信任、安装记录与 CLI 归 T-09.3、T-09.4；V-27、V-28、V-41 整体未闭合 |
| T-08.2 / V-23、V-38 的 raw JSON 分项 | `internal/core/config/`、`internal/core/state/`；集成测试 `TestPendingConfigEdits`：用 `state.Open` 打开临时 `data_dir` 的真实状态库，经 `Service` 公开方法驱动；覆盖 base 不等于 Active 时 Replace/Patch 返回 FAILED_PRECONDITION、CommitActive 新代次后旧 Pending 不再返回且行保留、Patch 叠加在有效 Pending 或 Active 配置上、嵌套未知键与显式 `false`/`0`/`""`/`[]`/`{}`/`null` 与 `9007199254740993` 保留、ImportSnapshot 不改变 Active、`config.json` 缺失为 NOT_FOUND、不是对象为 INVALID_ARGUMENT 且已有 Pending 不变。本地测试（未提交）：Canonical 的键与数字处理、RFC 7396 附录 A 全部示例、state 新方法读写往返；审查者另以本地测试复核过期 Pending 不作 Patch 基础、陈旧提交不覆盖新 Pending、尾随数据被拒 | `go test -race -count=1 ./internal/core/config/ ./internal/core/state/` 通过（darwin/arm64） | 默认值、Schema 与 Secret 归 T-08.3、T-08.4；激活 Pending、显式 reload 与 WebUI/API 入口归 T-09；V-23、V-38 整体未闭合 |

发布阻塞：协议/Session 鉴权、Core 全局权限、数据代次唯一性、回滚/维护的真实性、安装包越界/来源安全及目标五平台构建失败，均不能通过文案豁免。v0.1 只对空闲常驻内存设预算（V-74），初值依据 D-52 记录的实测数据；吞吐与延迟不设通过阈值，T-10 只报告 V-33 的实测结果；**不杜撰 100 MB 等硬性能指标**。

## 4. 实施交接

1. 采用 v0.6-review 技术基线，按 PLAN 的 T-01 开始；任务完成条件及部分验收的后续承接以详细工作包为准。
2. 执行方确认 Go 编译器与目标 OS/arch CI 可用；源码、数据库、Telegram 凭证以及真实测试环境由授权范围决定。
3. T-01→T-02→T-03→T-04→T-05 完成主链路，T-08 可在 T-04 后并行，T-06/T-07 在 T-05 后并行，T-11 在 T-06、T-08 后推进，T-12 在 T-11 后推进，T-13 在 T-04、T-08 后推进，T-09 待相关依赖完成再集成，T-14 在 T-09、T-12 后推进。切片级的前置以 PLAN §3.1 为准。
4. 执行中发现契约缺口，先修对应契约、工作包和用例，在 adr/DECISIONS.md 记录原因；不能在代码中默默更改已接受的协议。

