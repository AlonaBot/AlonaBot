> 维护归档：此处保留 2026-11-20 的原始记录。现行分支、派发与授权规则见根 [AGENTS.md](../../AGENTS.md)；重新派发前须核对当前 main。

# 切片 goal 语句

本目录收录给执行者派发切片用的 goal 语句，按 [PLAN §3.1](../../spec/PLAN.md#31-执行切片) 的章节分文件；一个 goal 对应所属章节实施计划中的一个切片。并行派发的时机、共享文件和合并顺序按 [执行顺序](../execution-order.md)；执行者的工作方式按 [执行技能](../../.agents/skills/alonabot-execute/SKILL.md)。本目录只由审查者维护。

## 章节

| 章节 | goal 语句 | 实施计划 |
| --- | --- | --- |
| 核心 | [core.md](core.md) | [spec/plan/core.md](../../spec/plan/core.md) |
| Agent | [agent.md](agent.md) | [spec/plan/agent.md](../../spec/plan/agent.md) |
| 适配器 · Telegram | [adapters/telegram.md](adapters/telegram.md) | [spec/plan/adapters/telegram.md](../../spec/plan/adapters/telegram.md) |
| 适配器 · OneBot 11 | [adapters/onebot11.md](adapters/onebot11.md) | [spec/plan/adapters/onebot11.md](../../spec/plan/adapters/onebot11.md) |
| 插件 | [plugins.md](plugins.md) | [spec/plan/plugins.md](../../spec/plan/plugins.md) |
| WebUI | [webui.md](webui.md) | [spec/plan/webui.md](../../spec/plan/webui.md) |
| 集成与发布 | [release.md](release.md) | [spec/plan/release.md](../../spec/plan/release.md) |

## 用法

新任务先读取 dev 的 AGENTS 与对应历史依据，按当前 main 重写完整派发语句。产品 worktree 从 main 创建；本目录和实施计划留在 dev。对旧语句中要求在产品 PR 回写 PLAN、TEST-MATRIX 或本目录的步骤，改为由审查者在 dev 另行回写。用户未授权 commit/push 时交付本地 diff，待授权后继续。

1. 在章节文件的「总览」中找到状态为**待派发**的 goal。
2. 把该 goal 的整段语句（代码块内的全部文字）原样发给一个执行者。一个 goal 同时只交给一个执行者。
3. 执行者开出 PR 后，把 PR 链接交给审查者审查。
4. 审查者合并后，会更新依赖它的 goal 的状态和接口段。

各章节文件的「总览」和条目按 [执行顺序 §2](../execution-order.md#2-波次参考) 的全局波次排列，只列出本章节有切片的波次；同一波次内按切片号排列，因此章节内每个 goal 的前置都排在它前面，跨章节的前置按切片编号查找。波次以执行顺序为准；新增切片或前置变化时先改那张表，再调整章节文件中的位置。状态只写在总览和条目的状态行，状态变化时条目不移动。

状态只有以下三种：

- **待派发**：切片表中的前置都已完成，语句已按合并后的代码写到接口签名，可以直接派发。已派发、PR 尚未合并的 goal 保持此状态，直到合并。
- **等待前置**：前置尚未全部合并。语句可能已经写完整，也可能只写了目标、依据和要点（范围与要点已由 PLAN 确定，没有待决的设计问题）；前置合并后，审查者按实际代码核对或写出接口、行为和测试段，再改为待派发。每条的完整语义以所属章节实施计划中对应的工作包步骤及其失败停止条件为准，不另作设计。
- **已完成**：切片 PR 已由审查者合并，验证证据见 TEST-MATRIX §3.2。条目保留供后继切片查阅接口。

## 编写测试段

审查者写每条语句的「测试」段时遵守[编码规范「测试提交规则」](../coding-standards.md#测试提交规则)：

- 「提交」默认只列一个覆盖主路径的集成测试。场景只放规格要求的主路径和只有跨进程才会暴露的失败。
- 多列测试时，在该项后面写「多列理由：……」。改写既有测试（换 API、断言不变）不算新增。
- 其余场景放进「本地自测」，包括 V 编号的分项、边界穷举和审查中发现的回归；执行者在 PR 中写明结果。
- 前端纯函数的测试不进「提交」，放进「本地自测」，写在 `*.local.test.ts`。

## 所有 goal 的共同约束

每条 goal 语句都引用本节，执行者开始前必须完整读取。

1. 使用 `$alonabot-execute` 推进；大型任务再使用 goal-mode。goal 编号、分支和 worktree 使用语句里指定的值：从最新 `origin/main` 创建 `.worktrees/goal-N` 和分支 `goal-N`。
2. goal-mode 第一轮只做初始化：
   - `goal-N/input.md` 原样保存整段 goal 语句。目录被 `.gitignore` 忽略，执行记录只留在本地，不进仓库。
   - `plan.md` 和 `tasks.md` 按执行技能「切片卡格式」编写，把语句中的接口、行为和测试逐条落到 task 上，不增加也不删减范围。
3. 实现前读取 [AGENTS.md](../../AGENTS.md)、[编码规范](../coding-standards.md)、[执行顺序](../execution-order.md) 和语句「依据」列出的规格小节。
4. 只修改语句「范围」中列出的文件。公开接口照语句给出的签名实现；语句没写的内部辅助函数可以自定，但不新增公开符号。
5. 遇到执行技能列出的阻塞情况时标 `[!]`，不要自行决定。特别是：需要改 `.proto`、加依赖、改范围外文件、发现契约冲突。
6. Go 代码遵循 Google Go Style Guide，可用时加载 `go-google-style-guide`、`go-google-style-decisions`、`go-google-best-practices` 技能。包文档、命名、注释和测试按 [编码规范](../coding-standards.md) 执行，与 Google 风格冲突时以编码规范为准：每个新包在一个文件里写标准库风格的包文档，其他文件开头不写注释；名字用规格术语，不用含糊词和缩写；注释用简体中文，只写现状和原因，不写 T、goal 或 task 编号。
   - 只提交语句列出的集成测试，不自行增加。单元测试、回归和临时测试写在 `*_local_test.go`（前端纯函数写在 `*.local.test.ts`），它们被 git 忽略，不提交。
   - 语句的「测试」段要求向 `sdk/` 或 `internal/core/supervisor/` 提交测试，或新增 `testdata` 测试插件时，一律改为本地自测，PR 写明结果；本条优先于语句（D-86）。
   - 只在信任边界做校验，不写防御性的重复检查和兜底分支。
   - 日志、错误字符串、CLI 输出和协议 safe_message 一律用英文；有含义的数值和字符串写成命名常量。
   - 日志按[编码规范「日志用默认 Logger，依赖显式传递」](../coding-standards.md#日志用默认-logger依赖显式传递)取用默认 Logger，字段遵循 CONTRACTS §14，不用 fmt 直接输出。
   - 契约中可调的限额与期限从 `internal/tune` 读取，新增的可调值登记到 `internal/tune/params.go`（CONTRACTS §16）。
7. 每个 task 都运行语句中的完成条件命令。最后一个 task 核对现有完成条件结果；取得创建 PR 的明确授权后推送分支，开 PR 到 `main`：标题为「切片 ID + 标题」，正文按 PR 模板填写，并加一节「文档同步」，列出需要审查者同步到 ARCHITECTURE、README 或 spec 的内容。不合并 PR。
