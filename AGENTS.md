# AlonaBot 个人维护工作地图

本文件与 `.agents/`、goal 调度、实施计划、ADR、历史验收材料只在 `dev` 维护。产品与公共工程文档以 `main` 为准；运行中的日志、凭据、截图、缓存、临时 worktree 和 ignored 验收目录留在本机。

## 分支与资料归属

- `main` 保存产品代码、现行架构、协议契约、编码规范、公开指南和可复用测试工具。
- `dev` 叠加个人 Agent 规则、技能、goal、计划及历史依据，与仓库共享可见性。只归档已受版本控制的规则和可复用资料，勿收录私密运行数据。
- 产品任务从最新 `origin/main` 创建独立 `codex/<主题>` 或 `goal-N` 分支与 worktree，PR 指向 main。将本 dev worktree 的路径与完整任务要求交给执行者，个人规则由执行者在此读取；产品 worktree 无需复制 AGENTS、技能或维护目录。
- 产品变更合入 main 后，将 main 同步到 dev。同步时保留维护覆盖层；冲突按公共内容采用 main、个人内容留在 dev 处理。dev 全分支合并回 main 会携带个人资料，交付时只选择产品改动。
- goal 语句中的旧任务说明、每 task 自动 commit、同步 PLAN/TEST-MATRIX 到产品 PR 等文字按下文现行流程执行。历史材料用于追溯；再次派发前按当前 main 重写范围、接口与验证条件。

## 任务入口

| 任务 | 读取资料 | 完成检查 |
| --- | --- | --- |
| Go 实现、诊断或设计 | [开发技能](.agents/skills/alonabot-develop/SKILL.md)、产品 worktree 的 [架构](ARCHITECTURE.md)、[编码规范](docs/coding-standards.md)与[技术契约](spec/README.md) | 所有者、调用者和必要相邻修正完整 |
| 注释 | [注释技能](.agents/skills/alonabot-code-commenting/SKILL.md) | 包文档、字段与阶段注释符合现状 |
| 验证 | [测试技能](.agents/skills/alonabot-test/SKILL.md)、[开发指南](docs/development.md) | 命令、输入、平台、结果与未运行项明确 |
| 审查 | [审查技能](.agents/skills/alonabot-review/SKILL.md) | 行为契约与项目规范分别核对 |
| PR 或 CI | [PR 技能](.agents/skills/alonabot-pr/SKILL.md)、[公共贡献指南](CONTRIBUTING.md) | 授权、当前 head 与检查有效 |
| 文档或维护技能 | [文档技能](.agents/skills/alonabot-docs/SKILL.md) | main 公共内容自包含，维护内容只写 dev |
| 执行维护 goal | [goal 入口](docs/goals/README.md)、[执行技能](.agents/skills/alonabot-execute/SKILL.md) | 任务范围与实际 main 一致 |

本表中的公共文档链接便于在 dev 导航；执行产品任务时以产品 worktree 的对应文件为准。规格冲突先在产品分支修订现行契约及受影响实现、示例与验证；原因、计划和历史依据另外记在 dev。

## 项目约定

- 用简体中文沟通；文档、注释与 commit message 用简体中文，标识符保留英文。日志、错误、CLI 与协议 safe_message 用英文；界面和机器人回复通过 i18n 提供。
- Go 遵循 Google Go Style 与项目编码规范。首次发布前同步修改初始迁移、签名及全部调用方，按当前契约实现。
- 依赖只在已批准范围使用；新增前说明必要性、替代方案并取得明确批准。
- 生成代码修改源文件或生成器后重新生成。
- 只在外部输入和信任边界校验。业务副作用的重试、限额、期限及降级须有契约或实际故障依据；可调值统一登记在 `internal/tune`。
- 仓库只提交集成测试和协议编号检查；本地测试与证据保持 ignored。fixture、交叉构建和真实平台分别报告。
- 保留用户已有改动和数据。Git 写操作使用独立 worktree。跨模块大改动独立审查；委派前读取 `read-only-delegation`，实现由主代理负责。
- 大型任务才使用 goal-mode。资源开销已被接受，后续资源测量、压测与优化已停止；Windows 实跑和模板 hosted CI 的既有延期决定继续有效。

## Git 授权与交付

- 未经用户明确要求，不 commit、不 push。用户明确要求创建 PR 时，授权其必要的提交与推送；合并、tag、发布另按用户授权执行。
- 执行者交付产品 diff、验证与文档同步要点；PR 只含产品和公共文档。切片状态、派发语句、PLAN、TEST-MATRIX 与 ADR 的更新在 dev 另行完成。
- 审查者核对完整 diff、当前 head 的 CI 与关键复现；只有得到合并授权才 squash 合并。合并后确认 main 检查，再在 dev 更新依赖任务与验收来源。
- `goal-N/` 三份任务卡继续只留在本地。回复直接交付结果，避免新建单独报告。
