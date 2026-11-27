# 源码开发

[English](development.md)

从源码运行一个插件，再按改动范围进入 Core、SDK、Adapter、Agent 或 WebUI。贡献流程见[贡献指南](../CONTRIBUTING.zh-CN.md)，模块与数据所有者见[架构地图](../ARCHITECTURE.md)。本文维护开发环境、命令和验证方法；管理端点与运行机制见[管理参考](management-api.zh-CN.md)，安装与运行机器人见[用户文档](https://alonabot.github.io/docs/usage/install)。

## 准备环境

| 工具 | 版本 | 何时需要 |
| --- | --- | --- |
| Go | 1.27.1，见 `go.mod` | Go 开发与组件构建 |
| protoc | 36.0，输出 `libprotoc 36.0` | 修改 `.proto` 后重新生成 |
| protoc-gen-go | v1.36.12，由 `go.mod` 的 `tool` 指令管理 | 协议生成，不需要全局安装 |
| Node | 26.5.1，见 WebUI 仓库 `package.json` | 官方 WebUI 前端开发 |
| npm | 随 Node 提供 | 按前端锁文件安装依赖 |

macOS 开发使用 asdf 管理 Go 与 Node。其他平台也需选择与上述版本相符的工具链。只改文档时无需准备全部运行时。

```fish
git clone https://github.com/AlonaBot/AlonaBot.git
cd AlonaBot
go mod download
go build -o bin/alonabot ./cmd/alonabot
```

直接运行已有示例可以复用仓库的 module 与资源，无需创建外部工程或模型凭据：

```fish
bin/alonabot run ./examples/plugins/lifecycle --data-dir ./data
```

在同一个终端输入 `/ping`，预期回复 `pong`；Ctrl-C 正常停止。要并行运行多个 Core，为每个实例选择独立数据目录与监听端口。首次管理认证和监听设置见[Core 设置](management-api.zh-CN.md#core-设置与-http-入口)。

示例的 `/ping` 已显式调用 `.Menu()`，支持菜单的 Adapter 自动按聊天发布；未标记命令仍可调用。手动接管见[命令菜单](https://alonabot.github.io/docs/tutorial/commands#平台命令菜单)。

## 查询 CLI 版本

构建 CLI 后运行 `bin/alonabot --version`。普通源码构建输出两行：

```text
alonabot version dev
https://github.com/AlonaBot/AlonaBot
```

发布构建通过 `-ldflags` 注入版本号和 UTC 构建日期，例如：

```fish
go build -ldflags '-X main.version=v0.3.0 -X main.buildDate=2026-11-23' -o bin/alonabot ./cmd/alonabot
bin/alonabot --version
```

```text
alonabot version 0.3.0 (2026-11-23)
https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.0
```

版本号可带或省略前导 `v`，显示时统一省略，Release 地址统一带 `v`。未注入日期时省略括号日期。`--version` 向 stdout 输出并以退出码 0 结束，在调优参数解析和 Core 启动之前返回。发行脚本自动注入版本号与日期，同时保留现有 Core 管理 API 的版本注入。

## 依赖管理

Core 使用 `go.mod`、`go.sum` 中的锁定版本；官方组件各自维护 module 与锁文件，WebUI 另有 `package-lock.json`。新增依赖或扩大链接范围前，说明必要性和替代方案，等待维护者批准。当前直接依赖的允许用途如下：

| 模块 | 允许用途 |
| --- | --- |
| protobuf、x/sys、google/jsonschema-go | 协议、操作系统能力及 Schema；SDK 根包还通过消息模型链接下述 Goldmark |
| github.com/yuin/goldmark v1.7.16 | 已批准用于 `contracts/markdown` 的 CommonMark 与删除线解析，供 Core 发送/编辑和 SDK `MarkdownSegment.PlainText()` 共用 |
| modernc.org/sqlite、libtnb/sqlite、GORM | Core 状态库与按需导入的 SQLite/GORM SDK 包 |
| go.yaml.in/yaml/v3 | manifest、技能元数据等 YAML |
| Cobra、pflag、go-isatty；Windows 的 mousetrap | CLI、终端识别 |
| modelcontextprotocol/go-sdk | 官方 Agent 的 MCP 客户端 |
| openai-go/v3、anthropic-sdk-go/v2 | Agent Provider；不导入 Azure、Bedrock、Vertex 等厂商子包 |
| coder/websocket | OneBot 11 Adapter；Agent 依赖的传递使用随现行锁文件保留 |

Core 可执行文件允许链接协议、Goldmark、SQLite/GORM、yaml、x/sys、Schema 与 CLI 所需模块；模型 SDK、MCP 和平台连接客户端归对应组件。插件只在导入 `sdk/gormdb` 时链接 GORM。审查时核对实际链接范围：

```fish
go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' ./cmd/alonabot ./sdk | sort -u
```

前端使用已锁定的 React、Vite、TypeScript、vitest、Lucide、dnd-kit 和 Radix；Recharts 已批准但尚未进入当前锁文件。DOM 测试环境和 Testing Library 需要另行批准。JetBrains Mono 静态字体按其 SIL OFL 1.1 许可分发。

### GORM 来源

GORM 直接使用上游 `gorm.io/gorm v1.31.2`，版本由各 module 的 `go.mod` 锁定。

独立组件的 build、pack 和开发构建沿用作者主 module 的 go.mod/go.sum，库模式直接使用作者 module。正式 GORM 组件的 build-info 列出实际使用的模块版本。

## 官方组件开发

组件源码与 Go/npm 锁文件维护在各自仓库：

| 组件 | 开发入口 |
| --- | --- |
| WebUI | [源码开发](https://github.com/AlonaBot/alonabot-webui/blob/main/docs/development.md) |
| Agent | [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) |
| Telegram | [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) |
| OneBot 11 | [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) |

在组件工程执行 `alonabot check .`、`alonabot run .` 或 `alonabot pack .`。本仓库的 Core 与 SDK 保持独立 module；跨组件测试位于独立的 `tests/go.mod`，模型 SDK 与 Adapter 实现依赖留在该测试 module 中。

```fish
uv run --no-project python scripts/integration.py
# 按改动选择测试，例如 Agent 服务：
uv run --no-project python scripts/integration.py --run '^TestAgentServices$' --race
```

整合入口先运行 `tests/agent` 与 `tests/e2e`；需要 Core 私有装配入口的 Agent 服务测试通过临时 modfile 和 overlay 运行，复用 `tests/go.mod` 实际解析出的 Agent 版本，依赖和测试文件均不写回 Core module。测试依赖可以固定真实开发提交；整合发行单独选用最新正式组件。CLI 的端到端测试在 `tests/e2e`，`cmd/alonabot` 只保留正式命令。

## 代码生成

只修改 `.proto`、生成器或 changelog 源文件，随后重新生成：

```fish
go generate ./...
```

协议源文件位于 `contracts/proto/alonabot/v1/`，`go_package` 统一为 `github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1`。生成器核对 protoc 版本后重新生成 `contracts/pb/`，protoc-gen-go 通过 `go tool` 定位。

日常变更记录在 `changelog/unreleased/{en,zh-CN}.md`；发行准备时按[更新日志格式](../changelog/README.zh-CN.md)整理到 `changelog/v<SemVer>/{en,zh-CN}.md`。生成器按语言原样拼接正式文件，写入 `internal/core/manage/changelog_core.go`；API 从文件内标题提取版本与日期。更新正式日志后运行 `go generate ./internal/core/manage`，一同交付源文件与生成内容。

## 验证改动

根据改动影响选择检查，不重复运行已有效通过的无关检查：

```fish
git diff --check
# 例如修改 Router 时：
go vet ./internal/core/router
go test -race ./internal/core/router
```

PR 的 Go CI 包括格式、生成一致性、`go vet ./...`、受影响包的普通测试和 `go build ./...`。完整门禁只在手动触发和发布流程中执行；race 测试、跨组件整合测试、平台构建和发行检查拆成独立 job 并行运行。命令以[工作流](../.github/workflows/ci.yml)为准。WebUI 独立仓的 CI 另执行前端类型检查与构建。手动需要完整检查时执行：

```fish
go build ./...
go vet ./...
go test -race ./...
uv run --no-project python scripts/integration.py --race
```

改变插件页面套件时，运行 `go test -race -v ./sdk/page ./examples/plugins/showcase` 和 `node --check sdk/page/kit.js`，并在 WebUI 的 iframe 中检查关键交互、语言与主题。文档修改检查链接、锚点、代码块和实际命令。

跨平台修改同时核对真实运行与交叉构建。五个发行目标为 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64；构建结果与平台实跑结果分别记录，现有结果见[发布说明](../changelog/v0.1.0/zh-CN.md#added)。真实 Telegram 或付费 Provider 验证使用授权的测试凭据，模拟服务在验证结果中明确注明。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml)在目标为 main 的 PR 和手动触发时运行；Release 工作流在构建与发布产物前调用同一门禁。推送到 main 不触发此工作流。`Repository checks` 检查空白与 main 的文件范围，纯文档 PR 在此结束。

Go 或 CI 改动的 `PR checks` 执行格式、生成一致性、vet 和根 module 构建。普通测试覆盖改动 Go 文件所在的根 module 包；`contracts/` 改动增加协议测试，根 `go.mod`/`go.sum` 改动运行根 module 的全量普通测试，已删除的包跳过。`tests/` 改动运行测试 module 的普通测试，沿用串行执行与 30 分钟包期限。

手动触发和发布在仓库检查后并行运行 `Race tests`、`Integration tests`、`Platform builds` 与 `Release tools`；发布检查使用对应发行 tag。所有 Go job 按根 `go.sum` 与 `tests/go.sum` 启用依赖和构建缓存；完整 job 的失败摘要使用英文，并保留各步骤的完整日志。PR 成功记录轻量检查，race、整合与平台构建证据分别取自对应完整 job。main 路径检查排除个人维护资料，产品中的技能功能和测试工具照常保留。

[PR metadata](../.github/workflows/pr-metadata.yml)验证非草稿 PR 的说明；triage 和 auto-assign 只处理 GitHub 元数据，不执行 PR 代码。代码检查使用只读权限，不接触生产 secret。分支保护另由维护者配置。

## 开发运行与热部署

先构建 CLI：`go build -o bin/alonabot ./cmd/alonabot`。在仓库根目录用以下三步得到第一条回复（使用 fish）：

1. 创建工程：`bin/alonabot init ./alonabot-plugin-hello`。
2. 运行：`bin/alonabot run ./alonabot-plugin-hello`。
3. 在同一个终端输入 `/hello`，看到 `Hello AlonaBot`；Ctrl-C 停止 Core。

有开发目录时缺省自动重载；修改源码后再输入 `/hello` 即可观察新回复。`--reload=false` 关闭重载。

回环注入、配置、热重载与同进程调试见 [devbot 教程](../examples/devbot/README.zh-CN.md)。

组件作者的完整用法在文档站：命令与参数见 [CLI 参考](https://alonabot.github.io/docs/reference/cli)，插件测试见[插件测试](https://alonabot.github.io/docs/tutorial/testing)与[测试进阶](https://alonabot.github.io/docs/advanced/testing)，同进程运行见[架构页「同进程模式」](https://alonabot.github.io/docs/internals/architecture#同进程模式)。

修改相关实现时，对应的集成测试位于：

- 库模式公开入口：根目录 `alonabot_test.go` 的 `TestLibraryLifecycle`；装配、内部服务、候选与 WebUI：`internal/app/`；开发 CLI 与热部署：`tests/e2e/`。
- Adapter 调试目录与调用：`go test -race ./internal/core/adapter/remote -run '^TestDebugAdapter$'`。
- 插件测试套件：`go test -race -count=2 ./plugintest/... ./examples/plugins/...`；连接运行中 Core 的链路见 `plugintest/debug_test.go`，LLM 替身见 `plugintest/llm_test.go`。

## 协议与会话行为

组件协议版本 **1.0** 与产品版本 **0.3.0** 独立。当前消息类型与字段号由 [contracts/proto/alonabot/v1](../contracts/proto/alonabot/v1) 定义，[wire 运行时](../internal/wire) 实现分帧与 Session 调用。修改 `.proto` 源文件后重新生成 Go 产物。保留已发布的字段号与枚举值；删除的名称和编号须 reserved，已有字段类型保持兼容。[协议编号检查](../contracts/wirespec_test.go) 保护这些标识。

TCP 与 stdio 均使用四字节无符号大端长度，后接 Protobuf Envelope。长度不含前缀，须为 1～4,194,304 字节。零长度、超限、截断或无法解码的帧使 Session 失败，读取端不猜测下一帧边界。每条连接只有一个 writer。Core 持有回环 TCP listener（`127.0.0.1:0`）直到结束。每次启动组件创建新 Session 与 32 字节 secret；认证恒时核对 secret，并核对身份、digest 和协议版本。stdio 的 stdin/stdout 传输协议帧，stderr 输出日志。Bootstrap 中的秘密、配置、数据库及 HTTP 凭据须留在日志之外。

激活顺序为 HELLO → HELLO_ACK/Bootstrap → REGISTER → PREPARED → ACTIVATE → OnReady → ACTIVATED，激活后才开始普通业务调用。协议 major 必须相同，协商 minor 取双方支持的较小值；required capabilities 必须双方支持，`http_listener` 版本为 1。注册与生命周期载荷见 [control.proto](../contracts/proto/alonabot/v1/control.proto) 和 [route.proto](../contracts/proto/alonabot/v1/route.proto)。

调用按 `(session_id, initiator_side, call_id)` 匹配。call ID 非零、按方向单调递增，在同一 Session 内不复用。父调用须为同一 Session 中仍在途的 Core 发起调用。deadline 为绝对 UTC 毫秒，子调用沿用或缩短父期限。取消为协作式，已完成的副作用保留。载荷与 ErrorInfo 可同时存在以表示部分结果，调用方须保留两者。请求默认超时为 30 秒。Session 失败时，挂起调用取得真实错误；认证无效时关闭连接，避免暴露凭据。

用 `alonabot tune list` 查看完整的当前调优表，包括名称、单位、默认值、生效值、约束与来源；`--args NAME=value` 使用同样的校验。[internal/tune/params.go](../internal/tune/params.go) 定义数值，[插件休眠](management-api.zh-CN.md#插件休眠) 说明唤醒预算，[插件指南](https://alonabot.github.io/docs/tutorial/project) 说明延续、存储与任务。帧上限固定为 4 MiB。阶段预算默认 HELLO 10s、配置 60s、迁移 15m、OnStart 60s、OnReady 30s；停止总预算 10s，其中排空 6s、OnStop 4s。崩溃恢复窗口为五分钟，最多三次。组件自己的 Handler 与后台活动须响应取消。

### 菜单协议新增字段

Core→Adapter 的 `adapter.command.menu` 执行 set/get/delete；插件→Core 的 `core.command.menu` 执行 set/get/delete/status/sync，省略实例时取在途父调用的可信实例。两者使用 schema_version=1。`AdapterInstanceStartResponse.command_menu=5` 为可选消息，缺失时沿用 `adapter.commands.sync`；其 CommandMenuConfig 含 bool auto_sync_commands=1、clear_commands_on_disable=2。`CommandSpec.menu=14` 缺省 false；可信 `MediaSource.mime=6` 为 string。

共用消息为 `CommandMenuRequest{instance_id=1, operation=2, menu=3}`、`CommandMenu{scope=1, language_code=2, commands=3}`、`CommandMenuScope{type=1, chat_id=2, user_id=3}`、`CommandMenuEntry{name=1, description=2}`、`CommandMenuResponse{commands=1, config=2, ready=3, last_error=4, positions=5}`、`CommandMenuPosition{scope=1, language_code=2, mode=3, owner_plugin_id=4}`，commands/positions 均 repeated。scope 类型为 default/chat/chat_administrators/chat_member，插件只开放后三种，thread 不进入 scope。

`EventPayload.kind` 追加 command_menu_ready=13，string 字段为 adapter_id=1、adapter_instance_id=2、account_id=3、chat_id=4。Core 定向发送 `bot.command_menu_ready`，空 chat_id 表示实例手动就绪，非空表示聊天可用。Adapter 上报标准事件仍沿用原白名单。定义见 [Adapter 协议](../contracts/proto/alonabot/v1/adapter.proto)、[路由](../contracts/proto/alonabot/v1/route.proto)与[消息](../contracts/proto/alonabot/v1/message.proto)，生命周期与准入见[命令菜单](https://alonabot.github.io/docs/tutorial/commands#平台命令菜单)。

### 消息描述与只读预检协议

`core.message.describe` 接收 `DescribeMessageRequest{to=1}`，返回 `MessageDescription{formats=1, extensions=2, combinations=3}`；`core.message.check` 接收 `CheckMessageRequest{to=1, chain=2}`，返回 `CheckMessageResponse{status=1}`，均使用 protobuf schema version 1。`to` 缺省时按可信父目标补齐，目标与发送共用当前准入规则。SDK 的交互与定时用法见[出站消息 API](agent-api.zh-CN.md#按目标查询与预检出站消息)。

`AdapterDescriptor` 追加 message_formats=7、message_combinations=8；`ExtensionDeclaration` 追加 description=4、example_json=5。`MessageFormat` 字段顺序为 name、version、status、description、schema_json、example_json、sources、text_limit、button_label_limit、adapter_id（编号 1～10）；`MessageCombination` 为 name=1、formats=2、description=3、example_json=4。SupportStatus 的 0/1/2 分别表示 unknown/supported/unsupported。消息定义见 [message.proto](../contracts/proto/alonabot/v1/message.proto)。

可选 `adapter.message.check` 复用 `AdapterSendRequest` 的实例、目标、消息链和代次，返回 CheckMessageResponse；实例真实实现 `sdk/adapter.MessageChecker` 时，`adapter.operations` 公布 `sdk/message.check`。Core 先核对当前目录及 Session/代次，再调用 RPC。媒体 Store 的 `PrepareCheckSource` 只读授权元数据，保留原来源和 token，错误与预检状态可同时返回。

`ToolConfirmRequest.require_confirmation=4` 为 bool。`sdk/tool.ConfirmRequired` 设置该字段，使 allow 策略同样进入可信 Actor 的确认链；deny 仍优先。既有 Confirm 不设置该字段，沿原策略执行。

## Agent、流式服务与生命周期 API

[完整参数、能力状态、八阶段钩子与 Goal 恢复契约](agent-api.zh-CN.md)。
