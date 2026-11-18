# Admin API 与插件页面桥

WebUI 登录后通过同一个 Core 公开入口管理组件。本文登记当前 HTTP 接口和后继 WebUI 应遵循的插件页面桥格式；新增管理接口时同步追加本文件。行为依据为 [CONTRACTS §7.2～§7.5、§9.3](CONTRACTS-v0.1.md)，实现入口为 `internal/core/adminauth`、`internal/core/manage` 与根包 `alonabot.Start`。

## 1. 入口、版本与认证

当前主版本为 **1**，常量为 `manage.AdminAPIMajor uint32 = 1`。管理路径使用 `/api/v1/`，认证路径为 `/api/auth/`。同一主版本内只增加兼容能力。WebUI manifest 的 `admin_api_min_major <= 1 <= admin_api_max_major` 须成立；Install（含 replace）、Enable 与 Update 预检返回 HTTP 412 / `FAILED_PRECONDITION` 拒绝范围失配，提示包含 Core 主版本及组件声明范围。检查发生在停止、卸载或修改旧组件前。

Core 使用 `base_url` 的受控主机名与单一 Listener，错误 Host 返回 HTTP 421。远程部署使用 HTTPS，可由已配置可信代理终止 TLS；loopback 开发允许 HTTP。WebUI 文件允许登录前读取，管理接口与图标使用以下认证：

| 调用方式 | 请求凭证 | 变更请求的附加要求 |
| --- | --- | --- |
| 浏览器 | 登录签发的管理 Cookie | GET/HEAD 以外的方法携带受控 `Origin` 和 `X-Alonabot-CSRF: <csrf_token>` |
| CLI | `Authorization: Bearer <管理 token>` | 使用显式配置的 token |

Cookie 为 host-only、`Path=/`、`HttpOnly`、`SameSite=Strict`；HTTP 下名称 `alonabot_admin`，HTTPS 下为 `__Host-alonabot_admin` 并带 `Secure`。会话固定 **8h** 到期，同时检查认证代次。登出撤销当前 Cookie 会话；离线 reset 递增认证代次并撤销旧管理会话。管理 token 与聊天事件中的 Core Admin 用户列表分别管理。

缺少有效凭证，或 Cookie 变更请求的 CSRF/Origin 校验失败，返回 HTTP 401、英文纯文本 `unauthorized`。管理凭证尚未初始化返回 HTTP 503、`admin token is not initialized`；认证存储故障返回 HTTP 500、`internal server error`，详细错误记入 Core 日志。

## 2. 登录与登出

| Method / path | 认证与请求 | 成功响应 | 失败 |
| --- | --- | --- | --- |
| `POST /api/auth/login` | 受控 `Origin`；`Content-Type: application/json`；`{"token":"<管理 token>"}` | 200、`application/json`、`{"csrf_token":"..."}`，同时设置管理 Cookie；首次成功登录删除数据目录中的 `admin_token` 文件 | Origin 无效为 403 `forbidden`；Content-Type 无效为 415 `expected application/json`；token、JSON 或多份 JSON 无效为 401 `unauthorized`；管理凭证未初始化为 503；存储或首次 token 文件删除失败为 500 |
| `POST /api/auth/logout` | 管理认证；Cookie 调用须带 Origin/CSRF | 204、空正文；撤销当前 Cookie 会话并发送删除 Cookie | 401、503 或 500 的认证错误；撤销会话失败为 500 |

以上错误使用 `text/plain; charset=utf-8`，正文末尾有换行。登录请求自身也检查 Origin。

## 3. 组件管理

下表各端点使用第 1 节的管理认证；JSON 成功响应的 Content-Type 为 `application/json`，直接返回结果，无外层结果包装。写接口拒绝未知字段和多份 JSON，解析错误返回 400 / `INVALID_ARGUMENT`。字段大小写以下表为准。

### 3.1 安装与列表

`POST /api/v1/packages/install` 接收 `InstallRequest`，字段沿用 Go 导出名称：

```json
{
  "Archive": "/path/on/core/package.tar.gz",
  "ReleaseJSON": "<release.json 原文>",
  "Signature": "<release.json.sig 内容，可省略>",
  "PublicKey": "<release.pub 标准 base64 内容，可省略>",
  "Source": "",
  "Replace": false,
  "Confirm": false
}
```

本地安装同时提供 `Archive` 与 `ReleaseJSON`，文件路径属于 Core 所在主机。远程安装只填写 `Source: "owner/repo"`，本地文件字段留空，Core 从该 GitHub Release 读取清单及对应资产。来源与验证过的发布者公钥绑定会持久化。无签名安装仍执行长度、归档摘要、平台与 manifest 校验；首次信任、未签名来源、发布者变化与同类替换按需确认。

成功为 200，返回一个 `Package`。`GET /api/v1/packages` 返回 200 和 `Package[]`，空列表为 `[]`，按安装顺序排列，含保留的已卸载记录。`Package` 字段如下：

```json
{
  "ID": "demo",
  "Role": "webui",
  "AdapterID": "",
  "Version": "1.0.0",
  "Source": "github.com/example/demo",
  "PublisherKeyID": "",
  "Pin": "",
  "Enabled": true,
  "Uninstalled": false,
  "Signed": false,
  "Active": "",
  "State": "Activated",
  "pages": []
}
```

`Role` 取 `plugin/adapter/agent/webui`；`Signed` 根据绑定发布者公钥是否非空返回；`Active` 是当前代次标识，尚无代次时为空。`State` 来自实际运行状态或安装状态，包含 `Installed`、`Stopped`、`Activated`、`Degraded`。WebUI 无进程，启用时显示 `Activated`。

`pages` 为运行路由中声明的页面列表（§20）。普通 `Package` 响应使用英文源标题；`GET /api/v1/packages` 按请求语言选择标题。

Install 的 `Package` 额外携带可选 `Warnings: string[]`，按原顺序返回 manifest 的全部警告；没有警告时省略。List 与其他返回 `Package` 的接口省略 `Warnings`。缺少 `migration_mode` 或 `prepare_effects` 各产生一条警告，安装与更新继续执行；管理服务按条记录 WARN 日志，字段为 `plugin_id` 与 `warning`，准备与回退按没有框架外副作用处理。

同类组件遵循 Agent/WebUI 全局各一个、Adapter 每个 `adapter_id` 一个的安装限制。已存在时为 409 / `ALREADY_EXISTS`；`Replace: true` 经预检和确认后按缺省卸载语义替换，保留旧绑定与数据，旧组件已启用时启用新组件。WebUI 可执行全部管理动作，安装界面须展示来源与这一权限。

### 3.2 组件动作

全部为 POST，成功返回 200；update 返回 `UpdateResult`，其他动作返回 `{}`。动作共享解码结构 `id`、`confirm`、`purge`、`constraint`、`generation`、`install`，每个动作消费下表字段；`id` 为组件 ID。

| Path | 请求字段及示例 | 行为与特定失败 |
| --- | --- | --- |
| `/api/v1/packages/enable` | `{"id":"demo"}` | 校验依赖并启动；WebUI 只修改启用记录，检查 Admin API 范围；前置条件失败 412 |
| `/api/v1/packages/disable` | `{"id":"demo","confirm":false}` | 停止并保留安装/数据；依赖影响需确认 |
| `/api/v1/packages/restart` | `{"id":"demo"}` | 以当前 Active 换 Session；要求已启用且有 Active；WebUI 无进程，返回 412 |
| `/api/v1/packages/uninstall` | `{"id":"demo","purge":0,"confirm":false}` | 停止并清理选定范围，准确对象需确认；开发组件返回 412 |
| `/api/v1/packages/pin` | `{"id":"demo","constraint":"1.2.3"}` | 精确 SemVer；空串解除；格式错误 400 |
| `/api/v1/packages/reload` | `{"id":"demo"}` | 要求已激活的进程组件，优先激活有效 Pending；无 Pending 时导入 config.json 再激活；WebUI 返回 412 |
| `/api/v1/packages/update` | `{"id":"demo","confirm":false,"install":{"Source":"owner/repo"}}` | `install` 使用 3.1 的 `InstallRequest` 大写字段；外层 `confirm` 覆盖 `install.Confirm`；校验身份、来源、公钥、pin、版本与依赖影响，通过候选准备与提交更新版本，保持启用状态 |
| `/api/v1/packages/rollback` | `{"id":"demo","generation":"<保留代次>","confirm":false}` | 要求目标代次属于该组件且不是当前 Active，校验确认后交给候选编排器回退到该保留代次（T-09.7），成功返回 `{}` |

