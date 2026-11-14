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
| 1 | T-02.1、T-02.3、T-02.4、T-07.1、T-08.2、T-09.2、T-12.1 |
| 2 | T-02.2、T-12.9、T-12.10、T-12.11 |
| 3 | T-03.1、T-12.12 |
| 4 | T-03.2、T-03.3、T-03.5、T-04.1、T-11.3、T-12.13 |
| 5 | T-03.4、T-03.12、T-04.2、T-04.3、T-06.1、T-07.3 |
| 6 | T-03.6、T-03.7、T-03.23、T-06.3、T-07.4、T-08.1 |
| 7 | T-03.8、T-03.9、T-03.27、T-04.6、T-08.3、T-08.4、T-08.10、T-09.1 |
| 8 | T-03.11、T-04.4、T-04.7、T-07.6、T-08.5、T-08.7、T-08.8、T-09.16、T-15.1 |
| 9 | T-03.24、T-04.9、T-04.10、T-05.1、T-07.7、T-08.6、T-08.11、T-11.1、T-15.2 |
| 10 | T-04.8、T-07.2、T-08.9、T-08.12、T-08.13、T-13.1 |
| 11 | T-03.13、T-03.14、T-03.15、T-03.16、T-03.17、T-03.18、T-03.19、T-04.12、T-05.2、T-07.5、T-09.4、T-13.2 |
| 12 | T-03.20、T-03.22、T-04.11、T-04.13、T-04.14、T-05.3、T-05.4、T-07.8、T-08.14、T-09.12、T-13.3、T-14.1、T-14.2、T-15.4、T-15.5、T-15.6、T-16.1 |
| 13 | T-03.10、T-03.21、T-03.25、T-05.5、T-06.2、T-07.9、T-09.10、T-09.13、T-12.2、T-15.3 |
| 14 | T-03.26、T-04.5、T-06.4、T-11.2、T-12.8、T-12.14、T-13.4、T-15.7、T-16.2 |
| 15 | T-09.3、T-09.5、T-11.4、T-12.3、T-12.4、T-16.3 |
| 16 | T-04.15、T-09.6、T-12.5、T-12.6 |
| 17 | T-04.16、T-09.7、T-09.8、T-09.14、T-12.7 |
| 18 | T-09.9、T-09.15、T-13.5、T-14.3 |
| 19 | T-09.11、T-13.6、T-14.4、T-14.5、T-14.7 |
| 20 | T-14.6 |
| 21 | T-10.1 |
| 22 | T-10.2、T-10.3 |
| 23 | T-10.4 |

## 3. 共享文件与冲突规则

并行的切片可能碰到同一批文件。下面这些文件或资源有固定规则；其余文件只能改 goal 语句「范围」里列出的那些。

