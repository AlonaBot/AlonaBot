package sdk

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Command 记录一条已收集的命令注册。
type Command struct {
	Path     []string       // 完整命令路径（含分组前缀），大小写敏感
	Aliases  [][]string     // 完整别名路径，不附加分组前缀
	Priority int32          // 数值越小越先，缺省为 0
	Handler  any            // 已通过签名检查的处理函数
	Mode     pb.HandlerMode // 由签名推出的 Handler 形态
	Params   []pb.ParamType // Context 之后各参数的类型

	routeID string          // 文本匹配与 Raw 的显式路由 ID
	spec    *pb.CommandSpec // 已通过注册边界校验的命令描述
}

// CommandRegistry 收集命令；实例方法以方法值显式传入，不做自动 DI。
type CommandRegistry struct {
	prefix   []string          // 本分组的公共路径前缀
	commands []Command         // 已收集的命令，按注册顺序排列
	errs     []error           // 收集过程中发现的错误，由 Collect 一并返回
	seen     map[string]string // 路径或别名对应的 route_id；惰性创建
	routeIDs map[string]bool   // 命令路径生成的 route_id，跨分组检查重复
}

type routeOptions struct {
	aliases    [][]string
	priority   int32
	names      []string
	optional   []optionalArgument
	raw        bool
	usageMode  pb.UsageMode
	usage      string
	ignoreCase bool
}

type optionalArgument struct {
	name     string
	defaults []string
}

// RouteOption 调整命令或事件路由的静态声明。
type RouteOption func(*routeOptions)

// Alias 添加命令的完整别名路径，不附加命令组前缀；可多次使用。
func Alias(path ...string) RouteOption {
	snapshot := append([]string(nil), path...)
	return func(o *routeOptions) { o.aliases = append(o.aliases, snapshot) }
}

// Priority 设置路由优先级，数值越小越先；多次设置时取最后一次。
func Priority(p int32) RouteOption {
	return func(o *routeOptions) { o.priority = p }
}

// Args 依次指定参数名，未指定的位置沿用 arg1、arg2 等名称。
func Args(names ...string) RouteOption {
	snapshot := append([]string(nil), names...)
	return func(o *routeOptions) { o.names = snapshot }
}

// Optional 允许省略命名参数；def 最多给出一个缺失位置的默认值。
func Optional(name string, def ...string) RouteOption {
	snapshot := append([]string(nil), def...)
	return func(o *routeOptions) {
		o.optional = append(o.optional, optionalArgument{name: name, defaults: snapshot})
	}
}

// RawTail 让最后一个 string 参数接收剩余原文。
func RawTail() RouteOption { return func(o *routeOptions) { o.raw = true } }

// Usage 设置参数错误时回复的自定义用法。
func Usage(text string) RouteOption {
	return func(o *routeOptions) { o.usageMode, o.usage = pb.UsageMode_USAGE_MODE_CUSTOM, text }
}

// Silent 设置参数错误时保持沉默。
func Silent() RouteOption {
	return func(o *routeOptions) { o.usageMode = pb.UsageMode_USAGE_MODE_SILENT }
}

// IgnoreCase 对文本模式采用 Unicode 简单大小写折叠比较。
func IgnoreCase() RouteOption { return func(o *routeOptions) { o.ignoreCase = true } }

// CommandGroup 用公共路径前缀组织一组子命令。
type CommandGroup = CommandRegistry

// Group 在 name 下注册一组子命令。
func (r *CommandRegistry) Group(name string, fn func(*CommandGroup)) {
	sub := &CommandRegistry{prefix: append(append([]string{}, r.prefix...), name)}
	fn(sub)
	r.merge(sub)
}

// Command 注册一条命令。handler 必须是以下两种签名之一，参数类型限于 string/int64/float64/bool：
//
//	func(*sdk.Context, ...参数) (message.Chain, error)  // 自动回复
//	func(*sdk.Context, ...参数) error                   // 主动发送
func (r *CommandRegistry) Command(name string, handler any, opts ...RouteOption) {
	path := append(append([]string{}, r.prefix...), name)
	r.register(Command{Path: path, Handler: handler, spec: &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH}}, opts)
}

