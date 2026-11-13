# 开发与验证

负责范围：环境准备、日常命令、外部测试环境、规则与检查的对照。
更新时机：工具链或依赖版本变化、新增或调整检查入口、测试环境变化时。

## 环境

| 工具 | 版本与来源 | 备注 |
| --- | --- | --- |
| Go | 1.27.1（asdf，`~/.tool-versions`） | `go.mod` 的 go 指令固定为该版本 |
| protoc | 36.0（`protoc --version` 输出 `libprotoc 36.0`；生成文件头记为 v7.36.0） | 版本不符时 `contracts/internal/protocgen` 拒绝生成 |
| protoc-gen-go | v1.36.12（`go.mod` 的 `tool` 指令，与 `google.golang.org/protobuf` 同版本） | 通过 `go tool -n protoc-gen-go` 定位，不全局安装 |
| Node（前端） | 26.5.1（asdf） | 供 web/admin 使用：在 `web/admin` 下运行 `npm ci`、`npm run typecheck`、`npm run build`，`npm run dev` 打开用 `fixtures/` 渲染表单的开发预览页；CI 不跑前端，前端测试只有被忽略的 `*.local.test.ts`，没有本地测试时 `npm run test` 也通过 |

已锁定并经测试验证的版本：protobuf v1.36.12、go.yaml.in/yaml/v3 v3.0.5、gorm v1.31.2、libtnb/sqlite v1.2.2、modernc.org/sqlite v1.61.0（已验证 CRUD/Tx/WAL Backup/迁移账本）；golang.org/x/sys v0.49.0 为直接依赖，用于 internal/oslock 的 flock/LockFileEx，以及 internal/core/process 在 Darwin 上查询进程组成员；Cobra v1.10.2、pflag v1.0.9、仅 Windows 链接的 mousetrap v1.1.0。

已批准的 Go 依赖（2026-10-21）：`google.golang.org/protobuf`、`gorm.io/gorm`、`github.com/libtnb/sqlite`、`modernc.org/sqlite`、`golang.org/x/sys`、`go.yaml.in/yaml/v3`（`gorm.io/driver/postgres` 的批准于 2026-10-27 随 D-48 撤销）；2026-10-25 批准 `github.com/modelcontextprotocol/go-sdk`，只用于官方 Agent 组件（`agent/`）的 MCP 客户端，版本在 T-12.3 验证后锁定；2026-10-26 批准 `github.com/spf13/cobra`（随之引入 `github.com/spf13/pflag`），只用于 `cmd/alonabot` 的命令树与全局参数，T-02.2 已锁定 Cobra v1.10.2、pflag v1.0.9 与仅 Windows 链接的 mousetrap v1.1.0；2026-10-26 批准 `github.com/google/jsonschema-go`（MCP go-sdk 也直接依赖它），用于 JSON Schema 的表示、序列化、`Resolve`、`ApplyDefaults` 与 `Validate`：配置 Schema（`sdk/config`、构建期生成器、Core 的配置与表单 Schema）、Service 与工具的输入 Schema、Adapter 声明的扩展事件、extension 与平台动作的 Schema（`sdk/adapter` 用 `For` 从 Go 类型推导，D-113）、`agent/`。配置 Schema 由生成器构造，不用 `For`/`ForType` 推导（D-43）；版本在首个使用它的切片引入时锁定。2026-10-27 批准 `github.com/coder/websocket`，只用于官方 OneBot 11 Adapter（`adapters/onebot11/`）的 WebSocket 连接，T-15.1 已锁定 v1.8.15（D-46）。2026-11-10 批准 `github.com/openai/openai-go/v3` 与 `github.com/anthropics/anthropic-sdk-go/v2`，只用于官方 Agent 组件的 Provider 客户端（`agent/provider`），不导入两者的 Azure、Bedrock、Vertex 等云厂商子包；T-12.12 锁定 openai-go v3.98.0、anthropic-sdk-go v2.9.0。实际链接的传递依赖为 bahlo/generic-list-go v0.2.0、buger/jsonparser v1.1.2、coder/websocket v1.8.15、go-logr/logr v1.4.2、go-logr/stdr v1.2.2、invopop/jsonschema v0.14.0、pb33f/ordered-map/v2 v2.3.1、standard-webhooks/standard-webhooks/libraries v0.0.1、tidwall/gjson v1.20.0、tidwall/match v1.1.1、tidwall/pretty v1.2.1、tidwall/sjson v1.2.5、go.opentelemetry.io/auto/sdk v1.1.0、go.opentelemetry.io/otel（含 metric、trace）v1.33.0、go.yaml.in/yaml/v4 v4.0.0-rc.2、golang.org/x/sync v0.24.0。MVS 将 x/text 从 v0.39.0 抬至 v0.42.0；coder/websocket 保持 v1.8.15；MCP go-sdk 的模块下限为 v1.3.1，当前 Provider 未链接 MCP 包。前端已批准：Vite（含 `@vitejs/plugin-react`）、TypeScript、React 19、vitest；2026-11-05 批准开发依赖 `@types/react@19`、`@types/react-dom@19`（PR #91），只用于类型检查。T-08.7 锁定 react 与 react-dom 19.3.1、@types/react 与 @types/react-dom 19.3.0、@vitejs/plugin-react 6.2.1、TypeScript 7.0.2、Vite 8.3.5、vitest 5.1.0；DOM 测试环境（jsdom、happy-dom）与 Testing Library 未批准。新增清单外依赖需要先征得维护者同意。

