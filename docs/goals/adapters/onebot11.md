# goal 语句：适配器 · OneBot 11

本文件收录「适配器 · OneBot 11」章节的 goal 语句，切片表在[OneBot 11 实施计划](../../../spec/plan/adapters/onebot11.md)。用法、状态与所有 goal 的共同约束见 [README](../README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-78 | T-15.1 | 官方 OneBot 11 Adapter 的连接与接收 | T-03.3、T-03.4、T-04.6 | 已完成（PR #109） |
| goal-79 | T-15.2 | 官方 OneBot 11 Adapter 的发送、媒体与 extension | T-03.6、T-03.8、T-04.16、T-15.1 | 等待前置（T-04.16） |
| goal-133 | T-15.7 | 官方 OneBot 11 Adapter 填写 chat_title | T-03.25、T-04.16、T-15.1 | 等待前置（T-04.16） |
| goal-80 | T-15.3 | 跨 Adapter 一致性验收 | T-03.10、T-15.2 | 等待前置 |
| goal-114 | T-15.4 | 官方 OneBot 11 Adapter 的反向 WebSocket | T-03.19、T-04.16、T-15.1、T-15.2 | 等待前置（T-04.16、T-15.2） |
| goal-115 | T-15.5 | 官方 OneBot 11 Adapter 的新增片段与媒体 URL | T-03.13、T-03.14、T-15.2 | 等待前置 |
| goal-116 | T-15.6 | 官方 OneBot 11 Adapter 的扩展事件、平台动作与撤回 | T-03.15、T-03.16、T-15.2 | 等待前置 |

## 第 8 波

### goal-78 · T-15.1 官方 OneBot 11 Adapter 的连接与接收

状态：已完成（[PR #109](https://github.com/AlonaBot/AlonaBot/pull/109)，合并提交 `1b226a2`），证据见 TEST-MATRIX §3.2。派发时：前置 T-03.3（PR #32，合并提交 `db3c1c7`）、T-03.4（PR #39，合并提交 `604dc65`）与 T-04.6（PR #78，合并提交 `0a320c5`）均已合并。2026-11-08 按 D-69 改写：删去 TrustedRole/CHAT_ADMIN 的旧写法，Actor 改为给出平台权限（`Actor.permissions`），描述声明 `onebot11.group_owner`、`onebot11.group_admin`、`onebot11.private_friend` 及隐含关系，新增只含常量的 `adapters/onebot11/perm`；按 D-85 提供 `adapter_config_schema.json`（Core 侧的读取与补值由 T-08.12 实现）。已核对：`sdk/adapter` 的 `Adapter`（`Descriptor`、`Start`）、`Instance`、`Core`（`Deliver(ctx, ev, cursor) (accepted bool, err error)`、`SaveCursor`、`ReportFailure(ctx, *pb.ErrorInfo)`、`ReadMedia`）、`StartRequest`、`Account`；`pb.PermissionDeclaration{name, implies, label_key}` 与 `validate.AdapterRegister` 对权限声明的检查；`sdk.RunAdapter(adapter.Adapter) error`；`MessageCreatedPayload.to_me`、`Actor.locale` 已有，无需修改 proto；`github.com/coder/websocket` 已于 2026-10-27 批准（docs/development.md）。

```text
/goal 使用 $alonabot-execute 推进 goal-78（切片 T-15.1：官方 OneBot 11 Adapter 的连接与接收）。goal 目录 goal-78，分支 goal-78，worktree .worktrees/goal-78，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现官方 OneBot 11 Adapter 的连接与接收：正向 WebSocket 连接与 access_token、get_login_info 账号、按 echo 配对的 API 调用、有界退避重连、self_id 校验、数组格式检查、实例内有界事件队列与逐条交付、私聊/群聊/撤回/成员事件的标准化与 to_me，群主、群管理员与好友私聊的平台权限及常量包；提供本地 OneBot 11 模拟服务，供本切片与后续 OneBot 11 切片测试。

依据：PLAN T-15 步骤 1、2；CONTRACTS §2.3、§8（OneBot 11 段与 Adapter 权限声明）、§14.5；MESSAGE-MODEL §1、§1.1、§7、§7.1；DECISIONS D-46、D-69、D-85；TEST-MATRIX V-71 的接收分项、V-85 的 OneBot 11 分项；OneBot 11 标准原文（https://github.com/botuniverse/onebot-11 的 communication/ws.md、event/、api/public.md、message/segment.md），字段名以原文为准。

范围（只新增或修改这些）：
- go.mod、go.sum：引入 github.com/coder/websocket（已批准，只在 adapters/onebot11 使用，版本取当前最新的正式版，并把锁定的版本写进 docs/development.md 的批准清单该条）
- 新增 adapters/onebot11/（可导入的包 onebot11），含 adapter_config_schema.json
- 新增 adapters/onebot11/perm/（只含常量）
- 新增 adapters/onebot11/onebottest/（本地 OneBot 11 模拟服务）
- 新增 adapters/onebot11/cmd/alonabot-adapter-onebot11/main.go（只调用 sdk.RunAdapter(onebot11.New())；返回错误时与 adapters/telegram/cmd/alonabot-adapter-telegram 一样 slog.Error 后 os.Exit(1)）
- 上述包的测试文件
允许的新跨模块边：adapters/... -> sdk/...、-> contracts（ARCHITECTURE 已登记）。adapters 不 import internal/...，测试也一样；perm 不 import 任何包。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapters/onebot11
   const AdapterID = "onebot11"
   func New() adapter.Adapter // 实现 sdk/adapter 的 Adapter
   Descriptor()：adapter_id 为 AdapterID；capabilities 只声明接收相关的能力（发送能力由 T-15.2 补充）；permissions 依次为
     {name: perm.GroupOwner, implies: [perm.GroupAdmin], label_key: "permission.group_owner"}
     {name: perm.GroupAdmin, implies: ["chat_admin"], label_key: "permission.group_admin"}
     {name: perm.PrivateFriend, label_key: "permission.private_friend"}
   不声明扩展事件与 extension。
   adapter_config_schema.json（格式同 config_schema.json，CONTRACTS §5.1）：ws_url（string，required）、access_token（string，secret）、mode（string，options ["forward"]，default "forward"）；没有 $version 与 $migrations。
2. adapters/onebot11/perm
   const GroupOwner = "onebot11.group_owner"
   const GroupAdmin = "onebot11.group_admin"
   const PrivateFriend = "onebot11.private_friend"
   包文档说明插件以 sdk.PermissionName(perm.GroupOwner) 等方式引用，本包不链接 Adapter 运行时代码。
3. adapters/onebot11/onebottest
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
  1) 解析 StartRequest.ConfigJSON：ws_url 不是 ws/wss URL、mode 不是 "forward"（缺省按 "forward"）→ INVALID_ARGUMENT（错误只写字段名，不含配置原文与 access_token）。Core 按 adapter_config_schema.json 补值前（T-08.12 之前），Adapter 自己的这层检查不能省。
  2) 用 websocket.Dial 连接 ws_url，access_token 非空时带 Authorization: Bearer；握手 401/403 → PERMISSION_DENIED，其他失败 → UNAVAILABLE。Start 失败不重试。
  3) get_login_info：account_id 为 user_id 的十进制字符串，DisplayName 为 nickname，Username 留空（OneBot 11 没有可寻址的账号用户名，命令不接受 @ 形式，D-67）；失败 → UNAVAILABLE。
  4) 启动读循环与交付循环后返回 Instance：Account() 返回 get_login_info 的账号；Capabilities 与 Descriptor 的接收能力一致；Send 本切片返回 UNSUPPORTED 的 ErrorInfo（发送由 T-15.2 实现）。
- 交付给 Core 的事件：Event 的 adapter_id、adapter_instance_id（StartRequest.InstanceID）与 account_id（Account.ID）必须与 Target 中的同名字段一致；event_id、source_kind 与 capabilities 由 Core 重建，Adapter 不填；occurred_at_unix_ms 取上报的 time（秒）乘以 1000。
- API 调用：每个请求带实例内唯一的 echo（递增的十进制字符串），响应按 echo 配对；调用以 ctx 为期限，ctx 结束时返回 ctx 的错误；连接断开时全部在途调用以 UNAVAILABLE 结束。带 echo 的帧是响应，其余帧是上报。
- 读循环一直读取，不因等待交付而停止：
  - 响应交给配对的调用；meta_event 不入队（lifecycle 的 connect 只记 DEBUG）。
  - 其余上报放进实例事件队列（256 条，命名常量）；队列满时丢弃该上报，累加丢弃计数并记 WARN（带累计数）。
- 交付循环按顺序取出上报，标准化后经 Core.Deliver 交付，游标为 nil；收到应答（无论 accepted）后才取下一条。不交付的上报直接跳过。Deliver 返回错误时：实例停止后为 FAILED_PRECONDITION，交付循环退出；其他错误记 WARN 后继续下一条（OneBot 11 没有重放，与队列满时丢弃一致）。
- 格式与身份：message 上报的 message 字段是字符串（CQ 码格式）时，经 Core.ReportFailure 报告（safe_message "onebot implementation must post messages in array format"），关闭连接并停止接收。self_id 与 account_id 不符的上报丢弃并记 WARN。
- 断线：连接关闭或读错误（Stop 引起的除外）后按有上限的指数退避重连（初始 1s、上限 60s，命名常量），重连后重新 get_login_info；user_id 变化 → ReportFailure（"onebot account changed"）并停止；握手 401/403 → ReportFailure（"onebot implementation rejected the access token"）并停止。断线期间的上报不补收；重连与已提交的 Send 无关，不重放任何请求。
- Stop：关闭连接，等待读循环与交付循环退出后返回；Stop 之后不再交付。同一实例不会有两条并行的连接。
- 标准化（MESSAGE-MODEL §7.1；数字 ID 一律转为十进制字符串）：
  - Target：adapter_id、adapter_instance_id、account_id、chat_id（"private:<user_id>" 或 "group:<group_id>"），没有 thread_id。
  - message（private、group）→ message.created：platform_message_id 为 message_id，source_event_id 同 message_id；Actor.external_user_id 为 user_id，display_name 取非空的 sender.card，否则 sender.nickname，locale absent；sub_type 为 anonymous 时不构造 Actor。
  - 平台权限（Actor.permissions，只给平台事实，展开隐含关系由 Core 负责）：群聊中 sender.role 为 owner 时为 [perm.GroupOwner]，为 admin 时为 [perm.GroupAdmin]，其他为空；私聊中 sub_type 为 friend 时为 [perm.PrivateFriend]，group（群临时会话）等其他子类型为空。撤回与成员事件的 actor 不带平台权限（上报中没有其角色）。
  - chain：text → plain text；at → mention（qq 为 "all" 时不生成）；reply → reference；image、video 本切片不生成（T-15.2 接入入站媒体）；其他段不生成。platform_payload 为原始上报 JSON，chain 可以为空。
  - to_me：私聊；chain 中有 external_user_id 等于 account_id 的 mention；reply 引用的消息经 get_msg 查得 sender.user_id 等于 account_id。get_msg 失败按 false 处理并记 DEBUG。
  - notice.group_recall、notice.friend_recall → message.deleted（platform_message_ids 只含 message_id；operator_id 存在时作为 actor）；notice.group_increase、notice.group_decrease → member.joined、member.left（member_external_user_id 为 user_id；operator_id 非 0 且不等于 user_id 时作为 actor）；其他 notice 与全部 request 不交付，计数并记 DEBUG。
- 脱敏：access_token 不进入日志、错误与普通响应；ws_url 带 userinfo 或查询参数时，记录前去掉这两部分。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/onebot11/onebot11_test.go。用 onebottest.Server 与一个记录调用的 adapter.Core 测试替身驱动 onebot11.New() 的实例（不经 Core 进程）：
  - Start 得到 get_login_info 的 account_id 与 nickname；access_token 错误时 Start 返回 PERMISSION_DENIED。
  - 推入好友私聊、群聊消息（sender.role 为 owner）、匿名群消息、撤回、入群与一条 request：依次交付 5 个事件，字段与 MESSAGE-MODEL §7.1 一致（私聊 Actor 带 onebot11.private_friend，群主带 onebot11.group_owner，匿名消息没有 Actor），request 未交付；Core 替身对第二条返回 accepted=false 时后续照样交付。
  - Disconnect 后在退避期内重连（Connections 为 2），之后推入的消息正常交付，没有重复交付；重连后 get_login_info 返回其他 user_id 时 ReportFailure。
  - 错误与日志中不出现 access_token（用测试 token 字符串做断言）。多列理由：这是凭证的信任边界。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：配置的各种非法值；Descriptor 经 validate.AdapterRegister 检查通过，权限隐含关系为 group_owner → group_admin → chat_admin；adapter_config_schema.json 经 configschema.Parse 通过，SecretPaths 只含 access_token；群管理员与普通成员、群临时会话的平台权限；echo 配对与并发调用；ctx 超时与断线时的在途调用；退避上限；Stop 期间读循环退出；ws_url 的脱敏；to_me：私聊为 true；@本账号为 true；reply 本账号的消息为 true；reply 他人消息与 get_msg 失败为 false；字符串格式的 message 上报产生一次 ReportFailure，接收停止；self_id 不符的上报未交付；Core 替身的 Deliver 阻塞期间推入 300 条上报：读循环仍能完成一次 get_msg 调用，有上报被丢弃且丢弃计数与未交付数一致；解除阻塞后按原顺序交付队列中的事件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./adapters/...；go build ./...；go mod tidy 后 go.mod 与 go.sum 无差异；go list -deps ./cmd/alonabot ./sdk 不含 github.com/coder/websocket。

不做：发送、发送计划与回执、入站与出站媒体、extension（T-15.2）；Core 按 adapter_config_schema.json 补值与迁移（T-08.12）；权限文案的 i18n 文件（label_key 的文案随 Adapter 包的 i18n 目录由打包切片加入）；Core 侧装配、alonabot run 联验与一致性验收（T-15.3）；反向 WebSocket（v0.1 之外）。
```

## 第 18 波

### goal-79 · T-15.2 官方 OneBot 11 Adapter 的发送、媒体与 extension

状态：等待前置（T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：待派发；暂不派发（维护者指示只改写语句）。前置 T-03.6（PR #55，合并提交 b023130）、T-03.8（PR #136，3621fc1）、T-15.1（PR #109，1b226a2）均已合并。2026-11-13 按维护者指定的 main 0af1255 源码及本地完成记录重新核对。以下裁决均记为「审查者补齐，待维护者确认」：下载仅开放显式 media_download_origins；共享目录在 Start 校验条件必填；模拟端全量记录只用于小对象；同实例 incoming_ref 的 platform_file 映射与完整 V-89 由 T-15.5（goal-115）交付，本切片保留阶段分工。

已有接口与衔接：onebot11.New() adapter.Adapter、Start(context.Context, adapter.StartRequest, adapter.Core) (adapter.Instance, error)、instance.call(context.Context, string, any, any) error 均存在，Send 仍返回 UNSUPPORTED；receiveCapabilities 仅声明接收。adapter.Core.ReadMedia、adapter.MediaSource.ReadPlatformFile 的分块签名已存在，adapter.Extension[T] 接收包内 face 类型。onebottest.Server 已有 Calls、Respond、SetLogin、Disconnect；本语句列出的 Files、DisconnectOn 为新增。与 goal-114 共用 OneBot 连接、生命周期、Schema 和模拟端，先合本切片；T-04.8、T-03.13 已合并。

```text
/goal 使用 $alonabot-execute 推进 goal-79（切片 T-15.2：官方 OneBot 11 Adapter 的发送、媒体与 extension）。goal 目录 goal-79，分支 goal-79，worktree .worktrees/goal-79，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现 OneBot 11 的完整发送预检、确定的物理计划与逐项回执，入站 image/video 的惰性读取、出站 https_url 与共享目录 media_ref，以及类型化 onebot11.face extension；保留原始数据、顺序与部分结果。

依据：spec/PLAN.md §3.1；spec/plan/adapters/onebot11.md T-15.2、T-15 步骤 3、4 及「失败停止条件」；spec/CONTRACTS-v0.1.md §1.2～§1.4、§8、§14.5；spec/MESSAGE-MODEL.md §2、§4、§7.1 的 image/video 入站媒体、§7.2；spec/WIRE-FORMAT.md §2、§3.4、§5；spec/adr/DECISIONS.md D-46、D-73、D-88、D-113；spec/TEST-MATRIX.md V-71 的发送与媒体分项。OneBot 11 标准 communication/ws.md、api/public.md 的 send_msg、message/segment.md 的 text/at/reply/image/video/face 字段在实施时对照原文。

范围（只新增或修改这些）：
- adapters/onebot11/onebot11.go：Send、描述与实例能力、媒体配置接入。
- adapters/onebot11/connection.go：解析 media_dir、media_download_origins；保留 API 原始响应与写出阶段，旧登录/get_msg 调用语义同步保留。
- adapters/onebot11/lifecycle.go：仅在发送调用的断链收尾需要时调整，不改变接收重连规则。
- adapters/onebot11/inbound.go：image/video 入站映射。
- 新增 adapters/onebot11/plan.go、send.go、media.go：物理计划、发送与平台文件读取；face 载荷结构体放 plan.go，保持包内类型。
- adapters/onebot11/adapter_config_schema.json：追加 media_dir 对象与 media_download_origins 列表。
- adapters/onebot11/onebottest/server.go：send_msg 缺省响应、断线注入及共享文件读取记录。
- 新增 adapters/onebot11/send_test.go；按需更新 adapters/onebot11/onebot11_test.go 中共用的 Core 测试替身；本地 *_local_test.go。

允许的新跨模块边：无。沿用 adapters/... -> sdk/...、contracts；只使用已批准的标准库与 github.com/coder/websocket。测试同样遵守 ARCHITECTURE，依赖、Core、SDK 与 proto 变更按共同约束标 [!]。

接口（照此实现；内部辅助函数自定）：
- 保持 New() adapter.Adapter、Adapter.Start 的签名；实例实现
  Send(ctx context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error)
  ReadPlatformFile(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error)
- face 载荷：type face struct { ID string `json:"id"` }；Descriptor.Extensions 以 adapter.Extension[face]("face", 1) 声明，片段 AdapterID 为 "onebot11"。Schema 由 SDK 推导，ID 的非空十进制格式在发送预检中检查。
- 配置追加 "media_dir":{"local":string,"remote":string}；配置共享目录时 local 必填，remote 缺省等于 local。Schema 为 {"type":"object","items":{"local":{"type":"string"},"remote":{"type":"string"}}}：现有 Complete 会把可选对象和字符串补成空值，两个子字段不声明 required；Start 将两者均空识别为未配置，给了 remote 而 local 为空时 INVALID_ARGUMENT，local 非空而 remote 为空时按 local。这样保持未使用共享目录的已有实例可启动；不新增迁移。
- 媒体下载配置追加 media_download_origins []string，Schema 为 {"type":"list","default":[]}。每项必须为 http/https origin（scheme、host、可选 port，无 userinfo、非根 path、query、fragment）；Start 校验后按小写主机名和有效端口比较精确 origin。列表为空时停用平台 URL 取件；它独立于 media_dir。
- onebottest.Call 保留 Action、Params，追加 Files map[string][]byte（键为请求中的 file:// 原文）；用于断言调用期间读取的字节。
  func (s *Server) DisconnectOn(action string, nth int)
  第 nth 次该 action 请求完整收到后关闭连接且不响应。send_msg 的缺省响应为 status=ok、retcode=0，data.message_id 从 1001 起递增。
- 复用现有 proto 字段，不追加编号：SendResult 与 PartResult、MediaSource、IncomingMedia、AdapterMediaReadRequest/Response。

行为：
- Descriptor 与实例能力在原接收能力上追加 message.send、message.reference、media.read，version=1、supported=true；发送的 CapabilitiesUsed 只列实际采用的已声明能力。保留原平台权限声明。
- Send 在首个平台请求前完整验证 Chain：支持 plain 且无 entities 的 text、mention、image、video、reference 与本 Adapter 的 face v1。空链或仅有 reference 为 INVALID_ARGUMENT；markdown、entities、buttons、file、voice、forward、其他 extension 为 UNSUPPORTED。合法而无法表示的 caption/alt 等显式内容须明确拒绝，保留可表达的字段，不静默丢弃。
- session.Parts() 取得 chat_id；仅接受 private:<十进制号码>、group:<十进制号码>，无 thread_id。mention 的 ExternalUserID、引用句柄中的消息 ID 须为十进制；引用会话和实例与发送一致。结构错误 INVALID_ARGUMENT，平台无法表示 UNSUPPORTED。预检失败返回 SendRejected 的有效 SendResult 和安全错误，PartResults 为空，平台请求数为零。
- 物理计划按原位置扫描：连续 text、mention、image、face 合成一条段数组；video 单独成一条；reference 变为第一条最前的 reply{id}，计入第一项的 SegmentIndexes。每条消息一次 send_msg，message_type 为 private/group，目标号码写 user_id/group_id，message 为数组。每个操作一项 PartResult，PartIndex 从 0 起。
- https_url 原样写段的 file，要求 https。media_ref（跨实例 incoming_ref 已经 Core 准备）只在配置 media_dir 时可发送，否则预检返回 UNSUPPORTED 并提示配置 media_dir。执行当前部分时用 Core.ReadMedia 每块至多 1 MiB 顺序写入 media_dir.local 的随机临时文件（crypto/rand）；平台收到 file:// 加 remote 中同名路径。随机名独立于插件 filename，该部分结束、明确失败、UNKNOWN 或取消后都删除。删除失败保留完整错误并记脱敏诊断，返回可用回执与错误。其余临时来源支持由 T-15.5 交付；本切片遇到 InlineBase64、PlatformFile 时明确返回 UNSUPPORTED。
- ok + 可用 data.message_id 为 PartAccepted 并保留十进制 ID；async 为 PartAccepted、MessageIDsUnavailable=true，无消息 ID；failed 为 PartRejected，错误 FAILED_PRECONDITION 且 safe_message 带 retcode。请求已写出后断线或到期为 PartUnknown，错误分别 UNAVAILABLE/DEADLINE_EXCEEDED；发出前连接不可用为 PartRejected + UNAVAILABLE。保留响应与写出事实，不沿用 call 抹掉状态的旧写法。任何非 ACCEPTED 后停止，后续 PartNotAttempted；整体、首个 ID 和已确认 ID 按 MESSAGE-MODEL §4 汇总，不自动重试。
- 入站 image/video 生成对应片段。存在 http/https url 时 Incoming.Token 保留完整 url；仅当其 origin 在本实例 media_download_origins 中时标 Readable，否则标 Unreadable。没有 url 时 token 保留 file 标识、标 Unreadable，读取返回 UNSUPPORTED。原片段及 platform_payload 保留；Core 接纳后自行签发 Session token。
- ReadPlatformFile 在调用时才经 HTTP GET 按 offset/length 取件。允许的目标仅为本实例显式列出的 media_download_origins；Core 只会转交本 Adapter 可信上报登记的标识。未获准 origin 或非 http/https 来源返回 UNSUPPORTED，HTTP 请求数为零。实例复用一个 HTTP 客户端并使用 http.ProxyFromEnvironment；所有 3xx 均停止，禁止跟随重定向；其他非 2xx 为 UNAVAILABLE。管理员可显式列出所需的私网/loopback origin；生产与模拟服务使用同一规则。Range 成功校验 Content-Range；服务忽略 Range 时流式跳过 offset、最多读取 length，并按分块接口返回 next/eof，避免整份积累。
- access_token、URL userinfo 和查询凭证沿用脱敏；平台已接受、未知结果、临时文件清理错误均完整保留。

测试（按编码规范「测试提交规则」）：
- 提交：adapters/onebot11/send_test.go 一个集成测试 TestSendPlan，经 onebottest.Server 与 recordingCore 驱动 New().Start 得到的实例。子场景覆盖：reference+text+mention+image(media_ref)+face+video(https_url)+text 得到三次 send_msg，第一条 [reply,text,at,image,face]、第二条 video、第三条 text；共享文件在调用期间可读且结束后目录为空；SegmentIndexes 与三项回执正确。第二次 failed 得 PARTIAL / ACCEPTED、REJECTED、NOT_ATTEMPTED；DisconnectOn 得 UNKNOWN 且后续未尝试；async 得 ACCEPTED 且无可用 ID。多场景理由：V-71 明确要求区分三类响应与写出后断线，只有经实际 WebSocket 才覆盖 echo 和写出阶段。
- 本地自测（*_local_test.go，PR 写明命令与结果）：入站 image/video、无 url/未获准 origin 的 Unreadable、白名单精确比较（协议/主机/端口）、显式私网与 loopback、未配置时零下载、跨 origin 和同 origin 重定向均拒绝、平台文件分块/Range/EOF；face 类型推导及非法 ID；无 media_dir 时零平台请求；预检不支持内容；取消与期限、写文件/删除失败；大媒体常驻内存不随总大小增长（实现端替身流式读文件，不使用 Call.Files 保存大对象）；旧登录、get_msg、接收重连保持；-race 下并发 API 响应配对。
- 真实 OneBot 11 实现端：未运行（无凭证）；上述证据标明「onebottest 本地模拟服务」，记录实际 OS/架构、Go 版本、命令和结果，交给审查者回写 TEST-MATRIX §3.2。

完成条件：上述集成与本地自测覆盖本切片分项；gofmt -l 本次改动 Go 文件无输出；go vet ./...；go test -race ./adapters/onebot11/...；go build ./...。遇到 T-15「失败停止条件」、范围外修改、依赖或协议缺口标 [!]，保留阻塞证据。

不做：Core/根包/CLI 装配、T-15.3 跨 Adapter 一致性、message.edit、message.delete、voice/file/forward/at_all、临时媒体 URL 与 inline_base64/platform_file 的平台映射（T-15.5）、反向 WebSocket（T-15.4）、真实 QQ 端验证。
```

### goal-133 · T-15.7 官方 OneBot 11 Adapter 填写 chat_title

状态：等待前置（T-04.16）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：待派发。暂不派发（维护者指示只改写语句）。直接前置 T-15.1（goal-78，PR #109，合并提交 `1b226a2`）、T-03.25（goal-131，PR #166，合并提交 `6f72fc8`）均已合并。2026-11-13 按维护者指定的 main `9efee73` 只读副本，用 `rg` 核对定义、调用者与完整 Event 断言；`message.Event`、`adapter.Core`、`onebottest.Server.Respond` 另经离线 `go doc` 核对输出。以下补齐方案均为「审查者补齐，待维护者确认」。

按 main 核对：

- `sdk/message/event.go` 已有 `message.Event{ChatTitle string, ChatType message.ChatType, Session message.Session, ...}`；空 ChatTitle 表示未提供。`contracts/proto/alonabot/v1/message.proto` 已有 `ChatType chat_type = 15`、`optional string chat_title = 16`；`sdk/internal/convert/event.go` 的 `EventToProto(e *message.Event) (*pb.Event, error)` 与 `EventFromProto(e *pb.Event) *message.Event` 已处理该字段。
- `sdk/adapter/adapter.go`：`Core.Deliver(ctx context.Context, ev *message.Event, cursor []byte) (accepted bool, err error)`。`adapters/onebot11/onebot11.go` 已有 `func New() adapter.Adapter`、`func (*onebotAdapter) Start(ctx context.Context, request adapter.StartRequest, core adapter.Core) (adapter.Instance, error)`、`func (i *instance) Stop(context.Context) error`；每次成功 Start 建立独立 instance，交付循环消费其 256 条实例队列。
- `adapters/onebot11/inbound.go`：`func (i *instance) deliverEvents()` 顺序调用 `normalize(ctx context.Context, raw json.RawMessage) (*msg.Event, error)`，等待所属连接账号核对后交付，cursor 为 nil。`onebotPost.Sender` 已有 `Card`、`Nickname`、`Role string`，`GroupID`、`UserID` 为 `*int64`；`messageEvent(ctx context.Context, post onebotPost, event *msg.Event) (*msg.Event, error)`、`noticeEvent(post onebotPost, event *msg.Event) (*msg.Event, error)` 是既有标准化入口。当前尚未填写 ChatTitle、ChatType，也没有群名缓存。
- `adapters/onebot11/connection.go`：`func (i *instance) call(ctx context.Context, action string, params any, result any) error` 已按 echo 配对，并处理取消、断线、非 ok/非零 retcode 与 JSON 解码错误；读循环独立处理 API 响应与心跳。`lifecycle.go` 的重连复用同一个 instance；Stop 取消 receiveCtx、关闭连接并等待交付结束。
- `adapters/onebot11/onebottest/server.go` 已有 `NewServer(t testing.TB, accessToken string) *Server`、`URL() string`、`SetLogin(userID int64, nickname string)`、`Push(events ...json.RawMessage)`、`Calls() []Call`、`Respond(action string, fn func(params map[string]any) (status string, retcode int, data any))`；`Call` 已含 `Action string`、`Params map[string]any`，数字参数按 json.Number 保留。`defaultResponse(action string, params map[string]any) (string, int, any)` 尚无 get_group_info。所有动作已有统一调用记录，增加响应分支即可取得查询次数。
- 审查者补齐，待维护者确认：群名缓存归单个存活 instance，成功名与失败结果都缓存；同实例断线重连继续复用，Stop 后重新 Start 得到新缓存。查询在首次有效事件的顺序交付路径上完成，原语句的 5 秒以 Adapter 包内命名常量 groupInfoTimeout 保留，属于 CONTRACTS §16 明确排除的 Adapter 自有常量。空或缺失的 group_name 均保持空名称，缺字段或错误类型按平台响应失败记录一次；原有事件照常交付。
- 审查者补齐，待维护者确认：在已通过来源、结构与事件类型检查后统一填写现行 D-100 的 ChatType；private: → ChatPrivate（含群临时会话），group: → ChatGroup。goal-115 原先规划的类型赋值由本切片承担，审查者正式派发时同步其语句；基线已有时复用。
- 在途协调按本次维护者提供的信息：PR #177（goal-66）改 `internal/core/logging/`、`internal/core/manage/`、`sdk/log.go`、`sdk/adapter_run.go`、根包装配与 `cmd/alonabot/`；PR #178（goal-60）改 `agent/`；goal-41 改 Journal、state、manage、supervisor 与根包装配；goal-9 的生产修改位于 `adapters/telegram/plan.go`，另有 Telegram 示例和 CLI 集成测试。这些与本切片生产文件分离。尚待实施的 goal-79、114、115、116 会共用 `onebot11.go`、`inbound.go`、`onebottest/server.go`、`onebot11_test.go`；后合并者保留缓存、类型与已有收发逻辑，群名逻辑经共同交付路径复用。
- 旧在途提示按完成记录更新：PR #165（`330164f`）、#167（`3a9b4af`）、#168（`37b0ada`）、#169（`4dc4bcd`），以及 goal-59／PR #170（`eca6310`）、goal-69／PR #171（`3043c4e`）均已合并；#169 的 schema.sql／manage 清理清单及 goal-69 的根包／cmd 装配已是现有基线。最近 #173（`2bb8181`）、#175（`ac6083a`）、#176（`9785c22`）及维护者告知已合并的 #174 均保留，本切片复用现有启动、取消与 SDK 交付接口。
- 需维护者另行决定：新增依赖、改变已确认决策均无新增请求；真实 OneBot 11 实现端、QQ 测试账号与凭证由维护者决定，模拟验收可独立完成，真实平台结果留作「未运行」。

```text
/goal 使用 $alonabot-execute 推进 goal-133（切片 T-15.7：官方 OneBot 11 Adapter 填写 chat_title）。goal 目录 goal-133，分支 goal-133，worktree .worktrees/goal-133，从包含 PR #109、#166 的最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：OneBot 11 实例为私聊与群的已支持入站事件填写 chat_title；私聊取 sender.nickname，群首次经 get_group_info 取得名称并在实例进程内缓存，失败后同一实例运行期间保持无名称且事件继续交付。在同一构造路径填写现行契约规定的 chat_type。

依据：spec/PLAN.md §3.1、§5；spec/plan/adapters/onebot11.md T-15.7、T-15 步骤 1、2、9 及「失败停止条件」；spec/plan/core.md T-03 步骤 22 已合并的聊天目录；spec/CONTRACTS-v0.1.md §8（OneBot 11 实例交付与聊天目录）、§14.1（日志）、§16（Adapter 自有常量）；spec/MESSAGE-MODEL.md §1、§7.1；spec/WIRE-FORMAT.md §3.4（core.adapter.deliver）、§5A（Event 的追加字段）；spec/adr/DECISIONS.md D-46、D-86、D-100、D-130；spec/TEST-MATRIX.md V-117 的 OneBot 11 分项及 §3 的证据格式。平台请求按 OneBot 11 标准的 get_group_info、私聊/群聊消息与通知上报核对。

范围（只新增或修改这些）：
- adapters/onebot11/onebot11.go：instance 增加私有群名缓存，并在创建新实例时初始化；其余生命周期与公开方法保持现状。
- adapters/onebot11/inbound.go：在有效事件的共同标准化路径填写私聊/群名和既定 ChatType；包内 groupInfoTimeout 与群名查询辅助函数也放在本文件。
- adapters/onebot11/onebottest/server.go：defaultResponse 追加 get_group_info 缺省响应，复用现有请求记录及 Respond 覆盖。
- adapters/onebot11/onebot11_test.go：追加下述一个集成测试，更新既有完整 Event 比较的名称、类型预期，保留其余断言。
- adapters/onebot11/ 下本切片的 *_local_test.go：失败、期限、重连、重启、通知与日志验证，只作本地自测。
允许的新跨模块边：无；沿用 sdk/adapter、sdk/message、sdk/errcode 和本模块模拟服务；WebSocket 使用已有批准依赖。

接口（照此实现；内部辅助函数自定）：
- 保持 func New() adapter.Adapter、Adapter.Start、Instance.Stop、adapter.Core.Deliver 的既有签名；normalize、messageEvent、noticeEvent 与 call 的现有签名保持原样。
- 使用既有 message.Event.ChatTitle string 与 ChatType；协议字段沿用 Event.target = 8、Event.chat_type = 15、optional string Event.chat_title = 16，SDK 负责空串与 absent 的转换。
- instance 追加私有 groupTitles map[int64]string：键为群号，map 中存在该键表示已尝试查询，值为空可表示失败或平台没有给出名称。只由唯一交付循环读写，在 Start 创建的新实例上初始化；禁止为它增加后台刷新循环。
- 包内 const groupInfoTimeout = 5 * time.Second。每次首次查询从 normalize 的 ctx（生产为 receiveCtx）派生 context.WithTimeout，结束时 cancel；复用 i.call(ctx, "get_group_info", params, &result)。params 只有数值 group_id，沿用标准 no_cache 的缺省行为；读取 data.group_name，缺字段或非 string 为响应失败，空 string 视作未提供名称。group_id 在平台 JSON 中保持数值，缓存与日志按同一群号索引。
- onebottest 的既有公开 API 保持现状。defaultResponse("get_group_info", params) 返回 status="ok"、retcode=0、data 包含请求的 group_id 与 group_name="test_group"。测试经 Respond("get_group_info", fn) 控制不同群名、failed、不支持、畸形数据或不应答；通过遍历 Calls() 中 Action=="get_group_info" 的记录按 group_id 计数，复用 json.Number 读取数字。

行为：
- 先完成现有 self_id、支持的消息/通知类型、会话标识、消息数组与对应 payload 检查，再为即将交付的有效事件取名；被忽略或拒绝的上报保持原处理方式。私聊（含群临时会话）仅取该上报 sender.nickname；为空时 ChatTitle 留空。Actor.DisplayName 原有 card 优先规则继续生效，聊天名称单独按 nickname 取值。
- 所有已支持的 group: 事件在本实例首次遇到该群时，先尝试一次 get_group_info，再交付当前事件；成功时当前事件及随后同群事件都带 group_name 原文。匿名群消息和群撤回、成员通知适用同样规则；私聊通知缺少 sender.nickname 时保留空名称。没有 target 的事件以及没有会话的扩展事件保持空名称。
- 在相同位置按最终会话填写 ChatType：private: 为 ChatPrivate，group: 为 ChatGroup，群临时会话依然为私聊。已有相同映射时直接复用。
- 查询成功、调用失败、超时、实现端不支持、缺失/畸形响应都使该群进入已尝试状态。失败时本事件 ChatTitle 留空，继续原事件交付；同实例运行期间不再为该群查询。名称获取失败只记一条英文 DEBUG，业务字段仅 adapter_instance_id 与群号；名称、响应体、认证信息与可能带名称的错误原文均不进入日志。
- 顺序交付循环等待查询最多 groupInfoTimeout，之后继续当前事件及队列中后续事件；其他群照常排队，保持原有事件顺序。API 响应与心跳继续由独立读取循环处理，原有 256 条队列及满队列计数不变。查询只使用既有单次 API 调用；查询持锁范围和取消规则沿用 call。
- 同一个存活 instance 的 WebSocket 重连复用缓存。群改名后在该实例存活期间保持缓存值；Stop 后重新 Start 创建空缓存，Adapter 进程重启自然清空全部实例缓存；不同实例的缓存独立。群名查询失败只影响可选名称，原有致命格式错误、账号变化及 access_token 错误继续按各自规则处理。
- Stop 或 Session 停止沿既有 receiveCtx 取消查询并等待交付循环结束，停止后没有新的 Deliver。PlatformPayload 保持完整原始上报，消息、Actor、权限名集合、to_me 和 nil cursor 的语义保留。

测试（按编码规范「测试提交规则」）：
- 提交：在 adapters/onebot11/onebot11_test.go 新增一个 TestChatTitle 集成测试。用现有 onebottest.Server、SetLogin、Respond、New().Start 和录制 Core，经真实 WebSocket 依次推入同群两条消息及一条私聊消息；为 get_group_info 的该群返回固定完整名称，私聊同时提供不同 card 与 nickname。断言两条群事件都带该群名、ChatGroup，私聊带 nickname、ChatPrivate；Calls() 中该群 get_group_info 恰好一次且 group_id 数值正确；事件交付顺序、原始载荷与 nil cursor 保持正确。理由：覆盖真实 API echo 响应、实例缓存、平台事件标准化和交付的主路径。
- 既有测试同步：TestReceiveLifecycle 完整 Event 比较中，私聊使用已有 sender.nickname，群使用模拟端新缺省的 test_group，并补齐 ChatType；原先的群主/匿名/撤回/成员、重连、accepted=false 与凭证脱敏断言保留。此项属于既有测试的预期更新。
- 本地自测（adapters/onebot11/*_local_test.go，PR 记录命令、平台与结果）：failed/非零 retcode、不支持 API、超时不应答、缺字段/错误类型与空群名；每种情形同群两条事件均继续交付且总查询一次，失败情形诊断一次，合法空群名缓存空串。查询超时后后续群仍按序交付，API 读循环继续处理 echo 与心跳；Stop 在查询中及时取消、结束后无交付。
- 本地自测继续覆盖：首次事件为群撤回/成员通知、匿名群消息；群临时会话取昵称；缺 nickname 的私聊通知；两群各查一次；同实例断线重连不再查询；群改名保持缓存；同一进程 Stop 后重新 Start 再次查询；两个实例同群独立查询；不匹配 self_id、未知上报、非法消息不触发群名查询。DEBUG 日志不含群名、私聊昵称、响应原文与 access_token，错误文本保持英文。
- D-86：提交范围为上述平台集成测试；SDK 字段编解码、Core 目录持久化与查询复用 PR #166 的证据。本切片的本地自测也集中在 adapters/onebot11/，无需新增 SDK 测试或 testdata 插件。
- 真实 OneBot 11 实现端与 QQ：未运行（待维护者提供环境）。模拟证据与真实手测分别记录；执行者把 V-117 OneBot 11 分项的测试位置、命令、OS/架构、Go 版本、结果及未运行项交审查者汇总到 TEST-MATRIX §3.2。

完成条件：主路径集成及列明本地自测取得结果；每实例每群只尝试一次查询，成功与失败均缓存、事件顺序与停止语义保持有效，名称及凭证无日志泄漏；gofmt -l 本次修改的 Go 文件无输出；go vet ./adapters/onebot11/...；go test -race ./adapters/onebot11/...；go build ./...。对照 T-15「失败停止条件」：通用 Segment 语义、Core 平台分支、凭证泄漏、静默片段丢弃等问题先修复；需要范围外生产文件、proto、新依赖或改变已确认语义时标 [!] 交审查者处理。V-117 全项由跨任务汇总关闭，本切片交付其 OneBot 11 分项。

不做：Core 聊天目录与 schema；群成员、群列表拉取；按时间或改名事件刷新群名；公开的平台 get_group_info 动作（goal-116）；反向 WebSocket、发送与媒体能力；新增配置项、可调参数或 SDK 公开方法。
```

## 第 19 波

### goal-80 · T-15.3 跨 Adapter 一致性验收

状态：等待前置（T-03.10、T-15.2）。2026-11-09 按 D-88 补充：OneBot 11 的 media_ref 需要实例配置 media_dir（goal-115 合并后也可用临时媒体 URL），审查者改写时在 OneBot 11 的实例文件中配置 media_dir，并断言请求中没有 base64://。语句中 goal-9、goal-78、goal-79 提供的符号（alonabot run 的 --adapter-id/--adapter-entry、cmd/alonabot/telegram_e2e_test.go 的装配方式、telegramtest.Server、onebottest.Server）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：OneBot 11 的端到端链路经 `alonabot run <目录>` 运行，实例写在 `core_config.json`（D-77、D-82）。前置合并后由审查者改写。

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

### goal-114 · T-15.4 官方 OneBot 11 Adapter 的反向 WebSocket

状态：等待前置（T-04.16、T-15.2）：维护者在 PR #196 确认加入 T-04.16 前置，按 D-146 分包后的接口改写本语句后再派发（2026-11-14）。此前：等待前置（T-15.2）；暂不派发（维护者指示只改写语句）。原前置 T-03.19（PR #152，2f4284d）、T-15.1（PR #109，1b226a2）均已合并。2026-11-13 按 main 0af1255 重新核对，T-15.2（goal-79）的 Send 尚未实现，新增它为直接前置以保留完整反向收发验收。以下裁决均记为「审查者补齐，待维护者确认」：Core 保留 Start 响应时的 Account 展示名快照；凭证优先使用存在的 Authorization；条件必填由 Start 校验；统一 HTTP 入口使用已合并装配。

已有接口与衔接：sdk.AdapterHTTP.HTTP(route func(method, path string) *sdk.RouteBuilder)、RouteBuilder.Public/HandleFunc 及 AuthPlatformToken 均已存在；sdk.RunAdapter 托管 Listener、内部代理凭证与 Session。onebotAdapter 当前仍为空结构体，parseConfig 只接受 forward、Start 总先 Dial/get_login_info，无进程级实例表；startRead 与 receiveConnection.checked/valid 是现有复核门禁。onebottest.Server 仍仅提供正向端点，ConnectReverse 为本切片新增。根包管理装配（PR #163）、run/热部署（PR #171）、启动取消修复（PR #174）已合并；本切片复用它们，不改根包/CLI。goal-79 先合并，并在派发前重新核对 connection.go、lifecycle.go、Schema、模拟端和新增 send.go；media_dir、media_download_origins 与真实 Send 必须完整保留。

```text
/goal 使用 $alonabot-execute 推进 goal-114（切片 T-15.4：官方 OneBot 11 Adapter 的反向 WebSocket）。goal 目录 goal-114，分支 goal-114，worktree .worktrees/goal-114，从包含已合并 goal-79（T-15.2）的最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：OneBot 11 实例支持 mode=reverse，经 Core 统一端口的公开 /ws/ 前缀接收实现端连接；完成配置、逐项握手鉴权、单连接、get_login_info 账号核对、断开后等待重连；onebottest 支持反向客户端。

依据：spec/PLAN.md §3.1；spec/plan/adapters/onebot11.md T-15.4、T-15 步骤 6 及「失败停止条件」；spec/CONTRACTS-v0.1.md §7.2、§8、§14.5；spec/MESSAGE-MODEL.md §7、§7.1、§7.3；spec/WIRE-FORMAT.md §2、§3.3、§3.4（Adapter HTTP 与平台 token）；spec/adr/DECISIONS.md D-25、D-46、D-94、D-113、D-136；spec/TEST-MATRIX.md V-95 的 OneBot 11 分项；OneBot 11 标准 communication/ws-reverse.md、api/public.md 的 get_login_info。

范围（只新增或修改这些）：
- adapters/onebot11/onebot11.go：进程级实例登记、Start 分模式、Account 展示名同步与 Stop 撤销（Core 仍保留 Start 快照）。
- adapters/onebot11/connection.go、lifecycle.go：模式配置、连接复用、反向断线与在途调用收尾。
- 新增 adapters/onebot11/reverse.go：sdk.AdapterHTTP 登记与反向 WebSocket Handler。
- adapters/onebot11/adapter_config_schema.json：mode 的 reverse 选项、self_id 与条件字段。
- adapters/onebot11/onebottest/server.go 与新增 onebottest/reverse.go：反向客户端，共用响应、上报和断线控制。
- adapters/onebot11/onebot11_test.go：一个新增反向连接集成测试；本地 *_local_test.go。
- adapters/onebot11/send.go：在已合并 T-15.2 的发送实现接入 reverse 未连接/未复核的 UNAVAILABLE 门禁，保留物理计划与回执。派发基线须包含 goal-79；该文件和其私有函数名在前置合并后由审查者按实际 diff 核对。

允许的新跨模块边：无。沿用 adapters/... -> sdk/...、contracts，以及本 Adapter 已批准的 github.com/coder/websocket。HTTP 路由由 SDK/Core 托管，测试用 net/http/httptest；不改根包、cmd、Core、SDK 或 proto。

接口（照此实现；内部辅助函数自定）：
- 保持 func New() adapter.Adapter；返回对象新增实现
  func (a *onebotAdapter) HTTP(route func(method,path string) *sdk.RouteBuilder)
  在该方法一次登记 route(http.MethodGet,"/ws/").Public(sdk.AuthPlatformToken).HandleFunc(a.reverseHTTP)。HTTP 方法与路径使用现成 builder，禁止使用旧 HTTPRegistry/HTTPOption 签名。
- 配置 JSON：mode 为 forward（缺省）或 reverse；forward 的 ws_url 必填，access_token 可选；reverse 的 access_token 与 self_id 必填，self_id 为正十进制 QQ 号 string，ws_url 不使用。Schema 的 mode.options 追加 reverse，删除 ws_url 的无条件 required，ws_url.condition 为 {"mode":"forward"}，self_id 为可选 string、condition 为 {"mode":"reverse"}；access_token 保留 secret 且不加 required。CONTRACTS §5.1 的 condition 只控制显示，模式相关必填统一在 Start 的外部配置边界校验；Schema 补齐的空字符串按对应模式解释，不扩展 Schema 引擎。
- Account() adapter.Account：reverse 启动后 ID 为 self_id，DisplayName 初始为空，验证连入后更新为 nickname；Username 为空。读写同一 account 有同步，ID 在实例存活期固定。SDK/Core 当前只在 Start 响应取一次 Account：Core 管理面的 account_display_name 及初始 bot.online 取 SDK 制作 Start 响应时那次 Account() 的快照，尚未复核时为空；后续 nickname 仅供实例内 Account() 读取，Core 不刷新；本切片采用现有快照语义，不新增状态刷新服务。
- Instance.Send 和既有可选动作使用原接口签名，不新增平台动作或任意透传。
- onebottest 在既有 Server 上追加
  func (s *Server) ConnectReverse(ctx context.Context,endpoint string,headers http.Header) (status int,err error)
  endpoint 为 ws/wss 地址；headers 原样交握手，成功 status=101；失败保留 HTTP 状态及错误。成功后 Server 作为 OneBot 实现端读取 API 并回复、Push 推送上报；SetLogin/Respond/Calls/Disconnect/Connections 复用原接口。t.Cleanup 关闭并等待全部反向读写循环。
- 沿用已存在的 HTTP_AUTH_MODE_PLATFORM_TOKEN、REGISTER routes 与 HttpListener 字段及 AdapterInstanceStartResponse，不追加 proto。

行为：
- Adapter 进程持有按 InstanceID 索引的实例 map；Start 成功登记，失败和 Stop 完成时撤销；只对正在运行的 reverse 实例开放连接。forward 保持原 Dial、get_login_info、退避重连和致命错误规则，实例之间的启停互不影响。
- reverse Start 完成配置校验和接收生命周期初始化后立即返回，不等待外部连入、不主动 Dial；account.ID=self_id。外部地址为 <base_url>/plugins/<Adapter 包 ID>/ws/<instance_id>，Adapter Handler 收到的相对路径为 /ws/<instance_id>。Adapter 包 ID 与 adapter_id 按各自身份使用。
- 按以下顺序返回 HTTP 状态，均在 Upgrade 前：未知、未启动、非 reverse 或已停止实例 404；Authorization: Bearer 或 access_token 查询参数的凭证不匹配 401，恒时比较；X-Self-ID 不等于 self_id 或 X-Client-Role 不为 Universal 403；已有连接或同实例正在接受连接 409。Authorization 头存在时仅校验该头：必须是单个有效 Bearer 凭证；即使查询参数正确，头缺少有效 Bearer 或凭证不符仍返回 401。仅头完全缺失时读取单个非空 access_token 查询参数；多值/空值返回 401。恒时比较 token，不记录请求原文、认证头或查询值。
- 单连接的检查与占槽须覆盖并发握手；Upgrade 失败释放自身槽。旧连接的收尾仅撤销自己，避免把后来连接清掉。Stop 期间冻结新握手，关闭已有连接，等待读、交付及握手收尾后再返回。
- Upgrade 后读循环先就绪，通过同一 echo 配对调用 get_login_info；复核 user_id=self_id 后更新展示名、开放该连接的事件交付。账号不符或登录调用失败时关闭该连接并记安全诊断，丢弃复核前排队的上报，实例继续等待下一条连接。不可在账号核对前向 Core 交付消息。
- 复核成功后沿用 forward 的事件解析、self_id 检查、256 条队列、满队列计数、逐条 Deliver、空游标、API/心跳独立读取与数组格式检查。字符串消息仍是实例致命错误并 ReportFailure，符合原格式契约；单纯断线/登录不符的处理按 reverse 的规则等待。
- 断线后结束该连接的在途调用、清理连接槽，实例继续等待；不启动向外 Dial 的退避循环，不 ReportFailure、不合成 bot.offline。未连接或尚未完成账号复核时 Send/已有动作在首个平台业务请求前返回 UNAVAILABLE，重连成功恢复；已写出的发送在断开时仍保留 UNKNOWN 和可用回执，不重放。
- 生命周期 ctx 独立于 Start RPC，HTTP Upgrade 的连接寿命独立于握手请求的短期 context，Stop 与 SDK Session 失效能结束它。SDK 内部代理凭证照常先验证，平台 access_token 由本 Handler 再验；生产进程只使用 SDK 托管的 Listener。
- 脱敏沿用 access_token、URL userinfo/查询参数规则；所有状态码错误响应只含安全短文，不回显凭证。

测试（按编码规范「测试提交规则」）：
- 提交：在 onebot11_test.go 追加一个集成测试 TestReverseLifecycle，httptest 仅托管生产 reverseHTTP Handler，通过 onebottest.ConnectReverse 驱动真实 WebSocket；实例 reverse Start 在未连入时返回，账号为 self_id。正确握手后 get_login_info 核对，推入群消息正常 Deliver；第二个独立模拟客户端同实例连接得到 409；断开后重新连入、再次交付且无 ReportFailure。同一用例通过实例 Send 回发一条文本，断言 send_msg、ACCEPTED 及其消息 ID，断线期 Send 为 UNAVAILABLE，重连后同样恢复。T-15.2 为新增直接前置，收发验收完整保留。
- 本地自测（*_local_test.go，PR 写明结果）：404/401/403/409 的先后次序；Bearer 与查询参数；正确头+错误查询成功、错误头+正确查询 401、格式错误头+正确查询 401、多头/多查询 401；未知实例、forward 实例、Stop 后；并发握手仅一条成功；登录账号不符后可再次正确连入；未复核连接不交付、不发业务请求；同包多实例隔离；Stop/Session 失效关闭连接；连接收尾不影响新连接；前向模式回归、字符串格式仍致命；错误响应和日志无 token。
- 本地自测的 SDK/Core 集成：复用 PR #152 的真实 SDK Adapter 进程与统一代理模式，在临时测试中接入本 Adapter 的 HTTP 声明，确认 /plugins/<包 ID>/ws/<实例> 可达、Authorization/查询参数保留、内部凭证必验、Unbind 关闭 Upgrade；只在本地测试文件/临时目录建探针，提交范围保持上述一个平台集成测试。SDK 路由登记使用插件对象与 RouteBuilder，所需消息 Content 使用 sdk/message。
- 真实 OneBot 11 实现端：未运行（无凭证）。分别记录 onebottest 平台 Handler 模拟证据与 SDK/Core 代理集成证据；提供 OS/架构、Go 版本、命令及结果给审查者回写 TEST-MATRIX §3.2。

完成条件：握手、单连接、复核、断线重连及收发有证据，已在包含 T-15.2 的基线上完成收发验收；gofmt -l 本次改动 Go 文件无输出；go vet ./...；go test -race ./adapters/onebot11/...；go build ./...。遇到 T-15「失败停止条件」、范围外文件/依赖/proto 变更或未解决契约冲突标 [!]。

不做：Core 代理/根包/CLI 装配重写、额外公开 Listener、Telegram webhook、未声明平台动作、发送能力实现（T-15.2）、媒体 URL/新增片段（T-15.5）、扩展事件与动作清单（T-15.6）、真实实现端手测。
```

### goal-115 · T-15.5 官方 OneBot 11 Adapter 的新增片段与媒体 URL

状态：等待前置（T-03.13、T-03.14、T-15.2）。2026-11-09 按 D-87～D-91、D-95 新增，同日按 D-100 追加入站事件的 chat_type。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。

目标：OneBot 11 实例收发语音（record）、文件（upload_group_file、upload_private_file）与合并转发（send_group_forward_msg、send_private_forward_msg），注册 extension `onebot11.at_all` 并收发全体提及（入站 `at qq=all` 映射为该 extension），入站 face 映射为 extension 片段、forward 映射为只带 platform_forward_id 的 forward；media_ref 在没有 `media_dir` 时经 `core.media.url` 取得临时媒体 URL，inline_base64 原样传递，`platform_file` 原样作为 file；入站事件按 chat_id 前缀填写 chat_type。

依据：PLAN T-15 步骤 7；CONTRACTS §1.3、§8（OneBot 11 段）；MESSAGE-MODEL §2、§7.1、§7.2；DECISIONS D-88、D-89、D-90、D-91、D-95、D-100；TEST-MATRIX V-89、V-90、V-98 的 OneBot 11 分项，V-96 的入站片段分项。

要点：
- 物理计划在 goal-79 的基础上追加：语音、文件、合并转发各自单独一个物理操作；文件的平台响应没有 message_id，ACCEPTED 时标记 message_ids_unavailable；第一项为文件或合并转发时 reference 返回 UNSUPPORTED。
- 合并转发的 node：platform_message_id 为 `id`，chain 为 `content`（节点内的 media_ref 同样按本 goal 的来源规则），source 给出时为 `user_id`、`nickname`，否则为本账号与展示名。
- 媒体来源：只有插件给出 inline_base64 时请求中才出现 `base64://`；media_dir 优先于媒体 URL；媒体 URL 的前缀取实例的 `media_url_base`，缺省时不传前缀由 Core 取 base_url。
- chat_type：`private:` 的事件（含群临时会话与好友通知）为 PRIVATE，`group:` 的事件为 GROUP。
- Capabilities 追加 `message.voice`、`message.forward`，Descriptor 的 extensions 追加以 `adapter.Extension[AtAll]("at_all", 1)` 声明的 at_all（AtAll 为空结构体，推导出的 Schema 不允许任何属性，D-113）；`onebottest` 追加 upload_*_file 与 send_*_forward_msg 的缺省响应。
- 测试：提交在既有 `send_test.go` 中追加的用例：一条 Chain 依次含语音、合并转发与文件时的三次请求参数，以及 media_ref 以媒体 URL 传递且请求中没有 base64；其余进本地自测。

### goal-116 · T-15.6 官方 OneBot 11 Adapter 的扩展事件、平台动作与撤回

状态：等待前置（T-03.15、T-03.16、T-15.2）。2026-11-09 按 D-87、D-92、D-93、D-95 新增。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。

目标：在 Adapter 描述中以 Go 类型声明 MESSAGE-MODEL §7.1 的十种扩展事件（含 NapCat 的 input_status）并完成映射；以参数与结果的 Go 类型声明 §7.4 的平台动作（含群信息与成员查询），Schema 都由 SDK 推导（D-113），在 `adapters/onebot11/action` 为每个动作发布带类型的 Go 函数（第一个参数为 goal-106 的调用者接口，只依赖 `sdk` 根包），扩展事件常量发布在新增的 `adapters/onebot11/event`；`adapter.delete` 逐个调用 delete_msg，声明 `message.delete`。

依据：PLAN T-15 步骤 8；CONTRACTS §8（平台动作，撤回、输入状态、表情回应与草稿，扩展事件）；MESSAGE-MODEL §1.2、§4A、§7.1、§7.4；DECISIONS D-92、D-93、D-95、D-113；TEST-MATRIX V-91、V-92 的 OneBot 11 分项，V-96 的扩展事件分项。

要点：
- 扩展事件的 payload 类型只含 MESSAGE-MODEL §7.1 表中的字段，数字 ID 为十进制字符串，放在 `adapters/onebot11/event`：Adapter 以 `adapter.Event[T]` 声明，插件以 `sdk.AdapterEvent[T]` 引用同一类型，推导出的 Schema 不允许其他属性（D-113）；Target 与 Actor 按 §7.1 的规则；未声明的 notice 与 request 仍只计数。
- 动作参数中可由 Target 推出的 group_id、user_id 不重复出现；结果为响应 data 原文；failed 为 FAILED_PRECONDITION 并带 retcode；实现端扩展的动作不探测实现端类型。
- 撤回逐个 delete_msg，失败或结果未知后其余为 NOT_ATTEMPTED。
- 测试：提交在既有测试文件中追加的用例：推入好友请求得到 onebot11.friend_request，再用 set_friend_add_request 动作处理，断言请求参数；其余进本地自测。
