# 参与 AlonaBot 开发

[English](CONTRIBUTING.md)

欢迎修复问题、改善文档、完善平台支持和开发社区组件。本文说明从反馈问题到提交 PR 的流程；环境和命令见[开发指南](docs/development.zh-CN.md)，编码与测试规则见[编码规范](docs/coding-standards.zh-CN.md)。

组件变更提交到对应的[官方组件仓库](README.zh-CN.md#官方组件)。本仓库维护 Core、SDK、协议、示例与整合发行。

## 从一个明确的问题开始

使用求助与方案交流前往[讨论区](https://github.com/AlonaBot/AlonaBot/discussions)。安全漏洞按[安全政策](SECURITY.md)通过私密报告提交；[Wiki](https://github.com/AlonaBot/AlonaBot/wiki) 提供任务导航。

先搜索已有的 Issue、打开和已合并的 PR，并核对当前源码。已有修复时，可以帮助复现或审阅；准备较大的功能、协议或数据改动时，先开 Issue 说明问题、预期行为与方案，等待维护者确认方向。

- **Bug**：提供版本、操作系统与架构、复现步骤、预期结果、实际结果和完整报错。尽量缩小到可重复的示例。
- **功能建议**：说明具体使用场景、现有能力的缺口和期望效果。平台集成优先考虑独立组件。
- **文档改进**：指出读到哪一步遇到困难，以及对应命令或行为；小修正可以直接提交 PR。

分享日志、配置或诊断包前检查并脱敏，移除 Bot Token、管理 token、数据库凭据和真实用户数据。疑似凭据泄露或安全漏洞请按[安全政策](SECURITY.md)通过私密报告提交，避免在公开 Issue 中披露可利用细节。

## 准备开发环境

先按[源码开发](docs/development.zh-CN.md#准备环境)准备所需工具。只改 Go 代码通常只需要 Go；修改协议生成产物还需要 protoc，修改 WebUI 还需要 Node 与 npm。

从最新 main 创建一个用途明确的分支；通过 Fork 贡献时，先同步上游 main。分支名可以采用 `fix/问题`、`feat/功能` 或 `docs/主题`。一个 PR 解决一个连贯问题，避免夹带无关重构、个人配置或临时验收产物。使用 worktree 可以同时处理多个独立任务。

## 同步变更

- 修改生成源后重新生成产物。
- 行为变更同步受影响调用方、示例与指南；数据库、进程生命周期和公开协议变更说明失败处理与恢复方法。
- 现有依赖范围及链接开销见[依赖管理](docs/development.zh-CN.md#依赖管理)。

## 提交 PR

1. 使用简体中文 commit message，说明改了什么和必要原因，无固定类型前缀要求。
2. PR 标题统一为 `type/scope: 中文摘要`，例如 `fix/scheduler: 修复休眠唤醒后的任务交付`。`type` 使用 `feat`、`fix`、`docs`、`refactor`、`test`、`ci`、`build`、`chore`、`perf` 或 `revert`；`scope` 使用实际模块的小写英文名称，可用连字符连接。摘要说明技术变化，省略任务编号、模型与执行工具署名。
3. 将 PR 指向 main，按[PR 模板](.github/PULL_REQUEST_TEMPLATE.md)填写变更原因、验证、兼容性与风险，并完成清单。关联 Issue 时使用 `Closes #编号`。
4. 核对 diff：仅包含本次改动及必要相邻修改，排除凭据、个人工作记录、运行数据和临时测试。
5. 等待当前 head 的检查与维护者审阅；根据意见补充修正和受影响验证。

维护者确认范围、实现、验证及延期项后使用 squash 合并，并核对 main 的合并结果。贡献者通过 PR 交付；正式版本、签名和发布由维护者控制。

## CI 与审查

- [Quality gate](.github/workflows/ci.yml) 检查空白与 main 文件范围；PR 修改 Go 源码、协议或 CI 时，再运行格式、生成一致性、vet、受影响包的普通测试和 Go 构建，`tests/` 改动运行该 module 的普通测试。完整门禁只在手动运行和发布时执行，并行运行 race、跨组件整合测试、五目标 pure-Go 构建与发行工具检查。前端验证由改动者按开发指南执行。
- [PR metadata](.github/workflows/pr-metadata.yml) 检查打开且非草稿 PR 的标题格式、必要说明与确认项；维护者审查其实际内容。
- 路径标签、指派与审阅请求由 GitHub 自动化处理；元数据工作流保持代码与 secret 隔离。
- main 保存产品、现行规范和公共工程资料。个人 Agent 配置、任务调度记录、历史实施与验收材料由 CI 路径检查拒绝。

分支保护配置由维护者管理；具体 CI 命令与权限以工作流为准。

## 贡献许可

AlonaBot 与官方组件采用 [MIT License](LICENSE)。提交贡献时，请确认你有权以该许可证提供相关代码与文档；引用第三方材料时保留其版权与许可声明。

## 编码规范

Go 风格、命名、注释、日志、兼容性与测试提交规则见[编码规范](docs/coding-standards.zh-CN.md)。
