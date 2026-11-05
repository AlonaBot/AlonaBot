# goal 语句：核心

本文件收录「核心」章节的 goal 语句，切片表在[核心实施计划](../../spec/plan/core.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-65 | T-02.1 | 日志基线与日志文件 | — | 已完成 |
| goal-86 | T-02.3 | 状态库改用 GORM | — | 已完成（PR #63） |
| goal-23 | T-07.1 | 调度触发计算 | — | 已完成 |
| goal-29 | T-08.2 | 配置 JSON 语义与 Pending | — | 已完成 |
| goal-37 | T-09.2 | 安装包预检与安全解压 | — | 已完成 |
| goal-76 | T-02.2 | 调优参数 | T-02.1 | 已完成（PR #28） |
| goal-2 | T-03.1 | 首个可用 SDK 切片 | T-02.1、T-02.2 | 已完成（PR #29） |
| goal-3 | T-03.2 | Target 与发信许可 | T-03.1 | 已完成（PR #31） |
| goal-4 | T-03.3 | Adapter 进程、sdk/adapter 与上下线事件 | T-03.1 | 已完成（PR #32） |
| goal-6 | T-03.5 | 媒体上传与配额 | T-03.1 | 已完成（PR #38） |
| goal-10 | T-04.1 | 根插件 API 与静态注册索引 | T-03.1 | 已完成（PR #34） |
| goal-56 | T-11.3 | Agent 会话存储 | T-03.1 | 已完成（PR #35） |
| goal-99 | T-03.12 | 后台发信许可缺省为全部实例 | T-03.2 | 已完成 |
| goal-11 | T-04.2 | 构建期生成器 | T-04.1 | 已完成（PR #42） |
| goal-12 | T-04.3 | 命令选择、文本匹配与参数绑定 | T-04.1 | 已完成（PR #47） |
| goal-19 | T-06.1 | Service/Bus 描述与 Broker 准入 | T-04.1 | 已完成（PR #45） |
| goal-25 | T-07.3 | SDK HTTP Listener 与 Core 代理 | T-04.1 | 已完成（PR #46） |
| goal-7 | T-03.6 | 入站媒体与 Grant | T-03.2、T-03.4、T-03.5 | 已完成（PR #55） |
| goal-21 | T-06.3 | 非持久 Event Bus | T-06.1 | 已完成（PR #56） |
| goal-26 | T-07.4 | 两个 Host 与插件票据 | T-07.3 | 已完成（PR #53） |
| goal-28 | T-08.1 | 配置字段描述与构建检查 | T-04.2 | 已完成（PR #54） |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2 | 等待前置 |
| goal-8 | T-03.8 | 按钮、extension 片段与依赖解析 | T-03.2、T-03.7 | 等待前置 |
| goal-51 | T-03.9 | 消息编辑 | T-03.7 | 等待前置 |
| goal-87 | T-04.6 | 权限、扩展事件与命令解析的协议和 Core 侧 | T-03.3、T-03.6、T-04.3 | 已完成 |
| goal-30 | T-08.3 | SDK 配置准备顺序 | T-08.1、T-08.2 | 已完成 |
| goal-31 | T-08.4 | 不可变 Secret Store | T-02.3、T-08.1、T-08.2 | 已完成 |
| goal-90 | T-08.10 | 配置标签统一 | T-08.1 | 已完成 |
| goal-36 | T-09.1 | 首次管理认证 | T-02.3、T-07.4 | 已完成 |
| goal-13 | T-04.4 | 持久化命令策略 | T-03.2、T-04.1、T-04.6 | 已完成 |
| goal-88 | T-04.7 | SDK 链式注册 | T-04.2、T-04.6、T-06.1、T-06.3、T-08.1 | 待派发 |
| goal-97 | T-07.6 | Core 设置与首次启动 | T-07.3、T-09.1 | 待派发 |
| goal-32 | T-08.5 | 受管 SQLite 连接与迁移 | T-08.3 | 已完成 |
| goal-53 | T-08.8 | 用户可见文案目录 | T-04.1、T-04.3、T-04.6、T-08.1、T-08.10 | 待派发 |
| goal-100 | T-09.16 | manifest 拆分为作者信息与构建信息 | T-03.8、T-09.2 | 等待前置 |
| goal-91 | T-04.8 | SDK 自有类型与事件常量 | T-04.7 | 等待前置 |
| goal-96 | T-04.9 | 删除构建期标记 | T-04.7 | 等待前置 |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 待派发 |
| goal-98 | T-07.7 | 插件页面的沙箱隔离 | T-07.4、T-07.6、T-09.1 | 等待前置 |
| goal-33 | T-08.6 | SQLite 候选与资源交接 | T-08.4、T-08.5 | 待派发 |
| goal-95 | T-08.11 | 配置改由 config_schema.json 声明 | T-04.7、T-08.3、T-08.4、T-08.10 | 等待前置 |
| goal-39 | T-09.4 | 管理服务、安装记录与 serve | T-03.8、T-07.6、T-08.4、T-09.1、T-09.2、T-13.1 | 等待前置 |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-04.7、T-06.1 | 等待前置 |
| goal-69 | T-13.3 | 开发运行与热部署 | T-07.6、T-09.16、T-13.1 | 等待前置 |
| goal-9 | T-03.10 | Telegram 接入开发入口与故障矩阵 | T-03.6、T-03.8、T-03.9、T-07.6、T-13.3 | 等待前置 |
| goal-16 | T-05.2 | Rule、HandlerOutcome 与传播 | T-05.1、T-04.7、T-04.8 | 等待前置 |
| goal-24 | T-07.2 | schedule 注册与运行槽 | T-05.1、T-07.1 | 等待前置 |
| goal-35 | T-08.9 | configdb 示例与 T-08 验收 | T-08.6、T-08.7、T-08.8、T-08.11 | 等待前置 |
| goal-101 | T-08.12 | Adapter 实例配置与 Core 设置的版本迁移 | T-07.6、T-08.11 | 等待前置 |
| goal-84 | T-09.12 | WebUI 组件与组件图标 | T-07.7、T-09.1、T-09.4、T-13.1 | 等待前置 |
| goal-68 | T-13.2 | 开发组件的配置与数据库 | T-07.6、T-08.5、T-08.11、T-13.1 | 等待前置 |
| goal-14 | T-04.5 | 示例升级与 T-04 验收 | T-03.10、T-04.2、T-04.3、T-04.4、T-04.7、T-04.8、T-04.9 | 等待前置 |
| goal-17 | T-05.3 | 三阶段 Hook 与 send_forbidden | T-05.2 | 等待前置 |
| goal-52 | T-05.4 | 延续调用与多轮等待 | T-05.2 | 等待前置 |
| goal-27 | T-07.5 | 调度与 HTTP 启停集成及验收 | T-07.2、T-07.4 | 等待前置 |
| goal-66 | T-09.10 | 日志管理 | T-02.1、T-03.1、T-07.6、T-09.4、T-09.12 | 等待前置 |
| goal-92 | T-09.13 | 管理页所需的会话与上传接口 | T-09.1、T-09.4、T-09.12 | 等待前置 |
| goal-70 | T-13.4 | 开发运行示例与验收 | T-03.10、T-13.2、T-13.3 | 等待前置 |
| goal-81 | T-16.1 | 插件测试套件基础 | T-05.2、T-13.2 | 等待前置 |
| goal-18 | T-05.5 | 多进程事件流水线验收 | T-03.8、T-05.3、T-05.4 | 等待前置 |
| goal-20 | T-06.2 | 父调用链、取消与背压 | T-03.6、T-05.3、T-06.1 | 等待前置 |
| goal-38 | T-09.3 | 开发 CLI 与库模式脚手架 | T-04.5、T-08.8、T-08.11、T-09.16、T-13.3 | 等待前置 |
| goal-82 | T-16.2 | 按钮、多轮等待、schedule 与示例测试 | T-03.8、T-05.4、T-07.2、T-16.1 | 等待前置 |
| goal-22 | T-06.4 | Session 清理与 Service 端到端 | T-06.2、T-06.3 | 等待前置 |
| goal-55 | T-11.2 | 工具调用与用户确认 | T-03.8、T-06.2、T-08.8、T-11.1 | 等待前置 |
| goal-40 | T-09.5 | 候选 Journal 与准备编排 | T-09.4、T-08.6、T-06.4 | 等待前置 |
| goal-57 | T-11.4 | 工具与会话的验收 | T-11.2、T-11.3 | 等待前置 |
| goal-102 | T-16.3 | plugintest 的 LLM 替身 | T-12.8、T-16.1 | 等待前置 |
| goal-41 | T-09.6 | 唯一提交点与激活门禁 | T-07.5、T-09.5 | 等待前置 |
| goal-42 | T-09.7 | 崩溃恢复、回退与回收 | T-09.6 | 等待前置 |
| goal-93 | T-09.14 | 配置、权限、工具策略与会话的管理 API | T-07.6、T-08.8、T-08.11、T-08.12、T-09.6、T-09.12、T-11.1、T-11.3 | 等待前置 |
| goal-45 | T-09.11 | 正式流程故障注入与 T-09 验收 | T-09.3、T-09.7、T-09.9、T-09.10、T-09.12、T-09.15 | 等待前置 |

## 第 1 波

### goal-65 · T-02.1 日志基线与日志文件

状态：已完成（PR #13 已合并）。本 goal 早于 goals.md 派发，没有原始语句；下面记录合并后的实际接口，供后继切片引用。

目标：日志基线：Core 与 SDK 改用 slog，控制台与 JSON 格式、来源标签、插件 stderr 的 JSON 解析与非结构化行、panic 堆栈、Bootstrap 下发级别与进程内过滤、`--log-level`/`--log-format`、不输出 Bootstrap 与配置明文；按日期轮转的日志文件与保留上限
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

### goal-86 · T-02.3 状态库改用 GORM

状态：已完成（[PR #63](https://github.com/AlonaBot/AlonaBot/pull/63)），证据见 TEST-MATRIX §3.2。审查期间按维护者要求修订了四轮范围（见下），PR 按 FIX-1～FIX-10 修改：业务 CRUD 不写 SQL，表名改为单数，DDL 移入 `schema.sql`，插件 ID 与会话键在入口校验非空，持久化模型嵌入公开类型。派发时：前置 T-02 已完成。已核对 `internal/core/state` 的现有实现：`open` 以 `sql.Open("sqlite", path+"?"+pragmas)` 打开单连接，已有文件先经 `preflight` 只读预检，`migrations[0]` 是唯一的初始迁移，`tx` 封装事务；包外只经公开方法访问，没有调用方拿到 `*sql.DB` 或 `*sql.Tx`，只有 `state_test.go` 直接用了 `s.db.Exec`。进行中的 PR #53～#56 与 goal-15 都不修改此包。

范围修订（2026-11-02，维护者确认）：PR #63 把原 SQL 片段搬进了 `Where("... = ?")`、`Select("coalesce(max(seq), 0)")` 与 `Where("1 = 1")`，没有达到改用 GORM 的目的。维护者要求 Core 的业务 CRUD 不写 SQL，范围、排序与取最大值用 clause 结构，不引入 GORM CLI；规则写入编码规范「Core 状态库不写 SQL」，语句「读写」一条与范围（增加 conversation/service.go 的 Clear 校验）随之修订。PR #63 按审查评论的 FIX-1～FIX-3 修改。

范围修订（2026-11-02，维护者确认）：复审 PR #63 时发现仍有硬写的名字，`Where` 后面追加 `"PluginID"` 等字段名参数，upsert 列出了冲突列和更新列；维护者同时要求表名改用单数。规则写入编码规范「Core 状态库不写 SQL」，语句「目标」「Schema」「读写」三条与本地自测随之修订。PR #63 按审查评论的 FIX-4～FIX-6 修改。

范围修订（2026-11-02，维护者确认）：维护者要求 DDL 从 Go 代码中拆到 `*.sql` 文件。规则写入编码规范「Core 状态库不写 SQL」，语句「范围」「Schema」两条随之修订。PR #63 按审查评论的 FIX-7 修改。

```text
/goal 使用 $alonabot-execute 推进 goal-86（切片 T-02.3：状态库改用 GORM）。goal 目录 goal-86，分支 goal-86，worktree .worktrees/goal-86，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：internal/core/state 的读写改用 GORM，不再手写 INSERT/SELECT 与逐列扫描；表名改用单数；公开 API、列与全部现有行为不变。

依据：PLAN T-02 步骤 11；CONTRACTS §6.1；DECISIONS D-68；TEST-MATRIX V-36、V-75；docs/execution-order.md §3 中 `schema.sql` 的规则。

范围（只新增或修改这些）：
- internal/core/state/：state.go、adapter.go、conversation.go、pending.go、permissions.go；新增 schema.sql 放初始迁移；可新增 models.go 放 GORM 模型
- internal/core/state/state_test.go：只把直接使用 `s.db` 的地方改为 GORM 写法，断言不变
- internal/core/conversation/service.go：只在 Clear 中加入会话键校验
允许的新跨模块边：无。go.mod 已直接依赖 gorm.io/gorm 与 github.com/libtnb/sqlite，不加新依赖。

接口：
- 公开符号与签名全部不变（Open、Close、ReadPragmas、LockPlugin，Install、Role、Pragmas、PendingConfig、ConversationEntry、ConversationSummary 等类型，以及各读写方法），不新增公开符号。调用方（cmd/alonabot、internal/core/adapter/remote、config、conversation、policy、supervisor 与它们的测试）不改。
- Store 的连接字段改为 *gorm.DB。

行为：
- 打开：已有文件仍先经 preflight 只读预检（可以继续用 database/sql 的只读连接），通过后以 gorm.Open(sqlite.Open(path+"?"+pragmas)) 打开；DSN 与 PRAGMA 不变，SetMaxOpenConns(1)、SetMaxIdleConns(1)。ReadPragmas 的回读与核对、「损坏不重建」、先取 core.lock 再打开状态库的顺序都不变。
- GORM 的 logger 设为 Silent：错误照常返回给调用方，由调用方用 slog 记录；不向 stdout/stderr 直接输出。
- Schema：初始迁移的 DDL 从 state.go 移到 internal/core/state/schema.sql，用 //go:embed 读入，作为唯一的一条迁移执行；state.go 中不再有建表语句。按编码规范把表名直接改为单数：installs→install、generations→generation、pending_configs→pending_config、adapter_cursors→adapter_cursor、plugin_adapter_permissions→plugin_adapter_permission、conversation_entries→conversation_entry，外键引用同步修改；schema_version 与 plugin_state 不变，列、约束与 schema 版本号不变。首次发布前不写兼容层，不追加改名迁移。迁移照旧用一次 Exec 执行并更新 schema_version；不使用 AutoMigrate，不让 GORM 推断或修改表结构。每张表有对应的模型类型，用 TableName 与 column 标签对齐表名与列名；可空列用指针或 sql.Null* 表示，保持现有的 NULL 语义。
- 读写：查询、插入、更新、删除改用 GORM，按编码规范「Core 状态库不写 SQL」不写 SQL 文本、字符串条件与字段名参数：单主键的读取与删除用 clause.PrimaryColumn，RevokeAdapterInstance 用带复合主键值的目标结构体删除，部分列匹配用结构体条件；更新在事务内按主键读出、在 Go 中检查条件后 Save；全表更新用 AllowGlobalUpdate；ListConversation 的 seq 范围、排序与条数，以及末尾 seq、下一个 install_seq 的取值用 clause.Gte 与 clause.OrderByColumn（降序取一行，没有记录时按 0），列名写成包内常量；AllowedAdapterInstances 的排序与 Conversations 的汇总在 Go 中完成，汇总只选 conversation_key、seq、created_at 三列。现有 SQL 中的 ON CONFLICT 用 clause.OnConflict 表达，语义一致且不列出冲突列和更新列（adapter_cursor 与 pending_config 用 UpdateAll，plugin_adapter_permission 用 DoNothing）；包内列名常量只保留 seq 与 install_seq。gorm.ErrRecordNotFound 映射为现有的 ErrNotFound；SetEnabled 等原来按受影响行数判断不存在的地方，改为在事务内按主键读取，不存在时返回 ErrNotFound。
- 结构体条件忽略零值字段：conversation.Service.Clear 用与插件入口相同的 validKey 校验会话键，非法时返回错误，不调用状态库。
- 事务：tx 改为基于 db.Transaction，事务内只使用传入的 *gorm.DB；CommitActive 仍在同一事务内核对预期旧值并返回 ErrStaleGeneration；AppendConversation 仍在一个事务内读末尾序号再追加；事务内不等待插件 RPC。
- 排序、上限与返回顺序与现有实现一致（例如 ListConversation 按 seq 升序且至多 limit 条，Conversations 按会话键字典序，AllowedAdapterInstances 的顺序不变）。

测试（按编码规范「测试提交规则」）：
- 不新增提交的测试。现有提交的集成测试全部原样通过：internal/core/state/state_test.go，以及使用状态库的 internal/core/config、internal/core/message、internal/core/adapter/remote、internal/core/supervisor 与 cmd/alonabot 的测试。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：同一组操作分别经改写前的实现（从 main 复制到本地测试中）与新实现执行，按新旧表名对应比较各表内容一致；Silent logger 下错误仍完整返回。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./cmd/...；go build ./...；`go list -deps` 列出的 Core 可执行文件与 SDK 根包模块都在 docs/development.md 的允许清单中。

不做：新增表或列；AutoMigrate；Backup API（仍直接用 modernc）；修改公开 API；其他包的重构。
```

### goal-23 · T-07.1 调度触发计算

状态：已完成（PR 已合并）。

```text
/goal 使用 $alonabot-execute 推进 goal-23（切片 T-07.1：调度触发计算）。goal 目录 goal-23，分支 goal-23，worktree .worktrees/goal-23，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

状态：已完成（PR #23 已合并）。后继注意：`pending_config.plugin_id` 外键引用 `install`，删除安装记录前须先删除该插件的 Pending 行。

```text
/goal 使用 $alonabot-execute 推进 goal-29（切片 T-08.2：配置 JSON 语义与 Pending）。goal 目录 goal-29，分支 goal-29，worktree .worktrees/goal-29，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
func (s *Store) GenerationConfig(ctx context.Context, generationID string) (string, error) // 读 generation.config_snapshot
迁移：CREATE TABLE pending_config (plugin_id TEXT PRIMARY KEY REFERENCES install(plugin_id), base_generation TEXT NOT NULL, config_json TEXT NOT NULL, updated_at INTEGER NOT NULL);

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
/goal 使用 $alonabot-execute 推进 goal-37（切片 T-09.2：安装包预检与安全解压）。goal 目录 goal-37，分支 goal-37，worktree .worktrees/goal-37，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

## 第 2 波

### goal-76 · T-02.2 调优参数

状态：已完成（PR #28；实现、本地验收与评论修复完成）。

```text
/goal 使用 $alonabot-execute 推进 goal-76（切片 T-02.2：调优参数）。goal 目录 goal-76，分支 goal-76，worktree .worktrees/goal-76，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
/goal 使用 $alonabot-execute 推进 goal-2（切片 T-03.1：首个可用 SDK 切片）。goal 目录 goal-2，分支 goal-2，worktree .worktrees/goal-2，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
   func New(queueEvents int) *Router // queueEvents 取 tune.Values.DispatchQueueEvents；本切片测试与示例用 tune.Defaults() 派生，已有启动路径沿用 tune.Parse 的生效值
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
- 提交：internal/core/supervisor/echo_test.go（//go:build unix）。按 supervisor_test.go 构建示例的方式构建 examples/plugins/echo，用 memory Adapter + adapter.Registry + message.Service + Mux + Router 装配；Spec.Log 取自写到测试缓冲区的 logging.Output；构造 Router 前用 slog.SetDefault(out.Logger()) 设置同一输出，并用 t.Cleanup 恢复原值，测试不调用 t.Parallel。Launch 后 Attach 实例。
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
/goal 使用 $alonabot-execute 推进 goal-3（切片 T-03.2：Target 与发信许可）。goal 目录 goal-3，分支 goal-3，worktree .worktrees/goal-3，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
迁移：CREATE TABLE plugin_adapter_permission (plugin_id TEXT NOT NULL REFERENCES install(plugin_id), adapter_instance_id TEXT NOT NULL, PRIMARY KEY (plugin_id, adapter_instance_id));
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
/goal 使用 $alonabot-execute 推进 goal-4（切片 T-03.3：Adapter 进程、sdk/adapter 与上下线事件）。goal 目录 goal-4，分支 goal-4，worktree .worktrees/goal-4，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
   迁移：ALTER TABLE install ADD COLUMN role TEXT NOT NULL DEFAULT 'plugin'; ALTER TABLE install ADD COLUMN adapter_id TEXT NOT NULL DEFAULT ''; CREATE TABLE adapter_cursor (instance_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES install(plugin_id), cursor BLOB NOT NULL, updated_at INTEGER NOT NULL);
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
  - Core 先从 adapter_cursor 读出游标放进请求；SDK 调用 Adapter.Start，失败时返回其错误（不含配置原文）。
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

状态：已完成（[PR #38](https://github.com/AlonaBot/AlonaBot/pull/38)，合并提交 `51a7078`）。证据见 TEST-MATRIX §3.2。审查中把分块上限移到 `internal/wire`：接口段的 `media.MaxChunkBytes` 已改为 `wire.MaxMediaChunk`（定义在 internal/wire/frame.go，`media` 不再导出该常量）。条目保留供后继切片查阅接口。

```text
/goal 使用 $alonabot-execute 推进 goal-6（切片 T-03.5：媒体上传与配额）。goal 目录 goal-6，分支 goal-6，worktree .worktrees/goal-6，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
// 分块上限为 wire.MaxMediaChunk = 1 << 20（internal/wire/frame.go，CONTRACTS §1.3），不是调优参数
type Store struct{ ... }
func Open(dir string, lim Limits, now func() time.Time) (*Store, error) // dir 是 <data_dir>/media；打开时清空上次遗留的临时文件，清理失败返回 error
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
  - 单块超过 wire.MaxMediaChunk，或写入后超过声明的 size → INVALID_ARGUMENT，并终止。
  - 磁盘写入失败 → INTERNAL，并终止。
  - 成功时返回 next_offset。
- Commit：
  - 实际长度必须等于 size、SHA256 必须相符，否则 INVALID_ARGUMENT，并终止。
  - 成功后签发 MediaRef{token: 16 字节随机值的十六进制, size, mime, sha256, expires_at: Commit 时刻起 TTL}。引用绑定 owner Session，预留转为引用占用。
- 回收：上传终止、过期或 ReleaseSession 时，删除临时文件并扣回预留与占用。
  - 删除失败（不含文件已不存在）时用 slog 记一条 WARN，带 path 与 err 字段；预留与占用照常扣回，upload_id/token 照常失效。残留文件由下次 Open 清理，不重试删除。
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

状态：已完成（[PR #34](https://github.com/AlonaBot/AlonaBot/pull/34)，合并提交 `284961b`）。证据见 TEST-MATRIX §3.2。审查中另增两个导出函数，事件类型与命令 token 规则由 SDK 与 Core 共用：`validate.IsCommandToken(token string) bool`、`validate.IsEventType(eventType string) bool`（含 bot.online/bot.offline）。`CommandIndex.Lookup` 与 `Shadows` 返回的 Path、Shadowed 切片只读。条目保留供后继切片查阅接口。

```text
/goal 使用 $alonabot-execute 推进 goal-10（切片 T-04.1：根插件 API 与静态注册索引）。goal 目录 goal-10，分支 goal-10，worktree .worktrees/goal-10，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
// router.New(queueEvents int) *Router；(*Router).Attach(p Plugin)；(*Router).Detach(pluginID string)；FromInstance(in *host.Instance) Plugin。
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

状态：已完成（[PR #35](https://github.com/AlonaBot/AlonaBot/pull/35)，合并提交 `d0983f0`）。证据见 TEST-MATRIX §3.2。会话表按首次发布前规则并入唯一的初始迁移；Purge 供 T-09.4 的卸载 `--purge` 调用。条目保留供后继切片查阅接口。

```text
/goal 使用 $alonabot-execute 推进 goal-56（切片 T-11.3：Agent 会话存储）。goal 目录 goal-56，分支 goal-56，worktree .worktrees/goal-56，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 状态库按（plugin_id、会话键）保存 Agent 会话条目；插件经 core.conversation.append/list/clear 访问自己的会话，追加原子、序号连续，list 按帧上限分页；Core 提供给管理服务层使用的列出、清除与按插件 purge；SDK 提供调用这三个服务的函数。

依据：PLAN T-11 步骤 5；CONTRACTS §13.3、§9.3（卸载与 --purge）、§1.2（帧上限）；WIRE-FORMAT §3.5、§5A；TEST-MATRIX V-56。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：新增 conversation.proto，随后 go generate ./...（本切片获授权新增该 proto 文件）
- contracts/wirespec_test.go：追加新消息的编号检查
- internal/wire/services.go：末尾追加三个服务名常量
- internal/core/state/state.go：在唯一的初始迁移 migrations[0] 末尾加入新表；新增 internal/core/state/conversation.go
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
   迁移：CREATE TABLE conversation_entry (plugin_id TEXT NOT NULL REFERENCES install(plugin_id), conversation_key TEXT NOT NULL, seq INTEGER NOT NULL, entry_json TEXT NOT NULL, created_at INTEGER NOT NULL, PRIMARY KEY (plugin_id, conversation_key, seq));
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

### goal-99 · T-03.12 后台发信许可缺省为全部实例

状态：已完成（PR #90）。前置 T-03.2（PR #31）已合并。依据 D-83 新增。`AdapterInstanceAllowed`、`AdapterInstanceAccess`、`SetAdapterInstanceAccess` 与 `plugin_adapter_restriction` 已合入；未安装的插件 ID 返回包装了 `state.ErrNotFound` 的错误，重复实例按主键去重。验证证据见 TEST-MATRIX §3.2。

已核对现有实现（main `1c07284`）：
- `internal/core/state/schema.sql` 的 `plugin_adapter_permission(plugin_id, adapter_instance_id)` 保存逐实例许可；`state/permissions.go` 有 `AllowedAdapterInstances`、`GrantAdapterInstance`、`RevokeAdapterInstance`。
- `policy.Policy.CheckSend` 放行对原始完整 Target 的回复，其余目标在 `AllowedAdapterInstances` 不含该实例时返回 PERMISSION_DENIED（safe_message `adapter instance is not allowed for plugin`），每次准入重新读库。
- `media.Store.AddReadServices` 的最后一个参数是 `func(context.Context, string) ([]string, error)`，生产代码还没有调用它，测试中传入 `store.AllowedAdapterInstances`。
- 使用这些方法的测试：`internal/core/message/permissions_test.go`、`internal/core/supervisor/media_in_test.go`、`command_test.go`、`internal/core/adapter/remote/remote_test.go`。

```text
/goal 使用 $alonabot-execute 推进 goal-99（切片 T-03.12：后台发信许可缺省为全部实例）。goal 目录 goal-99，分支 goal-99，worktree .worktrees/goal-99，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件的后台发信许可缺省允许全部 Adapter 实例；管理员可以把某个插件收紧为指定实例的集合，也可以恢复为全部实例。准入与撤销的语义不变：每次 Send 与入站媒体读取时重新读取，修改在下一次准入时生效。

依据：CONTRACTS §8（后台发信许可）、§1.4、§17；DECISIONS D-83；TEST-MATRIX V-07。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移中增加 plugin_adapter_restriction 表；internal/core/state/permissions.go 改为下面的方法
- internal/core/policy/policy.go：CheckSend 改用新方法
- internal/core/media/read.go：AddReadServices 的许可参数改为按实例判定
- 上述方法的全部调用者与测试
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   CREATE TABLE plugin_adapter_restriction (plugin_id TEXT PRIMARY KEY REFERENCES install(plugin_id)); -- 有记录时该插件只能发往 plugin_adapter_permission 中的实例
   func (s *Store) AdapterInstanceAllowed(ctx context.Context, pluginID, instanceID string) (bool, error) // 没有收紧时为 true
   func (s *Store) AdapterInstanceAccess(ctx context.Context, pluginID string) (all bool, instances []string, err error) // all 为 true 时 instances 为 nil；否则按字典序
   func (s *Store) SetAdapterInstanceAccess(ctx context.Context, pluginID string, all bool, instances []string) error // 在一个事务中整体替换；all 为 true 时删除收紧记录与实例行；all 为 false 且 instances 为空表示不允许任何实例
   删除 AllowedAdapterInstances、GrantAdapterInstance、RevokeAdapterInstance。
2. internal/core/media
   func (s *Store) AddReadServices(m *router.Mux, readers *adapter.Registry, dispatch *router.Router, allowed func(ctx context.Context, pluginID, instanceID string) (bool, error))

行为：
- CheckSend：对原始完整 Target 的回复照旧豁免；其余目标调用 AdapterInstanceAllowed，读取失败返回 INTERNAL，不允许时返回 PERMISSION_DENIED，safe_message 不变。实例启用、account 绑定与全局禁止的检查不变。
- 入站媒体读取按同一判定，其余规则不变。
- 收紧记录与实例行和插件的其他权限数据一样，在卸载保留数据时保留。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/message/permissions_test.go 改写：缺省时后台 Send 发往任一已启用实例成功；收紧为实例 A 后发往 B 返回 PERMISSION_DENIED、发往 A 成功；恢复为全部实例后发往 B 成功；收紧为空集合时全部拒绝；原 Target 的回复在各种设置下都放行。internal/core/supervisor/media_in_test.go 改用新判定，覆盖收紧后读取其他实例的入站媒体被拒。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：SetAdapterInstanceAccess 在事务中途失败时不留下半状态；对未安装的插件 ID 设置时报错。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：管理 API 与页面（T-09.14、T-09.9）；plugintest 收紧许可的选项（T-16.1）。
```

### goal-11 · T-04.2 构建期生成器

状态：已完成（PR #42，squash commit `62d5afa`）。`sdk/generate` 的 Scan、Options、Result、Diagnostic、Command、Event 与 `FileName`、`sdk.Generated`、`sdk.UseGenerated`、`alonabot plugin generate/check` 已合入。Scan 用 `go list -overlay` 把现有生成文件排除在依赖闭包之外。验证证据见 TEST-MATRIX §3.2。

```text
/goal 使用 $alonabot-execute 推进 goal-11（切片 T-04.2：构建期生成器）。goal 目录 goal-11，分支 goal-11，worktree .worktrees/goal-11，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：sdk/generate 在构建期扫描一个插件构建入口及其本 module 依赖包（遵守 GOOS/GOARCH 与 build tags，排除外部依赖、vendor 与测试文件），找出唯一的根类型、至多一个 Config 类型和带标记的包级 Handler，生成稳定的 Go 注册代码，由编译器检查引用；问题以带位置的诊断报告。cmd/alonabot 增加 plugin generate 与 plugin check，两者调用同一生成逻辑。

依据：PLAN T-04 步骤 2；CONTRACTS §9（生成器段）、§2.2；DESIGN「根插件」一节；DECISIONS D-20、D-39、D-40、D-42；TEST-MATRIX V-10。

范围（只新增或修改这些）：
- 新增 sdk/generate/
- sdk/：新增 generated.go（生成代码调用的注册入口）；sdk.go 的 Collect 合并生成的注册项（只加调用点）
- cmd/alonabot/：新增 plugin 命令组及 generate、check 子命令（新文件）；commands.go 只加装配：execute 中 root.AddCommand 接入 plugin 命令组，并把 &code 传给它，进入业务执行后设为 exitFailure
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

状态：已完成（PR #47，合并提交 5cb1cfd；验证见 TEST-MATRIX §3.2）。前置 T-04.1 已由 PR #34 合并；goal-10 的符号已按合入代码核对：`sdk.RouteOption`（`func(*routeOptions)`）、`sdk.Alias(path ...string)`、`sdk.Priority(p int32)`、`(*CommandRegistry).Command(name, handler, opts ...RouteOption)`、`(*EventRegistry).On(routeID, eventType, handler, opts ...RouteOption)`、`sdk.Command{Path, Aliases, Priority, Handler, Mode, Params}`、`sdk.EventRoute{RouteID, EventType, Handler, Priority}`、`router.CommandEntry`、`router.Shadow`、`(*CommandIndex).Lookup(tokens []string) (CommandEntry, int, bool)`、`(*CommandIndex).Shadows() []Shadow`（返回切片只读）、`(*Router).Commands() *CommandIndex`；router.Router、router.Invocation、adapter.Registry 沿用已合入代码，`message.New(adapters *adapter.Registry, p *policy.Policy) *Service`（PR #31 起需要 policy）。

范围修订（2026-11-01，D-67）：PR #47 审查发现 `Named.AccountName()` 读的是 `account_display_name`。Telegram 的 username 与 OneBot 的 nickname 都写在这个字段里，`/cmd@昵称` 会被当成命令。现在追加 `account_username` 字段，并授权 SDK、remote 与 Telegram 的对应调用点，见语句「范围」「接口」与「测试」。正在执行的 goal-12 从 origin/main 同步后，按修订更新 `goal-12/input.md`。

```text
/goal 使用 $alonabot-execute 推进 goal-12（切片 T-04.3：命令选择、文本匹配与参数绑定）。goal 目录 goal-12，分支 goal-12，worktree .worktrees/goal-12，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 对 message.created 依次执行「普通命令（最长完整路径与别名）→ 文本匹配（Regex、keyword、startswith、endswith、fullmatch）→ Raw」的选择，已识别的命令消费本层；按契约分词并转换参数，参数错误时发一次 usage 或按声明保持沉默；SDK 把参数绑定到 Handler，并让自动回复型 Handler 的非空 Chain 由 Core 发出一次。

依据：PLAN T-04 步骤 4；CONTRACTS §2.1、§2.2 全部、§8（自动回复段）；MESSAGE-MODEL §6.1（@username）；DECISIONS D-67；WIRE-FORMAT §3.4（adapter.instance.start 的 account_username）；WIRE-FORMAT §3.5（CommandMatch 与 CommandSpec 的追加）；TEST-MATRIX V-11 的命令分项、V-49。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/route.proto：CommandMatch 追加 COMMAND_MATCH_KEYWORD = 4、STARTSWITH = 5、ENDSWITH = 6、FULLMATCH = 7；CommandSpec 追加 repeated string patterns = 10、bool ignore_case = 11。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加上述编号检查
- contracts/validate/register.go：文本匹配与参数描述的注册校验
- internal/core/adapter/：新增 naming.go（可选接口）；internal/core/adapter/remote 的实例实现它，remote.go 保存启动响应中的 account_username
- contracts/proto/alonabot/v1/adapter.proto：AdapterInstanceStartResponse 追加 optional string account_username = 4（D-67），随 route.proto 的追加一起 go generate；contracts/wirespec_test.go 的 TestAdapterFieldNumbers 追加该字段
- sdk/adapter/adapter.go：Account 增加 Username 字段；sdk/adapter_run.go 只在构造 AdapterInstanceStartResponse 处回填它
- adapters/telegram/：telegram.go 的 getMe 映射改为 Username 取 username、DisplayName 取 first_name；inbound.go 的 to_me mention 判断改用 Username；telegram_test.go 只同步这两处断言
- internal/core/router/：新增 command.go（命令派发）与 args.go（分词与类型转换）；router.go 中 message.created 改走命令层（只加调用点），New 按下方接口增加参数
- internal/core/supervisor/：echo_test.go、conversation_test.go、index_test.go 只同步 router.New 的新签名
- sdk/：registry.go（文本匹配、Raw 与参数选项）、dispatch.go（命令路由的分发与参数绑定）、session.go（buildRegister 输出新字段）；sdk.go 只给 Context 加捕获组与自动回复标记两个未导出字段，并在 Send 开头加自动回复检查；events.go 只改 commandRouteID：文本匹配与 Raw 的显式 route_id 保存在 Command 的未导出字段中，非空时原样返回，普通命令仍生成 cmd.<path>（Collect、buildRegister 与分发共用这一函数，调用方不改）
- 上述包的测试文件
不修改 internal/core/host：Regex 的 REGISTER 拒绝由 host 已调用的 validate.Register 完成，编译结果缓存在 Router 中。
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/adapter
   type Named interface { AccountName() string } // 平台账号的用户名（account_username，D-67），用于 /cmd@name；Instance 可选实现；平台没有用户名时返回空串
   sdk/adapter：type Account struct { ID string; DisplayName string; Username string } // Username 为平台上可寻址的账号标识，空值表示平台没有；DisplayName 只作面向人的名字
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
3. internal/core/router（message 已 import router，发送经注入的回调完成，避免循环依赖）
   func New(queueEvents int, adapters *adapter.Registry, send ServiceFunc) *Router
   // adapters 用于查找目标实例的 Named；send 由装配处传入 message.Service.Send。Router 以回复所属路由的插件构造 Caller{PluginID}，
   // 以 &Invocation{PluginID, Event, DefaultTarget} 作为 ParentAttach、SendRequest（不设 To）作为载荷调用 send，走与 core.send 相同的校验。

行为：
- 注册校验（Collect 与 validate.Register 两处）：Regex 编译失败、patterns 为空或含空串、Raw 带 patterns、必需参数出现在可选参数之后、RawTail 用于非 string 或非最后一个参数、Optional 指向不存在的参数名、default 不能按参数类型解析，都拒绝注册（INVALID_ARGUMENT）。文本匹配与 Raw 的 route_id 不得使用命令保留的 cmd. 前缀，与其他路由的 route_id 重复同样拒绝（Collect 处）。Regex 编译失败经 validate.Register 在 REGISTER 时拒绝；Router 在 Attach 时编译并缓存到该插件的路由快照，不在派发时编译。
- 选择（只对 message.created）：
  1) 文本取 chain 中各 text 片段按顺序直接连接的结果。以 "/" 开头时，按空白切出 token；第一个 token 形如 name@user 时，只有 user 与目标实例的 Named.AccountName() 不区分大小写相等才去掉 @user，否则整条消息不算命令（实例不实现 Named，或 AccountName() 为空时，不接受 @ 形式；AccountName 只取 account_username，不用展示名代替）。
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
- 提交：internal/core/adapter/remote/remote_test.go 的回环 Adapter 在 Account 中给出 Username 时，经真实 adapter.instance.start 后实例的 AccountName 等于它，DisplayName 不影响结果；不给出时为空。adapters/telegram/telegram_test.go：模拟 getMe 的 username 与 first_name 分别出现在 account_username 与 account_display_name 中，@username 的 mention 使 to_me 为 true，@first_name 不会。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：分词与类型转换的完整表（int64 溢出、NaN、空字符串参数、raw 为唯一参数）；注册校验的每种拒绝；SDK 参数绑定的类型不符；Unicode 大小写折叠。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。

不做：命令策略、权限与禁用（T-04.4）；Rule 与 ToMe（T-05.2）；Hook（T-05.3）；延续调用（T-05.4）；usage 的多语言文案（T-08.8）；生成器对新选项的标记语法（T-04.2 之后按需追加）。
```

### goal-19 · T-06.1 Service/Bus 描述与 Broker 准入

状态：已完成（PR #45，squash commit `dcb4ec2`）。已合入：`internal/core/broker`（New、Call）、`(*router.Mux).Forward`、`(*router.Router).Plugin`，以及 SDK 的 Services Hook、ServiceRegistry、Access 常量、WithAccess、JSONService、ProtoService、CallJSON、CallProto、Context.Caller。审查中经维护者授权，插件 ID 规则从 `manifest.ValidateID` 移到 `validate.PluginID`，同时移走的还有 `validate.MaxPluginIDLen` 与 `validate.MarketPrefix`；manifest 不保留旧入口，语句中的 manifest.ValidateID 以此为准。验证证据见 TEST-MATRIX §3.2。

```text
/goal 使用 $alonabot-execute 推进 goal-19（切片 T-06.1：Service/Bus 描述与 Broker 准入）。goal 目录 goal-19，分支 goal-19，worktree .worktrees/goal-19，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件可以声明 Service（名称、输入/输出 schema 与版本、编码、访问规则）并经 SDK 调用其他插件的 Service；Core 的 Broker 以 plugin_id/service_name 定位目标的已激活 Session，在调用目标函数前完成访问规则、编码、schema 标识与版本、JSON 语法的检查，再以新的 Core→目标请求转发。REGISTER 同时校验 Service 与 Bus 订阅描述（Bus 的运行在 T-06.3）。

依据：PLAN T-06 步骤 1、2；CONTRACTS §3 第 2、3 条及第 4 条中「每一跳单独编号」的部分、§1.2；WIRE-FORMAT §4、§5；DECISIONS D-43；TEST-MATRIX V-19。

范围（只新增或修改这些）：
- contracts/validate/register.go：Service 与 Bus 订阅描述的注册校验
- contracts/validate/envelope.go：只改 checkPayload 的 schema_version 检查——控制帧仍要求为 1，REQUEST/RESPONSE 的业务载荷只要求 ≥ 1（D-66）；contracts/wirespec_test.go 追加对应的帧检查
- internal/wire/session.go：只改 dispatch 对 validate.Envelope 错误的处理——REQUEST 的错误码为 INVALID_ARGUMENT（JSON 语法不符，Envelope 的结构检查此前已通过）时不关闭 Session：照常做 CheckPeer 与 onRequest 的 call_id 单调、阶段和父调用检查，再对该次调用回复 INVALID_ARGUMENT，不执行 Handler；其余校验错误与 RESPONSE 帧仍按协议错误关闭（CONTRACTS §3）
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
   type Access int // 常量 AccessPrivate（缺省，只允许 Core 与本插件）、AccessAllowlist、AccessPublic；带类型名前缀，避免与 HTTP 的 sdk.Public(mode) 冲突
   type ServiceOption func(*serviceOptions)
   func WithAccess(a Access, allowlist ...string) ServiceOption // AccessAllowlist 时给出允许的插件 ID
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
   func New(r *router.Router, requestTimeout time.Duration) *Broker // requestTimeout 取 tune.Values.RequestTimeout
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
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：validate 的 Service 与 Bus 订阅描述各条拒绝；Collect 的服务名重复与非法；期限取最早者；Context.Caller；SDK 解码失败不调用 h；Mux.Forward 不覆盖固定服务；wire.Session 收到 JSON 语法不符的 REQUEST 时回复 INVALID_ARGUMENT、不调用 Handler，同一 Session 的后续调用仍成功（SDK 会在发送前拒绝非法 JSON，这一项用原始帧驱动）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/core/... ./internal/wire/... ./sdk/...；go build ./...。

不做：父调用链、深度与 CALL_CYCLE、取消与背压（T-06.2）；Event Bus 的发布与投递（T-06.3）；工具描述（T-11.1）；用 jsonschema-go 校验业务字段（Core 不校验，SDK 以类型解码为准；工具的 Schema 校验在 T-11）。
```

### goal-25 · T-07.3 SDK HTTP Listener 与 Core 代理

状态：已完成（PR #46，squash commit `1c3c912`）。`internal/core/settings`（Settings、Defaults、Load、ValidateListen、FileName）、`internal/core/httpproxy`（Authorizer、Server、New、Bind、Unbind、ServeHTTP）、`wire.HeaderProxyAuth`、`wire.CapabilityHTTPListener` 与 SDK HTTP Listener 已合入。`ValidateListen` 是审查中经维护者认可追加的导出符号，供 YAML 与 `--listen` 共用。验证证据见 TEST-MATRIX §3.2。

```text
/goal 使用 $alonabot-execute 推进 goal-25（切片 T-07.3：SDK HTTP Listener 与 Core 代理）。goal 目录 goal-25，分支 goal-25，worktree .worktrees/goal-25，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件只注册 http.Handler：SDK 在存在 HTTP 路由时持有 127.0.0.1:0 的 Listener，并在 REGISTER 报告实际端点，用户 Handler 外包内部凭证校验与就绪门禁；Core 读取 <data_dir>/alonabot.yaml（不存在时以缺省值创建），在唯一的公开 Listener 上把 /plugins/{id}/... 经标准库 ReverseProxy.Rewrite 转发给当前已激活 Session 的端点，清除可伪造的身份信息并注入受信凭证；Session 结束或代次变化时撤销映射。

依据：PLAN T-07 步骤 1（HTTP）、4、5；CONTRACTS §7.2、§15.1、§14.6；WIRE-FORMAT §3.1（HTTP 能力）、§3.2（http_proxy_auth）、§3.3；DECISIONS D-25、D-65；TEST-MATRIX V-22 的组件分项。

范围（只新增或修改这些）：
- contracts/validate/register.go：HTTP 路由描述的注册校验
- internal/wire/：新增 http.go（内部认证头名与 HTTP 协议能力名常量）
- internal/tune/：新增导出的大小解析函数（供设置文件复用）
- 新增 internal/core/settings/（alonabot.yaml）
- 新增 internal/core/httpproxy/
- internal/core/host/prepare.go：插件在 HELLO 声明 HTTP 能力时在 Bootstrap 中下发 http_proxy_auth；internal/core/host/activate.go 只给 Instance 加 HTTP 端点与代理 token 两个字段，并在构造 Instance 处从 Prepared 赋值（装配处经 supervisor 返回的 Running.Instance 读取后调用 Bind，不修改 supervisor.go）
- sdk/：新增 http.go（Listener、门禁与凭证校验）；sdk.go 去掉「不支持 HTTP 路由」的拒绝，Run 中只加 Listener 的创建、向 prepare 与 activate 的传递和随 Run 返回的关闭，并把是否有 HTTP 路由传给 clientHandshake，Collect 不改；registry.go 只改 Handle 的校验分支，规则与 validate.Register 的 HTTP 路由校验一致（标准方法、规范路径与前缀路由的尾斜杠、不含 "?" 与 "#"、is_public 与 auth_mode 的对应），错误照旧经 errs 由 Collect 返回；activate.go 与 session.go 只加启动、上报与关闭 Listener 的调用点及持有 Listener 的未导出字段，session.go 另在构造 Hello 处声明 HTTP 能力
- cmd/alonabot/：run 读取设置文件、增加 --listen 并让 --log-level/--log-format 覆盖设置文件；启动 Core HTTP Listener 并接入插件的 HTTP 路由
- 新增 internal/core/supervisor/testdata/httpplugin/（集成测试用的插件程序）
- 上述包的测试文件
允许的新跨模块边：无（cmd/alonabot -> internal/core/...、-> internal/tune 已允许；yaml 使用已批准的 go.yaml.in/yaml/v3）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/wire/http.go
   const HeaderProxyAuth = "X-Alonabot-Proxy-Auth" // Core 注入、SDK 校验的内部凭证头；值为 http_proxy_auth.token 的十六进制
   const CapabilityHTTPListener = "http_listener" // HTTP 协议能力名，版本 1（WIRE-FORMAT §3.1）
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
   func New(auth Authorizer) *Server // auth 为 nil 时私有路由一律拒绝
   func (s *Server) Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor) // 绑定当前已激活 Session 的路由；替换同一插件的旧映射
   func (s *Server) Unbind(pluginID, generationID string) // 只有 generationID 与当前映射一致时才撤销
   func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)
