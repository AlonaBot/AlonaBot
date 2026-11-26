# Agent, service and lifecycle API

This page defines the shared public contract for SDK consumers and Agent/Core implementers. Matcher builders, OnSleep and the existing text generation call have SDK implementations. Agent runs, hook scheduling, injections, rewriting and persistent Goals require Agent services. Discovery, role authorization, reentry and streaming require the corresponding Core capabilities. The current Core advertises none of those four capabilities.

## Matcher builders and sleep

Plugin, SubPlugin and CommandGroup return `*CommandBuilder` from Command, `*TextBuilder` from Keyword/StartsWith/EndsWith/FullMatch, `*RegexBuilder` from Regex, and `*RawBuilder` from Raw. Every builder supports Permission, Priority, Rule, Continue and Handle. Each fluent method returns its own builder type.

Command alone supports Menu, Alias, Args, Optional, RawTail, Description, UsageKey and Silent. Text/Regex/Raw support ID. Text alone supports IgnoreCase; regular expressions use `(?i)`. Private registration state is shared. Handle ends registration and retains the existing signature checks.

Menu opts a command into publication. Publication still depends on the effective primary name being one token and existing permission, Profile, prefix and platform rules. Administrator renaming continues to affect eligibility.

`p.OnSleep(func(context.Context) error)` may be registered once. STOPPING with reason IDLE invokes OnSleep followed by OnStop, using the same original stopping deadline. Errors and panics are logged with their full cause and remaining cleanup proceeds. A callback that ignores cancellation is abandoned when the shared deadline expires. Other stop reasons invoke OnStop. Waking initializes through OnStart and OnReady.

## Generate and Run

```go
llm.Generate(ctx, req llm.Request) (llm.Response, error)
llm.GenerateStream(ctx, req llm.Request, onDelta func(llm.Delta) error) (llm.Response, error)
agent.Run(ctx, req agent.RunRequest) (agent.RunResult, error)
agent.RunStream(ctx, req agent.RunRequest, onEvent func(agent.RunEvent) error) (agent.RunResult, error)
```

Generate targets `@agent/llm.generate`; GenerateStream targets `@agent/llm.generate.stream`. Run targets the independent `@agent/agent.run` service; RunStream targets `@agent/agent.run.stream`. Complete and stream handlers have distinct registrations. One Generate makes one model request. A Run can make multiple requests and execute tools. Calling the SDK service returns execution results; reading and persisting chat history belongs to the Agent conversation entry point. Provider replay state stays in the Agent and must be returned to the Provider unchanged.

| Type/field | Contract |
|---|---|
| Message.Role | string, user or assistant |
| Message.Text | string, complete text, preserved as supplied |
| Request.Messages | []Message, at least one, in supplied order |
| Request.System | string, base system prompt |
| Request.Model | string; empty selects the configured chat model |
| Request.MaxOutputTokens | int; zero uses the model configuration; negative is InvalidArgument |
| Response.Text | string, complete rewritten response |
| Response.Model | string, actual model configuration name |
| Response.StopReason | string, normalized stop reason, unknown values preserved |
| Response.InputTokens/OutputTokens | *int; nil means unavailable, a pointer preserves an actual zero |
| Delta.RequestID | string, assigned by Agent per model request |
| Delta.Sequence | uint64, delivered sequence starting at zero within that request |
| Delta.Text | string, delta after streaming hooks |
| RunRequest.Input | llm.Request, initial input and model options |
| RunRequest.Tools | *ToolSelection; nil selects all tools accessible to this call |
| ToolSelection.Names | []string, only these full tool identifiers; an empty list disables tools |
| RunResult.RunID | string, execution identifier |
| RunResult.Text | string, concatenation of all delivered model text in this Run |
| RunResult.Steps | []RunStep, actual execution order |
| RunResult.InputTokens/OutputTokens | *int, summed only when every request supplies that usage |
| RunStep.Kind | StepKind, model or tool |
| RunStep.RequestID | string, originating model request |
| RunStep.Model/Tool | *llm.Response / *ToolResult; set exactly the branch selected by Kind |
| ToolCall.Name | string, full tool identifier |
| ToolCall.CallID | string, Provider tool call identifier |
| ToolCall.Arguments | json.RawMessage, complete JSON object after hooks |
| ToolResult.Call | ToolCall, actual executed call and arguments |
| ToolResult.Raw | json.RawMessage, original result |
| ToolResult.Text | string, result supplied to the model |
| ToolResult.Failure | *Failure, actual execution error if present |
| Failure.Code/Message | strings, error code and full cause |
| RunEvent.RunID/RequestID | strings, Run and model request identifiers |
| RunEvent.Sequence | uint64, starting at zero across the whole Run |
| RunEvent.Kind | EventKind: text_delta, tool_started or tool_finished |
| RunEvent.Text/Call/Result | string / *ToolCall / *ToolResult; use the corresponding Kind branch |

