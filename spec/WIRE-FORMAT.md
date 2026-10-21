# AlonaBot v0.1 RPC Wire Format

版本：v0.6-review · 2026-10-21。状态：**现行实施基线，待实施验证**。
权威范围：补充 [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) §1/§3/§12；消息载荷见 [MESSAGE-MODEL.md](MESSAGE-MODEL.md)。产品版本 v0.1 与传输协议版本 **1.0** 分开。T-01 按本文件生成可编译 `.proto` 和兼容 fixture；本文不代表编译或平台测试已通过。

## 1. Frame 与 Envelope

每帧为 **4-byte big-endian unsigned payload length + Protobuf Envelope**，长度不含 4-byte 前缀；有效范围 **1～4,194,304 bytes**。0、超限、无法解码或状态不合法的帧使当前会话失败，不继续猜测帧边界。TCP 与 stdio 使用完全相同的编码。

| Field No. | 字段 | Proto 类型 | 必需性及语义 |
| --- | --- | --- | --- |
| 1 | `protocol_major` | uint32 | 首发为 1；握手主版本不一致拒绝 |
| 2 | `protocol_minor` | uint32 | 首发为 0；握手选择双方支持的较小值，后续帧保持协商值 |
| 3 | `session_id` | bytes | HELLO 为空；HELLO_ACK 起使用 Core 分配的 16-byte 随机 ID |
| 4 | `initiator_side` | enum | UNSPECIFIED=0、CORE=1、PLUGIN=2；REQUEST/RESPONSE/CANCEL 使用原请求的发起方向 |
| 5 | `call_id` | uint64 | REQUEST/RESPONSE/CANCEL 非零；其他控制帧及 EVENT 为 0；每 Session 每发起方向单调递增、不复用 |
| 6 | `parent_call_id` | optional uint64 | 仅 PLUGIN 发起的 REQUEST 可设置，引用同一 Session 中仍在途的 CORE 发起父 REQUEST；缺省表示独立插件调用 |
| 7 | `deadline_unix_ms` | optional int64 | REQUEST 必需；UTC 毫秒，子请求不得延长父期限；控制帧按阶段定义 |
| 8 | `kind` | enum | UNSPECIFIED=0、HELLO=1、HELLO_ACK=2、REGISTER=3、PREPARED=4、ACTIVATE=5、ACTIVATED=6、REQUEST=7、RESPONSE=8、CANCEL=9、EVENT=10、STOPPING=11、HEARTBEAT=12、SESSION_ERROR=13 |
| 9 | `source_plugin_id` | string | 插件发出的帧必须等于其认证 ID；Core 发出的帧为空，原业务调用者由可信调用上下文传递，不冒充传输对端 |
| 10 | `target` | string | 已注册 route/service 或固定 Core 服务名；不用文件路径；无目标的控制帧为空 |
| 11 | `trace_id` | bytes | 有业务上下文时为 Core 关联的 16-byte ID；不是权限凭证；Core 不采信插件用它替换 actor/scope |
| 12 | `payload_encoding` | enum | UNSPECIFIED=0、PROTOBUF=1、JSON=2；控制帧使用 PROTOBUF |
| 13 | `schema_id` | string | 固定 `alonabot.*.v1` 或服务声明的契约 ID；有 payload 时必需 |
| 14 | `schema_version` | uint32 | 有 payload 时必需且首发为 1；独立于 transport major/minor |
| 15 | `payload` | bytes | 按 encoding/schema 解释，连同 Envelope 其余字段一起计入 4 MiB |
| 16 | `error` | message ErrorInfo | RESPONSE 或 SESSION_ERROR 失败时使用；RESPONSE 可同时包含有效结果 payload 与 error |

