// Package sdk 提供 AlonaBot Native Go 插件的公共入口。
//
// 插件的 New 构造 [Plugin]，登记路由和生命周期回调，构建资源由 CLI 自动注入。
// main 调用 Run(New())，取得 runtime.lock，与 Core 完成握手、准备和激活，再处理调用直到停止。
// Run 把 log/slog 的默认 Logger 换成写 stderr 的 JSON 行，按 Core 下发的级别过滤。
//
// 包的组成：[Plugin] 持有登记状态，[SubPlugin] 提供一层逻辑子插件，[CommandGroup] 组织命令路径与权限；
// 注册收集在运行前统一检查；类型化服务、事件、任务与按钮分别由 service、bus、task、button 子包声明。
// [Context] 在一次 Rule 与 Handler 调用期间有效，提供配置、翻译、日志、发送、编辑和停止传播入口。
// 自动回复型命令返回 [message.Content]，由 Core 在显式发送完成后发送一次。
// [Context.Stream] 按当前事件实例的能力发送文本增量并节流编辑。
// HTTP 路由由 SDK 托管 Listener 并校验 Session 凭证；定时任务由 Core Scheduler 触发。
// 组件生命周期、配置、迁移、路由与参数的用法见 https://alonabot.github.io/docs/tutorial/project。
package sdk

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/internal/dbhook"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Error 携带协议错误码，供插件判断失败原因。
type Error = errcode.Error

// ErrNotImplemented 表示该入口的运行行为尚未实现。
var ErrNotImplemented = &Error{Code: errcode.Unsupported, Message: "SDK runtime is not implemented yet"}

// Context 只在一次 Handler 调用期间有效。
type Context struct {
	plugin          *Plugin
	subID           string
	profileID       string          // Core 选择的配置文件；空串表示 default。
	config          config.Values   // 创建调用时冻结的完整插件视图。
	ctx             context.Context // 本次调用的 context
	propagationMu   sync.Mutex      // 串行化 StopPropagation 与调用结束的传播快照。
	stopped         bool            // 由 propagationMu 保护，只在有效调用内锁存。
	active          atomic.Bool     // Handler 返回时关闭，阻止继续使用父调用
	inbound         *wire.Inbound
	caller          string      // Core 下发的服务直接调用方
	actor           *pb.Actor   // Service 继承的可信 Actor；事件调用从 event 读取。
	defaultTarget   *pb.Target  // Service 继承的可信默认目标；事件调用从 event 读取。
	chatType        pb.ChatType // 无事件时继承的可信会话类型。
	commandConsumed bool        // Core 下发的本条消息命令消费结果。
	event           *pb.Event
	eventOnce       sync.Once
	publicEvent     *message.Event
	match           []string               // 本次 Regex 的捕获组
	autoReply       bool                   // 自动回复 Handler 通过返回 Content 发送
	continuations   *continuationFunctions // 所属 Session 的延续函数登记表。
	continuation    bool
	inHandler       bool       // Rule 执行完成后开放 Continue。
	continueMu      sync.Mutex // 串行化一次性登记。
	continued       bool
}

// Context 返回本次调用的 context（含 deadline 与取消）。
func (c *Context) Context() context.Context { return c.ctx }

// CommandConsumed 返回 Core 命令层是否已消费本条消息。
func (c *Context) CommandConsumed() bool { return c.commandConsumed }

// Event 返回本次 Handler 接收的标准事件。
func (c *Context) Event() *message.Event {
	c.eventOnce.Do(func() { c.publicEvent = convert.EventFromProto(c.event) })
	return c.publicEvent
}

// User 返回本次调用触发者的用户字符串，没有 Actor 时为空。
func (c *Context) User() string {
	actor := c.actor
	adapterID := c.defaultTarget.GetAdapterId()
	if c.event != nil {
		actor = c.event.Actor
		adapterID = c.event.GetAdapterId()
	}
	if actor == nil || (c.event == nil && c.defaultTarget == nil) {
		return ""
	}
	return message.UID(adapterID, actor.ExternalUserId)
}

// Session 返回本次调用的默认会话，没有默认目标时为空。
func (c *Context) Session() string {
	return convert.SessionFromProto(c.target())
}

// Instance 返回当前事件目标或继承的默认目标的实例名；没有目标时为空。
func (c *Context) Instance() string { return c.target().GetAdapterInstanceId() }

func (c *Context) target() *pb.Target {
	target := c.defaultTarget
	if c.event != nil {
		target = c.event.Target
	}
	return target
}

