# PDF 之后的 AlonaBot 讨论归档审计

> 历史审计记录（2026-10-19）。文中的“待定/需补齐/当前版本”描述保留当时状态；现行处置见 [DECISIONS.md](../DECISIONS.md) 的 P/Q 处置表 与 [SELF-REVIEW.md](SELF-REVIEW.md)。本轮未重新取得旧 PDF 原件。

**版本：** v0.2 · 2026-10-19（追加独立自审发现）
**目的：** 对照 2026-10-11 旧版 PRD **之后**的产品讨论与当前 `spec/` 文档，而非再做一次旧 PDF 对照。主实施计划在 [PLAN.md](../../PLAN.md)，既有 PDF 对照在 [LEGACY-PRD-COMPARISON.md](LEGACY-PRD-COMPARISON.md)。

**完整性声明：** 此次以当前可取得的 PDF 后设计讨论上下文、历史检索到的用户明确要求及当前工作区文档核对，**不是所有聊天逐字备份**。表格 01～37 为先前覆盖审核，38～42 是本次遗漏自审补项（部分来源旧 PDF 的规范要求，并非都源自 PDF 后新增决定）。不得把建议提升为已确认；发生先后冲突优先采用较新的明确用户决定。

## 一、PDF 后核心讨论清单（01～37）与本次补查（38～42）