5. sdk（HTTPRegistry 已存在）：Handle 的注册项随 REGISTER 上报；不新增公开符号。

行为：
- 注册校验（validate.Register 与 SDK Collect）：HTTP 路由的 method 为大写的标准方法之一；path 以 "/" 开头、是规范形式（path.Clean 的结果；原路径以 "/" 结尾且结果不是 "/" 时再补回一个 "/"，即前缀路由，"/" 本身是匹配全部路径的前缀路由）、不含 "?"、"#" 与 ".."；"/static" 与 "/static/" 是两条不同的路由；同插件的 (method, path) 唯一；is_public 为 true 时 auth_mode 必须是 CALLBACK_STATE 或 CALLBACK_SIGNATURE，为 false 时必须是 CORE_SESSION；有 HTTP 路由时 http_listener 必须存在且为 loopback 非零端口，没有 HTTP 路由时必须不存在。
- SDK：
  - 存在 HTTP 路由时在 HELLO 的 required_capabilities 中声明 {CapabilityHTTPListener, 1}，没有时不声明。
  - 只有存在 HTTP 路由时才在准备阶段监听 127.0.0.1:0，REGISTER 报告实际 host 与 port；Listener 在整个 Session 期间保持打开，不释放后重开。
  - 每个请求先校验 HeaderProxyAuth 等于 Bootstrap 下发 token 的十六进制（恒定时间比较），不符返回 401 且不进入用户 Handler；Session 尚未 ACTIVATED 时返回 503；校验通过后删除该头再交给用户 Handler。路由匹配按注册的 method 与 path（精确匹配 path，或注册为以 "/" 结尾的前缀）。
  - STOPPING、断链、激活失败与 Run 返回时关闭 Listener 并等待在途请求结束（受停止预算约束）。
- Core host：Core 的 supported 能力包含 {CapabilityHTTPListener, 1}；能力交集含该项时，Bootstrap.http_proxy_auth.token 为本次 Session 新生成的 32 字节随机值，否则不设置；token 不进入日志。REGISTER 带 HTTP 路由而交集不含该项时拒绝注册（INVALID_ARGUMENT）。
- Core httpproxy 与 SDK 用同一份未清理的请求路径做路由匹配，两侧对同一请求选出同一条路由，避免经公开前缀路由（如 /static/../hello）绕过私有路由的授权。
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

## 第 6 波

### goal-7 · T-03.6 入站媒体与 Grant

