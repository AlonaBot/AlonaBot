# 设计决策记录

版本：v0.6-review（2026-10-21；保留已确认 D-01～D-25，2026-10-25～2026-10-28 新增 D-26～D-64，不升版本）

D 表保留历史已确认方向；P/Q 在本文件集中记录议题、处置及当前状态。2026-10-21 的规格完善任务将剩余技术默认值收敛为实施基线，不追加虚构的历史确认记录；2026-10-22 将决策和历史记录归入 adr/。

**实施入口**：[PLAN.md](../PLAN.md) 给出 T-01～T-16 详细工作包与 §7 的休眠扩展 S-01～S-06；[CONTRACTS-v0.1.md](../CONTRACTS-v0.1.md) 固定行为，表内 § 编号均指该契约；验收见 [TEST-MATRIX.md](../TEST-MATRIX.md)。P/Q 处置表已从根目录 CONTRACTS 合并到本文件。D-07/D-18 的历史“Ready”统一解释为新状态机中完成 OnReady 并被 Core 接受的 `Activated`，不是 `Prepared`。

## 已确认

| ID | 决策 | 说明 |
| --- | --- | --- |
| D-01 | 只支持 Native Go 插件 | 每个插件包独立子进程；取消 WASM 运行时/Host ABI/SQL Proxy 设计；允许普通 goroutine、net/http、GORM |
| D-02 | 预构建与源码构建并存 | **预构建优先**，市场按 OS/arch 提供匹配产物；本地源码构建可选 |
| D-03 | 两种 IPC 传输方式 | 本地 TCP RPC 默认，stdio 可选；两者共享相同数据协议，SDK 封装实现 |
| D-04 | IPC 编码 | Length-Prefixed Protobuf；CallID、Session ID；可并发/乱序响应；大媒体不直接塞入常规 RPC |
| D-05 | 插件日志可归因 | Core 捕获 stdout/stderr 并标记插件名；stdio 模式 stdout 专用 IPC，普通日志重定向 stderr；原始 FD 写入仍有风险 |
| D-06 | 进程恢复 | Core 启动、监控、终止、有限自动重启；按此前接受的 5 分钟内连续崩溃 3 次停止自动重启；业务调用不自动重试 |
| D-07 | 注册静态、策略动态 | Handler 注册表在插件启动 Ready 阶段固定；权限、启用状态由 Core 动态更新 |
| D-08 | 命令权限 | 插件/命令组/子命令的权限继承及覆盖由 Core 执行，支持运行时修改并持久化；不得越过全局安全规则 |
| D-09 | 不开放每插件并发上限配置 | Native Handler、不同定时任务可并发；goroutine 开发者自行管理；Core 仍有内部保护，不增加 WebUI 调参 |
| D-10 | 统一事件处理模型 | 融合 AstrBot 的 Hook/优先级和 NoneBot2 的 Rule/Permission/停止传播；不引入额外复杂 Matcher 系统 |
| D-11 | Event Bus | 跨插件异步事件，非持久化尽力投递，离线不补发；不同于有结果的调用 |
| D-12 | Service Call | 由 Core 路由，允许 A→B→C 嵌套，避免同步环；支持超时/取消，不自动重试有副作用的调用 |
| D-13 | Go SDK 注册风格 | 一个插件包一个根插件类（`sdk.Base`，D-39 前为 `plugin.Base`），最多一个 Config 类；独立函数可经注解自动发现，实例方法显式注册；命令组/子命令 |
| D-14 | 数据库与配置 | 首发 SQLite/PostgreSQL（PostgreSQL 由 D-48 移出框架）；原生 GORM；插件 JSON 配置、WebUI schema、secret/hidden/description、i18/；未知配置键默认保留 |
| D-15 | HTTP 路由 | Core 统一公共 HTTP 服务，插件路由按 ID 前缀；不可信插件 UI 不能与 Admin UI 共用浏览器 Origin |
| D-16 | 插件身份与更新来源 | `plugin.yaml` 的 ID 必须有且是 HTTP 路由前缀；市场发布需 `alonabot_plugin_` 前缀，本地命名不规范只警告；更新同时检查 ID + 原可信来源，避免同 ID 不同仓库混淆 |
| D-17 | 首发范围 | Telegram 与最小机器人主链路优先；OneBot/QQ 官方/Discord 等其他 Adapter 后续扩展 |
| D-18 | RPC 握手 | 身份、主/次版本与能力协商，主版本不兼容拒绝启动；只有 Ready 才激活路由 |
| D-19 | 调度归 Core | Core 管理定时触发；不同任务在同一个 Native 插件内可以并发，不应彼此串行阻塞 |
| D-20 | **重复冲突两类规则** | **不同插件包 ID 重复：first-wins，保留最先注册者、后者报告冲突，不可静默替换**。**同插件自动+显式 Handler / HTTP 路由重复：报错**，不再使用较早的 first-wins 路由规则 |
| D-21 | GORM 和迁移 | 提供统一 `gormdb.Open(...)` SDK 入口（D-49 增加 `sqlitedb` 入口）；可选择不迁移、自行实现 `Migrate()`、或由 Core 协助的 `Migrations()`，后两种互斥；跨数据库搬迁由开发者负责 |
| D-22 | 配置与翻译细节 | 每包最多一个 Config，多 Config 构建失败；`i18/` 下翻译 JSON 采用顶层 `config` 嵌套字段，不增加额外 i18n tag；secret/hidden、固定排序遵守最小配置原则 |
| D-23 | SDK 开发体验 | plugin.yaml 元数据 ID 是真源，根类无需重复写 `ID()`；独立函数可自动注册，实例方法显式注册；辅助 Handler 可显式拿到根类指针；优先原生 Go Modules 与少反射 |
| D-24 | 保留两条跨插件协作途径 | 公共接口注册到 Core → Service Call（推荐）；同时**保留**通过 Core 动态调用另一插件已注册命令/事件路由的需求，不能直接持有另一进程 Go 对象；v0.1 只交付 Service Call，动态调用命令/事件路由不在 v0.1 范围内（Q-15） |
| D-25 | 统一 HTTP 入口、插件端口由 SDK 托管 | **用户确认**：所有框架托管的插件 HTTP 路由只由 Core 的统一 HTTP(S) 入口对外提供；插件开发者只注册 Go `http.Handler`/路由，**不需要设置、管理或告知用户开放端口**。SDK 只在有 HTTP 路由时自动启动 loopback 随机端口的内部 Listener；Core 维护 Session 绑定的内部反向代理、路由/权限及启停清理。部署用户只需开放 Core 对外端口，插件内部 Listener 不映射公网；命令/事件/Service Call 仍使用 Protobuf RPC。此确认不代表 Native 插件绝不能自行创建框架外 Listener，也不改变插件 HTML/JS 与 Admin UI 不同浏览器 Origin 的既有安全要求。 |
| D-26 | 可安装的 Adapter | **用户确认（2026-10-25）**：Adapter 与插件一样是安装包和独立 Native 进程，官方 Telegram 也是可卸载的安装包，Core 不静态链接平台实现。同一 Adapter 包只运行一个进程，承载它的全部账号实例；实例由 Core 单独启停，停用 Adapter 即停止该进程，以节省内存。插件可声明对 Adapter 的依赖并使用 Adapter 声明的 extension 片段（例如 Telegram 专有的按钮组件） |
| D-27 | Agent 的分工与范围 | **2026-10-28 由 D-60、D-64 修订**：记忆并入 Agent，Provider 增加官方 `openai` 协议。原文：**用户确认（2026-10-25）**：v0.1 包含 Agent。Core 只提供基础：工具目录（工具是带描述与输入 Schema 的 Service）、按工具的调用策略（allow/confirm/deny）与用户确认、Agent 会话存储。官方 Agent 插件运行对话循环：多个 Provider（OpenAI 兼容与 Anthropic Messages，用标准库实现 SSE 流式）、人格、历史、工具调用、MCP 客户端（使用 `github.com/modelcontextprotocol/go-sdk`）、技能与定时 Agent 任务；流式输出通过编辑同一条消息逐步显示 |
| D-28 | 记忆插件 | **已由 D-60 取代（2026-10-28）**，原文保留：**用户确认（2026-10-25）**：记忆是一种特殊插件，即实现 `sdk/memory` 标准记忆服务的普通插件，不新增 manifest role；Agent 同时只使用一个记忆插件。官方内置的记忆插件仿照 Hermes 的内置记忆：按会话的笔记与按用户的档案各有字符上限，会话开始时冻结快照，`memory` 工具支持 add/patch/remove。跨会话全文检索与同时启用多个记忆插件后置 |
| D-29 | NoneBot2 基础对齐 | **用户确认（2026-10-25）**：v0.1 补齐 keyword/startswith/endswith/fullmatch 文本匹配、基础多轮等待（非持久）、bot 上下线事件与插件依赖声明。为了让多轮等待与 Agent 的一轮对话不受普通请求 30s 期限和 64 条分发链的限制，规格完善时增加延续调用：Handler 可请求 Core 在本次调用结束后，以原可信上下文另发一次独立期限的调用 |
| D-30 | 用户可见文案的 i18n | **用户确认（2026-10-25）**：日志、错误、CLI 输出与 safe_message 只用英文；机器人回复与 WebUI 经 i18n 提供，英文为源语言，附简体中文。插件文案放在 `i18/` 的 `messages`，按 Actor locale → Core 默认 locale → en-US 选择 |
| D-31 | 统一日志 | **用户确认（2026-10-25）**：Core、插件与 Adapter 都使用标准库 `log/slog`。控制台格式仿照 AstrBot：`[时间] [来源] [级别] [文件:行号]: 消息`，来源标签为 `[Core]`、`[Plug:<id>]`、`[Adpt:<id>/<实例>]`，另可输出 JSON。默认 INFO；全局级别在启动时设置，也可在运行时修改。每个插件与 Adapter 可单独设置级别（含 off），持久化在 Core，修改立即生效。WebUI 的显示筛选与输出级别分开 |
| D-32 | 两种运行方式 | **用户确认（2026-10-25）**：发行包开箱即用，包含 Core、官方 Telegram Adapter、Agent、记忆插件与预装插件，首次运行自动安装，默认只启用 Agent、记忆与 help（D-50 改为只启用 help）；预装插件首批为 help、welcome、rss（规格完善时选定）。库模式供开发者把框架当作 Go 库使用：在自己的 `main.go` 中调用根包的 `alonabot.Run`，开发中的组件与 Core 同进程、走同一套协议；`-reload` 时 Core 常驻，只重新编译并重启开发中的组件。官方 Adapter 与官方插件在 v0.1 协议冻结后迁到独立仓库，`adapters/` 与 `plugins/` 不被本仓库其他包 import |
| D-33 | 日志文件 | **用户确认（2026-10-25）**：日志按本地日期轮转写入文件，便于查询前几天的记录；单个文件 100 MiB 后同日切分，合计不超过 1 GiB、保留不超过 30 天，先达到的条件生效。文件开关与上限属于普通设置，可修改 |
| D-34 | 调优参数 | **用户确认（2026-10-25）**：契约中可调的容量、期限与配额集中登记在一个参数文件中，只能在启动时用唯一的 `--args` 覆盖（直接写 `KEY=VALUE`，或给出文本文件），用于测试极端情况。日志级别与日志文件设置不属于调优参数。插件内存上限只做 Go 的 `GOMEMLIMIT` 软上限；渲染不进入 v0.1 |
| D-35 | 规格完整性审查的补充 | **规格完善时选定（2026-10-25）**，按用户「计划中不能留待后续讨论的说法」的要求补齐：message.created 增加 `to_me` 与 SDK 的 `ToMe` Rule；第一个 Core Admin 经 `/whoami` 与 WebUI/CLI 添加；设置文件增加出站代理 `proxy`；Telegram 只支持 long polling，实例配置为 `token` 与 `api_base_url`；开发入口与库模式的实例配置放在 `adapters/<adapter_id>/instances/`；V-33 的工作负载给出具体数值；发布由维护者用 `scripts/dist/` 发布到本仓库的 GitHub Release；Agent、rss、welcome 的命令权限与 rss 的抓取上限给出具体值；v0.1 不提供 webhook 接收与系统服务安装脚本（容器镜像由 D-36 改为提供） |
| D-36 | 容器镜像与 Telegram 代理 | **用户确认（2026-10-26）**：v0.1 提供官方容器镜像（linux/amd64、linux/arm64，发布到 GHCR），内容与发行归档相同；系统服务安装脚本不在 v0.1 范围内。官方 Telegram Adapter 的每个实例可以单独配置代理，优先于全局代理 |
| D-37 | Bootstrap 增量字段编号 | **规格完善时选定（2026-10-26）**：保留现有 Bootstrap 字段 1～9；追加 `LogLevel log_level = 10`（枚举 UNSPECIFIED、DEBUG、INFO、WARN、ERROR、OFF，UNSPECIFIED 按 INFO 处理）与 `SessionLimits session_limits = 11`。SessionLimits 字段固定为 `max_inflight = 1`、`max_payload_bytes = 2`、`request_timeout_ms = 3`；Core 先解析 `inherit`，跨 Session 的 Core 总 payload 预算不下发 SDK。T-02.2 必须在 T-02.1 合并后修改同一 proto/生成文件，避免字段分配与合并冲突 | WIRE §3.2、CONTRACTS §14/§16、PLAN T-02 |
| D-38 | Provider 统一流接口 | **规格完善时选定（2026-10-26）**：Agent Provider 以注入的 `http.Client` 和回调式 `Client.Stream(ctx, req, emit) error` 统一 OpenAI 兼容 Chat Completions 与 Anthropic Messages。文本增量按序发出，工具调用完整组装且参数必须是 JSON 对象；只有收到协议终止标记才算成功，提前 EOF、Provider 错误或回调错误均返回错误，不自动重试。回调式接口用于自然传播背压与调用方错误，避免额外通道隐藏流错误 | PLAN T-12 步骤 1、TEST-MATRIX V-57 |
| D-39 | 根插件 Hook 的默认实现 | **用户确认（2026-10-26）**：`plugin.Base` 移入 `sdk` 包改称 `sdk.Base`（`sdk` 已 import `sdk/plugin`，留在原包无法引用注册器类型），并为 `OnStart`/`OnReady`/`OnStop`、`Config`、`Commands`/`HTTP`/`Schedules` 提供空的默认实现，便于编辑器补全；SDK 原样调用，不区分是否覆盖。`sdk.Root` 列出全部 Hook，根插件覆盖时签名不符即编译失败；Config 的 `OnLoad`/`Validate` 签名不符时 Collect 报错。不强制插件实现任何方法，也不改用选项式注册，回调仍是根插件结构体上的方法 | DESIGN §D、CONTRACTS §4.1、PLAN T-04 步骤 1 |
| D-40 | 配置类型的默认实现，删除 Validate | **用户确认（2026-10-26）**：新增 `config.Base`，配置类型嵌入它即得到原样返回 raw 的默认 `OnLoad`，编辑器可补全；根插件的 `Config()` 改为返回 `config.Config`，覆盖 `OnLoad` 时签名不符在编译期失败，取代 D-39 中 Collect 的运行时检查。删除 `Validate` 回调：多数插件不需要自行校验，需要时在 `OnLoad` 中返回错误即可，失败路径相同（候选失败，不覆盖 Active）。保留 `Config()`，它是根插件上发现配置能力的补全入口，且不需要反射 | CONTRACTS §1.1/§4.1/§4.2/§5.2/§5.3、WIRE §3.2、DESIGN §D/§J、PLAN T-01/T-08.3、TEST-MATRIX V-38 |
| D-41 | `--args` 是全局参数 | **用户确认（2026-10-26）**：调优参数的 `--args` 属于 `alonabot` 根命令，可以写在子命令之前或之后，对 `run`、`serve`、`tune list` 等所有子命令生效；任一子命令执行前先完成解析与校验，出错时不执行子命令。 |
| D-42 | CLI 使用 cobra | **用户确认（2026-10-26）**：`cmd/alonabot` 用 `github.com/spf13/cobra` 组织命令树（`run`、`serve`、`tune`、`plugin`、`admin-token` 等命令组），`--args` 定义为根命令的 persistent flag；参数为 pflag 的 GNU 风格，长参数只接受 `--name`。依赖只用于 `cmd/alonabot`，SDK、Core 包与插件不 import cobra。 |
| D-43 | JSON Schema 使用 jsonschema-go | **用户确认（2026-10-26）**：JSON Schema 的表示、序列化、引用解析、补默认值与校验统一使用 `github.com/google/jsonschema-go`（draft 2020-12），用于配置 Schema、Service 与工具的输入 Schema 和 MCP 工具。配置 Schema 由构建期生成器读取 `json/form/description/secret/hidden` 标签构造 `*jsonschema.Schema`，**不用 `For`/`ForType` 推导**：它默认禁止未知键（与保留未知键冲突）、把未写 `omitempty` 的字段都当必需（与显式 `required` 冲突）、描述读 `jsonschema` 标签（与 `description` 冲突）。`secret`、`hidden` 与表单控件放在 `Extra` 的自定义关键字中，字段顺序用 `PropertyOrder` 保持 Go 声明顺序。默认值只经 `ApplyDefaults` 补到原始 JSON 对象的缺失属性上，`Resolve` 时开启 `ValidateDefaults`。校验以原始 JSON 的解码副本进行，持久化的仍是原文；大整数的精度与范围以 typed decode 的结果为准，首个使用它的切片须验证 `json.Number` 输入的行为。`format` 关键字不校验，需要格式约束时用 `pattern`。该库仍是 v0.x，升级版本按一次行为改动验证。 |
| D-44 | message.created 的命令层与事件路由 | **用户确认（2026-10-26）**：message.created 先经过命令层（普通命令 → 文本匹配 → Raw，规则见 CONTRACTS §2.2）；命令层结束且没有任何 Handler 返回 Stop 时，再按 priority/plugin_id/route_id 派发该类型的事件路由。已识别的命令只消费命令层，不阻止事件路由；Stop 同时结束两者。 |
| D-45 | Core 全局禁止 | **用户确认（2026-10-26）**：契约多处要求「Core 全局禁止」始终生效，但没有定义其内容。v0.1 定为两张持久化禁止表：用户（adapter_instance_id、external_user_id）与会话（adapter_instance_id、chat_id）。命令派发前先检查用户与会话，Send 准入先检查目标会话；策略覆盖与 Core Admin 都不能绕过。管理入口随 T-09 的 Core 设置提供。 |
| D-46 | 官方 OneBot 11 Adapter | **用户确认（2026-10-27）**：v0.1 增加官方 OneBot 11 Adapter（`adapters/onebot11/`），与 Telegram 一样是可安装的独立 Adapter 包。两者在按钮、消息编辑、提及、媒体来源与专属内容上差别很大，用来检验通用 Segment 与 Capability 的设计，并新增跨 Adapter 一致性验收。v0.1 只实现正向 WebSocket（Adapter 主动连接 OneBot 实现端），使用 `github.com/coder/websocket`，该依赖只用于 `adapters/onebot11`；实例配置的 `mode` 只接受 `forward`。反向 WebSocket（`mode: reverse`，实现端连入 Adapter）不在 v0.1 范围内，加入时需要先定 Adapter 的监听与鉴权方式（例如复用 CONTRACTS §7.2 的统一端口代理），已有字段不变；该库同时提供服务端，不需要新依赖。真实 QQ 端依赖非官方协议实现，有账号风险，只由维护者手动测试；CI 与验收使用本地 OneBot 11 模拟服务，没有手测记录时记为未运行。 | CONTRACTS §0/§8/§11/§15.2、MESSAGE-MODEL §7、PLAN T-15、TEST-MATRIX V-71/V-72 |
| D-47 | 插件测试套件 | **用户确认（2026-10-27）**：v0.1 提供公开的插件测试套件 `plugintest`（模块根目录下的包，参照 NoneBug 与 Koishi 的模拟适配器）。插件作者在 `go test` 中用它在同一进程启动真实 Core、套件自带的测试 Adapter 与被测组件，注入消息、按钮点击与 schedule 触发，取得这次分发产生的发送；全程走真实的握手、Session、权限与派发，不 mock SDK。为此根包增加 `Start`：全部组件激活后返回句柄，由调用方停止；`Run` 等于 `Start` 加等待信号，`Start` 也供把 Core 嵌入其他 Go 服务使用。分发结束的通知只经 `internal/inproc` 交给套件，不进入 wire 协议。套件属于 SDK 公共 API，兼容承诺与 `sdk` 相同。WebUI 中可交互的沙盒（参照 Koishi 的 sandbox）不在 v0.1 范围内。 | CONTRACTS §0/§11/§15.4/§17、PLAN T-13/T-16、TEST-MATRIX V-73、ARCHITECTURE |
| D-48 | 框架只托管 SQLite | **用户确认（2026-10-27）**：移除框架托管的 PostgreSQL 后端：Core 不再下发 PG 连接、不管理每插件 role/schema，SDK 删除 `OpenPostgres`，`gorm.io/driver/postgres` 与 pgx 移出依赖清单。需要外部数据库的插件自带驱动、以 secret 配置项取得连接参数并自行管理数据；迁移或 OnStart 会写外部数据库时声明 `prepare_effects=external`，已有安装的 reload、update 与 rollback 走维护门槛。原因：PG 驱动与 SQLite 实现在同一个 `sdk/gormdb` 包中，所有用 SQLite 的插件都被迫链接它，本机实测（macOS footprint）每个插件进程因此多 1.7 MiB 常驻内存、二进制多 9 MiB；PG 分支（role/schema 预置、search_path、resume 只核对）又占了 §6 的大半复杂度。`DATABASE_BACKEND_POSTGRES`、`DatabaseDescriptor.postgres` 与消息 `PostgresConnection` 删除，编号与名称 reserved | CONTRACTS §0/§1.1/§6/§9/§12、WIRE §3.2/§5A、DESIGN §I/§N、PLAN T-08～T-10、TEST-MATRIX V-26/V-31/V-39/V-40 |
| D-49 | 受管 SQLite 的两个入口 | **用户确认（2026-10-27）**：新增 `sdk/sqlitedb`，`sqlitedb.Open(ctx)` 返回 `*sql.DB`；`gormdb.Open(ctx)` 在同一连接上返回 `*gorm.DB`。两个包各定义一组迁移接口，被导入时向 SDK 登记识别方式，SDK 本身不 import GORM，插件只为用到的包付出内存。依据同一组实测：`database/sql` 加 modernc 的进程比空 Go 进程多 0.9 MiB，GORM 再多 1.5 MiB、二进制再多 4.7 MiB。GORM 仍是一等入口，D-01、D-21 不变 | CONTRACTS §6.1/§6.2、DESIGN §I、PLAN T-08 步骤 5/6/9 与 T-14 步骤 2、TEST-MATRIX V-26 |
| D-50 | AI 组件默认不启用 | **2026-10-28 由 D-60、D-62 修订**：记忆并入 Agent，首次运行同时启用官方 WebUI。原文：**用户确认默认不启用（2026-10-27），启用方式由规格完善时选定**：修订 D-32，首次运行只启用 help；Agent 与记忆插件随发行包安装但不启用，由管理员在 WebUI 或 CLI 按需启用，不用 AI 对话的部署不运行这两个进程。没有采用「配置第一个 Provider 时自动启用」：那需要 Core 理解 Agent 插件的配置内容，打破 Core 不感知插件业务配置的边界 | CONTRACTS §15.2、DESIGN §S、PLAN T-14 步骤 4/6、TEST-MATRIX V-66 |
| D-51 | 文本生成服务 | **2026-10-28 由 D-61 修订**：调用辅助以 `@agent` 为目标。原文：**用户确认（2026-10-27）**：官方 Agent 插件提供 JSON Service `llm.generate`，其他插件经 Service Call 复用已配置的 Provider，类型与调用辅助在 `sdk/llm`。服务声明 public，但 Agent 只接受列入 `llm_service_allowed_plugins`（缺省为空）的直接调用方；一次调用就是一次非流式 Provider 请求，不涉及工具、会话、记忆与人格。模型能力因此只存在于 Agent 插件一个进程中，Core 不链接任何模型 SDK 或 Provider 客户端 | CONTRACTS §13/§13.5、DESIGN §R、PLAN T-12 步骤 9 与 T-12.8、TEST-MATRIX V-76 |
| D-52 | Core 依赖允许清单与常驻内存预算 | **用户确认（2026-10-27），预算数值由规格完善时选定**：Core 可执行文件与 SDK 根包只链接 docs/development.md 列出的允许模块，`scripts/check-go-deps.mjs` 在 CI 中对五个目标逐一检查。空闲常驻内存设预算（linux/amd64 PSS）：Core ≤ 24 MiB、无数据库的示例插件 ≤ 10 MiB、`sqlitedb` 插件 ≤ 14 MiB、`gormdb` 插件 ≤ 16 MiB。初值依据 2026-10-27 的实测（Linux PSS，测量脚本未入库）：Core 约 11 MiB、轻量插件进程约 7.5 MiB、链接了 PG 驱动的 GORM SQLite 插件约 17.6 MiB；预算为之后的功能留出余量。超出视为回归，确需提高时在本文件记录新值与原因。插件作者的对应规则写在 [插件开发规范](../../docs/plugin-guide.md) | CONTRACTS §10、DESIGN §O、PLAN §2/T-10、TEST-MATRIX V-74/V-75 |
| D-53 | 插件休眠用正常进程退出 | **用户确认（2026-10-28）**：插件因空闲退出进程，不冻结进程；唤醒复用现有生命周期与 `activation_reason=resume`，正常空闲退出不计入有限重启的异常次数。只有 manifest 声明可休眠的插件参与，未声明的常驻，因为 Core 观察不到插件自行创建的 goroutine 与外部连接 | PLAN §7 S-01、S-02；adr/SLEEP-DESIGN.md §4 |
| D-54 | 已提交代次的路由快照 | **用户确认（2026-10-28）**：Core 持久化已提交代次的路由描述与摘要，插件不运行时仍能被路由定位；唤醒后 REGISTER 的摘要不一致时停止激活，不悄悄改变路由 | PLAN §7 S-01；adr/SLEEP-DESIGN.md §4.2 |
| D-55 | 官方 Adapter 缺少外部入口时常驻 | **用户确认（2026-10-28）**：Telegram long polling 与 OneBot 11 正向 WebSocket 在没有可靠外部入口时不休眠，也不为休眠改变通信模式；同一 Adapter 包仍有运行实例时不停止进程（D-26） | PLAN §7 S-04；adr/SLEEP-DESIGN.md §5 |
| D-56 | Core 整机休眠的前提 | **用户确认（2026-10-28）**：只有外部消息接管、未来任务的外部调度与实例唤醒都已确认时，Core 才整机休眠；任一条件缺失时保持运行。库模式不整机休眠 | PLAN §7 S-05；adr/SLEEP-DESIGN.md §6 |
| D-57 | 外部调度只决定触发时刻 | **用户确认（2026-10-28）**：外部调度器只负责何时唤醒，不接管插件执行、权限与业务数据；任务认领、去重与 `UNKNOWN` 记录仍由 Core 负责，语义为至多一次执行尝试（D-19） | PLAN §7 S-05；adr/SLEEP-DESIGN.md §7 |
| D-58 | 休眠扩展在 v0.1 之后 | **用户确认（2026-10-28）**：休眠以 S-01～S-06 列入计划，先组件级、后整机；S 工作包不是 T-10 的前置，不进入 v0.1 发布门槛，未启用时 v0.1 行为不变 | PLAN §7；TEST-MATRIX V-77～V-81 |
| D-59 | 组件与角色 | **用户确认（2026-10-28）**：可安装的包统称组件，manifest 的 `role` 取 `plugin`（缺省）、`adapter`、`agent` 或 `webui`。只有 `plugin` 是插件，可以安装任意多个；Adapter 每个 `adapter_id` 至多安装一个，Agent 与 WebUI 全局各至多安装一个，同类组件不能并存。任何组件都可以由社区发布并替换官方组件：`install --source <来源>` 指定安装来源，之后的 `update` 只从该来源取版本；同类组件已安装时必须加 `--replace` 并确认，Core 按 `uninstall` 的缺省语义卸载旧组件（保留其数据与来源绑定），再安装新组件，并沿用旧组件的启用状态。会话、记忆、按组件 ID 记录的策略与 Adapter 实例配置不在组件之间迁移。没有采用「可安装多个、只启用一个」：用户要求同类组件不能并存 | CONTRACTS §0/§7.1/§8/§9.1/§9.3/§13/§15、DESIGN §M/§R/§S/§T、PLAN T-03.8/T-09.4、TEST-MATRIX V-82 |
| D-60 | 记忆并入 Agent | **用户确认（2026-10-28），取代 D-28**：记忆不再是独立插件，由 Agent 组件自己实现；删除 `sdk/memory`、`plugins/memory/` 与 `memory.context`/`memory.prefetch`/`memory.sync` 三个服务。官方 Agent 保留 Hermes 式的会话笔记与用户档案、上限与冻结快照；`memory` 工具由 Agent 注册为普通工具（§13.1），模型的调用照常经 `core.tool.invoke`，管理员可以按工具设置策略。原因：记忆唯一的使用方是 Agent，独立插件让每轮对话多出三次跨进程调用和一个常驻进程；检索式记忆可以由社区写成工具插件，对记忆不满意时替换整个 Agent 组件 | CONTRACTS §0/§13.4/§15.2、DESIGN §R/§S、PLAN T-12（步骤 5、切片 T-12.4）、TEST-MATRIX V-59/V-66 |
| D-61 | Agent 组件与 `@agent` | **用户确认（2026-10-28）**：Agent 是 `role: agent` 的组件，全局至多一个。Core 把 Service 调用的目标 `@agent` 解析为已安装的 Agent 组件，`sdk/llm` 的调用辅助以 `@agent` 为目标，调用方不写死官方 Agent 的插件 ID；社区 Agent 实现 §13.5 的 `llm.generate` 即可被现有调用方使用。同时只能启用一个 Agent，避免两个 Agent 同时响应同一条自由对话消息 | CONTRACTS §3/§13/§13.5、DESIGN §R、PLAN T-06/T-12 步骤 9、TEST-MATRIX V-76/V-82 |
| D-62 | WebUI 组件 | **用户确认（2026-10-28）**：管理界面是 `role: webui` 的组件，只含静态文件，没有进程、配置与数据库；Core 不内嵌任何前端。Core 在 Admin 主机上提供已安装 WebUI 的文件，WebUI 只经 Admin API 管理 Core；Admin API 是带主版本号的公开契约，WebUI 在 manifest 中声明支持的主版本范围，接口参考由 T-09 交付。WebUI 在 Admin 源中运行并持有管理员会话，安装它等同授予完整管理权限，按 Native 组件一样展示来源与权限边界。没有安装 WebUI 时 Admin 主机只返回提示，管理员用 CLI 从发行包重装。首次运行时官方 WebUI 与 help 一起启用（修订 D-50）。没有采用 Core 内嵌后备前端：Core 不随前端版本变化 | CONTRACTS §7.3/§7.5/§9.1/§15、DESIGN §T、PLAN T-08.7/T-09.12/T-13/T-14、TEST-MATRIX V-66/V-83 |
| D-63 | 组件图标 | **用户确认（2026-10-28）**：图标随组件包发布：manifest 的 `icon` 是组件自身的图标，配置表单的选项可以各带一个包内图标（例如 Agent 的 Provider 预设）。Core 只提供声明过的 SVG、PNG 或 WebP 文件，响应带 `Content-Security-Policy: sandbox` 与 `nosniff`，WebUI 只用 `<img>` 引用，SVG 中的脚本不会在 Admin 源中执行。Core 与 WebUI 不内置任何厂商图标 | CONTRACTS §5.1/§7.5/§9.1、PLAN T-08.1/T-09.12、TEST-MATRIX V-83 |
| D-64 | 官方整合包与 install.sh | **用户确认（2026-10-28）**：官方整合包是同一套发行物的三种安装方式：发行归档、`deploy/compose.yaml`（使用官方镜像）与 `install.sh`。`install.sh` 支持 Linux（systemd）与 macOS（launchd）：下载发行归档并按同一 Release 的 `SHA256SUMS` 校验，安装程序，初始化管理 token，并注册、启动系统服务；再次运行即升级，`--uninstall` 删除服务与程序、保留数据。移除 §0 中「系统服务安装脚本」的排除项。另外，Provider 客户端增加 `openai` 协议（官方 OpenAI 用 `max_completion_tokens`），流结束时返回停止原因与用量，忽略未知的 Anthropic 事件，错误附带脱敏后的 Provider 说明 | CONTRACTS §0/§15.2/§15.3/§15.6、DESIGN §R/§S、PLAN T-12 步骤 1/T-14、TEST-MATRIX V-84 |

