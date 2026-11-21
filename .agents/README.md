# AlonaBot 个人维护工作地图

本文件与 `.agents/`、goal 调度、实施计划、ADR、历史验收材料只在 `dev` 维护。产品与公共工程文档以 `main` 为准；运行中的日志、凭据、截图、缓存、临时 worktree 和 ignored 验收目录留在本机。

## 分支与资料归属

- `main` 保存产品代码、现行架构、协议契约、编码规范、公开指南和可复用测试工具。
- `dev` 叠加个人 Agent 规则、技能、goal、计划及历史依据，与仓库共享可见性。只归档已受版本控制的规则和可复用资料，勿收录私密运行数据。
- 产品任务从最新 `origin/main` 创建独立 `codex/<主题>` 或 `goal-N` 分支与 worktree，PR 指向 main。将本 dev worktree 的路径与完整任务要求交给执行者，个人规则由执行者在此读取；产品 worktree 自带公共 AGENTS，个人技能和维护目录从 dev 读取。
- 产品变更合入 main 后，将 main 同步到 dev。同步时保留维护覆盖层；冲突按公共内容采用 main、个人内容留在 dev 处理。dev 全分支合并回 main 会携带个人资料，交付时只选择产品改动。
- goal 语句中的旧任务说明、每 task 自动 commit、同步 PLAN/TEST-MATRIX 到产品 PR 等文字按下文现行流程执行。历史材料用于追溯；再次派发前按当前 main 重写范围、接口与验证条件。

## 任务入口

| 任务 | 读取资料 | 完成检查 |
| --- | --- | --- |
| Go 实现、诊断或设计 | [开发技能](skills/alonabot-develop/SKILL.md)、产品 worktree 的 [架构](../ARCHITECTURE.md)、[编码规范](../docs/coding-standards.md)与[技术契约](../spec/README.md) | 所有者、调用者和必要相邻修正完整 |
| 注释 | [注释技能](skills/alonabot-code-commenting/SKILL.md) | 包文档、字段与阶段注释符合现状 |
| 验证 | [测试技能](skills/alonabot-test/SKILL.md)、[开发指南](../docs/development.md) | 命令、输入、平台、结果与未运行项明确 |
| 审查 | [审查技能](skills/alonabot-review/SKILL.md) | 行为契约与项目规范分别核对 |
| PR 或 CI | [PR 技能](skills/alonabot-pr/SKILL.md)、[公共贡献指南](../CONTRIBUTING.md) | 授权、当前 head 与检查有效 |
| 文档或维护技能 | [文档技能](skills/alonabot-docs/SKILL.md) | main 公共内容自包含，维护内容只写 dev |
| 执行维护 goal | [goal 入口](../docs/goals/README.md)、[执行技能](skills/alonabot-execute/SKILL.md) | 任务范围与实际 main 一致 |

本表中的公共文档链接便于在 dev 导航；执行产品任务时以产品 worktree 的对应文件为准。规格冲突先在产品分支修订现行契约及受影响实现、示例与验证；原因、计划和历史依据另外记在 dev。

## 本分支维护约定

公共工程约定以根 [AGENTS.md](../AGENTS.md) 为准。大型任务才使用 goal-mode；委派前读取 `read-only-delegation`，实现由主代理负责。

资源开销已被接受，后续资源测量、压测与优化保持停止。Windows 实跑继续沿用已批准的延期；主仓公开后，模板候选的 hosted CI 已通过。

## Git 授权与交付

- 提交、推送、合并、tag 与发布遵循根 AGENTS 和本次用户授权。
- 执行者交付产品 diff、验证与文档同步要点；PR 只含产品和公共文档。切片状态、派发语句、PLAN、TEST-MATRIX 与 ADR 的更新在 dev 另行完成。
- 审查者核对完整 diff、当前 head 的 CI 与关键复现；只有得到合并授权才 squash 合并。合并后确认 main 检查，再在 dev 更新依赖任务与验收来源。
- `goal-N/` 三份任务卡继续只留在本地。回复直接交付结果，避免新建单独报告。