Core 可执行文件（`./cmd/alonabot`）与 SDK 根包（`./sdk`）只能链接上述清单的一个子集（D-52，D-68 修订），由审查时用下方的 `go list -deps` 命令核对：Core 允许 protobuf、modernc SQLite 及其依赖、GORM 与 libtnb/sqlite 及其依赖（jinzhu/inflection、jinzhu/now、golang.org/x/text）、yaml、x/sys、cobra 与 pflag、仅 Windows 链接的 mousetrap、jsonschema-go；SDK 根包只允许 protobuf、x/sys 与 jsonschema-go。MCP go-sdk、coder/websocket 只出现在官方插件、Agent 或 Adapter 中，openai-go 与 anthropic-sdk-go 只出现在 Agent 中；插件只在导入 `sdk/gormdb` 时链接 GORM。修改子集与修改已批准清单一样，需要维护者同意。

## 日常命令（仓库根目录）

```sh
go generate ./...   # 重新生成 contracts/pb（入口：contracts/doc.go → contracts/internal/protocgen）
go test ./...
go build ./...
go vet ./...
go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' ./cmd/alonabot ./sdk | sort -u   # Core 与 SDK 根包链接的模块，对照允许清单（D-52）
```

生成器会先删除 `contracts/pb/` 下的 `*.pb.go` 再重新生成；新增 `.proto` 放在 `contracts/proto/alonabot/v1/`，`go_package` 统一为 `github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1`。

交叉编译矩阵（只证明能构建，不能代替目标平台的运行证据）：

```sh
for p in linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64; do
  GOOS=${p%/*} GOARCH=${p#*/} CGO_ENABLED=0 go build ./... && echo "OK $p" || echo "FAIL $p"
done
GOARCH=amd64 CGO_ENABLED=0 go test ./...   # Apple Silicon 上经 Rosetta 运行 darwin/amd64 测试
```

## 运行 Core 与管理组件

```sh
go build -o bin/alonabot ./cmd/alonabot
bin/alonabot run --data-dir ./data      # 不带子命令时等同 run；按安装记录启动全部已启用的组件
# Ctrl-C / SIGTERM：排空 → STOPPING/OnStop → 退出；同一 data_dir 只能运行一个 Core
```

`run` 与库模式（`alonabot.Start`）共用根包的同一套 Core 装配。组件经管理服务安装与启停，CLI 通过 `base_url` 上的管理 API 操作，token 取自 `ALONABOT_ADMIN_TOKEN` 或 `--token-file`，管理地址可用 `--admin-url` 覆盖：

```sh
bin/alonabot install --archive pkg.tar.gz --release release.json --signature release.json.sig --public-key release.pub --yes
bin/alonabot install --source owner/repo --yes      # 从 GitHub Release 下载 release.json、可选的 .sig 与 release.pub
bin/alonabot enable <id>
bin/alonabot list
bin/alonabot uninstall <id> --purge-data --yes      # 或 --purge；不带选项时只卸载
bin/alonabot adapter instance create <adapter-package-id> <instance-id> --config instance.json
bin/alonabot admin add telegram:123456
```

需要确认的操作不带 `--yes` 时列出受影响的对象后退出。开发中的组件可以按下节以目录运行，也可以用库模式，例如 `go run ./examples/devbot --data-dir ./data`。

## 开发运行与热部署

在包含 `plugin.yaml` 的组件工程中运行 `alonabot run --reload`，或者明确给出一个或多个工程目录：

```sh
bin/alonabot run ./myplugin ./myadapter --data-dir ./data --reload
bin/alonabot run ./examples/plugins/lifecycle --stdio
```

