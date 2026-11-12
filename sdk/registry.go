package sdk

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// UserID 携带命令 user 参数所指向的 external_user_id。
type UserID string

// Command 记录一条已收集的命令注册。
type Command struct {
	Path      []string       // 完整命令路径（含分组前缀），大小写敏感
	Aliases   [][]string     // 完整别名路径，不附加分组前缀
	Priority  int32          // 数值越小越先，缺省为 0
	Handler   any            // 已通过签名检查的处理函数
	mode      pb.HandlerMode // 由签名推出的 Handler 形态
	structure reflect.Type   // 结构体参数的值类型；位置参数形式为 nil。

	permission Permission
	routeID    string          // 文本匹配与 Raw 的显式路由 ID
	spec       *pb.CommandSpec // 已通过注册边界校验的命令描述
}

// CommandRegistry 收集命令；实例方法以方法值显式传入，不做自动 DI。
// 子组共享根注册器的构建链和冲突索引；Collect 在全部声明结束后解析继承权限。
type CommandRegistry struct {
	parent     *CommandRegistry
	groups     map[string]*CommandRegistry
	permission *Permission
	builders   []*CommandBuilder
	prefix     []string          // 本分组的公共路径前缀
	commands   []Command         // 已收集的命令，按注册顺序排列
	errs       []error           // 收集过程中发现的错误，由 Collect 一并返回
	seen       map[string]string // 路径或别名对应的 route_id；惰性创建
	routeIDs   map[string]bool   // 命令路径生成的 route_id，跨分组检查重复
}

type routeOptions struct {
	aliases    [][]string
	priority   int32
	names      []string
	argsSet    bool // 区分未调用 Args 与显式 Args()，结构体参数禁止后者。
	optional   []optionalArgument
	raw        bool
	usageMode  pb.UsageMode
	usageKey   string
	ignoreCase bool
}

type optionalArgument struct {
	name     string
	defaults []string
}

// Group 返回共享注册状态的子组；同名调用返回同一组。
func (r *CommandRegistry) Group(name string) *CommandRegistry {
	if r.groups == nil {
		r.groups = make(map[string]*CommandRegistry)
	}
	if r.groups[name] == nil {
		r.groups[name] = &CommandRegistry{parent: r, prefix: append(slices.Clone(r.prefix), name)}
	}
	return r.groups[name]
}

// Permission 设置本组及未显式覆盖的后代路由权限。
func (r *CommandRegistry) Permission(p Permission) *CommandRegistry { r.permission = &p; return r }

func (r *CommandRegistry) owner() *CommandRegistry {
	for r.parent != nil {
		r = r.parent
	}
	return r
}

// Command 开始构建相对于本组路径的命令。
func (r *CommandRegistry) Command(path ...string) *CommandBuilder {
	return r.start(Command{Path: append(slices.Clone(r.prefix), path...), spec: &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH}}, len(path) == 0)
}

// Regex 开始构建 Go regexp 文本匹配。
func (r *CommandRegistry) Regex(routeID, expr string) *CommandBuilder {
	return r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_REGEX, Regex: expr})
}

// Keyword 开始构建包含文本匹配。
func (r *CommandRegistry) Keyword(routeID string, patterns ...string) *CommandBuilder {
	return r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_KEYWORD, Patterns: slices.Clone(patterns)})
}

// StartsWith 开始构建开头文本匹配。
func (r *CommandRegistry) StartsWith(routeID string, patterns ...string) *CommandBuilder {
	return r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_STARTSWITH, Patterns: slices.Clone(patterns)})
}

// EndsWith 开始构建结尾文本匹配。
func (r *CommandRegistry) EndsWith(routeID string, patterns ...string) *CommandBuilder {
	return r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_ENDSWITH, Patterns: slices.Clone(patterns)})
}

// FullMatch 开始构建完整文本匹配。
func (r *CommandRegistry) FullMatch(routeID string, patterns ...string) *CommandBuilder {
	return r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_FULLMATCH, Patterns: slices.Clone(patterns)})
}

// Raw 开始构建最后执行的原始消息路由。
func (r *CommandRegistry) Raw(routeID string) *CommandBuilder {
	return r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_RAW})
}

func (r *CommandRegistry) text(routeID string, spec *pb.CommandSpec) *CommandBuilder {
	return r.start(Command{routeID: routeID, spec: spec}, routeID == "" || strings.HasPrefix(routeID, commandRoutePrefix) || strings.HasPrefix(routeID, buttonRoutePrefix))
}

