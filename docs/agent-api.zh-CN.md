# Agent、服务与生命周期 API

## 可用状态

本页规定 SDK 与 Agent/Core 实现方共用的公开契约。四类 Builder、OnSleep、llm.Generate 已提供 SDK 实现。Agent 工具循环、钩子调度、注入、改写及 Goal 执行由 Agent 服务实现；服务发现、重入、角色授权与增量路由要求 Core 实现并协商对应能力。当前 Core 尚未提供这四项能力。

按可用性分两组交付：

| 交付项 | 当前版本完成程度 |
|---|---|
| 四类 matcher builder | 完成 SDK 实现及调用方调整 |
| `OnSleep` | 完成 SDK 实现，消费已有 `STOPPING(IDLE)` |
| `llm.Generate` | 完成命名及调用接口调整，复用现有生成服务 |
| `agent.Run`、AI 钩子、流式接口、Goal | 定义完整接口、数据契约及 SDK 调用／注册入口 |
| 服务发现、回调重入、流式传输 | 完成通用协议与能力声明；Core 运行时实现另列后续任务 |

依赖尚未实现的能力时，沿用 capability 协商返回明确的未支持错误；Agent 服务尚未注册时保留 `NotFound`。文档分别标注“契约已定义”和“运行时已支持”。

### Matcher builder

`Plugin`、`SubPlugin`、`CommandGroup` 统一提供以下返回类型：

| 入口 | 返回类型 |
|---|---|
| `Command(path ...string)` | `*CommandBuilder` |
| `Keyword`、`StartsWith`、`EndsWith`、`FullMatch` | `*TextBuilder` |
| `Regex(expr string)` | `*RegexBuilder` |
| `Raw()` | `*RawBuilder` |

能力表固定为：

| 方法 | Command | Text | Regex | Raw |
|---|:---:|:---:|:---:|:---:|
| `Permission`、`Priority`、`Rule`、`Continue`、`Handle` | ✓ | ✓ | ✓ | ✓ |
| `Menu`、`Alias`、`Args`、`Optional`、`RawTail` | ✓ | | | |
| `Description`、`UsageKey`、`Silent` | ✓ | | | |
| `ID` | | ✓ | ✓ | ✓ |
| `IgnoreCase` | | ✓ | | |

共享私有注册状态，各链式方法返回自身 builder 类型。正则大小写通过表达式 `(?i)` 控制。

`Menu()` 表示参与菜单发布；实际发布资格继续取决于生效主名是否为单 token，以及现有权限、Profile、前缀和平台条件，保留管理员改名后的行为。

### 休眠生命周期

```go
func (p *Plugin) OnSleep(fn func(context.Context) error)
```

- 每个插件最多注册一次。
- 收到 `STOPPING(reason=IDLE)` 时执行 `OnSleep`，随后在剩余预算内执行 `OnStop`。
- 两者共享原有停止截止，无新增超时。
- `OnSleep` 返回错误或 panic 时记录完整错误，继续剩余关闭流程。
- 唤醒后的初始化继续使用 `OnStart`、`OnReady`。

## Generate 与 Run

统一采用结构化请求，调用方显式选择完整或流式模式。

```text
llm.Generate(ctx, req Request) (Response, error)
llm.GenerateStream(ctx, req Request, onDelta func(Delta) error) (Response, error)

agent.Run(ctx, req RunRequest) (RunResult, error)
agent.RunStream(ctx, req RunRequest, onEvent func(RunEvent) error) (RunResult, error)
```

`llm.Generate` 完成一次文本生成；`agent.Run` 完成一次可包含多个模型请求和工具执行的 Agent Run。`agent.Run` 使用独立的 `agent.run` 服务契约。

### `llm` 类型

