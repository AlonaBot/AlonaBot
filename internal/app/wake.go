package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

func (a *assembly) acquireActivity(c router.Caller) (func(), error) {
	if _, ok := a.router.PluginFor(c); !ok {
		return nil, &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "plugin Session is no longer bound"}}
	}
	if session, ok := a.wakeSessions.Load(c); ok {
		return a.waker.AcquireActivity(c.PluginID, session.(*wire.Session))
	}
	return func() {}, nil // 常驻开发组件、Adapter 与 WebUI 保持原生命周期。
}

func (a *assembly) ensureActive(ctx context.Context, id string, n int64) (func(), error) {
	if caller, ok := a.router.Caller(id); ok {
		if _, tracked := a.wakeSessions.Load(caller); !tracked {
			return a.acquireActivity(caller)
		}
	}
	return a.waker.EnsureActive(ctx, id, n)
}

// startWake 复用 Core 运行循环和原代次启动；共享启动的期限只受 Core 生命周期与原阶段预算约束。
func (a *assembly) startWake(_ context.Context, id string) (*supervisor.Running, error) {
	c := a.core
	var process *supervisor.Running
	err := c.operation(c.lifetime, func(ctx context.Context) (err error) {
		// 状态持久化留在 Core 串行生命周期内，关闭状态库前交还清理权。
		defer func() {
			if err == nil || ctx.Err() != nil {
				return
			}
			status, readErr := c.store.RuntimeStatus(ctx, id)
			if readErr == nil && (status.State == state.RuntimeSuspended || status.State == state.RuntimeWaking) {
				in, installErr := c.store.GetInstall(ctx, id)
				readErr = installErr
				if installErr == nil {
					readErr = c.recordRuntime(ctx, in, state.RuntimeFailed, err.Error())
				}
			}
			err = errors.Join(err, readErr)
			// 在运行循环内撤销失败的休眠计划，完整启动错误仍交给 Waker 发布。
			a.scheduler.Detach(id)
			slog.Error(fmt.Sprintf("Failed to wake component %s: %v", id, err), "plugin_id", id, "error", err)
		}()
		in, err := c.store.GetInstall(ctx, id)
		if err != nil {
			return err
		}
		enabled, err := c.store.Enabled(ctx, id)
		if err != nil {
			return err
		}
		if !enabled || in.Uninstalled {
			return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "plugin is disabled or uninstalled"}}
		}
		// 管理启动可能已先完成；借用它的完整绑定，保持本次请求只交付一次。
		for _, running := range c.components {
			if running != nil && running.id == id && running.instance.State() == host.StateActivated {
				process = running.process
				return nil
			}
		}
		status, err := c.store.RuntimeStatus(ctx, id)
		if err != nil {
			return err
		}
		if status.State != state.RuntimeSuspended {
			return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "plugin is no longer suspended"}}
		}
		// SUSPENDED 已确认旧树退出；直接恢复，保留本次正常休眠接纳的到期项。
		for i, old := range c.components {
			if old != nil && old.id == id {
				c.components[i] = nil
			}
		}
		running, err := c.startInstalled(ctx, c.dir, in)
		if running != nil {
			c.components = append(c.components, running)
			process = running.process
		}
		if err != nil && running != nil {
			return errors.Join(err, c.stopComponent(running, time.Now(), host.NewStopTiming(c.values), pb.StopReason_STOP_REASON_DISABLE))
		}
		return err
	})
	return process, err
}
