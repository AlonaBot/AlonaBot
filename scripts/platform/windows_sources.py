"""生成原生 Windows 场景源码，复用管理与恢复断言并替换 Unix 资源入口。"""

import argparse
import importlib.util
import json
from pathlib import Path
import sys


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--scenarios", type=Path, required=True)
    parser.add_argument("--mode", choices=("build", "build-recovery", "management", "recovery", "survivor"), required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    sys.dont_write_bytecode = True
    spec = importlib.util.spec_from_file_location("scenario_sources", args.scenarios / "run.py")
    scenario = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(scenario)
    root = args.scenarios.resolve().parents[1]
    read = lambda name: (root / "cmd/alonabot" / name).read_text()
    sources = {name: read(name) for name in ("e2e_test.go", "manage_e2e_test.go", "recover_e2e_test.go")}
    if args.mode == "build":
        sources["flow.go"] = scenario.fixture_source(read("t09_e2e_test.go"))
    elif args.mode == "management":
        sources["flow.go"] = scenario.management_source(read("t09_e2e_test.go"))
    else:
        sources["flow.go"] = (args.scenarios / "recovery.go.tmpl").read_text() + scenario.EVIDENCE_HELPERS
    if args.mode == "survivor":
        source = scenario.survivor_source(sources["recover_e2e_test.go"])
        source = scenario.replace_once(source, '\t\t\tif err := c.cmd.Process.Kill(); err != nil {',
            '\t\t\tif phase == "started" { platformHoldJob(t, c) }\n\t\t\tif err := c.cmd.Process.Kill(); err != nil {')
        source = source.replace("syscall.Kill(pid, syscall.SIGKILL)", "platformKill(pid)")
        source = scenario.replace_once(source,
            'syscall.Mkfifo(filepath.Join(pluginDir, "hold-runtime"), 0o600)',
            'os.WriteFile(filepath.Join(pluginDir, "hold-runtime"), nil, 0o600)')
        sources["recover_e2e_test.go"] = source
    if args.mode in ("build-recovery", "recovery", "survivor"):
        source = sources["flow.go"]
        start = source.index('\tpatch("internal/core/journal/commit.go", " select {}",')
        end = source.index('\tpatch("internal/app/assembly.go",', start)
        source = source[:start] + '''\tpatch("internal/core/journal/commit.go", " select {}", `
 if _,err:=windows.WaitForSingleObject(event,windows.INFINITE);err!=nil{panic(err)}`)
 patch("internal/app/assembly.go", "import (", "import (\\n \\\"golang.org/x/sys/windows\\\"")
''' + source[end:]
        source = scenario.replace_once(source,
            'name:=os.Getenv("ALONABOT_TEST_CHECKPOINT");if err:=os.WriteFile(name,[]byte("binding"),0600);err!=nil{panic(err)}\n  if _,err:=os.ReadFile(name+".release");err!=nil{panic(err)}',
            '''eventName:="Local\\\\AlonaBotPlatformEvent-"+fmt.Sprint(os.Getpid())
  event,err:=windows.CreateEvent(nil,1,0,windows.StringToUTF16Ptr(eventName));if err!=nil{panic(err)};defer func(){if err:=windows.CloseHandle(event);err!=nil{panic(err)}}()
  name:=os.Getenv("ALONABOT_TEST_CHECKPOINT");if err:=os.WriteFile(name,[]byte(eventName),0600);err!=nil{panic(err)}
  if _,err:=windows.WaitForSingleObject(event,windows.INFINITE);err!=nil{panic(err)}''')
        source = source.replace('\tscenarioRecoveryReleasePipe(t, dir)\n', '')
        source = source.replace('Path: name, RawQuery: "mode=ro"', 'Path: "/"+filepath.ToSlash(name), RawQuery: "mode=ro"')
        sources["flow.go"] = source
    for name, source in sources.items():
        source = source.replace('filepath.Join(binDir, "alonabot")', 'os.Getenv("ALONABOT_PLATFORM_CORE")')
        source = source.replace('filepath.Join(binDir,"alonabot")', 'os.Getenv("ALONABOT_PLATFORM_CORE")')
        source = source.replace('if err := c.cmd.Start(); err != nil {', 'platformPrepare(c.cmd)\n\tif err := c.cmd.Start(); err != nil {')
        for executable in ("core", "plugin", "fault-core", "observed-plugin", "probe", "adapter"):
            source = source.replace(f'filepath.Join(dir, "{executable}")', f'filepath.Join(dir, "{executable}.exe")')
        sources[name] = source.replace("bin/component", "bin/component.exe")
    sources["native.go"] = Path(__file__).with_name("windows_native.go.tmpl").read_text()
    args.out.write_text(json.dumps(sources, indent=2) + "\n")
    print(f"Native source inputs: {args.out}")


if __name__ == "__main__":
    main()
