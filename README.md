English · [简体中文](README.zh-CN.md)

<p align="center">
  <img src="docs/assets/mascot.png" alt="AlonaBot" width="200">
</p>

<h1 align="center">AlonaBot</h1>

<p align="center">
  <a href="https://github.com/AlonaBot/AlonaBot/actions/workflows/ci.yml"><img src="https://github.com/AlonaBot/AlonaBot/actions/workflows/ci.yml/badge.svg?branch=main" alt="Quality gate"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="MIT"></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/Go-1.27.1-00ADD8?style=flat-square&amp;logo=go&amp;logoColor=white" alt="Go 1.27.1"></a>
  <a href="https://github.com/AlonaBot/AlonaBot/releases/latest"><img src="https://img.shields.io/github/v/release/AlonaBot/AlonaBot?style=flat-square&amp;color=22B8CF" alt="Release"></a>
  <a href="https://hits.sh/github.com/AlonaBot/AlonaBot/"><img src="https://hits.sh/github.com/AlonaBot/AlonaBot.svg?style=flat-square&amp;label=Views&amp;color=22B8CF" alt="Views"></a>
</p>

<p align="center">A lightweight, cross-platform chatbot framework in Go.<br>Install the components you need to build your bot.</p>

<p align="center">
  <a href="https://alonabot.github.io/docs/en/">Documentation</a> ·
  <a href="docs/plugin-guide.md">Plugin development</a> ·
  <a href="changelog/README.md">Changelog</a> ·
  <a href="CONTRIBUTING.md">Contributing</a> ·
  <a href="https://github.com/AlonaBot/AlonaBot/wiki">Wiki</a> ·
  <a href="https://github.com/AlonaBot/AlonaBot/discussions">Discussions</a> ·
  <a href="https://github.com/AlonaBot/AlonaBot/issues/new/choose">Report an issue</a> ·
  <a href="SECURITY.md">Security reports</a>
</p>

## Core features

- **Fault isolation**: With separate-process deployment, plugins, Adapters and Agents each run in their own process. A component crash leaves Core running; Core manages startup, shutdown and automatic restarts.
- **Decoupled modules**: Platform connectivity (Adapters), AI capabilities (Agents) and the management interface (WebUI) are separate from Core. You can replace official implementations with community or custom components at any time.
- **Low resource usage**: Disabled components do not start processes. Plugins that support hibernation exit when idle to release memory and wake automatically for messages, HTTP requests, scheduled tasks or calls from other components.
- **Plugin development**: Chainable command and event registration, hot reload, ready-to-use SQLite storage, configuration declarations rendered automatically as WebUI forms, i18n, and a test suite backed by a real Core.
- **Centralized operations**: Manage component updates, accounts and permissions, models, session settings and runtime logs through WebUI and CLI, with desktop and mobile support.

## Official components

| Component | Repository | Bundled deployment |
| --- | --- | --- |
| WebUI | [alonabot-webui](https://github.com/AlonaBot/alonabot-webui) | Preinstalled and enabled |
| Agent | [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) | Preinstalled; enable after configuration |
| Telegram | [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) | Install as needed |
| OneBot 11 | [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) | Install as needed; currently supports receiving |
| Help | [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) | Install as needed |

## Documentation and contributing

| What you want to do | Start here |
| --- | --- |
| Install, configure, upgrade or back up a bot | [Deployment](docs/deployment.md) |
| Write a plugin, Adapter or Agent | [Plugin guide](docs/plugin-guide.md) |
| Change Core, SDK or WebUI | [Development](docs/development.md) · [Architecture](ARCHITECTURE.md) |
| Report a problem, submit a fix or improve documentation | [Contributing](CONTRIBUTING.md) |
| Look up management APIs and protocol behavior | [Management API](docs/deployment.md#management-api) · [Protocol and sessions](docs/development.md#protocol-and-session-behavior) |
| Build and verify release artifacts | [Release maintenance](docs/releasing.md) |

Contributions to fixes, platform support, documentation and community components are welcome. Include the version, operating system and reproduction steps when reporting a problem, and redact logs and configuration before sharing them. Discuss larger changes in an Issue first.

## License

AlonaBot and the official components use the [MIT License](LICENSE).
