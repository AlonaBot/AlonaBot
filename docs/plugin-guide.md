# 插件开发规范

负责范围：插件作者（含本仓库的官方插件、Adapter、Agent 与示例）在依赖、存储、日志、生命周期与发布上必须遵守的规则，以及这样要求的原因。Adapter 与 Agent 是有进程的非插件组件（DESIGN §T），这里的规则同样适用于它们；WebUI 组件只含静态文件，只需遵守第 9 节与 CONTRACTS §7.5。行为的权威是 [CONTRACTS](../spec/CONTRACTS-v0.1.md)，本文只列规则并指向对应条款，两者冲突时以契约为准。
更新时机：契约中影响插件作者的条款变化、SDK 新增按需包或依赖政策变化时。

SDK 仍在实施中，文中的包、配置项与命令以契约为准，尚未全部实现；实际进度见 [spec/README.md](../spec/README.md)。

## 1. 运行模型与内存

- 每个插件是一个独立的 Native Go 进程（D-01），只经 SDK 与 Core 通信。崩溃只影响它自己，但插件仍是可信的本地代码，不是沙箱（CONTRACTS §7.1）。
- 没有启用的插件不运行进程，启用后它的常驻内存全部算在用户机器上。实测一个只用 SDK 的插件进程约 7.5 MiB（linux/amd64 PSS），每多链接一个库、多做一次初始化都会增加这个数。官方组件的预算见 TEST-MATRIX V-74（D-52）。

## 2. 依赖

- 根包 `sdk` 提供插件对象、Handler Context、注册 builder、权限、事件类型与运行入口。分类 API 按需导入 `sdk/task`、`sdk/service`、`sdk/bus`、`sdk/button`、`sdk/media`、`sdk/conversation`、`sdk/tool`、`sdk/action` 和 `sdk/options`。
- 只导入用到的包。SDK 根包只允许依赖 protobuf、x/sys 与 jsonschema-go（D-52）；数据库在 `sdk/sqlitedb` 与 `sdk/gormdb` 中，不导入就不链接（D-49）。
- 能用标准库就不引入第三方库。HTTP 用 `net/http`；不要内嵌模型厂商的 SDK 或自己的 Provider 客户端，需要模型时调用 Agent 组件的 `llm.generate`（第 6 节）。
- 包级变量与 `init` 不做重活：不加载大文件、不建大表、不批量编译正则，推迟到 OnStart 或首次使用。所有被链接的包的 `init` 都会在启动时执行，分配的内存会一直常驻。
- 优先纯 Go。确需 CGO 时，必须为每个声明支持的平台提供真实可运行的二进制，不支持的平台在 manifest 中不声明（CONTRACTS §7.1、§9.1）。
- 插件的依赖由它自己的 `go.mod` 管理，与其他插件和 Core 互不影响；代价是每个依赖都完整地进入这个插件的进程。
- 消息、事件、会话、消息句柄、发送回执与媒体使用 `sdk/message`，错误码使用 `sdk/errcode`；公共 API 的协议转换由 SDK 内部完成。自有 Protobuf 服务仍以自己的 `proto.Message` 类型配合 `service.Proto`、`service.Spec.Call` 使用（CONTRACTS §3、§8，D-70、D-71）。

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
- 取文案用 `c.T(key, name, value, ...)`，后台用 `sdk.T`，两者只返回字符串（D-106）。键与参数名尽量写成常量：`alonabot build` 与 `alonabot check` 会检查键是否存在于 en-US、参数是否成对、参数名是否为 `string` 类型且与占位符一致；键在运行时才确定时不检查，运行时找不到键或参数不符时返回键名或保留占位符原文，并在插件日志记一次 WARN（CONTRACTS §5.4）。
- 不记录 secret、token、完整配置与 DSN；含凭证的 URL 先脱敏（CONTRACTS §14）。

## 5. 生命周期与调用

