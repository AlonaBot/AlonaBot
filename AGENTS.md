# AlonaBot 开发约定

本仓库维护 Core、SDK、协议、示例和整合发行。官方组件在各自仓库维护，入口见 [README](README.zh-CN.md#官方组件)；安装与使用文档在[文档站](https://github.com/AlonaBot/alonabot.github.io)。

## 先读什么

维护技能位于 `.agents/skills/`，客户端未自动发现时直接读取下表链接；只加载本次任务需要的入口。

| 任务 | 先读 | 完成前检查 |
| --- | --- | --- |
| 实现、诊断、重构或设计 Go 代码 | [开发技能](.agents/skills/alonabot-develop/SKILL.md)、[架构地图](ARCHITECTURE.md)、[编码规范](docs/coding-standards.zh-CN.md) | 所有者、调用者、允许的依赖边与必要相邻修改完整 |
| 新增或修改代码注释 | [注释技能](.agents/skills/alonabot-code-commenting/SKILL.md) | 包文档、文档注释、字段与 step 注释符合现状 |
| 选择测试、复现失败或记录验证 | [测试技能](.agents/skills/alonabot-test/SKILL.md)、[源码开发](docs/development.zh-CN.md#验证改动) | 只提交集成测试与协议编号检查；本地测试写在 `*_local_test.go` |
| 审查分支、commit 或 PR | [审查技能](.agents/skills/alonabot-review/SKILL.md) | 行为与规范两轴，问题有位置、影响和修复方案 |
| 准备或更新 PR、诊断 GitHub Actions | [PR 技能](.agents/skills/alonabot-pr/SKILL.md)、[贡献流程](CONTRIBUTING.zh-CN.md) | 授权、模板、当前 head 的检查及真实未运行项 |
| 修改文档、AGENTS/CLAUDE 或仓库技能 | [文档技能](.agents/skills/alonabot-docs/SKILL.md) | 权威单一、链接有效、中英同步 |
| 新增包或跨包 import | [架构地图](ARCHITECTURE.md) | 依赖表允许这条边，或同步修改依赖表 |
| 编写官方插件、Adapter、Agent 或示例 | [插件指南](docs/plugin-guide.zh-CN.md)、[Agent API](docs/agent-api.zh-CN.md) | 依赖、存储与常驻内存符合指南 |
| 构建、依赖、生成代码 | [源码开发](docs/development.zh-CN.md) | 按需运行 `go generate ./...`、`go test`、`go build ./...` |
| 打包与发布 | [发行维护](docs/releasing.md) | 发布授权与版本记录 |

## 工作约定

- 沟通、内部维护说明、注释和 commit message 使用简体中文；用户使用与开发指南提供中英内容；标识符、日志、错误与 CLI 输出使用英文。用户界面通过 i18n 提供。
- 先读取受影响实现、调用方和当前 diff；保留用户已有改动，Git 写操作使用独立 worktree。提交、推送、合并和发布按本次明确授权执行。
- 新依赖先说明必要性与替代方案，取得维护者明确批准。生成代码修改源文件后重新生成，不手改生成产物。
- 只在外部输入与信任边界进行必要校验；错误保留原始原因，业务副作用的重试与超时遵循现行行为，可调参数登记在 `internal/tune`。
- 首次发布前不写兼容层：改签名或语义时同步全部调用方（[编码规范](docs/coding-standards.zh-CN.md#首次发布前不写兼容层)）。
- 验证与改动规模匹配。文档和模板检查内容、链接及空白；代码运行受影响检查，跨模块变更独立审查。不把 fixture、模拟服务或交叉编译结果记成真实平台的验证。
- ADR、spec、goal、计划及历史验收资料在 dev 分支维护，CI 拒绝它们进入 main；凭据、运行数据和临时文件留在本机。