| 类型／字段 | 参数规格 |
|---|---|
| `Message.Role string` | `user` 或 `assistant` |
| `Message.Text string` | 完整文本 |
| `Request.Messages []Message` | 至少一条消息，按传入顺序使用 |
| `Request.System string` | 当前版本生成的基础系统提示词 |
| `Request.Model string` | 空值使用 Agent 配置的聊天模型 |
| `Request.MaxOutputTokens int` | `0` 使用模型配置值；负数为参数错误 |
| `Response.Text string` | 改写后的完整文本 |
| `Response.Model string` | 实际使用的模型配置名 |
| `Response.StopReason string` | 规范化停止原因；未知原因保留原值 |
| `Response.InputTokens *int` | 缺失表示 Provider 未提供；指针保留实际零值 |
| `Response.OutputTokens *int` | 同上 |
| `Delta.RequestID string` | Agent 分配的模型请求标识 |
| `Delta.Sequence uint64` | 当前请求内从 `0` 开始的交付序号 |
| `Delta.Text string` | 经过流式钩子处理的文本增量 |

流式成功时，`Response.Text` 等于所有成功交付增量的拼接。中途失败时返回已确认的部分结果和错误。

### `agent` 类型

`RunRequest`：

| 字段 | 参数规格 |
|---|---|
| `Input llm.Request` | 初始输入、系统提示词、模型和输出 token 配置 |
| `Tools *ToolSelection` | `nil` 使用当前调用可用的工具目录 |
| `Tools.Names []string` | 非 nil 的选择对象表示仅使用这些完整工具标识；空列表关闭工具 |

Run 的工作上下文保存在当前版本调用内。聊天历史的读取与持久化由 Agent 的对话入口负责；SDK 服务调用本身返回执行结果。

`RunResult`：

| 字段 | 参数规格 |
|---|---|
| `RunID string` | 当前版本执行标识 |
| `Text string` | 当前版本 Run 所有已交付模型文本的拼接 |
| `Steps []RunStep` | 按实际顺序保存模型响应、工具调用及结果 |
| `InputTokens *int` | 全部请求均提供对应用量时返回合计 |
| `OutputTokens *int` | 同上 |

`RunStep` 分为模型步骤和工具步骤，保留对应的 `RequestID`。工具步骤包含完整工具标识、调用 ID、参数、原始结果、提供给模型的结果及执行错误。

`RunEvent` 包含 `RunID`、`RequestID`、序号和事件类型，支持：

- `TextDelta`：文本增量。
- `ToolStarted`：经过钩子处理、即将执行的工具调用。
- `ToolFinished`：工具实际结果及错误。

工具参数以完整 JSON 对象交付。Provider 私有的 replay 数据由 Agent 保留并原样回传。

调用回调返回错误时取消当前版本生成／Run，错误随部分结果返回。SDK 的相关编解码路径须保留协议已有的 `payload + error` 语义。

通过 `tool.Confirm` 提交的操作缺省为 `ALLOW`，直接返回且不发送确认消息，无 Actor 的调用同样适用。管理员通过 `/api/v1/tools/policies` 设置的覆盖与配置文件策略继续生效；`CONFIRM` 要求可信 Actor 与目标。工具服务继续采用自身声明的缺省策略。

## AI 钩子、提示词与响应修改

公共注册入口放在 `sdk/agent/hook`，接受 `sdk.Registrar`，适用于插件与子插件。

注册形式：

```go
hook.BeforeRequest(p, "weather").
    Priority(10).
    Handle(func(c *sdk.Context, in hook.RequestInput) (hook.RequestPatch, error) {
        // 返回本插件的修改。
    })
```

所有注册函数要求插件内稳定 ID，提供 `Priority(int32)` 和类型化 `Handle`。内部登记为普通 Service，附带通用重入声明和当前 `@agent` 调用授权。

Agent 按 `(priority, plugin_id, route_id)` 顺序执行，优先级数值越小越先。

### 阶段接口

