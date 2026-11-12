# goal 语句：核心

本文件收录「核心」章节的 goal 语句，切片表在[核心实施计划](../../spec/plan/core.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-65 | T-02.1 | 日志基线与日志文件 | — | 已完成（PR #13） |
| goal-86 | T-02.3 | 状态库改用 GORM | — | 已完成（PR #63） |
| goal-103 | T-02.4 | 移除绑定 SDK API 的提交测试 | — | 已完成（PR #99） |
| goal-23 | T-07.1 | 调度触发计算 | — | 已完成（PR #17） |
| goal-29 | T-08.2 | 配置 JSON 语义与 Pending | — | 已完成（PR #23） |
| goal-37 | T-09.2 | 安装包预检与安全解压 | — | 已完成（PR #19） |
| goal-76 | T-02.2 | 调优参数 | T-02.1 | 已完成（PR #28） |
| goal-2 | T-03.1 | 首个可用 SDK 切片 | T-02.1、T-02.2 | 已完成（PR #29） |
| goal-3 | T-03.2 | Target 与发信许可 | T-03.1 | 已完成（PR #31） |
| goal-4 | T-03.3 | Adapter 进程、sdk/adapter 与上下线事件 | T-03.1 | 已完成（PR #32） |
| goal-6 | T-03.5 | 媒体上传与配额 | T-03.1 | 已完成（PR #38） |
| goal-10 | T-04.1 | 根插件 API 与静态注册索引 | T-03.1 | 已完成（PR #34） |
| goal-56 | T-11.3 | Agent 会话存储 | T-03.1 | 已完成（PR #35） |
| goal-99 | T-03.12 | 后台发信许可缺省为全部实例 | T-03.2 | 已完成（PR #90） |
| goal-11 | T-04.2 | 构建期生成器 | T-04.1 | 已完成（PR #42） |
| goal-12 | T-04.3 | 命令选择、文本匹配与参数绑定 | T-04.1 | 已完成（PR #47） |
| goal-19 | T-06.1 | Service/Bus 描述与 Broker 准入 | T-04.1 | 已完成（PR #45） |
| goal-25 | T-07.3 | SDK HTTP Listener 与 Core 代理 | T-04.1 | 已完成（PR #46） |
| goal-7 | T-03.6 | 入站媒体与 Grant | T-03.2、T-03.4、T-03.5 | 已完成（PR #55） |
| goal-21 | T-06.3 | 非持久 Event Bus | T-06.1 | 已完成（PR #56） |
| goal-26 | T-07.4 | 两个 Host 与插件票据 | T-07.3 | 已完成（PR #53） |
| goal-28 | T-08.1 | 配置字段描述与构建检查 | T-04.2 | 已完成（PR #54） |
| goal-8 | T-03.8 | 按钮、extension 片段与依赖解析 | T-03.2、T-03.7 | 已完成（PR #136） |
| goal-51 | T-03.9 | 消息编辑 | T-03.7 | 已完成（PR #118） |
| goal-87 | T-04.6 | 权限、扩展事件与命令解析的协议和 Core 侧 | T-03.3、T-03.6、T-04.3 | 已完成（PR #78） |
| goal-30 | T-08.3 | SDK 配置准备顺序 | T-08.1、T-08.2 | 已完成（PR #73） |
| goal-31 | T-08.4 | 不可变 Secret Store | T-02.3、T-08.1、T-08.2 | 已完成（PR #74） |
| goal-90 | T-08.10 | 配置标签统一 | T-08.1 | 已完成（PR #77） |
| goal-36 | T-09.1 | 首次管理认证 | T-02.3、T-07.4 | 已完成（PR #72） |
| goal-13 | T-04.4 | 持久化命令策略 | T-03.2、T-04.1、T-04.6 | 已完成（PR #88） |
| goal-88 | T-04.7 | SDK 链式注册 | T-04.2、T-04.6、T-06.1、T-06.3、T-08.1 | 已完成（PR #89） |
| goal-97 | T-07.6 | Core 设置与首次启动 | T-07.3、T-09.1 | 已完成（PR #100） |
| goal-32 | T-08.5 | 受管 SQLite 连接与迁移 | T-08.3 | 已完成（PR #79） |
| goal-53 | T-08.8 | 用户可见文案目录 | T-04.1、T-04.3、T-04.6、T-08.1、T-08.10 | 已完成（PR #102） |
| goal-100 | T-09.16 | manifest 拆分为作者信息与构建信息 | T-03.8、T-09.2 | 已完成（PR #140） |
| goal-96 | T-04.9 | 删除构建期标记 | T-04.7 | 已完成（PR #110） |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 已完成（PR #67） |
| goal-98 | T-07.7 | 插件页面的沙箱隔离 | T-07.4、T-07.6、T-09.1 | 已完成（PR #116） |
| goal-33 | T-08.6 | SQLite 候选与资源交接 | T-08.4、T-08.5 | 已完成（PR #85） |
| goal-95 | T-08.11 | 配置改由 config_schema.json 声明 | T-04.7、T-08.3、T-08.4、T-08.10 | 已完成（PR #103） |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-04.7、T-06.1 | 已完成（PR #101） |
| goal-91 | T-04.8 | SDK 自有类型与事件常量 | T-03.9、T-03.24、T-04.7、T-04.9 | 已完成（PR #141） |
| goal-123 | T-08.13 | 文案函数只返回字符串 | T-04.9、T-08.8 | 已完成（PR #137） |
| goal-130 | T-08.14 | 配置项的选项来源 | T-04.12、T-06.1、T-08.11 | 等待前置 |
| goal-24 | T-07.2 | schedule 注册与运行槽 | T-05.1、T-07.1 | 已完成（PR #107） |
| goal-35 | T-08.9 | configdb 示例与 T-08 验收 | T-08.6、T-08.7、T-08.8、T-08.11 | 已完成（PR #117） |
| goal-101 | T-08.12 | Adapter 实例配置与 Core 设置的版本迁移 | T-07.6、T-08.11 | 已完成（PR #114） |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2、T-04.9、T-08.8、T-09.16 | 待派发 |
| goal-104 | T-03.13 | 媒体按来源传递 | T-03.6、T-04.8、T-07.7 | 待派发 |
| goal-105 | T-03.14 | 新增片段与字段（含会话类型） | T-03.8、T-04.8 | 待派发 |
| goal-106 | T-03.15 | 平台动作 | T-03.8、T-04.8 | 待派发 |
| goal-107 | T-03.16 | 撤回 | T-03.9、T-04.8 | 待派发 |
| goal-108 | T-03.17 | 命令菜单同步 | T-04.6、T-04.7、T-04.8、T-08.8 | 待派发 |
| goal-109 | T-03.18 | 流式输出辅助函数 | T-03.9、T-04.8 | 待派发 |
| goal-110 | T-03.19 | Adapter 的 HTTP 路由 | T-03.3、T-04.8、T-07.7 | 待派发 |
| goal-121 | T-03.24 | 插件给出的目标由 Core 补齐 | T-03.6、T-07.6 | 已完成（PR #135） |
| goal-131 | T-03.25 | 聊天目录 | T-03.3、T-08.14、T-09.4 | 等待前置 |
| goal-122 | T-04.10 | 管理员与封禁按 Adapter 登记 | T-04.4、T-07.6 | 已完成（PR #133） |
| goal-125 | T-04.12 | 插件对象、结构体参数与子插件 | T-04.8 | 待派发 |
| goal-16 | T-05.2 | Rule、HandlerOutcome 与传播 | T-05.1、T-04.7、T-04.8 | 待派发 |
| goal-27 | T-07.5 | 调度与 HTTP 启停集成及验收 | T-07.2、T-07.4 | 已完成（PR #119） |
| goal-39 | T-09.4 | 管理服务、安装记录与 serve | T-03.8、T-07.6、T-08.4、T-08.8、T-09.1、T-09.2、T-13.1 | 等待前置 |
| goal-68 | T-13.2 | 开发组件的配置与数据库 | T-07.6、T-08.5、T-08.11、T-13.1 | 等待前置 |
| goal-69 | T-13.3 | 开发运行与热部署 | T-07.6、T-09.16、T-13.1 | 等待前置 |
| goal-9 | T-03.10 | Telegram 接入开发入口与故障矩阵 | T-03.6、T-03.8、T-03.9、T-07.6、T-13.3 | 等待前置 |
| goal-17 | T-05.3 | 三阶段 Hook 与 send_forbidden | T-05.2 | 等待前置 |
| goal-52 | T-05.4 | 延续调用与多轮等待 | T-05.2 | 等待前置 |
| goal-124 | T-04.11 | 注册写法统一 | T-04.8、T-04.12、T-07.5 | 等待前置 |
| goal-126 | T-04.13 | Service 与 Event Bus 的带类型声明 | T-04.12 | 等待前置 |
| goal-129 | T-04.14 | 命令名覆盖的 Core 侧 | T-03.17、T-04.10 | 等待前置 |
| goal-127 | T-07.8 | 动态任务与调度持久化的协议和 Core 侧 | T-07.5、T-09.4 | 等待前置 |
| goal-84 | T-09.12 | WebUI 组件与组件图标 | T-07.7、T-09.1、T-09.4、T-13.1 | 等待前置 |
| goal-81 | T-16.1 | 插件测试套件基础 | T-05.2、T-13.2 | 等待前置 |
| goal-14 | T-04.5 | 示例升级与 T-04 验收 | T-03.10、T-04.2、T-04.3、T-04.4、T-04.7、T-04.8、T-04.9、T-04.11、T-04.12、T-04.13 | 等待前置 |
| goal-18 | T-05.5 | 多进程事件流水线验收 | T-03.8、T-05.3、T-05.4 | 等待前置 |
| goal-20 | T-06.2 | 父调用链、取消与背压 | T-03.6、T-05.3、T-06.1 | 等待前置 |
| goal-66 | T-09.10 | 日志管理 | T-02.1、T-03.1、T-07.6、T-09.4、T-09.12 | 等待前置 |
| goal-92 | T-09.13 | 管理页所需的会话与上传接口 | T-09.1、T-09.4、T-09.12 | 等待前置 |
| goal-70 | T-13.4 | 开发运行示例与验收 | T-03.10、T-13.2、T-13.3 | 等待前置 |
| goal-128 | T-07.9 | 动态任务与暂停恢复的 SDK | T-04.12、T-07.8 | 等待前置 |
| goal-22 | T-06.4 | Session 清理与 Service 端到端 | T-06.2、T-06.3 | 等待前置 |
| goal-82 | T-16.2 | 按钮、多轮等待、schedule 与示例测试 | T-03.8、T-05.4、T-07.2、T-07.9、T-16.1 | 等待前置 |
| goal-38 | T-09.3 | 开发 CLI 与库模式脚手架 | T-04.5、T-08.8、T-08.11、T-09.16、T-13.3 | 等待前置 |
| goal-55 | T-11.2 | 工具调用与用户确认 | T-03.8、T-06.2、T-08.8、T-11.1 | 等待前置 |
| goal-40 | T-09.5 | 候选 Journal 与准备编排 | T-09.4、T-08.6、T-06.4 | 等待前置 |
| goal-57 | T-11.4 | 工具与会话的验收 | T-11.2、T-11.3 | 等待前置 |
| goal-102 | T-16.3 | plugintest 的 LLM 替身 | T-12.8、T-16.1 | 等待前置 |
| goal-41 | T-09.6 | 唯一提交点与激活门禁 | T-07.5、T-09.5 | 等待前置 |
| goal-42 | T-09.7 | 崩溃恢复、回退与回收 | T-09.6 | 等待前置 |
| goal-93 | T-09.14 | 配置、权限、工具策略与会话的管理 API | T-03.25、T-04.14、T-07.6、T-08.8、T-08.11、T-08.12、T-08.14、T-09.6、T-09.12、T-11.1、T-11.3 | 等待前置 |
| goal-45 | T-09.11 | 正式流程故障注入与 T-09 验收 | T-09.3、T-09.7、T-09.9、T-09.10、T-09.12、T-09.15 | 等待前置 |

## 第 1 波

### goal-65 · T-02.1 日志基线与日志文件

状态：已完成（[PR #13](https://github.com/AlonaBot/AlonaBot/pull/13)，合并提交 `775f665`）。本 goal 早于 goals.md 派发，没有原始语句；下面记录合并后的实际接口，供后继切片引用。

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

状态：已完成（[PR #63](https://github.com/AlonaBot/AlonaBot/pull/63)，合并提交 `6b197d8`），证据见 TEST-MATRIX §3.2。审查期间按维护者要求修订了四轮范围（见下），PR 按 FIX-1～FIX-10 修改：业务 CRUD 不写 SQL，表名改为单数，DDL 移入 `schema.sql`，插件 ID 与会话键在入口校验非空，持久化模型嵌入公开类型。派发时：前置 T-02 已完成。已核对 `internal/core/state` 的现有实现：`open` 以 `sql.Open("sqlite", path+"?"+pragmas)` 打开单连接，已有文件先经 `preflight` 只读预检，`migrations[0]` 是唯一的初始迁移，`tx` 封装事务；包外只经公开方法访问，没有调用方拿到 `*sql.DB` 或 `*sql.Tx`，只有 `state_test.go` 直接用了 `s.db.Exec`。进行中的 PR #53～#56 与 goal-15 都不修改此包。

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

### goal-103 · T-02.4 移除绑定 SDK API 的提交测试

状态：已完成（[PR #99](https://github.com/AlonaBot/AlonaBot/pull/99)，合并提交 `892227b`）。按 D-86 删除 SDK 全部测试、supervisor 除 `supervisor_test.go` 外的测试与全部测试插件；两个保留测试文件只做派发指定的引用清理。产品代码、依赖与 `contracts/testdata` 保持原文，验证证据见 TEST-MATRIX §3.2。goal-96、goal-95、goal-54 已按删除后的代码清除旧测试迁移要求，可按共享文件规则派发。

```text
/goal 使用 $alonabot-execute 推进 goal-103（切片 T-02.4：移除绑定 SDK API 的提交测试）。goal 目录 goal-103，分支 goal-103，worktree .worktrees/goal-103，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：删除绑定 SDK API 和测试插件的提交测试。这些 API 与结构还在变动，每次改动都要同步这些测试，维护成本大于收益（D-86）。只删除，不改产品行为，不写替代测试。

依据：DECISIONS D-86；编码规范「测试提交规则」。

范围（只修改这些）：
- 删除 sdk/ 下全部 *_test.go：sdk/bus_test.go、sdk/runtimelock_test.go、sdk/generate/config_test.go、sdk/generate/generate_test.go、sdk/generate/service_test.go、sdk/gormdb/sqlite_test.go；派发时新增的也一并删除。
- 删除 internal/core/supervisor/ 下除 supervisor_test.go 以外的全部 *_test.go，包括 PR #67 带入的 dispatch_test.go。
- 删除整个 internal/core/supervisor/testdata/ 目录。
- internal/core/supervisor/supervisor_test.go：去掉 `rawPlugin` 中引用 `commandConflictFault` 的分支（`if fault == commandConflictFault { registration = conflictingCommands() }`），它们定义在被删的 index_test.go。其余用例不改。
- cmd/alonabot/e2e_test.go：从 `TestMain` 的构建表中去掉 httpplugin，删除只有它在用的 `TestHTTPPluginSettings`，以及因此不再使用的 import 与辅助函数。其余用例不改。

接口：不新增、不修改任何非测试代码的符号。

行为：产品代码零改动。`git diff --stat origin/main...HEAD` 中出现 *_test.go、testdata 以外的文件时停下来标阻塞。

测试：
- 提交：不新增。
- 本地自测（不提交，PR 中写明结果）：go vet ./...、go test -race ./...、go build ./...、go generate ./... 后无差异、go mod tidy 后无差异。

完成条件：
- `git ls-files 'sdk/*_test.go' 'sdk/**/*_test.go'` 输出为空。
- `git ls-files 'internal/core/supervisor/*_test.go'` 只有 supervisor_test.go。
- `git ls-files internal/core/supervisor/testdata` 输出为空。
- 上述本地检查全部通过，PR 正文写明删除的文件数与行数。

不做：修改 spec、docs（TEST-MATRIX 与切片状态由审查者在合并前同步）；删除 contracts/testdata（goal-91 的语句要用其中的事件 fixture）；改动 internal/core/adapter/remote、cmd/alonabot 的其他用例、internal/core/state、internal/core/process、internal/wire、contracts/wirespec_test.go 等范围外的测试；把被删的测试搬到别的目录；修复不稳定测试。
```

### goal-23 · T-07.1 调度触发计算

状态：已完成（[PR #17](https://github.com/AlonaBot/AlonaBot/pull/17)，合并提交 `0271377`）。

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

状态：已完成（[PR #23](https://github.com/AlonaBot/AlonaBot/pull/23)，合并提交 `95caa52`）。后继注意：`pending_config.plugin_id` 外键引用 `install`，删除安装记录前须先删除该插件的 Pending 行。

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

状态：已完成（[PR #19](https://github.com/AlonaBot/AlonaBot/pull/19)，合并提交 `f8e4c11`）。

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

状态：已完成（[PR #28](https://github.com/AlonaBot/AlonaBot/pull/28)，合并提交 `793b581`；实现、本地验收与评论修复完成）。

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

状态：已完成（[PR #35](https://github.com/AlonaBot/AlonaBot/pull/35)，合并提交 `d0983f0`）。证据见 TEST-MATRIX §3.2。会话表按首次发布前规则并入唯一的初始迁移；Purge 供 T-09.4 的卸载 `--purge` 调用（D-127 起会话属于数据，`--purge-data` 与 `--purge` 都调用）。条目保留供后继切片查阅接口。

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

状态：已完成（[PR #90](https://github.com/AlonaBot/AlonaBot/pull/90)，合并提交 `ae57979`）。前置 T-03.2（PR #31）已合并。依据 D-83 新增。`AdapterInstanceAllowed`、`AdapterInstanceAccess`、`SetAdapterInstanceAccess` 与 `plugin_adapter_restriction` 已合入；未安装的插件 ID 返回包装了 `state.ErrNotFound` 的错误，重复实例按主键去重。验证证据见 TEST-MATRIX §3.2。

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

状态：已完成（[PR #42](https://github.com/AlonaBot/AlonaBot/pull/42)，合并提交 `62d5afa`）。`sdk/generate` 的 Scan、Options、Result、Diagnostic、Command、Event 与 `FileName`、`sdk.Generated`、`sdk.UseGenerated`、`alonabot plugin generate/check` 已合入。Scan 用 `go list -overlay` 把现有生成文件排除在依赖闭包之外。验证证据见 TEST-MATRIX §3.2。

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

状态：已完成（[PR #47](https://github.com/AlonaBot/AlonaBot/pull/47)，合并提交 `5cb1cfd`；验证见 TEST-MATRIX §3.2）。前置 T-04.1 已由 PR #34 合并；goal-10 的符号已按合入代码核对：`sdk.RouteOption`（`func(*routeOptions)`）、`sdk.Alias(path ...string)`、`sdk.Priority(p int32)`、`(*CommandRegistry).Command(name, handler, opts ...RouteOption)`、`(*EventRegistry).On(routeID, eventType, handler, opts ...RouteOption)`、`sdk.Command{Path, Aliases, Priority, Handler, Mode, Params}`、`sdk.EventRoute{RouteID, EventType, Handler, Priority}`、`router.CommandEntry`、`router.Shadow`、`(*CommandIndex).Lookup(tokens []string) (CommandEntry, int, bool)`、`(*CommandIndex).Shadows() []Shadow`（返回切片只读）、`(*Router).Commands() *CommandIndex`；router.Router、router.Invocation、adapter.Registry 沿用已合入代码，`message.New(adapters *adapter.Registry, p *policy.Policy) *Service`（PR #31 起需要 policy）。

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

状态：已完成（[PR #45](https://github.com/AlonaBot/AlonaBot/pull/45)，合并提交 `dcb4ec2`）。已合入：`internal/core/broker`（New、Call）、`(*router.Mux).Forward`、`(*router.Router).Plugin`，以及 SDK 的 Services Hook、ServiceRegistry、Access 常量、WithAccess、JSONService、ProtoService、CallJSON、CallProto、Context.Caller。审查中经维护者授权，插件 ID 规则从 `manifest.ValidateID` 移到 `validate.PluginID`，同时移走的还有 `validate.MaxPluginIDLen` 与 `validate.MarketPrefix`；manifest 不保留旧入口，语句中的 manifest.ValidateID 以此为准。验证证据见 TEST-MATRIX §3.2。

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

状态：已完成（[PR #46](https://github.com/AlonaBot/AlonaBot/pull/46)，合并提交 `1c3c912`）。`internal/core/settings`（Settings、Defaults、Load、ValidateListen、FileName）、`internal/core/httpproxy`（Authorizer、Server、New、Bind、Unbind、ServeHTTP）、`wire.HeaderProxyAuth`、`wire.CapabilityHTTPListener` 与 SDK HTTP Listener 已合入。`ValidateListen` 是审查中经维护者认可追加的导出符号，供 YAML 与 `--listen` 共用。验证证据见 TEST-MATRIX §3.2。

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

状态：已完成（[PR #55](https://github.com/AlonaBot/AlonaBot/pull/55)，合并提交 `b023130`），证据见 TEST-MATRIX §3.2。审查期间按复审评论追加 `message.New` 末尾的 `requestTimeout`（见接口 3b），Telegram 取件按 Range 分块并在每轮 offset 0 刷新平台路径。派发时：T-03.2、T-03.4、T-03.5 已合并（PR #31、#39、#38）。已核对媒体存储入口 `media.Open(dir string, lim Limits, now func() time.Time) (*Store, error)`、`AddServices`、`ReleaseSession`、`Sweep`、`Lookup(token string, c router.Caller) (File, error)`、`media.File{Path, Size, MIME, SHA256, ExpiresAt}` 与 `wire.MaxMediaChunk`；回收删除失败记 WARN，safe_message 为 `media storage failed`。Telegram 可复用 `New() adapter.Adapter`、私有 `instance`、`telegramMessage`、`messageEvent`、`botClient.call(ctx context.Context, method string, request any, result any) error`；`telegramtest.NewServer(t testing.TB, token string) *Server` 与接收模拟入口已合入。`sdk/adapter.Core` 当前含 Deliver、SaveCursor、ReportFailure，ReadMedia 由本切片按下文接口追加。

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

## 第 7 波

### goal-8 · T-03.8 按钮、extension 片段与依赖解析

状态：已完成（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`），证据见 TEST-MATRIX §3.2。审查者决定：Chain 接纳 extension 片段的结构校验、REGISTER 时解析 extension 声明（授权修改 `contracts/validate/chain.go` 与 `adapter.go`）；`alonabot run` 只启动一个开发插件，Installed 为空。审查期间按 FIX-03（空 value 的按钮省略该字段）与 FIX-04（编辑签发失败返回唯一拒绝分项）修改；extension 数值精度一项因 jsonschema-go 不支持 `json.Number` 撤回，记为 TEST-MATRIX 未闭合项。新增内部包 `sdk/internal/messagecodec`。按钮服务尚未装配，`BUTTON_TOKEN_TTL` 在装配注入前读缺省值。派发时：goal-51（T-03.9，[PR #118](https://github.com/AlonaBot/AlonaBot/pull/118)，合并提交 `1cbfb2e`）已先合并：`internal/core/message/edit.go` 的 `Service.Edit` 在 Edit Chain 校验之后直接调用 `adapter.Editor.Edit`，buttons 片段原样传给 Adapter；Telegram 的 `buildEditParams` 已复用 `renderButtons`（缺 core_token 时 INVALID_ARGUMENT）。本切片按范围在 Edit 中接入同一签发调用。2026-11-09 按 D-107 改为带类型的按钮动作 `sdk.Action[T]` 与 `sdk.OnButton`，删除 `ButtonValue`。前置 T-03.2（PR #31）与 T-03.7（[PR #108](https://github.com/AlonaBot/AlonaBot/pull/108)，合并提交 `d89e9b7`）均已合并。2026-11-09 审查者按 main `3dec9a0` 的代码核对并改写：SDK 的事件路由已改为链式注册（`EventRegistry.On(routeID, eventType) *EventBuilder`，以 `Handle` 结束，`sdk/session.go` 由 `EventRoute` 生成只含 event_type_filter 与 handler_mode 的 EventSpec），按钮路由相应改为返回 `*EventBuilder` 的 `OnButton`；`message.Service` 还没有 BeforeSend（T-05.3），签发放在 `validate.Chain` 之后、媒体租约之前；`message.New` 与 `router.Options` 以注入函数接入 buttons，不让 message、router 反向依赖 buttons；`policy.CheckCommand(ctx, pluginID, route, actor, target)`、`router.Router.Caller(pluginID)`（当前 Active 代次）与 `Router.Plugin(pluginID).Routes()` 已存在；`validate.Chain` 已要求按钮有 action_id 与 label、value 为合法 JSON；T-03.7 合入的 `Button.core_token = 4` 已由 Telegram 渲染为 callback_data（缺 token 时 INVALID_ARGUMENT），callback_query 已在 allowed_updates 中但尚未处理；`Segment` oneof 已用到 forward = 8；`AdapterDescriptor.extensions = 3`（ExtensionDeclaration）已在协议中；`cmd/alonabot run` 只启动单个开发插件（`--data-dir`、`--plugin-id`、`--entry`），没有 Router 与 Adapter 装配（归 goal-67），依赖检查的 Installed 只来自状态库。

```text
/goal 使用 $alonabot-execute 推进 goal-8（切片 T-03.8：按钮、extension 片段与依赖解析）。goal 目录 goal-8，分支 goal-8，worktree .worktrees/goal-8，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 为发送中的每个按钮签发持久化的 token（绑定插件、button.clicked 路由、代次、Target、value 与 24h 期限），Telegram 点击先 answerCallbackQuery，再由 Core 验证 token、当前 Actor 与策略后以普通请求调用正确的 Session；extension 片段按目标 Adapter 注册时声明的 Schema 在发送前校验；manifest 解析 role（plugin/adapter/agent/webui）、adapter_id、icon、WebUI 的 assets 与 Admin API 主版本范围、adapter_dependencies 与 plugin_dependencies，开发入口启用插件时检查必需依赖。

依据：PLAN T-03 步骤 7；CONTRACTS §8（extension、依赖与按钮段）、§9（manifest）、§7.5（图标与 WebUI）；MESSAGE-MODEL §1.1、§6.3；DECISIONS D-26、D-43、D-59、D-62、D-63、D-107；TEST-MATRIX V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项、V-105。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto：Segment 的 oneof 追加 ExtensionSegment extension = 9。随后 go generate ./...（本切片获授权只追加这一分支与消息）
- contracts/wirespec_test.go：追加编号检查
- contracts/manifest/：manifest.go 增加 role、adapter_id、icon、assets、admin_api_min_major、admin_api_max_major、adapter_dependencies、plugin_dependencies 字段与按角色的校验；新增 deps.go（依赖检查）
- contracts/validate/register.go：button.clicked 事件路由的 action 校验
- internal/core/state/schema.sql：在唯一的初始迁移末尾加入新表；新增 internal/core/state/buttons.go
- 新增 internal/core/buttons/
- internal/core/message/：New 增加按钮签发的注入函数，Send 在结构校验后签发按钮并校验 extension（只加调用点）；edit.go 的 Edit 在 Edit Chain 校验之后、调用 Editor 之前用同一个注入函数签发按钮；同步全部 message.New 调用方
- internal/core/adapter/：新增 extension.go（可选接口 ExtensionChecker）；remote 在已有的 declarations 中缓存 extension 声明并实现它；memory 增加可配置的 extension 声明并实现它
- internal/core/router/：Options 增加 Buttons 解析函数，button.clicked 改由它解析后派发（只加调用点）
- sdk/：新增 buttons.go（Action、ButtonAction、OnButton、ButtonBuilder），session.go 为按钮路由写入 EventSpec 的 action_id 与 action_schema_version，dispatch.go 把 value 解码后交给 Handler；sdk/message 增加 Buttons 与 Extension 片段构造
- adapters/telegram/：callback_query 的处理与 answerCallbackQuery；extension 声明为空
- cmd/alonabot/：run 在启动插件前读取 --entry 所在目录的 plugin.yaml（存在时）并执行依赖检查
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
   func New(store *state.Store, now func() time.Time, current func(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool)) *Service // current 返回插件当前 Active 代次与路由快照，装配处用 Router.Caller 与 Router.Plugin 实现
   func (s *Service) Issue(ctx context.Context, caller router.Caller, target *pb.Target, chain *pb.MessageChain) error // 为 chain 中每个按钮写入 core_token 并持久化映射；失败时返回 *wire.CallError
   func (s *Service) Resolve(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error) // 校验 button.clicked 事件并改写 payload 为插件注册的 action 与 value
   func (s *Service) Purge(ctx context.Context, pluginID string) error // 卸载时调用
4. internal/core/message 与 internal/core/router（注入点）
   message.New(adapters, p, lease, issue, requestTimeout) // issue 为 buttons.Service.Issue；nil 时不签发，只用于不涉及按钮的测试
   adapter.ExtensionChecker interface { CheckExtension(name string, version uint32, payloadJSON string) *pb.ErrorInfo } // 实例可选实现；未声明为 UNSUPPORTED，不符合 Schema 为 INVALID_ARGUMENT
   router.Options 增加 Buttons func(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error) // nil 时 button.clicked 一律丢弃并以 INFO 记录
5. sdk
   type ButtonAction[T any] struct{ ... } // 按钮动作：动作名与 value 的 Go 类型（D-107）
   func Action[T any](name string) ButtonAction[T] // name 即 action_id；版本取 T 的 SchemaVersion() uint32（值接收者），没有该方法时为 1
   func (a ButtonAction[T]) Button(label string, value T) message.Button // value 用 encoding/json 编码为 Button.Value
   func OnButton[T any](r *EventRegistry, a ButtonAction[T]) *ButtonBuilder[T] // 事件类型为 button.clicked，route_id 为 "button." 加动作名
   type ButtonBuilder[T any] struct{ ... } // 可链的方法与 EventBuilder 相同（Priority、Permission 等），返回 *ButtonBuilder[T]
   func (b *ButtonBuilder[T]) Handle(h func(*Context, T) error) // 链的终点；漏写时注册失败
   sdk/message：func Buttons(rows ...[]Button) Segment；type Button struct { ActionID, Label, Value string }（另有不导出的编码错误，由 ButtonAction.Button 设置）；func Extension(adapterID, name string, version uint32, payload any) Segment // payload 用 encoding/json 编码；编码失败的片段由 sdk 的 Send 在发出请求前返回 INVALID_ARGUMENT

行为：
- 按钮签发（message.Service 在 validate.Chain 之后、媒体租约与实例 Send 之前；T-05.3 加入 BeforeSend 后仍在其之前）：插件给出的 core_token 一律清除；每个按钮的 action_id 必须对应发送插件当前注册的 button.clicked 路由（EventSpec.action_id），否则 INVALID_ARGUMENT。token 为 32 字节随机值的无 padding base64url（43 字节），持久化 (token, plugin_id, route_id, action_id, action_schema_version, generation_id, target 的 instance/chat/thread, value, expires_at=签发起 BUTTON_TOKEN_TTL)。value 不超过 1024 字节（命名常量）。持久化成功后才把 token 写入 core_token 并继续发送。
- SDK 的按钮动作（D-107）：Collect 检查动作名（与 action_id 的规则相同）；route_id 为 `button.` 加动作名，同插件内重复时注册失败；`button.` 是保留前缀，文本匹配与事件路由不能使用。ButtonAction.Button 编码失败时，sdk 的 Send 在发出请求前返回 INVALID_ARGUMENT。点击时 SDK 把 value 解码为 T 再调用 Handler；解码失败时记 WARN（只记 route_id 与错误，不记 value），不调用 Handler，正常结束本次调用。
- Telegram 点击：callback_query 先调用 answerCallbackQuery（失败只记日志），再交付 button.clicked：Target 取 callback_query.message 的 chat 与 thread（没有 message 的 inline 回调不交付，推进游标并以 INFO 记录），Actor 取 from，payload 的 value 暂放 callback_data，origin_platform_message_id 为消息 ID，callback_query_id 原样。
- Core 解析（Router 在派发 button.clicked 前调用 Resolve）：token 不存在、过期、generation 已不是该插件的当前 Active、或 Target 与签发时不同 → 丢弃并以 INFO 记录，不派发、不回复、不进入 Raw 或文本匹配。有效时把 payload 改写为签发时的 action_id、action_schema_version 与 value，只派发给该插件的该路由；派发前执行 T-04.4 的 CheckCommand 等价检查（全局禁止与路由启用、权限），每次点击都按本次 Actor 重新检查；同一 token 可被多人、多次点击。token 不带任何角色。
- extension：remote 的 `compileDeclarations` 已缓存权限闭包与扩展事件的已解析 Schema，同样在注册时解析每个 ExtensionDeclaration（jsonschema-go 的 Resolve，Loader 为 nil；解析失败按注册失败），实例实现 ExtensionChecker。发送前对每个 extension 片段：目标实例的 AdapterID() 必须等于片段的 adapter_id，实例必须实现 ExtensionChecker 且 (name, version) 已声明，payload_json 必须通过该声明的 Schema；否则 UNSUPPORTED（Adapter 不符、未实现或未声明）或 INVALID_ARGUMENT（不符合 Schema），在媒体租约与调用 Adapter 之前返回。Telegram 本切片不声明 extension，所有 extension 片段发往 Telegram 都是 UNSUPPORTED。
- manifest：按 CONTRACTS §9.1 校验各角色：role 为 adapter 时必须有 adapter_id，且不能有 adapter_dependencies、plugin_dependencies、migration_mode 或 prepare_effects；agent 与 plugin 的字段规则相同；webui 必须有 assets 与 Admin API 主版本范围、os/arch 为 any，不能有 entrypoint、协议主版本、required_capabilities、config_version 与依赖；市场 ID 的前缀与 role 一致（alonabot_plugin_、alonabot_adapter_、alonabot_agent_、alonabot_webui_）；依赖项的 ID 格式、版本 SemVer、重复项都要校验；必需的 plugin_dependencies 不能直接依赖自己。环的检测需要全部安装记录，由 T-09.4 完成。
- CheckDependencies：必需的 Adapter 依赖要求该 adapter_id 已安装且声明了全部所需 extension（版本相等）；必需的插件依赖要求已启用且版本 ≥ MinVersion；可选依赖缺失也报告（Required=false），调用方只记诊断。
- cmd/alonabot run：插件目录下有 plugin.yaml 时解析并检查依赖（Installed 来自状态库中已启用的安装记录与当前运行的 Adapter 描述），有必需项问题时以 exitFailure 退出并列出全部缺失项。

测试（按编码规范「测试提交规则」）：
- 提交：
  - adapters/telegram/callback_test.go：telegramtest 推入 callback_query：先有 answerCallbackQuery 调用，再交付 button.clicked；没有 message 的 inline 回调不交付但推进游标。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：manifest 新字段的每种非法值；CheckDependencies 的判定表；Action 的版本与动作名检查、编码失败的按钮、解码失败的点击、`button.` 前缀的注册冲突；token 的编码长度与唯一性；value 长度边界；Purge；Resolve 对 generation 的判断；internal/core/supervisor/buttons_local_test.go（//go:build unix）：memory Adapter 与测试插件（以 `vote := sdk.Action[voteValue]("vote")` 与 `sdk.OnButton(r, vote).Handle(...)` 注册，收到 /poll 时发送 `vote.Button` 生成的两个按钮），用真实状态库装配：发送后 memory 收到的按钮都有 43 字节的 core_token，状态库有映射；插件伪造的 core_token 被替换；注入以该 token 为 callback_data 的 button.clicked：插件的 Handler 被调用，收到的值与发送时相同；两个不同用户点击各调用一次；未知 token、过期 token（注入时钟）、另一个 chat 中的点击都不调用插件，也不进入 Raw；按钮指向未注册的 action 时发送被拒，memory 没有收到；插件产生新的 Active 代次后，旧 token 失效；extension 片段发往不声明它的 Adapter 返回 UNSUPPORTED；发往声明了它的 memory Adapter 时 payload 不符合 Schema 返回 INVALID_ARGUMENT，符合时发出；contracts/manifest 对四种 role、icon、WebUI 字段与依赖字段的解析。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/... ./cmd/...；go build ./...。

不做：正式管理入口的依赖检查、确认与环检测（T-09.4）；消息编辑（T-03.9）；BeforeSend/AfterSend Hook（T-05.3）；Telegram 的 extension（之后按需由 Adapter 声明）。
```

### goal-51 · T-03.9 消息编辑

状态：已完成（[PR #118](https://github.com/AlonaBot/AlonaBot/pull/118)，合并提交 `1cbfb2e`），证据见 TEST-MATRIX §3.2。合入时 goal-8 尚未合并，`core.edit` 把 buttons 片段原样交给 Adapter，签发由 goal-8 补上。派发时：前置 T-03.7（[PR #108](https://github.com/AlonaBot/AlonaBot/pull/108)，合并提交 `d89e9b7`）已合并。2026-11-09 审查者按 main `3dec9a0` 的代码核对并改写：`internal/core/adapter.Instance`（AdapterID、InstanceID、AccountID、Capabilities、Start、Stop、Send）与 `sdk/adapter.Instance`（Account、Capabilities、Send、Stop）已存在，Editor 都作为可选接口；`internal/wire/services.go` 现有 `ServiceAdapterSend = "adapter.send"`，RunAdapter 在 `sdk/adapter_run.go` 分派；`message.Service.Send` 的目标准入顺序为默认目标、NO_DEFAULT_TARGET、`policy.CheckSend`、`validate.Chain`、实例与账号绑定；Telegram 的能力表为 `platformCapabilities()`，由 `Capabilities()` 与 `Descriptor()` 共用；`telegramtest` 已有 `Requests()`、`FailMethod`、`ZeroMessageID`、`Drop` 与 `Calls(method)`，已支持 editMessageReplyMarkup，尚无 editMessageText。goal-8（T-03.8）与本切片可以并行：两者都修改 `internal/core/message`、`adapters/telegram` 与 `telegramtest`；按钮的 token 签发归 goal-8，后合并的一方按先合并的代码变基，并让 core.edit 复用 Send 的签发调用。

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
- adapters/telegram/：新增 edit.go；platformCapabilities() 增加 message.edit
- adapters/telegram/telegramtest/：server.go 的分派与 send.go 的请求记录加入 editMessageText，现有故障注入同样适用
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
- core.edit（Core）：确定 Target 与准入的步骤与 core.send 完全相同（默认目标、NO_DEFAULT_TARGET、policy.CheckSend、实例与账号绑定）。platform_message_id 为空 → INVALID_ARGUMENT。Chain 必须恰好一个 text 片段、至多一个 buttons 片段，不能有其他片段，否则 INVALID_ARGUMENT。目标实例的 Capability 没有 supported 的 message.edit，或实例不实现 Editor → UNSUPPORTED。按钮的 token 签发与 Send 相同：goal-8 已合并时复用它的签发注入；尚未合并时 buttons 片段原样交给 Adapter（Telegram 缺 core_token 时 INVALID_ARGUMENT），由后合并的一方补上。然后调用 Editor.Edit，原样返回 SendResult 与 ErrorInfo。不自动重试，失败时不改为发送新消息。
- Hook：BeforeSend/AfterSend 的输入标明操作为 edit 与被编辑的 platform_message_id，由 T-05.3 接入；本切片在 core.edit 中留出与 Send 相同的位置，以一行 shortcut 注释写明。
- Telegram：editMessageText（chat_id、message_id、message_thread_id 不传）；text 的格式与实体规则同发送；buttons 转为 reply_markup。平台返回 "message is not modified" 或消息不可编辑时为 REJECTED（ErrorInfo 带平台描述）。只编辑文本消息，不编辑媒体说明（目标为媒体消息时平台拒绝，同样 REJECTED）。成功时 SendResult 为 ACCEPTED，platform_message_ids 为被编辑消息的 ID。
- 能力：Telegram 的 platformCapabilities() 增加 message.edit（version 1，supported=true），实例与 Descriptor 同时声明。

测试（按编码规范「测试提交规则」）：
- 提交：
  - adapters/telegram/edit_test.go：telegramtest 驱动：编辑成功的参数正确；"message is not modified" 为 REJECTED；带 buttons 时 reply_markup 正确；Requests() 中只有 editMessageText，没有发送类方法。
  - contracts/wirespec_test.go：新消息的编号检查。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Edit Chain 的每种非法组合；后台 Edit 缺少 to；adapter.edit 在实例不实现 Editor 时的 UNSUPPORTED；Telegram 编辑的实体 UTF-16 偏移；internal/core/supervisor/edit_local_test.go（//go:build unix）：memory Adapter 实现 Editor 并声明 message.edit；测试插件收到消息后发送一条再编辑它：memory 先收到发送、再收到编辑，ID 一致；Chain 含图片时被拒；memory 不声明 message.edit 时返回 UNSUPPORTED；编辑其他会话（没有许可）返回 PERMISSION_DENIED。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。

不做：出站 Hook 的执行（T-05.3）；编辑媒体或说明；message.edited 入站事件（已在 T-03.4）。
```

### goal-87 · T-04.6 权限、扩展事件与命令解析的协议和 Core 侧

状态：已完成（[PR #78](https://github.com/AlonaBot/AlonaBot/pull/78)，合并提交 `0a320c5`）。审查中经维护者批准扩大范围两处（SCOPE-01、SCOPE-02，见下方范围的最后两项与接口 5）。

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

状态：已完成（[PR #73](https://github.com/AlonaBot/AlonaBot/pull/73)，合并提交 `225f132`），证据见 TEST-MATRIX §3.2。审查期间按授权改写 `sdk/generate/config_test.go` 的候选断言为规范 JSON，并在 `sdk/config` 包文档加入 Prepare 入口；`sdk/session.go` 另把 `PhaseLimits.config_ms` 缺失并入 Bootstrap 预算缺失的 PROTOCOL_ERROR。派发时：前置 T-08.1（PR #54）、T-08.2 已合并。goal-28 的符号已按合入代码核对：
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

状态：已完成（[PR #74](https://github.com/AlonaBot/AlonaBot/pull/74)，合并提交 `37e6cac`），证据见 TEST-MATRIX §3.2。审查中发现原语句「状态库文件已是 0600」的前提不成立（data_dir 为 0700，文件按进程 umask 创建）；维护者在 PR 中确认扩围，范围、行为与测试已按合入代码改写。派发时：前置 T-02.3（PR #63，合并提交 `6b197d8`）、T-08.1（PR #54）、T-08.2 均已合并。

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

状态：已完成（[PR #77](https://github.com/AlonaBot/AlonaBot/pull/77)，合并提交 `f34b1e8`），证据见 TEST-MATRIX §3.2。实现按语句完成，`internal/core/supervisor` 中配置准备与 Secret Store 的测试插件一并改用新标签。派发时：前置 T-08.1 已由 PR #54 合并（合并提交 `021e972`）；依据 D-74 新增。

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

状态：已完成（[PR #72](https://github.com/AlonaBot/AlonaBot/pull/72)，合并提交 `1ba7e5b`），证据见 TEST-MATRIX §3.2。审查期间按裁决改 `New` 的签名：原签名拿不到 Origin 校验所需的 admin_base_url，改为接收该 URL，并由 Base URL 的 scheme 推导 secure（见接口 2）。派发时：前置 T-02.3（PR #63，合并提交 `6b197d8`）、T-07.4（PR #53）均已合并。已按 #63 的模型与命名约定核对表结构：
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

状态：已完成（[PR #88](https://github.com/AlonaBot/AlonaBot/pull/88)，合并提交 `c8dd516`）。`policy` 的覆盖、Admin 与禁止名单方法、`router.CommandPolicy` 与 `router.New` 的第六个参数已合入；生产装配尚未调用，Core Admin 列表由 goal-39 接入，接口中 SetAdmins 注释的「T-07.6」以此为准。验证证据见 TEST-MATRIX §3.2。

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

状态：已完成（[PR #89](https://github.com/AlonaBot/AlonaBot/pull/89)，合并提交 `3e37cd6`）。链式注册、`Permission`、`Generated.ServiceSchemas` 与生成器的 Service Schema 已合入，示例与测试插件已迁移。语句按维护者对审查意见的裁决修订了三处：类型约束由生成器检查、切片与 map 接受 null、校验详情不出 SDK。验证证据见 TEST-MATRIX §3.2。

派发时的核对：前置 T-04.2、T-04.6（PR #78，合并提交 `0a320c5`）、T-06.1、T-06.3、T-08.1 均已合并。依据 D-70 新增；2026-11-04 经维护者同意，把原切片拆为两部分：本切片只做链式注册，SDK 自有类型、事件常量与 `sdk.AdapterEvent` 移到 T-04.8（goal-91）。符号已按合入代码核对：
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
   JSON 类型的 schema_id 为 Go 类型的「包路径.类型名」（reflect 的 PkgPath() + "." + Name()）；版本取类型的 SchemaVersion() uint32 方法（值接收者），没有该方法时为 1。In、Out、T 必须是包级的具名非泛型类型：生成器对 sdk.JSON、sdk.Subscribe、sdk.CallJSON、sdk.Publish 的实例统一检查，不满足时给出带位置的诊断；运行期只检查反射能判定的部分（Collect 报错，CallJSON 与 Publish 返回 INVALID_ARGUMENT），不判断词法作用域。
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
  - Schema 只依据 json 标签与 Go 类型，不读取 config 标签，支持的类型与配置 Schema 相同；不支持的类型给出带位置的诊断。切片生成 `"type":["array","null"]`，map 生成 `"type":["object","null"]`，与 encoding/json 对 nil 的编码一致。
  - In 或 Out 不是具名类型时报告诊断；两个不同类型得到同一 schema_id 时也报告。
  - Collect 时 Generated.ServiceSchemas 缺少某个 JSON Service 的 schema_id，报错 "service schema for <id> is not generated; run alonabot plugin generate"。
  - 调用时，SDK 先用 jsonschema-go 把输入 JSON 按 Schema 校验，失败返回 INVALID_ARGUMENT 与固定文本 `invalid service input`，不调用 Handler；Handler 返回后，输出按 Schema 校验，失败返回 INTERNAL 并记 ERROR（只记 schema_id，校验错误的文本带载荷，不写入响应和日志）。
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

状态：已完成（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`），证据见 TEST-MATRIX §3.2。审查中经审查者授权超出语句两处：导出 `adminauth.FirstTokenFile`；`Update` 以 Rename 为提交点，之后目录同步失败时仍发布新快照并返回 `settings saved but directory sync failed` 错误。派发时：前置 T-07.3（PR #46）与 T-09.1（PR #72）均已合并。依据 D-82 新增。2026-11-05 修订：`cmd/alonabot` 没有装配 Router，命令前缀与 Core Admin 列表改由 goal-39 接入 goal-67 建立的根包 Core 装配，本切片只提供读取设置的函数。

已核对现有实现（main `fb1000d`）：
- `internal/core/settings` 读取 `<data_dir>/alonabot.yaml`（`FileName`），`Settings` 有 `Listen`、`AdminBaseURL`、`PluginBaseURL`、`TrustedProxies`、`LogLevel`、`LogFormat`、`LogFile{Enabled, MaxSize, MaxTotal, MaxAgeDays}`、`Proxy`；`Load` 严格解析（未知键、重复键与类型错误带完整键名和行号），文件不存在时以 0600 创建缺省内容；`Defaults`、`ValidateListen` 供调用方使用。
- `cmd/alonabot/main.go` 用 `settings.Load` 的结果设置日志、日志文件、出站代理、`webhost.New`（`AdminBaseURL` 与 `PluginBaseURL`）与 Listener，命令行的 `--listen`、`--log-level`、`--log-format` 只覆盖本次运行。`run` 经 supervisor 启动单个插件，`cmd/alonabot` 没有 import `router` 与 `internal/core/policy`；`router.New` 与 `policy.SetAdmins` 目前只在测试中调用。
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
- cmd/alonabot/main.go：装配新的设置与首次 token；cmd/alonabot/admin_token.go：增加 show
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
   type File struct{ /* 不导出 */ } // 方法可以并发调用：Router 在派发路径上读取前缀，Update 与 Reload 同时替换当前设置
   func Load(dataDir string) (*File, error)                  // 文件不存在时以 0600 创建缺省内容；展开 ${...}，校验全部键
   func (f *File) Settings() Settings                         // 当前设置的副本
   func (f *File) CommandPrefixes(adapterInstanceID string) []string // 该实例的命令前缀：command_prefixes_by_instance 有该键时取覆盖，否则取 command_prefixes；读当前设置，Update 与 Reload 成功后下一次调用即返回新值；返回值只读，不复制。签名与 Router 的 prefixes 相同，由 goal-39 接入
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
- 文件格式：JSON 对象，键与缺省值见 CONTRACTS §15.1 的表；这一阶段 `base_url` 仍是 `admin_base_url` 与 `plugin_base_url` 两个键（缺省值与现在相同），由 T-07.7 合并。未知键、重复键、类型或取值错误时 Load 失败，错误包含完整键路径（例如 `adapters.telegram.instances.main`）与行列号。`command_prefixes` 与 `command_prefixes_by_instance` 的前缀列表为空时同样失败：Router 要求非空列表，不带前缀的命令写 `[""]`。`$version` 当前为 1，缺失时按 1 处理并在下次写回时补上；版本迁移由 T-08.12 实现，本切片遇到大于 1 的版本时报错。
- ${环境变量}：只有整个字符串值等于 `${NAME}`（NAME 为字母、数字与下划线，不以数字开头）时展开；变量不存在时 Load 失败并指出键路径。Update 在原文上修改，写回时保留原来的 `${...}`，不把展开后的值写回。
- 写回：键保持文件中的原有顺序，新键追加在所在对象末尾，缩进两个空格，末尾换行。
- 命令行参数 `--listen`、`--log-level`、`--log-format` 照旧只覆盖本次运行，不写回文件。
- 装配：main.go 用 Settings 设置日志、日志文件、代理、webhost 与 Listener。`run` 没有 Router 与 Policy，本切片不接入命令前缀与 Core Admin 列表，两者由 goal-39 接入根包的 Core 装配（执行顺序的共享文件规则）。LogSources、DefaultLocale 与 Adapters 只解析与保存，由后续切片使用（T-09.10、T-09.4、T-13.3）。
- 首次启动：Core 打开状态库后调用 FirstToken；stdin 与 stderr 都是终端时把返回的 token 只写一次到 stderr；否则写入 admin_token，并在日志中以 INFO 输出该文件的路径（不输出 token）。admin-token init 与 reset 的行为不变。

测试（按编码规范「测试提交规则」）：
- 提交：
  - cmd/alonabot 的集成测试（//go:build unix）：非终端首次启动后数据目录出现 0600 的 admin_token，`admin-token show` 输出的 token 能登录，登录后文件被删除、再次 show 报错。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：终端首次启动只向 stderr 写一次；写回中途失败时原文件不变；Update、Reload 与 CommandPrefixes 并发（-race）；Windows 上的文件权限（交叉编译检查，不在 Windows 运行）；internal/core/settings 的本地测试：缺省创建与 0600 权限；${NAME} 展开与缺失时的键路径；未知键与重复键的位置；Update 修改一个键后其他键的顺序与 ${...} 原写法不变，校验失败时文件不变；Reload 返回变化的顶层键；`$version` 为 2 时报错；CommandPrefixes 在有覆盖的实例上取覆盖、其他实例取 command_prefixes、缺省为 ["/"]，Update 与 Reload 后下一次调用返回新值；前缀列表为空时 Load 失败并给出键路径。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/settings/... ./internal/core/adminauth/... ./cmd/alonabot/...；go build ./...。

不做：把命令前缀、Core Admin 列表与默认 locale 接入 Router、Policy 和组件启动（T-09.4）；合并两个 Base URL 与删除插件 host（T-07.7）；core_config.json 与 Adapter 实例配置的版本迁移和备份（T-08.12）；Admin API 与 CLI 对这些设置的修改入口（T-09.4、T-09.14）；`alonabot run --reload` 对该文件的监视（T-13.3）；日志级别的运行时修改（T-09.10）。
```

### goal-32 · T-08.5 受管 SQLite 连接与迁移

状态：已完成（[PR #79](https://github.com/AlonaBot/AlonaBot/pull/79)，合并提交 `75c16d3`）。审查中经维护者批准扩大范围一处（FIX-3，见下方范围的最后一项）。

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

状态：已完成（[PR #102](https://github.com/AlonaBot/AlonaBot/pull/102)，合并提交 `f685c92`），证据见 TEST-MATRIX §3.2。2026-11-07 维护者在 PR #102 [裁决](https://github.com/AlonaBot/AlonaBot/pull/102#issuecomment-6707514440) Router 取默认 locale 的方式，下面语句的范围、接口第 4、6 项与 Core 行为已按裁决修订。派发时：前置 T-04.1、T-04.3、T-04.6（PR #78，合并提交 `0a320c5`）、T-08.1、T-08.10（PR #77）均已合并。2026-11-04 按 D-69～D-74 改写，符号已按合入代码核对；同日按 D-76 修订 config 根节点的格式（与 config_schema.json 同形），与 schema 的核对归 goal-95；按 D-80 改用 `i18n/` 目录并允许顶层的 `metadata`（D-78）；按 D-82 修订：默认 locale 写在 `core_config.json` 的 `default_locale`，由 goal-39 在装配处从 goal-97 的 settings 取得，本切片不建 `core_setting` 表：
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
- internal/core/host/：ActivateOptions 与 Instance 增加 Messages，Instance 增加 DefaultLocale；Bootstrap 写入 default_locale 的调用点，Prepare 保存该值的快照
- internal/core/supervisor/supervisor.go：Spec 增加 I18n 与 DefaultLocale，Launch 解析 I18n 并传给 Activate（只加这一段）
- internal/core/router/：Plugin 接口与 instancePlugin 增加 Messages 与 DefaultLocale；commandUsage 改用目录，删除其中的 shortcut 注释
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
   Instance 增加 DefaultLocale string // 本次 Session 的 Bootstrap.default_locale 快照；Prepare 存进 Prepared 的私有字段，Activate 复制，激活后只读
5. internal/core/supervisor
   Spec 增加 I18n map[string][]byte // 当前包的 i18n/*.json 原文，来自安装包或库模式的嵌入副本；nil 表示没有文案资源
   Spec 增加 DefaultLocale string // 写入 Bootstrap.default_locale；装配处由 core_config.json 的 default_locale 取得（goal-97 提供读取，goal-39 接入），空值按 "en-US"
6. internal/core/router
   Plugin 接口增加 Messages() *i18n.Catalog 与 DefaultLocale() string；instancePlugin 返回 instance.Messages 与 instance.DefaultLocale。Router 经接口读取，不做类型断言

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
  - 缺省 usage：文案键为 core.usage.*（en-US 与 zh-CN 两份），占位符为本次匹配到的前缀、命令路径、别名与参数用法（参数类型标注 :int、:float、:bool、:user、:image、:message 保持不变）。语言按 [事件 Actor 的 locale, 接收插件的 Plugin.DefaultLocale()] 选择，最后回落到 en-US，与该插件 Context.T 使用的 Bootstrap 快照一致；修改默认 locale 后，运行中的插件仍用旧值，下一个 Session 才用新值，Router 不读取当前设置。router.New、router.Options 与 supervisor.Spec 不因此改动。
  - 自定义 usage（USAGE_MODE_CUSTOM）：用 Plugin.Messages() 的 usage_key 按同一候选格式化，参数为 prefix（本次匹配到的前缀）与 command（前缀加规范路径），多余参数照常忽略。Messages 为 nil、缺少该键或格式化失败时，记 WARN（plugin_id、route_id、错误），改发缺省 usage。
- 日志、错误字符串与 safe_message 仍只用英文。

测试（按编码规范「测试提交规则」）：
- 提交：
  - contracts/wirespec_test.go：Bootstrap.default_locale = 12。
- 已有集成测试中断言英文缺省 usage 原文的地方（如 command_test.go）按 en-US 目录的文案同步，覆盖的场景不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 占位符与花括号转义；Format 的未知键与缺少参数。
  - BCP 47 规范化。
  - Core 目录的 en-US 与 zh-CN 键集合一致。
  - Collect 拒绝未知的 usage_key；Messages 为 nil 时回落到缺省 usage 并记 WARN。
  - internal/core/supervisor/i18n_local_test.go（//go:build unix），测试插件带 en-US、zh-CN、zh-TW 三个文件，收到消息时用 Context.T 回复：Actor.locale 为 zh-CN 时得到中文；zh-HK 回落到主语言 zh 中字典序最小的 zh-CN；ja 与没有 Actor 时用默认 locale；把默认 locale 设为 zh-CN 并重启插件后，ja 得到中文；参数错误时的缺省 usage：zh-CN 用户收到中文，en-US 用户收到英文；带 UsageKey 的命令参数错误时，按调用者语言发送插件的自定义 usage；Spec.I18n 有问题（如 zh-CN 多出键）时 Launch 返回 INVALID_ARGUMENT，且没有启动进程。
  - sdk/generate/i18n_local_test.go：沿用 generate_test.go 的临时 module 辅助函数。缺少 en-US、zh-CN 多出的键、占位符集合不同、非法 JSON、多余的顶层键、config 根节点中出现不允许的键或类型不对时各产生诊断；合法时 Generated.I18n 被嵌入。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：WebUI 文案、语言切换与配置字段翻译的渲染（T-08.7）；确认消息的文案（T-11.2 使用本目录）；修改默认 locale 的管理 API（T-09.14）；从安装包与库模式读取 i18n/ 并填入 Spec.I18n 的装配（T-09.5、T-13.1）；把 default_locale 填入 Spec.DefaultLocale 的装配（T-09.4）；Adapter 权限的 label_key 显示（T-09.9）。
```

### goal-100 · T-09.16 manifest 拆分为作者信息与构建信息

状态：已完成（[PR #140](https://github.com/AlonaBot/AlonaBot/pull/140)，合并提交 `896991b`），证据见 TEST-MATRIX §3.2。合入的符号：`manifest.ParseProject(data) (*Result, error)`、包内 `Parse` 要求 name 并接收 description、author、`BuildInfo` 与 `Generate(project *Manifest, build BuildInfo) ([]byte, error)`；`cmd/alonabot` 的依赖检查改用 ParseProject。未闭合：包内 manifest 缺 migration_mode 或 prepare_effects 时仍只警告（CONTRACTS §9.1 要求拒绝），归属待维护者确认。派发时：前置 T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）与 T-09.2（[PR #19](https://github.com/AlonaBot/AlonaBot/pull/19)）已合并，审查者按合入代码写出接口与测试段。依据 D-78 新增；T-03.8 已为 `contracts/manifest.Manifest` 加入 role、adapter_id、icon、assets、Admin API 版本范围与依赖字段，`CheckDependencies` 接收 `*Manifest`；`Parse` 仍只解析包内 manifest，`packages.Preflight` 与 `cmd/alonabot` 的 `checkPluginDependencies` 都调用它。现有工程声明为 `examples/plugins/configdb/plugin.yaml` 与 `web/admin/plugin.yaml`，均已含 name 与作者侧字段。

```text
/goal 使用 $alonabot-execute 推进 goal-100（切片 T-09.16：manifest 拆分为作者信息与构建信息）。goal 目录 goal-100，分支 goal-100，worktree .worktrees/goal-100，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：组件工程中的 plugin.yaml 只写作者信息；contracts/manifest 分别解析工程 manifest 与包内 manifest，并提供由工程 manifest 加构建信息生成包内 YAML 的函数。包内 manifest 沿用现有严格解析与角色校验，增加 name、description 与 author，name 必填；安装预检只接收包内 manifest。

依据：PLAN T-09 步骤 2；spec/plan/core.md 的 T-09.16；CONTRACTS §9.1、§5.4（metadata）；DECISIONS D-78；TEST-MATRIX V-01。

范围（只新增或修改这些）：
- contracts/manifest/：manifest.go、deps.go；可新增 project.go 与 build.go，分别实现工程解析与包内 YAML 生成，共用身份、角色、路径与依赖校验
- contracts/testdata/manifest/market.yaml、local_no_prefix.yaml：现有包内 manifest 样例补齐 name，继续用 Parse 校验
- internal/core/packages/preflight.go：继续用包内解析；preflight_test.go、preflight_unix_test.go：同步包内 manifest fixture，并按下文只新增一个集成测试
- cmd/alonabot/dependencies.go：run 读取工程目录 plugin.yaml 的依赖检查改用工程解析；cmd/alonabot/ 中已有测试如构造 plugin.yaml，按其实际用途同步
- examples/plugins/configdb/plugin.yaml、web/admin/plugin.yaml：作为工程 manifest 核对，保留作者字段；上述范围内测试源码或字符串中的工程声明用工程解析，归档内声明补齐 name 并继续用包内解析
- 上述包的 *_local_test.go（本地自测）
允许的新跨模块边：无；复用 packages 与 cmd/alonabot 到 contracts/manifest 的已有依赖。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. contracts/manifest
   Manifest 增加 Name、Description、Author string，YAML 键分别为 name、description、author；其余字段、Result 与 CheckDependencies(m *Manifest, in Installed) []Problem 保持。工程与包内解析共用 Manifest 与 Result，工程结果的构建字段保持未设置。
   func ParseProject(data []byte) (*Result, error) // 严格解析工程 plugin.yaml；出现构建字段时报错并指明字段
   func Parse(data []byte) (*Result, error)        // 保留为包内 manifest 解析，增加 name 必填及 description、author 的接收
   type BuildInfo struct {
       ManifestVersion int                // manifest_version，必须为 1
       ProtocolMinMajor *uint32            // protocol_min_major；进程角色必填
       ProtocolMaxMajor *uint32            // protocol_max_major；进程角色必填
       RequiredCapabilities *[]Capability // required_capabilities；进程角色必填，无能力时指向空列表
       OS, Arch string                    // os、arch；WebUI 为 any/any
       Entrypoint string                  // entrypoint；进程角色的包内相对路径
       ConfigVersion *uint32               // config_version；进程角色必填
       MigrationMode string               // migration_mode；plugin/agent 为 none、self 或 versioned
       PrepareEffects string              // prepare_effects；plugin/agent 的构建值为 none 或 managed_db
   }
   func Generate(project *Manifest, build BuildInfo) ([]byte, error) // 工程 manifest 加构建信息，返回经过 Parse 校验的稳定包内 YAML
2. 现有调用方
   packages.Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error) 保持签名，仍调用 manifest.Parse；解析失败继续包装为 *validate.Error，Code 为 ERROR_CODE_INVALID_ARGUMENT。
   cmd/alonabot 的 checkPluginDependencies(pluginDir string) error 保持签名，改用 manifest.ParseProject，再把 Result.Manifest 交给现有 CheckDependencies。

行为：
- 工程解析：必填 id、name、SemVer version、source；description、author、role（缺省 plugin）、icon 可选。按角色接收 adapter_id、adapter_dependencies、plugin_dependencies、assets、admin_api_min_major、admin_api_max_major；角色禁用字段、ID/市场前缀、相对路径、图标、依赖版本/重复/必需自依赖等校验复用现有规则。工程不要求构建字段。
- 工程出现 manifest_version、protocol_min_major、protocol_max_major、required_capabilities、os、arch、entrypoint、config_version、migration_mode 时，按字段是否出现拒绝，显式零值、空串、空列表与 null 同样报错，错误中带具体字段名。prepare_effects 只允许 plugin/agent 显式声明 external；none、managed_db、空值与其他值报错并指明 prepare_effects。Adapter 与 WebUI 沿用现有的角色字段限制。
- 两种解析都拒绝未知字段、多个 YAML 文档、缺失必填项与非法取值；警告仍经 Result.Warnings 返回。包内解析沿用现有的 manifest_version、协议范围、required_capabilities、平台、entrypoint、config_version 与按角色的严格校验，增加 name 非空校验；description 与 author 可省略。migration_mode 与 prepare_effects 沿用当前解析规则与未声明警告。
- 生成：作者字段原样保留；构建值来自 BuildInfo，工程已声明 prepare_effects: external 时保留 external，否则采用构建传入的 none 或 managed_db。按角色只输出允许的字段：Adapter 省略依赖、迁移与 prepare_effects；WebUI 输出 manifest_version、os/arch=any 与作者侧静态声明，省略进程字段、迁移与 prepare_effects。生成结果必须通过 Parse；构建信息不合法时返回完整错误。同样的工程字段与 BuildInfo 生成逐字节相同的 YAML，字段顺序固定，依赖与能力列表保持输入顺序。
- run 保留现有行为：plugin.yaml 缺失时跳过依赖检查；存在时严格解析工程文件。当前 run 只启动单个开发插件，Installed 为两个空 map；必需依赖问题合并报错，可选问题记 WARN，保持不变。
- i18n 的 metadata.name 与 metadata.description 沿用 contracts/i18n 已有的 Parse 检查与 §5.4 翻译约定；本切片保存 manifest 的英文源文案，显示与语言选择由既有文案/后继管理入口处理。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 internal/core/packages/preflight_test.go 的 TestProjectManifestPreflight，复用现有 buildArchive、newPackage 与真实临时归档/预检路径，在同一个测试中覆盖三个场景：工程 manifest 写 entrypoint 时 ParseProject 报错且包含 entrypoint；合法工程经 ParseProject 与 Generate 得到包内 YAML，装入包含普通入口文件的归档后 Preflight 成功且 Candidate.Manifest 的 name、description、author 与输入一致；包内 YAML 去掉 name 后重新构造有效发布清单与归档，Preflight 返回 ERROR_CODE_INVALID_ARGUMENT，错误包含 name。现有 manifestYAML、validEntries 及各拒绝用例只补齐新必填字段，原有断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：两种解析的未知字段/多文档、每个构建字段的显式空值与 null、必填 name；四种角色的字段组合、平台与相对路径；依赖校验与 CheckDependencies；external 的保留、none/managed_db 的合成与非法 BuildInfo；同输入两次 Generate 的 YAML 逐字节相同且 Parse 成功；两个现有工程 plugin.yaml 能由 ParseProject 解析；run 对缺文件、合法工程、必需与可选依赖的行为。sdk/ 与 internal/core/supervisor/ 的测试及测试插件均只在本地自测（D-86）。

完成条件：gofmt -l 对改动 Go 文件无输出；go vet ./...；go test -race ./contracts/manifest/... ./internal/core/packages/... ./cmd/alonabot/...；go build ./...。

不做：build 与 pack 命令（T-09.3）；构建信息的取值来源（SDK 的协议与能力、目标平台、迁移接口、是否链接受管 SQLite、config_schema.json 的 $version）；新增管理页面或显示入口。
```

## 第 9 波

### goal-96 · T-04.9 删除构建期标记

状态：已完成（[PR #110](https://github.com/AlonaBot/AlonaBot/pull/110)，合并提交 `6aed25d`），证据见 TEST-MATRIX §3.2。派发时：goal-103（PR #99）已合并；按 D-86 清除已删除测试与 testdata 的迁移要求，跨进程自测由本地测试在临时目录生成插件源码与二进制。前置 T-04.7（PR #89，合并提交 `3e37cd6`）已合并。依据 D-79 新增。符号已按合入代码核对：
- 删除测试后，标记示例只剩 `generate.go` 的包文档；`examples/plugins/` 不用标记。
- `sdk/generate/marker.go`：`scan.marker`、`checkTokens`、路径与 route_id 的重复检查。`generate.go`：`Command`、`Event` 类型，`Result.Commands`、`Result.Events`，`scan.paths`、`scan.routes`，`readFile` 的 step 3 与收集 `handlers` 的部分，`Scan` 中对两个切片的排序。`source.go`：imports 中为标记函数加的别名、`Commands:` 与 `Events:` 两段。`cmd/alonabot/plugin.go` 只读 `Result.Diagnostics` 与 `Source`，不用改。
- `sdk/generated.go`：`Generated.Commands`、`Generated.Events`，`registerCommands`、`registerEvents`；`Collect` 调用这两个方法。
- supervisor 测试插件与三个生成文件已由 PR #99 删除；本切片只为本地临时插件运行 `alonabot plugin generate` 与 `check`。

```text
/goal 使用 $alonabot-execute 推进 goal-96（切片 T-04.9：删除构建期标记）。goal 目录 goal-96，分支 goal-96，worktree .worktrees/goal-96，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：删除 //alonabot:command、//alonabot:event 构建期标记。命令、文本匹配与事件只在根插件的注册方法中链式注册，包级函数也这样注册；生成器不再扫描注册，生成文件只有配置 Schema 与 Service Schema。

依据：PLAN T-04 步骤 2；CONTRACTS §9.1；DESIGN §D；DECISIONS D-70、D-79；TEST-MATRIX V-10。

范围（只新增或修改这些）：
- sdk/generate/：删除 marker.go；generate.go 与 source.go 删除标记的扫描、诊断、Result 字段与生成代码，同步包文档
- sdk/generated.go：删除 Commands、Events、registerCommands、registerEvents，同步 Generated 的注释；sdk/sdk.go 的 Collect 删除对应调用
允许的新跨模块边：无。

接口：
- 删除 generate.Command、generate.Event、Result.Commands、Result.Events。
- 删除 sdk.Generated 的 Commands、Events。其余公开符号不变；Generated 只剩 ServiceSchemas 与 ConfigSchema（goal-53 合并后还有 I18n）。

行为：
- 生成器的根类型检查、Config 检查、构建标签与扫描范围不变。
- 以 //alonabot: 开头的注释不再有任何含义：写在函数、方法、类型或变量上都不产生路由，也不产生诊断。
- 生成文件仍然调用 UseGenerated，没有 Schema 时 Generated 为空字面量。imports 只有 sdk。
- 不留兼容读取。旧的生成文件引用已删除的字段，编译失败；alonabot plugin check 报告过期。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：在临时 module 构造引用旧 Generated.Commands/Events 的生成文件，alonabot plugin check 报告过期，重新生成后 check 通过；新增 sdk/generate/generate_local_test.go：临时 module 的源码在 Commands、Events 方法中链式注册，覆盖根类型、构建标签与外部依赖扫描。源码中留一条写在函数上和一条写在方法上的 //alonabot:command：Diagnostics 为空，生成文件没有注册代码。两次生成逐字节相同。internal/core/supervisor/command_local_test.go（TCP 与 stdio；由本地测试在临时目录生成并构建最小命令插件）：插件声明一个包级函数 Handler，上方留一条 //alonabot:command stale 注释，并在 Commands 中以 r.Command("pkg").Handle(pkgHandler) 注册：/pkg 照常回复，/stale 不产生路由。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./sdk/... ./internal/core/... ./cmd/...；go build ./...。

不做：config_schema.json（T-08.11）；i18n 的嵌入（T-08.8）；alonabot build 与 check 命令（goal-38）。
```

### goal-15 · T-05.1 可信上下文与有界分发

状态：已完成（[PR #67](https://github.com/AlonaBot/AlonaBot/pull/67)，合并提交 `c6778be`）。`router.New(Options)`、`Router.Close(ctx) error` 与导出的 `wire.Budget.Acquire`/`Release` 已合入；后继 goal 中 `router.New` 的旧写法在各自改写时按 Options 调整。派发前核对的符号：
- `router.New(queueEvents int, adapters *adapter.Registry, send ServiceFunc, prepare func(*pb.Event, Caller) (*pb.Event, error), prefixes func(adapterInstanceID string) []string, policy CommandPolicy) *Router`。调用点只在 `internal/core/adapter/remote/remote_test.go` 与 `internal/core/supervisor` 的 bus、command、conversation、echo、index、media_in、policy、service 测试中；`cmd/alonabot` 还没有装配 Router。
- `Router.Start()` 启动唯一的 worker；`Router.Close()` 取消 `r.ctx` 并等待 worker，没有参数和返回值，调用点同上。`HasInFlight` 在 `r.ctx` 取消后返回 false。
- `dispatch` step 1 在链开始时计算 `actor.Permissions`：{user} ∪（来源权限中除 user、admin 以外的名字），经 `CommandPolicy.IsAdmin` 加入 admin。`call` 的顺序是 `prepare` → `bind` → 构造 HandlerRequest，`Invocation{PluginID, Event: prepared, DefaultTarget: ev.GetTarget()}`。
- `CommandPolicy` 有 `CheckCommand` 与 `IsAdmin`；Actor 为 nil 时 `IsAdmin` 返回 false，非空权限集合不满足。命令层只处理 `message.created`，`scheduled` 不进入命令层。
- `wire.Budget` 的 `acquire`、`release` 是包私有方法，调用方只有 `internal/wire/session.go`；nil 表示不限额。
- `tune.Values` 已有 `DispatchConcurrency`、`DispatchQueueEvents`、`DispatchQueueBytes` 与 `CoreMaxPayload`。
- `message.proto` 注明 PreEventResult 只能改写 `event_payload`，不能改写 Event 的其余字段。
- SDK 的事件路由目前不写 `required_permissions`；`command_test.go` 在 Attach 前直接修改路由描述来声明权限。T-04.7 已由 #89 合并（`3e37cd6`）并重写了注册 API：本切片的测试插件按新 API 迁移，见 PR #67 的审查评论。

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
- 更新 router.New 与 Router.Close 的全部调用者（remote_test.go 与 internal/core/supervisor 的测试；根包 assembly.go 已合入时也一并更新）。依赖事件先后顺序的现有测试传 Concurrency: 1
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

状态：已完成（[PR #116](https://github.com/AlonaBot/AlonaBot/pull/116)，合并提交 `e65eb4d`），证据见 TEST-MATRIX §3.2。审查中补齐正文结束后的响应 trailer 过滤，清除插件的 `Set-Cookie` 与凭证 CORS 头。派发时：前置 T-07.4（PR #53）、T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）与 T-09.1（PR #72）均已合并。依据 D-84 新增；2026-11-09 审查者按 main `3dec9a0` 的代码写出接口、行为与测试段。已核对：`internal/core/webhost` 的 `Options{AdminBaseURL, PluginBaseURL, SecondHost, Admin, Proxy, Sessions, Now}`、`Host.ServeHTTP` 按 Host 分派、`IssueTicket`/`exchange`/`Authorize`（ticket.go，插件 Cookie 与派生会话）与 policy.go 的 `pluginHeaders`（单主机时 CSP 为 `default-src 'none'; frame-ancestors 'none'`，主动内容返回 409）；`httpproxy.Server` 的 `Authorizer` 接口、私有路由在 `s.auth == nil` 或未授权时返回 401、`stripIdentity` 去掉 Authorization、Forwarded、X-Forwarded-* 与 `alonabot_` 前缀 Cookie；`adminauth.Auth` 的 `Require`（Bearer 或管理 Cookie，Cookie 的变更请求校验 `X-Alonabot-CSRF` 与 Origin）、`Lookup`、`Valid`；`settings.Settings` 的 `AdminBaseURL`、`PluginBaseURL` 与 `admin_base_url`/`plugin_base_url` 的校验和写回；`cmd/alonabot/main.go` 以 `pluginAuthorizer` 把 `webhost.Host.Authorize` 接给代理；`cmd/alonabot/e2e_test.go` 的 `TestAdminAuthRun` 写入两个 Base URL。goal-27（T-07.5）把 HTTP 的 Bind/Unbind 从 main.go 移进 supervisor，两者可以并行，后合并的一方按先合并的代码变基。

```text
/goal 使用 $alonabot-execute 推进 goal-98（切片 T-07.7：插件页面的沙箱隔离）。goal 目录 goal-98，分支 goal-98，worktree .worktrees/goal-98，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 只用一个主机名：删除 plugin_base_url、插件兑换票据与插件会话，core_config.json 的 admin_base_url 改名为 base_url。插件路由的全部响应加上不含 allow-same-origin 的 CSP sandbox 与 nosniff，并去掉插件设置的 Set-Cookie；非公开路由要求管理会话，非 GET 请求验证 CSRF。

依据：CONTRACTS §7.2～§7.5、§15.1；DESIGN §H；DECISIONS D-15、D-84；TEST-MATRIX V-47 的组件分项。

范围（只新增或修改这些）：
- internal/core/settings/：Settings 的 AdminBaseURL 改名为 BaseURL，删除 PluginBaseURL；键 admin_base_url 改为 base_url，删除 plugin_base_url 及其双主机校验；缺省内容与写回同步
- internal/core/webhost/：删除 ticket.go、插件 Cookie、插件主机分支与 SecondHost；policy.go 改为统一的沙箱响应头；host.go 的 Options 只保留单一 Base URL
- internal/core/httpproxy/proxy.go：只改授权调用处的注释（行为不变，私有路由的授权方由装配处换成 adminauth）
- internal/core/adminauth/：新增 Authorize，实现 httpproxy.Authorizer
- cmd/alonabot/main.go：改用 base_url，代理授权改接 adminauth
- cmd/alonabot/e2e_test.go；examples/plugins/lifecycle/main.go 增加三个 HTTP 路由供 e2e 使用（私有 GET /page 返回 HTML 并设置 Set-Cookie 与自己的 CSP，私有 POST /echo 回显收到的请求头，公开回调 POST /callback）。D-86 不允许新增 testdata 测试插件
- ARCHITECTURE.md、docs/development.md 中 admin_base_url、plugin_base_url 与插件票据的描述
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/settings
   Settings 的 `Listen, AdminBaseURL, PluginBaseURL string` 改为 `Listen, BaseURL string`；core_config.json 的键为 base_url，缺省 `http://127.0.0.1:7650`（CONTRACTS §15.1）。写有 admin_base_url 或 plugin_base_url 的文件按未知键报错（首次发布前不写兼容层）。
2. internal/core/webhost
   type Options struct { BaseURL *url.URL; Admin http.Handler; Proxy *httpproxy.Server; Now func() time.Time } // 删除 PluginBaseURL、SecondHost 与 Sessions
   删除 Host.IssueTicket、Host.Authorize 与 AdminSessions 接口；Host.ServeHTTP 只接受与 BaseURL 主机名相同的请求，/plugins/ 前缀交给 Proxy，其余交给 Admin；其他 Host 返回 421。
3. internal/core/adminauth
   func (a *Auth) Authorize(r *http.Request, pluginID string) bool // 与 Require 相同的认证与 CSRF 规则；pluginID 只用于日志
4. cmd/alonabot
   代理改为 httpproxy.New(auth)（auth 为本次 run 的 *adminauth.Auth），删除 pluginAuthorizer；非本机 listen 的检查改为要求 HTTPS base_url 或 trusted_proxies。

行为：
- 响应头：插件路由的每个响应（含代理错误与 Core 返回的 401/404/405/503）由 Core 设置 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff`，覆盖插件设置的同名头；`Set-Cookie` 全部删除并以 WARN 记录名称（不记录值）；`Access-Control-Allow-Credentials` 继续删除。不再对主动内容返回 409，缺失 Content-Type 时仍补 `application/octet-stream`。
- 授权：非公开路由按 adminauth 的规则：Bearer 管理 token 或有效管理 Cookie；以 Cookie 认证的非 GET/HEAD 请求要求 `X-Alonabot-CSRF` 与受控 Origin；不满足时一律 401，与 Admin API 的 Require 一致。公开路由（Public）不要求会话。
- 转发：`stripIdentity` 不变，管理 Cookie、Authorization 与 CSRF 头都不转发给插件（CSRF 头加入删除列表）。
- 删除插件主机、票据端点 `/_alonabot/ticket` 与插件会话，不留兼容读取；DECISIONS D-84 已记录取舍，不再新增条目。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/e2e_test.go 新增 TestPluginPageSandbox（真实 alonabot run 与 lifecycle 插件子进程，单一 base_url）：私有 GET 返回的 HTML 带 sandbox CSP 与 nosniff，插件设置的同名头被覆盖；插件设置的 Set-Cookie 不出现在响应中；未登录访问私有路由 401，登录后带 Cookie 的 GET 200、缺 CSRF 的 POST 401、带 CSRF 的 POST 200；Bearer 访问私有路由 200；公开回调路由无需会话；插件收到的请求中没有管理 Cookie、Authorization 与 CSRF 头（由插件回显请求头核对）；其他 Host 返回 421。TestAdminAuthRun 与 TestSettingsAndListenOverride 改用 base_url。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：settings 对 base_url 的校验、旧键报未知键、缺省写回；webhost 对 401/404/405/503 响应也加沙箱头；Set-Cookie 的多值与大小写；adminauth.Authorize 在 HEAD、OPTIONS、过期会话、认证代次变化时的结果。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/settings/... ./internal/core/webhost/... ./internal/core/httpproxy/... ./internal/core/adminauth/... ./cmd/alonabot/...；go build ./...。

不做：WebUI 的 iframe 与 postMessage 桥（T-09.12、T-09.15）；Admin API 的其他端点；插件 HTTP 路由的注册方式（已在 T-07.4）。
```

### goal-33 · T-08.6 SQLite 候选与资源交接

状态：已完成（[PR #85](https://github.com/AlonaBot/AlonaBot/pull/85)，合并提交 `7d554bd`）。`candidate.PrepareDatabase`、`Discard` 与 state 的两个方法已合入，调用顺序写在包文档；生产装配尚未调用。Windows 只做了交叉构建，实际运行留给 T-10。验证证据见 TEST-MATRIX §3.2。

派发时的核对：前置 T-08.4（PR #74）与 T-08.5（PR #79，合并提交 `75c16d3`）均已合并。符号已按合入代码核对：
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

状态：已完成（[PR #103](https://github.com/AlonaBot/AlonaBot/pull/103)，合并提交 `c840af7`），证据见 TEST-MATRIX §3.2。审查中经审查者授权超出语句两处：showcase 的 `hello → greeting` 改名改用 `$migrations` 声明，`OnConfigLoad` 只保留 `interval_seconds` 非负检查（FIX-01）；`supervisor.Spec.ConfigSchema` 的字段注释改为 config_schema.json 的规范 JSON（FIX-02）。派发时：goal-103（PR #99）已合并；按 D-86 清除已删除测试与 testdata 的迁移要求，跨进程自测由本地测试在临时目录生成插件源码与二进制。前置 T-04.7（PR #89，合并提交 `3e37cd6`）、T-08.3、T-08.4、T-08.10 均已合并。依据 D-76 新增，同日按 D-85 补充 `$version` 与 `$migrations` 的声明式迁移，按 D-80 改用 `i18n/`。

T-04.7 合并后核对时修订了原语句：T-04.7 没有复用配置的类型推导，`sdk/generate/service.go` 有自己的 `serviceBuilder` 与 `schemaImporter`，不依赖 `config.go`。因此 `config.go` 整个删除，不需要拆分或保留。

已核对现有实现（main `3e37cd6`）：
- `sdk.Root` 有 Hook `Config() config.Config`，`sdk.Base` 默认返回 nil；`sdk/config` 的 `Config` 接口只有 `OnLoad`，`config.Base` 提供原样返回 raw 的默认实现。
- `sdk/session.go` 的 `prepare` 调用 `config.Prepare(ctx, root.Config(), generatedFor(root).ConfigSchema, raw, boot.PreviousConfigVersion, boot.GetTargetConfigVersion(), restart)`，失败一律是 `*validate.Error`；`buildRegister` 用 `config.SchemaHash` 计算 `config_schema_hash`。
- `sdk/config/prepare.go` 用 jsonschema-go 补默认值与校验，再按 Go 类型逐层检查；`sdk/config/schema.go` 定义 `KeywordSecret`、`KeywordHidden`、`KeywordForm` 与 `MarshalSchema`、`SchemaHash`。
- `sdk/generate/config.go` 的 `scan.configSchema` 经 go/types 从 Config 声明构造 JSON Schema 并解析 `config` 标签，结果写入 `Result.ConfigSchema` 与生成代码中的 `Generated.ConfigSchema`。`generate.go` 的 `readFile` 把嵌入 `config.Base` 的类型记入 `scan.configs`，`checkTypes` 检查至多一个并写入 `Result.ConfigType`；`service.go` 的 `scan.schemas` 在有 Config 时调用 `configSchema`，这一分支随之删除，其余的 Service Schema 生成不变。
- `sdk/services.go` 与 `internal/core/adapter/remote/declarations.go` 仍用 jsonschema-go，go.mod 不变。
- `internal/core/secret` 的 `Extract` 沿 JSON Schema 的 `properties`、`items`、`additionalProperties` 查找 `x-alonabot-secret`（`path.go` 的 `schemaPaths`，`*` 表示任意下标或键）；`supervisor.Spec.ConfigSchema` 由调用方传入当前包的 Schema。Core 不按 Schema 校验配置。
- 使用旧写法的存活代码为 `examples/plugins/showcase`；原 SDK 与 supervisor 测试已由 PR #99 删除。

```text
/goal 使用 $alonabot-execute 推进 goal-95（切片 T-08.11：配置改由 config_schema.json 声明）。goal 目录 goal-95，分支 goal-95，worktree .worktrees/goal-95，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件配置改由与 plugin.yaml 同目录的手写 config_schema.json 声明，键名沿用 AstrBot 的 _conf_schema.json，另有 secret、required 与只能选一个模板的 template。`alonabot build` 与 `check` 检查并嵌入它；SDK 在准备阶段按 `$migrations` 迁移旧配置，再补缺失值并校验，插件经 sdk.Base 的 Config() 按类型取值；Core 按它抽取 secret。删除 config 标签、Go 配置类型与 config.Config、config.Base。

依据：CONTRACTS §1.1、§4.1、§5.1～§5.3、§9.1；DESIGN §D、§J；DECISIONS D-76；TEST-MATRIX V-25 的描述分项、V-37 与 V-38 的配置分项。

范围（只新增或修改这些）：
- 新增 contracts/configschema/：config_schema.json 的解析与检查、规范 JSON 与摘要、补缺失值与校验、secret 路径、i18n/ 中 config 覆盖与 schema 的核对；生成器、SDK 与 Core 共用
- sdk/config/：删除 Config、Base、Keyword*、MarshalSchema、SchemaHash 与基于 jsonschema-go 的准备；Prepare 改用 contracts/configschema 并返回取值对象；新增 Values 与 Template
- sdk/base.go、sdk/session.go、sdk/generated.go：Root 删除 Config Hook、增加 OnConfigLoad；Base 提供 OnConfigLoad 的默认实现与 Config() 取值方法；Generated.ConfigSchema 的注释
- sdk/generate/：删除 config.go、Config 类型发现（readFile、checkTypes、Result.ConfigType）与 scan.schemas 中的配置分支，改为读取、检查并嵌入 config_schema.json，核对 i18n/ 的 config 覆盖；service.go 的 Service Schema 生成不变
- internal/core/secret/：secret 路径改由 contracts/configschema 给出
- 使用旧写法的示例：examples/plugins/showcase；本地验证新建 *_local_test.go，测试插件由本地测试在临时目录生成并构建
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
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Values 各方法对未声明键与类型不符的 panic 消息；Canonical 保留键顺序与数字原文；int64 边界与非有限数字；display_item 的嵌套路径；Merge Patch 删除键后重新准备得到默认值；web/admin/fixtures/ 中已有的 *.config_schema.json 都能通过 Parse（goal-34 未合并时跳过并写明）；新建 sdk/generate/config_local_test.go：临时 module 带 config_schema.json（覆盖十种 type、options 与 labels、slider、condition、editor_mode、invisible、secret、required、template 与 template_list 的 display_item）与 i18n/zh-CN.json；两次生成逐字节相同，Generated.ConfigSchema 等于 Canonical，REGISTER 的 config_schema_hash 等于 Hash。诊断各一例：未知键、AstrBot 的 file 类型、`$migrations` 版本不连续、`rename` 的目标不在 schema 中、labels 多于 options、slider 缺 max、condition 引用不存在的字段、display_item 指向非 string 字段、secret 带 default、required 用于 int、字段键含 "."、i18 覆盖不存在的字段、i18 字段覆盖中出现 name；新建 internal/core/supervisor/config_prepare_local_test.go，在临时目录生成使用 config_schema.json 的测试插件并构建：缺失键补 default 与零值；显式 false、0、"" 与空数组保留；null 被拒绝；options 之外的值；slider 越界；template 含两个模板键；required 的空串；从版本 1 经 `rename`、`map`、`wrap`、`remove` 迁移到版本 3 后再校验，删除的字段之外的未知键保留；配置版本高于 schema 时失败；OnConfigLoad 在声明式迁移之后执行；restart 时改变配置失败。测试插件在 OnStart 中用 Config() 取值并经日志或回复报告，核对取到的是补齐后的值；新建 internal/core/supervisor/secret_local_test.go（插件同样在临时目录生成并构建）：secret 位于 template 与 template_list 内时，快照只存引用、Bootstrap 收到明文；空串不生成引用。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./sdk/... ./internal/core/...；go build ./...。

不做：
- 表单渲染（T-08.7）；i18n/ 中 messages 与 config 根节点的结构检查（T-08.8）；Admin API 返回 schema 与 Pending 的提前校验（T-09.14）；Adapter 实例配置的 schema。
- AstrBot 的 file 类型、_special 与 dict 的 template_schema（D-76 不支持）。
```

### goal-54 · T-11.1 工具描述、目录与策略

状态：已完成（[PR #101](https://github.com/AlonaBot/AlonaBot/pull/101)，合并提交 `de48f26`）。三插件目录与策略验证只在执行者本地运行（D-86），测试插件没有提交，后继切片在测试临时目录自行生成同类插件。以下为派发时的记录：goal-103（PR #99）已合并；按 D-86 清除已删除测试与 testdata 的迁移要求，跨进程自测由本地测试在临时目录生成插件源码与二进制。前置 T-04.4（PR #88，合并提交 `c8dd516`）、T-04.7（PR #89，合并提交 `3e37cd6`）与 T-06.1（PR #45）均已合并。原语句未列 T-04.7，但它依赖的写法与 goal-88 冲突：goal-88 删除 `ServiceOption`，JSON Service 的 Schema 由生成器产生（D-70、CONTRACTS §3），`jsonschema.For[In]` 得到的 Schema 可能与 Service 实际校验用的不一致，不符合 §13.1「输入 Schema 即该 Service 的输入 Schema」。

sdk 段已按合入代码核对：
- `sdk.JSON` 的 `prepare` 闭包里有 `inputID` 与 `r.schemas`（`Generated.ServiceSchemas`），`b.route` 构造 `ServiceSpec`；`setAccess` 在 `handles != 0` 时置 `afterHandle`，`Tool` 用同样的写法。
- 生成器对切片和 map 生成 `"type":["array","null"]`、`"type":["object","null"]`，结构体生成 `"type":"object"`。In 是具名 map 或切片时，根类型不是单一的 object，注册校验要明确拒绝，语句已写明。
- 原 svcplugin 已由 PR #99 删除；工具验证由本地测试在临时目录生成插件源码、经正式生成器产生 Service Schema 后构建。

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
- 上述包的本地 *_local_test.go（不提交），真实插件源码与生成文件放在测试临时目录
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
5. sdk
   func (s *JSONService[In, Out]) Tool(description string, defaultPolicy ToolDefault) *JSONService[In, Out] // type ToolDefault int：ToolAllow、ToolConfirm；ProtoService 没有此方法
   type ToolInfo struct { Name, Description string; InputSchema json.RawMessage; Policy string } // Policy 为 "allow" 或 "confirm"
   func ListTools(ctx context.Context) ([]ToolInfo, error)

行为：
- SDK：声明了 Tool 的 JSONService，input_schema_json 取 Generated.ServiceSchemas 中 In 的 schema_id 对应的 Schema 原文，与该 Service 校验输入用的是同一份，不另外生成；链式规则与 goal-88 相同（Handle 之后调用 Tool 在 Collect 时报错）。
- 注册校验：tool 只允许出现在 encoding 为 JSON 的服务上；description 非空且不超过 1024 字节（命名常量）；default_policy 为 ALLOW 或 CONFIRM；input_schema_json 必须能被 jsonschema 解析并 Resolve（不加载远程 $ref，ResolveOptions.Loader 为 nil），根的 type 恰好是 object（In 是结构体时如此；具名 map 生成的 ["object","null"] 与切片都被拒绝）。任一失败整体拒绝注册。
- core.tool.list：遍历 Router.Plugins() 中已激活插件的工具服务，按 ServiceCallable 只保留调用插件可以调用的，再去掉生效策略为 DENY 的；按 name 字典序返回。调用者自己的工具同样按规则出现。每次调用都读取当前策略，不缓存。
- 策略：覆盖可为 ALLOW、CONFIRM、DENY；生效策略 = 覆盖或缺省。SetPolicy 在事务提交成功后才返回；安装记录不存在时返回 NOT_FOUND，安装记录存在而插件未激活或服务尚未声明时仍可设置。Effective：有覆盖时返回覆盖；没有覆盖且插件未 Attach 或该服务不是工具时返回 NOT_FOUND。Policies 列出已 Attach 插件的全部工具与全部已保存的覆盖，取不到声明时 Default 为 UNSPECIFIED。修改在下一次 list 与 invoke 时生效。
- 工具调用（core.tool.invoke）与确认在 T-11.2。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：validate.Register 拒绝 Protobuf 服务上的 tool、含远程 $ref 或根类型不是 object 的 input_schema_json（含 ["object","null"]）；description 长度边界；Policies 的输出；未知 ToolPolicy 被拒；Effective 的 NOT_FOUND；internal/core/supervisor/tool_local_test.go（//go:build unix）：在临时目录生成并构建 a、b、c 三个最小服务插件，a 声明 Tool 服务（PUBLIC 的 a/search 为 allow，ALLOWLIST: b 的 a/admin 为 confirm）；b 与 c 收到消息时调用 sdk.ListTools 并回复结果：b 看到 a/admin 与 a/search，c 只看到 a/search；input_schema 与 In 的类型一致，policy 正确；SetPolicy(a, search, DENY) 后两者都看不到 a/search；删除覆盖后恢复；关闭并重新打开状态库后覆盖仍在；description 为空的工具使插件启动失败。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go mod tidy 后无差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：core.tool.invoke、确认与确认按钮（T-11.2）；Agent 组件与 MCP（T-12）；管理 API 与页面（T-09.14、T-09.9）。
```

### goal-121 · T-03.24 插件给出的目标由 Core 补齐

状态：已完成（[PR #135](https://github.com/AlonaBot/AlonaBot/pull/135)，合并提交 `f83ad95`），证据见 TEST-MATRIX §3.2。审查期间按 FIX-1（Reload 的变化键包含 default_instance，授权修改 `settings/file.go` 一行）与 FIX-2（发送许可拒绝时回执只回传插件原始目标，不带出补齐的实例身份）修改。合入的符号：`adapter.Registry.SetDefaults(map[string]string)`、`Registry.Complete(target, inflight *pb.Target) (*pb.Target, *pb.ErrorInfo)`、`settings.AdapterSettings.DefaultInstance`、`Settings.DefaultInstances() map[string]string`；装配尚未调用 SetDefaults（goal-39、goal-69）。派发时：T-08.12（[PR #114](https://github.com/AlonaBot/AlonaBot/pull/114)，合并提交 `a796766`）已合并。`settings.readAdapters` 的 `$version` 已经由 `positiveVersion` 校验；`PrepareInstances` 从原文迁移后再展开环境变量，`TrialInstances` 同用这一准备路径。追加 `default_instance` 时保留现有版本校验、迁移与备份行为。合并后由审查者按 main `2c1fdcc` 核对。前置 T-03.6（PR #55，合并提交 `b023130`）与 T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）均已合并。2026-11-09 按 D-102 新增，接口与测试段按 main（`cbd22a8`）的代码写出：`internal/core/message.Service.Send` 先取 To 或父调用的 `DefaultTarget`，再依次 `policy.CheckSend`（原 Target 豁免用 `proto.Equal` 比较）、`validate.Chain`（引用片段的 target 要求四项标识齐全）、`adapters.Get` 与 adapter_id、account_id 的绑定检查；`internal/core/media` 的 `grant` 用 `targetSubset`（`proto.Equal`）比较 GrantRequest.targets 与原范围，补齐用 Store 的 `readers`（`AddReadServices` 传入的 `*adapter.Registry`）；`adapter.Registry` 按实例 ID 保存运行中的实例，重复的 ID 由 `Add` 拒绝；`settings.readAdapters` 只接受 `$version` 与 `instances`，`validateSettings` 不检查实例名。与 goal-101 共享的 `readAdapters` 已按合入代码核对，按执行顺序 §3 保留两边的修改。

```text
/goal 使用 $alonabot-execute 推进 goal-121（切片 T-03.24：插件给出的目标由 Core 补齐）。goal 目录 goal-121，分支 goal-121，worktree .worktrees/goal-121，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件发出的 Target 可以只带 adapter_id、chat_id、thread_id 与可选的 adapter_instance_id。Core 在准入前补齐空字段：缺少实例时，调用链上在途事件的目标属于同一 Adapter 与 chat_id 则取其实例，否则取该 Adapter 的缺省实例，没有时 FAILED_PRECONDITION；实例已登记时补上 adapter_id 与 account_id。core_config.json 校验实例名并接受可选的 default_instance。SDK 的会话字符串由 T-04.8 提供，本切片只做 Core 侧。

依据：PLAN T-03 步骤 21；CONTRACTS §8（Target 段）、§15.1（adapters）；MESSAGE-MODEL §1（Target）；DECISIONS D-102；TEST-MATRIX V-101 的 Core 分项。

范围（只新增或修改这些）：
- internal/core/adapter/adapter.go：Registry 追加缺省实例表与补齐方法
- internal/core/message/send.go：To 与引用片段 target 的补齐调用点
- internal/core/media/grant.go：GrantRequest.targets 的补齐调用点
- internal/core/settings/settings.go：AdapterSettings 追加 DefaultInstance，readAdapters 接受 default_instance，validateSettings 末尾调用名称检查
- 新增 internal/core/settings/names.go：实例名检查与 DefaultInstances
- contracts/proto/alonabot/v1/identity.proto：只改 Target 的 adapter_instance_id 与 account_id 两行注释（插件发出时可以为空，由 Core 补齐）。随后 go generate ./...
- internal/core/message/permissions_test.go：追加本语句的测试
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/adapter
   func (r *Registry) SetDefaults(defaults map[string]string) // adapter_id → 缺省实例 ID；整体替换并复制参数，可与其他方法并发调用
   func (r *Registry) Complete(target, inflight *pb.Target) (*pb.Target, *pb.ErrorInfo) // 返回补齐后的副本，不修改参数；inflight 可以为 nil
2. internal/core/settings
   type AdapterSettings struct { Version int; Instances map[string]json.RawMessage; DefaultInstance string } // 在 goal-101 合并后的字段上追加
   func (s Settings) DefaultInstances() map[string]string // adapter_id → 缺省实例：default_instance 优先，否则该 Adapter 恰有一个实例时为它，其余不列出

行为：
- Complete：
  1) chat_id 为空，或 adapter_id 与 adapter_instance_id 都为空 → INVALID_ARGUMENT（safe_message 只写字段名）。
  2) adapter_instance_id 为空：inflight 非 nil 且 adapter_id、chat_id 都与 target 相同 → 取 inflight 的实例（thread_id 不参与这条比较）；否则取缺省实例表中该 adapter_id 的实例；都没有 → FAILED_PRECONDITION，safe_message 为 "no adapter instance selected; specify one or set default_instance"。
  3) 实例已登记：adapter_id 为空时补为实例的 AdapterID()，account_id 为空时补为实例的 AccountID()。已给出的字段一律不改，不符的值由调用方原有的绑定检查拒绝。实例未登记时不补，调用方按原有顺序返回 PERMISSION_DENIED 或 UNAVAILABLE，不因补齐泄露实例是否存在。
- core.send（send.go）：
  1) To 非 nil 时先调用 Complete(To, 父调用的 DefaultTarget)；失败时返回 REJECTED 回执（Target 为原 To）与该错误，不调用 CheckSend 与 Adapter。To 为 nil 时照旧使用 DefaultTarget，不调用 Complete。
  2) Chain 中 reference 片段带 target 时，用 Complete(该 target, 补齐后的发送目标) 原位替换，再执行 validate.Chain；失败时返回 REJECTED 回执与该错误。
  3) 之后的 CheckSend、validate.Chain、实例查找与绑定检查、媒体租约与发送顺序不变；原 Target 豁免对补齐后的目标生效。
- core.media.grant（grant.go）：解码后、核对原 token 之前，逐个 Complete(target, 父调用的 DefaultTarget) 并替换；任一失败返回其错误，不签发 token。
- 设置：
  1) readAdapters 在 `$version`、`instances` 之外接受 `default_instance`（字符串），其他键仍报 unknown key。
  2) validateSettings 末尾检查：实例名匹配 ^[A-Za-z0-9_-]{1,64}$；同一实例名出现在两个 Adapter 下 → 错误 "duplicate adapter instance name"；default_instance 非空时必须是该 Adapter 的实例。错误按现有方式指明键路径与行列号，不含配置内容。
  3) 设置的重新读取沿用同一校验；DefaultInstances 返回新 map，调用方修改它不影响 Settings。
- 装配不在本切片：goal-39 在读取设置后调用 Registry.SetDefaults(settings.DefaultInstances())，重新读取后同样调用（goal-69）。

测试（按编码规范「测试提交规则」）：
- 提交：在 internal/core/message/permissions_test.go 追加 TestSendTargetCompletion，复用 TestSendPermissions 的 state、memory Adapter（memory.New("memory", "a", "account-a") 与 memory.New("memory", "b", "account-b")）与 policy 装配：
  - 父调用的 DefaultTarget 为实例 a 的 original 会话与话题；To 只带 adapter_id "memory"、chat_id "original" 与同一 thread_id：发送经实例 a，Adapter 记录的 Target 带 account-a，插件的实例许可为空集合时仍放行（原 Target 豁免）。
  - To 只带 adapter_id 与 chat_id "other"：SetDefaults({"memory": "b"}) 时经实例 b 发出且带 account-b；SetDefaults(nil) 时返回 FAILED_PRECONDITION，两个 Adapter 都没有收到请求。
  - To 带 adapter_instance_id "b" 而不带 account_id：补为 account-b；带错误的 account_id：照旧 INVALID_ARGUMENT。
  - reference 片段的 target 只带 adapter_id 与 chat_id "original"：Adapter 收到的引用目标带实例 a 与 account-a。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Complete 的全部分支（inflight 为 nil、adapter_id 不同、chat_id 不同而 thread 相同、实例未登记、只给实例不给 adapter_id）；SetDefaults 与 Complete 在 -race 下并发；grant 的 targets 补齐后与原范围比较，以及补齐失败不签发 token；设置的实例名非法、跨 Adapter 重名、default_instance 指向不存在的实例、DefaultInstances 的三种情况与返回值隔离。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./internal/core/... ./contracts/...；go build ./...。

不做：SDK 的会话字符串、`sdk.Via` 与类型转换（T-04.8，goal-91）；装配处调用 SetDefaults（goal-39、goal-69）；编辑、撤回与平台动作的补齐（goal-51、goal-107、goal-106 复用 Complete）；WebUI 与 Admin API 中 default_instance 的表单（T-09.14）；跨实例的事件去重。
```

### goal-122 · T-04.10 管理员与封禁按 Adapter 登记

状态：已完成（[PR #133](https://github.com/AlonaBot/AlonaBot/pull/133)，合并提交 `3d618cb`），证据见 TEST-MATRIX §3.2。合入的符号：`policy.Admin{AdapterID, UserID}`、`Policy.IsAdmin(ctx, adapterID string, actor *pb.Actor)`、`Policy.Deny`/`Undeny(ctx, kind, adapterID, subjectID)`、`policy.Deny{Kind, AdapterID, SubjectID}`、`router.CommandPolicy.IsAdmin(ctx, adapterID, actor)`；状态库 `global_deny` 按 `adapter_id`；`settings.Admin{AdapterID, UserID}` 与 `Admin.String()`，admins 为 `"<adapter_id>:<external_user_id>"` 字符串数组。两个 Admin 类型的转换由装配完成（goal-39）。派发时：前置 T-04.4（[PR #88](https://github.com/AlonaBot/AlonaBot/pull/88)，合并提交 `c8dd516`）与 T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）均已合并。2026-11-09 按 D-105 新增，接口按 goal-13 与 goal-97 语句中合入的符号写出。与 goal-121 都修改 `internal/core/settings`，执行顺序 §3 写明了后合并者的处理。

```text
/goal 使用 $alonabot-execute 推进 goal-122（切片 T-04.10：管理员与封禁按 Adapter 登记）。goal 目录 goal-122，分支 goal-122，worktree .worktrees/goal-122，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core Admin 列表与全局禁止不再按实例登记。admins 的元素与用户禁止都用用户字符串 `<adapter_id>:<external_user_id>`，会话禁止按 `<adapter_id>:<chat_id>`；同一 Adapter 的全部实例共用这些设置（D-105）。

依据：PLAN T-04 步骤 10；CONTRACTS §2.3、§15.1；MESSAGE-MODEL §1；DECISIONS D-45、D-82、D-105；TEST-MATRIX V-103。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：global_deny 的 adapter_instance_id 列改为 adapter_id（直接改唯一的初始迁移）；internal/core/state/policy.go 中对应的读写
- internal/core/policy/：admins.go、denies.go、command.go、policy.go 中按实例判定的地方
- internal/core/router/：CommandPolicy 接口与 dispatch 中 IsAdmin 的调用点
- internal/core/settings/：Admin 类型与 admins 的读取、校验和写回
- 上述包的测试文件，以及 internal/core/supervisor/policy_test.go
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   CREATE TABLE global_deny (kind TEXT NOT NULL, adapter_id TEXT NOT NULL, subject_id TEXT NOT NULL, PRIMARY KEY (kind, adapter_id, subject_id)); -- kind 为 user 或 chat；chat 的 subject_id 是 chat_id，不含话题
2. internal/core/policy
   type Admin struct { AdapterID, UserID string } // 取代 { InstanceID, UserID }
   func (p *Policy) IsAdmin(ctx context.Context, adapterID string, actor *pb.Actor) (bool, error) // 按 (adapterID, actor.external_user_id) 判定
   func (p *Policy) Admins() []Admin // 按 AdapterID、UserID 排序
   type Deny struct { Kind DenyKind; AdapterID, SubjectID string }
   func (p *Policy) Deny(ctx context.Context, kind DenyKind, adapterID, subjectID string) error
   func (p *Policy) Undeny(ctx context.Context, kind DenyKind, adapterID, subjectID string) error
   SetAdmins、Denies、CheckCommand 与 CheckSend 的签名不变。
3. internal/core/router
   CommandPolicy 的 IsAdmin 改为 IsAdmin(ctx context.Context, adapterID string, actor *pb.Actor) (bool, error)；dispatch 传入事件的 adapter_id。
4. internal/core/settings
   type Admin struct { AdapterID, UserID string } // JSON 为用户字符串 "<adapter_id>:<external_user_id>"，例如 "telegram:123456789"
   func (a Admin) String() string // 用户字符串

行为：
- 用户字符串：第一个 `:` 之前是 adapter_id，满足 manifest 中 adapter_id 的规则（contracts/validate 已有的检查）；之后是 external_user_id，不能为空，可以含 `:`。不合法、重复或仍是对象写法时，Load 与 Update 报错并指出 admins 的下标，与其他键的校验方式相同。
- admin：Router 给 Actor 加 admin 的条件改为（事件的 adapter_id、actor.external_user_id）在列表中，与收到事件的实例无关。
- 全局禁止：用户禁止按（target.adapter_id、actor.external_user_id），命令总带 target；会话禁止按（target.adapter_id、target.chat_id），不看 thread_id。CheckCommand 与 CheckSend 的顺序、错误码以及「Admin 身份与策略覆盖都不能绕过」不变。
- 参数校验：adapterID 不满足规则或 subjectID 为空 → INVALID_ARGUMENT，取代原来对 instanceID 的检查。
- settings 与 policy 各有一个 Admin 类型，由装配处转换（goal-39 接入）；本切片只同步已有的调用点与测试。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/policy_test.go 改用新签名，原断言保持；追加：同一 Adapter 的两个 memory 实例中，只登记一次的用户在两个实例都得到 admin；会话禁止后，两个实例发往该会话的 sdk.Send 都返回 PERMISSION_DENIED；另一个 Adapter 中 external_user_id 相同的用户不是 admin。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：用户字符串的解析与 String 往返（用户 ID 含 `:`）；不合法的 adapter_id、空用户、重复与对象写法使 Load 报错并指出下标；Update 写回字符串数组且保留其他键的写法；禁止表的增删幂等；Deny 的非法参数。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：admin add/remove 与禁止的 CLI、管理 API 和 WebUI（goal-39、goal-93 按新格式实现）；/whoami 的输出（goal-71）；多轮等待的键与 Agent 记忆的用户范围（仍按实例）。
```

## 第 10 波

### goal-91 · T-04.8 SDK 自有类型与事件常量

状态：已完成（[PR #141](https://github.com/AlonaBot/AlonaBot/pull/141)，合并提交 `ac4aded`），证据见 TEST-MATRIX §3.2。审查者决定：授权在 `sdk/services.go` 中让 Service 调用保存 DefaultTarget；往返等价不计 Core 补齐的 account_id 与 MediaRef 会话实例。审查期间按 FIX-4（扩展事件解码失败的 WARN 不记载荷）、FIX-5（省略 target 的入站引用继承所属会话与实例）、FIX-6（会话字符串、MessageRef 与用户字符串校验 adapter_id）修改。派发时：前置 T-03.9、T-03.24、T-04.7、T-04.9 均已合并；审查者按 main（合入 PR #136 后）核对。T-03.24（[PR #135](https://github.com/AlonaBot/AlonaBot/pull/135)，合并提交 `f83ad95`）已合并，`adapter.Registry.Complete(target, inflight *pb.Target) (*pb.Target, *pb.ErrorInfo)` 与本语句一致：插件给出的会话字符串转换为只含 adapter_id、chat_id、thread_id 与可选实例的 Target，其余由 Core 补齐。T-03.8（goal-8，[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）已合并，`sdk/buttons.go`、`sdk/message/buttons.go`、`sdk/internal/messagecodec/codec.go` 与 Telegram 的 `callback.go` 已并入范围与接口。T-03.9（[PR #118](https://github.com/AlonaBot/AlonaBot/pull/118)，合并提交 `1cbfb2e`）已合并。2026-11-10 按 D-118、D-120、D-121、D-123 加入 `c.Reply` 与后台用的 `sdk.SendText`、用户字符串 `message.User` 与 `c.User`、`c.Session`、收媒体来源的 `sdk.OpenMedia`、`errcode` 的 `New`、`Errorf` 与 `Of`；当时列入的 `internal/core/supervisor/media_in_test.go` 与 `testdata/` 已按 D-86 删除，媒体读取与按钮测试插件的验证改在本地临时目录完成。2026-11-09 按 D-104 加入消息句柄：编辑、撤回与引用只收 `message.MessageRef`，本切片把 T-03.9（goal-51，PR #118）新增的 `c.Edit` 与 `sdk.Edit` 改为句柄形式，因此前置加入 T-03.9；合入的 `sdk/edit.go` 与 `sdk/adapter/edit.go` 已补入范围。同日按 D-108 加入片段构造函数与 `message.Content`，Send、编辑与自动回复收 Chain 或单个片段；按 D-109 把插件的媒体读取改为 `sdk.OpenMedia`、上传改为 `Upload(ctx, r, mime)`，goal-7 当时合入的媒体读取测试随之成为调用方（现已按 D-86 删除）；按 D-110 把时间与时长字段改为 `time.Time` 与 `time.Duration`；按 D-113 让 sdk/adapter 的扩展事件与 extension 用 Go 类型声明，Schema 由 SDK 推导；按 D-102 改写插件侧的目标：`message.Target` 结构改为会话字符串 `message.Session`，新增可选的 `sdk.Via`，account_id 由 T-03.24（goal-121）在 Core 补齐，因此前置加入 T-03.24；本次已按 T-03.24 合入代码核对 `Registry.Complete` 的签名。T-04.9（PR #110，合并提交 `6aed25d`）已合并：生成器不再输出注册代码，`sdk.Generated` 当前为 `ServiceSchemas`、`I18n` 与 `ConfigSchema`（文案字段由 T-08.8 加入）。T-03.11（PR #106，合并提交 `3913d5a`）与 T-15.1（PR #109，合并提交 `1b226a2`）已合并，`adapters/telegram/perm`、`adapters/onebot11/`（含 `onebottest` 与测试）随之成为调用方，已加入语句「范围」的调用方一项。依据 D-70，2026-11-04 从 T-04.7 拆出。T-04.7（PR #89，合并提交 `3e37cd6`）已合并，接口、行为与测试段已按合入代码写出。前置加入 T-04.9：`sdk/generate/source.go` 为标记事件生成 `r.On(id, "类型字符串")`，本切片把 `On` 改为事件常量后，这段生成代码会编译失败，而标记本身由 T-04.9 删除；该生成路径现已删除。T-08.13（PR #137）已合并，`c.T(key, args...) string` 与 `sdk.T(ctx, key, args...) string` 保持；示例的自动回复改签名时继续用 `message.Text(c.T(...))`。

符号已按合入代码核对：
- 公开 API 中的 pb 类型：`Context.Event() *pb.Event`、`To(*pb.Target)`、`Context.Send` 与 `sdk.Send` 的 `*pb.SendResult` 和 `*pb.Target`、`Upload` 的 `*pb.MediaRef`、`GrantMedia` 的 `...*pb.Target`、`Error.Code pb.ErrorCode`、`Public(pb.HttpAuthMode)` 与 `HTTPRoute.AuthMode`、`Command.Mode` 与 `Command.Params`；`message.Chain.Proto() *pb.MessageChain`；`sdk/adapter` 的 `Core.Deliver`、`Core.ReportFailure`、`Instance.Capabilities`、`Instance.Send`、`Editor.Edit`、`Adapter.Descriptor`。T-03.8 的 `Action[T]`、`OnButton`、`ButtonBuilder[T]` 已使用 SDK 声明，内部仍使用 pb 错误码与字符串事件名；`message.Extension` 内部保存 `*pb.ExtensionSegment`。
- 命令参数：`sdk/registry.go` 的 `paramTypes` 把 `*pb.ImageSegment`、`*pb.MessageChain` 映射到 `ParamType`；`sdk/dispatch.go` 把 `ArgValue` 的 image、message 分支直接传给 Handler；当前自动回复签名为 `(message.Chain, error)`。
- `sdk` 导入 `sdk/adapter`（`sdk/adapter_run.go`），`sdk/adapter` 不能反向导入 `sdk`，所以两边共用的类型放在更底层的包。`sdk/message` 已被两边依赖。
- `sdk/message` 的 `Segment` 是带未导出 `toProto` 的接口，当前有 `Text`、`Buttons` 与 `Extension`；`Button` 是 `messagecodec.Button` 的别名，公开字段为 `ActionID`、`Label`、`Value`，另存未导出的编码错误。`messagecodec` 的 `NewButton`、`ButtonError`、`Failure`、`NewFailure`、`Error` 负责把编码失败传到发送入口，本身没有 pb 类型。协议已有 `Button.core_token = 4`，Telegram 的 `renderButtons` 依赖它；SDK 的 Adapter 消息转换须保留该字段。
- 调用方：`adapters/telegram/`（含 `callback.go` 与 `callback_test.go` 的 `TestCallbackDelivery`）、`adapters/onebot11/`（含 `onebottest` 与测试）、`internal/core/adapter/remote/remote_test.go` 的回环 Adapter/插件、`examples/plugins/` 与 `cmd/alonabot/e2e_test.go`。`internal/core/supervisor/` 当前只提交 `supervisor_test.go`，无 `testdata/`；T-03.8 的按钮插件及此前媒体、命令插件均在本地自测临时目录生成。T-03.24 后插件发送的 Target 只需 adapter_id、chat_id、thread_id，Via 另填可选 adapter_instance_id，account_id 由 Core 补齐。
- 用户自己的 Protobuf 消息继续使用 `proto.Message`：`sdk.Proto` 与 `sdk.CallProto` 的约束保持，内部协议消息仍由 SDK 编解码。

```text
/goal 使用 $alonabot-execute 推进 goal-91（切片 T-04.8：SDK 自有类型与事件常量）。goal 目录 goal-91，分支 goal-91，worktree .worktrees/goal-91，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：
- 插件与 Adapter 的公共 API 不再暴露生成的 pb 类型：SDK 自己定义消息、事件、会话字符串、Actor、SendResult 与错误码，pb 只留在 SDK 内部。发送目标是会话字符串，机器人是可选参数 sdk.Via（D-102）；编辑、撤回与引用所指的消息是消息句柄 message.MessageRef（D-104）。
- 每种片段都有构造函数；Send、编辑与自动回复收 message.Content，即 Chain 或单个片段（D-108）。
- 插件经 sdk.OpenMedia 顺序读取媒体；Upload 只收内容与 MIME，大小与 SHA-256 由 SDK 计算（D-109）。
- SDK 类型中的时间用 time.Time，时长用 time.Duration（D-110）。
- 常用操作的简写：c.Reply 回一段文本，c.User 与 c.Session 给出用户字符串与会话字符串，OpenMedia 直接收片段的媒体来源，errcode 提供构造与判断的辅助函数（D-118、D-120、D-121、D-123）。
- 新增标准事件常量（带 payload 的 Go 类型）与 sdk.AdapterEvent；事件注册改用常量，Collect 核对事件 Handler 的 payload 参数类型。
- 命令参数类型改为 SDK 自有类型。
- sdk/adapter 的接口改用 SDK 类型，扩展事件与 extension 用 Go 类型声明、Schema 由 SDK 推导（D-113）；官方 Telegram Adapter 与已合并的全部调用方同步。
- 不改协议：proto 与 Core 侧不变，SDK 在边界内转换。

依据：PLAN T-04 步骤 9（自有类型部分）；CONTRACTS §1.3、§1.4、§2.1、§8；MESSAGE-MODEL §1～§3；DESIGN §D；DECISIONS D-70、D-71、D-102、D-104、D-108～D-110、D-113、D-118、D-120、D-121、D-123；TEST-MATRIX V-14 的类型分项、V-86 的 SDK 分项、V-101 与 V-102 的 SDK 分项、V-106～V-108、V-111 的事件与 extension 分项、V-113 的 SDK 类型分项。

范围（只新增或修改这些）：
- 新增 sdk/errcode/
- sdk/message/：message.go、buttons.go 及新增类型文件；扩展为下面的类型，删除 Chain.Proto 与各片段的 toProto，把 pb 转换移到 sdk/internal/convert
- 新增 sdk/internal/convert/：pb 与 SDK 类型的双向转换，只有 sdk 与 sdk/adapter 的实现使用
- sdk/internal/messagecodec/codec.go：同步 Button 的 SDK 自有字段与编码错误传递，保留现有失败机制
- sdk/：sdk.go、send.go、edit.go、media.go、registry.go、events.go、buttons.go、dispatch.go、http.go、adapter_run.go、session.go（只改转换的调用点）；其余构造 sdk.Error 的文件只把错误码改为 errcode
- sdk/adapter/adapter.go、edit.go；可新增声明的实现文件
- 调用方：adapters/telegram/（含 callback.go、callback_test.go 及发送、编辑测试）、adapters/onebot11/（含 onebottest 与测试）、examples/plugins/、internal/core/adapter/remote/remote_test.go、cmd/alonabot/e2e_test.go，以及这些测试中嵌入的插件源码。internal/core/supervisor/ 的按钮、媒体与命令测试插件只在 *_local_test.go 的临时目录中生成；已删除的 testdata/、media_in_test.go 不恢复
- 上述包的测试文件
允许的新跨模块边：无（都在 sdk 模块内；Core 与 contracts 不变）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/errcode
   type Code int32 // 数值与协议的错误码相同
   const ( InvalidArgument Code = 1; Unauthenticated; PermissionDenied; NotFound; AlreadyExists; FailedPrecondition; Unsupported; ResourceExhausted; DeadlineExceeded; Cancelled; Unavailable; Internal; ProtocolError; CallCycle; PluginNotReady; NoDefaultTarget; NeedsManualMigration )
   func (c Code) String() string // 与现在的 pb 枚举名相同，如 ERROR_CODE_INVALID_ARGUMENT
   type Error struct { Code Code; Message string }
   func (e *Error) Error() string // 「错误码: 说明」，与现在的 sdk.Error 相同
   func (c Code) New(message string) *Error                 // D-123
   func (c Code) Errorf(format string, args ...any) *Error  // D-123
   func Of(err error) Code // err 的链中有 *Error 时返回它的 Code，否则（含 nil）返回 0（D-123）
2. sdk/message（可选字段的零值表示 absent，转换时对应未设置的 optional 字段；时间为 time.Time，时长为 time.Duration，D-110）
   身份：
   type Session string // 会话字符串 <adapter_id>:<chat_id>[#<thread_id>]（MESSAGE-MODEL §1，D-102）；插件可以直接写字面量
   func NewSession(adapterID, chatID, threadID string) Session // chat_id 与 thread_id 中的 % 与 # 写作 %25、%23；threadID 为空表示会话本身
   func (s Session) Parts() (adapterID, chatID, threadID string, ok bool) // 格式不合法时 ok 为 false
   type MessageRef string // 消息句柄 <会话字符串>/<platform_message_id>[@<实例名>]（MESSAGE-MODEL §4，D-104）；可以原样存储
   func NewMessageRef(s Session, platformMessageID, instanceID string) MessageRef // platform_message_id 中的 %、/、@ 写作 %25、%2F、%40；instanceID 为空时省略 @ 部分
   func (r MessageRef) Parts() (s Session, platformMessageID, instanceID string, ok bool) // 句柄或其中的会话字符串不合法时 ok 为 false
   type User string // 用户字符串 <adapter_id>:<external_user_id>，格式与 core_config.json 的 admins 相同（D-105、D-120）；第一个 : 之前是 adapter_id，之后原样
   func NewUser(adapterID, externalUserID string) User
   func (u User) Parts() (adapterID, externalUserID string, ok bool) // 没有 : 或任一部分为空时 ok 为 false
   type Actor struct { ExternalUserID, AdapterInstanceID, DisplayName, Locale string; Permissions []string }
   type Capability struct { Name string; Version uint32; Supported bool }
   片段（Chain 中的片段都是值类型，type switch 按值匹配）：
   type Content interface{ content() Chain } // Chain 与每种片段都实现（D-108）
   type Segment interface{ Content; segment() }
   type Chain []Segment
   func (c Chain) Text() string // 依次拼接全部 TextSegment 的文本
   func Text(s string) Segment  // 保留，返回 TextSegment
   type TextFormat int; const ( FormatPlain TextFormat = iota + 1; FormatMarkdown ) // 零值表示未指定
   type EntityType int; const ( EntityBold EntityType = iota + 1; EntityItalic; EntityUnderline; EntityStrikethrough; EntitySpoiler; EntityCode; EntityPre; EntityTextLink; EntityTextMention )
   type TextEntity struct { Type EntityType; OffsetUTF16, LengthUTF16 uint32; URL, Language, ExternalUserID string }
   type TextSegment struct { Text string; Format TextFormat; Entities []TextEntity }
   type Readability int; const ( Readable Readability = iota + 1; Unreadable )
   type IncomingMedia struct { Token string; Size uint64; MIME string; Readability Readability; ExpiresAt time.Time } // Size 为 0 表示未知
   type MediaSource struct { MediaRef string; Incoming *IncomingMedia; HTTPSURL string } // 恰有一个
   type MediaRef struct { Token string; Size uint64; MIME string; SHA256 []byte; ExpiresAt time.Time; ConsumerPluginIDs []string; Sessions []Session }
   func (m *MediaRef) Source() MediaSource       // 以该 MediaRef 为来源
   func (m *IncomingMedia) Source() MediaSource  // 以该入站对象为来源
   func ImageURL(httpsURL string) Segment        // 以下构造函数都返回对应的值类型片段（D-108）
   func VideoURL(httpsURL string) Segment
   func FileURL(httpsURL, filename string) Segment
   func Image(src MediaSource) Segment
   func Video(src MediaSource) Segment
   func File(src MediaSource, filename string) Segment
   func Mention(externalUserID string) Segment
   type ImageSegment struct { Source MediaSource; Alt, Caption string }
   type VideoSegment struct { Source MediaSource; Duration time.Duration; Caption string }
   type FileSegment struct { Source MediaSource; Filename, MIME, Caption string }
   type MentionSegment struct { ExternalUserID, Display string }
   type ReferenceSegment struct { Ref MessageRef } // 句柄省略实例时，Core 以补齐后的发送目标为在途目标补齐（CONTRACTS §8）
   func Quote(ref MessageRef) Segment // 返回 ReferenceSegment
   type Button struct { ActionID, Label, Value, CoreToken string } // CoreToken 对应现有协议 core_token，Core 签发、Adapter 渲染；另有未导出的编码错误
   type ButtonsSegment struct { Rows [][]Button } // 保留未导出的编码错误
   func Buttons(rows ...[]Button) Segment // 保留现有构造签名，返回 ButtonsSegment 值
   type ExtensionSegment struct { AdapterID, Name string; Version uint32; PayloadJSON string } // 保留未导出的编码错误
   func Extension(adapterID, name string, version uint32, payload any) Segment // 保留现有构造签名，JSON 编码 payload，返回 ExtensionSegment 值
   type ForwardSource struct { DisplayName, ExternalUserID string }
   type ForwardItem struct { Source *ForwardSource; Chain Chain }
   type ForwardSegment struct { Items []ForwardItem }
   事件：
   type SourceKind int; const ( SourceAdapter SourceKind = iota + 1; SourceScheduler )
   type PlatformMessage struct { PlatformMessageID string; Chain Chain; MediaGroupID string }
   type Payload interface{ payload() } // 下面的载荷类型以指针实现
   type Event struct { ID, Type string; Source SourceKind; OccurredAt time.Time; AdapterID, AdapterInstanceID, AccountID string; Session Session; Actor *Actor; Message *PlatformMessage; Payload Payload; Capabilities []Capability; PlatformPayload []byte; SourceEventID string }
   func (e *Event) Ref() MessageRef // 本事件所指的消息：message.created 与 message.edited 的消息、reaction.added 被回应的消息、button.clicked 按钮所在的消息（有 origin_platform_message_id 时）；其他事件为空
   type CreatedPayload struct { PlatformMessageID string; Chain Chain; ToMe bool }
   type EditedPayload struct { PlatformMessageID string; Chain, OldChain Chain }
   type DeletedPayload struct { PlatformMessageIDs []string }
   type ButtonClickedPayload struct { ActionID string; ActionSchemaVersion uint32; Value, CallbackQueryID, OriginPlatformMessageID string }
   type MemberPayload struct { MemberExternalUserID string } // member.joined 与 member.left 共用，由 Event.Type 区分
   type ReactionAddedPayload struct { PlatformMessageID, Reaction string }
   type ScheduledPayload struct { JobID, RunID string; ScheduledAt time.Time }
   type BotOnlinePayload struct { AdapterID, AdapterInstanceID, AccountID, AccountDisplayName string }
   type OfflineReason int; const ( OfflineStopped OfflineReason = iota + 1; OfflineAdapterExited; OfflineInstanceFailed )
   type BotOfflinePayload struct { AdapterID, AdapterInstanceID, AccountID string; Reason OfflineReason; SafeError *errcode.Error }
   type AdapterEventPayload struct { Version uint32; PayloadJSON string }
   发送结果：
   type SendStatus int; const ( SendAccepted SendStatus = iota + 1; SendPartial; SendRejected; SendUnknown )
   type PartStatus int; const ( PartAccepted PartStatus = iota + 1; PartRejected; PartUnknown; PartNotAttempted )
   type PartResult struct { PartIndex uint32; SegmentIndexes []uint32; Status PartStatus; PlatformMessageIDs []string; MessageIDsUnavailable bool; Error *errcode.Error }
   type SendResult struct { Status SendStatus; Session Session; AdapterInstanceID string; PlatformMessageIDs []string; PrimaryPlatformMessageID string; ProviderAcceptedAt time.Time; CapabilitiesUsed []Capability; PartResults []PartResult }
   func (r *SendResult) Ref() MessageRef    // PrimaryPlatformMessageID 与 AdapterInstanceID 的句柄；没有确认的消息时为空
   func (r *SendResult) Refs() []MessageRef // PlatformMessageIDs 依次的句柄
3. sdk
   type Error = errcode.Error // 插件继续用 errors.As 取 *sdk.Error
   func (c *Context) Event() *message.Event
   func To(s message.Session) SendOption
   func Via(instanceID string) SendOption // 指定发送的机器人（实例）；缺省由 Core 选择（CONTRACTS §8）
   func (c *Context) Send(m message.Content, opts ...SendOption) (*message.SendResult, error)
   func (c *Context) Reply(text string) error // 等于 c.Send(message.Text(text)) 并丢弃发送结果（D-118）
   func (c *Context) User() message.User       // 触发者的用户字符串；没有 Actor 时为空串（D-120）
   func (c *Context) Session() message.Session // 本次调用的默认会话；没有默认目标时为空串（D-120）
   func Send(ctx context.Context, to message.Session, m message.Content, opts ...SendOption) (*message.SendResult, error) // opts 只接受 Via
   func SendText(ctx context.Context, to message.Session, text string, opts ...SendOption) error // 等于 Send(ctx, to, message.Text(text), opts...) 并丢弃发送结果（D-118）
   func (c *Context) Edit(ref message.MessageRef, m message.Content) (*message.SendResult, error) // 取代 T-03.9 的 Edit(platformMessageID, chain, opts...)
   func Edit(ctx context.Context, ref message.MessageRef, m message.Content) (*message.SendResult, error) // 后台编辑；取代 Edit(ctx, to, platformMessageID, chain)
   func Upload(ctx context.Context, r io.Reader, mime string) (*message.MediaRef, error) // 大小与 SHA-256 由 SDK 计算；mime 为空时按内容判断（D-109）
   func OpenMedia(ctx context.Context, src message.MediaSource) (io.ReadCloser, error) // 顺序读取入站媒体或获授权的媒体，参数是片段的 Source（D-121）；取代 ReadMedia(ctx, token, offset, length)，后者删除（D-109）
   func GrantMedia(ctx context.Context, token, granteePluginID string, sessions ...message.Session) (string, error)
   type HTTPAuth int; const ( AuthCoreSession HTTPAuth = iota + 1; AuthCallbackState; AuthCallbackSignature; AuthCallbackNonce )
   func Public(mode HTTPAuth) HTTPOption // HTTPRoute.AuthMode 同步改为 HTTPAuth
   Command 的 Mode 与 Params 改为不导出的字段
   type EventType struct{ ... } // 事件名与 payload 类型；零值在 Collect 时报错
   var ( MessageCreated, MessageEdited, MessageDeleted, ButtonClicked, MemberJoined, MemberLeft, ReactionAdded, Scheduled, BotOnline, BotOffline EventType ) // 分别对应 MESSAGE-MODEL §1.1 的 event_type 与上面的载荷类型
   func AdapterEvent[T any](name string, version uint32) EventType // name 为 <adapter_id>.<name>；T 必须是结构体，Handler 的参数为 *T
   func (r *EventRegistry) On(routeID string, event EventType) *EventBuilder
   func (b *EventBuilder) Handle(h any) // func(*Context) error，或 func(*Context, P) error；EventRoute.Handler 同步改为 any，EventRoute.EventType 改为 EventType
   type ButtonAction[T any] struct{ ... } // 保留 T-03.8 的动作名与版本声明
   func Action[T any](name string) ButtonAction[T]
   func (a ButtonAction[T]) Button(label string, value T) message.Button
   func OnButton[T any](r *EventRegistry, a ButtonAction[T]) *ButtonBuilder[T] // 内部改用 ButtonClicked 事件常量
   type ButtonBuilder[T any] struct{ ... }
   func (b *ButtonBuilder[T]) Priority(p int32) *ButtonBuilder[T]
   func (b *ButtonBuilder[T]) Permission(p Permission) *ButtonBuilder[T]
   func (b *ButtonBuilder[T]) Handle(h func(*Context, T) error)
   func (c *Context) T(key string, args ...any) string // 保持 T-08.13 合入签名
   func T(ctx context.Context, key string, args ...any) string // 保持 T-08.13 合入签名
4. sdk/adapter
   type ExtensionDeclaration struct{ ... } // 只能由 Extension 构造（D-113）
   func Extension[T any](name string, version uint32) ExtensionDeclaration // T 为该 extension 的 payload 类型，必须是结构体
   type PermissionDeclaration struct { Name string; Implies []string; LabelKey string }
   type EventDeclaration struct{ ... } // 只能由 Event 构造（D-113）
   func Event[T any](name string, version uint32) EventDeclaration // name 为完整名；T 为 payload 类型，必须是结构体，与插件侧 sdk.AdapterEvent[T] 用同一类型
   type Descriptor struct { AdapterID string; Capabilities []message.Capability; Extensions []ExtensionDeclaration; Permissions []PermissionDeclaration; Events []EventDeclaration }
   Core.Deliver(ctx context.Context, ev *message.Event, cursor []byte) (accepted bool, err error)
   Core.ReportFailure(ctx context.Context, safe *errcode.Error) error
   Instance.Capabilities() []message.Capability
   Instance.Send(ctx context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error) // Adapter 用 to.Parts() 取 chat_id 与 thread_id
   Editor.Edit(ctx context.Context, to message.Session, platformMessageID string, chain message.Chain) (*message.SendResult, *errcode.Error) // T-03.9 的可选编辑入口同步换类型
   Adapter.Descriptor() *Descriptor

行为：
- 转换集中在 sdk/internal/convert：入站事件与命令参数从 pb 转为 SDK 类型，出站 Chain、目标、错误与 Adapter 描述反向转换。每个字段双向保持，空串与零值对应未设置的 optional。message 与 sdk/adapter 的公开 API 不出现 pb。wire 与 Core 不变（Core 侧的补齐由 T-03.24 提供）。
- 按钮与 extension：保留 Action 从值接收者 SchemaVersion() uint32 取版本（缺省 1）、button. 保留前缀、动作名/版本校验与 OnButton 的带类型 value 解码。ButtonBuilder 的 Priority、Permission、Handle 顺序检查照旧；Handle 仍接收 T。ButtonAction.Button 与 message.Extension 的 JSON 编码错误经 messagecodec 原样保留，在显式 Send、Edit 或自动回复发出链前返回 INVALID_ARGUMENT。ExtensionSegment 与 pb.ExtensionSegment 的 adapter_id、name、version、payload_json 双向保持；Buttons 的行序、action_id、label、value 与 core_token 双向保持，插件给出的 core_token 仍由 Core 清除并重新签发。Telegram callback 仍先 answerCallbackQuery，再交付 ButtonClickedPayload；inline 回调的游标与跳过行为保持。
- 时间与时长（D-110）：协议的毫秒值不为 0 时转为 time.UnixMilli(ms) 与 time.Duration(ms)*time.Millisecond，0 与未设置转为零值；反向用 UnixMilli() 与 Milliseconds()，零值转为 0 或未设置，亚毫秒部分舍去。出站的负 Duration 在发出请求前返回 INVALID_ARGUMENT。
- 会话字符串（D-102）：
  - pb.Target → Session 为 NewSession(adapter_id, chat_id, thread_id)；接收或发送的实例由 Event、SendResult 的 AdapterInstanceID 给出，account_id 由 Event 的 AccountID 给出。
  - Session → pb.Target：Parts 得到 adapter_id、chat_id、thread_id；Via 给出时填 adapter_instance_id；account_id 留空，由 Core 补齐。格式不合法时在发出请求前返回 INVALID_ARGUMENT。
  - c.Send：To 与 Via 都没有时照旧发往默认目标；只有 Via 时发往默认目标的会话（含 thread），经指定实例发出；只有 To 时由 Core 选择实例。
  - Adapter 交付的 Event 由 Session 与 Event 的 AdapterID、AdapterInstanceID、AccountID 组成完整的 pb.Target；Core 交给 Adapter 的完整 Target 转为 Session。
- 消息句柄（D-104）：
  - SendResult 与 Event 转换后，由 Session、AdapterInstanceID 与消息 ID 组成句柄；Event.Ref 按载荷取消息 ID。
  - Edit：句柄拆成 EditRequest 的 to 与 platform_message_id，to 的 adapter_instance_id 取句柄中的实例，省略时留空由 Core 补齐；c.Edit 带本次调用的父调用，sdk.Edit 不带。句柄不合法时在发出请求前返回 INVALID_ARGUMENT。
  - ReferenceSegment：Ref 拆成引用的 target（adapter_id、chat_id、thread_id 与可选的 adapter_instance_id）与 original_platform_message_id；入站的引用反向组成句柄。
- Content（D-108）：单个片段按只含它的 Chain 处理。自动回复型 Handler 的返回类型为 (message.Content, error)，nil 与空 Chain 表示不回复；checkHandler 不再接受 (message.Chain, error)，报错带 route_id 与期望的签名。显式 Send 与 Edit 收到 nil 或空 Chain 时返回 INVALID_ARGUMENT，与现在相同。
- 媒体（D-109）：
  - OpenMedia 的 token 取自来源（D-121）：src.Incoming 不为 nil 时用它的 Token，否则用 src.MediaRef；两者都没有（https_url 等其他来源）时在发出请求前返回 INVALID_ARGUMENT。只持有 token 的调用方传 message.MediaSource{MediaRef: token}。
  - OpenMedia 以 offset 0 读取第一块（入站对象在此时取件），失败时返回该错误，不返回 reader。之后 Read 用 OpenMedia 的 ctx 按 next_offset 依次读取，每次至多 1 MiB（wire.MaxMediaChunk），eof 后返回 io.EOF；错误为 *sdk.Error，ctx 取消后返回 ctx 的错误；不自动重试（CONTRACTS §1.4）。Close 只停止读取，不撤销 token。
  - Upload：r 实现 io.Seeker 时先读一遍计算大小与 SHA-256，再回到起始位置分块上传；否则边读边写入 os.CreateTemp 的临时文件（0600）并计算，上传结束（含失败）后删除。不整份读入内存。mime 为空时取开头至多 512 字节经 http.DetectContentType 判断。配额与 CommitUpload 的摘要校验照旧由 Core 负责：内容在两遍之间改变时上传失败并返回 Core 的错误。
- Context：Event 每次调用只转换一次，返回值只供本次 Handler 读取，修改它不影响 Send 的默认目标与权限判定。Service 调用的 Event 仍为 nil。Has 与默认 Target 继续使用 SDK 内部的 pb 值。
- 命令参数：`*message.ImageSegment` 对应 image，`message.Chain` 对应 message，其余不变；无默认值的可选参数缺失时取 Handler 声明类型的零值（指针与 Chain 为 nil），与现在相同。checkHandler 的错误文案同步。
- 事件路由：
  - Handle 的签名只能是 func(*Context) error，或第二个参数为该事件的 payload 类型（标准事件是上面的指针类型，AdapterEvent[T] 是 *T）。其他签名在 Collect 时报错，错误中带 route_id、期望的类型与实际的类型。
  - 标准事件：SDK 把 Event.Payload 传给 Handler。实际载荷与路由的类型不符时返回 INTERNAL 并记 ERROR，不调用 Handler。
  - AdapterEvent：Collect 检查 validate.IsAdapterEventType(name)、version 不为 0、T 是结构体。调用时 SDK 比较 version，并把 payload_json 解码为 *T；版本不符或解码失败时记 WARN（只记 route_id 与版本或错误，不记载荷），不调用 Handler，正常结束本次调用。func(*Context) error 形式不解码，Handler 用 c.Event().Payload 读取。
  - 注册未安装 Adapter 的扩展事件照常成功，诊断由 Core 负责（goal-87 已实现）。
- 错误：SDK 向调用者返回的 *sdk.Error 与现在的错误码和文案相同；validate.PublicCode 的映射留在转换之前。
- Adapter 声明（D-113）：Event 与 Extension 用 jsonschema-go 的 For（缺省选项）从 T 推导 Schema，即不允许未声明的属性、没有 omitempty 的字段为必需，与 encoding/json 对同一类型的编码一致；RunAdapter 组装注册描述时把它序列化为 schema_json。T 不是结构体或推导失败时，RunAdapter 在发出 REGISTER 之前返回错误，错误中带声明名。已合并的 remote_test 回环 Adapter 的 loop.poke 改用 Event 声明；当前 OneBot 11 只有接收能力且未声明 face extension，后继切片按同一类型声明规则追加。
- Telegram 与 OneBot 11 Adapter 只换类型，现有平台映射、错误、回执与游标行为保持。sdk/adapter 的 Core.ReadMedia、MediaSource.ReadPlatformFile、Account、StartRequest、Start、Stop 签名保持；Adapter 侧的底层分块读取继续供现有发送实现使用。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/adapter/remote/remote_test.go 的 TestRemotePermissionsAndEvents（TCP 与 stdio，回环 Adapter 改用 sdk/adapter 的新类型）：插件以 `sdk.AdapterEvent[poke]("loop.poke", 1)` 注册，Handler 收到解码后的 *poke；未安装的 other.poke 路由注册成功且不触发；原有断言保持。
  - adapters/telegram 的现有测试（含 callback_test.go 的 TestCallbackDelivery、发送与编辑的按钮 core_token 断言）、adapters/onebot11 的现有集成测试改用 SDK 类型，断言不变。
  - 其余已有集成测试改用新 API，断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 转换的往返：全部片段类型（含 ButtonsSegment 的 core_token 与 ExtensionSegment）、全部载荷类型（含 member.joined 与 member.left）、contracts/testdata/events 的 fixture，pb → SDK → pb 后 proto.Equal；空串与未设置的 optional 相互对应。
  - 按钮 API：Action 的缺省/显式版本、Button 的 SDK 字段与 JSON 值、ButtonBuilder 的 Priority/Permission/Handle 与保留前缀、value 解码失败记 WARN 后结束；编码失败的按钮与 extension 在 Send、Edit、自动回复前返回 INVALID_ARGUMENT，保留原始编码错误。internal/core/supervisor/buttons_local_test.go 的插件在临时目录生成，以 Action 与 OnButton 注册；Core 签发 token 后 SDK Adapter 能读到同一 CoreToken 并渲染，点击后 Handler 收到原 T 值，错误码与路由绑定保持。
  - errcode 的数值与 pb.ErrorCode 逐个一致，String 与 pb 枚举名相同。
  - 时间与时长：各字段的毫秒值往返不变；0 与未设置对应零值；亚毫秒部分舍去；负的 VideoSegment.Duration 在发出请求前返回 INVALID_ARGUMENT。
  - 版本不符与 payload_json 解码失败时：不调用 Handler、记 WARN、日志没有载荷。
  - AdapterEvent 的名称不合法、version 为 0、T 不是结构体；Handle 的各种错误签名；零值 EventType。
  - Event 与 Extension 推导的 Schema：必需字段、omitempty、嵌套结构与切片、空结构体；T 不是结构体或类型无法推导时 RunAdapter 返回错误；回环 Adapter 交付带多余字段的 loop.poke 被 Core 拒绝。
  - Chain.Text 跳过非文本片段；修改 c.Event() 的返回值不影响 c.Send 的默认目标。
  - Session：Telegram 与 OneBot 11 形式的往返、话题、chat_id 中的 `%` 与 `#`、各种不合法格式；To、Via 的四种组合对应的 SendRequest.to。
  - MessageRef：带与不带实例、话题、消息 ID 中的 `%`、`/`、`@` 的往返与各种不合法格式；SendResult.Ref、Refs 与各类事件的 Event.Ref；c.Edit 与 sdk.Edit 的 EditRequest 带句柄中的实例；Quote 与入站引用的往返。
  - 构造函数生成的片段与对应的结构体字面量相同；单个片段、Chain 与 nil 经 c.Send、sdk.Send、Edit 与自动回复得到的请求；DESIGN §D 的示例代码能编译。
  - OpenMedia 与 Upload：可 Seek 与不可 Seek 的 r 上传后经 OpenMedia 读回的字节相同（恰好 1 MiB、1 MiB+1 与跨多块）；MediaRef 的 Size 与 SHA256 与内容一致；mime 为空时的判断；无效 token 时 OpenMedia 返回错误；读取中取消 ctx；不可 Seek 时临时文件在上传结束与失败后都已删除。
  - `go doc -all` 查看 ./sdk、./sdk/message、./sdk/adapter、./sdk/errcode，输出中没有 pb 类型。
  - 简写（D-118、D-120、D-121、D-123）：c.Reply 的 SendRequest 与 c.Send(message.Text(text)) 相同，没有默认目标时返回 NO_DEFAULT_TARGET；sdk.SendText 的 SendRequest 与 sdk.Send(ctx, to, message.Text(text), opts...) 相同（含 Via），会话字符串不合法时在发出请求前返回 INVALID_ARGUMENT；c.User 与 c.Session 在消息事件、没有 Actor 的事件与 Service 调用中的值；NewUser 与 Parts 的往返，external_user_id 含 `:`，各种不合法格式；OpenMedia 收入站片段的 Source、MediaRef.Source() 与只含 token 的来源读到相同字节，https_url 来源在发出请求前返回 INVALID_ARGUMENT；errcode 的 New 与 Errorf 得到的 *Error，Of 对直接、经 fmt.Errorf("%w") 包装、非 *Error 与 nil 的结果。
  - internal/core/supervisor/command_local_test.go（在临时目录生成命令插件，TCP 与 stdio）：事件路由写成 `r.On("typed", sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error {...})`，插件回复 e.Chain.Text()、e.ToMe、c.Event().Actor.Permissions 与 c.Event().Session，与 Adapter 交付的值一致。Handler 的第二个参数写成 *message.ReactionAddedPayload 时（环境变量控制）插件启动失败，日志带 route_id。
  - 现有命令参数用例改用 *message.ImageSegment 与 message.Chain，断言不变。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./...；go build ./...。

不做：Rule（T-05.2）；T 的翻译逻辑与构建期检查（沿用 T-08.8、T-08.13）；Adapter 常量包的新能力（T-03.11、T-15.1 的已合入调用方随类型同步）；docs/plugin-guide.md 的更新（由审查者在合并前同步）。
```

### goal-24 · T-07.2 schedule 注册与运行槽

状态：已完成（[PR #107](https://github.com/AlonaBot/AlonaBot/pull/107)，合并提交 `fc2c247`），证据见 TEST-MATRIX §3.2。2026-11-09 维护者确认：原始 REGISTER 同时带 `cron` 与 `interval_ms` 时按解码后的 oneof 验收（见 DECISIONS），注册校验中「恰有 interval_ms 或 cron 之一」指解码后的 `ScheduleSpec.trigger`。派发时：前置 T-05.1（PR #67，合并提交 `c6778be`）与 T-07.1（PR #17，合并提交 `0271377`）均已合并。2026-11-08 按合入代码核对并改写：`contracts/schedule` 的 `FromSpec`/`Trigger.Next`、`wire.Request`（已有 `Attach`，本切片只加 `Late`）、`router.Plugin` 与 `router.Invocation`（定时调用经 `Attach` 绑定可信上下文，子调用 `core.send` 按无 Actor 的后台调用准入）、`tune.Values.JobTimeout`、SDK 已有的 `ScheduleRegistry.Every`/`Cron` 与 `Registration.Schedules`；按 D-86 去掉 testdata 测试插件，真实进程的验证在本地测试的临时目录中生成插件。

```text
/goal 使用 $alonabot-execute 推进 goal-24（切片 T-07.2：schedule 注册与运行槽）。goal 目录 goal-24，分支 goal-24，worktree .worktrees/goal-24，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件注册的 interval/cron 定时任务经 REGISTER 校验后，由 Core Scheduler 按 contracts/schedule 计算触发：每次触发以 Core→插件 REQUEST 调用任务（独立 run_id、JOB_TIMEOUT 期限、无 Actor、不默认 Admin），同一 job 有运行中的调用时跳过本次；到期只结束调用结果并取消，运行槽保留到插件真正返回（迟到响应只释放槽）或旧进程确认退出；不补跑错过的 tick。

依据：PLAN T-07 步骤 1（schedule）、3；CONTRACTS §4.3、§8（scheduled 的 Send 必须显式 To）；MESSAGE-MODEL §1.1（scheduled）；WIRE-FORMAT §4；TEST-MATRIX V-20、V-46、V-16 与 V-05 的 scheduled 分项。

范围（只新增或修改这些）：
- contracts/validate/register.go：SCHEDULE 路由的注册校验（调用 schedule.FromSpec）
- internal/wire/session.go：Request 增加迟到响应回调（只加字段与触发点）
- 新增 internal/core/scheduler/
- sdk/sdk.go：删除 Run 中「this SDK version does not support scheduled jobs yet」的拒绝（只改该处）
- sdk/session.go：buildRegister 输出 SCHEDULE 路由
- sdk/dispatch.go：businessHandler 分发 SCHEDULE 路由
- 上述包的 *_local_test.go（不提交）
允许的新跨模块边：internal/core/scheduler -> contracts/schedule、internal/core/router、internal/wire；contracts/validate -> contracts/schedule（同属 contracts，与已允许的 internal/core -> contracts 一致）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/wire
   Request 增加字段：Late func(Response) // 非 nil 时，若响应在调用方已因期限结束之后才到达，读循环调用一次；期限内到达时不调用。Stats.LateResponses 照常计数。
2. internal/core/scheduler
   type Scheduler struct{ ... }
   type Options struct { JobTimeout time.Duration; Now func() time.Time } // JobTimeout 取 tune.Values.JobTimeout；Now 为 nil 时用 time.Now
   func New(opt Options) *Scheduler
   func (s *Scheduler) Attach(p router.Plugin, activatedAt time.Time) // 为插件的 SCHEDULE 路由安排下一次触发；anchor 为激活时刻；同一插件 ID 再次 Attach 时替换为新实例，旧实例仍占用的槽保留
   func (s *Scheduler) Detach(pluginID string)                        // 停止安排新的触发；运行槽保持，直到 Released
   func (s *Scheduler) Released(pluginID string)                       // 旧进程确认退出后调用：释放该插件的全部运行槽
   func (s *Scheduler) Close()                                         // 停止全部定时器；不等待在途调用
   type JobStats struct { Runs, Skipped, TimedOut, Late uint64; Running bool }
   func (s *Scheduler) Stats(pluginID, jobID string) JobStats

行为：
- 注册校验：SCHEDULE 路由的 route_id 为 "job." + job_id；job_id 非空且同插件内唯一；ScheduleSpec 必须恰有 interval_ms 或 cron 之一，经 schedule.FromSpec 校验（interval 至少 1s、cron 五字段、时区合法）；job_deadline_ms 必须为 0，期限统一取 JOB_TIMEOUT（SDK 不提供逐任务期限）。任一项不合法时整次注册失败。
- SDK：buildRegister 为 Registration.Schedules 中每个任务输出 Kind 为 ROUTE_KIND_SCHEDULE、DefaultEnabled 为 true 的路由，Every 填 interval_ms，Cron 填 cron 与 timezone。businessHandler 收到 SCHEDULE 路由的请求时，以 invokeHandler 构造的 Context（没有 Actor 与默认 Target，Send 必须显式 To）调用 Schedule.Handler；ctx 被取消时仍等用户函数真正返回后才发送唯一的 Response。
- 触发：每个 job 用 Trigger.Next(anchor, now) 得到下一次时刻并设定定时器；到点时若该 job 的运行槽空闲则发起一次调用、Runs 加一，否则 Skipped 加一、不排队；之后再按 Next 安排（不补跑，跳过的 tick 不累积）。不同 job 并发运行。插件 Activated() 为 false 时跳过本次且不计 Skipped。
- 调用：直接调用插件 Session，不经事件队列与派发链。Request 为：
  - Target：route_id。
  - Payload：HandlerRequest{route_id, context: InvocationContext{actor 为空, default_target 为空}, event}；event 由 Core 生成：event_id 为新的随机 ID，event_type 为 "scheduled"，source_kind 为 EVENT_SOURCE_KIND_SCHEDULER，payload 为 ScheduledPayload{job_id, run_id（16 字节随机值的十六进制）, scheduled_at_unix_ms}，没有 Actor、Target 与 adapter 字段。编码与 router 的 Handler 调用相同（PROTOBUF、validate.SchemaID、schema 版本 1）。
  - Attach：&router.Invocation{PluginID, Event: 上述 event, DefaultTarget: nil}，使任务内的 core.send 等子调用得到可信上下文，并按后台调用要求显式 To 与发信许可。
  - Deadline：触发时刻 + JobTimeout；Late：释放运行槽。
- 期限：到期时调用方得到 DEADLINE_EXCEEDED、Cancel 已发出，TimedOut 加一，但运行槽保留；迟到响应（Late 回调）到达时释放槽并计入 Late。期限内的响应（成功或错误）立即释放槽；HandlerOutcome 与错误只记日志，不重试。Session 断开时不释放槽，直到 Released 被调用。
- 本切片不接统一启停（Draining 时停止新 tick 等属于 T-07.5），也不改 cmd/alonabot 的装配。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：注册校验的各条拒绝（route_id 前缀、重复 job_id、两种触发都有或都没有、interval 不足 1s、非法 cron 与时区、非零 job_deadline_ms）；wire 的 Late 回调只在期限后触发且只一次；cron 任务用注入时钟的 Next；Detach 后不再触发；Close 释放定时器；internal/core/supervisor/scheduler_local_test.go（//go:build unix）：在临时目录生成并构建一个插件（每 1s 的 interval 任务，按环境变量决定执行时长，以日志记录 run_id；另一个任务在执行中对显式 To 调用 Send），Launch 后 Attach 到 Scheduler（JobTimeout 设为 2s）：执行 0.2s 时，若干秒内有相应次数的调用，run_id 各不相同，事件没有 Actor；不带 To 的 Send 得到 NO_DEFAULT_TARGET；执行 5s 时：第 2s 到期后 TimedOut=1、Running 仍为 true，期间到点的 tick 被 Skipped；第 5s 迟到响应到达后 Late=1、槽释放，下一次 tick 正常调用；任务执行中杀掉插件进程：槽不释放；调用 Released 后释放，新进程 Attach 后从未来的 tick 开始，不补跑。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：Scheduler 与 HTTP 的统一启停、Draining 与停止预算（T-07.5）；生产装配（cmd/alonabot）；Agent 的定时任务（T-12.6）；持久化的运行记录；逐任务的期限。
```

### goal-35 · T-08.9 configdb 示例与 T-08 验收

状态：已完成（[PR #117](https://github.com/AlonaBot/AlonaBot/pull/117)，合并提交 `2c1fdcc`），证据见 TEST-MATRIX §3.2。维护者裁决声明字段为 null 时按现行 CONTRACTS §5.1 拒绝（CONFIG/INVALID_ARGUMENT），原语句的「null（指针字段）全部保留」已改正，本地两项 `declared_null_contract_conflict` 保持 SKIP。BLOCK-02 的修复经维护者授权修改 `internal/core/process`：Darwin 上对只剩僵尸或正在退出的进程组返回 EPERM 时继续由 `WaitTree` 等待回收。启动记录改写独立表 `configdb_startup`，避免覆盖用户的同名笔记。派发时：前置 T-08.6（PR #85）、T-08.7（goal-34，PR #91）、T-08.8（PR #102）与 T-08.11（goal-95，PR #103）均已合并。T-08.8（PR #102，合并提交 `f685c92`）已合并，合入的符号：`contracts/i18n` 的 `Parse(files) (*Catalog, []Problem)`、`Catalog.Has`/`Format(locales, key, args)`、`Placeholders`；`sdk.Context.T`、后台 `sdk.T` 与 `Generated.I18n`；`internal/core/i18n.Messages()`；`host.Instance.Messages`/`DefaultLocale`；`router.Plugin` 的 `Messages()`/`DefaultLocale()`；`supervisor.Spec.I18n`/`DefaultLocale`；`Bootstrap.default_locale = 12`。T-08.6（PR #85）已合并。goal-95（PR #103，合并提交 `c840af7`）已合并，合入的接口：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)` 与 `Schema` 的 `Canonical`、`Hash`、`Version`、`Migrate(config, from)`、`Complete`、`Validate`、`SecretPaths`、`CheckI18n(config)`；`sdk/config` 的 `Prepare(ctx, load, schema, raw, fromVersion, targetVersion, restart)`、`Loader`、只读的 `Values`（`String`/`Int`/`Float`/`Bool`/`Strings`/`Object`/`Objects`/`Dict`/`Template`/`TemplateList`）与 `Template.Key`；`sdk.Base.Config() config.Values` 与根插件的 `OnConfigLoad(ctx, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`；`supervisor.Spec.ConfigSchema` 为 config_schema.json 的规范 JSON。生成器读取手写 `config_schema.json`，配置标签解析与 `config.Config`/`config.Base` 已删除。goal-34（PR #91，合并提交 `9789fc8`）已合并，合入的符号：`src/schema/model.ts` 的 `fieldsFromSchema`、`Field`、`Template`；`src/schema/patch.ts` 的 `buildPatch(original, edited, fields)`、`MissingSecretsError`、`isSecretSet`；`src/i18n/catalog.ts` 的 `pickLocale`、`fieldText`、`templateText`；`src/i18n/UI.tsx` 的 `UIContext`、`useUI`；`src/components` 的 `ConfigForm`（props 为 `schema`、`value`、`messages`、`onSubmit(patch)`）与 `LanguageSwitch`；四个手写 fixture 在 `web/admin/fixtures/`。前端纯函数测试只写在被忽略的 `*.local.test.ts`，`npm run test` 为 `vitest run --passWithNoTests`。2026-11-09 审查者按 main `3dec9a0` 的代码核对并改写：配置 Schema 是手写的 `config_schema.json`（D-85），不再由 Go 结构生成；CLI 为 `alonabot plugin generate`/`plugin check`；数据库入口为 `sqlitedb.Open(ctx)`、`sqlitedb.RunVersioned`、`gormdb.Open(ctx)` 与 `gormdb.RunVersioned`；候选的准备与丢弃为 `candidate.PrepareDatabase` 与 `candidate.Discard`，Core 侧经 `supervisor.Launch(ctx, store, Spec{..., ConfigSchema, I18n, NewCandidate, ExpectedOld, Secrets, KeptSecretPaths})` 启动真实插件进程；开发入口 `alonabot run` 只启动单个插件，候选配置固定为 `{}`、数据库为 NONE，不装配配置、Secret 与受管数据库（库模式与开发组件的配置归 T-13.1、T-13.2），因此本切片的跨进程验收只能经 `supervisor.Launch`，按 D-86 写在本地测试，不提交 `cmd/alonabot` 的 e2e。

```text
/goal 使用 $alonabot-execute 推进 goal-35（切片 T-08.9：configdb 示例与 T-08 验收）。goal 目录 goal-35，分支 goal-35，worktree .worktrees/goal-35，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：新增 examples/plugins/configdb，覆盖配置迁移、显式空值、未知与 secret 字段、CRUD 与事务、versioned 迁移与失败、候选 OnStart 写库后失败；在临时 data_dir 上分层验证，数据库部分 sqlitedb 与 gormdb 各跑一遍；用 web/admin 的 fixture 验证表单的字段顺序、hidden、i18 回退、掩码与 PATCH；在 TEST-MATRIX §3.2 记录 T-08 的组件证据，并单独列出需要 T-09 正式入口与 Journal 才能关闭的场景。

依据：PLAN T-08 步骤 9 与「停止和失败条件」；CONTRACTS §5、§6；TEST-MATRIX V-23～V-26、V-38、V-39、V-54 的 T-08 分项。

范围（只新增或修改这些）：
- 新增 examples/plugins/configdb/（含 plugin.yaml、手写 config_schema.json、i18n/、OnConfigLoad 与 versioned 迁移，`//go:generate go run ../../../cmd/alonabot plugin generate` 生成的 alonabot_gen.go）
- web/admin/fixtures/：加入 configdb 的 config_schema.json 副本（configdb.config_schema.json）与对应的 i18n 文案 fixture
- spec/TEST-MATRIX.md §3.2：追加 T-08 的证据行（本切片获授权修改这一处）
- internal/core/supervisor/t08_local_test.go（//go:build unix，不提交）
允许的新跨模块边：无。

接口：不新增公开 Go 符号。

验收矩阵（各为一个子测试）：
1) 配置迁移：config_version 1 的旧配置只有经 OnConfigLoad 才满足 2 的 Schema；先迁移再校验，结果写入候选快照。
2) 显式空值：false、0、""、[]、{} 全部保留；缺失键补默认值；未知键与嵌套未知键（含未知字段的 null）保留在快照中。声明字段为 null 时按 CONTRACTS §5.1 拒绝（CONFIG/INVALID_ARGUMENT；维护者裁决），原「null（指针字段）全部保留」不再作为要求。
3) secret：快照只有引用；Schema 移除 secret 字段但候选保留该键时仍为引用；日志中没有明文。
4) DB：CRUD 与事务（tx 回调）；versioned 两步迁移；第二步失败时账本只有第一步、候选失败、Active 不变。
5) 候选：reload 后 OnStart 写入候选库再失败：Discard，Active 的数据与配置不变，旧版重启正常；新旧进程没有同时写同一个库（runtime.lock 证明）。
6) 两个入口：同一插件同时用 sqlitedb.Open 与 gormdb.Open 时共用同一连接，一边写入另一边立即可读；只用 sqlitedb 的示例变体不链接 gorm.io/gorm（`go list -deps` 核对）。
7) i18n：Context.T 的语言选择与回退；`alonabot plugin check` 拒绝缺少 en-US 与占位符不一致。
8) 表单（本地 web/admin/src/schema/configdb.local.test.ts，不提交，PR 中写明结果）：configdb 的 Schema 字段顺序与声明一致；hidden 不进 patch；未修改的 secret 不进 patch；i18 标签回退。
- 证据：TEST-MATRIX §3.2 追加 T-08 一行；需要 T-09 正式 reload、WebUI/API 与 Journal 崩溃恢复才能关闭的 V-23～V-25、V-38 场景单独列为「未闭合」，不标成通过。

测试（按编码规范「测试提交规则」）：
- 提交：无。理由：dev 入口不装配配置与受管数据库，跨进程验收只能经 supervisor.Launch，D-86 不允许提交这类测试；示例插件本身随切片提交。
- 本地自测（不提交，PR 中写明结果）：internal/core/supervisor/t08_local_test.go 以真实 configdb 进程、临时 data_dir、真实 state.Store 与 secret.Store 跑验收矩阵 1)～7)，数据库部分 sqlitedb 与 gormdb 各跑一遍；npm run build；示例的 `alonabot plugin check`。

完成条件：gofmt -l 无输出；`go generate ./...` 后无未提交的生成差异；go vet ./...；go test -race ./...；go build ./...；web/admin 下 npm run test 通过。

不做：正式 reload 与管理入口（T-09）；Journal 与崩溃恢复（T-09.5～T-09.7）。
```

### goal-101 · T-08.12 Adapter 实例配置与 Core 设置的版本迁移

状态：已完成（[PR #114](https://github.com/AlonaBot/AlonaBot/pull/114)，合并提交 `a796766`），证据见 TEST-MATRIX §3.2。按 D-117 使用 `adapter_config_schema.json`；Telegram 声明按 goal-117 合并后的 `Start` 补了 `proxy_mode`（缺省 `none`，`proxy` 带 `custom` 条件）；CONTRACTS §8 列出的 `file_base_url` 当前 `Start` 未读取，暂未声明；审查中修复同毫秒备份重名，备份名同名时顺延 1 毫秒。派发时：前置 T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）与 T-08.11（goal-95，PR #103，合并提交 `c840af7`）均已合并。依据 D-85 新增；2026-11-09 审查者按 main `3dec9a0` 的代码写出接口、行为与测试段。2026-11-10 按 D-117，文件名由 `instance_schema.json` 改为 `adapter_config_schema.json`，语句中的文件名已随之修改；PR #114 在合并前改用新名字。已核对：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)`（同时检查 `$version` 与 `$migrations`）、`Schema.Version() int`、`Migrate(config map[string]any, from int) error`、`Complete(config map[string]any)`、`Validate(config map[string]any) error`、`SecretPaths() [][]string`；`internal/core/settings` 的 `Settings.Adapters map[string]AdapterSettings{Version int; Instances map[string]json.RawMessage}`（实例配置已展开 `${...}`）、`File.Update(edit func(*Document) error)` 以磁盘原文为底稿并保留 `${...}` 原写法、`readVersion` 只接受 `currentVersion = 1`；`remote.Process.StartInstance(ctx, instanceID, configJSON)` 启动实例，但目前没有装配处从 `Settings.Adapters` 启动实例（归 goal-39 与 goal-9）；Telegram 在 `New().Start` 中解析 `token`（必填）、`api_base_url`（缺省 `https://api.telegram.org`）与 `proxy`；`alonabot plugin generate`/`check`（`sdk/generate.Scan`）只处理插件工程，`alonabot build`/`pack` 尚未实现。goal-117 已于 [PR #125](https://github.com/AlonaBot/AlonaBot/pull/125)（合并提交 `f3af986`）合并，`Start` 现在解析 `proxy_mode`。2026-11-09 新增的 goal-117（D-101）给 Telegram 实例配置追加 `proxy_mode`：goal-117 先合并时，本切片的 Telegram adapter_config_schema.json 按执行顺序 §3 一并声明它（options none、core、custom，缺省 none，proxy 带 condition）。

```text
/goal 使用 $alonabot-execute 推进 goal-101（切片 T-08.12：Adapter 实例配置与 Core 设置的版本迁移）。goal 目录 goal-101，分支 goal-101，worktree .worktrees/goal-101，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Adapter 包用 adapter_config_schema.json 声明实例配置（格式与 config_schema.json 相同），Core 按它迁移、补缺失值并校验 core_config.json 中的实例；core_config.json 自身的 $version 由 Core 内置的迁移处理。写回前备份，失败与降级时报错，不丢配置。官方 Telegram Adapter 补上 adapter_config_schema.json。

依据：PLAN T-08 步骤 1；CONTRACTS §5.5、§8、§15.1；DECISIONS D-85；TEST-MATRIX V-38 的迁移分项、V-88。

范围（只新增或修改这些）：
- internal/core/settings/：新增 instances.go（实例配置的准备与试算）与 migrate.go（Core 设置的内置迁移与备份）；settings.go 的 readVersion 改为经 migrate.go 处理
- sdk/generate/：新增 instance.go，Scan 在 --dir 中有 adapter_config_schema.json 时检查它
- cmd/alonabot/plugin.go：--dir 只有 adapter_config_schema.json、没有插件根类型时 check 只做这一项检查
- 新增 adapters/telegram/cmd/alonabot-adapter-telegram/adapter_config_schema.json
- internal/core/settings/instances_test.go；其他 *_local_test.go（不提交）
允许的新跨模块边：无（internal/core 引用 contracts 已在依赖表中）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/settings
   type InstanceProblem struct { InstanceID string; Path string; Reason string } // Path 为 JSON 路径；Reason 不含取值
   func (f *File) PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []InstanceProblem, error)
   // 对该 Adapter 的全部实例执行迁移、补缺失值与校验；返回可启动实例的完整配置。存储版本低于 schema 时，迁移全部成功后才备份并经 Update 写回迁移结果与新的 $version（写回的是迁移后的原文结构，${...} 写法不展开）；有任一实例失败时不写回，失败实例列入 problems，其余实例照常返回。error 只用于写回或备份失败。
   func (f *File) TrialInstances(adapterID string, schema *configschema.Schema) []InstanceProblem
   // 只读试算：不写回、不备份，供 Adapter 更新的预检在停止旧版本前调用（T-09.5 接入）。
   const BackupDir = "backups" // <data_dir>/backups/core_config.<旧的 Core $version>.<UTC 时间，20060102T150405.000Z>.json
2. sdk/generate
   Scan 对 adapter_config_schema.json 调用 configschema.Parse，问题按 config_schema.json 相同的 Diagnostic 格式（文件:JSON 路径）报告；不嵌入生成代码。

行为：
- 实例配置：core_config.json 中 `adapters.<adapter_id>.$version` 是实例配置的存储版本（缺省 1）。存储版本高于 schema 的 `$version` 时整组失败，problem 写明两个版本号；低于时对每个实例按 `$migrations` 迁移，再 Complete 与 Validate。secret 字段在写回时保留原写法，不在日志与 problem 中输出取值。
- 备份：任一写回前先把磁盘上的原文件复制到 BackupDir（目录 0700、文件 0600，名称如上）；复制失败时不写回并返回 error。
- Core 设置：settings 内置按版本排列的迁移列表（当前为空，`currentVersion` 仍为 1）。readVersion 对高于 currentVersion 的版本拒绝加载，错误写明两个版本号与文件路径；低于时按列表迁移、备份并写回。迁移失败时 Load 返回错误，原文件不变。
- Telegram：adapter_config_schema.json 声明 `token`（string、secret、required）、`api_base_url`（string，缺省 `https://api.telegram.org`）与 `proxy`（string，缺省空），`$version` 为 1；与 `New().Start` 的现有解析一致，不改 Adapter 代码。
- CLI：`alonabot plugin check --dir adapters/telegram/cmd/alonabot-adapter-telegram` 只检查 adapter_config_schema.json 并以 0 退出；有问题时逐行输出诊断并以 exitFailure 退出。generate 对没有插件根类型的目录照旧报错。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/settings/instances_test.go（真实文件系统上的 core_config.json）：实例配置 v1 → v2 的 rename 迁移写回、$version 更新、`${...}` 原写法保留、backups/ 下生成备份；一个实例迁移后校验失败时文件与备份目录都不变、其余实例照常返回；存储版本高于 schema 时整组失败；core_config.json 的 $version 高于程序时 Load 失败且文件不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：map、wrap、remove 与 `*` 路径；secret 不进 problem；TrialInstances 不写文件；备份失败时不写回；sdk/generate 对 adapter_config_schema.json 的各类诊断；Telegram 的 adapter_config_schema.json 经 Parse 与 Complete 后与 Start 的缺省一致。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/settings/... ./sdk/... ./cmd/alonabot/... ./adapters/...；go build ./...。

不做：从 Settings.Adapters 启动实例的装配（goal-39、goal-9）；Adapter 更新的预检与提交（T-09.5）；`alonabot build`/`pack` 与包内携带 adapter_config_schema.json（T-14.3）；Admin API 的实例编辑（T-09.14）。
```

### goal-67 · T-13.1 库模式入口与同进程组件

状态：待派发。前置 T-02.1、T-03.3、T-04.2、T-04.9（PR #110）、T-08.8（PR #102）、T-09.16（PR #140）均已合并；T-03.8（PR #136）与 T-08.13（PR #137）也已合并。2026-11-12 按维护者决定 D-134 补齐 Adapter 工程入口、生成登记与固定代次；审查者按 main `896991b` 核对现有符号、文件与签名。D-134（[PR #142](https://github.com/AlonaBot/AlonaBot/pull/142)，合并提交 `1e71494`）已合并。宜在 T-04.8（goal-91，[PR #141](https://github.com/AlonaBot/AlonaBot/pull/141)）合并后派出：两者都改 `sdk/sdk.go`、`adapter_run.go`、`session.go` 等文件，派发前审查者按合入的 T-04.8 代码核对语句中的 sdk 符号。

```text
/goal 使用 $alonabot-execute 推进 goal-67（切片 T-13.1：库模式入口与同进程组件）。goal 目录 goal-67，分支 goal-67，worktree .worktrees/goal-67，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：根包 alonabot 提供 Run、Start 与 Plugin、Adapter、DataDir、Args 选项，在当前进程装配一个 Core。已安装包沿用独立进程；开发插件与 Adapter 在 goroutine 中经 loopback TCP 走完整的握手、注册、激活与停止。组件身份取自生成器嵌入的工程 manifest，来源为 dev、固定代次为 dev:<组件 ID>。Adapter 的工程入口与资源登记归可导入包，库模式与独立进程共用身份。Start 在全部组件激活后返回句柄，ctx 取消时按停止预算停止；Run 加上 SIGINT/SIGTERM 处理。组件在预算内未退出时 Wait 与 Run 返回错误。

依据：spec/plan/core.md T-13 步骤 1、2 及失败停止条件；CONTRACTS §15.5、§15.4、§4.2、§5.5、§8、§9、§16；DECISIONS D-47、D-77～D-79、D-86、D-117、D-134；docs/execution-order.md §3 的根包、cmd/alonabot/main.go 与 sdk/generated.go 各行；ARCHITECTURE 的根包与依赖表；TEST-MATRIX V-64 的生命周期分项。

范围（只新增或修改这些）：
- 根包 alonabot.go（Run、Start、Core 与选项）、assembly.go（Core 装配，供库模式使用，后续 alonabot run 复用）。
- 新增 internal/inproc/launch.go：由根包向 SDK 传递启动参数，只使用标准库类型。
- sdk/：generated.go 增加 Manifest、AdapterConfigSchema 与 Adapter 登记、资源读取入口；新增 inproc.go；sdk.go、adapter_run.go、session.go、runtimelock.go、activate.go 及 sdk/ 下读取当前 Session、调用上下文、配置准备、HTTP 与日志的相关调用处，接入显式启动参数与同进程会话隔离。sdk/internal/dbhook/ 与 sdk/sqlitedb/open.go、sdk/gormdb/open.go 同步 Session 归属，两个数据库包现有 Open(ctx) 签名保持不变。
- sdk/generate/：检查并嵌入工程 plugin.yaml，生成 Adapter 资源登记并嵌入实例 schema；保留 ConfigSchema、I18n、ServiceSchemas 与文案函数检查。cmd/alonabot/plugin.go 同步生成结果的 Adapter 分支；本切片沿用 plugin generate/check 命令。
- adapters/telegram/：新增 plugin.yaml、生成指令与 alonabot_gen.go；将 adapters/telegram/cmd/alonabot-adapter-telegram/adapter_config_schema.json 原样移动到 adapters/telegram/adapter_config_schema.json。cmd/alonabot-adapter-telegram/main.go 继续作为调用 sdk.RunAdapter(telegram.New()) 的薄 main（当前已符合，只在必要时调整）。当前没有工程 plugin.yaml，需在可导入包内新增。
- internal/core/state/：新增安装记录枚举与 dev 摘要刷新入口；复用安装、启用状态、代次与 ActiveGeneration 接口，状态库表结构保持不变。
- 新增 examples/devbot/main.go、examples/devbot/plugin/ 与 examples/devbot/adapter/：主程序导入两个组件包，用信号派生的 ctx 调用 Start，成功后输出就绪日志并 Wait；插件与回环 Adapter 包各有 plugin.yaml、生成指令及 alonabot_gen.go，Adapter 包另有 adapter_config_schema.json。它们是后继开发运行示例的基础。
- 根包 alonabot_test.go；上述包的 *_local_test.go（不提交）。
允许的新跨模块边：根包 -> internal/inproc、sdk/... -> internal/inproc（已登记）；根包 -> internal/core/...、sdk/...、contracts，examples/devbot -> 根包、sdk/...（已登记）。internal/inproc 只依赖标准库。ARCHITECTURE 的同步由审查者负责。

现有接口（已按 main 核对，以下均为现有符号）：
- contracts/manifest：ParseProject(data []byte) (*Result, error)，Result 含 Manifest *Manifest 与 Warnings []string；从 result.Manifest.ID、Role、AdapterID 读取身份，Role 缺省为 plugin。Parse(data []byte) (*Result, error) 校验含完整构建信息的包内 manifest；Generate(project *Manifest, build BuildInfo) ([]byte, error) 供 build/pack 合成包内 YAML。
- sdk.Generated 当前字段为 ServiceSchemas map[string][]byte、I18n map[string][]byte、ConfigSchema []byte。UseGenerated(root Root, g Generated) 按根类型的 reflect.Type 登记；generatedFor(root Root) Generated 是包内函数。Run(root Root) error 与 RunAdapter(a adapter.Adapter) error 为现有进程入口，adapter 指 sdk/adapter。
- sdk/generate：Scan(ctx context.Context, opt Options) (*Result, error)；Options 含 Dir、GOOS、GOARCH、Tags；generate.FileName 为 alonabot_gen.go。Result 当前含 RootType、ConfigSchema、Source、Diagnostics；有实例声明但没有插件根类型时仅检查，RootType 为空、Source 为 nil。
- router：New(opt Options) *Router、NewCaller(pluginID, generationID string) Caller；(*Router).Attach(p Plugin, c Caller)、Close(ctx context.Context) error。Options 已有 Concurrency、QueueEvents、QueueBytes、Global、Adapters、Send、Prepare、Prefixes、Buttons、Policy。
- remote：New(opt Options) *Process；(*Process).Bind(ctx context.Context, in *host.Instance, caller router.Caller) error、StartInstance(ctx context.Context, instanceID string, configJSON []byte) error、Unbind()。Options 含 AdapterID、PluginID、Store、Registry、Sink、RegisterIncoming。
- supervisor：Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error)；(*Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult。Spec 的 Role、AdapterID、ConfigSchema、I18n、DefaultLocale、Handler、Global、Scheduler、HTTP 均已存在；Running.Instance 是 *host.Instance，Running.Process 是 *process.Process，Launch 始终创建独立进程。
- host：Prepare(ctx context.Context, conn *wire.Conn, exp Expect, boot *pb.Bootstrap, lim Limits) (*Prepared, error)、Activate(ctx context.Context, conn *wire.Conn, prep *Prepared, opt ActivateOptions) (*Instance, error)、NewStopTiming(v tune.Values) StopTiming。
- tune：Parse(args []string) (Values, []Override, error)。state：Open(ctx context.Context, dataDir string) (*Store, error)；(*Store).AddInstall(ctx context.Context, in Install) error、GetInstall(ctx context.Context, pluginID string) (Install, error)、AddGeneration(ctx context.Context, pluginID, generationID, configSnapshot string, configVersion uint32) error、CommitActive(ctx context.Context, pluginID, expectedOld, newGen string) error、Active(ctx context.Context, pluginID string) (string, error)、GenerationConfig(ctx context.Context, generationID string) (string, error)、SetEnabled(ctx context.Context, pluginID string, enabled bool) error、Enabled(ctx context.Context, pluginID string) (bool, error)。Install.Source 已是 string，直接写 "dev"；Role 用 state.RolePlugin/RoleAdapter，AdapterID 来自 manifest。
- settings：Load(dataDir string) (*File, error)、(*File).Settings() Settings、(*File).PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []InstanceProblem, error)。设置与 Adapter 实例的接入归后继切片；本切片按执行顺序不读 core_config.json。

本切片要求新增的接口（照此实现；内部辅助函数自定）：
1. 根包 alonabot
   type Option func(*options)
   type Core struct{ ... }
   func Run(opts ...Option) error
   func Start(ctx context.Context, opts ...Option) (*Core, error)
   func (c *Core) Wait() error
   func Plugin(root sdk.Root) Option
   func Adapter(a adapter.Adapter) Option // adapter 指 sdk/adapter。
   func DataDir(dir string) Option // 缺省为工作目录下的 data/。
   func Args(values ...string) Option
   // 此处按 main 的 sdk.Root；插件对象 API 由 T-04.12 同步调用方。
2. internal/inproc
   type Launch struct { PluginID string; Secret []byte; Addr string; BuildDigest []byte; RuntimeLockPath string }
3. sdk
   func RunInProcess(root Root, l *inproc.Launch) error
   func RunAdapterInProcess(a adapter.Adapter, l *inproc.Launch) error
   func UseGeneratedAdapter(a adapter.Adapter, g Generated)
   func GeneratedFor(component any) (Generated, bool)
   Generated 增加 Manifest []byte // 工程入口包内 plugin.yaml 的原文。
   Generated 增加 AdapterConfigSchema []byte // adapter_config_schema.json 的规范 JSON；插件为 nil。
   // 生成 init 传入具体类型的 nil 指针；UseGenerated 保留现有根类型键，UseGeneratedAdapter 按去掉指针后的具名类型登记，重复登记 panic。
   // GeneratedFor 按组件类型读取；Adapter 的值与指针读取同一登记，bool 表示是否登记。原 generatedFor 委托此读取，保留现有调用。
4. sdk/generate
   Result 增加 AdapterType string // 导入路径.具体类型名；插件或仅检查实例声明时为空，RootType 保留插件含义。
   // Scan、Options 与 FileName 保持现有签名和含义，Source 同时承载插件与 Adapter 的生成结果。
5. internal/core/state
   func (s *Store) Installs(ctx context.Context) ([]Install, error) // 按 install_seq 递增，供冲突检查与启动已启用包。
   func (s *Store) SetDevBuildDigest(ctx context.Context, pluginID string, digest []byte) error
   // 仅刷新已存在的 source=dev 安装记录的摘要；不改变身份、安装顺序、启用状态或代次，缺记录或来源不符时返回错误。

行为：
- Adapter 工程入口：以可导入包为生成器 Dir。先用 manifest.ParseProject 校验该目录的工程 manifest，role 为 adapter 时，在遵守 GOOS/GOARCH、build tags 且排除生成文件与测试文件的入口包中，用 go/types 查找唯一实现 sdk/adapter.Adapter 的具体具名类型（含指针方法集）；零个或多个时给出带位置的诊断。无需导出该类型，生成 init 调用 sdk.UseGeneratedAdapter((*具体类型)(nil), sdk.Generated{...})，嵌入 Manifest 与 AdapterConfigSchema，并按已有规则嵌入 I18n；不执行构造函数、不扫描注册标记。插件分支沿用 sdk.Base 根类型发现与 UseGenerated。
- 生成结果：有工程 plugin.yaml 时先 ParseProject，再原样嵌入 Generated.Manifest；缺少该文件时沿用当前独立进程工程的生成/实例声明检查，库模式在登记校验时报缺失。role=adapter 的工程要求有效的 adapter_config_schema.json。cmd/alonabot/plugin.go 根据 RootType 或 AdapterType 判断是否有可写入/核对的 Source，Adapter 的 plugin generate/check 同样写入/比较 alonabot_gen.go；仅检查实例声明的目录仍保持 check 行为。保留 ConfigSchema、I18n、ServiceSchemas 及 T-08.13 的常量文案键/参数检查。Telegram 与示例中的生成指令通过现有 plugin generate 命令生成，生成产物只由生成器写入。
- 启动前校验：通过 sdk.GeneratedFor 取得组件资源，对 Manifest 调用 manifest.ParseProject，保留解析错误与 Warnings，使用 ID、Role、AdapterID。缺少登记/Manifest、选项与角色不符、重复 ID/adapter_id 或与已安装包冲突时，在启动任何组件前返回错误。已有同 ID、同角色、同 adapter_id 的 dev 记录在重启时复用，其他冲突报错。工程作者字段沿用 ParseProject；BuildInfo 与 Generate 的包内构建信息留给 build/pack。
- Telegram 的工程 manifest 放在 adapters/telegram，adapter_id 为现有 AdapterID（telegram），组件 ID 使用 alonabot_adapter_telegram；schema 移动时保留现有字段、版本与内容。cmd 只调用 sdk.RunAdapter，库模式从同一个 telegram.New() 的具体类型取得包内登记。build/pack 从可导入包取资源并构建 cmd 薄入口的接入由 T-09.3、T-14.3 完成（D-134）。
- 参数与装配：ALONABOT_ARGS 有值时作为首个输入，再追加 Args 的各项，统一调用 tune.Parse；值含 = 时按逗号分隔，否则作为文件路径。使用 tune.Values 与 Override 记录覆盖 WARN，沿用日志与内存软上限的契约。state.Open 取得 core.lock 后装配 adapter.Registry、Router、message.Service、固定服务 Mux 等已合入组件。
- Router 使用 New(Options)，并发、队列与全局预算取 tune.Values；媒体与按钮接入 Prepare、Buttons。Prefixes 传 nil，Core Admin 列表为空，DefaultLocale 使用 en-US，三者由 T-09.4 接入。每个 Session 的 Router.Attach、Mux.For 与 remote.Process.Bind 使用同一个 router.Caller；remote.Options.RegisterIncoming 接媒体 Store 的入站登记入口。
- 已安装包：枚举 Installs，按 Enabled 启动已启用的非 dev 包，经 supervisor.Launch 使用其已提交的 ActiveGeneration 与包资源；当前选项中未登记的旧 dev 记录保留，不自动启动。同进程组件经 loopback 复用 host.Prepare/Activate，退出等待由根包装配持有；Running.Process 仍只描述独立进程。新装配供后继复用；cmd/alonabot 当前 run 只启动单个插件，其调用方迁移归 T-09.4、T-13.3。
- 开发记录：Source 为 "dev"，Entrypoint 为空，按 manifest 填 Role 与 AdapterID；初建使用 AddInstall，重启时经 SetDevBuildDigest 刷新本次随机 BuildDigest，该值与 host.Expect、SDK 启动参数一致。代次 ID 固定为 "dev:" + manifest.ID（Adapter 使用包 ID），沿用现有全局主键与外键；首次建立 AddGeneration 并在激活前 CommitActive，重启复用同一代次。激活前失败后再次启动时也复用已存在的代次行，保留真正的数据库错误；不创建候选目录、不走 Journal、不改表结构、不重建按钮密钥，按钮 token 重启后仍有效（D-134）。
- 启动参数：每个组件有独立 Session Secret、loopback 地址、BuildDigest 与绝对 RuntimeLockPath（<data_dir>/plugins/<id>/runtime.lock），由 inproc.Launch 传入；SDK 的启动读取、取锁与运行流程共用内部逻辑。同进程启动使用显式参数，环境变量与工作目录保持进程级含义。
- 同进程隔离：现有 sdk.activeSession/currentSession 与 sdk/internal/dbhook.current 是进程级状态，改为按组件 Session 保存生命周期、调用、文案与数据库归属；通过 SDK 传给生命周期、Handler、HTTP 与迁移的 ctx 及其派生 ctx 选择所属组件，停止一个组件只清理它自己的状态。现有 SDK 公开调用签名保持不变；独立进程入口沿用默认 Session 用法，同进程后台调用使用所属组件的生命周期 ctx，缺失归属时返回现有 FAILED_PRECONDITION，不选取另一组件的 Session。完整的开发配置与 dev.sqlite 接入归 T-13.2。
- 配置与实例：本切片开发插件 Bootstrap 的配置为 "{}"，数据库为 NONE。开发 Adapter 的实例按 PLAN T-13 步骤 3 由 T-13.2 从 core_config.json 的 adapters.<adapter_id>.instances 读取，经 settings.PrepareInstances 后交给 remote.Process.StartInstance；本切片完成 Adapter 包 Session 激活并提供嵌入的实例 schema。
- 停止：全部已安装包与同进程组件从同一个 t0 并行停止，共享 host.NewStopTiming(tune.Values) 的总截止，冻结、排空、STOPPING 与资源关闭遵循 §4.2。到总截止仍未退出的同进程组件记含组件 ID 的 ERROR，Wait 与 Run 返回错误；全部正常退出后释放资源与 core.lock。激活前失败时停止已启动的部分并返回完整错误。
- SDK 同进程入口复用 Run/RunAdapter 的协议与生命周期，使用 Core 已设置的默认 Logger，避免同进程入口改写全局 Logger；组件级日志字段与级别按 Session 绑定。未捕获的 panic 使进程退出，不执行有限重启。

测试（按编码规范「测试提交规则」）：
- 提交：仅根包 alonabot_test.go 的 TestLibraryLifecycle 一个主路径集成测试（//go:build unix）：构建 examples/devbot，在临时 data_dir 下以子进程运行；真实 Core 与生成登记的插件、Adapter 经 loopback 完成激活，核对两条 source=dev 记录及各自的 dev:<组件 ID> 代次；发送 SIGTERM，进程在共同 STOP_TOTAL 内成功退出。Adapter 实例收发随 T-13.2 接入后验证。
- 本地自测（*_local_test.go，不提交，PR 写明结果）：缺少登记/Manifest、manifest 不合法、角色不符、重复 ID/adapter_id 与已安装包冲突；Adapter 具体类型零个/多个、未导出类型登记、Adapter 值与指针读取同一登记、GOOS/GOARCH/tags、仅实例声明目录的原有行为；Telegram 生成资源与 cmd 独立进程、库模式取得同一身份；ALONABOT_ARGS 与 Args 的先后；激活失败后清理并再次启动；多个组件的 Session 归属、调用与独立停止；dev 记录跨启动复用、失败后复用代次、重启后按钮 token 仍有效；ctx 取消后 Wait；停止超时返回错误与日志；未捕获 panic 退出；已安装包与同进程组件混合运行；生成器保留 Schema、I18n、ServiceSchemas 与文案检查。sdk/ 与 internal/core/supervisor/ 的相关验证全部在本地完成，不新增 testdata 测试插件（D-86）。

完成条件：gofmt -l 无输出；go generate ./... 后生成产物与生成源一致；go vet ./...；go test -race ./...；go build ./...。

后继接入：config.json、dev.sqlite、开发 Adapter 实例与 dev 来源的 update/rollback/uninstall 限制归 T-13.2；目录构建与热部署归 T-13.3；管理与设置装配归 T-09.4；顶层 build/check/pack 的 Adapter 入口与资源读取归 T-09.3，发行构建与打包归 T-14.3，沿用 D-134。
```

### goal-123 · T-08.13 文案函数只返回字符串

状态：已完成（[PR #137](https://github.com/AlonaBot/AlonaBot/pull/137)，合并提交 `d74b63d`），证据见 TEST-MATRIX §3.2。合入的符号：`c.T(key, args ...any) string`、`sdk.T(ctx, key, args ...any) string`、`contracts/i18n` 不返回 error 的取文案函数；生成器检查以常量写出的键与参数名（参数名须为 `string` 类型），诊断经 `plugin generate` 与 `plugin check` 输出；审查者决定 build 按当前的 `plugin generate` 验收。审查期间按 FIX-1（多返回值实参使生成器越界）与 FIX-2（带类型的参数名常量）修改。派发时：前置 T-04.9（[PR #110](https://github.com/AlonaBot/AlonaBot/pull/110)，合并提交 `6aed25d`）与 T-08.8（[PR #102](https://github.com/AlonaBot/AlonaBot/pull/102)，合并提交 `f685c92`）均已合并。2026-11-09 按 D-106 新增，接口按 goal-53 语句中合入的符号写出。与同在第 10 波的 goal-101 都会修改 `sdk/generate`，按执行顺序 §3 的生成器一行处理。

```text
/goal 使用 $alonabot-execute 推进 goal-123（切片 T-08.13：文案函数只返回字符串）。goal 目录 goal-123，分支 goal-123，worktree .worktrees/goal-123，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件取文案时不再处理错误，`c.T` 与 `sdk.T` 只返回字符串。代码里以常量写出的键与参数名改在 `alonabot check` 与 `build` 时检查；运行时仍有问题的，返回键名或保留占位符原文，并在插件日志记 WARN（D-106）。

依据：PLAN T-08 步骤 10；CONTRACTS §5.4；DECISIONS D-106；TEST-MATRIX V-104。

范围（只新增或修改这些）：
- contracts/i18n/：增加不返回 error 的取文案函数
- sdk/：T 所在的文件（goal-53 合入）
- sdk/generate/：新增对 T 调用的检查，接入 check 与 build 已有的诊断输出
- 调用方：仓库内全部 c.T 与 sdk.T 的调用（示例、测试插件与以字符串嵌入的插件源码）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. contracts/i18n
   func (c *Catalog) Text(locales []string, key string, args map[string]string) (text string, problems []string) // 未知键时 text 为键名；缺少的参数保留占位符原文；多余的参数忽略；problems 依次用英文描述问题，不含参数值，没有问题时为 nil
2. sdk
   func (c *Context) T(key string, args ...any) string // 取代 (string, error)；args 仍为成对的名字与值，值用 fmt.Sprint 格式化
   func T(ctx context.Context, key string, args ...any) string // 取代 (string, error)；没有 Actor 时从 Core 默认 locale 开始

行为：
- 运行时：T 经 Catalog.Text 取文案，语言选择顺序不变（CONTRACTS §5.4）。problems 非空或 args 不成对时（忽略最后一个落单的名字），在插件日志记 WARN（key 与问题，不记参数值），仍返回取到的文本；同一 key 的同一问题在一个 Session 内只记一次。插件没有 i18n 资源时返回键名，同样记 WARN。
- 构建期：生成器在类型检查后找出对 `(*sdk.Context).T` 与 `sdk.T` 的调用；key 是字符串常量时检查：
  - en-US 中有该键；
  - 其余参数个数为偶数，名字位置上都是字符串常量，名字集合与该键的占位符集合相同。
  问题按 `文件:行:列: 说明` 输出，check 与 build 失败。key 不是常量的调用不检查，留给运行时。
- Core 自己的文案目录（internal/core/i18n）仍用严格的 Format，不变。

测试（按编码规范「测试提交规则」）：
- 提交：不新增；已有调用 T 的集成测试改为单返回值，断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Catalog.Text 的未知键、缺少参数、多余参数与语言回落；生成器对缺键、参数不成对、名字不是常量、名字集合不符的诊断与位置，key 不是常量时不报；运行时 WARN 的去重且不含参数值；没有 i18n 资源的插件调用 T。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./sdk/...；go build ./...。

不做：Core 文案目录的行为；check 已有的目录检查（goal-53）；docs/plugin-guide.md 的更新（由审查者在合并前同步）。
```

## 第 11 波

### goal-104 · T-03.13 媒体按来源传递

状态：待派发。前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.6（PR #55）、T-07.7（PR #116，合并提交 `e65eb4d`）提供入站媒体与统一入口。2026-11-09 按 D-87、D-88 新增。当前 `media.Store.LeaseForSend` 返回只读 token，`message.New` 通过函数注入取得它；本切片把该返回值扩为准备好的媒体来源。`settings.Settings.BaseURL` 与 `webhost.Options.BaseURL` 已存在；`cmd/alonabot/main.go` 当前只装配认证、插件代理与调度，完整媒体装配由 T-13 接入下文的 Store 与 Host 接口。回环 Adapter 已在 `internal/core/adapter/remote/remote_test.go` 的 `TestMain` 中，无独立的 `examples/adapters/loopback/` 工程。

```text
/goal 使用 $alonabot-execute 推进 goal-104（切片 T-03.13：媒体按来源传递）。goal 目录 goal-104，分支 goal-104，worktree .worktrees/goal-104，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：出站媒体按来源传递，不统一转为 base64。MediaSource 追加插件可用的 `inline_base64` 与只由 Core 填写的 `platform_file`；Send 的目标实例就是入站对象的来源实例时，Core 把平台文件标识交给 Adapter 复用，不取件；Adapter 可以经 `core.media.url` 为本次发送的媒体取得临时 URL，由平台端从 Core 统一端口拉取；SDK 的字节构造函数在 1 MiB 内用内联数据，超出时先上传。

依据：spec/plan/core.md T-03.13、T-03 步骤 10 与失败停止条件；CONTRACTS §1.2～§1.4、§7.2、§15.1（base_url）；WIRE-FORMAT §3.4、§3.5 末段；MESSAGE-MODEL §2；DECISIONS D-86～D-88、D-109、D-121；TEST-MATRIX V-89 的 Core 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto（MediaSource）、adapter.proto（媒体 URL 载荷）；对应 contracts/pb/alonabotv1/ 生成文件；contracts/validate/chain.go；contracts/wirespec_test.go
- internal/wire/services.go：追加 ServiceMediaURL = "core.media.url"
- internal/core/media/：lease.go、store.go、incoming.go；新增 url.go（RPC 签发与 HTTP 读取），沿用现有对象与租约回收
- internal/core/message/send.go：媒体准备注入函数与调用处；permissions_test.go 中的现有 New 调用随签名核对，断言保持原义
- internal/core/webhost/host.go：Options 追加 Media http.Handler，Host 校验主机名后把 /media/ 交给它，Media 为空时该保留路径返回 404
- sdk/message/media.go、sdk/internal/convert/message.go；sdk/adapter/adapter.go 的 Core 接口；sdk/adapter_run.go 的 adapterCore；sdk/media.go（OpenMedia 的来源检查）；新增 sdk/media_bytes.go
- adapters/telegram/telegram_test.go：仅给既有 recordingCore 补上新增 MediaURL 方法，返回明确的 UNSUPPORTED；sendTestCore 已嵌入它，沿用该实现，现有测试场景与断言保持
- internal/core/adapter/remote/remote_test.go：复用并扩展现有回环进程与装配；新增 media_source_test.go，仅提交下文一个集成测试；上述包的 *_local_test.go
允许的新跨模块边：无。webhost 经 http.Handler 注入媒体处理器；Core 生产代码沿用现有依赖，回环测试复用现有 SDK 入口。
协议授权：本切片明确授权上述 proto 追加。与 T-03.14～T-03.19 可并行；语句不预分配字段号。后合并者同步最新 main，在已合并字段之后顺延编号，重新 go generate ./... 并运行 contracts/wirespec_test.go；保留先合并者的字段与转换分支。

接口（下列标为新增的符号由本切片引入；内部辅助自定）：
1. proto：MediaSource.source 追加 string inline_base64、string platform_file；adapter.proto 新增 MediaURLRequest{string token; string base_url;}、MediaURLResponse{string url; int64 expires_at_unix_ms;}，固定服务载荷版本为 1。
2. internal/core/media：把现有 LeaseForSend 改为
   func (s *Store) LeaseForSend(ctx context.Context, c router.Caller, readerSessionKey string, target *pb.Target, source *pb.MediaSource, deadline time.Time) (prepared *pb.MediaSource, release func(), err error)
   新增 func (s *Store) AddURLService(m *router.Mux, baseURL *url.URL)
   新增 func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request)
   AddURLService 注册 ServiceMediaURL 并保存已校验的缺省前缀；ServeHTTP 只服务临时 URL。
3. internal/core/message：现有 New 同步为
   func New(adapters *adapter.Registry, p *policy.Policy, lease func(context.Context, router.Caller, string, *pb.Target, *pb.MediaSource, time.Time) (*pb.MediaSource, func(), error), issue func(context.Context, router.Caller, *pb.Target, *pb.MessageChain) error, requestTimeout time.Duration) *Service
   Send 用 prepared 替换来源，发送结束统一 release。
4. SDK：message.MediaSource 追加 InlineBase64 string、PlatformFile string；后者只供 Adapter 读取 Core 准备的来源，插件发送入口拒绝它。转换仍在 sdk/internal/convert。
   sdk/adapter.Core 新增 MediaURL(ctx context.Context, token, baseURL string) (url string, expiresAt time.Time, err error)，由现有 adapterCore 实现；baseURL 为空时用 Core 配置。
   sdk 根包新增：
   func ImageBytes(ctx context.Context, data []byte) (message.Segment, error)
   func VideoBytes(ctx context.Context, data []byte) (message.Segment, error)
   func FileBytes(ctx context.Context, data []byte, filename, mime string) (message.Segment, error)
   构造结果分别交给 message.Image、Video、File；上传复用现有 func Upload(ctx context.Context, r io.Reader, mime string) (*message.MediaRef, error)，图片与视频 MIME 留空由 Upload 判断。字节入口放在 sdk 根包以复用上传，sdk/message 保持纯数据构造。

行为：
- contracts/validate.Chain 在插件出站边界校验标准 base64 与解码后 1 B～1 MiB，非法为 INVALID_ARGUMENT；platform_file 同样拒绝。HTTPS 保留现有校验；SDK 转换统计四种插件来源恰有一个，Adapter 解码另接纳 Core 的 platform_file。
- incoming_ref 先按现有 incomingAccess 核对 Session、有效期与完整目标范围。同实例时从 incomingObject.source 取 fileID，返回 platform_file，保留有截止时间的对象租约，跳过 loadIncoming；readability 为不可读仍可复用有效的平台标识。跨实例沿用现有惰性取件、配额与只读 token。已上传 media_ref 沿用只读租约；https_url、inline_base64 原样交给 Adapter，不下载、不转码。
- core.media.url 只接受绑定调用者 Adapter Session 的本次发送只读 token（entry.lease），普通上传/入站 token 与其他 Session 的 token 均拒绝。url_token 单独生成 32 字节随机值并用无 padding base64url 编码（现有媒体 token() 使用 16 字节，保持其用途）；绑定一个对象与同一消费租约，有效期取租约期限，release、Session 撤销与 Sweep 一并使 URL 失效。
- /media/<url_token> 保留给 Core，不经管理认证。有效请求只接受 GET、HEAD，使用对象 MIME 并支持 Range；到期或租约结束后 404。文件读取期间沿用 Store 的回收同步，避免返回另一对象或已回收文件。日志若记录 token，只写前 8 个字符。
- URL 前缀优先取请求 base_url，否则取 settings 的 base_url；请求前缀须为绝对 http(s) URL，只用于拼接地址，Core 不向它发起下载。Host 的主机名校验保持现有规则，代理部署按配置让地址抵达同一 Listener。
- 字节构造函数拒绝空数据；1 B～1 MiB 直接标准 base64 编码，超出时以 bytes.Reader 调用 Upload 并返回 media_ref，上传失败完整返回。现有 OpenMedia(ctx, src message.MediaSource) 仍只读 Incoming 或 MediaRef，对 InlineBase64、HTTPSURL 与 PlatformFile 返回 errcode.InvalidArgument。
- 当前出站 Hook 尚未实现；本切片的来源校验与准备作为 Send 的公共路径保留，后继 T-05.3 在 Hook 改写后复用。T-03.14 若先合并，同步它的 voice 来源遍历；本切片先合并时由后者接入。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：internal/core/adapter/remote/media_source_test.go 的 TestMediaSources，一个集成测试复用 remote_test.go 的真实回环 Adapter 进程，装配真实 message.Service、媒体 Store、Mux 与 webhost：inline_base64 原样到达；同实例入站对象以 platform_file 到达且 ReadPlatformFile 调用数为零；Adapter 在发送在途时取得 URL，GET 读出原字节，发送完成后该 URL 返回 404。
- 协议编号检查：更新 contracts/wirespec_test.go。多列理由：新增协议字段和载荷须保护已有编号，该项独立于主路径集成测试。
- 本地自测（*_local_test.go，不提交，PR 写明结果）：V-89 其余 Core 分项，含大小与 base64 边界、伪造 platform_file、跨实例取件与目标范围、URL 的 HEAD/Range/期限/Session 撤销与错误 token、SDK 内联与上传及失败、OpenMedia 来源检查。sdk/ 与 internal/core/supervisor/ 的验证全部留在本地，不新增 testdata 测试插件。

完成条件：gofmt -l 无输出；go generate ./...；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./adapters/...；go build ./...。V-89 的 Core 分项在 PR 写明命令、平台与结果；集成装配把 AddURLService 与 webhost.Options.Media 接到同一媒体 Store，文档同步列出 T-13 正式装配需复用的 Store、Mux 注册与 Host 注入点，使用生效的 media.NewLimits(v) 与配置的 base_url。出现来源统一 base64、插件可给出 platform_file、取件/Grant 扩权、租约泄漏或失败后重发时停止合入。

不做：两个官方 Adapter 的平台发送实现（T-03.20、T-15.5）；宿主路径构造函数；任意 URL 下载代理；完整开发组件/正式安装装配（T-13、T-09）；新增出站 Hook。
```

### goal-105 · T-03.14 新增片段与字段

状态：待派发。前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.8（PR #136，合并提交 `3621fc1`）已提供 `buttons.Service.Issue`、`adapter.ExtensionChecker` 与带类型的按钮动作。SDK 的 `message.Button` 是 `sdk/internal/messagecodec.Button` 的别名，新增 URL 字段须改其定义；SDK 的收发转换在 `sdk/internal/convert`。2026-11-09 按 D-87、D-89、D-91、D-97 新增，同日按 D-100 追加会话类型；字段编号按并行切片规则在实现时追加。

```text
/goal 使用 $alonabot-execute 推进 goal-105（切片 T-03.14：新增片段与字段）。goal 目录 goal-105，分支 goal-105，worktree .worktrees/goal-105，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在通用消息模型中追加语音片段 voice，按 D-91、D-97 补齐既有合并转发的引用节点、入站转发标识与 URL 按钮，Core 在调用 Adapter 前完成结构校验与能力判定，SDK 提供对应的构造函数与入站类型；Event 追加会话类型 chat_type，SDK 提供会话类型的判断方法。全体提及等其他平台专有片段按 D-87 留给该 Adapter 的 extension。

依据：spec/plan/core.md T-03.14、T-03 步骤 11 与失败停止条件；CONTRACTS §8；WIRE-FORMAT §3.4、§3.5 末段；MESSAGE-MODEL §1、§2、§6.3；DECISIONS D-86、D-87、D-89、D-91、D-97、D-100、D-108、D-110；TEST-MATRIX V-90、V-98 的 Core 分项、V-108 的语音分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto、对应 contracts/pb/alonabotv1/message.pb.go；contracts/validate/chain.go；contracts/wirespec_test.go
- internal/core/message/send.go、extension.go（整链能力预检）；internal/core/buttons/buttons.go（跳过 URL 按钮）
- internal/core/adapter/remote/services.go（入站 chat_type 校验）、media.go（voice 来源登记）；internal/core/media/incoming.go（voice 的 Session token）、internal/core/message/send.go（voice 媒体租约遍历）
- sdk/message/message.go、event.go、buttons.go；sdk/internal/messagecodec/codec.go；sdk/internal/convert/message.go、event.go；新增 sdk/chat_type.go
- internal/core/adapter/remote/remote_test.go：复用现有回环 Adapter/插件；新增 segments_test.go，仅提交下文一个集成测试；上述包的 *_local_test.go
允许的新跨模块边：无。消息能力预检仍归 message，按钮 token 仍经现有 issue 函数注入，入站校验归 remote。
协议授权：本切片明确授权上述 proto 追加，与 T-03.13～T-03.19 可并行；不预分配字段号。后合并者同步 main 后顺延追加编号，保留先合并字段，再运行 go generate ./... 与 contracts/wirespec_test.go。

接口（新增，既有字段与构造函数保留）：
1. proto：新增 VoiceSegment{MediaSource source; optional uint64 duration_ms; optional string caption;} 并追加 Segment.voice；ForwardItem 追加 optional string platform_message_id，ForwardSegment 追加 optional string platform_forward_id，Button 追加 optional string url；新增枚举 ChatType（CHAT_TYPE_UNSPECIFIED 为零，另有 CHAT_TYPE_PRIVATE、CHAT_TYPE_GROUP、CHAT_TYPE_CHANNEL），Event 追加 ChatType chat_type。
2. sdk/message：
   type VoiceSegment struct { Source MediaSource; Duration time.Duration; Caption string } // 实现 Segment 与 Content
   func Voice(src MediaSource) Segment
   func VoiceURL(httpsURL string) Segment
   func Forward(items ...ForwardItem) Segment
   func URLButton(label, url string) Button
   ForwardItem 追加 PlatformMessageID string；ForwardSegment 追加 PlatformForwardID string；messagecodec.Button 追加 URL string（message.Button 的既有别名随之可用）。
   type ChatType int；常量 ChatUnspecified（零值）、ChatPrivate、ChatGroup、ChatChannel；Event 追加 ChatType ChatType。
3. sdk.Context：
   func (c *Context) ChatType() message.ChatType
   func (c *Context) IsPrivate() bool
   func (c *Context) IsGroup() bool
   func (c *Context) IsChannel() bool
   值取本次可信 Event；没有 Event 或类型为 UNSPECIFIED 时返回零值，三个判断均为 false。
4. 沿用 func validate.Chain(c *pb.MessageChain) error、func (s *buttons.Service) Issue(ctx context.Context, caller router.Caller, target *pb.Target, chain *pb.MessageChain) error，以及现有 message.New 的函数注入；此切片不改变它们的签名。SDK 的 ChainToProto/ChainFromProto、EventToProto/EventFromProto 在 convert 内扩展对应分支。

行为：
- voice 的来源复用当前 MediaSource 校验，来源恰有一个；Duration 以毫秒往返，零值为 absent、亚毫秒舍去，负值在 SDK 发出请求前返回 errcode.InvalidArgument。来源登记、为插件签入站 token 与发送消费租约都须包含 voice，覆盖 Event.Message 与新旧 payload 中的 Chain。
- ForwardItem 的 chain 与 platform_message_id 恰有一个；引用节点指同一发送会话的既有平台消息。节点 chain 的合法类型按 D-91 为 text、mention、image、video、voice 与本 Adapter 的 extension；嵌套 forward 按 D-91 返回 UNSUPPORTED。出站 platform_forward_id 为 INVALID_ARGUMENT；入站允许只带该标识的 forward，转换到 SDK 后保留原值。SDK 与 Adapter 的转换要区分出站预检与入站读取，入站空 Chain 的既有行为保持。
- Send 在首个 Adapter 调用前完成整链预检，递归核对 voice、forward 所需的 message.voice、message.forward 能力（supported=true），未声明为 UNSUPPORTED。结构错误为 INVALID_ARGUMENT；传播 validate.Error 的实际错误码，避免把 D-91 的 UNSUPPORTED 改写为 INVALID_ARGUMENT。能力拒绝发生在按钮签发和媒体取件之前。
- Button 的 url 与 action_id 恰有一个，label 必填；url 为 https 或 tg 的绝对链接，URL 按钮带 value 为 INVALID_ARGUMENT。Issue 仍先清除所有外来 core_token，只为 action 按钮查路由、校验值与签 token；纯 URL 键盘直接通过，混排时 URL 按钮保留顺序且不要求 button.clicked 路由。
- remote 在持久化游标、登记媒体与派发之前，拒绝没有 target 却带非 UNSPECIFIED chat_type 的事件（INVALID_ARGUMENT）；未知枚举值同样在入站边界拒绝。有效值随可信 Event 保留，chat_type 不进入 Target、不改变目标补齐、原 Target 回复豁免与发信准入。Core 自造的 bot.online、bot.offline、scheduled 事件保持零值。
- 与 T-03.13 可并行。它先合并时，voice 遍历采用其准备好的媒体来源接口；本切片先合并时由它保留 voice 分支。SDK 公共类型沿用 message、errcode，pb 转换留在 convert；本切片不增加 VoiceBytes，通用字节上传可先取得 MediaRef.Source() 再用 Voice。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：internal/core/adapter/remote/segments_test.go 的 TestMessageSegments，一个集成测试复用真实回环 Adapter/插件进程，装配 message.Service 与 buttons.Service：声明能力的实例收到 voice 与 URL 按钮（来源、顺序、URL 保持且 URL 按钮没有 token），无对应能力的实例在调用前被拒；chat_type=GROUP 的入站事件经真实 Session 到达插件并保持 GROUP。
- 协议编号检查：更新 contracts/wirespec_test.go。多列理由：新增字段、oneof 分支与枚举需要协议编号保护。
- 本地自测（*_local_test.go，不提交，PR 写明结果）：V-90 的结构错误、forward 两种节点/入站标识/嵌套限制、URL 与 action 混排、URL 按钮带 value、实际 token 表；V-98 的缺 target/未知值/三个判断与目标准入；V-108 的语音 Duration；voice 入站 token、媒体租约与转换往返。sdk/ 与 internal/core/supervisor/ 只本地验证，不新增 testdata 测试插件。

完成条件：gofmt -l 无输出；go generate ./...；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。PR 写明 V-90、V-98 的 Core 分项与 V-108 的语音分项证据；按 T-03 的失败停止条件，已知不支持却先发出平台请求、媒体授权扩大或新增 D-87 范围外通用能力时停止合入。

不做：两个官方 Adapter 的平台映射、全体提及 extension、平台转码、转发内容展开、聊天目录与 chat_title（T-03.25）；新增出站 Hook。
```

### goal-106 · T-03.15 平台动作

状态：待派发。前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.8（PR #136，合并提交 `3621fc1`）提供 extension 校验与 `manifest.CheckDependencies(m *Manifest, in Installed) []Problem`；当前安装快照的 `Installed.Adapters` 是 `map[string][]ExtensionRequirement`。SDK 已有 `adapter.Event[T]`、`adapter.Extension[T]` 的类型推导与内部 convert.Descriptor；平台动作声明须同时扩展两个 Descriptor。`sdk.Action[T]` 是按钮动作，平台动作使用 `sdk/adapter.Action[P,R]`。2026-11-09 按 D-87、D-92 新增，同日按 D-113 追加类型推导。

```text
/goal 使用 $alonabot-execute 推进 goal-106（切片 T-03.15：平台动作）。goal 目录 goal-106，分支 goal-106，worktree .worktrees/goal-106，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Adapter 在注册描述中声明平台动作，插件经 `core.action` 调用；Core 按发信准入检查目标、核对声明与参数 Schema 后经 `adapter.action` 交给实例，并按结果 Schema 校验返回值。动作以参数与结果的 Go 类型声明，Schema 由 SDK 推导（D-113）。

依据：spec/plan/core.md T-03.15、T-03 步骤 12 与失败停止条件；CONTRACTS §8（平台动作、目标准入、adapter_dependencies）；WIRE-FORMAT §3.4、§3.5；DECISIONS D-86、D-87、D-92、D-102、D-113；TEST-MATRIX V-91 的 Core 分项、V-111 的动作分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/adapter.proto、对应 contracts/pb/alonabotv1/adapter.pb.go；contracts/validate/adapter.go；contracts/wirespec_test.go
- contracts/manifest/deps.go；cmd/alonabot/dependencies.go 中安装快照构造随新增字段同步，保持当前单插件开发入口的依赖行为
- internal/wire/services.go：追加 ServiceAction = "core.action"、ServiceAdapterAction = "adapter.action"
- internal/core/adapter/ 新增 action.go（可选接口与 account 目标补齐）；remote/ 新增 action.go，扩展 declarations.go 的动作 Schema 缓存
- internal/core/policy/policy.go：抽出共享实例许可检查；internal/core/message/ 新增 action.go，在现有 Service 上提供 Action
- sdk/adapter/ 新增 action.go，declaration.go 的 Descriptor 追加动作；sdk/internal/convert/declaration.go 同步类型推导与描述转换，动作载荷转换辅助放在此内部包
- sdk/ 新增 action.go；sdk/adapter_run.go 追加 adapter.action 分派
- internal/core/adapter/remote/remote_test.go：复用并扩展现有回环进程；新增 action_test.go，仅提交下文一个集成测试；上述包的 *_local_test.go
允许的新跨模块边：无。动作服务沿用 message 对 adapter、policy、router 的依赖；Adapter 类型推导复用已批准的 jsonschema-go，后台调用复用 SDK 当前 Session。
协议授权：明确授权上述 proto 追加，与 T-03.13～T-03.19 可并行；不预分配字段号。后合并者同步 main，顺延追加字段与枚举，再运行 go generate ./... 与 contracts/wirespec_test.go。

接口（新增；SDK 公共签名只使用自有类型）：
1. proto：新增 ActionScope（ACTION_SCOPE_UNSPECIFIED 为零，另有 ACTION_SCOPE_CHAT、ACTION_SCOPE_ACCOUNT）、ActionDeclaration{string name; uint32 version; ActionScope scope; string params_schema_json; string result_schema_json;}，AdapterDescriptor 追加 repeated ActionDeclaration actions。
   新增 ActionRequest{Target to; string adapter_instance_id; string name; uint32 version; string params_json;}、ActionResponse{string result_json;}；AdapterActionRequest{string instance_id; Target target; string name; uint32 version; string params_json;}，响应复用 ActionResponse。固定服务载荷版本为 1。
2. internal/core/adapter：
   type ActionExecutor interface {
       CheckAction(name string, version uint32, paramsJSON string) (pb.ActionScope, *pb.ErrorInfo)
       Action(ctx context.Context, target *pb.Target, name string, version uint32, paramsJSON string) (resultJSON string, safe *pb.ErrorInfo)
   }
   func (r *Registry) CompleteAccount(adapterID, instanceID string) (*pb.Target, *pb.ErrorInfo)
   CompleteAccount 只补 Adapter、实例、账号，chat/thread 保持空；显式实例优先，否则取该 Adapter 已配置的缺省实例，没有时 FAILED_PRECONDITION。remote.instance 实现 ActionExecutor，保持创建它的 Session 绑定。
3. internal/core/policy：
   func (p *Policy) CheckInstance(ctx context.Context, pluginID string, inv *router.Invocation, instanceID string) *pb.ErrorInfo
   同插件可信在途事件的原实例免实例许可，其余逐次调用现有 state.Store.AdapterInstanceAllowed；读库失败 INTERNAL、无许可 PERMISSION_DENIED。CheckSend 的完整 Target 豁免与会话禁止顺序保持原状，只复用其非豁免分支的读库逻辑。
   internal/core/message：func (s *Service) Action(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response
   装配处通过现有 (*router.Mux).Handle(wire.ServiceAction, service.Action) 注册；当前生产入口尚未装配 Send/Adapter，主路径在真实进程集成装配中接通，正式装配由 T-13/T-09 复用。
4. sdk/adapter：
   type ActionScope int；常量 ActionScopeUnspecified（零值）、ActionScopeChat、ActionScopeAccount
   type ActionDeclaration // 与现有 EventDeclaration 一样，载荷 Schema 与推导错误封装在 convert 内
   func Action[P, R any](name string, version uint32, scope ActionScope) ActionDeclaration
   Descriptor 追加 Actions []ActionDeclaration；convert.Descriptor 同步。
   type ActionExecutor interface {
       Action(ctx context.Context, to message.Session, name string, version uint32, params json.RawMessage) (json.RawMessage, *errcode.Error)
   }
   account 动作的 to 为空；sdk/adapter.Instance 保持原接口，以可选接口提供动作。
5. sdk 根包：
   type ActionCaller interface {
       CallAction(to message.Session, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error)
   }
   func (c *Context) CallAction(to message.Session, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error)
   type ActionClient struct { /* 保存后台任务的 context */ }
   func NewActionClient(ctx context.Context) *ActionClient
   func (c *ActionClient) CallAction(to message.Session, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error)
   两者实现 ActionCaller，opts 只接受既有 Via；chat 动作传明确会话（Handler 可用 c.Session()），Via 写进 to 的 adapter_instance_id；account 动作传空会话，Via 写进请求的 adapter_instance_id。Context 保留父调用与期限，过期返回 errcode.FailedPrecondition；后台客户端没有父调用，ctx 负责取消与期限。params 由 SDK 编码，结果保留 JSON 原文及同时返回的安全错误。
6. contracts/manifest：新增 ActionRequirement{Name string; Version uint32}（YAML 键 name、version）；AdapterDependency 追加 Actions []ActionRequirement（键 actions）；Installed 追加 Actions map[string][]ActionRequirement，按 adapter_id 索引。现有 Adapters map[string][]ExtensionRequirement 与 CheckDependencies(m *Manifest, in Installed) []Problem 的签名保留，调用者按新增字段提供动作快照。

行为：
- 注册边界复用 validate.IsAdapterEventType 检查本 Adapter 前缀与名称，版本大于零、scope 为 CHAT/ACCOUNT；同名同版本重复拒绝。参数与结果 Schema 都必须可解析、Resolve 成功，失败使注册失败。remote 在 Bind 时缓存 Resolved Schema，实例不采用新 Session 的声明。
- SDK 用 jsonschema-go 从 P、R 推导；P 必须为结构体，参数拒绝未声明属性、无 omitempty 的字段为必需。结果 Schema 允许额外属性（含嵌套结果对象），符合平台原文的用法。推导失败在 RunAdapter 发出 REGISTER 前报告声明名与错误；不提供手写 Schema 入口。
- core.action 校验请求后选择目标：chat 路径通过 Registry.Complete(to, inv.DefaultTarget) 补齐，再依次 CheckSend、实例启用与 Adapter/account 绑定检查；account 路径按动作完整名取得 adapter_id，用 CompleteAccount 取显式实例或该 Adapter 的缺省实例，再按 CheckInstance 与实例绑定检查准入。account 请求没有 chat_id，沿用 D-102 的缺省实例选择；显式选择可信事件的原实例时，仍享有原实例的许可豁免。两个目标字段同时给出为 INVALID_ARGUMENT；始终核对动作所属 Adapter。
- 准入后以 ActionExecutor.CheckAction 核对声明、版本与参数 Schema；未声明/版本不符或实例未实现接口为 UNSUPPORTED，作用范围与目标形状不符、参数不符为 INVALID_ARGUMENT。所有拒绝都发生在 adapter.action 之前。
- remote.Action 通过创建实例时的 Session 只调用一次 adapter.action；account 动作的 AdapterActionRequest 只带 instance_id，target 为空，内部补齐的账号只用于 Core 绑定检查。SDK 运行时在锁外调用可选接口，未实现时 UNSUPPORTED。Core 对存在的 JSON 结果做结果 Schema 校验，成功响应缺少结果或结果不符为 INTERNAL，并记录不含参数、凭证的诊断。有效 JSON 结果与安全错误可同时返回，保留二者；断线与超期按 UNAVAILABLE、DEADLINE_EXCEEDED 报告，保持结果未知语义，不自动重试。
- adapter_dependencies.actions 按名称与精确版本检查，与 extension 一样报告 Problem；必需依赖缺失拒绝启用，可选依赖只记诊断。保留当前 extension 与 plugin 依赖检查及所有调用者的原行为。
- 动作不按 Actor 再造权限模型，插件命令上的 Permission 与实例许可按既有契约生效。当前 T-05.3 尚未实现，由它把 core.action 加入 send_forbidden；同步 main 时若其已合并，允许仅修改 internal/core/router/ 的现有出站 Hook 禁止名单追加本服务，并本地核对拒绝行为。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：internal/core/adapter/remote/action_test.go 的 TestPlatformAction，一个集成测试复用真实回环 Adapter 进程与真实 Core Mux、message.Service、policy 和 SQLite：已声明动作经 core.action→adapter.action 成功并保留 JSON；不符参数 Schema 时 Adapter 调用数不变；收紧实例许可后后台调用返回 PERMISSION_DENIED。
- 协议编号检查：更新 contracts/wirespec_test.go。多列理由：新增声明、服务载荷与 scope 枚举需要编号保护。
- 本地自测（*_local_test.go，不提交，PR 写明结果）：V-91 的声明错误、版本/范围/绑定、原 Target 与原实例豁免、许可撤销、结果 Schema、写出后断线/取消且不重试、依赖缺失；V-111 的 P/R 类型推导、omitempty/嵌套/切片/额外结果字段与推导失败；ActionCaller 的 Handler/后台两条路径。sdk/ 与 internal/core/supervisor/ 的验证只留在本地，不新增 testdata 测试插件。

完成条件：gofmt -l 无输出；go generate ./...；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/... ./cmd/...；go build ./...。PR 写明 V-91 Core 分项与 V-111 动作分项的命令、平台、结果，文档同步列出接口、动作依赖快照与服务注册点。出现动作绕过发信准入、未声明动作透传、丢失有效结果或自动重试时停止合入。

不做：官方 Adapter 的具体平台动作及常量包（T-03.21、T-15.6）；原始客户端透传；新增管理入口与完整生产装配；扩展 sdk.Action[T] 的按钮语义；新增出站 Hook。
```

### goal-107 · T-03.16 撤回

状态：待派发。前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.9（PR #118，合并提交 `1cbfb2e`）提供 Core 与 SDK 两侧可选的 `adapter.Editor`；SDK 的 `c.Edit`、`sdk.Edit` 已收 `message.MessageRef`，句柄拆分复用 `MessageRef.Parts` 与内部 `convert.RefToProto`。Core 新增撤回仍须先调用已合入的 `Registry.Complete` 补齐省略的实例与账号，再复用准入。2026-11-09 按 D-87、D-93 新增，同日按 D-104 改为只收句柄。

```text
/goal 使用 $alonabot-execute 推进 goal-107（切片 T-03.16：撤回）。goal 目录 goal-107，分支 goal-107，worktree .worktrees/goal-107，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现 Core 固定服务 `core.delete` 与 `adapter.delete`，按 `message.delete` 能力判定，准入与 Edit 相同，按输入顺序汇总每个 ID 的结果；SDK 提供撤回方法。输入状态与表情回应只有 Telegram 提供，是它的平台动作（goal-112），不在此实现。

依据：spec/plan/core.md T-03.16、T-03 步骤 13 与失败停止条件；CONTRACTS §1.2、§8；WIRE-FORMAT §3.4、§3.5；MESSAGE-MODEL §4、§4A；DECISIONS D-86、D-87、D-93、D-102、D-104；TEST-MATRIX V-92 的 Core 分项、V-102 的撤回分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto（撤回请求与结果）、adapter.proto（Adapter 请求）；对应 contracts/pb/alonabotv1/ 生成文件；contracts/wirespec_test.go
- internal/wire/services.go：追加 ServiceDelete = "core.delete"、ServiceAdapterDelete = "adapter.delete"
- internal/core/adapter/ 新增 delete.go；remote/ 新增 delete.go；internal/core/message/ 新增 delete.go，复用现有 Service 的 adapters 与 policy
- sdk/message/ 新增 delete.go；sdk/adapter/ 新增 delete.go；sdk/ 新增 delete.go；sdk/internal/convert/ 新增 delete.go；sdk/adapter_run.go 只追加 adapter.delete 分派
- internal/core/adapter/remote/remote_test.go：复用并扩展现有回环进程；新增 delete_test.go，仅提交下文一个集成测试；上述包的 *_local_test.go
允许的新跨模块边：无。Core 复用已有 adapter、policy、router 依赖，SDK 通过当前 Session 与内部 convert 转换，不改两个 Instance 基础接口。
协议授权：明确授权上述 proto 追加，与 T-03.13～T-03.19 可并行；不预分配字段号。后合并者同步 main 后顺延追加编号，重新 go generate ./... 与 contracts/wirespec_test.go。

接口（新增；SDK 侧只用自有类型）：
1. proto：message.proto 新增 DeleteRequest{Target to; repeated string platform_message_ids;}、DeleteItemResult{string platform_message_id; PartStatus status; ErrorInfo error;}、DeleteResult{Target target; repeated DeleteItemResult items;}；adapter.proto 新增 AdapterDeleteRequest{string instance_id; Target target; repeated string platform_message_ids;}，响应复用 DeleteResult。状态复用现有 PartStatus，固定服务载荷版本为 1。
2. internal/core/adapter：
   type Deleter interface {
       Delete(ctx context.Context, target *pb.Target, platformMessageIDs []string) (*pb.DeleteResult, *pb.ErrorInfo)
   }
   remote.instance 实现此可选接口，经其创建时绑定的 Session 调用 adapter.delete。
   internal/core/message：func (s *Service) Delete(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response
   装配处用现有 (*router.Mux).Handle(wire.ServiceDelete, service.Delete) 注册。当前生产入口尚未装配 Send/Adapter，本切片在真实进程集成装配中接通，正式装配由 T-13/T-09 复用。
3. sdk/message：
   type DeleteItemResult struct { PlatformMessageID string; Status PartStatus; Error *errcode.Error }
   type DeleteResult struct { Session Session; AdapterInstanceID string; Items []DeleteItemResult }
   Items 按输入顺序与消息句柄一一对应，不新增整体成功状态。
4. sdk 根包：
   func (c *Context) Delete(refs ...message.MessageRef) (*message.DeleteResult, error)
   func Delete(ctx context.Context, refs ...message.MessageRef) (*message.DeleteResult, error)
   c.Delete 保留当前父调用与期限，Context 过期返回 errcode.FailedPrecondition；包级方法为后台调用。二者复用 MessageRef.Parts 与 convert.RefToProto 拆分，不接受独立 Target、裸 ID 或 Via 参数。
5. sdk/adapter：
   type Deleter interface {
       Delete(ctx context.Context, to message.Session, platformMessageIDs []string) (*message.DeleteResult, *errcode.Error)
   }
   sdk/adapter_run.go 通过可选接口分派，未实现时返回 UNSUPPORTED。convert 内转换请求目标、结果及错误，SDK 公共类型不引用 pb。

行为：
- SDK 先校验 1～100 个合法句柄，全部会话（含 thread）与实例部分相同；省略实例统一保留为空，显式/省略混用按实例部分不一致拒绝。空输入、不合法句柄、混合会话或实例在发送 RPC 前返回 errcode.InvalidArgument；保留输入顺序与重复项。
- Core 校验载荷、非空 ID 与数量；显式 To 通过 Registry.Complete(to, inv.DefaultTarget) 补齐实例与账号，再沿用 policy.CheckSend 的会话禁止、完整原 Target 豁免、当前实例许可，以及 Send 的实例启用与 Adapter/account 绑定检查。省略实例的句柄按 D-102 选择；不以消息 ID 推断实例。
- 目标实例须声明 message.delete 且 supported=true，并实现可选 Deleter，否则在 adapter.delete 前返回 UNSUPPORTED。通过准入后只调用一次 Adapter，不执行 BeforeSend/AfterSend。
- Adapter 按 MESSAGE-MODEL §4A 保留每个 ID 的 ACCEPTED、REJECTED、UNKNOWN、NOT_ATTEMPTED 与安全错误；明确失败或结果未知后停止后续平台操作，其余为 NOT_ATTEMPTED。已不存在或超过平台期限为 REJECTED。合并平台请求的 ID 可共享同一结果，顺序与逐项身份仍须保持。
- Core 校验 Adapter 结果：成功响应须有结果，项数、各位置的 ID 与输入一致，状态须为上述四种，否则 INTERNAL；有效结果与顶层安全错误同时存在时完整返回。写出后超期或断线按 DEADLINE_EXCEEDED/UNAVAILABLE 保留未知语义，已有逐项结果不丢失，不推断失败等于未执行，不自动重试。
- 当前 T-05.3 尚未实现，由它把 core.delete 加入 send_forbidden；同步 main 时若其已合并，允许仅修改 internal/core/router/ 的现有出站 Hook 禁止名单追加本服务，并本地核对拒绝行为。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：internal/core/adapter/remote/delete_test.go 的 TestDeleteMessages，一个集成测试复用真实回环 Adapter 进程，装配真实 Core Mux、message.Service、policy 与 SQLite：按输入顺序返回 ACCEPTED、REJECTED、NOT_ATTEMPTED 的部分结果及安全错误；未声明 message.delete 的实例收到的 adapter.delete 调用数为零。
- 协议编号检查：更新 contracts/wirespec_test.go。多列理由：新增请求与结果载荷需要保护协议编号。
- 本地自测（*_local_test.go，不提交，PR 写明结果）：V-92 的 0/1/100/101 个 ID、UNKNOWN 后停止、数量/ID/状态不符、取消与断链且不重试；V-102 的句柄转义、不同会话/话题/实例、显式与省略实例、同会话两个机器人及 SDK 前置拒绝；原目标豁免、许可撤销、账号绑定、有效结果与错误并存。sdk/ 与 internal/core/supervisor/ 只本地验证，不新增 testdata 测试插件。

完成条件：gofmt -l 无输出；go generate ./...；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。PR 写明 V-92 Core 分项与 V-102 撤回分项的命令、平台及结果，文档同步列出新接口与服务注册点。按 T-03 的失败停止条件，实例串用、准入扩权、丢失已确认结果或失败后自动重试时停止合入。

不做：Telegram/OneBot 11 的实际撤回（T-03.21、T-15.6）；通用输入状态与表情回应；新增消息发送或编辑行为；新增出站 Hook；完整生产装配。
```

### goal-108 · T-03.17 命令菜单同步

状态：待派发。2026-11-12 依据维护者决定 D-138（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`），把说明字段、SDK builder、Core 翻译、生成器结构检查与 Collect 覆盖检查纳入本切片；前置均已合并：T-04.6（PR #78）、T-04.7（PR #89）、T-04.8（PR #141，合并提交 `ac4aded`）、T-08.8（PR #102）。按 main `c04b2ca` 的代码核对。D-128 的命令名覆盖仍由后继 goal-129 接入；与 goal-125 同改 SDK 注册与生成器，后合并者同步 main 保留 Description 及覆盖检查，迁到插件对象上的同一 builder。

```text
/goal 使用 $alonabot-execute 推进 goal-108（切片 T-03.17：命令菜单同步）。goal 目录 goal-108，分支 goal-108，worktree .worktrees/goal-108，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：命令通过链式 Description 登记英文说明，其他语言按完整规范路径覆盖，生成器只检查 commands 结构，Collect 检查覆盖路径与已注册命令及其英文说明的对应；实例声明 commands.sync 时，Core 在实例启动及路由/命令策略变化后，把可用命令名与翻译说明经 adapter.commands.sync 交给它。说明的最小交付在本切片完成，goal-71 后续复用。

依据：spec/plan/core.md 的 T-03.17、T-03 步骤 14 与失败停止条件；CONTRACTS §2.2、§2.3、§5.4、§8（命令菜单）；MESSAGE-MODEL §6.5；WIRE-FORMAT §3.4、§3.5；DECISIONS D-79、D-86、D-87、D-96、D-128、D-138；TEST-MATRIX V-93 的 Core 分项；docs/goals/README.md「所有 goal 的共同约束」。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/route.proto、adapter.proto：CommandSpec 追加可选英文 description，追加菜单项、adapter.commands.sync 请求与空响应；contracts/pb/alonabotv1/ 只经 go generate ./... 生成；contracts/wirespec_test.go 按既有方式同步编号检查。T-03.13～T-03.19 已获授权只追加 proto，字段号不预分配，后合并者同步 main、顺延编号、重新生成并运行编号检查。
- sdk/registry.go、sdk/session.go、sdk/i18n.go：Description setter、REGISTER 字段和 Collect 的实际注册表覆盖检查；sdk/generate/：生成器不扫描注册，check/build 对顶层 commands 只做结构检查（键为字符串，值为含字符串 description 的对象）。若 goal-125 先合并，同步其插件对象与资源绑定入口。
- contracts/i18n/：commands 对象解析、覆盖检查、locale 枚举与命令说明选择；保留 messages 的键、占位符与 T/Format/Text 语义。
- internal/wire/services.go：末尾追加 ServiceAdapterCommandsSync = "adapter.commands.sync"；internal/core/adapter/ 与 remote/：可选同步接口、绑定 Session 的调用及启停通知；保留 Instance、Sink 的现有接口。
- internal/core/router/、internal/core/policy/：菜单计算、策略继承共用和变化通知，Core 持有按运行实例隔离的同步状态；internal/core/supervisor/supervisor.go：激活与撤销时的菜单变化通知。
- sdk/adapter/、sdk/adapter_run.go、sdk/internal/convert/：SDK 自有菜单项、可选接口、固定服务分派与协议转换。
- internal/core/adapter/remote/remote_test.go：复用测试二进制的 Adapter/插件 helper，增加一个主路径集成测试；以上包的其余验证写在 *_local_test.go，sdk/ 和 supervisor/ 的测试不提交，不新增 testdata 插件。
允许的新跨模块边：无；沿用 SDK/Core -> contracts、internal/wire，router -> adapter、policy -> router 及 remote/supervisor 的现有依赖；策略与生命周期通知通过下列函数注入，router 不反向导入 policy。

接口（照此实现；内部辅助函数自定）：
sdk
   func (b *CommandBuilder) Description(text string) *CommandBuilder
   // 当前入口保持：
   func (r *CommandRegistry) Command(path ...string) *CommandBuilder
   func (r *CommandRegistry) Group(name string) *CommandRegistry
   func (b *CommandBuilder) Handle(handler any)
contracts/i18n
   func (c *Catalog) Locales() []string
   // 由 SDK Collect 传入实际注册的完整规范路径与英文原文，检查覆盖对应关系。
   func (c *Catalog) CheckCommands(commands map[string]string) []Problem
   func (c *Catalog) CommandDescription(locales []string, path []string, source string) string
internal/core/adapter
   type CommandSynchronizer interface {
       SyncCommands(ctx context.Context, commands []*pb.CommandMenuItem) *pb.ErrorInfo
   }
sdk/adapter
   type CommandMenuItem struct { Name string; Descriptions map[string]string }
   type CommandSynchronizer interface {
       SyncCommands(ctx context.Context, commands []CommandMenuItem) *errcode.Error
   }
internal/core/policy
   func (p *Policy) CommandState(ctx context.Context, pluginID string, route *pb.RouteDescriptor) (enabled bool, permissions []string, err error)
   func (p *Policy) SetCommandsChanged(fn func())
internal/core/router
   func (r *Router) StartCommandSync(in adapter.Instance)
   func (r *Router) StopCommandSync(in adapter.Instance)
   func (r *Router) CommandsChanged()
   // Options 追加字段，由装配方传入 Policy.CommandState；原 CommandPolicy 接口保持。
   CommandState func(context.Context, string, *pb.RouteDescriptor) (bool, []string, error)
internal/core/adapter/remote（Options 追加字段，由装配方传入 Router 的同名方法）
   StartCommandSync func(adapter.Instance)
   StopCommandSync func(adapter.Instance)
internal/core/supervisor（Spec 追加字段，由装配方传入 Router.CommandsChanged）
   CommandsChanged func()
协议按 WIRE-FORMAT §3.4：CommandSpec 的 optional string description 保存英文原文；CommandMenuItem 含 string name 与 map<string, string> descriptions；AdapterCommandsSyncRequest 含 string instance_id 与 repeated CommandMenuItem commands；AdapterCommandsSyncResponse 为空，错误沿用 ErrorInfo。SDK 与 pb 的转换只在 sdk/internal/convert。

行为：
- 使用当前实际链式写法 r.Command("set").Description("Set a value").Handle(Set)，r 为 *sdk.CommandRegistry；Description 与其他 setter 一样在 Handle 前调用，Collect 按现有顺序规则报错；空串与未调用均表示未提供说明。T-04.12 改用插件对象后仍返回同一 CommandBuilder，生成器不包装 New。
- 英文原文就是 en-US；其他语言用 i18n/<locale>.json 的 {"commands":{"rss add":{"description":"添加订阅"}}} 覆盖。键为包含全部 Group 前缀的规范路径，以一个 ASCII 空格连接 token；只对应按路径注册的命令，别名和管理员改名仍用原规范路径。说明作为普通文本使用，不替换占位符；en-US.json 的 commands 覆盖报错。仅含 commands 的目录可省略 en-US.json，含其他文案时沿用既有源语言要求。
- 生成器不扫描注册（D-79 不变）；check/build 对顶层 commands 只检查结构：它是对象，键为字符串，值为含字符串 description 的对象。SDK 的 Collect 用本次实际注册的完整路径→英文原文调用 CheckCommands；覆盖路径没有对应的已注册命令，或对应命令未用 Description 登记英文原文时，注册失败，错误为英文并带出路径，随原注册错误返回。动态注册的命令按本次实际注册结果检查。
- Locales 返回目录内语言的字典序副本。CheckCommands 接收全部规范路径（无说明的值为空），对未知路径/无英文说明的覆盖返回带路径的英文 Problem，由 Collect 返回注册错误。CommandDescription 按传入的语言优先序逐键选择覆盖，各级沿 §5.4 的精确/主语言匹配；选中的语言缺少覆盖则继续下一级，最终返回 source，source 为空时返回空串。请求 en-US 时直接用 source。Core 对 en-US 及该插件目录中的每个 locale 调用此入口，顺序为 [目标 locale, 插件默认 locale]；无 Catalog 时直接用英文，未提供说明时各语言均用命令名。
- 只列该实例生效前缀含 `/`、规范路径为一个 token、未被遮蔽、生效权限集合恰为 {user} 且各级策略启用的命令，按命令名字典序返回；别名不单列。按 Router 现有优先级先确定胜出者，禁用的胜出者仍占路径。当前使用 CommandSpec.path；goal-129 后续接生效主名，覆盖键仍为声明路径。
- CommandState 与 CheckCommand 共用插件→各祖先组→路由的继承计算：父层禁用保留，权限取最近显式值，路由 default_enabled 按现有规则应用；菜单按返回的完整权限集合筛选，不以虚构 Actor 的 CheckCommand 代替。读取失败保留原错误，跳过该次同步并记完整诊断，保留已发布菜单。
- Attach/Detach 后与激活/撤销后调用 CommandsChanged；SetOverride 仅在持久化成功后通知。SetCommandsChanged 在开始处理命令策略前装配一次，既有 New(store)、SetOverride 和事件派发入口保持。CommandsChanged 对受影响的运行实例调度重新计算；回调与网络调用都在状态锁外执行。
- 远端实例启动成功并登记后 StartCommandSync，停止或 Session 撤销时 StopCommandSync，未声明 supported=true 的 commands.sync 时不调用同步。启动同步一次；同实例后续变化从首次变化起 2 秒内合并为一次，不因持续变化无限推迟，同一实例串行，不同实例独立。绑定到运行实例及 Session，旧实例的停止通知只清理自身任务；Router.Close 取消同步任务。未提供同步可选接口而声明能力时返回 UNSUPPORTED 并记诊断；空列表照常发送以清空菜单。
- 同步 RPC 失败保留完整错误诊断，不重试、不影响实例运行；下一次变化重新同步。服务分派核对 instance_id 已启动，转换并调用可选接口；沿用固定 Adapter 服务的期限与取消路径。SDK 不提供平台特例。

测试（按编码规范「测试提交规则」）：
- 提交：只在 internal/core/adapter/remote/remote_test.go 增加 TestRemoteCommandSync 一个主路径集成测试，复用真实 SDK Adapter/插件进程、Session、Router 和临时 SQLite 策略库。命令链式登记英文说明，资源提供 zh-CN 路径覆盖；启动声明 commands.sync 的实例、激活插件后观察同步，断言仅有符合规则的根命令、英文原文与中文覆盖，成功禁用命令后菜单更新。复用 helper，不新增 testdata 插件。
- 本地自测：V-93 的其余 Core 分项；生成器的 commands 结构检查（commands 非对象、值非对象、description 缺失或非字符串）、动态注册不触发生成器覆盖诊断；Collect 对含 Group 前缀的实际注册路径、未知路径、未用 Description 登记英文原文与动态注册命令的覆盖检查，注册失败时英文错误带出路径；en-US 覆盖、commands-only 无 en-US、messages 原有规则；locale 枚举与精确/主语言、默认 locale/英文回落及无说明；2 秒合并、持续变化、同实例串行、实例间独立、空菜单、前缀排除、权限与父级禁用、遮蔽、失败后下一次变化、停止与换 Session 的清理。平台名字校验、截断与 setMyCommands 去重由 goal-113 验收。

完成条件：满足上述行为与 T-03 失败停止条件；go generate ./...；go test ./contracts；go vet ./...；go test -race ./internal/wire/... ./internal/core/... ./sdk/...；go build ./...；gofmt -l 对改动 Go 文件无输出，重复生成无新增差异。PR 写明 V-93 本地结果及真实平台未运行项。

不做：Telegram 的 setMyCommands（goal-113）；命令名覆盖（goal-129）；help 与 core.command.list（goal-71）；管理 API 与页面；自动重试。
```

### goal-109 · T-03.18 流式输出辅助函数

状态：待派发。2026-11-12 依据维护者决定 D-135（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`） 补齐能力元数据、SDK 接口与测试；前置均已合并：T-03.9（PR #118，合并提交 `1cbfb2e`）、T-04.8（PR #141）。按 main `c04b2ca` 的代码核对。与 T-03.13～T-03.19 的其他切片并行时按执行顺序 §3 追加 proto；与 goal-125 同改 SDK 的 Context/装配和 remote 测试 helper，后合并者同步 main 保留对方接口。

```text
/goal 使用 $alonabot-execute 推进 goal-109（切片 T-03.18：流式输出辅助函数）。goal 目录 goal-109，分支 goal-109，worktree .worktrees/goal-109，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Adapter 在 message.send 能力声明单条文本上限与计量单位，SDK 的 c.Stream 自动使用当前事件所属实例的能力：支持 message.edit 时先发送再节流编辑，按声明上限换条；没有编辑能力时缓冲到结束后发送一次。未声明文本上限时不自动换条，插件无需了解平台差异。官方 Telegram Adapter 同步声明自己的上限。

依据：spec/plan/core.md 的 T-03.18、T-03 步骤 15 与失败停止条件；CONTRACTS §8；MESSAGE-MODEL §4、§4A、§5；WIRE-FORMAT §3.4、§5A；DECISIONS D-86、D-98、D-102、D-104、D-108、D-135；TEST-MATRIX V-94 的 Core 分项；docs/goals/README.md「所有 goal 的共同约束」。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/identity.proto：Capability 的可选 TextLimit 与 TextUnit；contracts/pb/alonabotv1/ 只经 go generate ./... 生成；contracts/wirespec_test.go 只按既有方式同步编号检查。T-03.13～T-03.19 已获授权追加 proto，字段号不预分配，后合并者同步 main、顺延并重新生成。
- sdk/message/event.go、sdk/internal/convert/message.go：SDK 自有类型及双向转换，沿用 Event、SendResult、AdapterDescriptor 与实例启动的能力载体。
- contracts/validate/adapter.go、internal/core/adapter/remote/remote.go：在 REGISTER 与实例启动返回处校验文本限额；Core 现有能力存储与事件发布保留元数据，internal/core/adapter/ 与 internal/core/router/ 仅在现有复制能力的接点需要时补齐。
- sdk/stream.go（新增）、sdk/sdk.go、sdk/session.go：流调用与 Session 内同目标节流状态，复用 Context、Send、Edit；Session 结束清理状态。
- adapters/telegram/telegram.go：platformCapabilities 的 message.send 声明复用 render.go 的 maxTextUTF16 和 UTF-16 单位，描述与实例能力保持一致。
- internal/core/adapter/remote/remote_test.go：扩展现有测试二进制的 Adapter/插件 helper，新增一个主路径集成测试；以上包的 *_local_test.go 放其余验证，不新增 testdata 插件。
允许的新跨模块边：无；沿用 SDK/Adapter/Core 现有依赖，remote_test 可沿用 Core 内部装配与现有 SDK helper。

接口（照此实现；内部辅助函数自定）：
sdk/message
   type TextUnit int
   const (
       TextUTF16CodeUnits TextUnit = iota + 1
       TextUnicodeCodePoints
       TextBytes
   )
   type TextLimit struct { Max uint64; Unit TextUnit }
   // Capability 保留 Name、Version、Supported，追加以下字段；nil 表示未声明。
   TextLimit *TextLimit
sdk
   func (c *Context) Stream(next func(context.Context) (string, error)) ([]*message.SendResult, error)
contracts/validate（新增，由 AdapterRegister 与实例启动结果接纳共用）
   func Capabilities(capabilities []*pb.Capability) error
现有入口保持：
   func (c *Context) Context() context.Context
   func (c *Context) Event() *message.Event
   func (c *Context) Send(m message.Content, opts ...SendOption) (*message.SendResult, error)
   func (c *Context) Edit(ref message.MessageRef, m message.Content) (*message.SendResult, error)
   func (r *message.SendResult) Ref() message.MessageRef
协议按 WIRE-FORMAT §3.4：Capability 追加 TextLimit text_limit，TextLimit 为 uint64 max 与 TextUnit unit，枚举含 UNSPECIFIED、UTF16_CODE_UNITS、UNICODE_CODE_POINTS、BYTES；SDK 与 pb 的转换只在 sdk/internal/convert。保留现有能力版本与名称，不新增流式 Core 服务。

行为：
- Stream 固定使用当前事件的完整 Target（包括实例、账号与话题），从该事件的 Capabilities 读取 message.edit 和 message.send 的 TextLimit。无默认事件目标返回 NoDefaultTarget，Context 已失效返回 FailedPrecondition；不调用 next、不发请求。next 为 nil 返回 InvalidArgument。
- 依次调用 next(c.Context())，每次先处理返回的文本；io.EOF 正常结束并刷新剩余内容，其他输入错误或取消立即停止，返回已有回执与原错误。空增量跳过，整个输入为空时返回空列表且不发请求；等待节流时也响应取消。
- 支持 message.edit 时首个非空增量经 c.Send(message.Text(...)) 发出，后续增量累积后以该次真实 SendResult.Ref() 编辑；同一 SDK Session、同一完整 Target 的全部 Stream 共用编辑节流，实际 Edit 发起间隔至少 1 秒，结束刷新仍遵守间隔。不同 Target 独立；并发等待可取消，网络 RPC 在状态锁外执行。
- 有 TextLimit 时按其单位在 Unicode 码点边界拆分，满条后的文本从新消息继续，换条前把旧条剩余编辑刷新完成；一个增量可跨多条。UTF-16 的非 BMP 字符计 2，Unicode 码点计 1，bytes 计 UTF-8 字节数。单个码点已超过上限返回 InvalidArgument，保留之前结果；未声明上限不自动换条。没有 message.edit 时一直缓冲到正常结束，调用 Send 一次。
- REGISTER 与实例启动结果只允许 supported=true 的 message.send 携带 TextLimit；Max=0、单位缺失/未知或其他能力携带它时返回 INVALID_ARGUMENT，实例不得发布为已启动。能力元数据经现有 Core 事件与 SendResult 转换完整保留。Telegram 的 Max 取 maxTextUTF16（4096），Unit 为 TextUTF16CodeUnits。
- 返回列表按请求顺序保留全部 Send/Edit 的完整回执；失败请求同时带回执时也追加该回执，然后停止并返回错误，不重试、不合并或丢弃回执。平台已接受但没有可编辑消息 ID 时保留原结果，后续需要编辑才返回 FailedPrecondition，不补发新消息。
- 所有 Send/Edit 沿现有路径保留父调用和取消；出站 Hook 随 T-05.3 在原路径接入。当前可观察真实 Send/Edit 请求，完整 Hook 验收由其所有者闭合。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 internal/core/adapter/remote/remote_test.go 的 TestRemoteStream，一个主路径集成测试。复用测试二进制中的真实 SDK Adapter 与插件、Session、Router 和 message.Service；Adapter 在 REGISTER/实例启动声明 message.send 的 UTF-16 上限与 message.edit，事件进入插件后驱动 c.Stream，断言先 Send、至少一次 Edit、超长后第二次 Send，返回列表保留相应回执和实际消息 ID。用含非 BMP 字符的增量断言计量，节流以 Adapter 观察到的时间验证；不新增 testdata 插件。
- 本地自测：sdk/stream_local_test.go 与相关包 *_local_test.go 覆盖三种单位、一个增量跨多条、未声明上限、无编辑能力一次发送、空输入、EOF 携带末段、输入错误/取消、同目标并发节流与不同目标独立、结束刷新、平台拒绝及部分回执、没有可用消息 ID、非法能力声明与元数据双向转换；Telegram 模拟服务核对声明值与现有文本校验一致。sdk/ 与 supervisor/ 的测试按 D-86 本地运行，结果写入 PR。

完成条件：满足上述行为与 T-03 失败停止条件；go generate ./...；go test ./contracts；go vet ./...；go test -race ./internal/wire/... ./internal/core/... ./sdk/... ./adapters/telegram/...；go build ./...；gofmt -l 对改动 Go 文件无输出，重复生成无新增差异。PR 写明 V-94 本地结果，以及 Hook、真实平台的未运行项。

不做：流式 Core 服务；Telegram 草稿辅助函数（goal-112）；普通 Send 的自动切分；Hook 实现；失败补发。
```

### goal-110 · T-03.19 Adapter 的 HTTP 路由

状态：待派发。2026-11-12 依据维护者决定 D-136（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`） 统一 Adapter 角色规则与平台凭证转交；前置均已合并：T-03.3（PR #32）、T-04.8（PR #141）、T-07.7（PR #116，合并提交 `e65eb4d`）。生命周期接入 T-07.5（PR #119，合并提交 `cc19e99`）也已合并。按 main `c04b2ca` 的代码核对；T-03.13～T-03.19 的 proto 追加按执行顺序 §3，SDK HTTP 注册器后续由 T-04.11 改为 builder，本切片沿用现有 Handle/Public。若 goal-67 先合并，沿其同进程入口复用同一 Adapter HTTP 生命周期。

```text
/goal 使用 $alonabot-execute 推进 goal-110（切片 T-03.19：Adapter 的 HTTP 路由）。goal 目录 goal-110，分支 goal-110，worktree .worktrees/goal-110，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Adapter 的 REGISTER 携带 HTTP 路由与 loopback endpoint，sdk.RunAdapter 在有路由时托管 Listener；统一端口代理 WebSocket Upgrade 并在绑定撤销时关闭连接。Adapter 包路由的 Authorization 保留给 Adapter 校验，插件路由继续清除；内部代理认证与其他身份头的清理按 CONTRACTS §7.2。

依据：spec/plan/core.md 的 T-03.19、T-03 步骤 16 与失败停止条件；CONTRACTS §7.2、§7.3、§8、§9.1；WIRE-FORMAT §3.1～§3.4；DECISIONS D-25、D-65、D-84、D-86、D-94、D-134、D-136；TEST-MATRIX V-95 的 Core 分项；docs/goals/README.md「所有 goal 的共同约束」。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/route.proto：HttpAuthMode 只追加 HTTP_AUTH_MODE_PLATFORM_TOKEN，字段与已有枚举值保持；contracts/pb/alonabotv1/ 经 go generate ./... 生成，contracts/wirespec_test.go 按既有方式同步编号检查。T-03.13～T-03.19 已获 proto 追加授权，编号按实际 main 追加，不预分配。
- contracts/validate/adapter.go、register.go：Adapter 只接纳 HTTP 路由与 loopback endpoint，复用方法、规范路径、冲突和 Listener 校验；平台 token 模式仅 Adapter 的公开路由可用，插件原有认证模式保持。
- sdk/adapter_run.go、sdk/http.go、sdk/session.go、sdk/activate.go：SDK 根包的 AdapterHTTP 可选接口、AuthPlatformToken、HELLO 的 http_listener v1、托管 Listener、REGISTER 与激活/停止接入。sdk/adapter/adapter.go 只补可选 HTTP 入口的文档；同进程 sdk/inproc.go 若已由 goal-67 合入，同步复用本路径。
- internal/core/httpproxy/proxy.go：可信角色决定 Authorization 转发、升级连接的登记与撤销回收；internal/core/supervisor/supervisor.go 的 Bind 调用传安装角色判断。internal/core/host/prepare.go 仅补必要注册/生命周期接点；internal/core/webhost/ 仅补 Upgrade 所需的 Hijack/响应策略接点。
- internal/core/supervisor/supervisor_test.go 的既有 HTTP 生命周期 fixture 沿用原断言；若 goal-67 先合并，根包 assembly.go 的 Bind 调用也同步传可信角色。
- internal/core/adapter/remote/remote_test.go：扩展现有测试二进制，新增一个统一入口主路径集成测试；以上包的 *_local_test.go 放其余验证，不新增 testdata 插件。
允许的新跨模块边：无；沿用 SDK、Core、webhost、supervisor 与 remote_test 的已有依赖。可选接口声明在 sdk 根包，避免 sdk/adapter 导入 sdk 的环。

接口（照此实现；内部辅助函数自定）：
sdk（新增）
   type AdapterHTTP interface { HTTP(r *HTTPRegistry) }
   const AuthPlatformToken HTTPAuth = HTTPAuth(pb.HttpAuthMode_HTTP_AUTH_MODE_PLATFORM_TOKEN)
   // 上行表示值与新 proto 枚举一致；公开 SDK 声明不依赖调用者导入 pb。
SDK 既有入口保持：
   func RunAdapter(a adapter.Adapter) error
   func (r *HTTPRegistry) Handle(method, path string, h http.Handler, opts ...HTTPOption)
   func Public(mode HTTPAuth) HTTPOption
   type HTTPRoute struct { Method, Path string; Handler http.Handler; Public bool; AuthMode HTTPAuth }
Core（替换 Bind 签名并同步全部调用点）：
   func (s *Server) Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor, isAdapter bool)
   func (s *Server) Drain(pluginID string, deadline time.Time)
   func (s *Server) Unbind(pluginID, generationID string)
现有校验入口保持：
   func validate.AdapterRegister(r *pb.Register, adapterID string) error
   func validate.Register(r *pb.Register) error
isAdapter 由已认证安装记录决定，supervisor 传 spec.Role == state.RoleAdapter；插件及测试调用显式传 false，Adapter 调用传 true。不得从 ID、路径、外部请求头或进程自报推断角色。

行为：
- RunAdapter 检查 a 是否实现 AdapterHTTP，调用 HTTP 收集并校验路由；存在路由时在 HELLO 前声明 http_listener v1，Bootstrap 后、REGISTER 前持续持有 127.0.0.1:0 的 Listener 并上报实际 endpoint。无路由不创建 Listener。内部凭证、ready gate、排空/停止预算复用现有 SDK HTTP 路径。
- AdapterRegister 允许 HTTP 路由与 endpoint 成对出现，拒绝其他插件路由、业务数据库与迁移；插件 Register 继续拒绝 Adapter 描述。SDK 的 HTTPRoute 转换与已有 REGISTER 载体复用，保留重复路由、非法 endpoint、路径与方法的完整错误。
- 外部路径为 /plugins/<Adapter 包 ID>/...，精确/尾斜杠前缀匹配按 D-65；私有路由仍经 adminauth.Auth.Authorize，Cookie 的非 GET/HEAD 请求先验证 CSRF 与 Origin。Adapter 的公开平台 token 路由用 Public(AuthPlatformToken)，Core 校验角色与声明，用户 Handler 校验平台 Authorization 或查询参数；错误 token 的平台响应由 Adapter 给出。
- Core 代理对 Adapter 包全部路由保留入站 Authorization，对插件路由清除 Authorization。两者均清除外部传入的内部代理凭证、X-Alonabot-CSRF、Cookie、Forwarded 与全部 X-Forwarded-*，再注入当前绑定的 HeaderProxyAuth；SDK 在用户 Handler 前恒时比较内部凭证，并在进入用户 Handler 前移除它。平台凭证不进入日志、错误或响应。
- ReverseProxy 转发 Upgrade 并双向传输字节，连接寿命独立于普通 RPC 期限。每个 binding 记录自己的升级连接；Bind 替换时关闭旧绑定连接，Unbind、Session 失效与代次变化时关闭所属连接，Drain 冻结新请求。即使 generation_id 相同，新 Session 替换也必须回收旧连接；旧绑定的异步清理只作用于自身。
- SDK 也记录并关闭自身 hijacked 连接，覆盖 http.Server.Shutdown/Close 未回收的部分；停止与断链都清理。Core 与 SDK 在锁外关闭连接，保留现有停止预算。代理继续使用受控 Host、沙箱响应头、nosniff 与 Set-Cookie 过滤。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 internal/core/adapter/remote/remote_test.go 的 TestRemoteHTTPUpgrade，一个主路径集成测试。复用测试二进制的真实 SDK Adapter，经 REGISTER/ACTIVATED、httpproxy、webhost 与统一 Listener，以 Authorization: Bearer 携带平台凭证完成公开路由 WebSocket 往返，Handler 验证收到的平台 token 与已清理的身份头；Adapter 停止后客户端观察连接关闭。握手与最小帧用标准库实现，复用现有 helper，不加 WebSocket 依赖或 testdata 插件。
- 本地自测：无路由无 Listener；角色互斥、非法端点、路径/方法/冲突和认证模式；header/query 两种平台凭证；插件 Authorization 清除与两者其余身份头清理；未激活/Draining 拒绝；私有路由认证和响应隔离；断链、同代次新 Session、换代与旧清理并发；升级连接跨普通请求期限保持；SDK hijacked 连接停止回收。sdk/ 与 supervisor/ 的新增验证按 D-86 写 *_local_test.go，结果写入 PR。

完成条件：满足上述行为与 T-03 失败停止条件；go generate ./...；go test ./contracts；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...；gofmt -l 对改动 Go 文件无输出，重复生成无新增差异。PR 写明 V-95 Core 分项与真实 OneBot 11 环境的未运行项。

不做：OneBot 11 reverse 实例、平台握手与重连（goal-114）；新增公开端口；Adapter 工程入口与资源迁移（D-134 由 goal-67 实现）；管理认证改造；自动重试。
```

### goal-125 · T-04.12 插件对象、结构体参数与子插件

状态：待派发。2026-11-12 依据维护者决定 D-137（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`） 补齐生成资源与插件对象的显式绑定，承接 D-114～D-116；前置均已合并（T-04.8 为 PR #141），按 main `c04b2ca` 的代码核对。goal-67（T-13.1）正在按旧的类型登记实现：goal-125 合并后同步库模式的资源读取；若 goal-67 先合并，本切片范围已包含其插件选项、SDK 同进程入口与示例的同步迁移，Adapter 资源规则沿用 goal-67。与 goal-101、goal-123、goal-124 共改生成器时按执行顺序 §3；与 goal-16、goal-68 并行时，后合并者同步 main 保留对方的注册与装配语义。

```text
/goal 使用 $alonabot-execute 推进 goal-125（切片 T-04.12：插件对象、结构体参数与子插件）。goal 目录 goal-125，分支 goal-125，worktree .worktrees/goal-125，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件入口改为 `func New() *sdk.Plugin`，路由、生命周期回调、OnConfigLoad 与数据库迁移登记在插件对象上；删除 sdk.Base、sdk.Root、根插件的六个注册方法（Commands、Events、Services、Subscriptions、HTTP、Schedules）与对应的注册器类型；命令 Handler 可以用一个结构体接收参数；插件可以分成一层逻辑子插件。生成资源在 New 中显式绑定对象，独立进程与库模式共用，New 可重复调用。协议与 Core 行为不变。

依据：PLAN T-04 步骤 12 及其失败停止条件；DESIGN §D、§I、§J；CONTRACTS §2.2 第 5 条、§4.1、§5.1、§5.2、§6.2、§9.1、§14.1、§15.5；DECISIONS D-86、D-114、D-115、D-116、D-137；TEST-MATRIX V-112。

范围（只新增、修改或删除这些）：
- sdk/：base.go 由插件对象实现取代；sdk.go、registry.go、events.go、buttons.go、services.go、bus.go、generated.go、i18n.go、dispatch.go、session.go、activate.go 的注册、资源读取、配置和生命周期调用点；可按职责新增 plugin.go、subplugin.go 与参数绑定文件
- sdk/config/values.go：支持注册时取得、准备后读取的子插件配置视图；sdk/config/config.go 同步包文档中的配置入口
- sdk/sqlitedb/migrate.go、sdk/gormdb/migrate.go、sdk/gormdb/versioned.go、sdk/internal/dbhook/：迁移登记与准备阶段的接入；删除旧识别接口及其 init 登记，保留迁移执行与账本语义；sdk/sqlitedb/open.go 同步包文档中的迁移入口
- sdk/generate/：入口 New 的核对、生成资源绑定、迁移调用识别；保留现有 Service Schema、配置 schema、Adapter 实例声明与 T 调用检查
- cmd/alonabot/plugin.go：将 generate.Result.RootType 的判定改为 PluginConstructor，保留现有 plugin generate/check 命令
- goal-67 若已合并：根包 alonabot.go、assembly.go 的插件选项与资源读取，sdk/inproc.go 的插件同进程入口，以及 examples/devbot/ 的插件构造/资源绑定调用；只迁移到下列对象接口，沿用其装配与 Adapter 资源实现
- examples/plugins/{showcase,lifecycle,echo,configdb}/ 的 Go 源文件；showcase/config_schema.json 可增加子插件的同名 object；alonabot_gen.go 只由生成器重新生成
- internal/core/adapter/remote/remote_test.go：TestMain 的 sdk.Run 调用与 loopPlugin 的注册写法迁移，保持既有断言；新增一个插件对象主路径集成测试
- 上述 SDK 包的 *_local_test.go，以及 internal/core/supervisor/plugin_local_test.go（//go:build unix）；临时插件源码和二进制在测试临时目录生成
允许的新跨模块边：无；SDK 内部可用 sdk/sqlitedb、sdk/gormdb -> sdk 登记迁移，sdk 根包继续经 sdk/internal/dbhook 接入，保持 GORM 按需链接。remote_test 沿用已有 SDK helper 与 Core 内部装配。

接口（照此实现；以下 Plugin、SubPlugin、Registrar、CommandGroup 与登记函数由本切片新增）：
sdk
   type Plugin struct { /* 未导出状态 */ }
   type SubPlugin struct { /* 未导出状态 */ }
   type Registrar interface { /* 未导出方法，仅 Plugin 与 SubPlugin 实现 */ }
   type CommandGroup struct { /* 承接原命令组的路径与权限状态 */ }
   func NewPlugin() *Plugin
   func UseGenerated(p *Plugin, g Generated)
   func (p *Plugin) Generated() Generated
   // Generated 保留 ConfigSchema、ServiceSchemas、I18n 等现有字段；Manifest 若由 goal-67 提供则复用，否则本切片追加。
   Manifest []byte // Generated 的字段，入口包 plugin.yaml 原文；不存在时为 nil。
   func Run(p *Plugin) error
   func Collect(p *Plugin) (*Registration, error)
   func (p *Plugin) OnStart(fn func(context.Context) error)
   func (p *Plugin) OnReady(fn func(context.Context) error)
   func (p *Plugin) OnStop(fn func(context.Context) error)
   func (p *Plugin) OnConfigLoad(fn func(context.Context, uint32, json.RawMessage) (json.RawMessage, error))
   func (p *Plugin) Config() config.Values
   func (c *Context) Config() config.Values
   func (p *Plugin) Sub(id string) *SubPlugin
   func (s *SubPlugin) Config() config.Values
   func (s *SubPlugin) Permission(p Permission) *SubPlugin
   func (g *CommandGroup) Permission(p Permission) *CommandGroup

   // 以下命令入口在 *Plugin、*SubPlugin、*CommandGroup 上同签名，接收者以 r 表示。
   func (r *Plugin) Group(name string) *CommandGroup
   func (r *Plugin) Command(path ...string) *CommandBuilder
   func (r *Plugin) Regex(routeID, expr string) *CommandBuilder
   func (r *Plugin) Keyword(routeID string, patterns ...string) *CommandBuilder
   func (r *Plugin) StartsWith(routeID string, patterns ...string) *CommandBuilder
   func (r *Plugin) EndsWith(routeID string, patterns ...string) *CommandBuilder
   func (r *Plugin) FullMatch(routeID string, patterns ...string) *CommandBuilder
   func (r *Plugin) Raw(routeID string) *CommandBuilder
   // 以下入口在 *Plugin、*SubPlugin 上同签名；HTTP 与 Schedule 沿用当前参数形式，T-04.11 再改为 builder。
   func (p *Plugin) On(routeID string, eventType EventType) *EventBuilder
   func (p *Plugin) Handle(method, path string, h http.Handler, opts ...HTTPOption)
   func (p *Plugin) Every(jobID string, d time.Duration, fn func(*Context) error)
   func (p *Plugin) Cron(jobID, expr, timezone string, fn func(*Context) error)
   func JSON[In, Out any](r Registrar, name string) *JSONService[In, Out]
   func Proto[In, Out proto.Message](r Registrar, name string) *ProtoService[In, Out]
   func Subscribe[T any](r Registrar, routeID, eventName string) *Subscription[T]
   func OnButton[T any](r Registrar, a ButtonAction[T]) *ButtonBuilder[T]
sdk/sqlitedb
   func Migrate(p *sdk.Plugin, fn func(context.Context, *sql.DB) error)
   func Migrations(p *sdk.Plugin, list []Migration)
sdk/gormdb
   func Migrate(p *sdk.Plugin, fn func(context.Context, *gorm.DB) error)
   func Migrations(p *sdk.Plugin, list []Migration)
   // 两包现有 Migration、Open、RunVersioned、AppliedMigrations 的公开签名保持。

sdk/generate
   Result 的 RootType 替换为 PluginConstructor string // 导入路径.New；只有 Adapter 实例声明且没有插件 New 时为空。
   其余 Result、Options、Scan 的公开字段与签名保持。
库模式调用方（goal-67 已合并时同步；下列 alonabot 为根包）：
   func alonabot.Plugin(p *sdk.Plugin) Option
   func sdk.RunInProcess(p *Plugin, l *inproc.Launch) error
   // 根包经 p.Generated() 读取资源，SDK 内部 generatedFor(p *Plugin) 从对象字段读取。
生成代码与作者的调用约定：
   // alonabot_gen.go：生成器输出，包名沿用入口包；sdk import 别名沿用 alonabotgen0。
   var alonabotGenerated = alonabotgen0.Generated{ /* 本包 schema、i18n、manifest 等 */ }
   // 作者源码：资源绑定后登记路由与回调；生成器只核对 New 签名。
   func New() *sdk.Plugin {
       p := sdk.NewPlugin()
       sdk.UseGenerated(p, alonabotGenerated)
       // 在 p 上登记。
       return p
   }
生成器不生成或包装 New，不以 init 调用 UseGenerated；删除插件的 generatedRegistrations 类型表。Adapter 的生成资源登记继续按 goal-67/D-134；旧生成文件沿现有规则排除；扫描/类型检查视图为入口包补入 alonabotGenerated 的 sdk.Generated 类型声明，使尚无 alonabot_gen.go 的首次生成可完成 New 的检查，不读取旧资源内容，不把该声明写回作者源码。

行为：
- 生成资源：UseGenerated 把 Generated 保存在 p 上，Collect、配置准备、i18n 与库模式均读取该对象；同一对象重复绑定由 Collect 报错，Run 开始后绑定沿注册规则 panic。资源作为只读构建产物共用，每次 New 返回独立插件状态，不同包各自的 alonabotGenerated 互不覆盖；没有资源的字段沿现有 nil 语义，不伪造 schema 或 manifest。生成器保留所有已有资源，plugin.yaml 存在时将原文写入 Generated.Manifest。
- 删除 sdk.Base、sdk.Root、六个根注册方法 Commands/Events/Services/Subscriptions/HTTP/Schedules 与六个注册器类型；命令组由 CommandGroup 承接，builder 与 Handle 的既有校验保留。配置、回调、迁移、路由与子插件登记均随插件对象持有，重复调用 New 得到独立对象；sdk.Run 开始后再登记或修改已有 builder/命令组时 panic，消息为英文。
- OnStart/OnReady/OnStop/OnConfigLoad 各至多登记一次，重复登记由 Collect 报错；未登记的生命周期回调跳过，未登记 OnConfigLoad 时原样使用 raw。保留配置迁移、补值、校验、DB 迁移、OnStart、REGISTER、OnReady/ACTIVATED 的顺序与 runtime.lock 门禁。
- p.Config() 与 c.Config() 返回整个插件的 config.Values，读取时机保持 OnConfigLoad 之后、OnStart 之前就绪。s.Config() 可在注册时取得并保存，随后读取同名顶层 object；准备前读取 panic。Collect 检查调用过 s.Config() 的子插件在 schema 中有同名 object，错误指明 ID。Handler 内 c.Config() 始终指向整个插件配置。
- sqlitedb 与 gormdb 四个登记函数合计至多调用一次，Collect 核对重复登记与迁移列表（非零、唯一、递增 ID 和有效 Up）；两个包的 SelfMigrator、VersionedMigrator、init 识别与 gormdb.CheckMigrators 的旧根对象入口一并删除，同步 versioned.go 的校验调用。SDK 在准备阶段核对登记方式与 manifest 的 migration_mode；迁移事务、共用账本、启动原因与维护许可沿用 CONTRACTS §6。sdk 根包继续保持按需数据库依赖。
- 生成器核对入口包声明 func New() *sdk.Plugin，缺少或签名不符时给出源码位置；按同一模块构建依赖闭包中对四个迁移登记函数的调用识别 migration_mode，两种模式都出现时报构建诊断。保留 GOOS/GOARCH、tags、排除外部依赖/vendor/测试/旧生成文件的扫描规则，以及 Service Schema、配置与 T 调用的现有检查。
- 结构体参数按 CONTRACTS §9.1：字段声明顺序即参数位置；名字取 arg 标签，否则按首段大写转小写的规则生成（URL/url、FeedURL/feedURL、URLPath/urlPath），重复名报错；指针缺失为 nil，default 表示可选且与指针互斥；arg:",raw" 或 arg:"name,raw" 用于末位 string。必需先于可选，raw 与 message 为末位且互斥，user/image/message 没有默认值。未导出、内嵌、嵌套或不支持类型的字段在 Collect 报错并指明字段；结构体与位置参数混用或同时调用 Args/Optional/RawTail 报错。
- 两种参数形式都接受 int 与底层为 string/int64/int/float64/bool 的具名类型，沿用 sdk.UserID、*message.ImageSegment、message.Chain 的参数语义；time.Duration 及额外的文本解码接口按 D-115 排除。Collect 生成相同的 CommandParam；SDK 运行时只赋入 Core 已解析的值，保留位置参数原有调用方式与自动型 (message.Content, error) 返回值。
- 子插件 ID 为非空 ASCII 字母、数字、_、-，不合法或重复由 Collect 报错；SubPlugin 只提供路由登记、缺省权限与配置视图。route_id、命令路径/别名、Service 名、HTTP method/path、job ID 按整个插件查重，不加子插件前缀；移动路由保持其标识。缺省权限按路由、由近到远的命令组、子插件、SDK 缺省解析；s.Permission 只影响命令与文本匹配。子插件 Handler 的 c.Logger() 增加 sub_plugin，REGISTER 保持现有协议。
- 同步四个现有示例与 remote_test.go 中的真实 SDK 插件，包级函数与方法值统一在 New 中登记；生成文件通过生成器更新。正式 build 入口与 init 模板由 goal-38 按 *sdk.Plugin、sdk.UseGenerated(p, alonabotGenerated) 和 sdk.Run(New()) 接入；本切片保留现有 plugin generate/check 可用。goal-67 正在按旧类型登记执行，若先合并则在上述范围中同步 alonabot.Plugin、sdk.RunInProcess、根包资源读取与 devbot；若后合并则按本切片接口同步库模式资源读取。docs/plugin-guide.md 与 ARCHITECTURE 的描述由审查者在合并前同步。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 internal/core/adapter/remote/remote_test.go 的 TestPluginObjectCommand，一个主路径集成测试。复用现有测试二进制、真实 SDK 插件 Session、Router 与回环 Adapter；插件构造时显式绑定配置 schema 与 i18n 资源，登记一个子插件的结构体参数命令，经配置准备与激活后交付命令，核对字段绑定、子插件配置与翻译回复。临时资源沿现有 helper 构造，不新增 testdata 插件。既有测试仅迁移注册与 sdk.Run 调用，断言不变。
- 本地自测：sdk/*_local_test.go、sdk/generate/*_local_test.go、sdk/{sqlitedb,gormdb}/*_local_test.go 覆盖 V-112 的重复登记、Run 后登记、New 多次构造、两种参数的等价描述与绑定、结构体非法声明、子插件 ID/标识/权限/配置视图/日志、迁移互斥与 manifest 不一致；资源绑定覆盖重复 New、不同包的 schema/i18n/manifest 隔离、重复绑定、Run 后绑定、缺少所需资源时的原有错误、生成源码不含插件 init/类型登记或 New 包装、PluginConstructor/CLI 判定、删除旧生成文件后首次生成与重复生成一致性；goal-67 已合并时本地验证库模式读取同一对象资源。
- 本地集成：internal/core/supervisor/plugin_local_test.go 在临时目录生成并编译插件，用真实 Core 与 memory Adapter 完成 New→注册→配置/DB 迁移→激活→结构体命令→回复→停止；核对回调次序、子插件配置和 sub_plugin 日志。sqlite 与 gorm 构建分别验证，迁移失败保留完整错误与未完成账本；不新增 testdata 插件。结果写入 PR。

完成条件：V-112 的本切片分项有本地证据；gofmt -l 对改动 Go 文件无输出；go generate ./... 后生成文件一致；go vet ./...；go test -race ./sdk/... ./internal/core/... ./cmd/...；go build ./...；go build -tags configdb_gorm ./examples/plugins/configdb。重复路由被覆盖、扫描外部依赖、参数绑定改变既有消费语义、准备阶段迁移或配置错误被吞掉时，按 PLAN T-04 的停止条件暂停交付。

不做：协议或 Core 行为改造；Hook（T-05.3）；Rule（T-05.2）；HTTP/Schedule 的 builder、参数名生成与订阅 ID 缺省（T-04.11）；Service/Event Bus 的带类型声明（T-04.13）；build/init 的产品实现；根包新增能力（只同步已合入库模式调用方）；依赖新增；向 sdk/ 或 internal/core/supervisor/ 提交测试、增加 testdata 插件。
```

### goal-16 · T-05.2 Rule、HandlerOutcome 与传播

状态：待派发。前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-05.1（PR #67）、T-04.7（PR #89）已合并。2026-11-12 按当前 builder、SDK 自有类型、Permission 与 HandlerOutcome 改写，自动型显式 Send 沿用 D-70 和 MESSAGE-MODEL §3；PLAN T-05 步骤 3 中旧的自动型 Send 限制按 T-04.7 已完成行、同章 V-14 与现行消息契约执行。与 goal-125、goal-68 并行时，后合并者同步 main，在已有注册与装配上保留本切片的 Rule 和调用生命周期语义。

```text
/goal 使用 $alonabot-execute 推进 goal-16（切片 T-05.2：Rule、HandlerOutcome 与传播）。goal 目录 goal-16，分支 goal-16，worktree .worktrees/goal-16，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：SDK 在同一 Handler RPC 内先执行 Rule，再执行目标函数；多个 Rule 默认 AND，用 Any 显式 OR，提供 ToMe。固定两种命令 Handler 的 HandlerOutcome、StopPropagation 锁存、错误与已识别命令的消费语义；自动型与手动型均可显式 Send，自动型成功返回的非空内容由 Core 发送一次，排在显式发送之后。

依据：PLAN T-05 步骤 3、4 及其失败停止条件、T-04.7 已完成行；CONTRACTS §2.1、§2.2 第 3、6 条、§2.3、§8；MESSAGE-MODEL §1.1、§3、§4；WIRE-FORMAT §3.3 的 Rule 描述、§4 的 HandlerOutcome 载荷与 §5A 的 D-69 修订；DECISIONS D-69、D-70、D-86、D-108；TEST-MATRIX V-14、V-11 的 Rule 分项。自动型显式 Send 以 D-70、MESSAGE-MODEL §3 与 PLAN 的 V-14 为准。

范围（只新增或修改这些）：
- sdk/rule.go（新增）：Rule、Any、ToMe；sdk/registry.go、sdk/events.go、sdk/buttons.go 为已有 builder 增加 Rule 方法并在 Collect 固定注册顺序
- sdk/dispatch.go：businessHandler、invokeHandler 在同一请求中执行 Rule 与 Handler；sdk/sdk.go 的 Context.StopPropagation、Send 与 sdk/edit.go 的 Context.Edit 只做下述生命周期语义需要的相邻修改
- sdk/session.go：buildRegister 为命令、事件（含按钮）路由写入现有 RuleDescriptor
- internal/core/router/router.go 的 dispatch、internal/core/router/command.go 的 invokeCommand：按下文读取有效 HandlerOutcome，保持已识别命令的消费状态
- sdk/*_local_test.go 与 internal/core/supervisor/rule_local_test.go（//go:build unix）；临时插件源码和二进制在测试临时目录生成
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
sdk（以下 Rule、Any、ToMe 与三个 builder 的 Rule 方法由本切片新增）
   type Rule func(*Context) (bool, error)
   func (b *CommandBuilder) Rule(rules ...Rule) *CommandBuilder
   func (b *EventBuilder) Rule(rules ...Rule) *EventBuilder
   func (b *ButtonBuilder[T]) Rule(rules ...Rule) *ButtonBuilder[T]
   func Any(rules ...Rule) Rule
   func ToMe() Rule
现有调用点（保持公开签名）：
   func Collect(root Root) (*Registration, error)
   func (r *CommandRegistry) Command(path ...string) *CommandBuilder
   func (r *EventRegistry) On(routeID string, eventType EventType) *EventBuilder
   func OnButton[T any](r *EventRegistry, a ButtonAction[T]) *ButtonBuilder[T]
   func (c *Context) Event() *message.Event
   func (c *Context) StopPropagation()
   func (c *Context) Send(m message.Content, opts ...SendOption) (*message.SendResult, error)
   func (c *Context) Edit(ref message.MessageRef, m message.Content) (*message.SendResult, error)
内部接入点（当前签名）：
   func businessHandler(reg *Registration) wire.Handler
   func invokeHandler(ctx context.Context, in *wire.Inbound, event *pb.Event, autoReply bool, groups []string, handler func(*Context) (message.Content, error)) (response wire.Response)
   func buildRegister(reg *Registration, configJSON string, configVersion uint32, configSchema []byte) *pb.Register
Core：
   func New(opt Options) *Router
   func (r *Router) Attach(p Plugin, c Caller)
   func (r *Router) Deliver(ctx context.Context, ev *pb.Event) error
   func (r *Router) Close(ctx context.Context) error
goal-125 若先合并，Collect 改为收 *Plugin，命令、事件与 OnButton 的登记接收者按其 Plugin/SubPlugin/Registrar 实际接口接入；本切片的三个 builder.Rule 签名保持。内部函数可为传递 Rule 调整参数，并同步调用方；pb 只用于协议内部。

行为：
- builder.Rule 可多次调用，按调用顺序追加；同次参数顺序与各次登记顺序组成 AND，首个 false 或 error 即停止，无 Rule 时匹配。Any 按顺序短路，第一个 true 即为 true，全部 false 或没有参数时为 false，error 立即返回；nil Rule 在 Collect 报注册错误，Handle 后调用 Rule 沿用已有 builder 的顺序错误检查。
- ToMe 读取 c.Event() 的 message.created 与 *message.CreatedPayload.ToMe；其他事件或没有事件时为 false。权限继续用已有 Permission、sdk.ChatAdmin、sdk.Admin、builder.Permission 与 c.Has，由 Core 在 Rule 前准入。
- Rule 与目标函数在同一请求、同一 Context 中执行。开始执行 Rule 时 matched=false，全部通过后才改为 true 并调用 Handler；false/error/panic 均跳过目标函数、保留 matched=false，Rule false 时无 Chain。错误链中 *errcode.Error 的 Code/Message 按现有错误映射返回；其他错误转 errcode.Internal 并保留诊断，panic 记录 stack。Rule 与 Handler 共用现有调用边界的恢复、错误编码与 Context 失效逻辑。
- buildRegister 分别设置命令路由 CommandSpec.Rule.HasCustomRule 与事件路由 EventSpec.Rule.HasCustomRule（含按钮）；有已登记 Rule 时为 true。沿用现有 RuleDescriptor.has_custom_rule，require_chat_admin 已 reserved；权限写入现有 required_permissions，协议文件和生成代码保持。
- 传播：
  - StopPropagation 在有效调用内锁存，即使之后 Rule false/error/panic 或 Handler 返回错误/panic，HandlerOutcome 仍为 STOP，并可与 Envelope.ErrorInfo 同时返回。
  - Core 先读取有效 HandlerOutcome 中的传播结果；STOP 在 matched=false 或伴随错误时仍终止后续候选（含 D-44 中命令层之后的事件路由）。未 Stop 时，事件路由与文本匹配/Raw 的 matched=false 跳过当前项，业务错误以 WARN 记录后继续。
  - 已识别命令的胜出路由 Rule false 或出错时仍消费命令层（不回落到短路径、文本匹配或 Raw），不回复 usage；STOP 的效果同上。
- 命令沿用 func(*sdk.Context, ...) error 与 func(*sdk.Context, ...) (message.Content, error)；事件与按钮沿用各自既有的手动 Handler 签名。两种命令 Handler 均可多次 c.Send；Core 仅在 matched=true、调用成功且自动返回内容非空时发送一次，单片段与 Chain 共用现有转换。返回错误时即使带非空内容也不自动发送；显式 Send 已确认的消息和部分结果保持，Core 不补发、不撤销、不自动发送错误回复。
- Handler 返回或 panic 后 Context 失效并取消其派生 context，之后 Send、Edit 返回 errcode.FailedPrecondition；失效后的 StopPropagation 不影响已经返回的 HandlerOutcome。SDK 在请求边界固定 matched、propagation 与 Chain，后台 goroutine 持有旧 Context 无法继续借用父调用权限；独立后台发送沿用 sdk.Send(ctx, to message.Session, m message.Content, opts ...SendOption)。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86，本切片的主路径需要真实 SDK Handler；验证放入本地自测）。
- 本地自测（sdk/*_local_test.go，结果写入 PR）：三种 builder 的 Rule 链及 Handle 顺序、AND 与 Any 短路/空列表/错误、nil Rule、ToMe 的标准事件与其他事件、RuleDescriptor.HasCustomRule、Rule false/error/panic 与 Handler error/panic 的 matched/STOP/错误组合、单片段/Chain/nil 的自动返回，以及 Send/Edit/StopPropagation 的调用后失效。
- 本地集成（internal/core/supervisor/rule_local_test.go，//go:build unix）：在临时目录生成真实插件并编译，经 supervisor.Launch、router.New(Options{...})、Attach、memory Adapter 与真实 message.Service 运行，覆盖 V-14 和 V-11 的 Rule 分项。带 Rule(ToMe()) 的 sdk.MessageCreated 路由只接收 ToMe=true 的消息；Permission(sdk.ChatAdmin) 的命令先由 Core 拒绝，带自定义 false/error Rule 的已识别命令仍消费命令层、不发 usage、不落入短路径/文本匹配/Raw，未 Stop 时仍进事件层；keyword 的 Rule 出错后续候选继续；Rule 中 Stop 后 false/error/panic 和 Handler 中 Stop 后 error 都终止后续路由；带类型的 OnButton 回调先执行 Rule 再调用目标函数。
- 同一本地集成中，手动命令显式发送两条后出错，Adapter 只收到这两条；自动型显式发送两条再成功返回 message.Text，Adapter 按顺序收到三条、自动回复恰一次；自动型显式发送后返回非空内容与错误，仅保留显式消息。用同步信号在 Handler 返回后调用 Send/Edit/StopPropagation，核对 FailedPrecondition 与已返回传播结果；无 sleep 推测完成，不新增 testdata 插件。

完成条件：V-14、V-11 的本切片分项有本地测试与真实进程证据，PR 写明命令、平台和结果；gofmt -l 对改动 Go 文件无输出；go vet ./...；go test -race ./internal/core/... ./sdk/...；go build ./...。PLAN T-05 的 Stop 丢失、Rule false 仍执行、重复自动回复或已识别命令回落任一复现即停止交付。

不做：PreEvent 与出站 Hook（T-05.3）；Continue、Wait 与多轮等待（T-05.4）；新增权限 Rule 或角色枚举；协议字段及编号修改；插件对象重构（T-04.12）；向 sdk/ 或 internal/core/supervisor/ 提交测试、增加 testdata 插件。
```

### goal-27 · T-07.5 调度与 HTTP 启停集成及验收

状态：已完成（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`），证据见 TEST-MATRIX §3.2。派发时：前置 T-07.2（[PR #107](https://github.com/AlonaBot/AlonaBot/pull/107)，合并提交 `fc2c247`）与 T-07.4（PR #53）均已合并。2026-11-04 按 D-77～D-85 修订计划：两个 Host 与单 Host 主动内容的边界改为插件响应的沙箱头与 Set-Cookie 过滤，归 goal-98（T-07.7），本切片不再验收 Host 边界。2026-11-09 审查者按 main `3dec9a0` 的代码核对并改写，已核对：`internal/core/scheduler` 的 `New(Options{JobTimeout, Now})`、`Attach(p router.Plugin, activatedAt time.Time)`、`Detach(pluginID)`、`Released(pluginID)`、`Close()`、`Stats(pluginID, jobID) JobStats`，`wire.Request.Late`；`httpproxy.Server` 的 `New(auth)`、`Bind(pluginID, generationID, endpoint, token, routes)`、`Unbind(pluginID, generationID)`，未绑定时返回 503；`host.Instance.Stop(t0, reason, tree, timing)` 在内部完成冻结、排空与 STOPPING，`host.StopTiming{Drain, Total, OnStopAt}`；`supervisor.Launch`/`Running.Stop`/`Running.WaitExit` 与 `Spec`（目前没有调度与 HTTP 依赖）；`cmd/alonabot/main.go` 现在在 Launch 返回后自行 `proxy.Bind`，并在 Session 结束时 `proxy.Unbind`，没有装配 Scheduler，也没有为插件调用设置 `Spec.Handler`。D-86 之后 `internal/core/supervisor` 只保留不依赖 SDK API 的 `supervisor_test.go`（用测试二进制以原始协议帧扮演插件），本切片的提交测试写在那里。goal-98（T-07.7）也修改 `cmd/alonabot/main.go` 的代理装配，两者可以并行，后合并的一方按先合并的代码变基。

```text
/goal 使用 $alonabot-execute 推进 goal-27（切片 T-07.5：调度与 HTTP 启停集成及验收）。goal 目录 goal-27，分支 goal-27，worktree .worktrees/goal-27，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：把 Scheduler 与 HTTP 代理接入 T-02 的生命周期：插件 Activated 之后才 Attach/Bind；Draining 时停止新 tick 与新 HTTP 派发；停止阶段取消余留调用、撤销映射、关闭 Listener，全部复用全局 STOP_TOTAL 预算；Session 或代次变化后旧端点、凭证与运行记录不再接收新业务。完成 T-07 的时间、进程与 HTTP 生命周期验收，在 TEST-MATRIX §3.2 记录证据。

依据：PLAN T-07 步骤 7、8 与「停止和失败条件」；CONTRACTS §4.2、§4.3、§7.2；TEST-MATRIX V-20、V-21、V-22、V-46、V-47 的 T-07 分项。

范围（只新增或修改这些）：
- internal/core/host/stop.go：Instance.Stop 增加冻结与 STOPPING 两个可选回调（见接口），同步全部调用方
- internal/core/supervisor/：Spec 增加 Scheduler 与 HTTP 两个可选依赖；Launch 在激活之后、Running.Stop 与 WaitExit 在停止与退出时接入它们
- internal/core/scheduler/：增加 Drain（停止新 tick，保留运行槽）
- internal/core/httpproxy/：增加 Drain（新请求返回 503，等待在途请求到截止时间）
- cmd/alonabot/main.go：run 装配 Scheduler，把 Scheduler 与 proxy 交给 supervisor，删除 main 中自行 Bind/Unbind 的代码
- internal/core/supervisor/supervisor_test.go：原始协议帧扮演的测试插件增加 interval 调度与 HTTP Listener 两种故障模式
- spec/TEST-MATRIX.md §3.2：追加 T-07 的证据行（本切片获授权修改这一处）
- 其他 *_local_test.go（不提交）
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- internal/core/host
  type StopHooks struct { Frozen func(drainDeadline time.Time); Stopping func() } // 都可以为 nil
  func (in *Instance) Stop(t0 time.Time, reason pb.StopReason, tree Tree, timing StopTiming, hooks StopHooks) StopResult
- internal/core/supervisor
  Spec 增加字段：Scheduler *scheduler.Scheduler; HTTP *httpproxy.Server // 都可以为 nil
- internal/core/scheduler
  func (s *Scheduler) Drain(pluginID string) // 不再触发新的 tick；运行中的调用照常等待
- internal/core/httpproxy
  func (s *Server) Drain(pluginID string, deadline time.Time) // 新请求 503；在途请求最多等到 deadline，之后 Unbind 照常撤销

行为：
- 激活：supervisor 在 host.Activate 成功（ACTIVATED 已应用）之后才 Bind HTTP（Instance 的 HTTPListener、HTTPProxyToken 与 Routes）并 `Scheduler.Attach(router.FromInstance(in), 激活时刻)`；激活失败不 Attach、不 Bind。Bind 先于 Attach，日志 `plugin activated` 出现时两者都已生效。
- 停止（Running.Stop，复用同一个 STOP_TOTAL）：Frozen 回调里同时 Scheduler.Drain 与 HTTP.Drain（截止为 t0 + STOP_DRAIN）；Stopping 回调里 Scheduler.Detach 与 HTTP.Unbind（以本次 generation）；进程树清空后 Scheduler.Released。不为 Scheduler 或 HTTP 另起预算。多个插件并行停止时各自从同一 t0 计算截止。
- 崩溃与重启：WaitExit 发现 Session 断开或进程退出时 HTTP.Unbind（以旧 generation）与 Scheduler.Detach，旧进程树确认清空后 Released；重启成功后以新 Session 重新 Bind/Attach，旧 token 与端点不能再接请求。
- HTTP 失败不触发任何业务重试；未激活或已 Drain 的插件返回 503。

测试（按编码规范「测试提交规则」）：
- 提交：supervisor_test.go 新增 TestScheduleAndHTTPLifecycle（真实测试二进制进程，经 supervisor.Launch，TCP 传输），各为一个子测试：
  1) interval 1s 的 job 在激活后才开始触发，触发间隔与 Scheduler.Stats 一致；不响应的 job 超过 JOB_TIMEOUT 后槽不释放，迟到响应释放（Stats 计 Late）。
  2) Running.Stop：Drain 之后不再有新 tick，HTTP 新请求 503，Listener 端口在返回后已关闭；总耗时不超过 STOP_TOTAL。
  3) 进程崩溃后旧映射与运行记录不再接收请求，重启后新 Session 的路由可用，代理转发带的是新 Session 的 HTTPProxyToken（测试插件核对）。
  4) 两个插件并行停止，各自在 STOP_TOTAL 内结束。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：cron 触发时刻与 contracts/schedule 的期望一致（注入时钟）；Drain 的截止时间；激活失败不 Bind；重启后旧 generation 的 Unbind 不影响新映射；Scheduler 在 Detach 后不触发；伪造内部头、私有路由无会话、公开回调经公开 Listener 的结果；`alonabot run` 真实启动后调度与 HTTP 可用（e2e，本地）。
- 证据：TEST-MATRIX §3.2 追加 T-07 一行，写明测试位置、命令、平台与结果；Host 边界（goal-98）、首次认证后的真实管理页面与 T-09 装配才能完成的端到端项写为「未闭合」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：插件页面的沙箱头与单一主机名（T-07.7）；管理认证的接入（T-09.1 已完成，本切片不改 adminauth）；正式装配（T-09.4）；Agent 定时任务（T-12.6）。
```

### goal-39 · T-09.4 管理服务、安装记录与 serve

状态：等待前置（T-13.1）。T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）已合并：`internal/core/buttons` 的 `New(store, now, current)` 与 `Service.Issue`、`Resolve`、`Purge`；`message.New(adapters, p, lease, issue, requestTimeout)`；`router.Options.Buttons`；可选接口 `adapter.ExtensionChecker`；`manifest.CheckDependencies(m, Installed)`；SDK 的 `sdk.Action[T]`、`sdk.OnButton(r, a).Handle(func(*Context, T) error)`、`message.Buttons`、`message.Extension`；`Segment` oneof 已用到 `extension = 9`。装配用 `buttons.New` 创建服务（current 由 `Router.Caller` 与 `Router.Plugin(pluginID).Routes()` 实现），把 `Service.Issue` 传给 `message.New`、`Service.Resolve` 传给 `router.Options.Buttons`，卸载时调用 `Service.Purge`，并让 `BUTTON_TOKEN_TTL` 取 `tune.Values` 的生效值（目前读 `tune.Defaults()`）；正式安装的依赖检查调用 `manifest.CheckDependencies`，环检测由本切片完成。T-04.10（[PR #133](https://github.com/AlonaBot/AlonaBot/pull/133)，合并提交 `3d618cb`）与 T-03.24（[PR #135](https://github.com/AlonaBot/AlonaBot/pull/135)，合并提交 `f83ad95`）已合并：装配把 `settings.Admin{AdapterID, UserID}` 转为 `policy.Admin{AdapterID, UserID}` 后调用 `Policy.SetAdmins`，写回 admins 用 `settings.Admin.String()`；读取设置后调用 `adapter.Registry.SetDefaults(settings.DefaultInstances())`。T-08.12（[PR #114](https://github.com/AlonaBot/AlonaBot/pull/114)，合并提交 `a796766`）已合并。实例启动复用 `settings.File.PrepareInstances(adapterID, schema)`，返回可启动配置、`[]settings.InstanceProblem` 与备份/写回错误；`settings.BackupDir = "backups"`。T-07.7（[PR #116](https://github.com/AlonaBot/AlonaBot/pull/116)，合并提交 `e65eb4d`）已合并。装配使用 `Settings.BaseURL`（键 `base_url`）、`webhost.Options{BaseURL, Admin, Proxy, Now}` 与 `httpproxy.New(auth)`，私有路由经 `adminauth.Auth.Authorize`。T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`）已合并。装配把 Scheduler 与代理传入 `supervisor.Spec.Scheduler`/`HTTP`，由 `Launch`、`Running.Stop` 与 `WaitExit` 管理 Bind/Attach、Drain、撤销与清空后的 Released。合并后由审查者按 main `2c1fdcc` 核对。2026-11-10 按 D-127：卸载分为只卸载、`--purge-data` 与 `--purge` 三种清理范围，`Uninstall` 的 purge 参数由 bool 改为 `PurgeScope`；动态任务与调度状态的表由 goal-127 加入清理对象，本切片不涉及。2026-11-09 按 D-105：admin add/remove 与禁止的参数是用户字符串 `<adapter_id>:<external_user_id>` 与会话字符串，下文合入符号中的 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串） 由 goal-122 改为 `Admin{AdapterID, UserID}`，改写语句时按合入的代码核对。T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）已合并，合入的符号：`internal/core/settings` 的 `Load(dataDir) (*File, error)`、`Defaults()`、`File.Settings()`（深副本）、`File.CommandPrefixes(adapterInstanceID)`、`File.Update(edit func(*Document) error)`（Rename 为提交点，之后目录同步失败时仍发布快照并返回错误）、`File.Reload() (changed []string, err error)`、`Document.Set(path, value)`/`Delete(path)`，`Settings`（含 `Admins`、`CommandPrefixes`、`CommandPrefixesByInstance`、`DefaultLocale`、`LogSources`、`Adapters map[string]AdapterSettings{Version, Instances}`）与 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串）；`adminauth.FirstToken(ctx, st, dataDir, interactive)`、`FirstTokenFile` 与 `adminauth.New(store, now, baseURL, dataDir)`；`alonabot admin-token show`。T-08.8（PR #102，合并提交 `f685c92`）已合并，合入的符号：`contracts/i18n` 的 `Parse(files) (*Catalog, []Problem)`、`Catalog.Has`/`Format(locales, key, args)`、`Placeholders`；`sdk.Context.T`、后台 `sdk.T` 与 `Generated.I18n`；`internal/core/i18n.Messages()`；`host.Instance.Messages`/`DefaultLocale`；`router.Plugin` 的 `Messages()`/`DefaultLocale()`；`supervisor.Spec.I18n`/`DefaultLocale`；`Bootstrap.default_locale = 12`。语句中 goal-8、goal-31、goal-36、goal-37（packages）、goal-53、goal-67 提供的符号（manifest.CheckDependencies、secret.Store、adminauth.Auth、packages.Preflight、supervisor.Spec.DefaultLocale、根包的 assembly、remote.Process、policy 的 Admin 方法）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：`alonabot serve` 改为 `alonabot run`（D-77）；Adapter 实例与 Core Admin 的增删经 goal-97 的 settings 写回 `core_config.json`，实例凭证不再存为 Secret 引用（D-82）；首次 token 由 goal-97 实现；install 接受未签名的仓库来源并要求管理员确认（D-81）。前置合并后由审查者改写。2026-11-05：goal-67 首次装配 Router 时 prefixes 传 nil，goal-97 只提供 settings，命令前缀与 Core Admin 列表由本切片接入（见范围、行为与测试）。2026-11-06：默认 locale 也由本切片接入，前置加入提供 Spec.DefaultLocale 的 T-08.8；Router 经 `Plugin.DefaultLocale()` 读取该插件 Session 的 Bootstrap 快照（goal-53 已合入），装配处只需填好 Spec.DefaultLocale。2026-11-09 按 D-102 追加一项装配：goal-121（T-03.24）合并后，装配在读取设置后调用 `Registry.SetDefaults(settings.DefaultInstances())`，设置重新读取后（goal-69）同样调用；goal-121 先于本切片合并时，审查者把这一步写进接口与行为段。

```text
/goal 使用 $alonabot-execute 推进 goal-39（切片 T-09.4：管理服务、安装记录与 serve）。goal 目录 goal-39，分支 goal-39，worktree .worktrees/goal-39，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立唯一的管理服务，CLI 与（之后的）WebUI 都经它的 HTTP API 操作；四种角色的组件使用同一套 install、list、enable、disable、restart、uninstall（含 --purge-data 与 --purge）、pin 流程，所有变更按包串行；install 按角色执行同类唯一，`--source` 指定并持久化来源，`--replace` 显式替换同类组件；Adapter 实例的创建、启停与删除经此服务，配置与凭证写回 core_config.json 的 adapters；启用时检查必需依赖，停用、卸载被依赖的包时列出受影响者并要求显式确认；admin add/remove/list；alonabot run（不带子命令时相同）复用根包的 Core 装配，按安装记录启动全部已启用的包。reload、update、rollback 的入口与校验在此定义，候选编排由 T-09.5 起接入。

依据：PLAN T-09 步骤 3；CONTRACTS §2.2、§2.3、§9.1～§9.3、§8（依赖段）、§15.1、§15.2、§7.4、§5.3、§5.4；DECISIONS D-26、D-42、D-59、D-82、D-127；TEST-MATRIX V-40、V-48、V-52、V-82、V-01 的安装分项、V-88 的接入分项。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移的 install 表定义中加入新列、末尾加入新表，安装记录的写入与读取方法同步新列；新增 internal/core/state/installs_admin.go（安装记录的扩展字段）
- 新增 internal/core/manage/（管理服务与 HTTP API）
- 根包 assembly.go：增加按安装记录启动全部已启用包、挂载管理 API 的步骤，读取 core_config.json 并接入命令前缀、Core Admin 列表与默认 locale；新增导出的 Serve 供 cmd 使用
- cmd/alonabot/：serve（以及不带子命令时）调用根包 Serve；新增 install、list、enable、disable、restart、reload、update、rollback、uninstall、pin、adapter instance、admin 命令（经管理 API 访问）
- internal/core/buttons、conversation、policy、tool、secret：在卸载与 purge 中调用其已有的清理方法（只加调用点）
- 上述包的测试文件
允许的新跨模块边：无（cmd/alonabot -> 根包已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   迁移：install 表定义加入列 version TEXT NOT NULL, publisher_key_id TEXT NOT NULL, pin TEXT NOT NULL, uninstalled INTEGER NOT NULL（不设 DEFAULT，写入安装记录时给出全部列）；
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
   type PurgeScope int
   const (
       PurgeNone PurgeScope = iota // 只卸载：保留数据、配置与来源绑定
       PurgeData                   // 另外删除数据
       PurgeAll                    // 删除数据与配置，连同来源绑定与安装记录
   )
   func (s *Service) Uninstall(ctx context.Context, id string, purge PurgeScope, confirm bool) error
   func (s *Service) Pin(ctx context.Context, id, constraint string) error // 空串表示解除
   func (s *Service) Reload(ctx context.Context, id string) error
   func (s *Service) Update(ctx context.Context, id, archive, releaseJSON, signature string, confirm bool) error
   func (s *Service) Rollback(ctx context.Context, id, generation string, confirm bool) error
   func (s *Service) CreateInstance(ctx context.Context, adapterPackageID, instanceID string, config []byte) error // 配置与凭证经 settings.File.Update 写回 core_config.json，保留 ${...} 原写法
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
- enable：检查 manifest 的必需依赖（CheckDependencies，Installed 来自已启用的包与运行中的 Adapter 描述），有缺失时 FAILED_PRECONDITION 并列出全部缺失项；可选依赖缺失只记诊断。已有 Active 时以该代次启动；首次安装没有 Active 时以「首个候选」启动（新代次、首装 SQLite、配置为 {} 或编辑快照，提交后激活；Journal 由 T-09.5 接入，本切片用一行 shortcut 注释写明）。Adapter 包启用后按 core_config.json 的 adapters 读取该 Adapter 的实例，经 File.PrepareInstances 准备后逐个 StartInstance。
- disable：停止并保留安装与数据；被已启用包的必需依赖引用时，未带 confirm 返回 ErrNeedsConfirm（列出受影响者），确认后执行，依赖方继续运行，调用它返回 UNAVAILABLE。
- restart：以当前 Active 启动新 Session（不导入编辑快照，不产生新代次）。
- uninstall：同 disable 的确认规则；停止、解除运行注册、删除未被引用的程序包目录。清理范围按 CONTRACTS §9.3（D-127）：PurgeNone 保留数据、配置、来源绑定与 uninstalled=1 的记录。PurgeData（--purge-data）另外删除数据：全部代次（代次目录、SQLite 与代次内的配置快照）、会话、按钮 token、Adapter 的游标；保留配置、来源绑定与 uninstalled=1 的记录，之后同来源重装时首个候选的配置取编辑快照。PurgeAll（--purge）删除数据与配置（编辑快照、Pending、secret、策略、工具策略、发信许可、Adapter 实例）以及安装记录。两种清理都先列出将删除的对象，未带 confirm 返回 ErrNeedsConfirm（Affected 为对象列表；PurgeData 时编辑快照与 Active 的配置不同也列入），确认后删除，状态库中的对象在一个事务内删除；不删除插件自管的外部数据库与插件自管目录。清理对象的列表集中在一处，供之后的切片追加（goal-127 追加任务与调度状态）。
- pin：持久化版本约束（SemVer 范围，manifest 的语法规则），update 只接受满足 pin 的候选。
- reload/update/rollback：完成输入与前置校验（来源绑定、pin、版本不低于被依赖的 min_version 时需要 confirm），然后交给 Candidates；缺省实现返回 UNSUPPORTED。
- Adapter 实例：CreateInstance 按包内 adapter_config_schema.json 补缺失值并校验，经 File.Update 写回 adapters.<adapter_id>.instances，凭证保留原值或 ${...} 引用。启动前调用 File.PrepareInstances，逐个记录 InstanceProblem，返回的可启动配置交给 remote.Process.StartInstance(ctx, instanceID, configJSON)；备份或写回 error 原样返回。同一 adapter_id 下账号唯一由 remote 检查。DeleteInstance 经 File.Update 删除对应配置。实例配置与凭证不出现在 List 或日志中，管理读取按声明中的 secret 路径脱敏。
- Core 设置接入（assembly，Run 与 Serve 共用）：settings.Load 取得 *settings.File（goal-68 或 goal-69 已加入这一步时复用）；Router 的 prefixes（goal-15 合并后为 Options.Prefixes）传 File.CommandPrefixes，policy 传与 message.Service 相同的 *policy.Policy；启动任何包之前把 Settings().Admins 转为 []policy.Admin 并调用 SetAdmins。经 supervisor.Launch 启动组件时 Spec.DefaultLocale 取 Settings().DefaultLocale，同进程开发组件的 Bootstrap.default_locale 取同一个值；每次启动都读当前设置，修改后在之后建立的 Session 生效。Core 自己的文案（缺省 usage）由 Router 经 Plugin.DefaultLocale() 取该插件 Session 的 Bootstrap 快照（goal-53 已合入），不另外注入。
- admin add/remove 经 File.Update 写回 admins，成功后调用 SetAdmins。两步在同一把锁内执行，SetAdmins 用这次写回之后的列表。这两个操作没有包 ID，按包串行保护不到它们，不加锁时并发的 add 与 remove 会让内存列表与文件不一致。admin list 读 Policy.Admins。
- HTTP API（在 Core 的主机名下，全部经 adminauth.Require）：每个方法一个 JSON 端点，需要确认时返回 409 与 {"affected": [...]}；错误返回 {"code", "safe_message"}。CLI 命令从 --admin-url（缺省取设置文件的 base_url）与 ALONABOT_ADMIN_TOKEN 环境变量（或 --token-file）取得 Bearer token，不读取插件环境；确认通过 --yes 显式给出。CLI 不直接写状态库。
- WebUI 组件没有进程：enable/disable 只改记录，restart 返回 FAILED_PRECONDITION；文件的提供在 T-09.12。list 返回 role。
- serve：取得 core.lock；没有管理凭证时，stdin 与 stderr 都是终端则按 admin-token init 生成并显示一次后继续，否则拒绝启动并提示先运行 admin-token init；装配 Core（与 Run 共用 assembly），启动 HTTP Listener、webhost、管理 API，按 install_seq 顺序启动全部已启用的包。不带子命令运行 alonabot 等同 serve。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go（//go:build unix；多列理由：本切片首次接通管理 CLI、依赖、替换与设置写回，没有其他端到端入口）。测试中用 goal-37 的测试工具生成签名的插件包、Adapter 包与依赖它们的插件包，在临时 data_dir 运行 alonabot run（预先用测试钩子初始化管理 token），再用 CLI 命令操作：
  - install 后 list 显示未启用；enable 产生第一个 Active；restart 只换 Session（代次不变）。
  - 依赖 Adapter 的插件在 Adapter 未安装时 enable 失败并列出缺失项；安装 Adapter 后成功；disable Adapter 未加 --yes 时被拒并列出受影响插件，加 --yes 后执行，插件继续运行。
  - uninstall 后同来源重装复用数据；不同来源重装被拒；--purge-data 列出对象并在 --yes 后删除数据，secret 与策略仍在，同来源重装后以编辑快照为配置启动；--purge 列出对象并在 --yes 后删除数据、配置与安装记录。
  - 已安装 Agent 测试组件时安装另一个 Agent：不带 --replace 返回 ALREADY_EXISTS；带 --replace 与 --yes 后旧组件已卸载、新组件已安装并沿用旧组件的启用状态；替换 Adapter 时旧实例不出现在新 Adapter 下；不同 adapter_id 的两个 Adapter 可以并存。
  - 创建 Adapter 实例：配置写入 core_config.json，${...} 原写法保留；经 PrepareInstances 迁移、补值与校验后启动，失败实例单列且其余实例可启动；list 与日志中没有 token；实例可启停。
  - core_config.json 用 command_prefixes_by_instance 覆盖一个实例：该实例的消息按新前缀匹配命令，另一个实例仍用 command_prefixes；admins 中的用户发命令时插件收到的 Actor.permissions 含 admin，admin add 写回后新用户的下一条命令也含 admin。
  - core_config.json 的 default_locale 设为 zh-CN：没有 locale 的 Actor 触发命令时，插件的 Context.T 回复中文，参数错误的缺省 usage 也是中文。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：串行化（同一包的并发操作排队）；ErrNeedsConfirm 的 Affected 内容；API 的错误格式；没有管理凭证且非终端时 serve 拒绝启动；Package.State 的取值；--source 指向本地 Release 模拟服务：安装成功且来源被持久化；模拟服务上的签名被篡改时安装失败，已有组件不受影响；pin 持久化，重启 serve 后仍在；插件依赖成环的包在安装时被拒；reload、update、rollback 返回 UNSUPPORTED（等待 T-09.5）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：候选 Journal 与准备编排（T-09.5）；唯一提交点与激活门禁（T-09.6）；崩溃恢复、回退与回收（T-09.7）；WebUI 用的会话取回与上传接口（T-09.13）；管理页面（T-09.8、T-09.9）。
```

### goal-68 · T-13.2 开发组件的配置与数据库

状态：等待前置（T-13.1）。T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）已合并，合入的符号：`internal/core/settings` 的 `Load(dataDir) (*File, error)`、`Defaults()`、`File.Settings()`（深副本）、`File.CommandPrefixes(adapterInstanceID)`、`File.Update(edit func(*Document) error)`（Rename 为提交点，之后目录同步失败时仍发布快照并返回错误）、`File.Reload() (changed []string, err error)`、`Document.Set(path, value)`/`Delete(path)`，`Settings`（含 `Admins`、`CommandPrefixes`、`CommandPrefixesByInstance`、`DefaultLocale`、`LogSources`、`Adapters map[string]AdapterSettings{Version, Instances}`）与 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串）；`adminauth.FirstToken(ctx, st, dataDir, interactive)`、`FirstTokenFile` 与 `adminauth.New(store, now, adminBaseURL, dataDir)`；`alonabot admin-token show`。goal-95（PR #103，合并提交 `c840af7`）已合并，合入的接口：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)` 与 `Schema` 的 `Canonical`、`Hash`、`Version`、`Migrate(config, from)`、`Complete`、`Validate`、`SecretPaths`、`CheckI18n(config)`；`sdk/config` 的 `Prepare(ctx, load, schema, raw, fromVersion, targetVersion, restart)`、`Loader`、只读的 `Values`（`String`/`Int`/`Float`/`Bool`/`Strings`/`Object`/`Objects`/`Dict`/`Template`/`TemplateList`）与 `Template.Key`；`sdk.Base.Config() config.Values` 与根插件的 `OnConfigLoad(ctx, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`；`supervisor.Spec.ConfigSchema` 为 config_schema.json 的规范 JSON。生成器读取手写 `config_schema.json`，配置标签解析与 `config.Config`/`config.Base` 已删除。语句中 goal-30、goal-32、goal-67 提供的符号（sdk/config.Prepare、gormdb 的描述、根包 assembly 与 dev 安装记录）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：开发 Adapter 的实例取自 `core_config.json`（D-82）；WebUI 修改开发组件配置时写回 `config.json` 并重启该组件（Admin API 在 goal-93）；dev 来源同时拒绝 uninstall（D-77）。前置合并后由审查者改写。

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
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：config.json 非对象；dev.sqlite 不存在时的 activation_reason；显式设置级别时不被覆盖；日志中有该组件的 DEBUG 记录。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：热部署（T-13.3）；库模式示例与验收汇总（T-13.4）。
```

### goal-69 · T-13.3 开发运行与热部署

状态：等待前置（T-13.1）。T-09.16（[PR #140](https://github.com/AlonaBot/AlonaBot/pull/140)，合并提交 `896991b`）已合并：工程 manifest 用 `manifest.ParseProject`，包内 manifest 用 `Parse`，`Generate` 由工程 manifest 加 `BuildInfo` 生成包内 YAML。T-03.24（[PR #135](https://github.com/AlonaBot/AlonaBot/pull/135)，合并提交 `f83ad95`）已合并：重新读取设置后 `Reload` 的 changed 含 `adapters` 时调用 `adapter.Registry.SetDefaults(settings.DefaultInstances())`。T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）已合并，合入的符号：`internal/core/settings` 的 `Load(dataDir) (*File, error)`、`Defaults()`、`File.Settings()`（深副本）、`File.CommandPrefixes(adapterInstanceID)`、`File.Update(edit func(*Document) error)`（Rename 为提交点，之后目录同步失败时仍发布快照并返回错误）、`File.Reload() (changed []string, err error)`、`Document.Set(path, value)`/`Delete(path)`，`Settings`（含 `Admins`、`CommandPrefixes`、`CommandPrefixesByInstance`、`DefaultLocale`、`LogSources`、`Adapters map[string]AdapterSettings{Version, Instances}`）与 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串）；`adminauth.FirstToken(ctx, st, dataDir, interactive)`、`FirstTokenFile` 与 `adminauth.New(store, now, adminBaseURL, dataDir)`；`alonabot admin-token show`。2026-11-04 按 D-77 改写：热部署不再是库模式的 `Reload` 选项与 `ALONABOT_RELOAD`，而是 `alonabot run [目录...] --reload`，开发组件以独立进程运行。语句只写了目标、依据和要点，前置合并后由审查者按 goal-67 的装配、goal-97 的 settings 与 goal-100 的工程 manifest 写出接口与测试段。

目标：`alonabot run [目录...] [--reload]` 把给出的目录（没有给出且当前目录有 plugin.yaml 时为当前目录）作为开发组件运行：按 `alonabot build` 的方式构建到 `<data_dir>/dev/bin/`，插件、Adapter 与 Agent 以独立进程运行，WebUI 以目录中的 assets 代替已安装的 WebUI。`--reload` 时轮询源码、资源与配置，源码变化时重新构建并只重启变化的组件，编译失败时旧组件继续运行，只有配置变化时不重新构建。

依据：PLAN T-13 步骤 4；CONTRACTS §15.4、§15.1、§16（DEV_RELOAD_POLL）；DESIGN §S；DECISIONS D-77；TEST-MATRIX V-64 的独立进程分项、V-65。

要点：
- 组件身份取自目录中的工程 plugin.yaml（goal-100 的解析）；构建失败或 ID、adapter_id 与已安装的包冲突时，在启动任何组件前报错退出。
- 构建复用生成器与 `go build`，提供供 goal-38 的 `alonabot build` 复用的内部函数，不另写一套。
- 开发组件使用 goal-67 的 dev 来源安装记录与固定代次；进程的启动、停止与有限重启沿用 supervisor，与已安装组件相同。
- 轮询：每 DEV_RELOAD_POLL 用标准库检查每个组件目录中的 `.go` 文件、`//go:embed` 引用的文件、go.mod、go.sum、plugin.yaml、config_schema.json、adapter_config_schema.json 与 i18n/（排除数据目录、vendor、以 . 或 _ 开头的目录与 testdata），以及插件的 config.json 与 core_config.json 中该 Adapter 的实例；源码或资源变化时重新构建，成功后依次重启变化的组件；只有配置变化时只重启该组件，Adapter 只重启变化的实例；core_config.json 的其他变化交给 goal-97 的 Reload：返回的键含 admins 时调用 policy.SetAdmins，命令前缀经 File.CommandPrefixes 读当前设置，不需要额外处理；开发组件的 Spec.DefaultLocale 在每次启动时取当前设置，default_locale 变化后在之后建立的 Session 生效，不为此重启运行中的组件。
- 删除现在 `run` 的 `--plugin-id`、`--entry` 参数，不留兼容；`--data-dir`、`--stdio`、`--listen` 与 `--args` 保留。
- 提交的集成测试（//go:build unix）在临时目录用 `alonabot init` 生成插件工程并以 `--reload` 运行：改动回复文本后插件重启并回复新文本、Core 与回环 Adapter 不重启；写入语法错误时旧插件继续回复；只改 config.json 时不重新构建但新配置生效；改动 config_schema.json 后重新构建。

## 第 12 波

### goal-9 · T-03.10 Telegram 接入开发入口与故障矩阵

状态：等待前置（T-13.3）。T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）已合并：`internal/core/buttons` 的 `New(store, now, current)` 与 `Service.Issue`、`Resolve`、`Purge`；`message.New(adapters, p, lease, issue, requestTimeout)`；`router.Options.Buttons`；可选接口 `adapter.ExtensionChecker`；`manifest.CheckDependencies(m, Installed)`；SDK 的 `sdk.Action[T]`、`sdk.OnButton(r, a).Handle(func(*Context, T) error)`、`message.Buttons`、`message.Extension`；`Segment` oneof 已用到 `extension = 9`。T-08.12（[PR #114](https://github.com/AlonaBot/AlonaBot/pull/114)，合并提交 `a796766`）已合并。Telegram 实例声明已位于 `adapters/telegram/cmd/alonabot-adapter-telegram/adapter_config_schema.json`，含 `token`、`api_base_url`、`proxy_mode` 与 `proxy`；启动复用 `settings.File.PrepareInstances(adapterID, schema)` 的可启动配置与 `InstanceProblem`，再调用 `remote.Process.StartInstance(ctx, instanceID, configJSON)`。`sdk/generate.Scan` 与 `plugin check` 已检查实例声明。合并后由审查者按 main `2c1fdcc` 核对。T-03.9（[PR #118](https://github.com/AlonaBot/AlonaBot/pull/118)，合并提交 `1cbfb2e`）已合并。T-03.7（[PR #108](https://github.com/AlonaBot/AlonaBot/pull/108)，合并提交 `d89e9b7`）已合并，合入的符号：`Button.core_token = 4`（optional string，发送时校验 1～64 字节）；`adapters/telegram` 实例的 `Send`（整链预检、确定的物理计划、每个物理操作一个 PartResult；末条物理消息为相册且带按钮时首发前 UNSUPPORTED）与 message.send/markdown/album/reference/buttons 能力；`telegramtest` 的 `Request{Method, Params, Files}`、`Requests()`、`FailMethod`、`ZeroMessageID`、`Drop`，模拟服务已支持 editMessageReplyMarkup。T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）已合并，合入的符号：`internal/core/settings` 的 `Load(dataDir) (*File, error)`、`Defaults()`、`File.Settings()`（深副本）、`File.CommandPrefixes(adapterInstanceID)`、`File.Update(edit func(*Document) error)`（Rename 为提交点，之后目录同步失败时仍发布快照并返回错误）、`File.Reload() (changed []string, err error)`、`Document.Set(path, value)`/`Delete(path)`，`Settings`（含 `Admins`、`CommandPrefixes`、`CommandPrefixesByInstance`、`DefaultLocale`、`LogSources`、`Adapters map[string]AdapterSettings{Version, Instances}`）与 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串）；`adminauth.FirstToken(ctx, st, dataDir, interactive)`、`FirstTokenFile` 与 `adminauth.New(store, now, baseURL, dataDir)`；`alonabot admin-token show`。语句中 goal-4～goal-8、goal-50、goal-51 提供的符号（remote.Process、supervisor.Spec 的 Role、adapters/telegram、telegramtest、media.Store、buttons.Service、message 的 Send/Edit）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：开发入口改由 goal-69 的 `alonabot run <目录>` 提供，本切片不再给 run 加 `--adapter-id` 等参数；Adapter 实例写在 `core_config.json` 的 `adapters` 中，不再读取 instances/ 目录。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-9（切片 T-03.10：Telegram 接入开发入口与故障矩阵）。goal 目录 goal-9，分支 goal-9，worktree .worktrees/goal-9，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：alonabot run 可以同时启动一个 Adapter 包与一个插件，Adapter 的实例配置取自 <data_dir>/core_config.json 的 adapters.<adapter_id>.instances；提供 echo 与媒体示例；用本地 Bot API 模拟服务跑完 PLAN T-03 步骤 8 的发送与接收故障矩阵，并在 TEST-MATRIX §3.2 记录 T-03 的证据。真实 Telegram 没有测试账号时记为未运行。

依据：PLAN T-03 步骤 8 与「停止和失败条件」；CONTRACTS §8、§15.1（`adapters`）、§15.4；MESSAGE-MODEL §4、§6；TEST-MATRIX V-07、V-08、V-09、V-35、V-44、V-45、V-48、V-51～V-54 的 T-03 分项。

范围（只新增或修改这些）：
- cmd/alonabot/：复用 goal-69 的 alonabot run <目录...> 开发入口，读取 core_config.json，经 File.PrepareInstances 准备 Adapter 实例后启动
- 新增 examples/plugins/echo/ 的媒体回送能力（收到图片时以 incoming_ref 原样回送）；新增 examples/plugins/media/（上传一个内置图片并发送相册与按钮）
- adapters/telegram/telegramtest/：补齐矩阵所需的故障注入（getUpdates 409、提交后超时、相册后 markup 失败等）
- 新增 cmd/alonabot/telegram_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-03 的证据行（本切片获授权修改这一处；其他 spec 文件不改）
- 上述包的测试文件
允许的新跨模块边：cmd/alonabot 的测试 -> adapters/telegram/telegramtest（测试代码，审查者确认允许）。

接口：
- goal-69 合并后复用其 alonabot run <目录...> 入口与组件身份；实例读取使用 settings.Load 与 File.PrepareInstances，schema 取 Adapter 工程的 adapter_config_schema.json；返回的配置经 remote.Process.StartInstance 下发，失败实例按 InstanceProblem 定位。
- 不新增公开 Go 符号。

行为：
- 启动顺序：取得 core.lock → 启动 Adapter 进程并 Bind → 启动全部实例 → 启动插件。插件停止与 Core 退出时先停插件、再停 Adapter。Adapter 进程异常退出按有限重启策略恢复，Bind 后恢复已启用的实例。
- 实例配置的 token、proxy 在 core_config.json 保留原值或 ${...} 引用，PrepareInstances 在运行副本展开后下发；它们只出现在 adapter.instance.start 请求中，不写日志。
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
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：实例配置的非法 JSON、迁移与校验失败、空实例集，以及失败实例不影响其余可启动实例。
- 若维护者提供了测试账号：在 PR 中另附一次真实 Telegram 的手动运行记录（命令、时间、结果），不写入提交的测试。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：正式安装与实例管理（T-09.4）；其他平台的 Adapter。
```

### goal-17 · T-05.3 三阶段 Hook 与 send_forbidden

状态：等待前置（T-05.2）。2026-11-09 按 D-112 把 SDK 接口段改为链式注册与 goal-91 的 SDK 类型（位置参数与 `RouteOption` 已由 D-70 删除），AfterSend 的 Handler 增加 error 返回值，与 MESSAGE-MODEL §5「Handler 只返回成功或 ErrorInfo」一致；T-05.2 排在 T-04.8 之后，这些类型届时已合入。同日按 D-92、D-93 补充：出站 Hook 中禁止调用的服务除 Send、Edit 外还有 `core.action` 与 `core.delete`；其中已合并的（goal-106、goal-107）由本切片一并加入禁止名单，未合并的由它们自己加入，审查者改写语句时按合并情况写入。语句中 goal-2、goal-12、goal-15、goal-16、goal-51、goal-91 提供的符号（router 的派发链与可信信息、message.Service 的 Send 与 Edit、sdk 的分发、事件常量与 SDK 类型）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-17（切片 T-05.3：三阶段 Hook 与 send_forbidden）。goal 目录 goal-17，分支 goal-17，worktree .worktrees/goal-17，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：运行 PreEvent、BeforeSend、AfterSend 三类 Hook：PreEvent 只能改写业务 payload 或 Stop，失败时结束本次入站链；每次显式或自动 Send（以及 Edit）独立执行一组 BeforeSend → Adapter.Send → AfterSend，BeforeSend 只能改 Chain 或拒绝，改后重新预检；AfterSend 只观察。出站 Hook 的执行上下文带 send_forbidden，该标记保存在 Core 的父调用记录中并随派生调用继承，Hook 内直接 Send 被拒。

依据：PLAN T-05 步骤 5、6；CONTRACTS §2.3 第 3～5 条、§8（Edit 的 Hook 段）；WIRE-FORMAT §3.5（Hook 输入的追加）；MESSAGE-MODEL §5；DECISIONS D-70、D-104、D-112；TEST-MATRIX V-15 的直接路径分项、V-110 的 Hook 分项。

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
2. sdk（链式注册与 SDK 类型，D-70、D-112）
   Root 增加 Hook：Hooks(*HookRegistry) // func (*Base) Hooks(*HookRegistry) 默认不注册
   type HookRegistry struct{ ... }
   func (r *HookRegistry) PreEvent(routeID string, event EventType) *PreEventBuilder // 一条路由对应一种事件，event_type_filter 只含该事件名
   type PreEventBuilder struct{ ... }
   func (b *PreEventBuilder) Priority(p int32) *PreEventBuilder
   func (b *PreEventBuilder) Handle(h any) // func(*Context, P) (P, error)；P 为该事件的 payload 类型，规则与事件路由相同（标准事件为 goal-91 的载荷指针，AdapterEvent[T] 为 *T）；返回 nil 表示不改写
   func (r *HookRegistry) BeforeSend(routeID string) *BeforeSendBuilder
   type BeforeSendBuilder struct{ ... }
   func (b *BeforeSendBuilder) Priority(p int32) *BeforeSendBuilder
   func (b *BeforeSendBuilder) Handle(h func(c *Context, to message.Session, m message.Chain) (message.Chain, error)) // 返回 nil 表示不改写
   func (r *HookRegistry) AfterSend(routeID string) *AfterSendBuilder
   type AfterSendBuilder struct{ ... }
   func (b *AfterSendBuilder) Priority(p int32) *AfterSendBuilder
   func (b *AfterSendBuilder) Handle(h func(c *Context, to message.Session, res *message.SendResult, sendErr error) error) // sendErr 为本次发送的错误；返回的错误只记日志
   func (c *Context) Editing() message.MessageRef // 出站 Hook 中：本次是编辑时为被编辑消息的句柄（D-104），发送时为空；其他 Handler 中为空
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
- SDK：按 route_id 分发 Hook。Handle 的签名在 Collect 时核对，不符或漏写 Handle 时报错并带 route_id 与期望的签名。PreEvent 的 P 由 SDK 从 event_payload 解码、改写后编码回同一分支（AdapterEvent 的 *T 重新编码为 payload_json），返回 nil 表示不改写；BeforeSend 返回 nil 表示不改写；AfterSend 返回的错误以 ErrorInfo 回报。to 由请求的 target 转为会话字符串；c.Editing() 在 operation 为 EDIT 时由 target 与 edited_platform_message_id 组成句柄。Context.StopPropagation 在 PreEvent 中使 propagation=STOP；出站 Hook 的 Context 中 Send/Edit 由 Core 拒绝。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Handle 的错误签名与漏写 Handle 在 Collect 时报错（PreEvent 的 P 与事件不符、BeforeSend 与 AfterSend 的参数不符）；PreEvent 返回不匹配分支的 payload（fixture）；BeforeSend 修改后 Chain 预检失败；排序的稳定性；插件自报 send_forbidden 被忽略；InvocationContext 中的标记在派生请求中保留（接口 fixture）；internal/core/supervisor/hook_local_test.go（//go:build unix）。两个测试插件：h 注册 PreEvent（把 text 改成大写）、BeforeSend（在 Chain 末尾追加一段文本；遇到含 "deny" 的 Chain 返回错误；尝试在 Hook 内 Send）、AfterSend（记录结果后返回错误）；e 注册 echo 事件路由与一个命令：入站 "hi" 被 PreEvent 改为 "HI" 后到达 e；e 回复时 memory 收到追加了文本的 Chain；AfterSend 收到 ACCEPTED 的结果，它返回的错误不影响 e 收到的 SendResult；一次 Handler 内两次 Send：BeforeSend 与 AfterSend 各执行两次；含 "deny" 的回复被拒，memory 没有收到，e 收到 REJECTED 与错误；AfterSend 未执行；BeforeSend 内的 Send 返回 FAILED_PRECONDITION；PreEvent 返回错误时 e 的路由与命令都没有被调用；Edit 时出站 Hook 的 c.Editing() 为被编辑消息的句柄，Send 时为空。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：Hook→Service→Send 的间接递归回归（T-06.2 的必交用例，本切片只保证标记随父调用记录保存）；延续调用（T-05.4）。
```

### goal-52 · T-05.4 延续调用与多轮等待

状态：等待前置（T-05.2）。2026-11-10 按 D-119 加入 builder 的 `.Continue()`：接口、行为与本地自测各追加一条，`sdk/registry.go`、`sdk/events.go` 与按钮路由的 builder 加入范围；前置合并后由审查者按合入的 builder 类型核对。语句中 goal-2、goal-15、goal-16、goal-76 提供的符号（router 的派发链与 Invocation、sdk 的 Context 与分发、tune.Values）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-52（切片 T-05.4：延续调用与多轮等待）。goal 目录 goal-52，分支 goal-52，worktree .worktrees/goal-52，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：由真实入站事件触发、具有 Actor 与默认 Target 的 Handler 可以经 core.continuation.start 请求一次延续调用：原调用成功结束后，Core 以同一插件、同一 route_id 与原可信上下文发出新的 Handler 请求（continuation_of 为原 call_id），使用独立的 CONTINUATION_TIMEOUT 期限，不占用分发链。延续调用内可以经 core.wait 等待同一 Actor 在同一 Target 的下一条消息，命中的消息不进入普通分发。SDK 提供 Context.Continue 与 Context.Wait。

依据：PLAN T-05 步骤 7；CONTRACTS §2.4；WIRE-FORMAT §3.5；DESIGN（延续调用示例）；DECISIONS D-119；TEST-MATRIX V-50、V-113 的延续调用分项。

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
   func (b *CommandBuilder) Continue() *CommandBuilder // 命令、文本匹配与 Raw 路由：Handler 整体在延续调用中执行（D-119）
   func (b *EventBuilder) Continue() *EventBuilder     // 事件路由同上；按钮路由的 builder 有同名方法
   func (c *Context) Wait(timeout time.Duration) (*Context, error) // 只能在延续调用中调用；返回命中消息的 Context（可用 Event()、Text() 等）
   func (c *Context) Text() string // message.created 的文本片段连接结果；其他事件为空

行为：
- core.continuation.start（必须带 parent_call_id，指向本 Session 中仍在途的 Core→插件 Handler 调用）：以下情况返回 FAILED_PRECONDITION：父调用不是由真实入站事件触发（例如 scheduled、Bus、Service）、缺少 Actor 或默认 Target、父调用带 send_forbidden（出站 Hook 内）、父调用本身就是延续调用、同一父调用已请求过。本 Session 进行中的延续调用（已登记待发出与正在运行的）达到 Max 时返回 RESOURCE_EXHAUSTED。成功时只登记，返回空结果。
- 发出：Finished(ok=true) 时为登记的父调用发出新的 Handler REQUEST：同一插件、同一 route_id，HandlerRequest.continuation_of=原 call_id，event 为原事件（Core 保留的可信信息重建），context 为原调用的 Actor、默认 Target、scope、trace 与插件链；期限为发出时刻起 Timeout。它不经事件队列与派发链，不计入 DISPATCH_CONCURRENCY；返回值不参与原事件的传播。Finished(ok=false) 时丢弃登记，不发出。
- core.wait（必须带 parent_call_id，父调用必须是延续调用）：等待键为（adapter_instance_id、chat_id、thread_id（无则为空）、external_user_id），取自延续调用的默认 Target 与 Actor。同键已有等待 → ALREADY_EXISTS。timeout_ms 不在范围内 → INVALID_ARGUMENT。等待在 Core 内存中，不持久化。
- 匹配：Router 对每个 message.created，在 PreEvent 之后、命令层之前调用 Match：Target 与 Actor 构成的键命中等待时，该事件作为 WaitResponse 返回给等待者，并且不再进入命令层与事件路由。不同 Actor 或不同 thread 不命中。
- 结束：超时 → DEADLINE_EXCEEDED；父请求被取消 → CANCELLED；Release（插件停用或 Session 失效）→ 等待以 UNAVAILABLE 结束，待发出的延续调用丢弃，不补发。延续调用的期限到达时，其中的等待一并结束。
- builder 的 Continue（D-119）：原调用中 SDK 照常执行 Rule，匹配后请求 core.continuation.start 并以 matched、Continue 成功结束，不调用 Handler；请求失败时原调用返回该错误。收到 continuation_of 的请求时调用该路由的 Handler，参数与原调用相同；自动回复型 Handler 返回的非空内容由 SDK 在延续调用结束前经 Send 发出一次。Handler 中的 StopPropagation 不起作用，再调用 c.Continue 返回 FAILED_PRECONDITION。
- SDK：Continue 在 Handler 返回前请求 core.continuation.start，并登记 fn（按原 call_id）；收到 continuation_of 的请求时调用对应 fn。Continue 在延续调用中调用返回 FAILED_PRECONDITION。Wait 发起 core.wait，以本次延续调用为父请求。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：用 `.Continue()` 注册的 /quiz2 与 Handler 内调用的 /quiz 行为相同，Rule 不匹配时不请求延续，自动回复型的返回内容只发出一次，Handler 中的 StopPropagation 不改变原调用的传播；原调用返回错误时不发出；父调用带 send_forbidden 时被拒；timeout_ms 边界；Release 与超时的并发；SDK 中按 continuation_of 找不到 fn 时返回 NOT_FOUND；internal/core/supervisor/continuation_local_test.go（//go:build unix）。测试插件注册 /quiz 命令：Continue 中依次 Wait 两次并把收到的回答回复出来；另有一个 echo 事件路由用来证明消息是否进入普通分发：用户 A 发 /quiz 后，A 的下两条消息被 quiz 收到并回答，它们没有到达 echo；原命令的 HandlerOutcome 决定的传播不受影响；同一群里 A 与 B 同时 /quiz，各自只收到自己的回答；不同 thread 的消息不命中；Wait 超时返回 DEADLINE_EXCEEDED；同键第二个 Wait 返回 ALREADY_EXISTS；延续调用中再次 Continue 返回 FAILED_PRECONDITION；没有 Actor 的 scheduled 事件 fixture 中请求 Continue 返回 FAILED_PRECONDITION；用很小的 Max：超过的 Continue 返回 RESOURCE_EXHAUSTED；等待期间停止插件：等待结束，重启后不补发；延续调用执行期间，其他事件照常派发（不占分发链）。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：持久化的等待与延续调用（契约不提供）；Agent 的流式对话（T-12.2 使用本切片）。
```

### goal-124 · T-04.11 注册写法统一

状态：等待前置（T-04.12）。T-04.8（[PR #141](https://github.com/AlonaBot/AlonaBot/pull/141)，合并提交 `ac4aded`）已合并：SDK 公共 API 改用自有类型（`message.Session`、`message.Content`、`message.MessageRef`、事件常量与带类型 payload、`errcode`），pb 只在 `sdk/internal/convert`。T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`）已合并。`supervisor.Spec` 已有 `Scheduler *scheduler.Scheduler` 与 `HTTP *httpproxy.Server`，激活成功后先 Bind 再 Attach；`host.StopHooks{Frozen, Stopping}` 把 Drain 与撤销接入停止流程。SDK 当前仍为 `ScheduleRegistry.Every(jobID, d, fn)`、`Cron(jobID, expr, timezone, fn)` 与 `HTTPRegistry.Handle(method, path string, h http.Handler, opts ...HTTPOption)` 的位置参数注册，本切片在保留既有生命周期行为的基础上改为 builder。合并后由审查者按 main `2c1fdcc` 核对。2026-11-09 按 D-112 新增。2026-11-10 按 D-114 前置加入 T-04.12：要点中的 `r` 指插件对象或子插件（`*sdk.Plugin`、`*sdk.SubPlugin`），`sdk.Subscribe` 的第一个参数是 `sdk.Registrar`；Handler 用结构体接收参数时不取参数名（D-115）。T-04.8（goal-91）改写 sdk 的注册与类型（含 `Public(mode HTTPAuth) HTTPOption`），T-04.12（goal-125）改写注册的接收者，待两者合并后补全接口段。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。本切片与 goal-101、goal-123 都修改 sdk/generate，按执行顺序 §3 的生成器一行处理。

目标：定时任务与 HTTP 路由改为与 D-70 相同的链式注册；命令省略 Args 时参数名取 Handler 的参数名；Event Bus 订阅的路由 ID 缺省为事件名。协议不变。Hook 的同样改写由 T-05.3（goal-17）完成。

依据：PLAN T-04 步骤 11；DESIGN §D；DECISIONS D-70、D-79、D-112；TEST-MATRIX V-110 的 SDK 分项。

要点：
- 定时任务：`r.Every(jobID string, d time.Duration)` 与 `r.Cron(jobID, expr string)` 返回同一种 builder，Cron 可选 `.Timezone(name)`（缺省 UTC），以 `.Handle(func(*sdk.Context) error)` 结束；路由 ID 仍为 "job." 加 jobID。删除现有的位置参数形式。
- HTTP：`r.Route(method, path string)` 返回 builder，可选 `.Public(mode sdk.HTTPAuth)`，以 `.Handle(http.Handler)` 或 `.HandleFunc(func(http.ResponseWriter, *http.Request))` 结束；删除 `HTTPOption` 与包级的 `Public`。注册校验、路由 ID 与 Listener 的行为不变。
- 漏写 Handle 时 Collect 报错并带 route_id，与命令相同。
- 命令参数名：省略 `.Args` 时取 Handler 中 `*sdk.Context` 之后的参数名。Go 运行时取不到参数名，由生成器为插件包中第一个参数是 `*sdk.Context` 的包级函数、方法与函数字面量记录参数名，嵌入 Generated；SDK 按 Handler 的函数名（方法值去掉 `-fm` 后缀）或函数字面量的源码位置查找。写了 `Args` 时以它为准，个数仍须与参数一致；查不到时保持 arg1、arg2。`Optional` 的 name 指同一套名字。
- 订阅：`sdk.Subscribe[T](r, eventName)` 的路由 ID 缺省为事件名，`.ID(routeID)` 可以另给；同一事件订阅两次且都没有给 ID 时，Collect 报重复的 route_id。
- 调用方：examples/plugins 与测试中以字符串嵌入的插件源码按新写法修改，与参数名相同的 Args 删去。
- 测试：注册描述与旧写法的等价性、参数名的各种来源与查不到时的回退进本地自测；已有调用定时任务、HTTP 注册与 Subscribe 的集成测试改用新写法，断言不变。

### goal-126 · T-04.13 Service 与 Event Bus 的带类型声明

状态：等待前置（T-04.12）。2026-11-10 按 D-122 新增。T-04.12（goal-125）把注册的接收者改为 `sdk.Registrar`，合并前写接口段只能按将要改动的代码写。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。本切片与 goal-124 都修改 `sdk.Subscribe`，与 goal-101、goal-123、goal-124、goal-125 都修改 sdk/generate，后合并者同步 main 后按对方的写法重做自己的改动。

目标：Service 与 Event Bus 事件各用一个带类型的声明表示，注册、调用、发布与订阅都从声明出发，名字与 Go 类型只写一处；删除按字符串目标与类型参数直接调用、发布与订阅的函数。协议与 Core 不变。

依据：PLAN T-04 步骤 13；DESIGN §D、§G；CONTRACTS §3、§9.1；DECISIONS D-70、D-107、D-112、D-114、D-122；TEST-MATRIX V-113 的声明分项。

要点：
- 声明：`sdk.JSON[In, Out any](target string) Service[In, Out]` 与 `sdk.Proto[In, Out proto.Message](target string) Service[In, Out]`，`Service[In, Out]` 是只含未导出字段的值类型，记录 target 与编码。target 为 `plugin_id/service_name` 或 `@agent/service_name`。
- 注册：`sdk.Serve[In, Out any](r Registrar, s Service[In, Out])` 返回现有的 Service builder，`Access`、`Tool` 与 `Handle(func(*Context, In) (Out, error))` 不变；Protobuf 声明调用 `Tool` 照旧报错。Collect 核对 target 的组件部分等于 manifest 的插件 ID（`role: agent` 的组件为 `@agent`）、声明不是零值、同一 target 只注册一次，错误带 target。REGISTER 中的服务名仍是 service_name 部分，描述与原写法相同。
- 调用：`func (s Service[In, Out]) Call(ctx context.Context, in In) (Out, error)`，请求与原 `sdk.CallJSON`、`sdk.CallProto` 相同；零值声明返回 INVALID_ARGUMENT，不发出请求。
- Event Bus：`sdk.BusEvent[T any](name string) BusEventType[T]`，name 为 `plugin_id/event_name`；`func (e BusEventType[T]) Publish(ctx context.Context, payload T) (eventID string, err error)`；`sdk.Subscribe[T any](r Registrar, e BusEventType[T])` 返回现有的订阅 builder，路由 ID 缺省为事件名，`.ID(routeID)` 另给（goal-124 先合并时沿用它的缺省规则，后合并时由本切片实现）。
- 删除：`sdk.CallJSON`、`sdk.CallProto`、`sdk.Publish`，以及第一个参数为注册器或 Registrar、第二个参数为名字的 `sdk.JSON`、`sdk.Proto`、`sdk.Subscribe`，不留别名。
- 生成器：对类型参数的检查（包级的具名非泛型类型、Schema 生成）改为针对 `sdk.JSON`、`sdk.BusEvent` 的声明处，诊断位置随之变化；ServiceSchemas 的内容不变。
- `sdk/llm`（goal-83 先合并时）：导出 `llm.GenerateService = sdk.JSON[Request, Response]("@agent/llm.generate")`，`llm.Generate` 与 `llm.Ask` 的签名不变，内部调用它；Agent 组件用 `sdk.Serve(p, llm.GenerateService)` 注册。goal-83 后合并时按本条书写。
- 调用方：examples/plugins、测试中以字符串嵌入的插件源码与已合并的官方插件、Agent 代码按新写法修改。
- 文档同步：docs/plugin-guide.md 第 6 节由审查者在合并前改写。
- 测试：按 D-86 不向 sdk/ 提交测试。V-113 的声明分项进本地自测；已有调用 Service 与 Event Bus 的集成测试改用声明，断言不变。


### goal-129 · T-04.14 命令名覆盖的 Core 侧

状态：等待前置（T-03.17）。T-04.10（[PR #133](https://github.com/AlonaBot/AlonaBot/pull/133)，合并提交 `3d618cb`）已合并：policy 的 Admin 与全局禁止改为按 Adapter（`IsAdmin(ctx, adapterID, actor)`、`Deny`/`Undeny(ctx, kind, adapterID, subjectID)`），`global_deny` 按 `adapter_id`，`CheckCommand` 与 `CheckSend` 签名不变。2026-11-10 按 D-128 新增。T-04.4（PR #88）已合并，合入的符号：`policy.Scope`（`ScopePlugin`、`ScopeGroup`、`ScopeRoute`）、`policy.Override{Permissions, Enabled}`、`Policy.SetOverride`、`Policy.Overrides`、`Policy.CheckCommand`，状态库的 `command_policy` 表；Router 的命令索引从注册描述读取路径与别名（`internal/core/router/index.go`）。T-04.10（goal-122）修改 policy，T-03.17（goal-108）提供命令菜单同步，合并前写接口段只能按将要改动的代码写。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。本切片修改 `internal/core/state/schema.sql`，按执行顺序的共享文件规则处理冲突；与 goal-71（T-14.1）都涉及 `core.command.list` 返回的命令名，后合并者同步。

目标：管理员可以给一条按路径匹配的命令保存命令名覆盖，取代插件声明的路径与别名；Core 的匹配、遮蔽、usage 回复与命令菜单都使用生效的命令名，route_id、命令组与策略的作用域不变。本切片提供服务层与 Router 的行为，管理 API 由 goal-93 提供。

依据：PLAN T-04 步骤 14；CONTRACTS §2.2、§2.3、§8（命令菜单）、§9.3；WIRE-FORMAT §3.5（`core.command.list`）；DECISIONS D-96、D-127、D-128；TEST-MATRIX V-115 的 Core 分项。

要点：
- 存储：`command_policy` 表增加保存命令名列表的列（写进唯一的初始迁移），只用于路由作用域；命令名是 token 列表，列表按保存的顺序保留，第一项为主名。
- 服务层：在 policy 中增加按 plugin_id 与 route_id 设置、删除与读取命令名覆盖的方法，成功返回前已完成持久化。校验：列表非空；每个命令名按注册描述中命令路径的规则校验；列表内不重复；目标路由是按路径匹配的命令，文本匹配与 Raw 返回 INVALID_ARGUMENT；与本插件其他命令声明的路径或别名、或其他命令的覆盖相同时返回 ALREADY_EXISTS，safe_message 带冲突的 route_id。目标路由当前未注册时允许保存（插件可能未启用），只做格式与覆盖之间的检查，声明名的检查留到建索引时。
- Router：建立命令索引时，有覆盖的命令只用覆盖的命令名参与匹配；与本插件其他命令声明的名字相同的覆盖项不进入索引并记入诊断，一项不剩时用声明的名字。覆盖变化后重建索引，下一次派发生效，不重启插件。禁用的命令占有的是它生效的命令名。
- 不变的部分：route_id、参数、Handler、命令组归属；策略的命令组作用域仍按声明的规范路径；日志与策略中的标识不变。协议与 SDK 不改。
- usage 回复：命令与别名部分取生效的命令名，主名在前。
- 遮蔽诊断：按生效的命令名计算，并标出失效的覆盖项与指向未注册路由的覆盖。
- 命令菜单（goal-108）：命令名取生效的主名，主名只有一个 token 的命令才算根命令；覆盖变化与命令策略变化一样触发受影响实例的同步。
- 卸载：覆盖随命令策略归入配置，`--purge` 时删除（goal-39 的清理对象已含命令策略，本切片只核对新列随行删除）。
- 测试：提交一个集成测试，覆盖保存覆盖后旧名不再匹配、新名匹配且 Handler 收到相同参数，删除覆盖后恢复；其余（重名拒绝、空列表、文本匹配路由、插件更新后覆盖项失效、跨插件遮蔽、usage 与菜单中的名字、Core 重启后保留）进本地自测。

不做：管理 API（goal-93）；页面（goal-44）；按命令组整体改名；同一插件内两条命令互换名字；`core.command.list` 的实现（goal-71）。

### goal-130 · T-08.14 配置项的选项来源

状态：等待前置（T-04.12）。2026-11-10 按 D-129 新增。T-06.1（PR #45）与 T-08.11（PR #103）已合并；T-04.12（goal-125）提供插件对象与 `sdk.Registrar`，合并前写接口段只能按将要改动的代码写。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。本切片修改 Service 描述的 proto 与 `sdk/generate` 的 schema 检查，按执行顺序的共享文件规则处理冲突。

目标：`config_schema.json` 的字段可以用 `options_from` 声明选项来自 Core 内置来源或某个组件的选项来源服务；组件用 `sdk.ServeOptions` 提供来源；Core 能按组件与字段路径查询选项。本切片提供 schema 检查、协议标记、SDK 注册、Core 的查询入口与内置来源 `core.adapter_instances`；管理 API 由 goal-93 提供，页面由 goal-44 提供。

依据：PLAN T-08 步骤 11；CONTRACTS §3（选项来源）、§5.1（`options_from` 与「选项来源」）、§16（`OPTIONS_PAGE`）；WIRE-FORMAT §3.5 的追加字段；DECISIONS D-129；TEST-MATRIX V-116 的 schema、SDK 与 Core 分项。

要点：
- schema 检查：`options_from` 只用于 `string` 与没有 `items` 的 `list`；与 `options`、`labels`、`secret` 并存时报错；`source` 必须是 `core.<name>`、`<组件 ID>/<service_name>` 或 `@agent/<service_name>`；内置来源的名称与 filter 的键必须已知；`filter` 的值必须是字符串。错误带 JSON 路径。`adapter_config_schema.json` 走同一检查。
- 配置的行为不变：补缺失值、校验、`Config()` 的读取、Merge Patch 与 secret 都不因 `options_from` 改变；不按来源校验取值。
- 协议：Service 描述追加选项来源标记，字段号由本 goal 分配；Core 在 REGISTER 时核对带标记的服务使用 SDK 固定的选项查询 Schema，访问规则只含 Core 与本组件，不符时拒绝注册。
- SDK：`sdk.ServeOptions(r, serviceName, fn)`、`sdk.OptionsQuery{Search, Filter, Locale, Limit, Values}`、`sdk.OptionsPage{Options, More}`、`sdk.Option{Value, Label, Description}`。同一服务名注册两次或 `fn` 为 nil 时 `Collect` 报错。`Values` 非空时按存储值返回，其余字段不用。
- Core 的查询入口：输入组件 ID、字段路径与查询；从该组件 Active 代次的 schema 找到字段，取来源名与 filter。字段不存在或没有 `options_from` 返回 INVALID_ARGUMENT。内置来源在进程内查询；服务来源经 Service Broker 以 Core 的身份调用，不带用户 Actor，只接受带标记的服务。目标未安装、未激活、没有标记或调用失败返回 UNAVAILABLE。`Limit` 取 `OPTIONS_PAGE`；来源返回的选项多于上限时只取前面的并把 `More` 置为 true。
- 内置来源的登记处：按来源名登记查询函数与允许的 filter 键，供 schema 检查与查询入口共用。本切片登记 `core.adapter_instances`（存储值为 adapter_instance_id，显示名为实例名，说明为 adapter_id，filter 可用 `adapter_id`）；`core.chats` 由 goal-131 登记，本切片的 schema 检查先接受这个名称及其两个 filter 键。
- `internal/tune` 登记 `OPTIONS_PAGE`（缺省 50，范围 1～200）。
- 测试：提交一个集成测试，覆盖一个插件为自己的配置字段提供来源、Core 按字段路径查到选项、按存储值取回显示名、目标未激活时返回 UNAVAILABLE；其余（schema 检查的各种拒绝、超出上限的截断、未带标记的服务被拒、`@agent` 目标、Adapter 的 schema）进本地自测。

不做：管理 API（goal-93）；页面（goal-44）；`core.chats`（goal-131）；`llm.providers`（goal-83）；选项之间的级联；按来源校验已保存的值；用户与插件列表的内置来源。

### goal-127 · T-07.8 动态任务与调度持久化的协议和 Core 侧

状态：等待前置（T-09.4）。T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`）已合并。已合入 `Scheduler.Drain(pluginID string)`（停止新 tick，保留运行槽）、`Detach(pluginID)` 与 `Released(pluginID)`；`supervisor.Spec.Scheduler *scheduler.Scheduler` 在 ACTIVATED 后 Attach，`Running.Stop` 通过 `host.StopHooks.Frozen` Drain、`Stopping` Detach，停止或断链后只在确认进程树清空时 Released。合并后由审查者按 main `2c1fdcc` 核对。2026-11-10 按 D-124～D-127 新增。T-09.4（goal-39）将提供管理服务的卸载与清理对象列表，待其合并后补全接口段。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。本切片修改 `internal/core/state/schema.sql`，按执行顺序的共享文件规则处理冲突。

目标：Core 支持插件在运行时创建、替换、取消与列出带参数的定时任务实例，并能暂停、恢复静态任务；任务定义、周期任务的起点与暂停状态保存在 Core 状态库，Core 重启、插件重启与更新后仍在；错过的触发一律作废；卸载按清理范围删除或保留这些状态。SDK 的写法由 goal-128 提供。

依据：PLAN T-07 步骤 9；DESIGN §H；CONTRACTS §4.3、§9.3、§16；WIRE-FORMAT §3.3、§3.5；DECISIONS D-124～D-127；TEST-MATRIX V-114 的 Core 分项。

要点：
- 协议：RouteKind 追加 TASK，RouteDescriptor 追加 TaskSpec（只有任务名），ScheduledPayload 追加 task_name、task_key、task_payload_json；动态触发的 job_id 为 `task.<name>`。REGISTER 校验任务名的格式、同名重复与 `task.` 保留前缀。
- 状态库（写进唯一的初始迁移）：动态任务表以（plugin_id、任务名、key）为主键，保存触发规则（at、interval 或 cron）、时区、参数、起点与创建时刻；调度状态表以（plugin_id、job_id）为主键，保存静态 interval 任务的起点、对应的间隔与暂停标记。
- 调度器以状态库为准：Attach 时读出该插件的调度状态与任务实例；静态 interval 任务没有起点或间隔已变时写入本次激活时刻；按 contracts/schedule 计算下一未来触发。过期的一次性实例在 Attach 时删除，记 WARN（plugin_id、任务名、key、原定时刻）并计数。任务名未在本 Session 注册的实例保留、不安排。
- 六个固定服务：`core.task.set`、`core.task.cancel`、`core.task.list`、`core.job.pause`、`core.job.resume`、`core.job.status`，校验与错误码按 CONTRACTS §4.3；只在 Activated 之后可用，不需要父调用链，调用方只能操作自己的任务。先写状态库，成功后才更新内存中的定时器。
- 派发：动态任务经每插件 `TASK_CONCURRENCY` 的名额，其余按到期时间排队；排队中的周期实例到了下一个触发时刻仍未执行时只保留一次；同一实例上一轮未结束时跳过本轮。一次性实例在派发的同一事务中删除，失败与超时不重试。运行槽、期限与迟到 Response 沿用静态任务的规则（goal-24）。
- cancel 删除行、摘除定时器与排队项，并取消该实例在途的调用（协作式）。暂停只让之后的触发作废，不打断在途的一轮；恢复后从下一个未来时刻继续，起点不变；重复暂停或恢复成功且不改变状态。
- 启停：Draining 与停止时摘除该 Session 的全部定时器并清空队列，队列中的触发按作废处理；状态库中的行不动。
- 卸载：把两张表加入管理服务的清理对象，PurgeData 与 PurgeAll 时按 plugin_id 删除，与其余数据在同一个状态库事务内，并列入确认对象；删除不依赖插件进程，也不要求 Session 存在。PurgeNone 时行保留，没有 Session 所以不触发；重装并激活后周期任务从下一个未来时刻继续，期间到期的一次性实例按过期处理。
- 上限 `TASKS_PER_PLUGIN`、`TASK_PAYLOAD_BYTES`、`TASK_CONCURRENCY` 登记到 contracts/limits。
- 测试：用注入时钟与真实插件进程覆盖创建、替换、取消在途实例、列出、配额与参数上限；Core 重启后任务仍在且 interval 沿原起点触发；停用期间越过多次触发后不补跑，过期的 at 被删除并有 WARN；并发名额与排队；暂停跨重启保持；三种卸载范围下两张表的去留。

不做：SDK 的 `sdk.Task` 等写法（goal-128）；管理界面对任务的查看与清理；补执行错过的触发；动态任务的暂停。

### goal-84 · T-09.12 WebUI 组件与组件图标

状态：等待前置（T-09.4、T-13.1）。T-07.7（[PR #116](https://github.com/AlonaBot/AlonaBot/pull/116)，合并提交 `e65eb4d`）已合并。已合入 `settings.Settings.BaseURL`（键 `base_url`）、`webhost.Options{BaseURL, Admin, Proxy, Now}` 与 `adminauth.Auth.Authorize`；装配用 `httpproxy.New(auth)`，WebUI 静态入口挂到 Admin，`/plugins/` 由 Proxy 提供统一沙箱响应。私有插件请求直接使用管理 Cookie/Bearer，Cookie 的非 GET/HEAD 请求校验 CSRF 与 Origin；页面的 iframe 与 postMessage 桥仍由后继实现。合并后由审查者按 main `2c1fdcc` 核对。语句只写了目标、依据和要点。依据 D-62、D-63 新增；前置合并后由审查者按实际的 webhost、adminauth、manage 与根包装配写出接口和测试段。2026-11-04 随前后端拆分从 WebUI 章节移入核心章节。2026-11-04 按 D-77～D-85 修订计划：Core 只有一个主机名（D-84），WebUI 与插件路由同在其下；spec/ADMIN-API.md 还要写出插件页面桥的消息格式。

目标：Core 以组件方式提供 WebUI 与组件图标：已启用 WebUI 组件的静态文件（保留 `/api/` 与内部路径，无扩展名的不存在路径回退 index.html，未安装时 503 与英文提示），Admin API 主版本常量与 WebUI 声明范围的检查（不兼容时拒绝安装与启用并提示），只提供 manifest 声明的图标并带 sandbox CSP 与 nosniff、要求管理员会话；写出 spec/ADMIN-API.md（主版本、全部接口、认证与错误格式）；根包增加 `alonabot.WebUI(fsys fs.FS)` 选项。

依据：PLAN T-09 步骤 9、§3.1 的 T-09.12 行；CONTRACTS §7.3、§7.5、§9.1、§15.5；DECISIONS D-62、D-63；TEST-MATRIX V-83。

要点：
- 文件只从不可变包目录读取，路径按 CONTRACTS §9.1 的相对路径规则校验，不跟随链接，不经组件进程。
- 图标的类型只按扩展名取 image/svg+xml、image/png、image/webp；未声明的包内文件即使存在也返回 404。
- spec/ADMIN-API.md 由本切片新建，写入此时已有的全部 Admin API（goal-36、goal-39 等）；之后新增接口的切片（goal-66、goal-92、goal-93）同步追加；同一主版本内只做向后兼容的增加。
- 原 goal-43 要求 Core 给 WebUI 静态文件加 `Content-Security-Policy: default-src 'self'`、`X-Frame-Options: DENY` 与 `nosniff`。CONTRACTS §7.5 目前只规定了图标的响应头，没有这一条；补写接口段前先由维护者确认，确认后写入 §7.5 再列入本语句。
- 提交的集成测试经真实 Core 与临时安装的 WebUI 测试包覆盖 V-83 的全部场景；含脚本的 SVG 在浏览器中不执行的检查没有浏览器时记为未运行。

### goal-81 · T-16.1 插件测试套件基础

状态：等待前置（T-05.2、T-13.2）。2026-11-09 按 D-105，Harness 的 `Admin` 改为 `{ AdapterID, UserID }`（与 goal-122 的 policy.Admin 一致）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。语句中 goal-4、goal-16、goal-67、goal-68 提供的符号（sdk/adapter 的 Adapter 与 Instance、Router 的分发流程、alonabot.Start 与 Core、Plugin/Adapter/DataDir 选项、dev 的 config.json 与 dev.sqlite、生成器的 Generated）只写名字；合并后由审查者核对签名，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：测试实例写在 `core_config.json` 而不是 instances/ 文件；发信许可缺省为全部实例，套件选项用于收紧（D-82、D-83）。前置合并后由审查者改写。

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
  - 用 go test -race -count=2 ./plugintest/... 运行两遍都通过。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有 Generated 的根类型；同一 Harness 并发 Message；Start 失败（probe 的 OnStart 返回错误）后数据目录可以再次 Start；Dispatched 只在自动回复发送之后调用；Config 选项的值经 /config 返回；/count 连续两次回复 1 与 2；/edit 产生 Op 为 send 与 edit 的两条记录；去掉 message.edit 后 edit 被 Core 以 UNSUPPORTED 拒绝；两个使用 t.Parallel 的测试各自 Start，计数互不影响。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./plugintest/... ./internal/core/router/... ./internal/inproc/...；go test -race . ；go build ./...。

不做：按钮点击、多轮等待、schedule 触发与后台发送（T-16.2）；用套件测试示例插件与开发文档（T-16.2）；WebUI 沙盒（v0.1 之外）。
```

## 第 13 波

### goal-14 · T-04.5 示例升级与 T-04 验收

状态：等待前置（T-03.10、T-04.11、T-04.12、T-04.13）。T-04.8（[PR #141](https://github.com/AlonaBot/AlonaBot/pull/141)，合并提交 `ac4aded`）已合并：SDK 公共 API 改用自有类型（`message.Session`、`message.Content`、`message.MessageRef`、事件常量与带类型 payload、`errcode`），pb 只在 `sdk/internal/convert`。2026-11-10 按 D-118～D-123 前置加入 T-04.13：示例用 `c.Reply`、`.Continue()` 与带类型的 Service、Event Bus 声明，V-113 的证据在本切片汇总。2026-11-09 按 D-112 前置加入 T-04.11，示例按统一后的注册写法升级。2026-11-10 按 D-114～D-116 前置加入 T-04.12：示例以 `func New() *sdk.Plugin` 为入口，并各有一处结构体参数与子插件的用法，V-112 的证据在本切片汇总。T-04.9（PR #110，合并提交 `6aed25d`）已合并。T-04.4（PR #88）、T-04.7（PR #89）已合并。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。T-04.2、T-04.3 已合并；已核对 `CommandRegistry.Command`、`Regex`、`Keyword`、`Raw`，以及 Args、Optional、RawTail、Usage、Silent、IgnoreCase 和 Context.Match。策略与 Adapter 开发入口的实际接口仍待前置合并后核对。2026-11-04 按 D-77～D-85 修订计划：示例只用链式注册（D-79，goal-96 删除标记），用 `alonabot check`、`build` 检查与构建，经 `alonabot run <目录>` 运行（D-77）。前置合并后由审查者改写。

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
- 提交：cmd/alonabot/t04_e2e_test.go。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：示例的每条命令在 Windows/Linux 交叉编译（只编译，不运行）；生成文件的稳定性；sdk/generate 中对 testdata fixture 的测试。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；在两个示例目录运行 alonabot check 通过。

不做：真实 Rule 与 Hook 的集成（T-05.5）；WebUI 修改策略（T-09.9）。
```

### goal-18 · T-05.5 多进程事件流水线验收

状态：等待前置（T-05.3、T-05.4）。T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）已合并：`internal/core/buttons` 的 `New(store, now, current)` 与 `Service.Issue`、`Resolve`、`Purge`；`message.New(adapters, p, lease, issue, requestTimeout)`；`router.Options.Buttons`；可选接口 `adapter.ExtensionChecker`；`manifest.CheckDependencies(m, Installed)`；SDK 的 `sdk.Action[T]`、`sdk.OnButton(r, a).Handle(func(*Context, T) error)`、`message.Buttons`、`message.Extension`；`Segment` oneof 已用到 `extension = 9`。语句中 goal-8、goal-15～goal-17、goal-52 提供的符号（按钮、Rule、Hook、延续调用与等待、Router 的 Stats）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-18（切片 T-05.5：多进程事件流水线验收）。goal 目录 goal-18，分支 goal-18，worktree .worktrees/goal-18，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：用多个真实插件进程跑完整的事件流水线：多个 priority 与同优先级逆安装顺序、自定义 Rule、两种 Handler 签名、三个 Hook、真实按钮回调、两个用户在同一群同时等待、等待超时与等待期间停用插件；用恶意载荷尝试修改可信字段，用延迟 Handler 验证控制帧仍推进；在 TEST-MATRIX §3.2 记录 T-05 证据。

依据：PLAN T-05 步骤 8 与「停止和失败条件」；CONTRACTS §2；MESSAGE-MODEL §1、§6.3；TEST-MATRIX V-11、V-13～V-16、V-43、V-49、V-50 的 T-05 分项。

范围（只新增或修改这些）：
- 新增 internal/core/supervisor/testdata/pipeline/（流水线测试插件，按环境变量配置 priority、Rule、Hook 与命令）
- 新增 cmd/alonabot/t05_e2e_test.go（//go:build unix，经 Telegram 模拟服务运行）
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

状态：等待前置（T-05.3）。T-03.6（PR #55）、T-06.1（PR #45）已合并。语句中 goal-7、goal-17、goal-19 提供的符号（broker.Broker、router.Invocation 与 send_forbidden、media 的 Grant、wire 的 ParentAttach 与预算）只写名字；合并后由审查者核对，再改为待派发。

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
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：期限取最小值；没有父请求的后台调用只有插件身份；插件载荷中伪造的 actor 被忽略；Cancel 不越过其 Request；internal/core/supervisor/chain_local_test.go（//go:build unix）。构建 testdata/chainplugin（按环境变量决定收到调用后转调哪个插件、睡眠多久、是否在 BeforeSend 中调用 Service），以 a、b、c 启动：A→B→C 正确返回，C 看到的 Actor 与 Target 是原事件的，Chain 为 [a, b]；A→B→A 返回 CALL_CYCLE，A 的第二次函数未运行；两条同时进行的独立链 A→B 与 C→B 都成功；把 MaxDepth 设为 2：A→B→C 返回 FAILED_PRECONDITION；A 的父事件 Handler 到期：B、C 的未完成调用被取消（C 收到取消，日志可见），A 得到错误；没有重试；杀掉 B 的进程：A 立即得到 UNAVAILABLE；出站 Hook 中经 Service 再 Send 被拒（FAILED_PRECONDITION）；用很小的 wire 预算使转发请求额满：立即 RESOURCE_EXHAUSTED，期间其他 Session 的 Response 照常完成。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./sdk/...；go build ./...。

不做：Event Bus（T-06.3）；Session 清理的完整矩阵（T-06.4）。
```

### goal-66 · T-09.10 日志管理

状态：等待前置（T-09.4、T-09.12）。T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）已合并，合入的符号：`internal/core/settings` 的 `Load(dataDir) (*File, error)`、`Defaults()`、`File.Settings()`（深副本）、`File.CommandPrefixes(adapterInstanceID)`、`File.Update(edit func(*Document) error)`（Rename 为提交点，之后目录同步失败时仍发布快照并返回错误）、`File.Reload() (changed []string, err error)`、`Document.Set(path, value)`/`Delete(path)`，`Settings`（含 `Admins`、`CommandPrefixes`、`CommandPrefixesByInstance`、`DefaultLocale`、`LogSources`、`Adapters map[string]AdapterSettings{Version, Instances}`）与 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串）；`adminauth.FirstToken(ctx, st, dataDir, interactive)`、`FirstTokenFile` 与 `adminauth.New(store, now, adminBaseURL, dataDir)`；`alonabot admin-token show`。语句中 goal-2、goal-39、goal-65（logging 的 Output、Levels、Source）、goal-84（spec/ADMIN-API.md）提供的符号只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按前后端拆分改写：日志页移到 goal-94（T-09.15），本切片不再等待 T-09.8。2026-11-04 按 D-77～D-85 修订计划：按来源的日志级别写在 `core_config.json` 的 `log.sources`，经 goal-97 的 settings 写回（D-82），不存状态库。前置合并后由审查者改写。

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
  - 写入前一天日期的日志文件后，logs --date 能查到。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：环形缓冲的边界；历史文件解析（多行值、转义）；SSE 的断开；每个来源最多保留 LOG_BUFFER_RECORDS 条（用很小的值）。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交差异；go vet ./...；go test -race ./...；go build ./...。

不做：日志页（T-09.15）；远程日志收集与外部格式；日志文件设置的运行时修改（启动设置，§15.1）。
```

### goal-92 · T-09.13 管理页所需的会话与上传接口

状态：等待前置（T-09.4、T-09.12）。T-09.1（PR #72）已合并。2026-11-04 按前后端拆分新增，内容取自原 goal-43（T-09.8）语句中的后端部分：goal-36 合并的 adminauth 只有 login 与 logout，goal-39 的安装与更新只接受本地路径，浏览器中的 WebUI 用不了。语句只写了目标、依据和要点，前置合并后由审查者按实际的 adminauth、管理服务与 spec/ADMIN-API.md 写出接口和测试段。

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

### goal-128 · T-07.9 动态任务与暂停恢复的 SDK

状态：等待前置（T-04.12、T-07.8）。2026-11-10 按 D-124、D-126 新增。T-04.12（goal-125）提供 `sdk.Registrar`，T-07.8（goal-127）提供协议与六个固定服务，合并前写接口段只能按将要改动的代码写。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。

目标：插件作者用一个带类型的声明表示动态任务，在 `New()` 中注册处理函数，运行时用声明创建、替换、取消与列出实例；静态任务可以暂停、恢复与查询。协议与 Core 的行为由 goal-127 提供，本切片不改。

依据：PLAN T-07 步骤 10；DESIGN §H；CONTRACTS §4.3；WIRE-FORMAT §3.3、§3.5；DECISIONS D-107、D-114、D-124、D-126；TEST-MATRIX V-114 的 SDK 分项。

要点：
- 声明：`sdk.Task[T any](name string) TaskType[T]`，只含未导出字段的值类型；T 的要求与 Service 的 JSON 类型相同，编码为 JSON 对象。
- 注册：`sdk.OnTask[T any](r Registrar, t TaskType[T])` 返回 builder，`Handle(func(*Context, T) error)`。Collect 时登记 kind 为 task 的路由 `task.<name>`；声明为零值、同名重复、名称不合规使收集失败，错误带任务名。处理函数的 Context 没有 Actor；处理函数需要自己的 key 时，由作者用 `sdk.Key` 指定并写进参数。
- 创建：`t.At(ctx, when time.Time, args T, opts ...TaskOption) (key string, err error)`、`t.Every(ctx, d time.Duration, args T, opts ...TaskOption)`、`t.Cron(ctx, expr string, args T, opts ...TaskOption)`，调用 `core.task.set`。`sdk.Key(key)` 指定 key，相同 key 再次创建即替换；不指定时 SDK 生成随机 key。`sdk.Timezone(name)` 只对 Cron 有效，用在 At 或 Every 上返回 INVALID_ARGUMENT，不发出请求。
- 取消与列出：`t.Cancel(ctx, key) error`；`t.List(ctx) ([]TaskInstance[T], error)`，每项含 key、参数、触发规则与下一次触发时刻，SDK 取完全部分页。
- 触发到达时 SDK 按任务名找到处理函数，把参数解码为 T 后调用；解码失败返回 INVALID_ARGUMENT 并记日志，不调用处理函数。
- 静态任务：`sdk.PauseJob(ctx, jobID) error`、`sdk.ResumeJob(ctx, jobID) error`、`sdk.JobStatus(ctx, jobID) (JobStatus, error)`（是否暂停、上一次与下一次触发时刻）。
- 全部函数只在 Activated 之后可用，之前调用返回 FAILED_PRECONDITION；Core 返回的错误码原样交给调用方。
- 示例与文档：定时示例加入一个提醒命令，用 At 创建、用 Cancel 取消；docs/development.md 的定时任务一节补充动态任务与暂停恢复。
- 测试：SDK 与 Core 的集成测试覆盖 At、Every、Cron 各一次触发、相同 key 替换、Cancel、List 的带类型往返、参数解码失败、PauseJob 与 ResumeJob；选项的非法组合与 Collect 的失败情形进本地自测。

不做：协议、状态库与调度器（goal-127）；plugintest 对动态任务的立即触发（goal-82）；官方 Agent 改用动态任务（goal-63）。

### goal-131 · T-03.25 聊天目录

状态：等待前置（T-08.14、T-09.4）。2026-11-10 按 D-130 新增。T-03.3（PR #32）已合并。goal-130（T-08.14）提供内置来源的登记处，goal-39（T-09.4）提供卸载的清理对象，合并前写接口段只能按将要改动的代码写。语句只写了目标、依据和要点，前置合并后由审查者按合入代码写出范围、接口与测试段。本切片修改 Event 的 proto 与 `internal/core/state/schema.sql`，按执行顺序的共享文件规则处理冲突。

目标：Event 带上可选的会话显示名 `chat_title`；Core 在接纳入站事件时记录见过的聊天，并把它登记为内置选项来源 `core.chats`；聊天目录随 Adapter 的数据清理。两个官方 Adapter 的填写由 goal-132 与 goal-133 提供。

依据：PLAN T-03 步骤 22；CONTRACTS §8（聊天目录）、§5.1（`core.chats`）、§9.3（卸载的数据范围）；MESSAGE-MODEL §1（Event）；WIRE-FORMAT §3.5 的追加字段；DECISIONS D-127、D-130；TEST-MATRIX V-117 的 Core 分项。

要点：
- 协议：Event 追加可选的 `chat_title`，字段号由本 goal 分配；`sdk/adapter` 的交付接口带上它。没有 target 的事件带 `chat_title` 时 Core 拒绝该事件并记诊断。插件收到的 Event 带有该字段，SDK 不新增读取方法。
- 存储：状态库增加聊天表（写进唯一的初始迁移），主键为 adapter_id 与 chat_id，另存 chat_type、名称、最近交付它的 adapter_instance_id、首次与最近活跃时刻。话题不单独成行。
- 写入：接纳带 target 的入站事件时，在保存游标的同一个事务中写入或更新；事件不带名称时保留已有名称，chat_type 为 UNSPECIFIED 时保留已有类型。Core 生成的 bot.online、bot.offline 与 scheduled 事件不写。
- 来源 `core.chats`：存储值为不含话题的会话字符串 `<adapter_id>:<chat_id>`，转义规则同 MESSAGE-MODEL §1；显示名为名称，没有名称时为 chat_id；说明含 adapter_id 与 chat_type。搜索词不区分大小写地匹配名称与 chat_id；filter 支持 `chat_type`（`private`、`group`、`channel`）与 `adapter_id`；按最近活跃时刻从新到旧排列。按存储值取回时，格式不合法或不存在的值不返回。
- 清理：把聊天表加入 goal-39 的清理对象。`--purge-data` 与 `--purge` 卸载 Adapter 时，在同一个状态库事务中删除该 adapter_id 的行，确认对象中列出条数；只卸载与 `--replace` 时保留。
- 名称不写入日志；诊断只写 adapter_id 与 chat_id。
- 测试：提交一个集成测试（经回环测试 Adapter），覆盖群与私聊各交付一条带名称的事件后 `core.chats` 能按名称搜到、按 `chat_type` 筛选、之后一条不带名称的事件不清空名称；其余（没有 target 却带名称被拒、Core 重启后保留、两个实例交付同一聊天、`--purge-data` 后删除而只卸载保留、按存储值取回）进本地自测。

不做：给插件的聊天列表服务；读取 `chat_title` 的 SDK 方法；话题的目录；成员列表；主动向平台拉取全部群；管理 API（goal-93）与页面（goal-44）；两个官方 Adapter 的填写（goal-132、goal-133）。

## 第 14 波

### goal-82 · T-16.2 按钮、多轮等待、schedule 与示例测试

状态：等待前置（T-05.4、T-07.9、T-16.1）。T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）已合并：`internal/core/buttons` 的 `New(store, now, current)` 与 `Service.Issue`、`Resolve`、`Purge`；`message.New(adapters, p, lease, issue, requestTimeout)`；`router.Options.Buttons`；可选接口 `adapter.ExtensionChecker`；`manifest.CheckDependencies(m, Installed)`；SDK 的 `sdk.Action[T]`、`sdk.OnButton(r, a).Handle(func(*Context, T) error)`、`message.Buttons`、`message.Extension`；`Segment` oneof 已用到 `extension = 9`。2026-11-10 按 D-124：前置加入 T-07.9，由第 13 波移到第 14 波；套件另需支持立即触发一个动态任务实例并测试示例中的提醒命令，下面语句的接口与测试段尚未写入这一项，T-07.9 合并后由审查者补写。T-07.2（PR #107，合并提交 `fc2c247`）已合并，合入的符号：`internal/core/scheduler` 的 `New(Options{JobTimeout, Now})`、`Scheduler.Attach(router.Plugin, activatedAt)`/`Detach`/`Released`/`Close`/`Stats(pluginID, jobID) JobStats`；`wire.Request.Late`；SDK 的 `ScheduleRegistry.Every`/`Cron` 已随 REGISTER 声明并分发到 Handler。语句中 goal-8、goal-24、goal-52、goal-81 提供的符号（Button.core_token 与 button.clicked 的派发、调度器的运行入口、core.wait 的等待状态、plugintest 与 inproc.Harness）只写名字；合并后由审查者核对签名，再改为待派发。

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
  - go test -race -count=2 ./plugintest/... ./examples/... 两遍都通过。
- 证据：TEST-MATRIX §3.2 追加一行，写明覆盖的场景、命令、平台与 Go 版本。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Click 找不到 action_id；同 job 重叠时的 Trigger；等待超时后再次注入按普通消息处理；后台：probe 的 OnReady 启动的后台发送出现在 Background 中，读取后清空。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./plugintest/... ./examples/... ./internal/core/... ./internal/inproc/...；go test -race . ；go build ./...。

不做：模拟时钟或缩短 schedule 的周期（只提供立即运行）；WebUI 沙盒（v0.1 之外）；为 Adapter 作者提供的测试工具。
```

### goal-22 · T-06.4 Session 清理与 Service 端到端

状态：等待前置（T-06.2）。T-06.3（PR #56）已合并。语句中 goal-15、goal-17、goal-19～goal-21 提供的符号（router、broker、bus、hook 与各自的 Stats）只写名字；合并后由审查者核对，再改为待派发。T-06.3 已由 PR #56 合并，bus 的实际接口：`bus.New(opt bus.Options) *Bus`（`Options{QueueEvents int; QueueBytes int64}`，装配处传入 tune.Values 的值）、`(*Bus).Register(m *router.Mux)`、`Attach(p router.Plugin)`、`Detach(pluginID string)`、`Stats(pluginID string) Stats`（`Stats{Published, Delivered, Dropped uint64}`，丢弃不分原因；本切片要按原因计数时在 bus 中扩展）；Router 不持有 Bus，Attach/Detach 由装配处与 Router 的 Attach/Detach 一起调用；`router.Plugin` 增加了 `SendEvent`，wire 的接收丢弃计数为 `wire.Stats.DroppedEvents`。

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

### goal-38 · T-09.3 开发 CLI 与库模式脚手架

状态：等待前置（T-04.5、T-13.3）。T-09.16（[PR #140](https://github.com/AlonaBot/AlonaBot/pull/140)，合并提交 `896991b`）已合并：工程 manifest 用 `manifest.ParseProject`，包内 manifest 用 `Parse`，`Generate` 由工程 manifest 加 `BuildInfo` 生成包内 YAML。T-08.8（PR #102，合并提交 `f685c92`）已合并，合入的符号：`contracts/i18n` 的 `Parse(files) (*Catalog, []Problem)`、`Catalog.Has`/`Format(locales, key, args)`、`Placeholders`；`sdk.Context.T`、后台 `sdk.T` 与 `Generated.I18n`；`internal/core/i18n.Messages()`；`host.Instance.Messages`/`DefaultLocale`；`router.Plugin` 的 `Messages()`/`DefaultLocale()`；`supervisor.Spec.I18n`/`DefaultLocale`；`Bootstrap.default_locale = 12`。goal-95（PR #103，合并提交 `c840af7`）已合并，合入的接口：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)` 与 `Schema` 的 `Canonical`、`Hash`、`Version`、`Migrate(config, from)`、`Complete`、`Validate`、`SecretPaths`、`CheckI18n(config)`；`sdk/config` 的 `Prepare(ctx, load, schema, raw, fromVersion, targetVersion, restart)`、`Loader`、只读的 `Values`（`String`/`Int`/`Float`/`Bool`/`Strings`/`Object`/`Objects`/`Dict`/`Template`/`TemplateList`）与 `Template.Key`；`sdk.Base.Config() config.Values` 与根插件的 `OnConfigLoad(ctx, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`；`supervisor.Spec.ConfigSchema` 为 config_schema.json 的规范 JSON。生成器读取手写 `config_schema.json`，配置标签解析与 `config.Config`/`config.Base` 已删除。2026-11-04 按 D-76 改为核对与打包 config_schema.json。语句中 goal-11、goal-37、goal-53、goal-67、goal-95 提供的符号（sdk/generate、packages 的 Preflight 与包格式、contracts/release、contracts/i18n、根包 Run、contracts/configschema）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：开发命令改为顶层的 init、build、check、test、pack（D-77），build 取代 plugin generate，init 在当前目录创建工程、不生成 main.go，`.gitignore` 忽略 data/；包内 manifest 由 goal-100 的函数补上构建信息（D-78）；签名可选（D-81）；`run` 由 goal-69 实现；代码中提到旧命令名的提示（生成文件头的 `alonabot plugin generate`、`Collect` 报错中的命令）一并改为新命令。前置合并后由审查者改写。

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

### goal-55 · T-11.2 工具调用与用户确认

状态：等待前置（T-06.2）。T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）已合并：`internal/core/buttons` 的 `New(store, now, current)` 与 `Service.Issue`、`Resolve`、`Purge`；`message.New(adapters, p, lease, issue, requestTimeout)`；`router.Options.Buttons`；可选接口 `adapter.ExtensionChecker`；`manifest.CheckDependencies(m, Installed)`；SDK 的 `sdk.Action[T]`、`sdk.OnButton(r, a).Handle(func(*Context, T) error)`、`message.Buttons`、`message.Extension`；`Segment` oneof 已用到 `extension = 9`。用户确认按钮使用上述 SDK 按钮动作。T-08.8（PR #102，合并提交 `f685c92`）已合并，合入的符号：`contracts/i18n` 的 `Parse(files) (*Catalog, []Problem)`、`Catalog.Has`/`Format(locales, key, args)`、`Placeholders`；`sdk.Context.T`、后台 `sdk.T` 与 `Generated.I18n`；`internal/core/i18n.Messages()`；`host.Instance.Messages`/`DefaultLocale`；`router.Plugin` 的 `Messages()`/`DefaultLocale()`；`supervisor.Spec.I18n`/`DefaultLocale`；`Bootstrap.default_locale = 12`。语句中 goal-8、goal-20、goal-53 提供的符号（buttons.Service、broker 的父链调用、internal/core/i18n）只写名字；合并后由审查者核对，再改为待派发。goal-54（PR #101，合并提交 `de48f26`）已合并，合入的符号：`tool.Catalog`（`New(store, r)`、`Register(m)` 注册 `core.tool.list`、`Effective`、`SetPolicy`、`Policies`，可见性按 `router.ServiceCallable` 与生效策略非 DENY 判定）、`router.Plugins`、tool.proto 的 ToolPolicy/ToolSpec/ToolInfo/ToolListRequest/ToolListResponse、sdk 的 `ToolDefault`/`ToolInfo`/`ListTools`；`Catalog` 目前只持有状态库与 Router，invoke 需要的 Broker 等依赖在本切片按实际前置写入接口段。

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
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：参数截断；token 只能用一次；父调用结束时记录作废；zh-CN 文案；策略在两次调用之间修改；internal/core/supervisor/tool_invoke_local_test.go（//go:build unix）。在测试临时目录生成工具服务与测试插件（写法同 goal-54 的本地测试；收到消息时调用 InvokeTool 或 ConfirmOperation 并回复结果），memory Adapter 支持 buttons 与 Editor：allow 工具直接返回结果；deny 工具返回 PERMISSION_DENIED 且工具未执行；confirm 工具：memory 收到带两个按钮的确认消息，内容含实际 JSON 参数；同一用户点「允许」后工具执行并返回结果；另一用户先点击被拒、不影响结果；点「拒绝」返回 PERMISSION_DENIED 且工具未执行；超时（注入较短的 TOOL_CONFIRM_TIMEOUT）返回 DEADLINE_EXCEEDED；确认消息没有经过插件的 BeforeSend Hook（测试插件注册了会改链的 BeforeSend，确认消息不受影响）；没有 Actor 的调用（scheduled fixture）遇到 confirm 返回 PERMISSION_DENIED，没有发消息；不支持 buttons 的实例返回 UNSUPPORTED；arguments 不符合 input schema 时 INVALID_ARGUMENT；ConfirmOperation 展示 summary 而不是参数。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：Agent 的工具循环与 MCP（T-12.3）；「总是允许」（契约不提供）；管理页面的工具策略（T-09.9）。
```

## 第 15 波

### goal-40 · T-09.5 候选 Journal 与准备编排

状态：等待前置（T-09.4、T-06.4）。T-08.12（[PR #114](https://github.com/AlonaBot/AlonaBot/pull/114)，合并提交 `a796766`）已合并。Adapter 更新停止旧版本前调用 `settings.File.TrialInstances(adapterID string, schema *configschema.Schema) []settings.InstanceProblem`；它只读试算，不写回、不备份，成功提交后的迁移写回复用 `PrepareInstances` 与 `BackupDir`。T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`）已合并。冻结与停止沿用 `Running.Stop(t0, reason)`，经 `host.StopHooks{Frozen, Stopping}` 调用 `Scheduler.Drain(pluginID)` 与 `HTTP.Drain(pluginID, deadline)`，进入 STOPPING 后 Detach/Unbind，确认进程树清空后 Released。合并后由审查者按 main `2c1fdcc` 核对。T-08.6（PR #85）已合并。语句中 goal-22、goal-33、goal-39 提供的符号（manage.Service 与 Candidates 接口、candidate.PrepareDatabase/Discard 与交接约定）只写名字；合并后由审查者核对，再改为待派发。

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
1) 预检：update 的包已经过 packages.Preflight（staged 目录）；reload 用 config.Service.ImportSnapshot 导入编辑快照为 Pending（base 为当前 Active）；first_install 没有旧 Active。Adapter update 先从新包读取 adapter_config_schema.json，经 configschema.Parse 后调用 settings.File.TrialInstances；有 InstanceProblem 时在停止旧版之前拒绝，保留完整 Path 与 Reason。所有静态错误在停止旧版之前返回。
2) 写入 Journal（phase=started，expected_old=当前 Active，new_generation 为新 ID）。
3) 冻结：Router/Broker/Bus/Scheduler/HTTP 停止向该插件派发新的根调用（Router/Broker/Bus 沿用 goal-22；Scheduler/HTTP 由 Running.Stop 的 Frozen 回调调用各自的 Drain），按停止流程停止旧进程并等待进程树退出；phase=stopped_old。进程树没有在停止预算内清空时失败（不 sleep 猜测）。
4) candidate.PrepareDatabase（取得 runtime.lock 后复制，完成后释放）；AddGeneration(new, 未提交)；phase=candidate_ready。
5) 启动候选：supervisor.Launch（NewCandidate=true，Bootstrap 带 activation_reason（first_install/reload/update）、manual_migration_approved（只有管理员为本次 Session/Generation 显式授予时为 true）、阶段预算、Pending 解析后的配置；Spec.DefaultLocale 取当前设置，同 goal-39），REGISTER 校验与 secret 抽取后得到快照；SetGenerationConfig 写入快照与 secret 路径；phase=prepared。此时不提交 Active、不发 ACTIVATE。
6) manifest 声明 prepare_effects=external 的插件，已有安装的 reload 与 update 在步骤 2 之前检查：没有本次操作的维护许可时返回 FAILED_PRECONDITION（safe_message "NEEDS_MANUAL_MIGRATION"），不写 Journal、不停止旧版；有许可时 Bootstrap 的 manual_migration_approved=true，其余照常。first_install 不受此门槛限制。
- 普通 resume（restart、Core 重启后的恢复）不经过本编排，不写新 Journal。
- Abort：phase 改为 failed → 停止候选进程并确认退出 → candidate.Discard（只删未受保护的候选资源）→ 以旧 Active 启动旧版的新 Session（first_install 时保持未启用）→ phase=cleaned。Abort 本身失败时保留 Journal 供 T-09.7 恢复。
- 任一步失败都调用 Abort；错误返回给管理服务（安全文本）。
- 日志：每次 phase 变化以 INFO 记录（op_id、plugin_id、kind、phase）。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：phase 只能前进；expected_old 与写入一致；manual_migration_approved 只在显式授予时为 true；resume 不写 Journal；internal/core/supervisor/journal_local_test.go（//go:build unix）。测试插件按环境变量在 OnStart 写库后失败、或迁移失败：reload 成功运行到 PREPARED：Journal phase 依次前进，候选数据库是独立副本，旧进程已退出，Active 未变；候选 OnStart 失败：Abort 后旧版以新 Session 运行，旧库不含候选写入，候选目录已删除，phase=cleaned；旧进程不退出（忽略 STOPPING 且持有 runtime.lock 的测试插件）：编排在停止预算内失败，没有复制数据库，旧版状态可诊断；磁盘空间不足（用测试钩子使备份失败）：不激活，候选目录被清理；first_install：成功到 PREPARED；失败时插件保持未启用。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：唯一提交点与激活门禁（T-09.6）；崩溃恢复与 rollback（T-09.7）；维护许可的管理入口（T-09.14、T-09.9）。
```

### goal-57 · T-11.4 工具与会话的验收

状态：等待前置（T-11.2）。T-11.3（PR #35）已合并。语句中 goal-54～goal-56 提供的符号与测试插件只写名字；合并后由审查者核对，再改为待派发。goal-54 已合并，它的测试插件只在执行者本地（D-86），本切片需要时在测试临时目录生成。

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

## 第 16 波

### goal-41 · T-09.6 唯一提交点与激活门禁

状态：等待前置（T-09.5）。T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`）已合并。已核对 `supervisor.Spec.Scheduler`/`HTTP`：`Launch` 在 `host.Activate` 成功后才 `HTTP.Bind`、`Scheduler.Attach(router.FromInstance(in), activatedAt)`；`host.Instance.Stop(t0, reason, tree, timing, hooks)` 接收 `StopHooks{Frozen, Stopping}`，分别接入 Scheduler/HTTP 的 Drain 与 Detach/Unbind。合并后由审查者按 main `2c1fdcc` 核对。`host.Instance.State()`、`Err()` 与 `StateDegraded` 已存在；`journal.Orchestrator` 与 `Result` 待 goal-40 合并后核对，再改为待派发。

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
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：CommitCandidate 的每个前置条件；事务中没有文件与 RPC 操作（代码审查加测试钩子）；Degraded 状态下 Scheduler 不触发；internal/core/supervisor/commit_local_test.go（//go:build unix）：reload 成功：Active 改为新代次，COMMITTED 存在，Journal 为 committed；ACTIVATED 之前发送的消息没有派发给候选（候选的事件路由调用计数为 0，消息被旧版处理或因冻结被丢弃并计数），之后由新代次处理；提交前把 Active 改成别的值（模拟并发）：CommitCandidate 失败，候选被 Abort，旧版恢复；提交后 OnReady 失败：Active 仍为新代次，插件 Degraded，没有启动旧代次；HTTP 返回 503；提交后旧代次的按钮点击不调用插件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Core 崩溃后的恢复判定与 rollback（T-09.7）；管理页面（T-09.8）。
```

## 第 17 波

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
  - rollback 到上一代次：产生新代次与新的数据库副本，内容等于旧代次的数据；旧代次签发的按钮仍无效。多列理由：回退是本切片的主路径之一。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Recover 对每个 Journal 阶段的分支；保留代次的淘汰顺序；Collect 对 Pending 引用的保护；声明 prepare_effects=external 的测试插件 rollback 返回 NEEDS_MANUAL_MIGRATION；Collect 不删除保留代次与 Active，删除失败候选的残留。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：管理页面（T-09.8）；外部数据的备份工具（由管理员处理）。
```

### goal-93 · T-09.14 配置、权限、工具策略与会话的管理 API

状态：等待前置（T-03.25、T-04.14、T-08.14、T-09.6、T-09.12）。T-08.12（[PR #114](https://github.com/AlonaBot/AlonaBot/pull/114)，合并提交 `a796766`）已合并。已合入 `settings.File.PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []settings.InstanceProblem, error)`、`TrialInstances(adapterID string, schema *configschema.Schema) []settings.InstanceProblem` 与 `InstanceProblem{InstanceID, Path, Reason}`；准备结果包含展开环境变量、补缺失值并校验后的可启动配置，迁移全部成功才备份并写回，失败实例单列且其余实例仍返回；`settings.BackupDir = "backups"`。`sdk/generate.Scan` 已检查 `adapter_config_schema.json`，仅有实例声明而没有插件根类型时 `plugin check` 只检查该声明，`generate` 仍报缺根。T-07.7（[PR #116](https://github.com/AlonaBot/AlonaBot/pull/116)，合并提交 `e65eb4d`）已合并。Core 地址取 `Settings.BaseURL`，插件私有路由经 `adminauth.Auth.Authorize` 与 `httpproxy.New(auth)` 授权。合并后由审查者按 main `2c1fdcc` 核对。2026-11-10 按 D-129、D-130：增加配置字段的选项查询接口，前置加入 T-03.25 与 T-08.14；语句中 goal-130、goal-131 提供的查询入口与内置来源只写名字，合并后由审查者核对。同日按 D-128：增加命令名覆盖的接口，前置加入 T-04.14；语句中 goal-129 提供的 policy 方法只写名字，合并后由审查者核对。2026-11-09 按 D-105：/core/admins 的条目是用户字符串，/core/denies 的用户与会话分别是用户字符串与不含话题的会话字符串。T-07.6（[PR #100](https://github.com/AlonaBot/AlonaBot/pull/100)，合并提交 `cb83eeb`）已合并，合入的符号：`internal/core/settings` 的 `Load(dataDir) (*File, error)`、`Defaults()`、`File.Settings()`（深副本）、`File.CommandPrefixes(adapterInstanceID)`、`File.Update(edit func(*Document) error)`（Rename 为提交点，之后目录同步失败时仍发布快照并返回错误）、`File.Reload() (changed []string, err error)`、`Document.Set(path, value)`/`Delete(path)`，`Settings`（含 `Admins`、`CommandPrefixes`、`CommandPrefixesByInstance`、`DefaultLocale`、`LogSources`、`Adapters map[string]AdapterSettings{Version, Instances}`）与 `Admin`（T-04.10 起为 `{AdapterID, UserID}`，admins 为用户字符串）；`adminauth.FirstToken(ctx, st, dataDir, interactive)`、`FirstTokenFile` 与 `adminauth.New(store, now, baseURL, dataDir)`；`alonabot admin-token show`。T-08.8（PR #102，合并提交 `f685c92`）已合并，合入的符号：`contracts/i18n` 的 `Parse(files) (*Catalog, []Problem)`、`Catalog.Has`/`Format(locales, key, args)`、`Placeholders`；`sdk.Context.T`、后台 `sdk.T` 与 `Generated.I18n`；`internal/core/i18n.Messages()`；`host.Instance.Messages`/`DefaultLocale`；`router.Plugin` 的 `Messages()`/`DefaultLocale()`；`supervisor.Spec.I18n`/`DefaultLocale`；`Bootstrap.default_locale = 12`。goal-95（PR #103，合并提交 `c840af7`）已合并，合入的接口：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)` 与 `Schema` 的 `Canonical`、`Hash`、`Version`、`Migrate(config, from)`、`Complete`、`Validate`、`SecretPaths`、`CheckI18n(config)`；`sdk/config` 的 `Prepare(ctx, load, schema, raw, fromVersion, targetVersion, restart)`、`Loader`、只读的 `Values`（`String`/`Int`/`Float`/`Bool`/`Strings`/`Object`/`Objects`/`Dict`/`Template`/`TemplateList`）与 `Template.Key`；`sdk.Base.Config() config.Values` 与根插件的 `OnConfigLoad(ctx, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`；`supervisor.Spec.ConfigSchema` 为 config_schema.json 的规范 JSON。生成器读取手写 `config_schema.json`，配置标签解析与 `config.Config`/`config.Base` 已删除。T-11.1（PR #101，合并提交 `de48f26`）已合并，工具策略端点经 `tool.Catalog` 的 `Policies` 与 `SetPolicy`。2026-11-04 按前后端拆分新增，内容取自原 goal-44（T-09.9）语句中的管理 API；页面留在 goal-44。原语句的明文查看接口写成 GET 并在请求体中带 token，浏览器的 fetch 不允许 GET 带请求体，这里改为 POST。同日按 D-77～D-85 修订：Core 设置经 goal-97 的 settings 写回 `core_config.json`，Core Admin 列表同时经 goal-13 的 `policy.SetAdmins` 生效；发信许可改为「全部实例」或指定实例（goal-99）；删除插件页面票据（D-84，页面经 WebUI 的桥访问插件路由）；开发组件的配置写回 `config.json` 并重启组件；Adapter 实例的增删改写回 `core_config.json` 并按 goal-101 的 adapter_config_schema.json 校验。语句中 goal-13、goal-29、goal-31、goal-36、goal-39～goal-41、goal-53～goal-56、goal-84、goal-87 提供的符号（policy 的覆盖与 Admin 方法、config.Service、secret.Store 与 Redact、adminauth.Reauthenticate、管理服务与 reload、i18n 的默认 locale、tool.Catalog、conversation.Service、spec/ADMIN-API.md、router.New 的 prefixes）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-93（切片 T-09.14：配置、权限、工具策略与会话的管理 API）。goal 目录 goal-93，分支 goal-93，worktree .worktrees/goal-93，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在管理服务中补齐 Admin API：插件配置（raw JSON、Merge Patch、base_generation 与 secret 语义，Pending 经 reload 激活）、命令权限与启用、命令名覆盖、配置字段的选项查询、发信许可（allowed_adapter_instances）、Core 设置（Core Admin 列表、默认 locale、命令前缀、全局禁止表）、工具调用策略、Adapter 实例（按 adapter_config_schema.json 校验）、Agent 会话的列出与清除，以及 secret 明文查看的再认证与审计；Core 设置与 Adapter 实例写回 core_config.json。接口追加到 spec/ADMIN-API.md；页面由 WebUI 章节的 T-09.9 实现。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §5、§2.3、§8（allowed_adapter_instances）、§13、§7.3、§5.4；DECISIONS D-45、D-62、D-128、D-129、D-130；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56、V-115、V-116、V-117 的管理 API 分项。

范围（只新增或修改这些）：
- internal/core/manage/：新增 config_api.go、policy_api.go、settings_api.go、tool_api.go、conversation_api.go；在管理服务的 HTTP 路由中追加注册
- spec/ADMIN-API.md：追加本切片的接口（本切片获授权修改该文件）
- 上述包的测试文件

接口（HTTP，全部在 /api/v1 下、经 adminauth.Require；JSON；错误格式与 goal-39 一致）：
- GET /plugins/{id}/config：{active_generation, active: Redact 后的快照, pending: {base_generation, config: Redact 后}|null, schema, i18n}；schema 是当前安装包中 config_schema.json 的规范内容（D-76，没有时为 null），i18n 是包内 i18n/*.json 的 config 根节点，按语言标签给出，供表单显示文案
- PUT /plugins/{id}/config/pending：{base_generation, config}（完整替换）；PATCH /plugins/{id}/config/pending：{base_generation, patch}（RFC 7396）；DELETE 丢弃 Pending；POST /plugins/{id}/config/import 导入 config.json；secret 字段的新值在写入 Pending 前经 secret.Extract 存为引用，省略或未修改的 secret 保持原引用；base 不是当前 Active 时 409 并返回 FAILED_PRECONDITION。
- POST /plugins/{id}/reload：经 goal-40/41 激活 Pending。
- GET/PUT /plugins/{id}/policies：命令覆盖（作用域、permission、enabled），以及 Router 的遮蔽诊断。GET 的每条命令另带声明的命令名、生效的命令名、是否有覆盖与失效的覆盖项（D-128）。
- PUT /plugins/{id}/commands/{route_id}/names：{names: [[token, ...], ...]}，整份替换该命令的命令名覆盖，第一项为主名；DELETE 删除覆盖，恢复插件声明的命令名。空列表与格式错误返回 INVALID_ARGUMENT，同一插件内重名返回 409 与 ALREADY_EXISTS（safe_message 带冲突的 route_id），文本匹配与 Raw 路由返回 INVALID_ARGUMENT。追加与删除单个命令名由调用方在列表上完成后整份提交。
- GET/PUT /plugins/{id}/allowed-instances：{all: bool, instances: [...]}，经 goal-99 的 AdapterInstanceAccess 与 SetAdapterInstanceAccess；缺省 all 为 true。
- GET/PUT/DELETE /adapters/{adapter_id}/instances/{instance_id}：实例配置的读取（secret 只返回是否已设置）与写回 core_config.json，声明由 configschema.Parse 读取 adapter_config_schema.json；提交配置先经 Schema.Complete 与 Validate 检查，再经 File.Update 写回并保留 ${...} 原写法；启动前复用 File.PrepareInstances，problem 的 Path 与 Reason 原样用于错误定位；GET /adapters/{adapter_id}/instance-schema 供页面生成表单。
- GET/PUT/DELETE /core/admins、/core/denies，GET/PUT /core/default-locale，GET/PUT /core/command-prefixes（全局与按 Adapter 实例的覆盖）。
- GET/PUT /tools/policies：工具策略（缺省与覆盖）。
- GET /plugins/{id}/conversations、DELETE /plugins/{id}/conversations/{key}。
- POST /secrets/{plugin}/{path}/reveal：请求体带当前管理 token 再认证（adminauth.Reauthenticate），通过后返回明文并以 INFO 审计记录（不含明文）。

行为：
- 配置：invisible 字段在读取中照常返回，是否展示由页面决定；PATCH 不会因为未提交 invisible 字段而删除它。
- 命令权限：有效权限与启用状态按 policy.CheckCommand 的同一规则计算后返回，覆盖分插件、命令组、路由三层；同时返回遮蔽关系。
- 命令名：保存与删除经 goal-129 的 policy 方法，校验与重名规则不在本切片重复实现；保存后下一次派发生效，遮蔽关系按生效的命令名返回。
- GET /plugins/{id}/config/options?path=…&q=…：按字段路径列出选项（D-129），返回 {options: [{value, label, description}], more}；POST /plugins/{id}/config/options/resolve：{path, values: [...]}，返回这些存储值当前的选项，不存在的值不返回。来源名与 filter 由 goal-130 的查询入口从该插件 Active 代次的 schema 取得，请求不能指定来源；locale 取管理会话的界面语言。字段不存在或没有 `options_from` 返回 INVALID_ARGUMENT，来源不可用返回 503 与 UNAVAILABLE。Adapter 实例配置提供同样的两个接口，挂在本切片已有的 Adapter 实例配置路径之下。
- GET /core/options/{source}?q=…&filter[键]=…：直接查询内置来源（`core.chats`、`core.adapter_instances`），供 Core 设置页选择全局禁止表中的会话；服务来源不能经此接口查询，返回 INVALID_ARGUMENT。POST /core/options/{source}/resolve 按存储值取回。
- 选项查询的本地自测（不提交，PR 中写明结果）：插件字段的列出与取回、请求中多出的来源参数被忽略、没有 `options_from` 的字段、Agent 未启用时 `@agent/llm.providers` 返回 UNAVAILABLE、`core.chats` 按名称搜索与按 `chat_type` 筛选、未认证被拒。
- Core 设置：Core Admin 列表、默认 locale、command_prefixes 与按实例的覆盖都经 goal-97 的 settings.Update 写回 core_config.json，保留原有的 ${...} 写法；写回成功后 Core Admin 列表调用 policy.SetAdmins，命令前缀在下一次派发生效，默认 locale 在之后建立的 Session 生效。alonabot admin add|remove|list 使用同一服务。全局禁止表仍在状态库。
- 开发组件（dev 来源）：PUT/PATCH 配置时校验通过后写回 <data_dir>/plugins/<id>/config.json 并重启该组件，不建 Pending（CONTRACTS §15.4）。
- 会话：只返回摘要（键、条目数、最后时间），不返回条目内容。
- 所有修改请求经 CSRF 检查；修改后的策略在下一次派发时生效，不需要重启插件。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：各端点的错误格式与 409；读取结果中没有 secret 明文；internal/core/supervisor/admin_api_local_test.go（//go:build unix）：经管理 API 完成：PATCH 配置（未修改的 secret 保持原引用、新 secret 生成新引用、invisible 不被删除）、陈旧 base 409、reload 后 Active 改变；设置命令覆盖后下一条消息按新权限处理；保存命令名覆盖后旧名不再触发、新名触发，空列表与同插件重名被拒，DELETE 后恢复声明的命令名；把发信许可收紧为其他实例后插件后台 Send 被拒、恢复为全部实例后成功；Core 设置写回后 core_config.json 中的 ${...} 原写法保留；设置默认 locale 后新 Session 使用；工具策略改为 deny 后 ListTools 不含它；清除会话；开发组件的配置写回 config.json 并重启；reveal 需要再认证。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：页面（T-09.9）；日志 API（T-09.10）；故障注入与验收汇总（T-09.11）。
```

## 第 19 波

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