所有枚举在 `.proto` 中使用带类型前缀的名称，首项为 `*_UNSPECIFIED=0`，表中省略前缀以便阅读。REQUEST/RESPONSE/CANCEL 的 initiator_side、所有帧的 kind 和带载荷帧的 encoding 不接受 UNSPECIFIED；未知 kind 不能当作成功忽略。第 6/7 字段使用 `optional`；其余字段按帧种类校验语义，minor=0 是有效值，不要求用缺失判断代替版本判断。

原有 Envelope 1～16 和错误码 1～17 保持不变。删除字段或枚举值时 reserved 编号及名称；不复用编号，不把修改字段类型解释为普通 minor 更新。

## 2. ErrorInfo 与结果

| Field No. | 字段 | 类型 | 说明 |
| --- | --- | --- | --- |
| 1 | `code` | enum | ERROR_UNSPECIFIED=0；有 ErrorInfo 时必须是非零错误 |
| 2 | `safe_message` | string | 对外诊断，不含 secret、堆栈或本地绝对路径 |
| 3 | `trace_id` | bytes | 与 Envelope 的业务追踪 ID 一致；认证前不返回详细诊断 |

固定错误码：INVALID_ARGUMENT=1、UNAUTHENTICATED=2、PERMISSION_DENIED=3、NOT_FOUND=4、ALREADY_EXISTS=5、FAILED_PRECONDITION=6、UNSUPPORTED=7、RESOURCE_EXHAUSTED=8、DEADLINE_EXCEEDED=9、CANCELLED=10、UNAVAILABLE=11、INTERNAL=12、PROTOCOL_ERROR=13、CALL_CYCLE=14、PLUGIN_NOT_READY=15、NO_DEFAULT_TARGET=16、NEEDS_MANUAL_MIGRATION=17。未知错误码对调用者映射为 INTERNAL，原数字保留在内部诊断。

RESPONSE 的 payload 与 ErrorInfo **不是 oneof**。例如部分发送成功时返回 SendResult 和下游错误；SDK 不能因 error 非空丢弃已确认的消息 ID。成功使用无 error 的结果；失败没有有效业务结果时才省略 payload。SESSION_ERROR 只表示会话阶段失败，不伪造业务 call_id。

## 3. 控制帧与 Bootstrap

### 3.1 帧方向和阶段

| Kind | 方向及 Payload | 前置条件与处理 |
| --- | --- | --- |
| HELLO | Plugin→Core：plugin_id、secret(32 bytes)、supported_major/minor、build_digest、supported_capabilities、required_capabilities | 首帧且仅一次；身份、包摘要或协议不符直接关闭；build_digest 由 Core 从已验证归档安装记录经启动上下文提供供 SDK 回显，不内嵌整包自摘要；未知必需能力拒绝，不安装路由 |
| HELLO_ACK | Core→Plugin：session_id、chosen_minor、capability_intersection、Bootstrap | HELLO 验证通过；SDK 以后仅用该 Session；Bootstrap 一次提供准备所需输入 |
| REGISTER | Plugin→Core：完整 RouteDescriptor 列表、config_schema_hash、规范化 candidate_config_json、candidate_config_version、可选 HTTP listener 描述、迁移完成信息 | 配置转换/默认值/校验和 OnStart 完成后仅发送一次；Core 原子校验整个注册集与候选配置，失败不保留半个注册表 |
| PREPARED | Core→Plugin：route_descriptor_hash、candidate_config_digest | REGISTER 校验通过；digest 由 Core 对准备持久化的快照计算，SDK 作为 opaque 确认值；仍不开放业务 |
| ACTIVATE | Core→Plugin：generation_id、commit_marker、OnReady deadline | Core 状态库已提交唯一 ActiveGeneration；SDK 执行 OnReady |
| ACTIVATED | Plugin→Core：generation_id、成功状态 | OnReady 成功后由唯一 writer 完整写出；Core 在读循环内同步应用该状态，随后才处理后续帧及开放路由 |
| REQUEST | 双向：target、payload、schema、deadline、非零 CallID | 正常业务要求 Activated；调用前由 Core 鉴权及 admission，最多完成一个 RESPONSE |
| RESPONSE | 被调用方→原发起方：原 Session/initiator_side/CallID、payload 和/或 ErrorInfo | 独立配对；重复、未知或取消后迟到的响应不二次完成；scheduled 的已知迟到完成响应仅释放任务运行槽 |
| CANCEL | 原发起方→被调用方：原 initiator_side/CallID、reason | 通知协作取消；接收方取消相应 context，不代表外部副作用撤销 |
| EVENT | Core→Plugin：BusEvent，call_id=0 | 仅用于 Event Bus 已接纳的异步广播；无业务 ACK、重放或完成承诺；scheduled 不使用此帧 |
| STOPPING | Core→Plugin：绝对停止 deadline、reason | Core 已冻结新派发并结束 drain/取消剩余调用；SDK 使用独立的剩余期限 context 执行 OnStop，然后退出 |
| HEARTBEAT | 双向：seq、monotonic timestamp | 仅用于会话诊断，不触发业务重试，不把一端单调时钟与另一端直接相减 |
| SESSION_ERROR | 双向：failed_phase、failed_kind；Envelope.error 为 ErrorInfo | 认证后用于配置/OnStart/注册/OnReady/状态机失败；发送后关闭会话，由 Core 按提交前后规则恢复或进入 Degraded |

