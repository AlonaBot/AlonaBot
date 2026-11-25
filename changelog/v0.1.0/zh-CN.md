## [0.1.0] - 2026-11-21

AlonaBot 首次正式发布。你可以用 Core 搭建机器人，再按需安装聊天平台 Adapter、功能插件和 Agent。

<a id="发行内容"></a>

### Added

- Go 开发者可以使用 `v0.1.0` SDK 编写独立安装的插件、Adapter 和 Agent。Telegram、OneBot 11 Adapter 和 help 插件均可单独安装；本版的五个官方组件按 AGPL-3.0-only 独立发布签名归档。
- [部署指南](https://github.com/AlonaBot/AlonaBot/blob/af4f68550dcef191d5bc9a5e345b48510196ae23/docs/deployment.md)提供原生安装和 Docker Compose 两种方式。镜像 `ghcr.io/alonabot/alonabot:0.1.0` 内含 Core、WebUI 和 Agent。Compose 首次启动时启用 WebUI；按启动提示领取一次性管理令牌，在 `http://127.0.0.1:7650` 登录。填写 Provider 和模型后即可启用 Agent。
- 更新组件时，Core 会保留来源、发布者绑定和锁定设置。回退会创建新的运行代次和数据库副本。普通卸载会保留组件数据和来源，以便从同一来源重新安装；卸载 Core 服务和程序也会保留数据目录。

### Changed

- 11 月 21 日，五个平台归档原地重包，只包含 Core 程序、`LICENSE` 和 `source.json`。Core 二进制、tag 和整合镜像保持原样，已经安装的组件继续保留。**如果手里仍是旧归档，请重新下载，旧摘要已失效。** 安装器会按 GitHub Release API 提供的摘要校验下载 ([#291](https://github.com/AlonaBot/AlonaBot/pull/291), [#292](https://github.com/AlonaBot/AlonaBot/pull/292))。
- 发行页提供 Linux 和 macOS 的 amd64、arm64 tar.gz 归档，以及 Windows amd64 ZIP 归档。五个归档均使用 `CGO_ENABLED=0` 构建，源码压缩包由 GitHub 提供。安装脚本从 `main` 获取，Compose 文件从版本 tag 获取。

### 已知限制

- OneBot 11 可以接收事件，发送会返回 `UNSUPPORTED`。welcome、rss 插件和[首发约定](https://github.com/AlonaBot/AlonaBot/blob/f0382910/spec/CONTRACTS-v0.1.md#152-发行包官方整合包与预装组件)列出的 Telegram 后续功能留待后续版本。
- Native 组件会执行管理员信任的本机代码。在库模式下，组件与 Core 共用进程；未捕获的 panic 会结束进程，停止超时后也无法强制结束 goroutine。声明 `prepare_effects=external` 的组件需要维护许可、备份许可和组件自己的外部数据恢复流程。
- 如需运行依赖 shell、Node 或 Python 的 MCP stdio 工具，请使用包含相应运行时的自定义镜像。

### 验证记录

拆仓前的整合候选在 darwin/arm64 原生、darwin/amd64 Rosetta、linux/amd64 虚拟机和 ARM 宿主的 Linux 容器中运行过。Intel Mac 原生运行、ARM Linux 虚拟机和实机未验证。Windows 交叉构建和静态检查通过；实跑和 CGO 验证于 2026-11-20 经维护者批准延期。业务联验使用本地 Telegram Bot API 和合成 Provider。逐项条件、失败和延期见[历史验收记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md)。

### 资源与依赖

历史 24/10/16 MiB 对照未达标，维护者接受现有开销，停止新增资源测量和压测。[资源实测记录](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/docs/benchmarks.md)保留完整条件和数据。测量使用 `github.com/AlonaBot/gorm v1.31.3-0.20261120022453-34b98340007d`；首发改用上游 `gorm.io/gorm v1.31.2`，产物体积和运行内存可能随之增加。

[0.1.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.0
