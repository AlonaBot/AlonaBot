# 可重复执行的管理与恢复场景

这里的场景供 T-10 使用指定的二进制、临时 `data_dir` 和测试包验收正式管理流程、崩溃恢复与残存写入者。正式流程复用 [管理集成测试](../../tests/e2e/management_e2e_test.go) 的全部断言，残存场景复用 [恢复集成测试](../../tests/e2e/recover_e2e_test.go) 的 `started` 子项，额外恢复与门禁断言位于 [场景模板](recovery.go.tmpl)。临时 Go overlay 替换输入、增加取证；管理动作始终调用 `--alonabot` 指定的程序。

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

## 构建恢复产物并执行两类故障场景

以下 fish 命令在仓库根目录运行。CLI 使用普通构建；`--overlay-core` 用于运行含精确检查点的临时 Core。`build-recovery` 独立构建产物，运行入口消费传入的二进制与包。

```fish
set recovery_artifacts (mktemp -d /tmp/alonabot-recovery.XXXXXX)
go build -o "$recovery_artifacts/alonabot" ./cmd/alonabot
uv run --no-project scripts/scenarios/run.py build-recovery \
    --out "$recovery_artifacts/fixtures"

set recovery_data (mktemp -d /tmp/alonabot-crash.XXXXXX)
uv run --no-project scripts/scenarios/run.py recovery \
    --alonabot "$recovery_artifacts/alonabot" \
    --overlay-core "$recovery_artifacts/fixtures/fault-core" \
    --data-dir "$recovery_data" \
    --recovery-package "$recovery_artifacts/fixtures/recovery/package.tar.gz" \
    --probe-package "$recovery_artifacts/fixtures/gateprobe/package.tar.gz" \
    --adapter-package "$recovery_artifacts/fixtures/adapter-package/package.tar.gz"

set survivor_data (mktemp -d /tmp/alonabot-survivor.XXXXXX)
uv run --no-project scripts/scenarios/run.py survivor \
    --alonabot "$recovery_artifacts/alonabot" \
    --overlay-core "$recovery_artifacts/fixtures/fault-core" \
    --data-dir "$survivor_data" \
    --recovery-package "$recovery_artifacts/fixtures/recovery/package.tar.gz"
```

`build-recovery --out` 同样接收尚未存在的目录。输出包含 `fault-core`、真实 `observed-plugin`、`probe`、`adapter`、对应生成源码，以及 `recovery/`、`gateprobe/`、`adapter-package/` 中的三个签名包。每个包目录保留 `package.tar.gz`、`release.json`、`release.json.sig`、`release.pub`，使用本次构建的同一把临时 Ed25519 密钥。`overlay.json` 复用原恢复钩子；`fault-overlay.json` 追加 FIFO 释放与绑定前检查点，其映射的修改源码全部保留在输出目录。

这些产物按执行平台构建。替换为其他产物时，`recovery` 身份为 `github.com/example/recovery`，版本 1.0.0、`migration_mode: none`、`prepare_effects: managed_db`；提供 SQLite `/value`、`starts`、`candidate-value`、`fail-ready` 和六类请求观测行为。`gateprobe` 提供 `/service`、`/bus`、`attempts` 与独立定时 `clock`；Adapter 为 `alonabot-adapter-fixture`，协议 `devloop`、实例 `one`。源码与声明以构建目录保留的实际文件为准。

### 恢复入口的检查点与断言

`recovery` 依次在 `<data_dir>/` 下创建七个独立运行目录：

| 子目录 | 故障窗口 | 恢复结果 |
| --- | --- | --- |
| `prepared` | PREPARED 后、提交前杀 Core | 旧库 41 保持；候选库 77 清理，旧 Active 用新 Session 恢复 |
| `committed` | CommitCandidate 成功后、ACTIVATE 前杀 Core | 仅新 Active 恢复，DB=77；旧库 41 保持，已提交资源继续被引用 |
| `first-prepared` | 首次 enable 的 PREPARED 后杀 Core | 无旧 Active，保持未激活与 HTTP 503，回收未提交候选 |
| `ready` | 候选 OnReady 读取 `fail-ready` 后失败，再杀 Core | 新 Active/committed/Degraded 保持；带故障重启仍选新代次，移除故障后恢复 DB=77 |
| `gates-prepared` | 候选已准备、尚未提交 | 检查点释放前请求日志保持，释放后六类入口均由新 PID 处理 |
| `gates-committed` | 已提交、ACTIVATE 前 | 同上 |
| `gates-binding` | ACTIVATED 后、Core 的完整业务绑定前 | 同上 |