未给目录时，当前目录含 `plugin.yaml` 就运行该工程；其他情况按安装记录启动 Core。组件身份与角色取自工程 manifest，全部工程构建成功并通过身份冲突检查后才启动组件。重复目录、重复 ID、与已安装包冲突的 ID 或 `adapter_id` 都会报错。

构建复用 SDK 生成器与 `go build`，可执行文件写到 `<data_dir>/dev/bin/`，生成资源与薄入口通过 overlay 提供，作者目录中的源码和 module 文件保持原样。插件与 Agent 构建入口包；Adapter 工程导出 `New()` 并提供可导入的包，构建器为其生成调用 `sdk.RunAdapter(New())` 的薄入口。

开发组件通过 supervisor 在独立进程中运行，缺省使用 TCP loopback，`--stdio` 切换本次开发进程的传输。组件固定使用 `dev:<组件 ID>` 代次，配置每次从 `<data_dir>/plugins/<id>/config.json` 读取，缺失时按空对象准备；准备后的配置保存在开发代次快照中。受管 SQLite 为同目录的 `dev.sqlite`，跨重启保留。Adapter 实例配置来自 `core_config.json` 的 `adapters`。开发日志缺省为 debug，显式 `log.sources` 设置优先。

`--reload` 每 `DEV_RELOAD_POLL` 检查源码、`//go:embed` 资源、`go.mod`、`go.sum`、`plugin.yaml`、两个配置 Schema 和 `i18n/`，排除数据目录、vendor、testdata 及以 `.`、`_` 开头的目录。轮询间隔缺省为 500ms，例如 `--args DEV_RELOAD_POLL=1s` 可以覆盖。

源码或资源变化后重新构建，成功时只重启变化的组件；编译失败保留编译器诊断，旧组件继续运行。只改组件 `config.json` 时直接重启该组件，只改 Adapter 的一个实例配置时只重启该实例，二进制保持原样。不带 `--reload` 时构建并运行一次。

目前，单独修改 Core 的 admins、命令前缀或缺省语言等设置后，需重启 Core 生效；Adapter 实例重启会同时刷新共享 Core 设置。WebUI 目录运行等待 WebUI 选项接通。

独立进程调试时将调试器附加到组件 PID；需要在同一进程调试 Core 和组件时使用下文库模式。

## Core 设置与 HTTP 入口

`run` 取得 core.lock 后读取 `<data_dir>/core_config.json`，Core 设置统一从这个 JSON 对象读取；文件不存在时以 0600 权限写入缺省内容。未知键、重复键、类型或取值错误时，以 usage 退出码报告完整键路径与行列号。

键与缺省值见 CONTRACTS §15.1；`base_url` 缺省为 `http://127.0.0.1:7650`，管理入口与插件路由共用这个地址。旧键 `admin_base_url` 与 `plugin_base_url` 按未知键报错；已有配置需改用 `base_url`。`admins`、`command_prefixes`、`command_prefixes_by_instance`、`default_locale`、`log.sources` 与 `adapters` 已支持解析与保存；本地 `run` 按监听、日志、代理与 Base URL 装配。

整个字符串值写成 `${NAME}` 时，读取时展开环境变量；NAME 由 `A-Z`、`a-z`、数字与下划线组成，首字符为字母或下划线。变量缺失会报告键路径并拒绝启动。

`(*settings.File).Update` 在未展开的文档上编辑，校验后以 0600 临时文件替换原文件，保留其他键的顺序与 `${...}` 写法，新键追加到所在对象末尾；写回使用两个空格缩进并以换行结束。`$version` 当前为 1，缺失时按 1 读取、下次写回时补上，高于 1 时拒绝读取。

部署运行中，手改设置后重启 Core 生效；开发运行的实例配置可按上节热部署。`--listen`、`--log-level`、`--log-format` 只在命令行显式给出时覆盖设置，只对本次运行有效，文件内容保持不变：

```sh
bin/alonabot run --data-dir ./data --listen 127.0.0.1:8080
```

Core 在 `listen` 地址（缺省 `127.0.0.1:7650`）上提供唯一的公开 HTTP 入口。请求 Host 的主机名须匹配 `base_url`：`/plugins/{id}/...` 转发给该插件当前 Session 的路由，其余路径交给管理处理器；其他 Host 返回 421。缺省设置下，lifecycle 示例的私有页面地址为 `http://127.0.0.1:7650/plugins/demo/page`。listen 为非 loopback 地址时，`base_url` 必须是 https，或者配置非空的 `trusted_proxies`（由可信反代终止 TLS），设置文件与 `--listen` 都按此校验。

