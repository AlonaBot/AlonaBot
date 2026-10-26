# 切片 goal 语句

本文件收录给执行者派发切片用的 goal 语句。一个 goal 对应 [PLAN §3.1](spec/PLAN.md#31-执行切片) 中的一个切片。并行派发的时机、共享文件和合并顺序按 [执行顺序](docs/execution-order.md)；执行者的工作方式按 [执行技能](.agents/skills/alonabot-execute/SKILL.md)。本文件只由审查者维护。

## 用法

1. 在「总览」中找到状态为**可派发**的 goal。
2. 把该 goal 的整段语句（代码块内的全部文字）原样发给一个执行者。一个 goal 同时只交给一个执行者。
3. 执行者开出 PR 后，把 PR 链接交给审查者审查。
4. 审查者合并后，会更新依赖它的 goal 的状态和接口段。

状态的含义：

- **可派发**：前置都已合并，语句已写到接口签名，可以直接派发。
- **等待前置**：语句已写完整，等前置合并、审查者核对接口后转为可派发。
- **待编写**：范围与要点已由 PLAN 确定，没有待决的设计问题；前置合并后，审查者按实际代码写出接口、行为和测试段，然后才能派发。

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
| goal-23 | T-07.1 | 调度触发计算 | — | 可派发 |
| goal-29 | T-08.2 | 配置 JSON 语义与 Pending | — | 可派发 |
| goal-37 | T-09.2 | 安装包预检与安全解压 | — | 可派发 |
| goal-2 | T-03.1 | 首个可用 SDK 切片 | T-02.1、T-02.2 | 等待前置 |
| goal-3 | T-03.2 | Target 与发信许可 | T-03.1 | 等待前置 |
| goal-6 | T-03.5 | 媒体上传与配额 | T-03.1 | 等待前置 |
| goal-65 | T-02.1 | 日志基线与日志文件 | — | 待编写 |
| goal-76 | T-02.2 | 调优参数 | — | 待编写 |
| goal-4 | T-03.3 | Adapter 进程、sdk/adapter 与上下线事件 | T-03.1 | 待编写 |
| goal-5 | T-03.4 | 官方 Telegram Adapter 的接收 | T-03.3 | 待编写 |
| goal-7 | T-03.6 | 入站媒体与 Grant | T-03.2、T-03.4、T-03.5 | 待编写 |
| goal-50 | T-03.7 | 官方 Telegram Adapter 的发送计划与回执 | T-03.4 | 待编写 |
| goal-8 | T-03.8 | 按钮、extension 片段与依赖解析 | T-03.2、T-03.7 | 待编写 |
| goal-51 | T-03.9 | 消息编辑 | T-03.7 | 待编写 |
| goal-9 | T-03.10 | Telegram 接入开发入口与故障矩阵 | T-03.6、T-03.8、T-03.9 | 待编写 |
| goal-10 | T-04.1 | 根插件 API 与静态注册索引 | T-03.1 | 待编写 |
| goal-11 | T-04.2 | 构建期生成器 | T-04.1 | 待编写 |
| goal-12 | T-04.3 | 命令选择、文本匹配与参数绑定 | T-04.1 | 待编写 |
| goal-13 | T-04.4 | 持久化命令策略 | T-03.2、T-04.1 | 待编写 |
| goal-14 | T-04.5 | 示例升级与 T-04 验收 | T-03.10、T-04.2、T-04.3、T-04.4 | 待编写 |
| goal-15 | T-05.1 | 可信上下文与有界分发 | T-03.3、T-04.3、T-04.4 | 待编写 |
| goal-16 | T-05.2 | Rule、HandlerOutcome 与传播 | T-05.1 | 待编写 |
| goal-17 | T-05.3 | 三阶段 Hook 与 send_forbidden | T-05.2 | 待编写 |
| goal-52 | T-05.4 | 延续调用与多轮等待 | T-05.2 | 待编写 |
| goal-18 | T-05.5 | 多进程事件流水线验收 | T-03.8、T-05.3、T-05.4 | 待编写 |
| goal-19 | T-06.1 | Service/Bus 描述与 Broker 准入 | T-04.1 | 待编写 |
| goal-20 | T-06.2 | 父调用链、取消与背压 | T-03.6、T-05.3、T-06.1 | 待编写 |
| goal-21 | T-06.3 | 非持久 Event Bus | T-06.1 | 待编写 |
| goal-22 | T-06.4 | Session 清理与 Service 端到端 | T-06.2、T-06.3 | 待编写 |
| goal-24 | T-07.2 | schedule 注册与运行槽 | T-05.1、T-07.1 | 待编写 |
| goal-25 | T-07.3 | SDK HTTP Listener 与 Core 代理 | T-04.1 | 待编写 |
| goal-26 | T-07.4 | 两个 Host 与插件票据 | T-07.3 | 待编写 |
| goal-27 | T-07.5 | 调度与 HTTP 启停集成及验收 | T-07.2、T-07.4 | 待编写 |
| goal-28 | T-08.1 | 配置字段描述与构建检查 | T-04.2 | 待编写 |
| goal-30 | T-08.3 | SDK 配置准备顺序 | T-08.1、T-08.2 | 待编写 |
| goal-31 | T-08.4 | 不可变 Secret Store | T-08.1、T-08.2 | 待编写 |
| goal-32 | T-08.5 | GORM 连接与迁移 | T-08.3 | 待编写 |
| goal-33 | T-08.6 | SQLite 候选与资源交接 | T-08.4、T-08.5 | 待编写 |
| goal-34 | T-08.7 | web/admin 骨架、配置表单组件与界面文案 | T-08.1 | 待编写 |
| goal-53 | T-08.8 | 用户可见文案目录 | T-04.1、T-04.3、T-08.1 | 待编写 |
| goal-35 | T-08.9 | configdb 示例与 T-08 验收 | T-08.6、T-08.7、T-08.8 | 待编写 |
| goal-36 | T-09.1 | 首次管理认证 | T-07.4 | 待编写 |
| goal-38 | T-09.3 | 开发 CLI 与库模式脚手架 | T-04.5、T-08.1、T-08.8、T-09.2、T-13.1 | 待编写 |
| goal-39 | T-09.4 | 管理服务、安装记录与 serve | T-03.8、T-08.4、T-09.1、T-09.2、T-13.1 | 待编写 |
| goal-40 | T-09.5 | 候选 Journal 与准备编排 | T-09.4、T-08.6、T-06.4 | 待编写 |
| goal-41 | T-09.6 | 唯一提交点与激活门禁 | T-07.5、T-09.5 | 待编写 |
| goal-42 | T-09.7 | 崩溃恢复、回退与回收 | T-09.6 | 待编写 |
| goal-43 | T-09.8 | Admin 管理页 | T-09.6、T-08.7 | 待编写 |
| goal-44 | T-09.9 | Admin 配置、权限、工具策略与插件页 | T-09.8、T-11.1、T-11.3 | 待编写 |
| goal-66 | T-09.10 | 日志管理与日志页 | T-02.1、T-03.1、T-09.8 | 待编写 |
| goal-45 | T-09.11 | 正式流程故障注入与 T-09 验收 | T-09.3、T-09.7、T-09.9、T-09.10 | 待编写 |
| goal-46 | T-10.1 | 证据汇总与跨组件联验 | T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.5 | 待编写 |
| goal-47 | T-10.2 | 五平台构建与运行 | T-10.1 | 待编写 |
| goal-48 | T-10.3 | 资源比较报告 | T-10.1 | 待编写 |
| goal-49 | T-10.4 | 发布资产与说明 | T-10.2、T-10.3 | 待编写 |
| goal-54 | T-11.1 | 工具描述、目录与策略 | T-04.4、T-06.1 | 待编写 |
| goal-55 | T-11.2 | 工具调用与用户确认 | T-03.8、T-06.2、T-08.8、T-11.1 | 待编写 |
| goal-56 | T-11.3 | Agent 会话存储 | T-03.1 | 待编写 |
| goal-57 | T-11.4 | 工具与会话的验收 | T-11.2、T-11.3 | 待编写 |
| goal-58 | T-12.1 | Provider 客户端 | — | 待编写 |
| goal-59 | T-12.2 | Agent 对话与流式显示 | T-03.9、T-04.3、T-05.4、T-08.3、T-08.4、T-11.3、T-12.1 | 待编写 |
| goal-60 | T-12.3 | 工具循环与 MCP | T-11.2、T-12.2 | 待编写 |
| goal-61 | T-12.4 | 记忆插件 | T-08.5、T-11.2、T-12.2 | 待编写 |
| goal-62 | T-12.5 | 技能 | T-12.3 | 待编写 |
| goal-63 | T-12.6 | 定时 Agent 任务 | T-07.2、T-08.5、T-12.3 | 待编写 |
| goal-64 | T-12.7 | 官方插件打包与 T-12 验收 | T-12.4、T-12.5、T-12.6 | 待编写 |
| goal-67 | T-13.1 | 库模式入口与同进程组件 | T-02.1、T-03.3、T-04.2 | 待编写 |
| goal-68 | T-13.2 | 开发组件的配置与数据库 | T-08.3、T-08.5、T-13.1 | 待编写 |
| goal-69 | T-13.3 | 热部署 | T-13.1 | 待编写 |
| goal-70 | T-13.4 | 库模式示例与验收 | T-03.10、T-13.2、T-13.3 | 待编写 |
| goal-71 | T-14.1 | help 插件与命令列表 | T-04.4、T-05.2、T-08.8 | 待编写 |
| goal-72 | T-14.2 | welcome 与 rss 插件 | T-05.2、T-07.2、T-08.3、T-08.5、T-08.8 | 待编写 |
| goal-73 | T-14.3 | 发行归档 | T-03.10、T-09.3、T-12.7、T-14.1、T-14.2 | 待编写 |
| goal-74 | T-14.4 | 首次运行与预装 | T-09.6、T-14.3 | 待编写 |
| goal-75 | T-14.5 | 发行包端到端验收 | T-09.11、T-14.4 | 待编写 |

## 第 1 批：可派发

### goal-23 · T-07.1 调度触发计算

状态：可派发。

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

状态：可派发。

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

状态：可派发。

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

## 第 2 批：等待前置

goal-2 等待日志基线与调优参数（T-02.1、T-02.2）；其余语句按 T-03.1 的接口编写。前置合并后，审查者核对实际代码，必要时修订，再改为可派发。

### goal-2 · T-03.1 首个可用 SDK 切片

状态：等待前置（T-02.1、T-02.2）。日志基线与调优参数合并后，审查者按实际的 Logger 与参数接口补充本语句，再改为可派发。

```text
/goal 使用 $alonabot-execute 推进 goal-2（切片 T-03.1：首个可用 SDK 切片）。goal 目录 goal-2，分支 goal-2，worktree .worktrees/goal-2，从最新 origin/main 创建。开始前完整读取 goals.md「所有 goal 的共同约束」。

目标：插件能注册 message.created 事件路由；Core 把标准事件经真实 Session 派发给已 Activated 的插件；插件在 Handler 里用 Context.Send 回复原事件 Target；Core 经 Adapter 接口发出，并把 SendResult（以及可能同时存在的错误）返回给插件。

依据：PLAN T-03 步骤 1；CONTRACTS §1.2、§8；MESSAGE-MODEL §1、§2、§3（本切片只做手动发送签名）、§4；WIRE-FORMAT §1、§4、§5A。

范围（只新增或修改这些；.proto、internal/core/state、cmd/alonabot 和所有文档都不改）：
- 新增 internal/wire/services.go
- 新增 internal/core/adapter/ 和 internal/core/adapter/memory/
- 新增 internal/core/router/ 和 internal/core/message/
- 修改 internal/core/host/activate.go：Instance 增加 Routes 字段，在 Activate 中从 prep.Register 赋值
- sdk/：新增 events.go、dispatch.go；修改 sdk.go（Context、Send、Registration、Collect）、activate.go（业务 Handler 与当前 Session）、session.go（buildRegister 输出事件路由）
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
   func New() *Router
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
   type EventProvider interface { Events(*EventRegistry) }
   type EventRegistry struct{ ... }
   func (r *EventRegistry) On(routeID, eventType string, handler func(*Context) error)
   type EventRoute struct { RouteID string; EventType string; Handler func(*Context) error }
   Registration 增加 Events []EventRoute；Collect 收集 EventProvider。
   func (c *Context) Event() *pb.Event
   Context.Send 由返回 UNSUPPORTED 改为真实实现，签名不变。
   func currentSession() (*wire.Session, error) // 未导出：返回已激活的 Session；未就绪或已停止时返回 Code=FAILED_PRECONDITION 的 *sdk.Error。后续的后台 Send、媒体客户端都通过它取 Session，不再改 activate.go

行为：
- Router.Deliver 放入容量 256 的队列，满时立即返回 adapter.ErrQueueFull。容量和单 worker 是临时实现，用一行 shortcut 注释写明：有界并发派发接入后替换。
- worker 逐条处理事件：
  - 候选是 Activated()==true 的 Plugin 上 kind=EVENT 且 event_type_filter 包含该事件类型的路由，按 (priority, plugin_id, route_id) 升序依次调用。
  - 每次调用发送 Request{Target: route_id, Payload: PROTOBUF 编码的 HandlerRequest（schema_id 用 validate.SchemaID，version 1）, TraceID: 16 字节随机值, Attach: &Invocation{PluginID, Event, DefaultTarget: ev.Target}}。HandlerRequest.context 只设置 actor 与 default_target。
  - 响应解码为 HandlerOutcome：propagation=STOP 时停止后续候选；调用错误或带 Err 的响应记日志后继续下一个候选。不重试。
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
  - 执行 Handler 时 recover panic，转为 INTERNAL。Handler 返回后 Context 失效，之后再调用 Send 返回 FAILED_PRECONDITION。
  - 成功时返回 HandlerOutcome{matched: true, propagation: 调用过 StopPropagation 为 STOP，否则 CONTINUE}。
  - Handler 返回 *sdk.Error 时，Err 使用它的 Code 与 Message；其他错误为 INTERNAL，safe_message 为 "handler returned an error"，原文只写 stderr。出错时仍返回 HandlerOutcome，保留 STOP。
- Context.Send：构造 SendRequest{to: To 选项, chain: chain.Proto()}，以本次 Inbound.CallID 为 ParentCallID 调用 wire.ServiceSend。有载荷就解码成 SendResult；有 Err 时返回 *sdk.Error（Code 经 validate.PublicCode）。结果与错误可以同时返回。
- EventRegistry.On 的以下情况都记为 Collect 错误：routeID 为空；与已注册的事件或命令 route_id 重复；eventType 不是 message.proto 中列出的 8 种之一；handler 为 nil。
- buildRegister 为每个事件生成 RouteDescriptor{route_id, kind=EVENT, default_enabled=true, required_permission=ANY, event: EventSpec{event_type_filter: [eventType], handler_mode=MANUAL}}，排在命令路由之后。
- examples/plugins/echo：根插件注册 On("echo", "message.created", ...)，把收到消息中的 text 片段原样回复；没有 text 片段时不回复。

测试（按编码规范「测试提交规则」：只提交集成测试；本地自测写在 *_local_test.go，不提交，PR 中写明结果）：
- 提交：internal/core/supervisor/echo_test.go（//go:build unix）。按 supervisor_test.go 构建示例的方式构建 examples/plugins/echo，用 memory Adapter + adapter.Registry + message.Service + Mux + Router 装配，Launch 后 Attach 实例。
  - Inject 一条 message.created：memory 收到一条 Chain，文本与原文相同，Target 等于原 Target。
  - Inject 一条没有匹配路由的事件类型：memory 没有收到任何发送。
  - TCP 与 stdio 各跑一次。
- 本地自测：message.Service 的各条拒绝路径（无目标、非原目标、空 Chain、未知实例）和「结果与错误同时返回」；Router 的排序与 STOP、跳过未激活的 Plugin、第 257 条返回 ErrQueueFull、Mux 未知服务；SDK 的 EventRegistry 拒绝情况、buildRegister 输出、Handler 返回后 Send 失败、STOP 与错误并存、panic 转 INTERNAL。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/core/... ./internal/wire/... ./sdk/... ./examples/...；go build ./...。

不做：Telegram、显式目标许可、后台 Send、媒体、按钮、命令分发、Rule、Hook、并发派发、cmd/alonabot 装配。
```

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

## 第 3 批及以后：待编写

以下 goal 只写了范围和要点。前置合并后，审查者按实际代码补全「接口」「行为」「测试」三段并改为可派发；在此之前不要派发。每条的完整语义以 PLAN §4 对应步骤及其失败停止条件为准，不另作设计。

### goal-65 · T-02.1 日志基线与日志文件

状态：待编写。由审查者实现，维护者批准后开始。范围：Core 与 SDK 现有的直接输出、`control.proto` 的 Bootstrap（追加 `log_level`）、`cmd/alonabot` 的 `--log-level` 与 `--log-format`。要点：
- 控制台格式按 CONTRACTS §14.2：`[时:分:秒.毫秒] [Core|Plug:<id>|Adpt:<id>/<实例>] [DBUG|INFO|WARN|ERRO] [目录/文件:行号]: 消息 key=value`，trace 显示前 8 位，多行值缩进 4 个空格，只在终端且未设置 `NO_COLOR` 时着色；另有 JSON 输出。
- SDK 以 JSON 行写 stderr，Handler 的 Logger 附带 trace_id、route_id、call_id；调用边界捕获的 panic 记为带 `stack` 的 ERROR。
- Core 逐行解析插件 stderr，补上并保护来源字段；非 JSON 行作为 WARN 原样保留。
- SDK 与 Core 都按来源的生效级别过滤；持久化与运行时修改留给 T-09.10。
- 日志文件按 CONTRACTS §14.6：本地日期轮转，单文件 100 MiB 后同日续写编号文件，合计 1 GiB、保留 30 天，写入失败时继续运行并报告一次。
- 只用标准库 `log/slog`，不加依赖；日志不含 Secret、Bootstrap 与配置明文。
- 验收：V-62、V-63 的级别过滤分项、V-68 的写入与轮转分项。

### goal-76 · T-02.2 调优参数

状态：待编写。由审查者实现，维护者批准后开始。范围：新建 `internal/tune`，迁入 `internal/wire` 与 `internal/core` 中对应的现有常量；`cmd/alonabot` 的 `--args` 与 `tune list`；`control.proto` 的 Bootstrap（追加 `session_limits`）。要点：
- 参数表与缺省值按 CONTRACTS §16；不可调的协议、安全与平台常量留在原处。
- `--args` 可重复：值含 `=` 时是逗号分隔的直接参数，否则是文本文件路径；后出现的覆盖先出现的；大小只接受二进制单位。
- 未知键、格式错误、违反约束或相互矛盾时拒绝启动并指明出处；有覆盖时启动 WARN 列出全部覆盖项。
- SDK 使用 Bootstrap 下发的 `session_limits`；设置 `PLUGIN_MEMORY_SOFT_LIMIT` 时为子进程传入 `GOMEMLIMIT`。
- 验收：V-69。

### goal-4 · T-03.3 Adapter 进程、sdk/adapter 与上下线事件

状态：待编写（T-03.1）。范围：`contracts/proto` 的 `adapter.proto` 与生成代码（本切片获准修改 `.proto`）、manifest 的 `role`/`adapter_id` 解析、`internal/core/adapterhost/`（以 Adapter 进程为后端实现 T-03.1 的 `adapter.Instance`）、`internal/core/state` 的游标表、`sdk/adapter/`、`examples/adapters/loopback/`。要点：
- WIRE §3.4 的七个固定服务全部打通；Adapter 与插件的 REGISTER 按安装记录的角色互斥校验，角色不采信进程自报。
- 实例启停：`adapter.instance.start` 下发实例配置与已保存的游标，按返回的 account_id 检查同一 `adapter_id` 下的账号唯一；启停一个实例不重启进程，也不影响同包的其他实例。
- 交付：Core 先校验事件归属（adapter_id、adapter_instance_id、account_id 与本 Session 已启动的实例一致），再写入游标、接纳事件，最后应答；伪造实例的事件被拒并记诊断。event_id 由 Core 生成。
- 进程异常退出后按插件的有限重启策略恢复，只重新启动仍处于启用状态的实例，并从已保存的游标继续。
- 实例经 `adapter.instance.start` 成功启动后生成 `bot.online`；因停用、Adapter 进程退出或 `core.adapter.status` 的致命错误离开运行状态时生成带 reason 的 `bot.offline`（MESSAGE-MODEL §1.1）。两类事件不带 Actor 与 Target，Core 重启不补发。
- 回环测试 Adapter 用于本切片和 T-05 的测试，不依赖任何真实平台。
- 验收：V-48 的进程与实例分项、V-51。

### goal-5 · T-03.4 官方 Telegram Adapter 的接收

状态：待编写（T-03.3）。范围：`adapters/telegram/`（只经 `sdk/adapter` 与 Core 通信）与 `adapters/telegram/telegramtest/` 的本地 Bot API 模拟服务。要点：
- `adapter.instance.start` 中依次执行 getMe（account_id 为 bot 用户 ID 的十进制字符串）、getWebhookInfo（url 非空则失败，不调用 deleteWebhook），然后启动接收循环。
- long polling 的 allowed_updates 固定为 message、edited_message、channel_post、edited_channel_post、callback_query、my_chat_member、chat_member、message_reaction，timeout=30。
- 一批 update 逐条经 `core.adapter.deliver` 交付，游标为 update_id+1；收到应答后才处理下一条，未接纳的事件只计数。不支持的 update 也推进游标。
- 距上次接纳或轮次开始满 7 天时，先经 `core.adapter.cursor` 保存不带 offset 的新轮次，再拉取。
- 401/409 经 `core.adapter.status` 报告并停止该实例；网络错误与 5xx 只对接收退避重试，不用于发送。
- 标准化规则沿用 MESSAGE-MODEL §1、§6：带 sender_chat 的消息不构造 Actor，保留 media_group_id，每条 Telegram Message 一个事件；Actor.locale 取 from.language_code，按 BCP 47 规范化大小写；`to_me` 按 MESSAGE-MODEL §6.1 计算。
- Adapter 实现放在可导入的包中（`telegram.New()`），`main` 只调用 `sdk/adapter` 的运行函数，供库模式直接引用（CONTRACTS §15.3）。
- 验收：V-45、V-54 的 locale 分项。

### goal-50 · T-03.7 官方 Telegram Adapter 的发送计划与回执

状态：待编写（T-03.4）。范围：`adapters/telegram/` 的 `adapter.send` 实现与模拟服务的发送方法。要点：
- 发送前完成整条 Chain 的预检，生成有序的物理操作计划。文本、实体按 UTF-16 偏移；连续媒体按每组 2～10 项组成相册；Reference 只作用于首条，Buttons 作用于末条，相册后的按钮用 markup 更新补上。
- 超长、超限或无法表达的组合在首个请求前返回 UNSUPPORTED 或 INVALID_ARGUMENT，不截断、不降级、不重排。
- PartResult 区分 ACCEPTED、REJECTED、UNKNOWN、NOT_ATTEMPTED，逻辑状态按 MESSAGE-MODEL §4 汇总。遇到拒绝或结果未知即停止后续操作，保留已确认的 ID。message_id=0 不填入 ID，依赖它的步骤记为 NOT_ATTEMPTED。不自动重发。
- 验收：V-08、V-35。

### goal-7 · T-03.6 入站媒体与 Grant

状态：待编写（T-03.2、T-03.4、T-03.5）。范围：`internal/core/media`、`adapters/telegram/` 的 `adapter.media.read`、`sdk` 的 Read 客户端。要点：
- IncomingMedia token 只映射可信 Adapter 给出的平台文件标识。首次 Read 时经 `adapter.media.read` 按需取件，校验后缓存，多个 Session 共享同一份缓存。
- 出站媒体：Core 在 `adapter.send` 之前为本次发送签发只读 token，Adapter 经 `core.media.read` 读取。
- offset 与 length 的边界按 CONTRACTS §1.4。
- Grant 不扩大对象、操作、Target 或有效期。消费租约保护正在使用的对象。
- Bot Token 和宿主路径不进入插件进程。
- 验收：V-44、V-09。

### goal-8 · T-03.8 按钮、extension 片段与依赖解析

状态：待编写（T-03.2、T-03.7）。范围：Core 状态库的按钮 token 表、Telegram callback_query 处理、button.clicked 路由派发、extension 片段的声明与发送前校验、manifest 的 `adapter_dependencies` 与 `plugin_dependencies` 解析及开发入口的依赖检查。要点：
- 按钮 token 为 32 字节随机值的无 padding base64url，持久化绑定 plugin、route、generation、Target 和 24 小时期限。
- 先尽力调用 answerCallbackQuery，再逐次鉴权，Actor 取本次点击的用户。拒绝时不回落到 Raw/Regex，不发群消息。代次切换后 token 永久失效。
- extension 片段按目标实例所属 Adapter 在 REGISTER 中声明的 name/version/Schema 校验；目标不属于该 Adapter 或版本未声明时返回 UNSUPPORTED，不降级。
- 必需的 Adapter 依赖缺失时拒绝激活插件并列出缺失项；Adapter 停用后，发往它的 Send 返回 UNAVAILABLE。
- `plugin_dependencies`：plugin_id、可选 `min_version`、是否必需；必需依赖未安装、未启用或版本不满足时拒绝启用，可选依赖缺失只记诊断，必需依赖成环在解析时拒绝（CONTRACTS §8）。
- 验收：V-08 的按钮分项、V-48 的依赖分项、V-52 的解析分项。

### goal-51 · T-03.9 消息编辑

状态：待编写（T-03.7）。范围：`core.edit` 固定服务、`internal/core/adapter` 的可选接口 `Editor`、`adapter.edit`（获准修改 `adapter.proto` 与出站 Hook 载荷，只追加字段）、`adapters/telegram/` 的 editMessageText、SDK 的 `Edit`。要点：
- Target 准入与 Send 相同；Edit Chain 只能包含一个 text 片段和至多一个 buttons 片段，否则 INVALID_ARGUMENT。
- 每次 Edit 执行一组 BeforeSend → Adapter.Edit → AfterSend，Hook 输入标明 edit 与被编辑的 platform_message_id。
- 未声明 `message.edit` 的实例返回 UNSUPPORTED；平台拒绝（内容未变化、消息不可编辑）为 REJECTED，不改为发送新消息。
- SendResult 只有一项，status 只取 ACCEPTED/REJECTED/UNKNOWN（MESSAGE-MODEL §4）。
- 验收：V-53。

### goal-9 · T-03.10 Telegram 接入开发入口与故障矩阵

状态：待编写（T-03.6、T-03.8、T-03.9）。范围：`cmd/alonabot` 的开发装配（`alonabot run` 同时启动 Adapter 与插件；实例配置取自 `<data_dir>/adapters/<adapter_id>/instances/<instance_id>.json`，CONTRACTS §15.3）、`examples/plugins/echo` 与 media 示例、跨包故障矩阵测试、TEST-MATRIX 证据。要点：
- 完成 PLAN T-03 步骤 8 的全部场景，包括实例上下线事件与消息编辑；用请求记录证明没有自动补发。
- 真实 Telegram 没有测试账号时记为未运行。

### goal-10 至 goal-14 · T-04.1～T-04.5

状态：待编写。要点按 PLAN T-04 各步骤：
- 根插件与显式绑定；生成器只扫描当前模块的目标构建包。
- 重复 route 一律拒绝；跨插件遮蔽稳定，与启动顺序无关。
- 已识别的命令被拒绝后不回落，也不进入文本匹配。
- keyword、startswith、endswith、fullmatch 与 Regex 同属文本匹配层，统一按 priority/plugin_id/route_id 排序；空模式拒绝注册，ignore_case 按 Unicode 简单大小写折叠。
- 策略持久化，在下一次派发时生效；父级禁用优先于子级启用。

### goal-15 至 goal-18、goal-52 · T-05.1～T-05.5

状态：待编写。要点按 PLAN T-05 各步骤：
- 并发链与等待队列遵守 64 条、256 条/8 MiB 的上限。
- Rule 返回 false 时不执行 Handler；Stop 不会被错误清除；自动回复恰好发送一次。
- Hook 不能修改可信绑定，BeforeSend 修改后重新预检，Hook 中禁止 Send。
- goal-52（T-05.4）：延续调用保留原 Actor 与默认 Target，期限 15 分钟，不占分发链，每 Session 最多 128 个进行中；延续调用内再请求、无 Actor 或 Hook 调用链中的请求返回 FAILED_PRECONDITION。`core.wait` 的键为（adapter_instance_id、chat_id、thread_id、external_user_id），同键 ALREADY_EXISTS；命中的消息在 PreEvent 之后、命令匹配之前被等待消费；不持久化。获准在 HandlerRequest 上追加 `continuation_of` 字段（原调用的 call_id）。

### goal-19 至 goal-22 · T-06.1～T-06.4

状态：待编写。要点按 PLAN T-06 各步骤：
- 编码、版本与授权错误在执行目标函数之前拒绝。
- 循环返回 CALL_CYCLE，深度超过 16 返回 FAILED_PRECONDITION。
- Cancel 不越过它的 Request。
- Event Bus 每订阅者 256 条/4 MiB，满时 drop-newest。
- Hook 经 Service 再 Send 的路径被拒绝。

### goal-24 至 goal-27 · T-07.2～T-07.5

状态：待编写。要点按 PLAN T-07 各步骤：
- 同一 job 不重叠，离线错过的 tick 不补跑；运行槽要等 callback 真正返回或进程确认退出才释放。
- HTTP Listener 绑定 127.0.0.1:0，准备阶段不放行业务请求，旧 Session 不再代理。
- 两个 Host 的凭证隔离；所有停止共享同一个停止预算。

### goal-28、goal-30 至 goal-35、goal-53 · T-08.1、T-08.3～T-08.9

状态：待编写。要点按 PLAN T-08 各步骤：
- OnLoad 先于 Schema 校验；默认值只补缺失的键。
- secret 引用不可变，明文不进入日志和响应。
- 迁移的 Up 与账本写在同一事务；PG 需要维护许可。
- 候选 DB 写坏不影响旧数据。
- 前端只使用已批准的 Vite、TypeScript、React 19、vitest；界面文案提供 en-US 与 zh-CN 两份目录。
- goal-53（T-08.8）：`i18/*.json` 的 `messages` 检查（en-US 必须包含全部键，其他语言无多余键，占位符集合一致）；SDK 的 `T` 按 Actor locale → Core 默认 locale → en-US 选择，主语言回落按 CONTRACTS §5.4；Core 的 en-US/zh-CN 用户可见文案目录，usage 改为经目录生成。

### goal-36、goal-38 至 goal-45、goal-66 · T-09.1、T-09.3～T-09.11

状态：待编写。要点按 PLAN T-09 各步骤：
- 管理 token 只保存 SHA256；login 带 CSRF 校验。
- 所有管理动作经唯一的管理服务，按插件串行。
- 提交点是同一事务内的 ActiveGeneration+COMMITTED；崩溃后未提交恢复旧版、已提交恢复新版。
- 旧进程退出未确认时不启动新的写入者。
- 启用、停用、卸载与更新时检查插件依赖；影响依赖方的操作列出受影响插件并要求确认。
- Admin 页面包含工具调用策略、Agent 会话的列出与清除，以及界面语言切换。
- goal-38（T-09.3）的 `plugin init` 生成插件包与库模式调试用的 `main.go`，`plugin build` 生成调用 `sdk.Run` 的正式入口。
- goal-39（T-09.4）的 `alonabot serve` 复用 T-13 的 Core 装配，按安装记录启动全部已启用的包。
- goal-66（T-09.10）：按来源的日志级别持久化，经 `sdk.log.set_level` 立即生效；每个来源保留最近 1000 条记录，提供读取与 SSE，过去日期从日志文件按日期查询；`alonabot logs`、`log-level` 与 WebUI 日志页使用同一 API，页面筛选只在浏览器中进行。

### goal-54 至 goal-57 · T-11.1～T-11.4

状态：待编写。要点按 PLAN T-11 各步骤：
- 只有 JSON 编码的 Service 可以声明 `tool`；策略为 allow/confirm/deny，未设置时取 default_policy，修改在下一次调用生效。
- `core.tool.invoke` 与 `core.tool.confirm` 必须带 parent_call_id；没有 Actor 时 confirm 直接 PERMISSION_DENIED。
- 确认消息由 Core 直接经 Adapter 发送，不执行插件出站 Hook；只接受同一 Actor 在 5 分钟内的第一次点击；不支持按钮时 UNSUPPORTED。
- 会话存储按（plugin_id、会话键）隔离，追加原子、分页受帧上限约束、不截断；卸载默认保留，purge 删除。

### goal-58 至 goal-64 · T-12.1～T-12.7

状态：待编写。要点按 PLAN T-12 各步骤。goal-58（T-12.1）不依赖 Core 的新能力，是这一组里最早可以细化并派发的：
- 范围在 `plugins/agent/` 与 `plugins/memory/`，只 import `sdk/...` 与 `contracts`；`sdk/memory` 提供记忆服务的类型与 Schema。
- Provider 只用标准库实现 SSE；MCP 客户端用已批准的 `github.com/modelcontextprotocol/go-sdk`，goal-60 获准把它加入 `go.mod`。
- Provider 失败不自动重发；流式显示先发送再按间隔编辑，最后一次编辑为完整内容。
- 工具与 MCP 操作都经策略与确认；单轮工具步数上限与编辑间隔是命名常量。
- 记忆的字符上限为命名常量（2,200 与 1,375 码点），超限返回错误不丢已有条目；后台定时任务不注入记忆，confirm 工具一律拒绝。

### goal-67 至 goal-70 · T-13.1～T-13.4

状态：待编写。要点按 PLAN T-13 各步骤：
- 根包 `alonabot` 的 `Run` 与 `alonabot serve` 共用同一份 Core 装配；`internal/core`、`sdk`、`adapters/`、`plugins/` 都不能 import 根包。
- 开发中的组件在 goroutine 中经 loopback TCP 走完整协议，不另建协议路径；组件身份取自生成代码嵌入的 manifest。
- dev 来源使用固定代次 `dev`，配置直接读取 `config.json`，数据库为 `dev.sqlite`，不能 update 或 rollback。
- 热部署用 `Reload` 选项或 `ALONABOT_RELOAD=1` 启用，不解析程序的命令行参数；500ms 轮询，编译失败时旧组件继续运行，其他 module 的组件不重启。

### goal-71 至 goal-75 · T-14.1～T-14.5

状态：待编写。要点按 PLAN T-14 各步骤：
- `core.command.list` 只返回调用者可用的命令，说明按 Actor locale 翻译。
- rss 抓取只允许 http/https，解析出的地址与每次重定向后的地址都不能是 loopback、私网、链路本地、组播或未指定地址；响应大小与超时为命名常量。
- 预装包照常验签；只在数据目录首次初始化时安装，默认启用 Agent、记忆与 help；签名私钥不进入仓库与日志。

### goal-46 至 goal-49 · T-10.1～T-10.4

状态：待编写。要点按 PLAN T-10 各步骤。需要真实的 Telegram、五个目标平台和固定的测量机器，缺少的部分记为未运行。
