# goal 语句：WebUI

本文件收录「WebUI」章节的 goal 语句，切片表在[WebUI 实施计划](../../spec/plan/webui.md)。本章节只有 `web/admin` 的前端；页面调用的 Admin API 与 WebUI 的提供在[核心章节](core.md)（goal-66、goal-84、goal-92、goal-93）。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-34 | T-08.7 | web/admin 骨架、配置表单组件与界面文案 | T-08.1、T-08.10 | 待派发 |
| goal-43 | T-09.8 | Admin 管理页 | T-08.7、T-09.6、T-09.12、T-09.13 | 等待前置 |
| goal-44 | T-09.9 | Admin 配置、权限、工具策略与插件页 | T-09.8、T-09.14 | 等待前置 |
| goal-94 | T-09.15 | 日志页 | T-09.8、T-09.10 | 等待前置 |

## 第 8 波

### goal-34 · T-08.7 web/admin 骨架、配置表单组件与界面文案

状态：待派发。前置 T-08.1（PR #54，合并提交 `021e972`）与 T-08.10（PR #77，合并提交 `f34b1e8`）均已合并。2026-11-04 按 D-76 改写：配置改由手写的 `config_schema.json` 声明（CONTRACTS §5.1），表单直接读取这种格式，fixture 按契约手写，不再用生成器从 Go 类型生成。SDK 侧的解析与校验由 T-08.11（goal-95）实现，本切片只依赖契约，不等待它；goal-95 会用自己的解析器检查本切片的 fixture。

改写时沿用的已核对事实：
- 管理读取经 `secret.Store.Redact` 把每个已设置的 secret 原位替换为 `{"$secret_set": true}`，数组元素、模板对象与 dict 值中的 secret 也是如此。
- Merge Patch 整体替换数组，而掩码不能作为保持不变的协议值（CONTRACTS §5.1、§5.3），所以修改含 secret 的 list 与 template_list 时必须重新输入其中的全部 secret；Core 遇到相同的值会复用原引用。object、template 与 dict 是 JSON 对象，按键递归合并，不受此限。
- template 的值只能含一个模板键，切换模板时补丁必须把旧模板键置为 `null`，合并结果才只剩新模板。

