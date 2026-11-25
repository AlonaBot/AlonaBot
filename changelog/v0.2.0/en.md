## [0.2.0] - 2026-11-22

### Added

- Telegram can show different command menus for each chat and account. Adapter authors can use `CommandMenuController` to set, read, delete or sync menus and check sync status. Scoped menus require this Core release and an adapter that implements the interface ([#302](https://github.com/AlonaBot/AlonaBot/pull/302)).
- Adapters can use the trusted media type from Core to recognize GIFs and handle them according to platform capabilities.
- Users and component authors have deployment, development and contribution guides in English and Simplified Chinese. The community docs also explain how to report security issues privately.

### Changed

- A single platform archive is enough to upload a component in WebUI or install it locally with the CLI. For online installs and updates, Core selects the archive from GitHub Releases and checks the component and source against its `plugin.yaml`. Once official releases remove the old manifest and signature assets, Core versions that use those assets need this release for online installs and updates. Installed components can keep running ([#305](https://github.com/AlonaBot/AlonaBot/pull/305)).
- Root commands appear in menus only when marked with `.Menu()`. To keep an existing plugin's commands visible in menus, mark them with `.Menu()` and rebuild with this SDK. Users can still invoke the commands directly ([#302](https://github.com/AlonaBot/AlonaBot/pull/302)).
- Agent v0.2.0 can read quoted and forwarded messages and their images. WebUI v0.2.0 accepts single-archive uploads, help v0.2.0 declares its menu entries, and Telegram v0.2.0 adds media conversion and typed actions. The official OneBot 11 adapter also has a separate v0.2.0 release.
- Core, the SDK and the official components now use the MIT License.
- The release page provides five Core-only platform archives. The `ghcr.io/alonabot/alonabot:0.2.0` Compose image includes Agent and WebUI v0.2.0.

### Removed

- The single-archive workflow replaces `pack --key/--asset-base` and `install --release/--signature/--public-key`. In scripts, use `alonabot pack <directory> --out <output-directory>` to package a component and `alonabot install --source AlonaBot/<repository> --yes` or `alonabot install <archive> --yes` to install it. A successful update or reinstall clears the old publisher fingerprint and keeps the data and source binding.

### Recorded verification

Windows runtime, CGO and further resource measurements retain the deferrals recorded in [v0.1.0](https://github.com/AlonaBot/AlonaBot/blob/main/changelog/v0.1.0/en.md).

[0.2.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0
