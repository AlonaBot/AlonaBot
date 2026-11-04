// Package sdk 提供 AlonaBot Native Go 插件的公共入口。
//
// 插件的 main 调用 Run，传入嵌入了 [Base] 的根插件。Run 取得 runtime.lock，
// 收集根插件的注册项，与 Core 完成握手、准备和激活，然后一直运行到 Core 要求停止。
// Run 把 log/slog 的默认 Logger 换成写 stderr 的 JSON 行，按 Core 下发的级别过滤（CONTRACTS §14）。
// Base 为 [Root] 的全部 Hook 提供空的默认实现，根插件只覆盖需要的方法。
//
// 命令与事件 Handler 可通过 Context.Send 回复原事件目标；自动回复型命令返回 Chain，由 Core 发送一次。
// 文本匹配与 Raw 共用 Handler 调用上下文；HTTP 路由由 SDK 托管 Listener 并校验 Session 凭证。
// 定时任务尚未实现。
//
// 包的组成：
//   - 插件的 main 调用 [Run]；[Collect] 单独收集注册项，供检查与测试使用。
//   - [CommandRegistry]、[EventRegistry]、[HTTPRegistry] 与 [ScheduleRegistry] 收集命令、事件、HTTP 路由和定时任务。
//   - [ServiceRegistry] 收集类型化服务；[CallJSON] 与 [CallProto] 经 Core 调用目标插件。
//   - [SubscriptionRegistry] 收集 Event Bus 订阅；[Publish] 发布本插件命名空间下的 JSON 事件。
//   - [Context] 只在一次 Handler 调用期间有效，提供 [Context.Send] 与 [Context.StopPropagation]。
//   - 根插件嵌入 [Base]，覆盖 [Root] 中需要的 Hook。
package sdk

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// Error 携带协议错误码，供插件判断失败原因。
type Error struct {
	Code    pb.ErrorCode
	Message string
}

// Error 返回「错误码: 说明」。
func (e *Error) Error() string { return e.Code.String() + ": " + e.Message }

// ErrNotImplemented 表示该入口的运行行为尚未实现。
var ErrNotImplemented = &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "SDK runtime is not implemented yet"}

// Context 只在一次 Handler 调用期间有效。
type Context struct {
	ctx       context.Context // 本次调用的 context
	stopped   atomic.Bool     // 锁存的停止传播标记，一经设置不能清除
	active    atomic.Bool     // Handler 返回时关闭，阻止继续使用父调用
	inbound   *wire.Inbound
	caller    string // Core 下发的服务直接调用方
	event     *pb.Event
	match     []string // 本次 Regex 的捕获组
	autoReply bool     // 自动回复 Handler 通过返回 Chain 发送
}

// Context 返回本次调用的 context（含 deadline 与取消）。
func (c *Context) Context() context.Context { return c.ctx }

// Event 返回本次 Handler 接收的标准事件。
func (c *Context) Event() *pb.Event { return c.event }

// Logger 返回附带本次调用 trace_id、route_id 和 call_id 的默认 Logger。
func (c *Context) Logger() *slog.Logger {
	return slog.Default().With("trace_id", hex.EncodeToString(c.inbound.TraceID), "route_id", c.inbound.Target, "call_id", c.inbound.CallID)
}

// StopPropagation 锁存停止传播标记；一经设置不能重新打开传播，并且不受业务错误影响。
func (c *Context) StopPropagation() { c.stopped.Store(true) }

// SendOption 调整一次 Send。
type SendOption func(*sendOptions)

// sendOptions 汇总 SendOption 设置的发送参数。
type sendOptions struct {
	to *pb.Target // 显式目标；为 nil 时使用默认 Target
}

// To 指定显式目标；没有真实入站默认目标时必须使用。
func To(t *pb.Target) SendOption { return func(o *sendOptions) { o.to = t } }

// Send 发起一次逻辑发送。出错时仍可能返回已确认的部分结果（MESSAGE-MODEL §4）。
func (c *Context) Send(chain message.Chain, opts ...SendOption) (*pb.SendResult, error) {
	// 1. Handler 有效期与 Session 就绪状态均满足时才允许发起子请求。
	if !c.active.Load() {
		return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "handler context has expired"}
	}
	session, err := currentSession()
	if err != nil {
		return nil, err
	}
	options := &sendOptions{}
	for _, option := range opts {
		option(options)
	}
	request := &pb.SendRequest{To: options.to, Chain: chain.Proto()}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode SendRequest: " + err.Error()}
	}
	// 2. 父调用编号绑定 Core 记录的真实 Target；不回传插件自报的调用身份。
	response, callError := session.Call(c.ctx, wire.Request{Target: wire.ServiceSend,
		ParentCallID: c.inbound.CallID, TraceID: c.inbound.TraceID,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
	// 3. 即使出现 ErrorInfo，仍先解码并保留有效的部分发送结果。
	var result *pb.SendResult
	if response.Payload.SchemaID != "" {
		result = &pb.SendResult{}
		if err := decodeInvocation(response.Payload, result); err != nil {
			return nil, &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: err.Error()}
		}
	}
	if response.Err != nil {
		return result, &Error{Code: validate.PublicCode(response.Err.GetCode()), Message: response.Err.GetSafeMessage()}
	}
	if callError != nil {
		return result, &Error{Code: validate.PublicCode(wire.CodeOf(callError)), Message: callError.Error()}
	}
	return result, nil
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

	// 2. 收集静态注册项；HTTP 需要真实 Listener、调度需要 Core Scheduler，不静默丢弃这些声明。
	reg, err := Collect(root)
	if err != nil {
		return err
	}
	if len(reg.Schedules) > 0 {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "this SDK version does not support scheduled jobs yet"}
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
}

// Collect 调用根插件的注册入口，检查 Handler 签名与同插件内的重复项（D-20：同插件重复即报错）。
func Collect(root Root) (*Registration, error) {
	if root == nil {
		return nil, errors.New("root plugin is nil")
	}
	cmds, routes, jobs := &CommandRegistry{}, &HTTPRegistry{}, &ScheduleRegistry{}
	generated := generatedFor(root)
	generated.registerCommands(cmds)
	root.Commands(cmds)
	cmds.collect()
	events := &EventRegistry{seen: make(map[string]bool)}
	for _, command := range cmds.commands {
		events.seen[commandRouteID(command)] = true
	}
	generated.registerEvents(events)
	root.Events(events)
	events.collect()
	root.HTTP(routes)
	root.Schedules(jobs)
	// HTTP Listener 尚未创建，此处只核对跨注册器的 route_id。
	httpRoutes := &pb.Register{}
	appendHTTPRoutes(httpRoutes, routes.routes)
	for _, route := range httpRoutes.Routes {
		if events.seen[route.GetRouteId()] {
			routes.errs = append(routes.errs, fmt.Errorf("HTTP route %q conflicts with another route", route.GetRouteId()))
		}
	}
	reg := &Registration{Commands: cmds.commands, Events: events.events, HTTP: routes.routes, Schedules: jobs.jobs}
	errs := append(slices.Concat(cmds.errs, events.errs, routes.errs, jobs.errs), collectServices(root, reg))
	errs = append(errs, collectSubscriptions(root, reg))
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid registration of %T: %w", root, err)
	}
	return reg, nil
}
