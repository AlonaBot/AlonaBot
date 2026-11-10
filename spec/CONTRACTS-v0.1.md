# AlonaBot v0.1 技术契约

版本：v0.6-review · 2026-10-21  
状态：**现行实施基线，待实施验证**。本轮按用户“自动完善剩余细节、审阅可行性并修复缺陷”的授权收敛，不提升版本编号。

本文件与 [WIRE-FORMAT.md](WIRE-FORMAT.md)、[MESSAGE-MODEL.md](MESSAGE-MODEL.md) 共同定义首发行为；[PLAN.md](PLAN.md) 是唯一执行入口。保留 D-01～D-36，历史 PDF 与讨论记录只作背景，不覆盖当前规则。规格审阅不是运行时测试报告。

## 0. 范围和明确取舍

v0.1 交付 Core、可安装的 Adapter 机制与官方 Telegram、OneBot 11 Adapter 包（含 D-87 对标补齐的平台能力）、Native Go 插件与 Adapter SDK/子进程、双传输 RPC、命令/文本匹配/事件/权限、延续调用与基础多轮等待、Event Bus、Service Call、调度、插件 HTTP、受管 SQLite（`database/sql` 与 GORM 两种入口）、配置/WebUI、用户可见文案的 i18n、预构建安装/更新与有明确边界的回退，组件的四种角色与同类替换（§9.1、§9.3）、可替换的 WebUI 组件与组件图标（§7.5），Agent 基础（§13）与官方 Agent 组件，统一日志（§14），发行包、容器镜像、install.sh、预装插件、开发运行与库模式（§15），调优参数（§16），以及插件测试套件（§17）。保留 D-01～D-36；Telegram 与 OneBot 11 之外的官方 Adapter、WebUI 交互沙盒、TUI、WASM、插件实例池、蓝绿零停机、持久化的多轮会话状态与暂停 DSL、动态跨插件命令 Router、复杂表单、Telegram webhook 接收，以及由框架托管的 PostgreSQL 等外部数据库（插件可自带驱动自行管理，§6.1）均不在 v0.1 范围内。

运行时容量、期限和配额有固定的缺省值，统一登记为调优参数（§16）。缺省值是初始实现边界，须通过 TEST-MATRIX 实测后调整，不是性能保证；运维与测试可以在启动时覆盖，不提供 WebUI 或逐插件的调参。现行技术细节按本轮用户授权完成，保留 v0.6-review 编号。

## 1. RPC、握手、错误、媒体

### 1.1 连接、Bootstrap 与帧

- 每次 Native 进程启动创建新 Session。TCP 为默认传输：Core 使用 `net.Listen("tcp4", "127.0.0.1:0")` 或明确选定的 IPv6 loopback 地址直接绑定临时端口，持有 Listener 至会话结束；不采用先探测空闲端口、关闭后再绑定的流程。Core 为本次启动生成至少 256-bit 随机 Secret，经最小子进程环境或专用继承句柄传递。Hello 的插件 ID、包摘要、协议版本和 Secret 必须与 Core 启动记录一致；build_digest 为安装时已验证的归档 SHA256，由 Core 从安装记录随本次启动上下文下发供 SDK 回显，不在程序中内嵌自身整包摘要；Core 负责安装与启动前的实际文件完整性检查。Secret 用恒时比较，不接受跨插件凭证。该机制不构成对同宿主同权限恶意程序的沙箱。
- stdio 为可选传输：stdin/stdout 专用协议，stderr 专用普通日志；首帧同样认证。stdio stdout 出现非协议字节时关闭该 Session 并记录插件故障，不能把混入的日志当合法帧。TCP 模式下 Core 可同时捕获插件 stdout/stderr，并标记 plugin ID。
- 两种传输均为 `uint32 big-endian payload length + Protobuf Envelope`；长度不含 4-byte 前缀，合法范围为 **1～4,194,304 bytes**。先验证长度再分配受限缓冲；完整读取前缀和载荷，不能假设一次 Read 得到一帧。长度 0、超限、截断或不可解码的 Envelope 终止当前连接。每方向只有一个帧写入器，多个 goroutine 的帧不能交错。
- Hello 前不得发送其他帧。Protocol Major 不相等，或任一 required capability 不在双方能力交集中，拒绝握手；Minor 选双方支持范围内的较低版本。所有枚举保留零值 UNSPECIFIED，业务要求的枚举和身份字段解码后验证。Envelope 字段编号、控制帧方向及 schema 见 [WIRE-FORMAT.md](WIRE-FORMAT.md)，不另定义平行的 JSON 字节流协议。
- HelloAck 的 `bootstrap` 一次性下发本 Session 的 `generation_id`、原始候选配置 JSON、配置版本、DB descriptor、内部 HTTP 认证凭证和阶段预算。Core 此时只检查配置是合法 JSON 对象并符合大小限额，**不在配置迁移之前按新版 schema 拒绝旧配置**。运行所需 Secret 仅在该插件私有 IPC 中解引用，不继承其他插件或 Core 管理凭证；不得写入日志。Bootstrap 同样受 4 MiB 帧上限限制，超限明确拒绝启动。
- SDK 在 Preparing 中执行配置 `声明式迁移 → OnConfigLoad → 补缺失值 → schema 校验 → DB 迁移门槛核对及许可的迁移 → OnStart`。不存在的可选回调直接跳过。随后 REGISTER 一次性提交完整 RouteDescriptor、规范化配置及其版本、schema hash；声明 HTTP 路由时还提交实际已监听的 loopback endpoint。对新候选，Core 再次验证 Candidate，提取 Secret 为不可变引用后持久化；对已提交代次的普通重启，验证配置与原快照一致并复用原 Secret 引用，不创建新代次或改写快照。Core 校验所有路由和 endpoint 后才返回 PREPARED。明文配置只能经过私有 IPC 和必要的运行内存，不能进入 Journal 或普通配置快照。配置与 DB 细节分别以 §5、§6 为准。
- Preparing/Prepared/Activating 不开放业务 REQUEST；配置、DB 位置和内部 HTTP 凭证由 Bootstrap 提供，不需要准备期通用 Core RPC。认证后的 OnConfigLoad/迁移/OnStart/REGISTER/OnReady 失败使用 `SESSION_ERROR` 携带失败阶段和 ErrorInfo，然后关闭 Session；认证失败直接关闭，不暴露凭证信息或注册残留。

### 1.2 调用、取消与资源边界

- 业务调用的 `call_id` 非零，在同一 Session 的同一发起方向单调递增且不得复用。Request/Response/Cancel 按 `(session_id, initiator_side, call_id)` 配对；控制帧可使用 0。迟到、重复或未知 Response 不第二次完成调用；通常只计数丢弃，scheduled 的已知执行记录按 §4.3 仅据实际结束响应释放运行槽。Response 可同时携带业务结果和 ErrorInfo，例如发送部分成功时保留已确认的回执。
- 插件子请求的 `parent_call_id` 只引用同一 Session 中仍在途的 Core→该插件 Request；Core 从自身记录继承 actor、scope、trace、调用链及父截止时间，不采信插件自报管理员身份。无父 Request 的后台调用只有插件身份；父调用结束/取消后，不得继续使用其用户权限。§2.4 的延续调用由 Core 重新签发，不属于沿用已结束的父调用。跨插件转发是两段独立 Session 请求，Core 保存关联，不要求两段 CallID 相同。
- 普通业务请求默认 **30s**，子请求 deadline 为父截止时间与自身操作截止时间的较早值。SDK 到本地截止时间先完成失败返回，再尽力发 Cancel；对端协作取消 context。取消、超时和断链都不表示撤销已发出的消息或 SQL，Core 不自动重试业务请求。Scheduler callback 使用独立 **15 分钟** job deadline，不受 30s 默认值截断；延续调用同样使用独立的 **15 分钟** 期限（§2.4）。
- 每 Session 每方向最多 **256 个在途业务 Request**；请求和事件的在途及排队 payload 合计最多 **16 MiB**，Core 全部 Session 的此类 payload 合计最多 **64 MiB**。任一准入预算不足立即返回 RESOURCE_EXHAUSTED，不等待普通业务槽位。Event Bus 另按每订阅者 **256 条且 4 MiB** 的较先到达限额执行 drop-newest，见 §3。这里是受管 payload 预算，不是对 Native 进程或 Core 总 RSS 的承诺。
- 读帧、Response 配对、Cancel 和生命周期控制独立于 Handler 执行，不能因某个 Handler 阻塞而停止读取。Response/Cancel/控制使用有界优先通道，不占新业务 Request 的准入槽位，避免满队列时无法返回结果或取消。收发的独立协议缓冲仍受帧上限约束，不允许无限积压；优先通道仍保持同一调用的因果顺序：尚未写出的 Request 被取消则从队列撤出且不再发送；Request 完整写出后才可随后发送其 Cancel。若连有界控制/响应通道或阶段写入 deadline 都无法满足，关闭 Session，并以 UNAVAILABLE 结束未完成调用，不能等待业务 Handler 释放形成死锁。
- 错误包含稳定 `code`、不含敏感信息的 `safe_message`、Core 关联的 `trace_id`。固定错误码和整数值以 WIRE-FORMAT 的 ErrorInfo 为准；未知码按 INTERNAL 暴露，原数字仅保留诊断。跨插件错误不返回 Secret、堆栈或宿主任意路径。

### 1.3 媒体上传与租约

- 普通 RPC 不内嵌大媒体；插件通过 Core 的分块上传 API 得到 `MediaRef`。Begin 声明总大小、MIME 和 SHA-256，Core 检查配额并签发不可预测 `upload_id`；Chunk 必须带当前期望 offset，按顺序追加，每块最多 **1 MiB**。重复、越界、错序的 offset 明确拒绝并终止该上传，不提供同 upload_id 的自动重传或断点续传。插件 SDK 的 `sdk.Upload(ctx, r, mime)` 自行计算总大小与 SHA-256 并分块，插件只给出内容与 MIME；MIME 为空时由 SDK 按内容判断（D-109）。
- Commit 核对实际字节数、声明大小、SHA-256 和 MIME 后才签发引用。MediaRef 绑定 owner Session、允许的消费方/目标范围、size、mime、sha256 和 expires_at；不能用任意本地路径代替。转给其他插件时由 Core 签发新的范围受限 token，不直接放宽原 token。
- 内部首发限额：单对象 **100 MiB**；每 Session 最多 **4 个未完成上传**，其所有未回收的上传和引用对象合计最多 **256 MiB**；Core 临时媒体磁盘总预算 **2 GiB**。配额在接收字节前预留，释放时扣回；磁盘写入失败明确中止，不能产生可用 MediaRef。限额的缺省值登记为调优参数（§16），不增加 WebUI 调参。
- 未完成上传自 Begin 起 **15 分钟**到期，分块请求不延长期限；已提交引用自 Commit 起 **15 分钟**到期。失败、取消、过期或 owner Session 失效时回收无消费者租约的临时对象；正在发送且已被 Adapter 接管的对象由有截止时间的消费租约保护，直到发送结束或操作 deadline 到达。租约仍计入磁盘预算，不能无限延长引用。
- 出站媒体按来源传递，不统一转码（D-88）。Segment 的媒体来源恰有一个：`media_ref`、`incoming_ref`、`https_url`，或插件直接给出的 `inline_base64`（标准 base64，解码后 1 B～1 MiB，超出时 INVALID_ARGUMENT；它随 Send 请求计入 §1.2 的 payload 预算）。插件不能以宿主路径、`file://` 或其他 scheme 作为来源。Adapter 为各平台选择最合适的传递方式：https_url 与 inline_base64 原样交给平台支持的对应形式，media_ref 用平台的流式上传，或由平台端经 Core 签发的临时媒体 URL 拉取；Adapter 不把所有来源统一读成 base64 再发送。
- 临时媒体 URL：Adapter 可以对 Core 为本次发送签发的只读 token 调用 `core.media.url`，得到 `<媒体 URL 前缀>/media/<url_token>` 与到期时间；url_token 为 32 字节随机值的无 padding base64url，只对应这一个对象，到期时间不晚于该次发送的消费租约。Core 在统一端口上只以 GET 与 HEAD 提供该路径，支持 Range，Content-Type 取对象的 MIME，不要求管理认证，不记录完整 url_token；租约结束或到期后返回 404。媒体 URL 前缀缺省为 `base_url`（§15.1），Adapter 可以按实例配置传入另一个前缀（平台端经另一地址访问 Core 时，例如容器网络）。
- Core 不提供任意 HTTPS URL 下载代理。Native 插件可自行获取媒体再上传；Adapter 仅在其公布的平台能力和 URL 策略内使用平台支持的 URL 发送方式。若 Adapter 自己执行下载，必须明确允许的协议、重定向和目标范围，不能默认允许访问宿主私网。此约束不扩展为对 Native 插件网络访问的沙箱。

### 1.4 入站媒体与读取

- 入站图片/文件先使用 `IncomingMedia` 描述符：Core opaque token、声明 size/MIME、readability；Core 私下保存对应 Adapter 实例的 file_id。未知 MIME/大小为 absent，不能伪造已计算的 SHA256。首发不为每个入站媒体预先下载，也不把未取件对象当已提交 MediaRef。readability 只在来源实例的启动响应声明了能力 `media.read` v1 且 supported 时为 READABLE；Adapter 交付时把某个来源标为 UNSUPPORTED 的，Core 保留（D-73）。
- 媒体服务的 Read(token, offset, length) 返回 bytes、next_offset、eof，单次 length 为 1～1 MiB，offset 为非负字节偏移；offset 等于实际大小时返回空 bytes、原 offset、eof=true，大于实际大小时返回 INVALID_ARGUMENT，不产生空成功。错误返回明确 code。首次 Read 由 Core 内置 Adapter 使用受信平台文件 API 按需取件，写入 §1.3 的同一暂存区并校验实际大小/计算摘要后才返回内容；随后按 offset 读取。首次取件截止取调用者/父链与 15 分钟预算的较早者，不能绕过上游 deadline。一次取件失败（包括发起取件的读取者被取消或到期）只返回给这一轮等待的读取者，不记为对象状态；下一次 Read 重新取件，同一次调用内不自动重试（D-73）。插件 SDK 只提供 `sdk.OpenMedia(ctx, src)`（参数见本节末条）：打开时读取第一块，失败时直接返回错误；返回的 `io.ReadCloser` 按 next_offset 依次读取，每次至多 1 MiB（D-109）。
- 同一入站对象由 Core 持有，向接收它的各 Handler Session 签独立的目标受限 token，缓存仅保存一份。token 自签发 15 分钟有效，发送/读取期间用有截止时间的租约保护；一个 Session 失效只撤销其授权，不能删除其他有效消费者仍使用的对象。取件前预留实际可接受的容量，计入 Core 2 GiB 总预算与每个消费 Session 256 MiB 引用预算；未知长度最多按单对象 100 MiB 接收，超限立即失败。
- 超平台下载能力返回 UNSUPPORTED，超配额返回 RESOURCE_EXHAUSTED，事件仍保留不可读取的元数据。Token 仅映射可信 Adapter 已提供的文件标识，不接受用户指定的任意下载地址，不下发 Bot Token/宿主路径。平台文件标识不是插件可用的读取凭证：插件只能用 Core 签发的 token 读取或发送；Adapter 的 platform_payload 原样保留，其中出现的平台标识没有授权作用，Core 也不接受它作为媒体来源（D-73）。
- Segment 可用 `incoming_ref` 作为与 media_ref/https_url 互斥的来源；Send 重新验证目标授权，并核对本次发送的 Target 在原 token 的范围内；目标实例就是该对象的来源实例时，Core 把 Adapter 原先提供的平台文件标识作为 `platform_file` 来源交给 Adapter 直接复用，不取件（D-88），其他情况按需取件为已校验 MediaRef，再执行发送计划。`platform_file` 只出现在 Core 发往 Adapter 的请求中，插件给出时 INVALID_ARGUMENT。取件失败发生在平台发送前，不自动重试业务发送；共享对象到期且无租约后回收。
- 插件经 `sdk.OpenMedia(ctx, src)` 读取，参数是片段中的媒体来源 `message.MediaSource`（D-121）：来源是入站对象或 media_ref 时按其 token 顺序读取；其他来源（https_url、inline_base64）在发出请求前返回 INVALID_ARGUMENT，SDK 不代为下载。只持有 token 的调用方以该 token 作为 media_ref 来源。

## 2. 事件、命令、Rule、Hook 和传播

### 2.1 分发与结果

- 一个入站 Event 对应一条 Core 普通分发链，不同事件可并发。`Normalize → PreEvent → 多轮等待（§2.4）→ 命令/类型匹配与Core权限 → SDK Rule/Handler`；同类候选按 priority 升序，再按 plugin_id/route_id 字典序，顺序不依赖安装或启动时间。
- 保留 Handler 的 `error` 和 `(message.Chain,error)` 两种签名。`c.StopPropagation()` 锁存本次调用的停止标记，SDK 统一返回 `HandlerOutcome{matched,propagation,chain?}`；没有 Stop 默认为 Continue。Context 返回后失效，不允许后台改变结果；显式 Stop 优先于业务错误默认继续。完整编码见 MESSAGE-MODEL。
- 自定义 Go Rule 与 Handler 在同一 RPC 内执行，不另建 Rule RPC 框架。多个 Rule 默认 AND，显式 OR 使用 SDK 组合；False 返回 matched=false，error 标记本 Handler 失败，均不执行目标函数。普通事件的错误记录并继续；panic 由 SDK 在调用边界转 INTERNAL 并记录堆栈，不把失败返回为空成功。
- Core 普通入站事件最多同时执行 64 条分发链，等待队列最多 256 条且 8 MiB，先达到任一限额即拒绝/丢弃并计数；这些 payload 也计入 Core 总预算。不同 Adapter 实例共用此限额，Handler 不能占住读帧循环。已经发生的外部效果不回滚、不自动重放；Telegram 游标语义见 MESSAGE-MODEL。Event Bus 使用独立的非持久队列规则，见 §1/§3。

### 2.2 命令选择、参数与冲突

1. `message.created` 的文本命令以命令前缀开头。前缀由 `core_config.json` 的 `command_prefixes` 决定（§15.1，缺省 `["/"]`），可以用 `command_prefixes_by_instance` 按 Adapter 实例覆盖；空字符串表示不带前缀的命令，多个前缀都能匹配时取最长者（D-72）。匹配前去掉 chain 开头的 reference 片段与指向本账号的 mention 片段。路径 token 大小写敏感、按完整 token 匹配；不以字符串前缀把 `/rssx` 当 `/rss`。首个路径 token 带 `@username` 时，只在它等于当前实例的 account_username 时接受（比较不区分大小写，D-67），实例没有 account_username 时不接受。空白后的子命令按最长完整路径选择。
2. 普通命令/别名优先于文本匹配与 Raw。跨插件同路径以 priority/plugin_id/route_id 稳定选出**一个**胜出路由，并在管理诊断中显示被遮蔽者；同插件重复 route ID 或规范化路径/别名属于注册错误。禁用的已注册命令仍占有其路径，不能使同名低优先级路由接管。管理员保存了命令名覆盖的命令，按覆盖后的命令名参与本条与下一条的匹配、遮蔽与占位（§2.3，D-128）。
3. 一旦识别命令路径，权限拒绝、禁用、Rule false/error 或参数错误都消费该命令，不退到短路径、文本匹配或 Raw。没有识别到任何普通命令时才处理文本匹配，再处理 Raw；每类内部按固定顺序，可 Continue/Stop。命令层结束且没有 Stop 时，message.created 再进入该类型的事件路由（D-44）。其他 event_type 只进入对应事件路由。
4. 参数取自命令路径之后的片段：text 片段按下述规则分词，其间的每个 mention 与 image 片段各自是一个非文本 token，其余片段不参与分词。参数只做分词与类型转换，不执行 shell 展开。空白分隔；单引号内容原样，双引号/未引号中的反斜杠仅转义下一字符；未闭合引号/尾部转义报 INVALID_ARGUMENT。string 保留解码文本；int 为有符号十进制 int64；float 为有限 float64（拒绝 NaN/Inf）；bool 只接受 true/false；这四种类型只接受文本 token。`user` 接受 mention 片段，或被 Telegram `text_mention` 实体完整覆盖的文本 token，值为 external_user_id；`image` 接受 image 片段，值为该片段（含 incoming_ref）。非文本 token 出现在其他类型的位置属于参数错误（D-72）。显式空字符串不等于缺失。
5. 注册描述包含每个参数 name/type/required/default，必需参数在可选参数前，default 只用于缺失位置。raw args 只能是最后一个 string 参数，与该尾部的再次分词互斥，保留最后一个已消费参数后的原始尾部，去掉尾部起始分隔空白；raw 为唯一参数时，从命令路径后的首个非分隔空白开始。末位参数也可以是 `message` 类型：接收最后一个已消费参数之后的全部片段（含非文本片段），去掉起始分隔空白，不再分词，与 raw 互斥。多余参数、溢出或非法格式均失败；Go 函数类型与描述由生成器/显式适配器校验。Handler 用位置参数或一个结构体接收这些参数，两种方式生成相同的注册描述，Go 侧的规则见 §9.1（D-115）。
6. message.created 的 `to_me` 表示消息是否发给机器人（MESSAGE-MODEL §1.1）；SDK 提供对应的 `ToMe` Rule，与其他 Rule 一样组合。
7. 文本匹配包括 Regex 与 keyword、startswith、endswith、fullmatch，均由静态描述声明，作用于 message.created 中各 text 片段按顺序连接的文本。Regex 采用 Go regexp 语法，在启动时编译；非法表达式拒绝注册，捕获结果传专用 match 字段，不推测普通命令参数。keyword 在文本包含任一关键词时匹配；startswith、endswith、fullmatch 分别在文本以任一模式开头、结尾或与之完全相同时匹配。这四种方式可声明 ignore_case，按 Unicode 简单大小写折叠比较；空模式拒绝注册。文本匹配层内的路由统一按 priority/plugin_id/route_id 排序，不按匹配方式分组。
8. 命令解析失败默认发该命令 usage 一次。缺省 usage 由 Core 按参数描述生成；描述可声明 silent，或以插件 `i18n/` 中的文案键声明自定义 usage，Core 按 §5.4 的规则选择语言（D-72）。权限拒绝及系统故障仅保留管理日志，不自动向群组广播错误。
9. 事件路由的 event_type 取 MESSAGE-MODEL §1.1 的标准类型，或 Adapter 扩展事件的完整名 `<adapter_id>.<name>`（MESSAGE-MODEL §1.2，D-71），其他值拒绝注册。注册扩展事件路由不要求对应 Adapter 已安装；未安装或未声明该事件时，管理诊断标出这条路由，它不会被触发。