Successful streaming Response.Text equals the concatenation of successfully delivered deltas. Interrupted calls return confirmed partial results together with the error. Returning an error from a callback cancels the generation/Run and returns that error with the partial result. SDK encoding, handlers and calls preserve payload plus error.

Operations submitted through `tool.Confirm` default to `ALLOW` and return without sending a confirmation message, including calls with no Actor. Administrator overrides from `/api/v1/tools/policies` and profile policies remain effective; `CONFIRM` requires a trusted Actor and target. Tool services keep their declared default policies.

## Eight Agent hook stages

The `sdk/agent/hook` package accepts `sdk.Registrar`, so registration works on Plugin and SubPlugin. Each registration requires a nonempty stable ID within the plugin and supports Priority(int32) and typed Handle. Internally these are ordinary services with reentry enabled and an allowlist for the current `@agent` role. Agent executes them in `(priority, plugin_id, route_id)` order, smaller priorities first.

```go
hook.BeforeRequest(p, "weather").Priority(10).
    Handle(func(c *sdk.Context, in hook.RequestInput) (hook.RequestPatch, error) {
        return hook.RequestPatch{Injections: []hook.InjectionChange{
            {Op: hook.Upsert, Key: "temperature", Scope: hook.Run, Text: "temperature: 28 C"},
        }}, nil
    })
```

| Registration | Input | Output |
|---|---|---|
| BeforeRun | RunInput | RunPatch, change initial input and injections |
| BeforeRequest | RequestInput | RequestPatch, change the current model request and injections |
| RewriteResponse | ResponseInput | ResponsePatch, rewrite complete mode text |
| RewriteStream | StreamInput | StreamPatch, rewrite/buffer/flush streaming text |
| BeforeToolCall | ToolCallInput | ToolCallPatch, change arguments |
| AfterToolCall | ToolResultInput | ToolResultPatch, change model-visible result text |
| OnResponse | ResponseOutcome | error, observe rewritten response, actual usage and failure |
| AfterRun | RunOutcome | error, observe Run result and failure |

Every input contains Meta. Meta.RunID is empty for independent Generate, RequestID is set once a model request starts, Iteration starts at 1 inside a Run, and GoalID is set for Goal execution. Trusted caller identity, permissions, cancellation and deadlines come from `*sdk.Context`.

