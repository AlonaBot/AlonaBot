# AlonaBot v0.1 事件与消息契约

版本：v0.6-review · 2026-10-21。状态：**现行实施基线，待实施验证**。本文件补充 [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) §2/§8 和 [WIRE-FORMAT.md](WIRE-FORMAT.md)，不代表已构建 SDK 或调用平台 API。

## 1. Event、Actor 与 Target

公共 JSON 中的用户、账号、会话、消息、事件和代次标识使用字符串；时间用 UTC Unix 毫秒整数。字段 absent 表示不存在，不用虚构用户、空账号或零 ID 填补。

| 类型 | 字段 | 来源与验证 |
| --- | --- | --- |
| Event | event_id、event_type、source_kind、occurred_at_unix_ms、adapter_id?、adapter_instance_id?、account_id?、target?、actor?、message?、event_payload、capabilities、platform_payload? | event_id 由 Core 在入站标准化时产生；source_kind 为 ADAPTER/SCHEDULER，Core 为实例状态生成的 bot.online/bot.offline 也记为 ADAPTER。平台字段只对相应来源存在；原平台事件标识另存 source_event_id |
| Actor | external_user_id、adapter_instance_id、display_name?、locale?、trusted_roles | 来源必须为可信 Adapter 元数据及 Core policy；用户身份按实例作用域绑定；插件/外部 webhook 的 is_admin 不是身份依据。locale 是平台给出的用户语言（规范化的 BCP 47 标签），平台未提供时 absent，用于 CONTRACTS §5.4 的文案选择 |
| Target | adapter_id、adapter_instance_id、account_id、chat_id、thread_id? | Adapter 验证 account 与实例绑定；默认 Target 仅来自当前实际入站事件；主动任务须显式 To，按 CONTRACTS §8 的 allowed_adapter_instances 与原事件回复范围准入 |
| Capability | name、version、supported | 使用 Adapter 明确公布的能力；未知能力不推定支持 |

命令权限中的 User 指具有可信 Actor 的普通用户；Admin 指 Core 配置的、以 adapter_instance_id/external_user_id 标识的机器人管理员。平台群管理员只形成该会话的 chat_admin Rule，不因此取得 Core 管理、插件安装或其他会话权限。Core 全局禁止规则始终生效，动态策略在每次派发前重新检查。

### 1.1 按事件类型区分的 payload

| event_type | event_payload 必需内容 | 可缺失内容 |
| --- | --- | --- |
| message.created | platform_message_id、chain、to_me | actor 可缺省，例如平台未提供实际个人身份。to_me 为 true 表示消息发给本机器人：私聊、回复本机器人发出的消息或提及本机器人；平台无法判断时为 false |
| message.edited | platform_message_id、当前可见 chain | old_chain 可缺省，不为缺少旧内容伪造原消息 |
| message.deleted | platform_message_ids | 原消息内容和 actor 可缺省；平台没有删除通知就不发此事件 |
| button.clicked | action_id、action_schema_version、value、callback_query_id、origin_platform_message_id? | message 内容可缺省；Actor 必须来自平台本次点击用户，不能取发送按钮时的用户 |
| member.joined / member.left | member_external_user_id | 执行管理动作的 actor 可缺省，不把被操作成员自动当成操作人 |
| reaction.added | platform_message_id、reaction | actor 可缺省；平台未给出个人身份时不猜测 |
| scheduled | job_id、run_id、scheduled_at_unix_ms | adapter/account/target/actor/message 均可 absent；只具有 Core 授予的 job scope |
| bot.online | adapter_id、adapter_instance_id、account_id、account_display_name? | target/actor/message 均 absent；由 Core 在实例启动成功后生成 |
| bot.offline | adapter_id、adapter_instance_id、account_id、reason（STOPPED 停用 / ADAPTER_EXITED Adapter 进程退出 / INSTANCE_FAILED 实例致命错误）、safe_error? | target/actor/message 均 absent；safe_error 仅在 INSTANCE_FAILED 时给出 |

