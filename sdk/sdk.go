// Package sdk 提供 AlonaBot Native Go 插件的公共入口。
//
// 插件的 main 调用 Run，传入嵌入了 [Base] 的根插件。Run 取得 runtime.lock，
// 收集根插件的注册项，与 Core 完成握手、准备和激活，然后一直运行到 Core 要求停止。
// Run 把 log/slog 的默认 Logger 换成写 stderr 的 JSON 行，按 Core 下发的级别过滤（CONTRACTS §14）。
// Base 为 [Root] 的全部 Hook 提供空的默认实现，根插件只覆盖需要的方法。
//
// 命令与事件 Handler 可通过 [Context.Send] 回复原事件目标，通过 [Context.Edit] 编辑已发送的文本消息；
// 自动回复型命令返回 Content，由 Core 发送一次。后台 [Send] 显式给出会话，[Edit] 给出消息句柄。
// 文本匹配与 Raw 共用 Handler 调用上下文；HTTP 路由由 SDK 托管 Listener 并校验 Session 凭证。
// [Context.T] 按 Actor 与 Core 默认语言格式化文案；后台调用 [T] 使用 Core 默认语言。
// 定时任务随 REGISTER 声明，由 Core Scheduler 触发；其 Handler 没有 Actor 与默认 Target，Send 须显式 To。
//
// 包的组成：
//   - 插件的 main 调用 [Run]；[Collect] 单独收集注册项，供检查与测试使用。
//   - [CommandRegistry]、[EventRegistry]、[HTTPRegistry] 与 [ScheduleRegistry] 收集命令、事件、HTTP 路由和定时任务。
//   - [ServiceRegistry] 收集类型化服务；[CallJSON] 与 [CallProto] 经 Core 调用目标插件。
//   - [SubscriptionRegistry] 收集 Event Bus 订阅；[Publish] 发布本插件命名空间下的 JSON 事件。
//   - [Context] 只在一次 Handler 调用期间有效，提供 [Context.Send]、[Context.Edit] 与 [Context.StopPropagation]。
//   - 根插件嵌入 [Base]，覆盖 [Root] 中需要的 Hook。
package sdk

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Error 携带协议错误码，供插件判断失败原因。
type Error = errcode.Error

// ErrNotImplemented 表示该入口的运行行为尚未实现。
var ErrNotImplemented = &Error{Code: errcode.Unsupported, Message: "SDK runtime is not implemented yet"}

// Context 只在一次 Handler 调用期间有效。
type Context struct {
	ctx           context.Context // 本次调用的 context
	propagationMu sync.Mutex      // 串行化 StopPropagation 与调用结束的传播快照。
	stopped       bool            // 由 propagationMu 保护，只在有效调用内锁存。
	active        atomic.Bool     // Handler 返回时关闭，阻止继续使用父调用
	inbound       *wire.Inbound
	caller        string     // Core 下发的服务直接调用方
	actor         *pb.Actor  // Service 继承的可信 Actor；事件调用从 event 读取。
	defaultTarget *pb.Target // Service 继承的可信默认目标；事件调用从 event 读取。
	event         *pb.Event
	eventOnce     sync.Once
	publicEvent   *message.Event
	match         []string // 本次 Regex 的捕获组
	autoReply     bool     // 自动回复 Handler 通过返回 Content 发送
}

// Context 返回本次调用的 context（含 deadline 与取消）。
func (c *Context) Context() context.Context { return c.ctx }

// Event 返回本次 Handler 接收的标准事件。
func (c *Context) Event() *message.Event {
	c.eventOnce.Do(func() { c.publicEvent = convert.EventFromProto(c.event) })
	return c.publicEvent
}

// User 返回本次调用触发者的用户字符串，没有 Actor 时为空。
func (c *Context) User() message.User {
	actor := c.actor
	adapterID := c.defaultTarget.GetAdapterId()
	if c.event != nil {
		actor = c.event.Actor
		adapterID = c.event.GetAdapterId()
	}
	if actor == nil || (c.event == nil && c.defaultTarget == nil) {
		return ""
	}
	return message.NewUser(adapterID, actor.ExternalUserId)
}

// Session 返回本次调用的默认会话，没有默认目标时为空。
func (c *Context) Session() message.Session {
	return convert.SessionFromProto(c.target())
}

func (c *Context) target() *pb.Target {
	target := c.defaultTarget
	if c.event != nil {
		target = c.event.Target
	}
	return target
}

// Reply 向当前默认会话回复一段文本。
func (c *Context) Reply(text string) error { _, err := c.Send(message.Text(text)); return err }

// Logger 返回附带本次调用 trace_id、route_id 和 call_id 的默认 Logger。
func (c *Context) Logger() *slog.Logger {
	return slog.Default().With("trace_id", hex.EncodeToString(c.inbound.TraceID), "route_id", c.inbound.Target, "call_id", c.inbound.CallID)
}

// StopPropagation 锁存停止传播标记；一经设置不能重新打开传播，并且不受业务错误影响。
func (c *Context) StopPropagation() {
	c.propagationMu.Lock()
	defer c.propagationMu.Unlock()
	if c.active.Load() {
		c.stopped = true
	}
}

