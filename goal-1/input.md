# goal-1 原始输入（逐字保留）

## 用户初始指令（2026-10-21）

```
阅读spec的内容，然后准备分批实施计划，记得通过 $init-project 先初始化项目结构，然后通过 $goal-mode技能持续工作
```

## 用户追加消息（初始化过程中）

```
关于组织，我建议你先通过gh cli $github  帮我抢占这个组织名称
```

## 初始化阶段的问答（问题与用户所选答案原文）

1. 问：Go module path 用哪个？（T-01 固定后，插件作者也会 import 它）
   答：`github.com/AlonaBot*开头吧，我打算创立一个组织，帮我选一个组织名`
   追问：用哪个组织名？ → 答：`AlonaBot (Recommended)`
2. 问：是否批准规格指定的这组依赖？goal 期间遇到清单外的依赖时，我会把对应 task 标为阻塞，不自行添加。
   答：`批准规格清单 (Recommended)`
   （清单：google.golang.org/protobuf（protoc-gen-go 用 go.mod 的 tool 指令固定版本）、gorm.io/gorm、github.com/libtnb/sqlite、modernc.org/sqlite、gorm.io/driver/postgres、golang.org/x/sys（flock/LockFileEx/Job Object）、go.yaml.in/yaml/v3（解析 plugin.yaml）。Telegram 直接用 net/http，cron 自己实现。）
3. 问：web/admin 用什么技术栈？
   答：`Vite + TypeScript + 框架`
   追问：Vite + TypeScript 搭配哪个框架？我会按这个选择引入对应 npm 依赖（Vite、TypeScript、框架，再加 vitest 做组件测试），选定即视为批准。 → 答：`React 19`
4. 问：Git 怎么处理？goal-mode 的 SKILL.md 第 5 步要求每个 task 提交一次，这和你的规则「未经明确要求不 commit、Git 写操作在独立 worktree」冲突。
   答：`init + goal 分支每 task 提交 (Recommended)`
   （选项说明：git init，在 main 上做一次初始提交（spec+骨架），之后在独立 worktree 的 goal-1 分支上每个 task 提交一次；不 push）
5. 问：这个 goal 覆盖哪些批次？
   答：`只做 B1 (Recommended)`
   （选项说明：goal-1 = 项目骨架 + T-01 + T-02；终审后你看过结果，再开 goal-2 做 B2）
6. 问：T-01 需要验证 PostgreSQL 分支，是否允许用 docker 临时起一个 PG 容器？
   答：`允许临时容器 (Recommended)`
   （选项说明：使用 postgres:17 镜像，只绑定 127.0.0.1 的随机端口，容器名 alonabot-test-pg，测试结束后删除；PG 测试用环境变量启用，没有配置时跳过并标为未运行）

## 初始化时提出的分批草案（供参照，非用户原文）

B1：项目骨架 + T-01 + T-02；B2：T-03 + T-04；B3：T-05 + T-08；B4：T-06 + T-07；B5：T-09；B6：T-10。