func (r *CommandRegistry) start(c Command, invalid bool) *CommandBuilder {
	b := &CommandBuilder{registry: r, command: c, invalid: invalid, options: routeOptions{usageMode: pb.UsageMode_USAGE_MODE_DEFAULT}}
	r.owner().builders = append(r.owner().builders, b)
	return b
}

// CommandBuilder 收集一条路由；Handle 结束构建，Collect 统一验证。
type CommandBuilder struct {
	registry    *CommandRegistry
	command     Command
	options     routeOptions
	permission  *Permission
	handles     int
	afterHandle bool
	invalid     bool
}

func (b *CommandBuilder) touch() {
	if b.handles != 0 {
		b.afterHandle = true
	}
}

// Alias 追加完整别名路径。
func (b *CommandBuilder) Alias(path ...string) *CommandBuilder {
	b.touch()
	b.options.aliases = append(b.options.aliases, slices.Clone(path))
	return b
}

// Args 设置参数名。
func (b *CommandBuilder) Args(names ...string) *CommandBuilder {
	b.touch()
	b.options.names = slices.Clone(names)
	b.options.argsSet = true
	return b
}

// Optional 追加可选参数及默认值。
func (b *CommandBuilder) Optional(name string, def ...string) *CommandBuilder {
	b.touch()
	b.options.optional = append(b.options.optional, optionalArgument{name: name, defaults: slices.Clone(def)})
	return b
}

// RawTail 保留末位参数原文。
func (b *CommandBuilder) RawTail() *CommandBuilder { b.touch(); b.options.raw = true; return b }

// UsageKey 设置 usage 文案键。
func (b *CommandBuilder) UsageKey(key string) *CommandBuilder {
	b.touch()
	b.options.usageMode = pb.UsageMode_USAGE_MODE_CUSTOM
	b.options.usageKey = key
	return b
}

// Silent 关闭参数错误回复。
func (b *CommandBuilder) Silent() *CommandBuilder {
	b.touch()
	b.options.usageMode = pb.UsageMode_USAGE_MODE_SILENT
	b.options.usageKey = ""
	return b
}

// IgnoreCase 启用 Unicode 大小写折叠。
func (b *CommandBuilder) IgnoreCase() *CommandBuilder {
	b.touch()
	b.options.ignoreCase = true
	return b
}

// Priority 设置路由优先级。
func (b *CommandBuilder) Priority(p int32) *CommandBuilder {
	b.touch()
	b.options.priority = p
	return b
}

// Permission 覆盖本路由权限。
func (b *CommandBuilder) Permission(p Permission) *CommandBuilder {
	b.touch()
	b.permission = &p
	return b
}

// Handle 结束路由构建，签名规则由 Collect 检查。
func (b *CommandBuilder) Handle(handler any) { b.handles++; b.command.Handler = handler }

func (r *CommandRegistry) collect() {
	for _, b := range r.builders {
		c := b.command
		id := commandRouteID(c)
		if b.invalid || b.handles != 1 || b.afterHandle {
			r.errs = append(r.errs, invalidCommand("route %q (path %q) has invalid path/route_id or Handle sequence", id, strings.Join(c.Path, " ")))
			continue
		}
		p := b.permission
		for g := b.registry; p == nil && g != nil; g = g.parent {
			p = g.permission
		}
		if p == nil {
			p = &User
		}
		if err := checkPermission(*p); err != nil {
			r.errs = append(r.errs, invalidCommand("route %q: %v", id, err))
			continue
		}
		c.permission = *p
		r.register(c, b.options)
	}
}

