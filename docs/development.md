# 开发与验证

负责范围：环境准备、日常命令、外部测试环境、规则与检查的对照。
更新时机：工具链或依赖版本变化、新增或调整检查入口、测试环境变化时。

## 环境

| 工具 | 版本与来源 | 备注 |
| --- | --- | --- |
| Go | 1.27.1（asdf，`~/.tool-versions`） | `go.mod` 的 go 指令固定为该版本 |
| protoc | 36.0（`protoc --version` 输出 `libprotoc 36.0`；生成文件头记为 v7.36.0） | 版本不符时 `contracts/internal/protocgen` 拒绝生成 |
| protoc-gen-go | v1.36.12（`go.mod` 的 `tool` 指令，与 `google.golang.org/protobuf` 同版本） | 通过 `go tool -n protoc-gen-go` 定位，不全局安装 |
| Docker | 用于临时 PostgreSQL | 只在 PG 测试中使用 |
| Node（前端） | 26.5.1（asdf） | web/admin 从 T-08 开始使用 |

已批准的 Go 依赖（2026-10-21）：`google.golang.org/protobuf`、`gorm.io/gorm`、`github.com/libtnb/sqlite`、`modernc.org/sqlite`、`gorm.io/driver/postgres`、`golang.org/x/sys`、`go.yaml.in/yaml/v3`。前端已批准：Vite、TypeScript、React 19、vitest。新增清单外依赖需要先征得维护者同意。

## 日常命令（仓库根目录）

```sh
go generate ./...   # 重新生成 contracts/pb（入口：contracts/doc.go → contracts/internal/protocgen）
go test ./...
go build ./...
go vet ./...
```

生成器会先删除 `contracts/pb/` 下的 `*.pb.go` 再重新生成；新增 `.proto` 放在 `contracts/proto/alonabot/v1/`，`go_package` 统一为 `github.com/AlonaBot/alonabot/contracts/pb/alonabotv1`。

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
| 生成代码可重复生成，结果稳定 | PLAN T-01 | 执行 `go generate ./...` 后检查 `git diff --exit-code` | 已观测（G1-01，手动执行） |
| Envelope 字段号/枚举值与 WIRE-FORMAT 一致 | spec/WIRE-FORMAT.md §1/§2/§5A | `contracts/wirespec_test.go`（`go test ./contracts/`） | 已观测 |
| 五目标平台 `CGO_ENABLED=0` 可构建 | CONTRACTS §6.1/§7.1 | 交叉编译矩阵命令 | 已定义，未实现 |
| V-01～V-47 证据 | spec/TEST-MATRIX.md | 按编号记录测试位置、命令、平台和结果 | 全部未运行 |

状态说明：**已定义** 表示规则和检查方式已写下；**已实现** 表示检查代码已存在；**已观测** 表示检查真实运行过并有结果。目前还没有接入 CI。
