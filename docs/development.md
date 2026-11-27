# Source development

[简体中文](development.zh-CN.md)

Run a plugin from source, then work in Core, SDK, Adapter, Agent or WebUI as needed. See [contributing](../CONTRIBUTING.md) for workflow and [architecture](../ARCHITECTURE.md) for modules/data ownership. This guide maintains environments, commands and verification; management endpoints and runtime behavior are in the [management reference](management-api.md); installing and running a bot is in the [user documentation](https://alonabot.github.io/en/docs/usage/install).

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

Enter `/ping` in the same terminal for `pong`; Ctrl-C stops normally. Parallel Core instances need separate data directories/listener ports. See [Core settings](management-api.md#core-settings-and-http-entry-point) for first authentication/listener settings.

The example marks `/ping` with `.Menu()` for automatic chat menus on supported Adapters; unmarked commands remain callable. See [command menus](https://alonabot.github.io/en/docs/tutorial/commands#platform-command-menus) for manual control.

## Query the CLI version

After building the CLI, run `bin/alonabot --version`. An ordinary source build prints:

```text
alonabot version dev
https://github.com/AlonaBot/AlonaBot
```

Release builds inject the version and UTC build date through `-ldflags`; for example:

```fish
go build -ldflags '-X main.version=v0.3.0 -X main.buildDate=2026-11-23' -o bin/alonabot ./cmd/alonabot
bin/alonabot --version
```

```text
alonabot version 0.3.0 (2026-11-23)
https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.0
```

The version accepts an optional leading `v`; output omits it and Release links include it. An empty build date omits the parenthesized date. `--version` writes to stdout and exits with status 0 before parsing tuning parameters or starting Core. The distribution script injects both fields automatically and retains the existing Core API version injection.

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

Run `alonabot check .`, `alonabot run .` or `alonabot pack .` in a component project. Core/SDK retain an independent module. Cross-component tests use separate `tests/go.mod`, keeping model SDK/Adapter implementation dependencies there.

```fish
uv run --no-project python scripts/integration.py
# 按改动选择测试，例如 Agent 服务：
uv run --no-project python scripts/integration.py --run '^TestAgentServices$' --race
```

The integration entry point first runs `tests/agent`/`tests/e2e`. Agent service tests needing private Core assembly run via temporary modfiles/overlays using the Agent version resolved by `tests/go.mod`, without writing dependencies/tests back into Core's module. Test dependencies may pin real development commits; bundled releases resolve the latest published components separately. CLI end-to-end tests live in `tests/e2e`; `cmd/alonabot` contains production commands only.

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

PR Go CI runs formatting, generation, `go vet ./...`, affected-package tests and `go build ./...`. Manual dispatch and releases run the complete gate; race tests, cross-component integration tests, platform builds and release checks run as separate jobs. [The workflow](../.github/workflows/ci.yml) is authoritative. WebUI's separate CI also type-checks/builds the frontend. For a full manual check:

```fish
go build ./...
go vet ./...
go test -race ./...
uv run --no-project python scripts/integration.py --race
```

For page-kit changes run `go test -race -v ./sdk/page ./examples/plugins/showcase` and `node --check sdk/page/kit.js`, then check key interactions/language/theme in WebUI's iframe. For documentation, check links, anchors, code blocks and actual commands.

Cross-platform changes need actual runtime checks as well as cross-builds. Five targets: linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64. Record builds separately from real runs; current results are in [release notes](../changelog/v0.1.0/en.md#added). Use authorized test credentials for real Telegram/paid Provider checks; identify simulators in results.

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) runs on PRs targeting main and manual dispatch, and is called by the Release workflow before building or publishing artifacts. Pushes to main do not trigger this workflow. `Repository checks` checks whitespace and main file scope. Documentation-only PRs finish there.

`PR checks` runs formatting, generation, vet and a root-module build for Go or CI changes. Ordinary tests cover the root packages containing changed Go files; changes under `contracts/` also run protocol tests. Dependency changes (`go.mod`/`go.sum`) run all root-module ordinary tests. Deleted packages are omitted. Changes under `tests/` run the `tests` module without race detection, retaining its existing serial execution and 30-minute package timeout.

Manual dispatch and releases run `Race tests`, `Integration tests`, `Platform builds` and `Release tools` in parallel after repository checks. Release checks use the exact release tag. All Go jobs cache dependencies and build results with both `go.sum` and `tests/go.sum`; full jobs report failures in English and retain the complete step logs. PR success records the light checks; record race, integration and platform results from the corresponding full jobs. Main path checks exclude personal maintenance records while retaining product skills and test tools.

[PR metadata](../.github/workflows/pr-metadata.yml) validates nondraft PR descriptions. Triage/auto-assign only change GitHub metadata and never execute PR code. Code checks have read-only permissions without production secrets; maintainers configure branch protection separately.

## Development runs and hot reload

Build CLI first: `go build -o bin/alonabot ./cmd/alonabot`. From the repository root, get a reply in three steps (fish):

1. Create project: `bin/alonabot init ./alonabot-plugin-hello`.
2. Run: `bin/alonabot run ./alonabot-plugin-hello`.
3. Enter `/hello` in the same terminal for `Hello AlonaBot`; Ctrl-C stops Core.

Development directories enable reload by default. Change source and send `/hello` again to observe new replies; `--reload=false` disables it.

See the [devbot tutorial](../examples/devbot/README.md) for loopback injection, configuration, hot reload and same-process debugging.

Full component-author usage lives on the documentation site: commands and flags in the [CLI reference](https://alonabot.github.io/en/docs/reference/cli), plugin tests in [Testing](https://alonabot.github.io/en/docs/tutorial/testing) and [Advanced testing](https://alonabot.github.io/en/docs/advanced/testing), and in-process runs under [In-process operation](https://alonabot.github.io/en/docs/internals/architecture#in-process-operation).

When changing the related implementation, the integration tests are:

- Library-mode public entry: `TestLibraryLifecycle` in root `alonabot_test.go`; assembly, internal services, candidates and WebUI: `internal/app/`; development CLI and hot reload: `tests/e2e/`.
- Adapter debug catalog and calls: `go test -race ./internal/core/adapter/remote -run '^TestDebugAdapter$'`.
- Plugin test suite: `go test -race -count=2 ./plugintest/... ./examples/plugins/...`; the running-Core path is in `plugintest/debug_test.go` and the LLM double in `plugintest/llm_test.go`.

## Protocol and session behavior

Component protocol version **1.0** is independent of product version **0.3.0**. The current message types and field numbers are defined in [contracts/proto/alonabot/v1](../contracts/proto/alonabot/v1); the [wire runtime](../internal/wire) implements framing and Session calls. Change `.proto` sources and regenerate Go output. Preserve published field numbers and enum values; reserve removed names/numbers and keep existing field types. [Protocol-number checks](../contracts/wirespec_test.go) protect these identities.

TCP and stdio both encode a four-byte unsigned big-endian length followed by a Protobuf Envelope. The length excludes its prefix and must be 1–4,194,304 bytes. Zero, oversized, truncated or undecodable frames fail the Session; readers do not guess the next boundary. Each connection has one writer. Core holds its loopback TCP listener (`127.0.0.1:0`) until termination. Each component start receives a new Session and 32-byte secret; authentication verifies the secret in constant time plus identity, digest and protocol version. With stdio, stdin/stdout carry protocol frames and stderr carries logs. Keep Bootstrap secrets, configuration, database and HTTP credentials out of logs.

Activation proceeds through HELLO → HELLO_ACK/Bootstrap → REGISTER → PREPARED → ACTIVATE → OnReady → ACTIVATED. Ordinary business calls start only after activation. Protocol majors must agree; the negotiated minor is the smaller supported value. Required capabilities must be supported by both sides; `http_listener` is version 1. See [control.proto](../contracts/proto/alonabot/v1/control.proto) and [route.proto](../contracts/proto/alonabot/v1/route.proto) for registration and lifecycle payloads.

Calls match `(session_id, initiator_side, call_id)`. Call IDs are nonzero, monotonically increasing per direction and never reused in a Session. A parent must identify a still-active Core-initiated call in that same Session. Deadlines are absolute UTC milliseconds; children retain the parent's deadline or shorten it. Cancellation is cooperative and leaves completed side effects intact. Payload and ErrorInfo can coexist for partial results: preserve both. Default request timeout is 30 seconds. A Session failure completes pending calls with their actual errors; invalid authentication closes the connection without exposing credentials.

Use `alonabot tune list` for the full current tuning table, including names, units, defaults, effective values, constraints and sources. `--args NAME=value` applies the same validation. [internal/tune/params.go](../internal/tune/params.go) defines the values; [plugin hibernation](management-api.md#plugin-hibernation) explains wake budgets and [plugin development](https://alonabot.github.io/en/docs/tutorial/project) explains continuations, storage and tasks. Frame size remains fixed at 4 MiB. Phase budgets default to HELLO 10s, configuration 60s, migration 15m, OnStart 60s and OnReady 30s; stopping has 10s total, including 6s drain and 4s OnStop. Crash recovery permits three crashes per five-minute window. Components must honor cancellation in their own handlers and background work.

### Command menu wire additions

The Core→Adapter service `adapter.command.menu` handles set/get/delete; plugin→Core `core.command.menu` handles set/get/delete/status/sync, taking a missing instance from the in-flight parent's trusted instance. Both use schema_version=1. `AdapterInstanceStartResponse.command_menu=5` is optional; absence retains legacy `adapter.commands.sync`. Its CommandMenuConfig has bool auto_sync_commands=1 and clear_commands_on_disable=2. `CommandSpec.menu=14` defaults to false; trusted `MediaSource.mime=6` is a string.

The shared messages are `CommandMenuRequest{instance_id=1, operation=2, menu=3}`, `CommandMenu{scope=1, language_code=2, commands=3}`, `CommandMenuScope{type=1, chat_id=2, user_id=3}`, `CommandMenuEntry{name=1, description=2}`, `CommandMenuResponse{commands=1, config=2, ready=3, last_error=4, positions=5}` and `CommandMenuPosition{scope=1, language_code=2, mode=3, owner_plugin_id=4}`. commands and positions are repeated. Scope types are default/chat/chat_administrators/chat_member; plugins use the last three, without threads.

`EventPayload.kind` adds command_menu_ready=13, with string fields adapter_id=1, adapter_instance_id=2, account_id=3 and chat_id=4. Core targets `bot.command_menu_ready`; an empty chat_id indicates instance manual readiness, otherwise chat availability. Adapter-reported standard events retain the existing allowlist. See the [Adapter protocol](../contracts/proto/alonabot/v1/adapter.proto), [routes](../contracts/proto/alonabot/v1/route.proto) and [messages](../contracts/proto/alonabot/v1/message.proto) for definitions, and [command menus](https://alonabot.github.io/en/docs/tutorial/commands#platform-command-menus) for lifecycle and permission rules.

### Message description and read-only check protocol

`core.message.describe` takes `DescribeMessageRequest{to=1}` and returns `MessageDescription{formats=1, extensions=2, combinations=3}`. `core.message.check` takes `CheckMessageRequest{to=1, chain=2}` and returns `CheckMessageResponse{status=1}`. Both use protobuf schema version 1. An omitted to uses the trusted parent target; admission follows current send rules. See the [outbound message API](agent-api.md#query-and-check-outbound-messages-by-target) for interactive and scheduled SDK usage.

AdapterDescriptor appends message_formats=7 and message_combinations=8; ExtensionDeclaration appends description=4 and example_json=5. MessageFormat fields are name, version, status, description, schema_json, example_json, sources, text_limit, button_label_limit and adapter_id (numbers 1–10). MessageCombination uses name=1, formats=2, description=3 and example_json=4. SupportStatus values 0/1/2 mean unknown/supported/unsupported. Definitions live in [message.proto](../contracts/proto/alonabot/v1/message.proto).

Optional `adapter.message.check` reuses AdapterSendRequest's instance, target, chain and generation and returns CheckMessageResponse. `adapter.operations` advertises `sdk/message.check` when the actual instance implements `sdk/adapter.MessageChecker`. Core checks the current catalog and Session/generation before the RPC. Media Store's PrepareCheckSource reads authorized metadata while preserving source kinds and tokens. A check status and error can be returned together.

SendRequest.require_original_segment_indexes=3 is bool, set by `sdk.RequireOriginalSegmentIndexes()`. The default false retains existing send and Hook behavior. When true, Core tracks caller-input segments and rejects the first non-equal Hook response with FAILED_PRECONDITION and a REJECTED receipt before platform side effects. This uses the existing per-Hook validation callback; the Hook protocol and Edit interface remain unchanged.

ToolConfirmRequest.require_confirmation=4 is bool. `sdk/tool.ConfirmRequired` sets it to enter the trusted Actor confirmation flow even under allow; deny takes precedence. Existing Confirm leaves it unset and follows the existing policy.

## Agent, streaming services and lifecycle API

[Complete parameters, capability status, eight hook stages and Goal recovery](agent-api.md).