Scheduler 事件必须通过 REQUEST/RESPONSE 调用任务 Handler，使用独立 job deadline；到期只结束调用结果，实际运行槽需 callback 返回或旧进程确认退出才能释放，不能用 Event Bus 的 EVENT 帧推断任务结束。Event Bus 使用独立 BusEvent 载荷，不伪装成可信平台事件。

## 2. MessageChain 与 Segment

MessageChain 是一次**逻辑 Send** 的有序 Segment 列表。Segment JSON 结构为 `type + data`，以下字段构成首发公共模型；平台扩展须有 capability/schema 版本，不靠未知字段猜测。

| type | data 字段 | 约束 |
| --- | --- | --- |
| text | text、format?（plain/markdown）、entities? | 默认为 plain；format 与 entities 不同时指定。markdown 支持范围由 Adapter 公布，不能把未知 HTML 直接转交平台 |
| image | media_ref、incoming_ref 或 https_url、alt?、caption? | 三种来源恰有一个；引用/URL/大小/MIME 在发送前验证 |
| video | media_ref、incoming_ref 或 https_url、duration_ms?、caption? | 来源恰有一个；未知时长为 absent |
| file | media_ref、incoming_ref 或 https_url、filename?、mime?、caption? | filename 仅为无路径展示名称，不作为宿主文件访问路径 |
| mention | external_user_id、display? | 按目标平台实例解释，不能猜测跨实例身份；Adapter 无法表示时 UNSUPPORTED |
| reference | original_platform_message_id、target? | 一条 Chain 最多一个；目标必须与本次发送兼容，否则 UNSUPPORTED |
| buttons | rows[[{action_id,label,value?}]] | 一条 Chain 最多一个；action_id 指向当前插件已注册的 button.clicked 路由，value 通过该 action Schema 校验 |
| forward | items[{source?,chain}] | 仅平台明确支持合并转发时发送；不支持返回 UNSUPPORTED，不默默转成多次 Send |
| extension | adapter_id、name、version、payload（JSON 对象） | 平台专有能力，只能发往该 adapter_id 的实例；name/version 必须在 Adapter 注册的 extension 声明中，payload 按其 Schema 校验；不支持时 UNSUPPORTED，不降级为其他片段（CONTRACTS §8） |

`entities` 使用 `type、offset_utf16、length_utf16` 及类型需要的 url/language/external_user_id；偏移相对于本 text，必须落在完整字符边界，不能截断代理对。首发公共类型为 bold/italic/underline/strikethrough/spoiler/code/pre/text_link/text_mention；不支持的类型明确拒绝。SDK 从 Go 字符串创建实体时须转换 UTF-8/rune 索引，不能直接把字节索引交给 Telegram。

`incoming_ref` 表示尚未取件的入站对象授权，与已校验的 MediaRef 分开；描述包含 token、声明 size/MIME 与 readability。SDK media.Read 按需取件、读取和共享缓存的规则见 CONTRACTS §1.4。入站元数据不泄漏 Bot Token 或任意宿主路径。

Core/Adapter 先验证整个 Chain 的结构、顺序、片段数、媒体租约及目标能力，再发出第一条平台请求。结构错误返回 INVALID_ARGUMENT；合法但平台无法表示返回 UNSUPPORTED；不能先发送已知可用前半段，再因已知不支持的末尾丢失内容。

显式 `Send(empty Chain)` 返回 INVALID_ARGUMENT 和 REJECTED，不调用平台。没有默认 Target 且无显式 To 返回 NO_DEFAULT_TARGET。平台实现必须保留片段顺序；需要拆成多条物理消息时遵循 §4，不静默移动文字或丢弃按钮/附件。只含 reference/buttons 等发送元数据而无可发送内容的 Chain 在预检时拒绝。

## 3. Handler 与传播结果

保留两种开发者 Handler 形式：

- `func(c *Context, ...) (Chain, error)`：自动回复模式；成功返回非空 Chain 后由 Core 自动 Send 一次。返回空 Chain 且无错误代表不回复。该模式内部调用 `c.Send` 返回 FAILED_PRECONDITION。
- `func(c *Context, ...) error`：主动发送模式；Handler 可在内部调用一次或多次 `c.Send`，每次独立形成 SendResult。

