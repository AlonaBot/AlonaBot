"""以指定产物执行正式管理、崩溃恢复和残存写入者场景。"""

import argparse
import json
import os
from pathlib import Path
import platform
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[2]
TEST = ROOT / "tests/e2e/management_e2e_test.go"
HARNESS = ROOT / "tests/e2e/e2e_test.go"
RECOVERY = ROOT / "tests/e2e/recover_e2e_test.go"
RECOVERY_TEMPLATE = Path(__file__).with_name("recovery.go.tmpl")


# shortcut: 依赖集成测试的固定源码锚点，入口结构调整时同步修改场景生成器。
def replace_once(source, old, new):
    if source.count(old) != 1:
        raise ValueError(f"expected one source anchor: {old!r}")
    return source.replace(old, new, 1)


def fixture_source(source):
    # 包构建辅助保留原实现；只把输出从 t.TempDir 改为调用者指定的目录。
    function = source[source.index("func TestT09ManagementFlow(") :]
    function = function[: function.index("\tfirstPackage, secondPackage, thirdPackage :=")]
    function = replace_once(
        function,
        "\t\tdir := t.TempDir()",
        '\t\tdir := filepath.Join(os.Getenv("ALONABOT_SCENARIO_OUT"), version)\n'
        '\t\tif err := os.Mkdir(dir, 0o700); err != nil { t.Fatal(err) }',
    )
    return (
        '//go:build unix\n\npackage main\nimport ("fmt"; "os"; "os/exec"; '
        '"path/filepath"; "runtime"; "strings"; "testing")\n\n'
        + function
        + '\tfor _, version := range []string{"1.0.0", "2.0.0", "3.0.0"} {\n'
        + '\t\tt.Logf("built fixture %s", packageAt(version))\n\t}\n}\n'
    )


def management_source(source):
    # 保留正式流程的全部断言；包构建由独立 build 命令完成。
    start = source.index("\tconst schema =")
    end = source.index("\tfirstPackage, secondPackage, thirdPackage :=")
    source = source[:start] + source[end:]
    source = replace_once(
        source,
        "// 1. 复用恢复插件的 SQLite 读写，只给临时产物补充程序版本与配置的观测入口。",
        "// 1. 使用指定的测试包与数据目录，核对首次安装后的程序、配置与业务数据。",
    )
    source = replace_once(
        source,
        '\tfirstPackage, secondPackage, thirdPackage := packageAt("1.0.0"), packageAt("2.0.0"), packageAt("3.0.0")',
        '\tfirstPackage, secondPackage, thirdPackage := os.Getenv("ALONABOT_SCENARIO_PACKAGE_1"), '
        'os.Getenv("ALONABOT_SCENARIO_PACKAGE_2"), os.Getenv("ALONABOT_SCENARIO_PACKAGE_3")',
    )
    source = replace_once(source, "\tdir := t.TempDir()", '\tdir := os.Getenv("ALONABOT_SCENARIO_DATA_DIR")')
    for package in ("fmt", "os/exec", "runtime"):
        source = replace_once(source, f'\t"{package}"\n', "")
    source = replace_once(
        source,
        '\t"github.com/AlonaBot/AlonaBot/internal/core/buttons"',
        '\t"github.com/AlonaBot/AlonaBot/internal/core/manage"\n'
        '\t"github.com/AlonaBot/AlonaBot/internal/core/buttons"',
    )
    source = source.replace("startEmptyCore(t, ", "scenarioStartCore(t, ")
    source = source.replace("c.management(t, ", "scenarioManagement(t, c, ")
    source = source.replace("managementFailure(t, ", "scenarioFailure(t, ")
    source = source.replace("fixtureHTTP(t, ", "scenarioHTTP(t, ")
    source = source.replace("findPackage(t, ", "scenarioPackage(t, ")
    source = replace_once(source, "\texpected := map[string]struct", '\tscenarioRecord(t, "journals", records)\n\texpected := map[string]struct')
    return source + EVIDENCE_HELPERS


