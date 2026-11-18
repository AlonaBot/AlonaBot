package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/idle"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

const idleCheckInterval = time.Second

func (c *Core) sleepPolicy(ctx context.Context, in state.Install, sleep settings.Sleep) (idle.Policy, state.RouteSnapshot, error) {
	directory, err := installedDirectory(c.dir, in)
	if err != nil {
		return idle.Policy{}, state.RouteSnapshot{}, err
	}
	parsed, err := installedManifest(in, directory)
	if err != nil {
		return idle.Policy{}, state.RouteSnapshot{}, err
	}
	snapshot, err := c.store.ActiveRouteSnapshot(ctx, in.PluginID)
	if err != nil {
		return idle.Policy{}, state.RouteSnapshot{}, err
	}
	return idle.Resolve(sleep, parsed.Manifest, hasInstalledRoutes(in), snapshot.Routes), snapshot, nil
}

// checkIdle 与管理操作共用运行循环；StopIdle 再在逐插件锁和活动门禁内复核。
func (c *Core) checkIdle(ctx context.Context, now time.Time) error {
	installs, err := c.store.Installs(ctx)
	if err != nil {
		return err
	}
	sleep := c.assembly.settings.Settings().Sleep
	var result error
	for _, in := range installs {
		if !hasInstalledRoutes(in) {
			continue
		}
		if err := c.idleComponent(ctx, in, sleep, now); err != nil {
			result = errors.Join(result, fmt.Errorf("idle plugin %s: %w", in.PluginID, err))
		}
	}
	return result
}

func (c *Core) idleComponent(ctx context.Context, in state.Install, sleep settings.Sleep, now time.Time) error {
	runtime, err := c.store.RuntimeStatus(ctx, in.PluginID)
	if err != nil {
		return err
	}
	activity := c.assembly.waker.Activity(in.PluginID)
	policy, snapshot := idle.Policy{}, state.RouteSnapshot{}
	// 常驻、停用和已休眠插件也消费启动记录，避免无人参与空闲停止时积累历史。
	if !in.Uninstalled && sleep.Mode != "always_on" && activity.Ready {
		enabled, err := c.store.Enabled(ctx, in.PluginID)
		if err != nil {
			return err
		}
		if enabled {
			policy, snapshot, err = c.sleepPolicy(ctx, in, sleep)
			if err != nil {
				return err
			}
		}
	}
	canStop := c.assembly.idle.CanStop(in.PluginID, policy, runtime, activity, now)
	c.assembly.waker.AcknowledgeColdStarts(in.PluginID, activity.ColdStartSequence)
	if !canStop {
		return nil
	}
	for i, running := range c.components {
		if running == nil || !running.activity || running.id != in.PluginID || running.instance.Session != activity.Session || snapshot.GenerationID != running.instance.GenerationID {
			continue
		}
		err := c.stopComponentWithActivity(running, time.Now(), host.NewStopTiming(c.values), pb.StopReason_STOP_REASON_IDLE, &activity.LastActivityAt)
		if running.instance.State() == host.StateStopped && c.assembly.waker.Activity(running.id).Session != running.instance.Session {
			c.components[i] = nil // 旧观察者只归还自己的进程，不再触碰新 Session。
		}
		return err
	}
	return nil
}

// applySleep 在运行循环中读取生效策略和运行状态，在循环外逐个等待原代次唤醒。
// 保存成功后的部分唤醒失败一并返回，已保存设置继续生效。
func (c *Core) applySleep(ctx context.Context) error {
	var wake []string
	err := c.operation(ctx, func(ctx context.Context) error {
		installs, err := c.store.Installs(ctx)
		if err != nil {
			return err
		}
		sleep := c.assembly.settings.Settings().Sleep
		var result error
		for _, in := range installs {
			if in.Uninstalled || !hasInstalledRoutes(in) {
				continue
			}
			needed, err := c.sleepNeedsWake(ctx, in, sleep)
			if err != nil {
				result = errors.Join(result, fmt.Errorf("apply sleep for %s: %w", in.PluginID, err))
			} else if needed {
				wake = append(wake, in.PluginID)
			}
		}
		return result
	})
	for _, id := range wake {
		release, wakeErr := c.assembly.waker.EnsureActive(ctx, id, 0)
		if wakeErr != nil {
			err = errors.Join(err, fmt.Errorf("apply sleep for %s: %w", id, wakeErr))
			continue
		}
		release()
	}
	return err
}

func (c *Core) sleepNeedsWake(ctx context.Context, in state.Install, sleep settings.Sleep) (bool, error) {
	enabled, err := c.store.Enabled(ctx, in.PluginID)
	if err != nil || !enabled {
		return false, err
	}
	runtime, err := c.store.RuntimeStatus(ctx, in.PluginID)
	if err != nil {
		return false, err
	}
	if runtime.State != state.RuntimeSuspended && runtime.State != state.RuntimeQuiescing {
		return false, nil
	}
	policy, _, err := c.sleepPolicy(ctx, in, sleep)
	if err != nil {
		return false, errors.Join(err, c.recordRuntime(ctx, in, state.RuntimeFailed, err.Error()))
	}
	return !policy.Eligible, nil
}
