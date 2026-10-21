# AlonaBot 代理工作地图

轻量级跨平台 Go 聊天机器人框架：Core + 独立进程 Native Go 插件 + Telegram Adapter。目前处于实施初期。

## 先读什么

| 任务 | 先读 | 完成前检查 |
| --- | --- | --- |
| 任何实现工作 | [spec/PLAN.md](spec/PLAN.md) 中对应的 T 工作包，及其引用的 CONTRACTS/WIRE/MESSAGE 小节 | 工作包的「停止和失败条件」 |
| 新增包、跨包 import | [ARCHITECTURE.md](ARCHITECTURE.md) | 依赖表允许这条边；守卫测试通过 |
| 构建、测试、依赖、外部环境 | [docs/development.md](docs/development.md) | 按需运行 `go generate ./...`、`go test ./...`、`go build ./...` |
| 记录验收 | [spec/TEST-MATRIX.md](spec/TEST-MATRIX.md) §3 | 写明测试位置、命令、平台和结果；没运行的写「未运行」 |
| 契约有缺口或冲突 | [spec/README.md](spec/README.md)「执行与维护」 | 先修契约、PLAN 和 TEST-MATRIX，再在 spec/adr/DECISIONS.md 记录原因 |
| 推进批次 goal | 当前分支上的 `goal-N/input.md`、`plan.md`、`tasks.md` | 每个 task 的回写区填写完整 |

## 项目约定

- 代码注释、commit message、文档一律用简体中文；标识符和协议名保留英文。
- 依赖只用 docs/development.md 中列出的已批准清单；需要新依赖时先停下来征求维护者同意。
- 生成代码只改源文件（`.proto`、生成器），然后重新运行 `go generate ./...`，不手改生成产物。
- 不把 fixture、模拟服务或交叉编译结果记成真实平台/环境的验证结果。
- 业务副作用不加自动重试；限额、超时等常量以 CONTRACTS 为准，不擅自增减。
- `experiments/` 是独立的历史实验，不 import，也不作为验收依据。

## Git

- `main` 只接收维护者确认后的合并。批次工作在 `goal-N` 分支上进行，工作树位于 `.worktrees/goal-N`；每个 task 一个 commit，不 push。
