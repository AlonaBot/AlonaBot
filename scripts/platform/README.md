# 平台验收

`platform` 消费明确指定的 Core 与当前平台组件安装包，保存运行身份、命令和结果。当前入口完成候选身份与实际可执行性检查，整体状态为 `incomplete`；完整业务、进程生命周期与恢复验收随程序补齐后判定。

先从候选源码构建验收程序和 Core，将产物移至临时目录。以下为 fish 命令；`$example_packages` 指向当前平台 echo、showcase 及消息 Adapter 的真实安装包目录。

```fish
set artifacts (mktemp -d /tmp/alonabot-platform-build.XXXXXX)
go build -o "$artifacts/platform" ./scripts/platform
go build -o "$artifacts/alonabot" ./cmd/alonabot
"$artifacts/platform" \
    --alonabot "$artifacts/alonabot" \
    --packages "$example_packages" \
    --scenarios "$PWD/scripts/scenarios" \
    --out "$artifacts/result.json" \
    --environment native \
    --host 'macOS Apple Silicon'
```

参数与 `--help` 一致：

| 参数 | 输入 |
| --- | --- |
| `--alonabot` | 待验收 Core 的可执行文件路径 |
| `--packages` | 当前平台示例安装包目录，递归记录 `.tar.gz`、`.zip` 的路径与长度 |
| `--scenarios` | goal45 的 `scripts/scenarios` 目录 |
| `--out` | 新建 JSON 文件；父目录须存在，已有文件会被拒绝 |
| `--environment` | `native`、`vm`、`container` 或 `emulation`，必填 |
| `--host` | 操作者填写的宿主或 VM 说明，单独记为 `host_description` |

程序为每次运行创建独立的临时 `cwd` 与 `data` 目录，并在 JSON 中留下绝对路径。命令从该 `cwd` 运行；输入路径按启动位置解析为绝对路径。程序和 Core 应构建为相同目标；Rosetta 验收时两者及组件都使用 `darwin/amd64`，环境填 `emulation`，宿主说明填写 Apple Silicon/Rosetta。容器填 `container`，同时说明实际宿主与容器发行版。

JSON 包含运行时 OS/架构/Go 版本、系统版本原始输出、Core 的 Go build info（含存在的 VCS 设置和链接参数）、包路径与长度，以及各步骤的 UTC 时间、毫秒耗时、完整合并输出、退出码和错误。`host_description` 由操作者声明，系统命令输出与编译目标另行记录。组件目标与业务行为由后续实际安装和运行确认。

退出码 0 表示当前已执行步骤成功；整体验收状态以 JSON 的 `status` 为准。步骤失败时退出非零，JSON 标为 `failed` 并保留原始错误。缺少 `--out`、参数解析失败或结果文件创建失败时，错误直接输出到 stderr。运行目录与证据文件保留供排查；本程序无需新增依赖或安装工具。
