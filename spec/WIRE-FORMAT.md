# AlonaBot v0.1 RPC Wire Format

版本：v0.6-review · 2026-10-21。状态：**现行实施基线，待实施验证**。
权威范围：补充 [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) §1/§3/§12；消息载荷见 [MESSAGE-MODEL.md](MESSAGE-MODEL.md)。产品版本 v0.1 与传输协议版本 **1.0** 分开。T-01 按本文件生成可编译 `.proto` 和兼容 fixture；本文不代表编译或平台测试已通过。

身份字段中的组件 ID 原样使用 CONTRACTS §9.1 的拼写（D-172），例如 HELLO 的 `plugin_id` 与 Envelope 的 `source_plugin_id` 为 `alonabot-plugin-configdb`，Service 目标为 `alonabot-plugin-configdb/<服务名>`；Adapter 包身份如 `alonabot-adapter-telegram`，其平台 `adapter_id` 仍为 `telegram`。本次只改变 ID 字符串的值，字段号、枚举值与 schema 版本保持不变。

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
| 14 | `schema_version` | uint32 | 有 payload 时必需且 ≥ 1；控制帧与 Core 固定载荷首发为 1，REQUEST/RESPONSE 携带的 Service 业务载荷取服务注册的版本（D-66）；独立于 transport major/minor |
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
| REGISTER | Plugin→Core：完整 RouteDescriptor 列表、config_schema_hash、规范化 candidate_config_json、配置校正记录（D-161；`candidate_config_version` 已删除，字段号保留）、可选 HTTP listener 描述、迁移完成信息 | 配置转换/默认值/校验和 OnStart 完成后仅发送一次；Core 原子校验整个注册集与候选配置，失败不保留半个注册表 |
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

已定义的协议能力：`http_listener` 版本 1（D-65）。插件存在 HTTP 路由时 SDK 把它放进 HELLO 的 required；Core 支持它，交集含该项时才在 Bootstrap 中下发 `http_proxy_auth`，REGISTER 带 HTTP 路由而交集不含该项时拒绝注册。

StopReason 在现有枚举末尾追加 `STOP_REASON_IDLE`（goal-170），表示正常空闲退出，沿用 STOPPING 的绝对 deadline 与停止预算。编号由获授权的切片按当前 `.proto` 顺延，已有值保持不变。`SUSPENDED` 是 Core 管理运行状态，SessionPhase 沿用现有状态。唤醒复用 `ActivationReason.RESUME`、原 generation_id 与新 Session，不增加业务 RPC。

### 3.2 Bootstrap 的最小内容

Bootstrap 属于私有、已认证 IPC，和整个 HELLO_ACK 一起受 **4 MiB** 限制。超限明确失败；首发不增加分块 Bootstrap 或准备期业务 RPC。

| 字段 | 输入语义 |
| --- | --- |
| `generation_id` | Core 分配的不可变代次标识，插件只回显、不选择其他代次 |
| `candidate_config_json` | UTF-8 JSON 对象，Core 先验证对象/大小并解析既有 Secret 引用；允许旧版本字段进入 SDK 配置迁移，不先用新版 Schema 否决旧配置 |
| `config_schema_hash` | 与可信安装包对应的规范 Schema 摘要，含其中的 `example` 声明（CONTRACTS §5.1，D-160）；示例图片字节由包携带，经 Admin API 提供。无配置插件仍以空对象表示。`previous_config_version`（4）与 `target_config_version`（5）已由 D-161 删除，字段号保留不再使用 |
| `database_descriptor` | 后端类型（NONE 或受管 SQLite）、受管 SQLite 当前准备视图路径、迁移模式、已应用版本、activation_reason（first_install/resume/reload/update/rollback）、manual_migration_approved（仅Core显式维护许可且绑定当前Session/Generation/DB时true）；不下发 Core 管理数据库凭证 |
| `http_proxy_auth` | 本 Session 独立的内部代理认证材料；只在 HELLO 协商出 `http_listener` 能力时下发（§3.1），不复用 Admin 凭证 |
| `phase_limits`、`preparation_deadline_unix_ms`（字段号 8、9） | 准备、迁移和 Ready 的阶段预算及本轮绝对期限；迁移使用独立预算，不被普通业务 30s 截断 |
| `LogLevel log_level`（字段号 10） | 该来源当前生效的日志级别（CONTRACTS §14.3），SDK 据此在进程内过滤；枚举 `LogLevel` 的编号为 `LOG_LEVEL_UNSPECIFIED = 0`、`LOG_LEVEL_DEBUG = 1`、`LOG_LEVEL_INFO = 2`、`LOG_LEVEL_WARN = 3`、`LOG_LEVEL_ERROR = 4`、`LOG_LEVEL_OFF = 5`、`LOG_LEVEL_CRITICAL = 6`（D-158）。UNSPECIFIED 按 INFO 处理，OFF 关闭输出，CRITICAL 的阈值高于 ERROR；`inherit` 由 Core 先解析成全局级别后再下发 |
| `SessionLimits session_limits`（字段号 11） | 本次运行生效的每方向在途请求数、每 Session payload 预算与普通请求期限；Core 全局 payload 预算只由 Core 使用 |
| `string default_locale`（字段号 12） | Core 的全局默认语言，Session 内保持不变（CONTRACTS §5.4） |
| `string plugin_data_dir`（字段号 13） | 启动前创建的插件自管数据目录的绝对路径（CONTRACTS §6.1，D-165） |
| `profile_overrides_json`（T-04.18） | UTF-8 JSON 对象：配置文件 ID → `{字段路径: 值}`，该插件当前有效的全部覆盖值（CONTRACTS §5.6，D-168）；T-08.18 按此形状的编码字节数校验覆盖集合，实际帧另受 §1 的 Envelope 长度限制；按 §5A 的只追加规则编号 |

