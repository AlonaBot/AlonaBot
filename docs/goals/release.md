# goal 语句：集成与发布

本文件收录「集成与发布」章节的 goal 语句，切片表在[集成与发布实施计划](../../spec/plan/release.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-73 | T-14.3 | 发行归档 | T-03.10、T-09.3、T-12.7、T-14.1、T-14.2、T-15.2 | 等待前置 |
| goal-74 | T-14.4 | 首次运行与预装 | T-09.6、T-14.3 | 等待前置 |
| goal-77 | T-14.5 | 容器镜像 | T-07.6、T-14.3 | 等待前置 |
| goal-85 | T-14.7 | install.sh | T-14.3 | 等待前置 |
| goal-75 | T-14.6 | 发行包与镜像端到端验收 | T-09.11、T-14.4、T-14.5 | 等待前置 |
| goal-46 | T-10.1 | 证据汇总与跨组件联验 | T-02.3、T-03.11、T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.6、T-14.7、T-15.3、T-16.2 | 等待前置 |
| goal-47 | T-10.2 | 五平台构建与运行 | T-10.1 | 等待前置 |
| goal-48 | T-10.3 | 资源比较报告与常驻内存预算核对 | T-10.1 | 等待前置 |
| goal-49 | T-10.4 | 发布资产与说明 | T-10.2、T-10.3 | 等待前置 |

## 第 18 波

### goal-73 · T-14.3 发行归档

状态：等待前置（T-03.10、T-09.3、T-12.7、T-14.1、T-14.2、T-15.2）。语句中 goal-38 提供的命令（alonabot pack 与 --key）与各官方插件、Telegram 与 OneBot 11 Adapter 的目录只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-73（切片 T-14.3：发行归档）。goal 目录 goal-73，分支 goal-73，worktree .worktrees/goal-73，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：scripts/dist/ 为每个目标平台构建 Core 与全部官方组件（Telegram Adapter、OneBot 11 Adapter、Agent、WebUI、help、welcome、rss；WebUI 先在 web/admin 运行 npm ci 与 npm run build，再按 any 平台打包一次），用 alonabot pack 生成安装包与外部 release.json，以官方发布者私钥签名后放入 packages/，再打成 zip（Windows）或 tar.gz，并生成列出全部发行归档 SHA256 的 SHA256SUMS；发行版 Core 在构建时嵌入官方发布者公钥。

依据：PLAN T-14 步骤 3；CONTRACTS §15.2、§9.2；docs/development.md（交叉编译）；TEST-MATRIX V-66 的归档分项。

范围（只新增或修改这些）：
- 新增 scripts/dist/（Go 程序，go run ./scripts/dist 执行）
- 新增 internal/core/trust/（官方发布者公钥；构建时通过 -ldflags -X 注入，开发构建为空）
- internal/core/packages 或 manage：安装时把 trust 中的官方公钥视为已信任来源（只加调用点）
- docs/development.md：增加「发行构建」一节（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：scripts/dist -> contracts/release（ARCHITECTURE 已登记，用于自检签名）；其余只经 os/exec 调用 go build 与 alonabot 二进制。

接口：
- go run ./scripts/dist --version VERSION --key FILE --out DIR [--platforms linux/amd64,...]：缺省为五个目标平台；--key 为官方 Ed25519 私钥文件（不进入仓库与产物）。
- internal/core/trust：var officialKey string（base64 公钥，由 -ldflags "-X github.com/AlonaBot/AlonaBot/internal/core/trust.officialKey=..." 注入）；func Official() (ed25519.PublicKey, bool) // 未注入时 false

行为：
- 每个平台：CGO_ENABLED=0、-trimpath、-ldflags 注入版本与官方公钥，构建 alonabot；用该版本的 alonabot pack --goos --goarch --key 为每个官方包生成安装包、release.json 与 .sig；产物放入 alonabot-<version>-<os>-<arch>/（alonabot 可执行文件与 packages/），Windows 打 zip，其他打 tar.gz（文件按名称排序、时间戳固定，可重复）。
- 自检：每个包用 contracts/release.Verify 与公钥验证一次；失败则中止。把各归档的 SHA256 写入 SHA256SUMS（`sha256sum` 的格式：摘要、两个空格、文件名）。
- 官方公钥：发行版 Core 安装时把它视为已信任来源（预装包不逐个确认，签名与摘要照常校验）；开发构建没有内置公钥，行为不变。
- 私钥只从 --key 读取，不写日志；docs/development.md 的「发行构建」写明命令、所需工具与私钥保管要求。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/dist/dist_test.go：用临时 Ed25519 密钥只为当前平台运行一次（--platforms 当前平台），检查归档结构（alonabot 与 packages/ 下每个官方包的三个文件）、每个 release.json 验签通过、两次运行的归档字节相同、构建出的 alonabot 内置的公钥等于测试公钥（经一个只读的 version 输出或测试钩子）。
- 本地自测（不提交，PR 中写明结果）：五个平台的完整构建（交叉编译，只验证产物生成，不在其他平台运行）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./scripts/... ./internal/...；go build ./...。

不做：首次运行的预装（T-14.4）；容器镜像（T-14.5）；发布到 GitHub Releases（T-10.4）。
```

## 第 19 波

### goal-74 · T-14.4 首次运行与预装

状态：等待前置（T-09.6、T-14.3）。语句中 goal-39、goal-41、goal-73 提供的符号（manage.Service、serve、trust.Official、发行归档结构）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：首次 token 由 goal-97 实现（终端显示或写入 `admin_token`）；`core_setting` 表不再由 goal-53 建立（D-82），preinstalled 标记的存放位置由审查者改写；`core_config.json` 中写了实例的 Adapter 自动启用（D-82）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-74（切片 T-14.4：首次运行与预装）。goal 目录 goal-74，分支 goal-74，worktree .worktrees/goal-74，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：数据目录首次初始化时，按普通安装流程安装发行包 packages/ 中的全部包，只启用 help 与官方 WebUI；Agent 只安装不启用（D-50、D-62）；Telegram 与 OneBot 11 Adapter 只安装，管理员为它添加第一个实例时才启用；其他预装插件只安装不启用。之后不再自动安装，已卸载的预装包不重新安装；新版本发行包中版本更高的内置包作为可用更新列出，由管理员确认后按 update 流程更新。

依据：PLAN T-14 步骤 4；CONTRACTS §15.1、§15.2；TEST-MATRIX V-66。

范围（只新增或修改这些）：
- 根包 assembly.go：serve 首次初始化时的预装步骤（新增 preinstall.go，只在 Serve 中加调用点）
- internal/core/state/：记录「已完成首次初始化」的设置项（沿用 core_setting）
- internal/core/manage/：List 返回可用更新；CreateInstance 在实例所属的 Adapter 未启用时先启用它
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- 首次初始化：serve 取得 core.lock 后，若 core_setting 中没有 preinstalled 标记：查找 alonabot 可执行文件同目录下的 packages/（发行归档结构）；没有该目录时跳过（开发构建）。对其中每个包按 Install 的正常流程校验与安装；trust.Official 的公钥视为已信任来源，不逐个确认；签名或摘要校验失败的包不安装并记 ERROR，继续处理其余包。安装后只启用 help（按 manifest 的 id 识别）与 role 为 webui 的官方 WebUI，其余不启用。全部处理完后写入 preinstalled 标记（即使有包失败）。
- 首次管理 token：沿用 goal-39 的 serve 规则（交互终端下生成并显示一次，非终端拒绝启动）；预装在 token 初始化之后执行。
- 管理员之后可经管理 API 或 CLI 启用 Agent；Agent 在没有配置 Provider 时也能启用（/ask 回复提示）。
- 之后的启动不再自动安装；已卸载（uninstalled）的预装包不重新安装。
- 可用更新：List 时比较 packages/ 中每个包的版本与已安装版本，更高时在 Package 中标记 available_update（版本号），不自动更新；管理员经 update 选择该包。
- Telegram：CreateInstance 发现 Adapter 包未启用时先启用它，再创建实例。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/firstrun_e2e_test.go（//go:build unix）：用 goal-73 的 dist 工具以临时密钥为当前平台构建发行归档并解压到临时目录，运行其中的 alonabot run（干净 data_dir，测试钩子模拟交互终端）：
  - 全部预装包已安装；只有 help 已启用并有进程；Agent、记忆、welcome、rss、Telegram 与 OneBot 11 Adapter 未启用，没有对应进程。
  - 把其中一个包的签名改坏：该包未安装并有 ERROR，其余正常。
  - 卸载 rss 后重启 serve：不重新安装。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：没有 packages/ 时跳过；preinstalled 标记在部分失败时仍写入；非终端运行时拒绝启动并提示 admin-token init；用版本号更高的发行归档替换 packages/ 后重启：list 显示可用更新，没有自动更新；创建 Telegram 实例时 Adapter 被启用。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：容器镜像（T-14.5）；Core 自身的升级（替换可执行文件，由管理员操作）。
```

### goal-77 · T-14.5 容器镜像

状态：等待前置（T-07.6、T-14.3）。语句中 goal-25、goal-73 提供的符号（settings 的 listen 校验、scripts/dist）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：首次启动不再需要单独初始化，token 写入 `/data/admin_token`，用 `docker compose exec alonabot alonabot admin-token show` 读出（D-82）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-77（切片 T-14.5：容器镜像）。goal 目录 goal-77，分支 goal-77，worktree .worktrees/goal-77，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：仓库根目录的 Dockerfile 用多阶段构建产生 linux/amd64 与 linux/arm64 镜像，deploy/compose.yaml 是官方整合包的 compose 入口：Go 构建阶段以 CGO_ENABLED=0 构建 Core 与全部官方包，最终镜像只含 alonabot、packages/ 与 CA 证书，以 UID 10001 运行 alonabot run --data-dir /data，/data 为卷，设置 ALONABOT_CONTAINER=1。Core 只在该环境变量下允许 0.0.0.0 的 listen 搭配本机 http 的 Base URL，并在启动时输出 WARN。deploy/compose.yaml 给出卷、init: true、127.0.0.1:7650:7650 与首次 admin-token init 的用法。

依据：PLAN T-14 步骤 5；CONTRACTS §15.3、§15.1、§7.4；DECISIONS D-36；TEST-MATRIX V-70 的镜像分项。

范围（只新增或修改这些）：
- 新增 Dockerfile、.dockerignore、deploy/compose.yaml
- internal/core/settings/：ALONABOT_CONTAINER=1 时 listen 缺省为 0.0.0.0:7650，并放宽「非 loopback 必须 https 或可信反代」的校验（只在该变量下），启动时输出 WARN "listening on all interfaces inside a container; publish the port to the host loopback only"
- docs/development.md：增加「容器镜像」一节（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：无。

行为：
- Dockerfile：构建阶段使用 go.mod 锁定的 Go 版本的官方镜像（标签写明版本与 digest）；以 scripts/dist 的同一逻辑为目标平台构建 alonabot 与官方包（签名密钥以 BuildKit secret 传入，不进入镜像层；没有密钥时构建失败并提示）；最终阶段为 scratch 加 CA 证书（从构建阶段复制 /etc/ssl/certs/ca-certificates.crt），USER 10001:10001，VOLUME /data，ENV ALONABOT_CONTAINER=1，ENTRYPOINT ["/alonabot"]，CMD ["serve", "--data-dir", "/data"]。镜像中没有 shell。
- compose.yaml：image 固定为对应版本（发布流程把 :<version> 写入随 Release 上传的副本）、init: true、volumes（命名卷到 /data）、ports "127.0.0.1:7650:7650"，注释说明首次运行前执行 docker compose run --rm alonabot admin-token init，再 docker compose up -d。
- settings：没有 ALONABOT_CONTAINER 时行为不变（0.0.0.0 搭配 http Base URL 仍被拒绝）。
- docs：写明构建命令（docker buildx build --platform linux/amd64,linux/arm64 --secret id=release_key,src=...）、首次初始化、端口只发布到宿主本机、远程访问用 SSH 隧道或 HTTPS 反代、MCP stdio 需要 shell 或其他运行时时须自建镜像。

测试（按编码规范「测试提交规则」）：
- 提交：无。
- 本地自测（不提交，PR 中写明结果）：在有 Docker 的机器上 buildx 构建两个平台（写明 Docker 版本）；干净卷无 TTY 启动被拒；-it admin-token init 后正常启动并能访问 http://127.0.0.1:7650；容器内进程的 UID 为 10001；镜像文件列表只有 alonabot、packages/ 与 CA 证书；出站 HTTPS 可用；docker stop 后插件进程全部退出。没有 Docker 环境时以上各项写「未运行」；internal/core/settings 的本地测试：有与没有 ALONABOT_CONTAINER 时 listen 缺省值与校验的差异，以及 WARN。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/...；go build ./...。

不做：推送镜像到 ghcr.io（T-10.4）；Kubernetes 清单。
```

### goal-85 · T-14.7 install.sh

状态：等待前置（T-14.3）。语句只写了目标、依据和要点。依据 D-64 新增；T-14.3 合并后由审查者按实际的发行归档与 SHA256SUMS 写出范围和测试段。2026-11-04 按 D-77～D-85 修订计划：没有终端时 Core 不再拒绝启动，脚本跳过初始化并在启动服务后输出 `admin-token show` 的命令（D-82）。前置合并后由审查者改写。

目标：deploy/install.sh 按 CONTRACTS §15.6 在 Linux（systemd，root）与 macOS（launchd，当前用户）上安装 Core 与官方组件：下载并按 SHA256SUMS 校验发行归档，校验失败不改动已有安装；替换程序目录；尚无管理凭证时经 /dev/tty 运行 admin-token init，没有终端时不启动服务并输出初始化命令；写入 systemd 单元（系统用户 alonabot）或用户 LaunchAgent 并启动；再次运行即升级，`--uninstall` 保留数据目录；不支持的系统在改动前退出。Release 流程上传 install.sh 与 SHA256SUMS。

依据：PLAN T-14 步骤 7、§3.1 的 T-14.7 行；CONTRACTS §15.1、§15.2、§15.6；DECISIONS D-64；TEST-MATRIX V-84。

要点：
- 只依赖 POSIX sh、curl、tar 与 sha256sum/shasum，不引入新工具；用 shellcheck 检查，未安装时记为未运行。
- 下载地址的基址可由环境变量覆盖，供本地 Release 模拟服务测试；测试与真实安装都不写入开发机的系统目录之外的位置。
- V-84 需要真实的 systemd Linux 与 macOS；容器内的 systemd 与模拟环境的结果单独记录，不算真实结果。

## 第 20 波

### goal-75 · T-14.6 发行包与镜像端到端验收

状态：等待前置（T-09.11、T-14.4、T-14.5）。语句中 goal-43、goal-44、goal-73、goal-74、goal-77 提供的产物与入口（发行归档、容器镜像、WebUI 页面、管理 API）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按 D-77～D-85 修订计划：首次 token 的取得改为终端显示或 `admin-token show`（D-82）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-75（切片 T-14.6：发行包与镜像的端到端验收）。goal 目录 goal-75，分支 goal-75，worktree .worktrees/goal-75，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在干净的数据目录上，分别从发行归档与容器镜像依次完成：首次运行、登录 WebUI、添加 Telegram 实例（本地 Bot API 模拟服务）、启用 Agent 并配置 Provider（本地模拟服务）、/ask 对话、/help、启用 rss 并收到推送；在 TEST-MATRIX §3.2 记录 T-14 证据，真实 Telegram 与真实 Provider 没有凭证时记为未运行。

依据：PLAN T-14 步骤 6 与「停止和失败条件」；CONTRACTS §15；TEST-MATRIX V-66、V-67、V-70。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t14_e2e_test.go（//go:build unix；发行归档部分）
- 新增 deploy/e2e/（容器验收：Go 测试辅助，经 os/exec 调用 docker，运行镜像、模拟服务与断言）
- spec/TEST-MATRIX.md §3.2：追加 T-14 的证据行（本切片获授权修改这一处）
- 上述包的测试文件
允许的新跨模块边：无。

验收步骤（发行归档与容器镜像各跑一遍；每一步记录命令与结果）：
1) 首次运行：干净 data_dir；终端下初始化 token（容器用 -it admin-token init）；预装包状态符合 goal-74。
2) 登录：经管理 API 登录（等同页面登录），状态接口显示各包状态。
3) 添加 Telegram 实例：api_base_url 指向本地 Bot API 模拟服务（容器内经宿主地址访问），实例上线（bot.online）。
4) 配置 Provider：经配置 API 为 Agent 添加指向本地模拟 Provider 的条目（api_key 为 secret），reload 生效。
5) /ask：模拟服务中的用户发送 /ask，收到先发送再编辑的回复，最终内容正确。
6) /help 与 /whoami：列出可用命令；/whoami 返回实例与用户 ID；把该用户加入 Core Admin。
7) 启用 rss：Admin /rss add 一个测试 HTTP 源，授予发信许可，源新增条目后收到推送。测试源必须使用 rss 地址检查允许的地址（非 loopback、非私网）；做不到时说明实际做法（例如测试钩子），并在证据中标明。
8) 停止：Core 与全部插件进程退出（容器 docker stop 后进程全部消失）。
- V-70：镜像文件只含 alonabot、packages/ 与 CA 证书，UID 10001；无 TTY 时拒绝启动；端口只发布到宿主本机；出站 HTTPS 可用；没有 ALONABOT_CONTAINER 时 0.0.0.0 搭配 http Base URL 被拒绝。
- 证据：TEST-MATRIX §3.2 追加 T-14 一行（平台、Docker 版本、命令、结果）；真实 Telegram、真实 Provider 与没有运行的平台写「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t14_e2e_test.go；deploy/e2e。
- 本地自测（不提交，PR 中写明结果）：在有 Docker 的机器上运行容器部分；没有 Docker 时写「未运行」。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：五平台运行（T-10.2）；发布（T-10.4）；新的产品功能（只做验收与必要修复，修复逐条说明）。
```

## 第 21 波

### goal-46 · T-10.1 证据汇总与跨组件联验

状态：等待前置（T-03.11、T-05.5、T-08.9、T-09.11、T-11.4、T-12.7、T-13.4、T-14.6、T-14.7、T-15.3、T-16.2）。本切片只汇总与联验，不引入新接口。2026-11-04 按 D-77～D-85 修订计划：首次 token 的取得见 D-82；插件页面经 WebUI 的 iframe 与桥访问，不再使用票据（D-84）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-46（切片 T-10.1：证据汇总与跨组件联验）。goal 目录 goal-46，分支 goal-46，worktree .worktrees/goal-46，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：汇总 V-01～V-76 与 V-82～V-87 的证据，逐项标记已通过、失败、未运行或待联验；补齐组件之间的联验（首次管理登录 → 安装启用 → 配置/权限 → 真实消息 → 插件 HTTP → 重载/升级/回退，TCP 与 stdio）；用不同的真实插件二进制覆盖各能力组合；执行协议、权限、资源边界与完整恢复的故障注入，核对数据库内容与权威指针。

依据：PLAN T-10 步骤 1～4 与「停止和失败条件」；TEST-MATRIX 全部；CONTRACTS §12.4。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t10_e2e_test.go（//go:build unix）与必要的 testdata 插件
- spec/TEST-MATRIX.md：§3 的每个 V 条目更新为最新状态，新增「发布前证据汇总」一节（本切片获授权修改 TEST-MATRIX；其他 spec 文件只在发现契约缺陷时按 spec/README 的流程修改，并在 PR 中说明）
- 发现的缺陷：在对应包中修复，每项修复在 PR 中写明 V 编号、复现、原因与验证
允许的新跨模块边：无。

联验矩阵（各为一个子测试，TCP 与 stdio 各一遍）：
1) 组件链路：admin-token init → 登录 → 安装并启用 echo、configdb、http 示例与官方插件 → 配置与权限 → Telegram 模拟服务的真实消息 → 插件 HTTP 页面（经票据）→ reload → update → rollback。
2) 能力组合（每个能力使用不同的插件二进制，不用同一个 worker 的多份进程代替）：命令、文本匹配、多轮等待、主动 Send、Service、Event Bus、Scheduler、HTTP、GORM（含迁移），以及 Agent 与 WebUI 组件；既有无配置无数据库的插件，也有实际写入与迁移的插件。
3) 边界：伪造与过期 Session、命令遮蔽、无 Actor、实例许可、Hook 递归、媒体租约、满准入、Cancel 顺序、超时仍运行的 job；错误回执与已确认的外部结果一致。
4) 恢复：停止与持锁、SQLite WAL 备份、候选 OnStart 写入后失败、PREPARED 与提交前后 Core 退出、不可变 secret、external 插件的维护许可与 resume、显式 rollback 的新代次；核对数据库内容与状态库的 Active 指针，不能只看日志里的「恢复成功」。
- 汇总：TEST-MATRIX 中每个 V 条目写明测试位置、命令、平台与结果；失败的列为阻塞项；没有运行条件的（真实 Telegram、真实 Provider、非 darwin 平台等）写「未运行」与原因。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t10_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：完整测试套件在 -race 下运行一次的总耗时与结果。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；TEST-MATRIX 中没有「待联验」的条目。

