# 插件开发规范

负责范围：插件作者（含本仓库的官方插件、Adapter、Agent 与示例）在依赖、存储、日志、生命周期与发布上必须遵守的规则，以及这样要求的原因。Adapter 与 Agent 是有进程的非插件组件（DESIGN §T），这里的规则同样适用于它们；WebUI 组件只含静态文件，只需遵守第 9 节与 CONTRACTS §7.5。行为的权威是 [CONTRACTS](../spec/CONTRACTS-v0.1.md)，本文只列规则并指向对应条款，两者冲突时以契约为准。
更新时机：契约中影响插件作者的条款变化、SDK 新增按需包或依赖政策变化时。

SDK 仍在实施中，文中的包、配置项与命令以契约为准，尚未全部实现；实际进度见 [spec/README.md](../spec/README.md)。

## 1. 运行模型与内存

- 每个插件是一个独立的 Native Go 进程（D-01），只经 SDK 与 Core 通信。崩溃只影响它自己，但插件仍是可信的本地代码，不是沙箱（CONTRACTS §7.1）。
- 没有启用的插件不运行进程，启用后它的常驻内存全部算在用户机器上。实测一个只用 SDK 的插件进程约 7.5 MiB（linux/amd64 PSS），每多链接一个库、多做一次初始化都会增加这个数。官方组件的预算见 TEST-MATRIX V-74（D-52）。

## 2. 依赖

- 只导入用到的包。SDK 根包只允许依赖 protobuf、x/sys 与 jsonschema-go（D-52）；数据库在 `sdk/sqlitedb` 与 `sdk/gormdb` 中，不导入就不链接（D-49）。
- 能用标准库就不引入第三方库。HTTP 用 `net/http`；不要内嵌模型厂商的 SDK 或自己的 Provider 客户端，需要模型时调用 Agent 组件的 `llm.generate`（第 6 节）。
- 包级变量与 `init` 不做重活：不加载大文件、不建大表、不批量编译正则，推迟到 OnStart 或首次使用。所有被链接的包的 `init` 都会在启动时执行，分配的内存会一直常驻。
- 优先纯 Go。确需 CGO 时，必须为每个声明支持的平台提供真实可运行的二进制，不支持的平台在 manifest 中不声明（CONTRACTS §7.1、§9.1）。
- 插件的依赖由它自己的 `go.mod` 管理，与其他插件和 Core 互不影响；代价是每个依赖都完整地进入这个插件的进程。

## 3. 存储

- 业务数据放受管 SQLite：`sqlitedb.Open(ctx)` 返回 `*sql.DB`，`gormdb.Open(ctx)` 返回 `*gorm.DB`，两者共用同一个连接（CONTRACTS §6.1）。没有特别需要时用 `sqlitedb`：它比 GORM 少约 1.5 MiB 常驻内存。
- 迁移用所选包的 `Migrate` 或 `Migrations`，二者择一；辅助迁移的 ID 发布后不能改成另一项操作（CONTRACTS §6.2）。
- 受管 SQLite 在 reload 与 update 时使用独立候选副本，准备失败会回到旧数据（CONTRACTS §6.3）。框架外的文件与外部数据库不在这个保证内。
- 大数据放在磁盘上：索引、缓存、向量放进受管 SQLite 或插件数据目录中的文件，不整份载入内存；内存缓存必须有上限。
- 需要 PostgreSQL、MySQL 等外部数据库时自带驱动，连接参数做成 `secret:"true"` 的配置项由管理员填写。迁移或 OnStart 会写外部数据库时，manifest 声明 `prepare_effects=external`，此后已有安装的更新需要管理员进入维护（CONTRACTS §6.1、§6.3）。Core 不创建、不备份外部数据库。
- Core 托管的轻量 KV 尚未提供（Q-19），简单状态也用 `sqlitedb`。

## 4. 日志与文案

- 用标准库 `log/slog` 记录，SDK 把记录交给 Core 并补上来源（CONTRACTS §14）。不要往 stdout 写内容：stdio 传输下 stdout 是协议通道。
- 日志、错误字符串与 safe_message 用英文；给用户看的文字放进 `i18/` 的 `messages`，至少提供 en-US（D-30、CONTRACTS §5.4）。
- 不记录 secret、token、完整配置与 DSN；含凭证的 URL 先脱敏（CONTRACTS §14）。

## 5. 生命周期与调用

- 资源在 OnStart 中打开、在 OnStop 中关闭；后台 goroutine 随 OnStop 结束，并响应 ctx 取消（CONTRACTS §4）。
- 普通请求的期限是 30s。需要更久的工作放进延续调用或 schedule 回调（CONTRACTS §1.2、§2.4、§4.3）。
- 有副作用的业务调用不自动重试；超时与取消不代表对端没有执行（CONTRACTS §1.2）。

## 6. 跨插件协作

- 用 Service Call 与 Event Bus 协作，不直接读写其他插件的文件或数据库（CONTRACTS §3）。
- 需要模型能力时调用 Agent 组件的 `llm.generate`：`sdk/llm` 提供类型与以 `@agent` 为目标的调用辅助，官方 Agent 被替换后也不用修改；请管理员把插件加入 `llm_service_allowed_plugins`；一次调用是一次非流式请求，期限受调用方约束（CONTRACTS §13.5）。
- 要把能力交给 Agent 使用时，注册带 `tool` 描述的 JSON Service（CONTRACTS §13.1）。

## 7. 配置与 secret

- 根插件嵌入 `sdk.Base`，至多一个配置类型并嵌入 `config.Base`；校验写在 OnLoad 中（D-39、D-40、CONTRACTS §5）。
- secret 字段加 `secret:"true"`，不写进日志、错误或数据库明文（CONTRACTS §5.3）。

## 8. 测试

- 用插件测试套件 `plugintest` 在 `go test` 中启动真实 Core 与测试 Adapter，注入消息并检查发送（D-47、CONTRACTS §17）。
- 本仓库中的官方插件另遵守 [编码规范](coding-standards.md) 的测试提交规则。

## 9. 发布

- 市场 ID 按角色使用 `alonabot_plugin_`、`alonabot_adapter_`、`alonabot_agent_` 或 `alonabot_webui_` 前缀（CONTRACTS §9.1）。manifest 可以用 `icon` 声明组件图标，配置表单的选项也可以各带图标（SVG、PNG 或 WebP，CONTRACTS §7.5）。
- 用户用 `install --source owner/repo` 安装社区组件，之后只从该来源更新；Adapter、Agent 与 WebUI 同类不能并存，替换官方组件需要 `--replace`（CONTRACTS §9.3）。
- 用 `alonabot plugin check` 与 `alonabot plugin pack` 为每个支持的平台生成安装包与外部 `release.json`，用 Ed25519 签名（CONTRACTS §9.2、§9.3）。签名私钥由作者保管，不放进仓库或构建日志。