两个签名都通过 `c.StopPropagation()` 锁存本次 Handler 的停止标记；默认 Continue，不增加第三套返回签名。SDK 把本次执行结果转为统一的 `HandlerOutcome`：

| 字段 | 语义 |
| --- | --- |
| matched | 自定义 Rule 的结果；false 时未执行 Handler，不产生 Chain |
| propagation | CONTINUE 或 STOP；显式 Stop 一经设置不能重新打开传播 |
| chain? | 仅自动回复模式成功返回的非空 Chain；业务错误时不自动发送 |
| error | 使用 Envelope.ErrorInfo；可与有效 HandlerOutcome 并存以保留传播标记 |

普通事件的 Rule false 跳过当前 Handler，Rule error 记录后继续下一候选；Handler error 默认继续，但显式 Stop 仍停止后续普通 Handler。已识别命令的 Rule false/error、权限拒绝、禁用和参数错误均遵循主契约的“命令已消费”规则，不再选择短路径、同名候选或 Regex/Raw。Handler 失败不回滚内部已确认的 Send。

Context 的传播/默认 Target/用户调用链仅在本次调用期间有效；Handler 返回后后台 goroutine 不能继续改变 HandlerOutcome，不能持过期父 Request 发起用户授权调用。独立后台发送使用插件身份和显式 Target。需要在返回后继续以原上下文工作（等待用户回复、较长的生成）时使用延续调用（CONTRACTS §2.4）。

## 4. Send 与 SendResult

`Send(target,chain)` 是一次逻辑发送，可能需要多个平台 API 操作或返回多个平台 Message。SDK 返回 `(SendResult,error)`；error 非空时仍保留已得到的有效 SendResult。

| SendResult 字段 | 语义 |
| --- | --- |
| status | ACCEPTED / PARTIAL / REJECTED / UNKNOWN |
| target | 实际已校验目标；不得跨实例猜测 |
| platform_message_ids[] | 按物理发送和平台返回顺序保留所有已确认 Message ID；同一消息的 markup 更新不重复添加 ID |
| primary_platform_message_id? | 首个已确认的物理消息 ID；无确认消息时 absent |
| provider_accepted_time? | 平台确实返回接受时间时才设置；不伪造平台时间 |
| capabilities_used[] | 本次实际采用的已声明能力 |
| part_results[] | 各计划物理操作的明细，包含未尝试部分 |

每个 PartResult 包含 `part_index`（从 0 开始）、原 Chain 对应的 segment indexes、`status`、`platform_message_ids[]`、`message_ids_unavailable`（默认 false）、可选安全 ErrorInfo。部分状态为 ACCEPTED / REJECTED / UNKNOWN / NOT_ATTEMPTED；后者用于前序失败后没有调用的平台操作。一项平台相册操作可返回多个 Message ID。平台确认接受但尚未提供可用 ID 时，部分仍为 ACCEPTED，并显式标记 message_ids_unavailable，不填造 ID。

| 实际结果 | 逻辑 status |
| --- | --- |
| 所有计划的物理操作均获平台确认成功 | ACCEPTED |
| 至少一项确认成功，其余明确失败、结果未知或未尝试 | PARTIAL |
| 没有确认成功，且所有已尝试操作均明确失败；或发送前被校验/Hook 拒绝 | REJECTED |
| 没有确认成功，但至少一个已提交平台请求的结果未知 | UNKNOWN |

平台明确拒绝、包括权限/参数/限流失败，也属于 REJECTED，不限于发送前校验。前一项成功、后一项超时属于 PARTIAL，后一项在 part_results 标 UNKNOWN。每次物理请求失败或结果未知后停止后续操作；未尝试部分列出，Core 不自动补发。