不做：五平台运行（T-10.2）；资源报告（T-10.3）；发布（T-10.4）。
```

## 第 22 波

### goal-47 · T-10.2 五平台构建与运行

状态：等待前置（T-10.1）。需要真实目标平台；没有运行环境的组合保持未运行。

```text
/goal 使用 $alonabot-execute 推进 goal-47（切片 T-10.2：五平台构建与运行）。goal 目录 goal-47，分支 goal-47，worktree .worktrees/goal-47，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64 五个组合上分别构建并运行 Core 与插件，覆盖启动、停止、至少 100 次启停、异常退出、合作型子进程树与端口清理；记录 CGO 插件的实际支持范围。

依据：PLAN T-10 步骤 5；CONTRACTS §7.1（平台与进程树）、§12.2；TEST-MATRIX V-32、V-34。

范围（只新增或修改这些）：
- 新增 scripts/platform/（可在各平台运行的验收程序：Go 程序，调用构建好的 alonabot 与测试插件，输出结构化结果）
- 复用 goal-45 的 scripts/scenarios 在各平台运行
- spec/TEST-MATRIX.md：V-32、V-34 与各平台相关条目的证据（本切片获授权修改 TEST-MATRIX）
- 发现的平台缺陷：在对应包中修复（例如 Windows Job Object、路径与锁），每项在 PR 中写明平台、复现与验证
允许的新跨模块边：scripts/platform 只经 os/exec 调用构建产物，不 import 本 module 的包。