能力项使用 `name + version`，supported 列表是可用集合，required 是必须包含于交集的子集。Major 必须相同，minor 选择 `min(local,peer)`；不能凭相同 minor 假定对方支持未声明能力。

### 3.2 Bootstrap 的最小内容

Bootstrap 属于私有、已认证 IPC，和整个 HELLO_ACK 一起受 **4 MiB** 限制。超限明确失败；首发不增加分块 Bootstrap 或准备期业务 RPC。

| 字段 | 输入语义 |
| --- | --- |
| `generation_id` | Core 分配的不可变代次标识，插件只回显、不选择其他代次 |
| `candidate_config_json` | UTF-8 JSON 对象，Core 先验证对象/大小并解析既有 Secret 引用；允许旧版本字段进入 SDK 配置迁移，不先用新版 Schema 否决旧配置 |
| `config_schema_hash`、`previous_config_version`、`target_config_version` | 与可信安装包对应的 Schema 和配置迁移输入；无配置插件仍以空对象表示 |
| `database_descriptor` | 后端类型、受管 SQLite 当前准备视图路径或该插件 PostgreSQL 连接参数、迁移模式、已应用版本、activation_reason（first_install/resume/reload/update/rollback）、manual_migration_approved（仅Core显式维护许可且绑定当前Session/Generation/DB时true）；不下发 Core 管理数据库凭证 |
| `http_proxy_auth` | 本 Session 独立的内部代理认证材料；只在插件声明 HTTP 能力时需要，不复用 Admin 凭证 |
| `phase_limits`、`preparation_deadline_unix_ms` | 主契约固定的准备、迁移和 Ready 预算及本轮绝对期限；迁移使用独立预算，不被普通业务 30s 截断 |

SDK 按配置契约执行 `OnLoad → 补缺失默认值 → typed decode/Validate → DB 迁移门槛核对及许可的迁移 → OnStart`（PG resume 不重跑 Migrate/Up，见 CONTRACTS §6.3），并在 REGISTER 回传完整、规范化的候选 JSON；私有载荷可包含供插件使用的已解析 Secret。Core 只为新候选抽取 Secret 并持久化新的引用快照，再计算 PREPARED 的 candidate digest；普通 restart 仅验证原快照一致性，复用原 Generation、Secret 引用与 COMMITTED 记录。Bootstrap、候选完整 JSON、DB 凭证及 HTTP 认证材料不得进入普通日志或管理 API 响应。

