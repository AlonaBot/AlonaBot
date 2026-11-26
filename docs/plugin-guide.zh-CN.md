# 插件开发规范

[English](plugin-guide.md)

本文面向插件、Adapter、Agent 与 WebUI 作者，说明依赖、存储、日志、生命周期与发布规则。官方组件位于各自独立仓库；WebUI 可提供静态文件或附带 Go 后台，进程部分遵循相应 SDK 规则。协议字段以 [Protobuf 定义](../contracts/proto/alonabot/v1) 为准，管理调用见[部署指南](deployment.zh-CN.md#管理-api)。
更新时机：SDK、组件行为或依赖政策变化时。

## 从第一个命令开始

在 AlonaBot 源码仓库中创建插件，可以直接复用仓库的 SDK module。准备 Go 后，在仓库根目录执行：

```fish
go build -o bin/alonabot ./cmd/alonabot
bin/alonabot init ./alonabot-plugin-hello
bin/alonabot check ./alonabot-plugin-hello
bin/alonabot run ./alonabot-plugin-hello --data-dir ./data
```

在同一个终端输入 `/hello`，预期回复 `Hello AlonaBot`。修改源码后自动重载；Ctrl-C 停止。生成的 `New() *sdk.Plugin` 是组件入口，`plugin.yaml` 声明身份与来源，`plugin_config_schema.json` 提供配置 Schema。plugin/agent 模板显式标记 `.Menu()`，直接用 `c.Reply("Hello AlonaBot")` 回复，按需自行添加 `i18n/<locale>.json`。各角色生成的 Go 文件和 `plugin.yaml` 均无注释；发布前将 `source` 改为组件仓库。

```fish
bin/alonabot test ./alonabot-plugin-hello
bin/alonabot pack ./alonabot-plugin-hello --target darwin/arm64 --out ./alonabot-plugin-hello/bin
```

`test` 执行离线准备检查；消息与生命周期的集成验证使用[插件测试套件](development.zh-CN.md#插件测试)。打包 target 换成实际支持的平台，归档与分发见[发布](#9-发布)。

独立仓库中先建立自己的 Go module。正式 CLI 会写入对应 SDK 版本并整理依赖；当前源码构建的开发版需按提示选定可访问的 SDK 版本。详细说明见[开发运行](development.zh-CN.md#开发运行与热部署)。现成的[示例工程](../examples/plugins)可直接运行。

## 1. 运行模型与内存

- 独立进程部署时，每个插件是一个通过 SDK 与 Core 通信的 Native Go 进程，崩溃只影响该组件。插件仍是需要审查来源的可信本地代码。库模式组件与 Core 共用进程，详见[库模式](development.zh-CN.md#库模式运行同一进程)。
- 没有启用的插件不运行进程，启用后它的常驻内存全部算在用户机器上。2026-11-20 的固定 linux/amd64 实测中，echo 插件 PSS 三轮中位数约 10.094 MiB；链接库和初始化都会增加开销。官方组件的现有实测与维护者接受决定见 [发布说明](../changelog/v0.1.0/zh-CN.md#资源与依赖)。
- plugin/agent 缺省 `sleep_capable: true`；构建时省略的字段保持省略，安装使用同一默认值。需要保留进程内状态或独立后台活动的组件应显式声明 `sleep_capable: false`。可按重启语义恢复的组件继续遵循路由、活动引用与休眠策略准入。支持 SUSPENDED 后的命令/事件、Service 与 HTTP 按需唤醒、自动空闲退出、即时休眠策略，以及静态 interval/cron 与动态 at/interval/cron 自然到期唤醒。原计划、静态起点、暂停状态、动态实例与已接纳项跨同代次恢复保持；静态期限从原到期时刻计算，动态从取得并发名额开始计算，冷启动占用原 JOB_TIMEOUT。Handler 执行期间持有活动引用，异常离线期间的触发沿用跳过规则，过期 at 删除并记录 WARN。开发组件、Adapter 与 WebUI 保持常驻。

## 2. 依赖

- 根包 `sdk` 提供插件对象、Handler Context、注册 builder、权限、事件类型与运行入口。分类 API 按需导入 `sdk/task`、`sdk/service`、`sdk/bus`、`sdk/button`、`sdk/media`、`sdk/conversation`、`sdk/tool`、`sdk/action`、`sdk/options` 和 `sdk/command`。
- 只导入用到的包。SDK 根包链接 protobuf、x/sys、jsonschema-go 及消息模型使用的 Goldmark；数据库在 `sdk/sqlitedb` 与 `sdk/gormdb` 中，不导入就不链接。
- 能用标准库就不引入第三方库。HTTP 用 `net/http`；不要内嵌模型厂商的 SDK 或自己的 Provider 客户端，需要模型时调用 Agent 组件的 `llm.generate`（第 6 节）。
- 包级变量与 `init` 不做重活：不加载大文件、不建大表、不批量编译正则，推迟到 OnStart 或首次使用。所有被链接的包的 `init` 都会在启动时执行，分配的内存会一直常驻。
- 优先纯 Go。确需 CGO 时，必须为每个声明支持的平台提供真实可运行的二进制，不支持的平台在 manifest 中不声明。
- 插件的依赖由它自己的 `go.mod` 管理，与其他插件和 Core 互不影响；代价是每个依赖都完整地进入这个插件的进程。
- 消息、事件、会话、消息句柄、发送回执与媒体使用 `sdk/message`，错误码使用 `sdk/errcode`；公共 API 的协议转换由 SDK 内部完成。自有 Protobuf 服务仍以自己的 `proto.Message` 类型配合 `service.Proto`、`service.Spec.Call` 使用。

## 3. 存储

- 业务数据放受管 SQLite：`sqlitedb.Open(ctx)` 返回 `*sql.DB`，`gormdb.Open(ctx)` 返回 `*gorm.DB`，两者共用同一个连接。只需 SQL 时优先 `sqlitedb`，可以减少 GORM 的链接与初始化开销；现有实测条件见[发布说明](../changelog/v0.1.0/zh-CN.md#资源与依赖)。
- 迁移用所选包的 `Migrate` 或 `Migrations`，二者择一；辅助迁移的 ID 发布后不能改成另一项操作。
- 同代次 `RESUME` 复用原 generation、配置与业务 DB，SELF/VERSIONED 数据库迁移回调跳过，原迁移模式与账本保持；首次激活、更新与回退仍按迁移规则执行。配置继续执行 `OnConfigLoad`、补值/校正与 required 校验，完整结果须与原快照一致。
- 受管 SQLite 在 reload 与 update 时使用独立候选副本，准备失败会回到旧数据。框架外的文件与外部数据库不在这个保证内。
- 大数据放在磁盘上：索引、缓存、向量放进受管 SQLite 或 `sdk.PluginDataDir(ctx)` 中的文件，不整份载入内存；内存缓存必须有上限。
- 需要 PostgreSQL、MySQL 等外部数据库时自带驱动，连接参数做成在 `plugin_config_schema.json` 中声明 `"secret": true` 的配置项（§7），由管理员填写。迁移或 OnStart 会写外部数据库时，manifest 声明 `prepare_effects=external`，此后已有安装的更新需要管理员进入维护。Core 不创建、不备份外部数据库。
- 简单状态用 `sdk/kv`：数据放在受管 SQLite 的保留表 `_alonabot_kv` 中，与业务表一起随代次复制、回退和清理。链接 KV 会链接 SQLite，构建工具据此推导 `prepare_effects=managed_db`，保留表的初始化独立于插件迁移 ID。
- 插件自己的文件写到 `sdk.PluginDataDir(ctx)` 返回的数据目录；它跨 restart 保留，不随代次回退。只卸载与 disable 保留文件，`uninstall --purge-data` 和 `--purge` 在确认对象中列出目录并删除。

### 保存少量状态

在 Handler 或生命周期回调中传入所属 Session 的 `ctx`，例如保存订阅主题：

```go
import (
    "context"

    "github.com/AlonaBot/AlonaBot/sdk/kv"
)

func saveTopic(ctx context.Context, topic string) error {
    return kv.Set(ctx, "subscription/topic", []byte(topic))
}
```

读取用 `value, ok, err := kv.Get(ctx, key)`；`err == nil` 且 `ok == false` 表示键不存在，`kv.Delete(ctx, key)` 删除不存在的键也成功。键须为 1～512 字节的 UTF-8，拒绝恰为 `.` 或 `..` 的键；其他含点或斜杠的键按字面值保存。值最多 1 MiB，空值可以保存；非法参数返回 `INVALID_ARGUMENT`。没有受管 SQLite 的 Session 调用 KV 返回 `FAILED_PRECONDITION`。

`page, err := kv.List(ctx, "subscription/", "", 20)` 按键的字节序返回最多 20 个键；后续请求把 `page.Next` 作为 `cursor`，空 `Next` 表示已到末尾，非空 `cursor` 遵守键的约束。前缀按字面字节匹配，`%` 和 `_` 保持原义；空前缀列出全部键，`limit` 须为正数。

官方 WebUI 的插件「数据」页可浏览和编辑数据库表格、KV 与插件文件，并导出或清空数据；数据库保留代次只读。Agent 会话提供摘要与清除入口，接口与数据清理语义见 [管理 API](deployment.zh-CN.md#管理-api) 与 [管理 API](deployment.zh-CN.md#管理-api)。

`examples/plugins/configdb` 的 `/note kvset <key> <value>` 与 `/note kvget <key>` 演示保存和读取，沿用该示例的本地化回复。

### 保存插件文件

Core 在启动前创建目录，经 Bootstrap 下发。传入 Handler 或生命周期回调的 `ctx`（或其派生 context），SDK 优先用它确定所属插件 Session；同进程的多个插件各自取得自己的目录，独立进程在 `ctx` 未绑定时使用默认 Session。在 Session 就绪的 Handler 中通过 `sdk.PluginDataDir(ctx)` 取得 `<data_dir>/plugin_data/<id>/` 的绝对路径：

```go
import (
    "context"
    "os"
    "path/filepath"

    "github.com/AlonaBot/AlonaBot/sdk"
)

func saveCache(ctx context.Context, content []byte) error {
    return os.WriteFile(filepath.Join(sdk.PluginDataDir(ctx), "cache.json"), content, 0o600)
}
```

`OnStart`、`OnReady` 均早于 Session 就绪；此时或 Session 停止后调用 `PluginDataDir(ctx)` 会 panic：`plugin session is not ready`。目录在 Unix 上权限为 `0700`，Windows 继承状态根目录的当前账号专用 ACL。Core 负责建立与卸载清理，文件由插件管理。准备失败与回退保留已写文件，因此需要随代次回退的数据放在 KV 或受管 SQLite 中。

`examples/plugins/configdb` 的 `/note fileset <value>` 与 `/note fileget` 演示在该目录读写固定文件 `note.txt`。

## 4. 日志与文案

诊断包可收集组件日志及最近错误；日志和错误仍须避免记录凭证明文。导出只按规则遮盖列明的凭证格式和地址，聊天内容不作专门处理（[插件指南](plugin-guide.zh-CN.md)）。

- 用标准库 `log/slog` 记录，SDK 把记录交给 Core 并补上来源。不要往 stdout 写内容：stdio 传输下 stdout 是协议通道。
- 组件自己的数据损坏或无法继续运行、需要立即处理时，用 `slog.Log(ctx, sdk.LevelCritical, "component data is corrupt", "error", err)` 记录 CRITICAL；Handler panic 仍由 SDK 记录为 ERROR。
- 日志、错误字符串与 safe_message 用英文；给用户看的文字放进 `i18n/` 的 `messages`，至少提供 en-US。
- `messages` 键与命名占位符使用点分隔的标识符，每段由 ASCII 大小写字母、数字与 `_` 组成，例如 `storeEmpty` 与 `account.Name`。
- 取文案用 `c.T(key, name, value, ...)`，后台用 `sdk.T`，两者只返回字符串。键与参数名尽量写成常量：`alonabot build` 与 `alonabot check` 会检查键是否存在于 en-US、参数是否成对、参数名是否为 `string` 类型且与占位符一致；键在运行时才确定时不检查，运行时找不到键或参数不符时返回键名或保留占位符原文，并在插件日志记一次 WARN。
- 不记录 secret、token、完整配置与 DSN；含凭证的 URL 先脱敏。

## 5. 生命周期与调用

- 资源在 OnStart 中打开、在 OnStop 中关闭；后台 goroutine 随 OnStop 结束，并响应 ctx 取消。
- 普通请求的期限是 30s。需要更久的工作放进延续调用或 schedule 回调。
- 冷启动占用请求原有期限，等待取消只撤回该次请求；就绪后复核路由与权限，再交付一次。Core 按 Session 持有实际调用与资源租约的活动引用，本地取消先返回的业务持续占用至实际结束。HTTP 引用在反向代理返回时释放；自动退出进入 QUIESCING 后冻结新派发，新请求按原期限与预算等待旧树退出及下一次 Ready。排空、STOPPING、OnStop 与整树截止复用正常停止；HTTP/Bus 仍在执行的 SDK 回调沿用正常停止截止。
- 延续调用有两种写法，期限为 CONTINUATION_TIMEOUT，不占分发链：
  - Handler 内调用 `c.Continue(func(c *sdk.Context) error { ... })`：原 Handler 照常返回并决定传播，fn 在原调用成功结束后的独立调用中执行，原调用返回错误时 fn 不执行。
  - 命令、事件或按钮路由的 builder 写 `.Continue()`：SDK 在原调用中照常执行 Rule，匹配后整个 Handler 放到延续调用中执行。这类路由中 `StopPropagation` 不起作用；自动回复型 Handler 返回的非空内容在延续调用结束前发送一次。
  - 在延续调用中，`c.Wait(timeout)` 等待同一用户在同一会话（含话题）的下一条消息，返回它的 Context，可用 `Text()` 取文本；这条消息不再进入普通分发。只能由真实入站事件触发（需要 Actor 与默认 Target），延续调用中不能再次 Continue，等待与延续都不持久化，插件停止或替换时结束。
- 有副作用的业务调用不自动重试；超时与取消不代表对端没有执行。
- 插件包导出 `func New() *sdk.Plugin`：`p := sdk.NewPlugin()` 后在 p 上登记路由、`p.OnStart`/`OnReady`/`OnStop`、`p.OnConfigLoad` 与数据库迁移，最后返回 p；`New` 可以被多次调用，不依赖包级状态。逻辑模块用 `p.Sub(id)` 分成一层子插件，通常由子包导出 `func Register(s *sdk.SubPlugin)`；子插件的 ID 不进入 route_id、命令路径与 Service 名，生命周期回调与迁移只登记在插件对象上。
- 参数较多或需要可选、默认值时，命令 Handler 在 `*sdk.Context` 之后只收一个结构体：字段按顺序对应参数，名字取 `arg` 标签或把字段名开头的大写段转小写（`FeedURL` 为 `feedURL`），指针字段可选、缺失为 nil，`default:"…"` 给出默认值，`arg:",raw"` 接收末位 raw 尾部；与位置参数及 `Args`/`Optional`/`RawTail` 不能混用，全部检查在注册收集阶段完成。
- 位置参数的名字：省略 `Args` 时取 Handler 源码中的参数名，由构建工具自动记录，显式 `Args` 优先；`_` 与无名参数、同一行多个函数字面量（生成时给出 WARN）回退为 `argN`，这时要用 `Args` 显式命名后才能按名字写 `Optional`。
- HTTP 路由写作 `p.Route(method, path).Handle(h)`（或 `.HandleFunc(f)`），公开回调在 `Handle` 之前加 `.Public(sdk.AuthCallbackState)` 等认证方式；定时任务写作 `p.Every(jobID, d).Handle(fn)` 或 `p.Cron(jobID, expr).Timezone(tz).Handle(fn)`。每个 builder 必须恰好 `Handle` 一次，`Handle` 之后不能再设置选项，错误在注册收集阶段报告。
- 文本匹配直接写 `p.Regex(expr)`、`p.Keyword(patterns...)`、`p.StartsWith(patterns...)`、`p.EndsWith(patterns...)`、`p.FullMatch(patterns...)`，原始消息用 `p.Raw()`；子插件与命令组使用同样的写法。缺省 route ID 为小写类别名加冒号和以 `|` 连接的模式，Regex 保留表达式原文，Raw 为 `raw`。事件 ID 为事件名，Hook 用 `pre_event:<事件名>`、`before_send`、`after_send`。同插件内重复会报错，用 `.ID("stable-name")` 区分；模式会变化且需保留命令策略时也固定 ID。路径命令继续按命令路径生成 ID。
- Hook 直接写 `p.PreEvent(sdk.MessageCreated)`、`p.BeforeSend()`、`p.AfterSend()`，子插件同样适用。需要固定身份时在 `Handle` 前调用 `.ID(routeID)`。AfterSend 由 Core 异步管理，Send/Edit 在平台完成后返回完整回执与平台错误。平台操作前在原请求期限内预留观察容量；满载时等待，等待失败结束本次请求。每批观察按既有 Hook 顺序执行，从平台完成起独享 `REQUEST_TIMEOUT`，保留固定数据、Session、Trace、活动引用和出站禁令；Hook 失败记录 WARN。`AFTER_SEND_CONCURRENCY=4` 限制执行中的批次，`AFTER_SEND_CAPACITY=64` 包含已预留、等待和执行中的批次，均要求 ≥1。Core 停止时在现有排空预算内等待，随后取消观察并释放引用。
- 事件直接写 `p.On(eventType)`；Plugin、SubPlugin 与 CommandGroup 均提供此接口。命令组的事件沿用所属子插件归属，route ID 为事件名，需要稳定身份时在 `Handle` 前调用 `.ID(routeID)`。
- Handler 的 `c` 实现 `context.Context`，可直接传给 `kv.Set(c, ...)`、`task`、`service`、`llm`、`media`、`tool`、`command` 与 `conversation`。期限、取消与父调用关联沿用本次调用；`c.Context()` 仍可取得底层 context，Handler 返回后的失效规则保持。
- 命令、文本匹配、事件、Service 与 Event Bus 订阅都链式注册，以 `Handle` 结束；漏写 `Handle`、重复调用或在它之后继续调用，都会让注册收集阶段报错。
- 命令说明在注册处写英文原文：`p.Command("set").Description("Set a value").Handle(Set)`。其他语言在 `i18n/<locale>.json` 的顶层 `commands` 中以完整规范路径（含 Group 前缀，token 用一个空格连接）覆盖，例如 `{"commands":{"rss add":{"description":"添加订阅"}}}`；只有 `commands` 的语言文件可以省略 en-US。`check`/`build` 只检查 `commands` 的结构，注册收集阶段检查覆盖路径对应已注册且登记了英文说明的命令，否则注册失败。
- 权限写成 `sdk.Permission`：`sdk.Anyone`、`sdk.User`、`sdk.Admin`、`sdk.ChatAdmin`、`sdk.PermissionName`，用 `Or` 取并集。命令与文本匹配取最近一层显式设置（路由、最近的组、外层的组），都没有时要求 `user`；事件路由没有设置时为空集合。Handler 内用 `c.Has` 按当前 Actor 判断。
- `c.Send`、`sdk.Send`、编辑与自动回复都接收 `message.Content`，可以直接传 `message.Text(...)`、`message.ImageURL(...)` 等单个片段，也可以传 `message.Chain{...}`。自动回复型 Handler 的返回类型为 `(message.Content, error)`，成功返回的非空内容由 Core 在 Handler 返回后发送，排在本次调用的全部显式发送之后；nil 与空 Chain 表示不回复，显式发送与编辑的空内容返回 INVALID_ARGUMENT。
- 发送目标用 `string` 会话字符串：`<adapter_id>:<chat_id>[#<thread_id>]`，例如 `telegram:-100123#42` 或 `onebot11:group:123456`。用 `message.SID(adapterID, chatID, threadID)` 处理 chat_id 与 thread_id 中 `%`、`#` 的转义，用 `message.ParseSID(sid)` 校验并拆解；adapter_id 遵守 manifest 的本地 ID 规则。`c.Send(content)` 使用本次调用的默认会话，`c.SendTo(sid, content)` 发往指定会话，`sdk.Via(instanceID)` 指定机器人实例；只有 Via 时保留默认会话与话题。后台用 `sdk.Send(ctx, session, content, opts...)`，opts 只接受 Via。省略实例时由 Core 选择并补齐账号；非法会话在发出请求前返回 INVALID_ARGUMENT。
- 编辑、撤回与引用所指的消息统一用 `string` 消息句柄，格式为 `<会话字符串>/<platform_message_id>[@<实例名>]`，可以原样保存。用发送回执的 `Ref()`、`Refs()` 或 `c.Event().Ref()` 取得句柄，手工构造用 `message.Ref(sid, platformMessageID, instance)`，它负责消息 ID 中 `%`、`/`、`@` 的转义，用 `message.ParseRef(ref)` 校验并拆解。编辑用 `c.Edit(ref, content)` 或后台的 `sdk.Edit(ctx, ref, content)`，引用用 `message.Quote(ref)`；省略实例时由 Core 按目标选择规则补齐。
- 命令前缀与回复方式取当前会话的配置文件。处理 `message.created` 时，向原实例、会话与话题的首条 Send 由 Core 按设置加入群提及、引用与普通文本前缀；手动发送、自动回复、Service 派生与延续共享首条标记，BeforeSend 看到改写后的 chain。提及与前缀加入后超出实例文本上限时两项一起省略，引用保留；每次回复 Send 按 `forward_threshold` 与 `message.forward` 能力决定合并转发，组合仍须满足原消息校验。后台、定时任务与跨会话发送沿用原行为。
- 编辑首条回复时，Core 按当前插件 Session、完整目标与物理消息 ID 保留回执已确认发送的自动提及和前缀；重复编辑沿用已有修饰，实体偏移按 UTF-16 调整。Edit 接受一个 text（也可由通用 markdown 转换得到）与可选 buttons，需保留的普通文本前缀或自动提及与 Markdown 编辑组合无法表达时返回 UNSUPPORTED。Session 退出即释放关联；流式回复可直接沿用 `c.Edit` 与 `c.Stream`，修饰由 Core 处理。有效 mention/prefix 关联按需存入独立运行期 SQLite，空修饰跳过保存；整个 Session 内的早期消息仍可编辑。存储失败报告完整错误，已取得的平台回执与发送状态继续返回。
- 流式输出用 `c.Stream(next)`：next 每次返回一段增量文本，返回 `io.EOF` 正常结束（可以同时带上末段）。实例声明 `message.edit` 时先发送再节流编辑，超过 `TextLimit` 时换条；没有编辑能力时在正常结束后发送一次，输入出错时不发布缓存。返回全部回执，失败时保留已有回执。
- 回复文本用 `c.Reply(text)`；带片段时写 `c.Reply(text, message.ImageURL(url))`，文本与后续片段依次组成一条消息，只发送一次并返回错误。后台发文本用 `sdk.SendText(ctx, session, text, opts...)`，opts 只接受 Via，发送文本后丢弃回执，错误原样返回。
- `c.User()` 返回 `string` 用户字符串 `<adapter_id>:<external_user_id>`，没有 Actor 时为空串；`message.UID` 与 `message.ParseUID` 保留 external_user_id 中的冒号。`c.Session()` 返回默认会话，没有默认目标时为空串，未指定目标的发送返回 NO_DEFAULT_TARGET。Service 调用沿用调用上下文中的 Actor 与默认会话，`c.Event()` 为 nil。
- 事件用常量注册，例如 `p.On(sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error { return c.Reply(e.Chain.Text()) })`；也接受只收 `*sdk.Context` 的 Handler。其他标准事件同样使用对应常量与 payload 指针类型，注册收集阶段检查签名并在错误中给出 route_id。扩展事件用 `sdk.AdapterEvent[T]("adapter.name", version)`，T 须为结构体，带载荷的 Handler 收 `*T`；版本不符或 JSON 解码失败时记 WARN 并跳过 Handler，日志只含路由、事件名与版本或安全错误类别。
- 上传用 `media.Upload(ctx, r, mime)` 得到 `*message.MediaRef`，SDK 计算大小与 SHA-256，mime 为空时按内容判断；r 可 Seek 时读两遍，否则使用临时文件，调用方应保持两遍读取的内容一致。读取用 `media.Open(ctx, src)`，src 可以是片段的 `Source`、`MediaRef.Source()`、`IncomingMedia.Source()`，只持有 token 时用 `message.MediaSource{MediaRef: token}`。打开时先读第一块，失败直接返回错误；成功后顺序读取并关闭返回的 `io.ReadCloser`，读取响应 ctx 取消且不自动重试。HTTPS URL 来源在发出请求前返回 INVALID_ARGUMENT。向其他插件授权用 `media.Grant(ctx, token, granteePluginID, sessions...)`，会话参数为 `string`。
- 错误码用 `errcode`，例如 `errcode.InvalidArgument.New("invalid input")`、`errcode.InvalidArgument.Errorf("invalid field %s", name)`；用 `errcode.Of(err)` 查询错误链中的 SDK 错误码，没有 SDK 错误时返回 0。`sdk.Error` 是 `errcode.Error` 的别名，仍可用 `errors.As` 取得错误并读取 Code 与 Message。

发送带回调的按钮时，用 `a := button.Define[T](name)` 声明动作，通过 `button.On(r, a).Handle(...)` 注册签名为 `func(*sdk.Context, T) error` 的 Handler，再将 `a.Button(label, value)` 放入 `message.Buttons` 发送。`button.` 是按钮路由的保留前缀，注册其他路由时避开该前缀。value 用 `encoding/json` 编码，Core 限制编码结果最多 1024 字节；编码失败的按钮在发送前返回 INVALID_ARGUMENT。点击时 value 解码失败记 WARN，跳过 Handler 并正常结束本次调用。直接构造 `message.Button` 时，空的 `Value` 会省略该字段，省略值的点击同样按解码失败处理。`message.Extension(adapterID, name, version, payload)` 只能发往所属 Adapter 声明了对应 extension 名称与版本的实例，payload 按声明的 Schema 校验（[插件指南](plugin-guide.zh-CN.md)、[协议与会话行为](development.zh-CN.md#协议与会话行为)）。

读取合并转发时，当前 Handler 可直接调用 `nodes, err := c.ReadNodes(forward)`；返回值为递归的 `[]message.ForwardItem`，保留作者、`OccurredAt` 和节点消息链。节点中的图片、文件等来源已经按本次调用授权，可用 `media.Open(c, segment.Source)` 读取。查询错误可带部分节点，处理 `err` 后再决定如何使用。后台调用使用 `sdk.ReadNodes(ctx, sid, forward, sdk.Via(instanceID))`，沿用插件发信许可和指定实例。

通用 Markdown 使用对象方法转换为普通文本：

```go
md := message.Markdown("**标题**\n[链接](https://example.com)")
result, err := c.Send(md) // 支持 message.markdown 的平台由 Core 转换为实体。
result, err = c.Send(md.PlainText()) // 显式普通文本，方法无需参数。
```

通用语法采用 CommonMark 与删除线，保留块结构；富文本转换遇到图片或 HTML 时返回 `UNSUPPORTED`。平台未声明 `message.markdown` 时直接发送也返回该错误，调用者选择 `.PlainText()`。已有 `TextSegment.FormatMarkdown` 沿用 Adapter 的原生方言，Telegram 仍为 MarkdownV2。

Handler 内先调用 `client, err := c.Adapter()` 取得绑定当前可信实例与账号的 `*sdk.AdapterClient`，处理错误后调用 Adapter 包提供的带类型动作函数。Adapter 的 `client.From(c)` 可调用 `c.Adapter()` 并核对 `AdapterID()`；客户端还提供只读的 `InstanceID()`、`AccountID()` 与 `Session()`。没有可信平台身份时返回 `NO_DEFAULT_TARGET`，Handler 结束后调用返回 `FAILED_PRECONDITION`。

底层入口 `client.Call(name, version, params)` 与 `client.CallAction("", name, version, params)` 按 Adapter 声明的作用范围选择目标：chat 动作使用当前默认会话，account 动作只使用绑定实例、向执行器传空会话。账号事件也可取得客户端，调用 chat 动作时需显式提供 `to`。`CallAction(to, name, version, params, opts...)` 接受显式业务会话与 `sdk.Via(instanceID)`，跨会话、跨实例调用继续检查原有发信许可；结果 JSON 与错误同时保留。所有调用继承当前 Handler 的父调用、取消与期限。

`*sdk.AdapterClient`、`*sdk.Context` 与后台的 `*action.Client` 都实现 `action.Caller`。原入口 `c.CallAction` 保持显式目标语义，账号动作传空会话；后台沿用 `action.NewClient(ctx)`，由任务 context 控制取消与期限。

Adapter 作者用 `adapter.Event[T](name, version)` 与 `adapter.Extension[T](name, version)` 声明扩展事件与 extension；T 须为结构体，放在 Adapter 的常量与数据类型包中，供插件共用。SDK 从 Go 类型推导 Schema：没有 `omitempty` 的字段为必需，不允许未声明的属性；类型不符或推导失败时，`RunAdapter` 在 REGISTER 前返回带声明名的错误。

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

省略 `Auth` 时要求 Core 会话；`AuthPlatformToken` 由 Adapter 校验平台凭证；`AuthStatic` 公开提供 GET 与 HEAD 静态资源，无需管理会话或平台凭证。SDK 在连接 Core 前核对路径与重复路由。

查询当前会话的可用命令时，导入 `github.com/AlonaBot/AlonaBot/sdk/command`，在 Handler 内调用 `entries, err := command.List(c)`。成功时可用 `entry.Path` 和 `entry.Aliases` 显示生效命令名（多 token 路径用空格连接），用 `entry.Description` 显示本地化说明。调用失败时返回 `err`；后台或已结束的 Handler context 返回 `FAILED_PRECONDITION`。完整示例见[官方 help 插件](https://github.com/AlonaBot/alonabot-plugin-help/blob/main/help.go)。

### 命令菜单

根命令需要显示在平台菜单时写 `p.Command("hello").Menu().Description("Say hello").Handle(hello)`。缺省注册只提供处理器。自动菜单按聊天的 Profile、许可、公开权限、`/` 前缀和生效主名计算，论坛话题共用聊天菜单；关闭某个话题仍单独控制执行。

手动接管时，将目标 Telegram 实例的 `auto_sync_commands` 设为 false。`clear_commands_on_disable` 缺省 true，自动转手动先清理受管位置；false 保留到插件覆盖。首次手动及手动重启保留菜单。订阅 ready 后按聊天设置整份菜单：

```go
p.On(sdk.CommandMenuReady).Handle(func(c *sdk.Context, ready *message.CommandMenuReadyPayload) error {
    if ready.ChatID == "" {
        return nil
    }
    return command.Set(c, ready.AdapterInstanceID, command.Menu{
        Scope: command.Scope{Type: "chat", ChatID: ready.ChatID},
        Commands: []command.Item{{Name: "hello", Description: "Say hello"}},
    })
})
```

导入 `sdk/command` 和 `sdk/message`。需要翻译时逐 language_code 设置，空串为缺省语言，非空取两位小写主语言。另有 `chat_administrators` 与带 UserID 的 `chat_member` scope。每份最多 100 条；非法名字、说明、超限及语言映射冲突返回错误，保留已有成功菜单。每次成功 Set 转移该 scope+language 整份菜单的 owner，禁用/卸载旧 owner 保留新 owner 的内容。

`command.Get` 读取平台当前菜单，`Delete` 删除指定位置；`Status` 返回模式、Ready、LastError 和位置。普通自动同步在本 Session 跳过内容未变的位置。`command.Sync` 在自动模式全量重建活跃聊天菜单及受管 default，可用于修复远端漂移；手动模式重新处理清理与就绪通知。切换失败时读 LastError，修复原因后调用 Sync。插件的实例许可、聊天 Profile 与全局禁用在实际执行前重查；离群跳过写入，重入恢复。外部工具写入的历史 scope 需要明确给出位置再 Delete。

#### 菜单筛选、翻译与生命周期

自动筛选仅保留带 Menu 标记、主名胜出的单 token 路径，权限为空或包含 user，且插件可用、实例许可、聊天 Profile/插件开关、命令策略及 `/` 前缀均允许。可正常唤醒的休眠插件保留菜单。`.Description` 为英文原文；其他语言在 `i18n/<locale>.json` 顶层 `commands` 按完整规范路径覆盖，例如 `{"commands":{"rss add":{"description":"添加订阅"}}}`。路径含 Group，以一个 ASCII 空格连接；别名与生效主名共用规范路径。en-US 覆盖报错。check/build 检查结构，SDK Collect 拒绝未注册命令或未声明英文说明的覆盖路径。逐键按目标 locale、插件默认 locale、英文回落，未登记说明取主名；Core 将 locale 映射为两位小写主语言，同命令映射后说明冲突返回 INVALID_ARGUMENT。缺省菜单用英文。名字须匹配 `^[a-z0-9_]{1,32}$`，说明为1～256个 Unicode 字符，写前验证，错误保留成功菜单。

实例启动、插件安装/激活/卸载/禁用、许可、策略、Profile/assignment和前缀变化由实例串行 worker 从首次变化起合并2秒。首次消息和 Bot 自身入群后异步发布，普通消息继续派发；事件游标与账号聊天关系同事务保存。按 Adapter+account+scope+language 持久化模式及最后成功写入的插件，账号隔离，不逐条保存命令历史。受管 default 的已知语言菜单保持为空，thread 不进入位置。

本 Session 的内容快照仅在平台写入与位置记账均成功后保存；写入失败/UNKNOWN、删除、成员变化和模式切换使相关快照失效，新 Session 与重入重新发布。Set 整份覆盖并转移 owner；手动 owner 卸载、禁用、失去实例许可或被聊天 Profile 关闭时回收位置，自动模式删除失效 scope/语言。

自动转手动先停止调度并排空在途写入；首次手动、手动重启及只修改 clear 均保留菜单。离群位置保存进入手动时的清理决定，重入继续处理，之后单改 clear 保持原决定。清理失败保存完整错误并保持未就绪；成功后发 ready。手动转自动全量发布并移除失效手动位置。切换意图先持久化为新模式/未就绪；每项远端成功立即记账，错误或结果未知停止后续，存储错误完整返回，普通自动失败记诊断，后续变化或 Sync 可重处理。

Set/Delete 要求手动且 ready，插件只开放三个聊天 scope。调用检查当前 Session、插件可用性、allowed_adapter_instances；聊天操作再查 Profile 和全局禁用，空 instanceID 可取在途 Handler 的可信实例。`bot.command_menu_ready` 定向通知实例就绪、已有/新聊天、晚激活或重新获准；`CommandMenuReadyPayload` 含 AdapterID、AdapterInstanceID、AccountID、ChatID。空 ChatID 表示实例手动就绪，有值表示聊天可用；回调内可直接 Set。平台、传输与存储错误完整返回，业务写入无自动重试。

#### Adapter 菜单实现

声明 `commands.sync` v1 supported=true，并实现 `adapter.CommandMenuController`：`CommandMenuConfig()` 返回 `AutoSyncCommands` 与 `ClearCommandsOnDisable`；`SetCommandMenu(ctx, menu)`、`GetCommandMenu(ctx, scope, languageCode)`、`DeleteCommandMenu(ctx, scope, languageCode)` 每次只操作一个位置。Telegram 两项配置缺省均 true；其他实现由类型化配置报告。SDK 在启动响应报告配置，Core 经 `adapter.command.menu` 调用；未报告配置的既有 Adapter 沿用原 `CommandSynchronizer.SyncCommands`。

`CommandMenu{Scope, LanguageCode, Commands}` 中 Scope 含 Type、ChatID、UserID，Entry 含 Name、Description。Core 管理跨语言拆分、所有权、模式与持久化；Adapter 保留整份覆盖、空集合、完整平台错误及 UNKNOWN。SDK 实例执行锁跨平台调用持有（RPC取消后仍等平台调用返回），与旧 SyncCommands 共用锁；停止排空后才启动下一 worker。可信 member.left/member.joined 以成员ID等于账号ID识别 Bot 自身离群/重入。

使用新接口的 Adapter/插件需在 SDK/Core 正式发布后更新依赖；已有命令按需添加 `.Menu()`，官方 help 的 `/help`、`/whoami` 也需标记。

### 可信媒体 MIME 与平台助手

`message.MediaSource.MIME` 是 Core 准备发送时填写的可信来源元数据。MediaRef/File 取已保存的元数据；同实例入站对象以 `platform_file` 转发时取可信入站 MIME，空串表示未知。插件提交 MIME 返回 INVALID_ARGUMENT，`MediaRef.Source()` 只携带引用。租约与 SDK 转换保留 MIME，`adapter.Core.ReadMedia` 签名保持。Telegram 将 `image/gif` 作为动图，inline 内容检查 `GIF8` 前缀；HTTPS URL 不推断 MIME，按普通图片发送。

Adapter 公开 perm/event 包用 SDK builder 声明权限/事件；公开 action 助手通过 `sdk/action.Caller` 调用，可直接依赖 `sdk`、`sdk/action`、`sdk/message`、`sdk/errcode`，复用类型化选项、发送、消息句柄、流式助手与错误，保持不链接 Adapter 运行时。Telegram 输入状态助手每4秒刷新，ctx结束即停止；草稿助手结束时执行普通 Send。草稿动作不经过出站 Hook，最终 Send 经过 Hook。

## 6. 跨插件协作

- 用 Service Call 与 Event Bus 协作，不直接读写其他插件的文件或数据库。
- Service 与事件都先写成包级的声明，提供方与调用方共用同一个声明：`var Search = service.JSON[SearchIn, SearchOut]("weather/search")`（Protobuf 用 `service.Proto`），提供方在 `New()` 中 `service.Serve(p, Search).Handle(fn)`，调用方 `Search.Call(ctx, in)`；事件为 `var Updated = bus.Define[Update]("weather/updated")`，发布 `Updated.Publish(ctx, payload)`，订阅 `bus.Subscribe(p, Updated).Handle(fn)`，订阅的路由 ID 缺省为完整事件名，同一事件订阅多次时用 `.ID(routeID)` 区分。target 与事件名的组件部分必须是本插件的 ID，SDK 在握手取得插件 ID 后、REGISTER 之前检查。`@agent/<name>` 只应由 Agent 组件 Serve：SDK 只检查它的格式，Broker 只把 `@agent` 的调用转给 `role: agent` 的组件，普通插件写了也不会被找到。声明可以从其他 Go module 导入后直接调用，调用方的生成器会嵌入该声明的输入与输出 Schema，经包装函数调用时同样适用。
- Event Bus 是非持久广播：`Publish` 只能发布到本插件的命名空间，返回成功只表示 Core 准入，不代表有人收到；订阅方队列满、预算不足或离线时事件会丢失，也不会补发。需要回执或可靠送达时改用 Service Call。订阅回调串行执行，不要在其中阻塞：积压的事件占用本插件的接收预算，会让命令请求也被拒绝。回调里没有 Actor 与默认会话，发消息时用 `sdk.Send(ctx, session, content, ...)` 或 `sdk.SendText(ctx, session, text, ...)` 显式给出会话。
- `service.JSON`、`bus.Define` 与 `task.Define` 声明的类型参数必须是包级的具名非泛型类型，由生成器检查并给出带位置的诊断。JSON Service 与 Bus 事件的 schema_id 由类型决定，版本取类型的 `SchemaVersion() uint32`，没有该方法时为 1；改动类型后运行 `alonabot build`。切片与 map 接受 null，与 `encoding/json` 对 nil 的编码一致。
- JSON Service 的输入不符合 Schema 时，调用方收到 INVALID_ARGUMENT 和固定文本 `invalid service input`，Handler 不被调用；输出不符合时调用方收到 INTERNAL，日志只记 `schema_id`；调用方的 `Call` 也按声明校验响应，不符合时返回 INTERNAL 与 `invalid service output for "<target>"`。校验详情可能带业务载荷，不写入响应和日志。
- 通过 `tool.List(ctx)` 获取工具目录，`tool.Info.Effect` 返回声明的类别，策略按本次调用的可信默认 Target 所在会话选择配置文件；无 Target 或无父调用时使用 `default`。list、invoke 与 confirm 每次读取一次配置文件快照：插件关闭时为 deny，其余依次取配置文件覆盖、全局覆盖与工具声明缺省策略，确认操作无覆盖时为 confirm。工具与操作同名时分别匹配配置文件覆盖；管理更新从下一次操作生效，在途确认保留已接纳策略。用 `.Tool(description, tool.DefaultAllow)` 登记 JSON Service 工具，缺省需确认时使用 `tool.DefaultConfirm`。在 `.Tool(...)` 之后、`.Handle(...)` 之前可以接 `.Effect(tool.ReadOnly)`，另有 `tool.Write`、`tool.Destructive`；省略时为 `tool.Unspecified`。顺序或类别非法时收集注册报错；类别仅作提示，不改变策略。
- `sdk/llm.Generate(ctx, llm.Request{...})` 向 `@agent/llm.generate` 发送结构化文本生成请求，管理员通过 `llm_service_allowed_plugins` 授予调用权限。`sdk/agent.Run` 使用独立的 `@agent/agent.run` 契约。流式入口要求协商传输能力，完整参数与运行状态见 [Agent API](agent-api.zh-CN.md)。
- 在 Handler 或其延续调用中，`tool.Invoke(ctx, "plugin_id/service_name", args)` 按当前生效的工具策略调用工具（deny 返回 PERMISSION_DENIED；confirm 时 Core 先向当前会话发送带「允许」「拒绝」按钮的确认，展示实际参数，拒绝返回 PERMISSION_DENIED，超时返回 DEADLINE_EXCEEDED）；插件自行执行操作前用 `tool.Confirm(ctx, operation, summary, effect)` 请求许可，确认时展示插件给出的摘要，返回 nil 后可执行操作。无类别时传 `tool.Unspecified`，非法类别返回 INVALID_ARGUMENT。两者都需要父调用链；没有 Actor 时遇到确认返回 PERMISSION_DENIED，实例不支持按钮时返回 UNSUPPORTED。

### 插件页面

用 `p.Route("GET", "/subs/").Page("Subscriptions").Handle(h)` 将 HTTP 路由列为插件页面。`.Page` 写在 `Handle` 或 `HandleFunc` 前，参数为英文源标题；在 `i18n/zh-CN.json` 写入 `{"pages":{"/subs/":"订阅"}}` 即可覆盖该路径的标题。Core 在路由运行期间通过组件列表返回页面和标题（[管理 API](deployment.zh-CN.md#管理-api)）。

WebUI 提供沙箱 iframe 与 `postMessage` 桥（[管理 API](deployment.zh-CN.md#管理-api)）。声明 `.Page(title)` 后，SDK 自动在页面所在目录登记 `kit.css`、`kit.js` 和 `qrcode.js` 的 GET 路由，以公开静态资源（`sdk.AuthStatic`）提供；页面通过相对路径引用套件，即可使用与 WebUI 一致的基础样式、跟随明暗与语言切换，并转发 ⌘K/Ctrl+K 打开全局搜索。页面的第一行（插件名、标签与搜索）由 WebUI 绘制，页面里不要再画标题栏或搜索框。

官方 Agent 的五页实现见 `agent/pages.go` 与 `agent/pages/`：人格、MCP、技能、记忆与用户任务使用组件自带的 HTML/CSS/JS，标题写入 `pages`，文案写入 `messages`，通过桥请求本组件的管理路由（[管理 API](deployment.zh-CN.md#管理-api)）。记忆页逐页读取完整会话正文并管理笔记/档案；Core 的会话摘要与清除沿用 [管理 API](deployment.zh-CN.md#管理-api) 的管理入口。

#### 套件挂载与引入

在插件入口 `New()` 中声明页面：

```go
import (
    "net/http"
    "github.com/AlonaBot/AlonaBot/sdk"
)

func New() *sdk.Plugin {
    p := sdk.NewPlugin()

    p.Route(http.MethodGet, "/page").Page("Showcase").HandleFunc(servePage)
    return p
}
```

相对资源路径随页面 URL 解析：从 `/page` 打开时取 `/kit.css` 和 `/kit.js`，以 `/page/` 为基准时取 `/page/kit.css` 和 `/page/kit.js`（均在插件自己的 URL 前缀内）。作者已经登记的同方法、同路径路由优先保留；需要手动挂载套件时仍可使用 `page.Handler()`。页面运行在不透明源的沙箱中，加载样式、脚本、图片等子资源时不带管理 Cookie，所以页面自己的静态资源要以 `.Public(sdk.AuthStatic)` 登记（只限 GET 与 HEAD），否则浏览器只会拿到 401；页面 HTML 和数据接口保持缺省的管理会话认证，数据经 `window.AlonaBot.request` 由 WebUI 代为请求。完整示例见 [showcase](../examples/plugins/showcase/main.go)。

#### 页面 HTML 与控件样式

在返回的 HTML 中引用套件资源：

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <link rel="stylesheet" href="kit.css">
  <script src="kit.js"></script>
</head>
<body>
  <div class="ab-card">
    <div class="ab-card-header">插件管理</div>
    <div class="ab-card-body">
      <p id="greeting"></p>
      <button class="ab-btn ab-btn-primary" id="btn-save">保存</button>
      <input type="text" class="ab-input" placeholder="请输入...">
    </div>
  </div>
</body>
</html>
```

`kit.css` 只使用 `--ab-*` 设计令牌变量，为按钮（`button`、`.ab-btn`、`.ab-btn-primary`、`.ab-btn-danger`）、输入框（`input`、`textarea`、`.ab-input`）、选择框（`select`、`.ab-select`）、表格（`table`、`.ab-table`）、标签（`label`、`.ab-badge`）与卡片（`.ab-card`）提供好开箱即用的统一样式；未收到令牌时自动使用浅色缺省值，收到主题变更时即时跟随更新。

#### 客户端接口 `window.AlonaBot`

`kit.js` 加载后会在全局注入 `window.AlonaBot` 对象：

- **`request(method, path, body?)`**：经 WebUI 桥请求当前插件自己的 HTTP 路由。对象正文按 JSON 发送。Promise 返回的对象包含 `status`、`ok`（HTTP 2xx）、`headers`、`body`（UTF-8 文本）、`data`（解析成功的 JSON 或原文本）和原始 `body_base64`，也可调用 `text()` 或 `json()` 分别取得 `body` 或 `data`。桥错误会拒绝 Promise，HTTP 错误状态可从返回对象读取。
- **`t(key, args?) → Promise<string>`**：每次经桥发送 `translate`，传递 `key` 与可选的字符串参数对象 `args`，原样返回桥响应文案。桥按当前界面语言取插件文案并格式化 `{name}` 占位符；未知键返回键名，缺参数保留占位符。桥错误使 Promise 拒绝，独立打开时 Promise 返回键名。
- **账号配置**：`getConfigContext()` 与 `completeConfig({patch})` 分别发送 `config.context` 与 `config.complete`，见下文[账号配置页面](#adapter-账号配置页面)。
- **`onTheme(fn)`**：监听主题与令牌变化，回调签名 `fn(theme, tokens)`，注册时立即交付当前状态，返回取消订阅函数。
- **`onLocale(fn)`**：监听界面语言变化，回调签名 `fn(locale)`，注册时立即交付当前语言，返回取消订阅函数；页面在回调里重新 `await t(...)` 渲染。
- **快捷键转发**：页面内捕获 ⌘K（macOS）与 Ctrl+K（其他系统），自动向 WebUI 转发 `mod+k` 打开全局搜索。
- **独立打开**：自动应用浅色缺省值与字体；`request` 返回 rejected Promise，提示桥仅在 iframe 中可用。

在 `i18n/en-US.json` 的 `messages` 中声明 `"hello_message": "Hello, {name}!"`，其他语言文件提供同键翻译。页面在语言回调中等待文案后更新：

```js
AlonaBot.onLocale(async function(locale) {
  document.documentElement.lang = locale;
  try {
    document.getElementById('greeting').textContent =
      await AlonaBot.t('hello_message', { name: 'Alona' });
  } catch (err) {
    console.error('Translation failed:', err);
  }
});
```

### Adapter 账号配置页面

添加或编辑账号时，WebUI 可以在配置对话框中打开 Adapter 提供的页面，执行扫码登录、授权跳转或 Token 校验。在 `adapter.Descriptor.HTTP` 中登记一条 GET 路由，保留 `Auth` 的零值以要求管理会话。每个 Adapter 最多声明一个账号配置页面，`Title` 必填。

```go
return &adapter.Descriptor{
    AdapterID: "example",
    HTTP: []adapter.HTTPRoute{{
        Method: http.MethodGet,
        Path: "/account/",
        Handler: http.HandlerFunc(serveAccountPage),
        Page: &adapter.PageDescriptor{Title: "Connect account", Purpose: adapter.PageInstanceConfig},
    }},
}
```

Core 从 `RouteDescriptor.page_title` 取页面标题，并在组件列表的 `pages` 中发布 `{"path":"/account/","title":"Connect account","purpose":"instance_config"}`。协议字段 `page_purpose` 为空时按普通页面处理，列表中省略 `purpose`。

页面加载 `kit.js` 后，可通过 `alonabot.plugin` 桥调用：

| 接口 | op | 请求参数 | 成功结果 |
| --- | --- | --- | --- |
| `AlonaBot.getConfigContext()` | `config.context` | `{}` | `{mode: "create" \| "edit", instance_id: string, config: object, schema: object}` |
| `AlonaBot.completeConfig({patch})` | `config.complete` | `{patch: object}` | `{}` |

`config` 是当前表单配置，secret 仅提供脱敏值；`schema` 是 Adapter 实例配置 Schema。WebUI 收到补丁后按 RFC 7396 合入表单，仍由用户在对话框保存：省略字段保留原值，`null` 删除字段。新增或替换 secret 时在补丁中填明文值；保留原 secret 时省略字段，禁止提交 `$secret_set` 标记。若宿主当前没有账号配置上下文，两个调用都以带 `code: "FAILED_PRECONDITION"` 的 `Error` 拒绝。

首版流程只覆盖登录产出可长期使用的静态凭据。运行中需要刷新的凭据由 Adapter 按实例 ID 保存在自己的组件存储中，配置里只存账号标识。`create` 模式的 `instance_id` 只是草稿名，Adapter 不得用它持久化数据；登录任务使用 Adapter 自己生成的随机 ID。重新登录可在账号实例停止时发起，处理逻辑放在组件 HTTP 路由；状态查询使用约两秒一次的短轮询。

扫码页面在 `kit.js` 后单独加载同目录的 `qrcode.js`：

```html
<link rel="stylesheet" href="kit.css">
<div id="login-code" role="img" aria-label="Login QR code"></div>
<script src="kit.js"></script>
<script src="qrcode.js"></script>
```

拿到平台的扫码文本后，调用 `AlonaBot.renderQRCode(document.getElementById('login-code'), loginURL, {size: 256})`。函数以 UTF-8 字节编码，自动选择版本，使用 M 级纠错，并替换容器内容为跟随套件主题的 SVG。

## 7. 配置与 secret

Go 源文件可自由命名，入口由包级 `New` 及其签名决定；同包重复声明 `New` 会由 Go 编译器报告 `redeclared`。插件与 Agent 的 Schema 文件统一使用 `plugin_config_schema.json`，Adapter 继续使用 `adapter_config_schema.json`。

升级时先停用旧组件，备份数据并升级 Core，再安装使用新 SDK 重新构建的组件包，最后启用。配置值、secret 引用与受管数据沿用已有存储；无需手动重建配置。停用的旧包仍可经管理 API 更新。

在 `plugin.yaml` 同目录编写 `plugin_config_schema.json`，按字段键声明类型与英文源文案。运行 `alonabot build` 检查声明并把规范 JSON 经 overlay 编入程序；`alonabot check` 执行同样的检查。声明格式见 [插件指南](plugin-guide.zh-CN.md)，例如：

```json
{
  "greeting": {"type": "string", "description": "Greeting", "default": "Hello"},
  "enabled": {"type": "bool", "default": true},
  "token": {"type": "string", "description": "Access token", "secret": true}
}
```

- 配置没有版本，也不写迁移（[插件指南](plugin-guide.zh-CN.md)）：新增字段自动取默认值；删掉或改名的旧键保留为无效配置，管理员可通过管理 API 清理，WebUI 提示与清理按钮由 T-09.9 接入；类型不符的值先无损转换，不行再恢复默认值并提示管理员。
- 需要把旧值搬到新键，或检查业务取值时，用 `p.OnConfigLoad(fn)` 登记配置回调 `func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error)`。`raw` 是完整的原始配置，在补默认值与校正之前；判断旧键是否存在再搬移，保留其他键，响应 ctx，只做可重复的纯配置变换；没有登记时原样使用。
- 回调之后，SDK 给缺失或 `null` 的字段补 `default` 或类型零值，递归校正对象、列表与模板；类型不符时先无损转换，取值超出 options、slider 或数字范围时恢复默认值，失效的 template_list 元素被移除。每次转换、恢复和移除都记录路径、原因与原值，secret 字段省略原值；校正后 required 仍未满足时准备失败。未知键和数字精度保留。
- `GET /api/v1/plugins/{id}/config` 的 `corrections` 返回 Active 代次的校正记录，`unknown` 列出无效配置路径。`POST /api/v1/plugins/{id}/config/prune` 从 Active 删除所列无效键并写入 Pending，随后调用 reload 生效。普通 restart 保持同一配置快照，回调或校正改变结果时返回 `FAILED_PRECONDITION`。

Core 在首次成功激活后把配置导出到 `<data_dir>/plugins/<id>/config.json`，Secret 字段使用引用。编辑后执行 `alonabot reload <id>`：优先应用有效 Pending，否则导入该文件；文件缺失时从 Active 补建。成功应用后写回规范配置，准备期间的新编辑保留。更新、回退和应用 Pending 时，仅刷新缺失或仍与旧 Active 一致的文件，尚未应用的磁盘编辑保留。快照同步失败会报告错误，已提交的 Active 继续生效。

插件对象从 OnStart 开始通过 `p.Config()`（Handler 内为 `c.Config()`）取得 `config.Values`，子插件的 `s.Config()` 是 `plugin_config_schema.json` 中同名顶层 object 的只读视图；在 OnConfigLoad 中或更早调用会 panic。例如 `p.Config().String("greeting")` 读取上例的问候语。基础配置取值对象在同一 Session 内保持不变，`c.Config()` 冻结本次调用的配置文件视图；两者均可并发读取，集合 getter 返回可独立修改的副本。

给顶层字段或 `object` 子字段写 `"profile": true`，可在管理端按配置文件保存覆盖值。优先标记实际需要按会话变化的叶字段；整个复合字段被覆盖时，完整子树须无 `secret`，包括 list 元素与所有模板分支。在 OnStart 或后台读取的字段（令牌、连接、缓存）不要标记。Handler 的 `c.Config()` 已合并当前会话的有效覆盖，返回整个插件的只读配置；同一 Context 的视图保持不变。Rule、Service、Hook 与延续构造 Context 时同样冻结视图，覆盖更新从 SDK 接收之后开始的调用生效。`p.Config()`、`s.Config()`、OnStart 与后台读取保持基础配置。

覆盖以已经载入、补值及校正的配置为基线，按完整字段路径替换；object/list 整体替换，覆盖祖先 object 后子字段取新对象。`config.Values.WithOverrides(overrides map[string]json.RawMessage) (config.Values, error)` 可建立独立视图，保留当前子路径；覆盖只校验单值并替换，沿用原基础配置的载入与校正结果。

`sdk.ProfileOverrides(ctx context.Context) (map[string]map[string]json.RawMessage, error)` 可读取本插件各配置文件的有效覆盖值副本，外层键为配置文件 ID，内层为字段路径。map 与 RawMessage 均为深拷贝，无覆盖返回空 map；基础配置由 `p.Config()` 读取。传入所属 Handler、HTTP 或生命周期的 ctx，同进程插件按 Session 隔离；独立进程未绑定时沿用默认 Session，缺少 Session 或插件时返回 FAILED_PRECONDITION。每次读取取得当前已发布集合，通知后的读取取得新值，已有副本与 `c.Config()` 冻结视图保持原值。官方 Agent 在 `agent/personas.go` 中与基础配置、`core.profile.list` 合并计算人格直接及继承引用。

Core 提供 `core.profile.list`、`current`、`assign` 与 `set`（[协议与会话行为](development.zh-CN.md#协议与会话行为)）。current/assign/set 从同一 Session 的在途父调用取得可信会话与 Actor，assign/set 要求 Core Admin；set 只修改本插件标记了 `profile` 的字段。会话使用 `default` 时写入 Pending，已有 Pending 或管理互斥占用返回 FAILED_PRECONDITION；整条根调用及延续结束后触发 reload，响应的 `reloading` 为 true。配置文件关闭插件时，命令按禁用处理，文本、事件、Raw 与 PreEvent 候选跳过。

官方 Agent 的 `agent/profile.go` 提供调用示例：用包级 `service.Proto` 声明四个固定服务，调用传 `c`，用 `json.Marshal` 编码字段值，并按响应回复配置文件名称、会话数及 `reloading` 提示。`default` 的回复在当前 Handler 内完成，根调用结束后由 Core 重载。官方 Agent 的对话、定时任务和未显式指定模型的 `llm.generate` 均从 `c.Config()` 读取可覆盖字段。

| 声明类型 | 取值方法 |
| --- | --- |
| `string`、`text` | `String(key)` |
| `int`、`float`、`bool` | 分别为 `Int(key)`（int64）、`Float(key)`（float64）、`Bool(key)` |
| 没有 `items` 的 `list` | `Strings(key)` |
| `object`、有 `items` 的 `list` | `Object(key)`、`Objects(key)`，再按子字段键读取 |
| `dict` | `Dict(key)`，返回 `map[string]json.RawMessage` |
| `template`、`template_list` | `Template(key)`、`TemplateList(key)`，通过 `Key()` 取选中的模板键，嵌入的 `Values` 读取模板字段 |

可选 `template` 缺失时表示未选择，`Key()` 返回空串；选中时值须恰好包含一个已声明的模板键。读取未声明键或使用与声明类型不符的方法会 panic，消息指明字段路径。

secret 只用于 `string` 与 `text`，声明 `"secret": true`，同时省略 `default` 与 `options`。Core 按 schema 抽取对象、列表与模板中的秘密值：非空值生成不可变引用保存到配置快照，Bootstrap 解析为明文，`Config()` 取到明文；空串表示未设置。插件对秘密值保密，日志、错误和业务数据库中均禁止明文存放（[插件指南](plugin-guide.zh-CN.md)）。

配置翻译写在 `i18n/<locale>.json` 的 `config` 根节点，按 schema 的字段层级覆盖文案，例如：

```json
{"config": {"greeting": {"description": "问候语"}, "token": {"description": "访问令牌"}}}
```

字段可覆盖 `description`、`hint`、`labels`；子字段放在 `items`，模板放在 `templates`，模板文案可覆盖 `name`、`hint`。options 的存储值保持原文，翻译写在 labels 中。生成器核对覆盖键与 schema 的字段、模板和选项是否对应，问题诊断带文件名与路径。

- 字段的可选值来自运行时的数据（例如实例列表或本插件维护的条目）时，用 `options_from` 代替 `options`：`{"type": "string", "options_from": {"source": "core.adapter_instances", "filter": {"adapter_id": "telegram"}}}`，或指向本插件的选项服务 `"source": "<插件 ID>/<服务名>"`（[插件指南](plugin-guide.zh-CN.md)「选项来源」，）。选项服务用 `options.Serve(p, "feeds", func(ctx context.Context, q options.Query) (options.Page, error) {...})` 登记：按 `q.Search` 搜索、`q.Values` 按存储值取回显示名，最多返回 `q.Limit` 项，有更多结果时 `More` 为 true；输入、输出 Schema 由 SDK 固定提供，不用生成。该服务只由 Core 在管理页面打开表单时调用，访问固定为 PRIVATE。取值仍是普通字符串，来源不可用或值已不存在时配置照常加载，插件自己处理不认识的值。

### 配置字段的回复示例

要展示某个配置值对应的回复，在该字段上写 `example`。一个示例写成对象，多个示例写成数组；下面把问候语与一张包内图片组合起来：

```json
{
  "greeting": {
    "type": "string",
    "default": "Hello",
    "example": {
      "input": "hello",
      "reply": [
        {"text": "{greeting}"},
        {"image": "examples/greeting.png", "alt": "Greeting illustration"}
      ]
    }
  }
}
```

`input` 必填，写用户发出的文字并省略命令前缀；`reply` 必填，纯文字可以直接写字符串，混合回复按顺序列出片段。每个片段恰好写 `text` 或 `image`，图片可附 `alt`。图片放在工程的 `examples/` 下，格式为 PNG、JPEG、WebP 或 GIF，扩展名与内容须一致。`alonabot check`、`build` 和 `pack` 检查引用，`pack` 收集被引用的图片；安装与更新再次校验。

`input` 与 `text` 中用 `{字段路径}` 引用草稿中的 string、text、int、float 或 bool；对象子字段写成 `{group.greeting}`，带 labels 的选项显示当前语言的 label。模板用 `{mode.$key}` 取选中键，用 `{mode.template_key.child}` 取该模板的子字段，其他模板的值显示为空串。secret、list、dict、object 与 template_list 字段禁止作为引用值。字面花括号写成 `{{`、`}}`。

可选的 `condition` 按字段路径写期望值，例如 `"condition": {"mode.$key": "simple", "enabled": true}`，全部相等时显示该示例。字段被自身的显示条件隐藏时，示例一同隐藏。示例只声明固定回复；草稿替换与显示由 WebUI 完成。

翻译写在 `i18n/<locale>.json` 的 `config.<字段>.example` 下。沿用源 `example` 的对象或数组形状、源 `reply` 的字符串或片段数组形状；按位置覆盖 `input`、`reply` 字符串、片段的 `text` 或 `alt`。缺项沿用源文案，跳过数组中间一项时填 `{}`。`check`、`build` 和 `pack` 会核对覆盖键、文案类型、源形状、片段位置与占位符引用。

上例打包安装并激活后，Core 经管理认证提供 `GET /api/v1/packages/{id}/examples/greeting.png`。端点只提供 Active 包中 `plugin_config_schema.json` 引用的图片，并带与组件图标相同的安全响应头；详见 [管理 API](deployment.zh-CN.md#管理-api)。

## 8. 测试

- 用插件测试套件 `plugintest` 在 `go test` 中启动真实 Core 与测试 Adapter，注入消息并检查发送。
- 本仓库中的官方插件另遵守 [编码规范](coding-standards.zh-CN.md) 的测试提交规则。

## 9. 发布

- 市场 ID 按角色使用 `alonabot-plugin-`、`alonabot-adapter-`、`alonabot-agent-` 或 `alonabot-webui-` 前缀，后缀用单个连字符分词，完整 ID 不含下划线；插件正则为 `^alonabot-plugin-[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`，其他角色同形，长度上限为 64；本地规则仍为 `^[a-z][a-z0-9_-]{0,63}$`。manifest 可以用 `icon` 声明组件图标（SVG、PNG 或 WebP）。
- 改名时同步 `plugin.yaml`、依赖、HTTP 路径和 Service 原始名；`pack` 会使用新 ID 生成归档名。已有改名前的开发数据时，按[开发指南](development.zh-CN.md#开发运行与热部署)重建。
- 每次发布可附上 UTF-8 changelog，安装后可经[管理 API](deployment.zh-CN.md#管理-api)读取。二选一：根目录的 `CHANGELOG.md`（每版一节，例如 `## [1.2.1] - 2026-11-17`，日期可省略），或按版本、按语言存放，例如 `changelog/1.2.1/en-US.md` 和 `changelog/1.2.1/zh-CN.md`。目录名使用不带 `v`、且不高于 `plugin.yaml` 中 `version` 的 SemVer；文件名使用规范化的 BCP 47 标签，每个版本都要有 `en-US.md`。运行 `alonabot check` 核对格式，`alonabot pack` 原样打包。
- 官方 WebUI 预置插件源 `AlonaBot/plugin-store`，支持添加、移除和保留主动清空的源列表。商店读取源仓库默认分支的 `index.json`，从组件仓库最新 Release 的平台归档内 manifest 取得展示信息；发布条目的字段见 [release 定义](../contracts/release)。更新预览读取已绑定来源的 Release 正文，更新后读取包内 changelog；HTML 按文本显示，已读版本由 WebUI 后台按 SemVer 持久化并单调推进。
- 用户用 `install --source owner/repo` 安装社区组件，之后只从该来源更新；本地归档用 `install <归档>`，官方包用 `install <组件 ID>` 从可执行文件旁 `packages/<组件 ID>/` 读取。不带参数时在当前目录（作者工程为 `bin/`，含直接包子目录）按归档名选择唯一的本机包，静态 WebUI 使用 any/any，零个或多个候选时列出搜索目录和候选并报错；Adapter、Agent 与 WebUI 同类不能并存，替换官方组件需要 `--replace`。
- 用 `alonabot build` 检查资源并编译独立组件，`alonabot check` 只读核对。`alonabot pack [目录] --target linux/amd64 --target linux/arm64` 仅生成平台 tar.gz；省略目标时取本机平台，静态 WebUI 只生成一个 any 包，带 Go 入口的 WebUI 按平台构建。`--tags` 选择源码，`--out` 默认 `<工程>/bin`。Release 只上传这些归档，源码压缩包由 GitHub 自动提供。进程包包含可解析的 `build-info` 文本，静态 WebUI 包包含 manifest、静态资源和可选 changelog。

## 直接测试与构建

创建工程可运行 `alonabot init alonabot-plugin-hello`：ID 与外部 Go module 缺省取目录名，已有 module 内沿用其 go.mod。可用 `--id`、`--module`、`--role` 覆盖；Adapter 的 `--adapter-id` 缺省等于组件 ID。`my_plugin` 按本地 ID 创建并警告缺市场前缀；非法目录名报错并提示显式 `--id`。

`build`、`check`、`test`、`pack` 接受至多一个位置目录，省略时取当前目录，`--dir` 报未知参数。在组件目录运行 `go test ./...` 即可使用 plugintest。测试启动时按入口包现场扫描 manifest、Schema、i18n 和 Handler 参数名，并复用同一入口包的只读资源。作者的 `New()` 只创建插件并登记业务行为，工程里没有 `alonabot_gen.go`，也无需编写生成相关的代码或指令。

独立组件运行 `alonabot build .`；通过根包 `alonabot.Run` 或 `alonabot.Start` 装配组件的 main 包运行 `alonabot build ./cmd/bot`，执行产物为 `cmd/bot/bin/bot`（Windows 加 `.exe`）。库模式构建会只读扫描依赖中的组件，包括模块缓存中的第三方组件，并将资源经 main 包的 overlay 编入。直接 Go 构建的程序在启动缺少资源的组件时会提示使用 `alonabot build` 或 `alonabot run`。库模式的公开选项与停止用法见[开发指南](development.zh-CN.md#库模式运行同一进程)；独立进程开发运行与热部署使用 `alonabot run --reload`。

## Agent、流式服务与生命周期 API

[完整参数、能力状态、八阶段钩子与 Goal 恢复契约](agent-api.zh-CN.md)。