`Bootstrap` 的字段号固定为 `log_level = 10`、`session_limits = 11`、`default_locale = 12`、`plugin_data_dir = 13`。`SessionLimits` 的消息结构如下，字段号不得重排或复用：

```proto
message SessionLimits {
  uint32 max_inflight = 1;
  uint32 max_payload_bytes = 2;
  uint32 request_timeout_ms = 3;
}
```

`max_inflight` 是每个 Session、每个发起方向的在途业务请求上限；`max_payload_bytes` 是该 Session 在途与排队请求/事件载荷的合计预算；`request_timeout_ms` 是普通业务请求的缺省期限。T-02.2 依赖 T-02.1 合并后再修改同一 `.proto` 文件，避免并行生成代码冲突。
| `phase_limits`、`preparation_deadline_unix_ms` | 主契约固定的准备、迁移和 Ready 预算及本轮绝对期限；迁移使用独立预算，不被普通业务 30s 截断 |

SDK 按配置契约执行 `OnConfigLoad → 补缺失值与校正 → required 校验 → DB 迁移门槛核对及许可的迁移 → OnStart`，并在 REGISTER 回传完整、规范化的候选 JSON；私有载荷可包含供插件使用的已解析 Secret。Core 只为新候选抽取 Secret 并持久化新的引用快照，再计算 PREPARED 的 candidate digest；普通 restart 仅验证原快照一致性，复用原 Generation、Secret 引用与 COMMITTED 记录。Bootstrap、候选完整 JSON、DB 凭证及 HTTP 认证材料不得进入普通日志或管理 API 响应。

准备所需数据由 Bootstrap 一次下发，因此 OnStart/OnReady **不能发起普通业务 Send、Service、Publish，也不能等待自身 ready gate**。数据库操作遵循 Candidate 视图及主契约的副作用边界。后台任务可在 OnReady 注册启动意图；SDK 在唯一 writer 成功写完整个 ACTIVATED 后才放开 ready gate。Core 按同一连接帧顺序先处理 ACTIVATED，后处理紧随其后的业务 REQUEST；无额外 ActivatedAck。

### 3.3 注册描述与 HTTP Listener

RouteDescriptor 公共字段为 `route_id`、`kind`、`priority`、`default_enabled`、`required_permissions`（权限名集合，满足任一即可，D-69）、输入/输出 `schema_id/version`。ID 在一个插件内唯一且跨版本保持业务含义稳定；`target` 使用 `plugin_id/route_id`。各 kind 再提供必要字段：

| kind | 专属描述 |
| --- | --- |
| command | 规范命令路径、aliases、参数列表与默认值、Handler 模式、Rule 描述 |
| event | event_type_filter、Handler 模式、Rule 描述、可选 message_filter（CONTRACTS §2.5）；button.clicked 路由含 action schema version |
| hook | phase（PreEvent/BeforeSend/AfterSend）、event_type_filter（PreEvent 恰好一项，为已知标准事件或合法 Adapter 扩展事件名；BeforeSend/AfterSend 留空）、优先级及固定 Hook 输入/输出 Schema |
| service | service_name、公开/allowlist 能力、请求/响应 Schema |
| schedule | 稳定 job_id、cron/interval、timezone、固定 job deadline |
| task | 稳定任务名；实例的触发规则与参数不在注册描述中，由 `core.task.set` 在运行时给出（D-124） |
| http | method、插件相对 path、is_public、认证/回调验证模式 |
| bus_subscription | event_name、可接受的 Schema/版本、订阅回调 route_id |