准备所需数据由 Bootstrap 一次下发，因此 OnStart/OnReady **不能发起普通业务 Send、Service、Publish，也不能等待自身 ready gate**。数据库操作遵循 Candidate 视图及主契约的副作用边界。后台任务可在 OnReady 注册启动意图；SDK 在唯一 writer 成功写完整个 ACTIVATED 后才放开 ready gate。Core 按同一连接帧顺序先处理 ACTIVATED，后处理紧随其后的业务 REQUEST；无额外 ActivatedAck。

### 3.3 注册描述与 HTTP Listener

RouteDescriptor 公共字段为 `route_id`、`kind`、`priority`、`default_enabled`、`required_permission`、输入/输出 `schema_id/version`。ID 在一个插件内唯一且跨版本保持业务含义稳定；`target` 使用 `plugin_id/route_id`。各 kind 再提供必要字段：

| kind | 专属描述 |
| --- | --- |
| command | 规范命令路径、aliases、参数列表与默认值、Handler 模式、Rule 描述 |
| event | event_type_filter、Handler 模式、Rule 描述；button.clicked 路由含 action schema version |
| hook | phase（PreEvent/BeforeSend/AfterSend）、优先级及固定 Hook 输入/输出 Schema |
| service | service_name、公开/allowlist 能力、请求/响应 Schema |
| schedule | 稳定 job_id、cron/interval、timezone、固定 job deadline |
| http | method、插件相对 path、is_public、认证/回调验证模式 |
| bus_subscription | event_name、可接受的 Schema/版本、订阅回调 route_id |

自定义 Go Rule 是插件进程内与对应 Handler 一起执行的函数；Core 可表达的类型/命令/权限条件先执行，SDK 在同一 Handler REQUEST 中完成自定义 Rule，并以 HandlerOutcome.matched 返回结果，不另建 Rule 服务。

REGISTER 的完整描述必须装入一帧；超限 RESOURCE_EXHAUSTED，不部分接收或激活。配置 Schema 与服务 Schema 从经过包摘要校验的安装资源取得，REGISTER 中的 hash/版本必须一致，不能仅上报不可解析的 hash 而缺少 Schema 原文。

有 HTTP 路由时，SDK 先创建实际绑定的 loopback Listener，再上报 `host + port`；首发仅接受 `127.0.0.1` 或 `::1` 与合法非零端口，不能传任意 URL、文件路径或远端主机。Core 验证与当前 Session 的代理认证和就绪状态后保存端点，公开前缀仍由 Core 决定。无 HTTP 路由则不得要求创建 Listener。端点随 Session 失效撤销；外部请求不携带内部认证材料，Core 覆盖相应内部头。

## 4. 调用上下文、读写与取消