| 注册入口 | 输入 | 返回值与作用 |
|---|---|---|
| `BeforeRun` | `RunInput` | `RunPatch`：调整初始请求及注入 |
| `BeforeRequest` | `RequestInput` | `RequestPatch`：调整当前版本模型请求及注入 |
| `RewriteResponse` | `ResponseInput` | `ResponsePatch`：修改完整模式的响应文本 |
| `RewriteStream` | `StreamInput` | `StreamPatch`：修改、暂存及 flush 流式文本 |
| `BeforeToolCall` | `ToolCallInput` | `ToolCallPatch`：修改当前工具调用的参数 |
| `AfterToolCall` | `ToolResultInput` | `ToolResultPatch`：修改提供给模型的工具结果文本 |
| `OnResponse` | `ResponseOutcome` | `error`：观察最终响应、用量及错误 |
| `AfterRun` | `RunOutcome` | `error`：观察 Run 结果及错误 |

公共 `Meta` 包含：

- `RunID string`：独立 `llm.Generate` 调用时为空。
- `RequestID string`：进入模型请求阶段后设置。
- `Iteration int`：Run 中从 `1` 开始的模型请求轮次。
- `GoalID string`：由 Goal 发起时设置。

可信调用身份、权限、取消和截止通过 `*sdk.Context` 提供。

模型请求视图包含基础系统提示词、完整消息序列、模型、输出 token 配置及工具目录。消息能够表达 assistant 工具调用和对应工具结果。Provider 私有状态保持只读。

Patch 使用指针区分“保持原值”和“修改为空值”。响应改写仅修改文本；模型身份、实际 token 用量和执行结果保持真实记录。工具参数修改后按最终参数进行 Schema 校验及既有工具策略检查。

修改阶段错误终止当前操作；观察阶段错误完整记录，继续后续观察者。

### 提示词注入

```text
InjectionChange:
    Op    Upsert | Delete
    Key   string
    Scope Request | Run
    Text  string
```

`RunPatch`、`RequestPatch` 均可返回 `[]InjectionChange`。

规则固定为：

- 所有者由 Agent 根据钩子服务身份确定。
- 存储键为 `(plugin_id, scope, key)`。
- 同一键 `Upsert` 更新文本；`Delete` 显式删除。
- 同名 Request 注入在当前请求中覆盖本插件的 Run 注入；请求结束后恢复 Run 注入。
- Request 注入作用于对应的一次模型请求；`BeforeRun` 返回的 Request 注入作用于第一次请求。
- Run 注入持续到当前版本 Run 结束；独立 `llm.Generate` 接受 Request 注入。
- 注入排列按钩子顺序，同一路由内按 key 排序。
- 注入置于基础系统提示词之后，统一完成拼装。
- 钩子分别读取基础请求和有效注入列表，后续请求重新拼装。
- 注入内容保持运行态；长期历史由业务明确写入。

因此，A 的位置与 B 的温度可同时保留至 Run 结束，各自更新也互不覆盖。

### 流式修改

`StreamInput`：

| 字段 | 规格 |
|---|---|
| `Meta` | 当前 Run／模型请求信息 |
| `Phase` | `Start`、`Delta`、`End`、`Abort` |
| `Text string` | `Delta` 时的输入文本 |
| `State json.RawMessage` | 当前钩子上一次返回的私有流状态 |
| `Failure` | `Abort` 时的错误信息 |

`StreamPatch`：

| 字段 | 规格 |
|---|---|
| `Texts []string` | 输出零段、一段或多段文本 |
| `State json.RawMessage` | 后续事件继续使用的状态，返回 nil 清空 |

Agent 按“钩子服务＋RequestID”隔离状态。

- `Start` 初始化状态。
- `Delta` 可暂存文本，例如等待被拆开的关键词。
- `End` 输出剩余缓冲；上游 flush 的文本先经过下游 Delta 处理，再执行下游 End。
- `Abort` 用于清理；已断链时由 Agent 回收本地状态。
- 流式修改发生在交付前，已交付文本保持确定。
- `RewriteResponse` 用于完整模式；流式结束后通过 `OnResponse` 观察改写后的汇总结果。
- 后续模型上下文、Run 结果和聊天历史使用改写后的内容。

