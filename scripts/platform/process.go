package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// 观察与停止期限沿用 cmd/alonabot/e2e_test.go 的真实进程验收期限。
	observationDeadline = 20 * time.Second
	stopDeadline        = 15 * time.Second
	observationInterval = 20 * time.Millisecond
)

type activation struct {
	ID         string `json:"id"`
	Generation string `json:"generation"`
	PID        int    `json:"pid"`
}

type coreRun struct {
	Purpose     string       `json:"purpose"`
	Cycle       int          `json:"cycle,omitempty"`
	Command     []string     `json:"command"`
	Directory   string       `json:"directory"`
	Log         string       `json:"log"`
	PID         int          `json:"pid"`
	Address     string       `json:"address"`
	Activations []activation `json:"activations"`
	ExitCode    *int         `json:"exit_code,omitempty"`
	ExitError   string       `json:"exit_error,omitempty"`
	Tree        *treeState   `json:"tree,omitempty"`
	Running     *resources   `json:"running,omitempty"`
	Released    *resources   `json:"released,omitempty"`
}

type coreProcess struct {
	command *exec.Cmd
	run     *coreRun
	done    chan struct{} // Wait 的唯一所有者关闭后，等待方才读取退出状态。
	err     error
}

func (r *report) startCore(ctx context.Context, opt options) (*coreProcess, error) {
	var process *coreProcess
	err := r.step("core_start", func(step *stepResult) error {
		run := &coreRun{
			Command:   []string{opt.core, "run", "--data-dir", r.DataDir, "--listen", "127.0.0.1:0", "--log-format", "json", "--console=false"},
			Directory: r.Directory, Log: filepath.Join(r.Workspace, fmt.Sprintf("core-%d.jsonl", len(r.Runs))),
		}
		r.Runs = append(r.Runs, run)
		r.currentRun = run
		step.Command, step.Directory = run.Command, run.Directory
		file, err := os.OpenFile(run.Log, os.O_WRONLY|os.O_CREATE|os.O_EXCL, privateFileMode)
		if err != nil {
			return err
		}
		command := exec.Command(run.Command[0], run.Command[1:]...)
		command.Dir, command.Stdout, command.Stderr = run.Directory, file, file
		prepareCore(command)
		if err := command.Start(); err != nil {
			return errors.Join(err, file.Close())
		}
		run.PID = command.Process.Pid
		process = &coreProcess{command: command, run: run, done: make(chan struct{})}
		go func() {
			// Wait 与日志文件都归此 goroutine；close 将最终错误和退出码发布给等待方。
			process.err = errors.Join(command.Wait(), file.Close())
			if process.err != nil {
				run.ExitError = process.err.Error()
			}
			code := command.ProcessState.ExitCode()
			run.ExitCode = &code
			close(process.done)
		}()
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = r.step("core_ready", func(_ *stepResult) error {
		return observe(ctx, process, func() (bool, error) {
			if err := process.snapshot(); err != nil {
				return false, err
			}
			return process.run.Address != "", nil
		})
	})
	return process, err
}

func (c *coreProcess) snapshot() error {
	raw, err := os.ReadFile(c.run.Log)
	if err != nil {
		return err
	}
	var activations []activation
	// 末尾尚未写完的一行留到下一次读取，完整行的解码失败直接报告。
	lines := bytes.Split(raw, []byte("\n"))
	for _, line := range lines[:len(lines)-1] {
		if len(line) == 0 {
			continue
		}
		var event struct {
			Message    string `json:"msg"`
			Listen     string `json:"listen"`
			PluginID   string `json:"plugin_id"`
			Generation string `json:"generation_id"`
			PID        int    `json:"pid"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			return fmt.Errorf("decode Core log %s: %w", c.run.Log, err)
		}
		switch event.Message {
		case "HTTP listener started":
			c.run.Address = event.Listen
		case "plugin activated":
			activations = append(activations, activation{ID: event.PluginID, Generation: event.Generation, PID: event.PID})
		}
	}
	c.run.Activations = activations
	return nil
}

func observe(ctx context.Context, core *coreProcess, condition func() (bool, error)) error {
	ctx, cancel := context.WithTimeout(ctx, observationDeadline)
	defer cancel()
	ticker := time.NewTicker(observationInterval)
	defer ticker.Stop()
	var exited <-chan struct{}
	if core != nil {
		exited = core.done
	}
	for {
		ready, err := condition()
		if err != nil || ready {
			return err
		}
		select { // 按实际观测推进；Core 提前退出或观察期限到达立即结束。
		case <-ticker.C:
		case <-exited:
			return errors.Join(fmt.Errorf("Core exited before observation completed; log: %s", core.run.Log), core.err)
		case <-ctx.Done():
			return fmt.Errorf("observation deadline: %w", ctx.Err())
		}
	}
}

func (c *coreProcess) stop() error {
	select { // 已退出的进程直接交回原始退出结果。
	case <-c.done:
		return c.err
	default:
	}
	if err := interruptCore(c.command.Process); err != nil {
		killErr := c.command.Process.Kill()
		<-c.done
		return errors.Join(fmt.Errorf("interrupt Core: %w", err), killErr, c.err)
	}
	timer := time.NewTimer(stopDeadline)
	defer timer.Stop()
	select {
	case <-c.done:
		return errors.Join(c.err, c.snapshot())
	case <-timer.C:
		killErr := c.command.Process.Kill()
		<-c.done
		return errors.Join(fmt.Errorf("Core %d exceeded stop deadline; log: %s", c.run.PID, c.run.Log), killErr, c.err)
	}
}
