package generate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	buildOOMScore   = "1000" // Linux oom_score_adj 的最高值，编译子进程继承此值。
	buildTerminated = "build process was terminated; it may have run out of memory"
	buildStalled    = "NOTE: No build output for a while. Large packages can take minutes to compile; if this happens while downloading dependencies, set GOPROXY to a reachable mirror (e.g. https://goproxy.cn,direct) or configure a network proxy in System Settings."
)

// Output 收集标准输出。受管命令优先承受 Linux OOM，并在 stdout 和 stderr 同时沉默时提示。
func (t Toolchain) Output(ctx context.Context, cmd *exec.Cmd, stderr *bytes.Buffer) ([]byte, error) {
	if len(t.Env) == 0 {
		return cmd.Output()
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var watch *buildOutput
	if t.Stderr != nil && t.StallInterval > 0 {
		watch = &buildOutput{last: time.Now(), log: t.Stderr}
		cmd.Stdout = watch.track(cmd.Stdout)
		cmd.Stderr = watch.track(cmd.Stderr)
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if runtime.GOOS == "linux" {
		name := fmt.Sprintf("/proc/%d/oom_score_adj", cmd.Process.Pid)
		if err := os.WriteFile(name, []byte(buildOOMScore), 0); err != nil {
			slog.Debug("could not adjust source build OOM score", "pid", cmd.Process.Pid, "error", err)
		}
	} else {
		slog.Debug("source build OOM score adjustment is Linux only", "os", runtime.GOOS)
	}
	stop := func() {}
	if watch != nil {
		stop = watch.warnEvery(t.StallInterval)
	}
	err := cmd.Wait()
	stop()
	if err == nil {
		return stdout.Bytes(), nil
	}
	if ctx.Err() != nil {
		return stdout.Bytes(), errors.Join(ctx.Err(), err)
	}
	if status, ok := cmd.ProcessState.Sys().(interface{ Signaled() bool }); ok && status.Signaled() || compilerKilled(stderr.String()) {
		if t.Stderr != nil {
			fmt.Fprintln(t.Stderr, "WARNING: Build process was terminated; it may have run out of memory. Use alonabot pack on another machine and upload the archive, or install from a Release.")
		}
		err = fmt.Errorf("%s: %w", buildTerminated, err)
	}
	return stdout.Bytes(), err
}

// compilerKilled 处理编译器先被杀、go 自身仍以普通错误码退出的场景。
func compilerKilled(stderr string) bool {
	for line := range strings.Lines(stderr) {
		if strings.HasSuffix(strings.TrimSpace(line), ": signal: killed") {
			return true
		}
	}
	return false
}

type buildOutput struct {
	mu   sync.Mutex // 串行更新活动时间，并防止提示插进 stderr 的单次写入。
	last time.Time
	log  io.Writer
}

type buildOutputWriter struct {
	watch *buildOutput
	dest  io.Writer
}

func (b *buildOutput) track(dest io.Writer) io.Writer { return buildOutputWriter{b, dest} }

func (w buildOutputWriter) Write(p []byte) (int, error) {
	w.watch.mu.Lock()
	defer w.watch.mu.Unlock()
	w.watch.last = time.Now()
	return w.dest.Write(p)
}

func (b *buildOutput) warnEvery(interval time.Duration) func() {
	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-done:
				return
			case <-timer.C:
				b.mu.Lock()
				next := interval - time.Since(b.last)
				if next <= 0 {
					fmt.Fprintln(b.log, buildStalled)
					b.last, next = time.Now(), interval
				}
				b.mu.Unlock()
				timer.Reset(next)
			}
		}
	}()
	// Wait 收完输出后，等提示停止写日志再交还给调用方。
	return func() { close(done); <-stopped }
}
