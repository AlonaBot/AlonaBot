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
| goal-4 | T-03.3 | Adapter 进程、sdk/adapter 与上下线事件 | T-03.1 | 待编写 |
| goal-6 | T-03.5 | 媒体上传与配额 | T-03.1 | 等待前置 |
| goal-10 | T-04.1 | 根插件 API 与静态注册索引 | T-03.1 | 待编写 |
| goal-56 | T-11.3 | Agent 会话存储 | T-03.1 | 待编写 |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 待编写 |
| goal-11 | T-04.2 | 构建期生成器 | T-04.1 | 待编写 |
| goal-12 | T-04.3 | 命令选择、文本匹配与参数绑定 | T-04.1 | 待编写 |
| goal-13 | T-04.4 | 持久化命令策略 | T-03.2、T-04.1 | 待编写 |
| goal-19 | T-06.1 | Service/Bus 描述与 Broker 准入 | T-04.1 | 待编写 |
| goal-25 | T-07.3 | SDK HTTP Listener 与 Core 代理 | T-04.1 | 待编写 |
| goal-7 | T-03.6 | 入站媒体与 Grant | T-03.2、T-03.4、T-03.5 | 待编写 |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 待编写 |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 待编写 |
| goal-21 | T-06.3 | 非持久 Event Bus | T-06.1 | 待编写 |
| goal-26 | T-07.4 | 两个 Host 与插件票据 | T-07.3 | 待编写 |
| goal-28 | T-08.1 | 配置字段描述与构建检查 | T-04.2 | 待编写 |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-06.1 | 待编写 |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2 | 待编写 |
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

依据：PLAN T-02 步骤 10；CONTRACTS §1.2、§4.2、§16；WIRE-FORMAT §3.2（SessionLimits）；TEST-MATRIX V-69；DECISIONS D-41。

范围（只新增或修改这些）：
- 新增 internal/tune/（doc.go、params.go 与解析代码）
- contracts/proto/alonabot/v1/control.proto：追加 message SessionLimits 与 Bootstrap 的 SessionLimits session_limits = 11；随后 go generate ./... 重新生成 contracts/pb（本切片获授权修改该 proto，字段号按 WIRE-FORMAT，不改已有编号）
- contracts/wirespec_test.go：追加 SessionLimits 与 Bootstrap.session_limits 的字段号检查
- internal/wire/session.go：删除 DefaultDeadline、MaxInflightPerDirection、MaxBytesPerDirection、GlobalPayloadBytes 四个常量
- internal/core/host/：prepare.go、activate.go、stop.go
- internal/core/packages/preflight.go
- internal/core/supervisor/supervisor.go
- sdk/session.go、sdk/activate.go
- cmd/alonabot/main.go
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
  - --args 是根命令的全局参数（D-41）：命令行形如 alonabot [--args V]... <子命令> [子命令参数]。根 FlagSet 只定义 --args（flag.Value 收集，可重复），解析到第一个非参数位置即为子命令；子命令之后出现的 --args 按该子命令的未知参数处理。
  - 分派子命令之前调用 tune.Parse；出错时把错误写到 stderr 并以 exitUsage 退出，不执行任何子命令。
  - run 在 logging.Open 之后，有覆盖时以 Core Logger 输出一条 WARN，消息为 "tuning parameters overridden"，属性列出每项的名称、值与出处。
  - 新增子命令 alonabot tune list：调用 tune.WriteList(os.Stdout, v) 列出缺省值与生效值，以 0 退出；不需要 data_dir，也不输出 WARN。usage 文本写明全局参数 --args，并列出 run 与 tune list。
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
  - cmd/alonabot（构建真实二进制与 examples/plugins/lifecycle）：直接参数与文件各给一次 REQUEST_TIMEOUT 时以文件中的值生效，stderr 有一条列出生效值与文件行号的 "tuning parameters overridden" WARN；--args SESSION_MAX_PAYLOAD=100MB、未知键、CONTINUATION_MAX=256 各以 exitUsage 退出、不启动插件且 stderr 指明出处；子命令之后的 --args 被拒；alonabot --args REQUEST_TIMEOUT=2s tune list 以 0 退出并显示生效值 2s。
  - supervisor：用回显环境与 Bootstrap 的测试辅助插件，确认 MemorySoftLimit=256<<20 时插件收到 GOMEMLIMIT=268435456、为 0 时没有 GOMEMLIMIT；以 SESSION_MAX_INFLIGHT=2,REQUEST_TIMEOUT=2s 解析的 Values 派生 Limits 时，插件收到 max_inflight=2、request_timeout_ms=2000。
  - host/packages：NewLimits、NewStopTiming、NewRestartPolicy 与 packages.NewLimits 取值正确；SDK 在 session_limits 缺失或含 0 时以 INVALID_ARGUMENT 失败。
