// Package sdk 提供 AlonaBot Native Go 插件的公共入口。
//
// 插件的 main 调用 Run，传入嵌入了 [Base] 的根插件。Run 取得 runtime.lock，
// 收集根插件的注册项，与 Core 完成握手、准备和激活，然后一直运行到 Core 要求停止。
// Run 把 log/slog 的默认 Logger 换成写 stderr 的 JSON 行，按 Core 下发的级别过滤（CONTRACTS §14）。
// Base 为 [Root] 的全部 Hook 提供空的默认实现，根插件只覆盖需要的方法。
//
// 业务分发、Send、HTTP 路由与定时任务尚未实现；它们返回 UNSUPPORTED 错误，不会 panic，也不会假装成功。
//
// 包的组成：
//   - 插件的 main 调用 [Run]；[Collect] 单独收集注册项，供检查与测试使用。
//   - [CommandRegistry]、[HTTPRegistry] 与 [ScheduleRegistry] 收集命令、HTTP 路由和定时任务。
//   - [Context] 只在一次 Handler 调用期间有效，提供 [Context.Send] 与 [Context.StopPropagation]。
//   - 根插件嵌入 [Base]，覆盖 [Root] 中需要的 Hook。
package sdk

import (
	"context"
	"errors"
	"fmt"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
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
	if len(reg.HTTP) > 0 || len(reg.Schedules) > 0 {
		return &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "this SDK version does not support HTTP routes or scheduled jobs yet"}
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
	setLogLevel(ack.GetBootstrap().GetLogLevel())
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
func Collect(root Root) (*Registration, error) {
	if root == nil {
		return nil, errors.New("root plugin is nil")
	}
	cmds, routes, jobs := &CommandRegistry{}, &HTTPRegistry{}, &ScheduleRegistry{}
	root.Commands(cmds)
	root.HTTP(routes)
	root.Schedules(jobs)
	reg := &Registration{Commands: cmds.commands, HTTP: routes.routes, Schedules: jobs.jobs}
	errs := slices.Concat(cmds.errs, routes.errs, jobs.errs)
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid registration of %T: %w", root, err)
	}
	return reg, nil
}