状态：已完成（[PR #55](https://github.com/AlonaBot/AlonaBot/pull/55)），证据见 TEST-MATRIX §3.2。审查期间按复审评论追加 `message.New` 末尾的 `requestTimeout`（见接口 3b），Telegram 取件按 Range 分块并在每轮 offset 0 刷新平台路径。派发时：T-03.2、T-03.4、T-03.5 已合并（PR #31、#39、#38）。已核对媒体存储入口 `media.Open(dir string, lim Limits, now func() time.Time) (*Store, error)`、`AddServices`、`ReleaseSession`、`Sweep`、`Lookup(token string, c router.Caller) (File, error)`、`media.File{Path, Size, MIME, SHA256, ExpiresAt}` 与 `wire.MaxMediaChunk`；回收删除失败记 WARN，safe_message 为 `media storage failed`。Telegram 可复用 `New() adapter.Adapter`、私有 `instance`、`telegramMessage`、`messageEvent`、`botClient.call(ctx context.Context, method string, request any, result any) error`；`telegramtest.NewServer(t testing.TB, token string) *Server` 与接收模拟入口已合入。`sdk/adapter.Core` 当前含 Deliver、SaveCursor、ReportFailure，ReadMedia 由本切片按下文接口追加。

范围修订（2026-11-01，维护者确认）：执行者在 GN-02 报告的阻塞中，有三项现在补齐，见语句「范围」与「接口」：修改 media/store.go，接入共享对象与租约回收；`LeaseForSend` 增加目标 Adapter Session 的 SessionKey 输入；Adapter 实例提供该 SessionKey，remote 由此改为 Bind 时传入 Caller，入站登记经 Options 注入。仍待补写的是 Router 与 Grant 一侧的装配：Router 为接收 Session 签发 token 的回调注入、grantee 的已激活 Session 查询、同事件在途关联查询、`AddReadServices` 的许可查询参数。这些都落在 internal/core/router；#45（`Router.Plugin`）与 #47（`router.New` 签名）现已合并，已核对 `Router.New(queueEvents int, adapters *adapter.Registry, send router.ServiceFunc) *router.Router` 和 `Router.Plugin(pluginID string) (router.Plugin, bool)`。审查者仍需据此补写媒体回调和关联查询的装配签名，依赖这些装配的部分（Router 改写 incoming token、core.media.grant）保持阻塞，其余可以继续执行。


范围修订（2026-11-02，维护者按 PR #55 的审查建议确认）：执行者在 PR #55 提出的四个阻塞项按以下方式补齐，见语句「范围」「接口」「行为」与「测试」：
- GN-03 装配：Router 的 Attach 接收 Caller，新增 `Caller` 与 `HasInFlight`，`New` 增加 prepare 回调；media 新增 `PrepareIncoming`，`AddReadServices` 增加 Router 与实例许可查询；message 的 `New` 增加 lease 回调。依赖方向保持 media→router、message 不 import media。「在途调用关联」定义为 grantee 正在处理同一可信事件的实际调用，调用开始时登记，返回或取消时移除。
- FIX-3：`LeaseForSend` 增加最终发送的 Target，取件前检查原 token 的 Target 范围。
- FIX-1：可读性由实例启动响应中的能力 `media.read` v1 决定，SDK 按实例是否实现 MediaSource 生成；不改 proto。
- FIX-2：不过滤 `platform_payload`。平台文件标识不是插件可用的读取凭证，Core 只认自己签发的 token（CONTRACTS §1.4，D-73）。
- 审查发现的 P1：取件失败（含首个读取者的 ctx 取消或到期）不缓存为对象状态，下一次 Read 重新取件。
- `router.New` 与 goal-13 同时改签名，按 docs/execution-order.md §3 处理。
```text
/goal 使用 $alonabot-execute 推进 goal-7（切片 T-03.6：入站媒体与 Grant）。goal 目录 goal-7，分支 goal-7，worktree .worktrees/goal-7，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：入站媒体以 IncomingMedia 交给插件：Core 只保存可信 Adapter 给出的平台文件标识，向每个接收 Session 签发独立 token；首次 Read 时经 adapter.media.read 按需取件、校验后缓存一份供多个 Session 共享。插件经 core.media.read 读取、经 core.media.grant 在不扩大范围的前提下转给其他插件。出站媒体在 adapter.send 之前由 Core 签发只读 token，Adapter 经 core.media.read 读取；发送期间以租约保护对象。

依据：PLAN T-03 步骤 5（入站与授权）；CONTRACTS §1.3、§1.4、§8（媒体段）；WIRE-FORMAT §3.4、§5；TEST-MATRIX V-44、V-09。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/adapter.proto：追加 AdapterMediaReadRequest/Response。随后 go generate ./...（本切片获授权只追加这两个消息）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceAdapterMediaRead = "adapter.media.read"、ServiceMediaRead = "core.media.read"、ServiceMediaGrant = "core.media.grant"
- internal/core/media/：新增 incoming.go、read.go、grant.go、lease.go；store.go 只扩展 entry 与 Store 的字段，以及 ReleaseSession、Sweep、release 的回收条件（接入入站共享对象、按 Session 的 token 与发送租约），上传（begin、write、commit）与 Lookup 对已提交 MediaRef 的现有行为不变
- internal/core/adapter/：新增 media.go（两个可选接口）；internal/core/adapter/remote 与 internal/core/adapter/memory 实现它们；remote 的 deliver 改写入站媒体来源；remote.go 的 Bind 增加 Caller 参数，Options 增加入站登记回调（同步 remote_test.go 的调用点）
- internal/core/router/：Attach 增加 Caller 参数，新增 Caller 与 HasInFlight，New 增加 prepare 回调并在派发前对每个接收 Session 调用；同步 internal/core/supervisor 测试中全部 router.New 与 Attach 的调用点
- internal/core/message/：New 增加 lease 回调（同步全部调用点）；发送前解析媒体来源并签发 Adapter 只读 token
- sdk/：media.go 增加读取与授权函数；sdk/adapter 的 Core 增加 ReadMedia，Instance 可选实现平台文件读取；sdk 的 RunAdapter 处理 adapter.media.read，并按实例是否实现 MediaSource 在启动响应的 capabilities 中声明 `media.read` v1
- adapters/telegram/：新增 media.go（入站媒体片段与 getFile 取件）；adapters/telegram/telegramtest/ 增加 getFile 与文件下载
- 上述包的测试文件（含扩展 internal/core/adapter/remote/remote_test.go 的回环 Adapter）
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. adapter.proto
   message AdapterMediaReadRequest { string instance_id = 1; string file_id = 2; uint64 offset = 3; uint32 length = 4; } // length 为 1～1 MiB
   message AdapterMediaReadResponse { bytes data = 1; uint64 next_offset = 2; bool eof = 3; }
2. internal/core/adapter
   type MediaReader interface { ReadMedia(ctx context.Context, fileID string, offset uint64, length uint32) (data []byte, next uint64, eof bool, err error) } // Instance 可选实现
   type MediaSession interface { MediaSessionKey() string } // Instance 可选实现：该实例所属 Adapter Session 的 SessionKey，出站只读 token 绑定到它；adapter 包不 import router，因此只传 SessionKey
   remote：func (p *Process) Bind(ctx context.Context, in *host.Instance, caller router.Caller) error // caller 为该 Session 装配 Mux.For 时使用的同一个 Caller；实例的 MediaSessionKey 返回 caller.SessionKey
   remote：Options 增加 RegisterIncoming func(instanceID, fileID string, size *uint64, mime *string) string // 装配处传入 media.Store.RegisterIncoming；remote 不 import media
   memory：MediaSessionKey 返回该实例固定的 SessionKey，测试以 router.Caller{SessionKey: 该值} 调用 media.Store.Lookup 读出只读 token 的内容
3. internal/core/media（在 goal-6 的 Store 上追加）
   func (s *Store) RegisterIncoming(instanceID, fileID string, size *uint64, mime *string) string // 返回对象 ID；同一实例同一 file_id 复用同一对象
   func (s *Store) IssueIncoming(objectID string, c router.Caller, targets []*pb.Target) *pb.IncomingMedia // 给一个接收 Session 签发 token
   func (s *Store) PrepareIncoming(ev *pb.Event, c router.Caller) (*pb.Event, error) // 交给 router.New 的 prepare：克隆事件，把其中每个入站对象 ID 换成为 c 签发的 token（内部调用 IssueIncoming），token 记录签发它的原事件
   func (s *Store) AddReadServices(m *router.Mux, readers *adapter.Registry, dispatch *router.Router, allowedInstances func(ctx context.Context, pluginID string) ([]string, error)) // 注册 core.media.read 与 core.media.grant；allowedInstances 由装配处绑定 state.Store.AllowedAdapterInstances
   func (s *Store) LeaseForSend(ctx context.Context, c router.Caller, readerSessionKey string, target *pb.Target, source *pb.MediaSource, deadline time.Time) (readToken string, release func(), err error) // 供 message.Service：解析 media_ref 或 incoming_ref，签发只读 token 并绑定到 readerSessionKey（目标实例的 MediaSessionKey）
3a. internal/core/router
   func New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(ev *pb.Event, c Caller) (*pb.Event, error)) *Router // prepare 为 nil 时原样派发
   func (r *Router) Attach(p Plugin, c Caller) // c 与装配处传给 Mux.For 的是同一个 Caller
   func (r *Router) Caller(pluginID string) (Caller, bool) // 只返回当前已激活实例绑定的 Caller；替换、Detach 与退出后不再返回旧值
   func (r *Router) HasInFlight(c Caller, ev *pb.Event) bool // c 正在处理 ev（按 event_id）的实际调用
3b. internal/core/message
   func New(adapters *adapter.Registry, p *policy.Policy, lease func(ctx context.Context, c router.Caller, readerSessionKey string, target *pb.Target, source *pb.MediaSource, deadline time.Time) (string, func(), error), requestTimeout time.Duration) *Service // 装配处传入 media.Store.LeaseForSend 与 tune.Values.RequestTimeout；message 不 import media；没有 RPC 期限的发送（如自动回复）用 requestTimeout 计算媒体期限
4. sdk
   func ReadMedia(ctx context.Context, token string, offset uint64, length uint32) (data []byte, next uint64, eof bool, err error)
   func GrantMedia(ctx context.Context, token, granteePluginID string, targets ...*pb.Target) (string, error) // 返回给对方的新 token
5. sdk/adapter
   Core 增加：ReadMedia(ctx context.Context, token string, offset uint64, length uint32) ([]byte, uint64, bool, error)
   type MediaSource interface { ReadPlatformFile(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) } // Instance 可选实现；未实现时 adapter.media.read 返回 UNSUPPORTED

行为：
- 入站：Adapter 在 image/video/file 片段的 MediaSource.incoming_ref.token 中放平台文件标识（可带 size、mime）。remote 的 deliver 在接纳前把每个这样的来源登记为对象（RegisterIncoming），token 改为对象 ID；平台文件标识此后只留在 Core。Router 在每个接收 Session 派发前调用 prepare（装配处传入 PrepareIncoming），由它为该 Session 调用 IssueIncoming，把对象 ID 换成该 Session 独立的 token（16 字节随机值的十六进制，自签发起 MEDIA_TTL 有效，readability 为 READABLE；实例在 adapter.instance.start 的 capabilities 中声明了 `media.read` v1 且 supported 时为 READABLE，否则为 UNSUPPORTED；Adapter 交付时已把某个来源的 readability 标为 UNSUPPORTED 的，Core 保留。remote 只在实例声明了该能力时实现 MediaReader）。平台文件标识不是插件可用的读取凭证：插件只能用 Core 签发的 token 读取或发送；Adapter 的 platform_payload 原样保留，其中出现的平台标识没有授权作用，Core 也不接受它作为媒体来源。
- core.media.read（插件与 Adapter 都可调用，只认调用者 Session 持有的有效 token）：
  - length 为 0 或超过 1 MiB、offset 超过实际大小 → INVALID_ARGUMENT；offset 等于实际大小 → 空 data、原 offset、eof=true。
  - 已提交的 MediaRef（goal-6）直接按 offset 读取。
  - 入站对象首次被读取时取件：预留容量（已知 size 用它，未知按 MEDIA_MAX_OBJECT），计入 Core 总预算与读取者 Session 的预算，超限 RESOURCE_EXHAUSTED；按 1 MiB 依次调用 MediaReader 写入暂存区，总长超过预留或与声明的 size 不符时失败并删除；完成后计算 SHA256。取件期限取调用者期限与 MEDIA_TTL 中较早者。并发的首次读取只取件一次，其余等待结果。取件失败返回 UNAVAILABLE（平台错误的 safe_message 原样），同一次调用内不自动重试。失败（包括首个读取者的 ctx 取消或到期）只返回给本轮正在等待的读取者，不记为对象状态：对象恢复为未取件、预留归还，下一次 Read 由新的读取者重新取件。
- core.media.grant：只能授权自己持有的 token；grantee 必须是当前已激活的插件；targets 必须是原 token 范围的子集（原范围为空时 targets 必须为空）。入站对象还要求 grantee 与该事件有在途调用关联（Router.HasInFlight：grantee 正在处理签发原 token 的那个事件的实际调用，调用开始时登记、返回或取消时移除，候选与排队阶段不算；Grant 派生的 token 继承原事件），或 grantee 的 allowed_adapter_instances 含来源实例；插件自己上传的对象由 owner 直接委托。新 token 不延长有效期（取原 token 的到期时间）。不满足时 PERMISSION_DENIED。
- 出站：message.Service 在发信准入通过后、调用 Instance.Send 前，对每个 media_ref 或 incoming_ref 来源调用 lease（即 LeaseForSend，传入最终 Target 与目标实例的 MediaSessionKey）：先检查原 token 的 Target 范围，校验调用者对该 token 的权利，入站对象必要时先取件，再签发绑定目标 Adapter Session 的只读 token 并替换片段来源；租约在发送返回或期限到达时释放。租约期间对象不被 Sweep 回收，但仍计入磁盘预算。https_url 来源原样交给 Adapter。取件失败发生在平台发送之前，不调用 Send；任一来源失败时释放本次已取得的全部租约。目标实例不实现 adapter.MediaSession 而 Chain 含 media_ref 或 incoming_ref 时，在首个副作用前返回 UNSUPPORTED。
- Session 结束（ReleaseSession）只撤销该 Session 的 token；入站对象在所有 token 过期且没有租约时才删除。
- Adapter 端：RunAdapter 收到 adapter.media.read 时调用实例的 MediaSource；Adapter 用 Core.ReadMedia 读取出站只读 token 的内容。
- Telegram：photo（取最大尺寸）、video、document、animation、audio、voice 生成对应的 image/video/file 片段，incoming_ref.token 为 file_id，size 取 file_size（有时），mime 取 mime_type（有时），caption 写入片段；实例实现 MediaSource：首次读取时 getFile 得到 file_path，经实例的 http.Client（含代理）下载 /file/bot<token>/<file_path>，按 offset 返回；超过 Bot API 的 20 MiB 下载上限（命名常量）时返回 UNSUPPORTED。下载 URL 同样脱敏。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/media_in_test.go（//go:build unix）：memory Adapter 实现 MediaReader（内存中的文件），注入带图片的 message.created；两个测试插件都收到不同的 token，各自用 sdk.ReadMedia 分块读出完整内容；MediaReader 只被调用一次取件；插件 a 用 sdk.GrantMedia 授权给与事件无关联且无实例许可的 c 时被拒，给已有许可的 b 成功；a 用 incoming_ref 回发该图片，memory Adapter 的 SendFunc 收到只读 token，并能经 media.Store 读出相同字节。
  - internal/core/adapter/remote/remote_test.go 追加：回环 Adapter 实现 MediaSource；经真实 adapter.media.read 取件并在 Read 中得到相同内容；未实现 MediaSource 的实例使 readability 为 UNSUPPORTED。
  - contracts/wirespec_test.go：新消息的编号检查。
  - adapters/telegram/media_test.go：telegramtest 推入带 photo 与 document 的消息，片段的 incoming_ref.token 为 file_id；MediaSource 经 getFile 与下载返回正确内容，错误中不含 token。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：offset/length 边界；声明 size 与实际不符；未知大小超过上限；并发首次读取只取件一次；token 过期；Grant 扩大范围被拒；租约期间 Sweep 不删除；ReleaseSession 不影响其他 Session 的 token；首个读取者取消或到期后，另一 Session 仍能读到完整内容且 MediaReader 被再次调用；未声明 `media.read` 的实例 readability 为 UNSUPPORTED；Router.Caller 在 Detach 后不再返回旧值；LeaseForSend 对超出原 token 范围的 Target 返回 PERMISSION_DENIED。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。

不做：https_url 的下载代理（契约不提供）；按钮与 extension（T-03.8）；本地 Bot API 服务的大文件路径（T-03.10）。
```

### goal-21 · T-06.3 非持久 Event Bus

状态：已完成（[PR #56](https://github.com/AlonaBot/AlonaBot/pull/56)，合并提交 `cd35ff7`）。证据见 TEST-MATRIX §3.2。审查留下两项 P2 待跟进：SDK 侧订阅投递只受 wire 接收预算约束，卡住的回调会挤占本插件的命令请求（是否另设上限待维护者决定，见该 PR 的审查评论）；`bus.New` 的零值读取 `tune.Defaults()`，生产装配（T-06.4）应传入 `tune.Values` 的 BusQueueEvents/BusQueueBytes。条目保留供后继切片查阅接口。

派发时的核对：前置 T-06.1 已由 PR #45 合并；goal-2、goal-19 的符号已按合入代码核对：`router.NewMux() *Mux`、`(*Mux).Handle(name string, fn ServiceFunc)`、`(*Mux).Forward(fn ServiceFunc)`、`(*Mux).For(c Caller) wire.Handler`；`router.Plugin` 接口为 `PluginID() string`、`Activated() bool`、`Routes() []*pb.RouteDescriptor`、`Call(ctx, wire.Request) (wire.Response, error)`；`(*Router).Plugin(pluginID string) (Plugin, bool)`；SDK 的 `ServiceRegistry`、Root 的 `Services(*ServiceRegistry)` Hook，以及 dispatch.go 中先查服务路由、再查事件与命令路由的分发入口；BUS_SUBSCRIPTION 的注册校验在 `validate.Register`，事件名中的插件 ID 用 `validate.PluginID` 校验，callback_route_id 必须指向同一注册中的路由。goal-12（PR #47）已合并，sdk/dispatch.go 的分发入口现为 `invokeHandler`。

范围修订（2026-11-01，维护者确认）：执行者报告三处接口缺口，对照 origin/main 核实后按以下方式补齐，见语句「范围」「接口」「行为」与「测试」：
- `router.Plugin` 只有 `Call`：接口追加 `SendEvent`，唯一实现 `instancePlugin` 转调 `wire.Session.SendEvent`。原范围中的「Router 在 Attach/Detach 时通知 bus」取消：bus import router，Router 反向持有 Bus 会形成循环；Router 的 Attach/Detach 目前只在 supervisor 测试的装配处调用，由装配处在旁边同时调用 `bus.Attach`/`bus.Detach`。
- 现有 `SendEnvelopeSync` 走优先通道，通道满会关闭 Session，且不计业务预算：wire 新增 `SendEvent`，事件走业务 FIFO，按 CONTRACTS §1.2 计入本方向与全局 payload 预算，写出后才返回。只有写出后才返回，Bus 的每订阅者队列才是真正的积压上限；否则事件会全部堆进 wire 的 16 MiB 业务队列。
- EVENT 原本进 SDK 的 `activation.control`：改由 wire 的新钩子 `Options.Event` 接收。接收方向按同一预算准入，超额丢弃并计数；SDK 在 activate.go 设置该钩子（只加调用点），dispatch.go 不改。
- 原测试让 slow 在回调里阻塞，但 SDK 的读循环不停，事件写出后就离开 Core，Bus 队列不会满，Dropped 恒为 0。slow 改为在 Core 侧用带闸门的 `router.Plugin` 包装模拟写出受阻，丢弃数因此可以精确断言；回调阻塞的订阅者改测「不阻塞他人」。

```text
/goal 使用 $alonabot-execute 推进 goal-21（切片 T-06.3：非持久 Event Bus）。goal 目录 goal-21，分支 goal-21，worktree .worktrees/goal-21，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件经 core.bus.publish 向自己命名空间下的事件名发布，Core 只确认完成发布准入；Core 按订阅注册的事件名、schema 与版本，以无业务回执的 EVENT 帧分发给每个已激活的订阅者；每个订阅者有独立的 256 条、4 MiB 队列，满时丢弃新事件并计数，离线不缓存、不补发；慢订阅者不阻塞其他订阅者与发布者。

依据：PLAN T-06 步骤 5；CONTRACTS §3 最后一条、§1.2；WIRE-FORMAT §5；MESSAGE-MODEL §1（Event Bus 段）；TEST-MATRIX V-17。

范围（只新增或修改这些）：
- internal/wire/services.go：末尾追加 ServiceBusPublish = "core.bus.publish"
- internal/wire/session.go：新增 SendEvent、Options.Event 与 Stats.DroppedEvents，以及 EVENT 的收发预算
- 新增 internal/core/bus/
- internal/core/router/router.go：Plugin 接口追加 SendEvent，instancePlugin 实现它（Router 本身不改）
- sdk/：新增 bus.go（发布与订阅）；sdk.go 的 Collect 收集订阅（只加调用点）；activate.go 设置 wire.Options.Event（只加调用点）；session.go 的 buildRegister 输出订阅路由；base.go 的 Root 增加 Subscriptions Hook 与 Base 默认实现
- 新增 internal/core/supervisor/testdata/busplugin/
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/wire
   func (s *Session) SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error // Core 侧；target 为 callback_route_id
   Options 增加 Event func(env *pb.Envelope, release func()) // 为 nil 时 EVENT 仍交给 Control
   Stats 增加 DroppedEvents int64 // 接收方向预算不足而丢弃的 EVENT 数
2. internal/core/router
   Plugin 接口追加 SendEvent(ctx context.Context, target string, ev *pb.BusEvent) error // instancePlugin 转调 instance.Session.SendEvent
3. internal/core/bus
   type Bus struct{ ... }
   type Options struct { QueueEvents int; QueueBytes int64 } // BUS_QUEUE_EVENTS、BUS_QUEUE_BYTES
   func New(opt Options) *Bus
   func (b *Bus) Register(m *router.Mux) // 注册 core.bus.publish
   func (b *Bus) Attach(p router.Plugin) // 按插件的 BUS_SUBSCRIPTION 路由建立订阅与队列
   func (b *Bus) Detach(pluginID string) // 丢弃该插件队列中未发出的事件，取消正在等待写出的发送
   type Stats struct { Published, Delivered, Dropped uint64 }
   func (b *Bus) Stats(pluginID string) Stats // pluginID 为订阅者；Published 只对发布者有意义
   装配处在 router.Attach/Detach 旁边调用 bus.Attach/Detach，Router 不持有 Bus。
4. sdk
   Root 增加 Hook：Subscriptions(*SubscriptionRegistry) // func (*Base) Subscriptions(*SubscriptionRegistry) 默认不订阅
   type SubscriptionRegistry struct{ ... }
   func Subscribe[T any](r *SubscriptionRegistry, routeID, eventName, schemaID string, version uint32, h func(*Context, T) error) // JSON 载荷；eventName 为 "plugin_id/event_name"
   func Publish(ctx context.Context, eventName, schemaID string, version uint32, payload any) (eventID string, err error) // eventName 只写本插件命名空间内的名字，SDK 补上 plugin_id 前缀

行为：
- Publish（core.bus.publish，载荷为 PublishRequest，响应为 PublishResponse）：event_name 必须属于调用者的命名空间（"<调用者 plugin_id>/" 前缀，名字部分同 Service 名规则），否则 PERMISSION_DENIED；encoding 为 JSON 时载荷必须是合法 UTF-8 JSON；schema_id 与版本非空。准入通过后分配 event_id（16 字节随机值的十六进制），把 BusEvent 依次放进每个匹配订阅者的队列（入队操作不等待订阅者），然后立即响应。没有订阅者也成功。发布不携带任何可信 Actor 权限。
- 匹配：订阅的 event_name、schema_id 与 schema_version 都相等才投递；版本不同的订阅者不收到，也不计为丢弃。
- 队列：每个订阅者独立；条数达到 QueueEvents 或字节超过 QueueBytes 时丢弃新事件（drop-newest）并计数，以 WARN 记录（订阅者、事件名、累计计数）。订阅者未激活或已 Detach 时不入队、不缓存，Core 重启或插件重启不补发。
- 投递：每个订阅者一个发送 goroutine，按队列顺序逐条调用 Plugin.SendEvent（Target 为 callback_route_id，载荷为 BusEvent），等这一条写出后再发下一条，不等待业务回执；一个订阅者的 Session 慢或阻塞时只影响它自己的队列。SendEvent 返回错误（预算不足、Session 关闭、Detach 取消）时该事件计入 Dropped 并以 WARN 记录，不重试。Delivered 只表示已写出，不表示订阅者已处理。
- wire 发送（SendEvent）：构造 FRAME_KIND_EVENT（call_id 为 0，载荷为 protobuf 编码的 BusEvent）。入队前在 s.mu 下按本方向 outBytes/MaxBytes 与 Global 准入，不足立即返回 RESOURCE_EXHAUSTED，不等待；EVENT 不占在途 Request 条数。帧进入业务 FIFO，排在优先通道之后，与 REQUEST 共用唯一的 writer。写出后归还预算并返回；写出前 ctx 结束则撤出队列、归还预算并返回 ctx 的错误；Session 关闭返回 UNAVAILABLE。不经优先通道。
- wire 接收：Options.Event 非 nil 时，EVENT 不交给 Control，而是在读循环中按本方向 inBytes/MaxBytes（Global 非 nil 时也算）准入。不足时丢弃该帧、DroppedEvents 加一，Session 不关闭；通过时计入 inBytes 并调用 Event，release 归还预算，重复调用无副作用。Event 不得阻塞读循环。Core 侧不设置 Event，插件发来的 EVENT 仍由 Control 按协议违规关闭 Session；Adapter Session（sdk/adapter_run.go）不订阅，同样不设置。
- SDK：activate.go 设置 Event。ready gate 未打开或已收到 STOPPING 时，以 WARN 记录并立即 release，不关闭 Session。其余事件放进每个 Session 唯一的投递 goroutine，按到达顺序串行处理：按 callback_route_id 找到订阅，把 payload 按 JSON 解码为 T（失败以 WARN 记录并丢弃，不回应），以订阅插件自身身份调用 h（Context 没有 Actor 与默认 Target，Send 必须显式 To）。h 返回后 release。h 的错误与 panic 只记日志。广播不参与普通事件的 StopPropagation。
- Publish 在 Handler 内外都可以调用；在 Handler 内也不继承用户权限。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/bus_test.go（//go:build unix）：构建 testdata/busplugin，以 pub、fast、slow、stuck 四个插件 ID 启动。用 Bus{QueueEvents: 3} 装配，订阅者以 `router.FromInstance` 接入；slow 外面再包一层测试内的 router.Plugin，它的 SendEvent 先等测试关闭闸门再转调真实实现，模拟 Session 写出受阻；stuck 的订阅回调在插件内阻塞，直到测试放行。
  - 闸门关闭时 pub 先发布第 1 条，等包装层报告 SendEvent 已进入（第 1 条已离开队列）后，再发布第 2～10 条：Publish 全部立即返回；fast 收到全部 10 条且顺序不变；slow 的 Dropped 为 6（1 条等待写出、3 条排队）。开闸后 slow 按顺序恰好收到第 1～4 条，Delivered 为 4。
  - stuck 的回调阻塞期间，fast 与 pub 不受影响；放行后 stuck 按发布顺序处理已收到的事件。
  - pub 发布到 "fast/x"（不属于自己）返回 PERMISSION_DENIED。
  - schema 版本不同的订阅者收不到，Dropped 不变。
  - slow 停止（装配处 Detach）后再发布：不入队；slow 重新启动后收不到之前的事件。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：QueueBytes 的边界；非法 JSON；没有订阅者；订阅回调的 Context 没有 Actor；解码失败只记日志；Detach 丢弃未发出的事件并取消正在等待写出的 SendEvent；wire 的 SendEvent 在本方向或全局预算不足时立即返回 RESOURCE_EXHAUSTED、写出后归还预算、写出前 ctx 结束则撤出队列；接收方向预算不足时丢弃 EVENT 并计入 DroppedEvents，Session 不关闭。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：持久化与补发（契约不提供）；父调用链与取消（T-06.2）；Session 清理（T-06.4）。
```

### goal-26 · T-07.4 两个 Host 与插件票据

状态：已完成（[PR #53](https://github.com/AlonaBot/AlonaBot/pull/53)，合并提交 `07f0403`）。证据见 TEST-MATRIX §3.2。条目保留供后继切片查阅接口。

派发时的核对：前置 T-07.3 已由 PR #46 合并；goal-25 的符号已按合入代码核对：`httpproxy.Authorizer{ Authorize(r *http.Request, pluginID string) bool }`，`httpproxy.New(auth Authorizer) *Server`，以及 `(*Server).Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor)`、`Unbind(pluginID, generationID string)`、`ServeHTTP`；`settings.Settings` 的 `AdminBaseURL`、`PluginBaseURL` 为字符串，由 `settings.Load` 校验 scheme、端口与主机名，装配处用 `url.Parse` 转成 `*url.URL` 交给 webhost.Options；`wire.HeaderProxyAuth = "X-Alonabot-Proxy-Auth"`。cmd/alonabot 的 run 现在用 `httpproxy.New(nil)` 作为公开 Listener 的 Handler，并带 shortcut 注释。httpproxy.New 在构造时就要 Authorizer，而 webhost.Options 又需要 *httpproxy.Server：装配处先用一个转调 `host.Authorize` 的本地适配器构造 Server，再构造 Host，不需要改 httpproxy 的签名。

```text
/goal 使用 $alonabot-execute 推进 goal-26（切片 T-07.4：两个 Host 与插件票据）。goal 目录 goal-26，分支 goal-26，worktree .worktrees/goal-26，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
   func New(opt Options) *Host
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

状态：已完成（[PR #54](https://github.com/AlonaBot/AlonaBot/pull/54)，合并提交 `021e972`）。证据见 TEST-MATRIX §3.2。审查留给后继语句的 P3（`sdk/config` 能否供 Core 共用）已结清：goal-30 只在 SDK 内使用 `sdk/config`，goal-31 的 `internal/core/secret` 自行解析 Schema 并重复声明关键字常量，`MarshalSchema`、`SchemaHash` 不需要移到 `contracts/`。合并后核对 goal-34 时发现，本语句的标签键与 CONTRACTS 当时的嵌套写法 `config.<field>.label` 不一致。维护者按 D-74 改为仿照 AstrBot：标签统一为 `config` 标签，label 与 description 写源文案，翻译文件按字段名嵌套覆盖，取消选项图标，由 goal-90 修订生成器；下方语句中的 form、description、secret、hidden 标签与选项图标写法已被取代。条目保留供后继切片查阅接口。

```text
/goal 使用 $alonabot-execute 推进 goal-28（切片 T-08.1：配置字段描述与构建检查）。goal 目录 goal-28，分支 goal-28，worktree .worktrees/goal-28，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

### goal-67 · T-13.1 库模式入口与同进程组件

状态：等待前置。T-02.1、T-03.3、T-04.2 都已合并（PR #42 合入 sdk.Generated、sdk.UseGenerated 与生成器）。装配依赖的 `router.New` 正由 goal-12（PR #47）修改签名，`remote.Process.Bind` 正由 goal-7 增加 Caller 参数；两者合入后，审查者按实际代码核对语句中的 router、message.Service、remote.Process、supervisor、sdk.Run 与 sdk.RunAdapter 的内部流程以及 tune.Parse，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：数据目录缺省改为 `data/`，Core 装配与 `alonabot run` 共用（D-77）；构建期标记已删除（D-79），Adapter 包不用 `//alonabot:adapter` 标记，生成数据按包登记；开发 Adapter 的实例改由 goal-68 从 `core_config.json` 读取（D-82），本切片不读 instances/ 目录。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-67（切片 T-13.1：库模式入口与同进程组件）。goal 目录 goal-67，分支 goal-67，worktree .worktrees/goal-67，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：根包 alonabot 提供 Run(opts ...Option)、Start(ctx, opts ...Option) 与 Plugin、Adapter、DataDir、Args 选项：在当前进程装配一个 Core，已安装的包照常以独立进程运行，开发中的插件与 Adapter 在 goroutine 中经 loopback TCP 走完整的握手、注册、激活与停止。组件身份取自生成器嵌入的 manifest；Core 为它们创建 dev 来源的安装记录与固定代次 dev。Start 在全部组件激活后返回句柄，ctx 取消时按停止预算停止；Run 是 Start 加 SIGINT/SIGTERM 处理。组件在预算内没有退出时 Wait 与 Run 返回错误。

依据：PLAN T-13 步骤 1、2；CONTRACTS §15.5、§15.4（开发组件的规则）、§4.2、§9（manifest）、§16；DECISIONS D-47；ARCHITECTURE「根包」一行；TEST-MATRIX V-64 的生命周期分项。

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
- RunInProcess 与 RunAdapterInProcess 不替换 slog 默认 Logger（不调用 setupLogging）。
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

不做：config.json 直读、dev.sqlite 与拒绝 update/rollback（T-13.2）；热部署与 Reload（T-13.3）；alonabot run（T-09.4，复用本切片的装配）；HTTP 与设置文件的装配（随 T-07.5、T-09.4 接入）。
```

## 第 7 波

### goal-8 · T-03.8 按钮、extension 片段与依赖解析

状态：等待前置（T-03.2、T-03.7）。语句中 goal-2～goal-5、goal-50 提供的符号（message.Service、policy.Policy、router 的派发、remote.Process、AdapterDescriptor 与 ExtensionDeclaration、Button.core_token、adapters/telegram 的实例）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-8（切片 T-03.8：按钮、extension 片段与依赖解析）。goal 目录 goal-8，分支 goal-8，worktree .worktrees/goal-8，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 为发送中的每个按钮签发持久化的 token（绑定插件、button.clicked 路由、代次、Target、value 与 24h 期限），Telegram 点击先 answerCallbackQuery，再由 Core 验证 token、当前 Actor 与策略后以普通请求调用正确的 Session；extension 片段按目标 Adapter 注册时声明的 Schema 在发送前校验；manifest 解析 role（plugin/adapter/agent/webui）、adapter_id、icon、WebUI 的 assets 与 Admin API 主版本范围、adapter_dependencies 与 plugin_dependencies，开发入口启用插件时检查必需依赖。

依据：PLAN T-03 步骤 7；CONTRACTS §8（extension、依赖与按钮段）、§9（manifest）、§7.5（图标与 WebUI）；MESSAGE-MODEL §1.1、§6.3；DECISIONS D-26、D-43、D-59、D-62、D-63；TEST-MATRIX V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto：Segment 的 oneof 追加 ExtensionSegment extension = 9。随后 go generate ./...（本切片获授权只追加这一分支与消息）
- contracts/wirespec_test.go：追加编号检查
- contracts/manifest/：manifest.go 增加 role、adapter_id、icon、assets、admin_api_min_major、admin_api_max_major、adapter_dependencies、plugin_dependencies 字段与按角色的校验；新增 deps.go（依赖检查）
- contracts/validate/register.go：button.clicked 事件路由的 action 校验
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/buttons.go
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

状态：等待前置（T-03.7）。语句中 goal-2～goal-4、goal-50 提供的符号（message.Service、remote.Process、sdk/adapter 的 Instance、adapters/telegram 的发送与渲染）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-51（切片 T-03.9：消息编辑）。goal 目录 goal-51，分支 goal-51，worktree .worktrees/goal-51，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

### goal-87 · T-04.6 权限、扩展事件与命令解析的协议和 Core 侧

状态：已完成（PR #78）。审查中经维护者批准扩大范围两处（SCOPE-01、SCOPE-02，见下方范围的最后两项与接口 5）。

派发时：前置 T-03.3（PR #32）、T-03.6（PR #55，合并提交 `b023130`）、T-04.3（PR #47）均已合并。依据 D-69、D-71、D-72 新增。

已按合入代码核对：
- 旧写法的全部使用点：`remote/services.go` 交付时只保留 Adapter 给出的 `TRUSTED_ROLE_CHAT_ADMIN`；`router/command.go` 的 `commandTokens` 写死 `/` 前缀并对各 text 片段连接后的文本做 `strings.Fields`，`commandUsage` 读 `custom_usage`、usage 文本写死 `/`；`router/args.go` 的 `commandArgs(params, text)` 只处理文本；SDK 的 `buildRegister`、`appendServiceRoutes`、`appendSubscriptionRoutes`、`appendHTTPRoutes` 一律写 `PERMISSION_ANY`，`sdk.Usage(text)` 写 `custom_usage`；`supervisor/testdata/commanddispatch` 用了 `sdk.Usage`。Router 目前不检查任何路由权限。
- `validate.AdapterRegister` 只检查 capability 与 extension 声明；`validate.IsEventType` 是标准事件的封闭列表；`remote.adapterEventPayload` 按 event_type 与 payload 分支的对应关系接受平台事件。
- `router.New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(*pb.Event, Caller) (*pb.Event, error)) *Router`（goal-7 已追加 prepare）。生产装配尚未调用 `router.New`，调用点只在 supervisor 与 remote 的测试中。
- `sdk/adapter.Adapter.Descriptor()` 直接返回 `*pb.AdapterDescriptor`，Adapter 声明权限与扩展事件不需要新的 SDK 接口。Telegram 已把 `text_mention` 映射为带 `external_user_id` 的 `ENTITY_TYPE_TEXT_MENTION`，不声明平台权限（由 T-03.11 实现）。
- `remote_test.go` 的 loopAdapter 经 `sdk.RunAdapter` 以真实 Adapter 进程运行，可在 Descriptor 中声明权限与扩展事件。
- `github.com/google/jsonschema-go` 已批准且在 go.mod 中（`sdk/config` 使用），Core 与 SDK 根包的允许清单都包含它。

后继注意：
- `admin` 权限依据的 Core Admin 列表（`core_admin` 表）由 goal-13 实现。本切片的 Router 不赋予 `admin`；goal-13 按新契约改写时，以追加 `router.New` 参数的方式接入，并把 TrustedRoles 的写法改为权限名集合。
- 自定义 usage 的文案键由 goal-53（T-08.8）在 Core 文案目录接入时查找；goal-53 改写时补上这一项。
- `command_prefixes` 及按实例覆盖的持久化与管理由 goal-44（T-09.9）的 Core 设置提供；本切片只提供 Router 的注入点。

```text
/goal 使用 $alonabot-execute 推进 goal-87（切片 T-04.6：权限、扩展事件与命令解析的协议和 Core 侧）。goal 目录 goal-87，分支 goal-87，worktree .worktrees/goal-87，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按 WIRE-FORMAT §5A 的 2026-11-01 条目修改 proto 并重新生成；Core 校验 Adapter 的权限与扩展事件声明，交付时只保留已声明的平台权限并展开隐含关系，按声明校验扩展事件，路由按权限名集合判定；实现 command_prefixes 与按实例覆盖、去掉开头的 reference 与本账号 mention、非文本 token 与 user/image/message 参数、usage 文案键；同步已合并的 router、remote Adapter、SDK 与集成测试。

依据：PLAN T-04 步骤 8；CONTRACTS §2.2、§2.3、§8、§15.1；MESSAGE-MODEL §1、§1.2；WIRE-FORMAT §3.3、§5A；DECISIONS D-69、D-71、D-72；TEST-MATRIX V-85～V-87。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/ 的 identity.proto、route.proto、adapter.proto、message.proto、invocation.proto，然后 go generate ./...
- contracts/validate/：AdapterRegister 的权限与扩展事件声明、插件路由的 required_permissions、扩展事件的 event_type、新参数类型
- contracts/wirespec_test.go：新增与删除字段的编号检查
- internal/core/adapter/remote/：平台权限的过滤与展开、扩展事件的交付校验
- internal/core/router/：权限计算与判定、命令前缀、开头片段、按片段分词与新参数类型、usage
- sdk/：路由描述的 required_permissions、Usage 改为文案键、新参数类型的 Handler 绑定
- 上述包及 internal/core/supervisor 的测试与 testdata；router.New 的全部调用点
- 维护者批准（PR #78，SCOPE-01）：adapters/telegram/telegram_test.go 中 Actor 的断言由 TrustedRoles 改为 permissions，只改这一处
- 维护者批准（PR #78，SCOPE-02）：internal/core/host/activate.go 的 Instance 新增 Adapter 字段，Activate 从 REGISTER 保留；remote.Bind 从该字段读取权限与扩展事件声明
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto（删除的字段与枚举值一律 reserved 编号与名称，不复用编号）
   identity.proto：删除 enum TrustedRole；Actor 删除 trusted_roles = 4，新增 repeated string permissions = 6。
   route.proto：删除 enum Permission；RouteDescriptor 删除 required_permission = 5，新增 repeated string required_permissions = 17；RuleDescriptor 删除 require_chat_admin = 2；ParamType 新增 PARAM_TYPE_USER = 5、PARAM_TYPE_IMAGE = 6、PARAM_TYPE_MESSAGE = 7；CommandSpec 删除 custom_usage = 9，新增 string usage_key = 12（USAGE_MODE_CUSTOM 时为插件 i18/ 中 messages 的键）。
   adapter.proto：AdapterDescriptor 新增 repeated PermissionDeclaration permissions = 4、repeated EventDeclaration events = 5；
     message PermissionDeclaration { string name = 1; repeated string implies = 2; string label_key = 3; }  // name 与 implies 为完整名
     message EventDeclaration { string name = 1; uint32 version = 2; string schema_json = 3; }  // name 为完整名
   message.proto：EventPayload 的 oneof 新增 AdapterEventPayload adapter_event = 11；message AdapterEventPayload { uint32 version = 1; string payload_json = 2; }
   invocation.proto：ArgValue 的 oneof 新增 string user_id = 6、ImageSegment image = 7、MessageChain message = 8。
2. contracts/validate
   func IsPermissionName(name string) bool   // user、admin、chat_admin，或 <adapter_id>.<name>
   func IsAdapterEventType(eventType string) bool   // <adapter_id>.<name>
   完整名在最后一个 "." 处拆分：前半部分按 CONTRACTS §7.1 的本地 ID 规则（复用 validate.PluginID 的检查，且不带市场前缀），后半部分只用小写字母、数字与下划线且非空。
3. internal/core/router
   func New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(*pb.Event, Caller) (*pb.Event, error), prefixes func(adapterInstanceID string) []string) *Router
   // prefixes 为 nil 时所有实例取 ["/"]；返回值由设置层保证非空。
4. sdk
   func UsageKey(key string) RouteOption   // 取代 Usage(text)，删除 Usage
   type UserID string   // Handler 参数类型，对应 PARAM_TYPE_USER；值为 external_user_id
   Handler 参数类型映射新增：UserID → USER，*pb.ImageSegment → IMAGE，*pb.MessageChain → MESSAGE（goal-91 改为 SDK 自有类型）。
5. internal/core/host（SCOPE-02）
   Instance 新增字段 Adapter *pb.AdapterDescriptor   // 准备阶段已校验的 Adapter 描述，激活后只读；插件角色为 nil

行为：
- Adapter 注册（validate.AdapterRegister，失败则 Adapter 注册失败）：
  - 权限名与扩展事件名必须以 `<本 adapter_id>.` 开头且后半部分合法；同类声明内名称唯一。
  - implies 的每一项是本 Adapter 已声明的权限或 chat_admin；不能是 user、admin 或自身；隐含关系不能成环；label_key 非空。
  - 扩展事件的 version ≥ 1，同名只声明一个版本；schema_json 是 JSON 对象，并能用 jsonschema-go 解析与 Resolve。
- 平台权限（remote 交付时）：Adapter 在 Actor.permissions 中给出平台权限。Core 丢弃未声明的名字（包括 user、admin、chat_admin），每次丢弃记 WARN（plugin_id、instance_id、名称，作为诊断），再按声明展开隐含关系的传递闭包，结果去重后按字典序排列。Actor 为空时忽略。
- Actor 的权限集合（Router 在每个事件进入派发前计算，覆盖 remote 交付的事件、memory Adapter 与 Core 生成的事件）：actor 不为空时加入 user，再加上 remote 交付的平台权限与展开结果；来源事件中的 user 与 admin 不采信，本切片不赋予 admin。插件看到的 Actor.permissions 就是这个集合。
- 路由权限（Router，只作用于 COMMAND 与 EVENT 路由）：required_permissions 为空时放行；否则 Actor 权限集合与它有交集才放行，没有 Actor 不满足任何非空集合。
  - 已识别的命令路径被拒绝时消费该命令：不回复、不退到短路径、文本匹配或 Raw，记 DEBUG（plugin_id、route_id）；随后照常进入事件路由。
  - 文本匹配、Raw 与事件路由被拒绝时视为未匹配，继续下一条。
- 插件注册（validate.Register）：required_permissions 的每一项满足 IsPermissionName 且不重复；引用的 Adapter 不必已安装。事件路由的 event_type 是标准类型或满足 IsAdapterEventType；扩展事件路由不要求对应 Adapter 已安装，未安装或未声明时不会被触发。
- 扩展事件（remote 交付时）：event_type 为 `<本 adapter_id>.<已声明名>`，payload 分支为 adapter_event，version 等于声明的版本，payload_json 是符合声明 Schema 的 JSON 对象（注册时 Resolve 一次，交付时 Validate）；不符则按现有的非法事件处理拒绝并记 WARN。Actor、Target 与实例归属的校验与标准事件相同。扩展事件只进入 event_type_filter 含该完整名的事件路由，不进入命令层。
- 命令前缀（Router，message.created）：取 prefixes(事件的 adapter_instance_id)；去掉开头片段后的文本能匹配多个前缀时取最长者，空字符串前缀表示不带前缀；都不匹配时不识别命令路径，继续文本匹配与 Raw。首个路径 token 的 `@username` 规则保持不变。缺省 usage 中的前缀用本次匹配到的前缀。
- 开头片段：识别命令路径前，从 chain 开头连续去掉 reference 片段、external_user_id 等于事件 account_id 的 mention 片段以及只含空白的 text 片段，再去掉随后 text 的起始空白。文本匹配与 Raw 仍作用于完整 chain 中各 text 片段连接的文本。
- 分词与参数（替换 commandArgs 的文本入口）：
  - 路径 token 取开头片段之后、第一个非文本片段之前的文本中以空白分隔的词（与现有行为一致）。
  - 参数取路径之后的片段：每个 text 片段内部按现有引号与转义规则分词，引号与转义不跨片段，未闭合时报参数错误；每个 mention 与 image 片段各是一个非文本 token；其他片段不参与分词。
  - text 片段中范围恰好等于一个 text_mention 实体、且前后是空白或片段边界的文字作为一个文本 token（不在内部按空白切分），它可以绑定 user 参数，值为实体的 external_user_id；绑定 string 时值为其文字。
  - user 接受 mention token 或上述 text_mention token；image 接受 image token，值为该片段（含 incoming_ref）；string/int/float/bool 只接受文本 token。非文本 token 出现在其他类型的位置，或出现在 raw 尾部，都是参数错误。
  - message 只能是最后一个参数，与 raw 互斥：接收最后一个已消费参数之后的全部片段，去掉起始分隔空白，不再分词；没有剩余片段时视为缺失。
  - user、image、message 参数不能声明 default（validate 拒绝）。参数错误按现有规则回复一次 usage 或 silent。
- usage：缺省 usage 为 user、image、message 参数分别标出 `:user`、`:image`、`:message`。USAGE_MODE_CUSTOM 时 usage_key 必须非空（validate 拒绝空键）；在文案目录接入前按缺省 usage 发送，代码中以 `shortcut:` 注释写明由 T-08.8 的文案目录查找该键。
- SDK：命令与文本匹配路由的 required_permissions 缺省为 ["user"]；事件、Service、Bus 订阅与 HTTP 路由为空集合（权限的值对象与按路由设置由 goal-88 提供）。sdk.UsageKey 写入 usage_key。

测试（按编码规范「测试提交规则」）：
- 提交：
  - contracts/wirespec_test.go：上述新增字段与枚举值的编号，以及删除编号已 reserved。
  - internal/core/adapter/remote/remote_test.go：loopAdapter 声明 `<id>.owner`（隐含 `<id>.admin`）、`<id>.admin`（隐含 chat_admin）与一个扩展事件。错误前缀、隐含 user 或 admin、隐含成环、Schema 无法解析的声明各使 Adapter 注册失败；交付带 owner、未声明名与 user 的事件，插件收到的 Actor.permissions 为 user、`<id>.admin`、`<id>.owner`、chat_admin，并记丢弃诊断；扩展事件在名称未声明、版本不符、payload 不符合 Schema 时被拒绝，合法时只到达对应事件路由。
  - internal/core/supervisor/command_test.go（真实插件进程，TCP 与 stdio）：
    - 权限：required_permissions 为 ["<id>.admin"] 的命令被普通用户触发时不调用 Handler、不回复、不进入文本匹配；带该平台权限的用户成功；没有 Actor 的 message.created 不触发缺省 ["user"] 的命令。
    - 前缀：缺省 "/"；prefixes 返回 ["!", "!!", ""] 时 "!!x" 按 "!!" 匹配、无前缀的 "x" 也能触发；另一个实例仍只认 "/"。
    - 开头片段：reference + 指向本账号的 mention + " /kick" 识别为命令；指向他人的 mention 在开头时不识别。
    - 参数：`/kick @某人` 的 user 参数分别来自 mention 片段与 text_mention 实体；image 参数；末位 message 参数收到剩余的 text 与 mention 片段；mention 出现在 int 位置回复一次 usage。
    - 自定义 usage 键：参数错误时回复缺省 usage。
  - 已有集成测试按新字段与缺省权限同步（需要触发命令的事件补上 Actor），原断言保持。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：IsPermissionName 与 IsAdapterEventType 的边界；隐含闭包的多级传递；前缀的最长匹配与空前缀；引号不跨片段；text_mention 只覆盖部分词时不成为 user token；raw 尾部含 mention；message 参数为空；user/image/message 带 default 被拒绝；扩展事件路由在 Adapter 未安装时注册成功且不触发。

完成条件：gofmt -l 无输出；go generate ./... 后生成产物无差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。

不做：
- Core Admin 列表与 admin 权限、命令策略与全局禁止（T-04.4，goal-13）。
- command_prefixes 的持久化与管理（T-09.9）；usage 文案键的查找与语言选择（T-08.8）。
- 扩展事件路由的管理诊断展示（管理 API 切片）；PreEvent Hook 修改扩展事件 payload 后的重新校验（Hook 切片）；多轮等待（T-05）。
- SDK 的链式注册与 Permission 值对象（T-04.7，goal-88）；事件常量与 sdk.AdapterEvent（T-04.8，goal-91）。
- Telegram 与 OneBot 11 的平台权限映射（T-03.11、T-15.1）。
```

### goal-30 · T-08.3 SDK 配置准备顺序

状态：已完成（[PR #73](https://github.com/AlonaBot/AlonaBot/pull/73)），证据见 TEST-MATRIX §3.2。审查期间按授权改写 `sdk/generate/config_test.go` 的候选断言为规范 JSON，并在 `sdk/config` 包文档加入 Prepare 入口；`sdk/session.go` 另把 `PhaseLimits.config_ms` 缺失并入 Bootstrap 预算缺失的 PROTOCOL_ERROR。派发时：前置 T-08.1（PR #54）、T-08.2 已合并。goal-28 的符号已按合入代码核对：
- `sdk.Generated.ConfigSchema []byte` 是规范 JSON，没有 Config 时为 nil。
- `config.Config` 只有 `OnLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)` 一个方法，`config.Base` 的默认实现原样返回 raw。根插件经 `Config() config.Config` 提供配置，`sdk.Base` 返回 nil。
- 扩展关键字为 `config.KeywordSecret`、`KeywordHidden`、`KeywordForm`、`KeywordI18n`（`x-alonabot-*`）。
- 公开函数有 `config.MarshalSchema(*jsonschema.Schema) ([]byte, error)` 与 `config.SchemaHash(canonical []byte) []byte`。
- `sdk/session.go` 的 prepare 目前原样透传候选配置（step 1 注释为「目前原样透传」），REGISTER 由 `buildRegister(reg, boot.GetCandidateConfigJson(), boot.GetTargetConfigVersion(), generatedFor(root).ConfigSchema)` 构造。本切片在 OnStart 之前插入配置准备，并替换前两个参数。
- jsonschema-go v0.4.3 的 `Resolved.ApplyDefaults` 只沿 properties 递归，跳过 required 属性，也不进入 items 与 additionalProperties。生成器会给数组元素和 map 值内的属性写默认值，这部分由 Prepare 自行遍历补齐，见「行为」。

```text
/goal 使用 $alonabot-execute 推进 goal-30（切片 T-08.3：SDK 配置准备顺序）。goal 目录 goal-30，分支 goal-30，worktree .worktrees/goal-30，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
- 补默认值：用 jsonschema-go 的 Resolve（ValidateDefaults=true）与 ApplyDefaults，只给对象中缺失的属性补默认值；显式 false、0、""、[]、{} 与 null 都不覆盖；required 字段缺失且没有默认值时在下一步校验失败。ApplyDefaults 不进入 items 与 additionalProperties，数组的每个元素、map 的每个值若是对象，Prepare 按对应子 Schema 递归补齐其中缺失的属性。补默认值所用的 JSON 以 UseNumber 解码，默认值与原有数字都不经过 float64。
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
  - 数组元素与 map 值内缺失的属性得到默认值；元素中显式给出的值不被覆盖。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：规范 JSON 与 internal/core/config.Canonical 的结果逐字节一致（同一组输入）；ApplyDefaults 不触碰显式值；没有 Config 时的路径；OnLoad 超时；错误信息不含取值。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./sdk/... ./internal/core/...；go build ./...。

不做：Secret 的引用与解引用（T-08.4）；DB 迁移（T-08.5）；Pending 激活与 reload 流程（T-09）；表单（T-08.7）。
```

### goal-31 · T-08.4 不可变 Secret Store

状态：已完成（[PR #74](https://github.com/AlonaBot/AlonaBot/pull/74)），证据见 TEST-MATRIX §3.2。审查中发现原语句「状态库文件已是 0600」的前提不成立（data_dir 为 0700，文件按进程 umask 创建）；维护者在 PR 中确认扩围，范围、行为与测试已按合入代码改写。派发时：前置 T-02.3（PR #63，合并提交 `6b197d8`）、T-08.1（PR #54）、T-08.2 均已合并。

goal-28、goal-29 的符号已按合入代码核对：
- `config.KeywordSecret = "x-alonabot-secret"` 写在字段自身的属性 Schema 上，值为 `true`。secret 只允许 string 或 *string，可空时 type 为 `["string","null"]`。
- 数组元素与 map 值内的 secret 字段分别位于 `items.properties.<名>` 与 `additionalProperties.properties.<名>` 下，对应本语句的 `*` 路径段。
- `internal/core/config.Canonical(raw []byte) ([]byte, error)`：键排序、无多余空白，数字保持原文。
- `config.MarshalSchema` 只序列化 `*jsonschema.Schema`，本切片用不到。
- Core 不 import sdk：`internal/core/secret` 自行解析 Schema JSON，关键字名按「允许的新跨模块边」以常量重复声明。

已按 #63 合入的代码核对状态库：
- 表名 `secret`、`generation_secret_path` 符合单数命名。被引用的 `generation` 表主键是 `generation_id`，列 `config_snapshot` 存配置快照；`committed_at` 为 NULL 表示尚未提交。
- 现有模型都在 `models.go`；本切片把新模型和读写方法一起放在 `secrets.go`，不改 `models.go`，以减少并行冲突。
- `SetGenerationConfig` 改为在同一事务内写入快照和 secret 路径，避免快照和路径只写入一半。
- T-08.10（goal-90）把 secret 标记改为 `config:"secret"`。Schema 关键字 `x-alonabot-secret` 不变。测试插件按合并时 main 上的写法书写，后合并的一方同步。

```text
/goal 使用 $alonabot-execute 推进 goal-31（切片 T-08.4：不可变 Secret Store）。goal 目录 goal-31，分支 goal-31，worktree .worktrees/goal-31，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：配置快照中的 secret 字段只保存 {"$secret_ref": id}，Secret Store 保存不可变的 secret 版本。Core 在构造 Bootstrap 前把本插件有权访问的引用解析为明文，在 REGISTER 后按配置 Schema 的 secret 标记与此前保留的 secret 路径重新抽取，生成持久化快照，并以快照计算 PREPARED 的配置摘要；管理读取只返回引用是否已设置；按引用清理时保护 Active、Pending 与保留代次。