## P 类历史议题与当前处置

| ID | 议题 | 当前处置与实施规则 | 契约位置 | 状态 |
| --- | --- | --- | --- | --- |
| P-01 | 调度重叠/漏触发 | 同任务跳过重叠，离线漏触发不补。 | §4 | 已收敛 |
| P-02 | 资源/超时常量 | 由 D-34 修订：Session、队列的数量/字节预算与分阶段时限有契约规定的缺省值，可在启动时经调优参数覆盖；无 WebUI 与逐插件调参。 | §1/§2/§4/§16 | 已收敛，数值需测试 |
| P-03 | 代次与恢复 | 独占锁与 Journal 唯一提交点约束恢复；旧 Session 失效，不自动重放业务调用。 | §1/§4/§6/§12 | 已收敛 |
| P-04 | 多轮会话 | 由 D-29 取代：v0.1 提供延续调用内的基础多轮等待（非持久）；持久会话状态与暂停 DSL 仍后置。 | §2.4 | 已收敛 |
| P-05 | Handler 双形式 | 两种 Handler 签名共用 HandlerOutcome；自动回复与主动 Send 分明，一次 Chain 可产生多物理消息，手动模式可多次 Send。 | §8 | 已收敛 |
| P-06 | Adapter 形态 | 由 D-26 取代：Adapter 为安装包和独立进程，能力版本化，实例可单独启停。 | §8 | 已收敛 |
| P-07 | 更新恢复 | Generation/Journal 管理受管 SQLite 候选副本；提交后显式恢复，声明 `prepare_effects=external` 的不可隔离变更进入维护（D-48）。 | §6 | 已收敛 |
| P-08 | 配置激活 | Pending→Candidate→Prepared→ActiveGeneration+COMMITTED→OnReady/Activated；secret 引用不可变，提交前失败启动旧版的新 Session。 | §5/§12 | 已收敛 |
| P-09 | 匹配/传播 | 普通命令最长路径优先，识别后拒绝不回落；同事件 priority 小先、同级稳定 ID 排序，Regex/Raw 与 Stop 语义固定。 | §2/§3 | 已收敛 |
| P-10 | Adapter 动态装载 | 已有 Adapter 实例开关首发；新类型热加载后置。 | §8 | 已收敛 |
| P-11 | JSON 包装层 | 运行配置为普通 JSON，无 config 包装；i18 翻译使用 config 包装。 | §5 | 已收敛 |
| P-12 | 事件与 Actor | 最低事件清单采用类型化 payload；scheduled 等系统任务没有用户角色或默认 Admin。 | §2/§4 | 已收敛 |
| P-13 | 生命周期 | OnStart/OnReady/OnStop 首发，配置回调独立 OnLoad（D-40 删除 Validate）；OnConfigChange 后置。 | §4/§5 | 已收敛 |
| P-14 | 多账号 | Target 明确 adapter/instance/account/chat/thread，凭证按实例隔离；不内置 X/Twitter 动态凭证。 | §8 | 已收敛 |

