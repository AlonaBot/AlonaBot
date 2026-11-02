---
name: alonabot-execute
description: 执行者按 docs/goals/ 中的 goal 语句推进 AlonaBot 切片时使用；用 goal-mode 初始化切片卡，一轮一个 task，按语句给出的范围和接口实现并取证，完成后推送分支并开 PR，审查与合并由审查者负责。
---

# 执行 AlonaBot 切片

[PLAN §3.1](../../../spec/PLAN.md#31-执行切片) 把工作包分为若干章节，各章节实施计划（`spec/plan/`）再把工作包拆成切片，一个切片对应一个 goal：`goal-N` 目录、分支和 PR。审查者在 [docs/goals/](../../../docs/goals/README.md) 按章节写出每个 goal 的派发语句；执行者按本技能推进。并行时的共享文件和同步规则见 [执行顺序](../../../docs/execution-order.md)，角色和合并授权见 [贡献流程](../../../CONTRIBUTING.md#角色与合并)。

## 开始前

1. 从最新的 `origin/main` 创建 `.worktrees/goal-N` 和分支 `goal-N`，N 用 goal 语句指定的编号，不按 goal-mode 的默认规则自选。切片卡写在 worktree 的 `goal-N/`，该目录已被 `.gitignore` 忽略。
2. 完整读取 docs/goals/README.md「所有 goal 的共同约束」、根 [AGENTS.md](../../../AGENTS.md)、[编码规范](../../../docs/coding-standards.md)、执行顺序，以及 goal 语句「依据」列出的规格小节。实现时按 [开发技能](../alonabot-develop/SKILL.md)、[测试技能](../alonabot-test/SKILL.md) 和 [注释技能](../alonabot-code-commenting/SKILL.md) 执行。
3. goal-mode 第一轮只做初始化：
   - `goal-N/input.md` 原样保存整段 goal 语句。`goal-*/` 已被 git 忽略，切片卡和回写区只留在本地，不要提交。
   - `plan.md` 和 `tasks.md` 按下文「切片卡格式」编写。语句中的每个接口、行为要点和测试都要落到某个 task 上，不增加也不删减范围。

## 每一轮只做一个 task

1. 重新读取三份 goal 文件，选出第一个状态为 `[ ]` 的 task。已登记的 FIX-n 排在它们之前，优先执行。
2. 只修改该 task「范围」中列出的文件和包。按「接口」给出的签名实现，不自行增加公开接口、选项或语句没要求的功能。
3. 「测试」中列为「提交」的集成测试必须写出并通过；列为「本地自测」的写在 `*_local_test.go`，跑通后不提交，在回写区记下名字和结果。
4. 运行「完成条件」中的命令，结果必须与预期一致。改了 `.proto` 时运行 `go generate ./...`，并确认没有额外 diff。
5. 每个 task 一个 commit，message 用简体中文写明做了什么，并带上 task 编号。代码注释和错误消息里不写 task、goal 或 T 编号。
6. 回写 tasks.md：实际做了什么、验证证据（命令、结果、commit）、剩余风险、下一步。该文件被 git 忽略，只改本地，不放进 commit。

REVIEW task 要按 goal-mode 的集中检查清单复查前面几个 task，发现的问题登记为 FIX-n。

## 遇到以下情况标阻塞，不要自己决定

- 契约有缺口或冲突，或者需要修改 proto 字段号或枚举值。
- 需要清单之外的依赖。
- 需要修改「范围」之外的文件（gofmt 和生成代码的连带变化除外）。
- goal 语句给出的接口签名无法实现，或者与现有调用者冲突。
- 验证需要真实 Telegram、目标平台或凭证。
- 没有改动的测试在基线上就已失败。

把 task 标为 `[!]`，写明原因、已尝试的做法和需要谁提供什么，然后继续下一个不依赖它的 task。

## 不允许的操作

- 合并 PR、推送 `main`、force-push、rebase 已推送的分支、删除分支或 tag。
- 修改其他 goal 目录、CONTRACTS 或 WIRE；PLAN 和 TEST-MATRIX 只在 task 明确要求时更新。
- 把模拟服务、fixture 或交叉编译的结果写成真实平台的验证结果。
- 添加契约没有规定的限额、超时、重试或降级。

## 交付和处理审查意见

1. 所有 task 都已完成（`[x]`）或已标阻塞（`[!]`）后，在 goal 分支上运行 plan.md「验证方式」中的全部命令。
2. 推送 `goal-N`，按 [PR 模板](../../../.github/PULL_REQUEST_TEMPLATE.md) 开 PR 到 `main`。标题写成「切片 ID + 中文标题」，正文列出每个 task 的结果、验证证据、未运行的项和阻塞项，并加一节「文档同步」，写明需要审查者同步到 ARCHITECTURE、README 或 spec 的内容。有阻塞时开草稿 PR。无法创建 PR 时，只推送分支，并在 PR 描述里说明原因；`goal-N/` 的切片卡不进仓库，PR 正文就是唯一的执行记录。
3. 审查者通过 PR review 提出修改要求。每条意见在本地 tasks.md 末尾登记一个 FIX-n，一轮处理一个，之后推送普通 commit，并在 PR 中回复对应的 commit SHA。不同意某条意见时，在 PR 中附上证据回复，留给审查者决定。

维护者把 docs/goals/ 中该 goal 的整段语句原样发给执行者，以此启动切片。

## 切片卡格式

执行者在 goal-mode 初始化时，根据 goal 语句写出以下三份文件，并作为分支上的第一个 commit。

- **input.md**：整段 goal 语句原文，末尾补一行基线 commit。
- **plan.md**：
  - 依据（PLAN 步骤、CONTRACTS、WIRE、MESSAGE 小节）。
  - 可复用的现有代码（文件和符号）。
  - goal 语句已经给定的设计：类型、接口签名、文件位置、错误码。照抄，不改写。
  - 默认假设、不在本切片内的内容。
  - 验证方式（完整命令）和回滚方式。
- **tasks.md**：每个 task 按下面的格式写。每三个 task 之后插入一个 REVIEW，最后一个是交付 task（运行全部验证命令、推送、开 PR）。

```markdown
### [ ] GN-01 标题
- 依据：PLAN T-xx 步骤 k；CONTRACTS §x.y
- 范围：可新增或修改的文件和包；明确不得修改的部分
- 接口：要新增或修改的类型与函数签名
- 行为：必须满足的不变量和错误码
- 测试：提交的集成测试 TestXxx 断言什么；本地自测覆盖什么
- 完成条件：命令和预期结果
- 回写：
  - 实际做了什么：
  - 验证证据：
  - 剩余风险：
  - 下一步：
```

一个 task 要能在一轮内完成，通常只涉及一两个包。goal 语句没有覆盖、需要自行判断的地方，在 plan.md 写成默认假设；如果这个判断会改变公开接口、契约或范围，就标阻塞。

审查者编写 goal 语句时，要写到执行者不需要自己推断契约的程度：目标、依据、范围（含允许的新跨模块边）、接口签名、行为要点、测试（分提交的集成测试与本地自测）、完成条件和不做的事。
