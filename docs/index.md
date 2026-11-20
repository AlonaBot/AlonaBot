# AlonaBot 文档

从你要完成的事情开始选择文档。精确行为由技术规范定义；指南提供可以直接使用的步骤和示例。

## 使用机器人

- [安装与部署](deployment.md)：发行归档、系统服务、Docker、管理认证、配置、备份与诊断。
- [v0.1.0 发布说明](release-notes/v0.1.0.md)：发行内容、平台验证、升级与已知限制。

## 开发组件

- [插件开发指南](plugin-guide.md)：创建工程、命令、事件、配置、存储、模型调用、生命周期与发布。
- [插件测试](development.md#插件测试)：用真实 Core 注入消息、检查回复、任务和 LLM 替身。
- [可运行示例](../examples/plugins)：echo、lifecycle、showcase、configdb 和 media。
- [开发机器人示例](../examples/devbot/README.md)：回环 Adapter、热重载和同进程调试。

## 贡献源码

- [贡献指南](../CONTRIBUTING.md)：Issue、依赖讨论、PR、审查与交付。
- [源码开发](development.md)：环境、构建、热重载、WebUI、生成与验证。
- [编码规范](coding-standards.md)：命名、文件组织、注释、日志与测试提交规则。
- [架构地图](../ARCHITECTURE.md)：模块职责、依赖方向与状态归属。
- [技术规范](../spec/README.md)：API、生命周期、消息模型、IPC 与管理接口。

## 维护发行

- [发行维护](releasing.md)：发行构建、测试签名、候选核对与正式发布步骤。
