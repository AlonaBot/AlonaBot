# AlonaBot v0.1.0

v0.1.0 当前为发布候选。Core、官方组件和发行工具已完成实现；正式 SDK tag、官方密钥签名、GitHub Release 与 GHCR 镜像待维护者发布。以下说明列出候选的实际验证范围和发行使用方法。

## 发行内容

五个目标分别提供 `alonabot-0.1.0-<os>-<arch>.tar.gz`，Windows 使用 `.zip`。归档包含 Core 与相邻的 `packages/`，WebUI 与 Agent 两个组件各有平台包、`release.json`、`release.json.sig` 和 `release.pub`。整合包 Release 配套 `SHA256SUMS`、`install.sh` 与固定版本的 `docker-compose.yml`。构建同时生成外部组件资产，其元数据发布方式待维护者按下述限制完成。Core 与整合的两个组件的版本均为 `0.1.0`，组件后续更新使用各自的版本号。

首次启动预装 Agent 与 WebUI，启用 WebUI；Agent 配置 Provider 与模型后由管理员启用。Telegram、OneBot 11 Adapter 和 help 插件从各自独立仓库按需安装。整合发行、Dockerfile 与根目录 `docker-compose.yml` 保留在 Core 主仓。OneBot 11 当前仅支持接收，发送返回 `UNSUPPORTED`；welcome、rss 及 CONTRACTS §15.2 列出的 Telegram 后续功能留待后续发行。

以下平台与资源数字记录拆仓前候选的历史验收；组件拆仓后的定向验证随对应 PR 记录。

| 目标 | 已取得的实际运行结果 |
| --- | --- |
| darwin/arm64 | Apple Silicon 原生；发行归档、安装更新链及真实 launchd 安装通过。 |
| darwin/amd64 | Rosetta 仿真通过；Intel Mac 原生未运行。 |
| linux/amd64 | x86_64 Linux VM 平台验收及真实 systemd 安装通过；候选镜像另经 Rosetta 仿真通过。 |
| linux/arm64 | ARM 宿主上的 Docker 容器验收通过；ARM Linux VM/物理机未运行。 |
| windows/amd64 | 交叉构建与发行资产静态检查通过；Windows 实际运行和 CGO 未运行，2026-11-20 经维护者批准延期。 |

平台候选的四个可用环境分别通过 1082 步、104 次 Core 运行，包含 100 次启停与恢复场景；CGO 数学库 fixture 在四个环境实际返回 9。发行候选通过五目标静态检查、darwin/arm64 安装更新链、归档与双镜像业务联验。Windows 归档补入控制台进程组修复后，四个组件包、清单及测试签名的 16 项静态检查通过，归档 Core 的反汇编确认进程创建标志含 `CREATE_NEW_PROCESS_GROUP`；其余四目标与双镜像复用原有效结果。候选提交、工具链、逐项结果与原始失败见[历史验收记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md)。

## 安装与首次使用