行为：
- 每个平台：用 scripts/dist 的构建产物（或在该平台本地构建）运行 scripts/platform：启动 serve、安装并启用 echo 与 http 示例、收发一条消息、访问一次插件 HTTP、停止；再循环 100 次启停（每次检查插件进程与监听端口都已释放）；强制结束 Core（Unix 为 SIGKILL，Windows 为 TerminateProcess）后重启，确认插件进程已退出且能恢复；启动一个会派生子进程的测试插件，停止后子进程消失（合作型进程树）。
- 结果写成 JSON（平台、Go 版本、OS 版本、各步骤的结果与耗时），附在 PR 中，摘要写入 TEST-MATRIX。
- CGO：用一个 CGO 测试插件在各平台尝试构建与运行，记录实际支持范围；不支持的平台标为不支持，不发布。
- 没有某个平台的运行环境时，该平台写「未运行」；交叉编译成功不能代替运行结果。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/platform（本身就是验收程序）；平台缺陷修复所需的集成测试。
- 运行记录：每个平台的 JSON 结果附在 PR 中。

完成条件：gofmt -l 无输出；go vet ./...（含 GOOS=windows、GOOS=linux 的 vet）；go test -race ./...（当前平台）；每个可用平台的 scripts/platform 运行结果。

不做：资源报告（T-10.3）；发布（T-10.4）。
```

### goal-48 · T-10.3 资源比较报告与常驻内存预算核对

状态：等待前置（T-10.1）。需要固定机器与等价负载；NoneBot2 侧的 Python 依赖在引入前须经维护者确认。

```text
/goal 使用 $alonabot-execute 推进 goal-48（切片 T-10.3：资源比较报告与常驻内存预算核对）。goal 目录 goal-48，分支 goal-48，worktree .worktrees/goal-48，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在一台固定机器上，按 TEST-MATRIX V-33 的四种负载（空闲、消息、数据库、HTTP）测量 Core 加插件及附属进程的总 RSS/私有内存、Linux PSS、CPU 与 P50/P95/P99 延迟，并用同一模拟服务与功能相同的插件测量 NoneBot2；再按 V-74 测量 Core、echo、rss 与 configdb 的 gormdb 变体空闲时的 PSS 并对照预算。输出原始数据与复现命令，不预设优胜结论。

