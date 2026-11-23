## [0.3.0] - 2026-11-23

### Added

- Docker 部署首次从本机浏览器打开时直接领取管理令牌：容器内（`ALONABOT_CONTAINER=1`）额外信任默认网关来源，其余检查不变；登录可选择在这台设备上保持。见 [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312)。
- 从 GitHub 源码构建安装组件（`Build: true`）：按需下载 Go 工具链到数据目录，异步任务提供步骤与构建日志，构建产物仍走安装预检与来源信任。见 [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312)。
- 下载设置：GitHub 下载镜像前缀、API 地址、Go 工具链下载地址、GOPROXY 与网络代理；经镜像下载的 Release 资产按 GitHub 给出的 sha256 摘要校验。见 [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312)。
- 配置文件可选的会话访问：全部、只允许所选会话或屏蔽所选会话；不设置时行为不变，库模式不受影响。见 [PR #315](https://github.com/AlonaBot/AlonaBot/pull/315)。
- Adapter 可声明账号配置页面（`page_purpose = instance_config`），WebUI 在添加或编辑账号时打开，页面经桥交回配置补丁；页面套件新增 `getConfigContext`、`completeConfig` 与按需加载的 `qrcode.js`。见 [PR #317](https://github.com/AlonaBot/AlonaBot/pull/317)。
- Adapter 实时操作目录、管理认证的统一调用 API、`sdk/debug` 客户端与 `alonabot adapter instance list/actions/call`；plugintest 支持多 Adapter/实例、可配置平台结果、异步 Submit/Wait、经管理认证注入事件等。见 [PR #310](https://github.com/AlonaBot/AlonaBot/pull/310)。
- `alonabot --version` 显示版本、构建日期与 Release 地址。见 [PR #309](https://github.com/AlonaBot/AlonaBot/pull/309)。
- SDK 新增 `OnSleep`、结构化 `llm.Generate`、`agent.Run`、AI 钩子与 Goal 接口及流式服务协议。见 [PR #308](https://github.com/AlonaBot/AlonaBot/pull/308)。

### Changed

- 插件与 Agent 配置 Schema 文件改名为 `plugin_config_schema.json`；Adapter 保留 `adapter_config_schema.json`。
- 命令 Builder 按匹配方式拆分，只有路径命令提供 `Menu`、参数和别名；`llm.Generate` 取代 `Ask` 等旧入口。见 [PR #308](https://github.com/AlonaBot/AlonaBot/pull/308)。
- 受管源码构建串行编译并设置 `GOMEMLIMIT`、`GOGC`；Linux 上内存不足时内核优先结束编译进程而非 Core；可用内存偏低或长时间无输出时在构建日志提示，进程被杀时任务报告可能内存不足并给出替代做法。见 [PR #316](https://github.com/AlonaBot/AlonaBot/pull/316)。
- 包列表返回名称与简介，账号实例返回所属组件状态。见 [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312)。

### Fixed

- 更新处于 Degraded 状态的组件后，新代次不再被立即停用。见 [PR #316](https://github.com/AlonaBot/AlonaBot/pull/316)。

### 升级与兼容性

- 用新 SDK 构建的组件需要 Core v0.3.0。升级前停止 Core 并备份完整数据目录；先停用旧组件，升级 Core，再把各组件更新到 0.3.0 并启用。配置值、秘密和受管数据保留。
- 状态库自动迁移到 v3（新增会话访问表）。
- 官方组件锁定各自的 v0.3.0 发行；Compose 使用 `ghcr.io/alonabot/alonabot:0.3.0`，整合 WebUI 0.3.0 与 Agent 0.3.0。

[0.3.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.0
