// 本文件负责收集根插件的静态注册项：命令、HTTP 路由与定时任务，并在收集时检查签名与重复。
package sdk

import (
	"fmt"
	"net/http"
	pathpkg "path"
	"reflect"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Command 是一条已收集的命令注册。
type Command struct {
	// Path 是完整命令路径（含分组前缀），大小写敏感。
	Path []string
	// Handler 是已通过签名检查的处理函数。
	Handler any
	// Mode 是由签名推出的 Handler 形态。
	Mode pb.HandlerMode
	// Params 是 Context 之后各参数的类型。
	Params []pb.ParamType
}

// CommandRegistry 收集命令；实例方法以方法值显式传入，不做自动 DI。
type CommandRegistry struct {
	// prefix 是本分组的公共路径前缀。
	prefix []string
	// commands 是已收集的命令，按注册顺序排列。
	commands []Command
	// errs 是收集过程中发现的错误，由 Collect 一并返回。
	errs []error
	// seen 是已注册的完整路径，用于检查重复；惰性创建。
	seen map[string]bool
}

// CommandGroup 是带公共路径前缀的命令组。
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
func (r *CommandRegistry) Command(name string, handler any) {
	path := append(append([]string{}, r.prefix...), name)
	mode, params, err := checkHandler(handler)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("命令 /%s: %w", strings.Join(path, " "), err))
		return
	}
	r.add(Command{Path: path, Handler: handler, Mode: mode, Params: params})
}

// add 检查路径 token 并去重后登记命令；问题记入 errs。
func (r *CommandRegistry) add(c Command) {
	key := strings.Join(c.Path, " ")
	for _, tok := range c.Path {
		if tok == "" || strings.ContainsAny(tok, " \t\n") {
			r.errs = append(r.errs, fmt.Errorf("命令路径 %q 含空 token 或空白", key))
			return
		}
	}
	if r.seen == nil {
		r.seen = map[string]bool{}
	}
	if r.seen[key] {
		r.errs = append(r.errs, fmt.Errorf("命令 /%s 重复注册", key))
		return
	}
	r.seen[key] = true
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
		return 0, nil, fmt.Errorf("Handler 必须是非 nil 函数，实际 %T", h)
	}
	if t.IsVariadic() || t.NumIn() == 0 || t.In(0) != contextType {
		return 0, nil, fmt.Errorf("Handler 的第一个参数必须是 *sdk.Context，且不能是可变参数函数：%s", t)
	}
	var params []pb.ParamType
	for i := 1; i < t.NumIn(); i++ {
		p, ok := paramTypes[t.In(i)]
		if !ok {
			return 0, nil, fmt.Errorf("参数 %d 的类型 %s 不受支持（只支持 string/int64/float64/bool）", i, t.In(i))
		}
		params = append(params, p)
	}
	switch {
	case t.NumOut() == 1 && t.Out(0) == errorType:
		return pb.HandlerMode_HANDLER_MODE_MANUAL, params, nil
	case t.NumOut() == 2 && t.Out(0) == chainType && t.Out(1) == errorType:
		return pb.HandlerMode_HANDLER_MODE_AUTO_REPLY, params, nil
	}
	return 0, nil, fmt.Errorf("Handler 只能返回 error 或 (message.Chain, error)：%s", t)
}

// HTTPRoute 是一条插件 HTTP 路由；path 相对于 /plugins/{id}，端口由 SDK 托管（D-25）。
type HTTPRoute struct {
	// Method 是大写的 HTTP 方法。
	Method string
	// Path 是插件相对路径，以 / 开头且为规范形式。
	Path string
	// Handler 处理请求。
	Handler http.Handler
	// Public 为 true 时无需 Core 会话即可访问。
	Public bool
	// AuthMode 是认证方式；公开路由必须是回调验证方式之一。
	AuthMode pb.HttpAuthMode
}

// HTTPOption 调整 HTTP 路由声明。
type HTTPOption func(*HTTPRoute)

// Public 声明公开回调路由；插件必须自行验证 state、签名或 nonce。
func Public(mode pb.HttpAuthMode) HTTPOption {
	return func(r *HTTPRoute) { r.Public, r.AuthMode = true, mode }
}

