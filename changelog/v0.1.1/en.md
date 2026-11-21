# AlonaBot v0.1.1

## Added

- Add `Context.ReadNodes` to read standard message nodes from reply and forwarded-message handles, with media grants for the calling session. Typed node queries in OneBot 11 v0.1.1 require this Core version.
- Add portable Markdown message segments for Adapter-specific rendering, with explicit `PlainText()` conversion.

## Changed

- Pin official OneBot 11 and WebUI components to v0.1.1; help, Agent and Telegram remain at v0.1.0. WebUI v0.1.1 adds store search, sorting, role counts, nine-card pagination and submission checks. Submission opens a GitHub Issue template in the browser.
- Persist the management token in `admin_token` with mode `0600`; startup and token initialization/reset report its file path. Local WebUI supports token claim and confirmation, and management commands read credentials from the selected data directory.
- Public Releases continue to contain only five platform-specific Core archives, each containing Core, `LICENSE` and `source.json`. Compose uses `ghcr.io/alonabot/alonabot:0.1.1` with WebUI and Agent included. Existing release assets and tags are retained.

## Fixed

- Preserve media authorization and mount paths when serving media URLs, and register queried-node media for the calling session.
- Include released Core changelogs in the management API so the update view can display versioned release notes.

## Upgrade and compatibility

Stop Core normally and back up the complete data directory before upgrading. Replace the executable or update the pinned Compose image, then start with the same data directory. Installation records, source bindings, component generations and Adapter instance configuration are retained. Normal component uninstall retains data and source bindings for reinstallation from the same source.

Management commands use `--data-dir` in place of `--token-file` or `ALONABOT_ADMIN_TOKEN`; use SSH or `docker exec` for a remote Core. If an older login removed `admin_token`, stop Core and run `alonabot admin-token reset --data-dir <data_dir>` to restore local CLI access. Reset revokes existing management sessions.

OneBot 11 v0.1.1 can start on Core v0.1.0. Node queries use the new Core service and return `Core service not found` on the older version; upgrade Core to use them. Official components retain their independent signed-release installation flow. Help remains at v0.1.0.

This version was cross-built for all five targets and passed archive checks on darwin/arm64. CLI and HTTP management paths passed installation, activation, uninstall and same-source reinstallation of public signed components. Public OneBot 11 v0.1.1 and help v0.1.0 returned real `/help` replies in an authorized QQ test group before and after Core restart, retaining installation records, Active generations and instance configuration.

Other actual platform coverage and approved deferrals carry forward from the [v0.1.0 release record](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/en.md). Windows execution and CGO retain their approved deferrals.
