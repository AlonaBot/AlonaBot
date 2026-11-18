# goal 语句：插件

本文件收录「插件」章节的 goal 语句，切片表在[插件实施计划](../../spec/plan/plugins.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-71 | T-14.1 | help 插件与命令列表 | T-04.4、T-04.16、T-05.2、T-08.8 | 待派发 |
| goal-72 | T-14.2 | welcome 与 rss 插件 | T-04.16、T-05.2、T-07.2、T-08.5、T-08.8、T-08.11、T-09.4 | 等待前置（T-04.16）；不进入首个发行版（D-171） |

## 第 18 波

### goal-71 · T-14.1 help 插件与命令列表

状态：待派发。前置 T-04.4（PR #88）、T-04.16（PR #199，`046e752`）、T-05.2（PR #145）与 T-08.8（PR #102）已合并。相关的 T-03.17（PR #150，`8eb7846`）、T-04.12（PR #154，`43a6135`）、T-04.14（PR #156，`0726d66`）、T-04.15（PR #197，`ba2d323`）与 T-13.5（PR #200，`2e5ac2b`）也已合并。2026-11-17 按 main `bae7faa` 的代码核对，并按 D-171 精简为 `/help` 与 `/whoami`。审查者定稿时确定：`core.command.list` 不返回参数用法（WIRE-FORMAT §3.5 同步）；主名被其他路由遮蔽的命令不列出，与命令菜单一致；`/help` 不按插件分组，只按命令名排序，因此不需要插件 name（CONTRACTS §9.1 同步）。核对事实如下：

- `contracts/proto/alonabot/v1/route.proto` 已有 `CommandAlias { repeated string path = 1; }`，以及 `CommandSpec.description = 13`（`optional string`）；命令列表的请求、响应与 `CommandInfo` 尚未定义。`contracts/wirespec_test.go` 通过 `ProtoReflect().Descriptor().Fields().ByNumber(...)` 检查字段名、类型、列表、presence 与消息全名。
- `internal/wire/services.go` 的固定服务沿用带文档注释的 `const ServiceX = "core.x"`。固定服务由 `router.Mux.Handle(name string, fn ServiceFunc)` 注册；`ServiceFunc` 为 `func(context.Context, Caller, *wire.Inbound) wire.Response`。装配在 `internal/app/assembly.go` 的 `assemble(...)`，其中 `a.mux = router.NewMux()`，Router 建好后注册各固定服务；根包已只保留库模式接口。
- 索引入口为 `func (r *Router) Commands() *CommandIndex`；查询为 `func (x *CommandIndex) Lookup(tokens []string) (entry CommandEntry, consumed int, ok bool)`。`CommandIndex.paths` 保存按 priority/plugin_id/route_id 排序的全部候选，`Lookup` 返回胜出者；`Shadows() []Shadow` 提供遮蔽关系。`CommandEntry` 含 `PluginID`、`RouteID`、`Priority`、`Path`、`Alias`、`Names`，`Names` 是覆盖后且已处理同插件声明冲突的名字列表，主名在前。`rebuildCommands()` 与 `declaredCommandNames(command *pb.CommandSpec) [][]string` 在 `internal/core/router/index.go`。
- 名称覆盖由 `router.Options.CommandNames func(pluginID string) map[string][][]string` 接入；实际装配为 `func (p *Policy) CommandNames(pluginID string) map[string][][]string`。完整启用与权限继承为 `func (p *Policy) CommandState(ctx context.Context, pluginID string, route *pb.RouteDescriptor) (enabled bool, permissions []string, err error)`，通过 `Options.CommandState` 接入。实际准入为 `func (p *Policy) CheckCommand(ctx context.Context, pluginID string, route *pb.RouteDescriptor, actor *pb.Actor, target *pb.Target) *pb.ErrorInfo`，同时处理全局禁止、逐层禁用和权限集合。
- Router 中仍有 `func (r *Router) commandPermitted(c candidate, ev *pb.Event) bool`、`func commandUsage(c candidate, ev *pb.Event, prefix string) (string, error)`，以及菜单实现 `func (r *Router) commandMenu(ctx context.Context, instanceID string) ([]*pb.CommandMenuItem, error)`。`commandPermitted` 将策略错误压成布尔值；目录查询直接复用 `CheckCommand`，以保留系统故障。说明翻译仍是 `func (c *Catalog) CommandDescription(locales []string, path []string, source string) string`，目录与缺省语言分别来自 `router.Plugin.Messages() *i18n.Catalog`、`DefaultLocale() string`。
- 插件入口仍是 `func New() *sdk.Plugin`，内部使用 `sdk.NewPlugin()` 和 `p.Command(...).Description(...).Handle(...)`；文案仍用 `func (c *sdk.Context) T(key string, args ...any) string`。`sdk.UseGenerated` 已删除，资源由构建工具通过 overlay 按入口包登记。分类 API 放子包；本草案将原拟 `ListCommands` 写为 `sdk/command.List`，返回 `command.Entry`。`c.User() string` 已给出用户字符串，`c.Instance() string` 已给出默认目标的实例名。
- 官方组件结构参照 `agent/agent.go`、`agent/plugin.yaml`、`agent/cmd/alonabot-agent/main.go`；i18n 的 `messages` 与 `commands` 结构参照现有示例和 `contracts/i18n`。现有 CLI 为 `alonabot check --dir <工程>`、`alonabot pack --dir <工程> --asset-url <绝对 https URL>`；位置参数简化尚属后继切片。原根包的流水线测试已移到 `internal/app/pipeline_test.go`。

```text
/goal 使用 $alonabot-execute 推进 goal-71（切片 T-14.1：help 插件与命令列表）。goal 目录 goal-71，分支 goal-71，worktree .worktrees/goal-71，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 提供 core.command.list，返回调用者 Actor 在默认 Target 中可用的路径命令，包含生效的主名、其余未被遮蔽的生效命令名、按 locale 选择的说明与所属插件；官方 plugins/help/ 只注册 /help 与 /whoami。/help 按命令名排序列出命令名与说明，/whoami 返回用户字符串与收到消息的实例名，供添加第一个 Core Admin。

依据：spec/PLAN.md §3.1；spec/plan/plugins.md「T-14 的预装插件步骤」步骤 1；spec/plan/release.md「T-14」的前置输入与失败停止条件；CONTRACTS §2.2、§2.3、§5.4、§8、§9.1、§15.2；WIRE-FORMAT §3.5、§4、§5A；DECISIONS D-69、D-86、D-105、D-114、D-128、D-138、D-144～D-146、D-148、D-171；TEST-MATRIX V-67 的 help 分项、V-115 的 core.command.list 分项；ARCHITECTURE 的 SDK 分类与依赖方向；docs/plugin-guide.md。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/command.proto：新增下述三个消息，import 已有 route.proto 复用 CommandAlias；本语句明确授权这一个新 .proto。contracts/pb/alonabotv1/command.pb.go 由 go generate ./... 生成。
- contracts/wirespec_test.go：新增一项命令列表协议编号检查。
- internal/wire/services.go：追加 ServiceCommandList 常量。
- internal/core/router/commandlist.go：列表服务、筛选与注册；复用当前命令索引、策略接口与翻译。
- internal/app/assembly.go：在 assemble 中为现有 a.router 增加一次服务注册。
- sdk/command/command.go：新的按需类别包，提供 Entry 与 List。
- plugins/help/help.go、plugins/help/plugin.yaml、plugins/help/i18n/en-US.json、plugins/help/i18n/zh-CN.json、plugins/help/cmd/alonabot-help/main.go：插件、资源与薄入口。
- internal/core/router/、sdk/command/、plugins/help/ 下的 *_local_test.go，以及 internal/app/help_local_test.go：本地自测与验收取证。

允许的新跨模块边：sdk/command -> sdk、sdk/errcode、sdk/internal/registry、contracts/pb/alonabotv1、internal/wire，沿用分类 SDK 的方向；plugins/help -> sdk、sdk/command；plugins/help/cmd/alonabot-help -> plugins/help、sdk；internal/app 的 help_local_test.go -> internal/pluginkit，仅用于把实际 help 插件构建到测试临时目录。Core 服务留在 router 包内。根 sdk 包维持现有依赖方向。无新增第三方依赖。

接口（照此实现；内部辅助函数自定）：
1. contracts/proto/alonabot/v1/command.proto（package alonabot.v1，go_package 与其他 proto 一致）：
   message CommandListRequest {}
   message CommandInfo {
     string plugin_id = 1;
     string route_id = 2;
     repeated string path = 3;
     repeated CommandAlias aliases = 4;
     optional string description = 5;
   }
   message CommandListResponse { repeated CommandInfo commands = 1; }
   CommandSpec.description 和 CommandAlias 沿用已有定义。
2. internal/wire：const ServiceCommandList = "core.command.list"，按现有固定服务常量风格写文档注释。
3. internal/core/router：func (r *Router) RegisterCommandList(m *Mux)。内部 Handler 签名为 func (r *Router) listCommands(ctx context.Context, caller Caller, in *wire.Inbound) wire.Response；注册通过 m.Handle(wire.ServiceCommandList, r.listCommands)。internal/app/assembly.go 在 a.router 建好后调用 a.router.RegisterCommandList(a.mux)。
4. sdk/command：
   type Entry struct {
     PluginID, RouteID string
     Path []string
     Aliases [][]string
     Description string
   }
   func List(ctx context.Context) ([]Entry, error)
   Path 为生效主名；Aliases 为其余可列出的名字，保持配置顺序；Description 为空表示没有说明。原拟 ListCommands 收敛为 command.List，名字与现有 tool.List、conversation.List 一致。
5. plugins/help（package help）：func New() *sdk.Plugin。New 内直接创建对象并登记两个命令，Handler 都只收 *sdk.Context；两条命令的缺省权限均为 sdk.User，保持 Core 默认的参数错误 usage。
   cmd/alonabot-help/main.go 只调用 sdk.Run(help.New())；错误按现有构建薄入口的方式记英文 slog 并以非零状态退出。

行为：
- 固定服务的身份与协议：要求 in.ParentCallID 非零且 in.ParentAttach 为有效的 *router.Invocation，否则返回 FAILED_PRECONDITION。父调用仍在途、同 Session 关联与取消由 wire.Session 负责。Actor 只取 parent.Event.GetActor()，目标只取 parent.DefaultTarget；请求体为空，不接收插件自报的 Actor、权限或目标。请求与响应使用 Protobuf、validate.SchemaID 对应消息和版本 1；错误的请求 encoding/schema/version、无效 Protobuf 返回 INVALID_ARGUMENT；编码响应失败返回 INTERNAL，保留错误原因。缺少 Actor 的有效父调用仍可查询，按实际的空 Actor 权限规则筛选。
- 索引与遮蔽：在同一次 r.mu 临界区取得 r.commands 与 r.matches 的快照，随后解锁；策略读取、翻译和响应编码均在锁外执行。只枚举 COMMAND_MATCH_PATH 且插件仍 Activated 的路由，每个 plugin_id/route_id 最多一条。生效名字取同一索引快照的 names[pluginID][routeID]；沿用旧 goal 与 commandMenu 的主名胜出条件：Names[0] 的 Lookup 胜出者必须属于该路由，主名被其他路由遮蔽时整条省略。Path 取 Names[0]；Aliases 按 Names[1:] 原序逐个核对 Lookup，只保留胜出者仍属于该路由的名字。别名不另占一条目录记录；主名胜出时也收录多 token 命令。禁用或权限不满足的胜出者继续占有路径，遮蔽者保持隐藏；覆盖取代后的旧名字保持隐藏。
- 启用与权限：生产装配继续使用现有 Policy。每条候选调用 r.policy.CheckCommand(ctx, pluginID, route, actor, parent.DefaultTarget)，一次复用全局禁止、插件/组/路由启用继承与最近显式权限的语义。PERMISSION_DENIED、FAILED_PRECONDITION 对应的命令省略；其他错误作为本次列表失败原样返回，保留安全消息。r.policy 为 nil 的既有轻量装配按注册描述的 DefaultEnabled 与 RequiredPermissions 判定，权限语义复用 r.permitted；空权限集合允许没有 Actor。列表查询沿用策略本身的权限判断，不运行候选的 SDK Rule 或 Handler。
- 名称与说明：覆盖结果来自索引，路由身份、组策略与翻译键保持使用原规范路径。说明英文源文为 route.GetCommand().GetDescription()；有源文且有目录时调用 catalog.CommandDescription([]string{actor.GetLocale(), plugin.DefaultLocale()}, command.GetPath(), source)，缺目录直接使用英文源文。没有登记说明时 description 为 absent，SDK 转成空串，列表只显示名字。
- 排序：响应按 strings.Join(path, " ") 字典序排列。
- SDK：List 只在有效调用上下文内使用，调用方传 c.Context()。先用 sdk/internal/registry.Parent 检查有效父调用，后台或已结束的 Context 返回 errcode.FailedPrecondition；再用 registry.Proto[*pb.CommandListRequest, *pb.CommandListResponse](wire.ServiceCommandList).Call(ctx, &pb.CommandListRequest{}) 复用就绪门禁、父关联、取消、Schema 校验和错误转换。逐项转换为 Entry，不向插件公开 pb 类型；服务失败时返回原错误。
- /help：调用 command.List(c.Context())，失败直接返回错误。按返回顺序每条一行，列出 Path 和 Aliases（token 以一个 ASCII 空格连接，主名在前），有 Description 时附上说明；不按插件分组。标题、空列表和行格式经 c.T 读取 messages；没有说明的行只展示名字。名称原样显示，不硬编码斜杠前缀。列表包含调用者可用的 help 与 whoami 本身；筛选由 Core 完成，插件只负责排版。
- /whoami：通过 c.User() 与 c.Instance() 取得用户字符串和收到消息的实例名，调用 c.T("help.whoami", "user", c.User(), "instance", c.Instance()) 并回复。文案分别在 en-US、zh-CN 显示这两个原值；输出的用户字符串可直接用于 core_config.json 的 admins、WebUI Core 设置或 alonabot admin add（CONTRACTS §15.2）。命令要求 user 权限，首次使用无需已有 Core Admin。策略若被改为允许没有 Actor 的调用，c.User() 为空时回复 help.whoami.no_actor。
- 注册与资源：New 中直接登记 p.Command("help").Description("List available commands") 与 p.Command("whoami").Description("Show your user ID and instance")，然后 Handle；保留英文源文，zh-CN 的顶层 commands 分别以 help、whoami 作为规范路径覆盖说明。en-US 的英文说明留在代码，i18n/en-US.json 提供 messages，zh-CN 同时提供 messages 与 commands。无需配置、数据库、生命周期状态或子插件。资源由 alonabot build 的 overlay 按入口包登记。
- 工程 manifest：plugin.yaml 写 id: alonabot-plugin-help-official、name: Help、version: 0.1.0、author: AlonaBot、role: plugin、source: github.com/AlonaBot/AlonaBot 及英文 description；协议、平台、入口、配置版本、迁移方式与 prepare_effects 由工具推导。独立工程可导入，cmd 薄入口只负责运行。当前 check 使用 --dir；pack 所需的构建 manifest、build-info、发行 URL 与归档交由 T-14.3 产出。

测试（按编码规范「测试提交规则」与 docs/goals/README.md，D-86）：
- 提交：只在 contracts/wirespec_test.go 新增 TestCommandListFieldNumbers 一项编号检查。沿用现有反射断言风格，检查空请求、响应 commands = 1 的 repeated CommandInfo、CommandInfo 的五个字段的编号/名称/类型/cardinality、aliases 的消息类型 alonabot.v1.CommandAlias、description 的 optional presence；同时确认复用的 CommandAlias.path = 1 与 CommandSpec.description = 13。其余验证全部为本地自测。
- 本地自测（*_local_test.go，结果写入 PR）：
  - internal/app/help_local_test.go（unix）：复用 internal/app/pipeline_test.go 的真实 Core、Session 与回环 Adapter 装配，以 internal/pluginkit.BuildDev 将实际 plugins/help 构建到测试临时目录并作为独立进程运行。其他候选在本地测试文件中登记，资源通过已有 internal/generated 机制绑定；使用临时状态库与 core_config.json。普通用户发送 /help，只见可用命令及本地化说明；发送 /whoami，收到正确用户字符串与实例名；将所得用户加入 Core Admin 后，/help 才出现显式要求 sdk.Admin 的命令。主路径验证实际插件、SDK 调用、Mux 注册和 Core 筛选串联。
  - router 的本地测试：缺少 parent_call_id、失效父调用、错误附件与错误 encoding/schema/version/Protobuf；有父调用却没有 Actor 时只列空权限集合的命令；插件、父组和路由禁用；全局用户/会话禁止；Admin 与 chat_admin 的区别；被禁用或被拒的胜出者继续遮蔽低优先级命令；系统故障保留错误；文本匹配和 Raw 省略；排序、别名去重成单条、别名部分被遮蔽、主名被遮蔽而别名仍胜出时整条省略。
  - 名称与翻译：覆盖后只显示生效名字，调整主名后顺序正确，route_id 与规范路径翻译键保持；zh-CN、en-US、缺翻译时的插件默认 locale 与英文回落；没有登记说明；多 token 子命令。
  - SDK 与插件：command.List 的父上下文继承、返回值转换与 Core 错误；/help 的排序、主名/别名、空列表与缺说明排版；/help 多余参数由 Core 回复 usage，help Handler 保持零参数；/whoami 在相同 Adapter 的两个实例中返回同一用户字符串与各自的实例名，没有 Actor 的文案按上文取证；构造两次 New 得到独立插件对象。
  - 在仓库根执行 go run ./cmd/alonabot check --dir ./plugins/help，确认资源与入口通过检查。测试插件均留在本地测试文件中。

完成条件：上述本地证据齐备，在 PR 按 TEST-MATRIX §3 写明命令、平台、结果与未运行项；gofmt -l 对改动 Go 文件无输出；go generate ./... 生成命令列表协议，随后再次生成无新增差异；go vet ./...；go test -race ./...；go build ./...；help 工程的 check 通过。安装包、预装和真实平台证据由对应后继切片验收。出现身份/权限越界、遮蔽语义变化、需要增加依赖或触及本范围外文件时，按共同约束与工作包失败停止条件记录阻塞。

不做：/help <命令>；参数用法（由 Core 的 usage 回复给出）；按插件分组；welcome 与 rss（T-14.2，D-171 不进入首个发行版）；安装包、签名、发行归档与预装（T-14.3、T-14.4）；管理页面中的命令列表（T-09.9）；命令搜索、分页、缓存与自动重试；新增依赖。
```

### goal-72 · T-14.2 welcome 与 rss 插件

状态：待派发（老师确认暂不派发）。T-04.16（[PR #199](https://github.com/AlonaBot/AlonaBot/pull/199)，合并提交 `046e752`）已合并；2026-11-14 审查者按 main 核对：本语句只用到根包保留的 `sdk.ChatAdmin`、`sdk.Admin`、`p.Every`、`sdk.MemberJoined`、`sdk.Send` 与 `sqlitedb`，不涉及分包后的子包，无需改写。此前：等待前置（T-04.16），维护者在 PR #196 确认加入该前置。T-09.4（[PR #163](https://github.com/AlonaBot/AlonaBot/pull/163)，合并提交 `cb8ae3d`）已合并。T-05.2（[PR #145](https://github.com/AlonaBot/AlonaBot/pull/145)）、T-07.2（PR #107）、T-08.5（PR #79）、T-08.8（PR #102）与 T-08.11（PR #103）已合并。2026-11-12 审查者按 main 改写（老师确认只改写、暂不派发），前置加入 T-09.4：提交的测试在 plugins/ 中以根包库模式 `alonabot.Start` 运行插件与测试中定义的 Adapter，测试 Adapter 的实例经 `core_config.json` 启动、Core Admin 从同一文件读取，由 T-09.4 提供。改写要点：权限按 D-69 写成 `sdk.ChatAdmin.Or(sdk.Admin)`；定时任务按 T-04.11 写成 `p.Every(jobID, d).Handle(fn)`，注册时读不到配置，所以固定每分钟一次，按订阅的上次检查时间与 `interval_minutes` 决定是否抓取；数据库用 `sqlitedb.Open(ctx)` 与 `sqlitedb.Migrations(p, ...)`；事件为 `sdk.MemberJoined`（`*message.MemberPayload`）；后台发送用 `sdk.Send(ctx, session, content)`，发信许可缺省为全部实例（D-83）。此前的记录：T-07.2（PR #107，合并提交 `fc2c247`）已合并，合入的符号：`internal/core/scheduler` 的 `New(Options{JobTimeout, Now})`、`Scheduler.Attach(router.Plugin, activatedAt)`/`Detach`/`Released`/`Close`/`Stats(pluginID, jobID) JobStats`；`wire.Request.Late`；SDK 的 `ScheduleRegistry.Every`/`Cron` 已随 REGISTER 声明并分发到 Handler。T-08.8（PR #102，合并提交 `f685c92`）已合并，合入的符号：`contracts/i18n` 的 `Parse(files) (*Catalog, []Problem)`、`Catalog.Has`/`Format(locales, key, args)`、`Placeholders`；`sdk.Context.T`、后台 `sdk.T` 与 `Generated.I18n`；`internal/core/i18n.Messages()`；`host.Instance.Messages`/`DefaultLocale`；`router.Plugin` 的 `Messages()`/`DefaultLocale()`；`supervisor.Spec.I18n`/`DefaultLocale`；`Bootstrap.default_locale = 12`。T-08.5（PR #79）已合并。goal-95（PR #103，合并提交 `c840af7`）已合并，合入的接口：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)` 与 `Schema` 的 `Canonical`、`Hash`、`Version`、`Migrate(config, from)`、`Complete`、`Validate`、`SecretPaths`、`CheckI18n(config)`；`sdk/config` 的 `Prepare(ctx, load, schema, raw, fromVersion, targetVersion, restart)`、`Loader`、只读的 `Values`（`String`/`Int`/`Float`/`Bool`/`Strings`/`Object`/`Objects`/`Dict`/`Template`/`TemplateList`）与 `Template.Key`；`sdk.Base.Config() config.Values` 与根插件的 `OnConfigLoad(ctx, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`；`supervisor.Spec.ConfigSchema` 为 config_schema.json 的规范 JSON。生成器读取手写 `config_schema.json`，配置标签解析与 `config.Config`/`config.Base` 已删除。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。语句中 goal-16、goal-24、goal-32、goal-53、goal-95 提供的符号（ChatAdmin Rule、ScheduleRegistry.Every、gormdb.Open、Context.T、sdk.Send、Config() 与 OnConfigLoad）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：rss 的后台推送按缺省的发信许可即可发送（D-83）。

```text
/goal 使用 $alonabot-execute 推进 goal-72（切片 T-14.2：welcome 与 rss 插件）。goal 目录 goal-72，分支 goal-72，worktree .worktrees/goal-72，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：预装插件 plugins/welcome/ 在 member.joined 时向该聊天发送欢迎语，按聊天开关与自定义文案；plugins/rss/ 提供 /rss add|list|remove，订阅经 sqlitedb.Open 与 sqlitedb.Migrations 保存在插件数据库（不链接 GORM，D-49），固定 interval 的定时任务抓取 RSS 2.0 与 Atom，按 guid 或链接去重，只推送新条目；抓取只允许 http/https，拒绝 loopback、私网、链路本地、组播与未指定地址（含每次重定向之后），响应上限 5 MiB、超时 30s。

依据：PLAN T-14 步骤 2；CONTRACTS §15.2、§8（后台 Send 需要许可）、§5.4；TEST-MATRIX V-67。

范围（只新增或修改这些）：
- 新增 plugins/welcome/、plugins/rss/（各含可导入的包、cmd/ 下的 main、plugin.yaml、i18n/en-US.json 与 zh-CN.json）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> sdk/...、-> contracts（已允许）。只用标准库解析 XML（encoding/xml）与 HTTP。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- plugins/welcome：func New() *sdk.Plugin
- plugins/rss：func New() *sdk.Plugin；plugins/rss/config_schema.json 声明 {"interval_minutes": {"description": "Check interval (minutes)", "type": "int", "default": 10, "slider": {"min": 1, "max": 1440}}}，以 Config().Int("interval_minutes") 读取（D-76）

行为：
- welcome：
  - /welcome on|off|set <text:raw>：只有 chat_admin 或 Core Admin 能执行（命令的 Permission 为 sdk.ChatAdmin.Or(sdk.Admin)）；设置按聊天（instance、chat、thread）保存在插件数据库。默认关闭。
  - member.joined 时，该聊天已开启则发送欢迎语：自定义文案或 messages 中的 welcome.default；{name} 替换为新成员的显示名（没有时为 external_user_id）。回复原 Target，不需要许可。
  - 只响应 member.joined。
- rss：
  - /rss add <url> 与 /rss remove <url>：只有 chat_admin 或 Core Admin（Permission 为 sdk.ChatAdmin.Or(sdk.Admin)）；/rss list 不限。订阅（url、目标 Target、最后条目标识集合的摘要）保存在插件数据库。add 时先做一次抓取校验（地址规则同下），失败时回复原因（文案键）。
  - 定时任务：p.Every("check", time.Minute).Handle(fn)，每次只抓取上次检查早于 interval_minutes 的订阅（上次检查时间保存在插件数据库）；每个订阅抓取后解析 RSS 2.0（channel/item 的 guid、link、title）与 Atom（entry 的 id、link[rel=alternate 或缺省]、title）；条目标识为 guid（或 id），没有时为链接；只推送此前未见过的条目（首次订阅时只记录现有条目，不推送），按发布顺序发送，经 sdk.Send(ctx, 订阅的会话, content) 发送（发信许可缺省为全部实例；被管理员收紧而拒绝时记 WARN 并跳过）。
  - 抓取：只允许 http 与 https；自定义 http.Client 的 DialContext 在连接前解析地址并拒绝 IsLoopback、IsPrivate、IsLinkLocalUnicast、IsLinkLocalMulticast、IsMulticast、IsUnspecified 的地址；CheckRedirect 对每次重定向的 URL 同样检查 scheme，连接时再检查地址；响应体以 io.LimitReader 读取 5 MiB 加一字节，超过即失败；整个抓取超时 30s（命名常量）。失败记 WARN（订阅 URL 与原因），不推送、不重试到下一个周期之前。

测试（按编码规范「测试提交规则」）：
- 提交（多列理由：rss 与 welcome 是两个独立的插件）：
  - plugins/rss/rss_test.go（//go:build unix）：以根包库模式 alonabot.Start 运行 rss 插件与测试中定义的回环 Adapter（实例写在 core_config.json，声明成员事件与发送），httptest 提供订阅源。注意 httptest 监听 loopback，测试通过包内的测试钩子把「允许 loopback」打开（钩子只在测试中设置，生产默认关闭），并另有子测试在钩子关闭时验证 loopback 被拒：
    - 普通用户 /rss add 被拒；chat_admin 成功；list 显示订阅。
    - 首次抓取不推送；源中新增两个条目后下一次定时任务推送两条，顺序正确；再次运行不重复推送。
    - 响应超过 5 MiB、超过 30s（注入更短的超时）、重定向到 127.0.0.1（钩子关闭）、file:// 与 ftp:// 都失败。
  - plugins/welcome/welcome_test.go（//go:build unix，装配同上）：开关权限、自定义文案与 {name} 替换、关闭时不发送、只响应 member.joined。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：地址检查函数（IPv4/IPv6 私网、链路本地、映射地址 ::ffff:127.0.0.1）；条目标识的优先级；XML 中缺少字段的条目；发信许可被收紧时不推送并记 WARN；恢复缺省后推送；Atom 与 RSS 2.0 各一个源。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...；两个插件目录运行 alonabot plugin check 通过。

不做：发行归档与预装（T-14.3、T-14.4）；HTML 清洗与富文本推送；代理设置（沿用进程环境的 HTTP_PROXY）。
```
