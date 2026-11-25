## [0.1.1] - 2026-11-21

### Added

- Plugins can use `Context.ReadNodes` to read quoted or forwarded messages and access their media in the calling session. Node queries rely on the new Core service; calls to an older Core return `Core service not found` ([#293](https://github.com/AlonaBot/AlonaBot/pull/293)).
- Plugins can send Markdown segments for adapters to render in the platform's format. Call `PlainText()` when you need plain text.

### Changed

- Core stores the management token in `admin_token` with mode `0600`. Startup, initialization and reset show only the file path. Local WebUI can claim and confirm the token; management commands read it from the directory selected with `--data-dir`. Update scripts that use `--token-file` or `ALONABOT_ADMIN_TOKEN` to select the data directory instead. If an older sign-in removed the token file, stop Core and run `alonabot admin-token reset --data-dir <data_dir>` to regain access. A reset revokes existing management sessions.
- Plugin and agent examples created with `alonabot init` reply with `Hello AlonaBot` out of the box. Add i18n files as needed ([#295](https://github.com/AlonaBot/AlonaBot/pull/295)).
- Readers can switch between the English and Simplified Chinese READMEs and their documentation homepages.

### Fixed

- The management UI can read published Core notes through the changelog API. It serves English or Simplified Chinese based on `Accept-Language`, defaults to English when there is no match, and leaves unreleased entries in the repository ([#295](https://github.com/AlonaBot/AlonaBot/pull/295)).

### Recorded verification

All five targets were cross-built and their archives checked on darwin/arm64. The CLI and HTTP API passed installation, activation, uninstall and same-source reinstallation of signed components. An integration test in a real QQ group verified that the bot answered `/help` before and after Core restarted. Windows runtime and CGO checks remain deferred as recorded in the [v0.1.0 release notes](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/en.md#recorded-verification).

[0.1.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.1
