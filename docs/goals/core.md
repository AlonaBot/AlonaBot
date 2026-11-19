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
| goal-121 | T-03.24 | 插件给出的目标由 Core 补齐 | T-03.6、T-07.6 | 已完成（PR #135） |
| goal-96 | T-04.9 | 删除构建期标记 | T-04.7 | 已完成（PR #110） |
| goal-122 | T-04.10 | 管理员与封禁按 Adapter 登记 | T-04.4、T-07.6 | 已完成（PR #133） |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 已完成（PR #67） |
| goal-98 | T-07.7 | 插件页面的沙箱隔离 | T-07.4、T-07.6、T-09.1 | 已完成（PR #116） |
| goal-33 | T-08.6 | SQLite 候选与资源交接 | T-08.4、T-08.5 | 已完成（PR #85） |
| goal-95 | T-08.11 | 配置改由 config_schema.json 声明 | T-04.7、T-08.3、T-08.4、T-08.10 | 已完成（PR #103） |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-04.7、T-06.1 | 已完成（PR #101） |
| goal-91 | T-04.8 | SDK 自有类型与事件常量 | T-03.9、T-03.24、T-04.7、T-04.9 | 已完成（PR #141） |
| goal-24 | T-07.2 | schedule 注册与运行槽 | T-05.1、T-07.1 | 已完成（PR #107） |
| goal-35 | T-08.9 | configdb 示例与 T-08 验收 | T-08.6、T-08.7、T-08.8、T-08.11 | 已完成（PR #117） |
| goal-101 | T-08.12 | Adapter 实例配置与 Core 设置的版本迁移 | T-07.6、T-08.11 | 已完成（PR #114） |
| goal-123 | T-08.13 | 文案函数只返回字符串 | T-04.9、T-08.8 | 已完成（PR #137） |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2、T-04.9、T-08.8、T-09.16 | 已完成（PR #149） |
| goal-104 | T-03.13 | 媒体按来源传递 | T-03.6、T-04.8、T-07.7 | 已完成（PR #148） |
| goal-105 | T-03.14 | 新增片段与字段（含会话类型） | T-03.8、T-04.8 | 已完成（PR #143） |
| goal-106 | T-03.15 | 平台动作 | T-03.8、T-04.8 | 已完成（PR #146） |
| goal-107 | T-03.16 | 撤回 | T-03.9、T-04.8 | 已完成（PR #144） |
| goal-108 | T-03.17 | 命令菜单同步 | T-04.6、T-04.7、T-04.8、T-08.8 | 已完成（PR #150） |
| goal-109 | T-03.18 | 流式输出辅助函数 | T-03.9、T-04.8 | 已完成（PR #151） |
| goal-110 | T-03.19 | Adapter 的 HTTP 路由 | T-03.3、T-04.8、T-07.7 | 已完成（PR #152） |
| goal-125 | T-04.12 | 插件对象、结构体参数与子插件 | T-04.8 | 已完成（PR #154） |
| goal-16 | T-05.2 | Rule、HandlerOutcome 与传播 | T-05.1、T-04.7、T-04.8 | 已完成（PR #145） |
| goal-27 | T-07.5 | 调度与 HTTP 启停集成及验收 | T-07.2、T-07.4 | 已完成（PR #119） |
| goal-39 | T-09.4 | 管理服务、安装记录与 run | T-03.8、T-07.6、T-08.4、T-08.8、T-09.1、T-09.2、T-13.1 | 已完成（PR #163） |
| goal-68 | T-13.2 | 开发组件的配置与数据库 | T-07.6、T-08.5、T-08.11、T-13.1 | 已完成（PR #160） |
| goal-124 | T-04.11 | 注册写法统一 | T-04.8、T-04.12、T-07.5 | 已完成（PR #159） |
| goal-126 | T-04.13 | Service 与 Event Bus 的带类型声明 | T-04.12 | 已完成（PR #158） |
| goal-129 | T-04.14 | 命令名覆盖的 Core 侧 | T-03.17、T-04.10 | 已完成（PR #156） |
| goal-17 | T-05.3 | 三阶段 Hook 与 send_forbidden | T-05.2 | 已完成（PR #155） |
| goal-52 | T-05.4 | 延续调用与多轮等待 | T-05.2 | 已完成（PR #153） |
| goal-127 | T-07.8 | 动态任务与调度持久化的协议和 Core 侧 | T-07.5、T-09.4 | 已完成（PR #169） |
| goal-130 | T-08.14 | 配置项的选项来源 | T-04.12、T-06.1、T-08.11 | 已完成（PR #157） |
| goal-157 | T-08.17 | 插件的 KV 与数据目录 | T-08.5、T-09.4 | 已完成（PR #212） |
| goal-84 | T-09.12 | WebUI 组件与组件图标 | T-07.7、T-09.1、T-09.4、T-13.1 | 已完成（PR #167） |
| goal-69 | T-13.3 | 开发运行与热部署 | T-07.6、T-09.16、T-13.1、T-09.4 | 已完成（PR #171） |
| goal-81 | T-16.1 | 插件测试套件基础 | T-05.2、T-13.2、T-09.4 | 已完成（PR #168） |
| goal-9 | T-03.10 | Telegram 接入开发入口与故障矩阵 | T-03.6、T-03.8、T-03.9、T-07.6、T-13.3 | 已完成（PR #180） |
| goal-131 | T-03.25 | 聊天目录 | T-03.3、T-08.14、T-09.4 | 已完成（PR #166） |
| goal-18 | T-05.5 | 多进程事件流水线验收 | T-03.8、T-05.3、T-05.4 | 已完成（PR #162） |
| goal-20 | T-06.2 | 父调用链、取消与背压 | T-03.6、T-05.3、T-06.1 | 已完成（PR #161） |
| goal-128 | T-07.9 | 动态任务与暂停恢复的 SDK | T-04.12、T-07.8 | 已完成（PR #176） |
| goal-151 | T-08.15 | 配置字段的回复示例 | T-08.11、T-08.14、T-09.12 | 已完成（PR #211） |
| goal-66 | T-09.10 | 日志管理 | T-02.1、T-03.1、T-07.6、T-09.4、T-09.12 | 已完成（PR #177） |
| goal-92 | T-09.13 | 管理页所需的会话与上传接口 | T-09.1、T-09.4、T-09.12 | 已完成（PR #181） |
| goal-149 | T-02.5 | CRITICAL 日志级别 | T-02.1、T-09.10 | 已完成（PR #209） |
| goal-14 | T-04.5 | 示例升级与 T-04 验收 | T-03.10、T-04.2、T-04.3、T-04.4、T-04.7、T-04.8、T-04.9、T-04.11、T-04.12、T-04.13 | 已完成（PR #191） |
| goal-22 | T-06.4 | Session 清理与 Service 端到端 | T-06.2、T-06.3 | 已完成（PR #164） |
| goal-55 | T-11.2 | 工具调用与用户确认 | T-03.8、T-06.2、T-08.8、T-11.1 | 已完成（PR #165） |
| goal-70 | T-13.4 | 开发运行示例与验收 | T-03.10、T-13.2、T-13.3 | 已完成（PR #186） |
| goal-82 | T-16.2 | 按钮、多轮等待、schedule 与示例测试 | T-03.8、T-05.4、T-07.2、T-07.9、T-16.1 | 已完成（PR #183） |
| goal-38 | T-09.3 | 开发 CLI 与库模式脚手架 | T-04.5、T-08.8、T-08.11、T-09.16、T-13.3 | 已完成（PR #194） |
| goal-40 | T-09.5 | 候选 Journal 与准备编排 | T-09.4、T-08.6、T-06.4 | 已完成（PR #173） |
| goal-57 | T-11.4 | 工具与会话的验收 | T-11.2、T-11.3 | 已完成（PR #172） |
| goal-102 | T-16.3 | plugintest 的 LLM 替身 | T-12.8、T-16.1 | 已完成（PR #190） |
| goal-138 | T-04.15 | SDK 公开面收敛 | T-04.5、T-09.3、T-16.1 | 已完成（PR #197） |
| goal-41 | T-09.6 | 唯一提交点与激活门禁 | T-07.5、T-09.5 | 已完成（PR #179） |
| goal-142 | T-09.17 | 包校验收紧 | T-09.3、T-09.4 | 已完成（PR #201） |
| goal-139 | T-04.16 | SDK 按类别分包 | T-04.15 | 已完成（PR #199） |
| goal-42 | T-09.7 | 崩溃恢复、回退与回收 | T-09.6 | 已完成（PR #187） |
| goal-93 | T-09.14 | 配置、权限、工具策略与会话的管理 API | T-03.25、T-04.14、T-07.6、T-08.8、T-08.11、T-08.12、T-08.14、T-09.6、T-09.12、T-11.1、T-11.3 | 已完成（PR #192） |
| goal-168 | S-01.1 | 可休眠声明与持久化运行状态 | T-04.1、T-09.4、T-09.6 | 已完成（PR #225） |
| goal-169 | S-01.2 | 已提交代次的路由快照与索引 | T-04.1、T-09.4、T-09.6 | 已完成（PR #226） |
| goal-150 | T-04.17 | 插件页面套件 | T-04.16、T-07.7 | 已完成（PR #218） |
| goal-152 | T-08.16 | 去掉配置版本与迁移，配置宽松载入 | T-08.11、T-08.12、T-09.14 | 已完成（PR #239） |
| goal-160 | T-08.18 | 配置文件的存储与管理 | T-08.14、T-09.14 | 已完成（PR #214） |
| goal-147 | T-09.19 | 诊断包 | T-09.10、T-09.14 | 已完成（PR #213） |
| goal-153 | T-09.20 | WebUI 组件的后台进程与 command.invoked | T-09.12、T-09.14、T-04.16 | 已完成（PR #221） |
| goal-154 | T-09.21 | 插件数据浏览 | T-09.14、T-08.5、T-08.17 | 已完成（PR #234） |
| goal-155 | T-09.22 | WebUI 需要的补充管理 API | T-09.14、T-07.9、T-03.25、T-04.16 | 已完成（PR #220） |
| goal-158 | T-09.24 | 组件的 changelog | T-09.3、T-09.14 | 已完成（PR #215） |
| goal-148 | T-11.5 | 工具的副作用类别 | T-11.2、T-12.3、T-09.14 | 已完成（PR #219） |
| goal-140 | T-13.5 | 模块根包只保留库模式接口 | T-04.16 | 已完成（PR #200） |
| goal-170 | S-02.1 | 按需唤醒、活动引用与有界等待 | S-01.1、S-01.2、T-05.5、T-06.4、T-07.5 | 已完成（PR #241） |
| goal-176 | S-02.3 | Core 消息事件过滤与 SDK 注册 | T-04.16、T-08.11、T-09.14 | 已完成（PR #227） |
| goal-161 | T-04.18 | 配置文件的派发与 SDK | T-04.16、T-08.18 | 已完成（PR #242） |
| goal-171 | S-02.2 | 空闲退出与即时休眠策略 | S-02.1 | 已完成（PR #249） |
| goal-177 | T-09.27 | 组件市场 ID 改用连字符 | T-09.10、T-09.14、T-12.8、T-13.5 | 已完成（PR #222） |
| goal-162 | T-04.19 | 平台与消息 | T-04.18 | 已完成（PR #250） |
| goal-178 | T-04.20 | SDK 用法从简 | T-04.18、T-16.2 | 已完成（PR #251） |
| goal-45 | T-09.11 | 正式流程故障注入与 T-09 验收 | T-09.3、T-09.7、T-09.9、T-09.10、T-09.12、T-09.15 | 进行中 |
| goal-163 | T-11.6 | 按配置文件的工具策略 | T-04.18、T-11.4 | 已完成（PR #246） |
| goal-141 | T-13.6 | CLI 参数从简 | T-13.5、T-14.3 | 已完成（PR #233） |
| goal-179 | T-13.7 | CLI 开箱即用 | T-13.6、T-16.1 | 已完成（PR #253） |
| goal-172 | S-03.1 | Scheduler 唤醒与离线 tick 区分 | S-02.2、T-07.2 | 已完成（PR #256） |
| goal-173 | S-06.1 | 休眠管理 API、运行诊断与 CLI | S-02.2、S-03.1、T-09.14 | 已完成（PR #269） |
| goal-174 | S-06.2 | 休眠指标、故障注入与固定机器验收 | S-06.1 | 待派发 |

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
- manifest：按 CONTRACTS §9.1 校验各角色：role 为 adapter 时必须有 adapter_id，且不能有 adapter_dependencies、plugin_dependencies、migration_mode 或 prepare_effects；agent 与 plugin 的字段规则相同；webui 必须有 assets 与 Admin API 主版本范围、os/arch 为 any，不能有 entrypoint、协议主版本、required_capabilities、config_version 与依赖；市场 ID 的前缀与 role 一致（alonabot-plugin-、alonabot-adapter-、alonabot-agent-、alonabot-webui-）；依赖项的 ID 格式、版本 SemVer、重复项都要校验；必需的 plugin_dependencies 不能直接依赖自己。环的检测需要全部安装记录，由 T-09.4 完成。
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

### goal-67 · T-13.1 库模式入口与同进程组件

状态：已完成（[PR #149](https://github.com/AlonaBot/AlonaBot/pull/149)，合并提交 `bc39361`），证据见 TEST-MATRIX §3.2。合入：根包 `alonabot.Run(opts ...Option) error`、`alonabot.Start(ctx, opts...) (*Core, error)`、`(*Core).Wait() error`，选项 `DataDir(dir)`、`Plugin(root sdk.Root)`、`Adapter(a adapter.Adapter)`；`assembly.go` 的 Core 装配（固定服务 send、edit、delete、action、媒体上传与读取、会话、工具目录、Event Bus、Broker 转交，Router、按钮 `buttons.New(store, now, ttl, current)`、调度与插件 HTTP 代理绑定）；`internal/inproc` 的同进程组件启动参数；SDK 的 `currentSession(ctx)` 按 ctx 绑定 Session；开发组件 source=dev 与固定代次 `dev:<组件 ID>`；已安装组件按重启策略单独恢复；`examples/devbot`。库模式尚无公开 Listener，`core.media.url` 与 `/media/` 的装配留给后继装配切片。派发时：前置 T-02.1、T-03.3、T-04.2、T-04.9（PR #110）、T-08.8（PR #102）、T-09.16（PR #140）均已合并；T-03.8（PR #136）与 T-08.13（PR #137）也已合并。2026-11-12 按维护者决定 D-134 补齐 Adapter 工程入口、生成登记与固定代次；审查者按 main `896991b` 核对现有符号、文件与签名。D-134（[PR #142](https://github.com/AlonaBot/AlonaBot/pull/142)，合并提交 `1e71494`）已合并。宜在 T-04.8（goal-91，[PR #141](https://github.com/AlonaBot/AlonaBot/pull/141)）合并后派出：两者都改 `sdk/sdk.go`、`adapter_run.go`、`session.go` 等文件，派发前审查者按合入的 T-04.8 代码核对语句中的 sdk 符号。

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
- Telegram 的工程 manifest 放在 adapters/telegram，adapter_id 为现有 AdapterID（telegram），组件 ID 使用 alonabot-adapter-telegram；schema 移动时保留现有字段、版本与内容。cmd 只调用 sdk.RunAdapter，库模式从同一个 telegram.New() 的具体类型取得包内登记。build/pack 从可导入包取资源并构建 cmd 薄入口的接入由 T-09.3、T-14.3 完成（D-134）。
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

## 第 11 波

### goal-104 · T-03.13 媒体按来源传递

状态：已完成（[PR #148](https://github.com/AlonaBot/AlonaBot/pull/148)，合并提交 `cf74fa9`），证据见 TEST-MATRIX §3.2。合入：MediaSource 的 `inline_base64` 与 `platform_file`；同实例入站对象复用平台文件标识；`core.media.url`（`media.Store.AddURLService(mux, baseURL)`）与公开 Listener 上的 `/media/<url_token>`（`webhost.Options.Media`，GET/HEAD/Range，sandbox CSP 与 nosniff）；SDK 的 `ImageBytes`、`VideoBytes`、`FileBytes` 与 Adapter 侧 `MediaURL`。T-13 的正式装配需对同一 Mux 调用 `AddURLService`，并给 webhost 注入 `Media`。派发时：前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.6（PR #55）、T-07.7（PR #116，合并提交 `e65eb4d`）提供入站媒体与统一入口。2026-11-09 按 D-87、D-88 新增。当前 `media.Store.LeaseForSend` 返回只读 token，`message.New` 通过函数注入取得它；本切片把该返回值扩为准备好的媒体来源。`settings.Settings.BaseURL` 与 `webhost.Options.BaseURL` 已存在；`cmd/alonabot/main.go` 当前只装配认证、插件代理与调度，完整媒体装配由 T-13 接入下文的 Store 与 Host 接口。回环 Adapter 已在 `internal/core/adapter/remote/remote_test.go` 的 `TestMain` 中，无独立的 `examples/adapters/loopback/` 工程。

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

状态：已完成（[PR #143](https://github.com/AlonaBot/AlonaBot/pull/143)，合并提交 `15682e8`），证据见 TEST-MATRIX §3.2。合入：voice 片段、转发引用节点与入站转发标识、URL 按钮、`Event.chat_type` 与 SDK 的会话类型判断；Core 在按钮签发与媒体取件前递归核对 message.voice、message.forward 能力，嵌套 forward 仍为 UNSUPPORTED，入站非法 chat_type 在游标持久化前拒绝；URL 按钮清除外来 token；voice 进入现有来源登记、Session token 与消费租约路径。派发时：前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.8（PR #136，合并提交 `3621fc1`）已提供 `buttons.Service.Issue`、`adapter.ExtensionChecker` 与带类型的按钮动作。SDK 的 `message.Button` 是 `sdk/internal/messagecodec.Button` 的别名，新增 URL 字段须改其定义；SDK 的收发转换在 `sdk/internal/convert`。2026-11-09 按 D-87、D-89、D-91、D-97 新增，同日按 D-100 追加会话类型；字段编号按并行切片规则在实现时追加。

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

状态：已完成（[PR #146](https://github.com/AlonaBot/AlonaBot/pull/146)，合并提交 `ab5f8fb`），证据见 TEST-MATRIX §3.2。合入：`core.action` 与 `adapter.action`（`internal/core/message.Service.Action`：按 Adapter 声明的动作名、精确版本与 chat/account 范围准入，参数与结果 Schema 校验，`adapter.ActionExecutor` 可选接口，`Registry.CompleteAccount`，policy 的 `CheckInstance`）；协议 `ActionScope`、`ActionDeclaration`（`AdapterDescriptor.actions = 6`）、`ActionRequest`、`ActionResponse`、`AdapterActionRequest`；SDK 的 `adapter.Action[P,R]`、`ActionScope`、`ActionExecutor`、`ActionCaller`、`Context.CallAction` 与 `ActionClient`/`NewActionClient`；`AdapterDependency.Actions`。派发时：前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.8（PR #136，合并提交 `3621fc1`）提供 extension 校验与 `manifest.CheckDependencies(m *Manifest, in Installed) []Problem`；当前安装快照的 `Installed.Adapters` 是 `map[string][]ExtensionRequirement`。SDK 已有 `adapter.Event[T]`、`adapter.Extension[T]` 的类型推导与内部 convert.Descriptor；平台动作声明须同时扩展两个 Descriptor。`sdk.Action[T]` 是按钮动作，平台动作使用 `sdk/adapter.Action[P,R]`。2026-11-09 按 D-87、D-92 新增，同日按 D-113 追加类型推导。

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

状态：已完成（[PR #144](https://github.com/AlonaBot/AlonaBot/pull/144)，合并提交 `707b74c`），证据见 TEST-MATRIX §3.2。合入：`core.delete` 与 `adapter.delete`（`internal/core/message.Service.Delete`，Core 补齐目标并按 Send 的准入检查，实例须声明 `message.delete` 并实现可选的 `adapter.Deleter`，按输入顺序返回逐项结果与顶层安全错误）；SDK 的 `c.Delete(refs ...message.MessageRef)`、`sdk.Delete(ctx, refs...)`、`sdk/message.DeleteResult`/`DeleteItemResult` 与 `sdk/adapter.Deleter`；协议 `DeleteRequest`、`DeleteResult`、`DeleteItemResult`、`AdapterDeleteRequest`。派发时：前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-03.9（PR #118，合并提交 `1cbfb2e`）提供 Core 与 SDK 两侧可选的 `adapter.Editor`；SDK 的 `c.Edit`、`sdk.Edit` 已收 `message.MessageRef`，句柄拆分复用 `MessageRef.Parts` 与内部 `convert.RefToProto`。Core 新增撤回仍须先调用已合入的 `Registry.Complete` 补齐省略的实例与账号，再复用准入。2026-11-09 按 D-87、D-93 新增，同日按 D-104 改为只收句柄。

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

状态：已完成（[PR #150](https://github.com/AlonaBot/AlonaBot/pull/150)，合并提交 `8eb7846`），证据见 TEST-MATRIX §3.2。合入：`CommandSpec.description`、`CommandMenuItem`、`AdapterCommandsSyncRequest/Response` 与 `wire.ServiceAdapterCommandsSync`；`contracts/i18n` 的 commands 覆盖（`Locales`、`CheckCommands`、`CommandDescription`）；SDK 的 `(*CommandBuilder).Description(text)` 与 Collect 的实际注册表检查；`sdk/adapter.CommandSynchronizer`；Router 的 `StartCommandSync`/`StopCommandSync`/`CommandsChanged` 与 `router.Options.CommandState`；`Policy.CommandState` 与 `SetCommandsChanged`；`supervisor.Spec.CommandsChanged`；`remote.Options.StartCommandSync`/`StopCommandSync`；库模式装配已接入。派发时：2026-11-12 依据维护者决定 D-138（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`），把说明字段、SDK builder、Core 翻译、生成器结构检查与 Collect 覆盖检查纳入本切片；前置均已合并：T-04.6（PR #78）、T-04.7（PR #89）、T-04.8（PR #141，合并提交 `ac4aded`）、T-08.8（PR #102）。按 main `c04b2ca` 的代码核对。D-128 的命令名覆盖仍由后继 goal-129 接入；与 goal-125 同改 SDK 注册与生成器，后合并者同步 main 保留 Description 及覆盖检查，迁到插件对象上的同一 builder。

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

状态：已完成（[PR #151](https://github.com/AlonaBot/AlonaBot/pull/151)，合并提交 `ba14fd4`），证据见 TEST-MATRIX §3.2。合入：`Capability.text_limit`（`TextLimit{max, unit}`，三种计量单位）与注册校验，Telegram 的 `message.send` 声明 4096 UTF-16；SDK 的 `message.TextLimit` 与 `func (c *Context) Stream(next func(context.Context) (string, error)) ([]*message.SendResult, error)`；Core 的 `Edit` 对显式目标经 `Registry.Complete` 补齐实例与账号。派发时：2026-11-12 依据维护者决定 D-135（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`） 补齐能力元数据、SDK 接口与测试；前置均已合并：T-03.9（PR #118，合并提交 `1cbfb2e`）、T-04.8（PR #141）。按 main `c04b2ca` 的代码核对。与 T-03.13～T-03.19 的其他切片并行时按执行顺序 §3 追加 proto；与 goal-125 同改 SDK 的 Context/装配和 remote 测试 helper，后合并者同步 main 保留对方接口。范围修订（2026-11-12，审查者补充）：执行者发现 main 现有缺陷，Core 的 Edit 对显式目标未经 `Registry.Complete` 补齐账号，而 SDK 由 MessageRef 转出的目标不带 account_id，显式 Edit 一律被拒；授权修改 `internal/core/message/edit.go`，按 Send 的同一写法补齐显式目标，准入拒绝时只回传调用方已知的目标，由本切片的主路径集成测试证明。

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

状态：已完成（[PR #152](https://github.com/AlonaBot/AlonaBot/pull/152)，合并提交 `2f4284d`），证据见 TEST-MATRIX §3.2。合入：`HTTP_AUTH_MODE_PLATFORM_TOKEN` 与 `sdk.AuthPlatformToken`；`sdk.AdapterHTTP`（合入 main 时因 T-04.12 删除了导出的注册器类型，改为 `HTTP(handle func(method, path string, h http.Handler, opts ...HTTPOption))`，与插件对象的 `Handle` 同签名与同校验；T-04.11 会再迁到 HTTP builder）；`httpproxy.Server.Bind(pluginID, generationID, endpoint, token, routes, isAdapter)`、`Unbind(pluginID, generationID string, token []byte)` 与 WebSocket Upgrade 的转发与回收；`validate.AdapterRegister` 接受 HTTP 路由与端点；库模式装配按安装记录的角色调用 Bind。派发时：2026-11-12 依据维护者决定 D-136（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`） 统一 Adapter 角色规则与平台凭证转交；前置均已合并：T-03.3（PR #32）、T-04.8（PR #141）、T-07.7（PR #116，合并提交 `e65eb4d`）。生命周期接入 T-07.5（PR #119，合并提交 `cc19e99`）也已合并。按 main `c04b2ca` 的代码核对；T-03.13～T-03.19 的 proto 追加按执行顺序 §3，SDK HTTP 注册器后续由 T-04.11 改为 builder，本切片沿用现有 Handle/Public。若 goal-67 先合并，沿其同进程入口复用同一 Adapter HTTP 生命周期。范围修订（2026-11-12，审查者补充）：执行者报告 `sdk/registry.go` 的 HTTPRegistry.Handle 拒绝 `Public(AuthPlatformToken)`，语句漏列该文件；授权该校验接受公开平台 token 模式，插件注册路径（sdk.Run）与 Core 的 `validate.Register` 拒绝此模式，`RunAdapter` 与 `validate.AdapterRegister` 接受，其他 HTTP 校验与错误文本不变。

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

状态：已完成（[PR #154](https://github.com/AlonaBot/AlonaBot/pull/154)，合并提交 `43a6135`），证据见 TEST-MATRIX §3.2。合入：`sdk.NewPlugin() *Plugin` 与 `func New() *sdk.Plugin` 入口，`(*Plugin)` 的路由 builder、`OnStart`/`OnReady`/`OnStop`/`OnConfigLoad`、`Config()`、`Sub(id) *SubPlugin`、`Generated()`；`sdk.Registrar`（`*Plugin` 与 `*SubPlugin` 实现）；`sdk.UseGenerated(p *Plugin, g Generated)` 与生成的包内变量 `alonabotGenerated`；`sqlitedb`/`gormdb` 的 `Migrate(p, fn)` 与 `Migrations(p, list)`；结构体参数（`arg`、`default` 标签）；生成器的 `Result.PluginConstructor`；根包 `alonabot.Plugin(p *sdk.Plugin)` 与 `sdk/inproc.go` 的同进程入口；`sdk.Base`、`sdk.Root` 与旧注册器已删除。派发时：2026-11-12 依据维护者决定 D-137（规格 [PR #147](https://github.com/AlonaBot/AlonaBot/pull/147)，合并提交 `c04b2ca`） 补齐生成资源与插件对象的显式绑定，承接 D-114～D-116；前置均已合并（T-04.8 为 PR #141），按 main `c04b2ca` 的代码核对。goal-67（T-13.1）正在按旧的类型登记实现：goal-125 合并后同步库模式的资源读取；若 goal-67 先合并，本切片范围已包含其插件选项、SDK 同进程入口与示例的同步迁移，Adapter 资源规则沿用 goal-67。与 goal-101、goal-123、goal-124 共改生成器时按执行顺序 §3；与 goal-16、goal-68 并行时，后合并者同步 main 保留对方的注册与装配语义。范围修订（2026-11-12，审查者补充）：执行者报告 `sdk/adapter_run.go` 嵌入 `Base` 并充当 activation 的 root，语句漏列该文件；授权只迁移 adapterRuntime 到新插件对象模型下的最小内部结构（Adapter 行为与既有测试断言不变），activation 可在 `sdk/activate.go` 内用未导出接口区分插件与 Adapter，`sdk/config/config.go` 与 sdk 包文档中的 `sdk.Base` 说明同步改写。

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

状态：已完成（[PR #145](https://github.com/AlonaBot/AlonaBot/pull/145)，合并提交 `0f860a9`），证据见 TEST-MATRIX §3.2。合入：`sdk.Rule`、`sdk.Any`、`sdk.ToMe` 与命令、事件、按钮 builder 的 `Rule` 方法（Collect 校验 nil 与 Handle 顺序，REGISTER 写 `RuleDescriptor.HasCustomRule`）；Rule 与目标 Handler 在同一 RPC、同一 Context 中依次执行，Rule false/error/panic 为 matched=false；`StopPropagation` 在有效调用内锁存，调用结束时固定传播结果、使 Context 失效并取消派生 context；Core 自动回复要求调用成功、matched=true、自动模式且内容非空，排在显式发送之后。派发时：前置均已合并（T-04.8 为 PR #141（合并提交 `ac4aded`））；审查者按 main 合入 PR #141 后的代码核对。T-05.1（PR #67）、T-04.7（PR #89）已合并。2026-11-12 按当前 builder、SDK 自有类型、Permission 与 HandlerOutcome 改写，自动型显式 Send 沿用 D-70 和 MESSAGE-MODEL §3；PLAN T-05 步骤 3 中旧的自动型 Send 限制按 T-04.7 已完成行、同章 V-14 与现行消息契约执行。与 goal-125、goal-68 并行时，后合并者同步 main，在已有注册与装配上保留本切片的 Rule 和调用生命周期语义。

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

### goal-39 · T-09.4 管理服务、安装记录与 run

状态：已完成（[PR #163](https://github.com/AlonaBot/AlonaBot/pull/163)，合并提交 `cb8ae3d`），证据见 TEST-MATRIX §3.2。合入：根包 `Listen`、`LogLevel`、`LogFormat` 选项，`Plugin` 接受 role plugin 或 agent，Start 只加载一次 `settings.File` 并交给 `assemble(store, dir, values, file)`，HTTP 代理与管理认证在 Start 中创建；`internal/core/manage` 的 `New(Options)`、`Install`/`List`/`Enable`/`Disable`/`Restart`/`Uninstall(PurgeScope)`/`Pin`/`Reload`/`Update`/`Rollback`、实例与管理员方法、`Handler(auth)`；state 的 install 新列与 `RoleAgent`、`RoleWebUI`；`release.pub` 与签名可选的预检；`core_config.json` 的 `adapters.<id>.package` 归属；cmd/alonabot 的 run 调用根包，管理命令经管理 API。D-139 待维护者确认。派发时：前置均已合并。2026-11-12 审查者按 main `ce89453` 改写，要点如下（之前各轮的核对记录见本切片的 git 历史）：

- 根包 `alonabot.Start` 已按安装记录启动已启用的包（`startInstalled`），并以同进程方式运行登记的开发组件；但它没有读取 `core_config.json`、没有 HTTP Listener、webhost 与管理认证，Router 的 Prefixes 为 nil，也不启动 Adapter 实例。
- `cmd/alonabot` 的 `run` 仍是 T-02 时期的独立运行器（`main.go` 的 `run` 与 `launch`）：自己读取设置、开日志、交付首次管理凭证、建立 webhost 与 HTTP Listener，再用 `--plugin-id`、`--entry`、`--stdio` 运行一个插件可执行文件，没有 Router 与 Adapter。`cmd/alonabot/e2e_test.go` 的 15 个测试都经它运行 `examples/plugins/lifecycle`。
- 本切片把 `run` 改为调用根包，并把上面那些启动步骤移进根包，库模式与 `alonabot run` 从此共用一套装配（PLAN T-09 步骤 3「复用 T-13 的 Core 装配」）。`--plugin-id`、`--entry` 与 `--stdio` 随旧运行器删除；以目录运行开发组件的 `alonabot run [目录...]`、`--reload` 与 `--stdio` 由 goal-69 加入，goal-69 在本切片合并后派发。
- 已有的接口：`state.Store` 的 `AddInstall`、`GetInstall`、`Installs`、`SetEnabled`、`Enabled`、`Active`、`AddGeneration`、`CommitActive`、`LockPlugin`、`PurgeButtonTokens`、`PurgeConversations`、`DeleteSecrets` 等；`packages.Preflight(Input{ReleaseJSON, Signature, PublicKey, Source, OS, Arch, ArchivePath}, stagingRoot, Limits)` 与 `Candidate.MoveTo`/`Discard`；`contracts/release` 的 `Verify`、`KeyID`；`manifest.CheckDependencies(m, Installed{Plugins, Adapters, Actions})`；`settings.Load`、`File.Settings`、`File.Update`、`File.CommandPrefixes`、`File.PrepareInstances`、`Settings.DefaultInstances`、`settings.ValidateListen`；`adminauth.New`、`Auth.Routes`、`Auth.Require`、`FirstToken`、`FirstTokenFile`；`webhost.New(Options{BaseURL, Proxy, Admin})`；`remote.Process.StartInstance(ctx, instanceID, configJSON)`；`policy.Policy.SetAdmins`；`adapter.Registry.SetDefaults`；`supervisor.Spec.DefaultLocale`。
- 在途 PR：#157（goal-130）在 `assemble` 中加入 `settings.Load` 与按 Session 发布的选项 schema；goal-68 修改 `startDevelopment`、`internal/inproc` 与日志的同进程 Handler。后合并者同步 main，设置对象只加载一次、由装配共用。
- goal-68 原来的「dev 来源拒绝 update 与 rollback」移到本切片的管理服务（见行为）。
- PLAN T-13 步骤 3 的「开发中的 Adapter 的实例取自 core_config.json」也由本切片完成：它与已安装 Adapter 的实例启动共用设置对象与同一段代码（见行为 5）。
- 2026-11-10 按 D-127：卸载分为只卸载、`--purge-data` 与 `--purge` 三种清理范围；动态任务与调度状态的表由 goal-127 加入清理对象。2026-11-09 按 D-105：admin 的参数是用户字符串 `<adapter_id>:<external_user_id>`。

```text
/goal 使用 $alonabot-execute 推进 goal-39（切片 T-09.4：管理服务、安装记录与 run）。goal 目录 goal-39，分支 goal-39，worktree .worktrees/goal-39，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：库模式与 `alonabot run` 共用根包的一套 Core 装配：读取 core_config.json，提供 HTTP Listener、webhost、管理认证与管理 API，接入命令前缀、Core Admin 与默认 locale，启动已安装 Adapter 的实例。建立唯一的管理服务，CLI 经它的 HTTP API 操作；四种角色的组件使用同一套 install、list、enable、disable、restart、uninstall（含 --purge-data 与 --purge）、pin 流程，所有变更按包串行；install 按角色执行同类唯一，`--source` 指定并持久化来源，`--replace` 显式替换同类组件；Adapter 实例的创建、启停与删除经此服务写回 core_config.json；启用时检查必需依赖，停用、卸载被依赖的包时列出受影响者并要求显式确认；admin add/remove/list。reload、update、rollback 的入口与校验在此定义，候选编排由 T-09.5 接入。

依据：PLAN T-09 步骤 3；CONTRACTS §2.2、§2.3、§5.3、§5.4、§7.4、§8（依赖段）、§9.1～§9.3、§15.1、§15.2、§15.4；DECISIONS D-26、D-42、D-59、D-77、D-81、D-82、D-105、D-127；TEST-MATRIX V-40、V-48、V-52、V-82、V-01 的安装分项、V-88 的接入分项。

范围（只新增或修改这些）：
- 根包 alonabot.go、assembly.go：Start 读取设置并建立日志、首次管理凭证、HTTP Listener、webhost、管理认证与管理 API；接入命令前缀、Core Admin、默认 locale 与 Adapter 实例；新增三个选项（见接口）
- cmd/alonabot/：run 改为调用根包（删除 main.go 的 run、launch 及 dependencies.go 中只供它使用的部分）；新增 install、list、enable、disable、restart、reload、update、rollback、uninstall、pin、adapter instance、admin 命令（经管理 API 访问）
- cmd/alonabot/e2e_test.go：迁移到新的 run（见测试）
- internal/core/state/schema.sql 与 installs.go、state.go：install 表加列（写进唯一的初始迁移），读写同步；Role 增加 agent 与 webui
- internal/core/host/prepare.go：role 为 agent 的组件与 plugin 相同地准备与激活（CONTRACTS §9.1「Agent 的声明与插件相同」）
- 新增 internal/core/manage/（管理服务与 HTTP API）
- internal/core/buttons、conversation、policy、tool、secret、state：只在卸载与清理中调用已有的清理方法，或补一个按插件删除的方法
- 上述包的测试文件
允许的新跨模块边：根包 -> internal/core/manage；cmd/alonabot -> 根包（已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 根包
   func Listen(addr string) Option           // 覆盖本次运行的 listen，校验同 settings.ValidateListen；非回环地址要求 https base_url 或 trusted_proxies
   func LogLevel(level slog.Level) Option    // 覆盖本次运行的全局日志级别
   func LogFormat(format string) Option      // 覆盖本次运行的 stderr 日志格式：text 或 json
   Plugin(root) 改为接受工程 manifest 的 role 为 plugin 或 agent，按 manifest 记录角色（agent 同样只能登记一个）；Start 的其他行为见下文；Run、Adapter、DataDir、Args 不变。
2. internal/core/state
   install 表加入列 version TEXT NOT NULL、publisher_key_id TEXT NOT NULL（未签名为空串）、pin TEXT NOT NULL、uninstalled INTEGER NOT NULL（不设 DEFAULT，写入时给出全部列）；Install 结构同步这些字段；RoleAgent = "agent"、RoleWebUI = "webui"，AddInstall 对 agent 与 webui 各只允许一条未卸载的记录（与现有 adapter 按 adapter_id 唯一的写法相同）。读写方法自定。
3. internal/core/manage
   type Service struct{ ... }
   type Options struct { /* Store、Settings（*settings.File）、启动与停止组件的回调、Packages 根目录、ReleaseBaseURL（缺省 https://github.com，测试指向本地模拟服务）、Candidates 等，执行者按实际依赖列出 */ }
   func New(opt Options) *Service
   type ErrNeedsConfirm struct{ Affected []string } // 实现 error；调用方带 confirm 重试
   // 以下方法都按包串行（复用 state.Store.LockPlugin），错误为带 Code 的 *validate.Error 或 *ErrNeedsConfirm
   type InstallRequest struct {
       Archive, ReleaseJSON, Signature string // 本地包；Source 非空时由服务下载，三者留空；Signature 可以为空（D-81）
       Source  string // GitHub 仓库 owner/repo；从 <ReleaseBaseURL>/<owner>/<repo>/releases/latest/download/ 取 release.json 与 release.json.sig（可选），再按其中的资产 URL 取当前平台（WebUI 为 any）的归档
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
   func (s *Service) Update(ctx context.Context, id string, req InstallRequest) error
   func (s *Service) Rollback(ctx context.Context, id, generation string, confirm bool) error
   func (s *Service) CreateInstance(ctx context.Context, adapterPackageID, instanceID string, config json.RawMessage) error
   func (s *Service) StartInstance(ctx context.Context, instanceID string) error
   func (s *Service) StopInstance(ctx context.Context, instanceID string) error
   func (s *Service) DeleteInstance(ctx context.Context, instanceID string) error
   func (s *Service) AddAdmin(ctx context.Context, user string) error    // <adapter_id>:<external_user_id>
   func (s *Service) RemoveAdmin(ctx context.Context, user string) error
   func (s *Service) Admins(ctx context.Context) []string
   type Package struct { ID, Role, AdapterID, Version, Source, PublisherKeyID, Pin string; Enabled, Uninstalled, Signed bool; Active, State string }
   type Candidates interface { Reload(ctx context.Context, id string) error; Update(ctx context.Context, id, staged string) error; Rollback(ctx context.Context, id, generation string) error } // T-09.5 起实现；缺省实现返回 UNSUPPORTED
   func (s *Service) Handler(auth *adminauth.Auth) http.Handler // /api/v1/...，全部经 auth.Require

行为：
- 根包 Start（库模式与 run 共用），按顺序：
  1) 取得 core.lock；settings.Load（失败为用法错误）；Listen、LogLevel、LogFormat 选项覆盖本次运行的对应设置；按设置打开日志（格式、级别、日志文件），设置了 proxy 时写入 HTTP_PROXY/HTTPS_PROXY，并把 HTTP_PROXY、HTTPS_PROXY、NO_PROXY 传给组件进程。这些步骤从 cmd/alonabot/main.go 的 run 原样迁来，错误与日志文案不变。
  2) 首次管理凭证：adminauth.FirstToken，interactive 为 stdin 与 stderr 都是终端；交互时把 token 打印到 stderr，非交互时按现有写法提示 admin_token 文件。
  3) 装配 Core（assemble），Router 的 Prefixes 传 File.CommandPrefixes；启动任何包之前把 Settings().Admins 转为 []policy.Admin 并 SetAdmins，调用 Registry.SetDefaults(Settings().DefaultInstances())。经 supervisor.Launch 启动组件时 Spec.DefaultLocale 取 Settings().DefaultLocale，同进程开发组件的 Bootstrap.default_locale 取同一个值；每次启动读当前设置。
  4) HTTP：adminauth.New、管理 mux（Auth.Routes 与 manage 的 Handler）、httpproxy、webhost.New，监听 listen 并 Serve；Listener 失败时 Start 返回错误。停止时先 Shutdown HTTP，再按 §4.2 停止组件。
  5) 按 install_seq 启动全部已启用的包（现有 startInstalled）；Adapter 包激活并 Bind 后，按 core_config.json 的 adapters.<adapter_id>.instances 经 File.PrepareInstances 准备，逐个 remote.Process.StartInstance；失败实例按 InstanceProblem 记 WARN 并跳过，其余照常启动。然后启动登记的开发组件；开发 Adapter 激活后按同样的规则启动 core_config.json 中它的实例，schema 取自该 Adapter 包的生成资源。
- cmd/alonabot run：`alonabot run [--data-dir DIR] [--listen ADDR] [--log-level L] [--log-format F] [--args ...]` 调用 alonabot.Run，DataDir 缺省为当前目录下的 data（与库模式相同），三个选项只在给出参数时传入。不带子命令运行 alonabot 等同 run。退出码沿用 exitUsage（参数与设置错误）与 exitFailure。
- install：Preflight（有签名时验签；无签名时按 D-81 需要 confirm，Package.Signed 为 false；验证长度、SHA256、平台与内层 manifest）→ MoveTo 不可变目录 → 写安装记录（install_seq 递增、source、publisher_key_id、version、role、adapter_id）。仍处于已安装状态的同 ID 返回 ALREADY_EXISTS，不覆盖。同类组件（同一 adapter_id 的 adapter、任一 agent、任一 webui）已安装时：未带 Replace 返回 ALREADY_EXISTS，safe_message 列出已安装的组件 ID；带 Replace 未带 confirm 返回 ErrNeedsConfirm（Affected 为将被卸载的组件与不迁移的 Adapter 实例）；确认后先完成新组件的下载与 Preflight，再按 uninstall 的缺省语义卸载旧组件，最后写新安装记录；旧组件原来启用时启用新组件。被替换 Adapter 的实例记录留在 core_config.json 的旧 adapter_id 下，不转给新组件。新组件在卸载旧组件之后失败时旧组件保持已卸载。已卸载的保留记录只允许来源与发布者公钥都相同的重装（复用数据），否则 PERMISSION_DENIED。首次信任新来源与新发布者公钥需要 confirm（CONTRACTS §9.2）。安装时用全部安装记录的 plugin_dependencies 检查必需依赖成环，成环拒绝。
- enable：CheckDependencies（Installed 来自已启用的包与运行中的 Adapter 描述），有缺失时 FAILED_PRECONDITION 并列出全部缺失项；可选依赖缺失只记诊断。已有 Active 时以该代次启动；首次安装没有 Active 时以「首个候选」启动（新代次、首装 SQLite、配置为编辑快照或 {}，提交后激活；Journal 由 T-09.5 接入，用一行 shortcut 注释写明）。Adapter 包启用后按上文启动它的实例。
- disable：停止并保留安装与数据；被已启用包的必需依赖引用时，未带 confirm 返回 ErrNeedsConfirm（列出受影响者），确认后执行，依赖方继续运行，调用它返回 UNAVAILABLE。
- restart：以当前 Active 启动新 Session（不导入编辑快照，不产生新代次）。
- uninstall：同 disable 的确认规则；停止、解除运行注册、删除未被引用的程序包目录。清理范围按 CONTRACTS §9.3（D-127）：PurgeNone 保留数据、配置、来源绑定与 uninstalled=1 的记录。PurgeData 另外删除数据：全部代次（代次目录、SQLite 与代次内的配置快照）、会话、按钮 token、Adapter 的游标；保留配置、来源绑定与 uninstalled=1 的记录。PurgeAll 删除数据与配置（编辑快照、Pending、secret、策略、工具策略、发信许可、该 Adapter 的实例配置）以及安装记录。两种清理都先列出将删除的对象，未带 confirm 返回 ErrNeedsConfirm，确认后删除，状态库中的对象在一个事务内删除；不删除插件自管的外部数据库与目录。清理对象的列表集中在一处，供 goal-127 追加任务与调度状态。
- dev 来源（库模式登记的开发组件）：update、rollback 与 uninstall 返回 FAILED_PRECONDITION，safe_message 说明它是开发组件；list 显示 source 为 dev。
- pin：持久化版本约束（manifest 的 SemVer 范围语法），update 只接受满足 pin 的候选。
- reload/update/rollback：完成输入与前置校验（来源绑定、公钥一致、pin、低于被依赖的 min_version 时需要 confirm），然后交给 Candidates；缺省实现返回 UNSUPPORTED。
- Adapter 实例：CreateInstance 按包内 adapter_config_schema.json 补缺失值并校验，经 File.Update 写回 adapters.<adapter_id>.instances，凭证保留原值或 ${...} 引用；DeleteInstance 经 File.Update 删除。实例配置与凭证不出现在 List、API 响应或日志中。
- admin add/remove 经 File.Update 写回 admins，成功后 SetAdmins，两步在同一把锁内；admin list 读 Policy.Admins。
- HTTP API（全部经 adminauth.Require）：每个方法一个 JSON 端点，需要确认时返回 409 与 {"affected": [...]}；错误返回 {"code", "safe_message"}。CLI 从 --admin-url（缺省取设置文件的 base_url）与 ALONABOT_ADMIN_TOKEN（或 --token-file）取得 Bearer token；确认通过 --yes 给出。CLI 不直接写状态库。
- Agent 组件（role: agent）：安装、启停、更新与卸载与插件相同，只多同类唯一；@agent 目标解析由 T-12.8 接入。
- WebUI 组件没有进程：enable/disable 只改记录，restart 返回 FAILED_PRECONDITION；文件的提供在 T-09.12。

测试（按编码规范「测试提交规则」）：
- 提交：
  - cmd/alonabot/e2e_test.go 迁移：TestMain 除构建 alonabot 外，用测试中的辅助函数把 examples/plugins/lifecycle 打成签名的安装包（release.json、release.json.sig、归档，密钥在测试中生成）；各测试在临时 data_dir 先 install --yes 与 enable，再 alonabot run。原有断言不变；TestLifecycleOverTCPAndStdio 去掉 stdio 子项（goal-69 恢复）。
  - 新增 cmd/alonabot/manage_e2e_test.go（//go:build unix；多列理由：本切片首次接通管理 CLI、依赖、替换与设置写回，没有其他端到端入口），用上面的打包辅助生成插件包、Adapter 包（examples/devbot/adapter 的工程打成 adapter 包）与依赖它们的插件包：
    - install 后 list 显示未启用；enable 产生第一个 Active；restart 只换 Session（代次不变）。
    - 依赖 Adapter 的插件在 Adapter 未安装时 enable 失败并列出缺失项；安装 Adapter 后成功；disable Adapter 不加 --yes 被拒并列出受影响插件，加 --yes 后执行。
    - uninstall 后同来源重装复用数据；不同来源重装被拒；--purge-data 列出对象并在 --yes 后删除数据，secret 与策略仍在；--purge 删除数据、配置与安装记录。
    - 已安装 Agent 测试组件时安装另一个：不带 --replace 返回 ALREADY_EXISTS；带 --replace 与 --yes 后旧组件已卸载、新组件沿用旧组件的启用状态；不同 adapter_id 的两个 Adapter 可以并存。
    - adapter instance create：配置写入 core_config.json，${...} 原写法保留；重启 run 后实例经 PrepareInstances 启动，失败实例单列且其余实例启动；list 与日志中没有凭证。
    - core_config.json 的 command_prefixes_by_instance 覆盖一个实例的前缀；admins 中的用户发命令时插件收到的 Actor.permissions 含 admin，admin add 后新用户的下一条命令也含 admin；default_locale 为 zh-CN 时，没有 locale 的 Actor 收到的缺省 usage 为中文。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：同一包的并发操作排队；ErrNeedsConfirm 的 Affected；API 的错误格式与 409；Package.State 的取值；--source 指向本地 Release 模拟服务时安装成功且来源被持久化；签名被篡改时安装失败、已有组件不受影响；未签名来源需要 --yes 且 list 标明未签名；pin 在重启后仍在；插件依赖成环被拒；reload、update、rollback 返回 UNSUPPORTED；dev 来源拒绝 update、rollback 与 uninstall；库模式（alonabot.Start）同样提供管理 API 与 webhost。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：候选 Journal 与准备编排（T-09.5）；唯一提交点与激活门禁（T-09.6）；崩溃恢复、回退与回收（T-09.7）；以目录运行开发组件与热部署（T-13.3）；WebUI 的会话取回与上传接口（T-09.13）；管理页面（T-09.8、T-09.9）；@agent 解析（T-12.8）。
```

### goal-68 · T-13.2 开发组件的配置与数据库

状态：已完成（[PR #160](https://github.com/AlonaBot/AlonaBot/pull/160)，合并提交 `65ebc5f`），证据见 TEST-MATRIX §3.2。合入：根包 startDevelopment 读取 `plugins/<id>/config.json`，Bootstrap 的数据库为 `dev.sqlite`（first_install/resume，migration_mode UNSPECIFIED），日志级别缺省 debug（settings 的来源级别优先，读取装配共用的 `settings.File`）；`state.Store.SetDevGenerationConfig(ctx, generationID, snapshot, configVersion, secretPaths)`；`internal/inproc.Launch.Log` 与 `logging.Output.Component(pluginID, generationID)`；host 在 Bootstrap 方式为 UNSPECIFIED 时接受明确回报；SDK 对 `dev:` 代次的 RESUME 允许 OnConfigLoad 变换配置。执行中审查者扩充授权（host、logging、inproc、state）。配置写回随 T-09.14；dev 来源拒绝 update/rollback/uninstall 与开发 Adapter 的实例在 goal-39。派发时：前置均已合并：T-13.1（[PR #149](https://github.com/AlonaBot/AlonaBot/pull/149)，合并提交 `bc39361`）、T-04.12（[PR #154](https://github.com/AlonaBot/AlonaBot/pull/154)，合并提交 `43a6135`）、T-07.6（PR #100）、T-08.5（PR #79）、T-08.11（PR #103）。2026-11-12 审查者按 main `56550de` 的代码改写，现有代码如下：

- **开发组件的 Bootstrap**：根包 `assembly.go` 的 `startDevelopment` 用 `bootstrap(generation, initialConfigVersion)` 构造，固定为 `CandidateConfigJson: "{}"`、`DatabaseDescriptor{Backend: NONE, MigrationMode: NONE}` 与 `LogLevel: INFO`；资源来自 `entry.resources`（插件对象的 `p.Generated()`）。
- **代次与提交**：代次固定为 `dev:<id>`，首次以 `AddGeneration` 建快照，经 `host.Prepare`/`host.Activate` 激活并在首次 `CommitActive`；开发组件的数据目录为 `<data_dir>/plugins/<id>`。
- **迁移方式核对**：SDK 在 `prepareMigrations` 中要求实际登记的迁移方式等于 Bootstrap 的 `migration_mode`，该值本应来自构建写入的 manifest。库模式没有构建步骤，所以本切片规定库模式开发组件的 Bootstrap 填 `MIGRATION_MODE_UNSPECIFIED`，SDK 只在同进程启动（`launch.inProcess`）且收到 UNSPECIFIED 时接受实际登记的方式，REGISTER 照实回报。已安装组件与 `alonabot run` 不受影响，审查者在合并时把这一条补进 CONTRACTS §15.4。
- **update 与 rollback**：现在没有这两个入口，dev 来源的拒绝由 goal-39（T-09.4）的管理服务实现。

```text
/goal 使用 $alonabot-execute 推进 goal-68（切片 T-13.2：开发组件的配置与数据库）。goal 目录 goal-68，分支 goal-68，worktree .worktrees/goal-68，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：库模式的开发组件每次启动以 <data_dir>/plugins/<id>/config.json（缺失时为空对象）为配置，经 SDK 准备与校验后直接生效，不走候选与 Journal；受管 SQLite 为 <data_dir>/plugins/<id>/dev.sqlite，不复制、不回退；开发组件的缺省日志级别为 debug。

依据：PLAN T-13 步骤 3；CONTRACTS §15.4；TEST-MATRIX V-64。

范围（只新增或修改这些）：
- 根包 assembly.go：开发组件的 Bootstrap 配置、数据库描述与日志级别（只改 startDevelopment 与 bootstrap 的开发分支）
- sdk/session.go：prepareMigrations 在同进程启动且 migration_mode 为 UNSPECIFIED 时接受实际登记的方式
- examples/devbot/plugin/：增加 config_schema.json 与 OnStart 中的 dev.sqlite 读写；alonabot_gen.go 只由生成器重新生成
- 根包 alonabot_test.go：扩展 TestLibraryLifecycle；上述包的 *_local_test.go
允许的新跨模块边：无。

行为：
- 配置：
  - 开发组件启动时读取 config.json；文件不存在时为 {}；内容不是 JSON 对象时，Start 返回错误并指明路径。
  - candidate_config_json 为其内容；target_config_version 为该组件 config_schema 的 $version（没有 schema 时为 1）；previous_config_version 为 nil，每次都按首装处理 OnConfigLoad。
  - SDK 按 goal-95 准备与校验，校验失败时该组件启动失败，Start 返回错误。
  - 准备后的配置不回写 config.json，只覆盖代次 `dev:<id>` 的快照。
- Secret：配置中的 secret 字段在开发模式下原样下发，不经 Secret Store，代次快照中也保存原样（dev 来源不分发）；在包文档写明这一差异。
- 数据库：数据库描述为 SQLite，路径 <data_dir>/plugins/<id>/dev.sqlite；activation_reason 在文件不存在时为 first_install，否则为 resume；migration_mode 为 UNSPECIFIED，由同进程的 SDK 按实际登记执行迁移（goal-32）并在 REGISTER 回报。不创建候选副本，不回退。插件没有登记迁移、也不打开数据库时，照常运行。
- Adapter：开发中的 Adapter 不使用 dev.sqlite（Adapter 不得声明数据库），描述保持 NONE。
- 日志：开发组件 Bootstrap 的 log_level 缺省为 DEBUG；settings 对该组件来源有显式级别时，用显式级别。
- Adapter 实例目录已在 goal-67 中实现，本切片不改。

测试（按编码规范「测试提交规则」）：
- 提交：alonabot_test.go（根包）追加：examples/devbot 的插件加一个带默认值与 required 字段的 Config 并在 OnStart 写 dev.sqlite：
  - 没有 config.json 时以默认值启动；写入合法的 config.json 后重启，插件看到新值；写入缺少 required 的配置时 Run 返回错误。
  - 两次启动之间 dev.sqlite 中的数据保留，没有 generations/ 目录与候选副本。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：config.json 非对象；dev.sqlite 不存在时的 activation_reason；登记了 versioned 与 self 迁移时的回报；独立进程启动时 UNSPECIFIED 仍被拒；显式设置级别时不被覆盖；日志中有该组件的 DEBUG 记录。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：dev 来源的 update 与 rollback 拒绝（goal-39 的管理服务）；热部署（T-13.3）；库模式示例与验收汇总（T-13.4）。
```

## 第 12 波

### goal-124 · T-04.11 注册写法统一

状态：已完成（[PR #159](https://github.com/AlonaBot/AlonaBot/pull/159)，合并提交 `7872102`），证据见 TEST-MATRIX §3.2。合入：`(*Plugin)`/`(*SubPlugin)` 的 `Route(method, path) *RouteBuilder`（`Public(mode)`、`Handle(http.Handler)`、`HandleFunc`）与 `Every(jobID, d)`、`Cron(jobID, expr)` 返回的 `*ScheduleBuilder`（`Timezone`、`Handle`），删除 `HTTPOption`、`sdk.Public` 与旧的 Handle/Every/Cron 写法；`Generated.ParamNames`（键为 runtime 函数名，函数字面量为 `导入路径:文件:行`；同一行字面量冲突的键省略并 WARN，`_` 与无名参数记空串回退 argN），Collect 查找时规范化 main 与导入路径前缀、泛型实参与 `-fm`；`AdapterHTTP` 为 `HTTP(route func(method, path string) *RouteBuilder)`；Core 的公开路由允许 state、signature、nonce 与 Adapter 的 platform token。Subscribe 的缺省路由 ID 与 `.ID` 归 goal-126。派发时：前置均已合并：T-04.12（[PR #154](https://github.com/AlonaBot/AlonaBot/pull/154)，合并提交 `43a6135`）、T-04.8（[PR #141](https://github.com/AlonaBot/AlonaBot/pull/141)，合并提交 `ac4aded`）、T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)，合并提交 `cc19e99`）。2026-11-12 审查者按 main `aa11454` 的代码写出范围、接口与测试段，现有代码如下：

- **插件对象与子插件**（`sdk/plugin.go`、`sdk/subplugin.go`）上，HTTP 与定时任务仍是位置参数形式：
  - `Handle(method, path string, h http.Handler, opts ...HTTPOption)`，其中 `HTTPOption`、包级 `Public(mode HTTPAuth) HTTPOption` 与 `HTTPRoute` 在 `sdk/registry.go`，校验在未导出的 `httpRegistry.Handle`；
  - `Every(jobID string, d time.Duration, fn func(*Context) error)`；
  - `Cron(jobID, expr, timezone string, fn func(*Context) error)`。
- **订阅**：`sdk.Subscribe[T](owner Registrar, routeID, eventName string)`，由 goal-126 改为接收 `BusEventType[T]`，并提供缺省路由 ID 与 `.ID(routeID)`。
- **命令参数名**：省略 `Args` 时为 arg1、arg2；结构体参数（D-115）不需要参数名。
- **生成资源**：`sdk.Generated` 经 `UseGenerated(p, alonabotGenerated)` 绑定到对象（D-137）。
- **调用方**：`examples/plugins/lifecycle` 与 `showcase` 使用 Handle、Every 与 `sdk.Public`。T-03.19（goal-110，PR #152）给 Adapter 加了 `sdk.AdapterHTTP`，接收与 `Handle` 同签名的登记函数。

与 goal-126 都修改 `sdk.Subscribe`：goal-126 先合并时，本切片只核对其缺省路由 ID 与 `.ID` 符合下文；后合并时由 goal-126 沿用本切片的规则。与 goal-101、goal-123、goal-126 都修改 `sdk/generate`，按执行顺序 §3 的生成器一行处理。PR #152 先合并时，本切片一并迁移 `AdapterHTTP` 与其测试。

```text
/goal 使用 $alonabot-execute 推进 goal-124（切片 T-04.11：注册写法统一）。goal 目录 goal-124，分支 goal-124，worktree .worktrees/goal-124，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：定时任务与 HTTP 路由改为与 D-70 相同的链式注册；命令省略 Args 时，参数名取 Handler 的参数名；Event Bus 订阅的路由 ID 缺省为事件名。协议不变。Hook 的同样改写由 T-05.3（goal-17）完成。

依据：PLAN T-04 步骤 11；DESIGN §D；DECISIONS D-70、D-79、D-112、D-114；TEST-MATRIX V-110 的 SDK 分项。

范围（只新增或修改这些）：
- sdk/plugin.go、sdk/subplugin.go：HTTP 与定时任务的入口改为 builder；sdk/registry.go：HTTP 与定时任务的 builder、删除 HTTPOption 与包级 Public；sdk/sdk.go 的 Collect 只加漏写 Handle 的检查与参数名查找
- sdk/generated.go：Generated 追加参数名表；sdk/dispatch.go 只在命令参数绑定处读取参数名
- sdk/generate/：为 Handler 记录参数名并写入生成文件
- sdk/adapter_run.go（PR #152 已合并时）：AdapterHTTP 迁到新的 HTTP builder
- examples/plugins/{lifecycle,showcase}/ 的 Go 源文件；各处 alonabot_gen.go 只由生成器重新生成
- 提交的测试中以字符串嵌入的插件源码与测试插件（internal/core/adapter/remote/remote_test.go 等）只迁移注册写法，断言不变
- 上述 SDK 包的 *_local_test.go
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
sdk（替换与新增；*Plugin 与 *SubPlugin 各有一套同名方法）
   func (r *Plugin) Route(method, path string) *RouteBuilder
   type RouteBuilder struct{ ... }
   func (b *RouteBuilder) Public(mode HTTPAuth) *RouteBuilder // 公开路由；不调用时需要 Core 会话
   func (b *RouteBuilder) Handle(h http.Handler)
   func (b *RouteBuilder) HandleFunc(f func(http.ResponseWriter, *http.Request))
   func (r *Plugin) Every(jobID string, d time.Duration) *ScheduleBuilder
   func (r *Plugin) Cron(jobID, expr string) *ScheduleBuilder
   type ScheduleBuilder struct{ ... }
   func (b *ScheduleBuilder) Timezone(name string) *ScheduleBuilder // 只用于 Cron，缺省 UTC；Every 调用时由 Collect 报错
   func (b *ScheduleBuilder) Handle(fn func(*Context) error)
   Generated 追加 ParamNames map[string][]string // 键见下文；值为 *sdk.Context 之后的参数名
   AdapterHTTP（PR #152 已合并时）改为 HTTP(route func(method, path string) *RouteBuilder)
删除：(*Plugin)/(*SubPlugin).Handle(method, path, h, opts...)、原 Every(jobID, d, fn) 与 Cron(jobID, expr, timezone, fn)、HTTPOption、包级 Public，不留别名。

行为：
- HTTP：
  - Route 声明一条 method+path 路由，以 Handle 或 HandleFunc 结束。
  - 注册校验、错误文本、route_id 与 Listener 行为都与原 Handle 相同：私有路由需要 Core 会话，公开路由必须声明回调验证方式，Adapter 另可使用 AuthPlatformToken，插件不能。
- 定时任务：Every 与 Cron 返回同一种 builder，以 Handle 结束。路由 ID 仍为 "job." 加 jobID，描述与原写法逐字段相同。
- 漏写或重复 Handle：Collect 报错并带 route_id，与命令相同。
- 命令参数名：
  - 生成器为插件包中第一个参数是 `*sdk.Context` 的包级函数、方法与函数字面量，记录其后参数的名字，写入 Generated.ParamNames。
  - 键：函数与方法为 runtime 的函数全名（`importpath.Func`、`importpath.(*T).Method`、`importpath.T.Method`）；函数字面量为 `importpath:文件名:行号`，行号是 `func` 关键字所在行。
  - SDK 在 Collect 时用 `runtime.FuncForPC` 取 Handler 的名字（方法值去掉 `-fm` 后缀）或字面量的文件与行，查找参数名。
  - 写了 Args 时以 Args 为准，个数仍须与参数一致；查不到时保持 arg1、arg2；结构体参数不用参数名；Optional 的 name 指同一套名字。
- 订阅：路由 ID 缺省为事件名，`.ID(routeID)` 另给；同一事件订阅两次且都没有给 ID 时，Collect 报重复的 route_id（goal-126 已合并时只核对）。
- 调用方：examples/plugins 与测试中以字符串嵌入的插件源码按新写法修改，与参数名相同的 Args 删去。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86：不向 sdk/ 提交测试）。已有调用 HTTP、定时任务与 Subscribe 的集成测试（含 TestRemoteHTTPUpgrade、示例相关测试）只迁移写法，断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 新旧写法注册描述的等价性（HTTP 的各认证模式、Every 与 Cron 含 Timezone）；漏写、重复 Handle 与 Every 调用 Timezone 的报错；
  - 参数名的各种来源（包级函数、值与指针方法值、函数字面量）、Args 优先、查不到时的回退；
  - 生成器重复生成一致；
  - internal/core/supervisor 的本地集成（//go:build unix）：真实插件的 HTTP 路由、定时任务与按参数名绑定的命令。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race . ./cmd/... ./internal/core/... ./sdk/... ./adapters/...；go build ./...。

不做：Hook 的注册写法（goal-17）；Service 与 Event Bus 的声明（goal-126）；协议与 Core 的改动。
```

### goal-126 · T-04.13 Service 与 Event Bus 的带类型声明

状态：已完成（[PR #158](https://github.com/AlonaBot/AlonaBot/pull/158)，合并提交 `464ccfd`），证据见 TEST-MATRIX §3.2。合入：`sdk.JSON[In, Out](target) Service[In, Out]`、`sdk.Proto[In, Out](target)`、`sdk.Serve(r, s) *ServiceBuilder[In, Out]`（`Access`、`Tool`、`Handle`）、`Service.Call(ctx, in)`；`sdk.BusEvent[T](name) BusEventType[T]` 与 `Publish(ctx, payload)`、`sdk.Subscribe(r, e) *Subscription[T]`（路由 ID 缺省为事件名，`ID(routeID)` 另给，`Handle`）；删除 CallJSON、CallProto、Publish 与旧形式的 JSON、Proto、Subscribe。target 的组件部分在握手取得插件 ID 后、REGISTER 前核对，`@agent` 只检查格式；调用方生成器为依赖模块的声明嵌入输入、输出 Schema；Schema 校验失败的公开错误只含 target 与方向。范围修订：组件归属改在握手后核对（D-52，SDK 根包不链接 yaml）。派发时：前置 T-04.12（[PR #154](https://github.com/AlonaBot/AlonaBot/pull/154)，合并提交 `43a6135`）已合并。2026-11-12 审查者按 main `9983182` 的代码写出范围、接口与测试段，现有代码如下：

- **Service 注册**：`sdk.JSON[In, Out any](owner Registrar, name string) *JSONService[In, Out]` 与 `sdk.Proto[In, Out proto.Message](owner Registrar, name string) *ProtoService[In, Out]`（`sdk/services.go`），builder 有 `Access`、`Tool`（仅 JSON）与 `Handle`。
- **Service 调用**：`sdk.CallJSON[In, Out any](ctx, target string, in In) (Out, error)` 与 `sdk.CallProto[Out proto.Message](ctx, target string, in proto.Message) (Out, error)`。
- **Event Bus**：`sdk.Subscribe[T any](owner Registrar, routeID, eventName string) *Subscription[T]` 与 `sdk.Publish[T any](ctx, eventName string, payload T) (string, error)`（`sdk/bus.go`）；Publish 由 SDK 补上 `plugin_id/` 前缀。
- **`Registrar`**：由 `*sdk.Plugin` 与 `*sdk.SubPlugin` 实现，内部经 `owner.registrar()` 取得插件与子插件归属。
- **生成器**：`sdk/generate/service.go` 只检查 `JSON`、`Subscribe`、`CallJSON`、`Publish` 四个泛型入口的类型参数，都必须是包级具名非泛型类型，其中 JSON 注册生成 ServiceSchemas。
- **调用方**：main 上除 SDK 自身外，没有提交代码使用这些函数；`sdk/llm` 与 goal-83 均未合并。

本切片与 goal-124（T-04.11）都修改 `sdk.Subscribe`，与 goal-101、goal-123、goal-124 都修改 `sdk/generate`，后合并者同步 main，按对方的写法重做自己的改动。

```text
/goal 使用 $alonabot-execute 推进 goal-126（切片 T-04.13：Service 与 Event Bus 的带类型声明）。goal 目录 goal-126，分支 goal-126，worktree .worktrees/goal-126，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Service 与 Event Bus 事件各用一个带类型的声明表示，注册、调用、发布与订阅都从声明出发，名字与 Go 类型只写一处；删除按字符串目标与类型参数直接调用、发布与订阅的函数。协议与 Core 不变。

依据：PLAN T-04 步骤 13；DESIGN §D、§G；CONTRACTS §3、§9.1；DECISIONS D-70、D-107、D-112、D-114、D-122；TEST-MATRIX V-113 的声明分项。

范围（只新增或修改这些）：
- sdk/services.go：声明类型 Service[In, Out]、JSON 与 Proto 声明、Serve、Call；删除旧的 JSON、Proto（以注册器开头的形式）、CallJSON、CallProto
- sdk/bus.go：声明类型 BusEventType[T]、BusEvent、Publish 方法、新形式的 Subscribe；删除旧的 Subscribe 与 Publish
- sdk/sdk.go 的 Collect 只加声明相关的检查调用点
- sdk/generate/service.go：类型参数检查改为针对 JSON、Proto、BusEvent 的声明处；ServiceSchemas 内容不变
- docs/plugin-guide.md 第 6 节由审查者在合并前改写，执行者不改
- 上述 SDK 包的 *_local_test.go；sdk/generate 的 *_local_test.go
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
sdk（新增或替换）
   type Service[In, Out any] struct{ /* 只含未导出字段：target、编码 */ } // 值类型，零值无效
   func JSON[In, Out any](target string) Service[In, Out]                   // target 为 "plugin_id/service_name" 或 "@agent/service_name"
   func Proto[In, Out proto.Message](target string) Service[In, Out]
   func Serve[In, Out any](r Registrar, s Service[In, Out]) *ServiceBuilder[In, Out] // 返回服务 builder
   type ServiceBuilder[In, Out any] struct{ ... } // 由原 JSONService/ProtoService 合并而来
   func (b *ServiceBuilder[In, Out]) Access(a Access, allowlist ...string) *ServiceBuilder[In, Out]
   func (b *ServiceBuilder[In, Out]) Tool(description string, defaultPolicy ToolDefault) *ServiceBuilder[In, Out] // Proto 声明调用时由 Collect 报错
   func (b *ServiceBuilder[In, Out]) Handle(h func(*Context, In) (Out, error))
   func (s Service[In, Out]) Call(ctx context.Context, in In) (Out, error)
   type BusEventType[T any] struct{ /* 只含未导出字段：name */ }
   func BusEvent[T any](name string) BusEventType[T]                         // name 为 "plugin_id/event_name"
   func (e BusEventType[T]) Publish(ctx context.Context, payload T) (eventID string, err error)
   func Subscribe[T any](r Registrar, e BusEventType[T]) *Subscription[T]    // 路由 ID 缺省为事件名
   func (s *Subscription[T]) ID(routeID string) *Subscription[T]             // 另给路由 ID；Handle 之后调用时由 Collect 报错
删除：CallJSON、CallProto、Publish 函数，以及原 JSON(owner, name)、Proto(owner, name)、Subscribe(owner, routeID, eventName)、JSONService、ProtoService，不留别名。

行为：
- 注册：
  - Serve 对 Service 声明开始注册，Access、Tool、Handle 的语义与原 builder 相同，REGISTER 中的服务名是 target 的 service_name 部分，描述与原写法逐字段相同。
  - Collect 核对：target 的组件部分等于本组件的插件 ID（`role: agent` 的组件为 `@agent`）；声明不是零值；同一 target 只 Serve 一次；Proto 声明调用 Tool 报错。错误为英文并带 target。
- 调用：Call 发出的请求与原 CallJSON、CallProto 相同（JSON 时按 Schema 校验输入输出，Handler 内调用时关联父请求）。零值声明返回 INVALID_ARGUMENT，不发出请求。
- Event Bus：
  - Publish 核对 name 的组件部分等于本插件 ID，不一致时返回 INVALID_ARGUMENT，不发出请求；其余与原 Publish 相同，只是不再补前缀。
  - Subscribe 的路由 ID 缺省为事件名（D-112），`.ID(routeID)` 另给；订阅描述与原写法相同。
- 生成器：类型参数检查（包级具名非泛型类型）改为针对 `sdk.JSON`、`sdk.Proto`、`sdk.BusEvent` 的声明处，诊断位置落在声明处；JSON 声明照旧生成 ServiceSchemas，内容不变。
- 调用方：main 上没有提交代码使用这些函数。若执行时已有新的调用方（goal-83 的 `sdk/llm`、官方插件、测试中以字符串嵌入的插件源码），按新写法迁移；`sdk/llm` 导出 `llm.GenerateService = sdk.JSON[Request, Response]("@agent/llm.generate")`，`llm.Generate` 与 `llm.Ask` 的签名不变。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86：不向 sdk/ 提交测试）。已有调用 Service 与 Event Bus 的集成测试若存在，改用声明，断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 声明与 Collect 的拒绝项：零值声明、组件部分不符、重复 Serve、Proto 调用 Tool；
  - Call 的零值与真实 Session 往返（JSON 与 Proto），Handler 内调用关联父请求；
  - Publish 的组件检查与发布载荷，Subscribe 的缺省与 `.ID` 路由 ID；
  - 生成器诊断落在声明处、ServiceSchemas 与改写前逐字节相同；
  - internal/core/supervisor 的本地集成（//go:build unix）：两个真实插件进程，一个 Serve 并 Publish，另一个以同一声明 Call 与 Subscribe。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race . ./internal/core/... ./sdk/...；go build ./...。

不做：协议与 Core 的改动；`sdk/llm` 的新功能（goal-83）；HTTP、定时任务与 Hook 的注册写法（goal-124）。
```

### goal-129 · T-04.14 命令名覆盖的 Core 侧

状态：已完成（[PR #156](https://github.com/AlonaBot/AlonaBot/pull/156)，合并提交 `0726d66`），证据见 TEST-MATRIX §3.2。合入：`command_policy.names` 列；state 的 `PutCommandNames`、`AllCommandNames`，Put 与 Delete 各自只写自己的列，删除时名称清理与空行删除在一个事务内；policy 的 `LoadCommandNames`、`SetCommandNames(ctx, pluginID, routeID, names, routes)`、`DeleteCommandNames`、`CommandNames(pluginID)`；`router.Options.CommandNames`、`CommandEntry.Names`（主名在前）与 `CommandIndex.InvalidNames()`。goal-93 的管理 API 经 policy 的上述方法读写名称；goal-39 的 --purge 删除整行。派发时：前置均已合并：T-03.17（[PR #150](https://github.com/AlonaBot/AlonaBot/pull/150)，合并提交 `8eb7846`）、T-04.10（[PR #133](https://github.com/AlonaBot/AlonaBot/pull/133)，合并提交 `3d618cb`）、T-04.4（PR #88）。2026-11-12 审查者按 main `61bff90` 的代码写出范围、接口与测试段，要点如下：

- **状态库**：`command_policy` 表（`state.CommandPolicy{PluginID, Scope, Permissions, Enabled}`）现在由 `policy.SetOverride` 整行写入或删除（`PutCommandPolicy` 使用 `UpdateAll`，两项都为 nil 时 `DeleteCommandPolicy`）。按 PLAN 步骤 14，命令名写进同一张表的路由作用域行。为避免权限与启用的修改冲掉命令名，三个写入方法改为只更新各自的列，删除时只删三列都为 NULL 的行。
- **Router**：`rebuildCommands` 在持有 `r.mu` 时从已激活插件的 `Routes()` 建立 `CommandIndex`。`CommandEntry{PluginID, RouteID, Priority, Path, Alias}` 记录命中的路径与是否为别名。`commandUsage` 从 `route.GetCommand()` 的 path 与 aliases 拼出 usage。`commandMenu` 以 `command.GetPath()` 作为菜单名，并用 `index.Lookup(path)` 确认胜出者。`CommandsChanged()` 由 `Policy.SetCommandsChanged` 注入，目前只调度菜单同步。
- **索引读取不能做 I/O**：索引在锁内建立，所以命令名覆盖由 policy 保存在内存快照中，以注入函数交给 Router。
- **路径校验**：命令路径的格式校验为 `validate.IsCommandToken`，同插件路径冲突在 `validate.Register` 中报 ALREADY_EXISTS。
- **库模式装配**：`assembly.go` 已传入 `Policy`、`CommandState` 与 `SetCommandsChanged`。

与 goal-71（T-14.1）都涉及 `core.command.list` 返回的命令名，后合并者同步。goal-39（T-09.4）的 `--purge` 删除命令策略行时，命令名随之删除。

```text
/goal 使用 $alonabot-execute 推进 goal-129（切片 T-04.14：命令名覆盖的 Core 侧）。goal 目录 goal-129，分支 goal-129，worktree .worktrees/goal-129，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：管理员可以给一条按路径匹配的命令保存命令名覆盖，取代插件声明的路径与别名。Core 的匹配、遮蔽、usage 回复与命令菜单都使用生效的命令名，route_id、命令组与策略的作用域不变。本切片提供服务层与 Router 的行为，管理 API 由 goal-93 提供。

依据：PLAN T-04 步骤 14；CONTRACTS §2.2、§2.3、§8（命令菜单）、§9.3；WIRE-FORMAT §3.5（`core.command.list`）；DECISIONS D-96、D-127、D-128；TEST-MATRIX V-115 的 Core 分项。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一的初始迁移中给 command_policy 表加 names 列；internal/core/state/policy.go：行类型字段与读写方法
- internal/core/policy/：新增 names.go（命令名覆盖的校验、持久化与内存快照）；policy.go 只给 Policy 加快照字段
- internal/core/router/：
  - dispatch.go：Options 追加 CommandNames；
  - router.go：New 保存 CommandNames；
  - index.go：rebuildCommands 使用生效的命令名，CommandEntry 追加 Names，新增失效覆盖的诊断；
  - command.go：commandUsage 使用生效的命令名；
  - commands.go：commandMenu 使用生效的主名；CommandsChanged 先重建命令索引，再调度菜单同步。
- 根包 assembly.go：启动时载入命令名快照，并给 router.Options 传入 CommandNames
- internal/core/adapter/remote/remote_test.go：新增一个主路径集成测试
- 上述包的 *_local_test.go
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/core/state
   command_policy 表追加 names TEXT 列（NULL 表示没有命令名覆盖；只出现在路由作用域行）
   CommandPolicy 追加 Names [][]string `gorm:"column:names;serializer:json"` // nil 为 NULL；非 nil 时至少一项，第一项为主名
   func (s *Store) PutCommandPolicy(ctx context.Context, row CommandPolicy) error   // 改为只 upsert permissions 与 enabled 两列，保留 names
   func (s *Store) PutCommandNames(ctx context.Context, pluginID, scope string, names [][]string) error // 只 upsert names 列；nil 时置 NULL
   func (s *Store) DeleteCommandPolicy(ctx context.Context, pluginID, scope string) error // 改为只在 permissions、enabled 与 names 都为 NULL 时删除该行，幂等
   func (s *Store) AllCommandNames(ctx context.Context) ([]CommandPolicy, error)     // names 非 NULL 的全部行
   policy.SetOverride 两项都为 nil 时，先经 PutCommandPolicy 把两列置 NULL，再调用 DeleteCommandPolicy；其余行为与错误码不变
2. internal/core/policy
   func (p *Policy) LoadCommandNames(ctx context.Context) error // 从状态库载入内存快照；装配在创建 Router 之前调用
   func (p *Policy) SetCommandNames(ctx context.Context, pluginID, routeID string, names [][]string, routes []*pb.RouteDescriptor) error
   // routes 为该插件当前注册的路由（未注册时为 nil）；成功返回前已持久化并更新快照，随后调用 commandsChanged
   func (p *Policy) DeleteCommandNames(ctx context.Context, pluginID, routeID string) error // 删除覆盖即恢复声明；成功后调用 commandsChanged
   func (p *Policy) CommandNames(pluginID string) map[string][][]string // 内存快照的深副本，route_id → 命令名列表；不做 I/O
3. internal/core/router
   Options 追加 CommandNames func(pluginID string) map[string][][]string // nil 时只用声明的路径与别名；在持有 r.mu 时调用，实现不得做 I/O
   CommandEntry 追加 Names [][]string // 该路由生效的全部命令名，主名在前；Path 与 Alias 的含义不变（Alias 表示 Path 不是主名）
   type InvalidName struct {
       PluginID, RouteID string
       Name              []string // 失效的覆盖项；整条覆盖指向未注册路由时为 nil
       Reason            string   // "declared"：与本插件其他命令声明的名字相同；"unregistered"：路由未注册或不是按路径匹配的命令
   }
   func (x *CommandIndex) InvalidNames() []InvalidName // 按 plugin_id、route_id 与名字排序，返回的切片只读

行为：
- 校验（SetCommandNames）：
  - names 非空，否则 INVALID_ARGUMENT；每个命令名非空，每个 token 满足 validate.IsCommandToken；列表内不能重复（按完整 token 序列比较）。
  - routes 非 nil 时：目标路由必须存在且是 COMMAND_MATCH_PATH 的命令，文本匹配与 Raw 返回 INVALID_ARGUMENT；任一命令名与本插件其他命令声明的路径或别名相同时，返回 ALREADY_EXISTS，safe_message 带冲突的 route_id（不论那条命令自己是否被覆盖）。
  - routes 为 nil（插件未注册）时只做格式检查，以及与其他覆盖之间的检查。
  - 与本插件其他路由的覆盖相同时返回 ALREADY_EXISTS，并带出冲突的 route_id。
  - 校验失败时不触及状态库与快照。
- 生效：rebuildCommands 对有覆盖的命令只用覆盖的命令名参与匹配。与本插件其他命令声明的名字相同的覆盖项不进入索引，记为 InvalidName{Reason: "declared"}，一项不剩时退回声明的名字。快照中指向未注册、或不是按路径匹配的路由的覆盖，记为 "unregistered"。
- 不变的部分：route_id、参数、Handler、命令组归属不变；策略的命令组作用域仍按声明的规范路径。跨插件同名照旧按 priority、plugin_id、route_id 遮蔽，Shadows 按生效的名字计算。禁用的命令占有它生效的命令名。
- 变化：SetCommandNames 与 DeleteCommandNames 成功后经 commandsChanged 通知 Router。CommandsChanged 先在锁内重建命令索引，再调度菜单同步；下一次派发生效，不重启插件。
- usage 回复：命令与别名部分改用生效的命令名，主名在前，其余以「 | 」连接。自定义 usage 的 `command` 参数取生效的主名。
- 命令菜单：菜单名取生效的主名，主名只有一个 token 时才算根命令，以主名 Lookup 确认胜出者。说明仍按声明的规范路径查找（D-138：别名与生效的命令名共用原规范路径的说明）。
- 卸载：命令名随命令策略行保存，停用时不删除；`--purge` 由 goal-39 删除命令策略时一并删除，本切片只核对 DeleteCommandPolicy 的新语义不留下空行。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 internal/core/adapter/remote/remote_test.go 的 TestCommandNameOverride。复用测试二进制中的插件进程、真实 Session、Router、policy 与临时 SQLite：
  - 保存覆盖后旧名不再匹配；新名匹配，Handler 收到相同的参数；
  - 删除覆盖后恢复声明的名字；
  - 全程不重启插件。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 空列表、非法 token、列表内重复；文本匹配与 Raw 路由；
  - 与其他命令声明名或覆盖重名时的 ALREADY_EXISTS 与 route_id；未注册路由时允许保存；
  - 插件更新后声明名与覆盖相同时，覆盖项失效并出现在 InvalidNames 中，覆盖一项不剩时退回声明名；
  - 跨插件遮蔽按生效名计算；usage 与菜单中的名字；
  - LoadCommandNames 后 Core 重启仍然生效；CommandsChanged 触发的重建与菜单同步。
  - SetOverride 与 SetCommandNames 交替写入同一路由作用域时互不覆盖，三者都清空后行被删除。

完成条件：gofmt -l 无输出；go vet ./...；go test -race . ./internal/core/... ./sdk/...；go build ./...。

不做：管理 API（goal-93）；页面（goal-44）；按命令组整体改名；同一插件内两条命令互换名字；`core.command.list` 的实现（goal-71）；卸载清理（goal-39）。
```

### goal-17 · T-05.3 三阶段 Hook 与 send_forbidden

状态：已完成（[PR #155](https://github.com/AlonaBot/AlonaBot/pull/155)，合并提交 `53393aa`），证据见 TEST-MATRIX §3.2。合入：SDK 的 `(*Plugin)`/`(*SubPlugin)` 的 `PreEvent(routeID, event EventType) *PreEventBuilder`、`BeforeSend(routeID) *BeforeSendBuilder`、`AfterSend(routeID) *AfterSendBuilder`，各 builder 的 `Priority` 与 `Handle`（BeforeSend 为 `func(*Context, message.Session, message.Chain) (message.Chain, error)`，AfterSend 为 `func(*Context, message.Session, *message.SendResult, error) error`），`(*Context).Editing() message.MessageRef`；`internal/core/hook` 的 `New(r *router.Router, Options{Prepare, CheckEvent}) *Runner` 与 `Runner.PreEvent`、`BeforeSend`（带改写复检的 check 回调）、`AfterSend`；`router.Options.PreEvent` 与 `Invocation.SendForbidden`；`message.New(..., hooks)`；HookSpec 的 `event_type_filter = 2`。审查中范围修订：BeforeSend 的改写结果复检、PreEvent 改写的事件按接收 Session 准备媒体、Adapter 扩展事件的改写经实例复检（`sdk/sdk.go` 与 remote 的复检出口一并修改）。合入 main 时按 CONTRACTS §2 的 PreEvent → 多轮等待 → 命令顺序接入 T-05.4。派发时：前置 T-05.2（[PR #145](https://github.com/AlonaBot/AlonaBot/pull/145)，合并提交 `0f860a9`）已合并。2026-11-12 审查者按 main `3e4db62` 的代码改写接口段，要点如下：
- router 不能 import hook，PreEvent 改为 `router.Options.PreEvent` 注入函数；没有 `router.Dispatch` 类型。
- Router 的可信事件为 `trustedEvent{retained, payload}`：可改写的业务载荷只在 payload 中，`rebuild` 时与可信绑定合并。message.deleted、button.clicked 与 scheduled 没有业务载荷。
- `router.Invocation` 现有 PluginID、Event、DefaultTarget，作为 `wire.Request.Attach` 传出，服务函数从 `wire.Inbound.ParentAttach` 取回。
- `InvocationContext.send_forbidden = 5` 已在协议中；BeforeSendRequest 现用到字段 4，AfterSendRequest 现用到字段 6。`HookSpec{phase}` 与 `HookPhase` 已在 route.proto，`RouteDescriptor.event_type_filter` 为通用字段。
- `message.New(adapters, p, lease, issue, requestTimeout)` 的现有调用方为 `internal/core/message/permissions_test.go` 与 `internal/core/adapter/remote/segments_test.go`。
- SDK 事件类型为 `sdk.EventType`，标准事件为常量（例如 `sdk.MessageCreated`），Adapter 扩展事件为 `sdk.AdapterEvent[T]`，Handler 接收 *T。

core.delete（goal-107，PR #144）与 core.action（goal-106，PR #146）都在 `internal/core/message` 中实现，派发时尚未合并：执行者交付前同步最新 main，main 上已有的一并加入禁止名单。

与 goal-125 并行时，后合并者把 Hook 的注册迁到对应的写法：本切片先合并时，goal-125 把 `Hooks` 连同其余根插件注册方法一并删除，改为插件对象上的 `p.PreEvent`、`p.BeforeSend`、`p.AfterSend`，builder 与 Handler 签名不变；goal-125 先合并时，本切片直接在插件对象上提供这三个方法。与 goal-52 并行时，后合并者按「PreEvent → 等待匹配 → 命令层」的顺序接入调用点。与 goal-104、goal-109 同改 message.Service 的 Send 与 Edit，后合并者保留对方的来源准备与目标补齐。

```text
/goal 使用 $alonabot-execute 推进 goal-17（切片 T-05.3：三阶段 Hook 与 send_forbidden）。goal 目录 goal-17，分支 goal-17，worktree .worktrees/goal-17，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：运行 PreEvent、BeforeSend、AfterSend 三类 Hook。
- PreEvent 只能改写业务 payload 或 Stop，失败时结束本次入站链。
- 每次显式或自动 Send（以及 Edit）独立执行一组 BeforeSend → Adapter → AfterSend。BeforeSend 只能改 Chain 或拒绝，改后重新预检；AfterSend 只观察。
- 出站 Hook 的调用带 send_forbidden，该标记保存在 Core 的调用记录（router.Invocation）中。Hook 内直接调用 core.send、core.edit（以及 main 上已有的 core.action、core.delete）被拒。

依据：PLAN T-05 步骤 5、6；CONTRACTS §2.3 第 3～5 条、§8（Edit 的 Hook 段）；WIRE-FORMAT §3.5（Hook 输入的追加）；MESSAGE-MODEL §5；DECISIONS D-70、D-92、D-93、D-104、D-112；TEST-MATRIX V-15 的直接路径分项、V-110 的 Hook 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/invocation.proto：BeforeSendRequest 与 AfterSendRequest 各追加 SendOperation operation 与 optional string edited_platform_message_id，并新增 enum SendOperation。随后运行 go generate ./...。本切片获授权只追加这些。
- contracts/wirespec_test.go：追加编号检查
- contracts/validate/register.go：HOOK 路由的注册校验
- 新增 internal/core/hook/
- internal/core/router/：
  - mux.go：Invocation 追加 SendForbidden；
  - dispatch.go：Options 追加 PreEvent；
  - router.go：New 保存 PreEvent，dispatch 在 Core 权限计算之后、命令层之前调用它。
  - 不改其他派发逻辑。
- internal/core/message/：
  - send.go：New 追加 hooks 参数，Send 在调用 Adapter 前后执行出站 Hook；
  - edit.go：Edit 同上，并删除 goal-51 的 shortcut 注释；
  - Send、Edit 以及 main 上已有的 action.go、delete.go 拒绝 SendForbidden 的父调用；
  - 同步 message.New 的全部调用方（permissions_test.go、remote/segments_test.go 及派发时 main 上的其他调用点），传 nil。
- sdk/：
  - 新增 hooks.go（注册、分发与 Context.Editing）；
  - base.go 的 Root 增加 Hooks(*HookRegistry)，Base 提供默认实现；
  - session.go 的 buildRegister 输出 HOOK 路由；
  - dispatch.go 按 route_id 分发 Hook 请求。
- 上述包的 *_local_test.go，以及 internal/core/supervisor/hook_local_test.go（//go:build unix）
允许的新跨模块边：无（hook、message 与 router 同属 internal/core 模块；router 与 message 不 import hook）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   enum SendOperation { SEND_OPERATION_UNSPECIFIED = 0; SEND_OPERATION_SEND = 1; SEND_OPERATION_EDIT = 2; }
   BeforeSendRequest 追加 SendOperation operation = 5; optional string edited_platform_message_id = 6;
   AfterSendRequest 追加 SendOperation operation = 7; optional string edited_platform_message_id = 8;
2. internal/core/router
   Invocation 追加 SendForbidden bool // 出站 Hook 调用为 true；只由 Core 设置
   Options 追加 PreEvent func(ctx context.Context, ev *pb.Event) (payload *pb.EventPayload, stop bool, err error) // nil 时不执行；ev 为 rebuild 后的事件，返回的 payload 替换 trustedEvent 的业务载荷
3. internal/core/message
   type Hooks interface {
       BeforeSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, parent *router.Invocation) (*pb.MessageChain, *pb.ErrorInfo)
       AfterSend(ctx context.Context, op pb.SendOperation, edited string, target *pb.Target, chain *pb.MessageChain, result *pb.SendResult, errInfo *pb.ErrorInfo, parent *router.Invocation)
   }
   func New(adapters *adapter.Registry, p *policy.Policy, lease ..., issue ..., requestTimeout time.Duration, hooks Hooks) *Service // 其余参数不变；hooks 为 nil 时不执行出站 Hook
4. internal/core/hook（*Runner 实现 message.Hooks）
   type Runner struct{ ... }
   func New(r *router.Router) *Runner // 每次执行时从 r.Plugins() 读取已激活插件的 HOOK 路由，经 r.Caller 与 Plugin.Call 调用
   func (h *Runner) PreEvent(ctx context.Context, ev *pb.Event) (*pb.EventPayload, bool, error)
   func (h *Runner) BeforeSend(...)（同 message.Hooks）
   func (h *Runner) AfterSend(...)（同 message.Hooks）
   装配：Router 先于 Runner 创建，router.Options.PreEvent 传闭包调用之后创建的 Runner；message.New 的 hooks 传同一个 Runner。本切片在本地集成测试中装配，生产装配沿用 T-13 的 Core 装配。
5. sdk（链式注册与 SDK 类型，D-70、D-112）
   Root 增加 Hooks(*HookRegistry)；func (*Base) Hooks(*HookRegistry) 默认不注册
   type HookRegistry struct{ ... }
   func (r *HookRegistry) PreEvent(routeID string, event EventType) *PreEventBuilder // 一条路由对应一种事件，event_type_filter 只含该事件名
   type PreEventBuilder struct{ ... }
   func (b *PreEventBuilder) Priority(p int32) *PreEventBuilder
   func (b *PreEventBuilder) Handle(h any) // func(*Context, P) (P, error)；P 与同一事件的事件路由 Handler 的载荷类型相同（标准事件为 message 包的载荷指针，AdapterEvent[T] 为 *T）；返回 nil 表示不改写
   func (r *HookRegistry) BeforeSend(routeID string) *BeforeSendBuilder
   type BeforeSendBuilder struct{ ... }
   func (b *BeforeSendBuilder) Priority(p int32) *BeforeSendBuilder
   func (b *BeforeSendBuilder) Handle(h func(c *Context, to message.Session, m message.Chain) (message.Chain, error)) // 返回 nil 表示不改写
   func (r *HookRegistry) AfterSend(routeID string) *AfterSendBuilder
   type AfterSendBuilder struct{ ... }
   func (b *AfterSendBuilder) Priority(p int32) *AfterSendBuilder
   func (b *AfterSendBuilder) Handle(h func(c *Context, to message.Session, res *message.SendResult, sendErr error) error) // sendErr 为本次发送的错误；返回的错误只记日志
   func (c *Context) Editing() message.MessageRef // 出站 Hook 中：本次是编辑时为被编辑消息的句柄（D-104），发送时为空；其他 Handler 中为空

行为：
- 注册校验：
  - HOOK 路由的 phase 必须是三者之一。
  - PreEvent 的 event_type_filter 恰好一项，且为已知标准事件或合法的 Adapter 扩展事件名；BeforeSend、AfterSend 不带 event_type_filter。
- 候选：同一阶段的 Hook 按 (priority, plugin_id, route_id) 稳定排序，只包含已激活插件。Hook 不受命令策略约束（插件启用即生效）。
- PreEvent：
  - Router 在 Core 权限计算之后、命令层之前按事件类型依次调用，每个 Hook 收到当前业务副本（可信绑定已由 rebuild 合并，用于读取）。
  - 返回的 event_payload 分支必须与 event_type 一致，没有业务载荷的事件（message.deleted、button.clicked、scheduled）不得返回 event_payload，否则视为该 Hook 失败。可信字段不在 PreEventResult 中，无法改写。
  - 任一 Hook 返回错误、panic 或失败：以 WARN 记录并结束本次入站链（不派发命令与事件路由）。
  - propagation=STOP：结束本次入站链，此前的改写保留。
  - 正常结束时 Router 用最终 payload 替换业务副本，之后的命令与事件路由都收到改写后的载荷。
- 出站：message.Service 对每次 Send 与 Edit（包括 Router 经 Options.Send 代发的自动回复与 usage）依次：
  1) 完成现有的目标补齐、准入与 Chain 预检；
  2) 依次调用 BeforeSend：
     - 返回新 Chain 时替换，并对整条 Chain 重新做完整预检（Edit 还要满足编辑的片段限制），失败视为拒绝；
     - 返回错误即拒绝本次发送：返回 REJECTED 的 SendResult 加该 ErrorInfo，不调用 Adapter，也不执行 AfterSend；
     - Hook 不能修改 Target（请求中没有可回写的 Target）；
  3) 按改写后的 Chain 准备媒体租约与按钮签发并调用 Adapter；
  4) 依次调用 AfterSend，传入最终的 SendResult 与 ErrorInfo；错误只记日志，不改变返回给调用方的结果。
  每次 Send 是独立的一组，不合并。
- send_forbidden：
  - Runner 调用出站 Hook 时，Attach 的 Invocation 设置 SendForbidden=true，并在请求的 InvocationContext.send_forbidden 下发 true。
  - Hook 内（ParentAttach 为该 Invocation）调用 core.send、core.edit、core.action、core.delete 返回 FAILED_PRECONDITION。
  - 标记只来自 Core 的调用记录，插件自报的值被忽略。Broker 转发时继承标记由 T-06.2 负责。
- SDK：
  - 按 route_id 分发 Hook。Handle 的签名在 Collect 时核对，不符或漏写 Handle 时报错，并带 route_id 与期望的签名。
  - PreEvent 的 P 由 SDK 从 event_payload 解码，改写后编码回同一分支（AdapterEvent 的 *T 重新编码为 payload_json）；返回 nil 表示不改写。BeforeSend 返回 nil 表示不改写；AfterSend 返回的错误以 ErrorInfo 回报。
  - to 由请求的 target 转为会话字符串；c.Editing() 在 operation 为 EDIT 时由 target 与 edited_platform_message_id 组成句柄。
  - Context.StopPropagation 在 PreEvent 中使 propagation=STOP；出站 Hook 的 Context 中 Send、Edit 由 Core 拒绝。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86：SDK 与 supervisor 不提交测试；Core 侧由本地集成证明）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - Handle 的错误签名与漏写 Handle 在 Collect 时报错（PreEvent 的 P 与事件不符、BeforeSend 与 AfterSend 的参数不符）。
  - PreEvent 返回不匹配分支的 payload，以及无业务载荷的事件返回 payload（fixture）；BeforeSend 修改后 Chain 预检失败；排序的稳定性；插件自报 send_forbidden 被忽略；core.action、core.delete 在 Hook 父调用中被拒（main 上已有时）。
  - internal/core/supervisor/hook_local_test.go（//go:build unix）：在临时目录生成并编译两个测试插件，经 supervisor.Launch、router.New、Attach、memory Adapter 与真实 message.Service 运行。
    - 插件 h 注册 PreEvent（把 text 改成大写）、BeforeSend（在 Chain 末尾追加一段文本；遇到含 "deny" 的 Chain 返回错误；尝试在 Hook 内 Send）、AfterSend（记录结果后返回错误）。
    - 插件 e 注册 echo 事件路由与一个命令。
    - 覆盖：入站 "hi" 被 PreEvent 改为 "HI" 后到达 e；e 回复时 memory 收到追加了文本的 Chain；AfterSend 收到 ACCEPTED 的结果，它返回的错误不影响 e 收到的 SendResult。
    - 覆盖：一次 Handler 内两次 Send，BeforeSend 与 AfterSend 各执行两次；含 "deny" 的回复被拒，memory 没有收到，e 收到 REJECTED 与错误，AfterSend 未执行。
    - 覆盖：BeforeSend 内的 Send 返回 FAILED_PRECONDITION；PreEvent 返回错误时 e 的路由与命令都没有被调用；Edit 时出站 Hook 的 c.Editing() 为被编辑消息的句柄，Send 时为空。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：Hook→Service→Send 的间接递归回归（T-06.2 的必交用例，本切片只保证标记保存在调用记录中）；延续调用（T-05.4）；生产装配（T-13）。
```

### goal-52 · T-05.4 延续调用与多轮等待

状态：已完成（[PR #153](https://github.com/AlonaBot/AlonaBot/pull/153)，合并提交 `f22b538`），证据见 TEST-MATRIX §3.2。合入：SDK 的 `(*Context).Continue(fn func(*Context) error) error`、`(*Context).Wait(timeout) (*Context, error)`、`(*Context).Text()` 与 `(*EventBuilder).Continue()`；`internal/core/continuation` 的 `New(Options{Max, Timeout, Prepare, Caller})`、`Manager.Register(mux)`，等待按 Caller 的当前绑定匹配，`Release(c router.Caller)` 在 Session 结束时释放该 Session 的等待；`router.Options.Continuations`（`Finished`、`Match`、`Release`）与 `Router.CallContinuation(ctx, inv, continuationOf, deadline)`；proto 的 `HandlerRequest.continuation_of`、`ContinuationStartRequest`/`Response`、`WaitRequest{timeout_ms}`/`WaitResponse{event}`。审查时接受 `sdk/sdk.go` 中 Context 字段的改动（范围修订）。派发时：前置 T-05.2（[PR #145](https://github.com/AlonaBot/AlonaBot/pull/145)，合并提交 `0f860a9`）已合并：SDK 的 Rule 与目标 Handler 在同一调用、同一 Context 中执行，`StopPropagation` 锁存，调用结束时 Context 失效；Core 的自动回复排在显式发送之后。2026-11-12 审查者按 main `3e4db62` 的代码改写接口段，要点如下：
- `router.Invocation` 现有 PluginID、Event、DefaultTarget 三个字段，由 `Router.call` 作为 `wire.Request.Attach` 传出，插件以该调用为父请求时，服务函数从 `wire.Inbound.ParentAttach` 取回同一指针。
- 定时任务与 Service 调用的 ParentAttach 不是 `*router.Invocation`。
- 没有 `router.Dispatch` 类型。router 不能 import continuation，所以改为由 router 声明接口、在 Options 中注入。
- `CONTINUATION_MAX` 与 `CONTINUATION_TIMEOUT` 已在 `internal/tune`（`Values.ContinuationMax`、`ContinuationTimeout`）。
- `HandlerRequest` 现用到字段 4。
- SDK 的路由 builder 为 `*CommandBuilder`（Command、Regex、Keyword、StartsWith、EndsWith、FullMatch、Raw）、`*EventBuilder`（On）与 `*ButtonBuilder[T]`（OnButton），都已有 `Permission`、`Rule` 与 `Handle`；`message.Chain.Text()` 已存在。

与 goal-17（T-05.3）并行时，后合并者按「PreEvent → 等待匹配 → 命令层」的顺序接入两者的调用点，父调用带 send_forbidden 时的拒绝以 goal-17 合入的标记为准。与 goal-125 并行时，后合并者保留 builder 的 `Continue` 方法，并迁到插件对象返回的同一 builder 上。与 goal-109 同改 `sdk/sdk.go` 与 `sdk/session.go`，后合并者同步 main 并保留对方的接口。

```text
/goal 使用 $alonabot-execute 推进 goal-52（切片 T-05.4：延续调用与多轮等待）。goal 目录 goal-52，分支 goal-52，worktree .worktrees/goal-52，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：由真实入站事件触发、具有 Actor 与默认 Target 的 Handler 可以经 core.continuation.start 请求一次延续调用。原调用成功结束后，Core 以同一插件 Session、同一 route_id 与原 HandlerRequest 发出新的 Handler 请求（continuation_of 为原 call_id）。该请求使用独立的 CONTINUATION_TIMEOUT 期限，不占用分发链。延续调用内可以经 core.wait 等待同一 Actor 在同一 Target 的下一条消息，命中的消息不进入普通分发。SDK 提供 Context.Continue、Context.Wait，以及路由 builder 的 Continue（D-119）。

依据：PLAN T-05 步骤 7；CONTRACTS §2.4；WIRE-FORMAT §3.5；DESIGN（延续调用示例）；DECISIONS D-119；TEST-MATRIX V-50、V-113 的延续调用分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/invocation.proto：HandlerRequest 追加 optional uint64 continuation_of = 5；新增 ContinuationStartRequest/Response 与 WaitRequest/Response。随后运行 go generate ./...。本切片获授权只追加这些。
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceContinuationStart = "core.continuation.start"、ServiceWait = "core.wait"
- 新增 internal/core/continuation/
- internal/core/router/：
  - mux.go：Invocation 追加下列字段；
  - dispatch.go：Options 追加 Continuations；
  - router.go：New 保存 Continuations；dispatch 在命令层之前调用 Match；call 在 Handler 调用返回后调用 Finished，并为 CallContinuation 保存原 HandlerRequest；Detach 调用 Release；新增 CallContinuation。
  - 不改其他派发逻辑。
- sdk/：新增 continuation.go（Continue、Wait、Text）；registry.go、events.go、buttons.go 的 builder 各加 Continue；dispatch.go 处理 continuation_of 与 builder 的延续路由
- 上述包的 *_local_test.go，以及 internal/core/supervisor/continuation_local_test.go（//go:build unix）
允许的新跨模块边：无（continuation 与 router 同属 internal/core 模块；router 不 import continuation）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message ContinuationStartRequest {}
   message ContinuationStartResponse {}
   message WaitRequest { uint32 timeout_ms = 1; } // 1～CONTINUATION_TIMEOUT 的毫秒数
   message WaitResponse { Event event = 1; }
2. internal/core/router
   type Invocation struct {
       PluginID      string
       Event         *pb.Event
       DefaultTarget *pb.Target
       RouteID       string // 被调用的 route_id
       Continuation  bool   // 本调用为延续调用
       // 未导出字段：所属候选（插件 Session）与原 HandlerRequest，供 CallContinuation 使用
   }
   type Continuations interface {
       Finished(inv *Invocation, ok bool) // call 在 Handler 调用返回后调用；ok 为没有传输错误且 Response.Err 为 nil
       Match(ev *pb.Event) bool           // dispatch 在命令层之前对 message.created 调用；true 时事件不再进入命令层与事件路由
       Release(pluginID string)           // Detach 时调用
   }
   Options 追加 Continuations Continuations // nil 时不检查等待，也不调用 Finished
   func (r *Router) CallContinuation(ctx context.Context, inv *Invocation, continuationOf uint64, deadline time.Time) (wire.Response, error)
   // 以 inv 所属的插件 Session、route_id 与原 HandlerRequest（含 command 参数与已为该 Session 准备的事件）重发，HandlerRequest.continuation_of=continuationOf，Attach 为 Continuation=true 的新 Invocation，期限为 deadline；inv 所属 Session 已不是当前激活绑定时返回错误；不经事件队列，不计入 DISPATCH_CONCURRENCY，返回值不参与原事件的传播
3. internal/core/continuation（*Manager 实现 router.Continuations）
   type Options struct {
       Max     int           // CONTINUATION_MAX：每个插件 Session 已登记待发出与正在运行的延续调用总数上限
       Timeout time.Duration // CONTINUATION_TIMEOUT
       Call    func(ctx context.Context, inv *router.Invocation, continuationOf uint64, deadline time.Time) (wire.Response, error) // 装配处传 Router.CallContinuation
       Prepare func(*pb.Event, router.Caller) (*pb.Event, error) // 为等待者的 Session 准备命中事件，与 router.Options.Prepare 相同；nil 时原样返回
   }
   func New(opt Options) *Manager
   func (m *Manager) Register(mux *router.Mux) // 注册 core.continuation.start 与 core.wait
   func (m *Manager) Finished(inv *router.Invocation, ok bool)
   func (m *Manager) Match(ev *pb.Event) bool
   func (m *Manager) Release(pluginID string)
   装配：先 New(Options{Call: 闭包调用之后创建的 Router.CallContinuation})，再把它作为 router.Options.Continuations 传入；本切片在本地集成测试中装配，生产装配沿用 T-13 的 Core 装配。
4. sdk
   func (c *Context) Continue(fn func(*Context) error) error // 只能在 Handler 中调用一次；fn 在延续调用中执行
   func (b *CommandBuilder) Continue() *CommandBuilder       // 命令、文本匹配与 Raw 路由：Handler 整体在延续调用中执行（D-119）
   func (b *EventBuilder) Continue() *EventBuilder           // 事件路由同上
   func (b *ButtonBuilder[T]) Continue() *ButtonBuilder[T]   // 按钮路由同上
   func (c *Context) Wait(timeout time.Duration) (*Context, error) // 只能在延续调用中调用；返回命中消息的 Context（可用 Event()、Text() 等）
   func (c *Context) Text() string // message.created 时为消息 Chain 的 Text()；其他事件为空

行为：
- core.continuation.start（必须带 parent_call_id，指向本 Session 中仍在途的 Core→插件 Handler 调用）：
  - 以下情况返回 FAILED_PRECONDITION：in.ParentAttach 不是 *router.Invocation（scheduled、Bus、Service 等）；缺少 Actor 或默认 Target；父调用本身是延续调用（Continuation 为 true）；同一父调用已请求过。父调用带 send_forbidden 也拒绝，标记由 T-05.3 接入（见状态行）。
  - 调用方 Session 进行中的延续调用达到 Max 时，返回 RESOURCE_EXHAUSTED。
  - 成功时按父 Invocation 指针登记调用方 Caller 与 continuation_of（in.ParentCallID），返回空结果。
- 发出：
  - Finished(ok=true) 时，若该 Invocation 有登记，在新的 goroutine 中经 Options.Call 发出延续调用，期限为发出时刻起 Timeout，continuation_of 为登记的原 call_id。
  - Caller 已不是该插件的当前 Session 时丢弃登记。返回错误只记日志（不含参数值），不重试。
  - Finished(ok=false) 时丢弃登记，不发出。
- core.wait（必须带 parent_call_id，父 Invocation 的 Continuation 为 true）：
  - 等待键为（adapter_instance_id、chat_id、thread_id（无则为空）、Actor 的 external_user_id），取自父 Invocation 的默认 Target 与 Actor。
  - 同键已有等待时返回 ALREADY_EXISTS；timeout_ms 不在 1～Timeout 内时返回 INVALID_ARGUMENT。
  - 等待保存在 Core 内存中，不持久化。
- 匹配：Router 对每个 message.created 在命令层之前调用 Match。Target 与 Actor 构成的键命中等待时，Match 摘除该等待，事件经 Options.Prepare 为等待者的 Session 准备后作为 WaitResponse 返回，并返回 true；该事件不再进入命令层与事件路由。不同 Actor 或不同 thread 不命中。
- 结束：
  - 超时返回 DEADLINE_EXCEEDED；core.wait 的请求被取消（包括作为父请求的延续调用结束或期限到达）时返回 CANCELLED。
  - Release（插件停用或 Session 失效）时，该插件的等待以 UNAVAILABLE 结束，待发出的延续调用丢弃，不补发。
- builder 的 Continue（D-119）：
  - 原调用中 SDK 照常执行 Rule。匹配后请求 core.continuation.start，原调用以 matched、Continue 成功结束，不调用 Handler；请求失败时原调用返回该错误。
  - 收到 continuation_of 的请求时调用该路由的 Handler，参数与原调用相同；自动回复型 Handler 返回的非空内容由 SDK 在延续调用结束前经 Send 发出一次。
  - 这类路由中 StopPropagation 不起作用，再调用 c.Continue 返回 FAILED_PRECONDITION。
- SDK 的 Context.Continue：
  - 在 Handler 返回前请求 core.continuation.start（ParentCallID 为本次调用的 call_id），并按该 call_id 登记 fn；收到 continuation_of 的请求时调用对应 fn，找不到时返回 NOT_FOUND。
  - 在延续调用中调用 Continue 返回 FAILED_PRECONDITION。
  - Wait 以本次延续调用为父请求发起 core.wait。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86：SDK 与 supervisor 不提交测试；Core 侧由本地集成证明）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - 用 `.Continue()` 注册的 /quiz2 与 Handler 内调用的 /quiz 行为相同；Rule 不匹配时不请求延续；自动回复型的返回内容只发出一次；Handler 中的 StopPropagation 不改变原调用的传播。
  - 原调用返回错误时不发出；ParentAttach 不是 Invocation 时被拒；timeout_ms 边界；Release 与超时的并发；SDK 中按 continuation_of 找不到 fn 时返回 NOT_FOUND。
  - internal/core/supervisor/continuation_local_test.go（//go:build unix）：在临时目录生成并编译真实 SDK 插件，经 supervisor.Launch、router.New、Attach、memory Adapter 与真实 message.Service 运行。测试插件注册 /quiz 命令：Continue 中依次 Wait 两次，并把收到的回答回复出来；另有一个 echo 事件路由，用来证明消息是否进入普通分发。覆盖：
    - 用户 A 发 /quiz 后，A 的下两条消息被 quiz 收到并回答，没有到达 echo；原命令的 HandlerOutcome 决定的传播不受影响。
    - 同一群里 A 与 B 同时 /quiz，各自只收到自己的回答；不同 thread 的消息不命中。
    - Wait 超时返回 DEADLINE_EXCEEDED；同键第二个 Wait 返回 ALREADY_EXISTS；延续调用中再次 Continue 返回 FAILED_PRECONDITION。
    - 没有 Actor 的事件中请求 Continue 返回 FAILED_PRECONDITION；用很小的 Max 时，超过的 Continue 返回 RESOURCE_EXHAUSTED。
    - 等待期间停止插件，等待结束，重启后不补发；延续调用执行期间，其他事件照常派发（不占分发链）。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：持久化的等待与延续调用（契约不提供）；Agent 的流式对话（T-12.2 使用本切片）；生产装配（T-13）。
```

### goal-127 · T-07.8 动态任务与调度持久化的协议和 Core 侧

状态：已完成（[PR #169](https://github.com/AlonaBot/AlonaBot/pull/169)，合并提交 `4dc4bcd`），证据见 TEST-MATRIX §3.2。合入：服务 `core.task.set`、`core.task.cancel`、`core.task.list`、`core.job.pause`、`core.job.resume`、`core.job.status` 与 proto `TaskTrigger`、`TaskSetRequest/Response`、`TaskCancelRequest/Response`、`TaskListRequest/Response`、`TaskInstance`、`JobPauseRequest/Response`、`JobResumeRequest/Response`、`JobStatusRequest/Response`、`TaskSpec`（`ROUTE_KIND_TASK`）；状态库 `dynamic_task` 与 `scheduled_job`；tune `TASKS_PER_PLUGIN`、`TASK_PAYLOAD_BYTES`、`TASK_CONCURRENCY`；`wire.Request.Late` 在准入前期限已过时同样调用一次；根包任务服务等待本 Session 装配完成后才处理请求。SDK 的作者 API 由 goal-128 提供。

按 main 核对（2026-11-12，已用 rg 核对定义及调用者，并以离线 go doc 交叉核对 Scheduler.Options、manage.Service.Handler、state.Store.Uninstall）：

- `internal/core/scheduler/scheduler.go`：`func New(opt Options) *Scheduler`；当前 `Options` 只有 `JobTimeout time.Duration`、`Now func() time.Time`；`func (s *Scheduler) Attach(p router.Plugin, activatedAt time.Time)` 当前无返回值。`Drain(pluginID string)`、`Detach(pluginID string)`、`Released(pluginID string)`、`Close()`、`Stats(pluginID, jobID string) JobStats` 已存在；`JobStats` 为 `Runs, Skipped, TimedOut, Late uint64; Running bool`。当前定时器和起点均在内存中，运行槽按 plugin_id/job_id 保存。
- `contracts/schedule/schedule.go`：`Interval(d time.Duration) (Trigger, error)`、`Cron(expr, timezone string) (Trigger, error)`、`FromSpec(s *pb.ScheduleSpec) (Trigger, error)`、`func (t Trigger) Next(anchor, now time.Time) time.Time` 已存在；本切片复用其 interval、cron、IANA 与未来 tick 计算。
- `internal/core/router/mux.go`：`type ServiceFunc func(ctx context.Context, c Caller, in *wire.Inbound) wire.Response`；`(*Mux).Handle(name string, fn ServiceFunc)`、`(*Mux).For(c Caller) wire.Handler`；`Caller{PluginID, GenerationID, SessionKey string}`。调用身份由装配绑定。`internal/core/router/router.go` 的 `Plugin` 提供 `PluginID() string`、`Activated() bool`、`Routes() []*pb.RouteDescriptor`、`Call(ctx context.Context, req wire.Request) (wire.Response, error)` 等方法；`func (r *Router) PluginFor(c Caller) (Plugin, bool)` 核对 Session 身份。
- `assembly.go`：`assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error)` 在一处构造 Scheduler；`(*assembly).attach(ctx context.Context, r *runningComponent) error` 先接 Router，再为同进程组件 Attach。`internal/core/supervisor/supervisor.go` 的 `Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error)` 在 `host.Activate` 成功后为独立进程 Attach；冻结、撤销及确认进程退出分别调用 Drain、Detach、Released。两条 Attach 路径都须同步处理持久化错误。
- `internal/core/state/installs.go`：私有 `cleanupObjects(db *gorm.DB, id string, all bool) ([]cleanupObject, error)` 同时供 `PurgeObjects(ctx context.Context, id string, all bool) ([]string, error)` 和 `Uninstall(ctx context.Context, id string, data, all bool) error` 使用，后者在同一个事务内删除对象。`internal/core/manage/lifecycle.go` 已有 `PurgeNone`、`PurgeData`、`PurgeAll`；`(*Service).Uninstall(ctx context.Context, id string, purge PurgeScope, confirm bool) error` 先预览、确认和停止，再调 Store；无需另建清理回调。
- `contracts/proto/alonabot/v1/route.proto`：RouteKind 用到 7，RouteDescriptor.spec 用到 16，公共 `required_permissions = 17` 已占用；TaskSpec 可用字段 18。`message.proto` 的 ScheduledPayload 当前只有 `job_id = 1`、`run_id = 2`、`scheduled_at_unix_ms = 3`。`contracts/validate/register.go` 的 `Register(r *pb.Register) error` 原子检查注册，静态路由前缀为 `job.`。
- `internal/tune/params.go` 的 `Values` 和 `parameters` 是现有可调上限登记处；TASK 三项尚未登记；`contracts/limits/` 尚不存在。`internal/wire/services.go` 尚无六个 task/job 服务常量。
- 根包 `pipeline_test.go` 已有 `newPipeline(t *testing.T) *pipelineHarness`、`(*pipelineHarness).launch(id, mode string, priority int) *runningComponent` 与测试二进制子进程装配。SDK 尚无 Task API，本切片的动态任务验证使用原始 wire/pb；goal-128 负责作者 API。
- 在途冲突：goal-55、goal-131 同改 `internal/core/state/schema.sql`，goal-131 还涉及清理对象；goal-22、goal-55、goal-69、goal-81 与本切片共用 `assembly.go`，goal-22 可改 `pipeline_test.go`，goal-81 可改 Router，goal-69 改根包与 `cmd/alonabot`；goal-59 的产品代码在 `agent/`，与本切片无直接重叠。`contracts/wirespec_test.go`、生成目录与 `internal/wire/services.go` 也可能由 goal-55/goal-131 追加。后合并者保留各自对象、服务、装配点与编号，再统一生成；本切片无需修改 cmd 或 Agent。

- 审查者决定（2026-11-12）：可调上限沿用 internal/tune 登记；采纳本语句的协议字段号、表结构与任务名格式（D-124～D-126 约定由审查者提出）；Cancel 匹配持久实例或仍在途的同一实例（派发后的 at 在途时也可取消），两者都不存在时 NOT_FOUND，替换不取消在途的旧一轮；动态任务的 JobTimeout 从实际派发时开始计算，scheduled_at 仍为原到期时刻；JobStatus 的 last_at 为本次 Core 运行中最后一次实际派发的时刻，Core 重启后缺失，暂停时 next_at 缺失；既有 supervisor_test.go 只同步构造参数。
```text
/goal 使用 $alonabot-execute 推进 goal-127（切片 T-07.8：动态任务与调度持久化的协议和 Core 侧）。goal 目录 goal-127，分支 goal-127，worktree .worktrees/goal-127，从最新 origin/main 创建（包含 PR #119、PR #163）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 支持运行时创建、整体替换、取消与分页列出带参数的任务实例，并支持静态任务的暂停、恢复与状态查询。动态定义、interval 起点与暂停状态保存在 Core 状态库，跨 Core/插件重启与更新保留；错过的触发作废。按管理服务已有的三种卸载范围保留或事务删除这些数据。SDK 的作者 API 由 goal-128 提供。

依据：spec/PLAN.md §3.1；spec/plan/core.md T-07.8、T-07 步骤 2、3、7、9 及「停止和失败条件」，T-09 步骤 3 的清理语义；spec/DESIGN.md §H；spec/CONTRACTS-v0.1.md §1.2、§4.3、§9.3、§14.5、§16；spec/MESSAGE-MODEL.md §1 的 scheduled 来源与无 Actor、§3 的 HandlerOutcome；spec/WIRE-FORMAT.md §3.3、§3.5；spec/adr/DECISIONS.md D-124～D-127、D-86；spec/TEST-MATRIX.md V-114 的 Core 分项，并回归 V-20、V-21、V-46 的运行槽与启停分项。工作包在 spec/plan/core.md。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/route.proto、message.proto；新增 task.proto：下列声明、字段与服务载荷。授权只追加接口段列出的内容；运行 go generate ./... 更新 contracts/pb/alonabotv1/，生成产物按实际差异保留。
- contracts/validate/register.go：TASK 注册、任务名与 task. 前缀校验；contracts/wirespec_test.go：追加本次协议编号检查。
- internal/wire/services.go：末尾追加六个服务名常量。
- internal/tune/params.go 与 parse.go：登记三项可调上限与交叉约束，沿用现有调优入口（PLAN 中的 contracts/limits 已改为 internal/tune）。
- internal/core/state/schema.sql、installs.go；新增 tasks.go、jobs.go：唯一初始迁移中的两张表、GORM 读写和清理清单。各模型与方法放在所属主题文件。
- internal/core/scheduler/scheduler.go；可新增 tasks.go、services.go、jobs.go：持久化 Attach、六个固定服务、动态运行槽/队列、统计与生命周期。
- assembly.go：传入同一 Store 与生效上限、注册固定服务、处理同进程 Attach 错误；internal/core/supervisor/supervisor.go：处理独立进程 Attach 错误并走已有失败清理。
- 根包 pipeline_test.go：仅保留原断言的装配同步和测试子进程模式分流；新增 scheduler_tasks_test.go（unix）的一项集成测试。SDK 与 supervisor 目录不新增提交的测试；既有 internal/core/supervisor/supervisor_test.go 只同步 Scheduler 的构造参数，断言不变（审查者授权）。
- 上述相关包的 *_local_test.go：本地验证，不提交。

允许的新跨模块边：无。scheduler -> state/router 属于 internal/core 内部；Core -> contracts/internal/wire/internal/tune 已允许。contracts 与 internal/tune 保持现有依赖方向。生产 Core 保持通过协议调用组件。

接口（照此实现；字段名、任务名格式与分页细节已由审查者采纳，内部辅助函数自定）：
1. route.proto：
   enum RouteKind 追加 ROUTE_KIND_TASK = 8;
   message TaskSpec { string name = 1; }
   RouteDescriptor.spec 追加 TaskSpec task = 18;
   required_permissions = 17 和 reserved 5 保持原编号。
2. message.proto 的 ScheduledPayload 追加：
   string task_name = 4;
   string task_key = 5;
   string task_payload_json = 6;
   动态 job_id/route_id 为 task.<name>；静态触发的三个新增字段为空。
3. 新增 task.proto（均为 protobuf，schema_id 用 validate.SchemaID，schema_version=1）：
   message TaskTrigger {
     oneof trigger { int64 at_unix_ms = 1; uint64 interval_ms = 2; string cron = 3; }
     string timezone = 4;
   }
   message TaskSetRequest { string task_name = 1; string key = 2; TaskTrigger trigger = 3; string payload_json = 4; }
   message TaskSetResponse { int64 next_at_unix_ms = 1; }
   message TaskCancelRequest { string task_name = 1; string key = 2; }
   message TaskCancelResponse {}
   message TaskListRequest { string task_name = 1; optional string after_key = 2; }
   message TaskInstance { string key = 1; TaskTrigger trigger = 2; string payload_json = 3; optional int64 next_at_unix_ms = 4; }
   message TaskListResponse { repeated TaskInstance instances = 1; optional string next_after_key = 2; }
   message JobPauseRequest { string job_id = 1; }
   message JobPauseResponse {}
   message JobResumeRequest { string job_id = 1; }
   message JobResumeResponse {}
   message JobStatusRequest { string job_id = 1; }
   message JobStatusResponse { bool paused = 1; optional int64 last_at_unix_ms = 2; optional int64 next_at_unix_ms = 3; }
   next_at 的缺失表示当前无已安排的触发；last_at 的缺失表示本次 Core 运行中尚未派发。静态 job_id 为 ScheduleSpec.job_id。
4. internal/wire 追加常量：ServiceTaskSet="core.task.set"、ServiceTaskCancel="core.task.cancel"、ServiceTaskList="core.task.list"、ServiceJobPause="core.job.pause"、ServiceJobResume="core.job.resume"、ServiceJobStatus="core.job.status"。
5. internal/tune.Values 追加 TasksPerPlugin int、TaskPayloadBytes int64、TaskConcurrency int；登记 TASKS_PER_PLUGIN（10000，>=1）、TASK_PAYLOAD_BYTES（64KiB，1～4MiB）、TASK_CONCURRENCY（16，1～SESSION_MAX_INFLIGHT）。沿现有参数表/解析器校验和传递；全部构造方传入生效值。
6. internal/core/state（方法返回原始存储错误；事务内用 GORM）：
   type Task struct {
       PluginID, Name, Key string
       At time.Time; Interval time.Duration; Cron, Timezone string
       PayloadJSON string; Anchor, CreatedAt time.Time
   }
   type Job struct { PluginID, JobID string; Anchor time.Time; Interval time.Duration; Paused bool }
   func (s *Store) Tasks(ctx context.Context, pluginID, name string) ([]Task, error) // name 为空读取该插件全部实例
   func (s *Store) SetTask(ctx context.Context, task Task, maxTasks int) error // 原子检查该插件总数后按复合主键替换；替换不增加配额
   func (s *Store) DeleteTask(ctx context.Context, pluginID, name, key string) error // 无行返回 state.ErrNotFound；服务层另核对在途实例
   func (s *Store) Jobs(ctx context.Context, pluginID string) ([]Job, error)
   func (s *Store) SaveJob(ctx context.Context, job Job) error
   dynamic_task 以 (plugin_id,name,key) 为主键；scheduled_job 以 (plugin_id,job_id) 为主键。cron 的 Job 行同样保存暂停标记，Anchor/Interval 为零；起点仅用于 interval。以安装身份关联，普通卸载保留记录；清理经 cleanupObjects 在既有事务内完成。
7. internal/core/scheduler：
   Options 保留 JobTimeout、Now，追加必填 Store *state.Store、TasksPerPlugin int、TaskPayloadBytes int64、TaskConcurrency int。
   func New(opt Options) *Scheduler // 签名保持；所有构造点同步新字段
   func (s *Scheduler) Attach(ctx context.Context, p router.Plugin, activatedAt time.Time) error // 替换原 Attach；数据库失败向启动方返回
   func (s *Scheduler) Services(m *router.Mux, current func(router.Caller) (router.Plugin, bool)) // 注册六个服务；assembly 传 a.router.PluginFor
   保持 Drain/Detach/Released/Close/Stats 的签名和既有计数含义。
   type TaskStats struct { Expired uint64; Unregistered uint64; Running, Queued int }
   func (s *Scheduler) TasksStats(pluginID string) TaskStats // 内存诊断快照；本切片无管理 HTTP API
   动态运行槽的稳定键为 (plugin_id,name,key)，与静态 (plugin_id,job_id) 分开。

行为：
- REGISTER 原子拒绝错误的 kind/spec、空/非法任务名、重复任务名、任务路由 ID 与 task.<name> 不一致，以及其他路由占用 task. 前缀。任务名为 ^[a-z][a-z0-9_.-]{0,63}$。同名实例靠 key 区分。保留既有权限名集合的校验；TASK 触发只具有 job scope，事件及可信父附件均无用户 Actor、默认 Target 或默认 Admin。
- Attach 在安排定时器前从同一个状态库读取两类状态并完成必要写入：静态 interval 无起点或间隔变化时写 activatedAt，新间隔重置起点并保留暂停状态；未变化时用旧起点。静态 cron 保留暂停状态。全部读取/写入成功后发布本 Session 的注册和未来定时器；数据库失败走启动失败清理，错误保持可见。
- 动态 interval 的 Anchor 为创建/替换时刻，cron 复用 contracts/schedule，at 为一次性绝对时刻。Attach 清除已到期的 at，WARN 含 plugin_id、task_name、key、scheduled_at 并计 Expired；先处理过期 at，再安排已注册任务名的未来实例。当前 Session 缺少任务名的其余行保留，计 Unregistered。重启、更新与 rollback 均以 Core 状态库的任务为准。
- 服务入口以 Mux Caller 为身份，经 current 核对当前 Session 且处于 Activated；旧 Session 与未激活调用返回 FAILED_PRECONDITION，Draining 后拒绝新的任务变更，包括仍带父调用的请求。无需 parent_call_id。请求载荷无 plugin_id 字段。
- set 先完整校验，再写库：key 是 1～128 字节合法 UTF-8；三种触发恰有一种；interval >=1s、五字段 cron、有效 IANA、cron 缺省 UTC；timezone 只用于 cron；at 必须晚于 Now；参数为原样 JSON 对象且不超 TaskPayloadBytes。当前未注册任务名返回 NOT_FOUND；非法输入/过去的 at/参数超限返回 INVALID_ARGUMENT；新 key 超总数返回 RESOURCE_EXHAUSTED。相同 key 整体替换触发与参数，interval 起点重置。写入失败保持旧定时器、队列和数据，返回完整存储错误经协议安全错误通道处理。
- cancel 对该插件的实例撤销后续触发、定时器和排队项，取消其在途 wire 调用；删除持久行成功后才改内存。实例已在派发时删行但仍在途时，仍应成功取消该次调用；持久与在途记录均不存在才返回 NOT_FOUND。槽一直保留到真实 Response 或确认旧进程退出。替换后旧在途响应仅释放自身运行记录，禁止删除新实例或释放新调用的槽。
- list 按同一任务名、key 字典序返回完整参数与规则；after_key 为上一页最后返回的 key，下一页严格从其后开始。按实际 protobuf/Envelope 大小限制分页，完整保留单条数据；超帧的单条返回 RESOURCE_EXHAUSTED。next_after_key 仅有后续条目时设置。无并发写入时分页不丢不重；当前缺少处理函数的已保存任务也可列出和取消，其 next_at 缺失。任务名对应的零实例返回空页。
- tick 仍使用真实 REQUEST/RESPONSE、独立 run_id 与 JobTimeout；静态任务保持原期限和实际运行槽语义。动态同一实例有未结束调用时跳过本轮；每插件动态实际运行数受 TaskConcurrency 限制，其余按到期时刻排队。排队周期实例只保留一项；等待名额期间触发仍有效。一次性实例在取得名额、实际派发前的状态库事务中删除，删除失败禁止派发。排队动态调用期限拟从实际派发时刻计算，详见备注。
- 超时/协作取消只结束调用结果，实际槽及动态并发名额继续保留；真实迟到 Response 或 Released 才释放并推进仍有效的队列。失败与超时无业务重试。落后时只安排下一未来 tick。
- pause/resume/status 只操作已注册的静态 job_id，缺失返回 NOT_FOUND。pause 先保存标记，再撤销未来定时器，在途一轮继续；resume 先保存标记，再按原起点安排下一未来时刻。重复操作成功且保持状态与起点。status 的 last_at/next_at 语义按备注定稿。
- Drain、Detach、Close 摘除该 Session 全部静态/动态定时器，清空到期队列，持久定义保留；STOPPING 沿既有停止路径取消在途调用，共用全局停止预算。Session 替换后旧定时器、队列项与完成回调均按所属注册/运行身份失效。
- 将 dynamic_task、scheduled_job 添加到 state.cleanupObjects 的数据对象部分，预览和删除用同一清单。PurgeData/PurgeAll 在既有 Uninstall 事务中按 plugin_id 删除，并进入管理服务 ErrNeedsConfirm 的 affected；无需插件进程在线。PurgeNone 保留行；停止完成后无 Session 可再创建任务。同来源重装，周期任务从原起点的下一未来时刻继续，期间到期的 at 作废；PurgeData 保留配置/secret，PurgeAll 沿用全部配置清理。
- 内部存储、定时器与服务写操作采用同一串行顺序，覆盖 set/cancel 与到期派发、暂停和停止的交错；事务中只操作状态，不等待 IPC。数据库错误有返回/日志，参数 JSON 与 secret 不写日志。

测试（按编码规范「测试提交规则」）：
- 提交：只新增根包 scheduler_tasks_test.go 的 TestPersistentTasks（//go:build unix），测试子进程以原始 wire/pb 完成新 TASK 注册与服务请求，复用 pipeline_test.go 的生产装配、真实 Session、独立进程与临时 SQLite。主链：创建 interval/at、同 key 替换与 list、触发参数和无 Actor；重新打开同一数据目录后起点及任务保留、静态暂停保持、恢复从未来 tick 继续；已进入 Handler 的实例经 cancel 收到 Cancel；经真实 manage.Service 的三种卸载范围与同来源重装检查定义的去留和后续触发。主链可用同一测试的子场景，具体平台证据单列。
- 同步 contracts/wirespec_test.go 的编号断言。多列理由：公开协议追加需要固定编号与 presence，这是编码规范允许的协议编号检查。
- 范围内既有测试只同步新签名与必填装配字段，断言不变。SDK 与 supervisor 目录的测试、新 testdata 插件全部放本地；提交测试代码中的辅助进程直接在现有测试二进制内实现，不引入提交的 SDK Task 示例。遇到范围外既有 fixture 必须同步时标 [!]，交审查者处理。
- 本地自测（*_local_test.go，PR 写明结果）：注入 Now 验证固定起点、间隔变化、cron/夏令时、落后跳过、暂停跨重启；过期 at 的删除/WARN/计数，缺少任务名时的保留/诊断；输入错误码、配额替换、字节上限与跨帧完整分页；同插件超过并发名额后全部依次执行、周期排队合并与不同实例独立；超时/断链保槽和迟到响应；已派发 at 的 cancel、同 key 替换交错、Draining 清队列及旧 Session 变更拒绝；数据库失败不派发/不发布半注册，purge 清单准确和同事务回滚。
- 本地多进程补充：Core/插件停用越过多个 tick、插件更新缺少任务名、卸载时有排队/在途调用、重新安装的原起点恢复。原始协议进程保留参数和实际完成通知；其余 SDK API 验收交 goal-128。

完成条件：接口、持久化、实际槽、取消、暂停和三种清理语义均有证据；本草稿备注中的契约决定已由审查者写入正式语句/权威规格。执行阶段运行 go generate ./... 并核对差异仅为上述授权追加；gofmt -l 本次修改的 Go 文件无输出；go vet ./...；go test -race . ./contracts/... ./internal/core/... ./sdk/...；go build ./...。PR 记录命令、平台、Go 版本、提交测试新增行数及本地自测结果，V-114 的证据交审查者回写 TEST-MATRIX §3.2。出现工作包停止条件、范围外变更、新依赖或未决契约冲突时标 [!]。

不做：sdk.Task/OnTask、sdk.Key/Timezone、SDK 事件自有类型转换和作者示例（goal-128）；任务管理界面/API；动态任务暂停；错过触发的补执行；持久化到期队列；候选/rollback 改写 Core 任务；新增状态库迁移和兼容层；新依赖。
```

### goal-130 · T-08.14 配置项的选项来源

状态：已完成（[PR #157](https://github.com/AlonaBot/AlonaBot/pull/157)，合并提交 `e709135`），证据见 TEST-MATRIX §3.2。合入：`configschema.BuiltinOptionSources`、`Schema.OptionsFrom`；`ServiceSpec.options_source = 6`；`validate.OptionsQuerySchemaID`/`OptionsPageSchemaID`；tune 的 `OPTIONS_PAGE`；`sdk.ServeOptions(r, serviceName, fn)` 与 `OptionsQuery`、`OptionsPage`、`Option`（自有服务只声明名称，归属在注册时确定）；`internal/core/options` 的 `New(Options{Schema, Router, Broker, Instances, Page})` 与 `Query`；`Broker.CallSession(ctx, c, target, in)` 与 `Router.PluginFor(c)`；根包 assemble 加载 `settings.File` 并按 Session 发布选项 schema。范围修订：`@agent` 来源的解析归 T-12.8（goal-83），之前返回 UNAVAILABLE。派发时：前置均已合并：T-04.12（[PR #154](https://github.com/AlonaBot/AlonaBot/pull/154)，合并提交 `43a6135`）、T-06.1（PR #45）、T-08.11（PR #103）。2026-11-12 审查者按 main `2828864` 的代码写出范围、接口与测试段，现有代码如下：

- **配置 schema**：字段属性在 `contracts/configschema/declaration.go` 的属性 switch 中逐项校验。`options`、`labels` 只用于 string 与没有 items 的 list；secret 不能与 default、options 并存。
- **ServiceSpec**：现用到字段 5（tool）；`ServiceAccess_SERVICE_ACCESS_PRIVATE` 表示只允许 Core 与本插件。
- **Service 转发**：`broker.Broker.Call(ctx, c router.Caller, in *wire.Inbound)` 按目标插件当前 Active Session 的路由定位 Service，校验访问规则（`router.ServiceCallable`）与输入的编码和 Schema 后转发。
- **组件的配置 schema**：库模式装配按组件取得。开发组件来自 `p.Generated()`，已安装组件来自 `packageResources(dir)`，Adapter 来自 `Generated.AdapterConfigSchema`。
- **tune**：`internal/tune` 尚无 `OPTIONS_PAGE`。
- **并行切片**：goal-126（T-04.13）正在改写 `sdk/services.go` 的 Service 声明，本切片的 SDK 代码放在新文件中，复用其内部注册辅助，后合并者同步。

本切片修改 Service 描述的 proto 与 `sdk/generate` 的 schema 检查，按执行顺序的共享文件规则处理冲突。

```text
/goal 使用 $alonabot-execute 推进 goal-130（切片 T-08.14：配置项的选项来源）。goal 目录 goal-130，分支 goal-130，worktree .worktrees/goal-130，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：`config_schema.json` 的字段可以用 `options_from` 声明选项来自 Core 内置来源，或来自某个组件的选项来源服务；组件用 `sdk.ServeOptions` 提供来源；Core 能按组件与字段路径查询选项。本切片提供 schema 检查、协议标记、SDK 注册、Core 的查询入口与内置来源 `core.adapter_instances`；管理 API 由 goal-93 提供，页面由 goal-44 提供。

依据：PLAN T-08 步骤 11；CONTRACTS §3（选项来源）、§5.1（`options_from` 与「选项来源」）、§16（`OPTIONS_PAGE`）；WIRE-FORMAT §3.5 的追加字段；DECISIONS D-129；TEST-MATRIX V-116 的 schema、SDK 与 Core 分项。

范围（只新增或修改这些）：
- contracts/configschema/：options_from 的解析与检查，以及按字段路径读取 options_from 的方法；内置来源的名称与 filter 键登记在本包，供 schema 检查与 Core 共用
- contracts/proto/alonabot/v1/route.proto：ServiceSpec 追加选项来源标记；随后运行 go generate ./...；contracts/wirespec_test.go 追加编号检查。本切片获授权只追加这一个字段
- contracts/validate/register.go：带标记的服务的注册校验；固定 Schema 的标识常量放在 contracts/validate
- internal/tune：登记 OPTIONS_PAGE
- sdk/：新增 options.go（ServeOptions 与三个类型，Schema 由 SDK 固定提供），复用 services.go 的内部注册辅助；sdk/sdk.go 的 Collect 只加检查调用点
- 新增 internal/core/options/：查询入口与内置来源 core.adapter_instances
- 根包 assembly.go：构造查询入口并保存在装配中，供 T-09.14 的管理 API 使用
- internal/core/adapter/remote/remote_test.go：新增一个主路径集成测试
- 上述包的 *_local_test.go
允许的新跨模块边：无（options 与 router、broker 同属 internal/core 模块）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. contracts/configschema
   var BuiltinOptionSources = map[string][]string{"core.adapter_instances": {"adapter_id"}, "core.chats": {"chat_type", "adapter_id"}} // 来源名 → 允许的 filter 键；core.chats 由 goal-131 实现查询，本切片只登记名称
   type OptionsFrom struct { Source string; Filter map[string]string }
   func (s *Schema) OptionsFrom(fieldPath string) (OptionsFrom, bool) // fieldPath 规则同 CONTRACTS §5.1：以 . 连接，list 子字段不含下标，模板子字段为 <字段>.<模板键>.<子字段>
2. proto：ServiceSpec 追加 bool options_source = 6;
3. contracts/validate
   const OptionsQuerySchemaID, OptionsPageSchemaID = ..., ... // 与 SDK 由 Go 类型推导出的 schema_id 相同，版本均为 1
4. internal/tune：OPTIONS_PAGE，缺省 50，范围 1～200
5. sdk
   type OptionsQuery struct { Search string; Filter map[string]string; Locale string; Limit int; Values []string }
   type OptionsPage struct { Options []Option; More bool }
   type Option struct { Value, Label, Description string } // Description 可为空，JSON 中省略
   func ServeOptions(r Registrar, serviceName string, fn func(ctx context.Context, q OptionsQuery) (OptionsPage, error))
6. internal/core/options
   type Query struct { Search string; Locale string; Limit int; Values []string } // filter 由入口取自 schema，调用方不能指定
   type Page struct { Options []Option; More bool }
   type Option struct { Value, Label, Description string }
   type SchemaKind int // ConfigSchema、AdapterInstanceSchema
   type Options struct {
       Schema    func(componentID string, kind SchemaKind) (*configschema.Schema, bool) // 该组件 Active 代次的 schema
       Router    *router.Router
       Broker    *broker.Broker
       Instances func() []AdapterInstance // 已配置的 Adapter 实例
       Page      int                      // OPTIONS_PAGE
   }
   type AdapterInstance struct { InstanceID, AdapterID string }
   func New(opt Options) *Service
   func (s *Service) Query(ctx context.Context, componentID string, kind SchemaKind, fieldPath string, q Query) (Page, error)

行为：
- schema 检查：
  - `options_from` 只用于 string 与没有 items 的 list；与 options、labels、secret 并存时报错。
  - source 必须是 `core.<name>`、`<组件 ID>/<service_name>` 或 `@agent/<service_name>`；内置来源的名称与 filter 键必须在 BuiltinOptionSources 中；filter 的值必须是字符串。错误带 JSON 路径。
  - `adapter_config_schema.json` 走同一检查，`plugin check` 与生成器共用。
- 配置的行为不变：补缺失值、校验、`Config()` 的读取、Merge Patch 与 secret 都不因 options_from 改变，不按来源校验取值。
- 注册：
  - ServeOptions 注册一个 JSON Service，输入为 OptionsQuery、输出为 OptionsPage，访问规则固定为 PRIVATE，不可调用 Access 或 Tool，描述带 options_source=true。
  - 同一服务名注册两次或 fn 为 nil 时，Collect 报错。
  - Core 在 REGISTER 时核对带标记的服务：编码为 JSON，输入输出 schema_id 与版本等于 validate 的常量，访问为 PRIVATE 且没有 allowlist，没有 tool；不符时拒绝注册。
- 查询入口（Query）：
  - 从 Options.Schema 取该组件的 schema，按字段路径找到 options_from；字段不存在或没有 options_from 时返回 INVALID_ARGUMENT。
  - Limit 为 0 时取 Page，超过 Page 时按 Page 处理。来源返回的选项多于 Limit 时只取前面的，并把 More 置为 true。Values 非空时按存储值查询，不存在的值不返回。
  - 内置来源在进程内查询。`core.adapter_instances`：存储值为 adapter_instance_id，显示名为实例 ID，说明为 adapter_id，filter 的 adapter_id 按相等筛选。`core.chats` 尚未实现时返回 UNAVAILABLE。
  - 服务来源：组件部分 `@agent` 时取 `role: agent` 的组件。经 Router 找到目标的 Active 路由，只接受 options_source 的服务，再经 Broker 以 Core 的身份（不带用户 Actor）调用，filter 取自 schema 原样传入。
  - 目标未安装、未激活、没有标记、调用失败或响应不合法时，返回 UNAVAILABLE，不影响配置的读取与保存。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 internal/core/adapter/remote/remote_test.go 的 TestOptionSources，复用测试二进制中的插件进程、真实 Session、Router 与 Broker。覆盖：
  - 一个插件为自己的配置字段提供来源，Core 按字段路径查到选项；
  - 按存储值取回显示名；
  - 目标未激活时返回 UNAVAILABLE。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：
  - schema 检查的各种拒绝（类型、互斥、来源名格式、内置名称与 filter 键、filter 值类型、JSON 路径）；
  - Collect 对重复与 nil 的拒绝，REGISTER 对标记服务的各项拒绝；
  - 超出上限的截断与 More；未带标记的服务被拒；`@agent` 目标；Adapter 实例 schema；core.adapter_instances 的 filter。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race . ./contracts/... ./internal/core/... ./sdk/...；go build ./...。

不做：管理 API（goal-93）；页面（goal-44）；`core.chats` 的查询（goal-131）；`llm.providers`（goal-83）；选项之间的级联；按来源校验已保存的值；用户与插件列表的内置来源。
```

### goal-157 · T-08.17 插件的 KV 与数据目录

状态：已完成（PR #212），合并提交 `4d69978`。以下保留派发时语句。派发时：前置 T-08.5 与 T-09.4 已合并。2026-11-17 按 D-165 新增。

```text
/goal 使用 $alonabot-execute 推进 goal-157（切片 T-08.17：插件的 KV 与数据目录）。goal 目录 goal-157，分支 goal-157，worktree .worktrees/goal-157，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：新增 `sdk/kv`，在受管 SQLite 的保留表 `_alonabot_kv` 中提供键值存储；Core 在启动插件前建立 `<data_dir>/plugin_data/<id>/`，经 Bootstrap 的 `plugin_data_dir` 下发，SDK 以 `sdk.PluginDataDir()` 返回；卸载的 `--purge-data` 与 `--purge` 删除该目录。

依据：PLAN T-08 步骤 14；CONTRACTS §1.1（Bootstrap）、§6.1、§9.3（uninstall）；WIRE-FORMAT §3.2；DECISIONS D-165；TEST-MATRIX V-137 的 SDK 与 Core 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：Bootstrap 按 §5A 只追加 `string plugin_data_dir`；go generate 后的生成产物；contracts/wirespec_test.go 固定新编号
- sdk/kv/（新包）、sdk/sqlitedb/（初始化时建保留表）、sdk/ 中读取 Bootstrap 与导出 PluginDataDir 的文件
- internal/core/host 或 internal/core/process 中准备插件运行环境的位置（建立目录与权限）、internal/core/manage 中卸载清理的位置
- internal/pluginkit（链接 sdk/kv 的插件按链接受管 SQLite 推出 prepare_effects）
- examples/plugins/configdb：用 KV 与数据目录各写一处，作为示例
- docs/plugin-guide.md：KV 与数据目录的用法
- 上述的测试文件

接口：
- package kv
  func Get(ctx context.Context, key string) (value []byte, ok bool, err error)
  func Set(ctx context.Context, key string, value []byte) error
  func Delete(ctx context.Context, key string) error
  func List(ctx context.Context, prefix, cursor string, limit int) (Page, error)
  type Page struct { Keys []string; Next string } // Next 为空表示没有更多
  // 键为 1～512 字节的 UTF-8；值最大 1 MiB（命名常量）；ctx 取自 Handler 或生命周期回调，使用本 Session 的受管 SQLite
- package sdk：func PluginDataDir() string // Session 就绪后可用，之前调用 panic，消息为英文

行为：
- `_alonabot_kv(key TEXT PRIMARY KEY, value BLOB NOT NULL, updated_at INTEGER NOT NULL)` 随受管 SQLite 初始化建立，不占用插件的迁移 ID，候选副本与回退照常包含它；没有受管 SQLite 的插件调用 kv 返回 FAILED_PRECONDITION。
- 目录权限只有运行账号（Windows 用与状态库相同的 ACL）；Core 不读写目录内容。只卸载与 disable 保留目录；`--purge-data` 与 `--purge` 在卸载确认的对象列表中列出并删除它。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/plugin_cli_e2e_test.go 新增一个子测试（unix，真实 Core 与 configdb 示例）：示例写入 KV 与数据目录中的文件，restart 后仍能读到；`uninstall --purge-data` 之后目录不存在，重装后 KV 为空。
- 本地自测（*_local_test.go，结果写入 PR）：键与值的上限、List 的前缀与分页游标、没有受管 SQLite 时的错误、PluginDataDir 就绪前 panic、只卸载保留目录、目录权限；Windows 交叉编译。

完成条件：gofmt -l 无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...。

不做：管理 API 的 KV 与文件浏览（T-09.21）；WebUI 的存储视图（goal-156）；KV 的过期与事务。
```

### goal-84 · T-09.12 WebUI 组件与组件图标

状态：已完成（[PR #167](https://github.com/AlonaBot/AlonaBot/pull/167)，合并提交 `3a9b4af`），证据见 TEST-MATRIX §3.2。合入：`manage.AdminAPIMajor = 1`；`(*manage.Service).WebUI()`；根包选项 `alonabot.WebUI(fsys fs.FS)`（开发 WebUI，D-140 不做 Admin API 版本检查）；WebUI 静态托管（保留 /api/ 与内部路径，无扩展名的不存在路径回退 index.html，未安装时 503）与认证的组件图标端点；`spec/ADMIN-API.md` 登记现有管理端点。

按 main 核对（2026-11-12，已用 rg 核对签名、字段和路由；manage.Service.Handler 另以离线 go doc 核对）：

- 根包 `alonabot.go` 已有 `type Option func(*options)`、`Run(opts ...Option) error`、`Start(ctx context.Context, opts ...Option) (*Core, error)`、`(*Core).Wait() error`；选项含 `Plugin(root *sdk.Plugin)`、`Adapter(a adapter.Adapter)`、`DataDir(dir string)`、`Listen(addr string)`、`LogLevel(level slog.Level)`、`LogFormat(format string)`、`Args(values ...string)`，均返回 Option。当前无 WebUI 选项。Start 只读取一次 settings.File，然后装配认证、管理服务、HTTP Listener 与已启用组件。
- `alonabot.go` 当前依次 `authentication.Routes(adminMux)`、`adminMux.Handle("/", management.Handler(authentication))`，再把 adminMux 传给 webhost；静态入口须从这个根路由拆开，否则会落入管理认证/404。启动进程的安装遍历已跳过 `state.RoleWebUI`。
- `internal/core/webhost/host.go`：`New(opt Options) *Host`；`Options{BaseURL *url.URL, Admin http.Handler, Media http.Handler, Proxy *httpproxy.Server, Now func() time.Time}`。`(*Host).ServeHTTP` 验证 Host，先分流 `/plugins/`、`/media/`，其余交 Admin。保留这条真实分流及一个公开端口。
- `internal/core/adminauth/auth.go`：`New(st *state.Store, now func() time.Time, adminBaseURL *url.URL, dataDir string) *Auth`、`(*Auth).Routes(mux *http.ServeMux)`、`Require(next http.Handler) http.Handler`、`Authorize(r *http.Request, pluginID string) bool`；现有路由仅 POST `/api/auth/login`、POST `/api/auth/logout`。登录成功返回 `{"csrf_token":...}`；Cookie 变更请求使用 `X-Alonabot-CSRF` 并验证 Origin，CLI 使用 Bearer。
- `internal/core/manage/service.go`：`New(opt Options) *Service`，Options 已有 `Store *state.Store`、`Settings *settings.File`、`Packages string`、`Development map[string]*manifest.Manifest` 与生命周期函数；`(*Service).List(ctx context.Context) ([]Package, error)`。私有 `directory(in state.Install) string` 用安装版本与 BuildDigest 前缀定位包目录，`manifest(ctx context.Context, id string) (*manifest.Manifest, error)` 读包内或开发组件 manifest，静态入口应复用它们。
- `internal/core/manage/http.go`：`(*Service).Handler(auth *adminauth.Auth) http.Handler` 统一认证；已注册 `/api/v1/packages`、packages 的 install/enable/disable/restart/uninstall/pin/reload/update/rollback、`/api/v1/admins` 及 add/remove、`/api/v1/adapter/instances/` 的 create/start/stop/delete。成功结果直接编码结构体；409 确认响应是 `{"affected":[...]}`，其他管理错误是 `{"code":...,"safe_message":...}`；认证错误目前为英文纯文本。准确大小写及字段以源码 JSON 标签为准。
- `internal/core/manage/lifecycle.go` 的 `Enable(ctx context.Context, id string) error` 对 WebUI 只设 Enabled；尚无 Admin API 版本检查。`install.go` 的 `Install(ctx context.Context, req InstallRequest) (Package, error)` 先 prepare，再同类替换；兼容性检查须在替换旧包前完成。`candidates.go` 的 Update 也先 prepare，候选执行当前仍为 UNSUPPORTED，属后继切片。
- `contracts/manifest/manifest.go`：`Manifest{Role, Icon, Assets string; AdminAPIMinMajor, AdminAPIMaxMajor uint32; ...}`、`Parse(data []byte) (*Result, error)`、`ParseProject(data []byte) (*Result, error)`（project.go）、`ValidateRelPath(p string) error`。WebUI 已要求版本范围与 assets 路径；icon 已限制为 `.svg/.png/.webp`。`internal/core/packages/preflight.go`：`Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error)`、`(*Candidate).MoveTo(packagesRoot string) (string, error)`；解压已拒绝链接与越界路径，当前只核对进程 entrypoint，尚未核对 WebUI index.html 与图标实际文件。
- 在途冲突：goal-69、goal-81 同改根包 `alonabot.go`/`assembly.go`，goal-22、goal-55 也改装配；本切片集中在 alonabot.go 的 Option 和 HTTP 挂载，不改 cmd。goal-131 改 state/manage 的清理，本切片只改 manage 的静态入口、安装/启用门禁与 HTTP 注册；goal-59 在 agent/ 无直接文件冲突。本切片与 goal-127 可分别追加 manage/http.go 之外的新文件；state/schema.sql 由 goal-55/goal-131/goal-127 负责，本切片无需修改。后合并者同步保留其他 HTTP 路由与装配项。

- 审查者决定（2026-11-12）：开发 WebUI 不做 Admin API 版本检查并记 WARN（D-140，待维护者确认）；WebUI 静态响应只加契约写明的头；采纳本语句的桥消息格式；图标端点也接受 HEAD（与 GET 相同的状态与头、无正文）；本切片的图标只覆盖已安装组件，开发组件的图标另行安排；spec/ADMIN-API.md 如实记录当前端点（含字段大小写与纯文本认证错误），不借本切片统一格式。
```text
/goal 使用 $alonabot-execute 推进 goal-84（切片 T-09.12：WebUI 组件与组件图标）。goal 目录 goal-84，分支 goal-84，worktree .worktrees/goal-84，从最新 origin/main 创建（包含 PR #116、PR #72、PR #149、PR #163）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 提供已启用 WebUI 组件的静态文件、单页路由回退与缺失/版本不兼容的英文提示；定义 Admin API 主版本并拒绝不兼容的安装与启用；通过管理员认证的入口只提供 manifest 声明的图标并施加沙箱响应头。新增库模式 alonabot.WebUI(fsys fs.FS) 选项，以及包含当前全部管理 HTTP 接口、认证/错误格式和插件页面桥消息格式的 spec/ADMIN-API.md。

依据：spec/PLAN.md §3.1；spec/plan/core.md T-09.12、T-09 步骤 9、相关的步骤 3/7 与「对应 V 验收与停止条件」；spec/CONTRACTS-v0.1.md §7.2～§7.5、§9.1、§9.3、§15.5；spec/MESSAGE-MODEL.md §1 的 Target、Actor 与事件来源（代理既有消息/事件接口时沿用）；spec/WIRE-FORMAT.md §3.3 的现有 HTTP 路由描述与 Listener 凭证、§3.4 中 Adapter 平台 token 路由与内部代理凭证的区分；spec/adr/DECISIONS.md D-62、D-63、D-74、D-77、D-78、D-84、D-86、D-139；spec/TEST-MATRIX.md V-83，并回归 V-47 的管理认证、CSRF 与插件沙箱。对应工作包在 spec/plan/core.md。

范围（只新增或修改这些）：
- internal/core/manage/service.go：Options 增加开发 WebUI 文件系统；新增 static.go：WebUI 选择、版本检查、静态与图标提供。复用现有安装记录、manifest 读取和包目录定位。
- internal/core/manage/install.go、lifecycle.go：安装/替换预检与 Enable 的版本门禁；candidates.go 只在已有 Update 预检链补同一兼容性检查，候选执行保持后继所有者。
- internal/core/manage/http.go：追加图标端点，复用现有认证；新增静态入口方法与原管理 Handler 分开。
- internal/core/packages/preflight.go：验证 assets/index.html 与 icon 是包内普通文件，沿用已批准的解压与路径规则。
- 根包 alonabot.go：WebUI Option、options 字段、给同一个 manage.Service 传 fsys，以及 Admin API/静态入口的 HTTP 挂载；其余装配保持现有所有者。
- 新增根包 webui_test.go（unix）：一项经真实 Start/管理服务/公开 Listener/临时组件包的集成测试；上述包的 *_local_test.go。
- 新增 spec/ADMIN-API.md：当前 API 主版本、全部已实现端点和插件页面桥的格式。后继新增接口由 goal-66/goal-92/goal-93 追加。

允许的新跨模块边：无。静态托管与包身份解析都在 internal/core 内；根包经已允许的 internal/core 装配。fs.FS 使用标准库；Core 保持通过协议接入组件。

接口（照此实现；内部辅助函数自定）：
1. internal/core/manage：
   const AdminAPIMajor uint32 = 1
   Options 追加 DevelopmentWebUI fs.FS // 本次运行的静态根目录；未传时选择已启用的安装包
   func (s *Service) WebUI() http.Handler
   保持 New、Install、Enable、Handler 等现有公开签名。WebUI Handler 处理静态 GET；管理 API 仍由 Handler(auth) 提供。
2. 根包：
   func WebUI(fsys fs.FS) Option
   fsys 根目录即静态根目录；本次运行取代已安装的 WebUI，不将 dev WebUI 写成 Native 安装记录、不启动进程。开发 WebUI 没有 manifest，不做 Admin API 版本检查，启动时记一条 WARN（D-140，待维护者确认）；同一次 Start 登记两次 WebUI 选项返回错误。
3. HTTP：
   保持 POST /api/auth/login、POST /api/auth/logout 与全部 /api/v1/ 路由。
   GET /api/v1/packages/{id}/icon：只取该组件 manifest.Icon 的文件；无图标/未安装/已卸载/非文件返回 404，未认证返回 401。没有任意包路径参数；额外路径和查询参数均不得改变被选择的文件。
   Admin 挂载：authentication.Routes(adminMux)；/api/v1/ 交 management.Handler(authentication)；其余 /api/ 交同一管理错误入口；/ 根路由交 management.WebUI()。已登记的 auth 路由优先。/plugins/、/media/ 仍在 webhost 分流，所有其他 Core 保留路径在静态查找前拒绝。
4. spec/ADMIN-API.md 的桥消息格式（供后继 WebUI 页面实现；本切片只定义格式）：
   request: {channel:"alonabot.plugin", id:string, op:"request", method:string, path:string, headers?:Record<string,string>, body_base64?:string}
   locale/theme 请求: {channel:"alonabot.plugin", id:string, op:"locale"|"theme"}
   文案请求: {channel:"alonabot.plugin", id:string, op:"translate", key:string, args?:Record<string,string>}
   response: {channel:"alonabot.plugin", id:string, ok:true, result:{status:number,headers:Record<string,string>,body_base64:string}} // request
   response: {channel:"alonabot.plugin", id:string, ok:true, result:{locale:string}} // locale
   response: {channel:"alonabot.plugin", id:string, ok:true, result:{theme:string}} // theme
   response: {channel:"alonabot.plugin", id:string, ok:true, result:{text:string}} // translate：只查该组件的 messages
   error: {channel:"alonabot.plugin", id:string, ok:false, error:{code:string,safe_message:string}}
   body_base64 使用标准 base64，缺省请求正文为空；响应保留完整字节。translate 按当前界面语言及该组件已有 i18n 回落规则格式化 messages.key。以上格式已由审查者采纳，写入 spec/ADMIN-API.md，后继 WebUI 实现遵循它。

行为：
- 安装 WebUI 时，在 prepare 已完成且任何旧组件停止/卸载前，检查 AdminAPIMinMajor <= AdminAPIMajor <= AdminAPIMaxMajor；不兼容返回 FAILED_PRECONDITION，英文提示含 Core 主版本和组件声明范围。Enable 同样检查，失败保持原 Enabled。Update 复用同一预检，避免未来候选流程跳过版本门禁。保持同类唯一、--replace、来源绑定和确认语义。
- 每次提供已安装 WebUI 时，依据 Store.Installs/Enabled 选择 role=webui、未卸载且已启用的唯一记录，复用 directory 和 manifest 的 assets。Core 升级后范围失配时返回 503 与英文纯文本版本提示，管理 API 仍可用；缺少已启用 WebUI 时返回 503 与安装提示，明确使用 alonabot install 安装 WebUI 组件。安装损坏、读取/状态库错误保留并记录安全错误，响应 500，拒绝伪装成未安装。
- assets/index.html 与 icon 的存在性/普通文件检查在预检边界完成；已安装提供时同样保持文件边界。路径使用 fs.ValidPath/manifest.ValidateRelPath 的相应形式，拒绝 ..、反斜杠、绝对路径和链接；禁止先 Clean 后接受越界输入。包目录只读，不遍历暴露整个包。根静态路径映射为 index.html。
- GET 对应文件存在时提供该文件。不存在且 URL 最后一段无扩展名时提供 index.html；带扩展名的不存在路径返回 404；目录按静态根语义处理且不列目录。实际 I/O 错误返回错误，只有文件不存在才适用回退。非 GET 的静态请求返回 405。HEAD 的支持按备注定稿，保持静态与 API 的方法规则清晰。
- /api/、/plugins/、/media/ 及已有内部路径在回退前分流，即使 WebUI 包内有同名文件也不覆盖；未知 API 返回既有 API 错误，保留状态码和 Content-Type。所有静态请求继续由 webhost 验证受控 Host，错误 Host 为 421。无 WebUI 时登录、管理与插件/媒体入口照常可达。
- WebUI 是登录页面本身的来源，静态资源可在登录前获取；管理 API 和图标始终通过现有认证。图标只打开该组件自身的 manifest.Icon；MIME 只按扩展名取 image/svg+xml、image/png、image/webp。图标响应强制 Content-Security-Policy: sandbox; default-src 'none'; style-src 'unsafe-inline' 与 X-Content-Type-Options: nosniff，包括直接打开 SVG 的场景。未声明文件与其他扩展名一律拒绝提供。开发插件/Adapter 的图标若缺少可访问包目录，按备注定稿，禁止推测其工程路径。
- dev WebUI 选项只在本次运行覆盖安装包静态来源；不修改已安装组件的启用状态、来源或包文件。其资源使用同一回退与保留路径规则，不做 Admin API 版本检查（见接口 2）。
- 静态 WebUI 响应不额外加 CSP 或 X-Frame-Options（契约未规定，审查者决定只落实已写明的头）；图标与插件代理继续使用各自契约规定的响应头，避免将插件 CSP 套在持有管理会话的 WebUI 页面上。
- API 参考逐个登记当前源码的 method/path、认证方式、请求/响应 JSON 的准确字段大小写、状态码与确认流程；包括 auth、packages 全部管理动作、admins、Adapter 实例和新增图标。管理成功为直接结果、409 affected、其他错误 code/safe_message，认证与路由的纯文本错误单列。reload/update/rollback 写明当前候选执行返回 UNSUPPORTED（HTTP 501），后继实现再修订能力说明。上传与会话取回等后继端点由所属切片实现时追加。
- API 参考写明 Cookie/Origin/X-Alonabot-CSRF、CLI Bearer、8h 会话、缺认证时的行为、WebUI 版本范围与同主版本的兼容增加。插件路由的内部凭证由 SDK/Core 管理，浏览器桥无需取得它。
- 桥格式（已采纳）：请求来源按已嵌入的 iframe.contentWindow 绑定组件 ID；不透明源的 origin 为 null，父页面同时核对窗口身份。父页面向正确窗口回应，request.id 对应原请求。只接受插件相对路径，拒绝绝对 URL、越界路径和重定向；父页面将路径固定到该 iframe 所属组件的 /plugins/{id}/ 前缀，Core 现有代理负责已注册 method/path 的最终匹配，禁止访问其他组件和 Admin API。父页面添加管理员会话/CSRF，页面自带头中拒绝 Cookie、Authorization、CSRF 与 Core 内部凭证，响应也滤去 Set-Cookie 等凭证。语言/文案/主题只返回相应值。后继前端实现遵循本格式。

测试（按编码规范「测试提交规则」）：
- 提交：只新增根包 webui_test.go 的 TestWebUIComponents（//go:build unix）。用临时 data_dir、真实 alonabot.Start 的单一 HTTP Listener、真实认证与管理服务，在临时目录构造并安装 WebUI 组件包。一个测试内分子场景覆盖 V-83：启用后取 index/资源；无扩展名前端路由回退；带扩展名缺失 404；/api/ 和内部路径保持分流；未安装/停用 503 及英文提示；安装与启用时版本不兼容被拒且旧 WebUI 可继续提供；Core 主版本升级式模拟下，已安装不兼容 WebUI 返回版本提示；WebUI(fsys) 在契约定稿后通过同一 Start 路径覆盖已安装静态来源。
- 同一测试安装声明 SVG/PNG/WebP 图标的临时组件包，经图标公开端点验证认证、MIME、完整 CSP/nosniff、只读取声明文件；未声明、额外包路径、其他扩展名与越界路径拒绝。图标错误扩展名的包在安装预检被拒。测试包在 t.TempDir 动态生成，无新增 testdata 插件；包校验所必需的摘要沿用安装夹具，不作额外完整性计算。
- 本地自测（*_local_test.go，PR 写明结果）：所有 reserved 前缀、未知 API 的认证/404 与 MIME、非 GET 方法、编码后的路径与 symlink、目录/index 缺失和普通 I/O 故障；Host 校验；禁用/卸载/替换后的即时选择；--replace 拒绝不兼容候选时旧包与 Enabled 保持；Update 预检；图标 Cookie 与 Bearer、登录 CSRF 回归；重复 WebUI 选项按审查者定稿的规则处理。既有测试只做必需接口迁移，原断言保持。
- 浏览器本地验证：真实 Core 上登录，直接打开包含 script 的 SVG 图标，确认脚本无法访问 Admin 源或产生副作用；记录浏览器版本与实际结果。环境缺少浏览器时写“未运行”，V-83 此分项保留待联验；页面桥的执行验证由后继前端切片负责，本切片检查文档结构与安全规则完整。
- 不向 sdk/、internal/core/supervisor/ 增加提交测试；多进程本地自测如需插件页面，复用根包 pipeline_test.go 的真实 Session 装配。

完成条件：WebUI 提供/回退、保留路径、版本拒绝、图标认证与头部、库模式覆盖均有集成证据；spec/ADMIN-API.md 与当前 main 的所有端点逐项对应，桥格式已写入。gofmt -l 本次修改 Go 文件无输出；go vet ./...；go test -race . ./internal/core/manage/... ./internal/core/packages/... ./internal/core/webhost/... ./internal/core/adminauth/...；go build ./...。PR 给出命令、OS/架构、Go/浏览器版本、提交测试新增行数及本地验证结果，未运行项按原样交审查者回写 TEST-MATRIX §3.2。遇到 T-09 的停止条件、额外依赖、范围外文件或未决契约时标 [!]。

不做：WebUI 前端页面与桥运行代码（goal-43/goal-44）；Core 内嵌前端；插件页面票据/第二 host；会话取回与 multipart 上传（goal-92）；日志 API（goal-66）、配置/权限/工具/会话管理 API（goal-93）；候选 Journal、正式更新与 rollback 编排（goal-40）；任意包文件浏览、表单选项图标、组件进程静态代理；新依赖和 proto 变更。
```

### goal-69 · T-13.3 开发运行与热部署

状态：已完成（[PR #171](https://github.com/AlonaBot/AlonaBot/pull/171)，合并提交 `3043c4e`），证据见 TEST-MATRIX §3.2。合入：`internal/pluginkit.BuildDev(ctx, dir, outDir) (Build, error)` 与 `Build{Executable, Manifest, PackageManifest, Digest, Role, ID, AdapterID}`；根包 `Executable(path string, packageManifest []byte) Option`（第二个参数为构建生成的包内 manifest）、`Stdio() Option`、`(*Core).Restart(ctx, id)`、`(*Core).RestartInstance(ctx, adapterID, instanceID)` 与执行中审查者补齐的 `(*Core).ReloadSettings(ctx) error`（同步 Core Admin、缺省实例与全局日志级别，不重启组件，DECISIONS 2026-11-13，待维护者确认）；`alonabot run [目录...] [--reload] [--stdio]` 与目录 WebUI。以下为派发时的语句，接口段未含执行中的补齐。
- 根包：`alonabot.Run`/`Start` 与选项 `Plugin`（接受 role plugin 或 agent）、`Adapter`、`DataDir`、`Args`、`Listen`、`LogLevel`、`LogFormat`；Start 只加载一次 `settings.File` 并交给装配；开发组件由 `startDevelopment` 以同进程方式运行，配置取 `plugins/<id>/config.json`、数据库为 `dev.sqlite`（T-13.2），开发 Adapter 按 `core_config.json` 中属于它的实例启动（T-09.4）；已安装组件由 `startInstalled` 经 `supervisor.Launch` 以独立进程运行。根包目前没有「开发组件以独立进程运行」与「运行中只重启一个开发组件」的入口，本切片加入。
- `cmd/alonabot/commands.go` 的 run：`alonabot run [--data-dir] [--listen] [--log-level] [--log-format] [--args]` 调用 `alonabot.Run`，不接受位置参数；不带子命令时等同 run。
- 生成器 `sdk/generate` 与 `alonabot plugin generate` 已有；`internal/pluginkit` 尚不存在（ARCHITECTURE 已登记 `cmd/alonabot -> internal/pluginkit` 与 `internal/pluginkit -> sdk/generate` 的边，供 goal-38 的 `alonabot build` 复用）。
- tune 已登记 `DEV_RELOAD_POLL`（缺省 500ms，CONTRACTS §16）。
- `alonabot init` 由 goal-38 提供，goal-38 依赖本切片，所以测试不能用 init，工程文件由测试直接写出。

```text
/goal 使用 $alonabot-execute 推进 goal-69（切片 T-13.3：开发运行与热部署）。goal 目录 goal-69，分支 goal-69，worktree .worktrees/goal-69，从最新 origin/main 创建（须已包含 PR #163）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：`alonabot run [目录...] [--reload] [--stdio]` 把给出的目录（没有给出且当前目录有 plugin.yaml 时为当前目录）作为开发组件：按 `alonabot build` 将用的方式构建到 `<data_dir>/dev/bin/`，插件、Adapter 与 Agent 以独立进程运行，沿用固定代次 `dev:<组件 ID>`（D-134）与 T-13.2 的开发配置和数据库。`--reload` 时每 DEV_RELOAD_POLL 轮询源码、资源与配置：源码或资源变化时重新构建，成功后只重启变化的组件，编译失败时输出编译器信息、旧组件继续运行；只有配置变化时不重新构建，只重启该组件或变化的 Adapter 实例。

依据：PLAN T-13 步骤 4；CONTRACTS §15.4、§15.1、§16（DEV_RELOAD_POLL）；DESIGN §S；DECISIONS D-77、D-134；TEST-MATRIX V-64 的独立进程分项、V-65。

范围（只新增或修改这些）：
- 新增 internal/pluginkit/：开发构建（读取工程 plugin.yaml、运行生成器、go build 到指定目录；Adapter 工程包生成 cmd 薄入口，不在作者目录写文件），供 goal-38 复用
- 根包 alonabot.go、assembly.go：独立进程的开发组件与运行中重启单个开发组件（见接口）
- cmd/alonabot/commands.go 与新增 cmd/alonabot/dev.go：run 的位置参数、--reload、--stdio 与轮询
- cmd/alonabot/e2e_test.go：恢复 stdio 子项（改为以目录运行 examples/plugins/lifecycle）；新增 cmd/alonabot/dev_e2e_test.go
- docs/development.md 的开发运行一节
允许的新跨模块边：cmd/alonabot -> internal/pluginkit、internal/pluginkit -> sdk/generate 与 contracts/manifest（ARCHITECTURE 已登记）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. 根包
   func Executable(path string, projectManifest []byte) Option // 登记以独立进程运行的开发组件：projectManifest 为工程 plugin.yaml 原文（按 manifest.ParseProject 解析 id、role、adapter_id，规则同 Plugin/Adapter 的登记与冲突检查），path 为构建出的可执行文件
   func Stdio() Option                                         // 本次运行中 Executable 登记的组件使用 stdio 传输（缺省 TCP loopback）
   func (c *Core) Restart(ctx context.Context, id string) error // 只重启一个开发组件：重新读取可执行文件摘要与 config.json，按 §4.2 停止旧 Session 后以同一代次 dev:<id> 启动；Adapter 重启后按属于它的实例重新启动；id 不是本次运行的开发组件时返回错误
   func (c *Core) RestartInstance(ctx context.Context, adapterID, instanceID string) error // 只重启开发 Adapter 的一个实例（配置变化时）
2. internal/pluginkit
   func BuildDev(ctx context.Context, dir, outDir string) (Build, error) // 返回可执行文件路径、工程 manifest 原文与构建输入的摘要；编译失败返回带编译器原文的错误
   type Build struct { Executable string; Manifest []byte; Role, ID, AdapterID string }

行为：
- 组件身份取自目录中的工程 plugin.yaml；构建失败、ID 或 adapter_id 与已安装的包冲突、两个目录的 ID 相同时，在启动任何组件前报错退出。
- 构建复用 sdk/generate 与 go build；插件与 Agent 工程直接构建入口包；Adapter 目录为可导入的工程包，构建时为它生成只调用 sdk.RunAdapter(New()) 的薄入口（放在 <data_dir>/dev/ 下或经 go build -overlay，不写入作者目录）。
- 开发组件进程沿用 supervisor 的启动、停止与有限重启，与已安装组件相同；配置、数据库、日志级别与 T-13.2 相同（独立进程的开发组件由构建写入明确的 migration_mode）。
- 轮询（--reload）：每 DEV_RELOAD_POLL 用标准库检查每个组件目录中的 .go 文件、//go:embed 引用的文件、go.mod、go.sum、plugin.yaml、config_schema.json、adapter_config_schema.json 与 i18n/（排除数据目录、vendor、以 . 或 _ 开头的目录与 testdata），以及组件的 config.json 与 core_config.json 中属于该 Adapter 的实例。源码或资源变化：重新构建，成功后依次调用 Core.Restart；失败时输出编译器原文，旧组件继续运行。只有 config.json 变化：不构建，Core.Restart 该组件。只有某个 Adapter 实例的配置变化：Core.RestartInstance 该实例。core_config.json 的其他变化交给 settings 的 Reload（admins 变化时 SetAdmins，命令前缀经 File.CommandPrefixes 读取当前设置；default_locale 在之后建立的 Session 生效，不为此重启组件）。
- 不带 --reload 时只构建并运行一次。WebUI 目录（role: webui）以目录中的 assets 代替已安装的 WebUI（与 alonabot.WebUI 选项同义；该选项尚未实现时本项登记为阻塞，不在本切片实现 WebUI 托管）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - cmd/alonabot/dev_e2e_test.go（//go:build unix；多列理由：本切片首次接通目录运行与热部署，没有其他端到端入口）。测试在临时目录写出两个工程（各自的 go.mod 以 replace 指向本仓库）：一个插件（/hello 回复配置中的文本），一个回环 Adapter 工程包（声明私有 HTTP 路由，测试以管理 token 调用它注入 message.created，Adapter 把收到的发送写入文件），以 `alonabot run <插件目录> <Adapter 目录> --reload` 运行：
    - 改动插件的回复文本后，插件重启并回复新文本；Core 与 Adapter 进程不重启（PID 不变）。
    - 写入语法错误后，输出含编译器信息，旧插件继续回复。
    - 只改 config.json 时不重新构建（dev/bin 中的文件不变），新配置生效。
    - 改动 config_schema.json 后重新构建，新的缺省值生效。
  - cmd/alonabot/e2e_test.go：TestLifecycleOverTCPAndStdio 的 stdio 子项以 `alonabot run <examples/plugins/lifecycle> --stdio` 恢复，断言不变。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：ID 冲突与构建失败时在启动前退出；轮询的排除规则；只改 Adapter 一个实例时只重启该实例；同一目录两次给出时报错；Core.Restart 对非开发组件报错；不带 --reload 时不轮询。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：alonabot init、build、check、test、pack 的 CLI（T-09.3）；开发示例与文档汇总（T-13.4）；WebUI 的静态托管（T-09.12）。
```

### goal-81 · T-16.1 插件测试套件基础

状态：已完成（[PR #168](https://github.com/AlonaBot/AlonaBot/pull/168)，合并提交 `37b0ada`），证据见 TEST-MATRIX §3.2。合入：公开包 `plugintest`（`Start`、`Plugin`、`Config`、`Admin`、`AllowedInstances`、`Capabilities`、`Adapter`、`Private`、`Group`、`(*Harness).Message`/`Text`/`Event`、`Result`、`Sent`、`Outcome`）；`internal/inproc` 的 `Harness`、`RegisterHarness`、`HarnessFor`；`router.Options.Dispatched`（一个事件的分发链结束或进入等待时回调）。
- 插件对象：`*sdk.Plugin`（`sdk.NewPlugin`，`p.Generated().Manifest` 为工程 plugin.yaml 原文，D-137）；SDK 公开类型用 `sdk/message`（`message.Session`、`message.Chain`、`message.Content`、`message.SendResult`、`message.MessageRef`）与 `sdk/errcode`，pb 只在 SDK 内部（T-04.8）。
- 权限：Actor 的平台权限为权限名集合（D-68～D-72，`sdk.Permission`、`sdk.ChatAdmin`）；Core Admin 是 `core_config.json` 的 `admins` 中的 `<adapter_id>:<external_user_id>`（D-105）；发信许可缺省为全部实例（D-83）。
- 根包（PR #163 之后）：`alonabot.Start(ctx, DataDir, Plugin, Adapter, Listen, ...)`；Start 读取 `core_config.json`，开发 Adapter 按 `adapters.<adapter_id>` 中 `package` 属于它的实例启动；开发组件的配置为 `plugins/<id>/config.json`，数据库为 `dev.sqlite`（T-13.2）。
- `internal/inproc` 已有 `Launch`（同进程启动参数，含 `Log`）；Router 没有「一个事件的分发结束」回调，由本切片加入。
- 扩展声明：`sdk/adapter` 的 `PermissionDeclaration`、`Event[T](name, version)`、`Extension[T](name, version)`、`Action[P, R](name, version, scope)`，`sdk.AdapterEvent[T](name, version)` 为插件侧的扩展事件类型。

```text
/goal 使用 $alonabot-execute 推进 goal-81（切片 T-16.1：插件测试套件基础）。goal 目录 goal-81，分支 goal-81，worktree .worktrees/goal-81，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：新建公开包 plugintest：在 go test 中用 alonabot.Start 在临时数据目录上装配真实 Core，登记被测插件与套件自带的测试 Adapter；提供会话、带平台权限的用户、Core Admin、发信许可、Capability 子集、Adapter 声明与配置的选项；注入一条消息或一个已声明的扩展事件，在这个事件的全部分发链结束（或其中的 Handler 进入多轮等待）后返回期间产生的 Send、Edit 与 Handler 错误。分发结束由 Core 经 internal/inproc 通知，不靠等待时间。

依据：PLAN T-16 步骤 1、2 与「失败停止条件」；CONTRACTS §17、§15.5、§8（Target 准入）、§2、§2.4；DECISIONS D-47、D-68～D-72、D-82、D-83、D-105；ARCHITECTURE 的 plugintest 一行；TEST-MATRIX V-73 的基础分项。

范围（只新增或修改这些）：
- 新增 plugintest/（包 plugintest）与 plugintest/internal/probe/（测试用插件，生成文件只由生成器生成）
- internal/inproc/：新增 harness.go
- 根包 alonabot.go、assembly.go：Start 读取 internal/inproc 中登记给该数据目录的 Harness，在激活组件前写入发信许可，并把 Router 的分发结束回调接到 Harness（只加调用点）
- internal/core/router/：Options 增加分发结束回调，在每个事件的分发链结束或进入等待时调用（只加字段与调用点）
- 上述包的测试文件
允许的新跨模块边：plugintest -> 根包、-> sdk/...、-> internal/inproc、-> contracts（ARCHITECTURE 已登记）。internal/inproc 仍不 import 本 module 的任何包。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. internal/inproc（只用标准库类型）
   type Outcome struct { PluginID, RouteID string; ErrorCode int32; SafeMessage string } // ErrorCode 为 0 表示 Handler 成功
   type Dispatch struct { InstanceID, SourceEventID string; Waiting bool; Outcomes []Outcome } // Waiting 为 true 表示其中的 Handler 进入了多轮等待
   type Harness struct {
       AllowedInstances map[string][]string // plugin_id → 收紧后的 adapter_instance_id 列表；没有的插件保持缺省的全部实例
       Dispatched       func(Dispatch)      // 一个事件的全部分发链结束、且自动回复都已发送，或其中的 Handler 进入多轮等待时调用
   }
   func RegisterHarness(dataDir string, h *Harness) (unregister func()) // dataDir 为绝对路径；同一目录重复登记 panic
   func HarnessFor(dataDir string) *Harness                             // 没有登记时返回 nil
2. 根包：不新增公开符号。HarnessFor 返回 nil 时行为与之前完全相同。
3. plugintest
   const AdapterID = "plugintest"; const InstanceID = "test"; const AccountID = "bot"
   type Option func(*options)
   func Plugin(p *sdk.Plugin) Option                    // 可以多次使用
   func Config(p *sdk.Plugin, configJSON string) Option // 写入该插件的 config.json
   func Admin(userID string) Option                     // 写入 core_config.json 的 admins（<当前测试 Adapter 的 adapter_id>:<userID>）
   func AllowedInstances(p *sdk.Plugin, instanceIDs ...string) Option // 收紧该插件的发信许可；不给 instanceIDs 表示不允许任何实例
   func Capabilities(names ...string) Option            // 替换测试实例声明的能力
   func Adapter(id string, declarations ...any) Option  // 改用指定 adapter_id，并声明 sdk/adapter 的 PermissionDeclaration、EventDeclaration、ExtensionDeclaration 或 ActionDeclaration；其他类型 t.Fatal
   type Chat struct{ ... }
   func Private(userID string) Chat // chat_id 为 "private:<userID>"
   func Group(chatID string) Chat   // chat_id 为 "group:<chatID>"
   type User struct { ID, Name string; Permissions []sdk.Permission } // 该会话中的平台权限
   type Sent struct { Op string; To message.Session; Chain message.Chain; Edited message.MessageRef; Result *message.SendResult } // Op 为 "send" 或 "edit"
   func (s Sent) Text() string // 依次拼接 Chain 中 text 片段的文字
   type Outcome struct { PluginID, RouteID string; Err *errcode.Error } // Err 为 nil 表示 Handler 成功
   type Result struct { Sends []Sent; Outcomes []Outcome; Waiting bool }
   type Harness struct{ ... }
   func Start(t testing.TB, opts ...Option) *Harness
   func (h *Harness) Message(chat Chat, from User, content message.Content) Result
   func (h *Harness) Text(chat Chat, from User, text string) Result // 等于 Message(chat, from, message.Text(text))
   func (h *Harness) Event(chat Chat, from User, event sdk.EventType, payload any) Result // 只接受 Adapter 选项声明过的扩展事件

行为：
- Start：
  1) 数据目录为 t.TempDir()；写入 core_config.json：admins（Admin 选项）、adapters.<adapter_id> 的 instances.test（内容 {}，package 为测试 Adapter 的组件 ID）；写入每个 Config 的 plugins/<id>/config.json（插件 ID 取自 p.Generated().Manifest；没有生成资源时 t.Fatal）。
  2) 登记 inproc.Harness（AllowedInstances 与 Dispatched），调用 alonabot.Start(ctx, DataDir, 全部 Plugin, Adapter(测试 Adapter), Listen("127.0.0.1:0"))；返回错误时 t.Fatal。Start 返回时测试实例已启动。
  3) t.Cleanup：取消 ctx、Wait，返回错误时 t.Error；注销 Harness。
- 测试 Adapter（套件内的 sdk/adapter 实现，与被测插件一样经 loopback 走完整生命周期）：
  - 实例账号固定为 AccountID。缺省能力为 message.send、message.markdown、message.album、message.reference、message.buttons、message.forward、message.edit（version 1）；Capabilities 选项替换这一列表。Adapter 选项的声明原样放进 Descriptor。
  - Send 与 Edit：片段所需的能力未声明时由 Core 的 Capability 预检拒绝（UNSUPPORTED），不记录为已发送；其余一律 ACCEPTED，一项 PartResult 覆盖全部片段，消息 ID 从 1 起按顺序递增（十进制字符串）；每次都追加到套件的发送记录。
- Message 与 Event：
  - 构造事件：Target 为 (adapter_id, InstanceID, AccountID, chat_id)；Actor 为 from（permissions 为 User.Permissions 的名称）；platform_message_id 与 source_event_id 为套件内递增的唯一值；私聊或内容中 mention 了 AccountID 时 to_me 为 true。
  - 经测试实例交付；Core 不接纳时 t.Fatal。然后等待 SourceEventID 相同的 Dispatched 通知，返回从交付到通知之间追加的发送记录、Outcomes 与 Waiting。等待不设套件自己的超时，以测试的期限为准；同一 Harness 上并发调用 t.Fatal。
- 套件不跳过 Core 的任何检查：命令权限、Rule、Target 准入、Capability 预检、Hook 与文案都与正式运行相同。被测插件是库模式的开发组件，配置与 dev.sqlite 按 T-13.2 的规则处理。

测试（按编码规范「测试提交规则」）：
- 提交：plugintest/plugintest_test.go，用 plugintest/internal/probe（/ping 回复 pong；/admin 只允许 Core Admin；要求 sdk.ChatAdmin 的 /mod；发送带 buttons 的消息，收到 UNSUPPORTED 时改发纯文本；/config 回复配置中的值；/count 在 dev.sqlite 中计数并回复；/edit 先发送再编辑；/ask 进入多轮等待）：
  - /ping 的结果恰好一条 Sent，Text() 为 "pong"，To 为原会话。
  - /admin：普通用户被拒（没有 Sent，命令被消费），Admin 选项给出的用户成功。/mod：没有 chat_admin 权限时没有回复，有时回复。
  - 缺省能力下按钮消息 ACCEPTED；Capabilities 去掉 message.buttons 后，结果中只有纯文本那一条，Outcomes 无错误。
  - /ask 返回时 Waiting 为 true，下一条 Text 的结果是等待的回复。
  - 用 go test -race -count=2 ./plugintest/... 运行两遍都通过。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有生成资源的插件；同一 Harness 并发 Message；Start 失败（probe 的 OnStart 返回错误）后数据目录可以再次 Start；Dispatched 只在自动回复发送之后调用；Config 的值经 /config 返回；/count 连续两次回复 1 与 2；/edit 产生 send 与 edit 两条记录；去掉 message.edit 后 edit 被拒；AllowedInstances(p) 后插件的后台发送被拒；Adapter 选项声明的扩展事件经 Event 注入到插件；两个 t.Parallel 的测试各自 Start，计数互不影响。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./plugintest/... ./internal/core/router/... ./internal/inproc/...；go test -race .；go build ./...。

不做：按钮点击、schedule 触发与后台发送的读取（T-16.2）；用套件测试示例插件与开发文档（T-16.2）；LLM 替身（T-16.3）；WebUI 沙盒（v0.1 之外）。
```

## 第 13 波

### goal-9 · T-03.10 Telegram 接入开发入口与故障矩阵

状态：已完成（[PR #180](https://github.com/AlonaBot/AlonaBot/pull/180)，合并提交 `c128000`），证据见 TEST-MATRIX §3.2。合入：`manage.DependencySnapshot` 改为接收运行时 Adapter 描述（`descriptor func(string) *pb.AdapterDescriptor`）、新增 `manage.CheckDependencies`；开发入口先启动开发 Adapter，再按依赖快照检查开发插件，重启时在停止旧进程前检查依赖；Telegram renderMedia 消费 `platform_file`；示例插件 `examples/plugins/media`；`cmd/alonabot` 的 Telegram 开发入口与故障矩阵集成测试。真实 Telegram 平台未运行（维护者不提供凭证）。以下为派发时的语句。

按 main 核对：

- `internal/pluginkit/build.go`：`func BuildDev(ctx context.Context, dir, outDir string) (build Build, err error)`；`Build` 含 `Executable string`、`Manifest []byte`、`PackageManifest []byte`、`Digest []byte`、`Role, ID, AdapterID string`。`Manifest` 是工程原文；`PackageManifest` 是交给根包的包内构建元数据。构建使用 overlay，作者目录保持原样。
- `alonabot.go`：`func Start(ctx context.Context, opts ...Option) (*Core, error)`、`func Run(opts ...Option) error`、`func Executable(path string, packageManifest []byte) Option`、`func Stdio() Option`；`func (c *Core) Restart(ctx context.Context, id string) error`、`RestartInstance(ctx context.Context, adapterID, instanceID string) error`、`ReloadSettings(ctx context.Context) error`。`cmd/alonabot/dev.go` 的 `runDevelopment(dirs []string, dataDir string, reload, stdio bool, interval time.Duration, options []alonabot.Option) error` 已调用 BuildDev、Executable、Start；`commands.go` 已提供 `alonabot run [目录...] [--reload] [--stdio]`。
- `assembly.go` 的 `startInstances` 已执行 `settings.File.BindAdapter`、`PrepareInstances`、诊断记录和 `remote.Process.StartInstance`；`internal/core/settings/instances.go` 的真实签名为 `BindAdapter(adapterID, packageID string) (bool, error)`、`PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []InstanceProblem, error)`，`InstanceProblem` 含 `InstanceID, Path, Reason string`。`settings/settings.go` 的 `AdapterSettings` 还含 `Package string`、`Version int`、`Instances map[string]json.RawMessage`、`DefaultInstance string`。本切片复用这些装配。
- `adapters/telegram/telegram.go`：`func New() adapter.Adapter`；工程身份和实例声明位于 `adapters/telegram/plugin.yaml`、`adapter_config_schema.json`、`alonabot_gen.go`，薄入口为 `cmd/alonabot-adapter-telegram/main.go`。实例声明现有 `token`、`api_base_url`、`proxy_mode`、`proxy`；开发目录应传 `adapters/telegram`。
- `sdk/plugin.go`、`generated.go`：`NewPlugin() *Plugin`、`UseGenerated(p *Plugin, g Generated)`、`(*Plugin).Generated() Generated`；示例入口为 `func New() *sdk.Plugin`，main 调用 `sdk.Run(New())`。`sdk/sdk.go` 的 `(*Context).Send(m message.Content, opts ...SendOption) (*message.SendResult, error)` 与 `sdk/edit.go` 的 `(*Context).Edit(ref message.MessageRef, m message.Content) (*message.SendResult, error)` 已使用 SDK 自有类型。`examples/plugins/echo/main.go` 已用插件对象并显式绑定生成资源，但该目录尚缺 `plugin.yaml`；`examples/plugins/media/` 尚未建立。
- `sdk/adapter/adapter.go`：`Adapter.Start(ctx context.Context, req StartRequest, core Core) (Instance, error)`；`StartRequest{InstanceID string; ConfigJSON, Cursor []byte}`；`Core.Deliver(ctx context.Context, ev *message.Event, cursor []byte) (accepted bool, err error)`；`Instance.Send(ctx context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error)`。可选编辑接口在 `sdk/adapter/edit.go`。`sdk/adapter/declaration.go`、`action.go` 的声明函数为 `Event[T](name string, version uint32) EventDeclaration`、`Extension[T](name string, version uint32) ExtensionDeclaration`、`Action[P,R](name string, version uint32, scope ActionScope) ActionDeclaration`，权限声明用 `PermissionDeclaration{Name, Implies, LabelKey}`。
- `sdk/adapter_run.go` 的 `AdapterHTTP` 为 `HTTP(route func(method, path string) *RouteBuilder)`；`sdk/plugin.go` 的插件入口为 `Route(method, path string) *RouteBuilder`；`sdk/registry.go` 的终点是 `Handle(http.Handler)` 或 `HandleFunc(func(http.ResponseWriter, *http.Request))`。`sdk/permission.go` 的 `Permission` 是权限名集合，`PermissionName(name string) Permission`、`Permission.Names() []string`；`message.Actor.Permissions` 为 `[]string`。
- `internal/core/adapter/remote/remote.go`：`New(opt Options) *Process`、`(*Process).Bind(ctx context.Context, in *host.Instance, caller router.Caller) error`、`StartInstance(ctx context.Context, instanceID string, configJSON []byte) error`、`StopInstance(ctx context.Context, instanceID string) error`、`Unbind()`。`supervisor.Spec` 的 `Role state.Role`、`AdapterID`、`PluginID` 绑定可信启动身份。`internal/core/message/send.go` 的 `New` 现有末尾参数为 `requestTimeout time.Duration, hooks Hooks`；`internal/core/buttons/buttons.go` 的 `New` 现有显式 `ttl time.Duration` 参数，旧 goal 中省略它们的写法已过时。本切片以生产 assemble 复用完整装配。
- `adapters/telegram/telegramtest/server.go`、`send.go`、`media.go` 已提供 `NewServer(t testing.TB, token string) *Server`、`URL() string`、`SetMe(userID int64, username string)`、`Push(updates ...json.RawMessage)`、`Calls(method string) int`、`Offsets() []int64`、`AllowedUpdates() [][]string`、`FailNext(status int, n int)`、`Requests() []Request`、`FailMethod(method string, nth int, status int, description string)`、`ZeroMessageID(method string, nth int)`、`Drop(method string, nth int)`、`AddFile(fileID, filePath string, content []byte)`；`Request{Method string; Params map[string]any; Files map[string][]byte}`。
- `MESSAGE-MODEL §4/§6.2`、`adapters/telegram/plan.go` 与 `send.go` 一致：末尾相册带按钮在首发前返回 `UNSUPPORTED`；零 ID 对应的 PartResult 仍为 `ACCEPTED` 并置 `MessageIDsUnavailable`。相册后增加文本可承载按钮，文本请求失败时保留相册 ID。另发现 `internal/core/media/lease.go` 已把同实例 incoming_ref 转为 platform_file，而 Telegram 的 `renderMedia` 尚未处理该来源；下文提出限于此消费分支的修复。
- 根包 `pipeline_test.go` 已有 `newPipeline(t *testing.T, stdio ...bool) *pipelineHarness`、`newPipelineAt(t *testing.T, dir string, values tune.Values, stdio bool) *pipelineHarness`，后者调用 `assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error)`，可复用真实 Core/SDK 子进程、状态库及审计装配。`cmd/alonabot/dev_e2e_test.go` 已有临时工程和目录运行的测试方式。
- 旧在途快照中的 #165（goal-55，`330164f`）、#167（goal-84，`3a9b4af`）、#168（goal-81，`37b0ada`）、#169（goal-127，`4dc4bcd`）、goal-59（PR #170，`eca6310`）与 goal-69（PR #171，`3043c4e`）均已合并；#169 的 `state/schema.sql`、manage 清理清单及 #171 的根包/CLI 改动已在基线。根包测试装配还含 PR #172（goal-57，`15f9495`）的改动。
- 按维护者本次提供的在途信息：#173（goal-40）涉及 `internal/core/state/`、`internal/core/supervisor/`、`alonabot.go`、`assembly.go`；#174 涉及启动取消的 `alonabot.go`；#175（goal-83）涉及 Broker、options、agent 与 `assembly.go`；#176（goal-128）涉及 SDK 任务 API、生成器及根包本地测试；执行中的 goal-60 涉及 agent 工具循环，goal-66 涉及 logging、manage HTTP、`alonabot.go`、`assembly.go` 与 `cmd/alonabot/`。本切片新增 CLI 测试及本地复用装配时核对这些消费点，保留已合并的字段和服务；业务代码共享点主要是既有根包装配，草稿未给这些文件另加产品职责。goal-111 也规划修改 `adapters/telegram/plan.go` 的媒体来源分支，归属须在派发前确定。以上在途状态来自维护者，未联网查询 PR。

```text
/goal 使用 $alonabot-execute 推进 goal-9（切片 T-03.10：Telegram 接入开发入口与故障矩阵）。goal 目录 goal-9，分支 goal-9，worktree .worktrees/goal-9，从最新 origin/main 创建（须已包含 PR #55、#100、#108、#114、#118、#136、#160、#163、#171）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：复用已合并的 alonabot run <目录...>，同时以独立进程运行官方 Telegram Adapter、echo 与媒体示例；实例配置取自 <data_dir>/core_config.json 的 adapters.telegram.instances。补齐可运行示例，在本地 Bot API 模拟服务上用真实 Core/SDK 进程完成 PLAN T-03 步骤 8 的发送与接收故障矩阵，记录 T-03 分项证据。维护者不提供 Telegram 凭证，本切片真实 Telegram 平台验证明确记为「未运行」。

依据：
- spec/PLAN.md §3.1；spec/plan/core.md T-03.10、T-03 步骤 8，以及步骤 2～7 的既有语义、「验收与后续边界」和「失败停止条件」；T-13 步骤 4 的已合并开发运行入口。
- spec/plan/adapters/telegram.md T-03 步骤 4、6 的已实现接收/发送规则；步骤 17 中 platform_file 的既定语义仅用于下述最小衔接修复。
- spec/CONTRACTS-v0.1.md §1.2～§1.4、§2.3、§4.1～§4.2、§8、§9.1 的身份/依赖、§15.1、§15.4、§16 的既有调优参数；测试 HTTP 入口遵循 §7.2～§7.4。
- spec/MESSAGE-MODEL.md §1～§4、§5 的既有 Edit Hook、§6.1～§6.3；同实例媒体来源沿用 §6.4 的 platform_file。
- spec/WIRE-FORMAT.md §2、§3.2～§3.5、§4、§5/§5A 的既有消息、媒体、按钮与 Adapter 字段；本切片不分配新字段号。
- spec/adr/DECISIONS.md D-26、D-68～D-72、D-77、D-79、D-82、D-83、D-86、D-88、D-102、D-104、D-107、D-109、D-112、D-114、D-117、D-134、D-137；2026-11-13 的 ReloadSettings/Executable 参数补齐记录。
- spec/TEST-MATRIX.md V-07、V-08、V-09、V-35、V-44、V-45、V-48、V-51～V-54 的 T-03 分项与 §3 的证据格式。V-89 的 platform_file 仅记本次实际衔接分项，其余由 T-03.20 汇总。

范围（只新增或修改这些）：
- examples/plugins/echo/main.go；新增 examples/plugins/echo/plugin.yaml；仅经现有生成器更新 examples/plugins/echo/alonabot_gen.go：保留文本回送，增加入站图片的 incoming_ref 原样回送，并补齐目录运行需要的工程身份。
- 新增 examples/plugins/media/main.go、plugin.yaml、alonabot_gen.go、assets/sample.png、i18n/en-US.json、i18n/zh-CN.json：一个上传内置图片、发送相册与末尾文本按钮、处理按钮并编辑文本的最小示例。alonabot_gen.go 由现有生成器产生。
- 新增 examples/plugins/media/README.md：给出 core_config.json 的实例配置与同时运行两个示例的命令，使用环境变量占位，不写实际凭证。
- adapters/telegram/plan.go：只补 renderMedia 对 Core 给出的 PlatformFile 的消费，原样作为 Bot API file_id；保留现有计划、预检、顺序与错误处理。审查者已确认由本切片完成，goal-111 相应扣除。
- adapters/telegram/telegramtest/server.go、send.go、media.go：只在现有故障注入/记录不足以表达矩阵时补齐内部实现，复用现有公开方法，不新增公开 API。
- 新增 cmd/alonabot/telegram_e2e_test.go（//go:build unix）：一个目录运行主路径集成测试，复用该包既有 Core 子进程和临时目录辅助。
- 本地、忽略、不提交：根包 telegram_local_test.go、cmd/alonabot/telegram_local_test.go、adapters/telegram/telegram_local_test.go；按需在 telegramtest/ 下写 *_local_test.go。根包本地测试复用 pipeline_test.go 的装配，临时 SDK 进程工程只写 t.TempDir()。
- spec/TEST-MATRIX.md §3.2：追加本切片的证据行；本切片仅获授权修改这一节。

允许的新跨模块边：cmd/alonabot 的集成测试 -> adapters/telegram/telegramtest；根包的 *_local_test.go -> adapters/telegram、adapters/telegram/telegramtest、internal/pluginkit（均仅测试装配）。这些测试边已由审查者确认，本切片获授权把它们登记进 ARCHITECTURE.md 的依赖表（只改这一处）。示例沿用 examples/plugins/... -> sdk/...，Telegram 沿用 adapters/... -> sdk/...；产品代码不新增跨模块边。

接口（照此实现；除新示例的 New 外，复用 main 公开接口）：
1. 开发运行与 Core：
   func alonabot.Start(ctx context.Context, opts ...alonabot.Option) (*alonabot.Core, error)
   func alonabot.Executable(path string, packageManifest []byte) alonabot.Option
   func alonabot.Stdio() alonabot.Option
   func (c *Core) Restart(ctx context.Context, id string) error
   func (c *Core) RestartInstance(ctx context.Context, adapterID, instanceID string) error
   func (c *Core) ReloadSettings(ctx context.Context) error
   func pluginkit.BuildDev(ctx context.Context, dir, outDir string) (build pluginkit.Build, err error)
   type pluginkit.Build struct {
       Executable string
       Manifest, PackageManifest, Digest []byte
       Role, ID, AdapterID string
   }
   手工装配测试时传 alonabot.Executable(build.Executable, build.PackageManifest)。CLI 已有 --reload 与 --stdio，配置准备、实例恢复与热部署继续由现有 Core 所有者执行。

2. 设置与 Adapter 进程：
   func settings.Load(dataDir string) (*settings.File, error)
   func (f *File) BindAdapter(adapterID, packageID string) (bool, error)
   func (f *File) PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []settings.InstanceProblem, error)
   func (p *remote.Process) Bind(ctx context.Context, in *host.Instance, caller router.Caller) error
   func (p *remote.Process) StartInstance(ctx context.Context, instanceID string, configJSON []byte) error
   func (p *remote.Process) StopInstance(ctx context.Context, instanceID string) error
   func (p *remote.Process) Unbind()
   func manifest.CheckDependencies(m *manifest.Manifest, in manifest.Installed) []manifest.Problem
   core_config.json 示例结构：
   {"adapters":{"telegram":{"package":"alonabot-adapter-telegram","$version":1,"default_instance":"tg_a","instances":{"tg_a":{"token":"${TELEGRAM_BOT_TOKEN_A}","api_base_url":"https://api.telegram.org","proxy_mode":"none"},"tg_b":{"token":"${TELEGRAM_BOT_TOKEN_B}","api_base_url":"https://api.telegram.org","proxy_mode":"none"}}}}}
   模拟测试把 api_base_url 换成各自 Server.URL()；两个 Server 用不同 token 和 SetMe 账号。schema 来自 adapters/telegram/adapter_config_schema.json；配置归属与 PrepareInstances 已在 assembly.go 中接通。

3. 插件对象、消息与媒体：
   两个示例各提供 func New() *sdk.Plugin，New 中 p := sdk.NewPlugin()，sdk.UseGenerated(p, alonabotGenerated)；main 调用 sdk.Run(New())。
   func (p *Plugin) On(routeID string, eventType sdk.EventType) *sdk.EventBuilder
   func (p *Plugin) Command(path ...string) *sdk.CommandBuilder
   p.On(..., sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error { ... })
   p.Command("media").Handle(func(c *sdk.Context) error { ... })
   func (c *Context) Send(m message.Content, opts ...sdk.SendOption) (*message.SendResult, error)
   func sdk.Send(ctx context.Context, to message.Session, m message.Content, opts ...sdk.SendOption) (*message.SendResult, error)
   func sdk.To(s message.Session) sdk.SendOption
   func sdk.Via(instanceID string) sdk.SendOption
   func (c *Context) Edit(ref message.MessageRef, m message.Content) (*message.SendResult, error)
   func sdk.Upload(ctx context.Context, r io.Reader, mime string) (*message.MediaRef, error)
   func sdk.OpenMedia(ctx context.Context, src message.MediaSource) (io.ReadCloser, error)
   func sdk.GrantMedia(ctx context.Context, token, granteePluginID string, sessions ...message.Session) (string, error)
   func (m *message.MediaRef) Source() message.MediaSource
   func (m *message.IncomingMedia) Source() message.MediaSource
   message.Content 由 message.Segment 和 message.Chain 实现；入站图片用 message.ImageSegment.Source，其中 Incoming 为 *message.IncomingMedia。会话用 message.Session，编辑与引用用 message.MessageRef（result.Ref()/Refs()、event.Ref()），引用构造用 message.Quote(ref)。

4. 按钮、权限与可选 HTTP：
   func sdk.Action[T any](name string) sdk.ButtonAction[T]
   func (a ButtonAction[T]) Button(label string, value T) message.Button
   func sdk.OnButton[T any](r sdk.Registrar, a sdk.ButtonAction[T]) *sdk.ButtonBuilder[T]
   func (b *ButtonBuilder[T]) Handle(h func(*sdk.Context, T) error)
   func message.Buttons(rows ...[]message.Button) message.Segment
   路由 Permission 使用 sdk.Anyone、sdk.User、sdk.Admin、sdk.ChatAdmin 或 sdk.PermissionName(name)，Actor.Permissions 为权限名集合；Core Admin 取 admins 中的 "telegram:<用户 ID>"。
   测试 Adapter 按需实现：
   type sdk.AdapterHTTP interface { HTTP(route func(method, path string) *sdk.RouteBuilder) }
   route("POST", "/inject").HandleFunc(handler) 声明私有路由，由真实管理认证调用；插件的私有观测路由用 p.Route(method, path).Handle/HandleFunc。SDK 托管 Listener 和内部 Session 凭证。
   测试声明只使用 sdk/adapter.PermissionDeclaration，以及 adapter.Event[T](name, version)、Extension[T](name, version)、Action[P,R](name, version, scope)；按钮动作使用 sdk.Action[T](name)。

5. 既有 wire 字段（仅核对、使用）：
   contracts/proto/alonabot/v1/adapter.proto：AdapterInstanceStartRequest 的 instance_id=1、config_json=2、cursor=3；AdapterInstanceStopRequest.instance_id=1；AdapterDeliverRequest 的 instance_id=1、event=2、cursor=3；AdapterDeliverResponse.accepted=1。
   identity.proto：Target 的 adapter_id=1、adapter_instance_id=2、account_id=3、chat_id=4、optional thread_id=5；Actor 的 repeated string permissions=6（trusted_roles 已 reserved）。
   message.proto：Button.optional string core_token=4 由 Core 填写；SendResult.part_results=7；PartResult 的 platform_message_ids=4、message_ids_unavailable=5、error=6。
   这些字段由现有转换层消费；插件和 Adapter 作者代码使用 sdk/message 与 sdk/errcode。

行为：
- 示例可运行：在仓库根用 alonabot run adapters/telegram examples/plugins/echo examples/plugins/media --data-dir <临时数据目录>。echo 的文本行为保留；收到图片时以其原 Incoming 来源回送，保留图片说明与片段顺序，不读文件再上传来绕过来源语义。media 通过 sdk.Upload 上传嵌入图片（明确覆盖上传链路），用返回的 Source 构造连续图片相册；相册之后放一段文本承载按钮。按钮回调使用带类型的值，编辑文本消息的句柄取 result.Refs() 的末条。面向用户的固定文字与按钮标签经 i18n，错误和日志保持英文。
- platform_file 衔接：renderMedia 识别 Core 传入的非空 PlatformFile，原样填写单媒体/相册的 media 参数，不取件、不 base64 转换。插件直接给出 PlatformFile 仍由现有 Core 输入校验拒绝。只完成同实例图片回送必需的消费分支，保留其他来源行为。
- 生命周期沿用 Core：先持有 core.lock，再启动/激活组件、Bind Adapter、按归属准备并启动实例。测试用目录参数控制所需启动先后；全局停止沿用 Core 的共用停止起点、冻结和排空。一个实例的启停保持 Adapter PID 与其他实例不变；异常恢复只启动当前仍启用的实例。以观察插件记录 bot.online/offline；记录生命周期时保证观察插件已注册到对应事件路由。
- 配置：token、proxy 的原值或 ${...} 写法留在 core_config.json；PrepareInstances 的运行副本展开后仅经 adapter.instance.start 送到所属 Adapter。插件进程的环境、Bootstrap、事件、回执和日志均不含平台 token/proxy；测试使用自己生成的虚构凭证。配置失败记录 InstanceProblem 的路径与原因，其他可启动实例照常启动。
- 故障矩阵全部以本地 Bot API 模拟服务和真实 Core/SDK Session 完成；以下 1～13 的结果逐项留在 PR 与 TEST-MATRIX §3.2 的分项记录中：
  1) 两个不同账号实例分别启动、停用、恢复，核对 PID、实际发送账号及实例隔离；停用原因是 STOPPED。另验证已停用的实例在 Adapter 崩溃恢复时保持停用（CONTRACTS §8）。
  2) 两实例运行时杀掉测试拥有的 Adapter 进程：旧 Session 各产生一条 ADAPTER_EXITED 的 bot.offline，恢复后各一条 bot.online；从已保存游标继续，不重复交付。生命周期事件无 Actor/Target；Core 重启不补发历史上下线事件。
  3) 用真实 SDK 回环 Adapter 的 Deliver 交付错误 adapter_id、instance_id、account_id，或用另一 Adapter Session 试图交付不属于它的实例事件，Core 拒绝并记诊断，目标插件没有收到事件。保留官方 Telegram 实例的正常对照；测试控制只在本地测试进程中实现。
  4) 论坛 is_topic_message=true 的消息回复带正确 message_thread_id；普通超级群回复串与无 thread_id 的 General 话题按 MESSAGE-MODEL §6.1 处理。账号和话题的原 Target、显式 To/Via、收紧/撤销实例许可逐项核对；后台无目标返回 NO_DEFAULT_TARGET。
  5) 一条 Chain 产生多个物理消息，与两次独立 Send 分别记录逻辑调用数、Server.Calls 的增量及 Requests 的物理顺序；相邻文本、首条引用、末条按钮按计划执行，PartIndex/SegmentIndexes 与确认 ID 全部保留。
  6) 第一个物理请求完整到达模拟服务后用 Drop 断开响应：逻辑 UNKNOWN、该 part UNKNOWN，其余 NOT_ATTEMPTED；此前已有成功 part 时为 PARTIAL 并保留 ID。另用 FailMethod 覆盖平台明确拒绝的 REJECTED。各项均保留 payload+error，以后续实际调用数证明没有自动补发。
  7) 末尾相册加按钮：UNSUPPORTED/REJECTED，首个平台发送请求增量为 0。合法的「相册→带按钮的尾随文本」计划中，让尾随文本请求失败，结果 PARTIAL，保留相册 ID，sendMediaGroup 恰调用一次。另以「相册→文本→文件」计划让文本失败，核对文件 part 为 NOT_ATTEMPTED。此项覆盖旧语句所谓「相册后 markup 失败」的部分成功意图，按当前平台可表达的计划取证。
  8) ZeroMessageID 注入成功的单条/相册请求：对应 part ACCEPTED、MessageIDsUnavailable=true，ID 列表不含 "0"，所有计划成功时逻辑 ACCEPTED。缺少可用 ID 时不调用 Edit、不伪造句柄、不补发；若同一计划后续操作失败，整体为 PARTIAL，逐 part 保留具体状态。
  9) 入站图片在 Read 前 getFile 增量为 0；两个真实插件 Session 的不同 token 读取同一小文件，仅一次 getFile，内容相同且缓存共享。另验证同实例 incoming_ref 回送直接使用 file_id，getFile 增量仍为 0；Grant 的合法/拒绝路径、token 过期、源实例许可收紧、越界 Target、配额 RESOURCE_EXHAUSTED、首个读者取消与 owner 退出后的租约按 CONTRACTS §1.3～§1.4、§8 取证。配额/TTL 使用既有 tune.Values 或 --args，值和命令写进证据。
  10) 在持久游标已提交、Handler 尚未完成的位置以审计同步点杀掉测试 Core 子进程，清理它拥有的其余进程并等待退出后，在同一临时 data_dir 重启。核对 Offsets 与业务计数，已接纳 update 不再交付；允许该条未完成业务丢失。不得只用「等一段时间」猜测游标提交。
  11) getUpdates 401 与 409（FailNext）：接收实例停止并产生 INSTANCE_FAILED 的 bot.offline；其他实例继续运行。启动时已有 webhook、getMe 账号重复的拒绝，以及显式 allowed_updates、逐条游标、匿名 sender_chat、7 天接收轮次重建，复用既有证据并对本次有改动或缺失的项做本地补证。
  12) callback_query：先有 answerCallbackQuery 请求再执行对应发送插件的带类型按钮 Handler；非法/过期 token、其他 Target、inline-mode 缺少可核验 Target 时只拒绝和诊断。成功编辑及编辑带按钮用 editMessageText，保留原 ID；平台拒绝、不支持、含媒体的 Edit 按现有契约返回并保持 Send 请求数不变。每次 Edit 的既有出站 Hook 行为按生产装配验证。
  13) Actor.locale 与 to_me：language_code 的 zh-CN/zh-TW/缺失等映射，私聊、回复机器人、mention/text_mention 各按契约；权限验证使用现行权限名集合。开发入口的必需/可选插件依赖、版本条件与 Adapter extension 依赖沿用 CheckDependencies，缺失必需项在启用前拒绝；不满足目标 Adapter 的 extension 为 UNSUPPORTED。
- 请求数量以每个场景开始/结束的 Calls 差值与完整 Requests 对照，区分正常 long polling 和业务发送。现有 FailNext 已能注入 getUpdates 409，Drop 已能表达请求到达后的响应断开；只有取证确有缺口时才修改模拟器内部记录，保持模拟器与当前契约一致。
- 证据追加为「T-03.10 / 对应 V 分项」行，包含候选版本、测试位置、完整命令、实际 GOOS/GOARCH、Go 版本、传输、进程、模拟服务版本/故障方式、结果及日志位置。真实 Telegram 明确写「未运行：维护者未提供 Telegram 凭证和授权测试会话」；既有历史实测记录保留其原版本和范围。本次模拟通过后，仅更新实际完成的 T-03 分项。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：cmd/alonabot/telegram_e2e_test.go 的一个 TestTelegramDevelopment（unix）。通过真实 CLI 的 run 目录入口构建并运行官方 Telegram Adapter、echo、media，测试临时 data_dir 与两个账号的模拟服务：文本/话题回复、incoming 图片原样回送、Upload→相册→尾随文本按钮、点击后编辑原文本消息，以及带按钮尾随文本失败后保留相册且不重发。断言 Bot API 请求顺序、参数、上传字节、实例隔离和最终业务效果；需要检查回执的分项由本地多进程矩阵取得。理由：长期守住本切片新增的目录示例接入与跨进程媒体/按钮主路径，一个测试内用子测试组织。
- 本地自测：根包 telegram_local_test.go 的 TestTelegramFaultMatrix（unix）完成上述矩阵的其余分项，TCP 与 stdio 各一遍；复用 pipeline_test.go 的 newPipeline/newPipelineAt、assemble、真实 SDK 进程和 SQLite。必要的测试模式以现有 TestMain 在 m.Run() 中执行本地辅助 Test 的方式接入，或在 t.TempDir() 写临时工程经 BuildDev 构建；不修改现有 TestMain 的分流与断言。正常平台路径用官方 Adapter；伪造归属等攻击输入用真实 SDK 回环 Adapter，记录为 Core 信任边界证据。
- 本地自测：cmd/alonabot/telegram_local_test.go 覆盖目录身份/配置失败、非法 JSON、实例迁移与校验失败、空实例集、单个失败实例不影响其他实例；--reload 下配置变化只重启指定实例，Adapter PID、其他实例与插件 PID 保持；手工根包装配传递 PackageManifest，并核对 Restart/RestartInstance/ReloadSettings 的既有职责。
- 本地自测：adapters/telegram/telegram_local_test.go 覆盖新增 PlatformFile 消费及保留的媒体来源分支、当前相册按钮拒绝和零 ID 断言；telegramtest 的记录/故障辅助若改动，在该包 *_local_test.go 自测。
- 不向 sdk/、internal/core/supervisor/ 提交测试，不新增 testdata 测试插件；本地测试、辅助进程和临时工程不提交，PR 写明实际命令、平台与结果。真实 Telegram 在本切片固定记为「未运行」。

完成条件：
- echo/media 均能从工程目录由现有 run 构建并运行；所有新生成资源由现有生成器产生，生成一致性检查覆盖本次修改的示例。
- gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。记录主路径与本地矩阵的实际测试命令；需要单独复现时分别用 go test -race -count=1 ./cmd/alonabot -run '^TestTelegramDevelopment$' 和 go test -race -count=1 . -run '^TestTelegramFaultMatrix$'，本地辅助 Test 的实际运行方式同时记录。已有同一候选的通过结果可复用。
- 上述 1～13 每项有本次结果或适用的既有证据；新增链路与发送/接收故障必须经真实 Core/SDK 进程取证，未执行的分项写「未运行」并说明原因。现有证据只在实现与输入仍相同时复用。
- TEST-MATRIX §3.2 有本次 T-03 分项证据及真实 Telegram「未运行」记录；PR 附本地测试结果、提交测试新增行数、故障日志和待审查者同步的文档清单。
- 对照 T-03「失败停止条件」：凭证串用/泄漏、其他 Session 实例事件被接纳、实例启停牵连整个进程、成功被误报未发送、部分结果丢 ID、已知不支持却先发消息、失败后补发、媒体越权/超预算、游标重放或重复 owner，任一出现均须停止验收并修复；修复涉及范围外文件、proto、依赖或契约冲突时标阻塞交审查者，不用删断言或降级示例解除。

不做：重新实现 run/热部署或增加 --adapter-id 等 CLI 参数；Core 根包装配/启动取消/停止状态机改造；候选 Journal、schema.sql 和 manage 清理清单；Agent、Broker/options、SDK 任务 API 与日志管理；正式安装/卸载和管理流程的重复实现；T-03.20 的其他媒体功能、file_base_url、流式 multipart、voice/GIF/url 按钮与其他平台 Adapter；Service required_permissions 的新 SDK API；真实 Telegram 验证；除获授权的 TEST-MATRIX §3.2 外的 spec、PLAN、ADR、ARCHITECTURE 修改。
```

### goal-131 · T-03.25 聊天目录

状态：已完成（[PR #166](https://github.com/AlonaBot/AlonaBot/pull/166)，合并提交 `6f72fc8`），证据见 TEST-MATRIX §3.2。合入：Event 的 `chat_title = 16` 与 `message.Event.ChatTitle`；状态库 chat 表、`Store.PutAdapterCursor(ctx, pluginID, instanceID, cursor, chat)`、`Store.Chats(ctx, ChatQuery)`；`options.Options.Chats` 与内置来源 core.chats；Adapter 的清理对象加入聊天目录。派发时：前置 T-03.3（PR #32）、T-08.14（[PR #157](https://github.com/AlonaBot/AlonaBot/pull/157)，合并提交 `e709135`）与 T-09.4（[PR #163](https://github.com/AlonaBot/AlonaBot/pull/163)，合并提交 `cb8ae3d`）已合并。2026-11-10 按 D-130 新增；2026-11-12 审查者按 main `d6eb381` 写出范围、接口与测试段，要点如下：
- 协议：Event 在 `contracts/proto/alonabot/v1/message.proto`，最后一个字段为 `ChatType chat_type = 15`，`chat_title` 用 16；SDK 侧 `sdk/message.Event` 与 `sdk/internal/convert` 同步一个字段，`sdk/adapter` 的 `Core.Deliver(ctx, ev *message.Event, cursor)` 不变。
- 交付：`internal/core/adapter/remote/services.go` 处理 `AdapterDeliverRequest` 时先校验 Target 与 Actor，再 `Store.PutAdapterCursor(ctx, pluginID, instanceID, cursor)`，之后重建可信字段并派发。聊天记录与游标写在同一个状态库事务中。
- 选项来源：`configschema.BuiltinOptionSources` 已登记 `core.chats` 与 filter 键 `chat_type`、`adapter_id`；`internal/core/options` 的内置来源目前对它返回 UNAVAILABLE，本切片实现。
- 清理：`internal/core/manage/lifecycle.go` 的卸载按 `PurgeScope` 列出并删除受管对象，状态库对象在一个事务内删除。
- 本切片修改 Event 的 proto 与 `internal/core/state/schema.sql`，按执行顺序的共享文件规则处理冲突（goal-55 也在改 schema.sql 的按钮表）。

```text
/goal 使用 $alonabot-execute 推进 goal-131（切片 T-03.25：聊天目录）。goal 目录 goal-131，分支 goal-131，worktree .worktrees/goal-131，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

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

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/message.proto：Event 追加 `optional string chat_title = 16`，随后 go generate ./...；contracts/wirespec_test.go 追加编号检查
- sdk/message 与 sdk/internal/convert：Event 增加 ChatTitle 字段与转换
- internal/core/state/：schema.sql 增加 chat 表（写进唯一的初始迁移）；新增 chats.go；PutAdapterCursor 同步签名
- internal/core/adapter/remote/services.go：交付时校验并写入聊天
- internal/core/options/：实现内置来源 core.chats
- internal/core/manage/：把聊天表加入 Adapter 的清理对象
- internal/core/adapter/remote/remote_test.go：追加集成测试 TestChatDirectory
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
   // internal/core/state
   type Chat struct { AdapterID, ChatID, ChatType, Title, InstanceID string; FirstSeen, LastSeen time.Time } // ChatType 为 private、group 或 channel，空串表示未知
   func (s *Store) PutAdapterCursor(ctx context.Context, pluginID, instanceID string, cursor []byte, chat *Chat) error // chat 为 nil 时只保存游标；否则在同一事务中写入或更新聊天
   type ChatQuery struct { Search, ChatType, AdapterID string; Keys []ChatKey; Limit int }
   type ChatKey struct { AdapterID, ChatID string }
   func (s *Store) Chats(ctx context.Context, q ChatQuery) (chats []Chat, more bool, err error) // 按 LastSeen 从新到旧
   清理时删除某个 adapter_id 的聊天：在 manage 已有的清理事务中执行（state 提供按 adapter_id 删除的方法，名称自定）。

行为：按上文要点。补充：
- 没有 target 的事件带 chat_title 时，Core 以 INVALID_ARGUMENT 拒绝这次交付并记 WARN（只写 adapter_id 与 chat_id）。
- 写入：Title 为空时保留已有名称，ChatType 未知时保留已有类型，InstanceID 与 LastSeen 每次更新；Core 生成的 bot.online、bot.offline 与 scheduled 不写。
- core.chats：存储值为 `<adapter_id>:<chat_id>` 会话字符串（转义规则同 MESSAGE-MODEL §1，复用已有的会话字符串编码），显示名为 Title（为空时为 chat_id），说明为 adapter_id 与 chat_type；Search 不区分大小写地匹配名称与 chat_id；按值取回时格式不合法或不存在的值不返回；上限 OPTIONS_PAGE，超出时 More。
- 卸载 Adapter：--purge-data 与 --purge 删除该 adapter_id 的聊天并在确认对象中列出条数；只卸载与 --replace 保留。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/adapter/remote/remote_test.go 的 TestChatDirectory（经回环测试 Adapter 进程）：群与私聊各交付一条带名称的事件后，core.chats 能按名称搜到、按 chat_type 筛选；之后一条不带名称的事件不清空名称。contracts/wirespec_test.go 的编号检查。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有 target 却带名称被拒；Core 重启后保留；两个实例交付同一聊天时 InstanceID 更新为最近的一个；--purge-data 后删除而只卸载保留；按存储值取回；聊天名称不出现在日志中。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./...；go build ./...。

不做：给插件的聊天列表服务；读取 `chat_title` 的 SDK 方法；话题的目录；成员列表；主动向平台拉取全部群；管理 API（goal-93）与页面（goal-44）；两个官方 Adapter 的填写（goal-132、goal-133）。
```

### goal-18 · T-05.5 多进程事件流水线验收

状态：已完成（[PR #162](https://github.com/AlonaBot/AlonaBot/pull/162)，合并提交 `8c12972`），证据见 TEST-MATRIX §3.2。合入：根包 `pipeline_test.go`（unix）的 `TestPipeline`，经 assemble 取得生产装配、以测试二进制运行流水线插件与回环 Adapter 进程，覆盖矩阵 1～10；不修改产品代码。验收矩阵第 2、3、5 项已按契约改正（main `1f17170`）。派发时：前置 T-03.8（[PR #136](https://github.com/AlonaBot/AlonaBot/pull/136)，合并提交 `3621fc1`）、T-05.3（[PR #155](https://github.com/AlonaBot/AlonaBot/pull/155)，合并提交 `53393aa`）与 T-05.4（[PR #153](https://github.com/AlonaBot/AlonaBot/pull/153)，合并提交 `f22b538`）已合并。2026-11-12 审查者按 main `2a1c504` 改写：
- 测试装配：`cmd/alonabot` 的 run 目前不带 Router 与 Adapter（goal-39 改写），所以不经 cmd 与 Telegram 模拟服务。根包的 `assemble(store, dir, tune.Values)` 给出与生产相同的 Router、hook.Runner、continuation.Manager、buttons、policy 与 message 服务；插件与 Adapter 进程沿用 `internal/core/adapter/remote/remote_test.go` 的写法：测试二进制按环境变量以 `sdk.Run` 或 `sdk.RunAdapter` 运行，经 `supervisor.Launch` 启动，Adapter 用回环测试 Adapter。
- T-07.5（[PR #119](https://github.com/AlonaBot/AlonaBot/pull/119)）已合并，scheduled 用真实 Scheduler 验收，不再只用 fixture。
- 按钮点击前的 answerCallbackQuery 只在 Telegram Adapter 中，由 goal-9 的故障矩阵覆盖；本切片以回环 Adapter 交付 button.clicked。
- SDK 写法：插件对象（`sdk.NewPlugin`、`Command`/`On`/`PreEvent`/`BeforeSend`/`AfterSend` builder）、`sdk.Action[T]` 与 `sdk.OnButton`、`Context.Continue` 与 `Context.Wait(timeout)`、`sdk.ToMe` 与自定义 Rule。
- 与 goal-39 同改根包：本切片只新增测试文件，goal-39 先合并时按合并后的装配入口调整测试。

```text
/goal 使用 $alonabot-execute 推进 goal-18（切片 T-05.5：多进程事件流水线验收）。goal 目录 goal-18，分支 goal-18，worktree .worktrees/goal-18，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：用多个真实插件进程与回环 Adapter 进程跑完整的事件流水线：多个 priority 与同优先级逆安装顺序、自定义 Rule、两种 Handler 签名、三个 Hook、按钮回调、两个用户在同一群同时等待、等待超时与等待期间停用插件、真实 Scheduler 触发；用恶意载荷尝试修改可信字段，用延迟 Handler 验证控制帧仍推进；在 TEST-MATRIX §3.2 记录 T-05 证据。

依据：PLAN T-05 步骤 8 与「停止和失败条件」；CONTRACTS §2；MESSAGE-MODEL §1、§6.3；TEST-MATRIX V-11、V-13～V-16、V-43、V-49、V-50 的 T-05 分项。

范围（只新增或修改这些）：
- 新增根包 pipeline_test.go（//go:build unix）：经 assemble 取得生产装配，测试二进制按环境变量运行流水线插件与回环 Adapter
- spec/TEST-MATRIX.md §3.2：追加 T-05 的证据行（本切片获授权修改这一处）
允许的新跨模块边：无。

接口：不新增公开 Go 符号；不修改产品代码。测试发现产品缺陷时按 alonabot-execute 的阻塞规则登记，不在本切片修。

验收矩阵（pipeline_test.go 中各为一个子测试，以 Adapter 收到的发送与编辑、插件写出的调用次序记录与 HandlerOutcome 为证据）：
1) 三个插件 priority 为 0、0、-1，以两种不同的安装与启动顺序运行：调用顺序都是 (priority, plugin_id, route_id)。
2) 自定义 Rule 为 false 与出错：事件路由跳过；已识别的命令被消费且不回落到短路径、文本匹配与 Raw，没有 Stop 时 message.created 仍进入事件路由，StopPropagation 后不进入（CONTRACTS §2.2 第 3 条，D-44；V-11 的真实 Rule 分项）。
3) 两种 Handler：自动回复只发一次；手动多次 Send；自动回复型 Handler 内的显式 Send 照常发送，自动回复排在其后（MESSAGE-MODEL §3，D-70）。
4) Hook：PreEvent 改写载荷后后继路由看到改写结果；BeforeSend 改写链与拒绝发送；AfterSend 出错不改变发送结果；每次 Send 经过一组 Hook。
5) 恶意载荷：回环 Adapter 交付的事件与 PreEvent 的返回尝试改 Actor、Target：插件收到的仍是 Core 保留的值；PreEvent 对 button.clicked 与 scheduled 返回载荷（改 action/value 或 job_id）时按 Hook 失败终止入站链并记 WARN，没有 Handler 收到伪造值，去掉该 Hook 后原绑定照常工作（CONTRACTS §2.3）。
6) 按钮：发送带 `sdk.Action` 按钮的消息后，回环 Adapter 交付点击：调用正确的插件与动作；另一用户点击也调用一次并按该用户鉴权；过期 token 不调用、不落入 Raw。
7) 等待：两个用户在同一群同时 Wait，各自只收到自己的下一条消息，这些消息不进入普通分发；超时为 DEADLINE_EXCEEDED；等待期间停用插件，重新启动后不补发。
8) 延迟 Handler（睡眠数秒）执行期间，同一插件 Session 的另一请求的 Response 与 Cancel 正常完成，其他事件照常派发。
9) 权限组合（V-43）：chat_admin 不取得 Core Admin；组禁用优先于子项启用；最近显式权限覆盖父项；运行时修改在下一次派发生效，已开始的调用不被撤销。
10) scheduled：插件声明 Every 短间隔任务，真实 Scheduler 触发后插件收到 Core 的 job_id；伪造的 job_id 不改变它。
- 证据：TEST-MATRIX §3.2 追加 T-05 一行，写明测试位置、命令、平台与结果。

测试（按编码规范「测试提交规则」）：
- 提交：pipeline_test.go 的上述矩阵（多列理由：T-05 的验收矩阵本身，没有其他端到端入口）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：在 -race 下以 -count=10 运行矩阵无偶发失败。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：Service 嵌套与 Hook→Service→Send（T-06.2、T-06.4）；Telegram 的 answerCallbackQuery 顺序（T-03.10）；修改产品代码。
```

### goal-20 · T-06.2 父调用链、取消与背压

状态：已完成（[PR #161](https://github.com/AlonaBot/AlonaBot/pull/161)，合并提交 `b4610e1`），证据见 TEST-MATRIX §3.2。合入：`broker.New(broker.Options{Router, RequestTimeout, MaxDepth})`；`router.Invocation` 的 `Chain`、`Scopes` 与导出的 `TraceID`；Broker 从父附件继承 actor、default_target、scopes、send_forbidden 与 trace，链为父链加调用方，CALL_CYCLE 与深度检查在调用目标前返回；Hook 调用的附件继承父链加父插件、scopes 与 trace，Scheduler 的父附件带本次 tick 的 trace。父请求结束时的取消与背压由 wire 的现有机制提供，以三进程场景验证。执行中审查者授权导出 TraceID 并修改 hook 与 scheduler。派发时：T-05.3（[PR #155](https://github.com/AlonaBot/AlonaBot/pull/155)，合并提交 `53393aa`）、T-03.6（PR #55）、T-06.1（PR #45）已合并。2026-11-12 审查者按合并后的 main 核对：
- `broker.New(r *router.Router, requestTimeout time.Duration)`，调用方为根包 assembly.go（goal-130 的选项查询也使用它）；`Broker.Call` 已从 `in.ParentAttach.(*router.Invocation)` 派生 Actor 与默认 Target，`InvocationContext.plugin_chain` 目前只有调用方一跳，`send_forbidden` 与 `scopes` 没有继承；期限已取 REQUEST_TIMEOUT、调用方期限与 ctx 期限中最早者；`b.calls` 以子调用的 Invocation 记录调用来源。
- `router.Invocation{PluginID, Event, DefaultTarget, SendForbidden, RouteID, Continuation}`；`SendForbidden` 由 hook 包在出站 Hook 调用时设置，message 的 send、edit、delete、action 检查父附件的这个标记（goal-17）。
- `tune.Values.CallChainMaxDepth`（CALL_CHAIN_MAX_DEPTH，缺省 16）已登记。
- 测试插件按 main 上的插件对象写法（`sdk.NewPlugin`、Plugin.BeforeSend 等）编写；Service 的声明与调用使用合并时 main 上的写法（goal-126 先合并时为 `sdk.JSON[In, Out](target)` 与 `Call`）。

```text
/goal 使用 $alonabot-execute 推进 goal-20（切片 T-06.2：父调用链、取消与背压）。goal 目录 goal-20，分支 goal-20，worktree .worktrees/goal-20，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Broker 沿父调用链传播 Core 记录的可信上下文（actor、target、scope、trace、祖先插件链、send_forbidden），不接受插件覆盖；同链重复插件返回 CALL_CYCLE，深度超过 CALL_CHAIN_MAX_DEPTH 返回 FAILED_PRECONDITION；子调用期限取父与自身预算中较早者；父完成、超时、取消或任一 Session 断开时取消未完成的子调用；复用 wire 的在途数与字节预算，额满立即失败，不出现互相等待额度的死锁。

依据：PLAN T-06 步骤 3、4；CONTRACTS §3 第 4、5 条、§1.2、§2.3（send_forbidden）；WIRE-FORMAT §4；TEST-MATRIX V-18、V-42 的嵌套分项、V-15 的间接递归分项。

范围（只新增或修改这些）：
- internal/core/broker/：调用链记录、循环与深度检查、期限与取消传播
- internal/core/router/：Invocation 增加祖先插件链与 scopes（只加字段与填写处）
- 根包 assembly.go：broker.New 的调用改为传 Options
- internal/wire/：如需在父请求结束时通知子调用取消，只加回调字段与触发点（与 goal-24 的 Late 同样的方式）
- 新增 internal/core/supervisor/testdata/chainplugin/
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- internal/core/router：Invocation 增加字段 Chain []string（祖先插件链，最早者在前）与 Scopes []string；SendForbidden 沿用
- internal/core/broker：
   type Options struct { Router *router.Router; RequestTimeout time.Duration; MaxDepth int /* tune.Values.CallChainMaxDepth */ }
   func New(opt Options) *Broker // 替换 New(r, requestTimeout)，同步全部调用者

行为：
- 父请求：插件发起的 Service 调用带 parent_call_id 时，必须指向同一 Session 中仍在途的 Core→该插件请求（wire 已保证属于同一 Session）；Broker 从该父请求的 Invocation 派生上下文：actor、default_target、scopes、trace_id、send_forbidden 原样继承，写进 InvocationContext 与子调用的 Invocation，Chain = 父 Chain + [调用方]（InvocationContext.plugin_chain 下发同一列表）。插件载荷中的任何身份字段被忽略。没有父请求时只有插件身份：Chain = [调用方]，actor 与 target 为空。
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

### goal-128 · T-07.9 动态任务与暂停恢复的 SDK

状态：已完成（[PR #176](https://github.com/AlonaBot/AlonaBot/pull/176)，合并提交 `9785c22`），证据见 TEST-MATRIX §3.2。合入：`sdk.Task[T](name) TaskType[T]` 的 `At`、`Every`、`Cron`、`Cancel`、`List`（经 core.task.set、cancel、list 创建、替换、取消与分页列出，返回 `TaskInstance[T]`），`TaskBuilder[T]` 登记任务的 Handler，`sdk.JobPause`、`sdk.JobResume`、`sdk.JobStatus(ctx, jobID) (JobState, error)`；生成器对动态任务参数类型的检查；showcase 提醒示例按可信用户与会话划分内部 key。以下为派发时的语句。

按 main 核对：

- `sdk/plugin.go`：`func NewPlugin() *Plugin`；`type Registrar interface { registrar() *pluginScope }`，`*Plugin` 与 `*SubPlugin` 实现该接口；作者入口为 `func New() *sdk.Plugin`。`sdk/sdk.go`：`func Run(root *Plugin) error`、`func Collect(root *Plugin) (*Registration, error)`；Registration 的任务部分目前只有 `Schedules []Schedule`，动态任务尚无 SDK 作者入口。
- `sdk/plugin.go`、`sdk/subplugin.go`、`sdk/registry.go`：`Every(jobID string, d time.Duration) *ScheduleBuilder`、`Cron(jobID, expr string) *ScheduleBuilder`；`(*ScheduleBuilder).Timezone(name string) *ScheduleBuilder`、`Handle(fn func(*Context) error)`。静态 job ID 原样传给 `core.job.*`，其路由 ID 才加 `job.`。子插件 ID 只留在 SDK 的路由作用域与日志中。
- `sdk/session.go` 的 `buildRegister(reg *Registration, configJSON string, configVersion uint32, configSchema []byte) *pb.Register` 生成注册；`sdk/dispatch.go` 的 `businessHandler(reg *Registration) wire.Handler` 包装静态任务；`sdk/activate.go` 的 `currentSession(ctx context.Context) (*wire.Session, error)` 实施 Activated/停止门禁。`sdk/services.go` 的 `Proto[In, Out proto.Message](target string) Service[In, Out]`、`(Service[In, Out]).Call(ctx context.Context, in In) (Out, error)` 可复用固定服务调用与错误转换。
- `contracts/proto/alonabot/v1/route.proto` 已有 `ROUTE_KIND_TASK = 8`、`TaskSpec { string name = 1; }`、`RouteDescriptor.task = 18`、`required_permissions = 17`。`message.proto` 的 ScheduledPayload 字段为 `job_id = 1`、`run_id = 2`、`scheduled_at_unix_ms = 3`、`task_name = 4`、`task_key = 5`、`task_payload_json = 6`。`contracts/validate/register.go` 已校验任务名、`task.` 保留前缀和 task 路由。
- `contracts/proto/alonabot/v1/task.proto` 与 `internal/wire/services.go` 已提供六个 task/job 服务及全部载荷；准确字段见下文接口段。List 使用可缺省的 `after_key`、`next_after_key`；TaskInstance.next_at、JobStatusResponse.last_at/next_at 均为 optional Unix 毫秒，缺失须保留。
- `sdk/message/event.go` 的 `ScheduledPayload` 目前为 `{ JobID, RunID string; ScheduledAt time.Time }`；动态参数须在 SDK 的原始 HandlerRequest 边界解码后传给 T，沿用现有 Context 的调用有效期。`sdk/sdk.go` 已有 `(*Context).Context() context.Context`、`Session() message.Session`、`User() message.User`、`Reply(text string) error`、`Send(m message.Content, opts ...SendOption) (*message.SendResult, error)`；`sdk.To(s message.Session) SendOption`。任务无默认 Target，提醒须显式使用参数中保存的会话。
- 当前定时示例位于 `examples/plugins/showcase/main.go` 的 `(*Demo).Schedules(r *sdk.Plugin)`，生成项为同目录 `alonabot_gen.go`，入口已使用 `sdk.UseGenerated(p, alonabotGenerated)`；`docs/development.md` 尚无独立的动态任务一节。根包 `pipeline_test.go` 已有 `newPipeline(t *testing.T, stdio ...bool) *pipelineHarness`、`(*pipelineHarness).launch(id, mode string, priority int) *runningComponent`；`scheduler_tasks_test.go` 的 `persistentTaskProcess(mode string) error` 当前以原始协议验证 Core，`TestPersistentTasks` 保留。
- 旧状态快照中的 PR #165（goal-55，`330164f`）、#167（goal-84，`3a9b4af`）、#168（goal-81，`37b0ada`）、#169（goal-127，`4dc4bcd`）及 goal-59（PR #170，`eca6310`）均已合并。按用户提供的在途信息，goal-40 涉及 state、supervisor、根包装配，PR #171（goal-69）涉及根包与 `cmd/alonabot`，PR #172（goal-57）涉及 `pipeline_test.go`、`t11_e2e_test.go`。本切片与 #172 的潜在重叠是测试子进程分流；与 #171 的潜在重叠是 `docs/development.md`。#169 对 schema.sql、manage 清理清单的改动已在基线中，本切片复用其六个服务。

```text
/goal 使用 $alonabot-execute 推进 goal-128（切片 T-07.9：动态任务与暂停恢复的 SDK）。goal 目录 goal-128，分支 goal-128，worktree .worktrees/goal-128，从最新 origin/main 创建（包含 PR #154、#169）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件作者用带类型的任务声明，在 New() 中登记处理函数，Activated 后创建、替换、取消与列出带参数的实例；静态任务可暂停、恢复与查询。复用已合并的 Core 调度、持久化与六个固定服务，提供提醒示例和开发说明。

依据：spec/PLAN.md §3.1；spec/plan/core.md T-07 步骤 10（T-07.9），关联步骤 3、7、9 与「停止和失败条件」；spec/DESIGN.md §H；spec/CONTRACTS-v0.1.md §1.2、§3 的 JSON Service、§4.3、§8 的主动发送、§14.5、§16；spec/MESSAGE-MODEL.md §1/§1.1 的 scheduled 无 Actor 与显式 Target、§3 的 HandlerOutcome；spec/WIRE-FORMAT.md §3.3、§3.5；spec/adr/DECISIONS.md D-70、D-86、D-107、D-114、D-116、D-124～D-126；spec/TEST-MATRIX.md V-114 的 SDK 分项，复验 V-20、V-46 的实际运行槽。

范围（只新增或修改这些）：
- sdk/ 新增 task.go、job.go；sdk/plugin.go、sdk/sdk.go、sdk/session.go、sdk/dispatch.go、sdk/subplugin.go：任务登记、Collect、REGISTER、派发与子插件作用域映射所需的相邻修改。新增登记状态保持私有。
- sdk/generate/service.go：将 Task[T] 纳入既有 JSON 类型检查，复用 Service 的类型规则；任务参数顶层要求 JSON 对象。沿用现有生成结果结构，生成器公开 API 保持现状。
- examples/plugins/showcase/main.go：保留 tick，增加提醒创建与取消命令及具名参数类型；由现有生成入口更新同目录 alonabot_gen.go。
- docs/development.md：增加动态任务与静态任务暂停恢复的使用说明。
- 根包新增 scheduler_sdk_test.go（unix）；pipeline_test.go 仅增加该项集成需要的子进程模式分流与注册入口，保留已有断言和其他模式。
- 上述包的 *_local_test.go：本地自测，留在忽略文件中。

允许的新跨模块边：无。SDK 复用 contracts、internal/wire 与现有 SDK 子包；示例通过 SDK；根包测试复用生产装配。

接口（照此实现；内部辅助函数自定）：
1. sdk 新增作者 API；TaskType、TaskBuilder、TaskOption 的实现字段私有：
   type TaskType[T any] struct { /* 私有字段 */ }
   func Task[T any](name string) TaskType[T]
   type TaskBuilder[T any] struct { /* 私有字段 */ }
   func OnTask[T any](r Registrar, t TaskType[T]) *TaskBuilder[T]
   func (b *TaskBuilder[T]) Handle(fn func(*Context, T) error)
   type TaskOption func(*taskOptions)
   func Key(key string) TaskOption
   func Timezone(name string) TaskOption
   func (t TaskType[T]) At(ctx context.Context, when time.Time, args T, opts ...TaskOption) (string, error)
   func (t TaskType[T]) Every(ctx context.Context, d time.Duration, args T, opts ...TaskOption) (string, error)
   func (t TaskType[T]) Cron(ctx context.Context, expr string, args T, opts ...TaskOption) (string, error)
   func (t TaskType[T]) Cancel(ctx context.Context, key string) error
   func (t TaskType[T]) List(ctx context.Context) ([]TaskInstance[T], error)
   type TaskInstance[T any] struct {
       Key string
       Args T
       At time.Time
       Interval time.Duration
       Cron, Timezone string
       NextAt *time.Time
   }
   At、Interval、Cron 对应协议的三种互斥触发；未选择的字段为零值，Timezone 仅用于 Cron；NextAt nil 表示协议字段缺失。
   type JobState struct { Paused bool; LastAt, NextAt *time.Time }
   func PauseJob(ctx context.Context, jobID string) error
   func ResumeJob(ctx context.Context, jobID string) error
   func JobStatus(ctx context.Context, jobID string) (JobState, error)
   LastAt/NextAt nil 保留协议 absent；函数名 JobStatus 保留，结果类型使用 JobState 以满足 Go 的包级命名规则。
2. 使用以下已合并协议；字段编号、服务名及 Core 语义保持现状：
   ROUTE_KIND_TASK=8；RouteDescriptor.task=18；TaskSpec.name=1；route_id="task."+name。
   TaskTrigger.oneof { int64 at_unix_ms=1; uint64 interval_ms=2; string cron=3; }; string timezone=4。
   TaskSetRequest { string task_name=1; string key=2; TaskTrigger trigger=3; string payload_json=4; }
   TaskSetResponse { int64 next_at_unix_ms=1; }
   TaskCancelRequest { string task_name=1; string key=2; }; TaskCancelResponse {}。
   TaskListRequest { string task_name=1; optional string after_key=2; }
   TaskInstance（pb）{ string key=1; TaskTrigger trigger=2; string payload_json=3; optional int64 next_at_unix_ms=4; }
   TaskListResponse { repeated TaskInstance instances=1; optional string next_after_key=2; }
   JobPauseRequest/JobResumeRequest/JobStatusRequest 各只有 string job_id=1；Pause/Resume Response 为空。
   JobStatusResponse { bool paused=1; optional int64 last_at_unix_ms=2; optional int64 next_at_unix_ms=3; }
   ScheduledPayload 的任务字段为 task_name=4、task_key=5、task_payload_json=6；动态 job_id 与请求目标均为 task.<name>。
   固定服务及常量：wire.ServiceTaskSet="core.task.set"、ServiceTaskCancel="core.task.cancel"、ServiceTaskList="core.task.list"、ServiceJobPause="core.job.pause"、ServiceJobResume="core.job.resume"、ServiceJobStatus="core.job.status"。请求与响应使用现有 Protobuf Schema ID/version=1。
3. sdk.Registrar、Plugin/SubPlugin、Collect、Run 及既有静态任务 API 保持现有签名。示例通过 p.Command(...).Handle(...)、sdk.OnTask(p, declaration).Handle(...) 登记；自动回复用 message.Content，发送用 message.Text 与 sdk.To。

行为：
- Task 是带类型的声明值。T 遵循现有 Service JSON 类型规则，编码结果为 JSON 对象；声明零值、空或非法任务名、同名重复、缺失/重复 Handle、保留前缀冲突在 Collect 失败，错误带任务名。任务名沿用 Core 的 ^[a-z][a-z0-9_.-]{0,63}$。所有 route_id 在整个插件内唯一；子插件登记仍为 task.<name>，Handler 的日志保留 sub_plugin。
- At/Every/Cron 将参数完整编码后调用 core.task.set。Key 指定原样 key；省略 Key 时以标准库的密码学随机源生成符合既有 1～128 UTF-8 字节约束的 key，随机源失败向调用者返回错误。相同 key 交由 Core 整体替换；成功返回实际使用的 key。Timezone 在 At/Every 上返回 INVALID_ARGUMENT，调用前结束。编码失败、非对象参数明确返回 INVALID_ARGUMENT。
- 触发合法性、配额、并发、持久化与运行槽以 Core 为准：interval 至少 1s、五字段 cron、IANA 时区/缺省 UTC、at 晚于 Core 当前时间；替换保留在途旧一轮；一次性实例派发时删除，执行失败与超时沿既有行为结束。
- List 按 next_after_key 读取全部分页，顺序与 Core 的 key 顺序一致。保留每个实例的 key、完整参数、触发规则及 next_at 的存在性；任一页调用或解码失败则返回错误，完整成功后才交付列表。取消与静态 job 操作保留 Core 的 NOT_FOUND 等错误码。
- dynamic REQUEST 复用静态 scheduled 的 Context、panic/错误映射、取消与唯一 Response 路径。核对请求目标、任务名与载荷后将 task_payload_json 解码为 T；失败返回 INVALID_ARGUMENT 并记英文日志，Handler 未执行。Handler 没有 Actor 和默认 Target；需要 key、会话或用户时，作者写入 args。任务函数在取消后实际返回，才走既有 Response 完成与运行槽释放。
- 所有运行期客户端函数通过 currentSession 检查 Activated；OnStart/OnReady、停止后以及已失效的 Handler context 调用返回 FAILED_PRECONDITION。使用 c.Context() 保留现有父调用关联；独立后台 context 只有组件身份。协议错误码通过 sdk/errcode 原样传播。
- JobStatus 的 last_at 是本次 Core 运行最后一次实际派发；Core 重启后 absent。暂停时 next_at absent；暂停保持在途函数继续执行，恢复沿原 interval 起点安排未来触发。
- showcase 增加一条创建提醒命令和一条取消命令：在 New 中登记一次声明/Handler，命令从 c.Session() 保存会话字符串，用 sdk.Key 使作者掌握取消键，At 创建后向用户返回 key；回调以 args 中的会话显式发送。保留静态 tick，文档展示 PauseJob、ResumeJob、JobStatus 及 NextAt/LastAt 的缺失语义。

测试（按编码规范「测试提交规则」）：
- 提交：根包 scheduler_sdk_test.go 的一个 TestSchedulerSDK（//go:build unix）。复用 pipeline_test.go 的真实 Core 装配、SQLite、Session 与测试二进制子进程，插件使用本切片公开 SDK：At、Every、Cron 各验证一次真实触发及带类型参数；同 key 替换后 List 返回新规则/参数；Cancel 后停止后续触发；PauseJob/ResumeJob 配合 JobStatus 核对静态任务暂停与恢复。一次集成守住作者 API→协议→Core→typed Handler 的主路径。测试工程/子进程写在本测试或 t.TempDir()，沿用现有分流；sdk/、supervisor/ 和 testdata 不新增提交测试。
- 本地自测（*_local_test.go，PR 中记录命令、平台与结果）：非法 Task 声明/类型、根与子插件重复、保留前缀、nil/重复 Handle；Timezone 非法组合且未发请求；零值声明；随机 key；编码失败及非对象；全部分页和中间页错误；造入错误参数 JSON 时 INVALID_ARGUMENT、英文日志与 Handler 未执行；Activated 前/停止后/失效 Context；无 Actor/Target、显式发送准入；取消在途回调后仍等实际返回；Core 错误码；JobStatus optional 时刻和静态任务重启后的缺失语义。真实插件多进程验证优先复用根包装配。

完成条件：gofmt -l 对改动 Go 文件无输出；示例通过现有生成入口重建且重复生成内容一致；go vet ./...；go test -race ./...；go build ./...。逐项记录 V-114 的 SDK 证据、实际平台与未运行项；T-07 的同实例无重叠、取消/停止和无提前业务准入成立后结束。

不做：Core 协议、状态表、调度器与卸载清理（goal-127）；Service 的权限 API；plugintest 的动态任务立即触发（goal-82）；官方 Agent 的定时任务（goal-63）；新增任务管理页面或持久化执行队列。
```

### goal-151 · T-08.15 配置字段的回复示例

状态：已完成（PR #211），合并提交 `3b2b64c`。以下保留派发时语句。派发时：前置 T-08.11、T-08.14 与 T-09.12 已合并。2026-11-17 按 D-160 新增；预览由 WebUI 渲染（goal-44），本切片只做检查、随包校验与图片的提供。

```text
/goal 使用 $alonabot-execute 推进 goal-151（切片 T-08.15：配置字段的回复示例）。goal 目录 goal-151，分支 goal-151，worktree .worktrees/goal-151，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：`config_schema.json` 的字段可以带 `example`；`alonabot check`、`build` 与 `pack` 校验它与 i18n 覆盖，`examples/` 下被引用的图片随包校验；Core 经 `GET /api/v1/packages/{id}/examples/{path}` 提供 Active 代次中被引用的图片。

依据：PLAN T-08 步骤 12；CONTRACTS §5.1（`example` 与字段路径）、§5.3（i18n 覆盖）、§7.5；ADMIN-API §3.3 之后的 examples 端点说明；DECISIONS D-160；TEST-MATRIX V-131 的检查与 API 分项。

范围（只新增或修改这些）：
- contracts/configschema/：解析与检查 example（schema.go 或新文件 example.go）、i18n 覆盖（i18n.go）
- internal/pluginkit/：check、build 与 pack 校验并打包 examples/ 中被引用的文件
- internal/core/packages/preflight.go：安装与更新时校验被引用的图片
- internal/core/manage/：examples 端点
- examples/plugins/showcase：给 greeting 加一个文字与图片的示例
- docs/plugin-guide.md：example 的写法
- 上述的测试文件

接口（contracts/configschema）：
  type Example struct { Input string; Reply []ExampleSegment; Condition map[string]json.RawMessage }
  type ExampleSegment struct { Text string; Image string; Alt string } // Text 与 Image 恰有一个
  Field 增加 Examples []Example
  func (s *Schema) ExampleImages() []string // 全部被引用的 examples/ 相对路径，去重排序

行为：按 CONTRACTS §5.1 检查结构、片段、占位符（`{字段路径}`、`{字段.$key}`、`{{`、`}}`）与示例条件引用的字段及其类型，错误带 JSON 路径；图片只接受 PNG、JPEG、WebP、GIF，按内容核对扩展名。端点要求管理认证，只提供 Active 代次 schema 中被引用的路径，响应头同组件图标；未引用、缺失、目录或链接返回 404。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix）：showcase 装上后 examples 端点返回图片与图标相同的安全头，未引用的路径 404。
- 本地自测（*_local_test.go，结果写入 PR）：V-131 的检查分项逐项（每类错误的 JSON 路径）、占位符与花括号、i18n 覆盖不同形、图片扩展名与内容不符。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：WebUI 的预览渲染（goal-44）；预览的 html 类型（未确定）。
```

### goal-66 · T-09.10 日志管理

状态：已完成（[PR #177](https://github.com/AlonaBot/AlonaBot/pull/177)，合并提交 `99efcd1`），证据见 TEST-MATRIX §3.2。合入：服务 `sdk.log.set_level`（`LogSetLevelRequest/Response`）；`logging.Options.BufferRecords`、`(*Output).Recent`、`(*Output).Subscribe`、`logging.Query`、`logging.Record`；管理 API 的 `/api/v1/logs/recent`、`/stream`、`/history` 与日志级别修改，`alonabot logs` 与日志级别命令；组件日志的可信来源类别。以下为派发时的语句。

按 main 核对：

- `internal/core/logging/logging.go`：`type Options struct { Format Format; Level slog.Level; Stderr io.Writer }`；`Open(opt Options) (*Output, error)`；`(*Output).Logger() *slog.Logger`、`Levels() *Levels`、`EnableFile(opt FileOptions) error`、`Plugin(pluginID, generationID string) *Source`、`Component(pluginID, generationID string) slog.Handler`、`Close() error`。`Levels` 的 `SetGlobal(lv slog.Level)`、`SetSource(id string, lv slog.Level)`、`InheritSource(id string)`、`Source(id string) slog.Level` 已存在；`logging.ProtoLevel(lv slog.Level) pb.LogLevel` 与 `LevelOff` 可复用。目前没有 Recent、Subscribe 或 Query。
- `internal/core/logging/source.go`：`(*Source).Line(_, stream string, line []byte)` 在解析并补来源后按来源级别过滤；`SetSession(sessionID []byte)` 补 Session。`console.go` 的 `consoleHandler.Handle(ctx context.Context, r slog.Record) error` 写 text；`traceDigits = 8`，Plugin/Adapter 的 session/generation 字段在文本中省略。`file.go` 的 `FileOptions{Dir string; MaxSize, MaxTotal int64; MaxAgeDays int}` 及文件命名已实现，历史查询复用本地日期和编号续写规则。
- `internal/core/settings/settings.go`：`Load(dataDir string) (*File, error)`、`(*File).Settings() Settings`（深副本）；Settings 有 `LogLevel slog.Level`、`LogSources map[string]string`、`LogFile`。`file.go` 的 `(*File).Update(edit func(doc *Document) error) error` 以 Rename 为提交点，之后目录同步失败仍发布快照并返回错误；`document.go` 的 `Set(path []string, value any) error`、`Delete(path []string) error`。按来源级别允许 debug/info/warn/error/off/inherit，权威存储为 core_config.json。
- `sdk/log.go`：`setLogLevel(level *slog.LevelVar, lv pb.LogLevel)`；`sdk/inproc.go`：私有 `componentSession` 持有 `level slog.LevelVar`，`(*componentSession).handler(next wire.Handler) wire.Handler` 同时用于插件与 Adapter 的 RPC。`sdk/activate.go` 的 `activate(conn *wire.Conn, l launch, ack *pb.HelloAck, root *Plugin, reg *Registration) (session *wire.Session, active *activation, err error)` 和 `sdk/adapter_run.go` 的 `runAdapter(a adapter.Adapter, l launch) (err error)` 分别建立 Session。两条路径尚无 sdk.log.set_level 分派；`control.proto` 与 `internal/wire/services.go` 尚无 LogSetLevel 消息和服务常量。
- `internal/core/router/router.go`：`(*Router).Plugin(pluginID string) (Plugin, bool)`、`Plugins() []Plugin`；Plugin 有 `PluginID() string`、`Activated() bool`、`Call(ctx context.Context, req wire.Request) (wire.Response, error)`。可对当前 Activated Session 通知，使用组件 ID，保持跨 Session 身份检查。
- `internal/core/manage/service.go`：`New(opt Options) *Service`；Options 已持有 `Settings *settings.File`、`Store *state.Store`、`Development map[string]*manifest.Manifest` 等装配资源。`http.go` 的 `(*Service).Handler(auth *adminauth.Auth) http.Handler` 统一调用 `auth.Require`；私有 `respond(w http.ResponseWriter, result any, err error)` 处理确认与管理错误。`static.go` 的 `AdminAPIMajor uint32 = 1` 与 `(*Service).WebUI() http.Handler` 已合并。
- 根包 `alonabot.go`：`Start(ctx context.Context, opts ...Option) (*Core, error)` 只加载一次 settings.File，构造 `logging.Output`、manage、公开 Listener；`LogLevel(level slog.Level) Option` 为本次启动覆盖。`assembly.go` 的 `startDevelopment` 明确实现开发来源缺省 debug、显式 inherit 使用全局；`startInstalled` 的 Bootstrap 使用 `c.output.Levels().Source(install.PluginID)`。现有持久化 LogSources 在正式组件启动前的完整加载须随本切片接入。
- `cmd/alonabot/commands.go` 的 `execute(args []string) int` 注册命令；`manage.go` 的私有 `managementOptions`、`(*managementOptions).client() (*managementClient, error)`、`(*managementClient).request(ctx context.Context, method, path string, body any) (json.RawMessage, error)` 使用 Bearer 并拒绝重定向。request 当前对整段 path 赋 URL.Path，日志查询串须独立构造 RawQuery。`cmd/alonabot/manage_e2e_test.go` 已有真实安装包、Core 进程与管理 CLI 的装配。
- 旧在途列表中的 #165（goal-55，`330164f`）、#167（goal-84，`3a9b4af`）、#168（goal-81，`37b0ada`）、#169（goal-127，`4dc4bcd`）及 goal-59（#170，`eca6310`）均已合并。按用户提供的在途信息，goal-40 改 state、supervisor、根包装配，#171（goal-69）改 `alonabot.go`、`assembly.go`、`cmd/alonabot/commands.go` 等，#172（goal-57）改根包的测试装配。此切片主要与 goal-40/#171 重叠于根包，与 #171 重叠于 CLI 注册；contracts/wirespec_test.go、control.proto、wire 服务常量的后续追加均须保留当前 #165/#169 的编号与服务。日志级别走 settings，本切片范围无需 schema.sql 或 manage 清理清单。

```text
/goal 使用 $alonabot-execute 推进 goal-66（切片 T-09.10：日志管理）。goal 目录 goal-66，分支 goal-66，worktree .worktrees/goal-66，从最新 origin/main 创建（包含 PR #163、#167）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：经唯一管理服务读取与修改全局及按来源的输出级别，写回 core_config.json 并向运行中 SDK 下发 sdk.log.set_level；Core 为每个来源保存最近 LOG_BUFFER_RECORDS 条已输出记录，提供带筛选的读取、SSE 与按日期的文件查询，alonabot logs/log-level 和后继 WebUI 日志页共用这些 API。

依据：spec/PLAN.md §3.1；spec/plan/core.md T-09 步骤 7 的日志部分及 V-62/V-63/V-68「对应 V 验收与停止条件」（T-09.10）；spec/CONTRACTS-v0.1.md §7.3～§7.5、§14.1～§14.6、§15.1/§15.4、§16 的 LOG_BUFFER_RECORDS；spec/MESSAGE-MODEL.md §1 的来源与可信身份；spec/WIRE-FORMAT.md §3.2 的 Bootstrap.log_level、§3.5 的固定服务与请求/响应元数据；spec/adr/DECISIONS.md D-62、D-82、D-86；spec/TEST-MATRIX.md V-62/V-63/V-68 的 T-09.10 分项；spec/ADMIN-API.md §1、§5 的认证与错误约定。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：只追加 LogSetLevelRequest/Response；运行 go generate ./... 更新对应 contracts/pb/alonabotv1/ 产物。
- contracts/wirespec_test.go：追加上述两消息的协议编号检查。
- internal/wire/services.go：追加 ServiceLogSetLevel 常量，保留现有 tool/task/job 常量。
- internal/core/logging/：新增 buffer.go、query.go；logging.go、source.go 的最小输出接入点，console.go/file.go 仅复用或提取既有格式解析所需的私有格式逻辑。保留现有输出、轮转、脱敏及错误处理。
- internal/core/manage/service.go、http.go；新增 log_api.go：同一管理服务内的读取、SSE 与级别修改。
- sdk/log.go、sdk/inproc.go：共用的 sdk.log.set_level 分派、会话级与独立进程日志阈值更新。sdk/activate.go、sdk/adapter_run.go 仅在现有共用 handler 的接线确有需要时同步。
- 根包 alonabot.go、assembly.go：传入有效缓冲上限、接入管理资源与当前 Session 通知、正式/开发组件启动级别。只修改日志相关装配。
- cmd/alonabot/ 新增 logs.go；commands.go 只登记两个命令；manage.go 只复用客户端认证/错误解析并使日志请求正确携带 URL 查询串及流式读取。
- spec/ADMIN-API.md：追加日志 API、JSON/SSE 字段及历史数据精度；本切片获授权修改该文件。
- cmd/alonabot/logs_e2e_test.go（unix）；上述包 *_local_test.go；既有测试只同步增加的必填装配字段，原行为断言保留。

允许的新跨模块边：无。manage/logging/settings/router 属于 internal/core；根包、cmd、SDK 与协议的边均已登记。公开 SDK API 保持现有符号。

接口（照此实现；内部辅助函数自定）：
1. control.proto 追加：
   message LogSetLevelRequest { LogLevel level = 1; }
   message LogSetLevelResponse {}
   internal/wire 追加 const ServiceLogSetLevel = "sdk.log.set_level"。
   使用现有 LogLevel 的 DEBUG/INFO/WARN/ERROR/OFF；UNSPECIFIED 沿用 INFO。请求、响应为现有 Protobuf Schema ID、version=1；成功返回完整空消息元数据。
2. internal/core/logging：
   Options 追加 BufferRecords int，由 Core 的 tune.Values.LogBufferRecords 传入；同步 Open 的全部调用者。
   type Record struct {
       Time time.Time `json:"time"`
       Level slog.Level `json:"level"`
       Source string `json:"source"`
       Message string `json:"message"`
       Attrs map[string]any `json:"attrs"`
       TraceID string `json:"trace_id"`
   }
   func (o *Output) Recent(source string) []Record
   func (o *Output) Subscribe(ctx context.Context) <-chan Record
   func Query(dir string, date time.Time, source string, level slog.Level, traceID string) ([]Record, error)
   source 为空读取全部来源；实时 Source 为 core 或组件 ID（Adapter 使用包 ID）；Attrs 保留来源类别、平台/实例、Session/Generation、源码位置等实际存在的属性。Record.Time 的 JSON 为 RFC3339Nano 字符串，Level 使用 slog.Level 的标准 JSON 字符串。
3. manage.Options 追加：
   Logs *logging.Output
   NotifyLogLevel func(ctx context.Context, id string, level pb.LogLevel) error
   Logs 为同一 Core 的输出；日志目录取现有 Packages 父目录下 logs。NotifyLogLevel 由根包绑定当前 Router Session：未运行来源只保存设置，运行来源调用 sdk.log.set_level。保持 New、Handler、WebUI 的现有签名。
4. HTTP，全部经已有 auth.Require：
   GET /api/v1/logs/recent?source=&level=&trace= → 200，Record[]。
   GET /api/v1/logs/stream?source=&level=&trace= → 200，text/event-stream；event: log，data: 为单个 Record JSON。
   GET /api/v1/logs/history?date=YYYY-MM-DD&source=&level=&trace= → 200，Record[]。
   GET /api/v1/logs/levels → 200，{"global":"info","sources":{"组件ID":"debug"}}。
   PUT /api/v1/logs/levels 接收同形的部分更新；省略 global 或某个来源时保留当前值；sources 中某项为 inherit 时写入该来源的显式 inherit。成功返回与 GET 同形的当前有效全局值及文件中的来源设置。
   global 取 debug/info/warn/error；sources 值再允许 off/inherit。GET 的 global 为本次 Core 的有效全局值（含启动覆盖），sources 为配置的来源覆盖；空 sources 编码为 {}。level 查询表示最低级别，缺省 debug；source 为空为全部；实时 trace 按完整 trace_id 精确匹配。日期以 Core 本地时区解释，非法日期/级别返回 400/INVALID_ARGUMENT。
5. CLI：
   alonabot logs [--source ID] [--level L] [--trace T] [--follow] [--date YYYY-MM-DD]
   alonabot log-level [ID] [LEVEL]
   使用现有 --data-dir、--admin-url、--token-file 及 Bearer 客户端。log-level 无参数列出设置，单独 ID 显示该项，ID=global 修改全局；按来源允许 inherit/off。logs 缺省读 recent，--date 读 history，--follow 消费 stream（含初始快照）；date 与 follow 组合返回明确用法错误。复用现有控制台格式输出完整记录，流由命令 context 取消时结束。

行为：
- 缓冲在级别过滤后的共用输出路径接入，含 Core、自有插件、Agent 与 Adapter 包；只记录已输出的内容。每个来源上限使用 LOG_BUFFER_RECORDS，缺省1000，满时覆盖最旧；Core 重启后清空。Core 关于某组件的故障记录仍属于 core，其缓冲与级别都按全局处理。保留现有脱敏、非结构化输出与日志文件失败处理。
- 初始化时从同一个 settings.File 应用 log.sources，再构造组件 Bootstrap；已安装组件缺省 inherit，开发组件缺省 debug，显式 inherit 跟随本次有效全局值。安装进程、开发独立进程和同进程组件使用同一生效规则。
- 修改级别先经 File.Update 写回 log.level/log.sources；保存成功后更新 Core Levels 并给当前 Activated Session 下发解析后的级别，SDK 同时更新该 componentSession.level 和其独立进程 logLevel。插件、Agent、RunAdapter 及同进程模式均覆盖；同进程更新只作用于对应组件。来源为 off 时 SDK 停止输出，Core 关于它的记录照常由全局控制。
- 按现行 CONTRACTS §14.3，全局运行时修改也写回 log.level；重启取已保存设置，下一次显式 --log-level 仍只覆盖该次启动。改全局时通知显式 inherit 及已安装的缺省 inherit 来源，保留开发来源缺省 debug 和显式来源覆盖。
- File.Update 在 Rename 后报告目录同步失败时，快照已经发布：读取实际已发布快照同步内存与通知，并返回保存/通知的实际错误；Rename 前失败保持旧值。保存成功但通知失败保留已提交设置，报告错误及受影响来源，详细原因进入 Core 日志。下一次组件启动从已保存设置取级别；本切片只执行当前一次通知。
- SDK 在内部固定服务入口处理 set_level，使用现有 Session 准入与完整载荷校验，Handler 路由保持业务语义。SSE 写出前完成认证；Cookie 写级别需 Origin 与 X-Alonabot-CSRF；Bearer 沿用现状。未认证读取/订阅按既有规则返回401。
- stream 在输出侧原子建立订阅及初始缓冲快照，先交付快照再交付后续记录；--follow 可直接使用该快照。WebUI 按 §14.4 可先读 recent 再订阅，重复快照由消费者处理。订阅随请求 context 关闭；网络写入在输出锁外执行。每个订阅有 LOG_BUFFER_RECORDS 条的发送队列，队列满时不阻塞日志写入，Core 结束该订阅：SSE 先发 `event: overflow`（data 为已交付的最后一条记录的时间），再关闭流；客户端重新连接后用 recent 补齐。--follow 遇到 overflow 时输出一行说明并以非零状态退出。
- history 只读取选定日期的普通日志文件，按续写序号从主文件到 .1/.2... 读取；解析 §14.2 的来源标签、级别、位置、带引号字段、转义和多行值，保存文件中实际存在的内容。该日没有文件返回[]，读取或解析故障返回真实错误。文本中只有8位 trace 前缀及显示来源；历史 API 返回文件中的原始记录：source 按文件中实际的来源标签筛选（插件为去掉市场前缀后的显示名，Adapter 为 adapter_id/instance_id），trace 按 8 位前缀匹配，Record.Message 保留整行在来源标签之后的原文，Attrs 只含能可靠解析的字段；这些精度登记进 spec/ADMIN-API.md。完整 trace、包身份及已省略的 Session 字段仅由实时结构化记录提供。
- 所有管理接口沿用主版本1、已有成功/错误 JSON 及认证错误格式；CLI 的查询参数使用 URL.RawQuery，保留拒绝重定向和服务端错误，SSE 按流读取。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/logs_e2e_test.go 的一个 TestLogManagement（//go:build unix），复用真实 Core/CLI/临时安装包和文件装配；临时构建的日志组件放 t.TempDir()。验证 log-level 改来源为 debug 后 SDK 发出 DEBUG 且 logs 读到、改为 off 后进程记录消失而 Core 关于它的记录仍能读到、改为 inherit 后跟随全局；全局与来源设置写回后重启仍按文件生效，另一次启动的 --log-level 只覆盖本次。--follow 收到初始快照和随后新记录，未认证 recent/stream 被拒；放置前一天同格式文件后 logs --date 读到原消息。理由：同一个集成入口守住管理写回、跨进程动态通知、CLI、SSE 与文件查询主路径。
- contracts/wirespec_test.go 追加 LogSetLevelRequest.level=1 及空响应检查；多列理由：协议编号检查是编码规范允许的另一类提交，用于固定新消息布局。
- 本地自测（*_local_test.go，PR 记录结果）：小 LogBufferRecords 的逐来源覆盖；只缓存已输出记录；Adapter 包与实例归组、Agent 与 Core 记录；全部来源合并及筛选；SSE 快照与并发追加、取消与断开；text/json 输出均保留完整实时 trace；历史的续写顺序、多行/转义、缺失日期、损坏行和实际 I/O 错误；Cookie/CSRF；设置提交前失败与提交后目录同步失败、一个 Session 通知失败；同进程两个组件的级别隔离；SDK 与 RunAdapter 的 set_level 完整元数据、OFF/UNSPECIFIED 与停止门禁。sdk/、supervisor/、testdata 按 D-86 留在本地验证。

完成条件：gofmt -l 对改动 Go 文件无输出；go generate ./... 且生成结果与源定义一致；go vet ./...；go test -race ./...；go build ./...。记录 V-62/V-63/V-68 的本切片证据与实际未运行项。鉴权、日志敏感值或实时级别控制出现缺陷时修复后交付；历史精度与慢订阅行为按审查者裁决闭合后再声明验收完成。

不做：WebUI 日志页（goal-94/T-09.15）；远程日志收集、额外日志存储格式；日志文件设置的运行时修改；状态库日志级别表、额外调优参数和业务自动重试。
```

### goal-92 · T-09.13 管理页所需的会话与上传接口

状态：已完成（[PR #181](https://github.com/AlonaBot/AlonaBot/pull/181)，合并提交 `d507e5a`），证据见 TEST-MATRIX §3.2。合入：`GET /api/auth/session` 取回已登录会话的 `csrf_token` 与 `expires_at_unix_ms`（不续期）；`POST /api/v1/packages/install` 与 `/update` 同时接受 JSON 与 multipart（archive、release、可选 signature 与 public_key），每个请求独立的临时目录，结束即清理；`GET /api/v1/core` 返回 Core 版本（module replacement 取替换模块的版本，本地路径替换为 `(devel)`）、实际监听地址与 Admin API 主版本；经 HTTP 的更新在 T-09.6 合入后真实准备并提交，`TestAdminSessionAndUploads` 覆盖失败分支保持旧状态与有效更新激活新版本。spec/ADMIN-API.md 已登记。以下为派发时的语句。

按 main 核对：

- `internal/core/adminauth/auth.go`：`New(st *state.Store, now func() time.Time, adminBaseURL *url.URL, dataDir string) *Auth`；`(*Auth).Routes(mux *http.ServeMux)` 目前登记 POST `/api/auth/login`、POST `/api/auth/logout`；`Require(next http.Handler) http.Handler` 支持 Bearer 或有效 Cookie，Cookie 变更请求校验 Origin 和 `X-Alonabot-CSRF`；`Lookup(r *http.Request) (sessionID string, expiresAt time.Time, ok bool)` 只有会话ID/期限，并会把存储错误记录为日志。新增 session 端点可在同包直接复用私有 credential/cookieSession，保留存储错误语义。
- `internal/core/state/admin.go`：`type AdminSession struct { SessionID, CSRFToken string; AuthGeneration, ExpiresAt int64 }`（ExpiresAt 为 Unix 毫秒）；`(*Store).AdminSession(ctx context.Context, sessionID string) (AdminSession, error)`、`DeleteAdminSession(ctx context.Context, sessionID string) error` 已存在。有效 Cookie 会话须校验当前认证代次和固定8h到期时间。
- `internal/core/manage/install.go`：`type InstallRequest struct { Archive, ReleaseJSON, Signature string; PublicKey string; Source string; Replace, Confirm bool }`；Archive 是 Core 本机路径，ReleaseJSON/Signature 是文件原文，PublicKey 是 release.pub 的标准 base64 内容；`(*Service).Install(ctx context.Context, req InstallRequest) (Package, error)`。无签名包按 D-81 仍可确认安装；有签名时解码公钥并经 `release.Verify` 验证原始清单字节。
- `internal/core/manage/candidates.go`：`(*Service).Update(ctx context.Context, id string, req InstallRequest) error`（源码使用命名 err 返回）；`type Candidates interface { Reload(ctx context.Context, id string) error; Update(ctx context.Context, id, staged string) error; Rollback(ctx context.Context, id, generation string) error }`。默认 unsupportedCandidates 返回 `UNSUPPORTED`。T-09.5 的 main 语句已明确：goal-40 只交付 Prepare/Abort，计划把内部 Candidates.Update 改为完整 Candidate；生产管理动作接线由 goal-41 在 Commit 可用后完成。上传层只调用公开 Service.Update，其签名保持当前约定。
- `internal/core/manage/http.go`：`(*Service).Handler(auth *adminauth.Auth) http.Handler`，已有 POST `/api/v1/packages/install` 与 POST `/api/v1/packages/update`；JSON InstallRequest 使用 Go 导出名，update 外层为 `id`/`confirm`/`install`。私有 respond 统一成功结果、`ErrNeedsConfirm{Affected []string}` 的409和 `code/safe_message` 错误。认证错误沿用英文纯文本。
- `internal/core/manage/static.go`：`const AdminAPIMajor uint32 = 1`；`(*Service).WebUI() http.Handler`。`service.go` 的 Options 包含 `Packages string`、`Limits packages.Limits`、`Settings *settings.File`；`install.go` 的 prepare 使用 Packages 父目录下 staging 及相同预检。`internal/core/packages/preflight.go`：`Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error)`、`(*Candidate).MoveTo(packagesRoot string) (string, error)`、`Discard() error`，保留完整验签、身份、解压和来源检查。
- 根包 `alonabot.go` 的 `Start(ctx context.Context, opts ...Option) (*Core, error)` 创建真实管理服务和 Listener；`Listen(addr string) Option` 为启动覆盖，实际监听地址来自 `listener.Addr().String()` 并写入 `c.http.Addr`。目前未发现框架自身的版本常量/版本读取入口；`manage.AdminAPIMajor` 表示 Admin API 版本，Core 版本来源须另行确定。
- 根包 `webui_test.go` 已通过真实 Start、管理服务、临时组件包与 Listener 验证 WebUI；`cmd/alonabot/manage_e2e_test.go` 已有带签名归档和真实 CLI/API 的辅助设施。SDK 构造测试组件时使用 `sdk.NewPlugin()`、`sdk.UseGenerated(p, generated)`、`func New() *sdk.Plugin`；命令回复类型用 message.Content。若复用 Adapter HTTP 夹具，实际签名为 `HTTP(route func(method, path string) *sdk.RouteBuilder)`，通过 `route(...).HandleFunc(...)` 注册；私有路由沿用管理认证。
- 旧「在途」列表中的 #165（goal-55，`330164f`）、#167（goal-84，`3a9b4af`）、#168（goal-81，`37b0ada`）、#169（goal-127，`4dc4bcd`）与 goal-59（#170，`eca6310`）均已合并；#169 的 schema.sql 与 manage 清理对象已在基线。按用户提供的在途信息，goal-40 与本切片重叠 `manage/service.go`/`candidates.go` 和根包装配，#171（goal-69）重叠根包，#172（goal-57）重叠根包测试装配。本切片保持 Service.Update 的调用，保留 goal-40/41 的内部交接；与后继管理 API 切片共改 `manage/http.go`、`spec/ADMIN-API.md` 时按端点归属合并。

```text
/goal 使用 $alonabot-execute 推进 goal-92（切片 T-09.13：管理页所需的会话与上传接口）。goal 目录 goal-92，分支 goal-92，worktree .worktrees/goal-92，从最新 origin/main 创建（包含 PR #72、#163、#167）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：WebUI 刷新后可取回当前已登录会话的 csrf_token 与到期时刻；浏览器可上传安装/更新包，复用同一管理服务的预检、确认与安装/更新流程；总览可读取真实 Core 版本和实际监听地址。全部新增接口与现有 JSON 入口一起登记在 spec/ADMIN-API.md。

依据：spec/PLAN.md §3.1；spec/plan/core.md T-09 步骤 1、3、7 的管理页后端部分（T-09.13）及「对应 V 验收与停止条件」；spec/CONTRACTS-v0.1.md §7.3、§7.4、§7.5、§9.1～§9.3 的安装预检/来源/确认、§15.1 的监听与启动覆盖；spec/MESSAGE-MODEL.md §1 的可信 Actor/权限与 HTTP 管理身份的现行约定；spec/WIRE-FORMAT.md §3.3 的统一 HTTP Listener（仅在复用测试 Adapter 时涉及，管理上传沿用 HTTP JSON/multipart）；spec/adr/DECISIONS.md D-62、D-81、D-82、D-84、D-86；spec/TEST-MATRIX.md V-40、V-47 的 API 分项；spec/ADMIN-API.md §1～§3、§5。

范围（只新增或修改这些）：
- internal/core/adminauth/auth.go：Routes 新增 GET /api/auth/session；可新增 session.go 放私有 handler。复用现有会话模型/校验，保留 New/Require/Lookup 等签名。
- internal/core/manage/http.go：install/update 根据 Content-Type 分派到既有 JSON 或新增 multipart 解码；service.go 的 Options 只追加总览所需的运行信息注入；新增 upload.go、info_api.go。
- 根包 alonabot.go：向管理服务提供 Core 构建版本与实际监听地址；可新增 version.go 放私有版本读取。根包外部 Option、Start 与 Core 的现有公开签名保持现状。
- spec/ADMIN-API.md：追加会话、multipart 与总览接口，明确大小写、认证、可选签名、公钥、确认、清理及真实更新依赖。本切片获授权修改该文件。
- 根包新增 admin_upload_test.go（unix）：一个真实 Core 的会话/上传集成；webui_test.go 等已有测试仅复用辅助函数，若需要改动只同步装配参数，断言保留。
- 上述包 *_local_test.go：本地自测；临时组件工程在 t.TempDir()，沿用已有测试二进制/包生成装配。

允许的新跨模块边：无。管理端复用 adminauth、packages、settings 与现有服务；根包装配提供运行信息；SDK 测试组件通过既有公共 API。标准库 mime/multipart、os、runtime/debug 足够。

接口（照此实现；HTTP 字段为本稿候选，派发前由审查者采纳；内部辅助函数自定）：
1. GET /api/auth/session：仅有效管理 Cookie 会话可用，成功200、application/json：
   {"csrf_token":"当前会话原有值","expires_at_unix_ms":1790000000000}
   expiry 是已保存期限；沿用当前认证代次/到期校验及错误格式。只带 Bearer 且没有有效管理 Cookie 返回401。Cookie 无效/过期/已登出/旧认证代次返回401；存储错误为500并记录真实错误，凭证未初始化沿用503。成功响应 Cache-Control: no-store。
   通过 Auth.Routes 登记，复用该包的私有 credential/cookieSession，完整取得 CSRFToken。此功能无需新增公开认证函数、状态字段或协议消息。
2. 原有 POST /api/v1/packages/install、POST /api/v1/packages/update 增加 multipart/form-data 支持；application/json 保持原请求、字段大小写、成功与错误语义。
   multipart 文件 part：
   - archive：归档字节；文件名通常为 <id>-<version>-<platform>.tar.gz。
   - release：release.json 原始字节。
   - signature：release.json.sig 原始字节；已签名上传提供该 part。
   multipart 文本 part：
   - public_key：release.pub 的标准 base64 内容，映射 InstallRequest.PublicKey；签名路径须具备可验证的公钥。
   - confirm：可选 true/false，省略 false。
   - replace：install 可选 true/false，省略 false，映射 InstallRequest.Replace。
   - id：update 必填的已安装组件 ID；其 confirm 映射 InstallRequest.Confirm。
   archive/release 必填且各一份；signature 对无签名包可缺省，有签名的完整路径仍为三文件。未知/重复 part、必需 part 缺失、错误布尔值、截断 multipart 返回400/INVALID_ARGUMENT。已签名但公钥缺失/错误等验签失败交由既有服务返回403/PERMISSION_DENIED。
   multipart 固定为本地上传候选；ReleaseJSON/Signature 保留精确字节转 string，Archive 指向服务器为本请求创建的临时文件；Source 留空，来源从清单与既有绑定校验。用户文件名只作展示，服务器自行选择临时路径。
   install 成功200、Package（沿用导出字段）；update 成功200、{}。确认409：{"affected":[...]}；客户端展示后重新上传同一组文件并置 confirm=true；已有 JSON 客户端原样使用 Confirm 或外层 confirm。
3. GET /api/v1/core：经现有 auth.Require，成功200、application/json：
   {"version":"<Core构建版本>","listen":"127.0.0.1:7650","admin_api_major":1}
   manage.Options 追加 CoreInfo func() (version, listen string)，在请求时取同一 Core 的只读运行信息；admin_api_major 使用现有 manage.AdminAPIMajor。version 的来源按起草者备注由审查者确认；listen 为 listener.Addr().String() 的实际值，含启动覆盖与端口0分配结果。
4. 保持 internal/core/manage（package manage）的现有 Go 调用：
   func (s *Service) Install(ctx context.Context, req InstallRequest) (result Package, err error)
   func (s *Service) Update(ctx context.Context, id string, req InstallRequest) (err error)
   func (s *Service) Handler(auth *adminauth.Auth) http.Handler
   上传层只构造现有 InstallRequest，候选编排继续由管理服务所有者实现。无新增 proto/SDK 公开声明。

行为：
- 会话取回只读取当前 Cookie 对应的原 CSRF 与固定期限；请求成功后会话ID、CSRF、过期时刻、认证代次及 Cookie 都保持原值。返回字段只含 CSRF 与期限；使用原 csrf_token 和同源 Origin 可完成既有 Cookie 变更请求。
- 上传解码在现有 auth.Require 之后执行，Cookie 修改请求先验 Origin/CSRF，Bearer 沿用现有规则。所有文件字节经统一管理服务完成 release、签名、归档、manifest、来源、身份、平台与依赖检查；确认与预检全部发生在停止、卸载或修改原组件之前。
- 使用标准库流式 multipart 读取，避免把整个归档读入内存。临时目录位于现有 data_dir/staging 下，权限0700，临时文件0600；文件名由服务端生成。archive 大小沿用 packages.Limits.MaxArchiveBytes，解压和条目上限交由既有 Preflight。新增解析仅施加契约已有的限制。
- 临时目录生命周期覆盖整个 Install/Update 调用。成功、400/403/409、候选操作失败、客户端中断和文件写入失败均清理该请求目录；关闭文件与清理的实际错误保留在管理错误/日志中。JSON 上传确认重发规则与现有服务一致；处理结束后归档源可删除，已安装的不可变目录和由 Journal 接管的资源由各自所有者管理。
- 有签名时对 release.json 原文验签；错误签名、公钥与来源绑定按服务原码失败。无签名包继续走 D-81 的显式确认与其他完整预检，文档分开说明两条路径。更新继续检查已绑定来源、组件身份、公钥、pin 和依赖影响。
- GET /api/v1/core 经管理认证；返回实际运行元数据，字段中省略配置、token、secret、DSN、安装路径等无关信息。Admin API 主版本保持1；WebUI 静态入口与既有登录/管理分流保持现状。
- multipart Update 与 JSON Update 进入同一个 manage.Service.Update；当前 Candidates 返回的 UNSUPPORTED 原样传播（HTTP 501）。经 HTTP 真实更新成功的验收由 goal-41 完成。

测试（按编码规范「测试提交规则」）：
- 提交：根包 admin_upload_test.go 的一个 TestAdminSessionAndUploads（//go:build unix）。经真实 Start、公开 Listener、SQLite、签名组件包和文件系统：未登录 session 返回401；登录后取回的 CSRF 与原值/到期值一致且未续期，并能完成 Cookie 变更请求；带 Cookie/CSRF 上传安装包，遇首次信任409后重发确认成功，GET packages 返回真实安装；上传同一来源的新版本时预检与确认照常执行，更新请求返回现有的 UNSUPPORTED（501）且旧组件状态与数据保持；缺少必需文件或错误签名时拒绝且旧组件状态/数据保持，所有分支检查本请求上传临时目录已清理；GET core 的监听地址为实际分配地址。理由：同一集成守住浏览器凭证、文件暂存、管理预检/确认与真实更新的交接。根包复用已有包生成辅助函数；新增 SDK/测试插件源仅放 t.TempDir()，按 D-86 保持 sdk/、supervisor/ 与 testdata 的现有测试范围。
- 本地自测（*_local_test.go，PR 中记录结果）：只有 Bearer 的 session、过期/登出/reset 代次、状态库故障；GET session 无 Set-Cookie/no-store；缺 CSRF/错误Origin在读取上传前拒绝；part 任意顺序、重复/未知part、缺 archive/release、unsigned确认路径、signed缺公钥/坏公钥、保留清单字节；边界大小沿用既有上限、请求中断/写入失败/清理失败；有害上传文件名；409 后重传；现有 JSON install/update 行为；缺少正式候选接线时完整返回501；总览的启动覆盖/端口0、版本值、未认证拒绝和响应字段检查。

完成条件：gofmt -l 对改动 Go 文件无输出；go vet ./...；go test -race ./...；go build ./...。ADMIN-API 字段与实际 HTTP 响应一致；V-40/V-47 的 API 证据写明命令、真实 Core/组件进程、平台与结果。会话/CSRF、验签/来源、旧组件保护、临时文件清理成立，成立后完成本 goal。

不做：WebUI 登录与上传页面（T-09.8）；候选 Journal/Commit/恢复实现（T-09.5～T-09.7）；管理 API 中配置、工具策略或 Agent 会话的列出/清除（T-09.14）；插件页面桥、额外认证体系、上传文件持久缓存及新增公开 SDK 权限 API。
```

## 第 14 波

### goal-149 · T-02.5 CRITICAL 日志级别

状态：已完成（PR #209），合并提交 `d75e43d`。以下保留派发时语句。派发时：前置 T-02.1 与 T-09.10 已合并。2026-11-17 按 D-158 新增。

```text
/goal 使用 $alonabot-execute 推进 goal-149（切片 T-02.5：CRITICAL 日志级别）。goal 目录 goal-149，分支 goal-149，worktree .worktrees/goal-149，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：日志级别在 ERROR 之上增加 CRITICAL：协议枚举、SDK 常量与输出、Core 的解析、过滤、控制台与最近记录、设置与 CLI 都接受它；Core 在 D-158 列出的情况改用 CRITICAL。

依据：PLAN T-02 步骤 1（日志）；CONTRACTS §14.1、§14.3；WIRE-FORMAT §3.2；DECISIONS D-158；TEST-MATRIX V-129。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/control.proto：LogLevel 追加 `LOG_LEVEL_CRITICAL = 6`（OFF 保持 5）；生成产物；contracts/wirespec_test.go
- sdk/log.go（或现有日志文件）：导出常量与级别文本
- internal/core/logging/：解析、过滤、控制台、JSON 行与最近记录
- internal/core/settings/ 与 internal/core/manage/log_api.go：log.level、log.sources 与 PUT /logs/levels 接受 critical
- cmd/alonabot/：`--log-level`、`log-level` 与 `logs --level` 接受 critical
- D-158 列出的 Core 记录位置：状态库无法写入或核对失败、core_config.json 不是合法 JSON 而拒绝启动、core_config.json 写回失败、提交后的恢复失败、因故障退出前的最后一条
- 上述的测试文件

接口：package sdk：const LevelCritical = slog.Level(12)

行为：级别由低到高 DEBUG、INFO、WARN、ERROR、CRITICAL、OFF；JSON 行与控制台写作 `CRITICAL`；按来源设置为 critical 时只输出 CRITICAL；Bootstrap 下发 CRITICAL 时 SDK 按它过滤。插件崩溃与自动停用仍是 ERROR。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/logs_e2e_test.go 新增一个子测试：测试插件以 sdk.LevelCritical 记录一条，`logs --level critical` 只列出它；全局设为 critical 后 ERROR 不再出现在最近记录中。
- 本地自测（*_local_test.go，结果写入 PR）：各处解析 critical 与非法值；控制台文本；D-158 的每个记录点改用 CRITICAL。

完成条件：gofmt -l 无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...。

不做：日志页的颜色与计数（goal-94）；把插件崩溃改为 CRITICAL。
```

### goal-14 · T-04.5 示例升级与 T-04 验收

状态：已完成（[PR #191](https://github.com/AlonaBot/AlonaBot/pull/191)，合并提交 `862b7be`），证据见 TEST-MATRIX §3.2。合入：echo 与 media 示例按链式注册与 T-04.11～T-04.13 的 API 升级（命令组、子命令、别名、可选参数、Regex、keyword、Raw、结构体参数、子插件、带类型的 Service 与 Event Bus、i18n）；`cmd/alonabot/t04_e2e_test.go` 的 `TestSDKExamples`；SDK 的 schema_id 对齐：`UseGenerated` 记录入口包，go test 中只有入口包类型对齐到 `main.<Name>`（DECISIONS 2026-11-13 变更记录）。顶层 check/build 对示例的补证归 T-09.3。以下为派发时的语句。

按 main 核对：

- `sdk/plugin.go`：`NewPlugin() *Plugin`、`(*Plugin).Command(path ...string) *CommandBuilder`、`Group(name string) *CommandGroup`、`Regex(routeID, expr string) *CommandBuilder`、`Keyword(routeID string, patterns ...string) *CommandBuilder`、`Raw(routeID string) *CommandBuilder`、`On(routeID string, eventType EventType) *EventBuilder`。`sdk/generated.go`：`UseGenerated(p *Plugin, g Generated)`；每个示例已有 `func New() *sdk.Plugin`，显式绑定 `alonabotGenerated`。生成器只生成资源与参数名，路由在 New 中注册。
- `sdk/registry.go`（已用 `go doc ./sdk CommandBuilder` 复核）：`Alias(path ...string)`、`Args(names ...string)`、`Optional(name string, def ...string)`、`RawTail()`、`UsageKey(key string)`、`Silent()`、`IgnoreCase()`、`Permission(p Permission)`、`Rule(rules ...Rule)`、`Continue()` 均返回 `*CommandBuilder`，`Handle(handler any)` 结束注册。`sdk/arguments.go` 接受 `func(*sdk.Context, ...) error` 或 `func(*sdk.Context, ...) (message.Content, error)`；结构体参数独占 Context 后的位置，按 `arg`/`default` 标签绑定。
- `sdk/permission.go`：`Permission` 为值对象；`User`、`Admin`、`ChatAdmin`、`Anyone` 为值，`PermissionName(name string) Permission`、`(Permission).Or(others ...Permission) Permission`、`(*Context).Has(p Permission) bool`。`sdk/rule.go`：`type Rule func(*Context) (bool, error)`、`ToMe() Rule`、`NotCommand() Rule`；`sdk/sdk.go`：`(*Context).Reply(text string) error`、`CommandConsumed() bool`。事件回声与新增命令共享消息时，按 CONTRACTS §2.2 的消费状态处理，避免一条命令重复回声。
- `sdk/subplugin.go`：`(*Plugin).Sub(id string) *SubPlugin`、`(*SubPlugin).Permission(p Permission) *SubPlugin`、`Config() config.Values`；子插件 ID 不添加到命令路径或 route_id。`sdk/continuation.go`：`(*Context).Wait(timeout time.Duration) (*Context, error)`；builder 的 `.Continue()` 把 Handler 放进延续调用，原调用继续传播。
- `sdk/services.go`：`JSON[In, Out any](target string) Service[In, Out]`、`Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out]`、`(*ServiceBuilder[In, Out]).Handle(h func(*Context, In) (Out, error))`、`(Service[In, Out]).Call(ctx context.Context, in In) (Out, error)`。`sdk/bus.go`：`BusEvent[T any](name string) BusEventType[T]`、`Subscribe[T any](owner Registrar, e BusEventType[T]) *Subscription[T]`、`(*Subscription[T]).ID(routeID string) *Subscription[T]`、`Handle(h func(*Context, T) error)`、`(BusEventType[T]).Publish(ctx context.Context, payload T) (eventID string, err error)`。main 的提供方身份核对在握手后执行（`checkServiceTargets`），与 CONTRACTS §3 的 Collect 表述存在差异。
- `internal/core/policy/command.go`：`(*Policy).SetOverride(ctx context.Context, pluginID string, s Scope, o Override) error`、`Overrides(ctx context.Context, pluginID string) ([]ScopedOverride, error)`；`Scope{Kind, Group, RouteID}`，`Override{Permissions []string, Enabled *bool}`；权限 nil 为继承、显式空集合为放宽。`internal/core/router/index.go`：`(*Router).Commands() *CommandIndex`、`(*CommandIndex).Lookup(tokens []string) (CommandEntry, int, bool)`、`Shadows() []Shadow`；用这些真实入口取策略与遮蔽证据。
- `examples/plugins/echo/main.go` 已有保留 text/image 的事件回声；`examples/plugins/media/main.go` 已有 `/media` 上传、相册、带类型按钮、SQLite 中的消息句柄与编辑。`cmd/alonabot/telegram_e2e_test.go` 的 `TestTelegramDevelopment` 已通过真实 CLI/Core/官方 Adapter/两个示例进程及本地 Bot API 覆盖这条链；本切片保留原业务路径和断言。
- `sdk/generate/generate.go`：`Scan(ctx context.Context, opt Options) (*Result, error)`，`Options{Dir, GOOS, GOARCH, Tags}`，`Result{PluginConstructor, AdapterType, ConfigSchema, Source, Diagnostics}`。`cmd/alonabot/commands.go` 当前注册 `run` 与 `plugin generate/check`；`cmd/alonabot/plugin.go` 的 `plugin check --dir` 只读比较生成结果。顶层 `init/build/check/test/pack` 尚属 T-09.3 / goal-38；`internal/pluginkit/build.go` 的 `BuildDev(ctx context.Context, dir, outDir string) (Build, error)` 已由 `run <目录...>` 调用并实际编译。
- 在途冲突：按本次任务提供的信息，PR #183（goal-82）会新增五个 `examples/plugins/*/main_test.go`，本切片的源文件、注册描述或生成资源变化可能影响它的断言；由审查者在后合并者分支协调，只迁移必要调用与预期，不再新增同类 plugintest 用例。PR #181 涉及根包/manage HTTP，#182 涉及 broker/router/hook/agent，goal-42 涉及 journal/state/gc/assembly；本切片仅复用这些生产入口，不把其实现纳入范围。共享追加点为 `spec/TEST-MATRIX.md`；goal-70 的开发文档与本切片示例说明需各守所属段落。

```text
/goal 使用 $alonabot-execute 推进 goal-14（切片 T-04.5：示例升级与 T-04 验收）。goal 目录 goal-14，分支 goal-14，worktree .worktrees/goal-14，从最新 origin/main 创建，须包含直接前置 PR #180（c128000）、#42（62d5afa）、#47（5cb1cfd）、#88（c8dd516）、#89（3e37cd6）、#141（ac4aded）、#110（6aed25d）、#159（7872102）、#154（43a6135）、#158（464ccfd），并沿用 #153/#162 的真实延续调用与 Rule 实现。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：用 main 上同一个 SDK 完整升级 echo 与 media 示例，展示显式链式注册、命令组/子命令/别名、文本匹配、Raw、可选参数、显式实例方法、结构体参数与子插件；采用 c.Reply、builder.Continue() 及带类型的 Service/Event Bus 声明。以实际编译、激活和收发完成 T-04 步骤 6、7 的剩余验收，汇总 V-10、V-11、V-12、V-43、V-49 的 T-04 分项及 V-112、V-113。

依据：spec/plan/core.md 的 T-04 步骤 6、7、步骤 9/11/12/13 对现行 SDK 的修订及「失败停止条件」；spec/CONTRACTS-v0.1.md §2.1～§2.4、§3、§5.4、§8、§9.1、§9.3、§15.4；spec/MESSAGE-MODEL.md §1.1、§3、§4、§6.1～§6.4；spec/WIRE-FORMAT.md §3.2～§3.4、§4、§5A；DECISIONS D-69～D-72、D-77、D-79、D-86、D-112、D-114～D-116、D-118～D-123、D-137，消费状态沿用 main 的 D-141；TEST-MATRIX §1 的 V-10/V-11/V-12/V-43/V-49/V-112/V-113、§2 的分项归属、§3.2 的前置证据。

范围（只新增或修改这些）：
- examples/plugins/echo/：main.go 及本示例需要的普通 Go 源文件、plugin.yaml、config_schema.json、i18n/*.json、README.md、alonabot_gen.go；保留事件回送文本与入站图片，补命令、结构体参数、子插件、声明式 Service/Bus 与说明。
- examples/plugins/media/：main.go 及本示例需要的普通 Go 源文件、plugin.yaml、config_schema.json、i18n/*.json、README.md、alonabot_gen.go；保留上传/相册/按钮/消息句柄/编辑，补结构体参数与子插件的实际用法。保留现有 assets/sample.png。
- 新增 cmd/alonabot/t04_e2e_test.go（//go:build unix），只含下述一个主路径集成测试及其内部辅助；复用 e2e_test.go、dev_e2e_test.go、telegram_e2e_test.go 已有启动与 Bot API 辅助。
- 本地自测只写 sdk/generate/t04_local_test.go、sdk/t04_local_test.go、根包 t04_local_test.go 或上述示例的 *_local_test.go；临时构建工程写 t.TempDir()，不新增 testdata 测试插件。
- spec/TEST-MATRIX.md §3.2：追加 T-04.5 的分项汇总与实际新证据，引用既有行，不覆盖历史记录。
- examples/plugins/*/main_test.go 由 goal-82 持有；本语句不授权新增或改写这些文件，必要迁移由审查者协调到后合并者。
允许的新跨模块边：无。cmd/alonabot 集成测试复用已允许的 adapters/telegram/telegramtest；示例只经 sdk/... 与 contracts，测试使用既有装配，不导出生产测试入口。

接口（照此使用）：不新增框架公开符号；两个示例已有的 func New() *sdk.Plugin 保持签名。每次 New 创建独立对象，调用 sdk.UseGenerated(p, alonabotGenerated)，完成注册后返回；main 沿用 sdk.Run(New())，正式入口生成归 T-09.3。包级函数与对象方法值均以 builder.Handle 显式绑定。
- 命令用 p.Command(path...)、p.Group、p.Regex、p.Keyword、p.Raw 与各 builder，权限用 sdk.Permission 值、sdk.User/Admin/ChatAdmin/Anyone、Or；普通 Rule 用 sdk.Rule/ToMe，回声事件按已合并的 sdk.NotCommand() 读取 Core 消费状态。Regex 捕获用 sdk/dispatch.go 的 (*Context).Match() []string。
- 结构体 Handler 为 func(*sdk.Context, Args) error 或 (message.Content, error)；可选字段用指针或 default，raw 用末位 arg 标签。该路由不再同时调用 Args/Optional/RawTail；位置参数示例继续用这些 builder 方法。
- Service 用 sdk.JSON[In, Out](完整目标)、sdk.Serve(owner, declaration)...Handle，调用用同一类型声明的 Call；Bus 用 sdk.BusEvent[T](完整事件名)、sdk.Subscribe(owner, declaration)...Handle 与 declaration.Publish。输入/输出是包级具名类型，Schema 交给现有生成器；服务提供方身份按 main 的握手后检查执行。仅在有效 Handler 或延续调用内做业务调用。
- 纯文本主动回复用 c.Reply；媒体与需要 SendResult 的路径继续 c.Send(message.Content)。两种 Handler 签名、部分发送回执与错误均保留。

行为与验收矩阵：
1) echo：保留当前未消费消息的文本/图片回声；增加 /echo 的纯显式注册、/note add <text:raw>、/note list、/calc <a:int> [b:int=1]、Regex（例如 ^roll (\d+)d(\d+)$）、keyword 与 Raw。/note add、/note list 分别以 Alias("n", "add")、Alias("n", "list") 提供 /n 子命令别名。用显式创建的状态对象保存 note 列表并绑定至少一个实例方法。/calc 用结构体参数示范默认值，note 用位置参数示范 RawTail；至少一条路由属于 p.Sub。media 也用一条实际命令展示结构体参数与子插件；子插件标识与命令路径分别按契约声明。
2) 文本匹配按契约固定排序，Regex 捕获从 c.Match() 取得；普通命令被识别后，禁用、权限拒绝、参数错误或 Rule false/error 都保留消费状态。参数错误只发一次 usage，Silent 不发；无 Actor、admin、chat_admin 与平台权限的结果按权限名集合判定。keyword 使用专属示例短语，Raw 用 sdk.Rule 限定专属示例输入，其他输入的 Rule 返回 false，使普通文本/图片继续进入原有回声事件。事件回声避免重复回复已消费消息；保留 goal-9 对普通文本、图片与 /media 的现有行为。
3) 至少一条带可信 Actor 的交互路由用 builder.Continue() 与 c.Wait 展示延续调用，原调用继续传播；Handler 在延续调用中主动 Reply，或按契约发送自动回复一次。旧上下文不带入任意后台 goroutine。以同一类型声明提供一个真实 Service 路由，以及一次 Bus 发布/订阅；订阅只处理业务载荷，不继承用户权限。跨插件调用、Bus 和延续调用的完整失败矩阵引用既有证据，本切片只验证新增示例组合。
4) 构建与真实收发：两个示例经现有生成器产生资源并实际 go build；以 alonabot run <Adapter目录> <echo目录> <media目录> 在临时 data_dir 启动真实 CLI/Core/SDK 进程，使用现有 Telegram 本地 Bot API，验证新增命令的回复、别名/默认值/捕获与注册成功。只比较 alonabot_gen.go 文本不足以完成验收；生成产物只经现有生成入口更新。
5) 当前 main 的只读检查为 alonabot plugin check --dir <目录>，构建证据取实际 go build 与 run 内部 BuildDev；D-77 要求的 alonabot check/build 另列为 T-09.3 的待补入口证据（状态段的审查者裁决）；未实现的顶层命令写「未运行」，不得以旧入口结果填写其通过。
6) 构建拒绝与范围：引用 T-04.2（#42）、T-04.9（#110）、T-04.12（#154）的范围/build tags/GOOS/资源稳定性与新入口证据；补当前写法的本地检查：缺少或错误 New、重复 New 声明、错误 Handler 签名、重复资源/路由/别名、结构体参数非法组合、build tags 选择、外部依赖与残留标记均不参与路由生成、int64 溢出。旧的多根类、多 Config 类型与方法标记诊断按 D-79/D-114/D-76 的变更标为已替代，不恢复扫描器或旧配置类型。
7) 策略：引用 #34 的跨插件固定遮蔽、#47 的命令解析、#88 的服务层即时生效/持久化/禁止优先、#162 的真实 Rule 与命令消费证据；历史测试已随 D-86 删除的，明确写历史位置和当时结果。为 T-04 步骤 7 补当前 main 的本地两真实 Session 联验：同名命令的 Winner/Shadowed，禁用胜出者后路径保持占用；直接 SetOverride 后下一条消息变化、插件 PID 不变；关闭并重启 Core、复用同一临时 data_dir 后策略仍在。使用现有 policy 与 Router 索引，输入/输出/失败记录到 PR；不新增管理 HTTP。
8) V-112 汇总 #154 与新示例中的对象资源、结构体参数、子插件；V-113 分别引用 #141（SDK 类型）、#153（延续）、#158（Service/Bus 声明），加入本切片新增示例的实际运行证据。goal-9 的媒体/Telegram 故障矩阵与 goal-82 的示例 plugintest 测试按各自证据引用，不复制。V-12 的浏览器管理操作仍由 T-09 记录。

测试（按编码规范「测试提交规则」）：
- 提交：仅 cmd/alonabot/t04_e2e_test.go 的 TestSDKExamples 一个主路径集成测试，可用表驱动子场景覆盖新命令；理由：真实目录构建、CLI/Core/官方 Adapter/示例进程与本地 Bot API 首次联验本切片新增组合，补足已有 Telegram 主路径尚未覆盖的命令示例。复用已批准辅助，不增加 TestMain 分流或 SDK 测试插件。
- 既有测试：保持 TestTelegramDevelopment 原断言；若 goal-82 先合并，由审查者把必要的示例测试迁移交给后合并者，原有意图与覆盖保留。
- 本地自测（*_local_test.go，不提交，PR 写明命令与结果）：第 3/6/7 项的延续、Service/Bus、新签名诊断、构建选择、消费与策略；覆盖当前写法即可，已有有效证据的分项直接引用。sdk/ 与 internal/core/supervisor/ 不增加提交测试，不新增 testdata 测试插件。两个示例的 Linux/Windows 交叉构建只登记构建结果。

完成条件：改动 Go 文件 gofmt -l 无输出；go vet ./...、go test -race ./...、go build ./... 通过；go test -race -count=1 ./cmd/alonabot -run '^(TestSDKExamples|TestTelegramDevelopment)$' 作为上述 race 的定向取证，已有同候选结果时复用；现阶段 go run ./cmd/alonabot plugin check --dir examples/plugins/echo 与 media 均通过。TEST-MATRIX §3.2 按 §3 记录测试位置、完整命令、平台/工具版本、日志位置、实际结果和剩余归属，T-04 的构建/策略与 V-112/V-113 汇总逐项有出处。契约要求的顶层 check/build 按状态段的审查者裁决处理，不伪记通过。

停止条件：发现扫描依赖或隐式构造接收者、路由依启动顺序改变、重复路由静默覆盖、已识别命令回落、子项绕过父禁用、权限混淆或策略重启丢失，按 PLAN T-04 停止并交审查者；失败位于范围外时提供最小复现，先由所属切片修复，不用示例绕开。

不做：生成器/SDK/Router/Policy 的新设计；旧构建标记、TrustedRole/Permission 枚举、RouteOption 或 ChatAdmin Rule；testdata 插件与绑定 SDK 的提交测试；复制 goal-9/goal-82 已有矩阵；T-09.3 的顶层 CLI 实现、WebUI 策略页面、真实 Telegram 补测与 T-04.14 命令名覆盖的新功能。
```

### goal-22 · T-06.4 Session 清理与 Service 端到端

状态：已完成（[PR #164](https://github.com/AlonaBot/AlonaBot/pull/164)，合并提交 `039802f`），证据见 TEST-MATRIX §3.2。合入：`internal/core/metrics` 的 `Counters.Add`/`Snapshot`（以已知身份为界）；broker、bus、router 的 Options 增加 Metrics；Broker 为每次转发保存取消入口，Session 进入 STOPPING、被替换或断链时取消相关转发并归还预算；根包 `service_chain_test.go` 的 TCP、stdio 矩阵。执行中审查者授权 supervisor 透传停止钩子。派发时：前置 T-06.2（[PR #161](https://github.com/AlonaBot/AlonaBot/pull/161)，合并提交 `b4610e1`）与 T-06.3（PR #56）已合并。2026-11-12 审查者按 main `96856ec` 改写，要点如下：
- 已有的清理：根包 `assembly.detach` 在组件停止或 Session 替换时依次调用 `router.Detach`、`bus.Detach`、`scheduler.Detach`、`httpproxy.Server.Unbind` 与 Adapter 的 `remote.Unbind`；延续调用与等待随 `continuation.Manager.Release(caller)` 释放（T-05.4）；插件的入站请求在 Session 结束时由 wire 取消，Broker 转发时用的就是这个 ctx。
- Broker（T-06.2）：`broker.New(broker.Options{Router, RequestTimeout, MaxDepth})`，`b.calls` 以子调用的 Invocation 记录来源；CALL_CYCLE 与深度在调用前返回；额满由 wire 立即返回 RESOURCE_EXHAUSTED。
- Bus（T-06.3）：`bus.New(bus.Options{QueueEvents, QueueBytes})`、`Register`、`Attach`、`Detach`、`Stats(pluginID) Stats{Published, Delivered, Dropped}`（丢弃不分原因）；wire 的接收丢弃为 `wire.Stats.DroppedEvents`。
- 多进程装配：根包 `pipeline_test.go`（T-05.5）的 `newPipeline` 经 assemble 取得生产装配，以测试二进制运行插件与回环 Adapter 进程；本切片的测试在同一个包里复用它，并按需加入 stdio 传输。
- 与 PR #163（goal-39）同改 assembly.go 与 pipeline_test.go 的装配（#163 让 assemble 接收共用的 settings.File、HTTP 代理改在 Start 中创建），后合并者同步 main。
- 测试插件按 main 上的写法：插件对象、`sdk.JSON[In, Out](target)` 与 `sdk.Serve`/`Call`（T-04.13）、`Plugin.BeforeSend`（T-05.3）。

```text
/goal 使用 $alonabot-execute 推进 goal-22（切片 T-06.4：Session 清理与 Service 端到端）。goal 目录 goal-22，分支 goal-22，worktree .worktrees/goal-22，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：核对并补齐 Session 停止或失效时的清理：停止新的根调用，取消在途父链，清理路由目标、Bus 订阅队列、Broker 映射与预算，不把未完成的工作转交给同插件的新 Session；错误只返回安全的 ErrorInfo 与 trace。记录按原因区分的拒绝、取消、离线与丢弃计数。在 TCP 与 stdio 上完成 A/B/C 嵌套与故障的端到端验收（含 Hook→Service→Send），在 TEST-MATRIX §3.2 记录 T-06 证据。

依据：PLAN T-06 步骤 6、7 与「停止和失败条件」；CONTRACTS §3、§1.2、§4.2、§14.5；TEST-MATRIX V-15、V-17、V-18、V-19、V-42。

范围（只新增或修改这些）：
- 新增 internal/core/metrics/（按原因的计数，只在内存中，供管理面读取）
- internal/core/broker/、internal/core/bus/、internal/core/router/：Options 增加计数器并在各拒绝、取消与丢弃处计数；核对中发现的清理缺口在这些包内补齐
- 根包 assembly.go：创建计数器并传给上述三者（只加装配点）
- 新增根包 service_chain_test.go（//go:build unix），复用 pipeline_test.go 的装配辅助（可以为它加入 stdio 选项）
- spec/TEST-MATRIX.md §3.2：追加 T-06 的证据行（本切片获授权修改这一处）
允许的新跨模块边：internal/core/broker、bus、router -> internal/core/metrics；根包 -> internal/core/metrics。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/metrics
   type Reason string // 取值：denied、cancelled、offline、dropped_queue、dropped_budget、cycle、depth
   type Kind string   // 取值：service、bus、event
   type Counters struct{ ... } // 零值可用，方法可以并发调用
   func (c *Counters) Add(pluginID string, kind Kind, reason Reason)
   func (c *Counters) Snapshot() map[string]map[Kind]map[Reason]uint64 // plugin → kind → reason → count，返回副本
broker.Options、bus.Options 与 router.Options 增加 Metrics *metrics.Counters（nil 时不计数）。

行为：
- 冻结：插件进入 Draining 后，Broker 对以它为目标的新调用返回 UNAVAILABLE（计 offline），Bus 不再向它入队，Router 不再派发给它。核对现有行为，缺口在对应包内补齐。
- 失效（STOPPING 或断链）：以它为调用方或目标的转发请求被取消，调用方得到 UNAVAILABLE 或 CANCELLED（计 cancelled），占用的预算归还；Bus 丢弃它的订阅队列（计 offline）；Router 移除它的路由目标，等待与延续调用按 Release 释放。不转交给新 Session，新 Session 从空状态开始。
- 错误：返回给调用方的 ErrorInfo 只含 code、safe_message 与 trace_id；远端堆栈、Secret、文件路径只进 Core 日志。
- 计数：每次访问拒绝（denied）、取消、离线、队列满丢弃（dropped_queue）、预算不足丢弃（dropped_budget）、CALL_CYCLE（cycle）、深度超限（depth）都计数；不持久化，不建设可靠事件存储或自动重试。
- 验收矩阵（service_chain_test.go，TCP 与 stdio 各运行一遍；测试插件由测试二进制按环境变量扮演 A、B、C）：
  1) A→B→C、A→B→A（CALL_CYCLE）、深度边界、两条独立并发链：结果、Actor/Target、deadline 与取消在每一跳一致。
  2) A 的出站 Hook（BeforeSend）→ B 的 Service → Send：FAILED_PRECONDITION（闭合 V-15）；A 原本的发送照常完成。
  3) B/C 挂起、进程退出、满额、畸形输入与输出、排队前取消：没有死锁，没有重复的业务效果（以目标插件的调用计数证明）；B 退出后它的新 Session 不收到旧的未完成请求。
  4) Bus：慢订阅者、断线、条数满与字节满同时发生时，新事件按规则丢弃，其他订阅者继续收到；计数中有对应的原因。
  5) 每项在日志中记录实际进程、传输与命令。
- 证据：TEST-MATRIX §3.2 追加 T-06 一行，写明测试位置、命令、平台、传输与结果。

测试（按编码规范「测试提交规则」）：
- 提交：根包 service_chain_test.go 的上述矩阵（多列理由：T-06 的验收矩阵本身）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：Counters 的并发与 Snapshot 副本；失效后预算归还到初始值；ErrorInfo 中不含路径与堆栈；在 -race 下以 -count=5 运行矩阵无偶发失败。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：管理面对计数的读取接口（T-09.14）；可靠事件存储与服务重试（契约不提供）；修改 pipeline_test.go 的既有断言。
```

### goal-55 · T-11.2 工具调用与用户确认

状态：已完成（[PR #165](https://github.com/AlonaBot/AlonaBot/pull/165)，合并提交 `330164f`），证据见 TEST-MATRIX §3.2。合入：服务 `core.tool.invoke`、`core.tool.confirm`；`sdk.InvokeTool(ctx, name, arguments)`、`sdk.ConfirmOperation(ctx, operation, summary)`；`internal/core/tool` 的 `InvokeOptions{Call, Adapters, Buttons, ConfirmTimeout, DefaultLocale}` 与 `(*Catalog).RegisterInvoke(m, opt)`；`internal/core/buttons` 的 `IssueConfirm`、`RevokeConfirm`、`OnConfirm`、`InterceptConfirm`；`router.Options.Intercept`（确认点击在派发槽之外处理）。

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
- internal/core/state/schema.sql 与 buttons.go：button_token 增加类型列（plugin 或 confirm，写进唯一的初始迁移），读写同步
- 根包 assembly.go：注册两个新服务、把确认点击接到 buttons 的解析（只加装配点）
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
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：参数截断；token 只能用一次；父调用结束时记录作废；zh-CN 文案；策略在两次调用之间修改；根包 tool_invoke_local_test.go（//go:build unix），复用 pipeline_test.go 的装配：测试二进制扮演工具服务插件与调用方插件（收到消息时调用 InvokeTool 或 ConfirmOperation 并回复结果），回环 Adapter 声明 message.buttons 与 message.edit：allow 工具直接返回结果；deny 工具返回 PERMISSION_DENIED 且工具未执行；confirm 工具：memory 收到带两个按钮的确认消息，内容含实际 JSON 参数；同一用户点「允许」后工具执行并返回结果；另一用户先点击被拒、不影响结果；点「拒绝」返回 PERMISSION_DENIED 且工具未执行；超时（注入较短的 TOOL_CONFIRM_TIMEOUT）返回 DEADLINE_EXCEEDED；确认消息没有经过插件的 BeforeSend Hook（测试插件注册了会改链的 BeforeSend，确认消息不受影响）；没有 Actor 的调用（scheduled fixture）遇到 confirm 返回 PERMISSION_DENIED，没有发消息；不支持 buttons 的实例返回 UNSUPPORTED；arguments 不符合 input schema 时 INVALID_ARGUMENT；ConfirmOperation 展示 summary 而不是参数。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/... ./sdk/...；go build ./...。

不做：Agent 的工具循环与 MCP（T-12.3）；「总是允许」（契约不提供）；管理页面的工具策略（T-09.9）。
```

### goal-70 · T-13.4 开发运行示例与验收

状态：已完成（[PR #186](https://github.com/AlonaBot/AlonaBot/pull/186)，合并提交 `c38a294`），证据见 TEST-MATRIX §3.2。合入：examples/devbot 整理为 init 契约的工程形状（可导入插件包、命令、i18n、.gitignore、README 与 core_config.example.json），回环 Adapter 的私有注入路由经真实 Deliver 交付并观察回复（同一实例写入串行化，未接纳时返回 503）；`cmd/alonabot/t13_e2e_test.go` 的 `TestDevelopmentExample`。init 生成与模板比对仍为「未运行，T-09.3 补证」。以下为派发时的语句。

按 main 核对：

- `alonabot.go`：`Run(opts ...Option) error`、`Start(ctx context.Context, opts ...Option) (*Core, error)`（已用 `go doc . Start` 复核）、`(*Core).Wait() error`、`Plugin(root *sdk.Plugin) Option`、`Adapter(a adapter.Adapter) Option`、`DataDir(dir string) Option`、`Args(values ...string) Option`、`Listen(addr string) Option`、`LogLevel(level slog.Level) Option`、`LogFormat(format string) Option`。`Executable(path string, packageManifest []byte) Option`、`Stdio() Option` 供开发 CLI；`(*Core).Restart(ctx context.Context, id string) error`、`RestartInstance(ctx context.Context, adapterID, instanceID string) error`、`ReloadSettings(ctx context.Context) error` 已存在。库模式使用 Start/Run，热部署由 CLI 轮询驱动。
- `cmd/alonabot/commands.go`：`run [directories...]` 支持 `--reload`、`--stdio`、`--data-dir`、`--listen`、日志覆盖；`cmd/alonabot/dev.go` 的 `runDevelopment(dirs []string, dataDir string, reload, stdio bool, interval time.Duration, options []alonabot.Option) error` 构建后调用共用根包。`internal/pluginkit/build.go`：`BuildDev(ctx context.Context, dir, outDir string) (Build, error)`，`Build{Executable, Manifest, PackageManifest, Digest, Role, ID, AdapterID}`；以 overlay 生成资源与薄入口，不写工程源文件。没有公开的 `alonabot.Reload` 选项。
- `examples/devbot/main.go` 已用 Start、DataDir、Plugin(plugin.New())、Adapter(adapter.New()) 与信号 ctx 运行同进程示例；`examples/devbot/plugin/plugin.go` 的 `New() *sdk.Plugin` 显式绑定生成资源，按 `greetings` 配置写 startup 表并回声；`examples/devbot/adapter/adapter.go` 的 `New() adapter.Adapter` 返回可导入回环 Adapter，但当前仅实现 Start/Send/Stop，没有 Deliver 消息的交互入口。两个包已有工程 plugin.yaml 与 alonabot_gen.go；Adapter 的实例 schema 是 `{"$version":1}`。
- `internal/core/settings/instances.go`：`(*File).PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []InstanceProblem, error)`；实例来自 `core_config.json` 的 `adapters.<adapter_id>.instances`，`package` 绑定 Adapter 包 ID。devbot 的现有 ID 分别为 `alonabot-plugin-devbot`、`alonabot-adapter-devbot`，adapter_id 为 `devloop`；开发配置在 `<data_dir>/plugins/<id>/config.json`，受管库是 `dev.sqlite`，代次为 `dev:<组件 ID>`。
- `sdk/plugin.go`、`sdk/generated.go` 提供 `NewPlugin() *Plugin`、`UseGenerated(p *Plugin, g Generated)`；`sdk/sdk.go` 的 `(*Context).Reply(text string) error`、`sdk/rule.go` 的 `NotCommand() Rule` 可供新工程示例使用。`sdk/generate/generate.go` 的 `Scan(ctx context.Context, opt Options) (*Result, error)` 支持入口包与构建选择；插件资源绑定对象，Adapter 资源在可导入包中通过 `UseGeneratedAdapter(a adapter.Adapter, g Generated)` 登记。CLI 能为可导入包生成独立进程薄入口。
- `alonabot_test.go` 的 `TestLibraryLifecycle` 已构建并运行真实 devbot：核对两个 source=dev 安装记录与固定代次，验证缺省配置、required 拒绝、同一 data_dir 两次启动的配置/SQLite 记录与 STOP_TOTAL 内停止。TEST-MATRIX §3.2 的 #149/#160 行同时保留本地验证按钮跨重启、停止超时、身份预检等历史证据。
- `cmd/alonabot/dev_e2e_test.go` 的 `TestDevelopmentReload` 已验证独立进程开发插件/Adapter、源码成功替换只改变插件 PID、编译失败旧组件继续、只改 config.json 保持二进制字节与 mtime、改 config_schema.json 重建并生效；`e2e_test.go` 的 `TestLifecycleOverTCPAndStdio` 已恢复两种传输。#171 的 TEST-MATRIX 行还记录了 Adapter 单实例配置重启、ID 冲突、settings 重新读取等本地证据。该提交测试没有检查 i18n 变化或同时运行已安装组件。
- `cmd/alonabot/telegram_e2e_test.go` 的 `TestTelegramDevelopment` 已验证 Telegram、echo、media 的目录运行和真实进程收发，#180 的 TEST-MATRIX 附录另有 TCP/stdio 故障矩阵及 FIX-3 热重启依赖门禁。本切片引用其证据。`internal/core/manage/candidates.go` 的 Update/Rollback 与 `lifecycle.go` 的 Uninstall 已对 dev 来源返回 developmentError（FAILED_PRECONDITION）。
- 复用测试入口：`cmd/alonabot/e2e_test.go` 的 `startEmptyCore(t *testing.T, dataDir string, extra ...string) *core`、`startCore(...)`、`(*core).management(t *testing.T, dir, token string, args ...string) string`；`dev_e2e_test.go` 的 `developmentWrite/Read/PID`；`manage_e2e_test.go` 的 `findPackage`、`testPackage`。TestMain 已准备签名的 demo/lifecycle 安装包，可直接用于混合运行验收。
- `docs/development.md` 已有「开发运行与热部署」「库模式运行（同一进程）」两节；应补 init→devbot 的完整用法并整合现有内容。`cmd/alonabot/commands.go` 尚无 `init`、顶层 `build/check`；它们归未开始的 T-09.3 / goal-38，不能记录成 main 已可运行的命令。
- 在途冲突：PR #181（goal-92）可能改根包/manage HTTP，#182（goal-61）改 broker/router/hook/agent，goal-42 改 journal/state/gc/assembly；本切片不编辑这些生产文件。PR #183（goal-82）持有 plugintest、scheduler 与五个 `examples/plugins/*/main_test.go`，本切片不进入这些目录。`docs/development.md` 与 `spec/TEST-MATRIX.md` 是潜在共享文档；后合并者保留其他切片段落，具体远端 diff 未在本次任务中查询。

```text
/goal 使用 $alonabot-execute 推进 goal-70（切片 T-13.4：开发运行示例与验收）。goal 目录 goal-70，分支 goal-70，worktree .worktrees/goal-70，从最新 origin/main 创建，须包含 PR #180（c128000）、#160（65ebc5f）、#171（3043c4e）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按 PLAN T-13 步骤 5 把 examples/devbot 整理成组件工程式的插件与回环测试 Adapter，能以 alonabot run <插件目录> <Adapter目录> --reload 开发，并用现有根包示例对同一插件和 Adapter 做同进程调试；补齐开发文档、尚缺的开发入口证据，汇总 V-64/V-65。保留 #149/#160/#171/#180 的有效验收，仅补本语句列出的缺口。

依据：spec/plan/core.md 的 T-13 步骤 5、「失败停止条件」和 V-64/V-65 对照表，步骤 1～4 用于确定已实现入口；spec/CONTRACTS-v0.1.md §1.1、§4.1/§4.2、§5.2、§6.2、§8、§9.1/§9.3、§15.1、§15.4/§15.5、§16 的 DEV_RELOAD_POLL/STOP_TOTAL；spec/MESSAGE-MODEL.md §1/§1.1、§3、§4；spec/WIRE-FORMAT.md §3.1～§3.4、§4；DESIGN §S；DECISIONS D-77、D-78、D-82、D-86、D-114、D-134、D-137，事件消费沿用 main 的 D-141；TEST-MATRIX §1 的 V-64/V-65、§3/§3.2 的 T-13.1/T-13.2/T-13.3/T-09.4/T-03.10 证据。

范围（只新增或修改这些）：
- examples/devbot/main.go：保留同进程调试入口及现有信号/停止流程，按需要完善示例说明，不加入热部署选项。
- examples/devbot/plugin/：plugin.go 与本示例内部普通 Go 源文件、plugin.yaml、config_schema.json、i18n/*.json、.gitignore、alonabot_gen.go；整理为与 init 契约相同的可导入插件包、一个命令和所需资源，保留启动记录与现有配置含义。
- examples/devbot/adapter/：adapter.go 与本示例内部普通 Go 源文件、plugin.yaml、adapter_config_schema.json、alonabot_gen.go；补回环消息交付与可观察发送，使文档能实际驱动一次完整收发。沿用 dev_e2e_test.go 已有回环 Adapter 的私有 HTTP 注入与输出观察方式。
- 新增 examples/devbot/README.md、examples/devbot/core_config.example.json；样例为 core_config.json 的初始内容，实例按 adapters.devloop 配置，不创建 instances/。
- docs/development.md：仅整合「开发运行与热部署」「库模式运行（同一进程）」与紧邻的生成/检查命令说明，补 init→运行→调试的用法和当前命令实现状态；其他切片文档保留。
- 新增 cmd/alonabot/t13_e2e_test.go（//go:build unix），下述一个主路径集成测试及内部辅助，复用既有 CLI/管理/进程辅助；不修改 TestMain。
- 根包 alonabot_test.go：仅在示例迁移确有必要时同步既有 TestLibraryLifecycle 的构建/配置路径，原有断言与数据语义保留。
- 本地自测写根包 t13_local_test.go、cmd/alonabot/t13_local_test.go、examples/devbot 各包的 *_local_test.go；临时工程、data_dir 与日志用测试临时目录。
- spec/TEST-MATRIX.md §3.2：追加 T-13.4 的 V-64/V-65 汇总行与缺口分项证据，保留历史记录。
允许的新跨模块边：无。examples/devbot 沿用已允许的根包与 SDK 依赖，CLI 测试复用现有内部包；sdk、adapters、plugins、internal/core 不 import 根包。示例在同一仓库 module 内复用现有依赖，不新增 module 或第三方依赖。

接口（照此使用）：不新增框架公开符号。保留 examples/devbot/plugin.New() *sdk.Plugin、adapter.New() adapter.Adapter 与两个现有组件 ID/adapter_id，保留根包 Run/Start/Wait 及其 Option 的现有签名。
- 插件包导出 New 并显式绑定生成资源，在 New 内注册命令；入口包无手写 main 的工程可由现有 run/BuildDev 生成薄入口。examples/devbot/main.go 继续用该可导入包做库模式调试。
- 回环 Adapter 声明私有注入路由，实现 sdk/adapter_run.go 中已有的 sdk.AdapterHTTP：HTTP(route func(method, path string) *sdk.RouteBuilder)。消息经真实 adapter.Core.Deliver 交给 Core，回复由真实 Instance.Send 观察；采用既有测试辅助的请求与输出约定，文档写清实际路径、认证与输入格式。只演示正常标准事件，不创建旁路派发。
- 开发运行用 alonabot run examples/devbot/plugin examples/devbot/adapter --data-dir <临时目录> --reload；同进程用 go run ./examples/devbot -data-dir <另一个目录>。两个模式分别运行并使用各自 data_dir，Adapter 身份与资源来自同一可导入包。
- 实例为 core_config.json 的 adapters.devloop.package=alonabot-adapter-devbot 及 instances.<实例名>；配置与数据库为 plugins/alonabot-plugin-devbot/config.json、dev.sqlite。样例只给开发者可直接填写的初始设置，不引入另一份运行配置来源。

行为与验收矩阵：
1) 工程与使用：在保留现有 ID、greetings 配置与 startup 数据语义的基础上整理插件工程，示范 New、资源绑定、命令与 i18n；.gitignore 忽略工程中的 data/。回环 Adapter 可以按文档交付一个带可信 Actor/默认会话的 message.created 并观察回复。Plugin/Adapter 用现有生成器更新 alonabot_gen.go，生成产物随源码提交且能实际编译。
2) init 分项：D-77/CONTRACTS §9.3 要求的创建路径为 alonabot init [目录]，生成可导入插件包、作者 manifest、schema、i18n 与 .gitignore，不生成手写 main。main 尚未实现该命令，现阶段先整理为契约规定的工程形状，文档明确当前可运行的已有工程入口；实际 init 创建与模板一致性登记「未运行，T-09.3 补证」（状态段的审查者裁决）。执行者不在本切片实现 init。
3) 已覆盖的库模式验收：引用 TEST-MATRIX #149/#160 和 TestLibraryLifecycle 的真实生命周期、两个开发组件身份/固定代次、配置补值/required 错误、同一 data_dir 两次启动与 dev.sqlite 保留；引用 #149 的按钮跨重启有效、停止超时返回错误与启动冲突的本地结果。不再新增一组相同测试。修改示例后重跑既有测试，保持原有启动记录、配置原文、无候选副本等断言。
4) 已覆盖的独立进程验收：引用 #171 的 TestDevelopmentReload 与其本地配置/实例/settings/身份预检证据，引用 TestLifecycleOverTCPAndStdio；引用 #180 的 Telegram 目录运行、依赖检查及 FIX-3 热重启依赖门禁。源码变更、编译失败旧进程继续、config-only 不重建与 schema 重建的完整矩阵由这些证据承担，本切片仅按文档手动走一遍示例，不复制其提交测试。
5) 新增主路径缺口：把 examples/devbot 的插件/Adapter 工程复制到测试临时目录并使用临时 data_dir，以真实 CLI --reload 运行，按样例配置启动实例并完成一次收发；通过现有管理 CLI 安装并启用 TestMain 已准备的真实 lifecycle/demo 包。只改开发插件 i18n 资源后，以实际译文变化与插件 PID 变化确认重建/生效；Core PID、回环 Adapter Session/实例和已安装插件的 PID/Session 保持，已安装插件仍能处理一次真实请求或命令。由此补 V-65 尚缺的 i18n 与混合运行不中断证据，也证明示例遵守 T-13 的失败停止条件。
6) 剩余 V-64 本地取证：未捕获的库模式后台 panic 必须在受控子进程中使整个进程非零退出且无重启；SDK Handler 边界内已捕获的 panic 按既有错误语义处理。通过现有管理入口确认 dev 来源的 update/rollback/uninstall 三种操作均返回 FAILED_PRECONDITION，组件仍可用，配置/数据库未改变；已有 #163 的可定位同场景证据可直接引用。对握手/注册/权限/发送与已安装组件的一致性，引用 #149/#171 的共用装配、#162 的可信权限流水线和 #180 的真实发送；新增示例收发如暴露差异，按下面停止条件处理。
7) 文档：先给复制样例设置、已有工程 run --reload 与一条可见回复的完整操作，再给源码/配置/翻译改动、编译错误恢复和同进程调试。写清完整目录、实例 JSON、命令前缀、token 的既有读取/认证方式、缺省 data/、配置与 dev.sqlite、dev 来源限制，以及独立进程/同进程的 panic 与停止语义。复用现有章节，删除 instances/、Reload 选项、库模式热部署、go.mod 只输出 WARN 等旧描述。给用户手动执行的命令使用 fish 语法。
8) 汇总：V-64/V-65 按分项列已合并证据、本切片实际命令与结果、未运行项及承接切片。WebUI/CLI 修改开发组件配置后写回并重启属于 T-09.14，保留该分项未完成状态；真实 init 属 T-09.3，真实 Telegram 与其他 OS 的运行按已有记录写未运行。只据实际证据关闭分项。

测试（按编码规范「测试提交规则」）：
- 提交：仅 cmd/alonabot/t13_e2e_test.go 的 TestDevelopmentExample 一个主路径集成测试，覆盖第 5 项；理由：已有 TestDevelopmentReload 没有运行仓库中的 devbot 工程、没有验证 i18n 重建与已安装组件同时运行的连续性，这一条通过真实 CLI/Core/Adapter/插件/文件系统补齐三者。复用已准备的签名安装包与现有私有认证，无测试专用生产入口。
- 既有：TestLibraryLifecycle 如需迁移只改调用方式/路径，原断言保留；TestDevelopmentReload、TestLifecycleOverTCPAndStdio、TestTelegramDevelopment 保持原意。已有同候选 gate 结果可直接作为证据。
- 本地自测（*_local_test.go，不提交，PR 写明命令与结果）：第 6 项中未有准确历史证据的 panic/dev 来源拒绝；按文档在临时副本运行示例、做一次源码成功修改与编译失败恢复、一次仅配置变更并核对无重建；核对两种运行模式读取同一 Adapter 身份/schema。已证明的按钮/停止预算等分项引用来源，不重复造实现。SDK 与 supervisor 的本地验证遵守 D-86，不提交 sdk/ 测试、supervisor 新测试或 testdata 插件。

完成条件：改动 Go 文件 gofmt -l 无输出；go vet ./...、go test -race ./...、go build ./... 通过；定向取证命令为 go test -race -count=1 . -run '^TestLibraryLifecycle$' 与 go test -race -count=1 ./cmd/alonabot -run '^(TestDevelopmentExample|TestDevelopmentReload|TestLifecycleOverTCPAndStdio|TestTelegramDevelopment)$'，已有同候选结果时复用。现阶段 go run ./cmd/alonabot plugin check --dir examples/devbot/plugin 与 adapter 均通过。文档中的已实现命令在临时目录实际走通，TEST-MATRIX §3.2 按 §3 填写测试位置、完整命令、平台/工具版本、日志位置和结果；init 与外部后继分项按状态段的审查者裁决登记归属，不记成已通过。

停止条件：开发与安装组件的握手/注册/权限/发送出现差异；热部署重启 Core、断开已安装组件或 Adapter Session，或编译失败停止旧组件；发现 SDK/Adapter/插件/internal/core 反向 import 根包。均按 PLAN T-13 停止并给审查者提供最小复现与所属文件，由相应生产切片修复；本范围内的示例、调用或文档问题直接修正。

不做：重复开发运行器、轮询器、根包装配或管理服务；公开 Reload API；instances/ 配置目录；把 examples/plugins/echo 的 main 包作为库导入；goal-14 的 SDK 示例升级、goal-82 的 plugintest/五个 main_test.go；T-09.3 的 init/build/check CLI 实现；T-09.14 的配置写回 API/WebUI；journal/state/gc 修复、发行与预装、真实平台测试。
```

### goal-82 · T-16.2 按钮、多轮等待、schedule 与示例测试

状态：已完成（[PR #183](https://github.com/AlonaBot/AlonaBot/pull/183)，合并提交 `1f1c058`），证据见 TEST-MATRIX §3.2。合入：`plugintest.(*Harness).Click(from User, sent Sent, actionID string) Result`（点击事件带原会话类型）、`Trigger(p *sdk.Plugin, jobID string) Result`、`TriggerTask(p *sdk.Plugin, taskName, key string) Result`、`Background() []Sent`，单条 Outcome 的 `Waiting`；`scheduler.(*Scheduler).Trigger`/`TriggerTask`（暂停、重叠与未知任务返回 FAILED_PRECONDITION，一次性实例被消费并停止原定时器）；五个示例插件的 `main_test.go`。启动取消的本地自测失败源于 SDK 准备期断链的基线缺陷，由普通任务 PR #185 处理。以下为派发时的语句。

按 main 核对：

- `plugintest/plugintest.go`：`Start(t testing.TB, opts ...Option) *Harness`；`(*Harness).Message(chat Chat, from User, content message.Content) Result`、`Text(chat Chat, from User, text string) Result`、`Event(chat Chat, from User, event sdk.EventType, payload any) Result`。`Result{Sends []Sent; Outcomes []Outcome; Waiting bool}`；`Outcome{PluginID, RouteID string; Err *errcode.Error}`；`Sent{Op string; To message.Session; Chain message.Chain; Edited message.MessageRef; Result *message.SendResult}`，`(Sent).Text() string`。当前没有 Click、Trigger、TriggerTask 或 Background。
- `plugintest/options.go`：`Plugin(p *sdk.Plugin) Option`、`Config(p *sdk.Plugin, configJSON string) Option`、`Admin(userID string) Option`、`AllowedInstances(p *sdk.Plugin, instanceIDs ...string) Option`、`Capabilities(names ...string) Option`、`Adapter(id string, declarations ...any) Option`。`User.Permissions` 是 `[]sdk.Permission`；`sdk/permission.go` 的 `PermissionName(name string) Permission`、`(Permission).Or(others ...Permission) Permission` 与 `Names() []string` 表达权限名集合。缺省发信许可为全部实例；`AllowedInstances(p)` 收紧为空集合。
- `sdk/plugin.go`、`sdk/generated.go`：`NewPlugin() *Plugin`、`UseGenerated(p *Plugin, g Generated)`、`(*Plugin).Generated() Generated`；工程身份从 `Generated.Manifest` 解析。`(*Plugin).Every(jobID string, d time.Duration) *ScheduleBuilder`、`Cron(jobID, expr string) *ScheduleBuilder`；`sdk/registry.go` 的 `(*ScheduleBuilder).Handle(fn func(*Context) error)`。`sdk/continuation.go` 的 `(*Context).Continue(fn func(*Context) error) error`、`Wait(timeout time.Duration) (*Context, error)` 已可用。
- `internal/inproc/harness.go`：`Outcome{PluginID, RouteID string; ErrorCode int32; SafeMessage string}`、`Dispatch{InstanceID, SourceEventID string; Waiting bool; Outcomes []Outcome}`、`Harness{AllowedInstances map[string][]string; Dispatched func(Dispatch)}`；`RegisterHarness(dataDir string, h *Harness) (unregister func())`、`HarnessFor(dataDir string) *Harness`。`internal/core/router/observation.go` 已通过 `Invocation.OnContinuation`、`OnWait`、`OnWaitMatch` 覆盖延续调用的结束、首次等待和下一条注入的归属；`assembly.go` 已把 `router.Options.Dispatched func(router.Dispatch)` 转成 `inproc.Dispatch`。本切片只补单条 Outcome 的 Waiting，不重做这套通知。
- `internal/core/scheduler/scheduler.go`：`New(opt Options) *Scheduler`，`Options{Store *state.Store; TasksPerPlugin int; TaskPayloadBytes int64; TaskConcurrency int; JobTimeout time.Duration; Now func() time.Time}`；`(*Scheduler).Attach(ctx context.Context, p router.Plugin, activatedAt time.Time) error`、`Detach(pluginID string)`、`Released(pluginID string)`、`Close()`、`Stats(pluginID, jobID string) JobStats`；`tasks.go` 有 `TasksStats(pluginID string) TaskStats`。正常触发共用 `call(ctx, pluginID, reg, job, run, scheduledAt, now, task)`、静态运行槽及动态 `tickTask`/`pump`；当前没有立即运行入口。`assembly.go` 已把持久状态库、tune 限额与六个任务服务接入同一个 Scheduler。
- `sdk/task.go`：`Task[T any](name string) TaskType[T]`、`OnTask[T any](r Registrar, t TaskType[T]) *TaskBuilder[T]`、`(*TaskBuilder[T]).Handle(fn func(*Context, T) error)`；`(TaskType[T]).At(ctx context.Context, when time.Time, args T, opts ...TaskOption) (string, error)`，`Every(ctx context.Context, d time.Duration, args T, opts ...TaskOption) (string, error)`，`Cron(ctx context.Context, expr string, args T, opts ...TaskOption) (string, error)`，`Cancel(ctx context.Context, key string) error`，`List(ctx context.Context) ([]TaskInstance[T], error)`；`Key(key string) TaskOption`。静态暂停函数的真实名称在 `sdk/job.go`，是 `PauseJob`、`ResumeJob`、`JobStatus`。
- `sdk/buttons.go`：`Action[T any](name string) ButtonAction[T]`、`(ButtonAction[T]).Button(label string, value T) message.Button`、`OnButton[T any](r Registrar, a ButtonAction[T]) *ButtonBuilder[T]`；`sdk/message/buttons.go` 的 `Buttons(rows ...[]Button) Segment`；按钮的 `CoreToken string` 在 `sdk/internal/messagecodec/codec.go`。`internal/core/buttons/buttons.go` 的 `Resolve(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error)` 从点击载荷 `value` 读取 token，再恢复可信 action 与业务 value。`sdk/adapter/adapter.go` 的 `Core.Deliver(ctx context.Context, ev *message.Event, cursor []byte) (accepted bool, err error)` 是注入入口。
- `sdk/adapter/declaration.go`、`action.go`：`PermissionDeclaration`；`Event[T any](name string, version uint32) EventDeclaration`、`Extension[T any](name string, version uint32) ExtensionDeclaration`、`Action[P, R any](name string, version uint32, scope ActionScope) ActionDeclaration`。若本地多进程验证需要 Adapter HTTP，沿用 `sdk.AdapterHTTP.HTTP(route func(method, path string) *sdk.RouteBuilder)`（`sdk/adapter_run.go`）及 `RouteBuilder.Public(mode HTTPAuth) *RouteBuilder`、`HandleFunc(func(http.ResponseWriter, *http.Request))`，路径登记函数和 Builder 各有其职责。
- `examples/plugins/{echo,lifecycle,showcase,configdb}/main.go` 均有 `New() *sdk.Plugin`。echo 与 showcase 当前缺 `plugin.yaml`，对应 `alonabot_gen.go` 没有 Manifest；`plugintest.Start` 会在启动前拒绝。showcase 已有 `/demo remind <key> <seconds> <text>`、`/demo cancel <key>`，任务名 `remind`，route_id `task.remind`；内部 key 按 `fmt.Sprintf("%d:%s/%d:%s/%s", len(c.User()), c.User(), len(c.Session()), c.Session(), key)` 组成。仓库没有独立 `examples/plugins/media/`。
- 旧在途快照中的 #165（`330164f`）、#167（`3a9b4af`）、#168（`37b0ada`）、#169（`4dc4bcd`）、goal-59（#170，`eca6310`）、goal-69（#171，`3043c4e`）已在本地完成记录中；#169 的 `internal/core/state/schema.sql` 与 manage 清理清单、#171 的根包与 `cmd/alonabot/` 已属于基线。相关 #166（`6f72fc8`）、#173（`2bb8181`）同样已合并；启动取消修复 #174 按维护者信息已合并，合并号待审查者补齐。
- 当前在途按维护者清单：#177（goal-66）和 goal-41 与本切片共享 `assembly.go`；#177 另改 `alonabot.go`、logging、manage 和 CLI，goal-41 另改 state、journal、supervisor 及 manage。#178（goal-60）主要改 `agent/`，本地多进程验证复用根包时关注 `pipeline_test.go` 的 TestMain；goal-9 共享 `docs/development.md` 与 `spec/TEST-MATRIX.md`，并消费根包测试装配。goal-102 与本切片共享 `plugintest/{plugintest,options}.go`、probe、文档和证据行；建议本切片先合并，goal-102 按合入后的文件接续。上述在途状态未查远端 diff。

```text
/goal 使用 $alonabot-execute 推进 goal-82（切片 T-16.2：按钮、多轮等待、schedule 与示例测试）。goal 目录 goal-82，分支 goal-82，worktree .worktrees/goal-82，从最新 origin/main 创建（包含 PR #136、#153、#107、#168、#169、#176）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：补齐 plugintest 的按钮点击、等待中继续注入、静态 schedule 与动态任务实例的立即触发，以及分发链之外后台发送的顺序读取；用套件测试现有示例插件并补充开发文档，记录 V-73 的本切片证据。事件与任务都经真实 Core、SDK Session、权限、Hook、Target 准入及 Scheduler 执行。

依据：spec/PLAN.md §3.1、§5；spec/plan/core.md 的 T-16.2、T-16 步骤 3（LLM 替身归 T-16.3）、4 及「失败停止条件」；spec/CONTRACTS-v0.1.md §2.1～§2.4、§4.3、§8、§15.5、§16（JOB_TIMEOUT、动态任务限额）、§17；spec/MESSAGE-MODEL.md §1～§4、§5、§6.3；spec/WIRE-FORMAT.md §3.3（schedule/task 注册）、§3.5（任务与等待固定服务）、§4；spec/adr/DECISIONS.md D-47、D-68～D-72、D-77、D-83、D-86、D-102、D-105、D-107～D-108、D-119～D-120、D-124～D-127、D-134、D-137；spec/TEST-MATRIX.md V-73、§3（动态任务沿用 V-114 的既有语义）。

范围（只新增或修改这些）：
- plugintest/plugintest.go、adapter.go 及同目录新增的按钮、调度、后台记录实现文件；plugintest/options.go 仅在共用套件状态接入所需时修改。
- plugintest/internal/probe/：在已有插件内追加按钮动作、静态任务、动态任务和受测试同步控制的后台发送场景；修改相应源资源后重新生成 alonabot_gen.go。
- internal/inproc/harness.go：Outcome.Waiting 与调度桥接字段；字段在组件启动前装配完成，运行期间保持现有只读约定。
- internal/core/router/observation.go：等待通知附带进入等待的路由 Outcome；保持已存在的延续调用跟踪与匹配通知。
- internal/core/scheduler/scheduler.go、tasks.go 及同目录新增 trigger.go：立即运行入口，复用正常调度的派发、队列、运行槽、期限和完成处理。
- assembly.go：仅给该数据目录的 inproc.Harness 安装调度桥接，并传递 Outcome.Waiting。
- examples/plugins/showcase/plugin.yaml（以及派发时仍缺少 plugin.yaml 的其他示例；goal-9 合并后 echo 已有）：补齐可供 Start 解析的最小工程身份；对应 alonabot_gen.go 只由现有生成器重建。各示例 main.go 的业务行为保持原有含义。
- plugintest/plugintest_test.go；examples/plugins/{echo,lifecycle,showcase,configdb}/main_test.go：以下列出的集成测试。
- 上述包内被忽略的 *_local_test.go，以及根包的 *_local_test.go：只用于本地验证；多进程验证复用 pipeline_test.go 的 newPipeline/TestMain 装配，不新增 testdata 插件。
- docs/development.md：仅新增「插件测试」一节；spec/TEST-MATRIX.md §3.2：仅追加本切片证据行。

允许的新跨模块边：无。复用 ARCHITECTURE 已允许的 plugintest -> 根包、sdk/...、internal/inproc、contracts，以及 examples/plugins/... -> plugintest（仅测试）。internal/inproc 的字段仅使用标准库与本包类型；Core、SDK、根包产品代码均沿用既有依赖方向。

接口（照此实现；以下标明新增，其余保持 main 的真实签名；内部辅助函数自定）：
1. plugintest 新增：
   func (h *Harness) Click(from User, sent Sent, actionID string) Result
   func (h *Harness) Trigger(p *sdk.Plugin, jobID string) Result
   func (h *Harness) TriggerTask(p *sdk.Plugin, taskName, key string) Result
   func (h *Harness) Background() []Sent
   type Outcome struct { PluginID, RouteID string; Err *errcode.Error; Waiting bool }
   Result 的 Sends、Outcomes、Waiting 及 Message/Text/Event 的参数、返回值保持 main 的定义。
   TriggerTask 用任务名和完整实例 key 定位已经创建的实例，参数取 Core 已保存的数据。
2. internal/inproc 新增字段：
   Outcome.Waiting bool
   Dispatch.RunID string
   Harness.Trigger func(ctx context.Context, pluginID, jobID string, completed func(Dispatch)) error
   Harness.TriggerTask func(ctx context.Context, pluginID, taskName, key string, completed func(Dispatch)) error
   completed 只服务于这次显式触发：在本次 RPC 结果确定时调用一次；Dispatch 带 RunID 和路由 Outcome，InstanceID、SourceEventID 为空。函数返回准入/定位错误；已派发 Handler 的错误放入 Outcome。
3. internal/core/router：Outcome 增加 Waiting bool；Dispatch.Waiting 保持现有事件级语义。
4. internal/core/scheduler 新增内部跨包入口：
   func (s *Scheduler) Trigger(ctx context.Context, pluginID, jobID string, completed func(runID string, response wire.Response, err error)) error
   func (s *Scheduler) TriggerTask(ctx context.Context, pluginID, taskName, key string, completed func(runID string, response wire.Response, err error)) error
   根包将这两个入口适配成 Harness 字段；completed 在锁外调用，并在本次触发方法返回前完成。普通 timer 与任务服务继续复用同一派发及收尾代码。结果超时/取消后，实际运行槽仍按既有 Late/Released 规则释放。
5. 使用现有 wire 字段，字段号保持原值：
   Button.core_token=4（optional string）；ButtonClickedPayload { action_id=1; action_schema_version=2; value=3; callback_query_id=4; optional origin_platform_message_id=5; }。
   ScheduledPayload { job_id=1; run_id=2; scheduled_at_unix_ms=3; task_name=4; task_key=5; task_payload_json=6; }。
   动态请求目标与 job_id 均为 task.<任务名>，三个 task 字段来自保存的实例；静态任务的 task 字段为空。

行为：
- Click 在 sent.Chain 的 buttons 行中按 actionID 找到回调按钮，读取 CoreToken；不存在或找到的是 URL 按钮等测试输入错误时 t.Fatal。交付的 button.clicked 使用 sent.To、其 SendResult 的实例与原消息 ID，Actor 取本次 from，CallbackQueryID 为本次唯一值，Value 只放 token。通过测试 Adapter 的 Core.Deliver 进入原有 buttons.Resolve 与 Router；Core 恢复 action、版本和值并逐次鉴权。过期、代次失效或权限拒绝时返回无发送、无实际 Handler Outcome 的结果。
- 等待继续复用 main 的普通分发链和延续调用计数：某路由进入 core.wait 时，快照中该 Outcome.Waiting=true，Result.Waiting=true，注入立即返回；下一条同键消息命中等待后，后续发送归属本次新注入，直到延续调用结束或再次等待。其他用户或会话继续普通分发。等待中的路由最终返回的错误由后续对应注入取得。
- Trigger 只定位 p 对应的已激活插件及静态 job_id，立即运行一轮，复用 Scheduler 的调用身份、可信 job scope、运行槽、JOB_TIMEOUT 和错误处理。未知任务、未激活、暂停或同 job 已运行时返回有原因的错误，由套件报告 t.Fatal；自然 tick 仍按原有跳过重叠规则处理。
- TriggerTask 定位 p 下已经存在的 taskName/key，使用完整保存参数立即安排这一实例；经过同一实例运行槽、TASK_CONCURRENCY 与既有到期队列。同实例已运行或已排队时返回 FAILED_PRECONDITION；静态任务已暂停时 Trigger 返回 FAILED_PRECONDITION；全插件名额已满时经既有队列等待。一次性实例在实际派发时按现有规则删除；周期实例保留定义与持久起点，后续自然触发仍取未来时刻。测试触发时刻为本次立即触发的时刻，动态期限从实际派发时计算。替换、取消、缺失任务、存储错误均保留既有语义；取消实例后再次触发报告定位失败。
- Trigger/TriggerTask 在 completed 通知处冻结发送记录终点、收集 run_id 与 Outcome，之后返回 Result；记录接收者先登记，再发起运行，避免快速完成丢通知。Handler 错误与 DEADLINE_EXCEEDED 等进入 Outcome.Err；RPC 结果与实际运行结束的区分沿用 Scheduler。显式触发仍用 scheduled REQUEST/RESPONSE，所需消息目标由 Handler 显式 sdk.To 提供。
- Background 返回并清空尚未交给任何注入结果的发送/编辑记录，按测试 Adapter 接受的顺序；已经由 Result.Sends 交付的记录各消费一次。沿用 Harness 的串行注入约定和现有锁，在分发/调度完成通知处固定区间。真正位于这些区间之外的 OnReady 后台发送保留到 Background；读取时尚未发生的发送由之后的读取取得。
- 发信许可遵循 D-83：缺省全部实例可发；AllowedInstances(p) 收紧为空后，job 与后台显式发送返回 PERMISSION_DENIED。拒绝的发送保留错误而无 Sent；原可信入站目标的回复仍按现有豁免处理。
- 示例用现有 New 创建插件。echo 验证文本回复及混合片段中的文本提取；lifecycle 验证 /ping；showcase 验证 /demo hello、参数解析，以及创建提醒后通过完整内部 key 的 TriggerTask 得到原会话提醒、取消后查询消失；configdb 验证 Config、/note 命令与真实 dev.sqlite。media 示例（goal-9 新增）验证上传图片组成相册、相册后的文本承载按钮，以及去掉 message.album 能力后被拒。
- 文档给出 Start/Plugin、Text 与 Sent 断言的最小例子，再展示权限、Config/AllowedInstances/Capabilities、Click、Trigger、TriggerTask、等待与 Background；写明后台发送用测试自己的 channel 同步，套件靠 Core 完成通知返回。运行命令包含 -race -count=2。

测试（按编码规范「测试提交规则」，D-86）：
- 提交：扩展 plugintest/plugintest_test.go 的一个 TestPluginHarness，使用既有 probe 与真实 Core/Session/SQLite，按子测试覆盖：
  1) /menu 发两个动作按钮，点击获得对应回复；同群另一用户点同按钮以点击者身份执行；要求 sdk.Admin 的动作拒绝普通用户并允许 Admin。
  2) /ask 先发送问题再进入等待，返回 Waiting Outcome；另一用户消息正常分发，同一用户回答归入新注入；保留既有多轮等待测试断言。
  3) 静态 Trigger、动态 TriggerTask 各返回真实发送与路由 Outcome；动态以任务创建命令经 SDK 写入 Core，触发后核对一次性实例被删除。缺省发送允许；空 AllowedInstances 的另一套 Harness 中显式发送为 PERMISSION_DENIED。
  4) OnReady 只登记后台 goroutine，使用 channel 在 Start 返回后控制一次真实后台发送并等待该 Send 返回；Background 依次读到该发送、第二次为空，已归入 Text/Trigger 的发送不会重复读取。
  5) 保留命令、平台权限、Core Admin、Capability 降级的既有集成断言；补一条 Rule 主路径与一条媒体/能力拒绝路径，守住 V-73 所需组合。
- 提交：examples/plugins/{echo,lifecycle,showcase,configdb,media}/main_test.go 各一个主路径集成测试，均直接调用本包 New 与 plugintest.Start。showcase 的提醒通过实际创建命令和 TriggerTask 交付；configdb 通过 Config 写入问候词并执行笔记写入/读取。参数失败路径选 showcase，其他边界放本地。多列理由：T-16 步骤 4 要求交付示例测试；四个现有 package main 各自构建，测试分别运行真实组件与各自资源，套件内部 probe 测试无法替代这些工程的生成资源与入口验收。
- 本地自测（*_local_test.go，PR 写明实际命令、平台与结果）：按钮不存在、过期和代次失效；再次等待、超时与错误归属；发送与完成通知的竞态；同静态 job 重叠、暂停与未知 job；动态实例替换/取消、并发队列、存储失败与超时后的 Late 槽释放；后台 Edit 与顺序消费；收紧发信许可后原目标回复豁免；每份套件临时目录与配置/SQLite 隔离；Start 失败、启动 ctx 取消及清理后锁和组件释放。需要多进程对照时在根包 *_local_test.go 复用 pipeline_test.go 的装配与测试二进制分流，分别记录 TCP/stdio；sdk/、supervisor/、testdata 不新增提交测试。
- 证据：TEST-MATRIX §3.2 追加 T-16.2/V-73 一行，写测试位置、实际命令、平台、Go 版本、结果和真实未运行项；LLM 分项由 goal-102 追加，保留已有 T-16.1 证据。

完成条件：改动 Go 文件 gofmt -l 无输出；按现有入口生成 probe 与补齐身份的示例资源，go generate ./... 后重复生成内容一致；go vet ./...；go test -race -count=2 ./plugintest/... ./examples/plugins/...；go test -race ./internal/core/... ./internal/inproc/... .；go build ./...。核对 T-16 的三项失败停止条件：所有 Core 检查与正常派发仍生效，注入返回由完成通知决定，生产依赖方向符合 ARCHITECTURE；任一失败先修复。spec/TEST-MATRIX.md 证据与本地自测结果完整后结束。

不做：LLM 替身（goal-102）；模拟时钟、缩短正式周期、另建派发器或 wire 服务；动态任务 SDK、schema.sql 与卸载清理的返工；Adapter 作者测试框架；日志管理、Agent 工具循环、开发 CLI、Journal 激活流程；新增 SDK 公开符号或实现 ServiceBuilder.Permission；真实 Telegram 与 WebUI 沙盒；范围外 spec、PLAN、ADR 与 ARCHITECTURE 修改。
```

## 第 15 波

### goal-38 · T-09.3 开发 CLI 与库模式脚手架

状态：已完成（[PR #194](https://github.com/AlonaBot/AlonaBot/pull/194)，合并提交 `e9ff45e`），证据见 TEST-MATRIX §3.2。合入：顶层 `alonabot init`（四角色模板，`--module`）、`check`、`build`、`test`（离线 Core 侧会话）、`pack`（`--asset-url` 必填、`--key` Ed25519 签名、包根 build-info、归档外 release.json）；`internal/pluginkit` 的 `Check`、`Pack(ctx, options, assetURL, keyFile, outDir)` 与正式构建（`CGO_ENABLED=0`、资源快照、`manifest.Generate` 合成包内 manifest）；`sdk/generate` 的 `Result.HTTP` 与 `Options{Resources, CGOEnabled, RequireAdapterConstructor}`；仓库生成指令迁移到 `alonabot build`；echo、media 的顶层 check/build 与 devbot 的 init 模板比对补证。审查者补齐的约定见 DECISIONS 2026-11-13 变更记录。以下为派发时的语句。

按 main 核对：

- **生成器与入口。** `sdk/generate/generate.go`：`Scan(ctx context.Context, opt Options) (*Result, error)`；`Options{Dir string, GOOS, GOARCH string, Tags []string}`，目标为空时沿用 Go 环境；`Result{PluginConstructor string, AdapterType string, ConfigSchema []byte, Source []byte, Diagnostics []Diagnostic}`；`Diagnostic{Pos, Message string}`；`FileName = "alonabot_gen.go"`。`Scan` 排除旧生成文件，汇总诊断，成功时返回格式化源码；只有实例 schema、没有插件构造函数的旧检查路径仍存在。`sdk/generate/project.go` 按工程 manifest 的 adapter 角色查找入口包中唯一具体 Adapter 类型；`sdk/generate/source.go` 为插件生成包内变量 `alonabotGenerated`，为 Adapter 生成 `UseGeneratedAdapter` 登记。该生成器已有 schema/迁移、i18n、Service Schema、Handler 参数名与文案常量检查。
- **现有共享构建。** `internal/pluginkit/build.go`：`BuildDev(ctx context.Context, dir, outDir string) (build Build, err error)`；`Build{Executable string, Manifest []byte, PackageManifest []byte, Digest []byte, Role, ID, AdapterID string}`。它按宿主平台扫描，通过 overlay 注入资源与薄入口，使用临时 modfile，成功后发布可执行文件及其摘要命名的资源旁车；工程目录保持只读。可导入插件/Agent 调 `sdk.Run(component.New())`，Adapter 调 `sdk.RunAdapter(component.New())`；main 包已有 main 时沿用，否则注入入口。其私有 `migrationMode(ctx context.Context, dir, overlay, modFile string) (string, string, error)` 已识别 Migrate/Migrations/Open 的引用与互斥。当前尚未提供带 GOOS/GOARCH/tags 的正式构建、归档与签名接口。
- **工程与包 manifest。** `contracts/manifest/project.go`：`ParseProject(data []byte) (*Result, error)`；`contracts/manifest/manifest.go`：`Parse(data []byte) (*Result, error)`，`Result{Manifest *Manifest, Warnings []string}`。`contracts/manifest/build.go`：`Generate(project *Manifest, build BuildInfo) ([]byte, error)`；`BuildInfo{ManifestVersion int, ProtocolMinMajor, ProtocolMaxMajor *uint32, RequiredCapabilities *[]Capability, OS, Arch, Entrypoint string, ConfigVersion *uint32, MigrationMode, PrepareEffects string}`。Generate 保留作者字段和 external，按角色输出并再次 Parse。工程 manifest 显式写构建字段（含零值/null）会被拒绝；包 manifest 缺 migration_mode/prepare_effects 的遗留警告行为已在 goal-100 与 TEST-MATRIX 记为待裁决。
- **配置与翻译。** `contracts/configschema/schema.go`：`Parse(raw []byte) (*Schema, []Problem)`、`(*Schema).Canonical() []byte`、`Version() int`；`contracts/configschema/i18n.go`：`(*Schema).CheckI18n(config json.RawMessage) []Problem`。`contracts/i18n/catalog.go`：`Parse(files map[string][]byte) (*Catalog, []Problem)`；`sdk/generated.go` 的 `Generated` 含 Manifest、ConfigSchema、AdapterConfigSchema、ServiceSchemas、I18n 与 ParamNames。复用这些检查，作者不再声明 Config 类型或 sdk.Base。
- **SDK 与离线会话。** `sdk/plugin.go`：`NewPlugin() *Plugin`；`sdk/generated.go`：`UseGenerated(p *Plugin, g Generated)`、`(*Plugin).Generated() Generated`、`UseGeneratedAdapter(a adapter.Adapter, g Generated)`；`sdk/sdk.go`：`Run(root *Plugin) error`、`Collect(root *Plugin) (*Registration, error)`；`sdk/adapter_run.go`：`RunAdapter(a adapter.Adapter) error`。`internal/core/supervisor/supervisor.go`：`Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error)`、`(*Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult`；`Running.Instance` 是 `*host.Instance`，`internal/core/host/activate.go` 的 `Instance.Routes []*pb.RouteDescriptor` 为已校验路由。CLI test 可复用现有 Core 侧会话、状态与停止原语取实际注册结果。
- **release 与预检。** `contracts/release/release.go`：`Verify(raw, sigFile []byte, pub ed25519.PublicKey) (*Release, error)`、`KeyID(pub ed25519.PublicKey) string`、`DecodeSignature(sigFile []byte) ([]byte, error)`、`DecodePublicKey(raw []byte) (ed25519.PublicKey, error)`；`Release{PublisherKeyID, Source, PluginID, Version string, Assets []Asset}`，`Asset{OS, Arch, URL string, Size int64, SHA256 string}`，资产 URL 必须是绝对 https URL。`internal/core/packages/preflight.go`：`Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error)`、`(*Candidate).MoveTo(packagesRoot string) (string, error)`、`Discard() error`；`Input{ReleaseJSON, Signature []byte, PublicKey ed25519.PublicKey, Source, OS, Arch, ArchivePath string}`；`NewLimits(v tune.Values) Limits`。预检已完成验签、归档长度/摘要/来源/平台、受控解压与入口文件检查。
- **命令树与改名。** `cmd/alonabot/commands.go` 的 `execute(args []string) int` 注册 root 默认 run、`run [directories...]`、`tune list`、`plugin generate/check`、`admin-token init/reset/show`、管理命令与日志命令。管理命令为 install/list/enable/disable/restart/reload/update/rollback/uninstall/pin、adapter instance create/start/stop/delete、admin add/remove/list，日志为 logs/log-level。`plugin.go` 的 `pluginCommand(code *int) *cobra.Command` 与 `pluginGeneration(cmd *cobra.Command, options generate.Options, check bool) error` 共用扫描，check 逐字节比较生成代码；提示仍要求 `alonabot plugin generate`。本切片替换 plugin 命令组为顶层 init/build/check/test/pack，删除旧 generate/check 命令入口与别名；run 已由 #171 实现。`cmd/alonabot/dev.go` 的 `runDevelopment(dirs []string, dataDir string, reload, stdio bool, interval time.Duration, options []alonabot.Option) error` 已调用 BuildDev 与根包 Start。
- **库模式与示例补证。** `alonabot.go`：`Start(ctx context.Context, opts ...Option) (*Core, error)`、`Run(opts ...Option) error`、`(*Core).Wait() error`、`Plugin(root *sdk.Plugin) Option`、`Adapter(a adapter.Adapter) Option`、`Executable(path string, packageManifest []byte) Option`、`DataDir(dir string) Option`。`examples/devbot/plugin/plugin.go` 是可导入包，New 显式绑定资源，含 /hello 命令；相邻 plugin.yaml 只含作者字段，另有 config_schema.json、i18n/en-US.json、忽略 data/ 的 .gitignore。examples/devbot 根目录另有库模式 main.go、回环 Adapter 与 core_config.example.json。init 模板比对针对组件工程形状；库模式宿主由既有示例展示。goal-14 状态末尾、goal-70 状态及 TEST-MATRIX §3.2、DECISIONS 2026-11-13 变更记录共同把 echo/media 的顶层 check/build 与 init 模板比对交给本切片。
- **生成调用链与在途冲突。** `rg` 已定位 sdk/generate/source.go 的文件头、sdk/services.go 的缺 Schema 提示，以及 agent、Telegram、五个插件示例、devbot 两个组件与 plugintest/internal/probe 的 go:generate 指令仍使用旧命令。PR #191 持有 examples/plugins/echo、media、sdk/services.go、sdk/bus.go 和 cmd/alonabot/t04_e2e_test.go，并包含 main 包 schema_id 的 go test 回退；这些实现在本基线尚未合并。PR #188/#189 持有 agent/，#190 持有 plugintest/；本切片在这些目录仅迁移命令指令及重新生成资源。PR #187 持有 journal/state/gc、alonabot.go/assembly.go，#192 持有 manage/options、根包装配及 cmd/alonabot/manage.go 的 reload；本切片复用其接口，生产改动集中于自己的命令文件与 pluginkit。commands.go 注册行、spec/TEST-MATRIX.md 追加区亦需后合并者协调。PR 在途范围按本次任务提供的信息记录，未查询远端状态。

签名与字段均经本 worktree 的 `rg`/源码核对；BuildDev、BuildInfo、generate.Result 另用本机 Go 1.27.3 的 `go doc` 复核，其余主要入口也已读取 go doc 输出。Go 文档命令伴随未缓存模块的查询警告，文档结果与源码逐项核对；未运行编译或产品测试。

```text
/goal 使用 $alonabot-execute 推进 goal-38（切片 T-09.3：开发 CLI 与库模式脚手架）。goal 目录 goal-38，分支 goal-38，worktree .worktrees/goal-38，从最新 origin/main 创建。须包含直接前置 PR #102（f685c92）、#103（c840af7）、#140（896991b）、#171（3043c4e），以及审查者定稿时补记的 T-04.5 / PR #191 实际合并提交；沿用 #186（c38a294）的 devbot 工程及补证约定。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。发布 URL、密钥编码、build-info 与外部 module 模板的接口细节须由审查者在派发前按本稿备注写定。

目标：完成四种组件通用的顶层 alonabot init/build/check/test/pack。init 在当前或给定空目录创建作者工程，不生成 main.go；build 生成资源与正式薄入口并编译独立组件，产物不包含 Core；check 只读核对；test 在临时数据与离线会话验证加载、准备、注册和激活；pack 复用构建生成可安装 tar.gz、归档外 release.json 及可选签名。复用现有库模式与开发运行入口，补齐 goal-14、goal-70 留给本切片的顶层命令和模板证据。

依据：spec/PLAN.md §3.1 的 T-09.3；spec/plan/core.md 切片表 T-09.3、T-09 步骤 2 中的开发 CLI 部分及「对应 V 验收与停止条件」；T-09.16、T-13.3、T-13.4 的既有实现与证据用于界定复用。spec/CONTRACTS-v0.1.md §5.1/§5.3～§5.5、§6.2、§7.5、§8、§9.1～§9.3、§15.4/§15.5、§16；实际握手、注册、激活及停止沿用 §1.1/§4.1/§4.2 与 WIRE-FORMAT §3.1～§3.4。DECISIONS D-42/D-43、D-59、D-62/D-63 取其仍有效部分，并按 D-76～D-81、D-82、D-85/D-86、D-106、D-114、D-117、D-134、D-136/D-137、D-139 的现行修订执行；另读 2026-11-13 关于 T-04.5/T-13.4 向本切片移交补证的变更记录。TEST-MATRIX §1/§2 的 V-01、V-41、V-54 打包分项及 §3/§3.2；V-10、V-64/V-65 只补下述命令与模板分项。模块边按 ARCHITECTURE.md，测试按 docs/coding-standards.md「测试提交规则」。

范围（只新增或修改这些）：
- cmd/alonabot/commands.go：把 pluginCommand 的注册替换为五个顶层开发命令，保留其他命令注册；cmd/alonabot/plugin.go：迁移原扫描、诊断与只读生成比较，移除旧命令组；按职责新增 init.go、build.go、check.go、test.go、pack.go 中实际需要的文件。
- internal/pluginkit/build.go 及该目录必要的普通 Go 文件：复用并整理现有构建链，加入正式目标平台/tags、生成资源发布、脚手架与打包/清单/可选签名所需逻辑；保留 BuildDev 的签名与开发调用契约。薄入口继续由现有构建所有者生成，无须另建 SDK 入口生成器。
- cmd/alonabot/dev.go：仅在共享构建接口接线确有需要时调整调用；保留目录发现、轮询、重启与错误处理。
- sdk/generate/source.go：把生成文件头改为 alonabot build；sdk/services.go：只把缺 Schema 的建议命令改为 alonabot build，保留 #191 合入后的 schema_id 回退。生成器的扫描与 SDK 注册语义沿用前置。
- 下列源文件仅迁移 go:generate 的旧命令为 build（输出放到被忽略的 bin/ 或临时目录，避免工程目录混入二进制）：agent/agent.go；adapters/telegram/telegram.go；examples/plugins/{echo,media,showcase,configdb,lifecycle}/main.go；examples/devbot/plugin/plugin.go；examples/devbot/adapter/adapter.go；plugintest/internal/probe/probe.go。
- 以上十个工程目录的 alonabot_gen.go：只由更新后的生成命令重新生成；同时保留各工程按其既有 tags 选择的生成方式。echo/media 的作者逻辑与 #191 新增测试、agent/ 与 plugintest/ 的在途功能由各自切片持有。
- 新增 cmd/alonabot/plugin_cli_e2e_test.go（//go:build unix）：仅下述一个提交集成测试及必要内部辅助；复用该包 TestMain 和既有 CLI/临时目录辅助。本地自测写 cmd/alonabot/、internal/pluginkit/、sdk/generate/ 下的 *_local_test.go，临时工程全部写 t.TempDir()，不新增 testdata 测试插件。
- docs/development.md 的开发命令与 init→build→check→test→pack 用法、docs/plugin-guide.md 中生成/检查/打包命令与可选签名说明、examples/devbot/README.md 的命令及模板比对说明。只修改受命令迁移影响的内容。
- spec/TEST-MATRIX.md §3.2：追加 T-09.3 的实际证据及两个切片移交的补证，引用已有证据与剩余归属。README.md、ARCHITECTURE.md、docs/goals/、docs/execution-order.md 与规格的其他同步交给审查者，在 PR「文档同步」中逐项列出。
允许的新跨模块边：无。cmd/alonabot -> internal/pluginkit、sdk/generate、contracts、internal/core/...、根包，以及 internal/pluginkit -> contracts/...、sdk/generate 已由 ARCHITECTURE 登记。组件可执行文件仅链接公共 SDK；不为离线 test 新增 SDK 到 Core、pluginkit 到 Core/根包或根包到 plugintest 的边。

接口（照此实现）：
1) alonabot init [目录] --id ID [--role plugin|adapter|agent|webui] [--adapter-id ID] [--module PATH]。省略目录时为当前目录；role 缺省 plugin，目录非空时拒绝且保留原内容。工程 plugin.yaml 只含作者字段：id、name、version（初始 0.1.0）、source，及角色必需字段；source 的示例值明确提示作者修改。插件与 Agent 为可导入包，导出 func New() *sdk.Plugin，调用 sdk.NewPlugin 与 sdk.UseGenerated(p, alonabotGenerated)，显式登记一个带英文说明的示例命令；生成 config_schema.json、i18n/en-US.json 与忽略 data/ 的 .gitignore。Adapter 要求 --adapter-id，以可导入包提供 New() adapter.Adapter、唯一具体 Adapter 类型和 adapter_config_schema.json，遵守 D-134/D-136；WebUI 生成作者 manifest、assets/index.html 与 Admin API 主版本范围，使用静态工程形状。所有角色均生成忽略 data/ 的 .gitignore，均不生成 main.go；库模式宿主参考 examples/devbot/main.go。仓库内工程复用根 module，外部工程沿用组件自身 Go Modules；目标目录已位于某个 Go module 内时复用该 module、不写 go.mod；否则必须给出 --module PATH 并写 go.mod，`require github.com/AlonaBot/AlonaBot` 的版本取 CLI 自身的模块版本（与根包 version.go 的规则一致），为 (devel) 时不写 require，并以英文提示执行 `go get github.com/AlonaBot/AlonaBot@<version>`；集成测试在临时工程中用 replace 指向本 checkout。
2) alonabot check [--dir DIR] [--goos OS] [--goarch ARCH] [--tags a,b]。dir 缺省当前目录；目标缺省沿用 Go 环境，与 generate.Options 一致。用 manifest.ParseProject 与 generate.Scan 复用检查，并比较 alonabot_gen.go 是否最新；缺失或过期时失败，提示 alonabot build。检查工程作者字段、入口/角色、config_schema.json、adapter_config_schema.json、声明式迁移、i18n 及已有 Service/参数名/文案静态规则；保留全部诊断及位置，成功或失败均不写工程、生成文件、输出包或 data/。WebUI 直接检查 manifest、静态根及 index.html/icon 等包资源，跳过 Go 扫描和生成文件比较。各角色的检查按其真实资源执行。
3) alonabot build [--dir DIR] [--goos OS] [--goarch ARCH] [--tags a,b] [--out DIR]。dir/目标/tags 与 check 一致，out 的确定缺省值由实现写入帮助和文档。先执行共同静态检查，使用新扫描结果生成资源；已有生成文件过期由本命令更新，其他静态错误阻止发布。用已锁定 Go Modules、CGO_ENABLED=0 与 -trimpath 按目标编译；薄入口放临时目录或 overlay，插件/Agent 调 sdk.Run(New())，可导入包经 component.New() 调用，Adapter 调 sdk.RunAdapter(component.New())。main 包示例已有入口时继续使用；无入口时生成薄入口。正式二进制不链接根包 alonabot 或 Core。构建成功后发布可执行文件与由 manifest.Generate 合成的包内 manifest；工程 plugin.yaml 保持作者内容。正式入口不落为作者 main.go。WebUI 沿用静态检查路径，无 Go 二进制；pack 直接取静态资源。
4) 保持 pluginkit.BuildDev(ctx context.Context, dir, outDir string) (Build, error) 与 Build 字段的现有含义。共享生成与编译逻辑时，run 的宿主平台构建、工程只读、资源旁车、旧成功产物保留及编译器完整错误继续成立。正式 build 与 run 复用相同的角色、迁移、配置版本及资源选择，目标参数在扫描和编译时一致；不要为了新增顶层命令让 run 要求工程内生成文件已经最新。internal/pluginkit 的新增内部辅助与命令装配函数由执行者按现有结构确定，不新增框架公开 API。
5) 包内 manifest 统一调用 manifest.Generate(project.Manifest, BuildInfo{...})。进程角色填 manifest_version=1、链接 SDK 的协议范围与 required_capabilities、所选 os/arch、包内相对 entrypoint、config_version（schema.$version，无文件时 1）；插件/Agent 的 migration_mode 按实际登记取 none/self/versioned，prepare_effects 按受管 SQLite 使用取 none/managed_db，保留作者 external。Adapter 按其角色排除数据库与迁移字段。WebUI 使用 any/any，只合成其角色允许的字段；包内不写 entrypoint、协议范围、required_capabilities 与 config_version。SDK 能力来源复用前置与实际链接信息，当前 BuildDev 的空能力列表只用于 SDK 确实没有必需能力的构建。
6) alonabot test [--dir DIR]。构建宿主平台，在新临时数据目录建立真实离线 Core 侧会话，完成握手→配置准备/OnStart→REGISTER 校验→Activate/OnReady→Stop；从已校验的注册结果报告路由，任一步与停止失败均保留完整错误并非零退出。复用 supervisor/host/wire/state 的现有入口与 tune 预算，数据与配置均来自临时测试环境；不读取工作目录 data/、生产凭证或启动真实平台实例。Adapter 验证组件注册/描述，使用离线会话；WebUI 验证静态加载所需资源。此命令按 CONTRACTS §9.3 检查加载与注册，保留既有 plugintest 与 go test 的各自入口。
7) alonabot pack [--dir DIR] [--goos OS] [--goarch ARCH] [--tags a,b] --asset-url URL [--key FILE] [--out DIR]。与 build 使用同一目标与资源检查；合法但陈旧的生成文件由共同 build 更新，schema/i18n/manifest 等问题导致失败。进程角色先正式构建再写 <id>-<version>-<os>-<arch>.tar.gz，Windows 同格式；包根包含包内 plugin.yaml、entrypoint、存在时的 config_schema.json、adapter_config_schema.json、i18n/、build-info 与声明的 icon。WebUI 直接写 <id>-<version>-any.tar.gz，只含包内 plugin.yaml、icon 与 assets 静态树，跳过 Go 编译。归档外写 release.json，字段使用 contracts/release.Release/Asset 的现有 JSON 名，绑定来源、ID/版本、目标、资产 URL、实际长度及整个归档摘要。资产 URL 取必填的 --asset-url（绝对 https，一次 pack 对应一个 os/arch）。build-info 为包根的 `build-info` 文件，内容是 `debug/buildinfo.ReadFile(可执行文件)` 的 `(*debug.BuildInfo).String()` 文本（可被 `debug.ParseBuildInfo` 解析），其中 GOOS/GOARCH 与包内 manifest 的 os/arch 一致；WebUI 包没有 build-info。packages.Preflight 对 build-info 的严格核对不在本切片范围。
8) --key 可选。提供时用标准库 Ed25519 对已写出的 release.json 确切 UTF-8 字节签名，release.json.sig 为标准 base64 的 64 字节签名，另写 release.pub（标准 base64 的 32 字节公钥），publisher_key_id 取 release.KeyID；key 文件为去掉首尾空白的标准 base64，内容是 64 字节的 `ed25519.PrivateKey`，长度不符完整报错。未给 key 时输出归档与 unsigned release.json，不生成伪签名或伪公钥。密钥读取/签名失败完整报错，不切换到 unsigned 成功路径。归档不包含 release.json、sig、公钥身份材料、私钥、运行配置、data/ 或 DB/secret；release 及公钥文件留在归档外。摘要只用于契约要求的归档完整性与公钥标识。

行为与验收矩阵：
1) 命令树（D-77）：帮助列出五个顶层命令；旧 plugin generate/check 入口与建议文本完成迁移。核对命令注册修改后 run、默认运行、管理与日志命令仍沿用现有入口。仓库 go:generate 指令及其生成文件头统一使用 build，能实际重新生成；命令输出和错误为英文。
2) init 与模板补证（V-01、V-64 的入口分项）：在临时空目录省略目录参数执行 init，核对作者 manifest、可导入 New、一个命令、schema、i18n 与忽略 data/ 的 .gitignore，确认未生成 main.go。与 #186 的 examples/devbot/plugin 比较工程形状及资源绑定方式；用 ParseProject、configschema、i18n 和实际构建核对有效性。devbot 的启动计数 SQL、回声、回环 Adapter、库模式宿主与配置示例属于演示内容，模板只实现契约规定的最小命令。给定空目录同样可用；非空目录拒绝并保留原文件。Adapter/Agent/WebUI 逐角色核对实际形状。
3) check 只读与构建更新（V-01、V-54）：对合法工程先 build 再 check 成功；改动资源使生成代码过期后 check 失败且目录内容/mtime 保持，build 更新后 check 成功。工程手写构建字段（含零值/null）、错误 New、迁移不连续/非法目标、缺少 en-US、占位符不符、已有静态文案诊断都失败且带原位置。WebUI 缺 index.html、资源路径非法或角色字段不符时失败。复用前置规则与诊断，不恢复标记扫描和 Config 类型推导。
4) 正式构建（V-01、V-10 的入口分项）：同一目标贯穿 Scan、go list、编译、文件名与包内 manifest；Windows entrypoint 带 .exe，生成薄入口不污染作者源码。schema 的规范内容与版本、Adapter 实例 schema、i18n、Service Schema 与实际 binary 同源。以依赖清单核对独立组件未链接根包/Core。BuildDev 经共享逻辑调整后，用原有 TestDevelopmentReload 的场景守住编译失败保留旧组件与 config-only 不重建。
5) goal-14 移交补证：必须在包含 #191 的基线上分别执行顶层 check/build --dir examples/plugins/echo 与 examples/plugins/media，实际生成/编译成功，并记录完整命令、目标和产物；检查可导入包与既有 main 包形状均被支持。go test 下 Service/Bus schema_id 与生成器一致的回归引用 #191 的证据，保留其修复；本切片不再次实现 SDK 回退。已有 TestSDKExamples（#191 合入后）与 TestTelegramDevelopment 的命令/收发断言继续有效。
6) 离线 test（CONTRACTS §9.3）：实际运行新建插件到 Activated，报告其示例命令，再停止进程并释放锁；准备/注册错误与停止错误均使命令失败。Adapter 无真实平台实例启动；临时配置与数据隔离，测试结束后有明确清理结果。WebUI 静态路径与普通进程路径分别取证。阶段预算、进程树、锁与停止失败引用现有 supervisor/host 原语，失败时完整传播。
7) 四角色 pack（V-01、V-41）：插件、Adapter、Agent 使用同一构建规则，Adapter 工程取可导入包目录并把实例 schema 放包根；WebUI 为 any/any、只含静态内容。每种角色的合法归档经 packages.Preflight 成功，主路径插件再经 MoveTo 成功。新进程包明确写足 migration_mode/prepare_effects 等角色必需元数据，保留 external；不以现有 parser 的警告行为省略构建字段。包内资源与嵌入资源逐项核对。
8) 可重复与签名（V-41）：同源码、目标、tags、工具链与作者元数据两次构建/打包，产物逐字节相同；tar 条目按名称排序、时间戳和权限固定，元数据排除临时路径与当前时间。分别验证 unsigned 包及给测试 key 的 signed 包；release.Verify 对实际清单字节通过，改动签名字节或清单后拒绝。归档长度/摘要覆盖实际完整文件，包内无自包含归档摘要。包输入只收普通文件和目录，路径/链接/大小写冲突按 §9.1 拒绝。安全解压上限和完整拒绝矩阵直接引用 T-09.2 / #19、签名信任转换引用 T-09.4 / #163；本切片验证自己生成的包能接入预检。
9) i18n 打包（V-54）：合法 en-US/其他 locale 与 metadata/config/commands/messages 经原检查并完整进入进程包，config_schema.json 采用与嵌入一致的规范内容；缺少 en-US 或占位符不一致时 check 与 pack 均失败，不发布成功归档。SDK 的语言选择、运行期回落引用 #102、T-08.13/#137；配置 schema/迁移引用 #103 与 T-08.12/#114。本切片不重复关闭 WebUI 语言切换分项。
10) 既有开发与库模式证据：TEST-MATRIX §3.2 的 #149/#160/#171/#186 已覆盖库模式生命周期、固定开发代次、配置/数据库、双传输、热部署、混合运行与 panic。补本切片顶层命令及 init 比对后按分项追加，保留原有未运行平台与 T-09.14 配置管理承接；T-03.10/#180 的真实进程/本地 Bot API 与故障矩阵继续引用。

测试（按编码规范「测试提交规则」）：
- 提交：仅新增 cmd/alonabot/plugin_cli_e2e_test.go 的 TestDevelopmentCLI 一个主路径集成测试（unix）。在临时目录 init 插件→build 宿主平台→check→test（实际 Session/进程、注册示例命令并正常停止）→pack 使用临时测试 key→release.Verify→packages.Preflight→MoveTo。核对 init 契约形状、无 main.go、包内角色与资源。理由：首次通过真实 CLI、生成器、Go 编译器、SDK/离线 Core 会话、文件系统与包预检闭合开发发布主路径，补现有构建与预检各自测试间的接线空缺；只比对模板字符串不足以覆盖此协作路径。测试 key、工程与产物均在 t.TempDir() 生成。
- 本地自测（*_local_test.go，忽略、不提交，PR 列命令与结果）：上述矩阵其余分项，包括四角色、只读与陈旧生成文件、非法字段/迁移/i18n、命令帮助与旧入口、包资源/重复构建字节、unsigned/signature 拒绝、非空目录、编译/准备/停止失败，以及 echo/media 顶层命令和 devbot 模板语义比对。Go 侧真实失败用最小临时工程，不新建 testdata 插件。向 sdk/ 与 internal/core/supervisor/ 的新增测试一律留本地，保留现有 supervisor_test.go。
- 沿用测试：对新的共享构建和生成命令重跑现有 TestDevelopmentReload、TestDevelopmentExample、TestLifecycleOverTCPAndStdio、TestTelegramDevelopment，以及 #191 合并后的 TestSDKExamples；保留原断言。其余已有规则证据直接引用，无新增同类提交测试。
- 五平台 build 本地取证：linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64；记录交叉编译产物、manifest 与命令，平台实际运行单独记录。仓库内生成器的 host/target/tags 选择复用已有测试与本切片最小联验。

完成条件：
- 改动 Go 文件 gofmt -l 无输出；go vet ./...、go test -race ./...、go build ./... 通过。更新所有生成指令后 go generate ./... 成功，二次运行生成源码内容稳定；只通过生成源更新生成产物。生成二进制位于忽略的输出目录。
- 上述全仓 race 中定向确认 TestDevelopmentCLI 与沿用测试的结果；如需独立取证，运行 go test -race -count=1 ./cmd/alonabot -run '^(TestDevelopmentCLI|TestSDKExamples|TestDevelopmentReload|TestDevelopmentExample|TestLifecycleOverTCPAndStdio|TestTelegramDevelopment)$'，同候选已有有效结果时复用。
- 用构建好的 alonabot 分别运行 check --dir examples/plugins/echo、check --dir examples/plugins/media，以及 build --dir examples/plugins/echo --out <临时目录>、build --dir examples/plugins/media --out <临时目录>；为 init 的当前目录模式、给定空目录模式和 devbot 形状比对保存实际证据。
- TEST-MATRIX §3.2 按 §3 写测试位置、完整命令、平台/工具版本、日志位置、实际结果与未运行归属，明确列出 goal-14/goal-70 的三项补证。PR「文档同步」列出命令树、生成入口、开发文档与规格缺口的审查者同步项。
- 按 T-09 停止条件核对：任何协议/Session 校验、包路径/来源校验或 secret 泄漏失败，均阻塞交付；身份、平台或静态错误在管理入口停止旧版前拒绝，沿用现有预检链。本切片发现的范围外缺陷与接口缺口交审查者确定归属，不通过吞错、伪元数据或静默降级获得成功结果。

不做：T-09.4 的管理/安装服务重做，T-09.5～T-09.7 的 Journal/提交/恢复，T-09.14 的 reload 与管理配置；T-13.3 的 run/热部署重做、库模式宿主自动生成或新的根包公开选项；PR #191 的 SDK schema_id 修复和示例业务升级；Agent/plugintest 的在途功能；远程发布/市场、官方发行归档与发布签名流水线（T-14.3）；新增依赖、旧 CLI 别名或兼容层、构建期注册标记、sdk.Base/Config 类型恢复、新增 testdata 插件与绑定 SDK 的提交测试。
```

### goal-40 · T-09.5 候选 Journal 与准备编排

状态：已完成（[PR #173](https://github.com/AlonaBot/AlonaBot/pull/173)，合并提交 `2bb8181`），证据见 TEST-MATRIX §3.2。合入：状态库 `journal` 表与 `state.Journal`、`CreateJournal`、`UpdateJournal`、`GetJournal`、`Journals`、`GenerationConfigVersion`；`supervisor.Prepare(ctx, store, spec) (*Prepared, error)`、`(*Prepared).Activate(ctx, commit func(context.Context) error) (*Running, error)`、`(*Prepared).Close() error`；`internal/core/journal` 的 `Kind`（first_install、reload、update）、`Options{Store, Secrets, Config, Settings, DataDir, State, Spec, Stop, Resume}`、`Result{OpID, Generation, Prepared, Session}`、`New`、`(*Orchestrator).Prepare(ctx, pluginID, kind, staged)`、`(*Orchestrator).Abort(ctx, r, cause)`；根包装配结构的 `journal` 字段；`manage.Candidates.Update(ctx, id, staged *packages.Candidate)`；host.Prepare 的 REGISTER 绑定检查。以下为派发时的语句。
- 生产入口的分界（DECISIONS 2026-11-13 变更记录）：本切片只交付 Journal、两阶段 supervisor 入口与 Orchestrator 的 Prepare/Abort，不接入生产的管理动作与首次启用；`manage.Options.Candidates` 保持缺省的 `unsupportedCandidates`，`startInstalled` 的首次启用保持现状。管理动作接线、首次启用切换与没有 Active 的 Update 由 T-09.6（goal-41）在 Commit 可用时一并交付，避免出现「停旧版、准备、中止、返回 UNSUPPORTED」的过渡行为。
- `manage.Candidates.Update` 改为交接完整的 `*packages.Candidate`（BuildDigest、Release 与发布者随包对象交接）。
- 从 `supervisor.Launch` 提取只到 PREPARED 的 `supervisor.Prepare`/`Prepared`，Launch 保持现有签名与行为。
- external 获准维护后的失败按 CONTRACTS §6.3 保持停机；当前没有维护许可来源，一律下发 `manual_migration_approved=false`。
- Journal 表的列见接口 1；`state/installs.go` 把 Journal 纳入卸载预览与事务清理。
- 没有 Active 的 reload/update、role 为 webui 的组件与开发来源的组件，Prepare 在写 Journal 前返回 FAILED_PRECONDITION。
- REGISTER 的可信 schema 与配置版本绑定：WIRE-FORMAT §3.1 要求 REGISTER 带 `config_schema_hash` 与 `candidate_config_version`、Core 原子校验，main 的 Core 尚未在 Bootstrap 填入摘要，也未比对；本切片补齐（行为 7）。

已有接口（审查者按 main `37b0ada` 核对）：

- `internal/core/manage/candidates.go`：现有 `Candidates` 为 `Reload(ctx context.Context, id string) error`、`Update(ctx context.Context, id, staged string) error`、`Rollback(ctx context.Context, id, generation string) error`；默认 `unsupportedCandidates` 三项都返回 `UNSUPPORTED`。`Service.Update(ctx context.Context, id string, req InstallRequest) error` 完成预检后只传 `candidate.Dir`，退出时 defer `candidate.Discard()`。`packages.Candidate` 的 `Manifest`、`Release`、`Asset` 没有随目录交接。
- `internal/core/manage/lifecycle.go`：`Service.Enable(ctx context.Context, id string) error` 经 `enable` 检查依赖、调用 `Options.Start`，成功后才 `SetEnabled(true)`。`Service.Restart(ctx context.Context, id string) error` 要求 enabled 且有 Active，按 `Stop → Start` 恢复同一代次。首次启用的 shortcut 实际位于 `assembly.go` 的 `startInstalled`，当前直接 `AddGeneration → supervisor.Launch(NewCandidate:true)`。
- `internal/core/candidate/database.go`：`PrepareDatabase(ctx context.Context, req Request) (*Prepared, error)`；`Request` 字段为 `PluginID string`、`PluginDir string`、`Reason pb.ActivationReason`、`FromGeneration string`、`NewGeneration string`；`Prepared` 字段为 `PluginID string`、`Generation string`、`DatabasePath string`、`Created bool`。FIRST_INSTALL 只建目录，SDK 创建首库；RELOAD/UPDATE 以 Backup API 复制；RESUME 复用路径。返回前释放 `runtime.lock`。
- `internal/core/candidate/discard.go`：`Discard(ctx context.Context, st *state.Store, secrets *secret.Store, p *Prepared) error`；先删未提交代次与 secret 路径，再删候选目录，最后按全部代次与 Pending 保护 secret。已提交或 Active 代次拒绝删除。包文档要求调用方持管理操作锁，先确认旧树退出，候选失败时先确认候选树退出再清理。
- `internal/core/state/state.go`：`(*Store).LockPlugin(pluginID string) (unlock func())`、`AddGeneration(ctx context.Context, pluginID, generationID, configSnapshot string, configVersion uint32) error`、`Active(ctx context.Context, pluginID string) (string, error)`、`CommitActive(ctx context.Context, pluginID, expectedOld, newGen string) error`。现有 `CommitActive` 原子写 Active 与 `generation.committed_at`，尚无正式 Journal。`schema.sql` 没有 Journal 或代次程序包路径列。
- `internal/core/state/secrets.go`、`pending.go`、`candidate.go`：`SetGenerationConfig(ctx context.Context, generationID, snapshot string, secretPaths []string) error`、`GenerationSecretPaths(ctx context.Context, generationID string) ([]string, error)`、`GenerationConfig(ctx context.Context, generationID string) (string, error)`、`PluginGenerationConfigs(ctx context.Context, pluginID string) (map[string]string, error)`、`DeleteCandidateGeneration(ctx context.Context, pluginID, generationID string) error`；代次的 `config_version` 已存储，当前没有公开读取它的方法。
- `internal/core/config/service.go`：`New(store *state.Store, dataDir string) *Service`；`(*Service).ImportSnapshot(ctx context.Context, pluginID string) (Pending, error)`、`Pending(ctx context.Context, pluginID string) (Pending, bool, error)`。这些方法自行持 `Store.LockPlugin`；编辑快照缺失返回 NOT_FOUND，导入时只核对 JSON 对象，不先用新 schema 否决旧配置。
- `internal/core/settings/instances.go`：`(*File).TrialInstances(adapterID string, schema *configschema.Schema) []InstanceProblem`、`PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []InstanceProblem, error)`；问题包含 `InstanceID`、`Path`、`Reason`。Trial 不写文件；Prepare 的迁移写回会备份。`BindAdapter(adapterID, packageID string) (bool, error)` 保护实例包归属。
- `internal/core/supervisor/supervisor.go`：`Launch(ctx context.Context, store *state.Store, spec Spec) (*Running, error)`，当前从进程启动一直执行到 `host.Activate`，`NewCandidate=true` 时调用 `CommitActive`，随后 Bind HTTP、Attach Scheduler。当前不存在只到 PREPARED 的 supervisor 入口。`(*Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult` 与 `WaitExit() (abnormal bool, state string, err error)` 负责树清理；Stop 的 Frozen/Stopping 钩子处理 Scheduler/HTTP 与装配 Detach。
- `internal/core/host/prepare.go`、`activate.go`：`Prepare(ctx context.Context, conn *wire.Conn, exp Expect, boot *pb.Bootstrap, lim Limits) (*Prepared, error)` 已停在 PREPARED；`Prepared` 含 `SessionID`、`Register`、`ConfigSnapshot`、`RouteHash`、`ConfigDigest`。`Activate(ctx context.Context, conn *wire.Conn, prep *Prepared, opt ActivateOptions) (*Instance, error)` 独立执行提交回调与激活。预算由 `NewLimits(v tune.Values) Limits`、`NewStopTiming(v tune.Values) StopTiming` 提供。
- `alonabot.go`：`Start(ctx context.Context, opts ...Option) (*Core, error)` 装配同一 `settings.File`、state、管理服务；管理 Start/Stop 经 `Core.operation` 串行进入运行循环，当前未传 `manage.Options.Candidates`。`assembly.go`：`assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error)`；`(*Core).startInstalled(ctx context.Context, dir string, install state.Install) (*runningComponent, error)`、`stopComponent(r *runningComponent, t0 time.Time, timing host.StopTiming, reason pb.StopReason) error`；`(*assembly).detach(r *runningComponent)` 按 Session 撤销 Router、Bus、Scheduler、HTTP、remote 与 Broker 关联。
- `internal/core/packages/preflight.go`：`Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error)`、`(*Candidate).MoveTo(packagesRoot string) (string, error)`、`Discard() error`。MoveTo 用已验证的 `Asset.SHA256` 选择不可变目录，成功后清空 `Dir`；没有在该目录中保存完整 release 元数据。这里沿用现成摘要，不新增摘要计算。
- 同时在途、可能与本切片改同一批文件的 PR：#169（goal-127，`state/schema.sql`、`state/installs.go` 的清理清单、`assembly.go`）与 #170（goal-59，Router 与 SDK）；goal-69（根包与 cmd/alonabot 的开发运行）。从最新 main 创建；后合并者按最新 main 合并调用点、初始 schema 与清理清单。

```text
/goal 使用 $alonabot-execute 推进 goal-40（切片 T-09.5：候选 Journal 与准备编排）。goal 目录 goal-40，分支 goal-40，worktree .worktrees/goal-40，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：为 reload、update 与首次启用实现正式候选准备（本切片交付 Prepare 与 Abort，生产入口由 T-09.6 接入）：预检通过后，以预期旧 ActiveGeneration 持久化 Journal 及程序、配置、secret、数据库资源引用；冻结新根派发，排空并停止旧进程，确认进程树退出与 runtime.lock 可取；受管 SQLite 调用已有候选复制，创建新的 Generation 与 Session，以已有 Bootstrap 下发启动原因和阶段预算；REGISTER 校验及 secret 抽取后持久化候选资源，停在 PREPARED。提交与激活由 T-09.6 接续。受管候选提交前失败时清理候选，以旧版的新 Session 恢复；external 维护失败按 CONTRACTS §6.3 保持停机。

依据：
- spec/PLAN.md §3.1 → spec/plan/core.md 的 T-09.5 切片表、T-09「实施动作」步骤 4、「前置输入与边界」与「对应 V 验收与停止条件」；T-08 步骤 7 的候选交接及其停止条件。该 goal 属于 Core 章节，工作包位于 spec/plan/core.md。
- spec/CONTRACTS-v0.1.md §4.1、§4.2、§5.1～§5.3、§5.5、§6.1、§6.3、§8、§9.3、§12.1～§12.3、§14.5、§16。
- spec/WIRE-FORMAT.md §3.1、§3.2、§3.4：准备与激活分界、Bootstrap、Adapter 共用生命周期；spec/MESSAGE-MODEL.md §1、§1.2、§5、§6.3：可信事件、父调用链与按钮的代次归属沿用已有行为，准备期保持业务门禁。
- spec/adr/DECISIONS.md D-26、D-48、D-68、D-75、D-76、D-80、D-82、D-85、D-86、D-114、D-127、D-134；管理锁与实例归属沿用已合并的 D-139 实现，其维护者确认状态由审查者跟进。
- spec/TEST-MATRIX.md V-29；关联复验 V-24、V-36 的候选与锁分项、V-31/V-39 的维护拒绝与 resume 分项、V-21 的准备期门禁。V-30 的正式提交及崩溃恢复留给 T-09.6/T-09.7。

范围（只新增或修改这些）：
- internal/core/state/schema.sql：在唯一初始迁移加入 Journal 表；新增 journal.go：Journal 模型、读写、阶段转换及旧代次配置版本读取。
- internal/core/state/installs.go：仅把 Journal 纳入既有卸载预览与事务清理清单；保留 main 及 goal-127 的其他清理对象。
- 新增 internal/core/journal/：候选资源持久化、Prepare/Abort、管理动作入口；包内可按平台分文件处理落盘。
- internal/core/supervisor/supervisor.go；可新增 prepare.go：从现有 Launch 提取只到 PREPARED 的入口及资源关闭、继续激活的方法，Launch 与 Running 的现有调用契约保留。
- internal/core/manage/candidates.go 与 service.go 中的唯一调用点：把预检包的完整交接改为下面的签名，同步默认实现；预检、来源、信任、pin、依赖确认仍沿用现有流程。
- internal/core/host/prepare.go：REGISTER 的 config_schema_hash 与 candidate_config_version 绑定检查（行为 7）。
- 根包 alonabot.go、assembly.go；可新增 journal.go：在 assemble 中创建 Orchestrator 并保存在装配结构（供 T-09.6 接线与本地测试使用），以及共用的已安装包 Spec 构造（startInstalled 改用它，行为不变）。不改 manage.Options.Candidates 的生产装配与 startInstalled 的首次启用分支。
- 新增根包 journal_test.go（unix）的一项 SQLite/文件系统集成测试；pipeline_test.go 仅允许保留既有断言的装配同步。真实 SDK 候选进程验证写在根包 journal_local_test.go；上述相关包可写 *_local_test.go。

允许的新跨模块边：无。journal、candidate、config、secret、settings、state、supervisor 均属于 ARCHITECTURE 的 internal/core 模块；根包装配、Core 到 contracts/internal/wire/internal/oslock 的边已允许。Core 保持经协议接入 SDK 进程。

接口（以下为本切片新增或调整的接口；现有接口复用状态段所列签名；内部辅助函数自定）：
1. internal/core/state
   type Journal struct {
       OpID, PluginID, Kind, ExpectedOld, NewGeneration, Phase string
       PackageDir, DatabasePath string
       Entrypoint, Version, PublisherKeyID string
       BuildDigest []byte
       OwnsPackage bool
       WasRunning, WasEnabled bool
       SessionID []byte
       CreatedAt, UpdatedAt time.Time
   }
   func (s *Store) CreateJournal(ctx context.Context, record Journal) error
   func (s *Store) UpdateJournal(ctx context.Context, record Journal) error
   func (s *Store) GetJournal(ctx context.Context, opID string) (Journal, error)
   func (s *Store) Journals(ctx context.Context, pluginID string) ([]Journal, error)
   func (s *Store) GenerationConfigVersion(ctx context.Context, generationID string) (uint32, error)
   Journals 按 created_at、op_id 排序，pluginID 为空时列全部记录，缺单项返回 state.ErrNotFound。

   Journal 表的列固定对应上述字段：op_id TEXT PRIMARY KEY；plugin_id TEXT NOT NULL REFERENCES install(plugin_id)；kind、expected_old、new_generation、phase、package_dir、database_path、entrypoint、version、publisher_key_id 为 TEXT NOT NULL；build_digest、session_id 为 BLOB NOT NULL；owns_package、was_running、was_enabled 为 INTEGER NOT NULL；created_at、updated_at 为 INTEGER NOT NULL，使用 Unix 毫秒。session_id 在握手前为空 BLOB；NONE 后端的 database_path 为空串。new_generation 在 started 时已分配，generation 行随后创建，因此该列使用逻辑引用。
   kind：first_install、reload、update。phase：started、stopped_old、candidate_ready、prepared、committed、failed、cleaned。正常准备顺序为 started → stopped_old → candidate_ready → prepared；首装也记录 stopped_old，含义为旧写入者退出条件已满足。任一提交前阶段可转 failed，failed 在清理与恢复完成后转 cleaned；committed 的写入由 T-09.6 的提交事务完成。本切片的 UpdateJournal 拒绝写 committed。
   UpdateJournal 用事务复核原阶段、操作身份、expected_old/new_generation；身份与预期旧值保持不变，资源引用按实际取得/释放更新。清理错误保留 failed 及仍持有的引用。时间与阶段在同一次状态写入中更新；SQL 事务只处理元数据。

2. internal/core/supervisor
   type Prepared struct {
       Registration *host.Prepared
       // 其余字段私有：进程树、连接、Listener、Spec、日志与清理所有权。
   }
   func Prepare(ctx context.Context, store *state.Store, spec Spec) (*Prepared, error)
   func (p *Prepared) Activate(ctx context.Context, commit func(context.Context) error) (*Running, error)
   func (p *Prepared) Close() error

   Prepare 复用 Launch 的身份、secret、配置快照、预算和失败清理，只完成 host.Prepare。返回后进程与连接由 Prepared 持有；候选没有 host.Instance，也未接入业务装配。Close 关闭准备连接，按原 StopTiming 等待退出及必要的树终止，合并并返回清理错误；确认清空后才释放进程资源。Prepare 失败而树清空尚未确认时，返回非 nil 的句柄及错误，供调用方保留和关闭该进程。
   Activate 把既有 host.Activate 及激活成功后的接入提取为共用方法，成功后将资源所有权转交 Running；commit 为 nil 只用于已有提交记录的 resume。Launch 保持现有签名和调用行为，内部复用两阶段实现；其 NewCandidate 路径仍传既有 CommitActive 回调。journal.Prepare 只调用 Prepare；T-09.6 接续时才调用 Prepared.Activate。

3. internal/core/manage
   type Candidates interface {
       Reload(ctx context.Context, id string) error
       Update(ctx context.Context, id string, staged *packages.Candidate) error
       Rollback(ctx context.Context, id, generation string) error
   }
   Service.Update 在全部现有预检通过后传 candidate 本身；保留 defer candidate.Discard()。移动成功由 MoveTo 清空 Dir，之后由 Journal 持有不可变包资源。同步 unsupportedCandidates.Update；Reload/Rollback 和管理 HTTP、CLI 的公开参数保持现有签名。

4. internal/core/journal
   type Kind string
   const (
       KindFirstInstall Kind = "first_install"
       KindReload       Kind = "reload"
       KindUpdate       Kind = "update"
   )
   type Options struct {
       Store *state.Store
       Secrets *secret.Store
       Config *config.Service
       Settings *settings.File
       DataDir string
       State func(id string) string
       Spec func(ctx context.Context, in state.Install, packageDir string, boot *pb.Bootstrap) (supervisor.Spec, error)
       Stop func(ctx context.Context, id string, reason pb.StopReason) error
       Resume func(ctx context.Context, id string) error
   }
   type Orchestrator struct { /* 私有状态 */ }
   type Result struct {
       OpID, Generation string
       Prepared *candidate.Prepared // NONE 后端为 nil
       Session *supervisor.Prepared
       // 操作上下文与中途资源所有权保存在私有字段。
   }
   func New(opt Options) *Orchestrator
   func (o *Orchestrator) Prepare(ctx context.Context, pluginID string, kind Kind, staged *packages.Candidate) (*Result, error)
   func (o *Orchestrator) Abort(ctx context.Context, r *Result, cause error) error

   Prepare 的 staged 仅在 KindUpdate 时非 nil；其身份与 release 已由 manage 预检。静态阶段失败且尚未写 Journal 时返回 nil、error；写记录后失败返回可 Abort 的 Result、error，由调用方 Abort，避免重复清理。成功返回的 Result 由调用方负责后续提交或 Abort。
   Spec 在停旧版前构造完整 supervisor.Spec，复用根包的包路径、资源、schema、i18n、handler、日志、预算与 detach 装配；此时只读取与准备输入。Spec 的 Bootstrap 随后填入实际候选 DB 路径与 generation。Stop/Resume 使用 Core 的既有停止、已提交代次启动实现。
   FirstInstall、Reload、Update 的管理入口方法由 T-09.6 新增，Rollback 由 T-09.7 新增，本切片不提供。

5. 根包装配与既有 proto
   Orchestrator 在 assemble 中创建，不交给 manage.Service。Options.Stop/Resume 直接调用运行循环的既有停止与已提交代次启动实现，不经 Core.operation（T-09.6 的管理桥接会在 operation 内调用 Orchestrator，二者不能互相等待）。开发来源沿用 dev:<组件 ID>、config.json/dev.sqlite 与开发重启路径。
   Store.LockPlugin 继续由 config/supervisor 的各自方法取得，管理服务保留自己的逐包互斥；编排器调用这些方法期间保持管理操作串行，避免重复取得同一 Store 锁。

   复用 contracts/proto/alonabot/v1/control.proto 的现有字段，字段号与生成代码保持现状：
   Bootstrap：generation_id=1、candidate_config_json=2、config_schema_hash=3、optional previous_config_version=4、target_config_version=5、database_descriptor=6、phase_limits=8、preparation_deadline_unix_ms=9、log_level=10、session_limits=11、default_locale=12。
   DatabaseDescriptor：backend=1、sqlite_path=2、migration_mode=4、applied_migration_ids=5、activation_reason=6、manual_migration_approved=7。
   复用 FIRST_INSTALL/RESUME/RELOAD/UPDATE 枚举；previous_config_version 取旧代次元数据，target_config_version 与 schema 输入取候选包。schema 摘要复用 configschema.Schema.Hash，沿用既有协议摘要实现。

行为：
1) 停机前完成静态工作。读取安装记录与 Active，保存先前 enabled/运行状态。reload 用 ImportSnapshot 导入编辑快照为以当前 Active 为 base 的 Pending；update 取有效 Pending，否则取 Active 的完整配置；首装取编辑快照，文件缺失时沿用当前首装空对象语义。保留未知键、显式零值和 secret 引用，配置对象检查后把旧格式交给 SDK 迁移。读取候选 manifest、入口、schema 与 i18n 等资源，沿用既有错误和身份校验。
2) Adapter 更新在停机前解析新包 adapter_config_schema.json，使用同一个 settings.File 的 TrialInstances，只对归属于本组件的实例准备。失败完整返回 InstanceID、Path、Reason，保留旧进程、设置文件和旧 Active。准备阶段只试算；成功提交后的 PrepareInstances/备份写回由 T-09.6 接续。
3) external 的已有安装在 Journal 与停机前返回 FAILED_PRECONDITION，safe_message 为 NEEDS_MANUAL_MIGRATION。现有管理入口没有显式维护许可来源，全部下发 manual_migration_approved=false；首装与普通 resume 按契约放行。T-09.14/T-09.9 接入显式维护后，许可须绑定本次 Session/Generation/DB；获准 external 候选失败时保留停机及诊断，按插件的人工恢复步骤处理。
4) 分配 op_id/new_generation，保存 phase=started、expected_old 及程序引用。update 从 packages.Candidate.Asset 取得已验证的完整 BuildDigest，从 Manifest/Release 取得入口、版本、发布者；移动到不可变包目录，确保文件及目录持久化，再保存 Journal 的最终路径与所有权。install 的 Version、Entrypoint、BuildDigest 保持旧值，候选的程序身份暂存在 Journal，供 T-09.6 提交接续；显式确认后的发布者信任绑定沿用 manage 当前的写入时机。reload/首装复用已有不可变程序目录。
5) 冻结并停止旧进程。经原 Stop 路径进入 Draining，拒绝新根调用，既有合法子链按原 deadline 和停止 deadline 排空；Scheduler/HTTP 使用 Frozen 钩子，STOPPING 时按 Session 撤销 Router/Broker/Bus/HTTP/Adapter 等绑定。沿用原停止总预算并确认整棵树清空，然后记 stopped_old。停止未确认时保留诊断与 Journal，后续复制、启动候选及恢复旧版都等待该前提成立。首次安装也须核对 runtime.lock，处理已有残存进程。
6) 受管 SQLite 调用 candidate.PrepareDatabase，使用绝对 PluginDir、旧 Active 与新 Generation；FIRST_INSTALL 只准备目录，RELOAD/UPDATE 备份到独立库，成功后由该函数释放 runtime.lock。NONE 后端（含 Adapter）跳过数据库复制，以 oslock.TryLock 临时核对相同 runtime.lock 并释放，DatabasePath 留空。创建未提交 generation 行，配置版本取目标版本，初始快照只存引用或空对象；已有引用在所有准备阶段受保护。资源已登记后记 candidate_ready。
7) 调用 supervisor.Prepare(NewCandidate=true, ExpectedOld=旧 Active)。启动原因由操作 kind 映射；DefaultLocale、LogLevel 与各阶段预算取同一 Core 当前设置/生效值。SDK 顺序沿用配置迁移、OnConfigLoad、补值和校验、允许的 DB 准备、OnStart、REGISTER。Core 在 Bootstrap 填入候选包 config_schema.json 的 configschema.Schema.Hash（只对 role 为 plugin 或 agent 的包；没有该文件或 Adapter 时为空），host.Prepare 在 Bootstrap 的 config_schema_hash 非空时要求 REGISTER 的 config_schema_hash 相同，并要求 REGISTER 的 candidate_config_version 等于 Bootstrap 的 target_config_version，不符时返回 FAILED_PRECONDITION 并按准备失败处理（该检查对所有经 host.Prepare 的启动生效，普通启动经共用的 Spec 构造同样填入摘要）。Core 复用 host 的 REGISTER 验证与 secret.Extract，将不可变配置及 secret 路径写入新代次，再生成 PREPARED。准备成功后保存 SessionID，确认资源及引用持久化，记 prepared，返回 Result。保持 Active 与旧包安装身份，候选 OnReady 尚未执行，命令、事件、Service、Bus、Scheduler、HTTP 和 Adapter 实例均未向候选开放。
8) 本切片不接入生产入口：manage.Options.Candidates 保持缺省，startInstalled 的首次启用保持现状，二者由 T-09.6 在 Commit 可用时一并切换。Prepare 在写 Journal 之前对下列情况返回 FAILED_PRECONDITION：kind 为 reload 或 update 而没有 Active；role 为 webui；开发来源（固定代次 dev:<组件 ID>）；kind 与 staged 不匹配（只有 update 带 staged）。验证由本地测试直接调用 Prepare，检查 PREPARED 后显式 Abort。
9) Abort：先记 failed；停止并确认候选树退出，再清理本操作拥有的未提交 generation/数据库/新程序包与孤立 secret。SQLite 资源沿用 candidate.Discard；NONE 分支使用同一代次删除与 secret 保护规则。当前 Active、已提交代次、其他 Journal、Pending 及保留代次仍引用的资源持续保护；只清理本操作已取得所有权的资源。释放某项引用与更新 Journal 状态按实际清理结果进行。
10) 受管回退只在旧树及候选树均已确认退出后执行；以旧安装记录、旧 Active、旧快照、旧 DB 和新 Session 恢复原先运行的组件。失败发生在停旧版之前时，只清理已取得的候选资源，保留仍有效的旧 Session。首装失败保持未启用，原先停用的组件保留停用状态；external 维护失败保持停机。清理与恢复全部成功才记 cleaned，失败保留 failed 和尚需处理的引用，供 T-09.7 接续。管理请求取消后仍完成必要清理，沿用已有停止/准备预算，保留最初错误及清理/恢复错误。
11) restart、正常 Core 启动的已有 Active 恢复及异常恢复复用原提交记录，沿用 activation_reason=resume；这些路径保持 Generation/配置/secret 引用，Journal 数量不变。候选进程在 PREPARED 前后失败也归当前管理操作处理，避免进入普通异常进程自动重启分支。
12) Journal 每次阶段变化以 INFO 记 op_id、plugin_id、kind、phase；失败记录具体原因，日志及管理错误省略完整配置、secret、Bootstrap、内部 HTTP 凭证。卸载预览与删除沿用同一清单：两种 purge 将可清理的 Journal 与其关联数据在一个状态事务中处理；仍待恢复的操作先报告 FAILED_PRECONDITION，保留恢复材料。只卸载继续保留代次与恢复所需引用。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：根包 journal_test.go 的一个 TestCandidateJournalResources（//go:build unix）。通过公开 state、candidate、secret 入口协作真实 SQLite、WAL、文件系统和系统锁：创建旧 Active 与 Journal，取得候选副本并保存 generation/secret 引用，再经既有 candidate.Discard 清理本操作资源，核对独立 DB、旧值保护及 Journal 的持久化结果。该项验证编排依赖的真实资源交接；完整 Orchestrator.Prepare/Abort 使用下列本地进程测试取证。理由：本切片新增持久化资源引用，SQLite/文件/锁之间的交接需要长期集成回归。
- 本地自测：根包 journal_local_test.go 复用 pipeline_test.go 的生产 assemble、测试二进制分进程与 TCP/stdio 传输。必要的候选模式分流仅写本地测试，通过现有 TestMain 在执行 m.Run() 时以本地 Test 辅助进程运行 SDK；插件对象使用 sdk.NewPlugin、p.OnConfigLoad、p.OnStart、p.OnReady，受管迁移用 sqlitedb 的现有登记函数。补充临时包工程只写 t.TempDir()，按 D-86 留在本地测试，记录实际命令、平台、进程与结果。
  a. reload 到 PREPARED：旧进程已退出，新 Session 不同；WAL 内容进入独立 DB，候选写入未污染旧库；Journal 保存旧 Active、程序/DB/配置/secret 关联；Active、enabled 及旧安装身份不变；候选 OnReady/业务/HTTP/任务未执行；显式 Abort 后旧代次以新 Session 接单。
  b. update 到 PREPARED：以 packages.Preflight 得到的 *packages.Candidate（与 manage.Service.Update 交出的对象相同）调用 Prepare；新程序版本与 BuildDigest 用于 HELLO，候选配置迁移到目标版本，旧包、旧 schema 与旧 secret 可恢复；Abort 后候选新包按引用清理，旧身份保持。另以 manage.Service.Update 配一个本地测试用的 Candidates 实现，核对 defer Discard 不删除已转交的不可变目录。
  c. first_install 成功到 PREPARED：expected_old 为空，新 generation 未提交，首库由 SDK 创建，OnReady 与业务均未执行，显式 Abort 后仍未启用。OnConfigLoad、DB 迁移、OnStart、REGISTER 失败时，已有安装的旧版以新 Session 恢复，旧 DB/配置/secret 保持；初装失败仍未启用；清理或恢复失败时 Journal 保留 failed 与真实错误。
  d. 旧树/残存子进程未清空、runtime.lock 被占用、备份空间或校验失败：保持 Active，无第二写入者；使用已有真实锁与文件权限/故障手段，测试控制点写在 *_local_test.go。准备后的候选树未确认退出时，同样保留资源并停止恢复。
  e. Adapter/NONE 后端：无旧 db.sqlite 时照常准备，Adapter TrialInstances 失败发生在旧进程停止前且设置原文保持；PREPARED 前后实例未启动、游标仍归原组件；恢复保留本次运行已选择停用的实例。
  f. 已有 external 安装的 reload/update 在停机前返回 NEEDS_MANUAL_MIGRATION，Journal 数量和 Session 保持；FIRST_INSTALL 与 resume 的许可位为 false；插件自行声称获准不影响 Core 输入。有效维护入口与其失败停机分支由后继管理切片补端到端证据。
  g. 首装取编辑快照与缺省空对象、Pending base 陈旧、配置迁移版本、unknown/secret 保留、旧程序引用保护、清理失败重入、同包并发管理、无 DB 与未运行 update 的状态保持、phase 非法后退/伪 committed、resume 不增 Journal，Prepare 拒绝的四种情况（无 Active、webui、开发来源、kind 与 staged 不匹配），REGISTER 的摘要或配置版本与 Bootstrap 不符时准备失败并恢复旧版，以及 Store 锁不自锁。
- 既有提交测试的签名同步只保留原断言；SDK 与 supervisor 的新增测试全部用 *_local_test.go，不新增 testdata 插件。

完成条件：
- 范围内完成静态预检、两阶段进程入口、持久化 Journal、SQLite/NONE、REGISTER 绑定检查、Orchestrator 装配、Abort 与引用保护；所有公开调用者按最终接口同步；交接给 T-09.6 的 Result/Journal 可继续同一候选 Session。
- V-29 与本切片关联分项具有实际证据；按 TEST-MATRIX §3 的格式在 PR 写明测试位置、命令、平台、结果与未运行项；规格文档由审查者同步。
- gofmt -l . 无输出；go vet ./...；go test -race ./internal/core/...；go test -race -run '^TestCandidateJournalResources$' .；go test -race ./cmd/alonabot/...；go build ./...。本地进程测试另记实际命令和结果。
- 遇到候选污染旧数据、未确认旧树退出便启动写入者、external 未授权准备、资源被仍有效引用时删除、Journal/Active 权威含混，立即停止交付并修复。涉及范围外修复、proto、依赖或切片衔接裁决缺失时按共同约束记阻塞，由审查者补齐语句后继续。

不做：Active+COMMITTED+Journal 的唯一提交事务、候选正式激活与提交后 Degraded、生产的管理动作接线、首次启用切换与没有 Active 的 Update（T-09.6）；Core 崩溃扫描、rollback 和保留代次回收策略（T-09.7）；显式维护许可的 CLI/API/WebUI（T-09.14、T-09.9）；开发热部署、dev 代次改造、静态 WebUI 候选化；新增 proto、SDK 公共 API、依赖、调优值或通用审批机制。
```

### goal-57 · T-11.4 工具与会话的验收

状态：已完成（[PR #172](https://github.com/AlonaBot/AlonaBot/pull/172)，合并提交 `15f9495`），证据见 TEST-MATRIX §3.2。合入：根包 `t11_e2e_test.go` 的 `TestToolsAndConversations`（T-11 七项验收）与 `pipeline_test.go` 的 `newPipelineAt`（可传调优值、可重开同一数据目录）；未新增生产代码。
- 原语句的 `internal/core/supervisor/testdata` 测试插件按 D-86 改为根包测试二进制扮演的插件，测试位置改为根包。
- 第 6 项拆成两部分：Service 的访问规则（Access）与继承 Actor、默认 Target 用 SDK 写；`RouteDescriptor.required_permissions` 分项由原始 wire/pb 的 provider 模式驱动。CONTRACTS §2.3 的路由权限对 Service 同样适用、`internal/core/tool/invoke.go` 已检查，但 SDK 的 `ServiceBuilder` 还没有声明入口；SDK 入口不在本切片。
- 运行时修改策略与会话的服务层操作由父测试进程直接调用 `tool.New(h.store, h.assembly.router)` 与 `conversation.New(h.store)`：`Catalog.Effective` 每次从状态库读取，另建的 Catalog 与生产实例共享同一 Store。管理 API 与页面分别属于 T-09.14、T-09.9。

已有接口（审查者按 main 核对）：

- `internal/core/tool/catalog.go`：`New(store *state.Store, r *router.Router) *Catalog`；`(*Catalog).Register(m *router.Mux)`；`Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error)`；`SetPolicy(ctx context.Context, pluginID, service string, p pb.ToolPolicy) error`；`Policies(ctx context.Context) ([]ToolPolicyEntry, error)`。`ToolPolicyEntry` 含 `PluginID, Service string`、`Default, Override pb.ToolPolicy`。`SetPolicy` 成功时覆盖已提交，UNSPECIFIED 删除覆盖。
- `internal/core/tool/invoke.go`：`(*Catalog).RegisterInvoke(m *router.Mux, opt InvokeOptions)`；`InvokeOptions` 的实际字段与类型见下方接口段。`assembly.go` 已将 `Call` 接到 `a.broker.Call`，将 `ConfirmTimeout` 接到 `v.ToolConfirmTimeout`，并注册 `core.tool.list/invoke/confirm`。
- `internal/core/buttons/buttons.go`：`New(store *state.Store, now func() time.Time, ttl time.Duration, current func(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool)) *Service`；`IssueConfirm(ctx context.Context, caller router.Caller, target *pb.Target, recordID string) (allow, deny string, err error)`；`InterceptConfirm(ctx context.Context, ev *pb.Event) bool`；`OnConfirm(fn func(context.Context, string, bool, *pb.Event))`；`RevokeConfirm(ctx context.Context, recordID string) error`。Router 的 `Options.Intercept` 已接 `a.buttons.InterceptConfirm`，确认点击在事件队列前消费。
- `internal/core/conversation/service.go`：`New(store *state.Store) *Service`；`Register(m *router.Mux)`；`List(ctx context.Context, pluginID string) ([]state.ConversationSummary, error)`；`Clear(ctx context.Context, pluginID, key string) error`；`Purge(ctx context.Context, pluginID string) error`。`assembly.go` 已注册 `core.conversation.append/list/clear`。
- `sdk/tool.go`：`ListTools(ctx context.Context) ([]ToolInfo, error)`；`InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error)`；`ConfirmOperation(ctx context.Context, operation, summary string) error`。后两者要求活跃 Handler 或延续调用的 context。`sdk/conversation.go`：`AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error)`；`ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error)`；`ClearConversation(ctx context.Context, key string) error`。
- `sdk/services.go`：工具经 `sdk.Serve(p, sdk.JSON[In, Out](target)).Access(...).Tool(description, sdk.ToolAllow 或 sdk.ToolConfirm).Handle(fn)` 登记；`ServiceBuilder` 当前只有 `Access`、`Tool`、`Handle` 三个公开方法。`sdk/plugin.go` 的 `NewPlugin() *Plugin`、`sdk/generated.go` 的 `UseGenerated(p *Plugin, g Generated)` 为现行插件对象与资源绑定入口。Service 输入、输出 Schema 可按 `service_chain_test.go` 写入 `Generated.ServiceSchemas`。
- `internal/core/manage/service.go`、`http.go` 当前提供安装、生命周期、实例与管理员入口，工具策略和指定会话的管理方法及 HTTP 路由尚缺。T-11 的服务层验收分别经 `tool.Catalog` 与 `conversation.Service` 驱动；T-09.14（goal-93）负责后续管理 API，T-09.9（goal-44）负责页面。已存在的卸载入口为 `manage.Service.Uninstall(ctx context.Context, id string, purge PurgeScope, confirm bool) error`（`lifecycle.go`），三种范围为 `PurgeNone`、`PurgeData`、`PurgeAll`。
- 根包 `pipeline_test.go`：`newPipeline(t *testing.T, stdio ...bool) *pipelineHarness`；`(*pipelineHarness).install(id, mode string)`、`launch(id, mode string, priority int) *runningComponent`、`stop(c *runningComponent)`、`deliver(input pipelineInput) *pb.ErrorInfo`、`inject(input pipelineInput)`、`snapshot() []pipelineRecord`、`await(predicate func([]pipelineRecord) bool)`、`idle()`。`TestMain` 按 `ALONABOT_PIPELINE_MODE` 运行 `sdk.Run(pipelinePlugin(mode))` 或 `sdk.RunAdapter(&pipelineAdapter{})`。`service_chain_test.go` 已复用同一装配运行真实插件进程；参考条目为 goal-22（PR #164，合并提交 `039802f`）。
- `internal/tune/params.go`：名称为 `TOOL_CONFIRM_TIMEOUT`，字段为 `Values.ToolConfirmTimeout time.Duration`，缺省 `5*time.Minute`，覆盖约束为 `> 0`。`parse.go` 的 `Parse(args []string) (Values, []Override, error)` 接受 `NAME=value`、逗号分隔的覆盖或参数文件。`alonabot.go` 的 `Args(values ...string) Option` 经 `Start` 调用 `tune.Parse`；根包测试的 `assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error)` 直接接收生效值。
- 在途、可能与本切片改同一批文件的 PR：#170（goal-59，Router 与 SDK 的 InvocationContext）；goal-69 正在执行（根包与 cmd/alonabot）；goal-40 正在执行（根包装配、supervisor 与 state）。本切片共享 `pipeline_test.go` 与 `spec/TEST-MATRIX.md`，后合并者合并模式分流与追加行。

```text
/goal 使用 $alonabot-execute 推进 goal-57（切片 T-11.4：工具与会话的验收）。goal 目录 goal-57，分支 goal-57，worktree .worktrees/goal-57，从最新 origin/main 创建（须已包含 PR #101、#35、#165、#169）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：复用根包生产装配，以一个工具提供方插件、一个工具调用方插件和一个回环 Adapter 的真实进程，完成下列七项矩阵；验证工具目录与三种策略、确认的全部指定分支、策略的运行时修改、继承 Actor 的 Service 准入，以及会话的原子追加、分页、隔离、清除、Core 重启保留和服务层 purge；在 TEST-MATRIX §3.2 记录 T-11 证据。

依据：spec/PLAN.md §3.1；spec/plan/core.md T-11.4、T-11 步骤 6，以及步骤 1～5 的既有语义和「失败停止条件」；spec/CONTRACTS-v0.1.md §1.2（父调用与帧上限）、§2.3（权限名集合）、§3（Service 访问与调用链）、§5.4（确认文案）、§13.1～§13.3、§9.3（卸载清理归属）、§16（TOOL_CONFIRM_TIMEOUT）；spec/MESSAGE-MODEL.md §1（可信 Actor、默认 Target）、§2（buttons）、§3（Handler）、§5（Hook）、§6.3（token 与点击）；spec/WIRE-FORMAT.md §3.3（注册）、§3.5（固定服务）、§5A（追加字段）；spec/adr/DECISIONS.md D-07、D-12、D-69、D-86、D-102、D-105、D-114、D-127、D-137；spec/TEST-MATRIX.md V-55、V-56 的 T-11 分项与 §3 的证据格式。

范围（只新增或修改这些）：
- 新增根包 t11_e2e_test.go（//go:build unix，package alonabot）：一个顶层集成测试 TestToolsAndConversations，含下列七项子测试；测试二进制所需的工具插件、调用方插件及本切片专用回环 Adapter 行为也写在该文件。
- 根包 pipeline_test.go：只扩充 TestMain 的本切片子进程模式分流与必要的装配辅助（原始 wire/pb 的 provider 模式参照 scheduler_tasks_test.go 的 persistentTaskProcess；同一数据目录重建装配可复用 taskPipelineAt）。支持在 assemble 前传入测试 tune.Values、复用同一临时数据目录重建装配，以及记录确认的两枚 token、完整展示文本与工具执行审计；能力缩减、无 Actor 输入只作用于本切片模式。保留既有模式的行为与全部断言。
- 根包与 internal/core/tool/、internal/core/buttons/、internal/core/conversation/、sdk/ 中的 *_local_test.go：仅按本地自测段需要编写，不提交。
- spec/TEST-MATRIX.md §3.2：追加 T-11.4 对应 V-55、V-56 的证据行（本切片获授权修改这一处）。
允许的新跨模块边：无；根包继续通过既有的 internal/core、contracts、internal/wire、internal/tune 和 sdk 依赖装配测试。

接口（照此复用；本切片只增加测试与证据，公开符号和协议沿用 main；内部测试辅助函数自定）：
1. internal/core/tool
   func New(store *state.Store, r *router.Router) *Catalog
   func (c *Catalog) Register(m *router.Mux)
   func (c *Catalog) Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error)
   func (c *Catalog) SetPolicy(ctx context.Context, pluginID, service string, p pb.ToolPolicy) error
   func (c *Catalog) Policies(ctx context.Context) ([]ToolPolicyEntry, error)
   type ToolPolicyEntry struct {
       PluginID, Service string
       Default, Override pb.ToolPolicy
   }
   type InvokeOptions struct {
       Call func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response
       Adapters *adapter.Registry
       Buttons *buttons.Service
       ConfirmTimeout time.Duration
       DefaultLocale func() string
   }
   func (c *Catalog) RegisterInvoke(m *router.Mux, opt InvokeOptions)
   以上由 assembly.go 完成生产注册；测试通过插件 SDK 调用固定服务。运行时改策略由父测试进程持有 tool.New(h.store, h.assembly.router)，调用 SetPolicy；pluginID 为工具提供方，service 为 Service 的局部名。MCP 风格操作的策略键使用调用插件 ID 与完整操作名。覆盖取 pb.ToolPolicy_TOOL_POLICY_ALLOW/CONFIRM/DENY；UNSPECIFIED 删除覆盖。

2. internal/core/buttons
   func New(store *state.Store, now func() time.Time, ttl time.Duration, current func(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool)) *Service
   func (s *Service) Issue(ctx context.Context, caller router.Caller, target *pb.Target, chain *pb.MessageChain) error
   func (s *Service) Resolve(ctx context.Context, ev *pb.Event) (pluginID, routeID string, err error)
   func (s *Service) IssueConfirm(ctx context.Context, caller router.Caller, target *pb.Target, recordID string) (allow, deny string, err error)
   func (s *Service) InterceptConfirm(ctx context.Context, ev *pb.Event) bool
   func (s *Service) OnConfirm(fn func(ctx context.Context, recordID string, allow bool, ev *pb.Event))
   func (s *Service) RevokeConfirm(ctx context.Context, recordID string) error
   func (s *Service) Purge(ctx context.Context, pluginID string) error
   确认 token 由 Core 签发；测试从真实 Adapter 收到的 message.ButtonsSegment 中取得允许与拒绝 token，再经 Adapter Core.Deliver 注入可信 button.clicked。保留生产 Router 的 InterceptConfirm 分流。

3. internal/core/conversation 与 SDK 会话函数
   func New(store *state.Store) *Service
   func (s *Service) Register(m *router.Mux)
   func (s *Service) List(ctx context.Context, pluginID string) ([]state.ConversationSummary, error)
   func (s *Service) Clear(ctx context.Context, pluginID, key string) error
   func (s *Service) Purge(ctx context.Context, pluginID string) error
   // state.ConversationSummary 的字段：Key string、Entries int、LastSeq uint64、UpdatedAt time.Time。
   sdk 包中的签名：
   func AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error)
   func ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error)
   func ClearConversation(ctx context.Context, key string) error
   type ConversationEntry struct { Seq uint64; Entry json.RawMessage }
   插件调用的身份取自认证 Session；管理服务层查询、指定会话清除与 purge 由父测试进程调用 conversation.New(h.store) 的 List/Clear/Purge。

4. SDK 工具与登记（下列签名位于 sdk 包）
   func ListTools(ctx context.Context) ([]ToolInfo, error)
   func InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error)
   func ConfirmOperation(ctx context.Context, operation, summary string) error
   func NewPlugin() *Plugin
   func UseGenerated(p *Plugin, g Generated)
   func JSON[In, Out any](target string) Service[In, Out]
   func Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out]
   func (s *ServiceBuilder[In, Out]) Access(a Access, allowlist ...string) *ServiceBuilder[In, Out]
   func (s *ServiceBuilder[In, Out]) Tool(description string, defaultPolicy ToolDefault) *ServiceBuilder[In, Out]
   func (s *ServiceBuilder[In, Out]) Handle(h func(*Context, In) (Out, error))
   工具提供方参考 service_chain_test.go：p := sdk.NewPlugin()；用 sdk.UseGenerated 绑定输入、输出 ServiceSchemas；sdk.Serve(p, sdk.JSON[In, Out](providerID+"/run")).Access(sdk.AccessPublic).Tool("...", sdk.ToolConfirm).Handle(fn)。按测试场景设置缺省策略或经 Catalog 设置覆盖。ToolDefault 仅有 ToolAllow、ToolConfirm；deny 使用 Core 策略覆盖。
   调用方在 p.On(..., sdk.MessageCreated).Handle(func(c *sdk.Context) error { ... }) 中使用 c.Context() 调用 InvokeTool/ConfirmOperation；取结果、错误码和父链身份写入审计。自动回复若需要，使用现行 (message.Content, error) 签名。后台无 Actor 场景使用 p.Every(jobID, interval).Handle(func(c *sdk.Context) error { ... }) 的真实 scheduled 父调用，或真实 Service 根调用建立的无 Actor Handler。
   Adapter 使用 sdk/adapter 的 Adapter、Instance、Core.Deliver 与 sdk/message、sdk/errcode 类型；沿用 pipeline 的 sdkadapter.Action[pipelineInput, struct{}]("loop.inject", 1, sdkadapter.ActionScopeChat) 声明。需要新增本切片输入类型时相应声明真实的类型化 Action。权限为 Actor.Permissions 中的名称集合，Core Admin 按 settings 中的 <adapter_id>:<external_user_id> 取得。若测试采用 HTTP 输入，Adapter 的可选接口为 sdk.AdapterHTTP：HTTP(route func(method, path string) *sdk.RouteBuilder)，经 RouteBuilder 登记；首选复用已有 Action 输入。

5. 既有协议（均为 alonabot.v1，PROTOBUF 载荷，schema_version=1）：
   ServiceSpec.tool = 5：ToolSpec{description=1, default_policy=2, input_schema_json=3}。
   ToolPolicy：UNSPECIFIED=0，ALLOW=1，CONFIRM=2，DENY=3。
   ToolInfo：name=1、description=2、input_schema_json=3、policy=4；ToolListRequest 为空；ToolListResponse.tools=1（repeated）。
   ToolInvokeRequest：name=1、arguments_json=2；ToolInvokeResponse.result_json=1。
   ToolConfirmRequest：operation=1、summary=2；ToolConfirmResponse 为空。
   ConversationAppendRequest：conversation_key=1、entries_json=2（repeated）；ConversationAppendResponse.last_seq=1。
   ConversationListRequest：conversation_key=1、from_seq=2；ConversationEntry：seq=1、entry_json=2。
   ConversationListResponse：entries=1（repeated）、next_seq=2（optional uint64，absent 表示读完）。
   ConversationClearRequest.conversation_key=1；ConversationClearResponse 为空。
   工具 invoke/confirm 的 parent_call_id 由 SDK 的活跃 context 关联；会话服务无父调用要求。RouteDescriptor.required_permissions=17（repeated string），准入取继承 Actor 的权限名集合。

6. 根包测试装配与 tune
   func newPipeline(t *testing.T, stdio ...bool) *pipelineHarness
   func assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error)
   func Args(values ...string) Option
   internal/tune 包中的签名：
   func Defaults() Values
   func Parse(args []string) (Values, []Override, error)
   常规确认保留默认期限；超时子场景可经 tune.Parse([]string{"TOOL_CONFIRM_TIMEOUT=2s"}) 取得测试值，并在 assemble 之前传入。父请求的 deadline 应晚于确认期限，以收到确认服务产生的 DEADLINE_EXCEEDED。现有 newPipeline 内部直接取 Defaults，必要时提取可传值的未导出装配辅助并同步本文件调用处；成功路径继续使用正常期限。
   通过 Start 启动 Core 的测试可用 Args("TOOL_CONFIRM_TIMEOUT=2s")；CLI 对应 --args TOOL_CONFIRM_TIMEOUT=2s。ALONABOT_ARGS 由 Start 读取，现有 newPipeline 直接调用 assemble，因此本矩阵在构造装配时显式传 tune.Values。

行为：
- 测试二进制按环境变量扮演 provider、caller 与回环 Adapter，复用 newPipeline、install、launch、生产 handler/attach 与真实 SQLite/Session。本切片的 Adapter 子模式继续使用 role adapter、adapter_id loop，并由真实 StartInstance 启动实例。提交的矩阵默认 TCP；stdio 复用现有 stdio 选项做本地自测。所有数据和插件进程目录使用测试临时目录。
- 审计覆盖提供方每次实际执行、调用方的结束结果、Adapter 的实际发送与编辑，以及两枚确认 token。用明确的调用完成、Router 排空和审计屏障断言拒绝后执行计数不变；时间到期依据确认结果判断，沿用现有测试等待期限。
- 验收矩阵（按以下七项组织子测试，保留每项原有含义）：
  1) allow、confirm、deny：相同工具参数经 sdk.InvokeTool 走真实服务。allow 直接得到 JSON 输出；confirm 在同一用户允许后执行一次；deny 返回 PERMISSION_DENIED 且调用计数不变。sdk.ListTools 只列出访问允许且生效策略非 deny 的工具，核对名称、说明、输入 Schema 与生效策略。拒绝确认同样返回 PERMISSION_DENIED，计数不变。
  2) 确认的允许、拒绝、超时、他人先点、重复点击：核对确认文本含工具名和实际 JSON 参数（按 JSON 值核对，允许展示缩进）；他人点击后本次调用仍等待，原 Actor 允许后只执行一次；拒绝和超时后均未执行，超时为 DEADLINE_EXCEEDED。已完成记录的原 token 再点无额外执行，另一枚 token 同样失效。确认发送直接到 Adapter，调用方 BeforeSend 审计保持为空；支持编辑时核对允许/拒绝后的结果文本与按钮撤除。使用 Core 现有 locale 文案与真实 token 绑定。
  3) 无 Actor 与无按钮能力：在真实 scheduled 或 Service 根 Handler 内调用 confirm 策略工具，保留活跃父调用且缺少可信 Actor，得到 PERMISSION_DENIED，Adapter 没收到确认消息。另在具有 Actor 和默认 Target、但该测试实例缺少 message.buttons 能力时调用，得到 UNSUPPORTED，工具未执行。两种场景分别验证。
  4) MCP 风格操作：调用方在活跃 Handler 中执行 sdk.ConfirmOperation(c.Context(), "mcp/x/y", summary)，核对 Core 展示完整操作名及插件提供的摘要；同 Actor 允许返回 nil，随后由调用方记录一次操作。Core 的本次确认按调用插件 ID 与操作名识别；确认前该操作审计为空。
  5) 运行时修改策略：父测试进程用 tool.New(h.store, h.assembly.router).SetPolicy 修改已安装工具的覆盖；同一 provider/caller 进程上再次 InvokeTool，allow/confirm/deny 的下一次调用按新策略运行，PID 和 Session 身份保持不变。配合 Effective/Policies 核对服务层已保存的覆盖，SDK 目录同步反映 deny 的移除；删除覆盖后回到声明缺省策略。修改经 Core 内部服务层驱动。
  6) Service 访问与继承 Actor：provider 的 Service 访问规则使用 sdk.AccessPublic 或 sdk.AccessAllowlist，并另设当前 caller 无访问权的工具；目录按调用方过滤，直接 InvokeTool 对不可见工具返回当前实现的 NOT_FOUND，提供方计数不变。允许访问的调用中核对提供方收到原 Actor 与默认会话；带 required_permissions 的工具必须覆盖权限不足 PERMISSION_DENIED 与具备权限成功，策略 allow 或用户允许均保留该权限检查。required_permissions 分项由测试二进制的原始 wire/pb provider 模式驱动：参照 scheduler_tasks_test.go 的 persistentTaskProcess，以既有协议完成握手与 REGISTER（JSON Service 带 ToolSpec 与 RouteDescriptor.required_permissions），并应答工具调用；SDK 的 Service 权限声明入口不在本切片。
  7) 会话：provider 与 caller 均经 SDK 操作自己的同名会话。一次 AppendConversation 追加多条，随后再追加，返回连续的末尾序号；含无效 JSON 对象的一次批量追加整次失败，原条目与下一序号不变。多次合法追加的总量超过 4 MiB，单次请求保持帧内；可用六条约 1 MiB 的对象分批追加，ListConversation 从 0 起按 more/nextSeq 读完，至少两页，序号、顺序、内容原文不丢不重。两插件同键读取各自内容，ClearConversation 只清自己的会话。父测试进程经 conversation.Service.List/Clear 核对管理服务层的摘要与指定会话清除。停止全部旧子进程和装配、关闭 Store 后，用同一数据目录重新 Open、assemble 并 Launch，插件再次 ListConversation 与重启前一致。最后经 conversation.Service.Purge 删除一个插件的全部会话，另一个插件的内容仍保留；本切片覆盖服务层 purge，三种卸载流程沿用 T-09.4 的证据。
- T-11 的失败停止条件：未经确认执行、接受其他 Actor 的决定、无 Actor 通过确认、绕过 Service 访问或 required_permissions，以及会话丢失、重排、跨插件可见，均使验收失败。保留错误与原始数据，按实际失败位置记录；需要改范围外的生产代码、协议或契约时依共同约束标 [!] 并交审查者处理。
- 证据：TEST-MATRIX §3.2 追加 T-11.4 一行，关联 V-55、V-56，写明提交的测试与子场景、实际命令、OS/arch、Go 版本、传输和结果；本地自测另列。管理 API（T-09.14）、管理页（T-09.9）、Agent 工具循环（T-12.3）、未运行的真实平台与系统必须保留真实状态。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：根包 t11_e2e_test.go 的 TestToolsAndConversations，按上述七项组织子测试。多列理由：七项是 T-11.4 明确要求的跨进程验收矩阵，分别涉及工具策略与确认、可信调用链、状态持久化；共用同一装配。pipeline_test.go 仅纳入支持该矩阵的必要改动。
- 本地自测（范围所列目录的 *_local_test.go，不提交，PR 中写明命令和结果）：以 stdio 运行同一矩阵；在 -race 下 -count=5 检查偶发失败；父调用取消或 Session 失效后确认 token 作废；参数与摘要展示的 UTF-8/大小边界；MCP 操作 allow/deny 覆盖与相同操作名的插件隔离；会话并发追加和异常分页边界。SDK 客户端边界验证按 D-86 放 sdk/ 的本地测试。

完成条件：七项矩阵含 required_permissions 分项均取得证据；gofmt -l 对本次 Go 文件无输出；go vet ./...；go test -race ./...；go build ./...；TEST-MATRIX §3.2 写入实际证据和未运行项；既有 pipeline 与 service_chain 的断言保持有效。有失败停止条件命中时保留 [!]。

不做：Agent 对话与工具循环（T-12.2、T-12.3）；工具策略或会话的管理 API（T-09.14）与页面（T-09.9）；新增公开 SDK 权限 API；生产实现重构；协议字段与生成物、状态库 schema、manage 清理清单；新增依赖；向 sdk/、internal/core/supervisor/ 提交测试或新增 testdata 测试插件；改写 service_chain_test.go 的既有断言。
```

### goal-102 · T-16.3 plugintest 的 LLM 替身

状态：已完成（[PR #190](https://github.com/AlonaBot/AlonaBot/pull/190)，合并提交 `2ddaabd`），证据见 TEST-MATRIX §3.2。合入：`plugintest.LLM(handler func(ctx context.Context, req llm.Request) (llm.Response, error)) Option` 登记只提供 `llm.generate` 的测试 Agent（`plugintest/internal/llmagent`），被测插件经 `@agent` 的调用只做共享 Schema 的结构校验后交给测试函数，多次给出时最后一个生效，nil 函数在公开入口报错。以下为派发时的语句。

按 main 核对：

- `sdk/llm/llm.go`：`const AgentTarget = "@agent"`、`ServiceName = "llm.generate"`、`SchemaIn = "github.com/AlonaBot/AlonaBot/sdk/llm.Request"`、`SchemaOut = "github.com/AlonaBot/AlonaBot/sdk/llm.Response"`、`SchemaVersion uint32 = 1`；`var GenerateService = sdk.JSON[Request, Response](AgentTarget + "/" + ServiceName)`；`Generate(ctx context.Context, req Request) (Response, error)`、`Ask(ctx context.Context, prompt string) (string, error)`。`Response.InputTokens` 与 `OutputTokens` 均为 `*int`，保留缺失与零值的区别。
- `sdk/services.go`：`JSON[In, Out any](target string) Service[In, Out]`、`Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out]`；`(*ServiceBuilder[In, Out]).Access(a Access, allowlist ...string) *ServiceBuilder[In, Out]`、`Handle(h func(*Context, In) (Out, error))`；`(Service[In, Out]).Call(ctx context.Context, in In) (Out, error)`。真实服务调用在调用方和提供方都读取生成的 JSON Schema。现有公开 builder 方法只有 Access、Tool、Handle；D-142 的 `Permission` 将由 goal-61 加入。
- `sdk/generated.go`：`UseGenerated(p *Plugin, g Generated)`、`(*Plugin).Generated() Generated`；`Generated{Manifest []byte; AdapterConfigSchema []byte; ServiceSchemas map[string][]byte; I18n map[string][]byte; ParamNames map[string][]string; ConfigSchema []byte}`。`sdk/sdk.go` 的 `(*Context).Context() context.Context` 提供本次关联上下文；`sdk/services.go` 的 `(*Context).Caller() string` 读取 Core 给出的直接调用方。
- 根包 `alonabot.go`：`Plugin(root *sdk.Plugin) Option` 已接受 role plugin 或 agent；`Start(ctx context.Context, opts ...Option) (*Core, error)`；`assembly.go` 已向 Broker 注入 Agent 定位，`agent_target.go` 的 `(*assembly).agentID(ctx context.Context) (string, error)` 从 state 的安装记录选择未卸载的 `RoleAgent`。`internal/core/broker/broker.go` 的 Broker 处理 @agent 后以真实插件 ID 执行激活、访问、调用链、schema 与取消检查。
- `agent/agent.go` 的 `New() *sdk.Plugin` 已用 `sdk.Serve(p, llm.GenerateService).Access(sdk.AccessPublic).Handle(a.generate)` 注册真实服务；`agent/generate.go` 的 `(*component).generate(c *sdk.Context, req llm.Request) (llm.Response, error)` 按直接 Caller 检查默认 `["*"]` 的配置名单。测试 Agent 只需实现本切片的生成服务及默认放行语义。
- `plugintest/plugintest.go`、`options.go`：`Start(t testing.TB, opts ...Option) *Harness`、`Plugin(p *sdk.Plugin) Option`、`Config(p *sdk.Plugin, configJSON string) Option`；`(*Harness).Text(chat Chat, from User, text string) Result`，`Result{Sends []Sent; Outcomes []Outcome; Waiting bool}`，`Outcome{PluginID, RouteID string; Err *errcode.Error}`。现有 options 没有 LLM；Start 统一给每个开发组件绑定临时目录与真实生命周期，清理取消 ctx 并 Wait。
- `plugintest/internal/probe/probe.go` 的 `New() *sdk.Plugin` 是现有集成插件入口，已有生成资源与配置/SQLite、命令和等待场景，可在此追加调用 llm.Generate/Ask 的命令。`sdk/errcode` 的 `Of(err error) Code` 与 `(Code).New(message string) *Error` 用于返回与断言协议错误。
- `pipeline_test.go` 已有 `newPipeline(t *testing.T, stdio ...bool) *pipelineHarness` 及 TestMain 子进程分流；根包 `agent_service_test.go` 已用真实 SDK Session、Core、官方 Agent 与 httptest Provider 验证文本生成。多进程本地取证沿用这些装配，正式提交的套件测试直接通过 plugintest 的公开入口。
- 旧在途清单中的 #165（`330164f`）、#167（`3a9b4af`）、#168（`37b0ada`）、#169（`4dc4bcd`）、goal-59（#170，`eca6310`）和 goal-69（#171，`3043c4e`）均已在本地完成记录中；#169 的 schema.sql/manage 清理以及 #171 的根包/cmd 入口已在基线。相关 #166（`6f72fc8`）、#173（`2bb8181`）、#176（`9785c22`）已合并；#174 的已合并状态按维护者提供的信息处理。
- 当前在途按维护者清单：#177（goal-66）、goal-41 改根包装配与生命周期，但本切片无需修改它们的产品文件；#178（goal-60）改官方 agent 与其生成资源，本切片的替身放在 plugintest 内，不复用官方 Agent 业务。goal-9 与本切片共享 `docs/development.md`、`spec/TEST-MATRIX.md`。goal-82 直接共享 `plugintest/{options,plugintest}.go`、probe 与其生成物，建议 goal-82 先合并，再按新 options/Outcome 接入本切片。goal-61 将改 `sdk/services.go`，本切片仅消费现有 Access/Handle，保持其后续 Permission 的所有权。未读取远端 PR diff。

```text
/goal 使用 $alonabot-execute 推进 goal-102（切片 T-16.3：plugintest 的 LLM 替身）。goal 目录 goal-102，分支 goal-102，worktree .worktrees/goal-102，从最新 origin/main 创建（包含 PR #168、#175；若 goal-82 已合并，读取其 options、Outcome 与 probe 后接续）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现 plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error)) Option。Start 登记一个只提供 llm.generate 的测试 Agent，被测插件通过 llm.Generate、llm.Ask 或共享 GenerateService 经 @agent 的真实 Service 调用交给测试函数；保留完整请求、响应、错误码与取消。未给 LLM 选项时 @agent 沿现有目标未激活路径失败。

依据：spec/PLAN.md §3.1、§5；spec/plan/core.md T-16.3、T-16 步骤 3 的 LLM 替身项及「失败停止条件」；spec/CONTRACTS-v0.1.md §1.2、§2.3、§3、§4.1～§4.2、§13.5、§15.5、§17；spec/MESSAGE-MODEL.md §1、§3～§5（触发插件的真实事件、回复与 Hook）；spec/WIRE-FORMAT.md §3.3 的 Service 描述、§4 的可信上下文与取消、§5/§5A 的 ServiceInvocation；spec/adr/DECISIONS.md D-47、D-51、D-61、D-70、D-77、D-83、D-86、D-111、D-122、D-134、D-137、D-142 的 API 所有权；spec/TEST-MATRIX.md V-73 的 LLM 分项与 §3，沿用 V-76 的 Broker 及 V-109 的 sdk/llm 既有行为。

范围（只新增或修改这些）：
- plugintest/options.go：增加 LLM 选项及其测试函数存储；plugintest/plugintest.go：Start 将测试 Agent 作为开发组件登记，沿既有清理路径结束。
- 新增 plugintest/llm.go：选项接入所需的私有组装代码；新增 plugintest/internal/llmagent/：测试 Agent 的内部构造、plugin.yaml、生成的 alonabot_gen.go；只提供本切片指定的服务。生成资源按现有入口生成。
- plugintest/internal/probe/probe.go 与其生成资源：追加触发 llm.Generate/Ask 的测试命令，保留已有 probe 路由与资源。
- plugintest/llm_test.go：以下一个公开入口集成测试；上述包的 *_local_test.go 与根包 *_local_test.go：本地自测，不提交。
- docs/development.md「插件测试」一节：追加 LLM 替身示例；spec/TEST-MATRIX.md §3.2：只追加本切片证据行。

允许的新跨模块边：无。沿用 plugintest -> sdk/...、根包、internal/inproc、contracts；内部 llmagent 属于套件实现，使用 sdk 与 sdk/llm。测试 Agent 不链接 agent/、agent/provider/ 或外部 Provider 客户端；依赖方向按 ARCHITECTURE。

接口（照此实现；唯一新增的面向作者接口为 LLM）：
1. plugintest 新增：
   func LLM(handler func(ctx context.Context, req llm.Request) (llm.Response, error)) Option
   Start(t testing.TB, opts ...Option) *Harness、Plugin(p *sdk.Plugin) Option 及 Result/Outcome 保持合并后 main 的签名。
2. 测试 Agent 使用现有声明注册：
   sdk.Serve(p, llm.GenerateService).Access(sdk.AccessPublic).Handle(func(c *sdk.Context, req llm.Request) (llm.Response, error) {
       return handler(c.Context(), req)
   })
   内部构造函数与组件 ID 由实现自定，工程 manifest 明确 role: agent；Start 用既有 alonabot.Plugin 将其纳入同一临时 Core。组件身份和 Schema 均由源资源生成并绑定。
3. 共享输入输出，逐字段保持 sdk/llm/llm.go：
   type Message struct { Role string `json:"role"`; Text string `json:"text"` }
   type Request struct {
       Messages []Message `json:"messages"`
       System string `json:"system,omitempty"`
       Provider string `json:"provider,omitempty"`
       MaxOutputTokens int `json:"max_output_tokens,omitempty"`
   }
   type Response struct {
       Text string `json:"text"`
       Provider string `json:"provider"`
       Model string `json:"model"`
       InputTokens *int `json:"input_tokens,omitempty"`
       OutputTokens *int `json:"output_tokens,omitempty"`
   }
   以上为引用现有类型的说明，代码直接使用 llm.Request/Response/Message。
4. 服务目标 @agent/llm.generate，JSON 编码；schema_id 使用 llm.SchemaIn、SchemaOut，version 使用 llm.SchemaVersion=1。内部路由沿 sdk.Serve 生成 svc.llm.generate（sdk/services.go 的 serviceRoutePrefix="svc."）；请求由真实 ServiceInvocation 携带 route_id、可信 context、encoding、schema_id、schema_version、payload，响应为现有输出 Schema 的业务载荷。沿用已合并字段及编号，协议文件不变。

行为：
- LLM 选项只登记函数。Start 在启动前构造一份测试 Agent，与被测插件及测试 Adapter 一同完成 HELLO、配置准备、REGISTER、ACTIVATE、ACTIVATED；Start 成功后才接受注入。Agent 使用该 Harness 的临时数据目录、开发代次和既有 ctx/Wait 清理，不启动独立 Core。
- 每份 Harness 登记一份测试 Agent，沿用根包已有的重复组件/Agent 身份检查；LLM(nil) 作为公开选项的无效输入在启动前使测试明确失败。多个 LLM 选项与现有 Config/Capabilities 的覆盖方式一致，最后一个函数生效。每个 Harness 保存自己的函数，平行套件互不覆盖。
- @agent 继续由现有 Core Agent 定位解析；权限、激活门、调用链、parent_call_id、期限与取消、Schema ID/version/encoding、SDK 输入输出 Schema 校验均走真实 Broker/Session。提供方 AccessPublic 实现 llm_service_allowed_plugins 默认 ["*"] 的放行语义，无需测试作者再授权。
- 函数收到 c.Context() 及解码后的完整 llm.Request，调用一次；返回的文本、Provider、Model、nil/零值 token 用量原样经 SDK 编码。请求的语义结果由替身函数给出。错误沿已有 SDK 映射：*errcode.Error 保留 Code 与英文 SafeMessage，普通 Go error/panic 按 SDK 的 INTERNAL 处理；保留失败，不伪造空成功、不自动重试。
- parent 取消与期限照常传播给替身 ctx；测试函数按 ctx 协作结束，结束行为沿既有 Service 生命周期。调用与后续回复仍遵循真实出站 Hook 和 Target 检查。
- 未配置 LLM 时，套件保持当前组件集合，@agent 调用得到现有 UNAVAILABLE；注入正常结束并将调用错误作为 Handler Outcome 暴露。llm.Ask 继续保留 SDK 的非空白 prompt 检查，并将原始 prompt 作为唯一 user 消息。
- 测试 Agent 只登记 llm.generate，无官方对话、会话存储、人格、工具循环、llm.providers 或 Provider 请求。生成的 ServiceSchemas 同时供调用方 probe 和提供方使用，均从相应 Go 源的带类型声明生成，生成代码只经生成器更新。
- 文档给出一个 Plugin(p)+LLM(handler)+Text 的可运行测试例子：按 req.Messages 返回固定 Response，断言被测插件发出的文本；另用简短例子展示 errcode.Unavailable 和保留 nil 用量。读者运行 go test -race -count=2 ./plugintest/...。

测试（按编码规范「测试提交规则」，D-86）：
- 提交：plugintest/llm_test.go 的一个 TestLLMHarness。通过 plugintest.Start + 已有 probe 的新命令驱动真实 Core、SDK Session 和 Broker，在同一测试的子场景中验证：
  1) llm.Generate 经 @agent 到达替身一次，函数收到 Messages、System、Provider、MaxOutputTokens，返回的完整 Response 被 probe 核对后发送 Text；用量区分 nil 与指向 0。llm.Ask 到达同一服务，函数收到唯一 user 消息与原始 prompt，插件回复替身文本。
  2) 替身返回 errcode.Unavailable.New("synthetic generation failure")，调用插件将错误返回给 Handler；Result.Outcomes 中保留 UNAVAILABLE 与 SafeMessage，Result.Sends 为空，调用次数为一。
  3) 另一 Harness 不提供 LLM 选项，调用结果是 UNAVAILABLE，注入正常结束并留下 Handler 错误；组件关闭后清理沿 Start 的既有路径。
  这是一个套件跨组件集成测试，直接守住选项→开发 Agent→真实 @agent/Broker→共享 Schema→调用插件回复/错误的主路径。sdk/、internal/core/supervisor/ 与 testdata 均不新增提交测试。
- 本地自测（*_local_test.go，PR 记录实际命令、平台和结果）：替身 context 的直接调用方/父链、deadline 与取消；Schema/版本/编码拒绝且函数未执行、输出 Schema 失败；一般 error/panic 的 INTERNAL 映射；nil handler、选项覆盖、Agent 身份冲突；平行 Harness 函数隔离与多次调用；停止或启动失败后组件、锁及 goroutine 清理；缺省名单允许不同插件。需要跨进程对照时复用根包 pipeline_test.go 的 newPipeline/TestMain，在临时目录中写子进程入口或按既有模式分流，分别记录 TCP/stdio。
- 证据：TEST-MATRIX §3.2 追加 T-16.3/V-73 LLM 一行，写测试位置、命令、平台、Go 版本、结果、未运行项；保留 T-16.1/T-16.2 已有证据。

完成条件：改动 Go 文件 gofmt -l 无输出；通过已有生成入口重建内部测试 Agent 与 probe 的生成资源，go generate ./... 后重复生成内容一致；go vet ./...；go test -race -count=2 ./plugintest/...；go test -race .；go build ./...。逐项核对 T-16 失败停止条件：真实 Core 检查和派发完整、注入依完成通知返回、依赖方向符合 ARCHITECTURE；本地自测与 V-73 LLM 证据记录完整后结束。

不做：真实 Provider 请求、官方 Agent 的对话/工具/记忆/选项来源；跳过 Broker 的函数直调替身路径；可配置调用方名单、额外的公开 Agent 工厂或 Schema 导出；SDK 的新公开符号、ServiceBuilder.Permission（由 goal-61 提供）；根包启动/Journal/日志管理改造；goal-82 的按钮与调度；新 testdata 插件、依赖、wire 消息或状态表；范围外 spec、PLAN、ADR 与 ARCHITECTURE 修改。
```

## 第 16 波

### goal-138 · T-04.15 SDK 公开面收敛

状态：已完成（[PR #197](https://github.com/AlonaBot/AlonaBot/pull/197)，合并提交 `ba2d323`），证据见 TEST-MATRIX §3.2。合入：`sdk/message` 的 `SID`/`ParseSID`、`Ref`/`ParseRef`、`UID`/`ParseUID`，`c.SendTo(sid, content, opts...)`，`sdk.Send(ctx, sid, …)`、`sdk.SendText(ctx, sid, text, …)` 与 string 句柄的 Edit、Delete、CallAction、Session、User、Editing；`internal/generated`（`Register`、`Bind`、`Lookup`、`SetEntry`、`Entry`）与 `sdk/genimpl`（`Resources`、`Register`），`sdk.NewPlugin()` 记录入口包（还原符号名的 `%xx` 转义），资源在 Collect 与启动组件时取得；`internal/generate`（`Options.Package`、`Result.Generated`、登记源码）与 `pluginkit.CompileApp(ctx, options, outDir, goFlags...)`（库模式构建，不加 `-trimpath`），`alonabot build --dir` 在没有 plugin.yaml 的 main 包目录按库模式构建；plugintest 按入口包现场扫描；`adapter.Descriptor.HTTP`、`adapter.HTTPRoute`、`adapter.HTTPAuth`、`adapter.AuthPlatformToken`。删除 `sdk.To`、三种具名类型、`sdk` 的生成 API、`sdk.AdapterHTTP`、`sdk/generate`、仓库中的 `alonabot_gen.go` 与 alonabot 的 `go:generate`。审查追加的修复：入口包路径还原符号转义（FIX-1）、`Send`/`SendText` 参数名为 `sid`（FIX-2）、库模式构建不强加 `-trimpath`（FIX-3），以及两处测试复用 race 构建缓存。

```text
/goal 使用 $alonabot-execute 推进 goal-138（切片 T-04.15：SDK 公开面收敛）。goal 目录 goal-138，分支 goal-138，worktree .worktrees/goal-138，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：会话、消息句柄与用户字符串在 SDK 中一律是 string，发往当前事件以外的会话写 c.SendTo(sid, content)；作者工程与仓库中不再有生成文件，作者代码不写任何生成相关的内容，资源由 CLI 构建经 overlay 编入、由 plugintest 在 go test 中现场扫描，按入口包登记；库模式改由 alonabot build 构建；Adapter 的 HTTP 路由写进 adapter.Descriptor；示例只在 New() 中注册。协议、Core 行为与管理 API 不变。

依据：PLAN T-04 步骤 15 及其失败停止条件；CONTRACTS §7.2、§8、§9.1、§9.3、§15.5、§17；MESSAGE-MODEL §1、§4；DESIGN §D、§E；DECISIONS D-86、D-102、D-104、D-105、D-114、D-118、D-120、D-136、D-137、D-144、D-145、D-147；TEST-MATRIX V-112、V-120；ARCHITECTURE 中 sdk 与 internal/generate 的拟建说明。

范围（只新增、修改或删除这些）：
- sdk/message/：session.go 改为 SID/ParseSID、Ref/ParseRef、UID/ParseUID；event.go、result.go、delete.go、media.go、message.go 中以 Session、MessageRef、User 声明的字段与方法改为 string；包文档同步
- sdk/：sdk.go、send.go、edit.go、delete.go、action.go、hooks.go 的发送与句柄 API；plugin.go 记录入口包；generated.go 删除，资源改由内部登记表读取；registry.go、bus.go、services.go、i18n.go、session.go、subplugin.go、inproc.go 中读取资源的位置；adapter_run.go 与 http.go 的 Adapter 路由；sdk/internal/convert/ 的句柄转换
- sdk/adapter/：declaration.go 新增 HTTPRoute、HTTPAuth、AuthPlatformToken 与 Descriptor.HTTP；adapter.go 的包文档
- 新增 sdk/genimpl/ 与 internal/generated/
- sdk/generate/ 整体移到 internal/generate/（用 git mv 保留历史），以及调用方 cmd/alonabot/{build,check,pack,plugin,test}.go 与 internal/pluginkit/{build,capabilities,check,pack,init}.go；pluginkit 新增库模式构建
- 根包 assembly.go 的库模式资源读取；根目录 *_test.go 中的资源夹具、会话写法与组件构建方式
- plugintest/：plugintest.go、adapter.go、buttons.go，plugintest/internal/{probe,llmagent}/ 及其测试
- agent/、adapters/telegram/、adapters/onebot11/、examples/ 中使用上述 API 的源码与测试
- 删除 agent/、adapters/telegram/、examples/、plugintest/internal/ 下的 11 个 alonabot_gen.go 与全部 `//go:generate go run …/cmd/alonabot build` 指令；contracts/doc.go 的 Protobuf 指令不动
- cmd/alonabot/*_test.go、internal/core/adapter/remote/remote_test.go、internal/core/supervisor/supervisor_test.go 中随 API、资源夹具与组件构建方式改写的部分，以及下文的提交测试
- docs/plugin-guide.md、docs/development.md 的写法
- 上述包的 *_local_test.go
允许的新跨模块边：sdk、sdk/genimpl、根包、plugintest、internal/generate -> internal/generated；plugintest -> internal/generate；cmd/alonabot、internal/pluginkit -> internal/generate（取代 -> sdk/generate）；本 module 的测试代码 -> internal/generated。internal/generated 只 import 标准库。

接口（照此实现）：
sdk/message
   func SID(adapterID, chatID, threadID string) string
   func ParseSID(sid string) (adapterID, chatID, threadID string, err error)
   func Ref(sid, platformMessageID, instance string) string
   func ParseRef(ref string) (sid, platformMessageID, instance string, err error)
   func UID(adapterID, externalUserID string) string
   func ParseUID(uid string) (adapterID, externalUserID string, err error)
   // 转义与格式规则沿用现有 NewSession/NewMessageRef/NewUser 与 Parts；Parse* 在格式不合法时返回 code 为 INVALID_ARGUMENT 的 *errcode.Error（sdk/errcode）。
   // 删除 Session、MessageRef、User、NewSession、NewMessageRef、NewUser 与各自的 Parts，不留别名。
   // 以下成员只改类型为 string，名字不变：Event.Session、Event.Ref()、SendResult.Session、SendResult.Ref()、SendResult.Refs()（[]string）、ReferenceSegment.Ref、Quote(ref string)、DeleteResult.Session、media.go 中的 Sessions（[]string），以及其他以这三种类型声明的字段。
sdk
   func (c *Context) SendTo(sid string, m message.Content, opts ...SendOption) (*message.SendResult, error)
   func (c *Context) Send(m message.Content, opts ...SendOption) (*message.SendResult, error) // 只发往当前事件的默认目标
   func Send(ctx context.Context, sid string, m message.Content, opts ...SendOption) (*message.SendResult, error)
   func SendText(ctx context.Context, sid, text string, opts ...SendOption) error
   func Via(instanceID string) SendOption // 唯一的 SendOption
   func (c *Context) Session() string
   func (c *Context) User() string
   func (c *Context) Editing() string
   func (c *Context) CallAction(sid, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error)
   func (c *Context) Edit(ref string, m message.Content) (*message.SendResult, error)
   func Edit(ctx context.Context, ref string, m message.Content) (*message.SendResult, error)
   func (c *Context) Delete(refs ...string) (*message.DeleteResult, error)
   func Delete(ctx context.Context, refs ...string) (*message.DeleteResult, error)
   // 删除 To、Generated、UseGenerated、UseGeneratedAdapter、GeneratedFor、GeneratedProvider、(*Plugin).Generated 与 AdapterHTTP，不留别名。UserID、NewPlugin 与其余公开签名不变。
sdk/adapter
   type HTTPAuth int
   const AuthPlatformToken HTTPAuth = 1 // 平台 token 公开路由；零值要求 Core 会话
   type HTTPRoute struct {
       Method  string
       Path    string
       Auth    HTTPAuth
       Handler http.Handler
   }
   // Descriptor 新增字段 HTTP []HTTPRoute，其余字段不变。
sdk/genimpl（新包；包文档写明只供 CLI 注入的登记代码使用，不对作者承诺兼容）
   type Resources = generated.Resources
   func Register(pkgPath string, r Resources)
internal/generated（新包）
   type Resources struct {
       Manifest            []byte              // 工程入口目录 plugin.yaml 的原文
       ConfigSchema        []byte              // config_schema.json 的规范 JSON；没有该文件时为 nil
       AdapterConfigSchema []byte              // Adapter 实例的规范 JSON Schema；插件为 nil
       ServiceSchemas      map[string][]byte   // schema_id 对应的规范 JSON Service Schema
       I18n                map[string][]byte   // i18n 文件名对应的原文
       ParamNames          map[string][]string // 键与现有 Generated.ParamNames 相同
   }
   func Register(pkgPath string, r Resources)                   // 同一导入路径登记两次时 panic
   func Bind(component any, r Resources)                        // 按对象绑定，只供仓库测试夹具与 plugintest 的测试 Adapter
   func Lookup(component any, pkgPath string) (Resources, bool) // 先查对象绑定，再查入口包
   func SetEntry(p any, pkgPath string)                         // sdk.NewPlugin 记录插件对象的入口包
   func Entry(p any) (pkgPath string, ok bool)
   // 全部函数并发安全。
internal/generate（由 sdk/generate 移入）
   Options 增加 Package string // 要扫描的组件入口包导入路径；为空时扫描 Dir 本身。非空时按 Dir 所在 module 的构建列表解析，可以是模块缓存中的第三方包，只读。
   Result 增加 Generated generated.Resources // 本次扫描得到的全部资源
   // Result.Source 改为调用 genimpl.Register 的登记代码；删除 FileName。Scan 与其余字段、诊断规则保持。
internal/pluginkit
   func CompileApp(ctx context.Context, options generate.Options, outDir string, goFlags ...string) (executable string, err error)
   // 库模式构建；goFlags 只供仓库测试传入（例如 -race），CLI 不暴露。BuildDev、Compile、Pack、Check、Init 的签名不变。

行为：
- 实施次序：先句柄与 Adapter 路由；再资源登记，期间把现有 alonabot_gen.go 按新格式重新生成以保持编译；再 plugintest 现场扫描与库模式构建；最后删除生成文件，改 CLI、示例与文档。每一步结束时完成条件都通过。
- 句柄：SID、Ref、UID 只拼接与转义，不校验；ParseSID、ParseRef、ParseUID 的校验与现有 Parts 相同（adapter_id 规则、聊天非空、话题不含 `#`、实例字符集、external_user_id 非空且原样保留其中的 `:`），英文错误指明是哪一种字符串。sdk/internal/convert 与 Adapter SDK 的 Target、Event 映射改用这些函数，Wire 字段不变。
- 发送：c.SendTo 发往给出的 sid，可带 Via；c.Send 只发往当前事件的默认目标，没有时照旧返回 NO_DEFAULT_TARGET；sdk.Send、sdk.SendText 收 sid。发送、编辑、撤回与 CallAction 收到不合法的 sid 或句柄时，在发出请求前返回 INVALID_ARGUMENT。c.Session()、c.User() 没有默认目标或 Actor 时为空串。Agent 的任务目标保留现有 JSON 字段名与字符串内容，只把类型改为 string；Telegram、OneBot 11 Adapter 中的 Parts 调用改用 Parse*。
- 入口包：sdk.NewPlugin() 用 runtime.Caller 记录调用方所在的包，规则与现有 UseGenerated 相同（普通二进制中 package main 记为 `main`，go test 中记为导入路径），经 generated.SetEntry 保存。注册期间不读取资源。
- 取得资源：Collect、sdk.Run、sdk.RunAdapter、RunInProcess、RunAdapterInProcess 与库模式启动组件时，按 generated.Lookup 取得资源：插件用对象与记录的入口包，Adapter 用值与其具名类型所在的包。命令省略 Args 时的参数名、Service 与 Event Bus 的 schema ID、子插件配置的 schema 检查、配置准备与 i18n 都改在此时解析，规则不变；registry.go、bus.go 中现在注册时读取资源的位置一并后移。没有登记时，Collect 按无资源处理，与现在未绑定资源时相同，不启动组件的单元测试照旧可用；sdk.Run、sdk.RunAdapter 与库模式在连接 Core 前返回英文错误，指明入口包并提示 `alonabot build` 或 `alonabot run`。同包多次 New 共用一份只读资源，不同包互不影响。
- 仓库测试与夹具：原 sdk.UseGenerated(p, sdk.Generated{…}) 与 sdk.UseGeneratedAdapter 改为 generated.Bind(p, generated.Resources{…})，断言不变；本 module 之外的夹具源码（写入临时目录再编译的 main 程序）改为在 init 中调用 genimpl.Register。直接 `go build` 依赖生成资源的组件再运行的测试，改为经 CLI 或 pluginkit 构建：cmd/alonabot/e2e_test.go 的 TestMain 与 internal/core/supervisor/supervisor_test.go 构建 lifecycle 示例，alonabot_test.go 以 CompileApp 带 -race 构建 devbot，其余同类夹具照此处理；断言不变。
- 登记代码：internal/generate 产出的源码首行为 `// Code generated by alonabot. DO NOT EDIT.`，包名同入口包，import genimpl 后在 init() 中调用 genimpl.Register(键, genimpl.Resources{…})；键为入口包导入路径，package main 为 `main`。Adapter 包同样以其导入路径登记，不再按类型登记。生成器核对 New 签名、不包装 New 的规则不变。
- CLI：run、build、test、pack 只经 overlay 编入登记代码与正式入口，不写工程目录，删除 build.go 中非开发构建写入 alonabot_gen.go 的分支与扫描对该文件的排除和比对。check 只做扫描与资源检查，不比较生成文件。init 模板的 New 只含 sdk.NewPlugin() 与注册，工程中不生成 alonabot_gen.go。
- 库模式构建：`alonabot build --dir <目录>` 在目录没有 plugin.yaml、且是 import 根包 github.com/AlonaBot/AlonaBot 的 main 包时调用 CompileApp：用 `go list -deps` 找出其依赖中目录含 plugin.yaml 的组件入口包（包括模块缓存中的第三方 module），逐个以 Options.Package 只读扫描，把全部登记合成一个 overlay 文件放进该 main 包目录后编译。只输出可执行文件，缺省放在 `<目录>/bin/`，文件名为目录名，Windows 加 `.exe`；不生成安装包与 release.json。目录两者都不是时返回英文错误。overlay 只写入用户自己的 main 包目录，不替换模块缓存中的文件（go 命令拒绝这样做）。
- plugintest：Start 对每个被测插件，若 generated.Lookup 已有登记则直接使用；否则用 internal/generate 现场扫描其入口包：以测试进程的当前目录为 module 上下文，Options.Package 为入口包导入路径，GOOS、GOARCH 与 build tags 取自 debug.ReadBuildInfo 的 `GOOS`、`GOARCH` 与 `-tags`，扫描后 generated.Register。同一入口包在一个进程内只扫描一次，并发安全；扫描失败或有诊断时 t.Fatal 并给出原文。作者不需要先运行 CLI。测试 Adapter 改为启动前按对象 generated.Bind 本次 adapter_id 的资源，删除其 Generated 方法。
- Adapter 路由：runAdapter 先取 a.Descriptor()，再按 Descriptor.HTTP 建立 Listener：Auth 为零值时是 Core 会话私有路由，为 AuthPlatformToken 时是平台 token 公开路由。Method 与 Path 规范、重复检查和错误文本沿用插件路由的校验，不合法时在连接 Core 前返回英文错误；注册描述与原 AdapterHTTP 写法逐项相同。生成器在 Adapter 工程的入口包或其本模块依赖引用 adapter.HTTPRoute 时把 Result.HTTP 置为 true，取代按 AdapterHTTP 判定，宁可多报。devbot 的回环 Adapter 改在 Descriptor 中声明 `/inject`。
- 示例与文档：showcase 删除 Commands、HTTP、Schedules 方法，configdb 删除 Commands 方法，注册直接写在 New() 中，命令与行为不变（D-114）。全部示例、devbot、Agent 与 plugintest 的内部插件删除 UseGenerated 调用。docs/plugin-guide.md 与 docs/development.md 改为新写法：作者不写生成相关的内容，没有 alonabot_gen.go，go test 直接可用，库模式用 alonabot build；删除用 go generate 刷新生成文件的说明，Protobuf 的保留。

测试（按编码规范「测试提交规则」）：
- 提交：只改写并扩展 cmd/alonabot/plugin_cli_e2e_test.go 的 TestDevelopmentCLI，作为本切片唯一的主路径集成测试。init 之后工程中没有 alonabot_gen.go，plugin.go 不含 UseGenerated；在工程中写一个用 plugintest 发送 /hello 并核对 i18n 回复的 plugin_test.go，直接运行 `go test ./...`（不经 CLI）通过；build、check、test、pack 之后工程目录仍然没有 alonabot_gen.go；再写一个 import 根包、用 alonabot.Start 装配该插件、激活后打印结果并退出的 cmd/bot，`alonabot build --dir cmd/bot` 的产物激活成功，同一 main 包直接 `go build` 的产物在启动组件前以提示 `alonabot build` 的英文错误退出。既有断言按新模板改写，其余既有测试只迁移 API，断言不变。
- 本地自测（*_local_test.go，结果写入 PR）：V-120 的会话分项，即 SID/ParseSID、Ref/ParseRef、UID/ParseUID 的往返（含 `%`、`#`、`/`、`@` 与 external_user_id 中的 `:`）与各类不合法输入，SendTo 发往另一会话与同一会话的另一话题，没有默认目标时 c.Send 返回 NO_DEFAULT_TARGET，后台 sdk.Send 与 sdk.SendText，不合法的 sid 或句柄不发出请求；资源分项，即同包多次 New 共用资源、不同包隔离、重复 Register panic、没有登记时 Run、RunAdapter 与库模式的错误以及 Collect 不报错、在包级变量中调用 New 的组件在库模式下取得资源、plugintest 同一入口包只扫描一次与带 build tags 的扫描、经本地文件 GOPROXY 进入模块缓存的第三方组件经 `alonabot build` 库模式构建后取得资源、登记代码只含 genimpl.Register 且不包装 New、各命令构建后工程目录没有写入；Adapter 路由分项，即 Descriptor.HTTP 的私有路由与 platform token 公开路由、路径不合法与重复在连接 Core 前报错、注册描述与原写法一致、引用 HTTPRoute 时出现 http_listener；示例的命令与回复不变。

完成条件：V-120 中本切片的分项有本地证据；gofmt -l 对改动的 Go 文件无输出；go generate ./... 之后工作区没有变化（只剩 Protobuf 生成）；`git ls-files '*alonabot_gen.go'` 输出为空，仓库中没有 alonabot 的 go:generate 指令；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...；go build -tags configdb_gorm ./examples/plugins/configdb。注册语义改变、组件构建扫描了外部依赖、准备阶段的配置或迁移错误被吞掉时，按 PLAN T-04 的停止条件暂停交付。

不做：SDK 按类别分包（T-04.16，D-146）；根包装配移入 internal/app（T-13.5，D-148）；CLI 的位置参数与缺省值（T-13.6，D-149），本切片的库模式构建仍用 --dir；协议、Core 行为与管理 API 的改动；新增依赖；改写 plugins、adapters 章节未派发的 goal 语句；OneBot 11 反向 WebSocket（T-15.4）；向 sdk/ 或 internal/core/supervisor/ 提交新测试、增加 testdata 插件。
```

### goal-41 · T-09.6 唯一提交点与激活门禁

状态：已完成（[PR #179](https://github.com/AlonaBot/AlonaBot/pull/179)，合并提交 `8898251`），证据见 TEST-MATRIX §3.2。合入：`journal.(*Orchestrator).Commit(ctx, r) (*supervisor.Running, error)`，提交前失败经 Abort 以旧版新 Session 恢复，提交后激活失败返回 Degraded 的 Running；`state.(*Store).CommitCandidate` 在同一事务中复核并提交 Active、COMMITTED 与程序身份（含 BuildDigest、PublisherKeyID）；`state.(*Store).ReplaceProgram(ctx, pluginID, version, entrypoint, publisherKeyID string, buildDigest []byte)`；根包私有桥接 `managedCandidates` 经 `prepareCandidate`（journal.Prepare → 接入 → journal.Commit）实现 Reload 与 Update，`Rollback` 仍返回 UNSUPPORTED；首次启用改走候选准备与提交，Degraded 组件登记但不接入。接口调整见 DECISIONS 2026-11-13 变更记录。以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-41（切片 T-09.6：唯一提交点与激活门禁）。goal 目录 goal-41，分支 goal-41，worktree .worktrees/goal-41，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：确认候选的程序、配置快照、secret 与数据库引用都已持久化后，在 Core 状态库的同一事务中复核预期旧代次并提交 ActiveGeneration 与 COMMITTED（事务内不下载、不备份、不等待 RPC）；随后发送 ACTIVATE，只有 OnReady/ACTIVATED 完整成功才开放命令、事件、Service、Bus、任务与 HTTP；提交后启动或 OnReady 失败时保持新代次为权威并进入 Degraded，不自动倒退到旧数据库。旧代次的按钮随代次失效。

依据：PLAN T-09 步骤 5；CONTRACTS §12.1、§12.3、§6.3 第 2 条、§4.2；MESSAGE-MODEL §6.3；TEST-MATRIX V-30、V-21 的提交分项。

范围（只新增或修改这些）：
- internal/core/journal/：新增 commit.go（Commit 与提交后的激活）与管理入口 FirstInstall、Reload、Update（Prepare → Commit，任一步失败按 goal-40 的 Abort 处理）
- 根包 alonabot.go、assembly.go：把 Orchestrator 经根包私有桥接类型接到 manage.Options.Candidates（外部管理调用只进入一次 Core.operation，内部 Stop/Resume 直接调用运行循环方法，不在 operation 内再等待 operation；桥接的 Rollback 在 T-09.7 前返回现有 UNSUPPORTED）；startInstalled 的首次启用改走 FirstInstall，删除直接 AddGeneration → Launch(NewCandidate) 的 shortcut
- internal/core/manage/：没有 Active 的 Update 的程序身份替换（见行为）
- internal/core/state/：新增在同一事务中写 Active、COMMITTED 与 Journal 阶段的方法（新文件，不改 CommitActive 的已有行为）
- internal/core/supervisor/：激活结果的回调（只加调用点）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/state
   func (s *Store) CommitCandidate(ctx context.Context, opID, pluginID, expectedOld, newGen string) error // 同一事务：核对当前 Active == expectedOld、Journal 阶段为 prepared、代次快照已写入；然后更新 Active、写 COMMITTED、Journal 阶段改为 committed。不满足时返回 ErrStaleGeneration 或带原因的错误
internal/core/journal
   func (o *Orchestrator) Commit(ctx context.Context, r *Result) error // CommitCandidate → ACTIVATE → 等待 ACTIVATED；之后 Attach 到 Router/Broker/Bus/Scheduler/HTTP
   func (o *Orchestrator) FirstInstall(ctx context.Context, id string) error
   func (o *Orchestrator) Reload(ctx context.Context, id string) error
   func (o *Orchestrator) Update(ctx context.Context, id string, staged *packages.Candidate) error

行为：
- 提交前置：journal 阶段为 prepared、候选代次的 config_snapshot 与 generation_secret_path 已写入、候选目录与数据库文件存在（fsync 已完成），否则不提交，按 Abort 处理。
- 事务：CommitCandidate 只操作状态库元数据；任何文件操作或 RPC 都在事务之外完成。提交失败（Active 已被并发修改等）→ 不激活，Abort 候选，旧版以新 Session 恢复。
- 激活：提交成功后向候选 Session 发送 ACTIVATE（带 OnReady 截止），收到 ACTIVATED 后才 Attach 到 Router、Broker、Bus、Scheduler 与 HTTP（复用 goal-27 的接入点）；在此之前候选的路由不可见。
- 提交后失败（OnReady 失败、超时、进程退出）：新代次保持为 Active，插件进入 Degraded（不派发、HTTP 503、Scheduler 不触发），以 ERROR 记录；不切回旧代次、不使用旧数据库。之后的恢复由管理员 restart（以新代次的新 Session）或 T-09.7 的 rollback 处理。
- 按钮：提交新代次后，旧代次签发的按钮 token 在解析时因 generation 不是当前 Active 而失效（goal-8 已实现，本切片补测试）。
- 首次安装同样经此提交；first_install 提交后失败时插件为 Degraded 而不是未安装。首次启用切到 FirstInstall 后，cmd/alonabot/manage_e2e_test.go 等既有成功启用的断言保持不变。
- 没有 Active 的 Update（已安装但从未启用）：不建 Journal、不准备候选，管理预检通过后只把安装记录的程序身份（版本、入口、BuildDigest、发布者）与不可变包目录换成新包，首次启用时再走 first_install；role 为 webui 的 Update 保持 goal-84 的直接替换路径；开发组件不经 Journal。
- 不出现两个代次同时接单：旧进程在 goal-40 的冻结阶段已确认退出，提交后只有候选 Session 能 Attach。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：CommitCandidate 的每个前置条件；事务中没有文件与 RPC 操作（代码审查加测试钩子）；Degraded 状态下 Scheduler 不触发；internal/core/supervisor/commit_local_test.go（//go:build unix）：reload 成功：Active 改为新代次，COMMITTED 存在，Journal 为 committed；ACTIVATED 之前发送的消息没有派发给候选（候选的事件路由调用计数为 0，消息被旧版处理或因冻结被丢弃并计数），之后由新代次处理；提交前把 Active 改成别的值（模拟并发）：CommitCandidate 失败，候选被 Abort，旧版恢复；提交后 OnReady 失败：Active 仍为新代次，插件 Degraded，没有启动旧代次；HTTP 返回 503；提交后旧代次的按钮点击不调用插件。另经管理 HTTP 的现有 JSON 更新入口上传同一来源的新版本，以新 Active、版本与运行结果证明更新成功（goal-92 若已合并，同时经其 multipart 上传覆盖一次）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/...；go build ./...。

不做：Core 崩溃后的恢复判定与 rollback（T-09.7）；管理页面（T-09.8）。
```

### goal-142 · T-09.17 包校验收紧

状态：已完成（[PR #201](https://github.com/AlonaBot/AlonaBot/pull/201)，合并提交 `e04e50b`），证据见 TEST-MATRIX §3.2。合入：manifest 缺少 `migration_mode` 或 `prepare_effects` 时各给一条警告；`packages.Candidate.Warnings`；`manage.Package.Warnings`（`json:",omitempty"`，只在 Install 结果中填写）；`manage.Service.Update` 返回 `UpdateResult{Warnings}`（JSON 键 `warnings`，无警告时为 `{}`），管理服务逐条记 WARN 日志；plugin、adapter、agent 包的 build-info 严格核对（存在、可解析、与入口内嵌构建信息逐字一致、GOOS/GOARCH 与 manifest 一致），WebUI 包不要求；ADMIN-API §3、§10 同步。审查追加授权：`internal/app/webui_test.go` 的图标夹具补真实入口与 build-info（FIX-1）；`internal/app/admin_upload_test.go` 的 update 成功响应断言改为逐字核对市场前缀警告（FIX-2）；ADMIN-API §10 两处响应说明（FIX-3）。

```text
/goal 使用 $alonabot-execute 推进 goal-142（切片 T-09.17：包校验收紧）。goal 目录 goal-142，分支 goal-142，worktree .worktrees/goal-142，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：包内 manifest 缺少 migration_mode 或 prepare_effects 时只警告，安装与更新照常进行，警告出现在安装与更新的结果中并记 WARN 日志（D-150）；进程角色的包在安装与更新时严格核对 build-info，任一项不符时在停止旧版之前拒绝（D-151）。协议不变。

依据：PLAN T-09 步骤 10 及其失败停止条件；CONTRACTS §9.1、§9.2；ADMIN-API §3；DECISIONS D-150、D-151；TEST-MATRIX V-122。

范围（只新增、修改或删除这些）：
- contracts/manifest/manifest.go：缺少两项时的警告文案
- internal/core/packages/preflight.go：Candidate 携带 manifest 警告；build-info 核对
- internal/core/manage/：install.go、candidates.go、upload.go、http.go、service.go 中传出警告与记 WARN 日志
- spec/ADMIN-API.md §3：install 与 update 响应中的警告字段
- 既有测试中手工打包的夹具，只补 build-info，断言不变：internal/core/packages/preflight_test.go、cmd/alonabot/manage_e2e_test.go、cmd/alonabot/recover_e2e_test.go，以及 internal/app/admin_upload_test.go、internal/app/scheduler_tasks_test.go
- cmd/alonabot/manage_e2e_test.go 中新增的一个子测试（见「测试」）与上述包的 *_local_test.go
允许的新跨模块边：无（internal/core/packages 只增加标准库 debug/buildinfo 与 runtime/debug）。

接口（照此实现）：
- packages.Candidate 增加 Warnings []string：manifest.Parse 返回的警告，保持原顺序。
- manage.Package 增加 Warnings []string `json:",omitempty"`：只在 Install 的结果中填写，List 与其他返回 Package 的接口不填（JSON 中省略）；字段名与 Package 的其他字段一样沿用 Go 导出名。
- manage.Service.Update 改为 func (s *Service) Update(ctx context.Context, id string, req InstallRequest) (UpdateResult, error)；新增 type UpdateResult struct { Warnings []string `json:"warnings,omitempty"` }，与 update 的其他响应字段一样用小写键。HTTP 的 update（JSON 与 multipart）都返回 UpdateResult，没有警告时仍为 {}。
- 缺少的字段各给一条警告，文案固定为 "<字段名> not declared; preparation and rollback assume no external side effects"，<字段名> 为 migration_mode 或 prepare_effects；删除 "automatic upgrades are not allowed" 的说法。

行为：
- D-150：缺少 migration_mode、prepare_effects 或二者时，安装与更新照常完成；准备阶段按没有框架外副作用处理，与 none 相同，不返回 NEEDS_MANUAL_MIGRATION（internal/core/journal 只认 external 的现有判断保持不变）。Install 结果的 Warnings 与 Update 结果的 warnings 原样带出 manifest 的全部警告（包括组件 ID 缺少市场前缀等已有警告）；管理服务对每条警告记一条 WARN 日志，字段为 plugin_id 与 warning。CLI 原样打印响应，不改 CLI。取值不在规定集合内时照旧拒绝；REGISTER 回报的迁移方式与 manifest 不符时照旧准备失败。
- D-151：role 为 plugin、adapter、agent 的包，Preflight 在核对入口与图标文件之后依次检查：归档根目录有 build-info 文件；debug.ParseBuildInfo 能解析它；debug/buildinfo.ReadFile 读取的 entrypoint 内嵌构建信息的 String() 与 build-info 文件内容逐字相同；解析结果中 GOOS、GOARCH 两项设置分别等于 manifest 的 os 与 arch。任一项不符时返回 INVALID_ARGUMENT，英文错误分别为 "archive has no build-info"、"parse build-info: <原因>"、"build-info does not match the entrypoint"、"build-info GOOS does not match the manifest os"、"build-info GOARCH does not match the manifest arch"（文案写成命名常量）。这些检查与其他静态检查一样在停止旧版之前完成，被拒的更新不影响已安装版本的运行。role 为 webui 的包不要求 build-info，有也不核对。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix，真实 Core 与 pluginkit 打出的包，覆盖主路径）：缺 prepare_effects 的插件包安装成功，CLI 输出的 Package 带 Warnings、Core 日志有对应 WARN；更新为两项都缺的版本同样成功，响应带 warnings；build-info 改动一个字节的包与删除 build-info 的包更新被拒，错误指明问题，旧版仍在运行并响应命令。
- 既有测试：手工打包的夹具补上入口程序的 build-info（入口为真实 Go 程序，build-info 取 buildinfo.ReadFile(入口).String()）；preflight_test.go 中以伪造字节作入口的进程角色用例改用测试二进制作入口；断言不变。
- 本地自测（*_local_test.go，结果写入 PR）：缺 migration_mode、缺 prepare_effects、两项都缺的警告文案；build-info 缺失、不可解析、与入口不同、GOOS 不符、GOARCH 不符五种拒绝及错误文案；WebUI 包没有 build-info 照常预检；非法 migration_mode 与 prepare_effects 照旧拒绝；List 的 JSON 不含 Warnings。

完成条件：本地证据齐备；gofmt -l 对改动的 Go 文件无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...。缺字段的包被拒、被拒的更新停止了旧版，或需要改 .proto 时，按 PLAN T-09 的停止条件暂停交付。

不做：发布清单与签名格式的改动；WebUI 页面展示警告（T-09.9）；manifest 中其他字段的检查；新增依赖。
```

## 第 17 波

### goal-139 · T-04.16 SDK 按类别分包

状态：已完成（[PR #199](https://github.com/AlonaBot/AlonaBot/pull/199)，合并提交 `046e752`），证据见 TEST-MATRIX §3.2。合入：`sdk/task`（`Define`、`Type`、`On`、`Builder`、`Instance`、`Option`、`Key`、`Timezone`、`Pause`、`Resume`、`Status`、`State`）、`sdk/service`（`JSON`、`Proto`、`Spec`、`Serve`、`Builder`、`Access` 与其常量）、`sdk/bus`（`Define`、`Event`、`Subscribe`、`Subscription`）、`sdk/button`（`Define`、`Action`、`On`、`Builder`）、`sdk/media`（`Upload`、`Open`、`Grant`、`Image`、`File`、`Video`）、`sdk/conversation`（`Append`、`List`、`Clear`、`Entry`）、`sdk/tool`（`Invoke`、`List`、`Info`、`Confirm`、`Default`、`DefaultAllow`、`DefaultConfirm`）、`sdk/action`（`Client`、`NewClient`、`Caller`）与 `sdk/options`（`Serve`、`Query`、`Page`、`Option`）；根包与子包共用的注册状态在 `sdk/internal/registry`；`Collect`、`Registration` 及其描述类型不再导出，同进程入口 `internal/inproc.Plugin`、`internal/inproc.Adapter` 由 SDK 在 init 中登记，根包经它启动组件；生成器识别 `service.JSON`、`service.Proto`、`bus.Define` 与 `task.Define`。审查追加的修复：选项来源的固定 Schema ID 改为 `sdk/options.Query`、`sdk/options.Page`，删除运行期与生成器中按类型名改写旧 ID 的特例（FIX-1）；注释去掉已不导出的 `Collect` 与替换残留（FIX-2）。

```text
/goal 使用 $alonabot-execute 推进 goal-139（切片 T-04.16：SDK 按类别分包）。goal 目录 goal-139，分支 goal-139，worktree .worktrees/goal-139，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：SDK 的包级函数与类型不再集中在 sdk 根包。按 D-146 的对照把它们移入 sdk/task、sdk/service、sdk/bus、sdk/button、sdk/media、sdk/conversation、sdk/tool、sdk/action 与 sdk/options，名字去掉与包名重复的部分；根包只留插件对象、Handler 上下文、注册 builder、权限、事件类型与运行入口。RunInProcess、RunAdapterInProcess、Collect、Registration 及其描述类型不再导出。行为与协议不变，只改包与名字，调用方全部同步，不留别名。

依据：PLAN T-04 步骤 16 及其失败停止条件；CONTRACTS §1.3、§1.4、§3、§8、§9.1；DESIGN §D、§G、§H；DECISIONS D-86、D-114、D-122、D-146；TEST-MATRIX V-120；ARCHITECTURE 的 sdk 行。

范围（只新增、修改或删除这些）：
- sdk/：task.go、job.go、services.go、bus.go、buttons.go、media.go、media_bytes.go、conversation.go、tool.go、action.go、options.go 中移出的公开 API；根包其余文件中引用这些类型的位置；inproc.go 与 sdk.go 中改为不导出的入口；registry.go、plugin.go、subplugin.go 中改为经 sdk/internal 共享的注册状态
- 新增 sdk/task/、sdk/service/、sdk/bus/、sdk/button/、sdk/media/、sdk/conversation/、sdk/tool/、sdk/action/、sdk/options/，以及 sdk/internal/ 下承载共享注册状态的包
- sdk/llm/llm.go：GenerateService 改用 service.JSON
- internal/inproc/：增加由 sdk 在 init 中登记的同进程入口；根包 assembly.go 改经 internal/inproc 调用
- internal/generate/：对四个 JSON 泛型入口的识别改为 service.JSON、service.Proto、bus.Define、task.Define
- agent/、examples/、plugintest/、adapters/ 中使用这些 API 的源码与测试；根目录 *_test.go；internal/core/adapter/remote/remote_test.go；cmd/alonabot/*_test.go 中随包名改写的部分
- docs/plugin-guide.md、docs/development.md、README.md 与 examples/ 下 README 中的写法
- 上述包的 *_local_test.go
允许的新跨模块边：sdk/{task,service,bus,button,media,conversation,tool,action,options} -> sdk、-> sdk/message、-> sdk/errcode、-> sdk/internal/...，以及原根包对应代码已有的依赖；调用方 -> 这些子包。根包不 import 任何新子包；internal/inproc 仍不 import 本 module 的包。

接口（照此实现；类型的字段、方法集与函数的参数、返回值和行为保持不变，只换包与名字）：
| 原（sdk 根包） | 新 |
| --- | --- |
| Task[T](name) TaskType[T]、TaskType[T] | task.Define[T](name) task.Type[T]、task.Type[T] |
| OnTask[T](r, t) *TaskBuilder[T]、TaskBuilder[T] | task.On[T](r, t) *task.Builder[T]、task.Builder[T] |
| TaskInstance[T]、TaskOption、Key、Timezone | task.Instance[T]、task.Option、task.Key、task.Timezone |
| PauseJob、ResumeJob、JobStatus、JobState | task.Pause、task.Resume、task.Status、task.State |
| JSON、Proto、Service[In, Out] | service.JSON、service.Proto、service.Spec[In, Out] |
| Serve、ServiceBuilder[In, Out] | service.Serve、service.Builder[In, Out] |
| Access、AccessPrivate、AccessAllowlist、AccessPublic | service.Access、service.AccessPrivate、service.AccessAllowlist、service.AccessPublic |
| BusEvent[T](name) BusEventType[T]、BusEventType[T] | bus.Define[T](name) bus.Event[T]、bus.Event[T] |
| Subscribe、Subscription[T] | bus.Subscribe、bus.Subscription[T] |
| Action[T](name) ButtonAction[T]、ButtonAction[T] | button.Define[T](name) button.Action[T]、button.Action[T] |
| OnButton、ButtonBuilder[T] | button.On、button.Builder[T] |
| Upload、OpenMedia、GrantMedia | media.Upload、media.Open、media.Grant |
| ImageBytes、FileBytes、VideoBytes | media.Image、media.File、media.Video |
| AppendConversation、ListConversation、ClearConversation、ConversationEntry | conversation.Append、conversation.List、conversation.Clear、conversation.Entry |
| InvokeTool、ListTools、ToolInfo、ConfirmOperation | tool.Invoke、tool.List、tool.Info、tool.Confirm |
| ToolDefault、ToolAllow、ToolConfirm | tool.Default、tool.DefaultAllow、tool.DefaultConfirm（常量带类型名前缀，避免与 tool.Confirm 重名） |
| ActionClient、NewActionClient、ActionCaller | action.Client、action.NewClient、action.Caller |
| ServeOptions、OptionsQuery、OptionsPage、Option | options.Serve、options.Query、options.Page、options.Option |
| RunInProcess、RunAdapterInProcess | 不导出。sdk 在 init 中向 internal/inproc 登记两个入口，参数以 any 收插件对象与 Adapter，根包经 internal/inproc 调用 |
| Collect、Registration 与描述类型 Command、EventRoute、Schedule、HTTPRoute | 不导出，只在 sdk 内使用 |
根包保留：NewPlugin、Plugin、SubPlugin、Registrar；CommandBuilder、CommandGroup、EventBuilder、PreEventBuilder、BeforeSendBuilder、AfterSendBuilder、ScheduleBuilder、RouteBuilder；Context；Rule、Any、ToMe、NotCommand；Permission、PermissionName、Anyone、User、Admin、ChatAdmin；EventType、标准事件常量与 AdapterEvent；UserID；HTTPAuth 及其常量；Run、RunAdapter；Send、SendText、Edit、Delete、Via、SendOption；T；Error、ErrNotImplemented。

行为：
- 子包 import 根包取得 Registrar、*Context 与 Permission。路由登记、收集与分发要用的共享状态放在 sdk/internal，根包与子包都经它读写，不新增公开符号。子包登记的命令、任务、Service、订阅、按钮与选项来源，得到的注册描述、route_id、错误文本与检查次序和移出前相同。
- 子包函数的 Context 参数（例如任务与按钮 Handler 收 *sdk.Context）与泛型约束（Proto 的 proto.Message）不变；Handler 内的平台动作仍用 c.CallAction，action.Caller 由 *sdk.Context 与 *action.Client 实现。
- 生成器对四个 JSON 泛型入口的检查（类型须为包级具名非泛型）与 Service Schema 的生成改为识别 service.JSON、service.Proto、bus.Define 与 task.Define，诊断位置与文本不变；ParamNames、i18n 常量键与迁移方式的识别不受影响。
- 删除旧名，不留别名或转发函数。CONTRACTS 与 DESIGN 已按新名书写（PR #196）；plugin-guide、development、README 与示例 README 同步。

测试（按编码规范「测试提交规则」）：
- 提交：不新增测试。既有测试只迁移到新包名，断言不变；TestSDKExamples、TestAgentPackage、TestPluginHarness 与 TestDevelopmentCLI 覆盖分包后的示例、Agent、plugintest 与 CLI 的编译与运行。
- 本地自测（*_local_test.go，结果写入 PR）：V-120 的分包分项，即根包的导出集合只剩上文「根包保留」的标识符；各子包的导出与上表一致；`go list -deps ./sdk` 不含新子包；示例插件与 Agent 的 REGISTER 描述与移出前逐项相同（以 main 上的描述为基准对比）；生成器对新入口的 Schema 生成与诊断和移出前相同。

完成条件：V-120 中本切片的分项有本地证据；gofmt -l 对改动的 Go 文件无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...；go build -tags configdb_gorm ./examples/plugins/configdb；`go doc ./sdk` 不再列出上表左栏的标识符。注册描述或行为发生变化、根包 import 子包、出现循环依赖时，按 PLAN T-04 的停止条件暂停交付。

不做：协议与行为的改动；根包装配移入 internal/app（T-13.5）；CLI 的位置参数与缺省值（T-13.6）；改写 plugins、adapters 章节未派发的 goal 语句（合并后由审查者改写）；新增依赖；向 sdk/ 或 internal/core/supervisor/ 提交新测试、增加 testdata 插件。
```

### goal-42 · T-09.7 崩溃恢复、回退与回收

状态：已完成（[PR #187](https://github.com/AlonaBot/AlonaBot/pull/187)，合并提交 `864819e`），证据见 TEST-MATRIX §3.2。合入：`journal.(*Orchestrator).Recover(ctx) (blocked []string, err error)`（取不到 runtime.lock 的插件以 blocked 返回，根包不启动它们，管理状态为 Degraded）、`KindRollback` 与 `PrepareRollback(ctx, pluginID, generation string) (*Result, error)`、根包 `managedCandidates.Rollback`（须确认并按目标版本核对依赖）、`internal/core/gc.Collect`、状态库保留最近三个已提交代次；first_install 的提交同时写入启用意图；资源按当前 data_dir 与包身份定位；docs/development.md「备份与恢复」。接口调整见 DECISIONS 2026-11-13 变更记录。以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-42（切片 T-09.7：崩溃恢复、回退与回收）。goal 目录 goal-42，分支 goal-42，worktree .worktrees/goal-42，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 启动取得 core.lock 后按状态库的 Active 与 Journal 恢复：未提交的操作关闭候选、以旧程序启动新 Session（首次安装没有 Active 则保持未激活），已提交的只恢复当前代次；旧进程退出未确认时不启动新写入者。rollback 从选定的保留代次复制新的 SQLite 恢复候选并使用新代次，走同一准备与提交路径；插件自管的外部数据沿用显式维护。只回收无人引用的失败候选、旧包与 secret，恢复中的操作与保留快照受保护。

依据：PLAN T-09 步骤 6；CONTRACTS §12、§6.3、§9.3（rollback）；MESSAGE-MODEL §6.3（rollback 不复活旧按钮）；TEST-MATRIX V-30、V-31、V-36。

范围（只新增或修改这些）：
- internal/core/journal/：新增 recover.go（启动时恢复）与 rollback.go（回退候选的准备）；journal.go 只为回退复用准备流程做内部调整，不改 Prepare 的签名
- internal/core/state/：保留代次的读写（新文件；保留策略见下）
- 新增 internal/core/gc/（引用回收）
- 根包 assembly.go：serve/Run 启动时在启动任何包之前调用恢复；`managedCandidates.Rollback` 改为在 `operation` 中调用 `PrepareRollback`，再沿用 `apply` 的接入、Commit 与失败停止路径（可把 `prepareCandidate` 改为接收准备函数，不新增公开符号）
- docs/development.md：增加「备份与恢复」一节，说明整机备份覆盖整个 data_dir 与插件自管的外部数据（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
internal/core/journal
   func (o *Orchestrator) Recover(ctx context.Context) error // 处理全部未到 committed 或 cleaned 的 Journal
   const KindRollback Kind = "rollback"
   func (o *Orchestrator) PrepareRollback(ctx context.Context, pluginID, generation string) (*Result, error) // 准备回退候选，提交由调用方经 Commit 完成
internal/core/gc
   func Collect(ctx context.Context, st *state.Store, secrets *secret.Store, packagesDir string) (removed []string, err error)
internal/core/state
   保留代次：每个插件保留最近 3 个已提交的代次（命名常量）及其程序包与数据库；读写方法自定。

行为：
- Recover（Core 启动时，在启动任何包之前；已持 core.lock）：
  - Journal 阶段为 started、stopped_old、candidate_ready、prepared、failed：候选未提交。确认没有仍在运行的候选或旧进程（通过 runtime.lock 能否取得判断；取不到时该插件标为 Degraded 并以 ERROR 说明「旧进程可能仍在运行」，不启动任何实例）；然后 candidate.Discard 候选资源，Journal 改为 cleaned；有旧 Active 时稍后按正常启动流程启动旧版（新 Session），首次安装则保持未激活。
  - Journal 阶段为 committed 但插件没有成功 ACTIVATED 过：只按当前 Active 启动（新代次），失败则 Degraded；不切回旧代次。
  - 恢复判断只信状态库的 Active 指针与 Journal，不读插件目录猜测。
- Rollback（管理服务经 `manage.Candidates.Rollback` 调用根包的 `managedCandidates.Rollback`，持插件操作锁）：选定的代次必须是保留代次之一且不是当前 Active；`PrepareRollback` 以 Journal kind=rollback 沿用 goal-40 的准备流程：停止当前版本、从选定代次的数据库用 Backup API 复制到新代次（不是复用旧文件）、以选定代次的程序包与配置快照启动候选（activation_reason=rollback），返回 PREPARED 的 Result；根包再经 `journal.Commit` 提交并按 apply 的方式接入，提交前失败经 Abort 恢复当前版本。新代次 ID 为新值，旧代次签发的按钮不会因此复活（解析时比较的是当前 Active）。声明 prepare_effects=external 的插件 rollback 返回 FAILED_PRECONDITION（"NEEDS_MANUAL_MIGRATION"：需要维护与外部备份），不假装能恢复。
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

状态：已完成（[PR #192](https://github.com/AlonaBot/AlonaBot/pull/192)，合并提交 `0cff77c`），证据见 TEST-MATRIX §3.2（本地自测，D-86 不提交）。合入：internal/core/manage 的 config_api、policy_api、settings_api、tool_api、conversation_api 与路由注册（spec/ADMIN-API.md 已登记）；`options.(*Service).QueryBuiltin`；根包向管理服务注入 Router、policy、options.Service 与 tool.Catalog；reload 激活已有 Pending、没有时导入编辑快照，CLI reload 先 config/import（D-143，审查者补齐，待维护者确认）。以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-93（切片 T-09.14：配置、权限、工具策略与会话的管理 API）。goal 目录 goal-93，分支 goal-93，worktree .worktrees/goal-93，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在管理服务中补齐 Admin API：插件配置（raw JSON、Merge Patch、base_generation 与 secret 语义，Pending 经 reload 激活）、命令权限与启用、命令名覆盖、配置字段的选项查询、发信许可（allowed_adapter_instances）、Core 设置（Core Admin 列表、默认 locale、命令前缀、全局禁止表）、工具调用策略、Adapter 实例（按 adapter_config_schema.json 校验）、Agent 会话的列出与清除，以及 secret 明文查看的再认证与审计；Core 设置与 Adapter 实例写回 core_config.json。接口追加到 spec/ADMIN-API.md；页面由 WebUI 章节的 T-09.9 实现。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §5、§2.3、§8（allowed_adapter_instances）、§13、§7.3、§5.4；DECISIONS D-45、D-62、D-128、D-129、D-130；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56、V-115、V-116、V-117 的管理 API 分项。

范围（只新增或修改这些）：
- internal/core/manage/：新增 config_api.go、policy_api.go、settings_api.go、tool_api.go、conversation_api.go；在管理服务的 HTTP 路由中追加注册
- internal/core/options/：新增 `func (s *Service) QueryBuiltin(ctx context.Context, source string, filter map[string]string, q Query) (Page, error)`，只接受 `configschema.BuiltinOptionSources` 中的来源与其声明的 filter 键，其他来源或键返回 INVALID_ARGUMENT；复用现有内置来源的实现与分页上限
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
- GET /plugins/{id}/conversations、DELETE /plugins/{id}/conversations/{key}。key 是会话存储键（Agent 为 `<会话字符串>@<实例名>`，D-142），在路径中按段转义，服务端解码后原样交给 conversation.Service。
- POST /secrets/{plugin}/{path}/reveal：请求体带当前管理 token 再认证（adminauth.Reauthenticate），通过后返回明文并以 INFO 审计记录（不含明文）。

行为：
- 配置：invisible 字段在读取中照常返回，是否展示由页面决定；PATCH 不会因为未提交 invisible 字段而删除它。
- 命令权限：有效权限与启用状态按 policy.CheckCommand 的同一规则计算后返回，覆盖分插件、命令组、路由三层；同时返回遮蔽关系。
- 命令名：保存与删除经 goal-129 的 policy 方法，校验与重名规则不在本切片重复实现；保存后下一次派发生效，遮蔽关系按生效的命令名返回。
- GET /plugins/{id}/config/options?path=…&q=…：按字段路径列出选项（D-129），返回 {options: [{value, label, description}], more}；POST /plugins/{id}/config/options/resolve：{path, values: [...]}，返回这些存储值当前的选项，不存在的值不返回。来源名与 filter 由 goal-130 的查询入口从该插件 Active 代次的 schema 取得，请求不能指定来源；locale 取管理会话的界面语言。字段不存在或没有 `options_from` 返回 INVALID_ARGUMENT，来源不可用返回 503 与 UNAVAILABLE。Adapter 实例配置提供同样的两个接口，挂在本切片已有的 Adapter 实例配置路径之下。
- GET /core/options/{source}?q=…&filter[键]=…：经 `options.QueryBuiltin` 直接查询内置来源（`core.chats`、`core.adapter_instances`），filter 取自请求的 `filter[键]`，供 Core 设置页选择全局禁止表中的会话；服务来源不能经此接口查询，返回 INVALID_ARGUMENT。POST /core/options/{source}/resolve 按存储值取回。
- 选项查询的本地自测（不提交，PR 中写明结果）：插件字段的列出与取回、请求中多出的来源参数被忽略、没有 `options_from` 的字段、Agent 未启用时 `@agent/llm.providers` 返回 UNAVAILABLE、`core.chats` 按名称搜索与按 `chat_type` 筛选、未认证被拒。
- Core 设置：Core Admin 列表、默认 locale、command_prefixes 与按实例的覆盖都经 goal-97 的 settings.Update 写回 core_config.json，保留原有的 ${...} 写法；写回成功后 Core Admin 列表调用 policy.SetAdmins，命令前缀在下一次派发生效，默认 locale 在之后建立的 Session 生效。alonabot admin add|remove|list 使用同一服务。全局禁止表仍在状态库。
- 开发组件（dev 来源）：PUT/PATCH 配置时校验通过后写回 <data_dir>/plugins/<id>/config.json 并重启该组件，不建 Pending（CONTRACTS §15.4）。
- 会话：只返回摘要（键、条目数、最后时间），不返回条目内容。
- 所有修改请求经 CSRF 检查；修改后的策略在下一次派发时生效，不需要重启插件。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：各端点的错误格式与 409；读取结果中没有 secret 明文；根包 admin_api_local_test.go（//go:build unix，复用 admin_upload_test.go 的真实 Start 装配）：经管理 API 完成：PATCH 配置（未修改的 secret 保持原引用、新 secret 生成新引用、invisible 不被删除）、陈旧 base 409、reload 后 Active 改变；设置命令覆盖后下一条消息按新权限处理；保存命令名覆盖后旧名不再触发、新名触发，空列表与同插件重名被拒，DELETE 后恢复声明的命令名；把发信许可收紧为其他实例后插件后台 Send 被拒、恢复为全部实例后成功；Core 设置写回后 core_config.json 中的 ${...} 原写法保留；设置默认 locale 后新 Session 使用；工具策略改为 deny 后 ListTools 不含它；清除会话；开发组件的配置写回 config.json 并重启；reveal 需要再认证。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：页面（T-09.9）；日志 API（T-09.10）；故障注入与验收汇总（T-09.11）。
```

### goal-168 · S-01.1 可休眠声明与持久化运行状态

状态：已完成（PR #225），合并提交 `2cc32ff`。以下保留派发时语句。派发时：前置 T-04.1（PR #34）、T-09.4（PR #163）、T-09.6（PR #179）均已合并。main `bae7faa` 已核对：manifest.Manifest/ParseProject/Generate 尚无可休眠字段；state.pluginState 只保存 enabled 与 ActiveGeneration；manage.describe 在已启用但无运行实例时显示 Degraded。低层 host 状态与逐插件管理锁已可复用。与另一 S-01 切片共享状态库及装配文件，按执行顺序的冲突规则合并。

```text
/goal 使用 $alonabot-execute 推进 goal-168（切片 S-01.1：可休眠声明与持久化运行状态）。goal 目录 goal-168，分支 goal-168，worktree .worktrees/goal-168，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在作者与包内 manifest 保留可休眠声明；持久化与 enabled 分开的插件运行状态。现有 always_on 生命周期照常运行，后继切片可以准确识别 SUSPENDED，而管理读取不会唤醒插件。

依据：PLAN §7 S-01 步骤 1、2；CONTRACTS §4.4、§4.5、§9.1；ADMIN-API §24.2 的组件列表；DESIGN §A；DECISIONS D-53、D-153；TEST-MATRIX V-77 的 manifest、运行状态与未启用分项。

范围（只新增或修改这些）：
- contracts/manifest/：manifest.go、project.go、build.go，贯通 ParseProject、Generate 与 Parse
- internal/core/state/：schema.sql 与新文件 runtime.go；state.go 的安装初始化、installs.go 的清理调用点；已有模型需加列时同步 models.go，持久化只用 GORM
- internal/core/supervisor/：启动、准备、激活、停止、异常退出处的状态记录
- internal/app/：assembly.go、alonabot.go、journal.go 以及进程退出监视、候选提交/恢复的相邻调用点，覆盖独立进程与同进程开发组件
- internal/core/manage/：service.go、lifecycle.go 及既有更新/回退的状态读取调用点
- cmd/alonabot/manage_e2e_test.go；上述包的 *_local_test.go

允许的新跨模块边：无；manifest 属于 contracts，状态与生命周期属于 internal/core，internal/app 与 cmd 的装配沿用 ARCHITECTURE 已允许方向。

接口（main bae7faa 上的现有签名，保持可用）：
- manifest.ParseProject(data []byte) (*manifest.Result, error)
- manifest.Generate(project *manifest.Manifest, build manifest.BuildInfo) ([]byte, error)
- manifest.Parse(data []byte) (*manifest.Result, error)
- supervisor.Launch(ctx context.Context, store *state.Store, spec supervisor.Spec) (*supervisor.Running, error)
- supervisor.Prepare(ctx context.Context, store *state.Store, spec supervisor.Spec) (*supervisor.Prepared, error)
- (*supervisor.Prepared).Activate(ctx context.Context, commit func(context.Context) error) (*supervisor.Running, error)
- (*supervisor.Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult
- (*supervisor.Running).WaitExit() (abnormal bool, state string, err error)
- (*state.Store).SetEnabled(ctx context.Context, pluginID string, enabled bool) error
- (*manage.Service).List(ctx context.Context) ([]manage.Package, error)

接口（本切片新增，以下签名固定）：
- manifest.Manifest 追加 SleepCapable bool，YAML 名 sleep_capable，缺省 false，role=plugin 与 role=agent 均可声明；工程与包内值一致
- package state：type RuntimeState string；常量 RuntimeStopped="STOPPED"、RuntimeWaking="WAKING"、RuntimeActive="ACTIVE"、RuntimeQuiescing="QUIESCING"、RuntimeSuspended="SUSPENDED"、RuntimeFailed="FAILED"
- type RuntimeStatus struct { State RuntimeState; Reason string; ChangedAt time.Time; LastWakeReason string; LastWakeAt time.Time; LastSuspendAt time.Time }
- func (s *Store) SetRuntimeStatus(ctx context.Context, pluginID string, status RuntimeStatus) error
- func (s *Store) RuntimeStatus(ctx context.Context, pluginID string) (RuntimeStatus, error)
- manage.Package 追加 RuntimeState string（json:",omitempty"），plugin 与 agent 取上述六态，Adapter 与 WebUI 省略；State 保留既有阶段语义，读取到 RuntimeSuspended 时为 Suspended

行为：
- 以已有 plugin/agent 安装记录初始化 STOPPED；状态与 enabled 分开读写。WAKING 覆盖启动/准备/激活，ACTIVATED 确认后才写 ACTIVE；正常停止期间为 QUIESCING，完成为 STOPPED，准备/激活/退出失败记录 FAILED。持久化失败返回完整错误，资源清理由原所有者完成。
- 候选成功/失败、回退、崩溃恢复与逐插件管理锁全部接入；Core 启动先按现有 journal.Recover 和旧进程清理核对事实，再更新状态，落盘的 ACTIVE/WAKING 不直接作为存活证据。延迟退出通知须按实际代次/Session 所有权处理，保留新实例状态。
- LastWakeAt/LastSuspendAt 由实际唤醒/空闲退出的后继切片填写，本切片初始化为未发生。SUSPENDED 类型与存储可以使用，本切片不主动触发休眠。
- mode 尚未开启时维持 always_on；不替换已有 manage.Options.State 回调、Package.State 或底层 host 状态机，不给 Adapter 添加休眠语义。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 的真实安装/启停主路径增加一个子测试，构建声明 sleep_capable 的插件包，经现有管理 API enable 后列表 RuntimeState 为 ACTIVE、State 为 Activated，disable 后为 STOPPED 且 enabled=false；重新启动 Core 后核对状态与安装记录。复用现有动态构建的测试工程，不新增 testdata 插件。
- 本地自测（*_local_test.go，结果写入 PR）：manifest 缺省 false、plugin/agent 显式 true/false、非布尔值与 Adapter/WebUI 拒绝；SQLite 持久化、恢复时陈旧状态、准备/OnReady 失败、迟到退出通知；原管理生命周期与 dev/in-process 路径。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/state/... ./internal/core/manage/... ./internal/core/supervisor/... ./internal/app/... ./cmd/alonabot/...；go build ./...。PR 写明 V-77 本切片分项的测试位置、命令、平台与结果。

不做：路由快照（goal-169）、EnsureActive/活动引用/IDLE 协议及空闲停止（goal-170、goal-171）、sleep 设置/API/CLI 与测量；不改 .proto，不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-169 · S-01.2 已提交代次的路由快照与索引

状态：已完成（PR #226），合并提交 `413a6ff`。以下保留派发时语句。派发时：前置 T-04.1（PR #34）、T-09.4（PR #163）、T-09.6（PR #179）均已合并。main `bae7faa` 已核对：host.Prepared 提供 Register 与 RouteHash；supervisor.prepare 只保存配置快照，Prepared.Activate 通过 commit 回调调用唯一提交点；state.CommitCandidate 与 CommitActive 均未保存路由快照；Router 当前只索引 Attach 的活 Session，Detach 会撤销绑定。与另一 S-01 切片共享状态库及装配文件，按执行顺序的冲突规则合并。

```text
/goal 使用 $alonabot-execute 推进 goal-169（切片 S-01.2：已提交代次的路由快照与索引）。goal 目录 goal-169，分支 goal-169，worktree .worktrees/goal-169，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：把 REGISTER 的完整路由白名单、路由摘要和程序包摘要持久化，并与唯一 ActiveGeneration 提交原子生效；Core 重启与无进程期间仍可取得当前路由索引。重新注册与快照不一致时隔离插件，业务保持关闭。

依据：PLAN §7 S-01 步骤 1、3；CONTRACTS §4.4、§12.2、§12.3；WIRE-FORMAT §3.3、§5A；DESIGN §A、§N；DECISIONS D-54、D-153；TEST-MATRIX V-77 的快照与摘要分项。

范围（只新增或修改这些）：
- internal/core/state/：schema.sql、新文件 routes.go；state.go 的 CommitActive 与 commit.go 的 CommitCandidate 事务；candidate.go、installs.go、retained.go 中的代次删除/数据清理调用点
- internal/core/supervisor/：supervisor.go 的准备与 prepare.go 的激活前校验；从已验证 Registration 取白名单
- internal/core/router/：新文件 snapshots.go，必要的 Router 字段/初始化与索引读取调用点
- internal/core/journal/：准备、提交、abort/recover 与回收流程中快照绑定的必要调用点
- internal/core/gc/：随代次回收快照的调用点
- internal/core/manage/lifecycle.go：停用/卸载时撤销快照索引的必要调用点，沿用既有装配回调
- internal/app/：assembly.go、journal.go、alonabot.go 中已安装插件的准备、恢复、快照装载、停用与卸载索引清理；同进程 dev 路由沿用现有路径
- internal/app/journal_test.go；上述包的 *_local_test.go

允许的新跨模块边：无；state、supervisor、router、journal、gc 同属 internal/core；internal/app -> internal/core/... 与 internal/core/... -> contracts 沿用 ARCHITECTURE。Router 的快照参数只使用 pb.RouteDescriptor，不反向依赖装配层。

接口（main bae7faa 上的现有签名，保持可用）：
- host.Prepared.Register *pb.Register、RouteHash []byte；supervisor.Prepared.Registration *host.Prepared
- (*state.Store).CommitActive(ctx context.Context, pluginID, expectedOld, newGen string) error
- (*state.Store).CommitCandidate(ctx context.Context, opID, pluginID, expectedOld, newGen string) error
- (*journal.Orchestrator).Commit(ctx context.Context, r *journal.Result) (*supervisor.Running, error)
- (*journal.Orchestrator).Recover(ctx context.Context) (blocked []string, err error)
- (*supervisor.Prepared).Activate(ctx context.Context, commit func(context.Context) error) (*supervisor.Running, error)
- router.FromInstance(in *host.Instance) router.Plugin
- (*router.Router).Attach(p router.Plugin, c router.Caller)
- (*router.Router).Detach(pluginID string)

接口（本切片新增，以下签名固定）：
- package state：type RouteSnapshot struct { PluginID, GenerationID string; Routes []*pb.RouteDescriptor; RouteHash, BuildDigest []byte }
- func (s *Store) SaveRouteSnapshot(ctx context.Context, snapshot RouteSnapshot) error
- func (s *Store) ActiveRouteSnapshot(ctx context.Context, pluginID string) (RouteSnapshot, error)
- func (r *Router) SetRouteSnapshot(pluginID, generationID string, routes []*pb.RouteDescriptor)
- func (r *Router) RouteSnapshot(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool)
- func (r *Router) RemoveRouteSnapshot(pluginID string)

行为：
- SaveRouteSnapshot 只写匹配插件的未提交候选；已经提交的快照不可覆盖。新候选准备完成后先保存白名单快照；CommitActive/CommitCandidate 在同一事务复核快照身份及包/路由摘要，再提交当前代次；update/rollback 核对候选 Journal 中的新程序身份，ActiveRouteSnapshot 核对提交后的当前安装身份。已有 Adapter、静态 WebUI 与同进程 dev 的提交按原路径使用；快照用于已安装的 role=plugin 与 role=agent 独立进程。
- RouteHash 复用 host.Prepared.RouteHash 及既有 WIRE 确定性编码算法，BuildDigest 复用已验证程序身份；不另造摘要格式。快照只编码 RouteDescriptor 集合（S-02.3 已合并时含 message_filter 及配置字段路径；否则由后合并者补齐）和包资源绑定，不保存过滤配置值，不序列化整个 Register，排除 listener、配置 JSON、Session 与认证材料。Schema 从该代次不可变包资源核验后恢复。
- 普通 restart/resume 在准备完成、发送 ACTIVATE 之前比较当前快照与重新 REGISTER 的路由摘要，失败关闭候选并返回 FAILED_PRECONDITION；保持原 ActiveGeneration、提交标记与资源。新版本/配置候选可有新路由，只在同一提交后替换有效索引。
- 从状态库读取当前已 COMMITTED 的代次，核对安装身份与包摘要；恢复时先完成 Journal 权威判定，再装载所有已启用插件的快照索引。快照缺失或不一致返回实际错误并隔离该插件；其他可用插件继续运行。
- 快照索引与活 Session 绑定分别维护；Detach 撤销 Caller 与 Session 后仍保留快照，明确停用/卸载时 RemoveRouteSnapshot；回收代次与 purge-data/purge 同步清理持久快照。现有 Plugin/PluginFor/Call 的 Activated 门禁保持有效，快照本身只允许定位，不产生假 Caller，也不承诺无 Session 的业务交付。

测试（按编码规范「测试提交规则」）：
- 提交：internal/app/journal_test.go 增加一个真实子进程＋SQLite 的集成子测试，完成候选提交后停止实例，从 ActiveRouteSnapshot 读取完整路由，重启 Core 后重建索引并以原代次启动；断言快照对应的提交身份、路由摘要，且数据库持久内容排除 Listener、配置与 Session 认证材料。复用已有测试工程构建机制，不新增 testdata 插件。
- 本地自测（*_local_test.go，结果写入 PR）：候选未提交/提交失败保持旧快照可见，错误 plugin/generation/包摘要，重新 REGISTER 摘要不一致、Schema 资源校验失败、快照写入失败，停用/卸载与代次回收；Agent 快照往返；Adapter、WebUI 及 dev 路径原有提交回归。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/state/... ./internal/core/router/... ./internal/core/journal/... ./internal/core/supervisor/... ./internal/core/gc/... ./internal/app/...；go build ./...。PR 写明 V-77 快照分项的命令、平台、结果与未运行项。

不做：manifest 与运行状态（goal-168）、休眠与唤醒（goal-170、goal-171）、新业务 RPC、Scheduler 唤醒与管理 API；不改 .proto，不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

## 第 18 波

### goal-150 · T-04.17 插件页面套件

状态：已完成（PR #218），合并提交 `6d81be3`。以下保留派发时语句。派发时：前置 T-04.16 与 T-07.7 已合并。2026-11-17 按 D-159 新增；桥的 WebUI 一侧由 goal-44 实现。

```text
/goal 使用 $alonabot-execute 推进 goal-150（切片 T-04.17：插件页面套件）。goal 目录 goal-150，分支 goal-150，worktree .worktrees/goal-150，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：新增 `sdk/page`：以 embed 提供 `kit.css` 与 `kit.js`，插件把它挂在自己的 HTTP 路由下，页面引用后自动跟随 WebUI 的明暗、令牌与语言，并能请求本插件的路由。

依据：PLAN T-04 步骤 17；ADMIN-API §7（hello、theme、locale、shortcut 与令牌变量表）；DECISIONS D-84、D-159、D-169；TEST-MATRIX V-130 的 SDK 分项。

范围（只新增或修改这些）：
- sdk/page/（新包）：page.go（包文档与 Handler）、kit.css、kit.js
- examples/plugins/showcase：增加一个用套件的页面路由，作为示例
- docs/plugin-guide.md：插件页面与套件的用法
允许的依赖：无新增；kit.js 与 kit.css 不依赖第三方库。

接口：
- package page
  func Handler() http.Handler // 只响应路径末段为 kit.css 与 kit.js 的 GET，其余 404；Content-Type 正确，带基于 SDK 版本的 ETag
- kit.js 在 window.AlonaBot 上提供：request(method, path, body?) → Promise<Response 数据>（经桥请求本插件路由）；t(key, args?) → string（取当前语言的插件文案）；onTheme(fn)、onLocale(fn)

行为：
- kit.css 只使用 ADMIN-API §7 的 `--ab-*` 变量，给按钮、输入框、选择框、表格、标签与卡片写好样式；没有收到令牌时使用与 goal-166 浅色相同的缺省值（--ab-bg #F3F6F9、--ab-surface #FFFFFF、--ab-surface-2 #F4F7FA、--ab-line #DFE5EC、--ab-text #1F2A44、--ab-text-muted #5E6B82、--ab-accent #0E7C92、--ab-accent-soft #E2F5F8、--ab-success #1F9D6B、--ab-warning #D99A06、--ab-danger #F4505A、--ab-critical #C084FC，圆角 6/10/16px，字体同 goal-166）。
- kit.js 加载后向父窗口发送 `hello`，把收到的令牌写到 :root 并设置 data-theme；监听 theme 与 locale 事件；捕获 ⌘K/Ctrl+K 并发送 `shortcut`。不在 iframe 中打开时只应用缺省值，不报错。

测试（按编码规范「测试提交规则」）：
- 提交：无（D-86：sdk/ 不提交测试）。
- 本地自测（结果写入 PR）：Handler 的路径、类型与 ETag；用 showcase 的页面在浏览器中直接打开（缺省值）与在模拟父窗口中打开（发送 hello、令牌写入、主题与语言事件、⌘K 转发），写明浏览器与版本。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：WebUI 一侧的桥（goal-44）；把官方 WebUI 的组件库提供给插件。
```

### goal-152 · T-08.16 去掉配置版本与迁移，配置宽松载入

状态：已完成（PR #239），合并提交 `1a1d12c`。以下保留派发时语句。派发时：前置 T-08.11、T-08.12 与 T-09.14 已合并。2026-11-17 按 D-161 新增。改动面大，派发前确认同一波的 goal-160 与 goal-151 尚未改 `contracts/configschema/` 的同一批文件，否则先合并它们再派发本条。

```text
/goal 使用 $alonabot-execute 推进 goal-152（切片 T-08.16：去掉配置版本与迁移，配置宽松载入）。goal 目录 goal-152，分支 goal-152，worktree .worktrees/goal-152，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：删除配置的 `$version`、`$migrations` 与全部版本字段，改为按 CONTRACTS §5.5 宽松载入：缺失补 default、能无损转换的转换、不能转换的恢复 default、移除失效的模板元素，并产出校正记录；管理 API 返回校正记录并提供清理入口。

依据：PLAN T-08 步骤 13；CONTRACTS §5.1、§5.2、§5.5、§15.1；WIRE-FORMAT §3.2、§3.3；ADMIN-API §12.1、§13、§14；DECISIONS D-161；TEST-MATRIX V-37、V-38、V-88、V-132。

范围（只新增或修改这些）：
- contracts/configschema/：删除 migration.go、migration_path.go 及其检查；新增校正（correct.go）；prepare.go 改为先 OnConfigLoad 再校正
- contracts/manifest/：删除 config_version
- contracts/proto：Bootstrap 的 previous_config_version、target_config_version 与 Register 的 candidate_config_version 改为 reserved；Register 追加 `repeated ConfigCorrection config_corrections`；生成产物与 wirespec_test.go
- sdk/：OnConfigLoad 的签名与调用；internal/core/：代次元数据保存校正记录，settings 与 Adapter 实例的校正（不改写 core_config.json），管理 API 的校正记录与三个 prune 端点、GET /api/v1/core/settings/corrections
- 全部示例、Agent、官方 Adapter 与测试夹具中的 `$version`、`$migrations` 与 OnConfigLoad 调用；docs/plugin-guide.md、docs/development.md
- 上述的测试文件

接口：
- contracts/configschema：type Correction struct { Path string; Reason string /* converted、reset、removed */; Old json.RawMessage /* secret 字段为空 */ }；func (s *Schema) Correct(raw json.RawMessage) (json.RawMessage, []Correction, error)
- sdk：func (p *Plugin) OnConfigLoad(fn func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error))

行为：未知键保留但不生效；`core_config.json` 不是合法 JSON 对象时照旧拒绝启动（CRITICAL 由 goal-149 负责）；其余不合法的值按 §5.5 校正并记 WARN；清理只删除被列为无效的项。首次发布前不留兼容：旧写法的 `$version` 按未知键处理。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix）：插件配置中一个字段类型不符、一个未知键，插件照常启动，管理 API 列出 reset 的校正记录与无效路径，prune 后未知键被删除；改写既有测试中的 `$version` 写法，断言不变。
- 本地自测（*_local_test.go，结果写入 PR）：V-132 的各类校正（数字与字符串转换、bool、list 与 template_list 元素）、secret 不记原值、Adapter 实例与 Core 设置的校正、V-37 与 V-88 改写后的分项。

完成条件：gofmt -l 无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...；`grep -rn '\$version\|\$migrations' --include=*.go --include=*.json .` 只剩 DECISIONS 与历史证据中的文字。

不做：WebUI 的提示与清理按钮（goal-44）；配置文件的覆盖值（goal-160）。
```

### goal-160 · T-08.18 配置文件的存储与管理

状态：已完成（PR #214），合并提交 `ab0cd56`。以下保留派发时语句。派发时：前置 T-08.14 与 T-09.14 已合并。2026-11-17 按 D-168 新增；派发与 SDK 合并在 goal-161。与同一波的 goal-152 都改 contracts/configschema/，两者不同时派发，先合并的一条之后再派发另一条。

```text
/goal 使用 $alonabot-execute 推进 goal-160（切片 T-08.18：配置文件的存储与管理）。goal 目录 goal-160，分支 goal-160，worktree .worktrees/goal-160，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在 Core 状态库保存配置文件、会话的指定、字段覆盖值、插件开关、工具策略覆盖与平台设置；按 CONTRACTS §5.6 解析会话所用的配置文件；schema 接受 `profile` 标记；实现 ADMIN-API §23 与 `alonabot profile`。

依据：PLAN T-08 步骤 15；CONTRACTS §5.1（`profile`）、§5.6、§9.3（uninstall）、§15.1（`reply`）；ADMIN-API §23；DECISIONS D-168；TEST-MATRIX V-140。

范围（只新增或修改这些）：
- internal/core/state/schema.sql 与新文件 profile.go（并入初始迁移，不新增迁移）
- 新包 internal/core/profile/（解析与管理服务层）
- contracts/configschema/：`profile` 键的位置检查与按字段路径校验单个值
- internal/core/settings/：`reply` 键的解析、补缺与写回
- internal/core/manage/：ADMIN-API §23 的端点；卸载 `--purge` 与删除 Adapter 实例时清理对应记录
- cmd/alonabot/：profile 子命令
- 上述的测试文件

接口（internal/core/profile）：
  type Profile struct { ID, Name, Color string }
  func (s *Service) Resolve(instance, session string) Profile // 单独指定 → 实例缺省 → default
  func (s *Service) Overrides(pluginID string) map[string]map[string]json.RawMessage // 配置文件 ID → 字段路径 → 值，只含有效项
  以及 ADMIN-API §23 各端点对应的方法（名字自定，不导出到 internal/core 之外）

行为：`default` 不能删除、没有自己的覆盖值与平台设置，平台设置读写 core_config.json 的 command_prefixes 与 reply；写入覆盖值时用该插件 Active 代次的 schema 校验，插件更新后失效的项保留并在 GET 中列为 invalid；会话数按聊天目录计算；复制配置文件不复制会话的指定。本切片只存储与管理，派发、下发与 SDK 合并不在这里。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix）：经 CLI 新建配置文件、给天气类测试插件的标记字段写入覆盖值、指定实例缺省与一个会话，GET 返回对应的指定与覆盖；写入未标记字段返回 400；删除配置文件后指定回到 default。
- 本地自测（*_local_test.go，结果写入 PR）：V-140 的全部分项。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：InvocationContext 与下发、插件开关的派发过滤（goal-161）；命令前缀与回复方式的执行（goal-162）；工具策略覆盖的生效（goal-163）；页面（goal-165）。
```

### goal-147 · T-09.19 诊断包

状态：已完成（PR #213），合并提交 `6f6856c`。以下保留派发时语句。派发时：前置 T-09.10 与 T-09.14 已合并。2026-11-17 按 D-156 新增；页面在 goal-94。

```text
/goal 使用 $alonabot-execute 推进 goal-147（切片 T-09.19：诊断包）。goal 目录 goal-147，分支 goal-147，worktree .worktrees/goal-147，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按 CONTRACTS §14.7 生成诊断包 zip：日志、设置与配置、组件与环境三组内容可以分别取消，缺省脱敏；管理 API `POST /api/v1/diagnostics` 与 `alonabot diagnose` 共用同一实现，Core 未运行时 CLI 只读数据目录生成。

依据：PLAN T-09 步骤 7（诊断包）；CONTRACTS §14.5、§14.6、§14.7；ADMIN-API §19；DECISIONS D-156；TEST-MATRIX V-127 的 Core、API 与 CLI 分项。

范围（只新增或修改这些）：
- 新包 internal/core/diagnostics/（打包与脱敏）
- internal/core/manage/：diagnostics 端点
- cmd/alonabot/：diagnose 命令
- 上述的测试文件

接口：
- package diagnostics：type Options struct { Since time.Duration; Logs, Settings, Components, Redact bool }；func Write(ctx context.Context, w io.Writer, src Source, opts Options) error（Source 是运行中 Core 与只读数据目录的两种实现，名字自定）
- CLI：`alonabot diagnose [--since 24h] [--no-logs] [--no-settings] [--no-components] [--no-redact] [-o 文件]`，不给 -o 时写入当前目录的 `alonabot-diagnostics-<时间>.zip`

行为：根目录 manifest.json 记录格式版本 1、生成时间、时间范围、各组是否包含、是否脱敏与文件列表；不包含状态库、受管 SQLite、媒体、secret 存储、admin_token 与安装包；脱敏规则按 §14.7；生成失败不留下不完整的文件；CLI 在 Core 未运行时不写数据目录，组件运行状态记为未运行。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/logs_e2e_test.go 新增一个子测试（unix）：带一个 secret 配置的测试插件运行后，API 生成的诊断包含三组文件且 secret 为 [REDACTED]；停止 Core 后 diagnose 同样生成，组件为未运行。
- 本地自测（*_local_test.go，结果写入 PR）：§14.7 的每条日志遮盖规则、since 非法、三组都取消、写入失败时不留文件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...。

不做：日志页的导出对话框（goal-94）；上传诊断包到任何服务。
```

### goal-153 · T-09.20 WebUI 组件的后台进程与 command.invoked

状态：已完成（PR #221），合并提交 `bd386df`。以下保留派发时语句。派发时：前置 T-09.12、T-09.14 与 T-04.16 已合并。2026-11-17 按 D-162 新增；官方 WebUI 的后台实现在 goal-156。

```text
/goal 使用 $alonabot-execute 推进 goal-153（切片 T-09.20：WebUI 组件的后台进程与 command.invoked）。goal 目录 goal-153，分支 goal-153，worktree .worktrees/goal-153，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：`role: webui` 的组件可以带 Go 入口，这时它另有一个与插件相同的进程（可以使用配置、受管 SQLite、事件路由、Hook 与 HTTP 路由），HTTP 路由要求管理会话、与静态文件同源；Core 在命令层得到 Handler 结果后生成标准事件 `command.invoked`；插件源 `index.json` 的格式检查。

依据：PLAN T-09 步骤 7（WebUI 进程）；CONTRACTS §7.5、§8（标准事件）、§9.1（WebUI）、§9.2（插件源）；DECISIONS D-162；TEST-MATRIX V-133。

范围（只新增或修改这些）：
- contracts/manifest 与 internal/pluginkit：WebUI 有入口时按插件补全 manifest 并按平台出包，没有入口时照旧 any
- internal/core/manage、internal/core/webhost：WebUI 进程的启停随组件启停；其 HTTP 路由挂在管理源下并要求管理会话
- internal/core/router 或 bus：command.invoked（plugin_id、route_id、command_name、result 为 ok、error 或 denied、duration_ms），经 Event Bus 交给订阅者，不进入事件路由
- sdk/：command.invoked 的事件常量与载荷类型
- contracts/：index.json 的类型与解析（供 goal-156 使用）
- examples/：一个带 Go 入口的最小 WebUI 测试组件（只放在测试夹具中，不作为官方组件）
- 上述的测试文件

行为：WebUI 进程与插件同样经 Bootstrap、REGISTER 与 Activated，崩溃按插件规则恢复；没有入口的 WebUI 行为不变；command.invoked 不包含参数与聊天内容。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix）：安装带入口的 WebUI 测试组件后，未登录请求其路由被拒、登录后成功；发送一条命令后该组件收到 command.invoked。
- 本地自测（*_local_test.go，结果写入 PR）：V-133 的其余分项（denied 与 error 的结果、没有入口的 WebUI、index.json 的非法写法）。

完成条件：gofmt -l 无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...。

不做：官方 WebUI 的统计与商店（goal-156）；Core 的图表 API。
```

### goal-154 · T-09.21 插件数据浏览

状态：已完成（PR #234），合并提交 `3b3ca6f`。以下保留派发时语句。派发时：前置 T-09.14（PR #192，`0cff77c`）、T-08.5（PR #79，`75c16d3`）、T-08.17（PR #212，`4d69978`）均已合并。已按 main `947daa5` 核对：`internal/core/manage` 的 `(*Service).Handler(auth *adminauth.Auth) http.Handler` 统一注册管理路由；`generations_api.go` 已按安装来源选择代次 `db.sqlite` 或开发组件的 `dev.sqlite`；`internal/core/state` 有 `Active(ctx context.Context, pluginID string) (string, error)` 与 `Retained(ctx context.Context, pluginID string) ([]string, error)`。`sdk/sqlitedb/kv.sql` 的 `_alonabot_kv` 使用 `key TEXT PRIMARY KEY`、`value BLOB`、`updated_at INTEGER`，`internal/app/plugindata.go` 准备共用数据目录；`internal/core/candidate/database.go` 已使用 modernc SQLite Backup API。本条在这些入口上补 ADMIN-API §21。

```text
/goal 使用 $alonabot-execute 推进 goal-154（切片 T-09.21：插件数据浏览）。goal 目录 goal-154，分支 goal-154，worktree .worktrees/goal-154，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现 ADMIN-API §21：受管 SQLite 的表与行的浏览、增删改，KV 与数据目录文件的管理，整个插件数据的导出与清空。

依据：PLAN T-09 步骤 7（数据浏览）；CONTRACTS §6.1、§6.3；ADMIN-API §21；DECISIONS D-163、D-165；TEST-MATRIX V-134、V-137 的 API 分项。

范围（只新增或修改这些）：
- internal/core/manage/：新文件 data_api.go、data_db.go、data_kv.go、data_files.go；http.go 的路由登记，generations_api.go 中可共用的代次路径解析
- cmd/alonabot/manage_e2e_test.go；上述包的 *_local_test.go

允许的依赖：标准库与 docs/development.md 已批准的 modernc.org/sqlite、GORM/libtnb SQLite；Core 状态库沿用 state.Store，插件业务库按现有数据库规则访问。沿用 ARCHITECTURE 的依赖方向；不 import sdk/sqlitedb 或 sdk/kv，不新增依赖。

接口（现有签名保持可用）：
- (*manage.Service).Handler(auth *adminauth.Auth) http.Handler
- (*state.Store).GetInstall(ctx context.Context, pluginID string) (state.Install, error)
- (*state.Store).Active(ctx context.Context, pluginID string) (string, error)
- (*state.Store).Retained(ctx context.Context, pluginID string) ([]string, error)
- 新增接口限于以下 HTTP 端点，载荷与错误格式照 ADMIN-API §21、§5；辅助类型和函数保持包内可见：
  GET /api/v1/plugins/{id}/data/tables
  GET、POST /api/v1/plugins/{id}/data/tables/{table}/rows
  PATCH、DELETE /api/v1/plugins/{id}/data/tables/{table}/rows/{rowid}
  GET /api/v1/plugins/{id}/data/kv
  GET、PUT、DELETE /api/v1/plugins/{id}/data/kv/{key...}
  GET /api/v1/plugins/{id}/data/files
  GET、PUT /api/v1/plugins/{id}/data/files/content
  POST /api/v1/plugins/{id}/data/files/mkdir、rename、delete
  GET /api/v1/plugins/{id}/data/export
  POST /api/v1/plugins/{id}/data/clear
  KV 的公开路径语义仍是 §21 的 /kv/{key}；ServeMux 使用尾部通配以保留键中字面的 /。

行为：
- 只作用于受管 SQLite 与 <data_dir>/plugin_data/<id>/。复用管理认证和非 GET 的 CSRF 校验；从安装记录、Active 与 Retained 验证组件和代次，在逐插件管理锁内选定实际文件，防止与更新、回退、卸载及回收交错。安装插件用 plugins/<id>/generations/<generation>/db.sqlite，开发组件沿用 plugins/<id>/dev.sqlite；缺少受管数据库时表与 KV 端点返回 404，保留文件端点的独立可用性。
- generation 缺省为 Active，只允许本插件当前或保留的已提交代次；保留副本以只读连接浏览，写入返回 403。文件目录由各代次共用，generation 不选择历史文件；指定保留代次的 clear 请求拒绝写入，export 的 db.sqlite 选该代次、plugin_data/ 仍取当前共用目录。
- 列出表、行数、列名、类型、主键与 notnull；普通表按 rowid 增删改，WITHOUT ROWID 表、无主键视图与 _alonabot_ 保留表的通用行写入返回 403。_alonabot_kv 通过专用 KV 端点管理，并纳入 clear；迁移表只读且保留。查询按列等值筛选和排序，offset/limit 分页（缺省 50、上限 500），返回 total；BLOB 按 {"$base64":"…"} 往返。表、列与排序标识符先从实际 schema 核对并引用，取值用绑定参数。
- Active 的行与 KV 写入在插件运行中照常允许，共用原数据库文件的 WAL；使用 CONTRACTS §6.1 的连接设置（WAL、FULL、foreign_keys=ON、busy_timeout=5000ms），每个数据库请求一个事务。保留代次只读打开，缺失文件保持缺失。忙等待超时返回 409；约束错误返回 400 并保留 SQLite 英文说明。
- KV 复用 _alonabot_kv，键遵守 CONTRACTS §6.1；列表按键排序、支持字面 prefix 与分页，返回 size、updated_at、total。读取返回 value_base64，合法 UTF-8 时另给 text；写入接受二进制或文本并更新毫秒时间，读删不存在的键返回 404。每次写入记录 INFO 审计日志；表行只记组件、表与 rowid，KV 与文件记录定位信息，取值和文件正文不进日志。
- 文件路径相对数据目录，以 / 分隔；根目录列表可省略 path。拒绝空段、.、..、绝对路径、反斜杠及目录外路径，逐段访问时不跟随符号链接；上传、重命名、删除和导出同样遵守。列表目录在前、按名称排序；下载用 application/octet-stream 与 attachment。上传先在目标目录写临时文件再改名，已有文件且未给 overwrite=1 时 409；删除非空目录须 recursive=true。
- export 通过 SQLite Backup API 从所选数据库取一致快照，zip 中为 db.sqlite 与 plugin_data/；沿用 candidate/database.go 的驱动备份方式，数据库与文件收集错误原样返回。clear 须 confirm=true，否则 400；清空全部可写业务表的行、KV 与数据目录文件，保留表结构、迁移记录与目录本身，沿用每次数据库请求一个事务的规则。数据库与目录清理失败保留实际错误，不声称跨 SQLite/文件系统原子提交。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个真实 Core＋插件子进程＋SQLite 的数据管理集成子测试，复用现有动态构建工程：运行中经认证 API 浏览与写入业务行、KV 和文件，核对插件下一次读取可见；导出 zip 中数据库可打开、文件完整；确认 clear 后数据清空而表结构与迁移记录保留。复用既有测试工程构建方式，不新增 testdata 插件。
- 本地自测（*_local_test.go，结果写入 PR）：V-134、V-137 的其余 API 分项，包括分页/筛选/排序、BLOB 与 null、WITHOUT ROWID/视图/保留表、保留代次只读、无库、约束与 busy 超时、字面含 / 的 KV 键、越界和符号链接、上传覆盖与删除非空目录、CSRF、审计日志无取值；更新/回退/回收与请求交错、清理及 zip 写出失败保留错误。

完成条件：gofmt -l 无输出；go vet ./internal/core/manage/... ./cmd/alonabot/...；go test -race ./internal/core/manage/... ./cmd/alonabot/...；go build ./...。PR 按 TEST-MATRIX §3 写明 V-134、V-137 API 分项的测试位置、命令、平台、结果与未运行项。

不做：页面（goal-44 的数据库视图、goal-156 的存储视图）；外部数据库、自定义 SQL 执行端点、表结构编辑、数据迁移、CLI 数据浏览；不改 .proto，不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-155 · T-09.22 WebUI 需要的补充管理 API

状态：已完成（PR #220），合并提交 `b5e572e`。以下保留派发时语句。派发时：前置 T-09.14、T-07.9、T-03.25 与 T-04.16 已合并。2026-11-17 按 D-164 新增。

```text
/goal 使用 $alonabot-execute 推进 goal-155（切片 T-09.22：WebUI 需要的补充管理 API）。goal 目录 goal-155，分支 goal-155，worktree .worktrees/goal-155，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：实现 ADMIN-API §20 的端点：保留代次列表、静态任务的列出与暂停恢复、Adapter 实例列表、聊天目录查询、更换管理令牌，以及组件列表追加 `pages`（插件声明的页面与按语言的标题）。

依据：PLAN T-09 步骤 7（补充 API）；CONTRACTS §4.3、§7.4、§8；ADMIN-API §20；DECISIONS D-164；TEST-MATRIX V-135。

范围（只新增或修改这些）：
- internal/core/manage/ 与 internal/core/adminauth/：§20 的端点
- sdk/：路由的 `.Page(title)`，注册描述携带 page_title；contracts/proto 按 §5A 追加字段、生成产物与 wirespec_test.go
- contracts/i18n：`pages` 文案
- 上述的测试文件

接口：sdk 的 HTTP 路由构建器增加 `Page(title string)`（title 为英文源文案，i18n 覆盖在 `pages` 下）。

行为：按 ADMIN-API §20 的请求与响应；更换令牌先按 §12.2 再认证，删除全部管理会话（包括当前会话）；聊天目录查询支持 q、adapter_id、chat_type 与分页。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix）：暂停一个静态任务后到期不触发、恢复后触发；更换令牌后旧会话被拒、新令牌可登录。
- 本地自测（*_local_test.go，结果写入 PR）：V-135 的其余分项。

完成条件：gofmt -l 无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...。

不做：页面（goal-43、goal-44）；数据浏览（goal-154）；changelog（goal-158）。
```

### goal-158 · T-09.24 组件的 changelog

状态：已完成（PR #215），合并提交 `b911f27`。以下保留派发时语句。派发时：前置 T-09.3 与 T-09.14 已合并。2026-11-17 按 D-166 新增。

```text
/goal 使用 $alonabot-execute 推进 goal-158（切片 T-09.24：组件的 changelog）。goal 目录 goal-158，分支 goal-158，worktree .worktrees/goal-158，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：组件可以带 `CHANGELOG.md`（Keep a Changelog）或 `changelog/<版本>/<locale>.md`；check、build 与 pack 检查并打包；管理 API 按已安装的版本提供组件与 Core 自身的 changelog。

依据：PLAN T-09 步骤 7（changelog）；CONTRACTS §9.1；ADMIN-API §22；DECISIONS D-166；TEST-MATRIX V-138。

范围（只新增或修改这些）：
- internal/pluginkit/：检查与打包
- internal/core/packages/preflight.go：随包校验
- internal/core/manage/：两个 changelog 端点
- 仓库根目录 CHANGELOG.md（新建，Keep a Changelog，[Unreleased] 一节）与把它编进可执行文件的位置
- examples/plugins/showcase：加一个 CHANGELOG.md
- 上述的测试文件

行为：两种形式同时出现、文件不是 UTF-8、版本目录名不合规或高于 manifest 的 version、缺少 en-US.md 时报错并指出路径；`[Unreleased]` 不返回；按请求语言选择，缺少时回落到 en-US；Markdown 原样返回，Core 不渲染。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个子测试（unix）：showcase 装上后 changelog 端点按版本返回条目，不含 [Unreleased]。
- 本地自测（*_local_test.go，结果写入 PR）：V-138 的检查分项与目录形式的语言回落。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：显示与「已看过」的记录（goal-156，Core 不记录）。
```

### goal-148 · T-11.5 工具的副作用类别

状态：已完成（PR #219），合并提交 `b2348c9`。以下保留派发时语句。派发时：前置 T-11.2、T-12.3 与 T-09.14 已合并。2026-11-17 按 D-157 新增；页面在 goal-44。

```text
/goal 使用 $alonabot-execute 推进 goal-148（切片 T-11.5：工具的副作用类别）。goal 目录 goal-148，分支 goal-148，worktree .worktrees/goal-148，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：工具可以声明副作用类别 read_only、write 或 destructive（不声明为 unspecified）；注册、目录、确认记录与管理 API 带上类别；管理 API 的策略修改可以一次提交多项；官方 Agent 按 MCP 工具注解换算类别。类别不参与策略判断。

依据：PLAN T-11 步骤 7；CONTRACTS §13.1；WIRE-FORMAT §3.3、§3.5；ADMIN-API §18；DECISIONS D-157；TEST-MATRIX V-128 的 Core、SDK、Agent 与 API 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/tool.proto：ToolEffect（UNSPECIFIED、READ_ONLY、WRITE、DESTRUCTIVE）与 ToolSpec、ToolInfo、ToolConfirmRequest 的 `effect`；生成产物；contracts/wirespec_test.go
- internal/core/tool/、internal/core/state/tool.go：保存类别，confirm 记录操作的类别（同名以最近一次为准）
- internal/core/manage/tool_api.go：目录返回类别，PUT 接受 `{"items":[...]}` 并在一个事务内写入
- sdk/tool/tool.go 与 sdk/service/service.go：类别常量与构建器方法
- agent/mcp.go：按注解换算类别并在 confirm 中带上
- 上述的测试文件

接口：
- package tool：type Effect int；const ( Unspecified Effect = iota; ReadOnly; Write; Destructive )
- package service：func (s *Builder[In, Out]) Effect(e tool.Effect) *Builder[In, Out] // 在 Tool 之前调用时 Collect 报错
- SDK 的确认调用增加类别参数（随现有 confirm 函数追加一个 tool.Effect 参数，同步全部调用方）

行为：MCP 注解换算：没有 annotations 为 unspecified；readOnlyHint 为 true 为 read_only；否则 destructiveHint 为 false 为 write，为 true 或缺省为 destructive。非法枚举值使注册失败。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tools_test.go 新增一个子测试：带三种注解的 MCP 服务器经 confirm 记录的类别符合换算；cmd/alonabot/manage_e2e_test.go 改写既有的工具策略子测试为批量 PUT，断言目录中的类别。多列理由：Agent 的换算与 Core 的批量写入在两个进程中。
- 本地自测（*_local_test.go，结果写入 PR）：V-128 的其余分项（Effect 在 Tool 之前、非法枚举、批量中一项不合法时整批不写）。

完成条件：gofmt -l 无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...。

不做：权限页的按类别显示与批量设置（goal-44）；按类别自动决定策略。
```

### goal-140 · T-13.5 模块根包只保留库模式接口

状态：已完成（[PR #200](https://github.com/AlonaBot/AlonaBot/pull/200)，合并提交 `2e5ac2b`），证据见 TEST-MATRIX §3.2。合入：`internal/app` 的 `Options`、`Component`（字段不导出，由 `PluginComponent`、`AdapterComponent`、`ExecutableComponent` 构造并保留角色）、`Start(ctx, Options)` 与 `Core`（`Wait`、`Restart`、`RestartInstance`、`ReloadSettings`）；根包只留 `Run`、`Start`、`*Core`、`Wait` 与 `Plugin`、`Adapter`、`WebUI`、`DataDir`、`Args`、`Listen`、`LogLevel`、`LogFormat`，`type Option func(*app.Options)`；cmd/alonabot 改用 `internal/app`，不再 import 根包；装配、内部服务、候选与 WebUI 的 9 个集成测试移入 `internal/app`。审查追加的修复：语句原定的 `Component` 导出字段分不出 nil Adapter 与 nil Plugin，改为构造函数，两者的错误文本与移动前相同（FIX-2）；下面语句中的接口段已按此修订。

```text
/goal 使用 $alonabot-execute 推进 goal-140（切片 T-13.5：模块根包只保留库模式接口）。goal 目录 goal-140，分支 goal-140，worktree .worktrees/goal-140，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：模块根包 alonabot 只保留 CONTRACTS §15.5 的库模式接口——Run、Start、*Core 与 Wait，选项 Plugin、Adapter、WebUI、DataDir、Args、Listen、LogLevel、LogFormat——每个都是对新包 internal/app 的薄包装。Core 装配、候选与进程管理，以及只供开发 CLI 使用的 Executable、Stdio、Core.Restart、Core.RestartInstance 与 Core.ReloadSettings 移到 internal/app，cmd/alonabot 直接使用。行为与协议不变，只改包与调用方式。

依据：PLAN T-13 步骤 6 及其失败停止条件；CONTRACTS §15.4、§15.5；DECISIONS D-148；TEST-MATRIX V-64、V-65；ARCHITECTURE 的根包与 internal/app 行。

范围（只新增、修改或删除这些）：
- 根目录的 alonabot.go、assembly.go、agent_target.go、journal.go、version.go：除下文「根包」列出的库模式接口外，全部代码移入 internal/app（文件用普通 mv 后改包名，不做任何 git 写操作）；根包只留 alonabot.go 一个非测试文件
- 根目录测试 admin_upload_test.go、agent_service_test.go、journal_test.go、pipeline_test.go、scheduler_sdk_test.go、scheduler_tasks_test.go、service_chain_test.go、t11_e2e_test.go、webui_test.go 移入 internal/app；alonabot_test.go（TestLibraryLifecycle，只用公开接口）留在根包
- 新增 internal/app/
- cmd/alonabot/：commands.go 与 dev.go 改用 internal/app；随之改写的 cmd/alonabot/*_test.go
- docs/development.md、docs/plugin-guide.md、README.md 中提到这些能力或测试位置的写法
- 上述包的 *_local_test.go
允许的新跨模块边：根包 -> internal/app；cmd/alonabot -> internal/app；internal/app 承接原根包已有的依赖（-> internal/core/...、-> sdk/...、-> contracts/...、-> internal/generated、-> internal/inproc、-> internal/tune、-> internal/wire），其测试承接原根包测试的依赖（-> agent/...、-> internal/generate、-> internal/pluginkit、-> internal/oslock）。根包只 import internal/app、sdk、sdk/adapter 与标准库；cmd/alonabot 不再 import 根包；internal/core 仍不 import sdk 与 internal/app。

接口（照此实现）：
- internal/app：
  // Options 汇总一次运行的装配参数；指针字段为 nil 时不覆盖 core_config.json。
  type Options struct {
      DataDir    string        // 原样交给 filepath.Abs；不设缺省值
      Args       []string      // 调优参数与参数文件，后于 ALONABOT_ARGS 应用
      Components []Component   // 开发组件，按登记顺序
      Listen     *string
      LogLevel   *slog.Level
      LogFormat  *string
      Stdio      bool          // Executable 组件使用 stdio 传输
      WebUIs     []fs.FS       // 开发 WebUI；多于一个或含 nil 时返回原有的用法错误
  }
  // Component 保留开发组件的角色与装配资源，字段不导出（审查 FIX-2 修订）。
  type Component struct{ ... }
  func PluginComponent(root *sdk.Plugin) Component
  func AdapterComponent(a adapter.Adapter) Component
  func ExecutableComponent(path string, packageManifest []byte) Component // 复制 packageManifest，角色按包内 manifest 确定
  func Start(ctx context.Context, o Options) (*Core, error)
  type Core struct{ ... }
  func (c *Core) Wait() error
  func (c *Core) Restart(ctx context.Context, id string) error
  func (c *Core) RestartInstance(ctx context.Context, adapterID, instanceID string) error
  func (c *Core) ReloadSettings(ctx context.Context) error
- 根包（签名、文档注释与行为不变，函数体只做转发）：
  type Option func(*app.Options)
  func Plugin(root *sdk.Plugin) Option；func Adapter(a adapter.Adapter) Option；func WebUI(fsys fs.FS) Option；func DataDir(dir string) Option；func Args(values ...string) Option；func Listen(addr string) Option；func LogLevel(level slog.Level) Option；func LogFormat(format string) Option
  type Core struct{ ... }（只持有 *app.Core）；func (c *Core) Wait() error
  func Start(ctx context.Context, opts ...Option) (*Core, error)：以 app.Options{DataDir: "data"} 为起点依次应用选项，再调用 app.Start
  func Run(opts ...Option) error：保留现有的 SIGINT、SIGTERM 处理，调用 Start 与 Wait
  删除根包的 Executable、Stdio 与 Core 的 Restart、RestartInstance、ReloadSettings，不留别名或转发。

行为：
- 库模式、alonabot run 与 run --reload 的启动、组件检查、错误文本、退出码、日志、停止与热部署行为都与移动前相同：ALONABOT_ARGS 先于 Args 应用；根包缺省数据目录为 data，CLI 传入 --data-dir 的值；WebUI 重复登记与 nil 文件系统的用法错误、Listen 的安全检查、开发组件与已安装包的身份冲突都按原次序报告。
- 根包的 Start 返回的错误与 app.Start 相同（包括 context.Canceled），不另行包装。
- plugintest、examples/devbot 与 agent 测试继续经根包的 Start 运行，代码不改。cmd/alonabot/plugin_cli_e2e_test.go 中写给作者工程的库模式 main 仍 import 根包。
- 版本号（coreVersion、moduleVersion）随装配移入 internal/app，仍按模块路径 github.com/AlonaBot/AlonaBot 读取构建信息。

测试（按编码规范「测试提交规则」）：
- 提交：不新增测试。9 个根目录测试文件移入 internal/app，只改包名、装配写法（选项函数改为 Options 字面量或测试内的辅助函数）与 import，断言、子测试与超时不变；TestMain 按原环境变量重入组件进程的方式不变。根包保留 TestLibraryLifecycle；cmd/alonabot 的 TestDevelopmentCLI 与 dev 端到端测试覆盖 CLI 经 internal/app 的开发运行与热部署。
- 本地自测（*_local_test.go，结果写入 PR）：根包 `go doc` 的导出集合只剩上文「根包」列出的标识符；`go list -deps ./cmd/alonabot` 不含根包、`go list -deps ./internal/core/...` 不含 internal/app 与 sdk；移动后 internal/app 的测试函数与子测试名称集合与移动前根包的相同（以 main 为基准，排除 TestLibraryLifecycle）。

完成条件：本地证据齐备；gofmt -l 对改动的 Go 文件无输出；go generate ./... 之后工作区没有变化；go vet ./...；go test -race ./...；go build ./...；CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...；`go doc .` 只列出库模式接口。行为发生变化、出现循环依赖，或 internal/core 需要 import sdk 或 internal/app 时，按 PLAN T-13 的停止条件暂停交付。

不做：库模式接口的增删；协议与行为的改动；CLI 位置参数与缺省值（T-13.6）；新增依赖；在根包新增测试；改写 TEST-MATRIX 中已有的证据行（文件移动后的链接由审查者合并前同步）。
```

### goal-170 · S-02.1 按需唤醒、活动引用与有界等待

状态：已完成（PR #241），合并提交 `e0a38d7`。以下保留派发时语句。派发时：前置 S-01.1（PR #225，`2cc32ff`）、S-01.2（PR #226，`413a6ff`）、T-05.5（PR #162，`8c12972`）、T-06.4（PR #164，`039802f`）、T-07.5（PR #119，`cc19e99`）均已合并；S-02.3（PR #227，`86e6768`）的 Core 消息过滤也已合并。已按 main `947daa5` 核对：`internal/core/supervisor.Launch(ctx context.Context, store *state.Store, spec supervisor.Spec) (*supervisor.Running, error)` 与 `(*Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult` 共用逐插件锁，`runtime.go` 按实际启动所有权更新状态；`internal/core/state.ActiveRouteSnapshot` 与 `internal/core/router.SetRouteSnapshot/RouteSnapshot` 已保存和读取已提交路由，但 `Router.candidates`、Broker 与 HTTP 代理仍依赖活 Session。`router.Options.MessageConfig` 与 `matchMessageFilter` 已完成消息过滤，`internal/app/message_filter.go` 当前只读 Active 基础配置。本条接入快照唤醒与活动引用；与 goal-161 共享 Router 的 Options/Invocation、internal/app 装配及消息配置读取，后合并者同步 main，保留配置文件的会话解析与覆盖，不回退为基础配置读取。

```text
/goal 使用 $alonabot-execute 推进 goal-170（切片 S-02.1：按需唤醒、活动引用与有界等待）。goal 目录 goal-170，分支 goal-170，worktree .worktrees/goal-170，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在现有路由快照与运行状态上接入 Supervisor 的 EnsureActive、活动引用、并发唤醒合并与有界等待，保留请求的原始权限、Target、Trace、顺序与期限。

依据：PLAN §7 S-02 步骤 1～3 及其失败停止条件；CONTRACTS §2.5、§4.3、§4.4、§4.5、§12.2、§12.3、§16；WIRE-FORMAT §3.1、§4；DESIGN §A；DECISIONS D-53、D-54、D-153、D-173 及 2026-11-17 维护者确认的休眠补齐规则；TEST-MATRIX V-77 的首请求、并发、取消、期限与活动引用分项。

范围（只新增或修改这些）：
- internal/core/supervisor/：新文件 wake.go、activity.go，supervisor.go、prepare.go、runtime.go 的启动、活动登记、停止门禁与退出清理调用点
- internal/core/router/：快照候选、命令索引、派发、消息过滤接入位置与活动包装；internal/core/broker/ 的快照定位、唤醒准入与交付复核
- internal/core/httpproxy/：快照路径匹配、唤醒与活动引用；internal/core/media/ 的操作/租约活动登记；internal/core/continuation/ 的延续与等待活动登记；internal/core/hook/、bus/ 中已接纳工作的引用调用点
- internal/app/：assembly.go、alonabot.go、journal.go、message_filter.go 与按需唤醒的新装配文件，覆盖已安装独立进程的启动、停止竞争、绑定与退出监视
- internal/tune/params.go：追加 WAKE_QUEUE_REQUESTS、WAKE_QUEUE_BYTES 及 Values 字段
- contracts/proto/alonabot/v1/control.proto：仅追加 STOP_REASON_IDLE；相应 contracts/pb/alonabotv1/ 生成代码与 contracts/wirespec_test.go。该协议修改明确授权
- internal/app/wake_test.go（新集成测试）；上述包的 *_local_test.go

允许的依赖：标准库及现有已批准依赖；supervisor、router、broker 等沿用 ARCHITECTURE 的 internal/core 内部依赖，internal/app 装配；队列复用 internal/wire.Budget，限额复用 internal/tune。不增加依赖，不新增业务 RPC。

接口（现有签名保持可用）：
- supervisor.Launch(ctx context.Context, store *state.Store, spec supervisor.Spec) (*supervisor.Running, error)
- supervisor.Prepare(ctx context.Context, store *state.Store, spec supervisor.Spec) (*supervisor.Prepared, error)
- (*supervisor.Prepared).Activate(ctx context.Context, commit func(context.Context) error) (*supervisor.Running, error)
- (*supervisor.Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult；WaitExit() (abnormal bool, state string, err error)
- (*state.Store).ActiveRouteSnapshot(ctx context.Context, pluginID string) (state.RouteSnapshot, error)；RuntimeStatus(ctx context.Context, pluginID string) (state.RuntimeStatus, error)；SetRuntimeStatus(ctx context.Context, pluginID string, status state.RuntimeStatus) error
- (*router.Router).SetRouteSnapshot(pluginID, generationID string, routes []*pb.RouteDescriptor)；RouteSnapshot(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool)；PluginFor(c router.Caller) (router.Plugin, bool)
- router.Options.MessageConfig func(ctx context.Context, pluginID string, event *pb.Event) (json.RawMessage, error)
- (*httpproxy.Server).Bind(pluginID, generationID string, endpoint *pb.HttpListener, token []byte, routes []*pb.RouteDescriptor, role state.Role)；Drain(pluginID string, deadline time.Time)；Unbind(pluginID, generationID string, token []byte)
- (*wire.Budget).Acquire(n int64) bool；Release(n int64)；wire.Request.Deadline time.Time、Late func(wire.Response) 保留原有实际完成/迟到响应语义

接口（本切片新增，以下签名固定）：
- package supervisor：type WakeOptions struct { Store *state.Store; Global *wire.Budget; QueueRequests int; QueueBytes int64; Start func(context.Context, string) (*Running, error) }；type Waker struct（字段不导出）；func NewWaker(opt WakeOptions) *Waker；supervisor.Spec 追加 Waker *Waker，由独立插件装配传入
- func (w *Waker) EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)：合并启动并取得引用；Start 复用 internal/app 的原代次启动与完整绑定，绑定就绪后才成功返回
- func (w *Waker) Ready(running *Running)：supervisor 在开放 HTTP/Scheduler 前登记本次 Running/Session 的活动所有权，app 完成 Router/Bus 等绑定后调用 Ready；EnsureActive 等待该次绑定就绪，已有激活、候选提交与唤醒共用此入口
- func (w *Waker) AcquireActivity(pluginID string, session *wire.Session) (release func(), err error)：只为匹配的当前活 Session 取得引用，供已接纳工作登记，不触发唤醒；释放函数只释放自己持有的一次引用
- router.Options、broker.Options 各追加 EnsureActive func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)，均注入同一 Waker；原活 Session 路径在未装配该回调时照常使用
- func WithActivity(p Plugin, acquire func() (release func(), err error)) Plugin（package router）：为实际 Call/SendEvent 登记活动，保留 Plugin 的原身份、路由与调用结果，组合原 wire.Request.Late 回调；supervisor 的 Scheduler 绑定和 app 的 Router/Hook/Broker 绑定复用这一包装
- func (s *Server) SetWake(snapshot func(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool), ensure func(context.Context, string, int64) (func(), error))（package httpproxy）：装配已提交路由读取与同一唤醒入口
- continuation.Options、bus.Options 各追加 AcquireActivity func(c router.Caller) (release func(), err error)；func (s *Store) SetActivity(acquire func(c router.Caller) (release func(), err error))（package media）：装配方把 Caller 核对为当前 Session 后调用 Waker.AcquireActivity
- tune.Values 追加 WakeQueueRequests int、WakeQueueBytes int64，登记 WAKE_QUEUE_REQUESTS=256（≥1）、WAKE_QUEUE_BYTES=16MiB（≥4MiB 且 ≤CORE_MAX_PAYLOAD）；从本次 Core 的生效 Values 注入，不在业务调用点另写缺省值
- StopReason 在现有值之后追加 STOP_REASON_IDLE，当前 main 的下一号为 6；合并前同步最新源文件，已有编号保持不变，重新生成

行为：
- Waker 是每个 Core 唯一的唤醒队列与活动引用所有者，生命周期执行仍由 Supervisor 的 Launch/Prepare/Activate/Stop/WaitExit 完成。只从已启用、已提交且运行状态为 SUSPENDED 的独立 plugin/agent 自动唤醒；WAKING 的请求加入同次启动，ACTIVE 取得引用后交付；停用、STOPPED、FAILED 或隔离状态返回实际错误。Core、Adapter、WebUI 与同进程开发组件沿用原常驻生命周期。
- Router 根据已提交快照构建无进程候选和命令索引，维持 priority/plugin_id/route_id 稳定顺序、命令名覆盖、命令消费与权限规则；快照用于定位，实际交付仍取得新 Caller 与 Activated Plugin。Service Broker 在无活 Caller 时先从快照找 Service，保留访问规则、Schema 标识/版本、编码、调用链和 @agent 解析；初步准入通过后才唤醒。D-173：Service 只取快照已有的 Schema 标识/版本，工具输入 Schema 沿用 REGISTER 原文，配置 Schema 从绑定包资源核验；不另存 Service Schema 原文。
- HTTP 入口先用快照匹配方法/公共路径并执行现有公共/管理认证、CSRF 与隔离，再 EnsureActive；未匹配和未授权请求保持原错误。ACTIVATED 后重新读取当前 Listener/代理凭证并转发一次，HTTP 请求、流与升级连接按实际存续持有引用。等待期间不预读整份流式正文；已经缓存在 Core 内的请求载荷按实际字节计入等待预算，沿用原 HTTP 错误映射。
- 消息事件在 PreEvent 后、取得活动引用与 EnsureActive 前复用 S-02.3 的 matchMessageFilter：未命中不交付、不取得唤醒名额，读取失败保留诊断并跳过该路由；命中后仍执行 SDK Rule。常驻与快照候选共用同一匹配入口。goal-161 先合并时保留其按会话 Resolve 与有效 Overrides（含祖先 object）的 MessageConfig；本条先合并时保留现有注入点，由 goal-161 接入会话覆盖，禁止另起配置读取真源。
- 同一插件 100 个并发请求只建立一个新 Session。共享启动使用 Core 生命周期与原阶段预算，单个等待者取消只撤销自己的等待与预算，不取消其他等待者的启动。等待期间不持 Router/Broker/HTTP 映射锁或占着普通业务槽互相等待；逐插件管理锁、runtime.lock、退出树清理与有限重启继续由原所有者执行。
- 每插件等待队列同时限制 256 条与 16 MiB（取生效 tune 值），并计入同一个 Core 总 payload 预算，任一满额立即 RESOURCE_EXHAUSTED 并记录原因。按实际等待载荷取得预算，每份预留由同一所有者释放一次；取消或到期的未交付请求立即移除并释放名额与字节。保留原请求、可信 Actor、Target、Trace、顺序与绝对 deadline。
- 在开始等待前固定原调用期限；普通业务沿用 REQUEST_TIMEOUT 与父期限的较早者，交付时显式传原 wire.Request.Deadline。冷启动计入全部原预算；Scheduler 的 JOB_TIMEOUT 同样从到期入口算起、缺省 15 分钟，后继 goal-172 接入时沿用该绝对期限。EnsureActive 返回就绪后仍复核当前代次、当前 Caller、路由开关、权限与剩余期限，失败结束本次请求，原请求至多交付一次。
- 唤醒复用原 ActiveGeneration、Config、DB 与 activation_reason=resume；复用现有快照/安装身份检查与 REGISTER 摘要校验，通过 ACTIVATED 后才开放交付。摘要失配关闭候选、隔离路由并返回 FAILED_PRECONDITION；唤醒失败只返回实际错误，同时写 FAILED/原因，不返回伪造成功，不建立新候选、不重跑数据迁移、不重放已交付副作用。后继策略写回成功但唤醒失败时也按 ADMIN-API §5 只返回实际错误，由调用方 GET 核对已保存设置。
- 活动引用覆盖 Handler、Service、HTTP、媒体操作与租约、已接纳 Hook/Bus、延续调用、多轮等待及插件主动出站业务。Router 的通用 Plugin 包装同时覆盖当前 Scheduler 已运行的回调；到期唤醒入口留给 goal-172。Bus 只为已在线订阅者接纳的排队/写出工作持引用，离线仍丢弃、不积压、不唤醒。媒体的引用随真实租约、过期或 Session 释放；延续与等待在登记时取得，结束、取消或 Session 失效时释放。
- 实际业务响应/失败、取消后确认未交付、资源实际结束或旧进程树确认退出才释放相应引用；本地超时先返回时，经 wire.Request.Late 或原 Session/进程退出清理继续持有，不能仅因调用者收到超时就把插件记为空闲。插件出站在途调用也登记引用；迟到响应、旧 Session 清理和重复释放不得扣掉新 Session 的引用。
- 活动取得与停止共用逐插件生命周期门禁，供 goal-171 复用：空闲停止在发送 STOPPING 前可被新请求撤销；发送后新请求进入有界等待，确认旧进程树退出后再建立唯一新 Session。管理停用/卸载和 Core 停机关闭新接纳，沿用原排空与停止预算；活动引用阻止空闲退出，原管理强制停止仍按契约完成清理。
- 记录实际唤醒完成的 LastWakeAt/原因、真实活动结束时间及冷启动发生时间，供 goal-171 的空闲判定使用。既定规则保持：唤醒后最短运行 5 分钟；SLEEP_COLD_START_WINDOW 缺省 1 小时，窗口内冷启动超过 SLEEP_COLD_START_MAX（缺省 3 次）时，从触发时刻固定常驻 1 个窗口的冷却期；期间新冷启动不延长这次冷却期，结束后恢复空闲判定。SLEEP_* 参数登记、准入与空闲停止由 goal-171 交付，本条只追加 WAKE_QUEUE_*。

测试（按编码规范「测试提交规则」）：
- 提交：internal/app/wake_test.go 一个真实 Core＋已提交插件子进程＋SQLite 的集成测试，复用现有动态构建工程与回环 Adapter。测试装配将插件按正常停止流程置为 SUSPENDED，分别用首条命令、Service 与 HTTP 驱动同一唤醒主路径，核对只交付一次、Session 已换、原 generation/config/DB 保持，真实活动结束后释放引用；此测试不冒充 goal-171 的自动空闲策略验收。
- 提交：contracts/wirespec_test.go 固定 STOP_REASON_IDLE 追加编号且既有值不变。多列理由：授权修改控制协议，须独立核对编号，主路径进程测试无法证明旧编号稳定。
- 本地自测（*_local_test.go，结果写入 PR）：100 路并发只启一次、条数/字节/Core 预算分别满额、取消和 deadline 到达、首等待者取消而其他请求继续、唤醒失败及摘要失配、等待中权限/代次/路由变化、STOPPING 前后竞争、SDK Rule 前的 Core 过滤未命中/读取失败/会话覆盖衔接、Handler/Service/HTTP/媒体/schedule/Hook/Bus/延续/多轮等待/出站活动、超时后迟到响应与旧树退出、迟到清理不影响新引用；always_on、dev、Adapter/WebUI 原路径。supervisor/ 与 sdk/ 的新增测试全部写本地，不新增 testdata 插件。

完成条件：go generate ./... 后再次生成无差异；gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/wire/... ./internal/tune/... ./internal/core/supervisor/... ./internal/core/router/... ./internal/core/broker/... ./internal/core/httpproxy/... ./internal/core/media/... ./internal/core/continuation/... ./internal/core/hook/... ./internal/core/bus/... ./internal/app/...；go build ./...。PR 按 TEST-MATRIX §3 写明 V-77 本切片分项的测试位置、命令、平台、结果与未运行项。

不做：Idle Controller、休眠准入与策略写回/自动空闲停止（goal-171），Scheduler 到期唤醒（goal-172），休眠管理 API/CLI/WebUI 与固定机器测量；新增 Service Schema 包资源、业务 RPC 或自动重试；不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-176 · S-02.3 Core 消息事件过滤与 SDK 注册

状态：已完成（PR #227），合并提交 `86e6768`。以下保留派发时语句。派发时：前置 T-04.16（PR #199）、T-08.11（PR #103）、T-09.14（PR #192）均已合并。2026-11-17 按维护者决定新增；已对照合入最新 main 的当前工作树（HEAD `0d4cc26`）核对：EventSpec 只有类型、Handler、Rule 与按钮字段；Router.candidates 只按事件类型收集活 Session，dispatch 中执行权限与调用；sdk.EventBuilder.Rule 只在插件进程执行，buildRegister 在 sdk/session.go。Core 已持有可信 schema 和 Active 配置，state.Store.GenerationConfig 可读已提交快照；T-08.18/T-04.18 的配置文件实现尚未合并。本切片交付配置快照读取注入点及基础实现，后合并者按共享规则接入配置文件；无需等待任何休眠生命周期切片。

```text
/goal 使用 $alonabot-execute 推进 goal-176（切片 S-02.3：Core 消息事件过滤与 SDK 注册）。goal 目录 goal-176，分支 goal-176，worktree .worktrees/goal-176，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：插件与 Agent 能声明由 Core 在交付前匹配的消息事件过滤，按 to_me、会话类型或配置字段中的唤醒词任一命中交付；SDK 继续执行自定义 Rule。先在常驻模式完成注册、可信校验、Core 匹配与配置读取接口，供 Agent 与休眠唤醒复用同一判断。

依据：PLAN §7 S-02 步骤 5；CONTRACTS §2.5、§4.4、§4.5、§5.6；WIRE-FORMAT §3.3；MESSAGE-MODEL §1.1；DECISIONS D-167、D-168 及 2026-11-17 的本轮休眠记录；TEST-MATRIX V-77 的 Core 过滤分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/route.proto：导入 message.proto，在 EventSpec 末尾追加 message_filter，新增 MessageFilter；本切片获授权改这份 proto，字段号按最新 main 顺延，随后 go generate ./...，只保留相关生成变化
- contracts/validate/register.go 与新文件 message_filter.go：消息事件类型、过滤器结构、枚举与路径语法检查；contracts/configschema/ 新文件 message_filter.go：根据可信 schema 检查引用字段；contracts/wirespec_test.go 追加编号检查
- sdk/ 新文件 message_filter.go、events.go 与 session.go：SDK 自有声明、builder 收集、REGISTER 编码；不把进程内 Rule 改成另一条 RPC
- internal/core/host/prepare.go：Expect 增加可信配置 schema，REGISTER 校验时检查字段引用；internal/core/supervisor/supervisor.go 与 internal/app/assembly.go 的 host.Prepare 调用点传入安装包或生成资源的 schema
- internal/core/router/ 新文件 message_filter.go、router.go、dispatch.go：交付前匹配、配置快照读取注入点与错误诊断；不在 Router 锁内做配置 I/O
- internal/app/ 新文件 message_filter.go 与 assembly.go、journal.go、alonabot.go 中必要的装配/代次发布调用点：读取当前 Active 配置；覆盖已安装及同进程开发组件，配置候选失败时保留旧生效来源
- 若 S-01.2 已合并：其 internal/core/state/routes.go、internal/core/router/snapshots.go 与 supervisor 的快照白名单调用点，只补新字段的保存/摘要/恢复；若 T-04.18 已合并：internal/app 中复用其会话解析和覆盖值注入；若 S-02.1 已合并：在 Router 的既有活动引用及 EnsureActive 之前接入同一匹配入口
- internal/app/message_filter_test.go、pipeline_test.go 中现有测试宿主的必要注册；上述包的 *_local_test.go

允许的新跨模块边：无。沿用 sdk -> contracts、internal/core -> contracts、internal/app -> internal/core 与 contracts；通过函数注入读取配置，Router 不依赖 internal/app、SDK 或 Agent。

接口（已核对的现有签名，保持可用）：
- func (r *sdk.Plugin) On(routeID string, eventType sdk.EventType) *sdk.EventBuilder
- func (b *sdk.EventBuilder) Rule(rules ...sdk.Rule) *sdk.EventBuilder；func (b *sdk.EventBuilder) Handle(h any)
- func router.New(opt router.Options) *router.Router；func (r *router.Router) Deliver(ctx context.Context, ev *pb.Event) error
- func validate.Register(r *pb.Register) error
- func configschema.Parse(raw []byte) (*configschema.Schema, []configschema.Problem)
- func (s *state.Store) GenerationConfig(ctx context.Context, generationID string) (string, error)
- func host.Prepare(ctx context.Context, conn *wire.Conn, exp host.Expect, boot *pb.Bootstrap, lim host.Limits) (*host.Prepared, error)

接口（本切片新增，以下签名固定）：
- proto：EventSpec 追加 MessageFilter message_filter；MessageFilter 含 bool to_me、repeated ChatType chat_types、string wake_words_config、string triggers_config。沿用 message.proto 的 ChatType；不预分配编号。
- sdk：type MessageFilter struct { ToMe bool; ChatTypes []message.ChatType; WakeWordsConfig string; TriggersConfig string }
- func (b *EventBuilder) Filter(filter MessageFilter) *EventBuilder // 只设一次，在 Handle 前调用；复制切片；一个条件都没有的过滤器在 Collect 时拒绝
- contracts/configschema：func (s *Schema) CheckMessageFilter(filter *pb.MessageFilter) error // 检查类型及非 secret，错误只带字段路径
- host.Expect 追加 ConfigSchema *configschema.Schema // 来自已验证资源；无配置时 nil，存在配置引用时须校验失败
- router.Options 追加 MessageConfig func(ctx context.Context, pluginID string, event *pb.Event) (json.RawMessage, error) // 返回本次消息会话的生效配置只读快照，供本次匹配按路径取值，不发插件 RPC
- 注册示例：p.On("chat", sdk.MessageCreated).Filter(sdk.MessageFilter{TriggersConfig:"triggers", WakeWordsConfig:"wake_words"}).Rule(sdk.NotCommand()).Handle(handler)。静态私聊或群触发示例：Filter(sdk.MessageFilter{ChatTypes: []message.ChatType{message.ChatPrivate, message.ChatGroup}})。声明只含路径与条件，配置值不从 p.Config() 固化进 REGISTER。

行为：
- 按 CONTRACTS §2.5 作 OR：静态 to_me、任一 ChatTypes 或任一唤醒词命中即交付；TriggersConfig 模式按 mention/private/wake_word 多选，mention 限 GROUP 且 to_me，空列表不交付。triggers 模式与静态 ToMe/ChatTypes 互斥。文本仅顺序拼接 text 片段，非空词位置不限、拉丁字母不区分大小写；不修改 chain。未提供 to_me/文本的消息条件不命中，未知 chat_type 不命中类型条件。
- SDK 和 Core 依次校验各自信任边界：Filter 只允许一次且在 Handle 前；Core 拒绝非消息事件、混合其他事件、非法会话枚举和路径、schema 缺失/字段类型错误/secret 路径。SDK Rule 不决定 Core 侧准入；无过滤器保持原交付；一个条件都没有的过滤器由 SDK 与 Core REGISTER 拒绝（CONTRACTS §2.5）。注册或读取错误保留实际错误，日志省略配置值。
- 在 PreEvent 后、业务交付及唤醒之前判断过滤；继续执行原权限、路由开关、稳定顺序和 command_consumed 规则。与 S-02.1 的后合并者保证在同一判断通过后才取得活动引用并 EnsureActive，未命中不取得唤醒名额。
- MessageConfig 的基础实现一次读取当前已提交 GenerationConfig 并返回只读快照，Router 只按声明的点分路径取字段；候选/重启恢复继续使用原权威提交，不读 Pending。多个过滤字段在一次路由匹配中使用同一生效配置快照。没有配置引用的静态过滤无需调用此函数；有引用而未装配读取器时返回实际错误并跳过，不放行。
- T-08.18/T-04.18 尚未合并时，本切片以基础配置完成验收；它们与本切片的后合并者必须按执行顺序接入 Resolve 和有效 Overrides，使选择器和唤醒词均按会话生效、覆盖更新从后续匹配生效；运行值与 Handler 的 profile_id 使用同一配置文件解析规则。先通过注入值验证此接口，不把该结果记成真实配置文件验收。
- 与 S-01.2 的后合并者把 message_filter 及配置路径加入白名单和确定性摘要；配置/覆盖值变化不改变快照或摘要，唤醒后重新 REGISTER 照常校验。S-02.3 不建立快照存储或休眠状态机。

测试（按编码规范「测试提交规则」）：
- 提交：internal/app/message_filter_test.go 一个集成测试（unix），复用 pipeline_test.go 的真实 stdio 子进程测试宿主与现有资源登记：SDK 注册声明，Core 完成 REGISTER 校验后注入未命中和命中的消息，断言只有命中才产生 Handler REQUEST，命中后 SDK Rule 仍可拒绝；注入两个会话的不同字段值并切换覆盖快照，匹配随生效值变化，未声明过滤的路由行为保持原样。不新增 testdata 插件。
- 提交：contracts/wirespec_test.go 新字段编号检查。多列理由：协议编号固定检查与跨进程交付分别覆盖独立契约。
- 本地自测（*_local_test.go，结果写入 PR）：V-77 的 to_me/PRIVATE/GROUP/CHANNEL OR、空过滤器在 SDK 与 REGISTER 被拒绝、triggers 为空与多选、空唤醒词、分段文本/大小写、edited/deleted 缺失字段；配置路径/类型/secret 非法、读取失败、Pending 隔离、候选失败和原代次；builder 的重复/Handle 顺序；若已有 profile，核对真实会话指定/实例缺省/有效覆盖（包括 object 祖先覆盖）的交付；若已有快照，核对新字段往返和摘要不一致拒绝。sdk/ 与 supervisor/ 只放本地自测。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./contracts/... ./internal/core/router/... ./internal/core/host/... ./internal/core/supervisor/... ./internal/app/... ./sdk/...；go build ./...。PR 写明 V-77 本切片分项的命令、平台、结果和未运行项，以及快照/配置文件后合并者的接入情况。

不做：休眠生命周期、EnsureActive、IDLE、管理 API、WebUI、官方 Agent 的触发方式改造（goal-159/164）、新的配置存储或 profile 实现；不加依赖，不改其他 proto、schema.sql、tune 或 ARCHITECTURE/README/spec，文档同步写进 PR。
```

## 第 19 波

### goal-161 · T-04.18 配置文件的派发与 SDK

状态：已完成（PR #242），合并提交 `b816127`。以下保留派发时语句。派发时：前置 T-04.16（PR #199，`046e752`）、T-08.18（PR #214，`ab0cd56`）均已合并；消息过滤 S-02.3（PR #227，`86e6768`）也已合并。已按 main `947daa5` 核对：`internal/core/profile.New(ctx context.Context, opt profile.Options) (*profile.Service, error)` 建立一次操作的快照，`Resolve(instance, session string) profile.Profile`、`Overrides(pluginID string) map[string]map[string]json.RawMessage` 已可复用；`sdk/sdk.go` 的 `(*Context).Config() config.Values` 当前直接返回插件基础配置，`internal/app/message_filter.go` 只读 Active 基础配置，InvocationContext 仍止于字段 6、Bootstrap 止于字段 13。goal-152（T-08.16）已派发、正在进行；双方共享 `contracts/configschema/`（尤其 profile.go、prepare.go 的准备/校验入口）、`sdk/config/`、`sdk/session.go`、`sdk/plugin.go`、`contracts/proto/alonabot/v1/control.proto` 及生成产物，并可能相交于 Core 配置快照的读取调用点。后合并者同步最新 main，保留对方改动：goal-152 的去版本与宽松校正、goal-161 的 profile 标记与覆盖合并同时保留，字段号按已合并源文件顺延，重新生成并核对两条主路径。

```text
/goal 使用 $alonabot-execute 推进 goal-161（切片 T-04.18：配置文件的派发与 SDK）。goal 目录 goal-161，分支 goal-161，worktree .worktrees/goal-161，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：派发时按会话解析配置文件，InvocationContext 带 `profile_id`；Bootstrap 与 `sdk.profile.set_overrides` 下发插件的覆盖值，Handler 中的 `c.Config()` 合并；配置文件关闭的插件不接收该会话的派发；提供 `core.profile.list`、`current`、`assign` 与 `set`。

依据：PLAN T-04 步骤 18；CONTRACTS §5.6；WIRE-FORMAT §3.2、§3.5、§4；DECISIONS D-168；TEST-MATRIX V-141。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：invocation.proto、control.proto、新文件 profile.proto；相应 contracts/pb/alonabotv1/ 生成产物与 contracts/wirespec_test.go；internal/wire/services.go
- contracts/configschema/profile.go 及覆盖值应用的包内辅助文件
- sdk/：sdk.go、plugin.go、dispatch.go、session.go、服务/Hook/延续调用的 Context 构造点、config/ 与 internal/dbhook/ 的只读配置视图绑定；覆盖集合接收的包内新文件
- internal/core/profile/ 的运行时读取与服务辅助文件；internal/core/router/ 的 Options、Invocation、命令与事件派发、延续调用及根调用结束通知；internal/core/broker/、hook/、scheduler/ 中 InvocationContext 的构造与继承点
- internal/core/manage/profile_api.go、service.go 及配置文件变更后通知的相邻调用点；internal/core/config/ 仅为 default 上 Pending 的既有写入与互斥接入
- internal/app/：新文件 profile.go，message_filter.go、assembly.go、journal.go、alonabot.go 中配置文件快照、Bootstrap、固定服务、更新通知、根调用结束后 reload 的装配
- cmd/alonabot/manage_e2e_test.go；上述包的 *_local_test.go

允许的依赖：标准库及仓库已批准依赖；复用 contracts/configschema、internal/core/profile、config、policy、router 和 state。跨模块沿用 ARCHITECTURE 的 internal/app -> internal/core/...、internal/core 内部、sdk -> contracts/internal/wire 方向，SDK 保留现有分包。上述协议修改明确授权；只改生成源后运行 go generate ./...，不增加依赖。

接口（现有签名保持可用）：
- profile.New(ctx context.Context, opt profile.Options) (*profile.Service, error)，Options{Store *state.Store; Settings *settings.File; Schema func(context.Context, string) (*configschema.Schema, error)}
- (*profile.Service).Resolve(instance, session string) profile.Profile；List(ctx context.Context) ([]profile.Summary, error)
- (*profile.Service).Overrides(pluginID string) map[string]map[string]json.RawMessage；Plugins(id string) ([]profile.Plugin, error)
- (*profile.Service).Assign(ctx context.Context, instance, session string, profileID *string) error；PatchPlugin(ctx context.Context, id, pluginID string, patch profile.PluginPatch) error
- (*configschema.Schema).ProfilePaths() []string；ValidateProfileValue(path string, raw json.RawMessage) error
- (*config.Service).Pending(ctx context.Context, pluginID string) (config.Pending, bool, error)；Patch(ctx context.Context, pluginID, baseGeneration string, patch []byte) (config.Pending, error)
- (*sdk.Context).Config() config.Values；(*sdk.Plugin).Config() config.Values；(*sdk.SubPlugin).Config() config.Values
- router.Options.MessageConfig func(ctx context.Context, pluginID string, event *pb.Event) (json.RawMessage, error)

接口（本切片新增，以下签名和载荷固定）：
- InvocationContext 追加 string profile_id = 7（default 编码为空串）；Bootstrap 追加 string profile_overrides_json，当前 main 的下一号为 14，与并行切片合并时按最新 main 顺延
- 新文件 profile.proto：message ProfileSummary { string id = 1; string name = 2; uint64 sessions = 3; }；ProfileListRequest {}；ProfileListResponse { repeated ProfileSummary profiles = 1; }；ProfileCurrentRequest {}；ProfileCurrentResponse { ProfileSummary profile = 1; }；ProfileAssignRequest { string profile_id = 1; }；ProfileAssignResponse {}；ProfileSetRequest { string path = 1; string value_json = 2; }；ProfileSetResponse { ProfileSummary profile = 1; bool reloading = 2; }；ProfileSetOverridesRequest { string profile_overrides_json = 1; }；ProfileSetOverridesResponse {}
- 固定服务名追加 wire.ServiceProfileList="core.profile.list"、ServiceProfileCurrent="core.profile.current"、ServiceProfileAssign="core.profile.assign"、ServiceProfileSet="core.profile.set"、ServiceProfileSetOverrides="sdk.profile.set_overrides"。请求/响应使用上列同名 protobuf 消息，schema/version 沿用固定服务的 validate.SchemaID 与版本 1
- func (s *Schema) ApplyProfile(raw json.RawMessage, overrides map[string]json.RawMessage) (json.RawMessage, error)（contracts/configschema）：在基础对象中按完整字段路径替换有效覆盖，复用 ValidateProfileValue；返回新对象，保留未覆盖内容
- func (v Values) WithOverrides(overrides map[string]json.RawMessage) (Values, error)（sdk/config）：以同一 schema 和基础配置建立独立只读视图，复用 ApplyProfile，保留既有子插件配置路径
- router.Options 追加 Profile func(ctx context.Context, pluginID string, target *pb.Target) (profileID string, enabled bool, err error)；内部 default 使用 "default"，仅编码到 InvocationContext 时转为空串
- router.Invocation 追加 ProfileID string、Root *Invocation、OnFinished func()；根调用 Root 指向自身，派生调用保留根引用，OnFinished 在根调用结束时执行一次，用于延后 reload，执行期间不持 Router/Broker 锁
- manage.Options 追加 NotifyProfileOverrides func(ctx context.Context, pluginID string) error；装配方从新快照取完整有效集合，只通知当前在线 Session，失败交给原 Session 失败与清理路径

行为：
- 按 CONTRACTS §5.6 从可信默认 Target 解析实例名与 MESSAGE-MODEL §1 的会话字符串（含话题及转义），单独指定优先于实例缺省，再取 default；无默认 Target 的调用取 default。复用 profile.New/Resolve，一次操作使用同一快照，下一次操作观察管理变更；保留现有管理锁顺序。所有 Core→插件 InvocationContext 构造点及父子调用的可信附件同步 profile_id，保持 Actor、Target、Trace 与取消传播。
- Bootstrap 下发该插件当前有效的全部覆盖值，JSON 形状为配置文件 ID -> {字段路径: 值}；失效覆盖排除，复合 secret 规则与跨配置文件 4 MiB 编码预算复用 profile.Service。配置文件创建/复制/删除、字段覆盖写入/删除/prune，以及 Active schema 切换后，向受影响的在线 Session 下发完整集合；sdk.profile.set_overrides 整体替换，确认或发送失败按 Session 失败处理并返回实际错误。休眠/离线插件在下次 Bootstrap 取最新集合；与 goal-170 后合并者保留双方装配及清理。
- SDK 接收集合后原子发布快照；每次 Rule/Handler/Service/Hook 及延续调用建立 Context 时冻结本次 profile_id 与覆盖集合。c.Config() 合并对应覆盖，object/list 等字段按路径整体替换，覆盖祖先 object 时其子字段取新对象；本次调用的视图保持不变。p.Config()、s.Config()、OnStart 与后台基础配置视图保持原义。覆盖更新作用于接收更新后开始的调用，更新无需重启插件。
- 会话关闭插件时，命令按原禁用规则处理，文本匹配、事件路由与 Raw 跳过；在取得唤醒名额前应用同一开关。插件间 Service、定时任务与后台发送继续按原规则运行。工具策略的配置文件开关由 goal-163 接入。
- S-02.3 已合并，必须更新 internal/app/message_filter.go 向 router.Options.MessageConfig 提供同一会话的 Resolve -> Active 基础配置 -> 有效 Overrides 只读快照，排除 Pending，并支持祖先 object 覆盖。过滤与 Handler 选择会话的规则一致；管理更新从后续匹配生效，读取失败保留诊断并跳过该路由，快照不保存覆盖值。与 goal-170 后合并者保证过滤先于活动引用及 EnsureActive。
- core.profile.list 返回全部配置文件 ID、名称与会话数；current/assign/set 必须关联同一 Session 中仍在途的 parent_call_id，从 Core 的 router.Invocation 取可信 Target 与 Actor，无默认 Target 返回 FAILED_PRECONDITION。assign 的空 ID 取消单独指定，set 只接受调用插件自己标记 profile 的路径，JSON null 删除覆盖；两者均要求可信 Core Admin，越权返回 PERMISSION_DENIED。
- set 在非 default 上复用 PatchPlugin 的单值、secret 和预算检查；default 上在同一互斥操作中核对当前 Active 与 Pending，已有其他 Pending 返回 FAILED_PRECONDITION，以当前 Active 为基线写入 Pending，返回 reloading=true。用根调用结束通知触发既有 reload，调用链结束前保留原 Session；异步 reload 的实际错误记入既有状态与日志。字段未标记、无效 JSON 与类型错误保留路径及原错误，语义与管理入口一致。
- 与 goal-152 共享的配置载入沿用唯一实现：goal-152 合并前保留 main 的载入签名，后合并者同时保留 profile 标记、完整子树秘密检查、有效覆盖筛选，以及去版本后的 OnConfigLoad(ctx, raw) 与宽松校正。覆盖应用按 §5.6 单值校验与替换，不重新执行 OnConfigLoad 或给覆盖补默认值；随 goal-152 合并同步受影响调用点，不留旧版重载入口。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 新增一个真实 Core＋插件子进程的配置文件派发集成子测试，经现有管理 API 创建并指定非 default 配置文件，两个会话调用同一命令得到各自 c.Config()；更新覆盖后下一次调用生效且 Session 不变，再关闭其中会话的插件开关并核对禁用结果。复用现有动态构建与回环 Adapter，不新增 testdata 插件。
- 提交：contracts/wirespec_test.go 检查 profile_id、profile_overrides_json 与新消息的字段号及既有编号。多列理由：公开协议追加须独立固定编号，进程集成测试无法证明编号未漂移。
- 本地自测（*_local_test.go，结果写入 PR）：V-141 其余分项，包括同一调用冻结、p.Config/s.Config/OnStart/后台不变、Service/Hook/延续上下文、命令/文本/事件/Raw 开关、四个固定服务的 parent/Admin/Target 校验、default 与已有 Pending、根结束后 reload、Session 重启与通知失败；两个会话的唤醒词及 triggers/祖先 object 覆盖、过滤读取失败、更新后匹配；覆盖集合秘密与大小、Active schema 改变后的失效项、goal-152 校正后再合并。sdk/ 中只写本地自测。

完成条件：go generate ./... 后再次生成无差异；gofmt -l 无输出；go vet ./...；go test -race ./contracts/... ./internal/wire/... ./sdk/... ./internal/core/profile/... ./internal/core/config/... ./internal/core/router/... ./internal/core/broker/... ./internal/core/hook/... ./internal/core/scheduler/... ./internal/core/manage/... ./internal/app/... ./cmd/alonabot/...；go build ./...。PR 按 TEST-MATRIX §3 写明 V-141 与消息过滤衔接的测试位置、命令、平台、结果与未运行项。

不做：命令前缀与回复方式（goal-162）；工具策略（goal-163）；Agent 的命令（goal-164）；配置文件页面；重做 goal-152 的载入流程或另建配置真源；不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-177 · T-09.27 组件市场 ID 改用连字符

状态：已完成（PR #222），合并提交 `5940406`。以下保留派发时语句。派发时：前置 T-09.10（PR #177）、T-09.14（PR #192）、T-12.8（PR #175）、T-13.5（PR #200）均已合并，接口已按当前 main 核对。本切片与同时进行的其他 goal 都可能写到旧 ID；后合并者同步 main 后用 `rg -n 'alonabot_(plugin|adapter|agent|webui)_'` 把新出现的旧 ID 一并改掉，按执行顺序「组件 ID 的改名」处理。goal-141 的目录名缺省尚未实现，按下文衔接，不作为本切片前置。

```text
/goal 使用 $alonabot-execute 推进 goal-177（切片 T-09.27：组件市场 ID 改用连字符）。goal 目录 goal-177，分支 goal-177，worktree .worktrees/goal-177，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按 D-172 把市场 ID 的四角色前缀与后缀分词统一改为连字符，同步当前 main 上的官方组件、示例、校验调用者、工具原始名、日志、夹具与已有测试。本地 ID 的规则和 64 字符上限保持不变。首次发布前直接改名，状态库与开发数据目录需要重建。

依据：PLAN core T-09 步骤 11 及其失败停止条件；CONTRACTS §2、§8、§9.1、§14、§15.2、§15.4；DESIGN §B 的身份与 §D 的资源登记规则；ADMIN-API §8；WIRE-FORMAT §1/§3、MESSAGE-MODEL §1；DECISIONS D-172；TEST-MATRIX V-01、V-62、V-64 与 V-121 的相关分项；docs/coding-standards.md「测试提交规则」；docs/execution-order.md「组件 ID 的改名」。

范围（只改本次 ID 值、前缀判断和由此受影响的校验、错误文案、断言与资源；目录列项包括其既有测试和忽略的 *_local_test.go）：
- contracts/validate/plugin_id.go、adapter.go；contracts/manifest/manifest.go、project.go、deps.go；contracts/wirespec_test.go 仅在既有编号检查确有必要时调整，不改协议编号。
- contracts/testdata/manifest/market.yaml、release/release.json、frames/hello.bin、frames/request_with_parent.bin；internal/wire/frame_test.go（已有 golden 帧集成测试）。
- internal/core/logging/（console.go 的 sourceTag、query.go 的历史解析及既有测试）、manage/admins.go、manage/settings_api.go、options/chats.go、policy/denies.go、settings/settings.go；这些包的既有测试。
- sdk/message/session.go（validAdapterID 与 ParseSID）；SDK 测试只在 *_local_test.go 中取证。
- agent/memory.go、tools.go 与 agent/ 下既有测试；agent/plugin.yaml、adapters/telegram/plugin.yaml、web/admin/plugin.yaml；adapters/onebot11/ 中若同步 main 后出现工程 manifest 或包 ID 字符串，一并改名。
- examples/devbot/plugin/plugin.yaml、adapter/plugin.yaml、core_config.example.json、README.md；examples/plugins/configdb/plugin.yaml、examples/plugins/media/README.md，以及 examples/ 中本次改名影响的既有测试与配置。
- alonabot_test.go；internal/app/admin_upload_test.go、agent_service_test.go；internal/core/packages/preflight_test.go；cmd/alonabot/{plugin_cli_e2e,t04_e2e,t12_e2e,t13_e2e,telegram_e2e,manage_e2e}_test.go。
- cmd/alonabot/init.go；internal/pluginkit/init.go、build.go、pack.go 与既有测试；internal/generate/（project.go、source.go 与 Scan 的 manifest 读取及资源嵌入）；internal/generated/、internal/wire/env.go、internal/core/process/process.go、internal/core/state/、sdk/sqlitedb/、sdk/gormdb/ 只读核对 ID 的使用方式。
- docs/development.md、docs/plugin-guide.md 及上述 examples/ README 中与本次实现状态和数据重建有关的说明；spec/TEST-MATRIX.md 仅回填本切片实际证据。README.md、ARCHITECTURE.md 只读核对，需更新的实现状态列入 PR「文档同步」；它们与 docs/goals/、docs/execution-order.md、切片表均由审查者维护。
- 同步 main 后新出现的旧组件 ID，按共享文件规则同步其字符串、前缀判断、manifest、配置与夹具；若 plugins/help/、plugins/welcome/、plugins/rss/、packages/、scripts/ 或 deploy/ 已由其他 goal 引入，沿用已确定的 ID 把其中的下划线换为连字符，help 固定为 alonabot-plugin-help-official。该例外只授权改名及必要调用方，不扩展组件功能。

接口（保留现有公开名称、签名与协议字段）：
- contracts/validate：MaxPluginIDLen = 64 不变；MarketPrefix = "alonabot-plugin-"；marketID = regexp.MustCompile(`^alonabot-plugin-[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)。PluginID(id string) (warnings []string, err error)、IsAdapterEventType(eventType string) bool 的签名不变。
- contracts/manifest：Parse(data []byte) (*Result, error)、ParseProject(data []byte) (*Result, error) 不变；私有 componentID(id, role string) ([]string, error) 使用 "alonabot-" + marketRole + "-"，marketComponentSuffix 为 `^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`。四角色前缀分别为 alonabot-plugin-、alonabot-adapter-、alonabot-agent-、alonabot-webui-。
- sdk/message：ParseSID(sid string) (adapterID, chatID, threadID string, err error) 不变；私有 validAdapterID(id string) bool 与 options.chatKey(value string) (state.ChatKey, bool) 继续按本地平台 ID 规则检查。
- internal/pluginkit：Init(ctx context.Context, dir, id, role, adapterID, module string) (needsSDK bool, err error) 不变；internal/generate：Scan(ctx context.Context, opt Options) (*Result, error) 不变。生成资源按导入路径登记，manifest 的 ID 仍是字符串。
- agent：私有常量 memoryTool = "alonabot-agent-official/memory"；工具注册、目录、调用与按原名记录的策略使用同一新值。对模型的名字继续经过现有 mapTools/normalizedToolName，不新增映射协议或别名。
- 日志继续由 sourceTag(fields map[string]string) string 生成标签，logging.Query(dir string, date time.Time, source string, level slog.Level, traceID string) ([]Record, error) 的签名不变；仅插件市场前缀 alonabot-plugin- 从 [Plug:<id>] 中去掉，Adapter 标签仍使用平台 adapter_id。

行为：
1. 四角色市场前缀与 role 一致，后缀以小写字母开头，单个连字符分词，后续段由小写字母或数字组成；完整 ID 长度至多 64。带新市场前缀却不符合市场正则、或角色前缀错配时拒绝；无市场前缀的安全本地 ID 继续通过并警告。旧下划线市场拼写只按原有本地规则处理并警告，既不识别为市场 ID，也不自动映射为新身份。
2. 本地 ID 仍用 `^[a-z][a-z0-9_-]{0,63}$`，保留小写与 Windows 设备名检查。adapter_id、会话、权限/扩展事件、管理员、禁止表、设置与聊天选项中的平台 ID 按本地规则检查，原值原样存储与传递；同步各处 MarketPrefix 判断，避免把组件市场专用校验带入平台 ID。session.go 与 chats.go 现有检验副本中的替换针对旧分隔符，须同步处理；不得改写实际 adapter_id 或收紧本地语法。
3. 官方与示例直接改名：alonabot-adapter-telegram、alonabot-agent-official、alonabot-webui-admin、alonabot-plugin-configdb、alonabot-plugin-devbot、alonabot-adapter-devbot；help 为 alonabot-plugin-help-official。组件 ID 用在 HTTP 路径、Service/工具原始名、依赖、归档名、release、生成资源、core_config.json 的 package/log.sources、dev:<ID> 代次和数据目录时都保留新拼写。telegram、onebot11、devloop 等平台 adapter_id 与实例名保持原值。
4. Go 包名与标识符保持 plugin、component 等现有合法名字，不从市场 ID 直接生成符号；生成器按现有 %q 转义嵌入原始 manifest。ALONABOT_PLUGIN_ID 等固定环境变量的名字保持不变，ID 只放在值中；SQLite/GORM 中 ID 是绑定参数或字段值，表/列名仍用既有固定标识符。只读核对未发现把市场 ID 直接用作 Go 标识符、环境变量名或 SQL 标识符的路径。模型工具名已有 [A-Za-z0-9_] 归一化与 FNV 后缀，连字符由该映射处理，原始路由名保持完整新 ID；核对映射回调用目标与碰撞处理。
5. 当前 init 的 --id 仍为必填，显式给新市场 ID 时经 ParseProject 校验并生成可构建工程。goal-141 合并后，--id 缺省取目录名并走同一四角色/本地规则；目录 alonabot-plugin-hello 合法，本地目录 my_plugin 仍合法并警告，非法目录名报错并提示显式 --id，不自动替换目录名字符。后合并者保留两边实现；本切片不提前实现目录名缺省。
6. 新日志标签示例为 [Plug:configdb]，JSON 与实时 source 保留 alonabot-plugin-configdb，历史查询 source 为 configdb；Adapter 标签继续为 [Adpt:telegram/<instance_id>]。不为旧日志新增身份还原或前缀兼容层。
7. golden 帧使用现有 pb.Envelope 解码，修改其 ID 字符串后重新编码，并核对 4-byte big-endian 长度；保留原字段号、枚举值与其他载荷，不做二进制文本替换，不新增永久生成器。manifest/release 中的 ID 与归档 URL 同步。
8. 不做状态或数据的自动迁移。文档说明停止 Core 与开发组件后由开发者重建状态库与开发数据目录，并按新 ID 重配 package、log.sources、权限及工具策略；测试只使用临时数据目录，不改维护者现有数据。

测试（按编码规范「测试提交规则」与 docs/goals/README.md，D-86）：
- 提交：只改写范围内已有集成测试的 ID、路径、错误文本与资源夹具，保持原场景与断言含义；不新增提交测试。沿用 internal/wire/frame_test.go 的 TCP/stdio golden 帧验证和 contracts/wirespec_test.go 的编号检查；本切片没有协议字段变更，后者只在确有必要时改写既有检查。SDK 的验证全部在忽略的 *_local_test.go 中运行，不新增 testdata 测试插件。
- 本地自测（*_local_test.go，结果写入 PR）：contracts/validate 与 contracts/manifest 检查四角色正例、前缀与角色错配、后缀下划线、连续/首尾连字符、空后缀、首段数字、64/65 字符、小写、Windows 保留名；旧下划线拼写按本地规则警告且无别名；普通本地 ID 的下划线和连字符继续可用。adapter_id 的各入口同时覆盖形似市场前缀的合法本地字符串，确保 session/options 与注册、策略、管理设置一致。
- 本地自测：通过真实 CLI init --id alonabot-plugin-hello 创建工程，check/build/pack 后读取 manifest/release 与归档路径，核对新 ID 原样保留，编译生成资源；若 goal-141 已合并，追加省略 --id 的合法/非法目录名场景。通过现有 Agent 主路径与本地检查核对 memory 工具注册、模型名称映射和原名调用；日志核对控制台、JSON、实时与历史查询的上述 source 值。
- 验证量与改名匹配：已有测试只做必要改写；边界与回归只写本地测试，不新增独立测试框架、包或依赖。

完成条件：上述本地检查与已有主路径集成测试通过；gofmt -l 对改动 Go 文件无输出；go generate ./... 及生成一致性检查无协议生成差异；go vet ./...；go test -race ./...；go build ./...；带 configdb_gorm tag 的 configdb 构建通过。运行 rg -n 'alonabot_(plugin|adapter|agent|webui)_' 并解释全部命中：允许 DECISIONS 历史条目、D-172 的变更说明与 spec/adr/history/，其他被跟踪的当前文档、实现、配置、测试与夹具均清理；另以 rg --text 检查 contracts/testdata/frames，并搜索动态拼接的旧前缀。PR 按 TEST-MATRIX §3 写明命令、平台、结果、未运行项与提交测试新增行数，数据重建要求写在文档同步段。未通过任何现有主路径、四角色规则或本地 ID 兼容检查时按工作包失败条件阻塞，不改验收语义。

不做：旧市场 ID 别名、兼容层、自动迁移或删除维护者数据；修改本地 ID 规则；协议字段与枚举的改动；目录名缺省（T-13.6）；仓库、Go module、可执行入口与平台 adapter_id 改名；新功能、依赖、测试框架或发布动作。
```

### goal-45 · T-09.11 正式流程故障注入与 T-09 验收

状态：进行中。全部前置均已合并：T-09.3（PR #194，`e9ff45e`）、T-09.7（PR #187，`864819e`）、T-09.9（PR #259，`661cf9b1`）、T-09.10（PR #177）、T-09.12（PR #167）、T-09.15（PR #243，`d9e9089`）；external 正式许可 PR #265（`a97d4362`）已合并。按 main `661cf9b1` 复核 CLI、manage、Journal、配置面板和测试辅助，执行者已据此启动正式流程与故障验收。

```text
/goal 使用 $alonabot-execute 推进 goal-45（切片 T-09.11：正式流程故障注入与 T-09 验收）。goal 目录 goal-45，分支 goal-45，worktree .worktrees/goal-45，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：通过真实 CLI 与管理 API，在临时 data_dir 中跑完正式流程：首次安装 → enable → restart → reload → update → pin → disable → 同来源重装 → 显式 rollback；以包损坏、签名或公钥变化、external 插件的维护错误验证拒绝路径；在 PREPARED、提交前后与 OnReady 失败点终止 Core，连同残存旧进程、空间不足与同插件并发操作，验证没有双写、没有提前开放、没有静默的数据回滚。回填 T-08 留给集成的 V-23～V-25、V-39 证据，在 TEST-MATRIX §3.2 记录 T-09 证据，形成 T-10 可复用的恢复场景。

依据：spec/plan/core.md 的 T-09 步骤 8 与停止条件、PLAN §3.1；CONTRACTS（spec/CONTRACTS-v0.1.md）§5、§6.3、§7、§9、§12.2、§12.3、§14；ADMIN-API §1、§3、§5、§10、§12、§14～§18、§20；TEST-MATRIX V-23～V-25、V-27～V-31、V-39、V-40、V-52、V-54～V-56、V-62、V-63、V-82、V-83 及 §3 的分项证据；DECISIONS D-86。

范围（只新增或修改这些）：
- cmd/alonabot/t09_e2e_test.go（新增，//go:build unix）：一条正式管理主路径 TestT09ManagementFlow；e2e_test.go、manage_e2e_test.go、recover_e2e_test.go 中既有辅助的必要复用调整，保留原测试断言。
- scripts/scenarios/：供 T-10 重用的正式流程、崩溃恢复与残存写入者场景及入口说明；管理动作只调用传入的 alonabot 二进制。场景接收临时 data_dir、测试包与所需的临时 overlay 二进制，说明实际构建和执行命令、输入产物与检查点。平台相关的进程/锁场景标注支持范围，五平台实际运行由 T-10 取证。
- spec/TEST-MATRIX.md §3.2：追加 T-09 的证据行，并更新 T-08 行中留待集成的项（本切片获授权修改这两处）。
- cmd/alonabot 的 *_local_test.go；需要补充的插件源码、配置、签名密钥和 overlay 均在测试临时目录生成（D-86）。复用 examples/plugins/lifecycle、configdb、showcase 和现有 Adapter 测试源码。
允许的新跨模块边：无；沿用现有测试对 manage、state、buttons 等包的依赖与已批准依赖。产品缺陷由审查者确认修复归属并补足文件范围后继续。

接口与入口（现有，沿用原签名）：
- CLI 为顶层 install [archive|component-id]、enable/disable/restart/reload/update/uninstall <id>、pin <id> [version]、rollback <id> <generation>、list；install/update 使用 --archive、--release、--signature、--public-key 或 --source owner/repo，install 另有 --replace。pin 是精确 SemVer，省略 version 解除；uninstall 缺省保留数据，--purge-data/--purge 沿原确认语义。共享 --data-dir、--admin-url、--token-file 或 ALONABOT_ADMIN_TOKEN；需要确认时 --yes。维护 reload/update/rollback 另带 --manual-migration-approved，表示先 disable 并完成外部备份，普通确认仍用 --yes。CLI reload 先 POST /api/v1/plugins/{id}/config/import，再 POST /api/v1/packages/reload；API/UI reload 直接激活已有有效 Pending，缺 Pending 时导入编辑快照。
- manage.New(opt manage.Options) *manage.Service；(*manage.Service).Handler(auth *adminauth.Auth) http.Handler；List(ctx context.Context) ([]manage.Package, error)；Install(ctx context.Context, req manage.InstallRequest) (manage.Package, error)；Enable/Restart(ctx context.Context, id string) error；Reload(ctx context.Context, id string, manualMigrationApproved bool) error；Disable(ctx context.Context, id string, confirm bool) error；Uninstall(ctx context.Context, id string, purge manage.PurgeScope, confirm bool) error；Pin(ctx context.Context, id, constraint string) error；Update(ctx context.Context, id string, req manage.InstallRequest, manualMigrationApproved bool) (manage.UpdateResult, error)；Rollback(ctx context.Context, id, generation string, confirm, manualMigrationApproved bool) error。
- InstallRequest{Archive,ReleaseJSON,Signature,PublicKey,Source string; Replace,Confirm bool}；JSON install 保留这些大写字段，update 用 {id,confirm,install}，维护 reload/update/rollback 在请求顶层传 manual_migration_approved，其他动作经 POST /api/v1/packages/{action} 的 id/confirm/constraint/generation/purge，list 用 GET /api/v1/packages。错误保留 §5 的 code/safe_message；需确认时 409 及全部 affected；Cookie 写入校验 Origin/CSRF，Bearer 复用正式认证。
- Journal：journal.New(opt journal.Options) *journal.Orchestrator；Prepare(ctx context.Context, pluginID string, kind journal.Kind, staged *packages.Candidate, manualMigrationApproved bool) (*journal.Result, error)；PrepareRollback(ctx context.Context, pluginID, generation string, manualMigrationApproved bool) (*journal.Result, error)；Commit(ctx context.Context, r *journal.Result) (*supervisor.Running, error)；Abort(ctx context.Context, r *journal.Result, cause error) error；Recover(ctx context.Context) ([]string, error)。生产装配在 internal/app/assembly.go 的 managedCandidates 与 alonabot.go 的启动恢复，仍经唯一管理入口驱动。
- 状态取证使用 (*state.Store).Active(ctx context.Context, pluginID string) (string, error)、Journals(ctx context.Context, pluginID string) ([]state.Journal, error)、GetJournal(ctx context.Context, opID string) (state.Journal, error)、Retained(ctx context.Context, pluginID string) ([]string, error)、ReferencedGenerations(ctx context.Context, pluginID string) ([]string, error)。Journal 的 OpID/Kind/ExpectedOld/NewGeneration/Phase 与资源引用同查；只在 Core 退出后打开状态库核对，管理动作始终经 CLI/API。
- 测试辅助：startEmptyCore(t *testing.T, dataDir string, extra ...string) *core；(*core).management(t *testing.T, dir, token string, args ...string) string；managementFailure(t *testing.T, c *core, dir string, args ...string) string；findPackage(t *testing.T, c *core, dir, id string) manage.Package；fixtureHTTP(t *testing.T, c *core, dir, method, path string, body ...string) (int, []byte)。TestMain 已构建 Core、lifecycle 与 Adapter/command/job 夹具；testPackage 和 writeTestPackage 已生成真实签名安装包。
- 恢复辅助：buildRecoveryFixture(t *testing.T, dir string) (coreBinary, pluginBinary string)；recoveryPackage(t *testing.T, dir, plugin, version string) string；startRecoveryCore(t *testing.T, binary, dir, phase string) *core；recoveryCommand(t *testing.T, c *core, dir string, args ...string) *exec.Cmd；recoveryStore/recoveryWait/waitRecoveryRuntime。现有 TestRecoveryEndToEnd 覆盖 started/prepared/committed、残存写入者和 rollback。临时 overlay 的 ALONABOT_TEST_CRASH_PHASE/ALONABOT_TEST_CHECKPOINT 精确停在 Journal started 后停旧前、PREPARED 后提交前、CommitCandidate 成功后 ACTIVATE 前；fail-start/fail-ready 与 hold-runtime/runtime-held 是临时插件已有故障入口。
- 已合并的 goal-44：web/admin/src/pages/PluginConfigPanel.tsx 的 PluginConfigPanel({id})、api/admin.ts 的 getPluginConfig(id)、patchPluginConfig(id, baseGeneration, patch)、revealPluginSecret(id, path, token) 与 changePackage('reload', id)；读取 active_generation/active/pending/schema/i18n/messages/corrections/unknown，PATCH /config/pending 保存后由 API reload 激活。权限、工具策略、会话与页面桥沿 ADMIN-API 的已合并入口；页面实际操作和截图复用 goal-44/goal-94 的有效证据，按缺口补真实浏览器取证。

验收矩阵（对应提交主路径、既有回归或本地子测试，逐项记录命令、日志与持久状态）：
1) 正式流程：首次安装 → enable（首个 Active）→ restart（只换 Session）→ 编辑 config.json 后 CLI reload（新代次与 SQLite 副本）→ update（新包、新代次）→ pin → disable → uninstall 后同来源重装（复用数据）→ enable → rollback 到保留代次（新代次、旧按钮不复活）。核对 Active、Session、程序版本、配置与 DB 内容；pin 限制更新，解除后恢复。保留代次以 generations API/状态库实际结果选取。
2) 拒绝路径：归档损坏、摘要不符、签名错误、错误 OS/arch、来源 ID 冲突、../ 与符号链接逃逸均在停止旧版前拒绝；旧 Active/Session/DB 保持。已绑定签名公钥发生变化时 403 / PERMISSION_DENIED；原未签名绑定首次信任签名公钥须确认（409 affected，确认后执行）。签名失败保留拒绝结果。
3) external：首次安装与普通 restart/resume 放行；已有安装未许可的 external reload/update/rollback 在 Journal 和停机前拒绝，沿正式前置校验返回 412 / FAILED_PRECONDITION，external 候选拒绝的 safe_message 保留 NEEDS_MANUAL_MIGRATION，旧 Session 与 Journal 数保持。管理员先成功 disable 并自行完成外部备份，再以顶层 manual_migration_approved=true 或 CLI --manual-migration-approved 请求本次候选；--yes 单独承担原管理确认，rollback 仍需确认。串行候选内复核旧 Active、旧或目标包 external、停用记录、无运行 Session 与 runtime.lock；运行中、首装、开发或非 external 许可请求返回 412，其他动作携带 true 返回 400。许可仅绑定当次 Bootstrap 的 Session/Generation/DB，首装与后续 resume 为 false，安装记录与恢复不保存许可。成功完成 Prepared、提交与激活后停止候选并保持停用，管理员检查外部数据后显式 enable，以新 Active 的同代次 resume 开放业务。提交前失败保留旧 Active，提交后失败保留新 Active，两者均保持停用；Core 重启保留外部副作用并清理未提交候选，迁移与未知副作用执行次数保持。复用 PR #265 的 TestManualMigrationApproval 与独立恢复证据，补齐本轮候选上仍缺的正式流程分项。
4) 崩溃点：复用既有 overlay 和 TestRecoveryEndToEnd；PREPARED 后提交前杀 Core，清理未提交候选并恢复旧代次的新 Session；首次安装无旧 Active 时保持未激活；COMMITTED 后、ACTIVATED 前杀 Core，只恢复新代次。fail-ready 导致 OnReady 失败后保留新 Active/committed 与 Degraded，终止并重启 Core 仍按新代次恢复，故障移除后可恢复运行。各点记录 Journal、资源引用、业务 DB 和插件请求日志；ACTIVATED 及 Core 绑定完成前，命令、事件、Service、Bus、任务与 HTTP 保持门禁，任意时刻仅一代接单。
5) 残存旧进程：沿 hold-runtime/runtime-held 让旧进程断链后仍持 runtime.lock，重启 Core 显示 Degraded，进程启动计数不增、Active 与数据保留；确认旧树退出并释放锁后再恢复。清理测试自建进程及锁，保留完整失败日志。
6) 空间不足：沿临时 overlay 在 internal/core/candidate/database.go 的 backupDatabase 返回可识别的 ENOSPC（限故障场景），真实 CLI reload/update 验证失败无提交/无激活、旧 Active/DB/配置保留；已停止旧版时由 Abort 用旧代次的新 Session 恢复。保留真实备份错误与清理/恢复错误；故障注入与真实磁盘不足的证据分别标注。
7) 同插件并发操作：两个真实 CLI 分别 reload 与 disable，用检查点同步其重叠窗口，核对逐包串行、Journal 阶段与最终启用状态，无交错提交、双进程或死锁。
8) 依赖与四角色：复用 TestManagementDependencies/TestManagementReplacement/TestManagementUninstallScopes 的有效证据；必需依赖缺失拒启并列全缺项；停用被依赖者须确认，确认后依赖方仍运行且调用返回 UNAVAILABLE；成环拒装；同类唯一、--replace、--source 后续更新、同来源重装、默认保留与显式 purge 符合 V-40/V-52/V-82。V-83 复用 WebUI 静态/进程、回退、版本检查与图标路径/CSP 的真实入口证据。
9) 回填 T-08：configdb 的 fail_start 在候选 OnStart 写库后失败，fail_migration 覆盖版本迁移；两者分别按默认 SQLite 与 -tags configdb_gorm 构建打包运行。V-23 验证编辑 config.json 在 reload 前不生效及 Pending/base_generation；V-24 验证候选失败、Core 崩溃与 WAL 复制不污染旧库；V-25 验证 secret 修改、旧路径转未知键仍保密、invisible/显式空值/未知键保留、掩码拒写及查看需再认证，复用目标页面实际操作补齐；V-39 包含第 3 项完整维护门槛。
10) 日志、i18n、工具与会话：log-level [ID] [LEVEL]、logs 的 --source/--level/--trace/--date/--follow 经正式 API 核对即时级别、重启持久化、历史/最近/SSE；复用 TestLogManagement 的断言。核对 Core usage 和页面语言切换（V-54）、工具 allow/confirm/deny 的下一次调用生效（V-55）、会话列出/清除与 purge（V-56），沿管理 API 与 goal-44/goal-94 的真实浏览器证据取证。TEST-MATRIX 已指定由 T-09.11 汇总的后续分项复用所属切片记录，有缺口逐项列明。
- scripts/scenarios 把 1)、4)、5) 做成传入二进制和测试包即可重复执行的场景，交付 README 中完整命令与所需产物；崩溃钩子、故障插件及检测 marker 保持测试产物，不向正式 Core/SDK 增加开关或公开 API。
- 证据：TEST-MATRIX §3.2 的 T-09 与 T-08 待联验行按分项填写测试位置、命令、平台和结果，V 编号有未完成项就保留缺口；五平台的真实运行留给 T-10，写「未运行」。

测试（遵循 docs/goals/README.md「编写测试段」与 D-86）：
- 提交：cmd/alonabot/t09_e2e_test.go:TestT09ManagementFlow，一条真实 CLI＋认证 API＋已安装插件子进程＋SQLite/文件系统的正式流程；复用上述包构建、状态与 HTTP 辅助。scripts/scenarios 作为 T-10 的可运行交接产物提交。既有 TestRecoveryEndToEnd/TestManagementLifecycle 等仅做必要复用调整并保留断言，已覆盖场景引用原入口。
- 本地自测（*_local_test.go，不提交，PR 逐项写结果）：矩阵中主路径及现有有效证据尚未覆盖的拒绝、external、首次安装崩溃、OnReady 失败、ENOSPC、并发、配置/secret、权限/工具/会话分项。复用已有精确检查点，不靠固定 sleep 猜测崩溃窗口；实际失败修复后定向复验，不把已通过且无改动的崩溃测试无条件重复运行。

完成条件：改动 Go 文件 gofmt -l 无输出；git diff --check；go vet ./...；go test -race -count=1 ./cmd/alonabot -run '^(TestT09ManagementFlow|TestRecoveryEndToEnd|TestManagementLifecycle|TestManagementDependencies|TestManagementReplacement|TestManagementUninstallScopes|TestLogManagement)$'；新增本地分项按名称列出完整命令并运行；scripts/scenarios 三类场景按其 README 实际运行；最终 go test -race ./...、go build ./...。复用前置切片仍有效的专项/浏览器证据；全仓 race 与父审查者串行安排。PR 列提交测试新增行数、实际运行平台、未运行项及「文档同步」。

失败停止条件：鉴权/来源/归档越界或 secret 泄漏、Active 与 COMMITTED 分叉、候选污染旧数据、提交后自动退回旧 DB、旧树未退出就启动写入者、ACTIVATED/绑定前接单、external 未授权准备或获准后失败自动重跑、维护许可入口缺失、并发死锁/双写、静默吞错、清理仍被 Active/Pending/保留代次/恢复 Journal 引用的资源，均阻止交付。维护缺口由审查者补正式接口与范围后继续；其他范围外产品修复、proto、依赖或契约冲突按共同约束记录阻塞。

不做：五平台运行与发布（T-10）；新增产品功能、通用故障注入框架或维护审批系统。必要缺陷修复须由审查者补齐范围，PR 逐条说明并补对应证据。
```

### goal-171 · S-02.2 空闲退出与即时休眠策略

状态：已完成（PR #249），合并提交 `18f8af9`。以下保留派发时语句。派发时：前置 S-01.1（PR #225，合并提交 `2cc32ff7c7848e03621e356195e045736966e2fd`）、S-01.2（PR #226，合并提交 `413a6fffb47a9d4c774a9585b711fb9671eee150`）、S-02.1（PR #241，合并提交 `e0a38d7c1de9061abc51ba1c82423bfdbbd4805c`）均已合并。已按 main `25eea977561ff2beca5f4569a7e2c227acf68cd1` 核对：`contracts/manifest.Manifest.SleepCapable bool` 已贯通工程与安装包；`internal/core/state` 的 `plugin_runtime` 表及 `(*Store).RuntimeStatus(ctx context.Context, pluginID string) (RuntimeStatus, error)`、`SetRuntimeStatus(ctx context.Context, pluginID string, status RuntimeStatus) error` 保存六态、原因及最近唤醒/休眠时间；`ActiveRouteSnapshot(ctx context.Context, pluginID string) (RouteSnapshot, error)` 读取已提交路由。`internal/core/supervisor.(*Waker).EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error)`、`AcquireActivity(pluginID string, session *wire.Session) (func(), error)`、`Ready(running *Running)` 已装配，活动结束时间、Session 引用与冷启动时刻仍为包内记录；`(*Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult`、`WaitExit() (abnormal bool, exitState string, err error)` 已处理正常 IDLE 停止及 SUSPENDED 写入。`contracts/proto/alonabot/v1/control.proto` 已有 `STOP_REASON_IDLE = 6`；`internal/core/settings` 尚无 sleep 字段，`internal/core/idle/` 尚未建立，`internal/tune.Values` 尚无 SLEEP_* 参数。`internal/app.(*Core).ReloadSettings(ctx context.Context) error` 经运行循环重读设置，唤醒启动也经同一运行循环。2026-11-18 审查者裁定 D-179：空闲退出复用 CONTRACTS §4 的正常停止，不新增任何协议消息或字段；进入 QUIESCING 后到达的请求等待下一次唤醒。本条按该裁定替代旧的 SDK 回调排空补充及 S-02.1 的 QUIESCING 撤销行为；CONTRACTS §4.5、PLAN S-02 步骤 4 与 V-77 中残留的旧补充由审查者同步。

```text
/goal 使用 $alonabot-execute 推进 goal-171（切片 S-02.2：空闲退出与即时休眠策略）。goal 目录 goal-171，分支 goal-171，worktree .worktrees/goal-171，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按生效休眠策略执行安全空闲退出，设置修改立即生效，停止中到达的请求经同一生命周期门禁交接。

依据：PLAN §7 S-02 步骤 4 及其失败停止条件、spec/plan/core.md 的 S-02.2；CONTRACTS（spec/CONTRACTS-v0.1.md）§2.5、§4.2、§4.4、§4.5、§15.1、§16；ADMIN-API §24；DESIGN §A；DECISIONS D-53、D-153、D-179（本条的停止与 HTTP 回调规则按 D-179）；TEST-MATRIX V-77 的准入、空闲退出、停止竞争与策略生效分项。

范围（只新增或修改这些）：
- internal/core/idle/：新建 policy.go、idle.go，负责准入、空闲判定及固定冷却期；生命周期执行仍交给 supervisor
- internal/core/settings/：新建 sleep.go，settings.go、correction.go、file.go 中的默认文档、声明、解析校正、深副本、写回及 Reload 比较；复用 document.go 的 Set/Delete
- internal/core/supervisor/：activity.go、wake.go、supervisor.go、runtime.go 中的活动快照、空闲停止门禁与状态记录
- internal/app/：新建 idle.go，assembly.go、alonabot.go、wake.go、journal.go 的装配、设置应用、正常停止与退出清理相邻调用点
- internal/core/state/runtime.go：仅在既有运行状态与原因记录需要调整时修改，复用 plugin_runtime 表
- internal/tune/params.go：追加 SLEEP_MIN_UPTIME、SLEEP_COLD_START_WINDOW、SLEEP_COLD_START_MAX 及 Values 字段
- internal/app/idle_test.go（新集成测试）；上述包的 *_local_test.go

允许的依赖：标准库与现有已批准依赖；idle -> settings（生效策略）、idle -> supervisor/state（活动与运行状态）、idle -> contracts/manifest、contracts/pb/alonabotv1（已验证声明与路由）、idle -> internal/tune（调优值），internal/app -> idle（装配），沿用 ARCHITECTURE 的 internal/core 内部方向；由审查者随 PR 同步新包登记。反向通知通过装配注入函数，保持无 import 环。不加依赖。

接口（main 上的现有签名保持可用）：
- manifest.Parse(data []byte) (*manifest.Result, error)；manifest.Manifest.SleepCapable bool；已安装包的 manifest 读取复用 internal/app 的 installedDirectory 与 installedSpec 身份核对
- (*state.Store).Installs(ctx context.Context) ([]state.Install, error)；GetInstall(ctx context.Context, pluginID string) (state.Install, error)；Enabled(ctx context.Context, pluginID string) (bool, error)；Active(ctx context.Context, pluginID string) (string, error)；LockPlugin(pluginID string) (unlock func())
- (*state.Store).ActiveRouteSnapshot(ctx context.Context, pluginID string) (state.RouteSnapshot, error)；RuntimeStatus(ctx context.Context, pluginID string) (state.RuntimeStatus, error)；SetRuntimeStatus(ctx context.Context, pluginID string, status state.RuntimeStatus) error
- state.RuntimeStatus 的 State、Reason、ChangedAt、LastWakeReason、LastWakeAt、LastSuspendAt，及 RuntimeStopped/RuntimeWaking/RuntimeActive/RuntimeQuiescing/RuntimeSuspended/RuntimeFailed 六态
- supervisor.NewWaker(opt supervisor.WakeOptions) *supervisor.Waker；(*Waker).EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error)；AcquireActivity(pluginID string, session *wire.Session) (func(), error)；Ready(running *supervisor.Running)
- (*supervisor.Running).Stop(t0 time.Time, reason pb.StopReason) host.StopResult；WaitExit() (abnormal bool, exitState string, err error)；host.StopResult 的 Cancelled、Drained、StoppingSentAt、StopDeadline、Killed、Err；host.StopHooks.BeforeStopping func() bool
- (*settings.File).Settings() settings.Settings；Update(edit func(doc *settings.Document) error) error；Reload() (changed []string, err error)；(*settings.Document).Set(path []string, value any) error；Delete(path []string) error
- internal/app 的 (*Core).ReloadSettings(ctx context.Context) error、operation(ctx context.Context, apply func(context.Context) error) error、stopComponent(r *runningComponent, t0 time.Time, timing host.StopTiming, reason pb.StopReason) error；assembly.startWake 经 Core.operation 完成原代次启动与全部绑定

接口（本切片新增，以下签名固定）：
- package settings：type Sleep struct { Mode string; IdleTimeout time.Duration; Plugins map[string]string }；Settings 追加 Sleep Sleep；Defaults 提供 always_on、15m 与空覆盖表，Settings() 深复制 Plugins
- func (d *Document) SetSleep(mode, idleTimeout string) error：校验并只写 sleep.mode/idle_timeout，保留 plugins；func (d *Document) SetSleepPolicy(pluginID string, policy *string) error：非 nil 写入 never 或正时长，nil 删除该插件覆盖。两者供 File.Update 回调使用，复用 Set/Delete；插件安装身份由调用方核对
- package supervisor：type ActivityStatus struct { Session *wire.Session; Ready bool; References int; LastActivityAt time.Time; ColdStarts []time.Time }；func (w *Waker) Activity(pluginID string) ActivityStatus：在现有 w.mu 下读取当前 Session 的引用与实际活动结束时间，并复制冷启动时刻；Ready 仅在当前 Session 完整绑定、存活且接纳工作时为 true
- func (r *Running) StopIdle(t0 time.Time, lastActivityAt time.Time) host.StopResult：在原逐插件锁和 Waker 门禁内复核当前 Session、零引用及实际活动结束时间仍等于本次判定值，再复用 Stop 的同一内部停止实现与 STOP_REASON_IDLE。条件已变化时返回 Cancelled=true，保留当前 Session；普通 Stop 的签名与管理停止入口继续可用
- package idle：type Policy struct { Policy *string; EffectivePolicy string; SleepCapable bool; Eligible bool; Reason string; Warnings []string }；func Resolve(s settings.Sleep, m *manifest.Manifest, independent bool, routes []*pb.RouteDescriptor) Policy：只读准入，返回 ADMIN-API §24 的策略含义
- type Controller struct（字段不导出）；func New(values tune.Values) *Controller；func (c *Controller) CanStop(pluginID string, policy Policy, runtime state.RuntimeStatus, activity supervisor.ActivityStatus, now time.Time) bool：维护逐插件固定冷却期并作空闲决策，供 Core 串行生命周期调用，进程停止由装配方执行
- internal/app 新增 func (c *Core) applySleep(ctx context.Context) error：应用当前 settings 快照，对仍已启用且需常驻的 SUSPENDED 插件经同一 Waker.EnsureActive(ctx, id, 0) 唤醒，成功取得的活动引用立即释放；由 ReloadSettings 及后继管理入口复用，不新增根包导出接口
- tune.Values 追加 SleepMinUptime time.Duration、SleepColdStartWindow time.Duration、SleepColdStartMax int；分别登记 SLEEP_MIN_UPTIME=5m（>0）、SLEEP_COLD_START_WINDOW=1h（>0）、SLEEP_COLD_START_MAX=3（≥1），从本次 Core 生效 Values 注入。空闲时长只存 sleep，不增设调优项

行为：
- sleep.mode/idle_timeout/plugins 为唯一策略来源，模式只接受 always_on/component_idle，时长使用正的 Go duration，覆盖只接受 never 或正时长。auto 通过删除覆盖表达，写入字面量 auto、未知模式或零/负/非法时长返回实际校验错误；文件加载的缺失/非法值及环境变量展开沿用 settings 的声明校正与诊断，原文保留。sleep.plugins 按 dict 处理，支持逐值环境变量；Prune 保留声明内的 sleep 配置。显式写回保留无关字段及未修改的 ${...} 原文，Reload 比较有效 sleep 并返回顶层键 sleep。
- 只对已安装、已启用、声明 sleep_capable 且独立运行的 plugin/agent 判定自动休眠；Core、Adapter、WebUI 与同进程开发组件沿用常驻路径。Resolve 的常驻原因按 always_on、manifest_requires_resident、in_process、policy_never、route_requires_resident 顺序选择，允许休眠时 reason 为空；无覆盖的 Policy 为 nil，常驻 EffectivePolicy 为 never，允许时为生效时长。读取策略与运行状态沿用快照，不唤醒插件。
- HTTP、Hook、持续事件、未声明 Core message_filter 的 message.created/edited/deleted 与 Raw 路由缺省常驻；显式插件时长只放行这些路由准入，manifest 与进程形式限制保持有效，并返回 ADMIN-API §24 的 warnings。普通命令、Core 文本匹配与声明了 message_filter 的消息事件按既有命中规则唤醒；逐条检查全部路由，保留 CONTRACTS §2.5 的配置文件有效覆盖、过滤与权限顺序。
- Core 装配一个 Idle Controller，接入运行循环定时作空闲判定，检查随 Core 生命周期结束；调优项限于 CONTRACTS §16 已列出的三个 SLEEP_* 参数。仅当 RuntimeActive、当前 Session Ready、活动引用归零、空闲已达有效时长、LastWakeAt 已满 SLEEP_MIN_UPTIME 且冷却结束时请求空闲停止。复用 S-02.1 的 Handler、Service、HTTP、媒体操作/租约、已运行 schedule、Hook/Bus、延续、多轮等待及主动出站活动记录；运行中任务阻止休眠，未来 Scheduler 到期唤醒归 goal-172。Core 无法观察私建 goroutine 与连接，安全恢复前提由 manifest 作者声明承担。
- 空闲从当前 Session 最近一次真实活动归零起算，调整阈值保留该时间；初次无业务时沿用 Waker 登记该 Session 的时间。冷启动统计沿用 Waker 的实际启动时刻，在 SLEEP_COLD_START_WINDOW 内次数超过 SLEEP_COLD_START_MAX 时，从越过阈值的启动时刻固定常驻一个窗口；期间新启动保留原冷却截止，结束后恢复判定，已处理的启动记录不会反复触发同一冷却期。
- 空闲停止与升级、配置应用、回退、停用、卸载、Core 停机共用原生命周期串行化及逐插件管理锁；判定后如管理意图、代次、策略、Session 或实际活动时间已变化，撤回本次决定并重新判定。app 的自动空闲入口调用 StopIdle，并复用 stopComponent 的 Session、媒体及调度资源收尾；StopIdle 在同一活动门禁复核引用及时间后进入正常停止。进入 QUIESCING 前的新活动可使本次决定取消，保留 ACTIVE、原绑定与 LastWakeAt/LastWakeReason。
- D-179：进入 QUIESCING 即冻结新派发，停止沿用 CONTRACTS §4.2 的 STOP_TOTAL/STOP_DRAIN/STOP_ONSTOP（缺省 10s/6s/4s）。已接纳请求及合法子调用按原期限最多排空 6s，然后取消剩余请求、发送唯一 STOPPING；SDK 按原路径停止接单、取消 lifetime，用剩余停止 context 执行 OnStop 并关闭 DB/连接。停止截止为 min(t0+STOP_TOTAL, STOPPING 发送时刻+STOP_ONSTOP)，整棵进程树按 t0+STOP_TOTAL 确认退出或终止。复用 STOP_REASON_IDLE=6，不新增协议消息、字段、排空握手或 HTTP 回调完成回执，不增加 SDK 回调计数查询。
- 调整 Waker.EnsureActive、AcquireActivity 及停止门禁的现有 idleStop/stopCancelled/stopCommitted 分支：QUIESCING 开始后到达的新请求进入原有界等待，不撤销已开始的排空，也不交给旧 Session。保留原绝对 deadline、取消及 WAKE_QUEUE_* 和 Core 总预算；确认旧树退出并写 SUSPENDED 后才合并为一次新唤醒，完成 Ready 后复核准入并交付一次。已接纳工作的完成与引用释放仍归原 Session，旧 Session 的迟到清理保留新实例状态。
- HTTP 活动引用继续在 Core 反向代理返回时释放；客户端已取消而 SDK Handler 仍执行、或已写出 EVENT 后 SDK Bus 回调仍执行时，沿用正常停止截止处理。按 D-179 执行既有 SDK 停止路径，取代旧的“等待 SDK 报告全部回调结束或放弃退出”要求。
- 正常空闲停止确认旧树退出后写 SUSPENDED、STOP_REASON_IDLE 原因与 LastSuspendAt，enabled、ActiveGeneration、已提交路由、配置、任务及业务 DB 保持；空闲停止不计异常退出或触发有限重启，不新建候选、不重跑迁移。失败保留完整错误并记录 FAILED/原因，旧树未确认清空时保留运行槽；退出观察者按原实例所有权收尾。撤销与成功由 host.StopResult 区分，撤销后的调用方保留绑定与资源。
- File.Update 成功后应用生效策略，ReloadSettings 同样接入；时间覆盖与删除覆盖在下一次空闲判定使用，设置变更保持已运行 Session。改为 never 立即唤醒仍已启用的 SUSPENDED 插件，切回 always_on 按原代次唤醒全部已启用的休眠插件，停用插件保持停用；变更时已在 QUIESCING 的插件在本次正常停止结束后按新策略唤醒。恢复仍复用 S-02.1 的摘要校验、Ready 与原配置/数据库路径。
- 设置发布、空闲决定与管理动作按现有 Core 运行循环交接；EnsureActive 的 Start 回调也要进入 Core.operation，应用策略时在运行循环及设置锁、插件锁之外等待唤醒，避免持有运行循环等待自身启动。相关重读调用方同步接入。保存失败返回实际错误；保存成功后逐个应用应唤醒的插件并保留各失败，返回完整错误，已保存设置保留，对应失败运行状态记录实际原因，供后继管理 API 按 ADMIN-API §5 返回并重新 GET 核对。

测试（按 docs/goals/README.md「编写测试段」与编码规范「测试提交规则」）：
- 提交：internal/app/idle_test.go（//go:build unix）一个真实 Core＋声明 sleep_capable 的已提交插件子进程＋SQLite 的集成测试，复用 wake_test.go 的动态构建工程、真实 Start 与回环 Adapter，不新增 testdata 插件。注入较短的有效 SLEEP_* 与 sleep 时长，运行真实自动空闲判定，核对 ACTIVE→QUIESCING→SUSPENDED、OnStop/资源关闭、旧进程树退出及 IDLE 原因；首条命令经 Waker 唤醒后只执行一次、Session 更新且原 generation/config/DB 数据保留；再次自动休眠后经内部设置写回与应用路径改为 never，核对随即唤醒且后续保持 ACTIVE。通过同步点观察状态，不以手动 Stop 或直接写 SUSPENDED 替代自动空闲主路径。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：plugin/agent 的 manifest、进程形式及全部路由准入组合、常驻原因顺序与 warnings；设置缺省/校正/环境变量、非法显式写入、深副本、保留覆盖与原文、auto 删除、Prune 与 Reload；活动阈值即时变更、最短运行、固定冷却截止及结束恢复；各类引用、实际结束时间和旧 Session 释放；管理操作互斥、空闲判定后活动进入并结束的竞争、QUIESCING 前取消决定与进入后的新请求等待、STOPPING 后唯一恢复、取消/期限/队列预算；never/always_on 的全部已启用插件唤醒、停用保持、QUIESCING 中策略变更、保存后部分唤醒失败；HTTP 客户端取消但 Handler 仍执行、已写出 Bus 回调仍执行时按正常停止截止结束；停止或落盘失败、旧树未退出、迟到退出通知与数据保留。sdk/ 与 supervisor/ 验证全部留在本地，不提交测试插件。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/tune/... ./internal/core/idle/... ./internal/core/settings/... ./internal/core/state/... ./internal/core/supervisor/... ./internal/core/router/... ./internal/core/broker/... ./internal/core/httpproxy/... ./internal/core/media/... ./internal/core/continuation/... ./internal/core/hook/... ./internal/core/bus/... ./internal/app/...；go build ./...。PR 按 TEST-MATRIX §3 写明 V-77 本切片分项的测试位置、命令、平台、结果与未运行项，并列出 D-179 对 CONTRACTS/PLAN/TEST-MATRIX 的文档同步；双进程、重复交付、请求期限延长、休眠改代次或重复迁移均按 S-02 失败停止条件阻止合入。

不做：Scheduler 到期唤醒（goal-172），休眠管理 API/CLI/WebUI（goal-173、goal-175），固定机器资源测量与参数调优（goal-174）；私建后台工作的观测或新业务 RPC、自动重试。不改 .proto 或生成产物，不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

## 第 20 波

### goal-162 · T-04.19 平台与消息

状态：已完成（PR #250），合并提交 `f3a0645`。以下保留派发时语句。派发时：前置 T-04.18（PR #242，合并提交 `b8161272`）已合并；配置文件存储 T-08.18（PR #214，`ab0cd56`）、宽松载入 T-08.16（PR #239，`1a1d12c3`）均已合并。已按 main `9330f399` 核对：`internal/core/profile.New(ctx context.Context, opt profile.Options) (*profile.Service, error)`、`(*Service).Resolve(instance, session string) profile.Profile` 与 `Platform(id string) (profile.Platform, error)` 可读平台覆盖，`PutPlatform(ctx context.Context, id string, platform profile.Platform) error` 已支持 default 写回；`internal/core/settings.Settings.Reply` 与 `settings.Reply{MentionSender bool; Quote bool; Prefix string; ForwardThreshold int}` 已实现解析、保存和比较。Router 仍使用 `router.Options.Prefixes func(adapterInstanceID string) []string`，消息管道为 `internal/core/message.(*Service).Send/Edit(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response`，调用根由 `router.Invocation.Root`、`Begin() func()` 与 `AfterFinished(finished func())` 跟踪；平台设置的执行尚待接入。 审查者补充（2026-11-18）：首条回复与原消息的关联只在本 Session 内维持、不持久化；插件重启或换 Session 后不再按关联编辑早先的回复，按新回复处理。

与 goal-163 共享 `internal/core/profile/`（profile.go、platform.go、tools.go 的操作快照）、`internal/app/profile.go`、assembly.go、alonabot.go、`internal/core/manage/profile_api.go` 及 `cmd/alonabot/manage_e2e_test.go` 的配置文件测试装配。goal-162 负责平台有效值、前缀与回复，goal-163 负责工具策略；两者可并行，后合并者同步最新 main，保留双方的快照读取、回调注入、固定服务注册、Session 清理和测试断言，重跑两条主路径。与 goal-171 相交的 settings 与 app 文件同时保留 sleep、reply 及 T-08.16 的校正/原文保留行为。

```text
/goal 使用 $alonabot-execute 推进 goal-162（切片 T-04.19：平台与消息）。goal 目录 goal-162，分支 goal-162，worktree .worktrees/goal-162，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：命令前缀按会话所用的配置文件；删除 `command_prefixes_by_instance`；Core 在发送管道中为回复加上提及、引用、前缀与合并转发。

依据：PLAN T-04 步骤 19 及其失败停止条件；CONTRACTS §2.2、§5.5、§5.6（平台设置）、§8、§15.1；MESSAGE-MODEL §1、§2、§4、§4A、§5；ADMIN-API §13.1、§23；DECISIONS D-168、D-135；TEST-MATRIX V-87、V-142。

范围（只新增或修改这些）：
- internal/core/profile/platform.go 及平台有效值的包内辅助文件；internal/core/settings/settings.go、file.go、correction.go、reply.go 中旧实例前缀删除与现有 reply 的必要衔接
- internal/core/router/：dispatch.go、router.go、command.go、mux.go、trusted.go、profile.go、commandlist.go 中会话前缀、可信根关联与自动回复调用点；internal/core/broker/、hook/、continuation/ 中根附件继承的必要调用点
- internal/core/message/：新文件 reply.go，send.go、edit.go 中回复处理、首条回执关联、Hook 前改写与编辑限制接入
- internal/app/：profile.go、assembly.go、alonabot.go、journal.go 与装配/退出清理的必要辅助文件；internal/core/manage/settings_api.go 删除旧前缀路由，profile_api.go 保留现有平台接口
- cmd/alonabot/manage_e2e_test.go 中既有 V-87 测试改写及一个 V-142 集成子测试；docs/development.md 的相关说明；上述包的 *_local_test.go

允许的依赖：标准库与已批准依赖；复用 profile.Service、settings.Reply、router.Invocation、message.Service、adapter.Registry 及现有 TextEntity/Forward 载荷。沿用 ARCHITECTURE 的 internal/core 内部依赖与 internal/app 装配，通过函数注入平台读取；Core 不 import SDK 或平台 Adapter 实现。

接口（现有签名保持可用）：
- profile.New(ctx context.Context, opt profile.Options) (*profile.Service, error)；(*profile.Service).Resolve(instance, session string) profile.Profile；Platform(id string) (profile.Platform, error)；PutPlatform(ctx context.Context, id string, platform profile.Platform) error
- profile.Platform{CommandPrefixes *[]string; Reply profile.Reply}；profile.Reply 的 MentionSender/Quote/Prefix/ForwardThreshold 为指针，nil 跟随 default；settings.Reply 为对应完整值，零值关闭全部回复修饰
- (*settings.File).Settings() settings.Settings；Update(edit func(doc *settings.Document) error) error；Reload() (changed []string, err error)
- message.New(adapters *adapter.Registry, p *policy.Policy, lease func(context.Context, router.Caller, string, *pb.Target, *pb.MediaSource, time.Time) (*pb.MediaSource, func(), error), issue func(context.Context, router.Caller, *pb.Target, *pb.MessageChain) error, requestTimeout time.Duration, hooks message.Hooks) *message.Service
- (*message.Service).Send(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response；Edit(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response；message.Hooks.BeforeSend/AfterSend 沿用现有签名
- router.Invocation.Root *router.Invocation；(*router.Invocation).Begin() func()；AfterFinished(finished func())；(*router.Router).Profile(ctx context.Context, pluginID string, target *pb.Target) (string, bool, error)

接口（本切片新增或替换，以下签名固定）：
- func (s *profile.Service) EffectivePlatform(id string) (prefixes []string, reply settings.Reply, err error)：按键合并该配置文件覆盖与 default，Platform/PutPlatform 继续表示已保存覆盖。
- router.Options.Prefixes 替换为 func(ctx context.Context, target *pb.Target) ([]string, error)，同步 Router 内部字段和全部调用点；由装配用可信 Target 与已有 targetProfile/Resolve 读取有效前缀。删除 settings.Settings.CommandPrefixesByInstance、(*settings.File).CommandPrefixes(adapterInstanceID string) []string 及旧实例覆盖分支，不保留旧签名。
- func (s *message.Service) SetReply(resolve func(context.Context, *pb.Target) (settings.Reply, error))：装配同一会话解析与 EffectivePlatform；func (s *message.Service) Release(c router.Caller)：Session 退出时释放该 Caller 的首条回执关联。已有 New 签名保持可用，未注入时按全部回复设置关闭运行。

行为：
- 命令前缀按可信事件的完整会话解析：单独指定优先、实例缺省其次、default 最后，话题与转义沿用 internal/app/profile.go 的 targetProfile。其他配置文件未覆盖时取 default；仍支持非空列表、空字符串前缀与最长匹配、去开头 reference/本账号 mention、账号用户名和原命令消费规则。读取失败保留错误并结束本次匹配，诊断包含会话/插件定位，不绕回另一配置文件。usage 使用本次命中的前缀；core.command.list 与平台菜单保留各自既有命令目录语义。
- 删除 command_prefixes_by_instance 的声明、解析、默认文档、快照字段、比较和保存入口，删除 GET/PUT /api/v1/core/command-prefixes。磁盘已有旧键依 T-08.16 作为 Unknown 保留并忽略，管理员 prune 时清理，不做迁移；按账号区分前缀改为指定该实例的缺省配置文件。保留已实现的 core_config.json.reply 解析/保存与 ADMIN-API §23 的平台 GET/PUT，更新 docs/development.md，V-87 的旧实例覆盖断言改为会话配置文件断言。
- 回复仅指处理入站 message.created 的命令、文本匹配、事件路由及其合法派生/延续调用，发往该消息所在的同一实例、会话与话题。以 Core 保存的可信根事件、DefaultTarget 与 Root 判定；显式 Send、自动回复及 Service 子调用共享根的首条标记。后台、scheduled、其他事件和发往其他会话的发送保持原行为；插件修改载荷不能替换根事件绑定。
- 根调用及其延续向原会话的第一次 Send 原子取得首条标记，前三项只对它处理：群聊 mention_sender 提及可信 Actor 的用户，已有同用户提及则保留；quote 引用触发消息，已有 reference 则保留；prefix 放在第一个 text 前，没有 text 则新增首个 text。私聊不自动提及，缺 Actor 时不伪造发送人。保留 root 的 Begin/AfterFinished 引用与锁外 reload，首条判定不把平台 I/O 放进 Router/Broker 锁。
- 先按目标实例 message.send.text_limit 的单位检查加入提及/前缀后的单条长度；超过上限时两者都不加，引用照加，不截断原内容。发送/编辑在出站 Hook 之前改写，并走既有结构、能力、扩展和 Hook 复检；保留 text/entities 的 UTF-16 偏移。Edit 仍接受一个 text 和可选 buttons，自动提及可用现有 TEXT_MENTION 实体表达，复用既有实体及渲染规则；无法表达的组合沿原 UNSUPPORTED/INVALID_ARGUMENT 报错，不丢弃内容。
- 记录首条发送实际返回的平台消息 ID，以及实际采用的自动提及和前缀；部分结果与错误同时保留。此后编辑该 ID 保留这些修饰，不再追加引用，不重复叠加前缀/提及；其他消息的 Edit 保持原规则。关联按 Caller.SessionKey、完整 Target 与平台 ID 隔离，在 Release 时清理；根的首条标记在全部派生与延续结束后释放。
- forward_threshold 为正数且目标实例声明支持 message.forward 时，对每次 Send 的 text 片段按 Unicode 码点计总字数，超过阈值则将该条改为一个 forward 片段，节点来源为本账号。未声明时照常发送；已有 forward 的嵌套限制、reference、媒体及按钮组合继续按现有整链校验处理，保留原错误。合并转发也在 Hook 之前完成，阈值不作用于 Edit。
- 平台设置写入后后续派发/发送读取新快照，无需插件 reload；default 的值沿 settings.Update/Reload 生效。复用 goal-161 的配置文件解析、profile_id、插件开关、冻结视图和根结束通知，不另建配置真源。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 一个真实 Core＋插件子进程＋回环 Adapter 的 V-142 集成子测试，经既有 profile/platform/assignment 管理入口设置 default 和另一配置文件，用两个会话核对各自前缀与回复；首条手动 Send、后续自动回复共用首条状态，带编辑能力的回环实例核对首条 Edit 保留提及/前缀，Hook 观察改写后的 chain。既有 V-87 子测试同步删除旧接口断言，换用平台 API，保留原命令/权限断言，不另增同类测试或 testdata 插件。
- 本地自测（*_local_test.go，结果写入 PR）：V-87/V-142 其余分项，包括空前缀/最长匹配、实例缺省与话题/转义、平台按键继承和即时更新；命令/文本/事件/Service/延续共用首条、并发 Send、后台/异会话不变、已有提及/引用、无 Actor、失败/部分回执与 Session 隔离；UTF-8/UTF-16/码点上限、补前缀后实体偏移、Edit 的限制及重复修饰、支持/不支持 forward 与阈值、旧键 Unknown/prune、旧 HTTP 路由移除。SDK 只写本地自测。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/settings/... ./internal/core/profile/... ./internal/core/router/... ./internal/core/message/... ./internal/core/broker/... ./internal/core/hook/... ./internal/core/continuation/... ./internal/core/manage/... ./internal/app/... ./cmd/alonabot/...；go build ./...。PR 按 TEST-MATRIX §3 记录 V-87/V-142 的测试位置、命令、平台、结果及未运行项；回复作用于后台/异会话、破坏完整回执或绕过配置文件开关时停止交付。

不做：配置文件存储与 API 的重建（goal-160）、SDK 配置视图（goal-161）、工具策略（goal-163）、页面（goal-165）；不改 proto，不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-163 · T-11.6 按配置文件的工具策略

状态：已完成（PR #246），合并提交 `24e57d6`。以下保留派发时语句。派发时：前置 T-04.18（PR #242，合并提交 `b8161272`）与 T-11.4（PR #172，`15f94955`）均已合并；配置文件存储 T-08.18（PR #214，`ab0cd56`）、宽松载入 T-08.16（PR #239，`1a1d12c3`）也已合并。已按 main `9330f399` 核对：`internal/core/profile.(*Service).Tools(id string) (profile.ToolPolicies, error)` 与 `PutTools(ctx context.Context, id string, policies profile.ToolPolicies) error` 已提供工具/操作覆盖及 inherit 删除，ADMIN-API §23 的路由位于 `internal/core/manage/profile_api.go`；`internal/core/tool.New(store *state.Store, r *router.Router) *tool.Catalog`、`(*Catalog).Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error)` 当前只查全局覆盖与声明，`Register(m *router.Mux)`/`RegisterInvoke(m *router.Mux, opt tool.InvokeOptions)` 注册 list/invoke/confirm。`internal/app/profile.go` 的 targetProfile 按可信 Target 解析会话，`router.Invocation` 已携带 DefaultTarget/ProfileID/Root；本切片接入运行时策略。

与 goal-162 共享 `internal/core/profile/`（profile.go、platform.go、tools.go 的操作快照）、`internal/app/profile.go`、assembly.go、alonabot.go、`internal/core/manage/profile_api.go` 及 `cmd/alonabot/manage_e2e_test.go` 的配置文件测试装配。goal-163 负责工具策略，goal-162 负责平台有效值、前缀与回复；两者可并行，后合并者同步最新 main，保留双方的快照读取、回调注入、固定服务注册、Session 清理和测试断言，重跑两条主路径。配置读取保留 goal-152 已落地的无版本载入与宽松校正。

```text
/goal 使用 $alonabot-execute 推进 goal-163（切片 T-11.6：按配置文件的工具策略）。goal 目录 goal-163，分支 goal-163，worktree .worktrees/goal-163，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：工具策略先取调用所在会话的配置文件中的覆盖，再取全局策略；配置文件关闭了工具所属插件时视为 deny；`core.tool.list` 返回按本次调用生效的策略。

依据：PLAN T-11 步骤 8 及其失败停止条件；CONTRACTS §3、§5.6、§13.1、§13.2；WIRE-FORMAT §3.5、§4；ADMIN-API §18、§23（tools）；DECISIONS D-168；TEST-MATRIX V-143。

范围（只新增或修改这些）：
- internal/core/profile/tools.go 及策略读取的包内辅助文件，profile.go、plugins.go 仅限共用操作快照所需的调用点
- internal/core/tool/catalog.go、invoke.go、confirm.go 及生效策略的包内辅助文件
- internal/app/profile.go、assembly.go、alonabot.go 中快照读取与 Catalog 装配；internal/core/manage/profile_api.go、tool_api.go 仅为复用既有管理读写入口所需的调用点
- cmd/alonabot/manage_e2e_test.go 中一个配置文件工具策略集成子测试及既有宿主的必要装配；上述包的 *_local_test.go

允许的依赖：标准库与已批准依赖；复用 profile.Service、tool.Catalog、state.Store、router.Invocation、既有 Broker 与确认按钮服务，沿用 ARCHITECTURE 的 internal/core 内部依赖与 internal/app 装配。通过注入函数建立一次策略操作的 profile 快照，不另建存储或策略缓存。

接口（现有签名保持可用）：
- profile.New(ctx context.Context, opt profile.Options) (*profile.Service, error)；(*profile.Service).Resolve(instance, session string) profile.Profile；Plugins(id string) ([]profile.Plugin, error)；Tools(id string) (profile.ToolPolicies, error)；PutTools(ctx context.Context, id string, policies profile.ToolPolicies) error
- profile.ToolPolicies{Items []profile.ToolPolicy}；profile.ToolPolicy{Tool string; Operation *profile.Operation; Policy string}；profile.Operation{PluginID, Name string}
- tool.New(store *state.Store, r *router.Router) *tool.Catalog；(*tool.Catalog).Effective(ctx context.Context, pluginID, service string) (pb.ToolPolicy, error)；Policies(ctx context.Context) ([]tool.ToolPolicyEntry, error)；SetPolicy(ctx context.Context, rows ...state.ToolPolicy) error
- (*tool.Catalog).Register(m *router.Mux)；RegisterInvoke(m *router.Mux, opt tool.InvokeOptions)；tool.InvokeOptions.Call func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response 继续转交原 Broker
- router.Invocation.DefaultTarget *pb.Target、ProfileID string、Root *router.Invocation；wire.Inbound.ParentCallID 与 ParentAttach 继续由同一 Session 的在途父调用恢复

接口（本切片新增，以下签名固定）：
- func (s *profile.Service) ToolPolicy(id, pluginID, name string, operation bool) (policy pb.ToolPolicy, enabled bool, err error)：只读本次快照；operation=false 查工具键，true 查确认操作键；无覆盖返回 UNSPECIFIED，enabled 为该配置文件的插件开关，未设开关为 true。与 Tools/PutTools 使用同一快照条目及 allow/confirm/deny 编码。
- func (c *tool.Catalog) SetProfile(resolve func(ctx context.Context, target *pb.Target) (profiles *profile.Service, profileID string, err error))：装配用现有 targetProfile/Resolve 返回一次操作的快照及选择。每次 list/invoke/confirm 只调用一次，list 的全部条目复用结果；未注入时沿用原全局策略入口，已有全局 Effective/Policies/SetPolicy 的含义和签名保持可用。

行为：
- 仅从 Session 恢复的可信父附件取 DefaultTarget，按同一实例、会话字符串、话题及转义规则解析当前配置文件；无默认 Target 时使用 default。list 保留无父调用的现有用法，此时读取 default；携带 parent_call_id 时仍须关联该 Session 的有效在途父调用。invoke/confirm 保留现有有效父调用要求，不采信请求 JSON、Actor 或插件自行提供的 profile_id。
- 同次策略操作在同一快照按顺序查找：配置文件关闭工具所属插件时直接 deny；有该工具或操作覆盖时取覆盖；否则取原全局覆盖，再取工具声明 default_policy（操作缺省 confirm）。default 同样可保存插件开关和工具覆盖。继承键为空时沿原全局入口，不为未配置项伪造策略。
- 工具以 plugin_id/service 为键；core.tool.confirm 以调用插件 ID 与 operation 为键，工具覆盖与操作覆盖按现有 operation 标记分别查找。确认操作仍先记录名称与 effect，再判断策略；effect 的管理目录合并与提示语义保持原样。inherit 经现有 PutTools 删除指定覆盖，下次操作恢复全局/声明策略。
- core.tool.list 仍只列出当前已激活、调用者按 Service 访问规则可见且生效策略非 deny 的工具，返回本次配置文件下的 allow/confirm 与原 description、Schema、effect；配置文件关闭的插件工具同时从列表消失。策略读取失败保留实际错误并结束本次操作。
- invoke 的 deny 返回 PERMISSION_DENIED；allow 与确认通过均经原 Broker 继承父调用链执行一次，confirm 保留可信 Actor/默认 Target、按钮能力、同一用户、期限与取消规则。操作确认同样使用配置文件覆盖及插件开关；全局 allow 无法绕过配置文件 deny/关闭。新策略只影响随后开始的策略操作，保持已接纳调用的原策略快照。
- 管理继续使用 ADMIN-API §23 的 profile.Tools/PutTools 与 §18 的 Catalog.Policies/SetPolicy，各自保留原响应、原子写入和错误规则。修改即时生效、无 Pending/reload，不撤销在途调用；普通跨插件 Service、定时任务和后台发送继续按原规则运行。管理查询的全局策略仍表示全局覆盖，不能被本次聊天配置文件替换。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/manage_e2e_test.go 一个真实 Core＋工具提供插件/调用插件子进程＋SQLite＋回环 Adapter 的 V-143 集成子测试，复用已有动态构建宿主和确认按钮路径：全局 confirm，两个会话分别指定 allow/deny 配置文件，核对 list/invoke 的差异与真实调用计数；inherit 删除后恢复全局 confirm，并经真实点击确认执行；关闭工具所属插件后 list 隐藏、invoke 拒绝。复用原测试工程，不新增 testdata 插件。
- 本地自测（*_local_test.go，结果写入 PR）：V-143 其余分项，包括 core.tool.confirm 的操作覆盖与 effect 记录、工具/操作同名隔离、default 和无 Target、单独指定/实例缺省/话题/转义、list 无父调用及非法父调用；配置文件开关优先于 allow、管理更新下一次生效而在途确认保留原策略、读取错误、inherit 恢复、全局 API 原义、原 Service 权限/取消/期限与结果错误并存。SDK 中只写本地自测。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/profile/... ./internal/core/tool/... ./internal/core/router/... ./internal/core/broker/... ./internal/core/manage/... ./internal/app/... ./cmd/alonabot/...；go build ./...。PR 按 TEST-MATRIX §3 写明 V-143 的测试位置、命令、平台、结果与未运行项；deny/关闭仍执行、跨会话策略混用、丢失父调用权限或重复执行工具时停止交付。

不做：配置文件存储与管理 API 的重建（goal-160）、平台与消息（goal-162）、页面（goal-165）、新增工具发现/休眠唤醒机制；不改 proto，不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-141 · T-13.6 CLI 参数从简

状态：已完成（PR #233），合并提交 `4057ce2`。以下保留派发时语句。派发时：前置 T-13.5（[PR #200](https://github.com/AlonaBot/AlonaBot/pull/200)，合并提交 `2e5ac2b`）与 T-14.3（[PR #228](https://github.com/AlonaBot/AlonaBot/pull/228)，合并提交 `72d51c2`）均已合并。按 main `25e9064` 核对：CLI 已直接使用 `internal/app.Start(ctx context.Context, o app.Options) (*app.Core, error)`，根包只保留库模式接口；`internal/pluginkit` 已有 `Init(ctx context.Context, dir, id, role, adapterID, module string) (bool, error)`、`Compile(ctx context.Context, options generate.Options, outDir string) (Build, error)`、`CompileApp(ctx context.Context, options generate.Options, outDir string, goFlags ...string) (string, error)`、`Check(ctx context.Context, options generate.Options) error` 与单平台 `Pack(ctx context.Context, options generate.Options, assetURL, keyFile, outDir string) (string, error)`；`cmd/alonabot/plugin.go:componentFlags` 仍登记 `--dir`、`--goos`、`--goarch` 与 `--tags`，`init` 的 `--id` 和 `pack` 的 `--asset-url` 仍必填。`scripts/dist/main.go:packComponent` 逐平台调用 `pack --dir --goos --goarch`，合并各次的 assets 后签名；本切片同步将这些调用改为位置目录与单个 `--target`，保留 dist 的合并和签名流程。发行包实际布局为可执行文件旁的 `packages/<组件 ID>/`，各目录含本平台归档、合并全部平台的 `release.json`、`release.json.sig` 与 `release.pub`；当前静态 WebUI 为 `any/any`。`contracts/release.Verify(raw, sigFile []byte, pub ed25519.PublicKey) (*Release, error)`、`(*Release).Asset(os, arch string) (Asset, bool)` 与 `internal/core/trust.Official() (ed25519.PublicKey, bool)` 已可复用。T-09.27（[PR #222](https://github.com/AlonaBot/AlonaBot/pull/222)，合并提交 `5940406`）已合并，四角色 ID 统一经 `contracts/manifest.ParseProject(data []byte) (*Result, error)` 校验；V-121 的目录名缺省与 D-172 补验由本切片承接。

与第 20 波的 goal-74（T-14.4）共享发行包的 `packages/<组件 ID>/` 布局、CLI 安装入口及说明；本切片负责 `cmd/alonabot/manage.go` 的参数解析和查包，goal-74 负责 `internal/app/` 的首次预装与 `internal/core/manage/` 的更新提示、Adapter 启用。`packages/` 在本切片中只读，安装仍经管理 API。两者可并行；若 goal-74 定稿后也涉及 `cmd/alonabot/manage.go`、`cmd/alonabot/plugin_cli_e2e_test.go`、`cmd/alonabot/t12_e2e_test.go` 或 `docs/development.md`、`docs/plugin-guide.md`、`README.md`，后合并者同步最新 main，保留位置参数、包发现与首次预装两边的行为，重跑相应 CLI 和首次运行集成测试。合并本切片后，审查者按实际接口更新 goal-74 及模板仓库的 `scripts/release.sh`；模板仓库修改由审查者另行同步。

```text
/goal 使用 $alonabot-execute 推进 goal-141（切片 T-13.6：CLI 参数从简）。goal 目录 goal-141，分支 goal-141，worktree .worktrees/goal-141，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：CLI 的目录用位置参数，不给时取当前目录；不给的文件在当前目录寻找；其他参数取官方或能唯一推出的缺省值，推不出时报错并列出候选。根命令仍为 alonabot。协议与管理 API 不变；修改 CLI、其既有工程构建与打包实现、发行脚本调用和说明，落实 D-149 与 D-169。

依据：PLAN T-13 步骤 7 及其失败停止条件；CONTRACTS §9.3、§15.2，打包与验签沿用 §9.1、§9.2；DECISIONS D-149、D-169、D-172；TEST-MATRIX V-121（含 T-09.27 留下的目录名缺省分项）；ARCHITECTURE 的 cmd/alonabot、internal/pluginkit 与 scripts/dist 依赖边。

范围（只新增或修改这些）：
- cmd/alonabot/plugin.go、init.go、build.go、check.go、test.go、pack.go、manage.go：目录位置参数、缺省值、多平台打包入口与本地包发现；发现辅助函数放在 manage.go，保持未导出。
- internal/pluginkit/init.go、pack.go：工程缺省值与多平台打包，复用既有扫描、构建、归档及签名实现。
- scripts/dist/main.go：packComponent 改为每个组件调用一次 `alonabot pack <组件目录> --target ...（全部平台） --asset-base <前缀> --out <目录>`，删除自行合并清单的代码，签名、公钥注入与发行目录布局保持（D-174）。
- cmd/alonabot/plugin_cli_e2e_test.go、t12_e2e_test.go：下文允许的一个主路径集成测试及既有命令写法同步。
- docs/development.md、docs/plugin-guide.md、README.md：命令示例、参数说明与发行构建调用说明。
- 上述 Go 包的 *_local_test.go（本地自测，不提交）。
允许的依赖：标准库与 docs/development.md 已批准依赖；沿用 cmd/alonabot -> internal/pluginkit、internal/core/manage、contracts/...，internal/pluginkit -> internal/generate、contracts/...，scripts/dist -> contracts/release；scripts/dist 继续经 os/exec 调用主机 CLI。允许的新跨模块边：无。

接口（照此实现）：
- CLI：build [目录]、check [目录]、test [目录]、pack [目录] 均至多一个目录参数，缺省为当前目录；删除 --dir。init [目录] 保留 --role、--id、--module、--adapter-id；pack 增加可重复的 --target <os>/<arch>，删除 pack 的 --goos、--goarch 与 --asset-url，新增 --asset-base，保留 --tags、--out、--key（D-174）。build/check 的 --goos、--goarch 与 --tags 保持现有含义。
- internal/pluginkit 既有签名保持：Init(ctx context.Context, dir, id, role, adapterID, module string) (bool, error)；Compile(ctx context.Context, options generate.Options, outDir string) (Build, error)；CompileApp(ctx context.Context, options generate.Options, outDir string, goFlags ...string) (string, error)；Check(ctx context.Context, options generate.Options) error；单平台 Pack(ctx context.Context, options generate.Options, assetURL, keyFile, outDir string) (string, error)。generate.Options 的 Dir、GOOS、GOARCH 为 string，Tags 为 []string；既有直接调用 Pack 的集成测试继续使用单平台入口。
- 新增 internal/pluginkit.PackTargets(ctx context.Context, options generate.Options, targets []manifest.Platform, assetURL, keyFile, outDir string) (archives []string, err error)：由 CLI 调用；targets 为空时取 runtime.GOOS/runtime.GOARCH，显式目标使用 contracts/manifest.Platform{OS, Arch}；返回本次输出的全部归档路径，共享一份 release.json。复用单平台 Pack 的构建和归档实现，合并与最终签名归本入口；内部辅助函数自定。
- cmd/alonabot/manage.go 保持 localInstallRequest(archive, releaseFile, signatureFile, keyFile, source string) (manage.InstallRequest, error)，在 install 分支调用前完成归档定位。manage.InstallRequest 仍为 Archive、ReleaseJSON、Signature、PublicKey、Source string 与 Replace、Confirm bool；仍向 POST /api/v1/packages/install 发送，服务端沿用 (*manage.Service).Install(ctx context.Context, req manage.InstallRequest) (manage.Package, error)。
- 复用 contracts/manifest.ParseProject(data []byte) (*manifest.Result, error)，Result 含 Manifest *manifest.Manifest 与 Warnings []string；contracts/release.Verify(raw, sigFile []byte, pub ed25519.PublicKey) (*release.Release, error)、(*release.Release).Asset(os, arch string) (release.Asset, bool)；internal/core/trust.Official() (ed25519.PublicKey, bool)。release.Release.Assets 为 []release.Asset，资产含 OS、Arch、URL string、Size int64 与 SHA256 string；校验与信任仍由原所有者负责。

行为：
- build、check、test、pack 按位置目录工作；库模式 build 同样作用于给出的 main 包目录。目录参数、资源只读、overlay、编译器错误原文与离线测试生命周期沿用原入口；--dir 按未知参数报错。
- init 未给 --id 时取目标目录绝对路径的目录名，未处于已有 Go module 中且未给 --module 时取同一目录名；已有 module 内沿用原 module，不创建嵌套 go.mod。显式 --id、--module 继续生效。ID 通过 ParseProject 按 D-172 的角色市场前缀与本地规则校验：alonabot-plugin-hello 可直接作为插件市场 ID，my_plugin 作为本地 ID 并告知缺前缀警告；非法目录名报错并提示显式 --id，保留原字符。Adapter 未给 --adapter-id 时等于最终组件 ID，按既有本地平台 ID 规则检查；目录非空仍在写入前拒绝。
- 每个资产地址为 --asset-base 加 / 与归档名；未给 --asset-base 时，manifest 的 source 为 github.com/<owner>/<repo> 才推出 https://github.com/<owner>/<repo>/releases/download/v<version>；其他来源要求显式 --asset-base，沿用绝对 https 校验（D-174）。--out 缺省为工程 bin/，--key 无缺省，仍读取标准 base64 的 64 字节 Ed25519 私钥。
- pack --target 逐平台构建，沿用该平台的源码选择、资源检查、包内 manifest 与 build-info；输出每个平台的归档和一份含全部 assets 的 release.json。静态 WebUI 只输出一个 any/any 资产；有 Go 入口的 WebUI 按目标平台构建。给出 --key 时只对最终清单原始字节签名一次，写出 release.json.sig 与 release.pub；保留私钥资源排除、构建失败报错、可重复归档及 unsigned 输出清除旧签名的行为。每个资产沿用既有归档完整性字段，最终清单用 release.Verify 自检。
- scripts/dist 每个组件调用一次 alonabot pack <组件目录>，按 --platforms 给出全部 --target 与 --asset-base，由 pack 写出合并后的 release.json；dist 继续不向 pack 传 --key，自行签名并用 release.Verify 自检。当前静态 WebUI 的 any 包、各平台 packages/<组件 ID>/ 与 dist/packages/<组件 ID>/ 布局、公钥注入和私钥仅留在 dist 进程的约定保持原状。
- install 保留 --source owner/repo 的下载、--archive 等显式文件选项及其互斥规则；参数是已存在的普通文件时按本地包安装，否则按组件 ID 从 os.Executable 所在目录的 packages/<组件 ID>/ 读取清单，选择本机平台资产（静态 WebUI 取 any/any）及对应归档。来源取 release.json，沿用官方公钥信任、签名、归档预检与 --yes/--replace 确认；缺少官方包时报错并列出可用的官方组件 ID。
- install 无位置参数且未显式选择 --source 或 --archive 时，在当前目录寻找安装包；当前目录含工程 plugin.yaml 时使用其 bin/。按 release.json 的平台资产与归档文件筛选本机平台候选（含静态 WebUI 的 any/any），恰有一个才安装；零个或多个时报错并列出搜索目录及候选，在调用安装 API 前结束。归档外 release.json、release.json.sig、release.pub 的缺省位置与显式覆盖沿用 localInstallRequest；读取、解析和文件系统错误保留。
- 文档与 CLI 帮助同步以上用法；V-121 中 D-172 的目录名缺省分项由本切片提供证据。首次运行预装、版本更新提示与 Adapter 自动启用由 goal-74 实现；两者共享 packages/<组件 ID>/ 的只读布局。若并行改到 cmd/alonabot/manage.go、上述既有测试或三份说明，后合并者同步最新 main，保留参数解析、包发现与首次预装两边的行为，重跑相应 CLI 与首次运行集成测试，按 docs/execution-order.md §3 处理。

测试（按编码规范「测试提交规则」）：
- 提交：只扩展 cmd/alonabot/plugin_cli_e2e_test.go 的 TestDevelopmentCLI（//go:build unix）这一个主路径集成测试：在合法市场 ID 命名的临时工程目录省略 --id、--module 运行 init，经真实 CLI 的位置目录/当前目录 build、check、test，再不带 --asset-base 打包本机及另一个支持平台，临时密钥签最终清单；核对缺省 ID/module、两个资产的 URL、验签与本机包的既有预检断言；复用真实 Core 测试入口，从该工程 bin/ 不给安装参数安装本机包，核对登记的 ID 与来源。既有库模式 build 改为位置目录并保留断言；t12_e2e_test.go 中 pack --dir 改为位置目录，原断言不变。保留其他既有测试和直接调用单平台 Pack 的方式。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：V-121 其余分项——四个命令给出另一目录与省略目录、--dir 报未知参数；init 的四角色市场 ID、本地 my_plugin 警告、非法目录名与显式 ID、已有 module 内外、Adapter ID 缺省及显式值；GitHub/非 GitHub source 与显式 URL；pack 两个和三个目标、不给 --target 时本机平台、静态 WebUI any 去重、带进程 WebUI 的平台资源、unsigned/签名与私钥排除、失败保留错误；install 显式普通文件、官方 ID 存在/不存在、本机/any 资产选择、当前目录与工程 bin/ 的零/一/多个候选、显式元数据覆盖、--source/--archive 既有路径及歧义时无安装副作用。D-172 的 V-121 目录名缺省单列结果。
- 复用 scripts/dist/dist_test.go:TestDistribution 验证 CLI 调用同步后的归档布局、签名、内置公钥与可重复性，保留其既有断言；多平台发行调用在 scripts/dist/*_local_test.go 核对，不新增提交测试。其他平台只交叉构建时按交叉构建记录。

完成条件：gofmt -l 对改动 Go 文件无输出；git diff --check；go vet ./...；go test -race ./cmd/alonabot ./internal/pluginkit ./scripts/dist；go build ./...。PR 写明 V-121 各分项的测试位置、命令、平台、结果及未运行项；「文档同步」列出需审查者回写的 PLAN T-13.6、TEST-MATRIX §3.2、goal-74 接口与模板仓库 scripts/release.sh。遇到契约缺口、范围外修改或新依赖，按共同约束标阻塞，交审查者处理。

不做：根命令改名；缺省值扩大卸载、清理与回退等破坏性操作的对象；协议与管理 API 的改动；首次运行预装、更新提示与 Adapter 自动启用（goal-74）；重写 dist 的合并签名流程；直接改模板仓库；新增依赖。
```

### goal-172 · S-03.1 Scheduler 唤醒与离线 tick 区分

状态：已完成（PR #256），合并提交 `216bb7f9`。以下保留派发时语句。派发时：前置 S-02.2（goal-171，PR #249，合并提交 `18f8af95`）、T-07.2（goal-24，PR #107，合并提交 `fc2c2471`）均已合并；S-02.1（goal-170，PR #241，合并提交 `e0a38d7c`）已提供唤醒与活动门禁。已按 main `54ac4f8a` 核对：`internal/core/scheduler` 的 `New(opt Options) *Scheduler`、`(*Scheduler).Attach(ctx context.Context, p router.Plugin, activatedAt time.Time) error`、`Drain(pluginID string)`、`Detach(pluginID string)`、`Released(pluginID string)` 已接入生命周期；当前 `Drain` 调用 `Detach`，会停止定时器并清空注册，`tick`、`tickTask` 与 `pump` 依赖旧 `router.Plugin.Activated()`，尚无到期唤醒入口。`internal/core/supervisor.(*Waker).EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error)`、`AcquireActivity(pluginID string, session *wire.Session) (func(), error)` 与 `Ready(running *Running)` 已装配；`(*Running).StopIdle(t0 time.Time, lastActivityAt time.Time) host.StopResult` 经正常停止写入 SUSPENDED，Frozen、detach、release 分别调用 Scheduler 的 Drain、Detach、Released。`internal/core/idle.(*Controller).CanStop(pluginID string, policy Policy, runtime state.RuntimeStatus, activity supervisor.ActivityStatus, now time.Time) bool` 以 Ready、零引用及实际活动时间判定空闲。`internal/app/assembly.go`、`journal.go`、`wake.go`、`idle.go` 已串起 Scheduler、Waker 与自动空闲停止，`assembly.detach` 仍无条件撤销当前组件的调度注册。`internal/core/state.(*Store).Jobs(ctx context.Context, pluginID string) ([]Job, error)`、`Tasks(ctx context.Context, pluginID, name string) ([]Task, error)` 与 `RuntimeStatus(ctx context.Context, pluginID string) (RuntimeStatus, error)` 已提供持久调度与六态读取；复用现有表。空闲退出沿用 CONTRACTS §4 的正常停止（D-179）。

```text
/goal 使用 $alonabot-execute 推进 goal-172（切片 S-03.1：Scheduler 唤醒与离线 tick 区分）。goal 目录 goal-172，分支 goal-172，worktree .worktrees/goal-172，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：让静态与动态任务在插件正常休眠时按期唤醒执行，维持异常离线跳过、不追补及同一任务不重叠。

依据：PLAN §7 S-03 步骤 1、2 及其失败停止条件、spec/plan/core.md 的 S-03.1；CONTRACTS（spec/CONTRACTS-v0.1.md）§4.2～§4.5、§16；WIRE-FORMAT §5；DESIGN §H；DECISIONS D-19、D-53、D-124～D-126、D-153、D-179；TEST-MATRIX V-78 与 V-77 的 schedule 活动分项。

范围（只新增或修改这些）：
- internal/core/scheduler/：scheduler.go、tasks.go、trigger.go、jobs.go、services.go 与按职责拆出的休眠调度文件，覆盖 Attach/Drain/Detach/Released、静态 tick、动态到期队列、恢复与统一调用收尾
- internal/core/supervisor/：supervisor.go、prepare.go 中 Scheduler 的激活、空闲冻结、撤销与旧树退出通知；沿用 wake.go、activity.go、runtime.go 的现有门禁和状态接口
- internal/app/：assembly.go、journal.go、wake.go、idle.go、alonabot.go 中必要的注入、空闲停止、Session 解绑与退出清理调用点
- internal/app/scheduler_wake_test.go（新集成测试）；idle_test.go 中既有动态构建宿主的必要复用调整；上述包的 *_local_test.go

允许的依赖：标准库与现有已批准依赖；沿用 scheduler -> contracts/schedule、contracts/pb/alonabotv1、internal/core/router、internal/core/state、internal/wire，supervisor -> scheduler，internal/app 装配。Scheduler 通过函数注入调用同一个 Waker，保持依赖无环；状态读写继续归 state，空闲策略继续归 idle，调优值复用生效的 JOB_TIMEOUT、TASK_CONCURRENCY 与 WAKE_QUEUE_*。不增加依赖或调优参数。

接口（main 上的现有签名保持可用）：
- scheduler.Options 的 Store *state.Store、TasksPerPlugin int、TaskPayloadBytes int64、TaskConcurrency int、JobTimeout time.Duration、Now func() time.Time；func New(opt Options) *Scheduler
- (*scheduler.Scheduler).Attach(ctx context.Context, p router.Plugin, activatedAt time.Time) error；Drain(pluginID string)；Detach(pluginID string)；Released(pluginID string)；Close()
- (*scheduler.Scheduler).Stats(pluginID, jobID string) scheduler.JobStats；TasksStats(pluginID string) scheduler.TaskStats；Jobs(pluginID string) []scheduler.StaticJob；SetJobPaused(ctx context.Context, pluginID, jobID string, paused bool) error
- (*scheduler.Scheduler).Services(m *router.Mux, current func(router.Caller) (router.Plugin, bool))；Trigger(ctx context.Context, pluginID, jobID string, completed func(runID string, response wire.Response, err error)) error；TriggerTask(ctx context.Context, pluginID, taskName, key string, completed func(runID string, response wire.Response, err error)) error
- (*supervisor.Waker).EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error)；AcquireActivity(pluginID string, session *wire.Session) (func(), error)；Ready(running *supervisor.Running)；Activity(pluginID string) supervisor.ActivityStatus
- (*supervisor.Running).StopIdle(t0 time.Time, lastActivityAt time.Time) host.StopResult；Stop(t0 time.Time, reason pb.StopReason) host.StopResult；WaitExit() (abnormal bool, exitState string, err error)；supervisor.Spec.Detach func()
- router.WithActivity(p router.Plugin, acquire func() (release func(), err error)) router.Plugin；wire.Request.Deadline time.Time、Late func(wire.Response) 沿用实际响应与迟到响应的收尾规则
- (*state.Store).Jobs(ctx context.Context, pluginID string) ([]state.Job, error)；SaveJob(ctx context.Context, job state.Job) error；Tasks(ctx context.Context, pluginID, name string) ([]state.Task, error)；DeleteTask(ctx context.Context, pluginID, name, key string) error
- (*state.Store).Enabled(ctx context.Context, pluginID string) (bool, error)；Active(ctx context.Context, pluginID string) (string, error)；RuntimeStatus(ctx context.Context, pluginID string) (state.RuntimeStatus, error)；ActiveRouteSnapshot(ctx context.Context, pluginID string) (state.RouteSnapshot, error)
- (*idle.Controller).CanStop(pluginID string, policy idle.Policy, runtime state.RuntimeStatus, activity supervisor.ActivityStatus, now time.Time) bool；internal/app 的 assembly.ensureActive(ctx context.Context, id string, n int64) (func(), error)、assembly.startWake(context.Context, string) (*supervisor.Running, error) 与 Core.operation 继续负责完整恢复和串行生命周期

接口（本切片新增，以下签名固定）：
- scheduler.Options 追加 EnsureActive func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)，由 internal/app 注入 assembly.ensureActive；未注入时现有活 Session 调度入口继续可用
- func (s *Scheduler) Suspend(pluginID string)：在已取得正常 IDLE 停止权后撤销调度的旧 Session 交付绑定，保留计划、路由描述、静态起点、暂停状态、动态实例与未来定时器；与 Drain/Detach 的停用及异常离线清理分开。只调整调度内存状态，运行六态仍由 Supervisor 写入

行为：
- 正常 IDLE 退出沿用 D-179：StopIdle 在逐插件锁和活动门禁复核后进入 QUIESCING，按同一 STOP_DRAIN/STOP_ONSTOP/STOP_TOTAL 冻结、排空、STOPPING 与确认旧树退出，成功后写 SUSPENDED。Frozen 时通知 Scheduler.Suspend；随后 Supervisor.detach、Spec.Detach 回调、assembly.detach 与 app 的再次收尾只撤销该旧 Session 的 Router/Bus/HTTP 等绑定，保留此次正常休眠的调度计划。停止失败按原路径记录 FAILED 并作异常离线清理；退出观察者按原实例所有权处理，迟到清理保留新 Session。
- Drain、Detach、Close 继续分别服务管理排空、注册撤销与 Core 停机：停用、卸载、更新切代、异常断链/退出时停止未来定时器、清空到期队列并使旧注册回调失效，持久定义的去留沿用原规则。Core 未运行、停用、STOPPED 或 FAILED 期间到期的触发作废；恢复仅安排下一未来 tick，过期动态 at 删除并记 WARN/Expired，未注册任务名的实例保留但不触发。只有正常休眠保留的本次到期经既有 Waker 唤醒；QUIESCING 中到达时遵循原有界等待，等旧树退出、SUSPENDED 与唯一新 Session 的 Ready。
- 静态 tick 到期先占 plugin_id/job_id 的同一实际运行槽，动态到期仍按 plugin_id/name/key 去重并取得 TASK_CONCURRENCY 名额；唤醒等待计入运行槽和动态执行数。同一插件多个到期任务复用 Waker 合并启动，其他 job 仍可并发。同一静态 job 或动态实例已有运行槽时跳过本轮；动态名额已满时按原到期顺序排队，周期实例最多保留一次，排队等待名额沿用原规则。
- 在 Scheduler 锁、逐插件锁与 Core 运行循环之外执行 EnsureActive；等待载荷按实际请求字节占用既有 WAKE_QUEUE_* 与 Core 总预算，取消、期限与满额返回实际错误。静态任务在原 scheduled_at 固定 deadline=scheduled_at+JOB_TIMEOUT；动态任务取得名额后，在线从实际派发起、休眠从开始唤醒等待起固定同一绝对 deadline，排队等待名额保留原语义。期限通过等待 context 与 wire.Request.Deadline 传递，缺省 15 分钟，冷启动消耗原预算，新 Session 保留该截止与原 scheduled_at。
- EnsureActive 返回后在调度锁内重新取得 Attach 发布的当前注册，核对本次计划、已提交代次、路由、本轮取消状态、运行槽归属与剩余期限，Ready 后才向新 Plugin 发送一次既有 scheduled REQUEST。静态暂停停止后续 tick，已占槽的一轮沿用原执行规则。复用 Waker 的启用、代次与摘要检查，保留独立 run_id、job scope、无 Actor/默认 Admin/默认 Target 及动态任务参数；已取消或过期的等待结束本轮，callback 失败记录完整错误，不自动重试。
- 同代次 resume 的 Attach 重新绑定新 Session，保留已接纳的到期项、原 scheduled_at/deadline、稳定运行槽及动态并发计数；未来定时器按同一计划继续。避免普通恢复中的“只取未来 tick/删除过期 at”分支清掉本次正常休眠到期项；已为本轮消费的一次性实例在 Attach 重读状态库时仍保持待交付所有权。静态 interval 起点、暂停状态、动态周期起点与实例沿用 state.Jobs/Tasks，首次激活及新版本间隔变化仍按既有规则处理。
- 动态一次性实例沿用派发前事务删除及失败不重试；创建/替换、取消、缺少任务名、暂停与恢复的既有语义保持。替换实例保留旧在途槽，取消在途实例协作取消，本地结果结束后按实际执行状态释放。Trigger、TriggerTask、SetJobPaused 与插件固定服务继续核对当前 Activated Session，保持既有显式触发和服务准入规则。
- 唤醒等待与实际交付的引用连续交接，复用当前 router.WithActivity/AcquireActivity 的实际调用包装；任务执行期间 References 保持占用，idle.CanStop 与 StopIdle 沿用原复核。实际响应结束时释放；本地超时/取消先返回而请求已交付时，保留实际运行槽、原调用键与引用，迟到 Response 或确认所属旧树退出才释放。唤醒失败或确认尚未交付的取消/期限结束可以归还本轮等待与槽；Released 只清理归属退出 Session 的实际调用，保留等待下一 Session 的到期项及新 Session 的槽。旧响应与重复收尾只释放自己；失败有日志/计数，沿用现有 Runs/Skipped/TimedOut/Late、动态 Expired 与运行/排队统计。

测试（按 docs/goals/README.md「编写测试段」与编码规范「测试提交规则」）：
- 提交：internal/app/scheduler_wake_test.go（//go:build unix）的 TestSchedulerWake，一个真实 Core＋声明 sleep_capable 的已安装插件子进程＋SQLite 的集成测试，复用 idle_test.go 的 startIdleProbe 动态构建宿主与现有任务 API。以可控的 interval 和动态 at 依次覆盖自动空闲退出→自然到期唤醒→新 Session 执行一次→完成后再次休眠；核对原 scheduled_at、静态起点、原 generation/config/DB 与暂停状态保留，一次性实例按原规则消费；通过 Handler 同步点确认执行期间阻止休眠。使用真实定时器和正常 StopIdle/Ready 路径，插件源码与生成文件放测试临时目录，不新增 testdata 插件。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：12:00、16:00、22:00 三条独立 cron 与 interval 的正常休眠唤醒；动态 at、interval/cron、暂停及缺少任务名；同 job/实例重叠跳过、并发名额与到期队列、等待中替换/取消；多个任务同时到期唯一启动；QUIESCING 到期、旧树退出前保持等待、Ready 前交付门禁、resume Attach 保留已接纳项及已消费 at；静态与动态 deadline 起点、冷启动耗尽期限、取消与条数/字节/Core 预算满额；超时后真实迟到响应、断链后旧树退出才释放、Released 与新槽竞争、各类失败保留完整错误；异常离线/停用/Core 重启期间作废且不补跑、过期 at 的 WARN/Expired、固定起点及暂停状态恢复；显式 Trigger/TriggerTask 与固定服务的原准入、always_on 与开发组件原路径。复用 scheduler.Options.Now、contracts/schedule.Trigger.Next 和现有任务套件，不新建时钟框架；sdk/ 与 supervisor/ 的新增验证均留本地。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/scheduler/... ./internal/core/supervisor/... ./internal/core/idle/... ./internal/core/state/... ./internal/app/...；go build ./...。PR 按 TEST-MATRIX §3 写明 V-78 与 V-77 的 schedule 活动分项的测试位置、完整命令、平台、结果及未运行项；正常休眠丢失到期任务、异常离线 tick 被补执行、同任务重叠或唤醒延长原期限时阻止交付。「文档同步」列出由审查者回写的 PLAN/TEST-MATRIX 证据及后继 goal-173 的实际接口。

不做：修改休眠准入、空闲阈值或固定冷却策略；新增唤醒队列、任务持久队列、callback 自动重试或时钟框架；新增协议消息、字段、排空握手或业务 RPC；休眠管理 API/CLI/WebUI、指标扩展与固定机器测量（S-06）；新增依赖或修改状态库 schema；不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-178 · T-04.20 SDK 用法从简

状态：已完成（PR #251），合并提交 `43b05989`。以下保留派发时语句。派发时：前置 T-04.18（PR #242）、T-16.2（PR #183）均已合并。2026-11-18 按 D-180 新增。已按 main `d7c85b63` 核对：`sdk/plugin.go`、`sdk/subplugin.go`、`sdk/registry.go` 的 `Regex(routeID, expr string)`、`Keyword/StartsWith/EndsWith/FullMatch(routeID string, patterns ...string)`、`Raw(routeID string)`、`On(routeID string, eventType EventType)`，`sdk/hooks.go` 的 `PreEvent(routeID string, event EventType)`、`BeforeSend(routeID string)`、`AfterSend(routeID string)`；`sdk/sdk.go` 的 `(*Context).Reply(text string) error`；`sdk/registry.go` 的 `(*RouteBuilder).Page(title string)`；`sdk/bus/bus.go` 的 `(*Subscription[T]).ID(routeID string)` 是 `.ID` 的现成写法；`contracts/validate/register.go` 拒绝空的与重复的 route_id；`examples/plugins/showcase/main.go` 手动登记四条 `kit.css`/`kit.js` 路由；`plugintest/plugintest.go` 的 `Result{Sends []Sent; Outcomes []Outcome; Waiting bool}`。与 goal-162（T-04.19）都改 `sdk/` 的回复路径，后合并者同步 main 后保留双方改动；与 goal-179 无共享文件。

```text
/goal 使用 $alonabot-execute 推进 goal-178（切片 T-04.20：SDK 用法从简）。goal 目录 goal-178，分支 goal-178，worktree .worktrees/goal-178，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按 D-180 去掉作者不关心的参数：文本匹配、Raw、事件路由与 Hook 的注册不再收 route ID 并增加 `.ID`；`*sdk.Context` 实现 `context.Context`；`c.Reply` 可带片段；`.Page(title)` 自动登记套件资源；`plugintest.Result.RequireReply`。仓库内全部调用方同步，不留旧签名。

依据：PLAN T-04 步骤 20；CONTRACTS §9 的「注册从简（D-180）」、§7.3 的页面套件一条、§17 的 RequireReply 一条；DECISIONS D-180；TEST-MATRIX V-148；编码规范「首次发布前不写兼容层」。

范围（只新增或修改这些）：
- sdk/：plugin.go、subplugin.go、registry.go、events.go、hooks.go、sdk.go、http.go，Context 实现 context.Context 的新文件；sdk/page/ 与 sdk/internal/ 中为自动登记套件资源所需的最小改动
- plugintest/plugintest.go（RequireReply）与 plugintest 的既有测试
- 上述签名的全部调用方：agent/、plugins/、adapters/、examples/、internal/app/、internal/core/、cmd/alonabot/ 的源码与既有测试，internal/pluginkit/ 的 init 模板与构建期扫描（如按参数位置读取这些调用）
- docs/plugin-guide.md、docs/development.md 中这些调用的写法
- 上述包的 *_local_test.go

允许的依赖：标准库及仓库已批准依赖，不加依赖，不改 .proto。sdk 根包与 sdk/page 之间不得形成 import 环：套件资源的嵌入位置可移到 sdk/internal 下的新包，`page.Handler()` 的签名与行为保持。

接口（本切片修改或新增，以下签名固定；`*Plugin`、`*SubPlugin`、`*CommandGroup` 三处接收者一致）：
- Regex(expr string) *CommandBuilder；Keyword(patterns ...string)、StartsWith(patterns ...string)、EndsWith(patterns ...string)、FullMatch(patterns ...string) *CommandBuilder；Raw() *CommandBuilder
- On(eventType EventType) *EventBuilder；PreEvent(event EventType) *PreEventBuilder；BeforeSend() *BeforeSendBuilder；AfterSend() *AfterSendBuilder（Hook 只在 *Plugin 与 *SubPlugin 上）
- func (b *CommandBuilder) ID(routeID string) *CommandBuilder；*EventBuilder、*PreEventBuilder、*BeforeSendBuilder、*AfterSendBuilder 同名同形。Handle 之后调用按现有 builder 的顺序错误处理
- func (c *Context) Deadline() (time.Time, bool)；Done() <-chan struct{}；Err() error；Value(key any) any，全部转给 c.Context()
- func (c *Context) Reply(text string, more ...message.Segment) error
- func (r Result) RequireReply(t testing.TB, want string)（plugintest）
- button.On、task.On、bus.Subscribe、service.Serve 与 Command(path...) 的签名不变

行为：
- 缺省 route ID 在收集注册时生成：Regex 为 `regex:` 加表达式原文；Keyword、StartsWith、EndsWith、FullMatch 为 `keyword:`、`startswith:`、`endswith:`、`fullmatch:` 加各模式以 `|` 连接；Raw 为 `raw`；事件路由为事件名（EventType.Name()）；Hook 为 `pre_event:` 加事件名、`before_send`、`after_send`。`.ID` 给出时用给出的值。子插件前缀、命令路由的 route ID 规则不变。
- 同一插件内 route ID 重复（含缺省值之间、缺省值与显式值之间）时，注册收集返回错误，错误文本包含该 ID 并写明用 `.ID` 区分。先核对缺省 ID 能通过 contracts/validate 的现有 route_id 校验，并能经管理 API 的现有路径与编码被读取和设置命令策略；任一不成立时标 `[!]`，不自行改变缺省规则或校验。
- `*sdk.Context` 作为 context 传给 sdk/kv、sdk/task、sdk/service、sdk/llm、sdk/media、sdk/tool、sdk/command、sdk/conversation 时与传 `c.Context()` 行为相同（同一期限、取消与调用关联）；Handler 返回后的失效规则不变。
- `c.Reply(text, more...)` 把 `message.Text(text)` 与 more 依次组成一条 Chain，经 c.Send 发送一次，只返回错误；不带 more 时与原行为相同。
- `.Page(title)` 的路由收集时，SDK 在该页面路径所在目录下登记 `kit.css` 与 `kit.js` 两条 GET 路由（页面 `/page` 对应 `/kit.css`，页面 `/page/` 对应 `/page/kit.css`），内容与 `page.Handler()` 相同，认证与该页面路由相同；作者已登记同一方法与路径时不重复登记、不报错。showcase 删除手动登记的四条路由后页面照常加载套件；若沙箱 iframe 中以页面相同的认证取不到套件资源，标 `[!]` 并写明实测的请求与响应，不自行改为公开路由。
- RequireReply：调用 t.Helper()；任一 Outcome.Err 非 nil 时 t.Fatalf 并写出 PluginID、RouteID 与错误；`len(Sends) != 1` 或 `Sends[0].Text() != want` 时 t.Fatalf 写出实际发送。
- 全部调用方改为新签名。原来写了 route ID、且有测试、命令策略或文档引用该 ID 的路由，用 `.ID(原值)` 保持 ID 不变；其余改用缺省 ID。把示例与官方组件中的 `xxx(c.Context(), ...)` 调用改为直接传 `c`，`c.Context()` 方法保留。plugintest 既有测试与 examples 中自写的单条回复断言改用 RequireReply。

测试（按编码规范「测试提交规则」）：
- 提交：plugintest/plugintest_test.go 新增一个集成测试，被测插件用 `FullMatch("hello")`、`On(sdk.MessageCreated)`（不带 ID）登记，Handler 中 `kv.Set(c, ...)` 后 `c.Reply("hi", message.ImageURL(...))`；断言一次发送、链中两个片段，并用 RequireReply 断言另一条纯文本命令。既有测试只改写调用，不算新增。
- 本地自测（*_local_test.go，结果写入 PR）：V-148 其余分项：各类缺省 ID 的取值、重复 ID 的错误文本、`.ID` 在 Handle 之后、子插件前缀；Context 的期限与取消传播；`.Page` 在 `/page` 与 `/page/` 两种路径下的套件路由、作者已登记时不重复；RequireReply 的四种失败。sdk/ 中只写本地自测。

完成条件：gofmt -l 无输出；go vet ./...；go build ./...；`rg -n 'c\.Context\(\)' examples plugins` 的剩余项逐条说明理由；go test -race ./sdk/... ./plugintest/... ./examples/... ./plugins/... ./agent/... ./internal/pluginkit/... ./internal/app/... ./cmd/alonabot/...。PR 按 TEST-MATRIX §3 写明 V-148 的测试位置、命令、平台、结果与未运行项，并列出保留了显式 `.ID` 的路由及理由。

不做：命令路由的 route ID 规则；`c.Stream`、任务目标保存等其他接口；CLI 与终端对话（goal-179）；AlonaBot/docs 仓库；不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

## 第 21 波

### goal-179 · T-13.7 CLI 开箱即用

状态：已完成（PR #253），合并提交 `8a95d959`。以下保留派发时语句。派发时：前置 T-13.6（PR #233）、T-16.1（PR #168）均已合并。2026-11-18 按 D-181 新增。已按 main `d7c85b63` 核对：`cmd/alonabot/commands.go` 在根命令与 `run` 上登记 `--data-dir`、`--listen`、`--log-level`、`--log-format`、`--reload`、`--stdio`，`RunE` 调用 `runDevelopment(dirs []string, dataDir string, reload, stdio bool, interval time.Duration, options app.Options) error`（`cmd/alonabot/dev.go`）；`app.Options{DataDir, Args, Components []Component, Listen, LogLevel *slog.Level, LogFormat, Stdio, WebUIs}`，同进程组件用 `app.AdapterComponent(a adapter.Adapter) Component`、`app.PluginComponent(root *sdk.Plugin) Component`（`internal/app/assembly.go`）；`sdk/adapter` 的 `Adapter{Descriptor() *Descriptor; Start(ctx, StartRequest, Core) (Instance, error)}`、`Instance{Account(); Capabilities(); Send(ctx, to, chain); Stop(ctx)}`、`Core.Deliver(ctx, ev, cursor)`；`examples/devbot/adapter/adapter.go` 与 `plugintest/adapter.go` 是同类回环 Adapter 的现成写法；`cmd/alonabot/terminal_unix.go`、`terminal_windows.go` 的 `isTerminal(file *os.File) bool`；`internal/pluginkit/init.go` 的 `Init` 在发布版 CLI 下用 `go mod edit -require` 写入 SDK 版本，开发版返回 needsSDK。与 goal-178 无共享文件。

```text
/goal 使用 $alonabot-execute 推进 goal-179（切片 T-13.7：CLI 开箱即用）。goal 目录 goal-179，分支 goal-179，worktree .worktrees/goal-179，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：按 D-181 让作者 `alonabot init` 后直接 `alonabot run`，在同一个终端输入 `/hello` 就看到回复：内置终端 Adapter 与 `--console`，终端对话时 stderr 缺省只输出 WARN 及以上，开发目录缺省重载，发布版 `init` 生成可直接运行的工程。

依据：PLAN T-13 步骤 8；CONTRACTS §15.4 的「终端对话（D-181）」与缺省重载两条、§14.2；DECISIONS D-181；TEST-MATRIX V-149。

范围（只新增或修改这些）：
- cmd/alonabot/：commands.go、dev.go，终端 Adapter 的新文件 console.go，dev_e2e_test.go
- internal/app/：仅为本次运行追加 Adapter 实例与 Core Admin（不写回设置文件）以及 stderr 缺省级别所需的 Options 字段和读取点
- internal/pluginkit/init.go
- docs/development.md、examples/devbot/README.md 的上手步骤
- 上述包的 *_local_test.go

允许的依赖：标准库及仓库已批准依赖（含已使用的 golang.org/x/sys），不加依赖，不改 .proto。终端 Adapter 放在 cmd/alonabot（package main），只导入 sdk/adapter、sdk/message、sdk/errcode 与标准库；不新增跨模块依赖边，plugintest 不被非测试代码导入。

接口（本切片新增，以下固定）：
- `run` 与根命令的 `--console`（bool，缺省自动）：未显式给出时，`isTerminal(os.Stdin)` 为真且 core_config.json 的 adapters 没有任何实例时启用；`--console` 强制启用；`--console=false` 关闭
- `--reload` 的缺省值：有开发目录（含缺省取到的当前目录）时为 true；显式给出时取给出的值
- 终端 Adapter：Descriptor 的 adapter_id 为 `console`，组件 ID `alonabot-adapter-console`，实例名 `console`，账号 ID `bot`；Capabilities 为 message.send、message.edit、message.buttons、message.markdown、message.reference
- app.Options 追加的字段由执行者命名，写进 PR 的「文档同步」；语义限于：本次运行的 Adapter 实例、本次运行的 Core Admin、未显式指定时的 stderr 级别

行为：
- 启用时把终端 Adapter 作为同进程组件追加到 options.Components，并为本次运行提供实例 `console` 与 Core Admin `console:local`；二者都不写入 core_config.json，不出现在下次未启用终端对话的运行中。core_config.json 已有同名 adapter_id 的实例或已安装同 ID 组件时返回用法错误并说明冲突。
- 输入：按行读取 stdin，去掉行尾换行；空行忽略；`#<正整数>` 表示点击最近一条带按钮的已发送消息中第 N 个回调按钮，经与平台 Adapter 相同的按钮点击事件交付，编号无效时在 stdout 写一行英文提示；其他行作为用户 `local` 在私聊会话中的 `message.created`（一个 text 片段，to_me 为真）经 Core.Deliver 交付。stdin 读到 EOF 后停止读取，Core 继续运行直到信号。
- 输出：Send 把每条消息写到 stdout，一条消息一段：text 片段原样，其他片段写成 `[image]`、`[file: 名称]` 这样的方括号占位；回调按钮在消息后逐行写 `#1 标签`，URL 按钮写 `标签 <url>`。发送返回 Accepted 与递增的消息 ID。编辑写成以 `(edited) ` 开头的新段。输出与输入提示符 `> ` 不交错到同一行；stdout 不是终端时不写提示符。
- 终端对话启用且 `--log-level` 未显式给出时，stderr 的阈值为 WARN，日志文件与最近记录照常收全部级别；显式给出时按给出的值。首次启动的管理 token 显示规则不变。
- 重载缺省开启后，`--reload=false` 的行为与原先不带 `--reload` 相同；没有开发目录时不轮询。
- 发布版 CLI 的 init：写入 go.mod 的 SDK 版本后在工程目录运行 `go mod tidy`；失败时保留已生成的工程，返回前输出一行英文说明与需要手动执行的命令，退出码为失败。开发版 CLI 的提示与现状相同。
- docs/development.md 的开发运行一节与 examples/devbot/README.md 改为 init、run、在终端输入命令三步；回环 Adapter 与 curl 注入的说明保留为「不经终端」的做法。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/dev_e2e_test.go 新增一个集成测试，用真实 CLI 在临时目录 init 工程（在仓库 module 内），以 `run --console` 启动并经管道 stdin 写入 `/hello`，在 stdout 读到模板的回复；再核对 core_config.json 中没有写入 console 实例与 admin。复用该文件现有的构建与进程辅助。
- 本地自测（*_local_test.go，结果写入 PR）：V-149 其余分项：自动判定的三种情况（终端且无实例、已有实例、`--console=false`）、按钮编号点击与无效编号、编辑与非文本片段的输出、EOF 后 Core 继续运行、stderr 缺省级别与显式 `--log-level`、日志文件内容、`--reload=false`、修改源码后的自动重载、init 的发布版路径（可用临时 module 代理或记录「未运行」）。

完成条件：gofmt -l 无输出；go vet ./...；go build ./...；go test -race ./cmd/alonabot/... ./internal/app/... ./internal/pluginkit/...。PR 按 TEST-MATRIX §3 写明 V-149 的测试位置、命令、平台、结果与未运行项，并附一段真实终端中 init、run、输入 `/hello` 的原始输出。

不做：终端界面库、分栏与历史滚动；群聊与多用户模拟；媒体的真实读取；把终端 Adapter 做成可安装组件；SDK 注册签名（goal-178）；AlonaBot/docs 仓库；不加依赖，不改 ARCHITECTURE/README/spec，文档同步写进 PR。
```

### goal-173 · S-06.1 休眠管理 API、运行诊断与 CLI

状态：已完成（PR #269），合并提交 `f4bfe5b8`。以下保留派发时语句。派发时：S-02.2（goal-171，PR #249，`18f8af95`）、T-09.14（goal-93，PR #192，`0cff77c`）、S-03.1（goal-172，PR #256，`216bb7f9`）均已合并。2026-11-19 按 main `216bb7f9` 核对设置、策略、状态、管理装配及 Scheduler 的 EnsureActive/Suspend/Jobs/SetJobPaused 接口，与下述签名一致。组件列表的 `Package.RuntimeState` 与 `State=Suspended` 已实现，本切片保留并验收。

无路由快照的查询行为与本分支接入的 ADMIN-API §24 一致：前四类常驻原因可直接判定；需要路由而无已提交快照时返回 412 / FAILED_PRECONDITION，保持读取无唤醒副作用。派发语句与查询契约随本切片一并交付。

```text
/goal 使用 $alonabot-execute 推进 goal-173（切片 S-06.1：休眠管理 API、运行诊断与 CLI）。goal 目录 goal-173，分支 goal-173，worktree .worktrees/goal-173，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：提供 ADMIN-API §24 的六个管理端点与 alonabot sleep-policy，读出全局设置、插件有效策略、运行六态及最近原因；复用现有设置发布和唤醒路径，为 S-06.3（goal-175）提供 API。

依据：PLAN §3.1、§7 S-06 步骤 1 与 S-02/S-03 的失败停止条件；spec/plan/core.md 的 S-06.1；CONTRACTS（spec/CONTRACTS-v0.1.md）§4.3～§4.5、§9.3、§15.1、§16；ADMIN-API §1、§3、§5、§24；DECISIONS D-153、D-164、D-179；TEST-MATRIX V-81 的管理 API/CLI 分项及 V-77/V-78 的受影响回归；ARCHITECTURE 依赖表；docs/execution-order.md 的休眠生命周期及设置共享规则。

范围（只新增或修改这些）：
- internal/core/manage/sleep_api.go（新增）：六个端点、请求/响应、身份筛选、设置编辑与只读运行状态；http.go 的 Handler 路由登记，service.go 的 Options 注入字段。Package、List 与 describe 的既有六态展示保留，仅修正本切片直接暴露的问题。
- internal/app/alonabot.go 的 manage.Options 装配；idle.go 的管理策略只读适配，必要时新增同包 sleep_management.go；复用现有 sleepPolicy、operation 与 applySleep，不更改休眠生命周期。
- cmd/alonabot/sleep.go（新增）：命令、参数到管理 API 的映射及输出；manage.go 的 managementCommands 注册。commands.go 已统一接入 managementCommands，保持原入口。
- cmd/alonabot/sleep_e2e_test.go（新增的一条集成主路径）；manage_e2e_test.go 中现有构建/认证/进程工具的必要复用调整。上述包的 *_local_test.go 保留本地；直接使用现有 settings、idle、state、scheduler API，不改它们的实现。

允许依赖：标准库和现有已批准依赖。manage 复用既有 -> settings/state/contracts，管理端点沿用 adminauth；internal/app -> manage/idle/settings/state/contracts 已由装配边允许，策略结果经 Options 函数注入，manage 不新增 -> idle/supervisor/app 的 import。CLI 继续经 managementClient 请求 HTTP，使用既有 -> manage/settings/state/contracts 及 cobra，不导入根包。测试复用 cmd/alonabot 既有 Core 子进程、包构建、SQLite 与回环服务，不新增 testdata 插件或依赖。

接口（现有，保持签名和原调用可用）：
- settings.Sleep{Mode string; IdleTimeout time.Duration; Plugins map[string]string}；(*settings.File).Settings() settings.Settings；(*settings.File).Update(edit func(doc *settings.Document) error) error；Reload() ([]string, error)；WithPublisher(publish func(change func() error) error) *settings.File。
- (*settings.Document).SetSleep(mode, idleTimeout string) error；SetSleepPolicy(pluginID string, policy *string) error。nil 删除覆盖，字符串只接受 never 或正 Go duration；sleep.mode 只接受 always_on/component_idle，缺省 always_on，idle_timeout 缺省 15m，plugins 缺省空 map。
- idle.Policy{Policy *string; EffectivePolicy string; SleepCapable bool; Eligible bool; Reason string; Warnings []string}；idle.Resolve(s settings.Sleep, m *manifest.Manifest, independent bool, routes []*pb.RouteDescriptor) idle.Policy。原因顺序和全部路由警告由这个唯一实现决定。
- (*state.Store).GetInstall(ctx context.Context, id string) (state.Install, error)；Installs(ctx context.Context) ([]state.Install, error)；Enabled(ctx context.Context, pluginID string) (bool, error)；Active(ctx context.Context, pluginID string) (string, error)；RuntimeStatus(ctx context.Context, pluginID string) (state.RuntimeStatus, error)；ActiveRouteSnapshot(ctx context.Context, pluginID string) (state.RouteSnapshot, error)。RuntimeStatus 的字段为 State state.RuntimeState、Reason string、ChangedAt time.Time、LastWakeReason string、LastWakeAt time.Time、LastSuspendAt time.Time；六态常量为 RuntimeStopped/RuntimeWaking/RuntimeActive/RuntimeQuiescing/RuntimeSuspended/RuntimeFailed，线上值 STOPPED/WAKING/ACTIVE/QUIESCING/SUSPENDED/FAILED。
- manage.Options 的 Store *state.Store、Settings *settings.File、Development map[string]*manifest.Manifest、State func(string) string、Scheduler *scheduler.Scheduler；manage.New(opt manage.Options) *manage.Service；(*manage.Service).Handler(auth *adminauth.Auth) http.Handler；List(ctx context.Context) ([]manage.Package, error)。Package.RuntimeState string 的 JSON 名为 RuntimeState，省略空值；State 字段保留阶段语义。
- app 内部 (*Core).sleepPolicy(ctx context.Context, in state.Install, sleep settings.Sleep) (idle.Policy, state.RouteSnapshot, error)、operation(ctx context.Context, apply func(context.Context) error) error、applySleep(ctx context.Context) error，hasInstalledRoutes(in state.Install) bool。原 manage 装配传入 shared File.WithPublisher 视图；operation 返回前在循环及文件锁、插件锁外应用 sleep 变化并聚合错误。
- S-03.1 提供 scheduler.Options.EnsureActive func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)、(*scheduler.Scheduler).Suspend(pluginID string)、Jobs(pluginID string) []scheduler.StaticJob、SetJobPaused(ctx context.Context, pluginID, jobID string, paused bool) error；已有 (*supervisor.Waker).EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error) 与 (*assembly).ensureActive 装配保持原义。Jobs 可读正常休眠保留计划；显式调度操作仍要求 Activated。本切片不直接调用唤醒或调度方法完成读取。
- CLI 内部 managementCommands(code *int) []*cobra.Command、(*managementOptions).flags(command *cobra.Command)、(*managementOptions).client(required bool) (*managementClient, error)、(*managementClient).request(ctx context.Context, method, path string, body any) (json.RawMessage, error)、managementCall(cmd *cobra.Command, code *int, opt *managementOptions, method, path string, body any) error。复用既有认证发现、URL 组装、错误正文和退出码。

接口（本切片新增，固定以下公开形状）：
- manage.SleepPolicy：Policy *string（json:"policy"，nil 必须输出 null）、EffectivePolicy string（effective_policy）、SleepCapable bool（sleep_capable）、Eligible bool（eligible）、Reason string（reason）、Warnings []string（warnings,omitempty）。这是管理响应视图，字段原义与 idle.Policy 一致。
- manage.Options 追加 SleepPolicy func(ctx context.Context, in state.Install) (SleepPolicy, error)，由 app 的现有 Core 装配注入；参数是管理入口已核对的安装记录，回调只读策略，不启动进程。现有不使用这些端点的 manage.New 调用保持可用，不给旧调用增加初始化前置条件。其他新增 helper 和 HTTP DTO 保持私有，不新增根包或 SDK API。

行为：
- Handler 登记 GET/PUT /api/v1/core/sleep、GET/PUT/DELETE /api/v1/plugins/{id}/sleep-policy、GET /api/v1/plugins/{id}/runtime；使用同一管理 Bearer 或 Cookie/CSRF 保护，复用 decode/respond/failure 及 ADMIN-API §5 的 code/safe_message，成功形状严格按 §24。全局 PUT 两项必填，仅改 mode/idle_timeout，保留 plugins；策略 PUT 的 policy 必填；DELETE 无正文。
- 所有逐插件入口先通过 GetInstall 核对存在、未卸载且 role 为 plugin/agent；其他角色及已卸载 ID 返回 404/NOT_FOUND。未知 mode、auto 存储字面量、零/负/非法时长、缺失必填字段在写入前返回 400/INVALID_ARGUMENT；在 Update 的编辑回调中沿用 SetSleep/SetSleepPolicy 校验，将其输入错误按 invalid(err.Error()) 返回，保留完整说明；文件读取、格式、落盘及发布错误继续按原错误类型传递，不将 I/O 错误归为 INVALID_ARGUMENT。合法时长覆盖允许保存到 manifest 常驻组件，生效仍按 manifest 常驻原因。
- 策略 GET 读取本次生效 settings.Sleep，app 复用 idle.Resolve，输出原始覆盖、有效时长、准入原因和全部 warnings。installed manifest 的解析与身份检查复用既有路径；开发/同进程组件使用装配已有的 Development manifest 且 independent=false，避免访问不存在的安装包目录。always_on、manifest_requires_resident、in_process、policy_never 这些先于路由的原因可由 manifest/进程形式/设置直接决定；需要路由才能判定时读取已提交 ActiveRouteSnapshot。已安装但尚无已提交代次、快照缺失或身份不符时，保留状态库实际 FAILED_PRECONDITION/错误，不能以空 routes 生成 eligible=true；读取全局设置和 runtime 仍可独立使用。查询不触发恢复、候选准备或数据库迁移，不返回快照中的凭据/配置。
- runtime GET 从同一个持久状态来源组合 enabled、generation_id、state、reason、changed_at；时间 UTC RFC 3339。LastWakeAt 非零时输出 last_wake{reason:LastWakeReason,at:LastWakeAt}，LastSuspendAt 非零时输出 last_suspend{reason:"idle",at:LastSuspendAt}，缺失记录省略。保留实际英文原因与失败错误；不从 State 阶段字符串推断六态，不用当前请求时间伪造历史。组件列表沿用已存在的 RuntimeState 与 Suspended 展示。
- 写入只经 manage.Options.Settings 的受管视图 Update 和 Document.SetSleep/SetSleepPolicy；Update 已重读磁盘并保留其他字段、未知键和未修改的 ${...} 原文。提交前失败保留原文件与快照；rename 后目录同步失败仍发布已提交设置并保留真实错误。handler 返回 Update 的聚合结果，不再额外调用 applySleep/EnsureActive。调用者等待恢复时须已释放设置文件锁与逐包锁，保持既有 operation 的串行设置发布和锁外恢复路径，避免在 Handler 中持逐包锁等待唤醒。
- never 立即恢复已启用的 SUSPENDED 插件；QUIESCING 时先完成停止，再按最新设置恢复。always_on 恢复全部已启用休眠插件，停用插件保持停用。变更立即生效、不 reload；保持代次、配置、业务 DB、路由与任务，时长从下一次空闲判断使用。提交后失败保留已发布设置，各失败插件记录实际运行错误，健康插件继续恢复，所有持久化/唤醒错误聚合按 §5 返回。客户端重新 GET 核对，不增加 applied/persisted 等响应字段，不自动重试。
- CLI Use 为 sleep-policy [ID] [never|auto|duration]，只接受 0～2 个位置参数；沿用 --data-dir、--admin-url、--token-file。不带参数时读全局 sleep、组件列表，并按安装顺序筛选未卸载 plugin/agent，读取各自策略和 runtime；输出 JSON 含全局 sleep 与 plugins 数组（每项 id、policy 对象、runtime 对象）。只有 ID 时读策略与 runtime 并输出同样的单项信息；never/正时长以 PUT 写 {"policy":值}，auto 以 DELETE 删除覆盖，成功按既有管理命令输出 {}。各请求失败保留原错误并非零退出，聚合读取全部成功后才输出结果。CLI 无参/单 ID 仅 GET，不唤醒；全局 mode/idle_timeout 写入使用上述 PUT API，本切片不扩展 CLI 语法。

测试（遵循 docs/goals/README.md「编写测试段」及 D-86）：
- 提交：cmd/alonabot/sleep_e2e_test.go（//go:build unix）的 TestSleepPolicyManagement，一条真实 CLI＋认证管理 API＋已安装 sleep_capable 插件子进程＋SQLite 的集成路径。复用 TestManagementLifecycle 的 testPackage/startEmptyCore、core.management、findPackage 与原 HTTP 认证工具；组件若需新增探针源码，写入测试临时目录，不新增 testdata 插件。通过已有 --args 注入短的 SLEEP_MIN_UPTIME 与可观测空闲时长，不改产品缺省值。覆盖默认全局读取→PUT component_idle/时长→启用插件→自动 SUSPENDED→GET 全局/策略/runtime、CLI 列表及单 ID 查询仍保持原 Session 数且无唤醒→CLI never 恢复同代次新 Session→CLI auto 删除覆盖→再次休眠→PUT always_on 恢复。核对最近原因/时间、RuntimeState 与 State、未修改字段、原 generation/config/业务 DB 及迁移次数保持；设置已在线插件的合法时长时 Session 不变，证明没有 reload。集成主路径内验证现有 Cookie/CSRF 认证的一次真实策略写入。保留原 TestManagementLifecycle 的全部断言。
- 本地自测（*_local_test.go，不提交，PR 中逐项写结果）：六端点的无认证/错误 token、Cookie 缺/错 CSRF 与跨 Origin 写入；缺失 mode/idle_timeout/policy、未知 mode、auto 字面量、空/零/负/非法时长；未知/卸载/Adapter/WebUI ID、plugin/agent 一致；manifest 常驻保存时长与 reason、同进程 in_process、原覆盖 null/never/时长、固定原因优先序、全部危险路由 warnings；从未启用及快照缺失的实际错误与 runtime 可读，读取不泄漏凭据且不启动进程。
- 本地自测继续覆盖：保存前失败无发布/无唤醒、提交后目录同步失败仍发布及聚合错误、更新保留 ${...}/未知键/插件覆盖、QUIESCING 期间写入与生命周期串行且锁外恢复、never/always_on 的健康恢复与多插件失败聚合、停用保持、原代次/任务/业务数据保留；CLI 0/1/2/超量参数、auto DELETE、非法输入、错误响应/退出码以及凭据发现，API 请求不自动重试。复用原 settings/idle/app 测试机制，针对本切片新增路径取证，不把已验证底层实现的测试整批复制。
- 既有回归：internal/app 的 TestIdle、TestWake、TestSchedulerWake，以及 cmd/alonabot 的 TestManagementLifecycle、TestManagementInstancesAndCommands；新增 supervisor/sdk 验证只留本地，不提交这些目录的测试。

完成条件：gofmt -l 对改动 Go 文件无输出；git diff --check；go vet ./internal/core/manage/... ./internal/app/... ./cmd/alonabot/...；go test -race -count=1 ./cmd/alonabot -run '^(TestSleepPolicyManagement|TestManagementLifecycle|TestManagementInstancesAndCommands)$'；go test -race -count=1 ./internal/app -run '^(TestIdle|TestWake|TestSchedulerWake)$'；与本切片有关的 *_local_test.go 按测试名定向运行并列全命令；go build ./...。无协议生成变更，不为文档或未改动的生成源重跑生成；全仓 race 按父审查者约定统一串行，按 TEST-MATRIX §3 记录完整命令、平台、结果、提交测试新增行数和未运行项。PR「文档同步」列出 ADMIN-API §24、PLAN/TEST-MATRIX、架构/README/开发指南及后继 goal-175 可用的六端点与字段，审查者在合并前回写。

失败停止条件：未认证写入或跨角色接受、读取导致唤醒/迁移、策略与 idle.Resolve 分叉、缺失快照被当成可休眠、保存前失败仍发布或保存后错误被吞、持锁等待恢复而死锁、停用插件被恢复、策略变更触发 reload/代次变化/重复迁移、丢失已有设置或任务、双进程/重复交付/延长原期限，均阻止合入。发现契约冲突或必须改范围外产品文件、公开签名、proto、依赖/调优初值时按共同约束记录阻塞，不自行扩展。

不做：web/admin 页面（goal-175）、指标/固定机器测量及参数调整（goal-174）；新生命周期、Waker/调度实现、状态表或迁移、协议/SDK/根包 API、自动重试或新 CLI 语法；不改 spec、ARCHITECTURE、README 与 docs/execution-order.md，所需文档同步写入 PR。
```

## 第 22 波

### goal-174 · S-06.2 休眠指标、故障注入与固定机器验收

状态：待派发。前置 S-06.1（goal-173，PR #269，`f4bfe5b8`）已合并。按 main `f4bfe5b8` 复核六个管理端点、`sleep-policy` CLI、设置锁外发布与运行诊断；以下指标、装配、测量与测试约定和当前代码一致。固定 linux/amd64 机器由审查者预约，CI 与采样串行；WebUI 页面由 goal-175 独立接续。

```text
/goal 使用 $alonabot-execute 推进 goal-174（切片 S-06.2：休眠指标、故障注入与固定机器验收）。goal 目录 goal-174，分支 goal-174，worktree .worktrees/goal-174，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：补齐休眠计数与耗时记录，完成真实进程故障注入和固定 linux/amd64 机器上的资源测量，以原始数据复核初值并汇总 V-77、V-78、V-81。

依据：PLAN §7 S-06 全部步骤（实施步骤 2～4）；CONTRACTS §4.4、§4.5、§10、§14、§16；DESIGN §O；spec/adr/SLEEP-DESIGN.md §8、§10（现行行为以已并入契约为准）；TEST-MATRIX V-77、V-78、V-81、V-74 及 §3；ARCHITECTURE 的依赖方向；D-86、D-171（rss 后置沿用 goal-48 的首发范围决定）。

前置与范围：
- 开始实现前确认 S-06.1 已合并，并逐项核对下列管理接口；有签名或数据形状差异时交审查者修订本语句。运行状态、策略与设置写回沿用同一 Core 所有者。
- 产品改动限于 internal/core/metrics/counters.go 的种类/原因说明，internal/core/supervisor/{wake,activity,supervisor,prepare,runtime}.go 的计数、冷启动与状态变化日志，sdk/session.go 与 sdk/activate.go 的回调耗时日志，internal/app/assembly.go 的计数装配，以及 examples/plugins/{echo,configdb}/plugin.yaml 的 sleep_capable 声明。沿用 Idle Controller 与 Scheduler；发现行为故障时先报告具体契约、位置及受影响范围。
- 验证入口限于新增 internal/app/sleep_metrics_test.go（一个提交的集成测试）、scripts/sleep-measure/main.go（仅 linux 构建的标准库测量工具）与各相关包被忽略的 *_local_test.go；原始数据写入用户指定的临时输出目录。明确授权回写 spec/TEST-MATRIX.md §3 的证据，及在审查者先完成契约同步并批准新值后修改 internal/tune/params.go。其他规格、架构与页面同步交审查者。
- 固定机器为 linux/amd64，优先与 V-33/V-74 使用同一台机器；可用 instabox，由协调审查者预约独占时段并与 CI 串行。执行者取得时段后才采样。按 D-171 与 goal-48，本轮完整测 Core、echo、configdb 的 gormdb 变体；rss 的 V-74/V-81 样本分项注明随组件进入后续版本测量及原因，作为已裁定后置项保留。help 如有补测，保留 help 名称和独立结果。
- 复用既有 metrics、slog、管理 API、CLI、真实子进程与 SQLite，无新增依赖。测量工具只负责本次两组插件与 Core 基线取证；两组插件冷启动各 100 次。rss 后置独立登记，本轮现有样本的测量完整执行。

接口与装配（现有签名保留，新增公开面仅一项）：
- 现有 metrics.Counters.Add(pluginID string, kind metrics.Kind, reason metrics.Reason) 与 Snapshot() map[string]map[metrics.Kind]map[metrics.Reason]uint64 继续由 assembly 的唯一 *metrics.Counters 持有；零值、并发使用和快照复制语义保留。新增 kind 值 "wake"，原因取下文固定集合，完整错误只进日志。
- 在 supervisor.WakeOptions 新增字段 Count func(pluginID string, reason string)。assembly 创建唯一 Waker 时传入闭包 a.metrics.Add(pluginID, metrics.Kind("wake"), metrics.Reason(reason))；Count 为 nil 时关闭该处计数。supervisor 经函数注入调用，沿用 ARCHITECTURE 已允许的 app→metrics、app→supervisor 边。
- supervisor.NewWaker(opt WakeOptions) *Waker、EnsureActive(ctx context.Context, pluginID string, payloadBytes int64) (func(), error)、Ready(running *Running)、Activity(pluginID string) ActivityStatus、AcknowledgeColdStarts(pluginID string, through uint64) 均沿用现有签名。计数独立于冷却历史的消费与回收，Snapshot 及诊断读取保持只读。计数回调在活动门禁锁外调用，仅同步累加内存计数。
- idle.New(values tune.Values) *Controller 与 CanStop(pluginID string, policy Policy, runtime state.RuntimeStatus, activity supervisor.ActivityStatus, now time.Time) bool 继续由 Core 循环串行调用；scheduler.Options.EnsureActive func(ctx context.Context, pluginID string, payloadBytes int64) (release func(), err error)、Suspend(pluginID string)、Stats(pluginID, jobID string) JobStats 及 TasksStats(pluginID string) TaskStats 沿用现有生命周期与统计所有者。测量读取快照，实际空闲退出仍经 StopIdle；全部绑定完成后的 Ready 门禁继续决定交付时刻。
- S-06.1 候选提供 manage.Options.SleepPolicy func(ctx context.Context, in state.Install) (manage.SleepPolicy, error)，由 Core.managementSleepPolicy(ctx context.Context, in state.Install, development *manifest.Manifest) (manage.SleepPolicy, error) 装配。六个端点为 GET/PUT /api/v1/core/sleep、GET/PUT/DELETE /api/v1/plugins/{id}/sleep-policy、GET /api/v1/plugins/{id}/runtime；读取保持无唤醒副作用。全局 PUT 同时传 mode 与 idle_timeout，插件 PUT 传 policy，DELETE 删除覆盖。响应沿用 ADMIN-API §24 的 envelope 与 §5 错误；runtime 含 enabled、generation_id、state、reason、changed_at、可选 last_wake/last_suspend。CLI 沿用 alonabot sleep-policy [ID] [never|auto|duration]；全局模式经管理 API 修改。设置更新继续走 settings.File.Update 与既有发布者，失败保留实际错误并重新 GET 核对已保存设置。
- 测量入口：go run ./scripts/sleep-measure --alonabot <CLI绝对路径> --echo-package <echo归档> --configdb-package <configdb_gorm归档> --work-dir <新的临时目录> --out <原始数据目录> --phase <cold|resources|core> [--round <1|2|3>]。cold 执行两组各 100 次冷启动；resources、core 通过 --round 分别执行指定的一轮插件资源测量或 Core 基线。各次运行独立取证，按批次与轮次人工汇总原始结果，工具仅增加本次运行的阶段/轮次选择。工具仅 import 标准库，所有函数私有；输入缺失、平台不符、安装/启动/读数失败均返回完整错误和非零退出码，并保留已取得的数据。复用 CLI 的 install/run/restart 与认证管理 API，使用独立 data_dir；Core 与插件使用普通发布构建，编译、安装、首次迁移与测试工具自身耗时均在冷启动样本区间之外。

指标与日志口径：
- "wake"/"started"：每次 Waker 合并后实际调用 Start 一次记一次；"ready"：该次 Start 返回且完整绑定/Ready 已核对成功；"failed"：该次共享 Start 失败一次。100 个等待者共享一次启动时三者按启动次数统计。"cold_start"：process.Start 成功后的每个真实独立进程记一次，包括初次启用、管理重启与按需恢复；process.Start 本身失败保留错误，唤醒尝试计入 failed。
- 等待拒绝原因分别为 "dropped_queue"（逐插件条数）、"dropped_bytes"（逐插件字节）、"dropped_budget"（Core 总 payload）。已入队后未交付的请求每个只记一个终止原因："cancelled"、"deadline"、"dropped_failed"（观测到共享启动失败）或 "dropped_other"（其他实际错误）；排队循环与唤醒广播均不重复计数。等待取消/到期立即归还原预算，共享启动继续。区分每次启动的 failed 与每个等待者的丢弃数，Scheduler 到期唤醒复用同一入口。
- 每次计数同步写 DEBUG "wake metric"，含 plugin_id、reason；集成测试通过同一时间范围的日志聚合并与 Snapshot 对照。测量工具在独立 core_config.json 显式设置 log.level=debug、log.format=json、log.file.enabled=true，保留原日志与设置，插件来源继承 debug。错误日志保留完整 err 及可取得的 phase/code、plugin_id、generation_id、session_id、pid，按 CONTRACTS §14 脱敏。运行状态、唤醒与空闲停止原因沿用已有持久记录，并为真实状态变化补齐 INFO 日志；禁止用任意错误文本作为计数标签。
- 冷启动起点为 supervisor 调用 process.Start 前的既有 startedAt；终点为 host.Activate 成功返回、Core 已确认合法 ACTIVATED 的第一时间，先记录再写运行状态与装配业务绑定。用同一进程的单调时间求差；通过 Prepared 私有字段跨准备/激活阶段保存起点。成功记录 DEBUG "cold start completed"，含 duration_ns、plugin_id、generation_id、session_id、pid、activation_reason；另外记录 Core 可直接观察的 host.Prepare 调用区间 prepare_ns 与 host.Activate 调用区间 activate_ns。失败记录实际错误与已耗时，成功分布只包含完整成功样本。Ready 继续表示全部绑定完成，两类观测分别保留。
- 计时触点依据：当前 REGISTER 覆盖配置、迁移和 OnStart，ACTIVATE→ACTIVATED 覆盖 OnReady 及协议往返；control.proto/route.proto 没有回调耗时字段，Core 的上述可见阶段测量分别保留原名。CONTRACTS §4.5 要求单独记录 OnStart/OnReady，因此本语句将 sdk/session.go:prepare 的 root.onStart 和 sdk/activate.go:runOnReady 的 a.onReady 两个调用点的日志纳入最小范围，复用现行 componentLogger 输出英文消息与结构化字段。保持回调预算、watchdog、panic/error 处理及激活门禁；协议字段、wire、根包与 SDK 公开 API 保持现状。
- OnStart、OnReady 在 SDK 实际执行用户回调的 goroutine 内分别计时，起点为调用前，终点为实际返回或 panic。DEBUG "lifecycle callback completed" 含 phase（OnStart/OnReady）、executed=true、duration_ns、outcome（success/error/panic）及完整错误；缺少回调时记 executed=false、outcome=not_executed，耗时字段省略并在汇总标无样本。watchdog 超时时仅记录 "lifecycle callback deadline observed"、outcome=timeout 及 observed_ns 截止观测，保留原失败日志；实际返回前的 duration_ns 保持缺项，进程退出同样保留缺项原因。额外字段 lifecycle_session_id 取 HELLO_ACK 的 Session ID 十六进制，用于关联准备期日志（现有 Source.SetSession 在 host.Prepare 返回后调用）；Core 的来源、代次与 session_id 字段仍由原 Source 补充，完整冷启动耗时由 Core 直接测得。

行为与故障注入：
- 唤醒中崩溃：用临时真实插件在 OnReady 同步点退出，多个等待者收到实际启动错误，运行状态 FAILED、共享 failed 只增一次，未交付等待分别归还条数/字节/Core 预算，旧树确认退出后按既有恢复规则处理。无重复请求、双进程、重复迁移或额外业务重试；Scheduler 失败撤销计划沿用 S-03.1。
- STOPPING 中到达请求：用真实 OnStop 同步点冻结旧树，命令与自然到期项按原 deadline 等待；旧树退出后仅恢复一个原代次 Session，Ready 后各交付一次。取消的请求退出等待，过期项保持原期限；验证 StopReason=IDLE、一次 STOPPING，正常空闲退出继续保持异常重启次数。
- 切回 always_on：经认证管理 API 修改模式，逐个等待全部已启用且休眠的 plugin/agent 按原代次 ACTIVE；停用插件保持 STOPPED。核对 ActiveGeneration、配置、路由快照内容、静态起点/暂停状态、动态任务参数与业务 DB 标记保持，迁移次数保持，运行状态与最近原因可读。设置保存后的唤醒失败保留完整错误及 FAILED，GET 返回实际已保存值。
- 频繁唤醒：核对超过 SLEEP_COLD_START_MAX 后从触发时刻起固定常驻 SLEEP_COLD_START_WINDOW，期间新启动保留计数且冷却截止保持；结束后恢复空闲判定。既有测试与已合并证据有效时直接引用，指标读取与冷启动历史回收互不消耗。

固定机器测量与数据定义：
- 冷启动两组为 echo 与 configdb 的 configdb_gorm 构建，每组固定 100 次真实新进程冷启动尝试，序号 1～100 各执行一次。先在独立临时 data_dir 完成安装/首次迁移，再于 always_on 模式按原代次串行管理 restart，明确记录 activation_reason=resume、文件缓存条件与运行顺序；以 cold start completed 为区间，API/CLI 完成时间另存。该批记录进程冷启动，真实 SUSPENDED→WAKING→Ready 主路径由集成与故障注入另行取证。失败编号逐条保留、报告实际尝试数与失败率，继续独立样本取证；任一组少于 100 个完整成功样本时保留失败/未完成状态，修复后重测须对应新候选并保留原批次。rss 分项按 D-171 后置。
- 原始行至少含机器标识、产品 head、样本组/序号、代次/Session/pid、起止时间、cold_start_ns、prepare_ns、activate_ns、OnStart/OnReady 的实际耗时或缺项原因、结果与完整错误。成功样本按耗时升序，Pq 取第 ceil(q×N) 项，N 为有效成功样本数；完整批次 N=100，P50/P95/P99 分别第 50/95/99 项，存在缺项的分布标为部分结果。统一以 ms 展示并保留 ns 原值。echo P95 200ms 为初步目标；超出时定位阶段与差异并报告，其他插件只记录实测值。
- 资源测量允许按独立完整轮次跨多个正常运行窗口分批完成；全部批次固定同一 linux/amd64 机器、同一产品候选、同一构建包与采样方法，每个样本组、每轮从新的独立 Core/data_dir 开始，避免前 100 次启动触发的冷却期污染空闲测量。原始数据记录批次、轮次、实际运行环境与起止时间，逐轮保留日志并在最后汇总；每轮在预约窗口内完成采样及既有停止/清理流程，CI 与采样互斥。记录原样 sleep 与全部 tune 生效值；使用 CONTRACTS 缺省预算、15 分钟空闲时长、5 分钟最短运行时间、冷启动阈值与冷却窗口，managed runtime 的运行限制沿用现状，预约时段包含这些实际等待。通过管理 API 开启 component_idle 并按真实空闲条件等待自动 SUSPENDED。每组 ACTIVE 和确认 SUSPENDED 后分别稳定空闲 2 分钟，完整保留 3 轮及两组各 100 次冷启动要求；采集 /proc/<pid>/smaps_rollup 的 Pss、Rss、Private_Clean+Private_Dirty（KiB 转 MiB），同时记录各 PID 原值、插件进程树汇总和 Core/必要附属进程的单列与合计。仅 Core 的常驻基线按 V-74 单独测量。
- CPU 从同一 2 分钟区间内 /proc/<pid>/stat 的 utime+stime 差值，结合机器 getconf CLK_TCK 与实际单调墙钟差计算，100% 表示占满一个逻辑核；按进程树汇总并保留 ticks、频率与时间原值。确认旧插件整树已退出后，该树休眠期间的 PSS/RSS/CPU 记录为 0，并保留 Core 与附属进程实测值；读取权限、缺少文件或进程身份变化造成的缺项写明原因。测量平台为 linux/amd64，PSS 缺失时该项保持未运行。
- 保存 OS/arch、内核、CPU 型号/逻辑核数、总内存、Go 版本、产品/样本源码版本与完整构建命令（含 configdb_gorm 标签、CGO/构建参数）、包来源、环境限制、全部采样命令、逐行原始数据与汇总路径。测量机或本轮样本输入不足时记录具体缺口；实际机器与 V-33/V-74 的对应关系须写明。按 V-74 记录 ACTIVE 三轮中位数及原预算 Core ≤24 MiB、echo ≤10 MiB、configdb gormdb ≤16 MiB；rss ≤14 MiB 随组件后置，help 补测保持独立名称与结果，预算口径按各自样本登记。
- 初值沿用 CONTRACTS §16；测量支持调整时，在 PR 给出旧值、建议值、对应数据与影响，由审查者先同步 CONTRACTS/PLAN/TEST-MATRIX/ADR，再批准 internal/tune 的落值。保留原预算及 echo 200ms 初步目标，完整记录所有失败与测量缺项。

测试：
- 提交：internal/app/sleep_metrics_test.go:TestSleepMetrics，一个真实 Core＋认证管理 API＋已安装独立插件进程＋SQLite 的集成主路径，复用 wake/idle/scheduler 现有夹具。覆盖自动休眠→100 个并发请求共享恢复→再次休眠→always_on 恢复；核对每次实际启动及成功计数、逐条交付次数、冷启动与 SDK 阶段日志可关联，原代次/快照/任务/业务标记/迁移次数保持。测试使用显式记录的短时长与 tune 覆盖；真实固定机器测量沿用上述缺省值。测试插件源码在测试中动态构建，遵守 D-86。
- 本地自测：三种预算拒绝、取消/deadline、唤醒失败的逐启动/逐等待计数，快照并发与历史回收、无 OnStart/OnReady、回调错误/panic/超时、STOPPING 中请求与到期项、OnReady 崩溃、设置保存后唤醒失败、固定冷却期；按上述真实故障场景验证。SDK、supervisor 及 testdata 绑定测试一律 *_local_test.go/临时构建，PR 写明命令与结果。测量工具用本地测试校验 /proc 解析、进程身份、KiB/CPU 单位、nearest-rank 及输入/读数失败保留，禁止以 fixture 代替真实资源数据。
- 回归复用 TestWake、TestIdle、TestSchedulerWake 与合并后 S-06.1 的管理集成测试；已有有效 V-77/V-78 分项逐项引用源码位置、PR/head、命令/平台/结果，仅补本切片改变或仍缺的证据。

完成条件与停止条件：
- 开发环境执行 gofmt -w <本切片改动的Go文件>、git diff --check、go vet ./internal/core/metrics/... ./internal/core/supervisor/... ./internal/core/idle/... ./internal/core/scheduler/... ./sdk/... ./internal/app/...、go build ./...；非 Linux 的 go build ./... 按 Go 既有构建约束与包模式规则跳过全部源码仅适用于 Linux 的目录。运行 go test -race -count=1 ./internal/core/metrics/... ./internal/core/supervisor/... ./internal/core/idle/... ./internal/core/scheduler/... ./sdk/... ./internal/app/...；S-06.1 管理 API/CLI 测试按合并后的实际测试名定点回归。仅在 linux/amd64 测量机执行 go build ./scripts/sleep-measure 与 go test -count=1 ./scripts/sleep-measure，后者运行本地解析/计算自测。race 与采样分时执行，采样使用普通构建。
- 在预约的 linux/amd64 机器执行上述测量命令，保留两组各 100 次冷启动、ACTIVE/SUSPENDED 的 PSS/RSS/CPU、Core 基线、全部失败与丢弃原因及故障注入结果。在 TEST-MATRIX §3 回写测试位置、命令、平台/版本、产品与包来源、原始数据/日志路径和通过/失败/未运行，汇总 V-77、V-78、V-81 的每个分项；rss 单列 D-171 后置原因，留给随后续版本的组件验收。
- 出现双进程、重复交付、期限延长、预算未归还、空闲退出计入异常重启、代次/迁移/数据改变、快照混入凭证、Scheduler 丢失休眠计划或补执行异常离线 tick 时停止并报告；修复涉及本语句范围外的文件、协议或契约时交审查者决定。需要新增依赖、放宽预算/阈值或更换测量样本时同样停止对应工作。
- 测量机、本轮 echo/configdb 包或任何必需读数缺失时，受影响分项保持未运行；指标与独立故障注入继续。本轮必需测量缺项时 goal-174 及 V-81 总体验收保持未完成；rss 按 D-171 登记的后置状态随对应版本继续。该切片没有 WebUI 发布前置；页面证据归 S-06.3 / goal-175，关联 T-09.8/T-09.9 的真实浏览器操作，未合并/未运行明确留项，V-81 汇总逐项呈现状态。
- 最后一个 task 重跑本切片适用完成命令，推送 goal-174 并按模板开 PR 到 main，写明提交测试新增行数、本地测试、实际数据及待审查者同步内容。存在必需缺项时保留阻塞与未完成状态；由审查者验收后更新切片状态。
```
