# 核对候选发行资产与安装更新链

`release-check` 读取 `scripts/dist` 生成的候选目录，在新目录保留解包内容和 `result.json`。默认检查纯 Core 归档；`--bundle` 检查内部整合归档及签名组件。默认要求全部五目标；候选子集必须用 `--platforms` 明确指定。各平台二进制只进行静态读取。

先在仓库中构建检查器，再从任意工作目录运行。检查期间无需源码树、Go 工具链或网络：

```sh
go build -o /tmp/alonabot-release-check ./scripts/release-check
/tmp/alonabot-release-check --candidate /tmp/alonabot-candidate \
  --version 0.1.0 --public-key /tmp/trusted-release.pub \
  --out /tmp/alonabot-release-evidence
```

`--public-key` 使用独立保存、由调用者信任的公钥；正式发行公钥见 [`.github/release.pub`](../../.github/release.pub)。

| 参数 | 含义 |
| --- | --- |
| `--candidate` | 必填；`dist` 输出目录，含发行归档、`SHA256SUMS` 和 `install.sh`；整合模式额外要求 `packages/`。 |
| `--bundle` | 额外核对 Agent、WebUI 的签名组件及其整合副本。 |
| `--version` | 必填；外层发行归档和 Core 的构建版本。组件版本分别读取各自签名清单。 |
| `--public-key` | 必填；标准 base64 编码的 Ed25519 公钥文件。资产旁的公钥须与此输入一致。 |
| `--out` | 必填；尚不存在的证据目录，父目录须已存在，位置须在候选目录之外。已有目录原样保留。 |
| `--platforms` | 逗号分隔的目标，缺省 `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`。须与发行归档的目标集合一致；组件保留完整签名清单，其中须包含这些目标。 |
| `--asset-base` | 期望的资产 URL 前缀；缺省使用各组件仓库 `releases/download/v<组件版本>`；用于核对历史候选的自定义下载前缀。 |
| `--install-candidate` | 可选运行链的初始组件包目录：从现有 configdb 示例打包，使用独立第三方测试发布者。 |
| `--update-candidate` | 同 ID、来源、发布者，组件自身 SemVer 严格递增的包目录。 |
| `--wrong-source-candidate` | 同更新版本、ID、发布者，来源不同的有效签名包目录。 |
| `--wrong-publisher-candidate` | 同更新版本、ID、来源，发布者不同的有效签名包目录。 |
| `--runtime-timeout` | 运行链必填，例如 `2m`；检查器整条联验的截止时间，包含候选退出。 |
| `--help` | 显示实际参数并以 0 退出。 |

整合模式先执行下列 1～4 项；两种模式均执行第 5 项：

1. 对 WebUI 与 Agent 两个整合组件的 `release.json` 确切字节验签，核对可信发布者、组件 ID 和绑定来源。来源分别为 `github.com/AlonaBot/alonabot-webui` 与 `github.com/AlonaBot/alonabot-agent`。
2. 核对每个目标包的完整 URL、归档长度与 SHA256，安全解包后读取入口的实际 Go build-info，要求包内 `build-info` 与其逐字一致，平台相符且 CGO 关闭。包目录只包含各目标包与三份发布元数据。
3. 按 `pack` 的稳定输出核对包内 manifest 的 ID、版本、来源、角色、平台、入口和格式版本记录；检查二进制中作者 manifest 的共有身份记录。二进制内嵌作者 manifest，包内 manifest 增加了构建字段。此处针对正式 `dist/pack` 格式进行静态核对，完整 YAML、Schema 语义和运行时资源登记由候选 CLI 的安装、启用与握手验收。
4. 对包内存在的 `config_schema.json`、`adapter_config_schema.json` 和直属 `i18n/*.json` 检查 JSON 格式及二进制中的确切资源字节；在 JSON 中列出实际检查的资源。Schema 使用 `pack` 已规范化的原文。
5. 逐一检查 `SHA256SUMS` 中的发行归档，核对 Core 的实际目标、CGO 设置、已链接公钥与版本；核对 `source.json` 的来源和版本。纯 Core 归档仅含可执行文件、`LICENSE`、`source.json`，且来源记录没有组件；整合模式的包及 `release.json/.sig/.pub` 必须与外部文件逐字一致。两种模式的归档文件清单均须与预期相同。

