# Platform acceptance

[简体中文](README.zh-CN.md)

`platform` consumes an explicitly selected Core and current-platform packages and records runtime identities, commands and results. Unix/Windows entry points cover first authentication, real echo round trips, showcase HTTP, 100 start/stop cycles, Core kill/restart, cooperative subtree cleanup and full transactional recovery. CGO build records additionally enable a real C function call. All steps must succeed for `passed`; omitting CGO yields `incomplete`. Four Linux/macOS environments passed completely. Windows cross-builds completed; runtime is still “not run”, deferred by maintainers on 2026-11-20 under [Windows runtime deferral](../../changelog/v0.1.0/en.md#added). Actual candidates, environments and original failures are in [TEST-MATRIX §3.6](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md#36-五平台运行增量t-102--v-32v-34).

Build the acceptance program, Core and three example packages from candidate source, and prepare the local `platform-tree` plugin below. Keep fixture source/artifacts in a git-ignored local directory with the evidence. These are native fish commands; for cross-builds set both the program and Core targets and use the same `--target OS/arch` for `pack`.

```fish
set artifacts (mktemp -d /tmp/alonabot-platform-build.XXXXXX)
go build -o "$artifacts/platform" ./scripts/platform
go build -o "$artifacts/alonabot" ./cmd/alonabot
"$artifacts/alonabot" pack examples/plugins/echo --out "$artifacts/packages/echo"
"$artifacts/alonabot" pack examples/plugins/showcase --out "$artifacts/packages/showcase"
# telegram_source 指向独立 alonabot-adapter-telegram checkout。
"$artifacts/alonabot" pack "$telegram_source" --out "$artifacts/packages/telegram"
# 将 tree_source 设置为本地 platform-tree 工程的绝对路径。
"$artifacts/alonabot" pack "$tree_source" --out "$artifacts/packages/tree"
# cgo_source 指向下文约定的 ignored 本地 CGO 插件 main.go。
uv run --no-project --offline --no-python-downloads scripts/platform/cgo_build.py \
    --root "$PWD" --source "$cgo_source" --target darwin/arm64 --out "$artifacts/cgo"
"$artifacts/platform" \
    --alonabot "$artifacts/alonabot" \
    --packages "$artifacts/packages" \
    --scenarios "$PWD/scripts/scenarios" \
    --cgo "$artifacts/cgo/build.json" \
    --out "$artifacts/result.json" \
    --environment native \
    --host 'macOS Apple Silicon'
```

Options match `--help`:

| Option | Input |
| --- | --- |
| `--alonabot` | Candidate Core executable path |
| `--packages` | Current-platform package directory containing a unique target `.tar.gz` for each of echo, showcase, alonabot-adapter-telegram and platform-tree, with adjacent `release.json`; archive paths/lengths are recorded recursively |
| `--scenarios` | Candidate checkout's `scripts/scenarios`, including `run.py` and recovery templates |
| `--cgo` | Local `platform-cgo` build-record JSON; optional, but omission yields overall `incomplete` |
| `--out` | New JSON file; parent must exist; existing files are rejected |
| `--environment` | Required: `native`, `vm`, `container` or `emulation` |
| `--host` | Operator-supplied host/VM description, separately recorded as `host_description` |

Each run creates independent temporary `cwd`/`data` directories and records absolute paths in JSON. Commands run from that `cwd`; inputs resolve relative to the launch directory. Build the program and Core for the same target. Rosetta requires both and all components to use `darwin/amd64`, environment `emulation`, host description Apple Silicon/Rosetta. Containers use `container` with the actual host and container distribution described.

Rosetta builds may use host-native Go, Git and C compilers with explicit target `darwin/amd64`; launch acceptance with `arch -x86_64`. If `/usr/bin/git` under Rosetta fails because `xcrun` loads ARM-only `libxcrun`, find existing Command Line Tools' native Git through `PATH` during builds. Direct Clang invocation from that directory needs `CC`/`CXX` specifying `-isysroot` and the existing macOS SDK. Tool architectures, full arguments and failures are in [TEST-MATRIX](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md) §3.6.

For Docker Desktop, point `TMPDIR` to the container-local filesystem and copy the whole workspace to evidence afterward. Shared mounts may hold build inputs/final JSON. Existing comparisons found different Core-lock observations on shared and container-local filesystems; see original failures/comparisons in [TEST-MATRIX](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md) §3.6.

JSON records runtime OS/architecture/Go version, raw system-version output, Core/package Go build info, package paths/lengths, and each step's UTC times, millisecond duration, full combined output, exit code and error. `host_description` is operator-declared; system output/build targets are separate. `runs` records Core commands/PIDs/dynamic listeners/full log locations and component activation PIDs/generations. `purpose` separates initial business, cycles, kill and recovery; `cycle` is 1–100, and each step's `core_pid` identifies its run. `bot_calls` preserves real requests to the local Bot API simulator.

First startup has no TTY; the management token persists as `admin_token` with mode 0600 in this run's data directory. Explicitly obtain it with `admin-token show`, then use Cookie/Origin/CSRF to log in and enable components. Installation uses real CLI Bearer authentication, automatically reading the selected `--data-dir` credentials. A controlled `base_url` and Host forwarding to the actual dynamic listener follow existing integration tests. Credential values in tokens/login requests/successful responses are masked with `output_redacted`; data and credential files remain in temporary storage for reproduction. First login also checks `admin_token` remains.

Telegram Adapter, echo and showcase are real separate component processes; Bot API is a loopback simulator inside acceptance. Send `/echo`, check reply body/session, then access authenticated showcase `/status`, its page and bridge-script references.

Management sessions check every raw HTTP response; redirects retain original status, `Location` and body. A status mismatch fails. Bot API body-read failures preserve the full original error and received bytes in `%q` format for reproduction from failure JSON.

Local `platform-tree` uses the candidate SDK, manifest ID `platform-tree`, role `plugin`. `OnStart` uses `os/exec` to launch the same executable's child mode, inheriting process group/Job. The child accepts connections on `127.0.0.1:0`, closing the listener/exiting on stdin EOF. The parent holds the input pipe; `OnStop` closes it and calls `Wait`. Authenticated `GET /state` returns `{"parent_pid":123,"child_pid":124,"address":"127.0.0.1:34567"}` with actual PIDs/listener address. Normal and disconnect Core shutdown both follow this cooperative path.

After the first business run, reuse data_dir for 100 real start/stop cycles. Read management state/activation logs, check each component's persisted Active/generation/PID and subtree state. `running`/`released` record Core, four components, child PID, Unix process group, five locks and Core/child ports. During running, processes must live, locks be held and ports connect. After shutdown, PIDs/groups must disappear, existing locks be genuinely acquirable and original ports rebindable. Negative Unix PIDs denote groups; zombies count as present. Existing lock files are opened, released after checks and retained.

After cycles, kill Core: Unix checks SIGKILL and original `signal: killed`; Windows checks TerminateProcess exit code 1. Confirm the old tree/ports/locks released, restart, compare each same Active and repeat shutdown checks. Normal shutdown, kill, full transactional recovery and CGO results remain separate. Deadlines reuse real-process integration tests: 20 seconds, 15 for shutdown, observing real responses/OS/locks every 20ms. Unix uses interrupt/SIGKILL/flock; Windows uses targeted CTRL_BREAK, TerminateProcess, process-handle waits and LockFileEx first-byte locks. Start from a console-connected terminal: Core's separate process group receives normal stop signals; components stop by protocol and Job cleanup.

Unix invokes the [existing scenario entry point](../scenarios/README.md) with `os/exec`, builds three management package versions/checkpointed Core, then runs `management`, seven `recovery` windows and `survivor/started`. All original assertions remain: real DB, Active/Session/five committed Journals; DB/resource references/candidate cleanup after PREPARED/commit/OnReady failure; six admission gates; Degraded/start counts while a surviving writer holds locks and same Active/DB recovery.

This stage needs existing Go, Git, gofmt and `uv` with checkout-approved Go dependencies. Python uses `uv run --no-project --offline --no-python-downloads`. Candidates need `vcs.revision`; record source HEAD and use Git to check `cmd`, `internal`, `sdk`, `contracts`, `go.mod`, `go.sum` unchanged from that commit. Builds inherit `GOFLAGS`, explicitly set the program's OS/architecture and check seven artifacts' targets through build info. Source differences/missing tools/build failures retain original errors.

Compile in the source checkout. A temporary `go test -exec` launcher moves test processes and their Core/CLI children into independent `cwd` directories; each scenario has its own data directory. On cancellation, scenarios first clean up their own Core/lock-holding plugins through existing cleanup, then return cancellation. Wait for cleanup after interruption.

Windows `windows_sources.py` loads original source-generation functions, selects complete referenced Go declarations/original assertions by AST and compiles five native `scenario-test.exe` programs executed through Go testing. Builds retain source cwd; management/recovery/survivor TestMain switches to independent cwd. Archive entry points use `.exe`, read-only DB uses Windows file URIs, named Events hold/release checkpoints. Survivor's temporary Core names its Job; acceptance retains an extra Job handle so the real old SDK still holds runtime.lock after Core kill. Original Degraded/start count/Active/Journal/DB assertions continue, then old PID termination and recovery verification. This hook exists only in retained temporary overlays.

Windows also requires existing Go, Git, uv and all candidate target packages; Core/acceptance use `.exe`. Source-generation inputs, effective declarations, overlays, executables and per-scenario arguments stay in `scenario-evidence`. Maintainers deferred Windows runtime/CGO; rescheduling requires local tools and platform results. Original failures remain in the historical acceptance record above.

JSON `recovery` preserves source/candidate commits, build arguments, seven artifact identities, three scenario inputs and complete `events.jsonl`. Each scenario retains `go-test.log`, overlay, real `state.sqlite` and business DB. Required test success records and independent-cwd markers are checked; missing evidence fails. Partial events and corrupt-input raw bytes remain in failed results. All build/run/evidence paths are under JSON `workspace`, following original scenario layouts.

CGO fixture source/artifacts stay ignored: an independent `package main` using candidate SDK, `genimpl.Register("main", genimpl.Resources{})` and `sdk.Run`. Real `GET /native?value=81` parses input, calls C math `sqrt`, returns `{"value":9,"compiler":"C __VERSION__","os":"darwin","arch":"arm64","pid":123}` using actual platform/PID/compiler. Linux/Windows link math via `#cgo linux windows LDFLAGS: -lm`. `build/pack` fixes `CGO_ENABLED=0`; local `cgo_build.py` separately compiles existing CGO source and creates unsigned packages: ID `platform-cgo`, version `0.1.0`, standard archive/external `release.json`/required integrity fields.

`cgo_build.py` needs existing Go/C compiler/uv. `--root` is candidate checkout, `--source` the single Go file above, `--target` this run's `os/arch`, `--out` a new directory. Preserve target/CC/CXX/CGO arguments/GOFLAGS from `go env`; run actual compiler `--version` and `go build -x -work`. Full commands/cwd/start-end/durations/stdout/stderr/exit codes go to `build.json`; Go build workspace remains. A host descriptor reads target build-info only. Build/compiler failures still write records with original errors; obtain installation approval for missing tools.

Build-record `source`/`binary`/`archive` are absolute paths; `environment` is the object above. `compiler_version`/`build` contain `command`, `directory`, `started`, `duration_ms`, `output`, `exit_code`, `status`, optional `error`; `output` preserves full combined raw stdout/stderr. Local `go_env` uses the same format, preserving actual command/original error/exit code/known inputs including `GOEXPERIMENT` on failure and still writing `build.json`. External records may supply environment values. Acceptance reports recorded environment-command failure first, then checks compiler/build/paths, successful status/target/actual binary `CGO_ENABLED=1`, directly compares archive entry/binary bytes, and reads direct dependencies from Mach-O/ELF/PE. PE DLL names come from named imported symbols. Incomplete CGO builds keep failure evidence; actual builds/runs determine support.

CGO runs separately as `purpose=cgo`: after 100 cycles and kill/recovery, start real Core again, install/enable the package, access the C route through a management session and compare activation PID. Four original components run too; shutdown checks five components/cooperative subtree/six locks/released ports. Full results total 104 business Core runs, or original 103 without CGO. `cgo` stores build records/binary identity/direct libraries/status; `steps` stores raw HTTP responses.

Prepare independent assets for three local environments with Core/four ordinary packages/acceptance on the same target: native macOS `darwin/arm64`, Docker Linux `linux/arm64`, Rosetta `darwin/amd64`. Explicitly pass each target to ordinary `pack --target` and CGO builder `--target`. Linux CGO uses existing C tools in-container; run/check Rosetta x86_64 artifacts in that environment. Each JSON records `--packages`, `--cgo`, independent workspace/build identities; cross-builds are listed separately.

Exit 0 means executed steps succeeded; JSON `status` determines overall acceptance. Step failure exits nonzero, marks `failed` and keeps original errors. Missing `--out`, parse/result-creation errors print directly to stderr. Run/evidence directories remain for diagnosis; no new dependencies or installed tools are required.
