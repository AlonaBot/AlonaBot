# goal 语句：插件

本文件收录「插件」章节的 goal 语句，切片表在[插件实施计划](../../spec/plan/plugins.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-71 | T-14.1 | help 插件与命令列表 | T-04.4、T-05.2、T-08.8 | 等待前置 |
| goal-72 | T-14.2 | welcome 与 rss 插件 | T-05.2、T-07.2、T-08.3、T-08.5、T-08.8 | 等待前置 |

## 第 11 波

### goal-71 · T-14.1 help 插件与命令列表

状态：等待前置（T-04.4、T-05.2、T-08.8）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。语句中 goal-10、goal-12、goal-13、goal-16、goal-53 提供的符号（router.CommandIndex、policy.CheckCommand、contracts/i18n、sdk 的命令注册与 Context.T）只写名字；合并后由审查者核对签名，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-71（切片 T-14.1：help 插件与命令列表）。goal 目录 goal-71，分支 goal-71，worktree .worktrees/goal-71，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 提供 core.command.list：按调用者 Actor 与默认 Target 返回可用的命令（已启用、权限满足、未被遮蔽），说明按 CommandSpec 的说明键在所属插件的 messages 中按 Actor locale 翻译。官方 help 插件 plugins/help/ 注册 /help [命令] 与 /whoami。

依据：PLAN T-14 步骤 1；CONTRACTS §15.2、§2.2、§2.3、§5.4；WIRE-FORMAT §3.5（core.command.list 与 CommandSpec 的说明键）；TEST-MATRIX V-67 的 help 分项。

范围（只新增或修改这些）：
- contracts/proto/alonabot/v1/：route.proto 的 CommandSpec 追加 optional string description_key = 12；新增 CommandListRequest/Response 与 CommandInfo。随后 go generate ./...（本切片获授权只追加这些）
- contracts/wirespec_test.go：追加编号检查
- internal/wire/services.go：末尾追加 ServiceCommandList = "core.command.list"
- 新增 internal/core/commandlist/
- internal/core/supervisor 或 cmd/alonabot 的装配处：调用 SetMessages（只加调用点）
- sdk/：Description(key string) 路由选项；buildRegister 写入 description_key；新增 commands.go（ListCommands）
- 新增 plugins/help/（main 包与可导入的 help 包，含 plugin.yaml 与 i18n/en-US.json、zh-CN.json）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> sdk/...、-> contracts（已允许）。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. proto
   message CommandListRequest {}
   message CommandInfo { string plugin_id = 1; string route_id = 2; repeated string path = 3; repeated CommandAlias aliases = 4; string usage = 5; optional string description = 6; }
   message CommandListResponse { repeated CommandInfo commands = 1; }
2. internal/core/commandlist
   type Service struct{ ... }
   func New(r *router.Router, p *policy.Policy) *Service
   func (s *Service) Register(m *router.Mux)
   func (s *Service) SetMessages(pluginID string, files map[string][]byte) // 文件按 contracts/i18n 解析；有问题时记 WARN 并忽略该插件的说明
3. sdk
   func Description(key string) RouteOption // 命令说明在本插件 messages 中的键
   type CommandInfo struct { PluginID, RouteID string; Path []string; Aliases [][]string; Usage string; Description string }
   func ListCommands(ctx context.Context) ([]CommandInfo, error) // 必须在 Handler 内调用（需要父调用链）
4. plugins/help
   func New() sdk.Root // 根插件

行为：
- core.command.list（必须带 parent_call_id；Actor 与默认 Target 取自父调用的可信上下文，没有 Actor 时只列出 ANY 权限的命令）：遍历 Router 的 CommandIndex，只保留每个规范路径的胜出者（被遮蔽者不列出）；对每条命令执行 policy.CheckCommand（全局禁止、启用、权限），通过的才列出；按路径字典序返回。usage 与 goal-12 的 usage 生成规则相同（按 Actor locale 使用 Core 文案目录）；description 在所属插件的 messages 中按 [Actor.locale, 默认 locale, en-US] 翻译，键不存在时 absent。
- 插件的 messages：SDK 不在 REGISTER 中上传。commandlist.Service 提供 SetMessages(pluginID string, files map[string][]byte)，由装配方调用：已安装的包读取包目录下的 i18n/*.json；库模式的开发组件使用生成器嵌入的 Generated.I18n；alonabot run 的开发入口读取 --entry 所在目录的 i18n/（存在时）。取不到时 description absent。
- help 插件：
  - /help：调用 ListCommands，按插件分组列出路径与说明，文案经 Context.T（键 help.list.header 等）。
  - /help <路径>：只显示匹配命令的 usage、别名与说明；没有可用的匹配命令时回复 help.not_found。
  - /whoami：回复调用者的 adapter_instance_id 与 external_user_id（文案键 help.whoami，参数 instance 与 user）；没有 Actor 时回复 help.whoami.no_actor。
  - 所有回复都用 Context.Send 回复原 Target；用户可见文案只来自 i18n（en-US 与 zh-CN）。

测试（按编码规范「测试提交规则」）：
- 提交：internal/core/supervisor/help_test.go（//go:build unix）：构建 plugins/help 与两个测试插件（一个有 ADMIN 权限的命令，一个与另一插件同路径且 priority 较低的命令），用 memory Adapter 装配：
  - 普通用户的 /help 不包含 ADMIN 命令与被遮蔽的命令；Core Admin 能看到 ADMIN 命令。
  - zh-CN 用户看到中文说明，en-US 用户看到英文；说明键缺失的命令没有说明。
  - 禁用的命令不出现；/help 某命令显示 usage。
  - /whoami 回复正确的实例 ID 与用户 ID。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有父调用链时 core.command.list 返回 FAILED_PRECONDITION；没有 Actor 时只列 ANY；排序；help 的 i18n 文件在 plugin check 中通过。

完成条件：gofmt -l 无输出；go generate ./... 后无未提交的生成差异；go vet ./...；go test -race ./...；go build ./...。

不做：welcome 与 rss（T-14.2）；help 的安装包与预装（T-14.3、T-14.4）；管理页面中的命令列表（T-09.9）。
```

### goal-72 · T-14.2 welcome 与 rss 插件

状态：等待前置（T-05.2、T-07.2、T-08.3、T-08.5、T-08.8）。语句按 D-68～D-72 之前的契约编写，前置合并后由审查者按新契约改写其中的旧写法（TrustedRole 与 Permission 枚举、`RouteOption`、`ChatAdmin` Rule 等）。语句中 goal-16、goal-24、goal-30、goal-32、goal-53 提供的符号（ChatAdmin Rule、ScheduleRegistry.Every、Config 与 OnLoad、gormdb.Open、Context.T、sdk.Send）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-72（切片 T-14.2：welcome 与 rss 插件）。goal 目录 goal-72，分支 goal-72，worktree .worktrees/goal-72，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：预装插件 plugins/welcome/ 在 member.joined 时向该聊天发送欢迎语，按聊天开关与自定义文案；plugins/rss/ 提供 /rss add|list|remove，订阅经 sqlitedb.Open 保存在插件数据库（不链接 GORM，D-49），固定 interval 的定时任务抓取 RSS 2.0 与 Atom，按 guid 或链接去重，只推送新条目；抓取只允许 http/https，拒绝 loopback、私网、链路本地、组播与未指定地址（含每次重定向之后），响应上限 5 MiB、超时 30s。

依据：PLAN T-14 步骤 2；CONTRACTS §15.2、§8（后台 Send 需要许可）、§5.4；TEST-MATRIX V-67。

范围（只新增或修改这些）：
- 新增 plugins/welcome/、plugins/rss/（各含可导入的包、cmd/ 下的 main、plugin.yaml、i18n/en-US.json 与 zh-CN.json）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> sdk/...、-> contracts（已允许）。只用标准库解析 XML（encoding/xml）与 HTTP。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
- plugins/welcome：func New() sdk.Root
- plugins/rss：func New() sdk.Root；type Config struct { config.Base; IntervalMinutes int `json:"interval_minutes" form:"default=10;min=1;max=1440"` }

行为：
- welcome：
  - /welcome on|off|set <text:raw>：只有 chat_admin 或 Core Admin 能执行（Rule：Any(ChatAdmin(), CoreAdmin())）；设置按聊天（instance、chat、thread）保存在插件数据库。默认关闭。
  - member.joined 时，该聊天已开启则发送欢迎语：自定义文案或 messages 中的 welcome.default；{name} 替换为新成员的显示名（没有时为 external_user_id）。回复原 Target，不需要许可。
  - 只响应 member.joined。
- rss：
  - /rss add <url> 与 /rss remove <url>：只有 chat_admin 或 Core Admin（Any(ChatAdmin(), CoreAdmin())）；/rss list 不限。订阅（url、目标 Target、最后条目标识集合的摘要）保存在插件数据库。add 时先做一次抓取校验（地址规则同下），失败时回复原因（文案键）。
  - 定时任务：Every(interval_minutes 分钟) 检查全部订阅；每个订阅抓取后解析 RSS 2.0（channel/item 的 guid、link、title）与 Atom（entry 的 id、link[rel=alternate 或缺省]、title）；条目标识为 guid（或 id），没有时为链接；只推送此前未见过的条目（首次订阅时只记录现有条目，不推送），按发布顺序发送，经 sdk.Send 显式 To（需要 allowed_adapter_instances，缺少许可时记 WARN 并跳过）。
  - 抓取：只允许 http 与 https；自定义 http.Client 的 DialContext 在连接前解析地址并拒绝 IsLoopback、IsPrivate、IsLinkLocalUnicast、IsLinkLocalMulticast、IsMulticast、IsUnspecified 的地址；CheckRedirect 对每次重定向的 URL 同样检查 scheme，连接时再检查地址；响应体以 io.LimitReader 读取 5 MiB 加一字节，超过即失败；整个抓取超时 30s（命名常量）。失败记 WARN（订阅 URL 与原因），不推送、不重试到下一个周期之前。

测试（按编码规范「测试提交规则」）：
- 提交：
  - plugins/rss/rss_test.go（//go:build unix）：用 memory Adapter 与真实 Core 装配，httptest 提供订阅源。注意 httptest 监听 loopback，测试通过包内的测试钩子把「允许 loopback」打开（钩子只在测试中设置，生产默认关闭），并另有子测试在钩子关闭时验证 loopback 被拒：
    - 普通用户 /rss add 被拒；chat_admin 成功；list 显示订阅。
    - 首次抓取不推送；源中新增两个条目后下一次定时任务推送两条，顺序正确；再次运行不重复推送。
    - 没有发信许可时不推送并记 WARN；授予许可后推送。
    - 响应超过 5 MiB、超过 30s（注入更短的超时）、重定向到 127.0.0.1（钩子关闭）、file:// 与 ftp:// 都失败。
    - Atom 与 RSS 2.0 各一个源。
  - plugins/welcome/welcome_test.go（//go:build unix）：开关权限、自定义文案与 {name} 替换、关闭时不发送、只响应 member.joined。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：地址检查函数（IPv4/IPv6 私网、链路本地、映射地址 ::ffff:127.0.0.1）；条目标识的优先级；XML 中缺少字段的条目。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...；两个插件目录运行 alonabot plugin check 通过。

不做：发行归档与预装（T-14.3、T-14.4）；HTML 清洗与富文本推送；代理设置（沿用进程环境的 HTTP_PROXY）。
```