## Q 类历史问题与当前处置

| ID | 议题 | 当前处置与实施规则 | 契约位置 | 状态 |
| --- | --- | --- | --- | --- |
| Q-01 | RPC 精确格式 | WIRE-FORMAT 固定 Envelope、控制失败、Bootstrap、状态、错误与 deadline/Cancel。 | §1 | 已收敛 |
| Q-02 | Native 信任边界 | Native 插件按可信本地代码运行，不提供跨平台恶意代码沙箱。 | §7 | 已收敛；安全测试仍需做 |
| Q-03 | 真实多插件内存 | T-10 测量多个真实不同 Native 可执行文件的总资源；不能以相同 worker 的 PSS 预设结论。空闲常驻内存另有 V-74 的预算（D-52）。 | §10 | 性能数值只能实施期验证 |
| Q-04 | 五平台/CGO | T-01 锁定驱动/工具链，T-10 完成五平台真实构建与启动；CGO=0 优先，CGO 插件单独发布并记录条件。 | §7/§9 | 已收敛；需真实构建 |
| Q-05 | 事件 Hook/排序 | CONTRACTS 与 MESSAGE-MODEL 固定 priority、Rule 错误、StopPropagation、Hook 阶段、结果及允许修改范围。 | §2 | 已收敛 |
| Q-06 | 媒体 | Core 顺序分块上传，MediaRef 使用受限 token、TTL、配额和消费租约。 | §1.3 | 已收敛 |
| Q-07 | 版本/签名 | Core/protocol major.minor 与 SDK 编译版本区分；manifest、归档摘要、来源身份和外部 release 清单签名分别校验。 | §7/§9 | 已收敛 |
| Q-08 | Config/i18 | Go Schema/JSON tag/i18 嵌套对应原始 JSON；固定类型、迁移、默认、PATCH、未知键保留与翻译规则。 | §5 | 已收敛 |
| Q-09 | PostgreSQL | 由 D-48 撤销：框架不再托管 PostgreSQL，需要外部数据库的插件自带驱动并自行管理，按 `prepare_effects=external` 的维护门槛更新。 | §6 | 已撤销 |
| Q-10 | 第三方 Adapter | 由 D-26 收敛：Adapter 以安装包和独立进程交付，官方 Telegram 与第三方使用同一机制。 | §8 | 已收敛 |
| Q-11 | 回执/片段 | MESSAGE-MODEL 定义两种 Handler 输出、物理发送计划、部分成功回执与不支持片段的明确失败。 | §8 | 已收敛 |
| Q-12 | CLI/分发 | 交付已承诺的开发/管理 CLI、tar.gz、manifest、外部 release 清单与 GitHub Release 来源规则。 | §9 | 已收敛 |
| Q-13 | Journal/快照 | Journal/Generation、SQLite Backup 候选、进程锁和持久化次序共同保证提交恢复；声明 external 的不可逆变更拦截，恢复需故障注入。 | §6/§12 | 已收敛；恢复需故障注入 |
| Q-14 | SDK 生成 | 在插件目标构建范围内扫描本模块包级函数；单 root、自动/显式冲突报错、实例显式绑定、Schema 生成由 T-01/T-04 编译验证。 | §9 | 已收敛；编译器必须验证 |
| Q-15 | 动态命令调用 | v0.1 交付 Service Call/Event Bus；动态跨插件命令 Router 后置 v0.2，保留稳定 route ID 和需求。 | §3 | 明确后置（需求保留） |
| Q-16 | Reload/Ready | 配置变更通过新 Session 重启式应用；OnReady 完成并 Activated 后才接业务，OnStop 负责停止。 | §4 | 已收敛 |
| Q-17 | 配置包装 | 运行时 Config 直接使用 JSON 字段，无顶层包装。 | §5 | 已收敛 |
| Q-18 | 进程/密钥/账号 | 合作型进程锁/子进程树清理与实例凭证作用域有明确规则，需跨平台验证；Native 仍不是沙箱。 | §7/§8/§12 | 已收敛；安全测试仍需做 |
| Q-19 | Core 托管的轻量 KV | 只需保存少量状态（订阅、计数、开关）的插件，若由 Core 经 RPC 提供按插件隔离的 KV，就不必链接 SQLite，进程可保持在无数据库插件的水平。v0.1 不提供，简单状态继续用 `sqlitedb`。待定内容：容量与配额、是否纳入代次候选与回退、管理与 purge。 | — | 待定 |

