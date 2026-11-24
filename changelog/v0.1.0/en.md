## [0.1.0] - 2026-11-21

This is the first official AlonaBot release. Run Core to host a bot, then install the chat adapters, plugins and agents you need.

### Added

- Go developers can use the `v0.1.0` SDK to write independently installable plugins, adapters and agents. The Telegram and OneBot 11 adapters and the help plugin are available separately. All five official components publish signed archives under AGPL-3.0-only in this release.
- The [deployment guide](https://github.com/AlonaBot/AlonaBot/blob/af4f68550dcef191d5bc9a5e345b48510196ae23/docs/deployment.md) covers native installation and Docker Compose. The `ghcr.io/alonabot/alonabot:0.1.0` image includes Core, WebUI and Agent. Compose enables WebUI on first start. Follow the startup prompt to claim a one-time admin token and sign in at `http://127.0.0.1:7650`. Configure a provider and model before enabling Agent.
- Core keeps the component's source, publisher binding and version pin when you update it. Rollback creates a new generation and a database copy. A standard uninstall keeps component data and the source binding for reinstallation from that source. Uninstalling the Core service and program also preserves its data directory.

### Changed

- On November 21, the five platform archives were repackaged in place to contain only the Core executable, `LICENSE` and `source.json`. The Core binaries, tag and bundled Compose image stayed the same, and existing installations kept their components. **If you still have an old archive, download it again; the previous digest is obsolete.** The installer verifies downloads against the digest reported by the GitHub Release API. See [#291](https://github.com/AlonaBot/AlonaBot/pull/291) and [#292](https://github.com/AlonaBot/AlonaBot/pull/292).
- The release page provides tar.gz archives for Linux and macOS on amd64 and arm64, and a ZIP archive for Windows amd64. All five use `CGO_ENABLED=0`; GitHub provides the source archives. Get the install script from `main` and the Compose file from the version tag.

### Known limitations

- OneBot 11 can receive events; sending returns `UNSUPPORTED`. The welcome and rss plugins and the Telegram features listed in the [original release agreement](https://github.com/AlonaBot/AlonaBot/blob/f0382910/spec/CONTRACTS-v0.1.md#152-发行包官方整合包与预装组件) are deferred.
- Native components run local code trusted by the administrator. In library mode they share Core's process: an unhandled panic ends that process, and a stop timeout cannot forcibly end a goroutine. Components that declare `prepare_effects=external` require maintenance and backup approval and their own recovery procedure for external data.
- Use a custom image with the relevant runtime if an MCP stdio tool needs a shell, Node or Python.

### Recorded verification

Before the repositories were split, the bundled candidate ran on darwin/arm64 natively, darwin/amd64 under Rosetta, linux/amd64 in a VM, and Linux in Docker on an ARM host. Native Intel Mac execution and ARM Linux in a VM or on physical hardware were unverified. Windows cross-builds and static checks passed; the maintainer deferred Windows runtime and CGO checks on 2026-11-20. Workflow checks used a local Telegram Bot API and a synthetic provider. See the [historical acceptance record](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md) for conditions, failures and deferred checks.

### Resource results and dependencies

The historical 24/10/16 MiB comparison missed its targets. The maintainer accepted the overhead and stopped further resource measurements and load tests. The [original measurements](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/docs/benchmarks.md) record the data and conditions. They used `github.com/AlonaBot/gorm v1.31.3-0.20261120022453-34b98340007d`; the first release uses upstream `gorm.io/gorm v1.31.2`, which may increase binary size and memory use.

[0.1.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.0
