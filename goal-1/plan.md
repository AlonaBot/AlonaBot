# goal-1 计划：B1 = T-01 契约与工具链 + T-02 Core 状态、Native 进程与双 IPC

## 1. 目标与完成定义

按 [spec/PLAN.md](../spec/PLAN.md) 完成 T-01、T-02，每个工作包都满足其「停止和失败条件」：

- **T-01 完成：** `.proto` 可以编译，生成结果稳定（重复运行 `go generate ./...` 后 `git diff` 为空）；Envelope 字段号、枚举值与 WIRE-FORMAT 逐项一致；manifest、release、消息结构有校验入口和 fixture；SDK 最小公共类型和示例可以编译；pure-Go SQLite/GORM 的 CRUD、Tx、WAL Backup 通过，且 Up 与账本在同一事务内提交；PG 分支在临时容器上通过；五个目标平台 `CGO_ENABLED=0` 交叉编译通过；依赖版本写入 go.mod/go.sum。
- **T-02 完成：** 最小 Native 插件能通过 TCP 和 stdio 两种传输走完 Hello → Bootstrap → REGISTER → PREPARED → 提交 → ACTIVATE → ACTIVATED → STOPPING → 退出；core.lock、runtime.lock、调用表、准入预算、取消顺序、阶段预算、停止预算和有限重启都有测试；V-03/V-04/V-05/V-06/V-21/V-36/V-37/V-42 在 T-02 范围内的分项证据记入 tasks.md。
- 不属于 B1 的能力（Telegram、命令、配置迁移的实际实现、正式 Journal、HTTP、Scheduler）不实现，只留接入位置。

## 2. 已确定的决策（来自 input.md）

| 项 | 决定 |
| --- | --- |
| Module path | `github.com/AlonaBot/alonabot`（组织需由用户在网页上注册；GitHub 没有创建组织的 API，已验证返回 404） |
| Go 依赖 | 只使用已批准清单：protobuf（protoc-gen-go 通过 `tool` 指令固定）、gorm、libtnb/sqlite、modernc.org/sqlite、gorm postgres 驱动、golang.org/x/sys、go.yaml.in/yaml/v3 |
| 前端 | Vite + TypeScript + React 19 + vitest（B1 不涉及） |
| Git | 工作树 `.worktrees/goal-1`、分支 `goal-1`，每个 task 一个 commit；不 push，不合并到 main |
| PG | 允许使用 docker `postgres:17` 临时容器，名称 `alonabot-test-pg`，只绑定 127.0.0.1，用完即删；测试通过 `ALONABOT_TEST_PG_DSN` 启用 |
| 范围 | 只做 B1 |

## 3. 上下文

- 工具：Go 1.27.1（asdf），protoc 36.0（`/opt/homebrew/bin/protoc`，已安装，不重新安装），Docker 29.4.2。GOPROXY=goproxy.cn；直连失败时可临时使用代理 `127.0.0.1:7890`。
- 本机是 darwin/arm64，只能运行这一个平台。
- 架构与 import 边：[ARCHITECTURE.md](../ARCHITECTURE.md)。共享协议运行时放在 `internal/wire`，sdk 和 internal/core 都 import 它；contracts 只放类型、校验和 fixture。
- 文档、注释和 commit 统一用简体中文。

## 4. 默认假设（goal 期间不能提问，以下由我决定并记录）

