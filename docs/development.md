# 开发与验证

负责范围：环境准备、日常命令、外部测试环境、规则与检查的对照。
更新时机：工具链或依赖版本变化、新增或调整检查入口、测试环境变化时。

## 环境

| 工具 | 版本与来源 | 备注 |
| --- | --- | --- |
| Go | 1.27.1（asdf，`~/.tool-versions`） | `go.mod` 的 go 指令固定为该版本 |
| protoc | 36.0（`protoc --version` 输出 `libprotoc 36.0`；生成文件头记为 v7.36.0） | 版本不符时 `contracts/internal/protocgen` 拒绝生成 |
| protoc-gen-go | v1.36.12（`go.mod` 的 `tool` 指令，与 `google.golang.org/protobuf` 同版本） | 通过 `go tool -n protoc-gen-go` 定位，不全局安装 |
| Node（前端） | 26.5.1（asdf） | 供 `web/admin` 使用，见[官方 WebUI 开发](#官方-webui-开发) |

已锁定并经测试验证的版本：protobuf v1.36.12、go.yaml.in/yaml/v3 v3.0.5、gorm v1.31.2、libtnb/sqlite v1.2.2、modernc.org/sqlite v1.61.0（已验证 CRUD/Tx/WAL Backup/迁移账本）；golang.org/x/sys v0.49.0 为直接依赖，用于 internal/oslock 的 flock/LockFileEx，以及 internal/core/process 在 Darwin 上查询进程组成员；Cobra v1.10.2、pflag v1.0.9、仅 Windows 链接的 mousetrap v1.1.0。

已批准的 Go 依赖（2026-10-21）：`google.golang.org/protobuf`、`gorm.io/gorm`、`github.com/libtnb/sqlite`、`modernc.org/sqlite`、`golang.org/x/sys`、`go.yaml.in/yaml/v3`（`gorm.io/driver/postgres` 的批准于 2026-10-27 随 D-48 撤销）；2026-10-25 批准 `github.com/modelcontextprotocol/go-sdk`，只用于官方 Agent 组件（`agent/`）的 MCP 客户端，T-12.3 验证后锁定 v1.3.1；2026-10-26 批准 `github.com/spf13/cobra`（随之引入 `github.com/spf13/pflag`），只用于 `cmd/alonabot` 的命令树与全局参数，T-02.2 已锁定 Cobra v1.10.2、pflag v1.0.9 与仅 Windows 链接的 mousetrap v1.1.0；2026-10-26 批准 `github.com/google/jsonschema-go`（MCP go-sdk 也直接依赖它），用于 JSON Schema 的表示、序列化、`Resolve`、`ApplyDefaults` 与 `Validate`：配置 Schema（`sdk/config`、构建期生成器、Core 的配置与表单 Schema）、Service 与工具的输入 Schema、Adapter 声明的扩展事件、extension 与平台动作的 Schema（`sdk/adapter` 用 `For` 从 Go 类型推导，D-113）、`agent/`。配置 Schema 由生成器构造，不用 `For`/`ForType` 推导（D-43）；版本在首个使用它的切片引入时锁定。2026-10-27 批准 `github.com/coder/websocket`，只用于官方 OneBot 11 Adapter（`adapters/onebot11/`）的 WebSocket 连接，T-15.1 已锁定 v1.8.15（D-46）。2026-11-10 批准 `github.com/openai/openai-go/v3` 与 `github.com/anthropics/anthropic-sdk-go/v2`，只用于官方 Agent 组件的 Provider 客户端（`agent/provider`），不导入两者的 Azure、Bedrock、Vertex 等云厂商子包；T-12.12 锁定 openai-go v3.98.0、anthropic-sdk-go v2.9.0。实际链接的传递依赖为 bahlo/generic-list-go v0.2.0、buger/jsonparser v1.1.2、coder/websocket v1.8.15、go-logr/logr v1.4.2、go-logr/stdr v1.2.2、invopop/jsonschema v0.14.0、pb33f/ordered-map/v2 v2.3.1、standard-webhooks/standard-webhooks/libraries v0.0.1、tidwall/gjson v1.20.0、tidwall/match v1.1.1、tidwall/pretty v1.2.1、tidwall/sjson v1.2.5、go.opentelemetry.io/auto/sdk v1.1.0、go.opentelemetry.io/otel（含 metric、trace）v1.33.0、go.yaml.in/yaml/v4 v4.0.0-rc.2、golang.org/x/sync v0.24.0。MVS 将 x/text 从 v0.39.0 抬至 v0.42.0；coder/websocket 保持 v1.8.15；MCP go-sdk 的模块下限为 v1.3.1，当前 Provider 未链接 MCP 包。前端已批准：Vite（含 `@vitejs/plugin-react`）、TypeScript、React 19、vitest；2026-11-05 批准开发依赖 `@types/react@19`、`@types/react-dom@19`（PR #91），只用于类型检查。T-08.7 锁定 react 与 react-dom 19.3.1、@types/react 与 @types/react-dom 19.3.0、@vitejs/plugin-react 6.2.1、TypeScript 7.0.2、Vite 8.3.5、vitest 5.1.0；DOM 测试环境（jsdom、happy-dom）与 Testing Library 未批准。2026-11-17 批准（D-169）：`lucide-react`、`@dnd-kit/core`、`@dnd-kit/sortable`、`@radix-ui/react-dialog`、`@radix-ui/react-dropdown-menu`、`@radix-ui/react-popover`、`@radix-ui/react-tooltip` 与 `recharts`，版本在首个使用它的切片引入时锁定；JetBrains Mono 的 woff2（SIL OFL 1.1）作为静态资源随 WebUI 包提供，不是 npm 依赖。新增清单外依赖需要先征得维护者同意。

Core 可执行文件（`./cmd/alonabot`）与 SDK 根包（`./sdk`）只能链接上述清单的一个子集（D-52，D-68 修订），由审查时用下方的 `go list -deps` 命令核对：Core 允许 protobuf、modernc SQLite 及其依赖、GORM 与 libtnb/sqlite 及其依赖（jinzhu/inflection、jinzhu/now、golang.org/x/text）、yaml、x/sys、cobra 与 pflag、仅 Windows 链接的 mousetrap、jsonschema-go；SDK 根包只允许 protobuf、x/sys 与 jsonschema-go。MCP go-sdk、coder/websocket 只出现在官方插件、Agent 或 Adapter 中，openai-go 与 anthropic-sdk-go 只出现在 Agent 中；插件只在导入 `sdk/gormdb` 时链接 GORM。修改子集与修改已批准清单一样，需要维护者同意。

## 日常命令（仓库根目录）

```sh
go generate ./...   # 重新生成协议代码和 Core changelog
go test ./...
go build ./...
go vet ./...
go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' ./cmd/alonabot ./sdk | sort -u   # Core 与 SDK 根包链接的模块，对照允许清单（D-52）
```

协议生成器（入口：`contracts/doc.go` → `contracts/internal/protocgen`）会先删除 `contracts/pb/` 下的 `*.pb.go` 再重新生成；新增 `.proto` 放在 `contracts/proto/alonabot/v1/`，`go_package` 统一为 `github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1`。

更新仓库根目录的 [`CHANGELOG.md`](../CHANGELOG.md) 后，运行 `go generate ./internal/core/manage`，将生成的 `internal/core/manage/changelog_core.go` 一同提交。

修改插件页面套件后，用 `go test -race -v ./sdk/page ./examples/plugins/showcase` 检查受影响的 Go 包，并用 `node --check sdk/page/kit.js` 检查脚本语法。资源由 `sdk/page` 直接嵌入 Go 产物，无需 npm 构建。主题、语言和快捷键的端到端验收在 WebUI 的沙箱 iframe 中用 [showcase](../examples/plugins/showcase/main.go) 联验（V-130）。

交叉编译矩阵（只证明能构建，不能代替目标平台的运行证据）：

```sh
for p in linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64; do
  GOOS=${p%/*} GOARCH=${p#*/} CGO_ENABLED=0 go build ./... && echo "OK $p" || echo "FAIL $p"
done
GOARCH=amd64 CGO_ENABLED=0 go test ./...   # Apple Silicon 上经 Rosetta 运行 darwin/amd64 测试
```

## 官方 WebUI 开发

在仓库根目录安装锁定的依赖并启动 Vite：

```sh
npm --prefix web/admin ci
npm --prefix web/admin run dev
```

打开 Vite 输出的地址，首页跳转到 `/dev/forms`。预览使用 `web/admin/fixtures/`，可切换语言并查看表单提交生成的 JSON Merge Patch，无需启动 Core。接入新页面时按 [WebUI 实施计划 §3](../spec/plan/webui.md#3-页面接入外壳) 登记路由和文案。

本地运行 `npm --prefix web/admin run typecheck`、`npm --prefix web/admin run test` 和 `npm --prefix web/admin run build`。测试使用被忽略的 `*.local.test.ts`，无本地测试时命令也通过；CI 不运行前端检查。生产静态文件输出到 `web/admin/dist/`，其中包括站点图标、标志、JetBrains Mono woff2 和许可。

T-09.26 锁定 `lucide-react` 1.74.0、`@dnd-kit/core` 6.3.1、`@dnd-kit/sortable` 10.0.0、`@radix-ui/react-dialog` 1.3.1、`@radix-ui/react-dropdown-menu` 2.2.1 和 `@radix-ui/react-tooltip` 1.4.1；`@radix-ui/react-popover` 与 `recharts` 由首个使用它们的页面切片引入。

## 发行构建

在仓库根目录生成包含 Core、Telegram Adapter、Agent、WebUI 与 help 的发行候选：

```fish
go run ./scripts/dist --version 0.1.0 \
  --key ~/.config/alonabot/release.key --out ./dist
```

构建机需要上表中的 Go、Node 与 npm，并可读取已锁定的 Go 模块和 npm 包。程序先运行 `web/admin` 的 `npm ci` 与 `npm run build`，再由主机版 `alonabot pack <组件目录> --target <os>/<arch> ... --asset-base <前缀> --out <目录>` 为每个组件调用一次，传入全部目标；pack 写出合并清单，dist 再签名并自检。组件版本沿用各自 `plugin.yaml`。默认目标为 `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`，可用 `--platforms linux/amd64,linux/arm64` 选择子集。离线 npm 缓存已齐全时，可在命令前加 `env npm_config_offline=true`。

`--key` 文件保存标准 base64 编码的 64 字节 Ed25519 私钥，允许末尾空白。私钥放在仓库及组件资源目录之外，设置为仅发布维护者可读（0600），通过受控环境或离线介质保管和备份。程序只从该文件读取私钥，子进程不接收私钥；发行物只携带公钥与签名。发布前先用临时测试密钥验证流程，正式签名使用维护者控制的密钥。

`dist/` 中每个平台生成一个 `alonabot-<version>-<os>-<arch>.tar.gz`（Windows 为 `.zip`），解压后的同名目录包含 `alonabot`（Windows 为 `alonabot.exe`）和 `packages/<组件ID>/`。每个组件目录包含该平台安装包、合并全部目标平台的 `release.json`、`release.json.sig` 与 `release.pub`；当前静态 WebUI 只生成一份 `any` 安装包。另在 `dist/packages/<组件ID>/` 汇集各平台安装包与同一份签名清单，供后续发布使用。`dist/SHA256SUMS` 列出所有外层发行归档的摘要，格式为摘要、两个空格和文件名。

安装包下载前缀默认取 `https://github.com/AlonaBot/AlonaBot/releases/download/v<version>`，可用 `--asset-base https://example.com/releases/v0.1.0` 覆盖。发行 Core 通过 `-ldflags -X` 嵌入官方公钥，安装官方签名包时免首次来源与发布者确认；签名与摘要仍按安装流程验证，已有发布者绑定继续生效。未注入公钥的开发构建保留原确认流程。

当前 `--version` 决定发行归档名和默认下载前缀；Core 管理 API 的版本仍来自 Go module 构建信息，源码构建会显示 `(devel)`。正式发布前需在 `internal/app/version.go` 补齐 Core 版本注入入口。

## 容器镜像

从仓库根目录启动对应版本的官方整合包，然后读出首次登录的 token：

```fish
docker compose -f deploy/compose.yaml up -d
docker compose -f deploy/compose.yaml exec alonabot alonabot admin-token show --data-dir /data
```

打开 `http://127.0.0.1:7650`，使用读出的 token 登录。首次启动无需 TTY，Core 将一次性 token 写入 `/data/admin_token`，首次登录成功后删除该文件。镜像支持 `linux/amd64` 与 `linux/arm64`，compose 按宿主平台选择。compose 中固定镜像版本；T-10.4 随 Release 上传该文件的副本时，须与 Release 的镜像版本一致。

镜像以 `10001:10001` 运行，命名卷挂载到 `/data`，首次创建卷时继承镜像中该目录的所有者。改用宿主目录挂载时，先授予 UID/GID 10001 读写和遍历权限；已有卷也需保证这一权限。保留该卷即可保留配置、状态库和组件数据。

自行构建时，使用上表锁定的 Go、Node/npm 与 Docker BuildKit/buildx。`VERSION` 必填，签名私钥使用发行构建相同的 base64 Ed25519 格式，存放在构建上下文之外，权限为 0600；构建验收使用临时密钥，正式发行使用维护者控制的密钥。密钥通过必需的 BuildKit secret 挂载传入。构建阶段直接调用 `scripts/dist`，沿用其中的 npm ci/build、签名和四个官方包清单。

```fish
set version 0.1.0
set release_key ~/.config/alonabot/release.key
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  --output type=oci,dest=/tmp/alonabot-$version.oci.tar \
  -t ghcr.io/alonabot/alonabot:$version .
```

需要覆盖组件下载前缀时，加 `--build-arg ASSET_BASE=https://example.com/releases/v0.1.0`；省略时沿用 `scripts/dist` 的 GitHub Release 地址。下面分别构建并在可运行相应平台的主机上启动本地镜像；跨架构运行需要 Docker 已配置的模拟器。

```fish
docker buildx build --platform linux/amd64 --load \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  -t alonabot:$version-amd64 .
docker run -d --name alonabot-amd64 --platform linux/amd64 --init \
  -p 127.0.0.1:7650:7650 -v alonabot-amd64-data:/data alonabot:$version-amd64
docker exec alonabot-amd64 alonabot admin-token show --data-dir /data
# 完成该平台验收后停止，释放宿主端口，再运行另一个平台。
docker stop alonabot-amd64

docker buildx build --platform linux/arm64 --load \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  -t alonabot:$version-arm64 .
docker run -d --name alonabot-arm64 --platform linux/arm64 --init \
  -p 127.0.0.1:7650:7650 -v alonabot-arm64-data:/data alonabot:$version-arm64
docker exec alonabot-arm64 alonabot admin-token show --data-dir /data
```

端口只发布到宿主的 `127.0.0.1`。镜像设置 `ALONABOT_CONTAINER=1` 时，缺省监听为 `0.0.0.0:7650`，`base_url` 仍为 `http://127.0.0.1:7650`；显式 `listen` 优先。该容器例外实际生效时，启动日志输出一次 WARN 提醒宿主本机发布规则。远程管理可使用 `ssh -N -L 7650:127.0.0.1:7650 user@host`，再打开本机地址；HTTPS 反代部署按 [Core 设置与 HTTP 入口](#core-设置与-http-入口) 配置 `base_url` 和 `trusted_proxies`。

在 compose 服务的 `environment` 中传入 `TELEGRAM_BOT_TOKEN` 后，`/data/core_config.json` 中的实例凭证可写成 `"token": "${TELEGRAM_BOT_TOKEN}"`，设置写回保留这一引用。最终镜像采用 scratch，包含 Core、相邻 `packages/` 与 CA 证书；MCP stdio 服务器需要 shell、Node、Python 等运行时时，请自行构建包含所需运行时的镜像，并保留 UID 10001、可写数据卷与 `--init`。

## 用 install.sh 安装与升级

在支持 systemd 的 Linux 上以 root 运行；macOS 在当前用户的登录会话运行。支持两种系统的 amd64、arm64。先从同一 Release 下载脚本，再安装指定版本（以下命令使用 fish，把 `VERSION` 换成发行版本号，省略前导 `v`）：

set version VERSION
curl --fail --location --proto '=https' --proto-redir '=https' \
  -o install.sh https://github.com/AlonaBot/AlonaBot/releases/download/v$version/install.sh
# Linux
sudo sh install.sh --version $version
# macOS
sh install.sh --version $version

省略 `--version` 会先通过 `/latest` 固定一个 Release，再下载该版本的归档与 `SHA256SUMS`。`ALONABOT_RELEASE_BASE_URL` 可覆盖 Release 基址，基址和所有重定向均须使用 HTTPS。脚本先检查平台、权限和系统工具，完成下载、校验和临时解压后才停服并替换 `alonabot` 与完整的 `packages/`。

Linux 使用 `/opt/alonabot/`、数据目录 `/var/lib/alonabot` 和系统服务 `alonabot.service`；服务以系统用户 `alonabot` 运行并启用开机启动。macOS 使用 `~/Library/Application Support/AlonaBot/`、其中的 `data/` 和用户 LaunchAgent `io.github.alonabot`，登录后启动。工作目录设为程序目录，启动参数为 `run --data-dir <完整数据路径>`。

有控制终端时，脚本在停服后以服务身份运行 `admin-token init --if-missing`，stdin 与 stderr 连接 `/dev/tty`；新 token 仅显示一次，已有凭证保持原值。经管道执行同样使用控制终端。无控制终端时跳过初始化，服务照常启动，Core 首次启动写入 `admin_token`。文件存在时按脚本输出的完整命令领取：

# Linux，以服务身份读取
sudo runuser -u alonabot -- /opt/alonabot/alonabot admin-token show --data-dir /var/lib/alonabot
# macOS，以当前用户读取
"$HOME/Library/Application Support/AlonaBot/alonabot" admin-token show --data-dir "$HOME/Library/Application Support/AlonaBot/data"

首次登录成功后 Core 删除 `admin_token`。再次运行安装命令即升级，保留数据、`core_config.json`、状态库、已安装组件与未领取的 token。内置包的预装、默认启用与可用更新由 Core 处理。

# Linux 卸载
sudo sh install.sh --uninstall
# macOS 卸载
sh install.sh --uninstall

卸载停止并解除服务注册，删除服务定义、程序与内置包，输出保留的数据目录位置；macOS 的 `data/` 保留在原位。

本地先运行 `sh -n deploy/install.sh`、`shellcheck deploy/install.sh`（缺工具记录「未运行」）和 `go test -race ./cmd/alonabot ./scripts/dist`。临时 HTTPS Release 和系统命令模拟放在 `*_local_test.go`，分别验证版本/latest、下载校验失败门禁、服务参数与数据保留。开发机只在临时目录模拟；V-84 的真实安装须使用专用、可恢复的 systemd Linux（root）与 macOS（当前用户）环境，逐项记录全新安装、管道有/无终端、升级、校验失败与卸载，核对身份、token、旧安装及数据；缺少环境的项目记「未运行」。容器内 systemd 与命令模拟的结果分别记录。

`go run ./scripts/dist --version VERSION --key FILE --out DIR` 会在输出目录附带与 `deploy/install.sh` 内容相同的 `install.sh`，`SHA256SUMS` 继续仅列外层发行归档。当前上传由 T-10.4 的维护者执行：将 `install.sh`、`SHA256SUMS` 和发行归档上传至同一个 Release，并按发行步骤上传各组件安装包及 `release.json`、`.sig`、`release.pub`。

## 运行 Core 与管理组件

```sh
go build -o bin/alonabot ./cmd/alonabot
bin/alonabot run --data-dir ./data      # 不带子命令时等同 run；按安装记录启动全部已启用的组件
# Ctrl-C / SIGTERM：排空 → STOPPING/OnStop → 退出；同一 data_dir 只能运行一个 Core
```

`run` 与库模式（`alonabot.Start`）共用 `internal/app` 的同一套 Core 装配；CLI 直接调用它，根包提供库模式的薄包装。组件经管理服务安装与启停，CLI 通过 `base_url` 上的管理 API 操作，token 取自 `ALONABOT_ADMIN_TOKEN` 或 `--token-file`，管理地址可用 `--admin-url` 覆盖：

```sh
bin/alonabot install pkg.tar.gz --yes             # release.json、签名和公钥缺省取归档同目录
bin/alonabot install alonabot-webui-admin --yes    # 从可执行文件旁 packages/<组件 ID>/ 安装
bin/alonabot install --yes                        # 在当前目录或工程 bin/ 中选择唯一的本机包
bin/alonabot install --source owner/repo --yes      # 从 GitHub Release 下载 release.json、可选的 .sig 与 release.pub
bin/alonabot enable <id>
bin/alonabot list
bin/alonabot uninstall <id> --purge-data --yes      # 或 --purge；不带选项时只卸载
bin/alonabot adapter instance create <adapter-package-id> <instance-id> --config instance.json
bin/alonabot admin add telegram:123456
```

`install` 的位置参数为已存在的普通文件时使用该归档，其他值按组件 ID 查找可执行文件旁的 `packages/<组件 ID>/`；缺失时列出可用的官方组件 ID。不带位置参数时，当前目录为作者工程便搜索 `bin/`，其他情况搜索当前目录；也检查直接包子目录的清单。按 `release.json` 选择本机平台资产，静态 WebUI 使用 `any/any`，恰好一个候选才安装；零个或多个时列出搜索目录和候选后退出。发现阶段只读，安装经管理 API 完成。

`--archive` 仍可显式指定归档；`--release`、`--signature` 与 `--public-key` 分别覆盖元数据文件。签名和公钥未指定时取 release.json 同目录的 `release.json.sig` 与 `release.pub`。位置参数、`--archive` 和 `--source` 互斥；`--source` 与本地元数据选项也互斥。来源与签名确认、归档预检及 `--yes`、`--replace` 沿用管理服务。

help 插件启用后，在聊天中发送 `/whoami`，把回复中的用户字符串作为 `admin add` 的参数；回复中的实例名用于核对来源。

排查 Core 设置时，读取 `GET /api/v1/core/settings/corrections`：`corrections` 列出转换、恢复默认值与移除记录，`unknown` 列出声明之外的键。Core 载入时校正运行副本，原文件中的值与 `${环境变量}` 写法保持；每条校正在日志中只记路径与原因。`admins` 值有误时恢复为空列表，修订后重新读取设置。

Adapter 实例的记录在 `GET /api/v1/adapters/{adapter_id}/instances/{instance_id}` 的 `corrections` 与 `unknown` 中。校正后 required 仍缺失的实例留待修订，其余实例独立准备。清理 Core 的未知键调用 `POST /api/v1/core/settings/prune`，清理某个实例调用 `POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune`；这两个入口经原有管理认证和文件写回流程，只删除声明之外的键，保留其他原值。

需要确认的操作不带 `--yes` 时列出受影响的对象后退出。开发中的组件可以按下节以目录运行，也可以用库模式，先 `alonabot build ./examples/devbot`，再运行 `examples/devbot/bin/devbot --data-dir ./data`。

使用同一套管理凭证运行 `bin/alonabot profile list` 查看配置文件 ID 和会话数。`bin/alonabot profile assign <配置文件ID> <实例名>` 设置实例缺省；再给出会话字符串可单独指定会话。`profile set` 与 `profile unset` 编辑 `default` 时写入插件的 Pending，沿用配置重载流程。全部子命令见 [CONTRACTS §5.6](../spec/CONTRACTS-v0.1.md#56-配置文件)，平台设置与工具策略见 [ADMIN-API §23](../spec/ADMIN-API.md#23-配置文件)。

### 导出诊断包

```fish
alonabot diagnose --data-dir ./data --since 24h -o diagnose.zip
```

默认收集近 24 小时的日志、当前设置和 Active 配置、组件与环境信息，并开启脱敏。`--no-logs`、`--no-settings`、`--no-components` 可分别省略一组，至少保留一组。`-o` 也可写作 `--output`；省略时在当前目录生成 `alonabot-diagnostics-<UTC 时间>.zip`。

在线使用管理 API 和 Bearer token：优先读 `--token-file`，其次取 `ALONABOT_ADMIN_TOKEN`，两者均未指定时尝试 `<data_dir>/admin_token`。`--admin-url` 可覆盖管理地址。连接被拒绝后尝试停机导出，无需 token；数据目录须已有可读的 `core.lock` 且锁未被占用。离线导出只读原目录，组件状态为 `Stopped`。

缺少可用的 Adapter 实例 schema 时，默认遮盖该组实例配置的全部叶值，保留结构和完整的 `${环境变量}` 表达式。`--no-redact` 会导出凭证明文；日志可能包含聊天内容，提交 issue 前请先检查归档。

## 开发运行与热部署

先从仓库示例完成一次可见回复（以下使用 fish）：

```fish
go build -o bin/alonabot ./cmd/alonabot
set -gx DEVBOT_DATA (mktemp -d)
set -gx DEVBOT_OUTPUT (mktemp)
cp examples/devbot/core_config.example.json $DEVBOT_DATA/core_config.json
printf 'DEVBOT_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_DATA $DEVBOT_OUTPUT
bin/alonabot run examples/devbot/plugin examples/devbot/adapter --data-dir $DEVBOT_DATA --reload
```

另开终端，设置相同的 `DEVBOT_DATA`、`DEVBOT_OUTPUT`。终端首次启动的 token 在 stderr 显示一次；非终端启动可用 `bin/alonabot admin-token show --data-dir $DEVBOT_DATA` 读取 `admin_token`。令 `token` 为该值后：

```fish
curl --fail-with-body -X POST -H "Authorization: Bearer $token" \
  'http://127.0.0.1:7650/plugins/alonabot-adapter-devbot/inject?instance=one&text=%2Fhello'
cat $DEVBOT_OUTPUT
```

注入返回 204，异步发送后文件内容为 `Hello from devbot!`。完整实例 JSON、目录结构、认证、源码／翻译／配置改动与编译失败恢复操作见 [devbot README](../examples/devbot/README.md)。样例把 `adapters.devloop.package` 绑定到 `alonabot-adapter-devbot`，实例在 `instances.one`，命令前缀为 `/`。两个开发组件沿用真实 Deliver／Send 流水线。

组件市场 ID 使用连字符前缀，例如 `alonabot-plugin-hello`；本地 ID 仍允许下划线，规则见 [CONTRACTS §9.1](../spec/CONTRACTS-v0.1.md#91-manifest-与构建产物)。从改名前的开发环境更新时，先停止 Core 和开发组件，备份原目录，再用新数据目录启动以重建状态库和组件数据。按新组件 ID 重配组件设置、实例的 `package`、`log.sources`、权限和工具策略；平台 `adapter_id` 与实例名沿用原值。Core 不自动迁移旧数据。

验证官方 help 工程可运行 `bin/alonabot check ./plugins/help`。使用上面的回环 Adapter 试用时，在 `run` 的组件目录列表追加 `./plugins/help`，再把注入请求的 `text` 改为 `%2Fhelp` 或 `%2Fwhoami`，在同一输出文件查看回复。

先构建 CLI（`go build -o bin/alonabot ./cmd/alonabot`），再创建与验证组件：

```fish
bin/alonabot init ./alonabot-plugin-hello
bin/alonabot build ./alonabot-plugin-hello
bin/alonabot check ./alonabot-plugin-hello
bin/alonabot test ./alonabot-plugin-hello
bin/alonabot pack ./alonabot-plugin-hello
```

`init`、`build`、`check`、`test`、`pack` 至多接受一个位置目录，省略时使用当前目录；旧的 `--dir` 会报未知参数。`init` 要求空目录，非空时在写入前报错。`--id` 缺省取目标目录绝对路径的目录名；例如 `alonabot-plugin-hello` 直接作为市场 ID，`my_plugin` 保持本地 ID 并输出缺市场前缀的警告。非法目录名会保留原字符并提示显式给出 `--id`。缺省角色为 plugin，`--role agent|adapter|webui` 创建对应工程；Adapter 的 `--adapter-id` 缺省等于最终组件 ID，可显式覆盖。插件与 Agent 导出 `New()`，包含一个英文示例命令、schema、i18n；Adapter 提供实例 schema，WebUI 提供 `assets/index.html`。模板忽略 `data/` 与 `bin/`，正式薄入口通过 overlay 提供。库模式宿主参考 [devbot](../examples/devbot/README.md)。

位于现有 Go module 中时复用该 module，保持其 go.mod；外部空目录的 module 缺省同为目录名，可用 `--module example.com/myplugin` 覆盖。发布版 CLI 用自身模块版本写 SDK require，开发版 `(devel)` 会提示执行 `go get github.com/AlonaBot/AlonaBot@<version>`，由作者选定版本。发布前修改 `plugin.yaml` 中的示例 source。

在包含 `plugin.yaml` 的组件工程中运行 `alonabot run --reload`，或者明确给出一个或多个工程目录：

```sh
bin/alonabot run ./alonabot-plugin-hello ./myadapter --data-dir ./data --reload
bin/alonabot run ./alonabot-plugin-hello ./mywebui --data-dir ./data --reload
bin/alonabot run ./examples/plugins/lifecycle --stdio
```

未给目录时，当前目录含 `plugin.yaml` 就运行该工程；其他情况按安装记录启动 Core。组件身份与角色取自工程 manifest，完成工程检查与进程组件构建后才启动组件。重复目录、重复 ID 都会报错；进程组件的 ID 或 `adapter_id` 与已安装包冲突时也会报错。

构建复用内部生成器与 `go build`，可执行文件写到 `<data_dir>/dev/bin/`，生成资源与薄入口通过 overlay 提供，作者目录中的源码和 module 文件保持原样。插件与 Agent 构建入口包；Adapter 工程导出 `New()` 并提供可导入的包，构建器为其生成调用 `sdk.RunAdapter(New())` 的薄入口。

开发组件通过 supervisor 在独立进程中运行，缺省使用 TCP loopback，`--stdio` 切换本次开发进程的传输。组件固定使用 `dev:<组件 ID>` 代次，配置每次从 `<data_dir>/plugins/<id>/config.json` 读取，缺失时按空对象准备；准备后的配置保存在开发代次快照中。受管 SQLite 为同目录的 `dev.sqlite`，跨重启保留。Adapter 实例配置来自 `core_config.json` 的 `adapters`。开发日志缺省为 debug，显式 `log.sources` 设置优先。

`--reload` 每 `DEV_RELOAD_POLL` 检查源码、`//go:embed` 资源、`go.mod`、`go.sum`、`plugin.yaml`、两个配置 Schema 和 `i18n/`，排除数据目录、vendor、testdata 及以 `.`、`_` 开头的目录。轮询间隔缺省为 500ms，例如 `--args DEV_RELOAD_POLL=1s` 可以覆盖。

源码或资源变化后重新构建，成功时只重启变化的组件；编译失败保留编译器诊断，旧组件继续运行。只改组件 `config.json` 时直接重启该组件，只改 Adapter 的一个实例配置时只重启该实例，二进制保持原样。不带 `--reload` 时构建并运行一次。

`--reload` 也监视 `core_config.json` 的其他设置：`admins` 更新后立即改变 Core Admin 权限，命令前缀从共享设置读取当前值，`default_locale` 对之后建立的 Session 生效。修改 Adapter 的 `default_instance` 后，未指定实例的发送立即改选新的缺省实例；`log.level` 立即更新全局日志级别，本次运行显式给出的 `--log-level` 或库模式 `LogLevel` 选项持续优先。这些修改保持组件进程与实例运行；同时修改设置与开发 Adapter 实例时，先刷新设置，再重启变化的实例。重新读取失败时记录错误并保留旧设置。开发 CLI 经 `internal/app.Core.ReloadSettings(ctx)` 与其他管理操作串行刷新同一份设置。

`role: webui` 的工程直接以 manifest 中 `assets` 指定的目录为静态根目录，覆盖已安装的 WebUI，等同于 `alonabot.WebUI(os.DirFS(<assets 目录>))`。一次运行只接受一个 WebUI 目录，重复时在启动任何组件前报错。开发 WebUI 按 D-140 跳过 Admin API 主版本检查；静态文件按请求读取，修改后立即可见，无需构建或重启，也无需 `--reload`。

独立进程调试时将调试器附加到组件 PID；需要在同一进程调试 Core 和组件时使用下文库模式。

## Core 设置与 HTTP 入口

`run` 取得 core.lock 后读取 `<data_dir>/core_config.json`，Core 设置统一从这个 JSON 对象读取；文件不存在时以 0600 权限写入缺省内容。JSON 语法、重复键或根节点错误时，以 usage 退出码报告键路径与行列号。未知键保留并忽略，类型或取值错误在运行副本中校正，记录路径与原因。

键与缺省值见 CONTRACTS §15.1；`base_url` 缺省为 `http://127.0.0.1:7650`，管理入口与插件路由共用这个地址。旧键 `admin_base_url` 与 `plugin_base_url` 作为无效配置保留，运行时使用 `base_url`。`admins`、`command_prefixes`、`reply`、`default_locale`、`log.sources` 与 `adapters` 已支持解析与保存；`command_prefixes_by_instance` 仍由当前运行时读取，配置文件的前缀与回复行为由 T-04.19 接入。本地 `run` 按监听、日志、代理与 Base URL 装配。

整个字符串值写成 `${NAME}` 时，读取时展开环境变量；NAME 由 `A-Z`、`a-z`、数字与下划线组成，首字符为字母或下划线。变量缺失时该值按不合法处理，恢复为缺省值并记校正记录（D-161）。

`(*settings.File).Update` 在未展开的文档上编辑，校验后以 0600 临时文件替换原文件，保留其他键的顺序与 `${...}` 写法，新键追加到所在对象末尾；写回使用两个空格缩进并以换行结束。文件没有格式版本（D-161）：不是合法 JSON 对象时拒绝启动，其中不合法的值按 CONTRACTS §5.5 校正，不改写文件。

插件由 SDK 先执行 `OnConfigLoad(ctx, raw)`，再补值与校正；`GET /api/v1/plugins/{id}/config` 返回 Active 的 `corrections` 和 `unknown`。`POST /api/v1/plugins/{id}/config/prune` 从 Active 删除无效键并保存为 Pending，随后调用管理 API 的 reload 生效。开发组件也按此顺序清理：reload 时原子更新其编辑文件，成功启动后清除 Pending；之后的 restart 保持清理结果。

部署运行中，手改设置后重启 Core 生效；带 `--reload` 的运行按上节刷新设置与开发实例，监听地址等启动时装配的设置仍需重启 Core。`--listen`、`--log-level`、`--log-format` 只在命令行显式给出时覆盖设置，只对本次运行有效，文件内容保持不变：

```sh
bin/alonabot run --data-dir ./data --listen 127.0.0.1:8080
```

用 `alonabot logs --level critical` 只看严重故障，或用 `alonabot run --log-level critical` 为本次启动提高全局输出阈值。运行中用 `alonabot log-level global critical` 保存全局阈值，用 `alonabot log-level <组件 ID> critical` 单独设置来源。查询筛选与输出设置见 [ADMIN-API §8](../spec/ADMIN-API.md#8-日志读取与输出级别)。

Core 在 `listen` 地址（缺省 `127.0.0.1:7650`）上提供唯一的公开 HTTP 入口。请求 Host 的主机名须匹配 `base_url`：`/plugins/{id}/...` 转发给该插件当前 Session 的路由，其余路径交给管理处理器；其他 Host 返回 421。缺省设置下，lifecycle 示例的私有页面地址为 `http://127.0.0.1:7650/plugins/demo/page`。listen 为非 loopback 地址时，`base_url` 必须是 https，或者配置非空的 `trusted_proxies`（由可信反代终止 TLS），设置文件与 `--listen` 都按此校验。

插件路由的全部响应（含 Core 的错误响应）统一设置 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff`，覆盖插件的同名头。Core 清除插件设置的全部 `Set-Cookie` 和 `Access-Control-Allow-Credentials`（含 trailer），Cookie 只以 WARN 记录名称；缺失 Content-Type 时补 `application/octet-stream`。HTML 页面按此沙箱策略返回。

设置了 `proxy` 时，Core 与子进程使用它作为 `HTTP_PROXY`/`HTTPS_PROXY`；为空时原样传递环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。

首次启动且尚无管理凭证时，Core 自动生成管理 token：stdin 与 stderr 都是终端时只向 stderr 显示一次；系统服务、容器等非终端启动时，以 0600 权限写入 `<data_dir>/admin_token`，INFO 日志记录文件路径。读取文件中的 token：

```sh
bin/alonabot admin-token show --data-dir ./data
```

`show` 直接读取文件，可在 Core 运行时执行，也可从非终端调用；文件不存在时返回英文错误。第一次登录成功后，Core 删除该文件。登录接口为 `base_url` 下的 `POST /api/auth/login`，要求受控 Origin、`Content-Type: application/json` 与正文 `{"token":"<token>"}`，成功后返回管理 Cookie 和 `csrf_token`；CLI 可用 `Authorization: Bearer <token>`。

私有插件路由复用管理认证：有效管理 Cookie 或 Bearer token；Cookie 认证的非 GET/HEAD 请求还须携带 `X-Alonabot-CSRF` 与受控 Origin，校验失败返回 401。公开路由无需管理会话，由插件按其声明的回调来源校验规则处理。转发给插件前，Core 清除管理 Cookie、Authorization 与 CSRF 头。

需要在启动前初始化凭证时，停止 Core 后在本机终端运行 `bin/alonabot admin-token init --data-dir ./data`。token 遗失时用 `bin/alonabot admin-token reset --data-dir ./data` 重新生成，旧 token 与全部管理会话随之失效。`init` 与 `reset` 都要求 stdin 与 stderr 是终端，只向 stderr 显示 token 一次；Core 运行中执行会因 core.lock 失败。

Core 运行且仍持有当前 token 时，可调用 `POST /api/auth/token/rotate` 在线轮换（[ADMIN-API §20](../spec/ADMIN-API.md#20-webui-需要的补充接口)）。从响应保存新 token，更新 CLI 使用的 `ALONABOT_ADMIN_TOKEN` 或 `--token-file`，并用新 token 重新登录。轮换会撤销全部旧管理会话。

调优参数在启动时解析；修改后重启生效。`default` 的前缀与 `reply` 属于 `core_config.json`；配置文件覆盖集合沿用固定 4 MiB 帧上限。查看全部调优参数及当前生效值：

```sh
bin/alonabot tune list
bin/alonabot --args REQUEST_TIMEOUT=2s tune list
bin/alonabot run --data-dir ./data --args REQUEST_TIMEOUT=2s --args ./tuning.txt
```

`--args` 可重复，放在子命令前后均生效，后出现的覆盖先出现的。直接参数支持逗号分隔的 `KEY=VALUE`；文件每行一项，空行和以 `#` 开头的注释行会被忽略。时长使用 `2s` 等 Go 时长格式；大小使用 `B`、`KiB`、`MiB`、`GiB`，`MB` 等十进制单位会报错。`run` 将最终覆盖值及出处记为 WARN。

`OPTIONS_PAGE` 的默认值和范围见 [CONTRACTS §16](../spec/CONTRACTS-v0.1.md#16-调优参数)，可用 `alonabot tune list` 查看；聊天目录每页的上限取当前 Core 生效值的十倍，请求的 `limit` 仍缺省为 50。

插件包导出 `func New() *sdk.Plugin`，在 `sdk.NewPlugin()` 返回的对象上登记命令、生命周期回调与业务行为。构建工具自动扫描 manifest、配置 Schema、Service Schema、i18n 和参数名，经 overlay 编入程序；作者工程中没有 `alonabot_gen.go`，作者代码无需生成相关内容。构建或只读检查：

```fish
bin/alonabot build ./alonabot-plugin-hello --goos linux --goarch amd64 --tags a,b
bin/alonabot check ./alonabot-plugin-hello --goos linux --goarch amd64 --tags a,b
```

`build/check` 使用 `--goos`、`--goarch`，目标缺省沿用 Go 环境；`pack` 使用可重复的 `--target <os>/<arch>`，省略时取本机平台。三者均用 `--tags` 选择源码；扫描、依赖分析和正式编译统一使用 `CGO_ENABLED=0`。开发运行的 BuildDev 沿用开发环境的 cgo 设置。`build` 默认输出到 `<工程>/bin`，可用 `--out DIR` 覆盖。进程产物包含独立的 `component`（Windows 为 `component.exe`）、资源旁车及包内 `plugin.yaml`；作者 manifest 保持原样。WebUI build 发布静态资源与包内 manifest。`check` 只读扫描入口与资源，保留全部静态诊断与位置。

`test` 在全新临时数据目录中构建宿主平台组件，完成真实握手、准备、注册、激活与停止，报告注册命令及清理结果；Adapter 保持零平台实例，WebUI 检查静态加载资源。进程组件的 Bootstrap 下发临时状态根内 `plugin_data/<id>/` 的绝对路径，测试结束时随临时目录清理。准备、注册、停止或清理失败均非零退出。

`pack` 为每个 `--target` 分别扫描源码、检查资源并构建，输出 `<id>-<version>-<os>-<arch>.tar.gz`；静态 WebUI 只输出一个 `<id>-<version>-any.tar.gz`，带 Go 入口的 WebUI 按平台构建。全部归档共享一份列出本次全部资产的 `release.json`，`--out` 默认同 build。资产地址为 `--asset-base` 加 `/` 与归档名；省略时，`source: github.com/<owner>/<repo>` 推出 `https://github.com/<owner>/<repo>/releases/download/v<version>`，其他来源须显式给出绝对 https 前缀，前缀保留路径且无 query/fragment。进程包的 `build-info` 是 `debug.BuildInfo.String()` 文本；生成代码、资源旁车与归档共用扫描时校验的资源快照，配置 schema 使用同一规范内容；扫描后保存的工程资源留待下次构建。

配置字段的 `example` 引用图片时，`check`、`build`、`pack` 检查两种配置 schema 中的引用及图片签名；被引用的图片随资源快照进入构建目录和归档。示例写法见 [插件开发规范](plugin-guide.md#配置字段的回复示例)。

```fish
bin/alonabot pack ./alonabot-plugin-hello --target linux/amd64 --target linux/arm64 \
  --asset-base https://example.com/releases/v0.1.0 --key ./release.key --out ./dist
```

`--key` 可选，文件内容为标准 base64 编码的 64 字节 Ed25519 私钥，允许首尾空白。提供时只对最终合并清单的原始字节签名一次，额外写 `release.json.sig` 和 `release.pub`，分别为标准 base64 签名及公钥；省略时生成 unsigned release，并清除输出目录中旧的签名与公钥。密钥读取或签名失败直接报错。私钥、运行配置与数据目录均应留在打包资源之外。

真实进程验收（Unix；构建 Core、lifecycle 与临时热部署工程）：

```sh
go test -count=3 ./cmd/alonabot/ ./internal/core/supervisor/
go test -race -count=3 ./internal/core/supervisor/
```

文档改动用 `git diff --check` 检查空白错误，链接与技能元数据由审查核对。包文档、命名、注释与测试提交范围由 [编码规范](coding-standards.md) 与 [注释技能](../.agents/skills/alonabot-code-commenting/SKILL.md) 指导审查。

## 库模式运行（同一进程）

根包 `alonabot` 提供 `Run`、`Start`、`*Core` 与 `Wait`，选项为 `Plugin`、`Adapter`、`WebUI`、`DataDir`、`Args`、`Listen`、`LogLevel`、`LogFormat`。它们转发到 `internal/app`，在当前进程中运行 Core 与开发中的组件，示例见 `examples/devbot`：

```go
core, err := alonabot.Start(ctx, alonabot.DataDir(dir), alonabot.Plugin(plugin.New()), alonabot.Adapter(adapter.New()))
if err != nil {
	return err
}
return core.Wait() // ctx 取消（例如 SIGTERM）后在 STOP_TOTAL 内停止全部组件
```

```fish
bin/alonabot build ./examples/devbot
examples/devbot/bin/devbot -data-dir ./data
```

组件以 source=dev 登记，代次固定为 `dev:<组件 ID>`。库模式由 `alonabot build <main 目录>` 构建，扫描 main 依赖中的组件入口（含模块缓存中的第三方组件），将全部资源登记经 main 包的 overlay 编入。默认只输出 `<main 目录>/bin/<目录名>` 可执行文件（Windows 加 `.exe`），可用 `--out` 修改输出目录。组件源码与模块缓存保持只读。同包多次 New 共用资源；main 的包级变量也可调用组件的 New。直接 `go build` 的库模式产物在启动缺少资源的组件前返回提示 `alonabot build` 或 `alonabot run` 的英文错误。

库模式更改源码或资源后重新运行 `alonabot build` 并重启程序。未捕获的后台 panic 使整个进程非零退出且无有限重启；SDK Handler 边界内捕获的 panic 记录 stack 并返回 `INTERNAL`。同进程组件超过停止预算仍未退出时，`Wait` 返回错误；独立进程由 supervisor 隔离并有限恢复，两种模式的 SIGINT/SIGTERM 停止总预算缺省均为 `STOP_TOTAL=10s`。开发来源的 update、rollback、uninstall 返回 `FAILED_PRECONDITION`；WebUI/CLI 修改开发配置后写回并重启由 T-09.14 承接，手改配置的轮询行为见上节。

库模式公开入口由根目录 `alonabot_test.go` 的 `TestLibraryLifecycle` 验证；装配、内部服务、候选与 WebUI 的 9 个集成测试文件位于 `internal/app/`。开发 CLI 与热部署测试位于 `cmd/alonabot/`，覆盖经 `internal/app` 运行的独立进程开发组件与单组件重启。

## Adapter 的 HTTP 路由

Adapter 在 `adapter.Descriptor.HTTP` 中用 `[]adapter.HTTPRoute` 声明路由，对外位于统一端口的 `/plugins/<Adapter 包 ID>/...` 下声明路由，例如 OneBot 11 反向 WebSocket。路由规则与插件相同：私有路由需要 Core 会话，非 GET/HEAD 的 Cookie 请求还要验证 CSRF 与 Origin。`Auth` 的零值要求 Core 会话；公开路由写 `Auth: adapter.AuthPlatformToken`：Core 代理保留入站的 `Authorization` 交给 Adapter，由 Adapter 校验平台凭证，并给出错误时的平台响应。插件路由的 `Authorization` 一律被清除。两者都会清除入站的 Cookie、CSRF、Forwarded 与 X-Forwarded-*，并注入内部代理凭证；SDK 先恒定时间比较该凭证，比较后移除，再进入用户 Handler。WebSocket Upgrade 由代理双向转发，连接寿命不受普通请求期限限制，路由撤销或 Session 失效时由 Core 关闭。

## 备份与恢复

整机备份前先停止 Core 和全部插件进程，再复制完整的 `data_dir`。备份须同时覆盖 Core 元数据、程序包、插件配置及各代次的受管 SQLite 数据库。单独备份 `state.sqlite` 只保存 Core 元数据，不包含插件业务数据。

插件自管的外部数据库与框架目录之外的文件，按插件的维护说明单独备份。恢复时保持 Core 和插件停止，还原完整的 `data_dir`，并同步恢复插件自管的外部数据，再启动 Core。恢复旧备份会丢弃备份之后的新写入，执行前须确认要恢复的数据时点。

## 外部测试环境

- **Telegram：** 真实测试需要授权的测试 Bot，目前没有，相关项记为「未运行」；Adapter 测试改用本地 Bot API 模拟服务。
- **目标平台：** 本机只能运行 darwin/arm64；其余四个组合只做 `CGO_ENABLED=0` 交叉编译，尚无运行证据。

## 规则与检查

| 规则 | 依据 | 检查 | 状态 |
| --- | --- | --- | --- |
| 生成代码可重复生成，结果稳定 | AGENTS「项目约定」 | 执行 `go generate ./...` 后检查 `git diff --exit-code` | 已观测（手动执行） |
| WIRE-FORMAT 规定的协议字段号与枚举值（含 `ToolEffect` 及三个 `effect` 字段） | spec/WIRE-FORMAT.md §1/§2/§3.3/§3.5/§5A | `contracts/wirespec_test.go`（`go test ./contracts/`） | 已观测 |
| Core 与 SDK 根包只链接允许的模块 | CONTRACTS §10、D-52 | 审查时运行上方的 `go list -deps` 命令（需要时按目标平台设 GOOS/GOARCH） | 人工/agent 审查规则，不由 CI 自动证明 |
| 五目标平台 `CGO_ENABLED=0` 可构建 | CONTRACTS §6.1/§7.1 | 下方交叉编译命令（手动执行） | 已观测（只验证了构建；运行见 V-32） |
| 包文档、命名、注释与测试提交范围 | coding-standards.md | 代码审查与 PR 清单 | 人工/agent 审查规则，不由 CI 自动证明 |
| 验收证据 | spec/TEST-MATRIX.md §3 | 按编号记录测试位置、命令、平台和结果 | 各分项的运行结果与未运行项见证据表 |

状态说明：**已定义** 表示规则和检查方式已写下；**已实现** 表示检查代码已存在；**已观测** 表示检查真实运行过并有结果。工作流本身不代表目标提交已经验证。

## GitHub Actions

[Quality gate](../.github/workflows/ci.yml) 在 PR 到 main、main push 和手动触发时运行 `git diff --check`。存在产品 Go 文件时，才运行 Go 格式、Protobuf 生成一致性、vet、race 测试和五目标平台 `CGO_ENABLED=0` 构建；无产品代码时明确标注 Go 检查未运行。

生成检查从官方 Protobuf release 下载当前约定的 protoc 36.0；工具链变更时同时更新生成器、开发指南和工作流。代码 job 使用只读 token，不使用生产凭据。

[PR metadata](../.github/workflows/pr-metadata.yml) 校验非草稿 PR 的实际说明和清单；[triage](../.github/workflows/pr-triage.yml) 同步标签与 assignee，[auto-assign](../.github/workflows/auto-assign.yml) 请求非作者维护者审阅。这些 `pull_request_target` 流程不 checkout/执行 PR 代码，新工作流合入默认分支后才开始处理后续事件。Issue 表单自带 assignee，不依赖工作流执行。

远端合并保护由维护者单独管理，不能因存在这些 YAML 就声称检查不可绕过。本次建设不添加发布、真实 Telegram 调用或自动关闭 PR 的流程。

## 动态提醒与静态任务暂停恢复

需要十分钟后给当前会话发提醒时，先在 `New()` 中登记任务声明和处理函数，命令执行时再创建实例。完整示例见 [showcase](../examples/plugins/showcase/main.go)：`/demo remind meeting 600 开会` 创建当前用户在当前会话内的 `meeting` 提醒，`/demo cancel meeting` 取消。同用户、同会话、同 key 再次创建会替换提醒；不同用户或会话各自独立，回复始终使用用户输入的 key。

```go
type ReminderArgs struct {
    Session string `json:"session"`
    Text    string `json:"text"`
}

// 创建与取消共用编码，将可信身份与用户 key 一起放入内部 key。
func reminderKey(c *sdk.Context, key string) (string, error) {
    user, session := c.User(), c.Session()
    internalKey := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, key)
    if len(internalKey) > 128 {
        return "", errcode.InvalidArgument.New("reminder key with user and session scope exceeds 128 bytes")
    }
    return internalKey, nil
}

// New 中创建并登记一次，命令闭包持有同一个声明值。
reminder := task.Define[ReminderArgs]("remind")
task.On(p, reminder).Handle(func(c *sdk.Context, args ReminderArgs) error {
    _, err := c.SendTo(args.Session, message.Text(args.Text))
    return err
})
p.Command("remind").Handle(func(c *sdk.Context, key, text string) (message.Content, error) {
    internalKey, err := reminderKey(c, key)
    if err != nil {
        return nil, err
    }
    _, err = reminder.At(c.Context(), time.Now().Add(10*time.Minute),
        ReminderArgs{Session: c.Session(), Text: text}, task.Key(internalKey))
    if err != nil {
        return nil, err
    }
    return message.Text(key), nil
})
p.Command("cancel").Handle(func(c *sdk.Context, key string) error {
    internalKey, err := reminderKey(c, key)
    if err != nil {
        return err
    }
    return reminder.Cancel(c.Context(), internalKey)
})
```

`reminderKey` 使用 `c.User()` 与 `c.Session()` 的可信身份，以长度前缀区分完整字段。128 字节上限包含身份作用域和用户 key；超长返回明确的 `INVALID_ARGUMENT`，保留输入原文。示例需导入标准库 `fmt`、`sdk/errcode` 与 `sdk/task`。

任务回调没有 Actor 与默认会话；将所需会话、用户或取消键写入参数，并用 `c.SendTo(sid, content)` 指定发送目标。主动发送遵循 Core 的 `allowed_adapter_instances` 准入。参数类型遵循 JSON Service 的类型规则，使用包级具名、非泛型的值类型，顶层编码为 JSON 对象；生成器检查嵌套字段与自定义编码限制。新增声明后直接运行测试，或运行 `alonabot build` 构建组件。

`reminder.Every(ctx, time.Hour, args, task.Key(internalKey))` 创建周期实例；`reminder.Cron(ctx, "0 9 * * *", args, task.Key(internalKey), task.Timezone("Asia/Shanghai"))` 按 IANA 时区运行，其中 `internalKey` 沿用上述身份作用域编码。`At`、`Every` 上设置 `Timezone` 返回 `INVALID_ARGUMENT`。省略 `Key` 时 SDK 生成随机 key，成功返回实际使用的 key；同任务名、同 key 再次创建会整体替换规则和参数，在途旧一轮继续执行。`Cancel` 删除实例并协作取消在途调用，函数实际返回后释放运行槽。未知实例保留 Core 的 `NOT_FOUND`。

`reminder.List(ctx)` 读取全部分页，返回完整的 `task.Instance[ReminderArgs]`。`At`、`Interval`、`Cron` 中仅对应触发规则的字段有值，`Timezone` 随 Cron 返回；`NextAt == nil` 表示协议没有下一次触发时刻。任一页读取或参数解码失败都会返回错误。

所有任务客户端函数在 Activated 后可用；`OnStart`、`OnReady`、停止后与已失效的 Handler Context 返回 `FAILED_PRECONDITION`。在 Handler 中传 `c.Context()` 以保留父调用关联；独立后台 context 只有组件身份。Core 保存任务定义与 interval 起点，重启后安排下一未来触发；离线、未激活与暂停期间错过的触发作废，一次性过期实例删除。

静态任务继续在 `New()` 中声明，例如 `p.Every("tick", time.Minute).Handle(tick)`。在已激活的 Handler 或后台工作中暂停、恢复与查询：

```go
if err := task.Pause(ctx, "tick"); err != nil {
    return err
}
state, err := task.Status(ctx, "tick")
if err != nil {
    return err
}
if state.LastAt != nil {
    slog.Info("last tick", "at", *state.LastAt)
}
return task.Resume(ctx, "tick")
```

`task.State.LastAt` 记录本次 Core 运行最后一次实际派发，首次派发前和 Core 重启后为 `nil`；暂停时 `Paused` 为 true、`NextAt` 为 `nil`。暂停保留在途函数继续执行，恢复沿保存的 interval 起点安排未来触发。暂停状态跨重启保存，重复暂停或恢复成功，未登记的 job ID 返回 `NOT_FOUND`。

## 插件测试

在插件包的 `main_test.go` 中直接调用 `New()`，用 `plugintest.Start` 运行真实 Core、SDK Session 与测试 Adapter。例如 lifecycle 的 `/ping`：

```go
package main

import (
    "testing"

    "github.com/AlonaBot/AlonaBot/plugintest"
)

func TestPing(t *testing.T) {
    p := New()
    h := plugintest.Start(t, plugintest.Plugin(p))
    alice := plugintest.User{ID: "alice"}
    result := h.Text(plugintest.Private(alice.ID), alice, "/ping")
    if len(result.Sends) != 1 || result.Sends[0].Text() != "pong" {
        t.Fatalf("ping = %+v", result)
    }
    for _, outcome := range result.Outcomes {
        if outcome.Err != nil {
            t.Fatal(outcome.Err)
        }
    }
}
```

直接运行 `go test ./...`。plugintest 按被测插件的入口包自动扫描工程身份与资源，沿用测试二进制的 GOOS、GOARCH 与 build tags；同一入口包在一个进程中只扫描一次。每次 Start 创建独立的临时目录、配置和 `dev.sqlite`，在测试清理时停止组件；同一 Harness 的注入与读取按顺序调用。

`Message(chat, user, message.Chain{...})` 可以注入混合片段，`Sent.Text()` 按顺序提取文本。`Sent.To`、`Chain`、`Result` 保留目标、消息链与回执；`Op` 区分 send/edit，`Edited` 保存编辑句柄。`Event` 注入通过 `Adapter` 选项声明的扩展事件。

**权限与配置。** `Admin("root")` 登记 Core Admin；`User.Permissions` 携带该会话的平台权限，需通过 `Adapter` 声明它们。例如：

```go
moderator := plugintest.User{
    ID: "mod",
    Permissions: []sdk.Permission{sdk.PermissionName("plugintest.moderator")},
}
h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Admin("root"),
    plugintest.Adapter(plugintest.AdapterID, adapter.PermissionDeclaration{
        Name: "plugintest.moderator", Implies: []string{"chat_admin"}, LabelKey: "moderator",
    }))
result := h.Text(plugintest.Group("room"), moderator, "/mod")
```

其中 `sdk` 与 `adapter` 分别来自 `sdk`、`sdk/adapter`。配置按被测插件的 Schema 提供：

```go
h := plugintest.Start(t, plugintest.Plugin(p),
    plugintest.Config(p, `{"greeting":"Hello"}`),
    plugintest.AllowedInstances(p, plugintest.InstanceID),
    plugintest.Capabilities("message.send", "message.edit"))
```

`AllowedInstances` 缺省允许全部实例，`AllowedInstances(p)` 收紧为空集合，job 和后台的显式发送会得到 `PERMISSION_DENIED`；可信入站事件的原目标回复沿用原有豁免。`Capabilities` 完整替换缺省能力，上例可验证按钮、相册等片段的 `UNSUPPORTED` 错误及插件降级行为。拒绝的发送在 Outcome 中保留错误。

**按钮。** 从发送结果点击动作名，用户取本次点击者；Core 恢复 token 中的 action、版本与业务值，并逐次鉴权：

```go
menu := h.Text(plugintest.Group("room"), alice, "/menu")
clicked := h.Click(plugintest.User{ID: "bob"}, menu.Sends[0], "first")
```

缺失的动作、URL 按钮及缺少已发送回执属于测试输入错误，Click 通过 `t.Fatal` 报告。过期 token、失效代次和权限拒绝会返回空发送及空 Handler Outcome。

**静态与动态任务。** `h.Trigger(p, "tick")` 立即执行已登记静态任务；`h.TriggerTask(p, "remind", key)` 定位已创建实例，使用 Core 保存的参数。两者通过 Scheduler 的正常 REQUEST/RESPONSE、job scope、运行槽和期限执行，Handler 使用 `c.SendTo(sid, content)` 指定发送目标。准入及定位错误使测试失败；已派发 Handler 的错误保留在 `Result.Outcomes`。

showcase 的提醒先通过命令创建，再用含用户与会话的完整内部 key 触发：

```go
h.Text(chat, alice, "/demo remind lunch 3600 soup")
user := message.UID(plugintest.AdapterID, alice.ID)
session := message.SID(plugintest.AdapterID, "private:alice", "")
key := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, "lunch")
reminder := h.TriggerTask(p, "remind", key)
```

完整示例见 [showcase/main_test.go](../examples/plugins/showcase/main_test.go)。一次性实例在实际派发时消费；周期实例保留起点并继续安排未来 tick。静态任务暂停或运行中、动态实例运行中或排队时，显式触发返回 `FAILED_PRECONDITION`；动态全插件并发名额已满时沿用到期队列。

**多轮等待。** 进入 `core.wait` 后注入立即返回 `Result.Waiting=true`，进入等待的路由 `Outcome.Waiting=true`；问题发送仍在本次 `Sends` 中。下一条同用户、同会话消息由普通入口注入：

```go
question := h.Text(chat, alice, "/ask")
if !question.Waiting {
    t.Fatalf("ask = %+v", question)
}
answer := h.Text(chat, alice, "answer")
```

命中等待后的发送和最终错误归入新注入，直到延续调用结束或再次等待；其他用户或会话继续正常分发。注入返回由 Core 的完成或等待通知决定。

**后台发送。** `h.Background()` 按 Adapter 接受顺序返回并清空尚未交付的发送和编辑。注入开始到完成通知之间的发送随 `Result.Sends` 交付，区间外的留给 Background，每条记录消费一次。测试用自己的 channel 控制后台工作，并等待 Send 返回后读取：

```go
send, done := make(chan struct{}), make(chan error, 1)
var cancel context.CancelFunc
p.OnReady(func(ready context.Context) error {
    ctx, stop := context.WithCancel(context.WithoutCancel(ready))
    cancel = stop
    go func() {
        select {
        case <-send:
        case <-ctx.Done():
            return
        }
        _, err := sdk.Send(ctx, "plugintest:private:alice", message.Text("background"))
        done <- err
    }()
    return nil
})
p.OnStop(func(context.Context) error { cancel(); return nil })
h := plugintest.Start(t, plugintest.Plugin(p))
send <- struct{}{}
if err := <-done; err != nil {
    t.Fatal(err)
}
records := h.Background()
```

`OnReady` 的回调 context 在返回时取消，上例保留其 Session 归属，后台生命周期由 OnStop 管理。channel 确定发送时刻，Core 通知确定注入终点。

仓库完整示例覆盖 echo、lifecycle、showcase、configdb 与 media。configdb 的 KV 与文件命令见[插件指南](plugin-guide.md#保存少量状态)：

```sh
go test -race -count=2 ./plugintest/... ./examples/plugins/...
```

**LLM 替身。** 测试调用文本生成的插件时，用 `LLM(handler)` 提供结果。插件仍经 `@agent/llm.generate` 调用同一 Core 中的测试 Agent。先在插件的 `New()` 返回前登记命令（导入 `sdk/llm` 与 `sdk/message`）：

```go
p.Command("llm").Handle(func(c *sdk.Context, prompt string) (message.Content, error) {
    text, err := llm.Ask(c.Context(), prompt)
    if err != nil {
        return nil, err
    }
    return message.Text(text), nil
})
```

在同包的 `main_test.go` 中加入，plugintest 会自动扫描调用方的 `llm.Request`、`llm.Response` Schema：

```go
package main

import (
    "context"
    "testing"

    "github.com/AlonaBot/AlonaBot/plugintest"
    "github.com/AlonaBot/AlonaBot/sdk/errcode"
    "github.com/AlonaBot/AlonaBot/sdk/llm"
)

func TestLLMReply(t *testing.T) {
    p := New()
    h := plugintest.Start(t, plugintest.Plugin(p),
        plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error) {
            if len(req.Messages) != 1 || req.Messages[0].Role != "user" || req.Messages[0].Text != "hello" {
                return llm.Response{}, errcode.InvalidArgument.New("unexpected prompt")
            }
            return llm.Response{
                Text: "fixed answer", Provider: "synthetic", Model: "test-model",
                InputTokens: nil, OutputTokens: nil,
            }, nil
        }))
    alice := plugintest.User{ID: "alice"}
    result := h.Text(plugintest.Private(alice.ID), alice, "/llm hello")
    if len(result.Sends) != 1 || result.Sends[0].Text() != "fixed answer" {
        t.Fatalf("LLM reply = %+v", result)
    }
    for _, outcome := range result.Outcomes {
        if outcome.Err != nil {
            t.Fatal(outcome.Err)
        }
    }
}
```

上例的 nil 用量经 Service 编码后仍为 nil；已知用量为零时，返回指向 `0` 的 `*int`。验证失败路径时，将 handler 换成：

```go
plugintest.LLM(func(context.Context, llm.Request) (llm.Response, error) {
    return llm.Response{}, errcode.Unavailable.New("synthetic generation failure")
})
```

同一条命令会得到空 `result.Sends`，`result.Outcomes[0].Err` 保留 `errcode.Unavailable` 与 `Message`。未设置 LLM 时，`@agent` 调用同样返回 `UNAVAILABLE`。多次设置 LLM 时最后一个函数生效；`LLM(nil)` 在 Start 启动前使测试失败。handler 接收关联的 context，并应响应其取消与期限。Schema 负责结构校验，业务规则与错误由 handler 决定。

仓库的完整集成测试见 [plugintest/llm_test.go](../plugintest/llm_test.go)。运行：

```sh
go test -race -count=2 ./plugintest/...
```