依据：PLAN T-10 步骤 6；CONTRACTS §10；DECISIONS D-52；TEST-MATRIX V-33、V-74。

范围（只新增或修改这些）：
- 新增 bench/（负载生成器与采样程序：Go 程序）；bench/nonebot2/（NoneBot2 侧的等价插件与运行说明，Python 依赖用 uv 管理且只在该目录内）
- 新增 docs/benchmarks.md（报告：机器、版本、命令、原始数据位置、结果表，不写优胜结论）
- spec/TEST-MATRIX.md：V-33、V-74 的证据（本切片获授权修改这一处）
允许的新跨模块边：bench 只经 os/exec 与 HTTP 调用被测程序。NoneBot2 及其 Python 依赖只用于基准，不进入 Go module；在 bench/nonebot2 中引入前标阻塞，列出版本，等维护者确认。

行为：
- 负载（V-33）：空闲——Core 与 echo、configdb、HTTP 示例三个插件运行 10 分钟；消息——每秒 20 条文本、持续 10 分钟，echo 逐条回复；数据库——每秒 20 条命令，每条在 SQLite 中写入并读取一行，持续 10 分钟；HTTP——每秒 50 个请求访问插件路由，持续 10 分钟。两边都使用同一个本地 Bot API 模拟服务。
- 采样：每秒采样一次全部相关进程（Core、插件、Adapter 及其子进程；NoneBot2 的进程）的 RSS、私有内存（Linux 读 smaps_rollup 的 Private_*；macOS 写明使用的替代指标）、Linux PSS、CPU；延迟为模拟服务从发出 update 到收到回复的时间（消息与数据库负载），HTTP 为请求往返时间。
- 输出：每次运行的原始 CSV 与汇总 JSON，复现命令写入 docs/benchmarks.md；报告只给数值与条件，不写结论。
- V-74：linux/amd64，各进程 Activated 后空闲 2 分钟，读 smaps_rollup 的 Pss，取 3 轮中位数；超出预算时在报告中列出超出项与可见原因（例如新增依赖、init 分配），不自行调整预算。
- TEST-MATRIX 记录机器、OS、版本与结果文件位置；没有在 Linux 上运行时 PSS 写「未运行」。

