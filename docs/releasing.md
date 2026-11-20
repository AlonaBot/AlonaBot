# 发行维护

整合发行保留在 Core 主仓，默认包含 Core、WebUI 与 Agent。Telegram、OneBot 11 Adapter 和 help 插件在各自仓库发布，用户按需安装。安装步骤见[部署指南](deployment.md)，组件作者的发布流程见[插件指南](plugin-guide.md#9-发布)。

## 固定源码与发行说明

1. 选择已审查的 Core 提交和发行版本，确认相应检查通过。
2. 在 `components.lock.json` 固定五个官方组件的 Go module 精确版本。整合构建读取其中的 Agent 和 WebUI；跨组件测试还使用其他条目。组件版本来自各自 `plugin.yaml`，可独立于 Core 版本升级。
3. 将用户可见变更整理到 `changelog/v<版本>/en.md` 与 `zh-CN.md`。保留真实平台结果、资源实测失败和已批准的延期。日常条目放在 `changelog/unreleased/`，约定见[更新日志](../changelog/README.zh-CN.md)。
4. 正式发行前固定 SDK tag，并更新组件和插件模板的引用。主仓公开后补验模板 CI。

生成的 Release 正文英文在前、简体中文在后，自动去掉 HTML 编辑注释：

```fish
go run ./scripts/release-notes --version 0.1.0 > /tmp/alonabot-release-notes.md
```

## 构建整合归档

在主仓根目录执行：

```fish
go run ./scripts/dist --version 0.1.0 \
  --key ~/.config/alonabot/release.key --out ./dist
```

构建需要 Go、Node/npm 及已锁定的 Go/npm 依赖，版本见[开发环境](development.md#准备环境)。工具按发行锁下载 Agent 与 WebUI，在临时副本中构建前端和平台包；Core 源码取当前 checkout。源仓库中的 `node_modules`、`dist` 和依赖文件保持原样。默认目标为 `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`，用 `--platforms linux/amd64,linux/arm64` 选择子集。

`--key` 文件保存标准 base64 编码的 64 字节 Ed25519 私钥，可带末尾空白。密钥置于仓库与组件资源之外，权限设为 0600。候选验收使用临时测试密钥，正式签名由维护者使用官方私钥执行。私钥只由 dist 读取，子进程只得到构建所需的公钥。

每个平台生成 `alonabot-<version>-<os>-<arch>.tar.gz`，Windows 为 `.zip`；归档含 Core 与 `packages/<组件ID>/`。每个组件目录含本平台安装包和完整的 `release.json`、`release.json.sig`、`release.pub`。WebUI 包包含后台进程和同一份前端产物。

`dist/packages/<组件ID>/` 汇集该组件全部平台包及三份元数据。默认下载 URL 指向该组件仓库的 `releases/download/v<组件版本>`；`--asset-base https://example.com/releases/v0.1.0` 可显式覆盖。Core 版本由 `--version` 注入，并通过 `GET /api/v1/core` 读取。

输出目录还包含：

- `SHA256SUMS`：整合归档的完整性摘要。
- `install.sh`：系统服务安装与升级入口。
- `docker-compose.yml`：固定版本的整合镜像安装入口。
- `release-notes.md`：从当前版本双语 changelog 生成的正式正文。

## 构建容器镜像

根目录 Dockerfile 复用同一 dist 入口，镜像只预装 Core、WebUI 和 Agent。签名密钥通过 BuildKit secret 传入；签名阶段重新构建，使本次密钥生效：

```fish
set version 0.1.0
set release_key ~/.config/alonabot/release.key
docker buildx build --platform linux/amd64,linux/arm64 --no-cache-filter build \
  --build-arg VERSION=$version --secret id=release_key,src=$release_key \
  --output type=oci,dest=/tmp/alonabot-$version.oci.tar \
  -t ghcr.io/alonabot/alonabot:$version .
```

需要覆盖组件下载前缀时，加 `--build-arg ASSET_BASE=https://example.com/releases/v0.1.0`。本机验收可用 `--load` 构建单平台镜像，再执行根目录 `docker compose up -d`；Compose 默认使用正式 GHCR tag，验证本地候选时将 `image` 改为相应的本地 tag。

镜像采用 scratch、UID/GID 10001、独立 `/data` 卷与 init，端口只发布到 loopback。使用需要 shell、Node 或 Python 的 MCP stdio 工具时，自行构建含相应运行时的镜像。

## 核对候选

[release-check](../scripts/release-check/README.md)读取实际归档、两组件的签名元数据、包内容和 build-info，并在新目录保存逐项结果。按改动影响选择目标和生命周期检查，复用仍适用的历史证据。

跨组件测试运行 `uv run --no-project python scripts/integration.py`。平台、崩溃恢复与安装更新分别使用[平台工具](../scripts/platform/README.md)和[场景工具](../scripts/scenarios/README.md)。交叉构建结果与实际运行结果分别记录。Windows 实跑已批准延期，原结果见[v0.1.0 候选说明](../changelog/v0.1.0/zh-CN.md)。

修改 `install.sh` 后执行 `sh -n deploy/install.sh`。真实 systemd/launchd 安装、升级、校验失败和卸载在授权、可恢复的环境验收。

## 发布组件与整合发行

每个组件仓库的同版本 Release 上传该组件平台包，以及固定名称的 `release.json`、`release.json.sig`、`release.pub`。各组件有独立仓库，清单名称互不冲突。来源、组件 ID、签名公钥与归档摘要必须一致；用 `alonabot install --source AlonaBot/<组件仓库>` 验证公开下载与后续更新。

正式发布按以下顺序执行：

1. 发布正式 SDK tag，更新并验证组件与插件模板的 SDK 引用。
2. 按固定源码和官方密钥构建组件；在对应仓库发布各组件的同版本签名资产。Adapter 与 help 使用各自的 `alonabot pack` 输出。
3. 构建、核对整合归档，将五平台归档、`SHA256SUMS`、`install.sh` 和 `docker-compose.yml` 上传 Core Release；正文从 `dist/release-notes.md` 读取：`gh release create v0.1.0 --notes-file dist/release-notes.md ...`。
4. 推送核对过的 linux/amd64、linux/arm64 镜像与多架构清单，确认版本 tag、latest 和 Compose 版本一致。
5. 正式资产可用后更新 README 的默认安装入口。

当前仍为候选；正式 SDK tag、官方签名、GitHub Release 与 GHCR 发布由维护者执行。发布失败时保存完整错误、产物和当前线上版本，修复受影响步骤后继续。
