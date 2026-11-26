# AlonaBot 设计导读

[English](design.md)

本文面向第一次阅读 AlonaBot 源码的开发者，按「组件 → 进程 → 协议 → 消息 → 状态 → 信任」的顺序说明现行设计，以及每个设计为什么这样做。细节以文中链接的权威文档和源码为准；核心机制变化时同步修改本文。逐步动画演示见文档站的[架构页](https://alonabot.github.io/docs/internals/architecture)。

## 一句话概括

Core 是唯一的宿主和路由中心。平台接入（Adapter）、功能（插件）、AI（Agent）和管理界面（WebUI）都是可安装、可替换的组件，缺省各自运行在独立进程里，只通过 Core 和同一套协议通信。

## 组件与角色

| 角色 | 做什么 | 官方实现 | 进程 |
| --- | --- | --- | --- |
| Core | 启停组件、身份与权限、事件派发、发送消息、媒体、配置、调度、存储代次、管理 API | 本仓库 | 一个 |
| Adapter | 把一个平台的多个账号接入 Core，上报事件、执行发送与平台动作 | Telegram、OneBot 11 | 每个 Adapter 组件一个，进程内承载它的全部账号实例 |
| 插件（Plugin） | 注册命令、事件、Hook、服务、定时任务和 HTTP 路由 | echo、showcase | 每个插件一个 |
| Agent | 提供模型接入、会话和工具调用，运行方式与插件相同 | alonabot-agent | 一个 |
| WebUI | 管理界面的静态资源，可附带一个后台进程 | alonabot-webui | 只有静态资源时没有进程 |

组件包里的 `plugin.yaml` 声明角色、身份、版本与来源；Core 只经 [`contracts/manifest`](../contracts/manifest) 解析它。组件的 Go 代码只依赖公共 SDK（[`sdk/`](../sdk)）和协议包，不 import Core 内部实现，允许的依赖边见[架构地图](../ARCHITECTURE.md)。

**为什么**：平台、功能和 AI 的变化速度不同，拆成独立组件后可以单独发布、单独升级，也可以换成社区实现，Core 只需要保持协议稳定。

## 进程模型

- **一个组件一个进程**：Core 启动组件进程、完成握手，并持续监控。组件崩溃不影响 Core 和其他组件。异常退出后自动重启；`RESTART_WINDOW`（5 分钟）内异常退出 `RESTART_MAX_CRASHES`（3 次）后停用该组件。Core 不自动重试业务调用。实现见 [`internal/core/supervisor`](../internal/core/supervisor)。
- **两种传输**：缺省是 Core 持有的回环 TCP 监听（`127.0.0.1:0`），也可以用 stdio。每次启动都会创建新的 Session，并下发一次性的 32 字节 secret，组件凭它认证。
- **休眠**：开启 `sleep.mode = component_idle` 后，空闲的插件会退出，在有命令、事件、服务调用、HTTP 请求或定时任务时由唯一的 Waker 原样唤醒，同一代次复用配置和数据库。自动策略下，声明了 Hook、HTTP、Bus 订阅等常驻路由的插件不休眠；为单个插件显式设置空闲时长可以覆盖这一限制。频繁冷启动时插件会被固定常驻一段时间。准入规则见 [`internal/core/idle`](../internal/core/idle) 与[插件休眠](management-api.zh-CN.md#插件休眠)。
- **同进程模式**：Core 和组件也可以编进同一个 Go 程序（库模式），装配代码与独立进程共用 [`internal/app`](../internal/app)。[`plugintest`](../plugintest) 用它在 `go test` 中启动真实 Core。同进程时组件崩溃会带走整个程序，只适合测试、调试和嵌入。

**为什么**：插件是任意 Go 代码，进程隔离让一个插件的崩溃、内存泄漏或依赖版本冲突不扩散；按进程也能直接看出谁占了内存。代价是跨进程通信的开销，以及需要一套握手和调用协议。

## 协议与 Session

协议定义在 [`contracts/proto/alonabot/v1`](../contracts/proto/alonabot/v1)，Core 和 SDK 共用 [`internal/wire`](../internal/wire) 实现。

**分帧**：每帧是 4 字节无符号大端长度加一个 Protobuf [`Envelope`](../contracts/proto/alonabot/v1/envelope.proto)，整帧上限 4 MiB。Envelope 带帧种类、Session ID、调用方向与 call ID、期限、trace ID、目标、payload 及其契约 ID，以及错误信息。零长度、超限或无法解码的帧会直接让 Session 失败。

**握手**：组件启动后按固定顺序进入可用状态：

1. `HELLO`：组件出示身份与 secret，双方协商协议版本和能力。
2. `HELLO_ACK` 附带 `Bootstrap`：Core 下发代次、候选配置、数据库描述、日志级别、阶段与 Session 限额等私有输入。
3. `REGISTER`：组件完成配置准备与 `OnStart` 后，一次性提交全部路由，Core 原子校验整份注册集。
4. `PREPARED`：Core 确认持久化快照。
5. `ACTIVATE`：Core 在状态库提交唯一的 Active 代次后通知组件。
6. `ACTIVATED`：组件的 `OnReady` 成功，开始接收业务调用。

**调用**：调用按 `(Session, 发起方向, call ID)` 匹配，期限是绝对时间。插件在处理 Core 的请求时发起的子调用，用 `parent_call_id` 挂在这个仍在进行的请求之下，继承它的可信上下文，期限只能沿用或缩短。取消是协作式的，已经发生的副作用会保留。payload 和 error 可以同时出现，用来表示「部分成功」，调用方两者都要保留。

**为什么**：注册在启动时一次性固定，权限和启用状态由 Core 在运行时动态决定，这样 Core 能在派发前就知道每个组件声明了什么。同一套帧格式同时用于 TCP、stdio 和同进程模式，SDK 不需要区分运行方式。

## 事件、消息与路由

**身份只信 Core**：[`Event`](../contracts/proto/alonabot/v1/message.proto) 描述一次平台事件，`Actor` 是触发者，`Target` 唯一定位到某个 Adapter 实例下的会话或话题（[`identity.proto`](../contracts/proto/alonabot/v1/identity.proto)）。Adapter 上报事件后，Core 会重写事件 ID、来源和平台权限；`Actor` 只由可信的 Adapter 元数据和 Core 策略生成，插件自报的身份一律不作依据。是否是管理员只看 Core 设置。

**一条消息的处理顺序**（[`internal/core/router`](../internal/core/router)）：

1. Adapter 提交事件，Core 核对实例归属、持久化接收游标、重建可信字段。
2. Router 按条数、字节和全局预算接纳；预算不足返回 queue full，由 Adapter 记一次丢弃。
3. 按配置文件检查会话访问，计算 user/admin 权限。
4. PreEvent Hook 可以改写业务载荷或终止分发；正在等待下一句话的多轮对话先取走匹配的消息。
5. 按当前会话的前缀匹配命令，绑定参数。
6. 插件休眠时先唤醒，醒来后复核代次、路由和权限，再为接收方签发媒体令牌并调用 Handler。
7. Handler 回复时，消息服务做发送策略检查、执行 BeforeSend Hook、预留 AfterSend 容量，再调用一次 Adapter 发送。
8. AfterSend Hook 作为独立任务观察发送结果，不拖慢回复。

**路由种类**：命令、事件、Hook、服务、定时任务、HTTP 和 Bus 订阅（[`route.proto`](../contracts/proto/alonabot/v1/route.proto)）。命令层返回 `STOP` 时整条分发链终止；事件路由逐个执行，Handler 可以选择继续传播或停止。

**Hook 的边界**：PreEvent 只能改写事件载荷，不能改身份；BeforeSend 只能改写消息链，改写后重新完整预检；AfterSend 只观察。出站 Hook 内禁止再发消息，避免递归。

**跨插件协作**：

- **服务调用**：带结果的同步调用，经 [`internal/core/broker`](../internal/core/broker) 转发到目标插件当前的 Session，目标休眠时按需唤醒。`InvocationContext` 继承父事件的 `actor`、`default_target` 和 `chat_type`；Service Handler 通过 `c.User()`、`c.Session()` 和 `c.ChatType()` 读取，`c.Event()` 为空。没有真实入站事件时，`chat_type` 为 `UNSPECIFIED`。
- **Event Bus**：发布者只能用自己的命名空间，Core 接纳即返回；不持久化、不确认、不补发，离线的订阅者收不到。

**媒体**：媒体不塞进普通调用。上传按 1 MiB 分块并在提交时校验 SHA-256，得到绑定所有者 Session 和消费范围的 `MediaRef`；授权只能等同或收窄（[`media.proto`](../contracts/proto/alonabot/v1/media.proto)）。

**为什么**：所有路径都经过 Core，权限、限额和审计只需要在一处实现；Hook 的可改范围被协议限死，插件之间不能互相伪造身份或发送目标。

## 状态、代次与更新

**状态归属**：Core 的元数据（安装、代次、策略、配置文件等）在 `state.sqlite`，只有 Core 读写；每个插件的业务数据在自己的受管 SQLite 里，由插件读写；Core 负责复制候选和准入，并为管理界面提供浏览、导出和清空。完整清单见[架构地图「状态归属」](../ARCHITECTURE.md#状态归属)。

**代次（Generation）**：一次已提交的「程序 + 配置 + 数据」组合。更新、重载配置都通过候选代次完成（[`internal/core/journal`](../internal/core/journal)、[`internal/core/candidate`](../internal/core/candidate)）：

1. 停止旧进程并确认进程树退出，把受管数据库复制成候选。
2. 用候选启动新 Session，走到 `PREPARED`。
3. 在唯一的提交点，同一事务内切换 Active 代次并记录 COMMITTED，然后激活。
4. 提交前任何一步失败，丢弃候选并用旧代次恢复运行；提交后失败则保留新代次，标为 Degraded 供诊断。

保留的旧代次可以连同当时的数据一起回退。

**安装**：组件包先预检（受控解压、manifest 核对、平台与 build-info），再移入不可变的包目录。在线安装绑定来源仓库，用 GitHub Release 提供的摘要核验归档，之后的更新继续核对来源和身份（[`internal/core/packages`](../internal/core/packages)、[管理参考](management-api.zh-CN.md#组件安装)）。

**为什么**：更新失败时能回到一个完整、自洽的旧状态，而不是「程序是新的、数据迁移了一半」。Core 状态库的结构变化按顺序追加迁移，规则见[编码规范「版本与兼容性」](coding-standards.zh-CN.md#版本与兼容性)。

## 信任边界

- **插件是受信代码**：它和 Core 以同一个系统用户运行，能做这个用户能做的任何事。进程隔离只用来限制故障，不是安全沙箱，所以只安装信任的组件。
- **插件页面不受信**：插件的 HTTP 响应被强制加上沙箱 CSP，Core 会清除插件设置的 Cookie，转发前去掉管理凭据。
- **平台用户不受信**：身份和权限只来自 Adapter 的可信元数据与 Core 设置。
- **管理入口**：管理 API 只在 `base_url` 的主机名下提供，使用管理 token、Cookie 加 CSRF；监听非回环地址时，`base_url` 必须是 https，或者配置可信反向代理。见[管理参考](management-api.zh-CN.md#core-设置与-http-入口)。

## 继续阅读

| 想了解 | 读这里 |
| --- | --- |
| 模块职责、允许的依赖边、状态归属 | [架构地图](../ARCHITECTURE.md) |
| 协议字段与消息结构 | [`contracts/proto/alonabot/v1`](../contracts/proto/alonabot/v1) |
| 构建、生成、验证与 CI | [源码开发](development.zh-CN.md) |
| 管理 API、设置、休眠与下载镜像 | [管理参考](management-api.zh-CN.md) |
| 用 SDK 写组件 | [组件开发文档](https://alonabot.github.io/docs/tutorial/project)、[Agent API](agent-api.zh-CN.md) |
| 编码与测试规则 | [编码规范](coding-standards.zh-CN.md) |
