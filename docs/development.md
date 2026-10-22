# 开发与验证

负责范围：环境准备、日常命令、外部测试环境、规则与检查的对照。
更新时机：工具链或依赖版本变化、新增或调整检查入口、测试环境变化时。

## 环境

| 工具 | 版本与来源 | 备注 |
| --- | --- | --- |
| Go | 1.27.1（asdf，`~/.tool-versions`） | `go.mod` 的 go 指令固定为该版本 |
| protoc | 36.0（libprotoc） | 生成入口会核对版本，由 T-01 落地 |
| protoc-gen-go | 由 `go.mod` 的 `tool` 指令固定 | 通过 `go tool` 调用，不全局安装；T-01 落地 |
| Docker | 用于临时 PostgreSQL | 只在 PG 测试中使用 |
| Node（前端） | 26.5.1（asdf） | web/admin 从 T-08 开始使用 |

已批准的 Go 依赖（2026-10-21）：`google.golang.org/protobuf`、`gorm.io/gorm`、`github.com/libtnb/sqlite`、`modernc.org/sqlite`、`gorm.io/driver/postgres`、`golang.org/x/sys`、`go.yaml.in/yaml/v3`。前端已批准：Vite、TypeScript、React 19、vitest。新增清单外依赖需要先征得维护者同意。

## 日常命令（仓库根目录）

```sh
go generate ./...   # 重新生成 contracts（T-01 建立入口）
go test ./...
go build ./...
go vet ./...
```

在 T-01 落地之前，`./...` 下还没有包；`go build` 只会给出 "matched no packages" 警告，`go vet/test` 会以非零退出。

仓库文档与技能检查使用已安装的 Node（只用标准库，无 npm 依赖）：

```sh
node --test scripts/check-repository.test.mjs
node scripts/check-repository.mjs
git diff --check
```

检查器验证维护入口的内联本地文件链接、技能名称/描述及可选 UI 提示，不联网，不验证全部历史规格、页内锚点或注释语义。文件职责、逐字段说明和 step comment 由 [编码规范](coding-standards.md) 与 [注释技能](../.agents/skills/alonabot-code-commenting/SKILL.md) 指导审查。

## 外部测试环境

- **PostgreSQL：** 测试读取环境变量 `ALONABOT_TEST_PG_DSN`；未设置时跳过，在证据中记为「未运行」。临时容器：

  ```sh
  docker run -d --rm --name alonabot-test-pg -e POSTGRES_PASSWORD=alonabot -p 127.0.0.1::5432 postgres:17
  docker port alonabot-test-pg 5432   # 查看实际映射到的端口
  # 测试结束后
  docker stop alonabot-test-pg
  ```

- **Telegram：** 真实测试需要授权的测试 Bot，目前没有，相关项记为「未运行」；T-03 改用本地 Bot API 模拟服务。
- **目标平台：** 本机只能运行 darwin/arm64；其余四个组合只做 `CGO_ENABLED=0` 交叉编译，运行证据留给 T-10。

## 规则与检查

| 规则 | 依据 | 检查 | 状态 |
| --- | --- | --- | --- |
| 模块 import 方向符合架构表 | [ARCHITECTURE.md](../ARCHITECTURE.md) | 架构守卫测试（goal-1 任务） | 已定义，未实现 |
| 生成代码可重复生成，结果稳定 | PLAN T-01 | 执行 `go generate ./...` 后检查 `git diff --exit-code` | 已定义，未实现 |
| Envelope 字段号/枚举值与 WIRE-FORMAT 一致 | spec/WIRE-FORMAT.md | contracts 字段号 fixture 测试 | 已定义，未实现 |
| 五目标平台 `CGO_ENABLED=0` 可构建 | CONTRACTS §6.1/§7.1 | 交叉编译矩阵命令 | 已定义，未实现 |
| 维护文档与技能入口可从干净 checkout 读取 | AGENTS、CONTRIBUTING | `node scripts/check-repository.mjs`，CI 的 Repository contract | 已实现；运行结果见对应 PR |
| 文件职责头、逐字段说明和按逻辑段的 step comment | coding-standards.md | 代码审查与 PR 清单 | 人工/agent 审查规则，不由 CI 自动证明 |
| V-01～V-47 证据 | spec/TEST-MATRIX.md | 按编号记录测试位置、命令、平台和结果 | 全部未运行 |

状态说明：**已定义** 表示规则和检查方式已写下；**已实现** 表示检查代码已存在；**已观测** 表示检查真实运行过并有结果。工作流本身不代表目标提交已经验证。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) 在 PR 到 main、main push 和手动触发时运行维护入口检查。存在非 `experiments/` 的产品 Go 文件时，才运行 Go 格式、Protobuf 生成一致性、vet、race 测试和五目标平台 `CGO_ENABLED=0` 构建；无产品代码时明确标注 Go 检查未运行。

Go job 使用一次性 `postgres:17` service，测试 DSN 只指向 job 内的本地测试库。生成检查从官方 Protobuf release 下载当前约定的 protoc 36.0；工具链变更时同时更新生成器、开发指南和工作流。代码 job 使用只读 token，不使用生产凭据。

[PR metadata](../.github/workflows/pr-metadata.yml) 校验非草稿 PR 的实际说明和清单；[triage](../.github/workflows/pr-triage.yml) 同步标签与 assignee，[auto-assign](../.github/workflows/auto-assign.yml) 请求非作者维护者审阅。这些 `pull_request_target` 流程不 checkout/执行 PR 代码，新工作流合入默认分支后才开始处理后续事件。Issue 表单自带 assignee，不依赖工作流执行。

远端合并保护由维护者单独管理，不能因存在这些 YAML 就声称检查不可绕过。本次建设不添加发布、真实 Telegram 调用或自动关闭 PR 的流程。