部分成功/未知响应同时返回 SendResult 与 ErrorInfo（例如 UNAVAILABLE/DEADLINE_EXCEEDED），不新增临时 wire 错误码，不把 PARTIAL 错报为 FAILED_PRECONDITION。平台连接中断后的结果不确定不能被重新包装为“完全未发送”。如果到达调用者的 RPC 本身断开，调用者仍可能无法获知 Core 已有的回执；v0.1 不增加可靠发信/结果查询协议，也不自动重试。

Edit（CONTRACTS §8）同样返回 SendResult：part_results 恰有一项，status 只取 ACCEPTED / REJECTED / UNKNOWN，成功时 platform_message_ids 与 primary_platform_message_id 为被编辑消息的 ID。Edit 失败不改为发送新消息。

ACCEPTED 只表示平台 API 确认接受，不保证用户阅读。跨多条平台请求无事务性；发送中的媒体租约保护到消费完成/失败/超时，不能因原插件退出就提前删除平台仍在读取的文件。

## 5. Hook 的明确边界

- 每次显式 Send、Edit 或成功的自动回复都执行一次 `BeforeSend → Adapter.Send/Edit → AfterSend`；多次 Send 有多组 Hook，没有发送则没有出站 Hook。BeforeSend 的输入标明操作为 send 或 edit，edit 时附带被编辑的 platform_message_id；修改后的 Chain 仍须满足 Edit 的片段限制。
- PreEvent 可以修改业务 payload 或 StopPropagation；Core 保存的 actor、target、来源、调用链和权限范围不可修改。
- PreEventResult 使用 `event_payload + propagation`，不回传可覆写身份的完整 Event。button action_id/action_schema_version/value、scheduled job_id/run_id 与原平台消息 ID 属于 Core 验证的路由/来源绑定，不接受 Hook 修改，后续匹配按原绑定执行。BeforeSendResult 使用 `chain` 或 ErrorInfo 拒绝；不能修改 Target。修改后重新执行 Chain 结构/能力验证。拒绝产生 REJECTED 和审计，AfterSend 可观察该结果，不虚构平台成功。
- AfterSendRequest 包含最终 SendResult/错误；Handler 只返回成功或 ErrorInfo，无修改 SendResult 的返回字段。其自身失败记日志，不改变已确认的平台发送结果，不触发重发。
- **所有出站 Hook 中禁止 Send**，包含 BeforeSend 与 AfterSend。调用链标记由 Core 维护，沿该调用链间接经 Service 再调用 Send 也不能绕过；Hook 不得创建脱离该调用链的后台发送来规避此契约。Native 自行创建的任意代码副作用仍受其可信代码假设约束。

## 6. 官方 Telegram Adapter 映射

官方 Telegram Adapter 是可安装的 Adapter 包（CONTRACTS §8），一个进程承载全部已配置实例，实例可单独启停；以下映射在 Adapter 进程内实现，游标持久化、按钮 token 与派发仍在 Core。每实例绑定一个 Bot 账号，使用 getMe 得到的身份校验目标账号，凭证不跨实例复用。chat_id 与 thread_id 精确映射到 chat_id/message_thread_id；首发普通私聊、群、频道及论坛话题之外的专属会话形式需显式能力支持，否则 UNSUPPORTED。

### 6.1 接收、游标和实例故障