// Reply 将文本与后续片段作为一条消息回复到当前默认会话。
func (c *Context) Reply(text string, more ...message.Segment) error {
	chain := append(message.Chain{message.Text(text)}, more...)
	_, err := c.Send(chain)
	return err
}

// Logger 返回附带本次调用 trace_id、route_id 和 call_id 的默认 Logger。
func (c *Context) Logger() *slog.Logger {
	logger := componentLogger(c.ctx).With("trace_id", hex.EncodeToString(c.inbound.TraceID), "route_id", c.inbound.Target, "call_id", c.inbound.CallID)
	if c.subID != "" {
		logger = logger.With("sub_plugin", c.subID)
	}
	return logger
}

// StopPropagation 锁存停止传播标记；一经设置不能重新打开传播，并且不受业务错误影响。
func (c *Context) StopPropagation() {
	c.propagationMu.Lock()
	defer c.propagationMu.Unlock()
	if c.active.Load() {
		c.stopped = true
	}
}

// SendOption 调整一次消息发送或平台动作的机器人实例。
type SendOption func(*sendOptions)

type sendOptions struct {
	via                 string
	actionDefaultTarget bool // 仅绑定客户端按声明选择 Handler 的默认目标。
}

// Via 指定消息发送或平台动作的机器人实例。
func Via(instanceID string) SendOption { return func(o *sendOptions) { o.via = instanceID } }

// Send 向当前事件的默认会话发送内容，保留已确认的部分结果。
func (c *Context) Send(m message.Content, opts ...SendOption) (*message.SendResult, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	if c.target() == nil {
		return nil, errcode.NoDefaultTarget.New("no default target")
	}
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	var target *pb.Target
	if options.via != "" {
		var err error
		target, err = convert.TargetToProto(c.Session(), options.via)
		if err != nil {
			return nil, err
		}
	}
	return callSend(c.ctx, target, m, c.inbound, c.target())
}

// SendTo 向指定会话发送内容，关联当前 Handler 的父调用并保留部分结果。
func (c *Context) SendTo(sid string, m message.Content, opts ...SendOption) (*message.SendResult, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	target, err := convert.TargetToProto(sid, options.via)
	if err != nil {
		return nil, err
	}
	return callSend(c.ctx, target, m, c.inbound, c.target())
}

// Run 作为插件 main 的入口：连接 Core、完成握手并处理请求，直到停止。
// 顺序：取得 runtime.lock → 收集注册项 → 握手 → 准备阶段 → 激活（OnReady/ACTIVATED）→ 服务。
func Run(root *Plugin) error {
	if err := requireResources(root); err != nil {
		return err
	}
	root.freeze()
	setupLogging()
	l, err := readLaunch()
	if err != nil {
		return err
	}
	return run(root, l)
}

func run(root *Plugin, l launch) (err error) {
	// 1. 先取 runtime.lock，防止旧实例仍在执行用户代码。
	lock, err := acquireRuntimeLock(l.runtimeLockPath)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Unlock()) }()
	l.runtime = newComponentSession(l.inProcess)
	defer func() { err = errors.Join(err, l.runtime.close()) }()

	// 2. 收集静态注册项；HTTP 路由与定时任务随 REGISTER 一并声明。
	reg, err := collect(root)
	if err != nil {
		return err
	}
	l.serviceSchemas = generatedFor(root).ServiceSchemas
	l.entryPackage = entryPackage(root)

	// 3. 连接 Core 并依次完成握手、准备与激活，然后进入服务循环。
	l.http = root.http.newRuntime(reg.HTTP, l.runtime)
	defer func() { l.closeHTTP(time.Now().Add(onStopBudget)) }()
	conn, err := dial(l)
	if err != nil {
		return err
	}
	defer conn.Close()
	ack, err := clientHandshake(conn, l)
	if err != nil {
		return err
	}
	l.runtime.configure(l, ack, "plugin")
	if _, err := prepare(conn, l, ack, root, reg); err != nil {
		return err
	}
	sess, a, err := activate(conn, l, ack, root, reg)
	if err != nil {
		return err
	}
	return serve(sess, a) // 收到 STOPPING 正常返回；断链时执行 OnStop 后返回错误，随后进程退出
}