## Core 通用接口与传输规范

当前版本只新增协议、SDK 接口及规范。以下运行时处理由后续 Core 通信任务实现。

### 服务发现

```text
service.Discover(ctx, Filter) ([]Descriptor, error)
```

`Filter` 按输入／输出 Schema ID 及版本筛选。

`Descriptor` 包含完整服务目标、插件 ID、route ID、priority、编解码方式、输入／输出 Schema 和流式／重入声明。发现结果限定为当前调用身份可访问的有效服务，并覆盖可被按需唤醒的注册快照。

AI 阶段通过共享 Schema 识别，阶段定义保存在 Agent 契约中。

### 受控回调重入

```text
service.Builder.Reentrant()
service.Builder.Priority(int32)
```

- 服务显式声明可重入。
- 重入调用继续检查直接调用方、权限、调用深度、期限和有效 Session。
- 普通服务保留原有调用环检查。
- allowlist 支持已有角色别名 `@agent`，由 Core 按当前安装身份解析。

由此支持 `A → Agent.Run → A 的钩子／工具服务`，授权范围由 A 的服务声明决定。

### 通用流式服务

```text
service.StreamJSON[In, Delta, Out](target)
service.ServeStream(owner, declaration)
StreamSpec.Call(ctx, in, onDelta) (Out, error)
```

服务端 Handler：

```text
func(*sdk.Context, In, emit func(Delta) error) (Out, error)
```

传输规范采用已有双向 REQUEST／RESPONSE：

- 初始请求保持在途，终态 RESPONSE 返回最终结果。
- 提供方通过与原调用绑定的增量交付接口发送数据。
- Core 根据认证 Session 和在途父调用定位接收方。
- 接收方处理完成后确认本段，`emit` 随后返回，形成背压。
- 增量按调用内顺序交付，使用独立序号。
- 取消、超时、断链沿用原调用链。
- 最终结果和错误允许同时存在。
- 沿用既有帧大小、在途预算及期限规则。

分别定义服务发现、角色别名授权、重入和流式传输的 capability。运行时实现完成后才声明支持。

## Goal 管理与恢复

Goal 公共入口放在 `sdk/agent/goal`：

```text
goal.Create(ctx, CreateRequest) (Goal, error)
goal.Get(ctx, id string) (Goal, error)
goal.List(ctx) ([]Goal, error)
goal.Pause(ctx, id string) (Goal, error)
goal.Resume(ctx, id string) (Goal, error)
goal.Cancel(ctx, id string) (Goal, error)
```

### 创建参数与状态

| `CreateRequest` 字段 | 规格 |
|---|---|
| `Condition string` | 必需，描述可验证的完成条件 |
| `Run agent.RunRequest` | 初始执行输入及工具选择 |
| `EvaluatorModel string` | 空值使用当前版本 Run 选定的模型配置，独立发起评估请求 |
| `Checker string` | 可选的确定性验收服务完整目标 |
| `RecoveryPolicy` | 默认 `Manual`；契约预留 `Automatic` |

`Goal` 返回 ID、创建者插件、条件、恢复策略、状态、已执行 Run 数、最新评估、暂停／失败原因，以及创建和更新时间。时间统一使用 `time.Time`，JSON 编码为 RFC3339。

状态为 `Pending`、`Running`、`Paused`、`Completed`、`Failed`、`Cancelled`。

- 默认每次 Run 结束后，由独立模型评估条件和已记录的执行证据。
- 评估结果为 `NotMet`、`Met`、`Impossible`，均包含原因；`NotMet` 的原因作为下一次 Run 的指导。
- 指定 Checker 时，由其结果决定验收。
- Checker 注册函数接受 `sdk.Registrar`，输入包含 Goal 条件及已记录的执行步骤，返回同一评估结构。
- 运行或评估发生操作错误时暂停并保存原因，等待恢复。
- `Pause` 协作取消当前 Run并保存已确认步骤；`Cancel` 进入终态。
- 管理服务按真实创建者插件归属鉴权；Agent 的交互管理入口使用可信用户权限。

