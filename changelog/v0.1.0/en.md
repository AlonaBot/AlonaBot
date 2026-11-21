## [0.1.0] - 2026-11-21

The first official AlonaBot release.

### Added

- Core and the `v0.1.0` Go SDK for independently installed plugins, adapters and agents.
- Five platform archives: linux/amd64, linux/arm64, windows/amd64, darwin/amd64 and darwin/arm64. Windows uses ZIP; other targets use tar.gz. Standard builds use `CGO_ENABLED=0`.
- An integrated Core + WebUI + Agent image, `ghcr.io/alonabot/alonabot:0.1.0`, with Docker Compose. WebUI is enabled on first start; configure a Provider and model before enabling Agent. Compose uses UID 10001, init, a persistent `/data` volume and a loopback port binding.
- Separately installable Telegram and OneBot 11 adapters and the help plugin. All five official components use AGPL-3.0-only and publish signed releases; `components.lock.json` pins their versions. Image bundles verify and preserve the published component assets.
- Component updates that retain source/publisher bindings and pin settings, and rollback into a new generation with a database copy. Installer uninstall removes services and programs while retaining data; ordinary component uninstall retains generations, data and source bindings for reinstallation from the same source.

### Changed

- On 2026-11-21, with maintainer approval, the five v0.1.0 archives were **repackaged in place as Core-only distributions**, containing the Core executable, `LICENSE` and `source.json`. The original Core binaries and tag are preserved; Compose images remain integrated. Existing installed components are retained. **Previous archive digests are obsolete: download the new archives and verify their new GitHub Release API asset digests.** See [PR #292](https://github.com/AlonaBot/AlonaBot/pull/292).
- Release attachments contain only the five platform archives; GitHub provides source archives automatically. Download `deploy/install.sh` from `main` and `docker-compose.yml` from the version tag. The installer verifies official downloads against GitHub Release API asset digests. See [PR #291](https://github.com/AlonaBot/AlonaBot/pull/291).

### Install and upgrade

Follow the [deployment guide](https://github.com/AlonaBot/AlonaBot/blob/af4f68550dcef191d5bc9a5e345b48510196ae23/docs/deployment.md). Native archives and the installer provide Core; install WebUI, Agent and adapters as needed. Compose preinstalls WebUI and Agent. For v0.1.0, claim the one-time administration token as prompted at startup; WebUI is available at `http://127.0.0.1:7650` after installation and activation. Use an SSH tunnel or HTTPS reverse proxy for remote management.

Before upgrading from a development build, stop Core and all components and back up the complete data directory, programs, packages, deployment configuration and external component data: component IDs, state layouts and initial migrations changed during development. Preserve the same volume when recreating containers. Follow [backup and recovery](https://github.com/AlonaBot/AlonaBot/blob/af4f68550dcef191d5bc9a5e345b48510196ae23/docs/deployment.md#备份与恢复) for upgrade, rollback and full restoration.

### Known limitations

- OneBot 11 receives events; sending returns `UNSUPPORTED`. The welcome and rss plugins and the Telegram follow-up features listed in CONTRACTS §15.2 remain planned.
- Native components execute trusted local code. Library components share the Core process: an unhandled panic ends it, and a timed-out goroutine cannot be forcibly stopped.
- Components declaring `prepare_effects=external` require explicit maintenance, backup permission and component-specific recovery for external effects.
- The scratch image includes CA certificates. MCP stdio tools requiring shell, Node or Python need a custom image containing those runtimes.

### Recorded verification

Historical acceptance covers the integrated candidate before repository extraction: darwin/arm64 ran natively, darwin/amd64 under Rosetta, linux/amd64 in an x86_64 VM and linux/arm64 in Docker on an ARM host. Native Intel Mac and ARM Linux VM/physical-host execution were not run. Windows cross-compilation and static checks passed; runtime and CGO checks were deferred with maintainer approval on 2026-11-20. Business checks used a local Telegram Bot API and synthetic Provider; final-candidate live Telegram and paid LLM credentials were not exercised. Detailed results, original failures and deferred checks remain in the [historical acceptance record](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md). Checks for extracted components are recorded in their integration PR.

### Resource results and dependencies

The maintainer accepted current overhead and stopped further resource measurements, load tests and optimization. Historical 24/10/16 MiB comparisons remain FAIL; new four-workload and NoneBot2 comparison runs were cancelled. Complete figures and conditions remain in the [original measurements](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/docs/benchmarks.md).

Those measurements used `github.com/AlonaBot/gorm v1.31.3-0.20261120022453-34b98340007d`; v0.1.0 uses upstream `gorm.io/gorm v1.31.2`. Binary size and runtime memory may increase after restoring upstream. Builds use Go 1.27.1 and Node 26.5.1 with locked npm dependencies; Linux ARM acceptance used Go 1.27.3.

[0.1.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.0
