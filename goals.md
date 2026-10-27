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
| goal-76 | T-02.2 | 调优参数 | T-02.1 | 可派发 |
| goal-2 | T-03.1 | 首个可用 SDK 切片 | T-02.1、T-02.2 | 等待前置 |
| goal-3 | T-03.2 | Target 与发信许可 | T-03.1 | 等待前置 |
| goal-4 | T-03.3 | Adapter 进程、sdk/adapter 与上下线事件 | T-03.1 | 等待前置 |
| goal-6 | T-03.5 | 媒体上传与配额 | T-03.1 | 等待前置 |
| goal-10 | T-04.1 | 根插件 API 与静态注册索引 | T-03.1 | 等待前置 |
| goal-56 | T-11.3 | Agent 会话存储 | T-03.1 | 等待前置 |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 等待前置 |
| goal-11 | T-04.2 | 构建期生成器 | T-04.1 | 等待前置 |
| goal-12 | T-04.3 | 命令选择、文本匹配与参数绑定 | T-04.1 | 等待前置 |
| goal-13 | T-04.4 | 持久化命令策略 | T-03.2、T-04.1 | 等待前置 |
| goal-19 | T-06.1 | Service/Bus 描述与 Broker 准入 | T-04.1 | 等待前置 |
| goal-25 | T-07.3 | SDK HTTP Listener 与 Core 代理 | T-04.1 | 等待前置 |
| goal-7 | T-03.6 | 入站媒体与 Grant | T-03.2、T-03.4、T-03.5 | 等待前置 |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 等待前置 |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 等待前置 |
| goal-21 | T-06.3 | 非持久 Event Bus | T-06.1 | 等待前置 |
| goal-26 | T-07.4 | 两个 Host 与插件票据 | T-07.3 | 等待前置 |
| goal-28 | T-08.1 | 配置字段描述与构建检查 | T-04.2 | 等待前置 |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-06.1 | 等待前置 |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2 | 等待前置 |
| goal-8 | T-03.8 | 按钮、extension 片段与依赖解析 | T-03.2、T-03.7 | 待编写 |
| goal-51 | T-03.9 | 消息编辑 | T-03.7 | 待编写 |
| goal-16 | T-05.2 | Rule、HandlerOutcome 与传播 | T-05.1 | 待编写 |
| goal-24 | T-07.2 | schedule 注册与运行槽 | T-05.1、T-07.1 | 待编写 |
| goal-30 | T-08.3 | SDK 配置准备顺序 | T-08.1、T-08.2 | 待编写 |
| goal-31 | T-08.4 | 不可变 Secret Store | T-08.1、T-08.2 | 待编写 |
| goal-34 | T-08.7 | web/admin 骨架、配置表单组件与界面文案 | T-08.1 | 待编写 |
| goal-53 | T-08.8 | 用户可见文案目录 | T-04.1、T-04.3、T-08.1 | 待编写 |
| goal-36 | T-09.1 | 首次管理认证 | T-07.4 | 待编写 |
| goal-69 | T-13.3 | 热部署 | T-13.1 | 待编写 |
| goal-9 | T-03.10 | Telegram 接入开发入口与故障矩阵 | T-03.6、T-03.8、T-03.9 | 待编写 |
| goal-17 | T-05.3 | 三阶段 Hook 与 send_forbidden | T-05.2 | 待编写 |
| goal-52 | T-05.4 | 延续调用与多轮等待 | T-05.2 | 待编写 |
| goal-27 | T-07.5 | 调度与 HTTP 启停集成及验收 | T-07.2、T-07.4 | 待编写 |
| goal-32 | T-08.5 | GORM 连接与迁移 | T-08.3 | 待编写 |
| goal-39 | T-09.4 | 管理服务、安装记录与 serve | T-03.8、T-08.4、T-09.1、T-09.2、T-13.1 | 待编写 |
| goal-71 | T-14.1 | help 插件与命令列表 | T-04.4、T-05.2、T-08.8 | 待编写 |
| goal-14 | T-04.5 | 示例升级与 T-04 验收 | T-03.10、T-04.2、T-04.3、T-04.4 | 待编写 |
| goal-18 | T-05.5 | 多进程事件流水线验收 | T-03.8、T-05.3、T-05.4 | 待编写 |
| goal-20 | T-06.2 | 父调用链、取消与背压 | T-03.6、T-05.3、T-06.1 | 待编写 |
| goal-33 | T-08.6 | SQLite 候选与资源交接 | T-08.4、T-08.5 | 待编写 |
| goal-59 | T-12.2 | Agent 对话与流式显示 | T-03.9、T-04.3、T-05.4、T-08.3、T-08.4、T-11.3、T-12.1 | 待编写 |
| goal-68 | T-13.2 | 开发组件的配置与数据库 | T-08.3、T-08.5、T-13.1 | 待编写 |
| goal-72 | T-14.2 | welcome 与 rss 插件 | T-05.2、T-07.2、T-08.3、T-08.5、T-08.8 | 待编写 |
| goal-22 | T-06.4 | Session 清理与 Service 端到端 | T-06.2、T-06.3 | 待编写 |
| goal-35 | T-08.9 | configdb 示例与 T-08 验收 | T-08.6、T-08.7、T-08.8 | 待编写 |
| goal-38 | T-09.3 | 开发 CLI 与库模式脚手架 | T-04.5、T-08.1、T-08.8、T-09.2、T-13.1 | 待编写 |
| goal-55 | T-11.2 | 工具调用与用户确认 | T-03.8、T-06.2、T-08.8、T-11.1 | 待编写 |
| goal-70 | T-13.4 | 库模式示例与验收 | T-03.10、T-13.2、T-13.3 | 待编写 |
| goal-40 | T-09.5 | 候选 Journal 与准备编排 | T-09.4、T-08.6、T-06.4 | 待编写 |
| goal-57 | T-11.4 | 工具与会话的验收 | T-11.2、T-11.3 | 待编写 |
| goal-60 | T-12.3 | 工具循环与 MCP | T-11.2、T-12.2 | 待编写 |
| goal-61 | T-12.4 | 记忆插件 | T-08.5、T-11.2、T-12.2 | 待编写 |
| goal-41 | T-09.6 | 唯一提交点与激活门禁 | T-07.5、T-09.5 | 待编写 |
| goal-62 | T-12.5 | 技能 | T-12.3 | 待编写 |
| goal-63 | T-12.6 | 定时 Agent 任务 | T-07.2、T-08.5、T-12.3 | 待编写 |
| goal-42 | T-09.7 | 崩溃恢复、回退与回收 | T-09.6 | 待编写 |
| goal-43 | T-09.8 | Admin 管理页 | T-09.6、T-08.7 | 待编写 |
| goal-64 | T-12.7 | 官方插件打包与 T-12 验收 | T-12.4、T-12.5、T-12.6 | 待编写 |
| goal-44 | T-09.9 | Admin 配置、权限、工具策略与插件页 | T-09.8、T-11.1、T-11.3 | 待编写 |
| goal-66 | T-09.10 | 日志管理与日志页 | T-02.1、T-03.1、T-09.8 | 待编写 |
| goal-73 | T-14.3 | 发行归档 | T-03.10、T-09.3、T-12.7、T-14.1、T-14.2 | 待编写 |
| goal-45 | T-09.11 | 正式流程故障注入与 T-09 验收 | T-09.3、T-09.7、T-09.9、T-09.10 | 待编写 |
| goal-74 | T-14.4 | 首次运行与预装 | T-09.6、T-14.3 | 待编写 |
| goal-77 | T-14.5 | 容器镜像 | T-07.3、T-14.3 | 待编写 |
| goal-75 | T-14.6 | 发行包与镜像端到端验收 | T-09.11、T-14.4、T-14.5 | 待编写 |
| goal-46 | T-10.1 | 证据汇总与跨组件联验 | T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.6 | 待编写 |
| goal-47 | T-10.2 | 五平台构建与运行 | T-10.1 | 待编写 |
| goal-48 | T-10.3 | 资源比较报告 | T-10.1 | 待编写 |
| goal-49 | T-10.4 | 发布资产与说明 | T-10.2、T-10.3 | 待编写 |

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

