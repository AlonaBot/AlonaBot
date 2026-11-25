## [0.3.0] - 2026-11-23

### Added

- On the first visit to WebUI from the Docker host, you can claim the admin token in the browser. You can also choose to stay signed in on the current device ([#312](https://github.com/AlonaBot/AlonaBot/pull/312)).
- Administrators can install components by building them from GitHub source and follow progress in the build log. Download settings accept a GitHub mirror, API address, Go download address, `GOPROXY` and a network proxy. Core still verifies mirrored release assets against the original digest from GitHub ([#312](https://github.com/AlonaBot/AlonaBot/pull/312)).
- Administrators can allow all chats, allow only selected chats, or block selected chats for each profile. The setting applies to managed components ([#315](https://github.com/AlonaBot/AlonaBot/pull/315)).
- Adapter authors can declare an account setup page with `page_purpose = instance_config`. WebUI opens it when an account is added or edited. The page can read the setup context, show a QR code and return the user's changes to WebUI ([#317](https://github.com/AlonaBot/AlonaBot/pull/317)).
- Developers can use `alonabot adapter instance list/actions/call` or `sdk/debug` to inspect and call actions on running adapter instances. For integration testing, `plugintest` supports multiple adapters and instances, configurable platform responses, and asynchronous event submission and waiting ([#310](https://github.com/AlonaBot/AlonaBot/pull/310)).
- `alonabot --version` prints the version, build date and release page URL ([#309](https://github.com/AlonaBot/AlonaBot/pull/309)).
- Component authors can handle sleep with `OnSleep`, start multi-turn tasks with `agent.Run`, and participate in task execution through AI hooks and Goal interfaces. The service protocol supports streaming responses ([#308](https://github.com/AlonaBot/AlonaBot/pull/308)).

### Changed

- Plugin and agent authors should rename their configuration schema to `plugin_config_schema.json` before rebuilding. Adapters still use `adapter_config_schema.json`. If you installed Agent separately, replace the old package with a build targeting this Core release before re-enabling it. Configuration and managed data are retained.
- For a single model call, the SDK now takes a structured request through `llm.Generate`. Replace calls to `llm.Ask` or the old `agent.Generate` with that entry point. Only path command builders expose `Menu`, arguments and aliases ([#308](https://github.com/AlonaBot/AlonaBot/pull/308)).
- Source builds compile one at a time. The build log warns when memory is low or output stalls. If the compiler is killed, the task reports possible memory exhaustion and suggests alternatives. On Linux, the compiler is targeted first when memory runs out, reducing the risk that Core also exits ([#316](https://github.com/AlonaBot/AlonaBot/pull/316)).
- The management API includes names and descriptions in the component list and reports the owning component's state for each account instance ([#312](https://github.com/AlonaBot/AlonaBot/pull/312)).

### Fixed

- Updating a degraded component now keeps the successfully activated new version running. Previously, Core could stop it immediately after the update ([#316](https://github.com/AlonaBot/AlonaBot/pull/316)).

[0.3.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.0