目标、进度、执行记录和评估记录归 Agent 数据库；调度复用现有动态任务。

默认恢复契约为 `Manual`：重启后，未完成的执行先暂停。`Automatic` 留给 Agent 后续实现，只恢复重启前仍在运行的目标。用户主动暂停的目标等待显式 `Resume`。

恢复从记录继续安排新的 Run；结果未知的副作用工具调用交由后续步骤核实。每个 Run 重新获取位置、温度等注入信息。


## SDK 路径、Schema 与协议目标

| SDK 包 | 入口／服务目标 |
|---|---|
| `sdk/llm` | `Generate` → `@agent/llm.generate`；`GenerateStream` → `@agent/llm.generate.stream` |
| `sdk/agent` | `Run` → `@agent/agent.run`；`RunStream` → `@agent/agent.run.stream` |
| `sdk/agent/hook` | 八个注册函数 → 本插件 `agent.hook.<阶段小写>.<id>` 服务，route ID 为 `svc.` 加服务名 |
| `sdk/agent/goal` | `Create/Get/List/Pause/Resume/Cancel` → `@agent/goal.create/get/list/pause/resume/cancel` |
| `sdk/agent/goal` | `Checker(owner,id)` → 本插件 `goal.checker.<id>` |
| `sdk/service` | `Discover` → `core.service.discover`；流式初始调用 → `core.service.stream.call`；提供方增量 → `core.service.stream.emit`；接收方回调 → `sdk.service.stream.delta` |

`service.JSON` 与 `StreamJSON` 的 target 使用 `plugin_id/service_name`；服务端以 `/service_name` 声明本插件自有服务，身份由 Core 的认证 Session 决定。稳定 ID 必须非空并使最终服务名满足既有命名规则。完整与流式模式分别登记目标；LLM 与 Run 的流式服务名追加 `.stream`。`ServeStream` 的 Handler 需要流式调用上下文。

Service 调用通过 `InvocationContext` 继承父事件的 `actor`、`default_target` 和 `chat_type`（编号为 8）。Service Handler 通过 `c.User()`、`c.Session()` 和 `c.ChatType()` 读取这些字段，`c.Event()` 为空。没有真实入站事件时，`c.ChatType()` 返回 `message.ChatUnspecified`。

JSON Schema ID 为包导入路径加 `.` 加类型名，版本固定为 1。例如 `github.com/AlonaBot/AlonaBot/sdk/agent.RunRequest`、`sdk/agent/hook.RequestInput`（后者也使用完整模块前缀）。生成器识别 `JSON`、`StreamJSON`、公开 Agent 封装、`json.RawMessage` 与 `time.Time`。RawMessage 表示原始 JSON 值，工具 Arguments 进一步要求对象；time.Time 编码为 RFC3339 字符串。其余自定义编解码仍遵循已有生成规则。

能力名 `service_discover`、`service_role_access`、`service_reentrant`、`service_stream` 的版本均为 1，使用 HELLO/HELLO_ACK 的 supported/required/intersection 精确协商。SDK 消费协商结果，Core 当前未广告支持。流式调用和发现缺能力返回 `Unsupported`；声明角色授权、重入或流式服务的插件在激活前同样明确返回 `Unsupported`。普通 `agent.run` 与 Goal 调用沿用真实服务查找的错误，服务缺失为 `NotFound`；当前 Agent 角色不可用时保留原有 `Unavailable`。完整 `llm.Generate` 复用已存在的生成服务。