```text
/goal 使用 $alonabot-execute 推进 goal-34（切片 T-08.7：web/admin 骨架、配置表单组件与界面文案）。goal 目录 goal-34，分支 goal-34，worktree .worktrees/goal-34，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立 web/admin 前端工程（Vite、TypeScript、React 19、vitest）。它是官方 WebUI 组件（`role: webui`）的源码，构建产物 dist/ 是组件的静态根目录。提供由 config_schema.json 驱动的表单组件：
- 字段按 config_schema.json 的键顺序展示；invisible 字段不显示，也不会被删除。
- 支持 string、text、int、float、bool、object、list、dict、template、template_list 十种类型。
- secret 字段显示「已设置」状态，不把掩码当成保存值。
- 提交时生成 RFC 7396 Merge Patch。
- 字段与模板的文案按 i18 回退规则显示。
WebUI 自身的文案有 en-US 与 zh-CN 两个目录，并提供语言切换组件。真实的管理页面与 API 在 T-09 接入。

依据：PLAN T-08 步骤 9（前端）；CONTRACTS §5.1、§5.3、§5.4 最后一条、§7.5、§9.1（WebUI 字段）；DECISIONS D-43、D-62、D-63、D-76；docs/development.md（Node 版本与已批准的前端依赖）；TEST-MATRIX V-25 的表单分项、V-54 的界面组件分项。

范围（只新增或修改这些）：
- 新增 web/admin/：package.json、package-lock.json、tsconfig.json、vite.config.ts、index.html、src/、fixtures/、i18n/，以及 manifest plugin.yaml（`role: webui`、`assets: dist`、Admin API 主版本范围 1～1）
- .gitignore：追加 web/admin/node_modules/ 与 web/admin/dist/
- docs/development.md 不改。Node 使用 asdf 管理的 26.5.1，包管理用 npm。
允许的依赖：只用 react、react-dom、typescript、vite、@vitejs/plugin-react、vitest（都属于已批准的 Vite、TypeScript、React 19、vitest）。需要 DOM 测试环境（jsdom、happy-dom）或 Testing Library 时标阻塞，等待维护者批准。

接口（模块与导出，照此实现）：
- src/schema/model.ts
  export type FieldType = 'string'|'text'|'int'|'float'|'bool'|'object'|'list'|'dict'|'template'|'template_list'
  export type Field = { key: string; schemaPath: string[]; type: FieldType; description?: string; hint?: string; obviousHint: boolean; invisible: boolean; secret: boolean; required: boolean; defaultValue?: unknown; options?: string[]; labels?: string[]; editorMode: boolean; editorLanguage: string; slider?: { min: number; max: number; step?: number }; condition?: Record<string, unknown>; children?: Field[]; templates?: Template[] }
  export type Template = { key: string; schemaPath: string[]; name?: string; hint?: string; displayItem?: string; hideHintInList: boolean; children: Field[] }
  // schemaPath 是该节点在 config_schema.json 中的路径，含 items 与 templates 两层，例如 ["music","templates","netease","items","cookie"]；i18 覆盖按同一路径查找
  // description、hint、labels、name 是英文源文案；editorLanguage 缺省为 "json"
  // children：object 的 items、有 items 的 list 的元素字段、模板的 items；templates：template 与 template_list 按 JSON 键顺序
  export function fieldsFromSchema(schema: unknown): Field[] // 按 JSON 键顺序；invisible 字段保留在模型中并标记；遇到 CONTRACTS §5.1 没有的类型或键时抛错
- src/schema/patch.ts
  export function isSecretSet(value: unknown): boolean // value 恰为 {"$secret_set": true}
  export function buildPatch(original: unknown, edited: unknown, fields: Field[]): Record<string, unknown>
  // RFC 7396：只包含改变的键；仍为 {"$secret_set": true} 的 secret 与 invisible 字段不出现；恢复默认生成 null（删除键，由 SDK 按默认值补齐）；数组整体替换；切换模板时旧模板键为 null；结果中任何位置都不能出现 {"$secret_set": true}
  // 修改含 secret 的 list 或 template_list 而其中仍有 {"$secret_set": true} 时抛出 MissingSecretsError，列出缺少的值路径
- src/i18n/catalog.ts
  export function pickLocale(available: string[], preferred: readonly string[]): string // 精确匹配 → 主语言（多个时取字典序最小）→ en-US
  export function fieldText(locale: Record<string, unknown>, field: Field): { description: string; hint?: string; labels?: string[] }
  export function templateText(locale: Record<string, unknown>, template: Template): { name: string; hint?: string }
  // 按 schemaPath 在 locale.config 中查找覆盖 → 源文案 → 字段键（模板为模板键）；hint 与 labels 没有覆盖也没有源文案时不显示；locale 是 pickLocale 选出的插件语言文件
- src/components/ConfigForm.tsx：export function ConfigForm(props: { schema: unknown; value: unknown; messages: Record<string, unknown>; onSubmit: (patch: Record<string, unknown>) => void })
  // value 是管理读取的配置（已补齐缺失值，secret 已原位替换为 {"$secret_set": true}）；messages 是插件语言文件
- src/components/LanguageSwitch.tsx：export function LanguageSwitch(props: { value: string; onChange: (tag: string) => void })
- i18n/en-US.json、i18n/zh-CN.json：WebUI 文案，两个文件的键集合一致

行为：
- 控件：
  - string：单行文本框；有 options 时为下拉框，显示 labels（缺少的显示存储值），提交 options 中的存储值。
  - text：多行文本框。string 与 text 的 editor_mode 为 true 时用代码编辑区（等宽多行，editorLanguage 只作标注，不引入编辑器依赖）。
  - int、float：数字框；有 slider 时用滑块（min、max、step），并按 min、max 限制。
  - bool：开关。
  - object：分组，显示其 children。
  - list：没有 children 时每行一个文本框，可增删行；有 options 时为多选；有 children 时每行一个子表单。
  - dict：键值对行，可增删；值的类型可选 string、number、boolean 与 JSON 文本，JSON 文本不合法时阻止提交。
  - template：模板选择器（显示模板 name，下方显示 hint）与所选模板的子表单；没有选择时只显示选择器。
  - template_list：增删行，每行先选模板，再填子表单；折叠显示模板 name，有 displayItem 时附上该字段的当前值（例如 `URL: https://…`），hideHintInList 为 true 时折叠行不显示模板 hint。
  - description 作为标题，hint 作为提示，obviousHint 为 true 时醒目显示；required 的字段标记为必填。
  - 每个字段提供「恢复默认」，生成 null。