- 匹配键为 `(session_id, initiator_side, call_id)`。两个方向可以同时使用 call_id=7；Core 转发 A→B 时建立两段 Session 的内部映射，不要求转发后的 CallID 相同。
- PLUGIN REQUEST 有 parent_call_id 时，Core 必须找到同一 Session 中仍在途的 CORE REQUEST，才能继承该父调用的 actor、target、scope、trace 和调用链；不存在或已经结束则 FAILED_PRECONDITION。无 parent 的调用只具有插件身份，不接受 payload 中的 Admin 标志作为授权。Core 发出的 Request 不靠该字段跨 Session 指认父调用，映射由 Broker 内部维护。
- 可信调用上下文随 Core→Plugin 的 Invocation 载荷下发供 SDK 使用；插件回传的 actor、scope、trace 等值不能覆盖 Core 记录。所有子调用采用 `min(parent_deadline,operation_deadline)`；父调用完成/取消时取消尚未结束的子调用。独立后台工作必须使用插件身份和独立 context。
- 同一 Service 调用链的插件 ID 重复返回 CALL_CYCLE，链深超过 16 返回 FAILED_PRECONDITION；独立并发链互不影响。Event Bus 回调没有可继承的用户调用链，后续 Service 使用插件身份。
- 每连接只有一个连续读帧/分流循环、一个串行 writer；Handler 不在读循环内执行。reader 同步处理 Session 状态变化，再将业务提交给执行器；Response/Cancel/控制帧及时分流，不排在正在等待 RPC 的 Handler 后面。
- 每 Session 每方向最多 256 个 outstanding 业务请求；请求/事件载荷另受每 Session 每方向 16 MiB、Core 合计 64 MiB 的在途及排队字节预算约束；任一满额立即 RESOURCE_EXHAUSTED，不等待槽位。Response、Cancel、状态控制走独立、有界且优先的通道，不占用普通业务 admission；无法维持协议时使会话失败，不能无界缓存。优先队列不得使 Cancel 越过自己的 Request：尚未写出的被取消 Request 直接撤出且不再发送，已完整写出后才可随后发送 Cancel。
- 本地超时/取消先完成调用者失败返回，再尽力发送 Cancel；已发送外部消息/SQL 不撤销，不自动重试。Session 失效时清理映射、取消调用并返回 UNAVAILABLE；已确认的 SendResult 仍须保留。
- Core 停用时先冻结该插件的新根调用/新派发，最多 drain 6s；已有父 Request 链中的受限子请求与响应可继续。drain 结束后取消剩余调用并发送唯一 STOPPING，帧 deadline 为 min(t0+10s,发送STOPPING时刻+4s)，OnStop 与资源关闭使用这一独立 context；整个进程树停止预算仍为 10s。Core 同时停止多个插件时共享同一绝对截止时间，不逐个累加预算。

## 5. Event Bus、媒体与平台载荷

- Event Bus 的发布通过普通 REQUEST 返回“是否被 Core 接纳”的结果；它不确认订阅者已处理。Core 向订阅者发 EVENT。BusEvent 至少包括 event_name、Core event_id、source_plugin_id、occurred_at、schema_id/version、payload。发布者只能发布自身 `plugin_id/event_name` 命名空间；订阅者 Actor 缺省，业务 payload 不产生管理权限。
- 每订阅者最多 256 个且 4 MiB 待处理事件；任一满额丢弃新事件并计数，插件离线不积压、不补发。业务处理无 ACK、自动重试或 StopPropagation。Scheduler 的 scheduled 通过有 CallID/独立 deadline 的 REQUEST/RESPONSE 判定完成与重叠。
- 服务 JSON 使用 `payload_encoding=JSON` 和注册的 schema/version，严格验证 UTF-8 JSON 及结构；若使用 Protobuf Any，必须是真实的 type_url 和 Protobuf bytes。未知 encoding/schema/version 明确拒绝。
- MediaRef 为 Core 分配、受 Session 和目标范围约束的 capability；上传请求包含 upload_id、offset、chunk bytes。顺序分块只接受当前已确认 offset；重复或跳跃 offset 返回 INVALID_ARGUMENT，不静默追加。提交校验声明总量与 sha256 后才产生 MediaRef；未提交上传、过期引用和消费租约按主契约清理。
- MessageChain/Event/Actor/Target/HandlerOutcome/SendResult 的载荷遵循 [MESSAGE-MODEL.md](MESSAGE-MODEL.md)。公共 JSON 标识符使用字符串；媒体数据不直接塞进普通 MessageChain JSON。

Scheduler 到 deadline 后保留实际执行槽；SDK 只在 callback 实际返回后发送唯一 Response，迟到结果只用于释放 running，不把超时重新标为成功。未取得实际完成响应时，须确认旧进程退出才释放；详见 CONTRACTS §4.3。

Core 媒体服务固定为 BeginUpload/WriteChunk/CommitUpload/Read/Grant：Read 输入 token、offset、length（1～1 MiB），返回 bytes、next_offset、eof；Grant 只对已有授权签发等同或更窄范围的新 token，不允许调用者自行扩大 Target/接收插件。IncomingMedia 首次取件由内置 Adapter 解析，按 CONTRACTS §1.4 校验和配额处理。各请求/响应的 Protobuf 字段号在 T-01 固定，媒体块始终计入帧与字节预算。

