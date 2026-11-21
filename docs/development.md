# 源码开发

从源码运行一个插件，再按改动范围进入 Core、SDK、Adapter、Agent 或 WebUI。贡献流程见[贡献指南](../CONTRIBUTING.md)，模块与数据所有者见[架构地图](../ARCHITECTURE.md)。本文维护开发环境、命令和验证方法；部署操作见[安装与部署](deployment.md)。

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

在同一个终端输入 `/ping`，预期回复 `pong`；Ctrl-C 正常停止。要并行运行多个 Core，为每个实例选择独立数据目录与监听端口。首次管理认证和监听设置见[部署指南](deployment.md#core-设置与-http-入口)。

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

在组件工程执行 `alonabot check .`、`alonabot run .` 或 `alonabot pack .`。本仓库的 Core 与 SDK 保持独立 module，`components.lock.json` 固定整合构建使用的组件源码；跨组件测试位于独立的 `tests/go.mod`，模型 SDK 与 Adapter 实现依赖留在该测试 module 中。

```fish
uv run --no-project python scripts/integration.py
# 按改动选择测试，例如 Agent 服务：
uv run --no-project python scripts/integration.py --run '^TestAgentServices$' --race
```

整合入口先运行 `tests/agent` 与 `tests/e2e`；需要 Core 私有装配入口的 Agent 服务测试通过临时 modfile 和 overlay 运行，依赖和测试文件均不写回 Core module。CLI 的端到端测试在 `tests/e2e`，`cmd/alonabot` 只保留正式命令。

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

合并前 CI 的 Go 检查包括 `go vet ./...`、race 测试与五目标 pure-Go 构建；GitHub-hosted runner 串行测试包，命令以[工作流](../.github/workflows/ci.yml)为准。WebUI 独立仓的 CI 另执行前端类型检查与构建。手动需要完整检查时执行：

```fish
go build ./...
go vet ./...
go test -race ./...
```

改变插件页面套件时，运行 `go test -race -v ./sdk/page ./examples/plugins/showcase` 和 `node --check sdk/page/kit.js`，并在 WebUI 的 iframe 中检查关键交互、语言与主题。文档修改检查链接、锚点、代码块和实际命令。

跨平台修改同时核对真实运行与交叉构建。五个发行目标为 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64；构建结果与平台实跑结果分别记录，现有结果见[发布说明](../changelog/v0.1.0/zh-CN.md#发行内容)。真实 Telegram 或付费 Provider 验证使用授权的测试凭据，模拟服务在验证结果中明确注明。

仓库只提交集成测试及协议编号检查。单元测试、临时复现与局部前端测试使用 `*_local_test.go`、`*.local.test.ts`、`*.local.test.tsx` 或 `tmp/test/`；完整规则见[编码规范](coding-standards.md#测试提交规则)。PR 记录测试位置、命令、平台、输入条件、结果与未运行原因，以及提交测试新增行数。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml)在 main push、目标为 main 的 PR 和手动触发时运行。Repository job 检查空白与 main 的文件范围；Go job 检查格式、生成一致性、vet、race 和五平台构建。main 路径检查排除个人维护资料，产品中的技能功能和测试工具照常保留。

[PR metadata](../.github/workflows/pr-metadata.yml)验证非草稿 PR 的说明；triage 和 auto-assign 只处理 GitHub 元数据，不执行 PR 代码。代码检查使用只读权限，不接触生产 secret。分支保护另由维护者配置。


## 开发运行与热部署

先构建 CLI：`go build -o bin/alonabot ./cmd/alonabot`。在仓库根目录用以下三步得到第一条回复（使用 fish）：

1. 创建工程：`bin/alonabot init ./alonabot-plugin-hello`。
2. 运行：`bin/alonabot run ./alonabot-plugin-hello`。
3. 在同一个终端输入 `/hello`，看到 `Hello AlonaBot`；Ctrl-C 停止 Core。

已安装 CLI 时，可在工程目录直接运行 `alonabot run`。stdin 是终端且设置文件没有任何 Adapter 实例时，自动启用内置终端对话；`--console` 强制启用并接受管道输入，`--console=false` 关闭。终端用户 `console:local` 在本次运行中具有 Core Admin 权限；实例和管理员保持在内存中。回复中的媒体显示为 `[image]`、`[file: 名称]` 等占位，输入 `#1` 点击最近一条带按钮消息的第一个回调按钮，编辑显示为 `(edited)` 新段。stdin 结束后 Core 继续运行，直到收到停止信号。

有开发目录时缺省自动重载；修改源码后再输入 `/hello` 即可观察新回复。`--reload=false` 关闭重载。

### 不经终端：回环 Adapter 与 curl

需要通过 HTTP 注入消息时，使用仓库的回环示例：

```fish
go build -o bin/alonabot ./cmd/alonabot
set -gx DEVBOT_DATA (mktemp -d)
set -gx DEVBOT_OUTPUT (mktemp)
cp examples/devbot/core_config.example.json $DEVBOT_DATA/core_config.json
printf 'DEVBOT_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_DATA $DEVBOT_OUTPUT
bin/alonabot run examples/devbot/plugin examples/devbot/adapter --data-dir $DEVBOT_DATA --reload
```

另开终端，设置相同的 `DEVBOT_DATA`、`DEVBOT_OUTPUT`。首次启动将管理 token 以 0600 权限持久保存在 `$DEVBOT_DATA/admin_token`，启动只输出文件位置；管理 CLI 通过 `--data-dir $DEVBOT_DATA` 自动读取凭据。下面手工调用 HTTP API 时，先用 `bin/alonabot admin-token show --data-dir $DEVBOT_DATA` 显式读取 token，令 `token` 为该值后：

```fish
curl --fail-with-body -X POST -H "Authorization: Bearer $token" \
  'http://127.0.0.1:7650/plugins/alonabot-adapter-devbot/inject?instance=one&text=%2Fhello'
cat $DEVBOT_OUTPUT
```

注入返回 204，异步发送后文件内容为 `Hello from devbot!`。完整实例 JSON、目录结构、认证、源码／翻译／配置改动与编译失败恢复操作见 [devbot README](../examples/devbot/README.md)。样例把 `adapters.devloop.package` 绑定到 `alonabot-adapter-devbot`，实例在 `instances.one`，命令前缀为 `/`。两个开发组件沿用真实 Deliver／Send 流水线。

组件市场 ID 使用连字符前缀，例如 `alonabot-plugin-hello`；本地 ID 仍允许下划线，规则见 [CONTRACTS §9.1](../spec/CONTRACTS-v0.1.md#91-manifest-与构建产物)。

验证官方 help 工程时先检出独立的 `alonabot-plugin-help` 仓库，执行 `bin/alonabot check <help 工程目录>`。使用回环 Adapter 试用时，在 `run` 的组件目录列表追加该工程目录，再把注入请求的 `text` 改为 `%2Fhelp` 或 `%2Fwhoami`。

先构建 CLI（`go build -o bin/alonabot ./cmd/alonabot`），再创建与验证组件：

```fish
bin/alonabot init ./alonabot-plugin-hello
bin/alonabot build ./alonabot-plugin-hello
bin/alonabot check ./alonabot-plugin-hello
bin/alonabot test ./alonabot-plugin-hello
bin/alonabot pack ./alonabot-plugin-hello
```

`init`、`build`、`check`、`test`、`pack` 至多接受一个位置目录，省略时使用当前目录；旧的 `--dir` 会报未知参数。`init` 要求空目录，非空时在写入前报错。`--id` 缺省取目标目录绝对路径的目录名；例如 `alonabot-plugin-hello` 直接作为市场 ID，`my_plugin` 保持本地 ID 并输出缺市场前缀的警告。非法目录名会保留原字符并提示显式给出 `--id`。缺省角色为 plugin，`--role agent|adapter|webui` 创建对应工程；Adapter 的 `--adapter-id` 缺省等于最终组件 ID，可显式覆盖。插件与 Agent 导出 `New()`，包含一个直接回复 `Hello AlonaBot` 的英文示例命令与 schema；需要翻译时自行添加 i18n 目录；Adapter 提供实例 schema，WebUI 提供 `assets/index.html`。各角色生成的 Go 文件和 `plugin.yaml` 均无注释。模板忽略 `data/` 与 `bin/`，正式薄入口通过 overlay 提供。库模式宿主参考 [devbot](../examples/devbot/README.md)。

位于现有 Go module 中时复用该 module，保持其 go.mod；外部空目录的 module 缺省同为目录名，可用 `--module example.com/myplugin` 覆盖。发布版 CLI 用自身模块版本写 SDK require，并在生成的工程目录运行 `go mod tidy`，随后可直接 `run`。依赖解析失败时保留工程，命令以失败状态退出并显示手动补齐的命令。开发版 `(devel)`（含本地修改的 `+dirty` 构建，以及从仓库 checkout 构建时 Go 自动盖上的伪版本）会提示执行 `go get github.com/AlonaBot/AlonaBot@<version>`，由作者选定版本。发布前修改 `plugin.yaml` 中的示例 source。

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

`pack` 为每个 `--target` 分别扫描源码、检查资源并构建，输出 `<id>-<version>-<os>-<arch>.tar.gz`；静态 WebUI 只输出一个 `<id>-<version>-any.tar.gz`，带 Go 入口的 WebUI 按平台构建。全部归档共享一份列出本次全部资产的 `release.json`，`--out` 默认同 build。资产地址为 `--asset-base` 加 `/` 与归档名；省略时，`source: github.com/<owner>/<repo>` 推出 `https://github.com/<owner>/<repo>/releases/download/v<version>`，其他来源须显式给出绝对 https 前缀，前缀保留路径且无 query/fragment。进程包的 `build-info` 是 `debug.BuildInfo.String()` 文本；生成代码、资源旁车与归档共用扫描时校验的资源快照，配置 schema 使用同一规范内容；扫描后保存的工程资源留待下次构建。

配置字段的 `example` 引用图片时，`check`、`build`、`pack` 检查两种配置 schema 中的引用及图片签名；被引用的图片随资源快照进入构建目录和归档。示例写法见 [插件开发规范](plugin-guide.md#配置字段的回复示例)。

```fish
bin/alonabot pack ./alonabot-plugin-hello --target linux/amd64 --target linux/arm64 \
  --asset-base https://example.com/releases/v0.1.0 --key ./release.key --out ./dist
```

`--key` 可选，文件内容为标准 base64 编码的 64 字节 Ed25519 私钥，允许首尾空白。提供时只对最终合并清单的原始字节签名一次，额外写 `release.json.sig` 和 `release.pub`，分别为标准 base64 签名及公钥；省略时生成 unsigned release，并清除输出目录中旧的签名与公钥。密钥读取或签名失败直接报错。私钥、运行配置与数据目录均应留在打包资源之外。


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

库模式公开入口由根目录 `alonabot_test.go` 的 `TestLibraryLifecycle` 验证；装配、内部服务、候选与 WebUI 的 9 个集成测试文件位于 `internal/app/`。开发 CLI 与热部署测试位于 `cmd/alonabot/`，覆盖经 `internal/app` 运行的独立进程开发组件与单组件重启。

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

直接运行 `go test ./...`。plugintest 按被测插件的入口包自动扫描工程身份与资源，沿用测试二进制的 GOOS、GOARCH 与 build tags；同一入口包在一个进程中只扫描一次。每次 Start 创建独立的临时目录、配置和 `dev.sqlite`，在测试清理时停止组件；同一 Harness 的注入与读取按顺序调用。

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

**后台发送。** `h.Background()` 按 Adapter 接受顺序返回并清空尚未交付的发送和编辑。注入开始到完成通知之间的发送随 `Result.Sends` 交付，区间外的留给 Background，每条记录消费一次。测试用自己的 channel 控制后台工作，并等待 Send 返回后读取：

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

仓库完整示例覆盖 echo、lifecycle、showcase、configdb 与 media。configdb 的 KV 与文件命令见[插件指南](plugin-guide.md#保存少量状态)：

```sh
go test -race -count=2 ./plugintest/... ./examples/plugins/...
```

**LLM 替身。** 测试调用文本生成的插件时，用 `LLM(handler)` 提供结果。插件仍经 `@agent/llm.generate` 调用同一 Core 中的测试 Agent。先在插件的 `New()` 返回前登记命令（导入 `sdk/llm` 与 `sdk/message`）：

```go
p.Command("llm").Handle(func(c *sdk.Context, prompt string) (message.Content, error) {
    text, err := llm.Ask(c, prompt)
    if err != nil {
        return nil, err
    }
    return message.Text(text), nil
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
                Text: "fixed answer", Provider: "synthetic", Model: "test-model",
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