插件路由的全部响应（含 Core 的错误响应）统一设置 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff`，覆盖插件的同名头。Core 清除插件设置的全部 `Set-Cookie` 和 `Access-Control-Allow-Credentials`（含 trailer），Cookie 只以 WARN 记录名称；缺失 Content-Type 时补 `application/octet-stream`。HTML 页面按此沙箱策略返回。

设置了 `proxy` 时，Core 与子进程使用它作为 `HTTP_PROXY`/`HTTPS_PROXY`；为空时原样传递环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。

首次启动且尚无管理凭证时，Core 自动生成管理 token：stdin 与 stderr 都是终端时只向 stderr 显示一次；系统服务、容器等非终端启动时，以 0600 权限写入 `<data_dir>/admin_token`，INFO 日志记录文件路径。读取文件中的 token：

```sh
bin/alonabot admin-token show --data-dir ./data
```

`show` 直接读取文件，可在 Core 运行时执行，也可从非终端调用；文件不存在时返回英文错误。第一次登录成功后，Core 删除该文件。登录接口为 `base_url` 下的 `POST /api/auth/login`，要求受控 Origin、`Content-Type: application/json` 与正文 `{"token":"<token>"}`，成功后返回管理 Cookie 和 `csrf_token`；CLI 可用 `Authorization: Bearer <token>`。

私有插件路由复用管理认证：有效管理 Cookie 或 Bearer token；Cookie 认证的非 GET/HEAD 请求还须携带 `X-Alonabot-CSRF` 与受控 Origin，校验失败返回 401。公开路由无需管理会话，由插件按其声明的回调来源校验规则处理。转发给插件前，Core 清除管理 Cookie、Authorization 与 CSRF 头。

需要在启动前初始化凭证时，停止 Core 后在本机终端运行 `bin/alonabot admin-token init --data-dir ./data`。token 遗失时用 `bin/alonabot admin-token reset --data-dir ./data` 重新生成，旧 token 与全部管理会话随之失效。`init` 与 `reset` 都要求 stdin 与 stderr 是终端，只向 stderr 显示 token 一次；Core 运行中执行会因 core.lock 失败。

调优参数在启动时解析；修改后重启生效。查看全部 37 项参数及当前生效值：

```sh
bin/alonabot tune list
bin/alonabot --args REQUEST_TIMEOUT=2s tune list
bin/alonabot run --data-dir ./data --args REQUEST_TIMEOUT=2s --args ./tuning.txt
```

`--args` 可重复，放在子命令前后均生效，后出现的覆盖先出现的。直接参数支持逗号分隔的 `KEY=VALUE`；文件每行一项，空行和以 `#` 开头的注释行会被忽略。时长使用 `2s` 等 Go 时长格式；大小使用 `B`、`KiB`、`MiB`、`GiB`，`MB` 等十进制单位会报错。`run` 将最终覆盖值及出处记为 WARN。

插件包导出 `func New() *sdk.Plugin`，命令、文本匹配与事件都在插件对象上链式注册，包级函数也这样注册；构建期生成器不读取任何标记，只生成包内变量 `alonabotGenerated`（配置 Schema、Service Schema、i18n 与 manifest），由 `New` 显式调用 `sdk.UseGenerated(p, alonabotGenerated)` 绑定。在插件构建入口目录生成或检查 `alonabot_gen.go`：

```sh
bin/alonabot plugin generate --dir ./myplugin   # 内容不变时不改文件
bin/alonabot plugin check --dir ./myplugin --goos linux --goarch amd64 --tags a,b
```

两者按 `--goos`、`--goarch`、`--tags` 筛选源码（缺省取当前环境），只扫描入口及其依赖的本 module 包，不受现有生成文件的 import 影响。源码问题逐行输出为 `文件:行:列: 说明`，此时不写文件；`check` 在生成内容与现有文件不同时报告过期。两种情况都以退出码 1 结束。

真实进程验收（Unix；构建 Core、lifecycle 与临时热部署工程）：

```sh
go test -count=3 ./cmd/alonabot/ ./internal/core/supervisor/
go test -race -count=3 ./internal/core/supervisor/
```

文档改动用 `git diff --check` 检查空白错误，链接与技能元数据由审查核对。包文档、命名、注释与测试提交范围由 [编码规范](coding-standards.md) 与 [注释技能](../.agents/skills/alonabot-code-commenting/SKILL.md) 指导审查。

## 库模式运行（同一进程）

根包 `alonabot` 在当前进程中运行 Core 与开发中的组件，示例见 `examples/devbot`：

