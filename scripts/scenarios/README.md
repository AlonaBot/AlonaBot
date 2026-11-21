# Repeatable management and recovery scenarios

[简体中文](README.zh-CN.md)

These T-10 scenarios use specified binaries, temporary `data_dir` directories and test packages to verify the production management flow, crash recovery and surviving writers. The management flow reuses every assertion in the [management integration test](../../tests/e2e/management_e2e_test.go); the survivor scenario reuses `started` from the [recovery integration test](../../tests/e2e/recover_e2e_test.go). Additional recovery and admission assertions are in the [scenario template](recovery.go.tmpl). Temporary Go overlays replace inputs and collect evidence; management actions always invoke the program supplied by `--alonabot`.

## Build three test package versions and run the management flow

Use the project's required Go toolchain and `uv` from the candidate repository root. The entry point uses only Python's standard library and existing Go dependencies. It currently supports Unix platforms such as macOS/Linux, matching the original integration test's `unix` build tag; T-10 records actual runs on other platforms separately.

The following commands use fish. `artifacts` stores this run's independently built Core, fixture source, signed packages and raw logs. You can also replace `--alonabot` with the absolute path of the candidate Core.

```fish
set artifacts (mktemp -d /tmp/alonabot-scenarios.XXXXXX)
go build -o "$artifacts/alonabot" ./cmd/alonabot

uv run --no-project scripts/scenarios/run.py build --out "$artifacts/packages"

set data_dir (mktemp -d /tmp/alonabot-management.XXXXXX)
uv run --no-project scripts/scenarios/run.py management \
    --alonabot "$artifacts/alonabot" --data-dir "$data_dir" \
    "$artifacts/packages/1.0.0/package.tar.gz" \
    "$artifacts/packages/2.0.0/package.tar.gz" \
    "$artifacts/packages/3.0.0/package.tar.gz"
```

`build` reuses `recoveryFixtureSource` and `writeTestPackage` to compile real 1.0.0, 2.0.0 and 3.0.0 programs for `recovery` / `github.com/example/recovery`, adding the configuration schema, `/info` version/configuration endpoint and `OnReady` logs. All three packages use one temporary Ed25519 signing key. Each version directory contains `main.go`, `plugin`, `package.tar.gz`, `release.json`, `release.json.sig` and `release.pub`. The private key is used only in the build process's memory.

Alternative build outputs must provide the same identities, versions, public signing key, configuration fields and HTTP behavior, with release manifest, signature and public key beside each archive. `build --out` requires a new output directory; `management --data-dir` requires an existing empty temporary directory.

The management flow runs in this order:

1. install → enable; check initial Active, Session, actual program version, `label=initial` and initial SQLite value 1.
2. Write DB=41 and edit `config.json`; confirm the old configuration remains in use until reload.
3. restart keeps Active and replaces Session; CLI reload creates a new Active and independent DB, with configuration `reloaded`.
4. update to 2.0.0; stop Core, issue a button in the real state database and verify it is valid, then resume Core.
5. pin 2.0.0; rejection of update 3.0.0 preserves Active, Session and DB. Remove the pin, update successfully with the same package, then write DB=99.
6. disable → default uninstall → reinstall from the same source → enable preserves generation, configuration and DB=99.
7. Choose a 2.0.0 snapshot from the actual generations API retention list. Explicit rollback creates a new Active/Session and independent DB, restoring 2.0.0, `reloaded` and DB=41.
8. Stop Core cleanly. Check persisted Active and the operation/before-and-after generations/DB paths of 5 committed Journal entries; use real button resolution to confirm the generation change invalidated the old button.

## Build recovery artifacts and run both fault scenarios

Run these fish commands from the repository root. The CLI is a normal build; `--overlay-core` runs a temporary Core with exact checkpoints. `build-recovery` builds independent artifacts; runtime entry points consume the supplied binaries and packages.

```fish
set recovery_artifacts (mktemp -d /tmp/alonabot-recovery.XXXXXX)
go build -o "$recovery_artifacts/alonabot" ./cmd/alonabot
uv run --no-project scripts/scenarios/run.py build-recovery \
    --out "$recovery_artifacts/fixtures"

set recovery_data (mktemp -d /tmp/alonabot-crash.XXXXXX)
uv run --no-project scripts/scenarios/run.py recovery \
    --alonabot "$recovery_artifacts/alonabot" \
    --overlay-core "$recovery_artifacts/fixtures/fault-core" \
    --data-dir "$recovery_data" \
    --recovery-package "$recovery_artifacts/fixtures/recovery/package.tar.gz" \
    --probe-package "$recovery_artifacts/fixtures/gateprobe/package.tar.gz" \
    --adapter-package "$recovery_artifacts/fixtures/adapter-package/package.tar.gz"

set survivor_data (mktemp -d /tmp/alonabot-survivor.XXXXXX)
uv run --no-project scripts/scenarios/run.py survivor \
    --alonabot "$recovery_artifacts/alonabot" \
    --overlay-core "$recovery_artifacts/fixtures/fault-core" \
    --data-dir "$survivor_data" \
    --recovery-package "$recovery_artifacts/fixtures/recovery/package.tar.gz"
```