// Regex 注册 Go regexp 文本匹配，捕获组通过 Context.Match 获取。
func (r *CommandRegistry) Regex(routeID, expr string, handler any, opts ...RouteOption) {
	r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_REGEX, Regex: expr}, handler, opts)
}

// Keyword 注册包含任一模式的文本匹配。
func (r *CommandRegistry) Keyword(routeID string, patterns []string, handler any, opts ...RouteOption) {
	r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_KEYWORD, Patterns: append([]string(nil), patterns...)}, handler, opts)
}

// StartsWith 注册以任一模式开头的文本匹配。
func (r *CommandRegistry) StartsWith(routeID string, patterns []string, handler any, opts ...RouteOption) {
	r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_STARTSWITH, Patterns: append([]string(nil), patterns...)}, handler, opts)
}

// EndsWith 注册以任一模式结尾的文本匹配。
func (r *CommandRegistry) EndsWith(routeID string, patterns []string, handler any, opts ...RouteOption) {
	r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_ENDSWITH, Patterns: append([]string(nil), patterns...)}, handler, opts)
}

// FullMatch 注册与任一模式完全相同的文本匹配。
func (r *CommandRegistry) FullMatch(routeID string, patterns []string, handler any, opts ...RouteOption) {
	r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_FULLMATCH, Patterns: append([]string(nil), patterns...)}, handler, opts)
}

// Raw 注册普通命令与文本匹配之后执行的原始消息 Handler。
func (r *CommandRegistry) Raw(routeID string, handler any, opts ...RouteOption) {
	r.text(routeID, &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_RAW}, handler, opts)
}

func (r *CommandRegistry) text(routeID string, spec *pb.CommandSpec, handler any, opts []RouteOption) {
	if routeID == "" || strings.HasPrefix(routeID, commandRoutePrefix) {
		r.errs = append(r.errs, invalidCommand("text or raw route_id %q is empty or uses reserved %s prefix", routeID, commandRoutePrefix))
		return
	}
	r.register(Command{routeID: routeID, Handler: handler, spec: spec}, opts)
}

