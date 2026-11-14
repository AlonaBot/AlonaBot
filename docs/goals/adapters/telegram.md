# goal 语句：适配器 · Telegram

本文件收录「适配器 · Telegram」章节的 goal 语句，切片表在[Telegram 实施计划](../../../spec/plan/adapters/telegram.md)。用法、状态与所有 goal 的共同约束见 [README](../README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 已完成（PR #39） |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 已完成 |
| goal-89 | T-03.11 | 官方 Telegram Adapter 的平台权限 | T-04.6 | 已完成（PR #106） |
| goal-111 | T-03.20 | 官方 Telegram Adapter 的媒体与新增片段 | T-03.13、T-03.14、T-04.16 | 等待前置（T-04.16） |
| goal-112 | T-03.21 | 官方 Telegram Adapter 的平台动作与撤回 | T-03.11、T-03.15、T-03.16、T-03.20、T-04.16 | 等待前置（T-03.20、T-04.16） |
| goal-113 | T-03.22 | 官方 Telegram Adapter 的命令菜单 | T-03.17、T-04.16 | 等待前置（T-04.16） |
| goal-117 | T-03.23 | 官方 Telegram Adapter 的代理模式与话题判定 | T-03.4 | 已完成（PR #125） |
| goal-132 | T-03.26 | 官方 Telegram Adapter 填写 chat_title | T-03.4、T-03.25、T-04.16 | 等待前置（T-04.16） |
| goal-137 | T-03.27 | 官方 Telegram Adapter 的服务消息 | T-03.23、T-04.16 | 等待前置（T-04.16） |

## 第 5 波

### goal-5 · T-03.4 官方 Telegram Adapter 的接收

状态：已完成（[PR #39](https://github.com/AlonaBot/AlonaBot/pull/39)，合并提交 `604dc65`）。接收与标准化、`telegramtest.Server`、命令入口已合入；`Actor.locale = 5` 与 `MessageCreatedPayload.to_me = 3` 已追加，现有编号保持不变。模拟、真实私聊接收及未运行项见 TEST-MATRIX §3.2。

```text
/goal 使用 $alonabot-execute 推进 goal-5（切片 T-03.4：官方 Telegram Adapter 的接收）。goal 目录 goal-5，分支 goal-5，worktree .worktrees/goal-5，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现官方 Telegram Adapter 的接收部分：Bot API 客户端、getMe 账号绑定、webhook 冲突检查、后台 long polling、逐条交付与游标、7 天接收轮次重建、401/409 停止实例、媒体组标识、Actor.locale、to_me 与实例代理；提供本地 Bot API 模拟服务，供本切片与后续 Telegram 切片测试。

依据：PLAN T-03 步骤 4；CONTRACTS §8（Telegram 段）、§14.5、§15.1（proxy）、§15.4；MESSAGE-MODEL §1、§1.1、§6.1、§6.2 最后一条；TEST-MATRIX V-45、V-54 的 locale 分项、V-62 的 Bot Token URL 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：identity.proto 的 Actor 追加 optional string locale = 5；message.proto 的 MessageCreatedPayload 追加 bool to_me = 3。随后 go generate ./...（本切片获授权只追加这两个字段）
- contracts/wirespec_test.go：追加这两个字段的编号检查
- 新增 adapters/telegram/（可导入的包 telegram）
- 新增 adapters/telegram/telegramtest/（本地 Bot API 模拟服务）
- 新增 adapters/telegram/cmd/alonabot-adapter-telegram/main.go（只调用 sdk.RunAdapter(telegram.New())；返回错误时与 examples/plugins/echo 一样 slog.Error 后 os.Exit(1)）
- 上述包的测试文件
允许的新跨模块边：adapters/... -> sdk/...、-> contracts（ARCHITECTURE 已登记）。adapters 不 import internal/...，测试也一样。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/telegram
   const AdapterID = "telegram"
   func New() adapter.Adapter // 实现 sdk/adapter 的 Adapter；Descriptor 的 adapter_id 为 AdapterID
   实例配置 JSON：{"token": string（必填）, "api_base_url": string（缺省 "https://api.telegram.org"）, "proxy": string（可选，http/https/socks5 URL）}
2. adapters/telegram/telegramtest
   type Server struct{ ... }
   func NewServer(t testing.TB, token string) *Server // httptest 服务；t.Cleanup 时关闭
   func (s *Server) URL() string                    // 作为 api_base_url
   func (s *Server) SetMe(userID int64, username string)
   func (s *Server) SetWebhook(url string)
   func (s *Server) Push(updates ...json.RawMessage) // 追加待返回的 update（按 update_id 排序保存）
   func (s *Server) FailNext(status int, n int)      // 接下来 n 次 getUpdates 返回该 HTTP 状态与 Bot API 错误体
   func (s *Server) Offsets() []int64                // 依次记录每次 getUpdates 收到的 offset（未带 offset 记为 0）
   func (s *Server) AllowedUpdates() [][]string     // 依次记录每次 getUpdates 的 allowed_updates
   模拟服务按 Bot API 的 JSON 约定实现 getMe、getWebhookInfo、getUpdates（支持 offset、timeout、allowed_updates，timeout 内无数据时返回空数组），token 不符时返回 401。

行为：
- Start（sdk/adapter 的 Adapter.Start）：
  1) 解析实例配置；token 为空、api_base_url 或 proxy 不是合法 URL、proxy 的 scheme 不是 http/https/socks5 → INVALID_ARGUMENT（错误不含配置原文）。
  2) 构建该实例专用的 http.Client：设置了 proxy 时 Transport.Proxy 固定为它，否则用 http.ProxyFromEnvironment（Core 已按 §15.1 设置子进程环境）。代理不可达按网络错误处理，不改为直连。
  3) getMe：account_id 为机器人用户 ID 的十进制字符串，DisplayName 为 username；失败时 Start 返回错误（401 → PERMISSION_DENIED，其余 → UNAVAILABLE）。
  4) getWebhookInfo：url 非空 → FAILED_PRECONDITION，不调用 deleteWebhook。
  5) 启动后台接收循环后返回 Instance：Account() 返回 getMe 的账号；Capabilities 本切片只声明接收相关能力，发送能力由 T-03.7 补充；Send 本切片返回 UNSUPPORTED 的 ErrorInfo（发送由 T-03.7 实现）。Descriptor 的能力名称不能重复，本切片不声明 extension。
- 交付给 Core 的事件：Event 的 adapter_id、adapter_instance_id（StartRequest.InstanceID）与 account_id（Account.ID）必须与 Target 中的同名字段一致，否则 Core 以 INVALID_ARGUMENT 拒绝；event_id、source_kind 与 capabilities 由 Core 重建，Adapter 不填；Actor.trusted_roles 只保留 CHAT_ADMIN。 occurred_at_unix_ms 取 Telegram 的 date（秒）乘以 1000。
- 游标：Adapter 自定义的不透明 JSON 字节 {"offset": int64, "round_start_unix_ms": int64, "last_accepted_unix_ms": int64}；Start 收到空游标时以当前时刻为轮次起点、offset 为 0（不带 offset）。
- 接收循环：
  - 每次 getUpdates 显式传 allowed_updates=["message","edited_message","channel_post","edited_channel_post","callback_query","my_chat_member","chat_member","message_reaction"]、timeout=30、offset（为 0 时不传）。
  - 一批 update 按返回顺序逐条处理：能标准化的经 Core.Deliver 交付，游标 offset=update_id+1、last_accepted 为接纳时刻；收到应答（无论 accepted 是否为 true）后才处理下一条。不支持或本切片不处理的 update 经 Core.SaveCursor 推进游标，不交付。不得在交付第一条前保存整批最大 update_id。
  - 7 天重建：在一次 long poll 结束且没有请求在途时，若 max(round_start, last_accepted) 距今满 7 天（命名常量），先 SaveCursor 保存 {offset: 0, round_start: 现在}，再继续拉取。
  - Deliver 或 SaveCursor 返回错误：实例停止后为 FAILED_PRECONDITION，接收循环退出；其他错误不推进本地游标，退避后从上次保存的 offset 重新拉取（可能重复交付，MESSAGE-MODEL 不承诺 exactly-once）。
  - 401 或 409：经 Core.ReportFailure 报告（safe_message 写明 "telegram rejected the token" 或 "another getUpdates or webhook owns this bot"），停止接收，不再重试。
  - 网络错误、5xx 与 429：接收按有上限的指数退避重试（初始与上限为命名常量；429 使用 retry_after）。只用于接收，与发送无关。
  - Stop：取消在途的 getUpdates 并等待循环退出后返回；Stop 之后不再交付。同一实例的 Start/Stop/重连不会出现两个并行的接收循环。
- 标准化（每个 Telegram Message 一个事件，source_event_id 为 update_id 的十进制字符串）：
  - message、channel_post → message.created；edited_message、edited_channel_post → message.edited（old_chain absent）；my_chat_member、chat_member 中成员从非成员变为成员 → member.joined，反之 → member.left；message_reaction 新增的 reaction → reaction.added；callback_query 由 T-03.8 接入，本切片只推进游标。
  - Target：adapter_id、adapter_instance_id、account_id、chat_id（chat.id 十进制）、thread_id（message_thread_id，存在时）。
  - Actor：取 from；带 sender_chat 的消息不构造 Actor。external_user_id 为 from.id 十进制，display_name 为 username 或姓名，locale 为 from.language_code 按 BCP 47 规范化大小写（例如 "en-us" → "en-US"），平台未给出时 absent。trusted_roles 留空。
  - Chain：text 或 caption 生成一个 text 片段，entities 按 Telegram 的 UTF-16 offset/length 转换为 TextEntity；媒体片段由 T-03.6 接入，本切片不生成媒体片段。media_group_id 写入 PlatformMessage。
  - to_me：chat.type 为 private；或 reply_to_message.from.id 等于本机器人；或 entities 中有指向本机器人 username（不区分大小写）的 mention，或 user.id 等于本机器人的 text_mention。
  - platform_payload 为原始 update JSON。
- 脱敏：Bot API 请求 URL 中的 bot<token> 在任何错误与日志中替换为 bot<redacted>；net/http 返回的 *url.Error 先替换 URL 再记录或返回。token 与 proxy 的用户名、密码不进入日志与错误。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/telegram/telegram_test.go。用 telegramtest.Server 与一个记录调用的 adapter.Core 测试替身驱动 telegram.New() 的实例（不经 Core 进程，属于本包与模拟服务之间的集成测试）：
  - Start 得到 getMe 的 account_id 与 username；已有 webhook 时返回 FAILED_PRECONDITION。
  - 推入 3 条 update（含一条不支持的类型）：依次交付两条、推进游标三次，游标 offset 逐条为 update_id+1；Core 替身对第二条返回 accepted=false 时游标照样推进。
  - 每次 getUpdates 的 allowed_updates 与契约一致；重启实例（以保存的游标 Start）后第一次 getUpdates 的 offset 等于保存值。
  - 注入时钟使轮次满 7 天：先保存不带 offset 的新轮次，下一次 getUpdates 不带 offset。
  - FailNext(401) 与 FailNext(409)：各产生一次 ReportFailure，接收停止。FailNext(502, 2)：退避后恢复，无重复交付。
  - 私聊、回复本机器人、@本机器人、text_mention 的 to_me 为 true，群里的普通消息为 false；sender_chat 消息没有 Actor；language_code 规范化；media_group_id 保留。
  - 错误与日志中不出现 token（用测试 token 字符串做断言）。
  - 设置 proxy 为一个记录请求的本地 HTTP 代理时，全部 Bot API 请求经过它。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：配置的各种非法值；socks5 代理 URL 的解析；退避上限；Stop 期间在途请求被取消；成员事件的状态变化表；UTF-16 偏移在含 emoji 文本中的转换。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./adapters/... ./contracts/...；go build ./...。

不做：发送、发送计划与回执（T-03.7）；入站媒体与 IncomingMedia（T-03.6）；按钮与 callback_query（T-03.8）；编辑（T-03.9）；Core 侧装配与真实 Telegram 联验（T-03.10）。
```

## 第 6 波

### goal-50 · T-03.7 官方 Telegram Adapter 的发送计划与回执

状态：已完成（[PR #108](https://github.com/AlonaBot/AlonaBot/pull/108)，合并提交 `d89e9b7`），证据见 TEST-MATRIX §3.2；`Button.core_token = 4` 已追加，Core 签发与回调接入归 T-03.8。派发后 2026-11-09 维护者在 PR #108 确认：最后一条物理消息是相册时，按钮在首个平台请求前返回 UNSUPPORTED（见 DECISIONS），语句的按钮、物理计划、回执与本地自测段已按此修订。前置 T-03.4（PR #39，合并提交 `604dc65`）已合并；语句原先等待的 `Core.ReadMedia` 由 T-03.6（PR #55，合并提交 `b023130`）合入。2026-11-08 按合入代码核对并改写：`sdk/adapter.Core.ReadMedia(ctx, token, offset uint64, length uint32) ([]byte, uint64, bool, error)`；`internal/core/message` 的 Send 在调用实例前已把 media_ref 与 incoming_ref 都换成该 Adapter Session 的只读令牌并写回 media_ref，https_url 原样交给 Adapter，并已用 `validate.Chain` 做过结构校验（reference、buttons 各至多一个等）；`botClient.call` 只发 JSON，`apiError{status, description, retryAfter}` 已脱敏；实例的 `Capabilities()` 与 `Descriptor()` 共用 `receiveCapabilities()`，`media.read` 由 SDK 按实例是否实现 `adapter.MediaSource` 自动声明；`telegramtest.Server` 现有 `FailNext` 只作用于 getUpdates。`Segment` 还没有 extension 分支（随 T-03.8 加入），本切片不处理。goal-89 也修改 `telegram.go` 与 `telegramtest/server.go`，并在 telegramtest 增加 `Calls(method string) int`；本切片的请求记录用 `Requests` 命名以免冲突，后合并的一方按先合并的代码变基。

```text
/goal 使用 $alonabot-execute 推进 goal-50（切片 T-03.7：官方 Telegram Adapter 的发送计划与回执）。goal 目录 goal-50，分支 goal-50，worktree .worktrees/goal-50，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 实例的 Send 先对整条 Chain 做完整预检并生成确定的物理操作计划，再依次调用 Bot API；文本实体按 UTF-16 渲染，连续图片/视频组成媒体组，引用作用于第一条物理消息，按钮作用于最后一条；每个物理操作一个 PartResult，汇总为 ACCEPTED、PARTIAL、REJECTED 或 UNKNOWN，已发生的结果不丢失。

依据：PLAN T-03 步骤 6（发送）；CONTRACTS §1.3（媒体分块）、§8（Send、Chain 与计划段）；MESSAGE-MODEL §2、§4、§6.2；TEST-MATRIX V-08、V-35；Telegram Bot API 的 sendMessage、sendPhoto、sendVideo、sendDocument、sendMediaGroup 与 MessageEntity（字段名与限制以官方文档为准）。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto：Button 追加 optional string core_token = 4（由 Core 填写，T-03.8 接入；插件填写的值由 Core 清除）。随后 go generate ./...（本切片获授权只追加这一字段）
- contracts/wirespec_test.go：追加该字段的编号检查
- adapters/telegram/：新增 plan.go（预检与物理计划）、send.go（执行与回执）、render.go（文本、实体与按钮渲染）；client.go 增加 multipart 请求（复用 call 的响应解析与脱敏）；telegram.go 的实例 Send、Capabilities 与 Descriptor 的能力表
- adapters/telegram/telegramtest/：新增 send.go（发送类方法、请求记录与故障注入）；server.go 的方法分派只加这些方法
- adapters/telegram/send_test.go；其他 *_local_test.go（不提交）
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/telegram：实例实现 sdk/adapter 的 Instance.Send；不新增导出符号。
2. adapters/telegram/telegramtest（在现有 Server 上追加）
   type Request struct { Method string; Params map[string]any; Files map[string][]byte } // 一次发送类调用；Params 为 JSON 或 multipart 的非文件字段（multipart 中的 JSON 字段如 media、reply_markup 解析为对象），Files 为 multipart 文件字段名到内容
   func (s *Server) Requests() []Request // 按到达顺序返回发送类方法（上述六个）的请求副本；不含 getMe、getWebhookInfo、getUpdates、getFile
   func (s *Server) FailMethod(method string, nth int, status int, description string) // 该方法的第 nth 次调用（从 1 起）返回指定 HTTP 状态与 Bot API 错误体
   func (s *Server) ZeroMessageID(method string, nth int)                              // 第 nth 次调用成功，但返回的 message_id 为 0（sendMediaGroup 为该组全部消息）
   func (s *Server) Drop(method string, nth int)                                       // 第 nth 次调用在写出响应前断开连接（结果未知）
   成功的发送类调用返回 Message 对象（sendMediaGroup 返回数组），message_id 在一个 Server 内从 1 递增；editMessageReplyMarkup 返回 Message。

行为：
- 预检（任何 Bot API 调用与媒体读取之前，任一失败都返回 status 为 REJECTED、part_results 为空的 SendResult 与 ErrorInfo）：
  - 片段类型只支持 text、image、video、file、mention、reference、buttons；forward → UNSUPPORTED。
  - text：format 为 MARKDOWN 时按 MarkdownV2（parse_mode）发送，且不得同时带 entities；PLAIN 或缺省时 entities 按 UTF-16 偏移检查（越界、长度为 0、TEXT_LINK 缺 url、TEXT_MENTION 缺 external_user_id → INVALID_ARGUMENT），按 Bot API 的 MessageEntity 输出。合并后的单条文本超过 4096、caption 超过 1024 个 UTF-16 码元（命名常量）→ INVALID_ARGUMENT，不截断。
  - 媒体来源：media_ref 是 Core 换成的只读令牌，执行到该步时经 Core.ReadMedia 分块读取（每块至多 1 MiB，命名常量注明 CONTRACTS §1.3）后以 multipart 上传；https_url 只接受 https 方案，原样作为 photo/video/document 参数；incoming_ref 不应到达 Adapter → INVALID_ARGUMENT。
  - reference：original_platform_message_id 与 target.chat_id 必须是十进制整数，reference.target 若存在须与本次 target 的 chat 一致（否则 UNSUPPORTED，Telegram 本切片不做跨会话回复）。
  - buttons：每个按钮必须有 core_token（1～64 字节，命名常量），否则 INVALID_ARGUMENT；按钮作用于最后一条物理消息，最后一条物理消息是相册（sendMediaGroup）时 → UNSUPPORTED（媒体组不能携带或事后加上内联按钮）。
  - mention：渲染为该位置的 text_mention 实体（显示 display，缺省为 external_user_id），并入相邻文本；external_user_id 不是十进制整数 → UNSUPPORTED。
  - target：chat_id 必须是十进制整数；thread_id 存在时须为十进制整数，作为 message_thread_id。
- 物理计划（MESSAGE-MODEL §6.2）：相邻 text 与 mention 合并为一条 sendMessage，不改变顺序；独立 text 不改为 caption，caption 只取媒体片段自身的 caption；连续 image/video 每组 2～10 项用 sendMediaGroup，单项用 sendPhoto/sendVideo；file 用 sendDocument，不放进媒体组；文本或文件穿插时结束当前媒体组。reference 转为第一条物理请求的 reply_parameters；buttons 转为最后一条物理消息的 reply_markup（inline_keyboard，callback_data 为 core_token）；计划中没有 markup 更新步骤。
- 执行与回执：按计划顺序执行；每个物理操作一个 PartResult（part_index 从 0 起，segment_indexes 为它覆盖的原片段下标，reference 与 buttons 计入它们作用的那一步）。成功时记录 message_id；message_id 为 0 时 message_ids_unavailable=true，0 不进入 platform_message_ids。Bot API 返回错误体（apiError）时该步 REJECTED；连接断开、超时或响应无法解析时该步 UNKNOWN；ReadMedia 失败时该步 REJECTED（ErrorInfo 取 Core 返回的错误），不发起该步请求。任一步非 ACCEPTED 后，后续步骤全部 NOT_ATTEMPTED。整体：全部 ACCEPTED → ACCEPTED（不带 ErrorInfo）；第一步 REJECTED → REJECTED；第一步 UNKNOWN → UNKNOWN；其余 → PARTIAL；非 ACCEPTED 时同时返回该步的 ErrorInfo。platform_message_ids 按物理顺序保留全部已确认 ID；primary_platform_message_id 为第一条确认的消息 ID；capabilities_used 列出本次用到的已声明能力。
- 错误码：apiError 的 400 → INVALID_ARGUMENT，401 → UNAUTHENTICATED，403 → PERMISSION_DENIED，429 → RESOURCE_EXHAUSTED（safe_message 带 retry_after 秒数），其余 → UNAVAILABLE；safe_message 用已脱敏的描述。发送不自动重试，429 也不等待重发。
- Capabilities：在接收能力之外声明 message.send、message.markdown、message.album、message.reference、message.buttons（version 1，supported=true）；实例的 Capabilities 与 Descriptor 共用同一张表（media.read 仍由 SDK 自动声明，不写进表）。
- 脱敏沿用 goal-5：任何错误与日志不含 token；multipart 请求的错误同样经 safeError。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/telegram/send_test.go 的 TestSendPlan。用 telegramtest.Server 与包内的 Core 测试替身（ReadMedia 按令牌返回内存字节，分块读取）驱动真实实例：
  - text + image + image + text + file + buttons（另带一个 reference）：Requests() 依次为 sendMessage（带 reply_parameters）、sendMediaGroup（2 项，multipart 文件内容与替身一致）、sendMessage、sendDocument（带 reply_markup，callback_data 为 core_token）；整体 ACCEPTED，platform_message_ids 与 part_results 的 segment_indexes 正确。
  - FailMethod 使第二步返回 400：第一步 ACCEPTED、第二步 REJECTED、其余 NOT_ATTEMPTED，整体 PARTIAL 并带 INVALID_ARGUMENT。
  - Drop 使第一步断开：整体 UNKNOWN，其余 NOT_ATTEMPTED。
  - 预检失败（超长文本、forward、缺 core_token 的按钮、markdown 与 entities 同时出现）：整体 REJECTED，Requests() 为空，ReadMedia 未被调用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：媒体组 10 项与 11 项的拆分；文本与 mention 的合并规则；caption 与文本长度边界（以 UTF-16 计）；含 emoji 与组合字符的文本，entities 偏移在请求中正确；429 的错误码与 safe_message；thread_id；https_url 直传与非 https 拒绝；末尾为相册且带 buttons：整体 REJECTED（UNSUPPORTED），Requests() 为空，ReadMedia 未被调用；按钮跟在相册后的文本或单张媒体上时照常发送；ZeroMessageID 作用于相册：message_ids_unavailable=true、platform_message_ids 不含 0；ReadMedia 失败；错误与日志不含 token。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./adapters/... ./contracts/...；go build ./...。

不做：按钮 token 的签发与 callback_query（T-03.8）；extension 片段（T-03.8）；编辑（T-03.9）；平台权限（T-03.11）；BeforeSend/AfterSend Hook（T-05.3）；故障矩阵与真实 Telegram 验证（T-03.10）。
```

### goal-117 · T-03.23 官方 Telegram Adapter 的代理模式与话题判定

状态：已完成（[PR #125](https://github.com/AlonaBot/AlonaBot/pull/125)，合并提交 `f3af986`），证据见 TEST-MATRIX §3.2。`adapter_config_schema.json` 中的 `proxy_mode` 由 goal-101 补上。派发时：前置 T-03.4（[PR #39](https://github.com/AlonaBot/AlonaBot/pull/39)，合并提交 `604dc65`）已合并。2026-11-09 按 D-100、D-101 新增，接口与测试段按 main（`3f7a380`）的代码写出：`newBotClient(token, base, proxy string)` 在 proxy 为空时使用 `http.ProxyFromEnvironment`；`instance.messageEvent` 只要 Message 带 `message_thread_id` 就写入 Target 的 thread_id；`TestInstanceProxy` 与 `TestIncomingMedia` 只配置了 `proxy`。实例配置的 `adapter_config_schema.json` 由 goal-101（T-08.12）新建，处理方式见执行顺序 §3。

```text
/goal 使用 $alonabot-execute 推进 goal-117（切片 T-03.23：官方 Telegram Adapter 的代理模式与话题判定）。goal 目录 goal-117，分支 goal-117，worktree .worktrees/goal-117，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 实例配置追加 proxy_mode，代理可以不使用（缺省）、跟随 Core 或使用实例自己的 proxy；入站 Message 只在 is_topic_message 为 true 时把 message_thread_id 作为 Target 的 thread_id，非论坛超级群的回复串不再被当作话题。

依据：PLAN T-03 步骤 20；CONTRACTS §8（Telegram 段）、§15.1（proxy）；MESSAGE-MODEL §6、§6.1；DECISIONS D-100、D-101；TEST-MATRIX V-99、V-100；Bot API 原文 Message 的 message_thread_id 与 is_topic_message。

范围（只新增或修改这些）：
- adapters/telegram/telegram.go：实例配置的解析
- adapters/telegram/client.go：HTTP 客户端的代理选择
- adapters/telegram/inbound.go：telegramMessage 追加 is_topic_message，thread_id 的判定
- adapters/telegram/telegram_test.go、adapters/telegram/media_test.go：既有用例的配置与本语句的测试
- adapters/telegram/cmd/alonabot-adapter-telegram/adapter_config_schema.json：只在派发时 main 已有该文件（goal-101 已合并）时修改；没有时不新建，在 PR 正文写明，由后合并的 goal-101 按执行顺序 §3 补上
没有新的跨模块边。不改 proto，不改 telegramtest。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 实例配置 JSON：{"token": string（必填）, "api_base_url": string（缺省 "https://api.telegram.org"）, "proxy_mode": "none" | "core" | "custom"（缺省 "none"）, "proxy": string（只在 custom 时使用，http/https/socks5 URL，可含用户名与密码）}
2. adapter_config_schema.json（文件存在时）：在 proxy 之前追加
   "proxy_mode": {"type": "string", "options": ["none", "core", "custom"], "labels": ["No proxy", "Follow Core", "Custom"], "default": "none"}
   proxy 保留 secret，追加 "condition": {"proxy_mode": "custom"}；`$version` 不变（首次发布前不写迁移）。

行为：
- Start 解析配置：
  1) proxy_mode 缺失或为空串时按 "none"；不是三者之一 → INVALID_ARGUMENT，错误只写字段名 proxy_mode。
  2) none：Transport.Proxy 为 nil，直连，不读取 HTTP_PROXY、HTTPS_PROXY、NO_PROXY；proxy 有值也不使用、不校验。
  3) core：Transport.Proxy 为 http.ProxyFromEnvironment（Core 按 CONTRACTS §15.1 为子进程设置的代理环境）；proxy 不使用、不校验。
  4) custom：proxy 为空 → INVALID_ARGUMENT（错误只写字段名 proxy）；proxy 的校验、Transport.Proxy 固定为它、用户名与密码加入脱敏表的现有规则不变。
  5) 三种模式都作用于该实例的全部请求（Bot API、getFile 与文件下载共用同一个 http.Client）。代理不可达按网络错误处理，接收退避重连，不改为直连。
- 话题判定（inbound.go）：telegramMessage 追加字段 IsTopicMessage bool（JSON 键 is_topic_message）；messageEvent 只在 IsTopicMessage 为 true 且 message_thread_id 存在时设置 Target.thread_id。General 话题与非论坛超级群的回复串都不产生 thread_id。出站规则不变：Target 给出 thread_id 时照常作为 message_thread_id。
- 其余接收、发送、权限与脱敏行为不变。

测试（按编码规范「测试提交规则」）：
- 提交：修改既有用例并追加一个用例，都经 telegramtest.Server 与 recordingCore 驱动 New() 的实例：
  - TestInstanceProxy 与 TestIncomingMedia 的配置追加 "proxy_mode": "custom"，断言不变（全部请求经实例的 proxy）。
  - 新增 TestDirectProxyModeIgnoresEnvironment：t.Setenv("HTTP_PROXY", 记录请求的本地代理的 URL)，配置 proxy_mode 为 "none"、api_base_url 为 "http://telegram.invalid"、proxy 也指向该代理；Start 返回 UNAVAILABLE，代理没有收到任何请求。
  - TestMessageMetadata：带 message_thread_id 77 的用例改为 chat.type "supergroup" 并带 "is_topic_message": true，断言 thread_id 仍为 "77"；追加一条 chat.type "supergroup"、只带 "message_thread_id": 88（回复串，没有 is_topic_message）的消息，断言其 Target 没有 thread_id。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：proxy_mode 的非法值、custom 缺少或给出非法 proxy、none 与 core 时 proxy 为非法值仍能启动；core 模式在单独的进程中先设置 HTTP_PROXY 再启动实例，请求经过该代理（net/http 在进程内只读取一次代理环境变量，测试中其他用例的 httputil.ReverseProxy 会先触发读取，因此不放进提交的测试）；General 话题（没有 message_thread_id）的消息没有 thread_id。
- 真实平台（可选）：维护者提供论坛超级群与普通超级群时，各回复一次话题内消息与回复串消息，确认前者回到原话题、后者发到会话；没有条件时在 PR 中写「未运行」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./adapters/...；go build ./...。

不做：入站事件的 chat_type（T-03.20，依赖 T-03.14 的 proto 字段）；file_base_url（T-03.20）；Core 全局代理的设置与子进程环境（cmd/alonabot 已实现，不改）；OneBot 11 的代理。
```

## 第 7 波

### goal-137 · T-03.27 官方 Telegram Adapter 的服务消息

状态：等待前置（T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：待派发。前置 T-03.23（[PR #125](https://github.com/AlonaBot/AlonaBot/pull/125)，合并提交 `f3af986`）已合并。2026-11-11 按 D-133 新增，接口与测试段按 main（`db6317a`）的代码写出：`instance.normalize` 把 message、channel_post 与两种 edited 都交给 `messageEvent`，带服务字段的 Message 因此成为 chain 为空的 message.created；`Descriptor` 只声明权限，没有 `Events`；Core 的 `internal/core/adapter/remote` 已按声明的 Schema 校验扩展事件。`inbound.go` 与 goal-111、goal-132 共用，处理方式见执行顺序 §3。

```text
/goal 使用 $alonabot-execute 推进 goal-137（切片 T-03.27：官方 Telegram Adapter 的服务消息）。goal 目录 goal-137，分支 goal-137，worktree .worktrees/goal-137，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 的服务消息不再以 chain 为空的 message.created 交付：话题的创建、编辑、关闭与重开，置顶，修改群名，普通群升级为超级群映射为扩展事件；成员类与其他服务消息不交付，只推进游标。

依据：PLAN T-03 步骤 24；MESSAGE-MODEL §1.2、§6.1、§6.6；CONTRACTS §8（扩展事件的声明）；DECISIONS D-133；TEST-MATRIX V-119；Bot API 原文的 Message（描述为 service message 的字段）、ForumTopicCreated、ForumTopicEdited。

范围（只新增或修改这些）：
- adapters/telegram/inbound.go：服务字段的解析与映射
- adapters/telegram/telegram.go：Descriptor 追加 Events
- adapters/telegram/telegram_test.go：本语句的测试
- ARCHITECTURE.md：Telegram 行的实现说明追加一句
没有新的跨模块边。不改 proto，不改 telegramtest，不新增公开包。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. Descriptor().Events 声明 7 个事件，Version 都为 1，SchemaJson 为 type object、additionalProperties false 的 JSON Schema：
   - telegram.forum_topic_created：{"name": string, "icon_color": integer, "icon_custom_emoji_id": string}，required ["name", "icon_color"]
   - telegram.forum_topic_edited：{"name": string, "icon_custom_emoji_id": string}，都不 required
   - telegram.forum_topic_closed、telegram.forum_topic_reopened：没有属性
   - telegram.message_pinned：{"message_id": string}，required
   - telegram.chat_title_changed：{"new_chat_title": string}，required
   - telegram.chat_migrated：{"migrate_from_chat_id": string}，required
2. 事件：EventType 为上面的完整名，EventPayload 为 adapter_event{version: 1, payload_json: 对象}；payload 字段名照抄 Bot API，数字 ID 写作十进制字符串，平台没有给出的可选字段不写。

行为：
- 判定：message 与 channel_post 带任一服务字段时，在 messageEvent 之前分流，不再检查 text、caption 与媒体，也不产生 message.created。edited_message 与 edited_channel_post 不变。
- 映射：forum_topic_created、forum_topic_edited、forum_topic_closed、forum_topic_reopened、pinned_message（取其 message_id）、new_chat_title、migrate_from_chat_id 各产生一个对应事件。事件经 baseEvent 构造：Target 为该 Message 的 chat，thread_id 沿用 is_topic_message 规则；Actor 为 from，带 sender_chat 时 absent；platform_payload 为原 update；群中的平台权限照常经 addPlatformPermissions 添加。
- 不交付（只推进游标，不记 WARN）：new_chat_members、left_chat_member、migrate_to_chat_id，以及 Bot API 原文 Message 中其余描述为 service message 的字段（按原文逐项列出，例如 group_chat_created、supergroup_chat_created、channel_chat_created、new_chat_photo、delete_chat_photo、message_auto_delete_timer_changed、video_chat_scheduled、video_chat_started、video_chat_ended、video_chat_participants_invited、general_forum_topic_hidden、general_forum_topic_unhidden、write_access_allowed）。列表写在 inbound.go 的一处，注释写明来源与核对日期。
- 一条 Message 同时带两个服务字段不会出现；只按第一个命中的映射字段处理，不另做校验。
- 没有服务字段、也没有可映射内容的 Message（投票、位置等）不变，仍为 chain 为空的 message.created。
- 其余接收、游标、权限与脱敏行为不变。

测试（按编码规范「测试提交规则」）：
- 提交：在 TestNonMessageEvents 的 update 列表中追加用例（经 telegramtest.Server 与 recordingCore 驱动 New() 的实例），不新增测试函数：
  - 超级群中带 is_topic_message 与 message_thread_id 5 的 forum_topic_created（name "Ideas"、icon_color 7322096）：EventType 为 telegram.forum_topic_created，Target 的 thread_id 为 "5"，payload_json 为 {"icon_color":7322096,"name":"Ideas"}（按 JSON 比较），Actor 为 from。
  - pinned_message（被置顶消息 message_id 10）：telegram.message_pinned，payload {"message_id":"10"}。
  - 新超级群中的 migrate_from_chat_id -123：telegram.chat_migrated，Target 的 chat_id 为新超级群，payload {"migrate_from_chat_id":"-123"}。
  - 旧群中的 migrate_to_chat_id、new_chat_members 与 group_chat_created 各一条：不产生事件，游标仍推进（waitSaved 次数与 update 数一致，core.events 中没有它们）。
  - 断言 Descriptor().Events 的 7 个名称与版本。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：forum_topic_edited（只有 name、只有空串 icon_custom_emoji_id）、forum_topic_closed 与 reopened、new_chat_title、频道中的 pinned_message（Actor absent）、left_chat_member；7 个 SchemaJson 能被 jsonschema-go 解析，且各事件的 payload 通过对应 Schema 的校验；投票消息仍为 chain 为空的 message.created。
- 真实平台（可选）：维护者提供测试 Bot 与论坛超级群时，创建并关闭一个话题、置顶一条消息、修改群名，确认各产生一个对应事件且没有 message.created；没有条件时在 PR 中写「未运行」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./adapters/... ./internal/core/adapter/...；go build ./...。

不做：标准事件类型中的服务事件；事件名的常量包；OneBot 11 的对应处理；chat_type（goal-111）与 chat_title（goal-132）。
```

## 第 8 波

### goal-89 · T-03.11 官方 Telegram Adapter 的平台权限

状态：已完成（[PR #106](https://github.com/AlonaBot/AlonaBot/pull/106)，合并提交 `3913d5a`），证据见 TEST-MATRIX §3.2。派发时：前置 T-04.6（PR #78，合并提交 `0a320c5`）已合并。依据 D-69 新增，2026-11-08 按合入代码写出接口、行为与测试段。已核对：`adapters/telegram` 的 `Descriptor()`（目前只声明接收能力）、`instance.processUpdate` → `normalize` → `Core.Deliver` 的逐条交付、`baseEvent`/`actor` 构造 Actor 的位置、`botClient.call(ctx, method, request, result)`、`telegramtest.Server` 的方法分派；`pb.PermissionDeclaration{name, implies, label_key}` 与 `validate.AdapterRegister` 对权限声明的检查（名称带 `telegram.` 前缀、label_key 非空、隐含不成环）；`Actor.permissions` 由 Core 按声明过滤并展开。

```text
/goal 使用 $alonabot-execute 推进 goal-89（切片 T-03.11：官方 Telegram Adapter 的平台权限）。goal 目录 goal-89，分支 goal-89，worktree .worktrees/goal-89，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：官方 Telegram Adapter 在描述中声明 telegram.chat_owner 与 telegram.chat_admin；群与超级群中带 Actor 的事件按 getChatAdministrators 的会话缓存给出 Actor 的平台权限，缓存 10 分钟，该会话的 chat_member、my_chat_member update 使缓存作废，获取失败时本事件不带平台权限并记诊断；发布只含常量的 adapters/telegram/perm。

依据：PLAN T-03 步骤 9（Telegram 实施计划）；MESSAGE-MODEL §6.1；CONTRACTS §2.3、§8（Adapter 权限声明）；DECISIONS D-69；TEST-MATRIX V-85 的 Telegram 分项；Telegram Bot API 的 getChatAdministrators 与 ChatMember（字段名以官方文档为准）。

范围（只新增或修改这些）：
- adapters/telegram/telegram.go：Descriptor 增加权限声明；processUpdate 在交付前补上平台权限
- 新增 adapters/telegram/permissions.go：会话缓存与 getChatAdministrators 调用
- 新增 adapters/telegram/perm/（只含常量）
- adapters/telegram/telegramtest/server.go：增加 getChatAdministrators 的响应与调用计数
- adapters/telegram/telegram_test.go：增加一个测试；其他 *_local_test.go（不提交）
没有新的跨模块边；perm 不 import 任何包。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/telegram/perm
   const ChatOwner = "telegram.chat_owner"
   const ChatAdmin = "telegram.chat_admin"
   包文档说明插件以 sdk.PermissionName(perm.ChatAdmin) 等方式引用，本包不链接 Adapter 运行时代码。
2. adapters/telegram
   Descriptor() 在现有能力之外依次声明 permissions：
     {name: perm.ChatOwner, implies: [perm.ChatAdmin], label_key: "permission.chat_owner"}
     {name: perm.ChatAdmin, implies: ["chat_admin"], label_key: "permission.chat_admin"}
   缓存有效期为命名常量 10 分钟，经实例已有的 now 计时。
3. adapters/telegram/telegramtest
   func (s *Server) SetChatAdministrators(chatID int64, status int, members ...json.RawMessage) // status 为 200 时返回 members（ChatMember 对象）；其他值返回该 HTTP 状态与 Bot API 错误体；未设置的 chat 返回 400
   func (s *Server) Calls(method string) int // 某个 Bot API 方法累计被调用的次数

行为：
- 适用范围：chat.type 为 group 或 supergroup、且事件有 Actor 的事件（message.created、message.edited、member.joined/left 的操作人、reaction.added）。私聊、频道、没有 Actor 的事件（含 sender_chat 发送的消息）不带平台权限，也不调用 getChatAdministrators。
- 取值：在 processUpdate 中、normalize 之后、Deliver 之前，按 chat_id 取会话缓存的管理员列表；缓存缺失或超过 10 分钟时调用 getChatAdministrators（以接收循环的 ctx 为期限）并写入缓存。Actor.external_user_id 对应的成员 status 为 creator 时 Actor.permissions 为 [perm.ChatOwner]，为 administrator 时为 [perm.ChatAdmin]，不在列表中时为空。只给平台事实，隐含关系由 Core 展开。
- 作废：update 为 chat_member 或 my_chat_member 时，先作废该 chat 的缓存，再为本 update 的事件取值（因此成员变更事件使用变更后的列表）。
- 失败：getChatAdministrators 失败（网络错误、非 200、无法解析）时，本次 update 的事件照常交付但不带平台权限，记 WARN（adapter_instance_id、chat_id 与经 safeError 脱敏的错误，不含 token）；失败结果不写入缓存，下一个事件重新获取。401 等凭证失效不在此处改变实例状态，由接收循环原有的处理负责。
- 不改变游标推进、交付顺序与现有事件字段；一次 update 产生多个事件（reaction）时只获取一次。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/telegram/telegram_test.go 的 TestPlatformPermissions。用 telegramtest.Server 与现有的 Core 测试替身驱动实例：超级群中 creator、administrator 与普通成员各发一条消息，依次交付的 Actor.permissions 为 [telegram.chat_owner]、[telegram.chat_admin]、[]；三条消息只调用一次 getChatAdministrators；私聊消息不带平台权限且不增加调用次数；推入该群的 chat_member update 后，再发消息会重新获取；getChatAdministrators 返回 500 时消息照常交付且不带平台权限。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Descriptor 经 validate.AdapterRegister 检查通过，隐含关系为 telegram.chat_owner → telegram.chat_admin → chat_admin；注入时钟后 10 分钟内不重复获取、超过后重新获取；my_chat_member 作废；不同 chat 的缓存互不影响；sender_chat 消息与频道消息不调用；一次 reaction update 产生两个事件时只获取一次；失败的 WARN 日志不含 token；失败后下一个事件重新获取。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./adapters/...；go build ./...。

不做：Core 侧的权限过滤与展开（T-04.6 已完成）；按钮点击 Actor 的平台权限（随 T-03.8 的 callback_query 处理）；权限文案的 i18n 文件（label_key 的文案随 Adapter 包的 i18n 目录由打包切片加入）；OneBot 11 的平台权限（T-15.1）。
```

## 第 12 波

### goal-111 · T-03.20 官方 Telegram Adapter 的媒体与新增片段

状态：等待前置（T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：待派发；暂不派发（维护者指示只改写语句）。直接前置 T-03.13（PR #148，cf74fa9）、T-03.14（PR #143，15682e8）均已合并。2026-11-13 按 main 0af1255 重新核对。审查者补齐，待维护者确认：采用流式上传并隔离模拟端内存；file_base_url 的动态缺省在 Start 推导。按 DECISIONS 2026-11-13 变更记录，Telegram 对 Core 给出的 platform_file（同实例 incoming_ref）的发送消费已从本切片扣除，由 goal-9（T-03.10）以 renderMedia 最小分支交付。

- **需维护者决定：MediaRef MIME 的传递接口。** MESSAGE-MODEL §6.4/D-99 已要求按 MIME 选择 sendAnimation，但当前媒体来源仅传 token。建议先由 Core/SDK 的媒体所有者补齐「Core 从已校验对象取得 MIME 并随 Core→Adapter 来源传递」的最小协议前置；新增字段及具体编号须有明确契约授权，随后同步本 goal 的范围、接口、前置与 V-97 验收。当前语句保持 proto 原状，完整切片的派发与 MediaRef GIF 验收等待该决定；不撤销 D-99。

已有接口与衔接：buildSendPlan(to message.Session, chain message.Chain) (*sendPlan, *errcode.Error)、renderMedia(message.Segment, uint32) (map[string]any, *plannedUpload, *errcode.Error)、instance.readSendMedia(context.Context, string) ([]byte, error) 及 botClient.multipart(ctx context.Context, method string, params map[string]any, uploads []plannedUpload, files [][]byte, result any) error 仍为 main 上的实现；后两者的全量积累在本切片移除。newBotClient(token, base, proxyMode, proxy string)、botClient.fileBase、实例 ReadPlatformFile 已存在。MediaRef.MIME 存在，但 MediaRef.Source()、MediaSource、proto MediaSource 和 ReadMedia 的返回值均未传 MIME。VoiceSegment、URLButton、Event.ChatType 已存在；能力列表包含 message.send 的 4096 UTF-16 TextLimit。资源位于可导入包 adapters/telegram/；与 goal-9/112/113 共享文件，后合入者保留已合入能力、方法和平台发送分支。

```text
/goal 使用 $alonabot-execute 推进 goal-111（切片 T-03.20：官方 Telegram Adapter 的媒体与新增片段）。goal 目录 goal-111，分支 goal-111，worktree .worktrees/goal-111，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 出站媒体按来源传递，media_ref 用流式 multipart、inline_base64 解码上传；收发 voice、处理 GIF 与入站 animation/video_note/静态贴纸，支持 URL 按钮，追加 file_base_url 并为入站事件填写 chat_type。

依据：spec/PLAN.md §3.1；spec/plan/adapters/telegram.md T-03.20、T-03 步骤 17；spec/plan/core.md T-03「失败停止条件」；spec/CONTRACTS-v0.1.md §1.2～§1.4、§5.5、§8；spec/MESSAGE-MODEL.md §1、§2、§4、§6～§6.4；spec/WIRE-FORMAT.md §2、§3.4、§5；spec/adr/DECISIONS.md D-88、D-89、D-97、D-99、D-100、D-101、D-110、D-134；spec/TEST-MATRIX.md V-89、V-90、V-98 的 Telegram 分项与 V-97。

范围（只新增或修改这些）：
- adapters/telegram/telegram.go、client.go：file_base_url 解析与 HTTP 客户端、流式 multipart、message.voice 能力。
- adapters/telegram/plan.go、send.go：来源预检、voice/GIF 物理计划、流式执行与回执；删除 readSendMedia 的全量积累。
- adapters/telegram/render.go：URL 与 action 按钮渲染。
- adapters/telegram/inbound.go、media.go、callback.go：入站媒体类型、chat_type；callback.go 仅在回调 Target 的会话类型需要传递时修改。
- adapters/telegram/adapter_config_schema.json：追加 file_base_url；adapters/telegram/alonabot_gen.go 仅由该包既有生成命令更新。
- adapters/telegram/telegramtest/server.go、send.go：sendVoice/sendAnimation、可流式消费上传的模拟端。
- adapters/telegram/send_test.go：一个新增集成测试；media_test.go、telegram_test.go、edit_test.go、callback_test.go 仅调整受影响的共用替身或既有 API 调用，原断言保持；本地 *_local_test.go。

允许的新跨模块边：无。沿用 adapters/... -> sdk/...、contracts；实现用 io.Pipe、mime/multipart、encoding/base64 等标准库。SDK、Core、proto 及新依赖的改动遇到时标 [!]。

接口（照此实现；内部辅助函数自定）：
- New() adapter.Adapter、Instance.Send(ctx context.Context,to message.Session,chain message.Chain) (*message.SendResult,*errcode.Error) 与 adapter.MediaSource.ReadPlatformFile 保持现有签名。
- 配置 JSON 与 Schema 追加可选 file_base_url string，缺省语义为 <api_base_url>/file/bot<token>/；Schema 省略动态 default，由 Start 从同一实例的 api_base_url/token 推导。自定义前缀统一一个尾斜杠，然后连接 getFile 返回的 file_path；沿用 api_base_url 的 http/https URL 约束与实例代理、脱敏规则。Schema 的 $version 保持现值 1。
- 内部 multipart 调用删除 files [][]byte 参数，由实例提供当前操作的按块读取/写入函数；JSON 字段仍按计划编码，inline_base64 可使用至多 1 MiB 解码结果；不新增公开生产符号，不留下整份 []byte 请求体路径。
- telegramtest 增加
  func (s *Server) ConsumeUploads(fn func(method, field, filename string, src io.Reader) error)
  设置后逐个 multipart 文件 part 交给 fn 消费，Request.Files 不保存该文件的全量字节；未设置时 Requests() 的现有行为保持。回调错误通过模拟 HTTP 错误返回；用它验证大媒体内容和按块进展。
- 复用已有 proto：MediaSource.media_ref=1、incoming_ref=2、https_url=3、inline_base64=4、platform_file=5，Segment.voice、Button.url、Event.chat_type；沿用现有编号。MIME 传递的协议方案须先由维护者决定，本语句不授权新增 proto 字段。

行为：
- 先对整条 Chain 生成物理计划并验证能力、组合、长度与来源，保留 text/实体/相册/首条引用/末条按钮的现有语义。voice 每段单独 sendVoice，Duration 从 time.Duration 转成 Bot API 秒值；caption 仍来自该段。forward 在首个平台请求前 UNSUPPORTED。
- 来源：https_url 原样作为参数；platform_file 的发送消费由 T-03.10（goal-9）交付，本切片保留已合入分支且不实现该项；inline_base64 预检标准 base64 与大小，解码后上传；media_ref 用 Core.ReadMedia 每块至多 1 MiB 顺序读取，直接写入 multipart。一次只保留当前块及固定开销。
- 流式 multipart 使用 io.Pipe，写 goroutine 按计划写字段和文件，读取失败 CloseWithError；HTTP 返回/取消时关闭两端并回收写 goroutine。首个平台请求前已知的结构、来源与组合错误先拒绝；运行期读取失败按该请求的实际写出阶段判 PartRejected 或 PartUnknown。保留已确认结果，后续 PartNotAttempted，按 MESSAGE-MODEL §4 汇总。禁止业务重试。
- inline_base64 解码后以 GIF8 开头的 image 使用 sendAnimation，并切断前后连续媒体组；https_url 按普通 image，不额外下载探测。D-99 的 MediaRef MIME 判定保持原语义，但当前 MediaSource 与 ReadMedia 均无 MIME：该分项标 [!]，待状态段的协议决定及前置传递能力落地后补齐正式语句和验收，随后才派发完整切片。禁止以文件名、内容嗅探或额外 HTTP HEAD 代替已确认的 MIME 语义。
- 入站 voice → VoiceSegment；audio 仍 FileSegment；animation、video_note → VideoSegment；sticker 在 is_animated=false 且 is_video=false 时 → ImageSegment，其余只留 platform_payload。来源都为 IncomingMedia，保留平台 file_id、已给出的大小/MIME/时长，不为映射提前取件。
- URL 按钮输出 {text:label,url:URL}，无需 CoreToken；action 按钮继续输出 callback_data，并保留现有 token 检查。混排保持行列顺序，首发元数据限制不变。
- file_base_url 只影响文件下载，getFile 与其他 Bot API 仍经 api_base_url；proxy_mode none/core/custom 的语义全部沿用，token 及代理凭证仍脱敏，重定向规则保持。
- 每个带 Session 的入站 Event 按其 chat.type 写 ChatType：private=ChatPrivate、group/supergroup=ChatGroup、channel=ChatChannel；message、edited_message、channel_post、edited_channel_post、member、reaction、callback 同样处理。无 Session 保持 ChatUnspecified。thread_id 沿用 is_topic_message，ChatType 不写入会话字符串。
- Descriptor 与实例能力在原列表中只追加 message.voice v1；message.delete 与 commands.sync 分别由 T-03.21/T-03.22 追加，合并时保留已有项及 message.send 的 TextLimit。

测试（按编码规范「测试提交规则」）：
- 提交：send_test.go 追加一个集成测试 TestMediaSourcesAndNewSegments，经 telegramtest.Server 与真实 Adapter 实例；子场景验证 32 MiB media_ref multipart 的字节与块读取、voice 独立 sendVoice、可判定 GIF 独立 sendAnimation 且切断媒体组、URL/action 按钮混排。大媒体 Core 替身按 offset 生成每块，模拟端 ConsumeUploads 边读边比较，防止全量记录影响结果。流式内存证据同时用同一实现的隔离进程比较 32 MiB 与 100 MiB，对生产 Adapter 采样记录峰值与命令，验证额外常驻内存不随对象大小增长。多场景理由：来源与计划共用一次 Send 管道，V-89/V-97 要求同时验证流式与组合顺序。
- 本地自测（*_local_test.go，PR 写明结果）：非法 base64 与 1 MiB 边界；中途 ReadMedia 失败/取消的完整回执与 goroutine 收尾；相册前后 GIF、https_url GIF 按普通图片；入站 voice/audio/animation/video_note/静态与动态贴纸；file_base_url 自定义服务与代理；所有事件 ChatType、thread_id；纯 URL 按钮无 token；文本编辑的 URL 按钮复用；跨实例由 Core 准备的 media_ref；在已含 goal-9 的基线上只回归同实例 file_id 零 ReadMedia，发现回归交该实现的所有者修复；100 MiB 流式测量。MediaRef MIME 缺口未解决时该分项保持阻塞，不记录通过。
- 真实 Telegram：未运行（无凭证）。所有平台行为证据写明 telegramtest 本地模拟服务，附 OS/架构、Go 版本与命令；结果交审查者回写 TEST-MATRIX §3.2，区分 32 MiB 回归和 V-89 的 100 MiB 分项。

完成条件：本切片各映射、来源与内存分项有证据，状态段的 MIME 协议决定已落实并重新核对接口，MediaRef GIF 分项有证据后方可完成；gofmt -l 本次改动 Go 文件无输出；在执行分支运行 go generate ./adapters/telegram/... 并核对生成内容；go vet ./...；go test -race ./adapters/telegram/...；go build ./...。T-03 的失败停止条件逐项满足，范围外修改与协议缺口标 [!]。

不做：Telegram platform_file 发送消费（已移交 goal-9）、平台动作/撤回/命令菜单、Core 来源租约重写、通用流式 API（已有 c.Stream）、新媒体转码、chat_title、服务消息路由、OneBot 平台实现、真实平台验证。
```

### goal-112 · T-03.21 官方 Telegram Adapter 的平台动作与撤回

状态：等待前置（T-03.20、T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：等待前置（T-03.20）；暂不派发（维护者指示只改写语句）。原前置 T-03.11（PR #106，3913d5a）、T-03.15（PR #146，ab5f8fb）、T-03.16（PR #144，707b74c）均已合并。2026-11-13 按 main 0af1255 重新核对。新增 T-03.20（goal-111）为直接前置：DraftStream 的 c.IsPrivate() 使用事件 ChatType，当前 Telegram 尚未填写它。以下裁决均记为「审查者补齐，待维护者确认」：保留单次 100 个 ID 上限；action 使用现成 sdk/message、sdk/errcode 公共类型；结果常用字段类型化、完整 JSON 保留；管理员缓存采用同一同步和失效代次。

- **需维护者决定：Telegram Bot API 的原文基线及候选动作输入范围。** 本环境的网络和浏览器入口均未取得可用官方正文，详细证据见 adapters-decisions.md。当前 SendMessageDraftParams、ChatPermissions、MessageEntity/ReactionType 的候选字段未完成官方原文复核。建议按选定官方版本保留私聊限定、冻结下列最小输入字段并保留完整结果 Raw；审查者取得原文后逐字段核对和回写，维护者确认真实差异涉及的范围。原文齐备前保持本切片暂不派发，不把候选字段记成已验证的平台契约。

已有接口与衔接：adapter.Action[P,R](name, version, scope)、ActionExecutor.Action、Deleter.Delete、sdk.ActionCaller.CallAction、sdk.NewActionClient 均可沿用。sdk 根包没有 Session/MessageRef/SendResult 别名，类型在 sdk/message；错误码在 sdk/errcode。sdk/internal/convert/delete.go 的 maxDeleteMessages=100。instance.administrators 当前仅由接收循环读写，新增 Action 并发写方须同步整个缓存。Context.Instance 当前尚未合并，本切片使用现有父调用的目标补齐。与 goal-111/113 共享 telegram.go、client.go、send.go 和 telegramtest；RespondMethod 使用相同签名，先合入的实现直接复用。工具调用、WebUI、plugintest、动态任务、Agent 对话、run/热部署与启动取消修复均已合并；测试继续驱动平台 Adapter 与临时 SDK 插件，无需改其装配。

```text
/goal 使用 $alonabot-execute 推进 goal-112（切片 T-03.21：官方 Telegram Adapter 的平台动作与撤回）。goal 目录 goal-112，分支 goal-112，worktree .worktrees/goal-112，从包含已合并 goal-111（T-03.20）的最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 声明并实现 MESSAGE-MODEL §6.5 的十四个平台动作，发布带类型的 action 包与输入状态、草稿辅助函数；实现 deleteMessages 与 message.delete 能力，消息动作通过消息句柄选择原实例。

依据：spec/PLAN.md §3.1；spec/plan/adapters/telegram.md T-03.21、T-03 步骤 18；spec/plan/core.md T-03「失败停止条件」；spec/CONTRACTS-v0.1.md §1.2、§2.3、§8（平台动作、撤回）；spec/MESSAGE-MODEL.md §4、§4A、§6.1、§6.5；spec/WIRE-FORMAT.md §2、§3.4、§3.5；spec/adr/DECISIONS.md D-87、D-92、D-93、D-98、D-104、D-113、D-137；spec/TEST-MATRIX.md V-91、V-92、V-94 的 Telegram 分项、V-102 的动作分项、V-111 的类型声明。Bot API 原文对应十四个方法；状态段指定的原文版本及候选字段必须在派发前核定，冲突交审查者修订语句。

范围（只新增或修改这些）：
- 新增 adapters/telegram/action/：参数与结果类型、十四个带类型函数、Typing 与 DraftStream；不导入 Adapter 运行时。
- 新增 adapters/telegram/action.go、delete.go：动作调度、Bot API 转换与撤回回执。
- adapters/telegram/client.go：私有调用路径保留 result 原文与请求写出事实；现有 call 的调用者沿用原签名和行为。adapters/telegram/send.go 仅补共用安全错误映射中的取消/到期，不改发送计划。
- adapters/telegram/telegram.go：Descriptor.Actions、message.delete 能力、实例的管理员缓存同步字段。
- adapters/telegram/permissions.go：让接收与动作作废同一缓存时安全并发，保留缓存获取和失效语义。
- adapters/telegram/telegramtest/server.go、send.go 与新增 telegramtest/action.go：动作与 deleteMessages 的模拟响应、参数记录、故障注入。
- adapters/telegram/telegram_test.go：一个新增平台集成测试；其他既有测试只同步受影响的辅助调用；本地 *_local_test.go。action 包的纯函数/辅助函数测试放本地自测。

允许的新跨模块边：action 包使用 sdk、sdk/message 与 sdk/errcode（均为现有公共包），只用这些公共 API 与标准库，不 import telegram、sdk/adapter 或 Adapter 运行时。ARCHITECTURE 已有 adapters/... -> sdk/...，审查者需同步其中 action 常量包“只 import sdk 根包”的旧限制及 MESSAGE-MODEL §6.5/PLAN 的同句。其余方向不增加。

接口（照此实现；以下签名按现有 SDK 固定；平台字段的原文复核门禁见状态段）：
1. 实例实现
   Action(ctx context.Context,to message.Session,name string,version uint32,params json.RawMessage) (json.RawMessage,*errcode.Error)
   Delete(ctx context.Context,to message.Session,platformMessageIDs []string) (*message.DeleteResult,*errcode.Error)
   Descriptor.Actions 全部用 adapter.Action[P,R]("telegram.<snake_name>",1,adapter.ActionScopeChat)；参数 Schema 从 Go struct 推导，禁止额外属性，结果允许平台附加字段。
2. action 包的会话动作，保留 opts ...sdk.SendOption 供 sdk.Via：
   func BanChatMember(c sdk.ActionCaller,to message.Session,p BanChatMemberParams,opts ...sdk.SendOption) (bool,error)
   func UnbanChatMember(c sdk.ActionCaller,to message.Session,p UnbanChatMemberParams,opts ...sdk.SendOption) (bool,error)
   func RestrictChatMember(c sdk.ActionCaller,to message.Session,p RestrictChatMemberParams,opts ...sdk.SendOption) (bool,error)
   func GetChat(c sdk.ActionCaller,to message.Session,opts ...sdk.SendOption) (Chat,error)
   func GetChatMember(c sdk.ActionCaller,to message.Session,p GetChatMemberParams,opts ...sdk.SendOption) (ChatMember,error)
   func GetChatMemberCount(c sdk.ActionCaller,to message.Session,opts ...sdk.SendOption) (int64,error)
   func GetChatAdministrators(c sdk.ActionCaller,to message.Session,opts ...sdk.SendOption) ([]ChatMember,error)
   func SetChatTitle(c sdk.ActionCaller,to message.Session,p SetChatTitleParams,opts ...sdk.SendOption) (bool,error)
   func LeaveChat(c sdk.ActionCaller,to message.Session,opts ...sdk.SendOption) (bool,error)
   func SendChatAction(c sdk.ActionCaller,to message.Session,p SendChatActionParams,opts ...sdk.SendOption) (bool,error)
   func SendMessageDraft(c sdk.ActionCaller,to message.Session,p SendMessageDraftParams,opts ...sdk.SendOption) (bool,error)
3. 指向消息的动作从句柄取会话、消息与实例：
   func PinChatMessage(c sdk.ActionCaller,ref message.MessageRef,p PinChatMessageParams) (bool,error)
   func UnpinChatMessage(c sdk.ActionCaller,ref message.MessageRef) (bool,error)
   func SetMessageReaction(c sdk.ActionCaller,ref message.MessageRef,p SetMessageReactionParams) (bool,error)
   校验 ref.Parts()；调用 CallAction 时用句柄实例构造 sdk.Via，省略实例时交 Core 补齐。不另加可覆盖句柄实例的参数。
4. 参数类型的 JSON 字段（沿用 Bot API 名称；下列字段映射到同名 Go 导出字段，ID 用 string；可选字段用指针或 omitempty，须保留显式 false）：
   BanChatMemberParams：user_id string，until_date *int64,omitempty，revoke_messages *bool,omitempty。
   UnbanChatMemberParams：user_id string，only_if_banned *bool,omitempty。
   RestrictChatMemberParams：user_id string，permissions ChatPermissions，use_independent_chat_permissions *bool,omitempty，until_date *int64,omitempty。
   GetChatMemberParams：user_id string。
   SetChatTitleParams：title string。
   PinChatMessageParams：disable_notification *bool,omitempty；其消息字段由句柄包装成平台动作参数。
   SendChatActionParams：action string。
   SetMessageReactionParams：reaction []ReactionType（允许空列表），is_big *bool,omitempty；其消息字段由句柄包装成平台动作参数。
   SendMessageDraftParams（候选平台字段，派发前须按状态段复核）：draft_id string，text string，parse_mode string,omitempty，entities []MessageEntity,omitempty；draft_id 在平台编码处转十进制整数。
   空参数的动作使用 struct{}；带消息 ID 的声明具体为 PinChatMessageActionParams{MessageID string `json:"message_id"`; DisableNotification *bool `json:"disable_notification,omitempty"`}、UnpinChatMessageActionParams{MessageID string `json:"message_id"`}、SetMessageReactionActionParams{MessageID string `json:"message_id"`; Reaction []ReactionType `json:"reaction"`; IsBig *bool `json:"is_big,omitempty"`}。包装函数从句柄构造这三个类型，插件侧不重复传消息 ID。chat_id、message_thread_id 不进入参数 struct。
   ChatPermissions 是具体 struct，下列 Go 字段全部为 *bool，JSON 键为括号内名称并带 omitempty：CanSendMessages(can_send_messages)、CanSendAudios(can_send_audios)、CanSendDocuments(can_send_documents)、CanSendPhotos(can_send_photos)、CanSendVideos(can_send_videos)、CanSendVideoNotes(can_send_video_notes)、CanSendVoiceNotes(can_send_voice_notes)、CanSendPolls(can_send_polls)、CanSendOtherMessages(can_send_other_messages)、CanAddWebPagePreviews(can_add_web_page_previews)、CanChangeInfo(can_change_info)、CanInviteUsers(can_invite_users)、CanPinMessages(can_pin_messages)、CanManageTopics(can_manage_topics)。这些字段是当前草稿冻结的最小平台输入集合；完整支持范围和所选 Bot API 原文的核对结论须先写入状态段所述裁决，随后才派发。
   type ReactionType struct { Type string `json:"type"`; Emoji string `json:"emoji,omitempty"`; CustomEmojiID string `json:"custom_emoji_id,omitempty"` }
   type User struct { ID int64 `json:"id"`; IsBot bool `json:"is_bot"`; FirstName string `json:"first_name"`; LastName string `json:"last_name,omitempty"`; Username string `json:"username,omitempty"`; LanguageCode string `json:"language_code,omitempty"` }
   type MessageEntity struct { Type string `json:"type"`; Offset uint64 `json:"offset"`; Length uint64 `json:"length"`; URL string `json:"url,omitempty"`; User *User `json:"user,omitempty"`; Language string `json:"language,omitempty"`; CustomEmojiID string `json:"custom_emoji_id,omitempty"` }
   type Chat struct { ID int64 `json:"id"`; Type string `json:"type"`; Title string `json:"title,omitempty"`; Username string `json:"username,omitempty"`; FirstName string `json:"first_name,omitempty"`; LastName string `json:"last_name,omitempty"`; Raw json.RawMessage `json:"-"` }
   type ChatMember struct { Status string `json:"status"`; User User `json:"user"`; Raw json.RawMessage `json:"-"` }
   结果类型固定上述常用字段，完整对象由 Raw 保留；不按平台所有可选输出字段扩建 SDK 类型。Chat/ChatMember 的带类型函数同时保存各自完整的 result JSON 到 Raw；数组逐项保存各元素 JSON。Raw 由包装函数解码后复制，不给声明类型添加自定义 JSON 编码方法。平台数字结果 ID 保留 Bot API 原文的数值，动作输入 ID 使用字符串并在 Adapter 编码边界转换。不把未知平台结果字段丢弃；result Schema 允许附加字段。
5. 辅助函数：
   func Typing(c sdk.ActionCaller,ctx context.Context,to message.Session,action string,opts ...sdk.SendOption) error
   func DraftStream(c *sdk.Context,draftID string,next func(context.Context) (string,error)) (*message.SendResult,error)
   Typing 同步运行到 ctx 取消或首次错误，由调用者决定是否起 goroutine；DraftStream 固定当前 Handler 的会话与实例，最终调用 c.Send(message.Text(fullText))，保留返回回执与错误；草稿 CallAction 与最终 Send 均经当前 c 的父调用补齐原实例，无需尚未合并的 Context.Instance 方法。
6. telegramtest：沿用 Requests、FailMethod、Drop；新增
   func (s *Server) RespondMethod(method string,result json.RawMessage)
   覆盖方法成功时的 Bot API result，参数照常记录；默认布尔动作/deleteMessages 返回 true，查询动作返回对应结构，getChatAdministrators 沿用 SetChatAdministrators 的响应。不另造生产测试插件。
7. 沿用现有 ActionDeclaration、AdapterActionRequest、ActionResponse 与 AdapterDeleteRequest/DeleteResult proto；不改协议编号。

行为：
- 动作完整名称依次为 telegram.ban_chat_member、telegram.unban_chat_member、telegram.restrict_chat_member、telegram.pin_chat_message、telegram.unpin_chat_message、telegram.get_chat、telegram.get_chat_member、telegram.get_chat_member_count、telegram.get_chat_administrators、telegram.set_chat_title、telegram.leave_chat、telegram.send_chat_action、telegram.set_message_reaction、telegram.send_message_draft。每个均声明版本 1/chat 范围，并明确映射同名 Bot API 驼峰方法，禁止未声明动作透传。
- 会话与话题从 message.Session.Parts() 拆出，经 Core 准入的实例绑定到当前 instance；需要消息的动作从 JSON 的 message_id 取值。参数中的十进制 ID 在平台请求编码处转成 API 所需数字形式，错误在平台请求前 INVALID_ARGUMENT；结果返回 Bot API result JSON 原文，带类型函数解析它且保留可用结果与同时发生的错误。
- ban/unban/restrict 成功后作废该会话的 administrators 缓存，失败时保留；接收线程与动作并发访问统一同步，不在锁内发网络请求，读取、发布与失效共享同一把锁及每会话的失效代次：网络请求前记录代次，返回后仅在代次未变时发布和使用权限；代次已变时丢弃该次旧结果，本事件不附加旧权限，下个正常事件再获取。成功的 ban/unban/restrict 与成员变更都推进代次。权限仍由平台事实给出，Core 负责展开权限集合。包装函数沿用 ActionCaller 的父链与期限。
- 动作的平台明确拒绝返回现有安全错误映射；写出后断线或超期分别 UNAVAILABLE/DEADLINE_EXCEEDED，保留 result 原文（若有），不重试。日志和错误沿用 Bot token 脱敏。
- send_chat_action 带目标话题的 message_thread_id；Typing 立即调用一次，其后每 4 秒一次，每次调用前检查 ctx，ctx 结束停止后续请求，首个错误停止并返回。在途调用的取消/期限由传入 ActionCaller 的绑定 context 决定：Handler 传 c.Context()，后台传同一 ctx 创建的 sdk.NewActionClient(ctx)，保留父链；不创建脱离父链的后台客户端替换 c。不加入额外重试或可配置期限。
- set_message_reaction 是设置完整列表，空列表用于清空；消息来自句柄，不从当前事件猜测。参数字段与所需 message_id 由同一声明类型对应。
- send_message_draft 只用于私聊；Adapter 按合法 Telegram 数字 chat_id 的私聊语义（正值）判断，其他目标在请求前 UNSUPPORTED；不得调用 getChat 探测后才拒绝。draft_id 解析为非零十进制整数，text、parse_mode/entities 按已冻结的平台字段校验，parse_mode 与 entities 互斥。DraftStream 使用 T-03.20 已填写的可信 ChatType，c.IsPrivate() 为 false 时先返回 UNSUPPORTED；next 用 c.Context() 调用，累积文本发草稿，io.EOF 时处理最后增量并经普通 c.Send 正式发出一次；next 错误、取消或草稿请求失败停止并返回完整错误，不发伪成功正式消息。草稿走平台动作，正式发送走现有出站 Hook。该函数不调用 c.Stream，保留独立的草稿语义；空输入不发消息。
- Delete 受现有接口的 1～100 输入约束，合法列表一次 deleteMessages；平台要求的 chat_id 来自同一会话，不附加 thread_id。一次请求共享成功/拒绝/UNKNOWN 结果，逐项按原输入顺序生成，ID 不去重；HTTP 400 对该次每项为 REJECTED，写出后结果未知为 UNKNOWN，确认尚未写出时为 REJECTED。本入口最多一个合法批次，不生成不存在的后续批次。平台 result=false 按 FAILED_PRECONDITION 拒绝，禁止作为 ACCEPTED。顶层错误与逐项错误、可用回执一起保留。
- 追加 message.delete v1 能力，不改变原 message.send TextLimit、编辑、按钮、权限及其他切片新增能力。单次 101 个输入应由 SDK/Core 在平台请求前 INVALID_ARGUMENT；跨两次合法调用的 100+1 个 ID 产生两次平台请求。

测试（按编码规范「测试提交规则」）：
- 提交：telegram_test.go 一个新增集成测试 TestPlatformActionsAndDelete，经 telegramtest.Server 和 New().Start 实例，复用记录 Core：先由群消息获取管理员缓存，调用 Action 的 ban_chat_member，断言 chat_id、user_id、参数和 true 原文；下一群消息重新 getChatAdministrators。成功场景以两次独立 Deleter 调用提交 100 个和 1 个 ID，断言两次 deleteMessages 的参数与逐项 ACCEPTED；独立故障场景令第二次合法调用返回 400，断言该次每项 REJECTED。多场景理由：本切片两条独立平台入口、缓存并发失效与撤回共享回执均需通过 HTTP 集成验证。单次 101 个输入由本地 SDK/Core 自测确认平台请求前 INVALID_ARGUMENT。
- 本地自测（*_local_test.go，PR 写明结果）：十四个声明能类型推导、参数额外字段拒绝、result 额外字段保留；十四个动作参数与结果逐一对照原文；pin/unpin/reaction 的句柄含实例与话题、显式 false、回应设置与清空；Typing 立即一次/4 秒刷新/取消停止；DraftStream 私聊、群聊零请求、最终 Send 与 Hook、错误和取消；Delete 的 1/100 与 101 个前置拒绝、重复 ID、400、断线/超期及安全错误。辅助函数经 ActionCaller 记录替身自测；父链、最终 Hook 的集成使用临时 SDK 插件与既有 wire/Core 入口，测试文件仅本地，不新增 testdata 插件。
- 真实 Telegram：未运行（无凭证）。Bot API 原文核对与 telegramtest 模拟服务行为分别记证据，给审查者 TEST-MATRIX §3.2 回写所需的 OS/架构、Go 版本、命令与结果。

完成条件：状态段的平台原文复核已完成，T-03.20 的 ChatType 已合并，所有声明和主路径有证据；gofmt -l 本次改动 Go 文件无输出；go vet ./...；go test -race ./adapters/telegram/...；go build ./...。范围外修改、契约冲突、依赖新增与 Bot API 差异标 [!]，按 T-03「失败停止条件」处理。

不做：Core 准入、权限模型和通用协议改动、SDK 根包类型别名、通用输入状态或草稿服务、通用 c.Stream 重写、重试、Telegram 媒体与命令菜单、真实平台验证。
```

### goal-113 · T-03.22 官方 Telegram Adapter 的命令菜单

状态：等待前置（T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：待派发；暂不派发（维护者指示只改写语句）。前置 T-03.17（PR #150，8eb7846）已合并。2026-11-13 按 main 0af1255 重新核对。以下裁决均记为「审查者补齐，待维护者确认」：主语言键优先、否则完整 locale 字典序选一个菜单语言；撤销语言用 deleteMyCommands；有效命令超 100 条或说明为空在请求前拒绝；诊断保留、平台调用按成功快照去重。

- **需维护者决定：与 goal-112 共用的 Bot API 原文版本基线。** 本环境未取得可用官方正文，100 条、说明 1～256、两字母语言码和 deleteMyCommands 的平台语义须由审查者据选定版本复核后回写。当前方案已落入正文供审阅，真实平台差异涉及已确认决策时交维护者裁决，原文复核完成前保持暂不派发。

已有接口与衔接：adapter.CommandMenuItem 为 convert.CommandMenuItem 的别名，字段 Name string、Descriptions map[string]string；CommandSynchronizer.SyncCommands(context.Context, []CommandMenuItem) *errcode.Error 已存在。SDK 对实际回调持有实例的 commandSync 锁。Core 的 router.commandMenu 已生成 en-US 及目录 locale 的说明，platformCapabilities() 同时用于 Descriptor 与实例。main 尚无 setMyCommands、菜单缓存。与 goal-111/112 共用 telegram.go 和 telegramtest 分派；RespondMethod 按同一接口复用，保留现有方法、权限和 TextLimit。

```text
/goal 使用 $alonabot-execute 推进 goal-113（切片 T-03.22：官方 Telegram Adapter 的命令菜单）。goal 目录 goal-113，分支 goal-113，worktree .worktrees/goal-113，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 实例接收 adapter.commands.sync 的命令快照，以 setMyCommands 同步默认菜单与各语言说明，声明 commands.sync；按每个 language_code 记住成功内容并去重，失败保留错误且不阻断其他语言。

依据：spec/PLAN.md §3.1；spec/plan/adapters/telegram.md T-03.22、T-03 步骤 19；spec/plan/core.md T-03「失败停止条件」；spec/CONTRACTS-v0.1.md §5.4、§8（命令菜单）；spec/MESSAGE-MODEL.md §6.5；spec/WIRE-FORMAT.md §2、§3.4 的 adapter.commands.sync；spec/adr/DECISIONS.md D-96、D-138；spec/TEST-MATRIX.md V-93 的 Telegram 分项；Bot API setMyCommands/BotCommand/BotCommandScopeDefault 的原文。

范围（只新增或修改这些）：
- 新增 adapters/telegram/commands.go：菜单转换、成功快照与去重。
- adapters/telegram/telegram.go：commands.sync 能力与实例内缓存字段。
- adapters/telegram/telegramtest/server.go、send.go：setMyCommands、deleteMyCommands 的参数记录、true/false 响应与 FailMethod 故障注入。
- adapters/telegram/telegram_test.go：一个新增集成测试；本地 *_local_test.go。

允许的新跨模块边：无。只用既有 sdk/adapter、sdk/message、sdk/errcode 与标准库；无需配置、proto、Core、i18n 生成器或插件注册 API 改动。

接口（照此实现；内部辅助函数自定）：
- 实例实现 sdk/adapter.CommandSynchronizer：
  func (i *instance) SyncCommands(ctx context.Context,commands []adapter.CommandMenuItem) *errcode.Error
- 输入项沿用 adapter.CommandMenuItem{Name string,Descriptions map[string]string}。输入是当前可用命令的完整替换快照，空列表用于清空。
- Descriptor 与实例能力表追加 message.Capability{Name:"commands.sync",Version:1,Supported:true}；菜单成功缓存归实例，键为实际发出的 language_code（空串表示默认），值为上一次确认成功的完整 [{command,description}] 快照及存在标记。
- setMyCommands 请求 JSON 为 {"scope":{"type":"default"},"commands":[{"command":string,"description":string}],"language_code":string（默认菜单省略）}；deleteMyCommands 请求只带同一 scope 与待撤销的 language_code；响应 result 均为 true。
- telegramtest 沿用 Calls、Requests、FailMethod；setMyCommands 与 deleteMyCommands 进入 Requests 的顺序记录，默认返回 true；追加 func (s *Server) RespondMethod(method string, result json.RawMessage) 以注入 result=false，与 goal-112 同签名，只保留一份共用响应覆盖。无新增公开生产符号。
- 复用既有 CommandMenuItem、AdapterCommandsSyncRequest/Response；不追加 proto 字段。

行为：
- Core 已筛选根命令、权限为 user、策略启用且实例前缀含 / 的命令，并按 D-138 提供语言说明；Adapter 按输入顺序转换，不重新读取路由或权限。名称不匹配 ^[a-z0-9_]{1,32}$ 时跳过并记诊断；说明超过 256 Unicode 字符时在字符边界截断并记诊断。保留实际成功发送的内容作为比较依据。
- 默认 scope 的无 language_code 菜单使用输入 en-US 原文。从全部 Descriptions 键收集规范 locale，以主语言的小写两字母值作为 Telegram language_code；三字母、und、private-use 等无法映射的键跳过并记诊断。对同主语言的候选，优先无区域/脚本的主语言键；其余按完整 locale 的 ASCII 字典序取首个，整份菜单统一使用该 locale。某命令缺该 locale 时使用它的 en-US（Core 已做其已有 locale 的回落）；不重新执行 i18n。en-US 同时生成 language_code=en 的菜单，默认菜单始终存在。候选和实际语言排序后再处理，结果与 map 插入顺序无关。
- 全量预检在任何平台请求前完成：过滤非法命令名后最多 100 条；超出返回 INVALID_ARGUMENT，保留实际数量且零平台请求。说明按 Unicode 码点截断至 256，截断后长度须为 1～256；缺少 en-US、空说明返回 INVALID_ARGUMENT，错误带命令/语言且不补占位文案。输入中的有效命令保持顺序。
- 每个实际语言独立比较成功快照：相同内容跳过，不重复 setMyCommands；首次同步空列表仍须调用一次以清除可能残留的菜单，缓存存在标记与空内容分开。成功后复制保存，避免调用方随后修改切片/map 影响缓存。
- 顺序确定：默认菜单先发，之后实际 language_code 排序处理；某语言的平台明确失败继续其余语言，返回按处理顺序遇到的第一个安全错误；ctx 取消时停止发起后续请求，未处理项保留原缓存。失败语言不更新成功快照，下一次相同输入仍会尝试它；已成功语言跳过。平台 result=false 返回 FAILED_PRECONDITION 并保留原缓存。诊断在每次实际同步收到非法名或超长说明时各记录一次，即使转换结果与成功快照相同也保留诊断；去重只控制平台调用。
- 命令消失时以 setMyCommands 替换新的完整 commands 列表。对上次成功而本次完全消失的 language_code，用 deleteMyCommands（同一 default scope 和该语言码）撤销其旧菜单，使平台恢复语言回落；成功后移除该语言成功快照，失败保留以便下一次同步继续清理。新语言设置和旧语言撤销一起按 language_code 排序，默认菜单先处理；Stop 仍保持平台菜单。
- 菜单同步无自动重试、无新增定时器；Core 的两秒合并、同步通知及失败后下次变化处理全部沿用。SyncCommands 失败不 ReportFailure、不停实例；Stop 不调用平台删除菜单。缓存仅随实例存活，不持久化。
- 记录诊断时沿用默认 Logger 和 Bot token 脱敏，保留方法失败的安全错误；message.send 的 TextLimit、权限、其他切片能力保持。

测试（按编码规范「测试提交规则」）：
- 提交：telegram_test.go 一个集成测试 TestCommandMenuSync，经 telegramtest.Server 驱动 New().Start 得到实例，再转为 adapter.CommandSynchronizer：同步两次相同的 en-US 快照，默认菜单实际只出现一次 setMyCommands；断言默认及 en 的 scope、command、description、language_code；修改说明后各更新一次；加入 zh-CN 后生成 zh，移除该 locale 时有 deleteMyCommands(language_code=zh)，其余成功快照不重发；清空时默认 setMyCommands(commands=[])，旧 en 菜单撤销；Descriptor 与实例均声明 commands.sync。测试调用模拟 HTTP，不采用纯函数测试代替集成主路径。
- 本地自测（*_local_test.go，PR 写明结果）：en-US/zh-CN 等语言映射与冲突、旧语言移除；非法名称、256/257 字符与多字节文本；同内容不同 map 插入顺序；初次空列表；某语言失败后继续其他语言、下一次只重试失败内容；result=false、断线、ctx 取消；Stop 不删除菜单；不同实例缓存隔离、调用方修改原切片/map 不影响缓存。100/101 条有效命令、空说明和缺 en-US 的零平台请求；相同内容仍有非法/截断诊断；语言清理失败保留缓存并在下次同步再次清理。
- 真实 Telegram：未运行（无凭证）。证据明确写 telegramtest 本地模拟服务，提供 OS/架构、Go 版本与实际命令，交审查者回写 TEST-MATRIX §3.2；Core 的 V-93 分项沿用 PR #150 的证据。

完成条件：状态段的官方原文复核完成，上述语言映射、上限预检和撤销语义均有证据，菜单主路径与失败不污染缓存有证据；gofmt -l 本次改动 Go 文件无输出；go vet ./...；go test -race ./adapters/telegram/...；go build ./...。发生契约冲突、范围外修改或 T-03 失败停止条件时标 [!]。

不做：Core 的根命令筛选、两秒通知合并与策略/i18n 规则改写、生成器扫描注册、插件生命周期装配、Telegram 动作/媒体/撤回、菜单持久化、停止删除菜单、自动重试、真实平台验证。
```

## 第 14 波

### goal-132 · T-03.26 官方 Telegram Adapter 填写 chat_title

状态：等待前置（T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：待派发。暂不派发（维护者指示只改写语句）。直接前置 T-03.4（goal-5，PR #39，合并提交 `604dc65`）、T-03.25（goal-131，PR #166，合并提交 `6f72fc8`）均已合并。2026-11-13 按维护者指定的 main `9efee73` 只读副本，用 `rg` 核对定义、调用者与既有集成测试；`message.Event`、`adapter.Core`、`telegramtest.Server.Calls` 另经离线 `go doc` 核对输出。以下补齐方案均为「审查者补齐，待维护者确认」。

按 main 核对：

- `sdk/message/event.go`：`message.Event` 已有 `ChatTitle string`、`ChatType message.ChatType`、`Session message.Session`；`ChatTitle == ""` 表示未提供。`contracts/proto/alonabot/v1/message.proto` 已有 `ChatType chat_type = 15`、`optional string chat_title = 16`；`sdk/internal/convert/event.go` 的 `func EventToProto(e *message.Event) (*pb.Event, error)`、`func EventFromProto(e *pb.Event) *message.Event` 已双向转换名称。
- `sdk/adapter/adapter.go`：`Core.Deliver(ctx context.Context, ev *message.Event, cursor []byte) (accepted bool, err error)`。`adapters/telegram/telegram.go`：`func New() adapter.Adapter`、`func (a *telegramAdapter) Start(ctx context.Context, request adapter.StartRequest, core adapter.Core) (adapter.Instance, error)`；接收经 `processUpdate(ctx context.Context, raw json.RawMessage, saved cursor) (cursor, error)`、`save(ctx context.Context, event *msg.Event, next cursor) error` 交付，沿用逐条应答及游标规则。
- `adapters/telegram/inbound.go`：`telegramChat` 当前只有 `ID int64`、`Type string`；`func (i *instance) baseEvent(chat telegramChat, date int64, user *telegramUser) *msg.Event` 由消息、编辑、成员与 reaction 路径共用；`messageEvent(received *telegramMessage, edited bool) (*msg.Event, error)` 负责消息与话题。`adapters/telegram/callback.go` 的 `callback(ctx context.Context, queryJSON json.RawMessage, updateID int64, raw json.RawMessage) ([]*msg.Event, error)` 同样调用 `baseEvent(query.Message.Chat, ...)`；没有 message 的 inline 回调当前只确认并跳过交付。
- `adapters/telegram/telegramtest/server.go` 已有 `NewServer(t testing.TB, token string) *Server`、`(*Server).URL() string`、`Push(updates ...json.RawMessage)`、`Calls(method string) int`、`SetChatAdministrators(chatID int64, status int, members ...json.RawMessage)`；可以直接注入含名称的 Chat。`adapters/telegram/telegram_test.go` 已有 `recordingCore`、`newRecordingCore()`、`startRequest`、`waitSaved` 和 `registerStop`，足以覆盖真实 HTTP 接收链。
- 审查者补齐，待维护者确认：当前 `baseEvent` 尚未填写 `ChatType`；在同一构造点一并落实 MESSAGE-MODEL §6 已确认的 D-100 映射，让新目录名称与会话类型配套。goal-111 原先规划的类型赋值由本切片承担，审查者在正式派发时同步其语句；已经合并时复用现有赋值。私聊姓名按非空字段连接，保留字段内容，只有一项时直接取该项。
- 在途协调按本次维护者提供的信息：PR #177（goal-66）涉及 `internal/core/logging/`、`internal/core/manage/`、`sdk/log.go`、`sdk/adapter_run.go`、`alonabot.go`、`assembly.go`、`cmd/alonabot/`；PR #178（goal-60）涉及 `agent/`；goal-41 涉及 `internal/core/journal/`、`internal/core/state/`、`internal/core/manage/`、`internal/core/supervisor/` 与根包装配。这些文件与本切片生产修改范围分离。goal-9 的 Telegram 生产修改点是 `adapters/telegram/plan.go`，其 CLI 集成测试与本切片接收测试分离；派发时核对合并后的事件断言。另与尚待实施的 goal-111、goal-137 共用 `adapters/telegram/inbound.go`、`telegram_test.go`，后合并者保留完整字段和断言。
- 旧在途提示按完成记录更新：PR #165（`330164f`）、#167（`3a9b4af`）、#168（`37b0ada`）、#169（`4dc4bcd`），以及 goal-59／PR #170（`eca6310`）、goal-69／PR #171（`3043c4e`）均已合并。最近的 #173（`2bb8181`）、#175（`ac6083a`）、#176（`9785c22`）及维护者告知已合并的 #174 作为基线成果保留；本切片只消费现有 Adapter 交付接口。
- 需维护者另行决定：新增依赖、改变已确认决策均无新增请求；真实 Telegram 凭证与实测由维护者决定是否提供，本切片模拟验收可独立完成，真实平台结果留作「未运行」。

```text
/goal 使用 $alonabot-execute 推进 goal-132（切片 T-03.26：官方 Telegram Adapter 填写 chat_title）。goal 目录 goal-132，分支 goal-132，worktree .worktrees/goal-132，从包含 PR #39、#166 的最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 实例为每个已支持、带 target 的入站事件填写平台给出的 chat_title，群、超级群与频道取聊天名称，私聊取对方姓名；在共用构造点填写既定 chat_type，让 Core 已合并的聊天目录取得名称与类型。

依据：spec/PLAN.md §3.1、§5；spec/plan/adapters/telegram.md T-03.26、T-03 步骤 23；spec/plan/core.md T-03 的「失败停止条件」及步骤 22 的已合并聊天目录；spec/CONTRACTS-v0.1.md §8（Adapter 交付与聊天目录）、§14.1（日志）；spec/MESSAGE-MODEL.md §1、§6、§6.1、§6.3；spec/WIRE-FORMAT.md §3.4（core.adapter.deliver）、§5A（Event 的追加字段）；spec/adr/DECISIONS.md D-86、D-100、D-130；spec/TEST-MATRIX.md V-117 的 Telegram 分项及 §3 的证据格式。平台字段按 Telegram Bot API 的 Chat 对象核对。

范围（只新增或修改这些）：
- adapters/telegram/inbound.go：telegramChat 补字段；baseEvent 统一填写名称与既定类型；内部辅助函数放在本文件。
- adapters/telegram/telegram_test.go：追加下述一个集成测试；既有集成测试的 Event 预期随新增字段同步，保留原有行为断言。
- adapters/telegram/ 下本切片的 *_local_test.go：其余名称、事件路径与日志检查，只作本地自测。
允许的新跨模块边：无；沿用 Adapter 到 sdk/adapter、sdk/message 及本模块 telegramtest 的现有依赖。

接口（照此实现；内部辅助函数自定）：
- 保持 func New() adapter.Adapter、Adapter.Start、Instance 及 adapter.Core 的全部公开签名。
- telegramChat 在现有字段上追加 Title string `json:"title"`、FirstName string `json:"first_name"`、LastName string `json:"last_name"`；继续供 telegramMessage、membershipUpdate、reactionUpdate 与回调 Message 共用。
- 保持 func (i *instance) baseEvent(chat telegramChat, date int64, user *telegramUser) *msg.Event；返回的 msg.Event.ChatTitle 为 string，ChatType 为 msg.ChatType。消息、成员、reaction 与 callback 的现有调用签名保持原样。
- 使用既有协议字段：Event.target = 8、Event.chat_type = 15、Event.chat_title = 16；SDK 将非空 ChatTitle 编为 optional string，空串编为 absent。交付仍调用 Core.Deliver(ctx, event, cursor)。
- 测试复用 telegramtest.NewServer(t, token)、server.Push(...json.RawMessage)、server.Calls(method string) int；模拟服务的公开 API 保持现状。

行为：
- 在 baseEvent 按 chat.type 统一赋值。group、supergroup、channel 的 ChatTitle 取 chat.title 原文；private 按 first_name、last_name 的顺序，以一个 ASCII 空格连接存在的非空项，两项都为空时 ChatTitle 留空。私聊只使用 Chat 的姓名字段。
- ChatType 按 private → ChatPrivate、group/supergroup → ChatGroup、channel → ChatChannel；未知类型保留 ChatUnspecified 与空名称。复用已存在的同义映射，若基线已有该赋值则保留一个实现。
- message、channel_post、edited_message、edited_channel_post、chat_member、my_chat_member、message_reaction 与带 Message 的 callback_query 都从各自所在 Chat 取得名称。sender_chat 下 Actor 的既有处理继续生效。缺失名称时交付事件仍正常进行。
- 带话题的消息与回调仍按 is_topic_message 填写 thread_id；名称属于整个聊天，取所在 Chat。没有 target 的事件保持空名称；inline callback 的现有跳过行为保持原样。
- 名称来自现有上报；获取名称额外产生的 Bot API 请求数为零。原有权限查询、callback 确认、long polling、游标与错误处理照常执行。
- ChatTitle、私聊姓名和原始 Chat 对象均不写入 Adapter 日志；原始上报继续完整保存在 PlatformPayload。

测试（按编码规范「测试提交规则」）：
- 提交：在 adapters/telegram/telegram_test.go 新增一个 TestChatTitle 集成测试。经 telegramtest.Server 的真实 HTTP long polling 与 New().Start、Core.Deliver 回调，依次注入一条有 title 的超级群消息、一条 Chat 中同时有 first_name/last_name 的私聊消息、一条 Chat 无姓名但 from 有姓名的私聊消息。断言 ChatTitle 依次为完整群名、两项姓名以单空格连接、空串，ChatType 分别为 ChatGroup、ChatPrivate、ChatPrivate；既有 Session、消息内容、来源和逐条游标继续正确。用 server.Calls("getChat") == 0 及既有请求记录核对取名未增加平台查询。理由：覆盖平台 Chat 解码、共享事件构造与实际接收交付的主路径。
- 既有测试同步：如果完整 Event 比较受新增字段影响，只更新该字段的预期；原有权限、to_me、Actor、话题、回执与游标断言继续执行。
- 本地自测（adapters/telegram/*_local_test.go，PR 写明命令、平台、结果）：普通群与频道取 title；仅 first_name、仅 last_name、Unicode 与名称原文保留；群无 title 时为空；修改消息、成员变化、reaction 和带 Message 的 callback 都取所在 Chat；callback 的 from 与 Chat 私聊姓名不同、sender_chat、同一 update 的多条 reaction、论坛话题、inline callback 无交付；日志中无名称与凭证。以既有模拟服务和录制 Core 驱动，直接复用原辅助函数。
- D-86：提交范围为上述平台集成测试；SDK 字段编解码、Core 目录存储与查询的已有证据复用 PR #166。本切片的本地自测也集中在 adapters/telegram/，无需新增 SDK 测试或 testdata 插件。
- 真实 Telegram：未运行（待维护者提供环境）。将模拟证据与真实平台记录分别填写，执行者把 V-117 Telegram 分项的测试位置、命令、OS/架构、Go 版本、结果及未运行项交给审查者汇总到 TEST-MATRIX §3.2。

完成条件：上述一个集成测试与列明的本地自测取得结果；所有已支持的带 target 路径遵循相同名称规则，取名称新增平台请求为零；gofmt -l 本次修改的 Go 文件无输出；go vet ./adapters/telegram/...；go test -race ./adapters/telegram/...；go build ./...。对照 T-03「失败停止条件」，出现名称或凭证泄漏、重复接收 owner、游标重放、实例间干扰时先修复；需要范围外生产文件、proto、新依赖或改变已确认语义时标 [!] 交审查者处理。V-117 全项由跨任务汇总关闭，本切片交付其 Telegram 分项。

不做：Core 聊天目录、schema 或管理 API 修改；SDK 的名称读取方法；话题名称、getChat 查询与主动拉取聊天；新增服务消息或扩展事件（goal-137）；媒体、发送、开发入口及日志管理的实现。
```

