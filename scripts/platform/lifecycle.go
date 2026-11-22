package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"
)

const lifecycleCycles = 100

// treeState 由临时合作型插件通过受认证 HTTP 返回，父子都保持在 Core 管理的树内。
type treeState struct {
	ParentPID int    `json:"parent_pid"`
	ChildPID  int    `json:"child_pid"`
	Address   string `json:"address"`
}

type processObservation struct {
	PID    int  `json:"pid"` // Unix 负值表示检查整个进程组。
	Exists bool `json:"exists"`
}

type lockObservation struct {
	Path string `json:"path"`
	Held bool   `json:"held"`
}

type portObservation struct {
	Address string `json:"address"`
	Open    bool   `json:"open"`
	Error   string `json:"error,omitempty"`
}

type resources struct {
	Observed  time.Time            `json:"observed"`
	Processes []processObservation `json:"processes"`
	Locks     []lockObservation    `json:"locks"`
	Ports     []portObservation    `json:"ports"`
}

func (r *report) lifecycle(ctx context.Context, opt options, session *adminSession, ids []string) error {
	// 1. 首次业务路径与后续循环使用同一安装数据，逐次比对持久 Active。
	baseline := make(map[string]string)
	session.core.run.Purpose = "business"
	if err := session.ready(ctx, ids, baseline); err != nil {
		return err
	}
	if err := r.stopAndRelease(ctx, session.core, ids); err != nil {
		return err
	}
	for cycle := 1; cycle <= lifecycleCycles; cycle++ {
		if err := r.lifecycleRun(ctx, opt, session, ids, baseline, "cycle", cycle, false); err != nil {
			return fmt.Errorf("lifecycle cycle %d: %w", cycle, err)
		}
	}
	// 2. 强杀后的所有旧资源确认释放后，才允许恢复；记录信号原错作为预期故障证据。
	if err := r.lifecycleRun(ctx, opt, session, ids, baseline, "core_crash", 0, true); err != nil {
		return err
	}
	return r.lifecycleRun(ctx, opt, session, ids, baseline, "crash_recovery", 0, false)
}

func (r *report) lifecycleRun(ctx context.Context, opt options, original *adminSession, ids []string, baseline map[string]string, purpose string, cycle int, crash bool) (err error) {
	core, startErr := r.startCore(ctx, opt)
	if core == nil {
		return startErr
	}
	finished := false
	defer func() {
		if !finished { // 本次启动的清理只由本函数拥有，预期强杀结果在故障步骤单独核对。
			err = errors.Join(err, core.stop())
		}
	}()
	core.run.Purpose, core.run.Cycle = purpose, cycle
	if startErr != nil {
		return startErr
	}
	session := *original
	session.core = core
	if err := session.ready(ctx, ids, baseline); err != nil {
		return err
	}
	if !crash {
		err = r.stopAndRelease(ctx, core, ids)
		finished = true
		return err
	}
	if err := r.step("core_force_kill", func(step *stepResult) error {
		if err := core.command.Process.Kill(); err != nil {
			return err
		}
		<-core.done
		finished = true
		step.Output, step.ExitCode = core.run.ExitError, core.run.ExitCode
		var exit *exec.ExitError
		if !errors.As(core.err, &exit) || !killedState(exit.ProcessState) {
			return errors.Join(errors.New("Core did not exit from the requested force kill"), core.err)
		}
		return core.snapshot()
	}); err != nil {
		return err
	}
	return r.resourceCheck(ctx, core, ids, false)
}

