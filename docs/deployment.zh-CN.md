# 安装与部署

[English](deployment.md)

本文面向运行 AlonaBot 的管理员，说明安装、认证、配置、组件管理与恢复。源码开发见[开发指南](development.zh-CN.md)，构建发行物见[发行维护](releasing.md)。

## 选择安装方式

从 [v0.2.0 Release](https://github.com/AlonaBot/AlonaBot/releases/tag/v0.2.0) 下载正式资产。五平台归档与 `install.sh` 提供纯 Core；Docker Compose 提供 Core、WebUI 与 Agent 整合包。Adapter 和 help 插件按需安装。具体平台实跑与已知限制见[发布说明](../changelog/v0.2.0/zh-CN.md)。

| 方式 | 适用场景 |
| --- | --- |
| 发行归档 | 手动运行 Core，适用于五个发行目标 |
| install.sh | Linux systemd 系统服务与 macOS 当前用户 LaunchAgent |
| Docker Compose | Linux amd64/arm64 容器，持久化数据卷 |

## 手动运行发行归档

下载对应平台的归档，解压后可见 `alonabot`（Windows 为 `alonabot.exe`）、`LICENSE` 与 `source.json`。在该目录运行：
```fish
./alonabot run --data-dir ./data
```

后文管理命令假定 `alonabot` 已加入 PATH；从解压目录操作时使用 `./alonabot`。Windows 使用 `alonabot.exe`。首次启动的组件清单为空。按[管理认证](#core-设置与-http-入口)领取管理 token，按[管理组件](#管理组件)安装所需组件。安装并启用 WebUI 后可打开 `http://127.0.0.1:7650` 登录；Agent 配置 Provider 与模型后再启用。

## 系统服务安装与升级

在支持 systemd 的 Linux 上以 root 运行；macOS 在当前用户的登录会话运行。支持两种系统的 amd64、arm64。先从主仓下载安装脚本，再安装指定版本（以下命令使用 fish，把 `VERSION` 换成发行版本号，省略前导 `v`）：
```fish
set version VERSION
curl --fail --location --proto '=https' --proto-redir '=https' \
  -o install.sh https://raw.githubusercontent.com/AlonaBot/AlonaBot/main/deploy/install.sh
# Linux
sudo sh install.sh --version $version
# macOS
sh install.sh --version $version
```

省略 `--version` 会先通过 `/latest` 固定一个 Release，再下载该版本的归档，并与 GitHub Release API 的资产 SHA256 摘要核对。`ALONABOT_RELEASE_BASE_URL` 可覆盖 Release 基址；自定义镜像站继续从同一 Release 下载 `SHA256SUMS`。基址和所有重定向均须使用 HTTPS。脚本先检查平台、权限和系统工具，完成下载、校验和临时解压后才停服、替换 Core 程序并清理旧的相邻 `packages/` 安装介质；数据目录中的已安装组件保留。自定义镜像站若仍提供整合归档，脚本同时复制其 `packages/`。

Linux 使用 `/opt/alonabot/`、数据目录 `/var/lib/alonabot` 和系统服务 `alonabot.service`；服务以系统用户 `alonabot` 运行并启用开机启动。macOS 使用 `~/Library/Application Support/AlonaBot/`、其中的 `data/` 和用户 LaunchAgent `io.github.alonabot`，登录后启动。工作目录设为程序目录，启动参数为 `run --data-dir <完整数据路径>`。

脚本在停服后，有控制终端时以服务身份运行 `admin-token init --if-missing`，通过 `/dev/tty` 初始化；无控制终端时交给 Core 首次启动初始化。新 CLI 以 0600 权限持久保存凭据，只输出路径；已有凭据沿用。远程 WebUI 登录时可显式读取：
```fish
# Linux，以服务身份读取
sudo runuser -u alonabot -- /opt/alonabot/alonabot admin-token show --data-dir /var/lib/alonabot
# macOS，以当前用户读取
"$HOME/Library/Application Support/AlonaBot/alonabot" admin-token show --data-dir "$HOME/Library/Application Support/AlonaBot/data"
```

登录及确认均保留持久凭据。再次安装即升级，保留数据、配置、状态库和已安装组件。
```fish
# Linux 卸载
sudo sh install.sh --uninstall
# macOS 卸载
sh install.sh --uninstall
```

卸载停止并解除服务注册，删除服务定义、程序与内置包，输出保留的数据目录位置；macOS 的 `data/` 保留在原位。

## 容器镜像

下载发行的 Compose 文件，启动整合包并读出首次登录的 token：
```fish
curl --fail --location --proto '=https' --proto-redir '=https' \
  -o docker-compose.yml https://raw.githubusercontent.com/AlonaBot/AlonaBot/v0.2.0/docker-compose.yml
docker compose up -d
docker compose exec alonabot alonabot admin-token show --data-dir /data
```

打开 `http://127.0.0.1:7650`，使用读出的 token 登录。首次启动无需 TTY，Core 将管理 token 写入 `/data/admin_token`，登录后保留该文件供本机 CLI 使用。镜像支持 `linux/amd64` 与 `linux/arm64`，compose 按宿主平台选择。compose 中固定镜像版本，与所选 Release 的版本保持一致。

镜像以 `10001:10001` 运行，命名卷挂载到 `/data`，首次创建卷时继承镜像中该目录的所有者。改用宿主目录挂载时，先授予 UID/GID 10001 读写和遍历权限；已有卷也需保证这一权限。保留该卷即可保留配置、状态库和组件数据。

端口只发布到宿主的 `127.0.0.1`。镜像设置 `ALONABOT_CONTAINER=1` 时，缺省监听为 `0.0.0.0:7650`，`base_url` 仍为 `http://127.0.0.1:7650`；显式 `listen` 优先。该容器例外实际生效时，启动日志输出一次 WARN 提醒宿主本机发布规则。远程管理可使用 `ssh -N -L 7650:127.0.0.1:7650 user@host`，再打开本机地址；HTTPS 反代部署按 [Core 设置与 HTTP 入口](#core-设置与-http-入口) 配置 `base_url` 和 `trusted_proxies`。

按需安装 Adapter 后，在 compose 服务的 `environment` 中传入 `TELEGRAM_BOT_TOKEN` 后，`/data/core_config.json` 中的实例凭证可写成 `"token": "${TELEGRAM_BOT_TOKEN}"`，设置写回保留这一引用。最终镜像采用 scratch，包含 Core、相邻 `packages/` 与 CA 证书；MCP stdio 服务器需要 shell、Node、Python 等运行时时，请自行构建包含所需运行时的镜像，并保留 UID 10001、可写数据卷与 `--init`。

## 管理组件
```fish
alonabot run --data-dir ./data
```

`run` 与库模式（`alonabot.Start`）共用 `internal/app` 的同一套 Core 装配；CLI 直接调用它，根包提供库模式的薄包装。组件经管理服务安装与启停，CLI 通过 `base_url` 上的管理 API 操作，从所选 `--data-dir` 自动读取持久凭据并直连本机 listen，Host 沿用 base_url；本机连接地址可用 `--admin-url` 覆盖：
```fish
alonabot install pkg.tar.gz --yes             # 直接安装本地归档
alonabot install alonabot-webui-admin --yes    # 从可执行文件旁 packages/<组件 ID>/ 安装
alonabot install --yes                        # 在当前目录或工程 bin/ 中选择唯一的本机包
alonabot install --source owner/repo --yes      # 从 GitHub Release API 选择平台归档
alonabot install --source AlonaBot/alonabot-webui --yes
alonabot enable alonabot-webui-admin
alonabot install --source AlonaBot/alonabot-agent --yes
alonabot install --source AlonaBot/alonabot-adapter-telegram --yes
alonabot install --source AlonaBot/alonabot-plugin-help --yes
alonabot enable alonabot-plugin-help-official
alonabot enable alonabot-plugin-hello
alonabot list
alonabot uninstall alonabot-plugin-hello --purge-data --yes      # 或 --purge；不带选项时只卸载
alonabot adapter instance create alonabot-adapter-telegram primary --config instance.json
alonabot admin add telegram:123456
```

`install` 的位置参数为已存在的普通文件时使用该归档，其他值按组件 ID 查找可执行文件旁的 `packages/<组件 ID>/`；缺失时列出可用的官方组件 ID。不带位置参数时，当前目录为作者工程便搜索 `bin/`，其他情况搜索当前目录；也检查直接包子目录。按归档名选择本机平台资产，静态 WebUI 使用 `any/any`，恰好一个候选才安装；零个或多个时列出搜索目录和候选后退出。发现阶段只读，安装经管理 API 完成。

`--archive` 可显式指定归档。位置参数、`--archive` 和 `--source` 互斥；来源确认、归档预检及 `--yes`、`--replace` 沿用管理服务。浏览器上传安装或更新同样只需归档文件。历史 Release 清理旁车后，旧 Core 的在线安装与更新需要先升级至采用此流程的版本，发布顺序见[发行维护](releasing.md)。

help 插件启用后，在聊天中发送 `/whoami`，把回复中的用户字符串作为 `admin add` 的参数；回复中的实例名用于核对来源。

管理端点与配置读写详见 [管理 API](deployment.zh-CN.md#管理-api)。

### 导出诊断包
```fish
alonabot diagnose --data-dir ./data --since 24h -o diagnose.zip
```

默认收集近 24 小时的日志、当前设置和 Active 配置、组件与环境信息，并开启脱敏。`--no-logs`、`--no-settings`、`--no-components` 可分别省略一组，至少保留一组。`-o` 也可写作 `--output`；省略时在当前目录生成 `alonabot-diagnostics-<UTC 时间>.zip`。

在线使用管理 API 和 Bearer token，自动读取 `<data_dir>/admin_token`。默认直连本机 listen 并保留 base_url Host；`--admin-url` 覆盖本机连接地址。连接被拒绝后尝试停机导出；原目录的 core.lock 须未被占用。离线导出只读原目录。

缺少可用的 Adapter 实例 schema 时，默认遮盖该组实例配置的全部叶值，保留结构和完整的 `${环境变量}` 表达式。`--no-redact` 会导出凭证明文；日志可能包含聊天内容，提交 issue 前请先检查归档。

## Core 设置与 HTTP 入口

`run` 取得 core.lock 后读取 `<data_dir>/core_config.json`，Core 设置统一从这个 JSON 对象读取；文件不存在时以 0600 权限写入缺省内容。JSON 语法、重复键或根节点错误时，以 usage 退出码报告键路径与行列号。未知键保留并忽略，类型或取值错误在运行副本中校正，记录路径与原因。

键与缺省值见 [插件指南](plugin-guide.zh-CN.md)；`base_url` 缺省为 `http://127.0.0.1:7650`，管理入口与插件路由共用这个地址。`admins`、`command_prefixes`、`reply`、`default_locale`、`log.sources` 与 `adapters` 已支持解析与保存；`command_prefixes_by_instance` 按未知键保留原文并忽略，管理员 prune 时清理。按账号区分前缀时，为该实例指定缺省配置文件，再经平台 API 保存前缀覆盖；旧 `/api/v1/core/command-prefixes` 路由已移除。本地 `run` 按监听、日志、代理与 Base URL 装配。

整个字符串值写成 `${NAME}` 时，读取时展开环境变量；NAME 由 `A-Z`、`a-z`、数字与下划线组成，首字符为字母或下划线。变量缺失时该值按不合法处理，恢复为缺省值并记校正记录。

`(*settings.File).Update` 在未展开的文档上编辑，校验后以 0600 临时文件替换原文件，保留其他键的顺序与 `${...}` 写法，新键追加到所在对象末尾；写回使用两个空格缩进并以换行结束。文件没有格式版本：不是合法 JSON 对象时拒绝启动，其中不合法的值按配置 schema 校正，不改写文件。

插件由 SDK 先执行 `OnConfigLoad(ctx, raw)`，再补值与校正；`GET /api/v1/plugins/{id}/config` 返回 Active 的 `corrections` 和 `unknown`。`POST /api/v1/plugins/{id}/config/prune` 从 Active 删除无效键并保存为 Pending，随后调用管理 API 的 reload 生效。开发组件也按此顺序清理：reload 时原子更新其编辑文件，成功启动后清除 Pending；之后的 restart 保持清理结果。

Core 设置的记录由 `GET /api/v1/core/settings/corrections` 读取，清理由 `POST /api/v1/core/settings/prune` 写回。Adapter 实例的 GET 响应也带 `corrections` 与 `unknown`，`POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune` 只删除该实例的无效键，保留其他值与 `${...}` 原文；实例校正与 required 检查独立进行，失败实例保留可修订的配置，其余实例继续启动。端点与生效方式见 [管理 API](deployment.zh-CN.md#管理-api)。

配置准备沿用 `internal/tune` 的 `PHASE_CONFIG`（缺省 60s），覆盖 OnConfigLoad、补值与校正及 required 校验；`PHASE_MIGRATION`（缺省 15m）用于数据库迁移，参数表见 `alonabot tune list`。

部署运行中，手改设置后重启 Core 生效；[开发运行](development.zh-CN.md#开发运行与热部署)按变化刷新设置与开发实例，监听地址等启动时装配的设置仍需重启 Core。`--listen`、`--log-level`、`--log-format` 只在命令行显式给出时覆盖设置，只对本次运行有效，文件内容保持不变：
```fish
alonabot run --data-dir ./data --listen 127.0.0.1:8080
```

用 `alonabot logs --level critical` 只看严重故障，或用 `alonabot run --log-level critical` 为本次启动提高全局输出阈值。运行中用 `alonabot log-level global critical` 保存全局阈值，用 `alonabot log-level <组件 ID> critical` 单独设置来源。查询筛选与输出设置见 [管理 API](deployment.zh-CN.md#管理-api)。

Core 在 `listen` 地址（缺省 `127.0.0.1:7650`）上提供唯一的公开 HTTP 入口。请求 Host 的主机名须匹配 `base_url`：`/plugins/{id}/...` 转发给该插件当前 Session 的路由，其余路径交给管理处理器；其他 Host 返回 421。缺省设置下，lifecycle 示例的私有页面地址为 `http://127.0.0.1:7650/plugins/demo/page`。listen 为非 loopback 地址时，`base_url` 必须是 https，或者配置非空的 `trusted_proxies`（由可信反代终止 TLS），设置文件与 `--listen` 都按此校验。

插件路由的全部响应（含 Core 的错误响应）统一设置 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff`，覆盖插件的同名头。Core 清除插件设置的全部 `Set-Cookie` 和 `Access-Control-Allow-Credentials`（含 trailer），Cookie 只以 WARN 记录名称；缺失 Content-Type 时补 `application/octet-stream`。HTML 页面按此沙箱策略返回。

设置了 `proxy` 时，Core 与子进程使用它作为 `HTTP_PROXY`/`HTTPS_PROXY`；为空时原样传递环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。

首次启动统一以 0600 权限持久保存 `<data_dir>/admin_token`，仅输出管理地址和文件路径。本机 WebUI 自动展示待确认的 token；复制保存后点击“我已保存”。远程浏览器可显式读取：
```fish
alonabot admin-token show --data-dir ./data
```

`show` 直接读取持久凭据，可在 Core 运行时执行。普通登录使用 `POST /api/auth/login`，要求受控 Origin 和 JSON token，成功返回管理 Cookie 与 CSRF。登录和确认保留文件。

私有插件路由复用管理认证：有效管理 Cookie 或 Bearer token；Cookie 认证的非 GET/HEAD 请求还须携带 `X-Alonabot-CSRF` 与受控 Origin，校验失败返回 401。公开路由无需管理会话，由插件按其声明的回调来源校验规则处理。转发给插件前，Core 清除管理 Cookie、Authorization 与 CSRF 头。

启动前可运行 `alonabot admin-token init --data-dir ./data`。凭据文件遗失时显式执行 `admin-token reset`，旧凭据与会话随之失效，新代次重新待确认；init/reset 输出文件位置，无需 TTY，均须先停止 Core。

Core 运行且持有当前 token 时，可通过 `POST /api/auth/token/rotate` 在线轮换。响应和持久文件同步更新；旧会话全部撤销，新代次重新待确认。CLI 自动读取更新后的文件。

调优参数在启动时解析；修改后重启生效。`default` 的前缀与 `reply` 属于 `core_config.json`；配置文件覆盖集合沿用固定 4 MiB 帧上限，`core.profile.*` 与 `sdk.profile.set_overrides` 沿用普通 RPC 的期限与 Session 预算（`internal/tune/params.go`）。查看全部调优参数及当前生效值：
```fish
alonabot tune list
alonabot --args REQUEST_TIMEOUT=2s tune list
alonabot run --data-dir ./data --args REQUEST_TIMEOUT=2s --args ./tuning.txt
```

`--args` 可重复，放在子命令前后均生效，后出现的覆盖先出现的。直接参数支持逗号分隔的 `KEY=VALUE`；文件每行一项，空行和以 `#` 开头的注释行会被忽略。时长使用 `2s` 等 Go 时长格式；大小使用 `B`、`KiB`、`MiB`、`GiB`，`MB` 等十进制单位会报错。`run` 将最终覆盖值及出处记为 WARN。

`OPTIONS_PAGE` 的默认值和范围见 [插件指南](plugin-guide.zh-CN.md)，可用 `alonabot tune list` 查看；聊天目录每页的上限取当前 Core 生效值的十倍，请求的 `limit` 仍缺省为 50。

## 插件休眠

已安装 plugin/agent 从 SUSPENDED 按命令/事件、Service 与 HTTP 请求恢复原代次。Core 唯一 Waker 在 ACTIVATED、摘要核对及全部绑定后开放交付；等待期间沿用接纳时的绝对期限。同代次 RESUME 复用配置与 DB，跳过数据库迁移回调；配置仍执行加载、补值/校正与 required 校验，完整结果须与原快照一致。开发组件沿用当前绑定与原生命周期。

`internal/tune` 统一登记 `WAKE_QUEUE_REQUESTS=256`（≥1）与 `WAKE_QUEUE_BYTES=16MiB`（4MiB～`CORE_MAX_PAYLOAD`）；等待 payload 同时占用 Core 总预算。可经 `--args` 覆盖，名称、单位、上下界及来源错误沿用 现行指南。HTTP 等待只占条数，正文在交付后流式读取；唤醒失败分别映射为 429（预算满）、504（期限到达）或 503（其他失败）。自动空闲策略：`sleep.mode` 缺省 `always_on`，`sleep.idle_timeout` 缺省 `15m`，`sleep.plugins` 缺省为空；覆盖接受 `never` 或正时长，删除覆盖恢复自动准入。`SLEEP_MIN_UPTIME=5m`（>0）、`SLEEP_COLD_START_WINDOW=1h`（>0）与 `SLEEP_COLD_START_MAX=3`（≥1）统一登记，窗口内冷启动超过上限后固定常驻一个窗口，期间启动保持原冷却截止。 Scheduler 自然到期唤醒，复用同一 Waker 与上述等待预算；payload 按完整 HandlerRequest 的编码字节计入。静态 deadline 为原 scheduled_at 加 `JOB_TIMEOUT=15m`（>0），动态从取得并发名额开始计算，冷启动计入原期限；等待继续占用原运行槽与 `TASK_CONCURRENCY=16` 名额（≥1 且 ≤ SESSION_MAX_INFLIGHT）。

管理 CLI 可直接查询和调整逐插件策略（复用 `--data-dir` 与 `--admin-url`）：
```fish
alonabot sleep-policy
alonabot sleep-policy my-plugin
alonabot sleep-policy my-plugin never
alonabot sleep-policy my-plugin 15m
alonabot sleep-policy my-plugin auto
```

无参数返回全局设置与全部已安装 plugin/agent 的策略和运行状态；单 ID 返回对应对象。`auto` 删除覆盖，`never` 与正时长立即写回设置并生效。全局模式与时长通过[管理 API](deployment.zh-CN.md#管理-api) 的 `PUT /api/v1/core/sleep` 修改，两项必填。管理读取保持休眠；需要路由而尚无有效已提交快照时，策略查询返回 412 / `FAILED_PRECONDITION`，全局设置与 runtime 可独立读取。保存后的恢复失败保留实际错误；重新 GET 核对已保存设置。

唤醒与调度的实现、测试入口见[架构地图](../ARCHITECTURE.md)；开发验证见[源码开发](development.zh-CN.md#验证改动)。

## 备份与恢复

整机备份前先停止 Core 和全部插件进程，再复制完整的 `data_dir`。备份须同时覆盖 Core 元数据、程序包、插件配置及各代次的受管 SQLite 数据库。单独备份 `state.sqlite` 只保存 Core 元数据，不包含插件业务数据。

插件自管的外部数据库与框架目录之外的文件，按插件的维护说明单独备份。恢复时保持 Core 和插件停止，还原完整的 `data_dir`，并同步恢复插件自管的外部数据，再启动 Core。恢复旧备份会丢弃备份之后的新写入，执行前须确认要恢复的数据时点。

## 管理 API

使用前文的受控 `base_url` 与认证方式。Core API 主版本 1 位于 `/api/v1/`，认证位于 `/api/auth/`。JSON 写请求使用 `Content-Type: application/json`。CLI 发送 `Authorization: Bearer <管理 token>`；浏览器使用管理 Cookie，GET/HEAD 之外再带受控 `Origin` 与 `X-Alonabot-CSRF`。`POST /api/auth/login` 接收 `{"token":"…"}`，返回 `{"csrf_token":"…"}` 并设置 Cookie；刷新后用 `GET /api/auth/session` 恢复 CSRF，`POST /api/auth/logout` 撤销会话。登录和确认保留持久 token 文件。

管理错误格式为 `{"code":"FAILED_PRECONDITION","safe_message":"…"}`。HTTP 对应 400 INVALID_ARGUMENT、403 PERMISSION_DENIED、404 NOT_FOUND、409 ALREADY_EXISTS 或陈旧配置、412 FAILED_PRECONDITION、501 UNSUPPORTED、503 UNAVAILABLE、500 INTERNAL/其他错误码。认证错误为纯文本。未知管理路由先认证：无凭据为 401，认证后为 JSON 404。需确认的操作返回 409 与 `{"affected":["…"]}`：展示全部对象，管理员同意后重发原请求，install 写 `Confirm: true`，动作与实例创建写 `confirm: true`。解析前核对实际状态和 Content-Type。

| 操作 | 端点与请求 |
| --- | --- |
| Core 身份与组件 | `GET /api/v1/core`、`GET /api/v1/packages` |
| 安装 | `POST /api/v1/packages/install`，使用导出 `InstallRequest` 字段 `Archive`、`Source`、`ID`、`Confirm`、`Replace`，见[管理类型](../internal/core/manage/install.go) |
| 启用、停用、重启、重载 | `POST /api/v1/packages/{action}`，`{"id":"demo"}`；disable 可能要求 `confirm` |
| 更新 | `POST /api/v1/packages/update`，`{"id":"demo","confirm":false,"install":{"Source":"owner/repo"}}` |
| 回退 | `POST /api/v1/packages/rollback`，`{"id":"demo","generation":"retained-generation","confirm":false}` |
| 固定版本 | `POST /api/v1/packages/pin`，`{"id":"demo","constraint":"1.2.3"}`，空串解除 |
| 卸载 | `POST /api/v1/packages/uninstall`，`{"id":"demo","purge":0,"confirm":false}`，清理前展示实际对象 |
| Active 配置 | `GET /api/v1/plugins/{id}/config`，查看 `corrections` 与 `unknown` |
| 编辑 Pending | `PUT`/`PATCH /api/v1/plugins/{id}/config/pending`，带 `base_generation` 和完整 `config` 或 RFC 7396 `patch`；`DELETE` 丢弃 Pending |
| 导入与应用 | `POST /api/v1/plugins/{id}/config/import` 将编辑文件导入 Pending；`POST /api/v1/plugins/{id}/reload` 准备并提交 |
| Adapter 实例 | `GET /api/v1/adapter/instances`；`POST /api/v1/adapter/instances/{create,start,stop,delete}`；create 用 `package_id`、`instance_id`、`config`、可选 `confirm`，其余用 `instance_id` |
| 实例配置 | `GET /api/v1/adapters/{adapter_id}/instance-schema`；`GET`/`PUT`/`DELETE /api/v1/adapters/{adapter_id}/instances/{instance_id}`；PUT 为 `{"config":{}}` |
| 日志 | `GET /api/v1/logs/recent`、`/stream`（SSE）、`/history?date=YYYY-MM-DD`，可筛选 `source`、`level`、`trace`；`GET`/`PUT /api/v1/logs/levels` 用 `{"global":"info","sources":{}}` 部分更新 |
| 配置文件 | `GET`/`POST /api/v1/profiles`；`PATCH`/`DELETE /api/v1/profiles/{id}`；指定、插件、平台、工具覆盖见[配置文件路由](../internal/core/manage/profile_api.go) |
| 工具策略 | `GET`/`PUT /api/v1/tools/policies`，写 `{"plugin_id":"demo","service":"echo","policy":"deny"}` 或原子批量 `{"items":[…]}`；allow/confirm/deny/inherit，inherit 删除覆盖 |
| 会话 | `GET /api/v1/plugins/{id}/conversations` 返回摘要；`DELETE /api/v1/plugins/{id}/conversations/{key}` 清除完整不透明键 |
| 插件数据 | `/api/v1/plugins/{id}/data/` 下的 tables/rows、kv、files、export、clear，见[数据处理器](../internal/core/manage/data_api.go) |
| 诊断包 | `POST /api/v1/diagnostics`，缺省 `{"since":"24h","logs":true,"settings":true,"components":true,"redact":true}`，返回 no-store ZIP |

本地 JSON 安装提供 Core 主机上的 `Archive` 路径，身份和来源取自归档内 `plugin.yaml`；在线安装只提供 `Source: "owner/repo"`，Core 从最新 Release API 选择平台归档。可选 `ID` 约束期望身份。首次来源信任和同类替换按需确认；后续更新与保留记录重装继续核对来源和身份。`Package.PublisherKeyID`、`Package.Signed` 保留历史记录含义；新安装及成功更新或重装后分别为空和 false。

浏览器对 install/update 端点发送 `multipart/form-data`：必填一个文件 part `archive`；可选文本 `confirm=true|false`；仅 install 可带 `replace=true|false`，仅 update 必填文本 `id`。文件 part 带 `filename`，文本 part 省略；未知或重复 part、缺文件或更新 ID、错误布尔值和截断请求返回 400 / INVALID_ARGUMENT。认证先于上传解析；归档流式写入独占临时目录，沿用安装大小与预检限制，调用完成后清理。确认响应为 409 与完整 `affected`，确认后重新上传归档并带 `confirm=true`。预检在修改旧组件前完成，来源绑定不符返回 403 / PERMISSION_DENIED；更新沿用身份、pin 与依赖检查并保留用户数据。

Pending 编辑要求当前 `base_generation`，陈旧编辑为 409。读取时秘密用 `{"$secret_set":true}` 遮盖；只在需要修改处提交新明文，勿提交掩码或 Core 秘密引用。reload 建立独立受管 SQLite 候选代次，restart 保持 Active；提交后的 OnReady 失败保留新 Active 并进入 Degraded。`GET /api/v1/plugins/{id}/generations` 列出保留代次。开发组件的 update/rollback/uninstall 为 412。

会话键为 `<会话>@<实例>`，成员独立会话再加 `@<用户>`。将完整键百分号编码为一个路径段，服务端只解码一次，保留斜杠、冒号、反斜杠与点段；键须为 1～256 字节 UTF-8。数据浏览可选择只读保留代次。`/data/clear` 必须带 `{"confirm":true}`，删除可写表行（含 KV）和插件文件，保留表结构、迁移与目录。KV 遵循[存储规则](plugin-guide.zh-CN.md#保存少量状态)，键编码为不透明路径段。文件路径保持相对路径，拒绝穿越和符号链接；清理或编辑文件独立于代次回退。

休眠使用 `GET`/`PUT /api/v1/core/sleep`、`GET`/`PUT`/`DELETE /api/v1/plugins/{id}/sleep-policy` 及 `GET /api/v1/plugins/{id}/runtime`。逐插件 PUT 为 `{"policy":"never"}` 或正时长，DELETE 恢复自动策略；读取保持休眠。缺省值、预算和失败行为见[插件休眠](#插件休眠)。

### 更新日志与 WebUI 后台

认证后 `GET /api/v1/core/changelog` 返回 `{"locale":"en-US","entries":[{"version":"0.1.0","markdown":"…"}]}`。来源为正式 `changelog/v<SemVer>/{en,zh-CN}.md`，排除 unreleased，en 映射 en-US；按 `Accept-Language` 精确或主语言匹配，缺省或未匹配回退英文。条目按 SemVer 倒序，版本无 `v`，未知日期省略 `date`，保留 Markdown。`GET /api/v1/packages/{id}/changelog` 读取安装包：单文件 CHANGELOG.md 排除 Unreleased，locale 为 en-US；目录形式逐版本按偏好语言精确、规范标签、主语言匹配，最终回退 en-US.md，同级取文件名字典序最小者。未安装或 dev 组件为 404；无正式条目返回 en-US 与空 entries。

认证后 `GET /api/v1/packages/{id}/icon` 提供图标，`GET /api/v1/packages/{id}/examples/{path}` 只提供 Active 包引用的配置图片。插件页面在包的 `pages` 中列出，经 `/plugins/{id}` 在 WebUI 沙箱打开，见[页面套件](plugin-guide.zh-CN.md#插件页面)。

官方 WebUI 组件包含真实后台进程。认证前缀 `/plugins/alonabot-webui-admin` 下有 `/stats?range=24h|7d|30d`、`/users`、`/store`、`/updates?repo=owner/repo`、`/changelog/seen?component=core`。`PUT /changelog/seen` 仅接收 `{"component":"core","version":"0.1.0"}`，按 SemVer 单调持久化；GET 返回 `{"version":null}` 或版本。仅复制静态 `dist` 时这些后台路由不可用。模型、人格、MCP、技能、记忆和任务页面由 Agent 提供，见其 [HTTP 实现](https://github.com/AlonaBot/alonabot-agent/blob/main/pages.go)。

完整现行路由和请求/响应类型由[管理包](../internal/core/manage)维护，[管理集成测试](../tests/e2e/manage_e2e_test.go)通过真实 Core 验证。

### 组件商店

官方 WebUI 将自己的数据保存在其受管 SQLite 中。下列后台路由均相对于 `/plugins/alonabot-webui-admin`，沿用管理认证与 CSRF，响应带 `Cache-Control: no-store`。组件安装与更新仍调用 Core 管理 API。

使用 `GET /store?source=AlonaBot/plugin-store&role=plugin&q=&sort=stars&offset=0&limit=9&refresh=false`。`source` 必须为 `owner/repo`，显式指定非负 `offset` 与正整数 `limit`；role 为空或 plugin/adapter/agent/webui。响应包含 `source`、`fetched_at`、`stale`、`items`、`total` 与 `counts`（all 及四种角色）。items 包含 `repo,role,tags,id,name,version`，可选 `description,author,stars,icon_url`；stars 缺省表示未知，0 保持为 0。默认源为 `AlonaBot/plugin-store`；源列表在浏览器持久化，主动清空后保持空列表。

搜索对 name、description、repo、author 与空格连接的 tags 做不区分大小写的子串匹配，缺失值按空串处理。counts 在搜索后、role 筛选前计算，total 在角色筛选后计算。默认或空 `sort=stars` 将已知星标数降序排列，未知值随后；`sort=name` 按名称不区分大小写升序，同值保留索引顺序，其他排序值返回 400 / INVALID_ARGUMENT。WebUI 每页显示九张卡片，搜索、角色或排序变化后回到第一页。

`refresh` 仅接受字面量 `true` 或 `false`，空值也返回 400 / INVALID_ARGUMENT。省略或 true 会刷新并持久化完整成功快照；false 且有成功快照时，在快照上搜索、计数、排序与分页，全程无远端请求；无快照时执行首次抓取。快照读取返回 `stale=false` 和原 `fetched_at`。进入商店、切换源及支持键盘的手动刷新使用 true，搜索、筛选、排序与分页使用 false。刷新失败保留成功快照和原时间，返回 `stale=true` 及完整错误；无快照的一般失败返回 500 / INTERNAL，源索引缺失返回 404 / SOURCE_INDEX_MISSING。

`GET /updates?repo=owner/repo` 每次查询刷新，返回 `repo,version,release_url,markdown,fetched_at,stale`。最新 Release 标签去掉前导 v 后须为 SemVer，Markdown 保留 Release 正文。latest Release 的 404 返回 404 / NO_RELEASE，由 WebUI 的 `getLatestRelease` 映射为 null。源与 Release 缓存按小写 source/repo 隔离。链接仅允许 HTTP/HTTPS/mailto，HTML 按文本显示。更新预览使用组件绑定来源，更新后说明来自已安装包内 changelog，已读版本只推进到实际展示的版本。

收录时，`GET /store/check?repo=owner/repo` 读取最新 Release API 及其首个平台归档，保留持久缓存原状；核对仓库、Release 与资产 URL、标签 SemVer、资产名称和长度，以及包内 manifest 的 ID、版本与来源。返回 `repo,role,id,name,version,release_url`，可选 description/author，及保持顺序的 manifest `warnings`（空数组也返回）。响应 repo 取已验证 manifest Source 去掉 `github.com/` 后的规范大小写。Core 安装预检继续检查平台、build-info、资源及归档完整性。仅 latest Release 的 404 映射为 NO_RELEASE；坏资产 404 及其他失败返回 500 / INTERNAL，保留实际错误。

源索引为 `{"version":1,"plugins":[{"repo":"owner/repo","role":"plugin","tags":["utility"]}]}`，tags 可省略。检查后，用户填写标签，用检查所得规范 repo，在浏览器打开源仓库 `issues/new`，携带 `template=submit-plugin.yml` 以及经 URL 编码的 `repo`、`role`、`tags`、`release_url`，由浏览器提交 GitHub Issue。

每个远端 HTTP 请求的期限为 20 秒，覆盖响应体；串行依赖请求总耗时可更长，聚合刷新响应调用方取消。初始 URL 与每次重定向均检查 HTTPS 主机和路径，只允许支持的 GitHub API/Release/CDN 地址。组件安全上限为：元数据 1 MiB、索引 4 MiB、归档 64 MiB、每轮展开 256 MiB（最多两轮）、单文件 128 MiB、每轮 2,048 条、保留的 manifest 256 KiB、保留的图标 1 MiB。超限保留 `exceeds ... bytes/entries` 错误并沿用同样的缓存规则。当前实现见 [WebUI store.go](https://github.com/AlonaBot/alonabot-webui/blob/main/internal/admin/store.go)。