“已收敛”表示已有实施规则，后置需求仍保留；性能、驱动、生成器及故障恢复的结果须在 TEST-MATRIX 中取得实际证据。

## 旧版 PRD 冲突解决（现行方案优先）

- WASM-only、WASI Reactor、Host ABI/SQL Proxy、单实例串行：由 D-01/D-03/D-04/D-09 取代。
- 同包多个根插件类、方法注解必需：由 D-13 的单根类/独立函数注解/显式方法注册取代。
- 旧版默认 PruneUnknown=true：被 D-14 **默认保留未知键**取代，OnLoad 保留（Validate 由 D-40 删除）。
- 强制 Host FS/HTTP/Exec 权限拦截：Native 可绕开普通 SDK，无 OS 级隔离时不得宣称安全隔离。
- 通用事件自动补发/Outbox：Event Bus 按 D-11 尽力投递，业务 Service Call 不自动重试；可靠发信应单独定义。
- 同一个 .wasm 跨平台：改为同版本下按 OS/arch 分发不同 Native 可执行文件。

完整 FR-01～FR-30 迁移表见 [LEGACY-PRD-COMPARISON.md](history/LEGACY-PRD-COMPARISON.md)。

PDF **之后**的长时间设计讨论覆盖情况、仍未完整恢复的议题及冲突区分见 [POST-PRD-AUDIT.md](history/POST-PRD-AUDIT.md)。

