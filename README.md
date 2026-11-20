<p align="center">
  <img src="docs/brand/mascot.png" alt="AlonaBot" width="200">
</p>

<h1 align="center">AlonaBot</h1>

<p align="center">
  <a href="https://github.com/AlonaBot/AlonaBot/actions/workflows/ci.yml"><img src="https://github.com/AlonaBot/AlonaBot/actions/workflows/ci.yml/badge.svg?branch=main" alt="Quality gate"></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/Go-1.27.1-00ADD8?style=flat-square&amp;logo=go&amp;logoColor=white" alt="Go 1.27.1"></a>
  <a href="docs/release-notes/v0.1.0.md"><img src="https://img.shields.io/badge/v0.1.0-发布候选-22B8CF?style=flat-square" alt="v0.1.0 发布候选"></a>
</p>

<p align="center">轻量级、跨平台的 Go 聊天机器人框架。<br>按需安装组件，把机器人搭成你需要的样子。</p>

<p align="center">
  <a href="#快速开始">快速开始</a> ·
  <a href="docs/index.md">文档</a> ·
  <a href="docs/plugin-guide.md">插件开发</a> ·
  <a href="CONTRIBUTING.md">参与贡献</a>
</p>

## 为什么选择 AlonaBot

- **进程隔离**：Native Go 插件、Adapter 与 Agent 各自运行，Core 管理启动、停止与故障恢复。适合信任的本机组件；独立进程仍需管理员审查来源。
- **组件可替换**：聊天平台由 Adapter 接入，AI 能力由 Agent 提供，管理界面由 WebUI 提供。社区组件可以替换官方实现。
- **按需运行**：未启用的组件不启动进程；支持休眠的插件可在空闲时退出，由消息、服务、HTTP 或定时任务唤醒。
- **Go 开发体验**：链式注册命令与事件，支持热重载、受管 SQLite、配置表单、i18n，以及通过真实 Core 运行的插件测试套件。
- **统一管理**：WebUI 与 CLI 管理组件、账号、模型、权限、会话、配置、日志和更新，支持桌面与窄屏。

官方 Agent 提供流式对话、工具调用、MCP、技能、人格、记忆与定时任务。模型支持 OpenAI、OpenAI Compatible、Responses、Anthropic、Gemini 与 Ollama；模型 SDK 按需链接在 Agent 中。

## 快速开始

v0.1.0 目前为发布候选，官方发行归档和镜像待发布。现在可以从源码运行本地示例，无需聊天平台或模型凭据。准备 Go **1.27.1** 后执行：

```fish
git clone https://github.com/AlonaBot/AlonaBot.git
cd AlonaBot
go build -o bin/alonabot ./cmd/alonabot
bin/alonabot run ./examples/plugins/lifecycle --data-dir ./data
```

在同一个终端输入 `/ping`，即可看到 `pong`。修改示例源码后会自动重载；按 Ctrl-C 停止。使用独立的 `data` 目录，避免与正在运行的机器人共用数据。

需要接入聊天平台和 AI 时，官方整合包包含 Core、Telegram Adapter、Agent、WebUI 与 help。首次启动启用 help 和 WebUI，配置 Telegram 实例后启用 Adapter，再按需启用 Agent 并配置模型。正式发布后的归档、Docker Compose 与系统服务安装步骤见[安装与部署](docs/deployment.md)，平台实跑情况与已知限制见[v0.1.0 发布说明](docs/release-notes/v0.1.0.md)。

## 写一个插件

组件导出 `New()`，登记需要的行为，构建与运行入口由工具生成：

```go
package hello

import (
    "github.com/AlonaBot/AlonaBot/sdk"
    "github.com/AlonaBot/AlonaBot/sdk/message"
)

// New 创建提供 /ping 命令的插件。
func New() *sdk.Plugin {
    p := sdk.NewPlugin()
    p.Command("ping").Handle(func(*sdk.Context) (message.Content, error) {
        return message.Text("pong"), nil
    })
    return p
}
```

用 `alonabot init` 创建包含 manifest、配置 Schema 和 i18n 的工程，用 `alonabot run` 开发、`alonabot test` 验证、`alonabot pack` 打包。完整步骤见[插件开发指南](docs/plugin-guide.md#从第一个命令开始)；可运行的[示例](examples/plugins)涵盖命令、事件、媒体、数据库、服务与调度。

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
