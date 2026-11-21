# AlonaBot 架构地图

负责范围：模块职责、允许的 import 方向、状态归属。行为细节以 `spec/` 为准（见 [spec/README.md](spec/README.md)）。
更新时机：新增顶层包、改变包之间的依赖或状态归属时，同步修改本文件和 [docs/development.md](docs/development.md) 的检查表。

实现目录与职责如下；开发入口见[源码开发](docs/development.md)，精确行为见[技术规范](spec/README.md)。

## 模块与职责

Core 与 SDK 共用 `github.com/AlonaBot/AlonaBot` module。官方组件使用独立仓库与 module；`tests/` 隔离跨组件测试依赖，`components.lock.json` 固定整合发行源码。

| 模块（路径） | 负责 | 不负责 |
| --- | --- | --- |
| `contracts/` | `.proto` 源（`proto/`）与生成类型（`pb/alonabotv1`）、manifest/release/消息结构校验、调度触发计算（`schedule`）、通用 Markdown 解析（`markdown`，Goldmark）、兼容 fixture；生成工具 `contracts/internal/protocgen`（只由 go:generate 运行） | 任何运行时 I/O、会话状态、平台/GORM 依赖 |
| `internal/wire/` | 帧编解码（4 MiB 上限）、TCP/stdio 传输、Session 调用表、准入预算、优先通道、取消 | 业务路由、权限、插件生命周期决策 |
| `internal/tune/` | 调优参数表：名称、缺省值、单位、约束与 `--args` 解析（CONTRACTS §16） | 参数的业务含义（由使用方实现） |
| `internal/oslock/` | 跨进程 OS 独占文件锁（flock / LockFileEx），供 core.lock 与 runtime.lock 共用 | 锁的业务语义（何时加锁由调用方决定） |
| `internal/inproc/` | 库模式同进程组件的启动参数与运行入口（`internal/app` 构造参数；SDK 读取参数，并在 init 中登记入口） | 启动流程本身（归 sdk 与 `internal/app`） |
| `internal/pluginkit/` | 开发 CLI 的打包、release.json 生成与签名（plugin build/pack） | 安装与预检（归 `internal/core/packages`） |
| `sdk/`（含 `sdk/message`、`sdk/sqlitedb`、`sdk/gormdb`、`sdk/llm`、`sdk/agent` 等子包） | 组件侧 API：根包只放插件对象、Context/Send、注册 builder、权限、生命周期 Hook 与运行入口，动态任务、Service、Event Bus、按钮、媒体、会话历史、工具、平台动作、选项来源与命令目录按类别放在 `sdk/task`、`sdk/service`、`sdk/bus`、`sdk/button`、`sdk/media`、`sdk/conversation`、`sdk/tool`、`sdk/action`、`sdk/options`、`sdk/command`（D-146），根包与这些子包共用的注册状态在 `sdk/internal/registry`；runtime.lock；公开消息与事件 API 使用 `sdk/message` 的消息、事件、`SendResult` 与媒体类型，会话字符串、消息句柄与用户字符串是 string，由 `sdk/message` 的 `SID`/`ParseSID`、`Ref`/`ParseRef`、`UID`/`ParseUID` 拼接与拆分（D-144），错误使用 `sdk/errcode`；只供 CLI 注入的登记代码导入的资源登记入口 `sdk/genimpl`（D-145）；受管 SQLite 的 `database/sql` 与 GORM 两个按需入口（D-49）；`sdk/llm` 提供文本生成类型与 Ask，`sdk/agent` 提供以 `@agent` 为目标的 Generate | Core 状态库、管理凭证、其他插件 |
| `internal/core/` | 状态库与 core.lock、进程、RPC 编排、路由、权限、媒体、服务、调度、HTTP、配置、管理 | 平台协议细节（交 adapters）、插件进程内逻辑 |
| `internal/app/` | Core 装配、候选与进程管理、同进程运行开发组件，以及只供开发 CLI 使用的独立进程开发组件、单组件重启与重新读取设置（D-148）；放在 `internal/core` 之外，因为它要 import `sdk` | 库模式的公开接口（归根包）、业务逻辑 |
| `internal/generate/` | 构建期与测试期的扫描器：扫描入口包、检查资源与注册入口，生成经 overlay 编入的资源登记与正式入口；CLI 构建与 plugintest 的现场扫描共用（D-145） | 运行时资源读取（归 SDK）、写作者目录 |
| `internal/generated/` | 组件资源登记表：按入口包导入路径登记、按对象绑定与查找 manifest 原文、配置与实例 Schema、Service Schema、i18n 与 Handler 参数名，并记录插件对象的入口包（D-145） | 扫描与生成（归 `internal/generate`） |
| 根包 `alonabot`（模块根目录的 Go 文件） | 库模式入口（D-32）：只有 CONTRACTS §15.5 的库模式接口，每个都是对 `internal/app` 的薄包装（D-148） | Core 装配、开发 CLI 专用能力（归 `internal/app`）、业务逻辑 |
| [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) | 按需安装的官方插件，提供 `/help` 与 `/whoami`，只经 SDK 与 Core 通信 | 命令目录与权限判断（归 Core） |
| [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) | 官方 Agent 组件（`role: agent`，D-27、D-59～D-61）：对话循环、Provider、工具与 MCP、记忆、技能、定时任务与 `llm.generate`；只经 SDK 与 Core 通信，可被社区 Agent 替换 | Core 的工具策略、确认与会话存储（归 Core） |
| [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) | 官方 Telegram Adapter 程序：接收、标准化、发送计划、能力与平台映射（D-26） | Core 策略、命令路由、游标持久化与按钮 token（归 Core） |
| [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) | 官方 OneBot 11 Adapter 程序：正向与反向（D-94，经 Core 统一端口）WebSocket 连接、接收、标准化、发送计划、能力与平台映射，以及本地模拟服务（D-46） | Core 策略、命令路由与按钮 token（归 Core） |
| `plugintest/` | 插件测试套件（D-47）：在 `go test` 中用 `alonabot.Start` 启动真实 Core、测试 Adapter 与被测插件，注入事件并取得发送 | Core 的检查与派发（照常由 Core 执行，不另建路径） |
| `cmd/alonabot/` | Core 启停、开发者与管理 CLI 的装配入口 | 业务逻辑（只组装） |
| `examples/plugins/` | 随 SDK 编译的示例与集成样例 | 生产代码依赖 |
| [alonabot-webui](https://github.com/AlonaBot/alonabot-webui) | 官方 WebUI 组件（`role: webui`，D-62、D-162）：Vite + TypeScript + React 19 管理界面及 Go 后台进程；进程与同一份 `dist` 按平台同包发行 | Core 管理数据的所有权与写入规则（经 Admin API 操作，CONTRACTS §7.5） |
| `Dockerfile`、`docker-compose.yml`、`deploy/` | 官方容器镜像、compose 整合包入口与 `install.sh`（CONTRACTS §15.3、§15.6，D-64） | Core 行为（镜像与脚本只安装发行归档的内容） |
| `spec/` | 现行行为契约、协议与技术设计 | — |

生命周期所有者为每个 Core 唯一的 `supervisor.Waker`，由 `internal/app` 通过 `WakeOptions{Store, Global, QueueRequests, QueueBytes, Start}` 创建并注入 `supervisor.Spec.Waker`。`EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error)` 合并启动、管理三种等待预算并返回活动释放函数；`Ready(running *Running)` 在 ACTIVATED、摘要核对及 Router、Bus、HTTP、Scheduler 全部绑定后开放交付。`AcquireActivity(pluginID string, session *wire.Session) (func(), error)` 只取得当前 Session 引用，幂等释放始终归属原 Session。`Running.StopIdle(t0 time.Time, lastActivityAt time.Time) host.StopResult` 在逐插件锁与 Waker 活动门禁内复核管理状态、Session、Ready、零引用与实际活动结束时刻，变化时返回 Cancelled；进入 QUIESCING 后冻结新派发，新请求等待旧树退出、SUSPENDED 与下一次 Ready。wire 的 `Options.RequestActivity func(*wire.Inbound) (func(), *pb.ErrorInfo)` 经 host.ActivateOptions、supervisor.Prepare 同步接入 Waker，与 Session 在途登记共用临界区；正常 WaitIdle 覆盖整个准入区间。

Router 与 Broker 的 `Options.EnsureActive`、HTTP 的 `Server.SetWake(snapshot, ensure)` 共用同一唤醒入口；匹配、过滤及准入先于唤醒，完整绑定后按原绝对期限复核代次和路由。`router.WithActivity(p, acquire)` 包装实际调用与 EVENT 写出并组合原 `Request.Late`；Hook、已运行 schedule 同样经过该包装。Bus 的 `Attach(p router.Plugin, caller ...router.Caller)` 保留单参数入口，生产装配传订阅者自身 Caller，通过 `Options.AcquireActivity` 持有排队工作。媒体通过 `Store.SetActivity`、延续与等待通过 `continuation.Options.AcquireActivity` 接入；主动出站服务在生产 Handler 中持有至实际业务返回。HTTP 引用在反向代理返回时释放；HTTP/Bus 仍在执行的 SDK 回调沿用正常停止截止。

`Router.Options.MessageConfig` 按会话配置文件读取 Active 配置，并合并 `Resolve` 与有效 `Overrides`。Waker 通过 `Activity(pluginID string) ActivityStatus` 提供 Session、Ready、References、LastActivityAt、ColdStarts 与 ColdStartSequence 的独立快照；Core 空闲控制器消费后经 `AcknowledgeColdStarts(pluginID string, through uint64)` 回收历史前缀，常驻、停用和休眠实例同样消费。控制器按窗口清理记录并保持固定冷却截止。Scheduler 到期唤醒保留原 JOB_TIMEOUT，验证入口见[源码开发](docs/development.md#验证改动)。

`scheduler.Options.EnsureActive func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)` 由 app 注入同一 `assembly.ensureActive`；`func (s *Scheduler) Suspend(pluginID string)` 保留正常 IDLE 计划。静态绝对期限为原 scheduled_at 加 JOB_TIMEOUT，动态从取得并发名额、开始交付或唤醒等待时计算；等待保留运行槽与动态并发名额，按完整 HandlerRequest 编码字节占用既有唤醒预算。相同代次与路由的 resume Attach 保留计划、已接纳项、已消费 at 的待交付所有权、起点和暂停状态；清理未注册的过期 at 并更新 Expired/Unregistered、归还配额。Supervisor 独占独立进程的调度撤销与 Released，旧树退出只释放旧 Session 的实际槽；app 在 Core.operation 内撤销唤醒失败的计划并保留完整错误。

`internal/core/idle/` 提供：`Resolve(s settings.Sleep, m *manifest.Manifest, independent bool, routes []*pb.RouteDescriptor) Policy` 只读判定准入、原因与全路由 warnings；`New(values tune.Values) *Controller` 与 `CanStop(pluginID string, policy Policy, runtime state.RuntimeStatus, activity supervisor.ActivityStatus, now time.Time) bool` 判定实际空闲、最短运行及固定冷却，由 `internal/app` 的生命周期循环每秒检查并调用 StopIdle。管理设置使用 `File.WithPublisher(publish func(change func() error) error) *File` 共享文档、快照与锁，Update/Reload 交给 Core.operation 串行发布；有效 sleep 变化后，在循环及文件锁、插件锁外等待原代次恢复并聚合错误。

S-06.1 的 `manage.Options.SleepPolicy` 由 app 注入只读策略：复用 `idle.Resolve` 与已提交的身份匹配快照，管理层沿用原依赖方向。六个休眠管理端点共享管理认证和受管设置发布；CLI 经同一 HTTP 管理服务查询/写入。查询不启动组件，缺有效快照时返回实际 412；写入沿既有锁外恢复并完整返回持久化和唤醒错误。

`internal/core/profile` 负责配置文件的存储编排、实例与会话解析、Active schema 覆盖校验及管理方法。`profile.New(ctx, profile.Options{Store, Settings, Schema})` 建立一次操作的快照；管理写入后刷新状态快照，下一次独立操作重新构造。`Resolve(instance, session) Profile` 与 `Overrides(pluginID) map[string]map[string]json.RawMessage` 在内存中读取快照。HTTP 管理调用方串行化写入，并从读取 Active schema 到提交持有相应插件的管理锁。`internal/core/manage` 提供 ADMIN-API §23 及经 `RegisterProfiles(mux *router.Mux)` 登记的四个 `core.profile.*` 固定服务，`cmd/alonabot` 通过 `alonabot profile` 调用管理 API。固定服务 list/current 直接读快照，assign 跳过 schema，set 只读取调用插件的 schema；写入沿 catalog → instances → 包锁（仅 set）尝试互斥，占用时立即返回 FAILED_PRECONDITION。

`internal/app` 向 Router 注入 `Options.Profile func(context.Context, string, *pb.Target) (string, bool, error)` 与既有 `Options.MessageConfig func(context.Context, string, *pb.Event) (json.RawMessage, error)`；`profile.WithView` / `ReadView` 使一次准入与消息过滤共用配置文件选择、开关、Active 与有效覆盖。PreEvent 在媒体准备前检查开关。`router.Invocation.ProfileID` 与 `Root` 随 Handler、Service、Hook、定时任务与延续附件传递；`Begin() func()` 保留和释放根引用，`AfterFinished(func())` 在整条根调用及延续释放后锁外执行，根已结束时立即执行新登记回调，用于 default Pending 的延后 reload。

`profile.Service.EffectivePlatform(id string) (prefixes []string, reply settings.Reply, err error)` 按键合并当前 default 与覆盖，保留显式零值。`internal/app` 向 Router 注入 `Options.Prefixes func(context.Context, *pb.Target) ([]string, error)`，派发使用完整 Target，菜单使用实例身份与空会话选择实例缺省配置文件；同时经 `message.Service.SetReply(resolve func(context.Context, *pb.Target) (settings.Reply, error))` 注入回复读取。message 在 Hook 前改写回复并复检；可信消息根通过 `Invocation.Root` 在手动发送、自动回复、Service 派生与延续间共享首条标记，根结束时清理。Edit 修饰按 Caller Session、完整 Target 与选定物理消息 ID 的已确认片段关联；`internal/app` 在 Session 退出时调用 `message.Service.Release(router.Caller)`，配置读取中的准备及迟到回执随之失效。

`manage.Options.NotifyProfileOverrides func(context.Context, string) error` 由 `internal/app` 装配，读取新快照并向当前在线 Session 下发有效覆盖全集，失败关闭原 Session 并返回实际错误；离线插件在下次 Bootstrap 接收。候选 Bootstrap 按候选 schema 筛选覆盖集合。`contracts/configschema.Schema.ApplyProfile(raw json.RawMessage, overrides map[string]json.RawMessage) (json.RawMessage, error)` 提供字段路径替换，`sdk/config.Values.WithOverrides(overrides map[string]json.RawMessage) (Values, error)` 建立独立只读视图并保留子路径。SDK 原子发布集合，每次 Context 构造冻结视图，`c.Config()` 返回整个插件的调用配置，`p.Config()` 与 `s.Config()` 保持基础配置。

配置回复示例（D-160）沿用现有依赖方向：`contracts/configschema` 检查示例、字段引用与翻译的对应关系，并通过 `Schema.ExampleImages()` 返回去重排序的图片路径；`contracts/i18n` 检查通用词典中示例覆盖的结构与文案类型。`internal/pluginkit` 校验两种配置 schema 引用的图片并收集到资源快照，`internal/core/packages` 在安装和更新时复核。`internal/core/manage` 按 Active 代次的已提交 Journal 选包，只通过认证后的 GET/HEAD 提供包中 `config_schema.json` 引用的图片。

`sdk/kv` 提供基于 Session 的键值存储；`sdk.PluginDataDir(ctx)` 优先按 `ctx` 选择所属 Session，同进程插件各取自己的目录，独立进程在 `ctx` 未绑定时使用默认 Session。SQLite 驱动随 `sdk/sqlitedb`、`sdk/gormdb` 或 `sdk/kv` 按需链接，SDK 根包不引入驱动。

`internal/core/diagnostics` 负责诊断数据收集、脱敏与 zip 写入（CONTRACTS §14.7）。`internal/core/manage` 的诊断端点与 `cmd/alonabot` 的 `diagnose` 共用 `Write(ctx context.Context, w io.Writer, src Source, opts Options) error`。`Running` 复用当前 `state.Store`、运行状态与 schema 回调；`Directory` 对已有 `core.lock` 的只读句柄用 `x/sys` 取得与 Core 相同的 flock / LockFileEx 独占锁，用现有 GORM/SQLite 只读查询主库及已有 WAL 的临时副本。诊断包复用 `state` 的记录类型、`contracts/configschema` 的迁移与秘密路径及 `internal/core/secret` 的引用格式；认证、HTTP 响应与 CLI 输出替换由各入口负责。

组件 changelog（D-166）由 `internal/pluginkit` 在 `check`、`build`、`pack` 的资源扫描中校验，`pack` 保留文件原文；`internal/core/packages` 在安装包解压后复核。`internal/core/manage` 通过 `GET /api/v1/packages/{id}/changelog` 读取当前安装包，通过 `GET /api/v1/core/changelog` 读取由 `changelog/unreleased/en.md` 生成并编入 Core 的 `changelog_core.go`。

`sdk/service.Builder.Effect` 随 REGISTER 上报工具类别；`sdk/tool.List` 通过 `Info.Effect` 读回声明类别，Agent 将 MCP 注解换算后的类别交给 `sdk/tool.Confirm`。`internal/core/tool` 合并声明、确认记录与策略覆盖；管理层批量写入时先持有 catalog 锁，再按组件 ID 顺序取锁，与安装替换使用同一锁层级。`Catalog.SetPolicy(ctx, rows ...state.ToolPolicy)` 调用 `Store.SetToolPolicy(ctx, rows ...ToolPolicy)`，由状态库在同一事务内保存或删除覆盖。

`internal/app` 在注册工具服务前调用 `Catalog.SetProfile(resolve func(ctx context.Context, target *pb.Target) (profiles *profile.Service, profileID string, err error))`，回调复用 targetProfile/Resolve，返回本次操作的配置文件快照与 ID。list/invoke/confirm 从 Session 恢复的可信父附件读取 DefaultTarget，每次操作解析一次；无 Target 或无父 list 使用 `default`，list 全部条目复用该快照。`profile.Service.ToolPolicy(id, pluginID, name string, operation bool) (policy pb.ToolPolicy, enabled bool, err error)` 与 Tools/PutTools 共用快照条目，分别匹配工具和操作；无覆盖返回 UNSPECIFIED，未设插件开关为启用。插件关闭时为 deny，其余按配置文件覆盖、全局覆盖与工具声明缺省策略查找，确认操作无覆盖时为 confirm。管理更新从下一次操作生效，在途确认保留已接纳策略；全局 Effective/Policies/SetPolicy 保持原义，未注入 SetProfile 时使用全局入口。工具执行继续经 Broker 继承父调用权限、Actor、Target、取消与期限，并保留结果与错误。

SDK 的文本匹配、Raw 与事件注册在 Plugin、SubPlugin、CommandGroup 上省略 route ID；Hook 在 Plugin 与 SubPlugin 上提供，稳定身份由 builder 的 `.ID` 指定。`*sdk.Context` 实现 `context.Context`，分类客户端可直接接收 `c`；`c.Reply(text, more...)` 将文本与片段作为一条消息发送。插件测试的 `Result.RequireReply` 检查 Handler 错误与单次发送的文本（D-180，T-04.20）。

`sdk/page` 属于 SDK：根包在收集 `.Page(title)` 时调用 `page.Handler()`，自动补齐页面同目录的 GET `kit.css` 与 `kit.js`，以公开静态资源模式 `sdk.AuthStatic` 提供，页面 HTML 与数据路由继续要求管理会话，优先保留作者已登记的同方法、同路径路由；插件也可手动挂载 `Handler()`。`kit.js` 负责页面端的桥请求、异步翻译、主题与语言订阅和快捷键转发；沙箱、父桥与全局搜索由 WebUI 负责（[ADMIN-API §7](spec/ADMIN-API.md#7-插件页面-postmessage-桥)）。`examples/plugins/showcase` 演示页面声明与套件使用。

[Agent pages.go](https://github.com/AlonaBot/alonabot-agent/blob/main/pages.go) 经 `sdk/page` 与普通管理 HTTP 路由提供人格、MCP、技能、记忆、用户任务及嵌入资源，接口见 [ADMIN-API §15.2](spec/ADMIN-API.md#152-agent-插件页面与数据路由d-170d-182)。人格表 `persona` 与会话目录 `conversation_record` 纳入受管库初始迁移 ID:1，记忆保留 ID:2；正文归 Core，Agent 在追加成功后记录完整会话键，目录清理与追加/reset 复用会话锁。普通与定时轮次读取当前人格提示；技能采用轮次快照；MCP gate 协调状态、重连与停止锁存，Session 结束通知按身份隔离并保留最近工具目录。动态任务通过既有 SDK List/Cancel 管理。

`sdk.ProfileOverrides(ctx context.Context) (map[string]map[string]json.RawMessage, error)` 从 ctx 所属 Session 的 Plugin 读取当前有效覆盖集合并返回深拷贝，HTTP 与生命周期共用该所有者，同进程多个插件各自隔离；独立进程沿用默认 Session。数据复用 Bootstrap/通知的原子发布，Agent 将其与基础配置、`core.profile.list` 合并计算人格引用（D-182）。

`sdk.AdapterClient` 由 `Context.Adapter()` 绑定可信平台身份，复用 `Context.CallAction`、`core.action` 与 `adapter.action`。`ActionRequest.use_default_target` 仅供绑定客户端按已声明 scope 选取父调用默认会话或账号；声明、Schema、实例许可、出站 Hook 禁用与生命周期仍由现有链路执行。`sdk/action.Caller` 同时接受绑定客户端、Context 和后台 Client；Instance 的必需方法保持原有集合。

## 依赖方向

`A -> B` 表示 **A 的 Go 源码 import B**（静态依赖，含测试代码），不表示运行时调用或数据流。「模块」指上表每一行的路径前缀；同一模块内部的 import（如 `contracts/release -> contracts/manifest`、`sdk -> sdk/message`）不受限制。跨模块的边只允许下表列出的；标准库和已批准的第三方依赖不受此表限制。

| 允许的边 | 理由 |
| --- | --- |
| `internal/wire -> contracts` | 帧内承载生成的 Envelope 类型 |
| `internal/wire -> internal/tune`、`sdk/... -> internal/tune`、`internal/core/... -> internal/tune` | 读取调优参数的生效值 |
| `sdk/... -> contracts`、`sdk/... -> internal/wire`、`sdk/... -> internal/oslock` | 插件侧复用同一协议实现与 runtime.lock（Go 允许本 module 的 `sdk` import `internal`） |
| `internal/core/... -> contracts`、`internal/core/... -> internal/wire`、`internal/core/... -> internal/oslock` | Core 侧协议、core.lock 与停机复制时临时持有 runtime.lock |
| `internal/core/manage -> internal/core/profile`、`internal/core/router -> internal/core/profile`、`internal/core/tool -> internal/core/profile`；`internal/core/profile -> internal/core/state`、`-> internal/core/settings`、`-> contracts/configschema`、`-> internal/wire` | 管理 API 与固定服务建立配置文件快照，Router 复用准入和过滤的操作视图，工具目录、调用与操作确认读取同次策略操作的覆盖及插件开关；状态库存指定与覆盖、Core 设置存 `default` 平台设置、Active schema 校验覆盖值，覆盖集合预算沿用 `wire.MaxFrame` |
| `internal/core/idle -> internal/core/settings`、`-> internal/core/supervisor`、`-> internal/core/state`、`-> contracts/manifest`、`-> contracts/pb/alonabotv1`、`-> internal/tune`；`internal/app -> internal/core/idle` | 只读策略准入、活动与运行状态快照、调优值；Core 装配控制器并持有生命周期所有权 |
| `internal/core/buttons -> internal/core/state`、`internal/core/buttons -> internal/core/router` | 持久化按钮 token，复用可信 Caller；message 与 router 通过函数注入接入 buttons |
| `internal/core/state`、`internal/core/settings`、`internal/core/journal -> internal/core/logging` | 共用 Core 的 CRITICAL 级别记录状态库、设置与提交后故障；SDK 独立定义 `sdk.LevelCritical`，两侧共用协议的 `LogLevel` |
| `internal/core/manage -> internal/core/scheduler` | 列出保留计划中的静态任务（含正常休眠），暂停与恢复复用调度器的持久状态及当前 Activated Session 准入（D-164） |
| `internal/app -> internal/core/adapter/remote` | 装配 Adapter 实例运行状态查询，登记配置失败并在实例配置变更时清除旧快照（D-164） |
| `internal/core/broker -> internal/core/metrics`、`internal/core/bus -> internal/core/metrics`、`internal/core/router -> internal/core/metrics`、`internal/app -> internal/core/metrics` | Service、Bus 与事件派发按原因计数（只在内存中，管理面在 T-09.14 读取） |
| `sdk -> sdk/internal/convert`、`sdk/adapter -> sdk/internal/convert`；`sdk/internal/convert -> contracts`、`-> sdk/message`、`-> sdk/errcode` | 集中转换 SDK 自有类型与 pb，并组装 Adapter 的类型化声明；公开 API 使用 SDK 类型 |
| `sdk -> sdk/errcode`、`sdk/message -> sdk/errcode`、`sdk/adapter -> sdk/errcode` | 共用 SDK 错误码与错误类型；`sdk/errcode` 只依赖标准库 |
| `sdk/command -> sdk`、`-> sdk/errcode`、`-> sdk/internal/registry`、`-> contracts/pb`、`-> internal/wire` | `command.List` 复用 SDK 的父调用关联与固定服务协议，对外返回 `command.Entry` |
| `sdk/kv -> sdk/sqlitedb`、`sdk/kv -> sdk/errcode` | 复用受管 SQLite 的 Session 连接与 SDK 错误码 |
| `sdk/message -> contracts/validate` | 会话、消息句柄与用户字符串复用 Adapter ID 的校验规则 |
| `sdk -> sdk/internal/messagecodec`、`sdk/message -> sdk/internal/messagecodec`、`sdk/internal/convert -> sdk/internal/messagecodec` | 跨包传递按钮与片段的编码错误，在发送前报告；保留引用目标的省略信息 |
| `alonabot-adapter-{telegram,onebot11}/... -> sdk/...`、`-> contracts` | Adapter 程序与插件示例一样，只经公共 SDK（`sdk/adapter`）与 Core 通信 |
| `cmd/alonabot -> internal/core/...`、`-> contracts`、`-> internal/tune` | 装配；`run` 经 `internal/app` 复用库模式的 Core 装配；`--args` 与 `tune list` 使用调优参数表 |
| `internal/app -> internal/core/...`、`-> sdk/...`、`-> contracts`、`-> internal/wire` | 在同一进程中装配 Core 与开发中的组件 |
| `tests/app` overlay 测试 `-> alonabot-agent/...`（仅集成测试） | 通过官方 Agent 与本地 Provider 验证文本生成服务及选项来源 |
| `internal/app -> internal/inproc`、`sdk/... -> internal/inproc` | 库模式把同进程组件的启动参数从 `internal/app` 交给 SDK，SDK 在 init 中向它登记运行入口供 `internal/app` 调用；`internal` 保证外部模块无法构造这些参数 |
| `cmd/alonabot -> internal/generate` | `init`/`check`/`build`/`pack` 复用扫描与登记代码生成器 |
| `internal/pluginkit -> internal/generate`、`plugintest -> internal/generate`；`sdk`、`sdk/internal/registry`、`sdk/genimpl`、`internal/app`、`plugintest`、`internal/generate -> internal/generated`；本 module 的测试代码 `-> internal/generated`（D-145） | 生成器在内部包中，由 CLI 构建与 plugintest 的现场扫描共用；资源登记表只 import 标准库 |
| `根包 -> internal/app`、`cmd/alonabot -> internal/app`；`internal/app -> internal/core/...`、`-> sdk/...`、`-> contracts`、`-> internal/inproc`、`-> internal/tune`（D-148） | 根包只保留库模式接口，装配与开发 CLI 专用能力在 `internal/app`；`cmd/alonabot` 不 import 根包 |
| `cmd/alonabot -> internal/pluginkit`、`internal/pluginkit -> contracts/...`、`-> internal/generate` | 开发 CLI 的构建与打包复用生成器与契约包 |
| `examples/plugins/... -> sdk/...`、`-> contracts` | 示例只走公共 SDK |
| `tests/e2e -> alonabot-adapter-telegram/telegramtest`（仅集成测试） | Telegram 目录运行集成测试与本地故障矩阵复用官方 Adapter、Bot API 模拟服务和开发构建装配 |
| `scripts/dist -> contracts/manifest`、`-> contracts/release`、`-> internal/releasenotes`；`scripts/release-notes -> internal/releasenotes`；`internal/releasenotes -> contracts/manifest`；`scripts/release-check -> contracts/release` | 发行构建与发布复验只用契约包核对签名；其余步骤经 `os/exec` 调用 go 与 alonabot |
| `alonabot-webui -> sdk`、`-> sdk/message`、`-> sdk/sqlitedb`、`-> contracts/manifest`、`-> contracts/pluginindex`、`-> contracts/release` | 官方 WebUI 后台通过公共 SDK 观察事件、发送回执与管理会话路由，只读写自身受管 SQLite；契约包解析插件源、版本与发行清单，Core 管理数据由浏览器经 Admin API 操作 |
| `examples/devbot -> 根包`、`-> sdk/...` | 库模式示例 |
| `plugintest -> 根包`、`-> sdk/...`、`-> internal/inproc`、`-> contracts`；`examples/plugins/... -> plugintest`（仅测试代码） | 套件经根包的 `Start` 装配 Core，经 `internal/inproc` 接收分发结束通知；示例插件的测试使用套件 |
| `alonabot-plugin-help -> sdk/...`、`-> contracts`；`alonabot-agent/... -> sdk/...`、`-> contracts` | 官方插件与 Agent 组件只走公共 SDK；Agent 另用已批准的 MCP go-sdk，`alonabot-agent/provider` 另用 openai-go 与 anthropic-sdk-go（D-132） |
| `alonabot-plugin-*/...`、`alonabot-agent/...`、`examples/plugins/...`、`plugintest -> alonabot-adapter-<adapter_id>/perm`、`-> alonabot-adapter-<adapter_id>/event`、`-> alonabot-adapter-<adapter_id>/action` | 引用 Adapter 声明的权限、扩展事件与平台动作（D-69、D-71、D-92）；这三个包只含常量与数据类型，只 import `sdk` 根包，不链接 Adapter 的运行时 |

明确禁止（重点）：`internal/oslock`、`internal/tune` 与 `internal/inproc` 不 import 本 module 任何包；`contracts ->` 任何本 module 包；`internal/wire -> sdk | internal/core`；`sdk -> internal/core | adapters | plugins | agent | cmd`；`sdk` 根包 -> 分类子包（`sdk/task`、`sdk/service`、`sdk/bus`、`sdk/button`、`sdk/media`、`sdk/conversation`、`sdk/tool`、`sdk/action`、`sdk/options`、`sdk/command`，D-146）；`internal/core -> sdk | adapters | plugins | agent | cmd`；`adapters -> internal/core`；`plugins | agent -> internal/core | adapters（上表的常量包除外）| cmd`；`plugins`、`agent` 之间不互相 import；`internal/core | internal/app | sdk | adapters | plugins | agent -> 根包`；`internal/core | sdk | adapters | 根包 -> plugintest`，以及非测试代码 `-> plugintest`；任何包 `-> examples`。依赖图不允许环。

## 状态归属

| 状态 | 唯一所有者 |
| --- | --- |
| `<data_dir>/core.lock`、`state.sqlite`（安装、组件角色与来源、代次配置快照与校正记录、Journal、策略、secret 引用、工具策略、Agent 会话、按钮 token 映射 `button_token`、配置文件与指定及覆盖） | `internal/core` |
| WebUI 组件、组件图标与配置示例图片（不可变包目录） | `internal/core` 只读提供；内容归组件发布者 |
| 官方 WebUI 的主题、语言与侧栏偏好（按源保存于 `localStorage`） | `alonabot-webui` 的浏览器端 |
| `<plugins>/<id>/runtime.lock`、插件受管 DB 连接 | 插件进程内的 `sdk` |
| 插件业务数据（受管 SQLite 候选文件） | 插件；Core 只负责候选复制与准入 |
| `_alonabot_kv` | `sdk/sqlitedb` 初始化表结构，插件经 `sdk/kv` 管理键值；表随受管 SQLite 处理 |
| `<data_dir>/plugin_data/<id>/` | `internal/app` 在启动装配时创建并下发路径（离线 `test` 由 CLI 准备临时目录），`internal/core/manage` 清理；文件由插件管理 |
| 插件自管的外部数据（外部数据库、框架外文件） | 插件；Core 不创建、不校验、不备份（D-48） |

## 文件组织

模块内部按职责分文件，类型、方法和专属常量就近；独立共享数据契约按主题组织，实现状态仍归自己的所有者。不要求统一 DTO/常量文件或一个类型一个文件。文件内容、注释和拆分标准以 [编码规范](docs/coding-standards.md) 为准；这不改变上面的依赖表。

## 检查

依赖方向由代码审查对照上表把关；import 环由 Go 编译器拒绝。Core 可执行文件与 SDK 根包能链接的第三方模块由审查时用 `go list -deps` 对照 [docs/development.md](docs/development.md) 的允许清单核对（D-52）。