def survivor_source(source):
    source = replace_once(
        source, "binary, plugin := buildRecoveryFixture(t, fixture)",
        'binary, plugin := os.Getenv("ALONABOT_SCENARIO_OVERLAY_CORE"), "unused"',
    )
    source = replace_once(
        source, 'archive := recoveryPackage(t, fixture, plugin, "1.0.0")',
        'archive := os.Getenv("ALONABOT_SCENARIO_RECOVERY_PACKAGE")',
    )
    source = replace_once(
        source, "return startRecoveryCore(t, binary, dir, phase)",
        "return scenarioRecoveryStart(t, binary, dir, phase)",
    )
    source = replace_once(
        source, '\t\t\tdir := t.TempDir()',
        '\t\t\tdir := scenarioRecoveryDirectory(t, phase)',
    )
    source = replace_once(
        source, '\t\t\told := findPackage(t, c, dir, "recovery").Active',
        '\t\t\told := findPackage(t, c, dir, "recovery").Active\n'
        '\t\t\tscenarioRecoveryValue(t, c, dir, "41", true)',
    )
    source = replace_once(
        source, '\t\t\t\tif err := syscall.Kill(pid, syscall.SIGKILL); err != nil {',
        '\t\t\t\tscenarioRecord(t, "surviving-writer", map[string]any{"package":item,"pid":pid,"starts_before":string(before),"starts_after":string(after)})\n'
        '\t\t\t\tscenarioRecoveryDatabase(t, filepath.Join(pluginDir,"generations",old,"db.sqlite"), "41")\n'
        '\t\t\t\tif err := syscall.Kill(pid, syscall.SIGKILL); err != nil {',
    )
    source = replace_once(
        source, '\t\t\tstopTestCore(t, c)\n\t\t\tst = recoveryStore(t, dir)',
        '\t\t\tscenarioRecoveryValue(t, c, dir, "41", false)\n'
        '\t\t\tstopTestCore(t, c)\n\t\t\tst = recoveryStore(t, dir)',
    )
    source = replace_once(
        source, '\t\t\tcandidateDir := filepath.Join(',
        '\t\t\tscenarioRecord(t, "journal-recovered", got)\n\t\t\tcandidateDir := filepath.Join(',
    )
    source = replace_once(
        source, '\t\t\tst := recoveryStore(t, dir)',
        '\t\t\tscenarioRecord(t, "cli-process", map[string]any{"args":request.Args,"output":output.String(),"exit_code":request.ProcessState.ExitCode()})\n'
        '\t\t\tst := recoveryStore(t, dir)',
    )
    source = source.replace("c.management(t, ", "scenarioManagement(t, c, ")
    source = source.replace("findPackage(t, ", "scenarioPackage(t, ")
    return source


EVIDENCE_HELPERS = r'''
// scenarioRecord 保留原始管理响应与检查点，供场景执行者复查。
func scenarioRecord(t *testing.T, operation string, payload any) {
    t.Helper()
    file, err := os.OpenFile(filepath.Join(os.Getenv("ALONABOT_SCENARIO_EVIDENCE"), "events.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
    if err != nil { t.Fatal(err) }
    writeErr := json.NewEncoder(file).Encode(map[string]any{"operation":operation, "payload":payload})
    closeErr := file.Close()
    if writeErr != nil { t.Error(writeErr) }
    if closeErr != nil { t.Error(closeErr) }
}

func scenarioStartCore(t *testing.T, dir string, extra ...string) *core {
    t.Helper()
    c := startEmptyCore(t, dir, extra...)
    // 后登记的清理先执行，保留启动后的全部日志；成功流程已显式等待 Core 正常退出。
    t.Cleanup(func() { scenarioRecord(t, "core-log", c.log()) })
    return c
}

func scenarioManagement(t *testing.T, c *core, dir, token string, args ...string) string {
    t.Helper()
    output := c.management(t, dir, token, args...)
    scenarioRecord(t, "cli", map[string]any{"binary":filepath.Join(binDir,"alonabot"), "args":args, "output":output})
    return output
}

func scenarioFailure(t *testing.T, c *core, dir string, args ...string) string {
    t.Helper()
    output := managementFailure(t, c, dir, args...)
    scenarioRecord(t, "cli-rejection", map[string]any{"args":args, "output":output})
    return output
}

func scenarioPackage(t *testing.T, c *core, dir, id string) manage.Package {
    t.Helper()
    item := findPackage(t, c, dir, id)
    scenarioRecord(t, "package", item)
    return item
}

func scenarioHTTP(t *testing.T, c *core, dir, method, path string, body ...string) (int, []byte) {
    t.Helper()
    status, raw := fixtureHTTP(t, c, dir, method, path, body...)
    scenarioRecord(t, "http", map[string]any{"method":method, "path":path, "status":status, "body":string(raw)})
    return status, raw
}
'''