依据：PLAN T-08 步骤 4；CONTRACTS §5.3、§5.2（PREPARED 摘要）、§14.5、§12；WIRE-FORMAT §3.2；TEST-MATRIX V-25。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/secrets.go
- internal/core/state/state.go：Open 的建库与打开后核对；新增 internal/core/state/permissions_unix.go、permissions_windows.go（审查期间经维护者确认扩围）
- 新增 internal/core/secret/
- internal/core/host/prepare.go：Expect 增加快照回调；PREPARED 的摘要改为对快照计算；Prepared 增加 ConfigSnapshot
- internal/core/supervisor/supervisor.go：Spec 增加 ConfigSchema；新候选时把快照写入代次，普通 restart 时核对快照一致
- 上述包的测试文件
允许的新跨模块边：internal/core/secret -> sdk/config 不允许（internal/core 不 import sdk）；secret 标记的关键字名在 internal/core/secret 中以常量重复声明，并由测试核对与 sdk/config 的常量一致。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：
   CREATE TABLE secret (ref_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES install(plugin_id), value BLOB NOT NULL, created_at INTEGER NOT NULL);
   CREATE TABLE generation_secret_path (generation_id TEXT NOT NULL REFERENCES generation(generation_id), path TEXT NOT NULL, PRIMARY KEY (generation_id, path));
   func (s *Store) SetGenerationConfig(ctx context.Context, generationID, snapshot string, secretPaths []string) error // 只允许未提交的代次；同一事务内写入快照并替换该代次的 secret 路径
   其余读写方法自定。模型类型与读写方法都放在 secrets.go：TableName 指向表名，列名写在 column 标签，created_at 为毫秒，由调用方显式赋值（autoCreateTime:false）。
   读写遵守编码规范「Core 状态库不写 SQL」：
   - SetGenerationConfig 在事务内按主键读出代次，在 Go 中检查 committed_at 为 NULL（已提交时返回错误），改 config_snapshot 后 Save；再删除该代次原有的路径行，逐条写入新路径。
   - secret 按主键读取，在 Go 中核对 plugin_id。
   - 按插件或代次列出时用结构体条件，入口拒绝空的 plugin_id 与 generation_id。
   - Collect 先按插件列出 secret，在 Go 中算出未被引用的 ref_id，再用 clause.IN{Column: clause.PrimaryColumn, Values: ...} 删除；集合为空时不执行删除。
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
- REGISTER 之后：host 调用 Expect.Snapshot（supervisor 以 Extract 实现，schema=Spec.ConfigSchema，keptPaths=Spec.KeptSecretPaths 加上一代次的 generation_secret_path），得到快照与路径集合；PREPARED 摘要对快照计算。字段值为 null 或缺失时不生成引用；字段值已经是 {"$secret_ref": id} 时拒绝（插件不能回传引用）。
- 新候选：在提交 Active 之前，SetGenerationConfig 在同一事务内写入快照与 generation_secret_path。普通 restart：Extract 的结果必须与已提交的快照完全相同（明文未变时复用原引用），否则 FAILED_PRECONDITION，不提交新代次。
- 旧 secret 字段从 Schema 移除但仍作为未知键保留时，因为路径在 keptPaths 中而继续被抽取；只有候选中该键被删除，路径才不再保留。
- Redact 的结果用于管理 API；掩码不是协议值，表单提交时省略字段表示保持原值（T-09 与 T-08.7 遵守）。
- Collect 只删除不被任何 protected 快照引用的 secret；正在恢复或 Journal 中的引用由 T-09 加入 protected。
- 存储保护：Unix 上新状态库在 gorm.Open 之前以 O_CREATE|O_EXCL 和 0600 创建，WAL/SHM 由 SQLite 沿用主库权限；每次打开后核对主库及存在的 -wal、-shm 均为 0600，不符时 Open 返回带路径的错误并释放连接与锁，不自动修正。Windows 上在创建状态库之前为 data_dir 设置受保护、可继承、只授予当前账号完全控制的 DACL（golang.org/x/sys/windows，已批准），不检查 mode 位。
- 日志、错误与普通响应中不出现明文、Bootstrap 或完整配置；错误只写路径。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/secret_test.go（//go:build unix）。测试插件的配置含一个 secret 字段与一个普通字段，Spec.ConfigSchema 取生成器为它产生的 Schema：
  - 首次启动：插件收到明文；状态库代次快照中该字段为 {"$secret_ref": ...}，secret 表有一行；PREPARED 摘要等于快照的 SHA256。
  - 明文不变的新候选复用同一引用；明文改变生成新引用，旧代次快照中的引用不变。
  - 新 Schema 去掉该 secret 字段但候选保留该键：新快照中它仍是引用；候选删除该键后不再保留路径。
  - 普通 restart 时插件回传了不同的明文：FAILED_PRECONDITION，Active 不变。
  - Collect 在保护 Active 时不删除其引用，删除只被已丢弃候选引用的 secret。
  - 日志缓冲中不出现明文字符串。
  - 数组中的 secret（providers.*.api_key，两个元素）各自生成引用；调换元素顺序后按位置比较，明文变化的位置生成新引用。
- 提交：internal/core/state/permissions_test.go（unix）：子进程在 umask 022 下创建并写入真实状态库，主库、WAL、SHM 均为 0600；已有 0644 的库被拒绝，权限不被修改，失败后锁已释放。Windows DACL 只有交叉构建证据，实机验证记「未运行」。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：路径转义与 * 模式；Resolve 拒绝其他插件的引用；插件回传引用被拒；Redact 的输出；关键字常量与 sdk/config 一致。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Pending 中 secret 的写入与表单（T-09、T-08.7）；明文查看的再认证与审计（T-09.9）；Journal 中引用的保护（T-09.5）；Adapter 实例配置的 Secret（T-09.4）。
```

### goal-90 · T-08.10 配置标签统一

状态：已完成（[PR #77](https://github.com/AlonaBot/AlonaBot/pull/77)），证据见 TEST-MATRIX §3.2。实现按语句完成，`internal/core/supervisor` 中配置准备与 Secret Store 的测试插件一并改用新标签。派发时：前置 T-08.1 已由 PR #54 合并（合并提交 `021e972`）；依据 D-74 新增。

已核对现有实现：
- `sdk/generate/config.go` 的 `metadata` 分别读取 `description`、`secret`、`hidden` 三个独立标签。
- `configForm` 解析 `form` 标签，语法为 `key=value;flag`；default 按 JSON 字面量解码，因此可以含 `;`。
- options 的 `@图标路径` 由 `configBuilder.icon` 检查，写入 `x-alonabot-form.option_icons`。
- 每个字段写入 `x-alonabot-i18n`（`config.<json 路径>`）；`x-alonabot-form.label` 缺省等于它，description 为标签原文。
- `sdk/config/schema.go` 声明了 `KeywordI18n`，`sdk/config/config.go` 的包文档描述了旧标签语法。
- `sdk/generate/config_test.go` 的测试插件、期望 Schema 与诊断用例都使用旧标签。仓库中其他 Go 代码没有用到这些标签。
- 进行中的 goal-30 若先合并，它的测试插件使用旧标签，由本切片一并改写（docs/execution-order.md：后合并的一方同步）。

```text
/goal 使用 $alonabot-execute 推进 goal-90（切片 T-08.10：配置标签统一）。goal 目录 goal-90，分支 goal-90，worktree .worktrees/goal-90，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：仿照 AstrBot 的插件配置，Config 字段的表单与 secret 元数据统一写在 config 标签中，语法与 GORM 标签相同（key:value;key:value;flag）。label 与 description 是英文源文案，不再是 i18 键。表单特性补齐选项显示名与说明、多选、滑块、代码编辑器、醒目提示与显示条件；取消选项图标。

依据：CONTRACTS §5.1、§5.3、§7.5；DECISIONS D-74；TEST-MATRIX V-25 的描述分项。

范围（只新增或修改这些）：
- sdk/generate/config.go：标签的读取、解析、Schema 生成与诊断；删除选项图标的解析与检查
- sdk/config/schema.go：删除 KeywordI18n；sdk/config/config.go：包文档中的标签语法与示例
- sdk/generate/config_test.go：测试插件、期望 Schema 与诊断用例改用新写法
- 合并时 main 上已有的、使用旧标签的 Config（例如 goal-30 的测试插件）
允许的新跨模块边：无。

接口：删除 config.KeywordI18n，其余公开符号不变。

行为：
- 读取 config 标签。值键：label、description、widget、options、labels、subtitles、min、max、step、language、when、default；标志：required、secret、hidden、obvious。键与值以第一个 ":" 分隔。default 的值按 JSON 字面量解码（沿用现有解析，可以含 ";" 与 ":"），解码结束处之后必须是 ";" 或标签结尾；其余值键的值到下一个 ";" 为止。
- 各键的含义：
  - label、description：英文源文案，原样写入 Schema。没写 label 时不生成 label，表单回落到 JSON 名。
  - widget：text、password、textarea、number、slider、switch、select、code。
  - options：以 "|" 分隔的存储值，按字段（或 slice 元素）的类型解析；[]T 带 options 时为多选。
  - labels、subtitles：以 "|" 分隔的显示文案与选项说明，按下标与 options 对齐，可以比 options 短，不能更长。
  - min、max、step：数字字段的边界与步长；widget 为 slider 时必须有 min 与 max。
  - language：widget 为 code 时编辑器的语言，缺省为 json。
  - when：表单显示条件，形如 mode=regex，多个条件以 "," 连接。字段名指同一对象内的兄弟字段（按 JSON 名），值按该字段的类型解析：bool 与数字按 JSON 字面量，字符串原样。
  - obvious：提示醒目显示。
- Schema：
  - x-alonabot-form 写入 label、widget、options、labels、subtitles、min、max、step、language、when（对象：字段名 → 取值）、obvious；description 写入 Schema 的 description。
  - 不再写 x-alonabot-i18n 与 option_icons，其余关键字不变。
  - 翻译路径由 WebUI 按 Schema 的嵌套位置推出，不写入 Schema。
- 诊断（带位置，沿用现有诊断格式，英文）：
  - 未知键、标志带值、值键缺值、同一键出现两次。
  - widget 与类型不符：slider 与 number 只用于数字，switch 只用于 bool，select 需要 options，code 只用于 string。
  - labels 或 subtitles 多于 options；options 的取值不能解析为字段类型；slider 缺少 min 或 max；step 用于非数字字段；language 用于非 code 字段。
  - when 引用不存在的兄弟字段，或取值不能解析为该字段的类型。
  - 字段的 JSON 名为 label、description、labels 或 subtitles（与翻译文件的覆盖键冲突）。
  - 配置字段（含嵌套结构、slice 与 map 元素中的字段）上出现 form、description、secret 或 hidden 独立标签，提示改用 config 标签。
  - 现有诊断保持不变（例如 secret 只用于 string 与 *string）。
- 不保留旧标签与选项图标的兼容读取（编码规范「首次发布前不写兼容层」）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - sdk/generate/config_test.go 的 TestConfigSchemaIntegration 改用新写法，覆盖 labels、subtitles、多选、slider、code、obvious 与 when，期望 Schema 按上文更新，两次生成仍逐字节相同，REGISTER 的 config_schema_hash 与新 Schema 一致。
  - TestConfigDiagnosticsIntegration 覆盖上述每类新诊断各一例，其中旧标签一例放在嵌套结构的字段上。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - default 为含 ";" 与 ":" 的字符串、对象与数组字面量。
  - 标志与值键的不同顺序；未写 config 标签的字段。
  - when 引用 bool、数字与字符串字段。
  - 多选的 options 按元素类型解析。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./sdk/...；go build ./...。

不做：
- i18 文件内容的检查（T-08.8）；表单组件（T-08.7）；运行时的配置准备（T-08.3）。
- D-74 列为以后再考虑的特性：JSON 文件定义配置元数据、Core 数据源选择、文件上传、多模板列表。
```

### goal-36 · T-09.1 首次管理认证

状态：已完成（[PR #72](https://github.com/AlonaBot/AlonaBot/pull/72)），证据见 TEST-MATRIX §3.2。审查期间按裁决改 `New` 的签名：原签名拿不到 Origin 校验所需的 admin_base_url，改为接收该 URL，并由 Base URL 的 scheme 推导 secure（见接口 2）。派发时：前置 T-02.3（PR #63，合并提交 `6b197d8`）、T-07.4（PR #53）均已合并。已按 #63 的模型与命名约定核对表结构：
- `admin_credential`、`admin_session` 符合单数命名。
- 模型类型与读写方法一起放在 `admin.go`，不改 `models.go`。
- 单行表的整数主键要关闭 GORM 的自增推断。
- 全部会话的删除要显式允许全表操作。

符号已按合入代码核对：`webhost.AdminSessions{ Lookup(r *http.Request) (sessionID string, expiresAt time.Time, ok bool); Valid(sessionID string) bool }`，Host 只在 admin host 的请求上调用 Lookup 并记录会话期限，签发、兑换票据与每次插件授权都调用 Valid，所以 reset 与登出只要让 Valid 返回 false，派生的票据与插件会话就随之失效；`webhost.New(opt webhost.Options) *Host`，`Options{AdminBaseURL, PluginBaseURL *url.URL; SecondHost bool; Admin http.Handler; Proxy *httpproxy.Server; Sessions AdminSessions; Now func() time.Time}`，Admin 为 nil 时管理入口返回 404；`(*Host).IssueTicket(adminSessionID, pluginID string) (ticketURL string, err error)`。cmd/alonabot 的 run 现在只传 AdminBaseURL、PluginBaseURL、SecondHost 与 Proxy，本切片补上 Admin 与 Sessions。Host 与插件响应策略已过滤 `alonabot_` 与 `__Host-alonabot_` 前缀的 Cookie；插件会话 Cookie 为 `__Host-alonabot_plugin`（HTTPS）或 `alonabot_plugin`，管理会话 Cookie 不要与之同名。`settings.Settings` 的 AdminBaseURL、PluginBaseURL 是字符串，由 settings.Load 校验；命令树在 cmd/alonabot/commands.go（cobra）。

```text
/goal 使用 $alonabot-execute 推进 goal-36（切片 T-09.1：首次管理认证）。goal 目录 goal-36，分支 goal-36，worktree .worktrees/goal-36，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：alonabot admin-token init/reset 在持有 core.lock 时离线生成 32 字节管理 token，只在本机终端显示一次，状态库只存 SHA256 与认证代次；reset 递增代次并撤销全部管理会话、兑换票据与派生插件会话。admin host 提供 POST /api/auth/login（Origin 与登录 CSRF、恒时比较、8h 会话）与 POST /api/auth/logout；管理状态变更要求 CSRF；CLI 用显式 Bearer token。实现 webhost.AdminSessions，未初始化或认证无效时拒绝一切管理访问。

依据：PLAN T-09 步骤 1；CONTRACTS §7.3、§7.4、§15.1（listen 与 Base URL）、§14.5；TEST-MATRIX V-47 的认证分项。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/admin.go
- 新增 internal/core/adminauth/
- internal/core/settings/：非 loopback 的 listen 必须配合 https 的 admin_base_url 或非空的 trusted_proxies（只加校验）
- cmd/alonabot/：新增 admin-token 命令组（init、reset）；run 用 adminauth 作为 webhost 的 AdminSessions 并挂上登录接口
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：
   CREATE TABLE admin_credential (id INTEGER PRIMARY KEY CHECK (id = 1), token_sha256 BLOB NOT NULL, auth_generation INTEGER NOT NULL, updated_at INTEGER NOT NULL);
   CREATE TABLE admin_session (session_id TEXT PRIMARY KEY, csrf_token TEXT NOT NULL, auth_generation INTEGER NOT NULL, expires_at INTEGER NOT NULL);
   读写方法自定。模型类型与读写方法都放在 admin.go：TableName 指向表名，列名写在 column 标签，updated_at、expires_at 为毫秒，由调用方显式赋值（autoUpdateTime:false）。
   读写遵守编码规范「Core 状态库不写 SQL」：
   - admin_credential 的 id 字段标 primaryKey;autoIncrement:false，取值为命名常量 1，按 clause.Eq{Column: clause.PrimaryColumn, Value: 1} 读取。
   - InitToken 在事务内读不到凭证时才 Create，已有凭证时返回错误。
   - ResetToken 在事务内读出凭证（不存在时返回 ErrNotInitialized），auth_generation 加一并写入新摘要后 Save；同一事务内用 Session(&gorm.Session{AllowGlobalUpdate: true}) 删除全部 admin_session。
   - 会话按主键读取与删除，入口拒绝空的会话 ID；是否过期、代次是否一致在 Go 中判断。
   - 登录创建会话时，同一事务内删除已过期的会话：用 clause.Lte 与包内列名常量 expires_at。
2. internal/core/adminauth
   var ErrNotInitialized = errors.New("admin token is not initialized")
   func InitToken(ctx context.Context, st *state.Store) (token string, err error)  // 已有凭证时返回错误；token 为 32 字节随机值的 base64url（无 padding）
   func ResetToken(ctx context.Context, st *state.Store) (token string, err error) // 递增 auth_generation，删除全部管理会话
   type Auth struct{ ... }
   func New(st *state.Store, now func() time.Time, adminBaseURL *url.URL) *Auth // adminBaseURL 为 https 时 Cookie 用 __Host- 前缀与 Secure；Origin 比较 scheme、主机名与有效端口，不比较 path
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
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：替换终端检测后 init 只显示一次 token、reset 递增代次；恒时比较；Cookie 属性在 http 与 https 的 adminBaseURL 下的差异；settings 的非 loopback 校验；日志中不出现 token；InitToken 在已有凭证时失败；Reauthenticate。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：管理 API 的业务接口与页面（T-09.4、T-09.8）；serve 的首次启动生成 token（T-09.4）；多用户账号（契约不提供）；明文查看 secret 的接口（T-09.9）。
```

## 第 8 波

### goal-13 · T-04.4 持久化命令策略

状态：已完成（PR #88）。`policy` 的覆盖、Admin 与禁止名单方法、`router.CommandPolicy` 与 `router.New` 的第六个参数已合入；生产装配尚未调用。验证证据见 TEST-MATRIX §3.2。

派发时的核对：前置 T-03.2（PR #31）、T-04.1（PR #34）、T-04.6（PR #78，合并提交 `0a320c5`）均已合并。2026-11-04 按 D-69 的权限名集合改写；同日按 D-82 修订：Core Admin 列表写在 `core_config.json` 的 `admins`，本切片只在内存中保存由装配处设置的列表，不建 `core_admin` 表，增删由管理服务写回该文件（T-07.6、T-09.4）。TrustedRole 与 Permission 枚举已删除，路由权限是 `required_permissions`，Router 自己计算 Actor 的权限集合。符号已按合入代码核对：
- `router.New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(*pb.Event, Caller) (*pb.Event, error), prefixes func(adapterInstanceID string) []string) *Router`。调用点只在测试中：`internal/core/adapter/remote/remote_test.go` 与 `internal/core/supervisor` 的 index、media_in、echo、bus、service、command、conversation 测试；生产装配尚未调用。
- `router.go` 的 `dispatch` step 1：actor 不为空时，权限 = {user} ∪（来源权限中除 user、admin 以外的名字），排序去重后写回 `actor.Permissions`；不赋予 admin。`permitted(candidate, ev)` 按 `route.required_permissions` 判定：空集合放行，否则与 Actor 权限有交集才放行，拒绝时记 DEBUG（plugin_id、route_id）。
- `command.go` 的 `dispatchCommands`：已识别命令路径的胜出路由未激活或 `permitted` 失败时返回 false（消费命令，不回复，不进入文本匹配与 Raw，随后照常进入事件层）；文本匹配与 Raw 被拒时跳过该候选。参数解析与 usage 在 `invokeCommand` 中，位于权限判定之后。
- `RouteDescriptor` 有 `route_id`、`default_enabled = 4`、`required_permissions = 17`；`CommandSpec.path` 是规范路径（别名在 `aliases`）。
- `internal/core/policy`：`Policy{store}`、`New(store)`、`CheckSend(ctx, pluginID, inv *router.Invocation, target) *pb.ErrorInfo`：先做原 Target 回复豁免，再查实例许可。policy import router，所以 router 只能经接口使用策略。
- `validate.IsPermissionName(name)`：接受 user、admin、chat_admin 与 `<adapter_id>.<name>`。
- 状态库只有唯一的初始迁移 `internal/core/state/schema.sql`，读写经 GORM 的 `s.tx`。

```text
/goal 使用 $alonabot-execute 推进 goal-13（切片 T-04.4：持久化命令策略）。goal 目录 goal-13，分支 goal-13，worktree .worktrees/goal-13，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：
- 以 plugin_id 加作用域（插件、命令组、路由）持久化权限集合与启用覆盖。
- 命令派发前依次判定：Core 全局禁止 → 插件、祖先组或自身任一层禁用即拒绝 → 最近一层显式权限集合（都没有时用路由声明）。
- Core 依据 Core Admin 列表给 Actor 加上 admin 权限。
- Send 准入先检查会话禁止。
- 提供管理服务层的查询与修改方法：持久化成功后才返回成功，运行时修改在下一次派发时生效。

依据：PLAN T-04 步骤 5；CONTRACTS §2.2、§2.3 第 1、2 条、§7（Core Admin 列表）；MESSAGE-MODEL §1；DECISIONS D-45、D-69；TEST-MATRIX V-12、V-43 的服务层分项。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入下面两张表；新增 internal/core/state/policy.go
- internal/core/policy/：新增 command.go、admins.go、denies.go；policy.go 的 CheckSend 增加会话禁止检查
- internal/core/router/：router.go（New 追加参数、dispatch 加入 admin）、command.go（命令路由改用策略判定），只改这些调用点
- router.New 的全部调用者（上面列出的测试）同步新签名
- 上述包的测试文件，以及新增 internal/core/supervisor/policy_test.go
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state（schema.sql 与 policy.go）
   CREATE TABLE command_policy (plugin_id TEXT NOT NULL REFERENCES install(plugin_id), scope TEXT NOT NULL, permissions TEXT, enabled INTEGER, PRIMARY KEY (plugin_id, scope));
   -- permissions 为权限名的 JSON 数组（可以是 []），NULL 表示不覆盖；enabled 为 NULL 表示不覆盖
   CREATE TABLE global_deny (kind TEXT NOT NULL, adapter_instance_id TEXT NOT NULL, subject_id TEXT NOT NULL, PRIMARY KEY (kind, adapter_instance_id, subject_id)); -- kind 为 user 或 chat
   读写方法自定，放在 policy.go，全部按主键幂等。CheckCommand 需要的读取在一个只读事务中完成。
2. internal/core/policy
   type ScopeKind int // ScopePlugin、ScopeGroup、ScopeRoute
   type Scope struct { Kind ScopeKind; Group []string; RouteID string } // Group 用于 ScopeGroup（命令路径的前缀 token）；RouteID 用于 ScopeRoute
   type Override struct { Permissions []string; Enabled *bool } // Permissions 为 nil 表示不覆盖，非 nil 的空切片表示不要求权限；Enabled 为 nil 表示不覆盖
   type ScopedOverride struct { Scope Scope; Override Override }
   func (p *Policy) SetOverride(ctx context.Context, pluginID string, s Scope, o Override) error // 两项都不覆盖时等同删除；Permissions 按字典序保存
   func (p *Policy) Overrides(ctx context.Context, pluginID string) ([]ScopedOverride, error) // 按插件、组（路径字典序）、路由排序
   func (p *Policy) CheckCommand(ctx context.Context, pluginID string, route *pb.RouteDescriptor, actor *pb.Actor, target *pb.Target) *pb.ErrorInfo // nil 表示放行；actor.permissions 已由 Router 计算（含 admin）
   func (p *Policy) IsAdmin(ctx context.Context, actor *pb.Actor) (bool, error)
   type Admin struct { InstanceID, UserID string }
   func (p *Policy) SetAdmins(admins []Admin) // 整体替换内存中的 Core Admin 列表，可与 IsAdmin 并发；装配处在启动与重新读取 core_config.json 后调用（T-07.6），此前列表为空
   func (p *Policy) Admins() []Admin // 返回当前列表的副本，按 InstanceID、UserID 排序
   type DenyKind string // DenyUser = "user"、DenyChat = "chat"
   type Deny struct { Kind DenyKind; InstanceID, SubjectID string }
   func (p *Policy) Deny(ctx context.Context, kind DenyKind, instanceID, subjectID string) error
   func (p *Policy) Undeny(ctx context.Context, kind DenyKind, instanceID, subjectID string) error
   func (p *Policy) Denies(ctx context.Context) ([]Deny, error)
3. internal/core/router
   type CommandPolicy interface {
       CheckCommand(ctx context.Context, pluginID string, route *pb.RouteDescriptor, actor *pb.Actor, target *pb.Target) *pb.ErrorInfo
       IsAdmin(ctx context.Context, actor *pb.Actor) (bool, error)
   }
   func New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(*pb.Event, Caller) (*pb.Event, error), prefixes func(adapterInstanceID string) []string, policy CommandPolicy) *Router
   // policy 为 nil 时不赋予 admin，命令路由按 required_permissions 判定（与现在相同）；装配处传入同一份 *policy.Policy，router 不 import policy

行为：
- admin：Router 在 dispatch step 1 计算权限后，actor 不为空且 IsAdmin 为 true 时加入 "admin"，再排序去重。IsAdmin 出错时记 WARN（adapter_instance_id 与错误），不加 admin，事件照常派发。插件与 Adapter 给出的 user、admin 一律不采信（已有行为保持）。admin 不自动满足其他权限名。
- CheckCommand 只作用于命令层（PATH、文本匹配与 Raw 的路由），事件路由仍按 permitted 判定。顺序如下，任一步拒绝即返回：
  1) 全局禁止：actor 存在且 (adapter_instance_id, external_user_id) 在 user 禁止表中，或 target 存在且 (adapter_instance_id, chat_id) 在 chat 禁止表中 → PERMISSION_DENIED。Admin 身份与任何策略覆盖都不能绕过。
  2) 启用：插件作用域、规范路径（CommandSpec.path，与触发时用的别名无关）的每一层祖先组、路由自身，任一层显式 Enabled=false 即拒绝（FAILED_PRECONDITION，safe_message "command is disabled"）。路由自身没有显式 Enabled 时取 route.default_enabled。子项的 Enabled=true 不能抵消祖先的 false。
  3) 权限：取最近一层显式的 Permissions（路由 → 最长的祖先组 → 插件），都没有时取 route.required_permissions。空集合放行；否则 actor.permissions 与它有交集才放行，没有 Actor 不满足任何非空集合 → PERMISSION_DENIED。策略中引用未安装 Adapter 或未声明的名字时保留设置，这个名字不被任何事件满足。
  文本匹配与 Raw 路由没有路径，祖先组为空，只看插件与路由自身。
  读取状态库失败 → INTERNAL，按拒绝处理。
- 每次 CheckCommand 都从状态库读取当前策略，不缓存；修改在下一次派发时生效，已开始的调用不受影响。
- Router：policy 不为 nil 时，命令层用 CheckCommand 代替 permitted（在参数解析和 usage 发送之前）。拒绝后的处理与现在相同：已识别的命令路径被消费（不回复、不进入文本匹配与 Raw），随后照常进入事件层；文本匹配与 Raw 被拒时跳过该候选。拒绝记 DEBUG（plugin_id、route_id、错误码，不含消息内容）。
- CheckSend：最先检查 chat 禁止，目标 (adapter_instance_id, chat_id) 被禁止 → PERMISSION_DENIED，原 Target 回复豁免也不能绕过；之后保持原有顺序。
- 管理服务层：写操作在事务提交成功后才返回 nil。以下参数非法时返回 INVALID_ARGUMENT：
  - 未知的 ScopeKind；
  - ScopeGroup 的 Group 为空或含非法 token（validate.IsCommandToken）；
  - ScopeRoute 的 RouteID 为空；
  - Permissions 中的名字不满足 validate.IsPermissionName 或有重复；
  - kind 未知；instanceID、userID 或 subjectID 为空。
  遮蔽诊断直接使用已有的 Router.Commands().Shadows()，本切片不重复实现。
- 卸载插件时 command_policy 的清理由 T-09.4 负责（外键引用 install）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/policy_test.go（//go:build unix），沿用 command_test.go 的命令装配与 commanddispatch 测试插件，接入真实状态库上的 policy.Policy，TCP 与 stdio 各跑一次：
  - required_permissions 为 ["user"] 的 /rss list 普通用户可用；对插件作用域设 Permissions ["admin"] 后普通用户被拒（不回复、Handler 与文本匹配都未被调用）；加入 Core Admin 后放行，插件收到的 Actor.permissions 含 admin。
  - 组 rss 设 ["admin"]、路由 rss list 设 []：list 对普通用户放行（最近一层显式覆盖）；rss add 仍要求 admin。
  - 策略设为 ["loop.owner"] 这类未声明的名字：设置被保留，任何用户都被拒。
  - 组 rss 禁用、路由 rss list 显式启用：list 仍被拒（父层禁用优先）；取消组禁用后放行，插件不重启。
  - 用户加入 user 禁止表后，即使是 Core Admin 也被拒；会话加入 chat 禁止表后，命令被拒，且 sdk.Send 发往该会话（包括回复原 Target）返回 PERMISSION_DENIED。
  - Adapter 给出的 admin 被丢弃，非 Core Admin 用户的 Actor.permissions 不含 admin。
  - 关闭并重新打开状态库后策略仍在。
- 已有测试按 router.New 的新签名传 nil，原断言保持。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - CheckCommand 的完整判定表：default_enabled=false、无 Actor、多层组、别名触发时按规范路径取组、空集合覆盖。
  - SetOverride 两项都不覆盖时删除该行；Permissions 排序，重复返回 INVALID_ARGUMENT；各类非法参数。
  - IsAdmin 读取失败时不加 admin。
  - 并发修改与检查在 -race 下无数据竞争。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。

不做：管理 API、CLI 与 WebUI，以及在管理视图中标出未声明的权限名（T-09.4、T-09.14、T-09.9）；/whoami（T-14）；Rule 与 chat_admin Rule（T-05.2）；Hook（T-05.3）；SDK 的按路由权限设置（T-04.7）。
```

### goal-88 · T-04.7 SDK 链式注册

状态：待派发。前置 T-04.2、T-04.6（PR #78，合并提交 `0a320c5`）、T-06.1、T-06.3、T-08.1 均已合并。依据 D-70 新增；2026-11-04 经维护者同意，把原切片拆为两部分：本切片只做链式注册，SDK 自有类型、事件常量与 `sdk.AdapterEvent` 移到 T-04.8（goal-91）。符号已按合入代码核对：
- `sdk/registry.go`：
  - 命令：`CommandRegistry.Command(name, handler any, opts ...RouteOption)`、`Group(name string, fn func(*CommandGroup))`（`CommandGroup = CommandRegistry`）。
  - 文本匹配与 Raw：`Regex(routeID, expr, handler, opts...)`；`Keyword`、`StartsWith`、`EndsWith`、`FullMatch` 为 `(routeID string, patterns []string, handler, opts...)`；`Raw(routeID, handler, opts...)`。
  - `RouteOption`：`Alias`、`Priority`、`Args`、`Optional`、`RawTail`、`UsageKey`、`Silent`、`IgnoreCase`。
  - `register` 用 `checkHandler` 从函数签名推出 HandlerMode 与参数类型，再经 `validate.Register` 校验单条描述；`add` 检查路径、别名与 route_id 的冲突。
  - 命令的 route_id 由路径生成（`commandRouteID`），文本匹配与 Raw 必须显式给出且不能用保留前缀。
- 缺省权限：`sdk/session.go` 的 `buildRegister` 给命令与文本匹配写 ["user"]；`appendServiceRoutes`、`appendSubscriptionRoutes`、`appendHTTPRoutes` 与事件路由为空集合。SDK 目前没有按路由设置权限的入口。
- `sdk/events.go`：`EventRegistry.On(routeID, eventType string, handler func(*Context) error, opts ...RouteOption)`，事件类型须为 `validate.IsEventType` 的标准类型或 `validate.IsAdapterEventType`。
- `sdk/services.go`：
  - `JSONService[In, Out any](r, name, schemaIn, schemaOut string, version uint32, h, opts ...ServiceOption)`。
  - `ProtoService[In, Out proto.Message](r, name, h, opts...)`：schema_id 取自消息类型（`validate.SchemaID`），版本为 1。
  - `ServiceOption` 只有 `WithAccess(a Access, allowlist ...string)`。
  - route_id 为 `svc.` 加服务名；SDK 只按 schema_id 与 version 核对调用，不校验业务字段。
- `sdk/bus.go`：`Subscribe[T any](r *SubscriptionRegistry, routeID, eventName, schemaID string, version uint32, h func(*Context, T) error)`、`Publish(ctx, eventName, schemaID string, version uint32, payload any) (string, error)`；`CallJSON[Out any](ctx, target, schemaIn string, version uint32, in any) (Out, error)` 在 `services.go`。`validate.Register` 只要求 schema_id 非空、版本非零。
- 自动回复：`sdk.go` 的 `Context.Send` 在 `autoReply` 为 true 时返回 FAILED_PRECONDITION（"auto-reply handler cannot call Send"）；`dispatch.go` 在 Handler 成功返回非空 Chain 时把它放进 HandlerOutcome。
- 生成器：`sdk/generate/source.go` 生成 `Commands: func(r *sdk.CommandRegistry)`（用 `r.Group(..., func(r *sdk.CommandGroup){...})` 表达父 token）与 `Events: func(r *sdk.EventRegistry)`（`r.On(id, type, fn, sdk.Priority(n))`）。`config.go` 的 `configBuilder` 用 go/types 把 Go 类型转换为 JSON Schema，并只在有 Config 时做类型检查。
- HTTP 与 Schedule 的注册不在 D-70 的范围内，保持原样（`HTTPOption`、`Every`、`Cron`）。

```text
/goal 使用 $alonabot-execute 推进 goal-88（切片 T-04.7：SDK 链式注册）。goal 目录 goal-88，分支 goal-88，worktree .worktrees/goal-88，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：
- 命令、文本匹配、Raw、事件、JSON/Protobuf Service 与 Event Bus 订阅的注册改为链式构建，以 Handle 结束；删除 RouteOption、ServiceOption 与位置参数形式，不留别名。
- 新增 Permission 值对象（Or）与 c.Has；分组上设置的权限由组内路由继承。
- JSON Service 的输入/输出 Schema 由生成器从 Go 类型生成，SDK 用它校验业务字段；schema_id 与版本由 Go 类型决定，不再手写字符串。
- 自动回复型 Handler 也可以 c.Send。
- 生成器、示例与已合并的全部调用方同步修改。

依据：PLAN T-04 步骤 9（链式注册部分）；CONTRACTS §2.1、§2.3、§3；DESIGN §D；DECISIONS D-69、D-70；TEST-MATRIX V-10、V-14 的 SDK 分项。

范围（只新增或修改这些）：
- sdk/：registry.go、events.go、services.go、bus.go、sdk.go（Context.Send 的限制与 Has）、session.go（buildRegister 写入路由权限，只改这一处）、generated.go（Generated 增加 ServiceSchemas）；新增 permission.go
- sdk/generate/：source.go 生成链式调用；新增 service.go，从 sdk.JSON 的实例化类型生成 Schema（可以复用并拆分 config.go 中与配置标签无关的类型转换）
- 已合并的调用方：examples/plugins/、internal/core/supervisor/testdata/，以及测试中以字符串嵌入的插件源码（sdk/generate、internal/core/supervisor、internal/core/adapter/remote 的测试文件）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 权限（sdk/permission.go）
   type Permission struct{ ... }  // 不可变的权限名集合
   var (
       Anyone    Permission // 空集合：不要求 Actor
       User      Permission // {"user"}
       Admin     Permission // {"admin"}
       ChatAdmin Permission // {"chat_admin"}
   )
   func PermissionName(name string) Permission // Adapter 声明的完整名，如 "onebot11.group_owner"；名字不满足 validate.IsPermissionName 时，使用它的路由在 Collect 时报错
   func (p Permission) Or(others ...Permission) Permission // 并集，去重并按字典序
   func (p Permission) Names() []string // 返回副本
   func (c *Context) Has(p Permission) bool // p 为空集合时为 true；否则当前 Actor.permissions 与 p 有交集才为 true，没有 Actor 时为 false
