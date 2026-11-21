# Contributing to AlonaBot

[简体中文](CONTRIBUTING.zh-CN.md)

Contribute fixes, documentation, platform support and community components. This guide covers the process from reporting an issue to submitting a PR. See [development](docs/development.md) for tools and commands and [coding standards](docs/coding-standards.md) for code rules.

Submit component changes to the corresponding [official repository](README.md#official-components). This repository maintains Core, SDK, protocols, examples and bundled releases.

## Start with a concrete problem

Search existing Issues and open or merged PRs, and check current source. If a fix already exists, help reproduce or review it. For larger features, protocol changes or data changes, open an Issue describing the problem, expected behavior and approach, and wait for maintainer agreement.

- **Bugs**: provide version, operating system and architecture, reproduction steps, expected and actual results, and complete errors. Reduce the issue to a repeatable example where possible.
- **Features**: describe the actual use case, missing capability and expected result. Prefer separate components for platform integrations.
- **Documentation**: identify the step that caused difficulty and its command or behavior. Small corrections can go directly into a PR.

Before sharing logs, configuration or diagnostics, redact Bot Tokens, management tokens, database credentials and real user data. Report suspected credential leaks or security vulnerabilities privately using the contact details on the maintainer's GitHub profile. Avoid publishing exploitable details in public Issues.

## Prepare the environment

Follow [environment setup](docs/development.md#prepare-the-environment). Go changes usually need only Go. Protocol generation also needs protoc; WebUI changes need Node and npm.

Create a focused branch from the latest main. Fork contributors should first synchronize upstream main. Suggested names include `fix/problem`, `feat/feature` and `docs/topic`. Each PR should address one coherent problem. Keep unrelated refactoring, personal settings and temporary acceptance artifacts out of it. Worktrees support parallel independent tasks.

## Keep implementation and documentation aligned

- Read the [architecture](ARCHITECTURE.md) to identify module ownership, import directions and state ownership. See the [plugin guide](docs/plugin-guide.md) for supported component behavior.
- Follow [coding standards](docs/coding-standards.md): Simplified Chinese comments, English logs, errors and CLI output, and i18n for user-facing interfaces and bot replies. Public documentation defaults to English with a complete Simplified Chinese counterpart.
- Before adding a dependency or expanding its use, explain the need, alternatives and linking cost, and obtain explicit maintainer approval. See [dependency management](docs/development.md#dependency-management).
- Edit generator sources and regenerate their output. Preserve reproducibility.
- Update every caller when an interface changes. Add migrations for published state databases; see [versioning and compatibility](docs/coding-standards.md#versioning-and-compatibility).
- Update related contracts, examples, documentation and necessary verification when behavior changes. Explain failure handling and recovery for database, process lifecycle or public protocol changes.

## Verify the change

Choose the smallest sufficient checks for its actual effects; commands are in [development](docs/development.md#verify-changes). Check documentation content, links, anchors and `git diff --check`. Run affected Go tests and static checks. For frontend changes, check types, build and inspect the main interactions in the actual interface.

New nontrivial logic needs at least one meaningful test. The repository commits integration tests and protocol-number checks; unit tests, temporary reproductions and frontend local tests run in ignored local files. The default is one integration test covering the PR's main path; explain additional tests. See [test submission rules](docs/coding-standards.md#test-submission-rules).

Record actual commands, platform, results and local tests omitted from the commit in the PR. Mark checks lacking an environment as "not run" with the reason. Distinguish cross-compilation, emulated execution and real-platform results. Larger changes across modules require an independent review.

## Submit a PR

1. Write a Simplified Chinese commit message explaining the change and its reason. No fixed type prefix is required.
2. Use `type/scope: 中文摘要` for PR titles, for example `fix/scheduler: 修复休眠唤醒后的任务交付`. Types are `feat`, `fix`, `docs`, `refactor`, `test`, `ci`, `build`, `chore`, `perf` and `revert`. Use the affected module's lowercase English name for scope; hyphens are allowed. Describe the technical change without task numbers, model names or tool attribution in the title.
3. Target main and complete the [PR template](.github/PULL_REQUEST_TEMPLATE.md): reason, verification, compatibility, risks and checklist. Link an Issue with `Closes #number`.
4. Review the diff for scope and necessary adjacent changes. Exclude credentials, personal task records, runtime data and temporary tests.
5. Wait for checks on the current head and maintainer review. Make requested corrections and run affected verification.

After confirming scope, implementation, verification and deferred items, the maintainer squash-merges and checks main. Contributors deliver through PRs; maintainers control official versions, signatures and releases.

## CI and review

- [Quality gate](.github/workflows/ci.yml) checks whitespace and allowed main paths. Go, protocol or dependency changes additionally run formatting, generation consistency, vet, race tests and five-target pure-Go builds. Authors verify frontend changes following the development guide.
- [PR metadata](.github/workflows/pr-metadata.yml) checks the title, required description and checklist of open non-draft PRs. Maintainers review their actual contents.
- GitHub automation handles path labels, assignment and review requests. Metadata workflows keep code separate from secrets.
- main contains the product, current conventions and public engineering documentation. CI rejects personal Agent configuration, coordination records and historical implementation or acceptance artifacts.

Maintainers manage branch protection; workflows define actual CI commands and permissions.

## Contribution license

AlonaBot and official components use the [GNU Affero General Public License v3.0](LICENSE) (`AGPL-3.0-only`). Ensure you can contribute your code and documentation under this license. Preserve third-party copyright and license notices. Plugin templates follow the license declared by their own repository.