| Type/field | Contract |
|---|---|
| RunInput.Request | agent.RunRequest, initial request |
| RunInput.Injections / RequestInput.Injections | []Injection, effective injections, separately from the base request |
| RunPatch.Request | *agent.RunRequest, nil keeps the original |
| RunPatch.Injections / RequestPatch.Injections | []InjectionChange, explicit updates/deletions |
| RequestInput.Request | RequestView, complete base model request |
| RequestView.System | string, base prompt before injection assembly |
| RequestView.Messages | []hook.Message, complete model context |
| RequestView.Model/MaxOutputTokens | string/int, same model and token semantics as llm.Request |
| RequestView.Tools | []Tool, available tool catalog |
| hook.Message.Role/Text | strings; role is user, assistant or tool; complete text |
| hook.Message.ToolCalls | []agent.ToolCall, assistant tool calls |
| hook.Message.ToolResult | *agent.ToolResult, corresponding tool result |
| hook.Message.Replay | json.RawMessage, read-only Provider state |
| Tool.Name/Description/InputSchema | full identifier, description, raw JSON Schema |
| RequestPatch.System/Messages/Model/MaxOutputTokens/Tools | pointers to their RequestView field types; nil keeps the value, a pointer to an empty value clears it |
| ResponseInput.Response | llm.Response, actual complete-mode response |
| ResponsePatch.Text | *string, nil keeps text, pointer to empty clears it |
| ToolCallInput.Call | agent.ToolCall, actual proposed call |
| ToolCallPatch.Arguments | *json.RawMessage, nil keeps arguments; replacement must be a JSON object |
| ToolResultInput.Result | agent.ToolResult, real execution result |
| ToolResultPatch.Text | *string, only model-visible text is writable |
| ResponseOutcome.Response/Failure | llm.Response / *agent.Failure, confirmed response and failure |
| RunOutcome.Result/Failure | agent.RunResult / *agent.Failure, confirmed Run and failure |

Actual model identity, token counts and execution results stay truthful. Agent validates the final patched arguments against the Schema and existing tool policy. Modification hook errors terminate the operation. Observer errors are logged in full and remaining observers continue. The observer service returns `Observed {}` while the SDK Handle accepts a function returning only error. Replay state remains unchanged across patches.

### Injection ownership and lifetime

InjectionChange contains Op (Upsert or Delete), Key string, Scope (Request or Run), and Text string. Agent derives ownership from the trusted hook service. Injection exposes PluginID, RouteID, Key, Scope and Text. The storage key is `(plugin_id, scope, key)`.

Upsert replaces text for the same key; Delete explicitly removes it. A plugin's Request injection overrides its same-key Run injection for that model request, then the Run injection becomes effective again. Request injections returned by BeforeRun apply to the first request. Run injections last until the Run ends; independent Generate accepts Request injections. Injections are ordered by hook order and then key within one route, appended after the base system prompt. Hooks receive the base request and the effective injection list separately, and Agent rebuilds the prompt for each model request. Injection text stays in runtime state; persisting it into long-term history is an explicit business operation.

For example, plugin A Upserts Run key `location` with `location: Singapore` and plugin B Upserts Run key `temperature` with `temperature: 28 C`. Both survive across requests. A can Upsert `location` again without affecting B. B deletes its temperature with `{Op: Delete, Key: "temperature", Scope: Run}`. A Request-scoped location overrides A's Run location only for that request.

### Streaming rewrite state and flush

StreamInput contains Meta, Phase (Start/Delta/End/Abort), Text string (Delta input), State json.RawMessage (the preceding private state), and Failure *agent.Failure (Abort cause). StreamPatch contains Texts []string (zero, one or several outputs) and State json.RawMessage (nil clears state). Agent isolates state by hook service plus RequestID.

Start initializes state. Delta may buffer incomplete matches: receiving `sec` can emit no text and save that prefix in State; the next `ret` can emit `[redacted]`. End flushes the remaining buffer and clears State. An upstream End output must pass through downstream Delta processing before downstream End is invoked. Abort cleans up state; if the link has failed, Agent releases its local state. Rewriting happens before delivery, so delivered text is final. Complete mode uses RewriteResponse; streaming completion uses OnResponse to observe the rewritten aggregate. Subsequent model context, Run results and chat history use rewritten content.

## Generic service protocol