### 2.3 权限与 Hook

- 路由所需的权限是一组权限名，触发者满足其中任一即可；空集合表示不要求 Actor。权限名包括 Core 内置的 `user`（具有可信 Actor 的用户）、`admin`（`core_config.json` 的 `admins` 中以用户字符串 `<adapter_id>:<external_user_id>` 登记的机器人管理员，§15.1，D-105）、`chat_admin`（任一 Adapter 的会话管理员），以及 Adapter 声明的 `<adapter_id>.<name>`（§8，D-69）。Core 为每个事件计算 Actor 的权限集合：可信 Actor 含 `user`；位于 Core Admin 列表时含 `admin`；Adapter 在事件中给出的平台权限只保留该 Adapter 声明过的名字，并展开声明的隐含关系。`admin` 不自动满足其他权限名，需要 Core Admin 也能执行时显式列出 `admin`。SDK 的缺省权限为 `{user}`。平台权限与 `chat_admin` 只对该事件所在的会话有效，不取得 Core 安装、配置或跨会话权限。没有 Actor 的 job/后台调用只满足空集合。Core 全局禁止由状态库中的两张禁止表组成：按用户字符串禁止用户，按不含话题的会话字符串 `<adapter_id>:<chat_id>` 禁止会话，Core Admin 与两张表都对该 Adapter 的全部实例生效（MESSAGE-MODEL §1，D-105）；被禁止的用户触发的命令、发往被禁止会话的 Send 一律拒绝，任何策略覆盖与 Admin 身份都不能绕过（D-45）。
- 命令策略以稳定 plugin_id/route_id 持久化，权限使用最近显式子项覆盖，否则继承父项。策略中的权限同样是权限名集合，可以引用任一权限名；引用的 Adapter 未安装或未声明该名时保留设置并在管理诊断中标出，这个名字不被任何事件满足；插件/父组/当前项任一 disabled 均拒绝，再与 Core 全局安全限制 AND。运行时修改在下一次派发检查时生效，不承诺撤销已开始的调用。
- 命令名覆盖（D-128）：按路径匹配的命令有一个命令名列表，每个命令名是一条完整路径，缺省为插件声明的路径加别名，第一个是主名。管理员可以按 plugin_id/route_id 保存一份覆盖，它是完整的列表，取代缺省列表；列表至少一项，空列表返回 INVALID_ARGUMENT，删除覆盖即恢复缺省。命令名的格式校验与注册描述中的命令路径相同，列表内不得重复。覆盖只改变匹配用的文字，route_id、命令组归属、策略的作用域、参数与 Handler 不变。同一插件内，覆盖中的命令名与本插件其他命令声明的路径或别名（不论该命令是否被覆盖）或其他命令的覆盖相同时，保存返回 ALREADY_EXISTS 并指出冲突的 route_id。插件更新后新声明的名字与已有覆盖相同时，声明的名字优先，覆盖中的这一项失效并在管理诊断中标出；覆盖一项不剩时该命令使用声明的名字。跨插件同名按 §2.2 的遮蔽规则处理。路由不再注册时覆盖保留并在诊断中标出。usage 回复、命令菜单（§8）与 `core.command.list` 使用生效的命令名。文本匹配与 Raw 路由不能设置，返回 INVALID_ARGUMENT。覆盖在下一次派发时生效，属于命令策略，卸载时归入配置（§9.3）。
- Hook 注册为独立 route kind，phase 仅 PreEvent/BeforeSend/AfterSend，沿用 priority 与稳定排序。PreEvent 只可改变业务 payload 或 Stop，Core 保留的 Actor、Target、来源与 trace 不可改；button.clicked 的 action_id/action_schema_version/value、scheduled 的 job_id/run_id 及平台原消息 ID 也属于不可修改的路由/来源字段，后续匹配使用 Core 保留的原绑定；Hook error 终止本次入站链并记录，不在失去预处理条件时继续。
- 每次显式 Send 或自动返回非空 Chain 都执行 `BeforeSend → Adapter.Send → AfterSend`。BeforeSend 只修改 Chain 或拒绝，禁止修改 Target；变更后重新校验完整 Chain。BeforeSend 错误视为拒绝发送。AfterSend 只观察最终回执，错误记日志，不能改变已确认的发送事实。
- 所有出站 Hook 及其派生的 Service 子调用都继承 Core 的 `send_forbidden` 执行标记；再次 Send 返回 FAILED_PRECONDITION。不能只拦截 Hook 本地函数而允许 A-Hook→B-Service→Send 绕过递归保护。

### 2.4 延续调用与多轮等待

- 由真实入站事件触发、具有可信 Actor 与默认 Target 的 Handler 调用，可以在返回前经 `core.continuation.start` 请求一次延续调用。原调用成功结束后，Core 以同一插件、同一 route_id 向本 Session 发出新的 Handler REQUEST；原调用返回错误时不发出。新请求携带原调用的可信上下文（Actor、默认 Target、scope、trace 与插件链）和原事件。延续调用的期限为独立的 **15 分钟**，不占用 §2.1 的 64 条分发链与等待队列，也不影响原事件的传播：传播仍由原调用的 HandlerOutcome 决定，延续调用的返回值不再参与传播。
- 每个 Session 同时进行中的延续调用最多 **128** 个，超出时 `core.continuation.start` 返回 RESOURCE_EXHAUSTED，为普通派发保留 Core→插件方向 256 个在途请求中的其余名额。
- 延续调用内不能再请求延续调用；出站 Hook、send_forbidden 的调用链与没有 Actor 的调用也不能请求，均返回 FAILED_PRECONDITION。Session 失效、插件停用或代次变化时，尚未发出的延续调用丢弃，进行中的按 §4.2 取消；延续调用不持久化，不在新 Session 补发。
- 延续调用内可以调用 `core.wait`，等待同一 Actor 在默认 Target（含 thread）中的下一条 message.created。等待键为（adapter_instance_id、chat_id、thread_id、external_user_id），同一键已有等待时返回 ALREADY_EXISTS。命中的事件先经过 PreEvent：PreEvent Stop 时不交给等待，等待继续；否则作为 `core.wait` 的结果返回，不再进入普通分发，包括命令匹配。
- 等待的超时由调用方给出，不超过所在延续调用的剩余期限；超时返回 DEADLINE_EXCEEDED。等待也在调用取消或 Session 失效时结束。等待只保存在 Core 内存中，不持久化。
- SDK 有两种请求延续调用的写法，协议相同（D-119）：
  - 注册时声明：命令、文本匹配、Raw、事件与按钮路由的 builder 调用 `.Continue()`。SDK 在原调用中照常执行 Rule，匹配后请求延续调用并成功结束原调用，Handler 整体在延续调用中执行，签名不变。原调用的传播结果固定为 Continue，Handler 中的 `c.StopPropagation()` 不起作用。自动回复型 Handler 返回的非空内容由 SDK 在延续调用结束前经 Send 发出一次。请求延续调用失败时，原调用返回该错误，Handler 不执行。定时任务、Service、Event Bus 订阅、Hook 与 HTTP 路由没有这个方法。
  - Handler 内调用：`c.Continue(fn)`，fn 在延续调用中执行，原调用的其余部分（含 `c.StopPropagation()`）照常决定传播。对已经用 `.Continue()` 注册的路由，Handler 本身就在延续调用中，再调用 `c.Continue` 返回 FAILED_PRECONDITION。

## 3. 跨插件调用、动态 Router 与权限

- v0.1 交付显式 Service Call 与非持久 Event Bus；通过 Core 动态调用其他插件的命令/事件 Router 不在 v0.1 范围内（Q-15），v0.1 只固定稳定 route ID。
- Service 调用的目标可以写作 `@agent`：Core 在 Broker 定位时把它解析为已安装的 Agent 组件（§9.1、§13），访问规则、调用链与 CALL_CYCLE 都按解析后的组件 ID 执行；没有安装 Agent 组件时与目标未激活相同。`@` 不是合法的 ID 字符，不会与组件 ID 冲突（D-61）。
- Service 名为 `plugin_id/service_name`，描述包含输入/输出 schema_id/version、编码与调用规则。默认允许 Core 与本插件，其他插件必须位于发布者声明的 allowlist，或服务明确声明 public；public 仍仅表示已启用插件可调用，不赋予用户 Admin。目标须 Activated，权限/输入检查失败不运行函数。
- JSON 服务只使用 UTF-8 JSON + 明确 schema/version，Protobuf 服务使用对应类型与二进制编码；不得把任意 JSON 冒充 Any。Core 检查已注册 schema 标识/版本、编码、大小与 JSON 语法，不符时只拒绝该次调用（INVALID_ARGUMENT），不关闭调用方 Session；服务声明的版本可以大于 1（D-66）。JSON 服务的输入/输出 Schema 由生成器从 Go 类型生成（D-70），SDK 用生成类型/校验器验证业务字段；Schema 资源随包固定，不允许运行时访问网络 `$ref`。空响应也用声明的空对象/消息，不能把编解码失败吞成空结果。
- 每一跳由 Core 关联可信 Actor、权限 scope、trace、祖先插件链和执行限制。插件发起 Request 的 parent_call_id 只能引用**同一 Session 中仍在途的 Core→该插件 Request**；Core 从此记录派生上下文，不采信插件自报角色。父请求完成/超时/取消会取消未完成子调用；无父调用的后台/主动 Service 只有插件身份，不能沿用已结束用户权限。
- 同一调用链重复插件 ID 返回 CALL_CYCLE，最大插件链深 16，独立并发链互不影响。子调用截止取父 deadline 与自身操作预算的较早者；取消/中断不自动重试。read loop、response 与 admission 规则必须使 Handler 内 Send、A→B→C 及出站 Hook 能完成或明确失败，不能等待自己占据的槽位。
- Event Bus 发布使用带 CallID 的 Core Publish RPC，只确认 Core 已处理发布请求，不确认订阅者执行业务；实际分发使用无业务回执 EVENT。发布者只能使用自己的 `plugin_id/event_name`，订阅者明确注册事件名及 schema/version；离线或满队列丢弃并统计，慢订阅者不阻塞其他插件。payload 不能携带可信 Actor 权限，广播处理发起的业务默认使用订阅插件自身身份。SDK 对每个 Session 的订阅投递额外限制为 4 MiB（`BUS_QUEUE_BYTES` 缺省值），包含排队及正在处理的 EVENT protobuf 载荷；超出时 drop-newest、WARN 并立即 release wire 接收预算。Core 的 `BUS_QUEUE_BYTES` 覆盖仅调整 Core 等待队列，SDK 此上限保持 4 MiB，以给同 Session 的业务请求留出预算。
- SDK 中 Service 与 Event Bus 事件用带类型的声明表示（D-122），协议不变：
  - `sdk.JSON[In, Out](target)` 与 `sdk.Proto[In, Out](target)` 返回 `sdk.Service[In, Out]`，target 是完整的 `plugin_id/service_name`，或 Agent 组件的 `@agent/service_name`。`sdk.Serve(r, s)` 开始注册，之后照旧用 `Access`、`Tool` 与 `Handle`；target 的组件部分不是本插件的 ID（Agent 组件为 `@agent`）、声明为零值或同一声明注册两次时，`Collect` 报错。`s.Call(ctx, in)` 发起调用，返回 `(Out, error)`。
  - `sdk.BusEvent[T](name)` 返回事件声明，name 是完整的 `plugin_id/event_name`。`e.Publish(ctx, payload)` 发布，命名空间不是本插件时照旧被 Core 拒绝；`sdk.Subscribe(r, e)` 开始订阅，路由 ID 缺省为事件名，可用 `.ID(routeID)` 另给。
  - schema_id 与版本仍由 Go 类型决定。声明是普通的包级值，提供方导出后，调用方与订阅方导入同一个包使用；不提供按字符串目标与类型参数直接调用或订阅的函数。

## 4. 生命周期、Config Reload、调度

### 4.1 生命周期与就绪门

- v0.1 的可选生命周期回调为 `OnStart(ctx)`、`OnReady(ctx)`、`OnStop(ctx)`，在 `New()` 中经 `p.OnStart(fn)`、`p.OnReady(fn)`、`p.OnStop(fn)` 登记，各至多一次；没有登记的不执行，无需空函数（DESIGN §D，D-114）。OnConfigLoad 是配置处理回调，不是额外的进程生命周期 Hook。`OnConfigChange`/在线热应用后置；配置应用与包更新均由新 Session 执行。
- 新候选的标准顺序为：创建 Session → Hello/Bootstrap → 配置准备 → DB 迁移 → OnStart → REGISTER/Prepared → Core 原子提交唯一 ActiveGeneration → Activate → OnReady → Activated → 公开业务。首次启用、配置更新和包更新适用此提交路径。普通崩溃重启或 Core 恢复使用新 Session，复用原已 COMMITTED 的 Generation/Config/DB；准备验证通过后依据原提交记录发送 Activate，不新增 Generation 或重复提交 Journal。资源选择、同版本配置一致性检查和回退分支见 §5、§6、§12。
- OnStart 可以使用本 Session 的候选配置/受管 DB，并初始化本地资源；通过 Core 的 Send、Service、公开 HTTP、Scheduler 业务仍关闭。OnReady 是提交后的有限本地准备检查，不能等待自己尚未开放的业务路由或 SDK ready gate，不能同步执行需要 Activated 的业务 RPC；错误明确返回 PLUGIN_NOT_READY，不隐式排队等激活。
- SDK 完整写出 ACTIVATED 后才打开本地 ready gate；OnReady 创建的业务 goroutine 必须等待该 gate，同时响应 lifetime context。Core 在读循环中同步验证并应用 Activated 转换后再处理下一帧 REQUEST；两种传输共享该顺序，**不增加 ACTIVATED_ACK**。本地 gate 只表示允许发起业务 RPC，不等于 Core 已确认外部副作用，断链仍返回 UNAVAILABLE。
- OnStart/OnReady 每次新的 Session 都会执行，插件必须按重启语义编写；不提供一次性通知、只运行一次的迁移或外部发送保证。版本化迁移是否需要运行由 §6 的迁移记录决定。

### 4.2 阶段预算、停止与恢复

| 阶段 | 内部首发预算 | 超时后的处理 |
| --- | --- | --- |
| Hello / HelloAck | 10s | 关闭未认证或未完成握手的 Session，无公开路由 |
| 配置迁移 / 补缺失值 / schema 校验 | 60s | 候选失败，不覆盖 Active |
| DB 迁移 | 15 分钟 | 取消并中止候选；按受管/自管边界恢复或要求人工处理 |
| OnStart | 60s | 候选失败；已提交代次恢复失败则 Degraded |
| OnReady | 30s | 已提交代次 Degraded，禁止新业务派发 |
| 一次停用或整个 Core Shutdown | 总计 10s | 到总截止时间终止普通合作型进程树 |

SDK 对配置、迁移和 Hook 分别执行阶段 watchdog，不能只把 context 传进去便假定用户函数会退出；阶段超时报告 SESSION_ERROR 并结束该进程运行。Core 对 Preparing 另设上述适用阶段预算之和的硬截止，迁移的 15 分钟不被 OnStart 的 60s 默认值截断。这些值是可测试的首发缺省值（调优参数见 §16），不是已测性能基线。

- 停止从 t0 起冻结新入站派发、Scheduler tick 和该插件的新根业务调用；最多 **6s** 排空。已接纳 Request 的响应和合法子调用仍可在原 deadline 与停止 deadline 内完成，避免一冻结就破坏整个在途调用链。
- 排空完成或 6s 到期后，Core Cancel 剩余请求并发送唯一 STOPPING，携带 `min(t0+10s, STOPPING发送时刻+4s)` 的绝对截止时间；Core 终止进程树的总截止仍为 `t0+10s`。SDK 取消 lifetime、使 Session 调用门禁失效并停止 HTTP 接单，再用独立的剩余停止 context 执行 OnStop、关闭 DB/连接并退出；OnStop 及其资源关闭合计最多 **4s**，不复用已经取消的 lifetime context。Core Shutdown 中各插件并行停止，共享同一个总截止时间，不能逐插件串行累加 10s。
- STOPPING 可以紧跟 ACTIVATED 到达：SDK 先发布 ready，再按上一条处理 STOPPING，这种情况按正常停止退出，不算激活失败。
- 主 RPC EOF、协议不可恢复错误或 Core 消失时，SDK 立即使 Session 失效、停止新工作并取消 lifetime，按同一总停止预算退出；此时写 stdout 或 stderr 只返回错误，不能因 SIGPIPE 终止进程而跳过 OnStop；不把旧 Session 重新接回新 Core。Core 先撤销 Session 的所有路由/HTTP 映射，再处理恢复；在途业务 UNAVAILABLE，不重放。
- 正常服务中的异常退出按已确认规则有限恢复：5 分钟窗口内第三次非正常退出自动禁用并保留诊断；前两次在确认旧进程退出后启动新 Session。用户主动停用、正常退出不自动重启；手动启用重置失败窗口。候选配置/升级失败按 §6 回退，不把坏候选当作普通崩溃无限重启；已提交代次的 OnReady 失败为 Degraded，不自动选择旧数据继续服务。

### 4.3 Scheduler

