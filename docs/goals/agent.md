# goal 语句：Agent

本文件收录「Agent」章节的 goal 语句，切片表在[Agent 实施计划](../../spec/plan/agent.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-58 | T-12.1 | Provider 客户端 | — | 已完成（PR #16） |
| goal-118 | T-12.9 | Provider：OpenAI Responses | T-12.1 | 待派发 |
| goal-119 | T-12.10 | Provider：Gemini 原生协议 | T-12.1 | 待派发 |
| goal-120 | T-12.11 | Provider：Ollama 原生协议 | T-12.1 | 待派发 |
| goal-59 | T-12.2 | Agent 对话与流式显示 | T-03.9、T-04.3、T-05.4、T-08.4、T-08.11、T-11.3、T-12.1、T-12.9、T-12.10、T-12.11 | 等待前置 |
| goal-83 | T-12.8 | 文本生成服务 | T-03.8、T-06.1、T-12.2 | 等待前置 |
| goal-60 | T-12.3 | 工具循环与 MCP | T-11.2、T-12.2 | 等待前置 |
| goal-61 | T-12.4 | Agent 的记忆 | T-08.5、T-11.2、T-12.2 | 等待前置 |
| goal-62 | T-12.5 | 技能 | T-12.3 | 等待前置 |
| goal-63 | T-12.6 | 定时 Agent 任务 | T-07.9、T-08.5、T-12.3 | 等待前置 |
| goal-64 | T-12.7 | 官方 Agent 组件打包与 T-12 验收 | T-12.4、T-12.5、T-12.6、T-12.8 | 等待前置 |

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

状态：待派发。前置 T-12.1（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`）已合并。2026-11-09 审查者按 main `3f7a380` 的 `agent/provider` 编写：`provider.go` 有 `Protocol` 常量 `OpenAI`、`OpenAICompatible`、`AnthropicMessages`，`New` 逐个比较协议值，`Stream` 用 `anthropic` 布尔值在 `openAIBody`/`anthropicBody` 与 `streamOpenAI`/`streamAnthropic` 之间分派，`readSSE`、`describeError`、`validateRequest`、`isJSONObject` 可以直接复用；`provider_test.go` 的 `TestStream` 是表驱动的集成测试，`mockProvider` 按协议校验路径、认证头与请求体，`testRequest` 含两条 system、带两个工具调用的 assistant 与带两个工具结果的 user 消息。可以与 goal-119、goal-120 并行，共享文件按[执行顺序 §3](../execution-order.md#3-共享文件与冲突规则)「`agent/provider/` 的新协议」一行处理。

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

状态：待派发。前置 T-12.1（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`）已合并。main `3f7a380` 的 `agent/provider` 现状见 goal-118 的状态段。可以与 goal-118、goal-120 并行，共享文件按[执行顺序 §3](../execution-order.md#3-共享文件与冲突规则)「`agent/provider/` 的新协议」一行处理。

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

状态：待派发。前置 T-12.1（[PR #16](https://github.com/AlonaBot/AlonaBot/pull/16)，合并提交 `354c5b6`）已合并。main `3f7a380` 的 `agent/provider` 现状见 goal-118 的状态段；`New` 目前要求 APIKey 非空，`readSSE` 只适用于 SSE。可以与 goal-118、goal-119 并行，共享文件按[执行顺序 §3](../execution-order.md#3-共享文件与冲突规则)「`agent/provider/` 的新协议」一行处理。

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

## 第 13 波

### goal-59 · T-12.2 Agent 对话与流式显示

状态：等待前置（T-03.9、T-05.4、T-12.9、T-12.10、T-12.11）。2026-11-09 按 D-103 加入 T-12.9～T-12.11 为前置，预设增加 gemini、ollama 改用 ollama-chat，OnLoad 接受六种协议。T-08.4（PR #74）已合并。goal-95（PR #103，合并提交 `c840af7`）已合并，合入的接口：`contracts/configschema` 的 `Parse(raw) (*Schema, []Problem)` 与 `Schema` 的 `Canonical`、`Hash`、`Version`、`Migrate(config, from)`、`Complete`、`Validate`、`SecretPaths`、`CheckI18n(config)`；`sdk/config` 的 `Prepare(ctx, load, schema, raw, fromVersion, targetVersion, restart)`、`Loader`、只读的 `Values`（`String`/`Int`/`Float`/`Bool`/`Strings`/`Object`/`Objects`/`Dict`/`Template`/`TemplateList`）与 `Template.Key`；`sdk.Base.Config() config.Values` 与根插件的 `OnConfigLoad(ctx, fromVersion uint32, raw json.RawMessage) (json.RawMessage, error)`；`supervisor.Spec.ConfigSchema` 为 config_schema.json 的规范 JSON。生成器读取手写 `config_schema.json`，配置标签解析与 `config.Config`/`config.Base` 已删除。T-04.3、T-11.3、T-12.1 已合并；命令注册、Args/Optional、Context.Match 与自动回复接口已按 PR #47 核对。Context.Edit、Continue、配置及数据库入口仍待相关前置合并后核对。语句中的配置按 D-74 的 Go 配置类型与 `config` 标签编写；D-76 改为手写 `config_schema.json` 与 `Config()` 取值后，由审查者在 T-08.11 合并后改写，providers 可以写成 template_list，每个预设一个模板，由模板 items 的 default 给出 protocol 与 base_url。

```text
/goal 使用 $alonabot-execute 推进 goal-59（切片 T-12.2：Agent 对话与流式显示）。goal 目录 goal-59，分支 goal-59，worktree .worktrees/goal-59，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立官方 Agent 组件 agent/（manifest `role: agent`）的骨架与配置（Provider 与预设、人格、历史条数等），注册 /ask、/agent on|off、/persona、/provider、/reset 与 message 事件路由；一轮对话在延续调用中执行：读取会话历史、组装请求、经 Provider 流式接收，先发送一条消息再按 1s 间隔编辑，结束时以完整内容最后编辑，超长时在段落边界换新消息；结束后追加会话条目。Provider 出错或中途断开时保留已显示内容并按 Actor locale 报告，不自动重发。

依据：PLAN T-12 步骤 2、3；CONTRACTS §13（Agent 段）、§2.4、§8（Edit）、§5.4；DECISIONS D-27、D-38；TEST-MATRIX V-57。

范围（只新增或修改这些）：
- agent/：新增 agent 包（可导入的根组件 agent.New()）、cmd/ 下的 main、plugin.yaml（`role: agent`、`icon`）、i18n/en-US.json 与 zh-CN.json、assets/icons/（组件自身的图标）；provider 包已存在，不改
- 上述包的测试文件
允许的新跨模块边：agent/... -> sdk/...、-> contracts（已允许）。不 import internal/...。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
agent
   func New() *sdk.Plugin
   type Config struct {
       config.Base
       Providers       []ProviderConfig `json:"providers"`
       DefaultProvider string           `json:"default_provider"`
       Personas        map[string]string `json:"personas"`        // 名称 → 系统提示
       DefaultPersona  string           `json:"default_persona"`
       HistoryEntries  int              `json:"history_entries" config:"default:20;min:0;max:500"`
       MessageLimit    int              `json:"message_limit" config:"default:4096;min:256;max:65536"`
       MCPServers      []MCPServer      `json:"mcp_servers"`     // T-12.3 使用
       SkillsDir       string           `json:"skills_dir"`      // T-12.5 使用
   }
   type ProviderConfig struct { Name string `json:"name"`; Preset string `json:"preset"`; Protocol string `json:"protocol"`; BaseURL string `json:"base_url"`; APIKey string `json:"api_key" config:"secret"`; Model string `json:"model"`; MaxTokens int `json:"max_tokens" config:"default:2048;min:1"` }
   type MCPServer struct { Name string `json:"name"`; Transport string `json:"transport"`; Command string `json:"command"`; Args []string `json:"args"`; Env []SecretPair `json:"env"`; URL string `json:"url"`; Headers []SecretPair `json:"headers"` }
   type SecretPair struct { Name string `json:"name"`; Value string `json:"value" config:"secret"` }

行为：
- 配置：preset 的表单选项为 openai、anthropic、gemini、deepseek、ollama 与 custom，下拉框显示名称（config 标签写成 `options:openai|anthropic|gemini|deepseek|ollama|custom;labels:OpenAI|Anthropic|Gemini|DeepSeek|Ollama|Custom`，D-74）；preset 给出 protocol 与 base_url 的缺省值（gemini 为 gemini-generate-content 与 https://generativelanguage.googleapis.com/v1beta，ollama 为 ollama-chat 与 http://localhost:11434），配置中显式给出的值优先。OnLoad 校验 providers 的 name 唯一、protocol 为 openai、openai-compatible、openai-responses、anthropic-messages、gemini-generate-content 或 ollama-chat、api_key 只在 ollama-chat 时可以为空、default_provider 与 default_persona 指向存在的项（列表为空时允许为空）；错误时返回 OnLoad 错误。
- 数据：每个聊天（adapter_instance_id、chat_id、thread_id）当前的人格、Provider 与自由对话开关保存在 Agent 的数据库（gormdb.Open，versioned 迁移）。
- 命令：/ask <text:raw>（缺省 USER）、/agent on|off（ADMIN）、/persona [name]（ADMIN，无参数时列出）、/provider [name]（ADMIN）、/reset（USER），说明用 sdk.Description 指向 messages。自由对话开启时，该聊天中 to_me 的非命令 message.created 也触发一轮（message 事件路由带 ToMe Rule，并检查开关）；私聊的消息都是 to_me。默认关闭。
- 一轮对话（Handler 调用 Continue，在延续调用中执行）：
  1) 没有可用 Provider 时用 Context.T 回复 agent.no_provider，结束。
  2) 会话键为 "<instance>/<chat>/<thread>"；ListConversation 读取最近 history_entries 条（从末尾往前，按序号顺序组装）。条目为 JSON 对象 {"role": "user"|"assistant", "speaker": 外部用户 ID（群聊时）, "content": 文本}。
  3) 以人格为 system 消息，组装 provider.Request（MaxTokens 取 Provider 配置）。
  4) 事件的 capabilities 含 message.edit 时：发送一条占位消息（agent.thinking），之后流式接收：每累计到新的文本时，距上次编辑不少于 1s（命名常量）才 Edit；Edit 返回限流错误时把下一次编辑推迟 retry_after；生成结束后用完整内容最后编辑一次。不含 message.edit 时：不发占位消息，不发送中间增量，生成结束后一次发送完整内容。回复以纯文本发送（不使用 markdown 格式）。
  5) 当前消息的内容将超过 message_limit 时，在最近的段落边界（"\n\n"，没有时为 "\n"，再没有时为空白，最后硬切）结束当前消息，再发送新消息继续显示；不能编辑时按同样的切分发送多条。
  6) 结束后用 AppendConversation 追加本轮的 user 与 assistant 条目。
  7) Provider 返回错误或中途断开：保留已显示的内容（不能编辑时不发送已生成的部分），在其后编辑或发送 agent.error（按 Actor locale，参数为错误码），不追加 assistant 条目，不重发请求。
- /reset：ClearConversation 清除该会话。
- 工具、MCP、记忆与技能分别在 T-12.3、T-12.4、T-12.5 接入；本切片不调用它们。

测试（按编码规范「测试提交规则」）：
- 提交：agent/agent_test.go（//go:build unix）。用本仓库 Core 的测试装配（supervisor、memory Adapter，实现 Editor 与 Named）、真实状态库与本地模拟 Provider（openai、openai-compatible 与 anthropic-messages 各一次；其他协议的流解析由 TestStream 覆盖）：
  - /ask hello：memory 先收到一条发送，随后收到若干次编辑，最后一次编辑是完整回复；编辑间隔不小于 1s（用注入时钟或计时断言）。
  - 实例不声明 message.edit 时：memory 只收到完整回复的发送（超过 message_limit 时为多条），没有占位消息与编辑。
  - 模拟 Provider 中途断开：已显示内容保留，最后有错误说明，会话中没有 assistant 条目，Provider 只收到一次请求。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：段落切分函数；OnLoad 的校验；限流推迟；会话条目格式；配置 Schema 由生成器生成且 api_key 为 secret；回复超过 message_limit（测试中设为 256）时分成多条消息，切分点在段落边界；自由对话关闭时群聊中的 to_me 消息不触发；Admin /agent on 后触发；非 to_me 的消息不触发；/persona 切换后请求的 system 消息改变；/reset 后历史为空；普通用户执行 /agent 被拒（不回复）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... ./internal/core/...；go build ./...；在 agent 运行 alonabot check 通过。

不做：工具循环与 MCP（T-12.3）；记忆（T-12.4）；技能（T-12.5）；定时任务（T-12.6）；打包（T-12.7）；真实 Provider（T-10）。
```

## 第 14 波

### goal-83 · T-12.8 文本生成服务

状态：等待前置（T-03.8、T-12.2）。T-06.1（PR #45）已合并。语句中 goal-8、goal-19、goal-59 提供的符号（manifest 的 role 解析、Service Broker 的目标定位、sdk.JSONService、sdk.CallJSON、sdk.WithAccess、Context.Caller、agent 的 Config 与 Provider 装配）只写名字；合并后由审查者核对，再改为待派发。2026-10-28 按 D-61 改为以 `@agent` 为目标。2026-11-04 按 D-77～D-85 修订计划：llm_service_allowed_plugins 缺省为 `["*"]`，表示全部插件（D-83）；Agent 的配置改由 `config_schema.json` 声明（D-76）。2026-11-09 按 D-111 在 sdk/llm 增加一问一答的 `llm.Ask`。

```text
/goal 使用 $alonabot-execute 推进 goal-83（切片 T-12.8：文本生成服务）。goal 目录 goal-83，分支 goal-83，worktree .worktrees/goal-83，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 的 Service Broker 把目标 `@agent` 解析为已安装的 Agent 组件；新建 sdk/llm，提供 CONTRACTS §13.5 llm.generate 的请求/响应类型、Schema 标识、以 `@agent` 为目标的调用辅助与一问一答的 llm.Ask；官方 Agent 组件注册该服务（public），只接受配置项 llm_service_allowed_plugins 中的直接调用方，用已配置的 Provider 完成一次非流式请求，并返回 Provider 给出的用量。

依据：PLAN T-12 步骤 9；CONTRACTS §13.5、§3、§9.1；DECISIONS D-38、D-51、D-61、D-111；TEST-MATRIX V-76、V-82、V-109。

范围（只新增或修改这些）：
- internal/core/ 的 Service Broker：目标定位处解析 `@agent`（只改定位处，新增解析函数放在新文件）
- sdk/llm/：新建包（类型、Schema 标识、Generate、Ask）
- agent/：Config 增加字段；新增 generate.go（服务注册与处理）；provider 包不改
- 上述包的测试文件
允许的新跨模块边：无（agent -> sdk/... 已允许）。sdk/llm 只 import sdk 与标准库。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
1. sdk/llm
   const ServiceName = "llm.generate"
   const AgentTarget = "@agent" // CONTRACTS §3：由 Core 解析为已安装的 Agent 组件
   const SchemaIn, SchemaOut = "...", "..." // 随包固定的 schema 标识，按 goal-19 的命名规则选定并写入包文档
   const SchemaVersion uint32 = 1
   type Message struct { Role string `json:"role"`; Text string `json:"text"` } // Role 为 "user" 或 "assistant"
   type Request struct { Messages []Message `json:"messages"`; System string `json:"system,omitempty"`; Provider string `json:"provider,omitempty"`; MaxOutputTokens int `json:"max_output_tokens,omitempty"` }
   type Response struct { Text string `json:"text"`; Provider string `json:"provider"`; Model string `json:"model"`; InputTokens int `json:"input_tokens,omitempty"`; OutputTokens int `json:"output_tokens,omitempty"` }
   func Generate(ctx context.Context, req Request) (Response, error) // 经 sdk.CallJSON 调用 AgentTarget + "/" + ServiceName
   func Ask(ctx context.Context, prompt string) (string, error) // 等于 Generate(ctx, Request{Messages: []Message{{Role: "user", Text: prompt}}}) 后取 Text（D-111）
2. agent
   Config 增加字段：LLMServiceAllowedPlugins []string `json:"llm_service_allowed_plugins"`

行为：
- `@agent`：Broker 定位目标时，`@agent` 解析为当前安装记录（或库模式的 dev 组件）中 role 为 agent 的组件 ID，之后的访问规则、调用链、CALL_CYCLE 与激活检查按解析后的 ID 执行；没有 Agent 组件时与目标未激活相同。
- Ask：prompt 为空串或只含空白时返回 INVALID_ARGUMENT，不发出调用；否则按接口段的等价写法调用 Generate，只返回 Response.Text，错误原样返回。
- 注册：在 Services Hook 中以 JSONService 注册 ServiceName，WithAccess(sdk.AccessPublic)。
- 许可：Context.Caller() 为空或不在 LLMServiceAllowedPlugins 中时返回 PERMISSION_DENIED，不请求 Provider。
- 校验：messages 为空、role 不是 user 或 assistant、max_output_tokens 小于 0 时返回 INVALID_ARGUMENT；provider 非空且不存在时返回 INVALID_ARGUMENT；为空时使用 default_provider，没有任何可用 Provider 时返回 FAILED_PRECONDITION。
- 请求：把 system 与 messages 组装成 provider.Request，MaxTokens 取 max_output_tokens（为 0 时取该 Provider 配置的 max_tokens），不传工具；用 Provider 客户端的流式接口接收并拼接全部文本，收到终止标记才算成功；用量取自 provider.Result.Usage，为 nil 时不填写。
- 错误：Provider 返回错误或中途断开时返回 UNAVAILABLE，safe_message 不含 API Key 与消息内容（provider 返回的错误说明只写日志，不放进 safe_message）；不自动重试（D-38）。调用方期限到达时取消 Provider 请求。
- 不读写会话存储与记忆，不套用人格，不调用工具。
- 日志：每次调用以 INFO 记录调用方插件、provider、model、耗时与用量；不记录消息文本。

测试（按编码规范「测试提交规则」）：
- 提交：agent/generate_test.go（//go:build unix）。用 goal-59 的测试装配、本地模拟 Provider 与一个调用 llm.Generate 的测试插件：
  - 调用方未列入时 PERMISSION_DENIED，模拟 Provider 没有收到请求；列入后返回 text、provider、model 与模拟服务给出的用量，会话存储没有新增条目。
  - 测试插件只经 llm.Generate（`@agent`）调用，不知道 Agent 的组件 ID；把 Agent 组件换成另一个 ID 的测试 Agent 后同一调用到达新组件。
  - 测试插件以 llm.Ask 调用一次：模拟 Provider 收到只含一条 user 消息、没有 system 的请求，使用 default_provider，Ask 返回模拟服务给出的文本。
  - 模拟 Provider 出错与中途断开为 UNAVAILABLE，模拟服务只收到一次请求，safe_message 不含 API Key。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：校验分支；Ask 的 prompt 为空串与只含空白时不发出调用；三种协议各一次；provider 不存在为 INVALID_ARGUMENT；Agent 配置中没有 Provider 为 FAILED_PRECONDITION；调用方期限短于模拟 Provider 的延迟时，调用方得到 DEADLINE_EXCEEDED，模拟服务看到请求被取消；Agent 组件未启用或未安装时调用返回 UNAVAILABLE。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... ./sdk/... ./internal/core/...；go build ./...；`go list -deps` 列出的 Core 可执行文件与 SDK 根包模块都在 docs/development.md 的允许清单中。

不做：流式输出与工具调用（契约不提供）；llm_service_allowed_plugins 的专用管理页（配置表单随 T-09 由 Schema 生成）。
```

## 第 15 波

### goal-60 · T-12.3 工具循环与 MCP

状态：等待前置（T-11.2、T-12.2）。sdk.ListTools 已由 goal-54（PR #101，合并提交 `de48f26`）合入。语句中 goal-55、goal-58、goal-59 提供的符号（sdk.InvokeTool、sdk.ConfirmOperation、provider 的 Tool/ToolCall/ToolResult、agent 的一轮对话与 Config.MCPServers）只写名字；合并后由审查者核对，再改为待派发。本切片获授权引入已批准的 github.com/modelcontextprotocol/go-sdk（docs/development.md 只授权给 goal-60）。

```text
/goal 使用 $alonabot-execute 推进 goal-60（切片 T-12.3：工具循环与 MCP）。goal 目录 goal-60，分支 goal-60，worktree .worktrees/goal-60，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 组件每轮把 core.tool.list 返回的工具与已连接 MCP 服务器的工具交给模型；工具名按 Provider 的命名规则映射且一轮内稳定（六种协议都要覆盖，Gemini 的函数名规则最严）；模型交出的 provider.ToolCall 原样（含 Replay）放回下一次请求的 assistant 消息；Core 工具经 core.tool.invoke 调用，MCP 工具先经 core.tool.confirm（操作名 mcp/<server>/<tool>）再调用；拒绝、超时与工具错误作为工具结果交回模型；单轮工具步数达到上限时结束本轮并告知用户；MCP 服务器首次使用时连接、随插件停止关闭，stdio 子进程属于插件的进程树。

依据：PLAN T-12 步骤 4；CONTRACTS §13.1、§13.2、§7.1（进程树）；DECISIONS D-27、D-38；docs/development.md（MCP go-sdk 的批准范围）；TEST-MATRIX V-58。

范围（只新增或修改这些）：
- go.mod、go.sum：新增 github.com/modelcontextprotocol/go-sdk（取当前最新的 v1 版本，PR 中写明版本，并用 go mod why 说明它只被 agent 使用）
- agent/：新增 tools.go（工具目录与名称映射）、mcp.go（MCP 客户端管理）、loop.go（工具循环）；一轮对话接入循环（只改调用处）
- 新增 agent/mcptest/（stdio 与 streamable HTTP 的 MCP 模拟服务器，测试用）
- i18n：en-US.json 与 zh-CN.json 增加步数上限与工具错误的文案
- 上述包的测试文件
允许的新跨模块边：只有 agent 可以 import MCP go-sdk。

接口：agent 不新增导出符号（Config.MCPServers 已在 goal-59 声明）。mcptest 包导出测试所需的启动函数，由执行者在包文档列出。

行为：
- 工具目录：每轮开始时调用 sdk.ListTools，并对已连接（或首次使用时连接）的 MCP 服务器调用 tools/list；合并为一个列表交给模型。名称映射：Provider 只接受 ^[a-zA-Z0-9_-]{1,64}$ 的名字（两种协议都按此规则，命名常量），把 "plugin_id/service_name" 与 "mcp/<server>/<tool>" 转为合规名字（非法字符替换为 _，冲突或超长时加 _ 与短哈希）；映射在一轮内固定，结果交回模型时用映射后的名字。
- 循环：模型返回工具调用时依次执行，结果作为 ToolResult 交回，再继续请求；文本照常流式显示（沿用 goal-59）。单轮工具步数上限为命名常量（初值 20），达到时停止请求模型，回复 agent.tool_limit。
- Core 工具：sdk.InvokeTool（在延续调用内，继承父调用链与 Actor）；错误（PERMISSION_DENIED、DEADLINE_EXCEEDED、工具错误等）转为文本结果交回模型（"error: <code>: <safe_message>"），不重试。
- MCP 工具：先 sdk.ConfirmOperation("mcp/<server>/<tool>", 参数的 JSON 摘要)，允许后经 go-sdk 调用；拒绝与超时同样作为结果交回。MCP 工具的参数在交给模型前就以其 inputSchema 暴露，调用前不在 Agent 侧另做 Schema 校验（由服务器校验）。
- MCP 连接：stdio 以 command 与 args 启动子进程，env 为配置中解析后的 secret 值（SecretPair），子进程在插件的进程组内（由 Core 的进程树管理回收）；streamable HTTP 以 url 与 headers 连接。首次使用时连接，失败时该服务器本轮不可用并以 WARN 记录（不阻止对话）；插件 OnStop 时关闭全部会话并等待 stdio 子进程退出。
- 安全：MCP 的 env 与 headers 中的值不写日志、不进入工具结果。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tools_test.go（//go:build unix）。用 goal-59 的测试装配、模拟 Provider（脚本化地返回工具调用）、在测试临时目录生成的工具插件（goal-54、goal-55 的同类插件只在执行者本地）与 mcptest 的两种服务器：
  - 模型连续调用一个 Core 工具与一个 MCP 工具：两者都执行，结果交回模型，最终回复显示。
  - Core 工具为 confirm 策略且用户拒绝：结果为错误文本交回模型，工具未执行；MCP 工具的确认被拒同样如此。
  - 停止插件后 stdio MCP 子进程已退出（按 PID 检查）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：名称映射函数；MCP 连接失败时对话继续；secret 值不出现在日志中；HTTP 服务器的 headers；工具出错：错误作为结果交回，不重试；模型持续请求工具超过 20 步：本轮结束并回复步数上限文案；名称映射对含 / 与 . 的名字、冲突名字的结果稳定。

完成条件：gofmt -l 无输出；go mod tidy 后无差异；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：记忆（T-12.4）；技能（T-12.5）；MCP 的资源与提示词能力（只用 tools）；真实 Provider（T-10）。
```

### goal-61 · T-12.4 Agent 的记忆

状态：等待前置（T-11.2、T-12.2）。T-08.5（PR #79）已合并。`sdk.JSONService` 的 `Tool` 方法已由 goal-54（PR #101，合并提交 `de48f26`）合入。语句中 goal-19、goal-32、goal-59 提供的符号（sdk.JSONService、gormdb.Open、agent 的一轮对话）只写名字；合并后由审查者核对，再改为待派发。2026-10-28 按 D-60 改写：记忆并入 Agent，不再有 sdk/memory 与记忆插件。

```text
/goal 使用 $alonabot-execute 推进 goal-61（切片 T-12.4：Agent 的记忆）。goal 目录 goal-61，分支 goal-61，worktree .worktrees/goal-61，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：官方 Agent 组件 agent/ 实现 Hermes 式的记忆：用 gormdb 在 Agent 的受管 SQLite 中保存会话笔记（按默认 Target）与用户档案（按 Actor）；注册 memory 工具（allow）执行原子的 add/patch/remove；一轮对话在会话第一轮把会话笔记写为快照条目，每轮把当前 Actor 的档案附在用户消息之前。

依据：PLAN T-12 步骤 3、5；CONTRACTS §13.1、§13.4、§3；DECISIONS D-60；TEST-MATRIX V-59。

范围（只新增或修改这些）：
- agent/：新增 memory.go（存储、工具与注入）、迁移追加记忆表；一轮对话只在调用处加两处调用；根组件注册 memory 工具（只改注册处）
- 上述包的测试文件
允许的新跨模块边：无（agent/... -> sdk/... 已允许）。

接口：agent 不新增导出符号。

行为：
- 范围：会话按默认 Target（adapter_instance_id、chat_id、thread_id），用户按 Actor（adapter_instance_id、external_user_id），都取自 Core 下发的可信上下文；memory 工具经 core.tool.invoke 调用时取工具调用的上下文，不采信参数中的范围。缺少所需的部分时返回 FAILED_PRECONDITION。
- memory 工具（JSON 服务加 sdk.Tool，default_policy=allow，只允许本组件经 core.tool.invoke 调用）：输入 {"operations": [{"op": "add"|"patch"|"remove", "target": "memory"|"user", "text": "...", "old_text": "..."}]}；patch 与 remove 用唯一的 old_text 定位（找不到或不唯一 → INVALID_ARGUMENT）；整组按结果检查上限（会话笔记 2,200、用户档案 1,375 个 Unicode 码点，命名常量）后在一个事务中写入；超限、完全重复的条目、含 Unicode Cf 类字符的内容返回错误，已有条目不变。条目按写入顺序拼接为文本（每条一行）。
- 一轮对话：
  - 会话的第一轮（会话为空时）读取会话笔记，作为 {"role": "system", "kind": "memory_snapshot", "content": ...} 写入会话；之后的轮次使用会话中的快照，不再读取（会话内快照不变、/reset 后下一会话重新取）。
  - 每轮读取当前 Actor 的档案，附在本轮用户消息之前。
  - 读取记忆失败时本轮按 agent.memory_error 报错并结束，不跳过记忆继续。
  - 模型的写入经工具循环调用 memory 工具（T-12.3 已接入工具循环；若本切片先于 goal-60 合并，工具调用留到 goal-60 之后验证）。

测试（按编码规范「测试提交规则」）：
- 提交：agent/memory_test.go（//go:build unix）。经真实 Core 装配与模拟 Provider：
  - 经 core.tool.invoke 调用 memory 工具：add、patch、remove 正确；超限时返回错误且已有条目不变；重复条目与含 U+200B 的内容被拒；群聊中两位用户各自只看到自己的档案；管理员把 memory 工具设为 deny 后调用返回 PERMISSION_DENIED。
  - 第一轮会话写入快照条目，第二轮请求中的快照与第一轮相同（会话内写入不进入快照）；/reset 后重新取快照；模拟 Provider 收到的本轮用户消息前附有当前 Actor 的档案。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：码点计数（含组合字符与 emoji）；old_text 不唯一；空操作组；缺少 Actor 或默认 Target 的 FAILED_PRECONDITION。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/...；go build ./...；在 agent 运行 alonabot check 通过。

不做：技能（T-12.5）；定时任务（T-12.6）；向量检索或外部记忆后端。
```

## 第 16 波

### goal-62 · T-12.5 技能

状态：等待前置（T-12.3）。语句中 goal-59、goal-60 提供的符号（agent 的 Config.SkillsDir、工具目录与循环）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-62（切片 T-12.5：技能）。goal 目录 goal-62，分支 goal-62，worktree .worktrees/goal-62，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 组件启动时读取技能目录下各子目录的 SKILL.md：YAML frontmatter 的 name 与 description 必填，其余内容为正文；系统提示列出各技能的名称与描述，load_skill 工具返回正文；不合法的文件记诊断并跳过，不阻止插件启动。

依据：PLAN T-12 步骤 6；CONTRACTS §13（Agent 段）；TEST-MATRIX V-60。

范围（只新增或修改这些）：
- agent/：新增 skills.go；系统提示与工具目录接入技能（只改调用处）
- 上述包的测试文件
允许的新跨模块边：无（YAML 用已批准的 go.yaml.in/yaml/v3）。

行为：
- 目录：Config.SkillsDir 为空时不启用技能；相对路径以插件工作目录为基准。只读取一级子目录中的 SKILL.md（文件名区分大小写），不跟随符号链接，单个文件上限 256 KiB（命名常量），超过则视为不合法。
- 解析：文件以 "---\n" 开头的 frontmatter 到下一个 "---\n" 为止，用 yaml.v3 严格解码（未知字段允许并忽略）；name 为 [a-z0-9_-]{1,64}，description 为 1～1024 字符的非空字符串；同名技能以目录名字典序第一个为准，其余记诊断。不合法的文件以 WARN 记录（路径与原因）并跳过。
- 系统提示：在人格提示之后追加一段（文案键 agent.skills.header），逐行列出「name: description」，按 name 排序；没有技能时不追加。
- load_skill：插件内工具（不经 Core 的工具目录与策略），在 goal-60 的工具列表中加入，输入 {"name": string}，返回正文；名字不存在时返回错误文本 "error: NOT_FOUND: unknown skill"。不计入额外的确认。
- 技能在插件启动时读取一次，修改后需要重启插件。

测试（按编码规范「测试提交规则」）：
- 提交：agent/skills_test.go（//go:build unix）：临时技能目录含两个合法技能、一个缺少 frontmatter、一个缺少 description、一个超过大小上限：模拟 Provider 收到的系统提示只含两个合法技能，日志有三条诊断，插件正常启动；模型调用 load_skill 得到正文；不存在的名字得到错误结果。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：frontmatter 的边界（CRLF、没有结束分隔）；同名技能；符号链接被忽略；name 的字符规则。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：技能的热更新；从网络下载技能。
```

### goal-63 · T-12.6 定时 Agent 任务

状态：等待前置（T-07.9、T-12.3）。T-08.5（PR #79）已合并。2026-11-10 按 D-124、D-125 修订：任务改用 SDK 的动态任务保存与触发，删去自建的任务表与每分钟的固定 schedule，前置由 T-07.2 改为 T-07.9，并增加一次性的 at。语句中 goal-128、goal-32、goal-55、goal-59、goal-60 提供的符号（sdk.Task、sdk.OnTask、Task 的 At/Every/Cron/Cancel/List、sdk.Key、sdk.Timezone、sdk.ConfirmOperation、Agent 的一轮对话与工具循环）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：定时任务的后台发送按发信许可缺省允许全部实例（D-83）。

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
- schedule_task：输入 {"prompt", "at"?, "interval_seconds"?, "cron"?, "timezone"?}，at（RFC 3339）、interval 与 cron 恰有一个，timezone 只随 cron；interval 至少 60s（命名常量），在调用 Core 之前检查。调用者（工具循环的当前 Actor）必须含 ADMIN 角色，否则返回错误结果 "error: PERMISSION_DENIED"；先 ConfirmOperation("agent/schedule_task", 摘要)，允许后按触发规则调用 At、Every 或 Cron（cron 带 sdk.Timezone），返回值中的 key 作为任务 ID 交回模型。Target 为当前对话的默认 Target。Core 返回的 INVALID_ARGUMENT（表达式不合法、at 不晚于当前时间）与 RESOURCE_EXHAUSTED 作为错误结果交回模型。
- list_tasks：调用 List，只返回 target 与当前默认 Target（instance、chat、thread）相同的任务，含 ID、提示词、触发规则与下一次触发时刻；任何人可用。
- cancel_task：输入 {"id"}；只有 ADMIN，先确认；先用 List 确认该 ID 属于本聊天，再调用 Cancel；不属于本聊天或不存在时返回错误结果 "error: NOT_FOUND"。
- 处理函数运行一轮对话：
  - 没有 Actor，只有 job scope；会话键为 "task/<key>"，与聊天会话分开；不调用记忆服务。
  - 回复经 sdk.Send 显式 To 任务的 Target，需要 allowed_adapter_instances，没有许可时 PERMISSION_DENIED 并记 WARN（周期任务保留）。
  - 工具循环中 confirm 策略的 Core 工具与全部 MCP 工具一律拒绝（不发起确认，因为没有 Actor），作为错误结果交回模型。
  - 一次运行受 JOB_TIMEOUT 约束；同一任务上一轮未结束时由 Core 跳过本轮。
  - 一次性任务在派发时已由 Core 删除，对话失败不重试，失败写日志。
- 插件离线期间到期的触发由 Core 作废：周期任务从下一个未来时刻继续，一次性任务删除。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tasks_test.go（//go:build unix）：用模拟 Provider（脚本化调用 schedule_task 等工具）与注入时钟的 Scheduler：
  - 非 Admin 调用 schedule_task 得到错误结果，List 为空；Admin 经确认后创建 at、interval 与 cron 任务各一个。
  - 时钟推进到到期：任务运行一轮，回复发往任务 Target；没有发信许可时没有发送并记 WARN，授予许可后下一次运行发送；at 任务运行一次后 List 中不再出现。
  - 停止插件、推进时钟越过多次触发、再启动：不补跑，周期任务只在下一次未来时刻运行。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：interval 下限；三种触发规则互斥；会话键隔离；后台一轮中模型调用 confirm 策略的工具得到拒绝结果；没有注入记忆（请求中没有快照与档案）；其他用户 list_tasks 只看到本聊天的任务；cancel_task 需要 Admin 与确认，不能取消其他聊天的任务；Agent 的数据库中没有任务表。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：打包与 T-12 验收（T-12.7）；任务的管理页面。
```

## 第 17 波

### goal-64 · T-12.7 官方 Agent 组件打包与 T-12 验收

状态：等待前置（T-12.4、T-12.5、T-12.6、T-12.8）。语句中 goal-38、goal-58～goal-63、goal-83 提供的命令与符号（alonabot pack、Agent 组件、mcptest、模拟 Provider）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：llm_service_allowed_plugins 缺省为 `["*"]`（D-83）。

```text
/goal 使用 $alonabot-execute 推进 goal-64（切片 T-12.7：官方 Agent 组件打包与 T-12 验收）。goal 目录 goal-64，分支 goal-64，worktree .worktrees/goal-64，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：官方 Agent 组件（`role: agent`）按 T-09 的打包流程生成可安装包；用本地 Provider 与 MCP 模拟服务、回环测试 Adapter（或 Telegram 模拟服务）以安装包方式运行 V-57～V-61 与 V-76 的全部场景；在 TEST-MATRIX §3.2 记录 T-12 证据，真实 Provider 与真实 Telegram 没有凭证时记为未运行。

依据：PLAN T-12 步骤 8 与「停止和失败条件」；CONTRACTS §13、§9.3；TEST-MATRIX V-57～V-61、V-76。

范围（只新增或修改这些）：
- agent/plugin.yaml：补齐打包所需字段（version、source、os/arch 由 pack 填写的部分除外）
- 新增 cmd/alonabot/t12_e2e_test.go（//go:build unix）
- spec/TEST-MATRIX.md §3.2：追加 T-12 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

验收矩阵（以 alonabot pack 产生的安装包经 alonabot run 与管理 API 安装、启用后运行；各为一个子测试）：
- V-57：openai、openai-compatible 与 anthropic-messages 三种协议的流式文本、工具调用、停止原因与用量（其他协议的流解析由 TestStream 覆盖）；/ask 与自由对话开关；群聊非 to_me；/provider 与人格切换；/reset；超长回复分段；Provider 出错或中途断开。
- V-58：连续调用 Core 工具与 MCP 工具（stdio、streamable HTTP）；确认拒绝；工具出错；超过单轮步数上限；MCP 进程随 Agent 停止。
- V-59：记忆 add/patch/remove、超限、重复、不可见字符、群聊多用户、会话中写入与 /reset、memory 工具设为 confirm 与 deny。
- V-60：合法与不合法的 SKILL.md、load_skill 不存在的技能。
- V-61：非 Admin 创建任务、interval/cron 任务到期运行、无发信许可、confirm 工具、插件离线错过。
- V-76：另一个测试插件经安装包调用 llm.generate：未列入与列入 llm_service_allowed_plugins、Provider 不存在、Provider 出错、Agent 未启用。
- 证据：TEST-MATRIX §3.2 追加 T-12 一行，写明测试位置、命令、平台与结果；Provider 与 MCP 标为本地模拟；真实 Provider 与 Telegram 为「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t12_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：Agent 组件的 alonabot check 与 pack（当前平台）；若维护者提供了 Provider 凭证，另附一次手动运行记录，不写入提交的测试。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：发行归档与预装（T-14.3、T-14.4）。
```
