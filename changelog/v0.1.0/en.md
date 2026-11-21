# AlonaBot v0.1.0

The first official AlonaBot release provides a signed component ecosystem, five platform archives, and an integrated Core + WebUI + Agent distribution through Docker Compose.

## Included

- Five archive targets: linux/amd64, linux/arm64, windows/amd64, darwin/amd64 and darwin/arm64. Windows uses ZIP; other targets use tar.gz.
- Core, WebUI and Agent form the integrated distribution. The repository retains its Dockerfile, installer and root `docker-compose.yml`. WebUI is enabled on first start; configure a Provider and model before enabling Agent.
- Telegram and OneBot 11 adapters and the help plugin are installed separately from their component repositories. OneBot 11 currently receives events; sending returns `UNSUPPORTED`.
- Each independent component publishes its own signed `release.json`, signature, public key and platform archives. `components.lock.json` pins the released versions. Integrated archives verify and preserve the published component assets without rebuilding them. All five official components use AGPL-3.0-only.

## Install and upgrade

Follow the [deployment guide](https://github.com/AlonaBot/AlonaBot/blob/main/docs/deployment.md). Archives retain Core alongside `packages/`, the license and `source.json`. The public image is `ghcr.io/alonabot/alonabot:0.1.0` (`latest` also points to this release). Open `http://127.0.0.1:7650` and use the one-time administration token. Compose binds the published port to loopback, runs as UID 10001 with init, and preserves `/data` in a named volume.

Stop Core and components before upgrading. Back up the complete data directory, deployed programs, packages and external component data. Replace the complete distribution or recreate the container with the same volume. Component updates retain source/publisher bindings and pin settings; rollback creates a new generation and database copy. The installer removes the service and programs on uninstall while retaining data.

The scratch image includes CA certificates. MCP stdio tools requiring shell, Node or Python need a custom image containing those runtimes. Use an SSH tunnel or HTTPS reverse proxy for remote management.

## Recorded verification

These platform results describe the candidate before repository extraction. Checks for extracted components are recorded in their integration PR.

| Target | Actual runtime evidence |
| --- | --- |
| darwin/arm64 | Native Apple Silicon; distribution, install/update chain and real launchd installation passed. |
| darwin/amd64 | Passed under Rosetta; native Intel Mac execution was not run. |
| linux/amd64 | Passed in an x86_64 Linux VM with real systemd installation; the candidate image also passed under Rosetta. |
| linux/arm64 | Passed in Docker on an ARM host; an ARM Linux VM or physical Linux host was not run. |
| windows/amd64 | Cross-compilation and static release checks passed; runtime and CGO checks were deferred with maintainer approval on 2026-11-20. |

Each of the four available environments completed 1,082 steps and 104 Core runs, including 100 start/stop and recovery cycles. The CGO fixture returned 9 in all four. Five-target static checks, the darwin/arm64 install/update chain, archive and dual-image business flows passed. The corrected Windows archive passed 16 static checks across its four historical component packages; disassembly confirmed `CREATE_NEW_PROCESS_GROUP`. Original conditions and failures remain in the [historical acceptance record](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/spec/TEST-MATRIX.md).

## Resource results and dependencies

The maintainer accepted the current overhead and stopped further resource measurements, load tests and optimization. Median PSS from three fixed Linux runs was Core-only **26.61328125 MiB**, echo **10.09375 MiB**, configdb GORM **16.21875 MiB**. The historical 24/10/16 MiB comparisons remain FAIL. New four-workload and NoneBot2 comparison runs were cancelled. See the [original measurements](https://github.com/AlonaBot/AlonaBot/blob/d8b9678931013b6d46a044dbdbb0db9182262065/docs/benchmarks.md).

Builds use Go 1.27.1, Node 26.5.1 and locked npm dependencies; Linux ARM platform acceptance used Go 1.27.3. Normal distributions use `CGO_ENABLED=0`. GORM now uses upstream `gorm.io/gorm v1.31.2` directly. The historical resource measurements above used `github.com/AlonaBot/gorm v1.31.3-0.20261120022453-34b98340007d`, whose constant-name hook lookups helped the linker discard unrelated methods. Binary size and runtime memory may increase after restoring upstream; further resource measurements remain stopped.

## Known limitations

- Native components execute trusted local code. Library components share the Core process: an unhandled panic ends it, and a timed-out goroutine cannot be forcibly stopped.
- Components declaring `prepare_effects=external` require explicit maintenance, backup permission and component-specific recovery.
- Business acceptance used a local Telegram Bot API and synthetic Provider. Final-candidate live Telegram and paid LLM credentials were not exercised. Windows runtime and CGO checks remain deferred.
- Back up development data before moving to this first official release; component IDs, state layouts and initial migrations changed during development.
- The SDK is tagged `v0.1.0`. Components first pin a reviewed Core commit so component publication and final integration do not depend on each other's future tags. The plugin template is updated to the official SDK tag after Core publication.

See [release maintenance](https://github.com/AlonaBot/AlonaBot/blob/main/docs/releasing.md) and [release-check](https://github.com/AlonaBot/AlonaBot/blob/main/scripts/release-check/README.md) for the operational steps.