目标：一个纯计算包：校验 interval 与五字段 cron（含 IANA 时区），并给出下一次触发时间。Core Scheduler 与官方 Agent 插件的定时任务都使用它。不接入 REGISTER，不触发任何调用。

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

状态：已完成（PR #16 已合并）。本 goal 的原始语句未保留；下面记录合并后的实际接口，供后继切片引用。

目标：Provider 客户端：OpenAI 兼容与 Anthropic Messages 的流式请求、SSE 解析、工具调用解析、本地模拟服务
前置：T-02。
依据：PLAN §3.1 的 T-12.1 行、§4 的 T-12 步骤 1（接口与流语义）；验收：V-57 的 Provider 分项。

接口（main 上的实际代码）：

```go
// plugins/agent/provider
type Protocol string // OpenAICompatible = "openai-compatible"、AnthropicMessages = "anthropic-messages"
type Config struct { Protocol Protocol; BaseURL string; APIKey string; Model string }
type Message struct { Role string; Content string; ToolCalls []ToolCall; ToolResults []ToolResult } // Role：system、user、assistant
type Tool struct { Name string; Description string; InputSchema json.RawMessage }
type Request struct { Messages []Message; Tools []Tool; MaxTokens int } // MaxTokens 必须大于 0
type ToolCall struct { ID string; Name string; Arguments json.RawMessage }
type ToolResult struct { ToolCallID string; Content string }
type Event struct { Text string; ToolCall *ToolCall } // ToolCall 非 nil 时是组装完整的一次工具调用
func New(config Config, httpClient *http.Client) (*Client, error) // httpClient 由调用方注入
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) error
// Stream 不重发请求；已交给 emit 的文本增量由调用方保留；错误不含 API key 与认证头。
```

## 第 2 波

### goal-76 · T-02.2 调优参数

状态：可派发。

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

状态：等待前置（T-02.2）。日志接口已按 T-02.1 的实现写入，SDK 根插件已按 sdk.Base 与 Root Hook 的现行模型修订，Router 队列容量改为取 goal-76 的 tune.Values.DispatchQueueEvents；调优参数合并后，审查者按实际的参数接口补充本语句，再改为可派发。

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
   func New(log *slog.Logger, queueEvents int) *Router // log 为 Core 的 Logger（logging.Output.Logger）；queueEvents 取 tune.Values.DispatchQueueEvents
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
  - 响应解码为 HandlerOutcome：propagation=STOP 时停止后续候选；调用错误或带 Err 的响应以 WARN 记录（带 plugin_id、route_id、trace_id）后继续下一个候选。不重试。
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

状态：等待前置（T-03.1）。语句中 goal-2 提供的符号（router.Invocation、message.Service、currentSession）在 goal-2 合并后由审查者核对签名，再改为可派发。

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

状态：等待前置（T-03.1）。语句中 goal-2 提供的符号（adapter.Instance、adapter.Sink、adapter.Registry、router.Mux、router.Caller、wire.ServiceSend 等）只写名字；goal-2 合并后，审查者按实际签名核对，再改为可派发。

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
- 新增 internal/core/adapter/remote/
- sdk/：新增 adapter_run.go（RunAdapter）；sdk/events.go 的事件类型白名单加入 bot.online、bot.offline
- 新增 sdk/adapter/
- 上述包的测试文件
允许的新跨模块边：无（internal/core/adapter/remote 在 internal/core 内；sdk -> sdk/adapter 属同一模块）。sdk/adapter 只 import contracts，不 import sdk，避免循环。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
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

状态：等待前置（T-03.1）。配额与期限已按 T-02.2 的调优参数表改为从 tune.Values 取值；语句中 goal-2 提供的符号（router.Mux、router.Caller、currentSession）在 goal-2 合并后由审查者核对签名，再改为可派发。

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
// internal/wire/services.go 追加
const ServiceMediaBeginUpload = "core.media.begin_upload"
const ServiceMediaWriteChunk  = "core.media.write_chunk"
const ServiceMediaCommitUpload = "core.media.commit_upload"
// internal/core/media
type Limits struct { MaxFileBytes, MaxSessionBytes, MaxCoreBytes int64; MaxUploads int; TTL time.Duration }
func NewLimits(v tune.Values) Limits // 取 MEDIA_MAX_OBJECT、MEDIA_SESSION_BUDGET、MEDIA_DISK_BUDGET、MEDIA_MAX_UPLOADS、MEDIA_TTL 的生效值
const MaxChunkBytes = 1 << 20 // 协议固定的分块上限（CONTRACTS §1.3），不是调优参数
type Store struct{ ... }
func Open(dir string, lim Limits, now func() time.Time) (*Store, error) // dir 是 <data_dir>/media；打开时清空上次遗留的临时文件
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

状态：等待前置（T-03.1）。语句中 goal-2 提供的符号（sdk.EventRegistry、router.Router、router.Plugin、router.FromInstance 等）只写名字；goal-2 合并后，审查者按实际签名核对，再改为可派发。

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

