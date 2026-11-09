# 实施计划：适配器 · Telegram

总目标、章节划分与跨章节依赖见 [PLAN](../../PLAN.md)；本章节的 goal 语句见 [telegram.md](../../../docs/goals/adapters/telegram.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本子章节覆盖官方 Telegram Adapter `adapters/telegram/` 的平台实现：Bot API 客户端、接收与标准化、发送计划与回执、平台权限，以及本地 Bot API 模拟服务 `adapters/telegram/telegramtest/`。章节总述与扩展约定见[适配器实施计划](README.md)。

- 本子章节的切片：T-03.4、T-03.7、T-03.11，D-87 对标补齐的 T-03.20～T-03.22（协议与 Core 侧在核心切片 T-03.13～T-03.17），以及代理模式与话题判定的 T-03.23（D-100、D-101）。
- 核心切片中的 Telegram 部分：T-03.6（入站媒体的 getFile 取件）、T-03.8（callback_query 与 answerCallbackQuery）、T-03.9（editMessageText 与 `message.edit` 能力）、T-03.10（模拟服务上的发送与接收故障矩阵）。这些切片随 Core 功能一起交付，进度记在[核心实施计划](../core.md)。
- 跨章节依赖：核心切片 T-03.8、T-03.9 以 T-03.7 为前置，需要用 Telegram 的发送计划渲染按钮与编辑。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-03.4 | goal-5 | 官方 Telegram Adapter 的接收：Bot API 客户端、本地 Bot API 模拟服务、getMe、webhook 冲突、long polling、逐条交付与游标、7 天轮次重建、401/409、媒体组、Actor.locale、`to_me`、实例代理 | T-03 步骤 4 | T-03.3 | V-45、V-54 的 locale 分项，V-62 的 Bot Token URL 分项 | 已完成（PR #39；TEST-MATRIX §3.2） |
| T-03.7 | goal-50 | 官方 Telegram Adapter 的发送计划与回执：Chain 完整预检、物理操作计划、UTF-16 实体、相册/引用/按钮渲染、PartResult 汇总 | T-03 步骤 6（发送） | T-03.4 | V-08、V-35 | 已完成（[PR #108](https://github.com/AlonaBot/AlonaBot/pull/108)） |
| T-03.11 | goal-89 | 官方 Telegram Adapter 的平台权限：`telegram.chat_owner`/`telegram.chat_admin` 的声明、`getChatAdministrators` 的会话缓存与作废、获取失败时不带权限，`adapters/telegram/perm` 常量包 | T-03 步骤 9 | T-04.6 | V-85 的 Telegram 分项 | 已完成（[PR #106](https://github.com/AlonaBot/AlonaBot/pull/106)） |
| T-03.20 | goal-111 | 官方 Telegram Adapter 的媒体与新增片段：按来源传递与流式 multipart、`platform_file` 复用 file_id、voice 的收发、动图、贴纸与 video_note、url 按钮、`file_base_url`、入站事件的 chat_type | T-03 步骤 17 | T-03.13、T-03.14 | V-89、V-90、V-98 的 Telegram 分项，V-97 | 未开始 |
| T-03.21 | goal-112 | 官方 Telegram Adapter 的平台动作与撤回：动作声明（含 sendChatAction、setMessageReaction、sendMessageDraft）与 `adapters/telegram/action` 的带类型函数及输入状态、草稿两个辅助函数，deleteMessages | T-03 步骤 18 | T-03.11、T-03.15、T-03.16 | V-91、V-92、V-94 的 Telegram 分项 | 未开始 |
| T-03.22 | goal-113 | 官方 Telegram Adapter 的命令菜单：setMyCommands 同步 | T-03 步骤 19 | T-03.17 | V-93 的 Telegram 分项 | 未开始 |
| T-03.23 | goal-117 | 官方 Telegram Adapter 的代理模式与话题判定：实例配置 `proxy_mode`（none 缺省、core、custom），thread_id 只认 `is_topic_message` | T-03 步骤 20 | T-03.4 | V-99、V-100 | 未开始 |

## 3. 工作包

### T-03 的 Telegram 步骤

以下步骤原属 T-03，编号保持不变。T-03 的前置输入、拟目录、验收（V-08、V-35、V-45、V-53、V-54）与失败停止条件见[核心实施计划的 T-03](../core.md#t-03标准消息可安装-adapter-与官方-telegram-adapter)。

4. **接通 Telegram 接收与生命周期。** 在官方 Telegram Adapter 程序中实现 getMe 的账号/用户名绑定、已有 webhook 冲突检查、后台 long polling 和经 `core.adapter.status` 的状态上报；同账号唯一 owner 由 Core 检查。显式设置规定的 allowed_updates；只生成平台实际给出的事件，`sender_chat` 的兼容 fake sender 不构造个人 Actor。每条 update 经 `core.adapter.deliver` 交给 Core，Core 先持久化游标再投入非持久队列，Adapter 收到应答后才处理下一条；满队列丢弃计数且继续推进，不能提前提交整个 batch 的最大 ID。实现契约规定的连续 7 天空闲接收轮次重建。短暂读失败使用有界退避，401/409 停止该实例并报告故障；Start、Stop 和重连均不得创建重复接收 owner。入站媒体组保留 media_group_id，逐 Message 派发。Actor.locale 取 from.language_code，message.created 的 `to_me` 按 MESSAGE-MODEL §6.1 计算。实例配置的 `proxy` 优先于全局代理，作用于该实例的全部 Bot API 请求与文件下载；代理失败按网络错误退避，不改为直连。Adapter 实现放在可导入的包中（`telegram.New()`），程序入口只调用 `sdk/adapter` 的运行函数，供库模式直接引用。

6. **实现可预测的发送计划及回执。** 在官方 Telegram Adapter 中，每条 Chain 先完成全部结构、目标能力、媒体可用性与平台限制检查，再生成按顺序执行的物理操作列表。落实文本/实体 UTF-16 偏移、单媒体/2～10 项相册、首条引用、末条按钮（末条为相册时按钮无法表达，首发前返回 UNSUPPORTED）；不静默截断、重排或降级。每个操作记录 PartResult，明确 ACCEPTED、REJECTED、UNKNOWN、NOT_ATTEMPTED，并据事实汇总逻辑状态。遇明确拒绝或结果未知停止后续操作；保留已确认 ID，同时通过 Envelope 返回有效 payload 和错误。平台确认但 message_id=0 时不伪造可编辑 ID。显式空 Chain 和仅有元数据的 Chain 在调用平台前拒绝。

9. **接通 Telegram 的平台权限。** 按 MESSAGE-MODEL §6.1 在 Adapter 描述中声明 `telegram.chat_owner` 与 `telegram.chat_admin`；群与超级群中的事件按 `getChatAdministrators` 的会话缓存给出 Actor 的平台权限，缓存 10 分钟，收到该会话的 `chat_member`、`my_chat_member` 时作废；获取失败时本事件不带平台权限并记诊断。发布只含常量的 `adapters/telegram/perm`。

17. **媒体与新增片段（D-88、D-89、D-97、D-99）。** 按 MESSAGE-MODEL §6.4 改写发送计划与执行：multipart 请求体改为边经 `core.media.read` 分块读取边写出，不在内存中拼出整份请求；inline_base64 解码后上传，`platform_file` 直接作为 file_id。voice 用 sendVoice，单独成一个物理操作；image/gif 用 sendAnimation 并结束当前媒体组；url 按钮映射为 InlineKeyboardButton 的 url。入站 voice、animation、video_note 与静态贴纸按 §6.4 映射，入站 audio 仍为 file。实例配置追加 `file_base_url`，文件下载经它发出，代理规则不变。入站事件按 MESSAGE-MODEL §6 由 chat.type 填写 chat_type（D-100）。在 Capabilities 中追加 `message.voice`；`telegramtest` 追加对应方法的记录与响应。

18. **平台动作与撤回（D-92、D-93、D-98）。** 在 Adapter 描述中声明 MESSAGE-MODEL §6.5 的动作及其参数、结果 Schema（含 send_chat_action、set_message_reaction、send_message_draft），在 `adapters/telegram/action` 为每个动作发布带类型的 Go 函数（只依赖 `sdk` 根包），并提供输入状态与草稿流式两个辅助函数；实现 deleteMessages（同一会话每次最多 100 个），追加 `message.delete` 能力。限权相关动作执行后作废该会话的管理员缓存（T-03.11）。sendMessageDraft 的参数在实施时按 Bot API 原文核对，与本计划不符时标阻塞。

19. **命令菜单（D-96）。** `adapter.commands.sync` 按 MESSAGE-MODEL §6.5 调用 setMyCommands，记住上一次成功同步的内容，相同时不调用平台；追加 `commands.sync` 能力。

20. **代理模式与话题判定（D-100、D-101）。** 实例配置追加 `proxy_mode`，按 CONTRACTS §8 为该实例的 HTTP 客户端选择直连、Core 的代理环境或实例的 `proxy`，缺省直连；custom 缺少或给出不合法的 `proxy` 时启动返回 INVALID_ARGUMENT。入站 Message 只在 `is_topic_message` 为 true 时把 message_thread_id 写入 Target 的 thread_id（MESSAGE-MODEL §6.1）。
