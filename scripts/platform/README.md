# 平台验收

`platform` 消费明确指定的 Core 与当前平台组件安装包，保存运行身份、命令和结果。当前入口覆盖首次认证、真实 echo 消息往返、showcase HTTP 与一次正常退出；整体状态为 `incomplete`，100 次启停、异常与子树、完整恢复及 CGO 验收继续补齐。

先从候选源码构建验收程序、Core 与三个组件包。以下为本机 fish 命令；交叉构建时分别为验收程序与 Core 设置目标，并为 `pack` 指定同一 `--target OS/arch`。

```fish
set artifacts (mktemp -d /tmp/alonabot-platform-build.XXXXXX)
go build -o "$artifacts/platform" ./scripts/platform
go build -o "$artifacts/alonabot" ./cmd/alonabot
"$artifacts/alonabot" pack examples/plugins/echo --out "$artifacts/packages/echo"
"$artifacts/alonabot" pack examples/plugins/showcase --out "$artifacts/packages/showcase"
"$artifacts/alonabot" pack adapters/telegram --out "$artifacts/packages/telegram"
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
| `--packages` | 当前平台组件包目录；分别包含 echo、showcase、alonabot-adapter-telegram 的唯一目标 `.tar.gz`，每包保留相邻 `release.json`；递归记录归档路径与长度 |
| `--scenarios` | goal45 的 `scripts/scenarios` 目录 |
| `--out` | 新建 JSON 文件；父目录须存在，已有文件会被拒绝 |
| `--environment` | `native`、`vm`、`container` 或 `emulation`，必填 |
| `--host` | 操作者填写的宿主或 VM 说明，单独记为 `host_description` |

程序为每次运行创建独立的临时 `cwd` 与 `data` 目录，并在 JSON 中留下绝对路径。命令从该 `cwd` 运行；输入路径按启动位置解析为绝对路径。程序和 Core 应构建为相同目标；Rosetta 验收时两者及组件都使用 `darwin/amd64`，环境填 `emulation`，宿主说明填写 Apple Silicon/Rosetta。容器填 `container`，同时说明实际宿主与容器发行版。

JSON 包含运行时 OS/架构/Go 版本、系统版本原始输出、Core 与组件包的 Go build info、包路径与长度，以及各步骤的 UTC 时间、毫秒耗时、完整合并输出、退出码和错误。`host_description` 由操作者声明，系统命令输出与编译目标另行记录。`runs` 保存 Core 命令、PID、动态监听地址、完整日志位置和组件激活 PID/代次；`bot_calls` 保留本地 Bot API 模拟服务收到的真实请求。

首次启动无 TTY，经 `admin-token show` 取得凭证后，使用 Cookie/Origin/CSRF 登录及启用组件；安装走真实 CLI Bearer 认证。程序显式配置受控的 `base_url`，按现有集成测试的方式转发 Host 到实际动态 Listener。管理 token、登录请求和成功登录响应的凭证值以 `output_redacted` 标记遮蔽；CLI 的私有 token 文件留在临时目录，供故障复现。首次登录后核对 `admin_token` 已删除。

Telegram Adapter 与 echo、showcase 均为独立真实组件进程；Bot API 为验收程序内的 loopback 模拟服务。发送一条 `/echo` 后核对回复正文与会话，再访问受认证的 showcase `/status`、页面与桥脚本引用。观察期限沿用现有真实进程集成测试的 20 秒与停止 15 秒；等待以实际响应、进程退出和日志为依据。当前正常退出通过 Unix interrupt 执行；Windows 原生停止和恢复路径后续补齐，实际运行保持未运行。

退出码 0 表示当前已执行步骤成功；整体验收状态以 JSON 的 `status` 为准。步骤失败时退出非零，JSON 标为 `failed` 并保留原始错误。缺少 `--out`、参数解析失败或结果文件创建失败时，错误直接输出到 stderr。运行目录与证据文件保留供排查；本程序无需新增依赖或安装工具。
