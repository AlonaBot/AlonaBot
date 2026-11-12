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
- 需要 PostgreSQL、MySQL 等外部数据库时自带驱动，连接参数做成在 `config_schema.json` 中声明 `"secret": true` 的配置项（§7），由管理员填写。迁移或 OnStart 会写外部数据库时，manifest 声明 `prepare_effects=external`，此后已有安装的更新需要管理员进入维护（CONTRACTS §6.1、§6.3）。Core 不创建、不备份外部数据库。
- Core 托管的轻量 KV 尚未提供（Q-19），简单状态也用 `sqlitedb`。

## 4. 日志与文案

- 用标准库 `log/slog` 记录，SDK 把记录交给 Core 并补上来源（CONTRACTS §14）。不要往 stdout 写内容：stdio 传输下 stdout 是协议通道。
- 日志、错误字符串与 safe_message 用英文；给用户看的文字放进 `i18/` 的 `messages`，至少提供 en-US（D-30、CONTRACTS §5.4）。
- 取文案用 `c.T(key, name, value, ...)`，后台用 `sdk.T`，两者只返回字符串（D-106）。键与参数名尽量写成常量：`alonabot plugin generate` 与 `plugin check` 会检查键是否存在于 en-US、参数是否成对、参数名是否为 `string` 类型且与占位符一致；键在运行时才确定时不检查，运行时找不到键或参数不符时返回键名或保留占位符原文，并在插件日志记一次 WARN（CONTRACTS §5.4）。
- 不记录 secret、token、完整配置与 DSN；含凭证的 URL 先脱敏（CONTRACTS §14）。

## 5. 生命周期与调用

- 资源在 OnStart 中打开、在 OnStop 中关闭；后台 goroutine 随 OnStop 结束，并响应 ctx 取消（CONTRACTS §4）。
- 普通请求的期限是 30s。需要更久的工作放进延续调用或 schedule 回调（CONTRACTS §1.2、§2.4、§4.3）。
- 有副作用的业务调用不自动重试；超时与取消不代表对端没有执行（CONTRACTS §1.2）。
- 命令、文本匹配、事件、Service 与 Event Bus 订阅都链式注册，以 `Handle` 结束；漏写 `Handle`、重复调用或在它之后继续调用，都会让 `Collect` 报错（D-70）。
- 权限写成 `sdk.Permission`：`sdk.Anyone`、`sdk.User`、`sdk.Admin`、`sdk.ChatAdmin`、`sdk.PermissionName`，用 `Or` 取并集。命令与文本匹配取最近一层显式设置（路由、最近的组、外层的组），都没有时要求 `user`；事件路由没有设置时为空集合。Handler 内用 `c.Has` 按当前 Actor 判断。
- 两种 Handler 都可以 `c.Send`。自动回复型 Handler 返回的非空 Chain 由 Core 在 Handler 返回后发送，排在本次调用的全部显式发送之后。

