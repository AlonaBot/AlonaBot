# Core 管理 API 与运行机制

[English](management-api.md)

本文面向 WebUI、CLI 与组件开发者及 Core 贡献者，说明管理端点、认证、设置语义、安装与源码构建、休眠、下载镜像和日志。安装、运行与日常管理机器人见[用户文档](https://alonabot.github.io/docs/usage/install)。源码开发见[开发指南](development.zh-CN.md)，构建发行物见[发行维护](releasing.md)。

## Core 设置与 HTTP 入口

`run` 取得 core.lock 后读取 `<data_dir>/core_config.json`，Core 设置统一从这个 JSON 对象读取；文件不存在时以 0600 权限写入缺省内容。JSON 语法、重复键或根节点错误时，以 usage 退出码报告键路径与行列号。未知键保留并忽略，类型或取值错误在运行副本中校正，记录路径与原因。

键与缺省值见 [插件指南](https://alonabot.github.io/docs/tutorial/project)；`base_url` 缺省为 `http://127.0.0.1:7650`，管理入口与插件路由共用这个地址。`admins`、`command_prefixes`、`reply`、`default_locale`、`log.sources` 与 `adapters` 已支持解析与保存；`command_prefixes_by_instance` 按未知键保留原文并忽略，管理员 prune 时清理。按账号区分前缀时，为该实例指定缺省配置文件，再经平台 API 保存前缀覆盖；旧 `/api/v1/core/command-prefixes` 路由已移除。本地 `run` 按监听、日志、代理与 Base URL 装配。

整个字符串值写成 `${NAME}` 时，读取时展开环境变量；NAME 由 `A-Z`、`a-z`、数字与下划线组成，首字符为字母或下划线。变量缺失时该值按不合法处理，恢复为缺省值并记校正记录。

`(*settings.File).Update` 在未展开的文档上编辑，校验后以 0600 临时文件替换原文件，保留其他键的顺序与 `${...}` 写法，新键追加到所在对象末尾；写回使用两个空格缩进并以换行结束。文件没有格式版本：不是合法 JSON 对象时拒绝启动，其中不合法的值按配置 schema 校正，不改写文件。

插件由 SDK 先执行 `OnConfigLoad(ctx, raw)`，再补值与校正；`GET /api/v1/plugins/{id}/config` 返回 Active 的 `corrections` 和 `unknown`。`POST /api/v1/plugins/{id}/config/prune` 从 Active 删除无效键并保存为 Pending，随后调用管理 API 的 reload 生效。开发组件也按此顺序清理：reload 时原子更新其编辑文件，成功启动后清除 Pending；之后的 restart 保持清理结果。

Core 设置的记录由 `GET /api/v1/core/settings/corrections` 读取，清理由 `POST /api/v1/core/settings/prune` 写回。Adapter 实例的 GET 响应也带 `corrections` 与 `unknown`，`POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune` 只删除该实例的无效键，保留其他值与 `${...}` 原文；实例校正与 required 检查独立进行，失败实例保留可修订的配置，其余实例继续启动。端点与生效方式见 [管理 API](#管理-api)。

配置准备沿用 `internal/tune` 的 `PHASE_CONFIG`（缺省 60s），覆盖 OnConfigLoad、补值与校正及 required 校验；`PHASE_MIGRATION`（缺省 15m）用于数据库迁移，参数表见 `alonabot tune list`。

部署运行中，手改设置后重启 Core 生效；[开发运行](development.zh-CN.md#开发运行与热部署)按变化刷新设置与开发实例，监听地址等启动时装配的设置仍需重启 Core。`--listen`、`--log-level`、`--log-format` 只在命令行显式给出时覆盖设置，只对本次运行有效，文件内容保持不变：
```fish
alonabot run --data-dir ./data --listen 127.0.0.1:8080
```

用 `alonabot logs --level critical` 只看严重故障，或用 `alonabot run --log-level critical` 为本次启动提高全局输出阈值。运行中用 `alonabot log-level global critical` 保存全局阈值，用 `alonabot log-level <组件 ID> critical` 单独设置来源。查询筛选与输出设置见 [管理 API](#管理-api)。

Core 在 `listen` 地址（缺省 `127.0.0.1:7650`）上提供唯一的公开 HTTP 入口。请求 Host 的主机名须匹配 `base_url`：`/plugins/{id}/...` 转发给该插件当前 Session 的路由，其余路径交给管理处理器；其他 Host 返回 421。缺省设置下，lifecycle 示例的私有页面地址为 `http://127.0.0.1:7650/plugins/demo/page`。listen 为非 loopback 地址时，`base_url` 必须是 https，或者配置非空的 `trusted_proxies`（由可信反代终止 TLS），设置文件与 `--listen` 都按此校验。

插件路由的全部响应（含 Core 的错误响应）统一设置 `Content-Security-Policy: sandbox allow-scripts allow-forms allow-popups` 与 `X-Content-Type-Options: nosniff`，覆盖插件的同名头。Core 清除插件设置的全部 `Set-Cookie` 和 `Access-Control-Allow-Credentials`（含 trailer），Cookie 只以 WARN 记录名称；缺失 Content-Type 时补 `application/octet-stream`。HTML 页面按此沙箱策略返回。

Adapter 在 `adapter.Descriptor.HTTP` 中用 `[]adapter.HTTPRoute` 声明路由，对外位于统一端口的 `/plugins/<Adapter 包 ID>/...` 下声明路由，例如 OneBot 11 反向 WebSocket。路由规则与插件相同：私有路由需要 Core 会话，非 GET/HEAD 的 Cookie 请求还要验证 CSRF 与 Origin。`Auth` 的零值要求 Core 会话；公开路由写 `Auth: adapter.AuthPlatformToken`：Core 代理保留入站的 `Authorization` 交给 Adapter，由 Adapter 校验平台凭证，并给出错误时的平台响应。插件路由的 `Authorization` 一律被清除。两者都会清除入站的 Cookie、CSRF、Forwarded 与 X-Forwarded-*，并注入内部代理凭证；SDK 先恒定时间比较该凭证，比较后移除，再进入用户 Handler。WebSocket Upgrade 由代理双向转发，连接寿命不受普通请求期限限制，路由撤销或 Session 失效时由 Core 关闭。

设置了 `proxy` 时，Core 与子进程使用它作为 `HTTP_PROXY`/`HTTPS_PROXY`；为空时原样传递环境中的 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。

`go_download_base`（缺省 `https://go.dev/dl/`，可填 `https://golang.google.cn/dl/`）与 `goproxy`（缺省 `https://proxy.golang.org,direct`）只用于[源码构建](#源码构建)。`go_download_base` 须使用以 `/` 结尾的 HTTPS 地址；两项修改均从下一次下载生效，无需重启。`GET`/`PUT /api/v1/core/downloads` 读写这两项以及网络代理 `proxy`；修改 `proxy` 后要重启 Core。GitHub API 与归档地址见[组件下载镜像](#组件下载镜像)。

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

`OPTIONS_PAGE` 的默认值和范围见 [插件指南](https://alonabot.github.io/docs/tutorial/project)，可用 `alonabot tune list` 查看；聊天目录每页的上限取当前 Core 生效值的十倍，请求的 `limit` 仍缺省为 50。

### 容器内的监听与首次领取

端口只发布到宿主的 `127.0.0.1`。镜像设置 `ALONABOT_CONTAINER=1` 时，缺省监听为 `0.0.0.0:7650`，`base_url` 仍为 `http://127.0.0.1:7650`；显式 `listen` 优先。该容器例外实际生效时，启动日志输出一次 WARN 提醒宿主本机发布规则。远程管理可使用 `ssh -N -L 7650:127.0.0.1:7650 user@host`，再打开本机地址；HTTPS 反代部署按 [Core 设置与 HTTP 入口](#core-设置与-http-入口) 配置 `base_url` 和 `trusted_proxies`。

先保持宿主端口绑定 `127.0.0.1`，在本机完成首次令牌确认后，再改为绑定 `0.0.0.0` 或其他非回环地址。Docker 的 `userland-proxy` 可能让局域网请求也显示为网关来源；确认前，能访问该端口的人可能领取令牌并取得完整管理权限。

首次领取接口 `POST /api/auth/bootstrap` 接受本机直连。`ALONABOT_CONTAINER=1` 时，也接受经 Docker 网桥访问、来源为 `/proc/net/route` 中 IPv4 默认网关的请求。Host 与 Origin 仍须匹配 `base_url` 的本机地址；带 `Forwarded`、`X-Forwarded-*` 或 `X-Real-IP` 的请求会被拒绝。确认后接口停止返回令牌。

## 管理 API

使用前文的受控 `base_url` 与认证方式。Core API 主版本 1 位于 `/api/v1/`，认证位于 `/api/auth/`。JSON 写请求使用 `Content-Type: application/json`。CLI 发送 `Authorization: Bearer <管理 token>`；浏览器使用管理 Cookie，GET/HEAD 之外再带受控 `Origin` 与 `X-Alonabot-CSRF`。`POST /api/auth/login` 接收 `{"token":"…","remember":true}`，返回 `{"csrf_token":"…"}` 并设置 Cookie：`remember` 为真时会话 30 天且 Cookie 持久，省略或为假时会话 8 小时、Cookie 随浏览器关闭失效；刷新后用 `GET /api/auth/session` 恢复 CSRF，`POST /api/auth/logout` 撤销会话。登录和确认保留持久 token 文件。

管理错误格式为 `{"code":"FAILED_PRECONDITION","safe_message":"…"}`。HTTP 对应 400 INVALID_ARGUMENT、403 PERMISSION_DENIED、404 NOT_FOUND、409 ALREADY_EXISTS 或陈旧配置、412 FAILED_PRECONDITION、501 UNSUPPORTED、503 UNAVAILABLE、500 INTERNAL/其他错误码。认证错误为纯文本。未知管理路由先认证：无凭据为 401，认证后为 JSON 404。需确认的操作返回 409 与 `{"affected":["…"]}`：展示全部对象，管理员同意后重发原请求，install 写 `Confirm: true`，动作与实例创建写 `confirm: true`。解析前核对实际状态和 Content-Type。

| 操作 | 端点与请求 |
| --- | --- |
| Core 身份与组件 | `GET /api/v1/core`、`GET /api/v1/packages` |
| 安装 | `POST /api/v1/packages/install`，使用导出 `InstallRequest` 字段 `Archive`、`Source`、`ID`、`Confirm`、`Replace`、`Build`、`Ref`，见[管理类型](../internal/core/manage/install.go)；`Build` 返回 202 与任务 ID，见[源码构建](#源码构建) |
| 后台任务 | `GET /api/v1/tasks`、`GET`/`DELETE /api/v1/tasks/{id}` |
| 构建环境 | `GET`/`DELETE /api/v1/core/build-environment`，DELETE 可带 `?part=toolchain\|cache` |
| Go 下载与网络代理 | `GET`/`PUT /api/v1/core/downloads`，`{"go_download_base":"https://go.dev/dl/","goproxy":"https://proxy.golang.org,direct","proxy":""}`；PUT 只更新给出的字段，非法值返回 400 |
| 启用、停用、重启、重载 | `POST /api/v1/packages/{action}`，`{"id":"demo"}`；disable 可能要求 `confirm` |
| 更新 | `POST /api/v1/packages/update`，`{"id":"demo","confirm":false,"install":{"Source":"owner/repo"}}` |
| 回退 | `POST /api/v1/packages/rollback`，`{"id":"demo","generation":"retained-generation","confirm":false}` |
| 固定版本 | `POST /api/v1/packages/pin`，`{"id":"demo","constraint":"1.2.3"}`，空串解除 |
| 卸载 | `POST /api/v1/packages/uninstall`，`{"id":"demo","purge":0,"confirm":false}`，清理前展示实际对象 |
| Active 配置 | `GET /api/v1/plugins/{id}/config`，查看 `corrections` 与 `unknown` |
| 编辑 Pending | `PUT`/`PATCH /api/v1/plugins/{id}/config/pending`，带 `base_generation` 和完整 `config` 或 RFC 7396 `patch`；`DELETE` 丢弃 Pending |
| 导入与应用 | `POST /api/v1/plugins/{id}/config/import` 将编辑文件导入 Pending；`POST /api/v1/plugins/{id}/reload` 准备并提交 |
| Adapter 实例 | `GET /api/v1/adapter/instances`；`POST /api/v1/adapter/instances/{create,start,stop,delete}`；create 用 `package_id`、`instance_id`、`config`、可选 `confirm`，其余用 `instance_id`；列表的 `package_state` 为所属 Adapter 组件的 `enabled`、`disabled` 或 `not_installed`（未安装或已卸载，实例配置保留） |
| 实例配置 | `GET /api/v1/adapters/{adapter_id}/instance-schema`；`GET`/`PUT`/`DELETE /api/v1/adapters/{adapter_id}/instances/{instance_id}`；PUT 为 `{"config":{}}` |
| 日志 | `GET /api/v1/logs/recent`、`/stream`（SSE）、`/history?date=YYYY-MM-DD`，可筛选 `source`、`level`、`trace`；`GET`/`PUT /api/v1/logs/levels` 用 `{"global":"info","sources":{}}` 部分更新 |
| 配置文件 | `GET`/`POST /api/v1/profiles`；`PATCH`/`DELETE /api/v1/profiles/{id}`；指定、插件、平台、工具覆盖见[配置文件路由](../internal/core/manage/profile_api.go) |
| 工具策略 | `GET`/`PUT /api/v1/tools/policies`，写 `{"plugin_id":"demo","service":"echo","policy":"deny"}` 或原子批量 `{"items":[…]}`；allow/confirm/deny/inherit，inherit 删除覆盖 |
| 会话 | `GET /api/v1/plugins/{id}/conversations` 返回摘要；`DELETE /api/v1/plugins/{id}/conversations/{key}` 清除完整不透明键 |
| 插件数据 | `/api/v1/plugins/{id}/data/` 下的 tables/rows、kv、files、export、clear，见[数据处理器](../internal/core/manage/data_api.go) |
| 诊断包 | `POST /api/v1/diagnostics`，缺省 `{"since":"24h","logs":true,"settings":true,"components":true,"redact":true}`，返回 no-store ZIP |

### 配置文件的会话访问

在配置文件的「平台与消息」中设置会话访问。Core 按会话单独指定、实例缺省、`default` 的顺序选择配置文件，再应用该配置文件的模式：`all`（缺省）允许所有会话；`allow` 只允许列表中的会话，空列表拒绝全部；`deny` 只拒绝列表中的会话，空列表允许全部。被拒绝的会话不会触发命令或进入 Agent、插件事件处理，Core 也会拒绝发送。现有全局禁止名单继续在原有准入路径生效，任一规则拒绝即拒绝。

用 `GET`/`PUT /api/v1/profiles/{id}/session-access` 读取或整体替换模式和列表，例如 `{"mode":"allow","sessions":[{"instance":"main","session":"telegram:-100123#42"}]}`。条目以实例和完整会话 ID 配对；话题与聊天分别匹配。每次最多提交 1000 项，重复项只保留第一次。省略 `sessions` 或传空数组会清空列表；`all` 模式下列表仍保存，但不参与判断。

### 其他管理约定

本地 JSON 安装提供 Core 主机上的 `Archive` 路径，身份和来源取自归档内 `plugin.yaml`；在线安装只提供 `Source: "owner/repo"`，Core 从最新 Release API 选择平台归档。可选 `ID` 约束期望身份。首次来源信任和同类替换按需确认；后续更新与保留记录重装继续核对来源和身份。`Package.PublisherKeyID`、`Package.Signed` 保留历史记录含义；新安装及成功更新或重装后分别为空和 false。

浏览器对 install/update 端点发送 `multipart/form-data`：必填一个文件 part `archive`；可选文本 `confirm=true|false`；仅 install 可带 `replace=true|false`，仅 update 必填文本 `id`。文件 part 带 `filename`，文本 part 省略；未知或重复 part、缺文件或更新 ID、错误布尔值和截断请求返回 400 / INVALID_ARGUMENT。认证先于上传解析；归档流式写入独占临时目录，沿用安装大小与预检限制，调用完成后清理。确认响应为 409 与完整 `affected`，确认后重新上传归档并带 `confirm=true`。预检在修改旧组件前完成，来源绑定不符返回 403 / PERMISSION_DENIED；更新沿用身份、pin 与依赖检查并保留用户数据。

Pending 编辑要求当前 `base_generation`，陈旧编辑为 409。读取时秘密用 `{"$secret_set":true}` 遮盖；只在需要修改处提交新明文，勿提交掩码或 Core 秘密引用。reload 建立独立受管 SQLite 候选代次，restart 保持 Active；提交后的 OnReady 失败保留新 Active 并进入 Degraded。`GET /api/v1/plugins/{id}/generations` 列出保留代次。开发组件的 update/rollback/uninstall 为 412。

会话键为 `<会话>@<实例>`，成员独立会话再加 `@<用户>`。将完整键百分号编码为一个路径段，服务端只解码一次，保留斜杠、冒号、反斜杠与点段；键须为 1～256 字节 UTF-8。数据浏览可选择只读保留代次。`/data/clear` 必须带 `{"confirm":true}`，删除可写表行（含 KV）和插件文件，保留表结构、迁移与目录。KV 遵循[存储规则](https://alonabot.github.io/docs/tutorial/storage#键值存储)，键编码为不透明路径段。文件路径保持相对路径，拒绝穿越和符号链接；清理或编辑文件独立于代次回退。

休眠使用 `GET`/`PUT /api/v1/core/sleep`、`GET`/`PUT`/`DELETE /api/v1/plugins/{id}/sleep-policy` 及 `GET /api/v1/plugins/{id}/runtime`。逐插件 PUT 为 `{"policy":"never"}` 或正时长，DELETE 恢复自动策略；读取保持休眠。缺省值、预算和失败行为见[插件休眠](#插件休眠)。

### 更新日志与 WebUI 后台

认证后 `GET /api/v1/core/changelog` 返回 `{"locale":"en-US","entries":[{"version":"0.1.0","markdown":"…"}]}`。来源为正式 `changelog/v<SemVer>/{en,zh-CN}.md`，排除 unreleased，en 映射 en-US；按 `Accept-Language` 精确或主语言匹配，缺省或未匹配回退英文。条目按 SemVer 倒序，版本无 `v`，未知日期省略 `date`，保留 Markdown。`GET /api/v1/packages/{id}/changelog` 读取安装包：单文件 CHANGELOG.md 排除 Unreleased，locale 为 en-US；目录形式逐版本按偏好语言精确、规范标签、主语言匹配，最终回退 en-US.md，同级取文件名字典序最小者。未安装或 dev 组件为 404；无正式条目返回 en-US 与空 entries。

认证后 `GET /api/v1/packages/{id}/icon` 提供图标，`GET /api/v1/packages/{id}/examples/{path}` 只提供 Active 包引用的配置图片。包列表的 `Name`、`Description`、`Author` 取自包内 `plugin.yaml`，读不到时省略。组件页面在包的 `pages` 中列出，包含 `path`、`title`，账号配置页面另带 `purpose: "instance_config"`。WebUI 经 `/plugins/{id}` 在沙箱打开页面，账号配置页面在添加或编辑账号的对话框里打开。见[页面套件](https://alonabot.github.io/docs/advanced/http#插件页面)与[账号配置页面](https://alonabot.github.io/docs/develop/adapter#账号配置页面)。

官方 WebUI 组件包含真实后台进程。认证前缀 `/plugins/alonabot-webui-admin` 下有 `/stats?range=24h|7d|30d`、`/users`、`/store`、`/changelog/seen?component=core`。`PUT /changelog/seen` 仅接收 `{"component":"core","version":"0.1.0"}`，按 SemVer 单调持久化；GET 返回 `{"version":null}` 或版本。仅复制静态 `dist` 时这些后台路由不可用。模型、人格、MCP、技能、记忆和任务页面由 Agent 提供，见其 [HTTP 实现](https://github.com/AlonaBot/alonabot-agent/blob/main/internal/agent/pages.go)。

完整现行路由和请求/响应类型由[管理包](../internal/core/manage)维护，[管理集成测试](../tests/e2e/manage_e2e_test.go)通过真实 Core 验证。

### 组件商店

官方 WebUI 将自己的数据保存在其受管 SQLite 中。下列后台路由均相对于 `/plugins/alonabot-webui-admin`，沿用管理认证与 CSRF，响应带 `Cache-Control: no-store`。组件安装与更新仍调用 Core 管理 API。

使用 `GET /store?source=AlonaBot/plugin-store&mirror=auto&role=plugin&q=&sort=stars&offset=0&limit=9&refresh=false`。商店读取源仓库 `index` 分支上的 `store.json`，这是该仓库 CI 生成的快照（见 [plugin-store README](https://github.com/AlonaBot/plugin-store#readme)），不再逐条调用 GitHub API 或下载归档。`mirror` 选择读取线路：`auto`（缺省）同时请求 `https://raw.githubusercontent.com/<source>/index/store.json` 与 `https://cdn.jsdelivr.net/gh/<source>@index/store.json`，取先成功者；`github` 或 `jsdelivr` 只用对应地址；https 地址前缀拼在 raw 地址之前（GitHub 加速服务的用法）；其他值返回 400 / INVALID_ARGUMENT。仓库或角色不合法的条目被丢弃，非 `data:image/` 的图标被去掉。`source` 必须为 `owner/repo`，显式指定非负 `offset` 与正整数 `limit`；role 为空或 plugin/adapter/agent/webui。响应包含 `source`、`fetched_at`、`stale`、`items`、`total` 与 `counts`（all 及四种角色）。items 包含 `repo,role,tags,id,name,version`，可选 `description,author,stars,icon_url`；stars 缺省表示未知，0 保持为 0。默认源为 `AlonaBot/plugin-store`；源列表在浏览器持久化，主动清空后保持空列表。

搜索对 name、description、repo、author 与空格连接的 tags 做不区分大小写的子串匹配，缺失值按空串处理。counts 在搜索后、role 筛选前计算，total 在角色筛选后计算。默认或空 `sort=stars` 将已知星标数降序排列，未知值随后；`sort=name` 按名称不区分大小写升序，同值保留索引顺序，其他排序值返回 400 / INVALID_ARGUMENT。WebUI 每页显示九张卡片，搜索、角色或排序变化后回到第一页。

`refresh` 仅接受字面量 `true` 或 `false`，空值也返回 400 / INVALID_ARGUMENT。省略或 true 会刷新并持久化完整成功快照；false 且有成功快照时，在快照上搜索、计数、排序与分页，全程无远端请求；无快照时执行首次抓取。快照读取返回 `stale=false` 和原 `fetched_at`。进入商店、切换源及支持键盘的手动刷新使用 true，搜索、筛选、排序与分页使用 false。刷新失败保留成功快照和原时间，返回 `stale=true` 及完整错误；无快照的一般失败返回 500 / INTERNAL，源仓库没有 `store.json` 时返回 404 / SOURCE_INDEX_MISSING。

安装和更新使用的版本检查与 Release 预览通过 Core 的 `GET /api/v1/core/github/release?source=owner/repo` 查询，返回 `repo,version,release_url,markdown,fetched_at`。它与安装共用 GitHub 镜像设置、HTTP 客户端及原始仓库和标签校验，查询失败完整脱敏返回，不使用旧快照。latest Release 的 404 返回 404 / NOT_FOUND，由 WebUI 的 `getLatestRelease` 映射为 null。链接仅允许 HTTP/HTTPS/mailto，HTML 按文本显示。更新预览使用组件绑定来源，更新后说明来自已安装包内 changelog，已读版本只推进到实际展示的版本。

收录时，`GET /store/check?repo=owner/repo` 读取最新 Release API 及其首个平台归档，保留持久缓存原状；核对仓库、Release 与资产 URL、标签 SemVer、资产名称和长度，以及包内 manifest 的 ID、版本与来源。返回 `repo,role,id,name,version,release_url`，可选 description/author，及保持顺序的 manifest `warnings`（空数组也返回）。响应 repo 取已验证 manifest Source 去掉 `github.com/` 后的规范大小写。Core 安装预检继续检查平台、build-info、资源及归档完整性。仅 latest Release 的 404 映射为 NO_RELEASE；坏资产 404 及其他失败返回 500 / INTERNAL，保留实际错误。

源索引为 `{"version":1,"plugins":[{"repo":"owner/repo","role":"plugin","tags":["utility"]}]}`，tags 可省略。检查后，用户填写标签，用检查所得规范 repo，在浏览器打开源仓库 `issues/new`，携带 `template=submit-plugin.yml` 以及经 URL 编码的 `repo`、`role`、`tags`、`release_url`，由浏览器提交 GitHub Issue。

WebUI 商店每个远端 HTTP 请求的期限为 20 秒，覆盖响应体；串行请求的总耗时可更长，刷新响应调用方取消。`/store/check` 读取 GitHub API 与 Release 归档，初始 URL 与每次重定向均检查 HTTPS 主机和路径。安全上限为：商店快照 32 MiB（含内联图标）、元数据 1 MiB、归档 64 MiB、每轮展开 256 MiB（最多两轮）、单文件 128 MiB、每轮 2,048 条、保留的 manifest 256 KiB、保留的图标 1 MiB。超限保留 `exceeds ... bytes/entries` 错误并沿用同样的缓存规则。当前实现见 [WebUI store.go](https://github.com/AlonaBot/alonabot-webui/blob/main/internal/admin/store.go)。

## 组件安装

`run` 与库模式（`alonabot.Start`）共用 `internal/app` 的同一套 Core 装配；CLI 直接调用它，根包提供库模式的薄包装。组件经管理服务安装与启停，CLI 通过 `base_url` 上的管理 API 操作，从所选 `--data-dir` 自动读取持久凭据并直连本机 listen，Host 沿用 base_url；本机连接地址可用 `--admin-url` 覆盖：

`install` 的位置参数为已存在的普通文件时使用该归档，其他值按组件 ID 查找可执行文件旁的 `packages/<组件 ID>/`；缺失时列出可用的官方组件 ID。不带位置参数时，当前目录为作者工程便搜索 `bin/`，其他情况搜索当前目录；也检查直接包子目录。按归档名选择本机平台资产，静态 WebUI 使用 `any/any`，恰好一个候选才安装；零个或多个时列出搜索目录和候选后退出。发现阶段只读，安装经管理 API 完成。

`--archive` 可显式指定归档。位置参数、`--archive` 和 `--source` 互斥；来源确认、归档预检及 `--yes`、`--replace` 沿用管理服务。浏览器上传安装或更新同样只需归档文件。历史 Release 清理旁车后，旧 Core 的在线安装与更新需要先升级至采用此流程的版本，发布顺序见[发行维护](releasing.md)。

管理端点与配置读写详见 [管理 API](#管理-api)。

### 源码构建

安装请求带 `Build: true` 时，Core 不下载 Release 资产，而是从源码构建，例如 `POST /api/v1/packages/install` 的 `{"Source":"AlonaBot/alonabot-adapter-telegram","Build":true,"Ref":"v0.3.0","Confirm":true}`：

1. Core 经 `github_api_base` 将 `Ref`（分支、标签或提交；空串为默认分支）解析为提交，下载 `https://github.com/<owner>/<repo>/archive/<提交>.tar.gz`，解压到 `<data_dir>/staging`。配置了 `github_download_prefix` 就经该前缀下载。GitHub 未公开源码 tarball 的摘要；使用前缀时，构建任务日志会警告源码经前缀获取、内容未经摘要校验。请使用可信的下载镜像。
2. 按需下载与 Core 构建版本相同的 Go 发行版到 `<data_dir>/toolchain/go`：从 `go_download_base` 的 JSON 列表（`?mode=json&include=all`，`go.dev/dl/` 与 `golang.google.cn/dl/` 都提供）读取大小与 SHA-256 并核对。
3. 按 `alonabot pack` 的方式为 Core 所在平台打包：`CGO_ENABLED=0`、`GOTOOLCHAIN=auto`，模块经 `goproxy` 下载并按 GOSUMDB（缺省 `sum.golang.org`）校验。`GOCACHE`、`GOMODCACHE`、`GOTMPDIR`、`TMPDIR` 与 `HOME` 都在 `<data_dir>/cache` 下，因此 scratch 镜像也能构建。Go 子进程只从 Core 继承 `PATH`、`HTTP_PROXY`、`HTTPS_PROXY` 和 `NO_PROXY`，Go 设置由构建显式指定（`GOENV=off`、`GOWORK=off`）。
4. 归档照常经过安装预检、来源信任与同类替换确认，来源绑定为 `github.com/<owner>/<repo>`。`Ref` 是 SemVer 标签时版本取标签，并核对 `plugin.yaml` 的版本；其他提交的版本是 `plugin.yaml` 的版本加 `+<12 位提交>`，例如 `0.3.0+4769d92b0a8f`。

更新使用 `POST /api/v1/packages/update` 的 `{"id":"…","confirm":true,"install":{"Build":true,"Ref":"…"}}`，省略 `Source` 时构建已绑定的来源。

构建作为后台任务运行，请求立即返回 202 与 `{"task_id":"…"}`。`GET /api/v1/tasks/{id}` 返回 `{"id","action":"install|update","status":"running|succeeded|failed","created_at","finished_at","log","result"}`；`log` 包含各步骤与 go 命令的完整标准错误，成功时 `result` 与同步端点的结果相同。失败时 `http_status` 与 `error` 与同步端点的错误响应一致：需要确认时为 409 与 `{"affected":[…]}`，确认后带 `Confirm: true` 重新提交，构建缓存使重建更快。`GET /api/v1/tasks` 列出任务（不含 `log`），`DELETE /api/v1/tasks/{id}` 取消运行中的任务或删除已结束的记录。同一时间只运行一个构建，其余任务在日志中提示等待；任务只保存在内存中，Core 重启后消失，停机时取消运行中的构建。

`GET /api/v1/core/build-environment` 返回 `{"go_version":"go1.27.1","toolchain_bytes":0,"cache_bytes":0}`，未下载工具链时 `go_version` 为空串。`DELETE /api/v1/core/build-environment?part=toolchain|cache` 删除工具链或缓存，省略 `part` 时两者都删除并返回新的占用；构建运行中返回 412。工具链约占 250 MiB，缓存随构建增长。

只依赖 AlonaBot SDK 的普通插件构建很轻（实测几百 MiB、十几秒）；依赖较大的组件（例如带 LLM SDK 的 Agent）建议预留约 **2 GiB 可用内存**，扣除 Core 和其他负载的占用。Core 串行编译；Linux 上可用内存偏低时，构建日志会提示，任务继续运行。小内存机器优先从 Release 安装，或在另一台机器运行 `alonabot pack --target <os>/<arch>`，再上传适配 Core 平台的归档。

任务日志提示依赖下载可能受阻时，在系统设置中将 `goproxy` 改为可达的镜像，例如 `https://goproxy.cn,direct`，或设置网络代理 `proxy` 后重启 Core。`goproxy` 修改从下一次构建生效，详见 [Go 下载与代理设置](#core-设置与-http-入口)。静默提示只追加日志，构建继续运行。

用 `alonabot tune list` 查看启动调优项：`BUILD_MEMORY_SOFT_LIMIT`（每个 Go 进程 768 MiB）、`BUILD_GOGC`（50）、`BUILD_MEMORY_WARN_BELOW`（1536 MiB，仅 Linux）和 `BUILD_STALL_INTERVAL`（60s）。这些参数仅作用于 Core 的受管工具链。

源码构建不支持 cgo，也不支持私有仓库与私有模块：Core 不向 GitHub 或模块代理发送凭据，也没有 `GOPRIVATE`/`GONOSUMDB` 设置。镜像不含 git，`goproxy` 中的 `direct` 回退在容器内不可用，依赖须能经代理取得。module 须位于仓库根目录，源码不能含符号链接。Windows 主机与以开发版 Go 构建的 Core 尚不支持源码构建。

### 导出诊断包
```fish
alonabot diagnose --data-dir ./data --since 24h -o diagnose.zip
```

默认收集近 24 小时的日志、当前设置和 Active 配置、组件与环境信息，并开启脱敏。`--no-logs`、`--no-settings`、`--no-components` 可分别省略一组，至少保留一组。`-o` 也可写作 `--output`；省略时在当前目录生成 `alonabot-diagnostics-<UTC 时间>.zip`。

在线使用管理 API 和 Bearer token，自动读取 `<data_dir>/admin_token`。默认直连本机 listen 并保留 base_url Host；`--admin-url` 覆盖本机连接地址。连接被拒绝后尝试停机导出；原目录的 core.lock 须未被占用。离线导出只读原目录。

缺少可用的 Adapter 实例 schema 时，默认遮盖该组实例配置的全部叶值，保留结构和完整的 `${环境变量}` 表达式。`--no-redact` 会导出凭证明文；日志可能包含聊天内容，提交 issue 前请先检查归档。

## 组件下载镜像

Core 的 `core_config.json` 可独立保存 GitHub API 基址与归档下载前缀，WebUI 的「设置 → 组件下载」读写同一份设置。保存后对下一次安装或更新生效。

```json
{
  "github_api_base": "https://api-mirror.example",
  "github_download_prefix": "https://proxy.example/"
}
```

两个值缺省均为空：API 使用 `https://api.github.com`，归档使用原始资产 URL。下载地址严格按「前缀 + 原始 URL」拼接，例如 `https://proxy.example/https://github.com/owner/repo/releases/download/v1.0.0/package.tar.gz`；填写前缀时保留所需的末尾斜线。两个地址及重定向均使用 HTTPS。API 基址可含路径，须省略查询参数；两个地址均省略用户凭据与 fragment。下载前缀允许查询参数形式，例如 `https://proxy.example/?url=`。

从 Release 安装和更新时，Core 使用 Release API 返回的 SHA-256 摘要核验归档。`github_api_base` 决定摘要的信任根，只应填可信的 GitHub API 反向代理。旧资产缺少摘要时，清空下载前缀再试；直连原资产会校验大小并记一条警告。源码 tarball 见[源码构建](#源码构建)。

CLI 可逐字段覆盖本次操作，覆盖值保留在当前请求中：

```fish
alonabot install --source AlonaBot/alonabot-agent --github-api-base https://api-mirror.example --github-download-prefix https://proxy.example/ --yes
alonabot update alonabot-agent-official --github-api-base https://api-mirror.example --github-download-prefix https://proxy.example/ --yes
# 本次直连归档，API 继续沿用持久设置
alonabot update alonabot-agent-official --github-download-prefix '' --yes
```

管理 API 使用认证后的 `GET/PUT /api/v1/core/github` 读写两个持久字段，PUT 须同时提供两个字符串。安装请求的 `github_api_base`、`github_download_prefix` 可选，更新请求放在 `install` 对象内；省略沿用设置，显式空串恢复对应官方地址。Core 先核对原始 GitHub Release 与资产的仓库、标签和地址，再转换访问 URL；下载后核对归档摘要、包身份、平台、版本与 build-info。错误及诊断设置省略下载前缀的查询串，跨主机重定向移除 `Referer`。下载无自动重试或地址回退。

`INSTALL_DOWNLOAD_TIMEOUT` 默认每个 GitHub 请求 5 分钟，覆盖重定向和响应体读取；`INSTALL_MAX_RELEASE_JSON` 默认限制 GitHub API 响应为 4 MiB，也适用于源码引用解析；`INSTALL_MAX_ARCHIVE` 同时限制 Release 资产与源码 tarball；`INSTALL_MAX_HTTP_ERROR` 默认保留非 200 响应体的前 4 KiB。用 `alonabot tune list` 查看生效值。

商店的索引发现与收录检查仍使用所选源的 GitHub 索引及展示快照，其独立抓取规则见上文；首次浏览商店需要索引源可达。按仓库直接安装、已安装组件的更新检查及确认预览直接调用 Core，随后安装或更新也使用同一份镜像设置。

## 插件休眠

plugin/agent 的 `sleep_capable` 缺省为 true，显式 false 要求常驻。旧版构建器曾把源码中的缺省写成 false；用修复后的构建器重新构建这些包，显式 false 继续保留原义。全局 `always_on`、开发组件、Adapter、WebUI、活动引用与现有路由判断继续生效。HTTP、Hook、Raw、Bus 订阅及持续事件路由在自动策略下返回 `route_requires_resident`；显式逐组件正时长覆盖该路由要求并返回逐路由警告，manifest 的常驻声明继续生效。

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

无参数返回全局设置与全部已安装 plugin/agent 的策略和运行状态；单 ID 返回对应对象。`auto` 删除覆盖，`never` 与正时长立即写回设置并生效。全局模式与时长通过[管理 API](#管理-api) 的 `PUT /api/v1/core/sleep` 修改，两项必填。管理读取保持休眠；需要路由而尚无有效已提交快照时，策略查询返回 412 / `FAILED_PRECONDITION`，全局设置与 runtime 可独立读取。保存后的恢复失败保留实际错误；重新 GET 核对已保存设置。

唤醒与调度的实现、测试入口见[架构地图](../ARCHITECTURE.md)；开发验证见[源码开发](development.zh-CN.md#验证改动)。

删除 Adapter 实例时，Core 同步清除它的事件游标、Profile 指定和会话访问名单条目；名单的访问模式及其他实例条目保留。同名重建使用全新游标。普通停用与重启保留这些状态。

## 日志摘要与详情

每条日志正文（包括 DEBUG）用自然语言说明发生了什么，直接包含组件或路由、状态变化与失败原因。结构化字段用于筛选和关联。正文已包含的错误在控制台只显示一次。

Core 控制台展示摘要。如需查看完整的结构化字段、Trace、源码位置和错误，可读取日志文件或查询管理 API。新写入的日志以 JSON 行保存完整管理 API Record（`time`、`level`、`source`、`message`、`attrs`、`trace_id`），控制台格式与文件格式独立。历史查询同时读取既有文本记录和新 JSON 记录，保留来源、级别及 Trace 筛选。`--log-format=json` 继续提供完整结构化 stderr。INFO 记录接入、处理、发送及休眠唤醒摘要，内部事务与中间状态为 DEBUG；正文、模型输入输出、工具参数和凭据沿用现有保护规则。运行期 `runtime/reply.sqlite` 只保存本次 Session 的回复修饰关联，启动时清理旧运行数据，Session 结束时删除对应行。