正式资产发布后，从 v0.1.0 GitHub Release 选择同一版本的资产，按[安装与部署](https://github.com/AlonaBot/AlonaBot/blob/main/docs/deployment.md)操作。当前可按 [README 快速开始](https://github.com/AlonaBot/AlonaBot/blob/main/README.md#快速开始)从源码运行本地示例。

预编译运行无需 Go 或 Node。Linux 自动安装需要 systemd 和 root，macOS 使用当前登录用户的 launchd。脚本依赖 POSIX sh、curl、tar 和 sha256sum/shasum，下载校验通过后才替换安装。

手动解压时保留 Core 与 `packages/` 的相邻关系，在解压目录运行 `./alonabot run --data-dir ./data`；Windows 对应 `alonabot.exe`。浏览器打开 `http://127.0.0.1:7650`，按启动提示领取一次性管理 token，登录后按需安装 Adapter 并配置实例。需要 AI 对话时，配置 Agent 的 Provider 与模型再启用。help 插件可独立安装。

Linux 服务的程序目录为 `/opt/alonabot`、数据目录为 `/var/lib/alonabot`；macOS 数据位于 `~/Library/Application Support/AlonaBot/data`。容器以 UID 10001 运行，使用独立 `/data` 卷、init 和 loopback 端口发布。镜像包含 Core、组件与 CA 证书；需要 shell、Node 或 Python 的 MCP stdio 服务应自行构建含相应运行时的镜像。远程管理使用 SSH 隧道或按[HTTP 入口说明](https://github.com/AlonaBot/AlonaBot/blob/main/docs/deployment.md#core-设置与-http-入口)配置 HTTPS 反向代理。

## 升级、回退与恢复

1. 升级前正常停止 Core 与全部组件，备份完整 `data_dir`，同时保存当前 Core、`packages/` 和部署配置。外部数据库与框架目录外文件按组件说明另外备份；具体范围见[备份与恢复](https://github.com/AlonaBot/AlonaBot/blob/main/docs/deployment.md#备份与恢复)。
2. 再次运行 `install.sh --version <版本>` 升级 Core，或替换手动安装的 Core 与完整 `packages/`；容器更新固定版本 tag 后重建服务并保留同一数据卷。设置、安装记录和组件数据保留，更高版本的内置组件先列为可用更新，由管理员确认。
3. 组件更新保持来源与发布者绑定，遵守 pin。用 WebUI 查看保留代次，或通过[代次 API](https://github.com/AlonaBot/AlonaBot/blob/main/spec/ADMIN-API.md)选择恢复目标，再执行 `alonabot rollback <组件ID> <代次ID>`；CLI 显示将舍弃的数据并要求确认。回退建立独立新代次与数据库副本。
4. 需要整套恢复时，保持 Core 与组件停止，恢复备份对应的程序、完整数据目录及外部数据，再启动并核对登录、组件状态和业务数据。恢复备份会舍弃备份之后的写入；升级失败时先保存完整错误与现场。

`install.sh --uninstall` 删除服务和程序，保留数据目录；组件的普通 `uninstall` 保留代次、数据与来源，允许同来源重装。`--purge-data`、`--purge` 的删除对象和确认规则见 [CONTRACTS §9.3](https://github.com/AlonaBot/AlonaBot/blob/main/spec/CONTRACTS-v0.1.md#93-cli-的最小完整动作)。

## 资源与依赖

维护者已接受现有资源开销并停止新增资源测量、压测和优化。固定 Linux 三轮 PSS 中位数为 Core-only **26.61328125 MiB**、echo **10.09375 MiB**、configdb GORM **16.21875 MiB**；历史 24/10/16 MiB 比较保留 FAIL。条件、完整数列和原始数据位置见[资源实测记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/docs/benchmarks.md)与[历史验收记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md#s-062-最新资源实测与-d-184-接受决定)。原四负载及 NoneBot2 对照取消新增运行。

构建使用 Go 1.27.1、WebUI 使用 Node 26.5.1 与锁定的 npm 依赖；平台验收的 Linux ARM 容器使用 Go 1.27.3。普通发行物为 `CGO_ENABLED=0`。GORM 保留 `gorm.io/gorm v1.31.2` 身份，来源固定为获批准的 `github.com/AlonaBot/gorm v1.31.3-0.20261120022453-34b98340007d`；补丁、传播方式与依赖清单见[源码开发](https://github.com/AlonaBot/AlonaBot/blob/main/docs/development.md#依赖管理)。

## 已知限制与待完成项

- Native 组件是管理员信任的本机代码；签名证明发布来源。库模式共享 Core 进程，未捕获 panic 会结束进程，停止超时无法强制结束 goroutine。
- 声明 `prepare_effects=external` 的组件需显式维护、备份许可及组件专用恢复流程；外部写入和已发送消息按各自业务处理。
- 候选业务联验使用本地 Telegram Bot API 与合成 Provider。最终候选的真实 Telegram、付费 LLM Provider 凭据未运行；Windows 实跑与 CGO 已批准延期，原计费失败和未运行项保留在历史验收记录。
- 首次发布前的组件 ID、状态结构与初始迁移直接更新；已有开发数据先保留完整备份，再按契约重建。
- 官方[插件模板](https://github.com/AlonaBot/alonabot-plugin-template)的修复位于 [PR #3](https://github.com/AlonaBot/alonabot-plugin-template/pull/3)，当前固定候选 SDK `v0.0.0-20261120071651-46c35d0eebcb`。网络下载与本地功能验证通过；公开模板 CI 下载私有主仓库 SDK 时认证失败，维护者决定待主仓库正式公开后再验收，正式发行前替换为正式 SDK tag。
- 官方组件各自使用独立仓库和 Release，分别上传自己的 `release.json`、签名、公钥及平台包。整合归档只预装 WebUI 与 Agent；正式公开下载与更新链仍须在组件 Release 发布后验收。
- 维护者仍须使用正式 SDK tag 与官方私钥构建签名资产，发布 GitHub Release 与 `ghcr.io/alonabot/alonabot:0.1.0`、`:latest`。当前候选使用测试密钥，操作步骤见[发行维护](https://github.com/AlonaBot/AlonaBot/blob/main/docs/releasing.md)。

候选核对程序的输入和结果格式见 [release-check](https://github.com/AlonaBot/AlonaBot/blob/main/scripts/release-check/README.md)。