发现 Filter 的空 Schema ID 或零版本表示该字段不限，Descriptor 返回 Target、PluginID、RouteID、Priority、Encoding、InputSchemaID/Version、OutputSchemaID/Version、DeltaSchemaID/Version 与 Reentrant。delta Schema 为空表示普通服务。通用协议源为 [service.proto](../contracts/proto/alonabot/v1/service.proto)、[route.proto](../contracts/proto/alonabot/v1/route.proto) 与 [invocation.proto](../contracts/proto/alonabot/v1/invocation.proto)。ServiceSpec 新增 reentrant=7、delta_schema_id=8、delta_schema_version=9；排序使用已有的 RouteDescriptor.priority。ServiceInvocation.stream_id=7 由 Core 绑定真实在途调用。StreamCall 的 target/stream_id/encoding/schema_id/schema_version/payload 为 1～6；StreamDelta 的 stream_id/sequence/encoding/schema_id/schema_version/payload 为 1～6。父调用、Session、权限、期限、序号必须在 Core 边界复核；stream_id 是关联值，认证仍由 Session 与在途父调用提供。

`RunStep.Kind` 为 `model` 或 `tool`，恰好设置 `Model *llm.Response` 或 `Tool *agent.ToolResult`。ToolCall 的 Name 是完整目标，CallID 是 Provider 工具调用标识，Arguments 为完整 JSON 对象。ToolResult 的 Call 保留实际执行参数，Raw 保留原始 JSON 结果，Text 为模型可见文本，Failure 可选。Failure.Code 保存错误码，Message 保存完整原因。RunEvent.Kind 为 `text_delta/tool_started/tool_finished`，分别使用 Text、Call、Result；Sequence 在整个 Run 内从零递增，RequestID 关联模型请求。RunStep 的顺序是实际执行顺序。

钩子输入统一携带 Meta。RunInput 的 Request 为初始 RunRequest；RequestInput 的 Request 为 RequestView；两者 Injections 都为当前有效列表。RequestView 包含 System、Messages、Model、MaxOutputTokens、Tools。Message 的 Role 可为 user/assistant/tool，ToolCalls 保存 assistant 调用，ToolResult 保存对应结果，Replay 为 Provider 私有只读 JSON。Tool 的 Name 为完整标识，Description 为工具说明，InputSchema 为 JSON Schema。Injection 包含可信 PluginID/RouteID、Key、Scope、Text。RequestPatch 的 System、Messages、Model、MaxOutputTokens、Tools 为指针；RunPatch.Request 也为指针；nil 保持值，指向空值允许显式清空。Replay 保持原样。输出 token 规则沿用 llm.Request。

ResponseInput/ResponseOutcome.Response 为实际模型响应，后者另含 Failure。RunOutcome.Result 为已确认 RunResult，另含 Failure。ToolCallInput.Call、ToolResultInput.Result 保留实际调用和结果。ToolCallPatch.Arguments 的 nil 保持参数，非 nil 必须为有效 JSON 对象；ToolResultPatch.Text 和 ResponsePatch.Text 的 nil 保持文本，指向空串显式清空。观察者的公共服务输出为 `Observed {}`，SDK Handle 仍只返回 error。

Goal 的 IDRequest.ID 为目标 ID；ListRequest 为空对象，ListResponse.Goals 返回创建者全部目标。Goal.RunCount 为 uint64，LatestEvaluation 可缺失；Reason 保存暂停或失败原因。CheckerInput 包含 GoalID、Condition、Runs（按顺序的已保存 RunResult）；返回 Evaluation{Verdict,Reason}。CreateRequest.RecoveryPolicy 空值采用 Manual；Automatic 仅为已定义策略，Agent 尚未实现时应明确拒绝。所有管理接口都须按可信创建者鉴权，未知 ID 返回 NotFound。操作错误进入 Paused，Impossible 进入 Failed，Met 进入 Completed，Cancel 进入 Cancelled。

## 注入、跨块改写和手动恢复示例