- 现有测试必须继续通过；引用 DefaultLimits、DefaultStopTiming、NewRestartPolicy() 与被删除 wire 常量的测试改为 tune.Defaults() 派生的值。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./internal/... ./sdk/... ./contracts/... ./cmd/...；go build ./...。

不做：alonabot serve、库模式的 alonabot.Args 与 ALONABOT_ARGS（分别由 T-09.4、T-13.1 复用 tune.Parse 接入）；尚无调用方的参数（DISPATCH_*、BUS_*、MEDIA_*、JOB_TIMEOUT 等）只登记不接线；运行中修改参数；WebUI。
```

## 第 3 波

### goal-2 · T-03.1 首个可用 SDK 切片

状态：等待前置（T-02.2）。日志接口已按 T-02.1 的实现写入，SDK 根插件已按 sdk.Base 与 Root Hook 的现行模型修订；调优参数合并后，审查者按实际的参数接口补充本语句，再改为可派发。

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
   func New(log *slog.Logger) *Router // log 为 Core 的 Logger（logging.Output.Logger）
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
- Router.Deliver 放入容量 256 的队列，满时立即返回 adapter.ErrQueueFull。容量和单 worker 是临时实现，用一行 shortcut 注释写明：有界并发派发接入后替换。
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
- 本地自测：message.Service 的各条拒绝路径（无目标、非原目标、空 Chain、未知实例）和「结果与错误同时返回」；Router 的排序与 STOP、跳过未激活的 Plugin、第 257 条返回 ErrQueueFull、Mux 未知服务；SDK 的 EventRegistry 拒绝情况、buildRegister 输出、Handler 返回后 Send 失败、STOP 与错误并存、panic 转 INTERNAL。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./internal/wire/... ./sdk/... ./examples/...；go build ./...。

不做：Telegram、显式目标许可、后台 Send、媒体、按钮、命令分发、Rule、Hook、并发派发、cmd/alonabot 装配。
```

## 第 4 波

### goal-3 · T-03.2 Target 与发信许可

状态：等待前置（T-03.1）。

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

状态：待编写。

目标：Adapter 进程与 `sdk/adapter`：Adapter 角色的注册校验、实例启停、`adapter.send`、事件交付与游标持久化、事件归属校验、账号唯一；实例上下线时生成 `bot.online`/`bot.offline`；回环测试 Adapter
前置：T-03.1。
依据：PLAN §3.1 的 T-03.3 行、§4 的 T-03 工作包；验收：V-48 的进程与实例分项、V-51。

- WIRE §3.4 的七个固定服务全部打通；Adapter 与插件的 REGISTER 按安装记录的角色互斥校验，角色不采信进程自报。
- 实例启停：`adapter.instance.start` 下发实例配置与已保存的游标，按返回的 account_id 检查同一 `adapter_id` 下的账号唯一；启停一个实例不重启进程，也不影响同包的其他实例。
- 交付：Core 先校验事件归属（adapter_id、adapter_instance_id、account_id 与本 Session 已启动的实例一致），再写入游标、接纳事件，最后应答；伪造实例的事件被拒并记诊断。event_id 由 Core 生成。
- 进程异常退出后按插件的有限重启策略恢复，只重新启动仍处于启用状态的实例，并从已保存的游标继续。
- 实例经 `adapter.instance.start` 成功启动后生成 `bot.online`；因停用、Adapter 进程退出或 `core.adapter.status` 的致命错误离开运行状态时生成带 reason 的 `bot.offline`（MESSAGE-MODEL §1.1）。两类事件不带 Actor 与 Target，Core 重启不补发。
- 回环测试 Adapter 用于本切片和 T-05 的测试，不依赖任何真实平台。
- 验收：V-48 的进程与实例分项、V-51。

### goal-6 · T-03.5 媒体上传与配额

