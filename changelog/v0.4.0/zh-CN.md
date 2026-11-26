## [0.4.0] - 2026-11-26

### Added

- 插件可以在工程根目录放一个 `assets/` 目录，打包时它会原样放进组件归档，不编进可执行文件。运行时用 `sdk.PackageDir(ctx)` 找到安装目录再读取，适合字体这类大文件；库模式下返回空串，插件需要自己准备降级方式 ([#344](https://github.com/AlonaBot/AlonaBot/pull/344))。
- 插件可以用 `sdk/packages` 列出已安装的组件，或按 `owner/repo` 从 GitHub Release 安装组件。只有触发这次调用的消息来自 Core 管理员时，Core 才会执行。安装视为管理员已确认来源信任与同类替换，并返回受影响的对象 ([#344](https://github.com/AlonaBot/AlonaBot/pull/344))。
- `sdk/command.List` 返回的命令额外带有所属插件的名称、说明（按调用者语言）以及命令参数，帮助类插件可以按插件分组展示用法 ([#344](https://github.com/AlonaBot/AlonaBot/pull/344))。
- 通过 Service 调用其他插件时，被调用方可以用 `c.ChatType()` 读到原消息的会话类型，与已有的 `c.User()`、`c.Session()` 一致 ([#344](https://github.com/AlonaBot/AlonaBot/pull/344))。

### Changed

- 组件通过 `tool.Confirm` 发起的操作，在管理员没有设置策略时直接执行，不再先向用户确认。需要确认的操作，请在 WebUI 的工具策略中改为 confirm 或 deny ([#344](https://github.com/AlonaBot/AlonaBot/pull/344))。
- 官方来源列表中的 help 插件换成 builtin 插件（`alonabot-plugin-builtin-official`）。已安装的 help 可继续运行，但安装和更新时会按普通来源确认信任 ([#344](https://github.com/AlonaBot/AlonaBot/pull/344))。

[0.4.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.4.0