状态：等待前置（T-03.1）。语句中 goal-2 提供的符号（router.Mux、router.Caller、wire.ServiceSend、sdk 的 currentSession、memory Adapter 与事件路由）只写名字；goal-2 合并后，审查者按实际签名核对，再改为可派发。

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

不做：管理 API、CLI 与页面（T-09.4、T-09.9）；卸载流程本身（T-09.4）；Agent 插件对会话的使用（T-12.2）；记忆服务（T-12.4）。
```
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

依据：PLAN T-08 步骤 1；CONTRACTS §5.1 第 2、3 条、§5.3（secret 路径）；WIRE-FORMAT §3.2（config_schema_hash）；DECISIONS D-40、D-43；TEST-MATRIX V-25 的描述分项。

范围（只新增或修改这些）：
- go.mod、go.sum：新增 github.com/google/jsonschema-go（D-43 已批准；取当前最新版本，PR 中写明版本）
- sdk/generate/：新增 config.go（Config 类型分析与 Schema 构造）；生成文件嵌入 Schema
- sdk/：generated.go 的 Generated 增加 ConfigSchema；session.go 的 buildRegister 写入 config_schema_hash
- 新增 sdk/config/schema.go（Schema 的规范序列化与摘要，供生成器、SDK 与 Core 共用）
- 上述包的测试文件
允许的新跨模块边：sdk/generate -> sdk/config（同属 sdk 模块）。jsonschema-go 只在 sdk/config 与 sdk/generate 中 import。

标签语法（本切片定下，写在 sdk/config 的包文档中）：
- form:"key=value;key=value;flag"，键：label（i18 键）、widget（text、password、textarea、number、switch、select）、options（以 | 分隔的取值）、min、max（数字）、default（JSON 字面量）、required（无值的标志）。未知键是诊断错误。
- description:"i18 键"；secret:"true"；hidden:"true"。
- 字段的 i18 路径为 config.<json 路径>（嵌套结构用 . 连接），未写 label 时以它为 label 键。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/config
   func MarshalSchema(s *jsonschema.Schema) ([]byte, error) // 规范 JSON：PropertyOrder 保持声明顺序，其余键排序，无多余空白
   func SchemaHash(canonical []byte) []byte                // SHA256
   const (
       KeywordSecret = "x-alonabot-secret" // bool
       KeywordHidden = "x-alonabot-hidden" // bool
       KeywordForm   = "x-alonabot-form"   // 对象：label、widget、options、min、max
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
- 诊断（带位置，继续检查其余字段）：上述之外的类型（interface、chan、func、complex、非 string 键的 map、time.Time 等）；类型实现了 json.Marshaler/json.Unmarshaler/encoding.TextMarshaler/TextUnmarshaler；递归类型；json 标签为 "-" 之外的重复名；default 不能解码为字段类型或超出整数范围、不在 options 中、超出 min/max；min > max；options 用于非 string/数字字段；widget 与类型不符；secret 用于非 string 字段。类型分析使用 go/types（以源码方式导入扫描到的包），不用反射。
- Schema 的根为 object，$schema 为 draft 2020-12；secret、hidden、form、i18 路径分别写入对应 Extra 关键字；default 写入 Default；不使用 jsonschema.For 推导（D-43）。
- 生成文件把规范 Schema 作为字节常量嵌入，经 Generated.ConfigSchema 交给 SDK；同样输入生成的字节完全相同。plugin check 在 Schema 变化时报告生成文件过期（沿用 goal-11 的比较）。
- 生成的 Schema 必须能被 jsonschema 的 Resolve（开启 ValidateDefaults）通过；不通过时作为诊断报告。

测试（按编码规范「测试提交规则」）：
- 提交：sdk/generate/config_test.go。沿用 goal-11 的临时 module 方式，写一个含嵌套 struct、指针、slice、map、secret、hidden、required、default、options 的 Config：
  - Result.ConfigSchema 的规范 JSON 与期望一致（字段顺序为声明顺序，扩展关键字正确，无 additionalProperties: false）。
  - 生成后的程序启动时（测试入口打印 REGISTER 内容后退出），config_schema_hash 等于 SchemaHash(ConfigSchema)。
  - 第二次生成逐字节相同。
  - 不支持的类型、实现 json.Unmarshaler 的类型、递归类型、错误 default、越界 default、未知 form 键各产生一条带位置的诊断。
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

不做：core.tool.invoke、确认与确认按钮（T-11.2）；Agent 插件与 MCP（T-12）；管理 API 与页面（T-09.9）。
```
### goal-67 · T-13.1 库模式入口与同进程组件

状态：等待前置（T-02.1、T-03.3、T-04.2）。语句中 goal-2、goal-4、goal-11、goal-76 提供的符号（router、message.Service、remote.Process、supervisor、sdk.Run 与 sdk.RunAdapter 的内部流程、sdk.Generated、tune.Parse）只写名字；合并后由审查者核对签名，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-67（切片 T-13.1：库模式入口与同进程组件）。goal 目录 goal-67，分支 goal-67，worktree .worktrees/goal-67，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：根包 alonabot 提供 Run(opts ...Option) 与 Plugin、Adapter、DataDir、Args 选项：在当前进程装配一个 Core，已安装的包照常以独立进程运行，开发中的插件与 Adapter 在 goroutine 中经 loopback TCP 走完整的握手、注册、激活与停止。组件身份取自生成器嵌入的 manifest；Core 为它们创建 dev 来源的安装记录与固定代次 dev。收到 SIGINT/SIGTERM 时按停止预算停止；组件在预算内没有退出时 Run 返回错误。

依据：PLAN T-13 步骤 1、2；CONTRACTS §15.4、§4.2、§9（manifest）、§16；ARCHITECTURE「根包」一行；TEST-MATRIX V-64 的生命周期分项。