2. 命令与文本匹配（sdk/registry.go）
   func (r *CommandRegistry) Group(name string) *CommandRegistry   // 返回子注册器；取代 Group(name, fn)，删除 CommandGroup 别名
   func (r *CommandRegistry) Permission(p Permission) *CommandRegistry // 本组的权限，组内路由与子组继承最近一层显式设置
   func (r *CommandRegistry) Command(path ...string) *CommandBuilder // 相对本组的路径 token，至少一个
   func (r *CommandRegistry) Regex(routeID, expr string) *CommandBuilder
   func (r *CommandRegistry) Keyword(routeID string, patterns ...string) *CommandBuilder
   func (r *CommandRegistry) StartsWith(routeID string, patterns ...string) *CommandBuilder
   func (r *CommandRegistry) EndsWith(routeID string, patterns ...string) *CommandBuilder
   func (r *CommandRegistry) FullMatch(routeID string, patterns ...string) *CommandBuilder
   func (r *CommandRegistry) Raw(routeID string) *CommandBuilder
   type CommandBuilder struct{ ... }
   func (b *CommandBuilder) Alias(path ...string) *CommandBuilder   // 完整别名路径，不附加组前缀；只用于 Command
   func (b *CommandBuilder) Args(names ...string) *CommandBuilder
   func (b *CommandBuilder) Optional(name string, def ...string) *CommandBuilder
   func (b *CommandBuilder) RawTail() *CommandBuilder
   func (b *CommandBuilder) UsageKey(key string) *CommandBuilder
   func (b *CommandBuilder) Silent() *CommandBuilder
   func (b *CommandBuilder) IgnoreCase() *CommandBuilder
   func (b *CommandBuilder) Priority(p int32) *CommandBuilder
   func (b *CommandBuilder) Permission(p Permission) *CommandBuilder
   func (b *CommandBuilder) Handle(handler any)   // 签名规则与现在的 checkHandler 相同
   删除 RouteOption 及 Alias、Priority、Args、Optional、RawTail、UsageKey、Silent、IgnoreCase 这些包级函数。
3. 事件（sdk/events.go）
   func (r *EventRegistry) On(routeID, eventType string) *EventBuilder
   type EventBuilder struct{ ... }
   func (b *EventBuilder) Priority(p int32) *EventBuilder
   func (b *EventBuilder) Permission(p Permission) *EventBuilder
   func (b *EventBuilder) Handle(h func(*Context) error)
   （T-04.8 把 eventType 改为带 payload 类型的事件常量）
4. Service 与 Event Bus（sdk/services.go、sdk/bus.go）
   func JSON[In, Out any](r *ServiceRegistry, name string) *JSONService[In, Out]
   type JSONService[In, Out any] struct{ ... }
   func (s *JSONService[In, Out]) Access(a Access, allowlist ...string) *JSONService[In, Out]
   func (s *JSONService[In, Out]) Handle(h func(*Context, In) (Out, error))
   func Proto[In, Out proto.Message](r *ServiceRegistry, name string) *ProtoService[In, Out]
   type ProtoService[In, Out proto.Message] struct{ ... }
   func (s *ProtoService[In, Out]) Access(a Access, allowlist ...string) *ProtoService[In, Out]
   func (s *ProtoService[In, Out]) Handle(h func(*Context, In) (Out, error))
   func Subscribe[T any](r *SubscriptionRegistry, routeID, eventName string) *Subscription[T] // eventName 为 plugin_id/event_name
   type Subscription[T any] struct{ ... }
   func (s *Subscription[T]) Handle(h func(*Context, T) error)
   func CallJSON[In, Out any](ctx context.Context, target string, in In) (Out, error)
   func Publish[T any](ctx context.Context, eventName string, payload T) (eventID string, err error)
   删除 JSONService、ProtoService、ServiceOption、WithAccess 与旧的 Subscribe、Publish、CallJSON 签名；CallProto 不变。
   JSON 类型的 schema_id 为 Go 类型的「包路径.类型名」（reflect 的 PkgPath() + "." + Name()）；版本取类型的 SchemaVersion() uint32 方法（值接收者），没有该方法时为 1。In、Out、T 必须是包级的具名非泛型类型，否则 Collect 报错（CallJSON 与 Publish 返回 INVALID_ARGUMENT）。
5. 生成项（sdk/generated.go）
   Generated 增加 ServiceSchemas map[string][]byte // schema_id → 规范 JSON Schema，覆盖本插件全部 JSON Service 的输入与输出类型

行为：
- 链：
  - 每个由注册器方法开始的链必须以 Handle 结束。漏写 Handle、Handle 调用两次、在 Handle 之后再调用其他方法，都在 Collect 时报错，错误中带 route_id（命令带路径）。
  - 同一方法重复调用时取最后一次（Alias 与 Optional 是追加）。
  - 只适用于命令路径的方法（Alias）用在文本匹配或 Raw 上是注册错误；原有的其他组合规则（RawTail 需要参数、Optional 的默认值个数等）保持不变。
- 权限：
  - 命令与文本匹配的 required_permissions 取最近一层显式 Permission（路由 → 最近的祖先组 → 更外层的组），都没有时为 ["user"]。
  - 事件路由显式设置时用其值，否则为空集合；Service、Bus 订阅与 HTTP 路由仍为空集合。
  - Anyone 写为空集合。Collect 时对每个名字检查 validate.IsPermissionName。
- 分组：Group(name) 返回的子注册器与父注册器共享冲突检查与错误收集；在不同位置对同名组多次调用 Group 时，得到同一组路径下的路由，冲突照常报告。
- JSON Service：
  - 生成器在构建入口所在 module 的已扫描包中，找出 sdk.JSON 的全部实例化（go/types 的 Instances），为每个 In 与 Out 类型生成 JSON Schema，按 schema_id 写入 Generated.ServiceSchemas。
  - Schema 只依据 json 标签与 Go 类型，不读取 config 标签，支持的类型与配置 Schema 相同；不支持的类型给出带位置的诊断。
  - In 或 Out 不是具名类型时报告诊断；两个不同类型得到同一 schema_id 时也报告。
  - Collect 时 Generated.ServiceSchemas 缺少某个 JSON Service 的 schema_id，报错 "service schema for <id> is not generated; run alonabot plugin generate"。
  - 调用时，SDK 先用 jsonschema-go 把输入 JSON 按 Schema 校验，失败返回 INVALID_ARGUMENT 且不调用 Handler；Handler 返回后，输出按 Schema 校验，失败返回 INTERNAL 并记 ERROR（只记 schema_id 与校验错误，不记载荷）。
- Proto Service：schema_id 与版本仍取自消息类型，只改为链式。
- 自动回复：删除 Context.Send 在自动回复型 Handler 中的限制。Handler 内的 c.Send 与手动型相同；Handler 返回的非空 Chain 照旧由 Core 在 Handler 返回后发送，因此排在本次调用的全部显式发送之后。
- 生成器：source.go 为标记生成的命令与事件输出链式调用（r.Group(...).Command(...).Alias(...).Priority(...).Handle(fn)、r.On(id, type).Priority(n).Handle(fn)），输出保持稳定；`alonabot plugin check` 照常比较。
- 已合并的调用方全部改为新 API，原有集成测试的断言保持。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/command_test.go（沿用 commanddispatch 测试插件，TCP 与 stdio）：
    - 组设 Permission(sdk.Admin.Or(sdk.ChatAdmin))：组内路由的 required_permissions 为 ["admin","chat_admin"]；其中一条路由设 Permission(sdk.Anyone) 后为空集合，没有 Actor 的 message.created 也能触发。
    - Handler 内 c.Has(sdk.ChatAdmin) 按 Actor 的权限返回。
    - 自动回复型 Handler 先 c.Send 一条，再返回 Chain：Adapter 依次收到显式发送与自动回复两条。
  - internal/core/supervisor/service_test.go（沿用 svcplugin）：JSON Service 的输入缺少必需字段或类型不符时返回 INVALID_ARGUMENT，Handler 未被调用；Handler 返回不符合 Schema 的输出时，调用方收到 INTERNAL；带 SchemaVersion() 返回 2 的类型以版本 2 注册，用 CallJSON 调用成功。
  - sdk/generate/service_test.go（沿用 generate_test.go 的临时 module 辅助函数）：两个 JSON Service 的输入/输出类型生成的 Schema 与固定期望逐字节一致，两次生成相同；非具名类型与不支持的类型各给出带位置的诊断；生成文件过期时 plugin check 报错。
  - 已有集成测试改用新 API，断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 漏写 Handle、Handle 两次、Handle 之后调用方法；Alias 用在 Regex 上；PermissionName 不合法。
  - 多层组的权限继承与最近一层覆盖；同名组多次 Group 的冲突检查。
  - Permission.Or 的去重与排序；Names 返回副本。
  - schema_id 的推导（具名、指针、泛型、匿名类型）与 SchemaVersion。
  - Publish 与 CallJSON 对非具名类型返回 INVALID_ARGUMENT。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./...；go build ./...。

不做：SDK 自有的 Message、Target、Actor、SendResult、错误码与入站片段，标准事件常量、sdk.AdapterEvent 与事件 Handler 的 payload 类型核对，sdk/adapter 的公共类型（T-04.8，goal-91）；Rule（T-05.2）；HTTP 与 Schedule 注册的写法；Service 的 tool 描述（T-12）；docs/plugin-guide.md 的更新（由审查者在合并前同步）。
```

### goal-97 · T-07.6 Core 设置与首次启动

状态：待派发。前置 T-07.3（PR #46）与 T-09.1（PR #72）均已合并。依据 D-82 新增。

已核对现有实现（main `1c07284`）：
- `internal/core/settings` 读取 `<data_dir>/alonabot.yaml`（`FileName`），`Settings` 有 `Listen`、`AdminBaseURL`、`PluginBaseURL`、`TrustedProxies`、`LogLevel`、`LogFormat`、`LogFile{Enabled, MaxSize, MaxTotal, MaxAgeDays}`、`Proxy`；`Load` 严格解析（未知键、重复键与类型错误带完整键名和行号），文件不存在时以 0600 创建缺省内容；`Defaults`、`ValidateListen` 供调用方使用。
- `cmd/alonabot/main.go` 用 `settings.Load` 的结果设置日志、日志文件、出站代理、`webhost.New`（`AdminBaseURL` 与 `PluginBaseURL`）与 Listener，命令行的 `--listen`、`--log-level`、`--log-format` 只覆盖本次运行；`router.New` 的 prefixes 参数传 nil，所有实例使用 `["/"]`。
- `cmd/alonabot/admin_token.go` 的 `admin-token init|reset` 要求 stdin 与 stderr 都是终端，先 `state.Open` 取得 core.lock，再调用 `adminauth.InitToken` 或 `ResetToken`，只向 stderr 写一次 token。Core 启动时不会自动初始化管理 token。
- `adminauth.Auth` 由 `adminauth.New(store, now, adminBaseURL)` 构造，`login` 成功后签发会话。
- Core Admin 列表、命令前缀的按实例覆盖、按来源的日志级别与 Adapter 实例目前都没有持久化实现。

```text
/goal 使用 $alonabot-execute 推进 goal-97（切片 T-07.6：Core 设置与首次启动）。goal 目录 goal-97，分支 goal-97，worktree .worktrees/goal-97，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：数据目录的 core_config.json 取代 alonabot.yaml，是 Core 设置的唯一来源：启动设置、Core Admin 列表、命令前缀与按实例的覆盖、默认 locale、按来源的日志级别与 Adapter 实例都在这个文件里。字符串值可以写 ${环境变量}，写回时保留原写法。Core 首次启动且尚无管理凭证时自动生成管理 token：终端中显示一次，非终端时写入数据目录的 admin_token，第一次登录成功后删除；新增 alonabot admin-token show。

依据：CONTRACTS §15.1、§15.3、§7.4、§2.2、§2.3、§14.3；DECISIONS D-82；TEST-MATRIX V-88、V-47 的认证分项、V-22 的组件分项。

范围（只新增或修改这些）：
- internal/core/settings/：改为读取与写回 core_config.json（删除 YAML 的读取与 alonabot.yaml）
- internal/core/adminauth/：首次 token 文件的生成与登录成功后的删除
- cmd/alonabot/main.go：装配新的设置、命令前缀与首次 token；cmd/alonabot/admin_token.go：增加 show
- 上述包的测试，以及仓库中写 alonabot.yaml 的测试与 testdata
- docs/development.md 中提到 alonabot.yaml 的段落
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/settings
   const FileName = "core_config.json"
   type Admin struct { InstanceID, UserID string } // JSON 为 {"instance": "...", "user": "..."}
   type AdapterSettings struct { Version int; Instances map[string]json.RawMessage } // Instances 为展开 ${...} 后的实例配置原文
   type Settings struct {
       Listen, AdminBaseURL, PluginBaseURL string // 两个 Base URL 由 T-07.7 合并为 base_url
       TrustedProxies []netip.Prefix
       LogLevel slog.Level; LogFormat string; LogFile struct { Enabled bool; MaxSize, MaxTotal int64; MaxAgeDays int }
       LogSources map[string]string // 组件 ID → debug、info、warn、error、off 或 inherit
       Proxy string
       Admins []Admin
       CommandPrefixes []string; CommandPrefixesByInstance map[string][]string
       DefaultLocale string
       Adapters map[string]AdapterSettings // 按 adapter_id
   }
   type File struct{ /* 不导出 */ }
   func Load(dataDir string) (*File, error)                  // 文件不存在时以 0600 创建缺省内容；展开 ${...}，校验全部键
   func (f *File) Settings() Settings                         // 当前设置的副本
   func (f *File) Update(edit func(doc *Document) error) error // 在未展开的原文上修改，校验通过后原子写回（临时文件加重命名，权限 0600）；失败时文件与当前设置都不变
   func (f *File) Reload() (changed []string, err error)       // 重新读取文件，返回值有变化的顶层键（按字典序）；失败时保留当前设置
   type Document struct{ /* 不导出 */ }
   func (d *Document) Set(path []string, value any) error      // 设置一个键，保留文件中其他键的顺序与原写法
   func (d *Document) Delete(path []string) error
   func Defaults() Settings、func ValidateListen(address string) error 保持
2. internal/core/adminauth
   func FirstToken(ctx context.Context, st *state.Store, dataDir string, interactive bool) (token string, err error)
   // 已有凭证时返回 "" 与 nil；interactive 为 true 时初始化并返回 token，由调用方写到终端；为 false 时初始化并写入 <dataDir>/admin_token（0600），返回 ""
   New 增加参数 dataDir string，login 成功后删除 <dataDir>/admin_token（文件不存在时忽略）
3. cmd/alonabot
   alonabot admin-token show --data-dir DIR：读出并输出 admin_token，文件不存在时以英文错误退出；不取 core.lock，不要求终端

行为：
- 文件格式：JSON 对象，键与缺省值见 CONTRACTS §15.1 的表；这一阶段 `base_url` 仍是 `admin_base_url` 与 `plugin_base_url` 两个键（缺省值与现在相同），由 T-07.7 合并。未知键、重复键、类型或取值错误时 Load 失败，错误包含完整键路径（例如 `adapters.telegram.instances.main`）与行列号。`$version` 当前为 1，缺失时按 1 处理并在下次写回时补上；版本迁移由 T-08.12 实现，本切片遇到大于 1 的版本时报错。
- ${环境变量}：只有整个字符串值等于 `${NAME}`（NAME 为字母、数字与下划线，不以数字开头）时展开；变量不存在时 Load 失败并指出键路径。Update 在原文上修改，写回时保留原来的 `${...}`，不把展开后的值写回。
- 写回：键保持文件中的原有顺序，新键追加在所在对象末尾，缩进两个空格，末尾换行。
- 命令行参数 `--listen`、`--log-level`、`--log-format` 照旧只覆盖本次运行，不写回文件。
- 装配：main.go 用 Settings 设置日志、日志文件、代理、webhost 与 Listener；router.New 的 prefixes 改为按实例取 CommandPrefixesByInstance，没有覆盖时取 CommandPrefixes。Core Admin 列表：goal-13（T-04.4）已合并时在装配处调用 `policy.SetAdmins`；尚未合并时由 goal-13 合并时补上这一行（执行顺序的共享文件规则）。LogSources、DefaultLocale 与 Adapters 只解析与保存，由后续切片使用（T-09.10、T-08.8、T-09.4、T-13.3）。
- 首次启动：Core 打开状态库后调用 FirstToken；stdin 与 stderr 都是终端时把返回的 token 只写一次到 stderr；否则写入 admin_token，并在日志中以 INFO 输出该文件的路径（不输出 token）。admin-token init 与 reset 的行为不变。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/settings/settings_test.go：缺省创建与 0600 权限；${NAME} 展开与缺失时的键路径；未知键与重复键的位置；Update 修改一个键后其他键的顺序与 ${...} 原写法不变，校验失败时文件不变；Reload 返回变化的顶层键；`$version` 为 2 时报错。
  - cmd/alonabot 的集成测试（//go:build unix）：非终端首次启动后数据目录出现 0600 的 admin_token，`admin-token show` 输出的 token 能登录，登录后文件被删除、再次 show 报错；core_config.json 中设置按实例的命令前缀后，该实例的消息按新前缀匹配命令。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：终端首次启动只向 stderr 写一次；写回中途失败时原文件不变；Windows 上的文件权限（交叉编译检查，不在 Windows 运行）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/settings/... ./internal/core/adminauth/... ./cmd/alonabot/...；go build ./...。

不做：合并两个 Base URL 与删除插件 host（T-07.7）；core_config.json 与 Adapter 实例配置的版本迁移和备份（T-08.12）；Admin API 与 CLI 对这些设置的修改入口（T-09.4、T-09.14）；`alonabot run --reload` 对该文件的监视（T-13.3）；日志级别的运行时修改（T-09.10）。
```

### goal-32 · T-08.5 受管 SQLite 连接与迁移

状态：已完成（PR #79）。审查中经维护者批准扩大范围一处（FIX-3，见下方范围的最后一项）。

派发时的核对：前置 T-08.3（PR #73，合并提交 `225f132`）已合并。符号已按合入代码核对：
- `sdk/session.go` 的 prepare：step 1 核对 Bootstrap 预算（缺 `config_ms`、`on_start_ms` 或准备截止时为 PROTOCOL_ERROR）；step 2 在 `config_ms` 预算内调用 `config.Prepare`，失败以 `SESSION_PHASE_CONFIG` 报告；随后启动 HTTP Listener，step 3 运行 OnStart，step 4 用 `buildRegister(reg, configJSON, configVersion, configSchema)` 构造 REGISTER。`buildRegister` 目前把 `Migration` 固定为 `MIGRATION_MODE_NONE`。Adapter 角色在 `sdk/adapter_run.go` 中自行构造 REGISTER，同样固定为 NONE，`validate.AdapterRegister` 也要求 Adapter 的 migration 为 NONE；本切片不改 `adapter_run.go`。
- `sdk/activate.go` 的 `shutdown` 分三步：关闭 ready gate 并取消 lifetime → OnStop → 关闭 Session。`currentSession()` 只在 ready 之后、STOPPING 之前可用，准备期（迁移与 OnStart）拿不到它。
- `sdk/gormdb` 现有公开符号：`OpenSQLite(ctx, path) (*gorm.DB, error)`、`PragmaValues`、`ReadPragmas(ctx, *gorm.DB)`、`SelfMigrator`、`Migration{ID uint32; Up func(context.Context, *gorm.DB) error}`、`VersionedMigrator`、`ModeNone`/`ModeSelf`/`ModeVersioned`、`CheckMigrators(root any) (string, error)`、`LedgerTable`、`RunVersioned`、`AppliedMigrations`。账本由 `AutoMigrate(&ledgerRow{})` 建立，列为 `id`（主键）与 `applied_at`（UTC 毫秒）。
- `github.com/libtnb/sqlite` v1.2.2 提供 `sqlite.New(sqlite.Config{Conn: gorm.ConnPool})`，驱动名常量 `sqlite.DriverName = "sqlite"`，即 modernc 注册的名字。
- 协议：`DatabaseDescriptor{backend, sqlite_path, migration_mode, applied_migration_ids, activation_reason, manual_migration_approved}`、`PhaseLimits.migration_ms`、`SESSION_PHASE_MIGRATION`、`MigrationState{mode, completed_migration_ids, self_migrate_completed}` 都已在 proto 中。
- `supervisor.Spec` 已有 `PluginDir`（`<plugins>/<id>`）、`GenerationID`、`Bootstrap`、`NewCandidate`；数据库描述目前由调用方写在 `Spec.Bootstrap.DatabaseDescriptor` 中（cmd/alonabot 与全部现有测试传 `BACKEND_NONE`）。

核对时修订了原语句的三处：
- 原「允许的新跨模块边」含 `sdk -> sdk/sqlitedb`。这会让每个插件都链接 modernc SQLite，违反 D-52（SDK 根包只允许 protobuf、x/sys 与 jsonschema-go）和 D-49（不导入就不链接）。改为 SDK 根包只 import `sdk/internal/dbhook`，Session 的数据库描述、连接与关闭都经 dbhook 交接。
- 原接口 3 新增 `Spec.Database`，nil 时生成 SQLite 描述。这与已有的 `Bootstrap.DatabaseDescriptor` 重复，并会把现有调用方的 nil 当成 SQLite。改为不加 Spec 字段，沿用 Bootstrap 中的描述，只由 supervisor 补默认路径；调用方给出的非空路径原样下发，供 goal-33 的候选副本与 goal-68 的 `dev.sqlite` 使用。
- 补充迁移阶段的预算校验与失败阶段、`Open` 在准备期的可用范围、准备失败时关闭连接、两个包共用的账本 DDL。

```text
/goal 使用 $alonabot-execute 推进 goal-32（切片 T-08.5：受管 SQLite 连接与迁移）。goal 目录 goal-32，分支 goal-32，worktree .worktrees/goal-32，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：受管 SQLite 有两个入口：新建的 sdk/sqlitedb 的 Open(ctx) 按当前 Session 的数据库描述返回 *sql.DB；sdk/gormdb 的 Open(ctx) 在同一个连接上返回原生 *gorm.DB。同一 Session 复用同一连接，SDK 停止时关闭。只导入 sqlitedb 的插件不链接 GORM。两个包各定义一组迁移接口，被导入时向 SDK 登记识别方式；SDK 在配置准备之后、OnStart 之前执行迁移，并在 REGISTER 中报告结果。Core 为每个代次下发 SQLite 路径。

依据：PLAN T-08 步骤 5、6；CONTRACTS §6.1～§6.3、§4.2（迁移预算）；WIRE-FORMAT §3.2（database_descriptor）；DECISIONS D-48、D-49；TEST-MATRIX V-26、V-39。

范围（只新增或修改这些）：
- sdk/sqlitedb/：新建包。连接设置与回读核对从 sdk/gormdb/sqlite.go 移入；open.go（Open 与 Session 绑定）、migrate.go（迁移接口、RunVersioned、AppliedMigrations）
- sdk/gormdb/：OpenSQLite 改为 sqlitedb.OpenFile 加 GORM 包装；新增 open.go（Open）；ReadPragmas 与 PragmaValues 移到 sqlitedb 后从本包删除
- sdk/internal/dbhook/：新建内部包，供 sqlitedb 与 gormdb 向 sdk 登记迁移识别与执行，形状由执行者自定，只用标准库与 contracts 类型
- sdk/：session.go 在配置准备之后调用迁移阶段并填写 REGISTER 的 migration，准备失败时关闭连接；activate.go 在停止时关闭连接（只加调用点）
- sdk/gormdb/sqlite_test.go：随 ReadPragmas 移到 sqlitedb 同步改写
- internal/core/supervisor/supervisor.go：backend 为 SQLITE 且 sqlite_path 为空时补默认代次路径，新候选时创建代次目录（只加这一段）
- internal/core/host/prepare.go：校验 REGISTER 的 migration 与 Bootstrap 描述一致（只加校验）
- 上述包的测试文件
- 维护者批准（PR #79，FIX-3）：cmd/alonabot/main.go、internal/core/adapter/remote/remote_test.go、sdk/generate/config_test.go 中已有的 DatabaseDescriptor 字面量补上 MigrationMode: MIGRATION_MODE_NONE，不做其他改动
允许的新跨模块边：无（以下均在 sdk 模块内）。包内依赖为 sdk -> sdk/internal/dbhook、sdk/sqlitedb -> sdk/internal/dbhook、sdk/gormdb -> sdk/sqlitedb 与 sdk/internal/dbhook。sdk 根包不得 import sdk/sqlitedb、sdk/gormdb、gorm.io/gorm、github.com/libtnb/sqlite 或 modernc.org/sqlite（D-49、D-52）。

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
   不新增公开符号。数据库描述沿用 Spec.Bootstrap.DatabaseDescriptor：backend、migration_mode、applied_migration_ids、activation_reason 由调用方给出；backend 为 SQLITE 且 sqlite_path 为空时，supervisor 在 Launch 使用的 Bootstrap 副本中填写 <PluginDir>/generations/<GenerationID>/db.sqlite（绝对路径），非空时原样下发。
4. sdk/internal/dbhook（内部包，形状自定）
   只用标准库与 contracts 类型。至少承载：已登记的迁移识别与执行函数；当前 Session 的数据库描述（准备阶段由 sdk 写入）；sqlitedb 打开的连接及其关闭函数（sdk 在停止与准备失败时调用）。并发访问加锁。

行为：
- Open：可用范围为迁移阶段开始到 Session 停止（含 Migrate、Up 与 OnStart 中的调用），范围外返回错误码为 FAILED_PRECONDITION 的错误。文件不存在时只有 activation_reason 为 first_install 时才创建，否则返回错误（不创建空库掩盖丢失；OpenFile 本身不做这项判断）。连接错误原样返回。插件拿不到 Core 状态库的任何路径或连接。
- 连接生命周期：同一 Session 只打开一次，sqlitedb.Open 与 gormdb.Open 共用它；activate 的 shutdown 在 OnStop 返回后、关闭 Session 之前关闭连接；准备阶段任一步失败（迁移、OnStart、REGISTER、等待 PREPARED）时也关闭。gormdb.Open 返回的 *gorm.DB 不另行关闭底层连接。
- 登记：sqlitedb 与 gormdb 在包 init 中向 sdk/internal/dbhook 登记一个识别函数（只追加函数值，不做 I/O）；SDK 的迁移阶段只调用已登记的识别函数，因此未导入某个包时不认识它的接口。
- 迁移阶段（SDK，配置准备之后、OnStart 之前，预算为 Bootstrap 的 PhaseLimits.migration_ms，与普通业务期限无关；step 1 的预算校验加入 migration_ms，缺失时与 config_ms 一样报 PROTOCOL_ERROR。迁移阶段的任何失败以 SESSION_PHASE_MIGRATION 报告，不进入 OnStart；超时为 DEADLINE_EXCEEDED）：
  - 根插件实现的迁移接口合计多于一个（含两个包混用，例如 sqlitedb 的 Migrate 加 gormdb 的 Migrations）时失败（INVALID_ARGUMENT）。
  - backend 为 NONE：跳过，migration.mode=NONE。
  - versioned：ID 必须非零、唯一、严格递增；每步 Up 与 _alonabot_migrations 的完成记录在同一事务中，失败不记成功。两个包用同一条 DDL 建账本：CREATE TABLE IF NOT EXISTS _alonabot_migrations (id INTEGER PRIMARY KEY, applied_at INTEGER NOT NULL)（applied_at 为 UTC 毫秒）；gormdb 的 RunVersioned 改用这条 DDL，不再 AutoMigrate。
  - self：调用 Migrate，成功后 self_migrate_completed=true。
  - REGISTER.migration：mode 与 completed_migration_ids（versioned，账本中的全部 ID，升序）或 self_migrate_completed（self）如实填写。Adapter 角色不执行迁移，仍为 NONE。
- prepare_effects=external 的维护门槛由 Core 在建立候选前判断（T-09），本切片的 SDK 不读取 manifest 的 prepare_effects。
- Core：supervisor 只在 NewCandidate 且使用默认路径时创建 generations/<g>/ 目录（0700），普通 restart 的 GenerationID 就是 Active 代次，沿用它的路径，不创建目录；调用方给出的 Bootstrap 不被修改（与 Secret 解析共用同一个副本）。host 校验 REGISTER.migration.mode 与 Bootstrap 的 migration_mode 一致、versioned 的 completed 包含 Bootstrap 的 applied 集合，否则注册失败。
- 事务内不得等待同一单连接池的另一条独立查询（文档注释写明，测试覆盖一次死锁会被 busy timeout 以外的方式暴露为错误）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/db_test.go（//go:build unix）：两个测试插件，一个只导入 sqlitedb、一个用 gormdb，各实现 Migrations（两步，第二步按环境变量可失败）并在 OnStart 中写一行：首装成功，REGISTER 报告 [1,2]，插件代次目录 <PluginDir>/generations/<g>/ 下有 db.sqlite 且 PRAGMA 值正确；第二步失败时 SESSION_ERROR 阶段为 MIGRATION、注册失败、账本只有 1；普通 restart 不重复执行迁移；同时实现两个迁移接口的插件在 CONFIG 之后以 MIGRATION 阶段失败。
  - 同一测试中断言只导入 sqlitedb 的测试插件二进制不含 gorm.io/gorm（`go version -m` 或 `go list -deps`），且 `go list -deps ./sdk` 不含 modernc.org/sqlite 与 gorm.io/gorm。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：两个入口的复用、共用连接与停止时关闭；准备失败时关闭连接；准备期之外 Open 返回 FAILED_PRECONDITION；backend NONE；文件缺失时非首装不创建；ID 非递增；一个包写的账本另一个包能读出相同的 ID；host 的 migration 一致性校验；调用方给出的非空 sqlite_path 原样下发。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./sdk/... ./internal/core/...；go build ./...；`go list -deps` 列出的 Core 可执行文件与 SDK 根包模块都在 docs/development.md 的允许清单中。

不做：SQLite 候选副本与 Backup API（T-08.6）；维护许可的管理入口、external 插件的候选门槛与 Journal（T-09）。
```

### goal-53 · T-08.8 用户可见文案目录

状态：待派发。前置 T-04.1、T-04.3、T-04.6（PR #78，合并提交 `0a320c5`）、T-08.1、T-08.10（PR #77）均已合并。2026-11-04 按 D-69～D-74 改写，符号已按合入代码核对；同日按 D-76 修订 config 根节点的格式（与 config_schema.json 同形），与 schema 的核对归 goal-95；按 D-80 改用 `i18n/` 目录并允许顶层的 `metadata`（D-78）；按 D-82 修订：默认 locale 写在 `core_config.json` 的 `default_locale`，由装配处从 goal-97 的 settings 取得，本切片不建 `core_setting` 表：
- 命令 usage：`router/command.go` 的 `commandUsage(spec *pb.CommandSpec, prefix string) *pb.MessageChain` 以英文 "Usage: " 加本次匹配到的前缀、规范路径、别名与参数标注拼出缺省 usage；USAGE_MODE_CUSTOM 时也按缺省 usage 发送，函数开头有一条 `shortcut:` 注释，说明待 T-08.8 查找 usage_key。usage 经注入的发送回调（message.Service.Send）发出。
- `CommandSpec.usage_key = 12` 是插件 i18n/ 中 messages 的键；`sdk.UsageKey(key) RouteOption` 写入 USAGE_MODE_CUSTOM 与 usage_key；`Usage(text)` 已删除。`sdk.Collect(root) (*Registration, error)` 在启动时合并生成项与注册。
- `sdk.Generated{ConfigSchema, Commands, Events}`，由生成代码在 init 中经 `sdk.UseGenerated` 登记。`sdk/generate` 只做静态扫描，看不到运行时注册的 UsageKey 参数。
- Bootstrap 已用到 11（session_limits），12 空闲。
- `router.Plugin` 接口有 PluginID、Activated、Routes、Call、SendEvent，`router.FromInstance(*host.Instance)` 是唯一的生产实现；`supervisor/bus_test.go` 的 `busWriteGate` 嵌入 `router.Plugin`。`host.ActivateOptions` 与 `host.Instance` 由 supervisor 在 Launch 中构造；`supervisor.Spec.ConfigSchema` 由调用方从当前安装包取得，按同样方式传入 i18 资源。
- 字段文案的 i18 格式按 D-76 改为与 config_schema.json 同形（CONTRACTS §5.3），由 goal-95 改写生成器；T-08.10 实现的按 JSON 字段名嵌套、覆盖 label、description、labels、subtitles 的格式不再使用。

核对时补上原语句缺的一处：CONTRACTS §2.2 与 §5.4 要求自定义 usage 按插件 i18n/ 的键与调用者语言发送，但没有写 Core 从哪里取插件的文案。这里定为仿照 ConfigSchema：调用方经 `supervisor.Spec.I18n` 传入当前包的 i18n/ 原文（安装包与库模式的嵌入副本都有，CONTRACTS §9、§15），Launch 解析后放进 Instance，Router 经 `Plugin.Messages()` 读取。不加 proto 字段，也不改 `router.New`，避免与 goal-13 同时修改它。未知的 usage_key 由 SDK 在 Collect 时报错：生成器看不到运行时注册的参数，TEST-MATRIX V-87「构建检查」的部分由 Collect 与插件测试承担。

```text
/goal 使用 $alonabot-execute 推进 goal-53（切片 T-08.8：用户可见文案目录）。goal 目录 goal-53，分支 goal-53，worktree .worktrees/goal-53，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：
- 插件 i18n/<locale>.json 中 messages 的键与占位符在构建期检查，并由生成器嵌入。
- SDK 提供 Context.T，按「Actor locale → Core 默认 locale → en-US」选择语言，支持主语言回落。
- Core 内置 en-US 与 zh-CN 的用户文案目录，缺省命令 usage 改为经它生成；自定义 usage 按插件 i18n/ 的 usage_key 与同一规则选择语言。
- Core 默认 locale 是一项持久化的全局设置，经 Bootstrap 下发给 SDK。

