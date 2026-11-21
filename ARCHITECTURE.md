# AlonaBot architecture map

[简体中文](ARCHITECTURE.zh-CN.md)

Scope: module responsibilities, allowed import directions and state ownership. Public behavior is described in [plugin development](docs/plugin-guide.md) and [deployment](docs/deployment.md); protocol types live in [contracts/proto](contracts/proto/alonabot/v1).
When adding top-level packages or changing dependencies/state ownership, update this file and the [development](docs/development.md) checklist together.

Implementation directories/responsibilities follow; start with [source development](docs/development.md) and [plugin development](docs/plugin-guide.md).

## Modules and responsibilities

Core/SDK share github.com/AlonaBot/AlonaBot. Official components have separate repositories/modules; tests/ isolates integration dependencies and components.lock.json pins bundled source.

| Module (path) | Owns | Excludes |
| --- | --- | --- |
| `contracts/` | .proto sources (proto/), generated pb/alonabotv1, manifest/release/message validation, schedule triggers, generic Markdown parsing (markdown/Goldmark), compatibility fixtures; contracts/internal/protocgen runs only through go:generate | Runtime I/O, Session state, platform/GORM dependencies |
| `internal/wire/` | Frame encoding (4 MiB cap), TCP/stdio, Session call tables, admission budgets, priority channels, cancellation | Business routing, permissions, lifecycle decisions |
| `internal/tune/` | Names/defaults/units/constraints and --args parsing | Business meaning implemented by consumers |
| `internal/oslock/` | Cross-process exclusive flock/LockFileEx for core.lock/runtime.lock | Business lock timing |
| `internal/inproc/` | Library-mode startup arguments/entry points: app creates arguments, SDK reads/registers entry points during init | Startup itself, owned by SDK/app |
| `internal/pluginkit/` | Developer CLI packaging, release.json/signing (build/pack) | Install/preflight, owned by internal/core/packages |
| `sdk/` and subpackages | Root plugin/Context/Send/builders/permissions/lifecycle/runtime API; category task/service/bus/button/media/conversation/tool/action/options/command APIs live under corresponding sdk/ paths, sharing sdk/internal/registry; runtime.lock; sdk/message messages/events/SendResult/media and string SID/ParseSID/Ref/ParseRef/UID/ParseUID; sdk/errcode; CLI-injected resource registration via sdk/genimpl; optional database/sql and GORM managed SQLite; sdk/llm types/Ask and sdk/agent Generate targeting @agent | Core state, management credentials, other plugins |
| `internal/core/` | State/core.lock, processes, RPC orchestration, routing, permissions, media, services, scheduling, HTTP, configuration, management | Platform protocol details and in-plugin process logic |
| `internal/app/` | Core assembly/candidates/processes, in-process development components; developer CLI separate components, individual restarts/settings reload. Outside core because it imports SDK | Root public library interfaces, business logic |
| `internal/generate/` | Build/test scanning of entries/resources/registration, overlay resources/production entry points shared by CLI and plugintest | Runtime resource access (SDK), author-directory writes |
| `internal/generated/` | Entry-import-path resource registry binding objects to raw manifests/configuration/instance/Service Schemas/i18n/Handler names, recording plugin entry packages | Scanning/generation |
| Root `alonabot` Go package | Thin public library wrappers around internal/app | Assembly/developer CLI internals/business logic |
| [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) | Optional official /help and /whoami, SDK-only Core access | Command catalog/permissions (Core) |
| [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) | Replaceable role: agent, conversation loop/Providers/tools/MCP/memory/skills/tasks/llm.generate, SDK-only Core access | Core tool policy/confirmation/conversation storage |
| [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) | Receive/normalize/send plans/capability/platform mapping | Core policy/command routing/cursor persistence/button tokens |
| [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) | Forward/reverse WebSocket (reverse via Core port), receive/normalize/send plans/capability/platform mapping/local simulator | Core policy/command routing/button tokens |
| `plugintest/` | go test harness using alonabot.Start for real Core/test Adapter/plugin, event injection/send collection | Separate validation/dispatch paths: Core performs them normally |
| `cmd/alonabot/` | Assembly for Core start/stop/development/management CLI | Business logic |
| `examples/plugins/` | SDK-compiled examples/integration samples | Production dependencies |
| [alonabot-webui](https://github.com/AlonaBot/alonabot-webui) | role: webui, Vite/TypeScript/React 19 UI and Go backend; platform processes bundle the same dist | Core management data ownership/write rules (Admin API) |
| `Dockerfile`, `docker-compose.yml`, `deploy/` | Official images, compose bundle entry point, install.sh | Core behavior: images/scripts install release archive contents |

Each Core's sole lifecycle owner is supervisor.Waker, created in app with WakeOptions{Store,Global,QueueRequests,QueueBytes,Start}, injected through supervisor.Spec.Waker. EnsureActive(ctx context.Context,pluginID string,payloadBytes int64)(func(),error) coalesces starts, manages three wait budgets and returns an activity release. Ready(running *Running) opens delivery only after ACTIVATED/digest verification/Router/Bus/HTTP/Scheduler binding. AcquireActivity(pluginID string,session *wire.Session)(func(),error) acquires current Session references with idempotent release tied to that Session. Running.StopIdle(t0,lastActivityAt time.Time) host.StopResult rechecks management/Session/Ready/zero references/actual activity-end time under plugin lock and Waker gates; changes yield Cancelled. QUIESCING freezes dispatch; new work waits for old-tree exit/SUSPENDED/new Ready. wire.Options.RequestActivity func(*wire.Inbound)(func(),*pb.ErrorInfo) connects synchronously via host.ActivateOptions/supervisor.Prepare, sharing in-flight registration's critical section; normal WaitIdle covers full admission.

Router/Broker Options.EnsureActive and HTTP Server.SetWake(snapshot,ensure) share waking. Match/filter/admit before waking; after binding recheck generation/routes within original deadlines. router.WithActivity(p,acquire) wraps real calls/EVENT writes and existing Request.Late, including Hooks/running schedules. Bus Attach(p router.Plugin,caller ...router.Caller) retains one-argument usage; production passes subscriber Caller and holds queued work through Options.AcquireActivity. Media uses Store.SetActivity; continuations/waits use continuation.Options.AcquireActivity. Production proactive outbound services retain references until business return. HTTP releases on proxy return; running SDK HTTP/Bus callbacks retain normal stop deadlines.

Router.Options.MessageConfig reads Active configuration by chat profile, merging Resolve/effective Overrides. Waker Activity(pluginID string) ActivityStatus independently snapshots Session/Ready/References/LastActivityAt/ColdStarts/ColdStartSequence. Idle controller consumes these then AcknowledgeColdStarts(pluginID string,through uint64) drops historical prefixes for resident/disabled/sleeping instances alike. It clears records by window while keeping fixed cooldown deadlines. Scheduler wakes retain original JOB_TIMEOUT; see [verification](docs/development.md#verify-changes).

scheduler.Options.EnsureActive func(ctx context.Context,pluginID string,payloadBytes int64)(release func(),err error) injects the same assembly.ensureActive; Scheduler.Suspend(pluginID string) retains normal IDLE plans. Static deadlines are original scheduled_at+JOB_TIMEOUT; dynamic start on concurrency acquisition/delivery/wake wait. Waiting retains run slots/dynamic concurrency and counts full encoded HandlerRequest against wake budgets. Same-generation/route resume Attach retains plans/admitted work/consumed-at pending ownership/origins/pause; unregistered expired at entries update Expired/Unregistered and return quota. Supervisor alone withdraws separate-process schedules/Released; old-tree exit releases only old Session slots. app cancels failed-wake plans inside Core.operation with full errors.

internal/core/idle provides read-only Resolve(s settings.Sleep,m *manifest.Manifest,independent bool,routes []*pb.RouteDescriptor) Policy for admission/reasons/all-route warnings. New(values tune.Values)*Controller and CanStop(pluginID string,policy Policy,runtime state.RuntimeStatus,activity supervisor.ActivityStatus,now time.Time)bool check idle/minimum uptime/fixed cooldown. app lifecycle checks each second and calls StopIdle. File.WithPublisher(publish func(change func()error)error)*File shares settings document/snapshots/locks, publishing Update/Reload through Core.operation. Effective sleep changes await same-generation resume outside loop/file/plugin locks and aggregate errors.

app injects read-only manage.Options.SleepPolicy using idle.Resolve and committed identity-matched snapshots, retaining dependency direction. Six sleep endpoints share management authentication/settings publication; CLI uses the same HTTP service. Reads keep components asleep and return actual 412 without valid snapshots. Writes resume outside locks and fully return persistence/wake errors.

internal/core/profile orchestrates storage/instance-chat resolution/Active-schema overrides/management. profile.New(ctx,profile.Options{Store,Settings,Schema}) snapshots once per operation; writes refresh state, independent next operations rebuild. Resolve(instance,session) Profile and Overrides(pluginID) map[string]map[string]json.RawMessage read in memory. HTTP writes serialize and hold relevant plugin management locks from Active schema read to commit. manage provides profile management and four core.profile.* services via RegisterProfiles(mux *router.Mux); CLI alonabot profile uses HTTP. list/current read snapshots, assign skips schema, set reads caller schema only. Writes try catalog → instances → package lock (set only); contention immediately returns FAILED_PRECONDITION.

app injects Router Options.Profile func(context.Context,string,*pb.Target)(string,bool,error) and Options.MessageConfig func(context.Context,string,*pb.Event)(json.RawMessage,error). profile.WithView/ReadView share profile choice/switch/Active/effective overrides for admission/filtering. PreEvent checks switches before media preparation. Invocation.ProfileID/Root propagate through Handler/Service/Hook/jobs/continuations. Begin()func() retains/releases roots; AfterFinished(func()) runs outside locks after root/continuations release, immediately if finished, enabling delayed default Pending reload.

profile.Service.EffectivePlatform(id string)(prefixes []string,reply settings.Reply,err error) merges default/overrides by key, preserving explicit zeros. app injects Router Options.Prefixes func(context.Context,*pb.Target)([]string,error): dispatch uses full Target; menus select instance-default profiles with empty chat. message.Service.SetReply(resolve func(context.Context,*pb.Target)(settings.Reply,error)) supplies replies. message rewrites before Hooks/revalidates; Invocation.Root shares first-send markers across manual/automatic/Service/continuation replies and cleans at root end. Edit decorations associate confirmed segments by Caller Session/full Target/physical message ID. app calls message.Service.Release(router.Caller) at Session exit, invalidating prepared/late receipts too.

app assembles manage.Options.NotifyProfileOverrides func(context.Context,string)error: new snapshots send complete effective overrides to online Sessions; failure closes the old Session and returns actual errors. Offline plugins receive Bootstrap; candidates filter by candidate schema. contracts/configschema.Schema.ApplyProfile(raw json.RawMessage,overrides map[string]json.RawMessage)(json.RawMessage,error) replaces paths; sdk/config.Values.WithOverrides(overrides map[string]json.RawMessage)(Values,error) creates independent read-only views retaining subpaths. SDK publishes atomically/freezes per Context; c.Config returns whole-plugin invocation configuration, p.Config/s.Config retain base.

Configuration reply examples follow dependency ownership: contracts/configschema checks examples/references/translations, Schema.ExampleImages() returns sorted unique image paths. contracts/i18n checks dictionary example shapes/text types. pluginkit validates both schemas' referenced images/collects snapshots; packages rechecks install/update. manage selects packages through Active's committed Journal and serves config_schema-referenced images only via authenticated GET/HEAD.

sdk/kv provides Session-based storage. sdk.PluginDataDir(ctx) selects owning Session; in-process plugins get separate directories, separate processes default if unbound. SQLite drivers link optionally through sdk/sqlitedb/gormdb/kv, never SDK root.

internal/core/diagnostics collects/redacts/zips. Management and CLI diagnose share Write(ctx context.Context,w io.Writer,src Source,opts Options)error. Running reuses state.Store/runtime/schema callbacks; Directory acquires Core-equivalent flock/LockFileEx through x/sys on existing read-only core.lock, querying primary SQLite and temporary existing-WAL copies read-only through GORM. Diagnostics reuse state records, configschema migrations/secret paths, secret references; entry points own authentication/HTTP/CLI replacement.

pluginkit validates component changelogs in check/build/pack scans; pack preserves text, packages revalidates after extraction. manage GET /api/v1/packages/{id}/changelog reads installed packages; GET /api/v1/core/changelog reads changelog_core.go generated/embedded from changelog/v<SemVer>/{en,zh-CN}.md.

sdk/service.Builder.Effect reports tool categories in REGISTER; sdk/tool.List reads Info.Effect; Agent converts MCP annotations for sdk/tool.Confirm. Core tool merges declarations/confirmations/policy overrides. Batch management writes hold catalog then component-ID-ordered locks, sharing installation hierarchy. Catalog.SetPolicy(ctx,rows ...state.ToolPolicy) calls Store.SetToolPolicy(ctx,rows ...ToolPolicy), atomically saving/deleting overrides.

Before tool registration, app sets Catalog.SetProfile(resolve func(ctx context.Context,target *pb.Target)(profiles *profile.Service,profileID string,err error)), reusing targetProfile/Resolve per-operation snapshots. list/invoke/confirm read trusted parent DefaultTarget restored by Session, resolving once; no Target/parent list uses default; all list entries share snapshots. profile.Service.ToolPolicy(id,pluginID,name string,operation bool)(policy pb.ToolPolicy,enabled bool,err error) shares Tools/PutTools entries, matching tools/operations independently. No override yields UNSPECIFIED; absent plugin switch means enabled. Disabled plugins deny, otherwise profile/global/declaration precedence; confirmation operations default confirm. Updates affect next operations; in-flight confirmations retain policy. Global Effective/Policies/SetPolicy retain meaning; absent SetProfile uses global paths. Broker execution inherits permissions/Actor/Target/cancellation/deadlines and preserves results/errors.

Plugin/SubPlugin/CommandGroup text/Raw/event APIs omit route IDs; Hooks on Plugin/SubPlugin use builder .ID for stable identity. *sdk.Context implements context.Context for category clients. c.Reply(text,more...) sends one text/segment message; Result.RequireReply checks Handler errors/single-send text.

sdk/page belongs to SDK. Collecting .Page(title) calls page.Handler(), adding public sdk.AuthStatic GET kit.css/kit.js beside pages while retaining management-authenticated HTML/data. Existing same-method/path author routes win; manual Handler mounting remains. kit.js handles bridge/async translation/theme/locale/shortcuts; WebUI owns sandbox/parent bridge/global search. See [plugin pages](docs/plugin-guide.md#plugin-pages) and examples/plugins/showcase.

[Agent pages.go](https://github.com/AlonaBot/alonabot-agent/blob/main/pages.go) uses sdk/page/management HTTP for personas/MCP/skills/memory/tasks/assets. persona/conversation_record belong to managed initial migration ID:1, memory retains ID:2. Core owns bodies; Agent records full conversation keys after successful append, sharing chat locks for directory clear/append/reset. Normal/scheduled rounds read current persona prompts; skills snapshot per round; MCP gate coordinates state/reconnect/stop latch. Session-end notices isolate identity/preserve latest tool catalog. Dynamic tasks use existing SDK List/Cancel.

sdk.ProfileOverrides(ctx context.Context)(map[string]map[string]json.RawMessage,error) deep-copies current effective overrides from ctx's Plugin Session, sharing HTTP/lifecycle ownership, isolating in-process plugins, defaulting in separate processes. Bootstrap/notification atomic publication is reused; Agent combines base/core.profile.list for persona references.

sdk.AdapterClient binds trusted identity through Context.Adapter(), reusing Context.CallAction/core.action/adapter.action. ActionRequest.use_default_target lets bound clients select parent chat/account by declared scope; existing declarations/Schema/instance permission/outbound Hook disable/lifecycle remain. sdk/action.Caller accepts bound client/Context/background Client; required Instance methods stay unchanged.

## Dependency direction

A -> B means A's Go source imports B (static, including tests). Runtime calls/data flow are separate. Modules are table path prefixes; imports within modules (contracts/release → contracts/manifest, sdk → sdk/message) are unrestricted. Only listed cross-module edges are allowed; standard/approved third-party dependencies are exempt.

| Allowed edge | Reason |
| --- | --- |
| `internal/wire -> contracts` | Generated Envelope payloads |
| `internal/wire -> internal/tune`, `sdk/... -> internal/tune`, `internal/core/... -> internal/tune` | Effective tuning values |
| `sdk/... -> contracts`, `sdk/... -> internal/wire`, `sdk/... -> internal/oslock` | Shared protocol/runtime.lock; Go permits this module's SDK to import internal |
| `internal/core/... -> contracts`, `internal/core/... -> internal/wire`, `internal/core/... -> internal/oslock` | Core protocol/core.lock and temporary runtime.lock during offline copying |
| `internal/core/manage -> internal/core/profile`, `internal/core/router -> internal/core/profile`, `internal/core/tool -> internal/core/profile`; `internal/core/profile -> internal/core/state`, `-> internal/core/settings`, `-> contracts/configschema`, `-> internal/wire` | Management/fixed-service profile snapshots; shared Router admission/filter views; same-operation tool/confirmation overrides/switches; state assignments/overrides, settings default platform, Active-schema validation, wire.MaxFrame override budget |
| `internal/core/idle -> internal/core/settings`, `-> internal/core/supervisor`, `-> internal/core/state`, `-> contracts/manifest`, `-> contracts/pb/alonabotv1`, `-> internal/tune`; `internal/app -> internal/core/idle` | Read-only policy/activity/runtime/tuning; app owns controller/lifecycle |
| `internal/core/buttons -> internal/core/state`, `internal/core/buttons -> internal/core/router` | Persist button tokens/trusted Caller; message/router inject functions |
| `internal/core/state`, `internal/core/settings`, `internal/core/journal -> internal/core/logging` | Shared CRITICAL for state/settings/postcommit faults; separate sdk.LevelCritical with common protocol LogLevel |
| `internal/core/manage -> internal/core/scheduler` | Retained static schedules including hibernation; persistent pause/resume and Activated admission |
| `internal/app -> internal/core/adapter/remote` | Instance runtime queries, config failure snapshots cleared on changes |
| `internal/core/broker -> internal/core/metrics`, `internal/core/bus -> internal/core/metrics`, `internal/core/router -> internal/core/metrics`, `internal/app -> internal/core/metrics` | In-memory Service/Bus/event counters by reason, exposed to management |
| `sdk -> sdk/internal/convert`, `sdk/adapter -> sdk/internal/convert`; `sdk/internal/convert -> contracts`, `-> sdk/message`, `-> sdk/errcode` | Central SDK/pb conversion and typed Adapter declarations; public SDK types |
| `sdk -> sdk/errcode`, `sdk/message -> sdk/errcode`, `sdk/adapter -> sdk/errcode` | Shared SDK errors/codes; errcode uses standard library only |
| `sdk/command -> sdk`, `-> sdk/errcode`, `-> sdk/internal/registry`, `-> contracts/pb`, `-> internal/wire` | Parent association/fixed services returning command.Entry |
| `sdk/kv -> sdk/sqlitedb`, `sdk/kv -> sdk/errcode` | Managed Session SQLite and SDK errors |
| `sdk/message -> contracts/validate` | Shared Adapter ID validation for sessions/handles/users |
| `sdk -> sdk/internal/messagecodec`, `sdk/message -> sdk/internal/messagecodec`, `sdk/internal/convert -> sdk/internal/messagecodec` | Cross-package button/segment errors reported before send, retaining omitted quote-target fields |
| `alonabot-adapter-{telegram,onebot11}/... -> sdk/...`, `-> contracts` | Adapters use only public sdk/adapter to access Core |
| `cmd/alonabot -> internal/core/...`, `-> contracts`, `-> internal/tune` | Assembly/run shares app library assembly; tuning table for --args/tune list |
| `internal/app -> internal/core/...`, `-> sdk/...`, `-> contracts`, `-> internal/wire` | Same-process Core/development assembly |
| `tests/app` overlay `-> alonabot-agent/...` (integration tests only) | Official Agent/local Provider generation and option sources |
| `internal/app -> internal/inproc`, `sdk/... -> internal/inproc` | app passes library startup parameters, SDK init registers entry points for app; internal prevents external construction |
| `cmd/alonabot -> internal/generate` | Shared init/check/build/pack scanning/generation |
| `internal/pluginkit -> internal/generate`, `plugintest -> internal/generate`; `sdk`, `sdk/internal/registry`, `sdk/genimpl`, `internal/app`, `plugintest`, `internal/generate -> internal/generated`; module tests `-> internal/generated` | Internal generator shared by CLI/plugintest; registry imports standard library only |
| Root `-> internal/app`, `cmd/alonabot -> internal/app`; `internal/app -> internal/core/...`, `-> sdk/...`, `-> contracts`, `-> internal/inproc`, `-> internal/tune` | Root exposes only library API; app owns assembly/developer CLI, cmd never imports root |
| `cmd/alonabot -> internal/pluginkit`, `internal/pluginkit -> contracts/...`, `-> internal/generate` | CLI build/packaging reuse generation/contracts |
| `examples/plugins/... -> sdk/...`, `-> contracts` | Examples use public SDK |
| `tests/e2e -> alonabot-adapter-telegram/telegramtest` (integration only) | Directory-run/fault matrix use official Adapter/Bot API simulator/development assembly |
| `scripts/dist -> contracts/manifest`, `-> contracts/release`, `-> internal/releasenotes`; `scripts/release-notes -> internal/releasenotes`; `internal/releasenotes -> contracts/manifest`; `scripts/release-check -> contracts/release` | Release builds/checks use contract signatures; other steps exec go/alonabot |
| `alonabot-webui -> sdk`, `-> sdk/message`, `-> sdk/sqlitedb`, `-> contracts/manifest`, `-> contracts/pluginindex`, `-> contracts/release` | SDK event observation/receipts/management routes and own managed SQLite; contracts parse sources/versions/releases; browser Admin API manages Core data |
| `examples/devbot -> root`, `-> sdk/...` | Library-mode example |
| `plugintest -> root`, `-> sdk/...`, `-> internal/inproc`, `-> contracts`; `examples/plugins/... -> plugintest` (tests only) | Start assembles Core, inproc provides dispatch-finished notices; example tests use harness |
| `alonabot-plugin-help -> sdk/...`, `-> contracts`; `alonabot-agent/... -> sdk/...`, `-> contracts` | Official SDK-only access; Agent also approved MCP go-sdk and provider openai-go/anthropic-sdk-go |
| `alonabot-plugin-*/...`, `alonabot-agent/...`, `examples/plugins/...`, `plugintest -> alonabot-adapter-<adapter_id>/perm`, `-> alonabot-adapter-<adapter_id>/event`, `-> alonabot-adapter-<adapter_id>/action` | Shared Adapter permission/event/action constants/types importing only SDK root, without Adapter runtime linkage |

Forbidden edges: internal/oslock, internal/tune, internal/inproc → any module package; contracts → any module package; internal/wire → sdk or internal/core; sdk → internal/core/adapters/plugins/agent/cmd; SDK root → category task/service/bus/button/media/conversation/tool/action/options/command; internal/core → sdk/adapters/plugins/agent/cmd; adapters → internal/core; plugins/agent → internal/core/adapters (except listed constants packages)/cmd; plugins ↔ agent; internal/core/app/sdk/adapters/plugins/agent → root; internal/core/sdk/adapters/root → plugintest; all nontest code → plugintest; any package → examples. Import cycles are forbidden.

## State ownership

| State | Sole owner |
| --- | --- |
| data_dir/core.lock, state.sqlite: installs/roles/sources/generation configuration/corrections/Journal/policy/secret references/tool policy/Agent conversations/button_token/profiles/assignments/overrides | internal/core |
| WebUI components/icons/configuration example images in immutable packages | Read-only serving by Core; publisher owns contents |
| Official WebUI theme/locale/sidebar preferences in per-origin localStorage | alonabot-webui browser frontend |
| plugins/<id>/runtime.lock, managed DB connections | sdk inside plugin process |
| Plugin business data in candidate SQLite files | Plugin; Core copies/admits candidates |
| _alonabot_kv | sdk/sqlitedb creates schema, sdk/kv manages values; follows managed SQLite |
| data_dir/plugin_data/<id>/ | app creates/sends paths at startup (offline CLI test prepares temporary paths), manage purges; plugin owns files |
| Plugin-owned external DB/files | Plugin; Core does not create/validate/back up |

## File organization

Organize by responsibility, keeping types/methods/constants nearby. Shared data contracts group by subject, while implementation state retains its owner. No mandatory DTO/constants files or one-type files. [Coding standards](docs/coding-standards.md) govern contents/comments/splitting; dependency rules above remain.

## Checks

Reviewers check import direction against the table; Go rejects cycles. Compare Core/SDK root actual go list -deps linkage with [development's](docs/development.md) approved list.