```go
service.Discover(ctx, filter service.Filter) ([]service.Descriptor, error)
service.StreamJSON[In, Delta, Out]("plugin_id/service_name")
service.ServeStream(owner, declaration).Handle(
    func(c *sdk.Context, in In, emit func(Delta) error) (Out, error) { /* ... */ })
declaration.Call(ctx, in, onDelta) (Out, error)
```

Service declarations on the provider may use `/service_name` to bind the authenticated plugin identity. Builder.Reentrant() explicitly allows callback reentry; Builder.Priority(int32) controls order. Access and Permission retain their meanings. Ordinary services retain cycle rejection. Reentrant calls still check the direct caller, permission, depth, deadline and valid Session. The `@agent` allowlist alias resolves to the currently installed Agent identity.

Service calls inherit the parent event's `actor`, `default_target` and `chat_type` through `InvocationContext` (`chat_type = 8`). Use `c.User()`, `c.Session()` and `c.ChatType()` in a Service handler; `c.Event()` is nil. Without a real inbound event, `c.ChatType()` returns `message.ChatUnspecified`.

Filter has InputSchemaID string, InputSchemaVersion uint32, OutputSchemaID string and OutputSchemaVersion uint32; empty IDs and zero versions independently mean unrestricted. Discovery returns only effective services accessible to the current caller, including valid snapshots that can be woken on demand. Descriptor contains Target, PluginID, RouteID, Priority, Encoding, InputSchemaID/Version, OutputSchemaID/Version, DeltaSchemaID/Version and Reentrant. Empty delta Schema means an ordinary service. Agent recognizes hook stages through their shared input/output Schema identities.

JSON Schema identities are the full package import path plus `.` plus the named type, at version 1. The generator recognizes JSON, StreamJSON and public Agent wrappers. It supports json.RawMessage as a raw JSON value and time.Time as an RFC3339 date-time string; other custom codec restrictions remain in effect. Tool Arguments receive the additional object requirement at Agent's boundary.

The four capability names are `service_discover`, `service_role_access`, `service_reentrant` and `service_stream`, each at version 1. They use the existing HELLO/HELLO_ACK supported/required/intersection negotiation. SDK offers its supported side; Core must advertise a capability only after implementing its runtime. Current Core offers none. Calls lacking discovery or streaming capability return Unsupported. Plugins declaring role authorization, reentry or streaming fail activation explicitly with Unsupported when the capability is absent. Ordinary Agent/Goal calls preserve lookup errors: missing services return NotFound; an unavailable Agent role preserves the existing Unavailable error.

The transport uses existing bidirectional REQUEST/RESPONSE frames. `core.service.discover` uses ServiceDiscoverRequest/Response. `core.service.stream.call` uses ServiceStreamCallRequest; the initial call stays in flight and the final RESPONSE carries the declared business result and optional error. Providers send ServiceStreamDeltaRequest to `core.service.stream.emit` under their trusted parent. Core binds the stream to that authenticated Session and parent, then forwards it to `sdk.service.stream.delta`. Receiver processing completes before its ACK, so emit supplies backpressure. Sequence starts at zero and increments per successful acknowledgement. Cancellation, deadlines, disconnects, frame limits and in-flight budgets follow the original call chain.

Protocol sources: [service.proto](../contracts/proto/alonabot/v1/service.proto), [route.proto](../contracts/proto/alonabot/v1/route.proto), [invocation.proto](../contracts/proto/alonabot/v1/invocation.proto). ServiceSpec adds reentrant=7, delta_schema_id=8 and delta_schema_version=9. Ordering uses the existing RouteDescriptor.priority. ServiceInvocation.stream_id=7 binds the trusted in-flight call. StreamCall fields target/stream_id/encoding/schema_id/schema_version/payload are numbered 1–6; StreamDelta fields stream_id/sequence/encoding/schema_id/schema_version/payload are numbered 1–6. Core must validate Session, parent, identity, access, deadline and sequence. The stream ID correlates calls; Session and parent supply authentication.

## Persistent Goal management