| 分项 | 讨论后形成的要求 | 本轮前覆盖情况 | 本轮处理 |
| --- | --- | --- | --- |
| 01. 总体目标 | 原生 Go、超轻量、低内存、多平台；与 NoneBot2 同负载对比 | 已有 | 保持；性能没有经真实产品验证 |
| 02. 插件运行时 | 最终改为 **仅 Native Go 独立子进程**，放弃 WASM 双运行时 | 已有 | 保持；WASM 方案只做历史背景 |
| 03. 插件更新 | Native 可启停、卸载、升级；失败可退回旧版 | 已有 | 保持，升级 Journal 属拟议实现 |
| 04. 插件根类 | 一个插件包只允许一个主类（可嵌入 plugin.Base）；可分散多个 Handler 文件/辅助结构体 | 已有 | 补充“不要求 Go ID() 方法”，元数据来自 YAML |
| 05. Config 类 | 一个插件包至多一个配置类；重复声明构建失败 | 大致已有 | 补充显式构建失败语义 |
| 06. Handler 自动/显式 | **包级独立函数**可经注解生成注册；**实例方法**必须有实例后显式注册；可混用 | 已有 | 补充构建期只扫描本插件代码，不扫描 Go Module 依赖 |
| 07. Handler 的根类访问 | 辅助 Handler 显式传主插件实例指针，访问主类属性/方法；不要求自动 DI | 已有 | 加回明确示例 |
| 08. 命令类型 | 普通命令、Regex 命令、普通消息/事件 Handler、多个别名、子命令/多级组 | 部分已有 | 补充普通消息 Handler 区别与缺少的筛选语义 |
| 09. 自动参数解析 | string/int/float/bool、引号、可选/默认/raw；失败默认用法回复但可覆写 | 部分已有 | 详细整理；路由优先级仍待冻结 |
| 10. 动态权限 | 插件/命令组/子命令权限；子命令可放宽或收紧组级，Core 全局规则不能绕过 | 已有 | 保持；不向 LLM 等路由回落 |
| 11. 事件过滤 | 集中/分散注册、event type 过滤；后来采纳 AstrBot Hook + NoneBot2 Rule/Permission/StopPropagation | 已有但不完整 | 事件 Hook、同优先级异常行为仍是待讨论 |
| 12. 主动发送 | `Reply` → `Send`；支持循环多次发送、跨 Adapter 会话、消息链和回执 | 已有 | 维持新 PDF 补充的消息模型 |
| 13. 生命周期 | 可选 OnStart/OnStop；加载、准备就绪、卸载/更新阶段由 Core 统管 | 部分已有 | 标出“hooks 精确列表未冻结”，避免伪造空生命周期方法 |
| 14. 插件 ID | 市场必须 `alonabot_plugin_` 前缀；本地允许但警告；ID 唯一且是 HTTP 前缀 | 已有 | 增补前缀格式约束及重复行为区别 |
| 15. **重复插件 ID** | **先注册的保留**，后同 ID 不得覆盖原插件；安装要显式提示冲突 | **文档错误：写成拒绝所有重复 ID** | **纠正：插件 ID first-wins**；不自动关联其他仓库 |
| 16. **重复路由** | 早期曾倾向 first-wins，后确认混用注册模式下 Handler 冲突报错 | 已有但未区分历史 | **以较新决定为准**：同插件 Handler/HTTP 注册冲突报错，不能静默覆盖 |
| 17. 跨插件协作 | Go Module 用于普通编译期复用；运行时通过 Event Bus、Service Call，经 Core Router；可嵌套但不循环 | 部分已有 | 补充编译期库和运行时调用的边界 |
| 18. HTTP | Core 共享监听器，按插件 ID 前缀注册；插件可以写 OAuth/Webhook/页面 | 已有 | 保持 Admin UI 不同 Origin 安全约束 |
| 19. 数据库后端 | SQLite/PostgreSQL 首发；统一 GORM Open；每插件 sqlite 文件/PG schema | 已有 | 补充 Open 的开发者入口与 schema 权限提醒 |
| 20. **数据库迁移** | 可不使用；开发者手动 `Migrate()` 或 Core 辅助 `Migrations()`；**两者互斥**；不干涉跨库搬迁 | **缺具体接口** | **补齐**；迁移与 Native Go 可以使用原生驱动 |
| 21. 配置文件 | JSON；默认保留未知键、仅缺字段填默认；允许手动编辑；OnLoad/Validate 可选 | 已有 | 保持新决定覆盖旧 PDF prune 默认 |
| 22. WebUI Schema | Go Config 结构体驱动 json/form/description/secret/hidden；只有必要字段 | 已有 | 增补单配置类校验、字段排序固定 |
| 23. i18 翻译 | **目录必须命名 `i18/`**；语言 JSON 顶层 `config`，再嵌套字段 label/description；不要再加 i18n tag | 已有但无示例 | 增补确切 JSON 结构示意 |
| 24. Secret/Hidden | secret 掩码，授权后可临时查看；hidden 仅在 WebUI 隐藏，手工 JSON 可编辑 | 已有 | 保持；不以隐藏代替真实权限限制 |
| 25. 调度 | Core 拥有触发权；不同任务并发，同任务是否跳过重叠仍待确认 | 已有 | 保持“不调 UI 并发上限” |
| 26. Native 内存 | 原型测了 20 个相同 worker；真实多依赖二进制/SQLite/CGO 尚未测 | 已有 | 确保 PSS 不是生产承诺 |
| 27. 安装与分发 | 预编译优先，源码编译可选；五类首发 OS/arch；平台差异需核验 | 已有 | 保持，market 来源核对与更新回滚 |
| 28. IPC | 本地 TCP 默认，stdio 可选，Protobuf 同协议，Session/CallID，stdout 日志问题 | 已有 | 保持；第三方原始 FD 写入仍是风险 |
| 29. 崩溃恢复 | Core 检测退出、有限重启，业务 RPC 不自动重试 | 已有 | 保持，重启阈值按历史认可的默认建议 |
| 30. Event Bus | 非持久化、尽力投递，失联不补发 | 已有 | 保持；不能悄悄改成持久消息队列 |
| 31. Service Call | Core 路由，可嵌套 A→B→C、禁止同步环，超时/取消 | 已有 | 保持，与过滤器链不同 |
| 32. 新事件管道 | AstrBot 阶段 Hook+优先级，NoneBot2 Rule/Permission+Stop；同事件同步排序 | 已有 | 保持，优先级/异常尚待精化 |
| 33. Adapter 规划 | Telegram 优先；OneBot、Discord、QQ 官方后续；已有 Adapter 开关，Adapter 类型运行时热加载非首发目标 | 部分已有 | 补充区别：开关 ≠ 动态加载 Adapter 实现 |
| 34. LLM 长期路线 | 初期最小聊天；未来多模型 Provider、Claude 式任务/子代理方向，非 V0.1 | 只写了非目标 | **补“未来方向”**，不提前开发 |
| 35. 开发者体验 | 尽量少用反射、完全原生 Go SDK、标准库与普通依赖，不引入逐插件大堆配置开关 | 部分已有 | 增补减少反射、使用生成器的原因 |
| 36. 文件 / 进程 | 插件可读文件、HTTP 下载和进程调用；Native 不构成安全沙箱 | 已有 | 保持权限风险，不假称 Host 强隔离 |
| 37. 用户推进方式 | 先讨论细节、逐步确认，再统一 PRD 与实施；不应主动扩大需求 | 文件约定已有 | 继续保持 Draft，后续迭代局部更新 |
| 38. 动态调用另一个插件命令/事件 Router | 较早讨论明确保留了**公开接口注册 Service Call + Core 动态 Router 调用**两条协作路径 | **遗漏：只有 Service Call/Event Bus** | 在 DESIGN §G 补回第二路径；权限与是否 v0.1 尚待批准 |
| 39. 事件类型和系统任务 Actor | 旧 PDF 已列 message.edited/deleted、button、member、reaction、scheduled；无 Actor 的任务不得默认 Admin | **遗漏事件清单和身份规则** | 在 DESIGN §E/F/H 与 PLAN T-05 补充；这些是旧版承接，不是 PDF 后新确认 |
| 40. 生命周期 reload/ready Hook | 较早讨论希望覆盖加载、Ready、重载、卸载等阶段；最新公开示例只保留 OnStart/OnStop | **过度简化** | 保留可选阶段，具体 Hook 名称/顺序标待确定 |
| 41. Config 包装层 `config` | 历史讨论**建议**运行时 JSON 可用顶层 `config`；翻译文件 `i18/*.json` 顶层 `config` 则已确认 | **两者被混同** | 区分确认与提议，普通配置文件外层结构待确认 |
| 42. 多账号 Target/凭证隔离与 Native 安全 | 用户曾关注多个账号/凭证隔离；Native Secret Store 不能自然强隔离恶意插件，外部子进程要回收 | **缺少适用性边界说明** | 保留为 Adapter/插件实例设计要求或扩展候选，不强加 X/Twitter Core 功能 |

