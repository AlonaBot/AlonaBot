# AlonaBot v0.1 详细实施计划

**技术基线：** v0.6-review  
**编排更新：** 2026-10-22  
**实施状态：** T-01、T-02 已实现，范围内分项验收的证据见 [TEST-MATRIX §3.1](TEST-MATRIX.md#31-已运行的分项证据t-01t-02)；其中 Linux/Windows 的真实运行留给 T-10，跨任务验收在各自集成点闭合。T-03～T-10 未开始，按 §3.1 的执行切片推进。本文件规定接下来要执行的工作、产出和完成条件。

从 T-01 开始，按前置任务推进。每项任务交付可编译、可验证的增量；跨任务验收在对应集成点闭合。协议和行为以 [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md)、[WIRE-FORMAT.md](WIRE-FORMAT.md)、[MESSAGE-MODEL.md](MESSAGE-MODEL.md) 为准，验收编号统一使用 [TEST-MATRIX.md](TEST-MATRIX.md)。

## 1. 最终交付

交付一个 Go Core、Go 插件 SDK、Telegram Adapter 和管理 CLI/WebUI。用户能够安装预构建插件，配置并启用它，在指定聊天/话题执行命令、发送消息和媒体，使用 Service、Event Bus、Scheduler、插件 HTTP 与 SQLite/PostgreSQL；管理动作能够按确定的代次和维护规则更新、停止及恢复。

首发目标为 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64。每插件一个 Native 子进程，TCP 默认、stdio 可选；框架托管 HTTP 只开放 Core 的一个公开端口。动态跨插件命令 Router、其他正式 Adapter、Agent、多语言 SDK、多轮会话 DSL 和额外客户端沿既定范围后置。

当前项目已有 `spec/` 和独立的 `experiments/ogre-dna/`，尚无正式 Core/SDK。实验不作为下述任务的实现或验收依据。

## 2. 实施目录和工作流

以下是**实施时拟新增的代码目录**，不是现有文件清单。采用一个 Go module；T-01 固定实际 module path、工具链和生成方式，后续任务复用它们。

| 拟建位置 | 负责内容 | 首次交付 |
| --- | --- | --- |
| `go.mod`、`go.sum` | 实际模块路径、Go 与依赖版本 | T-01 |
| `contracts/` | Protobuf 源文件、生成类型、消息/配置/manifest Schema 和兼容 fixture | T-01 |
| `sdk/` | Native 会话、Context/Send、注册、生命周期、媒体与按需 GORM 包 | T-01 给出类型，T-02～T-08 逐步实现 |
| `internal/core/` | 状态、进程、RPC、路由、权限、服务、调度、HTTP 与管理编排 | T-02 起 |
| `internal/adapters/telegram/` | Telegram 接收、发送、能力和平台映射 | T-03 |
| `cmd/alonabot/` | Core 启停、开发者与管理 CLI 装配 | T-02 起，T-09 完整 |
| `web/admin/` | 配置表单/i18 组件、管理界面与静态资源 | T-08 组件，T-09 完整 |
| `examples/plugins/` | 随 SDK 版本同步编译的最小示例和集成样例 | T-02 起 |
| 各包的 `testdata/` 与测试文件 | 非生产 fixture、错误注入和行为验证 | 随负责任务交付 |

生成代码由受版本约束的生成命令产生，禁止手工修改后依赖下一次生成碰巧保留。工具可执行文件和业务凭证不放进发布包。T-01 建立 `go generate ./...` 的实际入口；之后在仓库根执行以下命令作为日常编译与单元验证入口：

```sh
go generate ./...
go test ./...
go build ./...
```

上述是实施后要交付的工作流；T-01/T-02 已建立其中 `go.mod`、`contracts/`、`sdk/`、`internal/core/`、`cmd/alonabot/`、`examples/plugins/` 的部分，实际状态以 [ARCHITECTURE.md](../ARCHITECTURE.md) 为准。Telegram、PG、浏览器和目标平台测试单独记录环境与命令；缺少真实环境不能用普通单元测试结果替代。每个任务先验证直接变更及其边界，再做工作包要求的集成检查。

## 3. 任务依赖与交付次序

| 任务 | 前置任务 | 本项完成时的可用能力 | 当前状态 |
| --- | --- | --- | --- |
| T-01 | 无 | 可编译契约、固定生成方式、驱动/备份验证基线 | 已实现；分项证据见 TEST-MATRIX §3.1 |
| T-02 | T-01 | 最小真实 Native 进程的状态、双 IPC、启动和停止 | 已实现；darwin/arm64 运行通过，Linux/Windows 运行未执行（T-10） |
| T-03 | T-02 | Telegram 入站与显式 Send、媒体、按钮、真实回执 | 未开始 |
| T-04 | T-03 | 根插件 SDK、自动/显式注册、命令树与动态权限 | 未开始 |
| T-05 | T-04 | 完整事件链、Rule、Hook、传播与可信身份 | 未开始 |
| T-06 | T-05 | Service Broker 与有界 Event Bus | 未开始 |
| T-07 | T-05 | Core Scheduler、SDK HTTP 与统一代理组件 | 未开始 |
| T-08 | T-04 | 配置/Secret/Schema/i18 与插件 DB 候选准备能力 | 未开始 |
| T-09 | T-06、T-07、T-08 | 安装更新/回退、完整管理认证、CLI/API/WebUI | 未开始 |
| T-10 | T-02～T-09 | 全部验收闭合、目标平台产物与资源报告 | 未开始 |

T-08 可在 T-04 后与事件链工作并行；T-06、T-07 在 T-05 后可并行。T-10 的测试设施随任务建立，最终发布门槛在依赖全部完成后检查。

```mermaid
flowchart TD
  T01["T-01 契约与工具链"] --> T02["T-02 Core 与进程"]
  T02 --> T03["T-03 Telegram 与消息"]
  T03 --> T04["T-04 SDK 与命令"]
  T04 --> T05["T-05 事件与 Hook"]
  T04 --> T08["T-08 配置与数据库"]
  T05 --> T06["T-06 服务与广播"]
  T05 --> T07["T-07 调度与 HTTP"]
  T06 --> T09["T-09 管理与更新"]
  T07 --> T09
  T08 --> T09
  T09 --> T10["T-10 发布验收"]
```

### 3.1 执行切片

每个工作包按下表拆成切片，一个切片对应一个 goal（`goal-N` 目录、分支与 PR）。执行者实现，审查者审查并合并，角色与授权见 [CONTRIBUTING「角色与合并」](../CONTRIBUTING.md#角色与合并)。可直接派发的 goal 语句在 [goals.md](../goals.md)，并行派发、共享文件与合并顺序的约束在 [执行顺序](../docs/execution-order.md)。

切片只是把 §4 的步骤重新分组，工作包的范围、验收和失败停止条件不变。为了能并行执行，切片的前置按实际用到的成果列出，可以比 §3 的工作包依赖更细。例如调度触发计算、配置 JSON 语义和安装包预检不依赖消息链路，因此可以提前执行。

- **前置**只列直接依赖。前置都已合并的切片可以派发。
- **主要验收**列出该切片负责的 V 分项。分项证据写在该切片的 tasks.md 中；每个工作包的最后一个切片负责把证据汇总进 TEST-MATRIX §3。整项 V 何时关闭仍按 §5 的跨任务规则。
- **状态**由审查者在合并前写入该切片的 PR：未开始，或「已完成」。
- 需要真实外部环境的分项已在表中注明；缺少环境时，该分项记为未运行。

| 切片 | goal | 内容 | §4 步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-03.1 | goal-2 | 首个可用 SDK 切片：事件路由注册、Context/Send、业务分发；Core 的 Adapter 接口、内存 Adapter、固定服务 mux、事件派发与 `core.send`；echo 插件经真实 REGISTER/ACTIVATED 后收发 | T-03 步骤 1 | T-02 | V-07、V-08 的 SDK 分项 | 未开始 |
| T-03.2 | goal-3 | Target 与最小授权：`allowed_adapter_instances` 持久化、显式 To、后台 Send、逐次发送准入、NO_DEFAULT_TARGET、Core 内部许可读写入口 | T-03 步骤 2 | T-03.1 | V-07 | 未开始 |
| T-03.3 | goal-4 | Telegram 发送计划与回执：Chain 完整预检、物理操作计划、UTF-16 实体、相册/引用/按钮渲染、PartResult 汇总 | T-03 步骤 5 | T-03.4 | V-08、V-35 | 未开始 |
| T-03.4 | goal-5 | Telegram 实例与接收：Bot API 客户端、本地 Bot API 模拟服务、getMe 绑定、同账号唯一 owner、webhook 冲突、long polling、逐条游标、满队列丢弃计数、7 天轮次重建、401/409、媒体组 | T-03 步骤 3 | T-03.1 | V-45 | 未开始 |
| T-03.5 | goal-6 | 媒体上传与配额：BeginUpload/Chunk/Commit、大小/MIME/SHA256 校验、全部固定配额与期限、owner 退出释放 | T-03 步骤 4（上传） | T-03.1 | V-09 | 未开始 |
| T-03.6 | goal-7 | 入站媒体与 Grant：IncomingMedia token、首次 Read 惰性取件并校验缓存、多 Session 共享、Grant 范围、租约 | T-03 步骤 4（入站与授权） | T-03.2、T-03.4、T-03.5 | V-44、V-09 | 未开始 |
| T-03.7 | goal-8 | 按钮 token 与回调路由：token 生成与持久化绑定、answerCallbackQuery、逐次鉴权、代次失效 | T-03 步骤 6 | T-03.2、T-03.3 | V-08 的按钮分项 | 未开始 |
| T-03.8 | goal-9 | Telegram 接入开发入口，echo/media 示例与发送接收故障矩阵，汇总证据 | T-03 步骤 7 | T-03.6、T-03.7 | V-07、V-08、V-09、V-35、V-44、V-45；真实 Telegram 需要测试账号 | 未开始 |
| T-04.1 | goal-10 | 根插件 API 与显式注册；静态注册索引、重复拒绝、跨插件遮蔽的稳定选择 | T-04 步骤 1、3 | T-03.8 | V-10 的注册分项、V-02 的注册原子性 | 未开始 |
| T-04.2 | goal-11 | 构建期生成器 `sdk/generate`：限定扫描范围、build tags、生成注册代码、稳定输出；`cmd/alonabot` 的构建检查入口 | T-04 步骤 2 | T-04.1 | V-10 | 未开始 |
| T-04.3 | goal-12 | 命令选择与参数绑定：完整 token、@username、最长路径/别名→Regex→Raw、消费状态、参数语法、usage/silent | T-04 步骤 4 | T-04.1 | V-11 的命令分项 | 未开始 |
| T-04.4 | goal-13 | 持久化命令策略与管理服务层：继承、禁用、全局禁止、下一次派发生效 | T-04 步骤 5 | T-04.1 | V-12、V-43 的服务层分项 | 未开始 |
| T-04.5 | goal-14 | 用同一 SDK 升级示例，完成构建与策略验收，汇总证据 | T-04 步骤 6、7 | T-04.2、T-04.3、T-04.4 | V-10、V-11、V-12、V-43 的 T-04 分项 | 未开始 |
| T-05.1 | goal-15 | 可信事件上下文与有界同事件分发：64 条并发链、256 条/8 MiB 等待队列、稳定候选排序、链执行与读循环分离 | T-05 步骤 1、2 | T-04.5 | V-13、V-16 的 fixture 分项 | 未开始 |
| T-05.2 | goal-16 | SDK Rule、HandlerOutcome、自动/手动 Send、StopPropagation 锁存与错误分支 | T-05 步骤 3、4 | T-05.1 | V-14、V-11 的 Rule 分项 | 未开始 |
| T-05.3 | goal-17 | PreEvent/BeforeSend/AfterSend 三阶段 Hook，send_forbidden 的保存与继承 | T-05 步骤 5、6 | T-05.2 | V-15 的直接路径分项 | 未开始 |
| T-05.4 | goal-18 | 多进程事件流水线验收，汇总证据 | T-05 步骤 7 | T-05.3 | V-11、V-13～V-16、V-43 的 T-05 分项 | 未开始 |
| T-06.1 | goal-19 | Service/Bus 描述与编码校验、SDK 入口；Broker 定位与准入 | T-06 步骤 1、2 | T-05.4 | V-19 | 未开始 |
| T-06.2 | goal-20 | 父调用链上下文传播、循环与深度检查、跨段取消与背压 | T-06 步骤 3、4 | T-06.1 | V-18、V-42 的嵌套分项 | 未开始 |
| T-06.3 | goal-21 | 非持久 Event Bus：发布准入、逐订阅者 256 条/4 MiB 队列、drop-newest 计数 | T-06 步骤 5 | T-06.1 | V-17 | 未开始 |
| T-06.4 | goal-22 | Session 变更时的清理与计数，A/B/C 端到端验收（含 Hook→Service→Send），汇总证据 | T-06 步骤 6、7 | T-06.2、T-06.3 | V-15、V-17、V-18、V-19、V-42 | 未开始 |
| T-07.1 | goal-23 | interval/五字段 cron/IANA 时区/夏令时的表达式校验与下一次触发计算（纯计算包） | T-07 步骤 2 | T-02 | V-20 的计算分项 | 未开始 |
| T-07.2 | goal-24 | schedule 注册校验，运行槽与调度调用：独立 run_id、15 分钟期限、同 job 跳过重叠、迟到响应与断链 | T-07 步骤 1（schedule）、3 | T-07.1、T-05.4 | V-20、V-46、V-16 和 V-05 的 scheduled 分项 | 未开始 |
| T-07.3 | goal-25 | HTTP 注册校验，SDK 托管 HTTP Listener 与 Core Session 代理映射 | T-07 步骤 1（HTTP）、4、5 | T-04.5 | V-22 的组件分项 | 未开始 |
| T-07.4 | goal-26 | 两个 Host 的边界、插件兑换票据与插件会话 | T-07 步骤 6 | T-07.3 | V-47 的组件分项 | 未开始 |
| T-07.5 | goal-27 | 统一启停与清理，时间、进程和 HTTP 边界验收，汇总证据 | T-07 步骤 7、8 | T-07.2、T-07.4 | V-20、V-21、V-22、V-46、V-47 的 T-07 分项 | 未开始 |
| T-08.1 | goal-28 | 配置字段描述与构建检查（在 T-04 生成器上扩展） | T-08 步骤 1 | T-04.5 | V-25 的描述分项 | 未开始 |
| T-08.2 | goal-29 | 配置 JSON 语义与 Pending：原始对象保留、完整替换、Merge Patch、编辑快照导入、base_generation | T-08 步骤 2 | T-02 | V-23、V-38 | 未开始 |
| T-08.3 | goal-30 | SDK 配置准备顺序：OnLoad→补缺键默认→typed decode/Schema→Validate | T-08 步骤 3 | T-08.1、T-08.2 | V-37、V-21 的配置分项 | 未开始 |
| T-08.4 | goal-31 | 不可变 Secret Store：引用、解引用、重新抽取、脱敏读取、受保护清理 | T-08 步骤 4 | T-08.1、T-08.2 | V-25 | 未开始 |
| T-08.5 | goal-32 | `gormdb.Open` 连接生命周期、迁移互斥与 versioned 迁移、PG 门槛与维护许可 | T-08 步骤 5、6 | T-08.3 | V-26、V-39；PG 需要测试实例 | 未开始 |
| T-08.6 | goal-33 | SQLite 候选复制与资源交接，写出给 T-09 的调用约定 | T-08 步骤 7 | T-08.4、T-08.5 | V-24、V-36 的候选分项 | 未开始 |
| T-08.7 | goal-34 | `web/admin` 前端骨架，配置表单与 i18 组件及 fixture | T-08 步骤 8（前端） | T-08.1 | V-25 的表单分项 | 未开始 |
| T-08.8 | goal-35 | configdb 示例与分层验证，汇总证据 | T-08 步骤 8 | T-08.6、T-08.7 | V-23～V-26、V-38、V-39 的 T-08 分项 | 未开始 |
| T-09.1 | goal-36 | 首次管理认证：admin-token init/reset、login、会话、CSRF、CLI Bearer | T-09 步骤 1 | T-07.5 | V-47 的认证分项 | 未开始 |
| T-09.2 | goal-37 | 安装包预检与安全解压：验签、来源、平台、长度/摘要、受控解压与上限、不可变包目录 | T-09 步骤 2（预检） | T-02 | V-27、V-28、V-41 的预检分项 | 未开始 |
| T-09.3 | goal-38 | 开发 CLI：`plugin init/check/build/test/pack` | T-09 步骤 2（CLI） | T-04.5、T-08.1、T-09.2 | V-01、V-41 的打包分项 | 未开始 |
| T-09.4 | goal-39 | 唯一管理服务与安装记录：逐插件串行、来源/pin/卸载保留、各管理动作的区分 | T-09 步骤 3 | T-09.1、T-09.2 | V-40、V-01 的安装分项 | 未开始 |
| T-09.5 | goal-40 | 正式候选 Journal 与准备编排 | T-09 步骤 4 | T-09.4、T-08.6、T-06.4 | V-29 | 未开始 |
| T-09.6 | goal-41 | 唯一提交点（ActiveGeneration+COMMITTED）与激活门禁 | T-09 步骤 5 | T-09.5 | V-30、V-21 的提交分项 | 未开始 |
| T-09.7 | goal-42 | 崩溃恢复、显式回退与引用回收 | T-09 步骤 6 | T-09.6 | V-30、V-31、V-36 | 未开始 |
| T-09.8 | goal-43 | Admin 页面：真实登录、状态/日志、安装启停、更新/恢复 | T-09 步骤 7（管理页） | T-09.6、T-08.7 | V-40、V-47 的页面分项 | 未开始 |
| T-09.9 | goal-44 | Admin 页面：配置表单与 JSON 编辑、命令权限、插件页面代理 | T-09 步骤 7（配置与插件页） | T-09.8 | V-12、V-22、V-23、V-25、V-43 的页面分项 | 未开始 |
| T-09.10 | goal-45 | 正式流程的故障注入，回填 T-08 待联验项，汇总证据 | T-09 步骤 8 | T-09.3、T-09.7、T-09.9 | V-23～V-25、V-27～V-31、V-39、V-40 | 未开始 |
| T-10.1 | goal-46 | V-01～V-47 证据汇总与跨组件联验 | T-10 步骤 1～4 | T-09.10 | 全部；真实 Telegram 需要测试账号 | 未开始 |
| T-10.2 | goal-47 | 五个目标组合的构建与运行、100 次启停 | T-10 步骤 5 | T-10.1 | V-32、V-34；需要真实目标平台 | 未开始 |
| T-10.3 | goal-48 | 资源比较报告 | T-10 步骤 6 | T-10.1 | V-33；需要固定机器与等价负载 | 未开始 |
| T-10.4 | goal-49 | 发布资产、签名与发布说明 | T-10 步骤 7、8 | T-10.2、T-10.3 | V-28、V-41 的发布复验 | 未开始 |

## 4. 逐任务实施工作包

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
4. 固定 SDK 的最小公共类型与可编译示例：一个根插件、两类 Handler、Context.StopPropagation、Config.OnLoad/Validate、互斥迁移接口、HTTP Handler 和调度声明。此处先证明签名能组合；T-02 起实现运行行为。
5. 用主契约选定的 pure-Go SQLite/GORM 组合完成临时数据库 CRUD/Tx、连接设置、WAL Backup 与恢复读取；验证 Up 与成功账本可在同一插件 DB 事务中提交。对独立 PG schema 验证连接、初始化和事务失败结果。验证通过后锁定实际版本，不预填“最新版可用”。
6. 生成兼容 fixture，测试编码往返、枚举零值、未知必需能力、字段存在性、JSON/Protobuf 编码区分，以及 payload+error 同时返回。让 `go generate ./...` 可重复运行，生成结果稳定；运行构建和相关测试。
7. 对五个目标 OS/arch 验证所选默认驱动与契约包的 `CGO_ENABLED=0` 编译路径，记录每个组合。目标平台的进程启动、运行与清理由 T-02/T-10 完成，不能以交叉编译替代运行证据。

#### 可审查交付

- 已锁版本的模块和生成入口、可编译协议/Schema/SDK 类型、兼容 fixture 与示例。
- 数据库连接/事务/Backup 验证代码和真实结果，供 T-08 直接采用已验证路径。
- V-01/V-02 的静态契约结果，以及 V-26/V-32/V-37/V-39/V-41 在本阶段取得的部分证据；运行和安装语义分别交后续任务闭合。

#### 停止和失败条件

契约能编译、示例签名成立、生成稳定、默认数据库组合的关键事务/备份路径通过，才向 T-02 交付。字段无法表达、驱动无法满足候选复制或目标构建失败时，先定位并修正该具体缺口。若 PG 测试环境尚不可用，可推进已通过的 IPC 工作，但 PG 分支和发布验收保持未完成，不能用 SQLite 结果替代。

### T-02：Core 状态、Native 进程、双 IPC 与基础生命周期

#### 前置输入

- T-01 已交付可编译的协议类型、控制帧 fixture、manifest/ID 校验规则，以及选定并锁定的 Go、Protobuf 和 Core SQLite 驱动版本。
- 按 [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) §1、§4、§6.1、§7.1、§12 和 [WIRE-FORMAT.md](WIRE-FORMAT.md) 实施。状态、帧号和错误码直接复用协议定义。
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
7. **完成阶段超时、停止和有限恢复。** 按主契约落实握手 10s、配置 60s、迁移 15 分钟、OnStart 60s、OnReady 30s 的独立预算；用户函数不响应 context 时仍须由 watchdog 结束失败进程。停用先冻结新根调用并最多排空 6s，允许既有父链内合法子调用完成；随后 Cancel 并发送 STOPPING，其期限为 `min(t0+10s, 发送时刻+4s)`。OnStop 使用独立 context，Core 在总 10s 截止处理进程树；Core Shutdown 并行停止所有插件。EOF 使旧 Session 立即失效，确认旧进程退出后才允许重启；5 分钟内第三次异常退出自动禁用，正常退出不自动重启。
8. **建立真实进程的验收切片。** 最小示例覆盖正常启动/退出，测试辅助进程覆盖挂起、错帧、日志污染、迟到响应和突退。对同一 fixture 分别运行 TCP 与 stdio，检查进程和 Listener 已关闭；注入满额度、双 Core、并发管理操作和 Core 被终止场景。记录实际运行平台、命令、测试位置及结果，并把配置/DB/更新尚待后续集成的场景明确标出。

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

#### 停止和失败条件

本项范围内的协议、身份、门禁、锁、取消和停止用例通过，最小进程切片可以重复运行，且后续配置/数据库/安装模块接入位置清楚后即可结束 T-02。若出现双代次接单、旧写入者未退出即启动替代实例、满额度使控制帧永久阻塞、丢失有效错误结果或损坏状态库后默默初始化，必须先修复，不能继续把该基础作为已完成依赖。当前缺少某目标平台的真实运行环境时保留“未运行”，由 T-10 补齐；构建记录不能代替该平台进程锁和退出测试。

### T-03：标准消息、Telegram 与可运行的发送 SDK

**目标：** 在 T-02 的真实 Session 与激活流程上，运行能够接收消息、回复、主动发送和处理媒体/按钮的 Native 插件；平台发送失败后仍能判断已确认结果。完整命令生成、Rule 与 Hook 编排在后续工作包接入同一实现。

#### 前置输入

- T-01 已交付协议 1.0、Event/Target/Segment/SendResult/媒体载荷的生成类型及序列化 fixture；字段以 WIRE-FORMAT 与 MESSAGE-MODEL 为准。
- T-02 已能启动独立插件进程，完成鉴权、Bootstrap、显式静态描述注册、激活、REQUEST/RESPONSE、取消及停止；Core 状态库与 Session/Generation 身份可用。
- 读取 CONTRACTS §1.3～1.4、§8，MESSAGE-MODEL §1～4、§6。使用本地 Bot API 测试服务覆盖可控失败；真实 Telegram 实测使用实施环境已经授权的测试账号和会话。

#### 实施步骤

1. **交付首个可用 SDK 切片。** 在最终 `sdk/` 中实现标准 Event、Target、Chain 构造、手动 Handler 的调用适配、Context、Send 和媒体客户端；复用 T-02 Session 与 ready gate。通过显式传入实例/函数及 RouteDescriptor 构建最小 echo 插件，先采用返回 `error` 的手动发送签名。该插件必须经过真实 REGISTER/ACTIVATED 后才收业务；不依赖尚未实现的注解生成器或命令树。接口直接使用正式生成类型，不建立另一套临时协议；T-04 在其上增加声明发现和构建检查。

2. **落实 Target 与最小授权路径。** Core 保存真实入站的完整 Target 和可信来源；原事件调用及合法后续父链只可默认回复该 Target。后台发送必须显式 To，其他 Target 与后台调用须匹配 Core 状态库中该插件的 `allowed_adapter_instances`，其初值为空。发送准入逐次检查实例启用、account 对应关系、当前许可和全局禁止，不信插件自报 Actor/Admin；没有默认目标返回 NO_DEFAULT_TARGET。提供 Core 内部策略读写入口供测试及后续管理面复用，避免为本阶段开放未鉴权的管理接口。基础身份与授权随首个 Adapter 落实，T-04 扩展命令继承，T-05 集成完整事件权限链。

3. **接通 Telegram 接收与生命周期。** 实现 getMe 的账号/用户名绑定、同账号唯一 owner、已有 webhook 冲突检查、后台 long polling 和状态上报。显式设置规定的 allowed_updates；只生成平台实际给出的事件，`sender_chat` 的兼容 fake sender 不构造个人 Actor。每条 update 先持久化接纳游标再投入非持久队列；满队列丢弃计数且继续推进，不能提前提交整个 batch 的最大 ID。实现契约规定的连续 7 天空闲接收轮次重建。短暂读失败使用有界退避，401/409 停止该实例并报告故障；Start、Stop 和重连均不得创建重复接收 owner。入站媒体组保留 media_group_id，逐 Message 派发。

4. **实现同一 Core 媒体暂存与授权。** 接通 BeginUpload→顺序 Chunk→Commit，验证大小、MIME、SHA256，执行 1 MiB 分块、100 MiB 单对象、4 个未完成上传、每 Session 256 MiB、Core 2 GiB 及 15 分钟期限。入站只签 IncomingMedia 描述 token；首次 Read 才通过受信平台文件 API 获取完整对象、校验并缓存，随后按 offset/length 读取。多 Session 获得独立 token、共用一份缓存；上传者读自身对象及真实入站 token 的读授权不依赖主动发送许可。Grant 只给符合契约范围的 Activated 接收者签新 token；共享不会放宽 Send 的目标权限。配额预留、owner 退出、引用过期和消费租约的释放归同一对象管理，不能先返回可用引用再发现下载失败。

5. **实现可预测的发送计划及回执。** 每条 Chain 先完成全部结构、目标能力、媒体可用性与平台限制检查，再生成按顺序执行的物理操作列表。落实文本/实体 UTF-16 偏移、单媒体/2～10 项相册、首条引用、末条按钮以及相册后的 markup 更新；不静默截断、重排或降级。每个操作记录 PartResult，明确 ACCEPTED、REJECTED、UNKNOWN、NOT_ATTEMPTED，并据事实汇总逻辑状态。遇明确拒绝或结果未知停止后续操作；保留已确认 ID，同时通过 Envelope 返回有效 payload 和错误。平台确认但 message_id=0 时不伪造可编辑 ID；依赖该 ID 的步骤未尝试并报告 PARTIAL。显式空 Chain 和仅有元数据的 Chain 在调用平台前拒绝。

6. **接通按钮 token 与明确的回调路由。** 使用显式注册的 button.clicked 路由验证 action/schema/value；Core 生成 32 字节随机值的无 padding base64url token，持久化绑定插件、路由、Generation、Target 和 24 小时期限。点击先尽力 answerCallbackQuery，再验证 token、当前 actor/策略及目标，使用普通 REQUEST 调用正确 Session。多用户点击逐次鉴权；普通同代次重启保留 token，卸载/切换代次永久失效。拒绝不能落入 Raw/Regex，也不能发额外群错误消息；缺少可核验普通 Target 的 inline 回调明确拒绝。

7. **运行发送与接收故障矩阵并记录证据。** 构建 echo/媒体示例以及本地 Bot API 测试服务，验证两个账号/实例、话题、多物理消息、提交后超时、相册后 markup 失败、零 ID、媒体过期/共享/配额、游标提交后崩溃和接收冲突。以调用记录证明失败后没有自动补发，以平台请求的实际数量证明一条 Chain 与两次 Send 的区别。跨插件媒体授权先用两个真实 Session 和显式测试调用验证，不等待 T-06 Service。记录运行命令、实现变更、模拟/真实平台版本和日志；真实 Telegram 尚无可用测试环境时单独标「未运行」，不得用模拟结果冒充平台实测。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `contracts/` | 复用生成的消息/媒体/回执类型与 fixture；发现载荷不足时先修正同一契约，不新增平行 wire |
| `sdk/` | Context、Chain、手动 Handler、Send/media 客户端与 ready gate 的实际用法 |
| `internal/core/message/`、`internal/core/media/`、`internal/core/policy/` | 发送准入、回执汇总、对象/引用/租约、实例许可与按钮映射的状态操作 |
| `internal/adapters/telegram/` | 接收 owner、游标、标准化、媒体取件、发送计划及 Bot API 故障测试 |
| `examples/plugins/echo/`、`examples/plugins/media/` | 能通过 T-02 实际启动的显式注册插件；后续迭代保持既有行为 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-07 | 原 Target 回复、显式目标、无许可/撤销、实例账号与话题隔离、NO_DEFAULT_TARGET |
| V-08、V-35 | 完整预检与物理顺序；四类逻辑结果、部分结果、零 ID、payload+error、无自动重发 |
| V-09、V-44 | 上传校验、惰性 Read、授权 Grant、多 Session 共用缓存、配额/过期/租约与失败清理 |
| V-45 | owner/webhook/401/409、订阅范围、匿名发送者、逐条游标和 7 天轮次边界 |

T-04 的启动条件是本包 SDK、Adapter 和受控故障集成可运行且相关安全边界通过。真实平台验证尚未运行的部分继续保持未完成，由 T-10 汇总补齐；它不要求先实现 T-04 的生成器。完整普通事件排序、自动回复、Hook 和命令策略组合分别在 T-04/T-05 闭合。

#### 失败停止条件

- Target/账号凭证串用、默认实例许可意外放开、IncomingMedia 读取暴露 Bot Token/路径或 Grant 扩权，阻止本包合入。
- 平台已成功却被报告为完全未发送、部分结果丢 ID、发送前已知不支持却先产生消息、失败后自动重发，必须先修复发送管道。
- 单个 Handler 阻塞 Session 读写、媒体缓存/队列超过固定预算、接收游标可能重放已接纳业务或重复启动 owner，停止主链路集成并回查 T-02/T-03 的所有权与准入实现。
- 碰到平台无法表达的既定能力，记录 UNSUPPORTED 和具体受影响验收；不能用静默降级替代通过，也不能在本包扩展新平台功能。

### T-04：根插件、构建期注册与命令策略

**目标：** 将 T-03 的可运行 SDK 变成可以稳定构建、注册和管理命令的插件开发入口，保留显式注册方式；策略变更无需重启插件。

#### 前置输入

- T-03 的正式 SDK、手动 Handler 示例、发送通路及原事件目标授权可用；T-02 的注册验证、状态库和激活流程可复用。
- T-01 的 manifest、稳定 route ID、RouteDescriptor、Schema 与生成 fixture 可用。读取 DESIGN §D/§E，以及 CONTRACTS §2、§9 的当前注册和命令规则。
- 本包实现配置类型的唯一性/构建描述检查；完整 OnLoad、Secret、DB 候选流程由 T-08 接入，不以 T-08 已实现为前置条件。

#### 实施步骤

1. **完成根插件 API 与实例绑定。** 在现有 SDK 上实现恰好一个根类、嵌入 plugin.Base、至多一个 Config 的发现/验证；插件 ID 只来自 plugin.yaml。保留显式传入根实例或辅助接收者的注册接口，不建立自动 DI。生成器与显式适配器校验两种 Handler 签名及参数类型；执行时的完整 Rule/传播/自动回复组合由 T-05 完成，不另造第二套 Handler 返回协议。

2. **实现限定范围的构建期生成。** 只扫描当前插件在所在 module 中选定的目标构建包及其本模块依赖包，遵守 build tags、GOOS/GOARCH，排除依赖、vendor、测试文件。包级 command/event 标记生成 Go 注册代码，由编译器检查函数类型；实例方法必须通过已提供的实例显式绑定。仓库内各示例以各自构建入口限定扫描范围，不能全仓库扫描所有示例后误报多个根插件。产出稳定描述、schema 引用和构建诊断；同输入重复生成结果一致。通过 `cmd/alonabot/` 的构建检查入口调用同一生成逻辑，避免 CLI 和 SDK 各自维护一份扫描器。

3. **建立静态注册索引与诊断。** 合并生成及显式描述，通过一次 REGISTER 提交；同插件重复 route ID、规范化命令路径或别名拒绝注册。Core 为已经实现的命令/事件描述建立索引，保留后续 Hook/Service 共用的描述及校验入口，不提前开放其未实现的执行能力。跨插件同路径记录唯一胜出者和遮蔽者，按 priority/plugin_id/route_id 稳定确定，与启动次序无关。Session 内注册固定，运行时策略单独读取。

4. **实现命令选择和参数绑定。** 只对 message.created 的 `/` 命令匹配完整 token；Telegram @username 按已验证的当前 bot 处理。先最长普通路径/别名，再 Regex，最后 Raw。将「已识别普通命令」保存为分发状态，禁止禁用、权限拒绝、参数错误后转入更短路径或其他类别。按契约实现引号、反斜杠、int64、有限 float64、true/false、必需/可选/默认与末位 raw 尾部；Regex 启动编译，捕获结果独立于参数。对允许执行但参数非法的命令生成一次 usage 或采用声明的 silent；权限拒绝不自动发群错误。

5. **实现持久化命令策略。** 以 plugin_id/route_id 为键记录权限和禁用覆盖，复用 T-02 状态库以及 T-03 的实例许可入口。权限取最近显式子项，否则继承；插件、祖先组及自身任一禁用都拒绝，最后与 Core 全局禁止约束合并。每次派发前读取一致的当前策略；已开始调用不宣称被撤销。提供管理服务层的查询/修改/遮蔽诊断方法，持久化成功后才报告变更成功，供 T-09 鉴权后的管理 API/WebUI 使用。

6. **用同一个 SDK 升级示例。** 将 echo/媒体示例增加命令组、子命令、别名、Regex、Raw、可选参数和显式实例方法，同时保留一条纯显式注册路径。验证生成产物实际编译、启动、注册、接收命令并调用 T-03 Send；不以只比较生成文本代替编译。对 T-05 才执行的自定义 Rule false/error，先用 HandlerOutcome fixture 验证命令消费状态机，后续补真实 SDK Rule 集成证据。

7. **完成构建和策略验收。** 使用独立测试插件或构建 fixture 验证多 root/Config、错误方法签名、依赖中的伪标记、不同 build tags、重复路由和参数溢出；以两个实际 Session 验证交叉同名命令与禁用保留路径。直接调用 Core 管理服务层修改策略，确认下一次派发生效、进程未重启、Core 重启后保留；为后续 WebUI 提供同一方法的输入、输出和失败条件。将静态解析证据与待 T-05/T-09 集成的用例分项登记。

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
| V-12 | 策略服务层的即时生效、持久化和全局禁止优先；实际 WebUI 操作由 T-09 完成 |
| V-43 | 父禁用、子权限覆盖、Core Admin/chat_admin 分辨及下一次派发生效；完整多事件流程由 T-05 再验证 |

V-12 是跨任务验收：T-04 不依赖尚不存在的 WebUI，但必须交付后续 UI 实际调用的策略服务层。矩阵中保存分项证据；只有 T-09 的鉴权后浏览器操作也完成，才将整项 V-12 标为通过。T-04 完成后，T-05 和 T-08 可以分别使用同一注册器继续实施。

#### 失败停止条件

- 生成器扫描外部依赖或隐式构造接收者、同一源码因启动顺序而选择不同路由、重复 route 被静默覆盖，阻止生成器/注册器合入。
- 已识别但拒绝的命令落入低优先级同名项、短命令或 Regex/Raw，先修复分发状态，不用示例避开冲突。
- 子项启用绕过父禁用、平台 chat_admin 获得 Core Admin、策略重启后丢失或需要重启插件才生效，阻止策略功能完成。
- 为了测试 UI 而增加临时无鉴权管理端点，或为未来配置迁移建立与 T-08 相竞争的运行时，停止该实现并复用既定服务入口。

### T-05：普通事件分发、Rule、Hook 与可信权限链

**目标：** 组合 T-03 发送管道与 T-04 路由，形成可解释的入站处理顺序、两类 Handler、传播停止和 Hook 结果；为 T-06 保留完整受信上下文。

#### 前置输入

- T-04 注册器、命令消费状态、策略服务层以及 T-03 Adapter/Send/媒体已可运行。
- T-02 能根据在途父请求保存并验证 Session、actor、target、trace、deadline，子请求取消规则可用。
- 读取 CONTRACTS §2、MESSAGE-MODEL §3～5、WIRE-FORMAT 的 HandlerOutcome/Hook 载荷。T-06 Service 与 T-07 Scheduler 均不是本包前置；其专属集成场景明确交回后续任务。

#### 实施步骤

1. **收口可信事件上下文。** 将 Adapter 原始输入转为 Core 保留的身份、Target、来源消息 ID 及路由绑定，业务 payload 与这些保留信息分开。确认只有可信个人 Actor 才满足 User，Core 实例作用域管理员才满足 Admin；群管理员仅供 chat_admin Rule。没有 Actor 的事件和后台调用不补用户。按钮 action/schema/value 与 scheduled job/run 等已验证绑定不随可修改 payload 改变；scheduled 用规范 fixture 验证即可，实际 tick 由 T-07 接入。

2. **实现有界的同事件分发。** 接收入口统一执行 Normalize→PreEvent→命令/类型匹配与 Core 权限→SDK Rule/Handler。普通事件每条链顺序执行，跨事件最多 64 条并发；等待队列不超过 256 条且 8 MiB，并计入已有 Core payload 总预算。候选按 priority/plugin_id/route_id 稳定排序；入站超额明确拒绝/丢弃并计数，接收游标依 T-03 规则处理。链执行从 Session 读循环分离，慢 Handler 不能停止 Response/Cancel/控制帧处理。

3. **完成 SDK Rule 与 HandlerOutcome。** 自定义 Rule 与目标 Handler 在同一 RPC 中执行；默认 AND，显式 OR 通过 SDK 组合，false/error 均不运行目标函数。适配 `error` 与 `(Chain,error)` 两个签名，返回明确 matched/propagation；只有自动模式成功返回非空 Chain 时触发一次 Send，自动模式中的显式 Send 拒绝。手动模式允许多次 Send。panic 转 INTERNAL 并保留诊断，不吞成成功；Context 返回后失效。

4. **固定传播与错误分支。** c.StopPropagation() 在 SDK 锁存，结果即使同时携带 ErrorInfo 也保留 Stop；普通事件 Rule false 跳过、业务错误默认继续，显式 Stop 终止后续候选。已识别命令则沿用 T-04 的消费状态，不因 Rule false/error 回落。测试 Handler 已发送后失败的情况，确保不会补发、撤销已确认消息或自动发送错误回复；后台 goroutine 不得修改已返回结果或继续借用失效父调用。

5. **接入三个 Hook 阶段。** 使用独立 route kind 和稳定排序运行 PreEvent、BeforeSend、AfterSend。PreEvent 只接受允许的 payload/传播变化，拒绝覆盖可信身份、Target、来源与按钮/任务绑定；失败结束本次入站链。每次显式或自动 Send 独立执行一组 BeforeSend→Adapter.Send→AfterSend；修改后的整条 Chain 重新预检，拒绝不调用平台。AfterSend 只观察最终回执/错误，其失败记录诊断而不改写已确认结果。

6. **落实出站递归约束与父链继承。** Core 在 BeforeSend/AfterSend 的受信执行上下文设置 send_forbidden；通过直接 Send 验证拒绝。该标记与 actor、target、scope、deadline 同存于 T-02 父调用记录，并提供给后续 Broker 使用，不能依赖插件自报。当前任务不为验证间接递归而临时实现 Service；在接口 fixture 中证明派生上下文保留此标记，将真实 Hook→Service→Send 拒绝作为 T-06 的必交回归用例。

7. **运行实际多进程流水线。** 复用示例插件配置多个 priority、同优先级逆安装顺序、自定义 Rule、两种 Handler 与三个 Hook；逐项检查平台调用日志、候选调用次序及 HandlerOutcome。用恶意载荷尝试修改 Actor/Target/button/job 字段，用延迟 Handler 验证控制帧仍推进。将 V-11 的真实 Rule 消费、V-43 的权限组合回归与本包验收一起记录；仅使用 scheduled fixture 的地方标明尚未覆盖真实 Scheduler 执行。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `sdk/` | Rule 组合、两个 Handler 适配、StopPropagation 锁存、Context 生命周期及 Hook 适配 |
| `internal/core/router/`、`internal/core/policy/` | 同事件候选链、命令消费衔接、可信身份和逐次策略检查 |
| `internal/core/message/` | 三阶段 Hook 与现有 Send 管道集成、只读 AfterSend 结果 |
| `internal/core/` 的现有 Session/调用上下文实现 | 父链权限和 send_forbidden 的保存/继承接口；不建立平行 Session 管理器 |
| `examples/plugins/` 及所属包测试 | 可运行多插件分发场景与绑定/身份篡改 fixture |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-13 | 不同启动/安装次序下候选调用顺序稳定 |
| V-14 | Rule false/error、两类 Handler、Stop+error、自动型 Send 拒绝、自动回复恰一次 |
| V-15 | 三阶段 Hook、不可改绑定、直接递归 Send 拒绝、AfterSend 不改成功；真实 Service 间接路径由 T-06 闭合 |
| V-16 | 无 Actor fixture、伪造管理员、实际支持的按钮/编辑事件与不支持能力；真实 scheduled 执行由 T-07 补验 |
| V-11、V-43 | 将 T-04 的命令消费及权限策略接入真实 SDK Rule/事件链，完成相应组合回归 |

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
- 读取 CONTRACTS §1.2、§3、WIRE-FORMAT §4～5，保持动态命令/事件 Router 为 v0.2。
- 准备 A/B/C 三个真实插件进程的测试入口和可控阻塞点，不用进程内函数互调冒充 Service E2E。T-07 的调度、HTTP 与 T-09 的管理 UI 不作为前置。

#### 实施步骤

1. **补全 Service/Bus 描述与 SDK 入口。** 通过现有 REGISTER 提交 service_name、输入/输出 schema/version、编码、调用规则及订阅事件声明。支持既定 UTF-8 JSON 和 Protobuf 编码，固定随包 schema 资源；Core 校验标识、版本、大小和 JSON 语法，SDK 用生成类型/校验器检查业务字段。拒绝运行时网络 `$ref`、未声明编码、未知版本及编解码失败，空结果仍返回声明的空对象/消息。

2. **实现 Broker 定位与准入。** 以 plugin_id/service_name 定位当前 Activated Session，检查默认 Core/本插件规则以及显式 allowlist/public；public 不赋予用户 Admin。目标不存在、未激活、未授权或输入不合规时在执行目标函数前返回明确错误。把每一段 Plugin→Core 与 Core→Plugin 请求单独编号，保存实际关联，不依靠跨 Session 相同 CallID。

3. **沿父调用传播真实上下文。** parent_call_id 只接纳同 Session 仍在途的 Core→Plugin 请求；从 Core 记录继承 actor、target、scope、trace、祖先插件链和 send_forbidden，不接受插件覆盖。无父链的后台服务只有插件身份。每次追加插件检查同链重复和深度 16，同链重复返回 CALL_CYCLE，深度超过 16 返回 FAILED_PRECONDITION；独立并发链各自记录。服务中向原事件 Target Send 仍经 T-03 准入；媒体跨 Session 仍须 Core Grant，不能把原 token 当公共对象地址。

4. **接通跨段取消和背压。** 子调用 deadline 取父与自身操作预算较早者；父完成/超时/取消或任一 Session 断开，结束对应结果并取消未完成子调用，不重放。复用 T-02 的 256 在途、16 MiB/方向、Core 64 MiB 预算，额满立即失败。Broker、SDK 与各 Handler 不在等普通业务槽位时互相占锁；Response/Cancel/控制保持进展，尚未写出的 Request 被取消即撤出，Cancel 不超越其 Request。普通迟到响应不二次完成调用。

5. **实现独立的非持久广播。** Publish 使用普通带 CallID 的请求，仅确认 Core 完成发布准入；所有权限及命名空间检查在分发前完成。按订阅注册的事件名/schema/version发送无业务回执 EVENT；每个订阅者最多 256 条且 4 MiB，满时 drop-newest 并计数，离线不缓存、不补发。广播回调使用订阅插件身份，不沿用发布者的用户权限；一个慢订阅者不得阻塞其他订阅者或 Publish。广播不参与普通事件的 StopPropagation，也不伪装成 scheduled 请求。

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
- 按 [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) §4.3、§7.2～§7.4 和 [WIRE-FORMAT.md](WIRE-FORMAT.md) 的 REQUEST、注册及 HTTP Listener 契约实施。
- T-09 负责首次管理 token、离线 reset、真实登录/CLI 和 WebUI。本项先完成授权检查与插件票据接口，用测试会话验证组件；实际入口在没有有效管理会话时必须拒绝，不能为了提前演示而开放临时免认证路径。T-09 再完成真实登录后的端到端授权验收。

#### 拟涉及目录

`sdk/`（scheduled callback 包装、HTTP 注册和 Listener）、`internal/core/` 下的 Scheduler、HTTP 路由/代理与插件会话实现、`cmd/alonabot/`（Core 公共 Listener 和 Base URL 装配）、`examples/plugins/`（定时与 HTTP 示例）。继续使用单 Go module；`web/admin/` 的管理页面由 T-09 交付，本项不新建另一套管理前端。

#### 按顺序实施

1. **接入并校验注册描述。** schedule 注册固定 job_id、interval/cron、timezone 和 job deadline，HTTP 注册 method、插件相对 path、公开/私有及回调验证模式。复用既有完整 REGISTER 校验；同插件重复 ID/HTTP 路由、无效表达式和不合法 endpoint 使整次注册失败。路由描述准备完成不代表任务已启动或代理已公开。
2. **实现下一次触发计算。** interval 至少 1s，首个 tick 为激活时刻加间隔，后续沿固定起点计算；cron 固定五字段，默认 UTC，显式时区必须为有效 IANA 名称。按既有日/星期 OR、其他字段 AND 的规则求下一真实时间点，处理夏令时缺失和重复时刻。激活、恢复或落后时只安排下一未来 tick，不持久化漏 tick，不补跑。触发计算接受明确的当前时间，便于验证时间边界。
3. **实现实际运行槽与调度调用。** tick 使用 Core→Plugin REQUEST 触发 callback，无用户 Actor、不默认 Admin；每次采用独立 run_id 和 15 分钟 job deadline。不同 job 可并发，同一 job 有实际运行记录时跳过本轮。超时先结束调用结果并取消，但保留运行槽和原调用键；SDK 在用户 callback 真正返回后才发送唯一 Response，迟到响应只释放槽。RPC 断链不能释放仍存活进程的槽，必须等真实完成响应或确认旧进程退出；恢复后仍从未来 tick 开始。
4. **实现 SDK 托管 HTTP Listener。** 仅存在 HTTP 路由时绑定 `127.0.0.1:0`，保持 Listener 打开并在 REGISTER 上报实际端点；插件作者只提供 Handler。用户 Handler 外包 Session 认证检查和就绪门禁；准备阶段即使端口已监听也不执行业务。Listener 与 Session 生命周期绑定，停止、错误和激活失败均关闭，不增加插件端口配置。
5. **实现 Core 的 Session 代理映射。** 校验注册端点为合法 loopback 后保存，只有当前 Activated Session 能接单。公共前缀固定 `/plugins/{id}/...`，Core 匹配插件、method、route 和授权，剥掉插件前缀后交给标准库 ReverseProxy.Rewrite。覆盖内部认证头，清除 admin Cookie/Authorization 及未验证的转发身份信息，保留 Core 路径。公开 OAuth/Webhook 只对显式声明的路由生效，示例与测试覆盖插件执行 state/签名/nonce 验证；回调 URL 只从受控 Base URL 构造。
6. **完成两个 Host 的页面边界与插件授权。** 同一公开 Listener 按受控 Host 分配 Admin 和插件内容，两个 Base URL 使用同 scheme/有效端口、不同 host；插件 host 不提供 Admin API。实现 60s 一次性插件兑换票据、独立插件 Cookie、原管理会话失效后的派生授权撤销，以及票据 URL 清理和 no-referrer。管理会话校验接 T-09 的认证入口，T-07 测试注入的会话只存在于测试装配。禁止向 plugin Origin 开放管理凭证 CORS，限制插件 Set-Cookie；未配置第二 host 时保留 JSON API/Webhook，按契约拒绝主动页面并施加响应策略。
7. **接入统一启停和清理。** 在 T-02 的 Activated 转换后启用 Scheduler 和代理；Draining 时停止新 tick 和新 HTTP 派发，停止阶段取消余留调用、撤销映射、关闭 Listener。复用全局 10s 停止预算，不能为 Scheduler、HTTP 和每个插件另起一轮独立预算。Session/Generation 变化后，旧 endpoint、凭证和运行记录不能继续接收新业务；HTTP 失败不得触发自动业务重试。
8. **执行时间、进程和 HTTP 边界验收。** 用给定时间验证 cron/interval/夏令时，用真实插件进程验证不合作 callback、迟到 Response 和断链后的运行槽；用定时与 HTTP 示例验证作者不管理端口。HTTP 测试覆盖未激活、停用后请求、伪造内部头、私有路由、公开回调、两个 Host 和单 Host 响应限制。分别记录 T-07 组件结果与仍需 T-09 首次认证后完成的端到端结果。

#### 可审查交付

- Scheduler 注册、下一 tick 计算、运行槽和 callback 包装的代码；能展示不同 job 并发以及超时后同 job 不重叠的示例和测试记录。
- SDK HTTP Listener、Core Session 代理映射、受控 Host 分流、插件票据/会话和停止清理代码；示例只声明 Handler，不声明内部端口。
- 可查阅的授权接口及 T-09 接入点、默认拒绝行为和测试装配；公开/私有路由与管理 Origin 的验证结果分开记录。

#### 对照现有验收

| 现有条目 | T-07 要完成的验证 |
| --- | --- |
| V-20 | 不同 job 并发、同 job 跳过重叠、cron/interval/IANA 时区/夏令时、离线和落后不补跑 |
| V-46 | 超时后 callback 未结束、迟到 Response、断链但进程仍存活时，实际运行槽的保留与释放 |
| V-21 | Scheduler/HTTP 服从就绪顺序、失败隔离与 Core 多插件共享的停止预算 |
| V-22 | 自动 loopback Listener、无需插件端口配置、Session 代理、公开/私有路由和停止清理；私有页面与真实管理员登录的完整连通在 T-09 联验 |
| V-47 | 两个 Host、内部头/管理凭证隔离、插件兑换票据和单 Host 主动内容限制；首次 token、reset、真实登录/CLI 与原会话撤销在 T-09 完成端到端验收 |

#### 停止和失败条件

定时与 HTTP 组件完成上述验收，正常运行和停止都复用 T-02 的状态机，且 T-09 只需接入真实认证和页面即可集成时结束本项。若同 job 在实际 callback 未结束时再次启动、离线 tick 被补发、准备阶段公开路由、旧 Session 仍能代理、插件取得管理凭证或运行主动内容于 Admin Origin，必须修复后再交付。授权入口尚未接通时保持拒绝并标注联验未完成，不以测试会话宣称首次登录已可用；不为完善本工作包加入多用户账号、每插件独立域名或持久任务队列。

### T-08：配置、Secret 与数据库候选准备

#### 前置输入与边界

直接依赖 **T-04**，并复用其依赖链已交付的 T-01/T-02；可与 T-05 后续工作并行，不等待 T-06/T-07 或 T-09。

| 前置输入 | 开始本任务前须能取得的成果 |
| --- | --- |
| T-01 契约与依赖验证 | 可编译的 Bootstrap/REGISTER 类型、Config Schema 与 manifest 描述；锁定版本的 GORM、SQLite/Backup API、PG 驱动；真实 CRUD/事务/备份小验证。 |
| T-02 基础运行时 | Core 状态库存取、Generation/Session 标识、逐插件操作锁、`core.lock`/`runtime.lock`、阶段预算与准备期业务门禁；可在临时数据目录启动测试插件并观察失败结果。 |
| T-04 SDK/生成器 | 单 root/至多单 Config 的发现与注册、构建期类型检查、可扩展的配置字段描述；不重复制作第二套扫描器。 |
| 现行规格 | CONTRACTS §5/§6/§12、WIRE Bootstrap 与 REGISTER、TEST-MATRIX V-23～V-26/V-38/V-39；迁移接口、JSON 语义、默认驱动和 PG 维护分支均沿用原契约。 |

**本任务交付候选准备能力，不提前实现 T-09 的正式 Journal 和 Active 提交编排。** T-02 只有基础代次/会话原语，不能据此宣称已经完成升级恢复。T-08 对外提供 Pending 导入、候选配置/DB 准备、资源引用与清理能力；T-09 决定何时停止旧版、创建正式操作记录、提交 Active、恢复和回收。测试驱动可以调用这些能力，但不能成为第二套生产激活入口。

#### 实施动作

1. **完成配置描述与构建检查。** 在 T-04 生成器上实现 `json/form/description/secret/hidden` 字段元数据、Go 声明顺序、嵌套结构、受支持类型、必需字段、默认值和 i18 路径。为不支持的自定义编解码、递归类型、错误默认值和越界数字给出明确诊断；生成资源与本次构建的 Schema 一致。准备覆盖可空值、map/slice、secret 和 hidden 的最小示例，不扩展字段类型范围。
2. **实现完整 JSON 与 Pending 编辑能力。** 以原始对象保留嵌套未知键，区分缺失键与显式 `false/0/空串/空集合/null`；实现完整替换、RFC 7396 Merge Patch 和编辑快照导入。提交检查 `base_generation`，旧基线拒绝；直接编辑 `config.json` 不更改 Active。提供供 T-09 管理服务调用的应用层入口和结构化错误，不在此阶段开放未经认证的独立 HTTP 写接口。
3. **接入 SDK 配置准备顺序。** 使用 T-02 的私有 Bootstrap，按 `OnLoad(raw, fromVersion)` → 仅补缺键默认 → typed decode/Schema 校验 → `Validate` 处理，再进入该启动原因允许的 DB 准备。完整规范 JSON 随 REGISTER 回传，禁止用 typed marshal 替换原始文档；新 Schema 检查必须在 OnLoad 之后。普通 restart 重验原 Active，相同配置版本的 OnLoad 若改变运行配置则失败；所有阶段错误、超限及超时走既有控制错误，准备期不能发普通业务 RPC。
4. **实现不可变 Secret Store。** 新 secret 值生成新引用，Core 在所属 Session Bootstrap 前解引用，在 REGISTER 后按新 Schema 和保留的旧 secret 路径重新抽取；SDK 不计算持久化引用摘要。检查引用归属，管理读取返回引用/设置状态，掩码不成为保存值；日志、错误和普通响应不含明文 JSON、DSN 或 secret。实现运行账号权限/Windows ACL 与按引用清理能力；清理必须保护 Active、Pending、保留代次及恢复中操作引用，正式 Journal 引用由 T-09 接入。
5. **实现原生 GORM 与连接生命周期。** `gormdb.Open(ctx)` 按 Session 固定后端并复用连接池，SDK 停止时关闭；插件不可取得 Core 状态库连接。按契约应用并核对 SQLite WAL/FULL/外键/5s busy timeout 与默认单连接，事务回调使用传入的 tx；PG 只接收管理员预置的插件 role/schema 凭证并校验指定 schema/search_path。连接、权限或数据损坏时返回错误，不能创建空库或自动换后端掩盖问题。
6. **实现两种互斥迁移与 PG 门槛。** SDK 检查 `Migrate`/`Migrations` 互斥；versioned 模式校验非零、唯一、递增 ID，每步 Up 与 `_alonabot_migrations` 完成记录使用插件 DB 的同一事务，失败不记成功。PG `resume` 只核对已提交代次及账本/身份，不重新执行 Migrate/Up；首次空 schema、无待迁移且无副作用的自动候选、显式维护按各自分支执行。维护许可须匹配 Core 下发的本次 Session/Generation/DB；事务不支持的 SQL 不得偷偷转为事务外执行。
7. **形成可交接的候选资源结果。** 在已持逐插件操作锁、旧写入者退出且 Core 可取得 runtime.lock 的前提下，使用 Backup API 创建独立 SQLite 候选；首次安装创建空候选，reload/update 同样复制，普通 resume 复用 Active。复制完成后 Core 释放临时 runtime.lock，再由 SDK 取锁打开候选，避免互相等待。将准备好的程序/配置/secret/DB 引用、配置摘要、迁移完成状态与本次身份交给 T-09；本任务不写 Active 指针或伪造 COMMITTED。失败清理只处理未受保护的候选资源。
8. **完成示例与分层验证。** 在临时 data_dir 和独立测试 PG schema 上，运行配置迁移、显式空值、未知 secret、CRUD/事务、迁移失败以及候选 OnStart 写 DB 后失败的案例。为自动表单提供字段顺序、hidden、i18 回退、掩码和 PATCH 的组件/fixture 验证，真实管理页面在 T-09 接入。记录可独立通过的组件证据，以及仍需 T-09 正式管理入口和 Journal 才能关闭的场景，不把后者标成已通过。

#### 拟目录与交付

| 实施拟建或扩展路径 | 本任务职责 |
| --- | --- |
| `contracts/`、`sdk/config/` | 复用 T-01 描述，配置回调、默认值、原始对象保留、类型校验及对应测试。 |
| `sdk/gormdb/` | Session 连接池、原生 GORM 入口、迁移执行器、后端集成测试。 |
| `internal/core/config/`、`internal/core/secrets/` | Pending/候选配置、引用归属、不可变 secret、脱敏读取与受保护引用清理。 |
| `internal/core/storage/` | 受管 DB 描述、SQLite Backup/候选文件准备、PG 准入检查；复用 T-02 状态/锁原语，不另建状态库。 |
| `web/admin/` | 配置 Schema 消费、表单/i18 组件和 fixture；正式登录及管理 API 绑定属于 T-09。 |
| `examples/plugins/configdb/` | 使用真实 SDK 的配置与数据库示例；失败点由测试夹具控制，测试数据不使用生产凭证。 |

交付包括可编译实现、生成资源、示例、测试夹具，以及一份写在代码/测试附近的 T-09 调用约定：输入身份和旧代次、锁与旧进程前置条件、输出资源引用、错误及清理所有权。该约定是既有契约的实现交接，不建立新的公开插件 API 或平行规格。

#### 对应 V 验收与停止条件

| 验收编号 | T-08 必须取得的证据 | 后续闭合 |
| --- | --- | --- |
| V-23、V-38 | 原始/未知键保留，空值/默认/Merge Patch 正确，陈旧基线拒绝，restart 不改 Active。 | 真实 `plugin reload` 与 WebUI 提交链路在 T-09 复验。 |
| V-24 | 独立候选 DB 中 OnStart 写入后失败，旧 DB/配置/secret 不变；锁阻止同时写入。 | Core 在正式候选流程中被 kill 后的权威代次选择，在 T-09 结合 V-30 完成，T-08 不提前关闭整项。 |
| V-25 | 新旧 secret 引用分离、未知 secret 仍保密、hidden/i18/掩码与 PATCH fixture 正确。 | 真实页面、再认证与管理响应在 T-09 复验。 |
| V-26、V-39 | 两后端 CRUD/Tx，接口互斥，Up+账本同事务，PG 首装/resume/维护许可有效性。 | T-09 补自动更新及显式维护的真实入口验证。 |
| V-21、V-37 的配置/迁移分支 | 使用 T-02 会话验证阶段错误/超限/超时无半注册、无准备期业务请求。 | T-09 验证与正式提交前后恢复门槛组合。 |

建立上述包后，后端验证入口至少包括 `go test ./sdk/... ./internal/core/config/... ./internal/core/secrets/... ./internal/core/storage/...`；PG 使用专属测试角色/schema，按 T-01 已落地的环境方式提供连接。前端使用实际建立的组件测试入口，记录准确命令。所有结果按 TEST-MATRIX 记录代码位置、命令、版本、环境和日志；尚无环境时明确保持未运行。

出现候选污染旧数据、未知键/显式值丢失、secret 越权/明文泄漏、迁移成功记录与实际事务不一致、PG 未许可写生产数据、无法验证驱动/备份正确性时，停止向 T-09 交付并修复对应实现。缺少 T-02 锁/会话原语时先回补 T-02，不在本任务用睡眠、PID 猜测、复制 WAL 主文件或临时直写 Active 绕过依赖。

### T-09：安装分发、正式激活恢复与管理入口

#### 前置输入与边界

直接依赖 **T-06、T-07、T-08**。T-02 基础状态、锁和 Session 能力已在前置链中实现；**本任务扩展 T-02 的同一提交事务原语，完成带候选资源引用的正式 Journal、Active+COMMITTED 提交编排及崩溃恢复**；T-02 的最小提交切片和 T-08 的资源准备不能替代这些集成工作。

| 前置输入 | 开始本任务前须能取得的成果 |
| --- | --- |
| T-01/T-04 构建契约 | manifest/Schema/注册描述解析器、生成器、构建输入约束和已锁定工具链；运行期显式注册的完整校验仍在 REGISTER。 |
| T-02/T-06 基础运行时 | 会话/阶段原语、状态库、锁、普通请求与 Service/Bus 的停止准入接口；不在 Journal 事务中等待 IPC。 |
| T-07 HTTP/调度 | 当前 Session 代理与任务门禁、Core 授权检查接口、票据/插件会话和拒绝路径。T-07 可用测试会话验组件，但生产没有有效管理会话必须拒绝。 |
| T-08 候选能力 | Pending/配置准备、不可变 secret 引用、受管 SQLite 副本、GORM/迁移与 PG 准入、候选结果和清理所有权。 |
| 现行规格 | CONTRACTS §7.3/§7.4/§9/§12 与 TEST-MATRIX V-27～V-31/V-40/V-41/V-47；首次管理凭证与真实登录由本任务落地。 |

#### 实施动作

1. **落地首次管理认证与真实入口。** 实现 `admin-token init/reset` 的离线 core.lock 检查、32-byte token 一次终端输出、仅存 SHA256 和认证代次；reset 撤销旧管理会话及派生票据/插件会话。实现 admin host 的 POST login、Origin/登录 CSRF、恒时摘要比较、8h 会话、登出撤销与 token 再认证查看 secret；CLI 使用显式 Bearer token。接 T-07 的授权检查，未初始化或无有效认证时拒绝管理访问；纯 loopback 开发与非 loopback HTTPS 边界沿用契约。
2. **实现开发 CLI 与安装预检。** 把 `plugin init/check/build/test/pack` 接到既有生成器和锁定工具链；仓库内示例复用根 module，面向外部作者的脚手架沿用插件自身 Go Modules 的既有方式。pack 生成 tar.gz 与归档外 release 清单，签名流程对清单确切字节签名；安装预检先做 Ed25519 验签，再核对归档 SHA256/长度、来源、平台和 manifest。受控解压拒绝越界、链接、设备、重复及大小写冲突路径，并执行 256 MiB/1 GiB/10,000 条目上限；所有下载/身份/静态错误必须在停止旧版前拒绝。
3. **建立唯一管理服务与安装记录。** CLI/WebUI/API 调同一管理服务，所有变更按插件串行，持久化安装序号、来源/受信公钥、pin、安装/卸载保留状态。实现 install/list/enable/disable/restart/reload/update/uninstall/pin/rollback 的既定区分：重启原 Active 只换 Session，reload 导入编辑快照；卸载默认保留数据及绑定，仅原来源可显式重装，purge 只删除已列明的框架受管对象。首次信任、公钥变化、恢复丢弃新写入等确认只用于契约要求的具体操作，不增加一套通用审批系统。
4. **实现正式候选 Journal 与准备编排。** 通过预检后，以预期旧 ActiveGeneration 开始持久化管理操作记录及其资源引用；冻结新根派发，drain/停旧进程，确认树退出及 runtime.lock，再调用 T-08 复制 DB 和创建候选 Session。将启动原因、维护许可和阶段预算放入已有 Bootstrap；REGISTER 校验后持久化程序/配置/secret/DB 候选引用并形成 PREPARED。普通 resume 走原提交记录，不伪造新候选 Journal；PG 不可隔离变更在相应维护门槛停下。
5. **实现唯一提交点与激活门禁。** 确认候选文件及引用已持久化后，在 Core 状态库的同一事务复核预期旧代次并提交 ActiveGeneration+COMMITTED；事务内不做下载、备份或等待 RPC。随后发送 Activate，只有 OnReady/Activated 完整成功才开放命令、事件、Service、Bus、任务与 HTTP；复用 T-06/T-07 门禁。旧按钮按 Generation 失效；提交后启动/OnReady 失败保持新代次权威并进入 Degraded，不自动倒退到旧 DB。
6. **实现崩溃恢复、显式回退与引用回收。** Core 启动持锁后按状态库 Active/Journal 恢复：未提交的已有安装关闭候选后以旧程序启动新 Session，首次安装尚无 Active 则保持未激活；已提交只恢复当前代次。旧进程退出未确认时不启动新写入者。rollback 从选定保留代次复制新的 SQLite 恢复候选并使用新 Generation，继续走同一准备/提交路径；PG/外部恢复沿用显式维护。只有无人引用的失败候选/旧包/secret 才可回收，恢复中操作和保留快照受保护；备份说明覆盖整个 data_dir 与外部 PG。
7. **完成最小 Admin 页面与 API 绑定。** 接入真实登录、状态/日志、安装启停、更新/恢复、配置表单/完整 JSON 编辑和命令权限。使用 T-08 的 raw JSON/PATCH/base_generation/secret 语义，显示 Pending、候选失败、Degraded 与维护原因；动态权限复用 T-04。通过 T-07 的两个 host、Host 校验、CSRF、受限票据和代理身份处理提供插件页面；管理 Cookie/Authorization 不进入插件，单 host 不执行插件主动内容。页面只展示用户可操作状态和具体错误，不把内部协议步骤当作用户配置项。
8. **完成正式流程的故障注入与交付证据。** 通过真实 CLI/API 在临时目录跑首次安装→enable→restart→reload→update→pin→disable→同来源重装→显式 rollback，并以包损坏/签名变化/PG 维护错误验证拒绝路径。在 PREPARED、提交前后和 OnReady 失败点终止 Core，连同残存旧进程、空间不足、并发同插件操作一起验证无双写、无提前开放、无静默数据回滚。回填 T-08 留待集成的 V-23～V-25/V-39 证据，形成 T-10 可重复使用的发布产物与恢复场景；五平台真实发布结果仍由 T-10 完成。

#### 拟目录与交付

| 实施拟建或扩展路径 | 本任务职责 |
| --- | --- |
| `cmd/alonabot/` | admin-token 命令、开发 CLI、统一管理服务客户端；运行时状态变更不直接写 DB。 |
| `internal/core/packages/` | release 清单/签名、下载/限额/安全解压、来源绑定、不可变程序包；不建设中心商店。 |
| `internal/core/management/` | 逐插件管理动作、安装/pin/卸载保留语义、候选 Journal 编排、恢复与引用清理；复用 T-02/T-08 存储与运行时原语。 |
| `internal/core/auth/`、`internal/core/http/` | 补齐首次认证、管理会话/API，与 T-07 已建授权及代理包整合，避免并列两套认证。 |
| `internal/core/` 的 T-02 状态包 | 扩展 Journal、来源/保留引用、预期代次事务与恢复查询；沿用同一 state.sqlite。 |
| `web/admin/` | 真实登录、管理页、配置/权限页及端到端交互；复用 T-08 组件。 |
| `examples/plugins/`、相邻 `testdata/` | 可安装包、配置/DB 版本对、坏包与恢复故障夹具；不得包含生产 token、DB 或签名私钥。 |

交付为可运行的管理 CLI/API/WebUI、可安装的示例发布包、唯一候选提交/恢复实现、备份与 PG 维护操作说明，以及按 V 编号关联的测试证据。实现形成后，入口至少包括 `go build ./cmd/alonabot`、管理/包/认证包测试、真实临时 Core 上的 CLI 流程和浏览器端到端验证；把实际 CLI 参数、启动配置和前端测试命令记录在夹具说明中，不能依赖执行者猜测目录或凭证来源。

#### 对应 V 验收与停止条件

| 验收编号 | 本任务通过条件 |
| --- | --- |
| V-27、V-28、V-41 | 安全解压和上限生效；签名/摘要/平台/来源变化均停机前拒绝；无自包含摘要循环，无验签失败降级无签名。 |
| V-29、V-30，复验 V-21/V-24/V-36 | 真实 WAL 候选备份、空间/锁失败、逐提交点崩溃恢复符合契约；未提交选旧、已提交选新，旧进程残存时不另启写入者，未 Activated 不接单。 |
| V-31、V-39 | PG 首装/resume/自动候选/显式维护均有正确分支；非法维护许可及不可隔离生产变更明确拒绝；失败不假称自动恢复。 |
| V-40，复验 V-23/V-25/V-38 | 所有管理命令有真实状态变化和持久记录；restart/reload 区分、同来源重装、pin、恢复新 Generation/DB 副本与旧按钮失效正确；配置/secret 语义在真实页面/API 保持。 |
| V-47，复验 V-12/V-22/V-43 | 首次 init/login/CLI 可用，reset 撤销旧授权；管理状态变更经认证/CSRF；插件 host 不取得 Admin API/Cookie/凭证；真实页面修改权限在下一次准入生效。 |

协议/Session 鉴权、安装路径或来源校验、secret 泄漏、Active/Journal 原子性、崩溃后双写、提交后自动回退丢数据、PG 维护虚假成功，任何一项失败均阻塞向 T-10 交付。组件缺口回补对应前置任务；不能把授权缺口用测试会话带入生产，不能借 T-10 再处理本任务的确定性数据恢复缺陷。尚无目标平台、PG 或浏览器证据时如实标注缺口并准备可重复夹具，不把“可编译”写成“已发布通过”。

### T-10：集成、目标平台与发布验收

#### 前置输入

- T-02～T-09 的实现、正式示例、构建产物和各阶段测试证据；47 项用例的未闭合部分可逐项定位。
- 独立 Telegram 测试账号/聊天/话题、PG 测试 schema、五目标运行环境，以及资源比较使用的机器和等价工作负载。

#### 拟涉及目录

复用各包测试、`examples/plugins/` 和构建/发布脚本；结果按 [TEST-MATRIX.md](TEST-MATRIX.md) 的证据字段记录，链接到实际测试位置与报告。发布资产目录和 CI 配置在实施时建立，`experiments/ogre-dna/` 不参与产品通过判定。

#### 按顺序实施

1. 汇总 V-01～V-47 的证据，逐项标记已通过、失败、未运行或待联验。先补齐组件之间的联验：首次管理登录→安装启用→配置/权限→真实消息→插件 HTTP→重载/升级/回退；覆盖 TCP 与 stdio。
2. 使用不同的真实插件二进制验证命令、主动 Send、Service、Event Bus、Scheduler、HTTP 和 GORM 组合。示例既覆盖无配置/无 DB 插件，也覆盖实际写入与迁移的插件；不以同一个 worker 的多份进程代替。
3. 执行协议、权限和资源边界测试：伪造/过期 Session、命令遮蔽、无 Actor、实例许可、Hook 递归、媒体租约、满准入、Cancel 顺序、超时仍运行的 job。验证错误回执与已经确认的外部结果保持一致。
4. 执行完整恢复故障注入：停止与持锁、SQLite WAL 备份、候选 OnStart 写入后失败、PREPARED/提交前后 Core 退出、不可变 secret、PG 维护许可及 resume、显式 rollback 的新代次。核对数据库内容和权威指针，不能只检查日志出现“恢复成功”。
5. 在五个目标组合分别构建并运行 Core 与插件，覆盖启动、停止、至少 100 次启停、异常退出、合作型子进程树和端口清理。记录 CGO 插件实际支持范围；无运行环境的目标保持未运行。
6. 在固定机器、版本与等价负载下测量 Core+不同插件及附属进程的总 RSS/私有内存、Linux PSS、CPU、P50/P95/P99；包含空闲、并发消息、GORM、持续 HTTP/媒体与生命周期压力，并与 NoneBot2 的等价行为比较。输出原始数据和复现命令，不预设优胜结论或虚构内存阈值。
7. 生成五平台候选资产、外部 release.json 和签名；用干净测试 data_dir 从这些真实资产完成安装、更新和卸载重装。核对来源、长度、摘要、平台和 Schema/二进制一致性，确保安装不依赖开发机的未打包文件。
8. 关闭全部阻塞用例，形成版本、平台、依赖、限制和恢复步骤的发布说明。准备好的发布包与验收记录构成本项交付；实际外部发布按届时的发布任务执行。

#### 可审查交付

- V-01～V-47 的完整结果与日志、故障注入后的状态/数据证据。
- 五平台真实运行记录和可安装候选资产；可复现的资源比较报告。
- 安装、更新、PG 维护和恢复的用户操作步骤，以及与产物一致的 SDK 示例。

#### 停止和失败条件

协议/身份、Core 全局权限、数据代次唯一性、真实恢复边界、归档来源/越界以及目标平台构建运行等发布阻塞项必须实际通过。发现回归回到负责的 T 任务修复并重跑受影响场景；没有具体风险或必需门槛时不扩张测试范围。全部必要结果齐全后结束本项。

## 5. 验收与任务完成规则

任务完成需同时具备：本项列出的实现、可复现示例或测试、对应结果记录，以及供后续任务直接调用的明确接口。不得仅因某项接口已声明便把它记作完成。

跨任务 V 用例按 TEST-MATRIX 的阶段分工记录。较早任务只完成自己负责的组件断言；用例整体在最终闭合任务完成后才标为通过。例如动态权限的管理服务测试不等于 WebUI 操作通过，内置 Send 的 Hook 检查不等于跨插件 Service 递归已经验证。后续接入仍复用较早实现，测试装配不能成为生产放行路径。

每项 V 的证据记录包含实现版本、测试位置、完整运行命令、平台/依赖、结果与失败日志。T-10 汇总时按 V-01～V-47 逐项查证，不以其他名称的新清单取代现有编号。当前全部为未运行。

若接口无法表达既定行为，先修对应契约和 fixture，再推进消费者；若已经表达但测试失败，先修负责实现。驱动备份、进程所有权、权限和提交唯一性未通过，不得被下游功能完成状态覆盖。

## 6. 资料与维护

根目录维护当前实施工作和契约；[adr/DECISIONS.md](adr/DECISIONS.md) 保存 D/P/Q 决策及处置，[adr/README.md](adr/README.md) 索引历史审阅和依据。需要了解选择原因时查 ADR，执行任务时直接按本文件与契约推进。

后续修改行为时同步契约、工作包、依赖和验收；决策理由和历史审计保存在 adr/，已有编号保持稳定。