// registration 汇总从根插件收集到的静态注册项。
type registration struct {
	plugin        *Plugin
	subPlugins    map[string]string
	Commands      []command
	Events        []eventRoute
	HTTP          []httpRoute
	Schedules     []schedule
	services      []serviceRoute
	subscriptions []subscriptionRoute
	hooks         []hookRoute
	tasks         []taskRoute
	messages      *i18n.Catalog // 收集时解析，供本次 Session 只读使用。
}

// collect 检查插件对象的全部登记、资源与跨路由冲突。
func collect(root *Plugin) (*registration, error) {
	if root == nil {
		return nil, errors.New("plugin is nil")
	}
	cmds, events, routes, jobs := &root.commands, &root.events, &root.http, &root.schedules
	routes.collect()
	jobs.collect()
	cmds.collect()
	events.seen = make(map[string]bool)
	for _, command := range cmds.commands {
		events.seen[commandRouteID(command)] = true
	}
	events.collect()
	reg := &registration{plugin: root, subPlugins: make(map[string]string), Commands: cmds.commands, Events: events.events, HTTP: routes.routes, Schedules: jobs.jobs}
	errs := slices.Concat(root.errs, cmds.errs, events.errs, routes.errs, jobs.errs)
	errs = append(errs, root.checkSubPlugins(), dbhook.Validate(root), collectServices(root, reg), collectSubscriptions(root, reg), collectI18n(root, reg))
	errs = append(errs, collectHooks(root, reg))
	errs = append(errs, collectTasks(root, reg))
	// HTTP Listener 在准备阶段创建；这里只核对所有种类共享的 route_id。
	seen := make(map[string]bool)
	for _, route := range buildRegister(reg, "{}", nil).GetRoutes() {
		if spec := route.GetHttp(); spec.GetAuthMode() == pb.HttpAuthMode_HTTP_AUTH_MODE_PLATFORM_TOKEN {
			errs = append(errs, fmt.Errorf("HTTP route %s %s: platform token authentication is only available to adapters", spec.GetMethod(), spec.GetPath()))
		}
		id := route.GetRouteId()
		if seen[id] {
			errs = append(errs, fmt.Errorf("route %q registered twice; use .ID to distinguish routes", id))
		}
		seen[id] = true
	}
	for _, c := range reg.Commands {
		reg.subPlugins[commandRouteID(c)] = c.subID
	}
	for _, e := range reg.Events {
		reg.subPlugins[e.RouteID] = e.subID
	}
	for _, job := range reg.Schedules {
		reg.subPlugins[scheduleRoutePrefix+job.JobID] = root.jobSubPlugins[job.JobID]
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid plugin registration: %w", err)
	}
	return reg, nil
}

// Config 返回整个插件在本次调用开始时冻结的配置文件视图。
func (c *Context) Config() config.Values { return c.config }

// handlerParamNames 在收集时解析函数身份，运行时绑定只读已收集的 CommandSpec。
func handlerParamNames(g generated.Resources, handler any) []string {
	if len(g.ParamNames) == 0 {
		return nil
	}
	fn := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
	packagePath, function := runtimeFunctionParts(handlerFunctionName(fn.Name()))
	if names, ok := g.ParamNames[packagePath+function]; ok {
		return names
	}
	// main 在普通二进制中用 main 前缀，在 go test 中用包的导入路径。
	if names, ok := g.ParamNames["main"+function]; ok {
		return names
	}
	file, line := fn.FileLine(fn.Entry())
	suffix := fmt.Sprintf(":%s:%d", filepath.Base(file), line)
	if names, ok := g.ParamNames[packagePath+suffix]; ok {
		return names
	}
	return g.ParamNames["main"+suffix]
}

// runtimeFunctionParts 先按符号分隔符切开包与函数，再还原包路径中的 Go %xx 转义。
func runtimeFunctionParts(name string) (pkgPath, function string) {
	start := strings.LastIndexByte(name, '/') + 1
	end := start + strings.IndexByte(name[start:], '.')
	// 编译器生成的包路径转义总是合法；函数部分保留符号原文。
	pkgPath, _ = url.PathUnescape(name[:end])
	return pkgPath, name[end:]
}

// handlerFunctionName 去掉实例类型实参和方法值包装后缀，对齐源码声明键。
func handlerFunctionName(name string) string {
	name = strings.TrimSuffix(name, "-fm")
	if !strings.ContainsRune(name, '[') {
		return name
	}
	var out strings.Builder
	depth := 0
	for _, char := range name {
		switch char {
		case '[':
			depth++
		case ']':
			depth--
		default:
			if depth == 0 {
				out.WriteRune(char)
			}
		}
	}
	return out.String()
}
