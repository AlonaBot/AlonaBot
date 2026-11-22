## [0.2.0] - 2026-11-22

### Added

- 新增按聊天与账号隔离的命令菜单：类型化 Set/Get/Delete/Status/Sync、就绪事件、所有权持久化及生命周期同步。见 [PR #302](https://github.com/AlonaBot/AlonaBot/pull/302)。
- Core 向 Adapter 准备媒体来源时保留可信 MIME，支持按已有媒体元数据识别 GIF。
- 提供双语文档导航、社区贡献指引和私密安全报告入口。

### Changed

- Core 直接使用平台归档安装和更新组件。在线安装通过 GitHub Release API 选择资产并读取包内 `plugin.yaml`，保留仓库、身份、版本、平台、build-info、资源、pin 与依赖检查；本地安装和管理上传使用单个归档。见 [PR #305](https://github.com/AlonaBot/AlonaBot/pull/305)。
- 命令注册缺省只提供处理器；需要自动显示的根命令添加 `.Menu()`。Telegram 分域菜单需要采用 `CommandMenuController` 的 Adapter；现有 Adapter 保留 `SyncCommands`。
- 提供中英文部署、组件开发和贡献指南；协议与管理 API 说明并入使用指南，协议字段以 `.proto` 源文件为准。
- Core 与 SDK 采用 MIT 许可证；整合的 Agent 和 WebUI 沿用各自锁定版本内附的许可证。
- 公开 Release 提供五个纯 Core 归档：四个 `.tar.gz` 和一个 Windows `.zip`，源码压缩包由 GitHub 提供。Compose 使用 `ghcr.io/alonabot/alonabot:0.2.0`，整合 Agent v0.1.0 与 WebUI v0.1.1。其他组件保持独立版本：help 与 Telegram v0.1.0、OneBot 11 v0.1.1。

### Removed

- Core 安装流程移除发布者签名要求及独立组件清单。组件成功更新或重装后清除退役的发布者指纹，保留已安装数据与来源绑定。
- 移除 `pack --key/--asset-base` 和 `install --release/--signature/--public-key`。打包使用 `alonabot pack <directory> --out <output-directory>`，安装使用 `alonabot install --source AlonaBot/<repository> --yes` 或 `alonabot install <archive> --yes`。

### 升级与兼容性

升级前正常停止 Core 并备份完整数据目录，替换程序或更新固定版本的 Compose 镜像后，使用同一数据目录启动。安装记录、组件代次、来源绑定与 Adapter 实例配置继续保留；命令菜单持久化通过新增状态迁移实现。

先升级 Core。本版可读取现有组件平台归档，包括仍附带历史签名旁车的 Release。组件与 WebUI 的发布工作流已在各自仓库适配，后续发行仅需平台归档。发布并验证适配后的 WebUI，再清理历史 `release.json`、`release.json.sig` 与 `release.pub`；清理后旧 Core 的在线安装与更新需要升级。整合的 WebUI v0.1.1 仍使用其已发布的上传界面，适配版 WebUI 发布前，单归档本地安装请使用 Core CLI。

自动命令菜单仅收录通过 `.Menu()` 显式声明的条目。已发布插件需要使用本版 SDK 添加 `.Menu()` 并重新构建，才能恢复自动菜单项；命令处理器继续可用。分域菜单控制需要同步升级 Core 与相应 Adapter。

Windows 实际运行、CGO 与资源测量的已批准延期沿用 [v0.1.0 发行记录](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/zh-CN.md)；交叉构建与模拟外部服务维持各自原有的验证范围。

[0.2.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0
