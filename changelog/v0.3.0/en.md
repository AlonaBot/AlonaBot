## [0.3.0] - 2026-11-23

### Added

- Docker deployments hand out the management token on the first visit from the host browser: inside a container (`ALONABOT_CONTAINER=1`) Core also trusts the default gateway as a source, with all other checks unchanged. Sign-in can be kept on the device. See [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312).
- Install components by building from GitHub source (`Build: true`). The Go toolchain is downloaded into the data directory on demand; an asynchronous task reports steps and build logs, and the result still passes install preflight and source trust. See [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312).
- Download settings: GitHub download mirror prefix, API base, Go toolchain download base, GOPROXY and network proxy. Release assets fetched through a mirror are verified against the sha256 digest GitHub reports. See [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312).
- Optional session access in profiles: all sessions, only selected sessions, or all except selected sessions. Behavior is unchanged when unset, and library mode is unaffected. See [PR #315](https://github.com/AlonaBot/AlonaBot/pull/315).
- Adapters can declare an account configuration page (`page_purpose = instance_config`) that WebUI opens when adding or editing an account; the page returns a configuration patch through the bridge. The page kit adds `getConfigContext`, `completeConfig` and an on-demand `qrcode.js`. See [PR #317](https://github.com/AlonaBot/AlonaBot/pull/317).
- Live Adapter action catalog, a management-authenticated call API, the `sdk/debug` client and `alonabot adapter instance list/actions/call`. plugintest supports multiple Adapters and instances, configurable platform results, asynchronous Submit/Wait and event injection through management authentication. See [PR #310](https://github.com/AlonaBot/AlonaBot/pull/310).
- `alonabot --version` prints the version, build date and Release URL. See [PR #309](https://github.com/AlonaBot/AlonaBot/pull/309).
- The SDK adds `OnSleep`, structured `llm.Generate`, `agent.Run`, AI hooks, Goal interfaces and a streaming service protocol. See [PR #308](https://github.com/AlonaBot/AlonaBot/pull/308).

### Changed

- Plugin and Agent configuration Schemas are now `plugin_config_schema.json`; Adapters keep `adapter_config_schema.json`.
- Command builders are split by matcher kind; only path commands offer `Menu`, arguments and aliases. `llm.Generate` replaces `Ask` and the old single-shot entry points. See [PR #308](https://github.com/AlonaBot/AlonaBot/pull/308).
- Managed source builds compile serially with `GOMEMLIMIT` and `GOGC`. On Linux the kernel prefers to kill the compiler rather than Core when memory runs out. Build logs warn about low available memory or long silence, and a killed build reports a possible out-of-memory failure with alternatives. See [PR #316](https://github.com/AlonaBot/AlonaBot/pull/316).
- The package list includes names and descriptions, and account instances report their component's state. See [PR #312](https://github.com/AlonaBot/AlonaBot/pull/312).

### Fixed

- Updating a Degraded component no longer stops the new generation right away. See [PR #316](https://github.com/AlonaBot/AlonaBot/pull/316).

### Upgrade and compatibility

- Components built with the new SDK require Core v0.3.0. Stop Core and back up the whole data directory before upgrading; disable old components, upgrade Core, then update each component to 0.3.0 and enable it. Configuration values, secrets and managed data are retained.
- The state database migrates to v3 automatically (adds the session access table).
- Official components are pinned to their v0.3.0 releases; Compose uses `ghcr.io/alonabot/alonabot:0.3.0` with WebUI 0.3.0 and Agent 0.3.0.

[0.3.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.0
