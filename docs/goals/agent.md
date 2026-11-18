# goal 语句：Agent

本文件收录「Agent」章节的 goal 语句，切片表在[Agent 实施计划](../../spec/plan/agent.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-58 | T-12.1 | Provider 客户端 | — | 已完成（PR #16） |
| goal-118 | T-12.9 | Provider：OpenAI Responses | T-12.1 | 已完成（PR #126） |
| goal-119 | T-12.10 | Provider：Gemini 原生协议 | T-12.1 | 已完成（PR #127） |
| goal-120 | T-12.11 | Provider：Ollama 原生协议 | T-12.1 | 已完成（PR #128） |
| goal-134 | T-12.12 | Provider 改用官方 SDK | T-12.9、T-12.10、T-12.11 | 已完成（PR #134） |
| goal-135 | T-12.13 | Provider 图片输入 | T-12.12 | 已完成（PR #138） |
| goal-59 | T-12.2 | Agent 对话与流式显示 | T-03.9、T-04.3、T-05.4、T-08.4、T-08.11、T-11.3、T-12.1、T-12.9、T-12.10、T-12.11、T-09.4 | 已完成（PR #170） |
| goal-83 | T-12.8 | 文本生成服务 | T-03.8、T-06.1、T-08.14、T-12.2 | 已完成（PR #175） |
| goal-136 | T-12.14 | 对话中的图片输入 | T-04.8、T-12.2、T-12.13 | 已完成（PR #189） |
| goal-60 | T-12.3 | 工具循环与 MCP | T-11.2、T-12.2 | 已完成（PR #178） |
| goal-61 | T-12.4 | Agent 的记忆 | T-08.5、T-11.2、T-12.2 | 已完成（PR #182） |
| goal-62 | T-12.5 | 技能 | T-12.3 | 已完成（PR #188） |
| goal-63 | T-12.6 | 定时 Agent 任务 | T-07.9、T-08.5、T-12.3 | 已完成（PR #193） |
| goal-64 | T-12.7 | 官方 Agent 组件打包与 T-12 验收 | T-12.4、T-12.5、T-12.6、T-12.8、T-12.14 | 已完成（PR #195） |
| goal-143 | T-12.15 | 模型两层配置 | T-12.7 | 已完成（PR #207） |
| goal-144 | T-12.16 | 预留的五类模型与模型管理路由 | T-12.15 | 待派发 |
| goal-145 | T-12.17 | 对话重试与回退模型、图像转述 | T-12.15 | 待派发 |
| goal-159 | T-12.18 | Agent 的触发方式、唤醒词与成员独立会话 | T-12.17 | 等待前置（T-12.17） |
| goal-164 | T-12.19 | Agent 接入配置文件 | T-04.18、T-12.18 | 等待前置（T-04.18、T-12.18） |
| goal-167 | T-12.20 | 官方 Agent 的插件页面 | T-04.17、T-12.19 | 等待前置（T-04.17、T-12.19） |

## 第 1 波

### goal-58 · T-12.1 Provider 客户端

状态：已完成（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`；2026-10-28 按 D-64 修正：增加 `openai` 协议、停止原因与用量，包移到 `agent/provider`）。本 goal 的原始语句未保留；下面记录当前的实际接口，供后继切片引用。

目标：Provider 客户端：OpenAI、OpenAI 兼容与 Anthropic Messages 的流式请求、SSE 解析、工具调用解析、停止原因与用量、本地模拟服务
前置：T-02。
依据：PLAN §3.1 的 T-12.1 行、§4 的 T-12 步骤 1（接口与流语义）；验收：V-57 的 Provider 分项。

接口（main 上的实际代码）：

```go
// agent/provider
type Protocol string // OpenAI = "openai"、OpenAICompatible = "openai-compatible"、AnthropicMessages = "anthropic-messages"
type Config struct { Protocol Protocol; BaseURL string; APIKey string; Model string }
type Message struct { Role string; Content string; ToolCalls []ToolCall; ToolResults []ToolResult } // Role：system、user、assistant
type Tool struct { Name string; Description string; InputSchema json.RawMessage }
type Request struct { Messages []Message; Tools []Tool; MaxTokens int } // MaxTokens 必须大于 0
type ToolCall struct { ID string; Name string; Arguments json.RawMessage }
type ToolResult struct { ToolCallID string; Content string }
type Event struct { Text string; ToolCall *ToolCall } // ToolCall 非 nil 时是组装完整的一次工具调用
func New(config Config, httpClient *http.Client) (*Client, error) // httpClient 由调用方注入
type StopReason string // StopEnd = "end"、StopToolUse = "tool_use"、StopMaxTokens = "max_tokens"；其他取值原样保留
type Usage struct { InputTokens int; OutputTokens int }
type Result struct { StopReason StopReason; Usage *Usage } // Usage 为 nil 表示 Provider 没有返回用量
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
// Stream 不重发请求；已交给 emit 的文本增量由调用方保留；错误不含 API key 与认证头，附带的 Provider 说明中回显的 key 替换为 [REDACTED]。
```

## 第 2 波

### goal-118 · T-12.9 Provider：OpenAI Responses

状态：已完成（[PR #126](https://github.com/AlonaBot/AlonaBot/pull/126)，合并提交 `594bbac`），证据见 TEST-MATRIX §3.2。派发时：前置 T-12.1（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`）已合并。2026-11-09 审查者按 main `3f7a380` 的 `agent/provider` 编写：`provider.go` 有 `Protocol` 常量 `OpenAI`、`OpenAICompatible`、`AnthropicMessages`，`New` 逐个比较协议值，`Stream` 用 `anthropic` 布尔值在 `openAIBody`/`anthropicBody` 与 `streamOpenAI`/`streamAnthropic` 之间分派，`readSSE`、`describeError`、`validateRequest`、`isJSONObject` 可以直接复用；`provider_test.go` 的 `TestStream` 是表驱动的集成测试，`mockProvider` 按协议校验路径、认证头与请求体，`testRequest` 含两条 system、带两个工具调用的 assistant 与带两个工具结果的 user 消息。可以与 goal-119、goal-120 并行，共享文件按[执行顺序 §3](../execution-order.md#3-共享文件与冲突规则)「`agent/provider/` 的新协议」一行处理。

```text
/goal 使用 $alonabot-execute 推进 goal-118（切片 T-12.9：Provider：OpenAI Responses）。goal 目录 goal-118，分支 goal-118，worktree .worktrees/goal-118，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：agent/provider 增加 openai-responses 协议：按无状态方式编码 Responses API 请求（store: false、完整历史放进 input），解析 SSE 事件得到文本增量、完整的工具调用、停止原因与用量；reasoning 项随它之后的工具调用写入新增的 ToolCall.Replay，并在后续请求中原样放回对应的 function_call 之前。语义与已有协议相同：只有终止事件才算成功，不自动重试，错误不含 API key。

依据：PLAN T-12 步骤 1（共同语义）与步骤 10「OpenAI Responses」；DECISIONS D-38、D-103；TEST-MATRIX V-57 的 Provider 分项。

范围（只新增或修改这些）：
- agent/provider/provider.go：追加常量 OpenAIResponses；New 的协议校验与 Stream 的分派各加一个分支；ToolCall 追加 Replay 字段；包文档的协议列表
- 新增 agent/provider/responses.go：请求编码与流解析
- agent/provider/provider_test.go：TestStream 的用例表、mockProvider 与 testRequest 只追加
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// agent/provider
const OpenAIResponses Protocol = "openai-responses"
type ToolCall struct {
    ID        string
    Name      string
    Arguments json.RawMessage
    Replay    json.RawMessage // 协议要求随这次调用原样回传的不透明 JSON，由流解析写入；调用方把整个 ToolCall 原样放回 assistant 消息，不解析。openai、openai-compatible 与 anthropic-messages 不产生也不读取它
}
（goal-119 先合并时 Replay 已存在，同步 main 后复用，不重复定义。）

行为：
- 请求：POST <BaseURL>/responses，Authorization: Bearer <APIKey>，Accept: text/event-stream。请求体为 model、input、stream: true、store: false、max_output_tokens（取 Request.MaxTokens）、include: ["reasoning.encrypted_content"]；有 system 消息时按输入顺序以 "\n\n" 连接写入 instructions（与 Anthropic 相同）；不发 previous_response_id。
- input：按消息顺序展开。user 消息先放每个 ToolResult 的 {"type": "function_call_output", "call_id", "output": Content}，再放非空文本 {"role": "user", "content"}；assistant 消息先放非空文本 {"role": "assistant", "content"}，再对每个 ToolCall 先放 Replay 中的 reasoning 项（Replay 为 JSON 数组时逐项展开），后放 {"type": "function_call", "call_id": ID, "name", "arguments": Arguments 的字符串形式}。function_call 不带 id 字段（store: false 时服务端没有保存这些项）。
- tools：每个 Tool 编码为 {"type": "function", "name", "description", "parameters": InputSchema, "strict": false}。必须显式发 strict: false：Core 与 MCP 工具的 Schema 不满足严格模式的要求。
- 流解析（复用 readSSE，按 data 中的 type 字段分派）：
  - response.output_text.delta：delta 非空时 emit 文本增量。
  - response.output_item.done：item.type 为 reasoning 时，去掉 item 的 id 后暂存；item.type 为 function_call 时，以 call_id 为 ID、name 与 arguments（字符串，必须是 JSON 对象）组成 ToolCall，暂存的 reasoning 项作为 JSON 数组写入其 Replay 并清空暂存，然后 emit。call_id 或 name 为空、arguments 不是对象时返回错误。不使用 response.function_call_arguments.delta，也不从 message 项重复发出文本。
  - response.completed：终止。本次流发出过工具调用时 StopReason 为 tool_use，否则为 end。
  - response.incomplete：终止。response.incomplete_details.reason 为 max_output_tokens 时为 max_tokens，其他取值原样保留。
  - 两种终止事件的 Usage 取 response.usage 的 input_tokens 与 output_tokens（usage 缺失时为 nil）。
  - response.failed 返回错误，说明取 response.error（经 describeError 脱敏）；type 为 error 的事件返回错误，说明取顶层 message（同样脱敏）。
  - 其他事件类型（created、in_progress、reasoning summary、refusal、内置工具等）忽略。
  - 流在终止事件之前结束返回错误；终止事件之后的数据忽略。
- 非 2xx 状态的处理与已有协议相同（describeError 读取 error 字段）。

测试（按编码规范「测试提交规则」）：
- 提交：agent/provider/provider_test.go 的 TestStream 追加 openai-responses 用例（改写既有测试，不新增测试函数）：
  - 主路径：两段文本增量、一个 reasoning 项后跟一个 function_call，结果为 tool_use、用量 5 输入 7 输出，交出的 ToolCall 为 c1/lookup/{"q":"x"}，Replay 是去掉 id 的 reasoning 项数组。mockProvider 校验路径 /v1/responses、Bearer 认证、store 为 false、instructions 为 "one\n\ntwo"、input 中 function_call_output 在 user 文本之前、function_call 不带 id、strict 为 false、max_output_tokens 为 20；用例中给 testRequest 副本的一个工具调用加上 Replay，校验对应的 reasoning 项出现在该 function_call 之前。
  - 缺少终止事件：流在 response.completed 之前结束返回错误，已交出的文本保留。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：response.incomplete 的 max_tokens 与其他原因；response.failed 与 error 事件的说明及 API key 脱敏；未知事件忽略；arguments 不是对象；emit 返回错误；多个工具调用各自带上之前的 reasoning 项；openai 与 anthropic-messages 编码时忽略 Replay。有 OpenAI API key 时，用推理模型对真实 API 跑一次「工具调用 → 带工具结果的第二次请求」，确认第二次请求被接受；没有时在 PR 中写「真实 Provider 未运行」。

完成条件：gofmt -l 无输出；go vet ./agent/...；go test -race -count=1 ./agent/provider/；go build ./...。

不做：Agent 的配置与预设（T-12.2）；工具名映射（T-12.3）；previous_response_id 与服务端会话；内置工具（web_search 等）；图片与文件输入。
```

### goal-119 · T-12.10 Provider：Gemini 原生协议

状态：已完成（[PR #127](https://github.com/AlonaBot/AlonaBot/pull/127)，合并提交 `4f1c39d`），证据见 TEST-MATRIX §3.2。派发时：前置 T-12.1（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`）已合并。main `3f7a380` 的 `agent/provider` 现状见 goal-118 的状态段。可以与 goal-118、goal-120 并行，共享文件按[执行顺序 §3](../execution-order.md#3-共享文件与冲突规则)「`agent/provider/` 的新协议」一行处理。

```text
/goal 使用 $alonabot-execute 推进 goal-119（切片 T-12.10：Provider：Gemini 原生协议）。goal 目录 goal-119，分支 goal-119，worktree .worktrees/goal-119，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：agent/provider 增加 gemini-generate-content 协议：编码 streamGenerateContent 请求（systemInstruction、user/model 内容、functionCall 与 functionResponse part、parametersJsonSchema），解析 SSE 中的 GenerateContentResponse 得到文本增量、完整的工具调用、停止原因与用量；functionCall part 的 thoughtSignature 写入 ToolCall.Replay 并在后续请求中写回；工具结果按 ID 查回调用的工具名。

依据：PLAN T-12 步骤 1（共同语义）与步骤 10「Gemini」；DECISIONS D-38、D-103；TEST-MATRIX V-57 的 Provider 分项。

范围（只新增或修改这些）：
- agent/provider/provider.go：追加常量 GeminiGenerateContent；New 的协议校验与 Stream 的分派各加一个分支；ToolCall 追加 Replay 字段（goal-118 已合并时复用）；包文档的协议列表
- 新增 agent/provider/gemini.go：请求编码与流解析
- 新增 agent/provider/toolcall.go：按 ToolCallID 查找工具名、生成调用 ID（goal-120 已合并时复用）
- agent/provider/provider_test.go：TestStream 的用例表、mockProvider 与 testRequest 只追加
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// agent/provider
const GeminiGenerateContent Protocol = "gemini-generate-content"
ToolCall.Replay json.RawMessage 的定义与 goal-118 相同（见 PLAN T-12 步骤 1 的代码块）。本协议写入的 Replay 是 JSON 字符串形式的 thoughtSignature。

行为：
- 请求：POST <BaseURL>/models/<Model>:streamGenerateContent?alt=sse（BaseURL 例如 https://generativelanguage.googleapis.com/v1beta；Model 不含 models/ 前缀，用 url.PathEscape 转义后作为一个路径段），请求头 x-goog-api-key: <APIKey>，不发 Authorization。错误文本不含完整 URL。
- 请求体：system 消息按输入顺序以 "\n\n" 连接，写入 systemInstruction: {"parts": [{"text"}]}；contents 按消息顺序：user 为 role "user"，parts 先放每个 ToolResult 的 {"functionResponse": {"id": ToolCallID, "name", "response": {"output": Content}}}，再放非空文本 {"text"}；assistant 为 role "model"，parts 先放非空文本，再放每个 ToolCall 的 {"functionCall": {"id", "name", "args": Arguments}}，Replay 非空时同一 part 加 "thoughtSignature": Replay 中的字符串。tools 为 [{"functionDeclarations": [{"name", "description", "parametersJsonSchema": InputSchema}]}]（没有工具时省略）；generationConfig.maxOutputTokens 取 Request.MaxTokens。
- 工具名查找（toolcall.go）：functionResponse 的 name 取同一 Request 中、该 ToolResult 所在消息之前最近的 assistant 消息里 ID 相同的 ToolCall 的 Name；找不到时 Stream 在发请求前返回错误。
- 流解析（复用 readSSE）：每个 data 是一个 GenerateContentResponse。带 error 字段时返回错误（经 describeError 脱敏）。只读 candidates[0]：
  - content.parts 中 thought 为 true 的 part 忽略；text 非空时 emit 文本增量；functionCall 为一次完整的工具调用：id 为空时由 toolcall.go 生成（crypto/rand，同一 Client 内不重复），name 必须非空，args 必须是 JSON 对象（缺失按 {} 处理），part 的 thoughtSignature 非空时以 JSON 字符串写入 Replay，然后 emit。
  - finishReason 非空时记下：STOP 在本次流发出过工具调用时为 tool_use，否则为 end；MAX_TOKENS 为 max_tokens；其他取值（SAFETY、MALFORMED_FUNCTION_CALL 等）原样保留。
  - usageMetadata 取最后一次出现的值：InputTokens 为 promptTokenCount（已含缓存命中的 token），OutputTokens 为 candidatesTokenCount 与 thoughtsTokenCount 之和。
  - 没有候选且 promptFeedback.blockReason 非空时返回错误，说明为该原因。
- 流在两个事件之间结束（EOF）且此前收到过 finishReason 才算成功；否则返回错误。
- 非 2xx 状态的处理与已有协议相同（Gemini 的错误体为 {"error": {"code", "message", "status"}}，describeError 取 message）。

测试（按编码规范「测试提交规则」）：
- 提交：agent/provider/provider_test.go 的 TestStream 追加 gemini-generate-content 用例（改写既有测试，不新增测试函数）：
  - 主路径：两段文本、一个带 id "c1" 与 thoughtSignature 的 functionCall、finishReason STOP、usageMetadata（promptTokenCount 5，candidatesTokenCount 4，thoughtsTokenCount 3），结果为 tool_use、用量 5/7，ToolCall 为 c1/lookup/{"q":"x"} 且 Replay 为该签名。mockProvider 校验路径 /v1/models/test-model:streamGenerateContent、查询 alt=sse、x-goog-api-key、systemInstruction 为 "one\n\ntwo"、model 角色的文本在 functionCall 之前、user 的 functionResponse 在文本之前且 name 为 lookup、parametersJsonSchema 与 maxOutputTokens；用例中给 testRequest 副本的一个工具调用加上 Replay，校验对应 part 的 thoughtSignature。
  - 缺少 finishReason：流结束前没有 finishReason 返回错误，已交出的文本保留。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：functionCall 不带 id 时生成不重复的 ID；ToolResult 找不到对应调用；只有 blockReason；MAX_TOKENS 与 SAFETY；thought part 不交出；data 中的 error 及 API key 脱敏；args 不是对象；emit 返回错误；Model 含需要转义的字符。有 Gemini API key 时，用 Gemini 3 模型对真实 API 跑一次「工具调用 → 带工具结果的第二次请求」，确认带回 thoughtSignature 的请求被接受；没有时在 PR 中写「真实 Provider 未运行」。

完成条件：gofmt -l 无输出；go vet ./agent/...；go test -race -count=1 ./agent/provider/；go build ./...。

不做：Agent 的配置与预设（T-12.2）；工具名映射（T-12.3）；Vertex AI 的认证与端点；cachedContents、代码执行与 Google 搜索等内置工具；图片与文件输入；text part 的 thoughtSignature（只有 functionCall 的签名要求回传）。
```

### goal-120 · T-12.11 Provider：Ollama 原生协议

状态：已完成（[PR #128](https://github.com/AlonaBot/AlonaBot/pull/128)，合并提交 `47a0cd5`），证据见 TEST-MATRIX §3.2。派发时：前置 T-12.1（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`）已合并。main `3f7a380` 的 `agent/provider` 现状见 goal-118 的状态段；`New` 目前要求 APIKey 非空，`readSSE` 只适用于 SSE。可以与 goal-118、goal-119 并行，共享文件按[执行顺序 §3](../execution-order.md#3-共享文件与冲突规则)「`agent/provider/` 的新协议」一行处理。

```text
/goal 使用 $alonabot-execute 推进 goal-120（切片 T-12.11：Provider：Ollama 原生协议）。goal 目录 goal-120，分支 goal-120，worktree .worktrees/goal-120，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：agent/provider 增加 ollama-chat 协议：编码 /api/chat 请求（消息、tool 角色的工具结果、对象形式的工具参数、options.num_predict），逐行解析 NDJSON 流得到文本增量、完整的工具调用、停止原因与用量；API key 可以为空，非空时用 Bearer 认证；工具调用没有 ID，由客户端生成，工具结果按 ID 查回工具名。

依据：PLAN T-12 步骤 1（共同语义）与步骤 10「Ollama」；DECISIONS D-38、D-103；TEST-MATRIX V-57 的 Provider 分项。

范围（只新增或修改这些）：
- agent/provider/provider.go：追加常量 OllamaChat；New 的协议校验与 Stream 的分派各加一个分支，New 对 ollama-chat 允许空 APIKey；包文档的协议列表
- 新增 agent/provider/ollama.go：请求编码与 NDJSON 流解析
- 新增 agent/provider/toolcall.go：按 ToolCallID 查找工具名、生成调用 ID（goal-119 已合并时复用）
- agent/provider/provider_test.go：TestStream 的用例表、mockProvider 与 testRequest 只追加；已有断言中工具调用 ID 为 "c1" 的检查对本协议改为非空（改写既有断言）
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// agent/provider
const OllamaChat Protocol = "ollama-chat"
// New：Protocol 为 OllamaChat 时 APIKey 可以为空，其他协议仍要求非空；Model 始终必填。

行为：
- 请求：POST <BaseURL>/api/chat（BaseURL 为服务根地址，例如 http://localhost:11434），Accept: application/x-ndjson；APIKey 非空时加 Authorization: Bearer <APIKey>，为空时不发认证头。
- 请求体：model、stream: true、options: {"num_predict": Request.MaxTokens}、messages、tools（没有工具时省略）。messages 按顺序：system 为 {"role": "system", "content"}；user 先放每个 ToolResult 的 {"role": "tool", "content": Content, "tool_name": 查到的工具名}，再放非空文本 {"role": "user", "content"}；assistant 为 {"role": "assistant", "content", "tool_calls": [{"function": {"name", "arguments": Arguments 原样作为 JSON 对象}}]}（没有工具调用时省略 tool_calls）。tools 每项为 {"type": "function", "function": {"name", "description", "parameters": InputSchema}}。
- 工具名查找（toolcall.go）：tool_name 取同一 Request 中、该 ToolResult 所在消息之前最近的 assistant 消息里 ID 相同的 ToolCall 的 Name；找不到时 Stream 在发请求前返回错误。
- 流解析：响应是逐行的 JSON，不是 SSE（不使用 readSSE）。空行忽略；每行解码为一个对象，不是合法 JSON 时返回错误：
  - 带 error 字段（字符串）时返回错误，说明经 describeError 脱敏。
  - message.content 非空时 emit 文本增量；message.thinking 忽略；message.tool_calls 中每项为一次完整的工具调用：ID 由 toolcall.go 生成（crypto/rand，同一 Client 内不重复），function.name 必须非空，function.arguments 必须是 JSON 对象，然后 emit。
  - done 为 true 的行是终止行：done_reason 为 stop 时，本次流发出过工具调用为 tool_use，否则为 end；length 为 max_tokens；其他取值原样保留，缺失为空。Usage 在终止行带 eval_count 时为 {InputTokens: prompt_eval_count（缺失为 0）, OutputTokens: eval_count}，否则为 nil。终止行之后的内容忽略。
  - 流在终止行之前结束，或最后一行没有换行且不是合法 JSON，返回错误。
- 非 2xx 状态的处理与已有协议相同（Ollama 的错误体为 {"error": "说明"}，describeError 已支持字符串）。

测试（按编码规范「测试提交规则」）：
- 提交：agent/provider/provider_test.go 的 TestStream 追加 ollama-chat 用例（改写既有测试，不新增测试函数）：
  - 主路径：两行文本增量、一行带一个工具调用（lookup、{"q":"x"}）、终止行 done_reason stop、prompt_eval_count 5、eval_count 7，结果为 tool_use、用量 5/7，ToolCall 的 ID 非空、name 与 arguments 正确。mockProvider 校验路径 /v1/api/chat（测试的 BaseURL 带 /v1）、Bearer 认证、options.num_predict 为 20、tool 角色消息在 user 文本之前且 tool_name 为 lookup、assistant 的 arguments 是对象。
  - 缺少终止行：流在 done 之前结束返回错误，已交出的文本保留。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：空 APIKey 时 New 成功且不发 Authorization（其他协议空 APIKey 仍被拒）；ToolResult 找不到对应调用；生成的 ID 不重复；done_reason length；错误行及 API key 脱敏；畸形行；arguments 不是对象；emit 返回错误；message.thinking 不交出。本机有 Ollama 时对一个支持工具的模型跑一次「工具调用 → 带工具结果的第二次请求」；没有时在 PR 中写「真实 Provider 未运行」。

完成条件：gofmt -l 无输出；go vet ./agent/...；go test -race -count=1 ./agent/provider/；go build ./...。

不做：Agent 的配置与预设（T-12.2）；工具名映射（T-12.3）；/api/generate、模型拉取与管理接口；think 与 format 参数；图片输入。
```

## 第 3 波

### goal-134 · T-12.12 Provider 改用官方 SDK

状态：已完成（[PR #134](https://github.com/AlonaBot/AlonaBot/pull/134)，合并提交 `b679a15`），证据见 TEST-MATRIX §3.2。审查期间按 FIX-3 修改：SDK 解码到类型错误的字段时返回错误（`sdkFieldsValid` 核对 `respjson.Field` 的状态与原值类型）。合入后：Chat Completions 的编码与流解析在 `chat.go`（`chatParams`、`streamOpenAI`），Anthropic 在 `anthropic.go`（`anthropicParams`、`streamAnthropic`），Responses 在 `responses.go`（`responsesParams`、`streamResponses`）；锁定 openai-go v3.98.0、anthropic-sdk-go v2.9.0；anthropic-messages 的 BaseURL 为服务根。派发时：前置 T-12.9（[PR #126](https://github.com/AlonaBot/AlonaBot/pull/126)，合并提交 `594bbac`）、T-12.10（[PR #127](https://github.com/AlonaBot/AlonaBot/pull/127)，合并提交 `4f1c39d`）与 T-12.11（[PR #128](https://github.com/AlonaBot/AlonaBot/pull/128)，合并提交 `47a0cd5`）均已合并。2026-11-10 按 D-132 编写，2026-11-11 审查者按 main `1cbfb2e` 核对：文件划分与范围一致（`responses.go`、`gemini.go`、`ollama.go`、`toolcall.go`，Chat Completions 与 Anthropic 的编码与流解析仍在 `provider.go`）；`Stream` 按协议依次覆盖端点与请求体，认证头为 anthropic 的 `x-api-key`、gemini 的 `x-goog-api-key`，其余协议在 APIKey 非空时发 Bearer（ollama-chat 允许空 key）；gemini 与 ollama 仍使用 `readSSE`（仅 gemini）、`describeError`、`isJSONObject`、`validateRequest`，`ollama.go` 另有 `describeOllamaError`（空 key 时不做替换），`toolcall.go` 的 `newToolCallID` 与 `toolCallName` 供两者使用，这些都保留；`ToolCall.Replay` 已在 `provider.go`。本切片获授权引入 openai-go 与 anthropic-sdk-go（docs/development.md 只授权给 goal-134），并在 docs/development.md 登记锁定的版本。

```text
/goal 使用 $alonabot-execute 推进 goal-134（切片 T-12.12：Provider 改用官方 SDK）。goal 目录 goal-134，分支 goal-134，worktree .worktrees/goal-134，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：agent/provider 的 openai、openai-compatible 与 openai-responses 改经 github.com/openai/openai-go/v3 发出请求、读取流，anthropic-messages 改经 github.com/anthropics/anthropic-sdk-go/v2；删除这四种协议手写的 HTTP 请求与 SSE 解析。公开接口与 PLAN T-12 步骤 1、10 的字段规则不变，只有两处语义变化：两种 Chat Completions 协议的成功条件改为流无错误结束且收到过 finish_reason；anthropic-messages 的 BaseURL 改为服务根地址。gemini-generate-content 与 ollama-chat 不改。

依据：PLAN T-12 步骤 1 与步骤 10「OpenAI Responses」；DECISIONS D-38、D-103、D-132；docs/development.md（两个 SDK 的批准范围）；TEST-MATRIX V-57 的 Provider 分项。

范围（只新增或修改这些）：
- agent/provider/provider.go 与 responses.go；Chat Completions 与 Anthropic 的编码、流解析可以拆到新文件 chat.go、anthropic.go。gemini.go、ollama.go、toolcall.go 不改，它们仍在使用的 readSSE、describeError 等辅助保留
- agent/provider/provider_test.go：改写 TestStream 中这四种协议的用例与 mockProvider
- go.mod、go.sum：引入两个 SDK
- docs/development.md：在已批准依赖一段登记锁定的版本与实际链接的传递依赖
允许的新跨模块边：agent/provider -> github.com/openai/openai-go/v3 的根包与 option、packages/param、packages/ssestream、responses 等所需子包；-> github.com/anthropics/anthropic-sdk-go/v2 的根包与 option、packages/param、packages/ssestream。不导入两个 SDK 的 azure、bedrock、vertex 与其他云厂商子包。

接口：公开符号不变（Protocol 常量、Config、Message、Tool、Request、ToolCall、ToolResult、Event、StopReason、Usage、Result、Client、New、Stream）。

行为：
- 客户端：New 按协议构造一次 SDK 的 Service，保存在 Client 中供 Stream 复用。两个 SDK 都传入注入的 httpClient（option.WithHTTPClient）、Config.BaseURL、Config.APIKey 与 option.WithMaxRetries(0)（两个 SDK 缺省重试 2 次）。不读取环境变量：两个 SDK 的 NewClient 缺省读取 OPENAI_*、ANTHROPIC_* 环境变量（含凭证、组织与自定义请求头）；anthropic-sdk-go 传 option.WithoutEnvironmentDefaults()，openai-go 没有公开的等价选项，直接用 openai.NewChatCompletionService 与 responses.NewResponseService 构造，不经 openai.NewClient。
- 路径：openai 与 openai-compatible 为 <BaseURL>/chat/completions，openai-responses 为 <BaseURL>/responses（BaseURL 仍含 /v1，与 openai-go 的约定一致）；anthropic-messages 改为 <BaseURL>/v1/messages，BaseURL 是服务根地址（例如 https://api.anthropic.com），与 anthropic-sdk-go 及各兼容服务文档中的 base URL 写法一致。New 对 BaseURL 的校验不变。
- 请求体：按 PLAN T-12 步骤 1 与步骤 10 的字段规则用 SDK 的参数类型表达，与现有 mockProvider 校验的请求体语义相同：openai 用 max_completion_tokens 并请求 stream_options.include_usage，openai-compatible 用 max_tokens、不发 stream_options；Tool.InputSchema 原样作为参数 Schema 发出，不丢失 $defs 等字段；Responses 的 store: false、strict: false、include 与 Replay 中 reasoning 项的位置不变；Anthropic 的 system 合并与 content block 顺序不变。SDK 的参数类型没有对应字段时，用它提供的额外字段机制，不改为发出不同的请求。
- 流：遍历 SDK 的 Stream，按到达顺序 emit 文本增量；Chat Completions 的工具调用仍按 index 自行累积（id 与 name 取非空值而不拼接），不使用 SDK 的累积器。终止条件：
  - openai 与 openai-compatible：openai-go 在 [DONE] 与提前 EOF 时都以无错误结束，无法区分，因此流无错误结束且至少一个 choice 给出过 finish_reason 才算成功，否则返回错误；finish_reason 之后、用量 chunk 之前断开时成功，Usage 为 nil。
  - openai-responses 与 anthropic-messages：终止事件（response.completed、response.incomplete；message_stop）仍由客户端判断，流在终止事件之前结束返回错误；未知事件照旧忽略。
  - 停止原因与用量的取值规则不变。
- 错误：非 2xx 时用 errors.As 取 SDK 的错误类型（openai.Error、anthropic.Error），报告状态码与 Provider 的说明，说明中回显的 API key 替换为 [REDACTED]，格式与现有错误相同；不直接包装 SDK 错误的文本（anthropic-sdk-go 的含请求 URL）。流中的错误事件照旧附带脱敏说明。ctx 取消与 emit 返回错误时立即停止并关闭流。

测试（按编码规范「测试提交规则」）：
- 提交：改写 agent/provider/provider_test.go 的 TestStream（换实现，断言不变，不新增测试函数）：四种协议的既有用例继续通过；anthropic-messages 的 mockProvider 路径改为 /v1/messages，BaseURL 改为服务根；Chat Completions 的「缺少终止标记」用例改为「没有 finish_reason」时返回错误。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：finish_reason 之后、用量之前断开为成功且 Usage 为 nil；进程设有 OPENAI_API_KEY、OPENAI_BASE_URL、OPENAI_CUSTOM_HEADERS、ANTHROPIC_API_KEY、ANTHROPIC_AUTH_TOKEN、ANTHROPIC_BASE_URL、ANTHROPIC_CUSTOM_HEADERS 时，请求仍发往 Config 的地址，mockProvider 看不到这些变量带来的请求头；mockProvider 返回 429 与 500 时只收到一次请求；错误说明中的 API key 已脱敏；InputSchema 带 $defs 时原样到达；docs/development.md 的 `go list -deps` 命令输出不含两个 SDK 及其传递依赖（D-52）。有 API key 时对真实 OpenAI 与 Anthropic 各跑一次带工具调用的流，没有时写「真实 Provider 未运行」。

完成条件：gofmt -l 无输出；go vet ./agent/...；go test -race -count=1 ./agent/provider/；go build ./...；go mod tidy 后 go.mod 与 go.sum 没有变化。

不做：图片输入（T-12.13）；Gemini 与 Ollama（保持标准库）；Agent 的预设（T-12.2，anthropic 预设的 base_url 按新约定为 https://api.anthropic.com）；SDK 的 Azure、Bedrock、Vertex 认证（D-103）。
```

## 第 4 波

### goal-135 · T-12.13 Provider 图片输入

状态：已完成（[PR #138](https://github.com/AlonaBot/AlonaBot/pull/138)，合并提交 `9b3bb86`），证据见 TEST-MATRIX §3.2。合入的符号：`provider.Image{MediaType string; Data []byte}`、`Message.Images []Image`；Stream 发请求前校验 Images 只在 user 消息、MediaType 为 image/jpeg、image/png、image/gif 或 image/webp、Data 非空；六种协议把图片按顺序内联在同一条消息的文本之前。派发时：前置 T-12.12（[PR #134](https://github.com/AlonaBot/AlonaBot/pull/134)，合并提交 `b679a15`）已合并。2026-11-10 按 D-132 编写，2026-11-11 审查者按 main `b679a15` 核对：openai 与 openai-compatible 的编码在 `chat.go` 的 `chatParams`（openai-go 的 `ChatCompletionNewParams`；带图片的 user 消息用 `openai.UserMessage` 的 part 数组形式，图片 part 用 `openai.ImageContentPart`、文本用 `openai.TextContentPart`）；openai-responses 在 `responses.go` 的 `responsesParams`（`responses.ResponseInputImageParam` 与 `ResponseInputTextParam`）；anthropic-messages 在 `anthropic.go` 的 `anthropicParams`（`anthropic.NewImageBlockBase64`）；gemini-generate-content 与 ollama-chat 仍为标准库，编码在 `gemini.go` 的 `geminiBody` 与 `ollama.go` 的 `ollamaBody`。四种 SDK 协议用 SDK 的参数类型表达图片，请求 JSON 与下文一致；SDK 没有对应字段时用它的额外字段机制。

```text
/goal 使用 $alonabot-execute 推进 goal-135（切片 T-12.13：Provider 图片输入）。goal 目录 goal-135，分支 goal-135，worktree .worktrees/goal-135，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：agent/provider 的 user 消息可以带图片：新增 Image 类型与 Message.Images，在发出请求前校验；六种协议各按自己的格式把图片内联在同一条消息的文本之前。

依据：PLAN T-12 步骤 1 与步骤 10（各协议的图片编码）；DECISIONS D-132；TEST-MATRIX V-118 的 Provider 分项。

范围（只新增或修改这些）：
- agent/provider/provider.go：Image 类型、Message.Images 字段与请求校验
- 六种协议的编码函数：`chat.go` 的 `chatParams`、`responses.go` 的 `responsesParams`、`anthropic.go` 的 `anthropicParams`、`gemini.go` 的 `geminiBody`、`ollama.go` 的 `ollamaBody`；只改编码部分，流解析不改
- agent/provider/provider_test.go：TestStream 的用例表、mockProvider 与 testRequest 只追加
允许的新跨模块边：无。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
// agent/provider
type Image struct {
    MediaType string // image/jpeg、image/png、image/gif 或 image/webp
    Data      []byte // 图片的原始字节，编码时转为 base64
}
type Message struct {
    Role        string
    Content     string
    ToolCalls   []ToolCall
    ToolResults []ToolResult
    Images      []Image // 只出现在 user 消息中
}

行为：
- 校验（Stream 发请求前）：Images 只能出现在 user 消息；MediaType 必须是上述四种之一；Data 不能为空。不检查图片的大小、尺寸与张数，Provider 拒绝时按它的错误返回。
- 编码：图片按 Images 的顺序排在同一条消息的文本之前；只有图片没有文本的 user 消息同样发出；不带图片的请求体不变。
  - openai 与 openai-compatible：带图片的 user 消息 content 为 part 数组，每张图片一个 {"type": "image_url", "image_url": {"url": "data:<MediaType>;base64,<Data>"}}，Content 非空时最后一个是 {"type": "text", "text"}；工具结果的 tool 消息仍排在这条 user 消息之前。
  - openai-responses：带图片的 user 输入项的 content 为数组，每张图片一个 {"type": "input_image", "image_url": data URL}，文本为最后一个 {"type": "input_text", "text"}；function_call_output 仍排在它之前。
  - anthropic-messages：每张图片一个 {"type": "image", "source": {"type": "base64", "media_type", "data"}} block，排在 tool_result 之后、text 之前。
  - gemini-generate-content：每张图片一个 {"inlineData": {"mimeType", "data"}} part，排在 functionResponse 之后、文本之前。
  - ollama-chat：user 消息的 images 为 base64 字符串数组（不带 data URL 前缀）；工具结果的 tool 消息仍排在它之前。

测试（按编码规范「测试提交规则」）：
- 提交：TestStream 追加带图片的用例表项（每种协议一项，改写既有测试，不新增测试函数）：testRequest 副本的 user 消息带一张 PNG 与文本，mockProvider 校验该协议的图片表示与它在文本之前的位置。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：多张图片的顺序；只有图片没有文本；与工具结果同在一条消息；Images 出现在 assistant 或 system 消息、MediaType 为 image/bmp、Data 为空时在发请求前返回错误且 mockProvider 没有收到请求。有 API key 时对真实 Provider 各发一次带图片的请求，没有时写「真实 Provider 未运行」。

完成条件：gofmt -l 无输出；go vet ./agent/...；go test -race -count=1 ./agent/provider/；go build ./...。

不做：读取入站媒体、image_input 开关与历史回放（T-12.14）；图片的压缩与缩放；工具结果中的图片；llm.generate 的图片输入。
```

## 第 13 波

### goal-59 · T-12.2 Agent 对话与流式显示

状态：已完成（[PR #170](https://github.com/AlonaBot/AlonaBot/pull/170)，合并提交 `eca6310`），证据见 TEST-MATRIX §3.2。合入：`agent.New() *sdk.Plugin`（`alonabot-agent-official`，薄入口 `agent/cmd/alonabot-agent`）与命令 /ask、/agent on|off、/persona、/provider、/reset；受管 SQLite 的 `chat_settings`（Provider、人格、自由对话开关与 reset 代次）；`InvocationContext.command_consumed = 6`、`sdk.NotCommand() Rule` 与 `(*sdk.Context).CommandConsumed() bool`（D-141，待维护者确认）；`c.Stream` 无编辑时按实例 text_limit 换条。

```text
/goal 使用 $alonabot-execute 推进 goal-59（切片 T-12.2：Agent 对话与流式显示）。goal 目录 goal-59，分支 goal-59，worktree .worktrees/goal-59，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立官方 Agent 组件 agent/（工程 manifest `role: agent`）的骨架与配置（Provider 与预设、人格、历史条数），注册 /ask、/agent on|off、/persona、/provider、/reset 与 message.created 事件路由；一轮对话在延续调用中执行：读取会话历史、组装请求、经 Provider 流式接收，用 SDK 的 c.Stream 显示（D-135），结束后追加会话条目。Provider 出错或中途断开时保留已显示内容并按 Actor locale 报告，不自动重发。

依据：PLAN agent T-12 步骤 2、3（2026-11-12 按 D-135 修订的流式显示）；CONTRACTS §13（Agent 段）、§2.4、§5.5（配置）、§8、MESSAGE-MODEL §4A；DECISIONS D-27、D-38、D-59、D-76、D-98、D-103、D-132、D-135；TEST-MATRIX V-57。

范围（只新增或修改这些）：
- agent/：新增可导入的根组件包 agent（`func New() *sdk.Plugin`）、cmd/alonabot-agent/ 的薄入口（只调用 sdk.Run(agent.New())）、plugin.yaml（`role: agent`、`icon`）、config_schema.json、i18n/en-US.json 与 zh-CN.json、assets/icons/；生成文件 alonabot_gen.go 只由生成器生成；agent/provider 已存在，不改
- agent/agent_test.go（见测试）
允许的新跨模块边：agent/... -> sdk/...、-> contracts/...（已允许）；agent 的产品代码不 import internal/...，测试可以使用根包 alonabot。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
   package agent
   func New() *sdk.Plugin // 调用 sdk.UseGenerated(p, alonabotGenerated)
配置写在手写的 config_schema.json（D-76），Go 代码经 p.Config()/c.Config() 读取：
- providers：template_list，每个预设一个模板（openai、anthropic、gemini、deepseek、ollama、custom），模板 items 为 name（required）、protocol（options 为六种协议，default 取预设）、base_url（default 取预设）、api_key（secret）、model（required）、max_tokens（int，default 2048，min 1）；预设缺省值：openai 为 openai 与 https://api.openai.com/v1，anthropic 为 anthropic-messages 与服务根地址 https://api.anthropic.com，gemini 为 gemini-generate-content 与 https://generativelanguage.googleapis.com/v1beta，deepseek 为 openai-compatible 与 https://api.deepseek.com/v1，ollama 为 ollama-chat 与 http://localhost:11434，custom 为 openai-compatible 且 base_url 必填。image_input 由 T-12.14 加入。
- default_provider：string，options_from 不使用（T-12.8 提供 llm.providers 后由该切片接入）。
- personas：list，items 为 name（required）与 prompt（text）；default_persona：string。
- history_entries：int，default 20，min 0，max 500。
- mcp_servers 与 skills_dir 由 T-12.3、T-12.5 加入，本切片不声明。
OnConfigLoad 之后的校验在 OnStart 中完成：providers 的 name 唯一、api_key 只在 ollama-chat 时可以为空、default_provider 与 default_persona 指向存在的项（列表为空时允许为空）；不合法时 OnStart 返回错误。

行为：
- 数据：每个聊天（会话字符串 c.Session()）当前的人格、Provider 与自由对话开关保存在 Agent 的受管 SQLite（gormdb，versioned 迁移）。
- 命令：/ask（RawTail 的 text，缺省 sdk.User）、/agent on|off（sdk.Admin）、/persona [name]（Admin，无参数时列出）、/provider [name]（Admin）、/reset（User）；说明经 Description 指向 i18n。自由对话开启时，该聊天中 to_me 的非命令 message.created 也触发一轮（On("chat", sdk.MessageCreated).Rule(sdk.ToMe())，Handler 内检查开关）；私聊的消息都是 to_me。默认关闭。
- 一轮对话（Handler 调用 c.Continue，在延续调用中执行）：
  1) 没有可用 Provider 时用 c.T 回复 agent.no_provider，结束。
  2) 会话键为 c.Session() 的字符串；ListConversation 读取最近 history_entries 条（从末尾向前取，按序号顺序组装）。条目为 JSON 对象 {"role": "user"|"assistant", "speaker": 外部用户 ID（群聊时）, "content": 文本}。
  3) 以人格为 system 消息组装 provider.Request（MaxTokens 取 Provider 配置），provider.New 使用标准库 http.Client（沿用进程的代理环境变量）。
  4) 显示：用 c.Stream，回调每次返回 Provider 新到的文本增量，生成结束返回 io.EOF。编辑节流、平台单条上限（实例的 text_limit）与换条都由 c.Stream 处理（D-135），Agent 不另设单条上限、不自行切分。回复以纯文本发送。
  5) 结束后 AppendConversation 追加本轮的 user 与 assistant 条目。
  6) Provider 返回错误或中途断开：c.Stream 已发出的内容保留，之后用 c.Reply 发送 agent.error（按 Actor locale，参数为错误码），不追加 assistant 条目，不重发请求；Provider 的错误说明只写日志。
- /reset：ClearConversation 清除该会话。
- 工具、MCP、记忆、技能、图片输入与 llm.generate 分别在 T-12.3、T-12.4、T-12.5、T-12.14、T-12.8 接入；本切片不调用它们。

测试（按编码规范「测试提交规则」）：
- 提交：agent/agent_test.go（//go:build unix）：用根包库模式 alonabot.Start(ctx, alonabot.DataDir(dir), alonabot.Plugin(agent.New()), alonabot.Adapter(<测试中定义的回环 Adapter>))，Agent 的配置写在 <dir>/plugins/<agent id>/config.json（T-13.2），回环 Adapter 的实例写在 <dir>/core_config.json（T-09.4）；回环 Adapter 声明 message.send（带 text_limit）与 message.edit，记录收到的发送与编辑；Provider 为 httptest 模拟服务（openai、openai-compatible 与 anthropic-messages 各一次；其他协议的流解析由 TestStream 覆盖）：
  - /ask hello：先收到一条发送，随后若干次编辑，最后一次编辑是完整回复；会话中追加了 user 与 assistant 条目。
  - 实例不声明 message.edit：只收到完整回复的发送，没有中间编辑；回复超过 text_limit 时为多条。
  - 模拟 Provider 中途断开：已显示内容保留，最后有错误说明，会话中没有 assistant 条目，Provider 只收到一次请求。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：OnStart 的配置校验；预设缺省值与显式值优先；会话条目格式与历史条数；自由对话关闭时群聊中的 to_me 消息不触发、Admin /agent on 后触发、非 to_me 的消息不触发；/persona 切换后请求的 system 消息改变；/reset 后历史为空；普通用户执行 /agent 被拒；经 cpa 兼容网关的真实 Provider 一轮对话（凭证从仓库 tmp/agent-debug/providers.json 读取，文件不存在时跳过；证据写「经 cpa 兼容网关运行」，不写成官方服务；不打印、不提交凭证）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... . ./internal/core/...；go build ./...；在 agent 运行 alonabot plugin check 通过。

不做：工具循环与 MCP（T-12.3）；记忆（T-12.4）；技能（T-12.5）；定时任务（T-12.6）；打包（T-12.7）；llm.generate 与 @agent 解析（T-12.8）；图片输入（T-12.14）。
```

## 第 14 波

### goal-83 · T-12.8 文本生成服务

状态：已完成（[PR #175](https://github.com/AlonaBot/AlonaBot/pull/175)，合并提交 `ac6083a`），证据见 TEST-MATRIX §3.2。合入：公开包 `sdk/llm`（`Request`、`Response`（用量为 `*int`）、`Message`、`AgentTarget = "@agent"`、`ServiceName = "llm.generate"`、`GenerateService`、`Generate(ctx, req)`、`Ask(ctx, text)`）；Broker 与 options.Service 的 Agent 定位注入，根包 `agent_target.go`；官方 Agent 的 llm.generate（按 llm_service_allowed_plugins 授权）与 llm.providers 选项来源。以下为派发时的语句。

- `agent/agent.go`：`func New() *sdk.Plugin`；组件 ID 为 `alonabot-agent-official`。配置通过 `readConfiguration(values config.Values) (configuration, error)` 读取，`configuration` 与 `configuredProvider{name string; client *provider.Client; maxTokens int}` 均为私有类型；当前没有公开 `Config`，也没有保存供服务响应使用的 model、protocol 元数据。`agent/config_schema.json` 已有六个 Provider 模板及 default_provider。
- `sdk/services.go`：`func JSON[In, Out any](target string) Service[In, Out]`；`func Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out]`；`func (s *ServiceBuilder[In, Out]) Access(a Access, allowlist ...string) *ServiceBuilder[In, Out]`；`func (s *ServiceBuilder[In, Out]) Handle(h func(*Context, In) (Out, error))`；`func (s Service[In, Out]) Call(ctx context.Context, in In) (Out, error)`；`func (c *Context) Caller() string`。`AccessPublic` 为现有访问规则。类型的 schema_id 是 `<Go import path>.<type name>`，缺省版本为 1；Schema 原文来自生成器。旧 JSONService、CallJSON、WithAccess 和 Services Hook 的写法已过时。
- `sdk/options.go`：`func ServeOptions(r Registrar, serviceName string, fn func(ctx context.Context, q OptionsQuery) (OptionsPage, error))`。OptionsQuery 含 Search、Filter、Locale、Limit、Values；OptionsPage 含 Options、More；Option 含 Value、Label、Description；访问规则固定为 private。`contracts/proto/alonabot/v1/route.proto` 已有 `ServiceSpec.options_source = 6`。
- `internal/core/broker/broker.go`：`func New(opt Options) *Broker`；`func (b *Broker) Call(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response`；`func (b *Broker) CallSession(ctx context.Context, c, target router.Caller, in *wire.Inbound) wire.Response`。Options 已含 Router、RequestTimeout、MaxDepth、Metrics；定位处先用 validate.PluginID 校验，当前会拒绝 `@agent`。
- `internal/core/router/router.go`：`func (r *Router) Caller(pluginID string) (Caller, bool)` 与 `func (r *Router) PluginFor(c Caller) (Plugin, bool)`；后者校验绑定的 Session。`internal/core/options/options.go`：`func (s *Service) Query(ctx context.Context, componentID string, kind SchemaKind, fieldPath string, q Query) (Page, error)`；其 service 当前直接以来源的组件部分调用 Router.Caller，再检查 options_source，通过 Broker.CallSession 调用同一个 Session，尚缺 `@agent` 解析。
- `internal/core/state/installs.go`：`func (s *Store) Installs(ctx context.Context) ([]Install, error)`；`state/state.go` 的 Install 含 Role、PluginID、Source、Uninstalled，`RoleAgent Role = "agent"`。根包 `assembly.go` 的 `assemble(store *state.Store, dir string, v tune.Values, file *settings.File) (*assembly, error)` 同时构造 Broker 和 options.Service；库模式也创建 dev 来源、role agent 的安装记录。
- `agent/provider/provider.go`：`func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)`；Request 含 Messages、Tools、MaxTokens，Result 含 StopReason、Usage，Usage 为可空的 `*Usage`，其中 InputTokens、OutputTokens 为 int。`agent/conversation.go` 已在 c.Continue 中调用 c.Stream；`sdk.NotCommand() Rule`、`(*sdk.Context).CommandConsumed() bool` 和 `InvocationContext.command_consumed = 6` 已随 #170 合入。
- 测试可用根包 Start 的真实装配；根包 `pipeline_test.go` 已有 `newPipeline(t *testing.T, stdio ...bool) *pipelineHarness`、install、launch、stop、deliver 等辅助，可供本地多进程自测借用。`sdk/generated.go` 已有 `func (p *Plugin) Generated() Generated` 与 ServiceSchemas 资源字段；Agent 当前尚未生成 llm 类型的 Schema，本切片增加声明并生成后，测试调用插件可复用其中同一组 Request/Response Schema。
- 在途冲突：goal-40 涉及 `internal/core/state/`、`internal/core/supervisor/` 和根包装配；PR #171（goal-69）涉及 `alonabot.go`、`assembly.go`、`cmd/alonabot/`；PR #172（goal-57）涉及根包 `t11_e2e_test.go` 与 `pipeline_test.go`。本切片与前两者的直接共享点是 assembly.go；本地复用流水线时还须核对 #172 的辅助改动。goal-136 同改 agent.go、config_schema.json、i18n 和生成资源，按执行顺序 §3 的追加规则处理。#165、#167、#168、#169 和 goal-59 在本基线上均已合并。

```text
/goal 使用 $alonabot-execute 推进 goal-83（切片 T-12.8：文本生成服务）。goal 目录 goal-83，分支 goal-83，worktree .worktrees/goal-83，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 的 Service Broker 将 @agent 解析为当前安装的 Agent；新增 sdk/llm 的共享服务声明、请求响应类型、Schema 标识、Generate 与 Ask；官方 Agent 提供 public 的 llm.generate，按直接调用方许可完成一次文本生成；同时提供私有选项来源 llm.providers，并接通 @agent/llm.providers 的 Core 查询。

依据：spec/PLAN.md §3.1；spec/plan/agent.md §3 T-12 步骤 2、9 及「失败停止条件」，spec/plan/core.md T-06 步骤 2～4、6 的目标定位与 Session 约束、T-08 步骤 11 的选项查询；CONTRACTS §3、§5.1、§9.1、§13、§13.5、§14.1；WIRE-FORMAT §3.3、§4、§5A；MESSAGE-MODEL §4A（沿用现有对话流显示）；DECISIONS D-38、D-51、D-59、D-61、D-70、D-76、D-83、D-86、D-111、D-122、D-129、D-131；TEST-MATRIX V-76、V-82 的 @agent 分项、V-109、V-116 的 llm.providers 分项。

范围（只新增或修改这些）：
- 新增 sdk/llm/llm.go：共享类型、声明、Schema 标识与调用辅助；该目录只增加本地 *_local_test.go，不提交测试。
- internal/core/broker/broker.go：Options 注入 Agent 定位函数；只调整目标语法检查和定位接线；别名的专属辅助放在新增 internal/core/broker/agent.go。
- internal/core/options/options.go：Options 注入同一个 Agent 定位函数；service 将别名归一化后沿用标记检查和 CallSession。
- 新增根包 agent_target.go：从 state.Store 的可信安装记录定位当前 Agent；assembly.go 只向 Broker、options.Service 接入同一函数。
- agent/agent.go：私有 configuration 增加调用方名单，configuredProvider 保留 protocol、model；readConfiguration 读取新增字段；New 登记服务。
- 新增 agent/generate.go、agent/options.go：llm.generate 与 llm.providers 的处理和注册。
- agent/config_schema.json、agent/i18n/en-US.json、agent/i18n/zh-CN.json：追加 llm_service_allowed_plugins 及 config 文案；default_provider 接上本组件的 llm.providers 选项来源。
- agent/alonabot_gen.go：仅经现有生成器更新。
- ARCHITECTURE.md：依赖表登记「根包测试 -> agent/...（仅集成测试）」这条边（本切片获授权修改这一处）。
- 新增根包 agent_service_test.go（unix）：一个覆盖完整主路径的集成测试；本地自测放根包 agent_service_local_test.go 与上述包的 *_local_test.go。
允许的新跨模块边：根包 agent_service_test.go -> agent/...（仅集成测试，用官方 Agent 与本地 Provider）；产品代码沿用根包 -> internal/core/...、agent/... -> sdk/...，sdk/llm 只依赖 sdk、sdk/errcode 与标准库。Agent 定位通过根包注入，Broker 无需新增对 state 的依赖。

接口（照此实现；以下新增接口与 main 已有接口分开列出）：
1. 新增 sdk/llm：
   const ServiceName = "llm.generate"
   const AgentTarget = "@agent"
   const SchemaIn = "github.com/AlonaBot/AlonaBot/sdk/llm.Request"
   const SchemaOut = "github.com/AlonaBot/AlonaBot/sdk/llm.Response"
   const SchemaVersion uint32 = 1
   type Message struct {
       Role string `json:"role"`
       Text string `json:"text"`
   }
   type Request struct {
       Messages []Message `json:"messages"`
       System string `json:"system,omitempty"`
       Provider string `json:"provider,omitempty"`
       MaxOutputTokens int `json:"max_output_tokens,omitempty"`
   }
   type Response struct {
       Text string `json:"text"`
       Provider string `json:"provider"`
       Model string `json:"model"`
       InputTokens *int `json:"input_tokens,omitempty"`
       OutputTokens *int `json:"output_tokens,omitempty"`
   }
   var GenerateService = sdk.JSON[Request, Response](AgentTarget + "/" + ServiceName)
   func Generate(ctx context.Context, req Request) (Response, error)
   func Ask(ctx context.Context, prompt string) (string, error)
   Generate 调用 GenerateService.Call；Ask 调用 Generate。Schema 原文由生成器按类型生成，两个 Schema 标识与 SDK 的类型规则一致，版本使用既有缺省 1。
2. internal/core/broker.Options 与 internal/core/options.Options 各新增：
   Agent func(ctx context.Context) (string, error)
   生产装配由 assemble 显式注入同一个定位函数；其他装配要调用 @agent 时也须注入。已有仅处理普通组件 ID 或内置选项来源的构造方式继续可用，新字段仅在别名定位分支读取。根包的私有定位方法为：
   func (a *assembly) agentID(ctx context.Context) (string, error)
   它读取 a.store.Installs(ctx)，返回 RoleAgent 且 Uninstalled 为 false 的 PluginID；无记录为 "", nil；存储失败保留 error。
3. agent 的新增字段均为私有：
   configuration.llmServiceAllowedPlugins []string
   configuredProvider.protocol provider.Protocol
   configuredProvider.model string
   New() *sdk.Plugin 的签名保留；在 New 中 sdk.Serve(p, llm.GenerateService).Access(sdk.AccessPublic).Handle(...)，并调用 sdk.ServeOptions(p, "llm.providers", fn)。
4. 以下为 main 已有的 SDK 接口，直接复用：
   // sdk/options.go，JSON 键分别为 search、filter,omitempty、locale、limit、values,omitempty。
   type OptionsQuery struct { Search string; Filter map[string]string; Locale string; Limit int; Values []string }
   type OptionsPage struct { Options []Option; More bool } // JSON 键 options、more。
   type Option struct { Value string; Label string; Description string } // JSON 键 value、label、description,omitempty。
   func ServeOptions(r Registrar, serviceName string, fn func(ctx context.Context, q OptionsQuery) (OptionsPage, error))
   func (c *Context) Caller() string
   // internal/core/broker/broker.go
   func (b *Broker) Call(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response
   func (b *Broker) CallSession(ctx context.Context, c, target router.Caller, in *wire.Inbound) wire.Response
   // agent/provider/provider.go
   func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
   proto 字段沿用 ServiceSpec.options_source = 6 与 InvocationContext.plugin_chain = 4，不改 .proto；选项固定 Schema ID 沿用 contracts/validate 的 OptionsQuerySchemaID、OptionsPageSchemaID，均为对应 sdk 类型的 import path + 类型名，版本 1。

行为：
- 别名：先按完整目标分离组件与服务名，服务名继续校验；仅组件部分恰为 @agent 时调用 Agent 定位函数，再按真实 PluginID 执行现有 Session 定位、Activated、访问规则、输入输出准入、祖先链、CALL_CYCLE、深度和取消处理。普通组件 ID 继续使用既有规则；其他 @ 名称返回 INVALID_ARGUMENT。无 Agent、已卸载或未激活为 UNAVAILABLE。定位存储失败记录完整错误，Broker 返回安全的 INTERNAL；选项查询沿用来源不可用的 UNAVAILABLE，并保留内部错误原因。每次定位读取当前记录，使替换后的新 ID 生效；dev 来源的 Agent 同样适用。
- 选项查询：service 先经同一个 Agent 函数把 @agent 变成真实 ID，再取得 router.Caller 与 PluginFor，核对该 Session 的 options_source。CallSession 的 target 参数与 in.Target 的组件部分都使用这一真实 ID，保持标记检查与调用绑定到同一 Session；替换期间旧绑定失效时返回 UNAVAILABLE。
- 配置：llm_service_allowed_plugins 声明为 {"type":"list","description":"Plugins allowed to generate text","default":["*"]}（省略 items 的 list 为字符串列表），空列表允许表达无人获准；在私有 configuration 中读取。default_provider 的 string 字段增加 options_from: {"source":"@agent/llm.providers"}，保留原有配置读取和引用校验。
- Generate/服务校验：messages 至少一条，role 仅 user 或 assistant，max_output_tokens 小于 0 为 INVALID_ARGUMENT；指定不存在的 Provider 为 INVALID_ARGUMENT；未配置任何 Provider 为 FAILED_PRECONDITION。provider 省略或为空时使用 default_provider；max_output_tokens 为 0 时取所选 Provider 的 maxTokens。校验在实际输入边界完成，错误采用 sdk/errcode。
- Ask：prompt 为空串或仅空白时直接返回 INVALID_ARGUMENT；有效 prompt 原样作为唯一 user 消息，省略 system、provider、max_output_tokens，返回 Response.Text，保留 Generate 的错误。
- 许可：llm.generate 为 public。可信直接调用方取 c.Caller()；为空时拒绝；名单含 "*" 或调用方的完整 ID 时允许，其余返回 PERMISSION_DENIED，且不请求 Provider。沿用 §3 的可信调用链，无需用户 Admin 或额外服务权限。
- 请求：system 非空时作为 system 消息，其后按输入顺序转换 messages；不传 Tools。调用所选客户端的 Stream，拼接所有文本；成功条件沿用该协议现有实现（包括 Chat Completions 的 finish_reason 规则）。响应中的 provider、model 取所选配置；Usage 为 nil 时省略两项用量，有 Usage 时包含两项，数值为 0 时也保留。
- Provider 错误或断流映射为 UNAVAILABLE，safe_message 使用固定英文说明，Provider 详情只写日志；不重试。传入 c.Context()，调用方取消或期限到达时取消上游请求，调用方期限按现有 wire/Broker 返回 DEADLINE_EXCEEDED。llm.generate 的处理独立于对话历史、记忆、人格和工具循环。
- llm.providers：用 ServeOptions 的固定私有服务声明。每个已配置 Provider 一个 Option，Value 与 Label 都是 name，Description 为 protocol 与 model；不返回 base_url 或 api_key。Values 非空时只返回存在的这些名称，忽略 Search、Filter、Locale；否则 Search 不区分大小写地匹配 name，当前不设能力 filter；按配置顺序返回，遵守 Limit，剩余结果通过 More 表达。它按配置列出，省略 llm.generate 的调用方名单检查；外插件直接调用按 private 规则被拒。
- 单次 Provider 调用的调用方、provider、model、耗时与用量使用 DEBUG，字段沿用日志契约，消息正文和密钥不写日志。配置、Schema 与 i18n 只追加必要内容，更新生成资源。

测试（按编码规范「测试提交规则」）：
- 提交：根包 agent_service_test.go（//go:build unix）的一个 TestAgentServices 集成测试，分子场景复用真实 Start、临时 data_dir、真实 SDK Session、官方 agent.New() 与 httptest Provider；测试调用插件经 Handler 中的 llm.Generate、llm.Ask 调用。共用 llm.GenerateService 和官方 Agent 生成资源中的 Request/Response Schema；主路径不手写替代 Schema。测试位于根包以直接调用装配中的 options.Service.Query，管理 API 由后继切片提供。
  - 缺省名单调用成功，响应带文本、实际 Provider、model 和用量；收紧为名单后，名单内成功、名单外 PERMISSION_DENIED 且 Provider 请求数不增加；会话存储没有新增条目。用拒绝与允许两个启动配置驱动真实 Session。
  - Ask 的上游请求恰有一条 user 消息，无 system，使用默认 Provider，返回其文本。
  - 从真实字段 options_from 经 Core Query 查询 @agent/llm.providers：返回名称、协议与模型；搜索和按值取回均正确，响应不含地址与密钥；llm.generate 名单收紧仍可查询。外插件直接调用该来源被拒。
  - 停止并移除旧 Agent 的安装身份后，登记并启动另一个 ID、role agent 的测试 Agent（使用同一服务声明与生成 Schema），调用方继续使用 @agent，生成服务与选项查询都到达新 ID。沿用角色唯一性规则，避免同时登记两个 Agent；正式管理替换流程由 T-09.4 的 V-82 证据覆盖。
- 本地自测（*_local_test.go，不提交，PR 写明命令、平台和结果）：空 messages、非法 role、负 max_output_tokens、未知 Provider、没有 Provider；Ask 空白不发调用；Usage 缺失与两项为 0；Provider 非 2xx 与缺少终止条件均为 UNAVAILABLE、一次请求、safe_message 脱敏；调用期限与取消；无 Agent、停用、已卸载、dev Agent；调用链按真实 ID 查循环和深度；选项检查后 Session 被替换；按值查询忽略搜索且缺失值不返回、分页 More；指定 Provider 与 max_tokens 选择；三种原有协议的服务转换。根包 agent_service_local_test.go 可复用 pipeline_test.go 的装配运行 TCP/stdio 的真实多进程调用，不增加 testdata 或修改共享流水线辅助。临时生成一个只调用 llm.Generate/Ask 的独立插件工程，运行现有生成器并启动，核对包装函数里的共享声明能生成调用方所需 Schema。本轮保留 main 的 D-141 路由行为。

完成条件：上述主路径集成通过；gofmt -l 对本次 Go 文件无输出；go vet ./...；go test -race ./agent/... ./sdk/... ./internal/core/... .；go build ./...；go run ./cmd/alonabot plugin check --dir ./agent 通过。用 go generate ./agent 更新生成资源，随后 check 应显示资源最新；go list -deps 核对 ./cmd/alonabot 与 ./sdk 的第三方模块仍在 docs/development.md 清单内。PR 附各 V 分项证据与真实 Provider 的实际运行情况；来源替换后生成和选项查询都使用新的组件 ID，错误路径保持现有取消与 Session 契约。

不做：流式 llm.generate、工具与图片输入；llm_service_allowed_plugins 的专用管理页；plugintest.LLM（T-16.3）；Provider 客户端修改；新增 ServiceBuilder 权限 API；state/supervisor/管理服务重构；热部署、候选 Journal 与真实平台验收。
```

### goal-136 · T-12.14 对话中的图片输入

状态：已完成（[PR #189](https://github.com/AlonaBot/AlonaBot/pull/189)，合并提交 `2268366`），证据见 TEST-MATRIX §3.2（含经 cpa 兼容网关以 pixiv 插图运行的四种协议实测）。合入：Provider 条目的 `image_input`；agent/image.go 的图片读取（sdk.OpenMedia）、MIME 与大小校验、受管 SQLite 存储与按会话键编号，历史中按回放窗口重放、窗口外与关闭时占位，/reset 删除；`conversationEntry` 记录图片编号，`chatSettings` 记录会话图片计数。以下为派发时的语句。

- `agent/agent.go`：`func New() *sdk.Plugin`，私有 `configuration`、`configuredProvider` 与 `readConfiguration(values config.Values) (configuration, error)`；当前 configuredProvider 含 name、client、maxTokens。不存在公开 Config。New 通过 `gormdb.Migrations(p, []gormdb.Migration{{ID: 1, Up: ...}})` 对 chatSettings 执行 AutoMigrate；尚无图片表或图片计数。
- `agent/settings.go`：私有 `chatSettings{Session string; Provider string; Persona string; Enabled bool; ResetGeneration uint64}`，表名 chat_settings；`func (a *component) sessionLock(session string) *sync.Mutex`；`func (a *component) resetConversation(c *sdk.Context) error` 在同一会话锁内递增 ResetGeneration，再 `sdk.ClearConversation`。当前轮次最终追加也持该锁并核对 reset 代次。
- `agent/conversation.go`：`func (a *component) converse(c *sdk.Context, text string) error`；`func recentConversation(ctx context.Context, key string, count int) ([]conversationEntry, error)`；私有 conversationEntry 含 Role、Speaker、Content，对应 JSON role、speaker,omitempty、content。现有流是 provider.Client.Stream 经 iter.Pull2 接到 c.Stream；生成和显示成功后原子追加一组 user、assistant JSON，出错时跳过追加。/ask 与自由对话都在 c.Continue 内执行。
- `sdk/media.go`：`func OpenMedia(ctx context.Context, src message.MediaSource) (io.ReadCloser, error)`，立即读取首块，只接受 incoming 或 media_ref token。`sdk/message/media.go` 的 `MediaSource` 含 `MediaRef string`、`Incoming *IncomingMedia` 等互斥来源；IncomingMedia 含 Token、Size、MIME、Readability、ExpiresAt，`message.Readable` 与 `message.Unreadable` 表示读取能力。
- `sdk/message/message.go`：`type ImageSegment struct { Source MediaSource; Alt, Caption string }`；`func Image(src MediaSource) Segment`；`type Chain []Segment`；单个 Segment 与 Chain 均实现 message.Content。`sdk/sdk.go` 的 `func (c *Context) Event() *message.Event` 返回当前事件，`message.CreatedPayload{PlatformMessageID string; Chain message.Chain; ToMe bool}` 位于 `sdk/message/event.go`，当前图片应从该 Chain 取出。
- `agent/provider/provider.go`：`type Image struct { MediaType string; Data []byte }`；`Message.Images []Image`，Message 另含 Role、Content、ToolCalls、ToolResults；`func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)`。图片只在 user 消息出现，MediaType 限 image/jpeg、image/png、image/gif、image/webp，Data 非空；六种协议编码已实现。
- `sdk/gormdb/open.go`：`func Open(ctx context.Context) (*gorm.DB, error)`；`sdk/gormdb/migrate.go`：`type Migration struct { ID uint32; Up func(context.Context, *gorm.DB) error }`、`func Migrations(p *sdk.Plugin, list []Migration)`。事务中须使用传入的 tx，避免在单连接池上再独立取连接。
- `sdk/conversation.go`：`func AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error)`、`func ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error)`、`func ClearConversation(ctx context.Context, key string) error`。该存储属于 Core，与 Agent 的 SQLite 分属两处。
- `sdk/stream.go`：`func (c *Context) Stream(next func(context.Context) (string, error)) ([]*message.SendResult, error)`；编辑、节流、text_limit 与分段由 SDK 提供。`sdk/rule.go` 的 `func NotCommand() Rule`、`sdk/sdk.go` 的 `func (c *Context) CommandConsumed() bool` 与 proto `InvocationContext.command_consumed = 6` 已合入；沿用现有路由即可。
- `agent/agent_test.go` 的 agent_test 包已有 `startAgent(t *testing.T, loop *loopback, configuration map[string]any, coreExtra map[string]any, plugins ...*sdk.Plugin) *harness`、`newLoop(edit bool, limit uint64) *loopback`、harness 的 deliver、waitEntries 与只读 Core DB。loopback 当前没有媒体读取接口，deliver 只接文本，entries 解码为 map[string]string，需要按图片集成调整这些辅助。`sdk/adapter/adapter.go` 的可选接口为 `ReadPlatformFile(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error)`；能力为 media.read v1。
- 在途冲突：goal-40 涉及 state、supervisor 与根包装配；PR #171（goal-69）涉及根包和 cmd/alonabot；PR #172（goal-57）涉及根包 t11_e2e_test.go、pipeline_test.go。本切片的提交范围集中在 agent/，本地若借用根包 pipeline_test.go，则按其最新装配核对。goal-83 与本切片同改 agent.go、config_schema.json、i18n 和生成资源，后合并者同步 main、保留双方追加，再运行生成与检查。#165、#167、#168、#169、goal-59 已合并。

```text
/goal 使用 $alonabot-execute 推进 goal-136（切片 T-12.14：对话中的图片输入）。goal 目录 goal-136，分支 goal-136，worktree .worktrees/goal-136，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 的一轮对话读取入站图片，Provider 条目增加缺省 false 的 image_input；开启时本轮图片全部发送，历史图片按批大小 3 的边界回放，关闭时只留占位；图片字节与成功轮次的计数保存在 Agent 的受管 SQLite，Core 会话条目只保存文本与图片编号；删除越过边界的图片，/reset 清除图片并重新编号。

依据：spec/PLAN.md §3.1；spec/plan/agent.md §3 T-12 步骤 2、3、11 与「失败停止条件」；CONTRACTS §1.4、§2.4、§5.1、§6.1～§6.2、§13、§13.3；MESSAGE-MODEL §2、§4A（流显示）、§6.4（媒体来源）；WIRE-FORMAT §4、§5、§5A（既有入站媒体与会话协议）；DECISIONS D-49、D-60、D-76、D-86、D-109、D-121、D-132、D-135、D-141；TEST-MATRIX V-118 的对话分项，已有 V-57 的对话行为继续成立。

范围（只新增或修改这些）：
- agent/config_schema.json、agent/i18n/en-US.json、agent/i18n/zh-CN.json：所有六个 Provider 模板的 items 与 config 文案追加 image_input。
- agent/agent.go：configuredProvider 追加私有 imageInput bool 并读取；New 的受管迁移登记接入图片表和会话计数。
- 新增 agent/image.go：私有图片模型、读取、保存、回放边界、占位与清理；迁移所需模型归此文件。
- agent/conversation.go：conversationEntry 追加图片编号；在 converse 的历史/请求组装和成功轮次追加点接入图片处理，保留已有流和 reset 代次核对。
- agent/settings.go：chatSettings 追加会话图片计数，resetConversation 接入清理；已有聊天设置及其更新保留。
- agent/alonabot_gen.go：仅由现有生成器更新。
- 新增 agent/image_test.go（unix）的一个主路径集成测试；agent/agent_test.go 只修改共用回环 Adapter、事件注入和会话条目读取辅助以支持真实图片。其余测试放 agent/*_local_test.go；多进程自测可放根包 agent_image_local_test.go。
允许的新跨模块边：无；agent 产品代码沿用 sdk/...、gormdb 与 provider，agent_test 沿用根包库模式装配，不 import internal/core。根包本地自测可复用 pipeline_test.go 的生产装配。

接口（不新增导出符号；按既有签名接入）：
1. Agent 私有类型追加字段：
   configuredProvider.imageInput bool
   chatSettings.ImageCount uint64 `gorm:"not null;default:0"`
   conversationEntry.Images []uint64 `json:"images,omitempty"`
   保留 Role、Speaker、Content 及既有 JSON 键；Images 按该条 user 消息的到达顺序保存成功读取的图片编号。ImageCount 表示成功保存过的最大编号，删除旧字节后仍保留计数；/reset 归零。
2. 图片模型为 agent/image.go 的私有类型，主键为（会话键、图片编号），保存 MIME 与原始字节；用 gormdb 的受管库。图片模型与 chatSettings 的计数字段并入 New 中已有的初始迁移 ID 1（首次发布前不追加迁移；已运行旧初始迁移的开发数据库需重建），Up 使用传入的 tx 执行 AutoMigrate。迁移登记仍只有一次；保留 chat_settings 的所有原字段，不另建 Core 表或业务数据库连接。
3. image_input 的配置声明：每个 providers.templates.<预设>.items 增加 {"type":"bool","description":"Image input","default":false}；预设为 openai、anthropic、gemini、deepseek、ollama、custom。从 item.Bool("image_input") 读取到 configuredProvider；config 文案与模板同形追加。
4. 从 c.Event().Payload.(*message.CreatedPayload).Chain 中按顺序读取 message.ImageSegment 的 Source；读取调用 sdk.OpenMedia(c.Context(), source)，用 io.ReadAll 取得完整字节并关闭 reader。保留现有 converse(c *sdk.Context, text string) error、recentConversation、resetConversation、New 的签名。
5. 继续使用 provider.Image{MediaType: MIME, Data: bytes} 与 provider.Message.Images，调用现有 Client.Stream 和 c.Stream；Core 会话仍通过 AppendConversation/ListConversation/ClearConversation。proto 沿用 IncomingMedia.token = 1、mime = 3、readability = 4，MediaSource.incoming_ref = 2、ImageSegment.source = 1；无需改协议。

行为：
- 读取：本轮进入 converse 后，遍历原消息图片；Incoming 的 Readability 为 message.Readable 且 MIME 为 image/jpeg、image/png、image/gif、image/webp 之一时，用 OpenMedia 读取。不可读、MIME 不支持或读取失败的图片各留一个 [image] 占位，WARN 日志带会话键和实际原因，省略图片内容和凭证，这一轮继续。读取与关闭的错误保留；不改为直接下载 URL 或使用平台文件标识。image_input 为 false 的 Provider 也读取并保存合格图片，供之后切换 Provider 使用。
- 会话：键为 string(c.Session())；群聊中不同用户共用该聊天的计数，Speaker 继续记录发言人的外部 ID。入轮次的设置、历史、图片与计数快照应与 /reset 共用既有 sessionLock，读取完快照即释放，Provider 请求沿用本轮 Context。
- 回放：批大小 N 为命名常量 3；设本轮之前成功保存的最大编号为 m，m=0 时无历史图片，否则边界 b=((m-1)/N)*N。image_input 开启时，本轮成功读取的所有图片都放入当前 user 的 Images；历史 user 的编号大于 b 且字节仍保存时，按条目内编号顺序放回该条消息的 Images；其余历史图片逐张在所在条目文本前增加一行 [image]。关闭时，本轮和历史图片均只加占位，Images 为空。占位 [image] 用命名常量。
- 历史条目：Core 保存 role、speaker、content 和 images 数字数组，保持原有文本及群聊发言人信息。不可读图片的占位保存在条目文本中；成功读取图片的临时占位按本次所选 Provider 与回放结果生成，避免下一轮重复叠加。Core 条目保留越过窗口的旧编号，以便重放时产生占位；原始图片、base64 和媒体读取 token 均由 Agent 自己处理。
- 保存：生成与显示均成功、且最终持锁核对的 ResetGeneration 与入轮次一致时才追加本轮。在最终锁内读取最新 ImageCount，为本轮成功读取的图片连续分配编号，避免两个并发轮次复用编号；user、assistant 仍用一次 AppendConversation 追加。图片行、计数和越界删除在同一个 Agent DB 事务中写入；AppendConversation 返回错误时回滚该本地事务并保留错误。本轮 Provider、显示或读取以外的存储操作失败时返回实际错误；Provider/显示失败时不保存本轮图片、不推进计数，错误显示沿用当前逻辑。
- 清理：保存后按新的计数 m' 算出 b'=((m'-1)/N)*N，删除该会话编号 <= b' 的图片字节；m'=0 时无图片可删。清理与本轮图片/计数写入同事务；持久保存的图片最多 3 张，本轮发送可多于 3 张。一条消息带多图时全部交给开启图片的 Provider，保存后仍按边界清理。
- /reset：沿用同一会话锁及 ResetGeneration 递增机制；ClearConversation 成功后，在 Agent 受管 DB 事务内删除该会话全部图片并把 ImageCount 置零，保留 Provider、Persona、Enabled。旧轮次即使稍后完成也按代次跳过追加和图片保存；下一轮从 1 编号。清理错误完整返回，不静默报告已重置。
- 跨库的失败（Core 会话与 Agent 图片在两个库，不做跨库事务）：每张图片记录它所属的 Core 会话条目序号；回放只取 Core 会话中仍存在的条目对应的图片，其余忽略，并在下一次 /reset 时清理。Core 追加成功而 Agent 写入图片失败：本轮保留文字历史，图片不回放，以 WARN 记录；/reset 中 ClearConversation 成功而删除图片失败：向用户返回错误，残留图片因对应条目已不存在而不会回放；管理服务直接清除 Core 会话时同理，残留图片在下一次 /reset 时清理。
- 流显示、命令权限、自由对话开关、ToMe/NotCommand 规则、群聊 Speaker 和错误文案沿用 #170 的实现；超长输出继续交给 c.Stream 按实例的 text_limit 处理。
- 任一 Core 会话条目写入图片字节，或成功清理后一个会话仍持有超过窗口的图片，按 T-12 的失败停止条件阻止合入。Agent 只经 SDK 访问 Core。

测试（按编码规范「测试提交规则」）：
- 提交：agent/image_test.go（//go:build unix）的一个 TestAgentImages 集成测试。复用 agent/agent_test.go 的 Start、真实 SDK Session、Core 和 Agent 的 SQLite 及 httptest Provider；回环实例声明 media.read v1，实现 sdk/adapter.MediaSource 的 ReadPlatformFile，按平台文件 ID 返回测试字节。经真实 Adapter.Core.Deliver 交付 image 片段（IncomingMedia 的 Token 在 Adapter 输入端为平台文件 ID，带 MIME 和可读标记，由 Core 登记并签发授权），由真实 OpenMedia 取件，避免直接注入 Agent 图片存储。
  - 配置开启 image_input 的 openai Provider；同一会话依次交付五条各含一张 PNG 的自由对话消息（先由 Admin /agent on，消息带 ToMe；复用当前权限名集合）。逐轮等到会话追加完成再交付下一轮，验证请求图片数依次为 1、2、3、4、2；第 5 轮只发送编号 4 与本轮 5 的字节，编号 1～3 所在条目有占位。
  - 只读查询 Agent 受管 DB，验证第 5 轮后仅保存编号 4、5，计数为 5；Core 条目含数字编号及原文本/群聊 Speaker，图片字节、base64 和令牌均未写入会话条目。
  - /reset 后 Core 历史与 Agent 图片都为空、计数为 0；下一条带图消息的请求只有本轮图片，成功后编号为 1。
  - 共用 harness 的条目读取从 map[string]string 调整为能保留数字数组的表示，已有 TestAgentConversation 的文本断言继续成立；不另增重复的测试函数。
- 本地自测（*_local_test.go，不提交，PR 写明命令、平台和结果）：V-118 的 m=0～7 边界（b 依次 0、0、0、0、3、3、3、6，其中 m=0 单独处理）；关闭 image_input 只发送占位但照常保存，再经 /provider 切到开启的 Provider 后回放；一条消息 5 张图片全部发送、保存后最多 3 张；只有图片无文本的自由对话；不可读、未知/不支持 MIME、OpenMedia 打开或后续读取失败；Provider 非 2xx/断流与显示失败时图片/计数不变；不同会话隔离、群聊多用户共用计数；同一会话并发成功轮次编号不重复；/reset 与仍在生成的轮次交错后旧轮次无残留；Agent 重启后图片与计数仍可回放；历史条数为 0 时不回放历史；Core 追加失败时本地事务回滚，图片清理失败完整报错；anthropic-messages、gemini-generate-content 各一次带历史图片的请求。根包 agent_image_local_test.go 可复用 pipeline_test.go 的生产装配运行一次真实多进程媒体读取，测试源码与入口留在本地，无需修改共享流水线文件或新增 testdata。

- 本地自测（不提交）：新库经初始迁移建立图片表与计数字段，再次启动跳过已完成迁移；Agent 图片写入失败时文字历史保留、图片不回放；删除图片失败时 /reset 返回错误且残留图片不回放；管理服务清除 Core 会话后旧图片不回放，下一次 /reset 清理。

完成条件：上述图片主路径与既有 TestAgentConversation 通过；gofmt -l 对本次 Go 文件无输出；go vet ./...；go test -race ./agent/...；go build ./...；go run ./cmd/alonabot plugin check --dir ./agent 通过。用 go generate ./agent 更新生成资源，随后 check 应显示资源最新。PR 对应 V-118 记录测试位置、命令、平台和结果，将模拟 Provider 与真实 Provider 的证据分开；达到 T-12 的图片存储与回放停止条件要求。

不做：Hermes text 模式；图片压缩、缩放和额外张数/大小限制；工具结果及 llm.generate 的图片输入；文件、语音、视频输入；Provider 六种协议的编码修改；Core 会话协议或图片字节存储；候选 Journal、管理 API 和热部署。
```

## 第 15 波

### goal-60 · T-12.3 工具循环与 MCP

状态：已完成（[PR #178](https://github.com/AlonaBot/AlonaBot/pull/178)，合并提交 `a95b85b`），证据见 TEST-MATRIX §3.2。合入：官方 Agent 的工具循环（`agent/tools.go`、`agent/loop.go`）与 MCP 客户端（`agent/mcp.go`，stdio 与 streamable HTTP，配置 `mcp_servers`）；Core 工具经 `sdk.InvokeTool`，MCP 工具经 `sdk.ConfirmOperation` 确认后执行，单轮最多 20 次；`github.com/modelcontextprotocol/go-sdk` v1.3.1 锁定为 agent 的直接依赖。以下为派发时的语句。

已有接口（审查者按 main 核对）：

- `agent/agent.go`：`func New() *sdk.Plugin`；内部对象为 `component`，配置为未导出的 `configuration`；`func readConfiguration(values config.Values) (configuration, error)` 在 `p.OnStart(func(context.Context) error)` 中调用。当前配置只有 providers、personas、defaultProvider、defaultPersona、historyEntries，MCP 配置由本切片加入。`sdk/plugin.go` 已有 `func (p *Plugin) OnStop(fn func(context.Context) error)`；应在同一组件对象上管理 MCP 资源。
- `agent/conversation.go`：`func (a *component) converse(c *sdk.Context, text string) error`、`func recentConversation(ctx context.Context, key string, count int) ([]conversationEntry, error)`；`conversationEntry` 只有 `Role, Speaker, Content string`。当前通过 `c.Continue` 执行对话、`iter.Pull2` 衔接一次 Provider 请求与 `c.Stream`，Provider 的 ToolCall 尚未处理。`agent/settings.go` 的 `ResetGeneration`、`sessionLock` 已保护 reset 与成功轮次追加；接入循环须保留这些条件。
- `sdk/tool.go`：`ListTools(ctx context.Context) ([]ToolInfo, error)`；`ToolInfo{Name, Description string; InputSchema json.RawMessage; Policy string}`，Policy 为 `allow` 或 `confirm`；`InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error)`；`ConfirmOperation(ctx context.Context, operation, summary string) error`。后两者绑定活跃 Handler/延续调用的父调用；InvokeTool 可同时返回结果与错误。
- `agent/provider/provider.go`：`New(config Config, httpClient *http.Client) (*Client, error)`、`(*Client).Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)`。六种协议均已存在；`ToolCall` 的 `Replay json.RawMessage` 原样保留于本轮工具循环，D-103 要求其生命周期止于本轮。接口段列出实际载荷类型。
- `sdk/sdk.go`、`sdk/continuation.go`、`sdk/stream.go`、`sdk/rule.go`：`(*Context).Context() context.Context`、`Continue(func(*Context) error) error`、`Stream(func(context.Context) (string, error)) ([]*message.SendResult, error)`、`Event() *message.Event`、`Session() message.Session`、`CommandConsumed() bool`、`NotCommand() Rule`。自由对话仍用 `sdk.ToMe(), sdk.NotCommand()`（D-141）。Send 接受 `message.Content`，平台权限为 `message.Actor.Permissions []string`。
- `sdk/config/values.go`：`Values.Objects(key string) []Values`、`String(key string) string`、`Strings(key string) []string`；secret 的 String getter 取得已解析明文。secret 只适用于 string/text，MCP env、headers 应使用含 secret value 的对象列表，沿用 `docs/plugin-guide.md` §7，避免把整个 dict 标为 secret。
- `go.mod` 当前没有 MCP 的直接 require；`go.sum` 已有 `github.com/modelcontextprotocol/go-sdk v1.3.1`，本机该版本源码与 anthropic-sdk-go/v2 v2.9.0 的 go.mod 均可读，后者要求 MCP v1.3.1。MCP 的 `NewClient`、`Client.Connect`、`ClientSession.ListTools/CallTool/Close`、`CommandTransport` 与 `StreamableClientTransport` 已按缓存源码核对，签名见接口段；正式版本在执行时验证并锁定，最低 v1.3.1。
- `agent/agent_test.go`（外部包 `agent_test`，unix）：`startAgent(t *testing.T, loop *loopback, configuration map[string]any, coreExtra map[string]any, plugins ...*sdk.Plugin) *harness` 已经用 `alonabot.Start`、`agent.New()`、真实 Session、临时 SQLite 与回环 Adapter 验证对话，可直接复用。多进程本地补证据可复用根包 `pipeline_test.go` 的 `newPipeline(t *testing.T, stdio ...bool) *pipelineHarness`；其 `TestMain` 已有 tasks 模式分流。测试插件用 `sdk.NewPlugin`、`sdk.UseGenerated`、`sdk.Serve(p, sdk.JSON[In, Out](target)).Access(...).Tool(...).Handle(...)`；`ServiceBuilder` 的公开方法为 Access、Tool、Handle。
- `alonabot.go` 的 `Start(ctx context.Context, opts ...Option) (*Core, error)`、`Plugin(root *sdk.Plugin) Option` 已接受 Agent；`internal/core/manage/static.go` 的 `(*Service).WebUI() http.Handler` 已合入，管理端点以 `spec/ADMIN-API.md` 为准。本切片使用 SDK 工具接口，测试策略通过既有 Core 工具服务层或临时状态准备设置。
- 在途共享文件：goal-40 涉及 `internal/core/state/`、`internal/core/supervisor/`、`assembly.go` 等根包装配；PR #171（goal-69）涉及根包与 `cmd/alonabot/`；PR #172（goal-57）涉及根包 `t11_e2e_test.go`、`pipeline_test.go`。本切片仅本地多进程自测可能接触根包测试辅助，须读取最新分流后追加；正式测试放在 Agent 目录。goal-61 与本切片共享 `agent/{agent,conversation}.go`、生成物及 i18n；后合并者保留双方接入点。旧提示中的 #165/#167/#168/#169、goal-59 均按已合并处理，#169 的 `state/schema.sql` 与清理清单按现状复用。

```text
/goal 使用 $alonabot-execute 推进 goal-60（切片 T-12.3：工具循环与 MCP）。goal 目录 goal-60，分支 goal-60，worktree .worktrees/goal-60，从最新 origin/main 创建（须已包含 PR #101、#165、#170）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在官方 Agent 的每轮对话中接入 Core 工具和 MCP tools，依次执行模型返回的工具调用并把结果交回模型；按六种 Provider 协议映射工具名，保持本轮映射和 Replay；执行 Core 工具策略与 MCP 操作确认，达到单轮工具步数上限时结束本轮并告知用户；MCP 懒连接，插件停止时关闭会话并回收 stdio 子进程。

依据：spec/PLAN.md §3.1、§5；spec/plan/agent.md T-12.3、T-12 步骤 2（MCP 配置）、3（对话显示）、4，以及「失败停止条件」；spec/plan/core.md T-11 步骤 1～4 与「失败停止条件」；spec/CONTRACTS-v0.1.md §2.4、§3、§4.1～§4.2、§5.1、§5.3～§5.4、§7.1、§13.1～§13.3；spec/MESSAGE-MODEL.md §1、§2（按钮）、§3、§4A、§6.3；spec/WIRE-FORMAT.md §3.3、§3.5、§4；spec/adr/DECISIONS.md D-27、D-38（按 D-132 修订）、D-86、D-98、D-103、D-132、D-135、D-141；docs/development.md 的已批准依赖范围；TEST-MATRIX V-58、§3，保留 V-57 的已实现对话行为。

范围（只新增或修改这些）：
- go.mod、go.sum：把 github.com/modelcontextprotocol/go-sdk 显式列为直接依赖，版本不低于 v1.3.1，执行验证后锁定。
- agent/agent.go：内部 configuration 增加 MCP 配置，component 持有连接；在已有 OnStart 中读取配置，在 New 的注册阶段另登记一个 OnStop 关闭资源；保留 agent.New 的签名与既有注册。
- agent/config_schema.json：追加 mcp_servers；agent/i18n/en-US.json、agent/i18n/zh-CN.json：相应配置翻译、agent.tool_limit 及必要的用户错误文案。
- 新增 agent/tools.go、agent/mcp.go、agent/loop.go：目录与名称映射、MCP 管理、工具循环；agent/conversation.go 只修改请求生成/流处理的接入处，保留历史提交与 reset 代次核对。
- 新增 agent/mcptest/main.go：仅供测试启动的 stdio MCP 模拟服务器（命令包）；streamable HTTP 模拟服务放在测试辅助中。
- 新增 agent/tools_test.go（//go:build unix，package agent_test）；agent/agent_test.go 只补充共用测试 Adapter 的确认按钮能力、点击交付和生命周期辅助，保留原测试断言。
- agent/alonabot_gen.go：仅通过生成器更新上述源码与资源对应的生成内容。
- 本地自测文件：agent/*_local_test.go；如需独立 Agent 进程与进程树证据，在根包 *_local_test.go 复用 pipeline_test.go 的装配与测试二进制，不改根包被跟踪的测试辅助。
允许的新跨模块边：agent/... -> github.com/modelcontextprotocol/go-sdk/mcp（已批准）；其余沿用 agent/... -> sdk/...、contracts 的既有边。测试复用 agent/agent_test.go 已有的外部测试包 -> 根包库模式入口；产品代码只经 SDK 与 Core 通信。

接口（已有公开符号照此调用，Agent 新增类型和辅助函数均不导出）：
// agent
func New() *sdk.Plugin
// agent/provider 的现有类型与方法
type Tool struct { Name, Description string; InputSchema json.RawMessage }
type ToolCall struct { ID, Name string; Arguments, Replay json.RawMessage }
type ToolResult struct { ToolCallID, Content string }
type Message struct { Role, Content string; ToolCalls []ToolCall; ToolResults []ToolResult; Images []Image }
type Request struct { Messages []Message; Tools []Tool; MaxTokens int }
type Event struct { Text string; ToolCall *ToolCall }
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
// sdk
func ListTools(ctx context.Context) ([]ToolInfo, error)
func InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error)
func ConfirmOperation(ctx context.Context, operation, summary string) error
func (c *Context) Continue(fn func(*Context) error) error
func (c *Context) Stream(next func(context.Context) (string, error)) ([]*message.SendResult, error)
func (p *Plugin) OnStop(fn func(context.Context) error)

// 新增私有配置的字段含义，放进现有 configuration；名称可按周围代码整理。
type secretPair struct { Name, Value string }
type mcpServer struct {
    Name, Transport, Command, URL string
    Args []string
    Env, Headers []secretPair
}
// mcp_servers：objects，name、transport、command、args、env、url、headers。
// transport 为 stdio 或 streamable-http；env/headers 为 name、value 对象列表，value 为 secret string。
// stdio 使用 command/args/env；streamable-http 使用 url/headers；名称须在配置中唯一。
// 由 Config().Objects/String/Strings 读取，校验本传输实际使用的字段，秘密值来自已解析配置。

// github.com/modelcontextprotocol/go-sdk/mcp，v1.3.1 已核对的入口；以下按 mcp 包内签名列出。
// 采用其他版本时先核对实际签名。
func NewClient(impl *Implementation, options *ClientOptions) *Client
func (c *Client) Connect(ctx context.Context, t Transport, opts *ClientSessionOptions) (*ClientSession, error)
func (cs *ClientSession) ListTools(ctx context.Context, params *ListToolsParams) (*ListToolsResult, error)
func (cs *ClientSession) CallTool(ctx context.Context, params *CallToolParams) (*CallToolResult, error)
func (cs *ClientSession) Close() error
// mcp.CommandTransport{Command *exec.Cmd, TerminateDuration time.Duration}
// mcp.StreamableClientTransport{Endpoint string, HTTPClient *http.Client, MaxRetries int, ...}

协议复用（字段保持现状，无新增 proto）：
ToolInvokeRequest{name = 1, arguments_json = 2} -> ToolInvokeResponse{result_json = 1}；
ToolConfirmRequest{operation = 1, summary = 2} -> ToolConfirmResponse{}；
两次请求均通过 sdk 关联活跃父调用。结果和 ErrorInfo 可以同时存在。

行为：
- 每轮开始经 sdk.ListTools(c.Context()) 取得 Core 目录；按配置对尚未连接的 MCP 服务器首次连接，取得 tools/list 的全部分页，与 Core 目录合并。本轮固定映射和来源信息，随后请求复用；连接失败的服务器本轮不可用，脱敏 WARN 后其余工具与对话继续。
- 工具的来源名为 plugin_id/service_name 或 mcp/<server>/<tool>。映射为六种协议共同接受的名称：首字符为 ASCII 字母或下划线，后续只用 ASCII 字母、数字、下划线，长度至多 64；非法字符替换为下划线，首字符和超长/冲突按命名常量及稳定后缀处理，维护可逆的本轮查表。后缀算法使用标准库，保留每个工具的唯一映射。原始名用于执行与确认，模型侧名称随 ToolCall 原样回放。
- 整轮仍在已有延续调用中运行，使用其 context、Actor、Target、deadline 与取消。Provider 返回的文本增量继续交给同一次 c.Stream；每次响应的文本与完整 ToolCall 组成 assistant 消息，ToolCall（含 ID、Name、Arguments、Replay）原样放回下一请求；工具结果组成 user 消息的 ToolResults，ToolCallID 精确对应调用 ID，再继续请求模型。最终无工具调用时结束。会话仅按既有成功提交规则保存本轮 user/assistant 文本；ToolCall、Replay 与 ToolResult 只保留在本轮请求链。
- Core 工具只调用 sdk.InvokeTool。成功返回的 JSON 原文作为工具结果；失败保留 ErrorInfo 的 code、safe_message，编码为 error: <code>: <safe_message>，有同时返回的结果时一并交回模型。依次执行同一响应里的工具调用；不自动重试，不直接调用目标 Service。
- MCP 每次执行前调用 sdk.ConfirmOperation(c.Context(), "mcp/<server>/<tool>", 参数 JSON 摘要)，摘要完整表达实际参数。成功后才 CallTool；拒绝、超时和工具失败都交回模型。服务器校验工具业务 Schema；Agent 不另写一套 Schema 校验。MCP 返回的 content、structuredContent、isError 保留为可供 ToolResult.Content 携带的文本表示，禁止悄悄丢掉错误或非文本块。
- 单轮工具步数上限使用 Agent 的命名常量，初值 20，按实际执行的工具调用次数计（同一响应中的多个 ToolCall 各计一次，Core 工具与 MCP 工具合计）；模型请求第 21 次工具调用时不执行，停止继续工具与模型请求，按 Actor locale 回复 agent.tool_limit。父调用取消或结束时立即终止本轮，保留已有平台回执与显示内容；不继续请求模型处理已经失效的调用。
- stdio 使用 exec.Cmd 的 command/args/env；凭证只来自本服务器配置，不把 Core 管理凭证或其他服务器凭证传入。子进程继承插件的受管进程树。streamable HTTP 使用注入 headers 的 HTTP client；禁止自动重放 tools/call，核对 go-sdk 自带重连行为并关闭会导致自动重发的设置。
- 连接可跨轮复用，生存期归插件 Session；本轮 context 只控制本轮请求。OnStop 关闭全部 MCP 会话、等待 stdio 进程退出，服从 CONTRACTS §4.2 的剩余停止预算并保留关闭错误。v1.3.1 CommandTransport.Close 默认每阶段等待 5s，接入时须按实际剩余停止预算处理，避免直接套用该默认值。
- env、headers、认证头与含凭证的 URL 不进入日志、用户错误或工具结果；保留脱敏后的错误类别与说明。沿用 Provider 不重发、显示失败立即结束、reset 之前的轮次跳过历史追加、NotCommand 过滤以及配置为空时提示管理员的现状。

测试（按编码规范「测试提交规则」）：
- 提交一个集成测试 agent/tools_test.go 的 TestAgentTools（unix，表驱动）。复用 startAgent 的真实 Core、Agent SDK Session、受管 SQLite 与回环 Adapter；工具提供方使用真实 sdk.Plugin、带类型 JSON 声明与 Generated.ServiceSchemas；Provider 用 httptest 脚本，MCP 用 stdio 真实子进程和 streamable HTTP 模拟服务器。主路径让模型先调用 Core 工具、再调用 MCP 工具、最后输出文本；分别覆盖两种 MCP 传输，核对确认的实际参数、同一 Actor 允许、执行次数、ToolCallID、后续请求中的结果与最终显示。再覆盖确认拒绝及工具错误交回模型且目标执行次数正确；停止 Agent 后核对 stdio 子进程退出。多个子场景理由：V-58 明列两种传输、确认拒绝与工具错误，且关闭子进程是本切片的资源交付条件。
- 原 TestAgentConversation 只按必要的辅助签名改写，原断言保持；不向 sdk/、internal/core/supervisor/ 提交测试，不新增 testdata 插件。
- 本地自测（*_local_test.go，不提交，PR 写明实际结果）：六种协议收到映射后的目录并完成工具往返；Responses 的 reasoning Replay、Gemini thoughtSignature 原样回传；非法字符、首字符、超长与同名碰撞；同响应多个工具的顺序与结果对应；Core/MCP 的 allow、confirm、deny、确认超时、无 Actor、无按钮、父调用取消；结果与错误同时返回；Provider/工具失败的请求次数；持续调用达到 20 次（同一响应含多个 ToolCall 时跨过上限的那一次不执行）；首次连接失败时其余对话继续；tools/list 分页；headers/env 的到达与日志/结果脱敏；并发首次使用同服务器、跨轮复用及停止预算。独立 Agent 进程的 stdio 子进程树退出证据可在根包本地测试复用 pipeline 装配。
- 经 cpa 兼容网关的真实工具往返由审查者在合并前补跑（凭证只在审查者本机）。
- 证据记入 goal-60/tasks.md 与 PR，按 TEST-MATRIX §3 写明位置、命令、OS/arch、Go 版本、传输、结果；T-12.7 负责最终汇总，真实 Provider/平台未运行项如实列出。

完成条件：生成资源前后对应源码一致；gofmt -l 对改动的 Go 文件无输出；go vet ./agent/...；go test -race -count=1 ./agent/...；go build ./...；在 agent 目录运行 go run ../cmd/alonabot plugin check 通过。依赖验证运行 go mod tidy、go mod why github.com/modelcontextprotocol/go-sdk/mcp，并核对 Core 与 SDK 根包的 go list -deps 未链接 MCP 包；在 PR 记明实际版本。通过上述检查并落实 V-58 的指定证据后完成；未经策略/确认执行、自动重放业务请求、丢失必需 Replay、子进程脱离管理或 Agent 产品代码访问 internal/core 时按工作包停止。

不做：记忆实现（T-12.4）、技能（T-12.5）、动态任务 SDK 与定时 Agent（T-07.9、T-12.6）、llm.generate、图片输入、MCP 资源与提示词、真实 Provider 汇总（T-10）、管理页面、Core/SDK/proto 改造、根包产品装配修改、SHA256。
```

### goal-61 · T-12.4 Agent 的记忆

状态：已完成（[PR #182](https://github.com/AlonaBot/AlonaBot/pull/182)，合并提交 `90b6817`），证据见 TEST-MATRIX §3.2（含经 cpa 兼容网关运行的 openai-responses 与 gemini-generate-content 实测）。合入：agent/memory.go 的 `memory_record`（会话笔记按 `<会话字符串>@<实例名>`、用户档案按 Actor）与 memory 工具（AccessPrivate、缺省 allow，经 core.tool.invoke 自身调用，add/patch/remove 原子执行，上限 2,200 与 1,375 码点）；会话首轮冻结笔记快照（`conversationEntry.Kind` 区分），/reset 后重取，每轮只注入当前 Actor 的档案；agent 的会话键 `sessionKey(c)`；`sdk.(*Context).Instance() string`、`sdk.(*ServiceBuilder).Permission(p)`；Broker 对自身工具调用不判 CALL_CYCLE、Service 的 required_permissions 在通用准入检查、深度记在 `router.Invocation.Depth`；CONTRACTS §2.3、§3、§13.4 同步 D-142。以下为派发时的语句。

已有接口（审查者按 main 核对）：

- `agent/agent.go`：`func New() *sdk.Plugin`；通过 `sdk.NewPlugin()`、`sdk.UseGenerated(p, alonabotGenerated)` 注册；用 `gormdb.Migrations(p, []gormdb.Migration{{ID: 1, Up: ...}})` 创建 chat_settings。现有配置为未导出的 configuration，当前没有 memory 注册。保留一个根插件对象与现有生命周期回调。
- `sdk/gormdb/open.go`：`func Open(ctx context.Context) (*gorm.DB, error)`；返回当前 Session 的受管连接包装，SDK 负责关闭。`sdk/gormdb/migrate.go`：`type Migration struct { ID uint32; Up func(context.Context, *gorm.DB) error }`、`func Migrations(p *sdk.Plugin, list []Migration)`；事务里使用传入的 tx，单连接事务中再次独立 Open/查询会等待自身。
- `sdk/services.go`：`JSON[In, Out any](target string) Service[In, Out]`、`Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out]`；builder 的 `Access(a Access, allowlist ...string)`、`Tool(description string, defaultPolicy ToolDefault)` 返回同一 builder，`Handle(func(*Context, In) (Out, error))` 结束注册；使用现行 `sdk.ToolAllow`。旧名 JSONService 已删除。JSON Schema 由生成器根据包级具名值类型生成，默认版本 1；不要求业务 DTO 导出。
- `sdk/tool.go`：`ListTools(ctx context.Context) ([]ToolInfo, error)`、`InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error)`、`ConfirmOperation(ctx context.Context, operation, summary string) error`；按 Core 的工具策略执行并继承活跃父调用。`contracts/proto/alonabot/v1/route.proto` 已有 `RouteDescriptor.required_permissions = 17`，公开 ServiceBuilder 只有 Access、Tool、Handle，尚无声明 required_permissions 的入口。
- `sdk/conversation.go`：`AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error)`；`ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error)`；`ClearConversation(ctx context.Context, key string) error`；`ConversationEntry{Seq uint64; Entry json.RawMessage}`，list 需处理分页。
- `agent/conversation.go`：`func (a *component) converse(c *sdk.Context, text string) error`；`func recentConversation(ctx context.Context, key string, count int) ([]conversationEntry, error)`；私有 conversationEntry 只有 Role、Speaker、Content 字符串。recentConversation 在 history_entries=0 时直接返回；其余仅保留最近 N 条。快照初始化与读取须独立识别真实空会话并找回既存快照，保证 history_entries 截取后仍使用冻结值。
- `agent/settings.go`：`func (a *component) resetConversation(c *sdk.Context) error` 以 `sessionLock(string(c.Session()))` 串行化 ResetGeneration 递增及 ClearConversation；converse 成功追加前核对相同代次。现有会话/设置键为 `string(c.Session())`。
- `sdk/sdk.go`：`(*Context).Event() *message.Event`、`User() message.User`、`Session() message.Session`、`Caller() string`（后者位于 services.go）；Session/用户字符串都省略实例。事件 Handler 可以从 `Event().Actor`、`Event().AdapterInstanceID` 读取实例；Service 的 invokeService 只填私有 actor/defaultTarget，未填 event。`sdk/message/session.go` 的 `Session.Parts() (adapterID, chatID, threadID string, ok bool)` 提供聊天与话题，公开 Context 当前没有 Service 侧 Actor/defaultTarget 的完整读取入口。
- `internal/core/broker/broker.go` 的 call：复制父链后追加 c.PluginID，只要目标 pluginID 已在链中便返回 CALL_CYCLE；`internal/core/tool/invoke.go` 的 invoke 最终仍转交同一个 Broker。Agent 延续调用经 InvokeTool 调用自己的 memory，会命中此规则；本稿保留 CONTRACTS §13.4 的规定路径，交审查者修定契约或前置能力。
- `agent/agent_test.go`（unix，外部包 agent_test）的 `startAgent(t *testing.T, loop *loopback, configuration map[string]any, coreExtra map[string]any, plugins ...*sdk.Plugin) *harness` 经根包 Start 运行真实 Core、Agent Session、SQLite 与回环 Adapter，可复用为记忆集成装配；确认按钮能力需按实际合入顺序与 goal-60 合并。多进程本地自测可参考根包 `pipeline_test.go` 的 `newPipeline(t *testing.T, stdio ...bool) *pipelineHarness`。
- 在途共享文件：goal-40 触及 state、supervisor 与根包装配；PR #171（goal-69）触及根包与 cmd/alonabot；PR #172（goal-57）触及 t11_e2e_test.go 与 pipeline_test.go。本切片产品范围均在 Agent；若审查者另派 Core/SDK 前置修复，应与这些所有者协调。goal-60 与本切片共享 agent.go、conversation.go、i18n、生成物及测试辅助；模型写记忆的联调依赖其工具循环。旧在途清单中的 #165/#167/#168/#169、goal-59 已合并，#169 的 schema 与卸载清理现状保持由 Core 所有者维护。

```text
/goal 使用 $alonabot-execute 推进 goal-61（切片 T-12.4：Agent 的记忆）。goal 目录 goal-61，分支 goal-61，worktree .worktrees/goal-61，从最新 origin/main 创建（须已包含 PR #79、#101、#35、#165、#170 与 goal-60 的 PR #178）。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在官方 Agent 的受管 SQLite 中保存按可信默认 Target 划分的会话笔记与按 Actor 划分的用户档案，注册 memory 工具并按 Core 策略执行原子的 add/patch/remove；会话第一轮冻结笔记快照，此后沿用该快照，/reset 后重新取；每轮只注入当前 Actor 的档案，群聊用户互相隔离。

依据：spec/PLAN.md §3.1、§5；spec/plan/agent.md T-12.4、T-12 步骤 3、5 与「失败停止条件」；spec/plan/core.md T-11 步骤 1～5 与「失败停止条件」；spec/CONTRACTS-v0.1.md §2.4、§3、§6.1～§6.2、§13.1～§13.4；spec/MESSAGE-MODEL.md §1（Actor、Target 与实例）、§3、§4A；spec/WIRE-FORMAT.md §3.3、§3.5、§4；spec/DESIGN.md §R 的记忆；spec/adr/DECISIONS.md D-60、D-86、D-103（工具 Replay 只留本轮）、D-122（带类型声明）、D-135、D-141；TEST-MATRIX V-59、§3，保留已实现的 V-57 对话行为。

范围（只新增或修改这些）：
- 新增 agent/memory.go：私有记忆数据类型、GORM 存储、原子操作、memory 工具注册、读取与请求注入辅助。
- agent/agent.go：在现有 gormdb.Migrations 登记中加入记忆表迁移，并在同一个 p 上注册 memory 工具；保留已有聊天设置与对话注册。
- agent/conversation.go：conversationEntry 追加快照 kind；接入会话首轮快照、历史中的固定快照读取与本轮 Actor 档案注入，保留成功追加及 reset 代次检查。
- agent/settings.go：快照的初始化、清除与 reset 代次共享现有锁；聊天设置、reset 锁与会话历史的键改为 `<会话字符串>@<实例名>`（D-142，首次发布前不写兼容层，已运行旧键的开发数据库需重建）。
- internal/core/broker/、internal/core/tool/invoke.go：只供 core.tool.invoke 使用的自身工具调用放行（D-142），不新增 SDK 公开符号。
- sdk/sdk.go：新增 `(*Context).Instance() string`；sdk/services.go：新增 `(*ServiceBuilder[In, Out]).Permission(p Permission)`，写入 RouteDescriptor.required_permissions。
- agent/i18n/en-US.json、agent/i18n/zh-CN.json：增加 agent.memory_error，保持消息占位符一致。
- 新增 agent/memory_test.go（//go:build unix，package agent_test）；agent/agent_test.go 只补必要的共用装配、确认点击与会话条目读取辅助，原断言随新增快照条目做精确适配，保留原有对话、显示和失败条件。
- agent/alonabot_gen.go：仅由生成器更新，包含 memory 服务输入/输出 Schema 与 i18n。
- 本地自测为 agent/*_local_test.go；若需多进程补证据，根包 *_local_test.go 复用已有 pipeline 装配，不改被跟踪的根包测试文件。
允许的新跨模块边：无；沿用 agent/... -> sdk/...、contracts 与已批准的 GORM，internal/core 内部与 sdk 内部的现有边。外部测试包沿用 agent/agent_test.go 已有的根包库模式装配。

接口（现有公开符号照此调用；Agent 不新增导出符号；SDK 只新增下面两个）：
// sdk（D-142 新增）
func (c *Context) Instance() string // 当前事件 Target 或继承的默认 Target 的 adapter_instance_id；都没有时为空
func (s *ServiceBuilder[In, Out]) Permission(p Permission) *ServiceBuilder[In, Out] // 与 CommandBuilder.Permission 同语义：满足任一权限名即可，未调用时不要求
// internal/core/broker 与 internal/core/tool：私有放行方式自定（例如 broker 调用参数中的标志），只由 core.tool.invoke 设置
// agent
func New() *sdk.Plugin
// sdk/gormdb，以下按 gormdb 包内签名列出。
func Open(ctx context.Context) (*gorm.DB, error)
type Migration struct { ID uint32; Up func(context.Context, *gorm.DB) error }
func Migrations(p *sdk.Plugin, list []Migration)
// sdk，以下按 sdk 包内签名列出。
func JSON[In, Out any](target string) Service[In, Out]
func Serve[In, Out any](owner Registrar, declaration Service[In, Out]) *ServiceBuilder[In, Out]
func (s *ServiceBuilder[In, Out]) Access(a Access, allowlist ...string) *ServiceBuilder[In, Out]
func (s *ServiceBuilder[In, Out]) Tool(description string, defaultPolicy ToolDefault) *ServiceBuilder[In, Out]
func (s *ServiceBuilder[In, Out]) Handle(h func(*Context, In) (Out, error))
func InvokeTool(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error)
func AppendConversation(ctx context.Context, key string, entries ...json.RawMessage) (lastSeq uint64, err error)
func ListConversation(ctx context.Context, key string, fromSeq uint64) (entries []ConversationEntry, nextSeq uint64, more bool, err error)
func ClearConversation(ctx context.Context, key string) error

// 拟新增的私有 JSON DTO，输入/输出 Schema 由现有生成器产生，默认版本 1。
type memoryOperation struct {
    Op string `json:"op"`             // add、patch、remove
    Target string `json:"target"`     // memory、user
    Text string `json:"text,omitempty"`
    OldText string `json:"old_text,omitempty"`
}
type memoryRequest struct { Operations []memoryOperation `json:"operations"` }
type memoryResponse struct{}
// 注册目标与调用名均为 alonabot-agent-official/memory：
// sdk.Serve(p, sdk.JSON[memoryRequest, memoryResponse]("alonabot-agent-official/memory")).
//     Access(sdk.AccessPrivate).Tool(英文说明, sdk.ToolAllow).Handle(私有处理函数)
// AccessPrivate 为默认规则：Core 与本组件；普通第三方插件无法直接访问。
// 成功响应为 {}；失败返回 sdk/errcode 中相应的公开错误。

// 会话条目沿用已有结构，新增字段：
// Kind string `json:"kind,omitempty"`
// 笔记快照：{"role":"system","kind":"memory_snapshot","content":"..."}
// 普通 user/assistant 的 role、speaker、content 含义保持。
// 用户档案只进本轮 Provider 请求，不保存为聊天共享历史条目。

协议复用（无新增 proto）：ToolInvokeRequest{name = 1, arguments_json = 2}、ToolInvokeResponse{result_json = 1}；ServiceInvocation.context 继承可信 Actor/default_target，InvocationContext.actor = 1、default_target = 2、plugin_chain = 4；RouteDescriptor.required_permissions = 17 已存在。所有读取只经审查者定稿后的公开 SDK 接口，当前缺口按停止条件处理。

行为：
- 业务数据归 Agent 受管 SQLite，使用 gormdb.Open(c.Context())；在现有 versioned 迁移列表中追加记忆表步骤，保留 ID 1 的 chat_settings 与账本。Core 的 schema、卸载清单和会话表由 Core 既有实现管理。
- 会话笔记键为 `<会话字符串>@<实例名>`（`string(c.Session()) + "@" + c.Instance()`，即可信默认 Target 的 adapter、chat、thread 与实例），用户档案键为可信 Actor 的 (adapter_instance_id, external_user_id)。按工具 Service 收到的可信上下文（c.User()、c.Session()、c.Instance()）确定，输入不接受调用方自报范围。Agent 的会话历史、聊天设置与 reset 锁改用同一会话键，同一聊天的两个实例互不共用（V-57）。缺少所操作 target 必需的 Actor/默认 Target 时返回 FAILED_PRECONDITION；同一组中任一操作失败，整组回滚。
- memory 工具默认 allow，访问规则只含 Core 与本组件；模型写入统一经 sdk.InvokeTool 的 Core 工具路径。管理员改为 confirm 时由 Core 展示实际参数并等待同一 Actor；改为 deny、用户拒绝或超时都保留既有条目，结果由工具循环交回模型。执行时沿用 Service 访问/权限、父链、期限与取消规则。
- 输入是一组 operations，按顺序执行 add/patch/remove，target 为 memory/user。add 添加条目；patch 与 remove 的 old_text 是已有条目中的子串，恰好一个条目包含它时对该条目整条替换或删除，没有或多于一个条目包含时返回 INVALID_ARGUMENT；空操作组、add/patch 的空 text、patch/remove 的空 old_text、未知 op 或 target 都返回 INVALID_ARGUMENT。整组先在事务内得到最终结果，检查再写入；会话笔记上限 2,200、用户档案 1,375 个 Unicode 码点，均为命名常量。文本按写入顺序逐条以换行连接，码点计数包含实际保存/注入的文本；超限、完全重复的条目与 Unicode Cf 内容返回错误，原数据保持。
- 首轮以 Core 会话实际为空为准，读取对应会话笔记，将快照条目写入会话并放进本轮系统上下文。后续轮次从会话找回同一快照，即使本轮历史截取未包含它也保持；history_entries=0 只影响普通历史，不把现有会话误判为空。并发首轮与 /reset 共用已有会话锁/代次约束，避免重复快照及 reset 后旧快照回写。
- 每轮读取当前 Actor 档案，在本轮 user 消息之前注入 Provider 请求；档案随本轮用户而变，群聊中后续用户的请求不包含前一用户的私有档案。群聊普通 user 条目的 speaker 标记保持。记忆读取失败时，按 Actor locale 回复 agent.memory_error 并结束该轮，Provider 请求不继续。
- 会话内写入的笔记不改变冻结快照；/reset 清除会话快照与历史，保留已写入的长期笔记与档案，下一会话重新读取。保留 ResetGeneration 对在途旧轮次的保护。模型工具联调在 goal-60 工具循环可用后完成；本切片交付前必须提供该证据。
- 输出继续经 c.Stream 以纯文本发送；保留既有 Provider 错误时不自动重发、保留已显示内容、失败轮次不追加 user/assistant、NotCommand 过滤与人格/Provider 选择。
- 自身工具调用（D-142）：core.tool.invoke 的目标插件等于调用方插件时，Broker 不判 CALL_CYCLE、不把调用方再次加入调用链，但仍计入深度上限；其他重复仍返回 CALL_CYCLE；不新增其他放宽。memory 工具仍经 Core 的策略、确认、Service 访问与权限检查。
- ServiceBuilder.Permission 写入的 required_permissions 由 Core 现有的 invoke 检查执行；memory 工具本身不要求额外权限。

测试（按编码规范「测试提交规则」）：
- 提交一个集成测试 agent/memory_test.go 的 TestAgentMemory（unix，表驱动）。复用真实 Core、Agent SDK Session、受管 SQLite、回环 Adapter 与脚本化 httptest Provider，通过模型 -> 工具循环 -> core.tool.invoke 驱动 memory。主路径在群聊中执行 add/patch/remove，验证数据库结果和模型收到的工具结果；本会话写笔记后第二轮仍见原快照，/reset 后重取新快照；两位用户的请求各自只带本人档案。追加 confirm 的允许/拒绝与 deny 子场景，核对实际写入次数与原数据保持。多子场景理由：V-59 同时要求快照、群聊用户隔离及管理员策略，需在真实 Session/存储/工具链中一起证明。
- 原 TestAgentConversation 的会话条目断言适配明确的快照 kind，正常轮次仍按原顺序保存 user/assistant；断流仍不追加失败轮次的这两个条目。不得简单删除原断言。
- 本地自测（*_local_test.go，不提交，PR 写明结果）：2,200/1,375 码点边界与 emoji/组合字符；换行计数；超限/重复/U+200B 拒绝且原数据不变；old_text 缺失/不唯一，操作顺序与整组回滚，memory/user 混合组；空操作组与非法 op/target；缺少可信 Actor/默认 Target；同聊天跨实例与跨 thread 隔离；history_entries=0、小于会话长度及分页仍保留快照；并发首轮、reset 与旧轮次交错；读取失败时 Provider 未调用；confirm 超时/取消无写入；Core 重启后笔记、档案及冻结快照保持。需要独立进程的场景可复用根包 pipeline 装配。
- D-86：SDK、supervisor 和 testdata 测试插件的验证均留本地。goal-60 尚未合入时可先完成存储与通过 SDK 调用的独立验证，模型往返证据保持待完成；收尾时更新至含工具循环的 main 后完成同一集成测试。
- 证据写入 goal-61/tasks.md 与 PR，按 TEST-MATRIX §3 写明位置、命令、OS/arch、Go 版本、结果及真实外部环境未运行项；T-12.7 负责最终汇总。

完成条件：生成器输出与源码/资源一致；gofmt -l 对改动的 Go 文件无输出；go vet ./agent/...；go test -race -count=1 ./agent/...；go build ./...；在 agent 目录运行 go run ../cmd/alonabot plugin check 通过；V-59 的存储、快照、用户隔离与模型经 Core 工具链的策略证据完整。记忆超限导致已有数据丢失、用户档案串读、绕过策略/调用链、Agent 产品代码访问 internal/core，按工作包停止；前述契约/SDK 缺口未闭合时对应验收保持阻塞。

不做：sdk/memory、独立记忆插件、向量检索、外部记忆后端、跨会话检索、技能、定时任务、图片输入、Core 管理 API/页面、D-142 之外的 SDK 公开符号或调用链放宽、修改 Core schema/清理清单、根包产品装配重构、新依赖。
```

## 第 16 波

### goal-62 · T-12.5 技能

状态：已完成（[PR #188](https://github.com/AlonaBot/AlonaBot/pull/188)，合并提交 `c80276e`），证据见 TEST-MATRIX §3.2（含经 cpa 兼容网关运行的四种协议 load_skill 实测）。合入：agent/skills.go（配置项 `skills_dir`，启动时以 os.Root 读取一级子目录的 SKILL.md，严格 YAML frontmatter，不合法 WARN 跳过）；技能清单作为人格之后、会话历史之前的 system 消息；插件内工具 `load_skill`（agent/tools.go 的 roundTool 增加插件内工具）。以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-62（切片 T-12.5：技能）。goal 目录 goal-62，分支 goal-62，worktree .worktrees/goal-62，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 组件启动时读取技能目录下各子目录的 SKILL.md：YAML frontmatter 的 name 与 description 必填，其余内容为正文；系统提示列出各技能的名称与描述，load_skill 工具返回正文；不合法的文件记诊断并跳过，不阻止插件启动。

依据：PLAN T-12 步骤 6；CONTRACTS §13（Agent 段）；TEST-MATRIX V-60。

范围（只新增或修改这些）：
- agent/：新增 skills.go；系统提示与工具目录接入技能（只改调用处）
- agent/config_schema.json 与 agent/i18n/*.json：新增配置项 `skills_dir`（string，缺省空串）及其 config 文案；agent/agent.go 的配置读取加入该项；运行 `go generate ./agent/...` 更新生成产物
- 上述包的测试文件
允许的新跨模块边：无（YAML 用已批准的 go.yaml.in/yaml/v3）。

行为：
- 目录：配置项 `skills_dir` 为空时不启用技能；相对路径以插件工作目录为基准。只读取一级子目录中的 SKILL.md（文件名区分大小写），不跟随符号链接，单个文件上限 256 KiB（命名常量），超过则视为不合法。
- 解析：文件以 "---\n" 开头的 frontmatter 到下一个 "---\n" 为止，用 yaml.v3 严格解码（未知字段允许并忽略）；name 为 [a-z0-9_-]{1,64}，description 为 1～1024 字符的非空字符串；同名技能以目录名字典序第一个为准，其余记诊断。不合法的文件以 WARN 记录（路径与原因）并跳过。
- 系统提示：在人格提示之后、会话历史之前追加一条 system 消息（文案键 agent.skills.header），逐行列出「name: description」，按 name 排序；没有技能时不追加。
- load_skill：插件内工具（不经 Core 的工具目录与策略），在 goal-60 的工具列表中加入，输入 {"name": string}，返回正文；名字不存在时返回错误文本 "error: NOT_FOUND: unknown skill"。不计入额外的确认。
- 技能在插件启动时读取一次，修改后需要重启插件。

测试（按编码规范「测试提交规则」）：
- 提交：agent/skills_test.go（//go:build unix）：临时技能目录含两个合法技能、一个缺少 frontmatter、一个缺少 description、一个超过大小上限：模拟 Provider 收到的系统提示只含两个合法技能，日志有三条诊断，插件正常启动；模型调用 load_skill 得到正文；不存在的名字得到错误结果。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：frontmatter 的边界（CRLF、没有结束分隔）；同名技能；符号链接被忽略；name 的字符规则。

完成条件：gofmt -l 无输出；go generate ./agent/... 后产物无差异；go vet ./...；go test -race ./agent/...；go build ./...；在 agent 目录运行 `go run ../cmd/alonabot plugin check`。

不做：技能的热更新；从网络下载技能。
```

### goal-63 · T-12.6 定时 Agent 任务

状态：已完成（[PR #193](https://github.com/AlonaBot/AlonaBot/pull/193)，合并提交 `7f81469`），证据见 TEST-MATRIX §3.2（含经 cpa 兼容网关运行的四种协议定时任务往返）。合入：agent/tasks.go 与 scheduled.go（schedule_task、list_tasks、cancel_task，后台一轮，会话键 `task/<TaskKey>`，按创建时的目标文本上限快照分段发送，空值与 0 的可选字段视为未给出）；SDK 的 `message.ScheduledPayload.TaskKey`。以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-63（切片 T-12.6：定时 Agent 任务）。goal 目录 goal-63，分支 goal-63，worktree .worktrees/goal-63，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 组件提供 schedule_task、list_tasks、cancel_task 三个 Agent 内工具：只有 Core Admin 能创建与取消（之前经 core.tool.confirm），其他用户只能列出本聊天的任务。任务用 SDK 的动态任务保存与触发（D-124）：参数为创建者、Target 与提示词，触发规则的校验、持久化与错过即跳过由 Core 负责（D-125），Agent 不自建任务表，不注册每分钟的固定任务。任务的处理函数以没有 Actor、只有 job scope 的方式运行一轮对话；发送需要 allowed_adapter_instances，confirm 策略的工具一律拒绝，不注入记忆，会话键与聊天会话分开。

依据：PLAN T-12 步骤 7；CONTRACTS §4.3、§13；DECISIONS D-124、D-125；TEST-MATRIX V-61。

范围（只新增或修改这些）：
- agent/：新增 tasks.go（工具与任务声明）、scheduled.go（任务的处理函数与后台一轮）；Agent 根插件注册任务（只改注册处）
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- 任务声明：sdk.Task[AgentTask]("agent")，AgentTask 为 {creator（instance、user）、target（instance、chat、thread）、prompt、created_at}；sdk.OnTask 在 New() 中注册处理函数。
- schedule_task：输入 {"prompt", "at"?, "interval_seconds"?, "cron"?, "timezone"?}，at（RFC 3339）、interval 与 cron 恰有一个，timezone 只随 cron；interval 至少 60s（命名常量），在调用 Core 之前检查。调用者（工具循环的当前 Actor）必须满足 `c.Has(sdk.Admin)`，否则返回错误结果 "error: PERMISSION_DENIED"；先 ConfirmOperation("agent/schedule_task", 摘要)，允许后按触发规则调用 At、Every 或 Cron（cron 带 sdk.Timezone），返回值中的 key 作为任务 ID 交回模型。Target 为当前对话的默认 Target。Core 返回的 INVALID_ARGUMENT（表达式不合法、at 不晚于当前时间）与 RESOURCE_EXHAUSTED 作为错误结果交回模型。
- list_tasks：调用 List，只返回 target 与当前默认 Target（instance、chat、thread）相同的任务，含 ID、提示词、触发规则与下一次触发时刻；任何人可用。
- cancel_task：输入 {"id"}；只有满足 `c.Has(sdk.Admin)` 的调用者，先确认；先用 List 确认该 ID 属于本聊天，再调用 Cancel；不属于本聊天或不存在时返回错误结果 "error: NOT_FOUND"。
- 处理函数运行一轮对话：
  - 没有 Actor，只有 job scope；会话键为 "task/<key>"，与聊天会话分开；不调用记忆服务。
  - 回复经 sdk.Send 显式 To 任务的 Target，需要 allowed_adapter_instances，没有许可时 PERMISSION_DENIED 并记 WARN（周期任务保留）。
  - 工具循环中 confirm 策略的 Core 工具与全部 MCP 工具一律拒绝（不发起确认，因为没有 Actor），作为错误结果交回模型。
  - 一次运行受 JOB_TIMEOUT 约束；同一任务上一轮未结束时由 Core 跳过本轮。
  - 一次性任务在派发时已由 Core 删除，对话失败不重试，失败写日志。
- 插件离线期间到期的触发由 Core 作废：周期任务从下一个未来时刻继续，一次性任务删除。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tasks_test.go（//go:build unix）：经 agent_test.go 的真实 Start 装配，用模拟 Provider（脚本化调用 schedule_task 等工具），到期用接近当前时刻的 at 绝对时间触发（真实时钟，不注入）：
  - 非 Admin 调用 schedule_task 得到错误结果，List 为空；Admin 经确认后创建 at、interval 与 cron 任务各一个。
  - at 任务到期：运行一轮，回复发往任务 Target；没有发信许可时没有发送并记 WARN，授予许可后另一个 at 任务运行时发送；at 任务运行一次后 List 中不再出现；interval 与 cron 任务在 List 中带触发规则与 NextAt。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：interval 下限；三种触发规则互斥；会话键隔离；后台一轮中模型调用 confirm 策略的工具得到拒绝结果；没有注入记忆（请求中没有快照与档案）；其他用户 list_tasks 只看到本聊天的任务；cancel_task 需要 Admin 与确认，不能取消其他聊天的任务；Agent 的数据库中没有任务表。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/...；go build ./...；在 agent 目录运行 `go run ../cmd/alonabot plugin check`。

不做：打包与 T-12 验收（T-12.7）；任务的管理页面。
```

## 第 17 波

### goal-64 · T-12.7 官方 Agent 组件打包与 T-12 验收

状态：已完成（[PR #195](https://github.com/AlonaBot/AlonaBot/pull/195)，合并提交 `47ae5f6`），证据见 TEST-MATRIX §3.2（含审查者经 cpa 兼容网关对安装包路径的四种协议实测）。合入：`cmd/alonabot/t12_e2e_test.go` 的 `TestAgentPackage`（正式 pack 的 Agent 与 probe 安装包经管理 CLI 安装、启用后的主路径）；Agent 对入站图片空 MIME 读取后按内容判断（CONTRACTS §1.3）；PLAN T-12 步骤 8 按增量验收同步。以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-64（切片 T-12.7：官方 Agent 组件打包与 T-12 验收）。goal 目录 goal-64，分支 goal-64，worktree .worktrees/goal-64，从已包含 PR #193、#194 的最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：用正式 alonabot pack 产出的官方 Agent 安装包，经管理入口安装并启用，新增一条跨进程主路径集成测试，证明包内资源、首次启用、受管数据库与 Agent 关键交互接通；对话、Core/MCP 工具、记忆、技能、定时任务、llm.generate、图片各选最小代表。T-12 的其余验收分项在 TEST-MATRIX §3.2 汇总行引用前置切片的既有证据，保留各项实际平台、真实环境与未运行项。

依据：spec/PLAN.md §3.1、§5；spec/plan/agent.md T-12.7、T-12 步骤 8、验收与失败停止条件，以及 T-12.2～T-12.6、T-12.8、T-12.14 的已交付行为；spec/plan/core.md T-09.3；spec/CONTRACTS-v0.1.md §3、§4.3、§9.1～§9.3、§13、§15.1；spec/MESSAGE-MODEL.md 的 Actor、Target 与媒体片段；spec/adr/DECISIONS.md D-83、D-86、D-124、D-125、D-132、D-142；docs/coding-standards.md「测试提交规则」；TEST-MATRIX V-57～V-61、V-76、V-118 与 §3。执行前确认审查者已把工作包步骤 8 同步为「新增安装包主路径证据，其余分项引用既有证据」。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t12_e2e_test.go（//go:build unix，package main）：仅一个 TestAgentPackage 集成测试，私有辅助也放本文件；复用同包既有 CLI/Core/Telegram 辅助。
- spec/TEST-MATRIX.md §3.2：追加 T-12.7 / T-12 汇总证据行，记录本次安装包路径的增量并引用各前置行；本切片获授权修改这一处。
- 本地自测如确有缺口，写在 cmd/alonabot/*_local_test.go 或 agent/*_local_test.go；执行记录留 goal-64/ 与 PR。
agent/plugin.yaml 已具备工程身份，沿用现值；os、arch、entrypoint、manifest_version、协议版本范围、required_capabilities、config_version、migration_mode、prepare_effects 由 build/pack 生成。Agent 的产品实现、配置、i18n 与生成源码沿用前置交付；发现需要修改时，按范围外变更交审查者处理。
允许的新跨模块边：无。测试沿用 cmd/alonabot 已有的管理入口与 telegramtest；Agent 产品代码继续只经 SDK 与 Core 通信。现有 probe 与 mcptest 按目录构建和运行，测试无须跨 internal 可见性导入它们。

接口（按已读代码复用，公开接口保持现状）：

1. 工程与打包命令（从仓库根执行；实际测试使用 TestMain 已构建的 alonabot）：
   alonabot check --dir agent [--goos <os>] [--goarch <arch>] [--tags <tags>]
   alonabot build --dir agent --out <临时构建目录> [--goos <os>] [--goarch <arch>] [--tags <tags>]
   alonabot test --dir agent
   alonabot pack --dir agent --asset-url <绝对 https URL> --out <临时打包目录> [--key <私钥文件>] [--goos <os>] [--goarch <arch>] [--tags <tags>]
   --dir 缺省为当前目录；build/pack 的 --out 缺省为工程 bin/；--tags 为逗号分隔的列表。test 只有 --dir 这个组件参数，在临时离线会话加载、注册、激活、停止组件。
   --asset-url 必填且 scheme 为 https；本地安装直接读本次归档，无须访问这个发布 URL。--key 为包含标准 base64 编码的 64 字节 Ed25519 私钥的文件路径。
   当前平台产物：alonabot-agent-official-0.1.0-<GOOS>-<GOARCH>.tar.gz、归档外 release.json；指定 --key 时另有 release.json.sig 与 release.pub。归档含包内 plugin.yaml、component（Windows 为 component.exe）、build-info、规范 config_schema.json、i18n 及已声明图标。
   internal/pluginkit 的现有签名：
   func Pack(ctx context.Context, options generate.Options, assetURL, keyFile, outDir string) (archive string, err error)
   sdk/generate 包内：type Options struct { Dir string; GOOS, GOARCH string; Tags []string; Resources map[string][]byte; CGOEnabled string; RequireAdapterConstructor bool }（正式构建由 pluginkit 设 CGOEnabled="0" 与资源快照）
   build/pack 会写工程的 alonabot_gen.go；测试的产物与数据目录放 t.TempDir，生成源码应与前置提交一致。

2. 安装、启用与认证：
   alonabot run --data-dir <临时 data_dir> --listen 127.0.0.1:0 [<开发组件目录>...]
   alonabot install <归档绝对路径> --yes --admin-url <管理 URL> --token-file <token 文件>
   alonabot enable alonabot-agent-official --admin-url <管理 URL> --token-file <token 文件>
   alonabot list --admin-url <管理 URL> --token-file <token 文件>
   install 也接受 --archive；--release 显式指定 release.json，缺省读取归档同目录；可选 --signature、--public-key，缺省读取 release 文件同目录的 release.json.sig、release.pub。管理 CLI 也可从 ALONABOT_ADMIN_TOKEN 读取 token，--data-dir 用于发现 base_url。
   POST /api/v1/packages/install：Authorization: Bearer <token>，JSON 使用 manage.InstallRequest：
   type InstallRequest struct {
       Archive, ReleaseJSON, Signature string
       PublicKey string
       Source string
       Replace, Confirm bool
   }
   Archive 为 Core 所在本机的归档路径；ReleaseJSON、Signature、PublicKey 为文件内容。主路径使用本地包，Source 留空、Confirm 为 true；JSON 可按上述 Go 字段名编码。
   POST /api/v1/packages/enable：{"id":"alonabot-agent-official"}。
   GET /api/v1/packages：核对 Package 的 Role="agent"、Source="github.com/AlonaBot/AlonaBot"、Version、Enabled、Active、State；这些字段按 Go 字段名编码，首次 install 后为 Installed，enable 后为 Activated。
   Agent 配置在首次 enable 前写到 <data_dir>/plugins/alonabot-agent-official/config.json；配置沿用既有文件入口。
   Core Admin 可经 alonabot admin add telegram:88 写入；后台发信许可沿用 allowed_adapter_instances 缺省允许全部实例（D-83）。

3. Agent 已有配置与命令：
   agent 包内：func New() *sdk.Plugin
   providers 的条目是 template_list，例如：
   {"custom":{"name":"synthetic","protocol":"openai-compatible","base_url":"<模拟 Provider URL>","api_key":"synthetic-key","model":"test-model","max_tokens":128,"image_input":true}}
   default_provider="synthetic"；personas 为 {name,prompt} 列表，default_persona 引用其中名字；history_entries 缺省 20；skills_dir 缺省空串，测试指定临时技能目录的绝对路径；llm_service_allowed_plugins 缺省 ["*"]。
   mcp_servers 为列表，stdio 条目 {name,transport:"stdio",command,args,env:[{name,value}]}；另一传输名为 "streamable-http"，字段为 url、headers:[{name,value}]。
   命令：/ask <text>（message.Chain 尾参）、/agent on、/agent off、/provider [name]、/persona [name]、/reset。agent on/off、provider、persona 需要 sdk.Admin；ask、reset 需要 sdk.User；自由对话要求已开启且满足 ToMe 与 NotCommand。
   Core 工具原名 alonabot-agent-official/memory，AccessPrivate、缺省 allow：
   {"operations":[{"op":"add","target":"memory","text":"package note"}]} -> {}
   target 也可为 user；patch/remove 用 old_text。会话键为 string(c.Session())+"@"+c.Instance()；可信范围由 Core/SDK 上下文提供。
   Agent 内工具 load_skill：{"name":"package-skill"} -> 技能正文。
   T-12.6 契约的 Agent 内工具：schedule_task 输入 {prompt,at?,interval_seconds?,cron?,timezone?}，list_tasks 无参数，cancel_task 输入 {id}；创建/取消要求 Core Admin 并先确认；本次只选择 at 的创建与到期运行。工具名与参数以 main 上 agent/tasks.go 的注册为准。
   动态任务底座的已存在签名（以下为 sdk 包内签名，供核对 #193）：
   func Task[T any](name string) TaskType[T]
   func OnTask[T any](r Registrar, t TaskType[T]) *TaskBuilder[T]
   func (t TaskType[T]) At(ctx context.Context, when time.Time, args T, opts ...TaskOption) (string, error)
   func (t TaskType[T]) Every(ctx context.Context, d time.Duration, args T, opts ...TaskOption) (string, error)
   func (t TaskType[T]) Cron(ctx context.Context, expr string, args T, opts ...TaskOption) (string, error)
   func (t TaskType[T]) Cancel(ctx context.Context, key string) error
   func (t TaskType[T]) List(ctx context.Context) ([]TaskInstance[T], error)
   TaskInstance[T] 的字段为 Key string、Args T、At time.Time、Interval time.Duration、Cron/Timezone string、NextAt *time.Time。
   Core/MCP 的模型侧工具名由现有映射产生，脚本化 Provider 根据收到的工具目录识别目标并使用实际名字；直接调用原工具名的路径继续由 Agent 工具循环承担。

4. llm.generate 与既有调用方（以下为 sdk/llm 包内签名）：
   var GenerateService = sdk.JSON[Request, Response](AgentTarget + "/" + ServiceName)
   const AgentTarget = "@agent"
   const ServiceName = "llm.generate"
   func Generate(ctx context.Context, req Request) (Response, error)
   func Ask(ctx context.Context, prompt string) (string, error)
   type Message struct { Role string `json:"role"`; Text string `json:"text"` }
   type Request struct {
       Messages []Message `json:"messages"`
       System string `json:"system,omitempty"`
       Provider string `json:"provider,omitempty"`
       MaxOutputTokens int `json:"max_output_tokens,omitempty"`
   }
   type Response struct {
       Text string `json:"text"`; Provider string `json:"provider"`; Model string `json:"model"`
       InputTokens *int `json:"input_tokens,omitempty"`; OutputTokens *int `json:"output_tokens,omitempty"`
   }
   复用 plugintest/internal/probe（ID plugintest_probe）的 /llmask：其 Handler 已调用 llm.Ask(c.Context(), "  original prompt\n")，后者调用真实 llm.Generate。请求为 Messages:[{Role:"user",Text:"  original prompt\n"}]，其余字段取零值，Agent 使用配置的默认 Provider 与 max_tokens；模拟 Provider 返回 "generated"，探针把返回文本回复到原聊天。直接用 pack 构建现有 probe 工程并安装、启用，源码与既有断言保持现状。probe 的 /llmgenerate 对可选用量字段有针对替身的固定预期，主路径采用 /llmask。

5. 现有测试辅助的可用方式：
   同包直接复用：startEmptyCore(t *testing.T, dataDir string, extra ...string) *core；(*core).management(t *testing.T, dir, token string, args ...string) string；findPackage(t *testing.T, c *core, dir, id string) manage.Package；stopTestCore(t *testing.T, c *core)；developmentWrite/Read；telegramMessage、telegramUpdate、telegramEventually。管理辅助已有 token 与 Host 处理。
   telegramtest.NewServer(t testing.TB, token string) *telegramtest.Server；(*Server).URL() string、SetMe(userID int64, username string)、AddFile(fileID, filePath string, content []byte)、Push(updates ...json.RawMessage)、Requests() []telegramtest.Request、Calls(method string) int。Request 有 Method string、Params map[string]any、Files map[string][]byte；确认按钮从发送请求的 reply_markup 读取 callback_data，再以相同 Actor 的 callback_query 回送。
   Agent 的模拟 Provider 为各测试文件中的 httptest Handler，当前没有可跨测试包直接导入的统一 fixture。参考 agent/tools_test.go 的 completionRequest、writeToolCall、checkToolResult 与 agent/agent_test.go 的 writeDelta/writeEnd，按本次主路径在 t12_e2e_test.go 写最少的私有 HTTP/SSE 辅助；这些未导出函数属于另一个测试包。
   agent/provider/provider_test.go 的 mockProvider(t *testing.T, protocol Protocol, status int, stream string, images ...bool) http.HandlerFunc、testRequest() Request 用于 Provider 分项，留在原测试包。
   stdio MCP 复用 go build -o <临时 mcptest> ./agent/mcptest；参数 --pid-file <文件>、--ignore-eof（后者用于已有故障测试，本次无需启用）。现有 echo/tool 输入 {text:string,fail?:bool}，输出包含文本、结构化结果与调用计数；使用 stdio 一个传输即可。streamable HTTP 与故障场景引用 T-12.3 的既有证据。

行为与验收矩阵（一个 TestAgentPackage 内顺序执行、共享同一个安装包与 Core；下列为主路径阶段）：
- 打包与启用：当前平台调用正式 pack，--asset-url 使用明确的测试 https URL，输出到临时目录；主路径选择 unsigned 包，安装时 --yes 确认来源与未签名包。通过管理 CLI install，再 enable；核对包身份、版本、Activated 与有效 Active。首次启用使用包内 schema/i18n、生成入口和受管 SQLite。Agent 的运行来源始终为本次安装包，alonabot run 的开发目录仅用于真实 Telegram Adapter；probe 同样用现有工程 pack 后安装。Core/Agent/Adapter/probe 都运行真实实现与 Session。
- V-57：复用 Telegram Bot API 模拟服务，配置一个测试账号与实例；把 telegram:88 设为 Core Admin，/agent on 后发送一条普通私聊消息。只用 openai-compatible 一个协议，模拟 Provider 返回文本增量，最终回复经真实 Adapter 回到模拟服务；核对最终文本与该轮工具往返的请求记录。
- V-58、V-59、V-60：同一轮里模型依次调用 memory（add 一条会话笔记）、load_skill（一个合法 SKILL.md 的正文）、stdio MCP 的 echo/tool，然后生成最终回复。MCP 操作先在平台侧展示确认，原 Actor 点击允许后恰好执行一次；核对每次 ToolCallID 与回传结果、技能正文和 MCP 实际参数。memory 通过 Core 的自身工具路径执行，记忆的实际效果在 /reset 后新轮请求中验证。
- V-118 与 V-59 的持久结果：/reset 后发送一条带有效图片的消息，模拟 Provider 同时收到前轮写入的会话笔记与本轮图片的预期 MIME/字节；下一条纯文本消息仍回放这一张图片。再次 /reset 后的请求保留长期笔记、图片回放为空。图片经 telegramtest.AddFile、真实 Adapter、Core 媒体入口与 Agent 的 sdk.OpenMedia 取得；图片窗口边界、失败与多图规则引用既有证据。
- V-61：同一管理员发起新一轮，模型调用 schedule_task 创建一个未来 RFC 3339 at 任务，真实确认按钮允许后得到任务 ID；按真实时钟等待 Core Scheduler 派发，任务执行一轮并向原 Target 发送唯一的标记回复。核对后台 Provider 请求未注入会话笔记/用户档案，at 任务只运行一次。时刻由测试运行时计算，等待复用既有端到端测试的机制；interval/cron、无许可、confirm 拒绝和离线错过的分项引用 T-12.6 与 T-07.9 证据。
- V-76：通过已安装并启用的 plugintest_probe 收到 /llmask，驱动真实 Handler -> llm.Ask -> llm.Generate -> Core @agent 解析 -> 已安装 Agent -> 模拟 Provider -> 原聊天回复。核对最终文本为 generated，Provider 请求保留上述 prompt 的原始空白、使用配置的 model="test-model" 与 max_tokens=128、无工具和聊天历史。调用方使用 llm_service_allowed_plugins 的默认 ["*"]，用量、名单拒绝与其他错误分项引用 T-12.8。
- 收尾：经管理入口 disable Agent，随后正常停止 Core，回收 Agent 与其已启动的 stdio MCP 进程。用 mcptest 的 PID 记录核对该进程退出，保留完整失败输出。

TEST-MATRIX 汇总方式：
- 新增「T-12.7 / T-12 安装包主路径与汇总」一行，列 TestAgentPackage、实际运行命令、候选提交、OS/arch、Go 版本、结果及 goal-64/ 日志；明确外部环境为 httptest Provider、stdio MCP 与 Telegram Bot API 模拟服务。
- V-57 的其余对话场景引用 T-12.2（#170）行；协议客户端分项沿用 T-12.1、T-12.9～T-12.13 的现有证据。
- V-58、V-59、V-60、V-61 分别引用 T-12.3（#178）、T-12.4（#182）、T-12.5（#188）、T-12.6（#193）行；Scheduler 周期/错过的底座证据沿用 T-07.9（#176）。
- V-76 引用 T-12.8（#175）行，同时保留该行 V-82 的 @agent、V-109、V-116 的 llm.providers 分项；V-118 对话分项引用 T-12.14（#189）行与已有 Provider 图片证据。
- 保留前置行经 cpa 兼容网关运行真实 Provider 的日期、模型、协议和结果；汇总行写明引用来源。本次新跑的安装包结果单独记录，既有真实证据保持原样。
- 某行缺少证据时标明具体分项、责任人与补齐条件，交审查者处理；使用「已运行的增量＋既有证据引用＋未运行项」表达 T-12 验收状态。

测试（按编码规范「测试提交规则」与 D-86）：
- 提交：仅 cmd/alonabot/t12_e2e_test.go 的 TestAgentPackage 一个集成测试。理由：前置已验证功能分项，本切片新增的风险集中在正式打包、管理安装/启用、跨进程注册、资源读取与真实协作链接通；上述顺序阶段用一套已安装实例证明这一条主路径。
- D-86：复用现有 probe、mcptest 和 Telegram 模拟服务，新增私有辅助只服务本测试；SDK/supervisor 的额外验证、临时插件与故障复现留 *_local_test.go 或 tmp/test/。维持单个提交集成测试。
- 本地自测：对官方 Agent 实际运行 check、build、test、pack（当前平台；输出目录临时化），记录命令及结果。签名流程与通用包损坏/篡改的矩阵引用 T-09.3 的既有证据；只有实际发现的安装路径缺口才补本地复现。前置分项沿用既有有效证据。
- 真实 Provider：由审查者在本地使用已配置的 cpa 兼容网关，基于本切片 pack 后安装、启用的 Agent 跑一次代表性真实对话与工具往返，按 §3 记录模型、协议、命令、平台、结果及本地日志；本地测试使用 *_local_test.go，凭证留在既有本地配置。执行者提交的测试只连模拟 Provider。审查者补验前，本次「安装包＋真实 Provider」标为待补验，同时引用前置已完成的 cpa 实测。
- 真实 Telegram：当前没有凭证，记「未运行」，后续由 T-10 汇总。模拟 Bot API 的通过记录按模拟环境填写；其他操作系统实际运行沿用各行的现有状态。
- PR 写明提交测试的新增行数、本地运行结果与审查者待补项；证据回写 goal-64/tasks.md 与 TEST-MATRIX §3.2。

完成条件：改动 Go 文件 gofmt -l 无输出；go vet ./...；go test -race -count=1 ./cmd/alonabot -run '^TestAgentPackage$' -v；go test -race ./...；go build ./...；官方 Agent 的顶层 check、build、test、pack 当前平台本地记录齐全，生成资源与前置源码一致。若全仓测试已在同一最终候选中包含 TestAgentPackage 并实际运行，可复用该结果。TEST-MATRIX 汇总行能追到各前置证据、区分模拟与真实环境、保留未运行项；审查者负责合并前的 cpa 安装包补验回写。安装包激活失败、缺必要资源、绕过工具确认、Provider 失败自动重发、记忆串读或丢失、Core 会话写入图片字节、Agent 产品代码导入 internal/core，按工作包停止条件处理。

不做：复制 V-57～V-61、V-76、V-118 的全量分项到安装包测试；新增测试插件工程或公共测试框架；扩充 Agent 功能、配置和 SDK 接口；新增依赖；发行归档与预装（T-14.3、T-14.4）；真实 Telegram 接入和多平台实机验收（T-10）。
```
## 第 18 波

### goal-143 · T-12.15 模型两层配置

状态：已完成（PR #207），合并提交 `1e010d7`。2026-11-17 按该提交核对：`agent/config_schema.json` 的 `providers` 为 template_list，只保存 name、protocol、base_url 与 secret 的 api_key；`models` 为 object 列表，字段为 name、provider、model、type，chat 另有 max_tokens 与 image_input；`chat_model` 引用模型名。`agent/agent.go` 中 `configuredProvider` 为 `{name string; connection provider.Config}`，`configuredModel` 为 `{name string; provider *configuredProvider; modelType string; client *provider.Client; maxTokens int; model string; imageInput bool}`；配置类型为 `configuration`，持有 providers、models、chatModel、personas、defaultPersona、historyEntries、mcpServers、skillsDir、llmServiceAllowedPlugins。实际签名为 `func readConfiguration(values config.Values) (configuration, error)`、`func (s *configuration) provider(name string) *configuredProvider`、`func (s *configuration) model(name string) *configuredModel`；只有 chat 模型已构造 client，五类预留模型的 client 为 nil。`chatSettings.Model` 保存模型名，`func (a *component) selectModel(c *sdk.Context, name string) error` 处理 `/provider`。`sdk/llm.Request.Model` 为 `string`（JSON `model,omitempty`，空串使用 chat_model），`sdk/llm.Response.Model` 为 `string`（JSON `model`，实际使用的模型配置名）；两者已删除 Provider 字段。`agent/options.go` 的 `modelsServiceName = "llm.models"`，经 `options.Serve(p, modelsServiceName, a.models)` 注册；处理函数为 `func (a *component) models(_ context.Context, q options.Query) (options.Page, error)`，支持 filter.type、省略时列出全部六类、名称搜索与按 q.Values 取回，Value/Label 为模型名，Description 为「提供商 · 模型 ID」，不返回地址与密钥。以下保留派发时语句；预留五类客户端与管理路由在 goal-144，重试回退与图像转述在 goal-145。

```text
/goal 使用 $alonabot-execute 推进 goal-143（切片 T-12.15：模型两层配置）。goal 目录 goal-143，分支 goal-143，worktree .worktrees/goal-143，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 配置从「每个 Provider 自带模型」改为提供商 `providers` 与模型 `models` 两层，全局 `chat_model` 取代 `default_provider`；对话与 `/provider` 按模型名选择；`llm.generate` 的 `provider` 改为 `model`，选项来源 `llm.providers` 改为 `llm.models` 并接受 filter `type`。

依据：PLAN T-12 步骤 2、9；CONTRACTS §13.5；DECISIONS D-131、D-154；TEST-MATRIX V-76、V-116 的 `llm.models` 分项、V-123。

范围（只新增或修改这些）：
- agent/config_schema.json、agent/i18n/：providers（template_list，模板即预设：openai、anthropic、gemini、deepseek、ollama 等，每项 name、protocol、base_url、secret 的 api_key）与 models（object 列表：name、provider、model、type，type 为 chat 时另有 max_tokens 与缺省 false 的 image_input）、chat_model（options_from `@agent/llm.models` 且 filter type=chat）；删除 default_provider 与 Provider 条目中的模型字段
- agent/settings.go、agent/agent.go、agent/generate.go、agent/options.go：加载与校验、按模型名选择 Provider 客户端、`/provider` 命令改为选择模型
- sdk/llm/：Request.Provider 改为 Model（json `model`），Response.Provider 改为 Model；文档注释同步
- 全部调用方（examples、plugins 中调用 llm 的代码与测试）
- 上述的测试文件

接口：
- package llm：Request 字段 `Model string \`json:"model,omitempty"\``（空串使用 Agent 的 chat_model），删除 Provider；Response 字段 `Model string \`json:"model"\``（实际使用的模型名），删除 Provider
- 选项来源服务名 `llm.models`，filter 键 `type`（省略时列出全部类型），返回的显示名为模型名，说明为「提供商 · 模型 ID」，不含 base_url 与 API key

行为：名称在各自列表内唯一；模型引用不存在的提供商、类型与提供商协议不匹配（chat 用现有六种协议，其余五类只允许 openai 与 openai-compatible）、chat_model 不是 chat 类型时配置加载失败并带 JSON 路径；同一提供商的多个模型共用一份密钥；chat_model 为空时 `/ask` 回复提示管理员配置；首次发布前不写迁移。

测试（按编码规范「测试提交规则」）：
- 提交：agent/agent_test.go 改写既有的对话子测试为两层配置（一个提供商下两个 chat 模型），断言 `/provider` 切换模型后请求发往对应的模型 ID；llm.generate 的既有子测试改用 model 字段。
- 本地自测（*_local_test.go，结果写入 PR）：V-123 的全部校验分项；`llm.models` 的 type 筛选与按存储值取回；chat_model 为空。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：预留五类模型的客户端与模型管理路由（goal-144）；重试回退与图像转述（goal-145）；模型页（goal-146）；配置文件（goal-164）。
```

## 第 19 波

### goal-144 · T-12.16 预留的五类模型与模型管理路由

状态：待派发。唯一前置 T-12.15（PR #207，合并提交 `1e010d7`）已合并；2026-11-17 按该提交改写。两层配置、`configuredProvider`、`configuredModel`、`configuration.model`、`llm.Request.Model`/`Response.Model` 与 `llm.models` 的实际接口见 goal-143 状态段。预留五类已能加载但没有 client；Provider 现有入口是 `New` 与 `(*Client).Stream`，尚无五类单次请求与 discover。`New()` 已登记受管数据库迁移、`a.generate` 和选项服务；`runToolLoop` 丢弃 Stream 的 Result，`generate` 已读取 Usage；管理路由与用量表尚未实现。可与 goal-145 并行，双方会改 `agent/agent.go`、`conversation.go`、`loop.go`、`scheduled.go`、`generate.go`、`agent_test.go` 和 `provider/provider.go`：后合并者同步 main，保留对方的客户端、错误分类、重试、转述与用量统计改动，更新全部调用方；合并后转述和回退请求接入同一用量统计，每次请求只计一次。

```text
/goal 使用 $alonabot-execute 推进 goal-144（切片 T-12.16：预留的五类模型与模型管理路由）。goal 目录 goal-144，分支 goal-144，worktree .worktrees/goal-144，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：为 speech_to_text、text_to_speech、embedding、rerank 与 decision 提供单次请求客户端与本地模拟服务；Agent 提供非公开的模型管理路由 discover、test 与 usage，并按模型名与日期累加用量。

依据：spec/PLAN.md §3.1；spec/plan/agent.md T-12.16、T-12 步骤 12 及失败停止条件；spec/CONTRACTS-v0.1.md §7.2、§7.3、§13.5、§13.6；spec/adr/DECISIONS.md D-86、D-132、D-154；TEST-MATRIX V-124；ARCHITECTURE.md 的 Agent 依赖边与受管数据归属。

范围（只新增或修改这些）：
- agent/provider/：新增五类单次请求与模型发现的实现文件，provider.go 中复用或补齐连接、认证与脱敏处理；已有六种 chat 协议行为保持现状。
- agent/agent.go：为五类模型构造客户端，登记模型管理路由，用量表并入现有初始迁移；新增 agent/models.go（HTTP 处理）、agent/usage.go（受管 SQLite 的用量表与累加）、agent/model_test.wav（不超过 2 秒的短 WAV，用于嵌入模型测试请求）。
- agent/loop.go、agent/generate.go、agent/conversation.go、agent/scheduled.go：在真实请求返回用量时接入统计；如需调整内部签名，同步这些调用方。goal-145 已合入时，连同 agent/image.go 及其新增的重试、图像转述实现接入同一统计；尚未合入时由后合并者完成交接。
- 新增 agent/models_test.go 的一个主路径集成测试；agent/agent_test.go 只同步已有资源快照及必要的装配辅助。其余验证写在上述目录的 *_local_test.go。
允许的新跨模块边：无。Agent 继续经 sdk/gormdb 与 sdk.Plugin 的 HTTP 路由访问 Core 能力；provider 只用标准库和已批准的 Provider SDK。

接口（先复用 main 的真实签名；以下新增入口是本切片交付）：
// agent 包内已有
func New() *sdk.Plugin
func readConfiguration(values config.Values) (configuration, error)
func (s *configuration) provider(name string) *configuredProvider
func (s *configuration) model(name string) *configuredModel
func (a *component) generate(c *sdk.Context, req llm.Request) (llm.Response, error)
func runToolLoop(ctx context.Context, selected *configuredModel, request provider.Request, catalog toolCatalog, emit func(string) error) error
// provider 包内已有；模型 ID 从 configuredModel.model 写入连接副本的 Config.Model
type Config struct { Protocol Protocol; BaseURL string; APIKey string; Model string }
type Usage struct { InputTokens int; OutputTokens int }
type Result struct { StopReason StopReason; Usage *Usage }
func New(config Config, httpClient *http.Client) (*Client, error)
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
// provider 包内新增；五类函数各完成一次请求
func (c *Client) Transcribe(ctx context.Context, audio []byte) (string, error)
func (c *Client) Synthesize(ctx context.Context, text string) (audio []byte, mime string, err error)
func (c *Client) Embed(ctx context.Context, text string) ([]float64, error)
type RankedText struct { Index int; Score float64 }
func (c *Client) Rerank(ctx context.Context, query string, texts []string) ([]RankedText, error)
func (c *Client) Decide(ctx context.Context, prompt string) (float64, error)
func Discover(ctx context.Context, config Config, httpClient *http.Client) (ids []string, description string, err error)
// Discover 只使用连接字段；已保存的 configuredProvider.connection 没有 Model。
// sdk 包内已有；用 p.Route 注册，缺省 AuthCoreSession
func (r *Plugin) Route(method, path string) *RouteBuilder
func (b *RouteBuilder) HandleFunc(f func(http.ResponseWriter, *http.Request))
// sdk/gormdb 包内已有；HTTP Handler 使用 r.Context()
func Open(ctx context.Context) (*gorm.DB, error)

HTTP 接口按 CONTRACTS §13.6 实现：POST /models/discover 的 JSON 输入为 {provider} 或 {protocol,base_url,api_key}，输出模型 ID、推断类型和不支持列表时的说明；POST /models/test 输入 {model}（Active 配置中的模型名字），输出 ok、耗时毫秒数和失败时的英文说明；GET /models/usage 输出各模型当天的输入、输出 token 数。Core 外部路径为 /plugins/{agent_id}/models/...。JSON 承载结构与包内私有辅助就近定义；在 PR 中列出实际 JSON 字段，供 goal-146 接入。

行为：
- readConfiguration 保留两层引用与协议校验；五类模型也按模型条目复制连接并填入 Model 后构造客户端，共享提供商密钥。五类仅用于配置、发现与测试，调用方只在模型测试路由中选择它们。
- 五类函数不流式、不重试：Transcribe 为 multipart POST <BaseURL>/audio/transcriptions（WAV 与模型 ID），返回文本；Synthesize 为 POST /audio/speech，返回音频字节与 MIME；Embed 为 POST /embeddings，返回向量；Rerank 为 Jina/Cohere 格式的 POST /rerank，返回文本序号与分数；Decide 经 Chat Completions 请求 JSON 输出，按步骤 12 返回布尔问题的概率，非 JSON 或概率超出 0～1 返回错误。仅允许 openai 与 openai-compatible，错误脱敏沿用 Provider 规则，SDK 内置重试保持 0，配置只取显式参数。
- discover 接受已保存的提供商或未保存的完整连接，不写配置。openai、openai-compatible、openai-responses 为 GET <BaseURL>/models；anthropic-messages 为 GET /v1/models；gemini-generate-content 为 GET /models；ollama-chat 为 GET /api/tags，使用各协议的认证方式。ID 含 embed 推断 embedding、含 rerank 为 rerank、含 whisper 或 transcribe 为 speech_to_text、含 tts 为 text_to_speech，其余为 chat。Provider 不支持列表时返回空列表与说明；请求失败保留错误。
- test 只取 Active 中保存的模型：chat 与 decision 用一句短文本，speech_to_text 用嵌入的短 WAV，text_to_speech 用一句短文本，embedding 用一个词，rerank 用一个查询与两段文本。每次只请求一次，返回 ok、耗时与失败说明，测试请求不计用量。
- 三个路由保持非公开，沿用 Core 的管理会话与非 GET 的 CSRF 检查；响应与错误说明均去除 API key、认证头及带凭证的 URL，禁止直接回送 Provider 原始错误。
- 用量按模型配置名与 Agent 进程本地日期累加到受管 SQLite，usage 只返回当天。对话的每次工具循环请求、定时任务与 llm.generate 在 Result.Usage 非 nil 时累加，nil 保持缺失，不推算 token；并发累加原子完成，读写失败返回错误。goal-145 的转述已存在时接入同样统计，回退按实际使用的模型名字记账；后合并者保留这些统计点。当前迁移 ID 1 创建 chatSettings/conversationImage，ID 2 创建 memoryRecord，用量表追加到 ID 1 的 AutoMigrate，沿用已有 ID，不新增迁移版本。

测试（按编码规范「测试提交规则」）：
- 提交：agent/models_test.go（//go:build unix，package agent_test）的一个 TestAgentModels 集成测试，复用 agent_test.go 的真实 Start、loopback 与受管 SQLite 装配：管理会话经 Core 代理 discover 一个已保存的 openai-compatible 提供商，返回一个 chat 模型；经 test 测试该模型成功且用量仍为零；再通过 /ask 完成一次带用量的对话，usage 返回该模型当天的输入与输出 token 数。模拟 Provider 校验请求模型 ID 与认证。
- 本地自测（*_local_test.go，不提交，结果写入 PR）：V-124 的其余分项，包括五类单次请求的成功与错误、decision 的非 JSON/越界概率、密钥脱敏，discover 两种输入与全部列表端点、类型推断与不支持列表，test 六类及失败且不重试、不计量，无管理会话和缺少 CSRF 时由 Core 拒绝，usage 跨日期与并发累加、llm.generate 用量。goal-145 已合入时补图像转述与实际回退模型的计量；尚未合入时在 PR 写明该分项待 goal-145 接入。真实 Provider 有凭证时实测，否则写「真实 Provider 未运行」。不向 sdk/、internal/core/supervisor/ 或新 testdata 插件提交测试（D-86）。

完成条件：gofmt -l agent 无输出；go vet ./agent/...；go test -race -count=1 ./agent/...；go build ./...；go run ./cmd/alonabot check --dir agent。PR 写明实际运行平台、本地自测结果与未运行项。

不做：模型页（goal-146）；重试与图像转述的实现（goal-145）；为预留五类添加 Agent 功能、服务或引用它们的配置字段；decision 的选项概率与分档评分（D-154 列出的另两种结构化判断，v0.1 没有使用方，按 PLAN T-12 步骤 12 只做是/否概率，接入使用方时另定）；配置文件（goal-164）。
```

### goal-145 · T-12.17 对话重试与回退模型、图像转述

状态：待派发。唯一前置 T-12.15（PR #207，合并提交 `1e010d7`）已合并；2026-11-17 按该提交改写。配置与模型选择的实际接口见 goal-143 状态段。`conversation.go` 的 `converse` 与 `scheduled.go` 的 `runScheduled` 都调用包内 `runToolLoop`；后者直接调用 `selected.client.Stream`，尚无重试。`generate` 独立直调 Stream，继续按 CONTRACTS §13.5 单次执行。`image.go` 已有 readImages、storedImages、replayImages、appendImages，本轮文本模型的图片目前转为占位并保存以供切换模型后回放。Provider 的非 2xx 错误当前只有脱敏文本，需要在 Provider 层保留可分类的状态与原因。可与 goal-144 并行，共享文件与合并规则见 goal-144 状态段：后合并者同步 main，保留对方改动，特别是每次工具循环、llm.generate、定时任务与图像转述的用量统计；回退按实际模型名记账。

```text
/goal 使用 $alonabot-execute 推进 goal-145（切片 T-12.17：对话重试与回退模型、图像转述）。goal 目录 goal-145，分支 goal-145，worktree .worktrees/goal-145，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：对话轮次中的 Provider 请求在可重试错误且尚未交出增量或工具调用时，同一模型最多请求 max_attempts 次，用完后按 fallback_chat_models 回退；对话模型不接受图片而设置了 image_caption_model 时，先请转述模型把图片描述成文字。

依据：spec/PLAN.md §3.1；spec/plan/agent.md T-12.17、T-12 步骤 2、3、11 及失败停止条件；spec/CONTRACTS-v0.1.md §13.3、§13.5、§13.6；spec/adr/DECISIONS.md D-38、D-86、D-132、D-154、D-155；TEST-MATRIX V-125。步骤 11 与 V-125 均规定转述失败保留占位并记 WARN，按此实现。

范围（只新增或修改这些）：
- agent/config_schema.json、agent/i18n/en-US.json、agent/i18n/zh-CN.json、agent/agent.go：追加 max_attempts、fallback_chat_models、image_caption_model、image_caption_prompt 与加载校验，保留已合入的配置字段和注册。
- agent/loop.go、agent/conversation.go、agent/scheduled.go、agent/image.go：逐次 Provider 请求的重试与回退、本轮图片转述及会话写入；允许在 agent/ 新增 retry.go、caption.go 放专属私有辅助。
- agent/provider/provider.go、chat.go、responses.go、anthropic.go、gemini.go、ollama.go：只修改重试分类所需的错误包装与状态保留，增加下述 Retryable 入口；协议编解码与 SDK 内置重试保持现状。
- agent/generate.go：仅在共享内部签名变化时同步调用，以及保留 goal-144 已有的用量统计，继续直接单次请求；goal-144 已合入时同步其 agent/usage.go 与本切片新增调用点。
- 新增 agent/retry_test.go 的一个主路径集成测试；agent/agent_test.go 只同步配置、i18n 资源快照及必要装配辅助。其余验证写在上述目录的 *_local_test.go。
允许的新跨模块边：无。重试与回退由 Agent 控制；provider 负责错误分类并继续使用现有标准库与已批准的 SDK。

接口（main 的真实签名与本切片新增入口分别列出；内部辅助调整时同步全部调用方）：
// agent 包内
func readConfiguration(values config.Values) (configuration, error)
func (s *configuration) model(name string) *configuredModel
func (a *component) converse(c *sdk.Context, text string) error
func (a *component) runScheduled(c *sdk.Context, session string, task AgentTask) error
func runToolLoop(ctx context.Context, selected *configuredModel, request provider.Request, catalog toolCatalog, emit func(string) error) error
func (a *component) generate(c *sdk.Context, req llm.Request) (llm.Response, error)
func readImages(c *sdk.Context, text string) (string, []provider.Image)
func storedImages(ctx context.Context, db *gorm.DB, session string, count uint64) (map[uint64]conversationImage, error)
func replayImages(entry conversationEntry, images map[uint64]conversationImage, enabled bool) provider.Message
func (a *component) appendImages(c *sdk.Context, db *gorm.DB, generation uint64, user conversationEntry, assistant json.RawMessage, images []provider.Image) error
// provider 包内已有
type Event struct { Text string; ToolCall *ToolCall }
func (c *Client) Stream(ctx context.Context, req Request, emit func(Event) error) (Result, error)
// provider 包内新增：保留脱敏错误及原始原因，按状态/错误类型分类
func Retryable(err error) bool

配置接口：max_attempts 为 int（缺省 3，至少 1），fallback_chat_models 为模型名列表（缺省 []，各项为 chat），image_caption_model 为可空的模型名（非空时须引用 image_input=true 的 chat），image_caption_prompt 为转述提示词，按配置原文发送。两种模型引用用 options_from @agent/llm.models 且 filter.type=chat；image_input 的引用约束在 readConfiguration 中校验。配置错误附对应 JSON 路径；configuration 持有加载后的值，复用 configuredModel 的 name、client、maxTokens、imageInput。

行为：
- 在一次 Provider 请求的外层实现重试，覆盖 /ask、自由对话与定时 Agent 任务共用的工具循环。Retryable 只接受连接失败、HTTP 408、429 与 5xx；取消、期限结束、其他 HTTP 状态、请求校验与流解析错误立即结束。Provider 层通过状态和错误链分类，避免解析错误字符串；脱敏后仍保留取消与网络错误的判断依据。两个 SDK 的内置重试仍为 0，Stream 本身仍只发一次请求。
- 每次请求先用本轮选中的模型，同一模型最多共请求 max_attempts 次；可重试的失败耗尽次数后按 fallback_chat_models 顺序尝试，每个模型同样最多 max_attempts 次。等待间隔为命名常量，初值 1s，等待与请求均响应 context 取消；每次失败记 WARN。不可重试的错误立即结束；全部耗尽时沿用一次错误回复，safe_message 与日志均遵守凭证脱敏规则。
- 按每次尝试跟踪已交出的非空文本增量或 ToolCall；交出后失败立即结束，保留已显示文本并按既有路径回复错误。重试只重发当前失败的请求，保留已有的 assistant 工具调用、ToolResults 与步数，已执行工具各执行一次；ToolCall.Replay 继续原样回传。
- 回退请求使用该模型自己的 max_tokens 与 image_input；为每次尝试按该模型重新组装图片，保留本轮原始图片与历史图片对应关系，避免从主模型已替换的占位中恢复。成功后沿用既有的流式显示、reset 代次保护与一次会话追加。
- 实际使用的对话模型 image_input=false 且配置了 image_caption_model 时，本轮每张可读图片各向转述模型请求一次：提示词用 image_caption_prompt，输出上限用转述模型的 maxTokens，非流式拼接文本，不调用工具、不套用人格或会话历史、不重试。以 [image: <描述>] 取代对应占位并写入本轮 user 条目；转述后的条目作为普通文本保存，不分配图片编号或保存待回放字节。转述失败的图片保留 [image] 并记 WARN。对话模型接受图片时直接发送图片；没有配置转述模型时沿用现有的占位、保存与回放规则。
- llm.generate、图像转述与模型测试各自只发一次请求；重试工具循环调用处与这些单次入口分开。goal-144 的用量统计已存在时，每次返回用量的对话请求与图像转述接入同一统计点，按实际模型配置名累计，避免重复；goal-144 尚未合入时在 PR「文档同步」写明实际转述入口，供后合并者接入。

测试（按编码规范「测试提交规则」）：
- 提交：agent/retry_test.go（//go:build unix，package agent_test）的一个 TestAgentRetry 集成测试，复用 agent_test.go 的真实 Start、loopback 与模拟 Provider：max_attempts=3，主模型连续三次返回 503，首个回退模型成功生成文本；经 /ask 断言收到成功回复、请求顺序与次数符合配置，Core 会话只追加一组本轮 user/assistant。主路径只保留这一条回退链。
- 本地自测（*_local_test.go，不提交，结果写入 PR）：V-125 的其余分项，包括主模型两次失败后成功、全部失败、401 立即结束、408/429/5xx 与连接失败分类、重试间隔约 1s、间隔中取消，文本或 ToolCall 交出后不重试，第二次 Provider 请求返回 429 时工具只执行一次，回退顺序和两种模型不同 image_input 的组装；转述成功的 [image: <描述>] 写入与回放排除、失败占位和 WARN、转述模型引用非法时配置加载失败、对话模型接受图片时不调用转述；自由对话和定时任务复用重试，llm.generate、转述和模型测试保持单次；max_attempts 下限及模型引用路径。goal-144 已合入时核对实际回退模型与转述的用量各记一次，否则写明该计量分项待 goal-144 接入。真实 Provider 有凭证时实测，否则写「真实 Provider 未运行」。不向 sdk/、internal/core/supervisor/ 或新 testdata 插件提交测试（D-86）。

完成条件：gofmt -l agent 无输出；go vet ./agent/...；go test -race -count=1 ./agent/...；go build ./...；go run ./cmd/alonabot check --dir agent。PR 写明实际运行平台、本地自测结果与未运行项。

不做：模型管理路由与用量表的建立（goal-144）；模型页（goal-146）；触发方式与成员独立会话（goal-159）；配置文件（goal-164）；SDK 的自动重试。
```

## 第 20 波

### goal-159 · T-12.18 Agent 的触发方式、唤醒词与成员独立会话

状态：等待前置（T-12.17）。2026-11-17 按 D-167 新增。目标、依据与要点由 PLAN T-12 步骤 2、3 确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-159（切片 T-12.18：Agent 的触发方式、唤醒词与成员独立会话）。goal 目录 goal-159，分支 goal-159，worktree .worktrees/goal-159，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：自由对话开关改为可多选的触发方式 `triggers`（mention、wake_word、private），新增唤醒词 `wake_words` 与成员独立会话 `member_sessions`；`/agent on` 与 `/agent off` 改写当前聊天的触发方式。

依据：PLAN T-12 步骤 2、3；DECISIONS D-141、D-167；TEST-MATRIX V-139。

要点：
- 非命令消息（command_consumed 为 false）命中任一选中的触发方式时触发一轮；mention 为群聊中 to_me 的消息，private 为私聊消息，wake_word 为 text 片段按顺序拼接后包含任一唤醒词（位置不限，拉丁字母不区分大小写），唤醒词保留在交给模型的文本中。
- 缺省 triggers 为空，只有 `/ask` 触发；按聊天保存的触发方式优先于配置（goal-164 改为配置文件后删除按聊天的保存）。
- member_sessions 为 true 时群聊的会话键追加发言人的用户字符串，成员各自有历史与图片编号，`/reset` 只清除发言人的会话，Actor 缺省时用群的会话；记忆的会话笔记不受影响。

不做：回复方式（Core 的平台设置，goal-162）；配置文件（goal-164）。
```

## 第 21 波

### goal-164 · T-12.19 Agent 接入配置文件

状态：等待前置（T-04.18、T-12.18）。2026-11-17 按 D-168 新增。目标、依据与要点由 PLAN T-12 步骤 13 确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-164（切片 T-12.19：Agent 接入配置文件）。goal 目录 goal-164，分支 goal-164，worktree .worktrees/goal-164，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 的对话设置按会话所用的配置文件生效；`/agent`、`/persona` 与 `/provider` 改为修改当前会话的配置文件，新增 `/profile [名字]`；删除按聊天保存的人格、模型与触发方式。

依据：PLAN T-12 步骤 13；CONTRACTS §5.6；DECISIONS D-154、D-167、D-168；TEST-MATRIX V-144。

要点：
- config_schema.json 中 chat_model、image_caption_model、人格字段、triggers、wake_words 与 member_sessions 写 `profile: true`；对话路径从 Handler 的 `c.Config()` 读取。
- 命令经 `core.profile.set` 修改，回复写明配置文件名与会话数；`reloading` 为 true 时说明 Agent 即将重载。
- `/profile` 无参数时显示当前配置文件，给出名字时经 `core.profile.assign` 切换，「默认」或 default 取消单独指定，名字不存在时列出可用的配置文件；缺省权限为 Admin。
- 删除 Agent 数据库中按聊天的设置（并入初始迁移）。

不做：配置文件的 Core 实现（goal-160、goal-161）；人格的存储（goal-167）；页面（goal-165）。
```

## 第 22 波

### goal-167 · T-12.20 官方 Agent 的插件页面

状态：等待前置（T-04.17、T-12.19）。2026-11-17 按 D-170 新增。目标、依据与要点由 PLAN T-12 步骤 14 与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-167（切片 T-12.20：官方 Agent 的插件页面）。goal 目录 goal-167，分支 goal-167，worktree .worktrees/goal-167，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：用 `sdk/page` 套件提供人格、MCP、技能、记忆与用户任务五个插件页面；命名人格从配置移到 Agent 的受管 SQLite，并提供选项服务 `agent/personas`。

依据：PLAN T-12 步骤 14；CONTRACTS §5.1（options_from）、§13.4；ADMIN-API §7；DECISIONS D-129、D-159、D-170；TEST-MATRIX V-147。画板：「v4 · 插件 · Agent · 人格」「v4 · 插件 · Agent · MCP」「v4 · 插件 · Agent · 技能」「v4 · 插件 · Agent · 用户任务（插件页面）」「v4 · 插件 · Agent · 数据（会话记录）」。

要点：
- 页面路由要求管理会话；页面经 kit.js 请求本插件的路由，样式只用 kit.css 与 `--ab-*` 令牌。
- 人格：列表（名称、系统提示摘要、使用它的配置文件与会话数）、新建与编辑对话框、删除（被引用时拒绝并列出引用）；不提供「设为默认」，配置文件使用哪个人格在配置文件页选择；没有人格时使用内置的缺省系统提示。
- MCP：各服务器的连接状态、工具列表与类别、重新连接；服务器配置仍在插件的配置标签。
- 技能：只读列出名称、描述与校验问题，可以重新读取。
- 记忆：会话笔记与用户档案的查看、编辑与删除，写入沿用字符上限与校验。
- 用户任务：Agent 建的动态任务，可删除；插件代码中的任务在「定时任务」标签。

不做：Agent 的配置标签（goal-44 的通用配置页）；插件页面的 WebUI 桥（goal-44）。
```