## 决策变更记录

- 2026-10-19：建立正式设计记录；Native 子进程取代 WASM-only。
- 2026-10-19：确认预构建优先、源码构建可选；双传输 Protobuf、日志归属、动态命令权限与最小配置。
- 2026-10-19：确定 Event Bus 尽力投递、允许嵌套 Service Call、融合 AstrBot 与 NoneBot2 事件处理。
- 2026-10-19：v0.2 对照旧版 32 页 PRD，保留可迁移需求，显式标注过时的 WASM/Host 安全假设及仍待批准内容。
- 2026-10-19：v0.3 对 PDF 后的设计约束再次核查：纠正插件 ID first-wins 与同插件 Handler 冲突报错混淆；增补可选迁移接口、i18/ 文案结构、根插件与 Go 模块协作、未来方向。
- 2026-10-19：v0.4 自审补入旧 PDF 事件身份/权限基线、动态 Router 调用、关键生命周期阶段、配置文件包装层候选及多账号作用域；区分安全必需条件与未定实现细节。
- 2026-10-21：v0.5 新增 CONTRACTS-v0.1.md：整理**待批提案**中的配置重启、事件排序、RPC、服务权限、媒体、备份/更新、原生信任和跨平台验收。原 P/Q 状态保留，避免假定审批。
- 2026-10-21：**D-25 已确认**：框架托管的插件 HTTP 只注册 Handler，不管理端口；统一由 Core 公网入口反向代理到 SDK 自动管理的 loopback Listener。只更新这一条决定，不代表批准其他 v0.6 提案。


