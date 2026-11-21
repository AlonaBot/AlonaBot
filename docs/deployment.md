# 安装与部署

本文面向运行 AlonaBot 的管理员，说明安装、认证、配置、组件管理与恢复。源码开发见[开发指南](development.md)，构建发行物见[发行维护](releasing.md)。

## 选择安装方式

从 [v0.1.0 Release](https://github.com/AlonaBot/AlonaBot/releases/tag/v0.1.0) 下载正式资产。整合包包含 Core、WebUI 与 Agent；Adapter 和 help 插件按需安装。具体平台实跑与已知限制见[发布说明](../changelog/v0.1.0/zh-CN.md)。

| 方式 | 适用场景 |
| --- | --- |
| 发行归档 | 手动运行 Core，适用于五个发行目标 |
| install.sh | Linux systemd 系统服务与 macOS 当前用户 LaunchAgent |
| Docker Compose | Linux amd64/arm64 容器，持久化数据卷 |

## 手动运行发行归档

下载同一版本的归档，解压后保持 `alonabot` 与 `packages/` 相邻，在该目录运行：
```fish
./alonabot run --data-dir ./data
```

后文管理命令假定 `alonabot` 已加入 PATH；从解压目录操作时使用 `./alonabot`。Windows 使用 `alonabot.exe`。首次启动预装 WebUI 与 Agent，启用 WebUI。打开 `http://127.0.0.1:7650`，按[管理认证](#core-设置与-http-入口)领取首次 token 并登录。按需安装 Telegram 或 OneBot 11 Adapter 后配置实例；配置 Provider 与模型后再启用 Agent。

## 系统服务安装与升级

在支持 systemd 的 Linux 上以 root 运行；macOS 在当前用户的登录会话运行。支持两种系统的 amd64、arm64。先从同一 Release 下载脚本，再安装指定版本（以下命令使用 fish，把 `VERSION` 换成发行版本号，省略前导 `v`）：
```fish
set version VERSION
curl --fail --location --proto '=https' --proto-redir '=https' \
  -o install.sh https://github.com/AlonaBot/AlonaBot/releases/download/v$version/install.sh
# Linux
sudo sh install.sh --version $version
# macOS
sh install.sh --version $version
```

省略 `--version` 会先通过 `/latest` 固定一个 Release，再下载该版本的归档与 `SHA256SUMS`。`ALONABOT_RELEASE_BASE_URL` 可覆盖 Release 基址，基址和所有重定向均须使用 HTTPS。脚本先检查平台、权限和系统工具，完成下载、校验和临时解压后才停服并替换 `alonabot` 与完整的 `packages/`。

Linux 使用 `/opt/alonabot/`、数据目录 `/var/lib/alonabot` 和系统服务 `alonabot.service`；服务以系统用户 `alonabot` 运行并启用开机启动。macOS 使用 `~/Library/Application Support/AlonaBot/`、其中的 `data/` 和用户 LaunchAgent `io.github.alonabot`，登录后启动。工作目录设为程序目录，启动参数为 `run --data-dir <完整数据路径>`。

有控制终端时，脚本在停服后以服务身份运行 `admin-token init --if-missing`，stdin 与 stderr 连接 `/dev/tty`；新 token 仅显示一次，已有凭证保持原值。经管道执行同样使用控制终端。无控制终端时跳过初始化，服务照常启动，Core 首次启动写入 `admin_token`。文件存在时按脚本输出的完整命令领取：
```fish
# Linux，以服务身份读取
sudo runuser -u alonabot -- /opt/alonabot/alonabot admin-token show --data-dir /var/lib/alonabot
# macOS，以当前用户读取
"$HOME/Library/Application Support/AlonaBot/alonabot" admin-token show --data-dir "$HOME/Library/Application Support/AlonaBot/data"
```

首次登录成功后 Core 删除 `admin_token`。再次运行安装命令即升级，保留数据、`core_config.json`、状态库、已安装组件与未领取的 token。内置包的预装、默认启用与可用更新由 Core 处理。
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
  -o docker-compose.yml https://github.com/AlonaBot/AlonaBot/releases/download/v0.1.0/docker-compose.yml
docker compose up -d
docker compose exec alonabot alonabot admin-token show --data-dir /data
```

打开 `http://127.0.0.1:7650`，使用读出的 token 登录。首次启动无需 TTY，Core 将一次性 token 写入 `/data/admin_token`，首次登录成功后删除该文件。镜像支持 `linux/amd64` 与 `linux/arm64`，compose 按宿主平台选择。compose 中固定镜像版本，与所选 Release 的版本保持一致。

镜像以 `10001:10001` 运行，命名卷挂载到 `/data`，首次创建卷时继承镜像中该目录的所有者。改用宿主目录挂载时，先授予 UID/GID 10001 读写和遍历权限；已有卷也需保证这一权限。保留该卷即可保留配置、状态库和组件数据。

端口只发布到宿主的 `127.0.0.1`。镜像设置 `ALONABOT_CONTAINER=1` 时，缺省监听为 `0.0.0.0:7650`，`base_url` 仍为 `http://127.0.0.1:7650`；显式 `listen` 优先。该容器例外实际生效时，启动日志输出一次 WARN 提醒宿主本机发布规则。远程管理可使用 `ssh -N -L 7650:127.0.0.1:7650 user@host`，再打开本机地址；HTTPS 反代部署按 [Core 设置与 HTTP 入口](#core-设置与-http-入口) 配置 `base_url` 和 `trusted_proxies`。

按需安装 Adapter 后，在 compose 服务的 `environment` 中传入 `TELEGRAM_BOT_TOKEN` 后，`/data/core_config.json` 中的实例凭证可写成 `"token": "${TELEGRAM_BOT_TOKEN}"`，设置写回保留这一引用。最终镜像采用 scratch，包含 Core、相邻 `packages/` 与 CA 证书；MCP stdio 服务器需要 shell、Node、Python 等运行时时，请自行构建包含所需运行时的镜像，并保留 UID 10001、可写数据卷与 `--init`。

## 管理组件
```fish
alonabot run --data-dir ./data
```

`run` 与库模式（`alonabot.Start`）共用 `internal/app` 的同一套 Core 装配；CLI 直接调用它，根包提供库模式的薄包装。组件经管理服务安装与启停，CLI 通过 `base_url` 上的管理 API 操作，token 取自 `ALONABOT_ADMIN_TOKEN` 或 `--token-file`，管理地址可用 `--admin-url` 覆盖：
```fish
alonabot install pkg.tar.gz --yes             # release.json、签名和公钥缺省取归档同目录
alonabot install alonabot-webui-admin --yes    # 从可执行文件旁 packages/<组件 ID>/ 安装
alonabot install --yes                        # 在当前目录或工程 bin/ 中选择唯一的本机包
alonabot install --source owner/repo --yes      # 从 GitHub Release 下载 release.json、可选的 .sig 与 release.pub
alonabot install --source AlonaBot/alonabot-adapter-telegram --yes
alonabot install --source AlonaBot/alonabot-plugin-help --yes
alonabot enable alonabot-plugin-help-official
alonabot enable alonabot-plugin-hello
alonabot list
alonabot uninstall alonabot-plugin-hello --purge-data --yes      # 或 --purge；不带选项时只卸载
alonabot adapter instance create alonabot-adapter-telegram primary --config instance.json
alonabot admin add telegram:123456
```

`install` 的位置参数为已存在的普通文件时使用该归档，其他值按组件 ID 查找可执行文件旁的 `packages/<组件 ID>/`；缺失时列出可用的官方组件 ID。不带位置参数时，当前目录为作者工程便搜索 `bin/`，其他情况搜索当前目录；也检查直接包子目录的清单。按 `release.json` 选择本机平台资产，静态 WebUI 使用 `any/any`，恰好一个候选才安装；零个或多个时列出搜索目录和候选后退出。发现阶段只读，安装经管理 API 完成。

`--archive` 仍可显式指定归档；`--release`、`--signature` 与 `--public-key` 分别覆盖元数据文件。签名和公钥未指定时取 release.json 同目录的 `release.json.sig` 与 `release.pub`。位置参数、`--archive` 和 `--source` 互斥；`--source` 与本地元数据选项也互斥。来源与签名确认、归档预检及 `--yes`、`--replace` 沿用管理服务。

help 插件启用后，在聊天中发送 `/whoami`，把回复中的用户字符串作为 `admin add` 的参数；回复中的实例名用于核对来源。


管理端点与配置读写详见 [Admin API](../spec/ADMIN-API.md)。

### 导出诊断包
```fish
alonabot diagnose --data-dir ./data --since 24h -o diagnose.zip
```

默认收集近 24 小时的日志、当前设置和 Active 配置、组件与环境信息，并开启脱敏。`--no-logs`、`--no-settings`、`--no-components` 可分别省略一组，至少保留一组。`-o` 也可写作 `--output`；省略时在当前目录生成 `alonabot-diagnostics-<UTC 时间>.zip`。

在线使用管理 API 和 Bearer token：优先读 `--token-file`，其次取 `ALONABOT_ADMIN_TOKEN`，两者均未指定时尝试 `<data_dir>/admin_token`。`--admin-url` 可覆盖管理地址。连接被拒绝后尝试停机导出，无需 token；数据目录须已有可读的 `core.lock` 且锁未被占用。离线导出只读原目录，组件状态为 `Stopped`。

缺少可用的 Adapter 实例 schema 时，默认遮盖该组实例配置的全部叶值，保留结构和完整的 `${环境变量}` 表达式。`--no-redact` 会导出凭证明文；日志可能包含聊天内容，提交 issue 前请先检查归档。

## Core 设置与 HTTP 入口

`run` 取得 core.lock 后读取 `<data_dir>/core_config.json`，Core 设置统一从这个 JSON 对象读取；文件不存在时以 0600 权限写入缺省内容。JSON 语法、重复键或根节点错误时，以 usage 退出码报告键路径与行列号。未知键保留并忽略，类型或取值错误在运行副本中校正，记录路径与原因。

键与缺省值见 [CONTRACTS §15.1](../spec/CONTRACTS-v0.1.md#151-启动与-core_configjson)；`base_url` 缺省为 `http://127.0.0.1:7650`，管理入口与插件路由共用这个地址。`admins`、`command_prefixes`、`reply`、`default_locale`、`log.sources` 与 `adapters` 已支持解析与保存；`command_prefixes_by_instance` 按未知键保留原文并忽略，管理员 prune 时清理。按账号区分前缀时，为该实例指定缺省配置文件，再经平台 API 保存前缀覆盖；旧 `/api/v1/core/command-prefixes` 路由已移除。本地 `run` 按监听、日志、代理与 Base URL 装配。

整个字符串值写成 `${NAME}` 时，读取时展开环境变量；NAME 由 `A-Z`、`a-z`、数字与下划线组成，首字符为字母或下划线。变量缺失时该值按不合法处理，恢复为缺省值并记校正记录。

`(*settings.File).Update` 在未展开的文档上编辑，校验后以 0600 临时文件替换原文件，保留其他键的顺序与 `${...}` 写法，新键追加到所在对象末尾；写回使用两个空格缩进并以换行结束。文件没有格式版本：不是合法 JSON 对象时拒绝启动，其中不合法的值按 CONTRACTS §5.5 校正，不改写文件。

插件由 SDK 先执行 `OnConfigLoad(ctx, raw)`，再补值与校正；`GET /api/v1/plugins/{id}/config` 返回 Active 的 `corrections` 和 `unknown`。`POST /api/v1/plugins/{id}/config/prune` 从 Active 删除无效键并保存为 Pending，随后调用管理 API 的 reload 生效。开发组件也按此顺序清理：reload 时原子更新其编辑文件，成功启动后清除 Pending；之后的 restart 保持清理结果。

Core 设置的记录由 `GET /api/v1/core/settings/corrections` 读取，清理由 `POST /api/v1/core/settings/prune` 写回。Adapter 实例的 GET 响应也带 `corrections` 与 `unknown`，`POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune` 只删除该实例的无效键，保留其他值与 `${...}` 原文；实例校正与 required 检查独立进行，失败实例保留可修订的配置，其余实例继续启动。端点与生效方式见 [ADMIN-API §12～§14](../spec/ADMIN-API.md#12-组件配置pending-与秘密查看)。

配置准备沿用 `internal/tune` 的 `PHASE_CONFIG`（缺省 60s），覆盖 OnConfigLoad、补值与校正及 required 校验；`PHASE_MIGRATION`（缺省 15m）用于数据库迁移，参数表见 CONTRACTS §16。

部署运行中，手改设置后重启 Core 生效；[开发运行](development.md#开发运行与热部署)按变化刷新设置与开发实例，监听地址等启动时装配的设置仍需重启 Core。`--listen`、`--log-level`、`--log-format` 只在命令行显式给出时覆盖设置，只对本次运行有效，文件内容保持不变：
```fish
alonabot run --data-dir ./data --listen 127.0.0.1:8080
```

用 `alonabot logs --level critical` 只看严重故障，或用 `alonabot run --log-level critical` 为本次启动提高全局输出阈值。运行中用 `alonabot log-level global critical` 保存全局阈值，用 `alonabot log-level <组件 ID> critical` 单独设置来源。查询筛选与输出设置见 [ADMIN-API §8](../spec/ADMIN-API.md#8-日志读取与输出级别)。

Core 在 `listen` 地址（缺省 `127.0.0.1:7650`）上提供唯一的公开 HTTP 入口。请求 Host 的主机名须匹配 `base_url`：`/plugins/{id}/...` 转发给该插件当前 Session 的路由，其余路径交给管理处理器；其他 Host 返回 421。缺省设置下，lifecycle 示例的私有页面地址为 `http://127.0.0.1:7650/plugins/demo/page`。listen 为非 loopback 地址时，`base_url` 必须是 https，或者配置非空的 `trusted_proxies`（由可信反代终止 TLS），设置文件与 `--listen` 都按此校验。

插件路由的全部响应（含 Core 的错误响应）统一设置 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff`，覆盖插件的同名头。Core 清除插件设置的全部 `Set-Cookie` 和 `Access-Control-Allow-Credentials`（含 trailer），Cookie 只以 WARN 记录名称；缺失 Content-Type 时补 `application/octet-stream`。HTML 页面按此沙箱策略返回。

设置了 `proxy` 时，Core 与子进程使用它作为 `HTTP_PROXY`/`HTTPS_PROXY`；为空时原样传递环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。

首次启动且尚无管理凭证时，Core 自动生成管理 token：stdin 与 stderr 都是终端时只向 stderr 显示一次；系统服务、容器等非终端启动时，以 0600 权限写入 `<data_dir>/admin_token`，INFO 日志记录文件路径。读取文件中的 token：
```fish
alonabot admin-token show --data-dir ./data
```

`show` 直接读取文件，可在 Core 运行时执行，也可从非终端调用；文件不存在时返回英文错误。第一次登录成功后，Core 删除该文件。登录接口为 `base_url` 下的 `POST /api/auth/login`，要求受控 Origin、`Content-Type: application/json` 与正文 `{"token":"<token>"}`，成功后返回管理 Cookie 和 `csrf_token`；CLI 可用 `Authorization: Bearer <token>`。

私有插件路由复用管理认证：有效管理 Cookie 或 Bearer token；Cookie 认证的非 GET/HEAD 请求还须携带 `X-Alonabot-CSRF` 与受控 Origin，校验失败返回 401。公开路由无需管理会话，由插件按其声明的回调来源校验规则处理。转发给插件前，Core 清除管理 Cookie、Authorization 与 CSRF 头。

需要在启动前初始化凭证时，停止 Core 后在本机终端运行 `alonabot admin-token init --data-dir ./data`。token 遗失时用 `alonabot admin-token reset --data-dir ./data` 重新生成，旧 token 与全部管理会话随之失效。`init` 与 `reset` 都要求 stdin 与 stderr 是终端，只向 stderr 显示 token 一次；Core 运行中执行会因 core.lock 失败。

Core 运行且仍持有当前 token 时，可调用 `POST /api/auth/token/rotate` 在线轮换（[ADMIN-API §20](../spec/ADMIN-API.md#20-webui-需要的补充接口)）。从响应保存新 token，更新 CLI 使用的 `ALONABOT_ADMIN_TOKEN` 或 `--token-file`，并用新 token 重新登录。轮换会撤销全部旧管理会话。

调优参数在启动时解析；修改后重启生效。`default` 的前缀与 `reply` 属于 `core_config.json`；配置文件覆盖集合沿用固定 4 MiB 帧上限，`core.profile.*` 与 `sdk.profile.set_overrides` 沿用普通 RPC 的期限与 Session 预算（CONTRACTS §16、`internal/tune/params.go`）。查看全部调优参数及当前生效值：
```fish
alonabot tune list
alonabot --args REQUEST_TIMEOUT=2s tune list
alonabot run --data-dir ./data --args REQUEST_TIMEOUT=2s --args ./tuning.txt
```

`--args` 可重复，放在子命令前后均生效，后出现的覆盖先出现的。直接参数支持逗号分隔的 `KEY=VALUE`；文件每行一项，空行和以 `#` 开头的注释行会被忽略。时长使用 `2s` 等 Go 时长格式；大小使用 `B`、`KiB`、`MiB`、`GiB`，`MB` 等十进制单位会报错。`run` 将最终覆盖值及出处记为 WARN。

`OPTIONS_PAGE` 的默认值和范围见 [CONTRACTS §16](../spec/CONTRACTS-v0.1.md#16-调优参数)，可用 `alonabot tune list` 查看；聊天目录每页的上限取当前 Core 生效值的十倍，请求的 `limit` 仍缺省为 50。


## 插件休眠

已安装 plugin/agent 从 SUSPENDED 按命令/事件、Service 与 HTTP 请求恢复原代次。Core 唯一 Waker 在 ACTIVATED、摘要核对及全部绑定后开放交付；等待期间沿用接纳时的绝对期限。同代次 RESUME 复用配置与 DB，跳过数据库迁移回调；配置仍执行加载、补值/校正与 required 校验，完整结果须与原快照一致。开发组件沿用当前绑定与原生命周期。

`internal/tune` 统一登记 `WAKE_QUEUE_REQUESTS=256`（≥1）与 `WAKE_QUEUE_BYTES=16MiB`（4MiB～`CORE_MAX_PAYLOAD`）；等待 payload 同时占用 Core 总预算。可经 `--args` 覆盖，名称、单位、上下界及来源错误沿用 CONTRACTS §16。HTTP 等待只占条数，正文在交付后流式读取；唤醒失败分别映射为 429（预算满）、504（期限到达）或 503（其他失败）。自动空闲策略：`sleep.mode` 缺省 `always_on`，`sleep.idle_timeout` 缺省 `15m`，`sleep.plugins` 缺省为空；覆盖接受 `never` 或正时长，删除覆盖恢复自动准入。`SLEEP_MIN_UPTIME=5m`（>0）、`SLEEP_COLD_START_WINDOW=1h`（>0）与 `SLEEP_COLD_START_MAX=3`（≥1）统一登记，窗口内冷启动超过上限后固定常驻一个窗口，期间启动保持原冷却截止。 Scheduler 自然到期唤醒，复用同一 Waker 与上述等待预算；payload 按完整 HandlerRequest 的编码字节计入。静态 deadline 为原 scheduled_at 加 `JOB_TIMEOUT=15m`（>0），动态从取得并发名额开始计算，冷启动计入原期限；等待继续占用原运行槽与 `TASK_CONCURRENCY=16` 名额（≥1 且 ≤ SESSION_MAX_INFLIGHT）。

管理 CLI 可直接查询和调整逐插件策略（复用既有 `--data-dir`、`--admin-url` 与 `--token-file`）：
```fish
alonabot sleep-policy
alonabot sleep-policy my-plugin
alonabot sleep-policy my-plugin never
alonabot sleep-policy my-plugin 15m
alonabot sleep-policy my-plugin auto
```

无参数返回全局设置与全部已安装 plugin/agent 的策略和运行状态；单 ID 返回对应对象。`auto` 删除覆盖，`never` 与正时长立即写回设置并生效。全局模式与时长通过[管理 API §24](../spec/ADMIN-API.md#24-插件休眠策略与运行状态) 的 `PUT /api/v1/core/sleep` 修改，两项必填。管理读取保持休眠；需要路由而尚无有效已提交快照时，策略查询返回 412 / `FAILED_PRECONDITION`，全局设置与 runtime 可独立读取。保存后的恢复失败保留实际错误；重新 GET 核对已保存设置。

唤醒与调度的实现、测试入口见[架构地图](../ARCHITECTURE.md)；开发验证见[源码开发](development.md#验证改动)。

## 备份与恢复

整机备份前先停止 Core 和全部插件进程，再复制完整的 `data_dir`。备份须同时覆盖 Core 元数据、程序包、插件配置及各代次的受管 SQLite 数据库。单独备份 `state.sqlite` 只保存 Core 元数据，不包含插件业务数据。

插件自管的外部数据库与框架目录之外的文件，按插件的维护说明单独备份。恢复时保持 Core 和插件停止，还原完整的 `data_dir`，并同步恢复插件自管的外部数据，再启动 Core。恢复旧备份会丢弃备份之后的新写入，执行前须确认要恢复的数据时点。
