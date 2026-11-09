# goal 语句：适配器 · Telegram

本文件收录「适配器 · Telegram」章节的 goal 语句，切片表在[Telegram 实施计划](../../../spec/plan/adapters/telegram.md)。用法、状态与所有 goal 的共同约束见 [README](../README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 已完成（PR #39） |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 已完成 |
| goal-89 | T-03.11 | 官方 Telegram Adapter 的平台权限 | T-04.6 | 已完成（PR #106） |
| goal-111 | T-03.20 | 官方 Telegram Adapter 的媒体与新增片段 | T-03.13、T-03.14 | 等待前置 |
| goal-112 | T-03.21 | 官方 Telegram Adapter 的平台动作与撤回 | T-03.11、T-03.15、T-03.16 | 等待前置 |
| goal-113 | T-03.22 | 官方 Telegram Adapter 的命令菜单 | T-03.17 | 等待前置 |
| goal-117 | T-03.23 | 官方 Telegram Adapter 的代理模式与话题判定 | T-03.4 | 待派发 |

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

状态：待派发。前置 T-03.4（[PR #39](https://github.com/AlonaBot/AlonaBot/pull/39)，合并提交 `604dc65`）已合并。2026-11-09 按 D-100、D-101 新增，接口与测试段按 main（`3f7a380`）的代码写出：`newBotClient(token, base, proxy string)` 在 proxy 为空时使用 `http.ProxyFromEnvironment`；`instance.messageEvent` 只要 Message 带 `message_thread_id` 就写入 Target 的 thread_id；`TestInstanceProxy` 与 `TestIncomingMedia` 只配置了 `proxy`。实例配置的 `instance_schema.json` 由 goal-101（T-08.12）新建，处理方式见执行顺序 §3。

```text
/goal 使用 $alonabot-execute 推进 goal-117（切片 T-03.23：官方 Telegram Adapter 的代理模式与话题判定）。goal 目录 goal-117，分支 goal-117，worktree .worktrees/goal-117，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 实例配置追加 proxy_mode，代理可以不使用（缺省）、跟随 Core 或使用实例自己的 proxy；入站 Message 只在 is_topic_message 为 true 时把 message_thread_id 作为 Target 的 thread_id，非论坛超级群的回复串不再被当作话题。

依据：PLAN T-03 步骤 20；CONTRACTS §8（Telegram 段）、§15.1（proxy）；MESSAGE-MODEL §6、§6.1；DECISIONS D-100、D-101；TEST-MATRIX V-99、V-100；Bot API 原文 Message 的 message_thread_id 与 is_topic_message。

范围（只新增或修改这些）：
- adapters/telegram/telegram.go：实例配置的解析
- adapters/telegram/client.go：HTTP 客户端的代理选择
- adapters/telegram/inbound.go：telegramMessage 追加 is_topic_message，thread_id 的判定
- adapters/telegram/telegram_test.go、adapters/telegram/media_test.go：既有用例的配置与本语句的测试
- adapters/telegram/cmd/alonabot-adapter-telegram/instance_schema.json：只在派发时 main 已有该文件（goal-101 已合并）时修改；没有时不新建，在 PR 正文写明，由后合并的 goal-101 按执行顺序 §3 补上
没有新的跨模块边。不改 proto，不改 telegramtest。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 实例配置 JSON：{"token": string（必填）, "api_base_url": string（缺省 "https://api.telegram.org"）, "proxy_mode": "none" | "core" | "custom"（缺省 "none"）, "proxy": string（只在 custom 时使用，http/https/socks5 URL，可含用户名与密码）}
2. instance_schema.json（文件存在时）：在 proxy 之前追加
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

状态：等待前置（T-03.13、T-03.14）。2026-11-09 按 D-87、D-88、D-89、D-97、D-99 新增，同日按 D-100 追加入站事件的 chat_type。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。

目标：Telegram 实例按来源传递出站媒体：multipart 请求体边读边写、inline_base64 解码上传、`platform_file` 直接作为 file_id；收发 voice（sendVoice）；动图用 sendAnimation；入站 voice、animation、video_note 与静态贴纸按新映射交付（入站 audio 仍为 file）；url 按钮；实例配置追加 `file_base_url`；入站事件填写 chat_type。

依据：PLAN T-03 步骤 17；CONTRACTS §1.3、§8（Telegram 段）；MESSAGE-MODEL §6、§6.2、§6.3、§6.4；DECISIONS D-88、D-89、D-97、D-99、D-100；TEST-MATRIX V-89、V-90、V-98 的 Telegram 分项，V-97。

要点：
- 流式 multipart：用 `io.Pipe` 与 `mime/multipart`，写端按发送计划顺序写字段与文件，文件内容经 Core.ReadMedia 每次最多 1 MiB 读取后立即写出；读取失败时以该错误关闭管道，该物理操作按请求是否已开始写出判为 REJECTED 或 UNKNOWN。现有一次读完的实现删除，不保留。
- 计划：voice 每个片段单独一个物理操作；image/gif（MediaRef 的 MIME，或 inline_base64 解码后以 `GIF8` 开头）用 sendAnimation 并结束当前媒体组；caption 与 reference、按钮的规则不变；forward 返回 UNSUPPORTED。
- 入站：voice 为 voice，audio 仍为 file，animation 与 video_note 为 video，is_animated 与 is_video 都为 false 的 sticker 为 image，都带 incoming_ref；其余贴纸不进 chain。
- `file_base_url`：缺省为 `<api_base_url>/file/bot<token>`，写进 instance_schema.json；文件下载经它发出，代理规则不变，token 仍不进日志。
- chat_type：每个带 target 的入站事件按 chat.type 填写，private 为 PRIVATE，group 与 supergroup 为 GROUP，channel 为 CHANNEL；成员与 reaction 事件同样取其 chat。
- Capabilities 追加 `message.voice`；`telegramtest` 追加 sendVoice、sendAnimation 的记录与缺省响应。
- 测试：提交在既有 `send_test.go` 中追加的用例（经 telegramtest.Server）：一次发送 32 MiB 的 media_ref，断言请求体正确且常驻内存不超过数 MiB，以及 voice 与 GIF 的计划；其余进本地自测。

### goal-112 · T-03.21 官方 Telegram Adapter 的平台动作与撤回

状态：等待前置（T-03.15、T-03.16）。T-03.11（PR #106）已合并。2026-11-09 按 D-87、D-92、D-93、D-98 新增。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。

目标：Telegram 实例声明并实现 MESSAGE-MODEL §6.5 的平台动作（含群信息与成员查询、send_chat_action、set_message_reaction、send_message_draft），在 `adapters/telegram/action` 为每个动作发布带类型的 Go 函数，并提供输入状态与草稿流式两个辅助函数；实现撤回（deleteMessages），声明 `message.delete`。

依据：PLAN T-03 步骤 18；CONTRACTS §8（通用能力的范围、平台动作、撤回）；MESSAGE-MODEL §4A、§6.5；DECISIONS D-87、D-92、D-93、D-98；TEST-MATRIX V-91、V-92、V-94 的 Telegram 分项；Bot API 原文的 sendMessageDraft。

要点：
- 动作的参数沿用 Bot API 字段名，数字 ID 为十进制字符串，chat_id 与 message_thread_id 取自 Target；Schema 不允许其他属性；结果为 Bot API 的 `result` 原文。ban、unban、restrict 成功后作废该会话的管理员缓存。
- `adapters/telegram/action`：每个动作一个带类型的函数，第一个参数为 goal-106 的调用者接口，参数与结果为 Go 类型；只依赖 `sdk` 根包，不 import Adapter 运行时代码。输入状态辅助函数立即发送一次、之后每 4 秒一次，ctx 结束即停止；草稿辅助函数以 send_message_draft 流式显示，结束后经普通 Send 发出正式消息，群聊目标在发请求前返回 UNSUPPORTED。sendMessageDraft 的参数与本语句不符时标 [!]。
- 撤回：同一 Target 的 ID 每 100 个一次 deleteMessages，同一请求的 ID 共享结果；平台返回 400 时该请求的 ID 为 REJECTED，其余 NOT_ATTEMPTED。
- 测试：提交在既有测试文件中追加的用例（经 telegramtest.Server）：ban_chat_member 的参数与管理员缓存作废，以及 101 个 ID 的撤回分为两次请求；其余进本地自测。

### goal-113 · T-03.22 官方 Telegram Adapter 的命令菜单

状态：等待前置（T-03.17）。2026-11-09 按 D-87、D-96 新增。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。

目标：Telegram 实例实现 `adapter.commands.sync`（setMyCommands，default scope 加每个 language_code），声明 `commands.sync`。

依据：PLAN T-03 步骤 19；CONTRACTS §8（命令菜单）；MESSAGE-MODEL §6.5；DECISIONS D-96；TEST-MATRIX V-93 的 Telegram 分项；Bot API 原文的 setMyCommands。

要点：
- 命令名不匹配 `^[a-z0-9_]{1,32}$` 时跳过，说明超过 256 字符时截断，都记诊断；记住每个 language_code 上一次成功同步的内容，相同时不调用；某个语言失败不影响其他语言，返回第一个错误。
- 测试：提交在既有测试文件中追加的用例（经 telegramtest.Server）：两次相同的同步只调用一次 setMyCommands；其余进本地自测。
