# 开发与验证

负责范围：环境准备、日常命令、外部测试环境、规则与检查的对照。
更新时机：工具链或依赖版本变化、新增或调整检查入口、测试环境变化时。

## 环境

| 工具 | 版本与来源 | 备注 |
| --- | --- | --- |
| Go | 1.27.1（asdf，`~/.tool-versions`） | `go.mod` 的 go 指令固定为该版本 |
| protoc | 36.0（`protoc --version` 输出 `libprotoc 36.0`；生成文件头记为 v7.36.0） | 版本不符时 `contracts/internal/protocgen` 拒绝生成 |
| protoc-gen-go | v1.36.12（`go.mod` 的 `tool` 指令，与 `google.golang.org/protobuf` 同版本） | 通过 `go tool -n protoc-gen-go` 定位，不全局安装 |
| Node（前端） | 26.5.1（asdf） | 供 web/admin 使用（尚未创建） |

已锁定并经测试验证的版本：protobuf v1.36.12、go.yaml.in/yaml/v3 v3.0.5、gorm v1.31.2、libtnb/sqlite v1.2.2、modernc.org/sqlite v1.61.0（已验证 CRUD/Tx/WAL Backup/迁移账本）；golang.org/x/sys v0.49.0 为直接依赖，用于 internal/oslock 的 flock/LockFileEx；Cobra v1.10.2、pflag v1.0.9、仅 Windows 链接的 mousetrap v1.1.0。

已批准的 Go 依赖（2026-10-21）：`google.golang.org/protobuf`、`gorm.io/gorm`、`github.com/libtnb/sqlite`、`modernc.org/sqlite`、`golang.org/x/sys`、`go.yaml.in/yaml/v3`（`gorm.io/driver/postgres` 的批准于 2026-10-27 随 D-48 撤销）；2026-10-25 批准 `github.com/modelcontextprotocol/go-sdk`，只用于官方 Agent 组件（`agent/`）的 MCP 客户端，版本在 T-12.3 验证后锁定；2026-10-26 批准 `github.com/spf13/cobra`（随之引入 `github.com/spf13/pflag`），只用于 `cmd/alonabot` 的命令树与全局参数，T-02.2 已锁定 Cobra v1.10.2、pflag v1.0.9 与仅 Windows 链接的 mousetrap v1.1.0；2026-10-26 批准 `github.com/google/jsonschema-go`（MCP go-sdk 也直接依赖它），用于 JSON Schema 的表示、序列化、`Resolve`、`ApplyDefaults` 与 `Validate`：配置 Schema（`sdk/config`、构建期生成器、Core 的配置与表单 Schema）、Service 与工具的输入 Schema、`agent/`。配置 Schema 由生成器构造，不用 `For`/`ForType` 推导（D-43）；版本在首个使用它的切片引入时锁定。2026-10-27 批准 `github.com/coder/websocket`，只用于官方 OneBot 11 Adapter（`adapters/onebot11/`）的 WebSocket 连接，版本在 T-15.1 引入时锁定（D-46）。前端已批准：Vite、TypeScript、React 19、vitest。新增清单外依赖需要先征得维护者同意。

Core 可执行文件（`./cmd/alonabot`）与 SDK 根包（`./sdk`）只能链接上述清单的一个子集（D-52，D-68 修订），由审查时用下方的 `go list -deps` 命令核对：Core 允许 protobuf、modernc SQLite 及其依赖、GORM 与 libtnb/sqlite 及其依赖（jinzhu/inflection、jinzhu/now、golang.org/x/text）、yaml、x/sys、cobra 与 pflag、仅 Windows 链接的 mousetrap、jsonschema-go；SDK 根包只允许 protobuf、x/sys 与 jsonschema-go。MCP go-sdk、coder/websocket 只出现在官方插件、Agent 或 Adapter 中；插件只在导入 `sdk/gormdb` 时链接 GORM。修改子集与修改已批准清单一样，需要维护者同意。

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

## 本地运行一个插件（B1 开发入口）

```sh
go build -o bin/alonabot ./cmd/alonabot
go build -o bin/lifecycle ./examples/plugins/lifecycle
bin/alonabot run --data-dir ./data --plugin-id demo --entry bin/lifecycle          # TCP loopback（默认）
bin/alonabot run --data-dir ./data --plugin-id demo --entry bin/lifecycle --stdio  # stdio 传输
# Ctrl-C / SIGTERM：排空 → STOPPING/OnStop → 退出；同一 data_dir 只能运行一个 Core
```

首次运行会登记开发用安装记录并提交新代次（`mode=new`），之后复用已提交代次（`mode=restart`）。尚无正式的安装、来源校验与管理 CLI。

`run` 取得 core.lock 后读取 `<data_dir>/alonabot.yaml`；文件不存在时以 0600 权限写入缺省值（键与缺省值见 CONTRACTS §15.1），未知键、类型或取值错误时以 usage 退出码报告键名与行号。修改设置文件后重启生效。`--listen`、`--log-level`、`--log-format` 只在命令行显式给出时覆盖设置，只对本次运行有效，不改文件：

