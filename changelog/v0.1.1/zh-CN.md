## [0.1.1] - 2026-11-21

### Added

- 插件可以用 `Context.ReadNodes` 读取被引用或合并转发的消息，再通过当前会话读取其中的媒体。节点查询使用本版新增的 Core 服务；调用旧 Core 会返回 `Core service not found` ([#293](https://github.com/AlonaBot/AlonaBot/pull/293))。
- 插件可以发送 Markdown 消息片段，由 Adapter 转成平台格式；需要纯文本时，可以显式调用 `PlainText()`。

### Changed

- Core 将管理令牌保存在权限为 `0600` 的 `admin_token` 文件中。启动、初始化或重置令牌时只显示文件路径。本机 WebUI 可以领取并确认令牌；管理 CLI 从 `--data-dir` 指定的目录读取令牌。请把脚本中的 `--token-file` 或 `ALONABOT_ADMIN_TOKEN` 改为选择数据目录。若旧版登录已删除令牌文件，先停止 Core，再执行 `alonabot admin-token reset --data-dir <data_dir>` 恢复访问；重置会注销现有管理会话。
- `alonabot init` 生成的插件和 Agent 示例会直接回复 `Hello AlonaBot`，作者可按需添加 i18n 文件 ([#295](https://github.com/AlonaBot/AlonaBot/pull/295))。
- 项目提供中英文 README 和各自的文档首页，读者可直接切换语言。

### Fixed

- 管理界面可以通过更新日志 API 读取已发布的 Core 说明。API 按 `Accept-Language` 返回英文或简体中文，缺省或无匹配时使用英文；未发布条目留在仓库 ([#295](https://github.com/AlonaBot/AlonaBot/pull/295))。

### 验证记录

本版在 darwin/arm64 完成五目标交叉构建和归档检查，并通过 CLI 和 HTTP 验证签名组件的安装、启用、卸载和同源重装。真实 QQ 群联验确认，Core 重启前后均能回复 `/help`。Windows 实跑和 CGO 仍按 [v0.1.0 的验证记录](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/zh-CN.md#验证记录)延期。

[0.1.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.1