Core 使用 `ALONABOT_TEST_CRASH_PHASE` 与 `ALONABOT_TEST_CHECKPOINT=<子目录>/checkpoint` 定位窗口。崩溃分项看到 marker 后杀 Core；门禁分项通过 `<子目录>/checkpoint.release` FIFO 精确保持和释放窗口。门禁以独立插件的真实定时回调推进至少一个任务周期，配合 command/event/Service/Bus/HTTP 请求；正常绑定前后都验证六类入口实际执行。日志逐条记录 PID 与入口类型，交接后只接受新 PID。

恢复分项在 Core 退出、运行锁释放后读取真实状态库：记录 Journal 的 OpID/Kind/ExpectedOld/NewGeneration/Phase/Session/DB 路径，核对 Active、代次引用与资源文件，并只读查询业务 SQLite。API 响应、Core 原始日志、插件 `starts` 与 `requests` 文件保留。

### 残存写入者的退出与恢复

`survivor` 使用 `<data_dir>/started/`。在 Journal started 后、停旧进程前持有检查点，杀 Core；临时 SDK overlay 在断链退出时写入 `runtime-held` 并阻塞于 `hold-runtime` FIFO，旧插件继续持有真实 `runtime.lock`。重启 Core 后要求 Degraded、Active 不变、启动计数不增、DB=41；随后终止该自建旧进程，等实际锁释放，再经 CLI restart 恢复相同 Active/DB。未提交 Journal 变为 cleaned，候选目录回收，新 Session 仅激活一次。

故障进程、FIFO 与系统运行锁场景使用 Unix API，入口支持 macOS/Linux；其他目标平台的运行证据由 T-10 取证。临时 Core 与 SDK 检查点供本场景使用，构建和运行位置均写入证据。

## 读取执行证据

成功时输出 `PASS`，退出码为 0。失败时保留完整 Go 测试输出并返回其退出码；场景清理仅终止自己启动的进程，运行目录与产物保留。

| 位置 | 内容 |
| --- | --- |
| `<out>/scenario-evidence/` | 构建命令、输入与 overlay、构建原始输出 |
| `<data_dir>/scenario-evidence/inputs.json` | 指定 CLI、overlay Core 与包的绝对路径、执行平台、实际 Go 命令 |
| `<data_dir>/scenario-evidence/overlay.json`、`*_test.go` | 本次有效的测试入口与辅助源码，可检查输入替换及断言 |
| `<data_dir>/scenario-evidence/go-test.log` | 完整 stdout/stderr 与测试结果 |
| `<data_dir>/scenario-evidence/events.jsonl` | 管理输出、包与 HTTP 快照、Core 原始日志；恢复入口另记录子目录，残存入口记录旧 PID、启动计数与恢复 Journal |
| `<data_dir>/state.sqlite` 或 `<data_dir>/<子目录>/state.sqlite` | Core 已退出后的持久状态、代次与资源引用；Journal 及 DB 核对结果同时保留在 Go 日志 |
| 各运行目录的 `plugins/recovery/` | 配置与业务 DB、`starts`、`requests`、崩溃 checkpoint 及残存标记 |

CLI 的本机 HTTP 转发器沿用集成测试辅助，只把 Host 改为 Core 配置的 `base_url`，请求仍经真实 Bearer 认证与正式管理入口。按钮取证在 Core 退出后使用同一状态库，沿用集成测试的固定路由描述。

源码定位锚点变化时入口明确报错，更新场景入口后再运行。临时 overlay 保留原主路径断言，`TestMain` 在本场景中跳过常规 Core/夹具构建，使用指向指定二进制的临时符号链接。