// SendOption 调整一次消息发送的会话或机器人实例。
type SendOption func(*sendOptions)

type sendOptions struct {
	to       message.Session
	explicit bool
	via      string
}

// To 指定发送会话，实例缺省由 Core 选择。
func To(s message.Session) SendOption { return func(o *sendOptions) { o.to = s; o.explicit = true } }

// Via 指定发送的机器人实例。
func Via(instanceID string) SendOption { return func(o *sendOptions) { o.via = instanceID } }

// Send 向当前会话或显式目标发送内容，保留已确认的部分结果。
func (c *Context) Send(m message.Content, opts ...SendOption) (*message.SendResult, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	var target *pb.Target
	if options.explicit || options.via != "" {
		to := options.to
		if !options.explicit {
			to = c.Session()
			if to == "" {
				return nil, errcode.NoDefaultTarget.New("no default target")
			}
		}
		var err error
		target, err = convert.TargetToProto(to, options.via)
		if err != nil {
			return nil, err
		}
	}
	return callSend(c.ctx, target, m, c.inbound, c.target())
}

// Run 作为插件 main 的入口：连接 Core、完成握手并处理请求，直到停止。
// 顺序：取得 runtime.lock → 收集注册项 → 握手 → 准备阶段 → 激活（OnReady/ACTIVATED）→ 服务。
func Run(root Root) error {
	setupLogging()
	// 1. 先取 runtime.lock：注册入口（Commands 等）也是用户代码，旧实例仍在时不能执行（CONTRACTS §12.2）。
	lock, err := acquireRuntimeLock()
	if err != nil {
		return err
	}
	defer lock.Unlock()

	// 2. 收集静态注册项；HTTP 路由与定时任务随 REGISTER 一并声明。
	reg, err := Collect(root)
	if err != nil {
		return err
	}

	// 3. 连接 Core 并依次完成握手、准备与激活，然后进入服务循环。
	l, err := readLaunch()
	if err != nil {
		return err
	}
	l.http = newHTTPListener(reg.HTTP)
	defer func() { l.http.close(time.Now().Add(onStopBudget)) }()
	conn, err := dial(l)
	if err != nil {
		return err
	}
	defer conn.Close()
	ack, err := clientHandshake(conn, l)
	if err != nil {
		return err
	}
	setLogLevel(ack.GetBootstrap().GetLogLevel())
	if _, err := prepare(conn, l, ack, root, reg); err != nil {
		return err
	}
	sess, a, err := activate(conn, l, ack, root, reg)
	if err != nil {
		return err
	}
	return serve(sess, a) // 收到 STOPPING 正常返回；断链时执行 OnStop 后返回错误，随后进程退出
}

// Registration 汇总从根插件收集到的静态注册项。
type Registration struct {
	Commands      []Command
	Events        []EventRoute
	HTTP          []HTTPRoute
	Schedules     []Schedule
	services      []serviceRoute
	subscriptions []subscriptionRoute
	messages      *i18n.Catalog // 收集时解析，供本次 Session 只读使用。
}

// Collect 调用根插件的注册入口，检查 Handler 签名、usage 文案键与同插件内的重复项（D-20：同插件重复即报错）。
func Collect(root Root) (*Registration, error) {
	if root == nil {
		return nil, errors.New("root plugin is nil")
	}
	cmds, routes, jobs := &CommandRegistry{}, &HTTPRegistry{}, &ScheduleRegistry{}
	root.Commands(cmds)
	cmds.collect()
	events := &EventRegistry{seen: make(map[string]bool)}
	for _, command := range cmds.commands {
		events.seen[commandRouteID(command)] = true
	}
	root.Events(events)
	events.collect()
	root.HTTP(routes)
	root.Schedules(jobs)
	// HTTP Listener 尚未创建，此处只核对跨注册器的 route_id。
	httpRoutes := &pb.Register{}
	appendHTTPRoutes(httpRoutes, routes.routes)
	for _, route := range httpRoutes.Routes {
		if spec := route.GetHttp(); spec.GetAuthMode() == pb.HttpAuthMode_HTTP_AUTH_MODE_PLATFORM_TOKEN {
			routes.errs = append(routes.errs, fmt.Errorf("HTTP route %s %s: platform token authentication is only available to adapters", spec.GetMethod(), spec.GetPath()))
		}
		if events.seen[route.GetRouteId()] {
			routes.errs = append(routes.errs, fmt.Errorf("HTTP route %q conflicts with another route", route.GetRouteId()))
		}
	}
	reg := &Registration{Commands: cmds.commands, Events: events.events, HTTP: routes.routes, Schedules: jobs.jobs}
	errs := append(slices.Concat(cmds.errs, events.errs, routes.errs, jobs.errs), collectServices(root, reg))
	errs = append(errs, collectSubscriptions(root, reg), collectI18n(root, reg))
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid registration of %T: %w", root, err)
	}
	return reg, nil
}