- condition：条件不满足时字段不显示，已有值不删除，也不进 patch。条件只比较同一层兄弟字段的当前值。
- 默认值：管理读取的 value 已由 SDK 补齐；表单只在提交时生成 patch，不自行补默认值。
- secret：
  - 显示「已设置/未设置」和「修改」按钮。未点「修改」时不进 patch；点了并输入后以新值进 patch；「清除」生成 null。
  - 任何掩码字符串都不放进 patch。
  - list 与 template_list 的元素中含 secret 时：修改该数组（增删行或改任一元素）的同时，必须为其中每个 {"$secret_set": true} 重新输入值，否则阻止提交并标出缺少的项。object、template 与 dict 按键递归合并，不受此限。
- template：切换模板时，patch 中旧模板键为 null、新模板键为其对象（例如 {"music": {"netease": null, "local": {"dir": "music"}}}）；新模板的字段按 defaultValue 预填。
- invisible：不渲染，patch 永远不包含它，不会因未显示而被删除。
- 未知键：表单只渲染 config_schema.json 中的字段；patch 只包含被编辑的已知字段，未知键不受影响。
- 字段文案：标题、提示、选项与模板文案经 fieldText、templateText 取得。插件语言文件的 `config` 与 config_schema.json 同形（D-76）：字段可覆盖 description、hint、labels，模板可覆盖 name、hint。options 的存储值不翻译。
- 语言切换：初始语言按 navigator.languages 用 pickLocale 在 en-US 与 zh-CN 中选择；用户选择后保存在 localStorage，与机器人回复的语言无关。
- fixtures/：按 CONTRACTS §5.1 手写下面四个 schema 与一个插件语言文件，原样保存（键顺序即字段顺序）：
  nested.config_schema.json：
    {"welcome": {"description": "Welcome message", "type": "text", "required": true, "default": "Hello"},
     "reply": {"description": "Reply", "type": "object", "items": {
       "prefix": {"description": "Prefix", "hint": "Text placed before every reply", "type": "string", "default": "> "},
       "mode": {"description": "Mode", "type": "string", "options": ["fast", "safe"], "labels": ["Fast", "Safe"], "default": "fast"}}},
     "level": {"description": "Level", "type": "int", "slider": {"min": 1, "max": 5, "step": 1}, "default": 3},
     "advanced": {"description": "Advanced", "type": "bool"},
     "template": {"description": "Template", "hint": "Raw JSON template", "obvious_hint": true, "type": "text", "editor_mode": true, "condition": {"advanced": true}}}
  secrets.config_schema.json：
    {"token": {"description": "Bot token", "type": "string", "secret": true, "required": true},
     "backup": {"description": "Backup token", "type": "string", "secret": true},
     "internal": {"type": "string", "invisible": true, "default": "v1"},
     "endpoint": {"description": "Endpoint", "hint": "Base URL of the service", "type": "string", "default": "https://example.com"}}
  collections.config_schema.json：
    {"tags": {"description": "Tags", "type": "list", "options": ["news", "sports", "tech"], "labels": ["News", "Sports", "Tech"], "default": ["news"]},
     "keywords": {"description": "Keywords", "type": "list"},
     "accounts": {"description": "Accounts", "type": "list", "items": {
       "name": {"description": "Name", "type": "string"},
       "api_key": {"description": "API key", "type": "string", "secret": true}}},
     "headers": {"description": "Headers", "type": "dict"},
     "ratio": {"description": "Ratio", "type": "float", "slider": {"min": 0, "max": 1, "step": 0.1}, "default": 0.5}}
  templates.config_schema.json：
    {"music": {"description": "Music source", "type": "template", "templates": {
       "netease": {"name": "NetEase Cloud Music", "hint": "Log in with a cookie", "items": {"cookie": {"description": "Cookie", "type": "string", "secret": true, "required": true}}},
       "local": {"name": "Local files", "items": {"dir": {"description": "Directory", "type": "string", "default": "music"}}}}},
     "sources": {"description": "Sources", "type": "template_list", "templates": {
       "rss": {"name": "RSS", "display_item": "url", "items": {"url": {"description": "Feed URL", "type": "string", "required": true}, "every_minutes": {"description": "Interval", "type": "int", "default": 30}}},
       "webhook": {"name": "Webhook", "hint": "Receive pushes", "hide_hint_in_list": true, "items": {"token": {"description": "Token", "type": "string", "secret": true}}}}}}
  i18n/zh-CN.json：
    {"config": {"reply": {"items": {"mode": {"description": "模式", "labels": ["快速", "安全"]}}},
                "accounts": {"items": {"api_key": {"description": "API 密钥"}}},
                "endpoint": {"hint": "服务的基础地址"},
                "music": {"templates": {"netease": {"name": "网易云音乐", "items": {"cookie": {"description": "Cookie"}}}}}}}