1. **Proto 布局：** `contracts/proto/alonabot/v1/*.proto`，go_package 为 `github.com/AlonaBot/alonabot/contracts/pb/alonabotv1`；生成入口是 `contracts/generate.go` 中的 `//go:generate`，调用 `protoc` 和 `go tool protoc-gen-go`，使用 `paths=source_relative` 输出到 `contracts/pb/alonabotv1/`。生成脚本会核对 `protoc --version` 等于 36.0，不一致时失败（这是规格要求的生成可复现性检查）。
2. **WIRE 中没有给出字段号的消息**（控制帧 payload、Bootstrap、RouteDescriptor、Invocation、HandlerOutcome、SendResult、媒体、BusEvent、Event/Segment）由 T-01 在 .proto 中固定，字段号按规格表格中的字段顺序从 1 开始编号。冻结后的编号写回 WIRE-FORMAT 的「T-01 已固定」小节，并在 ADR 变更记录中加一行。
3. **SDK 最小类型（T-01 第 4 步）只证明签名能组合**：放在 `sdk`（Base、Context、Handler 适配类型、注册器、生命周期接口）、`sdk/message`（Chain/Segment 构造），以及 `sdk/config` 和 `sdk/gormdb` 的接口；尚未实现运行行为的方法返回明确的 `UNSUPPORTED` 类错误，不 panic，不伪造成功。
4. **数据库验证代码放在 `sdk/gormdb` 的包测试里**，后续由 T-08 扩展。Core 状态库直接使用 modernc 驱动（`database/sql`），放在 `internal/core/state`。
5. **进程测试** 使用 Go 测试二进制重入（`os.Args[0]` + 环境变量切换为辅助进程）和 `examples/plugins/minimal` 两种方式；不另建长期运行的测试服务。
6. **只写 T-02 需要的 Core 状态表**：installs、generations（含 COMMITTED 标记）、plugin_state（启用状态、失败窗口）、policy 占位。其余表由 T-08/T-09 用迁移追加。状态库 schema 使用内建版本号顺序迁移。
7. **Windows Job Object、LockFileEx 的代码** 用 build tag 编写，通过交叉编译验证；运行状态在证据中记为「未运行」。
8. **cmd/alonabot 在 B1 只提供开发用启停入口**（例如 `alonabot run --data-dir ... --plugin <dir>`），实际参数在 G1-19 中确定并写入 docs/development.md；正式管理 CLI 属于 T-09。
9. 依赖选择当时已发布的最新稳定版本，验证通过后锁定；如果某个版本验证失败，先定位具体缺口，可以换到同一库的其他版本，但不引入新的库。
10. Claude Code 没有 goal 登记工具，也不会自动推进下一轮；进度只以本目录的三份文件为准。用户可以用 `/loop` 驱动后续轮次。

## 5. 风险

| 风险 | 影响 | 应对 |
| --- | --- | --- |
| libtnb/sqlite 与 modernc 的版本组合不兼容，或 Backup API 无法通过 GORM 底层连接取得 | T-01 数据库门槛不通过 | 在 G1-08 中用 `database/sql` 的 `Conn.Raw` 获取 modernc 连接做 Backup；仍然失败就按 PLAN 记录具体缺口并标为阻塞 |
| 下载 docker 镜像或 Go 模块时网络失败 | G1-01/G1-08/G1-09 卡住 | 改用代理 7890 重试一次；仍然失败就把 task 标为阻塞，继续推进不依赖它的 task |
| 准入预算、优先通道、取消顺序的并发逻辑容易出现竞态 | T-02 的核心正确性 | 所有并发测试都加 `-race` 运行；复查 task 重点检查 |
| 进程树清理和锁的行为依赖平台 | 只能在 darwin 上运行验证 | Linux/Windows 的代码只做交叉编译，运行证据标为「未运行」交给 T-10 |
| 阶段超时测试拖慢测试套件（10s/60s 等预算） | 测试变慢 | 预算常量作为包级变量，测试里注入缩短的值；生产值不变，并有测试核对它们等于契约数值 |
| 规格表达不出某个字段 | 契约缺口 | 按 spec/README：先修契约和 fixture，在 DECISIONS.md 记录原因，再写代码 |

## 6. 执行方式

- 每轮：全量重读三份文件 → 取第一个未完成的 task → 实现 → 自检（基于 diff、测试输出等事实）→ 在工作树中 `git commit` → 回写 tasks.md（与代码放在同一个 commit 里，或紧跟着追加一个 commit）→ 简短汇报后停止。
- 每 3 个 task 后安排一次复查 task；跨模块的大改动用 `code-review-expert` 审查当前分支的 diff。
- 验证命令（在 `.worktrees/goal-1` 下执行）：`go generate ./... && git diff --exit-code`、`go vet ./...`、`go test -race ./...`、`go build ./...`，以及五平台交叉编译：
  `for p in linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64; do GOOS=${p%/*} GOARCH=${p#*/} CGO_ENABLED=0 go build ./... || echo FAIL $p; done`
- 证据格式按 TEST-MATRIX §3：测试位置、命令、平台与版本、结果、对应 commit。

## 7. 回滚方案

- 所有改动都在 `goal-1` 分支上，main 不受影响。丢弃整批：`git worktree remove .worktrees/goal-1 && git branch -D goal-1`（只在用户要求时执行）。
- 单个 task 有问题：在分支上追加修复 commit；不使用 reset --hard，也不 force。
- PG 容器带 `--rm`，执行 `docker stop alonabot-test-pg` 即可完全清理；测试数据只放在 `t.TempDir()` 中。

## 8. 红线（在 goal-mode 原有红线之外）

不 push，不合并 main，不修改用户全局配置，不安装运行时，不添加清单外依赖，不删除 experiments，不使用 sleep 或猜 PID 来证明进程已退出（CONTRACTS §12.2）。