```go
core, err := alonabot.Start(ctx, alonabot.DataDir(dir), alonabot.Plugin(plugin.New()), alonabot.Adapter(adapter.New()))
if err != nil {
	return err
}
return core.Wait() // ctx 取消（例如 SIGTERM）后在 STOP_TOTAL 内停止全部组件
```

```sh
go generate ./examples/devbot/...
go run ./examples/devbot -data-dir ./data
```

组件以 source=dev 登记，代次固定为 `dev:<组件 ID>`；资源取自各组件包内生成并嵌入的副本。

## Adapter 的 HTTP 路由

Adapter 可以实现 `sdk.AdapterHTTP`，在统一端口的 `/plugins/<Adapter 包 ID>/...` 下声明路由，例如 OneBot 11 反向 WebSocket。路由规则与插件相同：私有路由需要 Core 会话，非 GET/HEAD 的 Cookie 请求还要验证 CSRF 与 Origin。公开路由除回调校验方式外，还可以用 `sdk.Public(sdk.AuthPlatformToken)`：Core 代理保留入站的 `Authorization` 交给 Adapter，由 Adapter 校验平台凭证，并给出错误时的平台响应。插件路由的 `Authorization` 一律被清除。两者都会清除入站的 Cookie、CSRF、Forwarded 与 X-Forwarded-*，并注入内部代理凭证；SDK 先恒定时间比较该凭证，比较后移除，再进入用户 Handler。WebSocket Upgrade 由代理双向转发，连接寿命不受普通请求期限限制，路由撤销或 Session 失效时由 Core 关闭。

## 外部测试环境

- **Telegram：** 真实测试需要授权的测试 Bot，目前没有，相关项记为「未运行」；Adapter 测试改用本地 Bot API 模拟服务。
- **目标平台：** 本机只能运行 darwin/arm64；其余四个组合只做 `CGO_ENABLED=0` 交叉编译，尚无运行证据。

## 规则与检查

| 规则 | 依据 | 检查 | 状态 |
| --- | --- | --- | --- |
| 生成代码可重复生成，结果稳定 | AGENTS「项目约定」 | 执行 `go generate ./...` 后检查 `git diff --exit-code` | 已观测（手动执行） |
| Envelope 字段号/枚举值与 WIRE-FORMAT 一致 | spec/WIRE-FORMAT.md §1/§2/§5A | `contracts/wirespec_test.go`（`go test ./contracts/`） | 已观测 |
| Core 与 SDK 根包只链接允许的模块 | CONTRACTS §10、D-52 | 审查时运行上方的 `go list -deps` 命令（需要时按目标平台设 GOOS/GOARCH） | 人工/agent 审查规则，不由 CI 自动证明 |
| 五目标平台 `CGO_ENABLED=0` 可构建 | CONTRACTS §6.1/§7.1 | 下方交叉编译命令（手动执行） | 已观测（只验证了构建；运行见 V-32） |
| 包文档、命名、注释与测试提交范围 | coding-standards.md | 代码审查与 PR 清单 | 人工/agent 审查规则，不由 CI 自动证明 |
| V-01～V-117 证据 | spec/TEST-MATRIX.md | 按编号记录测试位置、命令、平台和结果 | §3.1 所列分项已观测（darwin/arm64）；其余未运行 |

状态说明：**已定义** 表示规则和检查方式已写下；**已实现** 表示检查代码已存在；**已观测** 表示检查真实运行过并有结果。工作流本身不代表目标提交已经验证。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) 在 PR 到 main、main push 和手动触发时运行 `git diff --check`。存在产品 Go 文件时，才运行 Go 格式、Protobuf 生成一致性、vet、race 测试和五目标平台 `CGO_ENABLED=0` 构建；无产品代码时明确标注 Go 检查未运行。

生成检查从官方 Protobuf release 下载当前约定的 protoc 36.0；工具链变更时同时更新生成器、开发指南和工作流。代码 job 使用只读 token，不使用生产凭据。

[PR metadata](../.github/workflows/pr-metadata.yml) 校验非草稿 PR 的实际说明和清单；[triage](../.github/workflows/pr-triage.yml) 同步标签与 assignee，[auto-assign](../.github/workflows/auto-assign.yml) 请求非作者维护者审阅。这些 `pull_request_target` 流程不 checkout/执行 PR 代码，新工作流合入默认分支后才开始处理后续事件。Issue 表单自带 assignee，不依赖工作流执行。

远端合并保护由维护者单独管理，不能因存在这些 YAML 就声称检查不可绕过。本次建设不添加发布、真实 Telegram 调用或自动关闭 PR 的流程。