依据：PLAN T-08 步骤 8；CONTRACTS §5.4、§5.3（config 根节点）、§2.2 第 8 条；MESSAGE-MODEL §1（Actor.locale）；DECISIONS D-72、D-76；TEST-MATRIX V-54 的 SDK 与 Core 分项、V-87 的 usage 语言分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：Bootstrap 追加 string default_locale = 12，随后运行 go generate ./...（本切片获授权只追加这一字段）
- contracts/wirespec_test.go：追加编号检查
- 新增 contracts/i18n/：文件格式解析、占位符、语言选择，Core 与 SDK 共用
- sdk/generate/：检查并嵌入 i18n/*.json；sdk/generated.go 的 Generated 增加 I18n
- sdk/：新增 i18n.go（Context.T 与后台 T）；Collect 检查 usage_key；session.go 读取 Bootstrap.default_locale（只加读取）
- 新增 internal/core/i18n/：Core 内置目录与 en-US.json、zh-CN.json
- internal/core/host/：ActivateOptions 与 Instance 增加 Messages；Bootstrap 写入 default_locale 的调用点
- internal/core/supervisor/supervisor.go：Spec 增加 I18n 与 DefaultLocale，Launch 解析 I18n 并传给 Activate（只加这一段）
- internal/core/router/：Plugin 接口与 instancePlugin 增加 Messages；commandUsage 改用目录，删除其中的 shortcut 注释
- 上述包的测试文件
允许的新跨模块边：无（contracts/i18n 在 contracts 内；sdk/... -> contracts、internal/core/... -> contracts 已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. contracts/i18n
   type Catalog struct{ ... } // 一个组件全部 locale 的 messages
   type Problem struct { File string; Key string; Message string }
   func Parse(files map[string][]byte) (*Catalog, []Problem) // 键为文件名（如 "en-US.json"）；有问题时 Catalog 为 nil；files 为空时返回 nil, nil
   func (c *Catalog) Has(key string) bool // en-US 中是否有该键
   func (c *Catalog) Format(locales []string, key string, args map[string]string) (string, error) // locales 为按优先级排列的候选（如 [Actor locale, 默认 locale]），最后总会回落到 en-US
   func Placeholders(message string) ([]string, error)
2. sdk
   func (c *Context) T(key string, args ...any) (string, error) // args 为成对的名字与值，值用 fmt.Sprint 格式化
   func T(ctx context.Context, key string, args ...any) (string, error) // 没有 Actor：从 Core 默认 locale 开始
   Generated 增加 I18n map[string][]byte // 文件名 → 原文
3. internal/core/i18n
   func Messages() *i18n.Catalog // Core 内置目录（嵌入的 en-US.json 与 zh-CN.json）
4. internal/core/host
   ActivateOptions 增加 Messages *i18n.Catalog
   Instance 增加 Messages *i18n.Catalog // 当前包的用户可见文案；没有 i18 资源时为 nil，激活后只读
5. internal/core/supervisor
   Spec 增加 I18n map[string][]byte // 当前包的 i18n/*.json 原文，来自安装包或库模式的嵌入副本；nil 表示没有文案资源
   Spec 增加 DefaultLocale string // 写入 Bootstrap.default_locale；装配处由 core_config.json 的 default_locale 取得（goal-97），空值按 "en-US"
6. internal/core/router
   Plugin 接口增加 Messages() *i18n.Catalog；instancePlugin 返回 instance.Messages

行为：
- 文件格式（Parse 校验，问题以 Problem 报告）：
  - 文件名必须是规范 BCP 47 标签加 .json；JSON 不合法是错误；顶层只允许 metadata、config 与 messages；metadata 只允许 name 与 description 两个字符串（D-78）。
  - messages 的键为点分隔的 ASCII 标识（每段为 [a-z0-9_]+），值为字符串。{name} 为占位符（name 遵循同样的标识规则），{{ 与 }} 为字面花括号，其余单独的花括号是错误。
  - 必须有 en-US.json；其他语言不得有 en-US 没有的键；同一键在各语言中的占位符集合必须与 en-US 相同。
  - config 根节点（D-76）与 config_schema.json 同形：顶层与 items 下是「字段键 → 字段覆盖」，字段覆盖只允许 description、hint（字符串）、labels（字符串数组）与嵌套的 items、templates；templates 下是「模板键 → 模板覆盖」，模板覆盖只允许 name、hint（字符串）与 items。只做结构检查，不做占位符检查，也不检查字段键与模板键是否存在于 schema（这项核对属于 goal-95）。
- 语言选择（Format）：按候选顺序逐级尝试，每一级先精确匹配，再按主语言子标签匹配（同一主语言有多个文件时取字典序最小者）；都不命中时用 en-US。目标语言缺少该键时回落到 en-US 的该键。en-US 也没有的键返回错误；缺少占位符参数返回错误；多余参数忽略。不用键名充当文案。
- 生成器：构建入口目录下有 i18n/ 时解析其中全部 .json，Problem 作为诊断报告（位置为文件名与键）；没有问题时把文件原文嵌入 Generated.I18n。alonabot check 报告同样的诊断。
- SDK：
  - Context.T 的候选为 [Actor.locale（存在时）, Bootstrap.default_locale]；后台 T 为 [default_locale]。插件没有 i18 资源时，T 返回 NOT_FOUND 的错误。
  - Collect：USAGE_MODE_CUSTOM 的路由，其 usage_key 不在 en-US messages 中（或插件没有 i18 资源）时返回错误。
- Core：
  - Launch 用 contracts/i18n.Parse 解析 Spec.I18n，有 Problem 时在启动进程前返回 INVALID_ARGUMENT 的错误；Bootstrap.default_locale 取 Spec.DefaultLocale。修改默认 locale 后，新值在之后建立的 Session 生效（已运行的插件在下次启动时看到新值）。
  - 缺省 usage：文案键为 core.usage.*（en-US 与 zh-CN 两份），占位符为本次匹配到的前缀、命令路径、别名与参数用法（参数类型标注 :int、:float、:bool、:user、:image、:message 保持不变）。语言按 [事件 Actor 的 locale, 默认 locale] 选择；Router 取默认 locale 的方式由执行者在 Router 内部自定，不改 router.New。
  - 自定义 usage（USAGE_MODE_CUSTOM）：用 Plugin.Messages() 的 usage_key 按同一候选格式化，参数为 prefix（本次匹配到的前缀）与 command（前缀加规范路径），多余参数照常忽略。Messages 为 nil、缺少该键或格式化失败时，记 WARN（plugin_id、route_id、错误），改发缺省 usage。
- 日志、错误字符串与 safe_message 仍只用英文。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/i18n_test.go（//go:build unix），测试插件带 en-US、zh-CN、zh-TW 三个文件，收到消息时用 Context.T 回复：
    - Actor.locale 为 zh-CN 时得到中文；zh-HK 回落到主语言 zh 中字典序最小的 zh-CN；ja 与没有 Actor 时用默认 locale。
    - 把默认 locale 设为 zh-CN 并重启插件后，ja 得到中文。
    - 参数错误时的缺省 usage：zh-CN 用户收到中文，en-US 用户收到英文。
    - 带 UsageKey 的命令参数错误时，按调用者语言发送插件的自定义 usage。
    - Spec.I18n 有问题（如 zh-CN 多出键）时 Launch 返回 INVALID_ARGUMENT，且没有启动进程。
  - sdk/generate/i18n_test.go：沿用 generate_test.go 的临时 module 辅助函数。缺少 en-US、zh-CN 多出的键、占位符集合不同、非法 JSON、多余的顶层键、config 根节点中出现不允许的键或类型不对时各产生诊断；合法时 Generated.I18n 被嵌入。
  - contracts/wirespec_test.go：Bootstrap.default_locale = 12。
- 已有集成测试中断言英文缺省 usage 原文的地方（如 command_test.go）按 en-US 目录的文案同步，覆盖的场景不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 占位符与花括号转义；Format 的未知键与缺少参数。
  - BCP 47 规范化。
  - Core 目录的 en-US 与 zh-CN 键集合一致。
  - Collect 拒绝未知的 usage_key；Messages 为 nil 时回落到缺省 usage 并记 WARN。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：WebUI 文案、语言切换与配置字段翻译的渲染（T-08.7）；确认消息的文案（T-11.2 使用本目录）；修改默认 locale 的管理 API（T-09.14）；从安装包与库模式读取 i18n/ 并填入 Spec.I18n 与 Spec.DefaultLocale 的装配（T-09.5、T-13.1）；Adapter 权限的 label_key 显示（T-09.9）。
```

### goal-100 · T-09.16 manifest 拆分为作者信息与构建信息

状态：等待前置（T-03.8）。依据 D-78 新增。goal-8（T-03.8）会给 `contracts/manifest` 增加 role、adapter_id、icon、assets、Admin API 版本范围与依赖字段，本切片在它合并后拆分。语句只写了目标、依据和要点，前置合并后由审查者按实际的 manifest 包写出接口与测试段。

目标：组件工程中的 plugin.yaml 只写作者信息（id、name、version、description、author、source、role，以及按角色需要的 adapter_id、icon、依赖、WebUI 的 assets 与 Admin API 版本范围，写外部数据时的 prepare_effects: external）；构建信息（manifest_version、协议主版本范围、required_capabilities、os/arch、entrypoint、config_version、migration_mode、prepare_effects 的 none 或 managed_db）由构建写入包内 manifest。包内 manifest 照常严格校验，并增加 name、description 与 author。

依据：PLAN T-09 步骤 2；CONTRACTS §9.1、§5.4（metadata）；DECISIONS D-78；TEST-MATRIX V-01。

要点：
- contracts/manifest 提供两种解析：工程 manifest（出现构建字段时报错，指明字段）与包内 manifest（现有的严格规则，加上 name 等新字段）；以及由工程 manifest 加构建信息生成包内 manifest 的函数，输出稳定（同样的输入得到逐字节相同的 YAML）。
- packages 的安装预检只接受包内 manifest；`name` 缺失的包被拒绝。
- 构建信息的取值来源（SDK 的协议版本与能力、目标平台、迁移接口、是否链接受管 SQLite、config_schema.json 的 $version）由调用方传入，本切片不实现 build 与 pack 命令（T-09.3）。
- 提交的集成测试覆盖：工程 manifest 写了 entrypoint 时报错；生成的包内 manifest 通过预检；缺 name 的包内 manifest 被拒绝。

## 第 9 波

### goal-91 · T-04.8 SDK 自有类型与事件常量

状态：等待前置（T-04.7）。依据 D-70，2026-11-04 从 T-04.7 拆出。语句只写了目标、依据和要点，T-04.7 合并后由审查者按合入代码写出接口、行为与测试段。

目标：
- 插件与 Adapter 的公共 API 不再暴露生成的 pb 类型：SDK 自己定义 Message（入站消息与片段）、Target、Actor、SendResult 与错误码，pb 只留在 SDK 内部。
- 新增标准事件常量（带 payload 的 Go 类型）与 sdk.AdapterEvent；事件注册改用常量，Collect 核对事件 Handler 的 payload 参数类型。
- 命令参数类型 *pb.ImageSegment、*pb.MessageChain 改为 SDK 自有类型。
- sdk/adapter 的 Adapter、Instance、Core 接口改用 SDK 类型，官方 Telegram Adapter 与已合并的全部调用方同步。

依据：PLAN T-04 步骤 9（自有类型部分）；CONTRACTS §2.1、§8；MESSAGE-MODEL §1～§3；DESIGN §D；DECISIONS D-70、D-71；TEST-MATRIX V-14 的类型分项、V-86 的 SDK 分项。

要点：
- 涉及 Context.Event、Context.Send 与 sdk.Send 的返回值、Upload 与 GrantMedia 的 Target 与 MediaRef、HTTP 的 Public(mode)、sdk.Error.Code，以及 sdk/message 的入站表示。
- 事件常量放在 sdk 中（标准事件），Adapter 的常量包（例如 adapters/onebot11/event、adapters/onebot11/perm）只含常量与数据类型，由 T-03.11、T-15.1 提供；扩展事件的 Handler 以 sdk.AdapterEvent 声明 payload 类型，SDK 按 adapter_event 的 payload_json 解码。
- 不改协议：proto 与 Core 侧不变，SDK 在边界内转换。

### goal-96 · T-04.9 删除构建期标记

状态：等待前置（T-04.7）。依据 D-79 新增。goal-88（T-04.7）正在把注册改为链式，并让生成器为标记输出链式调用；本切片在它合并后删除标记本身。语句只写了目标、依据和要点，前置合并后由审查者按实际代码写出接口与测试段。

目标：删除 `//alonabot:command`、`//alonabot:event` 构建期标记。命令、文本匹配与事件只在根插件的注册方法中链式注册，包级函数也这样注册；生成器不再扫描注册，只检查并嵌入 manifest、配置 schema、i18n 与 Service Schema。

依据：PLAN T-04 步骤 2；CONTRACTS §9.1；DESIGN §D；DECISIONS D-70、D-79；TEST-MATRIX V-10。

要点：
- sdk/generate 删除 marker.go 与标记的扫描、诊断和生成的注册代码；`sdk.Generated` 中只为标记服务的字段与 `UseGenerated` 中对应的登记一并删除，不留兼容读取（编码规范「首次发布前不写兼容层」）。
- 仓库中使用标记的示例、测试插件与 testdata 改为在注册方法中链式注册，行为不变。
- 残留的标记注释只是普通注释，不产生注册，也不报错。
- 提交的集成测试覆盖：包级函数经链式注册后照常派发；残留标记不再产生路由；生成两次的输出逐字节相同。

### goal-15 · T-05.1 可信上下文与有界分发

状态：待派发。前置 T-03.3、T-04.3、T-04.4（PR #88，合并提交 `c8dd516`）均已合并。本 goal 此前已派发：草稿 PR #67（分支 goal-15）只有两个 fixture，内容与语句一致，继续时在该分支先 `git merge origin/main`，不另开分支。符号已按合入代码核对：
- `router.New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(*pb.Event, Caller) (*pb.Event, error), prefixes func(adapterInstanceID string) []string, policy CommandPolicy) *Router`。调用点只在 `internal/core/adapter/remote/remote_test.go` 与 `internal/core/supervisor` 的 bus、command、conversation、echo、index、media_in、policy、service 测试中；`cmd/alonabot` 还没有装配 Router。
- `Router.Start()` 启动唯一的 worker；`Router.Close()` 取消 `r.ctx` 并等待 worker，没有参数和返回值，调用点同上。`HasInFlight` 在 `r.ctx` 取消后返回 false。
- `dispatch` step 1 在链开始时计算 `actor.Permissions`：{user} ∪（来源权限中除 user、admin 以外的名字），经 `CommandPolicy.IsAdmin` 加入 admin。`call` 的顺序是 `prepare` → `bind` → 构造 HandlerRequest，`Invocation{PluginID, Event: prepared, DefaultTarget: ev.GetTarget()}`。
- `CommandPolicy` 有 `CheckCommand` 与 `IsAdmin`；Actor 为 nil 时 `IsAdmin` 返回 false，非空权限集合不满足。命令层只处理 `message.created`，`scheduled` 不进入命令层。
- `wire.Budget` 的 `acquire`、`release` 是包私有方法，调用方只有 `internal/wire/session.go`；nil 表示不限额。
- `tune.Values` 已有 `DispatchConcurrency`、`DispatchQueueEvents`、`DispatchQueueBytes` 与 `CoreMaxPayload`。
- `message.proto` 注明 PreEventResult 只能改写 `event_payload`，不能改写 Event 的其余字段。
- SDK 的事件路由目前不写 `required_permissions`；`command_test.go` 在 Attach 前直接修改路由描述来声明权限。T-04.7（#89）未合并，它会重写注册 API 并迁移测试插件：本切片的测试插件用合并时 main 上的 API，后合并者同步。

核对时修订了原语句：
- Options 补上 goal-7 的 Prepare 与 goal-87 的 Prefixes，保留 nil 语义。
- `trusted_roles`、`TrustedRoles`、USER/ADMIN 改为 D-69 的权限名集合。权限沿用链开始时计算，不提前到 Deliver：事件可能排队，策略修改在下一次派发时生效。
- 导出 `wire.Budget` 的预留与释放，范围加入 `internal/wire`。
- `Close` 写明签名、排队事件与期限到期后的处理。
- 保留字段按 `message.proto` 写全：除 `event_payload` 外都保留，原语句漏了 `occurred_at_unix_ms`、`capabilities`、`platform_payload` 与 payload 内的平台消息 ID。
- 写明重建与 `prepare` 的先后。
- 测试分开覆盖条数、字节与 Global 三种限额（原语句用一组 Options 无法区分）；scheduled 改为验证事件路由的权限，并写明它没有默认 Target。

```text
/goal 使用 $alonabot-execute 推进 goal-15（切片 T-05.1：可信上下文与有界分发）。goal 目录 goal-15，分支 goal-15，worktree .worktrees/goal-15，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Router 把 Adapter 输入转为 Core 保留的可信信息（Actor、Target、来源消息 ID、路由绑定）与可修改的业务 payload 两部分，插件看到的 Event 由 Core 保留的部分重建；用有界的并发派发替换单 worker：最多 DISPATCH_CONCURRENCY 条链并发、等待队列受 DISPATCH_QUEUE_EVENTS 与 DISPATCH_QUEUE_BYTES 限制并计入 Core 总预算，同一事件的候选在一条链内顺序执行，链与 Session 读循环分离。

依据：PLAN T-05 步骤 1、2；CONTRACTS §2.1、§2.3 第 1 条、§1.2；MESSAGE-MODEL §1、§1.1；TEST-MATRIX V-13、V-16 的 fixture 分项。

范围（只新增或修改这些）：
- internal/core/router/：新增 dispatch.go（并发链与等待队列）、trusted.go（可信信息的保留与重建）；router.go 的 New、Start、Close、Deliver、dispatch、call 改用它们，删除 Start 中的单 worker shortcut 注释
- internal/wire/：budget.go 导出预留与释放，session.go 同步调用点
- 更新 router.New 与 Router.Close 的全部调用者（remote_test.go 与 internal/core/supervisor 的测试；cmd/alonabot 若已装配 Router 也一并更新）。依赖事件先后顺序的现有测试传 Concurrency: 1
- contracts/testdata/events/：button.clicked 与 scheduled 的规范 fixture（JSON 形式的 pb.Event，PR #67 已提交）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/router
   type Options struct {
       Concurrency int          // DISPATCH_CONCURRENCY
       QueueEvents int          // DISPATCH_QUEUE_EVENTS
       QueueBytes  int64        // DISPATCH_QUEUE_BYTES
       Global      *wire.Budget // Core 全局 payload 预算，排队中的事件占用它；nil 时不限额
       Adapters    *adapter.Registry // 账号用户名查询
       Send        ServiceFunc // usage 与自动回复发送
       Prepare     func(*pb.Event, Caller) (*pb.Event, error) // 为接收 Session 改写事件；nil 时原样派发
       Prefixes    func(adapterInstanceID string) []string // 实例的命令前缀；nil 时所有实例使用 ["/"]
       Policy      CommandPolicy // 动态命令策略与管理员身份；nil 时仅按声明权限判定且不赋予 admin
   }
   func New(opt Options) *Router // 取代位置参数的 New，不留旧签名
   type Stats struct { Accepted, DroppedQueueFull, DroppedBudget uint64; Running, Queued int }
   func (r *Router) Stats() Stats
   func (r *Router) Close(ctx context.Context) error // 取代无参数的 Close，不留旧签名
internal/wire
   func (b *Budget) Acquire(n int64) bool // 由 acquire 改名，不等待，超过上限时不占用并返回 false；接受 nil
   func (b *Budget) Release(n int64)      // 由 release 改名；接受 nil

行为：
- Deliver：所有事件先入队，按进入顺序开始执行。事件编码后的大小（proto.Size）计入 QueueBytes 与 Global。等待中的事件数已达 QueueEvents 或字节将超过 QueueBytes 时计入 DroppedQueueFull，Global 无法预留时计入 DroppedBudget，都立即返回 adapter.ErrQueueFull，不阻塞调用方。事件开始执行时释放它在队列与 Global 中的占用。Start 之前接纳的事件留在队列中；Close 之后 Deliver 照旧返回错误。
- 并发：最多 Concurrency 个事件同时执行；同一事件的全部候选（命令层与事件路由）在一条链内按 (priority, plugin_id, route_id) 顺序执行，不同事件互不等待。链执行在独立 goroutine 中，不在任何 Session 的读循环里运行；慢 Handler 不影响其他 Session 的 Response、Cancel 与控制帧。
- 可信信息：Deliver 时把 Event 拆成 Core 保留部分与业务副本。保留部分是 event_payload 以外的全部字段（event_id、event_type、source_kind、occurred_at_unix_ms、adapter_id、adapter_instance_id、account_id、target、actor、message、capabilities、platform_payload、source_event_id），以及 event_payload 中的路由与来源绑定：button_clicked 的全部字段；scheduled 的全部字段；message_created、message_edited、reaction_added 的 platform_message_id；message_deleted 的 platform_message_ids。业务副本是其余的 event_payload。每次调用 Handler 时先用保留部分和当前业务副本（后续 PreEvent 可改写的副本）重建事件，再交给 Prepare 与参数绑定；Prepare 只替换入站媒体 ID。InvocationContext 的 actor、default_target 与 Invocation.DefaultTarget 只取保留部分。T-05.3 的 PreEvent 只能改写业务副本。
- 身份：Actor 的权限集合沿用 dispatch step 1 的计算，在链开始时算一次并写入保留部分，来源带的 user、admin 不生效；没有 Actor 的事件（例如 scheduled 与 bot.online）不补用户，只满足空的 required_permissions。
- Router.Close：停止接纳新事件；丢弃等待中的事件并释放它们在队列与 Global 中的占用；等待运行中的链结束。链全部结束时返回 nil。ctx 先到期时取消链内的调用，等链退出后返回 ctx.Err()。未 Start 或重复调用时同样安全。链结束之前 HasInFlight 仍按原规则返回。
- 本切片不改变命令层与事件路由的匹配规则，也不实现 PreEvent、Rule 与传播的新语义。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/dispatch_test.go（//go:build unix）：构建一个测试插件，其事件路由按收到文本睡眠指定时长后回复，可用不同 priority 以三个插件 ID 启动。用 memory Adapter 装配，每项按需要给出 Options，等 Stats().Running 达到预期后再投递下一个事件：
  - 三个插件以不同顺序启动时，同一事件的调用顺序都是 (priority, plugin_id, route_id)。
  - 条数：Options{Concurrency: 2, QueueEvents: 1}，QueueBytes 取 tune 缺省值。两个慢事件执行期间，第三个事件进入队列，第四个得到 ErrQueueFull，DroppedQueueFull 为 1；慢事件结束后排队事件被执行。
  - 字节：Options{Concurrency: 1, QueueEvents: 3}，QueueBytes 等于一个事件的 proto.Size。一个慢事件执行期间，第二个进入队列，第三个得到 ErrQueueFull，DroppedQueueFull 为 1。
  - 全局预算：QueueBytes 取 tune 缺省值，Global 为 wire.NewBudget(一个事件的 proto.Size)。一个慢事件执行期间，第二个进入队列，第三个得到 ErrQueueFull，DroppedBudget 为 1；队列清空后 Global.Used() 为 0。
  - 慢 Handler 执行期间，同一插件的 Session 仍能完成另一条 core.send 的 Response（读循环未被占住）。
  - 读入 contracts/testdata/events 的 button.clicked 与 scheduled fixture 并 Deliver：插件收到的 action_id/value、job_id/run_id 与 fixture 一致。scheduled 没有默认 Target，测试插件把收到的字段写入测试给出的文件，不用回复上报。scheduled 没有 Actor：Attach 前把一条 scheduled 事件路由的 RequiredPermissions 改为 {"user"}（沿用 command_test.go 的做法），它不被调用，同一事件上权限为空的路由被调用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：QueueBytes 与 Global 的预留与释放在并发下正确（-race）；Close 时排队事件的占用被释放；Close 在 ctx 到期后取消运行中的链并返回 ctx.Err()；事件自带的 admin 权限不生效；可信字段在重建时覆盖业务副本中的同名值。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./internal/wire/... ./sdk/...；go build ./...。

不做：PreEvent 与出站 Hook（T-05.3）；Rule 与 HandlerOutcome 新语义（T-05.2）；延续调用（T-05.4）；实际的 scheduled tick（T-07.2）；按钮 token（T-03.8）。
```

### goal-98 · T-07.7 插件页面的沙箱隔离

状态：等待前置（T-07.6）。依据 D-84 新增。T-07.4 已实现两个主机名、插件兑换票据与插件会话（`internal/core/webhost` 的 host.go、policy.go、ticket.go），本切片把它改为单一主机名。语句只写了目标、依据和要点，前置合并后由审查者按 goal-97 合并后的 settings 与 webhost 写出接口与测试段。

目标：Core 只用一个主机名：删除 plugin_base_url、插件兑换票据与插件会话，core_config.json 的 admin_base_url 改名为 base_url。插件路由的全部响应加上不含 allow-same-origin 的 CSP sandbox 与 nosniff，并去掉插件设置的 Set-Cookie；非公开路由要求管理会话，非 GET 请求验证 CSRF。

依据：CONTRACTS §7.2～§7.5、§15.1；DESIGN §H；DECISIONS D-15、D-84；TEST-MATRIX V-47 的组件分项。

要点：
- 响应头 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff` 由 Core 在代理响应上设置，插件设置的同名头被覆盖，`Set-Cookie` 全部去掉。
- 非公开路由的授权使用 adminauth 的管理会话；非 GET 请求要求与 Admin API 相同的 CSRF 头。公开路由（OAuth、Webhook）不要求会话。
- 删除 webhost 中的插件 host 分支、ticket.go 与插件 Cookie，不留兼容读取；settings 与 main.go 同步改为 base_url。
- 提交的集成测试覆盖：插件返回 HTML 时带 sandbox CSP 与 nosniff；插件设置的 Cookie 不到达浏览器；未登录访问非公开路由被拒，登录后 GET 成功、缺 CSRF 的 POST 被拒；公开路由无需会话；管理 Cookie 与 Authorization 不转发给插件。

### goal-33 · T-08.6 SQLite 候选与资源交接

状态：待派发。前置 T-08.4（PR #74）与 T-08.5（PR #79，合并提交 `75c16d3`）均已合并。符号已按合入代码核对：
- `internal/oslock`：`TryLock(path string) (*Lock, error)`，文件不存在时以 0600 创建，已被占用时返回包装了 `oslock.ErrLocked` 的错误；`(*Lock).Unlock() error`。锁文件为 `<PluginDir>/runtime.lock`（`wire.RuntimeLockFile`）。
- `state.Store`：`LockPlugin(pluginID) (unlock func())`、`AddGeneration(ctx, pluginID, generationID, configSnapshot string, configVersion uint32)`、`Active(ctx, pluginID)`、`GenerationConfig(ctx, generationID)`、`SetGenerationConfig`、`GenerationSecretPaths`、`GetPendingConfig(ctx, pluginID) (PendingConfig, error)`（没有记录时为 `state.ErrNotFound`）。`schema.sql` 的 `generation` 表有 `plugin_id`、`config_snapshot`、`committed_at`；`generation_secret_path` 以外键引用 `generation`。状态库不存数据库路径，也没有列出某插件全部代次的方法。
- `secret.Store`：`secret.New(st)`、`Collect(ctx, pluginID string, protected [][]byte) (int, error)` 删除该插件中未被 protected 快照引用的版本。
- `supervisor`：`Spec` 已有 `PluginDir`、`GenerationID`、`Bootstrap`、`ConfigSchema`、`Secrets`、`KeptSecretPaths`、`NewCandidate`、`ExpectedOld`。`Launch` 在 backend 为 SQLITE 且 `sqlite_path` 为空时填写 `<PluginDir>/generations/<GenerationID>/db.sqlite`，非空时原样下发；新候选的配置快照与 secret 路径由 Launch 自己经 `Extract` 和 `SetGenerationConfig` 持久化，NewCandidate 时在 ACTIVATE 前调用 `CommitActive`。
- SDK（T-08.5）：非首装以 `mode=rw` 打开，文件缺失不创建；首装由 SDK 创建数据库文件。
- `modernc.org/sqlite` v1.61.0：驱动连接有 `NewBackup(dstUri string) (*sqlite.Backup, error)`，经 `(*sql.Conn).Raw` 取得；`Backup` 有 `Step(n int32) (more bool, err error)`（n 为负时一次复制全部页）与 `Finish() error`（同时关闭目标连接）。
- Core 的错误码沿用 `&validate.Error{Code: pb.ErrorCode_..., Message: ...}`（如 `internal/core/config`）。

核对时修订了原语句的四处：
- `Reason` 改为直接使用 `pb.ActivationReason`，它与 Bootstrap 的 `activation_reason` 是同一个值，不另定义枚举。
- `Prepared` 删除 `ConfigSnapshot`、`SecretPaths`、`ConfigDigest`、`Migration`：它们由 Launch 持久化或在 REGISTER 中返回，本切片没有调用方填写，交接方式改为写进包文档；新增 `PluginID`，`Discard` 需要它。
- `Discard` 改为接收 `*secret.Store`，并删除候选的未提交 `generation` 记录；否则保留的记录会引用已删除的 secret。保护范围在 T-09 引入 Journal 与保留代次之前定为：该插件除本候选外的全部代次快照与 Pending。
- 原「读取代次 DB 路径」的 state 方法不需要（路径按 CONTRACTS §6.1 的布局计算），改为上面两个 state 方法；测试范围补上 `internal/core/supervisor` 的测试文件。

```text
/goal 使用 $alonabot-execute 推进 goal-33（切片 T-08.6：SQLite 候选与资源交接）。goal 目录 goal-33，分支 goal-33，worktree .worktrees/goal-33，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在已持插件操作锁、旧进程已退出且 Core 能取得 runtime.lock 的前提下，用 SQLite Backup API 从 Active 代次的数据库创建独立的候选副本（首次安装只建目录，reload 与 update 复制，resume 复用 Active）；复制完成后释放 runtime.lock，由随后启动的 SDK 取锁打开候选。失败清理只删除本候选的未受保护资源。在包文档中写出供 T-09 使用的调用顺序。本包不写 Active，不写 COMMITTED。

依据：PLAN T-08 步骤 7；CONTRACTS §6.1、§6.3 第 1、2、5 条、§12.2、§12.3；TEST-MATRIX V-24、V-36 的候选分项。

范围（只新增或修改这些）：
- 新增 internal/core/candidate/
- internal/core/state/：新增 candidate.go，放下面两个方法（不改已有方法与 schema.sql）
- internal/core/supervisor/：新增 candidate_test.go；db_test.go 中 databaseSQLPlugin 的 OnStart 只追加一个环境变量控制的失败分支
- 上述包的测试文件
允许的新跨模块边：无（internal/core -> internal/oslock 已允许；Backup 使用已批准且已在 go.mod 中的 modernc.org/sqlite）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/candidate
   type Request struct {
       PluginID       string
       PluginDir      string                // <plugins>/<id>，绝对路径
       Reason         pb.ActivationReason   // FIRST_INSTALL、RELOAD、UPDATE、RESUME；其他值返回 INVALID_ARGUMENT
       FromGeneration string                // Active 代次；FIRST_INSTALL 时为空
       NewGeneration  string                // RESUME 时必须等于 FromGeneration
   }
   type Prepared struct {
       PluginID     string
       Generation   string
       DatabasePath string // <PluginDir>/generations/<Generation>/db.sqlite，绝对路径；调用方写入 Bootstrap 的 sqlite_path
       Created      bool   // 本次是否新建了候选数据库文件
   }
   func PrepareDatabase(ctx context.Context, req Request) (*Prepared, error)
   func Discard(ctx context.Context, st *state.Store, secrets *secret.Store, p *Prepared) error
2. internal/core/state（candidate.go）
   func (s *Store) PluginGenerationConfigs(ctx context.Context, pluginID string) (map[string]string, error) // generation_id → config_snapshot，含未提交代次
   func (s *Store) DeleteCandidateGeneration(ctx context.Context, pluginID, generationID string) error
   var ErrCommittedGeneration = errors.New("generation is committed or active")
   // 同一事务删除 generation_secret_path 与 generation 记录；代次已 COMMITTED 或是当前 Active 时返回包装了 ErrCommittedGeneration 的错误，不属于该插件时返回普通错误，二者都不删除；记录不存在时返回 nil

行为：
- 路径：目录 <PluginDir>/generations/<gen>/（0700），文件 db.sqlite，与 supervisor 的默认路径一致（CONTRACTS §6.1）。
- PrepareDatabase 先以 oslock.TryLock 取 <PluginDir>/runtime.lock。取不到（errors.Is(err, oslock.ErrLocked)）返回 FAILED_PRECONDITION，消息为 "plugin process still holds runtime.lock"，不打开任何数据库、不建目录。成功与失败都在返回前释放锁。
- FIRST_INSTALL：建候选目录，不建数据库文件（SDK 首装时创建），Created=false。
- RELOAD/UPDATE：
  - 以 `mode=rw` 打开 FromGeneration 的 db.sqlite。不用 create 模式，文件缺失时返回错误，不创建空库。
  - 建候选目录，用 Backup API 复制到同目录的临时文件（Step(-1) 后 Finish），再对临时文件执行 PRAGMA integrity_check，结果必须是 "ok"。
  - fsync 临时文件，重命名为 db.sqlite，再 fsync 目录。Created=true。
  - 不能只拷贝主文件：Active 处于 WAL 且有未检查点的写入时，副本也要包含这些写入。
  - 候选目录已存在时返回 FAILED_PRECONDITION，不覆盖。
- RESUME：返回 Active 代次的路径，不复制、不建目录，Created=false。
- 失败（打开、空间、备份、校验、fsync 或重命名）：删除临时文件和本次新建的目录后返回原错误，不触碰 FromGeneration 的任何文件。
- Discard（调用方持有插件锁）：
  1. Generation 等于当前 Active 时返回 FAILED_PRECONDITION，不删除任何东西。
  2. 先 DeleteCandidateGeneration；返回 ErrCommittedGeneration 时转为 FAILED_PRECONDITION，不删除目录。
  3. 再删除候选目录。
  4. 最后调用 Collect，protected 为 PluginGenerationConfigs 中其余全部代次的快照，加上 Pending 的 config_json（没有 Pending 时只用代次快照）。
  5. 任一步失败时，原样返回该步骤的错误。
- 包文档写明调用顺序，供 T-09.5 引用：
  1) 持插件锁，冻结并停止旧进程，确认进程树退出；
  2) PrepareDatabase；
  3) AddGeneration（未提交）；
  4) Launch：NewCandidate=true，ExpectedOld=FromGeneration，Bootstrap.database_descriptor 的 activation_reason 等于 Reason，sqlite_path 为 Prepared.DatabasePath；
  5) PREPARED 后由 T-09.6 的唯一提交点在同一事务中提交 Active 与 COMMITTED；
  6) 提交前失败时 Discard，再以旧代次启动一个新 Session（NewCandidate=false）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/candidate_test.go（//go:build unix）。复用 db_test.go 的 buildDatabasePlugins、setup 与 databaseFile，SQL 插件的 OnStart 在写入 'start' 后按 DB_FAIL_START=1 返回错误：
  - 首装后 reload：PrepareDatabase(RELOAD) 产生独立文件；第二个候选的 OnStart 写入后失败；Discard 后候选目录与 generation 记录消失，Active 的行数与配置快照不变；以旧代次 restart 正常，看到的仍是原来的行。
  - 首装插件仍在运行（持有 runtime.lock）时 PrepareDatabase(RELOAD) 返回 FAILED_PRECONDITION，候选目录不存在。
  - 测试自己打开 Active 文件的一个连接（WAL，wal_autocheckpoint=0）写入几行，保持连接打开、不做检查点，此时 PrepareDatabase(RELOAD) 的副本能读到这些写入。
  - Discard 的 secret 保护：真实状态库中 Active 快照、Pending 和候选快照各引用不同的 secret，候选与 Active 还共享一个；Discard 后只删除仅被候选引用的那个。Discard 拒绝 Active 代次。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 临时文件的 fsync 与重命名。
  - integrity_check 不为 ok 时，临时文件和候选目录都被删除。
  - RESUME 不复制；FIRST_INSTALL 不创建数据库文件。
  - 源文件缺失时报错且不创建源文件；候选目录已存在时拒绝。
  - Reason 为 ROLLBACK 或 UNSPECIFIED 时返回 INVALID_ARGUMENT。
  - DeleteCandidateGeneration 拒绝已提交和属于其他插件的代次。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Journal、唯一提交点与恢复（T-09.5～T-09.7）；external 插件的候选门槛（T-09.5）；保留回退代次与 rollback（T-09.7）；把 Launch 内的提交移出（T-09.6）。
```

### goal-95 · T-08.11 配置改由 config_schema.json 声明

状态：等待前置（T-04.7）。依据 D-76 新增，同日按 D-85 补充 `$version` 与 `$migrations` 的声明式迁移，按 D-80 改用 `i18n/`；T-08.3、T-08.4、T-08.10 已合并。goal-88（T-04.7）正在重写 `sdk/generate` 与注册 API，并为 JSON Service 生成 Schema，用的是现有配置 Schema 的 go/types 类型推导；本切片在它合并后的代码上删除配置专用部分，保留这项推导。语句中涉及生成器内部结构的部分只写职责，goal-88 合并后由审查者按实际代码核对，再改为待派发。

已核对现有实现（main `1c07284`）：
- `sdk.Root` 有 Hook `Config() config.Config`，`sdk.Base` 默认返回 nil；`sdk/config` 的 `Config` 接口只有 `OnLoad`，`config.Base` 提供原样返回 raw 的默认实现。
- `sdk/session.go` 的 `prepare` 调用 `config.Prepare(ctx, root.Config(), generatedFor(root).ConfigSchema, raw, boot.PreviousConfigVersion, boot.GetTargetConfigVersion(), restart)`，失败一律是 `*validate.Error`；`buildRegister` 用 `config.SchemaHash` 计算 `config_schema_hash`。
- `sdk/config/prepare.go` 用 jsonschema-go 补默认值与校验，再按 Go 类型逐层检查；`sdk/config/schema.go` 定义 `KeywordSecret`、`KeywordHidden`、`KeywordForm` 与 `MarshalSchema`、`SchemaHash`。
- `sdk/generate/config.go` 的 `scan.configSchema` 经 go/types 从 Config 声明构造 JSON Schema 并解析 `config` 标签，结果写入 `Result.ConfigSchema` 与生成代码中的 `Generated.ConfigSchema`。
- `internal/core/secret` 的 `Extract` 沿 JSON Schema 的 `properties`、`items`、`additionalProperties` 查找 `x-alonabot-secret`（`path.go` 的 `schemaPaths`，`*` 表示任意下标或键）；`supervisor.Spec.ConfigSchema` 由调用方传入当前包的 Schema。Core 不按 Schema 校验配置。
- 使用旧写法的代码：`examples/plugins/showcase`；`internal/core/supervisor` 的 `config_prepare_test.go`、`secret_test.go`、`db_test.go`；`sdk/generate` 的 `config_test.go`、`generate_test.go`。

```text
/goal 使用 $alonabot-execute 推进 goal-95（切片 T-08.11：配置改由 config_schema.json 声明）。goal 目录 goal-95，分支 goal-95，worktree .worktrees/goal-95，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件配置改由与 plugin.yaml 同目录的手写 config_schema.json 声明，键名沿用 AstrBot 的 _conf_schema.json，另有 secret、required 与只能选一个模板的 template。`alonabot build` 与 `check` 检查并嵌入它；SDK 在准备阶段按 `$migrations` 迁移旧配置，再补缺失值并校验，插件经 sdk.Base 的 Config() 按类型取值；Core 按它抽取 secret。删除 config 标签、Go 配置类型与 config.Config、config.Base。

依据：CONTRACTS §1.1、§4.1、§5.1～§5.3、§9.1；DESIGN §D、§J；DECISIONS D-76；TEST-MATRIX V-25 的描述分项、V-37 与 V-38 的配置分项。

范围（只新增或修改这些）：
- 新增 contracts/configschema/：config_schema.json 的解析与检查、规范 JSON 与摘要、补缺失值与校验、secret 路径、i18n/ 中 config 覆盖与 schema 的核对；生成器、SDK 与 Core 共用
- sdk/config/：删除 Config、Base、Keyword*、MarshalSchema、SchemaHash 与基于 jsonschema-go 的准备；Prepare 改用 contracts/configschema 并返回取值对象；新增 Values 与 Template
- sdk/base.go、sdk/session.go、sdk/generated.go：Root 删除 Config Hook、增加 OnConfigLoad；Base 提供 OnConfigLoad 的默认实现与 Config() 取值方法；Generated.ConfigSchema 的注释
- sdk/generate/：删除 Config 类型发现与 config 标签解析，改为读取、检查并嵌入 config_schema.json，核对 i18n/ 的 config 覆盖；Service Schema 使用的 go/types 类型推导保留
- internal/core/secret/：secret 路径改由 contracts/configschema 给出
- 使用旧写法的代码与测试：examples/plugins/showcase、internal/core/supervisor 的测试插件、sdk/generate 的测试
- docs/plugin-guide.md §7；ARCHITECTURE.md 中 sdk/、contracts/ 与 internal/core/ 的实现状态
允许的新跨模块边：无（sdk/... 与 internal/core/... 到 contracts 已允许）。

接口：
1. contracts/configschema
   type Schema struct{ /* 不导出 */ }
   type Problem struct { Path, Message string } // Path 是 schema 中的路径，例如 music.templates.netease.items.cookie；Message 为英文
   func Parse(raw []byte) (*Schema, []Problem)           // 按 CONTRACTS §5.1 检查；有 Problem 时返回 nil
   func (s *Schema) Canonical() []byte                    // 保留键顺序、无多余空白的规范 JSON，数字保留原文
   func (s *Schema) Hash() []byte                         // Canonical 的 SHA-256，用作 config_schema_hash
   func (s *Schema) Version() int                         // $version，缺省 1
   func (s *Schema) Migrate(config map[string]any, from int) error // 按 CONTRACTS §5.5 依次执行 from+1 到 Version() 的步骤；from 大于 Version() 时返回 FAILED_PRECONDITION，写明两个版本号
   func (s *Schema) Complete(config map[string]any)       // 只补缺失键：default 或类型零值，递归进入 object、list 与 template_list 的元素、template 的模板对象
   func (s *Schema) Validate(config map[string]any) error // 失败为 *validate.Error（INVALID_ARGUMENT），消息带配置中的 JSON 路径、不含取值
   func (s *Schema) SecretPaths() [][]string              // 配置中的路径模式，"*" 表示任意数组下标；模板路径含模板键，例如 ["sources","*","webhook","token"]
   func (s *Schema) CheckI18n(config json.RawMessage) []Problem // i18n/<locale>.json 的 config 根节点与 schema 同形核对（CONTRACTS §5.3）