- npm scripts：dev、build、test（vitest run）、typecheck（tsc --noEmit）。

测试（按编码规范「测试提交规则」）：
- 提交：web/admin/src/schema/schema.test.ts（vitest，不依赖 DOM），用 fixtures 驱动。这是 config_schema.json 格式与表单之间的契约测试，覆盖：
  - fieldsFromSchema：nested 的字段顺序（welcome、reply、level、advanced、template）与 reply 的子字段顺序；required、invisible、options、slider、condition、editorMode；accounts 的 children；music 与 sources 的 templates 顺序、displayItem 与 hideHintInList；未知类型与未知键抛错。
  - buildPatch：invisible 不进 patch；未修改的 secret 不进 patch；修改后的 secret 进 patch；恢复默认生成 null；数组整体替换；未知键不受影响；修改含 secret 的 accounts 与 sources 时缺少重新输入的值抛出 MissingSecretsError；music 从 netease 切换到 local 生成旧键为 null 的 patch；删除 headers 的一个键生成该键为 null。
  - pickLocale、fieldText 与 templateText 的回退：当前语言的覆盖 → 源文案 → 字段键或模板键。
- 本地自测（不提交，PR 中写明结果）：npm run build、npm run typecheck；在浏览器中手动打开 dev 页面，用四个 fixture 渲染表单并切换语言（写明浏览器与版本，属于手动验证）。

完成条件：在 web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 全部成功；go build ./... 不受影响。

不做：真实管理 API、登录与页面路由（T-09.1、T-09.8）；CI 中的前端任务（另由维护者决定是否加入）；DOM 渲染测试（需要新依赖）；超出 JavaScript 安全整数范围的整数值的编辑（这类字段未修改时不进 patch，不受影响）；AstrBot 的 file 类型、_special 与 dict 的 template_schema（D-76 不支持）。
```

## 第 16 波

### goal-43 · T-09.8 Admin 管理页

状态：等待前置（T-08.7、T-09.6、T-09.12、T-09.13）。语句中 goal-34、goal-36、goal-39～goal-41、goal-84、goal-92 提供的符号（web/admin 的工程与组件、登录接口、管理服务的 HTTP API、Admin API 参考、会话取回与上传接口）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按前后端拆分改写：会话取回与 multipart 上传移到 goal-92（T-09.13）；原语句中嵌入 dist 的 `internal/core/adminui` 与 D-62「Core 不内嵌任何前端」冲突，已删去，静态文件由 goal-84（T-09.12）以 WebUI 组件的方式提供。2026-11-04 按 D-77～D-85 修订计划：Adapter 实例管理的表单按 goal-101 的 `instance_schema.json` 生成，增删改经 goal-93 写回 `core_config.json`（D-82、D-85）。前置合并后由审查者改写。

```text
/goal 使用 $alonabot-execute 推进 goal-43（切片 T-09.8：Admin 管理页）。goal 目录 goal-43，分支 goal-43，worktree .worktrees/goal-43，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在官方 WebUI 组件 web/admin 中实现真实的登录、总览状态、各角色组件的安装、启停与替换（`--source`、`--replace` 的页面入口，组件图标用 `<img>` 显示）、Adapter 实例管理、更新与恢复（rollback）页面，全部只调用 spec/ADMIN-API.md 列出的 Admin API。构建产物 dist/ 是 WebUI 组件的静态根目录，由 Core 按 T-09.12 提供，本切片不改 Go 代码。页面显示 Pending、候选失败、Degraded 与维护原因，只展示用户可操作的状态和具体错误。

依据：PLAN T-09 步骤 7（管理页）；CONTRACTS §7.3、§7.4、§7.5、§9.3、§12.1；DECISIONS D-25、D-62、D-63；TEST-MATRIX V-40、V-47 的页面分项。

范围（只新增或修改这些）：
- web/admin/src/：新增 api/（Admin API 客户端，带 CSRF 头）、pages/（Login、Overview、Packages、Instances、Updates）、路由（基于 window.location 与 history 的简单路由，不引入路由库）
- web/admin/i18n/：补充页面文案（en-US 与 zh-CN 键集合一致）
- 上述的测试文件
允许的依赖：前端仍只用已批准的 React、Vite、TypeScript、vitest；不引入路由、状态管理或 UI 库。

