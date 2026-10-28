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

已锁定并经测试验证的版本：protobuf v1.36.12、go.yaml.in/yaml/v3 v3.0.5、gorm v1.31.2、libtnb/sqlite v1.2.2、modernc.org/sqlite v1.61.0（已验证 CRUD/Tx/WAL Backup/迁移账本）；golang.org/x/sys v0.49.0 为直接依赖，用于 internal/oslock 的 flock/LockFileEx。

已批准的 Go 依赖（2026-10-21）：`google.golang.org/protobuf`、`gorm.io/gorm`、`github.com/libtnb/sqlite`、`modernc.org/sqlite`、`golang.org/x/sys`、`go.yaml.in/yaml/v3`（`gorm.io/driver/postgres` 的批准于 2026-10-27 随 D-48 撤销）；2026-10-25 批准 `github.com/modelcontextprotocol/go-sdk`，只用于官方 Agent 插件的 MCP 客户端，版本在 T-12.3 验证后锁定；2026-10-26 批准 `github.com/spf13/cobra`（随之引入 `github.com/spf13/pflag`），只用于 `cmd/alonabot` 的命令树与全局参数，版本在 T-02.2 引入时锁定；2026-10-26 批准 `github.com/google/jsonschema-go`（MCP go-sdk 也直接依赖它），用于 JSON Schema 的表示、序列化、`Resolve`、`ApplyDefaults` 与 `Validate`：配置 Schema（`sdk/config`、构建期生成器、Core 的配置与表单 Schema）、Service 与工具的输入 Schema、`plugins/agent`。配置 Schema 由生成器构造，不用 `For`/`ForType` 推导（D-43）；版本在首个使用它的切片引入时锁定。2026-10-27 批准 `github.com/coder/websocket`，只用于官方 OneBot 11 Adapter（`adapters/onebot11/`）的 WebSocket 连接，版本在 T-15.1 引入时锁定（D-46）。前端已批准：Vite、TypeScript、React 19、vitest。新增清单外依赖需要先征得维护者同意。

Core 可执行文件（`./cmd/alonabot`）与 SDK 根包（`./sdk`）只能链接上述清单的一个子集，子集维护在 `scripts/check-go-deps.mjs`，由 CI 在五个目标上检查（D-52）：Core 允许 protobuf、modernc SQLite 及其依赖、yaml、x/sys、cobra 与 pflag、仅 Windows 链接的 mousetrap、jsonschema-go；SDK 根包只允许 protobuf、x/sys 与 jsonschema-go。GORM、MCP go-sdk、coder/websocket 只出现在按需 SDK 包、官方插件或 Adapter 中。修改子集与修改已批准清单一样，需要维护者同意。

## 日常命令（仓库根目录）

```sh
go generate ./...   # 重新生成 contracts/pb（入口：contracts/doc.go → contracts/internal/protocgen）
go test ./...
go build ./...
go vet ./...
node scripts/check-go-deps.mjs   # Core 与 SDK 根包只链接允许的模块（D-52）
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

真实进程验收（Unix；会在测试中构建上述两个二进制）：

```sh
go test -count=3 ./cmd/alonabot/ ./internal/core/supervisor/
go test -race -count=3 ./internal/core/supervisor/
```

仓库文档与技能检查使用已安装的 Node（只用标准库，无 npm 依赖）：

```sh
node --test scripts/check-repository.test.mjs
node scripts/check-repository.mjs
git diff --check
```

检查器验证维护入口的内联本地文件链接、技能名称/描述及可选 UI 提示，不联网，不验证全部历史规格、页内锚点或注释语义。包文档、命名、注释与测试提交范围由 [编码规范](coding-standards.md) 与 [注释技能](../.agents/skills/alonabot-code-commenting/SKILL.md) 指导审查。

## 外部测试环境

- **Telegram：** 真实测试需要授权的测试 Bot，目前没有，相关项记为「未运行」；Adapter 测试改用本地 Bot API 模拟服务。
- **目标平台：** 本机只能运行 darwin/arm64；其余四个组合只做 `CGO_ENABLED=0` 交叉编译，尚无运行证据。

## 规则与检查

| 规则 | 依据 | 检查 | 状态 |
| --- | --- | --- | --- |
| 生成代码可重复生成，结果稳定 | AGENTS「项目约定」 | 执行 `go generate ./...` 后检查 `git diff --exit-code` | 已观测（手动执行） |
| Envelope 字段号/枚举值与 WIRE-FORMAT 一致 | spec/WIRE-FORMAT.md §1/§2/§5A | `contracts/wirespec_test.go`（`go test ./contracts/`） | 已观测 |
| Core 与 SDK 根包只链接允许的模块 | CONTRACTS §10、D-52 | `node scripts/check-go-deps.mjs`，CI 的 Go job | 已观测（本机检查五个目标，通过） |
| 五目标平台 `CGO_ENABLED=0` 可构建 | CONTRACTS §6.1/§7.1 | 下方交叉编译命令（手动执行） | 已观测（只验证了构建；运行见 V-32） |
| 维护文档与技能入口可从干净 checkout 读取 | AGENTS、CONTRIBUTING | `node scripts/check-repository.mjs`，CI 的 Repository contract | 已实现；运行结果见对应 PR |
| 包文档、命名、注释与测试提交范围 | coding-standards.md | 代码审查与 PR 清单 | 人工/agent 审查规则，不由 CI 自动证明 |
| V-01～V-76 证据 | spec/TEST-MATRIX.md | 按编号记录测试位置、命令、平台和结果 | §3.1 所列分项已观测（darwin/arm64）；其余未运行 |

状态说明：**已定义** 表示规则和检查方式已写下；**已实现** 表示检查代码已存在；**已观测** 表示检查真实运行过并有结果。工作流本身不代表目标提交已经验证。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) 在 PR 到 main、main push 和手动触发时运行维护入口检查。存在产品 Go 文件时，才运行 Go 格式、Protobuf 生成一致性、vet、依赖允许清单检查、race 测试和五目标平台 `CGO_ENABLED=0` 构建；无产品代码时明确标注 Go 检查未运行。

生成检查从官方 Protobuf release 下载当前约定的 protoc 36.0；工具链变更时同时更新生成器、开发指南和工作流。代码 job 使用只读 token，不使用生产凭据。

[PR metadata](../.github/workflows/pr-metadata.yml) 校验非草稿 PR 的实际说明和清单；[triage](../.github/workflows/pr-triage.yml) 同步标签与 assignee，[auto-assign](../.github/workflows/auto-assign.yml) 请求非作者维护者审阅。这些 `pull_request_target` 流程不 checkout/执行 PR 代码，新工作流合入默认分支后才开始处理后续事件。Issue 表单自带 assignee，不依赖工作流执行。

远端合并保护由维护者单独管理，不能因存在这些 YAML 就声称检查不可绕过。本次建设不添加发布、真实 Telegram 调用或自动关闭 PR 的流程。
