# Installation and deployment

[简体中文](deployment.zh-CN.md)

This guide covers installation, authentication, configuration, component management and recovery for AlonaBot administrators. For source work see [development](development.md); for release builds see [release maintenance](releasing.md).

## Choose an installation method

Download official assets from [v0.2.0 Release](https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0). Five-platform archives and `install.sh` provide Core only; Docker Compose bundles Core, WebUI and Agent. Install Adapters and help as needed. See [release notes](../changelog/v0.2.0/en.md) for actual platform runs and known limitations.

| Method | Use case |
| --- | --- |
| Release archive | Run Core manually on any of the five release targets |
| install.sh | Linux systemd service or macOS current-user LaunchAgent |
| Docker Compose | Linux amd64/arm64 containers with persistent data volumes |

## Run a release archive manually

Download/extract the archive for your platform. It contains `alonabot` (`alonabot.exe` on Windows), `LICENSE` and `source.json`. Run in that directory:
```fish
./alonabot run --data-dir ./data
```

Management commands below assume `alonabot` is on PATH; use `./alonabot` from the extracted directory, or `alonabot.exe` on Windows. The first startup has no components. Obtain the management token under [authentication](#core-settings-and-http-entry-point) and install needed [components](#manage-components). After installing/enabling WebUI, log in at `http://127.0.0.1:7650`; configure the Agent Provider/model before enabling Agent.

## Install and upgrade a system service

Run as root on systemd Linux, or in the current user's login session on macOS. Both support amd64/arm64. Download the installer from the main repository, then install a specific version (fish commands; replace `VERSION` with the release version without leading `v`):
```fish
set version VERSION
curl --fail --location --proto '=https' --proto-redir '=https' \
  -o install.sh https://raw.githubusercontent.com/AlonaBot/AlonaBot/main/deploy/install.sh
# Linux
sudo sh install.sh --version $version
# macOS
sh install.sh --version $version
```

Omitting `--version` first pins a Release through `/latest`, downloads its archive, and checks its SHA256 against the GitHub Release API asset digest. `ALONABOT_RELEASE_BASE_URL` overrides the Release base URL; custom mirrors still supply `SHA256SUMS` from that same Release. The base URL and all redirects must be HTTPS. The script checks platform/permissions/system tools, downloads/verifies/extracts to temporary storage, then stops the service, replaces Core and removes old adjacent `packages/` installation media. Installed components in the data directory remain. If a custom mirror still supplies bundled archives, their `packages/` is copied too.

Linux uses `/opt/alonabot/`, data directory `/var/lib/alonabot`, service `alonabot.service`, running as system user `alonabot` with boot startup enabled. macOS uses `~/Library/Application Support/AlonaBot/`, its `data/`, and user LaunchAgent `io.github.alonabot` starting at login. Working directory is the program directory; arguments are `run --data-dir <absolute data path>`.

After stopping the service, the script runs `admin-token init --if-missing` as the service identity through `/dev/tty` when a controlling terminal exists. Otherwise Core initializes on first startup. The new CLI persists credentials with mode 0600 and prints only the path; existing credentials are reused. Read explicitly for remote WebUI login:
```fish
# Linux，以服务身份读取
sudo runuser -u alonabot -- /opt/alonabot/alonabot admin-token show --data-dir /var/lib/alonabot
# macOS，以当前用户读取
"$HOME/Library/Application Support/AlonaBot/alonabot" admin-token show --data-dir "$HOME/Library/Application Support/AlonaBot/data"
```

Login/confirmation preserve persistent credentials. Rerunning installation upgrades while keeping data, configuration, state database and installed components.
```fish
# Linux 卸载
sudo sh install.sh --uninstall
# macOS 卸载
sh install.sh --uninstall
```

Uninstall stops/unregisters the service, removes its definition/program/bundled packages and prints the retained data path. macOS `data/` stays in place.

## Container images

Download the release Compose file and start the bundle:
```fish
curl --fail --location --proto '=https' --proto-redir '=https' \
  -o docker-compose.yml https://raw.githubusercontent.com/AlonaBot/AlonaBot/v0.2.0/docker-compose.yml
docker compose up -d
```

Open `http://127.0.0.1:7650` in a browser on the host. The WebUI shows the one-time management token; save it and confirm to log in. Without a browser on the host, use the SSH tunnel below to complete confirmation while the port is still bound to loopback. To read the token for manual login:
```fish
docker compose exec alonabot alonabot admin-token show --data-dir /data
```

After logging in, install plugins and Adapters from the store or [from source](#source-builds), create account instances, then configure an LLM. Behind restricted networks, configure the [GitHub endpoints](#component-download-mirrors) and, for source builds, the [Go download settings](#core-settings-and-http-entry-point). Initial startup needs no TTY: Core writes `/data/admin_token`, retaining it after login for the local CLI. Images support `linux/amd64` and `linux/arm64`; compose chooses the host platform. Keep the pinned image version in compose aligned with the chosen Release.

The image runs as `10001:10001`. Its named volume mounts at `/data`, inheriting that image directory's ownership on first creation. For host bind mounts, grant UID/GID 10001 read/write/traversal access first; existing volumes need the same permissions. Retain the volume to keep configuration, state and component data.

Ports publish only on host `127.0.0.1`. With `ALONABOT_CONTAINER=1`, the default listener is `0.0.0.0:7650`, while `base_url` remains `http://127.0.0.1:7650`; explicit `listen` wins. When this container exception applies, startup emits one WARN reminding you to publish only on host loopback. Remote management can use `ssh -N -L 7650:127.0.0.1:7650 user@host`, then the local address. For HTTPS reverse proxies configure `base_url`/`trusted_proxies` under [Core settings and HTTP](#core-settings-and-http-entry-point).

Keep the host port bound to `127.0.0.1` until you confirm the first-run token locally. Only then change the binding to `0.0.0.0` or another non-loopback address. With Docker's userland proxy, even LAN requests can appear to come from the gateway; anyone who reaches the published port before confirmation could claim the token and gain full management access.

The first-run endpoint `POST /api/auth/bootstrap` accepts direct local connections. With `ALONABOT_CONTAINER=1`, it also accepts the IPv4 default gateway from `/proc/net/route` for host requests routed through Docker's bridge. Host and Origin must still match the local `base_url`; requests with `Forwarded`, `X-Forwarded-*` or `X-Real-IP` are rejected. After confirmation, the endpoint stops returning the token.

After installing an Adapter as needed, pass `TELEGRAM_BOT_TOKEN` through the compose service's `environment`; instance credentials in `/data/core_config.json` can use `"token": "${TELEGRAM_BOT_TOKEN}"`, preserved on writeback. The final scratch image contains Core, adjacent `packages/` and CA certificates. MCP stdio servers needing shell/Node/Python require your own image with those runtimes, retaining UID 10001, a writable data volume and `--init`.

## Manage components
```fish
alonabot run --data-dir ./data
```

`run` and library mode (`alonabot.Start`) share Core assembly in `internal/app`: CLI calls it directly, the root package provides a thin library wrapper. Management services install/start/stop components. CLI calls the API at `base_url`, automatically reads persistent credentials from `--data-dir`, connects directly to local listen and preserves base_url Host. Override the local connection address with `--admin-url`:
```fish
alonabot install pkg.tar.gz --yes             # Install the local archive
alonabot install alonabot-webui-admin --yes    # 从可执行文件旁 packages/<组件 ID>/ 安装
alonabot install --yes                        # 在当前目录或工程 bin/ 中选择唯一的本机包
alonabot install --source owner/repo --yes      # Select a platform archive via the GitHub Release API
alonabot install --source AlonaBot/alonabot-webui --yes
alonabot enable alonabot-webui-admin
alonabot install --source AlonaBot/alonabot-agent --yes
alonabot install --source AlonaBot/alonabot-adapter-telegram --yes
alonabot install --source AlonaBot/alonabot-plugin-help --yes
alonabot enable alonabot-plugin-help-official
alonabot enable alonabot-plugin-hello
alonabot list
alonabot uninstall alonabot-plugin-hello --purge-data --yes      # 或 --purge；不带选项时只卸载
alonabot adapter instance create alonabot-adapter-telegram primary --config instance.json
alonabot admin add telegram:123456
```

An existing regular-file positional argument to `install` selects that archive; other values locate a component ID in adjacent `packages/<component ID>/`, listing available official IDs if missing. Without an argument, author projects search `bin/`; other directories search themselves, also checking direct package subdirectories. Archive names select host assets; static WebUI uses `any/any`. Exactly one candidate installs; zero/multiple candidates print search paths/candidates and exit. Discovery is read-only; management API performs installation.

`--archive` explicitly selects an archive. Positional arguments, `--archive` and `--source` are mutually exclusive. Source confirmation, archive preflight, `--yes` and `--replace` use the management service. Browser installation and updates also need only the archive. After historical Release sidecars are removed, older Core versions need an upgrade for online installation and updates; see [release sequencing](releasing.md).

After enabling help, send `/whoami` in chat and pass its returned user string to `admin add`. The returned instance name identifies its source.

See [management API usage](#management-api) for endpoints and configuration reads/writes.

### Source builds

With `Build: true`, Core builds from source instead of downloading a Release asset, for example `POST /api/v1/packages/install` with `{"Source":"AlonaBot/alonabot-adapter-telegram","Build":true,"Ref":"v0.2.0","Confirm":true}`:

1. Core resolves `Ref` (branch, tag or commit; empty means the default branch) through `github_api_base`, downloads `https://github.com/<owner>/<repo>/archive/<commit>.tar.gz` using `github_download_prefix` when configured, and unpacks it under `<data_dir>/staging`. GitHub publishes no digest for source tarballs. When a prefix is used, the task log warns that the archive came through the prefix without digest verification. Use a trusted download mirror.
2. Core downloads the Go release Core itself was built with into `<data_dir>/toolchain/go` when missing, checking size and SHA-256 against the JSON listing of `go_download_base` (`?mode=json&include=all`, served by both `go.dev/dl/` and `golang.google.cn/dl/`).
3. The project is packed for Core's platform like `alonabot pack`: `CGO_ENABLED=0`, `GOTOOLCHAIN=auto`, modules come from `goproxy` and are verified by GOSUMDB (default `sum.golang.org`). `GOCACHE`, `GOMODCACHE`, `GOTMPDIR`, `TMPDIR` and `HOME` all live under `<data_dir>/cache`, so the scratch image can build. The Go subprocess inherits only `PATH`, `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` from Core, with its Go settings set explicitly (`GOENV=off`, `GOWORK=off`).
4. The archive then goes through the usual install preflight, source trust and same-kind replacement confirmation, bound to source `github.com/<owner>/<repo>`. When `Ref` is a SemVer tag the version is the tag and must match `plugin.yaml`; any other commit gets the `plugin.yaml` version plus `+<12-character commit>`, for example `0.2.0+4769d92b0a8f`.

Updates use `POST /api/v1/packages/update` with `{"id":"…","confirm":true,"install":{"Build":true,"Ref":"…"}}`; without `Source` the bound source is built.

Builds run as background tasks: the request returns 202 with `{"task_id":"…"}` immediately. `GET /api/v1/tasks/{id}` returns `{"id","action":"install|update","status":"running|succeeded|failed","created_at","finished_at","log","result"}`; `log` holds each step and the complete standard error of the go command, and on success `result` equals the synchronous endpoint's result. On failure `http_status` and `error` match the synchronous error response, including 409 with `{"affected":[…]}` when confirmation is required; resubmit with `Confirm: true`, the build cache makes the rebuild faster. `GET /api/v1/tasks` lists tasks without `log`, and `DELETE /api/v1/tasks/{id}` cancels a running task or removes a finished record. Only one build runs at a time and queued tasks log that they are waiting. Tasks live only in memory, disappear when Core restarts, and running builds are cancelled on shutdown.

`GET /api/v1/core/build-environment` returns `{"go_version":"go1.27.1","toolchain_bytes":0,"cache_bytes":0}`, with an empty `go_version` before the toolchain is downloaded. `DELETE /api/v1/core/build-environment?part=toolchain|cache` removes the toolchain or the cache, both when `part` is omitted, and returns the new usage; it returns 412 while a build runs. The toolchain takes about 250 MiB and the cache grows with builds.

Ordinary plugins that only depend on the AlonaBot SDK build lightly (a few hundred MiB and seconds in testing); components with large dependencies, such as the Agent with LLM SDKs, need about **2 GiB of available memory**, after accounting for Core and other workloads. Core compiles one package at a time and logs a warning when available memory is low on Linux; the build continues. On smaller hosts, install a Release or run `alonabot pack --target <os>/<arch>` on another machine and upload the archive for Core's platform.

If the task log warns that dependency downloads may be blocked, use System Settings to set `goproxy` to a reachable mirror such as `https://goproxy.cn,direct`, or set `proxy` to a network proxy and restart Core. `goproxy` changes apply to the next build. See [Go download and proxy settings](#core-settings-and-http-entry-point). The warning does not cancel the build.

Use `alonabot tune list` to inspect the startup tuning settings: `BUILD_MEMORY_SOFT_LIMIT` (768 MiB per Go process), `BUILD_GOGC` (50), `BUILD_MEMORY_WARN_BELOW` (1536 MiB, Linux only) and `BUILD_STALL_INTERVAL` (60s). These settings apply only to Core's managed toolchain.

Source builds support neither cgo nor private repositories and modules: Core sends no credentials to GitHub or module proxies and has no `GOPRIVATE`/`GONOSUMDB` settings. The image has no git, so the `direct` fallback in `goproxy` does not work inside the container and dependencies must be available through a proxy. The module must sit at the repository root and the source must not contain symbolic links. Windows hosts and Core binaries built with a development Go version do not support source builds yet.

### Export diagnostics
```fish
alonabot diagnose --data-dir ./data --since 24h -o diagnose.zip
```

Defaults collect 24 hours of logs, current settings/Active configuration, component/environment information, with redaction enabled. `--no-logs`, `--no-settings`, `--no-components` each omit a group; retain at least one. `-o` also accepts `--output`; otherwise create `alonabot-diagnostics-<UTC time>.zip` in the current directory.

Online export uses management API/Bearer token, automatically reading `<data_dir>/admin_token`, connecting to local listen and preserving base_url Host; `--admin-url` overrides the connection. Connection refusal triggers an offline attempt requiring unheld core.lock. Offline export reads the original directory without modifying it.

Without a usable Adapter instance schema, redact all leaf values in that instance configuration, preserving structure and full `${ENV_VAR}` expressions. `--no-redact` exports plaintext credentials. Logs may contain chat content; inspect the archive before attaching it to an issue.

## Core settings and HTTP entry point

After acquiring core.lock, `run` reads `<data_dir>/core_config.json`; all Core settings come from this JSON object. A missing file is created with defaults and mode 0600. JSON syntax, duplicate keys and invalid roots report key paths/line/column with usage exit status. Unknown keys remain and are ignored; invalid types/values are corrected in the runtime copy with paths/reasons recorded.

Settings and defaults are defined in [`internal/core/settings`](../internal/core/settings). `base_url` defaults to `http://127.0.0.1:7650`, shared by management and plugin routes. `admins`, `command_prefixes`, `reply`, `default_locale`, `log.sources` and `adapters` support parsing/saving. `command_prefixes_by_instance` is preserved as an unknown key and ignored until administrator prune removes it. For per-account prefixes, assign an instance's default profile and save prefix overrides through the platform API. The old `/api/v1/core/command-prefixes` route was removed. Local `run` assembles listen/logging/proxy/Base URL settings.

A complete string `${NAME}` expands its environment variable when read. NAME contains `A-Z`, `a-z`, digits/underscores, beginning with a letter/underscore. Missing variables make the value invalid; it falls back to its default and records a correction.

`(*settings.File).Update` edits the unexpanded document, validates it and replaces it via a mode-0600 temporary file, preserving other key order and `${...}` expressions, appending new keys at their object's end. Writeback uses two spaces and a trailing newline. Files have no format version: invalid JSON objects prevent startup; invalid values are corrected in memory without rewriting the file.

SDK runs `OnConfigLoad(ctx, raw)` before filling/correcting plugin values. `GET /api/v1/plugins/{id}/config` returns Active `corrections`/`unknown`. `POST /api/v1/plugins/{id}/config/prune` removes invalid Active keys into Pending; management reload then applies them. Development components follow the same order: reload atomically updates their editable file, clears Pending after successful startup, and subsequent restart preserves cleanup.

Read Core corrections through `GET /api/v1/core/settings/corrections`; `POST /api/v1/core/settings/prune` writes cleanup. Adapter-instance GET also returns `corrections`/`unknown`; `POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune` removes only that instance's invalid keys while preserving other values and `${...}` text. Correction/required validation runs per instance; failed instances keep editable settings while others start. See [management API](#management-api) for application behavior.

Configuration preparation uses `internal/tune`'s `PHASE_CONFIG` (default 60s) for OnConfigLoad, filling/correction and required checks. `PHASE_MIGRATION` (default 15m) covers DB migrations; see [`internal/tune/params.go`](../internal/tune/params.go).

In deployed runs, restart Core after manually editing settings. [Development runs](development.md#development-runs-and-hot-reload) refresh settings/development instances on change, but startup-assembled settings such as listeners still require Core restart. Explicit CLI `--listen`, `--log-level`, `--log-format` override settings for that run only, leaving the file unchanged:
```fish
alonabot run --data-dir ./data --listen 127.0.0.1:8080
```

Use `alonabot logs --level critical` for serious faults only, or `alonabot run --log-level critical` to raise this run's global threshold. During runtime, `alonabot log-level global critical` saves the global threshold; `alonabot log-level <component ID> critical` sets a source. See [management API](#management-api) for filtering and output settings.

Core exposes one public HTTP entry point at `listen` (default `127.0.0.1:7650`). Request Host hostname must match `base_url`: `/plugins/{id}/...` forwards to current Session routes, other paths go to management, other Hosts return 421. With defaults, lifecycle's private page is `http://127.0.0.1:7650/plugins/demo/page`. A nonloopback listener requires https `base_url` or nonempty `trusted_proxies` terminating TLS; both settings and `--listen` are validated.

All plugin-route responses, including Core errors, set `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` and `X-Content-Type-Options: nosniff`, replacing plugin values. Core removes all plugin `Set-Cookie`/`Access-Control-Allow-Credentials` including trailers, logging only cookie names at WARN. Missing Content-Type becomes `application/octet-stream`. HTML uses this sandbox policy.

A configured `proxy` becomes Core/child `HTTP_PROXY`/`HTTPS_PROXY`; otherwise existing `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` pass through unchanged.

`go_download_base` (default `https://go.dev/dl/`, for example `https://golang.google.cn/dl/`) and `goproxy` (default `https://proxy.golang.org,direct`) are used only by [source builds](#source-builds). `go_download_base` must be an HTTPS URL ending with `/`. Changes apply to subsequent downloads without restarting. `GET`/`PUT /api/v1/core/downloads` reads and writes both settings and the network `proxy`; changing `proxy` requires restarting Core. Configure GitHub API and archive addresses separately under [Component download mirrors](#component-download-mirrors).

First startup persists `<data_dir>/admin_token` with mode 0600 and prints only management address/file path. Local WebUI displays the pending token; copy/save it and click “I've saved it”. Remote browsers can read it explicitly:
```fish
alonabot admin-token show --data-dir ./data
```

`show` reads persisted credentials directly and works while Core runs. Normal login uses `POST /api/auth/login`, requiring controlled Origin and JSON token; success returns management Cookie/CSRF. Login/confirmation preserve the file.

Private plugin routes use valid management Cookie or Bearer token. Cookie-authenticated methods other than GET/HEAD also need `X-Alonabot-CSRF`/controlled Origin; failure returns 401. Public routes use plugin-declared callback source validation without management sessions. Before forwarding, Core strips management Cookie, Authorization and CSRF headers.

Before startup, run `alonabot admin-token init --data-dir ./data` if needed. For lost credentials explicitly use `admin-token reset`, invalidating old credentials/sessions and making the new generation pending confirmation. Both init/reset print file locations, require no TTY and require stopped Core.

With Core running and the current token, `POST /api/auth/token/rotate` rotates online. Response/persistent file update together, all old sessions revoke and the new generation awaits confirmation. CLI reads the updated file automatically.

Tuning parameters parse at startup; restart after changes. `default` prefixes/`reply` belong to `core_config.json`. Profile override collections use the fixed 4 MiB frame cap; `core.profile.*`/`sdk.profile.set_overrides` use ordinary RPC deadlines/Session budgets ([`internal/tune/params.go`](../internal/tune/params.go)). List all parameters/effective values:
```fish
alonabot tune list
alonabot --args REQUEST_TIMEOUT=2s tune list
alonabot run --data-dir ./data --args REQUEST_TIMEOUT=2s --args ./tuning.txt
```

`--args` repeats before/after subcommands; later values override earlier ones. Inline arguments support comma-separated `KEY=VALUE`; files use one per line, ignoring blank lines and `#` comments. Durations use Go forms such as `2s`; sizes use `B`, `KiB`, `MiB`, `GiB`, rejecting decimal units such as `MB`. `run` logs final overrides/sources at WARN.

See `alonabot tune list` and [`internal/tune/params.go`](../internal/tune/params.go) for `OPTIONS_PAGE` defaults/range. Chat-directory page caps are ten times the effective Core value; requested `limit` still defaults to 50.

## Plugin hibernation

Installed plugin/agent components resume their original generation from SUSPENDED for commands/events, Service and HTTP requests. Core's sole Waker delivers only after ACTIVATED, digest verification and complete binding, retaining admission's absolute deadline while waiting. Same-generation RESUME reuses configuration/DB, skips migrations, but still loads/fills/corrects/validates required configuration and requires results identical to the snapshot. Development components retain current bindings/lifecycle.

`internal/tune` registers `WAKE_QUEUE_REQUESTS=256` (≥1), `WAKE_QUEUE_BYTES=16MiB` (4MiB–`CORE_MAX_PAYLOAD`); queued payloads also consume Core's total budget. Override with `--args` using the same name/unit/range/source validation. HTTP waiting consumes count only; bodies stream after delivery. Wake failures map to 429 (budget full), 504 (deadline) or 503 (other). Idle defaults: `sleep.mode=always_on`, `sleep.idle_timeout=15m`, empty `sleep.plugins`. Overrides accept `never` or positive durations; removal restores automatic admission. `SLEEP_MIN_UPTIME=5m` (>0), `SLEEP_COLD_START_WINDOW=1h` (>0), `SLEEP_COLD_START_MAX=3` (≥1) are registered. Exceeding the cold-start cap keeps a plugin resident for one window, retaining the same cooldown deadline across starts. Scheduler expiry uses the same Waker/budgets, counting complete encoded HandlerRequest bytes. Static deadlines are original scheduled_at plus `JOB_TIMEOUT=15m` (>0); dynamic deadlines start on acquiring a concurrency slot, including cold start. Waiting retains its run slot and `TASK_CONCURRENCY=16` slot (≥1, ≤ SESSION_MAX_INFLIGHT).

The management CLI reads/changes per-plugin policy using `--data-dir`/`--admin-url`:
```fish
alonabot sleep-policy
alonabot sleep-policy my-plugin
alonabot sleep-policy my-plugin never
alonabot sleep-policy my-plugin 15m
alonabot sleep-policy my-plugin auto
```

No arguments returns global settings and every installed plugin/agent's policy/runtime; one ID returns that component. `auto` removes overrides; `never`/positive durations persist/apply immediately. `PUT /api/v1/core/sleep` changes global mode/timeout, both required. Management reads preserve hibernation. Policy queries needing routes without a valid committed snapshot return 412 / `FAILED_PRECONDITION`; global settings/runtime are independently readable. Resume failures after saving retain actual errors; GET again to verify saved settings.

See [architecture](../ARCHITECTURE.md) for wake/scheduler implementation/test entry points and [development](development.md#verify-changes) for verification.

## Backup and recovery

Stop Core and every plugin process before copying the entire `data_dir`. Backups must include Core metadata, packages, plugin configuration and every generation's managed SQLite DB. `state.sqlite` alone holds Core metadata without plugin business data.

Back up plugin-owned external databases/files outside framework directories using each plugin's instructions. For recovery keep Core/plugins stopped, restore full `data_dir` and external data together, then start Core. Restoring an old backup discards later writes; confirm the desired recovery point first.

## Management API

Use the same controlled `base_url` and authentication described above. Core API major 1 lives at `/api/v1/`; authentication lives at `/api/auth/`. JSON writes use `Content-Type: application/json`. CLI clients send `Authorization: Bearer <admin token>`; browsers use the management Cookie and send controlled `Origin` plus `X-Alonabot-CSRF` on methods other than GET/HEAD. `POST /api/auth/login` takes `{"token":"…","remember":true}`, returns `{"csrf_token":"…"}` and sets the Cookie: with `remember` the session lasts 30 days and the Cookie persists; otherwise the session lasts 8 hours and the Cookie ends with the browser session. `GET /api/auth/session` restores CSRF after refresh; `POST /api/auth/logout` revokes the session. Login and confirmation keep the persisted token file.

Management errors use `{"code":"FAILED_PRECONDITION","safe_message":"…"}`. HTTP mappings are 400 INVALID_ARGUMENT, 403 PERMISSION_DENIED, 404 NOT_FOUND, 409 ALREADY_EXISTS or stale configuration, 412 FAILED_PRECONDITION, 501 UNSUPPORTED, 503 UNAVAILABLE and 500 INTERNAL/other codes. Authentication errors are plain text. Unknown management routes authenticate first (401 without credentials, JSON 404 afterward). Confirmation-required operations return 409 with `{"affected":["…"]}`: display every affected object and resend the same request after administrator approval using `Confirm: true` for install or `confirm: true` for actions/instance creation. Check actual status and Content-Type before parsing.

| Operation | Endpoint and request |
| --- | --- |
| Core identity and components | `GET /api/v1/core`, `GET /api/v1/packages` |
| Install | `POST /api/v1/packages/install`, using exported `InstallRequest` fields (`Archive`, `Source`, `ID`, `Confirm`, `Replace`, `Build`, `Ref`); see [management types](../internal/core/manage/install.go). `Build` returns 202 with a task ID, see [Source builds](#source-builds) |
| Background tasks | `GET /api/v1/tasks`, `GET`/`DELETE /api/v1/tasks/{id}` |
| Build environment | `GET`/`DELETE /api/v1/core/build-environment`; DELETE accepts `?part=toolchain\|cache` |
| Go downloads and network proxy | `GET`/`PUT /api/v1/core/downloads`, `{"go_download_base":"https://go.dev/dl/","goproxy":"https://proxy.golang.org,direct","proxy":""}`; PUT updates only the fields present, invalid values return 400 |
| Enable, disable, restart, reload | `POST /api/v1/packages/{action}` with `{"id":"demo"}`; disable may require `confirm` |
| Update | `POST /api/v1/packages/update`, `{"id":"demo","confirm":false,"install":{"Source":"owner/repo"}}` |
| Rollback | `POST /api/v1/packages/rollback`, `{"id":"demo","generation":"retained-generation","confirm":false}` |
| Pin | `POST /api/v1/packages/pin`, `{"id":"demo","constraint":"1.2.3"}`; empty constraint removes pin |
| Uninstall | `POST /api/v1/packages/uninstall`, `{"id":"demo","purge":0,"confirm":false}`; preview affected objects before purge |
| Active configuration | `GET /api/v1/plugins/{id}/config`; inspect `corrections` and `unknown` |
| Edit Pending | `PUT`/`PATCH /api/v1/plugins/{id}/config/pending` with `base_generation` and full `config` or RFC 7396 `patch`; `DELETE` discards Pending |
| Import and apply | `POST /api/v1/plugins/{id}/config/import` reads editable config.json into Pending; `POST /api/v1/plugins/{id}/reload` prepares/commits it |
| Adapter instances | `GET /api/v1/adapter/instances`; `POST /api/v1/adapter/instances/{create,start,stop,delete}`; create uses `package_id`, `instance_id`, `config`, optional `confirm`, others use `instance_id`; each listed instance has `package_state`: `enabled`, `disabled` or `not_installed` (not installed or uninstalled, configuration retained) for its Adapter component |
| Instance configuration | `GET /api/v1/adapters/{adapter_id}/instance-schema`; `GET`/`PUT`/`DELETE /api/v1/adapters/{adapter_id}/instances/{instance_id}`; PUT uses `{"config":{}}` |
| Logs | `GET /api/v1/logs/recent`, `/stream` (SSE), `/history?date=YYYY-MM-DD`; optional `source`, `level`, `trace`; `GET`/`PUT /api/v1/logs/levels` uses `{"global":"info","sources":{}}` partial updates |
| Profiles | `GET`/`POST /api/v1/profiles`; `PATCH`/`DELETE /api/v1/profiles/{id}`; assignments, plugin, platform and tool overrides are in [profile routes](../internal/core/manage/profile_api.go) |
| Tool policies | `GET`/`PUT /api/v1/tools/policies`; write `{"plugin_id":"demo","service":"echo","policy":"deny"}` or atomic `{"items":[…]}`; allow/confirm/deny/inherit, with inherit removing the override |
| Conversations | `GET /api/v1/plugins/{id}/conversations` returns summaries; `DELETE /api/v1/plugins/{id}/conversations/{key}` clears one full opaque key |
| Plugin data | `/api/v1/plugins/{id}/data/`: tables/rows, kv, files, export and clear; see [data handlers](../internal/core/manage/data_api.go) |
| Diagnostics | `POST /api/v1/diagnostics`, `{"since":"24h","logs":true,"settings":true,"components":true,"redact":true}` defaults; returns ZIP with no-store |

### Profile session access

Set session access under the profile's Platform and Messages settings. Core selects the session's assigned profile, the instance default, or `default`, in that order. `all` (the default) allows every session. `allow` admits only listed sessions; an empty list denies all. `deny` blocks only listed sessions; an empty list allows all. Blocked sessions skip command, Agent, and plugin event handling, and Core rejects sends to them. Existing global deny rules continue to apply at their usual checks; either policy can reject an operation.

Use `GET`/`PUT /api/v1/profiles/{id}/session-access` to read or replace the mode and list, for example `{"mode":"allow","sessions":[{"instance":"main","session":"telegram:-100123#42"}]}`. Entries pair an instance with a full session ID; chats and topics match separately. Submit up to 1,000 entries; the first occurrence of each pair is retained. Omitting `sessions` or passing an empty array clears the list. In `all` mode Core saves the list without applying it.

### Other management conventions

Local JSON installation provides an `Archive` path on the Core host; identity and source come from the archive’s `plugin.yaml`. Online installation provides only `Source: "owner/repo"`; Core selects a platform archive through the latest Release API. Optional `ID` constrains the expected identity. First source trust and same-kind replacement require confirmation when applicable. Updates and reinstalls with retained records preserve source and identity binding. `Package.PublisherKeyID` and `Package.Signed` retain their historical meaning; new installs and successful updates or reinstalls set them to empty and false.

Browsers send `multipart/form-data` to the install/update endpoints: one required file part `archive`, optional text `confirm=true|false`, install-only text `replace=true|false`, and update-only required text `id`. File parts have `filename`; text parts omit it. Unknown or duplicate parts, missing archives/update IDs, invalid booleans and truncated requests return 400 / INVALID_ARGUMENT. Authentication precedes parsing. Archives stream into exclusive temporary directories under existing installation limits and preflight checks; cleanup follows the service call. Confirmation returns 409 with all `affected` objects; resend the archive with `confirm=true` after approval. Preflight precedes changes to existing components. Source binding failures return 403 / PERMISSION_DENIED; updates retain identity, pin and dependency checks and preserve user data.

Pending edits require the current `base_generation`; stale edits return 409. Secrets are masked as `{"$secret_set":true}` on reads; submit new plaintext only where intended, never the mask or Core secret references. Reload creates a candidate generation with its own managed SQLite; restart keeps Active. A failed OnReady after commit leaves the new Active degraded rather than restoring the old database. Installed generation lists are available at `GET /api/v1/plugins/{id}/generations`. Development components reject update/rollback/uninstall with 412.

Conversation keys use `<session>@<instance>`, with `@<user>` appended for per-member sessions. Percent-encode the complete key as one path segment; it is decoded once, preserving slash/colon/backslash/dot segments. Keys require 1–256 UTF-8 bytes. Data browsing can select retained generations read-only. `/data/clear` requires `{"confirm":true}` and deletes writable rows (including KV) and plugin files while preserving schemas/migrations/directories. KV keys follow [storage rules](plugin-guide.md#store-small-amounts-of-state); encode them as opaque segments. File paths stay relative, disallow traversal/symlinks; clearing or editing files is independent of generation rollback.

For hibernation use `GET`/`PUT /api/v1/core/sleep` and `GET`/`PUT`/`DELETE /api/v1/plugins/{id}/sleep-policy`, plus `GET /api/v1/plugins/{id}/runtime`. PUT per-plugin policy is `{"policy":"never"}` or a positive duration; DELETE restores automatic policy. Reads keep plugins asleep. See [hibernation](#plugin-hibernation) for defaults, budgets and failure behavior.

### Changelog and WebUI backend

Authenticated `GET /api/v1/core/changelog` returns `{"locale":"en-US","entries":[{"version":"0.1.0","markdown":"…"}]}`. It uses released `changelog/v<SemVer>/{en,zh-CN}.md`, excludes unreleased, maps en to en-US and selects exact or primary-language `Accept-Language` matches; absent/unmatched headers fall back to English. Entries sort by SemVer descending; versions omit `v`, optional `date` is omitted when unknown, and Markdown is preserved. `GET /api/v1/packages/{id}/changelog` reads the installed package: single-file CHANGELOG.md sections omit Unreleased and use en-US; directory form selects per version by preferred exact/canonical/primary-language match then en-US.md, with lexicographic filename tie-breaking. Uninstalled or dev components return 404; no released entries returns en-US/empty entries.

Component icons use authenticated `GET /api/v1/packages/{id}/icon`. Referenced configuration images use `GET /api/v1/packages/{id}/examples/{path}` from the Active package only. Package `Name`, `Description` and `Author` come from the packaged `plugin.yaml` and are omitted when unreadable. Package `pages` lists each component page's `path` and `title`, with `purpose: "instance_config"` for account configuration pages. WebUI opens pages under `/plugins/{id}` in its sandbox and opens account configuration pages from the add/edit account dialog. See the [page kit](plugin-guide.md#plugin-pages) and [Adapter account configuration pages](plugin-guide.md#adapter-account-configuration-pages).

The official WebUI is a real component with its own backend, not just static `dist` files. Its authenticated prefix `/plugins/alonabot-webui-admin` includes `/stats?range=24h|7d|30d`, `/users`, `/store`, and `/changelog/seen?component=core`. `PUT /changelog/seen` accepts exactly `{"component":"core","version":"0.1.0"}` and monotonically advances persisted SemVer; GET returns `{"version":null}` or a version. Copying only `dist` leaves these backend routes unavailable. Agent model/persona/MCP/skill/memory/task pages similarly belong to the Agent; see its [HTTP implementation](https://github.com/AlonaBot/alonabot-agent/blob/main/pages.go).

The [management package](../internal/core/manage) owns the complete current route registrations and request/response types; [management integration tests](../tests/e2e/manage_e2e_test.go) exercise them through real Core.

### Component store

The official WebUI stores its data in its own managed SQLite database. Backend routes below are relative to `/plugins/alonabot-webui-admin`, use management authentication/CSRF and return `Cache-Control: no-store`. Component installation and updates still use Core's management APIs.

Use `GET /store?source=AlonaBot/plugin-store&mirror=auto&role=plugin&q=&sort=stars&offset=0&limit=9&refresh=false`. The store reads `store.json` from the source repository's `index` branch, a snapshot generated by that repository's CI (see the [plugin-store README](https://github.com/AlonaBot/plugin-store#readme)); it makes no per-entry GitHub API calls and downloads no archives. `mirror` selects the route: `auto` (default) requests `https://raw.githubusercontent.com/<source>/index/store.json` and `https://cdn.jsdelivr.net/gh/<source>@index/store.json` concurrently and uses the first success; `github` or `jsdelivr` uses only that address; an https URL prefix is prepended to the raw address (GitHub accelerator style); other values return 400 / INVALID_ARGUMENT. Entries whose repo or role is invalid are dropped and icons other than `data:image/` are removed. `source` must be `owner/repo`; specify nonnegative `offset` and positive `limit`. Role is empty or plugin/adapter/agent/webui. The response contains `source`, `fetched_at`, `stale`, `items`, `total` and `counts` (`all` and each role). Items contain `repo,role,tags,id,name,version` plus optional `description,author,stars,icon_url`; missing stars means unknown, while zero remains zero. The default source is `AlonaBot/plugin-store`; browser-persisted source lists retain an intentional empty list.

Search is a case-insensitive substring match over name, description, repo, author and space-joined tags; missing values are empty. Counts are calculated after search and before role filtering; total is calculated after role filtering. Default/empty `sort=stars` orders known star counts descending before unknown values; `sort=name` orders names case-insensitively. Ties preserve index order; other sort values return 400 / INVALID_ARGUMENT. WebUI shows nine cards per page and returns to page one after search, role or sort changes.

`refresh` accepts only literal `true` or `false`; even an empty value returns 400 / INVALID_ARGUMENT. Omitted/true refreshes and persists a successful full snapshot. With false, an existing successful snapshot is searched, counted, sorted and paginated without remote requests; no snapshot triggers the first fetch. Snapshot reads return `stale=false` and the original `fetched_at`. Entering the store, changing sources and keyboard-accessible manual refresh use true; search/filter/sort/pagination use false. A failed refresh retains the successful snapshot and original timestamp, returning `stale=true` plus the full error. Without a snapshot, ordinary failures return 500 / INTERNAL; a source without `store.json` returns 404 / SOURCE_INDEX_MISSING.

Version checks and Release previews for installation and updates use Core's `GET /api/v1/core/github/release?source=owner/repo`, returning `repo,version,release_url,markdown,fetched_at`. The lookup shares installation's GitHub mirror settings, HTTP client and original repository/tag validation. Failures return complete redacted errors without a stale snapshot. A latest-Release 404 returns 404 / NOT_FOUND, mapped to null by WebUI's `getLatestRelease`. Links allow only HTTP/HTTPS/mailto; HTML renders as text. Update previews use the component's bound source, installed changelogs supply post-update notes, and seen versions advance only to versions actually displayed.

For submission, `GET /store/check?repo=owner/repo` reads the latest Release API and its first platform archive without changing persistent caches. It checks repository, Release and asset URLs, tag SemVer, asset name/length, and packaged manifest ID/version/source. It returns `repo,role,id,name,version,release_url`, optional description/author and ordered manifest `warnings` (including an empty array). The returned repo preserves the verified manifest Source’s canonical case after removing `github.com/`. Core preflight also checks platform, build-info, resources and archive integrity. Only latest-Release 404 maps to NO_RELEASE; bad asset 404s and other failures return 500 / INTERNAL with the actual error.

A source index uses `{"version":1,"plugins":[{"repo":"owner/repo","role":"plugin","tags":["utility"]}]}`; tags are optional. After checking, the user supplies tags and opens the source repository's `issues/new` with `template=submit-plugin.yml` and URL-encoded `repo`, `role`, `tags`, `release_url`, using the canonical checked repo. The browser submits the GitHub Issue.

Each WebUI store request to a remote server has a 20-second deadline covering the response body; serial requests may take longer in total, and refresh follows caller cancellation. `/store/check` reads the GitHub API and Release archives. It checks HTTPS hosts and paths on the initial URL and each redirect. Safety limits are: store snapshot 32 MiB (including inline icons), metadata 1 MiB, archive 64 MiB, expanded data 256 MiB per scan (at most two scans), one file 128 MiB, 2,048 entries per scan, retained manifest 256 KiB and retained icon 1 MiB. Exceeding limits preserves the `exceeds ... bytes/entries` error and follows the same cache rules. The current implementation lives in [WebUI store.go](https://github.com/AlonaBot/alonabot-webui/blob/main/internal/admin/store.go).

## Component download mirrors

Set `github_api_base` and `github_download_prefix` independently in Core's `core_config.json`, or use **Settings → Component downloads** in WebUI. Saved values apply to the next install or update. Both default to an empty string: the API uses `https://api.github.com` and archives use the original asset URL.

```json
{
  "github_api_base": "https://api-mirror.example",
  "github_download_prefix": "https://proxy.example/"
}
```

The archive address is exactly **prefix + original URL**, for example `https://proxy.example/https://github.com/owner/repo/releases/download/v1.0.0/package.tar.gz`. Include any required trailing slash in the prefix. Both settings and redirects require HTTPS. Omit URL credentials and fragments. The API base may include a path and must omit query parameters. A download prefix may use a query form such as `https://proxy.example/?url=`.

For Release installs and updates, Core verifies the archive against the SHA-256 digest returned by the Release API. `github_api_base` is the trust root for that digest; use only a trusted GitHub API reverse proxy. When an older asset has no digest, clear the download prefix and retry. Direct downloads of those assets verify size and log a warning. For source tarballs, see [Source builds](#source-builds).

Override either field for one command; overrides leave the saved settings unchanged:

```fish
alonabot install --source AlonaBot/alonabot-agent --github-api-base https://api-mirror.example --github-download-prefix https://proxy.example/ --yes
alonabot update alonabot-agent-official --github-api-base https://api-mirror.example --github-download-prefix https://proxy.example/ --yes
# Download directly this time while retaining the saved API setting
alonabot update alonabot-agent-official --github-download-prefix '' --yes
```

Authenticated `GET/PUT /api/v1/core/github` reads/writes the persistent settings; PUT requires both strings. Install requests accept optional `github_api_base` and `github_download_prefix`; update requests place them inside `install`. Omission inherits settings; an explicit empty string restores the corresponding official address. Core validates the original GitHub repository, release tag and asset address before transforming request URLs, then checks the archive digest and the package's identity, platform, version and build-info. Errors and diagnostic settings omit the download-prefix query, and cross-host redirects remove `Referer`. Downloads have no automatic retry or fallback.

`INSTALL_DOWNLOAD_TIMEOUT` defaults to five minutes per GitHub request, including redirects and body reads. `INSTALL_MAX_RELEASE_JSON` defaults to 4 MiB for GitHub API responses, including source ref resolution. `INSTALL_MAX_ARCHIVE` caps both Release assets and source tarballs. `INSTALL_MAX_HTTP_ERROR` keeps the first 4 KiB of non-200 response bodies by default. Run `alonabot tune list` to see the effective values.

Store discovery and submission checks still use the selected source's GitHub index and display snapshot under the separate fetch rules above; first-time store browsing requires a reachable index source. Direct repository installation, installed-component update checks and confirmation previews call Core directly; installation and updates then use the same mirror settings.
