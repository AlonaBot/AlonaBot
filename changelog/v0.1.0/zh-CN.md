## [0.1.0] - 2026-11-21

AlonaBot 首个正式版本。

<a id="发行内容"></a>

### Added

- 提供 Core 与 `v0.1.0` Go SDK，支持独立安装的插件、Adapter 和 Agent。
- 提供 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64 五平台归档。Windows 使用 ZIP，其余使用 tar.gz；普通发行物为 `CGO_ENABLED=0`。
- 提供 Core + WebUI + Agent 整合镜像 `ghcr.io/alonabot/alonabot:0.1.0` 与 Docker Compose。首次启动启用 WebUI；Agent 配置 Provider 与模型后由管理员启用。Compose 使用 UID 10001、init、持久化 `/data` 卷和 loopback 端口绑定。
- 提供按需安装的 Telegram、OneBot 11 Adapter 与 help 插件。五个官方组件均采用 AGPL-3.0-only 并独立发布签名发行物，`components.lock.json` 固定组件版本；镜像整合包验签后原样收录已发布组件资产。
- 组件更新保留来源、发布者绑定与 pin 设置，回退建立独立新代次及数据库副本。安装器卸载服务和程序时保留数据；组件普通卸载保留代次、数据与来源，允许同来源重装。

### Changed

- 2026-11-21 经维护者批准，v0.1.0 五个平台归档**原地重包为纯 Core**，仅包含 Core 可执行文件、`LICENSE` 和 `source.json`。原 Core 二进制与 tag 保留，Compose 镜像继续提供整合包，数据目录中已有组件继续保留。**旧归档摘要失效，请重新下载并按 GitHub Release API 的新资产摘要校验。**见 [PR #292](https://github.com/AlonaBot/AlonaBot/pull/292)。
- Release 附件仅保留五个平台归档，源码压缩包由 GitHub 自动提供。`deploy/install.sh` 从 `main` 下载，`docker-compose.yml` 从版本 tag 下载；安装器使用 GitHub Release API 的资产摘要校验官方下载。见 [PR #291](https://github.com/AlonaBot/AlonaBot/pull/291)。

### 安装与升级

按[安装与部署](https://github.com/AlonaBot/AlonaBot/blob/af4f68550dcef191d5bc9a5e345b48510196ae23/docs/deployment.md)操作。原生归档与安装器提供 Core，WebUI、Agent 和 Adapter 按需安装；Compose 预装 WebUI 与 Agent。v0.1.0 按启动提示领取一次性管理 token，WebUI 安装并启用后可在 `http://127.0.0.1:7650` 登录。远程管理使用 SSH 隧道或 HTTPS 反向代理。

从开发版本升级前，停止 Core 与全部组件，备份完整数据目录、程序、安装包、部署配置和外部组件数据；开发期间组件 ID、状态结构和初始迁移已有调整。重建容器时保留同一数据卷。升级、回退与整套恢复按[备份与恢复](https://github.com/AlonaBot/AlonaBot/blob/af4f68550dcef191d5bc9a5e345b48510196ae23/docs/deployment.md#备份与恢复)操作。

### 已知限制

- OneBot 11 支持接收事件，发送返回 `UNSUPPORTED`；welcome、rss 插件与 [当时发行约定](https://github.com/AlonaBot/AlonaBot/blob/f0382910/spec/CONTRACTS-v0.1.md#152-发行包官方整合包与预装组件) 列出的 Telegram 后续功能留待后续发行。
- Native 组件是管理员信任的本机代码。库模式共享 Core 进程，未捕获 panic 会结束进程，停止超时无法强制结束 goroutine。
- 声明 `prepare_effects=external` 的组件需显式维护、备份许可及针对外部副作用的组件专用恢复流程。
- scratch 镜像包含 CA 证书；需要 shell、Node 或 Python 的 MCP stdio 工具应自行构建含相应运行时的镜像。

### 验证记录

历史验收对应拆仓前整合候选：darwin/arm64 原生运行，darwin/amd64 使用 Rosetta，linux/amd64 使用 x86_64 VM，linux/arm64 使用 ARM 宿主上的 Docker；Intel Mac 原生与 ARM Linux VM/物理机未运行。Windows 交叉构建与静态检查通过，实际运行和 CGO 于 2026-11-20 经维护者批准延期。业务联验使用本地 Telegram Bot API 与合成 Provider，最终候选的真实 Telegram 与付费 LLM 凭据未运行。逐项结果、原始失败和延期项见[历史验收记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md)；组件拆仓后的定向检查随对应整合 PR 记录。

### 资源与依赖

维护者已接受现有开销并停止新增资源测量、压测和优化。历史 24/10/16 MiB 比较保留 FAIL，原四负载及 NoneBot2 对照取消新增运行；完整数字和条件见[资源实测记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/docs/benchmarks.md)。

历史测量使用 `github.com/AlonaBot/gorm v1.31.3-0.20261120022453-34b98340007d`；v0.1.0 使用上游 `gorm.io/gorm v1.31.2`，恢复上游后的产物体积与运行内存可能增加。构建使用 Go 1.27.1、Node 26.5.1 与锁定的 npm 依赖，Linux ARM 平台验收使用 Go 1.27.3。

[0.1.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.0
