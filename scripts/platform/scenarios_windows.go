package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

func (r *report) runScenarios(ctx context.Context, opt options) error {
	fixtures, err := r.prepareScenarios(ctx, opt)
	if err != nil {
		return err
	}
	fixtures.Runner = filepath.Join(fixtures.SourceRoot, "scripts", "platform", "windows_sources.py")
	fixtures.Executor = "native Go test executable"
	fixtures.RunFlags = fixtures.BuildFlags
	packages, fault := filepath.Join(r.Workspace, "scenario-build"), filepath.Join(r.Workspace, "scenario-build-recovery")
	// 1. 构建独立原生 .exe；输入源码与生成的有效声明均保留在各场景证据目录。
	for _, name := range []string{"build", "build-recovery", "management", "recovery", "survivor"} {
		if err := ctx.Err(); err != nil {
			return err
		}
		directory := filepath.Join(r.Workspace, "scenario-"+name)
		evidence := filepath.Join(directory, "scenario-evidence")
		if err := os.MkdirAll(evidence, privateDirectoryMode); err != nil {
			return err
		}
		inputs := filepath.Join(evidence, "source-inputs.json")
		command := exec.Command("uv", "run", "--no-project", "--offline", "--no-python-downloads", fixtures.Runner, "--scenarios", opt.scenarios, "--mode", name, "--out", inputs)
		command.Dir = r.Directory
		if err := r.nativeCommand(ctx, "scenario_"+name+"_sources", command); err != nil {
			return err
		}
		root, selection, tests := "TestT09ManagementFlow", "^TestT09ManagementFlow$", []string{"TestT09ManagementFlow"}
		if name == "build-recovery" || name == "recovery" {
			root, selection, tests = "TestScenarioRecovery", "^TestScenarioRecovery$", []string{"TestScenarioRecovery"}
			if name == "recovery" {
				for _, phase := range []string{"prepared", "committed", "first-prepared", "ready", "gates-prepared", "gates-committed", "gates-binding"} {
					tests = append(tests, "TestScenarioRecovery/"+phase)
				}
			}
		} else if name == "survivor" {
			root, selection, tests = "TestRecoveryEndToEnd", "^TestRecoveryEndToEnd$/^started$", []string{"TestRecoveryEndToEnd", "TestRecoveryEndToEnd/started"}
		}
		var sources map[string]string
		raw, err := os.ReadFile(inputs)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &sources); err != nil {
			return err
		}
		source, err := nativeHarness(sources, []string{"TestMain", root})
		if err != nil {
			return err
		}
		generated := filepath.Join(evidence, "platform_native_test.go")
		if err := os.WriteFile(generated, source, privateFileMode); err != nil {
			return err
		}
		overlay := filepath.Join(evidence, "overlay.json")
		if err := writeScenarioJSON(overlay, map[string]any{"Replace": map[string]string{filepath.Join(fixtures.SourceRoot, "tests", "e2e", "platform_native_test.go"): generated}}); err != nil {
			return err
		}
		executable := filepath.Join(evidence, "scenario-test.exe")
		command = exec.Command("go", "test", "-overlay", overlay, "-c", "-o", executable, "./e2e")
		command.Dir = filepath.Join(fixtures.SourceRoot, "tests")
		if err := r.nativeCommand(ctx, "scenario_"+name+"_compile", command); err != nil {
			return err
		}
		// 2. Windows Event、Job 句柄和运行锁由原生场景拥有；取消时等场景清理后返回。
		command = exec.Command(executable, "-test.v", "-test.count=1", "-test.run="+selection)
		command.Dir = filepath.Join(fixtures.SourceRoot, "tests", "e2e")
		env := []string{"ALONABOT_SCENARIO_MODE=" + name, "ALONABOT_SCENARIO_OUT=" + directory, "ALONABOT_SCENARIO_DATA_DIR=" + directory, "ALONABOT_SCENARIO_EVIDENCE=" + evidence,
			"ALONABOT_PLATFORM_SOURCE=" + fixtures.SourceRoot, "ALONABOT_PLATFORM_CORE=" + opt.core,
			"ALONABOT_SCENARIO_OVERLAY_CORE=" + filepath.Join(fault, "fault-core.exe"),
			"ALONABOT_SCENARIO_RECOVERY_PACKAGE=" + filepath.Join(fault, "recovery", "package.tar.gz"),
			"ALONABOT_SCENARIO_PROBE_PACKAGE=" + filepath.Join(fault, "gateprobe", "package.tar.gz"),
			"ALONABOT_SCENARIO_ADAPTER_PACKAGE=" + filepath.Join(fault, "adapter-package", "package.tar.gz")}
		for index, version := range []string{"1.0.0", "2.0.0", "3.0.0"} {
			env = append(env, "ALONABOT_SCENARIO_PACKAGE_"+strconv.Itoa(index+1)+"="+filepath.Join(packages, version, "package.tar.gz"))
		}
		building := name == "build" || name == "build-recovery"
		if !building {
			env = append(env, "ALONABOT_PLATFORM_CWD="+r.Directory)
		}
		command.Env = append(os.Environ(), env...)
		if err := writeScenarioJSON(filepath.Join(evidence, "inputs.json"), map[string]any{"scenario": name, "binary": opt.core, "overlay_core": filepath.Join(fault, "fault-core.exe"), "command": command.Args, "environment": env, "directory": directory}); err != nil {
			return err
		}
		var run *scenarioRun
		if building {
			fixtures.BuildDirectories = append(fixtures.BuildDirectories, directory)
		} else {
			run = &scenarioRun{Name: name, Directory: directory}
			fixtures.Runs = append(fixtures.Runs, run)
		}
		commandErr := r.nativeCommand(ctx, "scenario_"+name, command)
		logErr := os.WriteFile(filepath.Join(evidence, "go-test.log"), []byte(r.Steps[len(r.Steps)-1].Output), privateFileMode)
		var evidenceErr error
		if run != nil {
			evidenceErr = r.collectScenario(run, tests)
		}
		if err := errors.Join(commandErr, logErr, evidenceErr); err != nil {
			return err
		}
		if name == "build-recovery" {
			if err := r.step("scenario_binary_targets", func(*stepResult) error {
				paths := []string{filepath.Join(fault, "fault-core.exe"), filepath.Join(fault, "observed-plugin.exe"), filepath.Join(fault, "probe.exe"), filepath.Join(fault, "adapter.exe")}
				for _, version := range []string{"1.0.0", "2.0.0", "3.0.0"} {
					paths = append(paths, filepath.Join(packages, version, "plugin.exe"))
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
		}
	}
	return nil
}

func (r *report) nativeCommand(ctx context.Context, name string, command *exec.Cmd) error {
	if err := ctx.Err(); err != nil {
		return r.step(name, func(*stepResult) error { return err })
	}
	if command.Env == nil {
		command.Env = os.Environ()
	}
	command.Env = append(command.Env, "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
	prepareCore(command)
	return errors.Join(r.execute(name, command), ctx.Err())
}

func writeScenarioJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), privateFileMode)
}
