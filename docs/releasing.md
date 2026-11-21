# 发行维护

Core 发行与整合镜像保留在主仓。公开五平台归档只包含 Core；Docker Compose 镜像包含 Core、WebUI 与 Agent。Telegram、OneBot 11 Adapter 和 help 插件在各自仓库发布，用户按需安装。安装步骤见[部署指南](deployment.md)，组件作者的发布流程见[插件指南](plugin-guide.md#9-发布)。

## 固定源码与发行说明

1. 选择已审查的 Core 提交和发行版本，确认相应检查通过。
2. 在 `components.lock.json` 固定五个官方组件的 Go module 精确版本。整合构建读取其中的 Agent 和 WebUI；跨组件测试还使用其他条目。组件版本来自各自 `plugin.yaml`，可独立于 Core 版本升级。
3. 将用户可见变更整理到 `changelog/v<版本>/en.md` 与 `zh-CN.md`，按[更新日志格式](../changelog/README.zh-CN.md)写入单版本标题和实际发行日期，执行 `go generate ./internal/core/manage` 更新 Core 管理 API 内嵌的双语原文。保留真实平台结果、资源实测失败和已批准的延期。日常条目放在 `changelog/unreleased/`。
4. 先将组件 SDK 固定到已验证的 Core 精确提交，待组件发布后再固定整合锁和 Core tag，避免相互引用尚未发布的版本。插件模板随后更新至正式 SDK tag。

生成的 Release 正文英文在前、简体中文在后，自动去掉 HTML 编辑注释：

```fish
go run ./scripts/release-notes --version 0.1.0 > /tmp/alonabot-release-notes.md
```

## 构建 Core 与内部整合归档

在主仓根目录执行：

```fish
go run ./scripts/dist --version 0.1.0 \
  --public-key .github/release.pub --out ./dist
```

构建需要 Go 和网络。默认只构建当前 checkout 的 Core。镜像使用的内部整合归档显式添加 `--bundle`，按 `components.lock.json` 的正式版本下载 Agent 与 WebUI 的公开 Release，先用可信公钥验签，再核对来源、ID、版本、平台、包长度与清单摘要。签名清单与组件包逐字收录到整合归档；前端在 WebUI 自身发布时构建。默认目标为 `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`，用 `--platforms linux/amd64,linux/arm64` 选择子集。

`--public-key` 指向独立信任的标准 base64 Ed25519 公钥；正式发行使用 [`.github/release.pub`](../.github/release.pub)。整合构建只需要公钥。组件仓库的 `pack --key` 接收维护者保存的 64 字节私钥，其本机备份权限设为 0600。

每个平台生成 `alonabot-<version>-<os>-<arch>.tar.gz`，Windows 为 `.zip`；默认归档仅含 Core 可执行文件、MIT 许可证与记录源码版本的 `source.json`。`--bundle` 额外收录 `packages/<组件ID>/`，每个组件目录含本平台安装包和完整的 `release.json`、`release.json.sig`、`release.pub`，源码记录列出实际整合的 Agent 与 WebUI。WebUI 包包含后台进程和同一份前端产物。

整合模式的 `<输出目录>/packages/<组件ID>/` 汇集所选平台包及三份元数据。默认下载 URL 指向该组件仓库的 `releases/download/v<组件版本>`。Core 版本由 `--version` 注入，并通过 `GET /api/v1/core` 读取。

纯 Core 归档可以先构建；内部整合归档与 Compose 镜像须在锁定的 Agent、WebUI 平台包及签名元数据公开后构建。商店正式验收使用匿名可读的公开索引和已发布的 WebUI；商店接口与收录流程见 [ADMIN-API §25](../spec/ADMIN-API.md#25-官方-webui-后台)。

公开 Release 仅上传五个平台归档；源码 zip 与 tar.gz 由 GitHub 自动提供。安装脚本从主仓 `main/deploy/install.sh` 下载，Compose 从对应 tag 的根目录下载，官方公钥保留在 `.github/release.pub`。安装器通过 GitHub Release API 的资产摘要校验官方归档。

构建和验收输出目录还包含以下辅助文件，保留在 Actions artifact 中：

- `SHA256SUMS`：该输出目录中归档的完整性摘要。
- `install.sh`：系统服务安装与升级入口。
- `docker-compose.yml`：固定版本的整合镜像安装入口。
- `release-notes.md`：从当前版本双语 changelog 生成的正式正文。

## 构建容器镜像

发布工作流将五平台 Core 归档写入 `dist/core`，将 Linux amd64/arm64 内部整合归档写入 `dist/bundle`，分别检查完整性和首次启动。公开 Release 只上传 `dist/core` 中的五份归档。两个原生架构 runner 从 `dist/bundle` 的对应归档构建 `published` target。源码构建仍可通过根 Dockerfile 的默认 target 调用 dist：

```fish
set version 0.1.0
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=$version \
  --output type=oci,dest=/tmp/alonabot-$version.oci.tar \
  -t ghcr.io/alonabot/alonabot:$version .
```

本机验收可用 `--load` 构建单平台镜像。发行工作流用 `scripts/release-smoke.py --compose docker-compose.yml --image <候选镜像>` 创建独立项目和数据卷，完成登录、默认组件、凭证持久化及停止验收后清理该项目。

镜像采用 scratch、UID/GID 10001、独立 `/data` 卷与 init，端口只发布到 loopback。使用需要 shell、Node 或 Python 的 MCP stdio 工具时，自行构建含相应运行时的镜像。

## 核对候选

[release-check](../scripts/release-check/README.md)默认核对 Core 归档的完整清单、来源、目标平台与 build-info；添加 `--bundle` 后同时核对两组件的签名元数据、包内容和 build-info，并在新目录保存逐项结果。按改动影响选择目标和生命周期检查，复用仍适用的历史证据。

跨组件测试运行 `uv run --no-project python scripts/integration.py`。平台、崩溃恢复与安装更新分别使用[平台工具](../scripts/platform/README.md)和[场景工具](../scripts/scenarios/README.md)。交叉构建结果与实际运行结果分别记录。Windows 实跑已批准延期，原结果见[v0.1.0 发布说明](../changelog/v0.1.0/zh-CN.md)。

修改 `install.sh` 后执行 `sh -n deploy/install.sh`。真实 systemd/launchd 安装、升级、校验失败和卸载在授权、可恢复的环境验收。

## 发布组件与整合发行

每个组件仓库的同版本 Release 上传该组件平台包，以及固定名称的 `release.json`、`release.json.sig`、`release.pub`。各组件有独立仓库，清单名称互不冲突。来源、组件 ID、签名公钥与归档摘要必须一致；用 `alonabot install --source AlonaBot/<组件仓库>` 验证公开下载与后续更新。

正式发布按以下顺序执行：

1. 将组件 SDK 固定到已审查的 Core 精确提交；验证并合并五个组件的源码 PR。
2. 在五组件仓库创建 `v<版本>` tag，触发各自 Release 工作流。工作流从固定源码构建五平台包，以官方密钥签名，正文读取 `changelog/<版本>/en-US.md` 与 `zh-CN.md`。
3. 在主仓 `components.lock.json` 固定已发布组件版本，验证并合并后创建 Core `v<版本>` tag。Release 工作流构建五平台 Core 归档和双 Linux 平台内部整合归档，核对归档、组件签名及许可证并验收两种首次启动，再创建只包含 Core 归档的 Release 草稿。正文读取 `dist/core/release-notes.md`。
4. 原生 amd64、arm64 runner 各自解压本次内部整合归档，通过 Dockerfile 的 `published` target 构建镜像并用发行 Compose 验收登录、默认组件、卷内凭证持久化和退出；通过后推送平台镜像并合并版本 tag 与 `latest`。
5. 首次发布时，先在组织 Settings → Packages → Package creation 中允许 Public，再将 GHCR 包的 visibility 设置为 public。工作流匿名拉取两架构并验收公开 Compose 后公开 Release；随后更新 README、同步 dev，并将插件模板切到正式 SDK tag。

官方公钥保存在 [`.github/release.pub`](../.github/release.pub)，五组件 Release 工作流通过 `ALONABOT_RELEASE_KEY` secret 读取同一发布者私钥；本机备份权限为 0600。工作流只记录公钥和签名产物。

发布失败时保留草稿、完整错误、产物与当前线上版本，修复受影响步骤后继续。`workflow_dispatch` 的 `tag` 指向已存在 tag，可恢复尚未完成的发行；常规发行保留已公开资产与 tag。2026-11-21 经维护者批准，v0.1.0 原地将五个整合归档替换为纯 Core 归档：复用原 Core 二进制，保留原 tag 和整合镜像，重新生成归档摘要与发行正文。旧归档摘要随之失效，用户需重新下载并按 Release API 摘要校验。该次重包手动消费原已验证产物；旧 tag 的工作流源码保留原状，新流程随之后的源码 tag 使用。