func (r *CommandRegistry) register(c Command, options routeOptions) {
	// 1. 从真实函数签名取得参数类型，选项仅描述名称与缺失时的行为。
	mode, structure, params, err := checkHandler(c.Handler)
	if err != nil {
		r.errs = append(r.errs, invalidCommand("command %q: %v", commandRouteID(c), err))
		return
	}
	c.mode, c.structure = mode, structure
	c.Aliases, c.Priority = options.aliases, options.priority
	spec := c.spec
	if structure != nil && (options.argsSet || len(options.optional) != 0 || options.raw) {
		r.errs = append(r.errs, invalidCommand("command %q uses struct parameters with Args, Optional or RawTail", commandRouteID(c)))
		return
	}
	spec.Path, spec.HandlerMode = c.Path, mode
	spec.UsageMode, spec.UsageKey, spec.IgnoreCase = options.usageMode, options.usageKey, options.ignoreCase
	spec.Params = params
	if len(options.names) > len(params) {
		r.errs = append(r.errs, invalidCommand("command %q has more argument names than parameters", commandRouteID(c)))
		return
	}
	for i, name := range options.names {
		spec.Params[i].Name = name
	}
	for _, optional := range options.optional {
		if len(optional.defaults) > 1 {
			r.errs = append(r.errs, invalidCommand("parameter %q has more than one default", optional.name))
			return
		}
		found := false
		for _, param := range spec.Params {
			if param.Name != optional.name {
				continue
			}
			found = true
			param.Required = false
			param.DefaultValue = nil
			if len(optional.defaults) == 1 {
				value := optional.defaults[0]
				param.DefaultValue = &value
			}
		}
		if !found {
			r.errs = append(r.errs, invalidCommand("command %q has no parameter %q", commandRouteID(c), optional.name))
			return
		}
	}
	if options.raw {
		if len(spec.Params) == 0 {
			r.errs = append(r.errs, invalidCommand("command %q has no parameter for RawTail", commandRouteID(c)))
			return
		}
		spec.Params[len(spec.Params)-1].Raw = true
	}
	for _, path := range c.Aliases {
		spec.Aliases = append(spec.Aliases, &pb.CommandAlias{Path: path})
	}
	if spec.GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH && len(c.Aliases) != 0 {
		r.errs = append(r.errs, invalidCommand("text or raw route %q cannot declare aliases", commandRouteID(c)))
		return
	}
	// 2. SDK 收集边界与 Core REGISTER 复用同一份描述校验。
	if err := validate.Register(&pb.Register{CandidateConfigJson: "{}", Routes: []*pb.RouteDescriptor{{
		RouteId: commandRouteID(c), Kind: pb.RouteKind_ROUTE_KIND_COMMAND, Spec: &pb.RouteDescriptor_Command{Command: spec},
	}}}); err != nil {
		r.errs = append(r.errs, err)
		return
	}
	r.add(c)
}

func invalidCommand(format string, args ...any) *Error {
	return &Error{Code: errcode.InvalidArgument, Message: fmt.Sprintf(format, args...)}
}

// add 对已校验的命令检查跨注册项冲突；问题记入 errs。
func (r *CommandRegistry) add(c Command) {
	// 1. 全部路径与别名检查跨注册项冲突，失败时不登记部分路径。
	var paths [][]string
	if c.spec.GetMatch() == pb.CommandMatch_COMMAND_MATCH_PATH {
		paths = append([][]string{c.Path}, c.Aliases...)
	}
	pending := make(map[string]bool, len(paths))
	id := commandRouteID(c)
	for _, path := range paths {
		key := strings.Join(path, " ")
		if previous, exists := r.seen[key]; exists {
			r.errs = append(r.errs, invalidCommand("command path %q conflicts between routes %q and %q", key, previous, id))
			return
		}
		pending[key] = true
	}
	if r.routeIDs[id] {
		r.errs = append(r.errs, invalidCommand("route %q registered twice", id))
		return
	}
	// 2. 全部通过后统一登记，所有分组共用根注册器的冲突索引。
	if r.seen == nil {
		r.seen = make(map[string]string)
	}
	if r.routeIDs == nil {
		r.routeIDs = make(map[string]bool)
	}
	for key := range pending {
		r.seen[key] = id
	}
	r.routeIDs[id] = true
	r.commands = append(r.commands, c)
}

// Handler 签名检查使用的公共入口与返回值类型。
var (
	contextType = reflect.TypeOf((*Context)(nil))
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contentType = reflect.TypeFor[message.Content]()
)

// HTTPRoute 记录一条插件 HTTP 路由；path 相对于 /plugins/{id}，端口由 SDK 托管（D-25）。
type HTTPRoute struct {
	Method string // 大写的 HTTP 方法
	Path   string // 插件相对路径，以 / 开头且为规范形式
	// Handler 处理请求。
	Handler http.Handler
	// Public 为 true 时无需 Core 会话即可访问。
	Public   bool
	AuthMode HTTPAuth // 认证方式；公开路由必须是回调验证方式之一
}