范围（只新增或修改这些）：
- 新增根包文件：alonabot.go（Run 与选项）、assembly.go（Core 装配，供 Run 使用，之后 T-09.4 的 serve 复用）
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
   func Run(opts ...Option) error
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
- Run 的顺序：解析 DataDir 与 Args（ALONABOT_ARGS 在前），出错立即返回；按 goal-76 的规则建立 tune.Values；打开日志与状态库（取得 core.lock）；装配 Router、message.Service、Mux、adapter.Registry 与 remote；按安装记录启动全部已启用的已安装包（独立进程，沿用 supervisor）；再启动开发组件；等待信号。
- 开发组件：
  - 从 Generated.Manifest 解析 manifest（contracts/manifest），id 即插件 ID，role 决定 Plugin 还是 Adapter；Plugin(...) 传入 Adapter 角色的包或反之是错误。与已安装包的 ID 冲突时返回错误，不覆盖安装记录。
  - 创建或复用 source 为 dev 的安装记录（entrypoint 为空，build_digest 为本次运行生成的随机值），代次固定为 "dev"：首次运行 AddGeneration("dev") 并提交；以后复用。dev 来源不走候选与 Journal。
  - Core 为每个组件生成 Session Secret，经 inproc.Launch 交给 RunInProcess/RunAdapterInProcess 在 goroutine 中运行，连接 Core 的 loopback 端点；握手、Bootstrap、注册、激活与停止与独立进程相同。runtime.lock 路径为 <data_dir>/plugins/<id>/runtime.lock。
  - 配置：本切片 Bootstrap 的候选配置为 "{}"（config.json 的读取与 dev.sqlite 属于 T-13.2）。开发 Adapter 的实例：读取 <data_dir>/adapters/<adapter_id>/instances/*.json，文件名（去掉 .json）为 instance_id，内容原样作为实例配置，启动时全部 StartInstance；Secret 字段的抽取在 Secret Store 接入后处理，本切片用一行 shortcut 注释写明。
- 停止：收到 SIGINT/SIGTERM 时先停开发组件、再停已安装包，复用 host 的停止预算（STOP_TOTAL）。开发组件的 goroutine 在预算内没有返回时记 ERROR（插件 ID），Run 返回错误；已安装包按原流程停止。
- 开发组件没有进程隔离：SDK 未捕获的 panic 使整个进程退出，Core 不做有限重启；RunInProcess 不 recover 生命周期之外的 panic。
- 生成器：构建入口目录下存在 plugin.yaml 时把原文作为字节常量嵌入 Generated.Manifest；不存在时不嵌入（独立进程运行时不需要）。Adapter 包由 //alonabot:adapter 标记其 Adapter 实现类型（语法与诊断沿用 goal-11），生成 UseGeneratedAdapter。

测试（按编码规范「测试提交规则」）：
- 提交：alonabot_test.go（根包，//go:build unix）。测试构建 examples/devbot 并以子进程运行（临时 data_dir，预先写入一个回环 Adapter 实例文件），只通过日志、状态库与进程退出码断言：
  - 启动后日志中依次出现开发插件与开发 Adapter 的激活记录，代次为 dev，状态库中两条 source=dev 的安装记录。
  - 回环实例启动后交付一条 "ping" 的 message.created；开发插件的事件路由用 Context.Send 回复 "pong"（回复原 Target，不需要许可），回环 Adapter 收到发送后记录日志；测试从日志断言这次往返完成。
  - 再次启动复用同一 dev 代次，不新增代次。
  - 发送 SIGTERM：进程在 STOP_TOTAL 内以 0 退出。
  - 用 Args 把 STOP_TOTAL 设得很小、开发插件的 OnStop 阻塞：SIGTERM 后 Run 返回错误、进程以非 0 退出，日志有该插件的 ERROR。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：缺少 Generated 或 Manifest 时的错误；Plugin 传入 Adapter 角色；与已安装包 ID 冲突；ALONABOT_ARGS 与 Args 的先后；instances 目录不存在时没有实例；inproc.Launch 为 nil。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./... ；go build ./...。

不做：config.json 直读、dev.sqlite 与拒绝 update/rollback（T-13.2）；热部署与 Reload（T-13.3）；alonabot serve（T-09.4，复用本切片的装配）；HTTP 与设置文件的装配（随 T-07.5、T-09.4 接入）。
```
### goal-8 · T-03.8 按钮、extension 片段与依赖解析

状态：待编写。

目标：按钮 token 与回调路由；extension 片段的声明与校验；manifest 的 `role`、`adapter_id`、`adapter_dependencies`、`plugin_dependencies` 解析与开发入口的依赖检查
前置：T-03.2、T-03.7。
依据：PLAN §3.1 的 T-03.8 行、§4 的 T-03 工作包；验收：V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项。

- 按钮 token 为 32 字节随机值的无 padding base64url，持久化绑定 plugin、route、generation、Target 和 24 小时期限。
- 先尽力调用 answerCallbackQuery，再逐次鉴权，Actor 取本次点击的用户。拒绝时不回落到 Raw/Regex，不发群消息。代次切换后 token 永久失效。
- extension 片段按目标实例所属 Adapter 在 REGISTER 中声明的 name/version/Schema 校验；目标不属于该 Adapter 或版本未声明时返回 UNSUPPORTED，不降级。
- 必需的 Adapter 依赖缺失时拒绝激活插件并列出缺失项；Adapter 停用后，发往它的 Send 返回 UNAVAILABLE。
- `plugin_dependencies`：plugin_id、可选 `min_version`、是否必需；必需依赖未安装、未启用或版本不满足时拒绝启用，可选依赖缺失只记诊断，必需依赖成环在解析时拒绝（CONTRACTS §8）。
- 验收：V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项。

### goal-51 · T-03.9 消息编辑

状态：待编写。

目标：消息编辑：`core.edit`、`adapter.edit`、Telegram editMessageText、`message.edit` 能力、出站 Hook 的 edit 输入
前置：T-03.7。
依据：PLAN §3.1 的 T-03.9 行、§4 的 T-03 工作包；验收：V-53。

- Target 准入与 Send 相同；Edit Chain 只能包含一个 text 片段和至多一个 buttons 片段，否则 INVALID_ARGUMENT。
- 每次 Edit 执行一组 BeforeSend → Adapter.Edit → AfterSend，Hook 输入标明 edit 与被编辑的 platform_message_id。
- 未声明 `message.edit` 的实例返回 UNSUPPORTED；平台拒绝（内容未变化、消息不可编辑）为 REJECTED，不改为发送新消息。
- SendResult 只有一项，status 只取 ACCEPTED/REJECTED/UNKNOWN（MESSAGE-MODEL §4）。
- 验收：V-53。

### goal-16 · T-05.2 Rule、HandlerOutcome 与传播

状态：待编写。

目标：SDK Rule（含 `ToMe`）、HandlerOutcome、自动/手动 Send、StopPropagation 锁存与错误分支
前置：T-05.1。

依据：PLAN §3.1 的 T-05.2 行、§4 的 T-05 步骤 3、4。

关键边界：Rule 返回 false 时跳过 Handler；Stop 不被后续错误清除；自动回复每个事件最多发送一次。

验收：V-14、V-11 的 Rule 分项。

### goal-24 · T-07.2 schedule 注册与运行槽

状态：待编写。

目标：schedule 注册校验，运行槽与调度调用：独立 run_id、15 分钟期限、同 job 跳过重叠、迟到响应与断链
前置：T-05.1、T-07.1。

依据：PLAN §3.1 的 T-07.2 行、§4 的 T-07 步骤 1（schedule）、3。

关键边界：同一 job 不重叠；离线错过的 tick 不补跑；运行槽只在 callback 真正返回或进程确认退出后释放，迟到响应不得提前释放。

验收：V-20、V-46、V-16 和 V-05 的 scheduled 分项。

### goal-30 · T-08.3 SDK 配置准备顺序

状态：待编写。

目标：SDK 配置准备顺序：OnLoad→补缺键默认→typed decode/Schema
前置：T-08.1、T-08.2。

依据：PLAN §3.1 的 T-08.3 行、§4 的 T-08 步骤 3。

关键边界：严格按 OnLoad→仅补缺失键的默认值→typed decode/Schema 顺序准备配置；显式 false、0、空串、空数组、空对象与 null 均保留。

验收：V-37、V-21 的配置分项。

### goal-31 · T-08.4 不可变 Secret Store

状态：待编写。

目标：不可变 Secret Store：引用、解引用、重新抽取、脱敏读取、受保护清理
前置：T-08.1、T-08.2。

依据：PLAN §3.1 的 T-08.4 行、§4 的 T-08 步骤 4。

关键边界：Secret 使用不可变引用；读取、日志、错误响应与配置快照不得泄漏明文；旧引用在候选失败时保持有效。

验收：V-25。

### goal-34 · T-08.7 web/admin 骨架、配置表单组件与界面文案

状态：待编写。

目标：`web/admin` 前端骨架，配置表单与 i18 组件及 fixture，界面文案的 en-US/zh-CN 目录与语言切换
前置：T-08.1。

依据：PLAN §3.1 的 T-08.7 行、§4 的 T-08 步骤 9（前端）。

关键边界：前端只使用仓库已批准的 Vite、TypeScript、React 19 与 vitest；界面文案提供 en-US、zh-CN 两份目录。

验收：V-25 的表单分项、V-54 的界面组件分项。

### goal-53 · T-08.8 用户可见文案目录

状态：待编写。

目标：用户可见文案目录：插件 `i18/` 的 `messages`、SDK 的 `T`、Core 的 en-US/zh-CN 文案目录（usage 等）、check 的键与占位符检查
前置：T-04.1、T-04.3、T-08.1。

依据：PLAN §3.1 的 T-08.8 行、§4 的 T-08 步骤 8。

关键边界：检查插件 i18/*.json 的 messages：en-US 包含全部键、其他语言无多余键、占位符集合一致；SDK 的 T 按 Actor locale→Core 默认 locale→en-US 回退；Core 用户可见文案提供 en-US 与 zh-CN。

验收：V-54 的 SDK 与 Core 分项。

### goal-36 · T-09.1 首次管理认证

状态：待编写。

目标：首次管理认证：admin-token init/reset、login、会话、CSRF、CLI Bearer
前置：T-07.4。

依据：PLAN §3.1 的 T-09.1 行、§4 的 T-09 步骤 1。

关键边界：管理 token 只保存 SHA-256 摘要；登录与状态变更使用 CSRF 校验，错误响应不得暴露凭证。

验收：V-47 的认证分项。

### goal-69 · T-13.3 热部署

状态：待编写。

目标：热部署：`Reload` 选项与 `ALONABOT_RELOAD`、源码轮询、重新编译、开发组件改为子进程并重启
前置：T-13.1。

依据：PLAN §3.1 的 T-13.3 行、§4 的 T-13 步骤 4。

关键边界：热部署由 Reload 选项或 ALONABOT_RELOAD=1 启用，500ms 轮询；编译失败时旧组件继续运行，重启开发组件，不重启来自其他 module 的组件。

验收：V-65。

## 第 8 波

### goal-9 · T-03.10 Telegram 接入开发入口与故障矩阵

状态：待编写。

目标：开发入口（`alonabot run` 同时启动 Adapter 与插件，实例配置取自 `adapters/<adapter_id>/instances/`），echo/media 示例与发送接收故障矩阵，汇总证据
前置：T-03.6、T-03.8、T-03.9。
依据：PLAN §3.1 的 T-03.10 行、§4 的 T-03 工作包；验收：V-07、V-08、V-09、V-35、V-44、V-45、V-48、V-51～V-54 的 T-03 分项；真实 Telegram 需要测试账号。

- 完成 PLAN T-03 步骤 8 的全部场景，包括实例上下线事件与消息编辑；用请求记录证明没有自动补发。
- 真实 Telegram 没有测试账号时记为未运行。

### goal-17 · T-05.3 三阶段 Hook 与 send_forbidden

状态：待编写。

目标：PreEvent/BeforeSend/AfterSend 三阶段 Hook，send_forbidden 的保存与继承
前置：T-05.2。

依据：PLAN §3.1 的 T-05.3 行、§4 的 T-05 步骤 5、6。

关键边界：Hook 不得修改可信 Actor、Target 或已验证路由绑定；BeforeSend 修改消息后必须重新预检；Hook 执行期间直接或间接 Send 均拒绝。

验收：V-15 的直接路径分项。

### goal-52 · T-05.4 延续调用与多轮等待

状态：待编写。

目标：延续调用与多轮等待：`core.continuation.start`、`core.wait`、等待键与消费、超时与 Session 失效；SDK 的 `Continue` 与 `Wait`
前置：T-05.2。

依据：PLAN §3.1 的 T-05.4 行、§4 的 T-05 步骤 7。

验收：V-50。

### goal-27 · T-07.5 调度与 HTTP 启停集成及验收

状态：待编写。

目标：统一启停与清理，时间、进程和 HTTP 边界验收，汇总证据
前置：T-07.2、T-07.4。

依据：PLAN §3.1 的 T-07.5 行、§4 的 T-07 步骤 7、8。

关键边界：端到端验证调度、HTTP、Session 变更与停止预算；旧 Session 的内部 HTTP 入口不得继续代理。

验收：V-20、V-21、V-22、V-46、V-47 的 T-07 分项。

### goal-32 · T-08.5 GORM 连接与迁移

状态：待编写。

目标：`gormdb.Open` 连接生命周期、迁移互斥与 versioned 迁移、PG 门槛与维护许可
前置：T-08.3。

依据：PLAN §3.1 的 T-08.5 行、§4 的 T-08 步骤 5、6。

关键边界：每个迁移的 Up 与版本账本写入同一事务；GORM 的自动迁移与自定义迁移互斥，PostgreSQL 迁移需显式维护许可。

验收：V-26、V-39；PG 需要测试实例。

### goal-39 · T-09.4 管理服务、安装记录与 serve

状态：待编写。

目标：唯一管理服务与安装记录：逐包串行、来源/pin/卸载保留、各管理动作的区分；Adapter 包与实例的管理（凭证经 Secret 引用）；启用、停用、卸载与更新时的 Adapter 依赖和插件依赖检查；`alonabot serve` 按安装记录启动全部已启用的包
前置：T-03.8、T-08.4、T-09.1、T-09.2、T-13.1。

依据：PLAN §3.1 的 T-09.4 行、§4 的 T-09 步骤 3。

关键边界：所有管理动作经唯一管理服务并按插件串行；启停、卸载、更新前检查依赖并列出受影响插件；serve 按安装记录启动已启用包，并复用 T-13 的 Core 装配。

验收：V-40、V-48、V-52、V-01 的安装分项。

### goal-71 · T-14.1 help 插件与命令列表

状态：待编写。

目标：help 插件与 `core.command.list`
前置：T-04.4、T-05.2、T-08.8。

依据：PLAN §3.1 的 T-14.1 行、§4 的 T-14 步骤 1。

关键边界：core.command.list 只返回调用者可用的命令，并按 Actor locale 翻译说明。

验收：V-67 的 help 分项。

## 第 9 波

### goal-14 · T-04.5 示例升级与 T-04 验收

状态：待编写。

目标：用同一 SDK 升级示例，完成构建与策略验收，汇总证据
前置：T-03.10、T-04.2、T-04.3、T-04.4。

依据：PLAN §3.1 的 T-04.5 行、§4 的 T-04 步骤 6、7。

关键边界：示例与验收需覆盖生成注册表、命令选择、文本匹配和持久化策略的交接；只验证实际行为，不新增无独立价值的字符串断言。

验收：V-10、V-11、V-12、V-43、V-49 的 T-04 分项。

### goal-18 · T-05.5 多进程事件流水线验收

状态：待编写。

目标：多进程事件流水线验收（含真实按钮回调与多轮等待），汇总证据
前置：T-03.8、T-05.3、T-05.4。

依据：PLAN §3.1 的 T-05.5 行、§4 的 T-05 步骤 8。

关键边界：必须以真实的多个插件进程完成流水线验收，覆盖真实按钮回调与多轮等待；进程内函数互调不能替代跨进程证据。

验收：V-11、V-13～V-16、V-43、V-49、V-50 的 T-05 分项。

### goal-20 · T-06.2 父调用链、取消与背压

状态：待编写。

目标：父调用链上下文传播、循环与深度检查、跨段取消与背压
前置：T-03.6、T-05.3、T-06.1。

依据：PLAN §3.1 的 T-06.2 行、§4 的 T-06 步骤 3、4。

关键边界：同一调用链重复进入插件返回 CALL_CYCLE，深度超过 16 返回 FAILED_PRECONDITION；Cancel 不得越过对应 Request，父调用结束后取消子调用且不重放。

验收：V-18、V-42 的嵌套分项。

### goal-33 · T-08.6 SQLite 候选与资源交接

状态：待编写。

目标：SQLite 候选复制与资源交接，写出给 T-09 的调用约定
前置：T-08.4、T-08.5。

依据：PLAN §3.1 的 T-08.6 行、§4 的 T-08 步骤 7。

关键边界：候选 SQLite 使用独立准备视图；候选失败不污染旧代次的数据、配置或 Secret，新旧进程不得同时成为同一数据库的写入者。

验收：V-24、V-36 的候选分项。

### goal-59 · T-12.2 Agent 对话与流式显示

状态：待编写。

目标：对话：`plugins/agent/` 骨架与配置、人格、历史、延续调用内的一轮对话、流式编辑显示、`/persona` 与 `/reset`
前置：T-03.9、T-04.3、T-05.4、T-08.3、T-08.4、T-11.3、T-12.1。

依据：PLAN §3.1 的 T-12.2 行、§4 的 T-12 步骤 2、3。

关键边界：流式回复先发送一条消息，再按 1 秒命名常量间隔编辑，结束时最后一次编辑写入完整内容；超长文本按段落边界拆成多条，模型输出按纯文本发送。

验收：V-57。

### goal-68 · T-13.2 开发组件的配置与数据库

状态：待编写。

目标：开发组件的配置与数据库：`config.json` 直读与准备、`dev.sqlite`、拒绝 update/rollback
前置：T-08.3、T-08.5、T-13.1。

依据：PLAN §3.1 的 T-13.2 行、§4 的 T-13 步骤 3。

关键边界：开发组件使用固定代次 dev、config.json 与 dev.sqlite；拒绝 update/rollback，配置和数据库资源按开发组件生命周期管理。

验收：V-64。

### goal-72 · T-14.2 welcome 与 rss 插件

状态：待编写。

目标：welcome 与 rss 插件
前置：T-05.2、T-07.2、T-08.3、T-08.5、T-08.8。

依据：PLAN §3.1 的 T-14.2 行、§4 的 T-14 步骤 2。

关键边界：RSS 只允许 http/https；每次请求与重定向都拒绝 loopback、私网、链路本地、组播和未指定地址，响应大小与超时使用命名常量。

验收：V-67。

## 第 10 波

### goal-22 · T-06.4 Session 清理与 Service 端到端

状态：待编写。

目标：Session 变更时的清理与计数，A/B/C 端到端验收（含 Hook→Service→Send），汇总证据
前置：T-06.2、T-06.3。

依据：PLAN §3.1 的 T-06.4 行、§4 的 T-06 步骤 6、7。

关键边界：完成 A/B/C 端到端链路并验证 Hook→Service→Send 的间接发送仍受 send_forbidden 约束；Session 失效时清理路由、订阅队列、Broker 映射与预算。

验收：V-15、V-17、V-18、V-19、V-42。

### goal-35 · T-08.9 configdb 示例与 T-08 验收

状态：待编写。

目标：configdb 示例与分层验证，汇总证据
前置：T-08.6、T-08.7、T-08.8。

依据：PLAN §3.1 的 T-08.9 行、§4 的 T-08 步骤 9。

关键边界：示例必须展示 JSON Pending、配置准备与数据库资源交接的正式路径；未完成的 PG 真实实例验收应明确标为未运行。

验收：V-23～V-26、V-38、V-39、V-54 的 T-08 分项。

### goal-38 · T-09.3 开发 CLI 与库模式脚手架

状态：待编写。

目标：开发 CLI：`plugin init/check/build/test/pack`；init 生成插件包与库模式调试用的 `main.go`，build 生成调用 `sdk.Run` 的正式入口
前置：T-04.5、T-08.1、T-08.8、T-09.2、T-13.1。

依据：PLAN §3.1 的 T-09.3 行、§4 的 T-09 步骤 2（CLI）。

关键边界：plugin init 生成可运行的插件包和库模式调试 main.go；plugin build 生成调用 sdk.Run 的正式入口；CLI 各命令复用相同的 manifest 与构建校验。

验收：V-01、V-41、V-54 的打包分项。

### goal-55 · T-11.2 工具调用与用户确认

状态：待编写。

目标：工具调用与用户确认：`core.tool.invoke`、`core.tool.confirm`、Core 内部的确认记录与按钮、同一 Actor 与 5 分钟限时、确认文案
前置：T-03.8、T-06.2、T-08.8、T-11.1。

依据：PLAN §3.1 的 T-11.2 行、§4 的 T-11 步骤 3、4。

关键边界：core.tool.invoke 与 core.tool.confirm 必须带 parent_call_id；没有可信 Actor 的调用遇 confirm 直接 PERMISSION_DENIED。确认消息由 Core 经 Adapter 直接发送，不执行插件出站 Hook；只接受同一 Actor 在 5 分钟内的第一次点击，不支持按钮时 UNSUPPORTED。

验收：V-55。

### goal-70 · T-13.4 库模式示例与验收

状态：待编写。

目标：库模式示例与开发文档，汇总证据
前置：T-03.10、T-13.2、T-13.3。

依据：PLAN §3.1 的 T-13.4 行、§4 的 T-13 步骤 5。

关键边界：示例与文档验证库模式和热部署的实际生命周期；真实平台或环境缺失的验收项标未运行。

验收：V-64、V-65。

## 第 11 波

### goal-40 · T-09.5 候选 Journal 与准备编排

状态：待编写。

目标：正式候选 Journal 与准备编排
前置：T-09.4、T-08.6、T-06.4。

依据：PLAN §3.1 的 T-09.5 行、§4 的 T-09 步骤 4。

关键边界：Journal 记录准备阶段与候选资源，任何准备失败都保留旧 Active；候选 DB 与 Secret 只在对应代次中可见。

验收：V-29。

### goal-57 · T-11.4 工具与会话的验收

状态：待编写。

目标：工具与会话的多进程验收，汇总证据
前置：T-11.2、T-11.3。

依据：PLAN §3.1 的 T-11.4 行、§4 的 T-11 步骤 6。

关键边界：验收使用多个真实插件进程，覆盖工具策略、确认与会话隔离；进程内模拟调用不能替代 Core Broker 路径。

验收：V-55、V-56 的 T-11 分项。

### goal-60 · T-12.3 工具循环与 MCP

状态：待编写。

目标：工具循环与 MCP：Core 工具、MCP 客户端（stdio、streamable HTTP）、确认、单轮步数上限
前置：T-11.2、T-12.2。

依据：PLAN §3.1 的 T-12.3 行、§4 的 T-12 步骤 4。

关键边界：MCP 只在本切片获准加入 github.com/modelcontextprotocol/go-sdk；所有 MCP 操作先经 core.tool.confirm，单轮工具步数上限为命名常量（初值 20），拒绝/超时/工具错误作为结果交回模型。

验收：V-58。

### goal-61 · T-12.4 记忆插件

状态：待编写。

目标：记忆：`sdk/memory` 服务契约、`plugins/memory/` 官方记忆插件、Agent 的记忆接入
前置：T-08.5、T-11.2、T-12.2。

依据：PLAN §3.1 的 T-12.4 行、§4 的 T-12 步骤 5。

关键边界：Agent 与记忆插件只经 sdk/... 和 contracts 交互；sdk/memory 提供三个固定 Service 的类型与 Schema。记忆按 Unicode 码点限制 2,200/1,375，超限、重复或含 Unicode Cf 的内容报错且保留旧数据；记忆插件不可用时本轮报错。

验收：V-59。

## 第 12 波

### goal-41 · T-09.6 唯一提交点与激活门禁

状态：待编写。

目标：唯一提交点（ActiveGeneration+COMMITTED）与激活门禁
前置：T-07.5、T-09.5。

依据：PLAN §3.1 的 T-09.6 行、§4 的 T-09 步骤 5。

关键边界：唯一提交点是在同一事务中写入 ActiveGeneration 与 COMMITTED；提交前失败恢复旧版，提交后失败恢复新版并进入 Degraded，不能提前开放路由。

验收：V-30、V-21 的提交分项。

### goal-62 · T-12.5 技能

状态：待编写。

目标：技能：技能目录、`SKILL.md` 解析、`load_skill`
前置：T-12.3。

依据：PLAN §3.1 的 T-12.5 行、§4 的 T-12 步骤 6。

关键边界：技能目录读取各子目录的 SKILL.md；name、description frontmatter 必填，无效文件记诊断并跳过；load_skill 对不存在的技能返回明确错误。

验收：V-60。

### goal-63 · T-12.6 定时 Agent 任务

状态：待编写。

目标：定时 Agent 任务：任务工具、任务表、分钟级固定 schedule、后台运行
前置：T-07.2、T-08.5、T-12.3。

依据：PLAN §3.1 的 T-12.6 行、§4 的 T-12 步骤 7。

关键边界：后台定时任务无 Actor、不注入记忆；confirm 策略的工具一律拒绝，离线错过的触发不补跑，发送仍需 allowed_adapter_instances。

验收：V-61。

## 第 13 波

### goal-42 · T-09.7 崩溃恢复、回退与回收

状态：待编写。

目标：崩溃恢复、显式回退与引用回收
前置：T-09.6。

依据：PLAN §3.1 的 T-09.7 行、§4 的 T-09 步骤 6。

关键边界：旧进程退出状态未确认时不得启动新的数据库写入者；恢复只信 Core 状态库的 Active 指针，并按引用关系回收旧资源。

验收：V-30、V-31、V-36。

### goal-43 · T-09.8 Admin 管理页

状态：待编写。

目标：Admin 页面：真实登录、状态、插件与 Adapter 的安装启停、Adapter 实例管理、更新/恢复
前置：T-09.6、T-08.7。

依据：PLAN §3.1 的 T-09.8 行、§4 的 T-09 步骤 7（管理页）。

关键边界：管理页面展示真实安装、启停和恢复状态；UI 状态来自管理 API，不以本地猜测代替 Core 状态。

验收：V-40、V-47 的页面分项。

### goal-64 · T-12.7 官方插件打包与 T-12 验收

状态：待编写。

目标：两个官方插件的打包与多进程验收，汇总证据
前置：T-12.4、T-12.5、T-12.6。

依据：PLAN §3.1 的 T-12.7 行、§4 的 T-12 步骤 8。

关键边界：官方 Agent 与记忆插件按 T-09 打包流程交付；使用本地 Provider/MCP 模拟服务与回环 Adapter 验收，真实 Provider/Telegram 无凭证时标未运行。

验收：V-57～V-61。

## 第 14 波

### goal-44 · T-09.9 Admin 配置、权限、工具策略与插件页

状态：待编写。

目标：Admin 页面：配置表单与 JSON 编辑、命令权限、发信许可、Core 设置（Core Admin 列表、默认 locale）、工具调用策略、Agent 会话的列出与清除、插件页面代理、界面语言切换
前置：T-09.8、T-11.1、T-11.3。

依据：PLAN §3.1 的 T-09.9 行、§4 的 T-09 步骤 7（配置与插件页）。

关键边界：Admin 页面包含工具调用策略、Agent 会话列出/清除、发信许可、Core Admin 列表、默认 locale 与界面语言切换；影响依赖方的操作须先展示影响并要求确认。

验收：V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56 的页面分项。

### goal-66 · T-09.10 日志管理与日志页

状态：待编写。

目标：日志管理：按来源级别的持久化与运行时修改（`sdk.log.set_level`）、最近记录缓冲、按日期查询日志文件、管理 API 与 SSE、`alonabot logs` 与 `log-level`、WebUI 日志页
前置：T-02.1、T-03.1、T-09.8。

依据：PLAN §3.1 的 T-09.10 行、§4 的 T-09 步骤 7（日志）。

关键边界：按来源级别持久化并通过 sdk.log.set_level 立即生效；每个来源保留最近 1000 条记录，提供受认证保护的查询与 SSE，历史日志按日期读取；CLI 与 WebUI 共用同一 API，页面筛选只在浏览器进行。

验收：V-62、V-63、V-68。

### goal-73 · T-14.3 发行归档

状态：待编写。

目标：发行归档：构建与打包全部官方包、签名、内置官方发布者公钥
前置：T-03.10、T-09.3、T-12.7、T-14.1、T-14.2。

依据：PLAN §3.1 的 T-14.3 行、§4 的 T-14 步骤 3。

关键边界：发行包的每个安装资产都按正式来源和签名流程验证；发布流程不把私钥、凭证或本机绝对路径打入归档。

验收：V-66 的归档分项。

## 第 15 波

### goal-45 · T-09.11 正式流程故障注入与 T-09 验收

状态：待编写。

目标：正式流程的故障注入，回填 T-08 待联验项，汇总证据
前置：T-09.3、T-09.7、T-09.9、T-09.10。

依据：PLAN §3.1 的 T-09.11 行、§4 的 T-09 步骤 8。

关键边界：用正式管理流程做故障注入并回填 T-08 待联验项；每个 V 项记录实际命令、环境、进程与日志，不以模拟服务替代真实平台证据。

验收：V-23～V-25、V-27～V-31、V-39、V-40、V-52、V-54～V-56、V-62、V-63。

### goal-74 · T-14.4 首次运行与预装

状态：待编写。

目标：首次运行：预装与默认启用、首次管理 token、内置包的可用更新
前置：T-09.6、T-14.3。

依据：PLAN §3.1 的 T-14.4 行、§4 的 T-14 步骤 4。

关键边界：预装包照常验签；只在数据目录首次初始化时安装，默认启用 Agent、记忆与 help；签名私钥不进入仓库与日志。

验收：V-66。

### goal-77 · T-14.5 容器镜像

状态：待编写。

目标：容器镜像：`Dockerfile` 多阶段构建、linux/amd64 与 linux/arm64、非 root、`/data` 卷、`ALONABOT_CONTAINER` 的监听规则、`deploy/compose.yaml`
前置：T-07.3、T-14.3。

依据：PLAN §3.1 的 T-14.5 行、§4 的 T-14 步骤 5。

关键边界：镜像只含 alonabot、packages/ 与 CA 证书，以 UID 10001 运行，/data 为卷且 Docker init 开启；只有 ALONABOT_CONTAINER=1 时允许 0.0.0.0 配合本机 http Base URL，并启动时 WARN。

验收：V-70 的镜像分项。

## 第 16 波

### goal-75 · T-14.6 发行包与镜像端到端验收

状态：待编写。

目标：从发行归档与容器镜像到 Agent 对话的端到端验收，汇总证据
前置：T-09.11、T-14.4、T-14.5。

依据：PLAN §3.1 的 T-14.6 行、§4 的 T-14 步骤 6。

关键边界：发行归档与镜像端到端验证安装、首次启动、预装插件与回退；真实目标平台缺失时逐项标未运行。

验收：V-66、V-67、V-70。

## 第 17 波

### goal-46 · T-10.1 证据汇总与跨组件联验

状态：待编写。

目标：V-01～V-70 证据汇总与跨组件联验
前置：T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.6。

依据：PLAN §3.1 的 T-10.1 行、§4 的 T-10 步骤 1～4。

关键边界：汇总 V-01～V-70 的逐项证据并完成跨组件闭合；真实 Telegram 与 Provider 凭证缺失时只标记对应项未运行，不将模拟结果记为真实平台通过。

验收：全部；真实 Telegram 与 LLM Provider 需要测试凭证。

## 第 18 波

### goal-47 · T-10.2 五平台构建与运行

状态：待编写。

目标：五个目标组合的构建与运行、100 次启停
前置：T-10.1。

依据：PLAN §3.1 的 T-10.2 行、§4 的 T-10 步骤 5。

关键边界：在五个目标组合上实际构建并运行，执行 100 次启停；缺少目标平台时记录未运行及其平台，不用交叉编译代替运行。

验收：V-32、V-34；需要真实目标平台。

### goal-48 · T-10.3 资源比较报告

状态：待编写。

目标：资源比较报告
前置：T-10.1。

依据：PLAN §3.1 的 T-10.3 行、§4 的 T-10 步骤 6。

关键边界：资源比较使用固定机器、等价负载和可复现命令；报告方法、环境、测量数据与限制，不能从单次结果推断通用性能保证。

验收：V-33；需要固定机器与等价负载。

## 第 19 波

### goal-49 · T-10.4 发布资产与说明

状态：待编写。

目标：发布资产、签名与发布说明
前置：T-10.2、T-10.3。

依据：PLAN §3.1 的 T-10.4 行、§4 的 T-10 步骤 7、8。

关键边界：发行资产包含可验证签名与发布说明；发布复验覆盖归档和平台清单，签名私钥不得进入仓库或日志。

验收：V-28、V-41 的发布复验。
