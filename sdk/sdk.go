// Package sdk 是 AlonaBot Native Go 插件的公共入口：Context、注册器与生命周期接口。
//
// T-01 只固定可编译的公共类型与注册收集（Collect）；与 Core 通信的运行行为（Run、Send）
// 从 T-02 起实现，在此之前返回 UNSUPPORTED 错误，不 panic，也不假装成功。
package sdk

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// Error 是带协议错误码的 SDK 错误。
type Error struct {
	Code    pb.ErrorCode
	Message string
}

func (e *Error) Error() string { return e.Code.String() + ": " + e.Message }

// ErrNotImplemented 表示该入口的运行行为尚未实现。
var ErrNotImplemented = &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "SDK 运行时尚未实现"}

// 可选的进程生命周期 Hook（CONTRACTS §4.1）；没有实现时无需空方法。
type (
	Starter interface{ OnStart(context.Context) error }
	Readier interface{ OnReady(context.Context) error }
	Stopper interface{ OnStop(context.Context) error }
)

// 根插件可选实现的注册入口；实例方法通过这些入口显式绑定。
type (
	CommandProvider  interface{ Commands(*CommandRegistry) }
	HTTPProvider     interface{ HTTP(*HTTPRegistry) }
	ScheduleProvider interface{ Schedules(*ScheduleRegistry) }
)

// Context 只在一次 Handler 调用期间有效。
type Context struct {
	ctx     context.Context
	stopped bool
}

// Context 返回本次调用的 context（含 deadline 与取消）。
func (c *Context) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// StopPropagation 锁存停止传播标记；一经设置不能重新打开传播，并且不受业务错误影响。
func (c *Context) StopPropagation() { c.stopped = true }

// SendOption 调整一次 Send。
type SendOption func(*sendOptions)

type sendOptions struct{ to *pb.Target }

// To 指定显式目标；没有真实入站默认目标时必须使用。
func To(t *pb.Target) SendOption { return func(o *sendOptions) { o.to = t } }

// Send 发起一次逻辑发送。出错时仍可能返回已确认的部分结果（MESSAGE-MODEL §4）。
func (c *Context) Send(chain message.Chain, opts ...SendOption) (*pb.SendResult, error) {
	return nil, ErrNotImplemented
}

// Run 是插件 main 的入口：连接 Core、完成握手并处理请求，直到停止。
// 顺序：收集注册项 → 取得 runtime.lock（任何用户回调之前）→ 握手 → 准备阶段 → 激活（OnReady/ACTIVATED）→ 服务。
func Run(root plugin.Root) error {
	reg, err := Collect(root)
	if err != nil {
		return err
	}
	if len(reg.HTTP) > 0 || len(reg.Schedules) > 0 {
		// HTTP 需要真实 Listener、调度需要 Core Scheduler（T-07）；不静默丢弃这些声明。
		return &Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "本版 SDK 尚不支持 HTTP 路由与定时任务（T-07）"}
	}
	lock, err := acquireRuntimeLock()
	if err != nil {
		return err
	}
	defer lock.Unlock()
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

// Registration 是从根插件收集到的静态注册项。
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