- 首发使用 long polling；getMe 固定该实例 account_id/username：`username` 为 account_username，`first_name` 为 account_display_name（D-67），同一个 Bot 账号只允许一个 Core 入站 owner。启动发现已有 webhook 时返回 FAILED_PRECONDITION，由管理员处理冲突，不擅自调用 deleteWebhook。
- 每次 getUpdates 显式传 `allowed_updates=["message","edited_message","channel_post","edited_channel_post","callback_query","my_chat_member","chat_member","message_reaction"]`，不依赖服务器旧设置或空列表默认值。成员/Reaction 事件还受 Bot 群管理员等平台权限限制，能力状态按实际权限呈现；遗留的未订阅 update 明确忽略并推进游标，不伪造支持。
- Message 带 sender_chat 时，不用兼容性的 fake from 构造个人 User/Admin；保留发送聊天的原始信息，缺少真实个人身份就让 Actor absent。button Actor 始终取本次 CallbackQuery.from。Actor.locale 取 from.language_code（平台给出时），按 BCP 47 规范化大小写。to_me 在以下情况为 true：chat.type 为 private；reply_to_message.from 是本机器人；文本实体中有指向本机器人用户名的 mention 或指向其用户 ID 的 text_mention。
- Adapter.Start 完成凭证/冲突检查和接收循环初始化后返回；循环在后台运行。401/409 等凭证失效/并发 owner 冲突终止实例接收并上报管理状态，不伪装为正常空闲。短暂读网络错误可按 Core 固定、带上限的退避重连；此接收恢复不允许重试已提交的 Send。
- getUpdates 每个 update 按返回顺序单独接纳：先在 Core 状态库持久化该实例的已接收游标，再交给非持久事件队列；重启从已持久化游标的下一个 offset 拉取，不重放已接纳业务。因此在持久化后、Handler 完成前崩溃，允许该事件丢失。首发交付的是尽力投递，不承诺 durable inbox 或完整补处理。
- Telegram 在至少一周没有新 update 后可能随机选择后续 ID，不能假定 ID 永久递增。Core 记录最后接纳时间与当前接收轮次起点；从两者较晚者起连续 7 天没有接纳新 update 时，在当前 long-poll 结束且无接收请求在途的边界持久化新接收轮次/起点、清除 offset，再开始拉取，避免旧高游标永远屏蔽新序列，也避免空闲时每次轮询都重置。平台更新最长仅保存 24h，该重建不承诺恢复空闲期早已过期的事件。
- 同 batch 逐条推进游标，不得在接纳第一条前就保存全批最大 update_id。达到已声明的队列上限时，该条记录丢弃计数并继续推进游标，避免无界积压或下次拉取重复执行业务；具体队列容量采用主契约的 Core 固定常量。
- 入站每个 Telegram Message 保持一个标准事件，保留 media_group_id 供插件使用；首发不等待媒体组聚合，不因迟到成员丢弃消息。出站相册能力独立于入站是否聚合。

### 6.2 顺序、相册和元数据

- 文本片段按原位置发送，邻接文本可合并但不改变顺序；首发不把独立 text 静默改成图片 caption。caption 只取媒体片段自身声明的 caption。
- 连续 image/video 使用媒体组，按原顺序每组最多 10 项；单项使用对应单媒体方法，2～10 项用 sendMediaGroup。文件使用相应文件方法；不把不兼容类型强行塞进媒体组。文字穿插时结束当前连续媒体组。
- Reference 作用于第一条物理消息/第一组媒体请求，后续部分不重复引用。Buttons 作用于最后一条物理消息；相册方法不能携带 reply_markup 时，计划中增加对已确认末条消息的 markup 更新，将该操作计入 part_results。更新失败为 PARTIAL，保留已发送媒体 IDs。
- Adapter 在发出请求前建立确定的物理操作计划并校验当前平台限制。超长文本、超限 caption、无法表达的格式或片段组合在首发明确拒绝，不自动截断。无跨 API 回滚或静默降级。
- Telegram 的 markdown 能力明确使用 MarkdownV2；实体采用平台要求的 UTF-16 单元偏移。原始输入须满足其格式规则，不能把任意 HTML 或其他 Markdown 方言当作已转换文本。
- Telegram 可能确认接受发送但暂返 message_id=0；此时 PartResult 标明 message_ids_unavailable，0 不进入 platform_message_ids。若计划的后续 markup 操作依赖该 ID，则该步 NOT_ATTEMPTED 并返回明确下游错误，逻辑结果 PARTIAL；保留已接受事实，不补发相册或轮询猜测消息 ID。
- Edit 使用 editMessageText，buttons 片段转为 reply_markup；文本与实体的规则同发送。v0.1 只编辑文本消息，不编辑媒体说明。平台返回内容未变化或消息不可编辑时为 REJECTED。实例启动时声明 `message.edit` 能力。
- mention 仅在平台可按 external_user_id 正确表示时渲染；forward 合并转发无等价能力时返回 UNSUPPORTED。edited/deleted 等事件仅映射平台实际提供的 update；不能轮询猜测来制造平台未提供的删除事件。

