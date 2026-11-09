# 实施计划：适配器 · OneBot 11

总目标、章节划分与跨章节依赖见 [PLAN](../../PLAN.md)；本章节的 goal 语句见 [onebot11.md](../../../docs/goals/adapters/onebot11.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本子章节覆盖官方 OneBot 11 Adapter `adapters/onebot11/` 的平台实现：正向与反向 WebSocket 连接与接收、发送计划与回执、媒体、`onebot11.face` extension、平台权限，以及本地模拟服务 `adapters/onebot11/onebottest/`，即工作包 T-15；T-15.4～T-15.6 是 D-87 对标补齐的部分，协议与 Core 侧在核心切片 T-03.13～T-03.16、T-03.19。T-15.3 的跨 Adapter 一致性验收同时覆盖 Telegram，放在本子章节是因为它在 OneBot 11 完成后才能进行。章节总述与扩展约定见[适配器实施计划](README.md)。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-15.1 | goal-78 | 官方 OneBot 11 Adapter 的连接与接收：正向 WebSocket、access_token、get_login_info、有界退避重连、self_id 校验、数组格式检查、实例事件队列、私聊/群聊/撤回/成员事件的标准化、to_me、群主、群管理员与好友私聊的平台权限及 `adapters/onebot11/perm`、本地 OneBot 11 模拟服务 | T-15 步骤 1、2 | T-03.3、T-03.4、T-04.6 | V-71 的接收分项、V-85 的 OneBot 11 分项 | 已完成（[PR #109](https://github.com/AlonaBot/AlonaBot/pull/109)） |
| T-15.2 | goal-79 | 官方 OneBot 11 Adapter 的发送计划与回执、入站与出站媒体（https_url 原样传递，media_ref 经共享目录 `media_dir`，D-88）、`onebot11.face` extension、能力声明 | T-15 步骤 3、4 | T-03.6、T-03.8、T-15.1 | V-71 的发送与媒体分项 | 未开始 |
| T-15.3 | goal-80 | 跨 Adapter 一致性验收与 OneBot 11 手测记录，汇总证据 | T-15 步骤 5 | T-03.10、T-15.2 | V-71、V-72；真实 QQ 端只手测 | 未开始 |
| T-15.4 | goal-114 | 官方 OneBot 11 Adapter 的反向 WebSocket：`mode: reverse` 与 `self_id`，注册 `/ws/` 路由，握手鉴权、403 与 409，等待连入与账号核对，断开后等待重连；`onebottest` 的反向客户端 | T-15 步骤 6 | T-03.19、T-15.1 | V-95 的 OneBot 11 分项 | 未开始 |
| T-15.5 | goal-115 | 官方 OneBot 11 Adapter 的新增片段与媒体 URL：语音、全体提及 extension、文件与合并转发的收发，入站 face 与 forward；media_ref 的临时媒体 URL、inline_base64 与 `platform_file` | T-15 步骤 7 | T-03.13、T-03.14、T-15.2 | V-89、V-90 的 OneBot 11 分项，V-96 的入站片段分项 | 未开始 |
| T-15.6 | goal-116 | 官方 OneBot 11 Adapter 的扩展事件、平台动作与撤回：十种扩展事件的声明与映射，动作声明与 `adapters/onebot11/action` 常量包，delete_msg | T-15 步骤 8 | T-03.15、T-03.16、T-15.2 | V-91、V-92 的 OneBot 11 分项，V-96 的扩展事件分项 | 未开始 |

## 3. 工作包

### T-15：官方 OneBot 11 Adapter

**目标：** 提供第二个官方 Adapter，以正向 WebSocket 接入 OneBot 11 实现端；它与 Telegram 在按钮、编辑、提及、媒体与专属内容上的差别用来检验通用 Segment 与 Capability 的设计（CONTRACTS §8、MESSAGE-MODEL §7，D-46）。

#### 前置输入

- T-03 的 Adapter 进程与 `sdk/adapter`、入站媒体与 Grant、extension 声明与校验，以及 Telegram 的模拟服务与故障矩阵（一致性验收需要）。
- 读取 CONTRACTS §8、MESSAGE-MODEL §2、§4、§7 与 OneBot 11 标准原文。依赖 `github.com/coder/websocket` 只用于 `adapters/onebot11/`。真实 QQ 端依赖非官方协议实现，有账号风险，只由维护者手动测试；CI 与验收使用本地模拟服务。

#### 实施步骤

1. **连接与账号。** `adapters/onebot11/` 为每个实例建立一条正向 WebSocket 连接（`mode` 只接受 `forward`），握手带 `access_token`；连接后调用 `get_login_info` 取得 account_id。握手 401/403 与字符串格式上报使实例致命失败并经 `core.adapter.status` 报告；断线按有界退避重连，重连不重放已提交的 Send。API 调用按 `echo` 配对，读循环同时处理响应、心跳与事件。实现放在可导入的包中（`onebot11.New()`），程序入口只调用 `sdk/adapter` 的运行函数。
2. **接收与标准化。** 事件进入实例内 256 条的有界队列，按到达顺序逐条交付，收到应答后才交付下一条；队列满时丢弃并计数。按 MESSAGE-MODEL §7.1 映射私聊/群聊消息、撤回与成员变动，chat_id 写作 `private:<id>` 与 `group:<id>`，匿名消息不构造 Actor，无通用对应的段只留在 platform_payload；`self_id` 不符的事件丢弃并记诊断；to_me 按私聊、mention 本账号与 `get_msg` 查得的引用发送者计算。同时实现本地 OneBot 11 模拟服务（`adapters/onebot11/onebottest`），可注入上报、记录 API 调用并控制响应。
3. **发送计划与回执。** 按 MESSAGE-MODEL §7.2 声明能力并生成物理操作：连续的 text、mention、image 与 extension 合成一条消息，video 单独成一条，reference 只加在第一条前；ok、async、failed、写出后断线与超时分别得到 ACCEPTED、ACCEPTED 加 message_ids_unavailable、REJECTED、UNKNOWN，遇到非 ACCEPTED 即停止并把其余部分记为 NOT_ATTEMPTED。
4. **媒体与 extension。** 入站 image、video 签发 IncomingMedia token，`adapter.media.read` 时用段中的 `url` 取件（只允许 http 与 https）；出站 https_url 原样传递；media_ref 与 incoming_ref（Core 取件后的 media_ref）在实例配置了 `media_dir` 时经 `core.media.read` 写入共享目录并以 `file://` 传递，没有配置时在首个平台请求前返回 UNSUPPORTED，临时媒体 URL 由步骤 7 加入；不把任何来源转成 base64（D-88）。注册 `onebot11.face` v1 extension 及其 Schema。
5. **一致性验收与手测。** 让同一个插件经 `alonabot run` 同时连接 Telegram 与 OneBot 11 两个模拟服务，按 TEST-MATRIX V-72 的同一组 Chain 与入站消息比较两边的结果，并证明未声明的能力在首个平台请求前被拒绝。维护者用真实 OneBot 11 实现端与测试 QQ 账号手测时，记录实现端名称、版本与结果；没有手测时记为未运行。
6. **反向 WebSocket（D-94）。** 按 CONTRACTS §8 与 MESSAGE-MODEL §7.3 实现 `mode: reverse`：实例配置校验 access_token 与 self_id 必填；Adapter 在 REGISTER 中注册公开的前缀路由 `/ws/`，按路径中的 instance_id 分派连接，依次检查实例存在、access_token、X-Self-ID 与 X-Client-Role、已有连接；连入后核对 get_login_info。断开后实例保持运行并等待重连，未连接时 Send 与动作返回 UNAVAILABLE。`onebottest` 增加以反向客户端身份连入的模式。
7. **新增片段与媒体 URL（D-88～D-91、D-95 的入站片段）。** 按 MESSAGE-MODEL §7.1、§7.2 收发语音（record）、文件（upload_group_file、upload_private_file）与合并转发（send_group_forward_msg、send_private_forward_msg），注册 extension `onebot11.at_all` 并收发全体提及；入站 face 为 extension 片段、forward 为只带 platform_forward_id 的 forward；media_ref 在没有 `media_dir` 时经 `core.media.url` 取得临时媒体 URL，inline_base64 原样以 `base64://` 传递，`platform_file` 原样作为 file。追加 `message.voice`、`message.forward` 能力。
8. **扩展事件、平台动作与撤回（D-92、D-93、D-95）。** 在 Adapter 描述中声明 MESSAGE-MODEL §7.1 的十种扩展事件及 Schema 并完成映射；声明 §7.4 的动作及参数、结果 Schema，在 `adapters/onebot11/action` 为每个动作发布带类型的 Go 函数（只依赖 `sdk` 根包）；`adapter.delete` 逐个调用 delete_msg，追加 `message.delete` 能力。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `adapters/onebot11/` | 官方 OneBot 11 Adapter：连接、接收、标准化、发送计划、媒体与 extension |
| `adapters/onebot11/onebottest/` | 本地 OneBot 11 模拟服务 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-71 | 连接、鉴权、重连、格式检查、事件映射、to_me、队列上限、发送计划、媒体与各种响应的回执 |
| V-72 | 同一组 Chain 与入站消息在两个 Adapter 上的结果对照，以及未声明能力在平台请求前被拒绝 |
| V-89～V-92、V-95、V-96 | 步骤 6～8 的 OneBot 11 分项 |

反向 WebSocket 由 D-94 纳入 v0.1（步骤 6）。

#### 失败停止条件

- 为了 OneBot 11 修改通用 Segment 的语义，或让 Core 按 adapter_id 分支，先回到契约讨论，不合入。
- 不支持的片段被静默丢弃、降级或拆成多次 Send，阻止合入。
- access_token 出现在日志、错误或普通响应中，立即修复。
- 插件没有给出 inline_base64 时请求中出现 `base64://`，或 media_dir 中的临时文件在物理操作结束后未删除，阻止合入。
