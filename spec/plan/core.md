# 实施计划：核心

总目标、章节划分与跨章节依赖见 [PLAN](../PLAN.md)；本章节的 goal 语句见 [core.md](../../docs/goals/core.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本章节覆盖 Core（`internal/core/`、`internal/wire/`、`internal/tune/`）、协议契约（`contracts/`）、插件与 Adapter SDK（`sdk/`，含 `sdk/adapter`、`sdk/generate` 与插件测试套件）、`cmd/alonabot`、库模式根包，以及作为 SDK 验收手段的 `examples/`。Adapter 的接口层面（`sdk/adapter`、`internal/core/adapter` 与 `remote`、`adapter.*` 与 `core.adapter.*` 服务、能力与 extension 声明的校验）属于本章节；各平台的实现在适配器章节。

- 工作包：T-01、T-02、T-03（Telegram 步骤除外）、T-04～T-08、T-09、T-11、T-13、T-16，以及 PLAN §7 的 S-01～S-03、S-06。T-11 是 Core 提供给任何 Agent 组件的工具目录、调用确认与会话存储，所以放在本章节。
- 含 Telegram 部分的核心切片：T-03.6（getFile 取件）、T-03.8（callback_query 与 answerCallbackQuery）、T-03.9（editMessageText）、T-03.10（Bot API 模拟服务上的故障矩阵）。这些切片用官方 Telegram Adapter 作为首个真实实现来验收 Core 功能，[Telegram 实施计划](adapters/telegram.md)同时列出它们。
- `web/admin` 的前端切片（T-08.7、T-09.8、T-09.9、T-09.15）归 WebUI 章节，步骤文字仍在本章节的 T-08 步骤 9 与 T-09 步骤 7。页面调用的 Admin API 与 WebUI 静态文件的提供（T-09.10、T-09.12～T-09.14）属于本章节。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-02.1 | goal-65 | 日志基线：Core 与 SDK 改用 slog（Core 由装配处、SDK 由 Run 设置默认 Logger），控制台与 JSON 格式、来源标签、插件 stderr 的 JSON 解析与非结构化行、panic 堆栈、Bootstrap 下发级别与进程内过滤、`--log-level`/`--log-format`、不输出 Bootstrap 与配置明文；按日期轮转的日志文件与保留上限 | T-02 步骤 9 | T-02 | V-62、V-63 的级别过滤分项、V-68 的写入与轮转分项 | 已完成（PR #13），darwin/arm64 运行通过（TEST-MATRIX §3.1） |
| T-02.2 | goal-76 | 调优参数：`internal/tune` 参数表，把现有常量迁入；`--args` 的直接参数与文件输入、校验与相互约束、启动 WARN、`alonabot tune list`；`SessionLimits session_limits = 11`（max_inflight、max_payload_bytes、request_timeout_ms）；`GOMEMLIMIT` 软上限 | T-02 步骤 10 | T-02.1 | V-69 | 已完成（PR #28），darwin/arm64 运行通过（TEST-MATRIX §3.2） |
| T-02.3 | goal-86 | 状态库改用 GORM：现有各表的读写改为 GORM 模型，经 `libtnb/sqlite` 复用同一单连接，schema 仍由唯一的初始迁移定义，锁、PRAGMA 核对与损坏不重建的行为不变 | T-02 步骤 11 | T-02 | V-36 的状态库分项与现有状态库集成测试的回归、V-75 | 已完成（PR #63） |
| T-02.4 | goal-103 | 移除绑定 SDK API 的提交测试：删除 `sdk/` 全部测试、`internal/core/supervisor/` 除 `supervisor_test.go` 外的测试与整个 `testdata/`（D-86） | T-02 步骤 12 | — | 无（只删除测试，V 编号的既有记录保留） | 已完成（PR #99） |
| T-02.5 | goal-149 | CRITICAL 日志级别：LogLevel 枚举追加 CRITICAL、SDK 的 `sdk.LevelCritical` 与输出文本、Core 的解析、过滤、控制台与最近及历史记录、管理 API 与 `log.level`/`log.sources`/`--log-level`/`log-level` 接受 critical；启动诊断与状态库、设置写回、提交后恢复及故障退出按 CONTRACTS §14 分级（D-158） | T-02 步骤 9（日志） | T-02.1、T-09.10 | V-129 | 已完成（PR #209） |
| T-03.1 | goal-2 | 首个可用 SDK 切片：事件路由注册、Context/Send、业务分发；Core 的 Adapter 实例接口、内存 Adapter、固定服务 mux、事件派发与 `core.send`；echo 插件经真实 REGISTER/ACTIVATED 后收发 | T-03 步骤 1 | T-02.1、T-02.2 | V-07、V-08 的 SDK 分项，V-62 的 Handler 分项 | 已完成（PR #29；TEST-MATRIX §3.2） |
| T-03.2 | goal-3 | Target 与最小授权：`allowed_adapter_instances` 持久化、显式 To、后台 Send、逐次发送准入、NO_DEFAULT_TARGET、Core 内部许可读写入口 | T-03 步骤 3 | T-03.1 | V-07 | 已完成（PR #31；TEST-MATRIX §3.3） |
| T-03.3 | goal-4 | Adapter 进程与 `sdk/adapter`：Adapter 角色的注册校验、实例启停、`adapter.send`、事件交付与游标持久化、事件归属校验、账号唯一；实例上下线时生成 `bot.online`/`bot.offline`；回环测试 Adapter | T-03 步骤 2 | T-03.1 | V-48 的进程与实例分项、V-51 | 已完成（PR #32；TEST-MATRIX §3.2） |
| T-03.5 | goal-6 | 媒体上传与配额：BeginUpload/Chunk/Commit、大小/MIME/SHA256 校验、全部固定配额与期限、owner 退出释放 | T-03 步骤 5（上传） | T-03.1 | V-09 | 已完成（PR #38；TEST-MATRIX §3.2） |
| T-03.6 | goal-7 | 入站媒体与 Grant：IncomingMedia token、经 `adapter.media.read` 首次取件并校验缓存、多 Session 共享、Grant 范围、租约；出站媒体经 `core.media.read` 交给 Adapter | T-03 步骤 5（入站与授权） | T-03.2、T-03.4、T-03.5 | V-44、V-09 | 已完成（PR #55；TEST-MATRIX §3.2） |
| T-03.8 | goal-8 | 按钮 token 与回调路由，带类型的按钮动作 `sdk.Action[T]`（D-107）；extension 片段的声明与校验；manifest 的 `role`（plugin/adapter/agent/webui）、`adapter_id`、`icon`、`adapter_dependencies`、`plugin_dependencies` 与 WebUI 字段的解析，开发入口的依赖检查 | T-03 步骤 7 | T-03.2、T-03.7 | V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项、V-105 | 已完成（PR #136；TEST-MATRIX §3.2） |
| T-03.9 | goal-51 | 消息编辑：`core.edit`、`adapter.edit`、Telegram editMessageText、`message.edit` 能力、出站 Hook 的 edit 输入 | T-03 步骤 6（编辑） | T-03.7 | V-53 | 已完成（PR #118；TEST-MATRIX §3.2） |
| T-03.10 | goal-9 | 开发入口：用 `alonabot run` 运行 Telegram Adapter 与示例插件（实例写在 `core_config.json`），echo/media 示例与发送接收故障矩阵，汇总证据 | T-03 步骤 8 | T-03.6、T-03.8、T-03.9、T-07.6、T-13.3 | V-07、V-08、V-09、V-35、V-44、V-45、V-48、V-51～V-54 的 T-03 分项；真实 Telegram 需要测试账号 | 已完成（PR #180；TEST-MATRIX §3.2） |
| T-03.12 | goal-99 | 后台发信许可缺省为全部实例：`allowed_adapter_instances` 改为「全部实例」或管理员收紧后的实例集合，准入与撤销语义不变（D-83） | T-03 步骤 3 | T-03.2 | V-07 | 已完成（PR #90；TEST-MATRIX §3.2） |
| T-03.13 | goal-104 | 媒体按来源传递：MediaSource 的 `inline_base64` 与只在 Core→Adapter 方向出现的 `platform_file`，同一实例的入站对象复用平台文件标识，`core.media.url` 与统一端口上的临时媒体 URL，SDK 字节构造函数的内联与上传选择（D-88） | T-03 步骤 10 | T-03.6、T-04.8、T-07.7 | V-89 的 Core 分项 | 已完成（PR #148；TEST-MATRIX §3.2） |
| T-03.14 | goal-105 | 新增片段与字段：voice 片段、forward 的 platform_message_id 与 platform_forward_id、Button 的 url、Event 的 chat_type；Chain 校验与能力判定，按钮签发跳过 url 按钮，入站事件的 chat_type 校验，SDK 构造函数、入站类型与会话类型方法（D-89、D-91、D-97、D-100、D-110） | T-03 步骤 11 | T-03.8、T-04.8 | V-90、V-98 的 Core 分项，V-108 的语音分项 | 已完成（PR #143；TEST-MATRIX §3.2） |
| T-03.15 | goal-106 | 平台动作：Adapter 描述的动作声明与校验，`core.action` 与 `adapter.action`，目标准入与参数、结果 Schema 校验，`adapter_dependencies` 的动作检查，SDK 的底层动作调用方法、供 Adapter 包带类型函数使用的调用者接口与 `sdk/adapter` 的可选接口（D-92），动作以参数与结果的 Go 类型声明（D-113） | T-03 步骤 12 | T-03.8、T-04.8 | V-91 的 Core 分项，V-111 的动作分项 | 已完成（PR #146；TEST-MATRIX §3.2） |
| T-03.16 | goal-107 | 撤回：`core.delete` 与 `adapter.delete`，`message.delete` 能力判定与准入，每个 ID 的结果汇总，SDK 方法只收消息句柄（D-93、D-104） | T-03 步骤 13 | T-03.9、T-04.8 | V-92 的 Core 分项、V-102 的撤回分项 | 已完成（PR #144；TEST-MATRIX §3.2） |
| T-03.17 | goal-108 | 命令菜单同步：`adapter.commands.sync`，按实例筛选命令与翻译说明，2 秒合并与失败诊断；本切片交付英文 description 的 proto 字段、SDK builder、commands 路径翻译、生成器结构检查与 Collect 覆盖检查，T-14.1 复用（D-96、D-138） | T-03 步骤 14 | T-04.6、T-04.7、T-04.8、T-08.8 | V-93 的 Core 分项 | 已完成（PR #150；TEST-MATRIX §3.2） |
| T-03.18 | goal-109 | SDK 的流式输出辅助函数 `c.Stream`：按 `message.edit` 选择节流编辑或缓冲后发送一次，按 `message.send` 声明的上限与单位换条，未声明时不换条，失败即停止；追加能力元数据与官方 Telegram 声明（D-98、D-135） | T-03 步骤 15 | T-03.9、T-04.8 | V-94 的 Core 分项 | 已完成（PR #151；TEST-MATRIX §3.2） |
| T-03.19 | goal-110 | Adapter 的 HTTP 路由：Adapter 的 REGISTER 携带 HTTP 路由与 endpoint，`sdk.RunAdapter` 托管 Listener，统一端口代理 WebSocket Upgrade 并在路由撤销时关闭连接，Adapter 路由保留 Authorization 交 Adapter 校验（D-94、D-136） | T-03 步骤 16 | T-03.3、T-04.8、T-07.7 | V-95 的 Core 分项 | 已完成（PR #152；TEST-MATRIX §3.2） |
| T-03.24 | goal-121 | 插件给出的目标由 Core 补齐：缺少实例时按在途事件的同一会话或 Adapter 的缺省实例选择，account_id 按实例补上；`core.send` 的 To 与引用片段、`core.media.grant` 的 targets 在准入前补齐；`core_config.json` 的实例名规则与 `default_instance`（D-102） | T-03 步骤 21 | T-03.6、T-07.6 | V-101 的 Core 分项 | 已完成（PR #135；TEST-MATRIX §3.2） |
| T-03.25 | goal-131 | 聊天目录：Event 追加 `chat_title`，Core 在接纳入站事件时记录聊天（键、chat_type、名称、最近的实例与活跃时刻），内置选项来源 `core.chats`，随 Adapter 的数据清理（D-130） | T-03 步骤 22 | T-03.3、T-08.14、T-09.4 | V-117 的 Core 分项 | 已完成（PR #166；TEST-MATRIX §3.2） |
| T-04.1 | goal-10 | 根插件 API 与显式注册；静态注册索引、重复拒绝、跨插件遮蔽的稳定选择 | T-04 步骤 1、3 | T-03.1 | V-10 的注册分项、V-02 的注册原子性 | 已完成（PR #34；TEST-MATRIX §3.2） |
| T-04.2 | goal-11 | 构建期生成器 `sdk/generate`：限定扫描范围、build tags、生成注册代码、稳定输出；`cmd/alonabot` 的构建检查入口 | T-04 步骤 2 | T-04.1 | V-10 | 已完成（PR #42；TEST-MATRIX §3.2） |
| T-04.3 | goal-12 | 命令选择与参数绑定：完整 token、@username、最长路径/别名→文本匹配（Regex、keyword、startswith、endswith、fullmatch）→Raw、消费状态、参数语法、usage/silent | T-04 步骤 4 | T-04.1 | V-11 的命令分项、V-49 | 已完成（PR #47；TEST-MATRIX §3.2） |
| T-04.4 | goal-13 | 持久化命令策略与管理服务层：继承、禁用、全局禁止、下一次派发生效 | T-04 步骤 5 | T-03.2、T-04.1、T-04.6 | V-12、V-43 的服务层分项 | 已完成（PR #88；TEST-MATRIX §3.2） |
| T-04.5 | goal-14 | 用同一 SDK 升级示例，完成构建与策略验收，汇总证据 | T-04 步骤 6、7 | T-03.10、T-04.2、T-04.3、T-04.4、T-04.7、T-04.8、T-04.9、T-04.11、T-04.12、T-04.13 | V-10、V-11、V-12、V-43、V-49 的 T-04 分项 | 已完成（PR #191；TEST-MATRIX §3.2） |
| T-04.6 | goal-87 | 权限、扩展事件与命令解析的协议和 Core 侧：proto 修改与生成；Adapter 权限与扩展事件声明的校验；平台权限的保留与隐含展开、按权限名集合判定；扩展事件的交付校验；`command_prefixes` 与按实例覆盖、去掉开头的 reference 与本账号 mention、非文本 token 与 `user`/`image`/`message` 参数、usage 文案键 | T-04 步骤 8 | T-03.3、T-03.6、T-04.3 | V-85、V-86、V-87 的 Core 分项 | 已完成（PR #78；TEST-MATRIX §3.2） |
| T-04.7 | goal-88 | SDK 链式注册：命令、文本匹配、事件、Service 与 Event Bus 订阅的 builder，Permission 值对象与 `c.Has`，Service Schema 由生成器生成且 schema_id 由 Go 类型决定，自动回复型可 `c.Send`；生成器、示例与已合并调用方同步 | T-04 步骤 9 | T-04.2、T-04.6、T-06.1、T-06.3、T-08.1 | V-10、V-14 的 SDK 分项 | 已完成（PR #89；TEST-MATRIX §3.2） |
| T-04.8 | goal-91 | SDK 自有类型与事件常量：Message/会话字符串与可选的 `sdk.Via`（D-102）/消息句柄（D-104）/片段构造函数与单片段发送（D-108）/媒体读取与上传助手（D-109）/时间与时长用 `time.Time`、`time.Duration`（D-110）/Adapter 声明用 Go 类型（D-113）/Actor/SendResult/错误码与入站片段，标准事件常量与 `sdk.AdapterEvent`，`Collect` 核对事件 Handler 的 payload 类型，`sdk/adapter` 的公共类型；Telegram Adapter、示例与已合并调用方同步 | T-04 步骤 9 | T-03.9、T-03.24、T-04.7、T-04.9 | V-14 的类型分项、V-86 的 SDK 分项，V-101、V-102 的 SDK 分项，V-106～V-108，V-111 的事件与 extension 分项，V-113 的 SDK 类型分项（`c.Reply`、用户字符串、按来源读取与错误码辅助函数，D-118、D-120、D-121、D-123） | 已完成（PR #141；TEST-MATRIX §3.2） |
| T-04.9 | goal-96 | 删除构建期标记：去掉 `//alonabot:command`、`//alonabot:event` 的扫描、诊断与生成，命令与事件只经链式注册，示例与测试随之改写（D-79） | T-04 步骤 2 | T-04.7 | V-10 | 已完成（[PR #110](https://github.com/AlonaBot/AlonaBot/pull/110)） |
| T-04.10 | goal-122 | 管理员与封禁按 Adapter 登记：`admins` 改用用户字符串，禁止表按 adapter_id，对该 Adapter 的全部实例生效（D-105） | T-04 步骤 10 | T-04.4、T-07.6 | V-103 | 已完成（PR #133；TEST-MATRIX §3.2） |
| T-04.11 | goal-124 | 注册写法统一：定时任务与 HTTP 路由的链式注册，命令省略 `Args` 时参数名取 Handler 的参数名（生成器记录），Event Bus 订阅的路由 ID 缺省为事件名（D-112） | T-04 步骤 11 | T-04.8、T-04.12、T-07.5 | V-110 的 SDK 分项 | 已完成（PR #159；TEST-MATRIX §3.2） |
| T-04.12 | goal-125 | 插件对象、结构体参数与子插件：入口改为 `func New() *sdk.Plugin`，路由、生命周期回调、`OnConfigLoad` 与数据库迁移登记在插件对象上，删除 `sdk.Base`、`sdk.Root`、根插件的六个注册方法（`Commands`、`Events`、`Services`、`Subscriptions`、`HTTP`、`Schedules`）与对应的注册器类型（D-114）；命令 Handler 可以用一个结构体接收参数（D-115）；一层逻辑子插件及其缺省权限、配置视图与日志字段（D-116）；包内 `alonabotGenerated` 经 `sdk.UseGenerated(p, g)` 显式绑定插件对象，独立进程与库模式共用（D-137）；生成器、示例与已合并的调用方同步 | T-04 步骤 12 | T-04.8 | V-112 | 已完成（PR #154；TEST-MATRIX §3.2） |
| T-04.13 | goal-126 | Service 与 Event Bus 的带类型声明：`sdk.JSON`、`sdk.Proto` 返回 `sdk.Service[In, Out]`，`sdk.Serve` 注册、`Call` 调用；`sdk.BusEvent[T]` 的 `Publish` 与 `sdk.Subscribe`；删除 `sdk.CallJSON`、`sdk.CallProto`、`sdk.Publish` 与旧形式的注册、订阅函数；`sdk/llm`、示例与已合并的调用方同步（D-122） | T-04 步骤 13 | T-04.12 | V-113 的声明分项 | 已完成（PR #158；TEST-MATRIX §3.2） |
| T-04.14 | goal-129 | 命令名覆盖的 Core 侧：命令策略保存按路由的命令名列表，Router 的索引、遮蔽诊断、usage 回复与命令菜单使用生效的命令名，保存时的格式与重名检查，插件更新后的失效处理（D-128） | T-04 步骤 14 | T-03.17、T-04.10 | V-115 的 Core 分项 | 已完成（PR #156；TEST-MATRIX §3.2） |
| T-04.15 | goal-138 | SDK 公开面收敛：会话、消息句柄与用户字符串改为 string，新增 `c.SendTo`，删除 `sdk.To`（D-144）；删除已提交的生成文件、`sdk.UseGenerated` 等生成相关 API 与 `//go:generate`，资源改由 CLI 构建经 overlay 或由 plugintest 现场扫描按入口包登记，库模式改由 `alonabot build` 构建，生成器移入 `internal/generate`（D-145）；Adapter 的 HTTP 路由写进 `adapter.Descriptor`，删除 `sdk.AdapterHTTP`（D-147）；示例只在 `New()` 中注册 | T-04 步骤 15 | T-04.5、T-09.3、T-16.1 | V-120 的会话、生成资源与 Adapter 路由分项 | 已完成（PR #197；TEST-MATRIX §3.2） |
| T-04.16 | goal-139 | SDK 按类别分包：动态任务、Service、Event Bus、按钮、媒体、会话历史、工具、平台动作与选项来源移入 `sdk/task`、`sdk/service`、`sdk/bus`、`sdk/button`、`sdk/media`、`sdk/conversation`、`sdk/tool`、`sdk/action`、`sdk/options`，根包只留插件对象、Context、注册 builder、权限与运行入口，`Collect` 等改为内部入口（D-146） | T-04 步骤 16 | T-04.15 | V-120 的分包分项 | 已完成（PR #199；TEST-MATRIX §3.2） |
| T-04.17 | goal-150 | 插件页面套件：`sdk/page` 的 `kit.css` 与 `kit.js`（桥的 hello、请求、异步文案、令牌应用、主题与语言事件、⌘K 转发），以及 showcase 插件页面使用套件（D-159） | T-04 步骤 17 | T-04.16、T-07.7 | V-130 的 SDK 分项 | 已完成（PR #218） |
| T-04.18 | goal-161 | 配置文件的派发与 SDK：`profile_id`、覆盖值下发、`c.Config()` 合并、按配置文件的插件开关、`core.profile.*`（D-168） | T-04 步骤 18 | T-04.16、T-08.18 | V-141 | 已完成（PR #242） |
| T-04.19 | goal-162 | 平台与消息：命令前缀按配置文件、删除按实例的前缀、回复的提及、引用、前缀与合并转发（D-168） | T-04 步骤 19 | T-04.18 | V-142 | 已完成（PR #250） |
| T-04.20 | goal-178 | SDK 用法从简：文本匹配、Raw、事件路由与 Hook 的注册去掉 route ID 参数并增加 `.ID`，`*sdk.Context` 实现 `context.Context`，`c.Reply` 可带片段，`.Page(title)` 自动登记套件资源，`plugintest.Result.RequireReply`；全部调用方与示例同步（D-180） | T-04 步骤 20 | T-04.18、T-16.2 | V-148 | 已完成（PR #251） |
| T-05.1 | goal-15 | 可信事件上下文与有界同事件分发：64 条并发链、256 条/8 MiB 等待队列、稳定候选排序、链执行与读循环分离；经回环测试 Adapter 运行，按钮与 scheduled 绑定先用规范 fixture | T-05 步骤 1、2 | T-03.3、T-04.3、T-04.4 | V-13、V-16 的 fixture 分项 | 已完成（PR #67；TEST-MATRIX §3.2） |
| T-05.2 | goal-16 | SDK Rule（含 `ToMe`）、HandlerOutcome、自动/手动 Send、StopPropagation 锁存与错误分支 | T-05 步骤 3、4 | T-05.1、T-04.7、T-04.8 | V-14、V-11 的 Rule 分项 | 已完成（PR #145；TEST-MATRIX §3.2） |
| T-05.3 | goal-17 | PreEvent/BeforeSend/AfterSend 三阶段 Hook（链式注册与 SDK 类型，D-112），send_forbidden 的保存与继承 | T-05 步骤 5、6 | T-05.2 | V-15 的直接路径分项，V-110 的 Hook 分项 | 已完成（PR #155；TEST-MATRIX §3.2） |
| T-05.4 | goal-52 | 延续调用与多轮等待：`core.continuation.start`、`core.wait`、等待键与消费、超时与 Session 失效；SDK 的 `Continue`（Handler 内调用与 builder 的 `.Continue()`，D-119）与 `Wait` | T-05 步骤 7 | T-05.2 | V-50，V-113 的延续调用分项 | 已完成（PR #153；TEST-MATRIX §3.2） |
| T-05.5 | goal-18 | 多进程事件流水线验收（含真实按钮回调与多轮等待），汇总证据 | T-05 步骤 8 | T-03.8、T-05.3、T-05.4 | V-11、V-13～V-16、V-43、V-49、V-50 的 T-05 分项 | 已完成（PR #162；TEST-MATRIX §3.2） |
| T-06.1 | goal-19 | Service/Bus 描述与编码校验、SDK 入口；Broker 定位与准入 | T-06 步骤 1、2 | T-04.1 | V-19 | 已完成（PR #45；TEST-MATRIX §3.2） |
| T-06.2 | goal-20 | 父调用链上下文传播、循环与深度检查、跨段取消与背压 | T-06 步骤 3、4 | T-03.6、T-05.3、T-06.1 | V-18、V-42 的嵌套分项 | 已完成（PR #161；TEST-MATRIX §3.2） |
| T-06.3 | goal-21 | 非持久 Event Bus：发布准入、逐订阅者 256 条/4 MiB 队列、drop-newest 计数 | T-06 步骤 5 | T-06.1 | V-17 | 已完成（PR #56） |
| T-06.4 | goal-22 | Session 变更时的清理与计数，A/B/C 端到端验收（含 Hook→Service→Send），汇总证据 | T-06 步骤 6、7 | T-06.2、T-06.3 | V-15、V-17、V-18、V-19、V-42 | 已完成（PR #164；TEST-MATRIX §3.2） |
| T-07.1 | goal-23 | interval/五字段 cron/IANA 时区/夏令时的表达式校验与下一次触发计算（纯计算包 `contracts/schedule`，Core 与官方 Agent 组件共用） | T-07 步骤 2 | T-02 | V-20 的计算分项 | 已完成（PR #17；TEST-MATRIX §3.2） |
| T-07.2 | goal-24 | schedule 注册校验，运行槽与调度调用：独立 run_id、15 分钟期限、同 job 跳过重叠、迟到响应与断链 | T-07 步骤 1（schedule）、3 | T-05.1、T-07.1 | V-20、V-46、V-16 和 V-05 的 scheduled 分项 | 已完成（[PR #107](https://github.com/AlonaBot/AlonaBot/pull/107)） |
| T-07.3 | goal-25 | HTTP 注册校验，SDK 托管 HTTP Listener 与 Core Session 代理映射；Core 设置文件 `alonabot.yaml` 的读取、缺省值与命令行覆盖 | T-07 步骤 1（HTTP）、4、5 | T-04.1 | V-22 的组件分项 | 已完成（PR #46；TEST-MATRIX §3.2） |
| T-07.4 | goal-26 | 两个 Host 的边界、插件兑换票据与插件会话 | T-07 步骤 6 | T-07.3 | V-47 的组件分项 | 已完成（PR #53） |
| T-07.5 | goal-27 | 统一启停与清理，时间、进程和 HTTP 边界验收，汇总证据 | T-07 步骤 7、8 | T-07.2、T-07.4 | V-20、V-21、V-22、V-46、V-47 的 T-07 分项 | 已完成（PR #119；TEST-MATRIX §3.2） |
| T-07.6 | goal-97 | Core 设置与首次启动：`core_config.json` 取代 `alonabot.yaml`（启动设置、Core Admin、命令前缀、默认 locale、按来源的日志级别、Adapter 实例），`${环境变量}` 展开、保留原写法的写回与供开发运行调用的重新读取；非终端首次启动把管理 token 写入 `admin_token`，`admin-token show`（D-82） | T-07 步骤 5 | T-07.3、T-09.1 | V-22 的组件分项、V-47 的认证分项、V-88 | 已完成（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)） |
| T-07.7 | goal-98 | 插件页面的沙箱隔离：删除插件 host、兑换票据与插件会话，插件路由响应的 CSP sandbox、nosniff 与 Set-Cookie 过滤，非公开路由的管理会话与 CSRF（D-84） | T-07 步骤 6 | T-07.4、T-07.6、T-09.1 | V-47 的组件分项 | 已完成（PR #116；TEST-MATRIX §3.2） |
| T-07.8 | goal-127 | 动态任务与调度持久化的协议和 Core 侧：proto 追加与生成；状态库的任务表与调度状态表；调度器改为以状态库为准，interval 按保存的起点计算，错过的触发跳过；`core.task.set`/`cancel`/`list` 与 `core.job.pause`/`resume`/`status`；动态任务的并发队列与 Cancel 取消在途调用；卸载清理范围中的任务分项（D-124～D-127） | T-07 步骤 9 | T-07.5、T-09.4 | V-114 的 Core 分项 | 已完成（PR #169；TEST-MATRIX §3.2） |
| T-07.9 | goal-128 | 动态任务与暂停恢复的 SDK：`sdk.Task[T]`、`sdk.OnTask`，实例的 `At`、`Every`、`Cron`、`Cancel`、`List`，`sdk.Key` 与 `sdk.Timezone`；`sdk.PauseJob`、`sdk.ResumeJob`、`sdk.JobStatus`；示例与已合并的调用方同步（D-124、D-126） | T-07 步骤 10 | T-04.12、T-07.8 | V-114 的 SDK 分项 | 已完成（PR #176；TEST-MATRIX §3.2） |
| T-08.1 | goal-28 | 配置字段描述与构建检查（在 T-04 生成器上扩展），含表单选项的图标路径 | T-08 步骤 1 | T-04.2 | V-25 的描述分项 | 已完成（PR #54；TEST-MATRIX §3.2） |
| T-08.2 | goal-29 | 配置 JSON 语义与 Pending：原始对象保留、完整替换、Merge Patch、编辑快照导入、base_generation | T-08 步骤 2 | T-02 | V-23、V-38 | 已完成（PR #23；TEST-MATRIX §3.2） |
| T-08.3 | goal-30 | SDK 配置准备顺序：OnLoad→补缺键默认→typed decode/Schema | T-08 步骤 3 | T-08.1、T-08.2 | V-37、V-21 的配置分项 | 已完成（PR #73；TEST-MATRIX §3.2） |
| T-08.4 | goal-31 | 不可变 Secret Store：引用、解引用、重新抽取、脱敏读取、受保护清理 | T-08 步骤 4 | T-02.3、T-08.1、T-08.2 | V-25 | 已完成（PR #74；TEST-MATRIX §3.2） |
| T-08.5 | goal-32 | `sqlitedb.Open` 与 `gormdb.Open` 的连接生命周期、迁移互斥与 versioned 迁移、external 插件的维护许可 | T-08 步骤 5、6 | T-08.3 | V-26、V-39 | 已完成（PR #79；TEST-MATRIX §3.2） |
| T-08.6 | goal-33 | SQLite 候选复制与资源交接，写出给 T-09 的调用约定 | T-08 步骤 7 | T-08.4、T-08.5 | V-24、V-36 的候选分项 | 已完成（PR #85；TEST-MATRIX §3.2） |
| T-08.8 | goal-53 | 用户可见文案目录：插件 `i18n/` 的 `messages`、SDK 的 `T`、Core 的 en-US/zh-CN 文案目录（usage 等）、check 的键与占位符检查 | T-08 步骤 8 | T-04.1、T-04.3、T-04.6、T-08.1、T-08.10 | V-54 的 SDK 与 Core 分项 | 已完成（[PR #102](https://github.com/AlonaBot/AlonaBot/pull/102)） |
| T-08.9 | goal-35 | configdb 示例与分层验证，汇总证据 | T-08 步骤 9 | T-08.6、T-08.7、T-08.8、T-08.11 | V-23～V-26、V-38、V-39、V-54 的 T-08 分项 | 已完成（PR #117；TEST-MATRIX §3.2） |
| T-08.10 | goal-90 | 配置标签统一：表单与 secret 元数据收进 `config` 标签（语法同 GORM 标签），label 与 description 改为源文案，补齐 labels、subtitles、多选、slider、code 编辑器、obvious 与 when，取消选项图标；残留的旧标签给出诊断（D-74） | T-08 步骤 1 | T-08.1 | V-25 的描述分项 | 已完成（PR #77） |
| T-08.11 | goal-95 | 配置改由 `config_schema.json` 声明：schema 的检查与嵌入、`$version` 与 `$migrations` 的声明式迁移、i18n/ 中 `config` 覆盖与 schema 的核对、SDK 与 Core 共用的补缺失值与 schema 校验、按 schema 抽取 secret、`sdk.Base.Config()` 按类型取值、根插件的 `OnConfigLoad`；删除 `config` 标签解析、Config 类型发现与 `config.Config`/`config.Base`，保留 go/types 类型推导供 Service Schema 使用（D-76、D-85） | T-08 步骤 1、3 | T-04.7、T-08.3、T-08.4、T-08.10 | V-25 的描述分项、V-37 与 V-38 的配置分项 | 已完成（PR #103；TEST-MATRIX §3.2） |
| T-08.12 | goal-101 | Adapter 实例配置与 Core 设置的版本迁移：`adapter_config_schema.json` 的构建检查、补缺失值与校验，实例配置与 `core_config.json` 的声明式迁移、写回前备份、降级与失败报错，官方 Telegram Adapter 补上 `adapter_config_schema.json`（D-85） | T-08 步骤 1 | T-07.6、T-08.11 | V-38 的迁移分项、V-88 | 已完成（PR #114；TEST-MATRIX §3.2） |
| T-08.13 | goal-123 | 文案函数只返回字符串：`c.T` 与 `sdk.T` 去掉 error，check 与 build 检查代码中的常量键与参数名，运行时问题回落并记 WARN（D-106） | T-08 步骤 10 | T-04.9、T-08.8 | V-104 | 已完成（PR #137；TEST-MATRIX §3.2） |
| T-08.14 | goal-130 | 配置项的选项来源：schema 的 `options_from` 与检查，Service 描述的选项来源标记，`sdk.ServeOptions`，Core 按组件与字段路径查询来源，内置来源 `core.adapter_instances`，`OPTIONS_PAGE`（D-129） | T-08 步骤 11 | T-04.12、T-06.1、T-08.11 | V-116 的 schema、SDK 与 Core 分项 | 已完成（PR #157；TEST-MATRIX §3.2；`@agent` 分项随 T-12.8） |
| T-08.15 | goal-151 | 配置字段的回复示例：`example` 与 i18n 覆盖的检查、`examples/` 图片随包校验、`GET /api/v1/packages/{id}/examples/{path}`（D-160） | T-08 步骤 12 | T-08.11、T-08.14、T-09.12 | V-131 的检查与 API 分项 | 已完成（PR #211） |
| T-08.16 | goal-152 | 去掉配置版本与迁移，配置宽松载入：删除 `$version`、`$migrations`、manifest 的 `config_version`、三个版本字段与迁移实现；`OnConfigLoad(ctx, raw)`；插件、Adapter 实例与 Core 设置的校正与校正记录；REGISTER 追加校正记录；读取校正记录与清理无效配置的管理 API（D-161） | T-08 步骤 13 | T-08.11、T-08.12、T-09.14 | V-37、V-38、V-88、V-132 | 已完成（PR #239） |
| T-08.17 | goal-157 | 插件的 KV 与数据目录：`sdk/kv` 与保留表 `_alonabot_kv`，Core 建立 `<data_dir>/plugin_data/<id>/` 并经 Bootstrap 的 `plugin_data_dir` 下发，`sdk.PluginDataDir(ctx)` 按 Session 返回路径，卸载清理归入数据（D-165） | T-08 步骤 14 | T-08.5、T-09.4 | V-137 的 SDK 与 Core 分项 | 已完成（PR #212） |
| T-08.18 | goal-160 | 配置文件的存储与管理：状态库表、会话的配置文件解析、schema 的 `profile` 检查、覆盖值校验与无效清理、ADMIN-API §23、`alonabot profile`（D-168） | T-08 步骤 15 | T-08.14、T-09.14 | V-140 | 已完成（PR #214） |
| T-09.1 | goal-36 | 首次管理认证：admin-token init/reset、login、会话、CSRF、CLI Bearer | T-09 步骤 1 | T-02.3、T-07.4 | V-47 的认证分项 | 已完成（PR #72；TEST-MATRIX §3.2） |
| T-09.2 | goal-37 | 安装包预检与安全解压：验签、来源、平台、长度/摘要、受控解压与上限、不可变包目录 | T-09 步骤 2（预检） | T-02 | V-27、V-28、V-41 的预检分项 | 已完成（PR #19；TEST-MATRIX §3.2） |
| T-09.3 | goal-38 | 开发 CLI：顶层的 `init/build/check/test/pack`（D-77）；init 在当前目录创建只含作者信息的组件工程，build 生成嵌入代码与调用 `sdk.Run` 的正式入口并编译，取代 `plugin generate`；pack 支持四种角色，WebUI 包为 `any` 平台，签名可选 | T-09 步骤 2（CLI） | T-04.5、T-08.8、T-08.11、T-09.16、T-13.3 | V-01、V-41、V-54 的打包分项 | 已完成（PR #194；TEST-MATRIX §3.2） |
| T-09.4 | goal-39 | 唯一管理服务与安装记录：逐包串行、来源/pin/卸载保留、各管理动作的区分；卸载的三种清理范围（D-127）；组件角色的同类唯一、`install --source` 与 `--replace`；Adapter 包与实例的管理（实例写回 `core_config.json`）；启用、停用、卸载与更新时的 Adapter 依赖和插件依赖检查；`alonabot run` 按安装记录启动全部已启用的包；Core 装配读取 `core_config.json`，接入命令前缀、Core Admin 列表与默认 locale | T-09 步骤 3 | T-03.8、T-07.6、T-08.4、T-08.8、T-09.1、T-09.2、T-13.1 | V-40、V-48、V-52、V-82、V-01 的安装分项、V-88 的接入分项 | 已完成（PR #163；TEST-MATRIX §3.2） |
| T-09.5 | goal-40 | 正式候选 Journal 与准备编排 | T-09 步骤 4 | T-09.4、T-08.6、T-06.4 | V-29 | 已完成（PR #173；TEST-MATRIX §3.2） |
| T-09.6 | goal-41 | 唯一提交点（ActiveGeneration+COMMITTED）与激活门禁 | T-09 步骤 5 | T-07.5、T-09.5 | V-30、V-21 的提交分项 | 已完成（PR #179；TEST-MATRIX §3.2） |
| T-09.7 | goal-42 | 崩溃恢复、显式回退与引用回收 | T-09 步骤 6 | T-09.6 | V-30、V-31、V-36 | 已完成（PR #187；TEST-MATRIX §3.2） |
| T-09.10 | goal-66 | 日志管理：按来源级别的持久化与运行时修改（`sdk.log.set_level`）、最近记录缓冲、按日期查询日志文件、管理 API 与 SSE、`alonabot logs` 与 `log-level` | T-09 步骤 7（日志） | T-02.1、T-03.1、T-07.6、T-09.4、T-09.12 | V-62、V-63、V-68 | 已完成（PR #177；TEST-MATRIX §3.2） |
| T-09.11 | goal-45 | 正式流程的故障注入，回填 T-08 待联验项，汇总证据 | T-09 步骤 8 | T-09.3、T-09.7、T-09.9、T-09.10、T-09.12、T-09.15、T-09.23 | V-23～V-25、V-27～V-31、V-39、V-40、V-52、V-54～V-56、V-62、V-63、V-82、V-83、V-127、V-128、V-129、V-130、V-131、V-132、V-133、V-134、V-135、V-136、V-137、V-138 | 未开始 |
| T-09.12 | goal-84 | WebUI 组件与组件图标：Core 提供 WebUI 静态文件与单页路由回退、未安装与版本不兼容的提示、Admin API 主版本与 `spec/ADMIN-API.md` 参考（含插件页面桥的消息格式）、带 sandbox CSP 的图标、库模式的 `alonabot.WebUI` 选项 | T-09 步骤 9 | T-07.7、T-09.1、T-09.4、T-13.1 | V-83 | 已完成（PR #167；TEST-MATRIX §3.2） |
| T-09.13 | goal-92 | 管理页所需的会话与上传接口：已登录会话取回 csrf_token、multipart 上传安装包与更新包、总览的 Core 版本与监听地址，追加到 `spec/ADMIN-API.md` | T-09 步骤 7（管理页） | T-09.1、T-09.4、T-09.12 | V-40、V-47 的 API 分项 | 已完成（PR #181；TEST-MATRIX §3.2） |
| T-09.14 | goal-93 | 配置、权限、工具策略与会话的管理 API：配置 raw JSON/PATCH/base_generation/secret、命令权限、命令名覆盖（D-128）、配置字段的选项查询（D-129）、发信许可、Core 设置（Core Admin 列表、默认 locale、命令前缀、全局禁止表）、工具调用策略、Agent 会话、插件页面票据、secret 明文查看的再认证 | T-09 步骤 7（配置与插件页） | T-03.25、T-04.14、T-07.6、T-08.8、T-08.11、T-08.12、T-08.14、T-09.6、T-09.12、T-11.1、T-11.3 | V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56、V-115、V-116、V-117 的管理 API 分项 | 已完成（PR #192；TEST-MATRIX §3.2） |
| T-09.16 | goal-100 | manifest 拆分：工程 `plugin.yaml` 只含作者信息（name、description、author 等），构建与打包补上协议、平台、入口、迁移方式与配置版本，包内 manifest 照常严格校验；`i18n/` 中 `metadata` 的翻译（D-78） | T-09 步骤 2 | T-03.8、T-09.2 | V-01 | 已完成（PR #140；TEST-MATRIX §3.2） |
| T-09.17 | goal-142 | 包校验收紧：manifest 缺少迁移方式或准备副作用时只警告，警告出现在安装与更新的结果中（D-150）；安装与更新时严格核对 build-info（D-151） | T-09 步骤 10 | T-09.3、T-09.4 | V-122 | 已完成（PR #201；TEST-MATRIX §3.2） |
| T-09.19 | goal-147 | 诊断包：管理服务的打包与脱敏、`POST /api/v1/diagnostics`、CLI `alonabot diagnose`（连接被拒绝后持有已有 `core.lock`，只读收集停机数据目录）（D-156） | T-09 步骤 7（诊断包） | T-09.10、T-09.14 | V-127 的 Core、API 与 CLI 分项 | 已完成（PR #213） |
| T-09.20 | goal-153 | WebUI 组件的可选进程与 `command.invoked`：manifest、build/pack 与安装对带入口 WebUI 的处理，WebUI 进程的 Session 与管理会话路由，标准事件 `command.invoked` 的 proto、Core 生成与 SDK 常量，插件源 `index.json` 的格式校验工具（D-162） | T-09 步骤 7（WebUI 进程） | T-09.12、T-09.14、T-04.16 | V-133 | 已完成（PR #221；TEST-MATRIX §3.2） |
| T-09.21 | goal-154 | 插件数据浏览的管理 API：表与列、分页查询、按 rowid 增删改、KV 与数据目录文件的管理、zip 导出、清空、保留代次只读浏览与审计日志（D-163、D-165） | T-09 步骤 7（数据浏览） | T-09.14、T-08.5、T-08.17 | V-134、V-137 的 API 分项 | 已完成（PR #234；TEST-MATRIX §3.2） |
| T-09.22 | goal-155 | 补齐的管理 API：保留代次、静态任务的列表与暂停恢复、Adapter 实例列表、聊天目录、更换管理令牌，插件页面的 `.Page` 声明、i18n `pages` 与组件列表的 `pages`（D-164） | T-09 步骤 7（补充 API） | T-09.14、T-07.9、T-03.25、T-04.16 | V-135 | 已完成（PR #220） |
| T-09.24 | goal-158 | 组件的 changelog：`check`/`build`/`pack` 对 `CHANGELOG.md` 与 `changelog/<版本>/<locale>.md` 的校验与打包、安装预检复核，`GET /api/v1/packages/{id}/changelog` 的切分与语言选择，Core 自身 `CHANGELOG.md` 的嵌入与 `GET /api/v1/core/changelog`（D-166） | T-09 步骤 7（changelog） | T-09.3、T-09.14 | V-138 | 已完成（PR #215） |
| T-09.27 | goal-177 | 组件市场 ID 改用连字符：四角色前缀与正则、校验调用者、日志来源、Agent 工具原始名、官方与示例 manifest、CLI/生成资源、既有测试与 fixture；本地 ID 规则保留（D-172） | T-09 步骤 11 | T-09.10、T-09.14、T-12.8、T-13.5 | V-01、V-62、V-64 的改名分项；V-121 的衔接 | 已完成（PR #222；TEST-MATRIX §3.2） |
| T-11.1 | goal-54 | 工具描述与目录：Service 描述的 `tool` 字段、注册校验、`core.tool.list`、调用策略的持久化与管理服务层 | T-11 步骤 1、2 | T-04.4、T-04.7、T-06.1 | V-55 的目录与策略分项 | 已完成（PR #101；TEST-MATRIX §3.2） |
| T-11.2 | goal-55 | 工具调用与用户确认：`core.tool.invoke`、`core.tool.confirm`、Core 内部的确认记录与按钮、同一 Actor 与 5 分钟限时、确认文案 | T-11 步骤 3、4 | T-03.8、T-06.2、T-08.8、T-11.1 | V-55 | 已完成（PR #165；TEST-MATRIX §3.2） |
| T-11.3 | goal-56 | Agent 会话存储：状态库表、`core.conversation.*`、管理服务层的列出与清除、purge | T-11 步骤 5 | T-03.1 | V-56 | 已完成（PR #35；TEST-MATRIX §3.2） |
| T-11.4 | goal-57 | 工具与会话的多进程验收，汇总证据 | T-11 步骤 6 | T-11.2、T-11.3 | V-55、V-56 的 T-11 分项 | 已完成（PR #172；TEST-MATRIX §3.2） |
| T-11.5 | goal-148 | 工具的副作用类别：ToolEffect 与三个 `effect` 字段、注册与目录、`core.tool.confirm` 记录操作的类别、SDK 的 `.Effect`、Agent 按 MCP 注解换算、管理 API 返回类别与批量修改策略（D-157） | T-11 步骤 7 | T-11.2、T-12.3、T-09.14 | V-128 的 Core、SDK、Agent 与 API 分项 | 已完成（PR #219） |
| T-11.6 | goal-163 | 按配置文件的工具策略（D-168） | T-11 步骤 8 | T-04.18、T-11.4 | V-143 | 已完成（PR #246） |
| T-13.1 | goal-67 | 库模式入口：根包 `alonabot.Run`、`alonabot.Start` 与选项、与 `alonabot run` 共用的 Core 装配、开发组件同进程经 loopback 连接、生成代码中嵌入的工程 manifest 身份、可导入 Adapter 包的工程入口与资源登记、dev 来源与固定代次 `dev:<组件 ID>`（D-134） | T-13 步骤 1、2 | T-02.1、T-03.3、T-04.2、T-04.9、T-08.8、T-09.16 | V-64 的生命周期分项 | 已完成（PR #149；TEST-MATRIX §3.2） |
| T-13.2 | goal-68 | 开发组件的配置与数据库：`config.json` 直读、准备与写回（WebUI 修改时重启组件）、`dev.sqlite`、拒绝 update/rollback/uninstall；沿用固定代次 `dev:<组件 ID>`（D-134） | T-13 步骤 3 | T-07.6、T-08.5、T-08.11、T-13.1 | V-64 | 已完成（PR #160；TEST-MATRIX §3.2；配置写回随 T-09.14 的配置管理 API，dev 来源拒绝 update/rollback/uninstall 与开发 Adapter 的实例在 T-09.4） |
| T-13.3 | goal-69 | 开发运行与热部署：`alonabot run [目录...] [--reload]` 构建并以独立进程运行开发组件，轮询源码、资源与配置，只重启变化的组件；Adapter 以可导入包为工程目录，沿用 `dev:<组件 ID>`（D-77、D-134） | T-13 步骤 4 | T-07.6、T-09.16、T-13.1、T-09.4 | V-65 | 已完成（PR #171；TEST-MATRIX §3.2） |
| T-13.4 | goal-70 | 开发运行示例与开发文档，汇总证据 | T-13 步骤 5 | T-03.10、T-13.2、T-13.3 | V-64、V-65 | 已完成（PR #186；TEST-MATRIX §3.2） |
| T-13.5 | goal-140 | 模块根包只保留库模式接口：Core 装配、候选与进程管理移入 `internal/app`，只供开发 CLI 使用的 `Executable`、`Stdio`、`Core.Restart`、`Core.RestartInstance`、`Core.ReloadSettings` 随之移入，根目录的端到端测试按归属迁移（D-148） | T-13 步骤 6 | T-04.16 | V-64、V-65 的回归（行为不变） | 已完成（PR #200；TEST-MATRIX §3.2） |
| T-13.6 | goal-141 | CLI 参数从简：`build`、`check`、`test`、`pack` 的工程目录改为位置参数并删除 `--dir`；`init` 的 ID 与 module、`pack` 的资产 URL 取缺省值；`install` 可按组件 ID 安装 `packages/` 中的官方包，不给参数时在当前目录寻找唯一的安装包（D-149）；`pack --target` 多平台一次打包与签名（D-169） | T-13 步骤 7 | T-13.5、T-14.3 | V-121 | 已完成（PR #233；TEST-MATRIX §3.2） |
| T-13.7 | goal-179 | CLI 开箱即用：`alonabot run` 的终端对话与 `--console`、终端对话时 stderr 的缺省级别、开发目录缺省重载、发布版 `init` 运行 `go mod tidy`（D-181） | T-13 步骤 8 | T-13.6、T-16.1 | V-149 | 已完成（PR #253；TEST-MATRIX §3.2） |
| T-16.1 | goal-81 | 插件测试套件基础：`plugintest.Start`、测试 Adapter、会话与身份、Capability 子集、消息注入、`internal/inproc` 的分发结束通知、配置与数据库 | T-16 步骤 1、2 | T-05.2、T-13.2、T-09.4 | V-73 的基础分项 | 已完成（PR #168；TEST-MATRIX §3.2） |
| T-16.2 | goal-82 | 按钮点击、多轮等待、schedule 触发（含动态任务实例，D-124）与后台发送；用套件测试示例插件，开发文档，汇总证据 | T-16 步骤 3、4 | T-03.8、T-05.4、T-07.2、T-07.9、T-16.1 | V-73 | 已完成（PR #183；TEST-MATRIX §3.2） |
| T-16.3 | goal-102 | `plugintest.LLM`：登记只提供 `llm.generate` 的测试 Agent，被测插件经 `@agent` 的调用交给测试函数（D-77） | T-16 步骤 3 | T-12.8、T-16.1 | V-73 的 LLM 分项 | 已完成（PR #190；TEST-MATRIX §3.2） |

| S-01.1 | goal-168 | 可休眠声明与持久化运行状态 | S-01 步骤 1、2 | T-04.1、T-09.4、T-09.6 | V-77 的 manifest、状态与未启用分项 | 已完成（PR #225；TEST-MATRIX §3.2） |
| S-01.2 | goal-169 | 已提交代次的路由快照与索引 | S-01 步骤 1、3 | T-04.1、T-09.4、T-09.6 | V-77 的快照、摘要校验与恢复索引分项 | 已完成（PR #226；TEST-MATRIX §3.2） |
| S-02.1 | goal-170 | 按需唤醒、活动引用与有界等待 | S-02 步骤 1～3 | S-01.1、S-01.2、T-05.5、T-06.4、T-07.5 | V-77 的首请求、并发、取消、期限与活动引用分项 | 已完成（PR #241） |
| S-02.2 | goal-171 | 空闲退出与即时休眠策略 | S-02 步骤 4 | S-02.1 | V-77 的准入、空闲退出、停止竞争与策略生效分项 | 已完成（PR #249） |
| S-02.3 | goal-176 | Core 消息事件过滤与 SDK 注册 | S-02 步骤 5 | T-04.16、T-08.11、T-09.14 | V-77 的 Core OR 匹配、配置取值与 SDK 注册分项（无需休眠进程） | 已完成（PR #227；TEST-MATRIX §3.2） |
| S-03.1 | goal-172 | Scheduler 正常休眠保留计划、自然到期经既有 Waker 恢复，静态/动态期限与运行槽连续；异常离线跳过 tick | S-03 步骤 1、2 | S-02.2、T-07.2 | V-78；V-77 的 schedule 阻止休眠分项（TEST-MATRIX §3.2） | 已完成（PR #256） |
| S-06.1 | goal-173 | 休眠管理 API、运行诊断与 CLI | S-06 步骤 1 | S-02.2、S-03.1、T-09.14 | V-81 的管理 API、CLI、状态与模式切换分项 | 已完成（PR #269） |
| S-06.2 | goal-174 | 休眠指标、故障注入与固定机器验收 | S-06 步骤 2～4 | S-06.1 | V-77、V-78、V-81 的证据汇总与固定机器测量 | 未开始 |

S 工作包的切片采用 `S-01.1` 形式：工作包编号加包内顺序号，一个切片对应一个可独立合并、可验收的 PR；工作包正文与失败停止条件沿用 [PLAN §7](../PLAN.md#7-休眠扩展)。S-01.1 与 S-01.2 的直接前置相同，可分别取证；共享提交与装配文件按执行顺序串行合并。S-02.3 只依赖已合并 T 切片，供 T-12.18 使用；过滤声明进入 S-01.2 快照，T-04.18 已接入会话生效配置读取（PR #242）；休眠唤醒联验由 S-02.1 的后合并者完成，具体共享规则见执行顺序。S-06.1 提供 WebUI 需要的 API；策略读取复用固定原因顺序，需要路由而无有效已提交快照时返回 412 / FAILED_PRECONDITION，全局设置与 runtime 保持独立可读（ADMIN-API §24）。固定机器测量由 S-06.2 独立完成；页面实现与手动取证在 WebUI 章节的 S-06.3（goal-175），S-06.2 汇总其 V-81 证据或明确标为未运行。WebUI 主线切片（T-09.8、T-09.9）不以任何 S 切片为前置，休眠不进入发布链。

## 3. 工作包

### T-01：契约、工具链与构建基线

#### 前置输入

- 现行 CONTRACTS、WIRE-FORMAT、MESSAGE-MODEL 和 TEST-MATRIX；实际项目位置与发布模块路径由实施时的项目配置确定。
- 可用的 Go 工具链、Protobuf 编译器和生成器；PG 初始化验证需要独立测试数据库。依赖版本须实际验证后写入项目文件。

#### 拟涉及目录

`go.mod`、`go.sum`、`contracts/`、`sdk/` 的公共类型与示例编译检查。它们是待建立的正式代码，数据库小验证放进对应包测试，不建立另一套长期运行时。

#### 按顺序实施

1. 核对项目实际目录、已有 Go/生成工具与模块路径，建立单 module 和可复现生成入口。固定 Go、protoc、Go Protobuf 生成器及所需依赖版本；记录开发环境与目标平台的差别。
2. 将 WIRE 字段表转换成 `.proto`：保留 Envelope 1～16、错误码 1～17、既有 Kind 与 SESSION_ERROR=13；补齐控制帧、Bootstrap、RouteDescriptor、可信 Invocation、HandlerOutcome、SendResult、媒体和 BusEvent 的类型及字段号。字段缺失/零值检查与编码结构分开，结果 payload 和 error 不做 oneof。
3. 为首发消息、配置支持类型、manifest、外部 release 清单建立实际 Schema 与验证入口。静态 fixture 覆盖安全 ID、平台、能力、来源摘要与签名对象；表明构建期能检查的内容及需 REGISTER 才能得到的实例路由。
4. 固定 SDK 的最小公共类型与可编译示例：一个根插件、两类 Handler、Context.StopPropagation、Config.OnLoad、互斥迁移接口、HTTP Handler 和调度声明。此处先证明签名能组合；T-02 起实现运行行为。
5. 用主契约选定的 pure-Go SQLite/GORM 组合完成临时数据库 CRUD/Tx、连接设置、WAL Backup 与恢复读取；验证 Up 与成功账本可在同一插件 DB 事务中提交。对独立 PG schema 验证连接、初始化和事务失败结果。验证通过后锁定实际版本，不预填“最新版可用”。
6. 生成兼容 fixture，测试编码往返、枚举零值、未知必需能力、字段存在性、JSON/Protobuf 编码区分，以及 payload+error 同时返回。让 `go generate ./...` 可重复运行，生成结果稳定；运行构建和相关测试。
7. 对五个目标 OS/arch 验证所选默认驱动与契约包的 `CGO_ENABLED=0` 编译路径，记录每个组合。目标平台的进程启动、运行与清理由 T-02/T-10 完成，不能以交叉编译替代运行证据。

#### 可审查交付

- 已锁版本的模块和生成入口、可编译协议/Schema/SDK 类型、兼容 fixture 与示例。
- 数据库连接/事务/Backup 验证代码和真实结果，供 T-08 直接采用已验证路径。
- V-01/V-02 的静态契约结果，以及 V-26/V-32/V-37/V-39/V-41 在本阶段取得的部分证据；运行和安装语义分别交后续任务闭合。

#### 停止和失败条件

契约能编译、示例签名成立、生成稳定、默认数据库组合的关键事务/备份路径通过，才向 T-02 交付。字段无法表达、驱动无法满足候选复制或目标构建失败时，先定位并修正该具体缺口。若 PG 测试环境尚不可用，可推进已通过的 IPC 工作，但 PG 分支和发布验收保持未完成，不能用 SQLite 结果替代。2026-10-27 起框架不再托管 PostgreSQL（D-48），本任务中的 PG 验证只作为当时的记录保留。

### T-02：Core 状态、Native 进程、双 IPC 与基础生命周期

#### 前置输入

- T-01 已交付可编译的协议类型、控制帧 fixture、manifest/ID 校验规则，以及选定并锁定的 Go、Protobuf 和 Core SQLite 驱动版本。
- 按 [CONTRACTS-v0.1.md](../CONTRACTS-v0.1.md) §1、§4、§6.1、§7.1、§12 和 [WIRE-FORMAT.md](../WIRE-FORMAT.md) 实施。状态、帧号和错误码直接复用协议定义。
- 本项先贯通无业务 DB 的最小插件。T-08 接入实际配置转换、Secret 和插件 DB 准备；T-09 接入包安装、完整更新 Journal 与管理入口。这里的基础状态库和生命周期是它们共同复用的实现。

#### 拟涉及目录

`contracts/`（复用协议及 fixture）、`sdk/`（会话、RPC、生命周期和进程锁）、`internal/core/` 下的状态、进程与 RPC 实现、`cmd/alonabot/`（Core 启停装配）、`examples/plugins/`（最小生命周期示例）。均为单 Go module 内的拟建目录，当前没有这些正式实现；内部子包按职责划分，不另拆 module。

#### 按顺序实施

1. **建立 Core 状态和操作所有权。** 启动时先取得 `<data_dir>/core.lock`，再打开 `state.sqlite`，核对单连接、WAL、FULL、外键和 busy timeout 的实际值。实现最小安装记录、ActiveGeneration/提交标记、启用状态与权限持久化基础，以及逐插件管理操作串行化。先交付可供无业务 DB 插件使用的提交事务原语；T-09 在同一实现上补齐候选资源引用、完整 Journal 与安装恢复编排。提交时核对预期旧代次；事务内只操作元数据，不等待插件 RPC。状态库损坏、锁失败或配置未生效时停止启动，不重建空库继续服务。
2. **建立 Native 启动与退出控制。** 从已校验的安装记录取得入口、插件 ID、平台与归档摘要，生成本次 Session Secret，构造精简环境并启动子进程；不把 Core 管理凭证传入插件。按平台管理普通进程树、取得退出状态并收集带插件 ID 的日志。SDK 在任何用户初始化或受管 DB 打开前取得稳定 `runtime.lock`，持有至退出；锁文件不删除或替换。为后续停机备份提供“确认退出、Core 暂持锁、释放后才能启动 SDK”的顺序。
3. **实现共用帧编解码与两种传输。** TCP 默认直接绑定 loopback 随机端口并持续持有 Listener；stdio 复用相同协议，把 stdout 留给帧、stderr 留给日志。完整读取长度前缀与载荷，在分配前拒绝零长度和超过 4 MiB 的帧。握手核对 Secret、插件 ID、启动摘要、协议和能力；每条连接只有一个 reader 和一个串行 writer，认证失败或不可恢复错帧只隔离当前 Session。
4. **实现双向调用、取消与资源准入。** 按 `(session_id, initiator_side, call_id)` 管理请求，验证父调用仍在途后才继承可信上下文；普通请求采用 30s 默认期限，子请求不延长父期限。落实每方向 256 个 Request、16 MiB payload 和 Core 合计 64 MiB 的限额，满额立即失败。Response、Cancel 和控制帧走有界优先通道；尚未发送的 Request 被取消后撤出队列，已发送请求的 Cancel 保持因果顺序。重复/迟到响应不二次完成，`payload + error` 同时存在时不丢失有效结果，断链不重放业务。
5. **贯通 Bootstrap 和准备阶段。** HELLO_ACK 一次下发 Generation、原始配置、版本、DB 描述、HTTP 认证材料及阶段期限；Core 在配置迁移前只检查 JSON 对象和大小。SDK 依次完成配置准备、许可的 DB 准备、OnStart，再发送完整 REGISTER；可选步骤不存在时跳过。T-02 用 T-01 的最小配置样例、无业务 DB 插件和阶段失败 fixture 验证顺序，实际迁移与 Secret 落地由 T-08 接入。REGISTER 原子校验，超限或失败不能留下半个注册表，敏感 Bootstrap 不进入日志。
6. **实现提交与激活门禁。** Prepared 仅表示准备完成。新候选只有在 Core 状态事务提交 ActiveGeneration/COMMITTED 后才能收到 ACTIVATE；普通 restart 复用既有提交记录、Generation 和配置引用。SDK 执行 OnReady，唯一 writer 完整写出 ACTIVATED 后才打开本地 ready gate；Core reader 同步应用状态后再读后续业务帧。Preparing、Prepared、Activating 均拒绝业务；提交后启动失败进入 Degraded，不自动选旧数据。以无业务 DB 的候选完成本项集成，受管 DB 副本的真实提交/恢复由 T-08、T-09 验证。
7. **完成阶段超时、停止和有限恢复。** 按主契约落实握手 10s、配置 60s、迁移 15 分钟、OnStart 60s、OnReady 30s 的独立预算；用户函数不响应 context 时仍须由 watchdog 结束失败进程。停用先冻结新根调用并最多排空 6s，允许既有父链内合法子调用完成；随后 Cancel 并发送 STOPPING，其期限为 `min(t0+10s, 发送时刻+4s)`。OnStop 使用独立 context，Core 在总 10s 截止处理进程树；Core Shutdown 并行停止所有插件。EOF 使旧 Session 立即失效，SDK 走完 OnStop 再退出，写 stdout 或 stderr 不因 SIGPIPE 终止进程；STOPPING 紧跟 ACTIVATED 到达时 SDK 先发布 ready，按正常停止退出。Core 确认旧进程退出并读完它的日志管道后才结束回收，确认旧进程退出后才允许重启；5 分钟内第三次异常退出自动禁用，正常退出不自动重启。
8. **建立真实进程的验收切片。** 最小示例覆盖正常启动/退出，测试辅助进程覆盖挂起、错帧、日志污染、迟到响应和突退。对同一 fixture 分别运行 TCP 与 stdio，检查进程和 Listener 已关闭；注入满额度、双 Core、并发管理操作和 Core 被终止场景。记录实际运行平台、命令、测试位置及结果，并把由 T-08、T-09 集成的配置/DB/更新场景明确标出。
9. **建立统一日志。** 按 CONTRACTS §14 把 Core 与 SDK 现有的直接输出改为 slog（Core 由装配处、SDK 由 Run 设置默认 Logger）：实现 §14.2 的控制台 Handler（来源标签、级别缩写、源码位置、`key=value`、多行缩进、按终端与 `NO_COLOR` 着色）与 JSON 输出，`--log-level`、`--log-format` 启动参数。SDK 把记录以 JSON 行写到 stderr，在调用边界（生命周期 Hook）捕获的 panic 记为带 `stack` 的 ERROR；Handler 的 Logger 附带 trace_id、route_id、call_id 与 Handler panic 的记录由 T-03.1 的业务分发实现，Bot API URL 中凭证的脱敏由 T-03.4 的 Bot API 客户端实现。Core 逐行解析插件 stderr，补上并保护来源字段，非 JSON 行作为 WARN 原样保留。Bootstrap 追加 `LogLevel log_level = 10`，SDK 与 Core 都按来源的生效级别过滤；按来源级别暂存在内存中，持久化与运行时修改由 T-09.10 接入。日志不输出 Secret、Bootstrap 与配置明文。按 CONTRACTS §14.6 写入按本地日期轮转的日志文件：同日达到单文件上限时续写编号文件，超过合计上限或保留天数时从最旧的删除，写入失败时继续运行并报告一次。
10. **建立调优参数。** 新建 `internal/tune`，按 CONTRACTS §16 的表登记全部可调参数，把 `internal/wire` 与 `internal/core` 中对应的现有常量迁入，调用方改为读取生效值。实现 `--args` 的直接参数与文件输入、值的格式、约束与相互约束检查（出错时指明参数或文件行号）、存在覆盖时的启动 WARN 与 `alonabot tune list`。Bootstrap 追加 `SessionLimits session_limits = 11`，字段 `max_inflight`、`max_payload_bytes`、`request_timeout_ms` 分别对应每方向在途数、每 Session payload 预算、普通请求期限；Core 全局预算不下发给 SDK；设置 `PLUGIN_MEMORY_SOFT_LIMIT` 时为子进程传入 `GOMEMLIMIT`。
11. **状态库改用 GORM。** 按 CONTRACTS §6.1 与 D-68，把 `internal/core/state` 现有各表的读写改为 GORM 模型，经 `libtnb/sqlite` 的 `Conn` 复用同一个单连接；schema 仍由唯一的初始迁移定义，不用 AutoMigrate。core.lock、PRAGMA 回读核对、损坏不重建、提交时核对预期旧代次与事务内不等待插件 RPC 的行为都不变，调用方与集成测试同步修改。之后新增的状态库表直接使用 GORM 模型。
12. **移除绑定 SDK API 的提交测试。** 按 D-86 删除 `sdk/` 全部测试、`internal/core/supervisor/` 除 `supervisor_test.go` 外的测试与整个 `testdata/`，产品代码不变。

#### 可审查交付

- 可以启动最小 Native 插件并走完 Bootstrap、Prepared、提交、Activated、停止的 Core/SDK 代码；两种传输使用同一协议实现。
- Core 状态表及事务边界、状态转换与门禁、进程锁和退出控制的实现位置；最小示例与故障辅助进程只使用测试数据。
- 按现有 V 编号记录的测试结果，包含失败阶段、退出结果和资源清理证据；不得把 fixture 验证记作真实插件数据库回退已通过。

#### 对照现有验收

| 现有条目 | T-02 要完成的验证 |
| --- | --- |
| V-03、V-04、V-05、V-06 | 双传输帧、认证和代次隔离、双向编号与取消、日志污染和有限崩溃恢复；V-05 的 Scheduler 迟到响应语义在 T-07 接入后补齐 |
| V-21、V-37 | 基础生命周期的阶段失败/挂起、合法 SESSION_ERROR、Bootstrap/REGISTER 上限、就绪顺序和全局停止预算；真实配置/DB 准备及更新恢复与 T-08、T-09 联验 |
| V-36 | 同 data_dir 双 Core、同插件并发操作、旧进程未退出时不能启动第二个受管写入者 |
| V-42 | Request 计数/字节满额、控制通道仍有进展、发送前取消的因果顺序；嵌套 Service 与 Event Bus 部分在 T-06 完成 |
| V-32、V-34 | 提供可复用的构建、启动/退出和进程树测试；五平台完整记录与 100 次启停在 T-10 汇总 |
| V-62、V-63 | 步骤 9：控制台与 JSON 格式、三类来源标签、非结构化行、panic、伪造来源字段、按级别过滤；持久化与查看由 T-09.10 完成 |
| V-68 | 步骤 9：按日期换文件、同日切分、合计与天数上限、写入失败；按日期查询由 T-09.10 完成 |
| V-69 | 步骤 10：直接参数与文件输入、优先级、未知键/格式/约束/矛盾拒绝启动、覆盖 WARN、SDK 使用下发值、GOMEMLIMIT |

#### 停止和失败条件

本项范围内的协议、身份、门禁、锁、取消和停止用例通过，最小进程切片可以重复运行，且后续配置/数据库/安装模块接入位置清楚后即可结束 T-02。若出现双代次接单、旧写入者未退出即启动替代实例、满额度使控制帧永久阻塞、丢失有效错误结果或损坏状态库后默默初始化，必须先修复，不能继续把该基础作为已完成依赖。当前缺少某目标平台的真实运行环境时保留“未运行”，由 T-10 补齐；构建记录不能代替该平台进程锁和退出测试。

### T-03：标准消息、可安装 Adapter 与官方 Telegram Adapter

**目标：** 在 T-02 的真实 Session 与激活流程上，运行能够接收消息、回复、主动发送和处理媒体/按钮的 Native 插件；Adapter 作为可安装包在独立进程中运行，官方 Telegram 是第一个 Adapter（D-26）；平台发送失败后仍能判断已确认结果。完整命令生成、Rule 与 Hook 编排在后续工作包接入同一实现。

#### 前置输入

- T-01 已交付协议 1.0、Event/Target/Segment/SendResult/媒体载荷的生成类型及序列化 fixture；字段以 WIRE-FORMAT 与 MESSAGE-MODEL 为准。
- T-02 已能启动独立插件进程，完成鉴权、Bootstrap、显式静态描述注册、激活、REQUEST/RESPONSE、取消及停止；Core 状态库与 Session/Generation 身份可用。
- 读取 CONTRACTS §1.3～1.4、§8，WIRE-FORMAT §3.4，MESSAGE-MODEL §1～4、§6，DESIGN §M。使用本地 Bot API 测试服务覆盖可控失败；真实 Telegram 实测使用实施环境已经授权的测试账号和会话。

#### 实施步骤

Adapter 以安装包和独立进程运行（CONTRACTS §8）。以下步骤中，Telegram 的接收、标准化、发送计划、answerCallbackQuery 与平台取件都在官方 Telegram Adapter 程序内实现；游标持久化、事件接纳、Target 授权、按钮 token、媒体缓存与派发留在 Core。

1. **交付首个可用 SDK 切片。** 在最终 `sdk/` 中实现标准 Event、Target、Chain 构造、手动 Handler 的调用适配、Context、Send 和媒体客户端；复用 T-02 Session 与 ready gate。通过显式传入实例/函数及 RouteDescriptor 构建最小 echo 插件，先采用返回 `error` 的手动发送签名。该插件必须经过真实 REGISTER/ACTIVATED 后才收业务；不依赖尚未实现的注解生成器或命令树。接口直接使用正式生成类型，不建立另一套临时协议；T-04 在其上增加声明发现和构建检查。

2. **接通 Adapter 进程与 SDK。** Core 按安装记录以 Adapter 角色启动进程，复用插件的握手、准备、激活、停止与有限重启；REGISTER 只接受 Adapter 描述。把进程内的每个账号实例包装为 Core 的 Adapter 实例接口，经 `adapter.instance.start/stop` 单独启停、经 `adapter.send` 发送，启停一个实例不重启进程。接收 `core.adapter.deliver` 时先校验事件属于本 Session 已启动的实例，再持久化游标、接纳事件并应答；`core.adapter.cursor` 保存无事件时的游标变化，`core.adapter.status` 报告实例致命错误。同一平台账号只允许一个已启动实例。实例进入运行状态后生成 `bot.online`，因停用、Adapter 进程退出或致命错误离开运行状态时生成带原因的 `bot.offline`。提供 `sdk/adapter`，并用一个回环测试 Adapter 验证完整链路。

3. **落实 Target 与最小授权路径。** Core 保存真实入站的完整 Target 和可信来源；原事件调用及合法后续父链只可默认回复该 Target。后台发送必须显式 To，其他 Target 与后台调用须匹配 Core 状态库中该插件的 `allowed_adapter_instances`；D-83 起它缺省为全部实例，管理员可以收紧为指定实例（T-03.12）。发送准入逐次检查实例启用、account 对应关系、当前许可和全局禁止，不信插件自报 Actor/Admin；没有默认目标返回 NO_DEFAULT_TARGET。提供 Core 内部策略读写入口供测试及后续管理面复用，避免为本阶段开放未鉴权的管理接口。基础身份与授权随首个 Adapter 落实，T-04 扩展命令继承，T-05 集成完整事件权限链。

4. **接通 Telegram 接收与生命周期。** 在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。

5. **实现同一 Core 媒体暂存与授权。** 接通 BeginUpload→顺序 Chunk→Commit，验证大小、MIME、SHA256，执行 1 MiB 分块、100 MiB 单对象、4 个未完成上传、每 Session 256 MiB、Core 2 GiB 及 15 分钟期限。入站只签 IncomingMedia 描述 token；首次 Read 才经 `adapter.media.read` 由所属 Adapter 调用平台文件 API 取件，Core 校验并缓存，随后按 offset/length 读取。多 Session 获得独立 token、共用一份缓存；上传者读自身对象及真实入站 token 的读授权不依赖主动发送许可。Grant 只给符合契约范围的 Activated 接收者签新 token；共享不会放宽 Send 的目标权限。配额预留、owner 退出、引用过期和消费租约的释放归同一对象管理，不能先返回可用引用再发现下载失败。出站媒体由 Core 为该次发送向 Adapter 签发只读 token，Adapter 经 `core.media.read` 读取。

6. **实现可预测的发送计划及回执。** 发送计划在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。**消息编辑：** Core 的 `core.edit` 按 Send 的 Target 准入执行出站 Hook（输入标明 edit），再经 `adapter.edit` 交给 Adapter；官方 Telegram Adapter 用 editMessageText 实现，只编辑文本消息，并声明 `message.edit` 能力；平台拒绝时不改为发送新消息。

7. **接通按钮 token、回调路由与 extension 片段。** 使用显式注册的 button.clicked 路由验证 action/schema/value，SDK 以带类型的按钮动作注册路由并负责 value 的编码与解码（D-107）；Core 生成 32 字节随机值的无 padding base64url token，持久化绑定插件、路由、Generation、Target 和 24 小时期限。点击先尽力 answerCallbackQuery，再验证 token、当前 actor/策略及目标，使用普通 REQUEST 调用正确 Session。多用户点击逐次鉴权；普通同代次重启保留 token，卸载/切换代次永久失效。拒绝不能落入 Raw/Regex，也不能发额外群错误消息；缺少可核验普通 Target 的 inline 回调明确拒绝。answerCallbackQuery 由 Adapter 在交付点击事件前调用。Adapter 在注册时声明 extension 及其 Schema，Core 在发送前按目标实例所属 Adapter 校验；解析 manifest 的 `role`（plugin/adapter/agent/webui）、`adapter_id`、`icon`、WebUI 的 `assets` 与 Admin API 主版本范围、`adapter_dependencies` 与 `plugin_dependencies`，按 CONTRACTS §9.1 校验各角色的必填与禁止字段，在开发入口启用插件时检查必需依赖（正式管理入口在 T-09）。

8. **运行发送与接收故障矩阵并记录证据。** 用 T-13.3 的 `alonabot run <目录...>` 同时运行 Telegram Adapter 与示例插件，Adapter 的实例写在 `core_config.json` 的 `adapters` 中（CONTRACTS §15.1、§15.4，D-77、D-82），构建 echo/媒体示例以及本地 Bot API 测试服务；覆盖同包两个实例单独启停、Adapter 进程崩溃后恢复、伪造实例事件被拒、实例上下线事件、消息编辑，验证两个账号/实例、话题、多物理消息、提交后超时、相册后 markup 失败、零 ID、媒体过期/共享/配额、游标提交后崩溃和接收冲突。以调用记录证明失败后没有自动补发，以平台请求的实际数量证明一条 Chain 与两次 Send 的区别。跨插件媒体授权先用两个真实 Session 和显式测试调用验证，不等待 T-06 Service。记录运行命令、实现变更、模拟/真实平台版本和日志；真实 Telegram 尚无可用测试环境时单独标「未运行」，不得用模拟结果冒充平台实测。

9. **接通 Telegram 的平台权限。** 在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。

步骤 10～19 补齐与 AstrBot 对标的平台能力（D-87）。协议与 Core 侧在本章节，两个平台的实现在各自的实施计划中；出站 Hook 中禁止调用的服务名单在 T-05.3 实现，T-05.3 先合并时由后合并的切片把自己的服务加入名单，反之由 T-05.3 一并加入。

10. **媒体按来源传递（D-88）。** MediaSource 追加 `inline_base64`：Core 校验标准 base64 与解码后 1 B～1 MiB，拒绝宿主路径与其他 scheme；追加只由 Core 填写的 `platform_file`，插件给出时 INVALID_ARGUMENT。Send 的目标实例就是 incoming_ref 的来源实例时，Core 把 Adapter 原先提供的平台文件标识作为 `platform_file` 交给 Adapter，不取件。实现 Adapter 调用的 `core.media.url`：为本次发送的只读 token 签发 32 字节随机 url_token，到期不晚于消费租约；Core 在统一端口以 GET、HEAD 提供 `/media/<url_token>`，支持 Range，到期或租约结束后 404，日志只记录 url_token 的前缀。SDK 的字节构造函数在 1 MiB 内生成 inline_base64，超出时先上传为 media_ref。

11. **新增片段与字段（D-89、D-91、D-97）。** 在 proto 中追加 voice 片段、ForwardItem 的 `platform_message_id`、ForwardSegment 的 `platform_forward_id`、Button 的 `url`，按 MESSAGE-MODEL §2 校验：来源恰有一个、chain 与 platform_message_id 恰有一个、嵌套 forward、platform_forward_id 只能出现在入站、url 与 action_id 恰有一个且 url 按钮不带 value。Core 在调用 Adapter 前按 `message.voice`、`message.forward` 判定能力；按钮签发跳过 url 按钮。SDK 追加构造函数与入站类型，语音片段的时长为 `time.Duration`（D-110）。只有一个平台具备的片段（例如全体提及）由该 Adapter 的 extension 提供，不在此追加（D-87）。Event 追加 `chat_type`（新增枚举 ChatType，D-100）：Core 在入站标准化时拒绝没有 target 却不是 UNSPECIFIED 的事件，chat_type 不参与 Target 的比较与准入；SDK 的 Context 提供 `ChatType()` 与 `IsPrivate`、`IsGroup`、`IsChannel`，UNSPECIFIED 时都为 false。

12. **平台动作（D-92）。** Adapter 描述追加动作声明，Core 在注册时校验名称、版本、Schema 与作用范围。实现 `core.action`：按 CONTRACTS §8 的发信准入检查目标，核对动作声明、版本与参数 Schema，经 `adapter.action` 交给实例，再按结果 Schema 校验；结果未知时返回 UNAVAILABLE 或 DEADLINE_EXCEEDED，不重试。manifest 的 `adapter_dependencies` 追加所需动作，启用检查与 extension 相同。SDK 提供底层动作调用方法，以及 Context 与后台客户端共同实现的调用者接口，供 Adapter 包的带类型函数使用；`sdk/adapter` 以可选接口让实例实现动作，动作以 `adapter.Action[P, R]` 声明，参数与结果的 Schema 由 SDK 从类型推导（D-113）。

13. **撤回（D-93）。** 实现 `core.delete` 与 `adapter.delete`，Target 准入与 Edit 相同，目标实例未声明 `message.delete` 时在调用 Adapter 前返回 UNSUPPORTED；按 MESSAGE-MODEL §4A 汇总每个 ID 的结果。撤回不执行出站 Hook。SDK 提供对应方法，只收消息句柄，一次撤回的句柄须属于同一会话与实例（D-104）；`sdk/adapter` 以可选接口让实例实现。输入状态与表情回应只有 Telegram 提供，是它的平台动作（步骤 18）。

14. **命令菜单同步（D-96、D-138）。** 实例声明 `commands.sync` 时，Core 在实例启动成功后与路由表或命令策略变化后，合并 2 秒内的变化，经 `adapter.commands.sync` 交给它可用的命令：只含该实例命令前缀含 `/`、权限只要求 `user` 且未被禁用的根命令。本切片在 CommandSpec 追加可选的英文 description，并实现 `func (b *CommandBuilder) Description(text string) *CommandBuilder`；当前由 CommandRegistry.Command 返回 builder，T-04.12 后沿用同一 builder。英文原文即 en-US，其他语言按 `commands[完整规范路径].description` 覆盖（含 Group 前缀，以一个 ASCII 空格连接 token）；生成器不扫描注册，check/build 只检查顶层 commands 的结构（键为字符串，值为含字符串 description 的对象）；SDK 的 Collect 对没有对应已注册命令或命令未用 Description 登记英文原文的覆盖路径拒绝注册，错误为英文并带出路径。Core 为 en-US 及插件目录中的 locale 选择说明，缺少覆盖时按插件默认 locale、英文原文回落，未登记说明时用命令名；T-14.1 复用该字段与翻译选择。同步失败只记诊断。

15. **流式输出（D-98、D-135）。** SDK 的 `c.Stream(next func(context.Context) (string, error)) ([]*message.SendResult, error)` 按 MESSAGE-MODEL §4A 处理文本增量、结束/错误与回执，选择节流编辑或缓冲后发送一次；同 Session 同完整 Target 的编辑间隔 1 秒，失败即停止，每次发送与编辑照常执行出站 Hook。在 `message.send` Capability 上追加 `text_limit{max, unit}`，单位为 UTF-16 码元、Unicode 码点或 UTF-8 字节，补齐 Adapter 描述、实例启动、事件与 SDK 的传递和边界校验；SDK 按声明换条，未声明不换条。官方 Telegram Adapter 用现有 `maxTextUTF16` 声明 4096 UTF-16 码元。proto 按 T-03.13～T-03.19 的授权只追加字段，字段号按合入时 main 分配；不新增流式服务。Telegram 的草稿是它的平台动作（步骤 18）。

16. **Adapter 的 HTTP 路由（D-94、D-136）。** Adapter 的 REGISTER 可以携带 HTTP 路由与实际监听的 loopback endpoint，复用插件的路径、方法与端点校验，角色限制按 CONTRACTS §9.1；SDK 根包的可选 `AdapterHTTP` 接口为 `HTTP(route func(method, path string) *RouteBuilder)`，`sdk.RunAdapter` 在声明了路由时托管 Listener。平台 token 路由用 `Public(AuthPlatformToken)`，追加对应 `HTTP_AUTH_MODE_PLATFORM_TOKEN` 枚举值；代理从可信安装角色区分路由，Adapter 包的 Authorization 保留并由 Adapter 校验，插件路由继续清除，两者均清除内部代理凭证、CSRF、Cookie、Forwarded 与 X-Forwarded-* 后注入内部凭证。统一端口代理转发 WebSocket Upgrade，升级后的连接不受普通请求期限限制，路由撤销、Session 失效或代次变化时关闭。

17. **Telegram 的媒体与新增片段。** 在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。

18. **Telegram 的平台动作（含输入状态、表情回应与草稿）与撤回。** 在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。

19. **Telegram 的命令菜单。** 在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。

20. **Telegram 的代理模式与话题判定。** 在[Telegram 实施计划](adapters/telegram.md#t-03-的-telegram-步骤)中实施。

21. **插件给出的目标由 Core 补齐（D-102）。** 在 `internal/core/adapter` 的实例表中实现目标补齐：只补空字段；缺少 adapter_instance_id 时，调用链上在途事件的目标属于同一 Adapter 与 chat_id 则取其实例，否则取该 Adapter 的缺省实例，没有时 FAILED_PRECONDITION；实例已登记时补上 adapter_id 与 account_id。`core.send` 的 To 与引用片段的 target、`core.media.grant` 的 targets 在准入前补齐，之后的编辑、撤回与平台动作复用同一函数；补齐后的目标照常经过原 Target 豁免、实例许可、会话禁止与实例绑定检查。`internal/core/settings` 校验实例名的字符集与跨 Adapter 唯一，读取可选的 `default_instance` 并给出各 Adapter 的缺省实例，由装配处在读取与重新读取设置后交给实例表。

22. **聊天目录（D-130，T-03.25）。** Event 追加可选的 `chat_title`，`sdk/adapter` 的交付接口带上它；没有 target 的事件带名称时 Core 拒绝并记诊断。状态库增加聊天表（写进唯一的初始迁移），以 adapter_id 与 chat_id 为键，不含话题。Core 接纳带 target 的入站事件时，在保存游标的同一个事务中写入或更新该聊天的 chat_type、名称、最近交付它的实例与最近活跃时刻，事件不带名称时保留已有名称。向 T-08.14 的来源登记处登记内置来源 `core.chats`：搜索词不区分大小写地匹配名称与 chat_id，支持 `chat_type` 与 `adapter_id` 两个 filter，按最近活跃时刻从新到旧排列，存储值为不含话题的会话字符串。把聊天表加入 T-09.4 的清理对象：`--purge-data` 与 `--purge` 卸载 Adapter 时在同一个状态库事务中删除该 adapter_id 的行，只卸载与 `--replace` 时保留。名称不写入日志。两个官方 Adapter 的填写由 T-03.26 与 T-15.7 提供。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `contracts/` | 复用生成的消息/媒体/回执类型与 fixture；发现载荷不足时先修正同一契约，不新增平行 wire |
| `sdk/`、`sdk/adapter/` | Context、Chain、手动 Handler、Send/media 客户端与 ready gate 的实际用法；Adapter 作者使用的接口与 `RunAdapter` |
| `internal/core/adapter/`、`internal/core/adapterhost/`、`internal/core/router/`、`internal/core/message/`、`internal/core/media/`、`internal/core/policy/` | Adapter 实例接口与进程外实现、事件派发、发送准入、回执汇总、对象/引用/租约、实例许可与按钮映射的状态操作 |
| `adapters/telegram/` | 官方 Telegram Adapter 程序：接收、标准化、平台取件、发送计划及 Bot API 故障测试 |
| `examples/adapters/loopback/` | 测试用回环 Adapter，用于不依赖平台的 Adapter 链路集成测试 |
| `examples/plugins/echo/`、`examples/plugins/media/` | 能通过 T-02 实际启动的显式注册插件；后续迭代保持既有行为 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-07 | 原 Target 回复、显式目标、无许可/撤销、实例账号与话题隔离、NO_DEFAULT_TARGET |
| V-08、V-35 | 完整预检与物理顺序；四类逻辑结果、部分结果、零 ID、payload+error、无自动重发 |
| V-09、V-44 | 上传校验、惰性 Read、授权 Grant、多 Session 共用缓存、配额/过期/租约与失败清理 |
| V-45 | owner/webhook/401/409、订阅范围、匿名发送者、逐条游标和 7 天轮次边界 |
| V-48 | Adapter 进程与实例启停、崩溃恢复、事件归属、extension 校验与依赖解析；正式安装/卸载与实例管理由 T-09 闭合 |
| V-51 | 启动、停用、进程崩溃与 401/409 时的上下线事件及原因 |
| V-52 | `plugin_dependencies` 解析与开发入口的启用检查；正式管理动作由 T-09 闭合 |
| V-53 | 编辑成功、平台拒绝、不支持与不合规的 Chain、每次编辑的出站 Hook |
| V-54 | Telegram language_code 到 Actor.locale 的映射；文案选择由 T-08 实现 |
| V-89～V-95 | 步骤 10～16 的协议与 Core 侧分项；两个官方 Adapter 的分项由 T-03.20～T-03.22 与 T-15.4～T-15.6 提供 |
| V-98 | 步骤 11 中 chat_type 的 Core 与 SDK 分项；两个官方 Adapter 的填写由 T-03.20 与 T-15.5 提供 |
| V-99、V-100 | Telegram 的代理模式与话题判定（步骤 20） |
| V-101 | 步骤 21 的 Core 分项；SDK 的会话字符串与 `sdk.Via` 由 T-04.8 提供 |
| V-102 | 步骤 13 中撤回只收消息句柄的 SDK 分项；句柄本身与编辑、引用由 T-04.8 提供，Telegram 中指向消息的平台动作由 T-03.21 提供 |
| V-105 | 步骤 7 中带类型的按钮动作：注册、按钮的编码与 Handler 收到的解码值 |
| V-108 | 步骤 11 中语音片段的 Duration 与协议毫秒值的往返；其余字段由 T-04.8 提供 |
| V-111 | 步骤 12 中动作的 Go 类型声明：参数 Schema 不允许未声明的属性，结果 Schema 允许；事件与 extension 分项由 T-04.8 提供 |
| V-117 | 步骤 22 中聊天的记录、名称的保留、`core.chats` 的搜索与筛选、随 Adapter 的数据清理；两个官方 Adapter 的填写由 T-03.26 与 T-15.7 提供，管理 API 与页面分项由 T-09.14 与 T-09.9 提供 |

T-04 的切片从首个 SDK 切片开始即可推进（见 §3.1），不要求本包的 Telegram 部分先完成；T-04.5 的示例升级在本包示例完成后进行。真实平台验证尚未运行的部分继续保持未完成，由 T-10 汇总补齐。完整普通事件排序、自动回复、Hook 和命令策略组合分别在 T-04/T-05 闭合。

#### 失败停止条件

- Target/账号凭证串用、默认实例许可意外放开、IncomingMedia 读取暴露 Bot Token/路径或 Grant 扩权，阻止本包合入。
- Adapter 能为非本 Session 已启动的实例交付事件、平台凭证进入插件进程或日志、启停一个实例需要重启整个 Adapter 进程，阻止 Adapter 链路合入。
- 平台已成功却被报告为完全未发送、部分结果丢 ID、发送前已知不支持却先产生消息、失败后自动重发，必须先修复发送管道。
- 单个 Handler 阻塞 Session 读写、媒体缓存/队列超过固定预算、接收游标可能重放已接纳业务或重复启动 owner，停止主链路集成并回查 T-02/T-03 的所有权与准入实现。
- 碰到平台无法表达的既定能力，记录 UNSUPPORTED 和具体受影响验收；不能用静默降级替代通过，也不能在本包扩展 D-87 之外的平台功能。
- 任何来源被统一转成 base64 发送、平台动作可以绕过发信准入或透传未声明的动作、`platform_file` 能由插件给出，阻止合入。
- 为只有一个官方 Adapter 支持的能力新增通用片段、服务或 SDK 方法，先回到契约讨论（D-87）。

### T-04：根插件、构建期注册与命令策略

**目标：** 将 T-03 的可运行 SDK 变成可以稳定构建、注册和管理命令的插件开发入口，保留显式注册方式；策略变更无需重启插件。

#### 前置输入

- T-03 的正式 SDK、手动 Handler 示例、发送通路及原事件目标授权可用；T-02 的注册验证、状态库和激活流程可复用。
- T-01 的 manifest、稳定 route ID、RouteDescriptor、Schema 与生成 fixture 可用。读取 DESIGN §D/§E，以及 CONTRACTS §2、§9 的当前注册和命令规则。
- 本包实现配置类型的唯一性/构建描述检查；完整 OnLoad、Secret、DB 候选流程由 T-08 接入，不以 T-08 已实现为前置条件。

#### 实施步骤

1. **完成根插件 API 与实例绑定。** 在现有 SDK 上实现恰好一个根类、嵌入 sdk.Base、至多一个 Config 的发现/验证；插件 ID 只来自 plugin.yaml。保留显式传入根实例或辅助接收者的注册接口，不建立自动 DI。生成器与显式适配器校验两种 Handler 签名及参数类型；执行时的完整 Rule/传播/自动回复组合由 T-05 完成，不另造第二套 Handler 返回协议。D-114 以 `func New() *sdk.Plugin` 取代根类与 `sdk.Base`，由 T-04.12 修改。

2. **实现限定范围的构建期生成。** 只扫描当前插件在所在 module 中选定的目标构建包及其本模块依赖包，遵守 build tags、GOOS/GOARCH，排除依赖、vendor、测试文件。包级 command/event 标记生成 Go 注册代码，由编译器检查函数类型；实例方法必须通过已提供的实例显式绑定。仓库内各示例以各自构建入口限定扫描范围，不能全仓库扫描所有示例后误报多个根插件。产出稳定描述、schema 引用和构建诊断；同输入重复生成结果一致。通过 `cmd/alonabot/` 的构建检查入口调用同一生成逻辑，避免 CLI 和 SDK 各自维护一份扫描器。D-79 删除构建期标记，命令与事件只经链式注册，标记的扫描与生成由 T-04.9 删除。

3. **建立静态注册索引与诊断。** 合并生成及显式描述，通过一次 REGISTER 提交；同插件重复 route ID、规范化命令路径或别名拒绝注册。Core 为已经实现的命令/事件描述建立索引，保留后续 Hook/Service 共用的描述及校验入口，不提前开放其未实现的执行能力。跨插件同路径记录唯一胜出者和遮蔽者，按 priority/plugin_id/route_id 稳定确定，与启动次序无关。Session 内注册固定，运行时策略单独读取。

4. **实现命令选择和参数绑定。** 只对 message.created 的 `/` 命令匹配完整 token；Telegram @username 按已验证的当前 bot 处理。先最长普通路径/别名，再文本匹配（Regex、keyword、startswith、endswith、fullmatch，统一按 priority/plugin_id/route_id 排序），最后 Raw。将「已识别普通命令」保存为分发状态，禁止禁用、权限拒绝、参数错误后转入更短路径或其他类别。按契约实现引号、反斜杠、int64、有限 float64、true/false、必需/可选/默认与末位 raw 尾部；Regex 启动编译，捕获结果独立于参数；空模式与非法 Regex 拒绝注册，ignore_case 按 Unicode 简单大小写折叠比较。对允许执行但参数非法的命令生成一次 usage 或采用声明的 silent；权限拒绝不自动发群错误。

5. **实现持久化命令策略。** 以 plugin_id/route_id 为键记录权限和禁用覆盖，复用 T-02 状态库以及 T-03 的实例许可入口。权限取最近显式子项，否则继承；插件、祖先组及自身任一禁用都拒绝，最后与 Core 全局禁止约束合并。每次派发前读取一致的当前策略；已开始调用不宣称被撤销。提供管理服务层的查询/修改/遮蔽诊断方法，持久化成功后才报告变更成功，供 T-09 鉴权后的管理 API/WebUI 使用。

6. **用同一个 SDK 升级示例。** 将 echo/媒体示例增加命令组、子命令、别名、Regex 与 keyword 等文本匹配、Raw、可选参数和显式实例方法，同时保留一条纯显式注册路径。验证生成产物实际编译、启动、注册、接收命令并调用 T-03 Send；不以只比较生成文本代替编译。对 T-05 才执行的自定义 Rule false/error，先用 HandlerOutcome fixture 验证命令消费状态机，后续补真实 SDK Rule 集成证据。

7. **完成构建和策略验收。** 使用独立测试插件或构建 fixture 验证多 root/Config、错误方法签名、依赖中的伪标记、不同 build tags、重复路由和参数溢出；以两个实际 Session 验证交叉同名命令与禁用保留路径。直接调用 Core 管理服务层修改策略，确认下一次派发生效、进程未重启、Core 重启后保留；为后续 WebUI 提供同一方法的输入、输出和失败条件。将静态解析证据与待 T-05/T-09 集成的用例分项登记。

8. **改造权限、扩展事件与命令解析的协议和 Core 侧。** 按 WIRE-FORMAT §5A 的 2026-11-01 条目修改 proto 并重新生成。Core 校验 Adapter 的权限与扩展事件声明（前缀、名称、隐含关系无环、Schema 可编译）；交付事件时只保留已声明的平台权限并展开隐含关系，按声明校验扩展事件；路由按权限名集合判定（CONTRACTS §2.3）。按 CONTRACTS §2.2 实现 `command_prefixes` 与按实例覆盖、去掉开头的 reference 与指向本账号的 mention、非文本 token、`user`/`image`/`message` 参数与 usage 文案键。同步修改已合并的 router、remote Adapter、`sdk/adapter`、Telegram Adapter 与集成测试（D-69、D-71、D-72）。

9. **SDK 改为链式注册并使用自有类型。** 按 D-70 把命令、文本匹配、事件、Service 与 Event Bus 订阅的注册改为链式构建，删除 `RouteOption`、`ServiceOption` 与位置参数形式；新增 Permission 值对象（`Or`）与 `c.Has`、标准事件常量与 `sdk.AdapterEvent`，`Collect` 核对 Handler 的参数类型；Service 的 Schema 由生成器从 Go 类型生成；Context、Send、事件 payload 与错误码改用 SDK 自有类型，发送目标用会话字符串，机器人用可选的 `sdk.Via`（D-102），编辑与引用用消息句柄（D-104），每种片段都有构造函数，Send、编辑与自动回复收 Chain 或单个片段（D-108），插件以 `sdk.OpenMedia` 顺序读取媒体，`sdk.Upload` 自行计算大小与 SHA-256（D-109），时间与时长用 `time.Time` 与 `time.Duration`（D-110），`sdk/adapter` 的扩展事件与 extension 声明给出 Go 类型、Schema 由 SDK 推导（D-113），自动回复型 Handler 可以 `c.Send`。构建期生成器、示例、`sdk/adapter` 的公共类型与已合并调用方同步修改。链式注册（含 Permission 与 Service Schema）为 T-04.7，自有类型、事件常量与 `sdk/adapter` 的公共类型为 T-04.8（2026-11-04 拆分）。

10. **管理员与封禁按 Adapter 登记（D-105）。** `core_config.json` 的 `admins` 与状态库的禁止表改用用户字符串 `<adapter_id>:<external_user_id>`，会话禁止按 `<adapter_id>:<chat_id>`；Router 按事件的 adapter_id 判定 admin，CheckCommand 与 CheckSend 按 adapter_id 判定禁止，同一 Adapter 的全部实例共用。同步已合并的 policy、router 与 settings；管理入口（T-09.4、T-09.14）与 /whoami（T-14.1）按新格式实现。

11. **注册写法统一（D-112）。** 定时任务与 HTTP 路由改为与 D-70 相同的链式构建：`r.Every(jobID, d)`、`r.Cron(jobID, expr)`（可选 `.Timezone(name)`）与 `r.Route(method, path)`（可选 `.Public(mode)`）开始，`Handle` 结束，删除 `HTTPOption` 与位置参数形式。命令省略 `Args` 时参数名取 Handler 的参数名：Go 运行时取不到参数名，由生成器为插件包中第一个参数是 `*sdk.Context` 的函数、方法与函数字面量记录参数名并嵌入 Generated，SDK 按 Handler 查找，查不到时仍为 arg1、arg2；写了 `Args` 时以它为准。Event Bus 订阅改为 `sdk.Subscribe[T](r, eventName)`，路由 ID 缺省为事件名，可用 `.ID(routeID)` 另给。协议不变，示例与已合并的调用方同步修改。Hook 的同样改写属于 T-05.3。本步骤排在步骤 12 之后，`r` 指步骤 12 的插件对象或子插件。

12. **插件对象、结构体参数与子插件（D-114～D-116）。** 协议不变，只改 SDK、生成器与调用方。
    - 插件对象：新增 `sdk.NewPlugin()` 与 `*sdk.Plugin`，把各注册器的方法移到它上面。删除 `sdk.Base` 与 `sdk.Root`；删除根插件的六个注册方法 `Commands(*CommandRegistry)`、`Events(*EventRegistry)`、`Services(*ServiceRegistry)`、`Subscriptions(*SubscriptionRegistry)`、`HTTP(*HTTPRegistry)`、`Schedules(*ScheduleRegistry)`，SDK 不再调用它们；删除这六个注册器类型。都不留别名。`sdk.JSON`、`sdk.Proto`、`sdk.Subscribe` 的第一个参数改为 `sdk.Registrar`。生命周期与配置回调改为 `p.OnStart`、`p.OnReady`、`p.OnStop`、`p.OnConfigLoad` 登记，配置经 `p.Config()` 与 `c.Config()` 读取。`sqlitedb` 与 `gormdb` 的迁移接口改为登记函数 `Migrate(p, fn)` 与 `Migrations(p, list)`，生成器按对这四个函数的调用写入 `migration_mode`，SDK 在准备阶段核对。`sdk.Run`、`sdk.Collect` 与库模式的 `alonabot.Plugin` 改收 `*sdk.Plugin`；生成器从查找根类型改为核对入口包的 `func New() *sdk.Plugin`，正式入口调用 `sdk.Run(New())`。重复登记由 `Collect` 报错，`sdk.Run` 开始后的注册 panic。
    - 生成资源（D-137）：生成器在入口包的 `alonabot_gen.go` 中生成 `var alonabotGenerated = sdk.Generated{…}`，作者在 `New()` 中显式调用 `sdk.UseGenerated(p, alonabotGenerated)`；`func UseGenerated(p *Plugin, g Generated)` 把资源保存到插件对象，`func (p *Plugin) Generated() Generated` 供 SDK 与库模式读取只读资源。保留 schema、i18n、Service Schema 与存在时的 Manifest 等字段，重复 New 的对象各自绑定，不按根类型登记、不包装 New。删除插件资源的 init 与类型表；扫描视图补入资源变量的类型声明，保证没有旧生成文件时仍可检查 New 并生成。生成器 Result 用 `PluginConstructor string`（导入路径.New）取代 RootType，同步 CLI 判定。T-13.1（goal-67）正在按旧类型登记实现；T-04.12 合并后同步库模式资源读取，若 T-13.1 先合并，本切片一并迁移其插件选项、SDK 同进程入口与示例调用。
    - 结构体参数：`Collect` 按 CONTRACTS §9.1 检查 Handler 的结构体参数（字段类型与导出、`arg` 与 `default` 标签、名字的缺省规则、指针与 default 的可选语义、与 `Args`/`Optional`/`RawTail` 互斥），生成与位置参数相同的 CommandParam；分发时把 Core 解析好的值赋给字段。两种形式同时接受 `int` 与底层为标量的具名类型。
    - 子插件：`p.Sub(id)` 与 `*sdk.SubPlugin`，ID 校验与重复检查；子插件中的路由照原规则生成 route_id 并与整个插件一起查重；`s.Permission` 进入缺省权限的解析次序；`s.Config()` 返回同名顶层 object 的视图，`Collect` 核对 schema 中有该字段；子插件 Handler 的 `c.Logger()` 带 `sub_plugin`。
    - examples/plugins、`alonabot init` 的模板（尚未实现时只改语句）与已合并的调用方按新写法修改；docs/plugin-guide.md 与 ARCHITECTURE 中对根插件的描述由审查者在合并前同步。

13. **SDK 的第二轮简化（D-118～D-123）。** 协议不变。
    - 属于 T-04.8 的部分：`c.Reply(text)` 与后台用的 `sdk.SendText(ctx, to, text)`；`message.User`、`c.User()` 与 `c.Session()`；`sdk.OpenMedia` 改收 `message.MediaSource`；`errcode` 的 `New`、`Errorf` 与 `Of`。
    - 属于 T-05.4 的部分：命令、文本匹配、Raw、事件与按钮路由 builder 的 `.Continue()`（T-05 步骤 7）。
    - 本步骤（T-04.13）：Service 与 Event Bus 改为带类型的声明。`sdk.JSON[In, Out](target)`、`sdk.Proto[In, Out](target)` 返回 `sdk.Service[In, Out]`，`sdk.Serve(r, s)` 开始注册，`s.Call(ctx, in)` 调用；`sdk.BusEvent[T](name)` 返回事件声明，`e.Publish(ctx, payload)` 发布，`sdk.Subscribe(r, e)` 订阅。`Collect` 核对 target 的组件部分是本插件（Agent 组件为 `@agent`）、声明不是零值、同一声明只注册一次。删除 `sdk.CallJSON`、`sdk.CallProto`、`sdk.Publish` 与旧形式的 `sdk.JSON(r, name)`、`sdk.Proto(r, name)`、`sdk.Subscribe[T](r, name)`；生成器对类型参数的检查改为针对声明处。`sdk/llm` 的 `Generate` 与 `Ask` 签名不变，内部改用声明并导出它；示例与已合并的调用方同步修改。
14. **命令名覆盖（D-128，T-04.14）。** 协议与 SDK 不变。命令策略的路由作用域增加命令名列表，写进状态库的命令策略表；保存时按注册描述中命令路径的规则校验每个命令名，拒绝空列表、列表内重复、文本匹配与 Raw 路由，并按 CONTRACTS §2.3 检查同一插件内的重名。Router 建立命令索引时，有覆盖的命令用覆盖的命令名代替声明的路径与别名；与本插件其他命令声明的名字相同的覆盖项不进入索引并记入诊断，一项不剩时用声明的名字；覆盖变化后重建索引，下一次派发生效。遮蔽诊断、usage 回复与命令菜单同步（T-03.17）改用生效的命令名，主名为列表第一项；覆盖变化触发受影响实例的菜单同步。命令组与策略的作用域仍按声明的规范路径。管理 API 由 T-09.14 提供，页面由 T-09.9 提供；`core.command.list`（T-14.1）返回生效的命令名，两者后合并者同步。

15. **SDK 公开面收敛（D-144、D-145、D-147）。** 协议不变，只改 SDK、生成器、CLI 与调用方。
    - 会话与句柄：删除 `message.Session`、`message.MessageRef`、`message.User` 及其构造与 `Parts`，删除 `sdk.To`；公开 API 的会话、消息句柄与用户字符串改为 string。新增 `c.SendTo(sid, content, opts...)`；`c.Send` 只发往当前事件的默认目标；`sdk.Send`、`sdk.SendText`、`c.CallAction` 与平台动作改收 sid。`sdk/message` 提供 `SID`/`ParseSID`、`Ref`/`ParseRef` 与 `UID`/`ParseUID`。
    - 生成资源：作者工程中不再有生成文件。`internal/generate` 生成调用 `genimpl.Register` 的登记代码，`run`、`build`、`test`、`pack` 经 overlay 编入入口包，不写作者目录；`build` 在没有 `plugin.yaml`、import 根包的 main 包目录中按库模式构建，扫描其依赖中的组件入口包（含模块缓存中的第三方 module），把全部登记编入该 main 包。plugintest 在 `Start` 时按被测组件的导入路径经 `go list` 定位源码，按测试二进制构建信息中的平台与 build tags 现场扫描并登记。`sdk.NewPlugin()` 只记录调用方所在的包，`sdk.Run`、`sdk.RunAdapter`、库模式与 plugintest 启动组件时才取得资源并解析命令参数名，缺少登记时在连接 Core 前返回英文错误。删除 `sdk.Generated`、`sdk.UseGenerated`、`sdk.UseGeneratedAdapter`、`sdk.GeneratedFor`、`sdk.GeneratedProvider` 与 `(*Plugin).Generated`；SDK、根包与 plugintest 经本 module 的内部包读取资源，仓库测试的资源夹具改经内部入口注入。`sdk/generate` 移到 `internal/generate`。删除官方组件、示例与 `plugintest/internal` 中已提交的 `alonabot_gen.go` 与 `//go:generate`，CI 的生成一致性检查随之只覆盖 Protobuf；`init` 模板不含生成文件，`check` 只检查资源与入口。新增的内部包与依赖边写进 ARCHITECTURE。
    - Adapter 路由：`adapter.Descriptor` 增加 `HTTP []HTTPRoute`，删除 `sdk.AdapterHTTP`；官方 OneBot 11 与 devbot 的回环 Adapter 改用新字段；生成器按 Adapter 包对 `adapter.HTTPRoute` 的引用写入 http_listener。
    - 示例：showcase 与 configdb 删除 `Commands`、`HTTP`、`Schedules` 这类按类别分组的方法，在 `New()` 中直接注册（D-114）。
    - docs/plugin-guide.md、docs/development.md、examples/devbot 的说明与 `alonabot init` 的模板按新写法同步。

16. **SDK 按类别分包（D-146）。** 协议与行为不变，只改包与名字。按 D-146 的对照把包级函数与类型移入 `sdk/task`、`sdk/service`、`sdk/bus`、`sdk/button`、`sdk/media`、`sdk/conversation`、`sdk/tool`、`sdk/action` 与 `sdk/options`；子包 import 根包取得 `Registrar`、`*Context` 与 `Permission`，共享的注册状态放在 `sdk/internal`，根包不 import 子包。`RunInProcess`、`RunAdapterInProcess`、`Collect`、`Registration` 及其描述类型改为内部入口。全部调用方（Adapter 的常量与动作包、Agent、plugintest、示例、根包与文档）同步修改，不留别名；ARCHITECTURE 的模块表与依赖边同步。
17. **插件页面套件（D-159）。** 协议与 Core 不变。`sdk/page.Handler() http.Handler` 提供嵌入的 `kit.css` 与 `kit.js`；插件挂到自己的 GET 路由，页面以相对路径引用。`kit.css` 以 ADMIN-API §7 的令牌为按钮、输入框、选择框、表格、标签与卡片提供基础样式及明暗缺省值。`kit.js` 在 iframe 内发送 `hello`，应用令牌、处理主题和语言事件，并将 ⌘K/Ctrl+K 转发为 `shortcut`。公开接口为 `request`、`t`、`onTheme`、`onLocale`；`t(key, args?) → Promise<string>` 每次经桥传递键与参数并原样返回 `text`，页面在语言回调中重新 `await t(...)` 渲染。独立打开时使用浅色缺省值，`t` 返回键名，`request` 拒绝。套件随 SDK 发布，无第三方依赖；接口细节见 ADMIN-API §7.3，挂载示例见 showcase。

18. **配置文件的派发与 SDK（D-168，T-04.18，已完成，PR #242）。** InvocationContext 追加 `profile_id = 7`，Bootstrap 追加必填对象 `profile_overrides_json = 14`，新增 `sdk.profile.set_overrides` 与 `core.profile.list`、`current`、`assign`、`set`（WIRE §3.2、§3.5、§4）。派发时按 CONTRACTS §5.6 解析会话所用的配置文件；关闭插件时命令按禁用处理，文本匹配、事件路由与 Raw 跳过，PreEvent 在媒体准备前检查开关。准入与消息过滤共用一次操作视图，以 Active 基础配置应用有效覆盖。覆盖值变化时向在线 Session 下发完整集合，SDK 原子替换并为每次调用冻结 `c.Config()`；`p.Config()` 与 `s.Config()` 保持基础视图。`core.profile.set` 只接受调用插件自己写了 `profile` 的字段，会话使用 `default` 时互斥核对 Active/Pending，写入 Pending，并在根调用、派生调用、自动回复与延续全部结束后 reload；assign 与 set 要求 Core Admin，遇管理互斥占用立即返回 FAILED_PRECONDITION。已与 T-08.16 的无版本配置载入及校正共同验证；休眠唤醒的装配与退出联验由 S-02.1 的后合并者完成。

19. **平台与消息（D-168，T-04.19，已完成，PR #250）。** 命令前缀按会话所用的配置文件读取；Router 的 `Options.Prefixes` 为 `func(context.Context, *pb.Target) ([]string, error)`，派发传完整 Target，菜单传带实例与账号、空会话的 Target，读取实例缺省配置文件。`profile.Service.EffectivePlatform(id string) (prefixes []string, reply settings.Reply, err error)` 按键合并当前 default 与覆盖。删除 `command_prefixes_by_instance` 的解析、保存与 `GET/PUT /api/v1/core/command-prefixes`，旧磁盘键按 Unknown 保留并忽略、prune 清理，docs/development.md 与 V-87 已同步；`core_config.json` 提供 `reply`。发送管道经 `message.Service.SetReply(func(context.Context, *pb.Target) (settings.Reply, error))` 注入配置读取，在出站 Hook 之前按 CONTRACTS §5.6 为可信消息根的首条回复加上提及、引用与普通文本前缀；手动发送、自动回复、Service 派生与延续共享首条标记。按选定物理 ID 的已确认片段记录修饰，Edit 保留提及与前缀，无法表达的 Markdown 组合返回 UNSUPPORTED；加上后超过单条上限时省略提及与前缀。`Release(router.Caller)` 在 Session 退出时释放关联并使在途准备失效。每次回复 Send 在目标实例声明 `message.forward` 时把超过阈值的消息改为一个 forward 片段。后台、定时任务与发往其他会话的消息不变。
20. **SDK 用法从简（D-180，T-04.20）。** 协议不变。按 D-180 修改 `sdk` 根包的注册签名并给 builder 增加 `.ID`，缺省 route ID 在收集注册时生成并查重。Plugin、SubPlugin 与 CommandGroup 均提供无 ID 的文本匹配、Raw 与 On；Hook 在 Plugin 与 SubPlugin 上登记，`.ID` 在 `Handle` 前设置。`*sdk.Context` 实现 `context.Context`；`c.Reply(text string, more ...message.Segment) error` 一次发送文本与后续片段；`.Page(title)` 自动补齐同目录的 GET 套件资源，按 D-183 以公开静态资源模式提供，作者已登记的同方法、同路径路由优先；`plugintest.Result` 增加 `RequireReply(t testing.TB, want string)`。同步仓库内全部调用方：官方组件（含 WebUI）、示例、probe、`init` 模板、测试与现行文档示例。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `sdk/`、`sdk/generate/` | 根插件/显式绑定接口、限定模块的构建扫描及编译可检验的生成代码 |
| `contracts/` | 静态描述和参数 fixture，复用现有 route/schema 标识 |
| `internal/core/router/`、`internal/core/policy/` | 路由索引、命令消费状态、参数解析、策略持久化和遮蔽诊断 |
| `cmd/alonabot/` | 复用生成器的构建检查入口；完整打包、来源验证和管理鉴权在其所属任务实现 |
| `examples/plugins/` | 同 SDK 下的显式注册、注解函数、实例方法和命令组示例 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-10 | 单 root/单 Config、函数/方法类型、限定扫描、构建标签及错误注册阻止激活 |
| V-11 | 路径/别名/遮蔽/参数/Regex/Raw 的实际调用；Rule 消费先有结果 fixture，真实 Rule 组合由 T-05 闭合 |
| V-49 | 文本匹配的注册校验、排序、ignore_case 与「已识别命令不进入文本匹配」；与 Rule 和传播的组合由 T-05 闭合 |
| V-12 | 策略服务层的即时生效、持久化和全局禁止优先；实际 WebUI 操作由 T-09 完成 |
| V-43 | 父禁用、子权限覆盖、权限名集合的判定（`admin`、`chat_admin` 与 Adapter 权限互不替代）及下一次派发生效；完整多事件流程由 T-05 再验证 |
| V-85～V-87 | Adapter 权限与扩展事件的声明校验、平台权限的保留与展开、命令前缀与非文本参数；Telegram 的平台权限由 T-03.11、OneBot 11 的由 T-15.1 提供 |
| V-102 | 消息句柄的格式与往返，SendResult 与 Event 给出句柄，编辑与引用只收句柄；撤回分项由 T-03.16 提供 |
| V-103 | 同一 Adapter 的多个实例共用 Core Admin 与禁止名单；用户字符串的解析、校验与写回 |
| V-106 | 片段构造函数与 `message.Content`：单个片段、Chain 与 nil 经 Send、编辑与自动回复的结果 |
| V-107 | 插件的媒体读取与上传：OpenMedia 顺序读取，Upload 自行计算大小与 SHA-256 |
| V-108 | SDK 类型中的时间与时长：与协议毫秒值的往返，零值对应 absent |
| V-110 | 步骤 11 中定时任务与 HTTP 路由的链式注册、命令参数名的缺省与订阅路由 ID 的缺省；Hook 分项由 T-05.3 提供 |
| V-111 | 步骤 9 中扩展事件与 extension 的 Go 类型声明与推导的 Schema；动作分项由 T-03.15 提供 |
| V-112 | 步骤 12 中插件对象的注册与回调登记、结构体参数的绑定与注册检查、子插件的路由标识、缺省权限与配置视图 |
| V-113 | 步骤 13 中 Service 与 Event Bus 的带类型声明；SDK 类型分项由 T-04.8 提供，延续调用分项由 T-05.4 提供 |
| V-115 | 步骤 14 中命令名覆盖的保存检查、匹配、遮蔽、usage 与菜单同步；管理 API 分项由 T-09.14 提供，页面分项由 T-09.9 提供 |
| V-120 | 步骤 15 与 16：会话、消息句柄与用户字符串的 string 形式与 `c.SendTo`；作者工程中没有生成文件与生成相关的写法，`go test` 经 plugintest 现场扫描取得资源，`alonabot build` 构建的库模式二进制取得依赖中各组件的资源，直接 `go build` 的库模式产物在启动组件前报错；Adapter 路由在描述中声明；分包后的公开面 |
| V-141 | 步骤 18 中配置文件的解析、覆盖值的下发与 `c.Config()` 合并、按配置文件的插件开关与 `core.profile.*`；存储与管理 API 分项由 T-08.18 提供 |
| V-142 | 步骤 19 中按配置文件的命令前缀与四项回复方式 |
| V-148 | 步骤 20 的注册从简、Context、Reply、页面套件与 RequireReply |

V-12 是跨任务验收：T-04 不依赖尚不存在的 WebUI，但必须交付后续 UI 实际调用的策略服务层。矩阵中保存分项证据；只有 T-09 的鉴权后浏览器操作也完成，才将整项 V-12 标为通过。T-04 完成后，T-05 和 T-08 可以分别使用同一注册器继续实施。

#### 失败停止条件

- 生成器扫描外部依赖或隐式构造接收者、同一源码因启动顺序而选择不同路由、重复 route 被静默覆盖，阻止生成器/注册器合入。
- 已识别但拒绝的命令落入低优先级同名项、短命令或 Regex/Raw，先修复分发状态，不用示例避开冲突。
- 子项启用绕过父禁用、平台权限或 chat_admin 满足 `admin`、Adapter 给出未声明的权限名仍被采信、策略重启后丢失或需要重启插件才生效，阻止策略功能完成。
- 配置文件关闭的插件仍收到该会话的派发、回复方式作用于后台或发往其他会话的消息、非 Admin 经 `core.profile.set` 或 `assign` 修改配置文件，阻止 T-04.18 或 T-04.19 合入。
- 为了测试 UI 而增加临时无鉴权管理端点，或为未来配置迁移建立与 T-08 相竞争的运行时，停止该实现并复用既定服务入口。

### T-05：普通事件分发、Rule、Hook、可信权限链与多轮等待

**目标：** 组合 T-03 发送管道与 T-04 路由，形成可解释的入站处理顺序、两类 Handler、传播停止和 Hook 结果，并提供延续调用与多轮等待；为 T-06 保留完整受信上下文。

#### 前置输入

- T-04 注册器、命令消费状态、策略服务层以及 T-03 Adapter/Send/媒体已可运行。
- T-02 能根据在途父请求保存并验证 Session、actor、target、trace、deadline，子请求取消规则可用。
- 读取 CONTRACTS §2、MESSAGE-MODEL §3～5、WIRE-FORMAT 的 HandlerOutcome/Hook 载荷。T-06 Service 与 T-07 Scheduler 均不是本包前置；其专属集成场景明确交回后续任务。

#### 实施步骤

1. **收口可信事件上下文。** 将 Adapter 原始输入转为 Core 保留的身份、Target、来源消息 ID 及路由绑定，业务 payload 与这些保留信息分开。确认只有可信个人 Actor 才带 `user`，Core 实例作用域管理员才带 `admin`；平台权限只来自 Adapter 的声明，只在该会话有效（CONTRACTS §2.3）。没有 Actor 的事件和后台调用不补用户。按钮 action/schema/value 与 scheduled job/run 等已验证绑定不随可修改 payload 改变；scheduled 用规范 fixture 验证即可，实际 tick 由 T-07 接入。

2. **实现有界的同事件分发。** 接收入口统一执行 Normalize→PreEvent→命令/类型匹配与 Core 权限→SDK Rule/Handler。普通事件每条链顺序执行，跨事件最多 64 条并发；等待队列不超过 256 条且 8 MiB，并计入已有 Core payload 总预算。候选按 priority/plugin_id/route_id 稳定排序；入站超额明确拒绝/丢弃并计数，接收游标依 T-03 规则处理。链执行从 Session 读循环分离，慢 Handler 不能停止 Response/Cancel/控制帧处理。

3. **完成 SDK Rule 与 HandlerOutcome。** 自定义 Rule 与目标 Handler 在同一 RPC 中执行；默认 AND，显式 OR 通过 SDK 组合，false/error 均不运行目标函数。SDK 提供按 message.created 的 `to_me` 判断的 `ToMe` Rule。适配 `error` 与 `(Content,error)` 两个签名（D-108），返回明确 matched/propagation；只有自动模式成功返回非空内容时触发一次 Send，排在 Handler 内的显式 Send 之后；自动型也可以显式 Send（D-70）。手动模式允许多次 Send。panic 转 INTERNAL 并保留诊断，不吞成成功；Context 返回后失效。

4. **固定传播与错误分支。** c.StopPropagation() 在 SDK 锁存，结果即使同时携带 ErrorInfo 也保留 Stop；普通事件 Rule false 跳过、业务错误默认继续，显式 Stop 终止后续候选。已识别命令则沿用 T-04 的消费状态，不因 Rule false/error 回落。测试 Handler 已发送后失败的情况，确保不会补发、撤销已确认消息或自动发送错误回复；后台 goroutine 不得修改已返回结果或继续借用失效父调用。

5. **接入三个 Hook 阶段。** 使用独立 route kind 和稳定排序运行 PreEvent、BeforeSend、AfterSend。PreEvent 只接受允许的 payload/传播变化，拒绝覆盖可信身份、Target、来源与按钮/任务绑定；失败结束本次入站链。每次显式或自动 Send 独立执行一组 BeforeSend→Adapter.Send→AfterSend；修改后的整条 Chain 重新预检，拒绝不调用平台。AfterSend 只观察最终回执/错误，其失败记录诊断而不改写已确认结果。SDK 的 Hook 与其他路由一样链式注册并使用 SDK 类型：`r.PreEvent(事件常量)`、`r.BeforeSend()`、`r.AfterSend()` 开始，需要稳定身份时链式调用 `.ID(routeID)`，`Handle` 结束；编辑时 `c.Editing()` 给出被编辑消息的句柄（D-112、D-180）。

6. **落实出站递归约束与父链继承。** Core 在 BeforeSend/AfterSend 的受信执行上下文设置 send_forbidden；通过直接 Send 验证拒绝。该标记与 actor、target、scope、deadline 同存于 T-02 父调用记录，并提供给后续 Broker 使用，不能依赖插件自报。当前任务不为验证间接递归而临时实现 Service；在接口 fixture 中证明派生上下文保留此标记，将真实 Hook→Service→Send 拒绝作为 T-06 的必交回归用例。

7. **实现延续调用与多轮等待。** `core.continuation.start` 只接受由真实入站事件触发、具有 Actor 与默认 Target、不在出站 Hook 或 send_forbidden 调用链中的在途 Handler 调用；原调用成功结束后，Core 以同一 route_id 和原可信上下文发出 Handler REQUEST，其 `continuation_of` 为原调用的 call_id；原调用返回错误时不发出。该请求使用独立的 15 分钟期限，不占用分发链与等待队列，传播仍只由原调用决定。延续调用内再次请求或无 Actor 的请求返回 FAILED_PRECONDITION。`core.wait` 在 Core 内存中登记（adapter_instance_id、chat_id、thread_id、external_user_id）键，同键已有等待返回 ALREADY_EXISTS；Core 在 PreEvent 之后、命令匹配之前检查等待，命中的 message.created 作为结果返回，不再进入普通分发。超时、取消与 Session 失效结束等待，不持久化、不补发。SDK 提供 `Continue` 与 `Wait`。

8. **运行实际多进程流水线。** 复用示例插件配置多个 priority、同优先级逆安装顺序、自定义 Rule、两种 Handler 与三个 Hook，以及两个用户在同一群同时等待、等待超时和等待期间插件停用；逐项检查平台调用日志、候选调用次序及 HandlerOutcome。用恶意载荷尝试修改 Actor/Target/button/job 字段，用延迟 Handler 验证控制帧仍推进。将 V-11 的真实 Rule 消费、V-43 的权限组合回归与本包验收一起记录；仅使用 scheduled fixture 的地方标明尚未覆盖真实 Scheduler 执行。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `sdk/` | Rule 组合、两个 Handler 适配、StopPropagation 锁存、Context 生命周期、Hook 适配、`Continue` 与 `Wait` |
| `internal/core/router/`、`internal/core/policy/` | 同事件候选链、命令消费衔接、可信身份和逐次策略检查、延续调用与等待表 |
| `internal/core/message/` | 三阶段 Hook 与现有 Send 管道集成、只读 AfterSend 结果 |
| `internal/core/` 的现有 Session/调用上下文实现 | 父链权限和 send_forbidden 的保存/继承接口；不建立平行 Session 管理器 |
| `examples/plugins/` 及所属包测试 | 可运行多插件分发场景与绑定/身份篡改 fixture |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-13 | 不同启动/安装次序下候选调用顺序稳定 |
| V-14 | Rule false/error、两类 Handler、Stop+error、自动型可 Send、自动回复恰一次且排在显式 Send 之后 |
| V-15 | 三阶段 Hook、不可改绑定、直接递归 Send 拒绝、AfterSend 不改成功；真实 Service 间接路径由 T-06 闭合 |
| V-16 | 无 Actor fixture、伪造管理员、实际支持的按钮/编辑事件与不支持能力；真实 scheduled 执行由 T-07 补验 |
| V-11、V-43、V-49 | 将 T-04 的命令消费、文本匹配及权限策略接入真实 SDK Rule/事件链，完成相应组合回归 |
| V-50 | 延续调用的上下文与期限、不占分发链、多轮等待的命中与消费、同键冲突、超时、Session 失效与嵌套请求 |
| V-110 | 步骤 5 中 Hook 的链式注册、Handler 签名与 `c.Editing()`；其余分项由 T-04.11 提供 |

T-05 完成条件包含直接 Hook 和上下文继承能力，不要求 T-06 尚未存在的 Broker 通过 E2E。整项 V-15 必须等 T-06 的间接递归用例通过后才关闭；T-06/T-07 可以在本包完成后并行推进。

#### 失败停止条件

- Stop 随错误丢失、Rule false 仍调用 Handler、自动回复重复发送、普通命令拒绝后回落，阻止事件流水线完成。
- 插件 payload 能改变可信身份、Target、按钮/任务绑定或沿用结束后的用户权限，先修复 Core 记录与 RPC 关联。
- BeforeSend 后未经复检就发平台请求、AfterSend 改写成功事实、直接 Hook 可再次 Send，阻止发送管道合入。
- 队列越界、一个 Handler 阻断 Session 控制或把超时当作已撤销副作用，回查准入/取消实现；不以扩大无限缓冲掩盖失败。

### T-06：Service Broker 与非持久 Event Bus

**目标：** 使用正式 Session 与受信父链完成跨插件服务调用，广播拥有独立的有界队列；闭合 T-05 的间接 Hook 递归检查。

#### 前置输入

- T-05 的上下文/Hook/权限约束、T-04 注册器与 schema 描述、T-03 Send/媒体和 T-02 双传输 RPC 均可复用。
- 读取 CONTRACTS §1.2、§3、WIRE-FORMAT §4～5，动态命令/事件 Router 不在 v0.1 范围内（Q-15）。
- 准备 A/B/C 三个真实插件进程的测试入口和可控阻塞点，不用进程内函数互调冒充 Service E2E。T-07 的调度、HTTP 与 T-09 的管理 UI 不作为前置。

#### 实施步骤

1. **补全 Service/Bus 描述与 SDK 入口。** 通过现有 REGISTER 提交 service_name、输入/输出 schema/version、编码、调用规则及订阅事件声明。支持既定 UTF-8 JSON 和 Protobuf 编码，固定随包 schema 资源；Core 校验标识、版本、大小和 JSON 语法，SDK 用生成类型/校验器检查业务字段。拒绝运行时网络 `$ref`、未声明编码、未知版本及编解码失败，空结果仍返回声明的空对象/消息。Envelope 校验对 REQUEST/RESPONSE 的业务载荷只要求 `schema_version` ≥ 1，版本不符由 Broker 返回 INVALID_ARGUMENT，不断开调用方 Session（D-66）。

2. **实现 Broker 定位与准入。** 以 plugin_id/service_name 定位当前 Activated Session，检查默认 Core/本插件规则以及显式 allowlist/public；public 不赋予用户 Admin。目标不存在、未激活、未授权或输入不合规时在执行目标函数前返回明确错误。把每一段 Plugin→Core 与 Core→Plugin 请求单独编号，保存实际关联，不依靠跨 Session 相同 CallID。

3. **沿父调用传播真实上下文。** parent_call_id 只接纳同 Session 仍在途的 Core→Plugin 请求；从 Core 记录继承 actor、target、scope、trace、祖先插件链和 send_forbidden，不接受插件覆盖。无父链的后台服务只有插件身份。每次追加插件检查同链重复和深度 16，同链重复返回 CALL_CYCLE，深度超过 16 返回 FAILED_PRECONDITION；独立并发链各自记录。服务中向原事件 Target Send 仍经 T-03 准入；媒体跨 Session 仍须 Core Grant，不能把原 token 当公共对象地址。

4. **接通跨段取消和背压。** 子调用 deadline 取父与自身操作预算较早者；父完成/超时/取消或任一 Session 断开，结束对应结果并取消未完成子调用，不重放。复用 T-02 的 256 在途、16 MiB/方向、Core 64 MiB 预算，额满立即失败。Broker、SDK 与各 Handler 不在等普通业务槽位时互相占锁；Response/Cancel/控制保持进展，尚未写出的 Request 被取消即撤出，Cancel 不超越其 Request。普通迟到响应不二次完成调用。

5. **实现独立的非持久广播。** Publish 使用普通带 CallID 的请求，仅确认 Core 完成发布准入；所有权限及命名空间检查在分发前完成。按订阅注册的事件名/schema/version发送无业务回执 EVENT；每个订阅者最多 256 条且 4 MiB，满时 drop-newest 并计数，离线不缓存、不补发。广播回调使用订阅插件身份，不沿用发布者的用户权限；一个慢订阅者不得阻塞其他订阅者或 Publish。广播不参与普通事件的 StopPropagation，也不伪装成 scheduled 请求。SDK 对每个 Session 的订阅投递额外限制为 4 MiB（`BUS_QUEUE_BYTES` 缺省值），包含排队及正在处理的 EVENT protobuf 载荷；超出时 drop-newest、WARN 并立即 release wire 接收预算。Core 的 `BUS_QUEUE_BYTES` 覆盖仅调整 Core 等待队列，SDK 此上限保持 4 MiB，以给同 Session 的业务请求留出预算。

6. **接入 Session 变更和可诊断结果。** 按 T-02 的冻结/drain/STOPPING 流程停止新根调用，处理在途父链和取消；Session 失效时清理路由目标、订阅队列、Broker 映射与预算，不能转交同插件新 Session 自动执行。错误只返回安全 ErrorInfo 与 trace，不暴露远端堆栈/Secret/路径。记录按原因区分的调用拒绝、取消、离线与队列丢弃计数，供后续管理面读取；不建设可靠事件存储或服务自动重试层。

7. **完成嵌套与故障端到端验收。** 分别在 TCP 和 stdio 运行 A→B→C、A→B→A、深度边界及两条独立并发链；检查结果、Actor/Target、deadline 和取消在每一跳一致。运行 A 的出站 Hook→B Service→Send，证明 Core 保留的 send_forbidden 使间接发送失败，从而闭合 V-15。注入 B/C 挂起、退出、满额、畸形输入/输出与排队前取消，检查没有死锁或重复业务效果。并行制造慢订阅者、断线、条数满和字节满，证明新事件按规则丢弃且其他订阅者继续。给每个 V 条目记录实际进程/传输/命令与日志。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `sdk/` | Service 客户端/服务适配、Bus Publish/订阅适配及生成类型校验 |
| `internal/core/service/` | 路由授权、跨 Session 调用映射、父链/循环检查和取消传播 |
| `internal/core/eventbus/` | 命名空间准入、独立订阅队列、drop-newest 与清理计数 |
| `contracts/` | 服务/广播编码 fixture，与 T-01 生成模型一致 |
| `examples/plugins/services/` 及所属包测试 | A/B/C 独立可执行示例、嵌套/循环/Hook 间接 Send 与故障场景 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-17 | 广播离线不积压、256 条/4 MiB 双限额、drop-newest 计数、慢订阅者隔离 |
| V-18 | 真实 A→B→C、循环拒绝、独立并发链互不误判，双传输结果一致 |
| V-19 | 编码/schema/version/授权/大小错误在目标函数前拒绝，正确输入输出可往返 |
| V-42 | 将 T-02 传输压力场景接入真实 Send/Service 嵌套；满额、取消、优先通道和 Event Bus 字节预算均有证据 |
| V-15 | 闭合 T-05 留下的 Hook→Service→Send 间接递归禁令；整个 V-15 在相关分项通过后才关闭 |

T-06 完成后，T-09 可以依赖稳定 Service/Bus 接口集成管理流程；仍不开放跨插件动态命令 Router。T-10 汇总本包已有证据，只有为解决具体剩余风险或目标平台门槛才扩大测试，不重复搭建第二套消息通道。

#### 失败停止条件

- 任何跨 Session Actor/权限串用、过期父链继续授权、Hook 经 Service 绕过发送限制，阻止 Broker 合入。
- 同链循环挂起、独立链误报循环、排队耗尽时 Response/Cancel 无进展或 Cancel 先于 Request，先修复关联和调度，不能用自动重试掩盖。
- 编码/版本/授权错误仍运行目标函数、错误丢有效业务结果或把远端异常变成空成功，阻止服务接口完成。
- Event Bus 超过固定条数/字节预算、慢订阅者拖住其他订阅者、重连补发已丢事件或把广播当任务完成确认，先修正队列和生命周期语义。

### T-07：Scheduler 与统一 HTTP 入口

#### 前置输入

- T-05 已提供事件/路由、可信调用上下文和权限检查；T-02 的 Session 状态、RPC、ready gate、进程退出通知和停止流程可复用。T-01 的 schedule/http 注册描述已能编译。
- 按 [CONTRACTS-v0.1.md](../CONTRACTS-v0.1.md) §4.3、§7.2～§7.4 和 [WIRE-FORMAT.md](../WIRE-FORMAT.md) 的 REQUEST、注册及 HTTP Listener 契约实施。
- T-09 负责首次管理 token、离线 reset、真实登录/CLI 和 WebUI。本项先完成授权检查与插件票据接口，用测试会话验证组件；实际入口在没有有效管理会话时必须拒绝，不能为了提前演示而开放临时免认证路径。T-09 再完成真实登录后的端到端授权验收。

#### 拟涉及目录

`contracts/schedule`（触发计算）、`sdk/`（scheduled callback 包装、HTTP 注册和 Listener）、`internal/core/` 下的 Scheduler、HTTP 路由/代理与插件会话实现、`cmd/alonabot/`（Core 公共 Listener 和 Base URL 装配）、`examples/plugins/`（定时与 HTTP 示例）。继续使用单 Go module；`web/admin/` 的管理页面由 T-09 交付，本项不新建另一套管理前端。

#### 按顺序实施

1. **接入并校验注册描述。** schedule 注册固定 job_id、interval/cron、timezone 和 job deadline，HTTP 注册 method、插件相对 path、公开/私有及回调验证模式。复用既有完整 REGISTER 校验；同插件重复 ID/HTTP 路由、无效表达式和不合法 endpoint 使整次注册失败。路由描述准备完成不代表任务已启动或代理已公开。
2. **实现下一次触发计算。** interval 至少 1s，首个 tick 为激活时刻加间隔，后续沿固定起点计算；cron 固定五字段，默认 UTC，显式时区必须为有效 IANA 名称。按既有日/星期 OR、其他字段 AND 的规则求下一真实时间点，处理夏令时缺失和重复时刻。激活、恢复或落后时只安排下一未来 tick，不持久化漏 tick，不补跑。D-125 起 interval 的起点保存在 Core 状态库，不再随每次激活重置（步骤 9）。触发计算接受明确的当前时间，便于验证时间边界；它放在 `contracts/schedule`，供 Core Scheduler 与官方 Agent 组件的定时任务共用，不读当前时间、不持有状态。
3. **实现实际运行槽与调度调用。** tick 使用 Core→Plugin REQUEST 触发 callback，无用户 Actor、不默认 Admin；每次采用独立 run_id 和 15 分钟 job deadline。不同 job 可并发，同一 job 有实际运行记录时跳过本轮。超时先结束调用结果并取消，但保留运行槽和原调用键；SDK 在用户 callback 真正返回后才发送唯一 Response，迟到响应只释放槽。RPC 断链不能释放仍存活进程的槽，必须等真实完成响应或确认旧进程退出；恢复后仍从未来 tick 开始。
4. **实现 SDK 托管 HTTP Listener。** 仅存在 HTTP 路由时绑定 `127.0.0.1:0`，保持 Listener 打开并在 REGISTER 上报实际端点；插件作者只提供 Handler。用户 Handler 外包 Session 认证检查和就绪门禁；准备阶段即使端口已监听也不执行业务。Listener 与 Session 生命周期绑定，停止、错误和激活失败均关闭，不增加插件端口配置。
5. **实现 Core 的 Session 代理映射。** Core 从 `<data_dir>/alonabot.yaml` 读取 listen、两个 Base URL、可信反代与日志设置，文件不存在时按 CONTRACTS §15.1 的缺省值创建，命令行参数只覆盖本次运行。校验注册端点为合法 loopback 后保存，只有当前 Activated Session 能接单。公共前缀固定 `/plugins/{id}/...`，Core 匹配插件、method、route 和授权，剥掉插件前缀后交给标准库 ReverseProxy.Rewrite。覆盖内部认证头，清除 admin Cookie/Authorization 及未验证的转发身份信息，保留 Core 路径。公开 OAuth/Webhook 只对显式声明的路由生效，示例与测试覆盖插件执行 state/签名/nonce 验证；回调 URL 只从受控 Base URL 构造。D-82 起设置文件改为 `core_config.json`，并集中 Core Admin、命令前缀、默认 locale、按来源的日志级别与 Adapter 实例（T-07.6）。
6. **完成两个 Host 的页面边界与插件授权。** 同一公开 Listener 按受控 Host 分配 Admin 和插件内容，两个 Base URL 使用同 scheme/有效端口、不同 host；插件 host 不提供 Admin API。实现 60s 一次性插件兑换票据、独立插件 Cookie、原管理会话失效后的派生授权撤销，以及票据 URL 清理和 no-referrer。管理会话校验接 T-09 的认证入口，T-07 测试注入的会话只存在于测试装配。禁止向 plugin Origin 开放管理凭证 CORS，限制插件 Set-Cookie；未配置第二 host 时保留 JSON API/Webhook，按契约拒绝主动页面并施加响应策略。D-84 改为单一主机名：删除插件 host、兑换票据与插件会话，插件路由的全部响应加不含 allow-same-origin 的 CSP sandbox 与 nosniff 并去掉 Set-Cookie，非公开路由用管理会话与 CSRF（T-07.7）。
7. **接入统一启停和清理。** 在 T-02 的 Activated 转换后启用 Scheduler 和代理；Draining 时停止新 tick 和新 HTTP 派发，停止阶段取消余留调用、撤销映射、关闭 Listener。复用全局 10s 停止预算，不能为 Scheduler、HTTP 和每个插件另起一轮独立预算。Session/Generation 变化后，旧 endpoint、凭证和运行记录不能继续接收新业务；HTTP 失败不得触发自动业务重试。
8. **执行时间、进程和 HTTP 边界验收。** 用给定时间验证 cron/interval/夏令时，用真实插件进程验证不合作 callback、迟到 Response 和断链后的运行槽；用定时与 HTTP 示例验证作者不管理端口。HTTP 测试覆盖未激活、停用后请求、伪造内部头、私有路由、公开回调，以及插件响应的沙箱头与 Set-Cookie 过滤。分别记录 T-07 组件结果与仍需 T-09 首次认证后完成的端到端结果。
9. **动态任务与调度持久化的协议和 Core 侧（D-124～D-127，T-07.8）。** proto 按 WIRE-FORMAT §3.3、§3.5 追加 RouteKind 的 TASK、TaskSpec 与 ScheduledPayload 的任务字段并重新生成。状态库新增两张表：动态任务表以（plugin_id、任务名、key）为主键，保存触发规则、时区、参数、起点与创建时刻；调度状态表以（plugin_id、job_id）为主键，保存静态 interval 任务的起点、对应的间隔与暂停标记。调度器以状态库为准：Activated 时读出该插件的静态任务状态与动态任务实例，起点缺失或间隔变化时写入本次激活时刻，再按步骤 2 的计算安排下一未来触发；过期的一次性实例在这时删除并记 WARN 与计数，未注册任务名的实例保留、不安排。实现 `core.task.set`、`core.task.cancel`、`core.task.list` 与 `core.job.pause`、`core.job.resume`、`core.job.status`，按 CONTRACTS §4.3 校验 key、触发规则、参数大小与 `TASKS_PER_PLUGIN`，调用方只能操作自己的任务，写状态库成功后才更新内存中的定时器。动态任务的派发经每插件 `TASK_CONCURRENCY` 的名额，其余按到期时间排队，排队中的周期实例只保留一次；一次性实例在派发的同一事务中删除。`cancel` 删除行、摘除定时器与排队项，并取消该实例在途的调用；暂停只让触发作废，不打断在途的一轮。Draining 与停止时摘除该 Session 的全部定时器并清空队列（步骤 7）。把任务与调度状态加入管理服务（T-09.4）`--purge-data` 与 `--purge` 的清理对象：按 plugin_id 删除两张表中的行，与其余数据的删除在同一个状态库事务中完成，并列入确认对象；删除不依赖插件进程，也不要求 Session 存在。三项上限登记到 `internal/tune`。
10. **动态任务与暂停恢复的 SDK（D-124、D-126，T-07.9）。** `sdk.Task[T](name)` 声明任务，`sdk.OnTask(r, t).Handle(fn)` 在 `New()` 中注册处理函数（`r` 为 `sdk.Registrar`，D-114），Collect 时登记路由 `task.<name>`，同名重复、名称不合规与保留前缀冲突使收集失败。`t.At`、`t.Every`、`t.Cron` 把参数编码为 JSON 对象后调用 `core.task.set` 并返回 key；`sdk.Key` 指定 key，不指定时 SDK 生成随机 key；`sdk.Timezone` 只对 `Cron` 有效。`t.Cancel` 与 `t.List` 对应 `core.task.cancel` 与 `core.task.list`，List 返回带类型的参数、触发规则与下一次触发时刻。触发到达时 SDK 按任务名找到处理函数，把参数解码为 T 后调用，解码失败返回 INVALID_ARGUMENT 并记日志，不调用处理函数。`sdk.PauseJob`、`sdk.ResumeJob`、`sdk.JobStatus` 对应三个 `core.job.*` 服务。全部函数只在 Activated 之后可用，之前调用返回 FAILED_PRECONDITION。定时示例加入一个提醒命令，演示 At 与 Cancel；开发文档同步。

#### 可审查交付

- Scheduler 注册、下一 tick 计算、运行槽和 callback 包装的代码；能展示不同 job 并发以及超时后同 job 不重叠的示例和测试记录。
- SDK HTTP Listener、Core Session 代理映射、受控 Host 分流、插件票据/会话和停止清理代码；示例只声明 Handler，不声明内部端口。
- 可查阅的授权接口及 T-09 接入点、默认拒绝行为和测试装配；公开/私有路由与管理 Origin 的验证结果分开记录。

#### 对照现有验收

| 现有条目 | T-07 要完成的验证 |
| --- | --- |
| V-20 | 不同 job 并发、同 job 跳过重叠、cron/interval/IANA 时区/夏令时、离线和落后不补跑 |
| V-114 | 步骤 9、10 中动态任务的增删与触发、调度状态的持久化与错过即跳过、静态任务的暂停与恢复；`--purge-data` 与 `--purge` 删除任务与调度状态、只卸载时保留但不触发 |
| V-46 | 超时后 callback 未结束、迟到 Response、断链但进程仍存活时，实际运行槽的保留与释放 |
| V-21 | Scheduler/HTTP 服从就绪顺序、失败隔离与 Core 多插件共享的停止预算 |
| V-22 | 自动 loopback Listener、无需插件端口配置、Session 代理、公开/私有路由和停止清理；私有页面与真实管理员登录的完整连通在 T-09 联验 |
| V-47 | 插件响应的 CSP sandbox 与 Set-Cookie 过滤、内部头/管理凭证隔离、非公开路由的会话与 CSRF；首次 token、reset、真实登录/CLI 与原会话撤销在 T-09 完成端到端验收 |

#### 停止和失败条件

定时与 HTTP 组件完成上述验收，正常运行和停止都复用 T-02 的状态机，且 T-09 只需接入真实认证和页面即可集成时结束本项。若同 job 在实际 callback 未结束时再次启动、离线 tick 被补发、准备阶段公开路由、旧 Session 仍能代理、插件取得管理凭证或运行主动内容于 Admin Origin，必须修复后再交付。授权入口尚未接通时保持拒绝并标注联验未完成，不以测试会话宣称首次登录已可用；不为完善本工作包加入多用户账号、每插件独立域名或持久任务队列。

### T-08：配置、Secret 与数据库候选准备

#### 前置输入与边界

直接依赖 **T-04**，并复用其依赖链已交付的 T-01/T-02；可与 T-05 后续工作并行，不等待 T-06/T-07 或 T-09。

| 前置输入 | 开始本任务前须能取得的成果 |
| --- | --- |
| T-01 契约与依赖验证 | 可编译的 Bootstrap/REGISTER 类型、Config Schema 与 manifest 描述；锁定版本的 GORM、SQLite/Backup API；真实 CRUD/事务/备份小验证。 |
| T-02 基础运行时 | Core 状态库存取、Generation/Session 标识、逐插件操作锁、`core.lock`/`runtime.lock`、阶段预算与准备期业务门禁；可在临时数据目录启动测试插件并观察失败结果。 |
| T-04 SDK/生成器 | 单 root/至多单 Config 的发现与注册、构建期类型检查、可扩展的配置字段描述；不重复制作第二套扫描器。 |
| 现行规格 | CONTRACTS §5/§6/§12、WIRE Bootstrap 与 REGISTER、TEST-MATRIX V-23～V-26/V-38/V-39/V-54；迁移接口（`sqlitedb` 与 `gormdb` 两组）、JSON 语义、默认驱动与 external 插件的维护分支以现行契约为准。 |

**本任务交付候选准备能力，不提前实现 T-09 的正式 Journal 和 Active 提交编排。** T-02 只有基础代次/会话原语，不能据此宣称已经完成升级恢复。T-08 对外提供 Pending 导入、候选配置/DB 准备、资源引用与清理能力；T-09 决定何时停止旧版、创建正式操作记录、提交 Active、恢复和回收。测试驱动可以调用这些能力，但不能成为第二套生产激活入口。

#### 实施动作

1. **完成配置 schema 与构建检查。** 配置由与 `plugin.yaml` 同目录的手写 `config_schema.json` 声明（D-76，CONTRACTS §5.1）。`alonabot build` 与 `check` 按 §5.1 校验类型、字段键与模板键、各键的适用类型、default、options 与 labels、slider、condition、display_item、secret 与 required，错误带 JSON 路径；通过后把规范内容嵌入程序，`config_schema_hash` 按保留键顺序的规范 JSON 计算，必须与安装包中的文件一致。i18n/ 的 `config` 覆盖按 §5.3 与 schema 同形核对。生成器的 go/types 类型推导只供 Service Schema 使用（T-04.7），不再生成配置 Schema。T-08.11、T-08.12 已交付的配置版本、声明式迁移与备份由 T-08.16 按 D-161 删除，现行配置准备使用 §5.5 的校正规则（步骤 13）。
2. **实现完整 JSON 与 Pending 编辑能力。** 以原始对象保留嵌套未知键，区分缺失键与显式 `false/0/空串/空集合/null`；实现完整替换、RFC 7396 Merge Patch 和编辑快照导入。提交检查 `base_generation`，旧基线拒绝；直接编辑 `config.json` 不更改 Active。提供供 T-09 管理服务调用的应用层入口和结构化错误，不在此阶段开放未经认证的独立 HTTP 写接口。
3. **接入 SDK 配置准备顺序。** 使用 T-02 的私有 Bootstrap，按 `OnConfigLoad(ctx, raw)` → 补缺失值与校正 → required 校验处理，再进入该启动原因允许的 DB 准备；补值、校正、required 校验与 secret 路径的实现放在 `contracts/`，SDK 与 Core 共用。完整规范 JSON 随 REGISTER 回传，禁止用只含已声明字段的结果替换原始文档；新 schema 检查必须在 OnConfigLoad 之后。准备完成后由 `p.Config()` 与 `c.Config()` 提供按类型取值的只读配置。普通 restart 重验原 Active，OnConfigLoad 或校正若改变完整运行配置则返回 FAILED_PRECONDITION；所有阶段错误、超限及超时走既有控制错误，准备期不能发普通业务 RPC。
4. **实现不可变 Secret Store。** 新 secret 值生成新引用，Core 在所属 Session Bootstrap 前解引用，在 REGISTER 后按新 Schema 和保留的旧 secret 路径重新抽取；SDK 不计算持久化引用摘要。检查引用归属，管理读取返回引用/设置状态，掩码不成为保存值；日志、错误和普通响应不含明文 JSON、DSN 或 secret。实现运行账号权限/Windows ACL 与按引用清理能力；清理必须保护 Active、Pending、保留代次及恢复中操作引用，正式 Journal 引用由 T-09 接入。
5. **实现受管 SQLite 的两个入口与连接生命周期。** 新建 `sdk/sqlitedb`：`sqlitedb.Open(ctx)` 按 Session 的 DB 描述打开受管 SQLite，同一 Session 复用同一个连接，SDK 停止时关闭；现有 `gormdb.OpenSQLite` 的连接设置与回读核对移入此包，按契约应用 WAL/FULL/外键/5s busy timeout 与单连接。`gormdb.Open(ctx)` 经 libtnb 的 `Conn` 包装同一个 `*sql.DB`，不另开连接。插件不可取得 Core 状态库连接，事务回调使用传入的 tx。连接或数据损坏时返回错误，不能创建空库掩盖问题。只导入 `sqlitedb` 的插件不链接 GORM，用 `go list -deps` 验证。
6. **实现两组互斥迁移与维护门槛。** `sqlitedb` 与 `gormdb` 各定义 SelfMigrator 与 VersionedMigrator，被导入时向 SDK 登记识别方式，SDK 本身不 import GORM；根插件最多实现一个迁移接口（D-114 把迁移接口改为登记函数，由 T-04.12 修改）。versioned 模式校验非零、唯一、递增 ID，每步 Up 与 `_alonabot_migrations` 完成记录使用插件 DB 的同一事务，失败不记成功，两个包共用同一账本格式。`prepare_effects=external` 的插件在已有安装的候选上没有维护许可时返回 NEEDS_MANUAL_MIGRATION；维护许可须匹配 Core 下发的本次 Session/Generation/DB；事务不支持的 SQL 不得偷偷转为事务外执行。
7. **形成可交接的候选资源结果。** 在已持逐插件操作锁、旧写入者退出且 Core 可取得 runtime.lock 的前提下，使用 Backup API 创建独立 SQLite 候选；首次安装创建空候选，reload/update 同样复制，普通 resume 复用 Active。复制完成后 Core 释放临时 runtime.lock，再由 SDK 取锁打开候选，避免互相等待。将准备好的程序/配置/secret/DB 引用、配置摘要、迁移完成状态与本次身份交给 T-09；本任务不写 Active 指针或伪造 COMMITTED。失败清理只处理未受保护的候选资源。
8. **实现用户可见文案目录。** 在 `alonabot build` 与 `check` 中检查 `i18n/*.json` 的 `messages`：en-US 必须存在且包含全部键，其他语言不得有 en-US 没有的键，同一键的占位符集合一致。SDK 提供 `T(key, args)`，按 Actor locale → Core 默认 locale → en-US 选择语言，主语言回落按 CONTRACTS §5.4；未知键与缺少参数返回错误。Core 内置 en-US 与 zh-CN 的用户可见文案目录，命令 usage 改为经该目录生成；Core 默认 locale 是一项全局设置。
9. **完成示例与分层验证。** 在临时 data_dir 上分别用 `sqlitedb` 与 `gormdb`，运行配置回调与校正、显式空值、未知 secret、CRUD/事务、数据库迁移失败以及候选 OnStart 写 DB 后失败的案例。为自动表单提供字段顺序、invisible、模板、i18 回退、掩码和 PATCH 的组件/fixture 验证，并建立 WebUI 文案的 en-US/zh-CN 目录与语言切换组件；`web/admin/` 是官方 WebUI 组件的源码；真实管理页面在 T-09 接入。记录可独立通过的组件证据，以及仍需 T-09 正式管理入口和 Journal 才能关闭的场景，不把后者标成已通过。
10. **文案函数只返回字符串（D-106）。** `c.T` 与 `sdk.T` 改为只返回字符串；生成器在 `check` 与 `build` 中检查以常量写出的键与参数名，问题带位置并使构建失败；运行时仍未知的键返回键名，缺少的参数保留占位符原文，并在插件日志记 WARN。Core 自己的文案目录不变。同步仓库内的全部调用。
11. **配置项的选项来源（D-129，T-08.14）。** 生成器的 schema 检查接受 `options_from` 并按 CONTRACTS §5.1 校验：只用于 `string` 与没有 `items` 的 `list`，不与 `options`、`labels`、`secret` 并存，来源名格式合法，内置来源的名称与 filter 的键已知。Service 描述追加选项来源标记；`sdk.ServeOptions(r, serviceName, fn)` 注册固定 Schema 的选项来源服务，访问规则固定为 Core 与本组件。Core 新增按组件与字段路径查询选项的入口：从 Active 代次的 schema 取来源名与 filter，内置来源在进程内查询，服务来源经 Service Broker 以 Core 的身份调用且只接受带标记的服务；条数上限取 `OPTIONS_PAGE`，多出的选项截去并标记还有更多；目标不可用时返回 UNAVAILABLE。内置来源由登记处管理，本切片登记 `core.adapter_instances`，`core.chats` 由 T-03.25 登记。配置的补值、校验与读取不变。管理 API 由 T-09.14 提供，页面由 T-09.9 提供。
12. **配置字段的回复示例（D-160）。** 配置 schema 检查 `example` 的结构、片段、占位符、条件引用与类型以及花括号转义；通用 i18n 目录检查覆盖键和文案类型，`Schema.CheckI18n` 核对源结构、片段位置与占位符。作者工具与安装预检检查两种 schema（`config_schema.json`、`adapter_config_schema.json`）引用的图片，打包收集这些图片。Core 按 Active 代次的已提交 Journal 选包，只提供包中 `config_schema.json` 引用的图片，响应头同组件图标。表单渲染与草稿值替换由 WebUI 完成，不调用插件。
13. **去掉配置版本与迁移（D-161）。** 删除 `config_schema.json` 与 `adapter_config_schema.json` 的 `$version`、`$migrations` 及其检查，manifest 的 `config_version`，Bootstrap 的 `previous_config_version`、`target_config_version` 与 REGISTER 的 `candidate_config_version`（保留字段号），`core_config.json` 与 Adapter 分组的 `$version`，以及 `contracts/` 中的迁移实现与备份；首次发布前不留兼容。`contracts/` 新增按 CONTRACTS §5.5 的校正：缺失补 default、无损转换、恢复 default、移除失效模板元素，并产出校正记录。SDK 的 `OnConfigLoad` 签名为 `func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error)`，在校正之前执行；共享校正入口为 `Schema.Correct(raw json.RawMessage) (json.RawMessage, []Correction, error)`。REGISTER 以字段号 8 追加 `repeated ConfigCorrection config_corrections`，Core 将记录与配置快照、秘密路径同事务存入代次元数据，普通 restart 复用记录。Core 用同一实现校正 Adapter 实例配置与 Core 设置，不改写 `core_config.json`，不是合法 JSON 对象时拒绝启动。`File.PrepareInstances(adapterID string, schema *configschema.Schema, replacement ...map[string]json.RawMessage) (map[string]json.RawMessage, []InstanceProblem, error)` 独立校正各实例的运行副本；显式 replacement 经 `File.Update` 提交，其他实例的原文保留，`TrialInstances(adapterID string, schema *configschema.Schema) []InstanceProblem` 只读试算。管理 API 返回校正记录与无效配置的路径，并提供三个清理入口（ADMIN-API §12.1）；插件从 Active 清理后写 Pending，开发组件 reload 原子发布 Pending 到编辑文件，启动成功后清除 Pending。
14. **插件的 KV 与数据目录（D-165）。** `sdk/sqlitedb` 初始化受管 SQLite 时建保留表 `_alonabot_kv`，不占用插件迁移 ID；`sdk/kv` 提供 Get、Set、Delete 与按前缀分页的 List，构建时按链接受管 SQLite 处理 `prepare_effects`。`internal/app` 在候选与常规启动装配时建立 `<data_dir>/plugin_data/<id>/`（Unix `0700`，Windows 继承状态根的账号专用 ACL），经 Bootstrap 的 `plugin_data_dir` 下发绝对路径；离线 `alonabot test` 在受保护的临时状态根内创建并下发目录。`sdk.PluginDataDir(ctx)` 优先返回 `ctx` 所属 Session 的目录，独立进程可使用默认 Session，就绪前以英文消息 panic。disable 和普通卸载保留目录，`--purge-data` 与 `--purge` 经管理服务确认后删除。configdb 示例演示 KV 与文件读写。

15. **配置文件的存储与管理（D-168，T-08.18）。** 在 Core 状态库（并入初始迁移）建立配置文件、实例缺省与单独指定、按（插件、配置文件、字段路径）的覆盖值、插件开关、工具策略覆盖与平台设置的表；`default` 不能删除，没有自己的覆盖值与平台设置。按 CONTRACTS §5.6 解析会话所用的配置文件，删除配置文件、Adapter 实例与插件 `--purge` 时清理对应记录。schema 检查接受 `profile` 并按 §5.1 校验位置；写入覆盖值时用 Active 代次的 schema 校验，插件更新后失效的覆盖值保留并列为无效，提供清理。实现 ADMIN-API §23 与 `alonabot profile`：`default` 的字段值经 §12 的配置 API，平台设置写回 `core_config.json` 的 `command_prefixes` 与 `reply`；会话数按聊天目录计算。

#### 拟目录与交付

| 实施拟建或扩展路径 | 本任务职责 |
| --- | --- |
| `contracts/`、`sdk/config/` | `config_schema.json` 的解析与检查、补缺失值与 schema 校验（SDK 与 Core 共用）、按类型取值、原始对象保留及对应测试。 |
| `sdk/sqlitedb/`、`sdk/gormdb/` | 受管 SQLite 的 Session 连接、`database/sql` 与 GORM 两个入口、两组迁移接口与执行器、集成测试。 |
| `internal/core/config/`、`internal/core/secret/` | Pending/候选配置、引用归属、不可变 secret、脱敏读取与受保护引用清理。 |
| `sdk/`、`internal/core/i18n/` | SDK 的 `T` 与 `messages` 加载；Core 的 en-US/zh-CN 用户可见文案目录与默认 locale。 |
| `internal/core/storage/` | 受管 DB 描述、SQLite Backup/候选文件准备、external 插件的维护准入；复用 T-02 状态/锁原语，不另建状态库。 |
| `web/admin/` | 配置 Schema 消费、表单/i18 组件和 fixture；正式登录及管理 API 绑定属于 T-09。 |
| `examples/plugins/configdb/` | 使用真实 SDK 的配置与数据库示例；失败点由测试夹具控制，测试数据不使用生产凭证。 |

交付包括可编译实现、生成资源、示例、测试夹具，以及一份写在代码/测试附近的 T-09 调用约定：输入身份和旧代次、锁与旧进程前置条件、输出资源引用、错误及清理所有权。该约定是既有契约的实现交接，不建立新的公开插件 API 或平行规格。

#### 对应 V 验收与停止条件

| 验收编号 | T-08 必须取得的证据 | 后续闭合 |
| --- | --- | --- |
| V-23、V-38 | 原始/未知键保留，空值/默认/Merge Patch 正确，陈旧基线拒绝，restart 不改 Active。 | 真实 `plugin reload` 与 WebUI 提交链路在 T-09 复验。 |
| V-24 | 独立候选 DB 中 OnStart 写入后失败，旧 DB/配置/secret 不变；锁阻止同时写入。 | Core 在正式候选流程中被 kill 后的权威代次选择，在 T-09 结合 V-30 完成，T-08 不提前关闭整项。 |
| V-25 | 新旧 secret 引用分离、未知 secret 仍保密、invisible/i18/掩码与 PATCH fixture 正确。 | 真实页面、再认证与管理响应在 T-09 复验。 |
| V-26、V-39 | 两个入口的 CRUD/Tx 共用连接，迁移接口互斥，Up+账本同事务，external 插件维护许可的有效性。 | T-09 补自动更新及显式维护的真实入口验证。 |
| V-21、V-37 的配置/迁移分支 | 使用 T-02 会话验证阶段错误/超限/超时无半注册、无准备期业务请求。 | T-09 验证与正式提交前后恢复门槛组合。 |
| V-54 | 文案选择顺序与主语言回落、check 的键与占位符检查、Core usage 文案、WebUI 文案组件。 | 真实页面的语言切换在 T-09 复验。 |
| V-104 | `c.T` 与 `sdk.T` 只返回字符串；check 与 build 检查常量键与参数名；运行时问题回落到键名或占位符原文并记 WARN。 | — |
| V-116 | `options_from` 的检查；`sdk.ServeOptions` 的注册与访问规则；Core 按字段路径查询内置来源与服务来源，条数上限与来源不可用。 | 管理 API 与页面在 T-09 复验；`llm.models` 由 T-12.15 提供（T-12.8 交付的 `llm.providers` 由它改名，D-154）。 |
| V-140 | 步骤 15 中配置文件的存储、解析、schema 的 `profile` 检查、覆盖值校验与清理（含复合 secret）、实例清理与提交边界、管理 API 与 CLI | 派发与 SDK 合并由 T-04.18 提供，页面由 T-09.25 提供 |

建立上述包后，后端验证入口至少包括 `go test ./sdk/... ./internal/core/config/... ./internal/core/secret/... ./internal/core/storage/...`。前端使用实际建立的组件测试入口，记录准确命令。所有结果按 TEST-MATRIX 记录代码位置、命令、版本、环境和日志；尚无环境时明确保持未运行。

出现候选污染旧数据、未知键/显式值丢失、secret 越权/明文泄漏、迁移成功记录与实际事务不一致、external 插件未经许可就准备候选、无法验证驱动/备份正确性时，停止向 T-09 交付并修复对应实现。缺少 T-02 锁/会话原语时先回补 T-02，不在本任务用睡眠、PID 猜测、复制 WAL 主文件或临时直写 Active 绕过依赖。

### T-09：安装分发、正式激活恢复与管理入口

#### 前置输入与边界

直接依赖 **T-06、T-07、T-08、T-11**。T-02 基础状态、锁和 Session 能力已在前置链中实现；**本任务扩展 T-02 的同一提交事务原语，完成带候选资源引用的正式 Journal、Active+COMMITTED 提交编排及崩溃恢复**；T-02 的最小提交切片和 T-08 的资源准备不能替代这些集成工作。

| 前置输入 | 开始本任务前须能取得的成果 |
| --- | --- |
| T-01/T-04 构建契约 | manifest/Schema/注册描述解析器、生成器、构建输入约束和已锁定工具链；运行期显式注册的完整校验仍在 REGISTER。 |
| T-02/T-06 基础运行时 | 会话/阶段原语、状态库、锁、普通请求与 Service/Bus 的停止准入接口；不在 Journal 事务中等待 IPC。 |
| T-07 HTTP/调度 | 当前 Session 代理与任务门禁、Core 授权检查接口、票据/插件会话和拒绝路径。T-07 可用测试会话验组件，但生产没有有效管理会话必须拒绝。 |
| T-08 候选能力 | Pending/配置准备、不可变 secret 引用、受管 SQLite 副本、两个数据库入口/迁移与 external 维护准入、候选结果和清理所有权。 |
| T-11 Agent 基础 | 工具调用策略与 Agent 会话的管理服务层。 |
| 现行规格 | CONTRACTS §7.3/§7.4/§9/§12 与 TEST-MATRIX V-27～V-31/V-40/V-41/V-47；首次管理凭证与真实登录由本任务落地。 |

#### 实施动作

1. **落地首次管理认证与真实入口。** 实现 `admin-token init/reset` 的离线 core.lock 检查、32-byte token 一次终端输出、仅存 SHA256 和认证代次；reset 撤销旧管理会话。D-82 起非终端首次启动把 token 写入数据目录的 `admin_token` 并提供 `admin-token show`（T-07.6），D-84 删除派生票据与插件会话（T-07.7）。实现 POST login、Origin/登录 CSRF、恒时摘要比较、8h 会话、登出撤销与 token 再认证查看 secret；CLI 使用显式 Bearer token。接 T-07 的授权检查，未初始化或无有效认证时拒绝管理访问；纯 loopback 开发与非 loopback HTTPS 边界沿用契约。
2. **实现开发 CLI 与安装预检。** 开发命令是对四种组件通用的顶层命令（D-77）：`init` 在当前目录创建只含作者信息的组件工程（`plugin.yaml`、根插件与示例命令、示例 `config_schema.json`、`i18n/en-US.json` 与忽略 `data/` 的 `.gitignore`，不生成 `main.go`）；`build` 生成嵌入代码与调用 `sdk.Run` 的正式入口并编译，取代 `plugin generate`，产物不包含 Core；`check` 只核对、不写文件；`test` 与 `pack` 接到同一构建。包内 manifest 由 build 与 pack 补上构建信息（D-78，T-09.16）。仓库内示例复用根 module，面向外部作者的脚手架沿用插件自身 Go Modules 的既有方式。pack 支持四种角色的组件，WebUI 包只含静态文件并以 `any` 为平台，生成 tar.gz 与归档外 release 清单，签名可选（D-81），有签名时对清单确切字节签名；安装预检有签名时先做 Ed25519 验签，再核对归档 SHA256/长度、来源、平台和 manifest。受控解压拒绝越界、链接、设备、重复及大小写冲突路径，并执行 256 MiB/1 GiB/10,000 条目上限；所有下载/身份/静态错误必须在停止旧版前拒绝。
3. **建立唯一管理服务与安装记录。** CLI/WebUI/API 调同一管理服务，四种角色的组件使用同一套安装、启停、更新与卸载流程，所有变更按包串行；按 CONTRACTS §9.1 执行同类唯一：同类组件已安装时 install 返回 ALREADY_EXISTS，`--replace` 在新组件通过预检后按 uninstall 的缺省语义卸载旧组件再安装，旧组件原来启用时启用新组件，确认对象列出被替换的组件及不迁移的 Adapter 实例；`--source owner/repo` 从 GitHub Release 下载并持久化安装来源，不给时安装本地包，update 只从安装时的来源取版本；`alonabot run`（不带子命令时等同，D-77）复用 T-13 的 Core 装配，按安装记录启动全部已启用的包；Core 装配读取 `core_config.json`，把命令前缀与 Core Admin 列表接入 Router 与策略，启动组件时传入默认 locale（T-07.6 只提供读取）；Adapter 实例的创建、启停与删除也经此服务，写回 `core_config.json` 的 `adapters`（D-82，T-07.6）；启用插件时检查其必需的 Adapter 依赖与插件依赖；卸载被依赖的 Adapter，或停用、卸载被依赖的插件、把它更新或回退到低于 `min_version` 的版本时，须列出受影响插件并显式确认。持久化安装序号、来源/受信公钥、pin、安装/卸载保留状态。实现 install/list/enable/disable/restart/reload/update/uninstall/pin/rollback 的既定区分：重启原 Active 只换 Session，reload 导入编辑快照；卸载按 D-127 分三种清理范围：缺省只卸载，保留数据、配置及绑定，仅原来源可显式重装；`--purge-data` 删除数据，`--purge` 删除数据与配置，两者只删除 CONTRACTS §9.3 列明的框架受管对象，状态库中的对象在一个事务内删除；任务与调度状态的表由 T-07.8 加入清理对象。首次信任（含未签名来源，D-81）、公钥变化、恢复丢弃新写入等确认只用于契约要求的具体操作，不增加一套通用审批系统。
4. **实现正式候选 Journal 与准备编排。** 通过预检后，以预期旧 ActiveGeneration 开始持久化管理操作记录及其资源引用；冻结新根派发，drain/停旧进程，确认树退出及 runtime.lock，再调用 T-08 复制 DB 和创建候选 Session。将启动原因、维护许可和阶段预算放入已有 Bootstrap；REGISTER 校验后持久化程序/配置/secret/DB 候选引用并形成 PREPARED。普通 resume 走原提交记录，不伪造新候选 Journal；`prepare_effects=external` 的候选在维护门槛停下。
5. **实现唯一提交点与激活门禁。** 确认候选文件及引用已持久化后，在 Core 状态库的同一事务复核预期旧代次并提交 ActiveGeneration+COMMITTED；事务内不做下载、备份或等待 RPC。随后发送 Activate，只有 OnReady/Activated 完整成功才开放命令、事件、Service、Bus、任务与 HTTP；复用 T-06/T-07 门禁。旧按钮按 Generation 失效；提交后启动/OnReady 失败保持新代次权威并进入 Degraded，不自动倒退到旧 DB。
6. **实现崩溃恢复、显式回退与引用回收。** Core 启动持锁后按状态库 Active/Journal 恢复：未提交的已有安装关闭候选后以旧程序启动新 Session，首次安装尚无 Active 则保持未激活；已提交只恢复当前代次。旧进程退出未确认时不启动新写入者。rollback 从选定保留代次复制新的 SQLite 恢复候选并使用新 Generation，继续走同一准备/提交路径；外部数据的恢复沿用显式维护。只有无人引用的失败候选/旧包/secret 才可回收，恢复中操作和保留快照受保护；备份说明覆盖整个 data_dir 与插件自管的外部数据。
7. **完成最小 Admin 页面与 API 绑定。** 页面属于官方 WebUI 组件（`web/admin/`），Core 管理操作调用 Admin API，统计、商店与已读版本调用 WebUI 自身后台（ADMIN-API §25）。接入真实登录、状态/日志、安装启停、更新/恢复、配置表单/完整 JSON 编辑、命令权限、命令名覆盖（D-128）、配置字段的选项查询与可搜索选择（D-129）、插件的发信许可（`allowed_adapter_instances`）、工具调用策略，以及界面语言切换。数据库表格、KV/文件存储、导出与清空、Agent 会话摘要与清除归 T-09.23；会话正文归 T-12.20（goal-167），配置、任务与插件页面入口沿用 T-09.9，权限共用用户页。Core 设置页管理 Core Admin 列表、命令前缀与默认 locale，与 `alonabot admin add|remove|list` 使用同一管理服务，修改写回 `core_config.json`。使用 T-08 的 raw JSON/PATCH/base_generation/secret 语义，显示 Pending、候选失败、Degraded 与维护原因；动态权限复用 T-04。插件页面由 WebUI 用沙箱 iframe 嵌入，经 postMessage 桥以管理员会话请求本插件的路由（D-84，T-07.7）；管理 Cookie/Authorization 不进入插件。页面只展示用户可操作状态和具体错误，不把内部协议步骤当作用户配置项。**日志：** 按来源的级别写在 `core_config.json` 的 `log.sources`，修改后经 `sdk.log.set_level` 通知 SDK 并立即生效；Core 为每个来源保留最近 1000 条记录，提供按来源、级别与 trace_id 的读取和 SSE 订阅，过去日期的记录从日志文件按日期查询；`alonabot logs`、`alonabot log-level` 与 WebUI 日志页使用同一 API，页面的筛选只在浏览器中进行（CONTRACTS §14.3、§14.4）。 **诊断包（D-156，CONTRACTS §14.7）：** 管理服务按所选时间范围与三组内容打包 zip，缺省脱敏；`POST /api/v1/diagnostics` 与 `alonabot diagnose` 共用该实现，CLI 在管理连接被拒绝且取得已有 `core.lock` 的独占锁后只读收集，组件运行状态记为 `Stopped`。WebUI 日志页只有终端视图，另有「清空」：清空只丢弃页面已载入的记录，不删除日志文件，实时订阅继续。 **WebUI 进程与补充 API（D-162～D-164）：** WebUI 包可以带 Go 入口，进程与插件相同，路由要求管理会话；Core 在命令层得到 Handler 结果后生成 `command.invoked`；数据浏览、保留代次、静态任务、实例列表、聊天目录、更换令牌与插件页面列表见 ADMIN-API §20、§21。 **Changelog（D-166）：** Core 只按已安装的版本提供组件包内与自身的 changelog（ADMIN-API §22），不比较版本、不记录是否看过。 **模型页（D-154）：** 侧栏的一级入口「模型」编辑 Agent 配置中的 `providers` 与 `models`（经同一配置 API 的 Pending 与 reload），并调用 Agent 的模型管理路由（CONTRACTS §13.6）拉取模型列表、测试模型与读取当天用量；按 `role: agent` 找到已安装的 Agent，没有时提示安装，路由不存在时只提供配置编辑。 **品牌、静态资源与外壳（D-169，T-09.26）：** 按 Figma v4 提供浅深色视觉令牌及随包资源（CONTRACTS §7.5），提供客户端路由、桌面侧栏、顶栏搜索与手机导航。管理页面按 [WebUI 实施计划 §3](webui.md#3-页面接入外壳) 接入；主题切换控件随系统设置页接入。 **配置文件页（D-168）：** 侧栏「配置文件」打开即是当前配置文件的设置，标题处切换或新建，「N 个会话」打开使用范围对话框（实例缺省与单独指定的会话），更多菜单为重命名、复制与删除（`default` 不能删除）；分区为 AI（已安装 Agent 的标记字段，没有 Agent 时不显示并提示安装）、平台与消息（命令前缀与四项回复方式）与插件（按插件的开关、标记字段的覆盖与无效覆盖的清理）；插件配置页可以切换配置文件，选中 `default` 以外的配置文件时只能编辑标记的字段；系统设置页不再有命令前缀。 Admin API 与日志管理在本章节（T-09.10、T-09.13、T-09.14、T-09.19～T-09.22、T-09.24），页面与官方后台在 WebUI 章节（T-09.8、T-09.9、T-09.15、T-09.23）。
8. **完成正式流程的故障注入与交付证据。** 通过真实 CLI/API 在临时目录跑首次安装→enable→restart→reload→update→pin→disable→同来源重装→显式 rollback，并以包损坏/签名变化/external 插件的维护错误验证拒绝路径。在 PREPARED、提交前后和 OnReady 失败点终止 Core，连同残存旧进程、空间不足、并发同插件操作一起验证无双写、无提前开放、无静默数据回滚。回填 T-08 留待集成的 V-23～V-25/V-39 证据，形成 T-10 可重复使用的发布产物与恢复场景；五平台真实发布结果仍由 T-10 完成。
9. **WebUI 组件与组件图标。** 按 CONTRACTS §7.5 由 Core 提供已启用 WebUI 组件的静态文件：保留 `/api/` 与内部路径，不存在且无扩展名的路径回退 `index.html`，未安装时返回 503 与英文提示，Admin API 主版本不在 WebUI 声明的范围内时拒绝安装与启用并返回提示。Admin API 定义主版本号常量，并在 `spec/ADMIN-API.md` 写出全部接口、认证、错误格式与插件页面桥的消息格式，官方 WebUI 只使用其中的接口。图标只提供 manifest 的 `icon` 声明的路径（表单选项不带图标，D-74），带 sandbox CSP 与 nosniff，并要求管理员会话。库模式增加 `alonabot.WebUI(fsys fs.FS)` 选项，以 dev 来源提供开发中的 WebUI。

10. **包校验收紧（D-150、D-151）。** 协议不变。`contracts/manifest` 对缺少 `migration_mode` 或 `prepare_effects` 的包只给警告，文案写明准备与回退按没有框架外副作用处理；管理服务把预检得到的警告放进安装与更新的结果（Admin API 增加 `warnings`，ADMIN-API 同步），CLI 原样输出，并记 WARN 日志。`packages.Preflight` 对进程角色的包严格核对 build-info：存在、可解析、与 entrypoint 内嵌的构建信息逐字一致、`GOOS`/`GOARCH` 与 manifest 一致，任一项不符时在停止旧版前拒绝；WebUI 包不要求 build-info。

11. **组件市场 ID 改用连字符（D-172，T-09.27）。** 四角色市场前缀、插件市场正则与 manifest 的角色后缀规则一起修改；本地 ID 语法和 64 字符上限保持不变。同步 MarketPrefix 的所有校验调用者、SDK 会话解析与聊天选项中的分隔符替换、日志来源标签和历史 source、Agent 的 memory 工具原始名、官方组件与示例 manifest、配置、路径、既有测试及 contracts/testdata（含 Protobuf golden 帧）。CLI init 与生成器通过原有 manifest 入口接受新 ID；当前 init 仍要求 --id，目录名缺省留给 T-13.6 并沿用本规则。组件 ID 只作字符串值，模型工具名继续由现有映射生成，Go 符号、环境变量名与 SQL 标识符不随之替换。直接改名并要求状态库与开发数据目录重建，不提供旧市场 ID 别名或数据迁移；只改写已有提交测试，边界与回归写在被忽略的 *_local_test.go，协议编号按 wirespec 核对。完整文件范围与真实签名见 goal-177。

#### 拟目录与交付

| 实施拟建或扩展路径 | 本任务职责 |
| --- | --- |
| `cmd/alonabot/` | admin-token 命令、开发 CLI、统一管理服务客户端；运行时状态变更不直接写 DB。 |
| `internal/core/packages/` | release 清单/签名、下载/限额/安全解压、来源绑定、不可变程序包；不建设中心商店。 |
| `internal/core/management/` | 逐插件管理动作、安装/pin/卸载保留语义、候选 Journal 编排、恢复与引用清理；复用 T-02/T-08 存储与运行时原语。 |
| `internal/core/auth/`、`internal/core/http/` | 补齐首次认证、管理会话/API，与 T-07 已建授权及代理包整合，避免并列两套认证。 |
| `internal/core/` 的 T-02 状态包 | 扩展 Journal、来源/保留引用、预期代次事务与恢复查询；沿用同一 state.sqlite。 |
| `web/admin/` | 官方 WebUI 组件：真实登录、管理页、配置/权限页及端到端交互；复用 T-08 组件。 |
| `spec/ADMIN-API.md` | Admin API 参考：主版本、全部接口、认证与错误格式（CONTRACTS §7.5）。 |
| `examples/plugins/`、相邻 `testdata/` | 可安装包、配置/DB 版本对、坏包与恢复故障夹具；不得包含生产 token、DB 或签名私钥。 |

交付为可运行的管理 CLI/API/WebUI、可安装的示例发布包、唯一候选提交/恢复实现、备份与维护操作说明，以及按 V 编号关联的测试证据。实现形成后，入口至少包括 `go build ./cmd/alonabot`、管理/包/认证包测试、真实临时 Core 上的 CLI 流程和浏览器端到端验证；把实际 CLI 参数、启动配置和前端测试命令记录在夹具说明中，不能依赖执行者猜测目录或凭证来源。

#### 对应 V 验收与停止条件

| 验收编号 | 本任务通过条件 |
| --- | --- |
| V-27、V-28、V-41 | 安全解压和上限生效；签名/摘要/平台/来源变化均停机前拒绝；无自包含摘要循环，无验签失败降级无签名。 |
| V-29、V-30，复验 V-21/V-24/V-36 | 真实 WAL 候选备份、空间/锁失败、逐提交点崩溃恢复符合契约；未提交选旧、已提交选新，旧进程残存时不另启写入者，未 Activated 不接单。 |
| V-31、V-39 | 首装、resume、自动候选与 external 插件的显式维护均有正确分支；非法维护许可及不可隔离生产变更明确拒绝；失败不假称自动恢复。 |
| V-40、V-48，复验 V-23/V-25/V-38 | 所有管理命令有真实状态变化和持久记录；restart/reload 区分、同来源重装、pin、恢复新 Generation/DB 副本与旧按钮失效正确；配置/secret 语义在真实页面/API 保持。 |
| V-62、V-63、V-68 | 按来源级别的持久化与运行时修改、最近记录与 SSE、按日期查询日志文件、CLI 与日志页、未认证拒绝。 |
| V-52、V-54、V-55、V-56 | 启用、停用、卸载与更新时的插件依赖检查和确认；界面语言切换；工具策略修改在下一次调用生效；会话清除与 purge。 |
| V-116、V-117 | 管理 API 按组件与字段路径查询选项、按存储值取显示名、来源不可用；`core.chats` 的搜索与筛选；页面的可搜索选择、手动输入与失效值的显示。 |
| V-145 | 配置文件页的一层结构、切换与使用范围、三个分区、没有 Agent 的状态与插件配置页的切换；`default` 的字段修改经 Pending 并提示 reload |
| V-82、V-83 | 同类组件不能并存，`--replace` 与 `--source` 的替换与后续更新正确；WebUI 静态文件、单页回退、未安装与版本不兼容的提示，图标的 CSP 与路径限制。 |
| V-47，复验 V-12/V-22/V-43 | 首次 init/login/CLI 可用，reset 撤销旧授权；管理状态变更经认证/CSRF；插件 host 不取得 Admin API/Cookie/凭证；真实页面修改权限在下一次准入生效。 |
| V-122（步骤 10） | 缺少迁移方式或准备副作用的包照常安装与更新，结果与日志带警告；build-info 缺失、无法解析、与 entrypoint 不一致或平台不符时在停止旧版前拒绝，已安装版本保持运行 |

协议/Session 鉴权、安装路径或来源校验、secret 泄漏、Active/Journal 原子性、崩溃后双写、提交后自动回退丢数据、维护虚假成功，任何一项失败均阻塞向 T-10 交付。组件缺口回补对应前置任务；不能把授权缺口用测试会话带入生产，不能借 T-10 再处理本任务的确定性数据恢复缺陷。尚无目标平台或浏览器证据时如实标注缺口并准备可重复夹具，不把“可编译”写成“已发布通过”。

### T-11：Agent 基础

**目标：** 在 Service Broker 与按钮 token 之上，提供 Agent 需要的工具目录、按工具的调用策略、调用前的用户确认和会话存储（CONTRACTS §13）。对话循环不在 Core 中实现。

#### 前置输入

- T-06 的 Service 描述、Broker 定位与准入、父调用链与取消；T-04 的策略持久化与管理服务层；T-03 的按钮 token、Target 准入与 Adapter 发送；T-08 的 Core 文案目录。
- 读取 CONTRACTS §3、§5.4、§13，WIRE-FORMAT §3.3、§3.5，MESSAGE-MODEL §6.3。

#### 实施步骤

1. **工具描述与注册校验。** Service 描述接受可选的 `tool`（description、default_policy）。只有 JSON 编码的 Service 可以声明工具；description 为空或策略取值非法时整次注册失败。Core 为已激活 Session 的工具建立目录，Session 失效时移除。
2. **调用策略与目录。** 以工具名持久化 allow/confirm/deny，未设置时取 default_policy；`core.tool.confirm` 的操作以「调用插件 ID + 操作名」为键，未设置时为 confirm。管理服务层提供查询与修改，修改在下一次调用生效。`core.tool.list` 只返回调用插件按 Service 访问规则可以调用、且生效策略不为 deny 的工具。
3. **调用路径。** `core.tool.invoke` 校验 parent_call_id 后按策略分支：deny 直接拒绝；confirm 进入步骤 4；allow 或确认通过后经 Broker 以父调用链调用目标 Service，结果与错误原样返回。`core.tool.confirm` 只执行策略与确认。两者都不重试工具调用。
4. **用户确认。** 父调用链没有可信 Actor 或默认 Target 时，confirm 直接返回 PERMISSION_DENIED。Core 创建确认记录和两个绑定该记录的按钮 token（沿用按钮 token 表），按 Actor locale 选择 Core 文案；invoke 展示实际的 JSON 参数，confirm 展示插件给出的摘要；确认消息直接经 Adapter 发送，不执行插件出站 Hook；目标不支持 buttons 时返回 UNSUPPORTED。只接受同一 Actor 在 5 分钟内的第一次点击；超时、父调用结束或 Session 失效时记录作废。
5. **会话存储。** 在 Core 状态库建立以（plugin_id、会话键）为键的条目表。`core.conversation.append` 在一个事务内追加并返回最后一条的序号，`list` 按序号分页并受帧上限约束，`clear` 删除全部条目。管理服务层提供列出与清除；卸载时默认保留，purge 时删除。D-127 起会话属于数据，`--purge-data` 与 `--purge` 都删除。
6. **多进程验收。** 用一个提供工具的插件和一个调用工具的测试插件，经回环测试 Adapter 覆盖三种策略、确认的允许/拒绝/超时/他人点击、无 Actor、目标不支持按钮和策略的运行时修改；会话存储覆盖原子追加、分页、跨插件隔离与 Core 重启。
7. **工具的副作用类别（D-157）。** 协议追加 ToolEffect 与 ToolSpec、ToolInfo、ToolConfirmRequest 的 `effect`，`contracts/wirespec_test.go` 固定编号（WIRE-FORMAT §3.3、§3.5）。`core.tool.list` 返回声明类别；管理目录按 CONTRACTS §13.1 合并同名声明与最近的确认记录。SDK 的 service builder 在 `Tool` 后、`Handle` 前接受 `Effect(tool.Effect)`；调用顺序或类别非法时 Collect 报错，原始 REGISTER 的非法类别由 Core 以 INVALID_ARGUMENT 拒绝。官方 Agent 按 MCP 工具注解换算类别（没有 annotations 为 unspecified；readOnlyHint 为 true 为 read_only；否则 destructiveHint 为 false 为 write，为 true 或缺省为 destructive），经 `tool.Confirm(ctx, operation, summary, effect)` 上报。管理 API 保留单项修改，批量项严格校验并在一个事务内保存或删除覆盖（ADMIN-API §18）。类别不参与策略判断。

8. **按配置文件的工具策略（D-168，T-11.6）。** `profile.Service.ToolPolicy(id, pluginID, name string, operation bool) (policy pb.ToolPolicy, enabled bool, err error)` 从本次快照读取工具或确认操作的覆盖及插件开关，与 ADMIN-API §23 的 Tools/PutTools 共用条目；无覆盖返回 UNSPECIFIED，未设插件开关为启用。`tool.Catalog.SetProfile(resolve func(ctx context.Context, target *pb.Target) (profiles *profile.Service, profileID string, err error))` 由 `internal/app` 在注册服务前注入，通过现有 targetProfile/Resolve 选择配置文件。每次 list/invoke/confirm 从 Session 恢复的可信父附件取默认 Target，建立一次操作快照；无 Target 或无父 list 使用 `default`，list 全部条目复用该快照。插件关闭时为 deny，其余按配置文件覆盖、全局覆盖、工具声明缺省策略查找；确认操作无覆盖时为 confirm，工具与操作同名时配置文件覆盖分别匹配。管理更新从下一次操作生效，在途确认保留已接纳策略；全局 Effective/Policies/SetPolicy 保持原义，未注入 SetProfile 时使用全局策略入口。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `internal/core/tool/` | 工具目录、调用策略、调用路径与确认记录 |
| `internal/core/conversation/` | 会话存储与管理服务层 |
| `internal/core/state` | 策略、确认记录与会话条目的表 |
| `sdk/` | 工具目录、调用、确认与会话存储的客户端 |
| `contracts/` | Service 描述的 `tool` 字段与 WIRE §3.5 服务的载荷，按只追加规则编号 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-55 | 三种策略、确认的全部分支、无 Actor、不支持按钮、运行时修改策略；管理页由 T-09 完成 |
| V-56 | 原子追加、分页、跨插件隔离、Core 重启与 purge；管理页由 T-09 完成 |
| V-143 | 步骤 8 中按配置文件的策略覆盖、关闭插件的工具与 `core.tool.list` 的生效策略 |

#### 失败停止条件

- 未经确认执行 confirm 策略的工具、接受非同一 Actor 的点击、后台调用通过确认，阻止本包合入。
- 工具调用绕过 Service 的访问规则或 required_permission，先修复 Broker 接入。
- 会话条目丢失、重排或被其他插件读取，停止会话存储的交付。

### T-13：开发运行与库模式

**目标：** 开发者用 `alonabot init` 创建组件工程，用 `alonabot run --reload` 在工程目录中运行：开发中的组件以独立进程运行，改动源码后自动重新构建、只重启变化的组件（CONTRACTS §15.4，D-77）。库模式只用于把 Core 嵌入其他 Go 程序和在同一进程中调试（CONTRACTS §15.5）。

#### 前置输入

- T-02 的 Session、启动流程与日志基线；T-03 的 Adapter 进程与 `sdk/adapter`；T-04 的注册与生成器；T-08 的配置准备与 `gormdb`；T-07.6 的 `core_config.json`；T-09.16 的作者 manifest。
- 读取 CONTRACTS §1、§14、§15 与 DESIGN §S。T-07 的 HTTP 入口、T-09 的管理服务合入后，开发运行与库模式直接装配它们，不另建一套。

#### 实施步骤

1. **根包与 Core 装配。** 根包 `alonabot` 提供 `Run(opts ...Option) error`、`Start(ctx, opts ...Option) (*Core, error)` 与 `Plugin`、`Adapter`、`WebUI`、`DataDir`、`Args` 等选项，数据目录缺省为工作目录下的 `data/`。`Start` 装配与 `alonabot run` 相同的 Core，在全部组件激活后返回句柄，ctx 取消时按 CONTRACTS §4.2 停止，`Wait` 返回停止结果；`Run` 是 `Start` 后等待信号再停止。
2. **开发组件的身份、来源与同进程运行。** SDK 与 `sdk/adapter` 的启动参数改为也可以由根包直接传入，外部模块不能构造这些参数。库模式的开发组件在 goroutine 中经 loopback TCP 连接 Core，走完整的握手、注册、激活与停止。生成器将工程入口目录的 `plugin.yaml` 原文嵌入 `sdk.Generated.Manifest`，并嵌入 schema 与 i18n 资源，根包通过 `manifest.ParseProject` 读取 id、role、adapter_id。Adapter 的工程入口是可导入的 Adapter 包，工程 `plugin.yaml`、`adapter_config_schema.json` 与生成登记都放在该包；生成器按 adapter 角色生成登记，`cmd/alonabot-adapter-<id>` 只保留调用 `sdk.RunAdapter` 的薄 main，库模式与独立进程共用身份（D-134）。T-13.1 在 `adapters/telegram` 新增工程 manifest，将 cmd 目录的实例声明移入该包并生成登记。Core 为开发组件创建 dev 来源的安装记录与固定代次 `dev:<组件 ID>`，组件 ID 为插件 ID 或 Adapter 包 ID；代次 ID 全局唯一、状态库结构不变，按钮 token 重启后仍有效（D-134）；未捕获的 panic 使进程退出；停止预算到期组件仍未退出时，`Run` 返回错误。
3. **开发组件的配置与数据库。** 每次启动以 `<data_dir>/plugins/<id>/config.json`（缺失时为空对象）为配置，经 `OnConfigLoad(ctx, raw)`、补值与校正及 required 校验后直接生效，不走候选与 Journal；WebUI 或 CLI 修改开发组件的配置时，校验通过后写回该文件并重启该组件。开发中的 Adapter 的实例与已安装的 Adapter 一样取自 `core_config.json` 的 `adapters`，schema 取自可导入 Adapter 包的生成资源；复用 `dev:<组件 ID>`（D-134）。受管 SQLite 为 `dev.sqlite`，不复制、不回退。dev 来源的组件拒绝 update、rollback 与 uninstall。
4. **开发运行与热部署。** `alonabot run [目录...] [--reload]` 把给出的目录（缺省为当前目录）作为开发组件；Adapter 目录为可导入的工程包，manifest 与实例声明从该包读取，构建 cmd 薄入口，代次沿用 `dev:<组件 ID>`（D-134）：按 `build` 构建到 `<data_dir>/dev/bin/`，插件、Adapter 与 Agent 以独立进程运行，WebUI 以目录中的 `assets` 代替已安装的 WebUI。`--reload` 时每 `DEV_RELOAD_POLL` 用标准库轮询组件目录中的源码、嵌入资源、`go.mod`、`go.sum`、`plugin.yaml`、`config_schema.json`、`adapter_config_schema.json`、`i18n/` 与组件的配置：源码变化时重新构建，成功后只重启变化的组件，失败时输出编译器信息、旧组件继续运行；只有配置变化时不重新构建，只重启该组件或变化的 Adapter 实例；`core_config.json` 的其他变化交给 T-07.6 的重新读取。
5. **示例与文档。** `examples/devbot/` 是一个用 `alonabot init` 生成的插件工程与回环测试 Adapter，覆盖开发运行、热部署、编译失败与同进程调试。docs/development.md 补充 `init`、`run --reload` 与库模式的用法。
6. **根包只保留库模式接口（D-148）。** 行为与协议不变。Core 装配、候选与进程管理从根包移到 `internal/app`，根包的 `Run`、`Start`、`*Core`、`Wait` 与选项改为对它的薄包装。只供开发 CLI 使用的 `Executable`、`Stdio`、`Core.Restart`、`Core.RestartInstance` 与 `Core.ReloadSettings` 移入 `internal/app`，`cmd/alonabot` 直接使用。根目录下测试装配与内部服务的端到端测试移到 `internal/app` 或行为所属的包，根包只留验证库模式公开接口的测试；ARCHITECTURE 同步模块表与依赖边。

7. **CLI 参数从简（D-149）。** 协议与管理 API 不变，只改 `cmd/alonabot` 与说明。`build`、`check`、`test`、`pack` 改为 `命令 [目录]`，删除 `--dir`；`init` 的 `--id`、`--module` 与 `--adapter-id` 按 D-149 取缺省值；`pack` 在 manifest 的 `source` 为 GitHub 仓库时推出 `--asset-url`；`install` 的位置参数按「已存在的普通文件」与「组件 ID」区分，组件 ID 从程序目录的 `packages/` 取官方包，不给参数时在当前目录或工程的 `bin/` 中寻找唯一的安装包。推不出唯一值时报错并列出候选。docs/development.md、docs/plugin-guide.md、README 与 CLI 帮助中的写法同步。 `pack` 增加可重复的 `--target <os>/<arch>`：逐个平台构建归档，写出一份列出全部资产的 `release.json`，给出 `--key` 时只签这一份；不给时为本机平台（D-169）。
8. **CLI 开箱即用（D-181，T-13.7）。** 协议与 Core 的分发路径不变。终端 Adapter 用 `sdk/adapter` 实现，由 `cmd/alonabot` 以同进程组件登记并在本次运行的设置中加入实例与 Core Admin，不写回 `core_config.json`；`--console` 三态、stderr 缺省级别、开发目录缺省重载与发布版 `init` 的 `go mod tidy` 按 D-181。docs/development.md 与 `examples/devbot/README.md` 的上手步骤同步。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| 根包 | `Run`、`Start`、`*Core` 与选项（库模式接口，D-148） |
| `internal/app/` | Core 装配、候选与进程管理、开发组件运行与重启（D-148） |
| `internal/core/devrun/` | dev 来源记录、开发组件的构建与进程、同进程组件、源码与配置轮询、重启 |
| `cmd/alonabot/` | `run` 的组件目录与 `--reload`；参数从简（步骤 7） |
| `sdk/`、`sdk/adapter/`、`sdk/generate/` | 可直接传入的启动参数；生成代码嵌入组件身份与资源 |
| `adapters/telegram/` 及其 cmd 薄入口 | 工程 manifest、实例声明与生成登记归可导入包，库模式与独立进程共享身份（D-134） |
| `examples/devbot/` | 开发运行示例 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-64 | 开发组件的完整生命周期（独立进程与同进程）、Adapter 包资源生成登记与共用身份、多个组件的固定代次 `dev:<组件 ID>` 与重启后按钮有效、dev 来源的限制、`config.json` 的读取与写回、panic 与停止超时、与已安装包混合运行 |
| V-65 | 改动源码后只重启变化的组件、编译失败旧组件继续、只改配置时不重新构建、已安装组件与 Adapter 连接不中断 |
| V-64、V-65（步骤 6） | 拆分根包后，库模式、开发运行与热部署的集成测试在新位置照常通过；根包只导出 CONTRACTS §15.5 的接口 |
| V-121（步骤 7） | 工程目录的位置参数与缺省、`init` 与 `pack` 的缺省值、`install` 按文件、组件 ID 与当前目录寻找安装包，推不出唯一值时报错并列出候选 |
| V-149（步骤 8） | 生成的工程不带参数运行并在终端得到回复、`--console` 三态、按钮点击、缺省重载与关闭、stderr 缺省级别与日志文件、`init` 的发布版与开发版路径 |

#### 失败停止条件

- 开发组件的握手、注册、权限或发送语义与已安装组件不同，阻止合入；不能为开发运行另建一条协议路径。
- `sdk`、`adapters/`、`plugins/` 或 `internal/core` import 根包，按 ARCHITECTURE 修正。
- 热部署时 Core 重启、已安装组件或 Adapter 连接中断，或编译失败导致正在运行的组件停止，先修复。

### T-16：插件测试套件

**目标：** 给插件作者一个类似 NoneBug 的集成测试工具：在 `go test` 中用真实 Core 与测试 Adapter 运行被测插件，注入事件并断言这次分发产生的发送（CONTRACTS §17，D-47）。

#### 前置输入

- T-13 的 `alonabot.Start`、同进程开发组件与 dev 配置/数据库；T-05 的 HandlerOutcome、多轮等待；T-03 的按钮；T-07 的 schedule。
- 读取 CONTRACTS §15.5、§17。套件是 SDK 公共 API，导出的名称与语义要稳定。

#### 实施步骤

1. **套件骨架与测试 Adapter。** 新建 `plugintest/`：`Start(t, opts...)` 在 `t.TempDir()` 上用 `alonabot.Start` 装配 Core，登记被测插件与套件自带的测试 Adapter（adapter_id 缺省 `plugintest`），在 `core_config.json` 中写入测试实例并在 `t.Cleanup` 中停止。提供 `Plugin`、`Config`、`Admin`、`Capabilities`、`Adapter`（指定 adapter_id 并声明其权限与扩展事件）与收紧 `allowed_adapter_instances` 的选项，以及私聊、群聊、带平台权限的用户的构造函数；可以注入已声明的扩展事件。测试 Adapter 的发送与编辑一律 ACCEPTED，消息 ID 递增。
2. **消息注入与分发结束通知。** Core 在一个事件的全部分发链结束，或其中的 Handler 进入多轮等待时，经 `internal/inproc` 通知套件；注入消息的方法据此返回这期间的全部 Send、Edit 与 Handler 错误。通知只在库模式装配中注册，独立进程运行的 Core 没有这条路径。
3. **按钮、多轮等待、schedule、后台发送与 LLM 替身。** 对结果中的按钮注入点击；多轮等待中的下一条消息按普通注入处理；立即触发一次 schedule 并在其调用结束后返回；后台发送按顺序读取。`plugintest.LLM(handler)` 登记只提供 `llm.generate` 的测试 Agent（D-77，T-16.3）。
4. **示例、文档与证据。** 用套件为 `examples/plugins/` 中的示例插件编写提交的集成测试，覆盖 V-73 的场景；docs/development.md 补充套件的用法。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `plugintest/` | 公开的插件测试套件与测试 Adapter |
| `internal/inproc/` | 分发结束通知 |
| `examples/plugins/` | 用套件编写的示例测试 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-73 | 套件覆盖命令、Rule、权限、Capability 子集、按钮、多轮等待、schedule、后台发送、配置与数据库；`-count=2 -race` 通过；Start 失败与停止后无残留 |

WebUI 中可交互的沙盒不在 v0.1 范围内（D-47）。

#### 失败停止条件

- 套件为测试跳过权限、Target 准入、Capability 预检或 Hook，或另建一条派发路径，阻止合入。
- 注入的返回依赖 sleep 或轮询时间，先改为 Core 的分发结束通知。
- 生产代码或 `internal/core` import `plugintest`，按 ARCHITECTURE 修正。
