## [0.3.1] - 2026-11-23

### Fixed

- Core 在激活后生成配置编辑快照，重载前从 Active 配置补建缺失文件；成功重载后写回规范配置，重载、更新和回退时保留磁盘上尚未应用的编辑。
- 删除 Adapter 实例时清除持久化事件游标，同名重建的实例使用全新游标。
- 删除实例时同步清理 Profile 会话访问名单中的对应条目，保留访问模式和其他实例的条目。

### 升级与兼容性

- 升级 Core 至 v0.3.1，官方组件沿用 v0.3.0。Compose 使用 `ghcr.io/alonabot/alonabot:0.3.1`，整合 Agent 和 WebUI v0.3.0。
- 保留已有配置值、Secret 引用和受管数据，状态库维持 v3。升级前备份完整数据目录。

[0.3.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.1
