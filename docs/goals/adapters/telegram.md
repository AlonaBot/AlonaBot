# goal 语句：适配器 · Telegram

本文件收录「适配器 · Telegram」章节的 goal 语句，切片表在[Telegram 实施计划](../../../spec/plan/adapters/telegram.md)。用法、状态与所有 goal 的共同约束见 [README](../README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 已完成（PR #39） |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 等待前置 |
| goal-89 | T-03.11 | 官方 Telegram Adapter 的平台权限 | T-04.6 | 等待前置 |

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

状态：等待前置。T-03.4 已由 PR #39 合并，已核对 `telegram.New() adapter.Adapter`、私有 `instance` 与 `botClient.call(ctx context.Context, method string, request any, result any) error`、`telegramtest.NewServer(t testing.TB, token string) *Server`，以及 `Instance.Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)`。媒体发送依赖本语句中的 `Core.ReadMedia`，该接口由 goal-7 / T-03.6 追加，当前 `sdk/adapter.Core` 尚未提供；待 T-03.6 合并并核对 ReadMedia 后再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-50（切片 T-03.7：官方 Telegram Adapter 的发送计划与回执）。goal 目录 goal-50，分支 goal-50，worktree .worktrees/goal-50，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Telegram 实例的 Send 先对整条 Chain 做完整预检并生成确定的物理操作计划，再依次调用 Bot API；文本实体按 UTF-16 渲染，连续图片/视频组成媒体组，引用作用于第一条物理消息，按钮作用于最后一条；每个物理操作一个 PartResult，汇总为 ACCEPTED、PARTIAL、REJECTED 或 UNKNOWN，已发生的结果不丢失。

依据：PLAN T-03 步骤 6（发送）；CONTRACTS §8（Send、Chain 与计划段）；MESSAGE-MODEL §2、§4、§6.2；TEST-MATRIX V-08、V-35。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto：Button 追加 optional string core_token = 4（由 Core 填写，T-03.8 接入；插件填写的值由 Core 清除）。随后 go generate ./...（本切片获授权只追加这一字段）
- contracts/wirespec_test.go：追加编号检查
- adapters/telegram/：新增 plan.go、send.go、render.go；实例的 Send 与 Capabilities 补全
- adapters/telegram/telegramtest/：增加 sendMessage、sendPhoto、sendVideo、sendDocument、sendMediaGroup、editMessageReplyMarkup 与故障注入
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/telegram：实例实现 sdk/adapter 的 Instance.Send；不新增导出符号。
2. adapters/telegram/telegramtest（在 goal-5 的 Server 上追加）
   type Call struct { Method string; Params map[string]any } // 每次 Bot API 调用的方法与解析后的参数
   func (s *Server) Calls() []Call
   func (s *Server) FailMethod(method string, nth int, status int, description string) // 第 nth 次调用该方法时返回指定错误
   func (s *Server) ZeroMessageID(method string, nth int)                              // 第 nth 次调用成功但 message_id 为 0
   func (s *Server) Timeout(method string, nth int)                                    // 第 nth 次调用在响应前断开（结果未知）

行为：
- 预检（任何 Bot API 调用之前，任一失败都返回 REJECTED 的 SendResult 加 ErrorInfo，part_results 为空）：
  - 片段类型只支持 text、image、video、file、reference、buttons、mention；forward 与其他类型 → UNSUPPORTED。extension 由 T-03.8 处理，本切片 → UNSUPPORTED。
  - text：format 为 markdown 时按 MarkdownV2 发送且不得同时带 entities；plain 时 entities 按 UTF-16 偏移校验（越界或重叠方式不被 Telegram 接受 → INVALID_ARGUMENT）。单条文本超过 4096 个字符、caption 超过 1024 个字符（以 UTF-16 计，命名常量）→ INVALID_ARGUMENT，不截断。
  - 媒体来源：media_ref 与 incoming_ref 由 Core 换成的只读 token 经 Core.ReadMedia 读取后上传（multipart）；https_url 只接受 https，原样交给 Telegram。
  - reference 只能出现一次且位于第一个物理操作；buttons 只能出现一次且位于末尾；每个按钮必须有 core_token（≤ 64 字节，命名常量），否则 INVALID_ARGUMENT。mention 只能渲染为 text_mention（需要 external_user_id），否则 UNSUPPORTED。
  - target.chat_id 必须是十进制整数；thread_id 存在时作为 message_thread_id。
- 物理计划（MESSAGE-MODEL §6.2）：相邻 text 片段可合并为一条消息，不改变顺序；独立 text 不改为 caption，caption 只取媒体片段自身的 caption；连续 image/video 每组 2～10 项用 sendMediaGroup，单项用 sendPhoto/sendVideo，file 用 sendDocument，不放进媒体组；文本穿插时结束当前媒体组。reference 转为第一条物理请求的 reply_parameters；buttons 转为最后一条物理消息的 reply_markup（inline_keyboard，callback_data 为 core_token），最后一步是 sendMediaGroup 时追加一次 editMessageReplyMarkup 作用于该组已确认的最后一条消息。
- 执行与回执：按计划顺序执行；每个物理操作一个 PartResult（part_index 从 0 起，segment_indexes 为它覆盖的片段）。成功时记录 message_id；message_id 为 0 时 message_ids_unavailable=true，0 不进入 platform_message_ids，依赖该 ID 的 markup 步骤为 NOT_ATTEMPTED。某一步被拒（4xx）后，后续步骤全部 NOT_ATTEMPTED；网络断开或超时使该步为 UNKNOWN，后续 NOT_ATTEMPTED。整体：全部 ACCEPTED → ACCEPTED；第一步就被拒 → REJECTED；第一步就未知 → UNKNOWN；其余 → PARTIAL，并同时返回 ErrorInfo。primary_platform_message_id 为第一条确认的消息 ID。发送不自动重试，429 也直接作为该步的错误返回（safe_message 带 retry_after）。
- Capabilities：本切片起声明 message.send、message.markdown、message.album、message.reference、message.buttons（version 1，supported=true）。
- 脱敏沿用 goal-5：任何错误与日志不含 token。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/telegram/send_test.go。用 telegramtest.Server 与 Core 测试替身（ReadMedia 返回内存字节）驱动真实实例：
  - text+image+image+text+file+buttons 的计划为 sendMessage、sendMediaGroup（2 项）、sendMessage、sendDocument（带 reply_markup），参数与顺序正确；reference 只出现在第一个请求。
  - 第二步 FailMethod(400)：第一步 ACCEPTED、第二步 REJECTED、其余 NOT_ATTEMPTED，整体 PARTIAL 并带 ErrorInfo；第一步 Timeout：整体 UNKNOWN。
  - 预检失败（超长文本、forward、缺 core_token 的按钮、markdown 与 entities 同时出现）时 Calls() 为空。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：媒体组 10 项与 11 项的拆分；文本合并规则；caption 长度边界；mention 的渲染；429 的 safe_message；thread_id；multipart 上传的字段；末尾为相册且带 buttons：追加 editMessageReplyMarkup；该步失败时整体 PARTIAL，媒体组的消息 ID 保留在结果中；ZeroMessageID 作用于相册末条：markup 步骤为 NOT_ATTEMPTED，整体 PARTIAL，platform_message_ids 不含 0；含 emoji 与组合字符的文本，entities 的 UTF-16 偏移在请求中正确。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./adapters/... ./contracts/...；go build ./...。

不做：按钮 token 的签发与 callback_query（T-03.8）；extension（T-03.8）；编辑（T-03.9）；入站媒体与 getFile（T-03.6 之后在 T-03.10 补 MediaSource）；BeforeSend/AfterSend Hook（T-05.3）。
```

## 第 8 波

### goal-89 · T-03.11 官方 Telegram Adapter 的平台权限

状态：等待前置（T-04.6）。语句只写了目标、依据和要点。依据 D-69 新增。

目标：Adapter 描述声明 `telegram.chat_owner` 与 `telegram.chat_admin`；群与超级群中的事件按 `getChatAdministrators` 的会话缓存给出 Actor 的平台权限，缓存 10 分钟，`chat_member`、`my_chat_member` 使缓存作废，获取失败时不带平台权限并记诊断；发布只含常量的 `adapters/telegram/perm`。

依据：PLAN T-03 步骤 9；MESSAGE-MODEL §6.1；CONTRACTS §2.3、§8；DECISIONS D-69；TEST-MATRIX V-85。
