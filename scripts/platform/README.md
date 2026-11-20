# 平台验收

`platform` 消费明确指定的 Core 与当前平台组件安装包，保存运行身份、命令和结果。当前入口覆盖首次认证、真实 echo 消息往返、showcase HTTP、100 次启停、Core 强杀后重启与合作型子树清理；整体状态为 `incomplete`，完整恢复场景与 CGO 验收继续补齐。

先从候选源码构建验收程序、Core 与三个示例组件包，并准备下文约定的本地 `platform-tree` 插件。fixture 源码和产物放在 git 忽略的本地目录，随本次证据保留。以下为本机 fish 命令；交叉构建时分别为验收程序与 Core 设置目标，并为 `pack` 指定同一 `--target OS/arch`。

```fish
set artifacts (mktemp -d /tmp/alonabot-platform-build.XXXXXX)
go build -o "$artifacts/platform" ./scripts/platform
go build -o "$artifacts/alonabot" ./cmd/alonabot
"$artifacts/alonabot" pack examples/plugins/echo --out "$artifacts/packages/echo"
"$artifacts/alonabot" pack examples/plugins/showcase --out "$artifacts/packages/showcase"
"$artifacts/alonabot" pack adapters/telegram --out "$artifacts/packages/telegram"
# 将 tree_source 设置为本地 platform-tree 工程的绝对路径。
"$artifacts/alonabot" pack "$tree_source" --out "$artifacts/packages/tree"
"$artifacts/platform" \
    --alonabot "$artifacts/alonabot" \
    --packages "$artifacts/packages" \
    --scenarios "$PWD/scripts/scenarios" \
    --out "$artifacts/result.json" \
    --environment native \
    --host 'macOS Apple Silicon'
```

参数与 `--help` 一致：

| 参数 | 输入 |
| --- | --- |
| `--alonabot` | 待验收 Core 的可执行文件路径 |
| `--packages` | 当前平台组件包目录；分别包含 echo、showcase、alonabot-adapter-telegram、platform-tree 的唯一目标 `.tar.gz`，每包保留相邻 `release.json`；递归记录归档路径与长度 |
| `--scenarios` | goal45 的 `scripts/scenarios` 目录 |
| `--out` | 新建 JSON 文件；父目录须存在，已有文件会被拒绝 |
| `--environment` | `native`、`vm`、`container` 或 `emulation`，必填 |
| `--host` | 操作者填写的宿主或 VM 说明，单独记为 `host_description` |

程序为每次运行创建独立的临时 `cwd` 与 `data` 目录，并在 JSON 中留下绝对路径。命令从该 `cwd` 运行；输入路径按启动位置解析为绝对路径。程序和 Core 应构建为相同目标；Rosetta 验收时两者及组件都使用 `darwin/amd64`，环境填 `emulation`，宿主说明填写 Apple Silicon/Rosetta。容器填 `container`，同时说明实际宿主与容器发行版。

JSON 包含运行时 OS/架构/Go 版本、系统版本原始输出、Core 与组件包的 Go build info、包路径与长度，以及各步骤的 UTC 时间、毫秒耗时、完整合并输出、退出码和错误。`host_description` 由操作者声明，系统命令输出与编译目标另行记录。`runs` 保存 Core 命令、PID、动态监听地址、完整日志位置和组件激活 PID/代次；`purpose` 区分首次业务、循环、强杀和恢复，`cycle` 记录 1–100，每一步的 `core_pid` 对应所属运行。`bot_calls` 保留本地 Bot API 模拟服务收到的真实请求。

首次启动无 TTY，经 `admin-token show` 取得凭证后，使用 Cookie/Origin/CSRF 登录及启用组件；安装走真实 CLI Bearer 认证。程序显式配置受控的 `base_url`，按现有集成测试的方式转发 Host 到实际动态 Listener。管理 token、登录请求和成功登录响应的凭证值以 `output_redacted` 标记遮蔽；CLI 的私有 token 文件留在临时目录，供故障复现。首次登录后核对 `admin_token` 已删除。

Telegram Adapter 与 echo、showcase 均为独立真实组件进程；Bot API 为验收程序内的 loopback 模拟服务。发送一条 `/echo` 后核对回复正文与会话，再访问受认证的 showcase `/status`、页面与桥脚本引用。

本地 `platform-tree` 工程使用当前候选 SDK，manifest 的 ID 为 `platform-tree`、role 为 `plugin`。`OnStart` 用 `os/exec` 启动同一可执行文件的 child 模式，继承进程组／Job；子进程保持一个 `127.0.0.1:0` TCP Listener 并接收连接，在其 stdin EOF 后关闭 Listener 并退出。父插件持有输入管道，`OnStop` 关闭管道并 `Wait` 回收子进程。父插件注册受认证的 `GET /state`，返回 `{"parent_pid":123,"child_pid":124,"address":"127.0.0.1:34567"}`；值取实际 PID 与 Listener 地址。Core 正常停止和断链停止均经过这一合作路径。

首次业务运行后，程序复用同一 data_dir 完成 100 次实际启动／停止。每次读取管理状态和激活日志，核对所有组件的持久 Active、代次与 PID，再访问子树状态。`running` 和 `released` 记录 Core、四组件、子 PID、Unix 进程组、五把锁及 Core／子进程监听端口：运行时要求存活、锁被持有且可连接；停止后要求 PID 和进程组消失、既有锁可真实取得、原端口可重新绑定。Unix 负 PID 表示进程组；僵尸仍视为存在。锁文件使用既有路径打开，检查结束即释放并保留文件。

循环结束后强杀 Core，核对实际 SIGKILL 退出与原始 `signal: killed`，确认旧树、端口和锁已释放后再重启；恢复时逐个比对相同 Active 并执行同样的停止检查。正常停止、Core 强杀、完整事务恢复与 CGO 的结果分别保留。观察期限沿用现有真实进程集成测试的 20 秒与停止 15 秒；每 20ms 按实际响应、OS 状态和取锁结果观察。当前 Unix 使用 interrupt／SIGKILL／flock；Windows 原生停止和观测路径后续补齐，实际运行保持未运行。

退出码 0 表示当前已执行步骤成功；整体验收状态以 JSON 的 `status` 为准。步骤失败时退出非零，JSON 标为 `failed` 并保留原始错误。缺少 `--out`、参数解析失败或结果文件创建失败时，错误直接输出到 stderr。运行目录与证据文件保留供排查；本程序无需新增依赖或安装工具。