EventSpec 追加可选消息 `MessageFilter message_filter`；MessageFilter 的字段为 `bool to_me`、`repeated ChatType chat_types`、`string wake_words_config`、`string triggers_config`。字段号在实施时按最新 proto 顺延，不在规格中预分配。字段路径、OR、触发方式选择、缺失数据与空列表的语义见 CONTRACTS §2.5；过滤器的存在与缺省有区别。仅允许标准 message.created/edited/deleted 事件，和其他事件类型混用时拒绝注册。SDK 用自有类型登记这些声明，不在公开 API 暴露 pb。过滤器及配置字段路径参与确定性路由编码、摘要和快照白名单，配置值及按配置文件的覆盖值保持在 Core 的配置所有者中。

自定义 Go Rule 是插件进程内与对应 Handler 一起执行的函数；Core 可表达的类型/命令/权限条件先执行，SDK 在同一 Handler REQUEST 中完成自定义 Rule，并以 HandlerOutcome.matched 返回结果，不另建 Rule 服务。

REGISTER 的完整描述必须装入一帧；超限 RESOURCE_EXHAUSTED，不部分接收或激活。配置 Schema 与服务 Schema 从经过包摘要校验的安装资源取得，REGISTER 中的 hash/版本必须一致，不能仅上报不可解析的 hash 而缺少 Schema 原文。

有 HTTP 路由时，SDK 先创建实际绑定的 loopback Listener，再上报 `host + port`；首发仅接受 `127.0.0.1` 或 `::1` 与合法非零端口，不能传任意 URL、文件路径或远端主机。Core 验证与当前 Session 的代理认证和就绪状态后保存端点，公开前缀仍由 Core 决定。无 HTTP 路由则不得要求创建 Listener。端点随 Session 失效撤销；外部请求不携带内部认证材料，Core 覆盖相应内部头。

