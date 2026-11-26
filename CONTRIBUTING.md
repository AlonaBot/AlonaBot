# Contributing to AlonaBot

[简体中文](CONTRIBUTING.zh-CN.md)

Contribute fixes, documentation, platform support and community components. This guide covers the process from reporting an issue to submitting a PR. See [development](docs/development.md) for tools and commands and [coding standards](docs/coding-standards.md) for coding and test rules.

Submit component changes to the corresponding [official repository](README.md#official-components). This repository maintains Core, SDK, protocols, examples and bundled releases.

## Start with a concrete problem

Use [Discussions](https://github.com/AlonaBot/AlonaBot/discussions) for usage questions and design discussions. Report vulnerabilities privately following the [security policy](SECURITY.md). The [Wiki](https://github.com/AlonaBot/AlonaBot/wiki) provides task navigation.

Search existing Issues and open or merged PRs, and check current source. If a fix already exists, help reproduce or review it. For larger features, protocol changes or data changes, open an Issue describing the problem, expected behavior and approach, and wait for maintainer agreement.

- **Bugs**: provide version, operating system and architecture, reproduction steps, expected and actual results, and complete errors. Reduce the issue to a repeatable example where possible.
- **Features**: describe the actual use case, missing capability and expected result. Prefer separate components for platform integrations.
- **Documentation**: identify the step that caused difficulty and its command or behavior. Small corrections can go directly into a PR.

Before sharing logs, configuration or diagnostics, redact Bot Tokens, management tokens, database credentials and real user data. Report suspected credential leaks or security vulnerabilities privately following the [security policy](SECURITY.md). Avoid publishing exploitable details in public Issues.

## Prepare the environment

Follow [environment setup](docs/development.md#prepare-the-environment). Go changes usually need only Go. Protocol generation also needs protoc; WebUI changes need Node and npm.

Create a focused branch from the latest main. Fork contributors should first synchronize upstream main. Suggested names include `fix/problem`, `feat/feature` and `docs/topic`. Each PR should address one coherent problem. Keep unrelated refactoring, personal settings and temporary acceptance artifacts out of it. Worktrees support parallel independent tasks.

## Keep changes in sync

- Update generator sources and regenerate their outputs.
- Update affected callers, examples and guides with behavior changes. For database, process lifecycle and public protocol changes, describe failure handling and recovery.
- Public dependency scope and linking costs are documented in [dependency management](docs/development.md#dependency-management).

## Submit a PR

1. Write a Simplified Chinese commit message explaining the change and its reason. No fixed type prefix is required.
2. Use `type/scope: 中文摘要` for PR titles, for example `fix/scheduler: 修复休眠唤醒后的任务交付`. Types are `feat`, `fix`, `docs`, `refactor`, `test`, `ci`, `build`, `chore`, `perf` and `revert`. Use the affected module's lowercase English name for scope; hyphens are allowed. Describe the technical change without task numbers, model names or tool attribution in the title.
3. Target main and complete the [PR template](.github/PULL_REQUEST_TEMPLATE.md): reason, verification, compatibility, risks and checklist. Link an Issue with `Closes #number`.
4. Review the diff for scope and necessary adjacent changes. Exclude credentials, personal task records, runtime data and temporary tests.
5. Wait for checks on the current head and maintainer review. Make requested corrections and run affected verification.

After confirming scope, implementation, verification and deferred items, the maintainer squash-merges and verifies the merged result on main. Contributors deliver through PRs; maintainers control official versions, signatures and releases.

## CI and review

- [Quality gate](.github/workflows/ci.yml) checks whitespace and allowed main paths. PRs with Go, protocol or CI changes run formatting, generation consistency, vet, affected-package ordinary tests and a Go build; changes under `tests/` run that module's ordinary tests. Only manual runs and releases run race tests, cross-component integration tests, five-target pure-Go builds and release-tool checks in parallel. Authors verify frontend changes following the development guide.
- [PR metadata](.github/workflows/pr-metadata.yml) checks the title, required description and checklist of open non-draft PRs. Maintainers review their actual contents.
- GitHub automation handles path labels, assignment and review requests. Metadata workflows keep code separate from secrets.
- main contains the product, current conventions and public engineering documentation. CI rejects personal Agent configuration, coordination records and historical implementation or acceptance artifacts.

Maintainers manage branch protection; workflows define actual CI commands and permissions.

## Contribution license

AlonaBot and official components use the [MIT License](LICENSE). Ensure you can contribute your code and documentation under this license. Preserve third-party copyright and license notices.

## Coding standards

See [coding standards](docs/coding-standards.md) for Go style, naming, comments, logging, compatibility and test-submission rules.
