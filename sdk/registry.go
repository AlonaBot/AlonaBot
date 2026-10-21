package sdk

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/alonabot/sdk/message"
)

// Command 是一条已收集的命令注册。
type Command struct {
	Path    []string
	Handler any
	Mode    pb.HandlerMode
	Params  []pb.ParamType
}

// CommandRegistry 收集命令；实例方法以方法值显式传入，不做自动 DI。
type CommandRegistry struct {
	prefix   []string
	commands []Command
	errs     []error
	seen     map[string]bool
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

func (r *CommandRegistry) merge(sub *CommandRegistry) {
	r.errs = append(r.errs, sub.errs...)
	for _, c := range sub.commands {
		r.add(c)
	}
}

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
	Method, Path string
	Handler      http.Handler
	Public       bool
	AuthMode     pb.HttpAuthMode
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
		r.errs = append(r.errs, fmt.Errorf("HTTP %s %s 的 Handler 为 nil", method, path))
	case !strings.HasPrefix(path, "/"):
		r.errs = append(r.errs, fmt.Errorf("HTTP 路径 %q 必须以 / 开头", path))
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
	JobID    string
	Interval time.Duration
	Cron     string
	Timezone string
	Handler  func(*Context) error
}

// ScheduleRegistry 收集定时任务。
type ScheduleRegistry struct {
	jobs []Schedule
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