// HTTPRegistry 收集 HTTP 路由；默认需要 Core 授权会话。
type HTTPRegistry struct {
	// routes 是已收集的路由。
	routes []HTTPRoute
	// errs 是收集过程中发现的错误。
	errs []error
}

// Handle 注册 method+path 的 Handler；同插件内重复的 method+path 报错。
func (r *HTTPRegistry) Handle(method, path string, h http.Handler, opts ...HTTPOption) {
	route := HTTPRoute{Method: strings.ToUpper(method), Path: path, Handler: h, AuthMode: pb.HttpAuthMode_HTTP_AUTH_MODE_CORE_SESSION}
	for _, o := range opts {
		o(&route)
	}
	switch {
	case h == nil:
		r.errs = append(r.errs, fmt.Errorf("HTTP %s %s 的 Handler 为 nil", method, path))
	case !strings.HasPrefix(path, "/") || pathpkg.Clean(path) != path:
		// 非规范写法（//、..、尾部 /）会让同一资源有多个路径，代理匹配时产生歧义。
		r.errs = append(r.errs, fmt.Errorf("HTTP 路径 %q 必须以 / 开头且是规范形式", path))
	case route.Public && route.AuthMode == pb.HttpAuthMode_HTTP_AUTH_MODE_CORE_SESSION:
		r.errs = append(r.errs, fmt.Errorf("公开路由 %s %s 必须声明回调验证方式", method, path))
	default:
		for _, e := range r.routes {
			if e.Method == route.Method && e.Path == route.Path {
				r.errs = append(r.errs, fmt.Errorf("HTTP %s %s 重复注册", route.Method, path))
				return
			}
		}
		r.routes = append(r.routes, route)
	}
}

// Schedule 是一个定时任务声明；interval 与 cron 互斥，触发权属于 Core。
type Schedule struct {
	// JobID 是插件内稳定的任务 ID。
	JobID string
	// Interval 是固定间隔；为零表示使用 Cron。
	Interval time.Duration
	// Cron 是五字段 cron 表达式；为空表示使用 Interval。
	Cron string
	// Timezone 是 cron 使用的 IANA 时区，空串表示 UTC。
	Timezone string
	// Handler 是每次触发时调用的函数。
	Handler func(*Context) error
}

// ScheduleRegistry 收集定时任务。
type ScheduleRegistry struct {
	// jobs 是已收集的任务。
	jobs []Schedule
	// errs 是收集过程中发现的错误。
	errs []error
}

// Every 声明固定间隔任务（至少 1s）。
func (r *ScheduleRegistry) Every(jobID string, d time.Duration, fn func(*Context) error) {
	if d < time.Second {
		r.errs = append(r.errs, fmt.Errorf("任务 %s 的间隔 %s 小于 1s", jobID, d))
		return
	}
	r.add(Schedule{JobID: jobID, Interval: d, Handler: fn})
}

// Cron 声明五字段 cron 任务；timezone 为 IANA 名称，空串表示 UTC。表达式在 REGISTER 时由 Core 校验。
func (r *ScheduleRegistry) Cron(jobID, expr, timezone string, fn func(*Context) error) {
	if len(strings.Fields(expr)) != 5 {
		r.errs = append(r.errs, fmt.Errorf("任务 %s 的 cron 表达式 %q 不是五字段", jobID, expr))
		return
	}
	r.add(Schedule{JobID: jobID, Cron: expr, Timezone: timezone, Handler: fn})
}

// add 检查 job_id、Handler 并去重后登记任务；问题记入 errs。
func (r *ScheduleRegistry) add(s Schedule) {
	switch {
	case s.JobID == "":
		r.errs = append(r.errs, fmt.Errorf("定时任务缺少 job_id"))
		return
	case s.Handler == nil:
		r.errs = append(r.errs, fmt.Errorf("任务 %s 的 Handler 为 nil", s.JobID))
		return
	}
	for _, j := range r.jobs {
		if j.JobID == s.JobID {
			r.errs = append(r.errs, fmt.Errorf("任务 %s 重复注册", s.JobID))
			return
		}
	}
	r.jobs = append(r.jobs, s)
}
