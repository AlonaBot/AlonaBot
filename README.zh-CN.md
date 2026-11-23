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
  <a href="https://qm.qq.com/q/YD36YEiCqI"><img src="https://img.shields.io/badge/QQ-Group-12B7F5?style=flat-square&amp;logo=qq&amp;logoColor=white" alt="QQ Group"></a>
</p>

<p align="center">轻量级、跨平台的 Go 聊天机器人框架。<br>按需安装组件，把机器人搭成你需要的样子。</p>

<p align="center">
  <a href="https://alonabot.github.io/docs/">文档</a> ·
  <a href="docs/plugin-guide.zh-CN.md">插件开发</a> ·
  <a href="changelog/README.zh-CN.md">更新日志</a> ·
  <a href="CONTRIBUTING.zh-CN.md">参与贡献</a> ·
  <a href="https://github.com/AlonaBot/AlonaBot/wiki">Wiki</a> ·
  <a href="https://github.com/AlonaBot/AlonaBot/discussions">讨论区</a> ·
  <a href="https://github.com/AlonaBot/AlonaBot/issues/new/choose">报告问题</a> ·
  <a href="SECURITY.md">安全报告</a>
</p>

## 核心特性

- **故障隔离**：独立进程部署时，插件、Adapter 与 Agent 各自运行，单个组件崩溃不会拖垮主程序，由 Core 负责启停与自动拉起。
- **模块解耦**：平台接入（Adapter）、AI 能力（Agent）与管理界面（WebUI）均与 Core 分离，不绑定官方实现，可随时替换为社区或自研组件。
- **轻量低耗**：未启用的组件不启动进程；支持休眠的插件空闲时自动退出以释放内存，在收到消息、HTTP 请求、定时任务或跨组件调用时自动唤醒。
- **插件开发**：提供链式命令与事件注册、热重载、开箱即用的 SQLite 存储、自动渲染为 WebUI 表单的配置声明、i18n，以及基于真实 Core 的测试套件。
- **集中运维**：通过 WebUI 与 CLI 统一处理组件更新、账号与权限、模型、会话配置和运行日志，同时适配桌面端与移动端。

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
| 安装、配置、升级或备份机器人 | [安装与部署](docs/deployment.zh-CN.md) |
| 编写插件、Adapter 或 Agent | [插件开发指南](docs/plugin-guide.zh-CN.md) |
| 修改 Core、SDK 或 WebUI | [源码开发](docs/development.zh-CN.md) · [架构地图](ARCHITECTURE.md) |
| 报告问题、提交修复或改善文档 | [贡献指南](CONTRIBUTING.zh-CN.md) |
| 查询管理 API 与协议行为 | [管理 API](docs/deployment.zh-CN.md#管理-api) · [协议与会话](docs/development.zh-CN.md#协议与会话行为) |
| 构建与核对发行物 | [发行维护](docs/releasing.md) |

欢迎修复问题、改善跨平台行为、补充文档和开发社区组件。提交问题时附上版本、系统和复现步骤，日志与配置先脱敏；较大的改动请先通过 Issue 讨论。

## 许可证

AlonaBot 与官方组件采用 [MIT License](LICENSE)。
