# 源码开发

[English](development.md)

从源码运行一个插件，再按改动范围进入 Core、SDK、Adapter、Agent 或 WebUI。贡献流程见[贡献指南](../CONTRIBUTING.zh-CN.md)，模块与数据所有者见[架构地图](../ARCHITECTURE.md)。本文维护开发环境、命令和验证方法；部署操作见[安装与部署](deployment.zh-CN.md)。

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

在同一个终端输入 `/ping`，预期回复 `pong`；Ctrl-C 正常停止。要并行运行多个 Core，为每个实例选择独立数据目录与监听端口。首次管理认证和监听设置见[部署指南](deployment.zh-CN.md#core-设置与-http-入口)。

示例的 `/ping` 已显式调用 `.Menu()`，支持菜单的 Adapter 自动按聊天发布；未标记命令仍可调用。手动接管见[命令菜单](plugin-guide.zh-CN.md#命令菜单)。

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
| Help | [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) |

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

已安装 CLI 时，可在工程目录直接运行 `alonabot run`。stdin 是终端且设置文件没有任何 Adapter 实例时，自动启用内置终端对话；`--console` 强制启用并接受管道输入，`--console=false` 关闭。终端用户 `console:local` 在本次运行中具有 Core Admin 权限；实例和管理员保持在内存中。回复中的媒体显示为 `[image]`、`[file: 名称]` 等占位，输入 `#1` 点击最近一条带按钮消息的第一个回调按钮，编辑显示为 `(edited)` 新段。stdin 结束后 Core 继续运行，直到收到停止信号。

有开发目录时缺省自动重载；修改源码后再输入 `/hello` 即可观察新回复。`--reload=false` 关闭重载。

回环注入、配置、热重载与同进程调试见 [devbot 教程](../examples/devbot/README.zh-CN.md)。

### 工程创建与 SDK 选择

组件市场 ID 使用连字符前缀，例如 `alonabot-plugin-hello`；本地 ID 仍允许下划线，规则见 [插件指南](plugin-guide.zh-CN.md)。

验证官方 help 工程时先检出独立的 `alonabot-plugin-help` 仓库，执行 `bin/alonabot check <help 工程目录>`。使用回环 Adapter 试用时，在 `run` 的组件目录列表追加该工程目录，再把注入请求的 `text` 改为 `%2Fhelp` 或 `%2Fwhoami`。

先构建 CLI（`go build -o bin/alonabot ./cmd/alonabot`），再创建与验证组件：

```fish
bin/alonabot init ./alonabot-plugin-hello
bin/alonabot build ./alonabot-plugin-hello
bin/alonabot check ./alonabot-plugin-hello
bin/alonabot test ./alonabot-plugin-hello
bin/alonabot pack ./alonabot-plugin-hello
```

`init`、`build`、`check`、`test`、`pack` 至多接受一个位置目录，省略时使用当前目录；旧的 `--dir` 会报未知参数。`init` 要求空目录，非空时在写入前报错。`--id` 缺省取目标目录绝对路径的目录名；例如 `alonabot-plugin-hello` 直接作为市场 ID，`my_plugin` 保持本地 ID 并输出缺市场前缀的警告。非法目录名会保留原字符并提示显式给出 `--id`。缺省角色为 plugin，`--role agent|adapter|webui` 创建对应工程；Adapter 的 `--adapter-id` 缺省等于最终组件 ID，可显式覆盖。插件与 Agent 导出 `New()`，包含一个直接回复 `Hello AlonaBot` 的英文示例命令与 schema；需要翻译时自行添加 i18n 目录；Adapter 提供实例 schema，WebUI 提供 `assets/index.html`。各角色生成的 Go 文件和 `plugin.yaml` 均无注释。模板忽略 `data/` 与 `bin/`，正式薄入口通过 overlay 提供。库模式宿主参考 [devbot](../examples/devbot/README.zh-CN.md)。

位于现有 Go module 中时复用该 module，保持其 go.mod；外部空目录的 module 缺省同为目录名，可用 `--module example.com/myplugin` 覆盖。发布版 CLI 用自身模块版本写 SDK require，并在生成的工程目录运行 `go mod tidy`，随后可直接 `run`。依赖解析失败时保留工程，命令以失败状态退出并显示手动补齐的命令。开发版 `(devel)`（含本地修改的 `+dirty` 构建，以及从仓库 checkout 构建时 Go 自动盖上的伪版本）会提示执行 `go get github.com/AlonaBot/AlonaBot@<version>`，由作者选定版本。发布前修改 `plugin.yaml` 中的示例 source。

### 运行、传输与重载

在包含 `plugin.yaml` 的组件工程中运行 `alonabot run`，或者明确给出一个或多个工程目录：

```sh
bin/alonabot run ./alonabot-plugin-hello ./myadapter --data-dir ./data --reload
bin/alonabot run ./alonabot-plugin-hello ./mywebui --data-dir ./data --reload
bin/alonabot run ./examples/plugins/lifecycle --stdio
```

未给目录时，当前目录含 `plugin.yaml` 就运行该工程；其他情况按安装记录启动 Core。组件身份与角色取自工程 manifest，完成工程检查与进程组件构建后才启动组件。重复目录、重复 ID 都会报错；进程组件的 ID 或 `adapter_id` 与已安装包冲突时也会报错。

构建复用内部生成器与 `go build`，可执行文件写到 `<data_dir>/dev/bin/`，生成资源与薄入口通过 overlay 提供，作者目录中的源码和 module 文件保持原样。插件与 Agent 构建入口包；Adapter 工程导出 `New()` 并提供可导入的包，构建器为其生成调用 `sdk.RunAdapter(New())` 的薄入口。

开发组件通过 supervisor 在独立进程中运行，缺省使用 TCP loopback，`--stdio` 切换本次全部 Native 独立进程的传输，覆盖开发组件、已安装组件及其候选、重启、重载、恢复和休眠唤醒。也可用 `alonabot run --stdio` 启动已安装组件；传输选择只作用于本次运行，纯静态 WebUI 与同进程组件沿用既有机制。组件固定使用 `dev:<组件 ID>` 代次，配置每次从 `<data_dir>/plugins/<id>/config.json` 读取，缺失时按空对象准备；准备后的配置保存在开发代次快照中。受管 SQLite 为同目录的 `dev.sqlite`，跨重启保留。Adapter 实例配置来自 `core_config.json` 的 `adapters`。开发日志缺省为 debug，显式 `log.sources` 设置优先。

`--reload` 每 `DEV_RELOAD_POLL` 检查源码、`//go:embed` 资源、`go.mod`、`go.sum`、`plugin.yaml`、两个配置 Schema 和 `i18n/`，排除数据目录、vendor、testdata 及以 `.`、`_` 开头的目录。轮询间隔缺省为 500ms，例如 `--args DEV_RELOAD_POLL=1s` 可以覆盖。

源码或资源变化后重新构建，成功时只重启变化的组件；编译失败保留编译器诊断，旧组件继续运行。只改组件 `config.json` 时直接重启该组件，只改 Adapter 的一个实例配置时只重启该实例，二进制保持原样。`--reload=false` 时构建并运行一次；没有开发目录时保持普通 Core 运行，停止轮询。

`--reload` 也监视 `core_config.json` 的其他设置：`admins` 更新后立即改变 Core Admin 权限，命令前缀从共享设置读取当前值，`default_locale` 对之后建立的 Session 生效。修改 Adapter 的 `default_instance` 后，未指定实例的发送立即改选新的缺省实例；`log.level` 立即更新全局日志级别，本次运行显式给出的 `--log-level` 或库模式 `LogLevel` 选项持续优先。这些修改保持组件进程与实例运行；同时修改设置与开发 Adapter 实例时，先刷新设置，再重启变化的实例。重新读取失败时记录错误并保留旧设置。开发 CLI 经 `internal/app.Core.ReloadSettings(ctx)` 与其他管理操作串行刷新同一份设置。

`role: webui` 的工程直接以 manifest 中 `assets` 指定的目录为静态根目录，覆盖已安装的 WebUI，等同于 `alonabot.WebUI(os.DirFS(<assets 目录>))`。一次运行只接受一个 WebUI 目录，重复时在启动任何组件前报错。开发 WebUI 按开发模式约定跳过 Admin API 主版本检查；静态文件按请求读取，修改后立即可见，无需构建或重启，也无需 `--reload`。

独立进程调试时将调试器附加到组件 PID；需要在同一进程调试 Core 和组件时使用下文库模式。

## 作者工具

插件包导出 `func New() *sdk.Plugin`，在 `sdk.NewPlugin()` 返回的对象上登记命令、生命周期回调与业务行为。构建工具自动扫描 manifest、配置 Schema、Service Schema、i18n 和参数名，经 overlay 编入程序；作者工程中没有 `alonabot_gen.go`，作者代码无需生成相关内容。构建或只读检查：

```fish
bin/alonabot build ./alonabot-plugin-hello --goos linux --goarch amd64 --tags a,b
bin/alonabot check ./alonabot-plugin-hello --goos linux --goarch amd64 --tags a,b
```

`build/check` 使用 `--goos`、`--goarch`，目标缺省沿用 Go 环境；`pack` 使用可重复的 `--target <os>/<arch>`，省略时取本机平台。三者均用 `--tags` 选择源码；扫描、依赖分析和正式编译统一使用 `CGO_ENABLED=0`。开发运行的 BuildDev 沿用开发环境的 cgo 设置。`build` 默认输出到 `<工程>/bin`，可用 `--out DIR` 覆盖。进程产物包含独立的 `component`（Windows 为 `component.exe`）、资源旁车及包内 `plugin.yaml`；作者 manifest 保持原样。纯静态 WebUI build 发布静态资源与包内 manifest；带 Go 入口的 WebUI 另外生成同平台进程，官方 WebUI 采用这一形式。`check` 只读扫描入口与资源，保留全部静态诊断与位置。

`test` 在全新临时数据目录中构建宿主平台组件，完成真实握手、准备、注册、激活与停止，报告注册命令及清理结果；Adapter 保持零平台实例，WebUI 检查静态加载资源。进程组件的 Bootstrap 下发临时状态根内 `plugin_data/<id>/` 的绝对路径，测试结束时随临时目录清理。准备、注册、停止或清理失败均非零退出。

`pack` 为每个 `--target` 分别扫描源码、检查资源并构建，输出 `<id>-<version>-<os>-<arch>.tar.gz`；静态 WebUI 只输出一个 `<id>-<version>-any.tar.gz`，带 Go 入口的 WebUI 按平台构建。`pack` 只生成组件归档，`--out` 默认同 build；Release 上传这些归档，源码由 GitHub 自动提供。进程包的 `build-info` 是 `debug.BuildInfo.String()` 文本；生成代码、资源旁车与归档共用扫描时校验的资源快照，配置 schema 使用同一规范内容；扫描后保存的工程资源留待下次构建。

配置字段的 `example` 引用图片时，`check`、`build`、`pack` 检查两种配置 schema 中的引用及图片签名；被引用的图片随资源快照进入构建目录和归档。示例写法见 [插件开发规范](plugin-guide.zh-CN.md#配置字段的回复示例)。

```fish
bin/alonabot pack ./alonabot-plugin-hello --target linux/amd64 --target linux/arm64 \
  --out ./dist
```

安装直接读取归档中的 `plugin.yaml`，在线安装使用 GitHub Release API 选择对应平台的归档。运行配置与数据目录应留在打包资源之外。

## 库模式运行（同一进程）

根包 `alonabot` 提供 `Run`、`Start`、`*Core` 与 `Wait`，选项为 `Plugin`、`Adapter`、`WebUI`、`DataDir`、`Args`、`Listen`、`LogLevel`、`LogFormat`。它们转发到 `internal/app`，在当前进程中运行 Core 与开发中的组件，示例见 `examples/devbot`：

```go
core, err := alonabot.Start(ctx, alonabot.DataDir(dir), alonabot.Plugin(plugin.New()), alonabot.Adapter(adapter.New()))
if err != nil {
	return err
}
return core.Wait() // ctx 取消（例如 SIGTERM）后在 STOP_TOTAL 内停止全部组件
```

```fish
bin/alonabot build ./examples/devbot
examples/devbot/bin/devbot -data-dir ./data
```

组件以 source=dev 登记，代次固定为 `dev:<组件 ID>`。库模式由 `alonabot build <main 目录>` 构建，扫描 main 依赖中的组件入口（含模块缓存中的第三方组件），将全部资源登记经 main 包的 overlay 编入。默认只输出 `<main 目录>/bin/<目录名>` 可执行文件（Windows 加 `.exe`），可用 `--out` 修改输出目录。组件源码与模块缓存保持只读。同包多次 New 共用资源；main 的包级变量也可调用组件的 New。直接 `go build` 的库模式产物在启动缺少资源的组件前返回提示 `alonabot build` 或 `alonabot run` 的英文错误。

库模式更改源码或资源后重新运行 `alonabot build` 并重启程序。未捕获的后台 panic 使整个进程非零退出且无有限重启；SDK Handler 边界内捕获的 panic 记录 stack 并返回 `INTERNAL`。同进程组件超过停止预算仍未退出时，`Wait` 返回错误；独立进程由 supervisor 隔离并有限恢复，两种模式的 SIGINT/SIGTERM 停止总预算缺省均为 `STOP_TOTAL=10s`。开发来源的 update、rollback、uninstall 返回 `FAILED_PRECONDITION`；WebUI/CLI 保存开发配置后写回并重启，手改配置的轮询行为见[开发运行](#开发运行与热部署)。

库模式公开入口由根目录 `alonabot_test.go` 的 `TestLibraryLifecycle` 验证；装配、内部服务、候选与 WebUI 的 9 个集成测试文件位于 `internal/app/`。开发 CLI 与热部署测试位于 `tests/e2e/`，覆盖经 `internal/app` 运行的独立进程开发组件与单组件重启。

## 调试运行中的 Adapter

先列出实例，再读取该实例当前支持的操作与 Schema，最后提交调用：

```fish
alonabot adapter instance list --data-dir ./data
alonabot adapter actions main --data-dir ./data
alonabot adapter actions main sdk/message.send --data-dir ./data
alonabot adapter call main sdk/message.send --version 1 --to '{"chatId":"private:alice"}' --params '{"chain":{"segments":[{"text":{"text":"hello"}}]}}' --data-dir ./data
```

`--params-file params.json` 从文件读取完整参数，与 `--params` 互斥。`--to` 使用 Target 的 protobuf JSON，可提供 `chatId` 和 `threadId`；Core 按实例绑定 Adapter 与账号。命令输出 JSON；业务失败以非零状态退出，stdout 仍保留 `payload`、`error` 与 `trace_id`，可检查部分成功的回执。管理连接沿用 `--data-dir`、`--admin-url` 与 `admin_token`。

SDK 在持有真实 Instance 的进程里识别 Send、Editor、Deleter、MediaSource、NodesReader、CommandSynchronizer、CommandMenuController 和 ActionExecutor，并结合实例能力提供目录。标准名称为 `sdk/message.send`、`sdk/message.edit`、`sdk/message.delete`、`sdk/media.read`、`sdk/message.read_nodes`、`sdk/commands.sync`、`sdk/command.menu`；平台动作沿用 Descriptor.Actions 的名称、版本、scope 与参数、结果 Schema。参数采用目录展示的 protobuf JSON；平台动作直接接收自身 JSON 参数。停止、重启和实例替换后重新查询目录。

公共客户端 `sdk/debug.Client{URL, Token, Host, HTTP}` 提供 `Instances(ctx)`、`Operations(ctx, instance, name)` 与 `Invoke(ctx, debug.Call{Instance, Name, Version, To, Params})`。`Invoke` 遇到业务错误时同时返回完整 Result 和 error；请求 context 传递取消，释放客户端保持 Core 运行。

目录端点为 `GET /api/v1/debug/adapters/{instance}/operations`，调用端点为 `POST /api/v1/debug/adapters/call`。两者使用现有管理认证。Send/Edit/Delete、节点和平台动作复用 Core 的业务准入；发送经过 Hook 与媒体授权，菜单沿用实例 worker 和 SDK 菜单锁。媒体读取使用 `fileId`、`offset`、`length`，其中 `fileId` 填本次运行实例已由 Core 登记的入站对象 ID，例如 `sdk/message.read_nodes` 结果中的 `incomingRef.token`。响应 `data` 为 base64。注入事件不接受入站平台文件引用。实际平台操作仍会产生真实副作用。

验证入口为 `go test -race ./internal/core/adapter/remote -run '^TestDebugAdapter$'`，通过真实 SDK 子进程与管理认证验证目录、调用、部分成功、媒体分块和实例替换。

## Adapter 的 HTTP 路由

Adapter 在 `adapter.Descriptor.HTTP` 中用 `[]adapter.HTTPRoute` 声明路由，对外位于统一端口的 `/plugins/<Adapter 包 ID>/...` 下声明路由，例如 OneBot 11 反向 WebSocket。路由规则与插件相同：私有路由需要 Core 会话，非 GET/HEAD 的 Cookie 请求还要验证 CSRF 与 Origin。`Auth` 的零值要求 Core 会话；公开路由写 `Auth: adapter.AuthPlatformToken`：Core 代理保留入站的 `Authorization` 交给 Adapter，由 Adapter 校验平台凭证，并给出错误时的平台响应。插件路由的 `Authorization` 一律被清除。两者都会清除入站的 Cookie、CSRF、Forwarded 与 X-Forwarded-*，并注入内部代理凭证；SDK 先恒定时间比较该凭证，比较后移除，再进入用户 Handler。WebSocket Upgrade 由代理双向转发，连接寿命不受普通请求期限限制，路由撤销或 Session 失效时由 Core 关闭。

## 动态提醒与静态任务暂停恢复

需要十分钟后给当前会话发提醒时，先在 `New()` 中登记任务声明和处理函数，命令执行时再创建实例。完整示例见 [showcase](../examples/plugins/showcase/main.go)：`/demo remind meeting 600 开会` 创建当前用户在当前会话内的 `meeting` 提醒，`/demo cancel meeting` 取消。同用户、同会话、同 key 再次创建会替换提醒；不同用户或会话各自独立，回复始终使用用户输入的 key。

```go
type ReminderArgs struct {
    Session string `json:"session"`
    Text    string `json:"text"`
}

// 创建与取消共用编码，将可信身份与用户 key 一起放入内部 key。
func reminderKey(c *sdk.Context, key string) (string, error) {
    user, session := c.User(), c.Session()
    internalKey := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, key)
    if len(internalKey) > 128 {
        return "", errcode.InvalidArgument.New("reminder key with user and session scope exceeds 128 bytes")
    }
    return internalKey, nil
}

// New 中创建并登记一次，命令闭包持有同一个声明值。
reminder := task.Define[ReminderArgs]("remind")
task.On(p, reminder).Handle(func(c *sdk.Context, args ReminderArgs) error {
    _, err := c.SendTo(args.Session, message.Text(args.Text))
    return err
})
p.Command("remind").Handle(func(c *sdk.Context, key, text string) (message.Content, error) {
    internalKey, err := reminderKey(c, key)
    if err != nil {
        return nil, err
    }
    _, err = reminder.At(c, time.Now().Add(10*time.Minute),
        ReminderArgs{Session: c.Session(), Text: text}, task.Key(internalKey))
    if err != nil {
        return nil, err
    }
    return message.Text(key), nil
})
p.Command("cancel").Handle(func(c *sdk.Context, key string) error {
    internalKey, err := reminderKey(c, key)
    if err != nil {
        return err
    }
    return reminder.Cancel(c, internalKey)
})
```

`reminderKey` 使用 `c.User()` 与 `c.Session()` 的可信身份，以长度前缀区分完整字段。128 字节上限包含身份作用域和用户 key；超长返回明确的 `INVALID_ARGUMENT`，保留输入原文。示例需导入标准库 `fmt`、`sdk/errcode` 与 `sdk/task`。

任务回调没有 Actor 与默认会话；将所需会话、用户或取消键写入参数，并用 `c.SendTo(sid, content)` 指定发送目标。主动发送遵循 Core 的 `allowed_adapter_instances` 准入。参数类型遵循 JSON Service 的类型规则，使用包级具名、非泛型的值类型，顶层编码为 JSON 对象；生成器检查嵌套字段与自定义编码限制。新增声明后直接运行测试，或运行 `alonabot build` 构建组件。

`reminder.Every(ctx, time.Hour, args, task.Key(internalKey))` 创建周期实例；`reminder.Cron(ctx, "0 9 * * *", args, task.Key(internalKey), task.Timezone("Asia/Shanghai"))` 按 IANA 时区运行，其中 `internalKey` 沿用上述身份作用域编码。`At`、`Every` 上设置 `Timezone` 返回 `INVALID_ARGUMENT`。省略 `Key` 时 SDK 生成随机 key，成功返回实际使用的 key；同任务名、同 key 再次创建会整体替换规则和参数，在途旧一轮继续执行。`Cancel` 删除实例并协作取消在途调用，函数实际返回后释放运行槽。未知实例保留 Core 的 `NOT_FOUND`。

`reminder.List(ctx)` 读取全部分页，返回完整的 `task.Instance[ReminderArgs]`。`At`、`Interval`、`Cron` 中仅对应触发规则的字段有值，`Timezone` 随 Cron 返回；`NextAt == nil` 表示协议没有下一次触发时刻。任一页读取或参数解码失败都会返回错误。

所有任务客户端函数在 Activated 后可用；`OnStart`、`OnReady`、停止后与已失效的 Handler Context 返回 `FAILED_PRECONDITION`。在 Handler 中传 `c` 以保留父调用关联；独立后台 context 只有组件身份。Core 保存任务定义与 interval 起点，重启后安排下一未来触发；离线、未激活与暂停期间错过的触发作废，一次性过期实例删除。

静态任务继续在 `New()` 中声明，例如 `p.Every("tick", time.Minute).Handle(tick)`。在已激活的 Handler 或后台工作中暂停、恢复与查询：

```go
if err := task.Pause(ctx, "tick"); err != nil {
    return err
}
state, err := task.Status(ctx, "tick")
if err != nil {
    return err
}
if state.LastAt != nil {
    slog.Info("last tick", "at", *state.LastAt)
}
return task.Resume(ctx, "tick")
```

`task.State.LastAt` 记录本次 Core 运行最后一次实际派发，首次派发前和 Core 重启后为 `nil`；暂停时 `Paused` 为 true、`NextAt` 为 `nil`。暂停保留在途函数继续执行，恢复沿保存的 interval 起点安排未来触发。暂停状态跨重启保存，重复暂停或恢复成功，未登记的 job ID 返回 `NOT_FOUND`。

## 插件测试

在插件包的 `main_test.go` 中直接调用 `New()`，用 `plugintest.Start` 运行真实 Core、SDK Session 与测试 Adapter。例如 lifecycle 的 `/ping`：

```go
package main

import (
    "testing"

    "github.com/AlonaBot/AlonaBot/plugintest"
)

func TestPing(t *testing.T) {
    p := New()
    h := plugintest.Start(t, plugintest.Plugin(p))
    alice := plugintest.User{ID: "alice"}
    result := h.Text(plugintest.Private(alice.ID), alice, "/ping")
    result.RequireReply(t, "pong")
}
```

`RequireReply(t, want)` 先检查每个 Handler Outcome 的错误，再要求恰好一次发送且 `Sent.Text()` 为 want；失败时报告插件与路由身份或实际发送。回复同时包含图片等片段时，可继续检查 `result.Sends[0].Chain`。

直接运行 `go test ./...`。plugintest 按被测插件的入口包自动扫描工程身份与资源，沿用测试二进制的 GOOS、GOARCH 与 build tags；同一入口包在一个进程中只扫描一次。每次 Start 创建独立的临时目录、配置和 `dev.sqlite`，在测试清理时停止组件；Submit 支持同一 Harness 的并行注入，结果按 trace 归属。

`Message(chat, user, message.Chain{...})` 可以注入混合片段，`Sent.Text()` 按顺序提取文本。`Sent.To`、`Chain`、`Result` 保留目标、消息链与回执；`Op` 区分 send/edit，`Edited` 保存编辑句柄。`Event` 注入通过 `Adapter` 选项声明的扩展事件。

**权限与配置。** `Admin("root")` 登记 Core Admin；`User.Permissions` 携带该会话的平台权限，需通过 `Adapter` 声明它们。例如：

```go
moderator := plugintest.User{
    ID: "mod",
    Permissions: []sdk.Permission{sdk.PermissionName("plugintest.moderator")},
}
h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Admin("root"),
    plugintest.Adapter(plugintest.AdapterID, adapter.PermissionDeclaration{
        Name: "plugintest.moderator", Implies: []string{"chat_admin"}, LabelKey: "moderator",
    }))
result := h.Text(plugintest.Group("room"), moderator, "/mod")
```

其中 `sdk` 与 `adapter` 分别来自 `sdk`、`sdk/adapter`。配置按被测插件的 Schema 提供：

```go
h := plugintest.Start(t, plugintest.Plugin(p),
    plugintest.Config(p, `{"greeting":"Hello"}`),
    plugintest.AllowedInstances(p, plugintest.InstanceID),
    plugintest.Capabilities("message.send", "message.edit"))
```

`AllowedInstances` 缺省允许全部实例，`AllowedInstances(p)` 收紧为空集合，job 和后台的显式发送会得到 `PERMISSION_DENIED`；可信入站事件的原目标回复沿用原有豁免。`Capabilities` 完整替换缺省能力，上例可验证按钮、相册等片段的 `UNSUPPORTED` 错误及插件降级行为。拒绝的发送在 Outcome 中保留错误。

**按钮。** 从发送结果点击动作名，用户取本次点击者；Core 恢复 token 中的 action、版本与业务值，并逐次鉴权：

```go
menu := h.Text(plugintest.Group("room"), alice, "/menu")
clicked := h.Click(plugintest.User{ID: "bob"}, menu.Sends[0], "first")
```

缺失的动作、URL 按钮及缺少已发送回执属于测试输入错误，Click 通过 `t.Fatal` 报告。过期 token、失效代次和权限拒绝会返回空发送及空 Handler Outcome。

**静态与动态任务。** `h.Trigger(p, "tick")` 立即执行已登记静态任务；`h.TriggerTask(p, "remind", key)` 定位已创建实例，使用 Core 保存的参数。两者通过 Scheduler 的正常 REQUEST/RESPONSE、job scope、运行槽和期限执行，Handler 使用 `c.SendTo(sid, content)` 指定发送目标。准入及定位错误使测试失败；已派发 Handler 的错误保留在 `Result.Outcomes`。

showcase 的提醒先通过命令创建，再用含用户与会话的完整内部 key 触发：

```go
h.Text(chat, alice, "/demo remind lunch 3600 soup")
user := message.UID(plugintest.AdapterID, alice.ID)
session := message.SID(plugintest.AdapterID, "private:alice", "")
key := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, "lunch")
reminder := h.TriggerTask(p, "remind", key)
```

完整示例见 [showcase/main_test.go](../examples/plugins/showcase/main_test.go)。一次性实例在实际派发时消费；周期实例保留起点并继续安排未来 tick。静态任务暂停或运行中、动态实例运行中或排队时，显式触发返回 `FAILED_PRECONDITION`；动态全插件并发名额已满时沿用到期队列。

**多轮等待。** 进入 `core.wait` 后注入立即返回 `Result.Waiting=true`，进入等待的路由 `Outcome.Waiting=true`；问题发送仍在本次 `Sends` 中。下一条同用户、同会话消息由普通入口注入：

```go
question := h.Text(chat, alice, "/ask")
if !question.Waiting {
    t.Fatalf("ask = %+v", question)
}
answer := h.Text(chat, alice, "answer")
```

命中等待后的发送和最终错误归入新注入，直到延续调用结束或再次等待；其他用户或会话继续正常分发。注入返回由 Core 的完成或等待通知决定。

**后台发送。** `h.Background()` 按 Adapter 完成顺序消费未归属事件或显式触发任务的发送和编辑。事件与显式任务的发送随各自的 `Result.Sends` 交付，每条记录消费一次。测试用自己的 channel 控制后台工作，并等待 Send 返回后读取：

```go
send, done := make(chan struct{}), make(chan error, 1)
var cancel context.CancelFunc
p.OnReady(func(ready context.Context) error {
    ctx, stop := context.WithCancel(context.WithoutCancel(ready))
    cancel = stop
    go func() {
        select {
        case <-send:
        case <-ctx.Done():
            return
        }
        _, err := sdk.Send(ctx, "plugintest:private:alice", message.Text("background"))
        done <- err
    }()
    return nil
})
p.OnStop(func(context.Context) error { cancel(); return nil })
h := plugintest.Start(t, plugintest.Plugin(p))
send <- struct{}{}
if err := <-done; err != nil {
    t.Fatal(err)
}
records := h.Background()
```

`OnReady` 的回调 context 在返回时取消，上例保留其 Session 归属，后台生命周期由 OnStop 管理。channel 确定发送时刻，Core 通知确定注入终点。

仓库完整示例覆盖 echo、lifecycle、showcase、configdb 与 media。configdb 的 KV 与文件命令见[插件指南](plugin-guide.zh-CN.md#保存少量状态)：

```sh
go test -race -count=2 ./plugintest/... ./examples/plugins/...
```

**LLM 替身。** 测试调用文本生成的插件时，用 `LLM(handler)` 提供结果。插件仍经 `@agent/llm.generate` 调用同一 Core 中的测试 Agent。先在插件的 `New()` 返回前登记命令（导入 `sdk/llm` 与 `sdk/message`）：

```go
p.Command("llm").Handle(func(c *sdk.Context, prompt string) (message.Content, error) {
    response, err := llm.Generate(c, llm.Request{Messages: []llm.Message{{Role: "user", Text: prompt}}})
    if err != nil {
        return nil, err
    }
    return message.Text(response.Text), nil
})
```

在同包的 `main_test.go` 中加入，plugintest 会自动扫描调用方的 `llm.Request`、`llm.Response` Schema：

```go
package main

import (
    "context"
    "testing"

    "github.com/AlonaBot/AlonaBot/plugintest"
    "github.com/AlonaBot/AlonaBot/sdk/errcode"
    "github.com/AlonaBot/AlonaBot/sdk/llm"
)

func TestLLMReply(t *testing.T) {
    p := New()
    h := plugintest.Start(t, plugintest.Plugin(p),
        plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error) {
            if len(req.Messages) != 1 || req.Messages[0].Role != "user" || req.Messages[0].Text != "hello" {
                return llm.Response{}, errcode.InvalidArgument.New("unexpected prompt")
            }
            return llm.Response{
                Text: "fixed answer", Model: "test-model",
                InputTokens: nil, OutputTokens: nil,
            }, nil
        }))
    alice := plugintest.User{ID: "alice"}
    result := h.Text(plugintest.Private(alice.ID), alice, "/llm hello")
    result.RequireReply(t, "fixed answer")
}
```

上例的 nil 用量经 Service 编码后仍为 nil；已知用量为零时，返回指向 `0` 的 `*int`。验证失败路径时，将 handler 换成：

```go
plugintest.LLM(func(context.Context, llm.Request) (llm.Response, error) {
    return llm.Response{}, errcode.Unavailable.New("synthetic generation failure")
})
```

同一条命令会得到空 `result.Sends`，`result.Outcomes[0].Err` 保留 `errcode.Unavailable` 与 `Message`。未设置 LLM 时，`@agent` 调用同样返回 `UNAVAILABLE`。多次设置 LLM 时最后一个函数生效；`LLM(nil)` 在 Start 启动前使测试失败。handler 接收关联的 context，并应响应其取消与期限。Schema 负责结构校验，业务规则与错误由 handler 决定。

仓库的完整集成测试见 [plugintest/llm_test.go](../plugintest/llm_test.go)。运行：

```sh
go test -race -count=2 ./plugintest/...
```

### 媒体、动作与完整记录

在插件中用公共 `sdk/media` 读取入站内容，再上传或转发；下面的命令可放进现有 `New()`（增加 `bytes`、`io`、`sdk/media` 导入）：

```go
p.FullMatch("copy").Handle(func(c *sdk.Context) error {
    source := c.Event().Payload.(*message.CreatedPayload).Chain[1].(message.FileSegment).Source
    reader, err := media.Open(c, source)
    if err != nil { return err }
    data, readErr := io.ReadAll(reader)
    closeErr := reader.Close()
    if readErr != nil { return readErr }
    if closeErr != nil { return closeErr }
    ref, err := media.Upload(c, bytes.NewReader(data), "image/png")
    if err != nil { return err }
    _, err = c.Send(message.Chain{message.Text("copied"), message.Image(ref.Source()), message.File(ref.Source(), "copy.png")})
    return err
})
```

测试中生成一个小图片，注入其平台文件 fixture，再检查 Adapter 实际取得的字节与消息顺序（增加 `image`、`image/png`）：

```go
var pixel bytes.Buffer
if err := png.Encode(&pixel, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil { t.Fatal(err) }
h := plugintest.Start(t, plugintest.Plugin(New()))
source := h.Fixture(plugintest.InstanceID, "pixel", plugintest.Media{Data: pixel.Bytes(), MIME: "image/png"})
alice := plugintest.User{ID: "alice"}
result := h.Message(plugintest.Private(alice.ID), alice,
    message.Chain{message.Text("copy"), message.File(source, "input.png")})
if len(result.Sends) != 1 { t.Fatalf("sends = %+v", result.Sends) }
sent := result.Sends[0]
if len(sent.Chain) != 3 || len(sent.Media) != 2 { t.Fatalf("send = %+v", sent) }
for _, output := range sent.Media {
    _, format, err := output.Image()
    if err != nil || format != "png" || output.MIME != "image/png" || !bytes.Equal(output.Data, pixel.Bytes()) {
        t.Fatalf("media = %+v, format = %s, error = %v", output, format, err)
    }
}
if sent.Media[1].Filename != "copy.png" { t.Fatalf("filename = %q", sent.Media[1].Filename) }
```

fixture 由真实 Adapter 媒体 RPC 取件；插件 `media.Open`、`media.Upload` 和出站 Adapter `ReadMedia` 使用同一 Core 的授权、存储和令牌校验。`Sent.Media` 按消息链遍历顺序保存实际内容；HTTPS 来源保留 URL，测试中的字节断言使用 fixture 或上传的受管媒体。

`Operations(Behavior{...})` 设置缺省实例的 Send/Edit/Action/Delete/Nodes/Menu 回调；多账号分别放在 `InstanceConfig.Behavior`。Send/Edit 收到的 `Call.SendResult` 已含缺省成功回执，可调整整体状态、逐段状态及错误后返回：

```go
behavior := plugintest.Behavior{
    Send: func(ctx context.Context, call plugintest.Call) (*message.SendResult, *errcode.Error) {
        receipt := call.SendResult
        return receipt, errcode.Unavailable.New("platform accepted the message with a warning")
    },
    Action: func(ctx context.Context, call plugintest.Call) (json.RawMessage, *errcode.Error) {
        return call.Params, errcode.Unavailable.New("platform returned a warning")
    },
}
h := plugintest.Start(t, plugintest.Plugin(New()), plugintest.Operations(behavior))
```

Send/Edit 可返回 Accepted、Rejected、Partial、Unknown 及结果和错误同时存在的任意合法组合。平台 Action 的参数、结果遵循 `Adapter(..., adapter.Action[P,R](...))` 或实例 Descriptor 的原声明；上述 Action 回显适用于参数和结果 Schema 相同的动作。已声明 Action 未配置回调时返回 `UNSUPPORTED`。Delete 缺省逐项接受，Nodes 缺省返回请求中的条目；平台 ID 查询可用 Nodes 回调返回展开节点。菜单缺省按 scope/language 保存 set/get/delete 状态，支持传统 SyncCommands。

`h.Calls()` 保留全部已完成的 Adapter 操作；`h.WaitCalls(ctx, count)` 等待数量。每条 Call 含操作、Adapter/实例/账号、目标、参数、实际结果、错误和 TraceID。一次事件的 `Result.Calls` 与 `Sends` 按 trace 归属。`Result.Outcomes` 保存实际调用 Handler 的结果；调用前的参数解析与权限等准入失败由 `Diagnostics` 保存，包含 Phase、PluginID、RouteID、TraceID 和 Err。`RequireReply` 继续检查 Handler 结果和回复，可正常断言参数错误的 usage 回复；真实 Handler 返回的 `INVALID_ARGUMENT` 仍在 Outcomes 中。媒体、菜单或动作直接联调也可用 `h.DebugClient().Invoke`，标准名称见上文的 Adapter 目录。

### 多账号、并行事件与可推进时钟

```go
p := New()
h := plugintest.Start(t, plugintest.Plugin(p),
    plugintest.Instance(plugintest.InstanceConfig{AdapterID: "first", InstanceID: "one", AccountID: "bot1"}),
    plugintest.Instance(plugintest.InstanceConfig{AdapterID: "first", InstanceID: "two", AccountID: "bot2"}),
    plugintest.Instance(plugintest.InstanceConfig{AdapterID: "second", InstanceID: "three", AccountID: "bot3"}),
    plugintest.DefaultInstance("first", "one"),
    plugintest.DefaultInstance("second", "three"),
    plugintest.AllowedInstances(p, "one", "three"))
alice := plugintest.User{ID: "alice"}
first := h.Submit(plugintest.Private(alice.ID).OnInstance("one"), alice, message.Text("/ping"))
second := h.Submit(plugintest.Private("bob").OnInstance("two"), plugintest.User{ID: "bob"}, message.Text("/ping"))
for _, pending := range []*plugintest.Pending{first, second} {
    result, err := pending.Wait(t.Context())
    if err != nil { t.Fatal(err) }
    result.RequireReply(t, "pong")
}
```

显式 Instance 列表替换缺省测试实例；入站用 `OnInstance` 选择账号，`DefaultInstance` 选择显式发信的默认账号，`AdminOn(adapterID, userID)` 指定平台管理员。`Descriptor.Capabilities` 提供平台声明全集，实例 `Capabilities` 选择实际启用的名称；nil 沿用 Harness 的能力选项，显式空列表表示无可选能力。同平台不同账号可共用 Descriptor 并分别收紧能力。未配置 Instance 时保留原单实例选项。`Submit` 立即提交，`Wait(ctx)` 可取消等待并再次调用；事件继续执行，最终 `TraceIDs`、Sends 和 Calls 关联自己的事件。`WaitBackground(ctx, count)` 等待并消费后台输出，超时等待结束后的回复也可从后台读取。

```go
clock := plugintest.NewClock(time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC))
t.Cleanup(clock.Close)
p := New()
p.Every("testtick", time.Second).Handle(func(c *sdk.Context) error {
    _, err := c.SendTo("plugintest:private:alice", message.Text("tick"))
    return err
})
h := plugintest.Start(t, plugintest.Plugin(p), plugintest.WithClock(clock))
clock.Advance(time.Second)
records, err := h.WaitBackground(t.Context(), 1)
if err != nil { t.Fatal(err) }
if records[0].Text() != "tick" { t.Fatalf("records = %+v", records) }
```

Scheduler 的自然定时任务、动态 `task.At/Every`、continuation 期限与 `c.Wait` 共用此时钟。测试创建绝对任务时用 `clock.Now()` 构造时间；进入 `Result.Waiting` 后再 Advance 可触发等待超时。Advance 交付到期定时器，异步 RPC 的完成使用 Wait/WaitBackground 等待。任务取消与 Harness.Close 会注销相应定时器；生产继续使用真实时钟，Session 协议期限使用真实绝对时间并接收逻辑时钟的取消。

### Connect 连接运行中的 Core

先在 Core 中启用待测插件和真实 Adapter 实例，准备本机 `admin_token`。下面的测试使用已有 `/ping` 命令；修改 URL 和实例身份以匹配运行配置（导入 `context`、`os`、`strings`、`sdk/debug`、`plugintest`）：

```go
func TestRunningCore(t *testing.T) {
    token, err := os.ReadFile("../data/admin_token")
    if err != nil { t.Fatal(err) }
    client := &debug.Client{URL: "http://127.0.0.1:7650", Token: strings.TrimSpace(string(token))}
    connection, err := plugintest.Connect(context.Background(), client,
        plugintest.InstanceConfig{AdapterID: "telegram", InstanceID: "main", AccountID: "123456"})
    if err != nil { t.Fatal(err) }
    defer connection.Close()
    alice := plugintest.User{ID: "987654"}
    result, err := connection.Text(t.Context(), plugintest.Private(alice.ID), alice, "/ping")
    if err != nil { t.Fatal(err) }
    result.RequireReply(t, "pong")
    if len(result.Outcomes) == 0 || len(result.TraceIDs) == 0 { t.Fatalf("dispatch = %+v", result) }
}
```

会话 ID 使用对应 Adapter 的实际格式；Telegram 的 `private:<userID>` 可通过上例构造。`Connection.Event(ctx, event)` 注入完整标准 SDK 事件；`Message` 发送混合链，`Submit` 配合 Pending.Wait 支持异步联调。`connection.Client.Operations/Invoke` 可发现并执行该真实 Adapter 的动作。

客户端经 `POST /api/v1/debug/events` 的管理认证注入标准 Event，Core 重新绑定来源身份、展开权限，执行正常命令匹配、参数解析与路由。结果保留实际 RouteID、TraceIDs、回复及同时存在的回执和错误；Outcome、Diagnostic 与回复各自携带所属 TraceID。调试事件及 Adapter 调用在管理元数据写锁和 GC 临界区外等待，独立会话可并行执行；元数据写入仍遵循原有锁与回收顺序。请求取消释放观察器，Close 只结束本客户端资源。既有 Core、Adapter 与插件继续运行。Core 的真实接收游标保持原值。

参数与权限诊断在本地 Harness、Connect 和 `sdk/debug.Client.Inject` 中使用相同字段。例如已有 `echo <text>` 命令缺少参数时：

```go
invalid, err := connection.Text(t.Context(), chat, alice, "/echo")
if err != nil { t.Fatal(err) }
invalid.RequireReply(t, "Usage: /echo <text>")
if len(invalid.Outcomes) != 0 || len(invalid.Diagnostics) != 1 { t.Fatalf("dispatch = %+v", invalid) }
diagnostic := invalid.Diagnostics[0]
if diagnostic.Phase != plugintest.PhaseArguments || errcode.Of(diagnostic.Err) != errcode.InvalidArgument {
    t.Fatalf("diagnostic = %+v", diagnostic)
}
```

示例需导入 `sdk/errcode`；`chat`、`alice` 为已配置实例的会话与用户。权限或生命周期准入失败的 Phase 为 `plugintest.PhaseAdmission`，并保留实际错误码；使用 `sdk/debug` 时对应常量为 `debug.PhaseArguments` 与 `debug.PhaseAdmission`。

同机集成测试 [plugintest/debug_test.go](../plugintest/debug_test.go) 使用真实 Core/Session、认证 HTTP 和 loopback 平台验证上述链路；真实平台联调需要相应账号凭据和实际聊天目标。

## 协议与会话行为

组件协议版本 **1.0** 与产品版本 **0.3.0** 独立。当前消息类型与字段号由 [contracts/proto/alonabot/v1](../contracts/proto/alonabot/v1) 定义，[wire 运行时](../internal/wire) 实现分帧与 Session 调用。修改 `.proto` 源文件后重新生成 Go 产物。保留已发布的字段号与枚举值；删除的名称和编号须 reserved，已有字段类型保持兼容。[协议编号检查](../contracts/wirespec_test.go) 保护这些标识。

TCP 与 stdio 均使用四字节无符号大端长度，后接 Protobuf Envelope。长度不含前缀，须为 1～4,194,304 字节。零长度、超限、截断或无法解码的帧使 Session 失败，读取端不猜测下一帧边界。每条连接只有一个 writer。Core 持有回环 TCP listener（`127.0.0.1:0`）直到结束。每次启动组件创建新 Session 与 32 字节 secret；认证恒时核对 secret，并核对身份、digest 和协议版本。stdio 的 stdin/stdout 传输协议帧，stderr 输出日志。Bootstrap 中的秘密、配置、数据库及 HTTP 凭据须留在日志之外。

激活顺序为 HELLO → HELLO_ACK/Bootstrap → REGISTER → PREPARED → ACTIVATE → OnReady → ACTIVATED，激活后才开始普通业务调用。协议 major 必须相同，协商 minor 取双方支持的较小值；required capabilities 必须双方支持，`http_listener` 版本为 1。注册与生命周期载荷见 [control.proto](../contracts/proto/alonabot/v1/control.proto) 和 [route.proto](../contracts/proto/alonabot/v1/route.proto)。

调用按 `(session_id, initiator_side, call_id)` 匹配。call ID 非零、按方向单调递增，在同一 Session 内不复用。父调用须为同一 Session 中仍在途的 Core 发起调用。deadline 为绝对 UTC 毫秒，子调用沿用或缩短父期限。取消为协作式，已完成的副作用保留。载荷与 ErrorInfo 可同时存在以表示部分结果，调用方须保留两者。请求默认超时为 30 秒。Session 失败时，挂起调用取得真实错误；认证无效时关闭连接，避免暴露凭据。

用 `alonabot tune list` 查看完整的当前调优表，包括名称、单位、默认值、生效值、约束与来源；`--args NAME=value` 使用同样的校验。[internal/tune/params.go](../internal/tune/params.go) 定义数值，[部署指南](deployment.zh-CN.md#插件休眠) 说明唤醒预算，[插件指南](plugin-guide.zh-CN.md) 说明延续、存储与任务。帧上限固定为 4 MiB。阶段预算默认 HELLO 10s、配置 60s、迁移 15m、OnStart 60s、OnReady 30s；停止总预算 10s，其中排空 6s、OnStop 4s。崩溃恢复窗口为五分钟，最多三次。组件自己的 Handler 与后台活动须响应取消。

### 菜单协议新增字段

Core→Adapter 的 `adapter.command.menu` 执行 set/get/delete；插件→Core 的 `core.command.menu` 执行 set/get/delete/status/sync，省略实例时取在途父调用的可信实例。两者使用 schema_version=1。`AdapterInstanceStartResponse.command_menu=5` 为可选消息，缺失时沿用 `adapter.commands.sync`；其 CommandMenuConfig 含 bool auto_sync_commands=1、clear_commands_on_disable=2。`CommandSpec.menu=14` 缺省 false；可信 `MediaSource.mime=6` 为 string。

共用消息为 `CommandMenuRequest{instance_id=1, operation=2, menu=3}`、`CommandMenu{scope=1, language_code=2, commands=3}`、`CommandMenuScope{type=1, chat_id=2, user_id=3}`、`CommandMenuEntry{name=1, description=2}`、`CommandMenuResponse{commands=1, config=2, ready=3, last_error=4, positions=5}`、`CommandMenuPosition{scope=1, language_code=2, mode=3, owner_plugin_id=4}`，commands/positions 均 repeated。scope 类型为 default/chat/chat_administrators/chat_member，插件只开放后三种，thread 不进入 scope。

`EventPayload.kind` 追加 command_menu_ready=13，string 字段为 adapter_id=1、adapter_instance_id=2、account_id=3、chat_id=4。Core 定向发送 `bot.command_menu_ready`，空 chat_id 表示实例手动就绪，非空表示聊天可用。Adapter 上报标准事件仍沿用原白名单。定义见 [Adapter 协议](../contracts/proto/alonabot/v1/adapter.proto)、[路由](../contracts/proto/alonabot/v1/route.proto)与[消息](../contracts/proto/alonabot/v1/message.proto)，生命周期与准入见[命令菜单](plugin-guide.zh-CN.md#命令菜单)。

## Agent、流式服务与生命周期 API

[完整参数、能力状态、八阶段钩子与 Goal 恢复契约](agent-api.zh-CN.md)。
