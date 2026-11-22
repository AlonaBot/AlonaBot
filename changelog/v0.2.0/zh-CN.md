## [0.2.0] - 2026-11-22

### Added

- 新增按聊天与账号隔离的命令菜单：类型化 Set/Get/Delete/Status/Sync、就绪事件、所有权持久化及生命周期同步。见 [PR #302](https://github.com/AlonaBot/AlonaBot/pull/302)。
- Core 向 Adapter 准备媒体来源时保留可信 MIME，支持按已有媒体元数据识别 GIF。
- 提供双语文档导航、社区贡献指引和私密安全报告入口。

### Changed

- Core 直接使用平台归档安装和更新组件。在线安装通过 GitHub Release API 选择资产并读取包内 `plugin.yaml`，保留仓库、身份、版本、平台、build-info、资源、pin 与依赖检查；本地安装和管理上传使用单个归档。见 [PR #305](https://github.com/AlonaBot/AlonaBot/pull/305)。
- 命令注册缺省只提供处理器；需要自动显示的根命令添加 `.Menu()`。Telegram 分域菜单需要采用 `CommandMenuController` 的 Adapter；现有 Adapter 保留 `SyncCommands`。
- 提供中英文部署、组件开发和贡献指南；协议与管理 API 说明并入使用指南，协议字段以 `.proto` 源文件为准。
- Core、SDK 与官方组件采用 MIT 许可证。
- Agent、WebUI、help、Telegram 与 OneBot 11 分别锁定其独立的 v0.2.0 发行。Agent 递归读取引用、转发节点及图片来源；WebUI 使用单归档上传；help 显式注册菜单项；Telegram 增加分域菜单、媒体映射与类型化动作。
- 公开 Release 提供五个纯 Core 归档：四个 `.tar.gz` 和一个 Windows `.zip`，源码压缩包由 GitHub 提供。Compose 使用 `ghcr.io/alonabot/alonabot:0.2.0`，整合 Agent v0.2.0 与 WebUI v0.2.0。

### Removed

- Core 安装流程移除发布者签名要求及独立组件清单。组件成功更新或重装后清除退役的发布者指纹，保留已安装数据与来源绑定。
- 移除 `pack --key/--asset-base` 和 `install --release/--signature/--public-key`。打包使用 `alonabot pack <directory> --out <output-directory>`，安装使用 `alonabot install --source AlonaBot/<repository> --yes` 或 `alonabot install <archive> --yes`。

### 升级与兼容性

升级前正常停止 Core 并备份完整数据目录，替换程序或更新固定版本的 Compose 镜像后，使用同一数据目录启动。安装记录、组件代次、来源绑定与 Adapter 实例配置继续保留；命令菜单持久化通过新增状态迁移实现。

先升级 Core，再更新 WebUI 与其他组件。本版可读取现有组件平台归档。官方组件 v0.2.0 Release 仅提供平台归档；历史 `release.json`、`release.json.sig` 与 `release.pub` 在 Core、WebUI 验证后清理。清理后旧 Core 在线安装与更新需要升级；已安装组件继续使用原有数据。

自动命令菜单仅收录通过 `.Menu()` 显式声明的条目。已发布插件需要使用本版 SDK 添加 `.Menu()` 并重新构建，才能恢复自动菜单项；命令处理器继续可用。分域菜单控制需要同步升级 Core 与相应 Adapter。

Windows 实际运行、CGO 与资源测量的已批准延期沿用 [v0.1.0 发行记录](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/zh-CN.md)；交叉构建与模拟外部服务维持各自原有的验证范围。

[0.2.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0
