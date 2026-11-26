# AlonaBot design guide

[简体中文](design.zh-CN.md)

This guide is for developers reading the AlonaBot source for the first time. It walks through the current design in the order components → processes → protocol → messages → state → trust, and explains why each part works the way it does. The linked documents and source files are authoritative for details; update this guide when a core mechanism changes. An animated step-by-step walkthrough is on the documentation site's [architecture page](https://alonabot.github.io/en/docs/internals/architecture).

## In one sentence

Core is the only host and routing hub. Platform connectivity (Adapters), features (plugins), AI (Agents) and the management interface (WebUI) are installable, replaceable components that run in separate processes by default and talk only to Core, over one protocol.

## Components and roles

| Role | What it does | Official implementation | Processes |
| --- | --- | --- | --- |
| Core | Starts and stops components; identity and permissions, event dispatch, sending, media, configuration, scheduling, storage generations, management API | This repository | One |
| Adapter | Connects one platform's accounts to Core, reports events, performs sends and platform actions | Telegram, OneBot 11 | One per Adapter component, hosting all of its account instances |
| Plugin | Registers commands, events, hooks, services, scheduled tasks and HTTP routes | help | One per plugin |
| Agent | Model providers, conversations and tool calls; runs like a plugin | alonabot-agent | One |
| WebUI | Static management interface assets, optionally with a backend process | alonabot-webui | None when static only |

Each component package declares its role, identity, version and source in `plugin.yaml`; Core parses it only through [`contracts/manifest`](../contracts/manifest). Component Go code depends only on the public SDK ([`sdk/`](../sdk)) and protocol packages, never on Core internals; allowed dependency edges are in the [architecture map](../ARCHITECTURE.md).

**Why**: platforms, features and AI change at different speeds. As separate components they release and upgrade independently and can be swapped for community implementations; Core only has to keep the protocol stable.

## Process model