- Core 维护稳定 `plugin_id/job_id`、触发规则和当前运行状态；触发经 REQUEST/RESPONSE 执行 `scheduled` callback，不使用无响应的 Event Bus EVENT。job context 无用户 Actor、无默认 Admin，仅有明确 job scope。
- 任务声明 interval 或 cron，两者互斥；cron 首发固定五字段（分、时、日、月、星期），timezone 为有效 IANA 名称，省略时 UTC。无效表达式在 REGISTER 时拒绝。interval 是至少 1s 的正时长，按固定起点计算，首个 tick 为起点加间隔：静态任务的起点是它首次激活的时刻，保存在 Core 状态库，不随重启改变，间隔在新版本中改变时重置为该次激活时刻（D-125）；睡眠或落后时跳过已错过 tick，只取下一未来触发。cron 的日与星期都受限时按 OR 匹配，其他字段按 AND；cron 求下一个真实时间点，夏令时不存在的本地时刻跳过，重复本地时刻对应的不同 UTC tick 可以分别触发。
- **不同 job 默认并发，同一 job 上一轮未结束时跳过本轮**，不排队补跑。Core 或插件离线期间的 tick 不保存、不补执行；恢复后从未来的下一次触发开始。一个任务超时/失败不串行阻塞其他任务，失败有日志/计数而不自动重试有副作用的 callback。
- job 的调用结果与实际执行结束分开：deadline 到达先返回 DEADLINE_EXCEEDED/Cancel，但 Scheduler 仍保留 running 槽与原调用键；SDK 等用户 callback 实际返回后才发唯一 Response。迟到 Response 只释放任务槽，不恢复业务成功；未收到结束响应时直到确认旧进程退出才释放。仅 RPC 断链不能证明 goroutine 已停止。下一轮只运行未来 tick，不补跑跳过项。
- **动态任务（D-124）。** 插件在注册描述中声明任务名（路由 ID 为 `task.` 加任务名，kind 为 task），运行时经 Core 固定服务增删任务实例：
  - `core.task.set` 创建或替换实例：任务名、key、触发规则（at、interval 或 cron 加可选的 timezone，三者恰有一个）与参数。key 为 1～128 字节的 UTF-8 字符串，在同一插件的同一任务名下唯一，已存在时整体替换并重新计算触发时刻。参数是 JSON 对象，至多 `TASK_PAYLOAD_BYTES`，Core 只校验它是合法的 JSON 对象。任务名未在本 Session 注册返回 NOT_FOUND；触发规则不合法、at 不晚于当前时间、参数超限返回 INVALID_ARGUMENT；该插件的实例数达到 `TASKS_PER_PLUGIN` 且 key 不存在时返回 RESOURCE_EXHAUSTED。interval 与 cron 的规则同上，interval 的起点是创建（或替换）时刻。
  - `core.task.cancel` 按任务名与 key 删除实例，不存在返回 NOT_FOUND；实例正在执行时同时取消这一轮调用（协作式，§1.2），运行槽照上一条的规则释放。`core.task.list` 按任务名返回实例的 key、触发规则、参数与下一次触发时刻，受帧上限约束时分页。
  - 这三个服务只在 Activated 之后可用，不需要父调用链；插件只能操作自己的任务。
  - 触发与静态任务一样经 `scheduled` 的 REQUEST 执行，载荷另带任务名、key 与参数，没有 Actor。同一实例上一轮未结束时跳过本轮。同一插件同时执行的动态任务至多 `TASK_CONCURRENCY` 个，其余到期的实例按到期时间先后排队等待；排队中的周期实例到了下一个触发时刻仍未执行时，只保留一次。一次性实例在派发时删除，执行失败或超时不重试。
- **持久化与错过的触发（D-125）。** 动态任务的定义与周期任务的起点保存在 Core 状态库，以 plugin_id 为键，Core 重启、插件重启与更新后保留，不属于插件数据库的候选与回退。到期时 Core 未运行、插件未处于 Activated 或任务被暂停的触发一律作废，不补执行：周期任务从下一个未来时刻继续；一次性实例过期即删除，记一条 WARN（plugin_id、任务名、key、原定时刻）并计数。排队等待并发名额不算错过，但插件离开 Activated 时队列清空，其中的触发按作废处理。新版本没有注册某个任务名时，该任务的实例保留但不触发，在管理诊断中标出。卸载时的去留见 §9.3。
- **静态任务的暂停与恢复（D-126）。** `core.job.pause` 与 `core.job.resume` 按 job_id 暂停、恢复本插件的静态任务，`core.job.status` 返回是否暂停、上一次与下一次触发时刻；job_id 未注册返回 NOT_FOUND，重复暂停或恢复成功且不改变状态。暂停状态保存在 Core 状态库，跨重启与更新保持。暂停期间的触发作废，不打断正在执行的一轮；恢复后从下一个未来时刻继续，interval 的起点不变。

## 5. 配置、Schema、i18、Secrets

### 5.1 配置真源、缺失值与编辑

- 每插件使用独立 JSON 对象，不加外层 `config`；`config` 只用于 `i18n/*.json` 的翻译根节点。`<data_dir>/plugins/<id>/config.json` 是可编辑快照，运行值以 Core 状态库的 Active Generation 为准。直接编辑后执行 `alonabot reload <id>` 才导入 Pending；`restart` 只重启已激活配置。已有状态库无法读取时必须报错，不能把编辑快照或空对象当成新的 Active。
- 插件配置由与 `plugin.yaml` 同目录的 `config_schema.json` 声明（D-76），没有该文件的插件只有空配置；插件不声明 Go 配置类型。文件是 JSON 对象，键为字段键、值为字段描述，键的顺序就是表单顺序。字段键与模板键是非空的 ASCII 字母、数字、`_` 与 `-`；顶层以 `$` 开头的键保留给 `$version` 与 `$migrations`（§5.5）。Adapter 实例配置用同样格式的 `instance_schema.json` 声明（§8）。键名沿用 AstrBot 的 `_conf_schema.json`，在此之上增加 `secret`、`required` 与 `template` 类型。`type` 必填：

  | type | 值 | 表单 |
  | --- | --- | --- |
  | `string` | 字符串 | 单行输入；有 `options` 时为下拉框，值必须是其中之一 |
  | `text` | 字符串 | 多行输入 |
  | `int` | 整数，范围同 int64 | 数字输入 |
  | `float` | 有限数字 | 数字输入 |
  | `bool` | 布尔值 | 开关 |
  | `object` | 对象，字段由 `items` 描述 | 分组 |
  | `list` | 数组。没有 `items` 时元素为字符串，有 `options` 时为多选，每个元素必须是其中之一；有 `items` 时元素为 `items` 描述的对象 | 增删行 |
  | `dict` | 对象，键为任意字符串，值为任意 JSON | 键值对编辑 |
  | `template` | 只含一个键的对象：键是选中的模板键，值是该模板 `items` 描述的对象，例如 `{"netease": {"cookie": "…"}}` | 模板选择与该模板的分组 |
  | `template_list` | 数组，每个元素的形式同 `template` | 增删行，每行选择一个模板 |

  字段的其他键：
  - `description`（标题，缺省显示字段键）、`hint`（说明）：英文源文案，翻译见 §5.3；`obvious_hint` 为 true 时醒目显示 hint。
  - `default`：该类型的合法值，按下文同样的规则校验；`object` 不写 default，由 `items` 中各字段的默认值组成。
  - `invisible`：为 true 时表单不显示该字段，值照常保存与校验。
  - `options`、`labels`：只用于 `string` 与没有 `items` 的 `list`。options 是存储值；labels 按下标给出显示文案，数量不能多于 options，缺少的显示存储值。
  - `editor_mode`、`editor_language`：只用于 `string` 与 `text`，表单改用代码编辑器，语言缺省为 `json`。
  - `slider`：`{"min": …, "max": …, "step": …}`，只用于 `int` 与 `float`；min、max 必填且 min < max，值超出范围时拒绝；step 可省略，只影响表单。
  - `condition`：`{"兄弟字段键": 期望值}`，全部相等时表单才显示该字段；只影响显示，不影响校验与存储。引用的字段须在同一层，期望值须符合该字段的类型。
  - `templates`：`template` 与 `template_list` 必填，模板键 → 模板描述。模板描述有 `name`（显示名，缺省显示模板键）、`hint`、`items`（可以为空对象）、`display_item`（指向该模板 items 中的一个 `string` 字段，可用 `.` 进入嵌套 object；`template_list` 折叠显示时附上它的当前值）与 `hide_hint_in_list`（为 true 时折叠列表中不显示模板的 hint）。
  - `secret`：为 true 时是 secret 字段，只用于 `string` 与 `text`，不能同时有 `default` 或 `options`；存储与读取见 §5.3。
  - `required`：为 true 时，`string` 与 `text` 须为非空串，`list`、`dict` 与 `template_list` 须非空，`template` 须选择了模板；其他类型不能写 required。
- `alonabot build` 与 `check` 校验 `config_schema.json`：未知键（包括 AstrBot 的 `file` 类型、`_special`、dict 的 `template_schema` 与 `editor_theme`）、键与类型不匹配、default 不合法、labels 多于 options、condition 或 display_item 引用不存在的字段，都报带 JSON 路径的错误。生成器把通过检查的内容嵌入程序；REGISTER 的 `config_schema_hash` 是其规范 JSON（保留键的顺序）的摘要，必须与安装包中的文件一致。
- 每个字段都有值。缺失的键用 `default` 补齐；没有 default 时取类型的零值：`string`、`text` 为 `""`，`int`、`float` 为 0，`bool` 为 false，`list`、`template_list` 为 `[]`，`dict` 为 `{}`，`object` 递归补齐 items；`template` 没有零值，缺失表示未选择模板。只补**缺失键**，显式的 `false`、`0`、`""`、空数组与空对象都保留；list 与 template_list 的元素、template 的模板对象内同样补齐。补齐后按 schema 校验，不接受 `null`：类型不符、超出 int64 或 slider 范围、不在 options 中、template 不是恰好一个已声明的模板键、required 不满足，均返回 INVALID_ARGUMENT，错误指明 JSON 路径，不含取值。
- 未知键默认保留，包括 object、list 元素与模板对象中的未知键；SDK 以完整原始 JSON 对象维护配置，不能用只含已声明字段的结果覆盖整个文档。
- 插件经插件对象的 `p.Config()` 或 Handler 内的 `c.Config()` 读取本 Session 的只读配置（D-114），两者都是整个插件的配置：它在 OnConfigLoad 之后、OnStart 之前就绪，Session 内不变，配置修改由新 Session 应用（§4.1）。按字段类型取值：`String`（string、text）、`Int`、`Float`、`Bool`、`Strings`（没有 items 的 list）、`Object`（object）、`Objects`（有 items 的 list）、`Dict`（dict，值保持 JSON 原样）、`Template`（template，返回模板键与该模板的对象，未选择时模板键为空串）、`TemplateList`（template_list）。`Object`、`Objects` 与模板返回的对象用同样的方法读取其 items。键未在该层声明，或方法与字段类型不符时 panic，消息为英文并指明键的路径；secret 字段返回解析后的明文。
- 子插件的配置是 `config_schema.json` 中与子插件 ID 同名的顶层 `object` 字段，不另设 schema 文件（D-116）。`s.Config()` 返回该对象的只读视图，与 `p.Config().Object(id)` 相同；它可以在注册时取得并保存，读取须在配置就绪之后，就绪前读取 panic。调用过 `s.Config()` 的子插件在 schema 顶层没有同名字段，或该字段不是 `object` 时，`Collect` 报错并指明子插件 ID；没有调用的子插件不要求有这个字段。`c.Config()` 在子插件的 Handler 中仍是整个插件的配置。表单、i18n 覆盖、secret、补缺失值与迁移路径照 `object` 的规则，配置版本每个插件只有一份。
- WebUI 的字段 PATCH 使用 JSON Merge Patch：对象递归合并、数组整体替换、未提交键保持、`null` 表示删除键；删除的键随后按默认值补齐，因此清空字段等于恢复默认值，再经必需字段校验。提交携带所基于的 `base_generation`；旧代次提交返回 FAILED_PRECONDITION，不能覆盖另一项已完成的配置修改。`invisible` 字段不出现在表单中，PATCH 不能因未显示就删除它。因为数组整体替换，WebUI 提交 `list` 与 `template_list` 的行时：行中隐藏（`invisible` 或 `condition` 不满足）的字段与未知字段提交该行未编辑的原值，编辑后又被隐藏的字段也提交原值，新行没有原值时省略该键；行内恢复默认是从该行移除键，由补齐默认值处理（对象差量仍用 `null`）；行中的 secret 须全部重新输入（§5.3）。只按 Schema 字段处理，`dict` 值中用户写的 `null` 原样提交。对象差量恰好形成已设置 secret 的掩码形状时，WebUI 一并提交该对象中未变的键以消除歧义；没有可带的非掩码键时拒绝提交，不能把掩码放进补丁。

### 5.2 配置回调与激活次序

插件可以用 `p.OnConfigLoad(fn)` 登记可选的配置回调，至多一次，只能登记在插件对象上；没有登记时原样使用 raw（D-114）。它是配置数据处理回调，与 `OnStart/OnReady/OnStop` 生命周期回调分开：

```go
func(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
```

- 配置格式版本是 `config_schema.json` 的 `$version`（§5.5），构建时写入包内 manifest 的 `config_version`；旧版本保存在 Core 代次元数据中，不塞入用户配置对象。SDK 先按 `$migrations` 把旧 JSON 迁移到本版本，再调用 OnConfigLoad：它输入经过声明式迁移的完整 JSON 与原始版本号，只做声明式迁移表达不了的转换；SDK 随后补缺失值并按 schema 校验，此时 `Config()` 尚未就绪。插件需要检查业务取值时在 OnConfigLoad 中返回错误，SDK 不另设校验回调。没有登记 OnConfigLoad 时仍保留原始键并执行补值与 schema 校验；不能转换的结构必须明确失败。回调应为可重复的纯配置变换，不执行数据库、文件、网络或发信副作用。
- 顺序固定为：Pending JSON 语法/对象检查 → 停止旧代次并准备候选资源 → HelloAck.bootstrap 下发候选运行 JSON → SDK 声明式迁移 → OnConfigLoad → 补缺失值 → schema 校验 → DB 迁移 → OnStart → REGISTER 回传规范候选 JSON及注册描述 → Core 复核并形成含 secret 引用的不可变快照 → PREPARED → Active Generation 提交 → Activate/OnReady/Activated。不能在 OnConfigLoad 之前用新 schema 否定仍需迁移的旧配置。
- Bootstrap 和 REGISTER 中的完整 JSON 只在所属插件的认证 IPC 内传输；持久化前由 Core 处理 secret，见 §5.3。PREPARED 的配置摘要由 Core 针对持久化快照计算，SDK 将其作为不透明确认值，不自行重建引用后计算摘要。提交前失败保留旧 Active；提交后失败按 §12 进入 Degraded，不把 OnReady 失败写成“任意阶段都能无损回退”。普通 restart/异常恢复只重验当前版本配置并复用原引用/代次；若 OnConfigLoad 在相同格式版本上产生不同配置，返回 FAILED_PRECONDITION，不能借 restart 隐式修改 Active。

### 5.3 Secret 与翻译

- **配置快照存引用，Secret Store 存不可变 secret 版本**。持久化 secret 字段使用 `{"$secret_ref":"opaque-id"}`；secret 字段是 `config_schema.json` 中带 `"secret": true` 的 `string` 或 `text` 字段，可以位于 object、list 元素与模板对象内；空串表示未设置，不生成引用。新值生成新引用，旧代次的引用不可原地改值；普通配置 API 只返回引用/是否已设置，表单据此显示掩码，不能把掩码当成保存或“保持不变”的协议值。省略字段保持原值，显式删除走 §5.1 的校验。
- Core 在构造本 Session 的 bootstrap 前解析其有权访问的引用，SDK 接收正常运行值完成 OnConfigLoad；REGISTER 回传后 Core 按新 Schema **以及此前保留的 secret 路径标记**重新抽取秘密值。旧 secret 字段从 Schema 移除但仍作为未知键保留时继续保密；只有迁移明确删除该键才移除对应路径标记。不能把其他插件的引用解析给当前插件，也不能把含明文的 bootstrap/REGISTER、DSN、配置差异写入日志或普通管理响应。
- 首发 Secret Store 与 Core 状态一起持久化，使用仅运行账号可读写的文件权限/Windows ACL；这保护普通管理访问，不承诺隔离同账号 Native 代码。管理员明文查看须再次认证并审计。失败候选的 secret 仅在未被 Active、Pending 或保留代次引用时回收；保留旧程序/DB 却提前删旧密钥不算可恢复快照。
- 字段与模板的源文案写在 `config_schema.json` 中。`i18n/<locale>.json` 的顶层 `config` 与 `config_schema.json` 同形：字段键 → 覆盖对象，object 与有 items 的 list 的子字段放在 `items` 下，模板放在 `templates` 下，例如 `{"config": {"mode": {"description": "模式", "labels": ["快速", "安全"]}, "music": {"templates": {"netease": {"name": "网易云音乐", "items": {"cookie": {"description": "Cookie"}}}}}}}`。字段可以覆盖 `description`、`hint`、`labels`，模板可以覆盖 `name`、`hint`；其他键或引用不存在的字段、模板时 `check` 报错。查找顺序为当前 locale 的文件 → `config_schema.json` 中的源文案 → 字段键（模板为模板键）；options 的存储值不翻译，显示文案由 labels 提供。缺少翻译不改变配置内容；`invisible` 不是权限控制，不另建字段排序或并发调参页面。

