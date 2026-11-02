# 实施计划：WebUI

总目标、章节划分与跨章节依赖见 [PLAN](../PLAN.md)；本章节的 goal 语句见 [webui.md](../../docs/goals/webui.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本章节覆盖官方 WebUI 组件 `web/admin/`：前端骨架、配置表单与 i18 组件、管理页面、配置与权限页面，以及 WebUI 组件的静态文件服务与组件图标（T-08.7、T-09.8、T-09.9、T-09.12）。Admin API、管理认证与管理服务层在[核心实施计划](core.md)；T-08.7 的步骤文字在核心计划的 T-08 步骤 9，T-09.10 的日志页随日志管理留在核心章节。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-08.7 | goal-34 | `web/admin` 前端骨架（官方 WebUI 组件的源码），配置表单与 i18 组件及 fixture（选项图标用 `<img>`），界面文案的 en-US/zh-CN 目录与语言切换 | T-08 步骤 9（前端） | T-08.1、T-08.10 | V-25 的表单分项、V-54 的界面组件分项 | 未开始 |
| T-09.8 | goal-43 | Admin 页面：真实登录、状态、各角色组件的安装启停与替换、Adapter 实例管理、更新/恢复 | T-09 步骤 7（管理页） | T-09.6、T-08.7、T-09.12 | V-40、V-47 的页面分项 | 未开始 |
| T-09.9 | goal-44 | Admin 页面：配置表单与 JSON 编辑、命令权限、发信许可、Core 设置（Core Admin 列表、默认 locale）、工具调用策略、Agent 会话的列出与清除、插件页面代理、界面语言切换 | T-09 步骤 7（配置与插件页） | T-09.8、T-11.1、T-11.3 | V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56 的页面分项 | 未开始 |
| T-09.12 | goal-84 | WebUI 组件与组件图标：Admin host 提供 WebUI 静态文件与单页路由回退、未安装与版本不兼容的提示、Admin API 主版本与 `spec/ADMIN-API.md` 参考、带 sandbox CSP 的图标、库模式的 `alonabot.WebUI` 选项 | T-09 步骤 9 | T-07.4、T-09.1、T-09.4、T-13.1 | V-83 | 未开始 |

## 3. 工作包

### T-09 的 WebUI 步骤

以下步骤原属 T-09，编号保持不变。T-09 的前置输入、拟目录、验收与失败停止条件见[核心实施计划的 T-09](core.md#t-09安装分发正式激活恢复与管理入口)。

7. **完成最小 Admin 页面与 API 绑定。** 页面属于官方 WebUI 组件（`web/admin/`），只调用 Admin API。接入真实登录、状态/日志、安装启停、更新/恢复、配置表单/完整 JSON 编辑、命令权限、插件的发信许可（`allowed_adapter_instances`）、工具调用策略、Agent 会话的列出与清除，以及界面语言切换。Core 设置页管理 Core Admin 列表与默认 locale，与 `alonabot admin add|remove|list` 使用同一管理服务。使用 T-08 的 raw JSON/PATCH/base_generation/secret 语义，显示 Pending、候选失败、Degraded 与维护原因；动态权限复用 T-04。通过 T-07 的两个 host、Host 校验、CSRF、受限票据和代理身份处理提供插件页面；管理 Cookie/Authorization 不进入插件，单 host 不执行插件主动内容。页面只展示用户可操作状态和具体错误，不把内部协议步骤当作用户配置项。**日志：** 按来源的级别持久化在状态库，修改后经 `sdk.log.set_level` 通知 SDK 并立即生效；Core 为每个来源保留最近 1000 条记录，提供按来源、级别与 trace_id 的读取和 SSE 订阅，过去日期的记录从日志文件按日期查询；`alonabot logs`、`alonabot log-level` 与 WebUI 日志页使用同一 API，页面的筛选只在浏览器中进行（CONTRACTS §14.3、§14.4）。

9. **WebUI 组件与组件图标。** 按 CONTRACTS §7.5 在 admin host 提供已启用 WebUI 组件的静态文件：保留 `/api/` 与内部路径，不存在且无扩展名的路径回退 `index.html`，未安装时返回 503 与英文提示，Admin API 主版本不在 WebUI 声明的范围内时拒绝安装与启用并在 admin host 提示。Admin API 定义主版本号常量，并在 `spec/ADMIN-API.md` 写出全部接口、认证与错误格式，官方 WebUI 只使用其中的接口。图标只提供 manifest 与表单选项声明的路径，带 sandbox CSP 与 nosniff，并要求管理员会话。库模式增加 `alonabot.WebUI(fsys fs.FS)` 选项，以 dev 来源提供开发中的 WebUI。