### 6.3 按钮与回调

Telegram callback_data 限制为 1～64 bytes，不能拼接完整插件 ID、路由和 JSON value。Core 为每个按钮签发 **32-byte 随机值的无 padding base64url token**（43 ASCII bytes），由 Core 记录映射：plugin_id、route_id/action_schema_version、generation_id、Target、value、expires_at。

- token TTL 固定 **24h**，作为内部首发常量；状态存于 Core 状态库，可跨同一 Generation 的普通进程/Core 重启使用。插件卸载、切换版本/配置 Generation 或 token 过期时永久失效，rollback 也不复活；旧 token 不自动绑定到新版同名路由。
- token 绑定发送 Target，不绑定某一个聊天用户：同聊天的多人可点击，但每次点击都以本次可信 Actor 重新检查当前路由、动态权限和目标实例。token 本身不提供 Admin 权限。
- callback 的 action_id 只能指向发送插件当前注册的 button.clicked 路由，value 必须符合该 action Schema；不能利用别的插件 token 调用任意 Service。
- Adapter 收到可信 CallbackQuery 后先尽力调用 answerCallbackQuery 结束客户端进度，再做 token/目标/权限验证与业务派发；此协议确认不表示业务成功，不发送额外群消息。拒绝/过期仅记录安全诊断；不得回落到 Raw/Regex 或其他命令。
- 业务派发通过普通 REQUEST/RESPONSE，事件携 callback_query_id 便于诊断；actor 来自 CallbackQuery.from。首发不支持缺少可核验普通 Target 的 inline-mode 回调，明确拒绝而不是补造 chat_id。
- 相同 update_id 遵循实例接收游标去重；不同用户点击或同用户再次点击作为独立事件处理。callback_query_id 用于这次平台确认与诊断，此入站接收策略不构成对外 Send 的可靠重试机制。

## 7. 官方 OneBot 11 Adapter 映射

