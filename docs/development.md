# Source development

[简体中文](development.zh-CN.md)

Run a plugin from source, then work in Core, SDK, Adapter, Agent or WebUI as needed. See [contributing](../CONTRIBUTING.md) for workflow and [architecture](../ARCHITECTURE.md) for modules/data ownership. This guide maintains environments, commands and verification; operations are in [deployment](deployment.md).

## Prepare the environment

| Tool | Version | Needed for |
| --- | --- | --- |
| Go | 1.27.1; see `go.mod` | Go development/component builds |
| protoc | 36.0, printing `libprotoc 36.0` | Regeneration after `.proto` changes |
| protoc-gen-go | v1.36.12, managed by `go.mod`'s `tool` directive | Protocol generation; no global installation |
| Node | 26.5.1; see WebUI `package.json` | Official WebUI frontend development |
| npm | Bundled with Node | Installing dependencies from the frontend lockfile |

macOS development uses asdf for Go/Node; other platforms also need matching versions. Documentation-only changes need no complete runtime setup.

```fish
git clone https://github.com/AlonaBot/AlonaBot.git
cd AlonaBot
go mod download
go build -o bin/alonabot ./cmd/alonabot
```

Existing examples reuse the repository's module/resources without an external project or model credentials:

```fish
bin/alonabot run ./examples/plugins/lifecycle --data-dir ./data
```