- 资源在 OnStart 中打开、在 OnStop 中关闭；后台 goroutine 随 OnStop 结束，并响应 ctx 取消（CONTRACTS §4）。
- 普通请求的期限是 30s。需要更久的工作放进延续调用或 schedule 回调（CONTRACTS §1.2、§2.4、§4.3）。延续调用有两种写法，见下一条。
- 延续调用有两种写法，期限为 CONTINUATION_TIMEOUT，不占分发链：
  - Handler 内调用 `c.Continue(func(c *sdk.Context) error { ... })`：原 Handler 照常返回并决定传播，fn 在原调用成功结束后的独立调用中执行，原调用返回错误时 fn 不执行。
  - 命令、事件或按钮路由的 builder 写 `.Continue()`：SDK 在原调用中照常执行 Rule，匹配后整个 Handler 放到延续调用中执行。这类路由中 `StopPropagation` 不起作用；自动回复型 Handler 返回的非空内容在延续调用结束前发送一次。
  - 在延续调用中，`c.Wait(timeout)` 等待同一用户在同一会话（含话题）的下一条消息，返回它的 Context，可用 `Text()` 取文本；这条消息不再进入普通分发。只能由真实入站事件触发（需要 Actor 与默认 Target），延续调用中不能再次 Continue，等待与延续都不持久化，插件停止或替换时结束（D-119）。
