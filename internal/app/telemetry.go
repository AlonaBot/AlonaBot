package app

import (
	"context"

	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/telemetry"
)

// telemetryTargets 构造当前已安装组件及开发组件的只读运行态快照，不进入 Core.run 操作队列。
func (c *Core) telemetryTargets(ctx context.Context) ([]telemetry.Target, error) {
	installs, err := c.store.Installs(ctx)
	if err != nil {
		return nil, err
	}
	targets := make([]telemetry.Target, 0, len(installs))
	seen := make(map[string]struct{}, len(installs))
	for _, in := range installs {
		if in.Uninstalled {
			continue
		}
		seen[in.PluginID] = struct{}{}
		enabled, err := c.store.Enabled(ctx, in.PluginID)
		if err != nil {
			return nil, err
		}
		active, err := c.store.Active(ctx, in.PluginID)
		if err != nil {
			return nil, err
		}
		current := "Stopped"
		if active == "" {
			current = "Installed"
		}
		inProcess := in.Role == state.RoleWebUI && in.Entrypoint == ""
		if enabled && inProcess {
			current = "Activated"
		} else if runningState := c.assembly.componentState(in.PluginID); runningState != "" {
			current = runningState
		} else if enabled && active != "" {
			current = "Degraded"
		}
		if in.Role == state.RolePlugin || in.Role == state.RoleAgent {
			status, err := c.store.RuntimeStatus(ctx, in.PluginID)
			if err != nil {
				return nil, err
			}
			if status.State == state.RuntimeSuspended {
				current = "Suspended"
			}
		}

		target := telemetry.Target{
			ID:        in.PluginID,
			Role:      string(in.Role),
			State:     current,
			InProcess: inProcess,
			Running:   enabled && inProcess,
		}
		if raw, ok := c.assembly.running.Load(in.PluginID); ok {
			rc := raw.(*runningComponent)
			alive := rc.instance != nil && rc.instance.State() != host.StateDegraded
			if alive {
				select {
				case <-rc.instance.Session.Done():
					alive = false
				default:
				}
			}
			if alive {
				target.Running = true
				if rc.process != nil && rc.process.Process != nil {
					target.InProcess = false
					target.Process = rc.process.Process
					target.StartedAt = rc.process.StartedAt
				} else {
					target.InProcess = true
					target.StartedAt = c.startedAt
				}
			}
		}
		targets = append(targets, target)
	}

	// 补充库模式/开发模式中未登记进 installs 表的活跃组件。
	c.assembly.running.Range(func(key, value any) bool {
		id := key.(string)
		if _, exists := seen[id]; exists {
			return true
		}
		rc := value.(*runningComponent)
		if rc.instance == nil || rc.instance.State() == host.StateDegraded {
			return true
		}
		select {
		case <-rc.instance.Session.Done():
			return true
		default:
		}
		target := telemetry.Target{
			ID:      id,
			Role:    string(rc.install.Role),
			State:   rc.instance.State().String(),
			Running: true,
		}
		if rc.process != nil && rc.process.Process != nil {
			target.Process = rc.process.Process
			target.StartedAt = rc.process.StartedAt
		} else {
			target.InProcess = true
			target.StartedAt = c.startedAt
		}
		targets = append(targets, target)
		return true
	})
	return targets, nil
}
