# 实施计划：Agent

总目标、章节划分与跨章节依赖见 [PLAN](../PLAN.md)；本章节的 goal 语句见 [agent.md](../../docs/goals/agent.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本章节覆盖官方 Agent 组件 `agent/`（含 Provider 客户端、对话、工具循环与 MCP、记忆、技能、定时任务与打包），即工作包 T-12。Agent 依赖的 Core 服务（工具目录、调用确认、会话存储，T-11）与 `sdk/llm` 的调用契约在[核心实施计划](core.md)。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-12.1 | goal-58 | Provider 客户端：OpenAI、OpenAI 兼容与 Anthropic Messages 的流式请求、SSE 解析、工具调用解析、停止原因与用量、本地模拟服务 | T-12 步骤 1 | T-02 | V-57 的 Provider 分项 | 已完成（PR #16；TEST-MATRIX §3.2） |
| T-12.2 | goal-59 | 对话：`agent/` 骨架与配置、人格、历史、延续调用内的一轮对话、流式编辑显示、`/persona` 与 `/reset` | T-12 步骤 2、3 | T-03.9、T-04.3、T-05.4、T-08.4、T-08.11、T-11.3、T-12.1、T-12.9、T-12.10、T-12.11、T-09.4 | V-57 | 已完成（PR #170；TEST-MATRIX §3.2） |
| T-12.3 | goal-60 | 工具循环与 MCP：Core 工具、MCP 客户端（stdio、streamable HTTP）、确认、单轮步数上限 | T-12 步骤 4 | T-11.2、T-12.2 | V-58 | 已完成（PR #178；TEST-MATRIX §3.2） |
| T-12.4 | goal-61 | 记忆：Agent 内的会话笔记与用户档案、`memory` 工具、会话快照与档案注入 | T-12 步骤 5 | T-08.5、T-11.2、T-12.2 | V-59 | 已完成（PR #182；TEST-MATRIX §3.2） |
| T-12.5 | goal-62 | 技能：技能目录、`SKILL.md` 解析、`load_skill` | T-12 步骤 6 | T-12.3 | V-60 | 已完成（PR #188；TEST-MATRIX §3.2） |
| T-12.6 | goal-63 | 定时 Agent 任务：任务工具、用动态任务（D-124）保存与触发、后台运行 | T-12 步骤 7 | T-07.9、T-08.5、T-12.3 | V-61 | 未开始 |
| T-12.7 | goal-64 | 官方 Agent 组件的打包与多进程验收，汇总证据 | T-12 步骤 8 | T-12.4、T-12.5、T-12.6、T-12.8、T-12.14 | V-57～V-61、V-76、V-118 | 未开始 |
| T-12.8 | goal-83 | 文本生成服务：Core Broker 的 `@agent` 目标解析，`sdk/llm` 的类型、Schema 与调用辅助（含一问一答的 `llm.Ask`，D-111），Agent 的 `llm.generate` 与调用方许可，选项来源 `llm.providers`（D-131） | T-12 步骤 9 | T-03.8、T-06.1、T-08.14、T-12.2 | V-76、V-82 的 `@agent` 分项，V-109，V-116 的 `llm.providers` 分项 | 已完成（PR #175；TEST-MATRIX §3.2） |
| T-12.9 | goal-118 | Provider 增加 OpenAI Responses（`openai-responses`）：请求编码、SSE 事件解析、reasoning item 的回传、停止原因与用量；`ToolCall.Replay` | T-12 步骤 10 | T-12.1 | V-57 的 Provider 分项 | 已完成（PR #126；TEST-MATRIX §3.2） |
| T-12.10 | goal-119 | Provider 增加 Gemini 原生协议（`gemini-generate-content`）：`streamGenerateContent` 的请求编码与 SSE 解析、thoughtSignature 的回传、工具结果按调用查名 | T-12 步骤 10 | T-12.1 | V-57 的 Provider 分项 | 已完成（PR #127；TEST-MATRIX §3.2） |
| T-12.11 | goal-120 | Provider 增加 Ollama 原生协议（`ollama-chat`）：`/api/chat` 的请求编码与 NDJSON 流解析、可选的 API key、工具结果按调用查名 | T-12 步骤 10 | T-12.1 | V-57 的 Provider 分项 | 已完成（PR #128；TEST-MATRIX §3.2） |
| T-12.12 | goal-134 | Provider 改用官方 SDK：`openai`、`openai-compatible` 与 `openai-responses` 经 openai-go，`anthropic-messages` 经 anthropic-sdk-go；不重试、不读环境变量、Chat Completions 的成功条件、Anthropic 的 BaseURL 改为服务根（D-132） | T-12 步骤 1、10 | T-12.9、T-12.10、T-12.11 | V-57 的 Provider 分项 | 已完成（PR #134；TEST-MATRIX §3.2） |
| T-12.13 | goal-135 | Provider 图片输入：`Image` 与 `Message.Images`、请求前的校验、六种协议的图片编码（D-132） | T-12 步骤 1、10 | T-12.12 | V-118 的 Provider 分项 | 已完成（PR #138；TEST-MATRIX §3.2） |
| T-12.14 | goal-136 | 对话中的图片输入：Provider 条目的 `image_input`、读取入站图片、Agent 内的图片存储、历史回放与占位、`/reset` 删除（D-132） | T-12 步骤 11 | T-04.8、T-12.2、T-12.13 | V-118 | 未开始 |

## 3. 工作包

### T-12：官方 Agent 组件

**目标：** 交付可安装、可被社区替换的官方 Agent 组件 `agent/`（`role: agent`）：对话循环、Provider、工具与 MCP、Hermes 式的记忆、技能、定时任务与 `llm.generate`（D-27、D-59～D-61）。它只经 SDK 与 Core 通信。

#### 前置输入

- T-05 的延续调用与多轮等待、T-03 的消息编辑、T-04 的命令选择、T-08 的配置准备、Secret 引用与 `gormdb`、T-07 的动态任务（D-124，T-07.9）、T-11 的工具与会话服务。
- 读取 CONTRACTS §2.4、§3、§4.3、§5.4、§9.1、§13 与 DESIGN §R、§T。Provider 与 MCP 使用本地模拟服务测试；真实 Provider 需要 API key，没有时记为未运行。
- Provider 的 `openai`、`openai-compatible` 与 `openai-responses` 使用已批准的 `github.com/openai/openai-go/v3`，`anthropic-messages` 使用 `github.com/anthropics/anthropic-sdk-go/v2`，`gemini-generate-content` 与 `ollama-chat` 用标准库（D-132），两个 SDK 的版本在 T-12.12 引入时锁定；MCP 客户端使用已批准的 `github.com/modelcontextprotocol/go-sdk`，版本在 T-12.3 实际验证后锁定（anthropic-sdk-go 已要求的版本是下限）。

#### 实施步骤

1. **Provider 客户端。** 在 `agent/provider/` 中实现 OpenAI（官方）与 OpenAI 兼容的 Chat Completions，以及 Anthropic Messages，前两者经 openai-go、后者经 anthropic-sdk-go 发出请求与读取流（D-132；T-12.1 的标准库实现由 T-12.12 替换）：流式请求、文本增量与工具调用组装、停止原因与用量、图片输入、错误转换；请求失败不自动重试。使用回调式流接口，避免 goroutine/通道错误被丢失，也让调用方背压自然传递：
   ```go
   type Protocol string
   const (
       OpenAI                Protocol = "openai"
       OpenAICompatible      Protocol = "openai-compatible"
       AnthropicMessages     Protocol = "anthropic-messages"
       OpenAIResponses       Protocol = "openai-responses"        // 步骤 10
       GeminiGenerateContent Protocol = "gemini-generate-content" // 步骤 10
       OllamaChat            Protocol = "ollama-chat"             // 步骤 10
   )
   type Config struct {
       Protocol Protocol
       BaseURL string // API base path，例如 OpenAI 的 https://api.openai.com/v1；Anthropic 为服务根地址 https://api.anthropic.com
       APIKey string
       Model string
   }
   type Message struct {
       Role string // system、user、assistant
       Content string
       ToolCalls []ToolCall
       ToolResults []ToolResult
       Images []Image // 只出现在 user 消息中（D-132）
   }
   type Image struct {
       MediaType string // image/jpeg、image/png、image/gif 或 image/webp
       Data []byte      // 图片的原始字节，编码时转为 base64
   }
   type Tool struct {
       Name string
       Description string
       InputSchema json.RawMessage
   }
   type Request struct {
       Messages []Message
       Tools []Tool
       MaxTokens int // 必须大于 0；由调用方显式提供，不在 Provider 内猜默认值
   }
   type ToolCall struct {
       ID string
       Name string
       Arguments json.RawMessage // 完整且合法的 JSON 对象
       Replay json.RawMessage // 步骤 10：同一协议要求随这次调用原样回传的不透明数据；调用方原样放回 assistant 消息，不解析
   }
   type ToolResult struct {
       ToolCallID string
       Content string
   }
   // Event 是流中的一个事件：ToolCall 非 nil 时是一次完整的工具调用，否则 Text 是文本增量。
   type Event struct {
       Text string
       ToolCall *ToolCall
   }
   // StopReason 统一两种协议的常见取值，其他取值原样保留。
   type StopReason string
   const (
       StopEnd       StopReason = "end"
       StopToolUse   StopReason = "tool_use"
       StopMaxTokens StopReason = "max_tokens"
   )
   type Usage struct {
       InputTokens  int // 含 Anthropic 的缓存写入与缓存读取
       OutputTokens int
   }
   type Result struct {
       StopReason StopReason // Provider 未给出时为空
       Usage      *Usage     // Provider 未返回用量时为 nil
   }
   type Client struct{ ... }
   func New(config Config, httpClient *http.Client) (*Client, error)
   func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
   ```
   `New` 校验协议、HTTP(S) base URL、model 和 API key，并要求注入非 nil 的 HTTP client，便于使用已有代理配置和本地测试服务。Request 的 MaxTokens 必须大于 0；Message.Role 只接受 system、user、assistant；ToolCalls 只能出现在 assistant 消息，ToolResults 只能出现在 user 消息，且每个 ToolResult.ToolCallID 必须非空。Images 只能出现在 user 消息，MediaType 只接受上述四种，Data 不能为空；客户端不检查图片的大小、尺寸与张数，超限时按 Provider 的错误返回。工具名必须非空且在同一 Request 中唯一，InputSchema 必须是 JSON 对象。OpenAI 把每个 ToolResult 转成对应的 tool 角色消息，同一条消息的文本随后作为 user 消息；Anthropic 把带工具调用、工具结果或图片的消息转成 content blocks，依次为 tool_result、image、text、tool_use，不丢弃与之同在的文本。图片按 Images 的顺序排在同一条消息的文本之前，只有图片没有文本的 user 消息同样发出：Chat Completions 中带图片的 user 消息 content 为 part 数组，每张图片一个 `image_url` part（`data:<MediaType>;base64,<Data>`），文本为最后一个 `text` part，不带图片的消息仍用字符串 content；Anthropic 的图片为 `source.type` 为 `base64` 的 `image` block。两种 OpenAI 协议的差别：`openai` 用 `max_completion_tokens` 并请求 `stream_options.include_usage`；`openai-compatible` 用 `max_tokens`，不发 `stream_options`（避免不认识它的兼容实现拒绝请求），兼容实现自行返回用量时照常读取。Anthropic 的 system 消息按输入顺序合并到顶层 system 字段，其他消息保留顺序。OpenAI 的 BaseURL 以 API 根路径结尾（含 `/v1`），追加 `/chat/completions`；Anthropic 的 BaseURL 是服务根地址，追加 `/v1/messages`，与 anthropic-sdk-go 及各兼容服务文档中的 base URL 写法一致。OpenAI 使用 Bearer 认证；Anthropic 使用 `x-api-key` 与 `anthropic-version` 请求头。
   SDK 的客户端由 `New` 构造一次，以注入的 HTTP client、Config 的 BaseURL 与 APIKey 构造，重试次数设为 0；不读取环境变量（两个 SDK 的 `NewClient` 缺省读取 `OPENAI_*`、`ANTHROPIC_*`，含凭证与自定义请求头）：anthropic-sdk-go 传 `option.WithoutEnvironmentDefaults()`，openai-go 没有公开的等价选项，直接构造 Chat Completions 与 Responses 的 Service，不经 `openai.NewClient`。请求体按本段与步骤 10 的字段规则用 SDK 的参数类型表达，InputSchema 原样发出，不丢失 `$defs` 等字段；SDK 没有对应字段时用它提供的额外字段机制。不导入两个 SDK 的 Azure、Bedrock、Vertex 等云厂商子包。
   `Stream` 逐事件读取流，不缓存完整响应；文本增量按到达顺序交给 emit。工具调用按协议的索引自行累积，OpenAI 分段中的 id 与 name 取非空值而不拼接（有的兼容实现在每个分段重复它们），只有收到完整结束标记且 arguments 是 JSON 对象后才发出带 ToolCall 的 Event。Anthropic 新增的事件类型与 block 类型按其版本策略忽略。只有收到协议的终止条件才算成功，成功时返回停止原因与用量：Anthropic 为 `message_stop`；两种 Chat Completions 协议为流无错误结束且至少一个 choice 给出过 `finish_reason`（openai-go 在 `[DONE]` 与提前 EOF 时都以无错误结束，无法区分，D-132），在 `finish_reason` 之后、用量之前断开时成功、Usage 为 nil。提前结束、无效事件、非成功 HTTP 状态、Provider 错误事件均返回错误，并附上 Provider 错误字段中的说明，其中回显的 API key 替换为 `[REDACTED]`；SDK 返回的错误用 `errors.As` 取出状态码与说明后照此格式报告，不沿用 SDK 错误的文本（anthropic-sdk-go 的含请求 URL）。emit 返回错误时立即停止读取并关闭响应体。错误文本不得包含 API key、认证头或带凭证的 URL；已交付的文本增量由调用方保留，不自动重发请求。
   为每种协议提供本地模拟服务，覆盖正常流、多个文本增量、完整工具调用、重复 id 的工具调用分段、未知 Anthropic 事件、停止原因与用量、非成功 HTTP 状态及其错误说明、中途断开（Chat Completions 分为 `finish_reason` 之前与之后）、畸形 SSE/JSON、取消 context、emit 返回错误、带图片的 user 消息与不合法的图片；模拟服务对非成功状态只收到一次请求，看不到 `OPENAI_*`、`ANTHROPIC_*` 环境变量带来的请求头。
2. **组件骨架与配置。** manifest 为 `role: agent`。配置包含：Provider 条目（名称、预设、协议、base_url、`secret` 标记的 api_key、model、缺省为 false 的 `image_input`，步骤 11）与默认 Provider，预设（openai、anthropic、gemini、deepseek、ollama 等）给出缺省协议与 base_url（anthropic 为服务根地址 `https://api.anthropic.com`，步骤 1；gemini 为 `gemini-generate-content`，ollama 为 `ollama-chat`，步骤 10），表单下拉框用 labels 显示预设名（D-76）；命名人格（系统提示）与默认人格；纳入上下文的历史条数；MCP 服务器（名称、传输方式；stdio 为 command、args 与 `secret` 标记的 env，streamable HTTP 为 url 与 `secret` 标记的 headers）；技能目录。Provider 列表可以为空，为空时 `/ask` 回复提示管理员配置。每个聊天当前的人格与 Provider 选择保存在 Agent 的数据库。
3. **一轮对话。** 注册 `/ask`、`/agent on|off`、`/persona`、`/provider`、`/reset` 命令与 message 事件路由。`/ask` 与 `/reset` 的缺省权限为 User，`/agent`、`/persona` 与 `/provider` 为 Admin，管理员可按 T-04 的命令策略调整。`/ask` 在任何聊天中触发一轮对话；Admin 用 `/agent on` 为某个聊天开启自由对话后，该聊天中 `to_me` 的非命令消息也触发（私聊的消息都是 `to_me`），默认关闭。群聊的会话条目标明发言人。触发的 Handler 请求延续调用，在其中：
   - 从 Core 会话存储读取最近的历史条目，条数取自配置；会话的第一轮读取该会话的笔记，作为快照条目写入会话；
   - 把当前 Actor 的档案附在本轮用户消息之前，组装请求并流式接收；本轮消息与历史中的图片按步骤 11 处理，会话条目只存文本与图片编号；
   - 流式显示交给 SDK 的 `c.Stream`（D-98、D-135，MESSAGE-MODEL §4A）：实例声明 `message.edit` 时先发送再按节流编辑，未声明时（例如 OneBot 11）生成结束后一次发送；单条上限由实例 `message.send` 能力的 `text_limit` 决定，超过时由 `c.Stream` 换条，Agent 不另设上限、不自行切分。回复以纯文本发送，不把模型输出的 Markdown 交给平台解析；
   - 本轮结束后追加会话条目。

   Provider 返回错误或中途断开时保留已显示的内容，按 Actor locale 回复错误说明，不自动重发请求。`/reset` 清除该会话，下一轮重新取记忆快照。
4. **工具循环与 MCP。** 每轮把 `core.tool.list` 返回的工具和已连接 MCP 服务器的工具交给模型；工具名按 Provider 的命名规则映射，映射在一轮内保持稳定；模型交出的 ToolCall 原样（含 `Replay`）放回下一次请求的 assistant 消息。Core 工具经 `core.tool.invoke` 调用；MCP 工具先经 `core.tool.confirm`（操作名为 `mcp/<server>/<tool>`）再调用。拒绝、超时与工具错误作为工具结果交回模型。单轮工具步数上限为命名常量（初值 20），达到时结束本轮并告知用户。MCP 服务器在首次使用时连接，随插件停止关闭；stdio 服务器的子进程属于插件的进程树。
5. **记忆。** 记忆是 Agent 的内部实现（D-60），用 `gormdb` 保存在 Agent 的受管 SQLite：会话笔记按默认 Target，用户档案按 Actor，字符上限为命名常量（初值 2,200 与 1,375，按 Unicode 码点计）：
   - 会话第一轮读取会话笔记作为快照（步骤 3），会话内的写入到下一个会话才进入系统提示；每轮附上当前 Actor 的档案；
   - Agent 注册 `memory` 工具（JSON Service，default_policy 为 allow），模型的调用经 `core.tool.invoke`；工具接受一组 add/patch/remove 操作，target 为 memory 或 user，patch 与 remove 用唯一的 old_text 定位；整组按结果检查上限后原子写入；范围取自 Core 下发的可信上下文；
   - 超限、完全重复的条目与含 Unicode Cf 类字符的内容返回错误，不丢弃已有条目。

6. **技能。** 启动时读取技能目录下各子目录的 `SKILL.md`：YAML frontmatter 中的 name 与 description 必填，其余内容为正文。系统提示列出各技能的名称与描述，`load_skill` 工具返回正文。不合法的文件记诊断并跳过，不阻止插件启动。
7. **定时 Agent 任务。** 提供 `schedule_task`、`list_tasks`、`cancel_task` 三个 Agent 内工具。只有 Core Admin 能创建与取消任务，其他用户只能列出本聊天的任务；创建与取消前先经 `core.tool.confirm`。任务用 SDK 的动态任务保存与触发（D-124，T-07.9）：Agent 声明一个任务 `sdk.Task[AgentTask]("agent")`，参数为创建者、Target 与提示词；`schedule_task` 按触发规则调用 `At`、`Every` 或 `Cron`，`list_tasks` 调用 `List` 后按聊天过滤，`cancel_task` 调用 `Cancel`。Agent 不自建任务表，不注册每分钟的固定任务，触发规则的校验、持久化与错过即跳过由 Core 负责（D-125）。任务的处理函数运行一轮对话，没有 Actor，只有 job scope：
   - 发送需要目标实例在 Agent 的 `allowed_adapter_instances` 中（缺省为全部实例，D-83）；
   - confirm 策略的工具一律拒绝；
   - 不注入记忆，会话键与聊天会话分开；
   - 一次性任务在派发时由 Core 删除，对话失败不重试，失败写日志。
8. **打包与验收。** Agent 组件按 T-09 的打包流程生成可安装包。用本地 Provider 与 MCP 模拟服务、回环测试 Adapter 运行 V-57～V-61 与 V-76 的全部场景；真实 Provider 与真实 Telegram 没有凭证时记为未运行。
9. **文本生成服务。** Core 的 Service Broker 按 CONTRACTS §3 把目标 `@agent` 解析为已安装的 Agent 组件（按 T-03.8 解析的 `role`），访问规则与调用链按解析后的 ID 执行，没有 Agent 组件时与目标未激活相同。`sdk/llm` 提供 CONTRACTS §13.5 `llm.generate` 的请求/响应类型、JSON Schema 与以 `@agent` 为目标的调用辅助，以及一问一答的 `llm.Ask(ctx, prompt)`（D-111）。Agent 注册该服务（public），配置增加 `llm_service_allowed_plugins`（缺省 `["*"]`，表示全部插件，D-83），按 Core 下发调用链中的上一跳插件检查许可；使用已有的 Provider 客户端完成一次非流式请求（流式接收后拼接），不调用工具、不读写会话与记忆、不套用人格。错误映射为 INVALID_ARGUMENT（Provider 不存在）与 UNAVAILABLE（Provider 出错或断开），Provider 客户端返回的错误说明只写日志，不进入 safe_message；不重试。Agent 另用 `sdk.ServeOptions` 注册选项来源 `llm.providers`（D-131，CONTRACTS §13.5）：每个 Provider 条目一个选项，存储值与显示名为名称，说明为协议与模型，按搜索词不区分大小写地匹配名称；不返回 base_url 与 API key；不检查 `llm_service_allowed_plugins`，因为它只由 Core 在管理 API 中调用。
10. **其他 Provider 协议（D-103）。** 在 `agent/provider/` 中增加三种原生协议，每种的编码与流解析放在自己的文件里，`Config`、`Request`、`Event`、`Result` 与 `Stream` 的语义不变：只收到协议的终止条件才算成功，请求失败不自动重试，错误不含 API key 与认证头、回显的 key 替换为 `[REDACTED]`，emit 返回错误时立即停止。`ToolCall.Replay` 是协议要求随工具调用原样回传的数据（不透明的 JSON），由流解析写入，调用方在后续请求中把整个 ToolCall 放回 assistant 消息；`openai`、`openai-compatible` 与 `anthropic-messages` 不产生也不读取它。`openai-responses` 经 openai-go 的 Responses Service 发出请求与读取事件，客户端的构造、重试与错误报告同步骤 1（D-132；T-12.9 的标准库实现由 T-12.12 替换）；`gemini-generate-content` 与 `ollama-chat` 用标准库。三种协议的图片同样按 Images 的顺序排在同一条消息的文本之前，规则同步骤 1。
   - **OpenAI Responses（`openai-responses`）。** BaseURL 追加 `/responses`，Bearer 认证。请求为无状态：`store: false`，不使用 `previous_response_id`，完整历史放进 `input`；system 消息按顺序合并到 `instructions`；user 与 assistant 的文本为 `{"role", "content"}` 输入项，带图片的 user 输入项的 content 为数组，每张图片一个 `{"type": "input_image", "image_url": data URL}`，文本为最后一个 `{"type": "input_text", "text"}`；工具结果为 `function_call_output`（`call_id`、`output`），排在同一条消息的文本之前；工具调用为 `function_call`（`call_id`、`name`、字符串形式的 `arguments`），排在文本之后，不带 `id`。工具为 `{"type": "function", "name", "description", "parameters", "strict": false}`（MCP 与 Core 工具的 Schema 不满足严格模式）；输出上限为 `max_output_tokens`；`include` 请求 `reasoning.encrypted_content`。流按事件 type 解析：`response.output_text.delta` 为文本增量；`response.output_item.done` 中的 `function_call` 项为一次完整的工具调用（`call_id` 为 ID）；同一响应中排在该调用之前、尚未附给其他调用的 `reasoning` 项去掉 `id` 后作为 JSON 数组写入该调用的 Replay，编码时原样放在对应 `function_call` 之前。`response.completed` 与 `response.incomplete` 是终止事件：前者有工具调用时为 `tool_use`、否则为 `end`，后者的 `incomplete_details.reason` 为 `max_output_tokens` 时为 `max_tokens`、其他原样保留；用量取 `usage.input_tokens` 与 `usage.output_tokens`。`response.failed` 与 `error` 事件返回错误并附说明；其他事件类型忽略。
   - **Gemini（`gemini-generate-content`）。** 请求 `POST <BaseURL>/models/<Model>:streamGenerateContent?alt=sse`（Model 不含 `models/` 前缀，按路径段转义），`x-goog-api-key` 认证。system 消息按顺序合并到 `systemInstruction`；user 为 `role: "user"`，assistant 为 `role: "model"`；工具结果为 `functionResponse` part（`id`、`name`、`response: {"output": Content}`），排在文本之前，name 取同一 Request 中此前 assistant 消息里 ID 相同的工具调用（找不到时 Stream 在发请求前返回错误）；图片为 `{"inlineData": {"mimeType", "data"}}` part（data 为 base64），排在 functionResponse 之后、文本之前；工具调用为 `functionCall` part（`id`、`name`、`args`），Replay 非空时其中的签名写回该 part 的 `thoughtSignature`。工具为一个 `functionDeclarations` 列表，Schema 原样放进 `parametersJsonSchema`；输出上限为 `generationConfig.maxOutputTokens`。流中每个 data 事件是一个 GenerateContentResponse，只读第一个候选：`thought` 为 true 的 part 忽略，text part 为文本增量，functionCall part 为一次完整的工具调用（Gemini 不给 `id` 时由客户端生成随机 ID；part 的 `thoughtSignature` 写入 Replay）。`finishReason` 为 `STOP` 时有工具调用为 `tool_use`、否则为 `end`，`MAX_TOKENS` 为 `max_tokens`，其他原样保留；用量取最后一次 `usageMetadata`，输入为 `promptTokenCount`，输出为 `candidatesTokenCount` 与 `thoughtsTokenCount` 之和。流在给出 `finishReason` 后结束才算成功；只有 `promptFeedback.blockReason` 而没有候选、或 data 中带 `error` 时返回错误。
   - **Ollama（`ollama-chat`）。** BaseURL 为服务根地址（例如 `http://localhost:11434`），追加 `/api/chat`；API key 可以为空（本地服务），非空时用 Bearer 认证（Ollama 云服务）。消息保留 system、user、assistant 角色与顺序；工具结果为 `role: "tool"` 的消息（`content`、`tool_name`，tool_name 按 ID 查找的规则同 Gemini），排在同一条消息的文本之前；图片为该 user 消息的 `images` 字段（base64 字符串数组，不带 data URL 前缀）；工具调用为 `tool_calls`（`function.name` 与对象形式的 `function.arguments`）。工具为 `{"type": "function", "function": {"name", "description", "parameters"}}`；输出上限为 `options.num_predict`。响应是逐行的 JSON（NDJSON），不是 SSE：`message.content` 为文本增量，`message.thinking` 忽略，`message.tool_calls` 中每项为一次完整的工具调用（没有 ID，由客户端生成）；`done` 为 true 的行是终止行，`done_reason` 为 `stop` 时有工具调用为 `tool_use`、否则为 `end`，`length` 为 `max_tokens`，其他原样保留；用量取 `prompt_eval_count` 与 `eval_count`（只有后者时输入记 0）。带 `error` 字段的行返回错误；流在终止行之前结束返回错误。

   Gemini 与 Ollama 生成的工具调用 ID 用 `crypto/rand` 产生，同一 Client 内不重复；它只用于把工具结果对应回调用，Gemini 在请求中原样回传，Ollama 不发送。三种协议各自提供本地模拟服务，覆盖的场景与步骤 1 相同，另加：Responses 的 reasoning 项随工具调用回传、`response.incomplete`；Gemini 的 thoughtSignature 回传、工具结果的 name 查找失败、只有 blockReason；Ollama 的空 API key 不发认证头、缺少终止行。
11. **对话中的图片输入（D-132）。** Provider 条目的 `image_input`（缺省 false）表示该模型接受图片。历史中的图片仿照 Hermes 的回放：图片字节由 Agent 保存，Core 会话条目只存文本与图片编号。
   - 读取：本轮消息的 image 片段可读且 MIME 为步骤 1 的四种之一时，用 `sdk.OpenMedia`（D-121）读入字节；不可读、类型不支持或读取失败的图片按占位处理并写 warn 日志，这一轮照常进行。
   - 编号与保存：每个会话的图片按到达顺序从 1 编号。本轮成功结束、追加会话条目时，把本轮读到的图片（MIME 与字节）按会话键与编号写进 Agent 的受管 SQLite，user 条目记录这些编号；本轮失败时不保存图片，也不占用编号。
   - 组装：`image_input` 为 true 时，本轮读到的图片全部放进本轮 user 消息的 Images；历史条目中编号大于回放边界、且仍保存着的图片放回所在条目的 Images。设本轮之前的最大编号为 m、批大小为 N（命名常量，初值 3），回放边界为 ⌊(m−1)/N⌋·N（m 为 0 时没有可回放的图片），因此回放 1～N 张，边界每 N 张图片才前移一次，请求前缀随之每 N 张才变化，Provider 的前缀缓存得以保持。其余图片在所在条目文本之前各留一个占位 `[image]`。`image_input` 为 false 时所有图片都只留占位，本轮的图片照常保存，切换到开启的 Provider 后可以回放。
   - 清理：本轮保存后，按新的最大编号计算下一轮的回放边界，删除编号不大于它的图片，因此每个会话至多保存 N 张。`/reset` 删除该会话的全部图片并从 1 重新编号。
   - 工具结果中的图片、`llm.generate` 的图片输入与图片的压缩缩放不在本步骤内。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `agent/` | 官方 Agent 组件：Provider 客户端与模拟服务、对话、工具循环、MCP、记忆、技能与定时任务 |
| `sdk/llm/` | 文本生成服务的类型、Schema 与调用辅助 |
| `internal/core/` 的 Service Broker | `@agent` 目标解析 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-57 | 六种协议的流式文本、工具调用、停止原因与用量（各协议的 Provider 分项）、编辑显示、人格切换、`/reset`、错误与断开 |
| V-58 | Core 工具与 MCP 工具的策略与确认、错误交回模型、单轮步数上限、MCP 进程随插件停止 |
| V-59 | 记忆的增改删、上限与非法内容、快照在会话内不变、群聊多用户、`memory` 工具的策略 |
| V-60 | 合法与不合法的 `SKILL.md`、`load_skill` |
| V-61 | 定时任务的触发、无发信许可、confirm 工具被拒、离线不补跑 |
| V-76 | `llm.generate` 的调用方许可、Provider 选择与错误映射、调用方期限、Agent 未启用；`@agent` 解析到替换后的 Agent |
| V-109 | `llm.Ask` 发出的请求与返回的文本，空白 prompt 不发出请求 |
| V-116 | `llm.providers` 列出已配置的 Provider、按存储值取回、不含 base_url 与 API key；其余分项由 T-08.14、T-09.14 与 T-09.9 提供 |
| V-118 | 六种协议的图片编码与校验；`image_input` 开启与关闭；历史图片的回放边界、占位、删除与 `/reset`；不可读的图片 |

Agent 组件的真实 Telegram 与真实 Provider 结果由 T-10 汇总。

#### 失败停止条件

- 未经策略与确认执行工具或 MCP 操作、后台任务通过确认、Provider 失败后自动重发请求（包括 SDK 的缺省重试），阻止合入。
- Provider 客户端从环境变量取得地址、凭证或请求头，或导入两个 SDK 的云厂商子包，阻止 T-12.12 合入。
- 图片字节写进 Core 会话存储，或一个会话保存的图片超过回放窗口仍不删除，阻止 T-12.14 合入。
- 某个协议丢弃了它要求回传的数据（Responses 的 reasoning 项、Gemini 的 thoughtSignature），使带工具调用的后续请求被真实 Provider 拒绝，阻止该协议合入。
- 记忆超限时丢弃已有条目、一个用户读到另一个用户的档案，阻止记忆合入。
- Agent import `internal/core` 或绕过 SDK 访问 Core 状态，停止并按 ARCHITECTURE 修正。