PATCH 语义依据：[RFC 7396](https://www.rfc-editor.org/rfc/rfc7396)。

### 5.4 用户可见文案

- 日志、错误字符串、CLI 输出与 safe_message 只用英文，不经 i18n。发给聊天用户的文案与 WebUI 文案经 i18n 提供，英文为源语言。
- 插件的用户可见文案放在 `i18n/<locale>.json` 的顶层 `messages` 对象中：键为点分隔的 ASCII 标识，值为文案字符串，`{name}` 是命名占位符，`{{` 与 `}}` 表示字面花括号。该文件的顶层键只有 `metadata`（组件的 name 与 description，§9.1）、`config`（§5.3）与 `messages`。
- 文件名是规范化的 BCP 47 标签，例如 `en-US.json`、`zh-CN.json`。`en-US.json` 是源语言，必须包含全部键；其他语言缺少的键回落到 en-US。`check`、`build` 与 `pack` 拒绝以下情况：缺少 en-US、其他语言含 en-US 没有的键、同一键的占位符集合与 en-US 不同、JSON 不合法。
- SDK 的 `T(key, args)` 按以下顺序选择语言：当前 Actor 的 locale（MESSAGE-MODEL §1）→ Core 全局默认 locale（缺省 en-US）→ en-US。每一级先精确匹配，再按主语言子标签匹配；同一主语言有多个文件时取字典序最小者。没有 Actor 的调用从 Core 默认 locale 开始。`T` 只返回字符串（D-106）：`check` 与 `build` 检查代码中以常量写出的键与参数名，键不在 en-US 中、参数不成对或名字集合与占位符集合不同时报错并指出位置；运行时仍未知的键返回键名，缺少的参数保留占位符原文，多余的参数忽略，并在插件日志记一条 WARN，只记键与问题，不记参数值。
- Core 自己发给聊天用户的文案（缺省的命令 usage、§13.2 的确认消息）使用 Core 内置的 en-US 与 zh-CN 目录；插件声明的自定义 usage 取该插件 `i18n/` 的键（§2.2）。选择规则相同。WebUI 提供 en-US 与 zh-CN，按浏览器语言或用户在界面中的选择显示，与机器人回复的语言选择无关。

### 5.5 配置版本与迁移

- 插件配置、Adapter 实例配置与 Core 设置共用同一套版本与迁移规则（D-85），实现放在 `contracts/`，Core、SDK 与构建命令共用：

  | 配置 | 声明 | 存储的版本 |
  | --- | --- | --- |
  | 插件配置 | 插件的 `config_schema.json` | Core 代次元数据（§5.2） |
  | Adapter 实例配置 | Adapter 包内的 `instance_schema.json`，格式与 `config_schema.json` 相同 | `core_config.json` 中该 Adapter 的 `$version` |
  | Core 设置 | Core 内置的 schema | `core_config.json` 的 `$version` |

- `config_schema.json` 与 `instance_schema.json` 的顶层可以写 `$version`（正整数，缺省 1）与 `$migrations`：按 `to` 递增排列的步骤，最后一步的 `to` 等于 `$version`。每步可以有以下操作，按此顺序执行：
  - `rename`：`{"旧路径": "新路径"}`，移动值，可以移到其他嵌套位置；目标已存在时迁移失败。
  - `map`：`{"路径": {"旧值": "新值"}}`，替换字符串或数字值，用于选项改名。
  - `wrap`：`["路径"]`，把不是数组的值包成单元素数组。
  - `remove`：`["路径"]`，删除键；被删除的值是 secret 时同时移除其路径标记（§5.3）。

  路径以 `.` 连接字段键，`*` 表示任意数组元素，template 与 template_list 的值中以模板键作为一段，例如 `sources.*.rss.url`；路径不存在时跳过该操作。示例：

  ```json
  {
    "$version": 3,
    "$migrations": [
      {"to": 2, "rename": {"hello": "greeting", "proxy": "network.proxy"}},
      {"to": 3, "map": {"mode": {"quick": "fast"}}, "wrap": ["keywords"]}
    ],
    "greeting": {"description": "Greeting", "type": "string"}
  }
  ```

- 迁移从存储的版本依次执行到当前版本。新增字段不需要迁移，补缺失值时得到默认值；从 schema 中删除的字段不会被自动删除，作为未知键保留（§5.1），需要删除时写 `remove`。插件的 `OnConfigLoad` 在声明式迁移之后执行（§5.2），只用于声明式迁移表达不了的转换；Adapter 实例配置只有声明式迁移。
- `alonabot build` 与 `check` 检查迁移：版本连续，最后一步的 `to` 等于 `$version`；`rename` 与 `wrap` 的目标路径、`map` 的新值在当前版本的 schema 中存在且合法。
- 迁移不能丢失配置：
  - Core 写回迁移后的 `core_config.json` 之前，把原文件复制到 `<data_dir>/backups/core_config.<旧版本>.<时间>.json`；插件的旧配置保存在旧代次的快照中（§6）。
  - 迁移或迁移后的校验失败时：Core 设置——Core 拒绝启动，原文件不变；Adapter 实例配置——在停止旧版本之前拒绝这次更新，开发运行中只是该实例不启动；插件配置——候选失败，旧 Active 继续（§5.2）。错误指明文件、版本与 JSON 路径，不含取值。
  - 存储的版本高于当前程序支持的版本（降级）时按上一条的失败处理，错误写明两个版本号；不猜测格式，也不删除无法识别的键。
- Core 设置的迁移内置在 Core 中，随 Core 升级发布。Adapter 实例配置的迁移在 Adapter 更新的准备阶段由 Core 执行，成功提交后写回 `core_config.json` 并更新该 Adapter 的 `$version`。

## 6. 数据库、更新、事务与回退

### 6.1 数据所有权、连接与默认依赖

- Core 元数据固定使用独立 `<data_dir>/state.sqlite`，记录安装/来源、代次、Journal、动态权限和 secret 引用关系，插件不复用其 DB 或凭证。Core 连接池固定为 1，采用 WAL、`synchronous=FULL`、`foreign_keys=ON`、`busy_timeout=5000`；每个物理连接都应用连接级设置，初始化核对实际生效值，失败停止启动。Core 的元数据事务只包含必要数据库操作，不在持锁事务中等待插件 RPC。状态库错误/损坏不能触发重建空库。
- 受管 SQLite 有两个入口，共用当前 Session 的同一个连接：`sqlitedb.Open(ctx) (*sql.DB, error)` 返回标准库 `database/sql` 对象；`gormdb.Open(ctx) (*gorm.DB, error)` 在同一连接上返回原生 GORM 对象。插件只导入用到的包，不用 GORM 的插件不链接 GORM。重复调用复用该连接，SDK 停止时关闭。事务回调须使用传入的 tx，不能在事务里等待同一单连接池的另一条独立查询。连接默认单连接、WAL、FULL、外键开启和 5s busy timeout；这不串行化整个插件的 goroutine/任务。连接错误直接返回，不回退为空数据库。
- SQLite 文件按 generation 分开：`plugins/<id>/generations/<g>/db.sqlite`。首次安装创建候选文件；配置 reload 和版本 update **一律创建独立 SQLite 候选副本**，不因“只改配置”而复用旧 DB，因为 OnStart 也能写入。普通 restart/异常恢复启动现有 Active，不创建新配置代次。框架外文件、任意外部数据库和出站副作用不纳入受管回退。
- 框架只托管 SQLite。需要 PostgreSQL、MySQL 等外部数据库的插件自行导入驱动并管理连接，连接参数作为插件的 secret 配置项由管理员填写；这些数据属于插件自管的外部数据，Core 不创建、不校验、也不备份外部数据库。迁移或 OnStart 会写外部数据库的插件须在 manifest 声明 `prepare_effects=external`（§9.1），按 §6.3 的维护门槛更新。
- `sqlitedb` 使用 `modernc.org/sqlite` 的 `database/sql` 驱动；`gormdb` 另用 `gorm.io/gorm` 与 `github.com/libtnb/sqlite`，经后者的 `Conn` 设置复用同一连接。Core 元数据经 GORM 访问，复用同一 modernc 连接（D-68）；schema 由唯一的初始迁移明确定义，不用 AutoMigrate 推断或改写表结构。Backup API 直接使用 modernc 驱动。T-01 已验证 GORM 组合的 CRUD/Tx/迁移、WAL Backup/恢复和五目标 `CGO_ENABLED=0` 构建，实际版本固定在 `go.mod/go.sum`，见 docs/development.md；不新增通用驱动兼容框架。

### 6.2 可选迁移接口

迁移回调在插件 SDK 进程内执行，只作用于受管 SQLite；Core 管理启动阶段、候选目录和维护准入，不跨进程直接调用 Go 函数。`sqlitedb` 与 `gormdb` 各提供一组形状相同的登记函数，只是数据库参数的类型不同，在 `New()` 中对插件对象调用（D-114）。自管迁移与辅助迁移互斥，四个函数合计至多调用一次，都没有调用即不自动迁移：

```go
// sdk/sqlitedb
func Migrate(p *sdk.Plugin, fn func(ctx context.Context, db *sql.DB) error) // 自管迁移

type Migration struct {
    ID uint32
    Up func(context.Context, *sql.Tx) error
}
func Migrations(p *sdk.Plugin, list []Migration) // 辅助迁移

// sdk/gormdb：同名函数，Migrate 的回调接收 *gorm.DB，Up 接收事务中的 *gorm.DB
```

- 迁移只能登记在插件对象上，不能登记在子插件上（D-116）；调用多于一次时 `Collect` 报错。SDK 自身不 import GORM，只导入 `sqlitedb` 的插件不链接 GORM。两个包的辅助迁移共用同一张完成表与同一套 ID 规则。`build` 按插件包及其本模块依赖中对这四个函数的调用写入 manifest 的 `migration_mode`（§9.1）；SDK 在准备阶段核对实际登记的方式与 manifest 一致，不一致时准备失败。
- 自管 Migrate 在 SDK 准备阶段调用，由作者处理版本记录和重复启动的幂等性。辅助 Migrations 的 ID 必须非零、唯一、递增；发布后不能把旧 ID 改成另一项操作。SDK 只执行尚未完成的 ID，每一步 Up 与完成 ID 一起写入**当前插件 DB 的同一事务**，失败不记录成功；完成表固定为框架保留名 `_alonabot_migrations`，业务代码不得挪作他用。不得把迁移成功记录只写入 Core 状态库而假称两库原子提交。
- 迁移门槛核对及本次启动原因允许的迁移均在 OnStart 之前完成；同一插件的启停、配置、迁移、更新和恢复按 §12 的管理锁串行。事务内不支持的 SQL 直接失败，不能偷偷转为事务外执行；需要这种操作的作者选自管迁移并进入相应维护流程。Native 仍能自行调用 GORM/SQL，自动回退只覆盖契约允许的受管路径。

### 6.3 SQLite 候选与维护

- 受管 SQLite：下载/校验候选 → 冻结旧版新派发并排空 → 停止旧进程及合作型后台资源，确认连接释放 → 使用 SQLite Backup API 将旧 Active DB 备份到独立候选文件 → SDK 配置校验/迁移/OnStart → PREPARED → 同一 Core 事务更新 ActiveGeneration 与 COMMITTED → Activate/OnReady/Activated 后开放路由。副本、程序和不可变配置先持久化再提交引用；备份/空间/校验失败不提交。不能只拷贝运行中 WAL 数据库的主文件。
- 提交前失败：保留旧 DB/程序/配置，重新启动旧版的**新 Session**。COMMITTED 后失败：新代次仍为权威，标记 Degraded 并关闭新派发；不自动切回已过时的旧数据。Core 崩溃恢复按 Journal 和 Active 指针判定，任意时刻只有一个代次可公开接单；完整时序见 §12。
- Bootstrap 的 database_descriptor 由 Core 下发 `activation_reason`（first_install/resume/reload/update/rollback）与 `manual_migration_approved`；后者只对经管理员显式维护许可的本次 Session/Generation/DB 为 true，不由插件请求自行设置。SDK 不需要在 Preparing 发起通用业务 RPC 获取许可。
- 声明 `prepare_effects=external` 的插件在准备期会改动受管 SQLite 之外的数据，候选副本无法隔离。首次安装照常准备；已有安装的 reload、update 与 rollback 返回 NEEDS_MANUAL_MIGRATION。管理员显式进入维护、停止插件并自行完成外部数据的备份后，Core 才以 `manual_migration_approved=true` 准备候选；失败保留停机状态，按该插件的恢复步骤处理，不自动重跑未知副作用；成功后照常完成 Prepared 与激活。普通 restart 与崩溃恢复（`activation_reason=resume`）不建立候选，不受此门槛限制。`prepare_effects` 是可信代码的声明，Core 不能静态证明任意 SQL 没有副作用。
- 人工整机备份须在 Core 与插件停止后覆盖完整 data_dir，以及插件自管的外部数据；单独备份 state.sqlite 不包含插件业务数据。受管 SQLite 的在线副本仍使用 Backup API。不能删除 Active、恢复中 Journal 或保留回退代次引用的程序/DB/配置/secret；候选失败清理只作用于无引用资源。

依据：[GORM 连接文档](https://gorm.io/docs/connecting_to_the_database.html)、[libtnb SQLite 驱动](https://github.com/libtnb/sqlite)、[modernc Backup API](https://pkg.go.dev/modernc.org/sqlite#Backup)、[SQLite Backup](https://sqlite.org/backup.html)、[SQLite PRAGMA](https://www.sqlite.org/pragma.html)、[database/sql](https://pkg.go.dev/database/sql)。这些主源支持选型与约束，AlonaBot 集成测试仍待实施。

## 7. Native 信任、HTTP 与来源边界

### 7.1 Native 进程与安装边界

- v0.1 Native 插件、Adapter 与 Agent 组件是**可信本地代码**；WebUI 组件在 Admin 源中运行并持有管理员会话（§7.5），同样可信。独立进程用于生命周期和故障隔离，不提供跨平台文件/网络/exec 沙箱；Secret Store、loopback、精简环境和工作目录都不能阻止同 OS 用户权限的恶意插件。安装界面展示实际来源和该权限边界，签名只证明发布来源。
- 子进程只接收本插件运行需要的配置/Secret、Session 凭证及必要环境，移除 Core 管理 token 和其他插件凭证。工作目录选在该插件目录；不得将工作目录描述为文件访问权限限制。
- 所有组件 ID 必须通过安全 slug 校验并限制路径长度，市场 ID 按角色使用 `alonabot_plugin_` 等前缀（§9.1）；精确 manifest 字段及长度约束见 §9。本地 ID 可不带市场前缀但不可包含路径分隔符、`..` 或其他非法路径字符。跨插件 ID 冲突按 Core 持久化安装顺序 first-wins；同插件重复 Handler/HTTP route 报错。更新同时核对 ID、绑定来源和发布者身份，不能用相同 ID 接管已安装包。
- 首发平台为 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64。CGO_ENABLED=0 优先；CGO 插件由发布者提供真实可运行的平台二进制。Core 与插件通过版本化 RPC 协作，不要求相同 Go patch 版本；不使用不能卸载且跨平台受限的 Go `plugin.Open`。
- 普通子进程树由 Windows Job Object（由 Core 独占持有、关闭时终止所管理进程）、Linux/macOS 进程组管理，必要的平台能力以真实测试为准。不能宣称可阻止恶意 detach；受管 SQLite 插件须关闭所有受管写入者，私自脱离管理的外部 DB 写入不属于自动回退保证。

### 7.2 统一端口的 HTTP 代理

- D-25 保持不变：只有 Core 公开 HTTP(S) Listener；插件作者注册 Go `http.Handler`，不选择、配置或向部署用户开放内部端口。SDK 仅在存在 HTTP route 时创建 `127.0.0.1:0` Listener，并向 REGISTER 报告实际 endpoint；无 HTTP route 不创建 Listener。SDK 持续持有该 Listener，不能释放端口后另起服务。
- Core 只接受当前 Session 注册的 loopback endpoint，使用标准 `httputil.ReverseProxy` 和 `Rewrite` 构造转发；不接受外部请求指定代理目标。外部路径固定 `/plugins/{id}/...`，Core 完成 ID、method、route 和权限匹配后去掉 `/plugins/{id}` 前缀，插件收到以 `/` 开头的相对应用路径。路由 path 以 `/` 开头且为规范形式：`path.Clean` 的结果，原路径以 `/` 结尾时保留一个尾斜杠并表示前缀路由（`/` 匹配全部路径），其余为精确路由；Core 与 SDK 用同一份未清理的请求路径匹配，两侧选出同一条路由（D-65）。Core 保留 Admin/API/healthz 和内部认证路径，不让插件覆盖。
- Bootstrap 中的 HTTP 凭证与 Session/Generation 绑定。Core 删除外部可伪造的内部身份/认证头，并在 Rewrite 阶段注入受信凭证；SDK 在进入用户 Handler 前校验。Core 不向插件转发 admin Cookie、admin Authorization 或未验证的 Forwarded/X-Forwarded 身份信息；路由停止、Session 失效或代次变化时撤销代理映射，SDK 关闭 Listener。
- 框架托管 HTTP 默认要求 Core 授权。公开 OAuth/Webhook 在描述中显式声明；插件必须验证对应 state、签名或 nonce，不能把“public”理解为免除业务来源验证。回调地址由已配置 Base URL 生成，不以任意 Host 或 Forwarded 请求头拼接；允许的 Host 与可信反代来源由 Core 全局设置验证。
- Adapter 包也可以注册 HTTP 路由（D-94），规则与插件相同，外部路径同样为 `/plugins/{Adapter 包 ID}/...`；用于平台端主动连入，例如 OneBot 11 反向 WebSocket。代理转发 WebSocket Upgrade 请求并在升级后双向转发字节，连接寿命不受普通请求期限限制，路由撤销或 Session 失效时 Core 关闭已升级的连接。
- HTTP 代理与普通命令/事件/Service 的 Protobuf RPC 分工不变；Native 插件自行建立的框架外 Listener 不享受统一入口和生命周期托管承诺。

### 7.3 管理界面与插件页面的浏览器边界

- v0.1 只有 **一个 Core Listener、一个公开端口、一个主机名**（`base_url`，D-84）：管理界面（WebUI 组件，§7.5）、Admin API 与插件路由 `/plugins/{id}/...` 都在这个主机名下，不需要为插件页面准备第二个域名。Core 验证 Host，不把客户端提供的 Host 当信任依据；远程反代须保留受控 Host 并按全局可信反代配置处理转发信息。
- 管理会话 Cookie 使用 host-only、HttpOnly、`SameSite=Strict`，并在 HTTPS 下使用 Secure 与 `__Host-` 前缀；Admin API 的状态变更验证 CSRF。公开安装/管理接口不能绕过管理员认证。
- 插件路由的全部响应由 Core 加上 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups`（不含 `allow-same-origin`）与 `X-Content-Type-Options: nosniff`，并去掉插件设置的 `Set-Cookie`；插件不能覆盖这些响应头。插件返回的页面因此运行在不透明源中：无论被嵌入还是直接打开，都读不到管理 Cookie、WebUI 的存储与 DOM，它发出的请求也不带管理 Cookie（D-15）。
- WebUI 用带 `sandbox` 属性的 iframe 嵌入插件页面，页面只经 WebUI 提供的 postMessage 桥与外部交互：请求本插件的 HTTP 路由，读取当前语言的文案与界面主题。WebUI 只转发方法与路径落在该插件已注册路由内的请求，用管理员会话与 CSRF 头发出，再把响应交回页面；页面拿不到管理会话、CSRF token 或其他插件的路由。桥的消息格式记录在 Admin API 参考中（§7.5）。
- 非公开的插件路由要求有效的管理会话，非 GET 请求与 Admin API 一样验证 CSRF；公开的 JSON API、OAuth 回调与 Webhook 按 §7.2 声明，不要求会话。插件页面与管理界面不同源，插件页面之间也互不同源（各自为不透明源），但插件仍是可信的本地代码（§7.1）。

### 7.4 首次管理认证与恢复

- 首发使用一个 Core 管理凭证，不引入多用户账号体系。Core 首次启动且尚无管理凭证时生成 32-byte 随机 token，状态库只存 SHA256（D-82）：stdin 与 stderr 都是终端时只在该终端显示一次；不是终端时（系统服务、容器）写入 `<data_dir>/admin_token`，文件只有运行账号可读，日志只输出文件路径，第一次登录成功后 Core 删除该文件。`alonabot admin-token show` 在本机读出并显示这个文件，文件不存在时报错，供没有 shell 的容器使用；`alonabot admin-token init` 在启动前离线完成同样的初始化，只向终端显示。`alonabot admin-token reset` 仅在 Core 停止并持锁后替换该凭证、递增认证代次并撤销旧管理会话。`init` 与 `reset` 先取得同一 data_dir 的 core.lock；这些命令都不走公开 HTTP，不把凭证写入日志。
- 浏览器通过 POST `/api/auth/login` 提交 token，Core 验证 Origin/登录 CSRF 后恒时比较摘要，成功签发随机的、固定 **8h** 到期的管理会话。登出撤销该会话及其派生授权；会话检查同时校验认证代次。修改管理状态继续采用 §7.3 的 Cookie/CSRF 规则，明文查看 secret 使用当前 token 再认证。
- 管理 CLI 仅使用用户显式配置的 Bearer token 访问 Core，不从插件环境继承凭证；该认证不依赖浏览器 Cookie。非 loopback 的管理入口必须 HTTPS，可由已配置可信反代终止 TLS；纯本机 loopback 开发允许 HTTP。Core 管理 token 与聊天事件中的 Core Admin 列表是不同入口，平台角色不能兑换管理会话。

### 7.5 组件静态文件

- Core 直接提供两类组件静态文件，只读取不可变包目录中的文件，不经组件进程（D-62、D-63）：
  - **WebUI 组件的文件**：除 Admin API（`/api/`）、插件路由（`/plugins/`）与 Core 保留的内部路径外，GET 请求按路径取已启用 WebUI 组件静态根目录中的文件；路径不存在且最后一段不含扩展名时返回 `index.html`，供单页应用的前端路由使用。没有已启用的 WebUI 时返回 503 与英文纯文本提示，说明用 `alonabot install` 安装 WebUI 组件。
  - **组件图标**：manifest 的 `icon`（§9.1），只提供声明过的路径。类型按扩展名取 SVG、PNG 或 WebP，其他扩展名在包校验时拒绝。响应带 `Content-Security-Policy: sandbox; default-src 'none'; style-src 'unsafe-inline'` 与 `X-Content-Type-Options: nosniff`，要求管理员会话；WebUI 只用 `<img>` 引用图标，直接打开图标 URL 时也只在隔离的源中渲染，SVG 中的脚本不会在 Admin 源中执行。
- Admin API 是带主版本号的公开契约，接口、主版本与插件页面桥的消息格式（§7.3）记录在 Admin API 参考（`spec/ADMIN-API.md`，T-09.12 交付）；同一主版本内只做向后兼容的增加。WebUI 在 manifest 中声明支持的主版本范围（§9.1），Core 的主版本不在范围内时拒绝安装与启用；Core 升级后不兼容的已安装 WebUI 不再提供文件，Core 返回说明版本不兼容的提示。
- WebUI 持有管理员会话，能执行全部管理动作，安装与替换界面展示这一点；签名只证明发布来源（§9.2）。

## 8. Adapter、Target 与消息

- Adapter 与插件一样以安装包分发，在 Core 之外以独立 Native 子进程运行，复用插件的握手、Session、准备/激活/停止状态机、代次、配置、Secret、安装与来源校验（D-26）。官方 Telegram Adapter 也是这样一个可卸载的安装包，Core 不静态链接任何平台实现。manifest 以 `role: adapter` 和 `adapter_id`（平台类型，例如 `telegram`）区分，同一 `adapter_id` 同时只能安装一个包；社区 Adapter 可以按 §9.3 的 `--replace` 替换同一 `adapter_id` 的官方 Adapter。
- 同一个 Adapter 包只运行一个进程，进程内承载该包的全部账号实例。实例配置（含凭证）写在 `core_config.json` 的 `adapters` 中（§15.1），由 Core 按 Adapter 包内 `instance_schema.json` 迁移、补缺失值并校验（§5.5），经 `adapter.instance.start`/`adapter.instance.stop` 单独启停；启停一个实例不重启进程，也不影响同包的其他实例。停用整个 Adapter 即按插件的停止流程停止该进程，全部实例随之停止。进程异常退出按插件的有限重启策略处理，重启后 Core 只重新启动仍处于启用状态的实例。
- Core 只接受 Adapter 为它在本 Session 中已启动的实例交付的事件：Target 与事件中的 adapter_id、adapter_instance_id、account_id 必须与该实例一致，否则拒绝并记诊断。event_id 由 Core 生成；Actor 的权限集合由 Core 按 §2.3 计算，Adapter 只提供平台给出的身份与平台权限；平台权限中未声明的名字由 Core 丢弃并记诊断。同一 `adapter_id` 下的同一平台账号只能有一个已启动实例（唯一入站 owner），由 Core 依据 `adapter.instance.start` 返回的 account_id 检查。
- 接收游标由 Core 持久化：Adapter 每交付一条事件都附带交付后的不透明游标，Core 先写状态库，再接纳事件并应答；Adapter 收到应答后才处理下一条。没有事件时的游标变化（例如空闲轮次重建）使用单独的游标保存调用。实例启动时，Core 把已保存的游标交给 Adapter。
- 平台凭证只在 `adapter.instance.start` 中下发给所属 Adapter 进程，不进入插件进程、日志或普通响应。入站媒体的平台文件标识留在 Adapter，Core 以 IncomingMedia token 映射，首次 Read 时经 `adapter.media.read` 取件；出站媒体由 Core 为该次发送向 Adapter Session 签发只读 token，Adapter 经 `core.media.read` 读取。
- 官方 Telegram Adapter 用 getMe 验证真实账号与实例配置。v0.1 只支持 long polling 接收，已有 webhook 时明确报错，不自行删除。实例配置为 `token`（Secret，必填）、`api_base_url`（缺省 `https://api.telegram.org`，用于本地 Bot API 服务或模拟服务）、`file_base_url`（可选，文件下载地址的前缀，缺省由 api_base_url 推出，D-99）、`proxy_mode`（`none`、`core` 或 `custom`，缺省 `none`，D-101）与 `proxy`（Secret，http、https 或 socks5 URL，可含用户名与密码，只在 custom 时使用并且必填）。代理模式决定该实例的全部 Bot API 请求（含 getFile 与文件下载）如何发出：none 直连，不读取代理环境变量，也不使用 §15.1 的全局代理；core 使用 Core 为子进程设置的代理环境（§15.1）；custom 经实例的 `proxy` 发出，`proxy` 为空或不合法时实例启动返回 INVALID_ARGUMENT。非 custom 模式下 `proxy` 不生效，原值保留。代理连接失败按网络错误处理，接收退避重连，不改为直连。
- 官方 OneBot 11 Adapter（D-46、D-94）有两种连接方式，同一连接都承载 API 调用与事件上报（通用端点）：`forward` 由 Adapter 主动连接实现端，`reverse` 由实现端经 Core 统一端口连入 Adapter 注册的 HTTP 路由（§7.2），外部地址为 `<base_url>/plugins/<Adapter 包 ID>/ws/<instance_id>`。实例配置为 `mode`（缺省 `forward`，可选 `reverse`）、`ws_url`（forward 必填，`ws://` 或 `wss://`）、`access_token`（Secret；forward 可选，握手时以 `Authorization: Bearer` 发送；reverse 必填，Adapter 以恒时比较校验实现端的 `Authorization: Bearer` 或 `access_token` 查询参数）、`self_id`（reverse 必填，十进制 QQ 号）、`media_url_base`（可选，实现端访问 Core 媒体 URL 时使用的前缀，缺省为 `base_url`，§1.3）与 `media_dir`（可选，`local` 为 Adapter 可写的共享目录，`remote` 为实现端看到的同一目录，缺省与 local 相同）。forward 模式用 `get_login_info` 取得账号；握手被拒（401、403）时实例致命失败；连接断开按与 Telegram 相同的有界退避重连。reverse 模式的 account_id 取 `self_id`，实例启动时不等待连入；握手的 `X-Self-ID` 不等于 self_id 或 `X-Client-Role` 不是 Universal 时返回 403，已有连接时返回 409；连入后调用 `get_login_info` 核对账号并取得展示名，不符时关闭连接并记诊断；断开后实例继续等待，不报告致命错误，未连接期间 Send 返回 UNAVAILABLE。两种模式重连都不重放已提交的 Send。上报事件的 `self_id` 必须等于该实例的账号，否则丢弃并记诊断。OneBot 11 没有事件重放：断线期间的事件不补收，游标保持为空。只接受数组格式的消息上报，收到字符串（CQ 码）格式时实例致命失败，并在 safe_error 中提示修改实现端的上报格式。映射、能力、发送计划与动作见 MESSAGE-MODEL §7。
- Target 为 adapter_id/adapter_instance_id/account_id/chat_id/thread_id?。插件的公开 API 用会话字符串 `<adapter_id>:<chat_id>[#<thread_id>]` 与可选的实例名表示目标（MESSAGE-MODEL §1，D-102），指向某条消息时用消息句柄（MESSAGE-MODEL §4，D-104）：插件发出的 Target 只带 adapter_id、chat_id、thread_id 与可选的 adapter_instance_id，Core 在准入前只补空字段。缺少实例时，调用链上在途事件的目标属于同一 Adapter 与 chat_id 则取其实例，否则取该 Adapter 的缺省实例（只配置了一个实例时为它，或 §15.1 的 `default_instance`），都没有时 FAILED_PRECONDITION；实例已登记时按其当前绑定的账号补上 account_id，未登记时不补，由下文的实例检查按原顺序拒绝。插件给出且不符的 adapter_id 或 account_id 仍由实例绑定检查返回 INVALID_ARGUMENT。引用片段的 target 以补齐后的发送目标为在途目标，按同一规则补齐；媒体授权、编辑、撤回与平台动作的目标同样补齐。默认目标仅来自入站真实事件；无默认目标且未指定 To 返回 NO_DEFAULT_TARGET，不能由回执 message ID 猜测凭证。
- Core 为每个 plugin_id 持久化后台发信许可 `allowed_adapter_instances`：缺省允许全部实例（D-83），管理员可以经既有插件权限配置把它收紧为指定实例的集合；协议 required_capabilities 不等于发信许可。持有仍在途的可信平台事件调用链时，插件及其合法 Service 子调用可向 Core 保留的原始完整 Target 回复，无需额外实例许可。独立后台、scheduled 或不同 Target 的 Send 必须显式 To，且目标实例在该插件许可集合中；再检查实例启用、account 绑定和 Core 全局禁止。许可覆盖该实例显式 chat/thread 目标，不赋予 User/Admin 或其他实例权限；无许可返回 PERMISSION_DENIED，动态撤销在下一次 Send 准入时生效。
- 上传、读取自己拥有的媒体不要求发信许可；Core 对真实事件签发的 IncomingMedia token 本身授予该 Session 的对象读取权。Read 只认有效 token/read 范围，Send 另行检查上述目标授权。Grant 不扩大对象、操作、Target 或有效期，只向当前已激活且满足原对象消费范围的 Session 签发受限 read token；入站对象及其派生对象的接收者还须与该事件有合法在途调用关联，或已获源实例许可；插件自己上传的对象可由 owner 显式委托当前 Activated 接收 Session 只读使用，Core 记录委托且保持原对象/有效期/Target 限制。上述关联由 Core 验证，接收者再次 Send 仍须通过自己的发信准入。
- MessageChain 为一次逻辑 Send 的有序片段。平台可能调用多次 API，必须保留所有确认 ID。接受、部分成功、明确拒绝、结果未知与未尝试部分按 MESSAGE-MODEL 处理；RPC 可同时返回 SendResult 和 ErrorInfo，错误不丢失已经发生的结果。
- Edit 修改一条已发送的平台消息，输入为 Target、platform_message_id 和新的 Chain，Target 准入与 Send 相同。v0.1 的 Edit Chain 只能包含一个 text 片段和至多一个 buttons 片段。每次 Edit 同样执行 BeforeSend → Adapter.Edit → AfterSend，Hook 输入标明本次操作为 edit。目标 Adapter 未声明 `message.edit` 能力时返回 UNSUPPORTED；平台拒绝（例如消息不可编辑或内容未变化）按 MESSAGE-MODEL §4 报告，不改为发送新消息。
- 自动回复 Handler `(Content,error)`（Content 是 Chain 或单个片段，D-108）成功且内容非空时，Core 在 Handler 返回后发送一次，排在该 Handler 内的显式 Send 之后；两种 Handler 都可以显式多次 Send（D-70）。StopPropagation 通过 Context/HandlerOutcome 表达。返回空 Chain 是无自动回复，显式 Send 空 Chain 是参数错误。
- 整条 Chain 先验证并生成有序物理操作计划，再开始平台调用；无法表达的组合在首个副作用前返回 UNSUPPORTED。Reference、按钮、相册和文本不能静默丢弃/重排。Telegram 按钮使用 Core token 关联真实点击 Actor、当前权限、路由代次与目标，协议确认不等于业务成功。
- 通用能力的范围（D-87）：通用的 Segment、Core 服务与 SDK 方法只收两个官方 Adapter 都支持且语义一致的能力；只有一个平台具备的能力由该 Adapter 以 extension 片段、扩展事件或平台动作提供，不新增通用接口。
- 撤回（D-93）：`core.delete` 的 Target 准入与 Edit 相同，Core 经 `adapter.delete` 交给目标实例；实例未声明 `message.delete` 时 Core 在调用 Adapter 前返回 UNSUPPORTED。撤回不带消息内容，不执行出站 Hook，但与 Send 一样不能在出站 Hook 中调用，也不自动重试。语义见 MESSAGE-MODEL §4A。
- 平台动作（D-92）：Adapter 在注册描述中声明动作，含名称（规则与扩展事件相同，完整名带 `<adapter_id>.` 前缀）、版本、参数与结果的 JSON Schema、作用范围（`chat` 需要 Target，`account` 只需要 adapter_instance_id）。插件经 `core.action` 调用：Core 先按本节的发信准入检查目标（持有在途可信事件调用链时，原 Target 及其实例无需额外许可；其余要求实例在该插件的 allowed_adapter_instances 中，并检查实例启用与 account 绑定），再核对动作已在目标实例所属 Adapter 中声明、版本一致、参数符合 Schema，然后经 `adapter.action` 交给实例；结果不符合结果 Schema 时返回 INTERNAL 并记诊断。Core 不提供未声明动作的透传，不按 Actor 检查动作（由插件为命令设置 Permission），不自动重试；结果未知时返回 UNAVAILABLE 或 DEADLINE_EXCEEDED。动作不能在出站 Hook 中调用。manifest 的 `adapter_dependencies` 可以列出所需动作及版本，启用检查与 extension 相同。Adapter 包为每个动作发布带类型的 Go 函数，只依赖 `sdk` 根包，开发者不直接书写动作名与 JSON；SDK 的通用动作方法只作底层入口。
- 命令菜单（D-96）：实例声明 `commands.sync` 能力时，Core 在该实例启动成功后与路由表或命令策略变化后（2 秒内的多次变化合并为一次）经 `adapter.commands.sync` 交给它当前可用的命令列表：只含该实例命令前缀含 `/`、权限只要求 `user` 且未被禁用的根命令，每项为命令名与按 locale 翻译的说明。命令名取生效的主名（§2.3，D-128），主名只有一个 token 的命令才算根命令。同步失败只记诊断，不影响实例运行，下一次变化时重新同步。
- 通用 Segment 在所有 Adapter 上语义一致，是否支持按 Adapter 公布的 Capability 判定。平台专有能力使用 `extension` 片段（adapter_id、name、version、JSON payload），其 Schema 由 Adapter 在注册时声明；Core 在发送前按目标实例所属 Adapter 的声明校验，目标不属于该 Adapter 或版本未声明时返回 UNSUPPORTED，不降级为其他片段。Adapter 包可以另行发布只含数据构造函数的 Go 模块供插件编译时使用，插件不链接 Adapter 的运行时代码。
- Adapter 在注册描述中声明本 Adapter 的权限与扩展事件（D-69、D-71）。权限声明含名称、隐含的权限（同一 Adapter 已声明的权限或 `chat_admin`，不能隐含 `user` 与 `admin`，不能成环）与供 WebUI 显示的文案键（Adapter 包 `i18n/` 中的 messages 键）。扩展事件声明含名称、版本与 payload 的 JSON Schema。名称只用小写字母、数字与下划线，完整名带 `<adapter_id>.` 前缀。Core 接受扩展事件前核对名称已声明、版本一致、payload 符合 Schema，不符时拒绝该事件并记诊断。Adapter 包按上一条的方式发布只含常量与数据类型的 Go 包，供插件引用这些权限与事件。Go SDK 中这些 Schema 不手写：Adapter 以 Go 类型声明扩展事件、extension 与平台动作，SDK 用 jsonschema-go 从类型推导 Schema（不允许未声明的属性，没有 omitempty 的字段为必需；动作结果允许未声明的属性），类型放在常量包中，插件使用同一类型（D-113）。
- 插件在 manifest 的 `adapter_dependencies` 中声明依赖的 `adapter_id`、所需 extension 及版本、是否必需。启用插件时，必需依赖的 Adapter 必须已安装且已声明所需 extension，否则拒绝启用并列出缺失项。Adapter 停用或卸载后，依赖它的插件继续运行，发往该 Adapter 的 Send 返回 UNAVAILABLE；卸载被必需依赖的 Adapter 时，须列出受影响的插件并要求显式确认。
- 插件在 manifest 的 `plugin_dependencies` 中声明依赖的插件：plugin_id、可选的最低版本 `min_version`（SemVer，含该版本）和是否必需。启用插件时，必需依赖必须已安装、已启用且版本满足，否则拒绝启用并列出缺失项；可选依赖缺失只记诊断。Core 不规定启动顺序，依赖暂时不可用时对它的 Service 调用返回 UNAVAILABLE。停用、卸载被必需依赖的插件，或把它更新、回退到低于 `min_version` 的版本时，须列出受影响的插件并要求显式确认，确认后依赖方继续运行。必需依赖不能成环，安装时发现环则拒绝。
- 实例经 `adapter.instance.start` 成功启动后，Core 生成 `bot.online` 事件；实例因停用、Adapter 进程退出或 `core.adapter.status` 报告的致命错误离开运行状态时，Core 生成 `bot.offline` 事件（payload 见 MESSAGE-MODEL §1.1）。这两类事件由 Core 生成而非 Adapter 交付，不带 Actor 与 Target，只进入声明了对应 event_type 的事件路由；Core 重启时不为重启前的状态变化补发。
- 平台能力、事件类型化 payload、Telegram 发送/接收边界及回执详见 MESSAGE-MODEL。Wire 字段号/错误/控制流详见 WIRE-FORMAT；这些是实施契约，尚未生成并编译 SDK。

## 9. 版本、SDK、打包 CLI 与构建

### 9.1 Manifest 与构建产物

所有 ID 最长 64 个 ASCII 字符。市场使用 `^alonabot_plugin_[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`，同时应用长度上限；本地使用 `^[a-z][a-z0-9_-]{0,63}$`，缺市场前缀只警告。为保证五平台路径有效，本地 ID 还拒绝 Windows 保留设备名 con/prn/aux/nul/com1～com9/lpt1～lpt9；大小写以规范小写保存，不能靠大小写形成两个安装。

- 组件工程中的 `plugin.yaml` 由作者编写，只写作者信息（D-78）：必填 `id`、`name`（显示名，英文源文案）、SemVer `version` 与 `source`；可选 `description`、`author`、`role`（缺省 `plugin`）、`icon`，以及下文按角色需要的字段；会写框架外数据时写 `prepare_effects: external`。`name` 与 `description` 可以在 `i18n/<locale>.json` 的顶层 `metadata` 中翻译（§5.4），WebUI、`list` 与 help 显示 name。
- `alonabot build` 与 `pack` 在包内的 manifest 中补上构建信息：`manifest_version=1`、`protocol_min_major/protocol_max_major` 与 `required_capabilities`（由所链接的 SDK 决定）、`os/arch`、相对可执行文件路径 `entrypoint`、`config_version`（`config_schema.json` 的 `$version`，没有该文件时为 1）、`migration_mode`（按实际实现的迁移接口）与 `prepare_effects`（作者没写 external 时，按是否链接受管 SQLite 取 `managed_db` 或 `none`）。作者在工程的 `plugin.yaml` 中写这些字段时 `check` 报错。包校验严格解析包内 manifest，同时核对 `config_schema.json`、`instance_schema.json`、注册描述、i18n 与 build-info；SDK 是编译依赖，不要求 Core 和插件 Go patch 相同，运行兼容以协议和能力协商为准。
- 可安装的包统称组件，使用同一份 manifest，`role` 取 `plugin`（缺省）、`adapter`、`agent` 或 `webui`（D-59）。只有 `plugin` 是插件，可以安装任意多个；`adapter` 每个 `adapter_id` 至多安装一个，`agent` 与 `webui` 全局各至多安装一个，同类组件的替换见 §9.3。市场 ID 依次使用 `alonabot_plugin_`、`alonabot_adapter_`、`alonabot_agent_`、`alonabot_webui_` 前缀，其余部分与插件市场 ID 的规则相同。
  - Adapter 另需 `adapter_id`（小写平台类型，规则同本地 ID），实例配置的 schema 写在包内的 `instance_schema.json`（§5.5、§8），不得声明插件路由、HTTP、数据库或迁移。
  - Agent 的声明与插件相同，可以有路由、HTTP、数据库与 schedule；它额外承担 §13 的职责，并是 `@agent` 的解析目标（§3）。
  - WebUI 只含静态文件，没有进程、配置、数据库与路由：作者写 `id`、`name`、`version`、`role`、`source`、静态根目录 `assets`（包内相对目录，含 `index.html`）与 Admin API 主版本范围 `admin_api_min_major`/`admin_api_max_major`（§7.5）；`pack` 只补上 `manifest_version`，不写 `entrypoint`、协议主版本、`required_capabilities` 与 `config_version`。WebUI 与平台无关，`os/arch` 写 `any`，归档名为 `<id>-<version>-any.tar.gz`。
- 任何角色都可以声明可选的 `icon`：包内图标文件的相对路径（§7.5）。插件可声明 `adapter_dependencies` 与 `plugin_dependencies`，语义见 §8。
- `migration_mode` 取 `none/self/versioned`，由构建按代码中登记的迁移方式写入（§6.2）；`prepare_effects` 固定 `none/managed_db/external`，描述配置回调以外的迁移/OnStart 在准备期的副作用。`managed_db` 仅表示写所属的受管 SQLite，候选可隔离；`external` 表示会写框架外的数据（例如插件自管的外部数据库），不具备自动回退承诺，按 §6.3 的维护门槛更新。包内 manifest 缺失这两项或与实际不一致时拒绝安装与自动升级，不能把未知值解释成安全。它们是插件包静态声明，不是 WebUI 调参开关。
- 命令、文本匹配与事件只在入口包的 `func New() *sdk.Plugin` 中对插件对象链式注册（D-79、D-114），生成器不扫描注册；`build` 检查并嵌入 manifest、`config_schema.json`、`i18n/`、Service Schema 与 Handler 的参数名（命令省略 `Args` 时使用，D-112），入口包必须声明 `func New() *sdk.Plugin`（缺少或签名不符时构建失败），配置 schema 只来自 `config_schema.json`（§5.1），重复 Handler/HTTP 路由失败。Schema/注册描述与 binary 来自同一构建，安装期检查能静态取得的 Schema/能力与自动注册项；实例方法显式注册的完整集合在本次启动 REGISTER 时校验，不要求静态推导任意 Go 控制流或重复手写整张路由表。最终路由在该 Session 固定，Schema/能力必须与包资源一致；Go 与 protoc 的版本已由 T-01 锁定，见 docs/development.md。
- 插件对象（D-114）：`sdk.NewPlugin()` 返回 `*sdk.Plugin`，路由的注册方法、`OnStart`/`OnReady`/`OnStop`/`OnConfigLoad` 的登记与数据库迁移的登记（§6.2）都作用在它上面；带类型参数的 `sdk.Serve`、`sdk.Subscribe` 与 `sdk.OnButton`（§3，D-122）的第一个参数是 `sdk.Registrar`，由 `*sdk.Plugin` 与 `*sdk.SubPlugin` 实现。注册在 `New()` 返回之前完成：同一回调登记两次由 `Collect` 报错；`sdk.Run` 开始之后调用注册或登记方法 panic，消息为英文。生成的正式入口调用 `sdk.Run(New())`；`New()` 可以被多次调用（例如测试），不得依赖包级的注册状态。
- 结构体参数（D-115）：命令 Handler 在 `*sdk.Context` 之后或者是零到多个位置参数，或者恰好是一个结构体值，不能混用。结构体的规则：
  - 字段按声明顺序对应参数位置；必须全部导出，类型为位置参数支持的类型或其指针，不能是结构体、内嵌字段或其他类型，否则 `Collect` 报错并指明字段。
  - 参数名取 `arg` 标签的名字；没有标签或名字为空时取字段名，并把开头的大写字母段转为小写，该段之后紧跟小写字母时保留段内最后一个字母（`URL` 为 `url`，`FeedURL` 为 `feedURL`，`URLPath` 为 `urlPath`）。同一结构体内参数名重复时报错。
  - 指针字段是可选参数，缺失时为 nil。`default:"…"` 给出缺失时的默认值，带 default 的字段是可选参数，不能是指针；默认值按 §2.2 该类型的语法检查。`arg:"name,raw"` 或 `arg:",raw"` 表示 raw 尾部，只能用于最后一个 string 字段。
  - §2.2 第 5 条的次序与互斥规则照常适用：必需在可选之前，raw 与 `message` 只能在末位且互斥，`user`、`image` 与 `message` 不能有默认值。
  - Handler 收结构体时，该路由调用 `Args`、`Optional` 或 `RawTail` 是注册错误；位置参数形式照旧使用它们。
  - 两种形式都接受 `int`（与 int64 同样解析）以及底层类型为 string、int64、int、float64、bool 的具名类型；SDK 不做其他类型转换。
  - 以上检查都在 `Collect` 完成。运行时 SDK 只把 Core 已解析的参数值赋给字段，不再分词或转换。
- 子插件（D-116）：`p.Sub(id)` 返回 `*sdk.SubPlugin`，注册方法与 `*sdk.Plugin` 相同，没有 `Sub`，也没有生命周期、`OnConfigLoad` 与迁移的登记。ID 为非空的 ASCII 字母、数字、`_` 与 `-`，同一插件内重复或不合法时 `Collect` 报错。子插件不出现在 REGISTER 中：其中路由的 route_id、命令路径、别名、Service 名、HTTP 路径与 job ID 按原规则生成，不带子插件 ID，并与整个插件的其他路由一起检查重复。`s.Permission(perm)` 设置其中命令与文本匹配路由的缺省权限，解析次序为路由、由近到远的命令组、子插件、SDK 缺省（§2.3）；子插件内创建的命令组属于该子插件。配置见 §5.1，日志字段见 §14.1。
- `pack` 输出 `<id>-<version>-<os>-<arch>.tar.gz`，Windows 同格式。归档内包括补上构建信息的 manifest、binary、存在时的 `config_schema.json` 与 `instance_schema.json`、`i18n/` 和 build-info，不包含运行配置/DB/secret/签名私钥。只接受普通文件与目录，拒绝绝对路径、`..`、symlink/hardlink、设备文件、重复归档路径及跨平台大小写冲突；解压到新临时目录、验证后再移动为不可变包目录，不能覆盖正在使用的程序。

### 9.2 来源、摘要与签名

- GitHub Releases 作为首发分发入口，不建设中心商店。**归档外**的 `release.json` 包含发布者公钥标识、绑定仓库来源、插件 ID/版本、各 OS/arch 资产 URL、字节长度和整个归档的 SHA256；归档内 manifest 不保存自身归档摘要，避免自包含循环。若记录单文件摘要，其对象必须明确是 binary 或资源，不与归档摘要混用。
- 发布者可以用 Ed25519 对 `release.json` 的**确切 UTF-8 文件字节**签名，输出外部 `release.json.sig`（D-81，可选）；验证先签名后解析，不能重新序列化 JSON 后验签。有签名时 Core 使用已信任公钥校验签名；有没有签名都验证下载长度/SHA256、平台和内层 manifest 一致性。实现使用 Go 标准库 `crypto/ed25519`，不引入通用 PKI 系统。
- 首次来源信任由管理员明确接受并持久化仓库，有签名时同时持久化发布者公钥指纹。有签名的来源：之后的更新必须由同一公钥签名，来源或公钥变化停止更新，重新信任是独立管理动作。没有签名的来源：安装界面与 `list` 标明未签名，之后的更新只从同一仓库取，照常校验长度、SHA256、平台与内层 manifest；它以后提供签名时，由管理员确认后改为按该公钥绑定。已签名来源永远不能静默降级为无签名，验签失败也不能自动降级为无签名信任；不能仅凭同 ID 接管已有安装。官方组件始终签名。
- 下载/签名/平台/静态检查都在停止旧版前完成；损坏或不可信候选直接拒绝。签名证明发布来源，不证明 Native 程序行为安全。新版本包只是安装候选，仍须 §6/§12 的准备与唯一提交流程。

### 9.3 CLI 的最小完整动作

Core 由 `alonabot run` 启动（§15.1），开发中的组件也由它运行（§15.4）。首次管理凭证与离线恢复使用 §7.4 的 `alonabot admin-token init/show/reset`。开发命令是对四种组件通用的顶层命令（D-77）：

| 命令 | 明确语义 |
| --- | --- |
| `init [目录]` | 在当前目录（或给定的空目录）创建组件工程：只含作者信息的 `plugin.yaml`、导出 `func New() *sdk.Plugin` 并注册一个示例命令的插件包、示例 `config_schema.json`、`i18n/en-US.json` 与忽略 `data/` 的 `.gitignore`；`--role` 选择角色，缺省 plugin；目录非空时拒绝。 |
| `build` | 生成嵌入代码（manifest、配置与 i18n 资源、Service Schema 与调用 `sdk.Run` 的正式入口）并编译，使用已锁定的 Go Modules；`--goos`、`--goarch` 选择目标平台。源码安装显式选择本地构建，预编译安装不要求服务器编译。 |
| `check` | 不写文件，核对生成代码是否最新、工程 manifest 是否只含作者信息、`config_schema.json`、`instance_schema.json` 及其迁移与 i18n。 |
| `test` | 在临时数据和离线测试会话检查加载与注册，不能调用生产凭证或宣称已测试平台发信。 |
| `pack` | 先 `build`，再生成安装包与归档外的 `release.json`；给出 `--key` 时签名，也可以只生成待离线签名的清单，私钥不进入包。 |


管理命令如下；修改均经 Core 管理接口，CLI 不直接写状态库，离线时只允许只读检查或明确的维护恢复流程：

| 命令 | 明确语义 |
| --- | --- |
| `install` / `list` | 校验包、首次来源与 ID 后登记；仍处于已安装状态的同 ID 不覆盖。`--source owner/repo` 从该 GitHub 仓库的 Release 下载并安装；不给时安装命令行给出的本地包，来源取自其 release.json（发行包 `packages/` 中的官方组件即官方来源）。之后的 `update` 只从安装时的来源取版本。同类组件（§9.1）已安装时返回 ALREADY_EXISTS 并列出它；加 `--replace` 时列出将被替换的组件并要求确认，在新组件的下载、签名与静态检查通过后，按 `uninstall` 的缺省语义卸载旧组件（保留其数据与来源绑定），再安装新组件；旧组件替换前已启用时，新组件安装后按普通 `enable` 流程启用，否则保持未启用。会话、记忆、按组件 ID 记录的策略，以及被替换 Adapter 的实例配置与凭证都不在组件之间迁移，新 Adapter 的实例需要重新创建；新组件准备或启用失败时旧组件保持已卸载，可以按原来源重装。已卸载的保留记录只允许匹配原来源的显式重装并复用数据，来源不符拒绝。list 返回实际安装/激活状态与角色。 |
| `enable` / `disable` | 已有 Active 时启动该代次；首次安装尚无 Active 时走首个候选 Prepared/提交。disable 停止并保留安装及数据。 |
| `restart` | 以当前 Active 程序/配置/DB 启动新 Session，不导入编辑中的 config.json。 |
| `reload` | 把编辑快照导入 Pending，走配置校验及候选激活；SQLite 同样复制 DB。 |
| `update` | 仅从绑定来源选择符合 pin 的候选版本并执行受控更新；不把 install 同 ID 当更新。 |
| `uninstall` | 停止、解除运行注册并删除未引用程序包，清理范围有三种（D-127）。不带选项时只卸载：保留数据、配置与来源绑定，以便同来源重装。`--purge-data` 另外删除数据，保留配置与来源绑定。`--purge` 删除数据与配置，连同来源绑定与安装记录。数据是框架受管的运行数据：全部代次（受管 SQLite、代次目录与代次内的配置快照）、Agent 会话（§13.3）、按钮 token、动态任务以及静态任务的起点与暂停状态（§4.3）、Adapter 的游标。配置是管理员与作者给出的设置：编辑快照 `config.json`、Pending、secret、命令策略、工具策略、发信许可、Adapter 在 `core_config.json` 中的实例。`--purge-data` 之后重装，首个候选的配置取编辑快照；编辑快照与被删除的 Active 配置不同时，确认对象中写明。两种清理都须展示并确认准确对象；不擅自删除插件自管的外部数据库或任意插件自管目录。卸载先按停止流程摘除该组件的定时器、清空排队的到期任务并取消在途的任务调用；清理时在同一个状态库事务中删除它的任务与调度状态，不经过插件进程。只卸载时任务保留但不触发。 |
| `pin` | 固定允许的更新版本；解除固定须显式操作，不能被自动更新忽略。 |
| `admin add` / `admin remove` / `admin list` | 管理 Core Admin 列表（按用户字符串 `<adapter_id>:<external_user_id>`），与 WebUI 的 Core 设置页使用同一管理服务，写回 `core_config.json` 的 `admins`。 |
| `logs` | 按来源、级别与 trace_id 读取最近记录，`--follow` 时持续输出（§14.4）。 |
| `log-level` | 修改全局或某个插件、Adapter 的日志级别，立即生效（§14.3）。 |
| `rollback` | 明确选择保留代次为输入，停机建立新的恢复候选 Generation；SQLite 从所选快照复制独立 DB，复用对应程序/配置/secret，走同一 Prepared/提交流程，不复用历史 Generation ID 或直接写保留快照；在执行前列明将舍弃该快照之后的本地写入，要求显式确认。声明 `prepare_effects=external` 的插件返回 NEEDS_MANUAL_MIGRATION，外部已发送消息不会撤回。 |

管理动作按插件串行；安装顺序、来源、pin、卸载保留与恢复目标均持久化，Core 重启不能靠无序目录遍历重新决定。安装器初始固定上限为压缩包 256 MiB、展开总量 1 GiB、10,000 个条目，超限在写入/激活前报 RESOURCE_EXHAUSTED；T-01 验证计数覆盖流式解压，不新增用户调参页面。发布验收至少覆盖当前五个目标平台，CGO 插件由发布者提供对应产物，不自动假称能跨编译。

签名实现依据：[Go crypto/ed25519](https://pkg.go.dev/crypto/ed25519)。

## 10. 实施验收与停止条件

沿用 T-01～T-16，不新建平行实施计划；全部 V 编号和通过条件集中维护在 [TEST-MATRIX.md](TEST-MATRIX.md)。每个任务同时完成对应实现与最小验证，不要求先跑尚不存在的应用测试。

详细步骤、拟目录、交付物和失败停止条件见 [PLAN.md](PLAN.md)。SDK/Proto 编译、数据库恢复、跨平台进程清理、Telegram 集成及真实性能在相应任务中取得证据；跨阶段用例的分工与最终闭合点按 TEST-MATRIX 执行。

Core 可执行文件与 SDK 根包只链接 [docs/development.md](../docs/development.md) 列出的允许模块，由审查用 `go list -deps` 核对（V-75）；Core 与示例插件空闲时的常驻内存不超过 V-74 的预算。超出预算视为回归：先定位原因，确需提高时在 adr/DECISIONS.md 记录新值与原因。

可开始实施的条件是关键行为已有确定契约与验收；发布条件是相关阻塞测试实际通过。测试失败必须修实现或回写有依据的规格修订，不用静默异常、伪默认值或新增兼容层掩盖问题。

## 11. 契约交付与集成位置

各任务的具体动作由 [PLAN.md](PLAN.md) 维护，本节将契约实现定位到对应工作包。历史决策及 P/Q 处置集中保存于 [adr/DECISIONS.md](adr/DECISIONS.md)。

| 契约范围 | 实现与集成位置 |
| --- | --- |
| §1、WIRE 字段/控制帧/会话 | T-01 生成类型，T-02 实现双 IPC；媒体交 T-03 |
| §2、MESSAGE 分发/权限/Hook | T-03 发送，T-04 命令/文本匹配/策略，T-05 事件链、延续调用与多轮等待；间接 Hook 约束在 T-06 闭合 |
| §3 服务/广播 | T-06 接入真实父调用链、编码校验与有界广播 |
| §4 生命周期/调度 | T-02 生命周期原语，T-07 调度/HTTP，T-08 准备回调，T-09 正式更新集成 |
| §5～§6 配置/Secret/数据库/文案 | T-01 先验驱动路径，T-08 候选准备与文案目录，T-08.11 `config_schema.json` 与插件配置迁移，T-08.12 Adapter 实例配置与 Core 设置的迁移，T-09 提交与恢复；Telegram locale 映射在 T-03 |
| §7 信任/HTTP/认证 | T-02 进程，T-07 代理与授权检查，T-07.7 插件页面的沙箱隔离，T-07.6 首次 token 文件，T-09 首次登录及管理入口；T-09.12 WebUI 组件、组件图标、Admin API 参考与插件页面桥 |
| §8 Adapter/消息 | T-03 Adapter 进程与 SDK、官方 Telegram Adapter 包、Target、媒体、按钮、extension 片段、回执、消息编辑、bot 上下线事件与依赖解析；T-15 官方 OneBot 11 Adapter 包与跨 Adapter 一致性验收；T-09 Adapter 包与实例的正式管理及启用时的依赖检查 |
| §9 构建/分发/管理 CLI | T-01 契约，T-03.8 manifest 角色解析，T-04 生成器，T-04.9 删除构建期标记，T-09.16 作者 manifest 与构建信息，T-09 完整工具、来源检查（含无签名来源）与同类组件替换 |
| §12 状态/锁/恢复 | T-02 建立共同原语，T-08 准备资源，T-09 复用原语完成正式 Journal 和恢复 |
| §13 Agent 基础 | T-11 工具目录、策略、确认与会话存储；T-06 `@agent` 目标解析；T-12 官方 Agent 组件（含记忆与 `llm.generate` 服务）；T-09 工具策略与会话清除的管理页 |
| §14 日志 | T-02.1 日志基线与日志文件；T-07.6 `core_config.json` 中的级别设置；T-09.10 按来源级别的写回、最近记录、按日期查询、管理 API 与 CLI，T-09.15 日志页；各切片按 §14 记录 |
| §15 运行方式 | T-07.6 `core_config.json` 与首次 token；T-09.4 `alonabot run` 与管理服务；T-13 开发运行、热部署与库模式；T-14 发行包、容器镜像、install.sh 与预装插件 |
| §17 插件测试套件 | T-13.1 根包 `Start`；T-16 `plugintest` 与示例插件的测试；T-16.3 LLM 替身 |
| §16 调优参数 | T-02.2 参数表、`--args` 与 Bootstrap 下发；之后的切片把新增的可调值登记到同一参数表 |
| 全部发布门槛 | 各任务积累证据，T-10 汇总 V-01～V-76、V-82～V-115 和五目标运行结果 |

## 12. 跨层状态机与恢复不变量

### 12.1 状态与可见性

| Core 中的状态 | 允许的动作 | 业务可见性 |
| --- | --- | --- |
| Installed / Stopped | 下载、静态验证、管理元数据、等待启用 | 无插件路由 |
| Starting / Handshaking | 新进程身份认证、Bootstrap | 无插件路由 |
| Preparing | 配置准备、受管 DB 迁移、OnStart、完整注册验证 | 禁止业务 REQUEST 和公开 HTTP |
| Prepared | 新候选资源/配置已验证持久化，或原提交代次已重验 | 路由不公开 |
| Activating | 新候选提交 ActiveGeneration / COMMITTED，或采用原提交记录；Activate / OnReady | 仍不派发公开业务 |
| Activated | 合法命令、事件、服务、调度、HTTP | 仅唯一 ActiveGeneration 接单 |
| Draining | 最多 6s 完成已接纳调用及其合法子调用 | 拒绝新根调用和新调度 |
| Stopping | Cancel 余留调用、STOPPING / OnStop、资源关闭 | 无新业务，至总 10s 截止退出 |
| Degraded / Crashed | 记录故障、隔离路由、按既定恢复策略处理 | 不擅自用旧数据继续服务 |

同一插件 ID 在任一时刻最多一个可接单 Generation。REGISTER 验证成功、Prepared、持久化提交、Activated 是不同节点，不能用一个含糊 Ready 替代；具体控制帧及失败回包见 WIRE-FORMAT，SDK 的本地 ready gate 见 §4。

### 12.2 进程和变更所有权

- Core 在启动 Adapter/插件或修改状态库之前取得 `<data_dir>/core.lock` 的 OS 独占锁，持有至退出。同一 data_dir 第二个 Core 明确拒绝启动，不依赖 SQLite 写事务锁充当运行实例锁；v0.1 的 data_dir 要求本机文件系统，不支持多主机共享数据目录运行。
- 每插件的 install/enable/disable/reload/update/rollback/uninstall 操作串行化。一次候选操作记录预期旧 ActiveGeneration；提交事务再次校验该值，已变化则失败，不能让陈旧候选覆盖新状态。全局锁不串行化不同插件的普通 Handler。
- 每插件使用跨代次稳定的 `<plugins>/<id>/runtime.lock`。SDK 在打开受管 DB 或运行任何初始化回调之前取得 OS 独占锁并持有至退出；锁句柄不传给无关子进程。Linux/macOS 使用 flock，Windows 使用 LockFileEx。锁文件不在解锁后删除或替换，防止并发进程锁住不同文件。
- 停机复制受管 SQLite 时，Core 在旧进程树退出后暂持同一 runtime.lock，完成副本操作再释放，由随后唯一启动的 SDK 取得；Core 不能持锁等待候选 SDK 启动造成自锁。Core 崩溃恢复时，旧 SDK 按断链规则退出，新实例必须在取锁成功后才打开 DB。拿不到锁时保持明确失败/Degraded，不能通过固定 sleep、删除锁文件或猜测 PID 证明旧写入者已停止。
- 锁和 Session 门禁约束遵循 SDK 的普通插件，不构成 Native 恶意代码沙箱；系统级信任边界见 §7。确认旧进程退出、资源关闭和唯一取锁是恢复前提，Journal 指针本身不停止仍活着的原生代码。

### 12.3 提交、恢复与回退边界

- 包、候选 Config、受管 DB 副本必须先持久化，Core 状态库再在同一事务提交 ActiveGeneration 与 COMMITTED。该事务仅控制 Core 元数据和已准备的引用，不是跨文件系统、外部数据库或外部 HTTP 的分布式事务。完整流程、迁移账本与备份要求以 §6 为唯一规范。
- 未提交候选失败：关闭候选并确认退出，使用旧包、旧配置和旧受管 DB 启动旧版本的**新 Session**。不是恢复已失效的旧 Session，也不重放在途业务。
- 事务已 COMMITTED：恢复时只启动新代次；OnStart/OnReady/资源打开失败则 Degraded并禁止接单，保留可恢复材料。不能直接把旧快照当最新业务数据，也不能承诺已发送消息、SQL或原生文件变更可自动撤销。
- 配置更新与版本升级在承诺受管 SQLite 候选回退时都使用独立副本；“没有 DDL”不能据此复用旧代次可写 DB。无业务 DB 的插件无需复制。插件自管的外部数据库与文件、框架外副作用按 §6 的人工维护边界处理，不纳入无损自动恢复承诺。

### 12.4 实证门槛

协议编译、五 OS/arch 的进程锁/树回收、Core 被 kill 后的 SDK 断链退出、配置 OnStart 写 DB 后失败、SQLite 提交各节点崩溃恢复、Bootstrap/REGISTER 超限、满准入下 Response/Cancel 可完成、HTTP 端口与两个 Origin、真实平台媒体和资源基准，都需要实施时的测试证据；本次规格修订不把它们记为已运行。对应场景合入现有 TEST-MATRIX，不另建平行验收体系。

## 13. Agent 基础

Core 只提供 Agent 需要的基础能力，不链接任何 Provider 客户端或模型 SDK；对话循环、Provider、MCP、技能、记忆与定时任务由 Agent 组件（`role: agent`，§9.1）实现（D-27、D-60），其他插件经 `@agent` 的服务使用模型（§13.5）。Agent 组件全局至多一个，官方 Agent 组件可以被社区 Agent 替换（D-59）；除同类唯一与 `@agent` 外，它与插件使用同一 SDK、生命周期、权限与本节的 Core 服务。

### 13.1 工具与调用策略

- 工具是带 `tool` 描述的 JSON Service：description 是给模型的说明，输入 Schema 即该 Service 的输入 Schema，输出是该 Service 的 JSON 输出；default_policy 取 allow 或 confirm。工具沿用 Service 的命名 `plugin_id/service_name`、访问规则、Broker 定位、调用链、深度、取消与准入（§3），不新增 route kind；Protobuf 编码的 Service 不能声明为工具。
- 调用策略按工具持久化，取值 allow、confirm 或 deny，未设置时使用工具声明的 default_policy。管理员经管理服务修改，下一次调用生效，不撤销已开始的调用。策略不替代 Service 自身的 required_permission：工具执行时仍按继承的 Actor 判定 User/Admin。
- `core.tool.list` 返回调用插件可以调用、且生效策略不为 deny 的工具：名称、description、输入 Schema 与生效策略。
- `core.tool.invoke` 输入工具名与 JSON 参数，必须带 parent_call_id（§1.2）。策略为 deny 时返回 PERMISSION_DENIED；为 confirm 时先按 §13.2 取得确认，用户拒绝返回 PERMISSION_DENIED，超时返回 DEADLINE_EXCEEDED；为 allow 或确认通过后，Broker 以父调用链调用该 Service 并原样返回结果。确认的等待计入父调用期限。
- `core.tool.confirm` 供插件在自己执行的操作（例如 MCP 工具）之前取得确认，输入操作名（调用插件内稳定的名称）与给用户看的参数摘要。Core 按「调用插件 ID + 操作名」查找策略，未设置时为 confirm，返回规则与 invoke 相同。Core 不检查插件随后是否执行了该操作；它记录出现过的操作名，供管理页列出并设置策略。

### 13.2 用户确认

- 确认要求调用链具有可信 Actor 与默认 Target。没有 Actor 的调用（后台、scheduled）遇到 confirm 策略时直接返回 PERMISSION_DENIED。
- Core 直接经 Adapter 向默认 Target 发送一条确认消息，内容为工具或操作名、要执行的参数以及「允许」「拒绝」两个按钮，语言按 §5.4 选择。`core.tool.invoke` 的参数由 Core 按实际 JSON 参数展示，不使用插件提供的摘要；`core.tool.confirm` 展示插件提供的参数摘要。这条消息不是插件的 Send，不执行插件出站 Hook。目标 Adapter 不支持 buttons 时返回 UNSUPPORTED，不改为文字确认；确认消息发送失败时返回该错误，不执行工具。
- 按钮 token 沿用 MESSAGE-MODEL §6.3 的格式与存储，但绑定 Core 内部的确认记录而非插件路由。只接受同一 Actor 在 **5 分钟** 内的第一次点击；其他用户的点击被拒绝，不影响结果。token 只能使用一次，超时、父调用结束或 Session 失效后作废。确认只对本次调用有效，不记住「总是允许」。

### 13.3 会话存储

- Core 状态库按插件保存 Agent 会话。会话键由插件给出，为 1～256 字节的 UTF-8 字符串；条目是插件定义的 JSON 对象，按追加顺序编号，Core 只校验它是合法的 JSON 对象。
- `core.conversation.append` 追加一条或多条条目，同一请求的条目原子写入，返回最后一条的序号。`core.conversation.list` 从给定序号起按顺序返回条目；响应受帧上限约束，Core 在超限前停止并返回下一个序号。`core.conversation.clear` 删除该会话的全部条目。插件只能访问自己的会话，这些服务不需要父调用链。
- Core 不截断、不过期、不压缩会话。管理服务可以列出插件的会话并清除指定会话；卸载插件默认保留会话，`--purge-data` 或 `--purge` 时一并删除（§9.3，D-127）。

### 13.4 记忆

- 记忆由 Agent 组件自己实现和存储，Core 不提供记忆服务（D-60）。官方 Agent 组件的记忆见 DESIGN §R 与 PLAN T-12 步骤 5，社区 Agent 可以采用其他实现。
- Agent 把写入记忆的操作注册为工具（§13.1），模型的调用照常经 `core.tool.invoke`，管理员可以按工具设置策略。
- 记忆的范围取自 Core 下发的可信上下文：会话按默认 Target（adapter_instance_id、chat_id、thread_id），用户按 Actor，不采信模型或插件在参数中给出的范围。缺少所需的 Actor 或默认 Target 时返回 FAILED_PRECONDITION。

### 13.5 文本生成服务

- Agent 组件注册 JSON Service `llm.generate`，让其他插件复用管理员已配置的 Provider，不必各自内嵌 Provider 客户端和 API Key（D-51）。Schema、Go 类型与调用辅助由 `sdk/llm` 提供，调用辅助以 `@agent` 为目标（§3、D-61），因此替换 Agent 组件后调用方不用修改；只需一问一答时用 `llm.Ask(ctx, prompt)`，它等于只含一条 user 消息、使用默认 Provider 的请求，返回生成的文本（D-111）；替换官方 Agent 的组件应实现本服务。
- 输入：`messages`（至少一条，每条为 `role`（user 或 assistant）与 `text`）、可选的 `system`、可选的 `provider`（Agent 配置中的 Provider 名称，缺省为默认 Provider）与可选的 `max_output_tokens`。输出：`text`、实际使用的 `provider` 与 `model`，以及 Provider 返回用量时的 `input_tokens` 与 `output_tokens`。
- 服务声明为 public（§3）。Agent 只接受直接调用方插件位于其配置项 `llm_service_allowed_plugins` 中的调用，缺省为 `["*"]`，表示全部插件（D-83），管理员可以改为插件 ID 名单；不在列表中返回 PERMISSION_DENIED，不请求 Provider。直接调用方取 Core 下发的调用链中的上一跳插件，不采信请求内容。
- 一次调用对应一次完整的 Provider 请求：不流式、不调用工具、不读写 Agent 会话与记忆，也不套用人格。`messages` 为空、`role` 非法或指定的 Provider 不存在返回 INVALID_ARGUMENT；没有配置任何 Provider 返回 FAILED_PRECONDITION；Provider 出错或中途断开返回 UNAVAILABLE，safe_message 不含 API Key；不自动重试（D-38）。
- 生成可用的时间受调用方期限约束（§1.2）。普通请求只有 30s，需要更长时间的调用方在延续调用（§2.4）或 schedule 回调中发起。没有安装或没有启用 Agent 组件时按 §3 的目标未激活处理。

## 14. 日志

### 14.1 记录与来源

- Core、插件与 Adapter 都用 Go 标准库 `log/slog` 记录。一条记录包含时间、级别、消息、源码位置（文件与行号）、来源与键值字段。来源分为 core、plugin、adapter 三类：plugin 与 adapter 另带来源 ID（plugin_id 或 adapter_id），与某个 Adapter 实例相关时带 instance_id。处理某次调用时，记录带上该调用的 trace_id（Envelope.trace_id）、route_id 与 call_id。记录中的字段名：来源类别 `component`（core、plugin、adapter）、`plugin_id`、`adapter_id`、`instance_id`、`session_id`（十六进制）、`generation_id`，调用字段 `trace_id`、`route_id`、`call_id`。子插件中的 Handler 经 `c.Logger()` 写出的记录另带 `sub_plugin`（子插件 ID，D-116）。
- 级别为 DEBUG、INFO、WARN、ERROR。逐条事件与逐次调用的记录只用 DEBUG：

| 级别 | 用途 |
| --- | --- |
| ERROR | 需要人处理的故障：插件崩溃或被自动停用、状态库错误、提交失败、实例致命错误、Handler panic |
| WARN | 被拒绝或丢弃，但系统继续运行：队列满丢弃、Handler 返回错误、确认超时、接收退避重连、非结构化输出 |
| INFO | 状态变化：Core 启停、插件激活与停止、实例上下线、安装、更新与回退、配置重载、管理员登录 |
| DEBUG | 单条事件与单次调用的过程：事件接收、匹配结果、调用耗时、发送回执、工具调用、Provider 请求 |

- 插件与 Adapter 的 SDK 把记录以每行一个 JSON 对象写到 stderr，字段为 time、level、msg、source（file、line）与键值字段。Core 逐行解析，补上来源类别、来源 ID、session 与 generation 后输出；插件给出的同名字段不能覆盖 Core 补充的值。
- 不是合法记录的行（直接写到 stderr 或 TCP 模式下 stdout 的文本、未被捕获的 panic 输出）作为 WARN 原样保留，消息为 `unstructured output`，带 stream 字段，原文在 output 字段，不丢弃。SDK 在调用边界捕获的 panic 记为一条 ERROR，堆栈放在 `stack` 字段。
- 插件或 Adapter 进程退出后，Core 先读完它的日志管道再结束这次回收，最后的输出不因 Core 随后退出而丢失；停止时这段等待受 §4.2 的总截止约束，到期仍未读完（例如脱离进程树的进程持有管道）时记一条 WARN 后继续。

### 14.2 输出格式

- Core 把全部记录写到自己的 stderr。`--log-format text`（缺省）为下面的控制台格式；`json` 为每行一个 JSON 对象，包含全部字段与完整的 trace_id。记录同时按 §14.6 写入日志文件。
- 控制台格式仿照 AstrBot：

  ```text
  [14:05:42.118] [Plug:weather] [ERRO] [weather/weather.go:57]: fetch forecast failed: context deadline exceeded trace=7c2e91d0
  ```

  - 时间为本地时区的「时:分:秒.毫秒」。
  - 来源标签：Core 为 `[Core]`；插件为 `[Plug:<id>]`；Adapter 为 `[Adpt:<adapter_id>]`，与实例相关时为 `[Adpt:<adapter_id>/<instance_id>]`。插件 ID 去掉 `alonabot_plugin_` 市场前缀显示。
  - 级别缩写为 DBUG、INFO、WARN、ERRO。
  - 源码位置为「所在目录名/文件名:行号」。
  - 消息之后按记录顺序输出 `key=value`，值含空白或引号时按 Go 的带引号字符串输出；trace_id 显示为 `trace=` 加前 8 位十六进制。多行的值（例如 `stack`）排在单行字段之后，值的每一行另起一行并缩进 4 个空格。
  - 来源字段已体现在标签中：插件与 Adapter 的记录不再输出 component、来源 ID、instance_id、session_id 与 generation_id，这些只在 JSON 输出中出现；Core 关于某个插件的记录照常输出 plugin_id 等字段。
  - 插件提供的消息、键与源码位置含换行或其他不可打印字符时按带引号的字符串输出，不能伪造出另一条记录。
  - 只有 stderr 是终端且未设置 `NO_COLOR` 时才着色：时间为绿色，DBUG 蓝色、INFO 青色、WARN 黄色、ERRO 红色。

### 14.3 级别

- 全局级别由 `core_config.json` 的 `log.level`（§15.1）设置，缺省 info；`--log-level` 只覆盖本次运行。管理员在运行时修改时立即生效并写回该文件。Core 自己的记录按全局级别输出。
- 每个插件与每个 Adapter 包可以单独设置级别：debug、info、warn、error、off，或 inherit（缺省，跟随全局）。该设置写在 `core_config.json` 的 `log.sources`（按组件 ID），经管理服务修改后立即生效并写回该文件。
- Core 在 Bootstrap 的 `LogLevel log_level = 10` 中下发该来源的生效级别；枚举值为 DEBUG、INFO、WARN、ERROR、OFF，UNSPECIFIED 按 INFO 处理，inherit 由 Core 先解析，运行时的变化经 `sdk.log.set_level` 通知。SDK 在进程内按这个级别过滤，off 时不输出任何记录；Core 对收到的记录再按同一级别过滤。
- 来源设为 off 只隐藏该进程自己的记录；Core 关于它的记录（例如崩溃、停用）仍以 `[Core]` 按全局级别输出。开发中的组件缺省为 debug（§15.4）。

### 14.4 查看

- Core 为 Core 自己、每个插件与每个 Adapter 包各保留最近 **1000** 条已输出的记录。这些记录只在内存中，Core 重启后清空。
- 管理 API 提供按来源、级别与 trace_id 读取这些记录，以及经 SSE 实时订阅，都要求管理员会话或 CLI 的 Bearer token；`alonabot logs` 使用同一 API。
- 过去日期的记录从日志文件读取：管理 API 与 `alonabot logs --date` 按日期、来源、级别与 trace 查询，WebUI 日志页可以选择日期。
- WebUI 日志页先读取最近记录再订阅。按来源、级别与 trace 的筛选只在浏览器中进行，不改变输出，筛选条件保存在浏览器本地；修改输出级别使用 §14.3 的管理服务。

### 14.5 脱敏

- 日志不包含 Secret 明文、Bootstrap、完整配置 JSON、DSN、管理 token 与会话凭证。含凭证的 URL（例如 Telegram Bot API 路径中的 `bot<token>`）在记录前脱敏；HTTP 客户端的错误先去掉 URL 中的凭证再记录。
- DEBUG 记录可以包含消息文本，不包含媒体内容。

### 14.6 日志文件

- `log.file.enabled` 为 true（缺省）时，Core 把全部记录以 §14.2 的控制台格式（不着色）写入 `<data_dir>/logs/alonabot-<YYYY-MM-DD>.log`，日期取本地时区。跨过零点时换到新日期的文件；当天的文件达到 `log.file.max_size`（缺省 **100 MiB**）时，依次续写 `alonabot-<日期>.1.log`、`alonabot-<日期>.2.log`。
- 全部日志文件合计超过 `log.file.max_total`（缺省 **1 GiB**），或文件日期早于 `log.file.max_age_days`（缺省 **30** 天）时，从最旧的文件开始删除，两个条件先达到的生效；正在写入的文件不删除。Core 启动时与每次换文件时检查。`max_size` 大于 `max_total` 时拒绝启动。
- 写入失败（例如磁盘已满）时 Core 继续运行，在 stderr 报告一次，到下一次换文件时重新打开。
- 这些是普通设置（§15.1），不属于调优参数。库模式的日志文件在 `.alonabot/logs/`。

## 15. 运行方式

### 15.1 启动与 core_config.json

- `alonabot run` 启动 Core，不带子命令运行 `alonabot` 与它相同（D-77）：取得 core.lock，打开状态库，读取 `core_config.json`，按安装记录启动全部已启用的组件（WebUI 没有进程，启用即提供文件）。`--data-dir` 缺省为工作目录下的 `data/`。给出组件目录，或当前目录是组件工程时，另外运行开发中的组件（§15.4）。
- Core 的设置都在 `<data_dir>/core_config.json`（D-82），仿照 AstrBot 的 `cmd_config.json`。文件不存在时以缺省值创建，权限只有运行账号可读写。它是这些设置的唯一来源，状态库不保存它们：

| 键 | 缺省值 | 含义 |
| --- | --- | --- |
| `$version` | 当前版本 | 文件的格式版本，迁移见 §5.5 |
| `listen` | `127.0.0.1:7650` | Core 唯一的公开 Listener（§7.2） |
| `base_url` | `http://127.0.0.1:7650` | Core 的地址，管理界面、Admin API 与插件路由都在这个主机名下（§7.3） |
| `trusted_proxies` | 空 | 可信反向代理的地址范围 |
| `log.level`、`log.format` | `info`、`text` | §14.2、§14.3 |
| `log.file.enabled`、`log.file.max_size`、`log.file.max_total`、`log.file.max_age_days` | `true`、`100MiB`、`1GiB`、`30` | §14.6 |
| `log.sources` | 空 | 按插件或 Adapter 包设置的日志级别（§14.3） |
| `proxy` | 空 | 出站代理 URL（http、https 或 socks5）；Telegram 实例只在 `proxy_mode` 为 core 时使用这里的代理环境（§8，D-101）；非空时 Core 为自身与所有子进程设置 `HTTP_PROXY`、`HTTPS_PROXY`。为空时，Core 把自身环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY` 原样传给子进程 |
| `admins` | 空 | Core Admin 列表，元素为用户字符串 `"<adapter_id>:<external_user_id>"`，例如 `"telegram:123456789"`、`"onebot11:10001"`（§2.3，D-105） |
| `command_prefixes` | `["/"]` | 命令前缀（§2.2）；`command_prefixes_by_instance` 按 Adapter 实例覆盖；列表不能为空，不带前缀的命令写 `[""]` |
| `default_locale` | `en-US` | 机器人回复的缺省语言（§5.4） |
| `adapters` | 空 | 按 adapter_id 分组的 Adapter 实例：`{"telegram": {"$version": 1, "instances": {"main": {"token": "${TELEGRAM_BOT_TOKEN}"}}}}`；实例配置按该 Adapter 的 `instance_schema.json` 补缺失值并校验（§5.5、§8）。实例名为 `^[A-Za-z0-9_-]{1,64}$`，跨 Adapter 也不能重名；每个 Adapter 可以写 `default_instance`（该 Adapter 的一个实例名），作为插件没有指定实例时的缺省发送实例（§8，D-102） |

- 任何字符串值都可以写成 `${环境变量}`，启动时展开；变量不存在时 Core 拒绝启动并指出键的路径。敏感值也可以直接写明文，文件权限保护它们。手写时可以省略 `$version`，Core 按当前版本读取，在下一次写回该文件时补上。
- WebUI 与 CLI 修改这些设置时写回该文件（先写临时文件再替换），保留原有的 `${...}` 写法，不把展开后的值写回。手动修改后重启 Core 生效；`alonabot run --reload` 时 Core 监视该文件：`admins`、`command_prefixes`、`command_prefixes_by_instance`、`default_locale`、`log.level`、`log.sources` 与 `adapters` 立即生效（Adapter 只重启变化的实例；`default_locale` 对运行中的插件从它的下一个 Session 起生效），其他键输出 WARN 提示需要重启。启动时文件不合法则拒绝启动并指出位置；运行中重新读取失败时保留当前设置并记 ERROR。
- 调优参数用 `--args` 覆盖（§16）。命令行参数覆盖 `core_config.json` 中的同名项，只对本次运行有效。Core 全局禁止表、插件的命令策略与发信许可等按插件或按用户记录的管理数据仍在状态库，经管理服务修改。缺省只监听本机地址。远程管理有两种方式：经 SSH 隧道访问本机地址，或改为非本机地址并满足 §7.3、§7.4 的 Host、HTTPS 与可信反代要求；容器镜像的例外见 §15.3。
- 首次启动且尚无管理凭证时按 §7.4 生成管理 token：stdin 与 stderr 都是终端时只在该终端显示一次；不是终端时写入 `<data_dir>/admin_token`，日志只输出该文件的路径。

### 15.2 发行包、官方整合包与预装组件

- 官方整合包是同一套发行物的三种安装方式（D-64）：发行归档、容器镜像与 `deploy/compose.yaml`（§15.3），以及 `install.sh`（§15.6）。三者包含相同的 Core 与官方组件。
- 每个目标平台发布一个发行归档 `alonabot-<version>-<os>-<arch>`（Windows 为 zip，其他为 tar.gz），包含 `alonabot` 可执行文件和 `packages/` 目录。`packages/` 中是官方 Telegram Adapter、官方 OneBot 11 Adapter、官方 Agent、官方 WebUI 与预装插件的组件包，以及各自归档外的 `release.json` 与 `.sig`（§9.2）。同一 Release 另附 `install.sh` 与列出全部发行归档 SHA256 的 `SHA256SUMS`。
- 发行版 Core 内置官方发布者公钥作为已信任来源：运行官方发行包即视为接受该来源，预装包不再逐个确认。签名与摘要照常校验，校验失败的包不安装并记 ERROR。
- 数据目录首次初始化时，Core 按普通安装流程安装 `packages/` 中的全部包，只启用 help 插件与官方 WebUI；Agent 组件与 help 以外的预装插件只安装不启用，不使用 AI 对话的部署不运行它们的进程（D-50、D-62）。官方 Telegram 与 OneBot 11 Adapter 只安装，`core_config.json` 的 `adapters` 中有它的实例时自动启用；某个 Adapter 写了实例却尚未安装时，Core 从 `packages/` 安装，没有发行包（例如经 `go install` 取得的 `alonabot`）时从官方 Release 安装（D-82）。需要 AI 对话时，管理员在 WebUI 或 CLI 启用 Agent；Agent 在没有配置 Provider 时也能启用，`/ask` 回复提示管理员配置。
- 管理员可以用 `install --source <来源> --replace` 把任一官方 Adapter、Agent 或 WebUI 换成社区组件（§9.3）；之后不再从官方来源更新该组件。卸载或损坏 WebUI 后，用 `alonabot install` 从 `packages/` 重装官方 WebUI。
- 之后不再自动安装，已卸载的预装包不会在下次启动时重新安装。新版本发行包中的包版本高于已安装版本时，作为可用更新列在管理界面与 `alonabot list` 中，由管理员确认后按 `update` 流程更新。Core 本身通过替换可执行文件升级，新版本按 §5.5 迁移 `core_config.json`。
- 首批预装插件为 help（按调用者权限列出可用命令与用法，`/whoami` 回复调用者的用户字符串与收到消息的实例名）、welcome（新成员入群时发送欢迎语）与 rss（订阅 RSS/Atom 并推送到聊天）。
- 第一个 Core Admin 的产生：用户在聊天中发送 `/whoami` 取得自己的用户字符串，管理员把它写入 `core_config.json` 的 `admins`，或在 WebUI 的 Core 设置页、`alonabot admin add` 中添加。预装插件与第三方插件一样以独立进程运行，只有启用的才运行。

### 15.3 容器镜像

- 每个版本发布官方容器镜像 `ghcr.io/alonabot/alonabot:<version>`（同时更新 `latest`），平台为 linux/amd64 与 linux/arm64。镜像内容与该平台的发行归档相同：`alonabot` 可执行文件与 `packages/`，另含 CA 证书；不含 shell、Node 或 Python，需要这些运行时的 MCP stdio 服务器须使用自行构建的镜像。
- 镜像以非 root 用户（UID 10001）运行 `alonabot run --data-dir /data`，数据目录 `/data` 是卷。容器须以 Docker 的 init 运行（`docker run --init` 或 compose 的 `init: true`），由它回收孤儿进程；Core 不承担 PID 1 的回收职责。
- 镜像设置环境变量 `ALONABOT_CONTAINER=1`，`core_config.json` 的缺省 `listen` 改为 `0.0.0.0:7650`，`base_url` 的缺省值不变（`http://127.0.0.1:7650`）。只有在这个环境变量下，Core 才允许非本机的 listen 搭配本机 http 的 `base_url`；容器的端口只应发布到宿主的本机地址（`127.0.0.1:7650:7650`），Core 启动时为此输出一条 WARN 提示。远程访问按 §15.1 使用 SSH 隧道或 HTTPS 反向代理。Adapter 的凭证可以经 `${环境变量}` 从容器环境传入。
- 容器首次启动时没有交互终端，Core 把一次性管理 token 写入 `/data/admin_token`（§7.4）。镜像不含 shell，用 `docker compose exec alonabot alonabot admin-token show` 读出它，登录成功后该文件被删除。
- 仓库根目录的 `Dockerfile` 用多阶段构建产生镜像。`deploy/compose.yaml` 是官方整合包的 compose 入口，随每个 Release 发布：使用对应版本的官方镜像，包含卷、`init: true` 与只发布到宿主本机的端口；`docker compose up -d` 即可启动，不需要先单独初始化。

### 15.4 开发运行与热部署

- `alonabot run [目录...] [--reload]` 同时运行开发中的组件（D-77）。每个目录是一个组件工程（含 `plugin.yaml`）；没有给出目录且当前目录有 `plugin.yaml` 时为当前目录。数据目录缺省仍为工作目录下的 `data/`，`alonabot init` 生成的 `.gitignore` 忽略它。
- 每个开发组件先按 `alonabot build` 构建到 `<data_dir>/dev/bin/`，再按 manifest 的 `role` 运行：插件、Adapter 与 Agent 以独立进程运行，进程、隔离与有限重启和已安装的组件相同；WebUI 以该目录的 `assets` 为静态根目录，代替已安装的 WebUI，按请求读取文件，不经安装与签名（Admin API 主版本检查照常进行）。构建失败或同一 ID、adapter_id 与已安装的包冲突时，`run` 在启动任何组件前报错退出。
- 开发组件的规则（`alonabot run` 与库模式 §15.5 相同）：
  - Core 为它们创建 dev 来源的安装记录，不做来源、签名与归档摘要校验，Hello 的 build_digest 使用 Core 为本次运行生成的值；dev 来源的组件不能 update、rollback 或 uninstall。
  - 它们使用固定代次 `dev`，不走候选与 Journal：配置取自 `<data_dir>/plugins/<id>/config.json`（缺失时为空对象），每次启动经迁移、补值与校验后直接生效（§5.2、§5.5）。WebUI 或 CLI 修改开发组件的配置时，校验通过后写回该文件并重启该组件。受管 SQLite 为 `<data_dir>/plugins/<id>/dev.sqlite`，不复制、不回退。按钮 token 因代次固定，重启后仍然有效。
  - 开发中的 Adapter 的实例与已安装的 Adapter 一样写在 `core_config.json` 的 `adapters` 中。
  - 日志级别缺省为 debug（§14.3）。
- `--reload` 时每 `DEV_RELOAD_POLL`（§16）用标准库轮询每个组件目录中的 `.go` 文件、`//go:embed` 引用的文件、`go.mod`、`go.sum`、`plugin.yaml`、`config_schema.json`、`instance_schema.json` 与 `i18n/`（排除数据目录、vendor、以 `.` 或 `_` 开头的目录与 testdata），以及该组件的配置（插件的 `config.json`，Adapter 在 `core_config.json` 中的实例）：
  - 源码或资源变化时重新构建；成功后依次停止并重启变化的组件，Core、其他组件与按钮 token 不受影响；失败时输出编译器信息，旧组件继续运行。
  - 只有配置变化时不重新构建，只重启该组件；Adapter 只重启配置变化的实例。
  - `core_config.json` 的其他变化按 §15.1 处理。
- 开发组件运行在独立进程中，断点调试时把调试器附加到该进程，或改用库模式（§15.5）在同一进程中运行。

### 15.5 库模式

- 根包 `github.com/AlonaBot/AlonaBot`（包名 `alonabot`）用于把 Core 嵌入其他 Go 程序，以及在同一进程中运行并调试组件（D-77）。`alonabot.Run` 在当前进程启动一个完整的 Core（§15.1 的启动与设置，数据目录缺省为工作目录下的 `data/`），并用选项登记开发中的组件：

  ```go
  func main() {
      if err := alonabot.Run(
          alonabot.Plugin(weather.New()),
          alonabot.Adapter(telegram.New()),
      ); err != nil {
          log.Fatal(err)
      }
  }
  ```

- `alonabot.Start(ctx, opts...)` 使用与 `Run` 相同的选项与装配，在全部已启用的包与开发组件激活后返回句柄 `*alonabot.Core`；任一组件在激活前失败时，停止已启动的部分并返回错误。ctx 取消时按 §4.2 停止，句柄的 `Wait()` 等待停止完成，返回与 `Run` 相同的错误。`Run` 等于 `Start` 后等待 SIGINT/SIGTERM 再停止；`Start` 不处理信号，供把 Core 嵌入其他 Go 服务与插件测试套件（§17）使用。同一数据目录仍只能运行一个 Core。
- 组件的 manifest、schema 与 i18n 资源取自 `alonabot build` 生成并嵌入的副本。同一插件 ID 或 adapter_id 被登记两次，或与已安装的包相同时，`Start` 在启动任何组件前返回错误，不以后登记的覆盖。登记的组件按 §15.4 的开发组件规则运行，只是由本进程的 goroutine 经 loopback TCP 连接 Core，握手、Bootstrap、注册、激活、调用与停止都与独立进程相同；已安装的其他包照常以独立进程运行。与独立进程的区别只有两点：
  - 没有进程隔离：SDK 未捕获的 panic 会使整个进程退出，也不执行有限重启；
  - Core 无法强制结束 goroutine：停止预算到期后组件仍未退出时记 ERROR，`Run` 返回错误。
- 库模式没有热部署，需要热部署时用 `alonabot run --reload`（§15.4）。
- 库模式用 `alonabot.Args(...)` 选项或环境变量 `ALONABOT_ARGS` 覆盖调优参数，语法与 `--args` 相同（§16）。
- 官方 Adapter 与官方插件以可导入的 Go 包提供组件（例如 `telegram.New()`），程序入口只调用 SDK 的运行函数，库模式可以直接引用它们。
- 库模式与 `alonabot run` 一样提供已安装的 WebUI 组件。`alonabot.WebUI(fsys fs.FS)` 选项提供 dev 来源的 WebUI：fsys 的根目录即静态根目录，本次运行中它代替已安装的 WebUI，不经安装、签名与 Admin API 版本检查之外的校验。
- 库模式也可以用于小规模部署，但不提供上述隔离与崩溃恢复；正式分发插件与 Adapter 仍使用安装包。

### 15.6 install.sh

- `install.sh` 在 Linux（需要 systemd）与 macOS（launchd）上安装 Core 与官方组件并注册系统服务（D-64）；其他系统与没有 systemd 的 Linux 在改动任何文件之前报错退出，提示改用发行归档或容器镜像。脚本只依赖 POSIX sh、`curl`、`tar`，以及 `sha256sum`（Linux）或 `shasum`（macOS）。

| | Linux | macOS |
| --- | --- | --- |
| 权限 | root | 当前用户，不需要 root |
| 程序目录 | `/opt/alonabot/` | `~/Library/Application Support/AlonaBot/` |
| 数据目录 | `/var/lib/alonabot` | `~/Library/Application Support/AlonaBot/data` |
| 服务 | systemd 单元 `alonabot.service`，以系统用户 `alonabot` 运行 `alonabot run --data-dir <数据目录>`，开机启动 | 用户 LaunchAgent `io.github.alonabot`，登录后启动 |

- 执行顺序：识别 OS/arch → 下载 `--version` 指定的版本（缺省为最新 Release）的发行归档与 `SHA256SUMS` 并校验，不一致时退出且不改动已有安装 → 停止已有服务 → 替换程序目录中的 `alonabot` 与 `packages/` → 数据目录尚无管理凭证时以服务的身份运行 `alonabot admin-token init`，token 只在当前终端显示一次 → 写入服务定义并启动服务。
- 以 `curl … | sh` 运行时，初始化从 `/dev/tty` 读写终端；没有可用终端时跳过初始化，Core 首次启动时按 §15.1 把一次性 token 写入数据目录的 `admin_token`，脚本输出读取它的命令（`alonabot admin-token show --data-dir <数据目录>`）。
- 再次运行即升级：数据目录、`core_config.json` 与已安装组件不变（Core 按 §5.5 迁移设置），新版本 `packages/` 中更高版本的官方组件按 §15.2 作为可用更新列出。`--uninstall` 停止并删除服务定义与程序目录，保留数据目录并输出其路径。
- 只经 HTTPS 下载。`SHA256SUMS` 与归档来自同一 Release，只防止传输损坏，不证明发布者身份；归档内的组件包仍由 Core 按 §9.2 验签。脚本不修改防火墙、反向代理与 `core_config.json`，Core 缺省只监听本机（§15.1）。

## 16. 调优参数

- 契约中可调的容量、期限与配额集中登记在 `internal/tune/params.go`：每项有名称、缺省值、单位与约束，缺省值就是本契约各节给出的数值。新增可调的值时，同时修改该文件与下表。
- 以下不属于调优参数：协议双方必须一致的常量（帧上限、分块大小）；安全相关的值（Secret 与 token 长度、管理会话与插件票据的期限）；平台规定的限制；插件与 Adapter 自己的常量；日志级别与日志文件设置（§14、§15.1）。
- 调优参数只在启动时覆盖，运行中不能修改，也不出现在 WebUI。`--args` 是 `alonabot` 根命令的全局参数，可以写在子命令之前或之后（例如 `alonabot --args REQUEST_TIMEOUT=2s run`），对所有子命令生效，可以重复使用：
  - 值中含 `=` 时是直接参数，多个参数用逗号分隔，例如 `--args DISPATCH_CONCURRENCY=1,REQUEST_TIMEOUT=2s`；
  - 否则是文本文件的路径，文件名任意，每行一个 `KEY=VALUE`，空行和以 `#` 开头的行忽略；
  - 后出现的覆盖先出现的。库模式的 `alonabot.Args(...)` 与 `ALONABOT_ARGS` 语法相同。
- 值的格式：整数用十进制；时长用 Go 的时长写法，例如 `2s`、`15m`；大小用 `B`、`KiB`、`MiB`、`GiB`，写成 `MB` 等十进制单位时报错并提示改用二进制单位。
- 未知的键、格式错误、违反约束或彼此矛盾的值使 Core 拒绝启动，错误指明出处（参数，或文件及行号）。存在任何覆盖时，Core 启动时输出一条 WARN，列出全部覆盖项及出处。`alonabot tune list` 列出全部参数的缺省值与约束。
- SDK 也要遵守的值（在途请求数、payload 预算、普通请求期限、各阶段预算）由 Core 经 Bootstrap 的 `SessionLimits session_limits = 11` 下发：`max_inflight` 对应 `SESSION_MAX_INFLIGHT`，`max_payload_bytes` 对应 `SESSION_MAX_PAYLOAD`，`request_timeout_ms` 对应 `REQUEST_TIMEOUT` 并转换为毫秒。`CORE_MAX_PAYLOAD` 是跨 Session 的 Core 总预算，只由 Core 使用。双方按同一组 Session 值执行。
- 设置 `PLUGIN_MEMORY_SOFT_LIMIT` 时，Core 为每个插件与 Adapter 进程传入环境变量 `GOMEMLIMIT`；`CORE_MEMORY_SOFT_LIMIT` 对 Core 自己生效。这是 Go 运行时的软上限：超过后 GC 更积极地回收，不会终止进程。v0.1 不提供硬上限。

| 参数 | 缺省值 | 约束 | 契约位置 |
| --- | --- | --- | --- |
| `SESSION_MAX_INFLIGHT` | 256 | ≥ 1 | §1.2 |
| `SESSION_MAX_PAYLOAD` | 16MiB | ≥ 帧上限 4MiB | §1.2 |
| `CORE_MAX_PAYLOAD` | 64MiB | ≥ `SESSION_MAX_PAYLOAD` | §1.2 |
| `REQUEST_TIMEOUT` | 30s | > 0 | §1.2 |
| `CALL_CHAIN_MAX_DEPTH` | 16 | ≥ 1 | §3 |
| `DISPATCH_CONCURRENCY` | 64 | ≥ 1 | §2.1 |
| `DISPATCH_QUEUE_EVENTS` | 256 | ≥ 1 | §2.1 |
| `DISPATCH_QUEUE_BYTES` | 8MiB | ≥ 帧上限 4MiB | §2.1 |
| `CONTINUATION_MAX` | 128 | ≥ 1 且 < `SESSION_MAX_INFLIGHT` | §2.4 |
| `CONTINUATION_TIMEOUT` | 15m | > 0 | §2.4 |
| `BUS_QUEUE_EVENTS` | 256 | ≥ 1 | §1.2、§3 |
| `BUS_QUEUE_BYTES` | 4MiB | ≥ 帧上限 4MiB | §1.2、§3 |
| `MEDIA_MAX_OBJECT` | 100MiB | ≤ `MEDIA_SESSION_BUDGET` | §1.3 |
| `MEDIA_MAX_UPLOADS` | 4 | ≥ 1 | §1.3 |
| `MEDIA_SESSION_BUDGET` | 256MiB | ≤ `MEDIA_DISK_BUDGET` | §1.3 |
| `MEDIA_DISK_BUDGET` | 2GiB | > 0 | §1.3 |
| `MEDIA_TTL` | 15m | > 0 | §1.3、§1.4 |
| `PHASE_HELLO` | 10s | > 0 | §4.2 |
| `PHASE_CONFIG` | 60s | > 0 | §4.2 |
| `PHASE_MIGRATION` | 15m | > 0 | §4.2 |
| `PHASE_ONSTART` | 60s | > 0 | §4.2 |
| `PHASE_ONREADY` | 30s | > 0 | §4.2 |
| `STOP_TOTAL` | 10s | ≥ `STOP_DRAIN` + `STOP_ONSTOP` | §4.2 |
| `STOP_DRAIN` | 6s | > 0 | §4.2 |
| `STOP_ONSTOP` | 4s | > 0 | §4.2 |
| `RESTART_WINDOW` | 5m | > 0 | §4.2 |
| `RESTART_MAX_CRASHES` | 3 | ≥ 1 | §4.2 |
| `JOB_TIMEOUT` | 15m | > 0 | §4.3 |
| `TASKS_PER_PLUGIN` | 10000 | ≥ 1 | §4.3 |
| `TASK_PAYLOAD_BYTES` | 64KiB | ≥ 1 且 ≤ 帧上限 4MiB | §4.3 |
| `TASK_CONCURRENCY` | 16 | ≥ 1 且 ≤ `SESSION_MAX_INFLIGHT` | §4.3 |
| `BUTTON_TOKEN_TTL` | 24h | > 0 | MESSAGE-MODEL §6.3 |
| `TOOL_CONFIRM_TIMEOUT` | 5m | > 0 | §13.2 |
| `LOG_BUFFER_RECORDS` | 1000 | ≥ 1 | §14.4 |
| `INSTALL_MAX_ARCHIVE` | 256MiB | > 0 | §9.3 |
| `INSTALL_MAX_EXPANDED` | 1GiB | ≥ `INSTALL_MAX_ARCHIVE` | §9.3 |
| `INSTALL_MAX_ENTRIES` | 10000 | ≥ 1 | §9.3 |
| `DEV_RELOAD_POLL` | 500ms | > 0 | §15.4 |
| `PLUGIN_MEMORY_SOFT_LIMIT` | 不设置 | > 0 | §16 |
| `CORE_MEMORY_SOFT_LIMIT` | 不设置 | > 0 | §16 |

## 17. 插件测试套件

- 包 `github.com/AlonaBot/AlonaBot/plugintest`（D-47）供插件作者在 `go test` 中做集成测试，只在测试代码中使用。`plugintest.Start(t, opts...)` 在 `t.TempDir()` 的数据目录上用 `alonabot.Start`（§15.5）装配真实 Core，登记被测插件（`plugintest.Plugin(root)`，可以有多个）与套件自带的测试 Adapter（adapter_id `plugintest`），并在 `t.Cleanup` 中停止；任一组件激活失败时测试失败。被测插件是库模式的开发组件，配置取自 `plugintest.Config(root, json)`，受管 SQLite 为 `dev.sqlite`。
- 测试 Adapter 只有一个实例和一个固定账号。它的 adapter_id 缺省为 `plugintest`；`plugintest.Adapter(id, declarations...)` 可改为指定的 adapter_id，并声明该 Adapter 的权限与扩展事件（通常取自其常量包），用于测试依赖平台权限或扩展事件的路由。会话分私聊与群聊，用户可带上该会话的平台权限；`plugintest.Admin(user)` 把用户加入 Core Admin。实例缺省声明 `message.send`、`message.markdown`、`message.album`、`message.reference`、`message.buttons`、`message.forward` 与 `message.edit`，`plugintest.Capabilities(...)` 可改为子集：未声明能力对应的片段在发送时返回 UNSUPPORTED，用来测试插件的降级分支。发送与编辑一律 ACCEPTED，消息 ID 按发送顺序递增。
- 注入事件：消息（message.created）、对已发送按钮的点击（button.clicked）、已声明的扩展事件与立即触发一次 schedule。每次注入在这个事件的全部分发链结束，或其中的 Handler 进入多轮等待（§2.4）后返回，结果包含这期间产生的全部 Send 与 Edit（Target、Chain、SendResult）和 Handler 返回的错误。分发结束由 Core 经 `internal/inproc` 通知套件，不靠等待时间判断。分发链之外的后台发送另行按顺序读取。
- `plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error))` 登记一个只提供 `llm.generate` 的测试 Agent（§13.5），被测插件经 `@agent` 的调用交给这个函数处理，不需要真实 Agent 与 Provider（D-77）；没有登记时，`@agent` 按目标未激活处理。
- 套件不跳过 Core 的任何检查：权限、Target 准入（后台发信许可与正式运行一样缺省为全部实例，可以用套件选项收紧）、Capability 预检、Hook 与文案都与正式运行相同。套件属于 SDK 公共 API，兼容承诺与 `sdk` 相同。