| 共享对象 | 规则 |
| --- | --- |
| `.proto` 与生成代码 | 只有 goal 语句明确授权的切片才能改；未授权却需要改时标阻塞。字段号和枚举值只能追加，不能改动已有编号。T-02.1 先追加 `Bootstrap.log_level = 10`；T-02.2 依赖 T-02.1 合并后再追加 `Bootstrap.session_limits = 11`，不得并行修改同一 proto 与生成文件。 T-03.13～T-03.19（D-87）都追加 proto 字段，可以并行开发：语句不预先指定字段号，后合并的一方同步最新 main 后在先合并者之后顺延编号、重新运行 `go generate ./...` 与 `contracts/wirespec_test.go`。 |
| `internal/core/state/schema.sql`（唯一的初始迁移） | 首次发布前只有一条初始迁移（[编码规范](coding-standards.md)「首次发布前不写兼容层」）：新表和新列直接写进 `schema.sql`，不新增迁移文件，不用 `ALTER TABLE`；新表的 GORM 模型和访问方法一起放在 `internal/core/state/` 下按主题新建的文件里，不改 `models.go`；读写遵守[编码规范](coding-standards.md)「Core 状态库不写 SQL」。同步 main 后如果和别的切片在初始迁移末尾冲突，两边的表都保留，把自己的表放在最后，再重跑 state 测试。`schema.sql` 与 GORM 模型由 T-02.3（goal-86，PR #63）引入；给已有表加列的切片同步修改 `models.go` 中对应的模型。 |
| `adapters/telegram/cmd/alonabot-adapter-telegram/adapter_config_schema.json` | goal-101（T-08.12）新建该文件；goal-117（T-03.23）追加 `proxy_mode` 并给 `proxy` 加 condition，goal-111（T-03.20）追加 `file_base_url`。字段只追加，不重排已有字段。goal-117 派发时该文件还不存在就不新建；goal-101 后合并时同步 main，按 CONTRACTS §8 声明 main 上 Start 已解析的全部字段（含 `proxy_mode`）。 |
| `adapters/telegram/inbound.go` 的入站标准化 | goal-137 在映射消息内容之前分出服务消息；goal-111 填写 chat_type、goal-132 填写 chat_title，二者都在 `baseEvent` 中进行。服务消息的扩展事件同样经 `baseEvent` 构造，因而自动带上两者。三者可以并行，后合并者同步 main 后保留先合并者的改动。 |
| `internal/core/adapter` 的 `Instance`、`Sink` 接口 | T-03.1 合并后不再修改。之后的切片需要更多 Adapter 能力时，在同一个包里新增可选接口，调用方用类型断言获取，不改已有接口。 |
| `internal/wire/services.go` | 固定服务名只追加常量，不重排、不改已有值。多个切片并行追加时，由先合并者保留；后合并者同步最新 main 后把自己的常量追加到文件末尾，再运行 `go test ./internal/wire/...` 和本切片服务测试。 |
| `router.New` 与 `Router.Attach` 的签名 | goal-7 给 `New` 追加 prepare、给 `Attach` 追加 Caller，goal-87 给 `New` 追加 prefixes，goal-13 给 `New` 追加 CommandPolicy。后合并的一方在先合并者的签名末尾追加自己的参数，并同步 supervisor 测试与装配处的全部调用点，不重排已有参数。goal-15 把 `New` 改为 `New(opt Options)`、`Close` 改为 `Close(ctx) error`：它合并后，新的注入项只追加 `Options` 字段，其他语句中 `router.New` 的位置参数写法按同名字段理解。 |
| `sdk/generated.go` 的 `Generated`、`sdk.Collect` 与 `sdk/generate` | goal-53 只做三处追加：给 `Generated` 加 `I18n` 字段，在 `Collect` 中检查 usage_key，在生成器中检查并嵌入 i18。goal-88 重写注册 API，并为 Service 生成 Schema。goal-95 在 goal-88 合并之后进行，删除配置专用的标签解析与 Config 类型发现、改为检查并嵌入 `config_schema.json`，`service.go` 的 Service Schema 生成不依赖配置部分，保持不变；i18n/ 中 `config` 的结构检查属于 goal-53，与 schema 的核对属于 goal-95。goal-96 在 goal-88 合并之后删除构建期标记的扫描与生成，不改其他生成内容。先合并者的代码保留；后合并者同步 main 后在新代码上重做自己的改动，不回退对方的修改。 goal-91 排在 goal-96 之后：生成器的标记事件输出用字符串事件类型，与事件常量不兼容。goal-123 只在生成器中新增对 T 调用的检查，接入已有的诊断输出，与 goal-101 冲突时同样由后合并者重做。goal-124 在生成器中记录 Handler 的参数名并嵌入 `Generated`，与 goal-101、goal-123 冲突时同样由后合并者重做。 |
| `router.Plugin` 接口 | goal-53 追加 `Messages()`。之后需要更多插件信息的切片同样只追加方法，并同步 `instancePlugin` 与测试中嵌入该接口的包装类型。 |
| `spec/ADMIN-API.md` 与管理服务的 HTTP 路由 | goal-84 新建 `spec/ADMIN-API.md`，goal-39 建立管理服务的 HTTP 路由。之后新增 Admin API 的切片（goal-66、goal-92、goal-93）在自己的新文件里实现端点，只在路由注册处和 `spec/ADMIN-API.md` 末尾追加，不改已有条目；并行时由先合并者保留，后合并者同步 main 后追加。 |
| `web/admin/src/api/` 与 `web/admin/i18n/*.json` | goal-43 建立 Admin API 客户端与页面文案；goal-44、goal-94 只追加自己的客户端函数与文案键，冲突时两边都保留，en-US 与 zh-CN 的键集合保持一致。 |
| `internal/core/router` 的固定服务 mux | 新的 Core 固定服务在各自包里实现，只在装配处注册一行；服务名常量追加在 `internal/wire/services.go`。 |
| `sdk/sdk.go`、`sdk/activate.go`、`sdk/session.go` | 新功能优先放进新文件，只在这几个文件里加必要的调用点；不重排、不顺手重写已有代码。 |
| `cmd/alonabot/main.go` | 只有 goal 语句写明负责装配的切片才能改。 |
| `sdk/` 与 `internal/core/supervisor/` 的测试和 testdata | goal-103 在 goal-96、goal-95、goal-54 之前合并，这样后者不必迁移即将删除的测试。goal-103 合并后，其他切片不再向这两处提交测试（`supervisor_test.go` 除外），也不新增 testdata 测试插件。 |
| `internal/core/settings` 与 `core_config.json` | goal-97 用 `core_config.json` 取代 `alonabot.yaml`，并提供读取、写回与重新读取的函数。之后修改 Core 设置的切片（goal-39 的 Adapter 实例、goal-66 的日志级别、goal-69 的重新读取、goal-93 的 Core 设置 API、goal-101 的版本迁移、goal-121 的实例名规则与 `default_instance`、goal-122 的 admins 格式）只调用或追加这些函数，不在别处直接读写该文件。goal-101 与 goal-121 都改 `readAdapters`：goal-121 只在其中接受 `default_instance` 一个键，名称检查与 `DefaultInstances` 放在新文件里；冲突时后合并者保留两边的改动。goal-122 只改 `Admin` 类型与 admins 的读取、校验和写回，与 goal-121 同在第 9 波，后合并者同样保留两边的改动。 |
| SDK 公共 API 中的发送目标 | goal-91 合并后，SDK 的公共签名不再有 `*pb.Target` 或目标结构：目标是会话字符串 `message.Session`，机器人用可选的 `sdk.Via`（D-102）。其他语句中 SDK 签名里的 `*pb.Target`（例如 goal-51 的 `sdk.Edit`）在 goal-91 合并后按此理解，由审查者在合并后改写语句；goal-17 的 Hook 已按 SDK 类型写出（D-112）。指向某条消息的参数一律用消息句柄 `message.MessageRef`（D-104），goal-51 的 `c.Edit` 与 `sdk.Edit` 由 goal-91 改写；Send 与编辑收 `message.Content`，自动回复型 Handler 返回 `(message.Content, error)`（D-108）。Core 与 SDK 的内部代码照旧使用 pb.Target。 |
| 插件入口与注册的接收者 | goal-125 合并后，插件入口是 `func New() *sdk.Plugin`，注册方法在插件对象或子插件上，回调与迁移用登记函数（D-114～D-116）。其他语句中的 `sdk.Root`、`sdk.Base`、根插件的注册方法、`*sdk.CommandRegistry` 等注册器参数与迁移接口在 goal-125 合并后按此理解，由审查者在合并后改写语句；goal-124 的要点已按插件对象注明。goal-125 与 goal-16、goal-68 可以并行，都改 sdk 的注册与装配，后合并者同步 main 后重做自己的改动。 |
| `cmd/alonabot` 的命令树 | goal-97 追加 `admin-token show`；goal-69 改写 `run` 的参数为组件目录与 `--reload`；goal-38 把 `plugin generate`、`plugin check` 改为顶层的 `build`、`check` 并追加 `init`、`test`、`pack`。各自在自己的命令文件中实现，只在命令注册处追加或替换自己的一行；并行时后合并者同步 main 后重做注册，不回退对方的修改。 |
| `go.mod`、`go.sum` | 不加依赖。需要依赖时标阻塞，由维护者决定。goal 语句明确授权的已批准依赖除外（MCP go-sdk 只授权给 goal-60，只在 `agent/` 使用；cobra 由 goal-76 引入，之后的 CLI 切片直接使用，不再改 go.mod；coder/websocket 由 goal-78 引入，只在 adapters/onebot11 使用；jsonschema-go 由首个使用它的 goal 语句授权引入，用法遵守 D-43）。 |
| 日志 | 按[编码规范「日志用默认 Logger，依赖显式传递」](coding-standards.md#日志用默认-logger依赖显式传递)取用默认 Logger，并遵循 CONTRACTS §14 的字段；不另建日志封装，不用 fmt 直接写 stderr/stdout（CLI 给用户的输出除外）。 |
| `agent/provider/` 的新协议 | goal-118、goal-119、goal-120 可以并行。每个切片把自己协议的请求编码与流解析放在新文件（`responses.go`、`gemini.go`、`ollama.go`），`provider.go` 只追加 Protocol 常量，并在 `New` 的协议校验与 `Stream` 的分派中各加一个分支；`TestStream` 的用例表、`mockProvider` 与 `testRequest` 只追加。两处共用的代码由先合并者加入，后合并者同步 main 后复用、不重复定义：`ToolCall.Replay`（goal-118 与 goal-119）放在 `provider.go`；按 ToolCallID 查找工具名与生成调用 ID 的辅助（goal-119 与 goal-120）放在 `toolcall.go`。goal-134（改用 SDK）与 goal-135（图片输入）要改全部协议文件，按前置在三者都合并后依次进行，不与它们并行。 |
| `agent/` 的配置（`config_schema.json` 与 i18n 的 `config`） | goal-83 与 goal-136 都在 goal-59 之后给 Agent 配置追加字段，可以并行：字段只追加，不重排已有字段；后合并者同步 main 后把自己的字段放在最后，再运行 `alonabot check`。 |
| `internal/inproc/harness.go` | 只由 goal-81、goal-82 修改，只用标准库类型；字段只追加。 |
| `internal/tune/params.go` | 只追加参数，不改已合并参数的名称与缺省值；新增参数同时追加到 CONTRACTS §16 的表，由审查者在合并前同步。 |
| 根包 `alonabot` 的 Core 装配 | 只有 goal 语句写明负责装配的切片才能改，与 `cmd/alonabot/main.go` 相同。goal-67 建立装配并首次调用 `router.New`，prefixes 传 nil，不读设置文件；goal-97 不改根包。命令前缀、Core Admin 列表与默认 locale 由 goal-39 接入。goal-39、goal-68、goal-69 都要在装配中读取 `core_config.json`：先合并者加入 `settings.Load` 这一步，后合并者复用同一个 `*settings.File`，不重复加载。 |
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