```sh
bin/alonabot run --data-dir ./data --plugin-id demo --entry bin/httpplugin --listen 127.0.0.1:8080
```

Core 在 `listen` 地址（缺省 `127.0.0.1:7650`）上提供唯一的公开 HTTP 入口，只按请求的 Host 分派：`admin_base_url` 的主机名进入管理入口，`plugin_base_url` 的主机名提供 `/plugins/{id}/...`，转发给该插件当前 Session 的路由；其他 Host 返回 421。本地调试时请求要带对应的 Host，例如缺省设置下用 `http://localhost:7650/plugins/demo/...` 访问插件。两个 Base URL 必须同 scheme、同有效端口、主机名不同（CONTRACTS §7.3）。listen 不是 loopback 地址时，`admin_base_url` 必须是 https，或者配置非空的 `trusted_proxies`（由可信反代终止 TLS），设置文件与 `--listen` 都按此校验。管理 API 先要初始化管理 token：停止 Core 后在本机终端运行 `bin/alonabot admin-token init --data-dir ./data`，token 只显示一次；遗失时用 `admin-token reset` 重新生成，旧 token 与全部管理会话随之失效。两个命令都要求 stdin 与 stderr 是终端，Core 运行中执行会因 core.lock 失败。登录接口为 admin host 上的 `POST /api/auth/login`，CLI 可用 `Authorization: Bearer <token>`。签发插件兑换票据的管理 API 尚未提供，私有插件路由目前仍无法从浏览器进入。设置了 `proxy` 时，Core 与子进程使用它作为 `HTTP_PROXY`/`HTTPS_PROXY`；为空时原样传递环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。

调优参数在启动时解析；修改后重启生效。查看全部 37 项参数及当前生效值：

```sh
bin/alonabot tune list
bin/alonabot --args REQUEST_TIMEOUT=2s tune list
bin/alonabot run --data-dir ./data --plugin-id demo --entry bin/lifecycle --args REQUEST_TIMEOUT=2s --args ./tuning.txt
```

`--args` 可重复，放在子命令前后均生效，后出现的覆盖先出现的。直接参数支持逗号分隔的 `KEY=VALUE`；文件每行一项，空行和以 `#` 开头的注释行会被忽略。时长使用 `2s` 等 Go 时长格式；大小使用 `B`、`KiB`、`MiB`、`GiB`，`MB` 等十进制单位会报错。`run` 将最终覆盖值及出处记为 WARN。

插件注册代码由构建期生成器从标记生成，标记语法见 `sdk/generate` 的包文档。在插件构建入口目录生成或检查 `alonabot_gen.go`：

```sh
bin/alonabot plugin generate --dir ./myplugin   # 内容不变时不改文件
bin/alonabot plugin check --dir ./myplugin --goos linux --goarch amd64 --tags a,b
```

两者按 `--goos`、`--goarch`、`--tags` 筛选源码（缺省取当前环境），只扫描入口及其依赖的本 module 包，不受现有生成文件的 import 影响。源码问题逐行输出为 `文件:行:列: 说明`，此时不写文件；`check` 在生成内容与现有文件不同时报告过期。两种情况都以退出码 1 结束。

真实进程验收（Unix；会在测试中构建上述两个二进制）：

```sh
go test -count=3 ./cmd/alonabot/ ./internal/core/supervisor/
go test -race -count=3 ./internal/core/supervisor/
```

文档改动用 `git diff --check` 检查空白错误，链接与技能元数据由审查核对。包文档、命名、注释与测试提交范围由 [编码规范](coding-standards.md) 与 [注释技能](../.agents/skills/alonabot-code-commenting/SKILL.md) 指导审查。

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
| V-01～V-84 证据 | spec/TEST-MATRIX.md | 按编号记录测试位置、命令、平台和结果 | §3.1 所列分项已观测（darwin/arm64）；其余未运行 |

状态说明：**已定义** 表示规则和检查方式已写下；**已实现** 表示检查代码已存在；**已观测** 表示检查真实运行过并有结果。工作流本身不代表目标提交已经验证。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) 在 PR 到 main、main push 和手动触发时运行 `git diff --check`。存在产品 Go 文件时，才运行 Go 格式、Protobuf 生成一致性、vet、race 测试和五目标平台 `CGO_ENABLED=0` 构建；无产品代码时明确标注 Go 检查未运行。

生成检查从官方 Protobuf release 下载当前约定的 protoc 36.0；工具链变更时同时更新生成器、开发指南和工作流。代码 job 使用只读 token，不使用生产凭据。

[PR metadata](../.github/workflows/pr-metadata.yml) 校验非草稿 PR 的实际说明和清单；[triage](../.github/workflows/pr-triage.yml) 同步标签与 assignee，[auto-assign](../.github/workflows/auto-assign.yml) 请求非作者维护者审阅。这些 `pull_request_target` 流程不 checkout/执行 PR 代码，新工作流合入默认分支后才开始处理后续事件。Issue 表单自带 assignee，不依赖工作流执行。

远端合并保护由维护者单独管理，不能因存在这些 YAML 就声称检查不可绕过。本次建设不添加发布、真实 Telegram 调用或自动关闭 PR 的流程。