官方 OneBot 11 Adapter 是可安装的 Adapter 包（CONTRACTS §8，D-46），一个进程承载全部实例，每个实例一条正向 WebSocket 连接、绑定一个 QQ 账号；游标、按钮 token 与派发仍在 Core。每次 API 调用带唯一的 `echo`，响应按 `echo` 配对。字段名以 [OneBot 11 标准](https://github.com/botuniverse/onebot-11) 为准，实施时逐项核对原文；各实现端的差异以模拟服务与手测记录为准。

### 7.1 接收与标准化

- 实例启动时建立连接并调用 `get_login_info`：`user_id` 的十进制字符串为 account_id，`nickname` 为 account_display_name；OneBot 11 没有可寻址的账号用户名，account_username absent，命令不接受 `/cmd@name` 形式（D-67）。同账号唯一 owner 按 CONTRACTS §8 由 Core 检查。
- 私聊的用户号与群号是不同的命名空间，chat_id 写作 `private:<user_id>` 与 `group:<group_id>`，不产生 thread_id。数字 ID 一律转为十进制字符串。
- Adapter 一边读连接一边把事件放进实例内的有界队列（固定常量 256 条），按到达顺序经 `core.adapter.deliver` 逐条交付，收到应答后才交付下一条；队列满时丢弃新事件并计数。API 响应与心跳不进这个队列，读循环不因等待交付而停下。

| OneBot 11 上报 | 标准事件 | 规则 |
| --- | --- | --- |
| `message`（private、group） | message.created | platform_message_id 为 message_id；Actor.external_user_id 为 user_id，display_name 取非空的 `sender.card`，否则取 `sender.nickname`，locale absent；群聊中 `sender.role` 为 owner 或 admin 时给出 chat_admin 事实；匿名群消息（`sub_type` 为 anonymous）不构造 Actor |
| `notice.group_recall`、`notice.friend_recall` | message.deleted | platform_message_ids 只含该 message_id；actor 取 operator_id（存在时） |
| `notice.group_increase`、`notice.group_decrease` | member.joined、member.left | member_external_user_id 为 user_id；operator_id 非 0 且与 user_id 不同时作为 actor |
| `meta_event`（lifecycle、heartbeat） | 不交付 | 只用于连接状态 |
| 其他 `notice` 与全部 `request` | 不交付 | 明确忽略并计数，不伪造事件 |

- 入站 Chain：`text` 为 plain 的 text；`at` 为 mention（`qq` 为 all 时没有通用对应）；`reply` 为 reference；`image`、`video` 为带 incoming_ref 的 image、video。没有通用对应的段（`at` all、`face`、`record`、`forward`、`json` 等）不进入 chain，原始消息数组完整保存在 platform_payload 中；因此 chain 可以为空。
- to_me 在以下情况为 true：私聊；chain 中有指向本账号的 mention；`reply` 引用的消息经 `get_msg` 查得发送者是本账号。`get_msg` 失败时按 false 处理并记诊断。
- 入站媒体：`adapter.media.read` 时 Adapter 用段中的 `url` 经 HTTP GET 取件，只允许 http 与 https；段中没有 `url` 时 readability 为 UNSUPPORTED。

### 7.2 能力与发送计划

- 支持的片段：plain 的 text、image、video、mention、reference，以及 extension `onebot11.face` v1（payload 为 `{"id": "<数字字符串>"}`，发送为 `face` 段）。实例声明 `message.send` 与 `message.reference`，不声明 `message.markdown`、`message.album`、`message.buttons` 与 `message.edit`。带 markdown 或 entities 的 text、buttons、forward 与 file 由 Adapter 在首个平台请求前返回 UNSUPPORTED；Edit 由 Core 按未声明的 `message.edit` 返回 UNSUPPORTED。
- 发送用 `send_msg`，`message_type` 与目标号码由 chat_id 决定。连续的 text、mention、image 与 extension 片段合成一条消息的段数组，按原顺序排列；video 单独成一条消息。reference 只加在第一条消息最前面的 `reply` 段上。每条消息是一个物理操作，对应一项 PartResult。
- 媒体来源：https_url 原样作为段的 `file`；media_ref 与 incoming_ref 由 Adapter 经 `core.media.read` 读出后以 `base64://` 传给实现端。
- 结果：响应 `status` 为 ok 时该部分 ACCEPTED，`data.message_id` 为平台消息 ID；为 async 时 ACCEPTED 并标记 message_ids_unavailable；为 failed 时 REJECTED，safe_message 带 retcode。请求帧已写出后连接断开，或到调用期限仍无响应，该部分为 UNKNOWN。连接未建立时在首个请求前返回 UNAVAILABLE。任一部分不是 ACCEPTED 后停止，其余部分 NOT_ATTEMPTED，按 §4 汇总。

## 8. 验收与依据

对应 [TEST-MATRIX.md](TEST-MATRIX.md)：Target/话题/多实例；Handler 两签名及停止传播；Rule false；空 Chain；按钮 token/过期/代次/多用户权限；相册多 ID；发送前拒绝、平台明确失败、先成功再超时；AfterSend 失败不改写平台结果；缺少 actor 和平台不支持事件；消息编辑；bot 上下线事件；Actor.locale 映射；OneBot 11 的接收映射、发送计划与跨 Adapter 一致性（V-71、V-72）。实际编译、平台调用和故障注入均留在实施阶段记录。

平台约束核对依据（2026-10-21）：[Telegram Bot API](https://core.telegram.org/bots/api)、[sendMediaGroup](https://core.telegram.org/bots/api#sendmediagroup)、[InlineKeyboardButton](https://core.telegram.org/bots/api#inlinekeyboardbutton)、[CallbackQuery](https://core.telegram.org/bots/api#callbackquery)。相册返回 Message 数组、每组 2～10 项、按钮载荷限额和回调确认来自官方接口；上面的拆分、token 生命周期和错误汇总属于 AlonaBot 首发选择。