解包只接受普通文件和目录，拒绝路径穿越、链接、重复条目、大小写冲突和文件/目录冲突。Linux/macOS 入口须带执行位。`install.sh` 须为非空普通文件；其功能、固定版本 compose 与镜像由发布候选联验覆盖。

成功返回 0，失败返回 1。stdout 始终输出结果 JSON（`--help` 除外），完整错误写入 `error`，已完成的检查与失败检查保留在 `checks`。证据目录创建成功后，同一 JSON 写入 `result.json`，失败解包现场保留。参数错误或证据目录创建失败时，可从 stdout 保存结果。stderr 保存参数帮助或失败日志。

仅给出原有静态参数时，`scope` 为 `asset-integrity`，`runtime_checks` 为 `not_run`。五目标静态检查结果分别记录所选平台。

要验证当前宿主的安装与更新，另给出上表四个组件目录与运行截止时间。每个目录包含正式 `alonabot pack` 生成的该平台 tar.gz、`release.json`、`release.json.sig` 和 `release.pub`；组件固定为现有 `examples/plugins/configdb` 示例，manifest 显式声明 `role: plugin`。初始与更新副本只改作者 manifest 中的版本/来源，错误来源副本只换来源，错误发布者副本只换测试签名密钥。私钥留在构建环境，检查器只读取发行文件。

```sh
/tmp/alonabot-release-check --candidate /tmp/alonabot-candidate \
  --version 0.1.0 --public-key /tmp/trusted-release.pub \
  --out /tmp/alonabot-lifecycle-evidence \
  --install-candidate /tmp/configdb-0.1.0 \
  --update-candidate /tmp/configdb-0.1.1 \
  --wrong-source-candidate /tmp/configdb-wrong-source \
  --wrong-publisher-candidate /tmp/configdb-wrong-publisher \
  --runtime-timeout 2m
```

`--platforms` 必须包含当前宿主。程序先完成全部静态检查，再在系统临时目录重新解包宿主归档，以独立空 cwd 和 data_dir 无 TTY 启动，不调用源码或构建工具。它通过 `admin-token show` 获取首次 token：Unix 核对文件权限 0600，Windows 核对从受保护 data_dir 继承的实际 DACL 只授予运行账号访问且允许读取。token 只留在内存和候选自身的私密文件中。管理 CLI 使用 Bearer token 文件；本机转发将随机监听端口的 Host/Origin 对齐候选实际 `base_url`。

运行链核对 HTTP 返回的 Core 版本与预装清单：纯 Core 为零个组件，整合模式核对官方两包的版本、来源、发布者和默认启用状态。随后安装并启用第三方 configdb，通过公开数据管理 API 写入 `configdb_note`，读取真实 SQLite 行与插件 OnStart 写入的 `configdb_startup`。错误签名、有效签名但来源不同、发布者不同与 pin 不匹配的更新均须被管理 CLI 拒绝；每次拒绝后核对身份、版本、Active 与数据库原内容。解除 pin 后执行真实高版本更新，要求 Active 变化、内容保留；普通卸载保留绑定与代次，拒绝异源/异发布者重装，再同源重装、启用，要求 Active 等于卸载前保留代次并读取原内容。

运行成功时 `scope=asset-integrity-and-lifecycle`、`runtime_checks=pass`。`runtime` 记录平台、保留的系统临时工作目录、Core PID/退出码和各阶段组件状态。`out/runtime/` 保留 Core 日志、管理 CLI 原始 stdout/stderr/参数/退出码、数据 API 原始响应；预期拒绝的原错同样保留。失败返回完整错误及已完成阶段；候选收到中断信号并等待退出，超过调用者给定截止时间则终止并记为失败。临时资产和 data_dir 保留供复查，整个程序成功也要求 Core 正常退出且日志未出现 token。真实平台运行结果由发布矩阵记录。
