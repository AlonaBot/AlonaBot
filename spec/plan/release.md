# 实施计划：集成与发布

总目标、章节划分与跨章节依赖见 [PLAN](../PLAN.md)；本章节的 goal 语句见 [release.md](../../docs/goals/release.md)。切片的拆分规则、状态与跨任务验收规则见 [PLAN §3.1](../PLAN.md#31-执行切片) 与 §5。

## 1. 范围

本章节覆盖 v0.1 的发布门槛：全部验收的汇总与跨组件联验、五个目标平台的构建与运行、资源报告、发布资产（T-10），以及发行归档、首次运行、容器镜像、install.sh 与端到端验收（T-14 步骤 3～7）。它依赖其他所有章节，不承担任何组件的功能实现。

发行包、预装配置与安装脚本按 CONTRACTS §15.2 的官方包 ID 使用连字符拼写（D-172）。与 T-09.27 并行的发布切片按共享文件规则同步新 ID，首次发布前使用重建的状态库与开发数据目录取证。

## 2. 执行切片

| 切片 | goal | 内容 | 工作包步骤 | 前置 | 主要验收 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| T-10.1 | goal-46 | V-01～V-76、V-82～V-117 证据汇总与跨组件联验 | T-10 步骤 1～4 | T-02.3、T-03.11、T-03.23、T-04.10、T-04.11、T-05.5、T-08.9、T-08.13、T-09.11、T-09.27、T-11.4、T-12.7、T-13.4、T-14.6、T-14.7、T-16.2（D-171） | 全部；真实 Telegram 与 LLM Provider 需要测试凭证 | 未开始 |
| T-10.2 | goal-47 | 五个目标组合的构建与运行、100 次启停 | T-10 步骤 5 | T-10.1 | V-32、V-34；需要真实目标平台 | 未开始 |
| T-10.3 | goal-48 | 资源比较报告与常驻内存预算核对 | T-10 步骤 6 | T-10.1 | V-33、V-74；需要固定机器与等价负载 | 未开始 |
| T-10.4 | goal-49 | 发布资产、签名与发布说明 | T-10 步骤 7、8 | T-10.2、T-10.3 | V-28、V-41 的发布复验 | 未开始 |
| T-14.3 | goal-73 | 发行归档：构建与打包首个发行版的组件（D-171）、签名、内置官方发布者公钥、`SHA256SUMS` | T-14 步骤 3 | T-03.10、T-09.3、T-12.7、T-14.1 | V-66 的归档分项 | 已完成（PR #228；TEST-MATRIX §3.2） |
| T-14.4 | goal-74 | 首次运行：预装与默认启用（help 与 WebUI）、首次管理 token、内置包的可用更新 | T-14 步骤 4 | T-09.6、T-14.3 | V-66 | 未开始 |
| T-14.5 | goal-77 | 容器镜像：`Dockerfile` 多阶段构建、linux/amd64 与 linux/arm64、非 root、`/data` 卷、`ALONABOT_CONTAINER` 的监听规则、作为官方整合包入口的 `deploy/compose.yaml` | T-14 步骤 5 | T-07.6、T-14.3 | V-70 的镜像分项 | 未开始 |
| T-14.6 | goal-75 | 从发行归档与容器镜像到 Agent 对话的端到端验收，汇总证据 | T-14 步骤 6 | T-09.11、T-14.4、T-14.5 | V-66、V-67、V-70 | 未开始 |
| T-14.7 | goal-85 | `install.sh`：Linux（systemd）与 macOS（launchd）的安装、`SHA256SUMS` 校验、管理 token 初始化、服务注册、升级与 `--uninstall` | T-14 步骤 7 | T-14.3 | V-84 | 未开始 |

## 3. 工作包

### T-10：集成、目标平台与发布验收

#### 前置输入

- T-02～T-09、T-11～T-16 的实现、正式示例、构建产物和各阶段测试证据；76 项用例的未闭合部分可逐项定位。
- 独立 Telegram 测试账号/聊天/话题、LLM Provider 的测试 API key、五目标运行环境，以及资源比较使用的机器和等价工作负载。

#### 拟涉及目录

复用各包测试、`examples/plugins/` 和构建/发布脚本；结果按 [TEST-MATRIX.md](../TEST-MATRIX.md) 的证据字段记录，链接到实际测试位置与报告。发布资产目录和 CI 配置在实施时建立。

#### 按顺序实施

1. 汇总 V-01～V-76 与 V-82～V-117 的证据，逐项标记已通过、失败、未运行或待联验。不进入首个发行版的组件（官方 OneBot 11 Adapter、welcome、rss 与 Telegram 的后续切片，D-171）的分项标为「随后续版本」，不阻塞本次发布。先补齐组件之间的联验：首次管理登录→安装启用→配置/权限→真实消息→插件 HTTP→重载/升级/回退；覆盖 TCP 与 stdio。
2. 使用不同的真实插件二进制验证命令、文本匹配、多轮等待、主动 Send、Service、Event Bus、Scheduler、HTTP、`sqlitedb` 与 GORM 两个数据库入口，以及官方 Agent 组件与 WebUI 组件。示例既覆盖无配置/无 DB 插件，也覆盖实际写入与迁移的插件；不以同一个 worker 的多份进程代替。
3. 执行协议、权限和资源边界测试：伪造/过期 Session、命令遮蔽、无 Actor、实例许可、Hook 递归、媒体租约、满准入、Cancel 顺序、超时仍运行的 job。验证错误回执与已经确认的外部结果保持一致。
4. 执行完整恢复故障注入：停止与持锁、SQLite WAL 备份、候选 OnStart 写入后失败、PREPARED/提交前后 Core 退出、不可变 secret、external 插件的维护许可与 resume、显式 rollback 的新代次。核对数据库内容和权威指针，不能只检查日志出现“恢复成功”。
5. 在五个目标组合分别构建并运行 Core 与插件，覆盖启动、停止、至少 100 次启停、异常退出、合作型子进程树和端口清理。记录 CGO 插件实际支持范围；无运行环境的目标保持未运行。
6. 在固定机器上按 TEST-MATRIX V-33 的四种负载（空闲、消息、数据库、HTTP）测量 Core+插件及附属进程的总 RSS/私有内存、Linux PSS、CPU 与 P50/P95/P99，并用同一模拟服务与功能相同的插件测量 NoneBot2。再按 V-74 测量 Core 与示例插件空闲时的 PSS，对照预算。输出原始数据和复现命令，不预设优胜结论；超出预算按 CONTRACTS §10 处理。
7. 生成五平台候选资产（首个发行版的组件：官方 Telegram Adapter、Agent、WebUI 与 help，D-171）、T-14 的发行归档、容器镜像与 install.sh、外部 release.json 和签名；用干净测试 data_dir 从这些真实资产完成安装、更新和卸载重装。核对来源、长度、摘要、平台和 Schema/二进制一致性，确保安装不依赖开发机的未打包文件。
8. 关闭全部阻塞用例，形成版本、平台、依赖、限制和恢复步骤的发布说明。准备好的发布包与验收记录构成本项交付。维护者用 T-14 的 `scripts/dist/` 在干净环境构建并签名，把发行归档、各安装包及其 `release.json`/`.sig` 发布到本仓库的 GitHub Release，把容器镜像推送到 `ghcr.io/alonabot/alonabot`；签名私钥只在维护者控制的环境中使用。

#### 可审查交付

- V-01～V-76、V-82～V-117 的完整结果与日志、故障注入后的状态/数据证据。
- 五平台真实运行记录和可安装候选资产；可复现的资源比较报告。
- 安装、更新、维护和恢复的用户操作步骤，以及与产物一致的 SDK 示例。

#### 停止和失败条件

协议/身份、Core 全局权限、数据代次唯一性、真实恢复边界、归档来源/越界以及目标平台构建运行等发布阻塞项必须实际通过。发现回归回到负责的 T 任务修复并重跑受影响场景；没有具体风险或必需门槛时不扩张测试范围。全部必要结果齐全后结束本项。

### T-14：官方整合包与预装组件

**目标：** 提供 AstrBot 式的开箱即用：用户从本仓库的 Release 用发行归档、`install.sh` 或 `docker compose` 安装官方整合包，在 WebUI 中配置即可使用；首个发行版的预装插件只有 help，welcome 与 rss 完成后随后续版本加入（CONTRACTS §15.2、§15.3、§15.6，D-32、D-64、D-171）。

#### 前置输入

- T-09 的安装、启用、更新、`alonabot run` 与 WebUI 组件；T-07.6 的 `core_config.json` 与首次 token 文件；T-12 的 Agent 组件；T-03 的 Telegram Adapter 包（T-15 的 OneBot 11 Adapter 包不进入首个发行版，D-171）；T-05、T-07、T-08 的事件、调度、配置、文案与数据库。
- 读取 CONTRACTS §9、§15.2、§15.3、§15.6。签名私钥由维护者保管，不进入仓库、构建日志或发行归档。

#### 实施步骤

1. **help 插件与 `core.command.list`。** 在[插件实施计划](plugins.md#t-14-的预装插件步骤)中实施。
2. **welcome 与 rss。** 在[插件实施计划](plugins.md#t-14-的预装插件步骤)中实施；不进入首个发行版，不是本章节切片的前置（D-171）。
   - `plugins/welcome/` 在 member.joined 时向该聊天发送欢迎语，欢迎语中的 `{name}` 替换为新成员的显示名。欢迎语按聊天配置，缺省使用插件文案；`/welcome on|off|set <文案>` 的权限为 `admin` 或 `chat_admin`，默认关闭。
   - `plugins/rss/` 提供 `/rss add|list|remove`，订阅经 `sqlitedb` 保存在插件数据库，不链接 GORM；add 与 remove 的权限为 `admin` 或 `chat_admin`，list 不限。一个固定 interval 的 schedule 检查订阅，间隔是配置项（缺省 10 分钟）。用标准库解析 RSS 2.0 与 Atom，按 guid 或链接去重，只推送新条目；后台推送需要 `allowed_adapter_instances`。
   - rss 抓取只允许 http 与 https；解析出的地址以及每次重定向后的地址都不能是 loopback、私网、链路本地、组播或未指定地址。响应大小上限 5 MiB、抓取超时 30s，均为命名常量，超过时本次抓取失败并记 WARN。
3. **发行归档。** `scripts/dist/` 为每个目标平台构建 Core 与首个发行版的组件：官方 Telegram Adapter、官方 Agent、官方 WebUI（含构建 `web/admin/` 得到的 WebUI 包）与 help（D-171），用 `alonabot pack` 生成安装包与外部 `release.json`，签名后放入 `packages/`，再打成 zip（Windows）或 tar.gz，并生成列出全部发行归档 SHA256 的 `SHA256SUMS`。发行版 Core 在构建时嵌入官方发布者公钥。
4. **首次运行。** 数据目录首次初始化时，按普通安装流程安装 `packages/` 中的包，只启用 help 与官方 WebUI，Agent 只安装不启用（D-50、D-62）；Telegram Adapter 在管理员为它添加第一个实例时启用。交互终端下按 CONTRACTS §15.1 初始化管理 token。之后不再自动安装；新版本发行包中版本更高的内置包作为可用更新列出。
5. **容器镜像。** 仓库根目录的 `Dockerfile` 用多阶段构建：在 Go 构建阶段以 `CGO_ENABLED=0` 构建 Core 与全部官方包，最终镜像只含 `alonabot`、`packages/` 与 CA 证书，以 UID 10001 运行，`/data` 为卷，设置 `ALONABOT_CONTAINER=1`。Core 在该环境变量下才允许 `0.0.0.0` 的 listen 搭配本机 http 的 `base_url`，并在启动时输出 WARN 提醒端口只发布到宿主本机。`deploy/compose.yaml` 是官方整合包的 compose 入口，随 Release 发布，固定对应版本的镜像，给出卷、`init: true`、`127.0.0.1:7650:7650`，以及首次启动后用 `docker compose exec alonabot alonabot admin-token show` 读出一次性 token 的用法（CONTRACTS §15.3）。镜像构建 linux/amd64 与 linux/arm64 两个平台。
6. **端到端验收。** 在干净的数据目录上，分别从发行归档与容器镜像依次完成：首次运行、登录 WebUI、添加 Telegram 实例（本地 Bot API 模拟服务）、启用 Agent 并配置 Provider（本地模拟服务）、`/ask` 对话、`/help` 与 `/whoami`（D-171 去掉 rss 一步）。真实 Telegram 与真实 Provider 没有凭证时记为未运行。
7. **install.sh。** `deploy/install.sh` 按 CONTRACTS §15.6 实现：只依赖 POSIX sh、curl、tar 与 sha256sum/shasum；不支持的系统与没有 systemd 的 Linux 在改动前退出；按 `SHA256SUMS` 校验后才停止旧服务、替换程序；尚无管理凭证时经 `/dev/tty` 运行 `admin-token init`，没有终端时跳过初始化，启动服务后输出 `admin-token show` 的命令；服务运行 `alonabot run --data-dir <数据目录>`；写入 systemd 单元（系统用户 `alonabot`）或用户 LaunchAgent 并启动；再次运行即升级，`--uninstall` 保留数据目录。Release 流程把它与 `SHA256SUMS` 一起上传。用 `shellcheck` 检查（未安装时记为未运行）；在真实的 systemd Linux 与 macOS 上各完成一次安装、升级与卸载，模拟环境与容器内的 systemd 不能代替。

#### 拟目录与交付

| 拟目录 | 本工作包交付 |
| --- | --- |
| `plugins/help/` | 首个发行版的预装插件（welcome 与 rss 见插件实施计划，D-171） |
| `internal/core/router/` | `core.command.list` |
| `scripts/dist/` | 发行归档的构建、打包与签名步骤 |
| `Dockerfile`、`deploy/compose.yaml` | 官方容器镜像与 compose 整合包入口 |
| `deploy/install.sh` | Linux 与 macOS 的安装脚本 |
| `internal/core/management/` | 首次初始化的预装与默认启用、内置包的可用更新 |

#### 验收与后续边界

| 验收 | 本工作包必须提供的证据 |
| --- | --- |
| V-66 | 首次运行的预装与默认启用、token 初始化、签名错误的包、卸载后不重装、内置包更新只列出（OneBot 11 的分项随它进入发行版时验收，D-171） |
| V-67 | help 的权限过滤与翻译、`/whoami`；welcome 与 rss 的分项随它们进入发行版时验收（D-171） |
| V-70 | 两个平台的镜像、首次运行与 token 初始化、本机端口发布、非 root、HTTPS 出站、停止容器时插件进程全部退出 |
| V-84 | install.sh 在真实 systemd Linux 与 macOS 上的安装、升级、卸载、校验失败与无终端的分支 |

五个目标平台的发行归档与两个平台的容器镜像运行结果由 T-10 汇总。

#### 失败停止条件

- 预装包绕过签名或摘要校验、首次运行之外自动安装或更新包，阻止合入。
- install.sh 在校验失败后改动已有安装，或升级、卸载时删除数据目录，阻止合入。
- rss 能访问 loopback、私网或链路本地地址，阻止 rss 合入（rss 进入发行版时适用）。
- 签名私钥或管理 token 出现在仓库、日志或发行归档中，立即停止并处理。