Enter `/ping` in the same terminal for `pong`; Ctrl-C stops normally. Parallel Core instances need separate data directories/listener ports. See [deployment](deployment.md#core-settings-and-http-entry-point) for first authentication/listener settings.

The example marks `/ping` with `.Menu()` for automatic chat menus on supported Adapters; unmarked commands remain callable. See [command menus](plugin-guide.md#command-menus) for manual control.

## Dependency management

Core uses versions locked in `go.mod`/`go.sum`; official components maintain their own modules/locks, and WebUI also has `package-lock.json`. Before adding dependencies or expanding linkage, explain necessity/alternatives and await maintainer approval. Approved uses of direct dependencies:

| Module | Approved use |
| --- | --- |
| protobuf, x/sys, google/jsonschema-go | Protocol, OS facilities, Schema; SDK root also links Goldmark through the message model |
| github.com/yuin/goldmark v1.7.16 | Approved CommonMark/strikethrough parsing in `contracts/markdown`, shared by Core send/edit and SDK `MarkdownSegment.PlainText()` |
| modernc.org/sqlite, libtnb/sqlite, GORM | Core state database and optionally imported SQLite/GORM SDK packages |
| go.yaml.in/yaml/v3 | Manifests, skill metadata and other YAML |
| Cobra, pflag, go-isatty; mousetrap on Windows | CLI/terminal detection |
| modelcontextprotocol/go-sdk | Official Agent MCP client |
| openai-go/v3, anthropic-sdk-go/v2 | Agent Providers; no vendor subpackages such as Azure/Bedrock/Vertex |
| coder/websocket | OneBot 11 Adapter; Agent's transitive usage remains under the current lockfile |

Core executables may link modules needed for protocol, Goldmark, SQLite/GORM, yaml, x/sys, Schema and CLI. Model SDKs, MCP and platform clients belong to their components. Plugins link GORM only when importing `sdk/gormdb`. Inspect actual linkage during review:

```fish
go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' ./cmd/alonabot ./sdk | sort -u
```

Frontend locks React, Vite, TypeScript, vitest, Lucide, dnd-kit and Radix. Recharts is approved but absent from the current lockfile. DOM testing environments/Testing Library require separate approval. JetBrains Mono static fonts are distributed under SIL OFL 1.1.

### GORM source

GORM uses upstream `gorm.io/gorm v1.31.2`, locked by each module's `go.mod`.

Separate-component build/pack/development builds use the author's main module go.mod/go.sum; library mode uses the author module directly. Release GORM components' build-info lists actual module versions.

## Official component development

Component source and Go/npm locks live in their own repositories:

| Component | Development entry point |
| --- | --- |
| WebUI | [Source development](https://github.com/AlonaBot/alonabot-webui/blob/main/docs/development.md) |
| Agent | [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) |
| Telegram | [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) |
| OneBot 11 | [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) |
| Help | [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) |

Run `alonabot check .`, `alonabot run .` or `alonabot pack .` in a component project. Core/SDK retain an independent module; `components.lock.json` pins component source for bundled builds. Cross-component tests use separate `tests/go.mod`, keeping model SDK/Adapter implementation dependencies there.

```fish
uv run --no-project python scripts/integration.py
# 按改动选择测试，例如 Agent 服务：
uv run --no-project python scripts/integration.py --run '^TestAgentServices$' --race
```

The integration entry point first runs `tests/agent`/`tests/e2e`. Agent service tests needing private Core assembly run via temporary modfiles/overlays using the Agent version resolved by `tests/go.mod`, without writing dependencies/tests back into Core's module. Test dependencies may pin real development commits; `components.lock.json` continues to pin published components for release builds. CLI end-to-end tests live in `tests/e2e`; `cmd/alonabot` contains production commands only.

## Code generation

Edit only `.proto`, generator or changelog sources, then regenerate:

```fish
go generate ./...
```

Protocol sources are in `contracts/proto/alonabot/v1/`, all using `go_package` `github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1`. Generation checks protoc version, regenerates `contracts/pb/`, and locates protoc-gen-go through `go tool`.

Record daily changes in `changelog/unreleased/{en,zh-CN}.md`; during release preparation, organize them into `changelog/v<SemVer>/{en,zh-CN}.md` using the [changelog format](../changelog/README.md). The generator concatenates released files unchanged per language into `internal/core/manage/changelog_core.go`; the API reads versions and dates from their headings. After editing released notes, run `go generate ./internal/core/manage` and deliver both source notes and generated content.

## Verify changes

Choose checks by impact; reuse valid results without repeating unrelated checks:

```fish
git diff --check
# 例如修改 Router 时：
go vet ./internal/core/router
go test -race ./internal/core/router
```

Premerge Go CI runs `go vet ./...`, race tests and five-target pure-Go builds. GitHub-hosted runners serialize test packages; [workflow](../.github/workflows/ci.yml) commands are authoritative. WebUI's separate CI also type-checks/builds frontend. For a full manual check:

```fish
go build ./...
go vet ./...
go test -race ./...
```

For page-kit changes run `go test -race -v ./sdk/page ./examples/plugins/showcase` and `node --check sdk/page/kit.js`, then check key interactions/language/theme in WebUI's iframe. For documentation, check links, anchors, code blocks and actual commands.

Cross-platform changes need actual runtime checks as well as cross-builds. Five targets: linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64. Record builds separately from real runs; current results are in [release notes](../changelog/v0.1.0/en.md#added). Use authorized test credentials for real Telegram/paid Provider checks; identify simulators in results.


## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) runs on main pushes, PRs targeting main and manual dispatch. Repository checks whitespace/main file scope; Go checks formatting, generation, vet, race and five-platform builds. Main path checks exclude personal maintenance records while retaining product skills and test tools.

[PR metadata](../.github/workflows/pr-metadata.yml) validates nondraft PR descriptions. Triage/auto-assign only change GitHub metadata and never execute PR code. Code checks have read-only permissions without production secrets; maintainers configure branch protection separately.

## Development runs and hot reload

Build CLI first: `go build -o bin/alonabot ./cmd/alonabot`. From the repository root, get a reply in three steps (fish):

1. Create project: `bin/alonabot init ./alonabot-plugin-hello`.
2. Run: `bin/alonabot run ./alonabot-plugin-hello`.
3. Enter `/hello` in the same terminal for `Hello AlonaBot`; Ctrl-C stops Core.

With an installed CLI, run `alonabot run` in the project directory. Console chat enables automatically when stdin is a terminal and settings contain no Adapter instances; `--console` forces it on with pipe support, `--console=false` disables it. User `console:local` has Core Admin for this run; instance/admin remain in memory. Media appear as `[image]`, `[file: name]` placeholders. Enter `#1` for the first callback button on the most recent button message; edits appear as new `(edited)` segments. Core runs after stdin ends until a stop signal.

Development directories enable reload by default. Change source and send `/hello` again to observe new replies; `--reload=false` disables it.

See the [devbot tutorial](../examples/devbot/README.md) for loopback injection, configuration, hot reload and same-process debugging.

### Project creation and SDK selection

Marketplace IDs use hyphenated prefixes such as `alonabot-plugin-hello`; local IDs still allow underscores. See [publishing](plugin-guide.md#9-publishing) for rules.

To check official help, first check out `alonabot-plugin-help` separately and run `bin/alonabot check <help project directory>`. For loopback trials, append its project to `run` directories and set injection `text` to `%2Fhelp` or `%2Fwhoami`.

Build CLI (`go build -o bin/alonabot ./cmd/alonabot`), then create/check components:

```fish
bin/alonabot init ./alonabot-plugin-hello
bin/alonabot build ./alonabot-plugin-hello
bin/alonabot check ./alonabot-plugin-hello
bin/alonabot test ./alonabot-plugin-hello
bin/alonabot pack ./alonabot-plugin-hello
```

`init`, `build`, `check`, `test`, `pack` accept at most one positional directory, defaulting to current; old `--dir` reports unknown option. `init` requires an empty directory, failing before writes otherwise. `--id` defaults to the absolute target basename: `alonabot-plugin-hello` is a marketplace ID; `my_plugin` stays a local ID with a missing-prefix warning. Invalid directory names preserve original characters in the error and request explicit `--id`. Default role is plugin; `--role agent|adapter|webui` creates others. Adapter `--adapter-id` defaults to final component ID and can be overridden. Plugin/Agent export `New()` with a schema and an English example command directly replying `Hello AlonaBot`; add i18n yourself when needed. Adapter provides an instance schema; WebUI provides `assets/index.html`. Generated Go/plugin.yaml have no comments. Templates ignore `data/`/`bin/`; overlays supply production thin entry points. See [devbot](../examples/devbot/README.md) for library hosts.

Inside an existing Go module, reuse its go.mod. External empty directories default the module to the directory name; override with `--module example.com/myplugin`. A release CLI requires SDK at its own module version and runs `go mod tidy` in the project, allowing immediate `run`. Dependency resolution failure keeps the project, exits unsuccessfully and prints manual recovery commands. Development `(devel)` builds (including local `+dirty` and automatic checkout pseudoversions) prompt `go get github.com/AlonaBot/AlonaBot@<version>` for the author to select. Change the example source in `plugin.yaml` before publishing.

### Run, transport and reload

Run `alonabot run` in a component project with `plugin.yaml`, or provide one/more directories explicitly:

```sh
bin/alonabot run ./alonabot-plugin-hello ./myadapter --data-dir ./data --reload
bin/alonabot run ./alonabot-plugin-hello ./mywebui --data-dir ./data --reload
bin/alonabot run ./examples/plugins/lifecycle --stdio
```

Without directories, current `plugin.yaml` runs that project; otherwise installed records start Core. Manifest defines identity/role; project validation/process builds finish before starting components. Duplicate directories/IDs fail, as do process-component ID/`adapter_id` conflicts with installed packages.

Builds reuse internal generation and `go build`, writing executables to `<data_dir>/dev/bin/`. Overlays supply resources/thin entry points without changing author source/module files. Plugin/Agent build their entry package; importable Adapter projects export `New()`, with generated entry point calling `sdk.RunAdapter(New())`.

Supervisor runs development components separately, using TCP loopback by default. `--stdio` switches all Native separate processes for this run: development/installed components, candidates, restarts, reloads, recovery and wakes. `alonabot run --stdio` also starts installed components. Transport selection is run-local; static WebUI/in-process mechanisms remain. Generations are `dev:<component ID>`; configuration loads each time from `<data_dir>/plugins/<id>/config.json`, treating absence as an empty object, and prepared configuration is saved in the generation snapshot. Managed SQLite `dev.sqlite` persists there across restarts. Adapter instances come from `core_config.json`'s `adapters`. Development logging defaults to debug; explicit `log.sources` wins.

Every `DEV_RELOAD_POLL`, `--reload` checks source, `//go:embed` resources, `go.mod`, `go.sum`, `plugin.yaml`, both configuration Schemas and `i18n/`, excluding data/vendor/testdata and dot/underscore-prefixed directories. Default 500ms; override with e.g. `--args DEV_RELOAD_POLL=1s`.

Source/resource changes rebuild and restart only changed components after success. Compile failures keep diagnostics and old components running. Only `config.json` changes directly restart that component; one Adapter-instance change restarts only it, preserving binary bytes. `--reload=false` builds/runs once; no development directories means ordinary Core operation without polling.

`--reload` also watches other `core_config.json` settings: `admins` immediately changes Core Admin, prefixes read shared current settings, `default_locale` affects new Sessions. Adapter `default_instance` changes immediately redirect sends without explicit instance; `log.level` immediately changes global logging, with explicit run `--log-level`/library `LogLevel` still taking priority. These preserve processes/instances. Combined settings/instance edits refresh settings first, then restart affected instances. Read failure logs errors and keeps old settings. CLI serializes refresh with management operations via `internal/app.Core.ReloadSettings(ctx)`.

`role: webui` serves the manifest's `assets` directory directly as static root, overriding installed WebUI like `alonabot.WebUI(os.DirFS(<assets directory>))`. Only one WebUI directory is accepted; duplicates fail before starting any component. Development WebUI skips Admin API major-version checks. Static files are read per request, so edits appear immediately without builds/restarts/`--reload`.

Attach debuggers to component PIDs for separate processes; use library mode below to debug Core/components together.

## Author tools

Plugin packages export `func New() *sdk.Plugin`, registering commands/lifecycle/business behavior on `sdk.NewPlugin()`. Build tools scan manifest, configuration/Service Schemas, i18n and argument names into overlays. Author projects have no `alonabot_gen.go` or generation-related author code. Build or check read-only:

```fish
bin/alonabot build ./alonabot-plugin-hello --goos linux --goarch amd64 --tags a,b
bin/alonabot check ./alonabot-plugin-hello --goos linux --goarch amd64 --tags a,b
```

`build/check` use `--goos`/`--goarch`, defaulting to Go environment; `pack` repeats `--target <os>/<arch>`, defaulting to host. All use `--tags`, and scan/dependency analysis/release compilation use `CGO_ENABLED=0`. Development BuildDev inherits development cgo. `build` outputs to `<project>/bin`, overridable with `--out DIR`. Process output includes `component` (`component.exe` on Windows), resource sidecars and packaged `plugin.yaml`, preserving author manifest. Static WebUI build publishes assets/manifest; Go-entry WebUI also builds a platform process, as official WebUI does. `check` scans entry/resources read-only, retaining all diagnostics/locations.

`test` builds a host component in a fresh temporary data directory, performs real handshake/preparation/registration/activation/shutdown, and reports commands/cleanup. Adapter has zero platform instances; WebUI checks static resource loading. Process Bootstrap supplies absolute `plugin_data/<id>/` under the temporary state root, removed at test end. Preparation/registration/stop/cleanup failure exits nonzero.

`pack` scans/checks/builds per `--target`, creating `<id>-<version>-<os>-<arch>.tar.gz`; static WebUI creates one `<id>-<version>-any.tar.gz`, Go-entry WebUI builds per platform. `pack` outputs component archives only; `--out` defaults as for build. Upload these archives to the Release; GitHub supplies source archives. Process `build-info` is `debug.BuildInfo.String()` text. Generated code/sidecars/archives share the validated scan snapshot and normalized schema; resources saved after scanning wait until the next build.

When configuration `example` fields reference images, `check`/`build`/`pack` validate both schemas' references/image signatures; referenced images enter build output/archive with the snapshot. See [reply examples](plugin-guide.md#reply-examples-for-configuration-fields).

```fish
bin/alonabot pack ./alonabot-plugin-hello --target linux/amd64 --target linux/arm64 \
  --out ./dist
```

Installation reads `plugin.yaml` inside the archive. Online installation selects platform assets through the GitHub Release API. Keep runtime configuration and data outside packaged resources.

## Library mode: one process

Root `alonabot` provides `Run`, `Start`, `*Core`, `Wait`, and options `Plugin`, `Adapter`, `WebUI`, `DataDir`, `Args`, `Listen`, `LogLevel`, `LogFormat`, forwarding to `internal/app` to run Core/development components together. See `examples/devbot`:

```go
core, err := alonabot.Start(ctx, alonabot.DataDir(dir), alonabot.Plugin(plugin.New()), alonabot.Adapter(adapter.New()))
if err != nil {
	return err
}
return core.Wait() // ctx 取消（例如 SIGTERM）后在 STOP_TOTAL 内停止全部组件
```

```fish
bin/alonabot build ./examples/devbot
examples/devbot/bin/devbot -data-dir ./data
```

Components register source=dev and generation `dev:<component ID>`. `alonabot build <main directory>` scans component entries in main dependencies (including third-party module cache), embedding resources through main's overlay. Default output is only `<main directory>/bin/<directory name>` (`.exe` on Windows), adjustable with `--out`. Component/module-cache source stays read-only; multiple New calls from one package share resources, and main package variables may call New. Plain `go build` outputs return an English prompt to use `alonabot build`/`alonabot run` before starting resource-missing components.

After library source/resource edits, rebuild with `alonabot build` and restart. Unhandled background panics exit the whole process nonzero without bounded restarts; caught SDK Handler panics log stacks/return `INTERNAL`. `Wait` errors if components exceed the stop budget. Supervisor isolates/separately recovers processes. Both modes default SIGINT/SIGTERM total shutdown to `STOP_TOTAL=10s`. Dev-source update/rollback/uninstall return `FAILED_PRECONDITION`; WebUI/CLI save development configuration and restart. See [development runs](#development-runs-and-hot-reload) for manual configuration polling.

Root `alonabot_test.go`'s `TestLibraryLifecycle` checks public library entry points; nine `internal/app/` integration-test files cover assembly/internal services/candidates/WebUI. Development CLI/hot-deployment tests cover separate-process components through `internal/app` and individual restarts; CLI end-to-end tests are in `tests/e2e/`.

## Debug a running Adapter

List instances, inspect the live operation schemas, then submit a call:

```fish
alonabot adapter instance list --data-dir ./data
alonabot adapter actions main --data-dir ./data
alonabot adapter actions main sdk/message.send --data-dir ./data
alonabot adapter call main sdk/message.send --version 1 --to '{"chatId":"private:alice"}' --params '{"chain":{"segments":[{"text":{"text":"hello"}}]}}' --data-dir ./data
```

`--params-file params.json` reads complete parameters from a file and is mutually exclusive with `--params`. `--to` uses Target protobuf JSON, including `chatId` and optional `threadId`; Core binds the Adapter and account from the selected instance. Output is JSON. Business failures exit nonzero while stdout preserves `payload`, `error` and `trace_id`, including partial receipts. Connections use the existing `--data-dir`, `--admin-url` and `admin_token`.

The SDK discovers interfaces on the real running Instance and combines them with declared capabilities: Send, Editor, Deleter, MediaSource, NodesReader, CommandSynchronizer, CommandMenuController and ActionExecutor. Standard names are `sdk/message.send`, `sdk/message.edit`, `sdk/message.delete`, `sdk/media.read`, `sdk/message.read_nodes`, `sdk/commands.sync` and `sdk/command.menu`. Platform actions retain Descriptor.Actions names, versions, scopes and parameter/result schemas. Standard parameters use the protobuf JSON shown in the catalog; platform actions accept their own JSON parameters. Query again after stopping, restarting or replacing an instance.

The public `sdk/debug.Client{URL, Token, Host, HTTP}` exposes `Instances(ctx)`, `Operations(ctx, instance, name)` and `Invoke(ctx, debug.Call{Instance, Name, Version, To, Params})`. Invoke returns both a complete Result and an error for business failures. Request contexts carry cancellation; releasing the client keeps Core running.

Endpoints are `GET /api/v1/debug/adapters/{instance}/operations` and `POST /api/v1/debug/adapters/call`, protected by existing management authentication. Send/Edit/Delete, node reads and platform actions reuse Core admission. Sending retains hooks and media authorization; menus share the instance worker and SDK menu lock. Media reads accept `fileId`, `offset`, and `length` and return base64 `data`. Use a Core-registered inbound object ID from the current instance run for `fileId`, such as `incomingRef.token` in an `sdk/message.read_nodes` result. Injected events reject inbound platform file references. Calls perform real platform operations.

Run `go test -race ./internal/core/adapter/remote -run '^TestDebugAdapter$'` to exercise the real SDK subprocess, authentication, catalog, calls, partial results, media chunks and instance replacement.

## Adapter HTTP routes

Adapters declare `[]adapter.HTTPRoute` in `adapter.Descriptor.HTTP` under `/plugins/<Adapter package ID>/...` on the shared port, such as OneBot 11 reverse WebSocket. Private routes need Core sessions; non-GET/HEAD Cookie requests need CSRF/Origin. Zero `Auth` requires Core session. Public `Auth: adapter.AuthPlatformToken` preserves incoming `Authorization` for Adapter platform-credential validation/platform error responses. Plugin routes always strip `Authorization`. Both strip incoming Cookie/CSRF/Forwarded/X-Forwarded-* and inject an internal proxy credential, constant-time checked/removed by SDK before user Handler. WebSocket Upgrade forwards bidirectionally without normal request deadlines; Core closes it on route withdrawal/Session invalidation.

## Dynamic reminders and static task pause/resume

To remind the current chat in ten minutes, register task/handler in `New()` and create its instance from the command. Full [showcase](../examples/plugins/showcase/main.go): `/demo remind meeting 600 开会` creates the current user's `meeting` reminder in this chat, `/demo cancel meeting` cancels. Same user/chat/key replaces it; different users/chats are independent. Replies preserve the user key.

```go
type ReminderArgs struct {
    Session string `json:"session"`
    Text    string `json:"text"`
}

// 创建与取消共用编码，将可信身份与用户 key 一起放入内部 key。
func reminderKey(c *sdk.Context, key string) (string, error) {
    user, session := c.User(), c.Session()
    internalKey := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, key)
    if len(internalKey) > 128 {
        return "", errcode.InvalidArgument.New("reminder key with user and session scope exceeds 128 bytes")
    }
    return internalKey, nil
}

// New 中创建并登记一次，命令闭包持有同一个声明值。
reminder := task.Define[ReminderArgs]("remind")
task.On(p, reminder).Handle(func(c *sdk.Context, args ReminderArgs) error {
    _, err := c.SendTo(args.Session, message.Text(args.Text))
    return err
})
p.Command("remind").Handle(func(c *sdk.Context, key, text string) (message.Content, error) {
    internalKey, err := reminderKey(c, key)
    if err != nil {
        return nil, err
    }
    _, err = reminder.At(c, time.Now().Add(10*time.Minute),
        ReminderArgs{Session: c.Session(), Text: text}, task.Key(internalKey))
    if err != nil {
        return nil, err
    }
    return message.Text(key), nil
})
p.Command("cancel").Handle(func(c *sdk.Context, key string) error {
    internalKey, err := reminderKey(c, key)
    if err != nil {
        return err
    }
    return reminder.Cancel(c, internalKey)
})
```

`reminderKey` uses trusted `c.User()`/`c.Session()` with length prefixes separating full fields. The 128-byte cap includes identity scope/user key; overflow explicitly returns `INVALID_ARGUMENT`, preserving original input. Import standard `fmt`, `sdk/errcode`, `sdk/task`.

Task callbacks have no Actor/default chat; encode needed chat/user/cancellation key in arguments and send with `c.SendTo(sid, content)`. Proactive sends follow Core `allowed_adapter_instances`. Argument types follow JSON Service rules: package-level named, nongeneric value types encoding top-level JSON objects. Generation checks nested fields/custom-encoding restrictions. Run tests directly or `alonabot build` after new declarations.

`reminder.Every(ctx, time.Hour, args, task.Key(internalKey))` creates intervals; `reminder.Cron(ctx, "0 9 * * *", args, task.Key(internalKey), task.Timezone("Asia/Shanghai"))` uses IANA timezones. `internalKey` retains the identity encoding above. `Timezone` on `At`/`Every` returns `INVALID_ARGUMENT`. Without `Key`, SDK generates/returns a random key. Recreating the same task/key replaces its full rule/arguments while an old in-flight run continues. `Cancel` deletes/cooperatively cancels; slots release when handlers return. Unknown instances preserve Core `NOT_FOUND`.

`reminder.List(ctx)` reads every page and returns complete `task.Instance[ReminderArgs]`. Only the relevant `At`, `Interval` or `Cron` is populated, with Cron `Timezone`; `NextAt == nil` means no next trigger in protocol. Any page/decode failure returns error.

Task clients work after Activated; `OnStart`, `OnReady`, after stop and expired Handler Contexts return `FAILED_PRECONDITION`. Pass `c` within Handlers to retain parent association; independent background contexts carry only component identity. Core persists definitions/interval origins, scheduling the next future trigger after restart. Missed offline/inactive/paused triggers lapse; expired one-shot instances are deleted.

Keep declaring static tasks in `New()`, e.g. `p.Every("tick", time.Minute).Handle(tick)`. Pause/resume/query from activated Handlers/background work:

```go
if err := task.Pause(ctx, "tick"); err != nil {
    return err
}
state, err := task.Status(ctx, "tick")
if err != nil {
    return err
}
if state.LastAt != nil {
    slog.Info("last tick", "at", *state.LastAt)
}
return task.Resume(ctx, "tick")
```

`task.State.LastAt` is the last actual dispatch during this Core run, `nil` before first dispatch/after restart. Paused means `Paused=true`, `NextAt=nil`. Pause lets in-flight handlers continue; resume schedules future triggers from persisted interval origin. Pause survives restart; repeated pause/resume succeeds; unregistered job IDs return `NOT_FOUND`.

## Plugin tests

Call `New()` directly from plugin `main_test.go`, using `plugintest.Start` for real Core/SDK Session/test Adapter. For lifecycle `/ping`:

```go
package main

import (
    "testing"

    "github.com/AlonaBot/AlonaBot/plugintest"
)

func TestPing(t *testing.T) {
    p := New()
    h := plugintest.Start(t, plugintest.Plugin(p))
    alice := plugintest.User{ID: "alice"}
    result := h.Text(plugintest.Private(alice.ID), alice, "/ping")
    result.RequireReply(t, "pong")
}
```

`RequireReply(t, want)` checks every Handler Outcome error, then requires exactly one send with `Sent.Text()` equal to want, reporting plugin/route identity or actual sends on failure. For image/mixed replies also inspect `result.Sends[0].Chain`.

Run `go test ./...` directly. plugintest scans identity/resources from the plugin entry package, using test-binary GOOS/GOARCH/build tags, once per entry package/process. Each Start creates independent temporary directory/configuration/`dev.sqlite`, stopping components at cleanup. Submit supports concurrent injection on one Harness with results assigned by trace.

`Message(chat, user, message.Chain{...})` injects mixed segments; `Sent.Text()` extracts ordered text. `Sent.To`, `Chain`, `Result` preserve target/chain/receipt; `Op` distinguishes send/edit, `Edited` holds edit handles. `Event` injects extension events declared through `Adapter`.

**Permissions and configuration.** `Admin("root")` registers Core Admin; `User.Permissions` carries platform permissions declared through `Adapter`. For example:

```go
moderator := plugintest.User{
    ID: "mod",
    Permissions: []sdk.Permission{sdk.PermissionName("plugintest.moderator")},
}
h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Admin("root"),
    plugintest.Adapter(plugintest.AdapterID, adapter.PermissionDeclaration{
        Name: "plugintest.moderator", Implies: []string{"chat_admin"}, LabelKey: "moderator",
    }))
result := h.Text(plugintest.Group("room"), moderator, "/mod")
```

Here `sdk`/`adapter` come from `sdk`/`sdk/adapter`. Supply configuration matching the tested plugin's Schema:

```go
h := plugintest.Start(t, plugintest.Plugin(p),
    plugintest.Config(p, `{"greeting":"Hello"}`),
    plugintest.AllowedInstances(p, plugintest.InstanceID),
    plugintest.Capabilities("message.send", "message.edit"))
```

`AllowedInstances` defaults to all; `AllowedInstances(p)` makes it empty, causing job/background explicit sends to return `PERMISSION_DENIED`, while trusted inbound replies retain their exemption. `Capabilities` replaces defaults entirely, allowing the example to check button/album `UNSUPPORTED` and plugin fallback. Rejected sends preserve errors in Outcome.

**Buttons.** Click an action name from sent output, specifying this click's user. Core restores action/version/business values from tokens and authorizes each click:

```go
menu := h.Text(plugintest.Group("room"), alice, "/menu")
clicked := h.Click(plugintest.User{ID: "bob"}, menu.Sends[0], "first")
```

Missing actions, URL buttons and absent send receipts are invalid test inputs reported through `t.Fatal`. Expired tokens, invalid generations and denied permissions return empty sends/Handler Outcomes.

**Static and dynamic tasks.** `h.Trigger(p, "tick")` runs a registered static task now; `h.TriggerTask(p, "remind", key)` finds an existing instance using Core's saved arguments. Both use normal Scheduler REQUEST/RESPONSE, job scope, slots/deadlines; Handlers target sends with `c.SendTo(sid, content)`. Admission/lookup errors fail tests; dispatched Handler errors remain in `Result.Outcomes`.

Create showcase reminders through commands, then trigger using the full internal user/chat key:

```go
h.Text(chat, alice, "/demo remind lunch 3600 soup")
user := message.UID(plugintest.AdapterID, alice.ID)
session := message.SID(plugintest.AdapterID, "private:alice", "")
key := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, "lunch")
reminder := h.TriggerTask(p, "remind", key)
```

See [showcase/main_test.go](../examples/plugins/showcase/main_test.go). One-shot instances consume on actual dispatch; recurring instances retain origins/future ticks. Explicit triggering of paused/running static tasks or running/queued dynamic instances returns `FAILED_PRECONDITION`; full plugin dynamic concurrency uses the existing due queue.

**Multi-turn waits.** After `core.wait`, injection returns immediately with `Result.Waiting=true` and the waiting route's `Outcome.Waiting=true`; question sends remain in current `Sends`. Inject the next same-user/chat message normally:

```go
question := h.Text(chat, alice, "/ask")
if !question.Waiting {
    t.Fatalf("ask = %+v", question)
}
answer := h.Text(chat, alice, "answer")
```

Sends/final errors after matching the wait belong to the new injection until continuation completes/waits again. Other users/chats dispatch normally. Core's completion/wait notification determines injection return.

**Background sends.** `h.Background()` consumes sends/edits without event or explicit-task ownership in Adapter completion order. Event and explicit-task sends arrive in their own `Result.Sends`; each record is consumed once. Control background timing through your own channels and read after Send returns:

```go
send, done := make(chan struct{}), make(chan error, 1)
var cancel context.CancelFunc
p.OnReady(func(ready context.Context) error {
    ctx, stop := context.WithCancel(context.WithoutCancel(ready))
    cancel = stop
    go func() {
        select {
        case <-send:
        case <-ctx.Done():
            return
        }
        _, err := sdk.Send(ctx, "plugintest:private:alice", message.Text("background"))
        done <- err
    }()
    return nil
})
p.OnStop(func(context.Context) error { cancel(); return nil })
h := plugintest.Start(t, plugintest.Plugin(p))
send <- struct{}{}
if err := <-done; err != nil {
    t.Fatal(err)
}
records := h.Background()
```

OnReady's callback context cancels on return. This example retains Session ownership while OnStop manages background lifetime. Channels determine send timing; Core notifications determine injection completion.

Complete examples cover echo/lifecycle/showcase/configdb/media. See [plugin guide](plugin-guide.md#store-small-amounts-of-state) for configdb KV/files:

```sh
go test -race -count=2 ./plugintest/... ./examples/plugins/...
```

**LLM substitute.** Use `LLM(handler)` to supply generation results. Plugins still invoke the test Agent through `@agent/llm.generate` in the same Core. First register commands before returning from `New()` (import `sdk/llm`, `sdk/message`):

```go
p.Command("llm").Handle(func(c *sdk.Context, prompt string) (message.Content, error) {
    response, err := llm.Generate(c, llm.Request{Messages: []llm.Message{{Role: "user", Text: prompt}}})
    if err != nil {
        return nil, err
    }
    return message.Text(response.Text), nil
})
```

Add this to same-package `main_test.go`; plugintest automatically scans caller `llm.Request`/`llm.Response` Schemas:

```go
package main

import (
    "context"
    "testing"

    "github.com/AlonaBot/AlonaBot/plugintest"
    "github.com/AlonaBot/AlonaBot/sdk/errcode"
    "github.com/AlonaBot/AlonaBot/sdk/llm"
)

func TestLLMReply(t *testing.T) {
    p := New()
    h := plugintest.Start(t, plugintest.Plugin(p),
        plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error) {
            if len(req.Messages) != 1 || req.Messages[0].Role != "user" || req.Messages[0].Text != "hello" {
                return llm.Response{}, errcode.InvalidArgument.New("unexpected prompt")
            }
            return llm.Response{
                Text: "fixed answer", Model: "test-model",
                InputTokens: nil, OutputTokens: nil,
            }, nil
        }))
    alice := plugintest.User{ID: "alice"}
    result := h.Text(plugintest.Private(alice.ID), alice, "/llm hello")
    result.RequireReply(t, "fixed answer")
}
```

Nil usage above stays nil after Service encoding; known zero usage returns `*int` pointing to `0`. For failure checks replace the handler with:

```go
plugintest.LLM(func(context.Context, llm.Request) (llm.Response, error) {
    return llm.Response{}, errcode.Unavailable.New("synthetic generation failure")
})
```

The same command yields empty `result.Sends`; `result.Outcomes[0].Err` retains `errcode.Unavailable`/`Message`. Without LLM, `@agent` also returns `UNAVAILABLE`. Last LLM function wins; `LLM(nil)` fails before Start. Handlers receive associated contexts and should honor cancellation/deadlines. Schema validates structure; handlers determine business rules/errors.

See full integration test [plugintest/llm_test.go](../plugintest/llm_test.go). Run:

```sh
go test -race -count=2 ./plugintest/...
```

### Media, actions and complete records

Use public `sdk/media` to read incoming content, then upload or forward it. Add this command to your existing `New()` (import `bytes`, `io` and `sdk/media`):

```go
p.FullMatch("copy").Handle(func(c *sdk.Context) error {
    source := c.Event().Payload.(*message.CreatedPayload).Chain[1].(message.FileSegment).Source
    reader, err := media.Open(c, source)
    if err != nil { return err }
    data, readErr := io.ReadAll(reader)
    closeErr := reader.Close()
    if readErr != nil { return readErr }
    if closeErr != nil { return closeErr }
    ref, err := media.Upload(c, bytes.NewReader(data), "image/png")
    if err != nil { return err }
    _, err = c.Send(message.Chain{message.Text("copied"), message.Image(ref.Source()), message.File(ref.Source(), "copy.png")})
    return err
})
```

Generate a small image in the test, inject its platform-file fixture and check the actual Adapter bytes and chain order (import `image` and `image/png`):

```go
var pixel bytes.Buffer
if err := png.Encode(&pixel, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil { t.Fatal(err) }
h := plugintest.Start(t, plugintest.Plugin(New()))
source := h.Fixture(plugintest.InstanceID, "pixel", plugintest.Media{Data: pixel.Bytes(), MIME: "image/png"})
alice := plugintest.User{ID: "alice"}
result := h.Message(plugintest.Private(alice.ID), alice,
    message.Chain{message.Text("copy"), message.File(source, "input.png")})
if len(result.Sends) != 1 { t.Fatalf("sends = %+v", result.Sends) }
sent := result.Sends[0]
if len(sent.Chain) != 3 || len(sent.Media) != 2 { t.Fatalf("send = %+v", sent) }
for _, output := range sent.Media {
    _, format, err := output.Image()
    if err != nil || format != "png" || output.MIME != "image/png" || !bytes.Equal(output.Data, pixel.Bytes()) {
        t.Fatalf("media = %+v, format = %s, error = %v", output, format, err)
    }
}
if sent.Media[1].Filename != "copy.png" { t.Fatalf("filename = %q", sent.Media[1].Filename) }
```

Fixtures are fetched through the real Adapter media RPC. Plugin `media.Open`, `media.Upload` and outgoing Adapter `ReadMedia` use the same Core authorization, storage and token checks. `Sent.Media` contains actual content in chain traversal order. HTTPS sources retain their URLs; use fixtures or uploaded managed media for byte assertions.

`Operations(Behavior{...})` configures the default instance's Send/Edit/Action/Delete/Nodes/Menu callbacks; configure multiple accounts through `InstanceConfig.Behavior`. Send/Edit receive a default successful receipt in `Call.SendResult`, which can be adjusted and returned with an error:

```go
behavior := plugintest.Behavior{
    Send: func(ctx context.Context, call plugintest.Call) (*message.SendResult, *errcode.Error) {
        receipt := call.SendResult
        return receipt, errcode.Unavailable.New("platform accepted the message with a warning")
    },
    Action: func(ctx context.Context, call plugintest.Call) (json.RawMessage, *errcode.Error) {
        return call.Params, errcode.Unavailable.New("platform returned a warning")
    },
}
h := plugintest.Start(t, plugintest.Plugin(New()), plugintest.Operations(behavior))
```

Send/Edit support Accepted, Rejected, Partial, Unknown and any valid result/error combination. Platform Action parameters/results follow the original `Adapter(..., adapter.Action[P,R](...))` or instance Descriptor declaration. The echo callback above suits actions with matching parameter/result Schemas. A declared Action without a callback returns `UNSUPPORTED`. Delete accepts each item by default; Nodes returns request items by default. Configure Nodes to expand a platform-ID query. Default menus persist set/get/delete state by scope/language and also support legacy SyncCommands.

`h.Calls()` retains all completed Adapter operations; `h.WaitCalls(ctx, count)` waits for a count. Each Call includes operation, Adapter/instance/account, target, parameters, actual results, error and TraceID. Event `Result.Calls` and `Sends` use trace ownership. `Result.Outcomes` records actual Handler calls. `Diagnostics` separately records argument parsing and admission failures with Phase, PluginID, RouteID, TraceID and Err. `RequireReply` checks Handler results and replies, including usage replies for invalid arguments. An actual Handler returning `INVALID_ARGUMENT` remains in Outcomes. Use `h.DebugClient().Invoke` to exercise media, menus or actions directly with standard names from the Adapter catalog above.

### Multiple accounts, concurrent events and manual time

```go
p := New()
h := plugintest.Start(t, plugintest.Plugin(p),
    plugintest.Instance(plugintest.InstanceConfig{AdapterID: "first", InstanceID: "one", AccountID: "bot1"}),
    plugintest.Instance(plugintest.InstanceConfig{AdapterID: "first", InstanceID: "two", AccountID: "bot2"}),
    plugintest.Instance(plugintest.InstanceConfig{AdapterID: "second", InstanceID: "three", AccountID: "bot3"}),
    plugintest.DefaultInstance("first", "one"),
    plugintest.DefaultInstance("second", "three"),
    plugintest.AllowedInstances(p, "one", "three"))
alice := plugintest.User{ID: "alice"}
first := h.Submit(plugintest.Private(alice.ID).OnInstance("one"), alice, message.Text("/ping"))
second := h.Submit(plugintest.Private("bob").OnInstance("two"), plugintest.User{ID: "bob"}, message.Text("/ping"))
for _, pending := range []*plugintest.Pending{first, second} {
    result, err := pending.Wait(t.Context())
    if err != nil { t.Fatal(err) }
    result.RequireReply(t, "pong")
}
```

An explicit Instance list replaces the default test instance. Select incoming identity with `OnInstance`, explicit sending defaults with `DefaultInstance`, and platform administrators with `AdminOn(adapterID, userID)`. Descriptor.Capabilities declares the platform capability set; instance Capabilities selects enabled names. Nil inherits Harness capability options; an explicit empty list enables no optional capabilities. Accounts on the same platform can share a Descriptor and restrict capabilities independently. Omitting Instance preserves existing single-instance options. `Submit` submits immediately; `Wait(ctx)` can be cancelled and called again while the event continues. Final TraceIDs, Sends and Calls belong to their own event. `WaitBackground(ctx, count)` waits for and consumes background output, including replies after a conversation wait times out.

```go
clock := plugintest.NewClock(time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC))
t.Cleanup(clock.Close)
p := New()
p.Every("testtick", time.Second).Handle(func(c *sdk.Context) error {
    _, err := c.SendTo("plugintest:private:alice", message.Text("tick"))
    return err
})
h := plugintest.Start(t, plugintest.Plugin(p), plugintest.WithClock(clock))
clock.Advance(time.Second)
records, err := h.WaitBackground(t.Context(), 1)
if err != nil { t.Fatal(err) }
if records[0].Text() != "tick" { t.Fatalf("records = %+v", records) }
```

Natural Scheduler jobs, dynamic `task.At/Every`, continuation deadlines and `c.Wait` share this clock. Use `clock.Now()` when constructing absolute task times. Advance after `Result.Waiting` to expire a conversation wait. Advance delivers due timers; use Wait/WaitBackground for asynchronous RPC completion. Task cancellation and Harness.Close unregister their timers. Production uses real time; Session protocol deadlines retain real absolute timestamps and receive logical-clock cancellation.

### Connect to a running Core

Enable the tested plugin and a real Adapter instance in Core and obtain its local `admin_token`. This test calls an existing `/ping` command. Match the URL and identity to your running configuration (import `context`, `os`, `strings`, `sdk/debug` and `plugintest`):

```go
func TestRunningCore(t *testing.T) {
    token, err := os.ReadFile("../data/admin_token")
    if err != nil { t.Fatal(err) }
    client := &debug.Client{URL: "http://127.0.0.1:7650", Token: strings.TrimSpace(string(token))}
    connection, err := plugintest.Connect(context.Background(), client,
        plugintest.InstanceConfig{AdapterID: "telegram", InstanceID: "main", AccountID: "123456"})
    if err != nil { t.Fatal(err) }
    defer connection.Close()
    alice := plugintest.User{ID: "987654"}
    result, err := connection.Text(t.Context(), plugintest.Private(alice.ID), alice, "/ping")
    if err != nil { t.Fatal(err) }
    result.RequireReply(t, "pong")
    if len(result.Outcomes) == 0 || len(result.TraceIDs) == 0 { t.Fatalf("dispatch = %+v", result) }
}
```

Use the Adapter's actual session-ID format; Telegram private chats use `private:<userID>` as constructed above. `Connection.Event(ctx, event)` injects a complete standard SDK event; Message sends a mixed chain, and Submit with Pending.Wait supports asynchronous integration. `connection.Client.Operations/Invoke` discovers and executes that real Adapter's operations.

The client injects standard Events through authenticated `POST /api/v1/debug/events`. Core rebinds trusted source identity, expands permissions and performs normal command matching, argument parsing and routing. Results retain actual RouteIDs, TraceIDs, replies and simultaneous receipts/errors; each Outcome, Diagnostic and reply carries its owning TraceID. Debug events and Adapter calls wait outside the management metadata write lock and GC critical section, allowing independent sessions to run concurrently. Metadata writes keep their existing locking and collection order. Request cancellation releases its observer; Close releases client resources. Existing Core, Adapter and plugin components keep running, and the real Adapter receive cursor is preserved.

Local Harness, Connect and sdk/debug.Client.Inject expose the same diagnostic fields. For an existing `echo <text>` command called without its argument:

```go
invalid, err := connection.Text(t.Context(), chat, alice, "/echo")
if err != nil { t.Fatal(err) }
invalid.RequireReply(t, "Usage: /echo <text>")
if len(invalid.Outcomes) != 0 || len(invalid.Diagnostics) != 1 { t.Fatalf("dispatch = %+v", invalid) }
diagnostic := invalid.Diagnostics[0]
if diagnostic.Phase != plugintest.PhaseArguments || errcode.Of(diagnostic.Err) != errcode.InvalidArgument {
    t.Fatalf("diagnostic = %+v", diagnostic)
}
```

Import sdk/errcode and use the configured instance's chat and user as `chat` and `alice`. Permission and lifecycle admission failures carry Phase `plugintest.PhaseAdmission` and their actual error code. The sdk/debug constants are `debug.PhaseArguments` and `debug.PhaseAdmission`.

The local integration test [plugintest/debug_test.go](../plugintest/debug_test.go) exercises real Core/Session, authenticated HTTP and a loopback platform. Real-platform integration requires platform credentials and actual chat targets.

## Protocol and session behavior

Component protocol version **1.0** is independent of product version **0.2.0**. The current message types and field numbers are defined in [contracts/proto/alonabot/v1](../contracts/proto/alonabot/v1); the [wire runtime](../internal/wire) implements framing and Session calls. Change `.proto` sources and regenerate Go output. Preserve published field numbers and enum values; reserve removed names/numbers and keep existing field types. [Protocol-number checks](../contracts/wirespec_test.go) protect these identities.

TCP and stdio both encode a four-byte unsigned big-endian length followed by a Protobuf Envelope. The length excludes its prefix and must be 1–4,194,304 bytes. Zero, oversized, truncated or undecodable frames fail the Session; readers do not guess the next boundary. Each connection has one writer. Core holds its loopback TCP listener (`127.0.0.1:0`) until termination. Each component start receives a new Session and 32-byte secret; authentication verifies the secret in constant time plus identity, digest and protocol version. With stdio, stdin/stdout carry protocol frames and stderr carries logs. Keep Bootstrap secrets, configuration, database and HTTP credentials out of logs.

Activation proceeds through HELLO → HELLO_ACK/Bootstrap → REGISTER → PREPARED → ACTIVATE → OnReady → ACTIVATED. Ordinary business calls start only after activation. Protocol majors must agree; the negotiated minor is the smaller supported value. Required capabilities must be supported by both sides; `http_listener` is version 1. See [control.proto](../contracts/proto/alonabot/v1/control.proto) and [route.proto](../contracts/proto/alonabot/v1/route.proto) for registration and lifecycle payloads.

Calls match `(session_id, initiator_side, call_id)`. Call IDs are nonzero, monotonically increasing per direction and never reused in a Session. A parent must identify a still-active Core-initiated call in that same Session. Deadlines are absolute UTC milliseconds; children retain the parent's deadline or shorten it. Cancellation is cooperative and leaves completed side effects intact. Payload and ErrorInfo can coexist for partial results: preserve both. Default request timeout is 30 seconds. A Session failure completes pending calls with their actual errors; invalid authentication closes the connection without exposing credentials.

Use `alonabot tune list` for the full current tuning table, including names, units, defaults, effective values, constraints and sources. `--args NAME=value` applies the same validation. [internal/tune/params.go](../internal/tune/params.go) defines the values; [deployment](deployment.md#plugin-hibernation) explains wake budgets and [plugin development](plugin-guide.md) explains continuations, storage and tasks. Frame size remains fixed at 4 MiB. Phase budgets default to HELLO 10s, configuration 60s, migration 15m, OnStart 60s and OnReady 30s; stopping has 10s total, including 6s drain and 4s OnStop. Crash recovery permits three crashes per five-minute window. Components must honor cancellation in their own handlers and background work.

### Command menu wire additions

The Core→Adapter service `adapter.command.menu` handles set/get/delete; plugin→Core `core.command.menu` handles set/get/delete/status/sync, taking a missing instance from the in-flight parent's trusted instance. Both use schema_version=1. `AdapterInstanceStartResponse.command_menu=5` is optional; absence retains legacy `adapter.commands.sync`. Its CommandMenuConfig has bool auto_sync_commands=1 and clear_commands_on_disable=2. `CommandSpec.menu=14` defaults to false; trusted `MediaSource.mime=6` is a string.

The shared messages are `CommandMenuRequest{instance_id=1, operation=2, menu=3}`, `CommandMenu{scope=1, language_code=2, commands=3}`, `CommandMenuScope{type=1, chat_id=2, user_id=3}`, `CommandMenuEntry{name=1, description=2}`, `CommandMenuResponse{commands=1, config=2, ready=3, last_error=4, positions=5}` and `CommandMenuPosition{scope=1, language_code=2, mode=3, owner_plugin_id=4}`. commands and positions are repeated. Scope types are default/chat/chat_administrators/chat_member; plugins use the last three, without threads.

`EventPayload.kind` adds command_menu_ready=13, with string fields adapter_id=1, adapter_instance_id=2, account_id=3 and chat_id=4. Core targets `bot.command_menu_ready`; an empty chat_id indicates instance manual readiness, otherwise chat availability. Adapter-reported standard events retain the existing allowlist. See the [Adapter protocol](../contracts/proto/alonabot/v1/adapter.proto), [routes](../contracts/proto/alonabot/v1/route.proto) and [messages](../contracts/proto/alonabot/v1/message.proto) for definitions, and [command menus](plugin-guide.md#command-menus) for lifecycle and permission rules.

## Agent, streaming services and lifecycle API

[Complete parameters, capability status, eight hook stages and Goal recovery](agent-api.md).
