## [0.2.0] - 2026-11-22

### Added

- Core 可以按聊天和账号同步不同的命令菜单。Adapter 作者可用 `CommandMenuController` 设置、查看、删除或同步菜单，并读取同步状态。分域菜单需要本版 Core 和实现该接口的 Adapter ([#302](https://github.com/AlonaBot/AlonaBot/pull/302))。
- Adapter 可以用 Core 提供的可信媒体类型识别 GIF，再按平台能力处理。
- 用户和组件作者可以分别查阅中英文部署、开发和贡献指南。社区文档也提供私下报告安全问题的入口。

### Changed

- 从 WebUI 上传组件或用 CLI 本地安装时，只需一个平台归档。在线安装和更新会从 GitHub Release 选择对应归档，再读取包内 `plugin.yaml` 核对组件和来源。官方发行清理旧清单和签名附件后，仍依赖这些附件的旧 Core 要先更新到本版，才能继续在线安装或更新；已安装组件可继续运行 ([#305](https://github.com/AlonaBot/AlonaBot/pull/305))。
- 根命令只有显式调用 `.Menu()` 才会自动出现在菜单里。若旧插件的命令需要继续显示在菜单中，作者应加上 `.Menu()` 并用本版 SDK 重新构建；原有命令仍可直接调用 ([#302](https://github.com/AlonaBot/AlonaBot/pull/302))。
- Core 和 SDK 改用 MIT 许可证。

### Removed

- 打包和安装改用单归档流程，移除了 `pack --key/--asset-base` 和 `install --release/--signature/--public-key`。请在脚本中改用 `alonabot pack <directory> --out <output-directory>`，安装时用 `alonabot install --source AlonaBot/<repository> --yes` 或 `alonabot install <archive> --yes`。组件更新或重装成功后，Core 会移除旧发布者指纹，保留数据和来源绑定。

### 验证记录

Windows 实跑、CGO 和新增资源测量沿用 [v0.1.0 的记录](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/zh-CN.md)，维持原延期决定。

[0.2.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0
