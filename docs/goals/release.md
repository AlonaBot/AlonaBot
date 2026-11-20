# goal 语句：集成与发布

本文件收录「集成与发布」章节的 goal 语句，切片表在[集成与发布实施计划](../../spec/plan/release.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-73 | T-14.3 | 发行归档 | T-03.10、T-09.3、T-12.7、T-14.1 | 已完成（PR #228） |
| goal-74 | T-14.4 | 首次运行与预装 | T-09.6、T-14.3 | 已完成（PR #235） |
| goal-77 | T-14.5 | 容器镜像 | T-07.6、T-14.3 | 已完成（PR #231） |
| goal-85 | T-14.7 | install.sh | T-14.3 | 已完成（PR #236） |
| goal-75 | T-14.6 | 发行包与镜像端到端验收 | T-09.11、T-14.4、T-14.5 | 已完成（PR #275） |
| goal-46 | T-10.1 | 证据汇总与跨组件联验 | T-02.3、T-03.11、T-03.23、T-04.10、T-04.11、T-05.5、T-08.9、T-08.13、T-09.11、T-09.27、T-11.4、T-12.7、T-13.4、T-14.6、T-14.7、T-16.2 | 实施与证据审查中（PR #278） |
| goal-47 | T-10.2 | 五平台构建与运行 | T-10.1 | 独立实施中；最终验收等待前置 |
| goal-48 | T-10.3 | 现有资源实测与维护者决定归档 | T-10.1 | 等待文档收口（D-184） |
| goal-49 | T-10.4 | 发布资产与说明 | T-10.2、T-10.3 | 等待前置 |

## 第 19 波

### goal-73 · T-14.3 发行归档

状态：已完成（PR #228），合并提交 `72d51c2`。以下保留派发时语句。派发时：前置 T-03.10、T-09.3、T-12.7 与 T-14.1（PR #210，`456a4df`）均已合并；2026-11-17 按 D-171 缩小组件范围，并按 main `08cee69` 核对：`alonabot pack` 的参数为 `--dir`、`--goos`、`--goarch`、`--tags`、`--asset-url`（必填，绝对 https 地址）、`--key`（标准 base64 的 64 字节 Ed25519 私钥文件）与 `--out`，带 `--key` 时只为单个平台签名（`internal/pluginkit.Pack` 以 `ed25519.Sign` 对 release.json 原始字节签名，写出 base64 加换行的 release.json.sig 与 release.pub）；`contracts/release` 提供 `Verify(raw, sigFile []byte, pub ed25519.PublicKey) (*Release, error)` 与 `KeyID(pub) string`，`Release.PublisherKeyID` 须等于验签公钥的 KeyID。组件目录：Core `cmd/alonabot`，Telegram Adapter `adapters/telegram`（薄入口 `cmd/alonabot-adapter-telegram`），Agent `agent`，help `plugins/help`（`alonabot-plugin-help-official`），WebUI `web/admin`（当前只有静态文件，按 any 打包一次；带 Go 入口由 goal-156 实现后再按平台打包）。`internal/core/trust` 尚不存在；安装时的信任判断在 `internal/core/manage/install.go`（`trusted` 与「trust publisher」确认项）和 `candidates.go:checkPublisher`（更新时的发布者绑定）。

```text
/goal 使用 $alonabot-execute 推进 goal-73（切片 T-14.3：发行归档）。goal 目录 goal-73，分支 goal-73，worktree .worktrees/goal-73，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：scripts/dist/ 为每个目标平台构建 Core 与首个发行版的组件（D-171：Telegram Adapter、Agent、WebUI、help；WebUI 先在 web/admin 运行 npm ci 与 npm run build，再与其他组件一样按平台打包），用 alonabot pack 生成安装包与外部 release.json，以官方发布者私钥签名后放入 packages/，再打成 zip（Windows）或 tar.gz，并生成列出全部发行归档 SHA256 的 SHA256SUMS；发行版 Core 在构建时嵌入官方发布者公钥。

依据：PLAN T-14 步骤 3；CONTRACTS §15.2、§9.2；docs/development.md（交叉编译）；TEST-MATRIX V-66 的归档分项。

范围（只新增或修改这些）：
- 新增 scripts/dist/（Go 程序，go run ./scripts/dist 执行）
- 新增 internal/core/trust/（官方发布者公钥；构建时通过 -ldflags -X 注入，开发构建为空）
- internal/core/manage/install.go 与 candidates.go：发布者 KeyID 等于 `trust.Official()` 的 KeyID 时视为已信任来源，不加入「trust source」「trust publisher」确认项（只加调用点）
- docs/development.md：增加「发行构建」一节（本切片获授权修改这一节）
- 上述包的测试文件
允许的新跨模块边：scripts/dist -> contracts/release（ARCHITECTURE 已登记，用于自检签名）；其余只经 os/exec 调用 go build 与 alonabot 二进制。

接口：
- go run ./scripts/dist --version VERSION --key FILE --out DIR [--platforms linux/amd64,...] [--asset-base URL]：缺省为五个目标平台；--key 为官方 Ed25519 私钥文件（格式同 alonabot pack --key，不进入仓库与产物）；--asset-base 缺省为 `https://github.com/AlonaBot/AlonaBot/releases/download/v<VERSION>`，每个安装包的 --asset-url 为该前缀加归档名。
- internal/core/trust：var officialKey string（base64 公钥，由 -ldflags "-X github.com/AlonaBot/AlonaBot/internal/core/trust.officialKey=..." 注入）；func Official() (ed25519.PublicKey, bool) // 未注入时 false

行为：
- 每个平台：CGO_ENABLED=0、-trimpath、-ldflags 注入版本与官方公钥，构建 alonabot。
- 每个官方包：先为每个平台运行该版本的 alonabot pack --goos --goarch（不带 --key），把各平台 release.json 的 assets 合并为一份，把 `publisher_key_id` 设为 `release.KeyID(公钥)`，按 CONTRACTS §9.2 用 --key 的私钥对这份 release.json 的确切字节签名（`crypto/ed25519`，sig 文件格式与 `pluginkit.Pack` 相同），写出 release.json.sig 与 release.pub（D-169：pack 的 --target 由 T-13.6 在本切片之后实现）。官方 WebUI 带 Go 入口（D-162、goal-156），与其他组件一样按平台打包；只有静态文件的 WebUI 才按 any 打包一次。
- 产物放入 alonabot-<version>-<os>-<arch>/（alonabot 可执行文件与 packages/，packages/ 中是该平台的安装包与合并后的 release.json、.sig、.pub），Windows 打 zip，其他打 tar.gz（文件按名称排序、时间戳固定，可重复）。
- 自检：每个包用 contracts/release.Verify 与公钥验证一次；失败则中止。把各归档的 SHA256 写入 SHA256SUMS（`sha256sum` 的格式：摘要、两个空格、文件名）。
- 官方公钥：发行版 Core 安装时把它视为已信任来源（预装包不逐个确认，签名与摘要照常校验）；开发构建没有内置公钥，行为不变。
- 私钥只从 --key 读取，不写日志；docs/development.md 的「发行构建」写明命令、所需工具与私钥保管要求。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/dist/dist_test.go：用临时 Ed25519 密钥只为当前平台运行一次（--platforms 当前平台），检查归档结构（alonabot 与 packages/ 下每个官方包的三个文件）、每个 release.json 验签通过、两次运行的归档字节相同、构建出的 alonabot 内置的公钥等于测试公钥（经一个只读的 version 输出或测试钩子）。
- 本地自测（不提交，PR 中写明结果）：五个平台的完整构建（交叉编译，只验证产物生成，不在其他平台运行）。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./scripts/... ./internal/...；go build ./...。

不做：首次运行的预装（T-14.4）；容器镜像（T-14.5）；发布到 GitHub Releases（T-10.4）。
```

## 第 20 波

### goal-74 · T-14.4 首次运行与预装

状态：已完成（PR #235），合并提交 `815e53c`。以下保留派发时语句。派发时：前置 T-09.6（PR #179，合并提交 `8898251`）与 T-14.3（PR #228，合并提交 `72d51c2`）均已合并。按 main `25e9064` 核对：T-09.4（PR #163，`cb8ae3d`）的管理服务现为 `internal/core/manage`，提供 `New(opt Options) *Service`、`(*Service).Install(ctx context.Context, req InstallRequest) (Package, error)`、`Enable(ctx context.Context, id string) error`、`List(ctx context.Context) ([]Package, error)`、`Update(ctx context.Context, id string, req InstallRequest) (UpdateResult, error)` 与 `CreateInstance(ctx context.Context, adapterPackageID, instanceID string, raw json.RawMessage, confirm bool) error`；T-13.5（PR #200，`2e5ac2b`）已将装配移到 `internal/app`，入口为 `Start(ctx context.Context, o Options) (*Core, error)`，CLI 为 `alonabot run`。T-07.6（PR #100，`cb83eeb`）提供 `internal/core/adminauth.FirstToken(ctx context.Context, st *state.Store, dataDir string, interactive bool) (string, error)` 与 `FirstTokenFile = "admin_token"`，无终端时写文件；状态库没有 `core_setting`。T-14.3 提供 `internal/core/trust.Official() (ed25519.PublicKey, bool)`、`scripts/dist` 与 `internal/app.releaseVersion` 注入；发行归档的同名根目录含 Core 与 `packages/<组件 ID>/`，每个组件目录含安装包、`release.json`、`release.json.sig`、`release.pub`，首版只有 Telegram Adapter、Agent、WebUI 与 help（D-171）。

与同在第 20 波的 goal-141（T-13.6）共享 `packages/` 的定位和读取规则、`internal/core/manage/bundled.go`（本切片新增）及其 CLI 调用方 `cmd/alonabot/manage.go`：本切片负责内置包读取、预装与 update 的本地候选；goal-141 负责 install 的位置参数和组件 ID 解析，复用本切片的 `manage.BundledPackage`，保持 `InstallRequest` 字段含义。goal-141 还须同步 `scripts/dist/main.go` 中现有的 `pack --dir ... --goos ... --goarch ...` 调用；本切片只消费 dist 产物。两者按 PR 就绪顺序合并；后合并者同步最新 main，保留双方改动，重跑各自完成条件并联验「首次运行预装、install 官方组件 ID、显式 update 内置包」；先合并 goal-141 时，本切片接入它已合并的解析入口，复用同一读取实现。审查者合并任一 PR 后同步另一条语句的实际接口。

```text
/goal 使用 $alonabot-execute 推进 goal-74（切片 T-14.4：首次运行与预装）。goal 目录 goal-74，分支 goal-74，worktree .worktrees/goal-74，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：数据目录首次初始化时，按普通安装流程安装发行包 packages/ 中的全部包，只默认启用 help 与官方 WebUI；Agent 只安装不启用（D-50、D-62）；Telegram Adapter 只安装，core_config.json 已写实例或管理员为它添加第一个实例时启用（D-82；首个发行版没有其他预装插件与 OneBot 11 Adapter，D-171）。之后不再自动预装，已卸载的预装包不重新安装；新版本发行包中版本更高的内置包作为可用更新列出，由管理员确认后按 update 流程更新。

依据：spec/plan/release.md T-14 步骤 4 及失败停止条件；CONTRACTS §15.1、§15.2、§9.2、§9.3；DECISIONS D-50、D-62、D-82、D-171、D-172；TEST-MATRIX V-66。

范围（只新增或修改这些）：
- internal/app/alonabot.go、assembly.go；新增 internal/app/preinstall.go：启动时预装、配置了实例的 Adapter 安装与启用，以及内置包目录的装配
- internal/core/state/schema.sql；新增 internal/core/state/preinstall.go：首次预装完成标记及 GORM 访问方法（直接修改唯一初始迁移）
- internal/core/manage/service.go、instances.go、candidates.go；新增 internal/core/manage/bundled.go：读取内置包、List 的可用更新、update 选择本地候选与 CreateInstance 的自动启用
- cmd/alonabot/manage.go：仅同步内置更新的选择；install 参数从简由 goal-141 负责
- 新增 cmd/alonabot/firstrun_e2e_test.go（//go:build unix）；上述包的 *_local_test.go

共享文件与合并：与第 20 波的 goal-141 共享 internal/core/manage/bundled.go、cmd/alonabot/manage.go 及 packages/ 的读取规则。本切片负责内置包读取、预装与 update 候选，goal-141 负责 install 参数解析并复用 BundledPackage；按 PR 就绪顺序合并，后合并者同步 main、保留双方行为，并联验首次预装、install 官方组件 ID 与显式 update。goal-141 的 pack 参数变更还须同步 scripts/dist/main.go 的调用，本切片沿用其最终产物结构。

允许的依赖：标准库及 docs/development.md 已批准且由上述包使用的依赖；复用 contracts/release、internal/core/packages、manage、state、settings、trust 的现有依赖边。允许的新跨模块边：无。

接口：
- 复用 internal/app.Start(ctx context.Context, o Options) (*Core, error) 与现有已安装组件启动、候选提交入口；根包的 alonabot.Start/Run 签名保持现状。
- internal/core/state 新增 (*Store).Preinstalled(ctx context.Context) (bool, error) 与 (*Store).MarkPreinstalled(ctx context.Context) error。以 core_initialization 单例行的 preinstalled 状态记录首次预装完成；无记录返回 false。它是运行状态，Core 设置继续由 core_config.json 持有。
- internal/core/manage.Options 新增 BundledPackages string：可执行文件同目录的 packages/ 的绝对路径；既有 Packages 仍为 data_dir 下的不可变安装目录。
- internal/core/manage 新增 BundledPackage(dir, id string) (InstallRequest, error)：读取 dir/<id>/，按 release.json 选择当前 GOOS/GOARCH 的资产（静态 WebUI 使用 any/any），返回 Archive 的本机路径及 ReleaseJSON、Signature、PublicKey 的文件原始内容；缺失目录或文件保留可识别的 os.ErrNotExist，其他错误完整返回。它只解析输入，正式校验仍交给 Install/Update。
- manage.Package 新增 AvailableUpdate string，JSON 名为 available_update，空值省略；保存可用的内置包版本。
- 复用 (*Service).Install(ctx context.Context, req InstallRequest) (Package, error)、Enable(ctx context.Context, id string) error、List(ctx context.Context) ([]Package, error)、Update(ctx context.Context, id string, req InstallRequest) (UpdateResult, error) 与 CreateInstance(ctx context.Context, adapterPackageID, instanceID string, raw json.RawMessage, confirm bool) error，签名保持现状。InstallRequest.Archive 是路径，ReleaseJSON、Signature、PublicKey 是内容；Confirm 保留显式确认语义。

行为：
- 首次初始化：Start 取得 core.lock、完成 FirstToken 与管理服务装配后，在启动已安装组件之前读取 Preinstalled；未完成时查找 os.Executable 所在目录的 packages/。没有该目录时跳过（开发构建）；存在时逐个通过 BundledPackage 与 Install 的正常流程安装，使用 trust.Official 的内置公钥核对官方发布者。签名、摘要、平台与包内身份照常校验，沿用已合并的官方来源免确认逻辑，预装调用不以 Confirm=true 跳过第三方信任或替换确认。缺少签名、发布者与内置官方公钥不符、签名或摘要错误的包不安装并记 ERROR，继续处理其余包。
- 只为本次成功安装的 alonabot-plugin-help-official 与官方 role=webui 的 alonabot-webui-admin 设置默认启用，其余保持停用；static WebUI 启用即提供文件，当前没有 WebUI 进程。预装后重新读取安装记录，help 与需要启动的 Adapter 经现有 startInstalled 的首次候选准备、唯一提交与激活流程启动。启动循环尚未运行时的预装步骤直接衔接该流程，避免等待依赖 Core.operation 的管理回调。
- 全部包处理完后写入 MarkPreinstalled（即使个别包校验或安装失败）；标记读写失败返回完整错误。中途退出后再次进入时保留已有安装、已卸载记录与管理员选择，不以安装列表是否为空或 token 是否存在代替首次标记。
- 首次管理 token：复用 adminauth.FirstToken；stdin 与 stderr 均为终端时生成并显示一次，非终端正常启动并写入 data_dir/admin_token（0600），日志只写路径，首次成功登录删除文件。预装在 token 初始化之后执行。
- 管理员之后可经管理 API 或 CLI 启用 Agent；Agent 在没有配置 Provider 时也能启用（/ask 回复提示）。
- 之后的启动不再遍历预装；已卸载（Uninstalled）的预装包保持卸载，已由社区组件替换的同类组件保持管理员选择。
- 配置中的 Adapter：每次启动读取 settings.File.Settings().Adapters，尊重 AdapterSettings.Package 的所属组件；已有实例的已安装 Adapter 自动启用。尚无安装记录时优先取发行包 packages/ 的官方包，没有发行包时从同平台的官方 Release 发行归档取得对应组件，仍经普通安装与验签流程；沿用 T-14.3 的外层归档及 packages/<组件 ID>/ 结构，不套用 InstallRequest.Source 的单组件仓库下载布局。trust.Official 返回 false（开发构建）时不自动下载补装，记 ERROR 并提示用 alonabot install 手动安装（D-175）。保留的卸载记录报出需手动安装的错误，下载、校验或实例配置失败完整记录，其他组件继续按原规则启动。
- 可用更新：List 对仍绑定官方来源和发布者的已安装组件读取对应内置包，按 SemVer 比较版本（含预发布，忽略 build metadata），更高时填 available_update；保持安装版本、Active、启用状态与 pin 不变。损坏候选不列为可用更新并记 ERROR，其余条目照常列出；目录缺失时沿用原列表。管理员显式 update 且未指定 Source/Archive 时优先使用同一内置候选，再走既有来源、发布者、pin、依赖与候选提交校验；显式给出的输入保持优先，社区替换的组件沿用其来源。
- Telegram：CreateInstance 在原有实例所属组件确认、配置补全、环境变量展开、校验与设置提交之后，对尚未启用的 Adapter 复用内部 enable；保持现有实例锁与包锁顺序，不在持有同一包锁时再调用公开 Enable。启用后按现有实例启动路径应用新配置；设置已提交而启用失败时返回实际错误，保留配置以供管理员修复。

测试（按编码规范「测试提交规则」）：
- 提交：只新增 cmd/alonabot/firstrun_e2e_test.go 的一个集成主路径。用临时 Ed25519 密钥运行 go run ./scripts/dist --version 0.1.0 --key FILE --out DIR --platforms <当前平台>，解压到临时目录后运行其中的 alonabot run（干净 data_dir、无终端）：四个官方组件均已安装，help 与 WebUI 已启用，只有 help 有进程，Agent 与未配置实例的 Telegram Adapter 停用；经 admin-token show --data-dir DIR 取 token 并登录，文件删除；管理 API 与 alonabot list 的状态相符；卸载 help 后重启，保持卸载。复用已有 CLI/HTTP 集成辅助，不新增 testdata 插件。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：交互终端 token 只显示一次；签名缺失、签名错误、摘要错误与非官方发布者的包被拒绝并记 ERROR，其余包照常预装且标记完成；目录缺失、标记读写失败和预装中断后的再进入；更高版本只列更新，经显式 update 才提交，pin、第三方来源与卸载选择保持生效；已有 Adapter 配置、从本地发行包或官方 Release 补装、CreateInstance 自动启用与失败后保留配置。官方 Release 用本地 HTTPS 模拟服务，记录与真实环境的区别；与 goal-141 的共享入口按状态段联验。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...。

不做：容器镜像（T-14.5）；Core 自身的升级（替换可执行文件，由管理员操作）；CLI 参数从简（T-13.6）；增加首版组件；管理页面改造；修改 spec/（字段与证据在 PR「文档同步」中交审查者更新）。
```

### goal-77 · T-14.5 容器镜像

状态：已完成（PR #231），合并提交 `97aafdf`。以下保留派发时语句。派发时：前置 T-07.6（PR #100，合并提交 `cb83eeb`）与 T-14.3（PR #228，合并提交 `72d51c2`）均已合并。按 main `25e9064` 核对：`internal/core/settings` 提供 `Defaults() Settings`、`Load(dataDir string) (*File, error)`、`(*File).Settings() Settings` 与 `ValidateListen(address string) error`；监听安全联合校验在 `settings.go:validateSettings`，CLI 的 `--listen` 覆盖另在 `internal/app/alonabot.go:Start` 校验，两处均尚无容器例外。`internal/core/adminauth.FirstToken(ctx context.Context, st *state.Store, dataDir string, interactive bool) (string, error)` 已支持无终端写 `FirstTokenFile = "admin_token"`；`cmd/alonabot/admin_token.go` 的 show 要求显式 `--data-dir`。`scripts/dist` 的入口为 `go run ./scripts/dist --version VERSION --key FILE --out DIR [--platforms OS/ARCH,...] [--asset-base URL]`，内部以 CGO_ENABLED=0 构建并注入 `internal/core/trust.officialKey` 与 `internal/app.releaseVersion`；它需要 Go、Node 与 npm，产物根为 `alonabot-<版本>-linux-<架构>/`，含 Core 与 `packages/<组件 ID>/`，首版组件按 D-171。与 goal-74 共享 `internal/app/alonabot.go`，后合并者同步 main 并保留预装调用点、两处监听校验及容器 WARN；预装效果由 goal-74 交付，镜像与发行归档的联验由 goal-75 接续。

```text
/goal 使用 $alonabot-execute 推进 goal-77（切片 T-14.5：容器镜像）。goal 目录 goal-77，分支 goal-77，worktree .worktrees/goal-77，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：仓库根目录的 Dockerfile 用多阶段构建产生 linux/amd64 与 linux/arm64 镜像，deploy/compose.yaml 是官方整合包的 compose 入口：Go 构建阶段以 CGO_ENABLED=0 构建 Core 与首版全部官方包（Telegram Adapter、Agent、WebUI、help，D-171），最终镜像只含 alonabot、packages/ 与 CA 证书，以 UID 10001 运行 alonabot run --data-dir /data，/data 为卷，设置 ALONABOT_CONTAINER=1。Core 只在该环境变量下允许 0.0.0.0 的 listen 搭配本机 http 的 base_url，并在启动时输出 WARN。deploy/compose.yaml 给出卷、init: true、127.0.0.1:7650:7650 与首次启动后读取 admin-token 的用法。

依据：spec/plan/release.md T-14 步骤 5 及失败停止条件；CONTRACTS §15.3、§15.1、§7.4；DECISIONS D-36、D-82、D-171；TEST-MATRIX V-70 的镜像分项。

范围（只新增或修改这些）：
- 新增 Dockerfile、.dockerignore、deploy/compose.yaml
- internal/core/settings/settings.go：容器的 listen 缺省值与监听安全联合校验
- internal/app/alonabot.go：--listen 覆盖的同一容器例外与启动 WARN
- docs/development.md：增加「容器镜像」一节（本切片获授权修改这一节）
- 上述包的 *_local_test.go

允许的依赖：标准库与现有 Go 依赖；构建环境使用 docs/development.md 锁定版本的 Go、Node/npm、Docker BuildKit/buildx 与 CA 证书。最终镜像沿用 scratch。允许的新跨模块边：无；Dockerfile 调用现有 scripts/dist，不复制一套打包、签名实现。

接口：
- docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=VERSION --secret id=release_key,src=FILE -t ghcr.io/alonabot/alonabot:VERSION .；VERSION 必填，release_key 格式同 scripts/dist --key；可用构建参数 ASSET_BASE 传给 scripts/dist --asset-base，未给时沿用 dist 缺省值。
- 构建阶段以 TARGETOS/TARGETARCH 给出 scripts/dist --platforms，执行 go run ./scripts/dist --version VERSION --key /run/secrets/release_key --out DIR --platforms TARGETOS/TARGETARCH；输出解压后把 alonabot 与相邻 packages/ 放在最终镜像根目录。
- 最终镜像：USER 10001:10001、VOLUME /data、ENV ALONABOT_CONTAINER=1、ENTRYPOINT ["/alonabot"]、CMD ["run", "--data-dir", "/data"]；设置 PATH 使 docker compose exec alonabot alonabot 可直接找到 /alonabot，预建 /data 并授予 UID 10001 写权限。
- 复用 settings.Defaults() Settings、Load(dataDir string) (*File, error)、ValidateListen(address string) error 与 app.Start(ctx context.Context, o Options) (*Core, error)，保持签名；不新增 Go 导出符号。

行为：
- Dockerfile：构建阶段使用 go.mod 与 docs/development.md 锁定的 Go、Node 版本的官方镜像（标签写明版本与 digest），主机构建阶段调用 dist 为目标平台交叉构建；WebUI 的 npm ci/build 沿用 dist。签名密钥只经必需的 BuildKit secret 挂载，缺少时构建失败；私钥及 token 不进入构建上下文、镜像层、日志或产物。最终阶段从构建阶段复制 /etc/ssl/certs/ca-certificates.crt、Core 与 packages/，镜像中没有 shell、Node/npm、Python 或构建缓存。
- compose.yaml：服务名 alonabot，image 固定为对应版本（随 Release 上传的副本与版本相符）、init: true、命名卷挂载 /data、ports "127.0.0.1:7650:7650"。docker compose -f deploy/compose.yaml up -d 即可启动；首次 token 写入 /data/admin_token，用 docker compose -f deploy/compose.yaml exec alonabot alonabot admin-token show --data-dir /data 读出，首次登录后文件删除。
- settings：仅 ALONABOT_CONTAINER 的值精确为 1 时，listen 缺省为 0.0.0.0:7650，base_url 缺省仍为 http://127.0.0.1:7650；允许非 loopback listen 搭配本机 http base_url，其余 Host、base_url、HTTPS 与可信反代规则继续生效。设置文件与 CLI --listen 覆盖采用同一条件；显式 listen 保持优先。容器例外实际生效时，由 Start 输出一次 WARN "listening on all interfaces inside a container; publish the port to the host loopback only"。变量未设置或值不同于 1 时沿用现行校验。
- docs：写明两个平台的构建与运行命令、首次 token 读取、卷的写权限、端口只发布到宿主本机、远程访问用 SSH 隧道或 HTTPS 反代；MCP stdio 需要 shell 或其他运行时时须自建镜像。Release 的 compose 副本交由 T-10.4 随对应版本上传。

测试（按编码规范「测试提交规则」）：
- 提交：无。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：settings 的缺省值、显式设置、ALONABOT_CONTAINER 未设置/为 1/为其他值，以及设置文件与 CLI --listen 两条路径的安全校验和 WARN；使用本地 TLS 服务检查出站 HTTPS。
- 容器运行记录（本地验收，不提交测试代码）：以临时密钥 buildx 构建 linux/amd64、linux/arm64，记录 Docker/buildx 版本；在可用平台以干净命名卷和无 TTY 启动，正常访问 http://127.0.0.1:7650，show --data-dir /data 读出 token、登录后文件消失；经环境变量传入 Telegram 凭证的配置原文保持引用。通过 docker inspect/top/export 核对 UID 10001、卷写权限、镜像文件清单、出站 HTTPS 与 docker stop 后插件进程全部退出，沿用外部 Docker CLI 取证。预装及 Telegram 实例启动在 goal-74 可用后联验；缺少 Docker、某个平台或真实凭证的分项写「未运行」，交叉构建与模拟服务分别记录。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./internal/...；go build ./...；两个平台镜像的构建结果与可用环境中的 V-70 运行记录，缺失环境逐项写「未运行」。

不做：推送镜像到 ghcr.io 或发布 Release（T-10.4）；首次预装的产品实现（T-14.4）；Kubernetes 清单；增加首版组件；修改 spec/（证据与文档同步交审查者）。
```

### goal-85 · T-14.7 install.sh

状态：已完成（PR #236），合并提交 `44d77cd`。以下保留派发时语句。派发时：前置 T-14.3（PR #228，合并提交 `72d51c2`）已合并。按 main `25e9064` 核对：`go run ./scripts/dist --version VERSION --key FILE --out DIR [--platforms OS/ARCH,...] [--asset-base URL]` 生成 `alonabot-<版本>-<os>-<arch>.tar.gz`（Windows 为 zip），归档内有同名根目录，其下为 `alonabot` 与 `packages/<组件 ID>/`；同一输出目录的 `SHA256SUMS` 逐行是摘要、两个空格与归档文件名。`internal/core/trust.Official() (ed25519.PublicKey, bool)` 读取嵌入公钥，Core 版本由 `internal/app.releaseVersion` 注入，首版组件按 D-171。现有 CLI 为 `alonabot run --data-dir DIR`、`alonabot admin-token init --data-dir DIR` 与 `show --data-dir DIR`；init 要求 stdin/stderr 为终端，内部经 `state.Open(ctx context.Context, dataDir string) (*state.Store, error)` 持有 core.lock，再调用 `adminauth.InitToken(ctx context.Context, st *state.Store) (string, error)`，已有凭证时报错；`(*state.Store).AdminCredential(ctx context.Context) (state.AdminCredential, error)` 可区分尚无凭证的 state.ErrNotFound。main 尚无 Release 上传 workflow，资产上传由 T-10.4 的维护者发布步骤接续。V-84 的无终端预期仍为旧规则，与 CONTRACTS §15.6、PLAN T-14 步骤 7 不一致；本语句按 D-82 后的两处正文执行，由审查者同步验收条目。

```text
/goal 使用 $alonabot-execute 推进 goal-85（切片 T-14.7：install.sh）。goal 目录 goal-85，分支 goal-85，worktree .worktrees/goal-85，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：deploy/install.sh 按 CONTRACTS §15.6 在 Linux（systemd，root）与 macOS（launchd，当前用户）上安装 Core 与官方组件（首版为 Telegram Adapter、Agent、WebUI、help，D-171）：下载并按 SHA256SUMS 校验发行归档，校验失败不改动已有安装；替换程序目录；尚无管理凭证时经 /dev/tty 运行 admin-token init，没有终端时跳过初始化，启动服务后输出 admin-token show 的命令（D-82）；写入 systemd 单元（系统用户 alonabot）或用户 LaunchAgent 并启动；再次运行即升级，--uninstall 保留数据目录；不支持的系统在改动前退出。Release 流程上传 install.sh 与 SHA256SUMS。

依据：spec/plan/release.md T-14 步骤 7、切片表 T-14.7 及失败停止条件；CONTRACTS §15.1、§15.2、§15.6、§7.4；DECISIONS D-64、D-82、D-171；TEST-MATRIX V-84（无终端预期已按 D-175 与 §15.6 对齐）。

范围（只新增或修改这些）：
- 新增 deploy/install.sh：POSIX sh 安装、升级、卸载及写出 systemd 单元或 LaunchAgent
- cmd/alonabot/admin_token.go：仅增加脚本所需的 init --if-missing，复用同一锁与凭证读取
- scripts/dist/main.go：在发行输出目录附带 deploy/install.sh；SHA256SUMS 继续只列外层发行归档
- scripts/dist/dist_test.go：仅同步既有发行集成测试的产物断言
- docs/development.md：增加 install.sh 的安装、升级、卸载与本地验收说明（本切片获授权修改这一节）
- cmd/alonabot、scripts/dist 或 deploy/ 下的 *_local_test.go；不新增提交测试

允许的依赖：POSIX sh 与系统标准工具、curl、tar、Linux 的 sha256sum 或 macOS 的 shasum；服务注册、用户创建与身份切换复用目标系统已有的 systemd/launchd 及账号管理工具，缺少时在改动前明确报错，不由脚本安装工具。Go 修改只用标准库与现有 adminauth、state；shellcheck 仅用于检查，未安装写「未运行」。允许的新跨模块边：无。

接口：
- sh deploy/install.sh [--version VERSION]；省略版本时选择本仓库最新 Release；sh deploy/install.sh --uninstall 停止并删除服务与程序，保留数据目录并输出其路径。只接受已列出的参数，错误输出与 CLI 一样使用英文。
- ALONABOT_RELEASE_BASE_URL 覆盖仓库 Release 的基址，缺省 https://github.com/AlonaBot/AlonaBot/releases；指定版本使用 /download/v<VERSION>/，省略版本先经 /latest 确定同一个 Release，再下载该版本的归档与 SHA256SUMS。VERSION 与 scripts/dist --version 一致，归档名不额外带 v。测试基址也使用 HTTPS；不增加数据目录、程序目录或跳过校验的覆盖开关。
- Linux：程序 /opt/alonabot/，数据 /var/lib/alonabot，服务 /etc/systemd/system/alonabot.service，系统用户 alonabot，ExecStart=/opt/alonabot/alonabot run --data-dir /var/lib/alonabot。
- macOS：程序 ~/Library/Application Support/AlonaBot/，数据为其 data/，服务 ~/Library/LaunchAgents/io.github.alonabot.plist（Label 为 io.github.alonabot）；ProgramArguments 分别传入可执行文件、run、--data-dir 与完整数据路径，按当前用户登录会话启动。
- alonabot admin-token init --data-dir DIR --if-missing：新增可选 --if-missing；在同一 state.Open/core.lock 内读取 Store.AdminCredential，已有凭证时成功返回且不输出 token，state.ErrNotFound 时沿用 InitToken 并只向终端输出一次，其他错误完整返回。沿用 stdin 与 stderr 的终端要求；原 init/reset/show 的参数与行为保持现状。
- alonabot admin-token show --data-dir DIR 复用现有入口，服务运行时可读取首次 token 文件。Linux 给出的读取命令以 alonabot 身份执行；macOS 使用当前用户。

行为：
- 先检查 OS、arch、权限、服务管理器及所需系统工具，再创建下载临时目录。支持 linux/amd64、linux/arm64、darwin/amd64、darwin/arm64；uname 的 x86_64 映射为 amd64，aarch64/arm64 映射为 arm64。不支持的系统、架构、没有 systemd 的 Linux 或权限不足时，在改动文件前报错并提示发行归档或容器入口。
- 只经 HTTPS 下载且重定向保持 HTTPS；归档与 SHA256SUMS 来自同一个已解析的 Release。精确选取目标归档文件名的校验行，缺失、歧义或摘要不符都失败退出。该校验是 CONTRACTS §15.6 要求的传输完整性检查。下载、校验及临时解压全部成功后才停止旧服务、创建运行账号或改动安装目录；临时文件通过 trap 清理，失败保留原始错误。
- 按 dist 的同名根目录取出 alonabot 与完整 packages/，只替换程序及内置包目录。Linux 以系统用户 alonabot 运行服务，程序与 packages/ 供其读取，数据目录仅该账号读写；macOS 由当前用户持有。升级保留数据、core_config.json、状态库、已安装组件与未领取的 token；macOS 的 data/ 位于程序目录内部，替换程序时必须单独保留。
- 服务停止且有可用的 /dev/tty 时，以服务运行身份执行新程序的 admin-token init --if-missing --data-dir DIR，stdin 与 stderr 都连接该终端；凭证保持原有值或只显示新 token 一次，脚本不捕获、不记录 token。没有终端时跳过此命令，写入服务定义并正常启动，Core 自行写入 admin_token；输出完整的 admin-token show --data-dir DIR 读取命令，说明首次登录后文件删除。存在凭证时的升级照常启动，输出以文件存在为条件的首次 token 领取说明。
- systemd 单元以 alonabot 用户运行并启用开机启动；用户 LaunchAgent 以 io.github.alonabot 注册并在登录后启动。服务命令均为 run，工作目录避开组件源码工程；路径中的空格按 shell 与 plist 各自规则正确传递。Core 的预装与默认启用由 goal-74 接续，脚本不自行安装、启用或更新 data_dir 中的组件。
- --uninstall 停止服务，解除 systemd enable 或 launchd 注册，删除服务定义、alonabot 与 packages/；保留完整数据目录并输出位置，尤其保留 macOS 程序目录下的 data/。脚本不修改 core_config.json、防火墙或反向代理。
- dist 在 --out 中附带 install.sh，使它与同目录的 SHA256SUMS 可一起交给发布步骤；更新既有发行集成测试的产物断言即可，不另加提交测试。当前仓库没有上传 Release 的 workflow，本切片在 docs/development.md 与 PR「文档同步」列出上传清单，由 T-10.4 的维护者把 install.sh、SHA256SUMS 与归档上传至同一 Release。

测试（按编码规范「测试提交规则」）：
- 提交：无新增测试；仅随 dist 的产物调整改写 scripts/dist/dist_test.go 中既有集成主路径的断言，检查附带的 install.sh 内容与源文件相同。
- 本地自测（*_local_test.go，不提交，PR 中写明结果）：用本地 HTTPS Release 模拟服务覆盖指定版本与 latest、真实 dist 布局、校验缺失/错误、下载失败与不支持的系统；模拟系统命令只在临时目录取证，核对校验成功前无停服或安装写入、macOS 的 data/ 在升级与卸载时保留。覆盖 init --if-missing 的已有凭证、无凭证、持锁与读取失败；保留原 init 的已有凭证错误。模拟环境的结果单独记录。
- 本地语法检查：sh -n deploy/install.sh；shellcheck deploy/install.sh（未安装时写「未运行」）。
- V-84 真实环境：在可恢复的真实 systemd Linux（root）与真实 macOS（当前用户）分别完成全新安装、经管道运行且有/无终端、再次运行升级、校验失败及 --uninstall；记录 OS/arch、命令和结果，核对服务身份、token 交付、旧安装与数据保留。测试仅在临时目录或专用验收机上进行系统安装；开发机只做临时目录中的模拟自测。缺少真实环境的项目写「未运行」，保留给审查者的 V-84 验收缺口；容器内 systemd 与模拟结果分别记录。

完成条件：sh -n deploy/install.sh；shellcheck 的结果或未运行原因；gofmt -l 无输出；go vet ./...；go test -race ./cmd/alonabot ./scripts/dist；go build ./...；PR 中列出 V-84 各真实环境的结果与未运行项，并列出 TEST-MATRIX 无终端预期的同步项。

不做：正式签名、上传 Release 或推送镜像（T-10.4）；直接更新已安装组件；删除数据目录；Windows、其他服务管理器、开端口或配置反向代理；增加依赖；修改 spec/。
```

## 第 21 波

### goal-75 · T-14.6 发行包与镜像端到端验收

状态：已完成（PR #275），合并提交 `8a039310`。最终 head `72183ded` 的 [Quality gate](https://github.com/AlonaBot/AlonaBot/actions/runs/44631679053) 与 [PR metadata](https://github.com/AlonaBot/AlonaBot/actions/runs/44634909628) 通过；验收证据见 TEST-MATRIX §3.2。以下保留派发时语句。T-09.11（goal-45，PR #272）的正式流程、三类 Unix 场景与 T-08/T-09 联验已完成，原始证据与未运行项见 TEST-MATRIX §3.2；T-14.4（PR #235，`815e53c`）、T-14.5（PR #231，`97aafdf`）和 T-09.9（PR #259，`661cf9b1`）已合并。2026-11-19 按 goal-45 候选 `8f62c0bb` 与 main `c2d4774d` 复核以下现行页面、API 与构建入口；执行前确认 PR #272 已合并。

```text
/goal 使用 $alonabot-execute 推进 goal-75（切片 T-14.6：发行包与镜像的端到端验收）。goal 目录 goal-75，分支 goal-75，worktree .worktrees/goal-75，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」，确认 T-09.11 / goal-45 已合并。

目标：在干净的数据目录上，分别从发行归档与容器镜像完成首次运行、登录 WebUI、添加 Telegram 实例（本地 Bot API 模拟服务）、启用 Agent 并配置 Provider（本地模拟服务）、/ask 对话、/help 与 /whoami，记录 V-66、V-67、V-70 的端到端证据。首发内容为 Core、Telegram Adapter、Agent、WebUI 与 help（D-171）；真实 Telegram 与真实 Provider 缺少凭证时记为未运行。

依据：spec/plan/release.md T-14 步骤 3～6 及「失败停止条件」；CONTRACTS §7.4、§15.1～§15.4；ADMIN-API §1～§4、§12.1、§13、§15；TEST-MATRIX V-66、V-67、V-70 与 §3；docs/development.md「发行构建」「容器镜像」。前置编号以 spec/plan/core.md 的 T-09.11 / goal-45、spec/plan/webui.md 的 T-09.9 / goal-44 为准。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t14_e2e_test.go（//go:build unix）：一个表驱动集成测试 TestReleaseEndToEnd，分别消费发行归档与容器镜像；同文件内放本测试所需的进程、Docker 与 HTTP 辅助。
- cmd/alonabot/*_local_test.go：其余本地分项回归。
- 上述端到端验收直接涉及的对应包、构建与部署文件，以及其直接调用方和回归：允许修复验收中可复现且阻碍既有契约的缺陷，采用最小充分改动。每项修复在 PR 中逐条说明复现输入与完整错误、根因、改动位置、验证命令与结果，保留原始错误和证据。
- spec/TEST-MATRIX.md §3.2：追加 T-14.6 的实际证据行（本切片获授权修改这一处）。
允许的依赖：标准库与 cmd/alonabot 已有集成测试依赖；沿用 cmd/alonabot -> adapters/telegram/telegramtest、internal/core/packages 等已有测试依赖边。无新增依赖或跨模块边。本切片包含验收与上述必要修复；发现契约缺口或冲突时按 spec/README.md「执行与维护」流程处理，保留完整错误与证据。

现行入口与可复用辅助：
- scripts/dist/main.go：go run ./scripts/dist --version 0.1.0 --key FILE --out DIR --platforms <GOOS>/<GOARCH>。FILE 为临时生成的 base64 Ed25519 私钥文件，位于构建上下文外；同名根目录含 alonabot 和 packages/<组件ID>/ 的安装包、release.json、release.json.sig、release.pub。签名与摘要验证复用现有流程。
- cmd/alonabot/firstrun_e2e_test.go:TestFirstRunDistribution 已展示临时密钥、dist、packages.Extract、无终端启动、admin-token show、登录删除文件和四组件状态比对；复用同包 startRecoveryCore、stopTestCore、core.address/waitFor 及必要 HTTP 观察辅助。运行阶段使用解压后的二进制与 packages，工作目录放在临时解压目录；验收期间不以 go run 或开发目录组件替换候选进程。
- adapters/telegram/telegramtest 提供 NewServer(t, token)、SetMe(userID, username)、URL()、Push(...json.RawMessage)、Calls(method)、Requests()。cmd/alonabot/telegram_e2e_test.go 的 telegramMessage/telegramUpdate/telegramRequests 展示 update 与发信观察；仅复用模拟端与观察辅助，Core 启动使用候选资产。
- internal/app/agent_service_test.go 的 agentServiceConfig 与 serviceStream 展示 Provider 配置和 SSE 响应格式；这些辅助属于另一测试包，按当前字段和协议在本集成测试中准备必要数据，不新增公开导出。

候选构建与运行环境：
- 发行归档：按宿主 GOOS/GOARCH 构建一次，构建机需 Go 1.27.1、Node 26.5.1/npm 与已锁定的 Go/npm 依赖；dist 自行暂存源码、npm ci/build、pack 和签名。只有缓存已齐时才设置 npm_config_offline=true。构建临时目录需容纳嵌套 Go 构建与 WebUI 产物；缺空间或依赖下载失败如实报错，保留输出。
- 镜像：从仓库根执行 docker buildx build --platform linux/<amd64或arm64> --build-arg VERSION=0.1.0 --secret id=release_key,src=FILE --load -t alonabot:t14-e2e .。单次 --load 只选择一个平台；两架构分别记录实际运行，跨架构仿真单列。Dockerfile 的 BuildKit secret 必填，镜像为 scratch，入口 /alonabot，UID/GID 10001:10001，ENV ALONABOT_CONTAINER=1，默认 run --data-dir /data。
- 用独立命名卷与 docker run --init -d -p 127.0.0.1:<宿主端口>:7650 --mount source=<本次卷>,target=/data alonabot:t14-e2e 启动，记录容器 ID；只清理本测试创建的容器、卷和临时资产。现行 deploy/compose.yaml 固定正式镜像 tag，不能直接把它当作本次候选镜像；可用隔离的临时 compose override 指定候选 tag 与卷，保持 init 与 loopback 端口发布。
- 容器访问宿主模拟服务：Docker Desktop 使用 host.docker.internal；Linux 启动时加 --add-host host.docker.internal:host-gateway。telegramtest.NewServer 固定监听宿主 loopback，容器路径用标准库 httputil.NewSingleHostReverseProxy 在宿主可达接口转发到它，保留原 Server 的 Push/Requests 观察；Provider 模拟服务直接监听可达接口。外部 URL 改用上述主机名和转发/监听端口，宿主观察仍保留本机 URL。明确区分宿主 loopback 与容器 loopback，两个模拟服务都先验证可达性。scratch 镜像无 shell/curl，用 docker exec <容器> /alonabot admin-token show --data-dir /data 读取凭证，以宿主侧 HTTP、docker inspect/export/top 检查镜像与进程。

验收步骤（归档与镜像各完成一遍）：
1) 首次运行使用干净 data_dir/卷，无 TTY 正常启动，/data/admin_token（归档为所选 data_dir/admin_token）权限 0600；admin-token show 读出 token，日志不含 token。初始四包均已安装且签名有效，help 与带后台进程的 WebUI 启用，Agent 与无实例的 Telegram Adapter 停用。离线 admin-token init/reset 要求 Core 停止且 stdin/stderr 均为终端，交互分项另行验证；运行中的容器通过 show 取首次 token。
2) 请求 Host 与 core_config.json 的 base_url 一致。POST /api/auth/login，Content-Type: application/json、受控 Origin、正文 {"token":"..."}；保留管理 Cookie 与返回的 csrf_token，后续 Cookie 变更请求带 Origin 和 X-Alonabot-CSRF。GET /api/v1/packages 与 /api/v1/core 成功，首次登录后 token 文件删除。Bearer 方式只用于 CLI 对照。临时监听端口与受控 base_url 不同时沿用首次预装测试的 Host 转发方式。
3) POST /api/v1/adapter/instances/create，正文 {"package_id":"alonabot-adapter-telegram","instance_id":"tg_e2e","config":{"token":"${T14_TELEGRAM_TOKEN}","api_base_url":"<可达模拟服务URL>","proxy_mode":"none"},"confirm":false}；启动 Core/容器时传入仅用于测试的 T14_TELEGRAM_TOKEN。核对磁盘保留环境变量引用、API 脱敏、实例上线与 getUpdates 已被调用，Telegram 包自动启用。模拟端 SetMe 与 token 对应。
4) POST /api/v1/packages/enable，正文 {"id":"alonabot-agent-official"}。先启用缺省空配置，再 GET /api/v1/plugins/alonabot-agent-official/config 取得 active_generation；PUT 同路径的 /pending，正文带 base_generation 与完整 config。最小 config 示例（URL 用模拟端可达地址替换）：
   {"providers":[{"custom":{"name":"Local","protocol":"openai-compatible","base_url":"<模拟Provider URL>/v1","api_key":"synthetic-provider-key"}}],"models":[{"name":"Chat","provider":"Local","model":"local-model","type":"chat","max_tokens":128}],"chat_model":"Chat"}
   POST /api/v1/plugins/alonabot-agent-official/reload（无正文），等待新 Active，GET 返回秘密设置状态。chat_model 引用 models.name，models.provider 引用 providers 条目的 name，远端模型 ID 在 models.model。使用 HTTP pending/reload；CLI reload 会先 import 磁盘编辑文件。
5) 模拟用户 88 私聊发送文本 /ask hello，附 update_id、message_id、date、from.id、from.first_name、chat.id、chat.type；命令包含 text 参数。Bot API 的 bot_command entity 可按真实格式附 offset=0、length=4，Adapter 当前保持文本交 Core 匹配。Provider 模拟 POST /v1/chat/completions，检查 synthetic API key、model 与用户正文，返回 text/event-stream 的两个内容增量、完成标记与 [DONE]；用同步观察控制增量，让请求记录能验证 sendMessage 后 editMessageText、同一消息 ID 与完整最终文本，避免单个瞬时结果掩盖编辑路径。
6) 同一用户依次发 /help 与 /whoami，按每次请求前的记录位置截取新增回复：help 列出当前可用命令（含 ask），whoami 包含 tg_e2e 与用户 88。POST /api/v1/admins/add，正文 {"user":"telegram:88"}，再 GET /api/v1/core/admins 并核对持久化；必要时再次发送 help 验证权限过滤。管理 token 与聊天 Core Admin 分别使用各自入口。
7) 真实浏览器打开候选 WebUI，从登录页完成登录，在组件/账号/模型页面核对配置与启用状态，页面保存与 reload 后的结果与上述 API 一致；使用前置合并后的真实页面路径。API 自动联验与浏览器操作分别记录。缺浏览器时保留页面分项为未运行，交审查者承接。
8) 归档发送 SIGTERM，等待 Core 与子进程退出；容器 docker stop 后检查 State.Running=false、无存活组件进程，保留停止日志。复用 V-66 的卸载后重启、错误签名与内置更新仅列出的已有证据，补齐候选资产上仍缺的分项，不伪造通过。

V-70 补充分项：核对 linux/amd64 与 linux/arm64 的实际镜像运行记录；镜像只携带 Core、四组件、CA 及必要 /data 目录，UID 10001；无 TTY 启动与 show、登录删除 token；端口发布限 loopback；通过真实出站 HTTPS 请求核对 CA 可用性（模拟 HTTP 无此证据）；在未设置 ALONABOT_CONTAINER 的归档进程中核对 0.0.0.0 搭配本机 http base_url 被拒绝。网络、Docker 或架构环境缺失时分别写未运行。

测试与证据（按编码规范「测试提交规则」）：
- 提交一个 TestReleaseEndToEnd，归档主路径默认运行；容器子场景用 ALONABOT_TEST_DOCKER=1 显式启用（测试运行开关），未启用时 t.Skip 并在证据中写未运行。启用后 Docker 构建、连接或断言失败按实际失败返回。两种资产共用上述业务断言；多场景理由：归档与镜像的进程入口、权限、网络与卷不同。
- 本地自测：交互 TTY token、错误签名、卸载不重装、更新只列出、容器静态检查、HTTPS 出站及浏览器操作；复用已合并辅助和证据，无需新增 testdata 插件或测试依赖。
- spec/TEST-MATRIX.md §3.2 写测试位置、资产版本与平台、Docker/buildx 版本、原生或仿真、命令、实际结果及未运行原因。真实 Telegram、真实 Provider、其余平台与正式发布分别记录，模拟结果按模拟环境列出。

完成条件：受影响测试与 vet/build 通过；go test -race ./cmd/alonabot -run '^TestReleaseEndToEnd$' 通过归档路径；有 Docker 时以 ALONABOT_TEST_DOCKER=1 执行同一测试并记录容器路径。完整 CI 的 go vet ./...、go test -race ./...、go build ./... 复用同候选有效结果，按共同约束记录；文档修改通过 git diff --check。缺少真实平台或外部环境的分项保留明确缺口与承接者，T-10 汇总时继续闭合。

不做：五平台运行（T-10.2）；正式发布（T-10.4）；首发范围外的 Adapter 或插件；新产品功能、依赖、runner 运维与容量清理。
```

## 第 22 波

### goal-46 · T-10.1 证据汇总与跨组件联验

状态：实施与证据审查中（PR #278）。T-10.1 的全部前置已合并；T-14.6 / goal-75（PR #275，`8a039310`）交付真实归档、双架构镜像与浏览器联验。2026-11-20 按该合并点复核现行入口，main `cb5acc79` 的 Quality gate（run 44647524835）已通过。跨组件实现与矩阵候选已交草稿 PR #278，审查者继续补浏览器证据并核对最终 head CI；GN11 接入 goal-174 的固定产品候选并补增量联验；D-184 已接受当前资源开销，资源测量收口。T-09.11（goal-45，PR #272）提供 `scripts/scenarios/run.py` 的 build/management/build-recovery/recovery/survivor 五个入口，实际调用方式见[场景说明](../../scripts/scenarios/README.md)；普通 Core、包与故障 overlay 分别构建后作为明确输入消费。Unix 场景使用 FIFO，Windows 的实际恢复由 goal-47 对应入口验收。完整前置以 spec/plan/release.md 的 T-10.1 行为准，包含已合并的 T-09.27 / goal-177 与 T-16.2 / goal-82。复用 goal-75 的有效资产及原始证据，在本切片更新 TEST-MATRIX 的历史总述和完整 CI 引用。S-06.2 / goal-174 已完成固定 Linux 实测并按 D-184 收口，最终消费稳定合并输入完成公开入口联验。

```text
/goal 使用 $alonabot-execute 推进 goal-46（切片 T-10.1：证据汇总与跨组件联验）。goal 目录 goal-46，分支 goal-46，worktree .worktrees/goal-46，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：汇总 V-01～V-76 与 V-82～V-117 的证据，逐项标记已通过、失败、未运行或待联验；补齐组件之间的联验（首次管理登录 → 安装启用 → 配置/权限 → 真实消息 → 插件 HTTP → 重载/升级/回退，TCP 与 stdio）；用不同的真实插件二进制覆盖各能力组合；执行协议、权限、资源边界与完整恢复的故障注入，核对数据库内容与权威指针。

依据：PLAN T-10 步骤 1～4 与「停止和失败条件」；TEST-MATRIX 全部；CONTRACTS §12.4。

范围（只新增或修改这些）：
- 新增 cmd/alonabot/t10_e2e_test.go（//go:build unix）；SDK 绑定的回归及临时插件写在 *_local_test.go，沿用 D-86，不新增 testdata 插件
- spec/TEST-MATRIX.md：§3 的每个 V 条目更新为最新状态，新增「发布前证据汇总」一节（本切片获授权修改 TEST-MATRIX；其他 spec 文件只在发现契约缺陷时按 spec/README 的流程修改，并在 PR 中说明）
- 发现的缺陷：在对应包中修复，每项修复在 PR 中写明 V 编号、复现输入、完整错误、根因、改动位置与验证命令和结果，连同直接调用方、fixture 与回归完成；契约缺口或冲突按 spec/README.md「执行与维护」处理
允许的新跨模块边：无。

现行入口与证据规则：
- 使用 cmd/alonabot 的 run、install、enable、restart、reload、update、pin、disable、uninstall、rollback 正式命令与 ADMIN-API；禁止沿用 serve、plugin generate 或旧 --dir/--goos/--goarch 参数。build/check/test/pack 目录为位置参数，pack 用 --target。复用 cmd/alonabot 的 manage_e2e_test.go、recover_e2e_test.go、firstrun_e2e_test.go、t12_e2e_test.go 及 goal-45 合并后的场景，真实 Core/组件进程由候选包启动。
- goal-45 场景：先用 `uv run --no-project scripts/scenarios/run.py build --out <新目录>` 构建三版包，management 接收 `--alonabot`、`--data-dir` 和三个版本包的位置参数；恢复产物通过 build-recovery 构建，recovery 接收 `--alonabot`、`--overlay-core`、`--data-dir`、`--recovery-package`、`--probe-package`、`--adapter-package`，survivor 接收 `--alonabot`、`--overlay-core`、`--data-dir`、`--recovery-package`。包、普通 Core 与故障 overlay 在当前 Unix 平台构建，运行阶段消费显式输入；完整命令和产物要求见 scripts/scenarios/README.md。
- 首次运行无 TTY 时正常启动并写 data_dir/admin_token（0600），show 只读取得一次性 token，首次登录删除文件。离线 init/reset 在 Core 停止且 stdin/stderr 为终端时单列验证。管理 Cookie、Origin、X-Alonabot-CSRF 与 CLI Bearer 分别按 ADMIN-API §1～§2；管理 token、聊天 Core Admin、internal HTTP 身份分别验证。
- 首发预装只包含 Telegram Adapter、Agent、WebUI、help；help 与带后台进程的 WebUI 默认启用，其余按配置/管理操作启用。额外示例仅为联验安装，保持发行清单。Telegram 使用 adapters/telegram/telegramtest，Provider 使用本地 HTTP 模拟；真实凭证仅通过已有批准机制读取，记录模拟与真实平台证据。
- 插件页面复用 examples/plugins/showcase 的 /status 与 /page，经 /plugins/<id>/、单一 base_url、管理认证、WebUI sandbox iframe/postMessage 桥与 sdk/page 验证（含 D-183 的公开套件与静态资源）；按当前 ADMIN-API、sdk/page 与前置合并后的桥协议逐条核对，禁止沿用页面票据流程。API、浏览器、TCP、stdio 的结果分别记录。
- configdb 的默认 sqlitedb 与 configdb_gorm 标签变体分别构建，fail_migration、fail_start 与 configdb_startup 用于真实候选/数据库观察；配置按当前 schema 的校正与 Pending/Active 流程验收，历史 OnConfigLoad/配置版本叙述按现行契约核对。
- V-01～V-76、V-82～V-117 完整保留，并逐项追踪 TEST-MATRIX 当前其他首发适用 V 条目及关联证据；D-171 后置的 OneBot 11、welcome、rss 等分项标随后续版本。历史票据或旧配置流程证据保留来源，新增当前契约的复验结论与缺口。external 维护许可、休眠等由对应 owner 完成，消费其合并后的公开入口与故障场景，不在本轮假定它们已交付。

联验矩阵（各为一个子测试，TCP 与 stdio 各一遍）：
1) 组件链路：干净 data_dir 的 alonabot run 首次启动 → admin-token show → Cookie 登录与 CSRF → 安装并启用 echo、configdb、showcase HTTP 示例与官方组件 → 配置与权限 → Telegram 模拟服务的真实消息 → 插件 HTTP 页面（同一 base_url 下，经管理认证与 WebUI 的 iframe/postMessage 桥）→ reload → update → rollback。
2) 能力组合（每个能力使用不同的插件二进制，不用同一个 worker 的多份进程代替）：命令、文本匹配、多轮等待、主动 Send、Service、Event Bus、Scheduler、HTTP、sqlitedb 与 GORM（含迁移），以及 Agent 与 WebUI 组件；既有无配置无数据库的插件，也有实际写入与迁移的插件。
3) 边界：伪造与过期 Session、命令遮蔽、无 Actor、实例许可、Hook 递归、媒体租约、满准入、Cancel 顺序、超时仍运行的 job；错误回执与已确认的外部结果一致。
4) 恢复：停止与持锁、SQLite WAL 备份、候选 OnStart 写入后失败、PREPARED 与提交前后 Core 退出、不可变 secret、external 插件的维护许可与 resume、显式 rollback 的新代次；核对数据库内容与状态库的 Active 指针，不能只看日志里的「恢复成功」。
- 汇总：TEST-MATRIX 中每个 V 条目写明测试位置、命令、平台与结果；失败的列为阻塞项；没有运行条件的（真实 Telegram、真实 Provider、非 darwin 平台等）写「未运行」与原因。

测试（按编码规范「测试提交规则」）：
- 提交：cmd/alonabot/t10_e2e_test.go。
- 本地自测（不提交，PR 中写明结果）：完整测试套件在 -race 下运行一次的总耗时与结果。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；首发适用条目逐项完成联验；仍缺真实凭证或环境的项保留「未运行」、原因与承接者，禁止用汇总文档替代实际测试。失败项按 T-10 停止条件修复或取得维护者明确裁定后继续。

不做：五平台运行（T-10.2）；资源报告（T-10.3）；发布（T-10.4）。
```

## 第 23 波

### goal-47 · T-10.2 五平台构建与运行

状态：独立实施中，最终验收等待 T-10.1 / goal-46。2026-11-20 从经父审查的 goal-46 候选 `eb01ed6c` 开始平台程序与可用环境的独立实施；最终以整合后的同一候选闭合五平台要求，后继候选变动补受影响验证。五平台均需实际运行；2026-11-19 GitHub-hosted 四平台短探测 run 44492249496 因付款或支出限额在分配 runner 前失败。现有 instabox linux/amd64、本机 darwin/arm64 与 Docker Desktop 的 linux/arm64 作为可用环境候选，执行时核对实际身份与架构；2026-11-19 主线程在本机执行 `arch -x86_64 /usr/bin/uname -m` 返回 `x86_64`，据此确认 Rosetta 可执行候选，完整 darwin/amd64 产品验收留待 goal-47，结果按架构仿真标记。Windows/amd64 的实际执行入口与 GitHub Actions 计费恢复按维护者决定留到最后处理；其余可用平台继续实施和实跑。

```text
/goal 使用 $alonabot-execute 推进 goal-47（切片 T-10.2：五平台构建与运行）。goal 目录 goal-47，分支 goal-47，worktree .worktrees/goal-47，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在 linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64 五个组合上分别构建并运行 Core 与插件，覆盖启动、停止、至少 100 次启停、异常退出、合作型子进程树与端口清理；记录 CGO 插件的实际支持范围。

依据：PLAN T-10 步骤 5；CONTRACTS §7.1（平台与进程树）、§12.2；TEST-MATRIX V-32、V-34。

范围（只新增或修改这些）：
- 新增 scripts/platform/（可在各平台运行的验收程序：Go 程序，调用构建好的 alonabot 与测试插件，输出结构化结果）
- Unix 平台复用 goal-45（PR #272）已交付的 scripts/scenarios；其中 run.py 与生成测试使用 Unix/FIFO。Windows 在 scripts/platform 内通过原生进程、锁和检查点驱动对应恢复场景，核对相同 Active/Journal/业务 DB、残存旧树与恢复结果；实际入口与产物要求见 scripts/scenarios/README.md
- spec/TEST-MATRIX.md：V-32、V-34 与各平台相关条目的证据（本切片获授权修改 TEST-MATRIX）
- 发现的平台缺陷：在对应包及直接调用方修复（例如 Windows Job Object、路径与锁），每项在 PR 中写明平台、复现、完整错误、原因与验证；SDK 绑定与临时进程树/CGO fixture 按 D-86 放本地测试，不新增 testdata 插件。契约冲突按 spec 流程处理
允许的新跨模块边：scripts/platform 只经 os/exec 调用构建产物，不 import 本 module 的包。

行为：
- 每个平台：用 scripts/dist 的构建产物（或在该平台本地构建）运行 scripts/platform：启动 alonabot run --data-dir <干净目录>、取得首次 token 并登录、安装并启用 echo 与 showcase HTTP 示例、收发一条消息、访问一次插件 HTTP、停止；再循环 100 次启停（每次检查插件进程与监听端口都已释放）；强制结束 Core（Unix 为 SIGKILL，Windows 为 TerminateProcess）后重启，确认插件进程已退出且能恢复；启动一个会派生子进程的测试插件，停止后子进程消失（合作型进程树）。
- scripts/platform 为本切片新增的可执行 Go 验收程序；输入明确的候选 Core 路径、当前平台示例安装包目录、goal-45 场景位置和结果输出路径，使用管理 CLI/HTTP 驱动，不 import 测试包私有辅助。具体 flag 在程序 --help 与运行说明中一致登记；程序运行工作目录和 data_dir 放在临时目录，避免源码开发运行改变行为。
- 五平台分别记录实际 OS/架构、Core/组件构建目标与执行环境（原生、虚拟机、容器或架构仿真）；Windows 真正运行 .exe，Intel macOS 使用真实 x86_64 执行环境。Docker Linux 结果标明容器与宿主，Rosetta/QEMU 结果标明仿真；跨编译、宿主短探测和跳过结果单列。CI 环境未落实时继续独立可运行平台，保持缺口，不改仓库权限、计费或 runner。
- 无 TTY 首次启动取得 admin-token 文件，通过 show 与登录完成认证；管理 HTTP、showcase 页面与桥用现行 Host/Origin/CSRF，动态监听地址按首次运行集成测试处理。进程树核对需记录实际 PID/子进程与监听端口，杀 Core 后等待旧插件树确认退出和 runtime.lock 释放，再验证恢复，保留故障与恢复前后数据。
- 结果写成 JSON（平台、Go 版本、OS 版本、各步骤的结果与耗时），附在 PR 中，摘要写入 TEST-MATRIX。
- CGO：记录 C 编译器、链接库及构建参数；需新工具/依赖时先列明并取得批准。用一个 CGO 测试插件在各平台尝试构建与运行，记录实际支持范围；不支持的平台标为不支持，不发布。
- 没有某个平台的运行环境时，该平台写「未运行」；交叉编译成功不能代替运行结果。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/platform（本身就是验收程序）；平台缺陷修复所需的集成测试。
- 运行记录：每个平台的 JSON 结果附在 PR 中。

完成条件：gofmt -l 无输出；go vet ./...（含 GOOS=windows、GOOS=linux 的 vet）；go test -race ./...（当前平台）；每个可用平台的 scripts/platform 运行结果。

不做：资源报告（T-10.3）；发布（T-10.4）。
```

### goal-48 · T-10.3 现有资源实测与维护者决定归档

状态：等待 goal-46 稳定输入后文档收口。2026-11-20 维护者接受当前开销并要求停止该方面测量（D-184）；原四负载及 NoneBot2 对照取消新增运行。已批准的两组基准依赖及本地兼容预检保留历史，不再扩展基准实现。

```text
/goal 使用 $alonabot-execute 推进 goal-48（切片 T-10.3：现有资源实测与维护者决定归档）。goal 目录 goal-48，分支 goal-48，worktree .worktrees/goal-48，从最新 origin/main 创建。

目标：在 docs/benchmarks.md 收录 goal-174 已取得的固定 linux/amd64 资源数据、测量条件、真实候选与原始记录位置，以及 D-184 的维护者接受决定；同步 TEST-MATRIX V-33、V-74。

范围：docs/benchmarks.md、spec/TEST-MATRIX.md 的相关证据。复用已核算的九段120s/15进程读数和原有效休眠/冷启动记录，列每轮 PSS、RSS、私有内存和中位数；保留旧预算比较及历史失败。V-33 四负载和 NoneBot2 对照标记“维护者取消新增测量”，保持未取得的指标缺项。

完成条件：数据与原始记录一致，git diff --check 通过，文档经审查后交付 PR。不新增采样、压测、性能优化、基准工具或依赖；后继 goal-49 消费该文档与 D-184。
```

## 第 24 波

### goal-49 · T-10.4 发布资产与说明

状态：等待前置（T-10.2 / goal-47、T-10.3 / goal-48）。正式签名与发布由维护者在其控制的环境中执行。

```text
/goal 使用 $alonabot-execute 推进 goal-49（切片 T-10.4：发布资产与说明）。goal 目录 goal-49，分支 goal-49，worktree .worktrees/goal-49，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：生成五平台候选资产（首个发行版的组件：官方 Telegram Adapter、Agent、WebUI 与 help，D-171）、T-14 的发行归档、容器镜像与 install.sh/compose、外部 release.json 与签名；用干净的测试 data_dir 从这些真实资产完成安装、更新与卸载重装，核对来源、长度、摘要、平台与 Schema/二进制一致性，确保安装不依赖开发机上未打包的文件；关闭全部阻塞用例，写出发布说明。签名私钥只在维护者控制的环境中使用；发布到 GitHub Release 与推送镜像由维护者执行。

依据：PLAN T-10 步骤 7、8；CONTRACTS §9.2、§15.2、§15.3；TEST-MATRIX V-28、V-41 的发布复验。

范围（只新增或修改这些）：
- 新增 docs/release-notes/v0.1.0.md（版本、平台、依赖、已知限制、未运行项、恢复步骤、升级与回退说明）
- 新增 scripts/release-check/（用候选资产在干净 data_dir 完成安装、更新、卸载重装与一致性核对的 Go 程序）
- spec/TEST-MATRIX.md：V-28、V-41 的发布复验证据，以及「发布前证据汇总」的最终状态（本切片获授权修改 TEST-MATRIX）
- README.md：安装与快速开始指向发布资产（本切片获授权修改 README 的这一节）
- 候选资产验收中发现的缺陷：对应包、构建/部署文件、直接调用方及必要回归，逐项保留复现、完整错误、原因和验证；契约缺口/冲突按 spec 流程处理
允许的新跨模块边：scripts/release-check -> contracts/release（ARCHITECTURE 已登记，与 scripts/dist 同类）；其余只经 os/exec 调用候选资产。

行为：
- 候选资产：执行者用测试密钥运行 scripts/dist 与镜像构建，得到与正式发布结构相同的候选资产；正式签名由维护者用官方私钥重新运行同一命令完成。
- release-check：在干净 data_dir 中只使用候选资产（不访问源码树）：首次运行预装、安装一个第三方测试包、用新版本资产 update、卸载并同来源重装；对每个包核对 release.json 的来源、长度、SHA256、平台与包内 Schema、manifest、二进制的一致性；签名错误、来源不符的资产被拒绝（V-28、V-41 复验）。
- 现行构建：go run ./scripts/dist --version <版本> --key <临时测试密钥文件> --out <目录>，缺省五平台，--platforms 可选择子集；归档内 Core 与 packages/<ID>/、dist/packages/<ID>/ 各平台资产、合并 release.json/.sig/.pub 及 SHA256SUMS 逐一登记。Core 版本由 internal/app.releaseVersion 注入，GET /api/v1/core 与构建版本一致；每个组件版本读取自己的 plugin.yaml，不能以外层归档版本代替更新版本。
- Dockerfile 用 VERSION 与 BuildKit release_key secret 构建 linux/amd64、linux/arm64；compose 副本固定本次候选镜像 tag，保留 init、UID 10001、独立 /data 卷、loopback 端口。将 deploy/install.sh 与版本固定的 deploy/compose.yaml 纳入候选资产清单，复用 T-14.7 的真实 systemd/macOS 安装、升级、卸载与失败保持原安装的证据，缺项按原环境继续实测。
- scripts/release-check 为本切片新增程序，输入候选发行目录、同来源递增组件版本的更新候选目录与临时结果目录，--help/文档登记实际参数。运行时解压并调用候选 alonabot，当前工作目录隔离源码树；无 TTY 启动正常预装后 admin-token show，按现行 Cookie/CSRF 或 CLI Bearer 管理入口操作。第三方样例从现有示例打包，SDK 绑定 fixture 按 D-86 放本地回归。
- 更新用每个组件 plugin.yaml 中实际更高的 SemVer，再打包签名并维护同一来源/发布者；验证 update 前后版本、Active 代次与数据库内容、pin 及卸载重装的来源绑定。摘要用于发布资产完整性校验，按清单核对文件长度、签名确切字节、平台选择与资源内容；不为普通文档检查增加摘要操作。
- 发布说明：写明支持的平台与实际运行结果（来自 T-10.2）、资源报告位置（T-10.3）、全部「未运行」项与原因、已知限制（Native 插件是可信代码、库模式无隔离、插件自管的外部数据不承诺自动回退等）、备份与恢复步骤、升级与回退方法。
- 阻塞用例：TEST-MATRIX 中标为失败的条目必须全部修复或经维护者同意延后（在发布说明中列出）。
- 插件模板（D-169）：核对 AlonaBot/alonabot-plugin-template 的 go.mod 指向本次发布的 SDK 版本、模板 CI 在候选版本上通过，并在发布说明中给出模板仓库；维护者在 AlonaBot 仓库公开时一并公开模板仓库。
- 发布：维护者在其环境中用 scripts/dist 构建并签名，把发行归档、各安装包及其 release.json/.sig/.pub、SHA256SUMS、install.sh 与版本固定的 compose 副本发布到本仓库的 GitHub Release，把镜像推送到 ghcr.io/alonabot/alonabot（:<version> 与 :latest）；执行者不持有私钥，不执行发布。

测试（按编码规范「测试提交规则」）：
- 提交：scripts/release-check（验收程序）。
- 运行记录：release-check 对候选资产的运行结果附在 PR 中。

完成条件：gofmt -l 无输出；go vet ./...；go test -race ./...；go build ./...；release-check 对候选资产通过；TEST-MATRIX 没有未处理的失败条目。

不做：执行正式签名与发布（维护者负责）；v0.1 之后的路线图。
```