发送带回调的按钮时，用 `a := sdk.Action[T](name)` 声明动作，通过 `sdk.OnButton(r, a).Handle(...)` 注册签名为 `func(*sdk.Context, T) error` 的 Handler，再将 `a.Button(label, value)` 放入 `message.Buttons` 发送。`button.` 是按钮路由的保留前缀，注册其他路由时避开该前缀。value 用 `encoding/json` 编码，Core 限制编码结果最多 1024 字节；编码失败的按钮在发送前返回 INVALID_ARGUMENT。点击时 value 解码失败记 WARN，跳过 Handler 并正常结束本次调用。直接构造 `message.Button` 时，空的 `Value` 会省略该字段，省略值的点击同样按解码失败处理。`message.Extension(adapterID, name, version, payload)` 只能发往所属 Adapter 声明了对应 extension 名称与版本的实例，payload 按声明的 Schema 校验（[CONTRACTS §8](../spec/CONTRACTS-v0.1.md)、[MESSAGE-MODEL §6.3](../spec/MESSAGE-MODEL.md#63-按钮与回调)）。

## 6. 跨插件协作

- 用 Service Call 与 Event Bus 协作，不直接读写其他插件的文件或数据库（CONTRACTS §3）。
- Event Bus 是非持久广播：`sdk.Publish` 只能发布到本插件的命名空间，返回成功只表示 Core 准入，不代表有人收到；订阅方队列满、预算不足或离线时事件会丢失，也不会补发。需要回执或可靠送达时改用 Service Call。订阅回调串行执行，不要在其中阻塞：积压的事件占用本插件的接收预算，会让命令请求也被拒绝。回调里没有 Actor 与默认 Target，发消息时要显式指定 To。
- `sdk.JSON`、`sdk.Subscribe`、`sdk.CallJSON`、`sdk.Publish` 的类型参数必须是包级的具名非泛型类型，由生成器检查并给出带位置的诊断。schema_id 由类型决定，版本取类型的 `SchemaVersion() uint32`，没有该方法时为 1；改动类型后运行 `alonabot plugin generate`。切片与 map 接受 null，与 `encoding/json` 对 nil 的编码一致。
- JSON Service 的输入不符合 Schema 时，调用方收到 INVALID_ARGUMENT 和固定文本 `invalid service input`，Handler 不被调用；输出不符合时调用方收到 INTERNAL，日志只记 `schema_id`。校验详情可能带业务载荷，不写入响应和日志。
- 需要模型能力时调用 Agent 组件的 `llm.generate`：`sdk/llm` 提供类型与以 `@agent` 为目标的调用辅助，官方 Agent 被替换后也不用修改；请管理员把插件加入 `llm_service_allowed_plugins`；一次调用是一次非流式请求，期限受调用方约束（CONTRACTS §13.5）。
- 要把能力交给 Agent 使用时，注册带 `tool` 描述的 JSON Service（CONTRACTS §13.1）。

## 7. 配置与 secret

在 `plugin.yaml` 同目录编写 `config_schema.json`，按字段键声明类型与英文源文案。运行 `alonabot plugin generate` 检查声明并把规范 JSON 嵌入生成文件；`alonabot plugin check` 执行同样的检查。声明格式见 [CONTRACTS §5.1](../spec/CONTRACTS-v0.1.md#51-配置真源缺失值与编辑)，例如：

```json
{
  "$version": 2,
  "$migrations": [{"to": 2, "rename": {"hello": "greeting"}}],
  "greeting": {"type": "string", "description": "Greeting", "default": "Hello"},
  "enabled": {"type": "bool", "default": true},
  "token": {"type": "string", "description": "Access token", "secret": true}
}
```

- `$version` 是配置格式版本，缺省为 1。SDK 从已保存的版本依次执行 `$migrations`，首装跳过迁移；上例把版本 1 的 `hello` 移到 `greeting`。每步按 `rename` → `map` → `wrap` → `remove` 执行，路径不存在时跳过，rename 目标已存在时失败；迁移路径与版本检查见 [CONTRACTS §5.5](../spec/CONTRACTS-v0.1.md#55-配置版本与迁移)。
- 声明式迁移后调用根插件的 `OnConfigLoad(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`。`raw` 是完整的迁移结果，`fromVersion` 保留迁移前的版本，首装时为目标版本；`sdk.Base` 默认原样返回。需要额外转换或业务取值检查时覆盖此方法，保留未知键，响应 ctx，只做可重复的纯配置变换；数据库、文件、网络与发信操作放在对应生命周期中。示例见 [showcase](../examples/plugins/showcase/main.go)。普通 restart 与恢复时，回调改变配置会返回 FAILED_PRECONDITION。
- 回调之后，SDK 只给缺失键补 `default` 或类型零值，递归进入对象、列表与已选择的模板，再校验。显式 `false`、`0`、空串与空集合保持，未知键和数字原文保留。声明字段为 `null`、类型不符或违反 options、slider、required 等约束时，准备失败，错误只含 JSON 路径和错误类别。

根插件以值嵌入 `sdk.Base`，从 OnStart 开始通过 `p.Config()` 取得 `config.Values`；在 OnConfigLoad 中或更早调用会 panic。例如 `p.Config().String("greeting")` 读取上例的问候语。取值对象在同一 Session 内保持不变，可并发读取；集合 getter 返回可独立修改的副本。

| 声明类型 | 取值方法 |
| --- | --- |
| `string`、`text` | `String(key)` |
| `int`、`float`、`bool` | 分别为 `Int(key)`（int64）、`Float(key)`（float64）、`Bool(key)` |
| 没有 `items` 的 `list` | `Strings(key)` |
| `object`、有 `items` 的 `list` | `Object(key)`、`Objects(key)`，再按子字段键读取 |
| `dict` | `Dict(key)`，返回 `map[string]json.RawMessage` |
| `template`、`template_list` | `Template(key)`、`TemplateList(key)`，通过 `Key()` 取选中的模板键，嵌入的 `Values` 读取模板字段 |

可选 `template` 缺失时表示未选择，`Key()` 返回空串；选中时值须恰好包含一个已声明的模板键。读取未声明键或使用与声明类型不符的方法会 panic，消息指明字段路径。

secret 只用于 `string` 与 `text`，声明 `"secret": true`，同时省略 `default` 与 `options`。Core 按 schema 抽取对象、列表与模板中的秘密值：非空值生成不可变引用保存到配置快照，Bootstrap 解析为明文，`Config()` 取到明文；空串表示未设置。插件对秘密值保密，日志、错误和业务数据库中均禁止明文存放（[CONTRACTS §5.3](../spec/CONTRACTS-v0.1.md#53-secret-与翻译)）。

配置翻译写在 `i18n/<locale>.json` 的 `config` 根节点，按 schema 的字段层级覆盖文案，例如：

```json
{"config": {"greeting": {"description": "问候语"}, "token": {"description": "访问令牌"}}}
```

字段可覆盖 `description`、`hint`、`labels`；子字段放在 `items`，模板放在 `templates`，模板文案可覆盖 `name`、`hint`。options 的存储值保持原文，翻译写在 labels 中。生成器核对覆盖键与 schema 的字段、模板和选项是否对应，问题诊断带文件名与路径。

## 8. 测试

- 用插件测试套件 `plugintest` 在 `go test` 中启动真实 Core 与测试 Adapter，注入消息并检查发送（D-47、CONTRACTS §17）。
- 本仓库中的官方插件另遵守 [编码规范](coding-standards.md) 的测试提交规则。

## 9. 发布

- 市场 ID 按角色使用 `alonabot_plugin_`、`alonabot_adapter_`、`alonabot_agent_` 或 `alonabot_webui_` 前缀（CONTRACTS §9.1）。manifest 可以用 `icon` 声明组件图标（SVG、PNG 或 WebP，CONTRACTS §7.5）。
- 用户用 `install --source owner/repo` 安装社区组件，之后只从该来源更新；Adapter、Agent 与 WebUI 同类不能并存，替换官方组件需要 `--replace`（CONTRACTS §9.3）。
- 用 `alonabot plugin check` 与 `alonabot plugin pack` 为每个支持的平台生成安装包与外部 `release.json`，用 Ed25519 签名（CONTRACTS §9.2、§9.3）。签名私钥由作者保管，不放进仓库或构建日志。