## 二、两项容易混淆的历史冲突

### C-01 插件 ID 重复 vs 路由重复

- **插件包 ID 冲突**：首次注册的插件保留，后续同 ID 不应覆盖或错误绑定不同来源；应报告冲突，并在安装/市场层阻止静默替换。
- **同插件的 Handler/HTTP 路由冲突**：较新讨论已确定应在注册/构建时报错；不采用早期“只保留第一个重复路由”的默默跳过行为。
- 这两种冲突的判断位置和用户体验不同，不能都写成“重复即拒绝”，也不能都做 first-wins。

### C-02 旧版迁移/Host 接口 vs Native 运行时

- `Migrate()` 自己执行或 `Migrations()` 交给 Core 协助版本调度是插件作者可选的迁移**接口风格**；当前不得把它理解成必须使用 WASM Host DB 或自动保证所有 DDL 回滚。
- Native 进程可绕过 Core 的 SDK 做文件读写、HTTP 或执行命令；原旧版 Host FS/HTTP/Exec 不是系统级沙箱。

## 三、仍需逐条设计而非假装定稿

以下不能从讨论摘要推导为已确定实现：

- RPC proto 精确字段、循环检测与取消/确认语义；大媒体传输与最大帧长度。
- 同一事件 Handler 优先级相同的顺序、Hook 细节、错误时是否继续传播。
- 同一定时任务重叠/错过触发、多轮会话状态存储。
- 插件版本对 Core SDK 的兼容契约与跨平台 CGO 签名/构建细节。
- Adapter 作为第三方扩展如何在 Native-only 系统中部署。
- Native 插件的 OS 级沙箱与可信市场安全模型。
- 不同 Native 二进制进程的真实内存与 SQLite/PostgreSQL 负载测试。
- LLM/Agent 子代理具体 interface 与 Provider 生命周期（后续版本）。
- Service Call 之外是否作为 V0.1 暴露通过 Core 的动态命令路由调用。
- 定时任务无 Actor 时的权限机制实现，以及 Adapter 身份不可被插件伪造。
- 配置 JSON 本身的顶层包装层与 i18 翻译文件 `config` 顶层的区别、生命周期 reload 钩子的接口。

**审计范围声明：** 42 项为已恢复会话信息与旧 PDF 的需求级覆盖检查，不是全文逐字聊天归档；无法据此证明历史对话绝对无遗漏。当前更严格的独立校验记录在 [SELF-REVIEW.md](SELF-REVIEW.md)，后续如恢复逐字记录需继续核对。

## 四、这轮同步的文档

- `DESIGN.md`：补充 SDK 注册冲突、迁移接口、示意配置 JSON、命令解析、Go Module 与未来方向。
- `DECISIONS.md`：标出插件 ID first-wins 与路由冲突报错的差别，补齐原生迁移与使用 Go SDK 的约束。
- `PLAN.md`：引用本审计作为新轮补全的依据；不把未确认候选直接标为已批准。
- 本文件作为 PDF 后设计变更的独立索引，后续仅追加真正的新决策。
- v0.2：独立自审核对缺漏 38～42，已同步至 v0.4-draft 的 PLAN/DESIGN/DECISIONS，未把未经批准的方案变成已确认。
