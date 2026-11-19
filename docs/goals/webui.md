# goal 语句：WebUI

本文件收录「WebUI」章节的 goal 语句，切片表在[WebUI 实施计划](../../spec/plan/webui.md)。本章节只有 `web/admin` 的前端，以及 T-09.23 中官方 WebUI 的后台进程；页面调用的 Admin API 与 WebUI 的提供在[核心章节](core.md)（goal-66、goal-84、goal-92、goal-93、goal-147、goal-153～goal-155、goal-158、goal-160）。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 设计稿与共同约定

WebUI 已定稿（D-169）。各 goal 按 Figma 文件 `hV9ILivl2VGGKwtbEEh3mC` 的「v4 pages」页（[节点 225:5739](https://www.figma.com/design/hV9ILivl2VGGKwtbEEh3mC/AlonaBot-WebUI?node-id=225-5739)）实现，语句中的「画板」都指这一页上以「v4 ·」开头的画板：浅色在 y=0～4080 的几排，深色在 y=5200 一排，配置文件与模型的补充画板在 y=6300 一排，手机在 y=7400 一排。设计稿与规格冲突时以规格为准，在 PR 中指出。执行者没有 Figma 访问时，向审查者要所需画板的截图，不凭印象补画。

以下约定适用于本章节的每条 goal，语句不再重复：

- 只调用 spec/ADMIN-API.md 列出的接口；修改请求带 CSRF 头；页面不读取、不保存管理 token 与 secret 明文。
- 一级页面不放大标题，侧栏高亮已说明位置；页面顶部直接放筛选与操作。不写解释性段落，不复述用户刚做的修改。
- 新建、编辑、确认用对话框（`@radix-ui/react-dialog`）；菜单、弹层与提示用 `@radix-ui/react-dropdown-menu`、`@radix-ui/react-popover`、`@radix-ui/react-tooltip`；图标只用 `lucide-react`。插件图标一律放进带 1px 描边、固定圆角的方块里按铺满裁切，只用 `<img>` 引用（D-63）。
- 颜色、圆角、字号与字体只用 goal-166 的 CSS 变量；浅色与深色都要正确。
- 每个列表按数量很多来设计：分页（数字页码）、筛选或分组折叠。
- 界面文案放在 `web/admin/i18n/` 的 en-US 与 zh-CN 中，两个文件的键集合一致；日志级别等按惯例用英文的词保持英文。
- 宽度小于 768px 时使用 goal-166 的手机布局（底部标签栏），页面内容单列排列；手机画板只画了代表性的页面，其余按同样规则排。
- 测试：CI 不跑前端，前端纯函数的测试写在被忽略的 `*.local.test.ts`；提交的测试为无。PR 中写明 `npm run typecheck`、`npm run test`、`npm run build` 的结果，以及在真实 Core 上用浏览器手动走过的流程（浏览器与版本，属于手动验证）。
- 完成条件：在 web/admin 下 `npm ci`、`npm run typecheck`、`npm run test`、`npm run build` 全部成功；`go build ./...` 不受影响。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-34 | T-08.7 | web/admin 骨架、配置表单组件与界面文案 | T-08.1、T-08.10 | 已完成（PR #91） |
| goal-166 | T-09.26 | 品牌、视觉令牌与外壳 | T-08.7 | 已完成（PR #217） |
| goal-43 | T-09.8 | 登录、总览、插件、账号与版本 | T-08.7、T-09.6、T-09.12、T-09.13、T-09.26 | 已完成（PR #238） |
| goal-44 | T-09.9 | 插件详情、权限、会话与系统设置 | T-09.8、T-09.14、T-11.5、T-04.17、T-08.15、T-08.16 | 待派发 |
| goal-94 | T-09.15 | 日志页 | T-09.8、T-09.10、T-09.19、T-02.5 | 已完成（PR #243） |
| goal-156 | T-09.23 | 官方 WebUI 的后台进程、统计、商店、用户与数据页 | T-09.8、T-09.20、T-09.21、T-09.22、T-09.24 | 已完成（PR #245） |
| goal-146 | T-09.18 | 模型页 | T-09.8、T-09.14、T-12.16 | 已完成（PR #244） |
| goal-165 | T-09.25 | 配置文件页 | T-04.19、T-11.6、T-12.19、T-09.18 | 待派发 |
| goal-175 | S-06.3 | 休眠策略与运行状态的页面 | T-09.8、T-09.9、S-06.1 | 等待前置 |

## 第 8 波

### goal-34 · T-08.7 web/admin 骨架、配置表单组件与界面文案

状态：已完成（[PR #91](https://github.com/AlonaBot/AlonaBot/pull/91)，合并提交 `9789fc8`）。审查中维护者确认三项（PR 评论）：批准 `@types/react@19`、`@types/react-dom@19`；数组行与 secret 掩码形状的补丁规则写进 CONTRACTS §5.1；下文「测试」中提交的 `schema.test.ts` 改为被忽略的 `schema.local.test.ts`，npm 的 test 改为 `vitest run --passWithNoTests`，本切片没有提交的测试。以下为派发时的记录：前置 T-08.1（PR #54，合并提交 `021e972`）与 T-08.10（PR #77，合并提交 `f34b1e8`）均已合并。2026-11-04 按 D-76 改写：配置改由手写的 `config_schema.json` 声明（CONTRACTS §5.1），表单直接读取这种格式，fixture 按契约手写，不再用生成器从 Go 类型生成。SDK 侧的解析与校验由 T-08.11（goal-95）实现，本切片只依赖契约，不等待它；goal-95 会用自己的解析器检查本切片的 fixture。

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

## 第 9 波

### goal-166 · T-09.26 品牌、视觉令牌与外壳

状态：已完成（PR #217）。合并提交 `acd5ba0`，验证证据见 TEST-MATRIX §3.2。按 `web/admin/src/` 核对，后继页面接入以下接口：

- `routes.tsx`：导出 `RouteEntry = { id: string; path: string; label: string; icon: LucideIcon; group: 'main' | 'bottom'; render: () => JSX.Element }` 与 `routes: RouteEntry[]`。当前仅开发模式登记 `/dev/forms`（ID `dev-forms`）；`main.tsx` 用 `usePath()` 精确匹配 `path` 并把页面交给外壳渲染。
- `shell/AppShell.tsx`：`AppShell(props: { version: string; onLogout?: () => void; children: React.ReactNode }): JSX.Element`。桌面侧栏收到 `onLogout` 才在标志菜单显示登出；现有入口传入的版本是 `v0.1.0`，还未接 API。侧栏版本按钮打开「关于」。
- `shell/router.ts`：`usePath(): string` 读取 `pathname` 并跟随 `popstate`；`navigate(path: string, options?: { replace?: boolean }): void` 用 History API 切换路径；`Link(props: LinkProps): JSX.Element`，其中 `LinkProps` 为 `{ to: string; className?: string; children: React.ReactNode }`，普通左键点击走站内导航。
- `shell/search.ts`：`SearchContext: React.Context<{ register(items: SearchItem[]): () => void }>`；`SearchItem = { id: string; title: string; section: string; keywords?: string[]; run: () => void }`；`filterSearch(items: SearchItem[], query: string): SearchItem[]` 按标题与关键词做大小写不敏感的子串匹配。`AppShell` 自动用路由的本地化标题和路径建立页面搜索项；额外操作由页面注册，卸载时调用返回的注销函数。
- `theme/theme.ts`：`ThemeChoice = 'light' | 'dark' | 'system'`、`resolveTheme(choice: ThemeChoice, prefersDark: boolean): 'light' | 'dark'`、`useTheme(): [ThemeChoice, (choice: ThemeChoice) => void]`。偏好存于 `alonabot.webui.theme`，缺省跟随系统，结果写入根元素 `data-theme`，由 `theme/tokens.css` 提供两套颜色令牌。
- `shell/sidebar.ts`：`SidebarState = { order: string[]; hidden: string[]; collapsed: boolean }`、`loadSidebar(available: string[]): SidebarState`、`saveSidebar(state: SidebarState): void`。桌面侧栏按 `main` 条目的 ID 保存顺序、隐藏和折叠状态，`bottom` 放在底部；手机按 `overview`、`plugins`、`profiles`、`logs` 的顺序取已登记的 `main` 条目，其余路由放入「更多」，无剩余项时省略「更多」。
- `i18n/UI.tsx`：`UIContext = createContext(english.messages)`、`useUI(): typeof english.messages`。`RouteEntry.label` 填 `web/admin/i18n/en-US.json` 与 `zh-CN.json` 共有的 `messages` 键，供导航、顶栏标题与搜索共用。

以下为派发时的语句。

```text
/goal 使用 $alonabot-execute 推进 goal-166（切片 T-09.26：品牌、视觉令牌与外壳）。goal 目录 goal-166，分支 goal-166，worktree .worktrees/goal-166，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：在 web/admin 中建立官方 WebUI 的外壳：浅色与深色两套 CSS 变量、标志与站点图标、JetBrains Mono 字体、侧栏、顶栏全局搜索的外壳、明暗切换、按路径切换页面的路由与手机布局。页面本身由后续 goal 实现；本切片只提供外壳与路由表，未实现的路由不出现在侧栏中。

依据：PLAN T-09 步骤 7（品牌与静态资源）；DECISIONS D-63、D-169；docs/brand/；TEST-MATRIX V-146。设计稿画板：「v4 · 总览」「v4 · 插件（侧栏收起）」「v4 · 总览（侧栏拖动排序）」「v4 · 总览（侧栏悬停与右键）」「v4 · 总览（编辑侧栏）」「v4 · 插件（搜索悬停展开）」「v4 · 插件（搜索点击展开）」、y=5200 一排的深色画板、y=7400 一排的手机画板。

范围（只新增或修改这些）：
- web/admin/package.json、package-lock.json：加入 lucide-react、@dnd-kit/core、@dnd-kit/sortable、@radix-ui/react-dialog、@radix-ui/react-dropdown-menu、@radix-ui/react-tooltip，锁定安装时的最新稳定版本
- web/admin/index.html：站点图标链接与在首帧前设置主题的内联脚本
- web/admin/public/：favicon.png（256）、favicon-32.png、apple-touch-icon.png（180）、icon-192.png、icon-512.png，均由 docs/brand/avatar.png 缩放得到；PR #202 已合并时替换其中的 favicon.png
- web/admin/src/assets/：mark.svg（复制 docs/brand/mark.svg）；fonts/JetBrainsMono-Regular.woff2、fonts/JetBrainsMono-Medium.woff2 与 fonts/OFL.txt，取自 JetBrains Mono 官方 GitHub Release 2.304 的 fonts/webfonts/，PR 中写明下载地址与两个文件的 SHA-256
- web/admin/src/theme/：tokens.css、theme.ts
- web/admin/src/shell/：AppShell.tsx、Sidebar.tsx、TopBar.tsx、GlobalSearch.tsx、MobileTabBar.tsx、AboutDialog.tsx、router.ts、sidebar.ts、search.ts
- web/admin/src/routes.tsx、src/main.tsx、src/style.css
- web/admin/i18n/en-US.json、zh-CN.json：外壳文案
- 上述的 *.local.test.ts
允许的依赖：只用上面列出的六个包与已批准的 React、Vite、TypeScript、vitest；recharts 与 @radix-ui/react-popover 由使用它们的页面切片引入。

接口（模块与导出，照此实现）：
- src/theme/tokens.css：在 `:root[data-theme="light"]` 与 `:root[data-theme="dark"]` 下定义以下变量（浅色 / 深色，取自设计稿变量集「AlonaBot v2」与「AlonaBot Tokens」的 console-*）：
  --color-bg #F3F6F9 / #0D121E；--color-surface #FFFFFF / #151C2B；--color-surface-2 #F4F7FA / #1C2538；--color-line #DFE5EC / #28334A；
  --color-ink #1F2A44 / #E5EAF3；--color-ink-muted #5E6B82 / #94A1B8；--color-accent #0E7C92 / #22B8CF；--color-accent-strong #0B6A80 / #67D7E8；--color-accent-soft #E2F5F8 / #103440；
  --color-warning #D99A06 / #FACC15；--color-warning-ink #8A5A00 / #FACC15；--color-yellow #FACC15 / #FACC15；--color-success #1F9D6B / #3DD6A0；--color-danger #F4505A / #FF6B74；--color-critical #C084FC / #C084FC；--color-pink #FF6F9C / #FF7AA8；
  --color-bubble #1F2A44 / #222E46；--color-on-bubble #FFFFFF / #EEF2FA；--color-chart-2 #1F2A44 / #94A3B8；
  --color-console-bg #0E1726 / #070D18；--color-console-line #1C2A3F / #111C2E；--color-console-text #C9D5E3 / #C9D5E3；--color-console-dim #6B7C93 / #6B7C93
  两个主题共用：--radius-sm 6px、--radius 10px、--radius-lg 16px；--text-xs 12px、--text-sm 14px、--text-md 16px、--text-lg 20px、--text-xl 28px；
  --font-sans: "PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC", "Noto Sans SC", system-ui, sans-serif；--font-mono: "JetBrains Mono", ui-monospace, monospace（@font-face 指向 src/assets/fonts 的两个 woff2，font-display: swap）。
  阴影只有两档：--shadow-popover 与 --shadow-dialog。
- src/theme/theme.ts
  export type ThemeChoice = 'light' | 'dark' | 'system'
  export function resolveTheme(choice: ThemeChoice, prefersDark: boolean): 'light' | 'dark'
  export function useTheme(): [ThemeChoice, (choice: ThemeChoice) => void] // 保存在 localStorage 'alonabot.webui.theme'，缺省 system；选择 system 时跟随 prefers-color-scheme 的变化；把结果写到 document.documentElement.dataset.theme
- src/shell/router.ts
  export function usePath(): string
  export function navigate(path: string, options?: { replace?: boolean }): void // 基于 history.pushState 与 popstate，不引入路由库
  export function Link(props: { to: string; className?: string; children: React.ReactNode }): JSX.Element
- src/routes.tsx
  export type RouteEntry = { id: string; path: string; label: string /* 文案键 */; icon: LucideIcon; group: 'main' | 'bottom'; render: () => JSX.Element }
  export const routes: RouteEntry[] // 本切片只登记开发预览 `/dev/forms`（只在 import.meta.env.DEV 时存在，复用 goal-34 的表单预览）；后续页面 goal 向这里追加条目
- src/shell/sidebar.ts
  export type SidebarState = { order: string[]; hidden: string[]; collapsed: boolean }
  export function loadSidebar(available: string[]): SidebarState // 读 localStorage 'alonabot.webui.sidebar'；丢弃已不存在的 id，新出现的 id 按 available 的顺序追加在末尾
  export function saveSidebar(state: SidebarState): void
- src/shell/search.ts
  export type SearchItem = { id: string; title: string; section: string; keywords?: string[]; run: () => void }
  export function filterSearch(items: SearchItem[], query: string): SearchItem[] // 不区分大小写的子串匹配 title 与 keywords，按 section 分组保持输入顺序
  export const SearchContext: React.Context<{ register(items: SearchItem[]): () => void }>
- src/shell/AppShell.tsx：export function AppShell(props: { version: string; onLogout?: () => void; children: React.ReactNode }): JSX.Element

行为：
- 侧栏（宽度 220px，收起时 64px 只显示图标、悬停出 tooltip）：左上角为 mark.svg 与「AlonaBot」字样，点击打开菜单（编辑侧栏、退出登录；没有 onLogout 时不显示退出）；收起与展开的把手在侧栏右边线上。入口来自 routes 中 group 为 main 的条目，按 SidebarState.order 排列，可以用 @dnd-kit 拖动排序（支持键盘）；右键菜单只有「从侧栏隐藏」与「编辑侧栏」，没有上移、下移；被隐藏的入口收进底部的「更多」并能从那里找回；「编辑侧栏」对话框列出全部入口，可勾选显示与拖动排序。底部从上到下是系统设置（group 为 bottom 的条目）与版本号；点版本号打开「关于」对话框（版本、项目链接、许可），版本号不是更新按钮。侧栏不放头像或用户名，不放搜索框。
- 顶栏：内容区顶部一行的右端是全局搜索，平时收成图标按钮；悬停或点击时在同一行的自动布局中展开为输入框，挤占左侧空白，不遮住页面内容。⌘K（Windows 与 Linux 为 Ctrl+K）聚焦搜索。结果来自各页面经 SearchContext 登记的条目，按 section 分组，Enter 执行 run 并收起。
- 明暗：useTheme 的选择在系统设置页由后续 goal 提供控件；本切片在「关于」对话框旁不放控件，只保证 data-theme 在首帧前设置（index.html 内联脚本读同一个 localStorage 键），切换时不闪白。
- 手机（宽度小于 768px）：侧栏隐藏，顶栏为标志、页名与搜索按钮；底部标签栏为总览、插件、配置文件、日志与「更多」（其余入口在「更多」中），标签只显示 routes 中已登记的条目，不足四个时按实际数量排。
- 站点图标：index.html 引用 favicon.png、favicon-32.png、apple-touch-icon.png，图标背景透明。
- 构建产物不请求任何外部字体、图标或脚本。

测试（按编码规范「测试提交规则」）：
- 提交：无（CI 不跑前端）。
- 本地自测（不提交，PR 中写明结果）：theme.local.test.ts（resolveTheme 的六种组合）；sidebar.local.test.ts（新 id 追加、删除的 id 丢弃、隐藏项保留）；search.local.test.ts（大小写、keywords、分组顺序）；在浏览器中手动验证浅色、深色与跟随系统切换后刷新保持且不闪白，侧栏收起、拖动排序、隐藏后从「更多」找回、编辑侧栏，⌘K 搜索，宽度 390px 的手机布局，浅色与深色标签栏上的站点图标（写明浏览器与版本）；`grep -rE "https?://" web/admin/dist` 中没有外部字体或脚本地址。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功；dist 中有 mark.svg、各尺寸站点图标、两个 woff2 与 OFL.txt。

不做：登录、各页面与 Admin API 的调用（goal-43 及之后）；系统设置页中的明暗与语言控件（goal-44）；插件页面的 `--ab-*` 令牌映射（goal-150 的 sdk/page 套件）；CJK 字体的下载（D-169 用系统字体栈）。
```

## 第 17 波

### goal-43 · T-09.8 登录、总览、插件、账号与版本

状态：已完成（PR #238），合并提交 `9123839`。以下保留派发时语句。派发时：按 `spec/plan/webui.md`、`spec/plan/core.md` 与 main 核对：T-08.7（PR #91，`9789fc8`）提供 `ConfigForm` 和双语文案；T-09.6（PR #179，`8898251`）接通更新的候选与提交；T-09.12（PR #167，`3a9b4af`）提供静态单页回退与组件图标；T-09.13（PR #181，`d507e5a`）提供会话取回、multipart 安装/更新及 Core 总览；T-09.26（PR #217，`acd5ba0`）提供路由、外壳、搜索和主题，签名见 goal-166。

main 上 T-09.7（PR #187）已接通回退执行，ADMIN-API §3.2 的 501 说明早于该实现；T-09.10（PR #177）提供日志接口，T-09.14（PR #192）提供 Adapter 实例 Schema 与脱敏配置读取。§20 的保留代次与实例列表尚未注册，归 T-09.22（goal-155）；会话数在该实例响应中也无字段。组件列表暂无展示名称、描述、故障原因字段；总览统计和新版来源归 goal-156（D-162）。下文先接可用 API，缺数据的部分保留设计稿规定的入口或明确显示缺口。D-172 已定稿，`web/admin/plugin.yaml` 仍是旧 ID，实际改名归 goal-177。

```text
/goal 使用 $alonabot-execute 推进 goal-43（切片 T-09.8：登录、总览、插件、账号与版本）。goal 目录 goal-43，分支 goal-43，worktree .worktrees/goal-43，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：接通真实登录、总览、插件管理（安装、替换、启停、更新、卸载及详情）与机器人账号管理，建立同源 Admin API 客户端。按 v4 保留版本记录、账号列表与 Core 新版提示的布局，随对应数据接口接通。

依据：PLAN T-09 步骤 7（管理页）及其停止和失败条件；spec/plan/webui.md §3；CONTRACTS §7.3～§7.5、§9.3、§12.1；ADMIN-API §1～§6、§8～§11、§14、§20；DECISIONS D-25、D-62、D-63、D-127、D-139、D-158、D-162、D-164、D-169、D-172；TEST-MATRIX V-40、V-47 的页面分项。画板：「v4 · 登录」「v4 · 总览」「v4 · 插件」「v4 · 插件（侧栏收起）」「v4 · 机器人账号」「v4 · 机器人账号（添加）」「v4 · 插件 · 天气 · 版本记录」「v4 · 插件 · 天气 · 回退确认」「v4 · 插件 · 天气 · 高级（卸载）」「v4 · 会话（Core 有新版本）」及对应浅色、深色和手机画板；会话画板只用于共用的新版提示。

范围（只新增或修改这些；路径相对 web/admin/）：
- src/main.tsx、src/routes.tsx、src/style.css、i18n/en-US.json、i18n/zh-CN.json。
- src/shell/AppShell.tsx、Sidebar.tsx、TopBar.tsx、MobileTabBar.tsx：版本提示、登出与插件详情的导航状态。
- src/pages/LoginPage.tsx、OverviewPage.tsx、PluginsPage.tsx、PluginDetailPage.tsx、AccountsPage.tsx，以及 src/pages/ 下由这些页面共用的卡片、对话框与样式。
- src/api/admin.ts、对应的 *.local.test.ts；package.json、package-lock.json 只在本切片实际首次使用已批准依赖时修改。

允许的依赖：只用 docs/development.md 已批准的前端依赖。按 D-169，recharts、@radix-ui/react-popover 由首次实际使用它的页面切片引入并锁定版本；空图表区不引入 recharts，已有 Radix 菜单和对话框能完成的交互不引入 popover。

接口（新增的前端导出与当前可调用的 HTTP 接口）：
- src/pages/ 下各同名模块导出 LoginPage(props: { onAuthenticated: () => void }): JSX.Element、OverviewPage(props: { core: CoreInfo }): JSX.Element、PluginsPage(): JSX.Element、PluginDetailPage(props: { id: string; generations?: readonly Generation[] }): JSX.Element、AccountsPage(): JSX.Element。routes 登记 overview → /overview、plugins → /components、accounts → /accounts，均为 main，label 用双语 messages 键。main.tsx 单独分派 /login 和 /components/<编码后的组件 ID>，两者不作为一级导航或重复搜索结果；已登录访问 / 跳至 /overview，详情仍高亮「插件」。按 ADMIN-API §6 避开 Core 保留的 /plugins 路径；仅开发模式保留 /dev/forms 预览。
- AppShell 保留原 props，增加可选 coreUpdate: { version: string; releaseURL: string }，转给 Sidebar 在当前版本旁显示新版入口；由 goal-156 接入真实数据，本切片默认不传该值。TopBar 复用 onLogout，让手机也可从标志菜单登出。
- src/api/admin.ts：导出 AdminAPIError（status: number、code?: string、affected?: string[]，message 保留 safe_message 或纯文本错误）。同源 Cookie 由浏览器管理；登录按 §2 以受控 Origin 提交 JSON，其余受保护的变更请求携带内存中的 X-Alonabot-CSRF。登录与认证失败按 §1、§2、§9 的纯文本解析，Admin JSON 错误按 §5 保留 code 与 safe_message；仅将带 affected 数组的 409 作为确认。每个操作按契约重送原请求，须由用户明确确认。
- login(token: string): Promise<void> → POST /api/auth/login，JSON {token}，200 取 csrf_token 存入内存；restoreSession(): Promise<{ expires_at_unix_ms: number }> → GET /api/auth/session，恢复 csrf_token 与固定过期时间；logout(): Promise<void> → POST /api/auth/logout，204 清空会话。入口用 onSessionExpired(handler: () => void): () => void 订阅受保护请求的 401，卸载时注销；login 的 401 留在登录页显示错误。
- getCore(): Promise<CoreInfo> → GET /api/v1/core，CoreInfo 为 { version: string; listen: string; admin_api_major: number }。listPackages(): Promise<Package[]> → GET /api/v1/packages，Package 按 §3.1 的大小写：ID、Role、AdapterID、Version、Source、PublisherKeyID、Pin、Enabled、Uninstalled、Signed、Active、State；安装响应可有 Warnings: string[]。packageIconURL(id: string): string 返回 /api/v1/packages/{id}/icon，供 img 使用，404 显示通用图标。
- installPackage(input: { Source: string; Replace?: boolean; Confirm?: boolean } | FormData): Promise<Package> → POST /api/v1/packages/install。远程安装传 Source: "owner/repo"；浏览器上传用 archive、release、可选 signature（文件）与 public_key（release.pub 的 base64 文本），以及可选 replace、confirm（文本 true/false）。保留签名文件原始字节，由浏览器设置 multipart boundary。遇到同类冲突时，管理员可选择发出带 Replace/replace 的请求；服务端返回受影响对象后，另行确认再提交。安装界面展示来源，并提示 WebUI 组件拥有完整管理权限。
- updatePackage(id: string, input: { confirm?: boolean } | FormData): Promise<{ warnings?: string[] }> → POST /api/v1/packages/update。JSON 发 {id, confirm, install:{}} 沿用原绑定来源；multipart 同 §10，追加 id，带 confirm，无 replace。显示服务端返回的 warnings。
- changePackage(action: 'enable' | 'disable' | 'restart' | 'reload', id: string, confirm?: boolean): Promise<void> → POST /api/v1/packages/{action}，发 {id}，仅 disable 附 confirm。uninstallPackage(id: string, purge: 0 | 1 | 2, confirm?: boolean): Promise<void> → POST /api/v1/packages/uninstall，发 {id,purge,confirm}。rollbackPackage(id: string, generation: string, confirm?: boolean): Promise<void> → POST /api/v1/packages/rollback，发 {id,generation,confirm}。这些动作成功均为 200 {}。Generation = { id: string; version: string; created_at: string; active: boolean; has_db: boolean } 对应 §20 的 generations；当前只在详情收到已知的保留代次时提供回退，等 goal-155 注册 GET /api/v1/plugins/{id}/generations 后再接列表。
- getLogLevels(): Promise<LogLevels> 与 setPackageLogLevel(id: string, level: SourceLogLevel): Promise<LogLevels> → GET/PUT /api/v1/logs/levels；LogLevels = { global: string; sources: Record<string, SourceLogLevel> }；PUT 仅发 {sources:{[id]:level}}。SourceLogLevel 取当前支持的 debug、info、warn、error、off、inherit；未设置的已安装组件显示继承。critical 的写入随 T-02.5 接入。
- getInstanceSchema(adapterID: string): Promise<unknown> → GET /api/v1/adapters/{adapter_id}/instance-schema，返回 Schema 对象。getInstanceConfig(adapterID: string, instanceID: string): Promise<{ config: Record<string, unknown> }> → GET /api/v1/adapters/{adapter_id}/instances/{instance_id}，凭证形如 {$secret_set:true}。createInstance(input: { package_id: string; instance_id: string; config: Record<string, unknown>; confirm?: boolean }): Promise<void> → POST /api/v1/adapter/instances/create。changeInstance(action: 'start' | 'stop' | 'delete', instanceID: string): Promise<void> → POST /api/v1/adapter/instances/{action}，发 {instance_id}；写操作均为 200 {}。§20 的 GET /api/v1/adapter/instances 将返回 {instances:[{instance_id,adapter_id,package_id,state,safe_error,default}]}，待 goal-155 注册后接持久列表。

行为：
- 登录：按画板提交管理 token，成功清空输入并进入 /overview，失败在表单显示原因。入口先取回会话再渲染管理页；刷新已登录页面保留路径，未登录的管理路径跳至 /login。登出成功或受保护请求收到 401 后清空内存中的 CSRF 并返回登录；认证 500/503 显示真实错误。token 和新填凭证只随对应请求提交，提交成功后清空输入，浏览器存储和日志不留明文。
- 总览：显示真实 Core 版本、监听地址与当前可从 Package.State 确认的故障概况；「查看日志」在 logs 路由已登记时可用。保留 v4 的状态行、时间范围（24 小时、7 天、30 天）和四组图表位置（消息量、各账号、热门插件、错误与警告；WARNING 黄色、CRITICAL 紫色）。goal-156 接入统计前显示无统计数据，不用日志缓冲推算完整统计。
- 插件列表：按类型（功能、适配器、Agent、界面）、名称筛选，支持「有问题的在前」和数字分页。「已安装」只列 Uninstalled 为 false 的组件；「商店」只保留入口。卡片展示图标、名称、版本、状态和 Enabled 开关；当前用 ID 作名称，缺少的描述或原因不编造，把真实错误展示在「查看原因」或动作反馈处。⋯ 菜单包含启用、停用、重启与卸载，按角色与状态提供动作。支持上传和远程安装、同类替换、沿原来源更新和上传更新。根据 §5/§10 逐项展示服务端返回的确认对象；卸载选项为只卸载、清理数据、清理数据与配置，分别对应 purge 0/1/2。成功后重读列表，失败显示原原因及安装、更新的真实结果。
- 插件详情：标题行有「插件 ›」面包屑、图标、名称、状态，右侧有重载和 ⋯。只做「版本记录」「高级」两个标签。「版本记录」显示当前版本和当前代次；收到保留代次时按服务端顺序列出，非 Active 项可回退，确认对话框显示目标和将丢弃的写入（以 affected 为准）。代次列表尚不可用时明确显示，并省略回退操作。高级提供该组件的日志级别和同一卸载流程；休眠控件见 goal-175。
- 机器人账号：按 v4 显示平台图标、实例名、状态、会话数；实例与会话数来源未就绪时明确标注，不填零。添加先选已安装 Adapter（区分 AdapterID 与包 ID），再用 src/components/ConfigForm.tsx 的 ConfigForm({schema, value, messages, onSubmit})，传返回的 Schema、空配置对象及可用文案，onSubmit 收到的新建 patch 作为 config 发送。成功后可对刚创建、ID 已确认的实例执行启停、删除和脱敏读取；已有实例列表接通后复用同一操作。凭证只按 $secret_set 标「已设置」，不把掩码回传。实例停用只影响本次 Core 运行（D-139）。
- 版本：外壳及「关于」使用 getCore 的 version；已有目标版本时显示「当前版本 → 目标版本」和带图标的操作。Core 有已知新版时在侧栏版本号旁加小按钮、同排显示，按画板提供提示并指向发行说明；版本号本身仍打开「关于」。实际新版检查由 goal-156 接入，本切片不直连 GitHub，Core 无浏览器内升级动作。
- 已登记页面自动进入搜索；当前可用的操作用 SearchContext 注册，离开页面时注销。组件 ID 和 AdapterID 取 API 原值；新写的市场 ID 示例用 D-172 连字符，manifest 改名见 goal-177。

测试：
- 提交：无。
- 本地自测：前端纯函数写在 *.local.test.ts，覆盖 409 确认与冲突的区别、上传 part 名、筛选与排序、未有新版时不显示更新入口。PR 写明 npm run typecheck、npm run test、npm run build 的结果。
- 浏览器手动验证（PR 写明浏览器和版本）：同源真实 Core 上登录失败、首次登录成功、详情路径刷新恢复、登出和受保护页面；深浅色、手机导航与搜索；插件上传、远程安装、替换/清理确认和取消、启停、重启、重载、更新与高级日志级别；新建账号后按真实 ID 读取脱敏配置、启动、停止及删除。核对 409 冲突、候选失败时页面与实际状态一致。代次列表、已有账号列表与会话数、统计与新版来源未接通的流程标明未运行；本地数据只用于验证对应界面和确认流程。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功，仓库根目录 go build ./... 成功；PR 记录真实 Core 上已有接口的浏览器验证及未接通项。

不做：Go 代码或新增 Admin API；插件配置、权限、会话、系统设置与用户（goal-44、goal-156）；日志页（goal-94）；商店实现、统计采集、新版检查、更新内容与存储视图（goal-156）；模型与配置文件页（goal-146、goal-165）；休眠策略、休眠状态及原因（goal-175）。
```

## 第 19 波

### goal-44 · T-09.9 插件详情、权限、会话与系统设置

状态：待派发。前置 T-09.8（PR #238，合并提交 `9123839`）、T-09.14（PR #192，`0cff77c`）、T-11.5（PR #219，`b2348c9`）、T-04.17（PR #218，`6d81be3`）、T-08.15（PR #211，`3b2b64c`）、T-08.16（PR #239，`1a1d12c`）均已合并。按 main `e250823` 核对：`web/admin/src/pages/PluginDetailPage.tsx` 已导出 `PluginDetailPage(props: { id: string; generations?: readonly Generation[] }): JSX.Element`，已有版本记录、回退、组件日志级别与卸载；`src/api/admin.ts` 已导出 `listPackages(): Promise<Package[]>`、`listGenerations(id: string): Promise<Generation[]>`、`listInstances(): Promise<AdapterInstance[]>`、`getLogLevels(): Promise<LogLevels>`、`setPackageLogLevel(id: string, level: SourceLogLevel): Promise<LogLevels>`、`AdminAPIError`、`isConfirmationError`，配置、权限、会话及令牌轮换尚无前端导出。`src/main.tsx` 已认证后分派 `/components/{id}` 与 `matchedRoute.render()`；`src/routes.tsx`、`src/shell/`、`src/i18n/UI.tsx` 和 `src/components/ConfigForm.tsx` 可直接复用。`ConfigForm({schema,value,messages,onSubmit})` 已调用 `fieldsFromSchema(schema: unknown): Field[]` 与 `buildPatch(original: unknown, edited: unknown, fields: Field[]): Record<string, unknown>`；`src/schema/model.ts` 尚未接入 `options_from`、`example`、`profile`，仍过滤 `$version`、`$migrations`，本切片一并修正。

后台实际包 `github.com/AlonaBot/AlonaBot/internal/core/manage` 的 `(*Service).Handler(auth *adminauth.Auth) http.Handler` 已注册配置、权限、工具、会话、任务、选项与三个 prune 端点；`github.com/AlonaBot/AlonaBot/contracts/configschema` 的 `(*Schema).Correct(raw json.RawMessage) (json.RawMessage, []Correction, error)` 已产出脱敏校正记录。`github.com/AlonaBot/AlonaBot/sdk/page` 已有 `Handler() http.Handler`、kit.js 与 kit.css，父桥和浏览器联验由本切片完成。沿用 D-128～D-130、D-157、D-159～D-161、D-168、D-169 的目标与要点；插件配置文件切换归 goal-165，休眠控件归 goal-175，存储与数据页归 goal-156。§7 的 messages 读取、系统设置的部分操作及会话配置文件映射仍有下文列出的契约缺口，交审查者裁定。

```text
/goal 使用 $alonabot-execute 推进 goal-44（切片 T-09.9：插件详情、权限、会话与系统设置）。goal 目录 goal-44，分支 goal-44，worktree .worktrees/goal-44，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：补齐插件详情页的配置、插件页面与定时任务标签，权限页的命令、工具调用与 Core 管理员/禁止操作，会话页与系统设置页；复用已有版本记录、回退确认与高级标签。已有目标中的用户目录和聊天目录选择接入 goal-156 的权限「用户」页，数据库表格、存储与会话记录视图统一由 goal-156 承接；配置文件页及插件配置文件切换由 goal-165 承接。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（配置与插件页）及其停止和失败条件；spec/plan/webui.md §2、§3；CONTRACTS §2.3、§5、§7.3～§7.5、§8、§13、§14.3；ADMIN-API §3.3、§4、§7、§8.3、§12～§18、§20、§21、§23；DECISIONS D-45、D-84、D-128、D-129、D-130、D-153、D-157、D-159、D-160、D-161、D-168、D-169；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56、V-115、V-116、V-117、V-128、V-130、V-131、V-132 的页面分项。画板：「v4 · 插件 · 天气」及深色、「v4 · 插件 · 天气（切换配置文件）」「v4 · 插件 · 天气 · 高级」「v4 · 插件 · 天气 · 高级（卸载）」「v4 · 插件 · 天气 · 高级（不能改休眠）」「v4 · 插件 · 天气 · 回退确认」「v4 · 插件 · 天气 · 定时任务」「v4 · 权限 · 命令」及深色、「v4 · 权限 · 命令（编辑命令名）」「v4 · 权限 · 工具调用」「v4 · 会话」「v4 · 系统设置」「v4 · 系统设置（更换管理令牌）」。截图由审查者提供在 goal-44/figma/；切换配置文件与休眠画板用于核对相邻布局，对应控件由 goal-165、goal-175 实现。原语句引用的「插件 · 天气 · 数据」「插件 · Agent · 用户任务（插件页面）」「权限 · 用户」「权限 · 用户（添加用户）」分别随 goal-156 的数据、用户页与 Agent 插件页面接入核对。

范围（只新增或修改这些；路径相对 web/admin/）：
- src/pages/PluginDetailPage.tsx、AccountsPage.tsx；新增 src/pages/PluginConfigPanel.tsx、PluginJobsPanel.tsx、PermissionsPage.tsx、ConversationsPage.tsx、SettingsPage.tsx，以及 src/pages/ 下这些页面专用的对话框和纯函数模块。AccountsPage 只补实例配置的校正记录、选项查询与清理入口，保留添加、启停、删除流程。
- src/api/admin.ts；新增 src/components/PluginPageFrame.tsx 与 src/bridge/pluginPage.ts；src/components/ConfigForm.tsx、src/schema/model.ts、src/i18n/catalog.ts。
- src/main.tsx、src/routes.tsx；新增 src/shell/settings.ts；src/shell/AppShell.tsx、TopBar.tsx、GlobalSearch.tsx、search.ts 只接通共享偏好状态与桥的搜索动作。
- src/style.css、i18n/en-US.json、i18n/zh-CN.json；上述模块对应的 *.local.test.ts。

允许的依赖：只用 web/admin/package.json 已安装的 React、Radix Dialog/DropdownMenu/Tooltip、lucide-react、TypeScript、Vite、vitest 与浏览器 API；复用现有 schema/patch、i18n/catalog、theme/theme、shell/router。需要新增依赖时按共同约束标阻塞。

接口（新增导出照此实现；先合并切片已有同名能力时复用，并保留其现有字段和调用语义）：
- src/pages/PluginDetailPage.tsx 保持 PluginDetailPage(props: { id: string; generations?: readonly Generation[] }): JSX.Element；新增 src/pages/PluginConfigPanel.tsx 的 export function PluginConfigPanel(props: { id: string }): JSX.Element、src/pages/PluginJobsPanel.tsx 的 export function PluginJobsPanel(props: { id: string }): JSX.Element。详情默认配置标签，增加任务与各声明页面标签，保留版本与高级标签、现有启停/restart/reload/rollback/uninstall 和 affected 确认流程。静态 WebUI 的配置接口返回 400，按实际组件能力显示标签。
- src/pages/PermissionsPage.tsx：export function PermissionsPage(): JSX.Element，内部切换命令与工具调用；src/pages/ConversationsPage.tsx：export function ConversationsPage(): JSX.Element；src/pages/SettingsPage.tsx：export function SettingsPage(): JSX.Element。routes 分别登记 {id:'permissions',path:'/permissions',label:'permissions',icon,group:'main',render:() => <PermissionsPage />}、{id:'conversations',path:'/conversations',label:'conversations',icon,group:'main',render:() => <ConversationsPage />}、{id:'settings',path:'/settings',label:'settings',icon,group:'bottom',render:() => <SettingsPage />}。main.tsx 沿用精确匹配和认证外壳，/components/{编码后的 ID} 继续独立分派并高亮插件；保留其他切片登记的路由。权限「用户」由 goal-156 的 UsersPage 接入同一权限导航，Core 管理员与禁止操作共用下述 API。
- src/shell/settings.ts：export type UISettings = { locale: string; setLocale: (locale: string) => void; theme: ThemeChoice; setTheme: (choice: ThemeChoice) => void; logout: () => Promise<void>; tokenRotated: (token: string) => void }；export const UISettingsContext: React.Context<UISettings | null>；export function useUISettings(): UISettings。main.tsx 在已认证页面上提供当前 locale、现有语言保存回调、useTheme() 的选择与 setter、现有登出动作；tokenRotated 接收一次性新 token，清理认证状态并由根组件持有结果对话框，确认关闭后清除 token、跳转 /login。结果对话框独立于已认证页面，后续 401 和页面卸载保留其展示。AppShell 移除重复的 useTheme 调用。保留语言键 alonabot.webui.locale 与主题键 alonabot.webui.theme。useUI(): typeof english.messages 保持原签名；新增界面文案放双语 messages，字段翻译继续用 pickLocale、fieldText、templateText。
- src/api/admin.ts 配置与校正类型、请求：
  export type Correction = { path: string; reason: 'converted' | 'reset' | 'removed'; original?: unknown }
  export type ConfigDiagnostics = { corrections: Correction[]; unknown: string[] }
  export type PendingConfig = { base_generation: string; config: Record<string, unknown> }
  export type ConfigSnapshot = ConfigDiagnostics & { active_generation: string; active: Record<string, unknown>; pending: PendingConfig | null; schema: unknown; i18n: Record<string, Record<string, unknown>> }
  export function getPluginConfig(id: string): Promise<ConfigSnapshot>
  export function patchPluginConfig(id: string, baseGeneration: string, patch: Record<string, unknown>): Promise<PendingConfig>
  export function discardPluginConfig(id: string): Promise<void>
  export function prunePluginConfig(id: string): Promise<PendingConfig>
  export function getCoreConfigDiagnostics(): Promise<ConfigDiagnostics>
  export function pruneCoreConfig(): Promise<void>
  export function pruneInstanceConfig(adapterID: string, instanceID: string): Promise<void>
  export function revealPluginSecret(id: string, path: string, token: string): Promise<{ value: string }>
  // 依次为 GET /api/v1/plugins/{id}/config、PATCH .../config/pending 发 {base_generation,patch}、DELETE .../config/pending、POST .../config/prune、GET /api/v1/core/settings/corrections、POST /api/v1/core/settings/prune、POST /api/v1/adapters/{adapter_id}/instances/{instance_id}/prune，以及 POST /api/v1/secrets/{plugin}/{path}/reveal 发 {token}。prune 无正文；插件 prune 返回 Pending，其余两个返回 {}。
- getInstanceConfig(adapterID: string, instanceID: string) 的响应扩展为 Promise<ConfigDiagnostics & { config: Record<string, unknown> }>，保留既有参数与 config 字段；AccountsPage 读取并展示顶层 corrections、unknown。插件/实例 ID 各按一个 URL 段编码。reveal 的 path 由实际值路径逐段将 ~ 转成 ~0、. 转成 ~1 后用点连接，最后整段 URL 编码；数组下标占一段（§12.2）。
- src/api/admin.ts 选项类型与导出：
  export type ConfigOption = { value: string; label: string; description?: string }
  export type OptionsPage = { options: ConfigOption[]; more: boolean }
  export function listConfigOptions(id: string, path: string, q: string, locale: string): Promise<OptionsPage>
  export function resolveConfigOptions(id: string, path: string, values: string[], locale: string): Promise<OptionsPage>
  export function listInstanceOptions(adapterID: string, instanceID: string, path: string, q: string, locale: string): Promise<OptionsPage>
  export function resolveInstanceOptions(adapterID: string, instanceID: string, path: string, values: string[], locale: string): Promise<OptionsPage>
  export function listCoreOptions(source: 'core.chats' | 'core.adapter_instances', q: string, filter: Record<string, string>, locale: string): Promise<OptionsPage>
  export function resolveCoreOptions(source: 'core.chats' | 'core.adapter_instances', values: string[], filter: Record<string, string>, locale: string): Promise<OptionsPage>
  export function pluginExampleURL(id: string, image: string): string
  // 插件 GET .../config/options?path=&q=、POST .../config/options/resolve 发 {path,values}；实例用 §17 的 .../instances/{instance_id}/options 与 /options/resolve；Core 用 /api/v1/core/options/{source} 与 /resolve，filter 编成 filter[键] 查询参数。均带 Accept-Language: locale。pluginExampleURL 的 image 是 schema 中的 examples/ 相对路径，返回 /api/v1/packages/{id}/examples/ 加去掉 examples/ 后逐段编码的路径。
- src/schema/model.ts 保持 fieldsFromSchema(schema: unknown): Field[]；新增 export type OptionsFrom = { source: string; filter?: Record<string, string> }、export type ExampleSegment = { text: string } | { image: string; alt?: string }、export type ReplyExample = { input: string; reply: string | ExampleSegment[]; condition?: Record<string, unknown> }；Field 增加 optionsFrom?: OptionsFrom、examples?: ReplyExample[]、profile: boolean，解析 example 的单对象/数组后统一成 examples。按 CONTRACTS §5.1 接受并检查这三个声明，保留既有类型、字段顺序与路径检查。删除 parseFields 中 $version/$migrations 过滤及旧注释，旧保留键按普通非法字段键报告路径；保留未知描述键报错。
- src/i18n/catalog.ts 新增 export function fieldExamples(locale: Record<string, unknown>, field: Field): ReplyExample[]，沿 schemaPath 读取 example 翻译，按源示例/片段位置合并 input、text、alt，缺项回退源值，沿用源图片路径与 condition；保持现有导出。
- src/components/ConfigForm.tsx 保留四个必填 props 与 onSubmit 的 patch 语义，增加可选 props：options?: { list: (path: string, q: string) => Promise<OptionsPage>; resolve: (path: string, values: string[]) => Promise<OptionsPage> }、exampleImageURL?: (image: string) => string、commandPrefix?: string、onDiscard?: () => void、onReveal?: (path: string[]) => Promise<string>。ConfigForm(props: { schema: unknown; value: unknown; messages: Record<string, unknown>; onSubmit: (patch: Record<string, unknown>) => void; options?: { list: (path: string, q: string) => Promise<OptionsPage>; resolve: (path: string, values: string[]) => Promise<OptionsPage> }; exampleImageURL?: (image: string) => string; commandPrefix?: string; onDiscard?: () => void; onReveal?: (path: string[]) => Promise<string> }): JSX.Element。有 onDiscard 时按编辑状态显示底部「放弃 / 保存」，其余调用者沿用现有提交方式。示例随内部草稿更新；onReveal 由页面弹出 token 再认证对话框，明文只在当前字段视图内存中显示。
- src/api/admin.ts 命令权限与名称：
  export type CommandScope = { kind: 'plugin' } | { kind: 'group'; group: string[] } | { kind: 'route'; route_id: string }
  export type CommandOverride = { scope: CommandScope; permissions: string[] | null; enabled: boolean | null }
  export type InvalidCommandName = { route_id: string; name: string[] | null; reason: 'declared' | 'unregistered' }
  export type CommandPolicy = { route_id: string; enabled: boolean; permissions: string[] | null; declared_names: string[][]; effective_names: string[][]; has_names_override: boolean; invalid_names: InvalidCommandName[] }
  export type CommandReference = { plugin_id: string; route_id: string }
  export type PluginPolicies = { overrides: CommandOverride[]; commands: CommandPolicy[]; shadows: { path: string[]; winner: CommandReference; shadowed: CommandReference[] }[]; invalid_names: InvalidCommandName[] }
  export function getPluginPolicies(id: string): Promise<PluginPolicies>
  export function setPluginPolicy(id: string, override: CommandOverride): Promise<void>
  export function setCommandNames(id: string, routeID: string, names: string[][]): Promise<void>
  export function resetCommandNames(id: string, routeID: string): Promise<void>
  // GET/PUT /api/v1/plugins/{id}/policies；PUT /api/v1/plugins/{id}/commands/{route_id}/names 发 {names}，DELETE 同路径恢复声明。PUT policies 发送一项完整覆盖，permissions/enabled 的 null 表示继承。
- src/api/admin.ts 工具、发信许可、会话与静态任务：
  export type ToolPolicyValue = 'allow' | 'confirm' | 'deny' | 'inherit'
  export type ToolPolicy = { plugin_id: string; service: string; identities: ('tool' | 'operation')[]; default: ToolPolicyValue; override: ToolPolicyValue; effect: 'read_only' | 'write' | 'destructive' | 'unspecified' }
  export type ToolPolicyUpdate = { plugin_id: string; service: string; policy: ToolPolicyValue }
  export function getToolPolicies(): Promise<{ policies: ToolPolicy[] }>
  export function setToolPolicy(item: ToolPolicyUpdate): Promise<void>
  export function setToolPolicies(items: ToolPolicyUpdate[]): Promise<void>
  export type AllowedInstances = { all: boolean; instances: string[] }
  export function getAllowedInstances(id: string): Promise<AllowedInstances>
  export function setAllowedInstances(id: string, value: AllowedInstances): Promise<void>
  export function listPluginConversations(id: string): Promise<{ key: string; entries: number; updated_at: string }[]>
  export function clearPluginConversation(id: string, key: string): Promise<void>
  export type PluginJob = { job_id: string; trigger: string; paused: boolean; last_run_at?: string; next_run_at?: string }
  export function listPluginJobs(id: string): Promise<{ jobs: PluginJob[] }>
  export function changePluginJob(id: string, jobID: string, action: 'pause' | 'resume'): Promise<void>
  // 工具 GET/PUT /api/v1/tools/policies，单项原样发 item，批量发 {items}；发信 GET/PUT /api/v1/plugins/{id}/allowed-instances；会话 §15 GET 数组、DELETE .../conversations/{key}；任务 §20 GET .../jobs、POST .../jobs/{job_id}/{pause|resume} 无正文。key、job_id 整个值编码成一个不透明 URL 段，保留原值。
- src/api/admin.ts Core 设置与管理认证：
  export function getCoreAdmins(): Promise<string[]>
  export function changeCoreAdmin(action: 'add' | 'remove', user: string): Promise<void>
  export function getCoreDenies(): Promise<{ users: string[]; chats: string[] }>
  export function setCoreDenies(value: { users: string[]; chats: string[] }): Promise<void>
  export function getDefaultLocale(): Promise<{ default_locale: string }>
  export function setDefaultLocale(locale: string): Promise<void>
  export function getDefaultPlatform(): Promise<{ command_prefixes: string[]; reply: { mention_sender: boolean; quote: boolean; prefix: string; forward_threshold: number } }>
  export type GlobalLogLevel = 'debug' | 'info' | 'warn' | 'error' | 'critical'
  export function setGlobalLogLevel(level: GlobalLogLevel): Promise<LogLevels>
  export function rotateAdminToken(token: string): Promise<{ token: string }>
  // 管理员 GET /api/v1/admins、POST /api/v1/admins/{add|remove} 发 {user}（§4，与 §13 的 core/admins 共用服务）；禁止 GET/PUT /api/v1/core/denies，写入两个数组；默认语言 GET/PUT /api/v1/core/default-locale，发 {default_locale}；getDefaultPlatform 只读 GET /api/v1/profiles/default/platform；级别 PUT /api/v1/logs/levels 只发 {global:level}，保留 sources；轮换 POST /api/auth/token/rotate 发 {token}。
- 所有普通请求复用 admin.ts 私有 request、Cookie、内存 X-Alonabot-CSRF、AdminAPIError 与 onSessionExpired。listPackages 保持原签名，发送当前界面的 Accept-Language 以取得本地化 pages，Package.pages 从 unknown[] 收窄为 {path:string;title:string}[]，保留其他字段。rotateAdminToken 成功清空内存 CSRF 并返回新 token，页面立即调用 tokenRotated(result.token)，在根组件的结果对话框显示一次，避免认证页面卸载丢失新 token。
- src/components/PluginPageFrame.tsx：export function PluginPageFrame(props: { id: string; path: string; title: string }): JSX.Element；src/bridge/pluginPage.ts：export function bindPluginPage(frame: HTMLIFrameElement, input: { id: string; locale: string; openSearch: () => void }): { setLocale: (locale: string) => void; dispose: () => void }；src/api/admin.ts 新增 export function requestPluginPage(id: string, input: { method: string; path: string; headers?: Record<string, string>; body_base64?: string }): Promise<{ status: number; headers: Record<string, string>; body_base64: string }>。父桥按 §7 的 JSON 消息收发，凭证和 fetch 留在 admin.ts；bindPluginPage 绑定 frame.contentWindow，卸载、导航和 iframe 重载时释放旧绑定。翻译数据源的缺口按下文交审查者处理，保留 translate 操作及其验收要求。
- src/shell/search.ts 的 SearchContext 值保留 register(items: SearchItem[]): () => void，增加 open(): void；AppShell 提供给桥。GlobalSearch(props) 保留原属性，增加可选 openRequest?: number；AppShell.open 递增请求序号，经 TopBar 传入后聚焦同一个搜索框（src/shell/TopBar.tsx 仅允许这项传参修改）。页面额外操作继续 register，卸载时注销。

行为：
- 配置编辑以 snapshot.pending?.config 或 snapshot.active 为基线，带 active_generation，只发 patch。schema 为 null 时显示无配置状态。i18n 每个语言值是 config 根节点，pickLocale 选中后包装成 {config: selectedMessages} 传给 ConfigForm。行左侧名称与说明、右侧控件；有改动时浮出「放弃 / 保存」。「放弃」恢复最近快照草稿；删除服务端 Pending 使用 discardPluginConfig 的独立显式操作。保存后显示 Pending 并提供重载，调用已有 changePackage('reload', id) 后重读；开发来源保存时已写文件并重启，按重读的 pending:null 展示。重载失败保留 Pending、实际错误、Degraded 与维护原因；陈旧代次 409 / FAILED_PRECONDITION 提示冲突并重读，草稿交用户处理，只有带 affected 数组的 409 进入原有确认流程。
- buildPatch 的 secret、数组整体替换、模板切换、invisible、condition 与未知键保留规则继续生效。含 secret 的数组发生修改时仍要求重新输入其中全部掩码值；保存成功清空明文草稿。secret 查看经 §12.2 的 token 再认证，只在当前视图内存中显示，关闭、切换组件、登出时清除输入与明文；403 保留原原因。另按 PLAN 提供完整 JSON 编辑（见下文裁定第 4 项）。
- options_from 为可搜索单选/多选，打开时 list，已存值用 resolve；path 从 schemaPath 去掉结构层 items/templates，保留模板键，省略数组下标。保存 value，显示 label/description；resolve 缺失的值仍显示原值并标「已不存在」，more 提示继续搜索，允许手输存储值。UNAVAILABLE 显示原错误并按 CONTRACTS 降为文本输入；其他失败保留原原因。来源与 filter 由服务端 Active schema 决定；插件表单、实例表单分别使用自己的 §17 入口，Core 选项供发信实例/聊天目录选择使用。
- example 与字段一起按 condition 显示；源示例单对象和数组、翻译稀疏覆盖、input、文字与图片片段按 D-160 渲染。占位符按当前草稿与当前语言 labels 即时替换，支持 template.$key、所选模板分支、{{ 与 }}；标量格式遵循 §5.1，未选模板分支为空串。示例 input 加所选配置文件的首个命令前缀；本切片用 getDefaultPlatform 只读 §23 的 default 平台设置取得 command_prefixes[0]，配置文件切换由 goal-165 接入。图片用 pluginExampleURL 和 img，失败显示实际加载状态；实例 schema 引用图片与静态端点的缺口见下文。
- 校正记录显示 path、reason 与服务端实际提供的 original，secret 相关记录省略的原值保持省略；unknown 路径逐项可见。「清理无效配置」前提示删除范围，插件 prune 以 Active 为基线、替换已有 Pending，存在草稿或 Pending 时明确确认丢弃这些编辑，成功显示返回 Pending 再显式 reload；开发来源 prune 同样先 Pending 再 reload。Core 与实例 prune 直接写回 core_config.json，成功重读对应配置与记录；保留其他字段和 ${...} 原文。普通读取和草稿放弃都不触发 prune。
- 插件 pages 从 listPackages 的当前运行声明取 path/title，语言切换后重读标题；iframe src 固定为 /plugins/{编码后的所属组件 ID} 加声明路径，sandbox="allow-scripts allow-forms allow-popups"，内容位于父页面插件名、标签与全局搜索下方。SDK kit.js 已提供 window.AlonaBot.request/t/onTheme/onLocale，复用 showcase 联验；插件页面内容由所属插件实现。
- 父桥检查 channel、消息结构、id、op 与 event.source === 绑定窗口，核对不透明源身份，响应只发回该窗口并原样保留 id。请求路径在规范化前拒绝绝对/协议相对 URL、..、编码越界/分隔符混淆、反斜杠、其他组件与 Admin API 路径；固定所属组件前缀，注册 method/path 的最终匹配交 Core 代理。忽略大小写拒绝 Cookie、Authorization、X-Alonabot-CSRF、X-Alonabot-Proxy-Auth 等凭证与内部身份头，父页面添加自己的 CSRF，Origin 由同源 fetch 控制。fetch 拒绝重定向，过滤 Set-Cookie 等响应凭证头；body_base64 保留完整请求/响应字节。插件 HTTP 非 2xx 仍以 ok:true 交付实际 status/headers/body_base64，桥校验/执行失败以 ok:false 的稳定 code 和英文 safe_message 返回；管理 401 同时触发现有过期处理。
- hello 返回 locale、实际 light/dark 与 §7.1 的全部 --ab-* 令牌：bg/surface/surface-2/line/accent/accent-soft/success/warning/danger/critical 映射 --color-*，text/text-muted 映射 --color-ink/--color-ink-muted，圆角与字体映射 --radius-*、--font-*。主题、系统明暗、令牌或语言变化只向已 hello 的窗口推送无 id 事件；locale/theme 只返回相应值。translate 按组件 messages 键与命名 args 格式化、未知键回键名、缺参数保留占位符。shortcut 只接受 mod+k，打开外壳第一行最右边的同一个搜索；其他键返回桥错误。
- 静态任务列表显示 trigger、last_run_at、next_run_at；缺失时刻显示未执行/未安排状态。开关发 pause/resume，成功重读；暂停状态持久保存，下次注册沿用，暂停前已派发调用可完成。保留原语句的代码任务只提供开关、无删除行为，动态用户任务交 Agent 插件页面。
- 命令按插件分组折叠，路径 token 构建命令树；plugin/group/route 三层覆盖显示显式值与 GET 返回的有效 permissions/enabled，页面沿用服务端有效结果，父级禁用保持生效。permissions:null 继承、[] 无权限要求，enabled:null 继承；编辑一项时携带另一项现有覆盖，两项 null 清除覆盖。显示 shadows 的 winner/shadowed 与失效覆盖诊断。命令名按 token 数组编辑，显示不加 /；替换、追加、删除、调整首项主名均 PUT 完整 names，至少留一项，恢复声明用 DELETE。Raw/文本匹配的名称数组为空时仅显示权限；409 / ALREADY_EXISTS 展示 safe_message 中冲突 route_id，保留草稿；保存后重读，无需重启。
- 工具按插件分组，类别为只读/写入/破坏性/未声明；展示 default 与 override，允许/询问/禁止分别写 allow/confirm/deny，恢复继承写 inherit。按类别批量设置把该类别选中目录项一次发 {items}，保持 MCP 完整操作名，服务端事务保证整批提交；失败保留 items[index] 的原错误并重读。后台发信许可放在详情高级标签，复用 listInstances 与 core.adapter_instances 选项，all:true 全部、all:false 按 instances，空数组拒绝全部；保存后下次发信生效。
- Core Admin 列表、默认 locale 与禁止表使用 §4、§13，在系统设置提供名单与默认语言编辑；管理员操作共用 changeCoreAdmin，禁止保存带 users/chats 两数组并保留未编辑项，失败后 GET 核对可能已提交部分。权限「用户」的统一表（身份与禁止状态、添加时从聊天目录选择）由 goal-156 接入，复用这些客户端导出；聊天目录返回会话，用户身份仍按 <adapter_id>:<external_user_id> 保存。
- 会话按插件列出 key、entries、updated_at，支持插件筛选、搜索与数字分页，清除前确认并把完整 key 交 clearPluginConversation，成功重读；读取失败展示实际错误。保留配置文件列与筛选目标，映射缺口见下文；配置文件管理页由 goal-165 提供，条目正文由 Agent 插件页面负责。
- 系统设置按桌面后台样式左名称/说明、右控件：界面语言只改本机偏好，Core default_locale 单独标识并 PUT，主题经共享 useTheme 状态选浅色/深色/跟随系统。全局级别使用 getLogLevels 与 setGlobalLogLevel，含 CRITICAL，组件来源覆盖保持；失败重读有效值并显示保存/通知原错误。管理令牌对话框输入当前 token 再认证，成功提示全部会话已撤销，一次显示新 token 供复制，确认后到登录页；新旧 token 均只在当前对话框内存中，关闭后清空。登出沿用现有 logout 与登录跳转。网络代理、日志保留与右上角重启 Core 保留画板要求，接口缺口按下文处理；命令前缀在配置文件页，休眠控件在 goal-175。
- 所有列表按共同约定提供筛选、分组折叠或数字分页，深浅色、窄于 768px 单列与手机「更多」沿用外壳。与 goal-94、goal-146、goal-156 共享 admin.ts、routes、详情、外壳和文案，后合并者同步 main，直接复用已加入的导出与请求处理。
- 审查者裁定（2026-11-17，记 D-177）：（1）`GET /api/v1/plugins/{id}/config` 的响应在 `i18n` 旁追加 `messages`：按包内 `i18n/*.json` 的语言标签返回各文件的 `messages` 根节点（开发来源读当前工程资源），缺失时为 `{}`；本切片获授权修改 `internal/core/manage` 中对应的资源读取与响应、`cmd/alonabot/manage_e2e_test.go` 中既有断言的同步，以及 ADMIN-API §12 的说明；父桥的 `translate` 用它取文案。（2）网络代理、日志保留与重启 Core 没有管理接口，本切片的系统设置页不显示这三项控件，验收不含它们，PR 中列为后继。（3）会话列表的配置文件列只对能解析出会话的键（`<会话字符串>@<实例名>`，以及成员会话追加的 `@<用户字符串>` 前缀部分）按 §23 assignments 的会话指定显示，其余键显示为空，不另造映射。（4）完整 JSON 编辑按 PLAN 保留：插件详情「高级」标签提供整份配置的 JSON 编辑（secret 仍只显示设置状态，提交走同一 PATCH 语义）。（5）回复示例图片只对插件 `config_schema.json` 提供（CONTRACTS §7.5 的端点范围）；Adapter 实例 schema 中的示例只渲染文字与图片的 alt，不请求图片。

测试（按 README「编写测试段」与编码规范「测试提交规则」）：
- 提交：无。
- 本地自测：*.local.test.ts 覆盖 options_from/example/profile 解析与旧 $version/$migrations 拒绝、嵌套字段路径、搜索/resolve/more/已不存在/UNAVAILABLE、示例翻译回退/条件/模板/占位符/图片路径、Merge Patch 与 secret/数组保留、Pending 与陈旧代次、三种 prune 的基线与原值脱敏、命令三层覆盖/名称顺序/冲突/遮蔽、工具类别与批量错误、会话和 job_id 整段编码、全局级别与来源保留、轮换后的 token 展示和认证切换。其余 Go 回归若确需核对，写被忽略的 *_local_test.go，PR 记录用例与结果。
- 父桥本地自测用 *.local.test.ts 覆盖窗口绑定/卸载、坏消息与关联 id、路径原始校验/编码绕过、请求与响应凭证头/重定向、完整二进制及 HTTP 错误状态、hello 前后事件、全部令牌映射、locale/theme/translate/shortcut 各操作；只用现有 vitest 与浏览器 API，DOM 测试依赖按共同约束处理。
- 浏览器手动验证：真实 Core 上配置保存/Pending/reload/失败/代次冲突、secret 再认证、回复示例与 options_from、插件/Core/实例三种校正提示与清理、版本回退和卸载、任务 pause/resume 后重新注册、命令名与权限、工具单项和按类别策略、发信许可、会话清除、默认语言/界面语言/主题/全局日志级别/令牌轮换及旧会话失效。用真实 showcase 的 sdk/page 联验沙箱、首次翻译、换语言重新取词、主题与搜索；检查深浅色、390px、键盘和错误展示。PR 写明浏览器与版本、实际环境、已完成流程与未运行项，契约缺口保留对应验收状态。

完成条件：npm --prefix web/admin run typecheck、npm --prefix web/admin run test、npm --prefix web/admin run build 全部成功；PR 记录命令与真实 Core 的手动验证结果。每项原目标均有实现或明确的承接切片；上述裁定的实现与验收完成后交付，提交内容只含允许范围。审查者按 TEST-MATRIX §3 记录页面证据，合并后按真实导出同步 goal-45、goal-165、goal-175 等后继语句，并在同一 goals PR 中把前置已齐的 goal 改写为待派发，总览与状态行同步。

不做：Go 代码、管理 API、spec 修改；日志页（goal-94）；商店、统计、用户目录、存储与数据页（数据库表格/保留代次只读、KV、文件、导出清空、会话记录，由 goal-156 承接）；插件页面自身内容与动态用户任务；模型与配置文件页及插件配置文件切换（goal-146、goal-165）；休眠控件、锁定原因与运行模式（goal-175）；新依赖、CI 前端任务与提交前端测试。
```

### goal-94 · T-09.15 日志页

状态：已完成（PR #243），合并提交 `d9e9089`。以下保留派发时语句。派发时：前置 T-09.8（PR #238，合并提交 `9123839`）、T-09.10（PR #177，`99efcd1`）、T-09.19（PR #213，`6f6856c`）、T-02.5（PR #209，`d75e43d`）均已合并。按 main `50244fe` 核对：`web/admin/src/api/admin.ts` 已有 `AdminAPIError`、`onSessionExpired(handler: () => void): () => void`、`listPackages(): Promise<Package[]>`、`listInstances(): Promise<AdapterInstance[]>` 与含 `critical` 的 `SourceLogLevel`，`request<T>` 与 CSRF 状态仍为模块私有；日志读取与二进制下载尚无前端导出。`internal/core/manage/log_api.go` 已注册 recent、stream、history，`diagnostics_api.go` 已注册诊断包端点；`src/routes.tsx` 的 `RouteEntry` 与 `src/main.tsx` 的精确路径分派可直接挂载 `/logs`，总览已有按 `id: logs` 显示「查看日志」的入口。`src/i18n/UI.tsx` 导出 `useUI(): typeof english.messages`；`src/components/ConfigForm.tsx` 的 `ConfigForm({schema, value, messages, onSubmit})` 保持原用途，本页的筛选与诊断包选项使用普通控件。沿用 D-169 的 v4 画板、D-156 与 D-158；级别设置继续由 goal-44、goal-43 负责。

```text
/goal 使用 $alonabot-execute 推进 goal-94（切片 T-09.15：日志页）。goal 目录 goal-94，分支 goal-94，worktree .worktrees/goal-94，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：实现日志页：终端视图实时显示日志，按级别、来源与日期筛选，搜索，清空已载入的记录，导出诊断包。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（日志、诊断包）及其停止和失败条件；spec/plan/webui.md §3；CONTRACTS §14.1、§14.3、§14.4、§14.7；ADMIN-API §8、§19；DECISIONS D-156、D-158、D-169；TEST-MATRIX V-62、V-63、V-127、V-129 的页面分项。画板：「v4 · 日志」「v4 · 日志（筛选）」「v4 · 日志（导出诊断包）」「v4 · 日志（深色）」「v4 · 手机 · 日志」。审查者将截图放在 goal-94/figma/；首轮由前端执行者实现，审查与修复交 GPT。

范围（只新增或修改这些；路径相对 web/admin/）：
- src/pages/LogsPage.tsx 及 src/pages/ 下日志页专用的筛选、记录格式化与诊断包对话框模块，src/api/admin.ts。
- src/routes.tsx、src/style.css、i18n/en-US.json、i18n/zh-CN.json；对应的 *.local.test.ts。
- package.json、package-lock.json 仅用于首次实际引入已批准的 @radix-ui/react-popover。

允许的依赖：现有 React、Radix、lucide-react、TypeScript、Vite、vitest；筛选弹层可首次引入 docs/development.md 已批准的 @radix-ui/react-popover 并锁版本。

接口（新增导出写在指定模块；HTTP 格式按 ADMIN-API）：
- src/pages/LogsPage.tsx：export function LogsPage(): JSX.Element。src/routes.tsx 的 routes 增加 {id: 'logs', path: '/logs', label: 'logs', icon, group: 'main', render: () => <LogsPage />}，icon 取 lucide-react。main.tsx 已在认证后执行 matchedRoute.render()，沿用 AppShell、手机底栏与全局搜索；无需再建页面分派。label 与新增文案放在双语 JSON 的 messages 下，经 useUI() 读取；日志级别保留英文显示。
- src/api/admin.ts 复用私有 request 的同源 Cookie、内存 CSRF、错误解析与 401 通知，补齐 SSE 和 Blob 响应的处理，保持现有导出及行为：
  export type LogRecord = { time: string; level: string; source: string; message: string; attrs: Record<string, unknown>; trace_id: string }
  export type DiagnosticsOptions = { since: string; logs: boolean; settings: boolean; components: boolean; redact: boolean }
  export function getRecentLogs(): Promise<LogRecord[]>
  export function getLogHistory(date: string): Promise<LogRecord[]>
  export function subscribeLogs(handlers: { onLog: (record: LogRecord) => void; onOverflow: (lastTime: string) => void; onError: (error: Error) => void }): () => void
  export function downloadDiagnostics(input: DiagnosticsOptions): Promise<{ blob: Blob; filename: string }>
- getRecentLogs → GET /api/v1/logs/recent；getLogHistory → GET /api/v1/logs/history?date=YYYY-MM-DD；subscribeLogs → GET /api/v1/logs/stream。三者均不传 source、level、trace，取得未筛选数据后在页面多选和搜索；服务端 level 是最低级别，source 是单值，无法表达页面的多选。recent 返回时间升序的 Record[]，history 保留文件续写顺序。SSE 监听 event: log（JSON Record）和 event: overflow（JSON 字符串，最后交付时间），返回函数取消请求并释放流；页面退出、登出、关闭实时或选择过去日期时调用。
- LogRecord.time 保留 RFC3339Nano 原串；实时级别来自 slog（警告为 WARN，界面显示 WARNING；CRITICAL 为独立级别）。源码位置取 attrs.source 的 {file,line,function}，顶层 source 是来源身份。历史时间仅毫秒精度、trace_id 最多为 8 位前缀，source 是显示标签，message 含原控制台正文，按 §8.2 原样保留。
- downloadDiagnostics → POST /api/v1/diagnostics，JSON 为 DiagnosticsOptions；对话框缺省 {since: '24h', logs: true, settings: true, components: true, redact: true}。成功读取 application/zip 与 Content-Disposition 附件名，生成下载后释放对象 URL；失败保留 JSON/纯文本错误。400 包括非法正时长、三组全关、非法或未知 JSON 字段；失败响应按状态解析，保持失败对话框和选择。
- 来源选择复用 listPackages、listInstances，按 Core、组件、账号分组；实时组件 source 按包 ID，账号按 attrs 的实例字段区分，历史按 §8.2 的标签筛选，保留无法还原身份的原标签。页面操作经 SearchContext.register(items: SearchItem[]): () => void 注册，卸载时注销。与 goal-156、goal-146 共用 admin.ts、routes、样式和双语文案，后合并者复用已合入的请求处理与导出，保留其他页面。

行为：
- 只有终端视图：终端区域用 --color-console-* 与 --font-mono，每条记录一行，列定宽对齐：时间、来源、级别、源码位置（取 Record.attrs.source 中 file 的最后两段与 line，悬停显示 function 与完整路径，点击复制）、消息；消息过长时换行并对齐到消息列；方括号内不补空格。ERROR 整行红色，WARNING 整行黄色（--color-yellow，不用棕色或琥珀色），CRITICAL 整行紫色，DEBUG 用暗色；不用整行色块，不加柱状分布图、字号按钮或列表与终端切换。
- 工具栏：「筛选」按钮打开弹层（级别多选含 DEBUG 与 CRITICAL；来源分 Core、插件与账号多选，可搜索；日期），筛选生效时按钮带圆点；CRITICAL、ERROR 与 WARNING 计数，点击只看该级别；搜索内容或 trace；实时开关；「清空」只丢弃页面已载入的记录，不删除日志文件；「导出诊断包」打开对话框，选择时间范围与三组内容，缺省脱敏（ADMIN-API §19）。
- 今天的记录先读 recent，再订阅 SSE（会先重发缓冲中的记录），只对 recent 与初始快照的重叠部分按完整时间与记录内容去重，后续实时记录保留重复出现的次数。过去日期查询 history，日期按 Core 本地时区语义查询；历史的精度与字段保留按 §8.2 展示。筛选与搜索只在浏览器中进行，条件保存在 localStorage；日志正文与诊断包留在当前页面内存。
- overflow 或异常断流显示实际状态与最后交付时间，提供重新读取 recent 并订阅的操作；重新连接的快照按同一重叠规则合并，缓冲之外的缺口保持可见。「清空」保持订阅并丢弃当前显示记录，后续新记录继续追加；计数对应当前已载入的记录，点击级别计数应用该级别筛选。
- 手机上每条记录两行：第一行时间、级别、来源与源码位置（截断），第二行消息；实时、清空与导出收进 ⋯。

测试（按编码规范「测试提交规则」）：
- 提交：无。
- 本地自测：*.local.test.ts 覆盖初始快照重叠去重与后续同文记录、WARN/WARNING 显示、CRITICAL、源码位置与缺字段、历史原文与 trace 前缀、筛选/搜索/计数、清空后接收新记录、SSE 分块与 overflow、诊断包错误解析与下载文件名。
- 浏览器手动验证：真实 Core 上走 recent → SSE、过去日期与无文件、级别/来源/账号/日期筛选、搜索、关闭实时、清空、断流后重连、未认证与会话过期、导出缺省脱敏包与取消；核对深浅色、390px 手机两行布局、源码复制与长消息换行。PR 写明浏览器与版本、命令结果及未运行项，诊断包核对三组选项与真实脱敏内容。

完成条件：npm --prefix web/admin ci、npm --prefix web/admin run typecheck、npm --prefix web/admin run test、npm --prefix web/admin run build 全部成功；go build ./... 不受影响，PR 写明实际验证结果。审查者按 TEST-MATRIX §3 记录页面证据。

不做：日志 API、诊断包与 CLI（T-09.10、T-09.19）；级别设置（goal-44、goal-43）；Go 代码、spec 修改、CI 前端任务与提交前端测试。
```

### goal-156 · T-09.23 官方 WebUI 的后台进程、统计、商店、用户与数据页

状态：已完成（PR #245），合并提交 `f6cf850`。以下保留派发时语句。派发时：前置 T-09.8（PR #238，合并提交 `9123839`）、T-09.20（PR #221，`bd386df`）、T-09.21（PR #234，`3b3ca6f`）、T-09.22（PR #220，`b5e572e`）、T-09.24（PR #215，`b911f27`）均已合并。按 main `50244fe` 核对：`web/admin/plugin.yaml` 的实际 ID 为 `alonabot-webui-admin`，assets 为 dist，尚无 Go 源码；`github.com/AlonaBot/AlonaBot/sdk` 提供 `NewPlugin() *Plugin`、`(*Plugin).On(routeID string, eventType EventType) *EventBuilder`、`(*Plugin).AfterSend(routeID string) *AfterSendBuilder`、`(*Plugin).Route(method, path string) *RouteBuilder`，可用 `sdk.MessageCreated` 与 `sdk.CommandInvoked`。`sdk/sqlitedb.Open(ctx context.Context) (*sql.DB, error)` 与 `Migrations(p *sdk.Plugin, list []Migration)` 可保存受管数据；`contracts/pluginindex.Parse(raw []byte) (*Index, error)` 已解析源索引。`src/api/admin.ts` 已有 `listPackages`、`installPackage`、`updatePackage`、`listGenerations(id: string): Promise<Generation[]>`、`listInstances`；`OverviewPage({core})` 的四组图表仍为空态，`PluginsPage()` 已有商店标签入口，`PluginDetailPage({id, generations?})` 已有版本记录与高级标签，`AppShell` 已接收可选 `coreUpdate: {version, releaseURL}`。`scripts/dist/main.go` 已逐组件调用多平台 pack，但 WebUI 在脱离根 Go module 的临时副本构建，带 Go 入口后须同步暂存方式。下文保留 D-162～D-166、D-169 的目标；日志级别统计来源、会话正文读取与官方源地址的缺口在行为段列明，交审查者裁定。

```text
/goal 使用 $alonabot-execute 推进 goal-156（切片 T-09.23：官方 WebUI 的后台进程、统计、商店、用户与数据页）。goal 目录 goal-156，分支 goal-156，worktree .worktrees/goal-156，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：为官方 WebUI 组件加上 Go 入口（D-162）：订阅 `command.invoked` 与消息事件，按小时分桶统计收发量、命令触发与错误，提供统计路由；记录见过的用户，读取插件源 `index.json` 并缓存，检查 Core 与已安装组件的更新。前端接入总览图表、插件商店与发布插件、权限「用户」页、插件详情的「存储」与会话记录视图、版本记录与 changelog 显示。

依据：PLAN T-09 步骤 7（WebUI 后台）及其停止和失败条件；spec/plan/webui.md §2、§3；CONTRACTS §6、§7.5、§9.1（WebUI 的 Go 入口）、§9.2（插件源）、§8（标准事件 command.invoked）；ADMIN-API §4、§13、§15、§20～§22；DECISIONS D-162～D-166、D-169；TEST-MATRIX V-136，V-134、V-135、V-138 的页面分项。画板：「v4 · 总览」「v4 · 插件 · 商店」「v4 · 插件 · 商店（发布插件）」「v4 · 权限 · 用户」「v4 · 权限 · 用户（添加用户）」「v4 · 插件 · 天气 · 数据（存储）」「v4 · 插件 · Agent · 数据（会话记录）」及其深色与手机画板。审查者将截图放在 goal-156/figma/；首轮仍由前端执行者负责整个切片，含 Go 后台和打包接入，审查与修复交 GPT。Go 部分按下文单独验证。

范围（只新增或修改这些）：
- Go 后台：web/admin/ 下的可导入入口包（package admin），含 New、事件/Hook、受管 SQLite 的 *.sql 与迁移登记、统计/用户/商店/更新/已读版本 HTTP 处理；web/admin/plugin.yaml 保留现有 ID、role、assets 与 Admin API 主版本范围。沿用仓库根 go.mod，不新建嵌套 module。
- 前端：web/admin/src/api/admin.ts、src/api/webui.ts；src/pages/OverviewPage.tsx、PluginsPage.tsx、PluginDetailPage.tsx、UsersPage.tsx，及 src/pages/、src/components/ 下由这些页面使用的图表、商店、存储、用户与 changelog 模块；src/main.tsx、src/routes.tsx、src/shell/AppShell.tsx、Sidebar.tsx、AboutDialog.tsx；src/style.css、i18n/en-US.json、i18n/zh-CN.json。页面路径均相对 web/admin/。
- 打包：scripts/dist/main.go 中 WebUI 的暂存、前端构建与 pack 调用；保留同一根 module 的 Go 入口、SDK 和依赖可解析性，产物包括各平台进程与同一份 dist。沿用现有签名与发行目录，T-14.3 的文档同步写入 PR「文档同步」。
- web/admin/package.json、package-lock.json 仅用于首次实际引入已批准的 recharts 或 @radix-ui/react-popover；以上模块对应的 *.local.test.ts、*_local_test.go 仅作本地自测。

允许的依赖：
- Go：标准库、sdk、sdk/message、sdk/sqlitedb、contracts/pluginindex、contracts/release、contracts/manifest；本地集成自测可复用 plugintest 与现有测试 Adapter。SQLite 使用现有已批准依赖，经 SDK 连接。后台只使用自己的受管库和 SDK，Core 管理数据由前端调用 Admin API。
- 前端：现有 React、Radix、lucide-react、TypeScript、Vite、vitest，以及 docs/development.md 已批准的 recharts、@radix-ui/react-popover（首次使用时锁版本）。Markdown 解析依赖尚无批准项，需要引入时按共同约束向维护者申请。
- 打包：复用现有 internal/pluginkit、生成器与 scripts/dist 的依赖；Core、SDK、协议与安装器保持现有实现。

接口（复用 main 的真实导出；以下 WebUI 私有路由由本切片新增）：
- web/admin 的 Go 入口导出 func New() *sdk.Plugin，使用 sdk.NewPlugin；通过 sqlitedb.Migrations(p, []sqlitedb.Migration{...}) 登记数据库表，回调与 HTTP 使用 sqlitedb.Open(ctx)，连接由 SDK 管理。New 内登记 p.PreEvent(..., sdk.MessageCreated).Handle(func(c *sdk.Context, payload *message.CreatedPayload) (*message.CreatedPayload, error))、p.On(..., sdk.CommandInvoked).Handle(func(c *sdk.Context, payload *message.CommandInvokedPayload) error) 与 p.AfterSend(...).Handle(func(c *sdk.Context, to string, result *message.SendResult, sendErr error) error)。PreEvent 返回原载荷，只观察；普通事件路由位于多轮等待与命令 STOP 之后，入站采集复用已有 PreEvent 观察点。读取 c.Event() *message.Event、c.Context() context.Context、c.User() string；入站账号取 Event.AdapterInstanceID，发送账号取 result.AdapterInstanceID，c.Instance() 只表示继承的默认目标，跨账号发送按实际回执归属。
- HTTP 用 p.Route(method, path).HandleFunc(func(http.ResponseWriter, *http.Request))，保持默认管理会话认证。SDK 登记的路径相对组件；公开 URL 为 /plugins/alonabot-webui-admin 加下列路径，GET 返回 JSON，变更请求由 Core 校验 Cookie、Origin、CSRF。它们属于 D-162 的官方 WebUI 接口；本章「只调用 ADMIN-API」对 Core 管理接口的约束继续适用。
- src/api/webui.ts 的新增导出与后台 JSON 使用下列同形类型：
  export type StatisticsRange = '24h' | '7d' | '30d'
  export type HourBucket = { hour: string; incoming: number; outgoing: number; send_errors: number; accounts: { instance_id: string; incoming: number; outgoing: number }[]; commands: { plugin_id: string; route_id: string; command_name: string; result: 'ok' | 'error' | 'denied'; count: number; duration_ms: number }[] }
  export type Statistics = { timezone: string; observed_since: string; buckets: HourBucket[] }
  export type SeenUser = { user: string; display_name: string; first_seen_at: string; last_seen_at: string }
  export type StoreEntry = { repo: string; role: string; tags: string[]; id: string; name: string; description?: string; version: string; author?: string; stars?: number; icon_url?: string }
  export type StorePage = { source: string; fetched_at: string; stale: boolean; error?: string; items: StoreEntry[]; total: number }
  export type LatestRelease = { repo: string; version: string; release_url: string; markdown: string; fetched_at: string; stale: boolean; error?: string }
  export function getStatistics(range: StatisticsRange): Promise<Statistics>
  export function listSeenUsers(): Promise<SeenUser[]>
  export function getStore(input: { source: string; role: string; q: string; offset: number; limit: number }): Promise<StorePage>
  export function getLatestRelease(repo: string): Promise<LatestRelease>
  export function getSeenVersion(component: string): Promise<{ version: string | null }>
  export function setSeenVersion(component: string, version: string): Promise<void>
- 对应 HTTP：GET /stats?range=24h|7d|30d；GET /users（完整见过用户数组，由前端合并身份后筛选和数字分页）；GET /store?source=&role=&q=&offset=&limit=；GET /updates?repo=owner/repo；GET /changelog/seen?component=core或组件ID；PUT /changelog/seen 发 {component,version}，成功 {}。商店查询返回精确 total，分页值由请求显式提供，source 为源仓库 owner/repo，读取其默认分支根目录的 index.json；无已读版本返回 null。hour 为带时区偏移的 RFC3339 小时起点，timezone 为后台本地时区，observed_since 为开始采集的实际时间；统计时窗按后台当前时间选取最近 24 小时、7 天、30 天。后台错误返回非 2xx 与 {code,safe_message}；外部抓取失败但有缓存时返回 stale: true 及 error，保留缓存原时间，无缓存时返回实际失败。
- admin.ts 复用私有 request 的 Cookie、CSRF、AdminAPIError、401 通知并补齐 Blob 下载处理；给 webui.ts 增加导出 adminRequest<T>(path: string, options?: { method?: string; headers?: Record<string, string>; body?: string | FormData | Blob }): Promise<T>，只将现有请求能力提供给同源 WebUI 客户端，CSRF 状态继续私有。若先合并切片已有同等请求导出，直接复用其名字和签名。
- Core 管理继续使用已有 listPackages(): Promise<Package[]>、listInstances(): Promise<AdapterInstance[]>、listGenerations(id: string): Promise<Generation[]>、installPackage(input: {Source: string; Replace?: boolean; Confirm?: boolean} | FormData): Promise<Package>、updatePackage(id: string, input: {confirm?: boolean} | FormData): Promise<{warnings?: string[]}> 与 isConfirmationError。版本记录复用现有详情组件及 rollbackPackage，不新建回退流程；按 ADMIN-API §20 读取任务与 pages 的工作沿用 goal-44，后合并者保留其标签与导航入口。
- admin.ts 补齐存储导出（响应字段原样采用 ADMIN-API §21）：
  export function listPluginKV(id: string, input: { prefix: string; offset: number; limit: number }): Promise<{ items: { key: string; size: number; updated_at: string }[]; total: number }>
  export function getPluginKV(id: string, key: string): Promise<{ value_base64: string; text?: string }>
  export function setPluginKV(id: string, key: string, value: { text: string } | { value_base64: string }): Promise<void>
  export function deletePluginKV(id: string, key: string): Promise<void>
  export function listPluginFiles(id: string, path: string): Promise<{ entries: { name: string; dir: boolean; size: number; modified_at: string }[] }>
  export function downloadPluginFile(id: string, path: string): Promise<Blob>
  export function uploadPluginFile(id: string, path: string, file: File, overwrite: boolean): Promise<void>
  export function makePluginDirectory(id: string, path: string): Promise<void>
  export function renamePluginFile(id: string, from: string, to: string): Promise<void>
  export function deletePluginFile(id: string, path: string, recursive: boolean): Promise<void>
  export function exportPluginData(id: string): Promise<Blob>
  export function clearPluginData(id: string, confirm: true): Promise<void>
  // 依次调用 §21 的 data/kv、data/kv/{key} GET/PUT/DELETE、data/files、data/files/content GET/PUT（原始文件体，overwrite=1）、data/files/mkdir、rename、delete、data/export、data/clear；文件路径放在 query/JSON，KV 的整个 key 编码为一个不透明路径段。私有请求处理增加 File/Blob 正文支持，不设置 JSON Content-Type。
- admin.ts 补齐用户、会话摘要与 changelog 导出：
  export function getCoreAdmins(): Promise<string[]>
  export function changeCoreAdmin(action: 'add' | 'remove', user: string): Promise<void>
  export function getCoreDenies(): Promise<{ users: string[]; chats: string[] }>
  export function setCoreDenies(value: { users: string[]; chats: string[] }): Promise<void>
  export function listPluginConversations(id: string): Promise<{ key: string; entries: number; updated_at: string }[]>
  export function clearPluginConversation(id: string, key: string): Promise<void>
  export type Changelog = { locale: string; entries: { version: string; date?: string; markdown: string }[] }
  export function getPackageChangelog(id: string, locale: string): Promise<Changelog>
  export function getCoreChangelog(locale: string): Promise<Changelog>
  // 分别为 GET /api/v1/admins、POST /api/v1/admins/{action} 发 {user}、GET/PUT /api/v1/core/denies（两个数组均必填）、§15 conversations GET/DELETE（key 整段转义）、§22 两个 changelog GET（Accept-Language）。
- UsersPage(): JSX.Element 在 src/pages/UsersPage.tsx 导出，作为权限「用户」入口接入 routes（id: 'users', path: '/permissions/users', label: 'users', group: 'main'）；若 goal-44 已有权限页入口，则作为其用户标签挂载，复用同一列表与动作。商店放进 PluginsPage 已有 store 标签，存储与会话记录放进 PluginDetailPage 的数据标签；保持 /components/{id} 由 main.tsx 分派，避开 /plugins/ 保留路径。已有页面导出与 AppShell({version,onLogout?,coreUpdate?,children}) 保持，main.tsx 取得真实新版信息后传 coreUpdate；Sidebar 和 AboutDialog 使用同一 Core 版本与 changelog。
- 界面复用 UIContext、useUI()，routes.label 与文案写入 en-US/zh-CN 的 messages，键集合一致；导航用 shell/router 的 navigate、Link，搜索用 SearchContext.register 并在卸载时注销，深浅色用 theme/theme 的 useTheme。ConfigForm({schema,value,messages,onSubmit}) 仍服务于插件配置；存储的 KV、文件与用户操作使用专用控件。goal-44、goal-94、goal-146 与本切片共享 routes、main、admin.ts、详情、外壳和文案，按执行顺序串行合并，后合并者按 main 的实际导出复用。

行为：
- 后台进程放在 web/admin 的 Go 入口包中，按插件的方式使用受管 SQLite 保存小时桶；路由要求管理会话，前端同源调用。统计只依赖 Core 的基础设施（事件与 command.invoked），Core 不提供图表 API。
- 总览：消息量（收到与发出）、各账号、热门插件（只算常规插件的命令触发次数）、错误与警告图表，时间范围 24 小时、7 天、30 天；故障标在曲线上。
- 收到消息在 message.created 的 PreEvent 按可信 Event.ID/OccurredAt/AdapterInstanceID 记账，覆盖随后进入多轮等待与命令 STOP 的消息；command.invoked 的 PluginID、RouteID、CommandName、Result、DurationMS 分别累计，命令载荷的 Result 保留 ok/error/denied。发出消息以 AfterSend 的成功发送事实计数，sendErr 记录发送失败；c.Editing() 非空的编辑操作单独排除。小时桶以绝对时间保存并返回后台时区偏移，跨午夜与夏令时保留实际小时身份。记录处理、桶更新与用户更新时间在受管事务内完成，重启后读取已落库结果；停机时释放本切片持有的资源，数据库连接交 SDK 关闭。
- 用户：从入站可信 Actor 记录 user、显示名及首次/最近见过时间，用户键用 c.User()，跨同一 Adapter 的实例合并。前端把见过的用户与当前 Core Admin、denies.users 合并，管理员和禁止表中尚未见过的用户也列出；管理员为身份、禁止为状态。按画板搜索、筛选与数字分页；授予/移除管理员调用 §4，禁止/解除调用 §13，写回保留 denies.chats 与其他用户，失败显示实际原因并重读当前状态。和 goal-44 共用一个权限「用户」页。
- 商店：网格卡片，分类只按组件类型，作者标签显示在卡片上；只显示 GitHub 星数；官方组织（alonabot/*）的插件名旁加认证图标；可以切换插件源（缺省官方源）；分页用数字页码；「发布插件」入口说明如何把仓库加入 index.json，并链接插件模板仓库 AlonaBot/alonabot-plugin-template（D-169）。
- 源索引用 pluginindex.Parse 校验；名字、描述、版本与图标从 Release 和包内 manifest 取得，stars 从 GitHub 仓库信息取得，缺字段保持可见的缺失状态。图标只经 img 引用，icon_url 可使用缓存图片的 data URL。缓存源索引、展示元数据与 Release 响应，切换源后按实际源隔离；网络失败、损坏索引与 GitHub 限流返回原错误，有缓存时标出缓存时间。安装仍调用 installPackage({Source: repo})，替换与信任确认沿用 §5 的 affected 对话框，成功重读包列表；更新沿原来源调用 updatePackage，Core 新版按钮打开发行说明。浏览器直接请求本组件后台，外部抓取由后台执行。
- 存储视图：KV（键列表、查看与编辑、删除）、数据目录文件（浏览、上传、下载、新建目录、改名、删除）、整个插件数据的导出与清空（ADMIN-API §21）；会话记录视图列出 Core 托管的 Agent 会话摘要，可清除（正文查看归 goal-167）。
- 更新后在插件详情与「关于」中显示 changelog（ADMIN-API §22，Markdown 由前端渲染，不执行其中的 HTML）。
- KV 分页、保留原始文本或 base64，文件上传覆盖 409 与递归删除分别显式确认；导出下载 zip、清空发 {confirm:true}。Active 库写入、文件目录共用与 404 无库按 §21 显示；数据库表格与保留代次只读视图沿用 goal-44，存储视图和导出不冒充代次文件快照。错误保留完整服务端响应，清空成功后重读各列表。
- 更新前显示 GitHub Release 正文预览并标出来源，更新后按 §22 读取已安装包与 Core 的发布条目。后台按 component 保存管理员已看版本，关闭更新内容对话框后写入本次实际已展示版本；再次打开只提示其后版本，详情和「关于」仍可主动查看完整条目。新旧版本按 SemVer 比较，省略 Unreleased；Markdown 渲染保留文本与链接，原始 HTML 作为文本处理，避免执行主动内容。
- 后台进程的 build 与 pack 按 D-162 照插件出平台包；官方 WebUI 从此按平台打包（T-14.3 同步）。
- 审查者裁定（2026-11-17）：（1）「错误与警告」曲线与故障标记的数据来自 Core 现有的日志查询接口（ADMIN-API 日志一节，按级别 WARNING/CRITICAL 与时间范围筛选后按小时汇总）与组件列表的状态（Degraded、候选失败），不新增 Core 接口或事件；send_errors 与命令 error 仍按各自含义统计。（2）Agent 会话正文的查看属于官方 Agent 的插件页面（T-12.20，goal-167），本切片的会话记录视图只列出 ADMIN-API §15 的会话摘要并支持清除，不读取正文。（3）预置的官方插件源为 GitHub 仓库 `AlonaBot/plugin-store`（维护者 2026-11-17 指定，D-176），作为源列表的缺省项；商店另提供添加与移除插件源。

测试（按编码规范「测试提交规则」）：
- 提交：无。按本章前端切片约定，Go 后台与发行脚本的验证也写入本地文件，PR 记录用例名、命令与结果。
- Go 本地自测（*_local_test.go）：在真实 Core 中加载官方 WebUI 后台，使用现有测试 Adapter 与插件注入消息和命令，验证小时桶、账号、等待/STOP 消息采集、常规插件过滤、AfterSend 成功/失败及编辑区别、跨账号发送归属、用户合并、Core/WebUI 重启后已落库统计与已读版本；HTTP 未登录与缺 CSRF 的拒绝。V-136 其余分项覆盖午夜/夏令时、索引解析、源切换、抓取失败、GitHub 限流后带错误的缓存；外部站点用本地 HTTP 模拟时明确标为模拟。打包本地自测覆盖暂存后 Go module 可解析、进程与 dist 同包及各平台资产，不新增 testdata 测试插件。
- 前端本地自测（*.local.test.ts）：统计分组与真实零值、用户去重和状态合并、过滤/分页、KV 的 UTF-8/base64、路径编码、覆盖/清空确认、SemVer 与已读版本、changelog 语言与 HTML 安全处理。纯函数自测均不提交。
- 浏览器手动验证：同源真实 Core 上的图表范围、商店切换/缓存错误/安装/替换确认、用户管理、KV 与文件的完整操作、数据导出与清空、会话摘要与清除、版本记录、更新前预览与更新后再次打开、深浅色和手机布局。PR 写明浏览器/版本、真实与模拟环境、缺口受影响的未运行项；审查者按 TEST-MATRIX §3 记录 V-136 与 V-134/V-135/V-138 页面证据。

完成条件：npm --prefix web/admin ci、npm --prefix web/admin run typecheck、npm --prefix web/admin run test、npm --prefix web/admin run build 全部成功；Go 改动 gofmt -l 无输出，go vet ./web/admin/... ./scripts/dist/...、go test -race ./web/admin/... ./scripts/dist/...、go build ./... 成功。通过现有 CLI 对 web/admin 执行 check、build 与多目标 pack，核对五平台包包含进程入口与 dist；交叉编译按构建证据记录。真实 Core 上完成后台采集、认证路由与页面主路径验证，原目标中的契约缺口经审查者裁定并完成后再交付；PR「文档同步」列出 T-14.3、ARCHITECTURE 与 TEST-MATRIX 的实际同步项。

不做：Core 的事件、API 与插件进程的支持（T-09.20～T-09.22、T-09.24）；直读 Core 或其他组件数据库、代管管理 token；模型与配置文件页（goal-146、goal-165）；goal-44 的数据库表格、配置与插件页面桥；自动安装更新、自建商店服务器；改 spec、CI 前端任务或新增未批准依赖。
```

## 第 20 波

### goal-146 · T-09.18 模型页

状态：已完成（PR #244），合并提交 `a989b5f`。以下保留派发时语句。派发时：前置 T-09.8（PR #238，合并提交 `9123839`）、T-09.14（PR #192，`0cff77c`）、T-12.16（PR #223，`d27ce82`）均已合并。按 main `50244fe` 核对：`web/admin/src/api/admin.ts` 已导出 `listPackages(): Promise<Package[]>`、`changePackage(action: 'enable' | 'disable' | 'restart' | 'reload', id: string, confirm?: boolean): Promise<void>`、`AdminAPIError` 与 `isConfirmationError`，配置与模型请求尚无前端导出。官方 Agent 的实际包为 `github.com/AlonaBot/AlonaBot/agent`，`agent.New() *sdk.Plugin` 在 `agent/agent.go` 注册 `/models/discover`、`/models/test`、`/models/usage`；`agent/models.go` 与 ADMIN-API §15.1 一致。`agent/config_schema.json` 的 providers 为 template_list、models 为 list；ConfigForm 已支持两者。页面沿用 `src/routes.tsx` 的 RouteEntry 与 `src/main.tsx` 的认证外壳；AppShell、Sidebar、MobileTabBar 当前尚无按 Agent 安装状态禁用模型入口的属性。沿用 D-154、D-155 与 D-169 的模型页目标和画板。

```text
/goal 使用 $alonabot-execute 推进 goal-146（切片 T-09.18：模型页）。goal 目录 goal-146，分支 goal-146，worktree .worktrees/goal-146，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：侧栏一级入口「模型」：编辑 Agent 配置中的 providers 与 models，调用 Agent 的模型管理路由拉取模型列表、测试模型与读取当天用量。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（模型页）及其停止和失败条件；spec/plan/webui.md §3；CONTRACTS §5.1、§5.3、§13.5、§13.6；ADMIN-API §12、§15.1；DECISIONS D-154、D-155、D-169；TEST-MATRIX V-126。画板：「v4 · 模型」「v4 · 模型（添加提供商）」及其深色画板。审查者将截图放在 goal-146/figma/；首轮由前端执行者实现，审查与修复交 GPT。

范围（只新增或修改这些；路径相对 web/admin/）：
- src/pages/ModelsPage.tsx 及 src/pages/ 下模型列表、提供商对话框与配置 patch 的专用模块；src/api/admin.ts。
- src/routes.tsx、src/shell/AppShell.tsx、src/shell/Sidebar.tsx、src/shell/MobileTabBar.tsx：登记页面、按已确认的 Agent 安装状态禁用入口与搜索动作；src/pages/PluginsPage.tsx 仅补齐从 URL 打开商店标签的入口。
- src/style.css、i18n/en-US.json、i18n/zh-CN.json；对应的 *.local.test.ts。

允许的依赖：只用 web/admin/package.json 已安装的 React、Radix、lucide-react、TypeScript、Vite、vitest；复用现有 ConfigForm、schema/patch 与 i18n/catalog。

接口（新增导出写在指定模块；已由先合并切片加入的同名能力直接复用）：
- src/pages/ModelsPage.tsx：export function ModelsPage(): JSX.Element。routes 登记 {id: 'models', path: '/models', label: 'models', icon, group: 'main', render: () => <ModelsPage />}，main.tsx 沿用 matchedRoute.render()、UIContext 与 AppShell。label 及新增文案写入两份 i18n JSON 的 messages，useUI() 读取，键集合一致。站内导航复用 navigate(path, options?) 与 Link，页面操作复用 SearchContext.register，卸载时注销。
- AppShell 保留现有 version、onLogout、coreUpdate、children，在内部随 usePath() 的路径变化及 onPackagesChanged 通知通过 listPackages 的 Role === 'agent' && !Uninstalled 确认 Agent 安装状态。Sidebar 与 MobileTabBar 的现有 props 增加可选 agentInstalled?: boolean，由 AppShell 传入；明确为 false 时模型入口置灰、停止点击和键盘激活，搜索省略该路由。直接访问 /models 仍显示安装空状态。加载失败显示真实错误，保留尚未确认的状态；安装或卸载后重读并同步入口状态。
- src/api/admin.ts：
  export function onPackagesChanged(handler: () => void): () => void
  export type PendingConfig = { base_generation: string; config: Record<string, unknown> }
  export type ConfigSnapshot = { active_generation: string; active: Record<string, unknown>; pending: PendingConfig | null; schema: unknown; i18n: Record<string, Record<string, unknown>> }
  export function getPluginConfig(id: string): Promise<ConfigSnapshot>
  export function patchPluginConfig(id: string, baseGeneration: string, patch: Record<string, unknown>): Promise<PendingConfig>
  export type ModelDiscoveryInput = { provider: string } | { protocol: string; base_url: string; api_key?: string }
  export type DiscoveredModel = { id: string; type: string }
  export type ModelUsage = { model: string; input_tokens: number; output_tokens: number }
  export function discoverAgentModels(id: string, input: ModelDiscoveryInput): Promise<{ models: DiscoveredModel[]; description?: string }>
  export function testAgentModel(id: string, model: string): Promise<{ ok: boolean; duration_ms: number; description?: string }>
  export function getAgentModelUsage(id: string): Promise<{ models: ModelUsage[] }>
- onPackagesChanged 沿用 onSessionExpired 的订阅/注销方式，由现有 installPackage、updatePackage、uninstallPackage 在操作成功后通知；AppShell 在卸载时注销。先合并切片已有组件变更通知时直接复用，保持既有安装、确认和错误语义。
- 配置读取 → GET /api/v1/plugins/{id}/config；PATCH .../config/pending 发 {base_generation,patch}。保存以 pending.config 或 active 为编辑基线，提交 active_generation，成功显示 Pending 后显式 changePackage('reload', id)，重读快照；开发来源保存时已按 §12 重启、pending 为 null，按返回实际状态展示。409 / FAILED_PRECONDITION 属于代次冲突，重读并提示用户处理草稿；仅带 affected 的 409 进入现有确认流程。
- 模型路由前缀 /plugins/{编码后的 Agent 包 ID}/models/：discoverAgentModels → POST discover，传已配置 provider 名，或 protocol/base_url/api_key 新连接；testAgentModel → POST test，发 {model}（Active 模型配置名）；getAgentModelUsage → GET usage。复用 admin.ts 的私有 request、同源 Cookie、内存 X-Alonabot-CSRF 与 onSessionExpired。扩展错误解析以保留 Agent JSON 的 description，同时保留 Core 的 safe_message、code 与纯文本认证错误；已有客户端调用语义保持。
- src/components/ConfigForm.tsx 的现有接口为 ConfigForm(props: { schema: unknown; value: unknown; messages: Record<string, unknown>; onSubmit: (patch: Record<string, unknown>) => void })。提供商编辑与无模型路由时的配置编辑复用它，只传 providers/models 对应的 schema 与值。getPluginConfig.i18n 每个语言值已是 config 根节点，按 pickLocale 选择后包装成 {config: selectedMessages} 传入；页面的 useUI 文案与插件字段文案分别处理。
- providers 每行为 {模板名: {name,protocol,base_url,api_key}}；预设与协议取返回 schema，不手写一套重复表单。models 每行为 {name,provider,model,type,max_tokens?,image_input?}：name 是配置引用名，provider 引用提供商 name，model 才是远端 ID。type 的协议值为 chat、speech_to_text、text_to_speech、embedding、rerank、decision。

行为：
- 按 `role: agent` 找到已安装的 Agent；没有时显示空状态与去商店的按钮，侧栏的「模型」置灰；Agent 没有模型管理路由时只提供配置编辑。
- 按类型筛选（对话、语音识别、语音合成、向量、重排、决策），按提供商分组列出模型，显示模型 ID、类型、当天用量与测试按钮；预留的五类只能添加、拉取与测试。
- 添加提供商对话框：预设、名称、Base URL、API Key（secret，只显示已设置），「拉取模型列表」后勾选要加入的模型，拉取失败时可以手动输入模型 ID。
- 修改经配置 API 的 Pending 与 reload（只提交 patch）；删除被 chat_model 或 fallback_chat_models 引用的模型时先提示引用位置。
- 拉取与测试针对 Active 配置：已有提供商按 name 拉取，新提供商可用对话框中的新连接直接 discover；Pending 模型先保存并 reload 后测试。discover 返回 models: [] 与 description 时展示原说明并允许手填；HTTP 200 的 ok: false 展示探测失败。usage 按 Agent 本地日期，只有实际用量记录的模型显示数值；缺失与读取失败分别显示状态，保留真实的零 token。
- 使用中的 API Key 只显示 $secret_set 状态。ConfigForm/buildPatch 对含 secret 的 providers 数组执行现有整体替换规则：提交前重新输入该数组仍为掩码的全部 secret；提交成功清空明文草稿。发现、测试、保存均不把掩码作为 api_key 发送。其余配置、未知键、invisible 字段及 chat_model/fallback_chat_models 保持原值；删除引用中的模型保留提示及候选失败的实际原因，等待用户修改引用。
- Agent 已安装但模型管理路由缺失时保留配置编辑，移除拉取、测试与用量操作；用 usage 的路由级 404 判断，单个模型/提供商的 404、上游 502、认证失败与离线错误分别展示原原因。没有 Agent 时「去商店」用 navigate('/components?tab=store') 打开已有商店标签，PluginsPage 从 URL 的 tab 初始化该标签；商店内容由 goal-156 接入并保留此入口。

测试（按编码规范「测试提交规则」）：
- 提交：无。
- 本地自测：*.local.test.ts 覆盖六类筛选与提供商分组、配置名和远端 ID 的区别、模板行与数组 patch、secret 掩码与重新输入、模型引用提示、Pending/Active、陈旧代次冲突、description/ok:false/真实零用量、路由缺失与普通 404 的区别。
- 浏览器手动验证：真实 Core 与官方 Agent 上添加提供商、拉取并勾选模型、失败后手填、更换 API Key、保存与 reload、测试模型、当天用量、删除被引用模型、无 Agent 与缺模型路由；深浅色、手机入口、置灰后的键盘操作与搜索。PR 写明浏览器与版本、实际 Provider 和六类中已验证的类型；本地模拟 Provider 记录为模拟环境，其余保持未运行。

完成条件：npm --prefix web/admin ci、npm --prefix web/admin run typecheck、npm --prefix web/admin run test、npm --prefix web/admin run build 全部成功；go build ./... 不受影响，PR 写明实际验证结果。审查者按 TEST-MATRIX §3 记录 V-126 页面证据，并在合并后按实际导出修订 goal-165 的模型选择接入。

不做：Agent 的模型管理路由或 Provider 实现（T-12.16）；配置文件中的模型选择（goal-165）；Go 代码、spec 修改、新依赖、DOM 测试依赖、CI 前端任务与提交前端测试。
```

## 第 22 波

### goal-165 · T-09.25 配置文件页

状态：待派发。前置 T-04.19（PR #250，合并提交 `f3a06458`）、T-11.6（PR #246，`24e57d60`）、T-12.19（PR #247，`f3959f6e`）、T-09.18（PR #244，`a989b5f1`）均已合并；配置文件管理 API 的 T-08.18（PR #214，`ab0cd565`）已合并。按 main `ce68e109` 核对页面基线；工具候选沿 2026-11-19 的 D-168 实施细化，按 §18 `identities` 展开为独立工具与确认操作，分别排除本配置文件已有覆盖。派发前先合入目录身份补齐；goal-44 先合并，本切片合并时复用其插件配置面板，冲突由后合并者按共享规则处理。

- `web/admin/src/api/admin.ts` 已导出 `listProfiles(): Promise<Profile[]>`、`getProfileAssignments(): Promise<ProfileAssignments>`、`assignInstanceProfile(instanceID: string, profile: string): Promise<void>`、`listInstances(): Promise<AdapterInstance[]>`、`listChats(): Promise<Chat[]>`，后者读取全部聊天目录；配置文件的其他写入、插件覆盖、平台与工具策略尚无前端导出。已有 `getPluginConfig(id: string): Promise<ConfigSnapshot>`、`patchPluginConfig(id: string, baseGeneration: string, patch: Record<string, unknown>): Promise<PendingConfig>`、`changePackage(action: 'enable' | 'disable' | 'restart' | 'reload', id: string, confirm?: boolean): Promise<void>`、`AdminAPIError` 与 `isConfirmationError`。
- `src/pages/ModelsPage.tsx` 导出 `ModelsPage(): JSX.Element`；`src/pages/ModelSelect.tsx` 导出 `ModelSelect({value,label,options,disabled,onChange})`，参数见下文；`src/pages/modelConfig.ts` 已有 `parseModels(config: Record<string, unknown>): ConfiguredModel[]`、`editingConfig(snapshot: ConfigSnapshot): Record<string, unknown>`。`src/components/ConfigForm.tsx` 的四个必填 props 为 schema、value、messages、onSubmit，另有 `footer?: ReactNode`；`src/schema/model.ts` 的 `fieldsFromSchema(schema: unknown): Field[]` 尚未接受 `profile` 与 `options_from`。
- `src/routes.tsx` 的 `RouteEntry` 与 `src/main.tsx` 的 `matchedRoute.render()`、`/components/{id}` 分派可复用。`AppShell({version,onLogout?,coreUpdate?,coreUpdateFailed?,children})` 已装配导航与搜索，`MobileTabBar` 已预留路由 ID `profiles`。`PluginDetailPage(props: { id: string; generations?: readonly Generation[] }): JSX.Element` 目前有版本、高级、数据标签；main 尚无 `PluginConfigPanel.tsx`、`SettingsPage.tsx`。本切片补齐配置文件切换所需的配置标签，goal-44 合并时复用同一入口，系统设置保持无命令前缀。
- 后台实际包 `github.com/AlonaBot/AlonaBot/internal/core/profile` 的 `New(ctx context.Context, opt Options) (*Service, error)`、`(*Service).EffectivePlatform(id string) (prefixes []string, reply settings.Reply, err error)`、`(*Service).ToolPolicy(id, pluginID, name string, operation bool) (policy pb.ToolPolicy, enabled bool, err error)` 已实现；`github.com/AlonaBot/AlonaBot/internal/core/manage` 的 `(*Service).Handler(auth *adminauth.Auth) http.Handler` 已注册 ADMIN-API §23。官方 Agent 实际包为 `github.com/AlonaBot/AlonaBot/agent`，`agent.New() *sdk.Plugin`；`agent/config_schema.json` 已标记 `chat_model`、`image_caption_model`、`default_persona`、`triggers`、`wake_words`、`member_sessions`，模型目录为基础配置 `models`，人格目录目前为基础配置 `personas`。

```text
/goal 使用 $alonabot-execute 推进 goal-165（切片 T-09.25：配置文件页）。goal 目录 goal-165，分支 goal-165，worktree .worktrees/goal-165，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：侧栏「配置文件」打开即是当前配置文件的设置（只有一层），分区为 AI、平台与消息与插件；切换、新建、使用范围、重命名、复制与删除；插件配置页可切换配置文件。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（配置文件页）及其停止和失败条件；spec/plan/webui.md §2、§3；CONTRACTS §5.1、§5.3、§5.6；ADMIN-API §12、§17、§18、§20、§23；DECISIONS D-167、D-168、D-169；TEST-MATRIX V-145。画板：审查者提供在 goal-165/figma/ 的「v4 · 配置文件」「配置文件（深色）」「配置文件（插件 · VIP 群）」「配置文件（没有 Agent）」「配置文件（使用范围）」及深色、「配置文件（平台与消息）」及深色、「配置文件（切换配置文件）」「手机 · 配置文件」。首轮由前端执行者实现，审查与修复交 GPT；画板缺失时按本章共同约定向审查者索取。

范围（只新增或修改这些；路径相对 web/admin/）：
- 新增 src/pages/ProfilesPage.tsx、ProfileAssignmentsDialog.tsx、ProfileAISection.tsx、ProfilePlatformSection.tsx、ProfilePluginsSection.tsx、ProfileToolsPanel.tsx、PluginConfigPanel.tsx，以及 src/pages/ 下这些模块专用的对话框与纯函数模块。
- src/pages/PluginDetailPage.tsx、src/api/admin.ts、src/routes.tsx；src/components/ConfigForm.tsx、src/schema/model.ts：补齐配置文件编辑所需的字段元数据与只读路径。src/main.tsx 只在页面挂载确有需要时修改，保留认证与组件详情分派。
- src/pages/SettingsPage.tsx 仅在先合并切片已提供该页时移除命令前缀控件；当前 main 无该文件，本切片沿用配置文件页的平台入口。
- src/style.css、i18n/en-US.json、i18n/zh-CN.json；上述模块对应的 *.local.test.ts。goal-44 若先合并，复用其 PluginConfigPanel、schema 元数据与选项查询，保留已有标签、表单 props 和配置操作。

允许的依赖：只用 web/admin/package.json 已安装的 React、Radix Dialog/DropdownMenu/Popover/Tooltip、lucide-react、TypeScript、Vite、vitest 与浏览器 API；复用 ConfigForm、schema/patch、i18n/catalog、ModelSelect、modelConfig、Pagination、shell/router、SearchContext 与现有主题令牌。

接口（新增导出照此实现；先合并切片已提供的同名能力直接复用，保留其现有字段与调用语义）：
- src/pages/ProfilesPage.tsx：export function ProfilesPage(): JSX.Element。routes 登记 {id:'profiles',path:'/profiles',label:'profiles',icon,group:'main',render:() => <ProfilesPage />}；main.tsx 沿用 matchedRoute.render()。AppShell、Sidebar 与 MobileTabBar 从 routes 自动接入桌面入口、手机固定标签与页面搜索，配置文件入口在没有 Agent 时仍可用。当前选择用 /profiles?profile={配置文件ID} 表示，首次进入选择 default；用 navigate 更新并响应 popstate，刷新保留选择。usePath() 只返回 pathname，查询参数从 URLSearchParams 读取，组件内同时响应同路径的切换。
- 页面专用组件：
  export function ProfileAssignmentsDialog(props: { profile: Profile; open: boolean; onOpenChange: (open: boolean) => void; onChanged: () => void }): JSX.Element
  export function ProfileAISection(props: { profileID: string; agent: Package; plugin: ProfilePlugin; onChanged: () => void }): JSX.Element
  export function ProfilePlatformSection(props: { profileID: string }): JSX.Element
  export function ProfilePluginsSection(props: { profileID: string; packages: Package[]; plugins: ProfilePlugin[]; onChanged: () => void }): JSX.Element
  export function ProfileToolsPanel(props: { profileID: string }): JSX.Element
  上述组件各在同名 .tsx 中导出。新建、重命名、复制、删除确认用页面内的 Radix 对话框；内部辅助函数自定。
- src/pages/PluginConfigPanel.tsx：export function PluginConfigPanel(props: { id: string }): JSX.Element。在 PluginDetailPage 增加配置标签并挂载 <PluginConfigPanel id={id} />，保留 PluginDetailPage 原签名、版本、高级、数据及其操作。插件行跳转 navigate('/components/' + encodeURIComponent(pluginID) + '?tab=config&profile=' + encodeURIComponent(profileID))；详情读取 tab=config 后选中配置标签，面板读取 profile 并提供同一配置文件选择器，默认 default，切换与浏览器前进后退同步。goal-44 已提供面板时直接扩展原面板。
- src/api/admin.ts 复用 Profile = {id:string;name:string;color:string;sessions:number}、ProfileAssignments = {instances:Record<string,string>;sessions:{instance:string;session:string;profile:string}[]}、listProfiles、getProfileAssignments、listInstances、listChats、listPackages、onPackagesChanged。assignInstanceProfile 的 profile 参数扩为 string | null，原 string 调用保持。新增：
  export function assignInstanceProfile(instanceID: string, profile: string | null): Promise<void>
  export function createProfile(input: { name: string; color: string; copy_from?: string }): Promise<{ id: string }>
  export function updateProfile(id: string, input: { name?: string; color?: string }): Promise<void>
  export function deleteProfile(id: string): Promise<void>
  export function assignSessionProfile(input: { instance: string; session: string; profile: string | null }): Promise<void>
  // 依次调用 §23 的 POST /api/v1/profiles、PATCH/DELETE /api/v1/profiles/{id}、PUT /api/v1/profiles/assignments/sessions；assignInstanceProfile 沿用 PUT .../assignments/instances/{instance}，发 {profile}，null 取消指定。
  export type ProfilePlugin = { plugin_id: string; enabled: boolean; fields: string[]; overrides: Record<string, unknown>; invalid: { path: string; reason: string }[] }
  export function getProfilePlugins(id: string): Promise<ProfilePlugin[]>
  export function patchProfilePlugin(id: string, pluginID: string, patch: { enabled?: boolean; overrides?: Record<string, unknown> }): Promise<void>
  export function pruneProfilePlugin(id: string, pluginID: string): Promise<{ removed: string[] }>
  // GET .../profiles/{id}/plugins、PATCH .../plugins/{plugin_id}、POST .../plugins/{plugin_id}/prune；overrides 的键是完整点分路径，值 null 删除该项。
  export type ProfilePlatform = { command_prefixes: string[] | null; reply: { mention_sender: boolean | null; quote: boolean | null; prefix: string | null; forward_threshold: number | null } }
  export function getProfilePlatform(id: string): Promise<ProfilePlatform>
  export function putProfilePlatform(id: string, platform: ProfilePlatform): Promise<void>
  // GET/PUT .../profiles/{id}/platform；PUT 发送完整当前覆盖对象。default 的所有键非 null。
  export type ProfileToolPolicy = ({ tool: string; operation?: never } | { operation: { plugin_id: string; name: string }; tool?: never }) & { policy: 'allow' | 'confirm' | 'deny' | 'inherit' }
  export function getProfileTools(id: string): Promise<{ items: ProfileToolPolicy[] }>
  export function putProfileTools(id: string, items: ProfileToolPolicy[]): Promise<void>
  export type ToolPolicy = { plugin_id: string; service: string; identities: ('tool' | 'operation')[]; default: ToolPolicyValue; override: ToolPolicyValue; effect: 'read_only' | 'write' | 'destructive' | 'unspecified' }
  export function getToolPolicies(): Promise<{ policies: ToolPolicy[] }>
  // GET/PUT .../profiles/{id}/tools，PUT 发 {items}；getToolPolicies 只读 §18 GET /api/v1/tools/policies，供「覆盖其他工具」查询全局目录与策略。
- 请求复用 admin.ts 的私有 request、现有 jsonBody、同源 Cookie、内存 CSRF 与 onSessionExpired；所有路径 ID 按一个 URL 段编码，保留 AdminAPIError 的 status、code、affected、responseBody 与原始错误。以上成功 {} 的写入由客户端归一为 Promise<void>，不改变既有请求导出。
- 复用 getPluginConfig、patchPluginConfig 与 ConfigSnapshot = {active_generation:string;active:Record<string,unknown>;pending:PendingConfig|null;schema:unknown;i18n:Record<string,Record<string,unknown>>}、PendingConfig = {base_generation:string;config:Record<string,unknown>}。default 以 editingConfig(snapshot) 为编辑基线，提交 snapshot.active_generation，保存后重读并显示 Pending；确认重载前列出将重载的插件，使用 changePackage('reload', id)。开发来源保存后按重读到的 pending:null 展示。保留 409 / FAILED_PRECONDITION 冲突、带 affected 的确认、候选失败与 Degraded 的真实反馈。
- src/pages/ModelSelect.tsx 复用 ModelSelect(props: { value: string; label: string; options: { value: string; label: string }[]; disabled?: boolean; onChange: (value: string) => void })；src/pages/modelConfig.ts 复用 parseModels(config: Record<string, unknown>): ConfiguredModel[] 与 editingConfig(snapshot: ConfigSnapshot): Record<string, unknown>。配置文件中的模型引用保存模型配置 name，候选取 parseModels(snapshot.active)，对话模型只列 chat，图像转述只列 chat 且 image_input:true，保留空字符串。人格候选取当前 Active 的 personas[].name；人格迁移若已合并，沿 ADMIN-API §17 的字段选项入口读取，不复制人格或模型管理功能。
- src/schema/model.ts 保持 fieldsFromSchema(schema: unknown): Field[]；Field 增加 profile:boolean（缺省 false）、optionsFrom?: OptionsFrom，新增 export type OptionsFrom = { source: string; filter?: Record<string, string> }。按 CONTRACTS §5.1 接受 profile 与 options_from，保留字段顺序、schemaPath、其他元数据及未知键报错；标记位置与复合 secret 的语义遵循 §5.6，已有复合标记仍可读取。
- src/components/ConfigForm.tsx 保留 schema:unknown、value:unknown、messages:Record<string,unknown>、onSubmit:(patch:Record<string,unknown>) => void、footer?:ReactNode，增加 editablePaths?: readonly string[] 与 options?: { list: (path: string, q: string) => Promise<OptionsPage>; resolve: (path: string, values: string[]) => Promise<OptionsPage> }。editablePaths 省略时保持原编辑行为；提供时仅这些点分路径及其被整体标记子树可编辑，祖先 object 继续作为分组，其余字段显示为禁用，恢复默认与增删行也遵循同一路径限制，提交只含可编辑路径的变更。onSubmit 继续返回 RFC 7396 patch，§23 的转换由面板负责。保留 invisible、condition、secret、数组与模板的现有处理及 footer；前端字段文案用 pickLocale/fieldText/templateText，getPluginConfig.i18n 选中语言后包装成 {config:selectedMessages}。
- options_from 所需请求补在 src/api/admin.ts：
  export type ConfigOption = { value: string; label: string; description?: string }
  export type OptionsPage = { options: ConfigOption[]; more: boolean }
  export function listConfigOptions(id: string, path: string, q: string, locale: string): Promise<OptionsPage>
  export function resolveConfigOptions(id: string, path: string, values: string[], locale: string): Promise<OptionsPage>
  // §17 GET /api/v1/plugins/{id}/config/options?path=&q=、POST .../options/resolve 发 {path,values}，带 Accept-Language。path 按字段值的点分路径，来源与 filter 由服务端 schema 决定。复用 goal-44 已有同名接口与可搜索控件；当前 main 未提供时只补这两个插件字段入口，保留 more、已存值 resolve、缺失值原文和来源不可用的实际错误。
- 界面用 useUI()，文案加入双语 messages，键集合一致；页面动作通过 SearchContext.register(items: SearchItem[]): () => void 登记并在卸载时注销。长列表复用 Pagination({page,total,size,onChange}) 或按本章约定筛选、分组折叠；颜色使用既有令牌。

行为：
- 标题「● 名称 ⌄」打开切换菜单（各配置文件的名称与会话数、新建配置文件）；旁边的「N 个会话」打开使用范围对话框（各账号作为缺省的开关，旁边写当前缺省；单独指定的会话可添加与移出）；⋯ 菜单为重命名、复制与删除，「默认」的删除置灰。页面没有返回按钮。
- 新建与复制调用 createProfile，复制传 copy_from，成功选择返回的 ID；复制只包含覆盖、插件开关、工具策略与平台设置。名称与颜色遵循 §5.6；名称重复的 409 显示原原因。删除先确认影响，成功选择 default 并重读列表、会话数与指定；default 的删除在鼠标与键盘操作中均置灰。
- 使用范围按 getProfileAssignments、listInstances 与 listChats 的真实数据展示；会话身份使用实例 ID 与 API 返回的 session，复用 accountState.ts 的 chatProfile(assignments: ProfileAssignments, instanceID: string, session: string): string 判断单独指定、实例缺省、default。实例开关开启时写当前 profile ID，关闭时写 null；单独指定可添加目录中的会话或完整会话字符串，移出写 profile:null。显式指定 default 与取消指定分别按 §23 保留含义。对话框区分当前配置文件的单独指定与继承使用，修改后重读指定与服务端会话数；筛选或分页只影响展示。
- AI 分区：已安装 Agent 的 schema 中写了 `profile` 的字段（D-168）——对话模型、图像转述、人格、触发方式（多选：@ 或回复机器人、唤醒词、私聊消息）、选中唤醒词时显示的唤醒词列表、群成员各自独立对话——以及按配置文件覆盖的工具策略（只列覆盖了全局的工具，「+ 覆盖其他工具」）。没有 Agent 时不显示 AI 分区，在分区上方提示并给出去商店的按钮。
- Agent 从 listPackages 的 Role === 'agent' && !Uninstalled 取得，使用实际包 ID；订阅 onPackagesChanged 并在卸载时注销，安装、卸载后重读。去商店沿用 navigate('/components?tab=store')。六个字段的协议名依次为 chat_model、image_caption_model、default_persona、triggers、wake_words、member_sessions；triggers 保存 mention、wake_word、private，空数组表示全部关闭。隐藏唤醒词列表时保留已有值。模型与人格选择沿用当前 Active 目录，原引用缺失时保留存储值并显示实际状态；修改 providers、models 与人格目录继续使用原页面。
- 平台与消息分区：命令前缀（可增删的标签），回复时 @ 发送人、引用原消息、回复前缀与长回复合并转发（标注只在支持的平台生效）。
- 平台设置分别读取当前配置文件与 default，区分已保存覆盖和继承后的展示值；其他配置文件每个键可跟随默认，保存 null，显式 false、空字符串与 0 保留。PUT 整体替换时带回未编辑键的原覆盖状态；default 始终提交完整非 null 设置。command_prefixes 至少一项，空字符串项表示无前缀命令；forward_threshold 为非负整数，0 关闭，正数按 Unicode 码点阈值、仅在支持 message.forward 的平台生效。mention_sender 只在群聊第一条回复生效，quote 与 prefix 作用于第一条回复。保存后重读，后续派发与 Send 立即使用新设置。
- 插件分区：每个插件一行，开关与「覆盖 N 项：字段名」或「跟随默认」，箭头进入该插件的配置页并预先选中这个配置文件；全局已停用的插件置灰。
- 列表以 getProfilePlugins 为准，与 Package.Enabled 合并展示全局状态；配置文件开关只调用 patchProfilePlugin({enabled})，保持全局启停状态。字段名称从对应 schema/i18n 取得，覆盖摘要与 invalid 分别展示；Adapter 与 WebUI 沿用各自管理页。
- 「默认」的字段修改走插件配置 API（Pending 与 reload，保存时提示会重载的插件）；其他配置文件经 ADMIN-API §23 立即生效。插件配置页选中「默认」以外的配置文件时，只有写了 `profile` 的字段可以编辑，其余字段置灰。无效覆盖值显示原因并可一键清理。
- 其他配置文件以 snapshot.active 为基础，按 ProfilePlugin.fields 中的点分路径应用有效 overrides；invalid 中的路径保留诊断并排除在生效预览外。覆盖按完整字段值替换，object/list 也整体替换，保留基础值的其他字段。ConfigForm 的 RFC 7396 patch 先应用到当前有效草稿，再取本次变更涉及的标记路径的完整值，形成 §23 的 overrides；恢复跟随默认写该路径:null。标记 object 内部的修改提交完整 object，保留未编辑成员；无关标记路径维持原覆盖状态。具有 secret 子树的复合标记按契约只读并展示实际校验原因，失效项用 prune 清理；禁止把掩码或未标记字段送入覆盖接口。
- default 的普通配置表单继续使用现有 buildPatch，数组、secret、invisible 与 condition 规则保持；插件配置页的非 default 草稿与 Pending 分开展示，切换后重建对应编辑基线。字段恢复默认在 default 中沿用配置 patch 的 null，在其他配置文件中表示跟随 default。prune 只调用 §23 的覆盖清理接口，成功重读 overrides/invalid；保留 §12 的普通配置操作。
- 工具区默认只列 getProfileTools 的覆盖项，可筛选全局目录来添加覆盖；allow、confirm、deny 分别保存对应策略，恢复全局发 inherit。PUT 只发本次更新项；已读工具保留 tool，已读确认操作保留 operation，包含同名项时分别编辑。插件关闭时展示其工具被禁用的实际含义，保留已存策略；默认配置文件的工具策略仍经 §23 管理。§18 的目录按 plugin_id/service 合并，按 identities 展开候选：tool 写入完整工具名，operation 写入 plugin_id/name；同名两类分别排除已有覆盖，[] 没有新增候选。已保存覆盖在目录身份消失后仍照原身份准确读写。
- 系统设置沿 PLAN 保持无命令前缀，平台设置统一在配置文件页。页面只调用上述 ADMIN-API 已有端点，错误显示完整原因；覆盖通知失败可能发生在保存之后，写入报错后重读已保存状态并保留错误，依服务端实际结果更新界面。

测试（按编码规范「测试提交规则」）：
- 提交：无。
- 本地自测：前端纯函数与请求格式写在 *.local.test.ts，覆盖配置文件切换与同路径查询参数、会话指定优先级与 null 取消、复制请求、平台继承与 false/空字符串/0、完整 PUT 保留未改键、profile/options_from 解析、只读路径、点分覆盖和 object/list 整体替换、恢复继承、invalid 排除与 prune、default 的 Pending/代次冲突、模型配置名与 Active 候选过滤、工具与 operation 的身份及 inherit。其余临时 Go 验证按共同约束写在被忽略的 *_local_test.go，PR 记录结果。
- 浏览器手动验证：同源真实 Core 与官方 Agent 上打开、切换、新建、重命名、复制与删除，default 删除置灰；使用范围中实例缺省、单独指定、取消指定和真实会话数；AI 六字段、唤醒词显隐与没有 Agent；平台五个键的覆盖和跟随默认；插件开关、箭头预选、非标记字段禁用、无效覆盖清理；default 保存提示、Pending 与 reload，其他配置文件修改后的命令前缀、回复与 Agent/工具策略；工具与操作的独立候选、同名分别覆盖、空身份行及目录身份消失后的原覆盖读写；深浅色、手机、键盘与搜索。PR 写明浏览器和版本、真实或模拟环境、用例名及未运行项。

完成条件：npm --prefix web/admin run typecheck、npm --prefix web/admin run test、npm --prefix web/admin run build 全部成功；按本章共同约定准备依赖并记录真实 Core 浏览器验证，完成 identities 驱动的独立工具与操作候选及原覆盖读写。PR 按 TEST-MATRIX §3 提供 V-145 页面证据与未运行项，审查者据实际结果同步；与 goal-44 的配置标签、表单和选项查询共用同一实现。

不做：配置文件的 Core、SDK 与 Agent 实现（T-08.18、T-04.18、T-04.19、T-11.6、T-12.19）；会话页的配置文件列（goal-44）；模型管理与人格管理、goal-44 的其余配置增强和系统设置实现；Go 产品代码、spec 修改、新 Admin API、新依赖、DOM 测试依赖、CI 前端任务与提交前端测试。
```

### goal-175 · S-06.3 休眠策略与运行状态的页面

状态：等待前置（T-09.9、S-06.1）。T-09.8（PR #238，`9123839`）已合并。2026-11-17 维护者决定把插件休眠（PLAN §7）纳入本轮时新增：休眠控件从 goal-43、goal-44 移到本切片，使 WebUI 主线与发布链不依赖休眠切片（S 工作包不进入 v0.1 发布门槛）。目标、依据与要点由 PLAN §7 S-06 步骤 1、ADMIN-API §24 与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-175（切片 S-06.3：休眠策略与运行状态的页面）。goal 目录 goal-175，分支 goal-175，worktree .worktrees/goal-175，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：在 goal-43 的插件详情「高级」标签加入休眠策略（跟随全局、永不休眠或时长；manifest 要求常驻或其他原因不能休眠时锁定并显示原因与提示），在 goal-44 的系统设置加入全局运行模式与「插件空闲休眠」时长；插件列表与详情显示运行状态（含「已休眠」）与最近的唤醒/休眠原因。本切片不改 Go 代码。

依据：PLAN §7 S-06 步骤 1；CONTRACTS §4.4、§4.5、§15.1；ADMIN-API §24；DECISIONS D-153；TEST-MATRIX V-81 的页面分项。画板：「v4 · 插件 · 天气 · 高级」「v4 · 插件 · 天气 · 高级（不能改休眠）」「v4 · 系统设置」。

要点：
- 只调用 ADMIN-API §24；读取不唤醒插件。保存后唤醒失败时显示实际错误并重新读取当前设置。
- 原因与警告的文案按 §24 的稳定值在 i18n 中翻译；未声明 Core 过滤的消息事件、Raw、HTTP、Hook 或持续监听的插件被显式给出时长时，提示会被频繁唤醒。

不做：休眠的 Core 实现与管理 API（S-01.1～S-06.1）；固定机器测量（S-06.2）。
```
