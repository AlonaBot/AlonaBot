# Plugin development guide

[简体中文](plugin-guide.zh-CN.md)

Scope: dependency, storage, logging, lifecycle and publishing rules for component authors, including official plugins, Adapters, Agents and examples, with their rationale. Adapters and Agents are separate component roles; these rules also apply to them. WebUI serves static files and may include a Go process: static releases follow section 9; processes follow the relevant SDK rules. The public behavior described here is implemented by the SDK and Core; see [architecture](../ARCHITECTURE.md) for ownership and [protocol sources](../contracts/proto/alonabot/v1) for wire types.
Update this guide when author-facing behavior, optional SDK packages or dependency policy changes.

## Start with your first command

Inside the AlonaBot source repository, plugins can reuse its SDK module. With Go ready, run from the repository root:

```fish
go build -o bin/alonabot ./cmd/alonabot
bin/alonabot init ./alonabot-plugin-hello
bin/alonabot check ./alonabot-plugin-hello
bin/alonabot run ./alonabot-plugin-hello --data-dir ./data
```

Enter `/hello` in the same terminal for `Hello AlonaBot`. Source edits reload automatically; Ctrl-C stops. Generated `New() *sdk.Plugin` is the entry point; `plugin.yaml` declares identity/source and `plugin_config_schema.json` supplies configuration Schema. Plugin/agent templates explicitly mark `.Menu()` and directly call `c.Reply("Hello AlonaBot")`; add `i18n/<locale>.json` when needed. Generated Go/plugin.yaml have no comments for any role. Change `source` to the component repository before release.

```fish
bin/alonabot test ./alonabot-plugin-hello
bin/alonabot pack ./alonabot-plugin-hello --target darwin/arm64 --out ./alonabot-plugin-hello/bin
```

