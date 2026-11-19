# 可重复执行的正式管理场景

这里的场景供 T-10 使用指定的 Core 二进制、临时 `data_dir` 和测试包验收。入口复用 [正式管理集成流程](../../cmd/alonabot/t09_e2e_test.go) 的全部断言；临时 Go overlay 只替换输入和增加证据记录，仓库中的测试源码保持原样。管理动作始终调用 `--alonabot` 指定的程序。

## 构建三版测试包并执行正式流程

在待验收版本的仓库根目录使用项目要求的 Go 工具链和 `uv`。入口只使用 Python 标准库与仓库已有 Go 依赖。当前支持 macOS、Linux 等 Unix 平台，与原集成测试的 `unix` build tag 一致；其他平台的实际运行由 T-10 单独取证。

以下为 fish 命令。`artifacts` 存放本次独立构建的 Core、夹具源码、签名包与原始日志；也可以把 `--alonabot` 换成待验收 Core 的绝对路径。

```fish
set artifacts (mktemp -d /tmp/alonabot-scenarios.XXXXXX)
go build -o "$artifacts/alonabot" ./cmd/alonabot

uv run --no-project scripts/scenarios/run.py build --out "$artifacts/packages"

set data_dir (mktemp -d /tmp/alonabot-management.XXXXXX)
uv run --no-project scripts/scenarios/run.py management \
    --alonabot "$artifacts/alonabot" --data-dir "$data_dir" \
    "$artifacts/packages/1.0.0/package.tar.gz" \
    "$artifacts/packages/2.0.0/package.tar.gz" \
    "$artifacts/packages/3.0.0/package.tar.gz"
```

`build` 复用现有 `recoveryFixtureSource` 和 `writeTestPackage`：为 `recovery` / `github.com/example/recovery` 分别编译实际的 1.0.0、2.0.0、3.0.0 程序，添加配置 schema、`/info` 版本与配置观测入口及 `OnReady` 日志。三版包由同一把临时 Ed25519 密钥签名；每个版本目录包含 `main.go`、`plugin`、`package.tar.gz`、`release.json`、`release.json.sig` 和 `release.pub`。私钥只在本次构建进程内存中使用。

传入其他构建产物时，三版包需提供上述相同身份、版本、签名公钥、配置字段和 HTTP 行为；每个归档旁保留对应的发行清单、签名和公钥。`build --out` 使用尚未存在的输出目录；`management --data-dir` 使用已存在的空临时目录。

正式流程依次执行：

1. install → enable；核对首个 Active、Session、实际程序版本、`label=initial` 与 SQLite 初值 1。
2. 写入 DB=41，编辑 `config.json`；确认 reload 前仍使用旧配置。
3. restart 保持 Active，换新 Session；CLI reload 创建新 Active 和独立 DB，配置变为 `reloaded`。
4. update 到 2.0.0；停止 Core 后在真实状态库签发按钮并验证当时有效，再恢复 Core。
5. pin 2.0.0；3.0.0 更新被拒绝时保留 Active、Session 和 DB；解除 pin 后同一包更新成功，再写入 DB=99。
6. disable → 默认 uninstall → 同来源重装 → enable，保留代次、配置和 DB=99。
7. 从 generations API 的实际保留列表选择 2.0.0 快照，显式 rollback 创建新 Active/Session 和独立 DB，恢复 2.0.0、`reloaded` 与 DB=41。
8. 正常停止 Core，核对持久 Active、5 条 committed Journal 的操作/前后代次/DB 路径，并通过真实按钮解析服务确认旧按钮因代次变化失效。

## 读取执行证据

成功时输出 `PASS`，退出码为 0。失败时保留完整 Go 测试输出并返回其退出码；场景清理仅终止自己启动的进程，运行目录与产物保留。

| 位置 | 内容 |
| --- | --- |
| `<out>/scenario-evidence/` | 构建命令、输入与 overlay、构建原始输出 |
| `<data_dir>/scenario-evidence/inputs.json` | 指定二进制与三版包的绝对路径、执行平台、实际 Go 命令 |
| `<data_dir>/scenario-evidence/overlay.json`、`*_test.go` | 本次有效的测试入口与辅助源码，可检查输入替换及断言 |
| `<data_dir>/scenario-evidence/go-test.log` | 完整 stdout/stderr 与测试结果 |
| `<data_dir>/scenario-evidence/events.jsonl` | 成功管理动作的完整 CLI 输出、预期 pin 拒绝、包状态快照、HTTP 响应、Journal 及 Core 原始 JSON 日志 |
| `<data_dir>/state.sqlite`、`plugins/recovery/` | Core 已退出后的持久状态、保留代次、配置与业务 DB |

CLI 的本机 HTTP 转发器沿用集成测试辅助，只把 Host 改为 Core 配置的 `base_url`，请求仍经真实 Bearer 认证与正式管理入口。按钮取证在 Core 退出后使用同一状态库，沿用集成测试的固定路由描述。

源码定位锚点变化时入口明确报错，更新场景入口后再运行。临时 overlay 保留原主路径断言，`TestMain` 在本场景中跳过常规 Core/夹具构建，使用指向指定二进制的临时符号链接。
