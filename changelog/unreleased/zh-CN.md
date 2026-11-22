# 未发布

## Added

- 新增按聊天与账号隔离的命令菜单：类型化 Set/Get/Delete/Status/Sync、手动就绪事件、所有权持久化及生命周期同步。
- Core 向 Adapter 准备媒体来源时保留可信 MIME，支持按已有媒体元数据识别 GIF。

## Changed

- 命令注册缺省只提供处理器；需要自动显示的根命令添加 `.Menu()`。Telegram 新菜单能力需要包含 CommandMenuController 的 SDK/Core，其他现有 Adapter 保留原 SyncCommands 入口。

- 为用户使用与开发指南提供中英内容，短导航和示例说明采用同页双语。内部维护与代理说明保留原有单语言；贡献与编码规则合并，共用示例操作步骤。删除退役的 `spec/` 目录，将协议与管理 API 帮助整合到已有指南；协议字段仍以 `.proto` 源文件定义为准。

- Agent、WebUI、Adapter、插件及模板的 Release 仅上传平台归档，源码压缩包由 GitHub 提供，停止发布 release.json 和签名旁车。
- Core 安装、更新与 WebUI 上传直接读取归档内 plugin.yaml。在线安装通过 GitHub Release API 选择资产，保留仓库、身份、版本、平台和资源检查；移除发布者签名要求，保留已安装数据和来源绑定。
- 移除 pack 的 --key/--asset-base 及 install 的 --release/--signature/--public-key。先发布并验证适配后的 Core 与 WebUI，再清理历史旁车；清理后旧 Core 在线安装和更新需要升级。
