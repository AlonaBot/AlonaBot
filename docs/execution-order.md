# 并行执行顺序

本文件约束多个执行者并行推进切片时的派发时机、共享文件和合并顺序。切片的内容与前置以各章节实施计划的切片表为准（章节划分见 [PLAN §3.1](../spec/PLAN.md#31-执行切片)），派发用的 goal 语句在 [docs/goals/](goals/README.md)，角色与合并授权见 [CONTRIBUTING「角色与合并」](../CONTRIBUTING.md#角色与合并)。

## 1. 派发规则

- 一个 goal 同时只交给一个执行者。
- 只派发 docs/goals/ 中状态为「待派发」的 goal。待派发要求切片表中它的全部前置都是「已完成」（前置可能在其他章节），而且审查者已经按合并后的代码补全了接口段。
- 执行者在派发时从最新的 `origin/main` 创建 `goal-N` 分支和 `.worktrees/goal-N`。goal 编号固定，由 goal 语句指定，不按 goal-mode 的默认规则自选。
- 某个前置 PR 已开但还没合并时，不派发依赖它的 goal，即使代码看起来已经够用。

## 2. 波次参考

下表按前置推出最早可以同时进行的批次，方便估计并行度。实际派发以第 1 节为准：一个 goal 的前置全部合并就可以派发，不必等同批的其他 goal 完成。

| 波次 | 切片 |
| --- | --- |
| 1 | T-02.1、T-02.3、T-07.1、T-08.2、T-09.2、T-12.1 |
| 2 | T-02.2 |
| 3 | T-03.1 |
| 4 | T-03.2、T-03.3、T-03.5、T-04.1、T-11.3 |
| 5 | T-03.4、T-04.2、T-04.3、T-06.1、T-07.3 |
| 6 | T-03.6、T-03.7、T-06.3、T-07.4、T-08.1、T-13.1 |
| 7 | T-03.8、T-03.9、T-04.6、T-08.3、T-08.4、T-08.10、T-09.1、T-13.3 |
| 8 | T-03.10、T-03.11、T-04.4、T-04.7、T-08.5、T-08.7、T-08.8、T-09.4、T-15.1 |
| 9 | T-04.5、T-05.1、T-08.6、T-09.12、T-11.1、T-13.2、T-15.2 |
| 10 | T-05.2、T-07.2、T-08.9、T-09.3、T-13.4、T-15.3 |
| 11 | T-05.3、T-05.4、T-07.5、T-14.1、T-14.2、T-16.1 |
| 12 | T-05.5、T-06.2、T-12.2、T-16.2 |
| 13 | T-06.4、T-11.2、T-12.8 |
| 14 | T-09.5、T-11.4、T-12.3、T-12.4 |
| 15 | T-09.6、T-12.5、T-12.6 |
| 16 | T-09.7、T-09.8、T-12.7 |
| 17 | T-09.9、T-09.10、T-14.3 |
| 18 | T-09.11、T-14.4、T-14.5、T-14.7 |
| 19 | T-14.6 |
| 20 | T-10.1 |
| 21 | T-10.2、T-10.3 |
| 22 | T-10.4 |

## 3. 共享文件与冲突规则

并行的切片可能碰到同一批文件。下面这些文件或资源有固定规则；其余文件只能改 goal 语句「范围」里列出的那些。

| 共享对象 | 规则 |
| --- | --- |
| `.proto` 与生成代码 | 只有 goal 语句明确授权的切片才能改；未授权却需要改时标阻塞。字段号和枚举值只能追加，不能改动已有编号。T-02.1 先追加 `Bootstrap.log_level = 10`；T-02.2 依赖 T-02.1 合并后再追加 `Bootstrap.session_limits = 11`，不得并行修改同一 proto 与生成文件。 |
| `internal/core/state/schema.sql`（唯一的初始迁移） | 首次发布前只有一条初始迁移（[编码规范](coding-standards.md)「首次发布前不写兼容层」）：新表和新列直接写进 `schema.sql`，不新增迁移文件，不用 `ALTER TABLE`；新表的 GORM 模型和访问方法一起放在 `internal/core/state/` 下按主题新建的文件里，不改 `models.go`；读写遵守[编码规范](coding-standards.md)「Core 状态库不写 SQL」。同步 main 后如果和别的切片在初始迁移末尾冲突，两边的表都保留，把自己的表放在最后，再重跑 state 测试。`schema.sql` 与 GORM 模型由 T-02.3（goal-86，PR #63）引入；给已有表加列的切片同步修改 `models.go` 中对应的模型。 |
| `internal/core/adapter` 的 `Instance`、`Sink` 接口 | T-03.1 合并后不再修改。之后的切片需要更多 Adapter 能力时，在同一个包里新增可选接口，调用方用类型断言获取，不改已有接口。 |
| `internal/wire/services.go` | 固定服务名只追加常量，不重排、不改已有值。多个切片并行追加时，由先合并者保留；后合并者同步最新 main 后把自己的常量追加到文件末尾，再运行 `go test ./internal/wire/...` 和本切片服务测试。 |
| `router.New` 与 `Router.Attach` 的签名 | goal-7 给 `New` 追加 prepare、给 `Attach` 追加 Caller，goal-13 给 `New` 追加 CommandPolicy。后合并的一方在先合并者的签名末尾追加自己的参数，并同步 supervisor 测试与装配处的全部调用点，不重排已有参数。 |
| `internal/core/router` 的固定服务 mux | 新的 Core 固定服务在各自包里实现，只在装配处注册一行；服务名常量追加在 `internal/wire/services.go`。 |
| `sdk/sdk.go`、`sdk/activate.go`、`sdk/session.go` | 新功能优先放进新文件，只在这几个文件里加必要的调用点；不重排、不顺手重写已有代码。 |
| `cmd/alonabot/main.go` | 只有 goal 语句写明负责装配的切片才能改。 |
| `go.mod`、`go.sum` | 不加依赖。需要依赖时标阻塞，由维护者决定。goal 语句明确授权的已批准依赖除外（MCP go-sdk 只授权给 goal-60，只在 `agent/` 使用；cobra 由 goal-76 引入，之后的 CLI 切片直接使用，不再改 go.mod；coder/websocket 由 goal-78 引入，只在 adapters/onebot11 使用；jsonschema-go 由首个使用它的 goal 语句授权引入，用法遵守 D-43）。 |
| 日志 | T-02.1 合并后，一律使用注入的 `*slog.Logger` 与 CONTRACTS §14 的字段；不另建日志封装，不用 fmt 直接写 stderr/stdout（CLI 给用户的输出除外）。 |
| `internal/inproc/harness.go` | 只由 goal-81、goal-82 修改，只用标准库类型；字段只追加。 |
| `internal/tune/params.go` | 只追加参数，不改已合并参数的名称与缺省值；新增参数同时追加到 CONTRACTS §16 的表，由审查者在合并前同步。 |
| 根包 `alonabot` 的 Core 装配 | 只有 goal 语句写明负责装配的切片才能改，与 `cmd/alonabot/main.go` 相同。 |
| 包文档 | 只在包的职责确实变化时更新，不顺手重写。 |
| `ARCHITECTURE.md`、`README.md`、`spec/` | 执行者不改，需要同步的内容在 PR 正文「文档同步」一节列出，由审查者在合并前补上。例外是 goal 语句明确要求的证据记录。 |
| `docs/goals/`、`docs/execution-order.md` | 只由审查者维护。 |

新建 Go 包时先核对 [ARCHITECTURE.md](../ARCHITECTURE.md) 的依赖表。goal 语句会列出本切片允许新增的跨模块边；需要其他边时标阻塞。

## 4. 分支同步与合并顺序

- 执行者需要 main 的新提交时，用 `git fetch origin && git merge origin/main` 同步，不 rebase、不 force-push。同步后重跑该切片的完成条件命令。
- 审查者按 PR 就绪的先后逐个审查和合并。合并一个 PR 后，同一波次里还开着的 PR 如果和它冲突，或者 CI 需要在新 main 上重跑，由审查者通知执行者同步。
- 审查者合并前，在 PR 分支上推送一个 commit：把所属章节实施计划切片表里的状态改为「已完成」，补上 PR 正文列出的文档同步。合并后，再更新 docs/goals/ 中依赖它的 goal 的接口段和状态。

## 5. 阻塞与升级

- 执行者遇到 [执行技能](../.agents/skills/alonabot-execute/SKILL.md) 列出的阻塞情况时，标 `[!]` 并继续做不受影响的 task。所有 task 都完成或阻塞后，开草稿 PR，并在正文写明阻塞项。
- 审查者能在契约范围内解决的阻塞（例如接口细节），在 PR 中回复决定；涉及契约、依赖或范围取舍的，交给维护者决定。
- 一个 goal 被阻塞，不影响和它没有依赖关系的 goal 继续派发。