- 有副作用的业务调用不自动重试；超时与取消不代表对端没有执行（CONTRACTS §1.2）。
- 插件包导出 `func New() *sdk.Plugin`：`p := sdk.NewPlugin()` 后在 p 上登记路由、`p.OnStart`/`OnReady`/`OnStop`、`p.OnConfigLoad` 与数据库迁移，最后返回 p；`New` 可以被多次调用，不依赖包级状态（D-114、D-145）。逻辑模块用 `p.Sub(id)` 分成一层子插件，通常由子包导出 `func Register(s *sdk.SubPlugin)`；子插件的 ID 不进入 route_id、命令路径与 Service 名，生命周期回调与迁移只登记在插件对象上（D-116）。
- 参数较多或需要可选、默认值时，命令 Handler 在 `*sdk.Context` 之后只收一个结构体：字段按顺序对应参数，名字取 `arg` 标签或把字段名开头的大写段转小写（`FeedURL` 为 `feedURL`），指针字段可选、缺失为 nil，`default:"…"` 给出默认值，`arg:",raw"` 接收末位 raw 尾部；与位置参数及 `Args`/`Optional`/`RawTail` 不能混用，全部检查在注册收集阶段完成（D-115）。
- 位置参数的名字：省略 `Args` 时取 Handler 源码中的参数名，由构建工具自动记录，显式 `Args` 优先；`_` 与无名参数、同一行多个函数字面量（生成时给出 WARN）回退为 `argN`，这时要用 `Args` 显式命名后才能按名字写 `Optional`。
- HTTP 路由写作 `p.Route(method, path).Handle(h)`（或 `.HandleFunc(f)`），公开回调在 `Handle` 之前加 `.Public(sdk.AuthCallbackState)` 等认证方式；定时任务写作 `p.Every(jobID, d).Handle(fn)` 或 `p.Cron(jobID, expr).Timezone(tz).Handle(fn)`。每个 builder 必须恰好 `Handle` 一次，`Handle` 之后不能再设置选项，错误在注册收集阶段报告。
- 命令、文本匹配、事件、Service 与 Event Bus 订阅都链式注册，以 `Handle` 结束；漏写 `Handle`、重复调用或在它之后继续调用，都会让注册收集阶段报错（D-70）。
- 命令说明在注册处写英文原文：`p.Command("set").Description("Set a value").Handle(Set)`。其他语言在 `i18n/<locale>.json` 的顶层 `commands` 中以完整规范路径（含 Group 前缀，token 用一个空格连接）覆盖，例如 `{"commands":{"rss add":{"description":"添加订阅"}}}`；只有 `commands` 的语言文件可以省略 en-US。`check`/`build` 只检查 `commands` 的结构，注册收集阶段检查覆盖路径对应已注册且登记了英文说明的命令，否则注册失败（D-138）。
- 权限写成 `sdk.Permission`：`sdk.Anyone`、`sdk.User`、`sdk.Admin`、`sdk.ChatAdmin`、`sdk.PermissionName`，用 `Or` 取并集。命令与文本匹配取最近一层显式设置（路由、最近的组、外层的组），都没有时要求 `user`；事件路由没有设置时为空集合。Handler 内用 `c.Has` 按当前 Actor 判断。
- `c.Send`、`sdk.Send`、编辑与自动回复都接收 `message.Content`，可以直接传 `message.Text(...)`、`message.ImageURL(...)` 等单个片段，也可以传 `message.Chain{...}`。自动回复型 Handler 的返回类型为 `(message.Content, error)`，成功返回的非空内容由 Core 在 Handler 返回后发送，排在本次调用的全部显式发送之后；nil 与空 Chain 表示不回复，显式发送与编辑的空内容返回 INVALID_ARGUMENT（CONTRACTS §8，MESSAGE-MODEL §2～§3，D-108）。
- 发送目标用 `string` 会话字符串：`<adapter_id>:<chat_id>[#<thread_id>]`，例如 `telegram:-100123#42` 或 `onebot11:group:123456`。用 `message.SID(adapterID, chatID, threadID)` 处理 chat_id 与 thread_id 中 `%`、`#` 的转义，用 `message.ParseSID(sid)` 校验并拆解；adapter_id 遵守 manifest 的本地 ID 规则。`c.Send(content)` 使用本次调用的默认会话，`c.SendTo(sid, content)` 发往指定会话，`sdk.Via(instanceID)` 指定机器人实例；只有 Via 时保留默认会话与话题。后台用 `sdk.Send(ctx, session, content, opts...)`，opts 只接受 Via。省略实例时由 Core 选择并补齐账号；非法会话在发出请求前返回 INVALID_ARGUMENT（CONTRACTS §8、§9.1，MESSAGE-MODEL §1，D-102）。
- 编辑、撤回与引用所指的消息统一用 `string` 消息句柄，格式为 `<会话字符串>/<platform_message_id>[@<实例名>]`，可以原样保存。用发送回执的 `Ref()`、`Refs()` 或 `c.Event().Ref()` 取得句柄，手工构造用 `message.Ref(sid, platformMessageID, instance)`，它负责消息 ID 中 `%`、`/`、`@` 的转义，用 `message.ParseRef(ref)` 校验并拆解。编辑用 `c.Edit(ref, content)` 或后台的 `sdk.Edit(ctx, ref, content)`，引用用 `message.Quote(ref)`；省略实例时由 Core 按目标选择规则补齐（CONTRACTS §8，MESSAGE-MODEL §4，D-104）。
- 流式输出用 `c.Stream(next)`：next 每次返回一段增量文本，返回 `io.EOF` 正常结束（可以同时带上末段）。实例声明 `message.edit` 时先发送再节流编辑，超过 `TextLimit` 时换条；没有编辑能力时在正常结束后发送一次，输入出错时不发布缓存。返回全部回执，失败时保留已有回执（D-135）。
- 只回一段文本时用 `c.Reply(text)`；后台发文本用 `sdk.SendText(ctx, session, text, opts...)`，opts 只接受 Via。两者分别等于对应 Send 发送 `message.Text(text)` 并丢弃回执，错误原样返回（D-118）。
- `c.User()` 返回 `string` 用户字符串 `<adapter_id>:<external_user_id>`，没有 Actor 时为空串；`message.UID` 与 `message.ParseUID` 保留 external_user_id 中的冒号。`c.Session()` 返回默认会话，没有默认目标时为空串，未指定目标的发送返回 NO_DEFAULT_TARGET。Service 调用沿用调用上下文中的 Actor 与默认会话，`c.Event()` 为 nil（CONTRACTS §1.2、§3、§8，MESSAGE-MODEL §1，D-120）。
- 事件用常量注册，例如 `p.On("created", sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error { return c.Reply(e.Chain.Text()) })`；也接受只收 `*sdk.Context` 的 Handler。其他标准事件同样使用对应常量与 payload 指针类型，注册收集阶段检查签名并在错误中给出 route_id。扩展事件用 `sdk.AdapterEvent[T]("adapter.name", version)`，T 须为结构体，带载荷的 Handler 收 `*T`；版本不符或 JSON 解码失败时记 WARN 并跳过 Handler，日志只含路由、事件名与版本或安全错误类别（MESSAGE-MODEL §1.1～§1.2，CONTRACTS §8，D-70、D-71、D-113）。
- 上传用 `media.Upload(ctx, r, mime)` 得到 `*message.MediaRef`，SDK 计算大小与 SHA-256，mime 为空时按内容判断；r 可 Seek 时读两遍，否则使用临时文件，调用方应保持两遍读取的内容一致（CONTRACTS §1.3，D-109）。读取用 `media.Open(ctx, src)`，src 可以是片段的 `Source`、`MediaRef.Source()`、`IncomingMedia.Source()`，只持有 token 时用 `message.MediaSource{MediaRef: token}`。打开时先读第一块，失败直接返回错误；成功后顺序读取并关闭返回的 `io.ReadCloser`，读取响应 ctx 取消且不自动重试。HTTPS URL 来源在发出请求前返回 INVALID_ARGUMENT。向其他插件授权用 `media.Grant(ctx, token, granteePluginID, sessions...)`，会话参数为 `string`（CONTRACTS §1.3～§1.4，D-102、D-109、D-121）。
- 错误码用 `errcode`，例如 `errcode.InvalidArgument.New("invalid input")`、`errcode.InvalidArgument.Errorf("invalid field %s", name)`；用 `errcode.Of(err)` 查询错误链中的 SDK 错误码，没有 SDK 错误时返回 0。`sdk.Error` 是 `errcode.Error` 的别名，仍可用 `errors.As` 取得错误并读取 Code 与 Message（CONTRACTS §1.2，D-123）。

