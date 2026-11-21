# AlonaBot v0.1.1

## Added

- 增加 `Context.ReadNodes`，按引用或合并转发的句柄读取标准消息节点，并为返回的媒体签发当前会话的读取权限。OneBot 11 v0.1.1 的类型化节点查询需要本版 Core。
- 增加通用 Markdown 消息片段，支持由 Adapter 映射到平台格式；`PlainText()` 可显式转换为纯文本。

## Changed

- 官方组件锁定 OneBot 11 v0.1.1 与 WebUI v0.1.1，help、Agent 和 Telegram 保持 v0.1.0。WebUI v0.1.1 提供商店搜索、排序、角色计数、九卡片分页和收录检查；收录申请通过浏览器打开 GitHub Issue 模板。
- 管理 token 以 `0600` 权限持久保存在 `admin_token`，启动与凭据初始化、重置只输出文件路径。本机 WebUI 支持领取和确认凭据，管理 CLI 自动读取所选数据目录中的凭据。
- 公开 Release 继续仅提供五个平台的纯 Core 归档，每个归档包含 Core、`LICENSE` 和 `source.json`。Compose 镜像使用 `ghcr.io/alonabot/alonabot:0.1.1`，整合 WebUI 与 Agent；旧版本资产和 tag 保留。

## Fixed

- 媒体 URL 读取沿用当前授权与挂载路径，节点查询中的媒体登记到调用会话，供后续读取。
- Core 管理 API 收录正式版本更新日志，使更新界面可以显示对应版本的发行说明。

## 升级与兼容性

升级前正常停止 Core 并备份完整数据目录，替换程序或更新固定版本的 Compose 镜像后，使用同一数据目录启动。安装记录、来源绑定、组件代次和 Adapter 实例配置继续保留。组件普通卸载保留数据与来源，可从同一来源重装。

管理 CLI 使用 `--data-dir` 替代 `--token-file` 或 `ALONABOT_ADMIN_TOKEN`；远程 Core 通过 SSH 或 `docker exec` 操作。若旧版登录已删除 `admin_token`，停止 Core 后执行 `alonabot admin-token reset --data-dir <data_dir>` 恢复本机 CLI 访问；重置会撤销已有管理会话。

OneBot 11 v0.1.1 可在 Core v0.1.0 启动；节点查询使用新增 Core 服务，旧版会返回 `Core service not found`，需要升级 Core。已有官方签名组件沿用其独立发行和验签安装流程，help 继续使用 v0.1.0。

本版在 darwin/arm64 完成五目标交叉构建与归档静态检查，并通过 CLI、HTTP 管理端交叉安装、启用、卸载和同源重装公开签名组件。公开 OneBot 11 v0.1.1 与 help v0.1.0 在真实 QQ 测试群完成 `/help` 回复；Core 重启后再次回复，安装记录、Active 代次和实例配置保留。

其余平台的实际运行覆盖与已批准延期沿用 [v0.1.0 发行记录](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/zh-CN.md)；Windows 实际运行和 CGO 延期保持原决定。
