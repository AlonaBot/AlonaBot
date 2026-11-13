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
  "State": "Activated"
}
```

`Role` 取 `plugin/adapter/agent/webui`；`Signed` 根据绑定发布者公钥是否非空返回；`Active` 是当前代次标识，尚无代次时为空。`State` 来自实际运行状态或安装状态，包含 `Installed`、`Stopped`、`Activated`、`Degraded`。WebUI 无进程，启用时显示 `Activated`。

同类组件遵循 Agent/WebUI 全局各一个、Adapter 每个 `adapter_id` 一个的安装限制。已存在时为 409 / `ALREADY_EXISTS`；`Replace: true` 经预检和确认后按缺省卸载语义替换，保留旧绑定与数据，旧组件已启用时启用新组件。WebUI 可执行全部管理动作，安装界面须展示来源与这一权限。

### 3.2 组件动作

全部为 POST，成功均返回 200 和 `{}`。动作共享解码结构 `id`、`confirm`、`purge`、`constraint`、`generation`、`install`，每个动作消费下表字段；`id` 为组件 ID。

| Path | 请求字段及示例 | 行为与特定失败 |
| --- | --- | --- |
| `/api/v1/packages/enable` | `{"id":"demo"}` | 校验依赖并启动；WebUI 只修改启用记录，检查 Admin API 范围；前置条件失败 412 |
| `/api/v1/packages/disable` | `{"id":"demo","confirm":false}` | 停止并保留安装/数据；依赖影响需确认 |
| `/api/v1/packages/restart` | `{"id":"demo"}` | 以当前 Active 换 Session；要求已启用且有 Active；WebUI 无进程，返回 412 |
| `/api/v1/packages/uninstall` | `{"id":"demo","purge":0,"confirm":false}` | 停止并清理选定范围，准确对象需确认；开发组件返回 412 |
| `/api/v1/packages/pin` | `{"id":"demo","constraint":"1.2.3"}` | 精确 SemVer；空串解除；格式错误 400 |
| `/api/v1/packages/reload` | `{"id":"demo"}` | 要求已激活的进程组件；当前候选执行返回 501 / `UNSUPPORTED`；WebUI 返回 412 |
| `/api/v1/packages/update` | `{"id":"demo","confirm":false,"install":{"Source":"owner/repo"}}` | `install` 使用 3.1 的 `InstallRequest` 大写字段；外层 `confirm` 覆盖 `install.Confirm`；校验身份、来源、公钥、pin、版本与依赖影响，当前候选执行返回 501 / `UNSUPPORTED` |
| `/api/v1/packages/rollback` | `{"id":"demo","generation":"<保留代次>","confirm":false}` | 要求目标代次属于该组件，校验确认后，当前候选执行返回 501 / `UNSUPPORTED` |

`purge` 为整数：0（缺省）保留数据、配置及来源绑定；1 额外清除受管数据；2 额外清除受管配置及安装绑定。保留记录重装要求匹配原来源与身份。Adapter 替换保留的实例归旧组件，新实例配置块换归属时另行确认。具体清理对象见 [CONTRACTS §9.3](CONTRACTS-v0.1.md#93-cli-的最小完整动作)。

Update 的本地候选可以提供 `install.Archive/ReleaseJSON/Signature/PublicKey`；省略候选来源和归档时，沿用安装时绑定来源。前置校验可能先返回 400、403、404、409 或 412；501 表示请求已到达当前尚未实现的候选执行入口，正式更新与回退编排由后继切片完成。

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

## 4. Core Admin 与 Adapter 实例

各端点经管理认证，JSON Content-Type 与错误规则同第 3 节。Core Admin 使用用户字符串 `<adapter_id>:<external_user_id>`，缺省所有值按原有校验处理。

| Method / path | 请求 | 成功响应 | 操作语义与特定错误 |
| --- | --- | --- | --- |
| `GET /api/v1/admins` | 无正文 | 200、排序的字符串数组，例如 `["telegram:123"]` | 空列表为 `[]` |
| `POST /api/v1/admins/add` | `{"user":"telegram:123"}` | 200、`{}` | 写回 `core_config.json` 并立即生效；重复添加保持集合；格式错 400 |
| `POST /api/v1/admins/remove` | `{"user":"telegram:123"}` | 200、`{}` | 删除用户并立即生效；格式错 400 |
| `POST /api/v1/adapter/instances/create` | `{"package_id":"adapter_package","instance_id":"bot","config":{},"confirm":false}` | 200、`{}` | 补全并校验实例配置，原样保留环境变量引用，写回设置；实例名跨包唯一，冲突 409；换配置块归属需确认；未知组件 404；组件已卸载或角色非 Adapter 为 412 |
| `POST /api/v1/adapter/instances/start` | `{"instance_id":"bot"}` | 200、`{}` | 由当前配置启动；要求 Adapter 激活，配置有效；已运行 409；未知实例 404 |
| `POST /api/v1/adapter/instances/stop` | `{"instance_id":"bot"}` | 200、`{}` | 保留配置，本次 Core 运行中持续停用，包括组件崩溃恢复；下次 Core 启动按设置启动；未运行 412 |
| `POST /api/v1/adapter/instances/delete` | `{"instance_id":"bot"}` | 200、`{}` | 停止仍运行的实例，删除配置与指向它的缺省选择；未知实例 404 |

实例响应仅包含操作结果，凭证与完整实例配置随设置保存，API 响应不返回它们。用户、消息与事件语义沿用 [MESSAGE-MODEL §1](MESSAGE-MODEL.md#1-eventactor-与-target)，平台身份须来自可信 Adapter。

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
| 409 | `ALREADY_EXISTS`，或下文确认响应 |
| 412 | `FAILED_PRECONDITION` |
| 501 | `UNSUPPORTED` |
| 500 | `INTERNAL` 及当前映射中其余错误码；保留实际 `code` |

需要显式确认时为 HTTP 409，正文为 `{"affected":["受影响对象", "..."]}`。客户端展示全部对象，在管理员同意后重发同一请求并置 `Confirm: true`（install）或 `confirm: true`（动作、实例创建）。它与 `ALREADY_EXISTS` 的 JSON 结构不同。

认证在管理路由查找前执行：未知 `/api/...` 无有效凭证返回 401；有效认证后返回 404、`application/json`、`{"code":"NOT_FOUND","safe_message":"management endpoint not found"}`。当前动作路径上的未登记方法同样落入这一管理 404。GET 列表与图标路由隐式接受 HEAD；POST 写接口按表格方法调用。

以下使用纯文本或标准路由响应，客户端按 HTTP 状态与 Content-Type 解析：

- 认证错误与登录错误见第 1/2 节，均为 `text/plain; charset=utf-8`。
- 图标缺失为 404；受控 Host 不匹配为 421；静态错误见第 6 节，均为纯文本。
- ServeMux 对子树入口缺少尾斜杠的规范路径（如 `/api/v1`）会产生标准重定向；客户端使用本文完整端点。原始路径含 `..`、重复分隔符、反斜杠或绝对路径形式时按所属入口拒绝。
- `/plugins/{id}/...` 与 `/media/...` 在 WebUI 之前分流，使用各自现有状态码与 MIME；插件路由受现有 sandbox CSP、nosniff 和凭证过滤保护。

## 6. 静态 WebUI

每次 GET 选取当前未卸载且已启用的唯一 WebUI，读取其 `assets` 根目录；`/` 对应 `index.html`。文件存在时提供原文件；不存在且最后一段无扩展名时回退根 `index.html`；带扩展名的缺失路径返回 404。目录仅提供该目录的 `index.html`，缺失返回 404，无目录列表。实际 I/O 故障、损坏的安装或状态库故障返回安全的 500 并记录错误，只有文件不存在才适用回退。

`/api`、`/plugins`、`/media`、`/admin`、`/healthz`、`/internal`、`/auth` 及其子路径由 Core 保留，在静态查找前分流或拒绝，包中同名文件也覆盖不了这些入口。文件路径须为有效的相对路径，逐级拒绝链接；静态 GET 以外的方法（含 HEAD）返回 405、`Allow: GET`。WebUI 文件响应按内容提供 MIME，未附加 WebUI 专用 CSP 或 X-Frame-Options。

没有已启用 WebUI 时为 503，英文提示包含 `use alonabot install to install a WebUI component`；Core 升级导致声明范围失配时同样为 503，并给出 Core 主版本与范围。管理、登录、插件和媒体入口继续可达。

库模式 `alonabot.WebUI(fsys fs.FS)` 以 FS 根作为本次运行的静态根，同一 `Start` 重复登记时报错。开发 WebUI 无 manifest，按审查者 D-140 决定跳过 Admin API 范围检查并在启动时记 WARN（待维护者确认）；已安装包的状态、来源和文件保持原值。

## 7. 插件页面 postMessage 桥

本节定义后继 WebUI 页面实现的消息契约。插件页面通过父 WebUI 请求自己所属组件的 HTTP 路由，并查询语言、主题和本组件文案。插件 HTTP 路由与 Listener 凭证见 [WIRE §3.3、§3.4](WIRE-FORMAT.md#33-注册描述与-http-listener)。内部代理凭证由 SDK/Core 管理，浏览器桥无需取得它。

### 7.1 消息格式

以下使用 TypeScript 对象类型表达 JSON 字段，`?` 为可省略；`id` 是请求方产生并随响应原样返回的关联字符串。所有消息的 `channel` 固定为 `alonabot.plugin`。

```ts
// HTTP 请求
{channel:"alonabot.plugin", id:string, op:"request", method:string, path:string,
 headers?:Record<string,string>, body_base64?:string}