`build-recovery --out` also requires a new directory. Output includes `fault-core`, real `observed-plugin`, `probe`, `adapter`, their generated source and three signed packages under `recovery/`, `gateprobe/` and `adapter-package/`. Each package directory retains `package.tar.gz`, `release.json`, `release.json.sig` and `release.pub`, using this build's shared temporary Ed25519 key. `overlay.json` reuses the original recovery hooks; `fault-overlay.json` adds FIFO release and pre-binding checkpoints. All modified source mapped by the overlay remains in the output directory.

Artifacts are built for the execution platform. Alternatives must use `recovery` identity `github.com/example/recovery`, version 1.0.0, `migration_mode: none`, `prepare_effects: managed_db`, SQLite `/value`, `starts`, `candidate-value`, `fail-ready` and observations of six request types. `gateprobe` provides `/service`, `/bus`, `attempts` and independent scheduled `clock`; the Adapter is `alonabot-adapter-fixture`, protocol `devloop`, instance `one`. The actual files retained in the build directory define the source and declarations.

### Recovery checkpoints and assertions

`recovery` creates seven independent run directories under `<data_dir>/`:

| Subdirectory | Failure window | Recovery result |
| --- | --- | --- |
| `prepared` | Kill Core after PREPARED, before commit | Old DB 41 remains; candidate DB 77 is removed; old Active returns with a new Session |
| `committed` | Kill Core after successful CommitCandidate, before ACTIVATE | Only new Active resumes with DB=77; old DB 41 remains; committed resources stay referenced |
| `first-prepared` | Kill Core after PREPARED on first enable | No old Active; remains inactive with HTTP 503; uncommitted candidate is reclaimed |
| `ready` | Candidate OnReady fails after reading `fail-ready`, then kill Core | New Active/committed/Degraded remains; faulted restart still selects the new generation; removing the fault restores DB=77 |
| `gates-prepared` | Candidate prepared, not committed | Request log stays unchanged until checkpoint release; all six entry points then run in the new PID |
| `gates-committed` | Committed, before ACTIVATE | Same as above |
| `gates-binding` | After ACTIVATED, before full Core business binding | Same as above |

Core locates windows with `ALONABOT_TEST_CRASH_PHASE` and `ALONABOT_TEST_CHECKPOINT=<subdirectory>/checkpoint`. Crash cases kill Core after observing the marker; admission cases hold/release the exact window through `<subdirectory>/checkpoint.release` FIFO. Admission checks advance at least one task cycle through a real timer callback in a separate plugin alongside command/event/Service/Bus/HTTP requests. All six entry points are verified before/after normal binding. Logs record PID and entry-point type for every request; after handover only the new PID is accepted.

After Core exits and releases runtime locks, recovery checks read the real state database: Journal OpID/Kind/ExpectedOld/NewGeneration/Phase/Session/DB paths, Active, generation references and resource files, plus read-only business SQLite queries. API responses, raw Core logs and plugin `starts`/`requests` files are preserved.

### Surviving writer shutdown and recovery

`survivor` uses `<data_dir>/started/`. Hold the checkpoint after Journal started but before stopping the old process, then kill Core. A temporary SDK overlay writes `runtime-held` on disconnect and blocks on `hold-runtime` FIFO, keeping the old plugin's real `runtime.lock`. Restarted Core must be Degraded with unchanged Active/start count and DB=41. Terminate that scenario-owned old process, wait for the real lock release, then CLI restart must restore the same Active/DB. The uncommitted Journal becomes cleaned, the candidate directory is reclaimed and the new Session activates only once.

Fault processes, FIFOs and system runtime locks use Unix APIs; the entry point supports macOS/Linux. T-10 records runs on other targets. Temporary Core/SDK checkpoints serve this scenario, with build/run locations recorded in evidence.

## Read execution evidence

Success prints `PASS` and returns 0. Failure preserves full Go test output and returns its exit code. Cleanup terminates only scenario-owned processes; run directories and artifacts remain.

| Location | Contents |
| --- | --- |
| `<out>/scenario-evidence/` | Build commands, inputs/overlays and raw build output |
| `<data_dir>/scenario-evidence/inputs.json` | Absolute paths to the specified CLI, overlay Core and packages; execution platform and actual Go command |
| `<data_dir>/scenario-evidence/overlay.json`, `*_test.go` | Effective test entry point and helper source for inspecting replaced inputs and assertions |
| `<data_dir>/scenario-evidence/go-test.log` | Full stdout/stderr and test results |
| `<data_dir>/scenario-evidence/events.jsonl` | Management output, package/HTTP snapshots and raw Core logs; recovery also records subdirectories, survivor records old PID, start count and recovery Journal |
| `<data_dir>/state.sqlite` or `<data_dir>/<subdirectory>/state.sqlite` | Persistent state, generations and resource references after Core exits; Journal/DB checks also appear in Go logs |
| `plugins/recovery/` in each run directory | Configuration/business DB, `starts`, `requests`, crash checkpoints and survivor markers |

The CLI's local HTTP forwarder reuses integration-test helpers, changing only Host to Core's configured `base_url`. Requests still use real Bearer authentication and production management entry points. Button evidence uses the same state database after Core exits and the integration test's fixed route description.

Source-anchor changes produce an explicit error; update the scenario entry point before rerunning. Temporary overlays retain original main-path assertions. This scenario's `TestMain` skips ordinary Core/fixture builds and uses temporary symbolic links to specified binaries.