测试（按编码规范「测试提交规则」）：
- 提交：基准程序本身是交付物，不另提交测试；负载生成器的本地自测在 PR 中写明。
- 运行记录：原始数据附在 PR 中（是否提交到 bench/results/ 由维护者决定）。

完成条件：gofmt -l 无输出；go vet ./...；go build ./...；一次完整基准运行的结果。

不做：性能优化（发现问题时另行处理）；修改 V-74 的预算（需维护者在 DECISIONS 记录）；为吞吐与延迟设定通过阈值（契约不设）。
```

## 第 23 波

### goal-49 · T-10.4 发布资产与说明

状态：等待前置（T-10.2、T-10.3）。正式签名与发布由维护者在其控制的环境中执行。

```text
/goal 使用 $alonabot-execute 推进 goal-49（切片 T-10.4：发布资产与说明）。goal 目录 goal-49，分支 goal-49，worktree .worktrees/goal-49，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：生成五平台候选资产（含官方 Telegram 与 OneBot 11 Adapter、Agent、记忆与预装插件）、T-14 的发行归档与容器镜像、外部 release.json 与签名；用干净的测试 data_dir 从这些真实资产完成安装、更新与卸载重装，核对来源、长度、摘要、平台与 Schema/二进制一致性，确保安装不依赖开发机上未打包的文件；关闭全部阻塞用例，写出发布说明。签名私钥只在维护者控制的环境中使用；发布到 GitHub Release 与推送镜像由维护者执行。