行为：
- 登录：POST /api/auth/login，成功后把 csrf_token 保存在内存中；刷新页面后经 goal-92 的 GET /api/auth/session 取回。登出。未登录访问任何页面跳转到登录页。
- Overview：Core 版本、监听地址（goal-92）、每个包的状态（未启用、运行、Pending、候选失败、Degraded 及原因、维护中）。
- Packages：经 goal-92 的上传接口安装（archive、release.json、.sig 三个文件）；启用、停用、重启、卸载（含 purge）、pin；需要确认时展示受影响列表并要求再次确认（API 的 409）。
- Instances：列出 Adapter 实例（不显示凭证，只显示「已设置」）、创建（表单由该 Adapter 包的配置 Schema 经 goal-34 的 ConfigForm 生成）、启停、删除。
- Updates：经 goal-92 的上传接口 update 新包、rollback 选择保留代次；显示操作结果与失败原因。
- 安全：所有修改请求带 X-Alonabot-CSRF；页面不读取或展示 token、secret 明文；组件图标只用 `<img>` 引用（D-63）。

测试（按编码规范「测试提交规则」）：
- 提交：web/admin/src/api/api.test.ts（vitest，不依赖 DOM）：用 fetch 的测试替身验证 API 客户端的 CSRF 头、会话取回、409 确认流程的数据结构、错误格式解析。
- 本地自测（不提交，PR 中写明结果）：npm run build 与 typecheck；用 goal-84 的 `alonabot.WebUI` 选项或安装构建出的 WebUI 包，在真实 Core 上用浏览器手动走一遍：登录 → 安装 → 启用 → 创建实例 → 停用 → rollback（写明浏览器与版本，属于手动验证）。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功。

不做：Admin API、会话取回与上传接口（T-09.4、T-09.13）；WebUI 静态文件的提供（T-09.12）；配置、命令权限、发信许可、工具策略、会话与插件页（T-09.9）；日志页（T-09.15）；Core 设置页（T-09.9）。
```

## 第 17 波

### goal-44 · T-09.9 Admin 配置、权限、工具策略与插件页

状态：等待前置（T-09.8、T-09.14）。语句中 goal-34、goal-43、goal-93 提供的符号（ConfigForm、LanguageSwitch、web/admin 的页面框架与 API 客户端、T-09.14 的 HTTP 接口）只写名字；合并后由审查者核对，再改为待派发。2026-11-04 按前后端拆分改写：原语句中的管理 API（internal/core/manage 的 *_api.go 与 admin_api_test.go）移到 goal-93（T-09.14）；嵌入构建产物的 `internal/core/adminui` 与 D-62 冲突，已删去。同日按 D-83、D-84 修订：发信许可改为「全部实例」或指定实例；插件页面改为在 WebUI 内用沙箱 iframe 打开，并由本切片实现 postMessage 桥，不再使用页面票据。

```text
/goal 使用 $alonabot-execute 推进 goal-44（切片 T-09.9：Admin 配置、权限、工具策略与插件页）。goal 目录 goal-44，分支 goal-44，worktree .worktrees/goal-44，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在官方 WebUI 组件 web/admin 中补齐页面：插件配置（表单与完整 JSON 编辑，显示 Pending 并经 reload 激活）、命令权限与启用、发信许可（allowed_adapter_instances）、Core 设置（Core Admin 列表、默认 locale、命令前缀、全局禁止表）、工具调用策略、Agent 会话的列出与清除、插件页面入口，以及界面语言切换。全部只调用 goal-93（T-09.14）写入 spec/ADMIN-API.md 的接口，本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §5、§2.3、§8（allowed_adapter_instances）、§13、§7.3、§5.4；DECISIONS D-45、D-62；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56 的页面分项。

范围（只新增或修改这些）：
- web/admin/src/pages/：新增 Config、Permissions、CoreSettings、Tools、Conversations；Packages 页增加「打开插件页面」入口；新增 src/pluginpage/：iframe 容器与 postMessage 桥
- web/admin/src/api/：追加 goal-93 接口的客户端函数
- web/admin/i18n/：补充文案（en-US 与 zh-CN 键集合一致）
- 上述的测试文件
允许的依赖：与 goal-43 相同，不新增前端依赖。

