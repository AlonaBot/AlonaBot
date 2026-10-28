# 切片 goal 语句

本文件收录给执行者派发切片用的 goal 语句。一个 goal 对应 [PLAN §3.1](spec/PLAN.md#31-执行切片) 中的一个切片。并行派发的时机、共享文件和合并顺序按 [执行顺序](docs/execution-order.md)；执行者的工作方式按 [执行技能](.agents/skills/alonabot-execute/SKILL.md)。本文件只由审查者维护。

## 用法

1. 在「总览」中找到状态为**可派发**的 goal。
2. 把该 goal 的整段语句（代码块内的全部文字）原样发给一个执行者。一个 goal 同时只交给一个执行者。
3. 执行者开出 PR 后，把 PR 链接交给审查者审查。
4. 审查者合并后，会更新依赖它的 goal 的状态和接口段。

「总览」和各条目按 [执行顺序 §2](docs/execution-order.md#2-波次参考) 的波次排列，同一波次内按切片号排列，因此每个 goal 的前置都排在它前面。波次以执行顺序为准；新增切片或前置变化时先改那张表，再调整本文件的位置。状态只写在总览和条目的状态行，状态变化时条目不移动。

状态的含义：

- **可派发**：前置都已合并，语句已写到接口签名，可以直接派发。
- **等待前置**：语句已写完整，等前置合并、审查者核对接口后转为可派发。
- **已完成**：切片 PR 已由审查者合并，验证证据见 TEST-MATRIX §3.2。条目保留供后继切片查阅接口。
- **待编写**：范围与要点已由 PLAN 确定，没有待决的设计问题；条目只写了目标、依据和要点，还不是可派发的语句。前置合并后，审查者按实际代码写出接口、行为和测试段并写成可复制的语句块，然后才能派发。每条的完整语义以 PLAN §4 对应步骤及其失败停止条件为准，不另作设计。

## 所有 goal 的共同约束

每条 goal 语句都引用本节，执行者开始前必须完整读取。

1. 使用 `$alonabot-execute` 和 goal-mode 推进。goal 编号、分支和 worktree 使用语句里指定的值：从最新 `origin/main` 创建 `.worktrees/goal-N` 和分支 `goal-N`。
2. goal-mode 第一轮只做初始化：
   - `goal-N/input.md` 原样保存整段 goal 语句。目录被 `.gitignore` 忽略，执行记录只留在本地，不进仓库。
   - `plan.md` 和 `tasks.md` 按执行技能「切片卡格式」编写，把语句中的接口、行为和测试逐条落到 task 上，不增加也不删减范围。
3. 实现前读取 [AGENTS.md](AGENTS.md)、[编码规范](docs/coding-standards.md)、[执行顺序](docs/execution-order.md) 和语句「依据」列出的规格小节。
4. 只修改语句「范围」中列出的文件。公开接口照语句给出的签名实现；语句没写的内部辅助函数可以自定，但不新增公开符号。
5. 遇到执行技能列出的阻塞情况时标 `[!]`，不要自行决定。特别是：需要改 `.proto`、加依赖、改范围外文件、发现契约冲突。
6. Go 代码遵循 Google Go Style Guide，可用时加载 `go-google-style-guide`、`go-google-style-decisions`、`go-google-best-practices` 技能。包文档、命名、注释和测试按 [编码规范](docs/coding-standards.md) 执行，与 Google 风格冲突时以编码规范为准：每个新包在一个文件里写标准库风格的包文档，其他文件开头不写注释；名字用规格术语，不用含糊词和缩写；注释用简体中文，只写现状和原因，不写 T、goal 或 task 编号。
   - 只提交集成测试。单元测试和临时测试写在 `*_local_test.go`，它们被 git 忽略，不提交。
   - 只在信任边界做校验，不写防御性的重复检查和兜底分支。
   - 日志、错误字符串、CLI 输出和协议 safe_message 一律用英文；有含义的数值和字符串写成命名常量。
   - 日志按 CONTRACTS §14 使用注入的 `*slog.Logger`，不用 fmt 直接输出。
   - 契约中可调的限额与期限从 `internal/tune` 读取，新增的可调值登记到 `internal/tune/params.go`（CONTRACTS §16）。
7. 每个 task 都运行语句中的完成条件命令。最后一个 task 在 goal 分支上重跑全部完成条件，然后推送分支，开 PR 到 `main`：标题为「切片 ID + 标题」，正文按 PR 模板填写，并加一节「文档同步」，列出需要审查者同步到 ARCHITECTURE、README 或 spec 的内容。不合并 PR。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-65 | T-02.1 | 日志基线与日志文件 | — | 已完成 |
| goal-23 | T-07.1 | 调度触发计算 | — | 已完成 |
| goal-29 | T-08.2 | 配置 JSON 语义与 Pending | — | 已完成 |
| goal-37 | T-09.2 | 安装包预检与安全解压 | — | 已完成 |
| goal-58 | T-12.1 | Provider 客户端 | — | 已完成 |
| goal-76 | T-02.2 | 调优参数 | T-02.1 | 已完成（PR #28） |
| goal-2 | T-03.1 | 首个可用 SDK 切片 | T-02.1、T-02.2 | 已完成（PR #29） |
| goal-3 | T-03.2 | Target 与发信许可 | T-03.1 | 已完成（PR #31） |
| goal-4 | T-03.3 | Adapter 进程、sdk/adapter 与上下线事件 | T-03.1 | 已完成（PR #32） |
| goal-6 | T-03.5 | 媒体上传与配额 | T-03.1 | 可派发 |
| goal-10 | T-04.1 | 根插件 API 与静态注册索引 | T-03.1 | 可派发 |
| goal-56 | T-11.3 | Agent 会话存储 | T-03.1 | 可派发 |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 等待前置 |
| goal-11 | T-04.2 | 构建期生成器 | T-04.1 | 等待前置 |
| goal-12 | T-04.3 | 命令选择、文本匹配与参数绑定 | T-04.1 | 等待前置 |
| goal-13 | T-04.4 | 持久化命令策略 | T-03.2、T-04.1 | 等待前置 |
| goal-19 | T-06.1 | Service/Bus 描述与 Broker 准入 | T-04.1 | 等待前置 |
| goal-25 | T-07.3 | SDK HTTP Listener 与 Core 代理 | T-04.1 | 等待前置 |
| goal-78 | T-15.1 | 官方 OneBot 11 Adapter 的连接与接收 | T-03.3 | 等待前置 |
| goal-7 | T-03.6 | 入站媒体与 Grant | T-03.2、T-03.4、T-03.5 | 等待前置 |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 等待前置 |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 等待前置 |
| goal-21 | T-06.3 | 非持久 Event Bus | T-06.1 | 等待前置 |
| goal-26 | T-07.4 | 两个 Host 与插件票据 | T-07.3 | 等待前置 |
| goal-28 | T-08.1 | 配置字段描述与构建检查 | T-04.2 | 等待前置 |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-06.1 | 等待前置 |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2 | 等待前置 |
| goal-8 | T-03.8 | 按钮、extension 片段与依赖解析 | T-03.2、T-03.7 | 等待前置 |
| goal-51 | T-03.9 | 消息编辑 | T-03.7 | 等待前置 |
| goal-16 | T-05.2 | Rule、HandlerOutcome 与传播 | T-05.1 | 等待前置 |
| goal-24 | T-07.2 | schedule 注册与运行槽 | T-05.1、T-07.1 | 等待前置 |
| goal-30 | T-08.3 | SDK 配置准备顺序 | T-08.1、T-08.2 | 等待前置 |
| goal-31 | T-08.4 | 不可变 Secret Store | T-08.1、T-08.2 | 等待前置 |
| goal-34 | T-08.7 | web/admin 骨架、配置表单组件与界面文案 | T-08.1 | 等待前置 |
| goal-53 | T-08.8 | 用户可见文案目录 | T-04.1、T-04.3、T-08.1 | 等待前置 |
| goal-36 | T-09.1 | 首次管理认证 | T-07.4 | 等待前置 |
| goal-69 | T-13.3 | 热部署 | T-13.1 | 等待前置 |
| goal-9 | T-03.10 | Telegram 接入开发入口与故障矩阵 | T-03.6、T-03.8、T-03.9 | 等待前置 |
| goal-17 | T-05.3 | 三阶段 Hook 与 send_forbidden | T-05.2 | 等待前置 |
| goal-52 | T-05.4 | 延续调用与多轮等待 | T-05.2 | 等待前置 |
| goal-27 | T-07.5 | 调度与 HTTP 启停集成及验收 | T-07.2、T-07.4 | 等待前置 |
| goal-32 | T-08.5 | 受管 SQLite 连接与迁移 | T-08.3 | 等待前置 |
| goal-39 | T-09.4 | 管理服务、安装记录与 serve | T-03.8、T-08.4、T-09.1、T-09.2、T-13.1 | 等待前置 |
| goal-71 | T-14.1 | help 插件与命令列表 | T-04.4、T-05.2、T-08.8 | 等待前置 |
| goal-79 | T-15.2 | 官方 OneBot 11 Adapter 的发送、媒体与 extension | T-03.6、T-03.8、T-15.1 | 等待前置 |
| goal-14 | T-04.5 | 示例升级与 T-04 验收 | T-03.10、T-04.2、T-04.3、T-04.4 | 等待前置 |
| goal-18 | T-05.5 | 多进程事件流水线验收 | T-03.8、T-05.3、T-05.4 | 等待前置 |
| goal-20 | T-06.2 | 父调用链、取消与背压 | T-03.6、T-05.3、T-06.1 | 等待前置 |
| goal-33 | T-08.6 | SQLite 候选与资源交接 | T-08.4、T-08.5 | 等待前置 |
| goal-84 | T-09.12 | WebUI 组件与组件图标 | T-07.4、T-09.1、T-09.4、T-13.1 | 待编写 |
| goal-59 | T-12.2 | Agent 对话与流式显示 | T-03.9、T-04.3、T-05.4、T-08.3、T-08.4、T-11.3、T-12.1 | 等待前置 |
| goal-68 | T-13.2 | 开发组件的配置与数据库 | T-08.3、T-08.5、T-13.1 | 等待前置 |
| goal-72 | T-14.2 | welcome 与 rss 插件 | T-05.2、T-07.2、T-08.3、T-08.5、T-08.8 | 等待前置 |
| goal-80 | T-15.3 | 跨 Adapter 一致性验收 | T-03.10、T-15.2 | 等待前置 |
| goal-22 | T-06.4 | Session 清理与 Service 端到端 | T-06.2、T-06.3 | 等待前置 |
| goal-35 | T-08.9 | configdb 示例与 T-08 验收 | T-08.6、T-08.7、T-08.8 | 等待前置 |
| goal-38 | T-09.3 | 开发 CLI 与库模式脚手架 | T-04.5、T-08.1、T-08.8、T-09.2、T-13.1 | 等待前置 |
| goal-55 | T-11.2 | 工具调用与用户确认 | T-03.8、T-06.2、T-08.8、T-11.1 | 等待前置 |
| goal-70 | T-13.4 | 库模式示例与验收 | T-03.10、T-13.2、T-13.3 | 等待前置 |
| goal-81 | T-16.1 | 插件测试套件基础 | T-05.2、T-13.2 | 等待前置 |
| goal-83 | T-12.8 | 文本生成服务 | T-03.8、T-06.1、T-12.2 | 等待前置 |
| goal-40 | T-09.5 | 候选 Journal 与准备编排 | T-09.4、T-08.6、T-06.4 | 等待前置 |
| goal-57 | T-11.4 | 工具与会话的验收 | T-11.2、T-11.3 | 等待前置 |
| goal-60 | T-12.3 | 工具循环与 MCP | T-11.2、T-12.2 | 等待前置 |
| goal-61 | T-12.4 | Agent 的记忆 | T-08.5、T-11.2、T-12.2 | 等待前置 |
| goal-82 | T-16.2 | 按钮、多轮等待、schedule 与示例测试 | T-03.8、T-05.4、T-07.2、T-16.1 | 等待前置 |
| goal-41 | T-09.6 | 唯一提交点与激活门禁 | T-07.5、T-09.5 | 等待前置 |
| goal-62 | T-12.5 | 技能 | T-12.3 | 等待前置 |
| goal-63 | T-12.6 | 定时 Agent 任务 | T-07.2、T-08.5、T-12.3 | 等待前置 |
| goal-42 | T-09.7 | 崩溃恢复、回退与回收 | T-09.6 | 等待前置 |
| goal-43 | T-09.8 | Admin 管理页 | T-09.6、T-08.7、T-09.12 | 等待前置 |
| goal-64 | T-12.7 | 官方 Agent 组件打包与 T-12 验收 | T-12.4、T-12.5、T-12.6、T-12.8 | 等待前置 |
| goal-44 | T-09.9 | Admin 配置、权限、工具策略与插件页 | T-09.8、T-11.1、T-11.3 | 等待前置 |
| goal-66 | T-09.10 | 日志管理与日志页 | T-02.1、T-03.1、T-09.8 | 等待前置 |
| goal-73 | T-14.3 | 发行归档 | T-03.10、T-09.3、T-12.7、T-14.1、T-14.2、T-15.2 | 等待前置 |
| goal-45 | T-09.11 | 正式流程故障注入与 T-09 验收 | T-09.3、T-09.7、T-09.9、T-09.10、T-09.12 | 等待前置 |
| goal-74 | T-14.4 | 首次运行与预装 | T-09.6、T-14.3 | 等待前置 |
| goal-77 | T-14.5 | 容器镜像 | T-07.3、T-14.3 | 等待前置 |
| goal-85 | T-14.7 | install.sh | T-14.3 | 待编写 |
| goal-75 | T-14.6 | 发行包与镜像端到端验收 | T-09.11、T-14.4、T-14.5 | 等待前置 |
| goal-46 | T-10.1 | 证据汇总与跨组件联验 | T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.6、T-14.7、T-15.3、T-16.2 | 等待前置 |
| goal-47 | T-10.2 | 五平台构建与运行 | T-10.1 | 等待前置 |
| goal-48 | T-10.3 | 资源比较报告与常驻内存预算核对 | T-10.1 | 等待前置 |
| goal-49 | T-10.4 | 发布资产与说明 | T-10.2、T-10.3 | 等待前置 |

## 第 1 波

### goal-65 · T-02.1 日志基线与日志文件

状态：已完成（PR #13 已合并）。本 goal 早于 goals.md 派发，没有原始语句；下面记录合并后的实际接口，供后继切片引用。

目标：日志基线：Core 与 SDK 改用注入的 slog Logger，控制台与 JSON 格式、来源标签、插件 stderr 的 JSON 解析与非结构化行、panic 堆栈、Bootstrap 下发级别与进程内过滤、`--log-level`/`--log-format`、不输出 Bootstrap 与配置明文；按日期轮转的日志文件与保留上限
前置：T-02。
依据：PLAN §3.1 的 T-02.1 行、§4 的 T-02 步骤 9；验收：V-62、V-63 的级别过滤分项、V-68 的写入与轮转分项。

接口（main 上的实际代码）：

```go
// internal/core/logging
const KeyComponent, KeyPluginID, KeyAdapterID, KeyInstanceID, KeySessionID, KeyGenerationID, KeyTraceID // 来源字段名
const ComponentCore, ComponentPlugin, ComponentAdapter
const LevelOff slog.Level // 来源级别 off
type Format string // FormatText、FormatJSON
type Options struct { Format Format; Level slog.Level; Stderr io.Writer }
func Open(opt Options) (*Output, error)
func (o *Output) Logger() *slog.Logger // component=core，按全局级别过滤
func (o *Output) Levels() *Levels
func (o *Output) EnableFile(opt FileOptions) error // 调用方须已持有 core.lock，只调用一次
func (o *Output) Close() error
func (o *Output) Plugin(pluginID, generationID string) *Source
type Source struct{ ... }
func (s *Source) SetSession(sessionID []byte)
func (s *Source) Line(pluginID, stream string, line []byte) // 作为 process 的输出行回调
type Levels struct{ ... } // SetGlobal、SetSource、InheritSource、Source
func ProtoLevel(lv slog.Level) pb.LogLevel
type FileOptions struct { Dir string; MaxSize, MaxTotal int64; MaxAgeDays int } // DefaultMaxFileSize、DefaultMaxTotalSize
// contracts/proto：Bootstrap.log_level = 10（enum LogLevel）
// sdk（未导出）：setLogLevel 按 Bootstrap.log_level 在进程内过滤；logPanic 记带 stack 的 ERROR；
//   Run 把 slog 默认 Logger 换成写 stderr 的 JSON 行
// cmd/alonabot run：--log-level、--log-format text|json
```

### goal-23 · T-07.1 调度触发计算

状态：已完成（PR 已合并）。

```text
/goal 使用 $alonabot-execute 推进 goal-23（切片 T-07.1：调度触发计算）。goal 目录 goal-23，分支 goal-23，worktree .worktrees/goal-23，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：一个纯计算包：校验 interval 与五字段 cron（含 IANA 时区），并给出下一次触发时间。Core Scheduler 与官方 Agent 组件的定时任务都使用它。不接入 REGISTER，不触发任何调用。

依据：PLAN T-07 步骤 2；CONTRACTS §4.3；TEST-MATRIX V-20。

范围：只新增 contracts/schedule/（实现与测试）。
允许的新跨模块边：无。contracts/schedule 只 import 同在 contracts 模块内的 contracts/pb。

接口：
type Trigger struct{ ... } // 已校验的触发规则；零值不可用
func Interval(d time.Duration) (Trigger, error)    // d < 1s 返回错误
func Cron(expr, timezone string) (Trigger, error)  // timezone 为空表示 UTC
func FromSpec(s *pb.ScheduleSpec) (Trigger, error) // interval_ms 与 cron 恰有一个；不检查 job_deadline_ms
func (t Trigger) Next(anchor, now time.Time) time.Time // 严格晚于 now 的下一次触发；anchor 是激活时刻，只用于 interval

行为：
- interval：触发时刻是 anchor + k×d（k ≥ 1）。Next 返回其中严格晚于 now 的最小值；已错过的直接跳过，不补。
- cron 固定 5 个字段：分 0–59、时 0–23、日 1–31、月 1–12、星期 0–7（0 与 7 都是周日）。
  - 每个字段支持 *、数字、a-b、逗号列表、*/n、a-b/n。
  - 以下情况报错：名称（JAN、MON）、@ 宏、L/W/#、秒字段、字段数不是 5、越界、步长为 0、a>b。
- 日与星期都不是 * 时按 OR 匹配；其余字段按 AND。
- 时区：timezone 必须能被 time.LoadLocation 解析，且不是 "Local"。包内 import _ "time/tzdata"，保证没有系统时区库的平台也能解析。
- 夏令时：按真实时刻（UTC 分钟）推进，用本地时间判断是否匹配。本地时刻不存在（春季跳过）时跳过；本地时刻重复（秋季回拨）时，两个不同的 UTC 时刻都算触发。
- 永远不会触发的表达式（例如 2 月 30 日）在 Cron 中报错：从 2000-01-01 起 8 年内找不到触发时刻，即视为无效。
- Next 对同样的输入总是返回同样的结果：不读当前时间，不持有状态。

测试（按编码规范「测试提交规则」）：
- 提交：无。这是没有外部协作方的纯计算包，集成测试在调度接入时补上。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：错过的 tick 被跳过；interval 小于 1s 被拒；每种 cron 语法错误；常见表达式的期望值；日与星期同时受限时按 OR 匹配；America/New_York 下 30 2 * * * 在 2026-03-08 不触发、下一次为 2026-03-09 06:30Z；30 1 * * * 在 2026-11-01 触发两次（05:30Z 与 06:30Z）；无效时区被拒、空时区等于 UTC；0 0 30 2 * 被拒；FromSpec 的两个分支与非法组合。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/schedule/；go build ./...。

不做：REGISTER 校验接入、运行槽、Core→插件调用、SDK 改动。
```

### goal-29 · T-08.2 配置 JSON 语义与 Pending

状态：已完成（PR #23 已合并）。后继注意：`pending_configs.plugin_id` 外键引用 `installs`，删除安装记录前须先删除该插件的 Pending 行。

```text
/goal 使用 $alonabot-execute 推进 goal-29（切片 T-08.2：配置 JSON 语义与 Pending）。goal 目录 goal-29，分支 goal-29，worktree .worktrees/goal-29，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 侧的配置 JSON 语义与 Pending 编辑能力：
- 保留原始对象和未知键；
- 支持完整替换、RFC 7396 Merge Patch、从 config.json 导入编辑快照；
- 用 base_generation 拒绝陈旧提交。
本切片只提供 Core 内部的应用层入口：不开放 HTTP/CLI，不激活 Pending，不处理默认值、Schema 和 secret。

依据：PLAN T-08 步骤 2；CONTRACTS §5.1；TEST-MATRIX V-23、V-38。

范围：
- 新增 internal/core/config/（JSON 文档操作与 Service）
- 修改 internal/core/state/state.go：只在 migrations 末尾追加一条
- 新增 internal/core/state/pending.go
- 上述包的测试文件
允许的新跨模块边：无。

接口：
// internal/core/config
func Canonical(raw []byte) ([]byte, error)
// 必须是 JSON 对象；输出按键排序、无多余空白、数字保留原文。非对象或语法错误返回 Code=INVALID_ARGUMENT 的 *validate.Error。
func MergePatch(doc, patch []byte) ([]byte, error) // RFC 7396；doc 与 patch 都必须是对象；输出为 Canonical 形式
type Pending struct { PluginID string; BaseGeneration string; ConfigJSON []byte; UpdatedAt time.Time }
type Service struct{ ... }
func New(store *state.Store, dataDir string) *Service
func (s *Service) Pending(ctx context.Context, pluginID string) (Pending, bool, error)
func (s *Service) Replace(ctx context.Context, pluginID, baseGeneration string, full []byte) (Pending, error)
func (s *Service) Patch(ctx context.Context, pluginID, baseGeneration string, patch []byte) (Pending, error)
func (s *Service) ImportSnapshot(ctx context.Context, pluginID string) (Pending, error) // 读 <dataDir>/plugins/<id>/config.json
func (s *Service) Discard(ctx context.Context, pluginID string) error
// internal/core/state/pending.go
type PendingConfig struct { PluginID string; BaseGeneration string; ConfigJSON string; UpdatedAt time.Time }
func (s *Store) GetPendingConfig(ctx context.Context, pluginID string) (PendingConfig, error) // 不存在时返回 ErrNotFound
func (s *Store) PutPendingConfig(ctx context.Context, p PendingConfig) error
func (s *Store) DeletePendingConfig(ctx context.Context, pluginID string) error
func (s *Store) GenerationConfig(ctx context.Context, generationID string) (string, error) // 读 generations.config_snapshot
迁移：CREATE TABLE pending_configs (plugin_id TEXT PRIMARY KEY REFERENCES installs(plugin_id), base_generation TEXT NOT NULL, config_json TEXT NOT NULL, updated_at INTEGER NOT NULL);

行为：
- Pending 只有在 base_generation 等于插件当前 Active 时才有效。Active 变化后，旧 Pending 在 Pending() 中视为不存在（不删除，由 Discard 或下一次写入覆盖）。插件尚无 Active 时，有效的 base 是空串。
- Replace 和 Patch：baseGeneration 必须等于当前 Active，否则返回 FAILED_PRECONDITION。
  - Patch 的基础文档：有有效 Pending 时用它，否则用 Active 代次的 config_snapshot，没有 Active 时用 {}。
  - Replace 直接使用 full。
  - 两者的结果都经 Canonical 后写入。
- ImportSnapshot：读取 config.json，base 为当前 Active，导入不改变 Active。
  - 内容不是 JSON 对象 → INVALID_ARGUMENT，已有的 Pending 不变。
  - 文件不存在 → NOT_FOUND。
- 语义：
  - 未知键和嵌套未知键全部保留；显式的 false、0、""、[]、{}、null 原样保留。
  - Merge Patch 中 null 删除键，数组整体替换，对象递归合并。
  - 超过 2^53 的整数（例如 9007199254740993）原样保留，不经过 float64。
- 错误一律返回带 Code 的 *validate.Error。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/config/service_test.go。用 state.Open 打开临时 data_dir 的真实状态库，经 Service 的公开方法做表驱动测试：
  - base 不等于 Active 时返回 FAILED_PRECONDITION。
  - CommitActive 新代次后，旧 Pending 不再返回。
  - Patch 有 Pending 时叠加在 Pending 上，没有时叠加在 Active 的配置上；结果保留未知键、显式空值和 9007199254740993 这样的大整数。
  - ImportSnapshot 不改变 Active；config.json 不是对象时被拒，已有 Pending 不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Canonical 的数字与键处理、非对象输入；MergePatch 跑通 RFC 7396 附录 A 的全部示例；state 新方法的读写往返。
- 现有 state 测试必须继续通过；如有断言 schema 版本的测试，更新为新版本号。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/config/ ./internal/core/state/；go build ./...。

不做：默认值与 Schema 校验、secret 引用、激活 Pending、HTTP/CLI/WebUI 入口。
```

### goal-37 · T-09.2 安装包预检与安全解压

状态：已完成（PR 已合并）。

```text
/goal 使用 $alonabot-execute 推进 goal-37（切片 T-09.2：安装包预检与安全解压）。goal 目录 goal-37，分支 goal-37，worktree .worktrees/goal-37，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：给定一份本地 tar.gz 归档和归档外的 release.json 及签名：
1. 先验签，并核对来源、平台、长度与 SHA256；
2. 再受控解压到新的暂存目录，核对内层 manifest；
3. 最后移动为不可变的包目录。
不做下载，不登记安装，不提供 CLI。

依据：PLAN T-09 步骤 2（预检部分）；CONTRACTS §7.1、§9.1、§9.2，以及 §9.3 末段的安装器上限；TEST-MATRIX V-27、V-28、V-41。

范围：只新增 internal/core/packages/（实现与测试）。测试 fixture 在测试中生成，不提交二进制文件。
允许的新跨模块边：internal/core/packages → contracts（已允许）。

接口：
type Limits struct { MaxArchiveBytes int64; MaxExpandedBytes int64; MaxEntries int }
var DefaultLimits = Limits{MaxArchiveBytes: 256 << 20, MaxExpandedBytes: 1 << 30, MaxEntries: 10000}
type Input struct {
    ReleaseJSON []byte          // release.json 原始字节
    Signature   []byte          // release.json.sig 原始字节
    PublicKey   ed25519.PublicKey // 已信任的发布者公钥
    Source      string          // 已绑定的来源，例如 github.com/owner/repo
    OS, Arch    string          // 本机平台
    ArchivePath string          // 已下载到本地的归档
}
type Candidate struct { Dir string; Manifest *manifest.Manifest; Release *release.Release; Asset release.Asset }
func Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error)
func Extract(archivePath, dest string, lim Limits) error
func (c *Candidate) MoveTo(packagesRoot string) (string, error) // 移入不可变包目录，返回最终目录；「Commit」专指代次提交，这里不用
func (c *Candidate) Discard() error                            // 删除暂存目录

行为：
- Preflight 按顺序检查，任何一步失败都删除已创建的临时目录，返回带 Code 的 *validate.Error：
  1) release.Verify(ReleaseJSON, Signature, PublicKey) 失败 → PERMISSION_DENIED。
  2) Release.Source 不等于 Input.Source → PERMISSION_DENIED。
  3) Release.Asset(OS, Arch) 不存在 → UNSUPPORTED。
  4) 归档超过 lim.MaxArchiveBytes → RESOURCE_EXHAUSTED；大小不等于 Asset.Size，或整个文件的 SHA256 不等于 Asset.SHA256 → INVALID_ARGUMENT。
  5) 在 stagingRoot 下用 os.MkdirTemp 新建唯一临时目录，并 Extract 到其中。
  6) 用 manifest.Parse 解析解压后的 plugin.yaml。以下情况 → INVALID_ARGUMENT：id、version 或 source 与 Release 不一致；os/arch 不等于本机平台；entrypoint 不是解压树中的普通文件。
- Extract：
  - 只接受 gzip 压缩的 tar，只接受普通文件和目录条目。
  - 以下情况 → INVALID_ARGUMENT：绝对路径；含 .. 组件；含反斜杠；空名字；symlink、hardlink、字符或块设备、FIFO；重复路径；按 strings.ToLower 比较相同的两个路径。
  - 条目数超过 lim.MaxEntries，或实际写出的字节总数超过 lim.MaxExpandedBytes → RESOURCE_EXHAUSTED。字节数在写入过程中累计。
  - 目录权限 0o700；普通文件 0o600，归档中带任一执行位的文件 0o700；不保留 setuid、setgid、sticky。
  - 失败时不保证 dest 为空，由 Preflight 删除整个暂存目录。
- MoveTo：
  - 最终目录为 <packagesRoot>/<plugin_id>/<version>-<归档 SHA256 前 12 位>。已存在 → ALREADY_EXISTS，绝不覆盖。
  - 用 os.Rename 从暂存目录移动（同一文件系统由调用方保证），失败时原样返回错误。

测试（按编码规范「测试提交规则」；fixture 在测试中用 archive/tar + compress/gzip 生成，签名用测试内生成的 ed25519 密钥）：
- 提交：internal/core/packages/preflight_test.go。在临时目录生成真实的 tar.gz、release.json 和签名，经 Preflight 与 MoveTo 做表驱动测试：
  - 合法包通过，MoveTo 后最终目录内容正确；对同一目标再次 MoveTo 返回 ALREADY_EXISTS。
  - 每种拒绝各一个子测试：签名错误、来源不符、缺少本平台资产、大小不符、摘要不符、manifest 的 id/version/source/平台/entrypoint 不符、../x、/abs、a/../../x、a\b、symlink、hardlink、设备文件、FIFO、A.txt 与 a.txt、重复路径、条目数和展开字节超过（用很小的 Limits）。每个子测试都断言 stagingRoot 下没有残留目录。
  - 文件权限（//go:build unix 的单独测试）：带执行位的文件为 0o700，其余为 0o600，setuid 被去掉。
- 本地自测（*_local_test.go，不提交）：Extract 的单项边界，按需编写。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/packages/；go build ./...。

不做：下载、首次来源信任的持久化、安装记录、ID 占用检查、CLI、pack。
```

### goal-58 · T-12.1 Provider 客户端

状态：已完成（PR #16 已合并；2026-10-28 按 D-64 修正：增加 `openai` 协议、停止原因与用量，包移到 `agent/provider`）。本 goal 的原始语句未保留；下面记录当前的实际接口，供后继切片引用。

目标：Provider 客户端：OpenAI、OpenAI 兼容与 Anthropic Messages 的流式请求、SSE 解析、工具调用解析、停止原因与用量、本地模拟服务
前置：T-02。
依据：PLAN §3.1 的 T-12.1 行、§4 的 T-12 步骤 1（接口与流语义）；验收：V-57 的 Provider 分项。

接口（main 上的实际代码）：

```go
// agent/provider
type Protocol string // OpenAI = "openai"、OpenAICompatible = "openai-compatible"、AnthropicMessages = "anthropic-messages"
type Config struct { Protocol Protocol; BaseURL string; APIKey string; Model string }
type Message struct { Role string; Content string; ToolCalls []ToolCall; ToolResults []ToolResult } // Role：system、user、assistant
type Tool struct { Name string; Description string; InputSchema json.RawMessage }
type Request struct { Messages []Message; Tools []Tool; MaxTokens int } // MaxTokens 必须大于 0
type ToolCall struct { ID string; Name string; Arguments json.RawMessage }
type ToolResult struct { ToolCallID string; Content string }
type Event struct { Text string; ToolCall *ToolCall } // ToolCall 非 nil 时是组装完整的一次工具调用
func New(config Config, httpClient *http.Client) (*Client, error) // httpClient 由调用方注入
type StopReason string // StopEnd = "end"、StopToolUse = "tool_use"、StopMaxTokens = "max_tokens"；其他取值原样保留
type Usage struct { InputTokens int; OutputTokens int }
type Result struct { StopReason StopReason; Usage *Usage } // Usage 为 nil 表示 Provider 没有返回用量
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
// Stream 不重发请求；已交给 emit 的文本增量由调用方保留；错误不含 API key 与认证头，附带的 Provider 说明中回显的 key 替换为 [REDACTED]。
```

## 第 2 波

### goal-76 · T-02.2 调优参数

状态：已完成（PR #28；实现、本地验收与评论修复完成）。

```text
/goal 使用 $alonabot-execute 推进 goal-76（切片 T-02.2：调优参数）。goal 目录 goal-76，分支 goal-76，worktree .worktrees/goal-76，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：建立调优参数表 internal/tune，把 internal/wire 与 internal/core 中可调的现有常量改为读取生效值；alonabot 根命令接受可重复的全局参数 --args，在执行任何子命令前校验并报告覆盖；新增 alonabot tune list；Core 经 Bootstrap.session_limits 把 Session 限额下发给 SDK，双方按同一组值执行；PLUGIN_MEMORY_SOFT_LIMIT 与 CORE_MEMORY_SOFT_LIMIT 设置 Go 运行时软上限。

依据：PLAN T-02 步骤 10；CONTRACTS §1.2、§4.2、§16；WIRE-FORMAT §3.2（SessionLimits）；TEST-MATRIX V-69；DECISIONS D-41、D-42。

范围（只新增或修改这些）：
- 新增 internal/tune/（doc.go、params.go 与解析代码）
- contracts/proto/alonabot/v1/control.proto：追加 message SessionLimits 与 Bootstrap 的 SessionLimits session_limits = 11；随后 go generate ./... 重新生成 contracts/pb（本切片获授权修改该 proto，字段号按 WIRE-FORMAT，不改已有编号）
- contracts/wirespec_test.go：追加 SessionLimits 与 Bootstrap.session_limits 的字段号检查
- internal/wire/session.go：删除 DefaultDeadline、MaxInflightPerDirection、MaxBytesPerDirection、GlobalPayloadBytes 四个常量
- internal/core/host/：prepare.go、activate.go、stop.go
- internal/core/packages/preflight.go
- internal/core/supervisor/supervisor.go
- sdk/session.go、sdk/activate.go
- cmd/alonabot/：main.go 改为 cobra 命令树，可按命令拆成同包的多个文件
- go.mod、go.sum：只新增 github.com/spf13/cobra 及其传递依赖（D-42 已批准；用 go get 取当前最新的 v1 版本，PR 中写明版本，并用 go mod why 确认编进二进制的只有 cobra、pflag 与 Windows 上的 mousetrap）
- 上述包中引用被删除或改名符号的测试文件（只做随之必要的调用修改）
允许的新跨模块边：cmd/alonabot -> internal/tune（ARCHITECTURE 已登记）。internal/wire、internal/core、sdk -> internal/tune 已允许；internal/tune 不 import 本 module 任何包。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/tune
   // Values 是一次运行的全部生效值，每个 CONTRACTS §16 参数一个字段；字段名为参数名的驼峰形式。
   type Values struct {
       SessionMaxInflight    int
       SessionMaxPayload     int64         // 字节
       CoreMaxPayload        int64
       RequestTimeout        time.Duration
       CallChainMaxDepth     int
       DispatchConcurrency   int
       DispatchQueueEvents   int
       DispatchQueueBytes    int64
       ContinuationMax       int
       ContinuationTimeout   time.Duration
       BusQueueEvents        int
       BusQueueBytes         int64
       MediaMaxObject        int64
       MediaMaxUploads       int
       MediaSessionBudget    int64
       MediaDiskBudget       int64
       MediaTTL              time.Duration
       PhaseHello            time.Duration
       PhaseConfig           time.Duration
       PhaseMigration        time.Duration
       PhaseOnStart          time.Duration
       PhaseOnReady          time.Duration
       StopTotal             time.Duration
       StopDrain             time.Duration
       StopOnStop            time.Duration
       RestartWindow         time.Duration
       RestartMaxCrashes     int
       JobTimeout            time.Duration
       ButtonTokenTTL        time.Duration
       ToolConfirmTimeout    time.Duration
       LogBufferRecords      int
       InstallMaxArchive     int64
       InstallMaxExpanded    int64
       InstallMaxEntries     int
       DevReloadPoll         time.Duration
       PluginMemorySoftLimit int64 // 0 表示不设置
       CoreMemorySoftLimit   int64 // 0 表示不设置
   }
   func Defaults() Values
   // Override 是一项覆盖；Origin 为 "--args"（直接参数）或 "<文件路径>:<行号>"。
   type Override struct { Name, Value, Origin string }
   // Parse 依次应用 args 中的每一项（每项是一次 --args 的值），后出现的覆盖先出现的，最后检查约束与相互约束。
   // 返回生效值与实际生效的覆盖（同一参数只保留最后一次），按参数表顺序排列。
   func Parse(args []string) (Values, []Override, error)
   // WriteList 按参数表顺序写出全部参数的名称、缺省值、v 中的生效值、约束与契约位置，供 alonabot tune list 使用。
   func WriteList(w io.Writer, v Values) error
2. contracts/proto（照 WIRE-FORMAT §3.2）
   message SessionLimits { uint32 max_inflight = 1; uint32 max_payload_bytes = 2; uint32 request_timeout_ms = 3; }
   Bootstrap 追加 SessionLimits session_limits = 11;
3. internal/core/host
   Limits 增加字段：SessionMaxInflight int; SessionMaxPayload int64; RequestTimeout time.Duration
   func NewLimits(v tune.Values) Limits         // 取代 var DefaultLimits
   func NewStopTiming(v tune.Values) StopTiming // 取代 var DefaultStopTiming
   func NewRestartPolicy(v tune.Values) *RestartPolicy // 窗口与次数取 RestartWindow、RestartMaxCrashes；删除 restartWindow、restartLimit
4. internal/core/packages
   func NewLimits(v tune.Values) Limits // 取代 var DefaultLimits
5. internal/core/supervisor
   Spec 增加字段：MemorySoftLimit int64 // 大于 0 时为插件进程设置 GOMEMLIMIT

行为：
- 参数表：internal/tune/params.go 按 CONTRACTS §16 表的顺序登记全部 37 个参数，每项有名称、单位类别（整数、时长、大小）、缺省值、约束与契约位置；Defaults 与表一致。PLUGIN_MEMORY_SOFT_LIMIT、CORE_MEMORY_SOFT_LIMIT 缺省不设置。帧上限 4MiB 在 tune 内写成命名常量并注明与 wire.MaxFrame 相同（tune 不能 import wire）。
- --args 每一项：
  - 含 "=" 时是直接参数，按逗号分隔为多个 KEY=VALUE；
  - 否则是文本文件路径：每行一个 KEY=VALUE，忽略空行与首个非空白字符为 # 的行；
  - 同一参数后出现的覆盖先出现的，包括跨项与同一项内。
- 值的格式：整数为十进制；时长用 time.ParseDuration；大小为十进制整数加 B、KiB、MiB、GiB 后缀之一；写成 KB、MB、GB 时报错并提示改用二进制单位。
- 以下情况 Parse 返回错误，错误文本指明出处（"--args" 中的哪一项，或文件路径与行号）与参数名：未知键、缺少 "="、空键、格式错误、违反单值约束、文件不可读。相互约束（SESSION_MAX_PAYLOAD 与 DISPATCH_QUEUE_BYTES、BUS_QUEUE_BYTES ≥ 4MiB；CORE_MAX_PAYLOAD ≥ SESSION_MAX_PAYLOAD；CONTINUATION_MAX < SESSION_MAX_INFLIGHT；MEDIA_MAX_OBJECT ≤ MEDIA_SESSION_BUDGET ≤ MEDIA_DISK_BUDGET；INSTALL_MAX_EXPANDED ≥ INSTALL_MAX_ARCHIVE；STOP_TOTAL ≥ STOP_DRAIN + STOP_ONSTOP）在全部覆盖应用后检查，错误列出相关参数的值与各自出处。超出字段类型或 uint32 毫秒/字节下发范围的值按格式错误处理。
- cmd/alonabot：
  - 用 cobra 组织命令树（D-42）：根命令 alonabot，子命令 run 与命令组 tune（含 list）。run 的现有参数（--data-dir、--plugin-id、--entry、--stdio、--log-level、--log-format）名称、缺省值与必填规则不变，改为 pflag 定义；--log-level 仍按 slog.Level 文本解析。
  - --args 定义为根命令的 persistent flag（StringArray，可重复），写在子命令前后都生效（D-41）；同一命令行中按出现顺序传给 tune.Parse。
  - 在根命令的 PersistentPreRunE 中调用 tune.Parse，执行任何子命令之前完成；出错时把错误写到 stderr 并以 exitUsage 退出，不执行子命令。参数错误与未知子命令同样以 exitUsage 退出；Execute 返回的运行错误以 exitFailure 退出。关闭 cobra 自动打印的错误与 usage（SilenceErrors、SilenceUsage），由 main 统一输出，避免重复。
  - 不生成 shell 补全命令（CompletionOptions.DisableDefaultCmd）。
  - run 在 logging.Open 之后，有覆盖时以 Core Logger 输出一条 WARN，消息为 "tuning parameters overridden"，属性列出每项的名称、值与出处。
  - 新增子命令 alonabot tune list：调用 tune.WriteList(os.Stdout, v) 列出缺省值与生效值，以 0 退出；不需要 data_dir，也不输出 WARN。帮助文本由 cobra 生成，Short 描述用英文。
  - Core 使用生效值：supervisor.Spec 的 Limits=host.NewLimits(v)、Stop=host.NewStopTiming(v)、Global=wire.NewBudget(v.CoreMaxPayload)、MemorySoftLimit=v.PluginMemorySoftLimit；重启策略为 host.NewRestartPolicy(v)。
  - CoreMemorySoftLimit > 0 时，启动时调用 debug.SetMemoryLimit。
- host：Prepare 按 Limits 写入 Bootstrap.session_limits（request_timeout_ms 为毫秒），与 phase_limits 一样由 Core 填写；Activate 用同一 Limits 的 SessionMaxInflight、SessionMaxPayload、RequestTimeout 设置 Core 侧 wire.Options 的 MaxInflight、MaxBytes、DefaultDeadline。
- wire：Options 的 MaxInflight、MaxBytes、DefaultDeadline 为零时改用 tune.Defaults() 的对应值（只供测试与零值兜底）；PriorityFrames 不是调优参数，保持原样。
- SDK：prepare 读取 Bootstrap.session_limits，activate 用它设置插件侧 wire.Options 的 MaxInflight、MaxBytes、DefaultDeadline。session_limits 缺失或任一字段为 0 时，按 CONFIG 阶段以 INVALID_ARGUMENT 的 SESSION_ERROR 失败。SDK 的 handshakeTimeout 与 onStopBudget 在收到 Bootstrap 之前或断链时使用，不在本切片改动。
- supervisor：MemorySoftLimit > 0 时，为插件进程的环境增加 GOMEMLIMIT=<字节数>（十进制整数，Go 运行时按字节解释）；0 时不设置。
- 日志级别、日志文件上限、协议帧上限、Secret 与票据期限等 §16 列为不可调的值保持原处，不进入参数表。

测试（按编码规范「测试提交规则」）：
- 提交：contracts/wirespec_test.go 追加协议编号检查：SessionLimits 字段 1～3 的名称与编号、Bootstrap 字段 11 为 session_limits。不提交其他测试。
- 本地自测（*_local_test.go，不提交，PR 中逐项写明结果）：
  - tune：Defaults 与 CONTRACTS §16 逐项一致；每种单位的合法与非法写法（含 1.5MiB、-1s、KB/MB/GB 的提示）；直接参数中的多项与同项重复；文件中的空行、注释、缺少 "="、行号；跨项后者覆盖前者；每条相互约束的边界两侧；Override 只保留最后一次且按参数表顺序；WriteList 输出全部 37 个参数及生效值。
  - cmd/alonabot（构建真实二进制与 examples/plugins/lifecycle）：直接参数与文件各给一次 REQUEST_TIMEOUT 时以文件中的值生效，stderr 有一条列出生效值与文件行号的 "tuning parameters overridden" WARN；--args SESSION_MAX_PAYLOAD=100MB、未知键、CONTINUATION_MAX=256 各以 exitUsage 退出、不启动插件且 stderr 指明出处；alonabot run ... --args X 与 alonabot --args X run ... 效果相同；alonabot --args REQUEST_TIMEOUT=2s tune list 以 0 退出并显示生效值 2s；未知子命令与缺少必填参数以 exitUsage 退出；现有 run 参数的写法不变。
  - supervisor：用回显环境与 Bootstrap 的测试辅助插件，确认 MemorySoftLimit=256<<20 时插件收到 GOMEMLIMIT=268435456、为 0 时没有 GOMEMLIMIT；以 SESSION_MAX_INFLIGHT=2,REQUEST_TIMEOUT=2s 解析的 Values 派生 Limits 时，插件收到 max_inflight=2、request_timeout_ms=2000。
  - host/packages：NewLimits、NewStopTiming、NewRestartPolicy 与 packages.NewLimits 取值正确；SDK 在 session_limits 缺失或含 0 时以 INVALID_ARGUMENT 失败。
- 现有测试必须继续通过；引用 DefaultLimits、DefaultStopTiming、NewRestartPolicy() 与被删除 wire 常量的测试改为 tune.Defaults() 派生的值。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go mod tidy 后 go.mod、go.sum 无差异；go vet ./...；go test -race ./internal/... ./sdk/... ./contracts/... ./cmd/...；go build ./...。

不做：alonabot serve、库模式的 alonabot.Args 与 ALONABOT_ARGS（分别由 T-09.4、T-13.1 复用 tune.Parse 接入）；尚无调用方的参数（DISPATCH_*、BUS_*、MEDIA_*、JOB_TIMEOUT 等）只登记不接线；运行中修改参数；WebUI。
```

## 第 3 波

### goal-2 · T-03.1 首个可用 SDK 切片

状态：已完成（[PR #29](https://github.com/AlonaBot/AlonaBot/pull/29)，合并提交 `c79d723`）。真实 TCP/stdio echo 与三项评审修复验证通过；合并后 main 的 Quality gate 已通过，证据见 TEST-MATRIX §3.2。

```text
/goal 使用 $alonabot-execute 推进 goal-2（切片 T-03.1：首个可用 SDK 切片）。goal 目录 goal-2，分支 goal-2，worktree .worktrees/goal-2，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件能注册 message.created 事件路由；Core 把标准事件经真实 Session 派发给已 Activated 的插件；插件在 Handler 里用 Context.Send 回复原事件 Target；Core 经 Adapter 接口发出，并把 SendResult（以及可能同时存在的错误）返回给插件。

依据：PLAN T-03 步骤 1；CONTRACTS §1.2、§8、§14.1；MESSAGE-MODEL §1、§2、§3（本切片只做手动发送签名）、§4；WIRE-FORMAT §1、§4、§5A。

范围（只新增或修改这些；.proto、internal/core/state、cmd/alonabot 和所有文档都不改）：
- 新增 internal/wire/services.go
- 新增 internal/core/adapter/ 和 internal/core/adapter/memory/
- 新增 internal/core/router/ 和 internal/core/message/
- 修改 internal/core/host/activate.go：Instance 增加 Routes 字段，在 Activate 中从 prep.Register 赋值
- sdk/：新增 events.go、dispatch.go；修改 base.go（Root 增加 Events Hook 与 Base 的默认实现）、sdk.go（Context、Send、Registration、Collect 与包文档的注册项列表）、activate.go（业务 Handler 与当前 Session）、session.go（buildRegister 输出事件路由）
- 新增 examples/plugins/echo/
- 新增 internal/core/supervisor/echo_test.go
- 上述包的测试文件
允许的新跨模块边：无（新包都在 internal/core 或 sdk 模块内部；examples → sdk 已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/wire/services.go
   const ServiceSend = "core.send"
   // Core 固定服务名不含 "/"，不会与 plugin_id/route_id 形式的目标冲突。
2. internal/core/adapter
   var ErrQueueFull = errors.New("Core event queue is full")
   type Sink interface {
       Deliver(ctx context.Context, ev *pb.Event) error // 只做接纳，不等待 Handler；满时返回 ErrQueueFull
   }
   type Instance interface {
       AdapterID() string
       InstanceID() string
       AccountID() string
       Capabilities() []*pb.Capability
       Start(ctx context.Context, sink Sink) error // 完成凭证与冲突检查、启动接收后返回
       Stop(ctx context.Context) error
       Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) // 两者可以同时非 nil
   }
   type Registry struct{ ... }
   func NewRegistry() *Registry
   func (r *Registry) Add(in Instance) error // instance_id 重复返回错误
   func (r *Registry) Get(instanceID string) (Instance, bool)
3. internal/core/adapter/memory（测试与示例用，不进入生产装配；并发安全）
   type Sent struct { Target *pb.Target; Chain *pb.MessageChain }
   type Adapter struct{ ... } // 实现 adapter.Instance
   func New(adapterID, instanceID, accountID string) *Adapter
   func (a *Adapter) Inject(ctx context.Context, ev *pb.Event) error // 经 Start 收到的 Sink 投递；未 Start 时返回错误
   func (a *Adapter) Sent() []Sent
   func (a *Adapter) SetSendFunc(fn func(*pb.Target, *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo))
   默认 Send：先记录，再返回 ACCEPTED；每个 Segment 一个 ACCEPTED PartResult，platform_message_ids 依次为 "m1"、"m2"……
4. internal/core/router
   type Caller struct { PluginID string; GenerationID string; SessionKey string }
   // SessionKey 每次启动唯一，标识这一次 Session；之后的媒体所有权等按它绑定。
   func NewCaller(pluginID, generationID string) Caller // SessionKey 为 16 字节随机值的十六进制
   type Invocation struct { PluginID string; Event *pb.Event; DefaultTarget *pb.Target }
   // Core 记录的可信调用上下文：派发时放进 wire.Request.Attach，插件的子请求经 Inbound.ParentAttach 取回。
   type ServiceFunc func(ctx context.Context, c Caller, in *wire.Inbound) wire.Response
   type Mux struct{ ... }
   func NewMux() *Mux
   func (m *Mux) Handle(name string, fn ServiceFunc) // 同名重复注册时 panic
   func (m *Mux) For(c Caller) wire.Handler          // 作为 supervisor.Spec.Handler；未注册的 Target 返回 NOT_FOUND
   type Plugin interface { // 一个可以接收派发的插件实例（Target 一词专指发送目标，这里不用）
       PluginID() string
       Activated() bool
       Routes() []*pb.RouteDescriptor
       Call(ctx context.Context, req wire.Request) (wire.Response, error)
   }
   func FromInstance(in *host.Instance) Plugin
   type Router struct{ ... }
   func New(log *slog.Logger, queueEvents int) *Router // log 为 Core 的 Logger（logging.Output.Logger）；queueEvents 取 tune.Values.DispatchQueueEvents；本切片测试与示例用 tune.Defaults() 派生，已有启动路径沿用 tune.Parse 的生效值
   func (r *Router) Attach(p Plugin)
   func (r *Router) Detach(pluginID string)
   func (r *Router) Start() // 启动单个派发 worker
   func (r *Router) Close() // 停止 worker；之后 Deliver 返回错误
   func (r *Router) Deliver(ctx context.Context, ev *pb.Event) error // 实现 adapter.Sink
5. internal/core/message
   type Service struct{ ... }
   func New(adapters *adapter.Registry) *Service
   func (s *Service) Send(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response // 注册为 Mux 的 wire.ServiceSend
6. sdk
   Root 增加 Hook：Events(*EventRegistry) // 在 Collect 时调用一次，注册事件路由；与 Commands、HTTP、Schedules 并列
   func (*Base) Events(*EventRegistry)       // 默认不注册事件路由
   type EventRegistry struct{ ... }
   func (r *EventRegistry) On(routeID, eventType string, handler func(*Context) error)
   type EventRoute struct { RouteID string; EventType string; Handler func(*Context) error }
   Registration 增加 Events []EventRoute；Collect 调用 root.Events，错误与其他注册项一起汇总。
   func (c *Context) Event() *pb.Event
   func (c *Context) Logger() *slog.Logger // slog.Default() 附带本次调用的 trace_id（十六进制）、route_id 与 call_id
   Context.Send 由返回 UNSUPPORTED 改为真实实现，签名不变。
   func currentSession() (*wire.Session, error) // 未导出：返回已激活的 Session；未就绪或已停止时返回 Code=FAILED_PRECONDITION 的 *sdk.Error。后续的后台 Send、媒体客户端都通过它取 Session，不再改 activate.go

行为：
- Router.Deliver 放入容量为 queueEvents 的队列（缺省 DISPATCH_QUEUE_EVENTS=256），满时立即返回 adapter.ErrQueueFull。单 worker 与只按条数限额是临时实现，用一行 shortcut 注释写明：有界并发派发（DISPATCH_CONCURRENCY、DISPATCH_QUEUE_BYTES）接入后替换。
- worker 逐条处理事件：
  - 候选是 Activated()==true 的 Plugin 上 kind=EVENT 且 event_type_filter 包含该事件类型的路由，按 (priority, plugin_id, route_id) 升序依次调用。
  - 每次调用发送 Request{Target: route_id, Payload: PROTOBUF 编码的 HandlerRequest（schema_id 用 validate.SchemaID，version 1）, TraceID: 16 字节随机值, Attach: &Invocation{PluginID, Event, DefaultTarget: ev.Target}}。HandlerRequest.context 只设置 actor 与 default_target。
  - 响应解码为 HandlerOutcome：propagation=STOP 时停止后续候选；调用错误或带 Err 的响应以 WARN 记录（带 plugin_id、route_id、trace_id）；仍解码有效 HandlerOutcome，显式 STOP 优先于业务错误，否则继续下一个候选。不重试。
- Mux.For 返回的 Handler 按 Inbound.Target 查找服务；Caller 由 For 的参数固定，不信任载荷里的任何身份。
- message.Service.Send 按顺序检查，1～5 的拒绝都同时返回 status=REJECTED 的 SendResult（已知 target 时填上）和 ErrorInfo：
  1) 载荷必须是 PROTOBUF、schema alonabot.send_request.v1、version 1，否则 INVALID_ARGUMENT。
  2) 默认目标取 in.ParentAttach 中的 *router.Invocation（没有父调用时为空）。wire 已保证父调用属于同一 Session，这里不再重复核对插件身份。
  3) 没有 To 也没有默认目标 → NO_DEFAULT_TARGET。有 To 但与默认目标不相等（proto.Equal）→ PERMISSION_DENIED，并用 shortcut 注释写明：显式目标的许可检查接入后替换。
  4) Chain 为空或只有元数据 → INVALID_ARGUMENT；validate.Chain 失败 → INVALID_ARGUMENT。
  5) 实例不存在 → UNAVAILABLE；target.adapter_id 或 account_id 与实例不符 → INVALID_ARGUMENT。
  6) 调用 Instance.Send，把 SendResult 作为 PROTOBUF 载荷（schema alonabot.send_result.v1）连同 ErrorInfo 一起返回，两者可以同时存在。不自动重试。
- SDK 业务分发（替换 businessHandler）：
  - 按 Inbound.Target 查找事件路由：找不到返回 NOT_FOUND；载荷不是 HandlerRequest、或 route_id 与 Target 不一致返回 INVALID_ARGUMENT；命令路由仍返回 UNSUPPORTED。
  - 执行 Handler 时 recover panic，经 logPanic 记为带 stack 的 ERROR，转为 INTERNAL。Handler 返回后 Context 失效，之后再调用 Send 返回 FAILED_PRECONDITION。
  - 成功时返回 HandlerOutcome{matched: true, propagation: 调用过 StopPropagation 为 STOP，否则 CONTINUE}。
  - Handler 返回 *sdk.Error 时，Err 使用它的 Code 与 Message；其他错误为 INTERNAL，safe_message 为 "handler returned an error"，原文经 Context.Logger 以 ERROR 记录。出错时仍返回 HandlerOutcome，保留 STOP。
- Context.Send：构造 SendRequest{to: To 选项, chain: chain.Proto()}，以本次 Inbound.CallID 为 ParentCallID 调用 wire.ServiceSend。有载荷就解码成 SendResult；有 Err 时返回 *sdk.Error（Code 经 validate.PublicCode）。结果与错误可以同时返回。
- EventRegistry.On 的以下情况都记为 Collect 错误：routeID 为空；与已注册的事件或命令 route_id 重复；eventType 不是 message.proto 中列出的 8 种之一；handler 为 nil。
- buildRegister 为每个事件生成 RouteDescriptor{route_id, kind=EVENT, default_enabled=true, required_permission=ANY, event: EventSpec{event_type_filter: [eventType], handler_mode=MANUAL}}，排在命令路由之后。
- examples/plugins/echo：根插件嵌入 sdk.Base，覆盖 Events 注册 On("echo", "message.created", ...)，把收到消息中的 text 片段原样回复；没有 text 片段时不回复。

测试（按编码规范「测试提交规则」：只提交集成测试；本地自测写在 *_local_test.go，不提交，PR 中写明结果）：
- 提交：internal/core/supervisor/echo_test.go（//go:build unix）。按 supervisor_test.go 构建示例的方式构建 examples/plugins/echo，用 memory Adapter + adapter.Registry + message.Service + Mux + Router 装配；Spec.Log 与 Router 的 Logger 取自写到测试缓冲区的 logging.Output。Launch 后 Attach 实例。
  - Inject 一条 message.created：memory 收到一条 Chain，文本与原文相同，Target 等于原 Target。
  - Inject 一条没有匹配路由的事件类型：memory 没有收到任何发送。
  - TCP 与 stdio 各跑一次。
- 本地自测：message.Service 的各条拒绝路径（无目标、非原目标、空 Chain、未知实例）和「结果与错误同时返回」；Router 的排序与 STOP、跳过未激活的 Plugin、超过 queueEvents 的那一条返回 ErrQueueFull、Mux 未知服务；SDK 的 EventRegistry 拒绝情况、buildRegister 输出、Handler 返回后 Send 失败、STOP 与错误并存、panic 转 INTERNAL。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./internal/wire/... ./sdk/... ./examples/...；go build ./...。

不做：Telegram、显式目标许可、后台 Send、媒体、按钮、命令分发、Rule、Hook、并发派发、cmd/alonabot 装配。
```

## 第 4 波

### goal-3 · T-03.2 Target 与发信许可

状态：已完成（[PR #31](https://github.com/AlonaBot/AlonaBot/pull/31)，合并提交 `33426cd`）。证据见 TEST-MATRIX §3.3。条目保留供后继切片查阅接口。

```text
/goal 使用 $alonabot-execute 推进 goal-3（切片 T-03.2：Target 与发信许可）。goal 目录 goal-3，分支 goal-3，worktree .worktrees/goal-3，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件可以向自己有许可的实例发送。显式 To 与后台 Send 都按 Core 持久化的 allowed_adapter_instances 逐次准入；回复原事件 Target 仍然不需要许可。提供 Core 内部的许可读写入口，供测试和后续管理服务调用，不开放外部接口。

依据：PLAN T-03 步骤 3；CONTRACTS §8 第 7、8 条；MESSAGE-MODEL §1；TEST-MATRIX V-07。

范围：
- 修改 internal/core/state/state.go：只在 migrations 末尾追加一条
- 新增 internal/core/state/permissions.go
- 新增 internal/core/policy/
- 修改 internal/core/message/：用 policy 替换「只允许默认目标」的临时规则，删除对应 shortcut 注释；同步更新 message.New 的全部调用者（包括 internal/core/supervisor/echo_test.go）
- 新增 sdk/send.go（通过 currentSession 取 Session，不修改 sdk/activate.go）
- 上述包的测试文件
允许的新跨模块边：无。

接口：
// T-03.1 已合入的服务接口：
// router.Caller { PluginID string; GenerationID string; SessionKey string }
// router.NewCaller(pluginID, generationID string) Caller
// type router.ServiceFunc func(context.Context, router.Caller, *wire.Inbound) wire.Response
// router.NewMux() *Mux；(*Mux).Handle(name string, fn ServiceFunc)；(*Mux).For(c Caller) wire.Handler
// sdk.currentSession() (*wire.Session, error)：未就绪或停止时为 FAILED_PRECONDITION 的 *sdk.Error；保持 sdk/activate.go 不变。
// router.Invocation { PluginID string; Event *pb.Event; DefaultTarget *pb.Target }；message.New(adapters *adapter.Registry) *Service（下文改为增加 policy 参数）。
// internal/core/state/permissions.go
func (s *Store) AllowedAdapterInstances(ctx context.Context, pluginID string) ([]string, error) // 按字典序返回
func (s *Store) GrantAdapterInstance(ctx context.Context, pluginID, instanceID string) error    // 幂等
func (s *Store) RevokeAdapterInstance(ctx context.Context, pluginID, instanceID string) error   // 幂等
迁移：CREATE TABLE plugin_adapter_permissions (plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), adapter_instance_id TEXT NOT NULL, PRIMARY KEY (plugin_id, adapter_instance_id));
// internal/core/policy
type Policy struct{ ... }
func New(store *state.Store) *Policy
func (p *Policy) CheckSend(ctx context.Context, pluginID string, inv *router.Invocation, target *pb.Target) *pb.ErrorInfo // nil 表示放行
// internal/core/message：构造函数改为
func New(adapters *adapter.Registry, p *policy.Policy) *Service
// sdk
func Send(ctx context.Context, to *pb.Target, chain message.Chain) (*pb.SendResult, error) // 后台发送，不关联任何入站调用

行为：
- message.Service.Send 中确定目标的步骤改为：有 To 用 To，否则用默认目标；两者都没有 → NO_DEFAULT_TARGET。然后调用 policy.CheckSend：
  a) inv 非 nil、inv.PluginID 等于调用插件、且 target 与 inv.DefaultTarget 用 proto.Equal 比较相等 → 放行。
  b) 否则 target.adapter_instance_id 必须在该插件的许可集合中，不在 → PERMISSION_DENIED。
- 每次 Send 都查询状态库，不缓存，撤销在下一次 Send 时生效。许可只是发信许可，不赋予 User/Admin。
- 实例存在性与 adapter_id/account_id 绑定检查保持不变，放在许可检查之后。Core 全局禁止规则尚无实现，本切片不新增。
- sdk.Send：
  - 没有已激活的 Session（尚未就绪或已停止）时返回 FAILED_PRECONDITION。
  - 不设 ParentCallID。
  - to 为 nil 时照样发出，由 Core 返回 NO_DEFAULT_TARGET。
  - Context.Send 保持原样。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/message/permissions_test.go。用临时 data_dir 的真实状态库、真实 policy 与 message.Service，加上 memory Adapter，做表驱动测试：
  - 回复原 Target 不需要许可。
  - 显式发往其他目标：无许可被拒；授权后放行；撤销后下一次 Send 被拒。
  - 许可按实例生效：实例 A 的许可不放行实例 B；同一实例下不同 chat/thread 都放行。
  - thread_id 不同就不是原 Target，需要许可。
  - 没有默认目标也没有 To 时返回 NO_DEFAULT_TARGET。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：state 许可方法的读写往返；sdk.Send 在就绪前返回 FAILED_PRECONDITION。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：管理 API/CLI/WebUI、Core 全局禁止规则、Telegram、媒体。
```

### goal-4 · T-03.3 Adapter 进程、sdk/adapter 与上下线事件

状态：已完成（[PR #32](https://github.com/AlonaBot/AlonaBot/pull/32)，合并提交 `db3c1c7`）。证据见 TEST-MATRIX §3.2。条目保留供后继切片查阅接口。

```text
/goal 使用 $alonabot-execute 推进 goal-4（切片 T-03.3：Adapter 进程、sdk/adapter 与上下线事件）。goal 目录 goal-4，分支 goal-4，worktree .worktrees/goal-4，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 按安装记录以 Adapter 角色启动 Adapter 进程，复用插件的握手、准备、激活、停止与有限重启；REGISTER 按角色互斥校验。进程内的每个账号实例包装为 Core 的 adapter.Instance，经 adapter.instance.start/stop 单独启停、经 adapter.send 发送；Adapter 经 core.adapter.deliver/cursor/status 交付事件、保存游标与报告致命错误。实例进入和离开运行状态时由 Core 生成 bot.online/bot.offline。提供 sdk/adapter 与 sdk.RunAdapter，并用回环测试 Adapter 验证完整链路。

依据：PLAN T-03 步骤 2；CONTRACTS §8（Adapter 段与 bot.online/offline 段）、§4.1、§12；WIRE-FORMAT §3.4、§5A；MESSAGE-MODEL §1、§1.1；DECISIONS D-26；TEST-MATRIX V-48、V-51。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：新增 adapter.proto；route.proto 的 Register 追加 AdapterDescriptor adapter = 7；message.proto 的 EventPayload 追加 bot_online = 9、bot_offline = 10 及其消息与枚举，并更新 Event 上 event_type 取值的注释。随后 go generate ./...（本切片获授权修改这些 proto，只追加字段与消息，不改已有编号）
- contracts/validate/：新增 adapter.go（Adapter 角色的 REGISTER 校验）；修改 register.go，让插件角色的 REGISTER 拒绝 adapter 描述
- contracts/wirespec_test.go：追加新消息与新字段的编号检查
- internal/wire/services.go：末尾追加 Adapter 服务名常量
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/adapter.go
- internal/core/host/prepare.go：Expect 增加角色，Prepare 按角色选择 REGISTER 校验
- internal/core/supervisor/supervisor.go：Spec 增加角色字段并传给 host.Expect
- 修改 internal/core/adapter/adapter.go：Registry 新增 Remove，支持停用与进程退出时撤销登记；保持 Instance、Sink 接口不变
- 新增 internal/core/adapter/remote/
- sdk/：新增 adapter_run.go（RunAdapter）；sdk/events.go 的事件类型白名单加入 bot.online、bot.offline
- 新增 sdk/adapter/
- 上述包的测试文件
允许的新跨模块边：无（internal/core/adapter/remote 在 internal/core 内；sdk -> sdk/adapter 属同一模块）。sdk/adapter 只 import contracts，不 import sdk，避免循环。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// T-03.1 已合入的服务接口：
// router.Caller { PluginID string; GenerationID string; SessionKey string }
// router.NewCaller(pluginID, generationID string) Caller
// type router.ServiceFunc func(context.Context, router.Caller, *wire.Inbound) wire.Response
// router.NewMux() *Mux；(*Mux).Handle(name string, fn ServiceFunc)；(*Mux).For(c Caller) wire.Handler
// adapter.NewRegistry() *Registry；(*Registry).Add(in Instance) error；(*Registry).Get(instanceID string) (Instance, bool)；wire.ServiceSend = "core.send"。
// 本 goal 新增：func (r *Registry) Remove(instanceID string)；并发安全，移除不存在的 ID 无副作用。
1. contracts/proto（adapter.proto，包 alonabot.v1）
   message AdapterDescriptor { string adapter_id = 1; repeated Capability capabilities = 2; repeated ExtensionDeclaration extensions = 3; }
   message ExtensionDeclaration { string name = 1; uint32 version = 2; string schema_json = 3; } // 本切片只随描述保存，校验在 T-03.8
   message AdapterInstanceStartRequest { string instance_id = 1; string config_json = 2; bytes cursor = 3; }
   message AdapterInstanceStartResponse { string account_id = 1; optional string account_display_name = 2; repeated Capability capabilities = 3; }
   message AdapterInstanceStopRequest { string instance_id = 1; }
   message AdapterInstanceStopResponse {}
   message AdapterSendRequest { string instance_id = 1; Target target = 2; MessageChain chain = 3; } // 响应载荷为 SendResult
   message AdapterDeliverRequest { string instance_id = 1; Event event = 2; bytes cursor = 3; }
   message AdapterDeliverResponse { bool accepted = 1; }
   message AdapterCursorRequest { string instance_id = 1; bytes cursor = 2; }
   message AdapterCursorResponse {}
   enum AdapterInstanceState { ADAPTER_INSTANCE_STATE_UNSPECIFIED = 0; ADAPTER_INSTANCE_STATE_FAILED = 1; }
   message AdapterStatusRequest { string instance_id = 1; AdapterInstanceState state = 2; ErrorInfo safe_error = 3; }
   message AdapterStatusResponse {}
   message.proto：
   message BotOnlinePayload { string adapter_id = 1; string adapter_instance_id = 2; string account_id = 3; optional string account_display_name = 4; }
   enum BotOfflineReason { BOT_OFFLINE_REASON_UNSPECIFIED = 0; BOT_OFFLINE_REASON_STOPPED = 1; BOT_OFFLINE_REASON_ADAPTER_EXITED = 2; BOT_OFFLINE_REASON_INSTANCE_FAILED = 3; }
   message BotOfflinePayload { string adapter_id = 1; string adapter_instance_id = 2; string account_id = 3; BotOfflineReason reason = 4; ErrorInfo safe_error = 5; } // safe_error 只在 INSTANCE_FAILED 时设置
   所有载荷用 PROTOBUF 编码，schema_id 用 validate.SchemaID，version 1。
2. contracts/validate
   func AdapterRegister(r *pb.Register, adapterID string) error
   // adapter 描述必须存在且 adapter_id 等于 adapterID；routes、http_listener 必须为空，migration 为 NONE；capability 名不重复；extension 的 (name, version) 不重复且 schema_json 是 JSON 对象。错误为带 Code 的 *validate.Error。
   Register(r) 增加：r.adapter 非空时返回 INVALID_ARGUMENT。
3. internal/wire/services.go（追加在末尾）
   const (
       ServiceAdapterInstanceStart = "adapter.instance.start"
       ServiceAdapterInstanceStop  = "adapter.instance.stop"
       ServiceAdapterSend          = "adapter.send"
       ServiceCoreAdapterDeliver   = "core.adapter.deliver"
       ServiceCoreAdapterCursor    = "core.adapter.cursor"
       ServiceCoreAdapterStatus    = "core.adapter.status"
   )
4. internal/core/state
   迁移：ALTER TABLE installs ADD COLUMN role TEXT NOT NULL DEFAULT 'plugin'; ALTER TABLE installs ADD COLUMN adapter_id TEXT NOT NULL DEFAULT ''; CREATE TABLE adapter_cursors (instance_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), cursor BLOB NOT NULL, updated_at INTEGER NOT NULL);
   type Role string // RolePlugin = "plugin"、RoleAdapter = "adapter"
   Install 增加字段：Role Role（为空时按 RolePlugin 写入）；AdapterID string（RoleAdapter 时必填）
   AddInstall/GetInstall 读写这两个字段；RoleAdapter 且 adapter_id 已被另一条安装记录占用时 AddInstall 返回错误（同一 adapter_id 只能安装一个包）
   func (s *Store) AdapterCursor(ctx context.Context, instanceID string) ([]byte, error) // 没有记录时返回 nil, nil
   func (s *Store) PutAdapterCursor(ctx context.Context, pluginID, instanceID string, cursor []byte) error
5. internal/core/host
   Expect 增加字段：Role state.Role; AdapterID string // RoleAdapter 时 Prepare 用 validate.AdapterRegister(r, AdapterID)，否则用 validate.Register(r)
6. internal/core/supervisor
   Spec 增加字段：Role state.Role; AdapterID string // 原样传给 host.Expect
7. internal/core/adapter/remote
   // Process 代表一个 Adapter 包进程及其实例；进程重启后用 Bind 绑定新 Session。方法可以并发调用。
   type Process struct{ ... }
   type Options struct {
       AdapterID string
       PluginID  string            // Adapter 包的安装 ID
       Store     *state.Store
       Registry  *adapter.Registry // 已启动的实例以 adapter.Instance 登记在这里，供 core.send 使用
       Sink      adapter.Sink      // 平台事件与 bot.online/offline 都交给它
       Log       *slog.Logger
   }
   func New(opt Options) *Process
   func (p *Process) Services(m *router.Mux) // 注册 core.adapter.deliver/cursor/status；只接受 Caller.PluginID 等于 PluginID 的调用
   func (p *Process) Bind(ctx context.Context, in *host.Instance) error // 绑定新激活的 Session，并重新启动全部已启用的实例
   func (p *Process) Unbind() // 进程退出或停止后调用：运行中的实例全部离开运行状态（ADAPTER_EXITED），保持启用
   func (p *Process) StartInstance(ctx context.Context, instanceID string, configJSON []byte) error
   func (p *Process) StopInstance(ctx context.Context, instanceID string) error // 停用：STOPPED，之后不再随 Bind 恢复
8. sdk/adapter
   type Account struct { ID string; DisplayName string } // DisplayName 为空表示平台没有给出
   type StartRequest struct { InstanceID string; ConfigJSON []byte; Cursor []byte }
   // Core 是 SDK 交给一个实例的 Core 客户端，只能为该实例交付；Deliver 返回 Core 是否接纳。
   type Core interface {
       Deliver(ctx context.Context, ev *pb.Event, cursor []byte) (accepted bool, err error)
       SaveCursor(ctx context.Context, cursor []byte) error
       ReportFailure(ctx context.Context, safe *pb.ErrorInfo) error // 实例不再接收；之后 Deliver 返回错误
   }
   type Instance interface {
       Account() Account
       Capabilities() []*pb.Capability
       Send(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)
       Stop(ctx context.Context) error
   }
   type Adapter interface {
       Descriptor() *pb.AdapterDescriptor
       // Start 完成凭证与账号检查、启动接收后返回；core 在实例停止前一直有效。
       Start(ctx context.Context, req StartRequest, core Core) (Instance, error)
   }
9. sdk
   func RunAdapter(a adapter.Adapter) error // Adapter 进程 main 的入口，与 Run 共用取锁、握手、准备、激活与停止

行为：
- 角色由安装记录决定：supervisor 按 Spec.Role 设置 host.Expect，Adapter 的 REGISTER 只经 validate.AdapterRegister，插件的只经 validate.Register，不一致按注册失败处理（SESSION_ERROR，阶段 REGISTER）。进程不自报角色。
- RunAdapter：与 Run 相同地取 runtime.lock、握手、在 OnStart 预算位置不执行用户代码，REGISTER 只携带 a.Descriptor()、原样回传 Bootstrap 的 candidate_config_json 与 target_config_version、migration 为 NONE；激活后把 adapter.instance.start/stop、adapter.send 交给对应实例。收到 STOPPING 或断链时，在 STOP_ONSTOP 预算内依次 Stop 全部实例后返回。Adapter 包级配置类型不在本切片。
- adapter.instance.start（Core 发起，期限 REQUEST_TIMEOUT）：
  - Core 先从 adapter_cursors 读出游标放进请求；SDK 调用 Adapter.Start，失败时返回其错误（不含配置原文）。
  - 响应的 account_id 为空 → Core 判定失败；同一 Process 中已有运行中的实例使用同一 account_id → Core 立即对新实例发 adapter.instance.stop，StartInstance 返回 ALREADY_EXISTS。
  - 成功后，以 adapter.Instance 登记到 Registry（AdapterID、InstanceID、AccountID、Capabilities 取自响应），标记为运行中与已启用，并生成 bot.online。
- StopInstance：发 adapter.instance.stop，从 Registry 移除，标记为未启用，生成 reason=STOPPED 的 bot.offline。实例不存在返回 NOT_FOUND。停一个实例不影响同进程的其他实例。
- adapter.send：Registry 中的远端实例 Send 发起 adapter.send，原样返回 SendResult 与 ErrorInfo（两者可以同时存在）；Session 不可用时返回 UNAVAILABLE。不自动重试。
- core.adapter.deliver，按顺序：
  1) Caller 不是本 Process 的 PluginID → PERMISSION_DENIED。
  2) instance_id 不是本 Session 中运行中的实例，或 event 的 adapter_id、adapter_instance_id、account_id 与该实例不一致，或 event.target 存在且这三项不一致 → INVALID_ARGUMENT，以 WARN 记诊断（含 plugin_id、instance_id，不含载荷），不写游标。
  3) 写入游标（PutAdapterCursor）；写失败返回 INTERNAL 且不接纳。
  4) Core 重写 event_id（16 字节随机值的十六进制）、source_kind=ADAPTER、capabilities=实例的 Capability；Actor 的 trusted_roles 只保留 Adapter 给出的 CHAT_ADMIN 事实，去掉 USER 与 ADMIN（由 Core 策略判定，T-04.4），然后交给 Sink.Deliver。
  5) 返回 accepted：Sink 返回 adapter.ErrQueueFull 时为 false，并递增该实例的丢弃计数，以 WARN 记录（含 plugin_id、instance_id 与累计计数）；游标已推进，不回退。
- core.adapter.cursor：同 1)、2) 的实例检查后写入游标。core.adapter.status：state=FAILED 时把实例移出 Registry、标记为未启用（不随 Bind 自动恢复），生成 reason=INSTANCE_FAILED、safe_error 为请求中 safe_error 的 bot.offline。
- Bind：记录新 Session 后，按 StartInstance 的同一流程重新启动全部「已启用」实例（实例配置由 StartInstance 时保存在内存中），游标取自状态库。Unbind：运行中的实例全部移出 Registry、生成 reason=ADAPTER_EXITED 的 bot.offline，保持已启用。Core 重启后内存中没有实例，不补发任何事件；实例配置的持久化由 T-09.4 负责。
- bot.online/bot.offline 事件：event_type 为 "bot.online"/"bot.offline"，source_kind=ADAPTER，adapter_id/adapter_instance_id/account_id 取实例，target、actor、message 为空，event_id 由 Core 生成，经 Sink.Deliver 交付；队列满时丢弃并计数，不重试。
- 实例配置 JSON（含凭证）只出现在 adapter.instance.start 的请求中，不写日志、不进入事件或错误文本。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/adapter/remote/remote_test.go（//go:build unix）：测试二进制在 TestMain 中按环境变量切换为回环测试 Adapter，经 sdk.RunAdapter 运行。回环实例的 Send 把收到的 text 原样作为该实例的 message.created 交付回 Core（游标为递增序号）；文本为 "forge" 时改用另一个 instance_id 交付；文本为 "fail" 时经 ReportFailure 报告失败。用 supervisor.Launch（Role=RoleAdapter）、state.Store、adapter.Registry、记录事件的 Sink 装配：
    - 启动两个实例：各有一条 bot.online；经 Registry.Get(...).Send 发送的文本作为 message.created 回到 Sink，event_id 由 Core 生成，游标写入状态库。
    - 停用其中一个：一条 STOPPED 的 bot.offline，另一个实例继续收发，进程 PID 不变。
    - "forge" 被拒，Sink 没有收到，游标不变。"fail" 产生一条带 safe_error 的 INSTANCE_FAILED bot.offline。
    - 第二个实例返回已运行实例的 account_id 时 StartInstance 返回 ALREADY_EXISTS，Registry 中没有它。
    - 杀掉 Adapter 进程：每个运行中的实例一条 ADAPTER_EXITED；重新 Launch 并 Bind 后只恢复仍启用的实例，Start 收到的游标等于状态库中的值。
    - 插件角色的进程发出带 adapter 描述的 REGISTER、Adapter 角色发出插件路由，都在 REGISTER 阶段失败。
  - contracts/wirespec_test.go：adapter.proto 全部消息与枚举、Register.adapter = 7、EventPayload.bot_online = 9 与 bot_offline = 10 的编号检查。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：validate.AdapterRegister 的每种拒绝；state 新方法与迁移（旧库升级后 role 为 plugin）、同 adapter_id 的第二条安装记录被拒；deliver 的 target 不一致、Caller 不符、游标写失败；队列满时 accepted=false 且游标已推进；配置原文不出现在日志中。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：adapter.edit（T-03.9）、adapter.media.read 与出站媒体 token（T-03.5、T-03.6）、extension 校验与 adapter_dependencies（T-03.8）、显式目标许可（T-03.2）、实例配置与 Secret 的持久化及管理入口（T-08.4、T-09.4）、Telegram（T-03.4）、cmd/alonabot 装配。
```

### goal-6 · T-03.5 媒体上传与配额

状态：可派发。前置 T-03.1 已由 PR #29 合并；router.Caller/Mux 与 currentSession 已核对；配额继续使用 tune.Values。

```text
/goal 使用 $alonabot-execute 推进 goal-6（切片 T-03.5：媒体上传与配额）。goal 目录 goal-6，分支 goal-6，worktree .worktrees/goal-6，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件可以经 Core 的固定服务分块上传媒体，得到绑定本次 Session 的 MediaRef。Core 执行全部固定配额和期限，在失败、过期或 Session 结束时回收。入站媒体、Read 与 Grant 在后续切片实现。

依据：PLAN T-03 步骤 5（上传部分）；CONTRACTS §1.3、§16；WIRE-FORMAT §5；TEST-MATRIX V-09。

范围：
- 新增 internal/core/media/
- 在 internal/wire/services.go 末尾追加 3 个服务名常量
- 新增 sdk/media.go
- 上述包的测试文件
允许的新跨模块边：无。

接口：
// T-03.1 已合入的服务接口：
// router.Caller { PluginID string; GenerationID string; SessionKey string }
// router.NewCaller(pluginID, generationID string) Caller
// type router.ServiceFunc func(context.Context, router.Caller, *wire.Inbound) wire.Response
// router.NewMux() *Mux；(*Mux).Handle(name string, fn ServiceFunc)；(*Mux).For(c Caller) wire.Handler
// sdk.currentSession() (*wire.Session, error)：未就绪或停止时为 FAILED_PRECONDITION 的 *sdk.Error；保持 sdk/activate.go 不变。
// internal/wire/services.go 追加
const ServiceMediaBeginUpload = "core.media.begin_upload"
const ServiceMediaWriteChunk  = "core.media.write_chunk"
const ServiceMediaCommitUpload = "core.media.commit_upload"
// internal/core/media
type Limits struct { MaxFileBytes, MaxSessionBytes, MaxCoreBytes int64; MaxUploads int; TTL time.Duration }
func NewLimits(v tune.Values) Limits // 取 MEDIA_MAX_OBJECT、MEDIA_SESSION_BUDGET、MEDIA_DISK_BUDGET、MEDIA_MAX_UPLOADS、MEDIA_TTL 的生效值
const MaxChunkBytes = 1 << 20 // 协议固定的分块上限（CONTRACTS §1.3），不是调优参数
type Store struct{ ... }
func Open(log *slog.Logger, dir string, lim Limits, now func() time.Time) (*Store, error) // dir 是 <data_dir>/media；打开时清空上次遗留的临时文件，清理失败返回 error
func (s *Store) AddServices(m *router.Mux)             // 把上面 3 个服务加到 Mux；「Register」专指协议的 REGISTER，这里不用
func (s *Store) ReleaseSession(sessionKey string)      // Session 结束时调用：回收该 Session 的未完成上传与引用
func (s *Store) Sweep()                                // 回收所有过期对象；由调用方定期调用
func (s *Store) Lookup(token string, c router.Caller) (File, error) // 供后续发送与 Read 使用：校验 token 属于该 Session 且未过期
type File struct { Path string; Size uint64; MIME string; SHA256 []byte; ExpiresAt time.Time } // 已提交的媒体文件
// sdk
func Upload(ctx context.Context, r io.Reader, size uint64, mime string, sum []byte) (*pb.MediaRef, error)
// 按 1 MiB 分块调用三个服务；通过 currentSession 取 Session，未就绪时返回 FAILED_PRECONDITION

行为：
- Begin：
  - size 为 0 或超过 MaxFileBytes、mime 为空、sha256 不是 32 字节 → INVALID_ARGUMENT。
  - 本 Session 已有 MaxUploads 个未完成上传 → RESOURCE_EXHAUSTED。
  - 预留 size 字节：Session 合计（未完成上传加未过期引用）超过 Session 上限，或全 Core 合计超过 Core 上限 → RESOURCE_EXHAUSTED，不预留。
  - upload_id 为 16 字节随机值的十六进制；期限为 Begin 时刻起 TTL，分块请求不延长期限。
- WriteChunk：
  - upload_id 必须属于调用的 Session，否则 NOT_FOUND。
  - offset 必须等于当前已确认的长度，否则 INVALID_ARGUMENT，并终止该上传、释放预留。
  - 单块超过 MaxChunkBytes，或写入后超过声明的 size → INVALID_ARGUMENT，并终止。
  - 磁盘写入失败 → INTERNAL，并终止。
  - 成功时返回 next_offset。
- Commit：
  - 实际长度必须等于 size、SHA256 必须相符，否则 INVALID_ARGUMENT，并终止。
  - 成功后签发 MediaRef{token: 16 字节随机值的十六进制, size, mime, sha256, expires_at: Commit 时刻起 TTL}。引用绑定 owner Session，预留转为引用占用。
- 回收：上传终止、过期或 ReleaseSession 时，删除临时文件并扣回预留与占用。
  - 删除失败（不含文件已不存在）时经注入的 Logger 记一条 WARN，带 path 与 err 字段；预留与占用照常扣回，upload_id/token 照常失效。残留文件由下次 Open 清理，不重试删除。
- 过期或已回收的 upload_id/token 一律视为不存在。
- 所有拒绝都不产生可用 MediaRef。
- 预留与占用的计数在并发请求下必须正确，用锁保护，不能出现负数或超额。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/media/upload_test.go。用临时目录打开真实 Store，经 AddServices 加到 Mux 的 3 个服务（直接构造 wire.Inbound 调用）做表驱动测试：
  - 分 3 块上传后 CommitUpload，Lookup 返回正确的 File。
  - 错序 offset 终止上传并释放预留；大小或摘要不符时被拒。
  - 第 5 个未完成上传被拒；Session 和 Core 总量超限被拒（用很小的 Limits）。
  - 注入 now：超过 TTL 后 WriteChunk 返回 NOT_FOUND，预留被释放。
  - ReleaseSession 释放该 Session 的全部占用；其他 Session 用不了这个 upload_id 或 token。
  - 并发上传时配额计数正确（-race）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：sdk.Upload 在就绪前返回 FAILED_PRECONDITION；单块超过 1 MiB 被拒。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./internal/wire/... ./sdk/...；go build ./...。

不做：IncomingMedia、Read、Grant、发送时消费租约、cmd/alonabot 装配。
```

### goal-10 · T-04.1 根插件 API 与静态注册索引

状态：可派发。前置 T-03.1 已由 PR #29 合并；EventRegistry.On、Router.Attach/Detach、Plugin 与 FromInstance 已按合入代码核对。

```text
/goal 使用 $alonabot-execute 推进 goal-10（切片 T-04.1：根插件 API 与静态注册索引）。goal 目录 goal-10，分支 goal-10，worktree .worktrees/goal-10，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：完成显式注册 API：命令可声明别名与 priority，事件路由可声明 priority；同插件重复的 route_id、规范化命令路径或别名在 SDK 的 Collect 与 Core 的 REGISTER 校验两处都拒绝，整份注册原子失败。Core 为已激活插件的命令建立静态索引：跨插件同一路径按 priority/plugin_id/route_id 稳定选出唯一胜出者，并记录被遮蔽者，结果与启动顺序无关。

依据：PLAN T-04 步骤 1、3；CONTRACTS §2.1、§2.2 第 1、2 条、§4.1；DECISIONS D-20、D-39；TEST-MATRIX V-10 的注册分项、V-02 的注册原子性。

范围（只新增或修改这些）：
- sdk/registry.go：Command 增加可变参数选项；新增路由选项类型与重复检查
- sdk/events.go：On 增加可变参数选项
- sdk/session.go：buildRegister 输出别名与 priority
- contracts/validate/register.go：命令与事件路由的注册校验
- internal/core/router/：新增 index.go（命令静态索引）；router.go 中 Attach/Detach 维护索引
- 新增 internal/core/supervisor/testdata/cmdplugin/（集成测试用的插件程序）
- 上述包的测试文件
允许的新跨模块边：无（testdata 下的插件程序只 import sdk）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// T-03.1 既有：EventRegistry.On(routeID, eventType string, handler func(*Context) error)；EventRoute { RouteID string; EventType string; Handler func(*Context) error }。
// router.New(log *slog.Logger, queueEvents int) *Router；(*Router).Attach(p Plugin)；(*Router).Detach(pluginID string)；FromInstance(in *host.Instance) Plugin。
// Plugin 提供 PluginID() string、Activated() bool、Routes() []*pb.RouteDescriptor、Call(context.Context, wire.Request) (wire.Response, error)。
1. sdk
   type RouteOption func(*routeOptions)
   func Alias(path ...string) RouteOption // 命令的别名，path 为完整 token 路径（不加分组前缀）；可多次使用
   func Priority(p int32) RouteOption     // 缺省 0；数值越小越先
   func (r *CommandRegistry) Command(name string, handler any, opts ...RouteOption) // 原签名增加 opts，已有调用不变
   func (r *EventRegistry) On(routeID, eventType string, handler func(*Context) error, opts ...RouteOption)
   Command 增加字段：Aliases [][]string; Priority int32
   EventRoute 增加字段：Priority int32
2. contracts/validate
   Register(r) 在现有检查之后增加（任一失败整体拒绝）：
   - COMMAND 路由：match=PATH 时 path 非空；path 与每个 alias 的 token 非空、不含空白、不以 "/" 开头；同一 Register 中全部命令的 path 与 alias（以 token 用单个空格连接为规范形式）互不相同，冲突返回 ALREADY_EXISTS 并指明两条 route_id。
   - EVENT 路由：event_type_filter 非空，每项是 message.proto 注释列出的 event_type 之一，同一路由内不重复。
3. internal/core/router
   type CommandEntry struct {
       PluginID string
       RouteID  string
       Priority int32
       Path     []string // 命中的规范路径（路由的 path 或某个 alias）
       Alias    bool     // Path 来自 alias
   }
   type Shadow struct { Path []string; Winner CommandEntry; Shadowed []CommandEntry } // Shadowed 按胜出顺序排列
   type CommandIndex struct{ ... } // 不可变；Router 每次 Attach/Detach 后整体重建
   func (x *CommandIndex) Lookup(tokens []string) (entry CommandEntry, consumed int, ok bool) // 按最长完整 token 路径匹配，大小写敏感
   func (x *CommandIndex) Shadows() []Shadow // 按规范路径字典序
   func (r *Router) Commands() *CommandIndex // 当前已 Attach 插件的索引快照

行为：
- SDK Collect（同插件，任一错误都使 Collect 失败，Run 不连接 Core）：
  - 命令的完整路径与各别名互不相同，也不与本插件其他命令的路径或别名相同；Alias 用于事件路由、Alias 为空路径、token 含空白都是错误。
  - 命令 route_id 仍为 "cmd." + 路径以 "." 连接；它与事件路由的 route_id 重复时报错（事件路由不得使用 cmd. 前缀）。
  - 同一选项重复给出时以最后一次为准；Alias 累加。
- buildRegister：CommandSpec.aliases 按注册顺序写入 CommandAlias{path}，RouteDescriptor.priority 写入 Priority；事件路由同样写入 priority。
- 根插件与 Config 的唯一性由 API 形状保证：Run 只接受一个 Root，Root.Config 只返回一个配置；本切片不扫描源码（属于 T-04.2 生成器）。插件 ID 只来自启动环境，不从代码读取。
- Core 索引：
  - 只包含 Router 中已 Attach（即已 Activated）插件的 COMMAND 路由；REGISTER 校验失败或未激活的插件不出现在索引中（注册原子性）。
  - 每个规范路径（path 或 alias）的候选按 (priority, plugin_id, route_id) 升序排列，第一个为胜出者，其余为被遮蔽者；同一路由的 path 与 alias 分别参与。
  - 索引不看命令是否被策略禁用：禁用的命令仍占有路径（策略在 T-04.4 接入）。
  - Lookup 按 token 从长到短尝试，只匹配完整 token，不按字符串前缀匹配；没有匹配时 ok=false。
  - 本切片不改变 goal-2 的事件派发；命令的实际派发由 T-04.3 使用 Lookup 接入。
- Router 在索引重建后，以 INFO 记录每个新出现的遮蔽关系（路径、胜出的 plugin_id/route_id、被遮蔽的 plugin_id/route_id），供管理诊断使用。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/index_test.go（//go:build unix）。按 supervisor_test.go 构建示例的方式构建 testdata/cmdplugin；该程序从测试设置的环境变量读取要注册的命令（路径、别名、priority），用 sdk.Run 运行。
  - 插件 a 注册 /rss（priority 0）、b 注册 /rss（priority 0）与别名 /feed、c 注册 /rss（priority -1）：三者以不同顺序 Launch 并 Attach，Lookup("rss") 的胜出者都是 c，Shadows 中 /rss 的被遮蔽者依次为 a、b；Lookup("feed") 为 b 的别名。
  - Lookup("rss","add","x") 在 a 另注册 /rss add 时命中 /rss add 且 consumed=2；Lookup("rssx") 不命中。
  - Detach c 后 /rss 的胜出者变为 a。
  - 用一个绕过 SDK Collect 的原始插件（沿用 supervisor_test.go 的 rawPlugin 方式）发出路径与别名冲突的 REGISTER：Launch 在 REGISTER 阶段失败（ALREADY_EXISTS），Router 索引不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Collect 的每种重复与非法选项；buildRegister 的 aliases 与 priority；validate.Register 的命令与事件路由各条拒绝；CommandIndex 在相同输入不同 Attach 顺序下结果一致。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：源码扫描、注解与生成器（T-04.2）；命令派发、参数绑定与文本匹配（T-04.3）；命令策略与禁用（T-04.4）；Hook 与 Service 路由的索引（T-05.3、T-06.1）；管理诊断的 API 与页面（T-09）。
```

### goal-56 · T-11.3 Agent 会话存储

状态：可派发。前置 T-03.1 已由 PR #29 合并；router.Caller/Mux、currentSession 与 echo 装配已核对；默认 message.New(adapters) 和 wire.ServiceSend 可直接复用。

```text
/goal 使用 $alonabot-execute 推进 goal-56（切片 T-11.3：Agent 会话存储）。goal 目录 goal-56，分支 goal-56，worktree .worktrees/goal-56，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 状态库按（plugin_id、会话键）保存 Agent 会话条目；插件经 core.conversation.append/list/clear 访问自己的会话，追加原子、序号连续，list 按帧上限分页；Core 提供给管理服务层使用的列出、清除与按插件 purge；SDK 提供调用这三个服务的函数。

依据：PLAN T-11 步骤 5；CONTRACTS §13.3、§9.3（卸载与 --purge）、§1.2（帧上限）；WIRE-FORMAT §3.5、§5A；TEST-MATRIX V-56。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：新增 conversation.proto，随后 go generate ./...（本切片获授权新增该 proto 文件）
- contracts/wirespec_test.go：追加新消息的编号检查
- internal/wire/services.go：末尾追加三个服务名常量
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/conversation.go
- 新增 internal/core/conversation/
- sdk/：新增 conversation.go
- 新增 internal/core/supervisor/testdata/convplugin/（集成测试用的插件程序）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// T-03.1 已合入的服务接口：
// router.Caller { PluginID string; GenerationID string; SessionKey string }
// router.NewCaller(pluginID, generationID string) Caller
// type router.ServiceFunc func(context.Context, router.Caller, *wire.Inbound) wire.Response
// router.NewMux() *Mux；(*Mux).Handle(name string, fn ServiceFunc)；(*Mux).For(c Caller) wire.Handler
// sdk.currentSession() (*wire.Session, error)：未就绪或停止时为 FAILED_PRECONDITION 的 *sdk.Error；保持 sdk/activate.go 不变。
// message.New(adapters *adapter.Registry) *Service；(*Service).Send(context.Context, router.Caller, *wire.Inbound) wire.Response；wire.ServiceSend = "core.send"。
1. contracts/proto（conversation.proto，包 alonabot.v1）
   message ConversationAppendRequest { string conversation_key = 1; repeated string entries_json = 2; }
   message ConversationAppendResponse { uint64 last_seq = 1; }
   message ConversationListRequest { string conversation_key = 1; uint64 from_seq = 2; } // from_seq 为 0 时从 1 开始
   message ConversationEntry { uint64 seq = 1; string entry_json = 2; }
   message ConversationListResponse { repeated ConversationEntry entries = 1; optional uint64 next_seq = 2; } // 没有更多条目时 next_seq absent
   message ConversationClearRequest { string conversation_key = 1; }
   message ConversationClearResponse {}
   载荷用 PROTOBUF 编码，schema_id 用 validate.SchemaID，version 1。
2. internal/wire/services.go（追加在末尾）
   const ( ServiceConversationAppend = "core.conversation.append"; ServiceConversationList = "core.conversation.list"; ServiceConversationClear = "core.conversation.clear" )
3. internal/core/state
   迁移：CREATE TABLE conversation_entries (plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), conversation_key TEXT NOT NULL, seq INTEGER NOT NULL, entry_json TEXT NOT NULL, created_at INTEGER NOT NULL, PRIMARY KEY (plugin_id, conversation_key, seq));
   type ConversationEntry struct { Seq uint64; EntryJSON string }
   type ConversationSummary struct { Key string; Entries int; LastSeq uint64; UpdatedAt time.Time }
   func (s *Store) AppendConversation(ctx context.Context, pluginID, key string, entries []string) (lastSeq uint64, err error) // 一个事务内写入，序号接续该会话当前最大值
   func (s *Store) ListConversation(ctx context.Context, pluginID, key string, fromSeq uint64, limit int) ([]ConversationEntry, error) // seq >= fromSeq，升序，至多 limit 条
   func (s *Store) ClearConversation(ctx context.Context, pluginID, key string) error
   func (s *Store) Conversations(ctx context.Context, pluginID string) ([]ConversationSummary, error) // 按 Key 字典序
   func (s *Store) PurgeConversations(ctx context.Context, pluginID string) error
4. internal/core/conversation
   type Service struct{ ... }
   func New(store *state.Store) *Service
   func (s *Service) Register(m *router.Mux) // 注册上面三个服务
   func (s *Service) List(ctx context.Context, pluginID string) ([]state.ConversationSummary, error) // 供管理服务层
   func (s *Service) Clear(ctx context.Context, pluginID, key string) error                      // 供管理服务层
   func (s *Service) Purge(ctx context.Context, pluginID string) error                           // 卸载 --purge 时由 T-09.4 调用
5. sdk
   type ConversationEntry struct { Seq uint64; Entry json.RawMessage }
   func AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error)
   func ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error)
   func ClearConversation(ctx context.Context, key string) error
   三者经 currentSession() 发起请求，不需要父调用链；Session 未就绪时返回 FAILED_PRECONDITION 的 *sdk.Error；Core 的错误转为 *sdk.Error（Code 经 validate.PublicCode）。

行为：
- 身份：三个服务只使用 Mux 给出的 Caller.PluginID，不读取载荷中的任何插件身份；插件只能访问自己的会话。没有父调用链也可以调用。
- 校验（信任边界，在 Core 侧）：载荷编码与 schema 不符 → INVALID_ARGUMENT；conversation_key 必须是 1～256 字节的合法 UTF-8，否则 INVALID_ARGUMENT；append 的 entries_json 至少一条，每条都必须是 JSON 对象（validate.IsJSONObject），任一条不合法时整次拒绝、不写入任何条目。
- append：一个事务内按请求顺序写入，序号从该会话当前最大值加 1 连续递增（新会话从 1 开始），返回最后一条的序号。同一插件同一会话的并发 append 不产生重复或空洞的序号（在事务内读取最大值并写入）。条目原样保存，不规范化、不截断。
- list：从 max(from_seq, 1) 起按序号升序返回。逐条累加 ConversationEntry 的编码大小，下一条会使响应超过帧上限减去信封余量（命名常量 64 KiB）时停止，并把下一条的序号放进 next_seq；全部返回完时 next_seq absent。第一条总会返回（它经 append 请求进入时已在帧上限内）。从状态库分批读取，每批条数为命名常量。
- clear：删除该会话的全部条目，不存在时也成功；之后再 append 从序号 1 开始。
- 管理层 List/Clear 与 Purge 不经过插件；Purge 删除该插件的全部会话。卸载默认保留会话，是否 purge 由 T-09.4 的卸载流程决定。Core 重启后数据保留。Core 不截断、不过期、不压缩会话。
- 日志不记录条目内容。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/conversation_test.go（//go:build unix）。按 goal-2 的 echo 测试方式装配 memory Adapter、message.Service、conversation.Service、Mux 与 Router，构建 testdata/convplugin。该插件注册 message.created 事件路由，按收到的文本执行命令（例如 "append k {json}..."、"list k 0"、"clear k"），用 sdk 的三个函数调用 Core，再把结果以文本回复。
  - 一次追加三条，返回 last_seq=3；再追加一条为 4；list 返回 1～4，顺序与内容不变。
  - 追加使总大小超过帧上限的多条条目（分多次 append，每次在帧上限内），从 0 开始反复 list 直到 next_seq absent：拿到的序号连续、不丢不重，内容与写入一致。
  - 两个插件使用同一会话键，各自只看到自己的条目。
  - 关闭 Store 并用 state.Open 重新打开同一 data_dir、重新 Launch 后，条目仍在；Service.Purge 后为空，另一插件不受影响。
  - append 中含一条非对象 JSON 时整次被拒，list 不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：会话键为空、257 字节、非法 UTF-8 时被拒；并发 append 的序号无重复与空洞；clear 后从 1 重新开始；单条接近帧上限的条目在 list 中单独返回；管理层 List 的摘要字段；schema 不符被拒。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：管理 API、CLI 与页面（T-09.4、T-09.9）；卸载流程本身（T-09.4）；Agent 组件对会话的使用（T-12.2）；Agent 的记忆（T-12.4）。
```

## 第 5 波

### goal-5 · T-03.4 官方 Telegram Adapter 的接收

状态：等待前置（T-03.3）。语句中 goal-4 提供的符号（sdk/adapter 的 Adapter、Instance、Core、StartRequest，sdk.RunAdapter，remote.Process）只写名字；goal-4 合并后，审查者按实际签名核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-5（切片 T-03.4：官方 Telegram Adapter 的接收）。goal 目录 goal-5，分支 goal-5，worktree .worktrees/goal-5，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：实现官方 Telegram Adapter 的接收部分：Bot API 客户端、getMe 账号绑定、webhook 冲突检查、后台 long polling、逐条交付与游标、7 天接收轮次重建、401/409 停止实例、媒体组标识、Actor.locale、to_me 与实例代理；提供本地 Bot API 模拟服务，供本切片与后续 Telegram 切片测试。

依据：PLAN T-03 步骤 4；CONTRACTS §8（Telegram 段）、§14.5、§15.1（proxy）、§15.4；MESSAGE-MODEL §1、§1.1、§6.1、§6.2 最后一条；TEST-MATRIX V-45、V-54 的 locale 分项、V-62 的 Bot Token URL 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：identity.proto 的 Actor 追加 optional string locale = 5；message.proto 的 MessageCreatedPayload 追加 bool to_me = 3。随后 go generate ./...（本切片获授权只追加这两个字段）
- contracts/wirespec_test.go：追加这两个字段的编号检查
- 新增 adapters/telegram/（可导入的包 telegram）
- 新增 adapters/telegram/telegramtest/（本地 Bot API 模拟服务）
- 新增 adapters/telegram/cmd/alonabot-adapter-telegram/main.go（只调用 sdk.RunAdapter(telegram.New())）
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
  5) 启动后台接收循环后返回；Instance.Capabilities 本切片只声明接收相关能力，发送能力由 T-03.7 补充。
- 游标：Adapter 自定义的不透明 JSON 字节 {"offset": int64, "round_start_unix_ms": int64, "last_accepted_unix_ms": int64}；Start 收到空游标时以当前时刻为轮次起点、offset 为 0（不带 offset）。
- 接收循环：
  - 每次 getUpdates 显式传 allowed_updates=["message","edited_message","channel_post","edited_channel_post","callback_query","my_chat_member","chat_member","message_reaction"]、timeout=30、offset（为 0 时不传）。
  - 一批 update 按返回顺序逐条处理：能标准化的经 Core.Deliver 交付，游标 offset=update_id+1、last_accepted 为接纳时刻；收到应答（无论 accepted 是否为 true）后才处理下一条。不支持或本切片不处理的 update 经 Core.SaveCursor 推进游标，不交付。不得在交付第一条前保存整批最大 update_id。
  - 7 天重建：在一次 long poll 结束且没有请求在途时，若 max(round_start, last_accepted) 距今满 7 天（命名常量），先 SaveCursor 保存 {offset: 0, round_start: 现在}，再继续拉取。
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

### goal-11 · T-04.2 构建期生成器

状态：等待前置（T-04.1）。语句中 goal-10 与 goal-76 提供的符号（sdk.RouteOption、sdk.Alias、sdk.Priority、EventRegistry.On 的选项、cmd/alonabot 的 cobra 命令树）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-11（切片 T-04.2：构建期生成器）。goal 目录 goal-11，分支 goal-11，worktree .worktrees/goal-11，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：sdk/generate 在构建期扫描一个插件构建入口及其本 module 依赖包（遵守 GOOS/GOARCH 与 build tags，排除外部依赖、vendor 与测试文件），找出唯一的根类型、至多一个 Config 类型和带标记的包级 Handler，生成稳定的 Go 注册代码，由编译器检查引用；问题以带位置的诊断报告。cmd/alonabot 增加 plugin generate 与 plugin check，两者调用同一生成逻辑。

依据：PLAN T-04 步骤 2；CONTRACTS §9（生成器段）、§2.2；DESIGN「根插件」一节；DECISIONS D-20、D-39、D-40、D-42；TEST-MATRIX V-10。

范围（只新增或修改这些）：
- 新增 sdk/generate/
- sdk/：新增 generated.go（生成代码调用的注册入口）；sdk.go 的 Collect 合并生成的注册项（只加调用点）
- cmd/alonabot/：新增 plugin 命令组及 generate、check 子命令（新文件）
- 上述包的测试文件
允许的新跨模块边：cmd/alonabot -> sdk/generate（ARCHITECTURE 已登记）。sdk/generate 只用标准库，不 import sdk 的其他包。

标记语法（本切片定下，写在 sdk/generate 的包文档中）：
- 命令：紧贴包级函数声明的注释行 //alonabot:command <路径 token...> [alias=<token[/token...]>]... [priority=<int32>]；alias 用 "/" 分隔多 token 路径，可重复。
- 事件：//alonabot:event <route_id> <event_type> [priority=<int32>]。
- 标记只对包级函数有效。写在方法、类型或变量上是诊断错误（实例方法必须在 Commands/Events 中显式注册）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/generate
   const FileName = "alonabot_gen.go" // 生成到构建入口包目录
   type Options struct { Dir string; GOOS, GOARCH string; Tags []string } // Dir 为构建入口包目录；GOOS/GOARCH 为空时取当前环境
   type Diagnostic struct { Pos string; Message string } // Pos 为 "文件:行:列"，相对 Dir
   type Command struct { Path []string; Aliases [][]string; Priority int32; Package, Func string }
   type Event struct { RouteID, EventType string; Priority int32; Package, Func string }
   type Result struct {
       RootType    string // 「导入路径.类型名」
       ConfigType  string // 没有 Config 时为空
       Commands    []Command // 按规范路径排序
       Events      []Event   // 按 route_id 排序
       Source      []byte    // 生成文件的完整内容（已 gofmt）；有诊断时为 nil
       Diagnostics []Diagnostic
   }
   func Scan(ctx context.Context, opt Options) (*Result, error) // error 只表示工具链或 IO 失败；源码问题放进 Diagnostics
2. sdk
   type Generated struct { Commands func(*CommandRegistry); Events func(*EventRegistry) }
   func UseGenerated(root Root, g Generated) // 只由生成代码在 init 中调用，root 为根类型的 nil 指针（例如 (*RSS)(nil)）；按根类型登记，同一根类型第二次调用 panic。库模式下一个进程可以有多个组件，各自登记
   Collect(root)：按 reflect.TypeOf(root) 找到登记的 Generated，在调用 root.Commands、root.Events 之前先调用它；生成项与显式注册项的重复按现有规则报错。
3. cmd/alonabot（cobra 子命令，在 goal-76 的命令树上追加）
   alonabot plugin generate [--dir DIR] [--goos OS] [--goarch ARCH] [--tags a,b]：写入或更新 DIR/alonabot_gen.go；内容不变时不改文件；有诊断时不写文件。
   alonabot plugin check（参数同上）：只扫描与比较，生成内容与现有文件不同时报 "alonabot_gen.go is out of date; run alonabot plugin generate"。
   两者把诊断逐行写到 stderr（"pos: message"），有诊断或文件过期时以 exitFailure 退出，否则 0。

行为：
- 扫描范围：在 Dir 下执行 go list -deps -json（设置 GOOS、GOARCH 环境与 -tags），只保留 Module.Main 为 true 的包（排除外部依赖与 vendor），只读取每个包的 GoFiles（已按构建约束筛选，不含测试文件），用 go/parser 解析注释。不扫描 module 中构建入口不依赖的包，所以同一仓库的多个示例互不干扰。go 命令不可用或失败时 Scan 返回 error，并带上 go list 的 stderr。
- 根类型：扫描到的包中嵌入 sdk.Base 的结构体（按导入路径解析选择器，不按名字猜）恰好一个，否则诊断（列出全部位置）。Config：嵌入 config.Base 的结构体至多一个，否则诊断。
- 标记检查（每条问题一条诊断，继续检查其余文件，最后一并报告）：语法错误、未知选项、priority 越界、event_type 不在已知列表、route_id 为空或以 "cmd." 开头、标记在非包级函数上、函数不在构建入口包时未导出、同一 Result 中规范路径或别名重复、route_id 重复。
- 生成文件：首行 "// Code generated by alonabot plugin generate. DO NOT EDIT."，package 与构建入口包相同；按导入路径排序导入扫描到的其他包（别名为 "alonabotgen" 加序号，序号按导入路径顺序分配）；在 init 中调用 sdk.UseGenerated((*<根类型>)(nil), ...)（根类型必须在构建入口包中，否则诊断），Commands 中按 Result.Commands 的顺序用 r.Command(...) 注册（带 sdk.Alias、sdk.Priority，路径按 token 用 r.Group 嵌套或直接给出完整路径，执行者选一种并在包文档写明），Events 中按 Result.Events 顺序用 r.On(...) 注册。签名问题由编译器和 Collect 现有的 checkHandler 报告。
- 稳定性：同样的输入多次生成字节完全相同（不含时间、绝对路径或 map 遍历顺序）。

测试（按编码规范「测试提交规则」）：
- 提交：sdk/generate/generate_test.go。测试在 t.TempDir() 中写出一个临时插件 module（go.mod 用 replace 指向本仓库根目录，包含 main 包、一个本 module 子包、一个只在 linux 构建的文件与一个 _test.go 文件），然后：
  - Scan 得到预期的 RootType、ConfigType、Commands、Events；子包中的导出函数被引用；_test.go 与不满足 GOOS 的文件中的标记不出现。
  - 把 Source 写为 alonabot_gen.go 后，go build 成功；运行该程序的一个测试入口（由临时 main 在设置了测试环境变量时调用 sdk.Collect 并打印注册项后退出）得到同样的命令与事件集合。
  - 第二次 Scan 的 Source 与第一次逐字节相同。
  - 两个根类型、两个 Config、方法上的标记、子包中未导出的函数、重复路径各产生一条带位置的诊断，Source 为 nil。
  - 构建 cmd/alonabot 并运行 plugin generate、plugin check：check 在生成后通过；手工改动生成文件后 check 以 exitFailure 退出并提示过期。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：标记语法的各种错误；alias 多 token；priority 边界；导入别名的确定性；同一根类型第二次 sdk.UseGenerated 时 panic、两个根类型各自登记互不影响；Collect 合并生成项后与显式注册重复时报错。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./sdk/... ./cmd/...；go build ./...。

不做：配置字段描述与 Schema（T-08.1）；i18 检查（T-08.8）；plugin init/build/test/pack 与正式入口生成（T-09.3）；命令派发（T-04.3）。
```

### goal-12 · T-04.3 命令选择、文本匹配与参数绑定

状态：等待前置（T-04.1）。语句中 goal-2、goal-4、goal-10 提供的符号（router.Router、router.CommandIndex、router.Invocation、message.Service、adapter.Registry、sdk.RouteOption 等）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-12（切片 T-04.3：命令选择、文本匹配与参数绑定）。goal 目录 goal-12，分支 goal-12，worktree .worktrees/goal-12，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 对 message.created 依次执行「普通命令（最长完整路径与别名）→ 文本匹配（Regex、keyword、startswith、endswith、fullmatch）→ Raw」的选择，已识别的命令消费本层；按契约分词并转换参数，参数错误时发一次 usage 或按声明保持沉默；SDK 把参数绑定到 Handler，并让自动回复型 Handler 的非空 Chain 由 Core 发出一次。

依据：PLAN T-04 步骤 4；CONTRACTS §2.1、§2.2 全部、§8（自动回复段）；MESSAGE-MODEL §6.1（@username）；WIRE-FORMAT §3.5（CommandMatch 与 CommandSpec 的追加）；TEST-MATRIX V-11 的命令分项、V-49。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/route.proto：CommandMatch 追加 COMMAND_MATCH_KEYWORD = 4、STARTSWITH = 5、ENDSWITH = 6、FULLMATCH = 7；CommandSpec 追加 repeated string patterns = 10、bool ignore_case = 11。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加上述编号检查
- contracts/validate/register.go：文本匹配与参数描述的注册校验
- internal/core/adapter/：新增 naming.go（可选接口）；internal/core/adapter/remote 的实例实现它
- internal/core/router/：新增 command.go（命令派发）与 args.go（分词与类型转换）；router.go 中 message.created 改走命令层（只加调用点）
- sdk/：registry.go（文本匹配、Raw 与参数选项）、dispatch.go（命令路由的分发与参数绑定）、session.go（buildRegister 输出新字段）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/adapter
   type Named interface { AccountName() string } // 平台账号的用户名，用于 /cmd@name；Instance 可选实现
2. sdk（选项沿用 goal-10 的 RouteOption）
   func Args(names ...string) RouteOption           // 依次命名参数，缺省为 arg1、arg2……
   func Optional(name string, def ...string) RouteOption // 该参数可省略；给出 def 时缺失位置用它（按参数类型解析），否则为类型零值
   func RawTail() RouteOption                        // 最后一个参数（必须是 string）接收原始尾部
   func Usage(text string) RouteOption               // 自定义 usage
   func Silent() RouteOption                         // 参数错误时不回复
   func IgnoreCase() RouteOption                     // 文本匹配不区分大小写
   func (r *CommandRegistry) Regex(routeID, expr string, handler any, opts ...RouteOption)
   func (r *CommandRegistry) Keyword(routeID string, patterns []string, handler any, opts ...RouteOption)
   func (r *CommandRegistry) StartsWith(routeID string, patterns []string, handler any, opts ...RouteOption)
   func (r *CommandRegistry) EndsWith(routeID string, patterns []string, handler any, opts ...RouteOption)
   func (r *CommandRegistry) FullMatch(routeID string, patterns []string, handler any, opts ...RouteOption)
   func (r *CommandRegistry) Raw(routeID string, handler any, opts ...RouteOption)
   func (c *Context) Match() []string // Regex 的捕获组（含整体匹配在 [0]）；其他路由为 nil
   文本匹配与 Raw 的 handler 只能是 func(*Context) error 或 func(*Context) (message.Chain, error)。

行为：
- 注册校验（Collect 与 validate.Register 两处）：Regex 编译失败、patterns 为空或含空串、Raw 带 patterns、必需参数出现在可选参数之后、RawTail 用于非 string 或非最后一个参数、Optional 指向不存在的参数名、default 不能按参数类型解析，都拒绝注册（INVALID_ARGUMENT）。Core 在 REGISTER 时编译 Regex 并缓存在 Session 的路由上，不在派发时编译。
- 选择（只对 message.created）：
  1) 文本取 chain 中各 text 片段按顺序直接连接的结果。以 "/" 开头时，按空白切出 token；第一个 token 形如 name@user 时，只有 user 与目标实例的 Named.AccountName() 不区分大小写相等才去掉 @user，否则整条消息不算命令（实例不实现 Named 时不接受 @ 形式）。
  2) 用 CommandIndex.Lookup 匹配最长完整路径或别名。命中时只调用该胜出路由：本层结束，不再进入文本匹配与 Raw（即使参数错误或 Handler 返回错误）。
  3) 没有命中任何命令时，依次处理文本匹配路由与 Raw 路由：各自按 (priority, plugin_id, route_id) 排序（文本匹配不按方式分组），逐个调用；HandlerOutcome 为 STOP 时停止后续候选。keyword/startswith/endswith/fullmatch 在 ignore_case 时按 Unicode 简单大小写折叠比较（strings.EqualFold 的规则）；Regex 用 FindStringSubmatch，捕获组放进 CommandInvocation.regex_groups。
  4) 命令层结束后，若没有 STOP，再按 goal-2 的规则派发 message.created 的 EVENT 路由。这一顺序是对 CONTRACTS §2.1「命令/类型匹配」的细化，审查者已在 DECISIONS 记录（D-44）。
- 参数（Core 侧，按 CONTRACTS §2.2 第 4、5 条）：空白分隔；单引号内原样；双引号内与未加引号时反斜杠只转义下一个字符；未闭合引号或末尾反斜杠为错误。string 原样；int 为有符号十进制 int64；float 为有限 float64（拒绝 NaN、Inf）；bool 只接受 true/false。必需参数缺失、多余参数、溢出或格式错误都是参数错误。raw 尾部保留最后一个已消费参数之后的原文并去掉开头的分隔空白；raw 是唯一参数时从命令路径后的第一个非空白字符开始。结果放进 CommandInvocation.args（按参数顺序）与 raw_tail。
- 参数错误：USAGE_MODE_SILENT 时不回复；否则经 message.Service 向原事件 Target 发送一次 usage（custom_usage 非空时用它，否则由路径、别名与参数生成），不调用插件。usage 文本暂为英文，用一行 shortcut 注释写明：T-08.8 的 Core 文案目录接入后替换。
- SDK 命令分发（替换 goal-2 中命令路由返回 UNSUPPORTED 的分支）：按 route_id 找到命令或文本匹配路由，把 args 按名字与类型绑定到 Handler 参数（类型与描述不符返回 INVALID_ARGUMENT），设置 Context.Match，调用并按 goal-2 的规则返回 HandlerOutcome。自动回复型 Handler 成功且 Chain 非空时放进 HandlerOutcome.chain；在其内调用 Context.Send 返回 FAILED_PRECONDITION。
- Core 收到 HandlerOutcome.chain 非空时经 message.Service 向原事件 Target 发送一次（走与 core.send 相同的校验），发送结果只记日志，不重试。
- 权限、禁用与 Rule 不在本切片：所有已识别命令都视为允许（T-04.4 接入策略后，权限拒绝同样消费命令且不回复）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/command_test.go（//go:build unix）。沿用 goal-2 的 echo 装配（memory Adapter 实现 Named），构建一个测试插件（testdata 下，注册 /rss、/rss add <url>、/rss list、别名 /feed、/calc <a:int> <b:int>、带 Silent 的 /quiet <n:int>、RawTail 的 /say、一个 Regex、一个 keyword 和一个 Raw）：
  - /rss add x 只调用 add；/rssx 不是命令，落到 Raw；/rss@bot 在用户名匹配时（大小写不同也算）是命令，@other 时整条不算命令。
  - /calc 1 2 回复 3（自动回复，由 Core 发一次）；/calc 1 x 回复 usage 一次且插件未被调用；/quiet x 不回复。
  - /say  "a b"  c 的 raw 尾部为 "\"a b\"  c"；引号与反斜杠规则（含未闭合引号报 usage）。
  - 命中命令时文本匹配与 Raw 都不被调用；未命中时 Regex 与 keyword 按 priority 依次调用，STOP 后 Raw 不再调用；Regex 的捕获组经 Context.Match 取到。
  - ignore_case 的 keyword 在大小写不同时命中。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：分词与类型转换的完整表（int64 溢出、NaN、空字符串参数、raw 为唯一参数）；注册校验的每种拒绝；SDK 参数绑定的类型不符；Unicode 大小写折叠。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：命令策略、权限与禁用（T-04.4）；Rule 与 ToMe（T-05.2）；Hook（T-05.3）；延续调用（T-05.4）；usage 的多语言文案（T-08.8）；生成器对新选项的标记语法（T-04.2 之后按需追加）。
```

### goal-13 · T-04.4 持久化命令策略

状态：等待前置（T-03.2、T-04.1）。语句中 goal-3、goal-10、goal-12 提供的符号（policy.Policy、policy.New、router 的命令派发、CommandIndex）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-13（切片 T-04.4：持久化命令策略）。goal 目录 goal-13，分支 goal-13，worktree .worktrees/goal-13，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：以 plugin_id 加作用域（插件、命令组、路由）持久化权限与启用覆盖；派发命令前按「最近显式权限、任一层禁用即拒绝、再与 Core 全局禁止合并」判定；Core 依据 Core Admin 列表给 Actor 赋予 USER/ADMIN 角色；提供管理服务层的查询与修改方法，持久化成功后才返回成功，运行时修改在下一次派发时生效。

依据：PLAN T-04 步骤 5；CONTRACTS §2.2 第 2、3、8 条、§2.3 第 1、2 条、§7（Core Admin 列表）；MESSAGE-MODEL §1；DECISIONS D-45；TEST-MATRIX V-12、V-43 的服务层分项。

范围（只新增或修改这些）：
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/policy.go
- internal/core/policy/：新增 command.go、admins.go、denies.go；policy.go 的 CheckSend 增加全局禁止检查
- internal/core/router/：命令派发在调用插件前检查策略；事件进入派发前由策略补全 Actor 的 trusted_roles（只加调用点）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：
   CREATE TABLE command_policies (plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), scope TEXT NOT NULL, permission INTEGER NOT NULL DEFAULT 0, enabled INTEGER, PRIMARY KEY (plugin_id, scope)); -- permission 0 表示不覆盖；enabled 为 NULL 表示不覆盖
   CREATE TABLE core_admins (adapter_instance_id TEXT NOT NULL, external_user_id TEXT NOT NULL, PRIMARY KEY (adapter_instance_id, external_user_id));
   CREATE TABLE global_denies (kind TEXT NOT NULL, adapter_instance_id TEXT NOT NULL, subject_id TEXT NOT NULL, PRIMARY KEY (kind, adapter_instance_id, subject_id)); -- kind 为 user 或 chat
   读写方法自定（放在 policy.go），全部按主键幂等。
2. internal/core/policy
   type ScopeKind int // ScopePlugin、ScopeGroup、ScopeRoute
   type Scope struct { Kind ScopeKind; Group []string; RouteID string } // Group 用于 ScopeGroup（命令路径前缀 token）；RouteID 用于 ScopeRoute
   type Override struct { Permission pb.Permission; Enabled *bool } // Permission 为 UNSPECIFIED、Enabled 为 nil 表示该项不覆盖
   type ScopedOverride struct { Scope Scope; Override Override }
   func (p *Policy) SetOverride(ctx context.Context, pluginID string, s Scope, o Override) error // 两项都不覆盖时等同删除
   func (p *Policy) Overrides(ctx context.Context, pluginID string) ([]ScopedOverride, error) // 按插件、组（路径字典序）、路由排序
   func (p *Policy) CheckCommand(ctx context.Context, pluginID string, route *pb.RouteDescriptor, actor *pb.Actor, target *pb.Target) *pb.ErrorInfo // nil 表示放行
   func (p *Policy) TrustedRoles(ctx context.Context, actor *pb.Actor) []pb.TrustedRole
   func (p *Policy) AddAdmin(ctx context.Context, instanceID, userID string) error
   func (p *Policy) RemoveAdmin(ctx context.Context, instanceID, userID string) error
   func (p *Policy) Admins(ctx context.Context) ([]Admin, error) // type Admin struct { InstanceID, UserID string }
   type DenyKind string // DenyUser = "user"、DenyChat = "chat"
   func (p *Policy) Deny(ctx context.Context, kind DenyKind, instanceID, subjectID string) error
   func (p *Policy) Undeny(ctx context.Context, kind DenyKind, instanceID, subjectID string) error
   func (p *Policy) Denies(ctx context.Context) ([]Deny, error) // type Deny struct { Kind DenyKind; InstanceID, SubjectID string }

行为：
- TrustedRoles：actor 为 nil 返回空。否则含 USER；(actor.adapter_instance_id, external_user_id) 在 Core Admin 列表中时再含 ADMIN；Adapter 给出的 CHAT_ADMIN 原样保留。Router 在每个事件进入派发前调用它重写 actor.trusted_roles，插件与 Adapter 自报的 USER/ADMIN 一律不采信。
- CheckCommand 的顺序（任一步拒绝即返回）：
  1) 全局禁止：actor 存在且 (实例, external_user_id) 在 user 禁止表中，或 target 存在且 (实例, chat_id) 在 chat 禁止表中 → PERMISSION_DENIED。
  2) 启用：插件作用域、该路由命令路径的每一层祖先组、路由自身中，任一层显式 Enabled=false → 拒绝（FAILED_PRECONDITION，safe_message "command is disabled"）。路由自身没有显式 Enabled 时取 route.default_enabled。子项的 Enabled=true 不能抵消祖先的 false。
  3) 权限：取最近的显式 Permission（路由 → 最长的祖先组 → 插件），都没有时取 route.required_permission。ANY 放行；USER 需要 TrustedRoles 含 USER；ADMIN 需要含 ADMIN；否则 PERMISSION_DENIED。没有 Actor 的调用不满足 USER 与 ADMIN。
  文本匹配与 Raw 路由的祖先组为空，只看插件与路由自身。
- 每次 CheckCommand 都从状态库读取当前策略，不缓存；同一次检查内的多次读取在一个只读事务中完成，保证看到一致的快照。修改在下一次派发时生效，已开始的调用不受影响。
- Router（goal-12 的命令层）：已识别的命令在调用插件前执行 CheckCommand；拒绝时消费该命令（不进入文本匹配与 Raw），不向会话回复，以 INFO 记录（plugin_id、route_id、错误码，不含消息内容）。文本匹配与 Raw 路由被拒时跳过该候选，继续后面的候选。
- CheckSend（goal-3）在许可检查之前先检查 chat 禁止：目标 (实例, chat_id) 被禁止 → PERMISSION_DENIED。
- 管理服务层：SetOverride、AddAdmin、Deny 等写操作在事务提交成功后才返回 nil；参数非法（未知 ScopeKind、ScopeGroup 的 Group 为空、ScopeRoute 的 RouteID 为空、Permission 不是已知枚举值、kind 未知）返回 INVALID_ARGUMENT。遮蔽诊断直接使用 goal-10 的 Router.Commands().Shadows()，本切片不重复实现。
- 卸载插件时策略行的清理由 T-09.4 负责（command_policies 外键引用 installs）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/policy_test.go（//go:build unix）。沿用 goal-12 的命令测试装配与测试插件，接入真实状态库上的 policy.Policy：
  - /rss list 的 required_permission 为 ANY 时任何人可用；对插件作用域设 ADMIN 后普通用户被拒（不回复、插件未被调用、文本匹配也未被调用），加入 Core Admin 后放行。
  - 组 rss 设 ADMIN、路由 rss list 设 ANY：list 对普通用户放行（最近显式覆盖），rss add 仍要求 ADMIN。
  - 组 rss 禁用、路由 rss list 显式启用：list 仍被拒（父禁用优先）；取消组禁用后放行，不重启插件。
  - 把用户加入 user 禁止表：即使是 Core Admin 也被拒；把会话加入 chat 禁止表：sdk.Send 发往该会话返回 PERMISSION_DENIED。
  - Adapter 自报的 ADMIN 角色被丢弃，插件收到的 Actor 只有 Core 判定的角色。
  - 关闭并重新打开状态库后策略仍在。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：CheckCommand 的完整判定表（含 default_enabled=false、无 Actor、多层组）；SetOverride 两项都不覆盖时删除行；非法参数；并发修改与检查在 -race 下无数据竞争。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：管理 API、CLI 与 WebUI（T-09.4、T-09.9）；/whoami（T-14）；Rule 与 chat_admin Rule（T-05.2）；Hook（T-05.3）。
```

### goal-19 · T-06.1 Service/Bus 描述与 Broker 准入

状态：等待前置（T-04.1）。语句中 goal-2、goal-10 提供的符号（router.Router、router.Mux、router.Plugin、router.Caller、sdk 的分发与 currentSession、RouteOption）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-19（切片 T-06.1：Service/Bus 描述与 Broker 准入）。goal 目录 goal-19，分支 goal-19，worktree .worktrees/goal-19，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件可以声明 Service（名称、输入/输出 schema 与版本、编码、访问规则）并经 SDK 调用其他插件的 Service；Core 的 Broker 以 plugin_id/service_name 定位目标的已激活 Session，在调用目标函数前完成访问规则、编码、schema 标识与版本、JSON 语法的检查，再以新的 Core→目标请求转发。REGISTER 同时校验 Service 与 Bus 订阅描述（Bus 的运行在 T-06.3）。

依据：PLAN T-06 步骤 1、2；CONTRACTS §3 第 2、3 条及第 4 条中「每一跳单独编号」的部分、§1.2；WIRE-FORMAT §4、§5；DECISIONS D-43；TEST-MATRIX V-19。

范围（只新增或修改这些）：
- contracts/validate/register.go：Service 与 Bus 订阅描述的注册校验
- 新增 internal/core/broker/
- internal/core/router/：Mux 增加对含 "/" 目标的转交入口；Router 增加按 plugin_id 查找已 Attach 插件的方法
- sdk/：新增 services.go（Service 注册与调用）；base.go 的 Root 增加 Services Hook 与 Base 默认实现；sdk.go 的 Collect 与 dispatch.go 的分发接入服务路由（只加调用点）；session.go 的 buildRegister 输出服务路由
- 新增 internal/core/supervisor/testdata/svcplugin/（集成测试用的插件程序）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk
   Root 增加 Hook：Services(*ServiceRegistry) // Collect 时调用一次；func (*Base) Services(*ServiceRegistry) 默认不注册
   type ServiceRegistry struct{ ... }
   type Access int // Private（缺省，只允许 Core 与本插件）、Allowlist、Public
   type ServiceOption func(*serviceOptions)
   func WithAccess(a Access, allowlist ...string) ServiceOption // Allowlist 时给出允许的插件 ID
   // JSON 服务：In/Out 为可 json 编解码的类型；schemaIn/schemaOut 为随包固定的 schema 标识，版本由 version 给出
   func JSONService[In, Out any](r *ServiceRegistry, name, schemaIn, schemaOut string, version uint32, h func(*Context, In) (Out, error), opts ...ServiceOption)
   // Protobuf 服务：schema 标识取 validate.SchemaID，版本 1
   func ProtoService[In, Out proto.Message](r *ServiceRegistry, name string, h func(*Context, In) (Out, error), opts ...ServiceOption)
   func CallJSON[Out any](ctx context.Context, target, schemaIn string, version uint32, in any) (Out, error) // target 为 "plugin_id/service_name"
   func CallProto[Out proto.Message](ctx context.Context, target string, in proto.Message) (Out, error)
   func (c *Context) Caller() string // 服务调用的直接调用方插件 ID；Core 发起时为空
   在 Handler 内调用时，Call* 以本次调用为父请求（parent_call_id）；在 Handler 之外调用时没有父请求，只带插件身份。
2. internal/core/router
   func (m *Mux) Forward(fn ServiceFunc) // Target 含 "/" 且没有同名固定服务时交给 fn；未设置时仍为 NOT_FOUND
   func (r *Router) Plugin(pluginID string) (Plugin, bool)
3. internal/core/broker
   type Broker struct{ ... }
   func New(r *router.Router, log *slog.Logger, requestTimeout time.Duration) *Broker // requestTimeout 取 tune.Values.RequestTimeout
   func (b *Broker) Call(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response // 注册为 Mux.Forward

行为：
- 描述与注册校验（Collect 与 validate.Register 两处，任一失败整体拒绝）：
  - SERVICE 路由：route_id 为 "svc." + service_name；service_name 为小写字母开头、只含小写字母、数字、下划线与点，长度 1～64，且同插件内唯一；encoding 为 JSON 或 PROTOBUF；input/output schema_id 与版本非空且版本 ≥ 1；ALLOWLIST 时 allowlist 非空且每项是合法插件 ID（manifest.ValidateID），PRIVATE/PUBLIC 时 allowlist 为空。
  - BUS_SUBSCRIPTION 路由：event_name 形如 "plugin_id/event_name"，schema_id 与版本非空，callback_route_id 指向本次注册中的一条路由。
- Broker.Call 的顺序（任一步失败都不调用目标函数）：
  1) target 拆成 plugin_id 与 service_name，格式不对 → INVALID_ARGUMENT。
  2) Router.Plugin 找不到或未 Activated → UNAVAILABLE；目标插件没有该服务 → NOT_FOUND。
  3) 访问规则：调用方等于目标插件，或 PUBLIC，或 ALLOWLIST 且调用方在列表中，才放行；否则 PERMISSION_DENIED。PUBLIC 只表示已启用插件可以调用，不赋予任何用户角色。
  4) 请求的 payload_encoding 必须等于服务的 encoding，schema_id 与 schema_version 必须等于服务的 input schema，否则 INVALID_ARGUMENT；JSON 编码时载荷必须是合法的 UTF-8 JSON（json.Valid），否则 INVALID_ARGUMENT。Core 不校验业务字段。
  5) 以新的 call_id 向目标 Session 发出 Request{Target: route_id, Payload: ServiceInvocation{route_id, context, encoding, schema_id, schema_version, payload}}。context 由 Core 构造：plugin_chain 为 [调用方]；有父请求时沿用父调用的 Actor 与默认 Target，没有父请求时两者为空。期限取父期限、请求自带期限与 requestTimeout 中最早者。Broker 记录「调用方请求 ↔ 转发请求」的对应关系，不复用调用方的 call_id。
  6) 目标的响应：有 Err 时原样返回给调用方（Code 经 validate.PublicCode）；成功时载荷的编码与 schema 必须等于服务的 output schema，否则以 INTERNAL 返回给调用方并以 WARN 记录（目标插件、服务名）。空结果也必须是声明的空对象或空消息。
- SDK 服务分发：按 route_id 找到服务，按 encoding 把 payload 解码为 In（JSON 用 encoding/json 解到 In，未知字段保留由 In 的类型决定；Protobuf 用 proto.Unmarshal），解码失败返回 INVALID_ARGUMENT 且不调用 h；h 返回的 Out 按同一编码编码，schema 为声明的 output schema。h 返回 *sdk.Error 时用其 Code，其他错误为 INTERNAL（safe_message "service returned an error"，原文以 ERROR 记日志）。panic 转 INTERNAL 并记录堆栈。
- CallJSON/CallProto：未就绪时返回 FAILED_PRECONDITION；把 Core 的错误转为 *sdk.Error；成功时按声明解码 Out，解码失败返回 INTERNAL。
- 父调用链深度、CALL_CYCLE、取消传播与背压属于 T-06.2，本切片不实现。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/service_test.go（//go:build unix）。构建 testdata/svcplugin，按环境变量以三个插件 ID 启动 a、b、c：a 提供 JSON 服务 a/echo（PRIVATE）、a/pub（PUBLIC）、a/allow（ALLOWLIST: b），以及一个 Protobuf 服务；b、c 各有一条事件路由，收到消息时按文本调用指定服务并把结果或错误码回复出来。用 Router、Mux（Forward 为 Broker）与 memory Adapter 装配：
  - b 调用 a/pub 与 a/allow 成功，结果正确；c 调用 a/allow 与 a/echo 返回 PERMISSION_DENIED，且 a 的函数没有被调用（a 用调用计数回复验证）。
  - 声明 JSON 服务时用错误的 schema 版本、错误的编码、非法 JSON 调用都返回 INVALID_ARGUMENT，a 未被调用；调用不存在的服务为 NOT_FOUND，调用未启动的插件为 UNAVAILABLE。
  - a 返回与 output schema 不符的结果时，调用方收到 INTERNAL。
  - Protobuf 服务的往返正确。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：validate 的 Service 与 Bus 订阅描述各条拒绝；Collect 的服务名重复与非法；期限取最早者；Context.Caller；SDK 解码失败不调用 h；Mux.Forward 不覆盖固定服务。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：父调用链、深度与 CALL_CYCLE、取消与背压（T-06.2）；Event Bus 的发布与投递（T-06.3）；工具描述（T-11.1）；用 jsonschema-go 校验业务字段（Core 不校验，SDK 以类型解码为准；工具的 Schema 校验在 T-11）。
```

### goal-25 · T-07.3 SDK HTTP Listener 与 Core 代理

状态：等待前置（T-04.1）。语句中 goal-2、goal-10、goal-76 提供的符号（router.Router、router.Plugin、sdk 的 HTTPRegistry 与激活流程、cobra 的 run 命令、internal/tune）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-25（切片 T-07.3：SDK HTTP Listener 与 Core 代理）。goal 目录 goal-25，分支 goal-25，worktree .worktrees/goal-25，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件只注册 http.Handler：SDK 在存在 HTTP 路由时持有 127.0.0.1:0 的 Listener，并在 REGISTER 报告实际端点，用户 Handler 外包内部凭证校验与就绪门禁；Core 读取 <data_dir>/alonabot.yaml（不存在时以缺省值创建），在唯一的公开 Listener 上把 /plugins/{id}/... 经标准库 ReverseProxy.Rewrite 转发给当前已激活 Session 的端点，清除可伪造的身份信息并注入受信凭证；Session 结束或代次变化时撤销映射。

依据：PLAN T-07 步骤 1（HTTP）、4、5；CONTRACTS §7.2、§15.1、§14.6；WIRE-FORMAT §3.2（http_proxy_auth）、§3.3；DECISIONS D-25；TEST-MATRIX V-22 的组件分项。

范围（只新增或修改这些）：
- contracts/validate/register.go：HTTP 路由描述的注册校验
- internal/wire/：新增 http.go（内部认证头名常量）
- internal/tune/：新增导出的大小解析函数（供设置文件复用）
- 新增 internal/core/settings/（alonabot.yaml）
- 新增 internal/core/httpproxy/
- internal/core/host/prepare.go：插件声明 HTTP 路由时在 Bootstrap 中下发 http_proxy_auth
- sdk/：新增 http.go（Listener、门禁与凭证校验）；sdk.go 去掉「不支持 HTTP 路由」的拒绝（只改该处）；activate.go 与 session.go 只加启动、上报与关闭 Listener 的调用点
- cmd/alonabot/：run 读取设置文件、增加 --listen 并让 --log-level/--log-format 覆盖设置文件；启动 Core HTTP Listener 并接入插件的 HTTP 路由
- 新增 internal/core/supervisor/testdata/httpplugin/（集成测试用的插件程序）
- 上述包的测试文件
允许的新跨模块边：无（cmd/alonabot -> internal/core/...、-> internal/tune 已允许；yaml 使用已批准的 go.yaml.in/yaml/v3）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/wire/http.go
   const HeaderProxyAuth = "X-Alonabot-Proxy-Auth" // Core 注入、SDK 校验的内部凭证头；值为 http_proxy_auth.token 的十六进制
2. internal/tune
   func ParseSize(s string) (int64, error) // 与 --args 中大小值相同的语法与错误提示（B、KiB、MiB、GiB；MB 等提示改用二进制单位）
3. internal/core/settings
   type Settings struct {
       Listen         string   // listen
       AdminBaseURL   string   // admin_base_url
       PluginBaseURL  string   // plugin_base_url
       TrustedProxies []netip.Prefix
       LogLevel       slog.Level
       LogFormat      string
       LogFile        struct { Enabled bool; MaxSize, MaxTotal int64; MaxAgeDays int }
       Proxy          string
   }
   const FileName = "alonabot.yaml"
   func Defaults() Settings // 与 CONTRACTS §15.1 的表一致
   func Load(dataDir string) (Settings, error) // 文件不存在时写入缺省内容（0600）并返回缺省值；未知键、类型不符、URL 或地址非法时返回带键名与行号的错误
4. internal/core/httpproxy
   type Authorizer interface { Authorize(r *http.Request, pluginID string) bool } // 私有路由的 Core 授权；T-07.4 提供插件会话实现
   type Server struct{ ... }
   func New(log *slog.Logger, auth Authorizer) *Server // auth 为 nil 时私有路由一律拒绝
   func (s *Server) Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor) // 绑定当前已激活 Session 的路由；替换同一插件的旧映射
   func (s *Server) Unbind(pluginID, generationID string) // 只有 generationID 与当前映射一致时才撤销
   func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)
5. sdk（HTTPRegistry 已存在）：Handle 的注册项随 REGISTER 上报；不新增公开符号。

行为：
- 注册校验（validate.Register 与 SDK Collect）：HTTP 路由的 method 为大写的标准方法之一；path 以 "/" 开头、是 path.Clean 后的规范形式、不含 "?"、"#" 与 ".."；同插件的 (method, path) 唯一；is_public 为 true 时 auth_mode 必须是 CALLBACK_STATE 或 CALLBACK_SIGNATURE，为 false 时必须是 CORE_SESSION；有 HTTP 路由时 http_listener 必须存在且为 loopback 非零端口，没有 HTTP 路由时必须不存在。
- SDK：
  - 只有存在 HTTP 路由时才在准备阶段监听 127.0.0.1:0，REGISTER 报告实际 host 与 port；Listener 在整个 Session 期间保持打开，不释放后重开。
  - 每个请求先校验 HeaderProxyAuth 等于 Bootstrap 下发 token 的十六进制（恒定时间比较），不符返回 401 且不进入用户 Handler；Session 尚未 ACTIVATED 时返回 503；校验通过后删除该头再交给用户 Handler。路由匹配按注册的 method 与 path（精确匹配 path，或注册为以 "/" 结尾的前缀）。
  - STOPPING、断链、激活失败与 Run 返回时关闭 Listener 并等待在途请求结束（受停止预算约束）。
- Core host：插件声明了 HTTP 路由时，Bootstrap.http_proxy_auth.token 为本次 Session 新生成的 32 字节随机值；不进入日志。
- Core httpproxy：
  - 只处理 /plugins/{id}/...；其余路径返回 404（Admin、API、healthz 归 T-07.4 与 T-09）。
  - 找到当前映射后按 method 与去掉 /plugins/{id} 后的路径匹配路由；没有映射（未激活、已停用或代次变化）返回 503，没有匹配路由返回 404，method 不符返回 405。
  - 私有路由调用 Authorizer，不通过返回 401；公开路由直接转发，插件自行验证 state、签名或 nonce。
  - 用 httputil.ReverseProxy 的 Rewrite：目标固定为映射中的 loopback 端点；删除入站的 HeaderProxyAuth、Cookie 中的管理会话 Cookie（名称由 T-07.4 确定前，删除全部以 "alonabot_" 开头的 Cookie）、Authorization、Forwarded 与 X-Forwarded-*，再注入 HeaderProxyAuth；转发路径为去掉前缀后的路径（以 "/" 开头），保留查询串。不使用 Director，不接受请求指定的目标。
  - 插件端点不可达时返回 502，不重试。
- 设置文件：缺省值与 §15.1 的表一致；键名与表相同（log.file.max_size 等用嵌套 YAML）。listen 必须是 host:port；两个 Base URL 必须是 http 或 https 的绝对 URL、端口相同、主机名不同；log.file.max_size、max_total 用 tune.ParseSize。proxy 非空时必须是 http、https 或 socks5 URL；按 §15.1 为 Core 与子进程设置 HTTP_PROXY/HTTPS_PROXY（为空时原样传递环境中的三个变量），子进程环境的设置经 process 的现有白名单机制传入。
- cmd/alonabot run：取得 core.lock 后读取设置文件；--listen、--log-level、--log-format 只在命令行显式给出时覆盖设置，只对本次运行有效；日志文件按设置启用。启动 Core HTTP Listener（listen 地址），插件激活后 Bind、停止或重启前 Unbind。Authorizer 暂为 nil（私有路由一律 401，由 T-07.4 接入插件会话），用一行 shortcut 注释写明。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/http_test.go（//go:build unix）：构建 testdata/httpplugin（注册私有 GET /hello、公开 POST /hook、前缀路由 /static/，Handler 把收到的路径、查询串、请求头写回响应），Launch 后 Bind 到 httpproxy.Server，用 httptest 驱动：
    - 公开路由转发成功，插件收到的路径以 "/" 开头且保留查询串；伪造的 HeaderProxyAuth、Authorization、Forwarded、X-Forwarded-For 与 alonabot_ 开头的 Cookie 都没有到达插件。
    - 私有路由在 Authorizer 为 nil 时返回 401；注入一个总是放行的 Authorizer 后转发成功。
    - 直接请求插件的 loopback 端点（不带或带错的 HeaderProxyAuth）返回 401，用户 Handler 未执行。
    - Stop 插件后同一路径返回 503，插件端口已关闭；重新 Launch 得到新端点与新 token，旧 generation 的 Unbind 不影响新映射。
    - 未注册的路径 404、method 不符 405。
  - cmd/alonabot/e2e_test.go 追加：首次运行在 data_dir 下生成与缺省值一致的 alonabot.yaml（权限 0600）；--listen 覆盖只对本次运行有效，文件不变；设置文件含未知键时 run 以 exitUsage 退出并指明键名与行号。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：settings 的每种非法值；tune.ParseSize；HTTP 路由描述的各条拒绝；SDK 在 ACTIVATED 前返回 503；没有 HTTP 路由时不创建 Listener 且 REGISTER 无 http_listener；恒定时间比较。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./cmd/...；go build ./...。

不做：两个 Host 的边界、插件兑换票据与插件会话（T-07.4）；管理认证与 Admin 页面（T-09）；HTTPS 与可信反代的 Host 校验（T-07.4、T-09）；alonabot serve（T-09.4）。
```

### goal-78 · T-15.1 官方 OneBot 11 Adapter 的连接与接收

状态：等待前置（T-03.3）。语句中 goal-4 提供的符号（sdk/adapter 的 Adapter、Instance、Core、StartRequest、Account，sdk.RunAdapter）只写名字；goal-4 合并后，审查者按实际签名核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-78（切片 T-15.1：官方 OneBot 11 Adapter 的连接与接收）。goal 目录 goal-78，分支 goal-78，worktree .worktrees/goal-78，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：实现官方 OneBot 11 Adapter 的连接与接收：正向 WebSocket 连接与 access_token、get_login_info 账号、按 echo 配对的 API 调用、有界退避重连、self_id 校验、数组格式检查、实例内有界事件队列与逐条交付、私聊/群聊/撤回/成员事件的标准化与 to_me；提供本地 OneBot 11 模拟服务，供本切片与后续 OneBot 11 切片测试。

依据：PLAN T-15 步骤 1、2；CONTRACTS §8（OneBot 11 段）、§14.5；MESSAGE-MODEL §1、§1.1、§7、§7.1；DECISIONS D-46；TEST-MATRIX V-71 的接收分项；OneBot 11 标准原文（https://github.com/botuniverse/onebot-11 的 communication/ws.md、event/、api/public.md、message/segment.md），字段名以原文为准。

范围（只新增或修改这些）：
- go.mod、go.sum：引入 github.com/coder/websocket（本切片获授权引入，只在 adapters/onebot11 使用，版本取当前最新的正式版）
- 新增 adapters/onebot11/（可导入的包 onebot11）
- 新增 adapters/onebot11/onebottest/（本地 OneBot 11 模拟服务）
- 新增 adapters/onebot11/cmd/alonabot-adapter-onebot11/main.go（只调用 sdk.RunAdapter(onebot11.New())）
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
  3) get_login_info：account_id 为 user_id 的十进制字符串，DisplayName 为 nickname；失败 → UNAVAILABLE。
  4) 启动读循环与交付循环后返回。Capabilities 本切片只声明接收相关的能力，发送能力由 T-15.2 补充。
- API 调用：每个请求带实例内唯一的 echo（递增的十进制字符串），响应按 echo 配对；调用以 ctx 为期限，ctx 结束时返回 ctx 的错误；连接断开时全部在途调用以 UNAVAILABLE 结束。带 echo 的帧是响应，其余帧是上报。
- 读循环一直读取，不因等待交付而停止：
  - 响应交给配对的调用；meta_event 不入队（lifecycle 的 connect 只记 DEBUG）。
  - 其余上报放进实例事件队列（256 条，命名常量）；队列满时丢弃该上报，累加丢弃计数并记 WARN（带累计数）。
- 交付循环按顺序取出上报，标准化后经 Core.Deliver 交付，游标为空；收到应答（无论 accepted）后才取下一条。不交付的上报直接跳过。
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

## 第 6 波

### goal-7 · T-03.6 入站媒体与 Grant

状态：等待前置（T-03.2、T-03.4、T-03.5）。语句中 goal-2～goal-6 提供的符号（media.Store、media.File、remote.Process、sdk/adapter 的 Core 与 Instance、router 的派发、message.Service、policy.Policy）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-7（切片 T-03.6：入站媒体与 Grant）。goal 目录 goal-7，分支 goal-7，worktree .worktrees/goal-7，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：入站媒体以 IncomingMedia 交给插件：Core 只保存可信 Adapter 给出的平台文件标识，向每个接收 Session 签发独立 token；首次 Read 时经 adapter.media.read 按需取件、校验后缓存一份供多个 Session 共享。插件经 core.media.read 读取、经 core.media.grant 在不扩大范围的前提下转给其他插件。出站媒体在 adapter.send 之前由 Core 签发只读 token，Adapter 经 core.media.read 读取；发送期间以租约保护对象。

依据：PLAN T-03 步骤 5（入站与授权）；CONTRACTS §1.3、§1.4、§8（媒体段）；WIRE-FORMAT §3.4、§5；TEST-MATRIX V-44、V-09。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/adapter.proto：追加 AdapterMediaReadRequest/Response。随后 go generate ./...（本切片获授权只追加这两个消息）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceAdapterMediaRead = "adapter.media.read"、ServiceMediaRead = "core.media.read"、ServiceMediaGrant = "core.media.grant"
- internal/core/media/：新增 incoming.go、read.go、grant.go、lease.go
- internal/core/adapter/：新增 media.go（可选接口）；internal/core/adapter/remote 与 internal/core/adapter/memory 实现它；remote 的 deliver 改写入站媒体来源
- internal/core/router/：派发前为每个接收 Session 改写 incoming token（只加调用点）
- internal/core/message/：发送前解析媒体来源并签发 Adapter 只读 token
- sdk/：media.go 增加读取与授权函数；sdk/adapter 的 Core 增加 ReadMedia，Instance 可选实现平台文件读取；sdk 的 RunAdapter 处理 adapter.media.read
- adapters/telegram/：新增 media.go（入站媒体片段与 getFile 取件）；adapters/telegram/telegramtest/ 增加 getFile 与文件下载
- 上述包的测试文件（含扩展 internal/core/adapter/remote/remote_test.go 的回环 Adapter）
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapter.proto
   message AdapterMediaReadRequest { string instance_id = 1; string file_id = 2; uint64 offset = 3; uint32 length = 4; } // length 为 1～1 MiB
   message AdapterMediaReadResponse { bytes data = 1; uint64 next_offset = 2; bool eof = 3; }
2. internal/core/adapter
   type MediaReader interface { ReadMedia(ctx context.Context, fileID string, offset uint64, length uint32) (data []byte, next uint64, eof bool, err error) } // Instance 可选实现
3. internal/core/media（在 goal-6 的 Store 上追加）
   func (s *Store) RegisterIncoming(instanceID, fileID string, size *uint64, mime *string) string // 返回对象 ID；同一实例同一 file_id 复用同一对象
   func (s *Store) IssueIncoming(objectID string, c router.Caller, targets []*pb.Target) *pb.IncomingMedia // 给一个接收 Session 签发 token
   func (s *Store) AddReadServices(m *router.Mux, readers *adapter.Registry) // 注册 core.media.read 与 core.media.grant
   func (s *Store) LeaseForSend(ctx context.Context, c router.Caller, source *pb.MediaSource, deadline time.Time) (readToken string, release func(), err error) // 供 message.Service：解析 media_ref 或 incoming_ref，签发给 Adapter Session 的只读 token
4. sdk
   func ReadMedia(ctx context.Context, token string, offset uint64, length uint32) (data []byte, next uint64, eof bool, err error)
   func GrantMedia(ctx context.Context, token, granteePluginID string, targets ...*pb.Target) (string, error) // 返回给对方的新 token
5. sdk/adapter
   Core 增加：ReadMedia(ctx context.Context, token string, offset uint64, length uint32) ([]byte, uint64, bool, error)
   type MediaSource interface { ReadPlatformFile(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) } // Instance 可选实现；未实现时 adapter.media.read 返回 UNSUPPORTED

行为：
- 入站：Adapter 在 image/video/file 片段的 MediaSource.incoming_ref.token 中放平台文件标识（可带 size、mime）。remote 的 deliver 在接纳前把每个这样的来源登记为对象（RegisterIncoming），token 改为对象 ID；平台文件标识此后只留在 Core。Router 为每个接收 Session 调用 IssueIncoming，把对象 ID 换成该 Session 独立的 token（16 字节随机值的十六进制，自签发起 MEDIA_TTL 有效，readability 为 READABLE；实例不实现 MediaReader 时为 UNSUPPORTED）。插件看不到平台文件标识。
- core.media.read（插件与 Adapter 都可调用，只认调用者 Session 持有的有效 token）：
  - length 为 0 或超过 1 MiB、offset 超过实际大小 → INVALID_ARGUMENT；offset 等于实际大小 → 空 data、原 offset、eof=true。
  - 已提交的 MediaRef（goal-6）直接按 offset 读取。
  - 入站对象首次被读取时取件：预留容量（已知 size 用它，未知按 MEDIA_MAX_OBJECT），计入 Core 总预算与读取者 Session 的预算，超限 RESOURCE_EXHAUSTED；按 1 MiB 依次调用 MediaReader 写入暂存区，总长超过预留或与声明的 size 不符时失败并删除；完成后计算 SHA256。取件期限取调用者期限与 MEDIA_TTL 中较早者。并发的首次读取只取件一次，其余等待结果。取件失败返回 UNAVAILABLE（平台错误的 safe_message 原样），不自动重试。
- core.media.grant：只能授权自己持有的 token；grantee 必须是当前已激活的插件；targets 必须是原 token 范围的子集（原范围为空时 targets 必须为空）。入站对象还要求 grantee 与该事件有在途调用关联（Router 记录的同一事件派发），或 grantee 的 allowed_adapter_instances 含来源实例；插件自己上传的对象由 owner 直接委托。新 token 不延长有效期（取原 token 的到期时间）。不满足时 PERMISSION_DENIED。
- 出站：message.Service 在调用 Instance.Send 前，对每个 media_ref 或 incoming_ref 来源调用 LeaseForSend：校验调用者对该 token 的权利，入站对象必要时先取件，再签发绑定目标 Adapter Session 的只读 token 并替换片段来源；租约在发送返回或期限到达时释放。租约期间对象不被 Sweep 回收，但仍计入磁盘预算。https_url 来源原样交给 Adapter。取件失败发生在平台发送之前，不调用 Send。
- Session 结束（ReleaseSession）只撤销该 Session 的 token；入站对象在所有 token 过期且没有租约时才删除。
- Adapter 端：RunAdapter 收到 adapter.media.read 时调用实例的 MediaSource；Adapter 用 Core.ReadMedia 读取出站只读 token 的内容。
- Telegram：photo（取最大尺寸）、video、document、animation、audio、voice 生成对应的 image/video/file 片段，incoming_ref.token 为 file_id，size 取 file_size（有时），mime 取 mime_type（有时），caption 写入片段；实例实现 MediaSource：首次读取时 getFile 得到 file_path，经实例的 http.Client（含代理）下载 /file/bot<token>/<file_path>，按 offset 返回；超过 Bot API 的 20 MiB 下载上限（命名常量）时返回 UNSUPPORTED。下载 URL 同样脱敏。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/media_in_test.go（//go:build unix）：memory Adapter 实现 MediaReader（内存中的文件），注入带图片的 message.created；两个测试插件都收到不同的 token，各自用 sdk.ReadMedia 分块读出完整内容；MediaReader 只被调用一次取件；插件 a 用 sdk.GrantMedia 授权给与事件无关联且无实例许可的 c 时被拒，给已有许可的 b 成功；a 用 incoming_ref 回发该图片，memory Adapter 的 SendFunc 收到只读 token，并能经 media.Store 读出相同字节。
  - internal/core/adapter/remote/remote_test.go 追加：回环 Adapter 实现 MediaSource；经真实 adapter.media.read 取件并在 Read 中得到相同内容；未实现 MediaSource 的实例使 readability 为 UNSUPPORTED。
  - contracts/wirespec_test.go：新消息的编号检查。
  - adapters/telegram/media_test.go：telegramtest 推入带 photo 与 document 的消息，片段的 incoming_ref.token 为 file_id；MediaSource 经 getFile 与下载返回正确内容，错误中不含 token。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：offset/length 边界；声明 size 与实际不符；未知大小超过上限；并发首次读取只取件一次；token 过期；Grant 扩大范围被拒；租约期间 Sweep 不删除；ReleaseSession 不影响其他 Session 的 token。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。

不做：https_url 的下载代理（契约不提供）；按钮与 extension（T-03.8）；本地 Bot API 服务的大文件路径（T-03.10）。
```

### goal-50 · T-03.7 官方 Telegram Adapter 的发送计划与回执

状态：等待前置（T-03.4）。语句中 goal-4、goal-5 提供的符号（adapters/telegram 的实例与 Bot API 客户端、telegramtest.Server、sdk/adapter 的 Instance 与 Core）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-50（切片 T-03.7：官方 Telegram Adapter 的发送计划与回执）。goal 目录 goal-50，分支 goal-50，worktree .worktrees/goal-50，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

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
  - 末尾为相册且带 buttons：追加 editMessageReplyMarkup；该步失败时整体 PARTIAL，媒体组的消息 ID 保留在结果中。
  - ZeroMessageID 作用于相册末条：markup 步骤为 NOT_ATTEMPTED，整体 PARTIAL，platform_message_ids 不含 0。
  - 第二步 FailMethod(400)：第一步 ACCEPTED、第二步 REJECTED、其余 NOT_ATTEMPTED，整体 PARTIAL 并带 ErrorInfo；第一步 Timeout：整体 UNKNOWN。
  - 预检失败（超长文本、forward、缺 core_token 的按钮、markdown 与 entities 同时出现）时 Calls() 为空。
  - 含 emoji 与组合字符的文本，entities 的 UTF-16 偏移在请求中正确。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：媒体组 10 项与 11 项的拆分；文本合并规则；caption 长度边界；mention 的渲染；429 的 safe_message；thread_id；multipart 上传的字段。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./adapters/... ./contracts/...；go build ./...。

不做：按钮 token 的签发与 callback_query（T-03.8）；extension（T-03.8）；编辑（T-03.9）；入站媒体与 getFile（T-03.6 之后在 T-03.10 补 MediaSource）；BeforeSend/AfterSend Hook（T-05.3）。
```

### goal-15 · T-05.1 可信上下文与有界分发

状态：等待前置（T-03.3、T-04.3、T-04.4）。语句中 goal-2、goal-4、goal-12、goal-13、goal-76 提供的符号（router.Router、router.New、remote.Process、policy.Policy、tune.Values、wire.Budget）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-15（切片 T-05.1：可信上下文与有界分发）。goal 目录 goal-15，分支 goal-15，worktree .worktrees/goal-15，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Router 把 Adapter 输入转为 Core 保留的可信信息（Actor、Target、来源消息 ID、路由绑定）与可修改的业务 payload 两部分，插件看到的 Event 由 Core 保留的部分重建；用有界的并发派发替换单 worker：最多 DISPATCH_CONCURRENCY 条链并发、等待队列受 DISPATCH_QUEUE_EVENTS 与 DISPATCH_QUEUE_BYTES 限制并计入 Core 总预算，同一事件的候选在一条链内顺序执行，链与 Session 读循环分离。

依据：PLAN T-05 步骤 1、2；CONTRACTS §2.1、§2.3 第 1 条、§1.2；MESSAGE-MODEL §1、§1.1；TEST-MATRIX V-13、V-16 的 fixture 分项。

范围（只新增或修改这些）：
- internal/core/router/：新增 dispatch.go（并发链与等待队列）、trusted.go（可信信息的保留与重建）；router.go 的 New、Deliver 与 worker 改用它们，删除 goal-2 的单 worker shortcut 注释
- 更新 router.New 的全部调用者（含测试与 cmd/alonabot 之外的装配代码；cmd/alonabot 若已装配 Router 也一并更新）
- 新增 contracts/testdata/events/：button.clicked 与 scheduled 的规范 fixture（JSON 形式的 pb.Event）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/router
   type Options struct {
       Concurrency int          // DISPATCH_CONCURRENCY
       QueueEvents int          // DISPATCH_QUEUE_EVENTS
       QueueBytes  int64        // DISPATCH_QUEUE_BYTES
       Global      *wire.Budget // Core 全局 payload 预算；排队中的事件占用它
   }
   func New(log *slog.Logger, opt Options) *Router // 取代 goal-2 的 New(log, queueEvents)
   type Stats struct { Accepted, DroppedQueueFull, DroppedBudget uint64; Running, Queued int }
   func (r *Router) Stats() Stats

行为：
- Deliver：事件编码后的大小计入 QueueBytes 与 Global；等待中的事件数达到 QueueEvents、字节超过 QueueBytes 或 Global 无法预留时，立即返回 adapter.ErrQueueFull 并计数（分别计入 DroppedQueueFull、DroppedBudget），不阻塞调用方。事件开始执行时释放它在队列与 Global 中的占用。
- 并发：最多 Concurrency 个事件同时执行；同一事件的全部候选（命令层与事件路由）在一条链内按 (priority, plugin_id, route_id) 顺序执行，不同事件互不等待。链执行在独立 goroutine 中，不在任何 Session 的读循环里运行；慢 Handler 不影响其他 Session 的 Response、Cancel 与控制帧。
- 可信信息：Deliver 时把 Event 拆成 Core 保留部分（event_id、event_type、source_kind、adapter_id、adapter_instance_id、account_id、target、actor、message 的 platform_message_id、source_event_id，以及 button.clicked 的 action_id/action_schema_version/value、scheduled 的 job_id/run_id/scheduled_at）与业务 payload。每次构造 HandlerRequest 时用保留部分重建这些字段，业务 payload 只来自当前（后续 PreEvent 可改写的）副本；InvocationContext 的 actor 与 default_target 只取保留部分。T-05.3 的 PreEvent 只能改写业务副本。
- 身份：Actor 的 trusted_roles 在 Deliver 时按 T-04.4 的 TrustedRoles 计算一次并保留；没有 Actor 的事件（例如 scheduled 与 bot.online）不补用户，命令的 USER/ADMIN 检查一律不满足。
- Router.Close：停止接收新事件，等待运行中的链在调用方给出的 ctx 内结束，然后释放全部队列占用。
- 本切片不改变命令层与事件路由的匹配规则，也不实现 PreEvent、Rule 与传播的新语义。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/dispatch_test.go（//go:build unix）：构建一个测试插件，其事件路由按收到文本睡眠指定时长后回复，可用不同 priority 以三个插件 ID 启动。用 memory Adapter 与 Options{Concurrency: 2, QueueEvents: 3, QueueBytes: 很小的值} 装配：
  - 三个插件以不同顺序启动时，同一事件的调用顺序都是 (priority, plugin_id, route_id)。
  - 两个慢事件执行期间，第三个事件进入队列，再多的事件得到 ErrQueueFull，Stats 计数正确；慢事件结束后排队事件被执行。
  - 慢 Handler 执行期间，同一插件的 Session 仍能完成另一条 core.send 的 Response（读循环未被占住）。
  - 读入 contracts/testdata/events 的 button.clicked 与 scheduled fixture 并 Deliver：插件收到的 action_id/value、job_id/run_id 与 fixture 一致；scheduled 事件没有 Actor，需要 USER 的命令不可用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：QueueBytes 与 Global 的预留与释放在并发下正确（-race）；Close 时排队事件的占用被释放；trusted_roles 不能被插件自报的值改变；可信字段在重建时覆盖业务副本中的同名值。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：PreEvent 与出站 Hook（T-05.3）；Rule 与 HandlerOutcome 新语义（T-05.2）；延续调用（T-05.4）；实际的 scheduled tick（T-07.2）；按钮 token（T-03.8）。
```

### goal-21 · T-06.3 非持久 Event Bus

状态：等待前置（T-06.1）。语句中 goal-2、goal-19 提供的符号（router.Mux、router.Router、router.Plugin、sdk 的 ServiceRegistry 与分发、BUS_SUBSCRIPTION 的注册校验）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-21（切片 T-06.3：非持久 Event Bus）。goal 目录 goal-21，分支 goal-21，worktree .worktrees/goal-21，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件经 core.bus.publish 向自己命名空间下的事件名发布，Core 只确认完成发布准入；Core 按订阅注册的事件名、schema 与版本，以无业务回执的 EVENT 帧分发给每个已激活的订阅者；每个订阅者有独立的 256 条、4 MiB 队列，满时丢弃新事件并计数，离线不缓存、不补发；慢订阅者不阻塞其他订阅者与发布者。

依据：PLAN T-06 步骤 5；CONTRACTS §3 最后一条、§1.2；WIRE-FORMAT §5；MESSAGE-MODEL §1（Event Bus 段）；TEST-MATRIX V-17。

范围（只新增或修改这些）：
- internal/wire/services.go：末尾追加 ServiceBusPublish = "core.bus.publish"
- 新增 internal/core/bus/
- internal/core/router/：Router 在 Attach/Detach 时通知 bus（只加调用点）
- sdk/：新增 bus.go（发布与订阅）；sdk.go 的 Collect 与 dispatch.go 接入订阅回调（只加调用点）；session.go 的 buildRegister 输出订阅路由；base.go 的 Root 增加 Subscriptions Hook 与 Base 默认实现
- 新增 internal/core/supervisor/testdata/busplugin/
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/bus
   type Bus struct{ ... }
   type Options struct { QueueEvents int; QueueBytes int64 } // BUS_QUEUE_EVENTS、BUS_QUEUE_BYTES
   func New(log *slog.Logger, opt Options) *Bus
   func (b *Bus) Register(m *router.Mux) // 注册 core.bus.publish
   func (b *Bus) Attach(p router.Plugin) // 按插件的 BUS_SUBSCRIPTION 路由建立订阅与队列
   func (b *Bus) Detach(pluginID string) // 丢弃该插件队列中未发出的事件
   type Stats struct { Published, Delivered, Dropped uint64 }
   func (b *Bus) Stats(pluginID string) Stats // pluginID 为订阅者；Published 只对发布者有意义
2. sdk
   Root 增加 Hook：Subscriptions(*SubscriptionRegistry) // func (*Base) Subscriptions(*SubscriptionRegistry) 默认不订阅
   type SubscriptionRegistry struct{ ... }
   func Subscribe[T any](r *SubscriptionRegistry, routeID, eventName, schemaID string, version uint32, h func(*Context, T) error) // JSON 载荷；eventName 为 "plugin_id/event_name"
   func Publish(ctx context.Context, eventName, schemaID string, version uint32, payload any) (eventID string, err error) // eventName 只写本插件命名空间内的名字，SDK 补上 plugin_id 前缀

行为：
- Publish（core.bus.publish，载荷为 PublishRequest，响应为 PublishResponse）：event_name 必须属于调用者的命名空间（"<调用者 plugin_id>/" 前缀，名字部分同 Service 名规则），否则 PERMISSION_DENIED；encoding 为 JSON 时载荷必须是合法 UTF-8 JSON；schema_id 与版本非空。准入通过后分配 event_id（16 字节随机值的十六进制），把 BusEvent 依次放进每个匹配订阅者的队列（入队操作不等待订阅者），然后立即响应。没有订阅者也成功。发布不携带任何可信 Actor 权限。
- 匹配：订阅的 event_name、schema_id 与 schema_version 都相等才投递；版本不同的订阅者不收到，也不计为丢弃。
- 队列：每个订阅者独立；条数达到 QueueEvents 或字节超过 QueueBytes 时丢弃新事件（drop-newest）并计数，以 WARN 记录（订阅者、事件名、累计计数）。订阅者未激活或已 Detach 时不入队、不缓存，Core 重启或插件重启不补发。
- 投递：每个订阅者一个发送 goroutine，按队列顺序以 FRAME_KIND_EVENT 发出（Target 为 callback_route_id，载荷为 BusEvent），不等待业务回执；一个订阅者的 Session 慢或阻塞时只影响它自己的队列。
- SDK：收到 EVENT 时按 callback_route_id 找到订阅，把 payload 按 JSON 解码为 T（失败以 WARN 记录并丢弃，不回应），以订阅插件自身身份调用 h（Context 没有 Actor 与默认 Target，Send 必须显式 To）。h 的错误与 panic 只记日志。广播不参与普通事件的 StopPropagation。
- Publish 在 Handler 内外都可以调用；在 Handler 内也不继承用户权限。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/bus_test.go（//go:build unix）：构建 testdata/busplugin，以 pub、fast、slow 三个插件 ID 启动（slow 的回调阻塞到测试放行）。用 Bus{QueueEvents: 3} 装配：
  - pub 发布 10 条：fast 收到全部 10 条且顺序不变；slow 只收到前若干条，其余计入 Dropped；pub 的 Publish 全部立即返回。
  - pub 发布到 "fast/x"（不属于自己）返回 PERMISSION_DENIED。
  - schema 版本不同的订阅者收不到，Dropped 不变。
  - slow 停止后再发布：不入队；slow 重新启动后收不到之前的事件。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：QueueBytes 的边界；非法 JSON；没有订阅者；订阅回调的 Context 没有 Actor；解码失败只记日志；Detach 丢弃未发出的事件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：持久化与补发（契约不提供）；父调用链与取消（T-06.2）；Session 清理（T-06.4）。
```

### goal-26 · T-07.4 两个 Host 与插件票据

状态：等待前置（T-07.3）。语句中 goal-25 提供的符号（httpproxy.Server、httpproxy.Authorizer、settings.Settings、wire.HeaderProxyAuth）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-26（切片 T-07.4：两个 Host 与插件票据）。goal 目录 goal-26，分支 goal-26，worktree .worktrees/goal-26，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：同一公开 Listener 按受控 Host 区分 admin host 与 plugin host：plugin host 只提供 /plugins/{id}/... 与票据兑换端点，不提供任何管理 API；私有插件页面经 admin host 上已认证的管理会话签发 60s 一次性兑换票据，在 plugin host 兑换为独立的插件 Cookie 会话，会话不超出原管理会话、原会话失效即撤销；未配置第二主机名时只保留 JSON API 与 Webhook，并施加响应策略。

依据：PLAN T-07 步骤 6；CONTRACTS §7.2、§7.3、§15.1；TEST-MATRIX V-47 的组件分项。

范围（只新增或修改这些）：
- 新增 internal/core/webhost/（Host 分派、票据、插件会话与响应策略）
- internal/core/httpproxy/：Authorizer 由 webhost 的插件会话实现；转发前施加响应策略（只加调用点）
- cmd/alonabot/：run 改用 webhost 作为公开 Listener 的 Handler（替换 goal-25 的 nil Authorizer 与 shortcut 注释）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/webhost
   // AdminSessions 校验管理会话；T-09.1 提供实现，T-07 的测试注入假实现。
   type AdminSessions interface {
       Lookup(r *http.Request) (sessionID string, expiresAt time.Time, ok bool) // 只在 admin host 的请求上调用
       Valid(sessionID string) bool                                            // 管理会话是否仍有效（登出、过期或认证代次变化后为 false）
   }
   type Options struct {
       AdminBaseURL, PluginBaseURL *url.URL // 来自 settings；两者 scheme 与端口相同、主机名不同
       SecondHost bool                      // plugin_base_url 的主机名与 admin 不同时为 true
       Admin      http.Handler              // admin host 上的处理器（T-09 提供；为 nil 时 404）
       Proxy      *httpproxy.Server
       Sessions   AdminSessions
       Now        func() time.Time
   }
   type Host struct{ ... }
   func New(log *slog.Logger, opt Options) *Host
   func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request)
   func (h *Host) IssueTicket(adminSessionID, pluginID string) (ticketURL string, err error) // 供 admin host 的页面入口调用
   func (h *Host) Authorize(r *http.Request, pluginID string) bool // 实现 httpproxy.Authorizer

行为：
- Host 分派：只按 r.Host（去掉端口后不区分大小写比较）与 Options 中的两个主机名匹配，不读 Forwarded/X-Forwarded-Host（可信反代的处理在 T-09）；不匹配任何受控主机名返回 421。
  - admin host：/plugins/ 下的路径返回 404；其余交给 Options.Admin。
  - plugin host：只提供 /plugins/{id}/... 与 /_alonabot/ticket；其余（包括 /api/）返回 404。
  - 未配置第二主机名时（SecondHost=false），/plugins/{id}/... 仍在 admin host 上提供，但只用于 JSON API 与 Webhook：请求的 Accept 或插件响应的 Content-Type 为 HTML、JavaScript、SVG 等主动内容时返回 409，正文说明需要配置 plugin_base_url。
- 票据：IssueTicket 要求 Sessions.Valid(adminSessionID)；票据为 32 字节随机值的十六进制，绑定 pluginID、管理会话 ID 与到期时间（签发起 60s，命名常量），只能兑换一次，不写日志、不转发给插件。返回 plugin host 上的 /_alonabot/ticket?t=<票据>&next=/plugins/{id}/。
- 兑换（plugin host GET /_alonabot/ticket）：票据有效且管理会话仍有效时，建立插件会话（随机 ID，到期取管理会话到期时间），设置 Cookie __Host-alonabot_plugin（HTTPS 时；HTTP 时为 alonabot_plugin），HttpOnly、SameSite=Lax、Path=/、无 Domain，然后 303 重定向到 next（只允许以 /plugins/{同一 id}/ 开头），响应带 Referrer-Policy: no-referrer。票据无效、过期、已用或会话失效返回 403。
- Authorize：插件会话存在、未过期、绑定的插件等于 pluginID，且 Sessions.Valid(原管理会话) 为 true 才放行；原管理会话失效后派生会话立即失效。
- 响应策略（所有 /plugins/ 响应）：设置 X-Content-Type-Options: nosniff；未配置第二主机名时另加 Content-Security-Policy: default-src 'none'; frame-ancestors 'none'。删除插件响应中带 Domain 属性的 Set-Cookie、名称以 alonabot_ 或 __Host-alonabot_ 开头的 Set-Cookie，以及 Access-Control-Allow-Credentials；以 WARN 记录（插件 ID、Cookie 名）。
- Admin API 不向 plugin host 的 Origin 返回带凭证的 CORS 头（Admin 处理器由 T-09 实现，本切片在 admin host 的响应上删除 Origin 为 plugin host 时的 Access-Control-Allow-Origin 与 Allow-Credentials）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/webhost_test.go（//go:build unix）：沿用 goal-25 的 httpplugin，注入假的 AdminSessions，用 httptest 以不同 Host 请求：
  - plugin host 上 /api/x 为 404；admin host 上 /plugins/p/hello 为 404（配置了第二主机名时）；未知 Host 为 421。
  - 私有路由没有插件会话时 401；IssueTicket 后兑换得到 Cookie 与 303，Referrer-Policy 为 no-referrer，之后私有路由放行；同一票据第二次兑换 403；61s 后兑换 403（注入时钟）。
  - 管理会话失效后，派生的插件会话立即被拒。
  - 插件设置带 Domain 的 Cookie 与 alonabot_ 开头的 Cookie 被删除；nosniff 存在。
  - 未配置第二主机名时，JSON 响应正常、HTML 响应返回 409，CSP 存在。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Host 比较的大小写与端口；next 参数的开放重定向被拒；票据与插件 ID 不符；并发兑换只成功一次；Cookie 属性在 HTTP 与 HTTPS 下的差异。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./cmd/...；go build ./...。

不做：管理会话、登录与 admin-token（T-09.1）；Admin 页面与 API（T-09.8）；可信反代与 HTTPS 终止（T-09）；调度与统一启停（T-07.5）。
```

### goal-28 · T-08.1 配置字段描述与构建检查

状态：等待前置（T-04.2）。语句中 goal-11 提供的符号（sdk/generate 的 Scan、Result、Diagnostic，sdk.Generated、sdk.UseGenerated，plugin generate/check）只写名字；合并后由审查者核对签名，再改为可派发。本切片首次引入已批准的 github.com/google/jsonschema-go（D-43）。

```text
/goal 使用 $alonabot-execute 推进 goal-28（切片 T-08.1：配置字段描述与构建检查）。goal 目录 goal-28，分支 goal-28，worktree .worktrees/goal-28，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：生成器读取 Config 类型的 json、form、description、secret、hidden 标签，按 Go 声明顺序构造该插件的配置 Schema（jsonschema-go 的 *jsonschema.Schema），嵌入生成文件；对不支持的类型、自定义编解码、递归类型、错误默认值和越界数字给出带位置的诊断。SDK 在 REGISTER 中回传该 Schema 的摘要。

依据：PLAN T-08 步骤 1；CONTRACTS §5.1 第 2、3 条、§5.3（secret 路径）、§7.5（图标）；WIRE-FORMAT §3.2（config_schema_hash）；DECISIONS D-40、D-43、D-63；TEST-MATRIX V-25 的描述分项、V-83 的图标构建检查分项。

范围（只新增或修改这些）：
- go.mod、go.sum：新增 github.com/google/jsonschema-go（D-43 已批准；取当前最新版本，PR 中写明版本）
- sdk/generate/：新增 config.go（Config 类型分析与 Schema 构造）；生成文件嵌入 Schema
- sdk/：generated.go 的 Generated 增加 ConfigSchema；session.go 的 buildRegister 写入 config_schema_hash
- 新增 sdk/config/schema.go（Schema 的规范序列化与摘要，供生成器、SDK 与 Core 共用）
- 上述包的测试文件
允许的新跨模块边：sdk/generate -> sdk/config（同属 sdk 模块）。jsonschema-go 只在 sdk/config 与 sdk/generate 中 import。

标签语法（本切片定下，写在 sdk/config 的包文档中）：
- form:"key=value;key=value;flag"，键：label（i18 键）、widget（text、password、textarea、number、switch、select）、options（以 | 分隔的取值；取值后可以用 @ 附一个包内图标的相对路径，例如 options=openai@assets/icons/openai.svg|custom）、min、max（数字）、default（JSON 字面量）、required（无值的标志）。未知键是诊断错误。
- description:"i18 键"；secret:"true"；hidden:"true"。
- 字段的 i18 路径为 config.<json 路径>（嵌套结构用 . 连接），未写 label 时以它为 label 键。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/config
   func MarshalSchema(s *jsonschema.Schema) ([]byte, error) // 规范 JSON：PropertyOrder 保持声明顺序，其余键排序，无多余空白
   func SchemaHash(canonical []byte) []byte                // SHA256
   const (
       KeywordSecret = "x-alonabot-secret" // bool
       KeywordHidden = "x-alonabot-hidden" // bool
       KeywordForm   = "x-alonabot-form"   // 对象：label、widget、options、option_icons（取值 → 图标路径）、min、max
       KeywordI18n   = "x-alonabot-i18n"   // 字段的 i18 路径
   )
2. sdk/generate
   Result 增加 ConfigSchema []byte // 规范 JSON；没有 Config 时为 nil
3. sdk
   Generated 增加 ConfigSchema []byte
   buildRegister：有 ConfigSchema 时 config_schema_hash = config.SchemaHash(ConfigSchema)，否则为空。

行为：
- 类型映射（CONTRACTS §5.1 第 2 条）：string → string；bool → boolean；有符号与无符号整数 → integer，并按 Go 类型设置 minimum/maximum；float32/float64 → number；struct → object（properties 按声明顺序，PropertyOrder 同序）；slice → array；map[string]T → object 的 additionalProperties；指针 → 允许 null（types 含 "null"）。
- 对象一律不设 additionalProperties: false（未知键保留）；required 只包含 form 中声明了 required 的字段；其余字段可缺失。
- 诊断（带位置，继续检查其余字段）：上述之外的类型（interface、chan、func、complex、非 string 键的 map、time.Time 等）；类型实现了 json.Marshaler/json.Unmarshaler/encoding.TextMarshaler/TextUnmarshaler；递归类型；json 标签为 "-" 之外的重复名；default 不能解码为字段类型或超出整数范围、不在 options 中、超出 min/max；min > max；options 用于非 string/数字字段；选项图标的路径不在模块目录内、文件不存在或扩展名不是 .svg、.png、.webp；widget 与类型不符；secret 用于非 string 或 *string 字段（slice 与 map 元素中的 secret 字段允许，例如 []ProviderConfig 的 api_key）。类型分析使用 go/types（以源码方式导入扫描到的包），不用反射。
- Schema 的根为 object，$schema 为 draft 2020-12；secret、hidden、form、i18 路径分别写入对应 Extra 关键字；default 写入 Default；不使用 jsonschema.For 推导（D-43）。
- 生成文件把规范 Schema 作为字节常量嵌入，经 Generated.ConfigSchema 交给 SDK；同样输入生成的字节完全相同。plugin check 在 Schema 变化时报告生成文件过期（沿用 goal-11 的比较）。
- 生成的 Schema 必须能被 jsonschema 的 Resolve（开启 ValidateDefaults）通过；不通过时作为诊断报告。

测试（按编码规范「测试提交规则」）：
- 提交：sdk/generate/config_test.go。沿用 goal-11 的临时 module 方式，写一个含嵌套 struct、指针、slice、map、secret、hidden、required、default、options 的 Config：
  - Result.ConfigSchema 的规范 JSON 与期望一致（字段顺序为声明顺序，扩展关键字正确，无 additionalProperties: false）。
  - 生成后的程序启动时（测试入口打印 REGISTER 内容后退出），config_schema_hash 等于 SchemaHash(ConfigSchema)。
  - 第二次生成逐字节相同。
  - 不支持的类型、实现 json.Unmarshaler 的类型、递归类型、错误 default、越界 default、未知 form 键、不存在的选项图标与 .gif 选项图标各产生一条带位置的诊断；合法的选项图标写入 option_icons。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：每种整数类型的 minimum/maximum；MarshalSchema 的确定性；Resolve 对生成 Schema 的校验；json 标签的 omitempty 不影响 required；字段 i18 路径。

完成条件：gofmt -l 无输出；go mod tidy 后无差异；go vet ./...；go test -race ./sdk/... ./cmd/...；go build ./...。

不做：配置准备顺序、默认值补齐与 Schema 校验的运行时（T-08.3）；Secret Store（T-08.4）；i18 文件的检查（T-08.8）；表单组件（T-08.7）。
```

### goal-54 · T-11.1 工具描述、目录与策略

状态：等待前置（T-04.4、T-06.1）。语句中 goal-13、goal-19 提供的符号（policy.Policy、broker.Broker、sdk.JSONService、ServiceOption、router.Mux）只写名字；合并后由审查者核对签名，再改为可派发。jsonschema-go 若尚未由 goal-28 引入，本切片按 D-43 引入。

```text
/goal 使用 $alonabot-execute 推进 goal-54（切片 T-11.1：工具描述、目录与策略）。goal 目录 goal-54，分支 goal-54，worktree .worktrees/goal-54，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：JSON Service 可以声明为工具（给模型的说明、缺省策略 allow 或 confirm，输入 Schema 为该服务的输入 JSON Schema）；Core 校验工具描述，提供 core.tool.list 返回调用插件可以调用、且生效策略不是 deny 的工具；工具策略按 plugin_id/service_name 持久化，提供管理服务层的查询与修改，下一次调用生效。

依据：PLAN T-11 步骤 1、2；CONTRACTS §13.1、§13.2（策略段）、§3；WIRE-FORMAT §3.5（core.tool.list 与 Service 描述的 tool）；DECISIONS D-43；TEST-MATRIX V-55 的目录与策略分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：route.proto 的 ServiceSpec 追加 ToolSpec tool = 5；新增 tool.proto（ToolPolicy、ToolInfo、ToolListRequest/Response）。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- contracts/validate/register.go：工具描述的注册校验
- internal/wire/services.go：末尾追加 ServiceToolList = "core.tool.list"
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/tool.go
- 新增 internal/core/tool/
- sdk/：services.go 增加 Tool 选项与工具输入 Schema 的生成；新增 tool.go（ListTools）
- go.mod、go.sum：仅在 jsonschema-go 尚未引入时新增（D-43）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   enum ToolPolicy { TOOL_POLICY_UNSPECIFIED = 0; TOOL_POLICY_ALLOW = 1; TOOL_POLICY_CONFIRM = 2; TOOL_POLICY_DENY = 3; }
   message ToolSpec { string description = 1; ToolPolicy default_policy = 2; string input_schema_json = 3; } // default_policy 只能是 ALLOW 或 CONFIRM
   message ToolInfo { string name = 1; string description = 2; string input_schema_json = 3; ToolPolicy policy = 4; } // name 为 "plugin_id/service_name"
   message ToolListRequest {}
   message ToolListResponse { repeated ToolInfo tools = 1; }
2. internal/core/state
   迁移：CREATE TABLE tool_policies (plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), service_name TEXT NOT NULL, policy INTEGER NOT NULL, PRIMARY KEY (plugin_id, service_name));
   读写方法自定（放在 tool.go），按主键幂等。
3. internal/core/tool
   type Catalog struct{ ... }
   func New(store *state.Store, r *router.Router) *Catalog
   func (c *Catalog) Register(m *router.Mux)                                  // 注册 core.tool.list
   func (c *Catalog) Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error) // 覆盖优先，否则缺省策略；供 T-11.2 的 invoke
   func (c *Catalog) SetPolicy(ctx context.Context, pluginID, service string, p pb.ToolPolicy) error // UNSPECIFIED 表示删除覆盖
   func (c *Catalog) Policies(ctx context.Context) ([]ToolPolicyEntry, error) // type ToolPolicyEntry struct { PluginID, Service string; Default, Override pb.ToolPolicy }
4. sdk
   func Tool(description string, defaultPolicy ToolDefault) ServiceOption // type ToolDefault int：ToolAllow、ToolConfirm；只用于 JSONService
   type ToolInfo struct { Name, Description string; InputSchema json.RawMessage; Policy string } // Policy 为 "allow" 或 "confirm"
   func ListTools(ctx context.Context) ([]ToolInfo, error)

行为：
- SDK：JSONService 带 Tool 选项时，输入 Schema 由 jsonschema.For[In] 生成（工具输入可以拒绝未知键，与配置 Schema 不同，D-43 只约束配置），序列化为 input_schema_json；Tool 用于 ProtoService 是 Collect 错误。
- 注册校验：tool 只允许出现在 encoding 为 JSON 的服务上；description 非空且不超过 1024 字节（命名常量）；default_policy 为 ALLOW 或 CONFIRM；input_schema_json 必须能被 jsonschema 解析并 Resolve（不加载远程 $ref，ResolveOptions.Loader 为 nil），根类型为 object。任一失败整体拒绝注册。
- core.tool.list：遍历 Router 中已激活插件的工具服务，按 goal-19 的访问规则只保留调用插件可以调用的，再去掉生效策略为 DENY 的；按 name 字典序返回。调用者自己的工具同样按规则出现。每次调用都读取当前策略，不缓存。
- 策略：覆盖可为 ALLOW、CONFIRM、DENY；生效策略 = 覆盖或缺省。SetPolicy 在事务提交成功后才返回；插件或服务不存在时仍可设置（安装记录必须存在，否则 NOT_FOUND）。修改在下一次 list 与 invoke 时生效。
- 工具调用（core.tool.invoke）与确认在 T-11.2。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/tool_test.go（//go:build unix）：沿用 goal-19 的 svcplugin，增加带 Tool 选项的服务（PUBLIC 的 a/search 为 allow，ALLOWLIST: b 的 a/admin 为 confirm）；b 与 c 收到消息时调用 sdk.ListTools 并回复结果：
  - b 看到 a/admin 与 a/search，c 只看到 a/search；input_schema 与 In 的类型一致，policy 正确。
  - SetPolicy(a, search, DENY) 后两者都看不到 a/search；删除覆盖后恢复；关闭并重新打开状态库后覆盖仍在。
  - 注册一个 description 为空或 input schema 含远程 $ref 的工具时 REGISTER 失败。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Tool 用于 Protobuf 服务的 Collect 错误；description 长度边界；Policies 的输出；未知 ToolPolicy 被拒。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go mod tidy 后无差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：core.tool.invoke、确认与确认按钮（T-11.2）；Agent 组件与 MCP（T-12）；管理 API 与页面（T-09.9）。
```

### goal-67 · T-13.1 库模式入口与同进程组件

状态：等待前置（T-02.1、T-03.3、T-04.2）。语句中 goal-2、goal-4、goal-11、goal-76 提供的符号（router、message.Service、remote.Process、supervisor、sdk.Run 与 sdk.RunAdapter 的内部流程、sdk.Generated、tune.Parse）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-67（切片 T-13.1：库模式入口与同进程组件）。goal 目录 goal-67，分支 goal-67，worktree .worktrees/goal-67，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：根包 alonabot 提供 Run(opts ...Option)、Start(ctx, opts ...Option) 与 Plugin、Adapter、DataDir、Args 选项：在当前进程装配一个 Core，已安装的包照常以独立进程运行，开发中的插件与 Adapter 在 goroutine 中经 loopback TCP 走完整的握手、注册、激活与停止。组件身份取自生成器嵌入的 manifest；Core 为它们创建 dev 来源的安装记录与固定代次 dev。Start 在全部组件激活后返回句柄，ctx 取消时按停止预算停止；Run 是 Start 加 SIGINT/SIGTERM 处理。组件在预算内没有退出时 Wait 与 Run 返回错误。

依据：PLAN T-13 步骤 1、2；CONTRACTS §15.4、§4.2、§9（manifest）、§16；DECISIONS D-47；ARCHITECTURE「根包」一行；TEST-MATRIX V-64 的生命周期分项。

范围（只新增或修改这些）：
- 新增根包文件：alonabot.go（Run、Start、Core 与选项）、assembly.go（Core 装配，供 Run 使用，之后 T-09.4 的 serve 复用）
- 新增 internal/inproc/（同进程启动参数，外部模块无法构造）
- sdk/：新增 inproc.go（接收 internal/inproc 的启动参数，走与 Run、RunAdapter 相同的流程）；session.go 的 readLaunch 改为可由该参数提供（只加分支）
- sdk/generate/：生成文件嵌入构建入口目录下 plugin.yaml 的内容；sdk 的 Generated 增加 Manifest []byte
- internal/core/state/：Install 的 Source 允许 "dev"（只加常量）
- ARCHITECTURE.md 已登记新边（见下），执行者不改
- 新增 examples/devbot/（库模式示例：一个插件，以及示例内自带的回环 Adapter：实例启动后交付一条 "ping" 消息，收到发送时以 INFO 记录发送的文本）
- 上述包的测试文件
允许的新跨模块边：根包 -> internal/inproc、sdk/... -> internal/inproc（审查者已登记）；根包 -> internal/core/...、-> sdk/...、-> contracts，examples/devbot -> 根包（已登记）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 根包 alonabot
   type Option func(*options)
   func Run(opts ...Option) error // 等于 Start(signal.NotifyContext(..., SIGINT, SIGTERM)) 后 Wait
   // Core 是 Start 返回的运行中 Core；ctx 取消后开始停止。
   type Core struct{ ... }
   func Start(ctx context.Context, opts ...Option) (*Core, error) // 全部已启用的包与开发组件激活后返回；任一组件在激活前失败时，停止已启动的部分并返回错误
   func (c *Core) Wait() error // 等待停止完成；返回值与 Run 相同
   func Plugin(root sdk.Root) Option       // 根类型必须有生成器登记的 Generated（含 Manifest），否则 Run 返回错误
   func Adapter(a adapter.Adapter) Option  // sdk/adapter 的 Adapter；manifest 取自同一包登记的 Generated
   func DataDir(dir string) Option         // 缺省为工作目录下的 .alonabot/
   func Args(values ...string) Option      // 与 --args 相同的语法；ALONABOT_ARGS 环境变量先于 Args 应用（逗号分隔的直接参数或文件路径）
2. internal/inproc
   type Launch struct { PluginID string; Secret []byte; Addr string; BuildDigest []byte; RuntimeLockPath string } // 与环境变量版本的启动信息字段一致
3. sdk
   func RunInProcess(root Root, l *inproc.Launch) error            // l 为 nil 时返回错误
   func RunAdapterInProcess(a adapter.Adapter, l *inproc.Launch) error
   Generated 增加 Manifest []byte // 生成器嵌入的 plugin.yaml 原文；UseGenerated 的 Adapter 版本见下
   func UseGeneratedAdapter(a adapter.Adapter, g Generated) // 生成代码为 Adapter 包登记；a 为该包 Adapter 实现类型的 nil 指针

行为：
- Start 的顺序：解析 DataDir 与 Args（ALONABOT_ARGS 在前），出错立即返回；按 goal-76 的规则建立 tune.Values；打开日志与状态库（取得 core.lock）；装配 Router、message.Service、Mux、adapter.Registry 与 remote；按安装记录启动全部已启用的已安装包（独立进程，沿用 supervisor）；再启动开发组件；全部激活后返回 *Core。任一步失败时按停止顺序停掉已启动的部分、释放 core.lock 后返回错误。Start 不处理信号；Run 只负责把 SIGINT/SIGTERM 变成 ctx 取消，再 Wait。
- 开发组件：
  - 从 Generated.Manifest 解析 manifest（contracts/manifest），id 即插件 ID，role 决定 Plugin 还是 Adapter；Plugin(...) 传入 Adapter 角色的包或反之是错误。同一插件 ID 或 adapter_id 被登记两次，或与已安装包的 ID 冲突时，在启动任何组件前返回错误，不以后登记的覆盖，也不覆盖安装记录。
  - 创建或复用 source 为 dev 的安装记录（entrypoint 为空，build_digest 为本次运行生成的随机值），代次固定为 "dev"：首次运行 AddGeneration("dev") 并提交；以后复用。dev 来源不走候选与 Journal。
  - Core 为每个组件生成 Session Secret，经 inproc.Launch 交给 RunInProcess/RunAdapterInProcess 在 goroutine 中运行，连接 Core 的 loopback 端点；握手、Bootstrap、注册、激活与停止与独立进程相同。runtime.lock 路径为 <data_dir>/plugins/<id>/runtime.lock。
  - 配置：本切片 Bootstrap 的候选配置为 "{}"（config.json 的读取与 dev.sqlite 属于 T-13.2）。开发 Adapter 的实例：读取 <data_dir>/adapters/<adapter_id>/instances/*.json，文件名（去掉 .json）为 instance_id，内容原样作为实例配置，启动时全部 StartInstance；Secret 字段的抽取在 Secret Store 接入后处理，本切片用一行 shortcut 注释写明。
- 停止：ctx 取消时先停开发组件、再停已安装包，复用 host 的停止预算（STOP_TOTAL）。开发组件的 goroutine 在预算内没有返回时记 ERROR（插件 ID），Wait 与 Run 返回错误；已安装包按原流程停止。
- 开发组件没有进程隔离：SDK 未捕获的 panic 使整个进程退出，Core 不做有限重启；RunInProcess 不 recover 生命周期之外的 panic。
- 生成器：构建入口目录下存在 plugin.yaml 时把原文作为字节常量嵌入 Generated.Manifest；不存在时不嵌入（独立进程运行时不需要）。Adapter 包由 //alonabot:adapter 标记其 Adapter 实现类型（语法与诊断沿用 goal-11），生成 UseGeneratedAdapter。

测试（按编码规范「测试提交规则」）：
- 提交：alonabot_test.go（根包，//go:build unix）。测试构建 examples/devbot 并以子进程运行（临时 data_dir，预先写入一个回环 Adapter 实例文件），只通过日志、状态库与进程退出码断言：
  - 启动后日志中依次出现开发插件与开发 Adapter 的激活记录，代次为 dev，状态库中两条 source=dev 的安装记录。
  - 回环实例启动后交付一条 "ping" 的 message.created；开发插件的事件路由用 Context.Send 回复 "pong"（回复原 Target，不需要许可），回环 Adapter 收到发送后记录日志；测试从日志断言这次往返完成。
  - 再次启动复用同一 dev 代次，不新增代次。
  - 发送 SIGTERM：进程在 STOP_TOTAL 内以 0 退出。
  - 用 Args 把 STOP_TOTAL 设得很小、开发插件的 OnStop 阻塞：SIGTERM 后 Run 返回错误、进程以非 0 退出，日志有该插件的 ERROR。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：缺少 Generated 或 Manifest 时的错误；Plugin 传入 Adapter 角色；同一 ID 登记两次；与已安装包 ID 冲突；ALONABOT_ARGS 与 Args 的先后；instances 目录不存在时没有实例；inproc.Launch 为 nil；开发插件 OnStart 返回错误时 Start 返回错误，且同一 data_dir 能立即再次 Start（core.lock 已释放）；Start 成功后取消 ctx，Wait 在 STOP_TOTAL 内返回 nil。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./... ；go build ./...。

不做：config.json 直读、dev.sqlite 与拒绝 update/rollback（T-13.2）；热部署与 Reload（T-13.3）；alonabot serve（T-09.4，复用本切片的装配）；HTTP 与设置文件的装配（随 T-07.5、T-09.4 接入）。
```

## 第 7 波

### goal-8 · T-03.8 按钮、extension 片段与依赖解析

状态：等待前置（T-03.2、T-03.7）。语句中 goal-2～goal-5、goal-50 提供的符号（message.Service、policy.Policy、router 的派发、remote.Process、AdapterDescriptor 与 ExtensionDeclaration、Button.core_token、adapters/telegram 的实例）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-8（切片 T-03.8：按钮、extension 片段与依赖解析）。goal 目录 goal-8，分支 goal-8，worktree .worktrees/goal-8，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 为发送中的每个按钮签发持久化的 token（绑定插件、button.clicked 路由、代次、Target、value 与 24h 期限），Telegram 点击先 answerCallbackQuery，再由 Core 验证 token、当前 Actor 与策略后以普通请求调用正确的 Session；extension 片段按目标 Adapter 注册时声明的 Schema 在发送前校验；manifest 解析 role（plugin/adapter/agent/webui）、adapter_id、icon、WebUI 的 assets 与 Admin API 主版本范围、adapter_dependencies 与 plugin_dependencies，开发入口启用插件时检查必需依赖。

依据：PLAN T-03 步骤 7；CONTRACTS §8（extension、依赖与按钮段）、§9（manifest）、§7.5（图标与 WebUI）；MESSAGE-MODEL §1.1、§6.3；DECISIONS D-26、D-43、D-59、D-62、D-63；TEST-MATRIX V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto：Segment 的 oneof 追加 ExtensionSegment extension = 9。随后 go generate ./...（本切片获授权只追加这一分支与消息）
- contracts/wirespec_test.go：追加编号检查
- contracts/manifest/：manifest.go 增加 role、adapter_id、icon、assets、admin_api_min_major、admin_api_max_major、adapter_dependencies、plugin_dependencies 字段与按角色的校验；新增 deps.go（依赖检查）
- contracts/validate/register.go：button.clicked 事件路由的 action 校验
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/buttons.go
- 新增 internal/core/buttons/
- internal/core/message/：发送前签发按钮 token、校验 extension（只加调用点）
- internal/core/adapter/remote/：保存 AdapterDescriptor 的 extension 声明，供 message 校验
- internal/core/router/：button.clicked 改由 buttons 解析后派发（只加调用点）
- sdk/：events.go 增加 OnButton；message 包增加 Buttons 与 Extension 片段构造
- adapters/telegram/：callback_query 的处理与 answerCallbackQuery；extension 声明为空
- cmd/alonabot/：run 在启动插件前读取插件目录下的 plugin.yaml（存在时）并执行依赖检查
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. message.proto
   message ExtensionSegment { string adapter_id = 1; string name = 2; uint32 version = 3; string payload_json = 4; }
2. contracts/manifest
   Manifest 增加字段：Role string（plugin、adapter、agent 或 webui，缺省 plugin）；AdapterID string（adapter 时必填，小写，规则同本地 ID）；Icon string（可选，包内相对路径，扩展名为 .svg、.png 或 .webp）；Assets string（webui 必填，包内相对目录）；AdminAPIMinMajor、AdminAPIMaxMajor uint32（webui 必填，min ≤ max）；AdapterDependencies []AdapterDependency；PluginDependencies []PluginDependency
   type AdapterDependency struct { AdapterID string; Extensions []ExtensionRequirement; Required bool }
   type ExtensionRequirement struct { Name string; Version uint32 }
   type PluginDependency struct { PluginID string; MinVersion string; Required bool } // MinVersion 为空表示不限
   type Installed struct { Plugins map[string]string; Adapters map[string][]ExtensionRequirement } // 插件 ID→版本（只含已启用的）；adapter_id→已声明的 extension
   type Problem struct { Dependency string; Required bool; Reason string }
   func CheckDependencies(m *Manifest, in Installed) []Problem // 只报告问题，必需项的问题由调用方据此拒绝
3. internal/core/buttons
   type Service struct{ ... }
   func New(store *state.Store, now func() time.Time) *Service
   func (s *Service) Issue(ctx context.Context, pluginID, generationID string, target *pb.Target, chain *pb.MessageChain) error // 为 chain 中每个按钮写入 core_token 并持久化映射
   func (s *Service) Resolve(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error) // 校验 button.clicked 事件并改写 payload 为插件注册的 action 与 value
   func (s *Service) Purge(ctx context.Context, pluginID string) error // 卸载时调用
4. sdk
   func (r *EventRegistry) OnButton(routeID, actionID string, schemaVersion uint32, handler func(*Context) error, opts ...RouteOption)
   func (c *Context) ButtonValue() string // button.clicked 的 value；其他事件为空
   message 包：func Buttons(rows ...[]Button) Segment；type Button struct { ActionID, Label, Value string }；func Extension(adapterID, name string, version uint32, payload any) Segment // payload 用 encoding/json 编码

行为：
- 按钮签发（message.Service 在 BeforeSend 之前、adapter.send 之前）：插件给出的 core_token 一律清除；每个按钮的 action_id 必须对应发送插件当前注册的 button.clicked 路由（EventSpec.action_id），否则 INVALID_ARGUMENT。token 为 32 字节随机值的无 padding base64url（43 字节），持久化 (token, plugin_id, route_id, action_id, action_schema_version, generation_id, target 的 instance/chat/thread, value, expires_at=签发起 BUTTON_TOKEN_TTL)。value 不超过 1024 字节（命名常量）。持久化成功后才把 token 写入 core_token 并继续发送。
- Telegram 点击：callback_query 先调用 answerCallbackQuery（失败只记日志），再交付 button.clicked：Target 取 callback_query.message 的 chat 与 thread（没有 message 的 inline 回调不交付，推进游标并以 INFO 记录），Actor 取 from，payload 的 value 暂放 callback_data，origin_platform_message_id 为消息 ID，callback_query_id 原样。
- Core 解析（Router 在派发 button.clicked 前调用 Resolve）：token 不存在、过期、generation 已不是该插件的当前 Active、或 Target 与签发时不同 → 丢弃并以 INFO 记录，不派发、不回复、不进入 Raw 或文本匹配。有效时把 payload 改写为签发时的 action_id、action_schema_version 与 value，只派发给该插件的该路由；派发前执行 T-04.4 的 CheckCommand 等价检查（全局禁止与路由启用、权限），每次点击都按本次 Actor 重新检查；同一 token 可被多人、多次点击。token 不带任何角色。
- extension：remote.Process 保存每个 Adapter 的 ExtensionDeclaration。发送前对每个 extension 片段：目标实例所属的 adapter_id 必须等于片段的 adapter_id，(name, version) 必须已声明，payload_json 必须通过该声明的 schema（jsonschema-go 的 Resolve 与 Validate，Loader 为 nil）；否则 UNSUPPORTED（Adapter 不符或未声明）或 INVALID_ARGUMENT（不符合 Schema），在调用 Adapter 之前返回。Telegram 本切片不声明 extension，所有 extension 片段发往 Telegram 都是 UNSUPPORTED。
- manifest：按 CONTRACTS §9.1 校验各角色：role 为 adapter 时必须有 adapter_id，且不能有 adapter_dependencies、plugin_dependencies、migration_mode 或 prepare_effects；agent 与 plugin 的字段规则相同；webui 必须有 assets 与 Admin API 主版本范围、os/arch 为 any，不能有 entrypoint、协议主版本、required_capabilities、config_version 与依赖；市场 ID 的前缀与 role 一致（alonabot_plugin_、alonabot_adapter_、alonabot_agent_、alonabot_webui_）；依赖项的 ID 格式、版本 SemVer、重复项都要校验；必需的 plugin_dependencies 不能直接依赖自己。环的检测需要全部安装记录，由 T-09.4 完成。
- CheckDependencies：必需的 Adapter 依赖要求该 adapter_id 已安装且声明了全部所需 extension（版本相等）；必需的插件依赖要求已启用且版本 ≥ MinVersion；可选依赖缺失也报告（Required=false），调用方只记诊断。
- cmd/alonabot run：插件目录下有 plugin.yaml 时解析并检查依赖（Installed 来自状态库中已启用的安装记录与当前运行的 Adapter 描述），有必需项问题时以 exitFailure 退出并列出全部缺失项。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/buttons_test.go（//go:build unix）：memory Adapter 与测试插件（注册 OnButton("vote", "vote", 1, ...)，收到 /poll 时发送两个按钮），用真实状态库装配：
    - 发送后 memory 收到的按钮都有 43 字节的 core_token，状态库有映射；插件伪造的 core_token 被替换。
    - 注入以该 token 为 callback_data 的 button.clicked：插件的 OnButton 被调用，ButtonValue 为原 value；两个不同用户点击各调用一次。
    - 未知 token、过期 token（注入时钟）、另一个 chat 中的点击都不调用插件，也不进入 Raw。
    - 按钮指向未注册的 action 时发送被拒，memory 没有收到。
    - 插件产生新的 Active 代次后，旧 token 失效。
    - extension 片段发往不声明它的 Adapter 返回 UNSUPPORTED；发往声明了它的 memory Adapter 时 payload 不符合 Schema 返回 INVALID_ARGUMENT，符合时发出。
  - adapters/telegram/callback_test.go：telegramtest 推入 callback_query：先有 answerCallbackQuery 调用，再交付 button.clicked；没有 message 的 inline 回调不交付但推进游标。
  - contracts/manifest 的现有集成测试（若有）追加四种 role、icon、WebUI 字段与依赖字段的解析。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：manifest 新字段的每种非法值；CheckDependencies 的判定表；token 的编码长度与唯一性；value 长度边界；Purge；Resolve 对 generation 的判断。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/... ./cmd/...；go build ./...。

不做：正式管理入口的依赖检查、确认与环检测（T-09.4）；消息编辑（T-03.9）；BeforeSend/AfterSend Hook（T-05.3）；Telegram 的 extension（之后按需由 Adapter 声明）。
```

### goal-51 · T-03.9 消息编辑

状态：等待前置（T-03.7）。语句中 goal-2～goal-4、goal-50 提供的符号（message.Service、remote.Process、sdk/adapter 的 Instance、adapters/telegram 的发送与渲染）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-51（切片 T-03.9：消息编辑）。goal 目录 goal-51，分支 goal-51，worktree .worktrees/goal-51，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件经 core.edit 修改一条已发送的平台消息：Core 按与 Send 相同的 Target 准入，校验 Edit Chain 只含一个 text 片段和至多一个 buttons 片段，经 adapter.edit 交给 Adapter；官方 Telegram Adapter 用 editMessageText 实现并声明 message.edit 能力；平台拒绝时按 MESSAGE-MODEL §4 报告，不改为发送新消息。

依据：PLAN T-03 步骤 6（编辑）；CONTRACTS §8（Edit 段）；MESSAGE-MODEL §4、§6.2（Edit 条）；WIRE-FORMAT §3.4、§3.5；TEST-MATRIX V-53。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：message.proto 新增 EditRequest；adapter.proto 新增 AdapterEditRequest。随后 go generate ./...（本切片获授权只追加这两个消息）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceEdit = "core.edit"、ServiceAdapterEdit = "adapter.edit"
- internal/core/adapter/：新增 edit.go（可选接口 Editor）；remote 与 memory 实现它
- internal/core/message/：新增 edit.go（core.edit）
- sdk/：Context 与后台各增加 Edit；sdk/adapter 的 Instance 可选实现 Edit；RunAdapter 处理 adapter.edit
- adapters/telegram/：新增 edit.go；Capabilities 增加 message.edit
- adapters/telegram/telegramtest/：增加 editMessageText 与故障注入
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message EditRequest { Target to = 1; string platform_message_id = 2; MessageChain chain = 3; } // to 缺省时用默认目标；响应为 SendResult
   message AdapterEditRequest { string instance_id = 1; Target target = 2; string platform_message_id = 3; MessageChain chain = 4; } // 响应为 SendResult
2. internal/core/adapter
   type Editor interface { Edit(ctx context.Context, target *pb.Target, platformMessageID string, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) }
3. sdk
   func (c *Context) Edit(platformMessageID string, chain message.Chain, opts ...SendOption) (*pb.SendResult, error)
   func Edit(ctx context.Context, to *pb.Target, platformMessageID string, chain message.Chain) (*pb.SendResult, error) // 后台编辑，必须给出 to
4. sdk/adapter
   type Editor interface { Edit(ctx context.Context, target *pb.Target, platformMessageID string, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) } // Instance 可选实现；未实现时 adapter.edit 返回 UNSUPPORTED

行为：
- core.edit（Core）：确定 Target 与准入的步骤与 core.send 完全相同（默认目标、NO_DEFAULT_TARGET、policy.CheckSend、实例与账号绑定）。platform_message_id 为空 → INVALID_ARGUMENT。Chain 必须恰好一个 text 片段、至多一个 buttons 片段，不能有其他片段，否则 INVALID_ARGUMENT。目标实例的 Capability 没有 supported 的 message.edit，或实例不实现 Editor → UNSUPPORTED。按钮的 token 签发与 Send 相同（T-03.8）。然后调用 Editor.Edit，原样返回 SendResult 与 ErrorInfo。不自动重试，失败时不改为发送新消息。
- Hook：BeforeSend/AfterSend 的输入标明操作为 edit 与被编辑的 platform_message_id，由 T-05.3 接入；本切片在 core.edit 中留出与 Send 相同的位置，以一行 shortcut 注释写明。
- Telegram：editMessageText（chat_id、message_id、message_thread_id 不传）；text 的格式与实体规则同发送；buttons 转为 reply_markup。平台返回 "message is not modified" 或消息不可编辑时为 REJECTED（ErrorInfo 带平台描述）。只编辑文本消息，不编辑媒体说明（目标为媒体消息时平台拒绝，同样 REJECTED）。成功时 SendResult 为 ACCEPTED，platform_message_ids 为被编辑消息的 ID。
- 能力：Telegram 实例启动时声明 message.edit（version 1，supported=true）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - adapters/telegram/edit_test.go：telegramtest 驱动：编辑成功的参数正确；"message is not modified" 为 REJECTED；带 buttons 时 reply_markup 正确；Calls() 中没有任何 send 方法。
  - internal/core/supervisor/edit_test.go（//go:build unix）：memory Adapter 实现 Editor 并声明 message.edit；测试插件收到消息后发送一条再编辑它：memory 先收到发送、再收到编辑，ID 一致；Chain 含图片时被拒；memory 不声明 message.edit 时返回 UNSUPPORTED；编辑其他会话（没有许可）返回 PERMISSION_DENIED。
  - contracts/wirespec_test.go：新消息的编号检查。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Edit Chain 的每种非法组合；后台 Edit 缺少 to；adapter.edit 在实例不实现 Editor 时的 UNSUPPORTED；Telegram 编辑的实体 UTF-16 偏移。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。

不做：出站 Hook 的执行（T-05.3）；编辑媒体或说明；message.edited 入站事件（已在 T-03.4）。
```

### goal-16 · T-05.2 Rule、HandlerOutcome 与传播

状态：等待前置（T-05.1）。语句中 goal-2、goal-12、goal-15 提供的符号（sdk 的分发、RouteOption、Context、HandlerOutcome 的处理、router 的命令层与事件路由派发）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-16（切片 T-05.2：Rule、HandlerOutcome 与传播）。goal 目录 goal-16，分支 goal-16，worktree .worktrees/goal-16，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：SDK 提供在同一 RPC 内执行的 Rule：多个 Rule 默认 AND，显式 OR 用 Any 组合，内置 ToMe 与 ChatAdmin；Rule 为 false 或出错时不运行目标函数并如实返回 matched 与错误。固定两种 Handler 签名的 HandlerOutcome、StopPropagation 锁存、错误分支与已识别命令的消费语义，并用测试证明发送后失败不会补发、撤销或自动回复错误。

依据：PLAN T-05 步骤 3、4；CONTRACTS §2.1、§2.2 第 3、6 条、§2.3 第 1 条、§8（自动回复段）；MESSAGE-MODEL §1.1（to_me）；TEST-MATRIX V-14、V-11 的 Rule 分项。

范围（只新增或修改这些）：
- sdk/：新增 rule.go；dispatch.go 在调用目标函数前执行 Rule；registry.go 与 events.go 接受 Rule 选项；session.go 的 buildRegister 写入 RuleDescriptor
- internal/core/router/：HandlerOutcome 的处理按下文统一（只改派发处的判断）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
sdk
   type Rule func(*Context) (bool, error)
   func WithRule(rules ...Rule) RouteOption // 可多次使用，全部 AND
   func Any(rules ...Rule) Rule              // OR：依次求值，第一个 true 即为 true；全部 false 为 false；出错立即返回错误
   func ToMe() Rule                           // message.created 的 to_me；其他事件为 false
   func ChatAdmin() Rule                      // 调用 Actor 的 trusted_roles 含 CHAT_ADMIN
   func CoreAdmin() Rule                      // 调用 Actor 的 trusted_roles 含 ADMIN（Core 判定的机器人管理员）

行为：
- Rule 在 SDK 的同一次 Handler 请求中、目标函数之前执行，顺序为注册顺序；任一 Rule 返回 false 时不运行目标函数，HandlerOutcome{matched:false, propagation:CONTINUE}；任一 Rule 返回错误时不运行目标函数，按 goal-2 的错误规则返回 Err（*sdk.Error 用其 Code，其他为 INTERNAL），matched=false。Rule panic 与 Handler panic 一样转 INTERNAL 并记录堆栈。
- RuleDescriptor：有任何自定义 Rule 时 has_custom_rule=true；只用 ChatAdmin 时 require_chat_admin=true（Core 不据此预先过滤，只用于诊断显示）。
- 传播：
  - StopPropagation 在调用的任何时刻调用都锁存，即使之后 Rule false、Handler 返回错误或 panic，HandlerOutcome 仍为 STOP。
  - Core 对事件路由与文本匹配/Raw：matched=false 跳过；业务错误以 WARN 记录后继续下一个候选；STOP 终止后续候选（含 D-44 中命令层之后的事件路由）。
  - 已识别命令的胜出路由 Rule false 或出错时仍消费命令层（不回落到短路径、文本匹配或 Raw），不回复 usage；STOP 的效果同上。
- 自动回复型 Handler 只在成功且 Chain 非空时由 Core 发送一次；返回错误时即使 Chain 非空也不发送；自动型内调用 Send 返回 FAILED_PRECONDITION（goal-12 已实现，本切片补齐测试）。手动型可多次 Send；返回错误时已发送的消息不撤销，Core 不自动发送错误回复。
- Context 在 Handler 返回后失效：之后的 Send、Edit、Wait 返回 FAILED_PRECONDITION；后台 goroutine 持有的 Context 不能改变已返回的 HandlerOutcome。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/rule_test.go（//go:build unix）。沿用 goal-12 的命令测试装配，测试插件注册：带 ToMe 的事件路由、带 ChatAdmin 的命令、带返回错误的自定义 Rule 的 keyword 路由、先 Stop 再返回错误的路由、发送两次后返回错误的手动命令、在自动型里调用 Send 的命令、在后台 goroutine 中于返回后调用 Send 的路由：
  - to_me=false 的群消息不触发 ToMe 路由，私聊触发。
  - 非 chat_admin 用户的命令被消费：不回 usage、不进入 Raw，插件目标函数未运行。
  - Rule 出错的 keyword 路由不运行，后续候选继续；先 Stop 再出错的路由使后续候选不再调用。
  - 手动命令的两条消息都已送达，之后没有任何补发或错误回复；自动型里的 Send 失败且只有返回的 Chain 被发出一次；返回后的后台 Send 返回 FAILED_PRECONDITION。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Any 的短路与错误；WithRule 多次使用的 AND；RuleDescriptor 的两个字段；Rule panic；Stop 在 Rule 中调用的锁存。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：PreEvent 与出站 Hook（T-05.3）；延续调用与多轮等待（T-05.4）；Core 侧预先执行 chat_admin 过滤。
```

### goal-24 · T-07.2 schedule 注册与运行槽

状态：等待前置（T-05.1、T-07.1）。语句中 goal-2、goal-15、goal-23（contracts/schedule）、goal-76 提供的符号（router.Router、router.Plugin、wire.Session 的 Call、schedule.Trigger、tune.Values、sdk 的 ScheduleRegistry 与分发）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-24（切片 T-07.2：schedule 注册与运行槽）。goal 目录 goal-24，分支 goal-24，worktree .worktrees/goal-24，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件注册的 interval/cron 定时任务经 REGISTER 校验后由 Core Scheduler 按 contracts/schedule 计算触发：每次触发以 Core→插件 REQUEST 调用任务（独立 run_id、JOB_TIMEOUT 期限、无 Actor、不默认 Admin），同一 job 有运行中的调用时跳过本次；到期只结束调用结果并取消，运行槽保留到插件真正返回（迟到响应只释放槽）或旧进程确认退出；不补跑错过的 tick。

依据：PLAN T-07 步骤 1（schedule）、3；CONTRACTS §4.3、§1.2；MESSAGE-MODEL §1.1（scheduled）、§1 末段；WIRE-FORMAT §4；TEST-MATRIX V-20、V-46、V-16 与 V-05 的 scheduled 分项。

范围（只新增或修改这些）：
- contracts/validate/register.go：SCHEDULE 路由的注册校验（调用 schedule.FromSpec）
- internal/wire/session.go：Request 增加迟到响应回调（只加字段与触发点）
- 新增 internal/core/scheduler/
- sdk/：sdk.go 去掉「不支持 scheduled jobs」的拒绝（只改该处）；dispatch.go 分发 SCHEDULE 路由；session.go 的 buildRegister 输出 SCHEDULE 路由
- 新增 internal/core/supervisor/testdata/jobplugin/
- 上述包的测试文件
允许的新跨模块边：internal/core/... -> contracts/schedule、contracts/validate -> contracts/schedule（同属 contracts 与已允许的 internal/core -> contracts）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/wire
   Request 增加字段：Late func(Response) // 非 nil 时，若响应在调用方期限之后才到达，读循环调用一次；期限内到达时不调用
2. internal/core/scheduler
   type Scheduler struct{ ... }
   type Options struct { JobTimeout time.Duration; Now func() time.Time; Log *slog.Logger } // JobTimeout 取 tune.Values.JobTimeout
   func New(opt Options) *Scheduler
   func (s *Scheduler) Attach(p router.Plugin, activatedAt time.Time) // 为插件的 SCHEDULE 路由安排下一次触发；anchor 为激活时刻
   func (s *Scheduler) Detach(pluginID string)                        // 停止安排新的触发；运行槽保持，直到 Released
   func (s *Scheduler) Released(pluginID string)                       // 旧进程确认退出后调用：释放该插件的全部运行槽
   func (s *Scheduler) Close()
   type JobStats struct { Runs, Skipped, TimedOut, Late uint64; Running bool }
   func (s *Scheduler) Stats(pluginID, jobID string) JobStats

行为：
- 注册校验：SCHEDULE 路由的 route_id 为 "job." + job_id；job_id 同插件内唯一；ScheduleSpec 必须恰有 interval_ms 或 cron 之一，经 schedule.FromSpec 校验（interval 至少 1s、cron 五字段、时区合法）；SDK 的 Every/Cron 生成对应描述。
- 触发：每个 job 用 Trigger.Next(anchor, now) 得到下一次时刻并设定定时器；到点时若该 job 的运行槽空闲则发起一次调用，否则 Skipped 加一、不排队；之后再按 Next 安排（不补跑，跳过的 tick 不累积）。
- 调用：Request{Target: route_id, Payload: HandlerRequest{route_id, context:{actor 为空, default_target 为空}, event: scheduled 事件（event_id 由 Core 生成、source_kind=SCHEDULER、payload 为 job_id、run_id（16 字节随机值的十六进制）、scheduled_at_unix_ms）}, Deadline: 触发时刻 + JobTimeout, Late: 释放运行槽}。直接调用插件 Session，不经事件队列与派发链。
- 期限：到期时调用方得到 DEADLINE_EXCEEDED、Cancel 已发出，TimedOut 加一，但运行槽保留；迟到响应（Late 回调）到达时释放槽并计入 Late。期限内的响应（成功或错误）立即释放槽。Session 断开时不释放槽，直到 Released 被调用。
- SDK：收到 SCHEDULE 路由的请求时以 Context（没有 Actor 与默认 Target，Send 必须显式 To）调用任务函数；ctx 被取消时仍等用户函数真正返回后才发送唯一的 Response。
- 本切片不接统一启停（Draining 时停止新 tick 等属于 T-07.5），也不改 cmd/alonabot 的装配。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/scheduler_test.go（//go:build unix）。构建 testdata/jobplugin（一个每 1s 的 interval 任务，按环境变量决定执行时长并以日志记录 run_id），Launch 后 Attach 到 Scheduler（JobTimeout 设为 2s）：
  - 执行 0.2s 时，若干秒内有相应次数的调用，run_id 各不相同，事件没有 Actor。
  - 执行 5s 时：第 2s 到期后 TimedOut=1、Running 仍为 true，期间到点的 tick 被 Skipped；第 5s 迟到响应到达后 Late=1、槽释放，下一次 tick 正常调用。
  - 任务执行中杀掉插件进程：槽不释放；调用 Released 后释放，新进程 Attach 后从未来的 tick 开始，不补跑。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：注册校验的各条拒绝；wire 的 Late 回调只在期限后触发且只一次；cron 任务用注入时钟的 Next；Detach 后不再触发；Close 释放定时器。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：Scheduler 与 HTTP 的统一启停、Draining 与停止预算（T-07.5）；Agent 的定时任务（T-12.6）；持久化的运行记录。
```

### goal-30 · T-08.3 SDK 配置准备顺序

状态：等待前置（T-08.1、T-08.2）。语句中 goal-28 提供的符号（sdk.Generated.ConfigSchema、sdk/config 的扩展关键字与 MarshalSchema）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-30（切片 T-08.3：SDK 配置准备顺序）。goal 目录 goal-30，分支 goal-30，worktree .worktrees/goal-30，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：SDK 在准备阶段严格按「OnLoad(fromVersion, raw) → 只给缺失键补默认值 → typed decode 与 Schema 校验」处理 Bootstrap 下发的候选配置，再执行 OnStart，并在 REGISTER 回传完整、规范化的候选 JSON（保留未知键，不用 typed marshal 覆盖原文档）。普通 restart 时 OnLoad 在相同配置版本上改变配置则失败。

依据：PLAN T-08 步骤 3；CONTRACTS §5.1 第 3 条、§5.2；WIRE-FORMAT §3.2；DECISIONS D-40、D-43；TEST-MATRIX V-37、V-21 的配置分项。

范围（只新增或修改这些）：
- 新增 sdk/config/prepare.go（准备流程，供 sdk 调用）
- sdk/session.go：prepare 在 OnStart 之前调用配置准备，REGISTER 使用准备后的 JSON（只改该段）
- 上述包的测试文件
允许的新跨模块边：无（jsonschema-go 已由 goal-28 引入）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
sdk/config
   type Prepared struct {
       JSON    []byte // 规范 JSON：键排序、无多余空白、数字保留原文；包含全部未知键
       Version uint32 // target_config_version
   }
   // Prepare 按顺序执行：OnLoad → 补缺失键默认 → typed decode 到 cfg → Schema 校验；任何一步失败返回带 Code 的错误。
   func Prepare(ctx context.Context, cfg Config, schema []byte, raw []byte, fromVersion *uint32, targetVersion uint32, restart bool) (Prepared, error)

行为：
- 输入：raw 为 Bootstrap.candidate_config_json；fromVersion 为 previous_config_version（首次安装为 nil，此时以 targetVersion 调用 OnLoad）；restart 为 true 表示 activation_reason 是普通 restart 或恢复。没有 Config（Root.Config 返回 nil）时跳过前三步，原样回传 raw 的规范形式。
- OnLoad：在零值的配置类型上调用（cfg 指向新的零值实例），输入完整原 JSON；返回值必须是 JSON 对象，否则 INVALID_ARGUMENT；OnLoad 返回错误时为 INVALID_ARGUMENT（safe_message "config OnLoad rejected the configuration"，原文只写日志）。OnLoad 的预算是 Bootstrap 的 config 阶段预算，超时为 DEADLINE_EXCEEDED。
- restart 时，OnLoad 的结果与输入的规范形式不同 → FAILED_PRECONDITION（不能借 restart 修改 Active 配置）。
- 补默认值：用 jsonschema-go 的 Resolve（ValidateDefaults=true）与 ApplyDefaults，只给对象中缺失的属性补默认值；显式 false、0、""、[]、{} 与 null 都不覆盖；required 字段缺失且没有默认值时在下一步校验失败。
- typed decode：把补完默认值的 JSON 用 encoding/json 解码到 cfg（数字超出字段范围、类型不符 → INVALID_ARGUMENT，错误指明 JSON 路径）；null 只允许出现在指针字段上。未知键不进入 cfg，但保留在 JSON 中。
- Schema 校验：用 Resolved.Validate 校验补完默认值后的 JSON 对象（以 UseNumber 解码的副本）；失败 → INVALID_ARGUMENT，错误指明 JSON 路径，不含取值。大整数在校验副本中可能失去精度时，以 typed decode 的范围检查为准（D-43）。
- 输出：Prepared.JSON 为补完默认值后的完整对象的规范形式（复用 internal/core/config.Canonical 的同一算法，在 sdk/config 中实现，不 import internal/core），Version 为 targetVersion。REGISTER 的 candidate_config_json 与 candidate_config_version 使用它们；SDK 不以 typed marshal 重建文档。
- 失败时 SDK 按现有流程发送 SESSION_ERROR（阶段 CONFIG），不进入 OnStart；错误文本不含配置取值。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/config_prepare_test.go（//go:build unix）：构建一个配置类型含 default、required、指针、嵌套与未知键场景的测试插件（经 goal-28 的生成器生成 Schema），以不同 Bootstrap 启动：
  - 缺失键得到默认值；显式 false、0、""、null（指针字段）保留；未知键与嵌套未知键出现在 REGISTER 的 candidate_config_json 中。
  - 旧版本配置（fromVersion=1）只有经 OnLoad 迁移后才满足新 Schema：启动成功；不覆盖 OnLoad 时同样的输入在校验阶段失败（阶段 CONFIG）。
  - required 缺失、int 字段超出范围、OnLoad 返回错误时，SESSION_ERROR 阶段为 CONFIG，OnStart 未执行。
  - restart 时 OnLoad 改变了配置：FAILED_PRECONDITION；OnLoad 保持不变时正常启动。
  - 超过 2^53 的整数原样保留在回传 JSON 中。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：规范 JSON 与 internal/core/config.Canonical 的结果逐字节一致（同一组输入）；ApplyDefaults 不触碰显式值；没有 Config 时的路径；OnLoad 超时；错误信息不含取值。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./sdk/... ./internal/core/...；go build ./...。

不做：Secret 的引用与解引用（T-08.4）；DB 迁移（T-08.5）；Pending 激活与 reload 流程（T-09）；表单（T-08.7）。
```

### goal-31 · T-08.4 不可变 Secret Store

状态：等待前置（T-08.1、T-08.2）。语句中 goal-28、goal-29 提供的符号（sdk/config 的 KeywordSecret、MarshalSchema，internal/core/config 的 Canonical）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-31（切片 T-08.4：不可变 Secret Store）。goal 目录 goal-31，分支 goal-31，worktree .worktrees/goal-31，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：配置快照中的 secret 字段只保存 {"$secret_ref": id}，Secret Store 保存不可变的 secret 版本。Core 在构造 Bootstrap 前把本插件有权访问的引用解析为明文，在 REGISTER 后按配置 Schema 的 secret 标记与此前保留的 secret 路径重新抽取，生成持久化快照，并以快照计算 PREPARED 的配置摘要；管理读取只返回引用是否已设置；按引用清理时保护 Active、Pending 与保留代次。

依据：PLAN T-08 步骤 4；CONTRACTS §5.3、§5.2（PREPARED 摘要）、§14.5、§12；WIRE-FORMAT §3.2；TEST-MATRIX V-25。

范围（只新增或修改这些）：
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/secrets.go
- 新增 internal/core/secret/
- internal/core/host/prepare.go：Expect 增加快照回调；PREPARED 的摘要改为对快照计算；Prepared 增加 ConfigSnapshot
- internal/core/supervisor/supervisor.go：Spec 增加 ConfigSchema；新候选时把快照写入代次，普通 restart 时核对快照一致
- 上述包的测试文件
允许的新跨模块边：internal/core/secret -> sdk/config 不允许（internal/core 不 import sdk）；secret 标记的关键字名在 internal/core/secret 中以常量重复声明，并由测试核对与 sdk/config 的常量一致。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：
   CREATE TABLE secrets (ref_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), value BLOB NOT NULL, created_at INTEGER NOT NULL);
   CREATE TABLE generation_secret_paths (generation_id TEXT NOT NULL REFERENCES generations(generation_id), path TEXT NOT NULL, PRIMARY KEY (generation_id, path));
   func (s *Store) SetGenerationConfig(ctx context.Context, generationID, snapshot string) error // 只允许未提交的代次
   其余读写方法自定（放在 secrets.go）。
2. internal/core/secret
   const RefKey = "$secret_ref"
   type Store struct{ ... }
   func New(st *state.Store) *Store
   func (s *Store) Resolve(ctx context.Context, pluginID string, snapshot []byte) ([]byte, error) // 把引用换成明文；引用不属于该插件或不存在 → PERMISSION_DENIED / NOT_FOUND
   // Extract 从候选 JSON 抽取 secret：schema 中标记的路径与 keptPaths 的并集。与 previous（上一快照）同一路径的明文相同则复用原引用，否则生成新引用（16 字节随机值的十六进制）。
   func (s *Store) Extract(ctx context.Context, pluginID string, schema, candidate, previous []byte, keptPaths []string) (snapshot []byte, paths []string, err error)
   func (s *Store) Redact(snapshot []byte) ([]byte, error) // 管理读取：每个引用替换为 {"$secret_set": true}
   func (s *Store) Collect(ctx context.Context, pluginID string, protected [][]byte) (removed int, err error) // 删除该插件中未被 protected 快照（Active、Pending、保留代次）引用的 secret
3. internal/core/host
   Expect 增加字段：Snapshot func(candidate []byte) (snapshot []byte, err error) // nil 时快照等于候选（没有 secret）
   Prepared 增加字段：ConfigSnapshot []byte
   PREPARED.candidate_config_digest = SHA256(ConfigSnapshot)
4. internal/core/supervisor
   Spec 增加字段：ConfigSchema []byte; Secrets *secret.Store; KeptSecretPaths []string

行为：
- 路径：secret 路径模式由从根到字段的 JSON 名用 "." 连接（属性名中的 "~" 与 "." 按 RFC 6901 的方式转义为 "~0" 与 "~1"），数组元素与 map 的任意键写作 "*"，例如 providers.*.api_key、mcp_servers.*.env.*.value。Extract 按模式遍历候选 JSON 的每个具体位置；复用引用时与 previous 中同一具体位置（同一数组下标或 map 键）的明文比较。
- Bootstrap：supervisor 在 Launch 前用 Resolve 把 Active（restart）或 Pending（新候选，由 T-09 提供）快照解析为明文，放进 candidate_config_json；明文 Bootstrap 只走认证 IPC。
- REGISTER 之后：host 调用 Expect.Snapshot（supervisor 以 Extract 实现，schema=Spec.ConfigSchema，keptPaths=Spec.KeptSecretPaths 加上一代次的 generation_secret_paths），得到快照与路径集合；PREPARED 摘要对快照计算。字段值为 null 或缺失时不生成引用；字段值已经是 {"$secret_ref": id} 时拒绝（插件不能回传引用）。
- 新候选：在提交 Active 之前，SetGenerationConfig 写入快照，并写入 generation_secret_paths。普通 restart：Extract 的结果必须与已提交的快照完全相同（明文未变时复用原引用），否则 FAILED_PRECONDITION，不提交新代次。
- 旧 secret 字段从 Schema 移除但仍作为未知键保留时，因为路径在 keptPaths 中而继续被抽取；只有候选中该键被删除，路径才不再保留。
- Redact 的结果用于管理 API；掩码不是协议值，表单提交时省略字段表示保持原值（T-09 与 T-08.7 遵守）。
- Collect 只删除不被任何 protected 快照引用的 secret；正在恢复或 Journal 中的引用由 T-09 加入 protected。
- 存储保护：状态库文件已是 0600；本切片在 Unix 上核对 data_dir 下状态库与其 WAL/SHM 文件的权限为 0600，在 Windows 上为状态库文件设置只允许当前账号访问的 ACL（golang.org/x/sys/windows，已批准）。
- 日志、错误与普通响应中不出现明文、Bootstrap 或完整配置；错误只写路径。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/secret_test.go（//go:build unix）。测试插件的配置含一个 secret 字段与一个普通字段，Spec.ConfigSchema 取生成器为它产生的 Schema：
  - 首次启动：插件收到明文；状态库代次快照中该字段为 {"$secret_ref": ...}，secrets 表有一行；PREPARED 摘要等于快照的 SHA256。
  - 明文不变的新候选复用同一引用；明文改变生成新引用，旧代次快照中的引用不变。
  - 新 Schema 去掉该 secret 字段但候选保留该键：新快照中它仍是引用；候选删除该键后不再保留路径。
  - 普通 restart 时插件回传了不同的明文：FAILED_PRECONDITION，Active 不变。
  - Collect 在保护 Active 时不删除其引用，删除只被已丢弃候选引用的 secret。
  - 日志缓冲中不出现明文字符串。
  - 数组中的 secret（providers.*.api_key，两个元素）各自生成引用；调换元素顺序后按位置比较，明文变化的位置生成新引用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：路径转义与 * 模式；Resolve 拒绝其他插件的引用；插件回传引用被拒；Redact 的输出；关键字常量与 sdk/config 一致；文件权限检查。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Pending 中 secret 的写入与表单（T-09、T-08.7）；明文查看的再认证与审计（T-09.9）；Journal 中引用的保护（T-09.5）；Adapter 实例配置的 Secret（T-09.4）。
```

### goal-34 · T-08.7 web/admin 骨架、配置表单组件与界面文案

状态：等待前置（T-08.1）。语句中 goal-28 定下的配置 Schema 格式（draft 2020-12 加 x-alonabot-secret、x-alonabot-hidden、x-alonabot-form、x-alonabot-i18n 关键字）只写名字；合并后由审查者用实际生成的 Schema 核对 fixture，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-34（切片 T-08.7：web/admin 骨架、配置表单组件与界面文案）。goal 目录 goal-34，分支 goal-34，worktree .worktrees/goal-34，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：建立 web/admin 前端工程（Vite、TypeScript、React 19、vitest），它是官方 WebUI 组件（`role: webui`）的源码，构建产物 dist/ 是组件的静态根目录；提供由配置 Schema 驱动的表单组件：字段按声明顺序展示，hidden 字段不显示也不被删除，secret 字段显示「已设置」状态而不把掩码当成保存值，提交生成 RFC 7396 Merge Patch；字段标签按 i18 回退规则显示；WebUI 自身文案有 en-US 与 zh-CN 目录和语言切换组件。真实管理页面与 API 在 T-09 接入。

依据：PLAN T-08 步骤 9（前端）；CONTRACTS §5.1 第 4 条、§5.3、§5.4 最后一条、§7.5、§9.1（WebUI 字段）；DECISIONS D-43、D-62、D-63；docs/development.md（Node 版本与已批准的前端依赖）；TEST-MATRIX V-25 的表单分项、V-54 的界面组件分项。

范围（只新增或修改这些）：
- 新增 web/admin/（package.json、package-lock.json、tsconfig.json、vite.config.ts、index.html、src/、fixtures/、i18n/，以及 manifest plugin.yaml：`role: webui`、`assets: dist`、Admin API 主版本范围 1～1）
- .gitignore：追加 web/admin/node_modules/ 与 web/admin/dist/
- docs/development.md 不改；Node 使用 asdf 管理的 26.5.1，包管理用 npm
允许的依赖：只用 react、react-dom、typescript、vite、@vitejs/plugin-react、vitest（均属已批准的 Vite、TypeScript、React 19、vitest）。需要 DOM 测试环境（jsdom、happy-dom）或 Testing Library 时标阻塞，等待维护者批准。

接口（模块与导出，照此实现）：
- src/schema/model.ts
  export type Field = { path: string[]; type: 'string'|'integer'|'number'|'boolean'|'object'|'array'|'map'; nullable: boolean; required: boolean; secret: boolean; hidden: boolean; widget?: string; options?: (string|number)[]; optionIcons?: Record<string, string>; min?: number; max?: number; defaultValue?: unknown; i18n: string; children?: Field[] }
  export function fieldsFromSchema(schema: unknown): Field[] // 按 properties 的声明顺序（Schema 的 JSON 键顺序）；hidden 字段保留在模型中，标记 hidden
- src/schema/patch.ts
  export type SecretState = { set: boolean } // 来自管理读取的 {"$secret_set": true}
  export function buildPatch(original: unknown, edited: unknown, fields: Field[]): Record<string, unknown> // RFC 7396：只包含改变的键；未触碰的 secret 与 hidden 字段不出现；清空可空字段生成 null 删除；数组整体替换
- src/i18n/catalog.ts
  export function pickLocale(available: string[], preferred: readonly string[]): string // 精确 → 主语言（字典序最小）→ en-US
  export function fieldLabel(messages: Record<string, unknown>, fallback: Record<string, unknown>, field: Field): string // config.<path>.label → en-US → 字段名
- src/components/ConfigForm.tsx：export function ConfigForm(props: { schema: unknown; value: unknown; secrets: Record<string, SecretState>; messages: Record<string, unknown>; onSubmit: (patch: Record<string, unknown>) => void })
- src/components/LanguageSwitch.tsx：export function LanguageSwitch(props: { value: string; onChange: (tag: string) => void })
- i18n/en-US.json、i18n/zh-CN.json：WebUI 文案（键集合一致）

行为：
- 表单：string 用文本框（widget=password 时为密码框，textarea 时为多行），integer/number 用数字框并按 min/max 限制，boolean 用开关，select 用下拉（options；有 optionIcons 时在选项旁用 `<img>` 显示图标，图标 URL 由调用方传入的函数生成，Admin API 的图标路径由 T-09.12 确定；不内联 SVG，不内置厂商图标），object 嵌套为分组，array 与 map 提供增删行。required 字段标记必填。表单只在提交时生成 patch，不自行补默认值（默认值由 SDK 在准备时补，T-08.3）。
- secret：显示「已设置/未设置」与「修改」按钮；未点「修改」时不出现在 patch 中；点了并输入后以新值出现；提供「清除」时（可空字段）生成 null。不把任何掩码字符串放进 patch。
- hidden：不渲染，但 patch 永远不包含它（不会因未显示而被删除）。
- 未知键：表单只渲染 Schema 中的字段；patch 只包含被编辑的已知字段，未知键不受影响。
- 字段文案：标签与说明按 pickLocale 选出的语言取 config.<path>.label / description，缺失时回落 en-US，再回落字段的 JSON 名；options 只翻译显示文案，提交值不变。
- 语言切换：初始语言按浏览器 navigator.languages 用 pickLocale 在 en-US 与 zh-CN 中选择；用户选择后保存在 localStorage，与机器人回复语言无关。
- fixtures/：至少三个 Schema fixture（嵌套与顺序、secret 与 hidden、可空与数组/map），由 goal-28 的生成器为对应 Go 类型实际生成后放入（审查者在派发前核对）。
- npm scripts：dev、build、test（vitest run）、typecheck（tsc --noEmit）。

测试（按编码规范「测试提交规则」）：
- 提交：web/admin/src/schema/schema.test.ts（vitest，不依赖 DOM）：用 fixtures 驱动 fieldsFromSchema 与 buildPatch，覆盖字段顺序、hidden 不进 patch、未修改的 secret 不进 patch、修改 secret 进 patch、清空可空字段生成 null、数组整体替换、未知键不受影响；pickLocale 与 fieldLabel 的回退。这是 Schema 格式与表单之间的契约测试。
- 本地自测（不提交，PR 中写明结果）：npm run build、npm run typecheck；在浏览器中手动打开 dev 页面，用 fixture 渲染表单并切换语言（写明浏览器与版本，属于手动验证）。

完成条件：在 web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 全部成功；仓库根目录 node scripts/check-repository.mjs 通过；go build ./... 不受影响。

不做：真实管理 API、登录与页面路由（T-09.1、T-09.8）；CI 中的前端任务（另行由维护者决定是否加入）；DOM 渲染测试（需要新依赖）。
```

### goal-53 · T-08.8 用户可见文案目录

状态：等待前置（T-04.1、T-04.3、T-08.1）。语句中 goal-11、goal-12、goal-28 提供的符号（sdk/generate 的 Scan 与诊断、sdk.Generated、Context、router 的 usage 生成）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-53（切片 T-08.8：用户可见文案目录）。goal 目录 goal-53，分支 goal-53，worktree .worktrees/goal-53，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件的 i18/<locale>.json 中 messages 的键与占位符在构建期检查，并由生成器嵌入；SDK 提供 Context.T 按「Actor locale → Core 默认 locale → en-US」选择语言、主语言回落；Core 内置 en-US 与 zh-CN 的用户文案目录，命令 usage 改为经它生成；Core 默认 locale 是一项持久化的全局设置。

依据：PLAN T-08 步骤 8；CONTRACTS §5.4、§5.3（config 根节点）、§2.2 第 8 条；MESSAGE-MODEL §1（Actor.locale）；TEST-MATRIX V-54 的 SDK 与 Core 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：Bootstrap 追加 string default_locale = 12。随后 go generate ./...（本切片获授权只追加这一字段）
- contracts/wirespec_test.go：追加编号检查
- 新增 contracts/i18n/（文件格式解析、占位符、语言选择，Core 与 SDK 共用）
- sdk/generate/：检查并嵌入 i18/*.json；sdk 的 Generated 增加 I18n
- sdk/：新增 i18n.go（Context.T 与后台 T）
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/settings.go
- 新增 internal/core/i18n/（Core 内置目录与 en-US.json、zh-CN.json）
- internal/core/router/：usage 改用 Core 目录，删除 goal-12 的 shortcut 注释
- internal/core/host/prepare.go：Bootstrap 写入 default_locale
- 上述包的测试文件
允许的新跨模块边：sdk/... -> contracts、internal/core/... -> contracts（已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. contracts/i18n
   type Catalog struct{ ... } // 一个组件全部 locale 的 messages
   type Problem struct { File string; Key string; Message string }
   func Parse(files map[string][]byte) (*Catalog, []Problem) // 键为文件名（如 "en-US.json"）；有问题时 Catalog 为 nil
   func (c *Catalog) Format(locales []string, key string, args map[string]string) (string, error) // locales 为按优先级的候选（如 [Actor locale, 默认 locale]），最后总会回落到 en-US
   func Placeholders(message string) ([]string, error)
2. sdk
   func (c *Context) T(key string, args ...any) (string, error) // args 为成对的名字与值，值用 fmt.Sprint 格式化
   func T(ctx context.Context, key string, args ...any) (string, error) // 没有 Actor：从 Core 默认 locale 开始
   Generated 增加 I18n map[string][]byte
3. internal/core/state
   迁移：CREATE TABLE core_settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
   func (s *Store) Setting(ctx context.Context, key string) (string, bool, error)
   func (s *Store) PutSetting(ctx context.Context, key, value string) error
4. internal/core/i18n
   func Messages() *i18n.Catalog // Core 内置目录（嵌入的 en-US.json 与 zh-CN.json）
   const SettingDefaultLocale = "default_locale"
   func DefaultLocale(ctx context.Context, st *state.Store) (string, error) // 未设置时 "en-US"
   func SetDefaultLocale(ctx context.Context, st *state.Store, tag string) error // 必须是规范的 BCP 47 标签

行为：
- 文件格式（Parse 校验，问题以 Problem 报告）：文件名必须是规范 BCP 47 标签加 .json；顶层只允许 config 与 messages 两个键；messages 的键为点分隔的 ASCII 标识（[a-z0-9_]+ 段），值为字符串；{name} 为占位符（name 同标识规则），{{ 与 }} 为字面花括号，其余单独的花括号是错误。必须有 en-US.json；其他语言不得有 en-US 没有的键；同一键在各语言的占位符集合必须与 en-US 相同；JSON 不合法是错误。config 根节点只检查是对象（字段翻译的路径检查属于表单，T-08.7）。
- 语言选择（Format）：按候选顺序，每一级先精确匹配，再按主语言子标签匹配（同一主语言有多个文件时取字典序最小者）；都不命中时用 en-US；目标语言缺少该键时回落到 en-US 的该键。未知键（en-US 也没有）返回错误；缺少占位符参数返回错误；多余参数忽略。不用键名充当文案。
- 生成器：构建入口目录下有 i18/ 时解析全部 .json，Problem 作为诊断报告（位置为文件名与键）；没有问题时把文件原文嵌入 Generated.I18n。plugin check 同样报告。
- SDK：Context.T 的候选为 [Actor.locale（存在时）, Bootstrap.default_locale]；后台 T 为 [default_locale]。插件没有 i18 资源时 T 返回 NOT_FOUND 的错误。
- Core：Bootstrap.default_locale 取 DefaultLocale；修改在之后建立的 Session 生效（已运行的插件在下次启动时看到新值）。命令 usage 的文案键为 core.usage.*（en-US 与 zh-CN 两份），按事件 Actor 的 locale 与默认 locale 选择语言；占位符为命令路径与参数用法。
- 日志、错误字符串与 safe_message 仍只用英文。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/i18n_test.go（//go:build unix）：测试插件带 en-US、zh-CN、zh-TW 三个文件，收到消息时用 Context.T 回复：Actor.locale 为 zh-CN 得到中文；zh-HK 回落到主语言 zh 中字典序最小的 zh-CN；ja 与没有 Actor 时用默认 locale；把默认 locale 设为 zh-CN 并重启插件后，ja 得到中文。参数错误的命令 usage 在 zh-CN 用户处为中文、en-US 用户处为英文。
  - sdk/generate/i18n_test.go：沿用 goal-11 的临时 module：缺少 en-US、zh-CN 多出的键、占位符集合不同、非法 JSON、多余顶层键各产生诊断；合法时 Generated.I18n 被嵌入。
  - contracts/wirespec_test.go：Bootstrap.default_locale = 12。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：占位符与花括号转义；Format 的未知键与缺少参数；BCP 47 规范化；core 目录的 en-US 与 zh-CN 键集合一致；SetDefaultLocale 拒绝非法标签。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：WebUI 文案与语言切换（T-08.7）；配置字段翻译的渲染（T-08.7）；确认消息的文案（T-11.2 使用本目录）；修改默认 locale 的管理 API（T-09.9）。
```

### goal-36 · T-09.1 首次管理认证

状态：等待前置（T-07.4）。语句中 goal-25、goal-26、goal-76 提供的符号（webhost.Host、webhost.AdminSessions、settings.Settings、cobra 命令树）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-36（切片 T-09.1：首次管理认证）。goal 目录 goal-36，分支 goal-36，worktree .worktrees/goal-36，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：alonabot admin-token init/reset 在持有 core.lock 时离线生成 32 字节管理 token，只在本机终端显示一次，状态库只存 SHA256 与认证代次；reset 递增代次并撤销全部管理会话、兑换票据与派生插件会话。admin host 提供 POST /api/auth/login（Origin 与登录 CSRF、恒时比较、8h 会话）与 POST /api/auth/logout；管理状态变更要求 CSRF；CLI 用显式 Bearer token。实现 webhost.AdminSessions，未初始化或认证无效时拒绝一切管理访问。

依据：PLAN T-09 步骤 1；CONTRACTS §7.3、§7.4、§15.1（listen 与 Base URL）、§14.5；TEST-MATRIX V-47 的认证分项。

范围（只新增或修改这些）：
- internal/core/state/state.go：migrations 末尾追加一条；新增 internal/core/state/admin.go
- 新增 internal/core/adminauth/
- internal/core/settings/：非 loopback 的 listen 必须配合 https 的 admin_base_url 或非空的 trusted_proxies（只加校验）
- cmd/alonabot/：新增 admin-token 命令组（init、reset）；run 用 adminauth 作为 webhost 的 AdminSessions 并挂上登录接口
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：
   CREATE TABLE admin_credential (id INTEGER PRIMARY KEY CHECK (id = 1), token_sha256 BLOB NOT NULL, auth_generation INTEGER NOT NULL, updated_at INTEGER NOT NULL);
   CREATE TABLE admin_sessions (session_id TEXT PRIMARY KEY, csrf_token TEXT NOT NULL, auth_generation INTEGER NOT NULL, expires_at INTEGER NOT NULL);
   读写方法自定（放在 admin.go）。
2. internal/core/adminauth
   var ErrNotInitialized = errors.New("admin token is not initialized")
   func InitToken(ctx context.Context, st *state.Store) (token string, err error)  // 已有凭证时返回错误；token 为 32 字节随机值的 base64url（无 padding）
   func ResetToken(ctx context.Context, st *state.Store) (token string, err error) // 递增 auth_generation，删除全部管理会话
   type Auth struct{ ... }
   func New(st *state.Store, now func() time.Time, secure bool) *Auth // secure 为 true 时 Cookie 用 __Host- 前缀与 Secure
   func (a *Auth) Routes(mux *http.ServeMux)        // POST /api/auth/login、POST /api/auth/logout
   func (a *Auth) Require(next http.Handler) http.Handler // 管理 API 的中间件：Cookie 会话（变更方法要求 CSRF）或 Authorization: Bearer
   func (a *Auth) Lookup(r *http.Request) (sessionID string, expiresAt time.Time, ok bool) // 实现 webhost.AdminSessions
   func (a *Auth) Valid(sessionID string) bool
   func (a *Auth) Reauthenticate(r *http.Request, token string) bool // 供明文查看 secret 前的再认证（T-09.9 使用）
3. cmd/alonabot
   alonabot admin-token init --data-dir DIR；alonabot admin-token reset --data-dir DIR
   两者都要求 stdin 与 stderr 都是终端，否则以 exitUsage 退出且不生成 token；先取得 core.lock（Core 正在运行时失败并提示先停止 Core）；token 只写到 stderr 一次，不进日志。

行为：
- 登录：只在 admin host 上提供（webhost 已按 Host 分派）。请求必须是 POST、Content-Type 为 application/json、Origin 等于 admin_base_url 的 origin（缺失或不等 → 403）；请求体 {"token": "..."}。未初始化 → 503（ErrNotInitialized 的 safe 文本）。SHA256 后用 subtle.ConstantTimeCompare 比较；失败 → 401（不区分原因）。成功时创建会话（随机 32 字节 ID，csrf_token 为另一随机值，到期为 8h 后，命名常量，绑定当前 auth_generation），设置 Cookie（__Host-alonabot_admin 或 alonabot_admin；HttpOnly、SameSite=Strict、Path=/），响应体 {"csrf_token": "..."}。
- Require：Cookie 会话必须存在、未过期且 auth_generation 等于当前值；GET/HEAD 之外的方法还要求请求头 X-Alonabot-CSRF 等于会话的 csrf_token 且 Origin 校验通过。或者 Authorization: Bearer <token> 与凭证摘要恒时比较相等（CLI 用，不需要 CSRF）。都不满足 → 401；未初始化 → 503。
- 登出：删除会话及其派生的插件会话（webhost 通过 Valid 立即感知），清除 Cookie。
- reset：auth_generation 加一并删除全部管理会话；派生票据与插件会话因 Valid 返回 false 而失效。
- 安全：token、会话 ID 与 csrf_token 不写日志、不出现在错误中；登录失败以 INFO 记录（来源地址，不含 token）。设置校验：listen 不是 loopback 地址时，admin_base_url 必须是 https，或 trusted_proxies 非空（TLS 由可信反代终止），否则 Load 返回错误。
- cmd/alonabot run：admin host 上把 /api/ 下的路由（目前只有认证路由）交给 Auth；webhost 使用 Auth 作为 AdminSessions。首次启动自动生成 token 的流程属于 serve（T-09.4）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - cmd/alonabot/e2e_test.go 追加：非终端运行 admin-token init 与 reset 时以 exitUsage 退出，状态库中没有凭证；Core 运行中执行 admin-token reset 时因 core.lock 失败。终端分支在本地自测中通过替换 cmd 内部的终端检测函数变量覆盖（不新增依赖）。
  - internal/core/supervisor/adminauth_test.go（//go:build unix）：用 webhost、httpproxy 与 httpplugin 装配：未初始化时登录 503；错误 token 401；正确 token 得到 Cookie 与 csrf_token；缺少 CSRF 的 POST 管理请求 401；Bearer 请求通过；Origin 不符的登录 403；登录后经票据进入私有插件页面，reset 后该插件会话与管理会话都失效；8h 后会话失效（注入时钟）；插件收到的请求中没有管理 Cookie 与 Authorization。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：替换终端检测后 init 只显示一次 token、reset 递增代次；恒时比较；Cookie 属性在 secure 与否时的差异；settings 的非 loopback 校验；日志中不出现 token；InitToken 在已有凭证时失败；Reauthenticate。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：管理 API 的业务接口与页面（T-09.4、T-09.8）；serve 的首次启动生成 token（T-09.4）；多用户账号（契约不提供）；明文查看 secret 的接口（T-09.9）。
```

### goal-69 · T-13.3 热部署

状态：等待前置（T-13.1）。语句中 goal-67 提供的符号（根包 Run 与选项、assembly、internal/inproc、开发组件的启动流程）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-69（切片 T-13.3：热部署）。goal 目录 goal-69，分支 goal-69，worktree .worktrees/goal-69，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Reload 选项或 ALONABOT_RELOAD=1 启用热部署：根类型所在包属于 main 包所在 module 的开发组件改为子进程运行；Run 把当前 main 包编译到 <data_dir>/dev/bin/，按正常启动流程启动，并在启动环境中指明要运行的组件；每 DEV_RELOAD_POLL 用标准库轮询 module 内 .go 文件与嵌入资源的修改时间，编译成功后依次重启这些组件，编译失败时旧组件继续运行；其他 module 的组件留在 Core 进程中不重启；go.mod 或 go.sum 变化时输出 WARN。

依据：PLAN T-13 步骤 4；CONTRACTS §15.4、§16（DEV_RELOAD_POLL）；TEST-MATRIX V-65。

范围（只新增或修改这些）：
- 根包：新增 reload.go；alonabot.go 增加 Reload 选项与子进程模式（只加分支）
- internal/inproc/：增加子进程模式下指明组件的环境变量名常量
- examples/devbot/：可选地演示 Reload（只改 main.go 的选项）
- 上述包的测试文件
允许的新跨模块边：无（编译使用 os/exec 调用 go build）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 根包 alonabot
   func Reload() Option // 等同 ALONABOT_RELOAD=1
2. internal/inproc
   const EnvComponent = "ALONABOT_DEV_COMPONENT" // 子进程模式：值为要运行的组件 ID；设置时 Run 只运行该组件，不装配 Core

行为：
- 判定：组件根类型（reflect）的包路径以 main 包所在 module 路径（runtime/debug.ReadBuildInfo 的 Main.Path）为前缀时为本 module 组件，否则为其他 module 组件。
- 编译：go build -o <data_dir>/dev/bin/<可执行文件名>（Windows 加 .exe），在 main 包目录执行，环境为当前环境；输出写到 Core 日志（INFO，失败时 ERROR 并附编译器输出）。
- 启动：本 module 组件以子进程运行该二进制，环境中设置 EnvComponent 与通常的 Session 变量（沿用 process 与 supervisor 的启动流程，安装记录仍为 dev 来源、代次 dev）；子进程中的 Run 看到 EnvComponent 时只以独立进程方式运行该组件（调用 sdk.Run 或 sdk.RunAdapter 的环境变量版本），不装配 Core。子进程异常退出按有限重启策略处理。其他 module 的组件仍按 goal-67 在 Core 进程的 goroutine 中运行。
- 轮询：每 DEV_RELOAD_POLL 遍历 module 根目录下的 .go 文件与 //go:embed 引用的文件（排除 <data_dir>、vendor、以 . 或 _ 开头的目录、testdata），比较修改时间与大小；有变化时编译，成功后依次停止并重新启动每个本 module 组件（每次一个，停止使用正常停止流程与预算）；Adapter 组件重启后由 remote.Process 恢复已启用的实例。编译失败时不停止旧组件。go.mod 或 go.sum 变化时只输出 WARN "go.mod or go.sum changed; restart alonabot to apply"，不编译。
- 停止 Run 时停止轮询，并按 goal-67 的顺序停止全部组件；dev/bin 下的二进制保留。

测试（按编码规范「测试提交规则」）：
- 提交：reload_test.go（根包，//go:build unix）：在 t.TempDir() 中写一个临时 module（go.mod replace 指向本仓库），含一个开发插件与一个来自本仓库 module 的回环 Adapter（属于「其他 module」）；以 ALONABOT_RELOAD=1 运行它的 main：
  - 启动后插件以子进程运行（日志中有 pid，与 Core 进程不同），回环 Adapter 在 Core 进程中运行。
  - 修改插件回复的文本并保存：在若干个轮询周期内插件重启，之后的回复为新文本；回环 Adapter 的实例没有 bot.offline。
  - 写入语法错误：日志有编译错误，旧插件继续回复旧文本；修复后恢复。
  - 修改 go.mod：日志有 WARN，没有重新编译。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：本 module 判定；轮询排除目录；Windows 可执行文件名（交叉检查，不在 Windows 运行）；子进程模式下 Run 只运行指定组件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./... ；go build ./...。

不做：配置与 dev.sqlite（T-13.2）；文件系统事件通知（只用标准库轮询）；Windows 上的实际运行验证（T-10）。
```

## 第 8 波

### goal-9 · T-03.10 Telegram 接入开发入口与故障矩阵

状态：等待前置（T-03.6、T-03.8、T-03.9）。语句中 goal-4～goal-8、goal-50、goal-51 提供的符号（remote.Process、supervisor.Spec 的 Role、adapters/telegram、telegramtest、media.Store、buttons.Service、message 的 Send/Edit）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-9（切片 T-03.10：Telegram 接入开发入口与故障矩阵）。goal 目录 goal-9，分支 goal-9，worktree .worktrees/goal-9，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：alonabot run 可以同时启动一个 Adapter 包与一个插件，Adapter 的实例配置取自 <data_dir>/adapters/<adapter_id>/instances/<instance_id>.json；提供 echo 与媒体示例；用本地 Bot API 模拟服务跑完 PLAN T-03 步骤 8 的发送与接收故障矩阵，并在 TEST-MATRIX §3.2 记录 T-03 的证据。真实 Telegram 没有测试账号时记为未运行。

依据：PLAN T-03 步骤 8 与「停止和失败条件」；CONTRACTS §8、§15.4（实例目录）；MESSAGE-MODEL §4、§6；TEST-MATRIX V-07、V-08、V-09、V-35、V-44、V-45、V-48、V-51～V-54 的 T-03 分项。

范围（只新增或修改这些）：
- cmd/alonabot/：run 增加 --adapter-id、--adapter-entry、--adapter-dir（可选；给出时以 Adapter 角色启动该包，读取实例目录并启动全部实例，再启动插件）
- 新增 examples/plugins/echo/ 的媒体回送能力（收到图片时以 incoming_ref 原样回送）；新增 examples/plugins/media/（上传一个内置图片并发送相册与按钮）
- adapters/telegram/telegramtest/：补齐矩阵所需的故障注入（getUpdates 409、提交后超时、相册后 markup 失败等）
- 新增 cmd/alonabot/telegram_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-03 的证据行（本切片获授权修改这一处；其他 spec 文件不改）
- 上述包的测试文件
允许的新跨模块边：cmd/alonabot 的测试 -> adapters/telegram/telegramtest（测试代码，审查者确认允许）。

接口：
- alonabot run 的新参数：--adapter-id ID --adapter-entry PATH [--adapter-dir DIR]。Adapter 的安装记录为 dev-local 来源、Role=adapter、AdapterID=ID；实例目录为 <data_dir>/adapters/<ID>/instances/，每个 *.json 是一个实例（文件名去掉 .json 为 instance_id），内容原样作为实例配置；目录不存在时没有实例并输出 WARN。
- 不新增公开 Go 符号。

行为：
- 启动顺序：取得 core.lock → 启动 Adapter 进程并 Bind → 启动全部实例 → 启动插件。插件停止与 Core 退出时先停插件、再停 Adapter。Adapter 进程异常退出按有限重启策略恢复，Bind 后恢复已启用的实例。
- 实例文件中的 Secret 字段（token、proxy）在本切片按原样下发；它们只出现在 adapter.instance.start 请求中，不写日志。Secret Store 对实例配置的处理由 T-09.4 接入，用一行 shortcut 注释写明。
- 故障矩阵（全部以 telegramtest 运行，在 telegram_e2e_test.go 中各为一个子测试，并以 telegramtest.Calls() 的实际请求数量为证据）：
  1) 同包两个实例单独启停；停一个不影响另一个，进程 PID 不变。
  2) Adapter 进程被杀后恢复：两条 ADAPTER_EXITED 的 bot.offline，恢复后各一条 bot.online，从已保存游标继续，不重复交付。
  3) 伪造其他实例的事件被拒（使用回环测试 Adapter 的方式无法覆盖时，以 Telegram 实例在 deliver 中填错 account_id 的测试钩子验证）。
  4) 话题（thread）消息的回复带 message_thread_id。
  5) 一条 Chain 生成多条物理消息与两次 Send 的区别：请求数分别为 N 与 2。
  6) 提交后超时：第一个物理请求断开，结果 UNKNOWN，此后没有任何补发请求。
  7) 相册后 markup 失败：PARTIAL，媒体组 ID 保留，没有重发相册。
  8) message_id=0：PARTIAL，依赖它的 markup 为 NOT_ATTEMPTED。
  9) 入站图片：两个插件共享一次 getFile 下载；授权与过期按 goal-7；超出配额 RESOURCE_EXHAUSTED。
  10) 游标提交后崩溃：交付后杀掉 Core 进程，重启后不重复交付该 update。
  11) getUpdates 409：实例停止并产生 INSTANCE_FAILED 的 bot.offline。
  12) 按钮点击回到发送插件；编辑消息成功，不改为发送新消息。
  13) Actor.locale、to_me 的映射（各一条）。
- 证据：TEST-MATRIX §3.2 追加 T-03 一行，写明测试位置、命令、平台（darwin/arm64 等实际运行平台）与结果，模拟服务的结果标为「本地 Bot API 模拟」；真实 Telegram 标「未运行」并写明原因，不用模拟结果代替。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/telegram_e2e_test.go（上述矩阵）；构建 cmd/alonabot、adapters/telegram/cmd/alonabot-adapter-telegram、examples/plugins/echo 与 media 后以真实进程运行。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：实例目录的各种异常（非法 JSON、空目录、非 .json 文件被忽略）；--adapter-* 参数的组合错误。
- 若维护者提供了测试账号：在 PR 中另附一次真实 Telegram 的手动运行记录（命令、时间、结果），不写入提交的测试。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：正式安装与实例管理（T-09.4）；实例配置的 Secret Store（T-09.4）；其他平台的 Adapter。
```

### goal-17 · T-05.3 三阶段 Hook 与 send_forbidden

状态：等待前置（T-05.2）。语句中 goal-2、goal-12、goal-15、goal-16、goal-51 提供的符号（router 的派发链与可信信息、message.Service 的 Send 与 Edit、sdk 的分发与 RouteOption）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-17（切片 T-05.3：三阶段 Hook 与 send_forbidden）。goal 目录 goal-17，分支 goal-17，worktree .worktrees/goal-17，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：运行 PreEvent、BeforeSend、AfterSend 三类 Hook：PreEvent 只能改写业务 payload 或 Stop，失败时结束本次入站链；每次显式或自动 Send（以及 Edit）独立执行一组 BeforeSend → Adapter.Send → AfterSend，BeforeSend 只能改 Chain 或拒绝，改后重新预检；AfterSend 只观察。出站 Hook 的执行上下文带 send_forbidden，该标记保存在 Core 的父调用记录中并随派生调用继承，Hook 内直接 Send 被拒。

依据：PLAN T-05 步骤 5、6；CONTRACTS §2.3 第 3～5 条、§8（Edit 的 Hook 段）；WIRE-FORMAT §3.5（Hook 输入的追加）；TEST-MATRIX V-15 的直接路径分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/invocation.proto：BeforeSendRequest 与 AfterSendRequest 各追加 SendOperation operation 与 optional string edited_platform_message_id，并新增 enum SendOperation。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- contracts/validate/register.go：HOOK 路由的注册校验
- 新增 internal/core/hook/
- internal/core/router/：派发链在命令层之前执行 PreEvent（只加调用点）
- internal/core/message/：Send 与 Edit 在调用 Adapter 前后执行出站 Hook；删除 goal-51 的 shortcut 注释
- internal/core/router/：调用记录（Invocation）增加 send_forbidden，并在 InvocationContext 中下发
- sdk/：新增 hooks.go（注册与分发）；base.go 的 Root 增加 Hooks Hook 与 Base 默认实现；session.go 的 buildRegister 输出 HOOK 路由
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   enum SendOperation { SEND_OPERATION_UNSPECIFIED = 0; SEND_OPERATION_SEND = 1; SEND_OPERATION_EDIT = 2; }
   BeforeSendRequest 追加 SendOperation operation = 5; optional string edited_platform_message_id = 6;
   AfterSendRequest 追加 SendOperation operation = 7; optional string edited_platform_message_id = 8;
2. sdk
   Root 增加 Hook：Hooks(*HookRegistry) // func (*Base) Hooks(*HookRegistry) 默认不注册
   type HookRegistry struct{ ... }
   func (r *HookRegistry) PreEvent(routeID string, eventTypes []string, h func(*Context, *pb.EventPayload) (*pb.EventPayload, error), opts ...RouteOption)
   func (r *HookRegistry) BeforeSend(routeID string, h func(*Context, *pb.Target, *pb.MessageChain) (*pb.MessageChain, error), opts ...RouteOption)
   func (r *HookRegistry) AfterSend(routeID string, h func(*Context, *pb.Target, *pb.SendResult, error), opts ...RouteOption)
   func (c *Context) Operation() pb.SendOperation // 出站 Hook 中可用
3. internal/core/hook
   type Runner struct{ ... }
   func New(r *router.Router, log *slog.Logger) *Runner
   func (h *Runner) PreEvent(ctx context.Context, ev *router.Dispatch) (stop bool, err error) // router 的派发对象；只改写其业务副本
   func (h *Runner) BeforeSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation) (*pb.MessageChain, *pb.ErrorInfo)
   func (h *Runner) AfterSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, result *pb.SendResult, errInfo *pb.ErrorInfo, parent *router.Invocation)

行为：
- 注册校验：HOOK 路由的 phase 必须是三者之一；PreEvent 的 event_type_filter 非空且为已知类型；BeforeSend/AfterSend 不带 event_type_filter。
- 候选：同一阶段的 Hook 按 (priority, plugin_id, route_id) 稳定排序，只包含已激活插件；Hook 自己不受命令策略约束（插件启用即生效）。
- PreEvent（Router 在可信信息保留之后、D-44 的命令层之前，按事件类型）：依次调用，每个 Hook 收到当前业务副本；返回的 event_payload 的分支必须与 event_type 一致，否则视为该 Hook 失败；可信字段（goal-15 列出的身份、Target、来源与按钮/任务绑定）不在 PreEventResult 中，无法改写。任一 Hook 返回错误或 panic：以 WARN 记录并结束本次入站链（不派发命令与事件路由）。propagation=STOP：结束本次入站链，此前改写保留。
- 出站：message.Service 对每次 Send 与 Edit（包括 Core 代发的自动回复与 usage）：
  1) 先完成 goal-2/goal-3/goal-51 的目标与准入、Chain 预检；
  2) 依次调用 BeforeSend：返回新 Chain 时替换并对整条 Chain 重新做完整预检（失败视为拒绝）；返回错误即拒绝本次发送，返回 REJECTED 的 SendResult 加该 ErrorInfo，不调用 Adapter，也不执行 AfterSend；Hook 不能修改 Target（请求中没有可回写的 Target）。
  3) 调用 Adapter；4) 依次调用 AfterSend，传入最终的 SendResult 与 ErrorInfo，错误只记日志，不改变返回给调用方的结果。
  每次 Send 是独立的一组，不合并。
- send_forbidden：Core 调用出站 Hook 时，在该次调用的 Invocation 中设置 send_forbidden=true，并经 InvocationContext.send_forbidden 下发；Hook 内（及以该调用为父请求的任何子请求，Broker 转发时沿用父调用的标记）调用 core.send 或 core.edit 返回 FAILED_PRECONDITION。标记只来自 Core 的调用记录，插件自报的值被忽略。
- SDK：按 route_id 分发 Hook；PreEvent 的 h 返回 nil payload 表示不改写；Context.StopPropagation 在 PreEvent 中使 propagation=STOP；出站 Hook 的 Context 中 Send/Edit 由 Core 拒绝。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/hook_test.go（//go:build unix）。两个测试插件：h 注册 PreEvent（把 text 改成大写）、BeforeSend（在 Chain 末尾追加一段文本；遇到含 "deny" 的 Chain 返回错误；尝试在 Hook 内 Send）、AfterSend（记录结果后返回错误）；e 注册 echo 事件路由与一个命令：
  - 入站 "hi" 被 PreEvent 改为 "HI" 后到达 e；e 回复时 memory 收到追加了文本的 Chain；AfterSend 收到 ACCEPTED 的结果，它返回的错误不影响 e 收到的 SendResult。
  - 一次 Handler 内两次 Send：BeforeSend 与 AfterSend 各执行两次。
  - 含 "deny" 的回复被拒，memory 没有收到，e 收到 REJECTED 与错误；AfterSend 未执行。
  - BeforeSend 内的 Send 返回 FAILED_PRECONDITION。
  - PreEvent 返回错误时 e 的路由与命令都没有被调用。
  - Edit 时出站 Hook 收到 operation=EDIT 与被编辑的消息 ID。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：PreEvent 返回不匹配分支的 payload；BeforeSend 修改后 Chain 预检失败；排序的稳定性；插件自报 send_forbidden 被忽略；InvocationContext 中的标记在派生请求中保留（接口 fixture）。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：Hook→Service→Send 的间接递归回归（T-06.2 的必交用例，本切片只保证标记随父调用记录保存）；延续调用（T-05.4）。
```

### goal-52 · T-05.4 延续调用与多轮等待

状态：等待前置（T-05.2）。语句中 goal-2、goal-15、goal-16、goal-76 提供的符号（router 的派发链与 Invocation、sdk 的 Context 与分发、tune.Values）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-52（切片 T-05.4：延续调用与多轮等待）。goal 目录 goal-52，分支 goal-52，worktree .worktrees/goal-52，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：由真实入站事件触发、具有 Actor 与默认 Target 的 Handler 可以经 core.continuation.start 请求一次延续调用：原调用成功结束后，Core 以同一插件、同一 route_id 与原可信上下文发出新的 Handler 请求（continuation_of 为原 call_id），使用独立的 CONTINUATION_TIMEOUT 期限，不占用分发链。延续调用内可以经 core.wait 等待同一 Actor 在同一 Target 的下一条消息，命中的消息不进入普通分发。SDK 提供 Context.Continue 与 Context.Wait。

依据：PLAN T-05 步骤 7；CONTRACTS §2.4；WIRE-FORMAT §3.5；DESIGN（延续调用示例）；TEST-MATRIX V-50。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/invocation.proto：HandlerRequest 追加 optional uint64 continuation_of = 5；新增 ContinuationStartRequest/Response 与 WaitRequest/Response。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceContinuationStart = "core.continuation.start"、ServiceWait = "core.wait"
- 新增 internal/core/continuation/
- internal/core/router/：派发链在 PreEvent 之后、命令层之前检查等待（只加调用点）；Handler 调用结束时通知 continuation（只加调用点）
- sdk/：新增 continuation.go（Continue 与 Wait）；dispatch.go 按 continuation_of 找到登记的延续函数
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message ContinuationStartRequest {}
   message ContinuationStartResponse {}
   message WaitRequest { uint32 timeout_ms = 1; } // 1～CONTINUATION_TIMEOUT
   message WaitResponse { Event event = 1; }
2. internal/core/continuation
   type Manager struct{ ... }
   type Options struct { Max int; Timeout time.Duration } // CONTINUATION_MAX、CONTINUATION_TIMEOUT
   func New(r *router.Router, log *slog.Logger, opt Options) *Manager
   func (m *Manager) Register(mux *router.Mux)
   func (m *Manager) Finished(inv *router.Invocation, callID uint64, ok bool) // Router 在 Handler 调用结束后调用；ok 为原调用成功（没有 Err）
   func (m *Manager) Match(ev *router.Dispatch) bool // Router 在命令层之前调用；命中等待时返回 true，事件不再进入普通分发
   func (m *Manager) Release(pluginID string) // Session 失效：结束该插件的全部等待与待发出的延续调用
3. sdk
   func (c *Context) Continue(fn func(*Context) error) error // 只能在 Handler 中调用一次；fn 在延续调用中执行
   func (c *Context) Wait(timeout time.Duration) (*Context, error) // 只能在延续调用中调用；返回命中消息的 Context（可用 Event()、Text() 等）
   func (c *Context) Text() string // message.created 的文本片段连接结果；其他事件为空

行为：
- core.continuation.start（必须带 parent_call_id，指向本 Session 中仍在途的 Core→插件 Handler 调用）：以下情况返回 FAILED_PRECONDITION：父调用不是由真实入站事件触发（例如 scheduled、Bus、Service）、缺少 Actor 或默认 Target、父调用带 send_forbidden（出站 Hook 内）、父调用本身就是延续调用、同一父调用已请求过。本 Session 进行中的延续调用（已登记待发出与正在运行的）达到 Max 时返回 RESOURCE_EXHAUSTED。成功时只登记，返回空结果。
- 发出：Finished(ok=true) 时为登记的父调用发出新的 Handler REQUEST：同一插件、同一 route_id，HandlerRequest.continuation_of=原 call_id，event 为原事件（Core 保留的可信信息重建），context 为原调用的 Actor、默认 Target、scope、trace 与插件链；期限为发出时刻起 Timeout。它不经事件队列与派发链，不计入 DISPATCH_CONCURRENCY；返回值不参与原事件的传播。Finished(ok=false) 时丢弃登记，不发出。
- core.wait（必须带 parent_call_id，父调用必须是延续调用）：等待键为（adapter_instance_id、chat_id、thread_id（无则为空）、external_user_id），取自延续调用的默认 Target 与 Actor。同键已有等待 → ALREADY_EXISTS。timeout_ms 不在范围内 → INVALID_ARGUMENT。等待在 Core 内存中，不持久化。
- 匹配：Router 对每个 message.created，在 PreEvent 之后、命令层之前调用 Match：Target 与 Actor 构成的键命中等待时，该事件作为 WaitResponse 返回给等待者，并且不再进入命令层与事件路由。不同 Actor 或不同 thread 不命中。
- 结束：超时 → DEADLINE_EXCEEDED；父请求被取消 → CANCELLED；Release（插件停用或 Session 失效）→ 等待以 UNAVAILABLE 结束，待发出的延续调用丢弃，不补发。延续调用的期限到达时，其中的等待一并结束。
- SDK：Continue 在 Handler 返回前请求 core.continuation.start，并登记 fn（按原 call_id）；收到 continuation_of 的请求时调用对应 fn。Continue 在延续调用中调用返回 FAILED_PRECONDITION。Wait 发起 core.wait，以本次延续调用为父请求。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/continuation_test.go（//go:build unix）。测试插件注册 /quiz 命令：Continue 中依次 Wait 两次并把收到的回答回复出来；另有一个 echo 事件路由用来证明消息是否进入普通分发：
  - 用户 A 发 /quiz 后，A 的下两条消息被 quiz 收到并回答，它们没有到达 echo；原命令的 HandlerOutcome 决定的传播不受影响。
  - 同一群里 A 与 B 同时 /quiz，各自只收到自己的回答；不同 thread 的消息不命中。
  - Wait 超时返回 DEADLINE_EXCEEDED；同键第二个 Wait 返回 ALREADY_EXISTS。
  - 延续调用中再次 Continue 返回 FAILED_PRECONDITION；没有 Actor 的 scheduled 事件 fixture 中请求 Continue 返回 FAILED_PRECONDITION。
  - 用很小的 Max：超过的 Continue 返回 RESOURCE_EXHAUSTED。
  - 等待期间停止插件：等待结束，重启后不补发。
  - 延续调用执行期间，其他事件照常派发（不占分发链）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：原调用返回错误时不发出；父调用带 send_forbidden 时被拒；timeout_ms 边界；Release 与超时的并发；SDK 中按 continuation_of 找不到 fn 时返回 NOT_FOUND。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：持久化的等待与延续调用（契约不提供）；Agent 的流式对话（T-12.2 使用本切片）。
```

### goal-27 · T-07.5 调度与 HTTP 启停集成及验收

状态：等待前置（T-07.2、T-07.4）。语句中 goal-24～goal-26 提供的符号（scheduler.Scheduler、httpproxy.Server、webhost.Host、host 的停止流程与 StopTiming、supervisor.Running）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-27（切片 T-07.5：调度与 HTTP 启停集成及验收）。goal 目录 goal-27，分支 goal-27，worktree .worktrees/goal-27，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：把 Scheduler 与 HTTP 代理接入 T-02 的生命周期：插件 Activated 之后才 Attach/Bind；Draining 时停止新 tick 与新 HTTP 派发；停止阶段取消余留调用、撤销映射、关闭 Listener，全部复用全局 STOP_TOTAL 预算；Session 或代次变化后旧端点、凭证与运行记录不再接收新业务。完成 T-07 的时间、进程与 HTTP 边界验收，在 TEST-MATRIX §3.2 记录证据。

依据：PLAN T-07 步骤 7、8 与「停止和失败条件」；CONTRACTS §4.2、§4.3、§7.2；TEST-MATRIX V-20、V-21、V-22、V-46、V-47 的 T-07 分项。

范围（只新增或修改这些）：
- internal/core/supervisor/：Launch 的激活之后与停止流程中加入 Scheduler 与 httpproxy 的接入点（Spec 增加两个可选依赖）
- internal/core/scheduler/：增加 Drain（停止新 tick，保留运行槽）
- internal/core/httpproxy/：增加 Drain（新请求返回 503，等待在途请求到截止时间）
- cmd/alonabot/：run 装配 Scheduler，并把两者交给 supervisor
- 新增 cmd/alonabot/t07_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-07 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- internal/core/supervisor
  Spec 增加字段：Scheduler *scheduler.Scheduler; HTTP *httpproxy.Server // 都可以为 nil
- internal/core/scheduler
  func (s *Scheduler) Drain(pluginID string) // 不再触发新的 tick；运行中的调用照常等待
- internal/core/httpproxy
  func (s *Server) Drain(pluginID string, deadline time.Time) // 新请求 503；在途请求最多等到 deadline

行为：
- 激活：supervisor 在 host.Activate 成功（ACTIVATED 已应用）之后才调用 Scheduler.Attach（anchor 为激活时刻）与 HTTP.Bind；激活失败不 Attach、不 Bind。
- 停止（复用 host 的停止流程与同一个 STOP_TOTAL）：冻结时同时 Scheduler.Drain 与 HTTP.Drain（截止取 STOP_DRAIN 的结束时刻）；STOPPING 之后 Scheduler.Detach、HTTP.Unbind；进程树清空后 Scheduler.Released。不为 Scheduler 或 HTTP 另起独立预算。多个插件并行停止时共享 Core 的全局停止截止。
- 崩溃与重启：Session 断开时 HTTP.Unbind（以旧 generation）、Scheduler.Detach；旧进程确认退出后 Released；重启成功后以新 Session 重新 Attach/Bind，旧 token 与端点不能再接请求。
- HTTP 失败不触发任何业务重试；未激活的插件返回 503。
- 验收矩阵（t07_e2e_test.go，以真实进程运行，各为一个子测试）：
  1) cron 与 interval 的触发时刻（用 goal-23 的规范期望值与短周期 interval 实测），夏令时规则以 contracts/schedule 的结果为准。
  2) 不合作 callback：job 超过 JOB_TIMEOUT 后槽不释放、迟到响应释放；断链后槽保持到进程退出。
  3) 停用插件时：新 tick 不再触发，HTTP 返回 503，Listener 端口已关闭；总耗时不超过 STOP_TOTAL。
  4) 未激活、停用后、伪造内部头、私有路由无会话、公开回调、两个 Host、单 Host 主动内容，各返回契约规定的结果。
  5) 两个插件并行停止共享同一 STOP_TOTAL。
- 证据：TEST-MATRIX §3.2 追加 T-07 一行，写明测试位置、命令、平台与结果；仍需 T-09 首次认证与真实管理页面才能完成的端到端项写为「未闭合」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t07_e2e_test.go（上述矩阵）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Drain 的截止时间；激活失败不 Bind；重启后旧 generation 的 Unbind 不影响新映射；Scheduler 在 Detach 后不触发。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：管理认证（T-09.1，测试注入假会话）；serve（T-09.4）；Agent 定时任务（T-12.6）。
```

### goal-32 · T-08.5 受管 SQLite 连接与迁移

状态：等待前置（T-08.3）。语句中 goal-30 提供的符号（sdk/config.Prepare 与 sdk 的准备流程）只写名字；sdk/gormdb 的现有函数（OpenSQLite、ReadPragmas、CheckMigrators、RunVersioned、AppliedMigrations）已在 main 上。合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-32（切片 T-08.5：受管 SQLite 连接与迁移）。goal 目录 goal-32，分支 goal-32，worktree .worktrees/goal-32，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：受管 SQLite 有两个入口：新建的 sdk/sqlitedb 的 Open(ctx) 按当前 Session 的数据库描述返回 *sql.DB；sdk/gormdb 的 Open(ctx) 在同一个连接上返回原生 *gorm.DB。同一 Session 复用同一连接，SDK 停止时关闭。只导入 sqlitedb 的插件不链接 GORM。两个包各定义一组迁移接口，被导入时向 SDK 登记识别方式；SDK 在配置准备之后、OnStart 之前执行迁移，并在 REGISTER 中报告结果。Core 为每个代次下发 SQLite 路径。

依据：PLAN T-08 步骤 5、6；CONTRACTS §6.1～§6.3、§4.2（迁移预算）；WIRE-FORMAT §3.2（database_descriptor）；DECISIONS D-48、D-49；TEST-MATRIX V-26、V-39。

范围（只新增或修改这些）：
- sdk/sqlitedb/：新建包。连接设置与回读核对从 sdk/gormdb/sqlite.go 移入；open.go（Open 与 Session 绑定）、migrate.go（迁移接口、RunVersioned、AppliedMigrations）
- sdk/gormdb/：OpenSQLite 改为 sqlitedb.OpenFile 加 GORM 包装；新增 open.go（Open）；ReadPragmas 与 PragmaValues 移到 sqlitedb 后从本包删除
- sdk/internal/dbhook/：新建内部包，供 sqlitedb 与 gormdb 向 sdk 登记迁移识别与执行，形状由执行者自定，只用标准库与 contracts 类型
- sdk/：session.go 在配置准备之后调用迁移阶段并填写 REGISTER 的 migration；activate.go 在停止时关闭连接（只加调用点）
- internal/core/supervisor/：Spec 增加 Database；新代次时创建 SQLite 代次目录并下发路径
- internal/core/host/prepare.go：校验 REGISTER 的 migration 与 Bootstrap 描述一致（只加校验）
- 上述包的测试文件
允许的新跨模块边：sdk -> sdk/sqlitedb、sdk/gormdb -> sdk/sqlitedb（同属 sdk 模块）。sdk 根包不得 import sdk/gormdb 或 gorm。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/sqlitedb
   var ErrNoDatabase = errors.New("this plugin has no database")
   func Open(ctx context.Context) (*sql.DB, error) // 未处于 Session 中返回 FAILED_PRECONDITION 的错误；backend 为 NONE 返回 ErrNoDatabase；同一 Session 重复调用返回同一个 *sql.DB
   func OpenFile(ctx context.Context, path string) (*sql.DB, error) // 原 gormdb.OpenSQLite 的设置与核对：WAL、FULL、外键、5s busy timeout、单连接
   type PragmaValues struct { JournalMode string; Synchronous int; ForeignKeys int; BusyTimeout int }
   func ReadPragmas(ctx context.Context, db *sql.DB) (PragmaValues, error)
   type SelfMigrator interface { Migrate(ctx context.Context, db *sql.DB) error }
   type Migration struct { ID uint32; Up func(context.Context, *sql.Tx) error }
   type VersionedMigrator interface { Migrations() []Migration }
   const LedgerTable = "_alonabot_migrations"
   func RunVersioned(ctx context.Context, db *sql.DB, migrations []Migration) ([]uint32, error)
   func AppliedMigrations(ctx context.Context, db *sql.DB) ([]uint32, error)
2. sdk/gormdb
   var ErrNoDatabase = sqlitedb.ErrNoDatabase
   func Open(ctx context.Context) (*gorm.DB, error) // 经 libtnb sqlite.Config{Conn: ...} 包装 sqlitedb.Open 返回的同一个 *sql.DB；同一 Session 重复调用返回同一个 *gorm.DB
   OpenSQLite、SelfMigrator、Migration、VersionedMigrator、RunVersioned、AppliedMigrations、CheckMigrators 保留现有签名；LedgerTable 改为引用 sqlitedb.LedgerTable
3. internal/core/supervisor
   Spec 增加字段：Database *pb.DatabaseDescriptor // nil 时 supervisor 为 SQLite 生成描述：<plugins>/<id>/generations/<generation>/db.sqlite，activation_reason 由调用方在 Bootstrap 中给出

行为：
- Open：文件不存在时只有 activation_reason 为 first_install 时才创建，否则返回错误（不创建空库掩盖丢失）。连接错误原样返回。插件拿不到 Core 状态库的任何路径或连接。
- 登记：sqlitedb 与 gormdb 在包 init 中向 sdk/internal/dbhook 登记一个识别函数（只追加函数值，不做 I/O）；SDK 的迁移阶段只调用已登记的识别函数，因此未导入某个包时不认识它的接口。
- 迁移阶段（SDK，配置准备之后、OnStart 之前，预算为 Bootstrap 的 migration 阶段预算，与普通业务期限无关）：
  - 根插件实现的迁移接口合计多于一个（含两个包混用，例如 sqlitedb 的 Migrate 加 gormdb 的 Migrations）时失败（INVALID_ARGUMENT）。
  - backend 为 NONE：跳过，migration.mode=NONE。
  - versioned：ID 必须非零、唯一、严格递增；每步 Up 与 _alonabot_migrations 的完成记录在同一事务中，失败不记成功。两个包的账本格式相同。
  - self：调用 Migrate，成功后 self_migrate_completed=true。
  - REGISTER.migration：mode 与 completed_migration_ids（versioned，升序）或 self_migrate_completed（self）如实填写。
- prepare_effects=external 的维护门槛由 Core 在建立候选前判断（T-09），本切片的 SDK 不读取 manifest 的 prepare_effects。
- Core：supervisor 新代次时创建 generations/<g>/ 目录（0700）并下发 SQLite 路径；普通 restart 沿用 Active 代次的路径。host 校验 REGISTER.migration.mode 与 Bootstrap 的 migration_mode 一致、versioned 的 completed 包含 Bootstrap 的 applied 集合，否则注册失败。
- 事务内不得等待同一单连接池的另一条独立查询（文档注释写明，测试覆盖一次死锁会被 busy timeout 以外的方式暴露为错误）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/db_test.go（//go:build unix）：两个测试插件，一个只导入 sqlitedb、一个用 gormdb，各实现 Migrations（两步，第二步按环境变量可失败）并在 OnStart 中写一行：首装成功，REGISTER 报告 [1,2]，状态库代次目录下有 db.sqlite 且 PRAGMA 值正确；第二步失败时注册失败、账本只有 1；普通 restart 不重复执行迁移；同时实现两个迁移接口的插件在 CONFIG 之后失败。
  - 同一测试中断言只导入 sqlitedb 的测试插件二进制不含 gorm.io/gorm（`go version -m` 或 `go list -deps`）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：两个入口的复用、共用连接与停止时关闭；backend NONE；文件缺失时非首装不创建；ID 非递增；host 的 migration 一致性校验。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./sdk/... ./internal/core/...；go build ./...；node scripts/check-go-deps.mjs 通过。

不做：SQLite 候选副本与 Backup API（T-08.6）；维护许可的管理入口、external 插件的候选门槛与 Journal（T-09）。
```

### goal-39 · T-09.4 管理服务、安装记录与 serve

状态：等待前置（T-03.8、T-08.4、T-09.1、T-09.2、T-13.1）。语句中 goal-8、goal-31、goal-36、goal-37（packages）、goal-67 提供的符号（manifest.CheckDependencies、secret.Store、adminauth.Auth、packages.Preflight、根包的 assembly、remote.Process、policy 的 Admin 方法）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-39（切片 T-09.4：管理服务、安装记录与 serve）。goal 目录 goal-39，分支 goal-39，worktree .worktrees/goal-39，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：建立唯一的管理服务，CLI 与（之后的）WebUI 都经它的 HTTP API 操作；四种角色的组件使用同一套 install、list、enable、disable、restart、uninstall（含 --purge）、pin 流程，所有变更按包串行；install 按角色执行同类唯一，`--source` 指定并持久化来源，`--replace` 显式替换同类组件；Adapter 实例的创建、启停与删除经此服务，凭证存为 Secret 引用；启用时检查必需依赖，停用、卸载被依赖的包时列出受影响者并要求显式确认；admin add/remove/list；alonabot serve（不带子命令时相同）复用根包的 Core 装配，按安装记录启动全部已启用的包。reload、update、rollback 的入口与校验在此定义，候选编排由 T-09.5 起接入。

依据：PLAN T-09 步骤 3；CONTRACTS §9.1～§9.3、§8（依赖段）、§15.1、§15.2、§7.4、§5.3；DECISIONS D-26、D-42、D-59；TEST-MATRIX V-40、V-48、V-52、V-82、V-01 的安装分项。

范围（只新增或修改这些）：
- internal/core/state/：migrations 末尾追加一条；新增 installs_admin.go（安装记录的扩展字段与 Adapter 实例表）
- 新增 internal/core/manage/（管理服务与 HTTP API）
- 根包 assembly.go：增加按安装记录启动全部已启用包、挂载管理 API 的步骤；新增导出的 Serve 供 cmd 使用
- cmd/alonabot/：serve（以及不带子命令时）调用根包 Serve；新增 install、list、enable、disable、restart、reload、update、rollback、uninstall、pin、adapter instance、admin 命令（经管理 API 访问）
- internal/core/buttons、conversation、policy、tool、secret：在卸载与 purge 中调用其已有的清理方法（只加调用点）
- 上述包的测试文件
允许的新跨模块边：无（cmd/alonabot -> 根包已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：ALTER TABLE installs ADD COLUMN version TEXT NOT NULL DEFAULT ''; ALTER TABLE installs ADD COLUMN publisher_key_id TEXT NOT NULL DEFAULT ''; ALTER TABLE installs ADD COLUMN pin TEXT NOT NULL DEFAULT ''; ALTER TABLE installs ADD COLUMN uninstalled INTEGER NOT NULL DEFAULT 0;
   CREATE TABLE adapter_instances (instance_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), config_snapshot TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, created_at INTEGER NOT NULL);
   读写方法自定。
2. internal/core/manage
   type Service struct{ ... }
   type Options struct { /* Store、Secrets、Supervisor 装配、Packages 根目录、Log、ReleaseBaseURL（缺省 https://github.com，测试指向本地 Release 模拟服务）等，执行者按实际依赖列出 */ }
   func New(opt Options) *Service
   type Confirm struct { Affected []string } // 需要确认时返回，调用方带 confirm=true 重试
   type ErrNeedsConfirm struct{ Affected []string } // 实现 error
   // 以下方法都按包串行（复用 state.Store.LockPlugin），错误为带 Code 的 *validate.Error 或 ErrNeedsConfirm
   type InstallRequest struct {
       Archive, ReleaseJSON, Signature string // 本地包；Source 非空时由服务下载，三者留空
       Source  string // GitHub 仓库 owner/repo；从 <ReleaseBaseURL>/<owner>/<repo>/releases/latest/download/ 取 release.json 与 release.json.sig，再按其中的资产 URL 取当前平台（WebUI 为 any）的归档
       Replace bool   // 同类组件已安装时替换它
       Confirm bool
   }
   func (s *Service) Install(ctx context.Context, req InstallRequest) (Package, error)
   func (s *Service) List(ctx context.Context) ([]Package, error)
   func (s *Service) Enable(ctx context.Context, id string) error
   func (s *Service) Disable(ctx context.Context, id string, confirm bool) error
   func (s *Service) Restart(ctx context.Context, id string) error
   func (s *Service) Uninstall(ctx context.Context, id string, purge, confirm bool) error
   func (s *Service) Pin(ctx context.Context, id, constraint string) error // 空串表示解除
   func (s *Service) Reload(ctx context.Context, id string) error
   func (s *Service) Update(ctx context.Context, id, archive, releaseJSON, signature string, confirm bool) error
   func (s *Service) Rollback(ctx context.Context, id, generation string, confirm bool) error
   func (s *Service) CreateInstance(ctx context.Context, adapterPackageID, instanceID string, config []byte) error // 配置中的 secret 字段经 Secret Store 存为引用
   func (s *Service) StartInstance(ctx context.Context, instanceID string) error
   func (s *Service) StopInstance(ctx context.Context, instanceID string) error
   func (s *Service) DeleteInstance(ctx context.Context, instanceID string) error
   type Package struct { ID, Role, AdapterID, Version, Source, Pin string; Enabled, Uninstalled bool; Active string; State string }
   type Candidates interface { Reload(ctx context.Context, id string) error; Update(ctx context.Context, id, staged string) error; Rollback(ctx context.Context, id, generation string) error } // T-09.5 起实现；缺省实现返回 UNSUPPORTED
   func (s *Service) Handler(auth *adminauth.Auth) http.Handler // /api/v1/...，全部经 auth.Require
3. 根包
   func Serve(ctx context.Context, dataDir string, args []string) error // cmd/alonabot serve 使用；与 Run 共用 assembly

行为：
- install：packages.Preflight（验签、来源、平台、大小、摘要）→ MoveTo 不可变目录 → 写安装记录（install_seq 递增、source、publisher_key_id、version、role、adapter_id）。仍处于已安装状态的同 ID 返回 ALREADY_EXISTS，不覆盖。同类组件（同一 adapter_id 的 adapter、任一 agent、任一 webui）已安装时：未带 Replace 返回 ALREADY_EXISTS，safe_message 列出已安装的组件 ID；带 Replace 未带 confirm 返回 ErrNeedsConfirm（Affected 为将被卸载的组件）；确认后先完成新组件的下载与 Preflight，再按 uninstall 的缺省语义卸载旧组件，最后写新安装记录；旧组件原来启用时按 Enable 启用新组件。被替换 Adapter 的实例记录留在旧组件名下，不转给新组件，Affected 中列出这些实例。新组件在卸载旧组件之后失败时旧组件保持已卸载。Source 持久化为来源，之后的 update 只从它取版本；已卸载的保留记录只允许来源与发布者公钥都相同的重装（复用数据），否则 PERMISSION_DENIED。安装时用全部安装记录的 plugin_dependencies 检查必需依赖成环，成环拒绝。首次信任新发布者公钥需要 confirm。
- enable：检查 manifest 的必需依赖（CheckDependencies，Installed 来自已启用的包与运行中的 Adapter 描述），有缺失时 FAILED_PRECONDITION 并列出全部缺失项；可选依赖缺失只记诊断。已有 Active 时以该代次启动；首次安装没有 Active 时以「首个候选」启动（新代次、首装 SQLite、配置为 {} 或编辑快照，提交后激活；Journal 由 T-09.5 接入，本切片用一行 shortcut 注释写明）。Adapter 包启用后按 adapter_instances 中 enabled 的实例 StartInstance。
- disable：停止并保留安装与数据；被已启用包的必需依赖引用时，未带 confirm 返回 ErrNeedsConfirm（列出受影响者），确认后执行，依赖方继续运行，调用它返回 UNAVAILABLE。
- restart：以当前 Active 启动新 Session（不导入编辑快照，不产生新代次）。
- uninstall：同 disable 的确认规则；停止、解除运行注册、删除未被引用的程序包目录；默认保留数据、secret、来源绑定与 uninstalled=1 的记录。--purge 时先列出将删除的框架受管对象（代次目录、SQLite、secret、会话、按钮 token、策略、工具策略、Pending、Adapter 实例），未带 confirm 返回 ErrNeedsConfirm（Affected 为对象列表），确认后删除这些对象与安装记录；不删除插件自管的外部数据库与插件自管目录。
- pin：持久化版本约束（SemVer 范围，manifest 的语法规则），update 只接受满足 pin 的候选。
- reload/update/rollback：完成输入与前置校验（来源绑定、pin、版本不低于被依赖的 min_version 时需要 confirm），然后交给 Candidates；缺省实现返回 UNSUPPORTED。
- Adapter 实例：CreateInstance 用该 Adapter 包的配置 Schema（安装包中的 Schema 资源）经 secret.Extract 存快照；Start 时 Resolve 后经 remote.Process 启动；同一 adapter_id 下账号唯一由 remote 检查。实例配置与凭证不出现在 List 或日志中，读取只返回 Redact 后的结果。
- admin add/remove/list 经 policy 的 Admin 方法。
- HTTP API（只在 admin host，全部经 adminauth.Require）：每个方法一个 JSON 端点，需要确认时返回 409 与 {"affected": [...]}；错误返回 {"code", "safe_message"}。CLI 命令从 --admin-url（缺省取设置文件的 admin_base_url）与 ALONABOT_ADMIN_TOKEN 环境变量（或 --token-file）取得 Bearer token，不读取插件环境；确认通过 --yes 显式给出。CLI 不直接写状态库。
- WebUI 组件没有进程：enable/disable 只改记录，restart 返回 FAILED_PRECONDITION；文件的提供在 T-09.12。list 返回 role。
- serve：取得 core.lock；没有管理凭证时，stdin 与 stderr 都是终端则按 admin-token init 生成并显示一次后继续，否则拒绝启动并提示先运行 admin-token init；装配 Core（与 Run 共用 assembly），启动 HTTP Listener、webhost、管理 API，按 install_seq 顺序启动全部已启用的包。不带子命令运行 alonabot 等同 serve。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go（//go:build unix）。测试中用 goal-37 的测试工具生成签名的插件包、Adapter 包与依赖它们的插件包，在临时 data_dir 运行 alonabot serve（预先用测试钩子初始化管理 token），再用 CLI 命令操作：
  - install 后 list 显示未启用；enable 产生第一个 Active；restart 只换 Session（代次不变）。
  - 依赖 Adapter 的插件在 Adapter 未安装时 enable 失败并列出缺失项；安装 Adapter 后成功；disable Adapter 未加 --yes 时被拒并列出受影响插件，加 --yes 后执行，插件继续运行。
  - uninstall 后同来源重装复用数据；不同来源重装被拒；--purge 列出对象并在 --yes 后删除。
  - 已安装 Agent 测试组件时安装另一个 Agent：不带 --replace 返回 ALREADY_EXISTS；带 --replace 与 --yes 后旧组件已卸载、新组件已安装并沿用旧组件的启用状态；替换 Adapter 时旧实例不出现在新 Adapter 下；不同 adapter_id 的两个 Adapter 可以并存。
  - --source 指向本地 Release 模拟服务：安装成功且来源被持久化；模拟服务上的签名被篡改时安装失败，已有组件不受影响。
  - pin 持久化，重启 serve 后仍在。
  - 创建 Adapter 实例：状态库中的配置只有 secret 引用；list 与日志中没有 token；实例可启停。
  - 插件依赖成环的包在安装时被拒。
  - reload、update、rollback 返回 UNSUPPORTED（等待 T-09.5）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：串行化（同一包的并发操作排队）；ErrNeedsConfirm 的 Affected 内容；API 的错误格式；没有管理凭证且非终端时 serve 拒绝启动；Package.State 的取值。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：候选 Journal 与准备编排（T-09.5）；唯一提交点与激活门禁（T-09.6）；崩溃恢复、回退与回收（T-09.7）；管理页面（T-09.8、T-09.9）。
```

### goal-71 · T-14.1 help 插件与命令列表

状态：等待前置（T-04.4、T-05.2、T-08.8）。语句中 goal-10、goal-12、goal-13、goal-16、goal-53 提供的符号（router.CommandIndex、policy.CheckCommand、contracts/i18n、sdk 的命令注册与 Context.T）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-71（切片 T-14.1：help 插件与命令列表）。goal 目录 goal-71，分支 goal-71，worktree .worktrees/goal-71，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 提供 core.command.list：按调用者 Actor 与默认 Target 返回可用的命令（已启用、权限满足、未被遮蔽），说明按 CommandSpec 的说明键在所属插件的 messages 中按 Actor locale 翻译。官方 help 插件 plugins/help/ 注册 /help [命令] 与 /whoami。

依据：PLAN T-14 步骤 1；CONTRACTS §15.2、§2.2、§2.3、§5.4；WIRE-FORMAT §3.5（core.command.list 与 CommandSpec 的说明键）；TEST-MATRIX V-67 的 help 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：route.proto 的 CommandSpec 追加 optional string description_key = 12；新增 CommandListRequest/Response 与 CommandInfo。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceCommandList = "core.command.list"
- 新增 internal/core/commandlist/
- internal/core/supervisor 或 cmd/alonabot 的装配处：调用 SetMessages（只加调用点）
- sdk/：Description(key string) 路由选项；buildRegister 写入 description_key；新增 commands.go（ListCommands）
- 新增 plugins/help/（main 包与可导入的 help 包，含 plugin.yaml 与 i18n/en-US.json、zh-CN.json）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> sdk/...、-> contracts（已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message CommandListRequest {}
   message CommandInfo { string plugin_id = 1; string route_id = 2; repeated string path = 3; repeated CommandAlias aliases = 4; string usage = 5; optional string description = 6; }
   message CommandListResponse { repeated CommandInfo commands = 1; }
2. internal/core/commandlist
   type Service struct{ ... }
   func New(r *router.Router, p *policy.Policy) *Service
   func (s *Service) Register(m *router.Mux)
   func (s *Service) SetMessages(pluginID string, files map[string][]byte) // 文件按 contracts/i18n 解析；有问题时记 WARN 并忽略该插件的说明
3. sdk
   func Description(key string) RouteOption // 命令说明在本插件 messages 中的键
   type CommandInfo struct { PluginID, RouteID string; Path []string; Aliases [][]string; Usage string; Description string }
   func ListCommands(ctx context.Context) ([]CommandInfo, error) // 必须在 Handler 内调用（需要父调用链）
4. plugins/help
   func New() sdk.Root // 根插件

行为：
- core.command.list（必须带 parent_call_id；Actor 与默认 Target 取自父调用的可信上下文，没有 Actor 时只列出 ANY 权限的命令）：遍历 Router 的 CommandIndex，只保留每个规范路径的胜出者（被遮蔽者不列出）；对每条命令执行 policy.CheckCommand（全局禁止、启用、权限），通过的才列出；按路径字典序返回。usage 与 goal-12 的 usage 生成规则相同（按 Actor locale 使用 Core 文案目录）；description 在所属插件的 messages 中按 [Actor.locale, 默认 locale, en-US] 翻译，键不存在时 absent。
- 插件的 messages：SDK 不在 REGISTER 中上传。commandlist.Service 提供 SetMessages(pluginID string, files map[string][]byte)，由装配方调用：已安装的包读取包目录下的 i18n/*.json；库模式的开发组件使用生成器嵌入的 Generated.I18n；alonabot run 的开发入口读取 --entry 所在目录的 i18n/（存在时）。取不到时 description absent。
- help 插件：
  - /help：调用 ListCommands，按插件分组列出路径与说明，文案经 Context.T（键 help.list.header 等）。
  - /help <路径>：只显示匹配命令的 usage、别名与说明；没有可用的匹配命令时回复 help.not_found。
  - /whoami：回复调用者的 adapter_instance_id 与 external_user_id（文案键 help.whoami，参数 instance 与 user）；没有 Actor 时回复 help.whoami.no_actor。
  - 所有回复都用 Context.Send 回复原 Target；用户可见文案只来自 i18n（en-US 与 zh-CN）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/help_test.go（//go:build unix）：构建 plugins/help 与两个测试插件（一个有 ADMIN 权限的命令，一个与另一插件同路径且 priority 较低的命令），用 memory Adapter 装配：
  - 普通用户的 /help 不包含 ADMIN 命令与被遮蔽的命令；Core Admin 能看到 ADMIN 命令。
  - zh-CN 用户看到中文说明，en-US 用户看到英文；说明键缺失的命令没有说明。
  - 禁用的命令不出现；/help 某命令显示 usage。
  - /whoami 回复正确的实例 ID 与用户 ID。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有父调用链时 core.command.list 返回 FAILED_PRECONDITION；没有 Actor 时只列 ANY；排序；help 的 i18n 文件在 plugin check 中通过。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./...；go build ./...。

不做：welcome 与 rss（T-14.2）；help 的安装包与预装（T-14.3、T-14.4）；管理页面中的命令列表（T-09.9）。
```

### goal-79 · T-15.2 官方 OneBot 11 Adapter 的发送、媒体与 extension

状态：等待前置（T-03.6、T-03.8、T-15.1）。语句中 goal-7、goal-8、goal-78 提供的符号（sdk/adapter 的 Core.ReadMedia 与实例可选实现的平台文件读取接口、ExtensionDeclaration 的校验、adapters/onebot11 的实例与 API 调用、onebottest.Server）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-79（切片 T-15.2：官方 OneBot 11 Adapter 的发送、媒体与 extension）。goal 目录 goal-79，分支 goal-79，worktree .worktrees/goal-79，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

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

## 第 9 波

### goal-14 · T-04.5 示例升级与 T-04 验收

状态：等待前置（T-03.10、T-04.2、T-04.3、T-04.4）。语句中 goal-9～goal-13 提供的符号与命令（标记语法、plugin generate/check、命令选择、策略服务层、alonabot run 的 Adapter 参数）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-14（切片 T-04.5：示例升级与 T-04 验收）。goal 目录 goal-14，分支 goal-14，worktree .worktrees/goal-14，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：用同一个 SDK 升级 echo 与媒体示例：命令组、子命令、别名、Regex 与 keyword 等文本匹配、Raw、可选参数与显式实例方法，并保留一条纯显式注册路径；验证生成产物实际编译、启动、注册、接收命令并调用 Send。完成 T-04 的构建与策略验收，在 TEST-MATRIX §3.2 记录证据。

依据：PLAN T-04 步骤 6、7 与「停止和失败条件」；CONTRACTS §2.2、§2.3、§9（生成器段）；TEST-MATRIX V-10、V-11、V-12、V-43、V-49 的 T-04 分项。

范围（只新增或修改这些）：
- examples/plugins/echo/、examples/plugins/media/：加入命令组与标记注册的 Handler（生成 alonabot_gen.go 并提交），同时保留 echo 中一条只用显式注册的路径
- 新增 sdk/generate/testdata/（构建 fixture：多 root、多 Config、错误签名、依赖中的伪标记、build tags、重复路由）
- 新增 cmd/alonabot/t04_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-04 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

接口：不新增公开 Go 符号。示例使用现有 SDK API 与 //alonabot:command、//alonabot:event 标记。

行为与验收矩阵（t04_e2e_test.go 与 sdk/generate 的测试，各为一个子测试）：
- 示例：echo 注册 /echo（显式注册，纯显式路径）、/note add <text:raw>、/note list、别名 /n、/calc <a:int> [b:int=1]、一个 Regex（例如 ^roll (\d+)d(\d+)$）、一个 keyword 与一个 Raw；其中 /note 组由标记生成，/calc 为实例方法显式注册。plugin check 对两个示例通过，生成文件与提交的版本一致。
- 构建：以真实 go build 编译示例，经 alonabot run（memory 之外，用 goal-9 的 Telegram 模拟服务或回环 Adapter）启动并收发：每条命令的回复正确；参数错误时发 usage 一次；Regex 捕获组正确；命中命令时 keyword 与 Raw 不运行。
- 生成器 fixture：多 root、多 Config、错误的方法签名（编译失败或 Collect 报错，以实际 go build 的结果为准）、依赖 module 中的伪标记不被扫描、不同 build tags 下命令集合不同、重复路由与别名冲突、参数溢出（int64 越界）各产生预期的诊断或运行结果。
- 策略：两个真实插件 Session 注册同名命令，CommandIndex 与 /help 之外的诊断显示胜出者与被遮蔽者；禁用胜出者后路径仍被占用（被遮蔽者不接管）；经 policy 服务层修改权限后下一条消息即生效，插件进程 PID 不变；关闭并重启 Core 后策略仍在。
- Rule 相关的命令消费在 T-05.2 之后才有真实 SDK Rule；本切片对 Rule false/error 的消费状态用 HandlerOutcome fixture 验证，并在证据中标明。
- 证据：TEST-MATRIX §3.2 追加 T-04 一行，写明测试位置、命令、平台与结果；只有 fixture 覆盖的项单独标明。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t04_e2e_test.go；sdk/generate 中对 testdata fixture 的测试。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：示例的每条命令在 Windows/Linux 交叉编译（只编译，不运行）；生成文件的稳定性。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；在两个示例目录运行 alonabot plugin check 通过；node scripts/check-repository.mjs。

不做：真实 Rule 与 Hook 的集成（T-05.5）；WebUI 修改策略（T-09.9）。
```

### goal-18 · T-05.5 多进程事件流水线验收

状态：等待前置（T-03.8、T-05.3、T-05.4）。语句中 goal-8、goal-15～goal-17、goal-52 提供的符号（按钮、Rule、Hook、延续调用与等待、Router 的 Stats）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-18（切片 T-05.5：多进程事件流水线验收）。goal 目录 goal-18，分支 goal-18，worktree .worktrees/goal-18，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：用多个真实插件进程跑完整的事件流水线：多个 priority 与同优先级逆安装顺序、自定义 Rule、两种 Handler 签名、三个 Hook、真实按钮回调、两个用户在同一群同时等待、等待超时与等待期间停用插件；用恶意载荷尝试修改可信字段，用延迟 Handler 验证控制帧仍推进；在 TEST-MATRIX §3.2 记录 T-05 证据。

依据：PLAN T-05 步骤 8 与「停止和失败条件」；CONTRACTS §2；MESSAGE-MODEL §1、§6.3；TEST-MATRIX V-11、V-13～V-16、V-43、V-49、V-50 的 T-05 分项。

范围（只新增或修改这些）：
- 新增 internal/core/supervisor/testdata/pipeline/（流水线测试插件，按环境变量配置 priority、Rule、Hook 与命令）
- 新增 cmd/alonabot/t05_e2e_test.go（//go:build unix，经 Telegram 模拟服务运行）或 internal/core/supervisor/pipeline_test.go（memory Adapter），执行者按需要真实按钮回调的项选择前者
- spec/TEST-MATRIX.md §3.2：追加 T-05 的证据行（本切片获授权修改这一处）
允许的新跨模块边：无。

接口：不新增公开 Go 符号。

验收矩阵（各为一个子测试，以平台调用记录、候选调用次序日志与 HandlerOutcome 为证据）：
1) 三个插件 priority 为 0、0、-1，以两种不同的安装与启动顺序运行：调用顺序都是 (priority, plugin_id, route_id)。
2) 自定义 Rule 为 false 与出错：事件路由跳过、已识别命令被消费且不回落（V-11 的真实 Rule 分项）。
3) 两种 Handler：自动回复只发一次；手动多次 Send；自动型内 Send 被拒。
4) Hook：PreEvent 改 payload、BeforeSend 改链与拒绝、AfterSend 出错不改结果；每次 Send 一组 Hook。
5) 恶意载荷：Adapter 侧（经模拟服务）与插件侧（PreEvent 返回）尝试改 Actor、Target、button 的 action/value、scheduled 的 job_id：插件收到的仍是 Core 保留的值。
6) 真实按钮：发送带按钮的消息后在模拟服务中点击：先有 answerCallbackQuery，再调用正确的插件；另一用户点击也调用一次；过期 token 不调用、不落入 Raw。
7) 等待：两个用户在同一群同时进入 Wait，各自只收到自己的下一条消息，且这些消息不进入普通分发；超时为 DEADLINE_EXCEEDED；等待期间停用插件，重启后不补发。
8) 延迟 Handler（睡眠数秒）执行期间，同一插件 Session 的另一请求的 Response 与 Cancel 正常完成，其他事件照常派发。
9) 权限组合（V-43）：chat_admin 不取得 Core Admin；组禁用优先于子项启用；最近显式权限覆盖父项；运行时修改在下一次派发生效，已开始的调用不被撤销。
10) scheduled：只用规范 fixture，证据中标明尚未覆盖真实 Scheduler（T-07.5 覆盖）。
- 证据：TEST-MATRIX §3.2 追加 T-05 一行，写明测试位置、命令、平台、结果与 fixture 覆盖的项。

测试（按编码规范「测试提交规则」）：
- 提交：上述矩阵测试。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：在 -race 下重复运行矩阵 10 次无偶发失败。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：Service 嵌套与 Hook→Service→Send（T-06.2、T-06.4）；真实 Scheduler（T-07.5）。
```

### goal-20 · T-06.2 父调用链、取消与背压

状态：等待前置（T-03.6、T-05.3、T-06.1）。语句中 goal-7、goal-17、goal-19 提供的符号（broker.Broker、router.Invocation 与 send_forbidden、media 的 Grant、wire 的 ParentAttach 与预算）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-20（切片 T-06.2：父调用链、取消与背压）。goal 目录 goal-20，分支 goal-20，worktree .worktrees/goal-20，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Broker 沿父调用链传播 Core 记录的可信上下文（actor、target、scope、trace、祖先插件链、send_forbidden），不接受插件覆盖；同链重复插件返回 CALL_CYCLE，深度超过 CALL_CHAIN_MAX_DEPTH 返回 FAILED_PRECONDITION；子调用期限取父与自身预算中较早者；父完成、超时、取消或任一 Session 断开时取消未完成的子调用；复用 wire 的在途数与字节预算，额满立即失败，不出现互相等待额度的死锁。

依据：PLAN T-06 步骤 3、4；CONTRACTS §3 第 4、5 条、§1.2、§2.3（send_forbidden）；WIRE-FORMAT §4；TEST-MATRIX V-18、V-42 的嵌套分项、V-15 的间接递归分项。

范围（只新增或修改这些）：
- internal/core/broker/：调用链记录、循环与深度检查、期限与取消传播
- internal/core/router/：Invocation 增加祖先插件链与 scope，并在 InvocationContext 中下发（只加字段与填写处）
- internal/wire/：如需在父请求结束时通知子调用取消，只加回调字段与触发点（与 goal-24 的 Late 同样的方式）
- 新增 internal/core/supervisor/testdata/chainplugin/
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- internal/core/router：Invocation 增加字段 Chain []string（祖先插件链，最早者在前）、Scopes []string、SendForbidden bool（goal-17 已有时沿用）
- internal/core/broker：Options 增加 MaxDepth int（tune.Values.CallChainMaxDepth）；New 的签名相应改为接收 Options，并更新全部调用者

行为：
- 父请求：插件发起的 Service 调用带 parent_call_id 时，必须指向同一 Session 中仍在途的 Core→该插件请求（wire 已保证属于同一 Session）；Broker 从该父请求的 Invocation 派生上下文：actor、default_target、scopes、trace_id、send_forbidden 原样继承，Chain = 父 Chain + [调用方]。插件载荷中的任何身份字段被忽略。没有父请求时只有插件身份：Chain = [调用方]，actor 与 target 为空。
- 检查：目标插件已在 Chain 中 → CALL_CYCLE；len(Chain)+1 > MaxDepth → FAILED_PRECONDITION。在调用目标函数之前返回。不同的独立链互不影响（检查只看本链）。
- 期限：子请求期限 = min(父请求期限, 调用方请求自带期限, REQUEST_TIMEOUT 起算的期限)。
- 取消：父请求完成（成功或失败）、到期、被取消，或调用方 Session、目标 Session 断开时，Broker 取消尚未完成的转发请求（发送 Cancel），并使调用方得到 CANCELLED 或 UNAVAILABLE；不重放、不重试。普通迟到响应不会再次完成已结束的调用。
- 背压：转发请求使用目标 Session 的现有准入（在途 256、每方向字节预算、Core 全局预算），额满时立即向调用方返回 RESOURCE_EXHAUSTED，不排队等待。Broker 不在持有任何锁时等待业务槽位；Response、Cancel 与控制帧的处理不依赖 Broker。
- send_forbidden 的间接路径：出站 Hook 内调用 Service、该 Service 再调用 core.send 时返回 FAILED_PRECONDITION（V-15 的必交回归）。
- 媒体：Service 中向原事件 Target Send 仍经 T-03 准入；跨 Session 使用媒体仍须 Grant。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/chain_test.go（//go:build unix）。构建 testdata/chainplugin（按环境变量决定收到调用后转调哪个插件、睡眠多久、是否在 BeforeSend 中调用 Service），以 a、b、c 启动：
  - A→B→C 正确返回，C 看到的 Actor 与 Target 是原事件的，Chain 为 [a, b]。
  - A→B→A 返回 CALL_CYCLE，A 的第二次函数未运行；两条同时进行的独立链 A→B 与 C→B 都成功。
  - 把 MaxDepth 设为 2：A→B→C 返回 FAILED_PRECONDITION。
  - A 的父事件 Handler 到期：B、C 的未完成调用被取消（C 收到取消，日志可见），A 得到错误；没有重试。
  - 杀掉 B 的进程：A 立即得到 UNAVAILABLE。
  - 出站 Hook 中经 Service 再 Send 被拒（FAILED_PRECONDITION）。
  - 用很小的 wire 预算使转发请求额满：立即 RESOURCE_EXHAUSTED，期间其他 Session 的 Response 照常完成。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：期限取最小值；没有父请求的后台调用只有插件身份；插件载荷中伪造的 actor 被忽略；Cancel 不越过其 Request。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./sdk/...；go build ./...。

不做：Event Bus（T-06.3）；Session 清理的完整矩阵（T-06.4）。
```

### goal-33 · T-08.6 SQLite 候选与资源交接

状态：等待前置（T-08.4、T-08.5）。语句中 goal-31、goal-32 提供的符号（secret.Store、supervisor.Spec 的 Database 与 ConfigSchema、gormdb 的 SQLite 设置）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-33（切片 T-08.6：SQLite 候选与资源交接）。goal 目录 goal-33，分支 goal-33，worktree .worktrees/goal-33，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：在已持插件操作锁、旧写入者已退出且 Core 能取得 runtime.lock 的前提下，用 SQLite Backup API 从 Active 代次的数据库创建独立的候选副本（首次安装创建空候选，reload 与 update 同样复制，普通 resume 复用 Active）；复制完成后释放临时 runtime.lock，由 SDK 取锁打开候选。把准备好的程序、配置快照、secret 引用、数据库路径、配置摘要、迁移状态与本次身份作为一个结果交给 T-09，不写 Active 指针、不伪造 COMMITTED；失败清理只处理未受保护的候选资源。

依据：PLAN T-08 步骤 7；CONTRACTS §6.1、§6.3 第 1、2 条、§12；TEST-MATRIX V-24、V-36 的候选分项。

范围（只新增或修改这些）：
- 新增 internal/core/candidate/
- internal/core/state/：读取代次 DB 路径与保留关系所需的方法（新增文件，不改已有方法）
- 上述包的测试文件
允许的新跨模块边：internal/core/candidate -> internal/oslock（已允许 internal/core -> internal/oslock）；SQLite Backup 使用已批准的 modernc.org/sqlite。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/candidate
   type Reason int // FirstInstall、Reload、Update、Resume
   type Request struct {
       PluginID     string
       Reason       Reason
       FromGeneration string // Active 代次；FirstInstall 时为空
       NewGeneration  string // Resume 时等于 FromGeneration
       PluginDir    string   // <plugins>/<id>
   }
   type Prepared struct {
       Generation   string
       DatabasePath string
       Created      bool     // 本次是否新建了候选文件
       // 以下由调用方在 SDK 准备完成后填写，再交给 T-09 的提交流程
       ConfigSnapshot []byte
       SecretPaths    []string
       ConfigDigest   []byte
       Migration      *pb.MigrationState
   }
   func PrepareDatabase(ctx context.Context, req Request) (*Prepared, error)
   func Discard(ctx context.Context, st *state.Store, p *Prepared) error // 删除未受保护的候选文件与只被它引用的 secret
   // 交接约定：本包的文档注释写明调用顺序（见下），供 T-09.5 引用

行为：
- 前置：调用方已持有 state.Store.LockPlugin(pluginID)，旧进程已停止。PrepareDatabase 先非阻塞地取得 <PluginDir>/runtime.lock（oslock.TryLock）；取不到 → FAILED_PRECONDITION（"plugin process still holds runtime.lock"），不打开任何数据库。
- FirstInstall：创建 generations/<new>/ 目录（0700），不创建数据库文件（SDK 首装时创建），Created=false。
- Reload/Update：打开 generations/<from>/db.sqlite（只读不可用时报错，不能因文件缺失而创建空库），用 modernc 的 Backup API 复制到 generations/<new>/db.sqlite（先写临时文件，完成后 fsync 并重命名），核对副本的 PRAGMA integrity_check 为 ok；Created=true。不能只复制主文件。
- Resume：返回 Active 代次的路径，不复制。
- 完成后（成功或失败）释放 runtime.lock，再返回；SDK 随后按自己的流程取锁打开。
- 失败：备份、空间或校验失败时删除临时文件与本次新建的目录，返回错误；不触碰 FromGeneration 的文件。
- Discard：只删除 NewGeneration 的目录及仅被该候选引用的 secret（secret.Collect，保护 Active、Pending 与保留代次）；Active 或保留代次的资源一律不删。
- 交接约定（写在包文档）：1) 持插件锁并停止旧进程 → 2) PrepareDatabase → 3) AddGeneration（未提交）→ 4) Launch（NewCandidate=true，Database 为候选路径）→ 5) PREPARED 后由 T-09.6 的唯一提交点在同一事务中提交 Active 与 COMMITTED → 6) 失败时 Discard 并重启旧版的新 Session。本包不写 Active，不写 COMMITTED。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/candidate_test.go（//go:build unix）。测试插件在 OnStart 中写一行，并按环境变量在写入后让 OnStart 失败：
  - 首装成功后 reload：候选是独立文件，候选中的写入不出现在 Active 文件中；OnStart 失败时 Discard，Active 的数据与配置不变，重启旧版正常。
  - 插件进程仍在运行（持有 runtime.lock）时 PrepareDatabase 返回 FAILED_PRECONDITION，没有打开数据库。
  - Active 处于 WAL 模式且有未检查点的写入时，候选副本包含这些写入（证明使用了 Backup API）。
  - Discard 不删除 Active 引用的 secret。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：临时文件的重命名；integrity_check 失败时的清理；Resume 不复制；FirstInstall 不创建文件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Journal、唯一提交点与恢复（T-09.5～T-09.7）；external 插件的候选门槛（T-09.5）。
```

### goal-84 · T-09.12 WebUI 组件与组件图标

状态：待编写（前置 T-07.4、T-09.1、T-09.4、T-13.1）。依据 D-62、D-63 新增；前置合并后由审查者按实际的 webhost、adminauth、manage 与根包装配写出接口和测试段。

目标：Core 在 admin host 上以组件方式提供 WebUI 与组件图标：已启用 WebUI 组件的静态文件（保留 `/api/` 与内部路径，无扩展名的不存在路径回退 index.html，未安装时 503 与英文提示），Admin API 主版本常量与 WebUI 声明范围的检查（不兼容时拒绝安装与启用并提示），只提供 manifest 与表单选项声明的图标并带 sandbox CSP 与 nosniff、要求管理员会话；写出 spec/ADMIN-API.md（主版本、全部接口、认证与错误格式）；根包增加 `alonabot.WebUI(fsys fs.FS)` 选项。

依据：PLAN T-09 步骤 9、§3.1 的 T-09.12 行；CONTRACTS §7.3、§7.5、§9.1、§15.4；DECISIONS D-62、D-63；TEST-MATRIX V-83。

要点：
- 文件只从不可变包目录读取，路径按 CONTRACTS §9.1 的相对路径规则校验，不跟随链接，不经组件进程。
- 图标的类型只按扩展名取 image/svg+xml、image/png、image/webp；未声明的包内文件即使存在也返回 404。
- spec/ADMIN-API.md 由本切片新建，之后新增接口的切片同步追加；同一主版本内只做向后兼容的增加。
- 提交的集成测试经真实 Core 与临时安装的 WebUI 测试包覆盖 V-83 的全部场景；含脚本的 SVG 在浏览器中不执行的检查没有浏览器时记为未运行。

### goal-59 · T-12.2 Agent 对话与流式显示

状态：等待前置（T-03.9、T-04.3、T-05.4、T-08.3、T-08.4、T-11.3、T-12.1）。语句中 goal-12、goal-30～goal-32、goal-51、goal-52、goal-56、goal-58 提供的符号（命令注册与选项、Context.Edit、Continue、会话函数、provider.Client、gormdb.Open、sdk.Description 等）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-59（切片 T-12.2：Agent 对话与流式显示）。goal 目录 goal-59，分支 goal-59，worktree .worktrees/goal-59，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：建立官方 Agent 组件 agent/（manifest `role: agent`）的骨架与配置（Provider 与预设、人格、历史条数等），注册 /ask、/agent on|off、/persona、/provider、/reset 与 message 事件路由；一轮对话在延续调用中执行：读取会话历史、组装请求、经 Provider 流式接收，先发送一条消息再按 1s 间隔编辑，结束时以完整内容最后编辑，超长时在段落边界换新消息；结束后追加会话条目。Provider 出错或中途断开时保留已显示内容并按 Actor locale 报告，不自动重发。

依据：PLAN T-12 步骤 2、3；CONTRACTS §13（Agent 段）、§2.4、§8（Edit）、§5.4；DECISIONS D-27、D-38；TEST-MATRIX V-57。

范围（只新增或修改这些）：
- agent/：新增 agent 包（可导入的根组件 agent.New()）、cmd/ 下的 main、plugin.yaml（`role: agent`、`icon`）、i18n/en-US.json 与 zh-CN.json、assets/icons/（Provider 预设与组件自身的图标）；provider 包已存在，不改
- 上述包的测试文件
允许的新跨模块边：agent/... -> sdk/...、-> contracts（已允许）。不 import internal/...。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
agent
   func New() sdk.Root
   type Config struct {
       config.Base
       Providers       []ProviderConfig `json:"providers"`
       DefaultProvider string           `json:"default_provider"`
       Personas        map[string]string `json:"personas"`        // 名称 → 系统提示
       DefaultPersona  string           `json:"default_persona"`
       HistoryEntries  int              `json:"history_entries" form:"default=20;min=0;max=500"`
       MessageLimit    int              `json:"message_limit" form:"default=4096;min=256;max=65536"`
       MCPServers      []MCPServer      `json:"mcp_servers"`     // T-12.3 使用
       SkillsDir       string           `json:"skills_dir"`      // T-12.5 使用
   }
   type ProviderConfig struct { Name string `json:"name"`; Preset string `json:"preset"`; Protocol string `json:"protocol"`; BaseURL string `json:"base_url"`; APIKey string `json:"api_key" secret:"true"`; Model string `json:"model"`; MaxTokens int `json:"max_tokens" form:"default=2048;min=1"` }
   type MCPServer struct { Name string `json:"name"`; Transport string `json:"transport"`; Command string `json:"command"`; Args []string `json:"args"`; Env []SecretPair `json:"env"`; URL string `json:"url"`; Headers []SecretPair `json:"headers"` }
   type SecretPair struct { Name string `json:"name"`; Value string `json:"value" secret:"true"` }

行为：
- 配置：preset 的表单选项为 openai、anthropic、deepseek、ollama 与 custom，各附 assets/icons/ 下的图标（选项图标的 tag 写法以 goal-28 合并后的生成器为准）；preset 给出 protocol 与 base_url 的缺省值，配置中显式给出的值优先。OnLoad 校验 providers 的 name 唯一、protocol 为 openai、openai-compatible 或 anthropic-messages、default_provider 与 default_persona 指向存在的项（列表为空时允许为空）；错误时返回 OnLoad 错误。
- 数据：每个聊天（adapter_instance_id、chat_id、thread_id）当前的人格、Provider 与自由对话开关保存在 Agent 的数据库（gormdb.Open，versioned 迁移）。
- 命令：/ask <text:raw>（缺省 USER）、/agent on|off（ADMIN）、/persona [name]（ADMIN，无参数时列出）、/provider [name]（ADMIN）、/reset（USER），说明用 sdk.Description 指向 messages。自由对话开启时，该聊天中 to_me 的非命令 message.created 也触发一轮（message 事件路由带 ToMe Rule，并检查开关）；私聊的消息都是 to_me。默认关闭。
- 一轮对话（Handler 调用 Continue，在延续调用中执行）：
  1) 没有可用 Provider 时用 Context.T 回复 agent.no_provider，结束。
  2) 会话键为 "<instance>/<chat>/<thread>"；ListConversation 读取最近 history_entries 条（从末尾往前，按序号顺序组装）。条目为 JSON 对象 {"role": "user"|"assistant", "speaker": 外部用户 ID（群聊时）, "content": 文本}。
  3) 以人格为 system 消息，组装 provider.Request（MaxTokens 取 Provider 配置）。
  4) 事件的 capabilities 含 message.edit 时：发送一条占位消息（agent.thinking），之后流式接收：每累计到新的文本时，距上次编辑不少于 1s（命名常量）才 Edit；Edit 返回限流错误时把下一次编辑推迟 retry_after；生成结束后用完整内容最后编辑一次。不含 message.edit 时：不发占位消息，不发送中间增量，生成结束后一次发送完整内容。回复以纯文本发送（不使用 markdown 格式）。
  5) 当前消息的内容将超过 message_limit 时，在最近的段落边界（"\n\n"，没有时为 "\n"，再没有时为空白，最后硬切）结束当前消息，再发送新消息继续显示；不能编辑时按同样的切分发送多条。
  6) 结束后用 AppendConversation 追加本轮的 user 与 assistant 条目。
  7) Provider 返回错误或中途断开：保留已显示的内容（不能编辑时不发送已生成的部分），在其后编辑或发送 agent.error（按 Actor locale，参数为错误码），不追加 assistant 条目，不重发请求。
- /reset：ClearConversation 清除该会话。
- 工具、MCP、记忆与技能分别在 T-12.3、T-12.4、T-12.5 接入；本切片不调用它们。

测试（按编码规范「测试提交规则」）：
- 提交：agent/agent_test.go（//go:build unix）。用本仓库 Core 的测试装配（supervisor、memory Adapter，实现 Editor 与 Named）、真实状态库与 agent/provider 的本地模拟 Provider（三种协议各一次）：
  - /ask hello：memory 先收到一条发送，随后收到若干次编辑，最后一次编辑是完整回复；编辑间隔不小于 1s（用注入时钟或计时断言）。
  - 回复超过 message_limit（测试中设为 256）时分成多条消息，切分点在段落边界。
  - 实例不声明 message.edit 时：memory 只收到完整回复的发送（超过 message_limit 时为多条），没有占位消息与编辑。
  - 模拟 Provider 中途断开：已显示内容保留，最后有错误说明，会话中没有 assistant 条目，Provider 只收到一次请求。
  - 自由对话关闭时群聊中的 to_me 消息不触发；Admin /agent on 后触发；非 to_me 的消息不触发。
  - /persona 切换后请求的 system 消息改变；/reset 后历史为空。
  - 普通用户执行 /agent 被拒（不回复）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：段落切分函数；OnLoad 的校验；限流推迟；会话条目格式；配置 Schema 由生成器生成且 api_key 为 secret。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... ./internal/core/...；go build ./...；在 agent 运行 alonabot plugin check 通过。

不做：工具循环与 MCP（T-12.3）；记忆（T-12.4）；技能（T-12.5）；定时任务（T-12.6）；打包（T-12.7）；真实 Provider（T-10）。
```

### goal-68 · T-13.2 开发组件的配置与数据库

状态：等待前置（T-08.3、T-08.5、T-13.1）。语句中 goal-30、goal-32、goal-67 提供的符号（sdk/config.Prepare、gormdb 的描述、根包 assembly 与 dev 安装记录）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-68（切片 T-13.2：开发组件的配置与数据库）。goal 目录 goal-68，分支 goal-68，worktree .worktrees/goal-68，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：库模式的开发组件每次启动以 <data_dir>/plugins/<id>/config.json（缺失时为空对象）为配置，经 SDK 准备与校验后直接生效，不走候选与 Journal；受管 SQLite 为 <data_dir>/plugins/<id>/dev.sqlite，不复制、不回退；dev 来源拒绝 update 与 rollback；开发组件的缺省日志级别为 debug。

依据：PLAN T-13 步骤 3；CONTRACTS §15.4；TEST-MATRIX V-64。

范围（只新增或修改这些）：
- 根包 assembly.go：开发组件的 Bootstrap 配置与数据库描述（只改开发组件分支）
- internal/core/manage/：dev 来源的 update 与 rollback 拒绝（若 goal-39 已合并；否则在 assembly 中拒绝对应入口）
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- 配置：开发组件启动时读取 config.json；文件不存在时为 {}；内容不是 JSON 对象时 Run 返回错误并指明路径。candidate_config_json 为其内容，previous_config_version 为 nil（每次都按首装处理 OnLoad），SDK 按 goal-30 准备与校验；校验失败时该组件启动失败，Run 返回错误。准备后的配置不回写 config.json，也不写入代次快照之外的地方（代次 dev 的快照每次覆盖）。
- Secret：配置中的 secret 字段在开发模式下原样下发，不经 Secret Store，代次快照中也保存原样（dev 来源不分发）；在包文档写明这一差异。
- 数据库：数据库描述为 SQLite，路径 <data_dir>/plugins/<id>/dev.sqlite，activation_reason 为 first_install（文件不存在时）或 resume，迁移按 goal-32 执行；不创建候选副本，不回退。
- dev 来源的 update、rollback 返回 FAILED_PRECONDITION（"dev components cannot be updated or rolled back"）。
- 日志：开发组件的来源级别缺省为 debug（只在没有显式设置时）。
- Adapter 实例目录已在 goal-67 中实现，本切片不改。

测试（按编码规范「测试提交规则」）：
- 提交：alonabot_test.go（根包）追加：examples/devbot 的插件加一个带默认值与 required 字段的 Config 并在 OnStart 写 dev.sqlite：
  - 没有 config.json 时以默认值启动；写入合法的 config.json 后重启，插件看到新值；写入缺少 required 的配置时 Run 返回错误。
  - 两次启动之间 dev.sqlite 中的数据保留，没有 generations/ 目录与候选副本。
  - 经管理服务（或装配入口）对 dev 组件执行 update 与 rollback 被拒。
  - 日志中有该组件的 DEBUG 记录。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：config.json 非对象；dev.sqlite 不存在时的 activation_reason；显式设置级别时不被覆盖。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：热部署（T-13.3）；库模式示例与验收汇总（T-13.4）。
```

### goal-72 · T-14.2 welcome 与 rss 插件

状态：等待前置（T-05.2、T-07.2、T-08.3、T-08.5、T-08.8）。语句中 goal-16、goal-24、goal-30、goal-32、goal-53 提供的符号（ChatAdmin Rule、ScheduleRegistry.Every、Config 与 OnLoad、gormdb.Open、Context.T、sdk.Send）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-72（切片 T-14.2：welcome 与 rss 插件）。goal 目录 goal-72，分支 goal-72，worktree .worktrees/goal-72，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：预装插件 plugins/welcome/ 在 member.joined 时向该聊天发送欢迎语，按聊天开关与自定义文案；plugins/rss/ 提供 /rss add|list|remove，订阅经 sqlitedb.Open 保存在插件数据库（不链接 GORM，D-49），固定 interval 的定时任务抓取 RSS 2.0 与 Atom，按 guid 或链接去重，只推送新条目；抓取只允许 http/https，拒绝 loopback、私网、链路本地、组播与未指定地址（含每次重定向之后），响应上限 5 MiB、超时 30s。

依据：PLAN T-14 步骤 2；CONTRACTS §15.2、§8（后台 Send 需要许可）、§5.4；TEST-MATRIX V-67。

范围（只新增或修改这些）：
- 新增 plugins/welcome/、plugins/rss/（各含可导入的包、cmd/ 下的 main、plugin.yaml、i18n/en-US.json 与 zh-CN.json）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> sdk/...、-> contracts（已允许）。只用标准库解析 XML（encoding/xml）与 HTTP。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- plugins/welcome：func New() sdk.Root
- plugins/rss：func New() sdk.Root；type Config struct { config.Base; IntervalMinutes int `json:"interval_minutes" form:"default=10;min=1;max=1440"` }

行为：
- welcome：
  - /welcome on|off|set <text:raw>：只有 chat_admin 或 Core Admin 能执行（Rule：Any(ChatAdmin(), CoreAdmin())）；设置按聊天（instance、chat、thread）保存在插件数据库。默认关闭。
  - member.joined 时，该聊天已开启则发送欢迎语：自定义文案或 messages 中的 welcome.default；{name} 替换为新成员的显示名（没有时为 external_user_id）。回复原 Target，不需要许可。
  - 只响应 member.joined。
- rss：
  - /rss add <url> 与 /rss remove <url>：只有 chat_admin 或 Core Admin（Any(ChatAdmin(), CoreAdmin())）；/rss list 不限。订阅（url、目标 Target、最后条目标识集合的摘要）保存在插件数据库。add 时先做一次抓取校验（地址规则同下），失败时回复原因（文案键）。
  - 定时任务：Every(interval_minutes 分钟) 检查全部订阅；每个订阅抓取后解析 RSS 2.0（channel/item 的 guid、link、title）与 Atom（entry 的 id、link[rel=alternate 或缺省]、title）；条目标识为 guid（或 id），没有时为链接；只推送此前未见过的条目（首次订阅时只记录现有条目，不推送），按发布顺序发送，经 sdk.Send 显式 To（需要 allowed_adapter_instances，缺少许可时记 WARN 并跳过）。
  - 抓取：只允许 http 与 https；自定义 http.Client 的 DialContext 在连接前解析地址并拒绝 IsLoopback、IsPrivate、IsLinkLocalUnicast、IsLinkLocalMulticast、IsMulticast、IsUnspecified 的地址；CheckRedirect 对每次重定向的 URL 同样检查 scheme，连接时再检查地址；响应体以 io.LimitReader 读取 5 MiB 加一字节，超过即失败；整个抓取超时 30s（命名常量）。失败记 WARN（订阅 URL 与原因），不推送、不重试到下一个周期之前。

测试（按编码规范「测试提交规则」）：
- 提交：
  - plugins/rss/rss_test.go（//go:build unix）：用 memory Adapter 与真实 Core 装配，httptest 提供订阅源。注意 httptest 监听 loopback，测试通过包内的测试钩子把「允许 loopback」打开（钩子只在测试中设置，生产默认关闭），并另有子测试在钩子关闭时验证 loopback 被拒：
    - 普通用户 /rss add 被拒；chat_admin 成功；list 显示订阅。
    - 首次抓取不推送；源中新增两个条目后下一次定时任务推送两条，顺序正确；再次运行不重复推送。
    - 没有发信许可时不推送并记 WARN；授予许可后推送。
    - 响应超过 5 MiB、超过 30s（注入更短的超时）、重定向到 127.0.0.1（钩子关闭）、file:// 与 ftp:// 都失败。
    - Atom 与 RSS 2.0 各一个源。
  - plugins/welcome/welcome_test.go（//go:build unix）：开关权限、自定义文案与 {name} 替换、关闭时不发送、只响应 member.joined。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：地址检查函数（IPv4/IPv6 私网、链路本地、映射地址 ::ffff:127.0.0.1）；条目标识的优先级；XML 中缺少字段的条目。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...；两个插件目录运行 alonabot plugin check 通过。

不做：发行归档与预装（T-14.3、T-14.4）；HTML 清洗与富文本推送；代理设置（沿用进程环境的 HTTP_PROXY）。
```

### goal-80 · T-15.3 跨 Adapter 一致性验收

状态：等待前置（T-03.10、T-15.2）。语句中 goal-9、goal-78、goal-79 提供的符号（alonabot run 的 --adapter-id/--adapter-entry、cmd/alonabot/telegram_e2e_test.go 的装配方式、telegramtest.Server、onebottest.Server）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-80（切片 T-15.3：跨 Adapter 一致性验收）。goal 目录 goal-80，分支 goal-80，worktree .worktrees/goal-80，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./cmd/alonabot/... ./adapters/...；go build ./...；node scripts/check-repository.mjs。

不做：修改两个 Adapter 的映射或 Core 的派发（发现不一致时标 [!] 并在 PR 中写明用例与差异，由审查者决定修契约还是修实现）；alonabot run 同时启动多个 Adapter；反向 WebSocket。
```

## 第 10 波

### goal-22 · T-06.4 Session 清理与 Service 端到端

状态：等待前置（T-06.2、T-06.3）。语句中 goal-15、goal-17、goal-19～goal-21 提供的符号（router、broker、bus、hook 与各自的 Stats）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-22（切片 T-06.4：Session 清理与 Service 端到端）。goal 目录 goal-22，分支 goal-22，worktree .worktrees/goal-22，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Session 停止或失效时，按 T-02 的冻结、drain、STOPPING 流程停止新的根调用，取消在途父链，并清理路由目标、Bus 订阅队列、Broker 映射与预算，不把未完成的工作转交给同插件的新 Session。错误只返回安全的 ErrorInfo 与 trace。记录按原因区分的拒绝、取消、离线与丢弃计数。在 TCP 与 stdio 上完成 A/B/C 嵌套与故障的端到端验收（含 Hook→Service→Send），在 TEST-MATRIX §3.2 记录 T-06 证据。

依据：PLAN T-06 步骤 6、7 与「停止和失败条件」；CONTRACTS §3、§1.2、§4.2、§14.5；TEST-MATRIX V-15、V-17、V-18、V-19、V-42。

范围（只新增或修改这些）：
- internal/core/broker/、internal/core/bus/、internal/core/router/：Session 失效时的清理入口与计数（新增文件或只加调用点）
- internal/core/supervisor/：停止流程中调用上述清理（只加调用点）
- 新增 internal/core/metrics/（按原因的计数，只在内存中，供管理面读取）
- 新增 cmd/alonabot/t06_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-06 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/metrics
   type Reason string // denied、cancelled、offline、dropped_queue、dropped_budget、cycle、depth
   type Counters struct{ ... }
   func (c *Counters) Add(pluginID string, kind string, r Reason) // kind 为 service、bus、event
   func (c *Counters) Snapshot() map[string]map[string]map[Reason]uint64 // plugin → kind → reason → count
broker、bus、router 的 Options 增加 Metrics *metrics.Counters（可为 nil）。

行为：
- 冻结：插件进入 Draining 后，Broker 对以它为目标的新调用返回 UNAVAILABLE（offline），Bus 不再向它入队，Router 不再派发给它。
- 失效（STOPPING 或断链）：Broker 取消它作为调用方或目标的全部转发请求（调用方得到 UNAVAILABLE 或 CANCELLED），释放这些请求占用的预算；Bus 丢弃它的订阅队列（计入 offline）；Router 移除它的路由目标与等待（与 goal-52 的 Release 一致）。不转交新 Session，新 Session 从空状态开始。
- 错误：返回给调用方的 ErrorInfo 只含 code、safe_message 与 trace_id；远端堆栈、Secret、文件路径只进 Core 日志。
- 计数：每次拒绝、取消、离线、队列丢弃、CALL_CYCLE、深度超限都计数；不持久化，不建设可靠事件存储或自动重试。
- 验收矩阵（t06_e2e_test.go，TCP 与 stdio 各运行一遍）：
  1) A→B→C、A→B→A、深度边界、两条独立并发链：结果、Actor/Target、deadline 与取消在每一跳一致。
  2) A 的出站 Hook → B 的 Service → Send：FAILED_PRECONDITION（闭合 V-15）。
  3) B/C 挂起、进程退出、满额、畸形输入与输出、排队前取消：没有死锁，没有重复的业务效果（以目标插件的调用计数证明）。
  4) Bus：慢订阅者、断线、条数满与字节满同时发生时，新事件按规则丢弃，其他订阅者继续收到。
  5) 每项记录实际进程、传输、命令与日志摘录。
- 证据：TEST-MATRIX §3.2 追加 T-06 一行，写明测试位置、命令、平台、传输与结果。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t06_e2e_test.go（上述矩阵，复用 goal-19～goal-21 的测试插件）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Counters 的并发；失效后预算归还到初始值；ErrorInfo 中不含路径与堆栈。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：管理面对计数的读取接口（T-09.9）；可靠事件存储与服务重试（契约不提供）。
```

### goal-35 · T-08.9 configdb 示例与 T-08 验收

状态：等待前置（T-08.6、T-08.7、T-08.8）。语句中 goal-28～goal-34、goal-53 提供的符号与工具（生成器、Prepare、secret.Store、gormdb、candidate、web/admin 的表单模型、contracts/i18n）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-35（切片 T-08.9：configdb 示例与 T-08 验收）。goal 目录 goal-35，分支 goal-35，worktree .worktrees/goal-35，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：新增 examples/plugins/configdb，覆盖配置迁移、显式空值、未知与 secret 字段、CRUD 与事务、versioned 迁移与失败、候选 OnStart 写库后失败；在临时 data_dir 上分层验证，数据库部分 sqlitedb 与 gormdb 各跑一遍；用 web/admin 的 fixture 验证表单的字段顺序、hidden、i18 回退、掩码与 PATCH；在 TEST-MATRIX §3.2 记录 T-08 的组件证据，并单独列出需要 T-09 正式入口与 Journal 才能关闭的场景。

依据：PLAN T-08 步骤 9 与「停止和失败条件」；CONTRACTS §5、§6；TEST-MATRIX V-23～V-26、V-38、V-39、V-54 的 T-08 分项。

范围（只新增或修改这些）：
- 新增 examples/plugins/configdb/（含 plugin.yaml、i18n/、Config 与 versioned 迁移）
- web/admin/fixtures/：加入由 configdb 的 Config 生成的 Schema fixture，并在 schema.test.ts 中追加对应用例
- 新增 cmd/alonabot/t08_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-08 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

接口：不新增公开 Go 符号。

验收矩阵（各为一个子测试）：
1) 配置迁移：config_version 1 的旧配置只有经 OnLoad 才满足 2 的 Schema；先迁移再校验，结果写入候选快照。
2) 显式空值：false、0、""、[]、{}、null（指针字段）全部保留；缺失键补默认值；未知键与嵌套未知键保留在快照中。
3) secret：快照只有引用；Schema 移除 secret 字段但候选保留该键时仍为引用；日志中没有明文。
4) DB：CRUD 与事务（tx 回调）；versioned 两步迁移；第二步失败时账本只有第一步、候选失败、Active 不变。
5) 候选：reload 后 OnStart 写入候选库再失败：Discard，Active 的数据与配置不变，旧版重启正常；新旧进程没有同时写同一个库（runtime.lock 证明）。
6) 两个入口：同一插件同时用 sqlitedb.Open 与 gormdb.Open 时共用同一连接，一边写入另一边立即可读；只用 sqlitedb 的示例变体不链接 gorm.io/gorm。
7) i18n：Context.T 的语言选择与回退；plugin check 拒绝缺少 en-US 与占位符不一致。
8) 表单（web/admin 的 vitest，非 DOM）：configdb 的 Schema 字段顺序与声明一致；hidden 不进 patch；未修改的 secret 不进 patch；i18 标签回退。
- 证据：TEST-MATRIX §3.2 追加 T-08 一行；需要 T-09 正式 reload、WebUI/API 与 Journal 崩溃恢复才能关闭的 V-23～V-25、V-38 场景单独列为「未闭合」，不标成通过。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t08_e2e_test.go；web/admin 的 schema.test.ts 追加用例。
- 本地自测（不提交，PR 中写明结果）：npm run build；示例的 plugin check。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；web/admin 下 npm run test 通过；node scripts/check-repository.mjs。

不做：正式 reload 与管理入口（T-09）；Journal 与崩溃恢复（T-09.5～T-09.7）。
```

### goal-38 · T-09.3 开发 CLI 与库模式脚手架

状态：等待前置（T-04.5、T-08.1、T-08.8、T-09.2、T-13.1）。语句中 goal-11、goal-28、goal-37、goal-53、goal-67 提供的符号（sdk/generate、packages 的 Preflight 与包格式、contracts/release、contracts/i18n、根包 Run）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-38（切片 T-09.3：开发 CLI 与库模式脚手架）。goal 目录 goal-38，分支 goal-38，worktree .worktrees/goal-38，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：完成开发命令 alonabot plugin init|check|build|test|pack：init 生成插件骨架与库模式调试用的 main.go；check 验证 CONTRACTS §9.1 的静态契约；build 用生成器为根插件生成调用 sdk.Run 的正式入口并按目标 OS/arch 编译，产物不含 Core；test 在临时数据与离线测试会话中检查加载与注册；pack 为四种角色的组件生成 tar.gz 安装包与归档外的 release.json（WebUI 只含 plugin.yaml、icon 与 assets 目录，平台为 any，不运行 build），并可用本地密钥对清单的确切字节签名。

依据：PLAN T-09 步骤 2（CLI）；CONTRACTS §9.1～§9.3、§15.4、§7.5；DECISIONS D-42、D-43、D-59、D-62、D-63；TEST-MATRIX V-01、V-41、V-54 的打包分项。

范围（只新增或修改这些）：
- cmd/alonabot/：plugin 命令组增加 init、build、test、pack，check 补齐静态契约（新文件）
- 新增 sdk/generate/entry.go（正式入口生成）
- 新增 internal/pluginkit/（打包、release.json 生成与签名，供 cmd 使用）；packages 的包格式与 goal-37 一致
- 上述包的测试文件
允许的新跨模块边：cmd/alonabot -> internal/pluginkit；internal/pluginkit -> contracts/...、-> sdk/generate（ARCHITECTURE 已登记）。

接口：
- alonabot plugin init <dir> --id ID [--role plugin|adapter] [--adapter-id ID]：生成 plugin.yaml（manifest_version 1、version 0.1.0、source 为占位并提示修改）、带 sdk.Base 的根插件、一个示例命令、i18n/en-US.json、main.go（调用 alonabot.Run(alonabot.Plugin(...))，供库模式调试）与 .gitignore（忽略 .alonabot/，其中的实例配置可能含平台凭证）；目录非空时拒绝。
- alonabot plugin check [--dir] [--goos] [--goarch] [--tags]：在 goal-11/goal-28/goal-53 的检查之外，核对 manifest（Parse 与 §9.1 的字段规则、role 与描述一致、migration_mode 与实际迁移接口一致、prepare_effects 已声明）、Schema 与 config_version、i18 文件；全部问题一次列出。
- alonabot plugin build [--dir] --goos OS --goarch ARCH [--out DIR]：生成正式入口（临时目录中的 main 包，调用 sdk.Run 或 sdk.RunAdapter，引用根类型），用锁定的 Go Modules 编译（CGO_ENABLED=0，-trimpath），输出可执行文件；不链接根包 alonabot。
- alonabot plugin test [--dir]：build 当前平台后，在临时 data_dir 中经 Core 的开发入口启动一次，完成握手、准备、注册与激活后停止，报告注册的路由；不连接任何真实平台，不读取生产凭证。
- alonabot plugin pack [--dir] --goos OS --goarch ARCH [--key FILE] [--out DIR]：build 后打包 tar.gz（plugin.yaml、entrypoint、config.schema.json、i18n/、build-info，以及 manifest 的 icon 与表单选项图标）；role 为 webui 时不需要 --goos/--goarch，只打包 plugin.yaml、icon 与 assets 目录，归档名为 <id>-<version>-any.tar.gz；生成 release.json（id、version、source、os/arch、归档 SHA256 与长度、发布者公钥 ID）；给出 --key（Ed25519 私钥文件）时对 release.json 的确切字节签名生成 release.json.sig，否则只输出待签名的 release.json。私钥不进入包。
internal/pluginkit 的导出函数由执行者按上述命令拆分，并在包文档列出。

行为：
- build 与 pack 的产物对同一源码与工具链可重复（tar 中的文件按名称排序、时间戳固定、权限固定）。
- pack 前自动运行 check；有问题时不打包。
- 签名：release.json 的签名对象是文件的确切字节（与 contracts/release.Verify 一致）；pack 输出的包必须能通过 goal-37 的 packages.Preflight（以测试中注入的公钥为受信来源）。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/plugin_cli_e2e_test.go（//go:build unix）：在临时目录 init 一个插件 → check 通过 → build 当前平台 → test 报告注册的命令 → pack 带测试密钥 → packages.Preflight 校验通过并 MoveTo 成功；改动 manifest 使 migration_mode 与实际不符时 check 失败；init 的 main.go 能以库模式编译；pack 两次的归档字节相同。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：五个目标平台的 build（交叉编译，只编译不运行）；Adapter 与 Agent 角色的 init 与 build；WebUI 包的 pack（缺少 assets/index.html 时 check 失败）；目录非空时 init 拒绝；没有 --key 时只输出 release.json。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./cmd/... ./sdk/... ./internal/...；go build ./...。

不做：安装与管理（T-09.4）；市场与远程发布；发行归档（T-14.3）。
```

### goal-55 · T-11.2 工具调用与用户确认

状态：等待前置（T-03.8、T-06.2、T-08.8、T-11.1）。语句中 goal-8、goal-20、goal-53、goal-54 提供的符号（buttons.Service、broker 的父链调用、internal/core/i18n、tool.Catalog）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-55（切片 T-11.2：工具调用与用户确认）。goal 目录 goal-55，分支 goal-55，worktree .worktrees/goal-55，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：core.tool.invoke 校验父调用链后按生效策略分支：deny 直接拒绝；confirm 先经用户确认；allow 或确认通过后经 Broker 以父调用链调用目标 Service，结果与错误原样返回。core.tool.confirm 只执行策略与确认。确认由 Core 直接经 Adapter 向默认 Target 发送带「允许」「拒绝」按钮的消息（不执行插件出站 Hook），只接受同一 Actor 在 TOOL_CONFIRM_TIMEOUT 内的第一次点击。

依据：PLAN T-11 步骤 3、4；CONTRACTS §13.1、§13.2、§5.4；MESSAGE-MODEL §6.3；WIRE-FORMAT §3.5；TEST-MATRIX V-55。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/tool.proto：追加 ToolInvokeRequest/Response 与 ToolConfirmRequest/Response。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceToolInvoke = "core.tool.invoke"、ServiceToolConfirm = "core.tool.confirm"
- internal/core/tool/：新增 invoke.go、confirm.go
- internal/core/buttons/：确认记录绑定的 token（与插件按钮共用表，增加确认类型）
- internal/core/i18n/：en-US.json 与 zh-CN.json 增加确认消息的文案
- sdk/：tool.go 增加 InvokeTool 与 ConfirmOperation
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message ToolInvokeRequest { string name = 1; string arguments_json = 2; } // name 为 "plugin_id/service_name"
   message ToolInvokeResponse { string result_json = 1; }
   message ToolConfirmRequest { string operation = 1; string summary = 2; } // 例如 operation "mcp/<server>/<tool>"
   message ToolConfirmResponse {}
2. sdk
   func InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) // 必须在 Handler 或其延续调用内调用
   func ConfirmOperation(ctx context.Context, operation, summary string) error // nil 表示用户允许

行为：
- 两个服务都必须带 parent_call_id（指向本 Session 在途的 Core→插件调用），否则 FAILED_PRECONDITION。
- invoke：
  1) name 必须是 core.tool.list 对调用者可见的工具（访问规则通过、生效策略不是 deny），否则 deny → PERMISSION_DENIED、不可见 → NOT_FOUND。arguments_json 必须是 JSON 对象，并通过工具的 input schema（jsonschema-go Validate）；否则 INVALID_ARGUMENT。
  2) 策略 confirm：执行确认（见下），拒绝或超时时返回相应错误，不调用工具。
  3) 经 Broker 以父调用链调用目标 Service（父链、Actor 与 send_forbidden 继承，工具仍按继承的 Actor 检查 Service 访问规则），结果原样放进 result_json，错误原样返回。不重试。
  每次调用都读取当前策略，修改在下一次调用时生效。
- confirm：operation 非空、summary 不超过 1024 字节（命名常量）；直接执行确认，不调用任何工具。
- 确认：
  - 父调用链没有可信 Actor 或默认 Target → PERMISSION_DENIED，不发消息。
  - 目标实例不支持 buttons → UNSUPPORTED，不改为文字确认。
  - Core 创建确认记录（随机 ID、Actor、Target、父调用、到期时间）与两个绑定它的按钮 token（沿用按钮 token 的格式与表，类型为 confirm，不绑定插件路由），按 Actor locale 选择 Core 文案：工具名或操作名；invoke 展示实际的 arguments_json（缩进格式化，超过 2048 字节时截断并注明，命名常量），confirm 展示插件给出的 summary。消息直接经 adapter.Instance.Send 发往默认 Target，不执行出站 Hook；发送失败时返回该错误，不执行工具。
  - 点击（button.clicked 解析到 confirm 类型的 token）：只接受该记录的 Actor 的第一次点击；其他用户的点击被拒绝（不影响结果，以 INFO 记录）；token 只能使用一次；「允许」使调用继续，「拒绝」返回 PERMISSION_DENIED。点击后 Core 把原消息编辑为结果文案（目标支持 message.edit 时；不支持时不编辑）。
  - 超时（TOOL_CONFIRM_TIMEOUT）→ DEADLINE_EXCEEDED；父调用结束或 Session 失效 → 记录作废，token 失效。确认只对本次调用有效。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/tool_invoke_test.go（//go:build unix）。沿用 goal-54 的工具服务与测试插件（收到消息时调用 InvokeTool 或 ConfirmOperation 并回复结果），memory Adapter 支持 buttons 与 Editor：
  - allow 工具直接返回结果；deny 工具返回 PERMISSION_DENIED 且工具未执行。
  - confirm 工具：memory 收到带两个按钮的确认消息，内容含实际 JSON 参数；同一用户点「允许」后工具执行并返回结果；另一用户先点击被拒、不影响结果；点「拒绝」返回 PERMISSION_DENIED 且工具未执行；超时（注入较短的 TOOL_CONFIRM_TIMEOUT）返回 DEADLINE_EXCEEDED。
  - 确认消息没有经过插件的 BeforeSend Hook（测试插件注册了会改链的 BeforeSend，确认消息不受影响）。
  - 没有 Actor 的调用（scheduled fixture）遇到 confirm 返回 PERMISSION_DENIED，没有发消息；不支持 buttons 的实例返回 UNSUPPORTED。
  - arguments 不符合 input schema 时 INVALID_ARGUMENT。
  - ConfirmOperation 展示 summary 而不是参数。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：参数截断；token 只能用一次；父调用结束时记录作废；zh-CN 文案；策略在两次调用之间修改。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：Agent 的工具循环与 MCP（T-12.3）；「总是允许」（契约不提供）；管理页面的工具策略（T-09.9）。
```

### goal-70 · T-13.4 库模式示例与验收

状态：等待前置（T-03.10、T-13.2、T-13.3）。语句中 goal-9、goal-67～goal-69 提供的符号与示例（根包 Run、Reload、examples/devbot、回环 Adapter）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-70（切片 T-13.4：库模式示例与验收）。goal 目录 goal-70，分支 goal-70，worktree .worktrees/goal-70，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：examples/devbot/ 用库模式运行 echo 插件与回环测试 Adapter，覆盖同进程运行、热部署与编译失败；docs/development.md 补充库模式与热部署的用法；在 TEST-MATRIX §3.2 记录 V-64、V-65 的证据。

依据：PLAN T-13 步骤 5 与「停止和失败条件」；CONTRACTS §15.4；TEST-MATRIX V-64、V-65。

范围（只新增或修改这些）：
- examples/devbot/：改为同时引用 examples/plugins/echo 的插件与示例内的回环 Adapter，提供 Reload 开关
- docs/development.md：新增「库模式与热部署」一节（本切片获授权修改这一节）
- 新增 t13_e2e_test.go（根包，//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-13 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

验收矩阵（各为一个子测试）：
1) 同进程：插件与 Adapter 的握手、注册、权限与独立进程一致（同一组命令的回复与 goal-14 的结果相同）；从 instances/ 读取实例配置。
2) dev 代次固定，重启后旧按钮仍有效。
3) dev 来源 update/rollback 被拒。
4) 未捕获的 panic 使进程退出且不重启。
5) 停止超时时 Run 返回错误。
6) 已安装包（独立进程）与开发组件混合运行。
7) 两次启动之间 config.json 与 dev.sqlite 的变化生效。
8) 热部署：改动插件源码后只重启开发组件，Adapter 连接与其他插件不中断；编译失败时旧组件继续；go.mod 变化输出 WARN；其他 module 的组件不重启。
- docs/development.md 的新节写明：Run 的选项、数据目录、实例目录、热部署的启用方式与限制（没有进程隔离、不能 update/rollback、Windows 未验证等），以及运行示例的命令（fish 与 bash 通用的写法）。
- 证据：TEST-MATRIX §3.2 追加 T-13 一行，写明测试位置、命令、平台与结果；Windows 与 Linux 未运行的项写「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：t13_e2e_test.go（上述矩阵，复用 goal-67～goal-69 的测试方式）。
- 本地自测（不提交，PR 中写明结果）：按文档中的命令手动运行一次示例并记录输出。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：发行与预装（T-14）。
```

### goal-81 · T-16.1 插件测试套件基础

状态：等待前置（T-05.2、T-13.2）。语句中 goal-4、goal-16、goal-67、goal-68 提供的符号（sdk/adapter 的 Adapter 与 Instance、Router 的分发流程、alonabot.Start 与 Core、Plugin/Adapter/DataDir 选项、dev 的 config.json 与 dev.sqlite、生成器的 Generated）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-81（切片 T-16.1：插件测试套件基础）。goal 目录 goal-81，分支 goal-81，worktree .worktrees/goal-81，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：新建公开包 plugintest：在 go test 中用 alonabot.Start 在临时数据目录上装配真实 Core，登记被测插件与套件自带的测试 Adapter；提供会话与用户构造、Core Admin、发信许可、Capability 子集与配置的选项；注入一条消息并在这个事件的全部分发链结束后返回期间产生的 Send、Edit 与 Handler 错误。分发结束由 Core 经 internal/inproc 通知，不靠等待时间。

依据：PLAN T-16 步骤 1、2 与「失败停止条件」；CONTRACTS §17、§15.4、§8（Target 准入）、§2；DECISIONS D-47；ARCHITECTURE 的 plugintest 一行；TEST-MATRIX V-73 的基础分项。

范围（只新增或修改这些）：
- 新增 plugintest/（包 plugintest）
- 新增 plugintest/internal/probe/（测试用插件，含 go generate 生成的注册文件）
- internal/inproc/：新增 harness.go
- 根包：Start 读取 internal/inproc 中登记给该数据目录的 Harness，在启动组件前写入其中的 Core Admin 与发信许可，并把 Router 的分发结束回调接到 Harness（只加调用点）
- internal/core/router/：Options 增加分发结束回调字段，在每个事件的分发结束时调用（只加字段与调用点）
- 上述包的测试文件
允许的新跨模块边：plugintest -> 根包、-> sdk/...、-> internal/inproc、-> contracts（ARCHITECTURE 已登记）。internal/inproc 仍不 import 本 module 的任何包。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/inproc（只用标准库类型）
   type Admin struct { InstanceID, UserID string }
   type Outcome struct { PluginID, RouteID string; ErrorCode int32; SafeMessage string } // ErrorCode 为 0 表示 Handler 成功
   type Dispatch struct { InstanceID, SourceEventID string; Outcomes []Outcome }
   type Harness struct {
       Admins           []Admin
       AllowedInstances map[string][]string // plugin_id → 授予的 adapter_instance_id
       Dispatched       func(Dispatch)      // Core 在一个事件的全部分发链结束、且其自动回复都已发送后调用
   }
   func RegisterHarness(dataDir string, h *Harness) (unregister func()) // dataDir 为绝对路径；同一目录重复登记 panic
   func HarnessFor(dataDir string) *Harness                             // 没有登记时返回 nil
2. 根包：不新增公开符号。Start 在取得 core.lock、创建开发组件的安装记录之后、激活组件之前，按 HarnessFor(DataDir) 写入 Admins 与 AllowedInstances（经 internal/core 已有的策略与许可写入方法），并设置 Router 的分发结束回调。HarnessFor 返回 nil 时行为与之前完全相同。
3. plugintest
   const AdapterID = "plugintest"
   const InstanceID = "test"
   const AccountID = "bot"
   type Option func(*options)
   func Plugin(root sdk.Root) Option                     // 可以多次使用
   func Config(root sdk.Root, configJSON string) Option  // 写入该插件的 config.json
   func Admin(userID string) Option                      // 加入 Core Admin（测试实例上的该用户）
   func Allow(root sdk.Root) Option                      // 授予该插件对测试实例的 allowed_adapter_instances
   func Capabilities(names ...string) Option             // 替换测试实例声明的能力
   type Chat struct{ ... }
   func Private(userID string) Chat                      // chat_id 为 "private:<userID>"
   func Group(chatID string) Chat                        // chat_id 为 "group:<chatID>"
   type User struct { ID, Name string; ChatAdmin bool }
   type Sent struct { Op string; Target *pb.Target; Chain *pb.MessageChain; EditedMessageID string; Result *pb.SendResult } // Op 为 "send" 或 "edit"
   func (s Sent) Text() string                           // 依次拼接 Chain 中 text 片段的文字
   type Outcome struct { PluginID, RouteID string; Err *pb.ErrorInfo } // Err 为 nil 表示 Handler 成功
   type Result struct { Sends []Sent; Outcomes []Outcome }
   type Harness struct{ ... }
   func Start(t testing.TB, opts ...Option) *Harness
   func (h *Harness) Message(chat Chat, from User, chain message.Chain) Result
   func (h *Harness) Text(chat Chat, from User, text string) Result // 等于 Message(chat, from, message.Chain{message.Text(text)})

行为：
- Start：
  1) 数据目录为 t.TempDir()；写入 adapters/plugintest/instances/test.json（内容 "{}"）与每个 Config 的 plugins/<id>/config.json（插件 ID 取自根类型的 Generated manifest；没有 Generated 时 t.Fatal）。
  2) 登记 inproc.Harness（Admins、AllowedInstances 与 Dispatched），调用 alonabot.Start(ctx, DataDir, 全部 Plugin, Adapter(测试 Adapter))；返回错误时 t.Fatal。Start 返回时测试实例已启动。
  3) t.Cleanup：取消 ctx、Wait，返回错误时 t.Error；注销 Harness。
- 测试 Adapter（套件内的 sdk/adapter 实现，与被测插件一样经 loopback 走完整生命周期）：
  - 实例账号固定为 AccountID。缺省能力为 message.send、message.markdown、message.album、message.reference、message.buttons、message.forward、message.edit（version 1）；Capabilities 选项替换这一列表。
  - Send：markdown 的 text、reference、buttons、forward 片段所需的能力（依次为 message.markdown、message.reference、message.buttons、message.forward）未声明时返回 REJECTED 与 UNSUPPORTED，不记录为已发送；否则 ACCEPTED，一项 PartResult 覆盖全部片段，消息 ID 从 1 起按顺序递增（十进制字符串）。Edit 同样记录（Op "edit"）。每次 Send 与 Edit 都追加到套件的发送记录。
- Message：
  - 构造 message.created：Target 为 (AdapterID, InstanceID, AccountID, chat_id)；Actor 为 from（ChatAdmin 时 trusted_roles 含 CHAT_ADMIN）；platform_message_id 与 source_event_id 为套件内递增的唯一值；to_me 在私聊或 chain 中 mention 了 AccountID 时为 true。
  - 经测试实例交付；Core 不接纳时 t.Fatal。然后等待 Dispatched 中 SourceEventID 与之相同的通知，返回从交付到通知之间追加的全部发送记录与 Outcomes。等待不设套件自己的超时，以测试的期限为准；Message 不能并发调用（同一 Harness 上并发调用时 t.Fatal）。
- 套件不跳过 Core 的任何检查：命令权限、Rule、Target 准入、Hook 与文案都与正式运行相同。被测插件是库模式的开发组件，配置与 dev.sqlite 按 goal-68 的规则处理。

测试（按编码规范「测试提交规则」）：
- 提交：plugintest/plugintest_test.go，用 plugintest/internal/probe（/ping 回复 pong；/admin 只允许 Core Admin；chat_admin Rule 的 /mod；发送带 buttons 的消息，收到 UNSUPPORTED 时改发纯文本；/config 回复配置中的值；/count 在 dev.sqlite 中计数并回复；/edit 先发送再编辑）：
  - /ping 的结果恰好一条 Sent，Text() 为 "pong"，Target 为原会话。
  - /admin：普通用户被拒（没有 Sent，命令被消费），Admin 选项给出的用户成功。/mod：ChatAdmin 为 false 时没有回复，为 true 时回复。
  - 缺省能力下按钮消息 ACCEPTED；Capabilities 去掉 message.buttons 后，结果中第一条为 REJECTED/UNSUPPORTED、第二条为纯文本。
  - Config 选项的值经 /config 返回；/count 连续两次回复 1 与 2。
  - /edit 产生 Op 为 send 与 edit 的两条记录；去掉 message.edit 后 edit 被 Core 以 UNSUPPORTED 拒绝。
  - 两个使用 t.Parallel 的测试各自 Start，计数互不影响。
  - 用 go test -race -count=2 ./plugintest/... 运行两遍都通过。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有 Generated 的根类型；同一 Harness 并发 Message；Start 失败（probe 的 OnStart 返回错误）后数据目录可以再次 Start；Dispatched 只在自动回复发送之后调用。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./plugintest/... ./internal/core/router/... ./internal/inproc/...；go test -race . ；go build ./...。

不做：按钮点击、多轮等待、schedule 触发与后台发送（T-16.2）；用套件测试示例插件与开发文档（T-16.2）；WebUI 沙盒（v0.1 之外）。
```

### goal-83 · T-12.8 文本生成服务

状态：等待前置（T-03.8、T-06.1、T-12.2）。语句中 goal-8、goal-19、goal-59 提供的符号（manifest 的 role 解析、Service Broker 的目标定位、sdk.JSONService、sdk.CallJSON、sdk.WithAccess、Context.Caller、agent 的 Config 与 Provider 装配）只写名字；合并后由审查者核对，再改为可派发。2026-10-28 按 D-61 改为以 `@agent` 为目标。

```text
/goal 使用 $alonabot-execute 推进 goal-83（切片 T-12.8：文本生成服务）。goal 目录 goal-83，分支 goal-83，worktree .worktrees/goal-83，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 的 Service Broker 把目标 `@agent` 解析为已安装的 Agent 组件；新建 sdk/llm，提供 CONTRACTS §13.5 llm.generate 的请求/响应类型、Schema 标识与以 `@agent` 为目标的调用辅助；官方 Agent 组件注册该服务（public），只接受配置项 llm_service_allowed_plugins 中的直接调用方，用已配置的 Provider 完成一次非流式请求，并返回 Provider 给出的用量。

依据：PLAN T-12 步骤 9；CONTRACTS §13.5、§3、§9.1；DECISIONS D-38、D-51、D-61；TEST-MATRIX V-76、V-82。

范围（只新增或修改这些）：
- internal/core/ 的 Service Broker：目标定位处解析 `@agent`（只改定位处，新增解析函数放在新文件）
- sdk/llm/：新建包（类型、Schema 标识、Generate）
- agent/：Config 增加字段；新增 generate.go（服务注册与处理）；provider 包不改
- 上述包的测试文件
允许的新跨模块边：无（agent -> sdk/... 已允许）。sdk/llm 只 import sdk 与标准库。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/llm
   const ServiceName = "llm.generate"
   const AgentTarget = "@agent" // CONTRACTS §3：由 Core 解析为已安装的 Agent 组件
   const SchemaIn, SchemaOut = "...", "..." // 随包固定的 schema 标识，按 goal-19 的命名规则选定并写入包文档
   const SchemaVersion uint32 = 1
   type Message struct { Role string `json:"role"`; Text string `json:"text"` } // Role 为 "user" 或 "assistant"
   type Request struct { Messages []Message `json:"messages"`; System string `json:"system,omitempty"`; Provider string `json:"provider,omitempty"`; MaxOutputTokens int `json:"max_output_tokens,omitempty"` }
   type Response struct { Text string `json:"text"`; Provider string `json:"provider"`; Model string `json:"model"`; InputTokens int `json:"input_tokens,omitempty"`; OutputTokens int `json:"output_tokens,omitempty"` }
   func Generate(ctx context.Context, req Request) (Response, error) // 经 sdk.CallJSON 调用 AgentTarget + "/" + ServiceName
2. agent
   Config 增加字段：LLMServiceAllowedPlugins []string `json:"llm_service_allowed_plugins"`

行为：
- `@agent`：Broker 定位目标时，`@agent` 解析为当前安装记录（或库模式的 dev 组件）中 role 为 agent 的组件 ID，之后的访问规则、调用链、CALL_CYCLE 与激活检查按解析后的 ID 执行；没有 Agent 组件时与目标未激活相同。
- 注册：在 Services Hook 中以 JSONService 注册 ServiceName，WithAccess(Public)。
- 许可：Context.Caller() 为空或不在 LLMServiceAllowedPlugins 中时返回 PERMISSION_DENIED，不请求 Provider。
- 校验：messages 为空、role 不是 user 或 assistant、max_output_tokens 小于 0 时返回 INVALID_ARGUMENT；provider 非空且不存在时返回 INVALID_ARGUMENT；为空时使用 default_provider，没有任何可用 Provider 时返回 FAILED_PRECONDITION。
- 请求：把 system 与 messages 组装成 provider.Request，MaxTokens 取 max_output_tokens（为 0 时取该 Provider 配置的 max_tokens），不传工具；用 Provider 客户端的流式接口接收并拼接全部文本，收到终止标记才算成功；用量取自 provider.Result.Usage，为 nil 时不填写。
- 错误：Provider 返回错误或中途断开时返回 UNAVAILABLE，safe_message 不含 API Key 与消息内容（provider 返回的错误说明只写日志，不放进 safe_message）；不自动重试（D-38）。调用方期限到达时取消 Provider 请求。
- 不读写会话存储与记忆，不套用人格，不调用工具。
- 日志：每次调用以 INFO 记录调用方插件、provider、model、耗时与用量；不记录消息文本。

测试（按编码规范「测试提交规则」）：
- 提交：agent/generate_test.go（//go:build unix）。用 goal-59 的测试装配、本地模拟 Provider 与一个调用 llm.Generate 的测试插件：
  - 调用方未列入时 PERMISSION_DENIED，模拟 Provider 没有收到请求；列入后返回 text、provider、model 与模拟服务给出的用量，会话存储没有新增条目。
  - 测试插件只经 llm.Generate（`@agent`）调用，不知道 Agent 的组件 ID；把 Agent 组件换成另一个 ID 的测试 Agent 后同一调用到达新组件。
  - provider 不存在为 INVALID_ARGUMENT；Agent 配置中没有 Provider 为 FAILED_PRECONDITION。
  - 模拟 Provider 出错与中途断开为 UNAVAILABLE，模拟服务只收到一次请求，safe_message 不含 API Key。
  - 调用方期限短于模拟 Provider 的延迟时，调用方得到 DEADLINE_EXCEEDED，模拟服务看到请求被取消。
  - Agent 组件未启用或未安装时调用返回 UNAVAILABLE。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：校验分支；三种协议各一次。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... ./sdk/... ./internal/core/...；go build ./...；node scripts/check-go-deps.mjs 通过。

不做：流式输出与工具调用（契约不提供）；llm_service_allowed_plugins 的专用管理页（配置表单随 T-09 由 Schema 生成）。
```

## 第 11 波

### goal-40 · T-09.5 候选 Journal 与准备编排

状态：等待前置（T-09.4、T-08.6、T-06.4）。语句中 goal-22、goal-33、goal-39 提供的符号（manage.Service 与 Candidates 接口、candidate.PrepareDatabase/Discard 与交接约定、supervisor 的停止流程）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-40（切片 T-09.5：候选 Journal 与准备编排）。goal 目录 goal-40，分支 goal-40，worktree .worktrees/goal-40，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：为 reload、update 与首次启用实现正式的候选编排：预检通过后，以预期的旧 ActiveGeneration 写入持久化的管理操作记录（Journal）及其资源引用；冻结新根派发、drain 并停止旧进程，确认进程树退出且 runtime.lock 可取；调用 T-08.6 复制数据库并创建候选 Session（启动原因、维护许可与阶段预算放进 Bootstrap）；REGISTER 校验后持久化程序、配置快照、secret 与数据库的候选引用，形成 PREPARED。提交与激活门禁留给 T-09.6；失败时清理候选并以旧版的新 Session 恢复。

依据：PLAN T-09 步骤 4；CONTRACTS §12、§6.3、§5.2、§4.2；TEST-MATRIX V-29。

范围（只新增或修改这些）：
- internal/core/state/：migrations 末尾追加一条；新增 journal.go
- 新增 internal/core/journal/（操作记录的状态机与编排，实现 manage.Candidates 的 Reload 与 Update）
- internal/core/manage/：首次启用改用 journal 编排，删除 goal-39 的 shortcut 注释（只改该处与装配）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：CREATE TABLE journal (op_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES installs(plugin_id), kind TEXT NOT NULL, expected_old TEXT NOT NULL, new_generation TEXT NOT NULL, phase TEXT NOT NULL, package_dir TEXT NOT NULL, database_path TEXT NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL);
   kind 为 first_install、reload、update；phase 为 started、stopped_old、candidate_ready、prepared、committed、failed、cleaned。读写方法自定，阶段只能前进。
2. internal/core/journal
   type Orchestrator struct{ ... }
   func New(opt Options) *Orchestrator // Options 由执行者按实际依赖列出：Store、Secrets、supervisor 启动函数、Router/Broker 的冻结入口、Log
   type Result struct { OpID string; Generation string; Prepared *candidate.Prepared }
   func (o *Orchestrator) Prepare(ctx context.Context, pluginID string, kind Kind, stagedPackage string) (*Result, error) // 运行到 PREPARED 为止
   func (o *Orchestrator) Abort(ctx context.Context, r *Result, cause error) error // 清理候选并以旧版的新 Session 恢复
   func (o *Orchestrator) Reload(ctx context.Context, id string) error   // 实现 manage.Candidates：Prepare → 交给提交入口（T-09.6 之前提交入口为 UNSUPPORTED 时 Abort 并返回）
   func (o *Orchestrator) Update(ctx context.Context, id, staged string) error

行为（Prepare，调用方已持插件操作锁）：
1) 预检：update 的包已经过 packages.Preflight（staged 目录）；reload 用 config.Service.ImportSnapshot 导入编辑快照为 Pending（base 为当前 Active）；first_install 没有旧 Active。所有静态错误在停止旧版之前返回。
2) 写入 Journal（phase=started，expected_old=当前 Active，new_generation 为新 ID）。
3) 冻结：Router/Broker/Bus/Scheduler/HTTP 停止向该插件派发新的根调用（沿用 goal-22 与 goal-27 的 Drain），按停止流程停止旧进程并等待进程树退出；phase=stopped_old。进程树没有在停止预算内清空时失败（不 sleep 猜测）。
4) candidate.PrepareDatabase（取得 runtime.lock 后复制，完成后释放）；AddGeneration(new, 未提交)；phase=candidate_ready。
5) 启动候选：supervisor.Launch（NewCandidate=true，Bootstrap 带 activation_reason（first_install/reload/update）、manual_migration_approved（只有管理员为本次 Session/Generation 显式授予时为 true）、阶段预算、Pending 解析后的配置），REGISTER 校验与 secret 抽取后得到快照；SetGenerationConfig 写入快照与 secret 路径；phase=prepared。此时不提交 Active、不发 ACTIVATE。
6) manifest 声明 prepare_effects=external 的插件，已有安装的 reload 与 update 在步骤 2 之前检查：没有本次操作的维护许可时返回 FAILED_PRECONDITION（safe_message "NEEDS_MANUAL_MIGRATION"），不写 Journal、不停止旧版；有许可时 Bootstrap 的 manual_migration_approved=true，其余照常。first_install 不受此门槛限制。
- 普通 resume（restart、Core 重启后的恢复）不经过本编排，不写新 Journal。
- Abort：phase 改为 failed → 停止候选进程并确认退出 → candidate.Discard（只删未受保护的候选资源）→ 以旧 Active 启动旧版的新 Session（first_install 时保持未启用）→ phase=cleaned。Abort 本身失败时保留 Journal 供 T-09.7 恢复。
- 任一步失败都调用 Abort；错误返回给管理服务（安全文本）。
- 日志：每次 phase 变化以 INFO 记录（op_id、plugin_id、kind、phase）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/journal_test.go（//go:build unix）。测试插件按环境变量在 OnStart 写库后失败、或迁移失败：
  - reload 成功运行到 PREPARED：Journal phase 依次前进，候选数据库是独立副本，旧进程已退出，Active 未变。
  - 候选 OnStart 失败：Abort 后旧版以新 Session 运行，旧库不含候选写入，候选目录已删除，phase=cleaned。
  - 旧进程不退出（忽略 STOPPING 且持有 runtime.lock 的测试插件）：编排在停止预算内失败，没有复制数据库，旧版状态可诊断。
  - 磁盘空间不足（用测试钩子使备份失败）：不激活，候选目录被清理。
  - first_install：成功到 PREPARED；失败时插件保持未启用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：phase 只能前进；expected_old 与写入一致；manual_migration_approved 只在显式授予时为 true；resume 不写 Journal。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：唯一提交点与激活门禁（T-09.6）；崩溃恢复与 rollback（T-09.7）；维护许可的管理入口（T-09.9）。
```

### goal-57 · T-11.4 工具与会话的验收

状态：等待前置（T-11.2、T-11.3）。语句中 goal-54～goal-56 提供的符号与测试插件只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-57（切片 T-11.4：工具与会话的验收）。goal 目录 goal-57，分支 goal-57，worktree .worktrees/goal-57，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：用一个提供工具的插件和一个调用工具的测试插件，经回环测试 Adapter（或 Telegram 模拟服务）以真实进程覆盖三种策略、确认的允许/拒绝/超时/他人点击、无 Actor、目标不支持按钮与策略的运行时修改；会话存储覆盖原子追加、分页、跨插件隔离与 Core 重启；在 TEST-MATRIX §3.2 记录 T-11 证据。

依据：PLAN T-11 步骤 6 与「停止和失败条件」；CONTRACTS §13；TEST-MATRIX V-55、V-56 的 T-11 分项。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t11_e2e_test.go（//go:build unix）
- 复用或补充 internal/core/supervisor/testdata 下的工具与会话测试插件
- spec/TEST-MATRIX.md §3.2：追加 T-11 的证据行（本切片获授权修改这一处）
允许的新跨模块边：无。

验收矩阵（各为一个子测试）：
1) allow、confirm、deny 三种策略各一次；deny 与拒绝返回 PERMISSION_DENIED 且工具未执行（以工具插件的调用计数证明）。
2) 确认：允许、拒绝、超时（DEADLINE_EXCEEDED）、他人先点击（被拒、不影响结果）、token 第二次点击无效；确认消息展示实际 JSON 参数。
3) 无 Actor 的后台调用遇到 confirm 直接 PERMISSION_DENIED，不发消息；目标不支持按钮时 UNSUPPORTED。
4) core.tool.confirm 用于 MCP 风格的操作名（mcp/x/y），展示插件给出的摘要。
5) 运行时修改策略：下一次调用生效，进程未重启。
6) 工具仍按继承 Actor 检查 Service 访问规则。
7) 会话：多条原子追加、跨帧上限分页不丢不重、两个插件同名会话互相隔离、Core 重启后保留、purge 后删除（卸载默认保留由 T-09.4 的测试覆盖，这里只覆盖服务层）。
- 证据：TEST-MATRIX §3.2 追加 T-11 一行，写明测试位置、命令、平台与结果；工具策略与会话清除的管理页面留给 T-09.9，记为未闭合。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t11_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：在 -race 下重复运行矩阵若干次无偶发失败。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：Agent 组件的工具循环（T-12.3）；管理页面（T-09.9）。
```

### goal-60 · T-12.3 工具循环与 MCP

状态：等待前置（T-11.2、T-12.2）。语句中 goal-54、goal-55、goal-58、goal-59 提供的符号（sdk.ListTools、sdk.InvokeTool、sdk.ConfirmOperation、provider 的 Tool/ToolCall/ToolResult、agent 的一轮对话与 Config.MCPServers）只写名字；合并后由审查者核对，再改为可派发。本切片获授权引入已批准的 github.com/modelcontextprotocol/go-sdk（docs/development.md 只授权给 goal-60）。

```text
/goal 使用 $alonabot-execute 推进 goal-60（切片 T-12.3：工具循环与 MCP）。goal 目录 goal-60，分支 goal-60，worktree .worktrees/goal-60，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Agent 组件每轮把 core.tool.list 返回的工具与已连接 MCP 服务器的工具交给模型；工具名按 Provider 的命名规则映射且一轮内稳定；Core 工具经 core.tool.invoke 调用，MCP 工具先经 core.tool.confirm（操作名 mcp/<server>/<tool>）再调用；拒绝、超时与工具错误作为工具结果交回模型；单轮工具步数达到上限时结束本轮并告知用户；MCP 服务器首次使用时连接、随插件停止关闭，stdio 子进程属于插件的进程树。

依据：PLAN T-12 步骤 4；CONTRACTS §13.1、§13.2、§7.1（进程树）；DECISIONS D-27、D-38；docs/development.md（MCP go-sdk 的批准范围）；TEST-MATRIX V-58。

范围（只新增或修改这些）：
- go.mod、go.sum：新增 github.com/modelcontextprotocol/go-sdk（取当前最新的 v1 版本，PR 中写明版本，并用 go mod why 说明它只被 agent 使用）
- agent/：新增 tools.go（工具目录与名称映射）、mcp.go（MCP 客户端管理）、loop.go（工具循环）；一轮对话接入循环（只改调用处）
- 新增 agent/mcptest/（stdio 与 streamable HTTP 的 MCP 模拟服务器，测试用）
- i18n：en-US.json 与 zh-CN.json 增加步数上限与工具错误的文案
- 上述包的测试文件
允许的新跨模块边：只有 agent 可以 import MCP go-sdk。

接口：agent 不新增导出符号（Config.MCPServers 已在 goal-59 声明）。mcptest 包导出测试所需的启动函数，由执行者在包文档列出。

行为：
- 工具目录：每轮开始时调用 sdk.ListTools，并对已连接（或首次使用时连接）的 MCP 服务器调用 tools/list；合并为一个列表交给模型。名称映射：Provider 只接受 ^[a-zA-Z0-9_-]{1,64}$ 的名字（两种协议都按此规则，命名常量），把 "plugin_id/service_name" 与 "mcp/<server>/<tool>" 转为合规名字（非法字符替换为 _，冲突或超长时加 _ 与短哈希）；映射在一轮内固定，结果交回模型时用映射后的名字。
- 循环：模型返回工具调用时依次执行，结果作为 ToolResult 交回，再继续请求；文本照常流式显示（沿用 goal-59）。单轮工具步数上限为命名常量（初值 20），达到时停止请求模型，回复 agent.tool_limit。
- Core 工具：sdk.InvokeTool（在延续调用内，继承父调用链与 Actor）；错误（PERMISSION_DENIED、DEADLINE_EXCEEDED、工具错误等）转为文本结果交回模型（"error: <code>: <safe_message>"），不重试。
- MCP 工具：先 sdk.ConfirmOperation("mcp/<server>/<tool>", 参数的 JSON 摘要)，允许后经 go-sdk 调用；拒绝与超时同样作为结果交回。MCP 工具的参数在交给模型前就以其 inputSchema 暴露，调用前不在 Agent 侧另做 Schema 校验（由服务器校验）。
- MCP 连接：stdio 以 command 与 args 启动子进程，env 为配置中解析后的 secret 值（SecretPair），子进程在插件的进程组内（由 Core 的进程树管理回收）；streamable HTTP 以 url 与 headers 连接。首次使用时连接，失败时该服务器本轮不可用并以 WARN 记录（不阻止对话）；插件 OnStop 时关闭全部会话并等待 stdio 子进程退出。
- 安全：MCP 的 env 与 headers 中的值不写日志、不进入工具结果。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tools_test.go（//go:build unix）。用 goal-59 的测试装配、模拟 Provider（脚本化地返回工具调用）、goal-54/55 的工具插件与 mcptest 的两种服务器：
  - 模型连续调用一个 Core 工具与一个 MCP 工具：两者都执行，结果交回模型，最终回复显示。
  - Core 工具为 confirm 策略且用户拒绝：结果为错误文本交回模型，工具未执行；MCP 工具的确认被拒同样如此。
  - 工具出错：错误作为结果交回，不重试。
  - 模型持续请求工具超过 20 步：本轮结束并回复步数上限文案。
  - 停止插件后 stdio MCP 子进程已退出（按 PID 检查）。
  - 名称映射对含 / 与 . 的名字、冲突名字的结果稳定。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：名称映射函数；MCP 连接失败时对话继续；secret 值不出现在日志中；HTTP 服务器的 headers。

完成条件：gofmt -l 无输出；go mod tidy 后无差异；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：记忆（T-12.4）；技能（T-12.5）；MCP 的资源与提示词能力（只用 tools）；真实 Provider（T-10）。
```

### goal-61 · T-12.4 Agent 的记忆

状态：等待前置（T-08.5、T-11.2、T-12.2）。语句中 goal-19、goal-32、goal-54、goal-59 提供的符号（sdk.JSONService 与 Tool 选项、gormdb.Open、agent 的一轮对话）只写名字；合并后由审查者核对，再改为可派发。2026-10-28 按 D-60 改写：记忆并入 Agent，不再有 sdk/memory 与记忆插件。

```text
/goal 使用 $alonabot-execute 推进 goal-61（切片 T-12.4：Agent 的记忆）。goal 目录 goal-61，分支 goal-61，worktree .worktrees/goal-61，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：官方 Agent 组件 agent/ 实现 Hermes 式的记忆：用 gormdb 在 Agent 的受管 SQLite 中保存会话笔记（按默认 Target）与用户档案（按 Actor）；注册 memory 工具（allow）执行原子的 add/patch/remove；一轮对话在会话第一轮把会话笔记写为快照条目，每轮把当前 Actor 的档案附在用户消息之前。

依据：PLAN T-12 步骤 3、5；CONTRACTS §13.1、§13.4、§3；DECISIONS D-60；TEST-MATRIX V-59。

范围（只新增或修改这些）：
- agent/：新增 memory.go（存储、工具与注入）、迁移追加记忆表；一轮对话只在调用处加两处调用；根组件注册 memory 工具（只改注册处）
- 上述包的测试文件
允许的新跨模块边：无（agent/... -> sdk/... 已允许）。

接口：agent 不新增导出符号。

行为：
- 范围：会话按默认 Target（adapter_instance_id、chat_id、thread_id），用户按 Actor（adapter_instance_id、external_user_id），都取自 Core 下发的可信上下文；memory 工具经 core.tool.invoke 调用时取工具调用的上下文，不采信参数中的范围。缺少所需的部分时返回 FAILED_PRECONDITION。
- memory 工具（JSON 服务加 sdk.Tool，default_policy=allow，只允许本组件经 core.tool.invoke 调用）：输入 {"operations": [{"op": "add"|"patch"|"remove", "target": "memory"|"user", "text": "...", "old_text": "..."}]}；patch 与 remove 用唯一的 old_text 定位（找不到或不唯一 → INVALID_ARGUMENT）；整组按结果检查上限（会话笔记 2,200、用户档案 1,375 个 Unicode 码点，命名常量）后在一个事务中写入；超限、完全重复的条目、含 Unicode Cf 类字符的内容返回错误，已有条目不变。条目按写入顺序拼接为文本（每条一行）。
- 一轮对话：
  - 会话的第一轮（会话为空时）读取会话笔记，作为 {"role": "system", "kind": "memory_snapshot", "content": ...} 写入会话；之后的轮次使用会话中的快照，不再读取（会话内快照不变、/reset 后下一会话重新取）。
  - 每轮读取当前 Actor 的档案，附在本轮用户消息之前。
  - 读取记忆失败时本轮按 agent.memory_error 报错并结束，不跳过记忆继续。
  - 模型的写入经工具循环调用 memory 工具（T-12.3 已接入工具循环；若本切片先于 goal-60 合并，工具调用留到 goal-60 之后验证）。

测试（按编码规范「测试提交规则」）：
- 提交：agent/memory_test.go（//go:build unix）。经真实 Core 装配与模拟 Provider：
  - 经 core.tool.invoke 调用 memory 工具：add、patch、remove 正确；超限时返回错误且已有条目不变；重复条目与含 U+200B 的内容被拒；群聊中两位用户各自只看到自己的档案；管理员把 memory 工具设为 deny 后调用返回 PERMISSION_DENIED。
  - 第一轮会话写入快照条目，第二轮请求中的快照与第一轮相同（会话内写入不进入快照）；/reset 后重新取快照；模拟 Provider 收到的本轮用户消息前附有当前 Actor 的档案。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：码点计数（含组合字符与 emoji）；old_text 不唯一；空操作组；缺少 Actor 或默认 Target 的 FAILED_PRECONDITION。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/...；go build ./...；在 agent 运行 alonabot plugin check 通过。

不做：技能（T-12.5）；定时任务（T-12.6）；向量检索或外部记忆后端。
```

### goal-82 · T-16.2 按钮、多轮等待、schedule 与示例测试

状态：等待前置（T-03.8、T-05.4、T-07.2、T-16.1）。语句中 goal-8、goal-24、goal-52、goal-81 提供的符号（Button.core_token 与 button.clicked 的派发、调度器的运行入口、core.wait 的等待状态、plugintest 与 inproc.Harness）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-82（切片 T-16.2：按钮、多轮等待、schedule 与示例测试）。goal 目录 goal-82，分支 goal-82，worktree .worktrees/goal-82，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：plugintest 支持对已发送按钮的点击、多轮等待中的注入、立即触发一次 schedule 与读取分发链之外的后台发送；用套件为示例插件编写提交的集成测试；docs/development.md 补充「插件测试」一节；在 TEST-MATRIX §3.2 记录 V-73 的证据。

依据：PLAN T-16 步骤 3、4 与「失败停止条件」；CONTRACTS §17、§2.4、§4.3、§8（后台 Send 的准入）、MESSAGE-MODEL §6.3 的按钮 token 规则；DECISIONS D-47；TEST-MATRIX V-73。

范围（只新增或修改这些）：
- plugintest/：Click、Trigger、Background 与等待状态的处理
- plugintest/internal/probe/：增加按钮、多轮等待、schedule 与后台发送的用例路由
- internal/inproc/harness.go：Outcome 增加 Waiting；Harness 增加 Trigger
- 根包：把调度器的立即运行入口接到 Harness.Trigger（只加调用点）
- internal/core/router/：Handler 进入 core.wait 时也调用分发结束回调（只加调用点）
- internal/core/ 中 goal-24 的调度器：没有立即运行一次的入口时增加一个（只加函数，复用正常的运行槽、期限与跳过重叠规则）
- examples/plugins/ 下的示例：为合并时已有的 echo、media 与命令示例各新增一个 *_test.go（用 plugintest）
- docs/development.md：新增「插件测试」一节（本切片获授权修改这一节）
- spec/TEST-MATRIX.md §3.2：追加 T-16 的证据行（本切片获授权修改这一处；其他 spec 文件不改）
允许的新跨模块边：examples/plugins/... -> plugintest（仅测试代码，ARCHITECTURE 已登记）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/inproc
   Outcome 增加 Waiting bool // Handler 进入多轮等待，分发链尚未结束
   Harness 增加 Trigger func(ctx context.Context, pluginID, jobID string) (runID string, err error) // 由根包设置；在调用结束后返回
   Dispatch 增加 RunID string // schedule 调用结束时设置，SourceEventID 为空
2. plugintest（在 goal-81 上追加）
   Outcome 增加 Waiting bool
   func (h *Harness) Click(from User, sent Sent, actionID string) Result // 在 sent.Chain 的 buttons 中按 action_id 找到按钮，以其 core_token 交付 button.clicked；找不到时 t.Fatal
   func (h *Harness) Trigger(root sdk.Root, jobID string) Result       // 立即运行该插件的一次 schedule，调用结束后返回期间的发送记录
   func (h *Harness) Background() []Sent                                // 返回并清空不属于任何一次注入的发送记录，按发生顺序

行为：
- 等待：Router 在某条分发链的 Handler 进入 core.wait 时，以 Waiting=true 的 Outcome 调用分发结束回调；Message 与 Click 收到它后返回。之后对同一会话与用户的 Message 按普通注入处理，Core 按等待键把它交给等待中的 Handler，该 Handler 的后续发送属于这次注入的结果。
- Click：button.clicked 的 Actor 为 from，Target 为 sent 的 Target，origin_platform_message_id 为 sent 的消息 ID；token 过期、插件代次变化或权限不足时由 Core 照常拒绝，结果中没有发送，Outcomes 为空。
- Trigger：经 Harness.Trigger 运行一次，走与定时触发相同的运行槽、15 分钟期限与跳过重叠规则；同 job 正在运行时返回该错误并 t.Fatal。返回这次调用期间的发送记录与 Outcome。
- Background：注入与 Trigger 之外发生的发送（例如插件在 OnReady 中启动的后台任务）进入后台记录。后台 Send 仍经 Core 的 Target 准入：没有 Allow 时结果为 PERMISSION_DENIED。
- 示例测试：每个示例至少覆盖它的主要命令或事件，以及一条失败路径（权限、参数错误或 UNSUPPORTED）；不依赖 time.Sleep。
- 文档：「插件测试」一节给出最小示例（Start、Text、断言 Sent）、各选项的用途、不跳过 Core 检查的说明，以及运行命令 go test -race ./...。

测试（按编码规范「测试提交规则」）：
- 提交：plugintest/plugintest_test.go 追加：
  - 按钮：/menu 发送带两个按钮的消息；Click 第一个按钮得到对应回复；另一个用户 Click 同一按钮也得到回复（按点击者鉴权）；按钮对应的 action 要求 Core Admin 时，普通用户 Click 没有回复。
  - 多轮等待：/ask 询问后 Message 立即返回，Outcomes 中有 Waiting；同一用户回答后得到确认回复；另一个用户的消息不被等待中的 Handler 消费。
  - schedule：probe 注册的 job 发送到 Allow 过的测试实例，Trigger 返回一条 Sent；未 Allow 时该发送为 PERMISSION_DENIED。
  - 后台：probe 的 OnReady 启动的后台发送出现在 Background 中，读取后清空。
  - go test -race -count=2 ./plugintest/... ./examples/... 两遍都通过。
- 证据：TEST-MATRIX §3.2 追加一行，写明覆盖的场景、命令、平台与 Go 版本。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Click 找不到 action_id；同 job 重叠时的 Trigger；等待超时后再次注入按普通消息处理。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./plugintest/... ./examples/... ./internal/core/... ./internal/inproc/...；go test -race . ；go build ./...；node scripts/check-repository.mjs。

不做：模拟时钟或缩短 schedule 的周期（只提供立即运行）；WebUI 沙盒（v0.1 之外）；为 Adapter 作者提供的测试工具。
```

## 第 12 波

### goal-41 · T-09.6 唯一提交点与激活门禁

状态：等待前置（T-07.5、T-09.5）。语句中 goal-27、goal-40 提供的符号（journal.Orchestrator 与 Result、supervisor 的激活与接入点、host 的 Degraded）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-41（切片 T-09.6：唯一提交点与激活门禁）。goal 目录 goal-41，分支 goal-41，worktree .worktrees/goal-41，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：确认候选的程序、配置快照、secret 与数据库引用都已持久化后，在 Core 状态库的同一事务中复核预期旧代次并提交 ActiveGeneration 与 COMMITTED（事务内不下载、不备份、不等待 RPC）；随后发送 ACTIVATE，只有 OnReady/ACTIVATED 完整成功才开放命令、事件、Service、Bus、任务与 HTTP；提交后启动或 OnReady 失败时保持新代次为权威并进入 Degraded，不自动倒退到旧数据库。旧代次的按钮随代次失效。

依据：PLAN T-09 步骤 5；CONTRACTS §12.1、§12.3、§6.3 第 2 条、§4.2；MESSAGE-MODEL §6.3；TEST-MATRIX V-30、V-21 的提交分项。

范围（只新增或修改这些）：
- internal/core/journal/：新增 commit.go（Commit 与提交后的激活）；Reload/Update 在 Prepare 之后调用 Commit（替换 goal-40 中「提交入口为 UNSUPPORTED」的分支）
- internal/core/state/：新增在同一事务中写 Active、COMMITTED 与 Journal 阶段的方法（新文件，不改 CommitActive 的已有行为）
- internal/core/supervisor/：激活结果的回调（只加调用点）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/state
   func (s *Store) CommitCandidate(ctx context.Context, opID, pluginID, expectedOld, newGen string) error // 同一事务：核对当前 Active == expectedOld、Journal 阶段为 prepared、代次快照已写入；然后更新 Active、写 COMMITTED、Journal 阶段改为 committed。不满足时返回 ErrStaleGeneration 或带原因的错误
internal/core/journal
   func (o *Orchestrator) Commit(ctx context.Context, r *Result) error // CommitCandidate → ACTIVATE → 等待 ACTIVATED；之后 Attach 到 Router/Broker/Bus/Scheduler/HTTP

行为：
- 提交前置：journal 阶段为 prepared、候选代次的 config_snapshot 与 generation_secret_paths 已写入、候选目录与数据库文件存在（fsync 已完成），否则不提交，按 Abort 处理。
- 事务：CommitCandidate 只操作状态库元数据；任何文件操作或 RPC 都在事务之外完成。提交失败（Active 已被并发修改等）→ 不激活，Abort 候选，旧版以新 Session 恢复。
- 激活：提交成功后向候选 Session 发送 ACTIVATE（带 OnReady 截止），收到 ACTIVATED 后才 Attach 到 Router、Broker、Bus、Scheduler 与 HTTP（复用 goal-27 的接入点）；在此之前候选的路由不可见。
- 提交后失败（OnReady 失败、超时、进程退出）：新代次保持为 Active，插件进入 Degraded（不派发、HTTP 503、Scheduler 不触发），以 ERROR 记录；不切回旧代次、不使用旧数据库。之后的恢复由管理员 restart（以新代次的新 Session）或 T-09.7 的 rollback 处理。
- 按钮：提交新代次后，旧代次签发的按钮 token 在解析时因 generation 不是当前 Active 而失效（goal-8 已实现，本切片补测试）。
- 首次安装同样经此提交；first_install 提交后失败时插件为 Degraded 而不是未安装。
- 不出现两个代次同时接单：旧进程在 goal-40 的冻结阶段已确认退出，提交后只有候选 Session 能 Attach。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/commit_test.go（//go:build unix）：
  - reload 成功：Active 改为新代次，COMMITTED 存在，Journal 为 committed；ACTIVATED 之前发送的消息没有派发给候选（候选的事件路由调用计数为 0，消息被旧版处理或因冻结被丢弃并计数），之后由新代次处理。
  - 提交前把 Active 改成别的值（模拟并发）：CommitCandidate 失败，候选被 Abort，旧版恢复。
  - 提交后 OnReady 失败：Active 仍为新代次，插件 Degraded，没有启动旧代次；HTTP 返回 503。
  - 提交后旧代次的按钮点击不调用插件。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：CommitCandidate 的每个前置条件；事务中没有文件与 RPC 操作（代码审查加测试钩子）；Degraded 状态下 Scheduler 不触发。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Core 崩溃后的恢复判定与 rollback（T-09.7）；管理页面（T-09.8）。
```

### goal-62 · T-12.5 技能

状态：等待前置（T-12.3）。语句中 goal-59、goal-60 提供的符号（agent 的 Config.SkillsDir、工具目录与循环）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-62（切片 T-12.5：技能）。goal 目录 goal-62，分支 goal-62，worktree .worktrees/goal-62，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Agent 组件启动时读取技能目录下各子目录的 SKILL.md：YAML frontmatter 的 name 与 description 必填，其余内容为正文；系统提示列出各技能的名称与描述，load_skill 工具返回正文；不合法的文件记诊断并跳过，不阻止插件启动。

依据：PLAN T-12 步骤 6；CONTRACTS §13（Agent 段）；TEST-MATRIX V-60。

范围（只新增或修改这些）：
- agent/：新增 skills.go；系统提示与工具目录接入技能（只改调用处）
- 上述包的测试文件
允许的新跨模块边：无（YAML 用已批准的 go.yaml.in/yaml/v3）。

行为：
- 目录：Config.SkillsDir 为空时不启用技能；相对路径以插件工作目录为基准。只读取一级子目录中的 SKILL.md（文件名区分大小写），不跟随符号链接，单个文件上限 256 KiB（命名常量），超过则视为不合法。
- 解析：文件以 "---\n" 开头的 frontmatter 到下一个 "---\n" 为止，用 yaml.v3 严格解码（未知字段允许并忽略）；name 为 [a-z0-9_-]{1,64}，description 为 1～1024 字符的非空字符串；同名技能以目录名字典序第一个为准，其余记诊断。不合法的文件以 WARN 记录（路径与原因）并跳过。
- 系统提示：在人格提示之后追加一段（文案键 agent.skills.header），逐行列出「name: description」，按 name 排序；没有技能时不追加。
- load_skill：插件内工具（不经 Core 的工具目录与策略），在 goal-60 的工具列表中加入，输入 {"name": string}，返回正文；名字不存在时返回错误文本 "error: NOT_FOUND: unknown skill"。不计入额外的确认。
- 技能在插件启动时读取一次，修改后需要重启插件。

测试（按编码规范「测试提交规则」）：
- 提交：agent/skills_test.go（//go:build unix）：临时技能目录含两个合法技能、一个缺少 frontmatter、一个缺少 description、一个超过大小上限：模拟 Provider 收到的系统提示只含两个合法技能，日志有三条诊断，插件正常启动；模型调用 load_skill 得到正文；不存在的名字得到错误结果。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：frontmatter 的边界（CRLF、没有结束分隔）；同名技能；符号链接被忽略；name 的字符规则。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：技能的热更新；从网络下载技能。
```

### goal-63 · T-12.6 定时 Agent 任务

状态：等待前置（T-07.2、T-08.5、T-12.3）。语句中 goal-24、goal-32、goal-55、goal-59、goal-60 提供的符号（ScheduleRegistry.Every、gormdb.Open、sdk.ConfirmOperation、Agent 的一轮对话与工具循环、contracts/schedule）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-63（切片 T-12.6：定时 Agent 任务）。goal 目录 goal-63，分支 goal-63，worktree .worktrees/goal-63，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Agent 组件提供 schedule_task、list_tasks、cancel_task 三个 Agent 内工具：只有 Core Admin 能创建与取消（之前经 core.tool.confirm），其他用户只能列出本聊天的任务；任务保存创建者、Target、提示词与触发规则（contracts/schedule 校验与计算）。一个每分钟一次的固定 schedule 检查到期任务，在其回调内以没有 Actor、只有 job scope 的方式运行一轮对话；发送需要 allowed_adapter_instances，confirm 策略的工具一律拒绝，不注入记忆，会话键与聊天会话分开，离线错过的触发不补跑。

依据：PLAN T-12 步骤 7；CONTRACTS §4.3、§13；TEST-MATRIX V-61。

范围（只新增或修改这些）：
- agent/：新增 tasks.go（工具与任务表）、scheduled.go（每分钟的 schedule 与后台一轮）；Agent 根插件注册 schedule（只改注册处）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> contracts/schedule（plugins -> contracts 已允许）。

行为：
- 任务表（gormdb，versioned 迁移）：id、creator（instance、user）、target（instance、chat、thread）、prompt、trigger（interval 秒数或 cron 与时区）、next_run_at、created_at。
- schedule_task：输入 {"prompt", "interval_seconds"?, "cron"?, "timezone"?}，恰有 interval 或 cron 之一，经 contracts/schedule 校验（interval 至少 60s，命名常量；cron 为五字段）；调用者（工具循环的当前 Actor）必须含 ADMIN 角色，否则返回错误结果 "error: PERMISSION_DENIED"；先 ConfirmOperation("agent/schedule_task", 摘要)，允许后写入，next_run_at 为 Trigger.Next(创建时刻, 现在)。Target 为当前对话的默认 Target。
- list_tasks：返回本聊天（当前默认 Target 的 instance、chat、thread）的任务；任何人可用。
- cancel_task：输入 {"id"}；只有 ADMIN，先确认，再删除；只能取消本聊天的任务。
- 每分钟的 schedule（Every 1 分钟）：插件激活时把每个任务的 next_run_at 重算为未来时刻（不补跑）；每次回调选出 next_run_at ≤ 现在的任务，每个任务本次只运行一次，然后用 Trigger.Next 更新 next_run_at。运行一轮对话：
  - 没有 Actor，只有 job scope；会话键为 "task/<id>"，与聊天会话分开；不调用记忆服务。
  - 回复经 sdk.Send 显式 To 任务的 Target，需要 allowed_adapter_instances，没有许可时 PERMISSION_DENIED 并记 WARN（任务保留）。
  - 工具循环中 confirm 策略的 Core 工具与全部 MCP 工具一律拒绝（不发起确认，因为没有 Actor），作为错误结果交回模型。
  - 一次回调的运行受 JOB_TIMEOUT 约束；同一任务在上一次运行未结束时跳过。
- 插件离线期间到期的任务在重新激活后不补跑。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tasks_test.go（//go:build unix）：用模拟 Provider（脚本化调用 schedule_task 等工具）与注入时钟的 Scheduler：
  - 非 Admin 调用 schedule_task 得到错误结果，任务表为空；Admin 经确认后创建 interval 与 cron 任务各一个。
  - 时钟推进到到期：任务运行一轮，回复发往任务 Target；没有发信许可时没有发送并记 WARN，授予许可后下一次运行发送。
  - 后台一轮中模型调用 confirm 策略的工具得到拒绝结果；没有注入记忆（请求中没有快照与档案）。
  - 停止插件、推进时钟越过多次触发、再启动：不补跑，只在下一次未来时刻运行。
  - 其他用户 list_tasks 只看到本聊天的任务；cancel_task 需要 Admin 与确认。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：interval 下限；一次回调中同一任务只运行一次；会话键隔离；JOB_TIMEOUT 内未结束时跳过。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：打包与 T-12 验收（T-12.7）；任务的管理页面。
```

## 第 13 波

### goal-42 · T-09.7 崩溃恢复、回退与回收

状态：等待前置（T-09.6）。语句中 goal-33、goal-39～goal-41 提供的符号（journal 的阶段与 Abort、state.CommitCandidate、candidate 的 PrepareDatabase/Discard、manage.Candidates.Rollback、secret.Collect）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-42（切片 T-09.7：崩溃恢复、回退与回收）。goal 目录 goal-42，分支 goal-42，worktree .worktrees/goal-42，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：Core 启动取得 core.lock 后按状态库的 Active 与 Journal 恢复：未提交的操作关闭候选、以旧程序启动新 Session（首次安装没有 Active 则保持未激活），已提交的只恢复当前代次；旧进程退出未确认时不启动新写入者。rollback 从选定的保留代次复制新的 SQLite 恢复候选并使用新代次，走同一准备与提交路径；插件自管的外部数据沿用显式维护。只回收无人引用的失败候选、旧包与 secret，恢复中的操作与保留快照受保护。

依据：PLAN T-09 步骤 6；CONTRACTS §12、§6.3、§9.3（rollback）；MESSAGE-MODEL §6.3（rollback 不复活旧按钮）；TEST-MATRIX V-30、V-31、V-36。

范围（只新增或修改这些）：
- internal/core/journal/：新增 recover.go（启动时恢复）与 rollback.go（实现 manage.Candidates.Rollback）
- internal/core/state/：保留代次的读写（新文件；保留策略见下）
- 新增 internal/core/gc/（引用回收）
- 根包 assembly.go：serve/Run 启动时在启动任何包之前调用恢复（只加调用点）
- docs/development.md：增加「备份与恢复」一节，说明整机备份覆盖整个 data_dir 与插件自管的外部数据（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/journal
   func (o *Orchestrator) Recover(ctx context.Context) error // 处理全部未到 committed 或 cleaned 的 Journal
   func (o *Orchestrator) Rollback(ctx context.Context, id, generation string) error
internal/core/gc
   func Collect(ctx context.Context, st *state.Store, secrets *secret.Store, packagesDir string) (removed []string, err error)
internal/core/state
   保留代次：每个插件保留最近 3 个已提交的代次（命名常量）及其程序包与数据库；读写方法自定。

行为：
- Recover（Core 启动时，在启动任何包之前；已持 core.lock）：
  - Journal 阶段为 started、stopped_old、candidate_ready、prepared、failed：候选未提交。确认没有仍在运行的候选或旧进程（通过 runtime.lock 能否取得判断；取不到时该插件标为 Degraded 并以 ERROR 说明「旧进程可能仍在运行」，不启动任何实例）；然后 candidate.Discard 候选资源，Journal 改为 cleaned；有旧 Active 时稍后按正常启动流程启动旧版（新 Session），首次安装则保持未激活。
  - Journal 阶段为 committed 但插件没有成功 ACTIVATED 过：只按当前 Active 启动（新代次），失败则 Degraded；不切回旧代次。
  - 恢复判断只信状态库的 Active 指针与 Journal，不读插件目录猜测。
- Rollback（管理服务调用，持插件操作锁）：选定的代次必须是保留代次之一且不是当前 Active；以 Journal kind=rollback 经 goal-40 的编排：停止当前版本、从选定代次的数据库用 Backup API 复制到新代次（不是复用旧文件）、以选定代次的程序包与配置快照启动候选（activation_reason=rollback）、经 goal-41 提交。新代次 ID 为新值，旧代次签发的按钮不会因此复活（解析时比较的是当前 Active）。声明 prepare_effects=external 的插件 rollback 返回 FAILED_PRECONDITION（"NEEDS_MANUAL_MIGRATION"：需要维护与外部备份），不假装能恢复。
- 回收（Collect，可在 serve 启动后与每次操作结束后调用）：删除不被 Active、保留代次、Pending 与未完成 Journal 引用的代次目录、程序包目录与 secret；绝不删除受保护资源；返回删除的路径列表并以 INFO 记录。
- 文档：docs/development.md 新增「备份与恢复」：整机备份须在 Core 与插件停止后复制完整 data_dir，并单独备份插件自管的外部数据；只备份 state.sqlite 不包含插件业务数据。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/recover_e2e_test.go（//go:build unix）。用测试钩子在各节点杀掉 Core 进程（PREPARED 之后提交之前、提交之后 ACTIVATED 之前），再重启 serve：
  - 提交前崩溃：只恢复旧版，候选目录被清理，Journal 为 cleaned。
  - 提交后崩溃：只恢复新代次。
  - 崩溃时旧插件进程仍存活并持有 runtime.lock（测试插件忽略断链）：重启后该插件为 Degraded，没有启动第二个写入者；旧进程退出后 restart 恢复。
  - rollback 到上一代次：产生新代次与新的数据库副本，内容等于旧代次的数据；旧代次签发的按钮仍无效。
  - 声明 prepare_effects=external 的测试插件 rollback 返回 NEEDS_MANUAL_MIGRATION。
  - Collect 不删除保留代次与 Active，删除失败候选的残留。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Recover 对每个 Journal 阶段的分支；保留代次的淘汰顺序；Collect 对 Pending 引用的保护。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：管理页面（T-09.8）；外部数据的备份工具（由管理员处理）。
```

### goal-43 · T-09.8 Admin 管理页

状态：等待前置（T-09.6、T-08.7、T-09.12）。语句中 goal-34、goal-36、goal-39～goal-41、goal-84 提供的符号（web/admin 的工程与组件、adminauth 的登录接口、manage 的 HTTP API、Admin API 参考与 WebUI 的提供）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-43（切片 T-09.8：Admin 管理页）。goal 目录 goal-43，分支 goal-43，worktree .worktrees/goal-43，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：在官方 WebUI 组件 web/admin 中实现真实的登录、总览状态、各角色组件的安装、启停与替换（`--source`、`--replace` 的页面入口，组件图标用 `<img>` 显示）、Adapter 实例管理、更新与恢复（rollback）页面，全部只调用 spec/ADMIN-API.md 列出的 Admin API；静态文件由 T-09.12 以 WebUI 组件的方式提供。页面显示 Pending、候选失败、Degraded 与维护原因，只展示用户可操作的状态和具体错误。

依据：PLAN T-09 步骤 7（管理页）；CONTRACTS §7.3、§7.4、§9.3、§12.1；DECISIONS D-25；TEST-MATRIX V-40、V-47 的页面分项。

范围（只新增或修改这些）：
- web/admin/src/：新增 api/（管理 API 客户端，带 CSRF 头）、pages/（Login、Overview、Packages、Instances、Updates）、路由（基于 window.location 与 history 的简单路由，不引入路由库）
- web/admin/i18n/：补充页面文案（en-US 与 zh-CN 键集合一致）
- 新增 internal/core/adminui/（嵌入 web/admin/dist 的静态资源并在 admin host 提供；go:embed）
- cmd/alonabot 或根包 assembly：admin host 的非 /api/ 路径交给 adminui（只加调用点）
- 构建流程：web/admin 的 npm run build 产物 dist/ 由 internal/core/adminui 的 go:generate 复制进包目录；生成物提交到仓库（与 proto 生成物同样处理），并在 docs/development.md 的生成命令中登记（本切片获授权修改该处）
- 上述包的测试文件
允许的依赖：前端仍只用已批准的 React、Vite、TypeScript、vitest；不引入路由、状态管理或 UI 库。

行为：
- 登录：POST /api/auth/login，成功后保存 csrf_token（内存中，刷新页面时重新登录或经 /api/auth/session 取回；若 goal-36 没有该端点，本切片在 adminauth 中追加 GET /api/auth/session 返回当前会话的 csrf_token，只在已登录时可用）；登出。未登录访问任何页面跳转到登录页。
- Overview：Core 版本、监听地址、每个包的状态（未启用、运行、Pending、候选失败、Degraded 及原因、维护中）。
- Packages：上传安装包（archive、release.json、.sig 三个文件，multipart 由 API 接收，若 goal-39 的 API 只接受路径，本切片在 manage 中追加 multipart 端点并写入临时目录后调用 Install）；启用、停用、重启、卸载（含 purge）、pin；需要确认时展示受影响列表并要求再次确认（API 的 409）。
- Instances：列出 Adapter 实例（不显示凭证，只显示「已设置」）、创建（表单由该 Adapter 包的配置 Schema 经 goal-34 的 ConfigForm 生成）、启停、删除。
- Updates：update 上传新包、rollback 选择保留代次；显示操作结果与失败原因。
- 安全：所有修改请求带 X-Alonabot-CSRF；页面不读取或展示 token、secret 明文；静态资源响应带 CSP（default-src 'self'）、X-Frame-Options: DENY 与 nosniff。
- adminui：只在 admin host 提供；未知路径返回 index.html（单页应用）；/plugins/ 与 /api/ 不经过它。

测试（按编码规范「测试提交规则」）：
- 提交：
  - web/admin/src/api/api.test.ts（vitest，不依赖 DOM）：用 fetch 的测试替身验证 API 客户端的 CSRF 头、409 确认流程的数据结构、错误格式解析。
  - internal/core/supervisor/adminui_test.go（//go:build unix）：admin host 上 / 返回 index.html 且带 CSP；plugin host 上不提供；未登录时 API 返回 401；登录后经 API 完成一次安装与启用（复用 goal-39 的测试包）。
- 本地自测（不提交，PR 中写明结果）：npm run build 与 typecheck；在浏览器中手动走一遍：登录 → 安装 → 启用 → 创建实例 → 停用 → rollback（写明浏览器与版本，属于手动验证）。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功；go generate ./... 后无未提交差异；gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：配置、命令权限、发信许可、工具策略、会话与插件页（T-09.9）；日志页（T-09.10）；Core 设置页（T-09.9）。
```

### goal-64 · T-12.7 官方 Agent 组件打包与 T-12 验收

状态：等待前置（T-12.4、T-12.5、T-12.6、T-12.8）。语句中 goal-38、goal-58～goal-63、goal-83 提供的命令与符号（plugin pack、Agent 组件、mcptest、模拟 Provider）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-64（切片 T-12.7：官方 Agent 组件打包与 T-12 验收）。goal 目录 goal-64，分支 goal-64，worktree .worktrees/goal-64，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：官方 Agent 组件（`role: agent`）按 T-09 的打包流程生成可安装包；用本地 Provider 与 MCP 模拟服务、回环测试 Adapter（或 Telegram 模拟服务）以安装包方式运行 V-57～V-61 与 V-76 的全部场景；在 TEST-MATRIX §3.2 记录 T-12 证据，真实 Provider 与真实 Telegram 没有凭证时记为未运行。

依据：PLAN T-12 步骤 8 与「停止和失败条件」；CONTRACTS §13、§9.3；TEST-MATRIX V-57～V-61、V-76。

范围（只新增或修改这些）：
- agent/plugin.yaml：补齐打包所需字段（version、source、os/arch 由 pack 填写的部分除外）
- 新增 cmd/alonabot/t12_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-12 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

验收矩阵（以 plugin pack 产生的安装包经 alonabot serve 与管理 API 安装、启用后运行；各为一个子测试）：
- V-57：三种协议的流式文本、工具调用、停止原因与用量；/ask 与自由对话开关；群聊非 to_me；/provider 与人格切换；/reset；超长回复分段；Provider 出错或中途断开。
- V-58：连续调用 Core 工具与 MCP 工具（stdio、streamable HTTP）；确认拒绝；工具出错；超过单轮步数上限；MCP 进程随 Agent 停止。
- V-59：记忆 add/patch/remove、超限、重复、不可见字符、群聊多用户、会话中写入与 /reset、memory 工具设为 confirm 与 deny。
- V-60：合法与不合法的 SKILL.md、load_skill 不存在的技能。
- V-61：非 Admin 创建任务、interval/cron 任务到期运行、无发信许可、confirm 工具、插件离线错过。
- V-76：另一个测试插件经安装包调用 llm.generate：未列入与列入 llm_service_allowed_plugins、Provider 不存在、Provider 出错、Agent 未启用。
- 证据：TEST-MATRIX §3.2 追加 T-12 一行，写明测试位置、命令、平台与结果；Provider 与 MCP 标为本地模拟；真实 Provider 与 Telegram 为「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t12_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：Agent 组件的 plugin check 与 pack（当前平台）；若维护者提供了 Provider 凭证，另附一次手动运行记录，不写入提交的测试。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：发行归档与预装（T-14.3、T-14.4）。
```

## 第 14 波

### goal-44 · T-09.9 Admin 配置、权限、工具策略与插件页

状态：等待前置（T-09.8、T-11.1、T-11.3）。语句中 goal-13、goal-29、goal-31、goal-34、goal-39、goal-43、goal-53～goal-56、goal-26 提供的符号（policy 的覆盖与 Admin 方法、config.Service、secret.Redact、ConfigForm、manage API、web/admin 页面框架、i18n 的默认 locale、tool.Catalog、conversation.Service、webhost 的票据）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-44（切片 T-09.9：Admin 配置、权限、工具策略与插件页）。goal 目录 goal-44，分支 goal-44，worktree .worktrees/goal-44，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：在管理 API 与 web/admin 中补齐：插件配置（表单与完整 JSON 编辑，使用 raw JSON、Merge Patch、base_generation 与 secret 语义，显示 Pending 并经 reload 激活）、命令权限与启用、发信许可（allowed_adapter_instances）、Core 设置（Core Admin 列表、默认 locale、全局禁止表）、工具调用策略、Agent 会话的列出与清除、插件页面入口（经兑换票据进入 plugin host），以及界面语言切换。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §5、§2.3、§8（allowed_adapter_instances）、§13、§7.3、§5.4；DECISIONS D-45；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56 的页面分项。

范围（只新增或修改这些）：
- internal/core/manage/：新增 config_api.go、policy_api.go、settings_api.go、tool_api.go、conversation_api.go、pluginpage_api.go
- web/admin/src/pages/：新增 Config、Permissions、CoreSettings、Tools、Conversations；Packages 页增加「打开插件页面」入口
- web/admin/i18n/：补充文案
- internal/core/adminui：重新生成嵌入的构建产物
- 上述包的测试文件
允许的依赖：与 goal-43 相同，不新增前端依赖。

接口（HTTP，全部在 /api/v1 下、经 adminauth.Require；JSON；错误格式与 goal-39 一致）：
- GET /plugins/{id}/config：{active_generation, active: Redact 后的快照, pending: {base_generation, config: Redact 后}|null, schema}
- PUT /plugins/{id}/config/pending：{base_generation, config}（完整替换）；PATCH /plugins/{id}/config/pending：{base_generation, patch}（RFC 7396）；DELETE 丢弃 Pending；POST /plugins/{id}/config/import 导入 config.json；secret 字段的新值在写入 Pending 前经 secret.Extract 存为引用，省略或未修改的 secret 保持原引用；base 不是当前 Active 时 409 并返回 FAILED_PRECONDITION。
- POST /plugins/{id}/reload：经 goal-40/41 激活 Pending。
- GET/PUT /plugins/{id}/policies：命令覆盖（作用域、permission、enabled），以及 Router 的遮蔽诊断。
- GET/PUT /plugins/{id}/allowed-instances：发信许可集合。
- GET/PUT/DELETE /core/admins、/core/denies，GET/PUT /core/default-locale。
- GET/PUT /tools/policies：工具策略（缺省与覆盖）。
- GET /plugins/{id}/conversations、DELETE /plugins/{id}/conversations/{key}。
- POST /plugins/{id}/page-ticket：返回 webhost.IssueTicket 的 URL（只给有 HTTP 路由的插件）。
- GET /secrets/{plugin}/{path}/reveal：要求请求体中的当前 token 再认证（adminauth.Reauthenticate），通过后返回明文并以 INFO 审计记录（不含明文）。

行为：
- 配置页：用 ConfigForm 编辑（只提交 patch）或切换到完整 JSON 编辑（textarea，提交完整替换）；显示 Pending 与 Active 的差异（按键列出，secret 只显示是否已设置）；「激活」调用 reload；候选失败与 Degraded 显示具体原因。hidden 字段只在完整 JSON 编辑中可见。
- 权限页：按命令树显示有效权限与启用状态（经 policy.CheckCommand 的同一规则计算），可设置插件、命令组、路由三层覆盖；显示遮蔽关系。
- 发信许可：列出全部 Adapter 实例，勾选授予。
- Core 设置：Core Admin 列表（与 alonabot admin add|remove|list 同一服务）、默认 locale、全局禁止表。
- 工具页：列出全部工具及缺省与覆盖策略。
- 会话页：按插件列出会话摘要（键、条目数、最后时间），可清除；不显示条目内容。
- 语言切换：沿用 goal-34 的组件，保存在 localStorage。
- 插件页面：点击后请求 page-ticket，在新窗口打开 plugin host 上的兑换 URL；未配置第二主机名时按钮禁用并提示。
- 所有修改请求带 CSRF 头；修改后的策略在下一次派发时生效（页面不需要重启插件）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/admin_api_test.go（//go:build unix）：经管理 API 完成：PATCH 配置（未修改的 secret 保持原引用、新 secret 生成新引用、hidden 不被删除）、陈旧 base 409、reload 后 Active 改变；设置命令覆盖后下一条消息按新权限处理；授予发信许可后插件后台 Send 成功；设置默认 locale 后新 Session 使用；工具策略改为 deny 后 ListTools 不含它；清除会话；page-ticket 只给有 HTTP 路由的插件；reveal 需要再认证。
  - web/admin 的 vitest：API 客户端对新端点的请求格式与 409 处理。
- 本地自测（不提交，PR 中写明结果）：npm run build；浏览器中手动走一遍各页面并切换语言（写明浏览器与版本）。

完成条件：web/admin 下 npm run typecheck、npm run test、npm run build 成功；go generate ./... 后无未提交差异；gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：日志页（T-09.10）；故障注入与验收汇总（T-09.11）。
```

### goal-66 · T-09.10 日志管理与日志页

状态：等待前置（T-02.1、T-03.1、T-09.8）。语句中 goal-2、goal-39、goal-43、goal-65（logging 的 Output、Levels、Source）提供的符号只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-66（切片 T-09.10：日志管理与日志页）。goal 目录 goal-66，分支 goal-66，worktree .worktrees/goal-66，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：按来源的日志级别持久化在状态库，经管理服务修改后通过 sdk.log.set_level 通知 SDK 并立即生效；全局级别可在运行时修改、Core 重启后恢复为启动设置；Core 为每个来源保留最近 LOG_BUFFER_RECORDS 条已输出的记录，提供按来源、级别与 trace_id 的读取和 SSE 订阅，过去日期从日志文件查询；alonabot logs 与 log-level 命令、WebUI 日志页使用同一 API。

依据：PLAN T-09 步骤 7（日志）；CONTRACTS §14.2～§14.6、§16（LOG_BUFFER_RECORDS）；WIRE-FORMAT §3.5（sdk.log.set_level）；TEST-MATRIX V-62、V-63、V-68。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：新增 LogSetLevelRequest/Response。随后 go generate ./...（本切片获授权只追加这两个消息）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceLogSetLevel = "sdk.log.set_level"
- internal/core/logging/：新增 buffer.go（最近记录）、query.go（按日期查询日志文件）；Output 在输出时写入缓冲（只加调用点）
- internal/core/state/：migrations 末尾追加一条；新增 loglevels.go
- internal/core/manage/：新增 log_api.go（读取、SSE、级别修改）
- sdk/：处理 sdk.log.set_level（新增文件；sdk/adapter 的 RunAdapter 同样处理）
- cmd/alonabot/：新增 logs 与 log-level 命令
- web/admin/src/pages/Logs：日志页；internal/core/adminui 重新生成
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message LogSetLevelRequest { LogLevel level = 1; }
   message LogSetLevelResponse {}
2. internal/core/state
   迁移：CREATE TABLE log_levels (source_id TEXT PRIMARY KEY, level TEXT NOT NULL); -- level 为 debug、info、warn、error、off；inherit 不存行
3. internal/core/logging
   type Record struct { Time time.Time; Level slog.Level; Source string; Message string; Attrs map[string]any; TraceID string }
   type Buffer struct{ ... } // 每个来源一个固定容量的环形缓冲
   func (o *Output) Recent(source string) []Record
   func (o *Output) Subscribe(ctx context.Context) <-chan Record // ctx 结束时关闭
   func Query(dir string, date time.Time, source string, level slog.Level, traceID string) ([]Record, error) // 读该日期的全部文件（含编号续写的文件）
4. HTTP（/api/v1，经 adminauth.Require）
   GET /logs/recent?source=；GET /logs/stream（SSE，事件为 JSON 的 Record）；GET /logs/history?date=YYYY-MM-DD&source=&level=&trace=；GET/PUT /logs/levels（全局与按来源）
5. cmd/alonabot
   alonabot logs [--source ID] [--level L] [--trace T] [--follow] [--date YYYY-MM-DD]；alonabot log-level [ID] [LEVEL]（不带参数时列出，ID 为 global 时修改全局）；都经管理 API 与 Bearer token

行为：
- 缓冲：只保存已经通过级别过滤并实际输出的记录；每个来源（core、每个插件、每个 Adapter 包）最多 LOG_BUFFER_RECORDS 条，满时丢弃最旧的；只在内存中，重启后清空。
- 级别：按来源的设置持久化；修改后先写状态库，再对运行中的 Session 发 sdk.log.set_level（inherit 时下发解析后的全局级别），并更新 Core 侧的 Levels；SDK 收到后立即按新级别过滤。全局级别的运行时修改只在内存中，Core 重启后恢复为 --log-level 或设置文件的值；inherit 的来源随之变化（同样发 set_level）。来源为 off 时 SDK 不输出，Core 关于它的记录仍按全局级别输出。
- 读取与订阅都要求管理员会话或 Bearer；SSE 先发送当前缓冲中的记录再推送新记录（与页面的「先读再订阅」配合，页面按时间去重）。
- 历史查询：按 CONTRACTS §14.6 的文件命名读取该日期的文件，解析控制台格式的行（goal-65 的文件格式）为 Record；文件不存在返回空列表。
- 日志页：选择来源、级别、trace 与日期（今天为实时，过去日期为历史），筛选只在浏览器中进行并保存在 localStorage；修改输出级别使用级别设置区域（调用 PUT /logs/levels）。
- 日志内容中的脱敏规则不变（goal-65 与 CONTRACTS §14.5）。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/logs_e2e_test.go（//go:build unix）：运行 serve 与一个持续输出日志的测试插件：
  - log-level 把该插件设为 debug 后，SDK 输出 DEBUG 记录且 logs 能读到；设为 off 后插件记录消失，Core 关于它的记录仍在；重启 serve 后按来源的设置保留，全局级别的运行时修改恢复为启动值。
  - logs --follow 收到新记录；未带 token 时 401。
  - 每个来源最多保留 LOG_BUFFER_RECORDS 条（用很小的值）。
  - 写入前一天日期的日志文件后，logs --date 能查到。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：环形缓冲的边界；历史文件解析（多行值、转义）；SSE 的断开；web/admin 日志页的构建与手动验证。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交差异；go vet ./...；go test -race ./...；go build ./...；web/admin 下 npm run test 与 npm run build 成功。

不做：远程日志收集与外部格式；日志文件设置的运行时修改（启动设置，§15.1）。
```

### goal-73 · T-14.3 发行归档

状态：等待前置（T-03.10、T-09.3、T-12.7、T-14.1、T-14.2、T-15.2）。语句中 goal-38 提供的命令（plugin pack 与 --key）与各官方插件、Telegram 与 OneBot 11 Adapter 的目录只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-73（切片 T-14.3：发行归档）。goal 目录 goal-73，分支 goal-73，worktree .worktrees/goal-73，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：scripts/dist/ 为每个目标平台构建 Core 与全部官方组件（Telegram Adapter、OneBot 11 Adapter、Agent、WebUI、help、welcome、rss；WebUI 先在 web/admin 运行 npm ci 与 npm run build，再按 any 平台打包一次），用 plugin pack 生成安装包与外部 release.json，以官方发布者私钥签名后放入 packages/，再打成 zip（Windows）或 tar.gz，并生成列出全部发行归档 SHA256 的 SHA256SUMS；发行版 Core 在构建时嵌入官方发布者公钥。

依据：PLAN T-14 步骤 3；CONTRACTS §15.2、§9.2；docs/development.md（交叉编译）；TEST-MATRIX V-66 的归档分项。

范围（只新增或修改这些）：
- 新增 scripts/dist/（Go 程序，go run ./scripts/dist 执行）
- 新增 internal/core/trust/（官方发布者公钥；构建时通过 -ldflags -X 注入，开发构建为空）
- internal/core/packages 或 manage：安装时把 trust 中的官方公钥视为已信任来源（只加调用点）
- docs/development.md：增加「发行构建」一节（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：scripts/dist -> contracts/release（ARCHITECTURE 已登记，用于自检签名）；其余只经 os/exec 调用 go build 与 alonabot 二进制。

接口：
- go run ./scripts/dist --version VERSION --key FILE --out DIR [--platforms linux/amd64,...]：缺省为五个目标平台；--key 为官方 Ed25519 私钥文件（不进入仓库与产物）。
- internal/core/trust：var officialKey string（base64 公钥，由 -ldflags "-X github.com/AlonaBot/AlonaBot/internal/core/trust.officialKey=..." 注入）；func Official() (ed25519.PublicKey, bool) // 未注入时 false

行为：
- 每个平台：CGO_ENABLED=0、-trimpath、-ldflags 注入版本与官方公钥，构建 alonabot；用该版本的 alonabot plugin pack --goos --goarch --key 为每个官方包生成安装包、release.json 与 .sig；产物放入 alonabot-<version>-<os>-<arch>/（alonabot 可执行文件与 packages/），Windows 打 zip，其他打 tar.gz（文件按名称排序、时间戳固定，可重复）。
- 自检：每个包用 contracts/release.Verify 与公钥验证一次；失败则中止。把各归档的 SHA256 写入 SHA256SUMS（`sha256sum` 的格式：摘要、两个空格、文件名）。
- 官方公钥：发行版 Core 安装时把它视为已信任来源（预装包不逐个确认，签名与摘要照常校验）；开发构建没有内置公钥，行为不变。
- 私钥只从 --key 读取，不写日志；docs/development.md 的「发行构建」写明命令、所需工具与私钥保管要求。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/dist/dist_test.go：用临时 Ed25519 密钥只为当前平台运行一次（--platforms 当前平台），检查归档结构（alonabot 与 packages/ 下每个官方包的三个文件）、每个 release.json 验签通过、两次运行的归档字节相同、构建出的 alonabot 内置的公钥等于测试公钥（经一个只读的 version 输出或测试钩子）。
- 本地自测（不提交，PR 中写明结果）：五个平台的完整构建（交叉编译，只验证产物生成，不在其他平台运行）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./scripts/... ./internal/...；go build ./...；node scripts/check-repository.mjs。

不做：首次运行的预装（T-14.4）；容器镜像（T-14.5）；发布到 GitHub Releases（T-10.4）。
```

## 第 15 波

### goal-45 · T-09.11 正式流程故障注入与 T-09 验收

状态：等待前置（T-09.3、T-09.7、T-09.9、T-09.10、T-09.12）。语句中 goal-38～goal-44、goal-66、goal-84 提供的命令与 API 只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-45（切片 T-09.11：正式流程故障注入与 T-09 验收）。goal 目录 goal-45，分支 goal-45，worktree .worktrees/goal-45，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：通过真实 CLI 与管理 API，在临时 data_dir 中跑完正式流程：首次安装 → enable → restart → reload → update → pin → disable → 同来源重装 → 显式 rollback；以包损坏、签名或公钥变化、external 插件的维护错误验证拒绝路径；在 PREPARED、提交前后与 OnReady 失败点终止 Core，连同残存旧进程、空间不足与同插件并发操作，验证没有双写、没有提前开放、没有静默的数据回滚。回填 T-08 留给集成的 V-23～V-25、V-39 证据，在 TEST-MATRIX §3.2 记录 T-09 证据，形成 T-10 可复用的恢复场景。

依据：PLAN T-09 步骤 8 与「停止和失败条件」；CONTRACTS §5、§6、§7、§9、§12、§14；TEST-MATRIX V-23～V-25、V-27～V-31、V-39、V-40、V-52、V-54～V-56、V-62、V-63。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t09_e2e_test.go（//go:build unix）与必要的测试插件（testdata 下）
- 新增 scripts/scenarios/（可重复执行的恢复场景脚本或 Go 测试辅助，供 T-10 在其他平台复用；只调用 alonabot 二进制）
- spec/TEST-MATRIX.md §3.2：追加 T-09 的证据行，并更新 T-08 行中留待集成的项（本切片获授权修改这两处）
- 上述包的测试文件
允许的新跨模块边：无。

验收矩阵（各为一个子测试，记录命令、日志摘录与状态库检查）：
1) 正式流程：首次安装 → enable（首个 Active）→ restart（只换 Session）→ 编辑 config.json 后 reload（新代次与 SQLite 副本）→ update（新包、新代次）→ pin → disable → uninstall 后同来源重装（复用数据）→ rollback 到上一代次（新代次、旧按钮不复活）。
2) 拒绝路径：归档损坏、摘要不符、签名错误、发布者公钥变化（需要确认）、错误的 OS/arch、来源 ID 冲突、../ 与符号链接逃逸的归档：都在停止旧版前拒绝，旧版继续运行。
3) external：声明 prepare_effects=external 的测试插件，已有安装的 update 返回 NEEDS_MANUAL_MIGRATION；授予维护许可后执行；rollback 返回 NEEDS_MANUAL_MIGRATION；普通 restart 不受影响。
4) 崩溃点：在 PREPARED 之后、提交之前、提交之后 ACTIVATED 之前、OnReady 失败时分别杀掉 Core，重启后的状态与 goal-42 的规则一致；任何时刻只有一个代次接单（以插件的事件处理日志证明）。
5) 残存旧进程（忽略断链、持有 runtime.lock）：不启动第二个写入者，插件 Degraded。
6) 空间不足（测试钩子使备份失败）：不激活，旧版不受影响。
7) 同插件并发操作（两个 CLI 同时 reload 与 disable）：串行执行，没有交错的 Journal。
8) 依赖：必需依赖缺失时拒绝启用并列出缺失项；停用被依赖的插件需要确认，确认后依赖方继续运行且调用返回 UNAVAILABLE；成环在安装时拒绝。
9) 回填 T-08：V-23（config.json 编辑在 reload 前不生效）、V-24（候选 OnStart 写库后失败不污染旧库）、V-25（secret 修改、旧 secret 字段转未知键、掩码不成为保存值）、V-39（external 插件的维护门槛）经正式入口验证。
10) 日志与 i18n（V-54、V-62、V-63）经正式入口的回归：日志级别修改、logs 查询、Core usage 的语言。
- scripts/scenarios：把 1)、4)、5) 写成只依赖 alonabot 二进制与测试包的场景，T-10 在其他平台运行。
- 证据：TEST-MATRIX §3.2 追加 T-09 一行；五平台的真实运行留给 T-10，写「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t09_e2e_test.go 与 scripts/scenarios。
- 本地自测（不提交，PR 中写明结果）：在 -race 下重复运行崩溃点场景若干次无偶发失败。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：五平台运行与发布（T-10）；新的产品功能（只做验收与必要的缺陷修复，修复需在 PR 中逐条说明）。
```

### goal-74 · T-14.4 首次运行与预装

状态：等待前置（T-09.6、T-14.3）。语句中 goal-39、goal-41、goal-73 提供的符号（manage.Service、serve、trust.Official、发行归档结构）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-74（切片 T-14.4：首次运行与预装）。goal 目录 goal-74，分支 goal-74，worktree .worktrees/goal-74，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：数据目录首次初始化时，按普通安装流程安装发行包 packages/ 中的全部包，只启用 help 与官方 WebUI；Agent 只安装不启用（D-50、D-62）；Telegram 与 OneBot 11 Adapter 只安装，管理员为它添加第一个实例时才启用；其他预装插件只安装不启用。之后不再自动安装，已卸载的预装包不重新安装；新版本发行包中版本更高的内置包作为可用更新列出，由管理员确认后按 update 流程更新。

依据：PLAN T-14 步骤 4；CONTRACTS §15.1、§15.2；TEST-MATRIX V-66。

范围（只新增或修改这些）：
- 根包 assembly.go：serve 首次初始化时的预装步骤（新增 preinstall.go，只在 Serve 中加调用点）
- internal/core/state/：记录「已完成首次初始化」的设置项（沿用 core_settings）
- internal/core/manage/：List 返回可用更新；CreateInstance 在实例所属的 Adapter 未启用时先启用它
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- 首次初始化：serve 取得 core.lock 后，若 core_settings 中没有 preinstalled 标记：查找 alonabot 可执行文件同目录下的 packages/（发行归档结构）；没有该目录时跳过（开发构建）。对其中每个包按 Install 的正常流程校验与安装；trust.Official 的公钥视为已信任来源，不逐个确认；签名或摘要校验失败的包不安装并记 ERROR，继续处理其余包。安装后只启用 help（按 manifest 的 id 识别）与 role 为 webui 的官方 WebUI，其余不启用。全部处理完后写入 preinstalled 标记（即使有包失败）。
- 首次管理 token：沿用 goal-39 的 serve 规则（交互终端下生成并显示一次，非终端拒绝启动）；预装在 token 初始化之后执行。
- 管理员之后可经管理 API 或 CLI 启用 Agent；Agent 在没有配置 Provider 时也能启用（/ask 回复提示）。
- 之后的启动不再自动安装；已卸载（uninstalled）的预装包不重新安装。
- 可用更新：List 时比较 packages/ 中每个包的版本与已安装版本，更高时在 Package 中标记 available_update（版本号），不自动更新；管理员经 update 选择该包。
- Telegram：CreateInstance 发现 Adapter 包未启用时先启用它，再创建实例。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/firstrun_e2e_test.go（//go:build unix）：用 goal-73 的 dist 工具以临时密钥为当前平台构建发行归档并解压到临时目录，运行其中的 alonabot serve（干净 data_dir，测试钩子模拟交互终端）：
  - 全部预装包已安装；只有 help 已启用并有进程；Agent、记忆、welcome、rss、Telegram 与 OneBot 11 Adapter 未启用，没有对应进程。
  - 非终端运行时拒绝启动并提示 admin-token init。
  - 把其中一个包的签名改坏：该包未安装并有 ERROR，其余正常。
  - 卸载 rss 后重启 serve：不重新安装。
  - 用版本号更高的发行归档替换 packages/ 后重启：list 显示可用更新，没有自动更新。
  - 创建 Telegram 实例时 Adapter 被启用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有 packages/ 时跳过；preinstalled 标记在部分失败时仍写入。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：容器镜像（T-14.5）；Core 自身的升级（替换可执行文件，由管理员操作）。
```

### goal-77 · T-14.5 容器镜像

状态：等待前置（T-07.3、T-14.3）。语句中 goal-25、goal-73 提供的符号（settings 的 listen 校验、scripts/dist）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-77（切片 T-14.5：容器镜像）。goal 目录 goal-77，分支 goal-77，worktree .worktrees/goal-77，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：仓库根目录的 Dockerfile 用多阶段构建产生 linux/amd64 与 linux/arm64 镜像，deploy/compose.yaml 是官方整合包的 compose 入口：Go 构建阶段以 CGO_ENABLED=0 构建 Core 与全部官方包，最终镜像只含 alonabot、packages/ 与 CA 证书，以 UID 10001 运行 alonabot serve --data-dir /data，/data 为卷，设置 ALONABOT_CONTAINER=1。Core 只在该环境变量下允许 0.0.0.0 的 listen 搭配本机 http 的 Base URL，并在启动时输出 WARN。deploy/compose.yaml 给出卷、init: true、127.0.0.1:7650:7650 与首次 admin-token init 的用法。

依据：PLAN T-14 步骤 5；CONTRACTS §15.3、§15.1、§7.4；DECISIONS D-36；TEST-MATRIX V-70 的镜像分项。

范围（只新增或修改这些）：
- 新增 Dockerfile、.dockerignore、deploy/compose.yaml
- internal/core/settings/：ALONABOT_CONTAINER=1 时 listen 缺省为 0.0.0.0:7650，并放宽「非 loopback 必须 https 或可信反代」的校验（只在该变量下），启动时输出 WARN "listening on all interfaces inside a container; publish the port to the host loopback only"
- docs/development.md：增加「容器镜像」一节（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- Dockerfile：构建阶段使用 go.mod 锁定的 Go 版本的官方镜像（标签写明版本与 digest）；以 scripts/dist 的同一逻辑为目标平台构建 alonabot 与官方包（签名密钥以 BuildKit secret 传入，不进入镜像层；没有密钥时构建失败并提示）；最终阶段为 scratch 加 CA 证书（从构建阶段复制 /etc/ssl/certs/ca-certificates.crt），USER 10001:10001，VOLUME /data，ENV ALONABOT_CONTAINER=1，ENTRYPOINT ["/alonabot"]，CMD ["serve", "--data-dir", "/data"]。镜像中没有 shell。
- compose.yaml：image 固定为对应版本（发布流程把 :<version> 写入随 Release 上传的副本）、init: true、volumes（命名卷到 /data）、ports "127.0.0.1:7650:7650"，注释说明首次运行前执行 docker compose run --rm alonabot admin-token init，再 docker compose up -d。
- settings：没有 ALONABOT_CONTAINER 时行为不变（0.0.0.0 搭配 http Base URL 仍被拒绝）。
- docs：写明构建命令（docker buildx build --platform linux/amd64,linux/arm64 --secret id=release_key,src=...）、首次初始化、端口只发布到宿主本机、远程访问用 SSH 隧道或 HTTPS 反代、MCP stdio 需要 shell 或其他运行时时须自建镜像。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/settings 的测试追加：有与没有 ALONABOT_CONTAINER 时 listen 缺省值与校验的差异，以及 WARN。
- 本地自测（不提交，PR 中写明结果）：在有 Docker 的机器上 buildx 构建两个平台（写明 Docker 版本）；干净卷无 TTY 启动被拒；-it admin-token init 后正常启动并能访问 http://127.0.0.1:7650；容器内进程的 UID 为 10001；镜像文件列表只有 alonabot、packages/ 与 CA 证书；出站 HTTPS 可用；docker stop 后插件进程全部退出。没有 Docker 环境时以上各项写「未运行」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/...；go build ./...；node scripts/check-repository.mjs。

不做：推送镜像到 ghcr.io（T-10.4）；Kubernetes 清单。
```

### goal-85 · T-14.7 install.sh

状态：待编写（前置 T-14.3）。依据 D-64 新增；T-14.3 合并后由审查者按实际的发行归档与 SHA256SUMS 写出范围和测试段。

目标：deploy/install.sh 按 CONTRACTS §15.6 在 Linux（systemd，root）与 macOS（launchd，当前用户）上安装 Core 与官方组件：下载并按 SHA256SUMS 校验发行归档，校验失败不改动已有安装；替换程序目录；尚无管理凭证时经 /dev/tty 运行 admin-token init，没有终端时不启动服务并输出初始化命令；写入 systemd 单元（系统用户 alonabot）或用户 LaunchAgent 并启动；再次运行即升级，`--uninstall` 保留数据目录；不支持的系统在改动前退出。Release 流程上传 install.sh 与 SHA256SUMS。

依据：PLAN T-14 步骤 7、§3.1 的 T-14.7 行；CONTRACTS §15.1、§15.2、§15.6；DECISIONS D-64；TEST-MATRIX V-84。

要点：
- 只依赖 POSIX sh、curl、tar 与 sha256sum/shasum，不引入新工具；用 shellcheck 检查，未安装时记为未运行。
- 下载地址的基址可由环境变量覆盖，供本地 Release 模拟服务测试；测试与真实安装都不写入开发机的系统目录之外的位置。
- V-84 需要真实的 systemd Linux 与 macOS；容器内的 systemd 与模拟环境的结果单独记录，不算真实结果。

## 第 16 波

### goal-75 · T-14.6 发行包与镜像端到端验收

状态：等待前置（T-09.11、T-14.4、T-14.5）。语句中 goal-43、goal-44、goal-73、goal-74、goal-77 提供的产物与入口（发行归档、容器镜像、WebUI 页面、管理 API）只写名字；合并后由审查者核对，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-75（切片 T-14.6：发行包与镜像的端到端验收）。goal 目录 goal-75，分支 goal-75，worktree .worktrees/goal-75，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：在干净的数据目录上，分别从发行归档与容器镜像依次完成：首次运行、登录 WebUI、添加 Telegram 实例（本地 Bot API 模拟服务）、启用 Agent 并配置 Provider（本地模拟服务）、/ask 对话、/help、启用 rss 并收到推送；在 TEST-MATRIX §3.2 记录 T-14 证据，真实 Telegram 与真实 Provider 没有凭证时记为未运行。

依据：PLAN T-14 步骤 6 与「停止和失败条件」；CONTRACTS §15；TEST-MATRIX V-66、V-67、V-70。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t14_e2e_test.go（//go:build unix；发行归档部分）
- 新增 deploy/e2e/（容器验收：Go 测试辅助，经 os/exec 调用 docker，运行镜像、模拟服务与断言）
- spec/TEST-MATRIX.md §3.2：追加 T-14 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

验收步骤（发行归档与容器镜像各跑一遍；每一步记录命令与结果）：
1) 首次运行：干净 data_dir；终端下初始化 token（容器用 -it admin-token init）；预装包状态符合 goal-74。
2) 登录：经管理 API 登录（等同页面登录），状态接口显示各包状态。
3) 添加 Telegram 实例：api_base_url 指向本地 Bot API 模拟服务（容器内经宿主地址访问），实例上线（bot.online）。
4) 配置 Provider：经配置 API 为 Agent 添加指向本地模拟 Provider 的条目（api_key 为 secret），reload 生效。
5) /ask：模拟服务中的用户发送 /ask，收到先发送再编辑的回复，最终内容正确。
6) /help 与 /whoami：列出可用命令；/whoami 返回实例与用户 ID；把该用户加入 Core Admin。
7) 启用 rss：Admin /rss add 一个测试 HTTP 源，授予发信许可，源新增条目后收到推送。测试源必须使用 rss 地址检查允许的地址（非 loopback、非私网）；做不到时说明实际做法（例如测试钩子），并在证据中标明。
8) 停止：Core 与全部插件进程退出（容器 docker stop 后进程全部消失）。
- V-70：镜像文件只含 alonabot、packages/ 与 CA 证书，UID 10001；无 TTY 时拒绝启动；端口只发布到宿主本机；出站 HTTPS 可用；没有 ALONABOT_CONTAINER 时 0.0.0.0 搭配 http Base URL 被拒绝。
- 证据：TEST-MATRIX §3.2 追加 T-14 一行（平台、Docker 版本、命令、结果）；真实 Telegram、真实 Provider 与没有运行的平台写「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t14_e2e_test.go；deploy/e2e。
- 本地自测（不提交，PR 中写明结果）：在有 Docker 的机器上运行容器部分；没有 Docker 时写「未运行」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs。

不做：五平台运行（T-10.2）；发布（T-10.4）；新的产品功能（只做验收与必要修复，修复逐条说明）。
```

## 第 17 波

### goal-46 · T-10.1 证据汇总与跨组件联验

状态：等待前置（T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.6、T-14.7、T-15.3、T-16.2）。本切片只汇总与联验，不引入新接口。

```text
/goal 使用 $alonabot-execute 推进 goal-46（切片 T-10.1：证据汇总与跨组件联验）。goal 目录 goal-46，分支 goal-46，worktree .worktrees/goal-46，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：汇总 V-01～V-76 与 V-82～V-84 的证据，逐项标记已通过、失败、未运行或待联验；补齐组件之间的联验（首次管理登录 → 安装启用 → 配置/权限 → 真实消息 → 插件 HTTP → 重载/升级/回退，TCP 与 stdio）；用不同的真实插件二进制覆盖各能力组合；执行协议、权限、资源边界与完整恢复的故障注入，核对数据库内容与权威指针。

依据：PLAN T-10 步骤 1～4 与「停止和失败条件」；TEST-MATRIX 全部；CONTRACTS §12.4。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t10_e2e_test.go（//go:build unix）与必要的 testdata 插件
- spec/TEST-MATRIX.md：§3 的每个 V 条目更新为最新状态，新增「发布前证据汇总」一节（本切片获授权修改 TEST-MATRIX；其他 spec 文件只在发现契约缺陷时按 spec/README 的流程修改，并在 PR 中说明）
- 发现的缺陷：在对应包中修复，每项修复在 PR 中写明 V 编号、复现、原因与验证
允许的新跨模块边：无。

联验矩阵（各为一个子测试，TCP 与 stdio 各一遍）：
1) 组件链路：admin-token init → 登录 → 安装并启用 echo、configdb、http 示例与官方插件 → 配置与权限 → Telegram 模拟服务的真实消息 → 插件 HTTP 页面（经票据）→ reload → update → rollback。
2) 能力组合（每个能力使用不同的插件二进制，不用同一个 worker 的多份进程代替）：命令、文本匹配、多轮等待、主动 Send、Service、Event Bus、Scheduler、HTTP、GORM（含迁移），以及 Agent 与 WebUI 组件；既有无配置无数据库的插件，也有实际写入与迁移的插件。
3) 边界：伪造与过期 Session、命令遮蔽、无 Actor、实例许可、Hook 递归、媒体租约、满准入、Cancel 顺序、超时仍运行的 job；错误回执与已确认的外部结果一致。
4) 恢复：停止与持锁、SQLite WAL 备份、候选 OnStart 写入后失败、PREPARED 与提交前后 Core 退出、不可变 secret、external 插件的维护许可与 resume、显式 rollback 的新代次；核对数据库内容与状态库的 Active 指针，不能只看日志里的「恢复成功」。
- 汇总：TEST-MATRIX 中每个 V 条目写明测试位置、命令、平台与结果；失败的列为阻塞项；没有运行条件的（真实 Telegram、真实 Provider、非 darwin 平台等）写「未运行」与原因。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t10_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：完整测试套件在 -race 下运行一次的总耗时与结果。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；node scripts/check-repository.mjs；TEST-MATRIX 中没有「待联验」的条目。

不做：五平台运行（T-10.2）；资源报告（T-10.3）；发布（T-10.4）。
```

## 第 18 波

### goal-47 · T-10.2 五平台构建与运行

状态：等待前置（T-10.1）。需要真实目标平台；没有运行环境的组合保持未运行。

```text
/goal 使用 $alonabot-execute 推进 goal-47（切片 T-10.2：五平台构建与运行）。goal 目录 goal-47，分支 goal-47，worktree .worktrees/goal-47，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：在 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64 五个组合上分别构建并运行 Core 与插件，覆盖启动、停止、至少 100 次启停、异常退出、合作型子进程树与端口清理；记录 CGO 插件的实际支持范围。

依据：PLAN T-10 步骤 5；CONTRACTS §7.1（平台与进程树）、§12.2；TEST-MATRIX V-32、V-34。

范围（只新增或修改这些）：
- 新增 scripts/platform/（可在各平台运行的验收程序：Go 程序，调用构建好的 alonabot 与测试插件，输出结构化结果）
- 复用 goal-45 的 scripts/scenarios 在各平台运行
- spec/TEST-MATRIX.md：V-32、V-34 与各平台相关条目的证据（本切片获授权修改 TEST-MATRIX）
- 发现的平台缺陷：在对应包中修复（例如 Windows Job Object、路径与锁），每项在 PR 中写明平台、复现与验证
允许的新跨模块边：scripts/platform 只经 os/exec 调用构建产物，不 import 本 module 的包。

行为：
- 每个平台：用 scripts/dist 的构建产物（或在该平台本地构建）运行 scripts/platform：启动 serve、安装并启用 echo 与 http 示例、收发一条消息、访问一次插件 HTTP、停止；再循环 100 次启停（每次检查插件进程与监听端口都已释放）；强制结束 Core（Unix 为 SIGKILL，Windows 为 TerminateProcess）后重启，确认插件进程已退出且能恢复；启动一个会派生子进程的测试插件，停止后子进程消失（合作型进程树）。
- 结果写成 JSON（平台、Go 版本、OS 版本、各步骤的结果与耗时），附在 PR 中，摘要写入 TEST-MATRIX。
- CGO：用一个 CGO 测试插件在各平台尝试构建与运行，记录实际支持范围；不支持的平台标为不支持，不发布。
- 没有某个平台的运行环境时，该平台写「未运行」；交叉编译成功不能代替运行结果。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/platform（本身就是验收程序）；平台缺陷修复所需的集成测试。
- 运行记录：每个平台的 JSON 结果附在 PR 中。

完成条件：gofmt -l 无输出；go vet ./...（含 GOOS=windows、GOOS=linux 的 vet）；go test -race ./...（当前平台）；每个可用平台的 scripts/platform 运行结果。

不做：资源报告（T-10.3）；发布（T-10.4）。
```

### goal-48 · T-10.3 资源比较报告与常驻内存预算核对

状态：等待前置（T-10.1）。需要固定机器与等价负载；NoneBot2 侧的 Python 依赖在引入前须经维护者确认。

```text
/goal 使用 $alonabot-execute 推进 goal-48（切片 T-10.3：资源比较报告与常驻内存预算核对）。goal 目录 goal-48，分支 goal-48，worktree .worktrees/goal-48，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：在一台固定机器上，按 TEST-MATRIX V-33 的四种负载（空闲、消息、数据库、HTTP）测量 Core 加插件及附属进程的总 RSS/私有内存、Linux PSS、CPU 与 P50/P95/P99 延迟，并用同一模拟服务与功能相同的插件测量 NoneBot2；再按 V-74 测量 Core、echo、rss 与 configdb 的 gormdb 变体空闲时的 PSS 并对照预算。输出原始数据与复现命令，不预设优胜结论。

依据：PLAN T-10 步骤 6；CONTRACTS §10；DECISIONS D-52；TEST-MATRIX V-33、V-74。

范围（只新增或修改这些）：
- 新增 bench/（负载生成器与采样程序：Go 程序）；bench/nonebot2/（NoneBot2 侧的等价插件与运行说明，Python 依赖用 uv 管理且只在该目录内）
- 新增 docs/benchmarks.md（报告：机器、版本、命令、原始数据位置、结果表，不写优胜结论）
- spec/TEST-MATRIX.md：V-33、V-74 的证据（本切片获授权修改这一处）
允许的新跨模块边：bench 只经 os/exec 与 HTTP 调用被测程序。NoneBot2 及其 Python 依赖只用于基准，不进入 Go module；在 bench/nonebot2 中引入前标阻塞，列出版本，等维护者确认。

行为：
- 负载（V-33）：空闲——Core 与 echo、configdb、HTTP 示例三个插件运行 10 分钟；消息——每秒 20 条文本、持续 10 分钟，echo 逐条回复；数据库——每秒 20 条命令，每条在 SQLite 中写入并读取一行，持续 10 分钟；HTTP——每秒 50 个请求访问插件路由，持续 10 分钟。两边都使用同一个本地 Bot API 模拟服务。
- 采样：每秒采样一次全部相关进程（Core、插件、Adapter 及其子进程；NoneBot2 的进程）的 RSS、私有内存（Linux 读 smaps_rollup 的 Private_*；macOS 写明使用的替代指标）、Linux PSS、CPU；延迟为模拟服务从发出 update 到收到回复的时间（消息与数据库负载），HTTP 为请求往返时间。
- 输出：每次运行的原始 CSV 与汇总 JSON，复现命令写入 docs/benchmarks.md；报告只给数值与条件，不写结论。
- V-74：linux/amd64，各进程 Activated 后空闲 2 分钟，读 smaps_rollup 的 Pss，取 3 轮中位数；超出预算时在报告中列出超出项与可见原因（例如新增依赖、init 分配），不自行调整预算。
- TEST-MATRIX 记录机器、OS、版本与结果文件位置；没有在 Linux 上运行时 PSS 写「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：基准程序本身是交付物，不另提交测试；负载生成器的本地自测在 PR 中写明。
- 运行记录：原始数据附在 PR 中（是否提交到 bench/results/ 由维护者决定）。

完成条件：gofmt -l 无输出；go vet ./...；go build ./...；一次完整基准运行的结果；node scripts/check-repository.mjs。

不做：性能优化（发现问题时另行处理）；修改 V-74 的预算（需维护者在 DECISIONS 记录）；为吞吐与延迟设定通过阈值（契约不设）。
```

## 第 19 波

### goal-49 · T-10.4 发布资产与说明

状态：等待前置（T-10.2、T-10.3）。正式签名与发布由维护者在其控制的环境中执行。

```text
/goal 使用 $alonabot-execute 推进 goal-49（切片 T-10.4：发布资产与说明）。goal 目录 goal-49，分支 goal-49，worktree .worktrees/goal-49，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：生成五平台候选资产（含官方 Telegram 与 OneBot 11 Adapter、Agent、记忆与预装插件）、T-14 的发行归档与容器镜像、外部 release.json 与签名；用干净的测试 data_dir 从这些真实资产完成安装、更新与卸载重装，核对来源、长度、摘要、平台与 Schema/二进制一致性，确保安装不依赖开发机上未打包的文件；关闭全部阻塞用例，写出发布说明。签名私钥只在维护者控制的环境中使用；发布到 GitHub Release 与推送镜像由维护者执行。

依据：PLAN T-10 步骤 7、8；CONTRACTS §9.2、§15.2、§15.3；TEST-MATRIX V-28、V-41 的发布复验。

范围（只新增或修改这些）：
- 新增 docs/release-notes/v0.1.0.md（版本、平台、依赖、已知限制、未运行项、恢复步骤、升级与回退说明）
- 新增 scripts/release-check/（用候选资产在干净 data_dir 完成安装、更新、卸载重装与一致性核对的 Go 程序）
- spec/TEST-MATRIX.md：V-28、V-41 的发布复验证据，以及「发布前证据汇总」的最终状态（本切片获授权修改 TEST-MATRIX）
- README.md：安装与快速开始指向发布资产（本切片获授权修改 README 的这一节）
允许的新跨模块边：scripts/release-check -> contracts/release（ARCHITECTURE 已登记，与 scripts/dist 同类）；其余只经 os/exec 调用候选资产。

行为：
- 候选资产：执行者用测试密钥运行 scripts/dist 与镜像构建，得到与正式发布结构相同的候选资产；正式签名由维护者用官方私钥重新运行同一命令完成。
- release-check：在干净 data_dir 中只使用候选资产（不访问源码树）：首次运行预装、安装一个第三方测试包、用新版本资产 update、卸载并同来源重装；对每个包核对 release.json 的来源、长度、SHA256、平台与包内 Schema、manifest、二进制的一致性；签名错误、来源不符的资产被拒绝（V-28、V-41 复验）。
- 发布说明：写明支持的平台与实际运行结果（来自 T-10.2）、资源报告位置（T-10.3）、全部「未运行」项与原因、已知限制（Native 插件是可信代码、库模式无隔离、插件自管的外部数据不承诺自动回退等）、备份与恢复步骤、升级与回退方法。
- 阻塞用例：TEST-MATRIX 中标为失败的条目必须全部修复或经维护者同意延后（在发布说明中列出）。
- 发布：维护者在其环境中用 scripts/dist 构建并签名，把发行归档、各安装包及其 release.json/.sig 发布到本仓库的 GitHub Release，把镜像推送到 ghcr.io/alonabot/alonabot（:<version> 与 :latest）；执行者不持有私钥，不执行发布。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/release-check（验收程序）。
- 运行记录：release-check 对候选资产的运行结果附在 PR 中。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；release-check 对候选资产通过；node scripts/check-repository.mjs；TEST-MATRIX 没有未处理的失败条目。

不做：执行正式签名与发布（维护者负责）；v0.1 之后的路线图。
```
