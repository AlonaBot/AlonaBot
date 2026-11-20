> 维护归档：此处保留 2026-11-20 的原始记录。现行分支、派发与授权规则见根 [AGENTS.md](../AGENTS.md)；重新派发前须核对当前 main。

# 资源实测记录

2026-11-20，维护者接受当前资源开销，并要求停止该方面测量。后续资源压测与 NoneBot2 对照按 [D-184](../spec/adr/DECISIONS.md) 收口；本文记录已经取得的数据。

## 当前开销

单位均为 MiB（1 MiB = 1024 KiB）。PSS 使用 Linux `smaps_rollup`，私有内存为 `Private_Clean + Private_Dirty`，每项取三轮中位数。

| 进程 | PSS | RSS | 私有内存 | 历史 PSS 预算 |
| --- | ---: | ---: | ---: | ---: |
| Core-only | 26.61328125 | 26.62109375 | 26.61328125 | 24 |
| echo | 10.09375 | 10.1015625 | 10.09375 | 10 |
| configdb GORM | 16.21875 | 16.2265625 | 16.21875 | 16 |

三项均超历史 24/10/16 MiB 预算。维护者基于上表接受当前开销；历史失败和原始数值保留在 [TEST-MATRIX](../spec/TEST-MATRIX.md)。

## 三轮读数

| 进程 | 轮次 | PSS | RSS | 私有内存 |
| --- | ---: | ---: | ---: | ---: |
| Core-only | 1 | 26.64453125 | 26.65234375 | 26.64453125 |
| Core-only | 2 | 26.35546875 | 26.36328125 | 26.35546875 |
| Core-only | 3 | 26.61328125 | 26.62109375 | 26.61328125 |
| echo | 1 | 10.0859375 | 10.09375 | 10.0859375 |
| echo | 2 | 10.359375 | 10.3671875 | 10.359375 |
| echo | 3 | 10.09375 | 10.1015625 | 10.09375 |
| configdb GORM | 1 | 16.1796875 | 16.1875 | 16.1796875 |
| configdb GORM | 2 | 16.21875 | 16.2265625 | 16.21875 |
| configdb GORM | 3 | 16.2578125 | 16.265625 | 16.2578125 |

## 测量条件与原始记录

- 产品：`8b838dd8d3018154c663cc429354a526145f7640`，Go1.27.1、`CGO_ENABLED=0` 普通构建；echo 与 configdb GORM 经正式 pack/install。GORM 固定到组织补丁提交 `34b98340007ddc2acd5cb8727a0b4b72ea295896`。
- 环境：instabox 固定 Linux/amd64 VM，内核 `6.18.60-instabox`；自然文件缓存、默认 GC，采样与 CI 串行。
- 方法：每轮依次 Core-only、echo、configdb GORM，共九个独立 data_dir；Core-only 关闭全部插件和 Adapter，其余在真实 ACTIVE 后空闲120s取样。
- 执行：2026-11-20 05:00:27–05:18:30 UTC，job `e0183236` 退出0。十五个进程对象的 PID/starttime、六字段 smaps/rollup 合计、正式包字节、状态/业务 SQLite 和退出证据均已核对。

历史执行命令（前两条在原远端目录，第三条在 goal-174 worktree）：

```sh
uv run --no-project python candidate_resources.py inputs out
uv run --no-project python after.py
uv run --no-project python tmp/test/v74-8b/check.py goal-174/v74-8b-linux/alonabot-v74-8b-174-20261120 tmp/test/v74-8b/summary.json
```

[PR #277](https://github.com/AlonaBot/AlonaBot/pull/277) 列出完整来源和验证记录。该切片的本地 ignored 归档为 `goal-174/v74-8b-evidence.tar.gz`，展开目录为 `goal-174/v74-8b-linux/alonabot-v74-8b-174-20261120/`，包含测量脚本、输入、逐轮日志、maps/smaps、状态及业务数据库。汇总位于 `tmp/test/v74-8b/summary.json`，审查者独立核算结果为同目录 `parent-summary.json`。这些路径相对保留原始证据的 goal-174 worktree。

原有效的各100次冷启动、长空闲休眠、唤醒与数据保持证据沿用 TEST-MATRIX 的 S-06.2 记录。V-33 四负载、NoneBot2 的资源/吞吐/延迟对照按维护者决定取消新增运行；已有依赖兼容预检保留在原工作区。rss 随组件后置，help 未纳入本轮样本。