`purge` 为整数：0（缺省）保留数据、配置及来源绑定；1 额外清除受管数据；2 额外清除受管配置及安装绑定。保留记录重装要求匹配原来源与身份。Adapter 替换保留的实例归旧组件，新实例配置块换归属时另行确认。具体清理对象见 [CONTRACTS §9.3](CONTRACTS-v0.1.md#93-cli-的最小完整动作)。

Update 成功响应为 `UpdateResult`：有警告时为 `{"warnings":["..."]}`，按原顺序返回 manifest 的全部警告；没有警告时为 `{}`。JSON 与 multipart 使用相同响应格式。

Update 的本地候选可以提供 `install.Archive/ReleaseJSON/Signature/PublicKey`；省略候选来源和归档时，沿用安装时绑定来源。前置校验可能先返回 400、403、404、409 或 412；重复上传已安装的同版本同归档会因不可变包目录已存在返回 409 / `ALREADY_EXISTS`。Rollback 经候选编排执行（T-09.7）：目标须为该组件的保留代次且不是当前 Active，否则返回 404 或 412；未确认时返回 409 并列出回退目标与丢弃的本地写入。

### 3.3 组件图标

`GET /api/v1/packages/{id}/icon` 经管理认证，只读取该已安装组件 `manifest.Icon` 声明的普通文件。`HEAD` 状态及响应头与 GET 相同，正文为空。查询参数忽略；额外路径返回 404，越界/非规范路径在路由匹配前拒绝。未安装、已卸载、无图标、缺失文件、目录或链接返回 404；开发组件图标本切片返回 404。

| 扩展名 | Content-Type |
| --- | --- |
| `.svg` | `image/svg+xml` |
| `.png` | `image/png` |
| `.webp` | `image/webp` |

安装预检拒绝其他扩展名、缺失或非普通图标文件。图标文件响应强制以下头；WebUI 用 `<img>` 引用，直接打开 SVG 也受同一策略约束：

```http
Content-Security-Policy: sandbox; default-src 'none'; style-src 'unsafe-inline'
X-Content-Type-Options: nosniff
```

图标 404 使用英文纯文本；包读取或状态库实际故障返回 500、纯文本 `internal server error` 并记录详细错误。

`GET /api/v1/packages/{id}/examples/{path}` 经管理认证。`{path}` 是相对于包内 `examples/` 目录的路径，可含子目录。Core 按 Active 代次的已提交 Journal 选包，只读取包内 `config_schema.json` 的 `example` 引用的普通文件；Pending 或未提交候选不参与选包。`HEAD` 与安全响应头同图标；`.png`、`.jpg`/`.jpeg`、`.webp`、`.gif` 分别返回 `image/png`、`image/jpeg`、`image/webp`、`image/gif`。

未安装、已卸载、开发组件、无 Active 代次、缺少 `config_schema.json`，或路径不规范、未引用、文件缺失、目录或链接时返回 404。Active 缺少已提交包记录、schema 无效或存储读取故障时，按图标规则返回 500。

## 4. Core Admin 与 Adapter 实例

各端点经管理认证，JSON Content-Type 与错误规则同第 3 节。添加 Core Admin 时，将 help 插件 `/whoami` 回复的用户字符串 `<adapter_id>:<external_user_id>` 原样填入 `user`。同一 Adapter 的各实例共用该用户字符串；回复中的实例名供核对消息来源（CONTRACTS §15.2）。缺省所有值按原有校验处理。

| Method / path | 请求 | 成功响应 | 操作语义与特定错误 |
| --- | --- | --- | --- |
| `GET /api/v1/admins` | 无正文 | 200、排序的字符串数组，例如 `["telegram:123"]` | 空列表为 `[]` |
| `POST /api/v1/admins/add` | `{"user":"telegram:123"}` | 200、`{}` | 写回 `core_config.json` 并立即生效；重复添加保持集合；格式错 400 |
| `POST /api/v1/admins/remove` | `{"user":"telegram:123"}` | 200、`{}` | 删除用户并立即生效；格式错 400 |
| `POST /api/v1/adapter/instances/create` | `{"package_id":"adapter_package","instance_id":"bot","config":{},"confirm":false}` | 200、`{}` | 补全并校验实例配置，原样保留环境变量引用，写回设置；实例名跨包唯一，冲突 409；换配置块归属需确认；未知组件 404；组件已卸载或角色非 Adapter 为 412 |
| `POST /api/v1/adapter/instances/start` | `{"instance_id":"bot"}` | 200、`{}` | 由当前配置启动；要求 Adapter 激活，配置有效；已运行 409；未知实例 404 |
| `POST /api/v1/adapter/instances/stop` | `{"instance_id":"bot"}` | 200、`{}` | 保留配置，本次 Core 运行中持续停用，包括组件崩溃恢复；下次 Core 启动按设置启动；未运行 412 |
| `POST /api/v1/adapter/instances/delete` | `{"instance_id":"bot"}` | 200、`{}` | 停止仍运行的实例，删除配置与指向它的缺省选择；配置提交后清除该实例的配置文件指定及运行选择（§23）；未知实例 404 |

本节的实例动作响应仅包含操作结果；脱敏配置读取与编辑见第 14 节。用户、消息与事件语义沿用 [MESSAGE-MODEL §1](MESSAGE-MODEL.md#1-eventactor-与-target)，平台身份须来自可信 Adapter。

## 5. 错误、确认和路由规则

管理错误通常返回下列 JSON，错误码沿用协议名称并去掉 `ERROR_CODE_` 前缀：

```json
{"code":"FAILED_PRECONDITION","safe_message":"Core Admin API major 1 is outside WebUI supported range [2, 2]"}
```

| HTTP 状态 | `code` / 格式 |
| --- | --- |
| 400 | `INVALID_ARGUMENT` |
| 403 | `PERMISSION_DENIED` |
| 404 | `NOT_FOUND`，含未知管理端点 |
| 409 | `ALREADY_EXISTS`、配置编辑的陈旧代次 `FAILED_PRECONDITION`，或下文确认响应 |
| 412 | `FAILED_PRECONDITION` |
| 501 | `UNSUPPORTED` |
| 503 | `UNAVAILABLE` |
| 500 | `INTERNAL` 及当前映射中其余错误码；保留实际 `code` |

需要显式确认时为 HTTP 409，正文为 `{"affected":["受影响对象", "..."]}`。客户端展示全部对象，在管理员同意后重发同一请求并置 `Confirm: true`（install）或 `confirm: true`（动作、实例创建）。它与 `ALREADY_EXISTS` 的 JSON 结构不同。

认证在管理路由查找前执行：未知 `/api/...` 无有效凭证返回 401；有效认证后返回 404、`application/json`、`{"code":"NOT_FOUND","safe_message":"management endpoint not found"}`。当前动作路径上的未登记方法同样落入这一管理 404。GET 列表与图标路由隐式接受 HEAD；POST 写接口按表格方法调用。

以下使用纯文本或标准路由响应，客户端按 HTTP 状态与 Content-Type 解析：

- 认证错误与登录错误见第 1/2 节，均为 `text/plain; charset=utf-8`。
- 图标缺失为 404；受控 Host 不匹配为 421；静态错误见第 6 节，均为纯文本。
- ServeMux 对子树入口缺少尾斜杠的规范路径（如 `/api/v1`）会产生标准重定向；客户端使用本文完整端点。原始路径含 `..`、重复分隔符、反斜杠或绝对路径形式时按所属入口拒绝；第 15 节会话清除的最后一段按不透明键接收。
- `/plugins/{id}/...` 与 `/media/...` 在 WebUI 之前分流，使用各自现有状态码与 MIME；插件路由受现有 sandbox CSP、nosniff 和凭证过滤保护。

## 6. 静态 WebUI

每次 GET 选取当前未卸载且已启用的唯一 WebUI，读取其 `assets` 根目录；`/` 对应 `index.html`。文件存在时提供原文件；不存在且最后一段无扩展名时回退根 `index.html`；带扩展名的缺失路径返回 404。实际目录缺少尾斜杠时，以 301 重定向到加上 `/` 的路径并保留查询参数，便于目录首页解析相对资源。目录仅提供该目录的 `index.html`，缺失返回 404，无目录列表。实际 I/O 故障、损坏的安装或状态库故障返回安全的 500 并记录错误，只有文件不存在才适用回退。

已安装文件及实现 `io.ReadSeeker` 的开发 FS 文件由打开的句柄流式提供，支持 Range 请求；句柄在本次请求结束时关闭。开发 FS 文件未实现 `io.ReadSeeker` 时整文件读入内存后提供，并记录 WARN；大型资源应使用支持 Seek 的文件系统，例如 `os.DirFS`。

`/api`、`/plugins`、`/media`、`/admin`、`/healthz`、`/internal`、`/auth` 及其子路径由 Core 保留，在静态查找前分流或拒绝，包中同名文件也覆盖不了这些入口。文件路径须为有效的相对路径，逐级拒绝链接；静态 GET 以外的方法（含 HEAD）返回 405、`Allow: GET`。WebUI 文件响应按内容提供 MIME，未附加 WebUI 专用 CSP 或 X-Frame-Options。

没有已启用 WebUI 时为 503，英文提示包含 `use alonabot install to install a WebUI component`；Core 升级导致声明范围失配时同样为 503，并给出 Core 主版本与范围。管理、登录、插件和媒体入口继续可达。

库模式 `alonabot.WebUI(fsys fs.FS)` 以 FS 根作为本次运行的静态根，同一 `Start` 重复登记时报错。开发 WebUI 无 manifest，按审查者 D-140 决定跳过 Admin API 范围检查并在启动时记 WARN（待维护者确认）；已安装包的状态、来源和文件保持原值。

官方 WebUI 的 `/dev/forms` 仅由 Vite 开发模式登记。生产静态服务对该无扩展名路径仍按本节规则回退到 `index.html`，由前端展示页面不可用提示。

## 7. 插件页面 postMessage 桥

本节定义 WebUI 父桥的消息契约与 SDK 页面客户端的接口。插件页面通过父 WebUI 请求自己所属组件的 HTTP 路由，并查询语言、主题和本组件文案。`theme` 传实际生效的浅色或深色，`tokens` 使用本节定义的 `--ab-*` 名称；官方外壳的主题偏好和内部 CSS 变量见 CONTRACTS §7.5。插件 HTTP 路由见 [WIRE §3.3](WIRE-FORMAT.md#33-注册描述与-http-listener)，Listener 凭证由 SDK/Core 经 Bootstrap 管理（WIRE §3.2），浏览器桥无需取得它。

### 7.1 消息格式

以下使用 TypeScript 对象类型表达 JSON 字段，`?` 为可省略；`id` 是请求方产生并随响应原样返回的关联字符串。所有消息的 `channel` 固定为 `alonabot.plugin`。

```ts
// HTTP 请求
{channel:"alonabot.plugin", id:string, op:"request", method:string, path:string,
 headers?:Record<string,string>, body_base64?:string}

// 语言、主题请求
{channel:"alonabot.plugin", id:string, op:"locale"|"theme"}

// 页面就绪（D-159），返回语言、明暗与设计令牌
{channel:"alonabot.plugin", id:string, op:"hello"}

// 快捷键转发（D-159），key 只接受 "mod+k"
{channel:"alonabot.plugin", id:string, op:"shortcut", key:"mod+k"}

// 文案请求
{channel:"alonabot.plugin", id:string, op:"translate", key:string, args?:Record<string,string>}

// HTTP 响应
{channel:"alonabot.plugin", id:string, ok:true,
 result:{status:number,headers:Record<string,string>,body_base64:string}}

// 语言响应
{channel:"alonabot.plugin", id:string, ok:true, result:{locale:string}}

// 主题响应
{channel:"alonabot.plugin", id:string, ok:true, result:{theme:string}}

// 文案响应
{channel:"alonabot.plugin", id:string, ok:true, result:{text:string}}

// hello 响应
{channel:"alonabot.plugin", id:string, ok:true,
 result:{locale:string, theme:"light"|"dark", tokens:Record<string,string>}}

// shortcut 响应
{channel:"alonabot.plugin", id:string, ok:true, result:{}}

// WebUI 推送的事件（没有 id）
{channel:"alonabot.plugin", event:"theme", data:{theme:"light"|"dark", tokens:Record<string,string>}}
{channel:"alonabot.plugin", event:"locale", data:{locale:string}}

// 桥错误
{channel:"alonabot.plugin", id:string, ok:false, error:{code:string,safe_message:string}}
```

`body_base64` 使用标准 base64；请求省略时正文为空，响应保留完整字节。HTTP 响应的 `status` 保留实际状态，桥成功交付 HTTP 错误响应时也使用 `ok:true`；桥校验或执行失败使用 `ok:false` 与稳定错误码及安全英文提示。

`hello` 由页面在加载后发送，此后明暗、令牌或语言变化时 WebUI 只向发过 hello 的页面推送事件。`tokens` 的键是 CSS 变量名，至少包含 `--ab-bg`、`--ab-surface`、`--ab-surface-2`、`--ab-line`、`--ab-text`、`--ab-text-muted`、`--ab-accent`、`--ab-accent-soft`、`--ab-success`、`--ab-warning`、`--ab-danger`、`--ab-critical`、`--ab-radius-sm`、`--ab-radius`、`--ab-radius-lg`、`--ab-font-sans`、`--ab-font-mono`，值为可直接用于 CSS 的字符串；同一主版本内只增加键（D-159）。`shortcut` 只接受 `mod+k`（macOS 为 ⌘K，其他系统为 Ctrl+K），WebUI 打开全局搜索；其他键返回 `ok:false`。

`translate` 仅查所属组件的 `messages.key`，以 `args` 格式化命名占位符。语言按当前界面语言、该组件已有主语言回落与 en-US 规则选择；未知键返回键名，缺参数保留占位符，遵循既有 i18n 行为。`locale/theme/translate` 分别只返回相应值。

### 7.2 窗口、路径与凭证

1. WebUI 用带 `sandbox` 的 iframe 嵌入组件页面，并将 `iframe.contentWindow` 与组件 ID 绑定。父页面接收时检查 channel、消息结构及 `event.source` 是否为该窗口。不透明源的 `event.origin` 为 `"null"`，同时核对窗口身份；相同 origin 字符串无法代替窗口匹配。响应只发回绑定的正确窗口，`id` 对应原请求。
2. `path` 只接受插件相对路径。父页面拒绝绝对 URL、协议相对 URL、`..`、编码后的越界/分隔符混淆与反斜杠，校验发生在路径规范化之前。随后将目标固定到该 iframe 所属组件的 `/plugins/{id}/` 前缀，保留已校验的插件相对路径及查询参数。其他组件与 Admin API 的路径拒绝。
3. 父页面将请求限定到所属组件已注册 method/path，Core 现有代理执行最终匹配。请求使用管理员会话，变更请求由父页面添加受控 Origin 与 CSRF。父页面拒绝插件自带的 `Cookie`、`Authorization`、`X-Alonabot-CSRF`、`X-Alonabot-Proxy-Auth` 等凭证/内部身份头，比较头名时忽略大小写。
4. 父页面拒绝重定向，向插件返回响应前滤去 `Set-Cookie` 等凭证头。管理会话、CSRF 与内部代理凭证均留在父页面和 Core 之间。Core 代理继续按角色过滤请求头：私有 Adapter 路由也清除 Core 认证凭证；公开 Adapter 平台 token 路由按其独立声明处理平台凭证。
5. 插件的 HTML 与全部插件路由响应继续带 `sandbox allow-scripts allow-forms allow-popups`（无 `allow-same-origin`）和 `nosniff`，清除插件设置的 Cookie；WebUI 自身继续使用第 6 节的静态规则。父桥与浏览器联验归 T-09.9。

### 7.3 SDK 页面客户端

`sdk/page` 的 `kit.js` 在 `window.AlonaBot` 上公开 `request(method, path, body?)`、`t(key, args?)`、`onTheme(fn)`、`onLocale(fn)`。嵌入时自动发送 `hello`，把令牌写到 `:root` 并设置 `data-theme`，接收主题与语言事件，转发全局搜索快捷键。两个订阅函数在注册时立即交付当前状态，并返回取消订阅函数。独立打开时使用 `light`、`en-US` 与浅色缺省令牌；`request` 返回拒绝的 Promise，不直接发送 HTTP 请求。

`t(key, args?) → Promise<string>` 在 iframe 内每次发送 `translate`，携带键和可选参数，原样返回响应的 `text`。页面在 `onLocale` 回调中重新 `await t(...)` 渲染；`hello` 和 `locale` 只交付 §7.1 所列字段，文案由 `translate` 逐次取得。独立打开时 `t` 返回兑现为键名的 Promise。

客户端按 `event.source === window.parent` 校验父窗口身份，校验 channel、响应关联 ID、对应操作的结果结构以及主题、语言事件的结构。无效消息不改变状态，也不消费在途请求。有效桥错误以 `Error` 拒绝 Promise，保留 `code`，并以 `safe_message` 为消息；`postMessage` 同步抛错时保留原错误并清理关联。客户端等待桥的有效响应，无独立固定超时。

## 8. 日志读取与输出级别

`alonabot logs`、`alonabot log-level` 与 WebUI 共用下列端点，均经第 1 节的管理认证。实时缓冲按 Core、组件 ID（Adapter 按包 ID）分别保留 `LOG_BUFFER_RECORDS` 条已输出记录，缺省每来源 1000 条；重启后清空。查询筛选只影响返回内容。

| Method / path | 查询或请求 | 成功响应 |
| --- | --- | --- |
| `GET /api/v1/logs/recent` | `source`、`level`、`trace` | 200，`Record[]`，空结果为 `[]`，按记录时间升序 |
| `GET /api/v1/logs/stream` | 同 recent | 200，`text/event-stream`，先交付初始快照，再交付新记录 |
| `GET /api/v1/logs/history` | 必填 `date=YYYY-MM-DD`；其余同 recent | 200，`Record[]`，按该日本地文件的续写序号读取；该日无文件为 `[]` |
| `GET /api/v1/logs/levels` | 无正文 | 200，`{"global":"info","sources":{}}` |
| `PUT /api/v1/logs/levels` | 同形部分更新，例如 `{"sources":{"demo":"debug"}}` | 200，返回当前有效全局值与文件中的全部来源设置 |

`source` 为空时返回全部来源，实时值为 `core` 或组件 ID。`level` 表示最低记录级别，允许 `debug/info/warn/error/critical`，省略或空串取 `debug`；`trace` 为空时返回全部追踪记录，实时查询按完整 `trace_id` 精确匹配。非法级别、非法/缺失历史日期返回 400 / `INVALID_ARGUMENT`。读取与解析历史文件的实际故障返回 500 / `INTERNAL`，并保留具体错误。

### 8.1 实时 Record 与 SSE

```json
{
  "time": "2026-11-13T10:20:30.123456789-08:00",
  "level": "DEBUG",
  "source": "demo",
  "message": "received event",
  "attrs": {
    "component": "plugin",
    "plugin_id": "demo",
    "session_id": "0123456789abcdef",
    "generation_id": "gen-example",
    "source": {"function":"example.handle","file":"/project/plugin.go","line":42},
    "trace_id": "123456789abcdef0123456789abcdef0"
  },
  "trace_id": "123456789abcdef0123456789abcdef0"
}
```

`time` 使用 RFC3339Nano；严重故障的 `level` 为 `"CRITICAL"`，其他级别沿用 `slog.Level` 的 JSON 字符串。此格式同样用于 SSE 和历史查询。`attrs` 保留实际存在的来源类别、平台/实例、Session/Generation、源码位置和调用字段。来源类别与组件身份由 Core 的安装记录绑定，Core 关于组件的故障记录仍归 `core`。实时结构化值保留完整 trace；stderr 采用 text 或 json 均相同。

每个 SSE 记录按以下格式发送：

```text
event: log
data: {"time":"...","level":"INFO","source":"core","message":"...","attrs":{},"trace_id":""}

```

订阅建立与初始快照在输出侧原子完成，快照之后的新记录进入该订阅独立的发送队列。网络写入在输出锁外进行，断开连接会取消订阅。Core 开始停机时先结束日志流并取消订阅，再等待 HTTP 请求排空；停机直接关闭流。每订阅的后续队列上限同为 `LOG_BUFFER_RECORDS`；队列满时结束该订阅，交付已在写出的记录后发送：

```text
event: overflow
data: "2026-11-13T10:20:30.123456789-08:00"

```

`overflow.data` 为本次流最后交付记录的时间；尚无已交付记录时为 `"0001-01-01T00:00:00Z"`。随后关闭流。消费者重新连接并用 recent 补齐；`alonabot logs --follow` 会打印一行说明并以非零状态退出。WebUI 可先读 recent 再订阅，重复快照由消费者处理。筛选发生在交付时，共享订阅队列接收全部已输出记录。

### 8.2 历史文件精度

历史查询使用 Core 本地时区的日期，只读取普通日志文件：先 `alonabot-<日期>.log`，再按数字序号读取 `.1.log`、`.2.log` 等。记录时间保留文件中的毫秒精度，日期与时区取查询日期和 Core 本地时区。跨夏令时重复出现的本地钟点按标准库的本地时间解析规则解释。

历史 `source` 与筛选值来自显示标签：Core 为 `core`，插件为去掉 `alonabot-plugin-` 的显示名，Adapter 为 `adapter_id` 或 `adapter_id/instance_id`。`trace_id` 保留文件中的最多 8 位前缀；传入完整 trace 时按其前 8 位匹配。文件已省略的完整包身份、Session 和 Generation 由实时结构化 API 提供。

历史 `message` 原样保留来源标签之后的全部文本，包括级别、源码位置、消息、字段及缩进多行内容。`attrs` 包含可解析的显示来源、位置、带引号/转义的字段和多行值；普通字段保留文本值。正文与属性共享文本格式，含未闭合引号、非法转义或有歧义边界的字段候选只留在 `message`，其余可靠字段继续解析。读取、I/O 和损坏记录头等实际错误继续返回。CLI 按文件原有控制台格式显示历史记录。

### 8.3 级别写回与运行时通知

`global` 允许 `debug/info/warn/error/critical`；`sources` 的值还允许 `off/inherit`。PUT 省略 global 或某个来源时保留当前值，来源设置为 `inherit` 时显式写入该值。空来源集合编码为 `{}`。GET 的 `global` 是本次 Core 的有效值，包含 `--log-level` 启动覆盖；`sources` 是设置文件中的覆盖项。级别设置使用小写 `critical`，日志记录使用大写 `CRITICAL`。

全局及来源修改先通过同一个 `settings.File.Update` 写回 `core_config.json`，然后更新 Core 阈值并向当前 Activated Session 通知 `sdk.log.set_level`。已安装组件缺省继承，开发组件缺省 debug；开发组件显式 inherit 跟随本次有效全局值。未运行来源只保存设置，下次启动使用已保存值。同进程组件只更新自己的 Session 阈值。

改全局时通知显式 inherit 与已安装的缺省继承来源，开发缺省 debug 及显式来源覆盖继续生效。来源 off 时 SDK 停止输出该进程日志，Core 关于它的记录仍由全局控制。重启沿用保存值，另一次显式 `--log-level` 只覆盖那次启动。

Rename 前保存失败保留旧设置与内存阈值；Rename 后目录同步失败按已经发布的快照同步内存并尝试当前一次通知，同时返回实际保存错误。通知失败保留已提交设置，错误响应列出受影响来源和实际原因，详细原因写入 Core 日志；下次启动继续从保存值取级别。

Cookie 写级别须同时提供受控 Origin 和 `X-Alonabot-CSRF`；Bearer 沿用管理客户端规则。认证在读取或开始 SSE 响应前完成，未认证的 recent/stream 返回 401。

CLI 用法：`alonabot logs [--source ID] [--level L] [--trace T] [--follow] [--date YYYY-MM-DD]`；缺省 recent，date 读取 history，follow 直接消费含快照的 stream；date 与 follow 同时给出时返回用法错误。`alonabot log-level [ID] [LEVEL]` 无参数列出全部设置，单独 ID 显示该设置（文件未覆盖时显示 inherit），`ID=global` 选择全局。两个命令复用 `--data-dir`、`--admin-url`、`--token-file` 与 Bearer 客户端。

## 9. 刷新后取回管理会话

浏览器刷新后，先用现有管理 Cookie 调用 `GET /api/auth/session`，取回当前会话原有的 CSRF 和固定到期时刻。成功为 200、`Content-Type: application/json`，并带 `Cache-Control: no-store`：

```json
{"csrf_token":"<当前会话原有值>","expires_at_unix_ms":1790000000000}
```

此接口只接受有效管理 Cookie。仅有 Bearer、Cookie 无效、过期、已登出或认证代次失配时返回 401。管理凭证未初始化返回 503；状态库故障返回 500，真实错误记录到 Core 日志。错误使用第 1/2 节的英文纯文本格式。成功读取保持会话 ID、CSRF、认证代次和已保存的到期时刻，响应无 `Set-Cookie`。后续变更请求继续携带这个 CSRF 与受控 Origin。

## 10. 浏览器上传安装与更新包

`POST /api/v1/packages/install` 和 `POST /api/v1/packages/update` 同时接受 `multipart/form-data; boundary=...`。现有 JSON 入口继续使用第 3 节的原请求结构、大小写和响应；JSON 的 Archive 仍为 Core 本机路径。

浏览器每次上传一组本地候选，part 名严格区分大小写：

| Part | 格式 | 规则 |
| --- | --- | --- |
| `archive` | 文件 | 必填且一份；归档字节，通常文件名为 `<id>-<version>-<platform>.tar.gz` |
| `release` | 文件 | 必填且一份；release.json 原始字节 |
| `signature` | 文件 | 有签名时提供 release.json.sig 原始字节；无签名包可省略 |
| `public_key` | 文本 | release.pub 的标准 base64 内容，映射 `InstallRequest.PublicKey` |
| `confirm` | 文本 | 可选，严格为 `true` 或 `false`；省略为 false，映射 `InstallRequest.Confirm` |
| `replace` | 文本 | 仅 install 可选，严格为 `true` 或 `false`；省略为 false，映射 `InstallRequest.Replace` |
| `id` | 文本 | 仅 update 必填，指定已安装组件 ID |

文件 part 在 Content-Disposition 中带 `filename` 参数（其值可为空），文本 part 省略该参数；各 part 可任意排序。未知或重复 part、缺少必需文件或更新 ID、错误布尔值和截断 multipart 返回 400 / `INVALID_ARGUMENT`。上传固定为本地候选，`Source` 留空；来源从 release/manifest 与既有绑定校验。客户端文件名只作展示，服务器自行生成目录与文件名。

有签名时对 release.json 的确切字节验签，提供 `archive`、`release`、`signature` 三个文件与 `public_key` 文本。缺少或错误公钥、错误签名与来源绑定等错误经既有服务返回 403 / `PERMISSION_DENIED`。无签名包省略 signature，继续执行归档长度、摘要、平台、manifest、身份、来源、依赖和显式信任确认；列表的 `Signed` 为 false。已经绑定签名的来源继续要求同一公钥，验签失败按原码返回。

认证先于上传解析：Cookie 变更请求先验证 Origin/CSRF，Bearer 沿用第 1 节规则。archive 流式写入 `<data_dir>/staging` 中独占的请求目录（0700），文件权限为 0600，大小沿用 `packages.Limits.MaxArchiveBytes`（缺省 256 MiB）；解压总量与条目上限由同一 Preflight 执行。超限沿用 `RESOURCE_EXHAUSTED`，其 HTTP 状态按第 5 节当前映射为 500。解析层保持 release/signature 精确字节，安装和更新均调用同一管理服务，预检与确认发生在停止或修改旧组件前。

install 成功为 200、§3 的 `Package`（有警告时带 `Warnings`）；update 成功为 200、§3 的 `UpdateResult`（没有警告时为 `{}`）。确认响应为 409、`{"affected":[...]}`，浏览器展示全部对象，经管理员确认后，重新上传同一组文件并设置文本 `confirm=true`。JSON 客户端继续使用 install 的 `Confirm` 或 update 外层 `confirm`。

临时目录覆盖完整 Install/Update 调用；成功、400/403/409、候选失败、请求中断和文件写入失败均执行清理。关闭或清理的真实错误保留在管理错误或日志中。服务调用结束后归档源可删除；已安装不可变目录及 Journal 接管资源继续由各自所有者管理。

multipart 与 JSON Update 均进入 `manage.Service.Update`，执行绑定来源、组件身份、公钥、pin 与依赖影响检查。Core 已接通候选准备与提交：有效的新版本上传成功返回 200、§3 的 `UpdateResult`（没有警告时为 `{}`），已启用的组件激活新代次，用户数据保留，受管 SQLite 在候选副本上执行迁移。缺文件或验签失败在修改旧组件前拒绝；重复上传已安装的同版本同归档返回 409 / `ALREADY_EXISTS`。

## 11. Core 总览

`GET /api/v1/core` 使用第 1 节的 Bearer 或有效 Cookie 管理认证，成功为 200、`application/json`：

```json
{"version":"v0.1.0","listen":"127.0.0.1:7650","admin_api_major":1}
```

`version` 来自 `runtime/debug.ReadBuildInfo` 中 AlonaBot module（`github.com/AlonaBot/AlonaBot`）的版本：作为主 module 取 Main，作为库取对应依赖项；依赖存在 `Replace` 时取替换模块的版本，本地路径替换或其他缺少版本的情况为 `(devel)`。`listen` 是当前 Core 的 `listener.Addr().String()`，包含启动覆盖及端口 0 实际分配结果。`admin_api_major` 使用 `manage.AdminAPIMajor`，保持为 1。响应只包含以上三个字段。

## 12. 组件配置、Pending 与秘密查看

本节及后续端点均使用第 1 节的管理认证；Cookie 变更请求携带 Origin 与 CSRF。路径中的 `{id}` 为已安装组件 ID。成功响应为 200 JSON，错误遵循第 5 节。

### 12.1 配置快照

`GET /api/v1/plugins/{id}/config` 返回：

```json
{
  "active_generation": "gen-example",
  "active": {"token":{"$secret_set":true},"hidden":"value"},
  "pending": {"base_generation":"gen-example","config":{"token":{"$secret_set":true},"hidden":"value"}},
  "schema": {"token":{"type":"string","secret":true}},
  "i18n": {"zh-CN":{"token":{"description":"访问凭证"}}},
  "corrections": [],
  "unknown": []
}
```

没有 Active 时 `active_generation` 为空串、`active` 为 `{}`；没有有效 Pending 时 `pending` 为 null。`schema` 是当前安装包 `config_schema.json` 的规范内容，缺失时为 null。`i18n` 按包内 `i18n/*.json` 的语言标签返回各文件的 `config` 根节点，缺失时为 `{}`。读取隐藏秘密引用、schema 声明的秘密及代次保留的秘密路径；`invisible` 字段照常返回，由表单决定是否展示。

| Method / path | 请求 | 成功响应与语义 |
| --- | --- | --- |
| `PUT /api/v1/plugins/{id}/config/pending` | `{"base_generation":"gen-example","config":{...}}` | 返回 `{base_generation,config}`，完整替换编辑值，响应脱敏 |
| `PATCH /api/v1/plugins/{id}/config/pending` | `{"base_generation":"gen-example","patch":{...}}` | 同上；按 RFC 7396 合并，null 删除键、数组整体替换、对象递归合并 |
| `DELETE /api/v1/plugins/{id}/config/pending` | 无正文 | `{}`，丢弃 Pending，保持 Active |
| `POST /api/v1/plugins/{id}/config/import` | 无正文 | 返回脱敏的 `{base_generation,config}`，导入 `<data_dir>/plugins/<id>/config.json`，保持 Active |
| `POST /api/v1/plugins/{id}/reload` | 无正文 | `{}`，要求组件 Activated，经候选编排激活有效 Pending；没有 Pending 时先导入编辑快照 |

PUT/PATCH 必须提供 `base_generation`，首次代次可提交空串；与当前 Active 不同时返回 **409 / `FAILED_PRECONDITION`**。编辑以有效 Pending 为起点，否则以 Active 为起点。新秘密在写入 Pending 前存为不可变引用；相同位置未改变的秘密复用原引用。PUT 省略的秘密路径保留；PUT/PATCH 提交数组时，仍存在的相同下标内省略的秘密保留，删除行或显式 null 沿用删除语义；PATCH 省略的字段继续保留，包含 invisible 字段。提交秘密引用或恰为 `{"$secret_set":布尔值}` 的读取掩码返回 400 / `INVALID_ARGUMENT`；含其他键的普通对象可保存 `$secret_set` 字段。导入文件缺失返回 404。

安装包保存操作只更新 Pending；两个 reload HTTP 入口共用候选编排，优先激活有效 Pending。`config/import` 总是显式导入磁盘编辑，保持 Active。CLI `alonabot reload <id>` 先调用 config/import，再调用 reload，因此直接编辑 config.json 后执行 CLI reload 继续导入文件内容。

开发来源从当前工程资源读取 schema/i18n，GET 返回固定 `dev:<id>` 代次的脱敏 Active、corrections 与 unknown。PUT/PATCH 使用同样的 base 和秘密省略规则，先 Complete/Validate，再以临时文件替换 `<data_dir>/plugins/<id>/config.json`，随后经 Core 的开发重启入口建立新 Session；返回脱敏 `{base_generation,config}`，保持 Pending 为 null。保存失败保留原文件；重启失败返回实际错误并保留已保存文件，调用方可修订配置后再次重启。dev 的 config/import 只读取编辑文件并返回脱敏结果，缺文件时返回脱敏 Active，CLI 随后的 reload 仍按原有路径启动 schema 缺省配置。dev 的 prune 从 Active 写入 Pending；管理 API reload 有有效 Pending 时将其原子写入编辑文件，成功启动后清除 Pending，失败返回实际错误并保留 Pending；后续 restart 读取已清理的文件。WebUI 配置返回 400。

**校正与无效配置（D-161）。** 插件配置、Core 设置与 Adapter 实例的读取响应各多两项：`corrections` 为校正记录数组，元素是 `{"path":"…","reason":"converted|reset|removed","original":…}`，涉及 secret 的字段或容器没有 `original`，与历史秘密路径相交的记录同样省略；`unknown` 为无效配置（schema 中没有的键）的 JSON 路径数组。插件配置的两项取自 Active 代次，在 `GET /api/v1/plugins/{id}/config` 的顶层；Core 设置用 `GET /api/v1/core/settings/corrections` 返回 `{"corrections":[…],"unknown":[…]}`；Adapter 实例在 `GET /api/v1/adapters/{adapter_id}/instances/{instance_id}` 的顶层。清理：`POST /api/v1/plugins/{id}/config/prune` 以 Active 为起点删除全部无效配置并写入 Pending，返回脱敏的 `{base_generation,config}`，随后照常 reload；`POST /api/v1/core/settings/prune` 与 `POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune` 经 `settings.File.Update` 删除这些键并写回 `core_config.json`，成功返回 `{}`，保留其他键的 `${...}` 写法。

### 12.2 再认证查看秘密

`POST /api/v1/secrets/{plugin}/{path}/reveal` 接收 `{"token":"<当前管理 token>"}`，在既有管理认证后再次调用 `adminauth.Reauthenticate`。失败返回 403 / `PERMISSION_DENIED`；成功返回 `{"value":"<秘密明文>"}`，附加 `Cache-Control: no-store`，并以 INFO 记录组件 ID 与字段路径。

`path` 使用秘密字段的点分路径，数组下标占一段；字段名中的 `.` 编码为 `~1`，`~` 编码为 `~0`，然后将整个路径作为一个 URL 段转义。读取有效 Pending 中的引用，否则读取 Active；开发来源按 schema 确认秘密路径后读取明文。路径未指向秘密字段时返回 400。审计记录省略明文与 token。

## 13. Core 设置与发信许可

### 13.1 Core 设置

| Method / path | 请求或成功响应 | 操作语义 |
| --- | --- | --- |
| `GET /api/v1/core/settings/corrections` | `{"corrections":[],"unknown":[]}` | 返回当前运行设置的校正记录与无效配置路径（§12.1），附加 `Cache-Control: no-store` |
| `POST /api/v1/core/settings/prune` | 无正文，成功返回 `{}` | 重读磁盘并删除 Core schema 之外的键，经 `settings.File.Update` 写回，管理员策略跟随同一次发布的设置 |
| `GET /api/v1/core/admins` | `["telegram:123"]` | 返回当前生效的、排序后的用户字符串 |
| `PUT /api/v1/core/admins` | 同形字符串数组 | 整份替换，空数组清空；格式错误或展开后重复的身份返回 400 |
| `DELETE /api/v1/core/admins` | 同形字符串数组 | 删除指定用户，未列出的用户保留 |
| `GET /api/v1/core/default-locale` | `{"default_locale":"en-US"}` | 当前有效语言 |
| `PUT /api/v1/core/default-locale` | 同形对象 | 必填且展开后非空，之后新建的 Session 使用新值 |
| `GET /api/v1/core/denies` | `{"users":["telegram:123"],"chats":["telegram:room"]}` | 全局禁止表，两个数组均返回，空列表为 `[]` |
| `PUT /api/v1/core/denies` | 同形对象 | 两个数组必填，整体替换 |
| `DELETE /api/v1/core/denies` | 同形对象 | 两个数组必填，删除指定用户与会话 |

命令前缀与回复方式按配置文件设置，`default` 的值写回 `core_config.json`，接口见 §23（D-168）。

写接口成功返回 `{}`。管理员、语言和前缀经 `settings.File.Update` 写回 `core_config.json`，保留未修改字段的 `${...}` 原文；请求中的完整环境变量引用按原文保存、按展开值校验。管理员列表写回后调用 `policy.SetAdmins`；现有 `/admins` 与 CLI `admin add/remove/list` 共用同一管理服务和设置。前缀由下一次消息派发读取，实例未配置覆盖时使用全局值。Rename 后目录同步失败时设置已发布，接口返回实际错误，同时管理员策略跟随当前设置更新。

禁止表保存在状态库。用户使用 `<adapter_id>:<external_user_id>`；会话使用不含话题的会话字符串，外部会话 ID 中的 `%`、`#` 分别编码为 `%25`、`%23`。输入包含话题或非规范编码时返回 400，检查发生在写入前。批量替换先增加新限制，再解除旧项；各项独立提交，发生存储错误时返回实际错误，已提交的限制保留，调用方可 GET 核对当前状态。

### 13.2 插件发信许可

`GET /api/v1/plugins/{id}/allowed-instances` 返回 `{"all":true,"instances":[]}`。PUT 接收同形对象并整份替换；省略 `all` 时为 true。`all:false` 将后台发信限定到 `instances`，空列表拒绝所有实例；`all:true` 允许全部实例。实例名格式错误返回 400。许可保存至共享状态库，下一次发信使用新策略。

## 14. Adapter 实例配置

| Method / path | 请求或成功响应 | 操作语义 |
| --- | --- | --- |
| `GET /api/v1/adapters/{adapter_id}/instance-schema` | `adapter_config_schema.json` 的规范 JSON 对象 | 声明通过 `configschema.Parse` 读取 |
| `GET /api/v1/adapters/{adapter_id}/instances/{instance_id}` | `{"config":{"token":{"$secret_set":true},"name":"bot"},"corrections":[],"unknown":[]}` | 返回校正后的实例配置及记录、无效配置路径；秘密字段只显示设置状态，附加 `Cache-Control: no-store` |
| `POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune` | 无正文，成功返回 `{}` | 从磁盘原文删除实例 schema 之外的键，保留其他值与 `${...}` 写法 |
| `PUT /api/v1/adapters/{adapter_id}/instances/{instance_id}` | `{"config":{...}}` | 创建或整体替换指定实例配置，成功返回 `{}` |
| `DELETE /api/v1/adapters/{adapter_id}/instances/{instance_id}` | 无正文 | 与 §4 的实例删除共用清理流程；实例归属的 Adapter 不匹配时返回 404，保留该实例及其指定 |

实例名跨 Adapter 唯一，其他 Adapter 已占用时返回 409 / `ALREADY_EXISTS`；未知 Adapter、未知实例或删除时 URL 的归属失配返回 404。配置块属于其他组件时返回 412。

提交先经 `Schema.Complete` 补全，再对展开环境变量后的副本执行 `Schema.Validate`，最后用 `File.Update` 写回原文。省略的秘密从原始文档保留，数组中仍存在的相同下标也遵循此规则。提交秘密引用或恰为 `{"$secret_set":布尔值}` 的掩码返回 400。读取时 `PrepareInstances(adapterID, schema)` 独立校正各实例的运行副本，再脱敏返回；required 失败的实例仍可读取校正结果与记录以便修订。显式编辑使用 `PrepareInstances(adapterID, schema, replacement)`，在含替换值的候选文档上校正其他实例；存在失败实例时返回带完整 `Path`、`Reason` 的 412，原文件保留。通过后仅将显式 replacement 经 `File.Update` 写回，其他实例保留原文。校正与清理均沿用 §12.1 的记录及路径格式。

保存配置后通过第 4 节的启停入口管理运行实例。启动前仍由 `File.PrepareInstances` 校正运行副本并检查 required，失败实例独立列入 `InstanceProblem`，其余实例继续启动；路径与原因原样用于定位，原文件保持。

## 15. Agent 会话管理

`GET /api/v1/plugins/{id}/conversations` 返回按键排序的摘要数组，空结果为 `[]`：

```json
[{"key":"telegram:room@assistant","entries":3,"updated_at":"2026-11-13T12:00:00Z"}]
```

`DELETE /api/v1/plugins/{id}/conversations/{key}` 清除该组件指定会话，成功返回 `{}`。`key` 是会话存储键，Agent 按 `<会话字符串>@<实例名>` 组织，开启成员独立会话（`member_sessions`）时群聊成员的键再追加 `@<用户字符串>`；调用方把整个键按一个 URL 段转义，服务端解码一次后原样交给 `conversation.Service.Clear`。该端点在路径清理前接收键，因此键中的斜杠、冒号、反斜杠和点段保留。键须为 1～256 字节的有效 UTF-8，格式错误返回 400；未知组件返回 404。列表仅返回键、条目数与最后时间。


### 15.1 Agent 模型管理路由（D-154）

路径前缀 `/plugins/{agent_id}/models/`。使用 Core 管理 Cookie；POST 通过 `Origin` 和 `X-Alonabot-CSRF` 校验。路由自身的响应为 JSON，Core 层的鉴权拒绝沿用既有响应。

| 方法与子路径 | 请求体 | HTTP 200 响应 |
| --- | --- | --- |
| `POST discover` | `{"provider":"配置名"}` 或 `{"protocol":"openai-compatible","base_url":"…","api_key":"…"}`；无密钥 Ollama 可省略 `api_key` | `{"models":[{"id":"远端模型 ID","type":"chat"}]}`；不支持列表时 `{"models":[],"description":"Provider does not support listing models."}`。 |
| `POST test` | `{"model":"Active 模型配置名"}` | `{"ok":true,"duration_ms":12}`；Provider 探测失败仍返回 200，`ok:false` 且附英文 `description`。 |
| `GET usage` | 无 | `{"models":[{"model":"模型配置名","input_tokens":7,"output_tokens":3}]}`，按模型名排序，仅含 Agent 本地日期当天实际收到用量的模型；空结果为 `{"models":[]}`。 |

应用层的 400（无效输入）、404（找不到配置项）、502（发现上游失败）和 500（读取用量失败）返回 `{"description":"…"}`，说明为受控英文。模型 ID 的推断规则依次匹配 `embed`、`rerank`、`whisper`/`transcribe`、`tts`，其余为 `chat`。测试路由不落账；有效的零 token 响应存为零，缺失用量不建记录。

## 16. 命令权限、启用与命令名

`GET /api/v1/plugins/{id}/policies` 返回：

```json
{
  "overrides":[{"scope":{"kind":"route","route_id":"cmd.ping"},"permissions":["admin"],"enabled":null}],
  "commands":[{"route_id":"cmd.ping","enabled":true,"permissions":["admin"],"declared_names":[["ping"]],"effective_names":[["check"],["c"]],"has_names_override":true,"invalid_names":[]}],
  "shadows":[{"path":["check"],"winner":{"plugin_id":"demo","route_id":"cmd.ping"},"shadowed":[{"plugin_id":"other","route_id":"cmd.check"}]}],
  "invalid_names":[]
}
```

`PUT /api/v1/plugins/{id}/policies` 接收一项 `{scope,permissions,enabled}`，整份替换该作用域的权限与启用覆盖，成功返回 `{}`。`scope.kind` 使用 `plugin/group/route`；group 另带 token 数组 `group`，route 另带 `route_id`。permissions 为 null 或省略时继承，`[]` 表示无权限要求；enabled 为 null 或省略时继承，false 禁用。两项都继承时清除该作用域覆盖。输入校验及保存复用 `policy.SetOverride`，GET 的有效权限/启用与派发的 `CheckCommand` 共用 `CommandState` 规则；子级启用保持父级禁用的效果。覆盖在下一次派发生效。

命令列表来自当前 Router 注册声明，权限覆盖保留未注册项。名称只对路径命令返回；Raw 与文本匹配的 declared_names/effective_names 为空数组。has_names_override 表示保存过覆盖，覆盖失效后仍为 true。invalid_names 为 `{route_id,name,reason}`，reason 为 `declared`（与其他当前声明名冲突）或 `unregistered`（路由未注册或非路径命令，name 为 null）；顶层包含当前插件的全部诊断，命令项包含该路由的诊断。shadows 只列涉及该插件的关系，来自 Router 按生效命令名建立的索引。

| Method / path | 请求 | 语义 |
| --- | --- | --- |
| `PUT /api/v1/plugins/{id}/commands/{route_id}/names` | `{"names":[["check"],["c"]]}` | 整份替换该路径命令的名称，第一项为主名，成功返回 `{}` |
| `DELETE /api/v1/plugins/{id}/commands/{route_id}/names` | 无正文 | 删除覆盖、恢复声明名称，也可清除已移除路由的失效覆盖，成功返回 `{}` |

名称修改复用 policy 的格式、声明与重名校验；空列表、空路径、非法 token、Raw 或文本匹配路由返回 400 / INVALID_ARGUMENT。同插件声明或覆盖重名返回 409 / ALREADY_EXISTS，safe_message 包含冲突 route_id。调用方增删单个名称时先在列表上编辑，再整份提交。保存和删除后下一次派发使用新索引，无需重启组件。

## 17. 配置字段与 Core 选项

所有选项响应为 `{options:[{value,label,description?}],more}`，空列表为 `[]`。列表请求用 `q` 传搜索文本；resolve 请求返回当前仍存在的存储值，缺失值省略，空 values 返回空列表。两种查询沿用共享 options.Service 的 OPTIONS_PAGE 上限。

| Method / path | 请求 |
| --- | --- |
| `GET /api/v1/plugins/{id}/config/options` | `?path=字段路径&q=搜索文本` |
| `POST /api/v1/plugins/{id}/config/options/resolve` | `{"path":"字段路径","values":["已存值"]}` |
| `GET /api/v1/adapters/{adapter_id}/instances/{instance_id}/options` | 同插件字段列表查询 |
| `POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/options/resolve` | 同插件字段 resolve |
| `GET /api/v1/core/options/{source}` | `?q=搜索文本&filter[键]=值` |
| `POST /api/v1/core/options/{source}/resolve` | `{"values":["已存值"]}`；可带同形 filter 查询参数 |

插件与 Adapter 字段的来源名和 filter 只从插件 Active 代次的 schema 取得；额外的 source/filter 查询参数忽略。字段使用 schema 的点分路径，列表字段省略下标；不存在的字段或没有 options_from 返回 400 / INVALID_ARGUMENT。Adapter 实例入口要求该 Adapter 和实例存在，归属不匹配返回 404。停用组件仍可查询内置来源；Core 重启后从 Active 安装包或开发资源读取声明。组件来源要求提供该来源的运行中 Session，服务来源、Agent 或 Active schema 不可用时返回 503 / UNAVAILABLE。

界面用 `Accept-Language` 传当前管理语言，取首选语言标签；缺省使用 Core default_locale。Core 直接查询仅接受 core.chats（filter 键 chat_type、adapter_id）和 core.adapter_instances（filter 键 adapter_id）；服务来源与未知 filter 键返回 400。core.chats 可按名称或聊天 ID 搜索并按类型筛选；按值取回沿用来源的 resolve 语义。

## 18. 工具调用策略

`GET /api/v1/tools/policies` 返回 `{"policies":[{"plugin_id":"demo","service":"echo","default":"allow","override":"inherit","effect":"read_only"}]}`；`effect` 取 `read_only`、`write`、`destructive` 或 `unspecified`（D-157）。目录按 `plugin_id` 和 `service` 合并当前 Attach 的工具声明、本次 Core 运行的确认记录和已保存覆盖。有确认记录时采用最近上报的类别（包括 `unspecified`），否则采用声明类别，仅有保存覆盖时为 `unspecified`。`default` 优先取声明的缺省策略，仅有确认记录时为 `confirm`，两者都没有时为 `inherit`；`override` 为 `inherit` 表示没有保存覆盖。

`PUT /api/v1/tools/policies` 接收 `{"plugin_id":"demo","service":"echo","policy":"deny"}`，成功返回 `{}`。policy 接受 allow、confirm、deny、inherit；inherit 删除覆盖。service 也接受目录已记录的完整操作名，例如 `mcp/filesystem/read_file`；操作名独立于 Service 名语法校验。未知策略、服务名格式错误或目录未记录的操作名返回 400，未知组件返回 404。写入与固定 ListTools 使用同一个 tool.Catalog 和状态库；下次查询或调用读取新策略，deny 的工具从 ListTools 隐藏，无需重启。

`PUT /api/v1/tools/policies` 也接收 `{"items":[{"plugin_id":"demo","service":"echo","policy":"deny"}]}`。逐项严格解码和校验，全部合法后，在一个事务内保存或删除覆盖；写入失败时整批回滚。项内字段类型错误、未知字段、非法值或未知组件返回 400（INVALID_ARGUMENT），`safe_message` 以 `items[index]:` 标注从 0 开始的项号（例如 `items[1]: unknown tool policy`）；校验失败时整批不写（D-157）。

## 19. 诊断包

`POST /api/v1/diagnostics` 经管理认证（Cookie 认证时校验 CSRF），接收 `{"since":"24h","logs":true,"settings":true,"components":true,"redact":true}`，各字段可省略，缺省值如上。`since` 无法解析为正时长、三组都为 `false`、JSON 无效、出现未知字段或正文含尾随内容时返回 400。

归档完整生成后返回 `application/zip`，`Content-Disposition` 的附件名为 `alonabot-diagnostics-<UTC 时间>.zip`，并设置 `Cache-Control: no-store`。生成失败按第 3 节返回 JSON 错误。内容与脱敏规则见 CONTRACTS §14.7（D-156）。

## 20. WebUI 需要的补充接口

以下端点经管理认证；使用 Cookie 的写入请求验证 CSRF，错误规则同第 3 节（D-164）。

| Method / path | 请求或成功响应 | 语义 |
| --- | --- | --- |
| `GET /api/v1/plugins/{id}/generations` | `{"generations":[{"id":"gen-…","version":"1.2.1","created_at":"RFC 3339","active":true,"has_db":true}]}` | 保留代次按提交时间降序；回退仍用 `/api/v1/packages/rollback` |
| `GET /api/v1/plugins/{id}/jobs` | `{"jobs":[{"job_id":"…","trigger":"cron 0 8 * * *","paused":false,"last_run_at":"…","next_run_at":"…"}]}` | 当前 Session 登记的静态任务；Detach 后返回 `{"jobs":[]}`。`trigger` 如 `interval 1h0m0s` 或 `cron 0 8 * * *`，没有记录的时刻省略 |
| `POST /api/v1/plugins/{id}/jobs/{job_id}/pause`、`…/resume` | 无正文，`{}` | 与 `core.job.pause/resume` 共用持久暂停状态（CONTRACTS §4.3，D-126）；未知任务返回 404 |
| `GET /api/v1/adapter/instances` | `{"instances":[{"instance_id":"main","adapter_id":"telegram","package_id":"…","state":"running","safe_error":"","default":true}]}` | 按实例 ID 排列已配置的实例；`state` 为 `running`、`stopped` 或 `failed`；`default` 遵循显式默认与单实例自动默认规则 |
| `GET /api/v1/chats` | `?q=词&adapter_id=telegram&chat_type=group&offset=0&limit=50`，返回 `{"chats":[{"session":"telegram:-100…","title":"…","chat_type":"group","instance_id":"main","last_active_at":"…"}],"total":53}` | 聊天目录（D-130），筛选后按最近活跃降序，`total` 为筛选后的精确总数 |
| `POST /api/auth/token/rotate` | `{"token":"<当前管理 token>"}`，返回 `{"token":"<新 token>"}` | 再认证后在同一事务中替换令牌并删除全部管理会话（包括当前会话）；响应清除会话 Cookie，带 `Cache-Control: no-store`，只交付一次新令牌 |

- **代次：** `created_at` 取 generation 元数据；安装来源的 `version` 取该代次的 committed Journal 程序快照，缺失时返回错误，`has_db` 检查 `plugins/{id}/generations/{generation}/db.sqlite`。开发组件的固定代次使用当前登记版本，`has_db` 检查 `plugins/{id}/dev.sqlite`。路径均相对于 `data_dir`。
- **任务：** 暂停前已派发的调用可以完成，恢复沿原时间锚点安排后续触发。`job_id` 按 SDK/REGISTER 规则取原值，客户端将其百分号编码为单个不透明路径段；Core 在认证后、路径规范化前只解码一次，保留冒号、斜杠、点段、字面百分号和反斜杠。
- **实例：** `failed` 和 `safe_error` 来自当前 Core 中保存的最近失败快照。启动失败使用协议 `SafeMessage` 或通用内部失败说明，配置失败只含路径和诊断原因。成功启动、成功停止或忘记实例配置时清除快照；`running` 和 `stopped` 的 `safe_error` 为空。
- **聊天：** `offset` 缺省 0，`limit` 缺省 50，超过生效 `OPTIONS_PAGE` 十倍时取上限。负数、零 `limit`、非整数和溢出的分页值返回 400。`chat_type` 可省略或为空，非空时接受 `private`、`group` 或 `channel`，其他值返回 400。`session` 按 MESSAGE-MODEL §1 将 chat ID 中的 `%` 与 `#` 分别编码为 `%25` 与 `%23`。
- **页面：** `GET /api/v1/packages` 的列表项带 `pages`，仅列当前运行路由中声明的页面，无运行路由登记时为 `[]`。例如 `i18n/zh-CN.json` 为 `{"pages":{"/subs/":"订阅"}}`，`Accept-Language: zh-CN` 时返回 `[{"path":"/subs/","title":"订阅"}]`；缺少匹配的覆盖时用 SDK 声明的英文标题。路径相对于插件，WebUI 用沙箱 iframe 打开 `/plugins/{id}` 加该路径（§7）。

## 21. 插件数据浏览

以下端点由 T-09.21 实施，经管理认证，非 GET 验证 CSRF（D-163）。表与 KV 作用于受管 SQLite；缺少受管数据库时，相应端点返回 404。数据库的 `generation` 查询参数缺省为 Active 代次，指定保留代次时只读。文件端点作用于插件当前的 `<data_dir>/plugin_data/<id>/`，该目录在代次之间共用（CONTRACTS §6.1）。

| Method / path | 请求或成功响应 | 语义 |
| --- | --- | --- |
| `GET /api/v1/plugins/{id}/data/tables` | `{"tables":[{"name":"subs","rows":42,"writable":true,"columns":[{"name":"id","type":"INTEGER","pk":true,"notnull":true}]}]}` | `_alonabot_migrations` 列出但 writable 为 false；`WITHOUT ROWID` 的表 writable 为 false |
| `GET /api/v1/plugins/{id}/data/tables/{table}/rows` | `?offset=&limit=&order=列&desc=1&where.列=值`，返回 `{"rows":[{"rowid":1,"values":{…}}],"total":42}` | limit 缺省 50、最多 500；`where` 为等值筛选，可以有多个；BLOB 以 `{"$base64":"…"}` 表示 |
| `POST /api/v1/plugins/{id}/data/tables/{table}/rows` | `{"values":{…}}`，返回 `{"rowid":43}` | 插入一行 |
| `PATCH /api/v1/plugins/{id}/data/tables/{table}/rows/{rowid}` | `{"values":{…}}`，`{}` | 修改给出的列 |
| `DELETE /api/v1/plugins/{id}/data/tables/{table}/rows/{rowid}` | 无正文，`{}` | 删除一行 |
| `GET /api/v1/plugins/{id}/data/export` | `application/zip` | 含 `db.sqlite`（经 SQLite Backup API 的一致快照）与 `plugin_data/`（插件数据目录），D-165 |
| `POST /api/v1/plugins/{id}/data/clear` | `{"confirm":true}`，`{}` | 删除全部可写表的行（含 `_alonabot_kv`）与数据目录中的全部文件，保留表结构、迁移记录与目录本身；没有 confirm 时 400 |
| `GET /api/v1/plugins/{id}/data/kv` | `?prefix=&offset=&limit=`，返回 `{"items":[{"key":"…","size":12,"updated_at":"…"}],"total":128}` | KV 键列表（D-165），按键排序 |
| `GET /api/v1/plugins/{id}/data/kv/{key}` | `{"value_base64":"…","text":"…"}` | 值是合法 UTF-8 时另给 text；不存在 404 |
| `PUT /api/v1/plugins/{id}/data/kv/{key}` | `{"value_base64":"…"}` 或 `{"text":"…"}`，`{}` | 写入或覆盖 |
| `DELETE /api/v1/plugins/{id}/data/kv/{key}` | 无正文，`{}` | 删除；不存在 404 |
| `GET /api/v1/plugins/{id}/data/files` | `?path=子目录`，返回 `{"entries":[{"name":"card.png","dir":false,"size":20480,"modified_at":"…"}]}` | 列出数据目录（D-165），目录在前、按名称排序 |
| `GET /api/v1/plugins/{id}/data/files/content` | `?path=文件`，`application/octet-stream` | 下载文件，`Content-Disposition: attachment` |
| `PUT /api/v1/plugins/{id}/data/files/content` | `?path=文件&overwrite=1`，请求体为文件内容，`{}` | 先写临时文件再改名；已存在且没有 overwrite 时 409 |
| `POST /api/v1/plugins/{id}/data/files/mkdir`、`…/rename`、`…/delete` | `{"path":"…"}`、`{"from":"…","to":"…"}`、`{"path":"…","recursive":false}`，`{}` | 删除非空目录须 recursive |

KV 键遵守 CONTRACTS §6.1 的键约束，`/` 可作为键的字面字符。文件路径是相对于数据目录的 `/` 分隔路径，拒绝空段、`.`、`..`、绝对路径与反斜杠，不跟随符号链接，指向数据目录之外时 400。写入在插件运行中照常进行：Core 以 busy_timeout 打开数据库，每个数据库请求一个事务，等待超时返回 409；约束错误返回 400 并带 SQLite 的英文说明。每次写入记一条 INFO 审计日志，表行写入只记组件、表与 rowid。

## 22. 更新内容（changelog）

| Method / path | 成功响应 | 语义 |
| --- | --- | --- |
| `GET /api/v1/packages/{id}/changelog` | `{"locale":"en-US","entries":[{"version":"1.2.1","date":"2026-11-17","markdown":"…"}]}` | 当前已安装版本包中的 changelog（D-166）。单文件按围栏代码块之外的 `## [版本] - 日期` 或 `## [版本]` 切分，版本须为 SemVer，省略 `[Unreleased]`；`locale` 为 `en-US`。目录形式逐版本选择语言，`locale` 为最新条目所用文件的标签，条目省略 `date` |
| `GET /api/v1/core/changelog` | 同上 | 读取编进可执行文件的根 `CHANGELOG.md`，按单文件规则返回 |

条目按 SemVer 倒序，`markdown` 保留原文（单文件只返回节正文），日期缺省时省略 `date`。无发布条目时返回 `locale: "en-US"` 与 `entries: []`。

目录形式按 `Accept-Language` 的偏好顺序为各版本独立选择语言。每个偏好依次尝试精确文件名、规范化后的标签相等、规范化后的主语言相等；同一级有多个候选时取文件名字典序最小者。无匹配时使用 `en-US.md`。

两个端点均需管理认证，错误规则同第 3 节；未安装、已卸载或以 `source=dev` 运行的组件返回 404。Markdown 原文由 WebUI 渲染，WebUI 不执行其中的 HTML。

## 23. 配置文件

配置文件的语义见 CONTRACTS §5.6（D-168）。ID `default` 表示「默认」。管理写入立即保存；`default` 的字段值经 §12 的配置 API 写入 Pending，平台设置写回 `core_config.json`。运行时应用由后续切片接入。

| Method / path | 请求或成功响应 | 操作语义 |
| --- | --- | --- |
| `GET /api/v1/profiles` | `[{"id":"default","name":"默认","color":"teal","sessions":50}]` | `default` 在首位，其余按创建顺序 |
| `POST /api/v1/profiles` | `{"name":"VIP 群","color":"navy","copy_from":"default"}` → `{"id":"…"}` | `copy_from` 可省略；名称重复 409 |
| `PATCH /api/v1/profiles/{id}` | `{"name":"…","color":"…"}` → `{}` | 两键均可省略 |
| `DELETE /api/v1/profiles/{id}` | `{}` | `default` 返回 400；指定它的会话与实例回到下一级 |
| `GET /api/v1/profiles/assignments` | `{"instances":{"qq-main":"vip"},"sessions":[{"instance":"qq-main","session":"onebot11:group:123","profile":"vip"}]}` | `instances` 只列出设置了缺省的实例；空集合为 `{}` 与 `[]` |
| `PUT /api/v1/profiles/assignments/instances/{instance}` | `{"profile":"vip"}`，`{"profile":null}` 或正文 `null` 取消 → `{}` | 实例不存在 404 |
| `PUT /api/v1/profiles/assignments/sessions` | `{"instance":"qq-main","session":"…","profile":"vip"}`，profile 为 `null` 取消 → `{}` | 会话字符串按 MESSAGE-MODEL §1 校验 |
| `GET /api/v1/profiles/{id}/plugins` | `[{"plugin_id":"weather","enabled":true,"fields":["city","unit"],"overrides":{"city":"北京"},"invalid":[{"path":"unit","reason":"type"}]}]` | 列出已安装的插件与 Agent。`fields` 为 Active schema 标记的字段路径；`default` 的 `overrides` 与 `invalid` 为空 |
| `PATCH /api/v1/profiles/{id}/plugins/{plugin_id}` | `{"enabled":false,"overrides":{"city":"北京","unit":null}}` → `{}` | 两键均可省略，字段值为 `null` 时删除该项；其余值按 Active schema 单值校验，完整子树须无 `secret`，无效时 400 带路径；`default` 提交非空 `overrides` 时返回 400 |
| `POST /api/v1/profiles/{id}/plugins/{plugin_id}/prune` | `{"removed":["unit"]}` | 只删除列为无效的覆盖值 |
| `GET /api/v1/profiles/{id}/platform` | `{"command_prefixes":["/","!"],"reply":{"mention_sender":true,"quote":null,"prefix":null,"forward_threshold":1500}}` | 没有覆盖的键为 `null`；`default` 返回完整值 |
| `PUT /api/v1/profiles/{id}/platform` | 同形对象 → `{}` | 整体替换，省略或 `null` 跟随 `default`；`default` 要求全部键非 `null`，写回 `core_config.json`。前缀数组非空，转发阈值须为非负整数 |
| `GET /api/v1/profiles/{id}/tools` | `{"items":[{"tool":"weather/subscribe","policy":"allow"},{"operation":{"plugin_id":"agent","name":"mcp/fs/write_file"},"policy":"deny"}]}` | 只列出覆盖了全局策略的工具与操作 |
| `PUT /api/v1/profiles/{id}/tools` | `{"items":[…]}`，policy 取 allow、confirm、deny 或 inherit → `{}` | 仅更新所列项；每项恰好指定 `tool` 或 `operation`，`inherit` 删除该项覆盖 |

`invalid` 列出失效覆盖的路径与原因（`unmarked`、`type`、`schema unavailable`）。已有复合字段标记仍可读取；覆盖路径与秘密路径相交或 Active schema 不可用时，普通 GET 在 `overrides` 中省略原值，`invalid` 与 `prune` 仍可定位该项。写入或复制使某插件的有效覆盖集合超过 4 MiB 时，按 §5 返回 HTTP 500 与 `RESOURCE_EXHAUSTED`。

§4 与 §14 的实例删除入口共用指定与运行选择的清理流程。实例删除、旧实例组替换与 Adapter 全量清除在设置提交后继续清理，即使请求已取消。提交前失败保留指定与运行选择；若设置已发布但目录同步报错，继续清理并返回同步错误及后续错误。再次删除已消失的实例仍清理该实例的遗留指定，清理成功后返回 404。

## 24. 插件休眠策略与运行状态

模式、准入、活动引用与停止规则见 CONTRACTS §4.4、§4.5，设置存储见 §15.1。以下接口沿用 §1 的管理认证、CSRF 与 §5 的错误格式；读取只查询状态，不唤醒插件。CLI `alonabot sleep-policy [ID] [never|auto|时长]` 与 WebUI 使用同一管理服务。

### 24.1 全局设置与插件覆盖

| Method / path | 请求或成功响应 | 操作语义 |
| --- | --- | --- |
| `GET /api/v1/core/sleep` | `{"mode":"always_on","idle_timeout":"15m","plugins":{}}` | 返回 `core_config.json` 中的生效设置；plugins 的值为 `"never"` 或时长，未覆盖的 ID 省略 |
| `PUT /api/v1/core/sleep` | `{"mode":"component_idle","idle_timeout":"15m"}` → `{}` | 两项必填，只修改全局模式和空闲时长，保留插件覆盖 |
| `GET /api/v1/plugins/{id}/sleep-policy` | 下文的策略对象 | 返回配置覆盖、有效准入及常驻原因 |
| `PUT /api/v1/plugins/{id}/sleep-policy` | `{"policy":"never"}` 或 `{"policy":"15m"}` → `{}` | 设置一个插件的覆盖；时长须为正的 Go duration |
| `DELETE /api/v1/plugins/{id}/sleep-policy` | 无正文 → `{}` | 删除覆盖，恢复准入规则及全局时长；CLI 的 `auto` 使用此接口 |
| `GET /api/v1/plugins/{id}/runtime` | 下文的运行对象 | 返回运行状态、代次与最近变化原因 |

策略对象为 `{"policy":null,"effective_policy":"never","sleep_capable":false,"eligible":false,"reason":"manifest_requires_resident"}`。`policy` 为文件中的覆盖值，无覆盖时为 `null`；`effective_policy` 为 `"never"` 或生效时长；`eligible` 表示当前模式、manifest、进程形式与路由策略是否允许自动休眠，实际空闲退出仍须满足活动条件。`reason` 使用稳定值：`always_on`、`manifest_requires_resident`、`in_process`、`policy_never`、`route_requires_resident`，允许休眠时为空串，按前述顺序选择首个成立原因。未声明 Core 过滤的消息事件、Raw、HTTP、Hook 与持续监听被显式时长放行时另返回 `warnings` 字符串数组；无警告时省略。面向用户的原因与提示由 WebUI i18n 翻译。

所有插件接口都只接受已安装且未卸载的 `role: plugin` 或 `role: agent` ID；ID 不存在、已卸载或属于其他角色返回 404 / `NOT_FOUND`。无 manifest 可休眠声明时，时长覆盖仍可保存，生效策略保持 `never`，WebUI 锁定时长选项并显示原因。未知模式、`auto` 字面量作为存储值、零/负/非法时长在写入前返回 400 / `INVALID_ARGUMENT`。

写入经 `settings.File.Update` 保存 `sleep`，保留其他字段及未修改的 `${...}` 原文，再向生命周期所有者应用生效策略。变更立即生效，无需 reload。`never` 立即唤醒已启用的 `SUSPENDED` 插件；切回 `always_on` 唤醒全部已启用插件，已停用插件保持停用。时长变更从下一次空闲判定使用。文件保存失败时返回实际错误；保存成功后唤醒失败时保留已保存设置，仅按 §5 返回实际错误，调用方重新 GET 核对设置；各失败插件的运行状态同时记录实际错误。

### 24.2 运行状态

运行对象为 `{"enabled":true,"generation_id":"…","state":"SUSPENDED","reason":"idle","changed_at":"2026-11-17T12:00:00Z"}`。`state` 取 CONTRACTS §4.4 的六个值，`changed_at` 为最近状态变化的 UTC RFC 3339 时间；`reason` 为无 secret 的诊断字符串（程序输出用英文），失败保留实际错误原因。最近一次唤醒与空闲停止另在存在记录时返回 `last_wake:{"reason":"…","at":"…"}` 与 `last_suspend:{"reason":"idle","at":"…"}`，缺失时省略。

`last_wake` 记录 WAKING→ACTIVE 的实际完成时间与原因；QUIESCING→ACTIVE 的空闲停止撤销保留原记录。`last_suspend` 在确认旧进程树退出、写入 SUSPENDED 时更新。S-02.1 已提供底层记录，本文休眠管理端点由 S-06.1 接入；ACTIVATED 后的业务交付仍须等待 Core 完成全部绑定。

组件列表继续保留 §3 的 `State` 字段及现有角色的语义；plugin 与 agent 追加 `RuntimeState`，取同一持久化记录中的六态值。休眠插件的 `State` 显示 `Suspended`，与 `Degraded` 区分；低层 Preparing、Activating、Activated 等阶段仍通过原字段展示。所有读取均排除快照中的私有认证材料与配置。
