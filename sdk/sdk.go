// Package sdk 提供 AlonaBot Native Go 插件的公共入口。
//
// 插件的 main 调用 Run，传入嵌入了 plugin.Base 的根插件。Run 取得 runtime.lock，
// 收集根插件的注册项，与 Core 完成握手、准备和激活，然后一直运行到 Core 要求停止。
// 根插件可以按需实现 Starter、Readier、Stopper 等生命周期接口。
//
// 业务分发、Send、HTTP 路由与定时任务尚未实现；它们返回 UNSUPPORTED 错误，不会 panic，也不会假装成功。
//
// 包的组成：
//   - 插件的 main 调用 [Run]；[Collect] 单独收集注册项，供检查与测试使用。
//   - [CommandRegistry]、[HTTPRegistry] 与 [ScheduleRegistry] 收集命令、HTTP 路由和定时任务。
//   - [Context] 只在一次 Handler 调用期间有效，提供 [Context.Send] 与 [Context.StopPropagation]。
//   - 根插件可以按需实现 [Starter]、[Readier]、[Stopper] 这几个生命周期接口。
package sdk

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// Error 携带协议错误码，供插件判断失败原因。
type Error struct {
	Code    pb.ErrorCode
	Message string
}

// Error 返回「错误码: 说明」。
func (e *Error) Error() string { return e.Code.String() + ": " + e.Message }

// ErrNotImplemented 表示该入口的运行行为尚未实现。
var ErrNotImplemented = &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "SDK 运行时尚未实现"}

// 可选的进程生命周期 Hook（CONTRACTS §4.1）；没有实现时无需空方法。
type (
	// Starter 的 OnStart 在准备阶段、REGISTER 之前执行；失败时本次启动失败。
	Starter interface {
		// OnStart 在配置准备之后运行，不能发起业务调用。
		OnStart(context.Context) error
	}
	// Readier 的 OnReady 在收到 ACTIVATE 之后执行；成功后才写出 ACTIVATED。
	Readier interface {
		// OnReady 在 OnReady 预算内完成就绪工作。
		OnReady(context.Context) error
	}
	// Stopper 的 OnStop 在 STOPPING 或断链后执行，使用独立的剩余停止期限。
	Stopper interface {
		// OnStop 关闭资源，最多使用 STOPPING 给出的期限。
		OnStop(context.Context) error
	}
)

// 根插件可选实现的注册入口；实例方法通过这些入口显式绑定。
type (
	// CommandProvider 注册命令。
	CommandProvider interface {
		// Commands 在 Collect 时调用一次。
		Commands(*CommandRegistry)
	}
	// HTTPProvider 注册 HTTP 路由。
	HTTPProvider interface {
		// HTTP 在 Collect 时调用一次。
		HTTP(*HTTPRegistry)
	}
	// ScheduleProvider 注册定时任务。
	ScheduleProvider interface {
		// Schedules 在 Collect 时调用一次。
		Schedules(*ScheduleRegistry)
	}
)

// Context 只在一次 Handler 调用期间有效。
type Context struct {
	ctx     context.Context // 本次调用的 context
	stopped bool            // 锁存的停止传播标记，一经设置不能清除
}

// Context 返回本次调用的 context（含 deadline 与取消）。
func (c *Context) Context() context.Context { return c.ctx }

// StopPropagation 锁存停止传播标记；一经设置不能重新打开传播，并且不受业务错误影响。
func (c *Context) StopPropagation() { c.stopped = true }

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
	return nil, ErrNotImplemented
}

// Run 作为插件 main 的入口：连接 Core、完成握手并处理请求，直到停止。
// 顺序：取得 runtime.lock → 收集注册项 → 握手 → 准备阶段 → 激活（OnReady/ACTIVATED）→ 服务。
func Run(root plugin.Root) error {
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
	if len(reg.HTTP) > 0 || len(reg.Schedules) > 0 {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "本版 SDK 尚不支持 HTTP 路由与定时任务"}
	}

	// 3. 连接 Core 并依次完成握手、准备与激活，然后进入服务循环。
	l, err := readLaunch()
	if err != nil {
		return err
	}
	conn, err := dial(l)
	if err != nil {
		return err
	}
	defer conn.Close()
	ack, err := clientHandshake(conn, l)
	if err != nil {
		return err
	}
	if _, err := prepare(conn, l, ack, root, reg); err != nil {
		return err
	}
	sess, a, err := activate(conn, l, ack, root)
	if err != nil {
		return err
	}
	return serve(sess, a) // 收到 STOPPING 正常返回；断链时执行 OnStop 后返回错误，随后进程退出
}

// Registration 汇总从根插件收集到的静态注册项。
type Registration struct {
	Commands  []Command
	HTTP      []HTTPRoute
	Schedules []Schedule
}

// Collect 调用根插件的注册入口，检查 Handler 签名与同插件内的重复项（D-20：同插件重复即报错）。
func Collect(root plugin.Root) (*Registration, error) {
	if root == nil {
		return nil, errors.New("根插件为 nil")
	}
	reg := &Registration{}
	var errs []error
	if p, ok := root.(CommandProvider); ok {
		r := &CommandRegistry{}
		p.Commands(r)
		reg.Commands, errs = r.commands, append(errs, r.errs...)
	}
	if p, ok := root.(HTTPProvider); ok {
		r := &HTTPRegistry{}
		p.HTTP(r)
		reg.HTTP, errs = r.routes, append(errs, r.errs...)
	}
	if p, ok := root.(ScheduleProvider); ok {
		r := &ScheduleRegistry{}
		p.Schedules(r)
		reg.Schedules, errs = r.jobs, append(errs, r.errs...)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("%T 的注册无效: %w", root, err)
	}
	return reg, nil
}