## 5A. T-01 已固定（2026-10-21）

字段号与枚举值的权威来源是 `contracts/proto/alonabot/v1/*.proto`（package `alonabot.v1`）；本节不再重复抄录编号。`contracts/wirespec_test.go` 逐项断言 §1/§2 表格中的编号，以及 SendStatus、PartStatus、Propagation、RouteKind、HookPhase 的取值，并检查所有枚举的零值都是 `*_UNSPECIFIED`。冻结之后，删除字段或枚举值时必须 reserved，不能复用编号。

| 文件 | 内容 |
| --- | --- |
| `envelope.proto`、`errors.proto` | §1 Envelope 1～16、§2 ErrorCode 1～17 与 ErrorInfo（编号照抄规格） |
| `control.proto` | §3.1/§3.2 的控制帧载荷与 Bootstrap |
| `route.proto` | REGISTER 载荷 `Register`、§3.3 RouteDescriptor 及各 kind 的描述 |
| `identity.proto`、`message.proto`、`media.proto` | MESSAGE-MODEL 的 Target/Actor/Capability、Event 与按类型区分的载荷、Segment、SendRequest/SendResult/PartResult，以及 §5 的媒体请求 |
| `invocation.proto`、`bus.proto` | §4 可信调用上下文、HandlerOutcome、Hook 载荷，以及 §5 的 Event Bus |

规格只给出语义、由 T-01 确定的实现选择：

- 没有编号表的消息，字段号按规格表格中的字段顺序从 1 开始编排。
- 新增枚举 CancelReason、StopReason、SessionPhase、DatabaseBackend（显式 `NONE` 表示没有业务 DB）、MigrationMode、ActivationReason、TrustedRole（USER/ADMIN/CHAT_ADMIN）、Permission（ANY/USER/ADMIN）、HandlerMode、CommandMatch（PATH/REGEX/RAW）、UsageMode、ParamType、ServiceAccess（PRIVATE/ALLOWLIST/PUBLIC）、HttpAuthMode、TextFormat、EntityType、MediaReadability。
- 为避免 import 循环，Target/Actor/Capability 单独放在 `identity.proto`；Segment 的 `incoming_ref` 直接携带 IncomingMedia 描述符，`media_ref` 是 MediaRef 的 token。
- Event 按类型区分的载荷放在 `EventPayload` oneof 中，PreEventResult 只能回传这一部分；按钮 `value` 与服务载荷之外的 JSON 一律以 UTF-8 文本表示。
- `ForwardSource` 的结构（display_name、external_user_id）由 T-01 补定；新增 `SendRequest{to, chain}` 作为 Core Send 服务的请求。
- Core→目标插件的 Service 调用载荷是 `ServiceInvocation{route_id, context, encoding, schema_id, schema_version, payload}`：可信上下文与业务载荷分开传输，服务响应直接使用 output schema 的业务载荷（REVIEW-1 补充）。
- 协议能力项使用 `ProtocolCapability{name, version}`，与 Adapter 平台能力 `Capability{name, version, supported}` 分开。

## 6. 实施验证与依据

T-01 将字段表转换为 `.proto`，冻结控制 Payload/RouteDescriptor 的字段号和 fixture；T-02 验证双向编号、乱序响应、失败阶段、帧上限、父调用关联、读循环进度和 ready gate。T-03/T-05/T-06 验证 payload+error、HandlerOutcome、Hook 约束和 Event Bus；对应 [TEST-MATRIX.md](TEST-MATRIX.md)。目前这些运行时测试均未执行。

Protobuf 枚举零值与字段存在性依据：[官方 proto3 指南](https://protobuf.dev/programming-guides/proto3/)。权限和准备期限制是 AlonaBot 的协议选择，不代表对同宿主恶意 Native 代码提供 OS 沙箱。