状态：等待前置（T-03.1）。

```text
/goal 使用 $alonabot-execute 推进 goal-6（切片 T-03.5：媒体上传与配额）。goal 目录 goal-6，分支 goal-6，worktree .worktrees/goal-6，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件可以经 Core 的固定服务分块上传媒体，得到绑定本次 Session 的 MediaRef。Core 执行全部固定配额和期限，在失败、过期或 Session 结束时回收。入站媒体、Read 与 Grant 在后续切片实现。

依据：PLAN T-03 步骤 5（上传部分）；CONTRACTS §1.3；WIRE-FORMAT §5；TEST-MATRIX V-09。

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
type Limits struct { MaxFileBytes, MaxSessionBytes, MaxCoreBytes int64; MaxUploads int; MaxChunkBytes int; TTL time.Duration }
var DefaultLimits = Limits{MaxFileBytes: 100 << 20, MaxSessionBytes: 256 << 20, MaxCoreBytes: 2 << 30, MaxUploads: 4, MaxChunkBytes: 1 << 20, TTL: 15 * time.Minute}
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

状态：待编写。

目标：根插件 API 与显式注册；静态注册索引、重复拒绝、跨插件遮蔽的稳定选择
前置：T-03.1。

依据：PLAN §3.1 的 T-04.1 行、§4 的 T-04 步骤 1、3。

关键边界：根插件与显式绑定必须唯一；跨插件同名路由的可见性与选择顺序稳定，不依赖启动顺序。

验收：V-10 的注册分项、V-02 的注册原子性。

### goal-56 · T-11.3 Agent 会话存储

状态：待编写。

目标：Agent 会话存储：状态库表、`core.conversation.*`、管理服务层的列出与清除、purge
前置：T-03.1。

依据：PLAN §3.1 的 T-11.3 行、§4 的 T-11 步骤 5。

关键边界：会话存储按 (plugin_id, session_key) 隔离，追加原子、分页受帧上限约束且不截断；卸载默认保留会话，只有显式 purge 才删除。

验收：V-56。

## 第 5 波

### goal-5 · T-03.4 官方 Telegram Adapter 的接收

状态：待编写。

目标：官方 Telegram Adapter 的接收：Bot API 客户端、本地 Bot API 模拟服务、getMe、webhook 冲突、long polling、逐条交付与游标、7 天轮次重建、401/409、媒体组、Actor.locale、`to_me`、实例代理
前置：T-03.3。
依据：PLAN §3.1 的 T-03.4 行、§4 的 T-03 工作包；验收：V-45、V-54 的 locale 分项。

- `adapter.instance.start` 中依次执行 getMe（account_id 为 bot 用户 ID 的十进制字符串）、getWebhookInfo（url 非空则失败，不调用 deleteWebhook），然后启动接收循环。
- long polling 的 allowed_updates 固定为 message、edited_message、channel_post、edited_channel_post、callback_query、my_chat_member、chat_member、message_reaction，timeout=30。
- 一批 update 逐条经 `core.adapter.deliver` 交付，游标为 update_id+1；收到应答后才处理下一条，未接纳的事件只计数。不支持的 update 也推进游标。
- 距上次接纳或轮次开始满 7 天时，先经 `core.adapter.cursor` 保存不带 offset 的新轮次，再拉取。
- 401/409 经 `core.adapter.status` 报告并停止该实例；网络错误与 5xx 只对接收退避重试，不用于发送。
- 标准化规则沿用 MESSAGE-MODEL §1、§6：带 sender_chat 的消息不构造 Actor，保留 media_group_id，每条 Telegram Message 一个事件；Actor.locale 取 from.language_code，按 BCP 47 规范化大小写；`to_me` 按 MESSAGE-MODEL §6.1 计算。
- 实例配置的 `proxy`（Secret，http/https/socks5）优先于全局代理，作用于全部 Bot API 请求与文件下载；代理失败按网络错误退避，不改为直连。
- Adapter 实现放在可导入的包中（`telegram.New()`），`main` 只调用 `sdk/adapter` 的运行函数，供库模式直接引用（CONTRACTS §15.4）。
- Bot API 请求的错误先去掉 URL 中的 `bot<token>` 再记录（CONTRACTS §14.5）。
- 验收：V-45、V-54 的 locale 分项，V-62 的 Bot Token URL 分项。

### goal-11 · T-04.2 构建期生成器

状态：待编写。

目标：构建期生成器 `sdk/generate`：限定扫描范围、build tags、生成注册代码、稳定输出；`cmd/alonabot` 的构建检查入口
前置：T-04.1。

依据：PLAN §3.1 的 T-04.2 行、§4 的 T-04 步骤 2。

关键边界：生成器只扫描当前 Go module 中符合目标构建条件的包；重复 route 必须在构建期明确报错。

验收：V-10。

### goal-12 · T-04.3 命令选择、文本匹配与参数绑定

状态：待编写。

目标：命令选择与参数绑定：完整 token、@username、最长路径/别名→文本匹配（Regex、keyword、startswith、endswith、fullmatch）→Raw、消费状态、参数语法、usage/silent
前置：T-04.1。

依据：PLAN §3.1 的 T-04.3 行、§4 的 T-04 步骤 4。

关键边界：命令一旦识别，权限拒绝、禁用或参数解析失败都不得回落到文本匹配；文本模式支持 keyword、startswith、endswith、fullmatch 与 Regex，按 priority/plugin_id/route_id 稳定排序，空模式拒绝注册，ignore_case 使用 Unicode 简单大小写折叠。

验收：V-11 的命令分项、V-49。

### goal-13 · T-04.4 持久化命令策略

状态：待编写。

目标：持久化命令策略与管理服务层：继承、禁用、全局禁止、下一次派发生效
前置：T-03.2、T-04.1。

依据：PLAN §3.1 的 T-04.4 行、§4 的 T-04 步骤 5。

关键边界：命令策略持久化并在下一次派发时生效；父级禁用优先于子级启用。

验收：V-12、V-43 的服务层分项。

### goal-19 · T-06.1 Service/Bus 描述与 Broker 准入

状态：待编写。

目标：Service/Bus 描述与编码校验、SDK 入口；Broker 定位与准入
前置：T-04.1。

依据：PLAN §3.1 的 T-06.1 行、§4 的 T-06 步骤 1、2。

关键边界：编码、版本和授权错误必须在执行目标函数之前拒绝；只接受已声明的编码与 Schema 版本，不允许运行时网络 $ref。

验收：V-19。

### goal-25 · T-07.3 SDK HTTP Listener 与 Core 代理

状态：待编写。

目标：HTTP 注册校验，SDK 托管 HTTP Listener 与 Core Session 代理映射；Core 设置文件 `alonabot.yaml` 的读取、缺省值与命令行覆盖
前置：T-04.1。

依据：PLAN §3.1 的 T-07.3 行、§4 的 T-07 步骤 1（HTTP）、4、5。

关键边界：HTTP 仅在插件声明路由时启动 loopback Listener，绑定 127.0.0.1:0；准备阶段不接业务请求，旧 Session 失效后立即停止代理。

验收：V-22 的组件分项。

## 第 6 波

### goal-7 · T-03.6 入站媒体与 Grant

状态：待编写。

目标：入站媒体与 Grant：IncomingMedia token、经 `adapter.media.read` 首次取件并校验缓存、多 Session 共享、Grant 范围、租约；出站媒体经 `core.media.read` 交给 Adapter
前置：T-03.2、T-03.4、T-03.5。
依据：PLAN §3.1 的 T-03.6 行、§4 的 T-03 工作包；验收：V-44、V-09。

- IncomingMedia token 只映射可信 Adapter 给出的平台文件标识。首次 Read 时经 `adapter.media.read` 按需取件，校验后缓存，多个 Session 共享同一份缓存。
- 出站媒体：Core 在 `adapter.send` 之前为本次发送签发只读 token，Adapter 经 `core.media.read` 读取。
- offset 与 length 的边界按 CONTRACTS §1.4。
- Grant 不扩大对象、操作、Target 或有效期。消费租约保护正在使用的对象。
- Bot Token 和宿主路径不进入插件进程。
- 验收：V-44、V-09。

### goal-50 · T-03.7 官方 Telegram Adapter 的发送计划与回执

状态：待编写。

目标：官方 Telegram Adapter 的发送计划与回执：Chain 完整预检、物理操作计划、UTF-16 实体、相册/引用/按钮渲染、PartResult 汇总
前置：T-03.4。
依据：PLAN §3.1 的 T-03.7 行、§4 的 T-03 工作包；验收：V-08、V-35。

- 发送前完成整条 Chain 的预检，生成有序的物理操作计划。文本、实体按 UTF-16 偏移；连续媒体按每组 2～10 项组成相册；Reference 只作用于首条，Buttons 作用于末条，相册后的按钮用 markup 更新补上。
- 超长、超限或无法表达的组合在首个请求前返回 UNSUPPORTED 或 INVALID_ARGUMENT，不截断、不降级、不重排。
- PartResult 区分 ACCEPTED、REJECTED、UNKNOWN、NOT_ATTEMPTED，逻辑状态按 MESSAGE-MODEL §4 汇总。遇到拒绝或结果未知即停止后续操作，保留已确认的 ID。message_id=0 不填入 ID，依赖它的步骤记为 NOT_ATTEMPTED。不自动重发。
- 验收：V-08、V-35。

### goal-15 · T-05.1 可信上下文与有界分发

状态：待编写。

目标：可信事件上下文与有界同事件分发：64 条并发链、256 条/8 MiB 等待队列、稳定候选排序、链执行与读循环分离；经回环测试 Adapter 运行，按钮与 scheduled 绑定先用规范 fixture
前置：T-03.3、T-04.3、T-04.4。

依据：PLAN §3.1 的 T-05.1 行、§4 的 T-05 步骤 1、2。

关键边界：并发链与等待队列遵守契约上限：事件等待队列最多 64 条；每 Session 最多 256 条/8 MiB 的相关请求与排队载荷。

验收：V-13、V-16 的 fixture 分项。

### goal-21 · T-06.3 非持久 Event Bus

状态：待编写。

目标：非持久 Event Bus：发布准入、逐订阅者 256 条/4 MiB 队列、drop-newest 计数
前置：T-06.1。

依据：PLAN §3.1 的 T-06.3 行、§4 的 T-06 步骤 5。

关键边界：Event Bus 每个订阅者最多 256 条或 4 MiB，先达到者为准；满时 drop-newest 并计数，离线不缓存、不补发，慢订阅者不能阻塞其他订阅者。

验收：V-17。

### goal-26 · T-07.4 两个 Host 与插件票据

状态：待编写。

目标：两个 Host 的边界、插件兑换票据与插件会话
前置：T-07.3。

依据：PLAN §3.1 的 T-07.4 行、§4 的 T-07 步骤 6。

关键边界：两个 Host 的插件票据与 Session 必须隔离；停止流程共享一个停止预算，不为每个插件重新开始计时。

验收：V-47 的组件分项。

### goal-28 · T-08.1 配置字段描述与构建检查

状态：待编写。

目标：配置字段描述与构建检查（在 T-04 生成器上扩展）
前置：T-04.2。

依据：PLAN §3.1 的 T-08.1 行、§4 的 T-08 步骤 1。

关键边界：配置描述生成必须拒绝重复 Config、重复路由与无法静态确定的字段；Schema 描述和运行时字段语义保持一致。

验收：V-25 的描述分项。

### goal-54 · T-11.1 工具描述、目录与策略

状态：待编写。

目标：工具描述与目录：Service 描述的 `tool` 字段、注册校验、`core.tool.list`、调用策略的持久化与管理服务层
前置：T-04.4、T-06.1。

依据：PLAN §3.1 的 T-11.1 行、§4 的 T-11 步骤 1、2。

关键边界：只有 JSON 编码的 Service 可以声明 tool；策略为 allow/confirm/deny，未设置时使用 default_policy，策略修改从下一次调用生效。

验收：V-55 的目录与策略分项。

### goal-67 · T-13.1 库模式入口与同进程组件

状态：待编写。

目标：库模式入口：根包 `alonabot.Run` 与选项、与 `serve` 共用的 Core 装配、开发组件同进程经 loopback 连接、生成代码中嵌入的组件身份、dev 来源与固定代次
前置：T-02.1、T-03.3、T-04.2。

依据：PLAN §3.1 的 T-13.1 行、§4 的 T-13 步骤 1、2。

关键边界：根包 alonabot.Run 与 alonabot serve 共用同一 Core 装配；internal/core、sdk、adapters/、plugins/ 不得 import 根包。

验收：V-64 的生命周期分项。

## 第 7 波

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