- 2026-10-21：按用户本轮“自动完善剩余细节、审阅可行性并修复缺陷”的授权，同版收敛 P/Q 与执行契约；D-01～D-25 历史内容保持，新增技术细节的来源就是本轮规格完善任务。
- 2026-10-22：按用户要求迁入 adr/；合并原 CONTRACTS §11 的 P/Q 处置表，根目录改为直接实施工作包与契约。D-01～D-25 原文和既有编号保持。
- 2026-10-21：T-01 冻结 `alonabot.v1` 的 Protobuf 字段号；对规格未给出编号或取值的消息与枚举做出实现选择，记录于 WIRE-FORMAT §5A。不改变既有 Envelope 1～16 与错误码 1～17。
- 2026-10-25：**D-26 已确认**：Adapter 改为可安装包与独立进程（同包一个进程、实例单独启停），插件可依赖 Adapter 的 extension 片段。同步修改 CONTRACTS §0/§8/§9.1/§11、WIRE §3.4、MESSAGE-MODEL §2/§6、DESIGN §M、PLAN T-03/T-09 与 TEST-MATRIX V-48；P-06、Q-10 改为已收敛。
- 2026-10-25：**D-27～D-30 已确认**：v0.1 加入 Agent 基础与官方 Agent 插件、记忆插件、NoneBot2 基础对齐（文本匹配、延续调用与多轮等待、bot 上下线事件、插件依赖）和用户可见文案的 i18n。同步修改 CONTRACTS §0/§1.2/§2/§5.4/§8/§9.1/§11/§13、MESSAGE-MODEL、WIRE §3.4/§3.5、DESIGN、PLAN（新增 T-11、T-12 与相关切片）及 TEST-MATRIX V-49～V-61；P-04 改为已收敛。
- 2026-10-25：**D-31、D-32 已确认**：统一日志（slog、AstrBot 风格控制台格式、按来源的级别）与两种运行方式（发行包与预装插件、库模式与热部署），并补上 Core 的启动命令与设置文件。同步修改 CONTRACTS §0/§9.3/§11/§14/§15、WIRE §3.2/§3.5、DESIGN §O/§S、PLAN（新增 T-02.1、T-09.10、T-13、T-14）、TEST-MATRIX V-62～V-67。
- 2026-10-25：**D-33、D-34 已确认**：日志文件按日期轮转并设默认上限；可调的容量、期限与配额集中登记为调优参数，经 `--args` 在启动时覆盖。同步修改 CONTRACTS §0/§11/§14/§15/§16、WIRE §3.2、PLAN（T-02.1 扩展、新增 T-02.2）、TEST-MATRIX V-68、V-69；P-02 改为由 D-34 修订。
- 2026-10-25：**D-35**：规格完整性审查，删除全部「后续再设计、届时、待定」类说法并补齐 D-35 所列内容；D-24 的描述改为已处置的结论；goals.md 的「待细化」状态改称「待编写」。同步修改 CONTRACTS §0/§2.2/§6.1/§7.3/§8/§9.1/§9.3/§15、MESSAGE-MODEL §1.1/§6.1、WIRE、DESIGN、PLAN T-02/T-03/T-05/T-09/T-10/T-12/T-13/T-14、TEST-MATRIX。
- 2026-10-26：**D-36 已确认**：v0.1 提供官方容器镜像，Telegram 实例支持单独的代理配置。同步修改 CONTRACTS §0/§8/§15（新增 §15.3，原 §15.3、§15.4 顺延为 §15.4、§15.5）、PLAN T-03/T-10/T-14、TEST-MATRIX V-45 与新增的 V-70。
- 2026-10-26：**D-39 已确认**：`plugin.Base` 改为 `sdk.Base` 并提供全部根插件 Hook 的空默认实现，签名不符的 Hook 在编译期或 Collect 时报错。同步修改 DESIGN §D、CONTRACTS §4.1、PLAN T-04 步骤 1。
- 2026-10-26：**D-40 已确认**：新增 `config.Base`，`Config()` 返回 `config.Config`；删除 Validate 回调，校验并入 OnLoad。同步修改 CONTRACTS §1.1/§4.1/§4.2/§5.2/§5.3、WIRE §3.2、DESIGN §D/§J、PLAN T-01/T-08.3、goals.md goal-30、TEST-MATRIX V-38。
- 2026-10-26：**D-41 已确认**：`--args` 改为根命令的全局参数。同步修改 CONTRACTS §16 与 goals.md goal-76。
- 2026-10-26：**D-42 已确认**：CLI 改用 cobra，`--args` 为 persistent flag，可写在子命令前后。同步修改 D-41、CONTRACTS §16、docs/development.md 的依赖清单、docs/execution-order.md 与 goals.md goal-76。
- 2026-10-26：**D-43 已确认**：JSON Schema 使用 `github.com/google/jsonschema-go`，配置 Schema 由生成器构造、不用 `For` 推导。同步修改 docs/development.md 的依赖清单与 docs/execution-order.md。
- 2026-10-26：**D-44、D-45 已确认**：message.created 在命令层之后进入事件路由；Core 全局禁止定为用户与会话两张禁止表。同步修改 CONTRACTS §2.2、§2.3 与 goals.md goal-12、goal-13。
- 2026-10-27：**D-46、D-47 已确认**：v0.1 增加官方 OneBot 11 Adapter（正向 WebSocket，依赖 `github.com/coder/websocket`）与插件测试套件 `plugintest`，根包增加 `Start`。同步修改 CONTRACTS §0/§8/§11/§15.2/§15.4 并新增 §17，MESSAGE-MODEL 新增 §7（原 §7 顺延为 §8），PLAN 新增 T-15、T-16 并修订 T-13 步骤 1，TEST-MATRIX 新增 V-71～V-73，同步 ARCHITECTURE、docs/development.md 的依赖清单、docs/execution-order.md 与 goals.md（新增 goal-78～goal-82，修订 goal-67、goal-46、goal-73）。
- 2026-10-27：评估「插件与 Adapter 合并进同一进程」后**未采纳**，维持 D-01、D-26 的一插件一进程。实测（测量脚本未入库）：20 个插件时，单进程的空闲内存是 NoneBot2 的 0.2～0.3 倍；代价是第三方插件必须在用户机器上构建（atri 上 512 MiB、1 CPU 首次构建三轮都 OOM，1 GiB、2 CPU 约 82 秒），依赖不再隔离，故障隔离与真正的热安装也随之失去。同期测得 Python 子进程插件每个进程的开销约为 Go 的 4～5 倍（macOS footprint），说明一插件一进程对 Go 仍然划算；需要控制的是 Core 与插件引入的依赖，由 D-48～D-52 处理。
- 2026-10-27：**D-48～D-52 已确认**，新增待定的 Q-19：框架只托管 SQLite，受管 SQLite 增加 `database/sql` 入口，AI 组件默认不启用，Agent 插件提供 `llm.generate`，Core 设依赖允许清单与常驻内存预算。同步修改 CONTRACTS §0/§1.1/§6/§9/§10/§11/§12/§13（新增 §13.5）/§15.2、WIRE §3.2/§5A、DESIGN §I/§N/§O/§R/§S、PLAN（§1、§2、T-08～T-10、T-12 新增步骤 9 与切片 T-12.8、T-14）、TEST-MATRIX（V-26/V-31/V-39/V-40/V-66，新增 V-74～V-76）；删除 `sdk/gormdb` 的 PostgreSQL 实现与测试、`contracts/manifest.PostgresIdentifier` 与 CI 的 PostgreSQL service；新增 `scripts/check-go-deps.mjs` 与插件开发规范 docs/plugin-guide.md；同步 ARCHITECTURE、README、CONTRIBUTING、docs/development.md、docs/coding-standards.md、docs/execution-order.md、测试技能与 goals.md（新增 goal-83，修订 goal-32、goal-64 等）。
- 2026-10-28：**D-53～D-58 已确认**：分层休眠与按需唤醒以 S-01～S-06 列入 PLAN §7，设计依据存为 adr/SLEEP-DESIGN.md，新增 TEST-MATRIX V-77～V-81；v0.1 契约不变，各 S 工作包实施时再把对应设计并入 CONTRACTS、WIRE 与 DESIGN（包括修订 CONTRACTS §4.3 的离线 tick 规则）。
- 2026-10-28：审阅「未实现能力与规格补充」草案的 G-01～G-09 后，补齐三处 v0.1 缺口：官方 Agent 插件按事件的 capabilities 选择回复方式，目标不声明 `message.edit`（如 OneBot 11）时生成结束后一次发送；会话键由 adapter_instance_id、chat_id 与可选的 thread_id 组成，只用于会话存储，不影响 Target 与显式 To；库模式中同一组件 ID 登记两次时启动前报错。`plugin init` 生成的 `.gitignore` 忽略 `.alonabot/`。同步修改 CONTRACTS §9/§15.4、DESIGN §R、PLAN T-09/T-12/T-13、TEST-MATRIX V-57/V-64 与 goals.md goal-38/goal-59/goal-67。**未采纳**：CLI 的 `--plugin . --reload`（与 T-13.3 重复成第二条热部署路径）；同 module 开发 Adapter 不随热部署重启（同 module 的 Adapter 通常正在开发，现有规则已保证其他 module 的组件不重启）；开发模式配置热加载与编辑器关联示例（重启即可生效，示例不属于规格）。G-02 的其余内容、OneBot 上的工具确认与 Provider 管理已由 D-26、CONTRACTS §13 与 T-12 覆盖；G-06～G-09 与休眠设计重复，以 adr/SLEEP-DESIGN.md 为准。
- 2026-10-28：**D-59～D-64 已确认**：组件分为插件、Adapter、Agent 与 WebUI 四种角色，后三者同类不能并存并可由社区组件替换；记忆并入 Agent（D-28 被取代）；`@agent` 目标；WebUI 改为可替换的静态组件；组件图标；官方整合包与 install.sh；Provider 客户端修正。同步修改 CONTRACTS §0/§3/§5.1/§7.1/§7.3/§8/§9.1/§9.3/§11/§13/§15（新增 §7.5、§15.6）、DESIGN §M/§R/§S（新增 §T）、PLAN（T-08、T-09、T-12、T-13、T-14，新增切片 T-09.12、T-14.7）、TEST-MATRIX（V-59、V-66、V-76，新增 V-82～V-84）、ARCHITECTURE、README、docs/plugin-guide.md、docs/execution-order.md 与 goals.md。