插件页面的 HTTP 请求由 WebUI 宿主转发，复用本节的路由注册与代理认证规则。主题、语言、令牌与快捷键消息由浏览器内的桥处理，消息格式见 [ADMIN-API §7](ADMIN-API.md#7-插件页面-postmessage-桥)。

路由快照只保存以上 RouteDescriptor 的确定性编码、内容摘要及其插件/已提交代次/包身份，规则见 CONTRACTS §4.4；摘要复用 PREPARED 的 `route_descriptor_hash` 算法。REGISTER 中的配置、HTTP Listener 与任何 Session 认证材料均排除在快照外。唤醒重新 REGISTER 后在业务开放前核对摘要，HTTP 重新绑定本 Session 的 Listener 与凭证。

### 3.4 Adapter 会话

Adapter 包使用与插件相同的帧、握手、Bootstrap、准备与激活流程（CONTRACTS §8、D-26），区别如下：

- 启动的是插件还是 Adapter 由 Core 依据安装记录决定，不采信进程自报。Adapter 的 REGISTER 必须携带 Adapter 描述（adapter_id、平台 Capability、extension 声明及其 Schema、平台动作声明，D-92），可以携带 HTTP 路由与实际监听的 loopback endpoint（D-94、D-136，路径、内部代理认证与生命周期规则同插件，CONTRACTS §7.2），不得携带其他插件路由或迁移信息；插件的 REGISTER 不得携带 Adapter 描述。不一致按注册失败处理。
- Core 与 Adapter 之间的固定服务如下（共十二个），服务名不含 `/`：

| 服务 | 方向 | 请求 → 响应 |
| --- | --- | --- |
| `adapter.instance.start` | Core→Adapter | instance_id、实例配置 JSON（含已解析凭证）、已保存的游标 → account_id、账号展示名、账号用户名（`account_username`，平台上可寻址的标识，用于 `/cmd@username` 与 mention；平台没有时 absent，D-67）、该实例的 Capability |
| `adapter.instance.stop` | Core→Adapter | instance_id → 空结果；实例停止接收，不再交付事件 |
| `adapter.send` | Core→Adapter | instance_id、Target、MessageChain → SendResult 与/或 ErrorInfo，语义同 MESSAGE-MODEL §4 |
| `adapter.edit` | Core→Adapter | instance_id、Target、platform_message_id、MessageChain → SendResult 与/或 ErrorInfo，语义同 MESSAGE-MODEL §4 的 Edit |
| `adapter.media.read` | Core→Adapter | instance_id、平台文件标识、offset、length（1～1 MiB）→ bytes、next_offset、eof |
| `adapter.action` | Core→Adapter | instance_id、Target?（作用范围为 chat 时必有）、动作完整名、版本、JSON 参数 → JSON 结果与/或 ErrorInfo（D-92） |
| `adapter.delete` | Core→Adapter | instance_id、Target、platform_message_ids（1～100）→ 每个 ID 的状态（ACCEPTED / REJECTED / UNKNOWN / NOT_ATTEMPTED）与安全错误（MESSAGE-MODEL §4A） |
| `adapter.commands.sync` | Core→Adapter | instance_id、命令列表（name、descriptions：locale → 文本，英文原文与路径覆盖按 CONTRACTS §8）→ 空结果或 ErrorInfo（D-96、D-138） |
| `core.adapter.deliver` | Adapter→Core | instance_id、Event、交付后的游标 → 是否接纳；Core 队列满时为未接纳并计数 |
| `core.adapter.cursor` | Adapter→Core | instance_id、游标 → 空结果；用于没有事件时的游标变化 |
| `core.adapter.status` | Adapter→Core | instance_id、实例状态、安全错误 → 空结果；用于 401/409 等使实例停止接收的错误 |
| `core.media.url` | Adapter→Core | Core 为本次发送签发的只读 token、可选的 URL 前缀 → 临时媒体 URL 与到期时间（CONTRACTS §1.3，D-88） |

- 命令说明（D-138）：T-03.17 在 `route.proto` 的 `CommandSpec` 追加 `optional string description`，保存 builder 登记的英文原文；覆盖键由现有规范 path（Group 前缀在内，token 以单个 ASCII 空格连接）推导。生成器不扫描注册，check/build 只检查顶层 commands 的结构（键为字符串，值为含字符串 description 的对象）；SDK 的 Collect 对没有对应已注册命令或命令未用 Description 登记英文原文的覆盖路径拒绝注册，错误为英文并带出路径。T-03.17 同时在 `adapter.proto` 追加 `CommandMenuItem{name, descriptions}`、`AdapterCommandsSyncRequest{instance_id, commands}` 与空的 `AdapterCommandsSyncResponse`，descriptions 为 string → string 的 locale 文本映射。字段号按合入时 main 追加，不预分配；`core.command.list` 复用本字段与翻译规则。
- 平台 Capability（D-135）在 `identity.proto` 的 `Capability` 中追加可选消息字段 `TextLimit text_limit`；`TextLimit` 含 `uint64 max` 与 `TextUnit unit`，枚举取 `TEXT_UNIT_UNSPECIFIED`、`TEXT_UNIT_UTF16_CODE_UNITS`、`TEXT_UNIT_UNICODE_CODE_POINTS`、`TEXT_UNIT_BYTES`（UTF-8 字节）。只有 supported=true 的 `message.send` 可以声明，max 必须大于 0，unit 必须为后三者之一；Core 在 REGISTER 与实例启动结果的信任边界校验。未声明用 absent 表示，不推定平台上限。AdapterDescriptor、实例启动结果、Event 与 SendResult 中的 Capability 保留该字段；SDK 的对应类型见 CONTRACTS §8 与 goal-109。由 T-03.18 追加字段与枚举，字段号不预分配，遵守 §5A 与并行切片的只追加规则。
- Adapter 的公开平台 token 路由声明 `HttpAuthMode` 新增值 `HTTP_AUTH_MODE_PLATFORM_TOKEN`（T-03.19，D-136）；仅 Adapter HTTP 路由可使用。Core 保留 Adapter 包路由的 Authorization 并由 Adapter 校验，插件路由照旧清除；两者均清除外部传入的内部代理凭证、CSRF、Cookie、Forwarded 与 X-Forwarded-*，再注入本 Session 的内部代理凭证。平台凭证不替代内部代理认证。
- 出站媒体沿用 `core.media.read`；Core 在 `adapter.send` 之前为本次发送签发只读 token。
- 这些载荷的 Protobuf 定义放在 `adapter.proto`，编号遵循 §5A 的只追加规则。

### 3.5 插件可调用的 Core 固定服务

以下服务与 `core.send`、`core.media.*` 一样服务名不含 `/`，载荷为 Protobuf，按 §5A 的只追加规则编号：

| 服务 | 请求 → 响应 | 契约 |
| --- | --- | --- |
| `core.edit` | Target、platform_message_id、MessageChain → SendResult 与/或 ErrorInfo | CONTRACTS §8、MESSAGE-MODEL §4 |
| `core.action` | Target 或 adapter_instance_id、动作完整名、版本、JSON 参数 → JSON 结果与/或 ErrorInfo | CONTRACTS §8（平台动作）、D-92 |
| `core.delete` | Target、platform_message_ids（1～100）→ 每个 ID 的状态与安全错误 | CONTRACTS §8、MESSAGE-MODEL §4A |
| `core.continuation.start` | 必须带 parent_call_id → 空结果，超过每 Session 128 个进行中的延续调用时 RESOURCE_EXHAUSTED；原调用成功结束后 Core 发出 Handler REQUEST，HandlerRequest 的 `continuation_of` 为原调用的 call_id，SDK 据此找到登记的延续函数 | CONTRACTS §2.4 |
| `core.wait` | 必须带 parent_call_id（延续调用）、超时 → 命中的 Event | CONTRACTS §2.4 |
| `core.tool.list` | 空 → 工具列表（名称、description、输入 Schema、生效策略） | CONTRACTS §13.1 |
| `core.tool.invoke` | 必须带 parent_call_id；工具名、JSON 参数 → 工具的 JSON 输出与/或 ErrorInfo | CONTRACTS §13.1、§13.2 |
| `core.tool.confirm` | 必须带 parent_call_id；操作名、参数摘要 → 空结果（允许）或 ErrorInfo | CONTRACTS §13.1、§13.2 |
| `core.conversation.append` | 会话键、一条或多条 JSON 条目 → 最后一条的序号 | CONTRACTS §13.3 |
| `core.conversation.list` | 会话键、起始序号 → 条目及序号、下一个序号 | CONTRACTS §13.3 |
| `core.conversation.clear` | 会话键 → 空结果 | CONTRACTS §13.3 |
| `core.command.list` | 必须带同一 Session 内仍在途的 `parent_call_id`，从父调用取可信 Actor 与默认 Target；空的 `CommandListRequest` → `CommandListResponse`（字段见 §5A）；参数用法由 Core 的 usage 回复给出，不在列表中（D-171） | CONTRACTS §2.3、§15.2 |
| `core.task.set` | 任务名、key、触发规则（at_unix_ms、interval_ms 或 cron 加可选的 timezone，恰有一个）、参数 JSON → 下一次触发时刻 | CONTRACTS §4.3、D-124 |
| `core.task.cancel` | 任务名、key → 空结果；实例正在执行时同时取消该次调用 | CONTRACTS §4.3、D-124 |
| `core.task.list` | 任务名、分页游标 → 实例（key、触发规则、参数 JSON、下一次触发时刻）与下一页游标 | CONTRACTS §4.3、D-124 |
| `core.job.pause` | job_id → 空结果 | CONTRACTS §4.3、D-126 |
| `core.job.resume` | job_id → 空结果 | CONTRACTS §4.3、D-126 |
| `core.job.status` | job_id → 是否暂停、上一次与下一次触发时刻 | CONTRACTS §4.3、D-126 |
| `core.profile.list` | 空 → 全部配置文件的 ID、名称与会话数 | CONTRACTS §5.6、D-168 |
| `core.profile.current` | 必须带 parent_call_id → 默认 Target 所在会话的配置文件 ID、名称与会话数 | CONTRACTS §5.6、D-168 |
| `core.profile.assign` | 必须带 parent_call_id，Actor 须为 Core Admin；配置文件 ID（空串取消单独指定）→ 空结果 | CONTRACTS §5.6、D-168 |
| `core.profile.set` | 必须带 parent_call_id，Actor 须为 Core Admin；字段路径、JSON 值（null 删除）→ 配置文件 ID、名称、会话数与 reloading | CONTRACTS §5.6、D-168 |

`core.command.list` 请求与响应均为 PROTOBUF，`schema_id` 分别为 `alonabot.command_list_request.v1`、`alonabot.command_list_response.v1`，`schema_version` 均为 1。缺少有效父调用返回 FAILED_PRECONDITION；通过帧校验后，服务发现请求编码、Schema 或载荷不合法时返回 INVALID_ARGUMENT。路径是生效的主名，主名被遮蔽时省略整条记录；别名只保留该路由仍胜出的生效名称，顺序不变（D-128）。结果按主名 token 用空格连接后的字典序排列；说明按规范路径和 Actor locale 选择，未登记说明时 `description` 不出现。

Core→插件与 Core→Adapter 方向另有 `sdk.log.set_level`：`LogSetLevelRequest`（`LogLevel level = 1`，与 Bootstrap 使用同一枚举）→ 空的 `LogSetLevelResponse`，SDK 收到后立即按新级别过滤（CONTRACTS §14.3）。Core→插件方向另有 `sdk.profile.set_overrides`：该插件当前有效的全部覆盖值（与 Bootstrap 的 `profile_overrides_json` 同形）→ 空结果，SDK 以它整体替换已有的覆盖值（CONTRACTS §5.6，D-168）。

注册描述的追加字段：Service 描述增加可选的 `tool`（description、default_policy、input_schema_json：该 Service 输入 Schema 的原文，Core 只从注册中得到 schema 标识，列出工具时需要 Schema 正文）；CommandMatch 追加 KEYWORD、STARTSWITH、ENDSWITH、FULLMATCH，CommandSpec 增加模式列表与 `ignore_case`；Actor 增加可选 `locale`；message.created 的载荷增加 `to_me`；Bootstrap 追加 `log_level`（字段号 10）与 `session_limits`（字段号 11），字段分配与切片顺序见 §3.2；CommandSpec 增加可选的英文说明 `description`（由 T-03.17 按 D-138 交付，翻译键取规范路径，供菜单与 `core.command.list` 共用）；BeforeSend/AfterSend 的输入增加操作类型与被编辑的 platform_message_id；D-87 的追加：MediaSource 追加 `inline_base64` 与只在 Core→Adapter 方向出现的 `platform_file`（D-88），Segment 追加 voice（D-89），ForwardItem 追加 `platform_message_id`、ForwardSegment 追加 `platform_forward_id`（D-91），Button 追加 `url`（D-97），AdapterDescriptor 追加动作声明（D-92），Event 追加 `chat_type`（新增枚举 ChatType：UNSPECIFIED、PRIVATE、GROUP、CHANNEL，D-100）；D-124 的追加：RouteKind 追加 TASK，RouteDescriptor 追加 TaskSpec（任务名），ScheduledPayload 追加 `task_name`、`task_key` 与 `task_payload_json`，动态任务触发时 `job_id` 为该任务的 route_id（`task.` 加任务名），静态任务触发时这三个字段为空；D-129、D-130 的追加：Service 描述增加可选的选项来源标记（带标记的 Service 输入与输出为 SDK 固定的选项查询 Schema，访问规则只含 Core 与本组件），Event 追加可选的 `chat_title`；D-157 的追加：新增枚举 ToolEffect（UNSPECIFIED、READ_ONLY、WRITE、DESTRUCTIVE），ToolSpec、ToolInfo 与 ToolConfirmRequest 追加 `effect`；D-161 的改动：Bootstrap 删除 `previous_config_version`、`target_config_version`，REGISTER 删除 `candidate_config_version`（字段号保留），REGISTER 追加 `repeated ConfigCorrection config_corrections`（path、reason、original_json）；D-164 的追加：RouteDescriptor 的 HTTP 路由追加 `page_title`（非空表示该路由是插件页面）；D-165 的追加：Bootstrap 追加 `string plugin_data_dir`（插件数据目录的绝对路径）；D-162 的追加：标准事件 `command.invoked` 的载荷消息 CommandInvokedPayload（plugin_id、route_id、command_name、result、duration_ms）。字段号由实现对应切片的 goal 授权分配。这些都是追加，不改变 `contracts/wirespec_test.go` 固定的枚举。

## 4. 调用上下文、读写与取消

- 匹配键为 `(session_id, initiator_side, call_id)`。两个方向可以同时使用 call_id=7；Core 转发 A→B 时建立两段 Session 的内部映射，不要求转发后的 CallID 相同。
- PLUGIN REQUEST 有 parent_call_id 时，Core 必须找到同一 Session 中仍在途的 CORE REQUEST，才能继承该父调用的 actor、target、scope、trace 和调用链；不存在或已经结束则 FAILED_PRECONDITION。无 parent 的调用只具有插件身份，不接受 payload 中的 Admin 标志作为授权。Core 发出的 Request 不靠该字段跨 Session 指认父调用，映射由 Broker 内部维护。
- 可信调用上下文随 Core→Plugin 的 Invocation 载荷下发供 SDK 使用；插件回传的 actor、scope、trace 等值不能覆盖 Core 记录。事件路由的 HandlerRequest 上下文另带 `command_consumed`：本条 message.created 已被命令层消费时为 true，只由 Core 设置（D-141）。调用上下文另带 `profile_id`：本次调用所用的配置文件，`default` 为空串（CONTRACTS §5.6，D-168）。所有子调用采用 `min(parent_deadline,operation_deadline)`；父调用完成/取消时取消尚未结束的子调用。独立后台工作必须使用插件身份和独立 context。
- 同一 Service 调用链的插件 ID 重复返回 CALL_CYCLE，链深超过 16 返回 FAILED_PRECONDITION；独立并发链互不影响。Event Bus 回调没有可继承的用户调用链，后续 Service 使用插件身份。
- 每连接只有一个连续读帧/分流循环、一个串行 writer；Handler 不在读循环内执行。reader 同步处理 Session 状态变化，再将业务提交给执行器；Response/Cancel/控制帧及时分流，不排在正在等待 RPC 的 Handler 后面。
- 每 Session 每方向最多 256 个 outstanding 业务请求；请求/事件载荷另受每 Session 每方向 16 MiB、Core 合计 64 MiB 的在途及排队字节预算约束；任一满额立即 RESOURCE_EXHAUSTED，不等待槽位。Response、Cancel、状态控制走独立、有界且优先的通道，不占用普通业务 admission；无法维持协议时使会话失败，不能无界缓存。优先队列不得使 Cancel 越过自己的 Request：尚未写出的被取消 Request 直接撤出且不再发送，已完整写出后才可随后发送 Cancel。
- 本地超时/取消先完成调用者失败返回，再尽力发送 Cancel；已发送外部消息/SQL 不撤销，不自动重试。Session 失效时清理映射、取消调用并返回 UNAVAILABLE；已确认的 SendResult 仍须保留。
- Core 停用时先冻结该插件的新根调用/新派发，最多 drain 6s；已有父 Request 链中的受限子请求与响应可继续。drain 结束后取消剩余调用并发送唯一 STOPPING，帧 deadline 为 min(t0+10s,发送STOPPING时刻+4s)，OnStop 与资源关闭使用这一独立 context；整个进程树停止预算仍为 10s。Core 同时停止多个插件时共享同一绝对截止时间，不逐个累加预算。

等待休眠插件的业务请求在 Session admission 之前受 CONTRACTS §4.5 的有界唤醒队列约束，保留可信上下文、原始 Target、Trace、顺序与绝对 deadline；确认 ACTIVATED 及快照一致后复核权限再交付，取消或到期的未交付请求移出队列。空闲 STOPPING 与请求准入共用门禁；旧进程树退出前保持唯一 Session 所有权。

## 5. Event Bus、媒体与平台载荷

- Event Bus 的发布通过普通 REQUEST 返回“是否被 Core 接纳”的结果；它不确认订阅者已处理。Core 向订阅者发 EVENT。BusEvent 至少包括 event_name、Core event_id、source_plugin_id、occurred_at、schema_id/version、payload。发布者只能发布自身 `plugin_id/event_name` 命名空间；订阅者 Actor 缺省，业务 payload 不产生管理权限。
- 每订阅者最多 256 个且 4 MiB 待处理事件；任一满额丢弃新事件并计数，插件离线不积压、不补发。业务处理无 ACK、自动重试或 StopPropagation。Scheduler 的 scheduled 通过有 CallID/独立 deadline 的 REQUEST/RESPONSE 判定完成与重叠。
- 服务 JSON 使用 `payload_encoding=JSON` 和注册的 schema/version，严格验证 UTF-8 JSON 及结构；若使用 Protobuf Any，必须是真实的 type_url 和 Protobuf bytes。未知 encoding/schema/version 明确拒绝：Envelope 校验只要求业务载荷的 `schema_version` ≥ 1，与目标服务注册的版本是否一致由 Broker 判断，不一致时对该次调用返回 INVALID_ARGUMENT，不作为协议错误关闭调用方 Session（D-66）。
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
| `command.proto`（T-14.1） | §3.5 的 `CommandListRequest`（空消息）；`CommandInfo`（`string plugin_id = 1`、`string route_id = 2`、`repeated string path = 3`、`repeated CommandAlias aliases = 4`、`optional string description = 5`）；`CommandListResponse`（`repeated CommandInfo commands = 1`）。复用 `route.proto` 的 `CommandAlias`（`repeated string path = 1`） |

规格只给出语义、由 T-01 确定的实现选择：

- 没有编号表的消息，字段号按规格表格中的字段顺序从 1 开始编排。
- 新增枚举 CancelReason、StopReason、SessionPhase、DatabaseBackend（显式 `NONE` 表示没有业务 DB）、MigrationMode、ActivationReason、HandlerMode、CommandMatch（PATH/REGEX/RAW）、UsageMode、ParamType（含 USER、IMAGE、MESSAGE，D-72）、ServiceAccess（PRIVATE/ALLOWLIST/PUBLIC）、HttpAuthMode、TextFormat、EntityType、MediaReadability。
- 为避免 import 循环，Target/Actor/Capability 单独放在 `identity.proto`；Segment 的 `incoming_ref` 直接携带 IncomingMedia 描述符，`media_ref` 是 MediaRef 的 token。
- Event 按类型区分的载荷放在 `EventPayload` oneof 中，PreEventResult 只能回传这一部分；按钮 `value` 与服务载荷之外的 JSON 一律以 UTF-8 文本表示。
- `ForwardSource` 的结构（display_name、external_user_id）由 T-01 补定；新增 `SendRequest{to, chain}` 作为 Core Send 服务的请求。
- Core→目标插件的 Service 调用载荷是 `ServiceInvocation{route_id, context, encoding, schema_id, schema_version, payload}`：可信上下文与业务载荷分开传输，服务响应直接使用 output schema 的业务载荷（REVIEW-1 补充）。
- 控制帧与 Core 固定载荷的 `schema_id` 为 `alonabot.<snake_case 消息名>.v1`（例如 `alonabot.hello_ack.v1`），控制帧一律使用 PROTOBUF 编码。控制帧的 `schema_version` 不为 1 属于协议错误；Core 固定服务的请求与响应由各自的接收方检查版本为 1（D-66）。是否携带载荷以 `schema_id` 是否为空来判断，而不看 payload 长度，因为空消息序列化后是 0 字节。校验入口为 `contracts/validate`。
- Core 启动插件进程时通过环境变量传递会话参数（常量定义在 `internal/wire/env.go`）：`ALONABOT_PLUGIN_ID`、`ALONABOT_SESSION_SECRET`（32 字节十六进制）、`ALONABOT_TRANSPORT`（`tcp`/`stdio`）、`ALONABOT_CORE_ADDR`（TCP loopback 地址）、`ALONABOT_BUILD_DIGEST`（归档 SHA256 十六进制）、`ALONABOT_RUNTIME_LOCK`（runtime.lock 的绝对路径）。其余环境变量只从白名单继承（PATH、HOME、临时目录、SystemRoot、LANG、TZ），Core 管理凭证不会进入插件环境；设置了调优参数 `PLUGIN_MEMORY_SOFT_LIMIT` 时另传 `GOMEMLIMIT`（CONTRACTS §16）；出站代理按 CONTRACTS §15.1 的 `proxy` 设置传递 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。
- 握手与准备：认证失败（Secret 恒时比较、插件 ID、build_digest 任一不符）时 Core 直接关闭连接，不回复任何帧；认证之后的失败（协议协商、Bootstrap 超限、REGISTER 不合法、准备期出现非控制帧）先发 SESSION_ERROR 再关闭，发送 SESSION_ERROR 限时 1s，对端不读时不会卡住失败路径。Preparing 的硬截止为配置、迁移、OnStart 三个阶段预算之和，写入 Bootstrap.preparation_deadline_unix_ms。PREPARED 中 route_descriptor_hash 为按顺序拼接各 RouteDescriptor 确定性编码后的 SHA256，candidate_config_digest 为候选配置 JSON 的 SHA256（T-08 引入 secret 引用后改为对持久化快照计算）。
- 协议能力项使用 `ProtocolCapability{name, version}`，与 Adapter 平台能力 `Capability{name, version, supported, text_limit?}` 分开（可选文本上限由 T-03.18 按 D-135 追加）。
- 2026-11-17 配置文件（D-168，T-04.18 待接入）：`InvocationContext` 追加 `string profile_id = 7`；Bootstrap 追加 `profile_overrides_json`；新增 Core 固定服务 `core.profile.list`、`current`、`assign`、`set` 与 Core→插件的 `sdk.profile.set_overrides`。T-08.18 提供存储与管理，现有协议字段、服务注册与 SDK 保持不变。
- 2026-11-13 命令消费标记（D-141）：`InvocationContext` 追加 `bool command_consumed = 6`。
- 2026-11-01 权限、扩展事件与参数类型（D-69、D-71、D-72）：删除枚举 `TrustedRole` 与 `Permission`、字段 `Actor.trusted_roles` 与 `RuleDescriptor.require_chat_admin`，编号与名称 reserved。`Actor` 新增 `repeated string permissions`；`RouteDescriptor.required_permission` 改为 `repeated string required_permissions`；`AdapterDescriptor` 新增 `repeated PermissionDeclaration permissions`（name、implies、label_key）与 `repeated EventDeclaration events`（name、version、schema_json）；`EventPayload` oneof 新增 `AdapterEventPayload adapter_event`（version、payload_json）；`ParamType` 新增 USER、IMAGE、MESSAGE；`CommandSpec.custom_usage` 改为插件 `i18/` 中的文案键。字段号由 T-04.6 编排。
- 2026-10-27 框架移除 PostgreSQL 后端（D-48）：删除枚举值 `DATABASE_BACKEND_POSTGRES = 3`、`DatabaseDescriptor.postgres = 3` 与消息 `PostgresConnection`，两处编号与名称都已 reserved。

## 6. 实施验证与依据

T-01 将字段表转换为 `.proto`，冻结控制 Payload/RouteDescriptor 的字段号和 fixture；T-02 验证双向编号、乱序响应、失败阶段、帧上限、父调用关联、读循环进度和 ready gate。T-03/T-05/T-06 验证 payload+error、HandlerOutcome、Hook 约束和 Event Bus；对应 [TEST-MATRIX.md](TEST-MATRIX.md)。T-01/T-02 的运行结果见 TEST-MATRIX §3.1；T-03/T-05/T-06 的部分尚未执行。

Protobuf 枚举零值与字段存在性依据：[官方 proto3 指南](https://protobuf.dev/programming-guides/proto3/)。权限和准备期限制是 AlonaBot 的协议选择，不代表对同宿主恶意 Native 代码提供 OS 沙箱。