插件 A 在 BeforeRun 中返回 `{Op: Upsert, Key: "location", Scope: Run, Text: "location: Singapore"}`；插件 B 返回 `{Op: Upsert, Key: "temperature", Scope: Run, Text: "temperature: 28 C"}`。两个插件各自保留所有权，Agent 按钩子顺序将两段内容放在基础系统提示词后。A 再次 Upsert 同键会更新位置；B 返回 `{Op: Delete, Key: "temperature", Scope: Run}` 明确删除温度。A 在 BeforeRequest 中为 location 返回 Request 注入只覆盖本请求，请求结束恢复 A 的 Run 位置。BeforeRun 返回的 Request 注入只用于第一次请求。

跨 chunk 改写可将 `"sec"` 暂存在 StreamPatch.State，在下一块 `"ret"` 到来时输出 `"[redacted]"`。每次 Delta 合并私有缓冲和当前文本，输出已可确定的前缀，把剩余内容编码回 State。End 必须输出余下缓冲并将 State 设为 nil。上游 End 输出先经过下游 Delta，再对下游执行 End；Abort 清理状态且无需补发文本。每个钩子服务与 RequestID 分别保存 State。

```go
g, err := goal.Create(c, goal.CreateRequest{
    Condition: "The requested report is saved and the checker accepts it",
    Run: agent.RunRequest{Input: llm.Request{
        Messages: []llm.Message{{Role: "user", Text: "Prepare the report"}},
    }},
    RecoveryPolicy: goal.Manual,
})
if err != nil { return err }
g, err = goal.Get(c, g.ID) // 重启后未完成目标处于 Paused。
if err != nil { return err }
_, err = goal.Resume(c, g.ID) // 从已保存记录安排新 Run，重新取得注入。
return err
```

调用取消使最终响应来不及时，GenerateStream 保留成功回调的文本，RunStream 保留成功事件的 RunID、模型文本及 ToolFinished 记录；尚未取得的模型身份和用量留空。成功终态的汇总文本须与已交付文本一致。

## 按目标查询与预检出站消息

交互回合使用当前 Handler 查询目标，随后沿同一 Context 发送；定时任务使用已保存的会话和实例，通过根级函数查询及发送。所有查询都执行当前的会话与发信准入；Context 方法保留可信父调用、Actor 和原目标回复许可。

```go
description, err := c.DescribeMessage(c.Session(), sdk.Via(c.Instance()))
if err != nil {
    return err
}
// description.Formats、Extensions 和 Combinations 供本轮工具描述与格式查询使用。
status, err := c.CheckMessage(c.Session(), content, sdk.Via(c.Instance()))
if err != nil {
    return err
}
if status == message.SupportUnsupported {
    return errcode.Unsupported.New("message is unsupported")
}
_, err = c.Send(content)
return err
```

根级签名为 `sdk.DescribeMessage(ctx, to, opts...) (*message.OutboundDescription, error)` 和 `sdk.CheckMessage(ctx, to, content, opts...) (message.SupportStatus, error)`，`opts` 使用现有 `sdk.Via(instanceID)`。`to` 使用会话 SID；实例单独由 Via 指定。根级入口沿用 `sdk.Send` 的独立调用许可，交互回复使用 Context 方法。

`message.StandardFormats()` 定义 text、markdown、image、video、voice、file、mention、quote、forward、buttons 的字段 Schema 和完整 `send_message` 例子。详情的标准 part 字段为 `text/entities`、`content`、`source/caption`（file 另带 `filename/mime`）、`user_id/display`、`message_ref`、`nodes`、`rows`；完整 Schema 与递归节点以 SDK 返回为准。实体的偏移和长度以 UTF-16 码元计，实体类别沿用 SDK 的数值枚举。forward 节点选择自编 `parts` 或同目标、同实例的确认 `message_ref`；选项按钮只公开 `label`。

描述 JSON 使用 snake_case。标准格式版本为 1；未声明的格式状态为 `unknown`，已声明的 `supported`/`unsupported` 保留。即使旧 Adapter 缺少格式声明，实例原有 `message.send.TextLimit` 仍返回到 text/markdown，状态保持 unknown。`Sources` 使用 `https_url`、`media_ref`、`incoming`、`inline_base64`、`platform_file`；Agent 将 incoming 显示为 attachment，并按管理员策略提供 local_path。`Sources` 缺省表示来源未知；`TextLimit`、`ButtonLabelLimit` 缺省表示未声明，存在时保留 SDK 的计量单位。

