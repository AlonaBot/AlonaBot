[English](README.md) · 简体中文

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

<p align="center">轻量级、跨平台的 Go 聊天机器人框架。<br>按需安装组件，把机器人搭成你需要的样子。</p>

<p align="center">
  <a href="https://alonabot.github.io/docs/">文档</a> ·
  <a href="docs/plugin-guide.md">插件开发</a> ·
  <a href="changelog/README.zh-CN.md">更新日志</a> ·
  <a href="CONTRIBUTING.md">参与贡献</a>
</p>

## 为什么选择 AlonaBot

- **进程隔离**：Native Go 插件、Adapter 与 Agent 各自运行，Core 管理启动、停止与故障恢复。适合信任的本机组件；独立进程仍需管理员审查来源。
- **组件可替换**：聊天平台由 Adapter 接入，AI 能力由 Agent 提供，管理界面由 WebUI 提供。社区组件可以替换官方实现。
- **按需运行**：未启用的组件不启动进程；支持休眠的插件可在空闲时退出，由消息、服务、HTTP 或定时任务唤醒。
- **Go 开发体验**：链式注册命令与事件，支持热重载、受管 SQLite、配置表单、i18n，以及通过真实 Core 运行的插件测试套件。
- **统一管理**：WebUI 与 CLI 管理组件、账号、模型、权限、会话、配置、日志和更新，支持桌面与窄屏。

## 官方组件

| 组件 | 仓库 | 整合包 |
| --- | --- | --- |
| WebUI | [alonabot-webui](https://github.com/AlonaBot/alonabot-webui) | 预装并启用 |
| Agent | [alonabot-agent](https://github.com/AlonaBot/alonabot-agent) | 预装，配置后启用 |
| Telegram | [alonabot-adapter-telegram](https://github.com/AlonaBot/alonabot-adapter-telegram) | 按需安装 |
| OneBot 11 | [alonabot-adapter-onebot11](https://github.com/AlonaBot/alonabot-adapter-onebot11) | 按需安装，当前支持接收 |
| Help | [alonabot-plugin-help](https://github.com/AlonaBot/alonabot-plugin-help) | 按需安装 |

## 文档与贡献

| 你想做什么 | 从这里开始 |
| --- | --- |
| 安装、配置、升级或备份机器人 | [安装与部署](docs/deployment.md) |
| 编写插件、Adapter 或 Agent | [插件开发指南](docs/plugin-guide.md) |
| 修改 Core、SDK 或 WebUI | [源码开发](docs/development.md) · [架构地图](ARCHITECTURE.md) |
| 报告问题、提交修复或改善文档 | [贡献指南](CONTRIBUTING.md) |
| 查询精确的 API 与协议行为 | [技术规范](spec/README.md) |
| 构建与核对发行物 | [发行维护](docs/releasing.md) |

欢迎修复问题、改善跨平台行为、补充文档和开发社区组件。提交问题时附上版本、系统和复现步骤，日志与配置先脱敏；较大的改动请先通过 Issue 讨论。

## 许可证

AlonaBot 与官方组件采用 [MIT License](LICENSE)。