func (s *adminSession) ready(ctx context.Context, ids []string, baseline map[string]string) error {
	if err := s.report.step("components_ready", func(_ *stepResult) error {
		return observe(ctx, s.core, func() (bool, error) {
			if err := s.core.snapshot(); err != nil {
				return false, err
			}
			return len(s.core.run.Activations) == len(ids), nil
		})
	}); err != nil {
		return err
	}
	// 重新读取正式管理状态，证明进程已激活且恢复的是同一 Active。
	raw, err := s.request(ctx, "persistent_active", http.MethodGet, "/api/v1/packages", nil, http.StatusOK, false)
	if err != nil {
		return err
	}
	if err := s.report.step("active_unchanged", func(_ *stepResult) error {
		var installed []struct {
			ID, Active, State string
			Enabled           bool
		}
		if err := json.Unmarshal(raw, &installed); err != nil {
			return err
		}
		current := make(map[string]string)
		for _, id := range ids {
			for _, item := range installed {
				if item.ID == id && item.Enabled && item.State == "Activated" && item.Active != "" {
					current[id] = item.Active
				}
			}
			if current[id] == "" {
				return fmt.Errorf("component %s is not activated: %s", id, raw)
			}
			found := false
			for _, activated := range s.core.run.Activations {
				if activated.ID == id && activated.Generation == current[id] && activated.PID > 0 {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("component %s has no matching activation PID: %s", id, raw)
			}
		}
		if len(baseline) == 0 {
			for id, active := range current {
				baseline[id] = active
			}
		} else if !maps.Equal(current, baseline) {
			return fmt.Errorf("Active changed: got %v, want %v", current, baseline)
		}
		return nil
	}); err != nil {
		return err
	}
	raw, err = s.request(ctx, "cooperative_tree", http.MethodGet, "/plugins/"+treePackageID+"/state", nil, http.StatusOK, false)
	if err != nil {
		return err
	}
	if err := s.report.step("tree_identity", func(_ *stepResult) error {
		var tree treeState
		if err := json.Unmarshal(raw, &tree); err != nil {
			return err
		}
		s.core.run.Tree = &tree
		if tree.ChildPID <= 0 || tree.ChildPID == tree.ParentPID || tree.Address == "" {
			return fmt.Errorf("invalid cooperative tree: %s", raw)
		}
		for _, activated := range s.core.run.Activations {
			if activated.ID == treePackageID && activated.PID == tree.ParentPID {
				return nil
			}
		}
		return fmt.Errorf("tree parent is absent from activation log: %s", raw)
	}); err != nil {
		return err
	}
	return s.report.resourceCheck(ctx, s.core, ids, true)
}

func (r *report) stopAndRelease(ctx context.Context, core *coreProcess, ids []string) error {
	if err := r.step("core_stop", func(_ *stepResult) error { return core.stop() }); err != nil {
		return err
	}
	return r.resourceCheck(ctx, core, ids, false)
}

func (r *report) resourceCheck(ctx context.Context, core *coreProcess, ids []string, running bool) error {
	name := "resources_released"
	if running {
		name = "resources_running"
	}
	return r.step(name, func(_ *stepResult) error {
		return observe(ctx, nil, func() (bool, error) {
			state := &resources{Observed: time.Now().UTC()}
			if running {
				core.run.Running = state
			} else {
				core.run.Released = state
			}
			match := true
			// 1. PID 与进程组均按 OS 实际状态判断，僵尸也必须消失。
			pids := []int{core.run.PID, core.run.Tree.ChildPID}
			for _, activated := range core.run.Activations {
				pids = append(pids, activated.PID)
				pids = append(pids, processGroup(activated.PID)...)
			}
			for _, pid := range pids {
				exists, err := processExists(pid)
				if err != nil {
					return false, fmt.Errorf("inspect PID/group %d: %w", pid, err)
				}
				state.Processes = append(state.Processes, processObservation{PID: pid, Exists: exists})
				match = match && exists == running
			}
			// 2. 对既有锁做真实非阻塞取锁，检查后立即释放，保留原文件。
			paths := []string{filepath.Join(r.DataDir, "core.lock")}
			for _, id := range ids {
				paths = append(paths, filepath.Join(r.DataDir, "plugins", id, "runtime.lock"))
			}
			for _, path := range paths {
				held, err := lockHeld(path)
				if err != nil {
					return false, fmt.Errorf("probe lock %s: %w", path, err)
				}
				state.Locks = append(state.Locks, lockObservation{Path: path, Held: held})
				match = match && held == running
			}
			// 3. 运行时连接真实 Core/子进程端口，退出后必须能重新绑定。
			for _, address := range []string{core.run.Address, core.run.Tree.Address} {
				port := portObservation{Address: address}
				if running {
					connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
					if err != nil {
						return false, fmt.Errorf("connect running port %s: %w", address, err)
					}
					if err := connection.Close(); err != nil {
						return false, err
					}
					port.Open = true
				} else {
					listener, err := net.Listen("tcp", address)
					if err != nil {
						if !addressInUse(err) {
							return false, fmt.Errorf("bind released port %s: %w", address, err)
						}
						port.Open, port.Error = true, err.Error()
						match = false
					} else if err := listener.Close(); err != nil {
						return false, err
					}
				}
				state.Ports = append(state.Ports, port)
			}
			return match, nil
		})
	})
}