扩展目录保留声明的完整名称、版本、用途、载荷 Schema、载荷例子和 AdapterID；Agent 将载荷例子包装为 `{"parts":[{"type":"<name>","version":1,"payload":{...}}]}`。组合规则包含名称、涉及格式、说明和完整发送例子，平台的实际判定由 Adapter 的可选接口执行。

```go
// sdk/adapter
func Extension[T any](name string, version uint32) ExtensionDeclaration
func (d ExtensionDeclaration) Describe(description string, example any) ExtensionDeclaration

type MessageChecker interface {
    CheckMessage(ctx context.Context, to string, chain message.Chain) (message.SupportStatus, *errcode.Error)
}
```

Adapter 在 `Descriptor.Formats` 和 `Descriptor.Combinations` 声明当前规则；扩展例子在 REGISTER 边界按同一载荷 Schema 校验。旧 `Extension[T](name, version)` 可继续注册。Core 按当前实例所属 Session 与代次读取描述，只为实现 MessageChecker 的实例公布 `sdk/message.check`。

预检先执行与发送一致的目标准入、基础结构与扩展 Schema 检查，再从媒体 Store 查询授权元数据。插件提交的 `MediaSource.MIME` 仍被拒绝；Store 核对 Session、期限与完整目标 Grant 后，读取已上传或已登记的 MIME，保留原 MediaRef/Incoming 和 token。预检不打开媒体文件、取件、执行出站 Hook、签发按钮、创建租约或发送平台请求。Checker 接收尚未签发 CoreToken 的按钮。缺少决定组合的元数据或缺少可选接口时返回 unknown；确定违反规则时返回 unsupported 及具体错误，权限、取消、目录和 RPC 错误完整返回。最终 Send 继续执行其完整校验与副作用链。

需要将物理回执关联回原始 parts 时，交互发送调用 `c.Send(content, sdk.RequireOriginalSegmentIndexes())` 或 `c.SendTo(to, content, sdk.Via(instance), sdk.RequireOriginalSegmentIndexes())`，定时发送调用 `sdk.Send(ctx, to, content, sdk.Via(instance), sdk.RequireOriginalSegmentIndexes())`。选项只适用于发送；DescribeMessage、CheckMessage、ReadNodes 和平台动作只接受 Via，传入该选项返回 INVALID_ARGUMENT。

启用 `RequireOriginalSegmentIndexes` 后，Core 为交互与无父调用的根级发送都跟踪原始片段下标。Core 自动添加的引用、提及和独立前缀没有调用方来源；修改已有文本保留其来源，按阈值封装的 forward 关联全部被封装的原始片段。每项物理结果、状态和错误完整保留。AfterSend、调试观察与回复修饰记账继续接收实际发送链及其内部平台下标。

Core 对每个 BeforeSend 回传立即比较消息链。等值回传与无 Hook 路径保留原下标映射；首次非等值改链（包括仅改 caption）返回 FAILED_PRECONDITION：`rich receipt attribution unavailable after BeforeSend rewrite`，在按钮签发、媒体租约和平台投递前终止，并返回 REJECTED 回执。多个 Hook 中首个改链即拒绝，后续恢复原内容的 Hook 不再执行。普通 Send、SendTo、SendText 和 Stream 保持现有调用前提和 Hook 改链行为；需要原始 parts 回执的富消息发送须显式传入此选项。

读取需要用户确认的本地文件前，调用 `tool.ConfirmRequired(ctx, operation, summary, effect)`，完整展示路径与目标。它在策略 allow 时仍运行现有确认流程，deny 优先拒绝；缺少可信交互 Actor 时失败。摘要沿用 1024 字节上限，超长摘要返回错误。`tool.Confirm` 保持按现有策略决定是否确认。