// HTTPOption 调整 HTTP 路由声明。
type HTTPOption func(*HTTPRoute)

// Public 声明公开回调路由；插件必须自行验证 state、签名或 nonce。
func Public(mode HTTPAuth) HTTPOption {
	return func(r *HTTPRoute) { r.Public, r.AuthMode = true, mode }
}

// HTTPRegistry 收集 HTTP 路由；默认需要 Core 授权会话。
type HTTPRegistry struct {
	routes []HTTPRoute
	errs   []error
}

// Handle 注册 method+path 的 Handler；同插件内重复的 method+path 报错。
func (r *HTTPRegistry) Handle(method, path string, h http.Handler, opts ...HTTPOption) {
	route := HTTPRoute{Method: strings.ToUpper(method), Path: path, Handler: h, AuthMode: AuthCoreSession}
	for _, o := range opts {
		o(&route)
	}
	switch {
	case h == nil:
		r.errs = append(r.errs, fmt.Errorf("handler for HTTP %s %s is nil", method, path))
	case !validHTTPMethod(method) || !validHTTPMethod(route.Method):
		r.errs = append(r.errs, fmt.Errorf("HTTP method must be standard and uppercase: input %q, registered %q", method, route.Method))
	case !canonicalHTTPPath(route.Path):
		r.errs = append(r.errs, fmt.Errorf("HTTP path %q must start with / and be canonical", route.Path))
	case route.Public && route.AuthMode != AuthCallbackState && route.AuthMode != AuthCallbackSignature && route.AuthMode != AuthCallbackNonce:
		r.errs = append(r.errs, fmt.Errorf("public HTTP route %s %s must declare state, signature or nonce verification", method, path))
	case !route.Public && route.AuthMode != AuthCoreSession:
		r.errs = append(r.errs, fmt.Errorf("private HTTP route %s %s must use Core session authentication", method, path))
	default:
		for _, e := range r.routes {
			if e.Method == route.Method && e.Path == route.Path {
				r.errs = append(r.errs, fmt.Errorf("HTTP %s %s registered twice", route.Method, path))
				return
			}
		}
		r.routes = append(r.routes, route)
	}
}

// Schedule 记录一个定时任务声明；interval 与 cron 互斥，触发由 Core 负责。
type Schedule struct {
	JobID    string               // 插件内稳定的任务 ID
	Interval time.Duration        // 固定间隔；为零表示使用 Cron
	Cron     string               // 五字段 cron 表达式；为空表示使用 Interval
	Timezone string               // cron 使用的 IANA 时区，空串表示 UTC
	Handler  func(*Context) error // 每次触发时调用的函数
}

// ScheduleRegistry 收集定时任务。
type ScheduleRegistry struct {
	jobs []Schedule
	errs []error
}

// Every 声明固定间隔任务（至少 1s）。
func (r *ScheduleRegistry) Every(jobID string, d time.Duration, fn func(*Context) error) {
	if d < time.Second {
		r.errs = append(r.errs, fmt.Errorf("interval %[2]s of job %[1]s is less than 1s", jobID, d))
		return
	}
	r.add(Schedule{JobID: jobID, Interval: d, Handler: fn})
}

// Cron 声明五字段 cron 任务；timezone 为 IANA 名称，空串表示 UTC。表达式在 REGISTER 时由 Core 校验。
func (r *ScheduleRegistry) Cron(jobID, expr, timezone string, fn func(*Context) error) {
	if len(strings.Fields(expr)) != 5 {
		r.errs = append(r.errs, fmt.Errorf("cron expression %[2]q of job %[1]s does not have five fields", jobID, expr))
		return
	}
	r.add(Schedule{JobID: jobID, Cron: expr, Timezone: timezone, Handler: fn})
}

// add 检查 job_id、Handler 并去重后登记任务；问题记入 errs。
func (r *ScheduleRegistry) add(s Schedule) {
	switch {
	case s.JobID == "":
		r.errs = append(r.errs, fmt.Errorf("scheduled job is missing job_id"))
		return
	case s.Handler == nil:
		r.errs = append(r.errs, fmt.Errorf("handler for job %s is nil", s.JobID))
		return
	}
	for _, j := range r.jobs {
		if j.JobID == s.JobID {
			r.errs = append(r.errs, fmt.Errorf("job %s registered twice", s.JobID))
			return
		}
	}
	r.jobs = append(r.jobs, s)
}
