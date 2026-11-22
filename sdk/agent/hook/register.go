package hook

import (
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/service"
)

// Builder 为修改阶段提供稳定 ID、顺序与类型化处理器。
type Builder[In, Out any] struct{ service *service.Builder[In, Out] }

// Priority 设置阶段内顺序，Agent 按 (priority, plugin_id, route_id) 排序。
func (b *Builder[In, Out]) Priority(p int32) *Builder[In, Out] { b.service.Priority(p); return b }

// Handle 冻结当前阶段处理器，返回错误终止当前修改操作。
func (b *Builder[In, Out]) Handle(fn func(*sdk.Context, In) (Out, error)) { b.service.Handle(fn) }

// ObserverBuilder 为观察阶段提供稳定 ID 和顺序。
type ObserverBuilder[In any] struct {
	service *service.Builder[In, Observed]
}

// Priority 设置观察者顺序。
func (b *ObserverBuilder[In]) Priority(p int32) *ObserverBuilder[In] { b.service.Priority(p); return b }

// Handle 登记观察者；Agent 记录错误后继续后续观察者。
func (b *ObserverBuilder[In]) Handle(fn func(*sdk.Context, In) error) {
	b.service.Handle(func(c *sdk.Context, in In) (Observed, error) { return Observed{}, fn(c, in) })
}

// BeforeRun 在插件或子插件内注册稳定 ID 的阶段服务。
func BeforeRun(owner sdk.Registrar, id string) *Builder[RunInput, RunPatch] {
	return &Builder[RunInput, RunPatch]{service: service.Serve(owner, service.JSON[RunInput, RunPatch](localTarget("agent.hook.beforerun", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// BeforeRequest 在插件或子插件内注册稳定 ID 的阶段服务。
func BeforeRequest(owner sdk.Registrar, id string) *Builder[RequestInput, RequestPatch] {
	return &Builder[RequestInput, RequestPatch]{service: service.Serve(owner, service.JSON[RequestInput, RequestPatch](localTarget("agent.hook.beforerequest", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// RewriteResponse 在插件或子插件内注册稳定 ID 的阶段服务。
func RewriteResponse(owner sdk.Registrar, id string) *Builder[ResponseInput, ResponsePatch] {
	return &Builder[ResponseInput, ResponsePatch]{service: service.Serve(owner, service.JSON[ResponseInput, ResponsePatch](localTarget("agent.hook.rewriteresponse", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// RewriteStream 在插件或子插件内注册稳定 ID 的阶段服务。
func RewriteStream(owner sdk.Registrar, id string) *Builder[StreamInput, StreamPatch] {
	return &Builder[StreamInput, StreamPatch]{service: service.Serve(owner, service.JSON[StreamInput, StreamPatch](localTarget("agent.hook.rewritestream", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// BeforeToolCall 在插件或子插件内注册稳定 ID 的阶段服务。
func BeforeToolCall(owner sdk.Registrar, id string) *Builder[ToolCallInput, ToolCallPatch] {
	return &Builder[ToolCallInput, ToolCallPatch]{service: service.Serve(owner, service.JSON[ToolCallInput, ToolCallPatch](localTarget("agent.hook.beforetoolcall", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// AfterToolCall 在插件或子插件内注册稳定 ID 的阶段服务。
func AfterToolCall(owner sdk.Registrar, id string) *Builder[ToolResultInput, ToolResultPatch] {
	return &Builder[ToolResultInput, ToolResultPatch]{service: service.Serve(owner, service.JSON[ToolResultInput, ToolResultPatch](localTarget("agent.hook.aftertoolcall", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// OnResponse 在插件或子插件内注册稳定 ID 的阶段服务。
func OnResponse(owner sdk.Registrar, id string) *ObserverBuilder[ResponseOutcome] {
	return &ObserverBuilder[ResponseOutcome]{service: service.Serve(owner, service.JSON[ResponseOutcome, Observed](localTarget("agent.hook.onresponse", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

// AfterRun 在插件或子插件内注册稳定 ID 的阶段服务。
func AfterRun(owner sdk.Registrar, id string) *ObserverBuilder[RunOutcome] {
	return &ObserverBuilder[RunOutcome]{service: service.Serve(owner, service.JSON[RunOutcome, Observed](localTarget("agent.hook.afterrun", id))).Access(service.AccessAllowlist, "@agent").Reentrant()}
}

func localTarget(stage, id string) string {
	if id == "" {
		return "/"
	}
	return "/" + stage + "." + id
}
