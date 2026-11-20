# 发行维护

本文面向发行维护者，说明从源码构建、核对候选到正式签名发布的步骤。组件作者使用[插件指南](plugin-guide.md#9-发布)发布自己的组件；用户安装见[部署指南](deployment.md)。

## 发布前准备

1. 固定已审查的源码提交与发行版本，确认该提交 CI 通过；核对 SDK、组件版本与说明一致。
2. 阅读[当前发布说明](release-notes/v0.1.0.md)，保留真实平台结果、已知限制及已批准的延期项。
3. 准备开发工具链和已锁定依赖，使用独立输出目录。先使用仓库外的临时测试签名密钥验证候选，正式私钥只由发行维护者控制。
4. 解决下述组件在线元数据发布问题，确认 SDK tag、模板依赖与主仓库可访问性相符。主仓库正式公开后补验模板 CI，正式发行前替换候选 SDK 版本。

## 发行构建

在仓库根目录生成包含 Core、Telegram Adapter、Agent、WebUI 与 help 的发行候选：

```fish
go run ./scripts/dist --version 0.1.0 \
  --key ~/.config/alonabot/release.key --out ./dist
```

构建机需要[开发指南](development.md#准备环境)中的 Go、Node 与 npm，并可读取已锁定的 Go 模块和 npm 包。程序在暂存目录保留根 `go.mod`、`go.sum` 及 `sdk/`、`contracts/`、`internal/`、`web/admin/`，跳过 WebUI 的旧 `node_modules` 与 `dist`；在暂存 WebUI 中运行 `npm ci` 与 `npm run build`，再由主机版 `alonabot pack <组件目录> --target <os>/<arch> ... --asset-base <前缀> --out <目录>` 为每个组件调用一次，传入全部目标；pack 写出合并清单，dist 再签名并自检。组件版本沿用各自 `plugin.yaml`。默认目标为 `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`，可用 `--platforms linux/amd64,linux/arm64` 选择子集。离线 npm 缓存已齐全时，可在命令前加 `env npm_config_offline=true`。

`--key` 文件保存标准 base64 编码的 64 字节 Ed25519 私钥，允许末尾空白。私钥放在仓库及组件资源目录之外，设置为仅发布维护者可读（0600），通过受控环境或离线介质保管和备份。程序只从该文件读取私钥，子进程不接收私钥；发行物只携带公钥与签名。发布前先用临时测试密钥验证流程，正式签名使用维护者控制的密钥。

`dist/` 中每个平台生成一个 `alonabot-<version>-<os>-<arch>.tar.gz`（Windows 为 `.zip`），解压后的同名目录包含 `alonabot`（Windows 为 `alonabot.exe`）和 `packages/<组件ID>/`。每个组件目录包含该平台安装包、合并全部目标平台的 `release.json`、`release.json.sig` 与 `release.pub`；官方 WebUI 按五个目标平台出包，分别包含 `component`（Windows 为 `component.exe`）与同一份 `dist`。另在 `dist/packages/<组件ID>/` 汇集各平台安装包与同一份签名清单，供后续发布使用。`dist/SHA256SUMS` 列出所有外层发行归档的摘要，格式为摘要、两个空格和文件名。

安装包下载前缀默认取 `https://github.com/AlonaBot/AlonaBot/releases/download/v<version>`，可用 `--asset-base https://example.com/releases/v0.1.0` 覆盖。发行 Core 通过 `-ldflags -X` 嵌入官方公钥，安装官方签名包时免首次来源与发布者确认；签名与摘要仍按安装流程验证，已有发布者绑定继续生效。未注入公钥的开发构建保留原确认流程。

`--version` 决定发行归档名、默认下载前缀及注入的 Core 版本；`GET /api/v1/core` 返回该发行版本。开发构建沿用 Go module 构建信息，源码构建通常显示 `(devel)`。

## 构建容器镜像

自行构建时，使用[开发指南](development.md#准备环境)锁定的 Go、Node/npm 与 Docker BuildKit/buildx。`VERSION` 必填，签名私钥使用发行构建相同的 base64 Ed25519 格式，存放在构建上下文之外，权限为 0600；构建验收使用临时密钥，正式发行使用维护者控制的密钥。密钥通过必需的 BuildKit secret 挂载传入。构建阶段直接调用 `scripts/dist`，沿用其中的 npm ci/build、签名和四个官方包清单。BuildKit 的缓存判定忽略 secret 内容；命令使用 `--no-cache-filter build` 重新执行签名阶段，确保同版本更换密钥时使用本次密钥。

```fish
set version 0.1.0
set release_key ~/.config/alonabot/release.key
docker buildx build --platform linux/amd64,linux/arm64 --no-cache-filter build \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  --output type=oci,dest=/tmp/alonabot-$version.oci.tar \
  -t ghcr.io/alonabot/alonabot:$version .
```

需要覆盖组件下载前缀时，加 `--build-arg ASSET_BASE=https://example.com/releases/v0.1.0`；省略时沿用 `scripts/dist` 的 GitHub Release 地址。下面分别构建并在可运行相应平台的主机上启动本地镜像；跨架构运行需要 Docker 已配置的模拟器。

```fish
docker buildx build --platform linux/amd64 --load --no-cache-filter build \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  -t alonabot:$version-amd64 .
docker run -d --name alonabot-amd64 --platform linux/amd64 --init \
  -p 127.0.0.1:7650:7650 -v alonabot-amd64-data:/data alonabot:$version-amd64
docker exec alonabot-amd64 alonabot admin-token show --data-dir /data
# 完成该平台验收后停止，释放宿主端口，再运行另一个平台。
docker stop alonabot-amd64

docker buildx build --platform linux/arm64 --load --no-cache-filter build \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  -t alonabot:$version-arm64 .
docker run -d --name alonabot-arm64 --platform linux/arm64 --init \
  -p 127.0.0.1:7650:7650 -v alonabot-arm64-data:/data alonabot:$version-arm64
docker exec alonabot-arm64 alonabot admin-token show --data-dir /data
```


## 核对发行候选

使用仓库提供的 [release-check](../scripts/release-check/README.md)读取实际归档、组件包、签名、公钥、manifest 与 build-info，保留原始 JSON 和完整错误。生命周期联验使用独立目录与测试组件包，操作入口和输入格式以检查器说明为准。

五目标构建与运行验收使用[平台验收工具](../scripts/platform/README.md)，安装、更新与恢复场景使用[场景工具](../scripts/scenarios/README.md)。已有有效证据在来源与产物保持适用时复用，只补本轮改变或缺失的分项。

`install.sh` 修改后执行 `sh -n deploy/install.sh`，有 shellcheck 时一同检查；实际安装、升级、校验失败和卸载验收需要可恢复的 systemd 或 launchd 环境。所有测试使用授权环境，保留数据恢复路径。

## 组件在线元数据发布

正式发布前需确定四个官方组件各自的元数据入口。`scripts/dist` 将清单分别写到 `dist/packages/<组件ID>/release.json`，并生成同名的 `.sig` 和 `release.pub`；GitHub Release 的资产名称在同一次发布中唯一。在线 `install --source owner/repo` 与更新逻辑按来源仓库读取固定名称的 `release.json`，一份清单只描述一个组件。

发行归档已将各组件清单放在各自的 `packages/<组件ID>/`，本地安装和首次预装使用这组元数据。外部组件包的在线安装、更新需要与读取入口、签名及来源绑定相符的发布安排；该安排确定并验证后，维护者再开放正式在线分发。现有候选验收覆盖归档和本地清单，尚缺这条公开下载链。

## 正式发布

1. 将 SDK 固定到正式 tag，更新插件模板及示例的相应引用并验证。
2. 使用官方私钥，从同一固定源码运行 dist 入口，重新构建并签名正式资产；核对发行公钥与可信来源。
3. 上传五平台归档、归档摘要、`install.sh` 及固定版本 `compose.yaml` 到同一 GitHub Release；归档内保留各组件的清单、签名与公钥。外部组件包及其元数据按上一节已解决并验证的发布安排上传。
4. 推送已核对的 linux/amd64、linux/arm64 镜像与多架构清单，核对版本 tag、latest 与 compose 中的版本关系。
5. 发布说明明确列出安装方法、升级与备份步骤、目标平台的真实验证结果及已知缺项。正式资产发布后再更新 README 的默认安装入口。

当前候选使用测试密钥；正式 SDK tag、官方签名、GitHub Release 和 GHCR 发布由维护者执行。发布失败时保留错误、产物与当前线上版本，修复受影响步骤后再继续。
