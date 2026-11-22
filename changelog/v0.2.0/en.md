## [0.2.0] - 2026-11-22

### Added

- Add chat and account scoped command menus with typed Set/Get/Delete/Status/Sync, ready events, persisted ownership and lifecycle synchronization. See [PR #302](https://github.com/AlonaBot/AlonaBot/pull/302).
- Preserve trusted MIME metadata when Core prepares media sources for adapters, including GIF identification.
- Add bilingual documentation navigation, community contribution guidance and private security reporting links.

### Changed

- Core installs and updates components directly from platform archives. Online installation selects GitHub Release assets and reads the inner `plugin.yaml`, retaining repository, identity, version, platform, build-info, resource, pin and dependency checks. Local installation and management uploads accept one archive. See [PR #305](https://github.com/AlonaBot/AlonaBot/pull/305).
- Command registration registers handlers by default; add `.Menu()` to root commands that should appear automatically. Scoped Telegram menus require an adapter built with `CommandMenuController`; existing adapters retain `SyncCommands`.
- Provide English and Simplified Chinese deployment, component development and contribution guides. Protocol and management API guidance now lives alongside usage instructions, with protocol fields defined in `.proto` sources.
- Core, SDK and the official components use the MIT License.
- Pin Agent, WebUI, help, Telegram and OneBot 11 to their independent v0.2.0 releases. Agent reads quoted and forwarded-message nodes with image sources; WebUI supports single-archive uploads; help explicitly registers its menu entries; Telegram adds scoped menus, media mappings and typed actions.
- Publish five Core archives: four `.tar.gz` files and one Windows `.zip`; GitHub supplies source archives. Compose uses `ghcr.io/alonabot/alonabot:0.2.0`, with Agent v0.2.0 and WebUI v0.2.0.

### Removed

- Remove publisher signature requirements and standalone component manifests from the Core installation flow. Successful component updates or reinstallation clear the retired publisher fingerprint while preserving installed data and source bindings.
- Remove `pack --key/--asset-base` and `install --release/--signature/--public-key`. Use `alonabot pack <directory> --out <output-directory>` and `alonabot install --source AlonaBot/<repository> --yes` or `alonabot install <archive> --yes`.

### Upgrade and compatibility

Stop Core normally and back up the complete data directory before upgrading. Replace the executable or update the pinned Compose image, then start with the same data directory. Installation records, component generations, source bindings and Adapter instance configuration are retained. Command menu persistence is added through a new state migration.

Upgrade Core first, then update WebUI and the other components. This version reads existing component platform archives. Official component v0.2.0 Releases use platform archives alone; historical `release.json`, `release.json.sig` and `release.pub` assets are removed after Core and WebUI verification. Older Core versions require an upgrade for online installation and updates after that cleanup. Installed components continue to use their existing data.

Automatic command menus now include only entries explicitly marked with `.Menu()`. Previously published plugins need a rebuild with the new SDK and `.Menu()` to restore those entries; their command handlers continue to work. Upgrade Core and the relevant Adapter to use scoped menu control.

Windows execution, CGO and the approved resource-measurement deferrals carry forward from the [v0.1.0 release record](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/en.md). Cross-builds and simulated external services retain their original verification scope.

[0.2.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0