2. sdk/config
   type Values struct{ /* 不导出 */ }
   func (v Values) String(key string) string   // string 与 text
   func (v Values) Int(key string) int64
   func (v Values) Float(key string) float64
   func (v Values) Bool(key string) bool
   func (v Values) Strings(key string) []string // 没有 items 的 list
   func (v Values) Object(key string) Values
   func (v Values) Objects(key string) []Values // 有 items 的 list
   func (v Values) Dict(key string) map[string]json.RawMessage
   func (v Values) Template(key string) Template
   func (v Values) TemplateList(key string) []Template
   type Template struct{ Values }
   func (t Template) Key() string // 未选择模板时为 ""，此时取任何键都 panic
   type Loader func(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
   func Prepare(ctx context.Context, load Loader, schema []byte, raw []byte, fromVersion *uint32, targetVersion uint32, restart bool) (Prepared, error)
   Prepared 增加 Values Values
3. sdk
   Root 删除 Config() config.Config，增加 OnConfigLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)
   func (*Base) OnConfigLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error) // 原样返回 raw
   func (b *Base) Config() config.Values // 配置准备完成后可用，之前调用 panic
   Generated.ConfigSchema 是 config_schema.json 的规范 JSON；没有该文件时为 nil

要点：
- 生成器：构建入口目录有 config_schema.json 时用 Parse 检查，Problem 作为诊断报告（位置为文件名与 schema 路径），通过后把 Canonical 嵌入 Generated.ConfigSchema；没有该文件时不嵌入。读取 i18n/*.json 时（goal-53 合并后由它读取并嵌入）对每个文件的 config 根节点调用 CheckI18n。alonabot check 报告同样的诊断。
- 迁移：Parse 检查 `$migrations` 的版本连续、最后一步的 `to` 等于 `$version`、`rename` 与 `wrap` 的目标路径和 `map` 的新值在当前 schema 中存在且合法；四种操作的语义与路径写法按 CONTRACTS §5.5，路径不存在时跳过，`rename` 的目标已存在时失败。包内 manifest 的 config_version 取 `$version`（goal-100 接入构建）。
- 不保留旧写法：删除 config 标签解析与 Config 类型发现，不为残留的标签或 Config 方法加诊断（编码规范「首次发布前不写兼容层」）；旧写法因 config.Config、config.Base 不存在而编译失败。
- SDK 准备顺序：Migrate（从 Bootstrap 的 previous_config_version，首装时不迁移）→ OnConfigLoad → Complete → Validate → 规范 JSON 与 Values，错误码、安全消息与阶段预算沿用 goal-30；restart 时 OnConfigLoad 改变配置仍返回 FAILED_PRECONDITION。准备成功后 SDK 把 Values 交给根插件嵌入的 sdk.Base（根插件未嵌入 sdk.Base 时没有 Config()）。REGISTER 的 config_schema_hash 改用 Schema.Hash。
- 取值：键未在该层声明时 panic，消息为 `config key "<路径>" is not declared`；方法与字段类型不符时 panic，消息为 `config key "<路径>" is <类型>, not <方法对应的类型>`；路径以 "." 连接，数组为下标。secret 字段返回解析后的明文；Values 只读，可并发读取。
- Core：secret.Extract 收到的 schema 是 config_schema.json 的规范内容，路径模式来自 SecretPaths；空串不生成引用。Redact 与路径匹配的其余逻辑不变。
- showcase 改为 config_schema.json 与 OnConfigLoad，OnConfigLoad 中的迁移与取值检查保持原意。

测试（按编码规范「测试提交规则」）：
- 提交：
  - sdk/generate/config_test.go 改写：临时 module 带 config_schema.json（覆盖十种 type、options 与 labels、slider、condition、editor_mode、invisible、secret、required、template 与 template_list 的 display_item）与 i18n/zh-CN.json；两次生成逐字节相同，Generated.ConfigSchema 等于 Canonical，REGISTER 的 config_schema_hash 等于 Hash。诊断各一例：未知键、AstrBot 的 file 类型、`$migrations` 版本不连续、`rename` 的目标不在 schema 中、labels 多于 options、slider 缺 max、condition 引用不存在的字段、display_item 指向非 string 字段、secret 带 default、required 用于 int、字段键含 "."、i18 覆盖不存在的字段、i18 字段覆盖中出现 name。
  - internal/core/supervisor/config_prepare_test.go 的场景改用 config_schema.json：缺失键补 default 与零值；显式 false、0、"" 与空数组保留；null 被拒绝；options 之外的值；slider 越界；template 含两个模板键；required 的空串；从版本 1 经 `rename`、`map`、`wrap`、`remove` 迁移到版本 3 后再校验，删除的字段之外的未知键保留；配置版本高于 schema 时失败；OnConfigLoad 在声明式迁移之后执行；restart 时改变配置失败。测试插件在 OnStart 中用 Config() 取值并经日志或回复报告，核对取到的是补齐后的值。
  - internal/core/supervisor/secret_test.go：secret 位于 template 与 template_list 内时，快照只存引用、Bootstrap 收到明文；空串不生成引用。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Values 各方法对未声明键与类型不符的 panic 消息；Canonical 保留键顺序与数字原文；int64 边界与非有限数字；display_item 的嵌套路径；Merge Patch 删除键后重新准备得到默认值；web/admin/fixtures/ 中已有的 *.config_schema.json 都能通过 Parse（goal-34 未合并时跳过并写明）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./sdk/... ./internal/core/...；go build ./...。

不做：
- 表单渲染（T-08.7）；i18n/ 中 messages 与 config 根节点的结构检查（T-08.8）；Admin API 返回 schema 与 Pending 的提前校验（T-09.14）；Adapter 实例配置的 schema。
- AstrBot 的 file 类型、_special 与 dict 的 template_schema（D-76 不支持）。
```

### goal-39 · T-09.4 管理服务、安装记录与 serve

状态：等待前置（T-03.8、T-07.6、T-13.1）。语句中 goal-8、goal-31、goal-36、goal-37（packages）、goal-67 提供的符号（manifest.CheckDependencies、secret.Store、adminauth.Auth、packages.Preflight、根包的 assembly、remote.Process、policy 的 Admin 方法）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：`alonabot serve` 改为 `alonabot run`（D-77）；Adapter 实例与 Core Admin 的增删经 goal-97 的 settings 写回 `core_config.json`，实例凭证不再存为 Secret 引用（D-82）；首次 token 由 goal-97 实现；install 接受未签名的仓库来源并要求管理员确认（D-81）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-39（切片 T-09.4：管理服务、安装记录与 serve）。goal 目录 goal-39，分支 goal-39，worktree .worktrees/goal-39，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立唯一的管理服务，CLI 与（之后的）WebUI 都经它的 HTTP API 操作；四种角色的组件使用同一套 install、list、enable、disable、restart、uninstall（含 --purge）、pin 流程，所有变更按包串行；install 按角色执行同类唯一，`--source` 指定并持久化来源，`--replace` 显式替换同类组件；Adapter 实例的创建、启停与删除经此服务，凭证存为 Secret 引用；启用时检查必需依赖，停用、卸载被依赖的包时列出受影响者并要求显式确认；admin add/remove/list；alonabot run（不带子命令时相同）复用根包的 Core 装配，按安装记录启动全部已启用的包。reload、update、rollback 的入口与校验在此定义，候选编排由 T-09.5 起接入。

依据：PLAN T-09 步骤 3；CONTRACTS §9.1～§9.3、§8（依赖段）、§15.1、§15.2、§7.4、§5.3；DECISIONS D-26、D-42、D-59；TEST-MATRIX V-40、V-48、V-52、V-82、V-01 的安装分项。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移的 install 表定义中加入新列、末尾加入新表，安装记录的写入与读取方法同步新列；新增 internal/core/state/installs_admin.go（安装记录的扩展字段与 Adapter 实例表）
- 新增 internal/core/manage/（管理服务与 HTTP API）
- 根包 assembly.go：增加按安装记录启动全部已启用包、挂载管理 API 的步骤；新增导出的 Serve 供 cmd 使用
- cmd/alonabot/：serve（以及不带子命令时）调用根包 Serve；新增 install、list、enable、disable、restart、reload、update、rollback、uninstall、pin、adapter instance、admin 命令（经管理 API 访问）
- internal/core/buttons、conversation、policy、tool、secret：在卸载与 purge 中调用其已有的清理方法（只加调用点）
- 上述包的测试文件
允许的新跨模块边：无（cmd/alonabot -> 根包已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：install 表定义加入列 version TEXT NOT NULL, publisher_key_id TEXT NOT NULL, pin TEXT NOT NULL, uninstalled INTEGER NOT NULL（不设 DEFAULT，写入安装记录时给出全部列）；
   CREATE TABLE adapter_instance (instance_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES install(plugin_id), config_snapshot TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, created_at INTEGER NOT NULL);
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
   func Serve(ctx context.Context, dataDir string, args []string) error // cmd/alonabot run 使用；与 Run 共用 assembly

行为：
- install：packages.Preflight（验签、来源、平台、大小、摘要）→ MoveTo 不可变目录 → 写安装记录（install_seq 递增、source、publisher_key_id、version、role、adapter_id）。仍处于已安装状态的同 ID 返回 ALREADY_EXISTS，不覆盖。同类组件（同一 adapter_id 的 adapter、任一 agent、任一 webui）已安装时：未带 Replace 返回 ALREADY_EXISTS，safe_message 列出已安装的组件 ID；带 Replace 未带 confirm 返回 ErrNeedsConfirm（Affected 为将被卸载的组件）；确认后先完成新组件的下载与 Preflight，再按 uninstall 的缺省语义卸载旧组件，最后写新安装记录；旧组件原来启用时按 Enable 启用新组件。被替换 Adapter 的实例记录留在旧组件名下，不转给新组件，Affected 中列出这些实例。新组件在卸载旧组件之后失败时旧组件保持已卸载。Source 持久化为来源，之后的 update 只从它取版本；已卸载的保留记录只允许来源与发布者公钥都相同的重装（复用数据），否则 PERMISSION_DENIED。安装时用全部安装记录的 plugin_dependencies 检查必需依赖成环，成环拒绝。首次信任新发布者公钥需要 confirm。
- enable：检查 manifest 的必需依赖（CheckDependencies，Installed 来自已启用的包与运行中的 Adapter 描述），有缺失时 FAILED_PRECONDITION 并列出全部缺失项；可选依赖缺失只记诊断。已有 Active 时以该代次启动；首次安装没有 Active 时以「首个候选」启动（新代次、首装 SQLite、配置为 {} 或编辑快照，提交后激活；Journal 由 T-09.5 接入，本切片用一行 shortcut 注释写明）。Adapter 包启用后按 adapter_instance 中 enabled 的实例 StartInstance。
- disable：停止并保留安装与数据；被已启用包的必需依赖引用时，未带 confirm 返回 ErrNeedsConfirm（列出受影响者），确认后执行，依赖方继续运行，调用它返回 UNAVAILABLE。
- restart：以当前 Active 启动新 Session（不导入编辑快照，不产生新代次）。
- uninstall：同 disable 的确认规则；停止、解除运行注册、删除未被引用的程序包目录；默认保留数据、secret、来源绑定与 uninstalled=1 的记录。--purge 时先列出将删除的框架受管对象（代次目录、SQLite、secret、会话、按钮 token、策略、工具策略、Pending、Adapter 实例），未带 confirm 返回 ErrNeedsConfirm（Affected 为对象列表），确认后删除这些对象与安装记录；不删除插件自管的外部数据库与插件自管目录。
- pin：持久化版本约束（SemVer 范围，manifest 的语法规则），update 只接受满足 pin 的候选。
- reload/update/rollback：完成输入与前置校验（来源绑定、pin、版本不低于被依赖的 min_version 时需要 confirm），然后交给 Candidates；缺省实现返回 UNSUPPORTED。
- Adapter 实例：CreateInstance 用该 Adapter 包的配置 Schema（安装包中的 Schema 资源）经 secret.Extract 存快照；Start 时 Resolve 后经 remote.Process 启动；同一 adapter_id 下账号唯一由 remote 检查。实例配置与凭证不出现在 List 或日志中，读取只返回 Redact 后的结果。
- admin add/remove/list 经 policy 的 Admin 方法。
- HTTP API（在 Core 的主机名下，全部经 adminauth.Require）：每个方法一个 JSON 端点，需要确认时返回 409 与 {"affected": [...]}；错误返回 {"code", "safe_message"}。CLI 命令从 --admin-url（缺省取设置文件的 admin_base_url）与 ALONABOT_ADMIN_TOKEN 环境变量（或 --token-file）取得 Bearer token，不读取插件环境；确认通过 --yes 显式给出。CLI 不直接写状态库。
- WebUI 组件没有进程：enable/disable 只改记录，restart 返回 FAILED_PRECONDITION；文件的提供在 T-09.12。list 返回 role。
- serve：取得 core.lock；没有管理凭证时，stdin 与 stderr 都是终端则按 admin-token init 生成并显示一次后继续，否则拒绝启动并提示先运行 admin-token init；装配 Core（与 Run 共用 assembly），启动 HTTP Listener、webhost、管理 API，按 install_seq 顺序启动全部已启用的包。不带子命令运行 alonabot 等同 serve。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go（//go:build unix）。测试中用 goal-37 的测试工具生成签名的插件包、Adapter 包与依赖它们的插件包，在临时 data_dir 运行 alonabot run（预先用测试钩子初始化管理 token），再用 CLI 命令操作：
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

不做：候选 Journal 与准备编排（T-09.5）；唯一提交点与激活门禁（T-09.6）；崩溃恢复、回退与回收（T-09.7）；WebUI 用的会话取回与上传接口（T-09.13）；管理页面（T-09.8、T-09.9）。
```

### goal-54 · T-11.1 工具描述、目录与策略

状态：等待前置（T-04.7）。T-04.4（PR #88，合并提交 `c8dd516`）与 T-06.1（PR #45）已合并。原语句未列 T-04.7，但它依赖的写法与 goal-88 冲突：goal-88 删除 `ServiceOption`，JSON Service 的 Schema 由生成器产生（D-70、CONTRACTS §3），`jsonschema.For[In]` 得到的 Schema 可能与 Service 实际校验用的不一致，不符合 §13.1「输入 Schema 即该 Service 的输入 Schema」。因此前置加入 T-04.7（波次 8，早于本切片，波次不变）。下面的 sdk 段按 goal-88 语句的接口编写，T-04.7 合并后由审查者按合入代码核对，再改为待派发。

Core 一侧已按合入代码核对并修订：
- `route.proto` 的 `ServiceSpec` 已用到字段号 4，`tool = 5` 可用；`schema.sql` 有 `install(plugin_id)`；`internal/wire/services.go` 末尾可追加。jsonschema-go v0.4.3 已在 go.mod，删除原语句的 go.mod 范围。
- `Router` 只有按 ID 查询的 `Plugin(pluginID)`，不能遍历，范围加入 `Plugins()`。
- Service 访问规则写在 `broker.Call` step 2 内，抽为 `router.ServiceCallable` 由 Broker 与 Catalog 共用，不复制。
- `(*Mux).Handle(name string, fn ServiceFunc)` 用于注册 `core.tool.list`。
- 写明 `Effective` 与 `Policies` 在插件未 Attach 时的结果，修正 `SetPolicy` 的表述。
- 生成的 Schema 无法从 SDK 写入远程 `$ref`，该用例移到 `validate.Register` 的本地自测。

```text
/goal 使用 $alonabot-execute 推进 goal-54（切片 T-11.1：工具描述、目录与策略）。goal 目录 goal-54，分支 goal-54，worktree .worktrees/goal-54，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：JSON Service 可以声明为工具（给模型的说明、缺省策略 allow 或 confirm，输入 Schema 为该服务的输入 JSON Schema）；Core 校验工具描述，提供 core.tool.list 返回调用插件可以调用、且生效策略不是 deny 的工具；工具策略按 plugin_id/service_name 持久化，提供管理服务层的查询与修改，下一次调用生效。

依据：PLAN T-11 步骤 1、2；CONTRACTS §13.1、§13.2（策略段）、§3；WIRE-FORMAT §3.5（core.tool.list 与 Service 描述的 tool）；DECISIONS D-43、D-70；TEST-MATRIX V-55 的目录与策略分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：route.proto 的 ServiceSpec 追加 ToolSpec tool = 5；新增 tool.proto（ToolPolicy、ToolInfo、ToolListRequest/Response）。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- contracts/validate/register.go：工具描述的注册校验
- internal/wire/services.go：末尾追加 ServiceToolList = "core.tool.list"
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/tool.go
- 新增 internal/core/tool/
- internal/core/router/：新增 Plugins 与 ServiceCallable
- internal/core/broker/broker.go：访问判定改用 router.ServiceCallable，行为不变
- sdk/：services.go 的 JSONService 增加 Tool；新增 tool.go（ListTools）
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
   迁移：CREATE TABLE tool_policy (plugin_id TEXT NOT NULL REFERENCES install(plugin_id), service_name TEXT NOT NULL, policy INTEGER NOT NULL, PRIMARY KEY (plugin_id, service_name));
   读写方法自定（放在 tool.go），按主键幂等。
3. internal/core/router
   func (r *Router) Plugins() []Plugin // 当前已 Attach 的插件快照，激活状态由调用方检查
   func ServiceCallable(spec *pb.ServiceSpec, callerPluginID, targetPluginID string) bool // callerPluginID 为空（Core）或等于目标时为 true，其余按 PUBLIC 与 ALLOWLIST 判定
4. internal/core/tool
   type Catalog struct{ ... }
   func New(store *state.Store, r *router.Router) *Catalog
   func (c *Catalog) Register(m *router.Mux)                                  // 注册 core.tool.list
   func (c *Catalog) Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error) // 覆盖优先，否则缺省策略；供 T-11.2 的 invoke
   func (c *Catalog) SetPolicy(ctx context.Context, pluginID, service string, p pb.ToolPolicy) error // UNSPECIFIED 表示删除覆盖
   func (c *Catalog) Policies(ctx context.Context) ([]ToolPolicyEntry, error) // type ToolPolicyEntry struct { PluginID, Service string; Default, Override pb.ToolPolicy }
5. sdk（T-04.7 合并后核对）
   func (s *JSONService[In, Out]) Tool(description string, defaultPolicy ToolDefault) *JSONService[In, Out] // type ToolDefault int：ToolAllow、ToolConfirm；ProtoService 没有此方法
   type ToolInfo struct { Name, Description string; InputSchema json.RawMessage; Policy string } // Policy 为 "allow" 或 "confirm"
   func ListTools(ctx context.Context) ([]ToolInfo, error)

行为：
- SDK：声明了 Tool 的 JSONService，input_schema_json 取 Generated.ServiceSchemas 中 In 的 schema_id 对应的 Schema 原文，与该 Service 校验输入用的是同一份，不另外生成；链式规则与 goal-88 相同（Handle 之后调用 Tool 在 Collect 时报错）。
- 注册校验：tool 只允许出现在 encoding 为 JSON 的服务上；description 非空且不超过 1024 字节（命名常量）；default_policy 为 ALLOW 或 CONFIRM；input_schema_json 必须能被 jsonschema 解析并 Resolve（不加载远程 $ref，ResolveOptions.Loader 为 nil），根类型为 object。任一失败整体拒绝注册。
- core.tool.list：遍历 Router.Plugins() 中已激活插件的工具服务，按 ServiceCallable 只保留调用插件可以调用的，再去掉生效策略为 DENY 的；按 name 字典序返回。调用者自己的工具同样按规则出现。每次调用都读取当前策略，不缓存。
- 策略：覆盖可为 ALLOW、CONFIRM、DENY；生效策略 = 覆盖或缺省。SetPolicy 在事务提交成功后才返回；安装记录不存在时返回 NOT_FOUND，安装记录存在而插件未激活或服务尚未声明时仍可设置。Effective：有覆盖时返回覆盖；没有覆盖且插件未 Attach 或该服务不是工具时返回 NOT_FOUND。Policies 列出已 Attach 插件的全部工具与全部已保存的覆盖，取不到声明时 Default 为 UNSPECIFIED。修改在下一次 list 与 invoke 时生效。
- 工具调用（core.tool.invoke）与确认在 T-11.2。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/tool_test.go（//go:build unix）：沿用 goal-19 的 svcplugin，增加声明了 Tool 的服务（PUBLIC 的 a/search 为 allow，ALLOWLIST: b 的 a/admin 为 confirm）；b 与 c 收到消息时调用 sdk.ListTools 并回复结果：
  - b 看到 a/admin 与 a/search，c 只看到 a/search；input_schema 与 In 的类型一致，policy 正确。
  - SetPolicy(a, search, DENY) 后两者都看不到 a/search；删除覆盖后恢复；关闭并重新打开状态库后覆盖仍在。
  - description 为空的工具使插件启动失败。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：validate.Register 拒绝 Protobuf 服务上的 tool、含远程 $ref 或根类型不是 object 的 input_schema_json；description 长度边界；Policies 的输出；未知 ToolPolicy 被拒；Effective 的 NOT_FOUND。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go mod tidy 后无差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：core.tool.invoke、确认与确认按钮（T-11.2）；Agent 组件与 MCP（T-12）；管理 API 与页面（T-09.14、T-09.9）。
```

### goal-69 · T-13.3 开发运行与热部署

状态：等待前置（T-07.6、T-09.16、T-13.1）。2026-11-04 按 D-77 改写：热部署不再是库模式的 `Reload` 选项与 `ALONABOT_RELOAD`，而是 `alonabot run [目录...] --reload`，开发组件以独立进程运行。语句只写了目标、依据和要点，前置合并后由审查者按 goal-67 的装配、goal-97 的 settings 与 goal-100 的工程 manifest 写出接口与测试段。

目标：`alonabot run [目录...] [--reload]` 把给出的目录（没有给出且当前目录有 plugin.yaml 时为当前目录）作为开发组件运行：按 `alonabot build` 的方式构建到 `<data_dir>/dev/bin/`，插件、Adapter 与 Agent 以独立进程运行，WebUI 以目录中的 assets 代替已安装的 WebUI。`--reload` 时轮询源码、资源与配置，源码变化时重新构建并只重启变化的组件，编译失败时旧组件继续运行，只有配置变化时不重新构建。

依据：PLAN T-13 步骤 4；CONTRACTS §15.4、§15.1、§16（DEV_RELOAD_POLL）；DESIGN §S；DECISIONS D-77；TEST-MATRIX V-64 的独立进程分项、V-65。

要点：
- 组件身份取自目录中的工程 plugin.yaml（goal-100 的解析）；构建失败或 ID、adapter_id 与已安装的包冲突时，在启动任何组件前报错退出。
- 构建复用生成器与 `go build`，提供供 goal-38 的 `alonabot build` 复用的内部函数，不另写一套。
- 开发组件使用 goal-67 的 dev 来源安装记录与固定代次；进程的启动、停止与有限重启沿用 supervisor，与已安装组件相同。
- 轮询：每 DEV_RELOAD_POLL 用标准库检查每个组件目录中的 `.go` 文件、`//go:embed` 引用的文件、go.mod、go.sum、plugin.yaml、config_schema.json、instance_schema.json 与 i18n/（排除数据目录、vendor、以 . 或 _ 开头的目录与 testdata），以及插件的 config.json 与 core_config.json 中该 Adapter 的实例；源码或资源变化时重新构建，成功后依次重启变化的组件；只有配置变化时只重启该组件，Adapter 只重启变化的实例；core_config.json 的其他变化交给 goal-97 的 Reload。
- 删除现在 `run` 的 `--plugin-id`、`--entry` 参数，不留兼容；`--data-dir`、`--stdio`、`--listen` 与 `--args` 保留。
- 提交的集成测试（//go:build unix）在临时目录用 `alonabot init` 生成插件工程并以 `--reload` 运行：改动回复文本后插件重启并回复新文本、Core 与回环 Adapter 不重启；写入语法错误时旧插件继续回复；只改 config.json 时不重新构建但新配置生效；改动 config_schema.json 后重新构建。

## 第 10 波

### goal-9 · T-03.10 Telegram 接入开发入口与故障矩阵

状态：等待前置（T-03.8、T-03.9、T-07.6、T-13.3）。语句中 goal-4～goal-8、goal-50、goal-51 提供的符号（remote.Process、supervisor.Spec 的 Role、adapters/telegram、telegramtest、media.Store、buttons.Service、message 的 Send/Edit）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：开发入口改由 goal-69 的 `alonabot run <目录>` 提供，本切片不再给 run 加 `--adapter-id` 等参数；Adapter 实例写在 `core_config.json` 的 `adapters` 中，不再读取 instances/ 目录。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-9（切片 T-03.10：Telegram 接入开发入口与故障矩阵）。goal 目录 goal-9，分支 goal-9，worktree .worktrees/goal-9，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：alonabot run 可以同时启动一个 Adapter 包与一个插件，Adapter 的实例配置取自 <data_dir>/adapters/<adapter_id>/instances/<instance_id>.json；提供 echo 与媒体示例；用本地 Bot API 模拟服务跑完 PLAN T-03 步骤 8 的发送与接收故障矩阵，并在 TEST-MATRIX §3.2 记录 T-03 的证据。真实 Telegram 没有测试账号时记为未运行。

依据：PLAN T-03 步骤 8 与「停止和失败条件」；CONTRACTS §8、§15.1（`adapters`）、§15.4；MESSAGE-MODEL §4、§6；TEST-MATRIX V-07、V-08、V-09、V-35、V-44、V-45、V-48、V-51～V-54 的 T-03 分项。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：正式安装与实例管理（T-09.4）；实例配置的 Secret Store（T-09.4）；其他平台的 Adapter。
```

### goal-16 · T-05.2 Rule、HandlerOutcome 与传播

状态：等待前置（T-05.1、T-04.7、T-04.8）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。语句中 goal-2、goal-12、goal-15 提供的符号（sdk 的分发、RouteOption、Context、HandlerOutcome 的处理、router 的命令层与事件路由派发）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-16（切片 T-05.2：Rule、HandlerOutcome 与传播）。goal 目录 goal-16，分支 goal-16，worktree .worktrees/goal-16，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

状态：等待前置（T-05.1、T-07.1）。语句中 goal-2、goal-15、goal-23（contracts/schedule）、goal-76 提供的符号（router.Router、router.Plugin、wire.Session 的 Call、schedule.Trigger、tune.Values、sdk 的 ScheduleRegistry 与分发）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-24（切片 T-07.2：schedule 注册与运行槽）。goal 目录 goal-24，分支 goal-24，worktree .worktrees/goal-24，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
   type Options struct { JobTimeout time.Duration; Now func() time.Time } // JobTimeout 取 tune.Values.JobTimeout
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

### goal-35 · T-08.9 configdb 示例与 T-08 验收

状态：等待前置（T-08.6、T-08.7、T-08.8、T-08.11）。语句中 goal-28～goal-34、goal-53 提供的符号与工具（生成器、Prepare、secret.Store、gormdb、candidate、web/admin 的表单模型、contracts/i18n）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-35（切片 T-08.9：configdb 示例与 T-08 验收）。goal 目录 goal-35，分支 goal-35，worktree .worktrees/goal-35，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
1) 配置迁移：config_version 1 的旧配置只有经 OnConfigLoad 才满足 2 的 Schema；先迁移再校验，结果写入候选快照。
2) 显式空值：false、0、""、[]、{}、null（指针字段）全部保留；缺失键补默认值；未知键与嵌套未知键保留在快照中。
3) secret：快照只有引用；Schema 移除 secret 字段但候选保留该键时仍为引用；日志中没有明文。
4) DB：CRUD 与事务（tx 回调）；versioned 两步迁移；第二步失败时账本只有第一步、候选失败、Active 不变。
5) 候选：reload 后 OnStart 写入候选库再失败：Discard，Active 的数据与配置不变，旧版重启正常；新旧进程没有同时写同一个库（runtime.lock 证明）。
6) 两个入口：同一插件同时用 sqlitedb.Open 与 gormdb.Open 时共用同一连接，一边写入另一边立即可读；只用 sqlitedb 的示例变体不链接 gorm.io/gorm。
7) i18n：Context.T 的语言选择与回退；alonabot check 拒绝缺少 en-US 与占位符不一致。
8) 表单（web/admin 的 vitest，非 DOM）：configdb 的 Schema 字段顺序与声明一致；hidden 不进 patch；未修改的 secret 不进 patch；i18 标签回退。
- 证据：TEST-MATRIX §3.2 追加 T-08 一行；需要 T-09 正式 reload、WebUI/API 与 Journal 崩溃恢复才能关闭的 V-23～V-25、V-38 场景单独列为「未闭合」，不标成通过。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t08_e2e_test.go；web/admin 的 schema.test.ts 追加用例。
- 本地自测（不提交，PR 中写明结果）：npm run build；示例的 alonabot check。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；web/admin 下 npm run test 通过。

不做：正式 reload 与管理入口（T-09）；Journal 与崩溃恢复（T-09.5～T-09.7）。
```

### goal-101 · T-08.12 Adapter 实例配置与 Core 设置的版本迁移

状态：等待前置（T-07.6、T-08.11）。依据 D-85 新增。语句只写了目标、依据和要点，前置合并后由审查者按 goal-95 的 contracts/configschema 与 goal-97 的 settings 写出接口与测试段。

目标：Adapter 包用 instance_schema.json 声明实例配置（格式与 config_schema.json 相同），Core 按它补缺失值、校验并迁移 core_config.json 中的实例；core_config.json 自身的 $version 由 Core 内置的迁移处理。迁移前备份，失败与降级时报错，不丢配置。官方 Telegram Adapter 补上 instance_schema.json。

依据：PLAN T-08 步骤 1；CONTRACTS §5.5、§8、§15.1；DECISIONS D-85；TEST-MATRIX V-38 的迁移分项、V-88。

要点：
- 复用 goal-95 的 contracts/configschema（Parse、迁移、补缺失值、校验与 secret 路径），不另写一套。
- build 与 check 对 role 为 adapter 的工程检查 instance_schema.json 及其迁移；包内带上该文件。
- Core 启动 Adapter 的实例前，按包内 instance_schema.json 迁移、补缺失值并校验 core_config.json 中的实例；版本提高时经 settings 的 Update 写回，并先把原文件备份到 data/backups/。迁移失败时该实例不启动并记 ERROR，其他实例照常。提供迁移试算函数，供 Adapter 更新的预检在停止旧版本前调用（T-09.5 接入）。
- core_config.json 的 $version：Core 内置迁移列表（当前为空），启动时按 §5.5 执行；版本高于程序时拒绝启动并写明两个版本号。
- 官方 Telegram Adapter 的 instance_schema.json 声明 token（secret、required）、api_base_url 与 proxy，缺省值与现在的解析一致。
- 提交的集成测试覆盖：实例配置的 rename 迁移与写回、备份文件的生成；迁移失败时实例不启动且文件不变；core_config.json 版本高于程序时拒绝启动。

### goal-84 · T-09.12 WebUI 组件与组件图标

状态：等待前置（T-07.7、T-09.4、T-13.1）。语句只写了目标、依据和要点。依据 D-62、D-63 新增；前置合并后由审查者按实际的 webhost、adminauth、manage 与根包装配写出接口和测试段。2026-11-04 随前后端拆分从 WebUI 章节移入核心章节。2026-11-04 按 D-77～D-85 修订计划：Core 只有一个主机名（D-84），WebUI 与插件路由同在其下；spec/ADMIN-API.md 还要写出插件页面桥的消息格式。

目标：Core 以组件方式提供 WebUI 与组件图标：已启用 WebUI 组件的静态文件（保留 `/api/` 与内部路径，无扩展名的不存在路径回退 index.html，未安装时 503 与英文提示），Admin API 主版本常量与 WebUI 声明范围的检查（不兼容时拒绝安装与启用并提示），只提供 manifest 声明的图标并带 sandbox CSP 与 nosniff、要求管理员会话；写出 spec/ADMIN-API.md（主版本、全部接口、认证与错误格式）；根包增加 `alonabot.WebUI(fsys fs.FS)` 选项。

依据：PLAN T-09 步骤 9、§3.1 的 T-09.12 行；CONTRACTS §7.3、§7.5、§9.1、§15.5；DECISIONS D-62、D-63；TEST-MATRIX V-83。

要点：
- 文件只从不可变包目录读取，路径按 CONTRACTS §9.1 的相对路径规则校验，不跟随链接，不经组件进程。
- 图标的类型只按扩展名取 image/svg+xml、image/png、image/webp；未声明的包内文件即使存在也返回 404。
- spec/ADMIN-API.md 由本切片新建，写入此时已有的全部 Admin API（goal-36、goal-39 等）；之后新增接口的切片（goal-66、goal-92、goal-93）同步追加；同一主版本内只做向后兼容的增加。
- 原 goal-43 要求 Core 给 WebUI 静态文件加 `Content-Security-Policy: default-src 'self'`、`X-Frame-Options: DENY` 与 `nosniff`。CONTRACTS §7.5 目前只规定了图标的响应头，没有这一条；补写接口段前先由维护者确认，确认后写入 §7.5 再列入本语句。
- 提交的集成测试经真实 Core 与临时安装的 WebUI 测试包覆盖 V-83 的全部场景；含脚本的 SVG 在浏览器中不执行的检查没有浏览器时记为未运行。

### goal-68 · T-13.2 开发组件的配置与数据库

状态：等待前置（T-07.6、T-08.11、T-13.1）。语句中 goal-30、goal-32、goal-67 提供的符号（sdk/config.Prepare、gormdb 的描述、根包 assembly 与 dev 安装记录）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：开发 Adapter 的实例取自 `core_config.json`（D-82）；WebUI 修改开发组件配置时写回 `config.json` 并重启该组件（Admin API 在 goal-93）；dev 来源同时拒绝 uninstall（D-77）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-68（切片 T-13.2：开发组件的配置与数据库）。goal 目录 goal-68，分支 goal-68，worktree .worktrees/goal-68，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：库模式的开发组件每次启动以 <data_dir>/plugins/<id>/config.json（缺失时为空对象）为配置，经 SDK 准备与校验后直接生效，不走候选与 Journal；受管 SQLite 为 <data_dir>/plugins/<id>/dev.sqlite，不复制、不回退；dev 来源拒绝 update 与 rollback；开发组件的缺省日志级别为 debug。

依据：PLAN T-13 步骤 3；CONTRACTS §15.4；TEST-MATRIX V-64。

范围（只新增或修改这些）：
- 根包 assembly.go：开发组件的 Bootstrap 配置与数据库描述（只改开发组件分支）
- internal/core/manage/：dev 来源的 update 与 rollback 拒绝（若 goal-39 已合并；否则在 assembly 中拒绝对应入口）
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- 配置：开发组件启动时读取 config.json；文件不存在时为 {}；内容不是 JSON 对象时 Run 返回错误并指明路径。candidate_config_json 为其内容，previous_config_version 为 nil（每次都按首装处理 OnConfigLoad），SDK 按 goal-95 准备与校验；校验失败时该组件启动失败，Run 返回错误。准备后的配置不回写 config.json，也不写入代次快照之外的地方（代次 dev 的快照每次覆盖）。
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

## 第 11 波

### goal-14 · T-04.5 示例升级与 T-04 验收

状态：等待前置（T-03.10、T-04.4、T-04.7、T-04.8、T-04.9）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。T-04.2、T-04.3 已合并；已核对 `CommandRegistry.Command`、`Regex`、`Keyword`、`Raw`，以及 Args、Optional、RawTail、Usage、Silent、IgnoreCase 和 Context.Match。策略与 Adapter 开发入口的实际接口仍待前置合并后核对。2026-11-04 按 D-77～D-85 修订计划：示例只用链式注册（D-79，goal-96 删除标记），用 `alonabot check`、`build` 检查与构建，经 `alonabot run <目录>` 运行（D-77）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-14（切片 T-04.5：示例升级与 T-04 验收）。goal 目录 goal-14，分支 goal-14，worktree .worktrees/goal-14，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
- 示例：echo 注册 /echo（显式注册，纯显式路径）、/note add <text:raw>、/note list、别名 /n、/calc <a:int> [b:int=1]、一个 Regex（例如 ^roll (\d+)d(\d+)$）、一个 keyword 与一个 Raw；其中 /note 组由标记生成，/calc 为实例方法显式注册。alonabot check 对两个示例通过，生成文件与提交的版本一致。
- 构建：以真实 go build 编译示例，经 alonabot run（memory 之外，用 goal-9 的 Telegram 模拟服务或回环 Adapter）启动并收发：每条命令的回复正确；参数错误时发 usage 一次；Regex 捕获组正确；命中命令时 keyword 与 Raw 不运行。
- 生成器 fixture：多 root、多 Config、错误的方法签名（编译失败或 Collect 报错，以实际 go build 的结果为准）、依赖 module 中的伪标记不被扫描、不同 build tags 下命令集合不同、重复路由与别名冲突、参数溢出（int64 越界）各产生预期的诊断或运行结果。
- 策略：两个真实插件 Session 注册同名命令，CommandIndex 与 /help 之外的诊断显示胜出者与被遮蔽者；禁用胜出者后路径仍被占用（被遮蔽者不接管）；经 policy 服务层修改权限后下一条消息即生效，插件进程 PID 不变；关闭并重启 Core 后策略仍在。
- Rule 相关的命令消费在 T-05.2 之后才有真实 SDK Rule；本切片对 Rule false/error 的消费状态用 HandlerOutcome fixture 验证，并在证据中标明。
- 证据：TEST-MATRIX §3.2 追加 T-04 一行，写明测试位置、命令、平台与结果；只有 fixture 覆盖的项单独标明。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t04_e2e_test.go；sdk/generate 中对 testdata fixture 的测试。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：示例的每条命令在 Windows/Linux 交叉编译（只编译，不运行）；生成文件的稳定性。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；在两个示例目录运行 alonabot check 通过。

不做：真实 Rule 与 Hook 的集成（T-05.5）；WebUI 修改策略（T-09.9）。
```

### goal-17 · T-05.3 三阶段 Hook 与 send_forbidden

状态：等待前置（T-05.2）。语句中 goal-2、goal-12、goal-15、goal-16、goal-51 提供的符号（router 的派发链与可信信息、message.Service 的 Send 与 Edit、sdk 的分发与 RouteOption）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-17（切片 T-05.3：三阶段 Hook 与 send_forbidden）。goal 目录 goal-17，分支 goal-17，worktree .worktrees/goal-17，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
   func New(r *router.Router) *Runner
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

状态：等待前置（T-05.2）。语句中 goal-2、goal-15、goal-16、goal-76 提供的符号（router 的派发链与 Invocation、sdk 的 Context 与分发、tune.Values）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-52（切片 T-05.4：延续调用与多轮等待）。goal 目录 goal-52，分支 goal-52，worktree .worktrees/goal-52，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
   func New(r *router.Router, opt Options) *Manager
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

状态：等待前置（T-07.2、T-07.4）。语句中 goal-24～goal-26 提供的符号（scheduler.Scheduler、httpproxy.Server、webhost.Host、host 的停止流程与 StopTiming、supervisor.Running）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：HTTP 边界验收中的两个 Host 与单 Host 主动内容改为插件响应的沙箱头与 Set-Cookie 过滤（D-84，goal-98）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-27（切片 T-07.5：调度与 HTTP 启停集成及验收）。goal 目录 goal-27，分支 goal-27，worktree .worktrees/goal-27，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：管理认证（T-09.1，测试注入假会话）；serve（T-09.4）；Agent 定时任务（T-12.6）。
```

### goal-66 · T-09.10 日志管理

状态：等待前置（T-07.6、T-09.4、T-09.12）。语句中 goal-2、goal-39、goal-65（logging 的 Output、Levels、Source）、goal-84（spec/ADMIN-API.md）提供的符号只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按前后端拆分改写：日志页移到 goal-94（T-09.15），本切片不再等待 T-09.8。2026-11-04 按 D-77～D-85 修订计划：按来源的日志级别写在 `core_config.json` 的 `log.sources`，经 goal-97 的 settings 写回（D-82），不存状态库。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-66（切片 T-09.10：日志管理）。goal 目录 goal-66，分支 goal-66，worktree .worktrees/goal-66，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按来源的日志级别持久化在状态库，经管理服务修改后通过 sdk.log.set_level 通知 SDK 并立即生效；全局级别可在运行时修改、Core 重启后恢复为启动设置；Core 为每个来源保留最近 LOG_BUFFER_RECORDS 条已输出的记录，提供按来源、级别与 trace_id 的读取和 SSE 订阅，过去日期从日志文件查询；alonabot logs 与 log-level 命令、WebUI 日志页（T-09.15）使用同一 API。

依据：PLAN T-09 步骤 7（日志）；CONTRACTS §14.2～§14.6、§16（LOG_BUFFER_RECORDS）；WIRE-FORMAT §3.5（sdk.log.set_level）；TEST-MATRIX V-62、V-63、V-68。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：新增 LogSetLevelRequest/Response。随后 go generate ./...（本切片获授权只追加这两个消息）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceLogSetLevel = "sdk.log.set_level"
- internal/core/logging/：新增 buffer.go（最近记录）、query.go（按日期查询日志文件）；Output 在输出时写入缓冲（只加调用点）
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/loglevels.go
- internal/core/manage/：新增 log_api.go（读取、SSE、级别修改）
- sdk/：处理 sdk.log.set_level（新增文件；sdk/adapter 的 RunAdapter 同样处理）
- cmd/alonabot/：新增 logs 与 log-level 命令
- spec/ADMIN-API.md：追加日志接口（本切片获授权修改该文件）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message LogSetLevelRequest { LogLevel level = 1; }
   message LogSetLevelResponse {}
2. internal/core/state
   迁移：CREATE TABLE log_level (source_id TEXT PRIMARY KEY, level TEXT NOT NULL); -- level 为 debug、info、warn、error、off；inherit 不存行
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
- 日志内容中的脱敏规则不变（goal-65 与 CONTRACTS §14.5）。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/logs_e2e_test.go（//go:build unix）：运行 serve 与一个持续输出日志的测试插件：
  - log-level 把该插件设为 debug 后，SDK 输出 DEBUG 记录且 logs 能读到；设为 off 后插件记录消失，Core 关于它的记录仍在；重启 serve 后按来源的设置保留，全局级别的运行时修改恢复为启动值。
  - logs --follow 收到新记录；未带 token 时 401。
  - 每个来源最多保留 LOG_BUFFER_RECORDS 条（用很小的值）。
  - 写入前一天日期的日志文件后，logs --date 能查到。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：环形缓冲的边界；历史文件解析（多行值、转义）；SSE 的断开。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交差异；go vet ./...；go test -race ./...；go build ./...。

不做：日志页（T-09.15）；远程日志收集与外部格式；日志文件设置的运行时修改（启动设置，§15.1）。
```

### goal-92 · T-09.13 管理页所需的会话与上传接口

状态：等待前置（T-09.1、T-09.4、T-09.12）。2026-11-04 按前后端拆分新增，内容取自原 goal-43（T-09.8）语句中的后端部分：goal-36 合并的 adminauth 只有 login 与 logout，goal-39 的安装与更新只接受本地路径，浏览器中的 WebUI 用不了。语句只写了目标、依据和要点，前置合并后由审查者按实际的 adminauth、管理服务与 spec/ADMIN-API.md 写出接口和测试段。

目标：补齐浏览器中的 WebUI 需要、goal-36 与 goal-39 没有提供的 Admin API：已登录时取回当前会话的 csrf_token（GET /api/auth/session）；以 multipart 上传安装包与更新包（archive、release.json、.sig 三个文件），写入临时目录后调用管理服务的 Install 与 Update；总览页显示的 Core 版本与监听地址。新接口追加到 spec/ADMIN-API.md。

依据：PLAN T-09 步骤 7（管理页）；CONTRACTS §7.3、§7.4、§7.5、§9.3；DECISIONS D-62；TEST-MATRIX V-40、V-47 的 API 分项。

要点：
- GET /api/auth/session 只在已登录时可用，返回当前会话的 csrf_token 与过期时间；不签发新会话，不返回管理 token。
- 上传端点经 adminauth.Require 与 CSRF 检查；预检、确认（409 与 affected）与错误格式和 goal-39 的 JSON 端点一致；临时文件在调用结束后删除，失败时不留残余。
- Core 版本与监听地址不含任何 secret；Admin API 主版本沿用 goal-84 的常量。
- 提交的集成测试经真实 Core 覆盖：未登录取会话返回 401；登录后取回的 csrf_token 可用于修改请求；multipart 安装与更新成功；缺少文件或签名错误时拒绝，已有组件不受影响，临时目录为空。

### goal-70 · T-13.4 开发运行示例与验收

状态：等待前置（T-03.10、T-13.2、T-13.3）。语句中 goal-9、goal-67～goal-69 提供的符号与示例（根包 Run、Reload、examples/devbot、回环 Adapter）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：示例改为 `alonabot init` 生成的工程，经 `alonabot run --reload` 运行，实例写在 `core_config.json`（D-77、D-82）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-70（切片 T-13.4：开发运行示例与验收）。goal 目录 goal-70，分支 goal-70，worktree .worktrees/goal-70，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：发行与预装（T-14）。
```

### goal-81 · T-16.1 插件测试套件基础

状态：等待前置（T-05.2、T-13.2）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。语句中 goal-4、goal-16、goal-67、goal-68 提供的符号（sdk/adapter 的 Adapter 与 Instance、Router 的分发流程、alonabot.Start 与 Core、Plugin/Adapter/DataDir 选项、dev 的 config.json 与 dev.sqlite、生成器的 Generated）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：测试实例写在 `core_config.json` 而不是 instances/ 文件；发信许可缺省为全部实例，套件选项用于收紧（D-82、D-83）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-81（切片 T-16.1：插件测试套件基础）。goal 目录 goal-81，分支 goal-81，worktree .worktrees/goal-81，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：新建公开包 plugintest：在 go test 中用 alonabot.Start 在临时数据目录上装配真实 Core，登记被测插件与套件自带的测试 Adapter；提供会话与用户构造、Core Admin、发信许可、Capability 子集与配置的选项；注入一条消息并在这个事件的全部分发链结束后返回期间产生的 Send、Edit 与 Handler 错误。分发结束由 Core 经 internal/inproc 通知，不靠等待时间。

依据：PLAN T-16 步骤 1、2 与「失败停止条件」；CONTRACTS §17、§15.5、§8（Target 准入）、§2；DECISIONS D-47；ARCHITECTURE 的 plugintest 一行；TEST-MATRIX V-73 的基础分项。

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

## 第 12 波

### goal-18 · T-05.5 多进程事件流水线验收

状态：等待前置（T-03.8、T-05.3、T-05.4）。语句中 goal-8、goal-15～goal-17、goal-52 提供的符号（按钮、Rule、Hook、延续调用与等待、Router 的 Stats）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-18（切片 T-05.5：多进程事件流水线验收）。goal 目录 goal-18，分支 goal-18，worktree .worktrees/goal-18，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：Service 嵌套与 Hook→Service→Send（T-06.2、T-06.4）；真实 Scheduler（T-07.5）。
```

### goal-20 · T-06.2 父调用链、取消与背压

状态：等待前置（T-03.6、T-05.3、T-06.1）。语句中 goal-7、goal-17、goal-19 提供的符号（broker.Broker、router.Invocation 与 send_forbidden、media 的 Grant、wire 的 ParentAttach 与预算）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-20（切片 T-06.2：父调用链、取消与背压）。goal 目录 goal-20，分支 goal-20，worktree .worktrees/goal-20，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

### goal-38 · T-09.3 开发 CLI 与库模式脚手架

状态：等待前置（T-04.5、T-08.8、T-08.11、T-09.16、T-13.3）。2026-11-04 按 D-76 改为核对与打包 config_schema.json。语句中 goal-11、goal-37、goal-53、goal-67、goal-95 提供的符号（sdk/generate、packages 的 Preflight 与包格式、contracts/release、contracts/i18n、根包 Run、contracts/configschema）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：开发命令改为顶层的 init、build、check、test、pack（D-77），build 取代 plugin generate，init 在当前目录创建工程、不生成 main.go，`.gitignore` 忽略 data/；包内 manifest 由 goal-100 的函数补上构建信息（D-78）；签名可选（D-81）；`run` 由 goal-69 实现；代码中提到旧命令名的提示（生成文件头的 `alonabot plugin generate`、`Collect` 报错中的命令）一并改为新命令。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-38（切片 T-09.3：开发 CLI 与库模式脚手架）。goal 目录 goal-38，分支 goal-38，worktree .worktrees/goal-38，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：完成开发命令 alonabot init|check|build|test|pack：init 生成插件骨架与库模式调试用的 main.go；check 验证 CONTRACTS §9.1 的静态契约；build 用生成器为根插件生成调用 sdk.Run 的正式入口并按目标 OS/arch 编译，产物不含 Core；test 在临时数据与离线测试会话中检查加载与注册；pack 为四种角色的组件生成 tar.gz 安装包与归档外的 release.json（WebUI 只含 plugin.yaml、icon 与 assets 目录，平台为 any，不运行 build），并可用本地密钥对清单的确切字节签名。

依据：PLAN T-09 步骤 2（CLI）；CONTRACTS §9.1～§9.3、§15.4、§7.5；DECISIONS D-42、D-43、D-59、D-62、D-63；TEST-MATRIX V-01、V-41、V-54 的打包分项。

范围（只新增或修改这些）：
- cmd/alonabot/：plugin 命令组增加 init、build、test、pack，check 补齐静态契约（新文件）
- 新增 sdk/generate/entry.go（正式入口生成）
- 新增 internal/pluginkit/（打包、release.json 生成与签名，供 cmd 使用）；packages 的包格式与 goal-37 一致
- 上述包的测试文件
允许的新跨模块边：cmd/alonabot -> internal/pluginkit；internal/pluginkit -> contracts/...、-> sdk/generate（ARCHITECTURE 已登记）。

接口：
- alonabot init <dir> --id ID [--role plugin|adapter] [--adapter-id ID]：生成 plugin.yaml（manifest_version 1、version 0.1.0、source 为占位并提示修改）、带 sdk.Base 的根插件、一个示例命令、i18n/en-US.json、main.go（调用 alonabot.Run(alonabot.Plugin(...))，供库模式调试）与 .gitignore（忽略 .alonabot/，其中的实例配置可能含平台凭证）；目录非空时拒绝。
- alonabot check [--dir] [--goos] [--goarch] [--tags]：在 goal-11/goal-53/goal-95 的检查之外，核对 manifest（Parse 与 §9.1 的字段规则、role 与描述一致、migration_mode 与实际迁移接口一致、prepare_effects 已声明）、config_schema.json（goal-95 的 Parse）与 config_version、i18 文件；全部问题一次列出。
- alonabot build [--dir] --goos OS --goarch ARCH [--out DIR]：生成正式入口（临时目录中的 main 包，调用 sdk.Run 或 sdk.RunAdapter，引用根类型），用锁定的 Go Modules 编译（CGO_ENABLED=0，-trimpath），输出可执行文件；不链接根包 alonabot。
- alonabot test [--dir]：build 当前平台后，在临时 data_dir 中经 Core 的开发入口启动一次，完成握手、准备、注册与激活后停止，报告注册的路由；不连接任何真实平台，不读取生产凭证。
- alonabot pack [--dir] --goos OS --goarch ARCH [--key FILE] [--out DIR]：build 后打包 tar.gz（plugin.yaml、entrypoint、存在时的 config_schema.json、i18n/、build-info，以及 manifest 的 icon）；role 为 webui 时不需要 --goos/--goarch，只打包 plugin.yaml、icon 与 assets 目录，归档名为 <id>-<version>-any.tar.gz；生成 release.json（id、version、source、os/arch、归档 SHA256 与长度、发布者公钥 ID）；给出 --key（Ed25519 私钥文件）时对 release.json 的确切字节签名生成 release.json.sig，否则只输出待签名的 release.json。私钥不进入包。
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

### goal-82 · T-16.2 按钮、多轮等待、schedule 与示例测试

状态：等待前置（T-03.8、T-05.4、T-07.2、T-16.1）。语句中 goal-8、goal-24、goal-52、goal-81 提供的符号（Button.core_token 与 button.clicked 的派发、调度器的运行入口、core.wait 的等待状态、plugintest 与 inproc.Harness）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-82（切片 T-16.2：按钮、多轮等待、schedule 与示例测试）。goal 目录 goal-82，分支 goal-82，worktree .worktrees/goal-82，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./plugintest/... ./examples/... ./internal/core/... ./internal/inproc/...；go test -race . ；go build ./...。

不做：模拟时钟或缩短 schedule 的周期（只提供立即运行）；WebUI 沙盒（v0.1 之外）；为 Adapter 作者提供的测试工具。
```

## 第 13 波

### goal-22 · T-06.4 Session 清理与 Service 端到端

状态：等待前置（T-06.2、T-06.3）。语句中 goal-15、goal-17、goal-19～goal-21 提供的符号（router、broker、bus、hook 与各自的 Stats）只写名字；合并后由审查者核对，再改为待派发。T-06.3 已由 PR #56 合并，bus 的实际接口：`bus.New(opt bus.Options) *Bus`（`Options{QueueEvents int; QueueBytes int64}`，装配处传入 tune.Values 的值）、`(*Bus).Register(m *router.Mux)`、`Attach(p router.Plugin)`、`Detach(pluginID string)`、`Stats(pluginID string) Stats`（`Stats{Published, Delivered, Dropped uint64}`，丢弃不分原因；本切片要按原因计数时在 bus 中扩展）；Router 不持有 Bus，Attach/Detach 由装配处与 Router 的 Attach/Detach 一起调用；`router.Plugin` 增加了 `SendEvent`，wire 的接收丢弃计数为 `wire.Stats.DroppedEvents`。

```text
/goal 使用 $alonabot-execute 推进 goal-22（切片 T-06.4：Session 清理与 Service 端到端）。goal 目录 goal-22，分支 goal-22，worktree .worktrees/goal-22，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：管理面对计数的读取接口（T-09.14）；可靠事件存储与服务重试（契约不提供）。
```

### goal-55 · T-11.2 工具调用与用户确认

状态：等待前置（T-03.8、T-06.2、T-08.8、T-11.1）。语句中 goal-8、goal-20、goal-53、goal-54 提供的符号（buttons.Service、broker 的父链调用、internal/core/i18n、tool.Catalog）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-55（切片 T-11.2：工具调用与用户确认）。goal 目录 goal-55，分支 goal-55，worktree .worktrees/goal-55，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

## 第 14 波

### goal-40 · T-09.5 候选 Journal 与准备编排

状态：等待前置（T-09.4、T-08.6、T-06.4）。语句中 goal-22、goal-33、goal-39 提供的符号（manage.Service 与 Candidates 接口、candidate.PrepareDatabase/Discard 与交接约定、supervisor 的停止流程）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-40（切片 T-09.5：候选 Journal 与准备编排）。goal 目录 goal-40，分支 goal-40，worktree .worktrees/goal-40，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：为 reload、update 与首次启用实现正式的候选编排：预检通过后，以预期的旧 ActiveGeneration 写入持久化的管理操作记录（Journal）及其资源引用；冻结新根派发、drain 并停止旧进程，确认进程树退出且 runtime.lock 可取；调用 T-08.6 复制数据库并创建候选 Session（启动原因、维护许可与阶段预算放进 Bootstrap）；REGISTER 校验后持久化程序、配置快照、secret 与数据库的候选引用，形成 PREPARED。提交与激活门禁留给 T-09.6；失败时清理候选并以旧版的新 Session 恢复。

依据：PLAN T-09 步骤 4；CONTRACTS §12、§6.3、§5.2、§4.2；TEST-MATRIX V-29。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/journal.go
- 新增 internal/core/journal/（操作记录的状态机与编排，实现 manage.Candidates 的 Reload 与 Update）
- internal/core/manage/：首次启用改用 journal 编排，删除 goal-39 的 shortcut 注释（只改该处与装配）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：CREATE TABLE journal (op_id TEXT PRIMARY KEY, plugin_id TEXT NOT NULL REFERENCES install(plugin_id), kind TEXT NOT NULL, expected_old TEXT NOT NULL, new_generation TEXT NOT NULL, phase TEXT NOT NULL, package_dir TEXT NOT NULL, database_path TEXT NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL);
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

不做：唯一提交点与激活门禁（T-09.6）；崩溃恢复与 rollback（T-09.7）；维护许可的管理入口（T-09.14、T-09.9）。
```

### goal-57 · T-11.4 工具与会话的验收

状态：等待前置（T-11.2、T-11.3）。语句中 goal-54～goal-56 提供的符号与测试插件只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-57（切片 T-11.4：工具与会话的验收）。goal 目录 goal-57，分支 goal-57，worktree .worktrees/goal-57，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：Agent 组件的工具循环（T-12.3）；管理页面（T-09.9）。
```

### goal-102 · T-16.3 plugintest 的 LLM 替身

状态：等待前置（T-12.8、T-16.1）。依据 D-77 新增。语句只写了目标、依据和要点，前置合并后由审查者按 sdk/llm 与 plugintest 的实际代码写出接口与测试段。

目标：plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error)) 登记一个只提供 llm.generate 的测试 Agent，被测插件经 @agent 的调用交给这个函数处理；没有登记时 @agent 按目标未激活处理。

依据：PLAN T-16 步骤 3；CONTRACTS §13.5、§17；DECISIONS D-77、D-83；TEST-MATRIX V-73 的 LLM 分项。

要点：
- 测试 Agent 是库模式的开发组件（role 为 agent），走真实的 Service Broker、调用链与 llm.generate 的 Schema 校验；函数返回的错误按 llm.generate 的错误码交回调用方。
- llm_service_allowed_plugins 按缺省的 ["*"] 处理，不需要额外授权。
- 提交的集成测试覆盖：插件调用 llm.generate 得到替身的回复；替身返回错误时插件收到对应错误码；没有登记替身时调用失败。

## 第 15 波

### goal-41 · T-09.6 唯一提交点与激活门禁

状态：等待前置（T-07.5、T-09.5）。语句中 goal-27、goal-40 提供的符号（journal.Orchestrator 与 Result、supervisor 的激活与接入点、host 的 Degraded）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-41（切片 T-09.6：唯一提交点与激活门禁）。goal 目录 goal-41，分支 goal-41，worktree .worktrees/goal-41，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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
- 提交前置：journal 阶段为 prepared、候选代次的 config_snapshot 与 generation_secret_path 已写入、候选目录与数据库文件存在（fsync 已完成），否则不提交，按 Abort 处理。
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

## 第 16 波

### goal-42 · T-09.7 崩溃恢复、回退与回收

状态：等待前置（T-09.6）。语句中 goal-33、goal-39～goal-41 提供的符号（journal 的阶段与 Abort、state.CommitCandidate、candidate 的 PrepareDatabase/Discard、manage.Candidates.Rollback、secret.Collect）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-42（切片 T-09.7：崩溃恢复、回退与回收）。goal 目录 goal-42，分支 goal-42，worktree .worktrees/goal-42，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：管理页面（T-09.8）；外部数据的备份工具（由管理员处理）。
```

### goal-93 · T-09.14 配置、权限、工具策略与会话的管理 API

状态：等待前置（T-07.6、T-08.8、T-08.11、T-08.12、T-09.6、T-09.12、T-11.1）。2026-11-04 按前后端拆分新增，内容取自原 goal-44（T-09.9）语句中的管理 API；页面留在 goal-44。原语句的明文查看接口写成 GET 并在请求体中带 token，浏览器的 fetch 不允许 GET 带请求体，这里改为 POST。同日按 D-77～D-85 修订：Core 设置经 goal-97 的 settings 写回 `core_config.json`，Core Admin 列表同时经 goal-13 的 `policy.SetAdmins` 生效；发信许可改为「全部实例」或指定实例（goal-99）；删除插件页面票据（D-84，页面经 WebUI 的桥访问插件路由）；开发组件的配置写回 `config.json` 并重启组件；Adapter 实例的增删改写回 `core_config.json` 并按 goal-101 的 instance_schema.json 校验。语句中 goal-13、goal-26、goal-29、goal-31、goal-36、goal-39～goal-41、goal-53～goal-56、goal-84、goal-87 提供的符号（policy 的覆盖与 Admin 方法、webhost 的票据、config.Service、secret.Store 与 Redact、adminauth.Reauthenticate、管理服务与 reload、i18n 的默认 locale、tool.Catalog、conversation.Service、spec/ADMIN-API.md、router.New 的 prefixes）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-93（切片 T-09.14：配置、权限、工具策略与会话的管理 API）。goal 目录 goal-93，分支 goal-93，worktree .worktrees/goal-93，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在管理服务中补齐 Admin API：插件配置（raw JSON、Merge Patch、base_generation 与 secret 语义，Pending 经 reload 激活）、命令权限与启用、发信许可（allowed_adapter_instances）、Core 设置（Core Admin 列表、默认 locale、命令前缀、全局禁止表）、工具调用策略、Adapter 实例（按 instance_schema.json 校验）、Agent 会话的列出与清除，以及 secret 明文查看的再认证与审计；Core 设置与 Adapter 实例写回 core_config.json。接口追加到 spec/ADMIN-API.md；页面由 WebUI 章节的 T-09.9 实现。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §5、§2.3、§8（allowed_adapter_instances）、§13、§7.3、§5.4；DECISIONS D-45、D-62；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56 的管理 API 分项。

范围（只新增或修改这些）：
- internal/core/manage/：新增 config_api.go、policy_api.go、settings_api.go、tool_api.go、conversation_api.go、pluginpage_api.go；在管理服务的 HTTP 路由中追加注册
- spec/ADMIN-API.md：追加本切片的接口（本切片获授权修改该文件）
- 上述包的测试文件

接口（HTTP，全部在 /api/v1 下、经 adminauth.Require；JSON；错误格式与 goal-39 一致）：
- GET /plugins/{id}/config：{active_generation, active: Redact 后的快照, pending: {base_generation, config: Redact 后}|null, schema, i18n}；schema 是当前安装包中 config_schema.json 的规范内容（D-76，没有时为 null），i18n 是包内 i18n/*.json 的 config 根节点，按语言标签给出，供表单显示文案
- PUT /plugins/{id}/config/pending：{base_generation, config}（完整替换）；PATCH /plugins/{id}/config/pending：{base_generation, patch}（RFC 7396）；DELETE 丢弃 Pending；POST /plugins/{id}/config/import 导入 config.json；secret 字段的新值在写入 Pending 前经 secret.Extract 存为引用，省略或未修改的 secret 保持原引用；base 不是当前 Active 时 409 并返回 FAILED_PRECONDITION。
- POST /plugins/{id}/reload：经 goal-40/41 激活 Pending。
- GET/PUT /plugins/{id}/policies：命令覆盖（作用域、permission、enabled），以及 Router 的遮蔽诊断。
- GET/PUT /plugins/{id}/allowed-instances：{all: bool, instances: [...]}，经 goal-99 的 AdapterInstanceAccess 与 SetAdapterInstanceAccess；缺省 all 为 true。
- GET/PUT/DELETE /adapters/{adapter_id}/instances/{instance_id}：实例配置的读取（secret 只返回是否已设置）与写回 core_config.json，按 instance_schema.json 补缺失值并校验；GET /adapters/{adapter_id}/instance-schema 供页面生成表单。
- GET/PUT/DELETE /core/admins、/core/denies，GET/PUT /core/default-locale，GET/PUT /core/command-prefixes（全局与按 Adapter 实例的覆盖）。
- GET/PUT /tools/policies：工具策略（缺省与覆盖）。
- GET /plugins/{id}/conversations、DELETE /plugins/{id}/conversations/{key}。
- POST /secrets/{plugin}/{path}/reveal：请求体带当前管理 token 再认证（adminauth.Reauthenticate），通过后返回明文并以 INFO 审计记录（不含明文）。

行为：
- 配置：invisible 字段在读取中照常返回，是否展示由页面决定；PATCH 不会因为未提交 invisible 字段而删除它。
- 命令权限：有效权限与启用状态按 policy.CheckCommand 的同一规则计算后返回，覆盖分插件、命令组、路由三层；同时返回遮蔽关系。
- Core 设置：Core Admin 列表、默认 locale、command_prefixes 与按实例的覆盖都经 goal-97 的 settings.Update 写回 core_config.json，保留原有的 ${...} 写法；写回成功后 Core Admin 列表调用 policy.SetAdmins，命令前缀在下一次派发生效，默认 locale 在之后建立的 Session 生效。alonabot admin add|remove|list 使用同一服务。全局禁止表仍在状态库。
- 开发组件（dev 来源）：PUT/PATCH 配置时校验通过后写回 <data_dir>/plugins/<id>/config.json 并重启该组件，不建 Pending（CONTRACTS §15.4）。
- 会话：只返回摘要（键、条目数、最后时间），不返回条目内容。
- 所有修改请求经 CSRF 检查；修改后的策略在下一次派发时生效，不需要重启插件。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/admin_api_test.go（//go:build unix）：经管理 API 完成：PATCH 配置（未修改的 secret 保持原引用、新 secret 生成新引用、invisible 不被删除）、陈旧 base 409、reload 后 Active 改变；设置命令覆盖后下一条消息按新权限处理；把发信许可收紧为其他实例后插件后台 Send 被拒、恢复为全部实例后成功；Core 设置写回后 core_config.json 中的 ${...} 原写法保留；设置默认 locale 后新 Session 使用；工具策略改为 deny 后 ListTools 不含它；清除会话；开发组件的配置写回 config.json 并重启；reveal 需要再认证。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：各端点的错误格式与 409；读取结果中没有 secret 明文。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：页面（T-09.9）；日志 API（T-09.10）；故障注入与验收汇总（T-09.11）。
```

## 第 18 波

### goal-45 · T-09.11 正式流程故障注入与 T-09 验收

状态：等待前置（T-09.3、T-09.7、T-09.9、T-09.10、T-09.12、T-09.15）。语句中 goal-38～goal-44、goal-66、goal-84、goal-92～goal-94 提供的命令与 API 只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-45（切片 T-09.11：正式流程故障注入与 T-09 验收）。goal 目录 goal-45，分支 goal-45，worktree .worktrees/goal-45，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：五平台运行与发布（T-10）；新的产品功能（只做验收与必要的缺陷修复，修复需在 PR 中逐条说明）。
```