行为：
- 配置页：用 ConfigForm 编辑（schema 与 messages 取自 GET /plugins/{id}/config 的 schema 与 i18n，messages 是按界面语言经 pickLocale 选出的文件；只提交 patch）或切换到完整 JSON 编辑（textarea，提交完整替换），都带 base_generation；显示 Pending 与 Active 的差异（按键列出，secret 只显示是否已设置）；「激活」调用 reload；候选失败与 Degraded 显示具体原因；409 时提示配置已被他处修改并重新读取。invisible 字段只在完整 JSON 编辑中可见。
- 权限页：按命令树显示 API 返回的有效权限与启用状态（页面不自行推导），可设置插件、命令组、路由三层覆盖；显示遮蔽关系。
- 发信许可：缺省选中「全部实例」；切换为「指定实例」后列出全部 Adapter 实例供勾选，保存为 {all, instances}。
- Core 设置：Core Admin 列表、默认 locale、命令前缀 `command_prefixes` 与按 Adapter 实例的覆盖、全局禁止表。
- 工具页：列出全部工具及缺省与覆盖策略。
- 会话页：按插件列出会话摘要（键、条目数、最后时间），可清除；不显示条目内容。
- secret 明文查看：要求重新输入管理 token，经 reveal 接口取回后只在当前视图显示，不写入任何存储。
- 语言切换：沿用 goal-34 的组件，保存在 localStorage。
- 插件页面：在 WebUI 内用带 `sandbox="allow-scripts allow-forms allow-popups"` 的 iframe 打开 `/plugins/{id}/` 下插件声明的页面；postMessage 桥只接受来自该 iframe 的消息，转发方法与路径落在该插件已注册路由内的请求（带管理会话与 CSRF 头），把响应、当前语言的插件文案与界面主题交回页面；页面拿不到会话、CSRF token 与其他插件的路由。消息格式按 spec/ADMIN-API.md（D-84）。
- 所有修改请求带 CSRF 头。

测试（按编码规范「测试提交规则」）：
- 提交：web/admin 的 vitest（不依赖 DOM）：API 客户端对 goal-93 端点的请求格式、base_generation 与 409 处理。
- 本地自测（不提交，PR 中写明结果）：npm run build 与 typecheck；在真实 Core 上用浏览器手动走一遍各页面并切换语言（写明浏览器与版本，属于手动验证）。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功。

不做：管理 API（T-09.14）；日志页（T-09.15）；故障注入与验收汇总（T-09.11）。
```

### goal-94 · T-09.15 日志页

状态：等待前置（T-09.8、T-09.10）。2026-11-04 从 goal-66（T-09.10）中拆出。语句中 goal-43、goal-66 提供的符号（web/admin 的页面框架与 API 客户端、日志接口）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-94（切片 T-09.15：日志页）。goal 目录 goal-94，分支 goal-94，worktree .worktrees/goal-94，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在官方 WebUI 组件 web/admin 中实现日志页：按来源、级别、trace 与日期查看日志，今天的记录实时订阅，过去日期查询历史；在页面中修改全局与按来源的输出级别。只调用 goal-66（T-09.10）写入 spec/ADMIN-API.md 的日志接口，本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（日志）；CONTRACTS §14.3、§14.4；TEST-MATRIX V-62、V-63 的页面分项。

范围（只新增或修改这些）：
- web/admin/src/pages/：新增 Logs
- web/admin/src/api/：追加日志接口的客户端函数
- web/admin/i18n/：补充文案（en-US 与 zh-CN 键集合一致）
- 上述的测试文件
允许的依赖：与 goal-43 相同，不新增前端依赖。

行为：
- 今天：先读 GET /logs/recent，再订阅 GET /logs/stream（SSE 会先重发缓冲中的记录），页面按时间去重。过去日期：GET /logs/history。
- 来源、级别、trace 与日期的筛选只在浏览器中进行，筛选条件保存在 localStorage。
- 级别设置区域：经 GET/PUT /logs/levels 读取和修改全局与按来源的级别。

测试（按编码规范「测试提交规则」）：
- 提交：web/admin 的 vitest（不依赖 DOM）：日志接口客户端的请求格式；SSE 记录的解析与按时间去重；筛选条件的保存与恢复。
- 本地自测（不提交，PR 中写明结果）：npm run build 与 typecheck；在真实 Core 上用浏览器手动验证实时订阅、历史查询与级别修改（写明浏览器与版本，属于手动验证）。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功。

不做：日志 API、CLI 与 sdk.log.set_level（T-09.10）。
```