发送带回调的按钮时，用 `a := button.Define[T](name)` 声明动作，通过 `button.On(r, a).Handle(...)` 注册签名为 `func(*sdk.Context, T) error` 的 Handler，再将 `a.Button(label, value)` 放入 `message.Buttons` 发送。`button.` 是按钮路由的保留前缀，注册其他路由时避开该前缀。value 用 `encoding/json` 编码，Core 限制编码结果最多 1024 字节；编码失败的按钮在发送前返回 INVALID_ARGUMENT。点击时 value 解码失败记 WARN，跳过 Handler 并正常结束本次调用。直接构造 `message.Button` 时，空的 `Value` 会省略该字段，省略值的点击同样按解码失败处理。`message.Extension(adapterID, name, version, payload)` 只能发往所属 Adapter 声明了对应 extension 名称与版本的实例，payload 按声明的 Schema 校验（[CONTRACTS §8](../spec/CONTRACTS-v0.1.md)、[MESSAGE-MODEL §6.3](../spec/MESSAGE-MODEL.md#63-按钮与回调)）。

Handler 内的平台动作使用 `c.CallAction`；后台使用 `action.NewClient(ctx)`，两者都实现 `action.Caller`。

Adapter 作者用 `adapter.Event[T](name, version)` 与 `adapter.Extension[T](name, version)` 声明扩展事件与 extension；T 须为结构体，放在 Adapter 的常量与数据类型包中，供插件共用。SDK 从 Go 类型推导 Schema：没有 `omitempty` 的字段为必需，不允许未声明的属性；类型不符或推导失败时，`RunAdapter` 在 REGISTER 前返回带声明名的错误（CONTRACTS §8，MESSAGE-MODEL §1.2、§2，D-113）。

Adapter 的 HTTP 路由在 `adapter.Descriptor.HTTP` 中声明，例如：

```go
return &adapter.Descriptor{
    AdapterID: "example",
    HTTP: []adapter.HTTPRoute{{
        Method: http.MethodPost,
        Path: "/webhook",
        Auth: adapter.AuthPlatformToken,
        Handler: http.HandlerFunc(handleWebhook),
    }},
}
```

省略 `Auth` 时要求 Core 会话；`AuthPlatformToken` 由 Adapter 校验平台凭证。SDK 在连接 Core 前核对路径与重复路由。

## 6. 跨插件协作

- 用 Service Call 与 Event Bus 协作，不直接读写其他插件的文件或数据库（CONTRACTS §3）。
- Service 与事件都先写成包级的声明，提供方与调用方共用同一个声明：`var Search = service.JSON[SearchIn, SearchOut]("weather/search")`（Protobuf 用 `service.Proto`），提供方在 `New()` 中 `service.Serve(p, Search).Handle(fn)`，调用方 `Search.Call(ctx, in)`；事件为 `var Updated = bus.Define[Update]("weather/updated")`，发布 `Updated.Publish(ctx, payload)`，订阅 `bus.Subscribe(p, Updated).Handle(fn)`，订阅的路由 ID 缺省为完整事件名，同一事件订阅多次时用 `.ID(routeID)` 区分。target 与事件名的组件部分必须是本插件的 ID，SDK 在握手取得插件 ID 后、REGISTER 之前检查。`@agent/<name>` 只应由 Agent 组件 Serve：SDK 只检查它的格式，Broker 只把 `@agent` 的调用转给 `role: agent` 的组件，普通插件写了也不会被找到。声明可以从其他 Go module 导入后直接调用，调用方的生成器会嵌入该声明的输入与输出 Schema，经包装函数调用时同样适用。
- Event Bus 是非持久广播：`Publish` 只能发布到本插件的命名空间，返回成功只表示 Core 准入，不代表有人收到；订阅方队列满、预算不足或离线时事件会丢失，也不会补发。需要回执或可靠送达时改用 Service Call。订阅回调串行执行，不要在其中阻塞：积压的事件占用本插件的接收预算，会让命令请求也被拒绝。回调里没有 Actor 与默认会话，发消息时用 `sdk.Send(ctx, session, content, ...)` 或 `sdk.SendText(ctx, session, text, ...)` 显式给出会话（CONTRACTS §3、§8，D-102、D-118）。
- `service.JSON`、`bus.Define` 与 `task.Define` 声明的类型参数必须是包级的具名非泛型类型，由生成器检查并给出带位置的诊断。JSON Service 与 Bus 事件的 schema_id 由类型决定，版本取类型的 `SchemaVersion() uint32`，没有该方法时为 1；改动类型后运行 `alonabot build`。切片与 map 接受 null，与 `encoding/json` 对 nil 的编码一致。
- JSON Service 的输入不符合 Schema 时，调用方收到 INVALID_ARGUMENT 和固定文本 `invalid service input`，Handler 不被调用；输出不符合时调用方收到 INTERNAL，日志只记 `schema_id`；调用方的 `Call` 也按声明校验响应，不符合时返回 INTERNAL 与 `invalid service output for "<target>"`。校验详情可能带业务载荷，不写入响应和日志。
- 通过 `tool.List(ctx)` 获取工具目录；用 `.Tool(description, tool.DefaultAllow)` 登记工具，需确认时改用 `tool.DefaultConfirm`。
- 需要模型能力时调用 Agent 组件的 `llm.generate`：`sdk/llm` 提供类型与以 `@agent` 为目标的调用辅助，官方 Agent 被替换后也不用修改；请管理员把插件加入 `llm_service_allowed_plugins`；一次调用是一次非流式请求，期限受调用方约束（CONTRACTS §13.5）。
- 要把能力交给 Agent 使用时，注册带 `tool` 描述的 JSON Service（CONTRACTS §13.1）。
- 在 Handler 或其延续调用中，`tool.Invoke(ctx, "plugin_id/service_name", args)` 按当前生效的工具策略调用工具（deny 返回 PERMISSION_DENIED；confirm 时 Core 先向当前会话发送带「允许」「拒绝」按钮的确认，展示实际参数，拒绝返回 PERMISSION_DENIED，超时返回 DEADLINE_EXCEEDED）；`tool.Confirm(ctx, operation, summary)` 只请求用户确认，展示插件给出的摘要，返回 nil 表示允许。两者都需要父调用链，后台任务中没有 Actor 时遇到确认返回 PERMISSION_DENIED；实例不支持按钮时返回 UNSUPPORTED（CONTRACTS §13.2）。

## 7. 配置与 secret

在 `plugin.yaml` 同目录编写 `config_schema.json`，按字段键声明类型与英文源文案。运行 `alonabot build` 检查声明并把规范 JSON 经 overlay 编入程序；`alonabot check` 执行同样的检查。声明格式见 [CONTRACTS §5.1](../spec/CONTRACTS-v0.1.md#51-配置真源缺失值与编辑)，例如：

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
- 声明式迁移后调用 `p.OnConfigLoad(fn)` 登记的配置回调 `func(ctx context.Context, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`。`raw` 是完整的迁移结果，`fromVersion` 保留迁移前的版本，首装时为目标版本；没有登记时原样返回。需要额外转换或业务取值检查时登记此回调，保留未知键，响应 ctx，只做可重复的纯配置变换；数据库、文件、网络与发信操作放在对应生命周期中。示例见 [showcase](../examples/plugins/showcase/main.go)。普通 restart 与恢复时，回调改变配置会返回 FAILED_PRECONDITION。
- 回调之后，SDK 只给缺失键补 `default` 或类型零值，递归进入对象、列表与已选择的模板，再校验。显式 `false`、`0`、空串与空集合保持，未知键和数字原文保留。声明字段为 `null`、类型不符或违反 options、slider、required 等约束时，准备失败，错误只含 JSON 路径和错误类别。

插件对象从 OnStart 开始通过 `p.Config()`（Handler 内为 `c.Config()`）取得 `config.Values`，子插件的 `s.Config()` 是 `config_schema.json` 中同名顶层 object 的只读视图；在 OnConfigLoad 中或更早调用会 panic。例如 `p.Config().String("greeting")` 读取上例的问候语。取值对象在同一 Session 内保持不变，可并发读取；集合 getter 返回可独立修改的副本。

字段写 `"profile": true` 时可以按配置文件覆盖（CONTRACTS §5.6，D-168）：Handler 中的 `c.Config()` 合并当前会话所用配置文件的覆盖值，`p.Config()` 与 `s.Config()` 仍是插件配置本身。在 OnStart 或后台读取的字段（令牌、连接、缓存）不要标记，secret、list 的元素与模板中的字段不能标记。

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

- 字段的可选值来自运行时的数据（例如实例列表或本插件维护的条目）时，用 `options_from` 代替 `options`：`{"type": "string", "options_from": {"source": "core.adapter_instances", "filter": {"adapter_id": "telegram"}}}`，或指向本插件的选项服务 `"source": "<插件 ID>/<服务名>"`（[CONTRACTS §5.5](../spec/CONTRACTS-v0.1.md)「选项来源」，D-129）。选项服务用 `options.Serve(p, "feeds", func(ctx context.Context, q options.Query) (options.Page, error) {...})` 登记：按 `q.Search` 搜索、`q.Values` 按存储值取回显示名，最多返回 `q.Limit` 项，有更多结果时 `More` 为 true；输入、输出 Schema 由 SDK 固定提供，不用生成。该服务只由 Core 在管理页面打开表单时调用，访问固定为 PRIVATE。取值仍是普通字符串，来源不可用或值已不存在时配置照常加载，插件自己处理不认识的值。

## 8. 测试

- 用插件测试套件 `plugintest` 在 `go test` 中启动真实 Core 与测试 Adapter，注入消息并检查发送（D-47、CONTRACTS §17）。
- 本仓库中的官方插件另遵守 [编码规范](coding-standards.md) 的测试提交规则。

## 9. 发布

- 市场 ID 按角色使用 `alonabot_plugin_`、`alonabot_adapter_`、`alonabot_agent_` 或 `alonabot_webui_` 前缀（CONTRACTS §9.1）。manifest 可以用 `icon` 声明组件图标（SVG、PNG 或 WebP，CONTRACTS §7.5）。
- 用户用 `install --source owner/repo` 安装社区组件，之后只从该来源更新；Adapter、Agent 与 WebUI 同类不能并存，替换官方组件需要 `--replace`（CONTRACTS §9.3）。
- 用 `alonabot build` 检查资源并编译独立组件，`alonabot check` 只读核对。`alonabot pack --asset-url https://example.com/component.tar.gz` 为所选平台生成安装包与归档外的 `release.json`；目标通过 `--goos/--goarch/--tags` 选择，`--out` 默认 `<工程>/bin`。可选 `--key FILE` 读取标准 base64 编码的 64 字节 Ed25519 私钥，对清单原始字节签名，并在归档外写 `release.json.sig` 与 `release.pub`；省略时输出 unsigned 清单（CONTRACTS §9.2、§9.3）。密钥读取或签名失败直接报错。进程包含可解析的 `build-info` 文本，WebUI 包只含 manifest 与静态资源。签名私钥由作者保管，留在打包资源、仓库和构建日志之外。

## 直接测试与构建

在组件目录运行 `go test ./...` 即可使用 plugintest。测试启动时按入口包现场扫描 manifest、Schema、i18n 和 Handler 参数名，并复用同一入口包的只读资源。作者的 `New()` 只创建插件并登记业务行为，工程里没有 `alonabot_gen.go`，也无需编写生成相关的代码或指令。

独立组件运行 `alonabot build --dir .`；通过根包 `alonabot.Run` 或 `alonabot.Start` 装配组件的 main 包运行 `alonabot build --dir ./cmd/bot`，执行产物为 `cmd/bot/bin/bot`（Windows 加 `.exe`）。库模式构建会只读扫描依赖中的组件，包括模块缓存中的第三方组件，并将资源经 main 包的 overlay 编入。直接 Go 构建的程序在启动缺少资源的组件时会提示使用 `alonabot build` 或 `alonabot run`。库模式的公开选项与停止用法见[开发指南](development.md#库模式运行同一进程)；独立进程开发运行与热部署使用 `alonabot run --reload`。