The `sdk/agent/goal` package exposes Create(ctx, CreateRequest), Get(ctx,id), Pause(ctx,id), Resume(ctx,id), Cancel(ctx,id), each returning (Goal,error), and List(ctx) returning ([]Goal,error). Targets are `@agent/goal.create`, `.get`, `.list`, `.pause`, `.resume` and `.cancel`. IDRequest contains ID string; ListRequest is an empty object and ListResponse contains Goals []Goal. Missing IDs return NotFound.

| Type/field | Contract |
|---|---|
| CreateRequest.Condition | required string, verifiable completion condition |
| CreateRequest.Run | agent.RunRequest, initial execution input and tools |
| CreateRequest.EvaluatorModel | string, empty selects the model chosen for this Run; assessment is a separate model request |
| CreateRequest.Checker | optional full target of a deterministic checking service |
| CreateRequest.RecoveryPolicy | Manual by default (including empty); Automatic is reserved and must be rejected explicitly by Agents that do not implement it |
| Goal.ID/CreatorPluginID | strings, ID and authenticated owner |
| Goal.Condition/RecoveryPolicy | completion condition and effective recovery policy |
| Goal.Status | Pending, Running, Paused, Completed, Failed or Cancelled; JSON values are lowercase |
| Goal.RunCount | uint64, number of executed Runs |
| Goal.LatestEvaluation | *Evaluation, latest saved assessment if available |
| Goal.Reason | string, full pause or failure reason |
| Goal.CreatedAt/UpdatedAt | time.Time, JSON RFC3339 |
| Evaluation.Verdict | NotMet, Met or Impossible; JSON not_met/met/impossible |
| Evaluation.Reason | string, evidence and explanation; NotMet guides the next Run |
| CheckerInput.GoalID/Condition | strings, identity and condition |
| CheckerInput.Runs | []agent.RunResult, saved execution evidence in order |

`goal.Checker(owner,id)` accepts sdk.Registrar and returns a typed service builder for CheckerInput → Evaluation. The service name is `goal.checker.<id>`, route `svc.goal.checker.<id>`, with current Agent access and explicit reentry. The eight hook service names are `agent.hook.<lowercase-stage>.<id>` and routes add `svc.`.

After every Run, the default evaluator independently assesses the condition and recorded evidence. When Checker is provided its Evaluation decides acceptance. Met completes the Goal, Impossible fails it, and NotMet starts another Run with the reason as guidance. Operational errors during running or assessment pause the Goal and save their full reason. Pause cooperatively cancels the Run and saves confirmed steps; Cancel enters the terminal Cancelled state. Management authorizes by real creator plugin; Agent's interactive entry uses trusted user permissions.

Goals, execution evidence, progress and assessments belong to the Agent database. Scheduling reuses existing dynamic tasks. Manual recovery pauses unfinished work after restart. Automatic is reserved for future Agent implementation and restores only Goals that were running before restart. Explicitly paused Goals require Resume. Recovery schedules a new Run from saved evidence; later steps verify side-effect calls with unknown outcomes. Each Run reacquires location, temperature and other injections.

```go
g, err := goal.Create(c, goal.CreateRequest{
    Condition: "The requested report is saved and the checker accepts it",
    Run: agent.RunRequest{Input: llm.Request{
        Messages: []llm.Message{{Role: "user", Text: "Prepare the report"}},
    }},
    RecoveryPolicy: goal.Manual,
})
if err != nil { return err }
g, err = goal.Get(c, g.ID) // After restart, unfinished Manual Goals are Paused.
if err != nil { return err }
_, err = goal.Resume(c, g.ID) // Schedule a new Run from saved evidence.
return err
```

If cancellation prevents the terminal response from arriving, GenerateStream retains successfully delivered text. RunStream retains the confirmed RunID, model text and ToolFinished records; unavailable model identity and usage stay empty. Successful terminal text must match delivered text.
