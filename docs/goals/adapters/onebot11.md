# goal 语句：适配器 · OneBot 11

本文件收录「适配器 · OneBot 11」章节的 goal 语句，切片表在[OneBot 11 实施计划](../../../spec/plan/adapters/onebot11.md)。用法、状态与所有 goal 的共同约束见 [README](../README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-78 | T-15.1 | 官方 OneBot 11 Adapter 的连接与接收 | T-03.3、T-03.4、T-04.6 | 等待前置 |
| goal-79 | T-15.2 | 官方 OneBot 11 Adapter 的发送、媒体与 extension | T-03.6、T-03.8、T-15.1 | 等待前置 |
| goal-80 | T-15.3 | 跨 Adapter 一致性验收 | T-03.10、T-15.2 | 等待前置 |

## 第 8 波

### goal-78 · T-15.1 官方 OneBot 11 Adapter 的连接与接收

状态：等待前置（T-03.3、T-03.4、T-04.6）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。T-03.3、T-03.4 已由 PR #32、#39 合并；已核对 `sdk/adapter` 的 Adapter、Instance、Core、StartRequest、Account 与 `sdk.RunAdapter(adapter.Adapter) error`。`MessageCreatedPayload.to_me` 为字段 3，`Actor.locale` 为 optional 字段 5，OneBot 本切片直接使用已生成类型，无需修改 proto。

```text
/goal 使用 $alonabot-execute 推进 goal-78（切片 T-15.1：官方 OneBot 11 Adapter 的连接与接收）。goal 目录 goal-78，分支 goal-78，worktree .worktrees/goal-78，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现官方 OneBot 11 Adapter 的连接与接收：正向 WebSocket 连接与 access_token、get_login_info 账号、按 echo 配对的 API 调用、有界退避重连、self_id 校验、数组格式检查、实例内有界事件队列与逐条交付、私聊/群聊/撤回/成员事件的标准化与 to_me；提供本地 OneBot 11 模拟服务，供本切片与后续 OneBot 11 切片测试。

依据：PLAN T-15 步骤 1、2；CONTRACTS §8（OneBot 11 段）、§14.5；MESSAGE-MODEL §1、§1.1、§7、§7.1；DECISIONS D-46；TEST-MATRIX V-71 的接收分项；OneBot 11 标准原文（https://github.com/botuniverse/onebot-11 的 communication/ws.md、event/、api/public.md、message/segment.md），字段名以原文为准。

范围（只新增或修改这些）：
- go.mod、go.sum：引入 github.com/coder/websocket（本切片获授权引入，只在 adapters/onebot11 使用，版本取当前最新的正式版）
- 新增 adapters/onebot11/（可导入的包 onebot11）
- 新增 adapters/onebot11/onebottest/（本地 OneBot 11 模拟服务）
- 新增 adapters/onebot11/cmd/alonabot-adapter-onebot11/main.go（只调用 sdk.RunAdapter(onebot11.New())；返回错误时与 examples/plugins/echo 一样 slog.Error 后 os.Exit(1)）
- 上述包的测试文件
允许的新跨模块边：adapters/... -> sdk/...、-> contracts（ARCHITECTURE 已登记）。adapters 不 import internal/...，测试也一样。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/onebot11
   const AdapterID = "onebot11"
   func New() adapter.Adapter // 实现 sdk/adapter 的 Adapter；Descriptor 的 adapter_id 为 AdapterID
   实例配置 JSON：{"ws_url": string（必填，ws:// 或 wss://）, "access_token": string（可选）, "mode": string（缺省 "forward"，只接受 "forward"）}
2. adapters/onebot11/onebottest
   type Call struct { Action string; Params map[string]any }
   type Server struct{ ... }
   func NewServer(t testing.TB, accessToken string) *Server // httptest 服务，在 "/" 提供 OneBot 11 正向 WebSocket 通用端点；accessToken 非空时校验 Authorization: Bearer，不符返回 401；t.Cleanup 时关闭
   func (s *Server) URL() string                       // ws:// 地址，作为 ws_url
   func (s *Server) SetLogin(userID int64, nickname string)
   func (s *Server) Push(events ...json.RawMessage)    // 按顺序推送上报；没有连接时排队，连接建立后推送
   func (s *Server) Disconnect()                       // 关闭当前连接，模拟断线
   func (s *Server) Connections() int                  // 累计接受的连接数
   func (s *Server) Calls() []Call                     // 依次记录每个 API 请求
   func (s *Server) Respond(action string, fn func(params map[string]any) (status string, retcode int, data any)) // 替换某个 action 的响应；status 为空字符串时不响应
   缺省响应：get_login_info 返回 SetLogin 的值；get_msg 在 Push 过的 message 上报中按 message_id 查找并返回其 sender，找不到返回 failed；其他 action 返回 status "failed"、retcode 1404。

行为：
- Start（sdk/adapter 的 Adapter.Start）：
  1) 解析实例配置：ws_url 不是 ws/wss URL、mode 不是 "forward" → INVALID_ARGUMENT（错误不含配置原文与 access_token）。
  2) 用 websocket.Dial 连接 ws_url，access_token 非空时带 Authorization: Bearer；握手 401/403 → PERMISSION_DENIED，其他失败 → UNAVAILABLE。Start 失败不重试。
  3) get_login_info：account_id 为 user_id 的十进制字符串，DisplayName 为 nickname，Username 留空（OneBot 11 没有可寻址的账号用户名，命令不接受 @ 形式，D-67）；失败 → UNAVAILABLE。
  4) 启动读循环与交付循环后返回 Instance：Account() 返回 get_login_info 的账号；Capabilities 本切片只声明接收相关的能力，发送能力由 T-15.2 补充；Send 本切片返回 UNSUPPORTED 的 ErrorInfo（发送由 T-15.2 实现）。
- 交付给 Core 的事件：Event 的 adapter_id、adapter_instance_id（StartRequest.InstanceID）与 account_id（Account.ID）必须与 Target 中的同名字段一致，否则 Core 以 INVALID_ARGUMENT 拒绝；event_id、source_kind 与 capabilities 由 Core 重建，Adapter 不填；Actor.trusted_roles 只保留 CHAT_ADMIN。 occurred_at_unix_ms 取上报的 time（秒）乘以 1000。
- API 调用：每个请求带实例内唯一的 echo（递增的十进制字符串），响应按 echo 配对；调用以 ctx 为期限，ctx 结束时返回 ctx 的错误；连接断开时全部在途调用以 UNAVAILABLE 结束。带 echo 的帧是响应，其余帧是上报。
- 读循环一直读取，不因等待交付而停止：
  - 响应交给配对的调用；meta_event 不入队（lifecycle 的 connect 只记 DEBUG）。
  - 其余上报放进实例事件队列（256 条，命名常量）；队列满时丢弃该上报，累加丢弃计数并记 WARN（带累计数）。
- 交付循环按顺序取出上报，标准化后经 Core.Deliver 交付，游标为空；收到应答（无论 accepted）后才取下一条。不交付的上报直接跳过。Deliver 返回错误时：实例停止后为 FAILED_PRECONDITION，交付循环退出；其他错误记 WARN 后继续下一条（OneBot 11 没有重放，与队列满时丢弃一致）。
- 格式与身份：message 上报的 message 字段是字符串（CQ 码格式）时，经 Core.ReportFailure 报告（safe_message "onebot implementation must post messages in array format"），关闭连接并停止接收。self_id 与 account_id 不符的上报丢弃并记 WARN。
- 断线：连接关闭或读错误（Stop 引起的除外）后按有上限的指数退避重连（初始 1s、上限 60s，命名常量），重连后重新 get_login_info；user_id 变化 → ReportFailure（"onebot account changed"）并停止；握手 401/403 → ReportFailure（"onebot implementation rejected the access token"）并停止。断线期间的上报不补收；重连与已提交的 Send 无关，不重放任何请求。
- Stop：关闭连接，等待读循环与交付循环退出后返回；Stop 之后不再交付。同一实例不会有两条并行的连接。
- 标准化（MESSAGE-MODEL §7.1；数字 ID 一律转为十进制字符串）：
  - Target：adapter_id、adapter_instance_id、account_id、chat_id（"private:<user_id>" 或 "group:<group_id>"），没有 thread_id。
  - message（private、group）→ message.created：platform_message_id 为 message_id，source_event_id 同 message_id；Actor.external_user_id 为 user_id，display_name 取非空的 sender.card，否则 sender.nickname，locale absent；群聊中 sender.role 为 owner 或 admin 时 trusted_roles 含 CHAT_ADMIN；sub_type 为 anonymous 时不构造 Actor。
  - chain：text → plain text；at → mention（qq 为 "all" 时不生成）；reply → reference；image、video 本切片不生成（T-15.2 接入入站媒体）；其他段不生成。platform_payload 为原始上报 JSON，chain 可以为空。
  - to_me：私聊；chain 中有 external_user_id 等于 account_id 的 mention；reply 引用的消息经 get_msg 查得 sender.user_id 等于 account_id。get_msg 失败按 false 处理并记 DEBUG。
  - notice.group_recall、notice.friend_recall → message.deleted（platform_message_ids 只含 message_id；operator_id 存在时作为 actor）；notice.group_increase、notice.group_decrease → member.joined、member.left（member_external_user_id 为 user_id；operator_id 非 0 且不等于 user_id 时作为 actor）；其他 notice 与全部 request 不交付，计数并记 DEBUG。
- 脱敏：access_token 不进入日志、错误与普通响应；ws_url 带 userinfo 或查询参数时，记录前去掉这两部分。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/onebot11/onebot11_test.go。用 onebottest.Server 与一个记录调用的 adapter.Core 测试替身驱动 onebot11.New() 的实例（不经 Core 进程）：
  - Start 得到 get_login_info 的 account_id 与 nickname；access_token 错误时 Start 返回 PERMISSION_DENIED。
  - 推入私聊消息、群聊消息（sender.role 为 admin）、匿名群消息、撤回、入群与一条 request：依次交付 5 个事件，字段与 MESSAGE-MODEL §7.1 一致，request 未交付；Core 替身对第二条返回 accepted=false 时后续照样交付。
  - to_me：私聊为 true；@本账号为 true；reply 本账号的消息为 true；reply 他人消息与 get_msg 失败为 false。
  - 字符串格式的 message 上报产生一次 ReportFailure，接收停止；self_id 不符的上报未交付。
  - Disconnect 后在退避期内重连（Connections 为 2），之后推入的消息正常交付，没有重复交付；重连后 get_login_info 返回其他 user_id 时 ReportFailure。
  - Core 替身的 Deliver 阻塞期间推入 300 条上报：读循环仍能完成一次 get_msg 调用，有上报被丢弃且丢弃计数与未交付数一致；解除阻塞后按原顺序交付队列中的事件。
  - 错误与日志中不出现 access_token（用测试 token 字符串做断言）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：配置的各种非法值；echo 配对与并发调用；ctx 超时与断线时的在途调用；退避上限；Stop 期间读循环退出；ws_url 的脱敏。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./adapters/...；go build ./...；go mod tidy 后 go.mod 与 go.sum 无差异。

不做：发送、发送计划与回执、入站与出站媒体、extension（T-15.2）；Core 侧装配、alonabot run 联验与一致性验收（T-15.3）；反向 WebSocket（v0.1 之外）。
```


## 第 9 波

### goal-79 · T-15.2 官方 OneBot 11 Adapter 的发送、媒体与 extension

状态：等待前置（T-03.6、T-03.8、T-15.1）。语句中 goal-7、goal-8、goal-78 提供的符号（sdk/adapter 的 Core.ReadMedia 与实例可选实现的平台文件读取接口、ExtensionDeclaration 的校验、adapters/onebot11 的实例与 API 调用、onebottest.Server）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-79（切片 T-15.2：官方 OneBot 11 Adapter 的发送、媒体与 extension）。goal 目录 goal-79，分支 goal-79，worktree .worktrees/goal-79，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：OneBot 11 实例的 Send 先对整条 Chain 做完整预检并生成确定的物理操作计划，再依次调用 send_msg；连续的 text、mention、image 与 extension 合成一条消息，video 单独成一条，reference 只加在第一条前；每条消息一个 PartResult，按 MESSAGE-MODEL §4 汇总。入站 image、video 生成带 incoming_ref 的片段并实现平台文件读取；出站媒体经 Core.ReadMedia 读出后以 base64 发送；注册 onebot11.face extension。

依据：PLAN T-15 步骤 3、4；CONTRACTS §8（Send、Chain 与计划段，extension 段）、§1.4；MESSAGE-MODEL §2、§4、§7.1 的入站媒体、§7.2；DECISIONS D-46；TEST-MATRIX V-71 的发送与媒体分项。

范围（只新增或修改这些）：
- adapters/onebot11/：新增 plan.go、send.go、media.go；实例的 Send、Capabilities 与 Descriptor 的 extensions 补全；标准化补上入站 image、video
- adapters/onebot11/onebottest/：增加 send_msg 的缺省响应与故障注入
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/onebot11：实例实现 sdk/adapter 的 Instance.Send 与 goal-7 定义的平台文件读取接口；不新增导出符号。
2. adapters/onebot11/onebottest（在 goal-78 的 Server 上追加）
   func (s *Server) DisconnectOn(action string, nth int) // 收到第 nth 次该 action 后不响应并关闭连接（请求已写出、结果未知）
   send_msg 的缺省响应：status "ok"，data.message_id 从 1001 起递增。

行为：
- 能力与 extension：Capabilities 声明 message.send、message.reference（version 1，supported=true）；不声明 message.markdown、message.album、message.buttons、message.edit。Descriptor 的 extensions 含 {name "face", version 1, schema_json 为只有必需字符串属性 "id"（pattern ^[0-9]+$）、不允许其他属性的 JSON Schema}；片段的 adapter_id 为 AdapterID。
- 预检（任何 send_msg 之前，任一失败都返回 REJECTED 的 SendResult 加 ErrorInfo，part_results 为空）：
  - 片段只支持 plain 且无 entities 的 text、image、video、mention、reference 与本 Adapter 的 face extension；带 markdown 或 entities 的 text、file、buttons、forward 与其他 extension → UNSUPPORTED。
  - target.chat_id 必须是 "private:<十进制>" 或 "group:<十进制>"，否则 INVALID_ARGUMENT；thread_id 存在 → UNSUPPORTED。mention 的 external_user_id、reference 的 original_platform_message_id 必须是十进制，否则 INVALID_ARGUMENT。
  - 媒体来源：https_url 只接受 https，原样作为段的 file；media_ref 与 incoming_ref 由 Core 换成的只读 token 在执行该部分时经 Core.ReadMedia 读完，编码为 "base64://..."。
- 物理计划：按原顺序扫描，连续的 text、mention、image、face 合成一条消息的段数组（text → text，mention → at{qq}，image → image{file}，face → face{id}）；video 单独成一条（video{file}）；reference 转为第一条消息最前面的 reply{id}。每条消息一个 send_msg（private 用 user_id，group 用 group_id，message 为段数组），对应一项 PartResult（part_index 从 0 起，segment_indexes 为它覆盖的片段，reference 计入第一项）。
- 执行与回执：按计划顺序执行。status ok → ACCEPTED，message_id 的十进制进入 platform_message_ids；status async → ACCEPTED 且 message_ids_unavailable=true；status failed → REJECTED，ErrorInfo 为 FAILED_PRECONDITION，safe_message 带 retcode；请求写出后连接断开或到 ctx 期限仍无响应 → UNKNOWN；发出前连接不可用 → 该部分 REJECTED，ErrorInfo 为 UNAVAILABLE。任一部分不是 ACCEPTED 后，其余部分 NOT_ATTEMPTED。整体状态与 primary_platform_message_id 按 MESSAGE-MODEL §4 汇总，部分成功时同时返回 ErrorInfo。不自动重试。
- 入站媒体：image、video 段生成 image、video 片段，incoming_ref.token 为段的 url（没有 url 时为 file，并在读取时返回 UNSUPPORTED）；平台文件读取接口首次被调用时经 HTTP GET 下载 url（只允许 http、https；响应非 2xx → UNAVAILABLE），按 offset 返回；下载使用实例自己的 http.Client 与 http.ProxyFromEnvironment。
- 脱敏沿用 goal-78：access_token 不进入日志与错误。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/onebot11/send_test.go。用 onebottest.Server 与 Core 测试替身（ReadMedia 返回内存字节）驱动真实实例：
  - reference+text+mention+image(media_ref)+face+video(https_url)+text 的计划为三次 send_msg：[reply, text, at, image(base64://), face]、[video(https URL)]、[text]，参数与顺序正确，三项 PartResult 的 segment_indexes 正确。
  - 第二次 send_msg 返回 failed：第一项 ACCEPTED、第二项 REJECTED、第三项 NOT_ATTEMPTED，整体 PARTIAL 并带 ErrorInfo；DisconnectOn 第一次 send_msg：整体 UNKNOWN，其余 NOT_ATTEMPTED；async 响应标记 message_ids_unavailable 且 platform_message_ids 为空。
  - 预检失败（buttons、file、forward、markdown、thread_id、非法 chat_id、其他 Adapter 的 extension）时 Calls() 中没有 send_msg。
  - adapters/onebot11/media_test.go：推入带 image 的群消息（url 指向测试 HTTP 服务），片段的 incoming_ref.token 为该 url；平台文件读取返回正确内容；url 为 file:// 时读取返回 UNSUPPORTED。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：face 的 Schema 校验正反例；只含 reference 的 Chain 被拒；大媒体的 base64 编码与分块读取；ctx 期限到达时的 UNKNOWN。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./adapters/...；go build ./...。

不做：Core 侧装配、alonabot run 联验与一致性验收（T-15.3）；消息编辑（OneBot 11 不支持，由 Core 按未声明的 message.edit 拒绝）；合并转发与文件上传（v0.1 不支持）。
```


## 第 10 波

### goal-80 · T-15.3 跨 Adapter 一致性验收

状态：等待前置（T-03.10、T-15.2）。语句中 goal-9、goal-78、goal-79 提供的符号（alonabot run 的 --adapter-id/--adapter-entry、cmd/alonabot/telegram_e2e_test.go 的装配方式、telegramtest.Server、onebottest.Server）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-80（切片 T-15.3：跨 Adapter 一致性验收）。goal 目录 goal-80，分支 goal-80，worktree .worktrees/goal-80，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：用同一个插件二进制，分别经 Telegram 与 OneBot 11 两个模拟服务收发同一组 Chain 与入站消息，对照两边的结果，证明通用片段在两边语义一致、每个实例只用自己声明的能力、未声明的能力在首个平台请求前被拒绝；经 alonabot run 跑 OneBot 11 的端到端链路；在 TEST-MATRIX §3.2 记录 V-71、V-72 的证据。真实 QQ 端只有维护者的手测记录，没有时记为未运行。

依据：PLAN T-15 步骤 5 与「失败停止条件」；CONTRACTS §8；MESSAGE-MODEL §2、§4、§6.2、§7；DECISIONS D-46；TEST-MATRIX V-71、V-72。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/testdata/conformance/（测试插件：收到 "case:<name>" 时发送下表中该用例的 Chain，并以一行 INFO 记录 SendResult 的摘要；收到其他消息时以一行 INFO 记录入站 chain 的片段类型序列与 to_me）
- 新增 cmd/alonabot/conformance_e2e_test.go（//go:build unix）
- docs/development.md：在 alonabot run 一节补充以 OneBot 11 Adapter 运行的示例命令与实例文件（本切片获授权修改这一处）
- spec/TEST-MATRIX.md §3.2：追加 T-15 的证据行（本切片获授权修改这一处；其他 spec 文件不改）
允许的新跨模块边：cmd/alonabot 的测试 -> adapters/telegram/telegramtest、-> adapters/onebot11/onebottest（测试代码，审查者确认允许）。

接口：不新增公开 Go 符号，不修改 alonabot run 的参数。测试插件的日志摘要格式自定，但须包含 status、每项 PartResult 的 status 与 segment_indexes、ErrorInfo 的 code。

行为与测试：
- conformance_e2e_test.go 构建 alonabot 与两个 Adapter 程序和测试插件，对每个 Adapter 各起一次 alonabot run（各自的临时 data_dir 与实例文件，同一个插件二进制）；经模拟服务注入 "case:<name>" 消息，从 Core 日志读出摘要，从模拟服务的 Calls() 读出实际平台请求，再按下表断言：
  | 用例 | Telegram | OneBot 11 |
  | text | ACCEPTED，1 项 | ACCEPTED，1 项 |
  | mention+text | ACCEPTED（text_mention） | ACCEPTED（at 段） |
  | reference+text | ACCEPTED，reply_parameters 只在第一个请求 | ACCEPTED，reply 段只在第一条消息最前 |
  | text+image+text | ACCEPTED，3 项 | ACCEPTED，1 项 |
  | text+buttons | ACCEPTED | REJECTED/UNSUPPORTED，send_msg 为 0 次 |
  | file | ACCEPTED | REJECTED/UNSUPPORTED，send_msg 为 0 次 |
  | markdown text | ACCEPTED | REJECTED/UNSUPPORTED，send_msg 为 0 次 |
  | forward | REJECTED/UNSUPPORTED，Bot API 请求为 0 次 | REJECTED/UNSUPPORTED，send_msg 为 0 次 |
  | onebot11.face extension | REJECTED/UNSUPPORTED，由 Core 拒绝，Bot API 请求为 0 次 | ACCEPTED |
  | Edit 文本 | ACCEPTED（editMessageText） | UNSUPPORTED，由 Core 按 message.edit 拒绝，send_msg 为 0 次 |
  两边都 ACCEPTED 的用例另外断言：每个片段恰好出现在一项 PartResult 的 segment_indexes 中，按项的顺序展开等于原片段顺序；platform_message_ids 的数量等于物理请求数。
- 入站对照：两边各注入一组语义相同的消息（纯文本；@本机器人加文本；回复本机器人消息；带图片的消息），测试插件记录的片段类型序列与 to_me 在两边相同。
- OneBot 11 的端到端链路（同一测试文件的子测试）：实例启动后出现 bot.online；模拟服务 Disconnect 后实例自动重连，之后的消息照常收发，没有 bot.offline；access_token 错误时实例失败，出现 reason=INSTANCE_FAILED 的 bot.offline；Core 日志与 data_dir 中的日志文件不含 access_token。
- 证据：TEST-MATRIX §3.2 追加一行，写明上表的实测结果、命令、平台与 Go 版本；真实 QQ 端：维护者提供手测记录（实现端名称与版本、通过的用例）时写入，否则写「未运行」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./cmd/alonabot/... ./adapters/...；go build ./...。

不做：修改两个 Adapter 的映射或 Core 的派发（发现不一致时标 [!] 并在 PR 中写明用例与差异，由审查者决定修契约还是修实现）；alonabot run 同时启动多个 Adapter；反向 WebSocket。
```
