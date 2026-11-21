English · [简体中文](README.zh-CN.md)

<p align="center">
  <img src="docs/assets/mascot.png" alt="AlonaBot" width="200">
</p>

<h1 align="center">AlonaBot</h1>

<p align="center">
  <a href="https://github.com/AlonaBot/AlonaBot/actions/workflows/ci.yml"><img src="https://github.com/AlonaBot/AlonaBot/actions/workflows/ci.yml/badge.svg?branch=main" alt="Quality gate"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-AGPL--3.0-blue?style=flat-square" alt="AGPL-3.0"></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/Go-1.27.1-00ADD8?style=flat-square&amp;logo=go&amp;logoColor=white" alt="Go 1.27.1"></a>
  <a href="https://github.com/AlonaBot/AlonaBot/releases/latest"><img src="https://img.shields.io/github/v/release/AlonaBot/AlonaBot?style=flat-square&amp;color=22B8CF" alt="Release"></a>
  <a href="https://hits.sh/github.com/AlonaBot/AlonaBot/"><img src="https://hits.sh/github.com/AlonaBot/AlonaBot.svg?style=flat-square&amp;label=Views&amp;color=22B8CF" alt="Views"></a>
</p>

<p align="center">A lightweight, cross-platform chatbot framework in Go.<br>Install the components you need to build your bot.</p>

<p align="center">
  <a href="https://alonabot.github.io/docs/en/">Documentation</a> ·
  <a href="docs/plugin-guide.md">Plugin development</a> ·
  <a href="changelog/README.md">Changelog</a> ·
  <a href="CONTRIBUTING.md">Contributing</a>
</p>

## Why AlonaBot

- **Process isolation**: Native Go plugins, Adapters and Agents run in separate processes. Core manages startup, shutdown and recovery. Components run as trusted local code; administrators still need to review their sources.
- **Replaceable components**: Adapters connect chat platforms, Agents provide AI capabilities, and WebUI provides the management interface. Community components can replace the official implementations.
- **Run on demand**: Disabled components start no processes. Plugins that support sleep can exit when idle and wake for messages, service calls, HTTP requests or scheduled tasks.
- **Go development**: Chain command and event registrations, with hot reload, managed SQLite, configuration forms, i18n and a plugin test harness that runs against a real Core.
- **Unified management**: Use WebUI or CLI to manage components, accounts, models, permissions, sessions, configuration, logs and updates. WebUI supports desktop and narrow screens.

The official Agent supports streaming conversations, tool calls, MCP, skills, personas, memory and scheduled tasks. Model support includes OpenAI, OpenAI Compatible, Responses, Anthropic, Gemini and Ollama; provider SDKs are linked into the Agent as needed.

## Official components

| Component | Repository | Bundled deployment |
| --- | --- | --- |
| WebUI | [alonabot-webui](https://github.com/AlonaBot/alonabot-webui) | Preinstalled and enabled |
| Agent | [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) | Preinstalled; enable after configuration |
| Telegram | [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) | Install as needed |
| OneBot 11 | [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) | Install as needed; currently supports receiving |
| Help | [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) | Install as needed |

The Core repository contains the SDK, protocol, examples, bundled releases and cross-component tests. Components build and release independently; the [component source lock](components.lock.json) pins the bundled versions.

## Documentation and contributing

| What you want to do | Start here |
| --- | --- |
| Install, configure, upgrade or back up a bot | [Deployment](docs/deployment.md) |
| Write a plugin, Adapter or Agent | [Plugin guide](docs/plugin-guide.md) |
| Change Core, SDK or WebUI | [Development](docs/development.md) · [Architecture](ARCHITECTURE.md) |
| Report a problem, submit a fix or improve documentation | [Contributing](CONTRIBUTING.md) |
| Look up exact API and protocol behavior | [Specifications](spec/README.md) |
| Build and verify release artifacts | [Release maintenance](docs/releasing.md) |

Contributions to fixes, platform support, documentation and community components are welcome. Include the version, operating system and reproduction steps when reporting a problem, and redact logs and configuration before sharing them. Discuss larger changes in an Issue first.

## License

AlonaBot and the official components use the [GNU Affero General Public License v3.0](LICENSE) (`AGPL-3.0-only`). Plugin templates follow the license of their own repository.
