"""为指定本地 platform-cgo 源码构建安装包，保留 C 工具链与完整构建证据。"""

import argparse
from datetime import datetime, timezone
import hashlib
import io
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import tarfile
import time


INSPECTOR = r'''// inspect 输出实际二进制的标准 Go build-info 文本。
package main
import ("debug/buildinfo"; "fmt"; "os")
func main() {
 info,err:=buildinfo.ReadFile(os.Args[1])
 if err!=nil { fmt.Fprintln(os.Stderr,err);os.Exit(1) }
 fmt.Print(info.String())
}
'''


def execute(command, directory, environment):
    started = datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")
    before = time.monotonic()
    result = {"command": command, "directory": str(directory), "started": started}
    try:
        completed = subprocess.run(command, cwd=directory, env=environment, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        result.update(output=completed.stdout.decode(), exit_code=completed.returncode, status="passed" if completed.returncode == 0 else "failed")
        if completed.returncode:
            result["error"] = f"command exited with status {completed.returncode}"
    except OSError as error:
        result.update(status="failed", error=str(error), output="")
    result["duration_ms"] = (time.monotonic() - before) * 1000
    return result


def checked(command, directory, environment):
    result = execute(command, directory, environment)
    if result["status"] != "passed":
        raise RuntimeError(json.dumps(result, indent=2))
    return result["output"]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--target", required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    root, source, out = args.root.resolve(strict=True), args.source.resolve(strict=True), args.out.resolve()
    out.mkdir(mode=0o700)
    goos, goarch = args.target.split("/")
    environment = dict(os.environ, GOOS=goos, GOARCH=goarch, CGO_ENABLED="1")
    keys = ("GOOS", "GOARCH", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOGCCFLAGS", "GOFLAGS", "GOEXPERIMENT")
    binary = out / ("component.exe" if goos == "windows" else "component")
    archive = out / f"platform-cgo-0.1.0-{goos}-{goarch}.tar.gz"
    # 环境探测也会失败；先保存明确输入，使原始命令和错误始终有结构化落点。
    record = {"source": str(source), "binary": str(binary), "archive": str(archive),
              "environment": {key: environment[key] for key in keys if key in environment}}
    try:
        record["go_env"] = execute(["go", "env", "-json", *keys], root, environment)
        if record["go_env"]["status"] != "passed":
            raise RuntimeError(json.dumps(record["go_env"], indent=2))
        settings = json.loads(record["go_env"]["output"])
        record["environment"] = settings
        compiler = shlex.split(settings["CC"])
        compiler[0] = shutil.which(compiler[0]) or compiler[0]
        record["compiler_version"] = execute([*compiler, "--version"], root, environment)
        if record["compiler_version"]["status"] != "passed":
            raise RuntimeError(json.dumps(record["compiler_version"], indent=2))
        record["build"] = execute(["go", "build", "-x", "-work", "-o", str(binary), str(source)], root, environment)
        if record["build"]["status"] != "passed":
            raise RuntimeError(json.dumps(record["build"], indent=2))
        # 描述器在宿主执行，只读取目标文件；Windows或Rosetta交叉构建不会在这里运行目标程序。
        host_environment = dict(os.environ, CGO_ENABLED="0", GOFLAGS="")
        host_environment.pop("GOOS", None)
        host_environment.pop("GOARCH", None)
        inspector = out / "inspect.go"
        inspector.write_text(INSPECTOR)
        buildinfo = checked(["go", "run", str(inspector), str(binary)], root, host_environment)
        manifest = f"""manifest_version: 1
id: platform-cgo
name: Platform CGO
version: 0.1.0
source: github.com/AlonaBot/AlonaBot
role: plugin
protocol_min_major: 1
protocol_max_major: 1
required_capabilities: []
entrypoint: bin/{binary.name}
migration_mode: none
prepare_effects: none
os: {goos}
arch: {goarch}
"""
        files = {"plugin.yaml": manifest.encode(), "build-info": buildinfo.encode(), f"bin/{binary.name}": binary.read_bytes()}
        with tarfile.open(archive, "w:gz") as package:
            for name, raw in files.items():
                header = tarfile.TarInfo(name)
                header.size, header.mode = len(raw), 0o700 if name.startswith("bin/") else 0o600
                package.addfile(header, io.BytesIO(raw))
        # 仅计算安装契约强制要求的归档完整性字段。
        asset = {"os": goos, "arch": goarch, "url": "https://example.invalid/" + archive.name, "size": archive.stat().st_size,
                 "sha256": hashlib.sha256(archive.read_bytes()).hexdigest()}
        release = {"publisher_key_id": "", "source": "github.com/AlonaBot/AlonaBot", "plugin_id": "platform-cgo", "version": "0.1.0", "assets": [asset]}
        (out / "release.json").write_text(json.dumps(release, indent=2) + "\n")
    finally:
        (out / "build.json").write_text(json.dumps(record, indent=2) + "\n")
    print(out / "build.json")


if __name__ == "__main__":
    main()
