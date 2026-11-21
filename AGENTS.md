# AlonaBot development conventions

[简体中文](AGENTS.zh-CN.md)

This repository maintains Core, SDK, protocols, examples and bundled releases. Official components live in their own repositories; see [official components](README.md#official-components).

## Development entry points

- Environment and checks: [development guide](docs/development.md).
- Modules and dependency directions: [architecture](ARCHITECTURE.md).
- Naming, comments and tests: [coding standards](docs/coding-standards.md).
- Public behavior: [plugin guide](docs/plugin-guide.md).
- Pull requests and licensing: [contributing](CONTRIBUTING.md).
- Packaging and releases: [release maintenance](docs/releasing.md).

## Working conventions

- Use Simplified Chinese for communication, comments and commit messages. Maintain public Markdown documentation in English by default with a complete Simplified Chinese counterpart and language links. Use English for identifiers, logs, errors and CLI output; user interfaces use i18n.
- Read affected implementations, callers and current changes first. Preserve other people's changes and use an isolated worktree for Git writes. Commit, push, merge and release only within the current explicit authorization.
- Explain the need and alternatives for a new dependency and obtain explicit maintainer approval. Edit generator sources and regenerate their output.
- Validate only at external input and trust boundaries. Preserve original errors; retries and timeouts for business side effects follow the existing contract.
- Match verification to the change. Check documentation and templates for content, links, anchors and whitespace. Run affected checks for Go changes and obtain an independent review for changes across modules. Follow the coding standards for committed tests.
- Personal skills, task cards, coordination and historical acceptance records belong on dev. Keep credentials, runtime data and temporary files local.
