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
| `/api/v1/packages/reload` | `{"id":"demo"}` | 要求已激活的进程组件，优先激活有效 Pending；无 Pending 时导入 config.json 再激活；WebUI 返回 412 |
| `/api/v1/packages/update` | `{"id":"demo","confirm":false,"install":{"Source":"owner/repo"}}` | `install` 使用 3.1 的 `InstallRequest` 大写字段；外层 `confirm` 覆盖 `install.Confirm`；校验身份、来源、公钥、pin、版本与依赖影响，通过候选准备与提交更新版本，保持启用状态 |
| `/api/v1/packages/rollback` | `{"id":"demo","generation":"<保留代次>","confirm":false}` | 要求目标代次属于该组件，校验确认后，当前候选执行返回 501 / `UNSUPPORTED` |

`purge` 为整数：0（缺省）保留数据、配置及来源绑定；1 额外清除受管数据；2 额外清除受管配置及安装绑定。保留记录重装要求匹配原来源与身份。Adapter 替换保留的实例归旧组件，新实例配置块换归属时另行确认。具体清理对象见 [CONTRACTS §9.3](CONTRACTS-v0.1.md#93-cli-的最小完整动作)。

Update 的本地候选可以提供 `install.Archive/ReleaseJSON/Signature/PublicKey`；省略候选来源和归档时，沿用安装时绑定来源。前置校验可能先返回 400、403、404、409 或 412；重复上传已安装的同版本同归档会因不可变包目录已存在返回 409 / `ALREADY_EXISTS`。Rollback 的正式候选编排由后继切片完成，当前返回 501 / `UNSUPPORTED`。

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

## 8. 日志读取与输出级别

`alonabot logs`、`alonabot log-level` 与 WebUI 共用下列端点，均经第 1 节的管理认证。实时缓冲按 Core、组件 ID（Adapter 按包 ID）分别保留 `LOG_BUFFER_RECORDS` 条已输出记录，缺省每来源 1000 条；重启后清空。查询筛选只影响返回内容。

| Method / path | 查询或请求 | 成功响应 |
| --- | --- | --- |
| `GET /api/v1/logs/recent` | `source`、`level`、`trace` | 200，`Record[]`，空结果为 `[]`，按记录时间升序 |
| `GET /api/v1/logs/stream` | 同 recent | 200，`text/event-stream`，先交付初始快照，再交付新记录 |
| `GET /api/v1/logs/history` | 必填 `date=YYYY-MM-DD`；其余同 recent | 200，`Record[]`，按该日本地文件的续写序号读取；该日无文件为 `[]` |
| `GET /api/v1/logs/levels` | 无正文 | 200，`{"global":"info","sources":{}}` |
| `PUT /api/v1/logs/levels` | 同形部分更新，例如 `{"sources":{"demo":"debug"}}` | 200，返回当前有效全局值与文件中的全部来源设置 |

`source` 为空时返回全部来源，实时值为 `core` 或组件 ID。`level` 表示最低记录级别，允许 `debug/info/warn/error`，省略或空串取 `debug`；`trace` 为空时返回全部追踪记录，实时查询按完整 `trace_id` 精确匹配。非法级别、非法/缺失历史日期返回 400 / `INVALID_ARGUMENT`。读取与解析历史文件的实际故障返回 500 / `INTERNAL`，并保留具体错误。

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

`time` 使用 RFC3339Nano，`level` 使用 `slog.Level` 的标准 JSON 字符串。`attrs` 保留实际存在的来源类别、平台/实例、Session/Generation、源码位置和调用字段。来源类别与组件身份由 Core 的安装记录绑定，Core 关于组件的故障记录仍归 `core`。实时结构化值保留完整 trace；stderr 采用 text 或 json 均相同。

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

历史 `source` 与筛选值来自显示标签：Core 为 `core`，插件为去掉 `alonabot_plugin_` 的显示名，Adapter 为 `adapter_id` 或 `adapter_id/instance_id`。`trace_id` 保留文件中的最多 8 位前缀；传入完整 trace 时按其前 8 位匹配。文件已省略的完整包身份、Session 和 Generation 由实时结构化 API 提供。

历史 `message` 原样保留来源标签之后的全部文本，包括级别、源码位置、消息、字段及缩进多行内容。`attrs` 包含可解析的显示来源、位置、带引号/转义的字段和多行值；普通字段保留文本值。正文与属性共享文本格式，含未闭合引号、非法转义或有歧义边界的字段候选只留在 `message`，其余可靠字段继续解析。读取、I/O 和损坏记录头等实际错误继续返回。CLI 按文件原有控制台格式显示历史记录。

### 8.3 级别写回与运行时通知

`global` 允许 `debug/info/warn/error`；`sources` 的值还允许 `off/inherit`。PUT 省略 global 或某个来源时保留当前值，来源设置为 `inherit` 时显式写入该值。空来源集合编码为 `{}`。GET 的 `global` 是本次 Core 的有效值，包含 `--log-level` 启动覆盖；`sources` 是设置文件中的覆盖项。

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

install 成功为 200、原 `Package` 导出字段；update 成功为 200、`{}`。确认响应为 409、`{"affected":[...]}`，浏览器展示全部对象，经管理员确认后，重新上传同一组文件并设置文本 `confirm=true`。JSON 客户端继续使用 install 的 `Confirm` 或 update 外层 `confirm`。

临时目录覆盖完整 Install/Update 调用；成功、400/403/409、候选失败、请求中断和文件写入失败均执行清理。关闭或清理的真实错误保留在管理错误或日志中。服务调用结束后归档源可删除；已安装不可变目录及 Journal 接管资源继续由各自所有者管理。

multipart 与 JSON Update 均进入 `manage.Service.Update`，执行绑定来源、组件身份、公钥、pin 与依赖影响检查。Core 已接通候选准备与提交：有效的新版本上传成功返回 200、`{}`，已启用的组件激活新代次，用户数据保留，受管 SQLite 在候选副本上执行迁移。缺文件或验签失败在修改旧组件前拒绝；重复上传已安装的同版本同归档返回 409 / `ALREADY_EXISTS`。

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
  "i18n": {"zh-CN":{"token":{"description":"访问凭证"}}}
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

PUT/PATCH 必须提供 `base_generation`，首次代次可提交空串；与当前 Active 不同时返回 **409 / `FAILED_PRECONDITION`**。编辑以有效 Pending 为起点，否则以 Active 为起点。新秘密在写入 Pending 前存为不可变引用；相同位置未改变的秘密复用原引用。PUT 省略的秘密路径保留；PUT/PATCH 提交数组时，仍存在的相同下标内省略的秘密保留，删除行或显式 null 沿用删除语义；PATCH 省略的字段继续保留，包含 invisible 字段。`$secret_ref` 与 `$secret_set` 属于读取标记，提交它们返回 400 / `INVALID_ARGUMENT`。导入文件缺失返回 404。

安装包保存操作只更新 Pending；两个 reload HTTP 入口共用候选编排，优先激活有效 Pending。`config/import` 总是显式导入磁盘编辑，保持 Active。CLI `alonabot reload <id>` 先调用 config/import，再调用 reload，因此直接编辑 config.json 后执行 CLI reload 继续导入文件内容。

开发来源从当前工程资源读取 schema/i18n，GET 返回固定 `dev:<id>` 代次的脱敏 Active。PUT/PATCH 使用同样的 base 和秘密省略规则，先 Complete/Validate，再以临时文件替换 `<data_dir>/plugins/<id>/config.json`，随后经 Core 的开发重启入口建立新 Session；返回脱敏 `{base_generation,config}`，保持 Pending 为 null。保存失败保留原文件；重启失败返回实际错误并保留已保存文件，调用方可修订配置后再次重启。dev 的 config/import 只读取编辑文件并返回脱敏结果，缺文件时返回脱敏 Active，CLI 随后的 reload 仍按原有路径启动 schema 缺省配置。WebUI 配置返回 400。

### 12.2 再认证查看秘密

`POST /api/v1/secrets/{plugin}/{path}/reveal` 接收 `{"token":"<当前管理 token>"}`，在既有管理认证后再次调用 `adminauth.Reauthenticate`。失败返回 403 / `PERMISSION_DENIED`；成功返回 `{"value":"<秘密明文>"}`，附加 `Cache-Control: no-store`，并以 INFO 记录组件 ID 与字段路径。

`path` 使用秘密字段的点分路径，数组下标占一段；字段名中的 `.` 编码为 `~1`，`~` 编码为 `~0`，然后将整个路径作为一个 URL 段转义。读取有效 Pending 中的引用，否则读取 Active；开发来源按 schema 确认秘密路径后读取明文。路径未指向秘密字段时返回 400。审计记录省略明文与 token。

## 13. Core 设置与发信许可

### 13.1 Core 设置

| Method / path | 请求或成功响应 | 操作语义 |
| --- | --- | --- |
| `GET /api/v1/core/admins` | `["telegram:123"]` | 返回当前生效的、排序后的用户字符串 |
| `PUT /api/v1/core/admins` | 同形字符串数组 | 整份替换，空数组清空；格式错误或展开后重复的身份返回 400 |
| `DELETE /api/v1/core/admins` | 同形字符串数组 | 删除指定用户，未列出的用户保留 |
| `GET /api/v1/core/default-locale` | `{"default_locale":"en-US"}` | 当前有效语言 |
| `PUT /api/v1/core/default-locale` | 同形对象 | 必填且展开后非空，之后新建的 Session 使用新值 |
| `GET /api/v1/core/command-prefixes` | `{"command_prefixes":["/"],"command_prefixes_by_instance":{"bot":["!"]}}` | 全局与按 Adapter 实例的覆盖；无覆盖时返回空对象 |
| `PUT /api/v1/core/command-prefixes` | 同形对象 | 两个字段必填，整体替换；前缀列表须非空，`[""]` 表示无前缀匹配 |
| `GET /api/v1/core/denies` | `{"users":["telegram:123"],"chats":["telegram:room"]}` | 全局禁止表，两个数组均返回，空列表为 `[]` |
| `PUT /api/v1/core/denies` | 同形对象 | 两个数组必填，整体替换 |
| `DELETE /api/v1/core/denies` | 同形对象 | 两个数组必填，删除指定用户与会话 |

写接口成功返回 `{}`。管理员、语言和前缀经 `settings.File.Update` 写回 `core_config.json`，保留未修改字段的 `${...}` 原文；请求中的完整环境变量引用按原文保存、按展开值校验。管理员列表写回后调用 `policy.SetAdmins`；现有 `/admins` 与 CLI `admin add/remove/list` 共用同一管理服务和设置。前缀由下一次消息派发读取，实例未配置覆盖时使用全局值。Rename 后目录同步失败时设置已发布，接口返回实际错误，同时管理员策略跟随当前设置更新。

禁止表保存在状态库。用户使用 `<adapter_id>:<external_user_id>`；会话使用不含话题的会话字符串，外部会话 ID 中的 `%`、`#` 分别编码为 `%25`、`%23`。输入包含话题或非规范编码时返回 400，检查发生在写入前。批量替换先增加新限制，再解除旧项；各项独立提交，发生存储错误时返回实际错误，已提交的限制保留，调用方可 GET 核对当前状态。

### 13.2 插件发信许可

`GET /api/v1/plugins/{id}/allowed-instances` 返回 `{"all":true,"instances":[]}`。PUT 接收同形对象并整份替换；省略 `all` 时为 true。`all:false` 将后台发信限定到 `instances`，空列表拒绝所有实例；`all:true` 允许全部实例。实例名格式错误返回 400。许可保存至共享状态库，下一次发信使用新策略。

## 14. Adapter 实例配置

| Method / path | 请求或成功响应 | 操作语义 |
| --- | --- | --- |
| `GET /api/v1/adapters/{adapter_id}/instance-schema` | `adapter_config_schema.json` 的规范 JSON 对象 | 声明通过 `configschema.Parse` 读取 |
| `GET /api/v1/adapters/{adapter_id}/instances/{instance_id}` | `{"config":{"token":{"$secret_set":true},"name":"bot"}}` | 返回实例配置，秘密字段只显示设置状态，附加 `Cache-Control: no-store` |
| `PUT /api/v1/adapters/{adapter_id}/instances/{instance_id}` | `{"config":{...}}` | 创建或整体替换指定实例配置，成功返回 `{}` |
| `DELETE /api/v1/adapters/{adapter_id}/instances/{instance_id}` | 无正文 | 停止仍运行的实例、删除配置和指向它的默认选择，成功返回 `{}` |

实例名跨 Adapter 唯一，其他 Adapter 已占用时返回 409 / `ALREADY_EXISTS`；未知 Adapter、未知实例或删除时 URL 的归属失配返回 404。配置块属于其他组件时返回 412。

提交先经 `Schema.Complete` 补全，再对展开环境变量后的副本执行 `Schema.Validate`，最后用 `File.Update` 写回原文。省略的秘密从原始文档保留，数组中仍存在的相同下标也遵循此规则。提交 `$secret_ref`、`$secret_set` 返回 400。已有实例需升级配置版本时，复用 `PrepareInstances`；问题中的 `Path`、`Reason` 保留在错误定位中。

保存配置后通过第 4 节的启停入口管理运行实例。启动前仍由 `File.PrepareInstances` 执行迁移、补全和校验，失败实例的路径与原因原样用于定位。

## 15. Agent 会话管理

`GET /api/v1/plugins/{id}/conversations` 返回按键排序的摘要数组，空结果为 `[]`：

```json
[{"key":"telegram:room@assistant","entries":3,"updated_at":"2026-11-13T12:00:00Z"}]
```

`DELETE /api/v1/plugins/{id}/conversations/{key}` 清除该组件指定会话，成功返回 `{}`。`key` 是会话存储键，Agent 按 `<会话字符串>@<实例名>` 组织；调用方把整个键按一个 URL 段转义，服务端解码一次后原样交给 `conversation.Service.Clear`。该端点在路径清理前接收键，因此键中的斜杠、冒号、反斜杠和点段保留。键须为 1～256 字节的有效 UTF-8，格式错误返回 400；未知组件返回 404。列表仅返回键、条目数与最后时间。

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

插件与 Adapter 字段的来源名和 filter 只从当前 Active Session 的 schema 取得；额外的 source/filter 查询参数忽略。字段使用 schema 的点分路径，列表字段省略下标；不存在的字段或没有 options_from 返回 400 / INVALID_ARGUMENT。Adapter 实例入口要求该 Adapter 和实例存在，归属不匹配返回 404。组件 schema、服务来源或 Agent 不可用时返回 503 / UNAVAILABLE。

界面用 `Accept-Language` 传当前管理语言，取首选语言标签；缺省使用 Core default_locale。Core 直接查询仅接受 core.chats（filter 键 chat_type、adapter_id）和 core.adapter_instances（filter 键 adapter_id）；服务来源与未知 filter 键返回 400。core.chats 可按名称或聊天 ID 搜索并按类型筛选；按值取回沿用来源的 resolve 语义。

## 18. 工具调用策略

`GET /api/v1/tools/policies` 返回 `{"policies":[{"plugin_id":"demo","service":"echo","default":"allow","override":"inherit"}]}`。目录合并当前 Attach 工具声明、本次运行出现的操作和已保存覆盖；default 为 inherit 表示当前无声明，override 为 inherit 表示没有覆盖。

`PUT /api/v1/tools/policies` 接收 `{"plugin_id":"demo","service":"echo","policy":"deny"}`，成功返回 `{}`。policy 接受 allow、confirm、deny、inherit；inherit 删除覆盖。未知策略或服务名格式错误返回 400，未知组件返回 404。写入与固定 ListTools 使用同一个 tool.Catalog 和状态库；下次查询或调用读取新策略，deny 的工具从 ListTools 隐藏，无需重启。
