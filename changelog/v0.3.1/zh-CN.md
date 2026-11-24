## [0.3.1] - 2026-11-23

### Fixed

- 首次启用组件后，Core 会生成可编辑的配置文件。重载时如果文件缺失，Core 会从当前生效的配置补建。成功重载后，Core 会将当前生效的配置写回文件；重载、更新或回退时会保留磁盘上尚未应用的编辑。见 [#321](https://github.com/AlonaBot/AlonaBot/pull/321)。
- 删除 Adapter 实例会清除它的旧事件游标，并从 Profile 的会话访问名单移除该实例的条目。同名重建的实例使用新的游标，访问模式和其他实例的条目保持原样。见 [#321](https://github.com/AlonaBot/AlonaBot/pull/321)。

Compose 镜像 `ghcr.io/alonabot/alonabot:0.3.1` 包含 Agent 和 WebUI v0.3.0。

[0.3.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.1