`test` performs offline preparation checks; use [plugin tests](development.md#plugin-tests) for message/lifecycle integration. Set actual supported packaging targets; see [publishing](#9-publishing) for archives/distribution.

For separate repositories create your own Go module first. Release CLI writes its SDK version and tidies dependencies; a source-built development CLI prompts you to choose an accessible SDK version. See [development runs](development.md#development-runs-and-hot-reload). Existing [example projects](../examples/plugins) run directly.

## 1. Runtime model and memory

- In separate-process deployment each plugin is a Native Go process communicating with Core only through SDK. A crash affects that component. Plugins remain trusted local code requiring source review. Library-mode components share Core's process; see [library mode](development.md#library-mode-one-process).
- Disabled plugins have no running process; all enabled resident memory counts against the user's machine. Fixed linux/amd64 measurements on 2026-11-20 found echo median PSS about 10.094 MiB across three runs. Linked libraries/initialization add costs. See [release notes](../changelog/v0.1.0/en.md#resource-results-and-dependencies) for official measurements and maintainer acceptance.
- Restart-safe plugin/agent components without independent background activity may declare `sleep_capable: true`. This supports command/event/Service/HTTP waking from SUSPENDED, automatic idle exit, immediate policy changes and natural expiry wakes for static interval/cron and dynamic at/interval/cron. Plans/static origins/pause states/dynamic instances/admitted work survive same-generation resume. Static deadlines start at original due time, dynamic deadlines on acquiring concurrency; cold starts consume original JOB_TIMEOUT. Handlers hold activity references. Abnormally offline triggers follow skip rules; expired at instances delete with WARN. Development components, Adapters and WebUI remain resident.

## 2. Dependencies

- Root `sdk` provides plugins, Handler Context, registration builders, permissions, event types and runtime entry points. Import category APIs as needed: `sdk/task`, `sdk/service`, `sdk/bus`, `sdk/button`, `sdk/media`, `sdk/conversation`, `sdk/tool`, `sdk/action`, `sdk/options`, `sdk/command`.
- Import only what you use. SDK root permits protobuf, x/sys, jsonschema-go and Goldmark through the message model. Databases live in optional `sdk/sqlitedb`/`sdk/gormdb`; unused packages are not linked.
- Prefer standard libraries, e.g. `net/http`. Request models through Agent `llm.generate` (section 6), keeping vendor SDKs/custom Provider clients out of plugins.
- Avoid heavy package variables/`init`: defer large files/tables/batch regex compilation to OnStart or first use. Every linked package's `init` runs at startup, keeping allocated memory resident.
- Prefer pure Go. CGO components must supply actually runnable binaries for every declared supported platform; omit unsupported platforms from the manifest.
- Plugin dependencies use their own `go.mod`, independently of Core/other plugins; each dependency fully enters that plugin process.
- Use `sdk/message` for messages/events/sessions/handles/receipts/media, `sdk/errcode` for errors. SDK converts public API protocol types internally. Custom Protobuf services still use their own `proto.Message` with `service.Proto`/`service.Spec.Call`.

## 3. Storage

- Store business data in managed SQLite: `sqlitedb.Open(ctx)` returns `*sql.DB`, `gormdb.Open(ctx)` returns `*gorm.DB`, sharing one connection. Prefer `sqlitedb` for SQL-only use to reduce GORM linkage/initialization; see [measurement conditions](../changelog/v0.1.0/en.md#resource-results-and-dependencies).
- Use either `Migrate` or `Migrations` from your chosen package. A released helper migration ID must retain its original operation.
- Same-generation `RESUME` reuses generation/configuration/business DB, skipping SELF/VERSIONED migration callbacks while preserving mode/ledger. First activation/update/rollback still follow migration rules. Configuration still runs `OnConfigLoad`, filling/correction/required checks, with full results identical to the original snapshot.
- Reload/update use independent candidate SQLite copies; preparation failures restore old data. External files/databases have their own recovery responsibilities.
- Keep large indexes/caches/vectors on disk in managed SQLite or files under `sdk.PluginDataDir(ctx)`, rather than loading everything. Bound memory caches.
- PostgreSQL/MySQL users supply drivers and administrator-filled connection configuration marked `"secret": true` in `plugin_config_schema.json` (§7). If migrations/OnStart write external DBs, declare `prepare_effects=external`; updates to existing installations then require administrator maintenance mode. Core neither creates nor backs up external DBs.
- Use `sdk/kv` for simple state: reserved managed SQLite table `_alonabot_kv` copies/rolls back/cleans with generations. KV links SQLite, so tools infer `prepare_effects=managed_db`; reserved-table initialization is independent of plugin migration IDs.
- Write plugin files under `sdk.PluginDataDir(ctx)`. They survive restart without generation rollback. Disable/ordinary uninstall preserve them; `uninstall --purge-data`/`--purge` lists and deletes the directory after confirmation.

### Store small amounts of state

Pass the owning Session's `ctx` in Handlers/lifecycle callbacks, for example to save a subscription:

```go
import (
    "context"

    "github.com/AlonaBot/AlonaBot/sdk/kv"
)

func saveTopic(ctx context.Context, topic string) error {
    return kv.Set(ctx, "subscription/topic", []byte(topic))
}
```

Read with `value, ok, err := kv.Get(ctx, key)`. `err == nil`/`ok == false` means absent; deleting absent keys with `kv.Delete(ctx, key)` succeeds. Keys are UTF-8, 1–512 bytes, excluding exactly `.`/`..`; other dots/slashes remain literal. Values allow up to 1 MiB including empty. Invalid arguments return `INVALID_ARGUMENT`; Sessions without managed SQLite return `FAILED_PRECONDITION`.

`page, err := kv.List(ctx, "subscription/", "", 20)` returns at most 20 keys in byte order. Use `page.Next` as subsequent `cursor`; empty means end. Nonempty cursors obey key constraints. Prefixes match literal bytes, including `%`/`_`; empty lists all, and `limit` must be positive.

WebUI plugin Data pages browse/edit DB tables, KV and files, export/clear data; retained DB generations are read-only. Agent conversations offer summaries/clear operations; see [management API](deployment.md#management-api) for interfaces/cleanup semantics.

`examples/plugins/configdb` commands `/note kvset <key> <value>` and `/note kvget <key>` demonstrate writes/reads with localized replies.

### Store plugin files

Core creates the directory before startup and sends it in Bootstrap. Pass Handler/lifecycle `ctx` or a derived context; SDK uses its owning Session. In-process plugins get separate directories; separate processes use default Session if unbound. In a ready Handler, `sdk.PluginDataDir(ctx)` returns absolute `<data_dir>/plugin_data/<id>/`:

```go
import (
    "context"
    "os"
    "path/filepath"

    "github.com/AlonaBot/AlonaBot/sdk"
)

func saveCache(ctx context.Context, content []byte) error {
    return os.WriteFile(filepath.Join(sdk.PluginDataDir(ctx), "cache.json"), content, 0o600)
}
```

OnStart/OnReady precede Session readiness; calling `PluginDataDir(ctx)` there or after stop panics with `plugin session is not ready`. Unix mode is `0700`; Windows inherits the state root's account-only ACL. Core creates/purge-cleans, plugins manage files. Preparation failures/rollback preserve written files; use KV/managed SQLite for data requiring generation rollback.

configdb `/note fileset <value>` and `/note fileget` read/write fixed `note.txt` there.

## 4. Logs and text

Diagnostics collect component logs/recent errors. Keep plaintext credentials out of logs/errors: export masks only specified credential/address patterns; chat content receives no special handling. See [diagnostics](deployment.md#export-diagnostics).

- Use standard `log/slog`; SDK forwards records to Core with source identity. Keep stdout free for the stdio protocol.
- For corrupt component data or inability to continue requiring immediate action, log CRITICAL with `slog.Log(ctx, sdk.LevelCritical, "component data is corrupt", "error", err)`. SDK still logs Handler panics as ERROR.
- Logs/error strings/safe_message use English. Put localized user text in `i18n/` `messages`, with at least en-US.
- `messages` keys/named placeholders are dot-separated identifiers, each part containing ASCII letters/digits/`_`, e.g. `storeEmpty`, `account.Name`.
- `c.T(key, name, value, ...)` and background `sdk.T` return strings only. Prefer constant keys/argument names. `build/check` verify en-US keys, paired arguments, string argument names and matching placeholders. Runtime-determined keys skip static checks. Missing/mismatched runtime text returns the key or original placeholders and logs one plugin WARN.
- Keep secrets/tokens/full configuration/DSNs out of logs; redact credential-bearing URLs first.

## 5. Lifecycle and calls

- Open resources in OnStart, close in OnStop. Background goroutines end with OnStop and honor ctx cancellation.
- Ordinary deadlines are 30s. Put longer work in continuations or schedule callbacks.
- Cold start consumes original deadlines; cancelling a wait withdraws only that request. Once ready, routes/permissions are rechecked and delivery occurs once. Session activity references track actual calls/resource leases; locally cancelled work remains active until actually finished. HTTP releases on reverse-proxy return. Automatic exit enters QUIESCING, freezes dispatch, and queues new requests within original deadlines/budgets until old-tree exit/new Ready. Drain/STOPPING/OnStop/tree deadlines reuse normal shutdown, including still-running HTTP/Bus SDK callbacks.
- Continuations use CONTINUATION_TIMEOUT without occupying dispatch chains:
  - `c.Continue(func(c *sdk.Context) error { ... })`: original Handler returns/controls propagation normally; fn runs separately after successful completion, and is skipped on original error.
  - Builder `.Continue()` on commands/events/buttons: Rules run in the original call; the matched Handler runs wholly in a continuation. `StopPropagation` has no effect there. Nonempty automatic reply sends once before continuation completion.
  - In continuations, `c.Wait(timeout)` returns the next message Context from the same user/chat/topic; `Text()` reads it and ordinary dispatch skips it. This requires real inbound Actor/default Target. Nested Continue is unavailable; waits/continuations are not persisted and end on stop/replacement.
- Business calls with side effects do not retry automatically. Timeout/cancellation can occur after remote execution.
- Export `func New() *sdk.Plugin`: create `p := sdk.NewPlugin()`, register routes/OnStart/OnReady/OnStop/OnConfigLoad/migrations and return p. New can run multiple times without package state. Split one level with `p.Sub(id)`, usually subpackages exporting `Register(s *sdk.SubPlugin)`. Subplugin IDs do not enter route_id/command paths/Service names. Register lifecycle/migrations only on Plugin.
- For many/optional/defaulted arguments, accept only one struct after `*sdk.Context`. Fields map positionally; names use `arg` tags or lowercase leading capital runs (`FeedURL` → `feedURL`). Pointers are optional/nil; `default:"…"` supplies defaults; `arg:",raw"` consumes final raw tail. Mixing with positional/Args/Optional/RawTail fails during registration collection.
- Without `Args`, tools record Handler source parameter names; explicit Args wins. `_`, unnamed arguments and multiple same-line function literals (generation WARN) fall back to `argN`; explicitly name with Args before named Optional in those cases.
- HTTP uses `p.Route(method, path).Handle(h)`/`.HandleFunc(f)`; add `.Public(sdk.AuthCallbackState)` or another mode before Handle. Timers use `p.Every(jobID, d).Handle(fn)` or `p.Cron(jobID, expr).Timezone(tz).Handle(fn)`. Exactly one Handle is required, with no later options; collection reports misuse.
- Match with `p.Regex(expr)`, `p.Keyword(patterns...)`, `p.StartsWith(patterns...)`, `p.EndsWith(patterns...)`, `p.FullMatch(patterns...)`, or `p.Raw()`, also on subplugins/groups. Default route IDs use lowercase category, colon, `|`-joined patterns; Regex preserves expressions, Raw is `raw`. Events use event names; Hooks use `pre_event:<event name>`, `before_send`, `after_send`. Duplicates fail; `.ID("stable-name")` distinguishes them or preserves policy across pattern changes. Path commands still derive IDs from command paths.
- Hooks use `p.PreEvent(sdk.MessageCreated)`, `p.BeforeSend()`, `p.AfterSend()`, also on subplugins. Set stable `.ID(routeID)` before Handle.
- `p.On(eventType)` works on Plugin/SubPlugin/CommandGroup. Group events retain subplugin ownership; IDs default to event names, overridable before Handle.
- Handler `c` implements `context.Context`; pass directly to kv/task/service/llm/media/tool/command/conversation to preserve deadline/cancellation/parent association. `c.Context()` still returns the underlying context; post-Handler invalidation remains.
- Command/text/event/Service/Bus registration chains end in Handle. Missing/duplicate Handle or later options fail collection.
- Register English descriptions, e.g. `p.Command("set").Description("Set a value").Handle(Set)`. Other locales override top-level `commands` by complete canonical path including Group prefixes/single spaces, e.g. `{"commands":{"rss add":{"description":"添加订阅"}}}`. Commands-only locales may omit en-US. check/build validates structure; collection requires paths to registered commands with English descriptions, otherwise fails.
- Permissions use `sdk.Permission`: Anyone/User/Admin/ChatAdmin/PermissionName, combined with `Or`. Command/text uses nearest explicit route/group permission, default `user`; events default empty. `c.Has` checks current Actor.
- Send/Edit/automatic replies accept `message.Content`: individual `message.Text(...)`/`message.ImageURL(...)` or `message.Chain{...}`. Automatic Handlers return `(message.Content, error)`; nonempty successful results send after return and after all explicit sends. Nil/empty Chain means no automatic reply; explicit empty sends/edits return INVALID_ARGUMENT.
- Session strings are `<adapter_id>:<chat_id>[#<thread_id>]`, e.g. `telegram:-100123#42`, `onebot11:group:123456`. `message.SID(adapterID, chatID, threadID)` escapes `%`/`#`; `message.ParseSID` validates/splits. adapter_id follows local manifest IDs. `c.Send` uses default chat; `c.SendTo(sid, content)` specifies it; `sdk.Via(instanceID)` selects account, retaining default chat/topic if alone. Background `sdk.Send(ctx, session, content, opts...)` accepts only Via options. Core fills omitted instances; invalid sessions return INVALID_ARGUMENT before requests.
- Edit/delete/quote use storable string handles `<session string>/<platform_message_id>[@<instance name>]`. Get them from receipt Ref/Refs or `c.Event().Ref()`; `message.Ref(sid, platformMessageID, instance)` escapes `%`/`/`/`@`, `message.ParseRef` validates/splits. Edit via `c.Edit`/`sdk.Edit`, quote via `message.Quote`; Core fills omitted instances using target selection.
- Prefix/reply settings come from the current chat's profile. For `message.created`, Core adds configured group mention/quote/plain-text prefix to the first Send to original instance/chat/topic. Explicit/automatic/Service-derived/continuation replies share that marker; BeforeSend sees rewritten chain. If mention+prefix exceeds instance text limits, omit both, retaining quote. Every reply Send uses forward_threshold/message.forward for merged forwarding, still requiring normal validation. Background/jobs/cross-chat sends retain behavior.
- Editing the first reply preserves receipt-confirmed automatic mentions/prefixes by plugin Session/full target/physical message ID. Repeated edits retain them with UTF-16 offsets. Edit accepts one text (possibly converted Markdown) plus optional buttons; unsupported combinations of retained plain prefix/mention with Markdown edits return UNSUPPORTED. Associations release on Session exit; Core also handles decorations for c.Edit/c.Stream.
- `c.Stream(next)` consumes incremental text until `io.EOF` (possibly with final text). With message.edit it sends then throttles edits, starting new messages beyond TextLimit. Without editing, send once on normal completion; input errors discard buffered content. Return all receipts, including existing ones on failure.
- `c.Reply(text)` or `c.Reply(text, message.ImageURL(url))` sends one ordered text/segment message once, returning errors. Background `sdk.SendText(ctx, session, text, opts...)` accepts only Via, discards receipts and preserves errors.
- `c.User()` is `<adapter_id>:<external_user_id>` or empty without Actor; message.UID/ParseUID preserve external-user colons. `c.Session()` is default chat or empty; untargeted send then returns NO_DEFAULT_TARGET. Services inherit Actor/default chat with nil `c.Event()`.
- Register constants such as `p.On(sdk.MessageCreated).Handle(func(c *sdk.Context, e *message.CreatedPayload) error { return c.Reply(e.Chain.Text()) })`, or Context-only Handlers. Standard events use matching payload pointers; collection checks signatures with route_id errors. Extensions use `sdk.AdapterEvent[T]("adapter.name", version)`, struct T, Handler *T. Version/JSON failure warns and skips, logging only route/event/version or safe error category.
- `media.Upload(ctx, r, mime)` yields `*message.MediaRef`; SDK determines size/SHA-256 and sniffs empty mime. Seekable readers read twice; others use temporary files. Keep both reads identical. `media.Open(ctx, src)` accepts segment Source, MediaRef.Source, IncomingMedia.Source, or `message.MediaSource{MediaRef: token}`. It reads the first chunk before success, then returns sequential io.ReadCloser honoring ctx without retries; close it. HTTPS URL sources return INVALID_ARGUMENT before requests. `media.Grant(ctx, token, granteePluginID, sessions...)` grants other plugins access using string chats.
- Use errcode constructors such as `errcode.InvalidArgument.New("invalid input")`/`.Errorf("invalid field %s", name)`. `errcode.Of(err)` finds SDK codes through error chains or returns 0. `sdk.Error` aliases errcode.Error; errors.As still exposes Code/Message.

Declare buttons with `a := button.Define[T](name)`, register `button.On(r, a).Handle(...)` taking `func(*sdk.Context, T) error`, and send `a.Button(label, value)` inside message.Buttons. Reserve `button.` for button route IDs. Values use encoding/json, capped at 1024 bytes by Core; encoding failures return INVALID_ARGUMENT before sends. Click decode failure warns/skips Handler and completes normally. Direct message.Button with empty Value omits it; omitted-value clicks also fail decode. `message.Extension(adapterID, name, version, payload)` requires the target Adapter instance to declare that extension/version and validates its Schema.

`nodes, err := c.ReadNodes(forward)` returns recursive `[]message.ForwardItem`, preserving authors/OccurredAt/chains. Image/file sources are authorized for this call; read with media.Open(c, segment.Source). Errors may include partial nodes; handle err before using them. Background `sdk.ReadNodes(ctx, sid, forward, sdk.Via(instanceID))` respects send permissions/instance selection.

Convert generic Markdown to plain text with the object method:

```go
md := message.Markdown("**标题**\n[链接](https://example.com)")
result, err := c.Send(md) // 支持 message.markdown 的平台由 Core 转换为实体。
result, err = c.Send(md.PlainText()) // 显式普通文本，方法无需参数。
```

Generic syntax is CommonMark plus strikethrough, retaining block structure. Rich conversion with images/HTML returns UNSUPPORTED. Sending without instance message.markdown also returns it; choose .PlainText(). Existing TextSegment.FormatMarkdown retains native Adapter dialect, Telegram MarkdownV2.

In Handlers, `client, err := c.Adapter()` binds trusted instance/account. Handle errors then call typed Adapter actions. Adapter `client.From(c)` can invoke c.Adapter and check AdapterID; clients expose read-only InstanceID/AccountID/Session. No trusted platform identity returns NO_DEFAULT_TARGET; calls after Handler completion return FAILED_PRECONDITION.

`client.Call(name, version, params)`/`client.CallAction("", name, version, params)` choose declared scope: chat uses default chat, account uses bound instance/empty executor chat. Account events may acquire clients, but chat actions need explicit to. `CallAction(to, name, version, params, opts...)` accepts explicit chat/Via; cross-chat/instance calls retain send-permission checks, JSON results/errors together. All inherit parent/cancellation/deadline.

`*sdk.AdapterClient`, `*sdk.Context` and background `*action.Client` implement action.Caller. Existing c.CallAction keeps explicit targets, empty chat for account actions. Background action.NewClient(ctx) uses task-context cancellation/deadlines.

Adapters declare extensions/events via `adapter.Event[T](name, version)`/`adapter.Extension[T](name, version)`. Struct T belongs in shared Adapter constants/data packages. SDK derives Schema: non-omitempty fields required, undeclared properties forbidden. Type/derivation failures make RunAdapter error with declaration name before REGISTER.

Declare Adapter HTTP routes in adapter.Descriptor.HTTP, e.g.:

```go
return &adapter.Descriptor{
    AdapterID: "example",
    HTTP: []adapter.HTTPRoute{{
        Method: http.MethodPost,
        Path: "/webhook",
        Auth: adapter.AuthPlatformToken,
        Handler: http.HandlerFunc(handleWebhook),
    }},
}
```

Omitted Auth requires Core session; AuthPlatformToken delegates platform credentials to Adapter; AuthStatic publicly serves GET/HEAD without management/platform credentials. SDK checks paths/duplicates before connecting.

Import sdk/command and call `entries, err := command.List(c)` in a Handler. Display Path/Aliases (space-joined multi-token paths) and localized Description. Return errors; background/finished Handler contexts return FAILED_PRECONDITION. See [official help](https://github.com/AlonaBot/alonabot-plugin-help/blob/main/help.go).

### Command menus

To show a root command in a platform menu, register `p.Command("hello").Menu().Description("Say hello").Handle(hello)`. Registration otherwise only installs a handler. Automatic menus use the chat's Profile, instance permission, public permission, `/` prefix and winning effective primary name. Forum topics share the chat menu; topic switches still control execution.

For manual control, set the Telegram instance's `auto_sync_commands` to false. `clear_commands_on_disable` defaults to true: automatic-to-manual transitions first clear managed positions; false retains them until a plugin replaces them. First manual startup and manual restarts preserve menus. Subscribe to ready and set a complete menu per chat:

```go
p.On(sdk.CommandMenuReady).Handle(func(c *sdk.Context, ready *message.CommandMenuReadyPayload) error {
    if ready.ChatID == "" {
        return nil
    }
    return command.Set(c, ready.AdapterInstanceID, command.Menu{
        Scope: command.Scope{Type: "chat", ChatID: ready.ChatID},
        Commands: []command.Item{{Name: "hello", Description: "Say hello"}},
    })
})
```

Import `sdk/command` and `sdk/message`. Set each language separately: an empty LanguageCode is the default, and nonempty codes are two lowercase primary-language letters. Scopes also include `chat_administrators` and `chat_member` with UserID. Each menu accepts at most 100 commands. Invalid names, descriptions, limits and conflicting language mappings return errors while retaining previously successful menus. Each successful Set transfers ownership of the entire scope+language position; disabling or uninstalling the previous owner preserves the new owner's menu.

`command.Get` reads the current platform menu; `Delete` removes a specified position. `Status` returns mode, Ready, LastError and positions. Ordinary automatic synchronization skips unchanged successful positions in this Session. `command.Sync` fully rebuilds active-chat menus and managed default positions in automatic mode to repair remote drift; manual mode reprocesses cleanup and readiness. Read LastError after a failed transition, fix its cause, then call Sync. Core rechecks instance permission, chat Profile and global disablement before execution. It skips writes after leaving a chat and resumes on rejoin. Historical scopes written by external tools require explicit positions for Delete.

#### Filtering, translation and lifecycle

Automatic selection requires Menu, a winning single-token primary path, empty permissions or a permission set containing user, plugin availability, instance permission, the chat Profile/plugin switch, command policy and a `/` prefix. Sleeping plugins that can wake remain listed. `.Description` is English source text. Other languages override full canonical paths in the top-level `commands` object of `i18n/<locale>.json`, e.g. `{"commands":{"rss add":{"description":"添加订阅"}}}`. Group tokens use one ASCII space; aliases and effective names share the canonical path. en-US overrides fail. check/build validate structure; SDK Collect rejects paths without a registered command or English description. Per-key fallback is target locale, plugin default locale, then English; without a description use the primary name. Core maps locales to two lowercase primary-language letters and returns INVALID_ARGUMENT for conflicting descriptions after mapping. The default menu uses English. Names match `^[a-z0-9_]{1,32}$`; descriptions contain 1–256 Unicode characters. Validation precedes writes and preserves successful menus on failure.

An instance's serial worker coalesces changes for two seconds from the first change: startup, plugin installation/activation/removal/disablement, permissions, policies, Profile/assignment and prefixes. First messages and the Bot's own join publish asynchronously while ordinary dispatch continues. Event cursors and account/chat membership commit together. Mode and last successful writer persist per Adapter+account+scope+language, isolating accounts without per-command history. Known languages at managed default positions remain empty; threads do not enter positions.

Session content snapshots are saved only after platform writes and position bookkeeping both succeed. Write failures/UNKNOWN, deletion, membership changes and mode switches invalidate relevant snapshots; new Sessions and rejoins publish again. Set replaces a whole position and transfers its owner. Manual owners lose positions on uninstall, disablement, loss of instance permission or chat Profile disablement; automatic mode removes obsolete scopes/languages.

Automatic-to-manual transitions stop scheduling and drain in-flight writes. First manual startup, manual restarts and changes to clear alone preserve menus. Positions in departed chats retain the cleanup decision made when entering manual mode; rejoin resumes it, and later clear-only changes preserve that decision. Cleanup failure retains the full error and not-ready state; success emits ready. Manual-to-automatic transitions republish everything and remove obsolete manual positions. Transition intent persists as the new mode/not-ready first. Each remote success is recorded immediately; errors or unknown outcomes stop subsequent operations, storage errors return in full. Ordinary automatic failures record diagnostics; later changes or Sync may reprocess them.

Set/Delete require manual mode and readiness; plugins can use only the three chat scopes. Calls check the current Session, plugin availability and allowed_adapter_instances; chat operations also check Profile and global disablement. An empty instanceID may use an in-flight Handler's trusted instance. `bot.command_menu_ready` targets instance readiness, existing/new chats, late activation and restored permissions. `CommandMenuReadyPayload` contains AdapterID, AdapterInstanceID, AccountID and ChatID: an empty ChatID means instance manual readiness; a value means that chat is available. Callbacks can call Set directly. Platform, transport and storage errors return in full; business writes have no automatic retries.

#### Implementing Adapter menus

Declare `commands.sync` v1 supported=true and implement `adapter.CommandMenuController`: `CommandMenuConfig()` returns `AutoSyncCommands` and `ClearCommandsOnDisable`; `SetCommandMenu(ctx, menu)`, `GetCommandMenu(ctx, scope, languageCode)` and `DeleteCommandMenu(ctx, scope, languageCode)` operate on one position at a time. Telegram defaults both settings to true; implementations report typed configuration. SDK reports it in the startup response and Core calls `adapter.command.menu`. Existing Adapters without that configuration retain `CommandSynchronizer.SyncCommands`.

`CommandMenu{Scope, LanguageCode, Commands}` uses Scope fields Type, ChatID, UserID and Entry fields Name, Description. Core owns language splitting, ownership, modes and persistence. Adapters preserve whole-menu replacement, empty sets, full platform errors and UNKNOWN. SDK holds the instance execution lock throughout platform calls, even after RPC cancellation until the call returns, sharing it with legacy SyncCommands. Stop drains it before the next worker starts. Trusted member.left/member.joined events identify the Bot's own departure/rejoin by matching member ID with account ID.

Adapters/plugins using the new interfaces must update dependencies after the SDK/Core release. Add `.Menu()` to commands that should be visible, including official help's `/help` and `/whoami`.

### Trusted media MIME and platform helpers

`message.MediaSource.MIME` is trusted source metadata filled by Core while preparing sends. MediaRef/File sources use saved metadata; same-instance incoming media forwarded as `platform_file` uses the trusted inbound MIME. Empty means unknown. Plugin-supplied MIME returns INVALID_ARGUMENT; `MediaRef.Source()` carries only the reference. Leases and SDK conversion preserve MIME; `adapter.Core.ReadMedia` keeps its signature. Telegram uses `image/gif` for animation, checks inline content for the `GIF8` prefix, and treats HTTPS URLs as ordinary images without MIME inference.

Adapter public perm/event packages declare permissions/events using SDK builders. Public action helpers call through `sdk/action.Caller` and may directly import `sdk`, `sdk/action`, `sdk/message` and `sdk/errcode` for typed options, sends, handles, streaming helpers and errors. These public packages keep Adapter runtime code out of plugin binaries. Telegram typing helpers refresh every four seconds until context cancellation; draft helpers finish with an ordinary Send. Draft actions bypass outbound Hooks; the final Send runs them.

## 6. Cross-plugin collaboration

- Use Service Calls/Bus, respecting other plugins' file/database ownership.
- Share package-level declarations: `var Search = service.JSON[SearchIn, SearchOut]("weather/search")` (Protobuf service.Proto); provider `service.Serve(p, Search).Handle(fn)`, caller Search.Call(ctx,in). Events: `var Updated = bus.Define[Update]("weather/updated")`, Updated.Publish, bus.Subscribe(...).Handle. Subscription IDs default to full event name; use .ID for duplicates. Target/event component namespace must match this plugin ID, checked after handshake/before REGISTER. Only Agent should Serve `@agent/<name>`: SDK checks format, Broker routes only to role: agent. Imported declarations work directly or through wrappers; caller generation embeds input/output Schemas.
- Bus broadcasts are nonpersistent. Publish only in your namespace; success means Core admission. Full subscriber queues/budgets/offline state drop events without replay. Use Service for receipts/reliable delivery. Callbacks execute serially; blocking backlog consumes plugin receive budget and can reject commands. No Actor/default chat is supplied; send with explicit sdk.Send/sdk.SendText sessions.
- service.JSON/bus.Define/task.Define require package-level named nongeneric types, checked with source diagnostics. JSON Service/Bus schema_id derives from type, version from SchemaVersion() uint32 or default 1. Rebuild after type changes. Slices/maps accept null matching encoding/json nil.
- Invalid JSON Service inputs return INVALID_ARGUMENT / `invalid service input` without calling Handler. Invalid output returns INTERNAL and logs only schema_id. Caller Call validates responses too, returning INTERNAL / `invalid service output for "<target>"`. Validation details can include business payloads; omit them from responses/logs.
- tool.List(ctx) returns catalog/declared Info.Effect. Current trusted default Target selects profile; no Target/parent uses default. Each list/invoke/confirm snapshots once: disabled plugin denies, then profile override/global override/declaration defaults; confirmation operations default confirm without override. Same-named tools/operations match independently. Updates affect next operation; in-flight confirmation retains admitted policy. Register JSON tools with .Tool(description, tool.DefaultAllow/DefaultConfirm). Between Tool and Handle, optional .Effect(tool.ReadOnly), Write, Destructive; default Unspecified. Invalid order/category fails collection; effects are hints without changing policy.
- `sdk/llm.Generate(ctx, llm.Request{...})` sends a structured text generation request to `@agent/llm.generate`. Administrators grant access through `llm_service_allowed_plugins`. `sdk/agent.Run` targets the separate `@agent/agent.run` contract. Streaming entry points require negotiated transport capability. Full parameters and runtime status are in the [Agent API](agent-api.md).
- In Handlers/continuations, `tool.Invoke(ctx,"plugin_id/service_name",args)` follows policy: deny → PERMISSION_DENIED; confirm sends actual arguments with Allow/Deny buttons, denial → PERMISSION_DENIED, timeout → DEADLINE_EXCEEDED. Before your own operations, `tool.Confirm(ctx,operation,summary,effect)` displays your summary; nil permits execution. Use Unspecified if unknown; invalid categories return INVALID_ARGUMENT. Both need parent chains; required confirmation without Actor returns PERMISSION_DENIED, without buttons UNSUPPORTED.

### Plugin pages

`p.Route("GET", "/subs/").Page("Subscriptions").Handle(h)` lists a page. Page precedes Handle/HandleFunc with English title; `{"pages":{"/subs/":"订阅"}}` in i18n/zh-CN.json translates it. Core exposes active pages/titles through component listings.
WebUI hosts sandboxed iframes and the `postMessage` bridge. `.Page(title)` registers `kit.css`, `kit.js`, and `qrcode.js` beside the page as public AuthStatic GET routes. Relative imports provide shared styles/theme/locale and ⌘K/Ctrl+K global search. WebUI draws the first row (plugin/tabs/search); omit duplicate title/search bars inside pages.

Official Agent's five pages live in [pages.go](https://github.com/AlonaBot/alonabot-agent/blob/main/pages.go) and pages/: personas/MCP/skills/memory/user tasks use component HTML/CSS/JS, titles in pages, text in messages and bridged component management routes. Memory pages paginate full conversations/manage notes/profiles; Core conversation summaries/clear use [management endpoints](deployment.md#management-api).

#### Mount and import the kit

Declare the page in New():

```go
import (
    "net/http"
    "github.com/AlonaBot/AlonaBot/sdk"
)

func New() *sdk.Plugin {
    p := sdk.NewPlugin()

    p.Route(http.MethodGet, "/page").Page("Showcase").HandleFunc(servePage)
    return p
}
```

Relative assets resolve with page URLs: /page uses /kit.css,/kit.js; /page/ uses /page/kit.css,/page/kit.js under the plugin prefix. Existing same-method/path author routes win; page.Handler() also mounts manually. Opaque-origin sandbox subresource requests omit management cookies: register your static assets .Public(sdk.AuthStatic), GET/HEAD only, to avoid 401. HTML/data remain management-authenticated; window.AlonaBot.request asks WebUI to fetch data. See [showcase](../examples/plugins/showcase/main.go).

#### Page HTML and controls

Reference kit assets in returned HTML:

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <link rel="stylesheet" href="kit.css">
  <script src="kit.js"></script>
</head>
<body>
  <div class="ab-card">
    <div class="ab-card-header">插件管理</div>
    <div class="ab-card-body">
      <p id="greeting"></p>
      <button class="ab-btn ab-btn-primary" id="btn-save">保存</button>
      <input type="text" class="ab-input" placeholder="请输入...">
    </div>
  </div>
</body>
</html>
```

kit.css uses only --ab-* tokens for buttons (button,.ab-btn,.ab-btn-primary,.ab-btn-danger), inputs (input,textarea,.ab-input), selects (select,.ab-select), tables (table,.ab-table), labels (label,.ab-badge), cards (.ab-card). Light defaults apply until tokens arrive; theme changes update immediately.

#### Client interface: window.AlonaBot

kit.js injects window.AlonaBot globally:

- **request(method,path,body?)** calls this plugin's routes through WebUI. Objects send as JSON. Promise results include status/ok (2xx)/headers/body (UTF-8)/data (parsed JSON or text)/original body_base64, with text()/json() aliases for body/data. Bridge errors reject; HTTP errors remain inspectable results.
- **t(key,args?) → Promise<string>** sends translate with key/optional string arguments on every call and returns bridge text unchanged. Current UI locale formats {name}; unknown keys return key, missing arguments retain placeholders. Bridge failure rejects; standalone pages resolve to key.
- **Account configuration:** `getConfigContext()` and `completeConfig({patch})` send `config.context` and `config.complete` respectively. See [Adapter account configuration pages](#adapter-account-configuration-pages).
- **onTheme(fn)** subscribes fn(theme,tokens), immediately delivers current state and returns unsubscribe.
- **onLocale(fn)** subscribes fn(locale), immediately delivers current locale and returns unsubscribe; rerender by awaiting t in the callback.
- **Shortcuts** capture ⌘K on macOS/Ctrl+K elsewhere and forward mod+k for global search.
- **Standalone pages** apply light defaults/fonts; request rejects explaining the bridge requires an iframe.

Declare `"hello_message": "Hello, {name}!"` under messages in i18n/en-US.json and matching translations elsewhere. Await text before updating on locale changes:

```js
AlonaBot.onLocale(async function(locale) {
  document.documentElement.lang = locale;
  try {
    document.getElementById('greeting').textContent =
      await AlonaBot.t('hello_message', { name: 'Alona' });
  } catch (err) {
    console.error('Translation failed:', err);
  }
});
```

### Adapter account configuration pages

When adding or editing an account, WebUI can open an Adapter page in the configuration dialog for QR login, an authorization redirect, or token validation. Register a GET route in `adapter.Descriptor.HTTP` and leave `Auth` at its zero value to require a management session. Each Adapter may declare one account configuration page, with a required `Title`.

```go
return &adapter.Descriptor{
    AdapterID: "example",
    HTTP: []adapter.HTTPRoute{{
        Method: http.MethodGet,
        Path: "/account/",
        Handler: http.HandlerFunc(serveAccountPage),
        Page: &adapter.PageDescriptor{Title: "Connect account", Purpose: adapter.PageInstanceConfig},
    }},
}
```

Core takes the title from `RouteDescriptor.page_title` and publishes `{"path":"/account/","title":"Connect account","purpose":"instance_config"}` in the component listing's `pages` array. An empty `page_purpose` means a regular page; the listing omits `purpose` in that case.

After loading `kit.js`, call the `alonabot.plugin` bridge:

| Method | op | Request arguments | Success result |
| --- | --- | --- | --- |
| `AlonaBot.getConfigContext()` | `config.context` | `{}` | `{mode: "create" \| "edit", instance_id: string, config: object, schema: object}` |
| `AlonaBot.completeConfig({patch})` | `config.complete` | `{patch: object}` | `{}` |

`config` is the current form configuration with secrets redacted; `schema` is the Adapter instance configuration schema. WebUI merges the patch into the form using RFC 7396, and the user saves it in the dialog. Omitted fields keep their values; `null` removes a field. Supply new or replacement secrets in plaintext, omit a secret field to retain its value, and never submit `$secret_set` markers. When the host has no account configuration context, both calls reject with an `Error` with `code: "FAILED_PRECONDITION"`.

The initial flow covers logins that produce long-lived static credentials. The Adapter keeps credentials that need runtime refresh in its own component storage, keyed by instance ID; configuration stores only the account identifier. In `create` mode, `instance_id` is a draft name: never persist data under it. Generate a random ID in the Adapter for each login task. Keep re-login in component HTTP routes so it works while the instance is stopped. Use short polling at about two-second intervals to check status.

For QR login, load `qrcode.js` separately after `kit.js` from the same directory:

```html
<link rel="stylesheet" href="kit.css">
<div id="login-code" role="img" aria-label="Login QR code"></div>
<script src="kit.js"></script>
<script src="qrcode.js"></script>
```

Once the platform returns the login text, call `AlonaBot.renderQRCode(document.getElementById('login-code'), loginURL, {size: 256})`. It encodes UTF-8 bytes, selects the version automatically, uses level M error correction, and replaces the container contents with a themed SVG.

## 7. Configuration and secrets

Go source filenames are unrestricted: the package-level `New` and its signature define the entry point. Duplicate `New` declarations in one package produce Go's `redeclared` error. Plugins and Agents use `plugin_config_schema.json`; Adapters retain `adapter_config_schema.json`.

For upgrades, disable old components, back up data and upgrade Core, update component packages rebuilt with the new SDK, then enable them. Configuration values, secret references and managed data retain their existing storage. Disabled old packages remain updateable through the management API.

Place plugin_config_schema.json beside plugin.yaml, declaring types/English source text by field key. alonabot build validates/embeds normalized JSON through overlays; check validates the same. For example:

```json
{
  "greeting": {"type": "string", "description": "Greeting", "default": "Hello"},
  "enabled": {"type": "bool", "default": true},
  "token": {"type": "string", "description": "Access token", "secret": true}
}
```

- Configuration has no version/migration. Added fields get defaults; deleted/renamed old keys remain invalid until administrator cleanup. WebUI displays corrections/cleanup. Wrong types convert losslessly first, otherwise restore defaults/notify administrators.
- To move old values/check business rules, register p.OnConfigLoad(fn), signature `func(ctx context.Context, raw json.RawMessage) (json.RawMessage, error)`. raw is full original configuration before defaults/corrections. Check old-key presence, preserve other keys, honor ctx and perform repeatable pure transformations. Without callback, keep raw unchanged.
- SDK then fills missing/null fields with default/type zero, recursively corrects objects/lists/templates, tries lossless type conversion, restores defaults outside options/slider/numeric bounds and removes invalid template_list elements. Record every path/reason/original value except secrets. Unmet required fields fail preparation. Preserve unknown keys/numeric precision.
- GET /api/v1/plugins/{id}/config corrections describes Active; unknown lists invalid paths. POST .../config/prune removes invalid Active keys into Pending, then reload applies. Ordinary restart preserves snapshot; changed callback/correction results return FAILED_PRECONDITION.

From OnStart, p.Config() (Handler c.Config()) returns config.Values; subplugin s.Config() reads its same-named top-level object. Calling during/before OnConfigLoad panics. p.Config().String("greeting") reads the example. Base values stay fixed per Session; c.Config freezes the invocation profile. Both support concurrent reads; collection getters return independently mutable copies.

Mark top-level/object children `"profile": true` for per-profile overrides. Prefer leaves that vary by chat. Overriding composites requires the whole subtree, list elements/all template branches included, to contain no secret. Keep OnStart/background tokens/connections/caches unmarked. c.Config returns whole-plugin read-only effective chat configuration, stable within Context. Rules/Services/Hooks/continuations also freeze when constructing Context; override notifications affect subsequent calls. p.Config/s.Config/OnStart/background retain base values.

Overrides replace full paths on loaded/filled/corrected base. Objects/lists replace wholly; replacing ancestor objects changes descendant values. `config.Values.WithOverrides(overrides map[string]json.RawMessage) (config.Values, error)` creates independent views retaining subpaths, validating single values and preserving base load/correction results.

`sdk.ProfileOverrides(ctx context.Context) (map[string]map[string]json.RawMessage, error)` returns deep-copied maps/RawMessages keyed by profile ID then field path, empty map without overrides; p.Config reads base. Pass owning Handler/HTTP/lifecycle ctx; in-process plugins isolate by Session, separate processes default if unbound. Missing Session/plugin returns FAILED_PRECONDITION. Reads see currently published collections; notifications change new reads while existing copies/c.Config retain values. Official Agent personas.go combines base/core.profile.list for direct/inherited persona references.

Core provides core.profile.list/current/assign/set. current/assign/set use trusted chat/Actor from same-Session in-flight parents; assign/set require Core Admin, set touches only this plugin's profile-marked fields. default writes Pending; existing Pending/management contention returns FAILED_PRECONDITION. After root and continuations finish, reload triggers with response reloading=true. Disabled profile plugins deny commands and skip text/event/Raw/PreEvent candidates.

Official Agent profile.go demonstrates four package-level service.Proto declarations, passing c, json.Marshal values and replying with profile names/chat counts/reloading. default replies finish in the current Handler before root-end reload. Agent conversation/jobs/llm.generate without explicit model read overridable c.Config fields.

| Declaration | Getter |
| --- | --- |
| string, text | String(key) |
| int, float, bool | Int(key) (int64), Float(key) (float64), Bool(key) |
| list without items | Strings(key) |
| object, list with items | Object(key), Objects(key), then child keys |
| dict | Dict(key), returning map[string]json.RawMessage |
| template, template_list | Template(key), TemplateList(key); Key() returns selected template, embedded Values reads fields |

Missing optional templates mean no selection and Key() returns empty; selected values contain exactly one declared template key. Undeclared keys/type-mismatched getters panic with field paths.

Secrets support string/text only: set secret=true, omit default/options. Core extracts secrets recursively through objects/lists/templates. Nonempty values become immutable snapshot references, resolved to plaintext in Bootstrap/Config(); empty means unset. Keep plaintext out of logs/errors/business DB.

Configuration translations live under config in i18n/<locale>.json, following schema field hierarchy:

```json
{"config": {"greeting": {"description": "问候语"}, "token": {"description": "访问令牌"}}}
```

Override description/hint/labels; nested fields use items, templates use templates with name/hint. Stored options remain unchanged; labels translate. Generation checks keys against schema fields/templates/options with file/path diagnostics.

- For runtime choices (instances/plugin entries), use options_from instead of options: `{"type":"string","options_from":{"source":"core.adapter_instances","filter":{"adapter_id":"telegram"}}}`, or plugin source `<plugin ID>/<service name>`. Register `options.Serve(p,"feeds",func(ctx context.Context,q options.Query)(options.Page,error){...})`: search q.Search, resolve labels by stored q.Values, return at most q.Limit with More=true if needed. SDK fixes input/output Schema without generation. Core calls only for management forms; access is PRIVATE. Values stay ordinary strings; unavailable sources/stale values still load, and plugins handle unknowns.

### Reply examples for configuration fields

Add example to a field to show its reply. One object or an array supports one/multiple examples. This combines a greeting with a packaged image:

```json
{
  "greeting": {
    "type": "string",
    "default": "Hello",
    "example": {
      "input": "hello",
      "reply": [
        {"text": "{greeting}"},
        {"image": "examples/greeting.png", "alt": "Greeting illustration"}
      ]
    }
  }
}
```

Required input is user text without command prefix; required reply accepts string or ordered segments, each with exactly text or image and optional image alt. Put PNG/JPEG/WebP/GIF images in project examples/, matching extension/content. check/build/pack validate references; pack collects referenced images; install/update revalidate.

Use {field.path} in input/text for draft string/text/int/float/bool, e.g. {group.greeting}; labeled options display current locale labels. Templates use {mode.$key} or {mode.template_key.child}; unselected branches render empty. Secret/list/dict/object/template_list cannot supply references. Escape literal braces as {{/}}.

Optional condition matches all path/value pairs, e.g. `"condition":{"mode.$key":"simple","enabled":true}`. Hiding the field by its own display condition also hides examples. Examples declare fixed replies; WebUI substitutes drafts/renders.

Translate under config.<field>.example in i18n/<locale>.json, preserving example object/array and reply string/segment-array shapes. Positionally override input/reply strings/segment text/alt; omitted entries fall back, `{}` skips intermediate array items. check/build/pack validate keys/types/shapes/positions/placeholders.

After packing/installing/activating, authenticated `GET /api/v1/packages/{id}/examples/greeting.png` serves only plugin_config_schema.json-referenced images in Active, with the same security headers as icons. See [management API](deployment.md#management-api).

## 8. Tests

- plugintest starts real Core/test Adapter in go test; inject messages/check sends. See [plugin tests](development.md#plugin-tests).
- Repository official plugins also follow [coding standards](../CONTRIBUTING.md) test-submission rules.

## 9. Publishing

- Marketplace role prefixes: alonabot-plugin-, alonabot-adapter-, alonabot-agent-, alonabot-webui-. Suffix words use single hyphens, no underscores. Plugin regex is `^alonabot-plugin-[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`; other roles match that shape; maximum length 64. Local rules remain `^[a-z][a-z0-9_-]{0,63}$`. Manifest icon supports SVG/PNG/WebP.
- Renames update plugin.yaml/dependencies/HTTP paths/raw Service names; pack uses new IDs for archives. Rebuild old development data using [development runs](development.md#development-runs-and-hot-reload).
- Releases may include UTF-8 changelog exposed by [management API](deployment.md#management-api). Choose root CHANGELOG.md with version headings (e.g. `## [1.2.1] - 2026-11-17`, date optional), or version/locale files such as changelog/1.2.1/en-US.md and zh-CN.md. Directories use SemVer without v, at most manifest version; filenames use canonical BCP 47 tags and every version requires en-US.md. check validates; pack preserves original text.
- WebUI defaults plugin sources to AlonaBot/plugin-store, supporting add/remove/persistently empty lists. It reads default-branch index.json and the manifest inside component latest-Release platform archives for display. See [contracts/pluginindex](../contracts/pluginindex) and [contracts/release](../contracts/release) for source/release formats. Update previews use bound-source Release bodies; installed updates use packaged changelog. HTML displays as text; backend persists monotonically increasing seen SemVers.
- Install community components with install --source owner/repo; later updates stay bound there. Local install <archive> reads archives; install <component ID> reads adjacent packages/<component ID>/. No-argument discovery uses current directory (author bin/, including direct package children), archive names to select one host candidate; static WebUI uses any/any. Zero/multiple candidates error/list paths. Same-category Adapter, Agent and WebUI cannot coexist; replacing official components requires --replace.
- `alonabot build` validates/compiles; `alonabot check` is read-only. `alonabot pack [directory] --target linux/amd64 --target linux/arm64` outputs platform tar.gz archives only. The default target is the host; static WebUI creates one any archive, and Go-entry WebUI builds per platform. `--tags` selects source; `--out` defaults to `<project>/bin`. Upload these archives to Releases; GitHub supplies source archives. Process packages include parseable `build-info`; static WebUI includes manifest, assets and optional changelog.

## Direct testing and builds

`alonabot init alonabot-plugin-hello` defaults ID/external module to directory name, reusing existing go.mod. Override --id/--module/--role; Adapter --adapter-id defaults to component ID. my_plugin creates local IDs with missing-prefix warnings; invalid names request explicit --id.

build/check/test/pack accept at most one positional directory, default current; --dir is unknown. Run go test ./... directly with plugintest. Startup scans manifest/Schema/i18n/Handler argument names by entry package, sharing read-only resources. New only creates/registers behavior; author projects contain no alonabot_gen.go/generation code/directives.

Separate components use alonabot build .; main packages assembling via alonabot.Run/Start use alonabot build ./cmd/bot, producing cmd/bot/bin/bot (.exe on Windows). Library builds read-only scan dependencies including third-party module cache and embed resources via main overlays. Plain Go binaries prompt build/run for resource-missing components. See [library mode](development.md#library-mode-one-process) for options/stopping; separate-process hot development uses alonabot run --reload.

## Agent, streaming services and lifecycle API

[Complete parameters, capability status, eight hook stages and Goal recovery](agent-api.md).