// 语言、主题请求
{channel:"alonabot.plugin", id:string, op:"locale"|"theme"}

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

// 桥错误
{channel:"alonabot.plugin", id:string, ok:false, error:{code:string,safe_message:string}}
```

`body_base64` 使用标准 base64；请求省略时正文为空，响应保留完整字节。HTTP 响应的 `status` 保留实际状态，桥成功交付 HTTP 错误响应时也使用 `ok:true`；桥校验或执行失败使用 `ok:false` 与稳定错误码及安全英文提示。

`translate` 仅查所属组件的 `messages.key`，以 `args` 格式化命名占位符。语言按当前界面语言、该组件已有主语言回落与 en-US 规则选择；未知键返回键名，缺参数保留占位符，遵循既有 i18n 行为。`locale/theme/translate` 分别只返回相应值。

### 7.2 窗口、路径与凭证

1. WebUI 用带 `sandbox` 的 iframe 嵌入组件页面，并将 `iframe.contentWindow` 与组件 ID 绑定。父页面接收时检查 channel、消息结构及 `event.source` 是否为该窗口。不透明源的 `event.origin` 为 `"null"`，同时核对窗口身份；相同 origin 字符串无法代替窗口匹配。响应只发回绑定的正确窗口，`id` 对应原请求。
2. `path` 只接受插件相对路径。父页面拒绝绝对 URL、协议相对 URL、`..`、编码后的越界/分隔符混淆与反斜杠，校验发生在路径规范化之前。随后将目标固定到该 iframe 所属组件的 `/plugins/{id}/` 前缀，保留已校验的插件相对路径及查询参数。其他组件与 Admin API 的路径拒绝。
3. 父页面将请求限定到所属组件已注册 method/path，Core 现有代理执行最终匹配。请求使用管理员会话，变更请求由父页面添加受控 Origin 与 CSRF。父页面拒绝插件自带的 `Cookie`、`Authorization`、`X-Alonabot-CSRF`、`X-Alonabot-Proxy-Auth` 等凭证/内部身份头，比较头名时忽略大小写。
4. 父页面拒绝重定向，向插件返回响应前滤去 `Set-Cookie` 等凭证头。管理会话、CSRF 与内部代理凭证均留在父页面和 Core 之间。Core 代理继续按角色过滤请求头：私有 Adapter 路由也清除 Core 认证凭证；公开 Adapter 平台 token 路由按其独立声明处理平台凭证。
5. 插件的 HTML 与全部插件路由响应继续带 `sandbox allow-scripts allow-forms allow-popups`（无 `allow-same-origin`）和 `nosniff`，清除插件设置的 Cookie；WebUI 自身继续使用第 6 节的静态规则。桥的运行代码和浏览器交互由后继 WebUI 切片实现。
