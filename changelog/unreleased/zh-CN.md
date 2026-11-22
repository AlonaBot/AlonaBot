# 未发布

## Added

- 新增按聊天与账号隔离的命令菜单：类型化 Set/Get/Delete/Status/Sync、手动就绪事件、所有权持久化及生命周期同步。
- Core 向 Adapter 准备媒体来源时保留可信 MIME，支持按已有媒体元数据识别 GIF。

## Changed

- 命令注册缺省只提供处理器；需要自动显示的根命令添加 `.Menu()`。Telegram 新菜单能力需要包含 CommandMenuController 的 SDK/Core，其他现有 Adapter 保留原 SyncCommands 入口。

- 为用户使用与开发指南提供中英内容，短导航和示例说明采用同页双语。内部维护与代理说明保留原有单语言；贡献与编码规则合并，共用示例操作步骤。删除退役的 `spec/` 目录，将协议与管理 API 帮助整合到已有指南；协议字段仍以 `.proto` 源文件定义为准。
