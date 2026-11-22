# 平台验收

`platform` 消费明确指定的 Core 与当前平台组件安装包，保存运行身份、命令和结果。Unix 与 Windows 入口覆盖首次认证、真实 echo 消息往返、showcase HTTP、100 次启停、Core 强杀后重启、合作型子树清理及完整事务恢复。传入 CGO 构建记录后还验证真实 C 函数调用；全部步骤成功才将该平台结果标为 `passed`，省略 CGO 输入时为 `incomplete`。四个 Linux/macOS 环境已完整通过；Windows 完成交叉构建，实际运行仍为「未运行」，2026-11-20 经维护者按 [Windows 实跑延期](../../changelog/v0.1.0/zh-CN.md#added) 批准延期。实际候选、环境与原失败见 [TEST-MATRIX §3.6](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md#36-五平台运行增量t-102--v-32v-34)。

先从候选源码构建验收程序、Core 与三个示例组件包，并准备下文约定的本地 `platform-tree` 插件。fixture 源码和产物放在 git 忽略的本地目录，随本次证据保留。以下为本机 fish 命令；交叉构建时分别为验收程序与 Core 设置目标，并为 `pack` 指定同一 `--target OS/arch`。

```fish
set artifacts (mktemp -d /tmp/alonabot-platform-build.XXXXXX)
go build -o "$artifacts/platform" ./scripts/platform
go build -o "$artifacts/alonabot" ./cmd/alonabot
"$artifacts/alonabot" pack examples/plugins/echo --out "$artifacts/packages/echo"
"$artifacts/alonabot" pack examples/plugins/showcase --out "$artifacts/packages/showcase"
# telegram_source 指向独立 alonabot-adapter-telegram checkout。
"$artifacts/alonabot" pack "$telegram_source" --out "$artifacts/packages/telegram"
# 将 tree_source 设置为本地 platform-tree 工程的绝对路径。
"$artifacts/alonabot" pack "$tree_source" --out "$artifacts/packages/tree"
# cgo_source 指向下文约定的 ignored 本地 CGO 插件 main.go。
uv run --no-project --offline --no-python-downloads scripts/platform/cgo_build.py \
    --root "$PWD" --source "$cgo_source" --target darwin/arm64 --out "$artifacts/cgo"
"$artifacts/platform" \
    --alonabot "$artifacts/alonabot" \
    --packages "$artifacts/packages" \
    --scenarios "$PWD/scripts/scenarios" \
    --cgo "$artifacts/cgo/build.json" \
    --out "$artifacts/result.json" \
    --environment native \
    --host 'macOS Apple Silicon'
```

参数与 `--help` 一致：

| 参数 | 输入 |
| --- | --- |
| `--alonabot` | 待验收 Core 的可执行文件路径 |
| `--packages` | 当前平台组件包目录；分别包含 echo、showcase、alonabot-adapter-telegram、platform-tree 的唯一目标 `.tar.gz`，每包保留相邻 `release.json`；递归记录归档路径与长度 |
| `--scenarios` | 候选源码 checkout 内的 `scripts/scenarios` 目录，含 `run.py` 与恢复模板 |
| `--cgo` | 本地 `platform-cgo` 的构建记录 JSON；可选，省略时整体验收为 `incomplete` |
| `--out` | 新建 JSON 文件；父目录须存在，已有文件会被拒绝 |
| `--environment` | `native`、`vm`、`container` 或 `emulation`，必填 |
| `--host` | 操作者填写的宿主或 VM 说明，单独记为 `host_description` |

程序为每次运行创建独立的临时 `cwd` 与 `data` 目录，并在 JSON 中留下绝对路径。命令从该 `cwd` 运行；输入路径按启动位置解析为绝对路径。程序和 Core 应构建为相同目标；Rosetta 验收时两者及组件都使用 `darwin/amd64`，环境填 `emulation`，宿主说明填写 Apple Silicon/Rosetta。容器填 `container`，同时说明实际宿主与容器发行版。

Rosetta 的构建工具可使用宿主原生 Go、Git 和 C 编译器，构建目标显式设为 `darwin/amd64`，验收程序经 `arch -x86_64` 启动。若 `/usr/bin/git` 在 Rosetta 中经 `xcrun` 加载 ARM 专用 `libxcrun` 失败，构建时从 `PATH` 查找已有 Command Line Tools 的原生 Git。直接调用该目录中的 Clang 时，`CC`／`CXX` 同时指定 `-isysroot` 与现有 macOS SDK；工具实际架构、完整参数及原始失败见 [TEST-MATRIX](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md) §3.6。

Docker Desktop 验收时，将 `TMPDIR` 指向容器本地文件系统中的目录，运行后把完整工作区复制到证据目录。共享挂载可保留构建输入和最终 JSON；既有对照发现，共享文件系统上的 Core 锁观测与容器本地文件系统不同，原始失败与对照结果见 [TEST-MATRIX](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md) §3.6。

JSON 包含运行时 OS/架构/Go 版本、系统版本原始输出、Core 与组件包的 Go build info、包路径与长度，以及各步骤的 UTC 时间、毫秒耗时、完整合并输出、退出码和错误。`host_description` 由操作者声明，系统命令输出与编译目标另行记录。`runs` 保存 Core 命令、PID、动态监听地址、完整日志位置和组件激活 PID/代次；`purpose` 区分首次业务、循环、强杀和恢复，`cycle` 记录 1–100，每一步的 `core_pid` 对应所属运行。`bot_calls` 保留本地 Bot API 模拟服务收到的真实请求。

首次启动无 TTY，管理 token 以 0600 权限持久保存在本次数据目录的 `admin_token`；经 `admin-token show` 显式取得凭证后，使用 Cookie/Origin/CSRF 登录及启用组件。安装走真实 CLI Bearer 认证，由 CLI 自动读取所选 `--data-dir` 的凭据。程序显式配置受控的 `base_url`，按现有集成测试的方式转发 Host 到实际动态 Listener。管理 token、登录请求和成功登录响应的凭证值以 `output_redacted` 标记遮蔽；数据目录及凭据文件留在临时目录，供故障复现。首次登录后核对 `admin_token` 仍保留。

Telegram Adapter 与 echo、showcase 均为独立真实组件进程；Bot API 为验收程序内的 loopback 模拟服务。发送一条 `/echo` 后核对回复正文与会话，再访问受认证的 showcase `/status`、页面与桥脚本引用。

管理会话核对每个端点的原始 HTTP 响应；跳转保留原状态、`Location` 和响应体，状态不符即判失败。Bot API 请求体读取失败时，错误同时保留完整原错与 `%q` 格式的已收字节，便于从失败 JSON 复现。

本地 `platform-tree` 工程使用当前候选 SDK，manifest 的 ID 为 `platform-tree`、role 为 `plugin`。`OnStart` 用 `os/exec` 启动同一可执行文件的 child 模式，继承进程组／Job；子进程保持一个 `127.0.0.1:0` TCP Listener 并接收连接，在其 stdin EOF 后关闭 Listener 并退出。父插件持有输入管道，`OnStop` 关闭管道并 `Wait` 回收子进程。父插件注册受认证的 `GET /state`，返回 `{"parent_pid":123,"child_pid":124,"address":"127.0.0.1:34567"}`；值取实际 PID 与 Listener 地址。Core 正常停止和断链停止均经过这一合作路径。

首次业务运行后，程序复用同一 data_dir 完成 100 次实际启动／停止。每次读取管理状态和激活日志，核对所有组件的持久 Active、代次与 PID，再访问子树状态。`running` 和 `released` 记录 Core、四组件、子 PID、Unix 进程组、五把锁及 Core／子进程监听端口：运行时要求存活、锁被持有且可连接；停止后要求 PID 和进程组消失、既有锁可真实取得、原端口可重新绑定。Unix 负 PID 表示进程组；僵尸仍视为存在。锁文件使用既有路径打开，检查结束即释放并保留文件。

循环结束后强杀 Core，Unix 核对 SIGKILL 退出与原始 `signal: killed`，Windows 核对 TerminateProcess 的退出码 1；确认旧树、端口和锁已释放后再重启，恢复时逐个比对相同 Active 并执行同样的停止检查。正常停止、Core 强杀、完整事务恢复与 CGO 的结果分别保留。观察期限沿用现有真实进程集成测试的 20 秒与停止 15 秒；每 20ms 按实际响应、OS 状态和取锁结果观察。Unix 使用 interrupt／SIGKILL／flock。Windows 使用定向 CTRL_BREAK、TerminateProcess、进程句柄等待与 LockFileEx 首字节锁；需从连接控制台的终端启动验收程序，Core 独立进程组接收正常停止信号，组件通过协议停止并由 Job 回收。

Unix 随后经 `os/exec` 运行[原有场景入口](../scenarios/README.md)，自动构建三版管理包与带检查点的 Core，再运行 `management`、`recovery` 七个窗口及 `survivor/started`。原入口的全部断言保留：管理流程的真实 DB 值、Active／Session 与五条 committed Journal；PREPARED／提交／OnReady 失败后的 DB、资源引用和候选回收；六类业务门禁；残存写入者持锁期间的 Degraded、启动计数和相同 Active／DB 恢复。

该阶段需要已有 Go、Git、gofmt 与 `uv`，使用 checkout 已批准的 Go 依赖。Python 通过 `uv run --no-project --offline --no-python-downloads` 执行。候选须含 `vcs.revision`；程序记录源码 HEAD，并用 Git 核对 `cmd`、`internal`、`sdk`、`contracts`、`go.mod` 与 `go.sum` 相对候选提交无变化。构建继承 `GOFLAGS`，显式指定验收程序的 OS／架构，并读取七个产物的 build info 核对目标。源码差异、工具缺失或构建失败均保留原错。

编译仍使用源码 checkout；运行场景时通过 `go test -exec` 的临时入口，将测试进程和其启动的 Core／CLI 切到独立 `cwd`。每个场景使用单独数据目录。运行中的场景收到验收取消后，先按原有测试清理自有 Core 和持锁插件，再返回取消错误；中断后需等待清理完成。

Windows 使用 `windows_sources.py` 加载原场景的源码生成函数，按 Go AST 选择入口引用的完整声明与原断言，再编译成五个原生 `scenario-test.exe`；实际执行通过 Go 测试程序完成。构建阶段保留源码 cwd，管理／恢复／残存写入者运行阶段在 TestMain 切到独立 cwd。归档入口使用 `.exe`，只读 DB 使用 Windows 文件 URI。检查点由命名 Event 阻塞和释放。残存写入者场景的临时 Core 为 Job 命名，验收程序保留额外 Job 句柄，Core 强杀后真实旧 SDK 仍持有 runtime.lock；原有 Degraded、启动计数、Active／Journal／业务 DB 断言继续执行，随后终止旧 PID 并验证恢复。该钩子只存在于保留的临时 overlay 中。

Windows 同样需要已存在的 Go、Git 和 uv，以及当前候选的全部目标组件包；Core 与验收程序使用 `.exe` 文件名。源码生成输入、有效 Go 声明、overlay、测试可执行文件及每场景运行参数均留在 `scenario-evidence`。Windows 实跑与 CGO 经维护者批准延期，重新安排时需补齐本地工具与平台运行结果；原始失败保存在上述历史验收记录中。

JSON 的 `recovery` 保存源码与候选提交、构建参数、七个产物的身份、三个场景的输入和完整 `events.jsonl`；每个场景另留 `go-test.log`、overlay、真实 `state.sqlite` 与业务数据库。程序同时核对所需测试名的成功记录和独立 cwd 标记，证据缺失即失败；部分事件和损坏输入的原始字节也保留在失败结果中。所有构建、运行目录与证据路径均位于 JSON 的 `workspace` 下，目录内容沿用原场景入口的布局。

CGO fixture 源码和产物留在 ignored 目录。它是使用当前候选 SDK 的独立 `package main`，以 `genimpl.Register("main", genimpl.Resources{})` 登记资源并经 `sdk.Run` 启动；真实 `GET /native?value=81` 路由解析输入、调用 C 数学库的 `sqrt`，返回 `{"value":9,"compiler":"C __VERSION__","os":"darwin","arch":"arm64","pid":123}`，平台、PID 和编译器字符串均取实际值。Linux／Windows 的 C 入口通过 `#cgo linux windows LDFLAGS: -lm` 链接数学库。`build/pack` 固定 `CGO_ENABLED=0`；本地 `cgo_build.py` 为此单独编译已有 CGO 源文件并生成 unsigned 安装包，fixture ID 固定为 `platform-cgo`、版本 `0.1.0`，使用标准归档、外置 `release.json` 和必需的完整性字段。

`cgo_build.py` 需要已有 Go、C 编译器与 uv，`--root` 指向候选源码 checkout，`--source` 指向上述单个 Go 源文件，`--target` 为本次 `os/arch`，`--out` 必须为尚未存在的目录。它保留 `go env` 的目标、CC/CXX、CGO 参数和 GOFLAGS，执行实际 C 编译器 `--version` 及 `go build -x -work`，完整命令、cwd、起止耗时、stdout/stderr 和退出码写入 `build.json`；Go 构建临时工作区也保留。宿主版描述器只读取目标文件的 build-info。构建或编译器失败仍写记录，保留原错；缺工具时先取得安装授权。

构建记录中的 `source`、`binary`、`archive` 均为绝对路径；`environment` 为上述构建环境对象，`compiler_version`、`build` 为含 `command`、`directory`、`started`、`duration_ms`、`output`、`exit_code`、`status` 和可选 `error` 的命令结果，`output` 完整保留合并的原始 stdout/stderr。本地构建器还写入同一格式的 `go_env`：探测失败时保留实际命令、原错、退出码及已知输入（含 `GOEXPERIMENT`），仍写 `build.json`。外部构建记录可直接提供环境值。验收程序优先报告已记录的环境命令失败，再检查编译器、构建与产物路径；核对成功状态与目标、实际二进制的 `CGO_ENABLED=1`，直接比对归档入口和二进制字节，再从 Mach-O／ELF／PE 读取直接动态依赖。PE 的 DLL 名从命名导入符号取得。未完成的 CGO 构建保持失败证据，平台支持范围由实际构建和运行结果确定。

CGO 运行单独记录为 `purpose=cgo`：在百次启停和强杀恢复结束后，再启动一次真实 Core，安装启用该包，经管理会话访问 C 函数路由并比对激活 PID。原有四组件一同运行，停止后检查五组件与合作子树、六把锁和端口释放。完整结果共 104 次业务 Core 运行；省略 CGO 时保留原有 103 次。`cgo` 保存构建记录、二进制身份、直接链接库和该项状态，HTTP 原始响应在 `steps` 中。

本地三种环境分别准备独立资产目录，Core、四个普通组件包和验收程序使用同一目标：macOS 原生为 `darwin/arm64`，Docker Linux 容器内为 `linux/arm64`，Rosetta 为 `darwin/amd64`。普通组件 `pack --target` 和 CGO 构建器 `--target` 均显式传入对应值；Linux CGO 在容器内使用其已有 C 工具链，Rosetta 的 x86_64 产物实际运行留在该执行环境核对。各环境的 `--packages`、`--cgo`、独立工作区和构建身份写入各自 JSON，交叉构建单列。

退出码 0 表示当前已执行步骤成功；整体验收状态以 JSON 的 `status` 为准。步骤失败时退出非零，JSON 标为 `failed` 并保留原始错误。缺少 `--out`、参数解析失败或结果文件创建失败时，错误直接输出到 stderr。运行目录与证据文件保留供排查；本程序无需新增依赖或安装工具。