依据：PLAN T-10 步骤 7、8；CONTRACTS §9.2、§15.2、§15.3；TEST-MATRIX V-28、V-41 的发布复验。

范围（只新增或修改这些）：
- 新增 docs/release-notes/v0.1.0.md（版本、平台、依赖、已知限制、未运行项、恢复步骤、升级与回退说明）
- 新增 scripts/release-check/（用候选资产在干净 data_dir 完成安装、更新、卸载重装与一致性核对的 Go 程序）
- spec/TEST-MATRIX.md：V-28、V-41 的发布复验证据，以及「发布前证据汇总」的最终状态（本切片获授权修改 TEST-MATRIX）
- README.md：安装与快速开始指向发布资产（本切片获授权修改 README 的这一节）
允许的新跨模块边：scripts/release-check -> contracts/release（ARCHITECTURE 已登记，与 scripts/dist 同类）；其余只经 os/exec 调用候选资产。

行为：
- 候选资产：执行者用测试密钥运行 scripts/dist 与镜像构建，得到与正式发布结构相同的候选资产；正式签名由维护者用官方私钥重新运行同一命令完成。
- release-check：在干净 data_dir 中只使用候选资产（不访问源码树）：首次运行预装、安装一个第三方测试包、用新版本资产 update、卸载并同来源重装；对每个包核对 release.json 的来源、长度、SHA256、平台与包内 Schema、manifest、二进制的一致性；签名错误、来源不符的资产被拒绝（V-28、V-41 复验）。
- 发布说明：写明支持的平台与实际运行结果（来自 T-10.2）、资源报告位置（T-10.3）、全部「未运行」项与原因、已知限制（Native 插件是可信代码、库模式无隔离、插件自管的外部数据不承诺自动回退等）、备份与恢复步骤、升级与回退方法。
- 阻塞用例：TEST-MATRIX 中标为失败的条目必须全部修复或经维护者同意延后（在发布说明中列出）。
- 发布：维护者在其环境中用 scripts/dist 构建并签名，把发行归档、各安装包及其 release.json/.sig 发布到本仓库的 GitHub Release，把镜像推送到 ghcr.io/alonabot/alonabot（:<version> 与 :latest）；执行者不持有私钥，不执行发布。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/release-check（验收程序）。
- 运行记录：release-check 对候选资产的运行结果附在 PR 中。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；release-check 对候选资产通过；TEST-MATRIX 没有未处理的失败条目。

不做：执行正式签名与发布（维护者负责）；v0.1 之后的路线图。
```