- **One process per component**: Core starts each component process, completes the handshake and keeps monitoring it. A crash leaves Core and other components running. Abnormal exits restart automatically; after `RESTART_MAX_CRASHES` (3) abnormal exits within `RESTART_WINDOW` (5 minutes) the component is disabled. Core never retries business calls automatically. See [`internal/core/supervisor`](../internal/core/supervisor).
- **Two transports**: by default Core holds a loopback TCP listener (`127.0.0.1:0`); stdio is also supported. Every start creates a new Session and hands the component a one-time 32-byte secret to authenticate with.
- **Hibernation**: with `sleep.mode = component_idle`, idle plugins exit and the single Waker restores the same generation when a command, event, service call, HTTP request or scheduled task needs them, reusing configuration and database. Under the automatic policy, plugins that declare resident routes (hooks, HTTP, Bus subscriptions and others) do not hibernate; an explicit per-plugin idle timeout overrides that. Plugins with frequent cold starts are pinned resident for a while. See [`internal/core/idle`](../internal/core/idle) and [plugin hibernation](management-api.md#plugin-hibernation).
- **In-process mode**: Core and components can also be compiled into one Go program (library mode), sharing assembly code with the multi-process mode in [`internal/app`](../internal/app). [`plugintest`](../plugintest) uses it to start a real Core in `go test`. An in-process component crash takes the whole program down, so this mode is for tests, debugging and embedding.

**Why**: plugins are arbitrary Go code. Process isolation keeps one plugin's crash, memory leak or dependency conflict from spreading, and makes memory use visible per process. The cost is cross-process communication and the need for a handshake and call protocol.

## Protocol and Sessions

The protocol is defined in [`contracts/proto/alonabot/v1`](../contracts/proto/alonabot/v1); Core and the SDK share the [`internal/wire`](../internal/wire) implementation.

**Framing**: each frame is a 4-byte unsigned big-endian length followed by a Protobuf [`Envelope`](../contracts/proto/alonabot/v1/envelope.proto), with a 4 MiB limit per frame. The Envelope carries the frame kind, Session ID, call direction and call ID, deadline, trace ID, target, payload with its schema ID, and error information. A zero-length, oversized or undecodable frame fails the Session.

**Handshake**: a component becomes available in a fixed order:

1. `HELLO`: the component presents its identity and secret; both sides negotiate protocol version and capabilities.
2. `HELLO_ACK` with `Bootstrap`: Core sends private inputs such as the generation, candidate configuration, database descriptor, log level, phase and Session limits.
3. `REGISTER`: after configuration preparation and `OnStart`, the component submits all of its routes once; Core validates the whole set atomically.
4. `PREPARED`: Core confirms the persisted snapshot.
5. `ACTIVATE`: Core notifies the component after committing the single Active generation in its state database.
6. `ACTIVATED`: the component's `OnReady` succeeded; business calls begin.

**Calls**: calls match on `(Session, initiating side, call ID)` and carry absolute deadlines. A sub-call a plugin makes while handling a Core request attaches to that in-flight request with `parent_call_id`, inherits its trusted context, and can only keep or shorten its deadline. Cancellation is cooperative; side effects that already happened stay. Payload and error can both be present to report partial success, and callers keep both.

**Why**: registration is fixed once at startup while permissions and enablement are decided by Core at run time, so Core knows every component's declarations before dispatching. The same framing serves TCP, stdio and in-process mode, so the SDK does not care how it runs.

## Events, messages and routing

**Only Core is trusted for identity**: an [`Event`](../contracts/proto/alonabot/v1/message.proto) describes a platform event, `Actor` is who triggered it, and `Target` identifies one conversation or topic under one Adapter instance ([`identity.proto`](../contracts/proto/alonabot/v1/identity.proto)). When an Adapter reports an event, Core rewrites its event ID, source and platform permissions. `Actor` comes only from trusted Adapter metadata and Core policy; identities claimed by plugins are never used. Admin status comes only from Core settings.

**How one message is processed** ([`internal/core/router`](../internal/core/router)):

1. The Adapter submits the event; Core checks instance ownership, persists the receive cursor and rebuilds trusted fields.
2. The Router admits the event against count, byte and global budgets; when they are exhausted it returns queue full and the Adapter records a drop.
3. Session access is checked against the profile, and user/admin roles are computed.
4. PreEvent hooks may rewrite the business payload or stop dispatch; a multi-turn conversation waiting for the next message takes a matching message first.
5. Commands are matched with the conversation's effective prefixes and arguments are bound.
6. A hibernating plugin is woken; after waking, Core re-checks generation, route and permission, issues media tokens for the receiver and calls the Handler.
7. When the Handler replies, the message service checks send policy, runs BeforeSend hooks, reserves AfterSend capacity and then calls the Adapter send once.
8. AfterSend hooks observe the result as a separate task and never delay the reply.

**Route kinds**: command, event, hook, service, schedule, HTTP and Bus subscription ([`route.proto`](../contracts/proto/alonabot/v1/route.proto)). A `STOP` from the command layer ends the whole dispatch chain; event routes run one by one and each Handler chooses to continue or stop propagation.

**Hook boundaries**: PreEvent may rewrite only the event payload, never identity; BeforeSend may rewrite only the message chain, which is then fully re-checked; AfterSend only observes. Outbound hooks cannot send messages, which prevents recursion.

**Cross-plugin collaboration**:

- **Service calls**: synchronous calls with results, forwarded by [`internal/core/broker`](../internal/core/broker) to the target plugin's current Session, waking it on demand.
- **Event Bus**: publishers use only their own namespace and Core returns as soon as it admits the event; no persistence, acknowledgement or redelivery, so offline subscribers miss it.

**Media**: media never travels inside ordinary calls. Uploads go in 1 MiB chunks and are checked against SHA-256 on commit, producing a `MediaRef` bound to its owner Session and consumer scope; grants can only stay equal or narrow ([`media.proto`](../contracts/proto/alonabot/v1/media.proto)).

**Why**: every path goes through Core, so permissions, limits and auditing live in one place. The protocol fixes what hooks can change, so plugins cannot forge each other's identity or send targets.

## State, generations and updates

**State ownership**: Core metadata (installs, generations, policies, profiles and more) lives in `state.sqlite`, read and written only by Core. Each plugin's business data lives in its own managed SQLite and is read and written by the plugin; Core copies candidates, admits them, and offers browse, export and clear operations to the management interface. The full list is in the [architecture map](../ARCHITECTURE.md#状态归属).

**Generation**: one committed combination of program, configuration and data. Updates and configuration reloads go through a candidate generation ([`internal/core/journal`](../internal/core/journal), [`internal/core/candidate`](../internal/core/candidate)):

1. Stop the old process, confirm its process tree exited, and copy the managed database into a candidate.
2. Start a new Session on the candidate and bring it to `PREPARED`.
3. At the single commit point, switch the Active generation and record COMMITTED in one transaction, then activate.
4. If any step fails before commit, discard the candidate and resume the old generation; if a failure happens after commit, keep the new generation marked Degraded for diagnosis.

Retained older generations can be rolled back to together with their data.

**Installation**: a package is preflighted (controlled extraction, manifest, platform and build-info checks) and then moved into an immutable package directory. Online installs bind the source repository and verify the archive against the digest from the GitHub Release; later updates keep checking source and identity ([`internal/core/packages`](../internal/core/packages), [management reference](management-api.md#component-installation)).

**Why**: a failed update returns to a complete, consistent old state instead of "new program, half-migrated data". Core state schema changes are added as ordered migrations; see [coding standards](coding-standards.md#versioning-and-compatibility).

## Trust boundaries

- **Plugins are trusted code**: they run as the same OS user as Core and can do anything that user can. Process isolation limits faults; it is not a security sandbox, so install only components you trust.
- **Plugin pages are untrusted**: plugin HTTP responses get a forced sandbox CSP, Core strips cookies the plugin sets, and management credentials are removed before forwarding.
- **Platform users are untrusted**: identity and permissions come only from trusted Adapter metadata and Core settings.
- **Management entry**: the management API is served only under the `base_url` host and uses the management token, cookies and CSRF; when listening on a non-loopback address, `base_url` must be https or trusted reverse proxies must be configured. See the [management reference](management-api.md#core-settings-and-http-entry-point).

## Further reading

| Topic | Read |
| --- | --- |
| Module responsibilities, allowed dependencies, state ownership | [Architecture map](../ARCHITECTURE.md) |
| Protocol fields and message structures | [`contracts/proto/alonabot/v1`](../contracts/proto/alonabot/v1) |
| Build, generation, verification and CI | [Development](development.md) |
| Management API, settings, hibernation and download mirrors | [Management reference](management-api.md) |
| Writing components with the SDK | [Plugin guide](plugin-guide.md), [Agent API](agent-api.md) |
| Coding and test rules | [Coding standards](coding-standards.md) |
