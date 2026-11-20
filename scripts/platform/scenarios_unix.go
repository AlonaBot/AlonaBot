//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func (r *report) runScenarios(ctx context.Context, opt options) error {
	// 1. 构建带检查点的临时 Core 前，核对其产品源码与输入候选的提交相同。
	fixtures, err := r.prepareScenarios(ctx, opt)
	if err != nil {
		return err
	}
	fixtures.Executor = filepath.Join(r.Workspace, "scenario-exec")
	if err := r.step("scenario_executor", func(step *stepResult) error {
		script := "#!/bin/sh\ncd '" + strings.ReplaceAll(r.Directory, "'", "'\\''") + "' || exit\nprintf 'Platform test cwd: %s\\n' \"$PWD\"\nexec \"$@\"\n"
		step.Output = script
		return os.WriteFile(fixtures.Executor, []byte(script), privateDirectoryMode)
	}); err != nil {
		return err
	}
	// 2. build 与 build-recovery 复用原生成器；构建产物和全部 overlay 源码均留在工作区。
	for _, name := range []string{"build", "build-recovery"} {
		directory := filepath.Join(r.Workspace, "scenario-"+name)
		fixtures.BuildDirectories = append(fixtures.BuildDirectories, directory)
		if err := r.scenarioCommand(ctx, "scenario_"+name, fixtures.BuildFlags, fixtures.Runner, name, "--out", directory); err != nil {
			return err
		}
	}
	packages, fault := fixtures.BuildDirectories[0], fixtures.BuildDirectories[1]
	if err := r.step("scenario_binary_targets", func(_ *stepResult) error {
		paths := []string{filepath.Join(fault, "fault-core"), filepath.Join(fault, "observed-plugin"), filepath.Join(fault, "probe"), filepath.Join(fault, "adapter")}
		for _, version := range []string{"1.0.0", "2.0.0", "3.0.0"} {
			paths = append(paths, filepath.Join(packages, version, "plugin"))
		}
		for _, path := range paths {
			build, err := executableCandidate(path)
			fixtures.Binaries = append(fixtures.Binaries, build)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// 3. -exec 仅改变测试进程 cwd；原管理、七窗口和 started 的全部真实断言保持。
	// 通过 PATH 查找固定文件名，避免临时目录中的空格进入 GOFLAGS 与 -exec 的两层解析。
	fixtures.RunFlags = strings.TrimSpace(fixtures.BuildFlags + " -exec=scenario-exec")
	for _, name := range []string{"management", "recovery", "survivor"} {
		directory := filepath.Join(r.Workspace, "scenario-"+name)
		if err := os.Mkdir(directory, privateDirectoryMode); err != nil {
			return err
		}
		run := &scenarioRun{Name: name, Directory: directory}
		fixtures.Runs = append(fixtures.Runs, run)
		args := []string{name, "--alonabot", opt.core, "--data-dir", directory}
		tests := []string{"TestT09ManagementFlow"}
		if name == "management" {
			for _, version := range []string{"1.0.0", "2.0.0", "3.0.0"} {
				args = append(args, filepath.Join(packages, version, "package.tar.gz"))
			}
		} else {
			args = append(args, "--overlay-core", filepath.Join(fault, "fault-core"), "--recovery-package", filepath.Join(fault, "recovery", "package.tar.gz"))
			if name == "recovery" {
				args = append(args, "--probe-package", filepath.Join(fault, "gateprobe", "package.tar.gz"), "--adapter-package", filepath.Join(fault, "adapter-package", "package.tar.gz"))
				tests = []string{"TestScenarioRecovery"}
				for _, phase := range []string{"prepared", "committed", "first-prepared", "ready", "gates-prepared", "gates-committed", "gates-binding"} {
					tests = append(tests, "TestScenarioRecovery/"+phase)
				}
			} else {
				tests = []string{"TestRecoveryEndToEnd", "TestRecoveryEndToEnd/started"}
			}
		}
		commandErr := r.scenarioCommand(ctx, "scenario_"+name, fixtures.RunFlags, fixtures.Runner, args...)
		evidenceErr := r.collectScenario(run, tests)
		if err := errors.Join(commandErr, evidenceErr); err != nil {
			return err
		}
	}
	return nil
}

func (r *report) scenarioCommand(ctx context.Context, name, flags, runner string, args ...string) error {
	if err := ctx.Err(); err != nil {
		return r.step(name, func(*stepResult) error { return err })
	}
	argv := append([]string{"run", "--no-project", "--offline", "--no-python-downloads", runner}, args...)
	command := exec.Command("uv", argv...)
	command.Dir = r.Directory
	command.Env = append(os.Environ(), "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH, "GOFLAGS="+flags,
		"PATH="+r.Workspace+string(os.PathListSeparator)+os.Getenv("PATH"))
	// 独立进程组隔离终端中断；已开始的场景自行完成清理后才响应取消，避免遗留持锁写入者。
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return errors.Join(r.execute(name, command), ctx.Err())
}