func (r *CommandRegistry) register(c Command, opts []RouteOption) {
	// 1. 从真实函数签名取得参数类型，选项仅描述名称与缺失时的行为。
	mode, params, err := checkHandler(c.Handler)
	if err != nil {
		r.errs = append(r.errs, invalidCommand("command %q: %v", commandRouteID(c), err))
		return
	}
	c.Mode, c.Params = mode, params
	var options routeOptions
	options.usageMode = pb.UsageMode_USAGE_MODE_DEFAULT
	for _, option := range opts {
		option(&options)
	}
	c.Aliases, c.Priority = options.aliases, options.priority
	spec := c.spec
	spec.Path, spec.HandlerMode = c.Path, mode
	spec.UsageMode, spec.CustomUsage, spec.IgnoreCase = options.usageMode, options.usage, options.ignoreCase
	if len(options.names) > len(params) {
		r.errs = append(r.errs, invalidCommand("command %q has more argument names than parameters", commandRouteID(c)))
		return
	}
	for i, param := range params {
		name := fmt.Sprintf("arg%d", i+1)
		if i < len(options.names) {
			name = options.names[i]
		}
		spec.Params = append(spec.Params, &pb.CommandParam{Name: name, Type: param, Required: true})
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
	return &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: fmt.Sprintf(format, args...)}
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
	// 2. 全部通过后统一登记，分组合并也走相同检查。
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

// merge 把子分组的命令与错误并入本注册器，重复检查跨分组生效。
func (r *CommandRegistry) merge(sub *CommandRegistry) {
	r.errs = append(r.errs, sub.errs...)
	for _, c := range sub.commands {
		r.add(c)
	}
}

// Handler 签名检查使用的反射类型，以及参数类型到协议参数类型的映射。
var (
	contextType = reflect.TypeOf((*Context)(nil))
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	chainType   = reflect.TypeOf(message.Chain(nil))
	paramTypes  = map[reflect.Type]pb.ParamType{
		reflect.TypeOf(""):         pb.ParamType_PARAM_TYPE_STRING,
		reflect.TypeOf(int64(0)):   pb.ParamType_PARAM_TYPE_INT,
		reflect.TypeOf(float64(0)): pb.ParamType_PARAM_TYPE_FLOAT,
		reflect.TypeOf(false):      pb.ParamType_PARAM_TYPE_BOOL,
	}
)

// checkHandler 校验两种 Handler 签名并返回模式与参数类型。
func checkHandler(h any) (pb.HandlerMode, []pb.ParamType, error) {
	t := reflect.TypeOf(h)
	if t == nil || t.Kind() != reflect.Func || reflect.ValueOf(h).IsNil() {
		return 0, nil, fmt.Errorf("handler must be a non-nil function, got %T", h)
	}
	if t.IsVariadic() || t.NumIn() == 0 || t.In(0) != contextType {
		return 0, nil, fmt.Errorf("handler's first parameter must be *sdk.Context and it must not be variadic: %s", t)
	}
	var params []pb.ParamType
	for i := 1; i < t.NumIn(); i++ {
		p, ok := paramTypes[t.In(i)]
		if !ok {
			return 0, nil, fmt.Errorf("type %[2]s of parameter %[1]d is not supported (only string/int64/float64/bool)", i, t.In(i))
		}
		params = append(params, p)
	}
	switch {
	case t.NumOut() == 1 && t.Out(0) == errorType:
		return pb.HandlerMode_HANDLER_MODE_MANUAL, params, nil
	case t.NumOut() == 2 && t.Out(0) == chainType && t.Out(1) == errorType:
		return pb.HandlerMode_HANDLER_MODE_AUTO_REPLY, params, nil
	}
	return 0, nil, fmt.Errorf("handler may only return error or (message.Chain, error): %s", t)
}

// HTTPRoute 记录一条插件 HTTP 路由；path 相对于 /plugins/{id}，端口由 SDK 托管（D-25）。
type HTTPRoute struct {
	Method string // 大写的 HTTP 方法
	Path   string // 插件相对路径，以 / 开头且为规范形式
	// Handler 处理请求。
	Handler http.Handler
	// Public 为 true 时无需 Core 会话即可访问。
	Public   bool
	AuthMode pb.HttpAuthMode // 认证方式；公开路由必须是回调验证方式之一
}

// HTTPOption 调整 HTTP 路由声明。
type HTTPOption func(*HTTPRoute)

// Public 声明公开回调路由；插件必须自行验证 state、签名或 nonce。
func Public(mode pb.HttpAuthMode) HTTPOption {
	return func(r *HTTPRoute) { r.Public, r.AuthMode = true, mode }
}

// HTTPRegistry 收集 HTTP 路由；默认需要 Core 授权会话。
type HTTPRegistry struct {
	routes []HTTPRoute
	errs   []error
}

// Handle 注册 method+path 的 Handler；同插件内重复的 method+path 报错。
func (r *HTTPRegistry) Handle(method, path string, h http.Handler, opts ...HTTPOption) {
	route := HTTPRoute{Method: strings.ToUpper(method), Path: path, Handler: h, AuthMode: pb.HttpAuthMode_HTTP_AUTH_MODE_CORE_SESSION}
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
	case route.Public && route.AuthMode != pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_STATE && route.AuthMode != pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_SIGNATURE:
		r.errs = append(r.errs, fmt.Errorf("public HTTP route %s %s must declare state or signature verification", method, path))
	case !route.Public && route.AuthMode != pb.HttpAuthMode_HTTP_AUTH_MODE_CORE_SESSION:
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
