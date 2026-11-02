# goal 语句：Agent

本文件收录「Agent」章节的 goal 语句，切片表在[Agent 实施计划](../../spec/plan/agent.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-58 | T-12.1 | Provider 客户端 | — | 已完成 |
| goal-59 | T-12.2 | Agent 对话与流式显示 | T-03.9、T-04.3、T-05.4、T-08.3、T-08.4、T-11.3、T-12.1 | 等待前置 |
| goal-83 | T-12.8 | 文本生成服务 | T-03.8、T-06.1、T-12.2 | 等待前置 |
| goal-60 | T-12.3 | 工具循环与 MCP | T-11.2、T-12.2 | 等待前置 |
| goal-61 | T-12.4 | Agent 的记忆 | T-08.5、T-11.2、T-12.2 | 等待前置 |
| goal-62 | T-12.5 | 技能 | T-12.3 | 等待前置 |
| goal-63 | T-12.6 | 定时 Agent 任务 | T-07.2、T-08.5、T-12.3 | 等待前置 |
| goal-64 | T-12.7 | 官方 Agent 组件打包与 T-12 验收 | T-12.4、T-12.5、T-12.6、T-12.8 | 等待前置 |

## 第 1 波

### goal-58 · T-12.1 Provider 客户端

状态：已完成（PR #16 已合并；2026-10-28 按 D-64 修正：增加 `openai` 协议、停止原因与用量，包移到 `agent/provider`）。本 goal 的原始语句未保留；下面记录当前的实际接口，供后继切片引用。

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


## 第 12 波

### goal-59 · T-12.2 Agent 对话与流式显示

状态：等待前置（T-03.9、T-05.4、T-08.3、T-08.4）。T-04.3、T-11.3、T-12.1 已合并；命令注册、Args/Optional、Context.Match 与自动回复接口已按 PR #47 核对。Context.Edit、Continue、配置及数据库入口仍待相关前置合并后核对。

```text
/goal 使用 $alonabot-execute 推进 goal-59（切片 T-12.2：Agent 对话与流式显示）。goal 目录 goal-59，分支 goal-59，worktree .worktrees/goal-59，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立官方 Agent 组件 agent/（manifest `role: agent`）的骨架与配置（Provider 与预设、人格、历史条数等），注册 /ask、/agent on|off、/persona、/provider、/reset 与 message 事件路由；一轮对话在延续调用中执行：读取会话历史、组装请求、经 Provider 流式接收，先发送一条消息再按 1s 间隔编辑，结束时以完整内容最后编辑，超长时在段落边界换新消息；结束后追加会话条目。Provider 出错或中途断开时保留已显示内容并按 Actor locale 报告，不自动重发。

依据：PLAN T-12 步骤 2、3；CONTRACTS §13（Agent 段）、§2.4、§8（Edit）、§5.4；DECISIONS D-27、D-38；TEST-MATRIX V-57。

范围（只新增或修改这些）：
- agent/：新增 agent 包（可导入的根组件 agent.New()）、cmd/ 下的 main、plugin.yaml（`role: agent`、`icon`）、i18n/en-US.json 与 zh-CN.json、assets/icons/（Provider 预设与组件自身的图标）；provider 包已存在，不改
- 上述包的测试文件
允许的新跨模块边：agent/... -> sdk/...、-> contracts（已允许）。不 import internal/...。

接口（照此实现，不增减公开符号；内部辅助函数自定）：
agent
   func New() sdk.Root
   type Config struct {
       config.Base
       Providers       []ProviderConfig `json:"providers"`
       DefaultProvider string           `json:"default_provider"`
       Personas        map[string]string `json:"personas"`        // 名称 → 系统提示
       DefaultPersona  string           `json:"default_persona"`
       HistoryEntries  int              `json:"history_entries" form:"default=20;min=0;max=500"`
       MessageLimit    int              `json:"message_limit" form:"default=4096;min=256;max=65536"`
       MCPServers      []MCPServer      `json:"mcp_servers"`     // T-12.3 使用
       SkillsDir       string           `json:"skills_dir"`      // T-12.5 使用
   }
   type ProviderConfig struct { Name string `json:"name"`; Preset string `json:"preset"`; Protocol string `json:"protocol"`; BaseURL string `json:"base_url"`; APIKey string `json:"api_key" secret:"true"`; Model string `json:"model"`; MaxTokens int `json:"max_tokens" form:"default=2048;min=1"` }
   type MCPServer struct { Name string `json:"name"`; Transport string `json:"transport"`; Command string `json:"command"`; Args []string `json:"args"`; Env []SecretPair `json:"env"`; URL string `json:"url"`; Headers []SecretPair `json:"headers"` }
   type SecretPair struct { Name string `json:"name"`; Value string `json:"value" secret:"true"` }

行为：
- 配置：preset 的表单选项为 openai、anthropic、deepseek、ollama 与 custom，各附 assets/icons/ 下的图标（选项图标的 tag 写法以 goal-28 合并后的生成器为准）；preset 给出 protocol 与 base_url 的缺省值，配置中显式给出的值优先。OnLoad 校验 providers 的 name 唯一、protocol 为 openai、openai-compatible 或 anthropic-messages、default_provider 与 default_persona 指向存在的项（列表为空时允许为空）；错误时返回 OnLoad 错误。
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
- 提交：agent/agent_test.go（//go:build unix）。用本仓库 Core 的测试装配（supervisor、memory Adapter，实现 Editor 与 Named）、真实状态库与 agent/provider 的本地模拟 Provider（三种协议各一次）：
  - /ask hello：memory 先收到一条发送，随后收到若干次编辑，最后一次编辑是完整回复；编辑间隔不小于 1s（用注入时钟或计时断言）。
  - 回复超过 message_limit（测试中设为 256）时分成多条消息，切分点在段落边界。
  - 实例不声明 message.edit 时：memory 只收到完整回复的发送（超过 message_limit 时为多条），没有占位消息与编辑。
  - 模拟 Provider 中途断开：已显示内容保留，最后有错误说明，会话中没有 assistant 条目，Provider 只收到一次请求。
  - 自由对话关闭时群聊中的 to_me 消息不触发；Admin /agent on 后触发；非 to_me 的消息不触发。
  - /persona 切换后请求的 system 消息改变；/reset 后历史为空。
  - 普通用户执行 /agent 被拒（不回复）。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：段落切分函数；OnLoad 的校验；限流推迟；会话条目格式；配置 Schema 由生成器生成且 api_key 为 secret。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... ./internal/core/...；go build ./...；在 agent 运行 alonabot plugin check 通过。

不做：工具循环与 MCP（T-12.3）；记忆（T-12.4）；技能（T-12.5）；定时任务（T-12.6）；打包（T-12.7）；真实 Provider（T-10）。
```


## 第 13 波

### goal-83 · T-12.8 文本生成服务

状态：等待前置（T-03.8、T-06.1、T-12.2）。语句中 goal-8、goal-19、goal-59 提供的符号（manifest 的 role 解析、Service Broker 的目标定位、sdk.JSONService、sdk.CallJSON、sdk.WithAccess、Context.Caller、agent 的 Config 与 Provider 装配）只写名字；合并后由审查者核对，再改为待派发。2026-10-28 按 D-61 改为以 `@agent` 为目标。

```text
/goal 使用 $alonabot-execute 推进 goal-83（切片 T-12.8：文本生成服务）。goal 目录 goal-83，分支 goal-83，worktree .worktrees/goal-83，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Core 的 Service Broker 把目标 `@agent` 解析为已安装的 Agent 组件；新建 sdk/llm，提供 CONTRACTS §13.5 llm.generate 的请求/响应类型、Schema 标识与以 `@agent` 为目标的调用辅助；官方 Agent 组件注册该服务（public），只接受配置项 llm_service_allowed_plugins 中的直接调用方，用已配置的 Provider 完成一次非流式请求，并返回 Provider 给出的用量。

依据：PLAN T-12 步骤 9；CONTRACTS §13.5、§3、§9.1；DECISIONS D-38、D-51、D-61；TEST-MATRIX V-76、V-82。

范围（只新增或修改这些）：
- internal/core/ 的 Service Broker：目标定位处解析 `@agent`（只改定位处，新增解析函数放在新文件）
- sdk/llm/：新建包（类型、Schema 标识、Generate）
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
2. agent
   Config 增加字段：LLMServiceAllowedPlugins []string `json:"llm_service_allowed_plugins"`

行为：
- `@agent`：Broker 定位目标时，`@agent` 解析为当前安装记录（或库模式的 dev 组件）中 role 为 agent 的组件 ID，之后的访问规则、调用链、CALL_CYCLE 与激活检查按解析后的 ID 执行；没有 Agent 组件时与目标未激活相同。
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
  - provider 不存在为 INVALID_ARGUMENT；Agent 配置中没有 Provider 为 FAILED_PRECONDITION。
  - 模拟 Provider 出错与中途断开为 UNAVAILABLE，模拟服务只收到一次请求，safe_message 不含 API Key。
  - 调用方期限短于模拟 Provider 的延迟时，调用方得到 DEADLINE_EXCEEDED，模拟服务看到请求被取消。
  - Agent 组件未启用或未安装时调用返回 UNAVAILABLE。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：校验分支；三种协议各一次。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/... ./sdk/... ./internal/core/...；go build ./...；`go list -deps` 列出的 Core 可执行文件与 SDK 根包模块都在 docs/development.md 的允许清单中。

不做：流式输出与工具调用（契约不提供）；llm_service_allowed_plugins 的专用管理页（配置表单随 T-09 由 Schema 生成）。
```


## 第 14 波

### goal-60 · T-12.3 工具循环与 MCP

状态：等待前置（T-11.2、T-12.2）。语句中 goal-54、goal-55、goal-58、goal-59 提供的符号（sdk.ListTools、sdk.InvokeTool、sdk.ConfirmOperation、provider 的 Tool/ToolCall/ToolResult、agent 的一轮对话与 Config.MCPServers）只写名字；合并后由审查者核对，再改为待派发。本切片获授权引入已批准的 github.com/modelcontextprotocol/go-sdk（docs/development.md 只授权给 goal-60）。

```text
/goal 使用 $alonabot-execute 推进 goal-60（切片 T-12.3：工具循环与 MCP）。goal 目录 goal-60，分支 goal-60，worktree .worktrees/goal-60，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 组件每轮把 core.tool.list 返回的工具与已连接 MCP 服务器的工具交给模型；工具名按 Provider 的命名规则映射且一轮内稳定；Core 工具经 core.tool.invoke 调用，MCP 工具先经 core.tool.confirm（操作名 mcp/<server>/<tool>）再调用；拒绝、超时与工具错误作为工具结果交回模型；单轮工具步数达到上限时结束本轮并告知用户；MCP 服务器首次使用时连接、随插件停止关闭，stdio 子进程属于插件的进程树。

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
- 提交：agent/tools_test.go（//go:build unix）。用 goal-59 的测试装配、模拟 Provider（脚本化地返回工具调用）、goal-54/55 的工具插件与 mcptest 的两种服务器：
  - 模型连续调用一个 Core 工具与一个 MCP 工具：两者都执行，结果交回模型，最终回复显示。
  - Core 工具为 confirm 策略且用户拒绝：结果为错误文本交回模型，工具未执行；MCP 工具的确认被拒同样如此。
  - 工具出错：错误作为结果交回，不重试。
  - 模型持续请求工具超过 20 步：本轮结束并回复步数上限文案。
  - 停止插件后 stdio MCP 子进程已退出（按 PID 检查）。
  - 名称映射对含 / 与 . 的名字、冲突名字的结果稳定。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：名称映射函数；MCP 连接失败时对话继续；secret 值不出现在日志中；HTTP 服务器的 headers。

完成条件：gofmt -l 无输出；go mod tidy 后无差异；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：记忆（T-12.4）；技能（T-12.5）；MCP 的资源与提示词能力（只用 tools）；真实 Provider（T-10）。
```

### goal-61 · T-12.4 Agent 的记忆

状态：等待前置（T-08.5、T-11.2、T-12.2）。语句中 goal-19、goal-32、goal-54、goal-59 提供的符号（sdk.JSONService 与 Tool 选项、gormdb.Open、agent 的一轮对话）只写名字；合并后由审查者核对，再改为待派发。2026-10-28 按 D-60 改写：记忆并入 Agent，不再有 sdk/memory 与记忆插件。

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

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./agent/...；go build ./...；在 agent 运行 alonabot plugin check 通过。

不做：技能（T-12.5）；定时任务（T-12.6）；向量检索或外部记忆后端。
```


## 第 15 波

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

状态：等待前置（T-07.2、T-08.5、T-12.3）。语句中 goal-24、goal-32、goal-55、goal-59、goal-60 提供的符号（ScheduleRegistry.Every、gormdb.Open、sdk.ConfirmOperation、Agent 的一轮对话与工具循环、contracts/schedule）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-63（切片 T-12.6：定时 Agent 任务）。goal 目录 goal-63，分支 goal-63，worktree .worktrees/goal-63，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：Agent 组件提供 schedule_task、list_tasks、cancel_task 三个 Agent 内工具：只有 Core Admin 能创建与取消（之前经 core.tool.confirm），其他用户只能列出本聊天的任务；任务保存创建者、Target、提示词与触发规则（contracts/schedule 校验与计算）。一个每分钟一次的固定 schedule 检查到期任务，在其回调内以没有 Actor、只有 job scope 的方式运行一轮对话；发送需要 allowed_adapter_instances，confirm 策略的工具一律拒绝，不注入记忆，会话键与聊天会话分开，离线错过的触发不补跑。

依据：PLAN T-12 步骤 7；CONTRACTS §4.3、§13；TEST-MATRIX V-61。

范围（只新增或修改这些）：
- agent/：新增 tasks.go（工具与任务表）、scheduled.go（每分钟的 schedule 与后台一轮）；Agent 根插件注册 schedule（只改注册处）
- 上述包的测试文件
允许的新跨模块边：plugins/... -> contracts/schedule（plugins -> contracts 已允许）。

行为：
- 任务表（gormdb，versioned 迁移）：id、creator（instance、user）、target（instance、chat、thread）、prompt、trigger（interval 秒数或 cron 与时区）、next_run_at、created_at。
- schedule_task：输入 {"prompt", "interval_seconds"?, "cron"?, "timezone"?}，恰有 interval 或 cron 之一，经 contracts/schedule 校验（interval 至少 60s，命名常量；cron 为五字段）；调用者（工具循环的当前 Actor）必须含 ADMIN 角色，否则返回错误结果 "error: PERMISSION_DENIED"；先 ConfirmOperation("agent/schedule_task", 摘要)，允许后写入，next_run_at 为 Trigger.Next(创建时刻, 现在)。Target 为当前对话的默认 Target。
- list_tasks：返回本聊天（当前默认 Target 的 instance、chat、thread）的任务；任何人可用。
- cancel_task：输入 {"id"}；只有 ADMIN，先确认，再删除；只能取消本聊天的任务。
- 每分钟的 schedule（Every 1 分钟）：插件激活时把每个任务的 next_run_at 重算为未来时刻（不补跑）；每次回调选出 next_run_at ≤ 现在的任务，每个任务本次只运行一次，然后用 Trigger.Next 更新 next_run_at。运行一轮对话：
  - 没有 Actor，只有 job scope；会话键为 "task/<id>"，与聊天会话分开；不调用记忆服务。
  - 回复经 sdk.Send 显式 To 任务的 Target，需要 allowed_adapter_instances，没有许可时 PERMISSION_DENIED 并记 WARN（任务保留）。
  - 工具循环中 confirm 策略的 Core 工具与全部 MCP 工具一律拒绝（不发起确认，因为没有 Actor），作为错误结果交回模型。
  - 一次回调的运行受 JOB_TIMEOUT 约束；同一任务在上一次运行未结束时跳过。
- 插件离线期间到期的任务在重新激活后不补跑。

测试（按编码规范「测试提交规则」）：
- 提交：agent/tasks_test.go（//go:build unix）：用模拟 Provider（脚本化调用 schedule_task 等工具）与注入时钟的 Scheduler：
  - 非 Admin 调用 schedule_task 得到错误结果，任务表为空；Admin 经确认后创建 interval 与 cron 任务各一个。
  - 时钟推进到到期：任务运行一轮，回复发往任务 Target；没有发信许可时没有发送并记 WARN，授予许可后下一次运行发送。
  - 后台一轮中模型调用 confirm 策略的工具得到拒绝结果；没有注入记忆（请求中没有快照与档案）。
  - 停止插件、推进时钟越过多次触发、再启动：不补跑，只在下一次未来时刻运行。
  - 其他用户 list_tasks 只看到本聊天的任务；cancel_task 需要 Admin 与确认。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：interval 下限；一次回调中同一任务只运行一次；会话键隔离；JOB_TIMEOUT 内未结束时跳过。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./plugins/...；go build ./...。

不做：打包与 T-12 验收（T-12.7）；任务的管理页面。
```


## 第 16 波

### goal-64 · T-12.7 官方 Agent 组件打包与 T-12 验收

状态：等待前置（T-12.4、T-12.5、T-12.6、T-12.8）。语句中 goal-38、goal-58～goal-63、goal-83 提供的命令与符号（plugin pack、Agent 组件、mcptest、模拟 Provider）只写名字；合并后由审查者核对，再改为待派发。

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

验收矩阵（以 plugin pack 产生的安装包经 alonabot serve 与管理 API 安装、启用后运行；各为一个子测试）：
- V-57：三种协议的流式文本、工具调用、停止原因与用量；/ask 与自由对话开关；群聊非 to_me；/provider 与人格切换；/reset；超长回复分段；Provider 出错或中途断开。
- V-58：连续调用 Core 工具与 MCP 工具（stdio、streamable HTTP）；确认拒绝；工具出错；超过单轮步数上限；MCP 进程随 Agent 停止。
- V-59：记忆 add/patch/remove、超限、重复、不可见字符、群聊多用户、会话中写入与 /reset、memory 工具设为 confirm 与 deny。
- V-60：合法与不合法的 SKILL.md、load_skill 不存在的技能。
- V-61：非 Admin 创建任务、interval/cron 任务到期运行、无发信许可、confirm 工具、插件离线错过。
- V-76：另一个测试插件经安装包调用 llm.generate：未列入与列入 llm_service_allowed_plugins、Provider 不存在、Provider 出错、Agent 未启用。
- 证据：TEST-MATRIX §3.2 追加 T-12 一行，写明测试位置、命令、平台与结果；Provider 与 MCP 标为本地模拟；真实 Provider 与 Telegram 为「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t12_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：Agent 组件的 plugin check 与 pack（当前平台）；若维护者提供了 Provider 凭证，另附一次手动运行记录，不写入提交的测试。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：发行归档与预装（T-14.3、T-14.4）。
```
