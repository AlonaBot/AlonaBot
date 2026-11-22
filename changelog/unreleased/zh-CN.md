# 未发布

## Added

- 新增按聊天与账号隔离的命令菜单：类型化 Set/Get/Delete/Status/Sync、手动就绪事件、所有权持久化及生命周期同步。
- Core 向 Adapter 准备媒体来源时保留可信 MIME，支持按已有媒体元数据识别 GIF。

- 提供组件与 Core 更新日志的管理 API。

## Changed

- 命令注册缺省只提供处理器；需要自动显示的根命令添加 `.Menu()`。Telegram 新菜单能力需要包含 CommandMenuController 的 SDK/Core，其他现有 Adapter 保留原 SyncCommands 入口。

- 提供中英文 README、语言互切、各语言文档首页与共享 Views 徽章，品牌图片移至 `docs/assets`。
- plugin/agent 初始工程直接回复 `Hello AlonaBot`，按需自行添加 i18n 目录；各角色生成的 Go 文件和 `plugin.yaml` 均无注释。

## Fixed

- Core 日志 API 返回正式版本内容，按 `Accept-Language` 选择英文或简体中文，缺省或未匹配时回退英文，并排除未发布变更。