def run(args):
    if os.name != "posix":
        raise ValueError("this scenario uses the Unix integration harness")
    building = args.scenario in ("build", "build-recovery")
    binary = None if building else args.alonabot.resolve(strict=True)
    overlay_core = args.overlay_core.resolve(strict=True) if args.scenario in ("recovery", "survivor") else None
    directory = args.out if building else args.data_dir
    directory = directory.resolve()
    if building:
        directory.mkdir(mode=0o700)
    elif not directory.is_dir() or any(directory.iterdir()):
        raise ValueError("--data-dir must be an existing empty temporary directory")

    evidence = directory / "scenario-evidence"
    evidence.mkdir(mode=0o700)
    binaries = evidence / "bin"
    binaries.mkdir(mode=0o700)
    if binary is not None:
        (binaries / "alonabot").symlink_to(binary)
    environment = os.environ.copy()
    environment.update(
        ALONABOT_SCENARIO_BIN=str(binaries),
        ALONABOT_SCENARIO_OUT=str(directory),
        ALONABOT_SCENARIO_DATA_DIR=str(directory),
        ALONABOT_SCENARIO_EVIDENCE=str(evidence),
        ALONABOT_SCENARIO_MODE=args.scenario,
    )
    source = TEST.read_text()
    harness = HARNESS.read_text()
    setup = ""
    packages = []
    selected = "^TestT09ManagementFlow$"
    if args.scenario in ("build-recovery", "recovery", "survivor"):
        source = RECOVERY_TEMPLATE.read_text() + EVIDENCE_HELPERS
        selected = "^TestScenarioRecovery$"
    if building:
        if args.scenario == "build":
            source = fixture_source(source)
        harness = replace_once(harness, 'import (', 'import (\n "crypto/ed25519"\n "crypto/rand"')
        setup = '''
        _, private, err := ed25519.GenerateKey(rand.Reader)
        if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
        packageKey = private
'''
    elif args.scenario == "management":
        source = management_source(source)
        packages = [str(path.resolve(strict=True)) for path in args.packages]
        for index, package in enumerate(packages, start=1):
            environment[f"ALONABOT_SCENARIO_PACKAGE_{index}"] = package
    else:
        environment["ALONABOT_SCENARIO_OVERLAY_CORE"] = str(overlay_core)
        for name in ("recovery", "probe", "adapter") if args.scenario == "recovery" else ("recovery",):
            package = str(getattr(args, f"{name}_package").resolve(strict=True))
            packages.append(package)
            environment[f"ALONABOT_SCENARIO_{name.upper()}_PACKAGE"] = package
        if args.scenario == "survivor":
            selected = "^TestRecoveryEndToEnd$/^started$"
    harness = replace_once(
        harness,
        "func TestMain(m *testing.M) {",
        'func TestMain(m *testing.M) {\n'
        ' if directory := os.Getenv("ALONABOT_SCENARIO_BIN"); directory != "" {\n'
        '  binDir = directory\n' + setup + '  os.Exit(m.Run())\n }',
    )
    replacements = {}
    sources = [(TEST, source), (HARNESS, harness)]
    if args.scenario == "survivor":
        sources.append((RECOVERY, survivor_source(RECOVERY.read_text())))
    for original, content in sources:
        overlay_file = evidence / original.name
        overlay_file.write_text(content)
        replacements[str(original)] = str(overlay_file)
    overlay = evidence / "overlay.json"
    overlay.write_text(json.dumps({"Replace": replacements}, indent=2) + "\n")
    subprocess.run(["gofmt", "-w", *replacements.values()], check=True)
    command = ["go", "test", "-overlay", str(overlay), "-count=1", "-v", "./e2e", "-run", selected]
    inputs = {
        "scenario": args.scenario,
        "platform": platform.platform(),
        "packages": packages,
        "directory": str(directory),
        "command": command,
    }
    if binary is not None:
        inputs["binary"] = str(binary)
    if overlay_core is not None:
        inputs["overlay_core"] = str(overlay_core)
    (evidence / "inputs.json").write_text(json.dumps(inputs, indent=2) + "\n")
    print(f"Evidence: {evidence}", flush=True)
    with (evidence / "go-test.log").open("w") as log:
        result = subprocess.run(command, cwd=ROOT / "tests", env=environment, stdout=log, stderr=subprocess.STDOUT)
    print((evidence / "go-test.log").read_text(), end="", flush=True)
    return result.returncode


def main():
    parser = argparse.ArgumentParser(description="Build fixtures or run management and recovery scenarios against supplied binaries.")
    commands = parser.add_subparsers(dest="scenario", required=True)
    build = commands.add_parser("build", help="build three recovery fixture packages")
    build.add_argument("--out", type=Path, required=True, help="new output directory")
    recovery_build = commands.add_parser("build-recovery", help="build checkpoint Core and recovery, probe and adapter packages")
    recovery_build.add_argument("--out", type=Path, required=True, help="new output directory")
    management = commands.add_parser("management", help="run the complete management flow")
    management.add_argument("--alonabot", type=Path, required=True)
    management.add_argument("--data-dir", type=Path, required=True, help="empty temporary directory")
    management.add_argument("packages", nargs=3, type=Path, metavar="ARCHIVE", help="1.0.0, 2.0.0 and 3.0.0 archives, in order")
    for name in ("recovery", "survivor"):
        scenario = commands.add_parser(name, help="run crash recovery and gates" if name == "recovery" else "run the surviving writer and lock recovery")
        scenario.add_argument("--alonabot", type=Path, required=True, help="CLI binary used for every management action")
        scenario.add_argument("--overlay-core", type=Path, required=True, help="Core with the documented temporary checkpoint overlay")
        scenario.add_argument("--data-dir", type=Path, required=True, help="empty temporary directory")
        scenario.add_argument("--recovery-package", type=Path, required=True)
        if name == "recovery":
            scenario.add_argument("--probe-package", type=Path, required=True)
            scenario.add_argument("--adapter-package", type=Path, required=True)
    return run(parser.parse_args())


if __name__ == "__main__":
    sys.exit(main())
