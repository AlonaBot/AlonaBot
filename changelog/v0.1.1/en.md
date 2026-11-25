## [0.1.1] - 2026-11-21

### Added

- Plugins can use `Context.ReadNodes` to read quoted or forwarded messages and access their media in the calling session. Node queries in OneBot 11 v0.1.1 require Core v0.1.1; older Core versions return `Core service not found`.
- Plugins can send Markdown segments for adapters to render in the platform's format. Call `PlainText()` when you need plain text.
- WebUI v0.1.1 lets you search, sort and browse store components by role. You can also check whether a component is listed and open a GitHub issue template to submit it.

### Changed

- Core stores the management token in `admin_token` with mode `0600`. Startup, initialization and reset show only the file path. Local WebUI can claim and confirm the token; management commands read it from the directory selected with `--data-dir`. Update scripts that use `--token-file` or `ALONABOT_ADMIN_TOKEN` to select the data directory instead. If an older sign-in removed the token file, stop Core and run `alonabot admin-token reset --data-dir <data_dir>` to regain access. A reset revokes existing management sessions.
- Plugin and agent examples created with `alonabot init` reply with `Hello AlonaBot` out of the box. Add i18n files as needed ([#295](https://github.com/AlonaBot/AlonaBot/pull/295)).
- Readers can switch between the English and Simplified Chinese READMEs and their documentation homepages.
- The `ghcr.io/alonabot/alonabot:0.1.1` Compose image includes WebUI v0.1.1 and Agent v0.1.0. The official OneBot 11 adapter was released separately at v0.1.1; help and Telegram stayed at v0.1.0.

### Fixed

- The management UI can read published Core notes through the changelog API. It serves English or Simplified Chinese based on `Accept-Language`, defaults to English when there is no match, and leaves unreleased entries in the repository ([#295](https://github.com/AlonaBot/AlonaBot/pull/295)).

### Recorded verification

All five targets were cross-built and their archives checked on darwin/arm64. The CLI and HTTP API passed installation, activation, uninstall and same-source reinstallation of signed components. OneBot 11 v0.1.1 and help v0.1.0 answered `/help` in a real QQ test group before and after Core restarted. Windows runtime and CGO checks remain deferred as recorded in the [v0.1.0 release notes](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/en.md#recorded-verification).

[0.1.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.1
