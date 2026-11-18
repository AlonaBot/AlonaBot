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
| goal-43 | T-09.8 | 登录、总览、插件、账号与版本 | T-08.7、T-09.6、T-09.12、T-09.13、T-09.26 | 待派发 |
| goal-44 | T-09.9 | 插件详情、权限、会话与系统设置 | T-09.8、T-09.14、T-11.5、T-04.17、T-08.15、T-08.16 | 等待前置 |
| goal-94 | T-09.15 | 日志页 | T-09.8、T-09.10、T-09.19、T-02.5 | 等待前置 |
| goal-156 | T-09.23 | 官方 WebUI 的后台进程、统计、商店、用户与数据页 | T-09.8、T-09.20、T-09.21、T-09.22、T-09.24 | 等待前置 |
| goal-146 | T-09.18 | 模型页 | T-09.8、T-09.14、T-12.16 | 等待前置 |
| goal-165 | T-09.25 | 配置文件页 | T-04.19、T-11.6、T-12.19、T-09.18 | 等待前置 |
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

状态：待派发。按 `spec/plan/webui.md`、`spec/plan/core.md` 与 main 核对：T-08.7（PR #91，`9789fc8`）提供 `ConfigForm` 和双语文案；T-09.6（PR #179，`8898251`）接通更新的候选与提交；T-09.12（PR #167，`3a9b4af`）提供静态单页回退与组件图标；T-09.13（PR #181，`d507e5a`）提供会话取回、multipart 安装/更新及 Core 总览；T-09.26（PR #217，`acd5ba0`）提供路由、外壳、搜索和主题，签名见 goal-166。

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

状态：等待前置（T-09.8、T-09.14、T-08.16）。T-11.5（PR #219，`b2348c9`）、T-04.17（PR #218，`6d81be3`）、T-08.15（PR #211，`3b2b64c`）已合并。2026-11-17 按 D-169 以 Figma v4 改写，并纳入 D-157（工具类别）、D-159（插件页面主题）、D-160（回复示例）、D-161（配置校正）与 D-128～D-130 的页面部分。目标、依据与要点由 PLAN T-09 步骤 7 与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-44（切片 T-09.9：插件详情、权限、会话与系统设置）。goal 目录 goal-44，分支 goal-44，worktree .worktrees/goal-44，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：补齐插件详情页的配置、插件页面、定时任务与数据（数据库视图）标签，权限页（命令、工具调用、用户），会话页与系统设置页。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §2.3、§5、§7.3、§8、§13、§14.3；ADMIN-API §7、§12、§13、§15～§18、§20；DECISIONS D-45、D-128、D-129、D-130、D-153、D-157、D-159、D-160、D-161、D-169；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56、V-115、V-116、V-117、V-128、V-130、V-131、V-132 的页面分项。画板：「v4 · 插件 · 天气」「v4 · 插件 · 天气（切换配置文件）」「v4 · 插件 · 天气 · 定时任务」「v4 · 插件 · 天气 · 数据」「v4 · 插件 · 天气 · 高级（不能改休眠）」「v4 · 插件 · Agent · 用户任务（插件页面）」「v4 · 权限 · 命令」「v4 · 权限 · 命令（编辑命令名）」「v4 · 权限 · 工具调用」「v4 · 权限 · 用户」「v4 · 权限 · 用户（添加用户）」「v4 · 会话」「v4 · 系统设置」「v4 · 系统设置（更换管理令牌）」及其深色画板。

要点：
- 配置标签：goal-34 的 ConfigForm 编辑（带 base_generation，只提交 patch），每行左边是名称与说明、右边是控件；字段有 `example` 时在下方显示回复预览（D-160：文字与图片片段、占位符替换为草稿值、按条件显示）；带 `options_from` 的字段为可搜索选择，取不到的值标「已不存在」（D-129）；有校正记录时显示提示并提供「清理无效配置」（D-161）。有改动时底部浮出「放弃 / 保存」，保存后显示 Pending 与重载；409 时提示已被他处修改并重新读取。不提供完整 JSON 视图。
- 插件页面：插件声明的页面作为详情页的标签（标题取 ADMIN-API 的 pages），在沙箱 iframe 中打开；postMessage 桥按 ADMIN-API §7 实现 hello、请求、文案、主题与语言事件、⌘K 转发（D-84、D-159），页面拿不到会话、CSRF token 与其他插件的路由。
- 定时任务：插件代码中的任务只有启用开关（下次启动会重新注册，不提供删除），显示触发规则、上一次与下一次时间。
- 数据标签的数据库视图：按受管 SQLite 的表浏览，可筛选、删除与新增行，可切换保留代次的副本（ADMIN-API §21）；KV、数据目录文件与导出、清空由 goal-156 的「存储」视图实现。
- 权限 · 命令：按插件分组折叠的命令树，三层覆盖与有效权限（页面不自行推导），遮蔽关系；命令名不加「/」显示，可以替换、追加、删除与调整主名，只剩一项时不能删除，409 时显示冲突的命令（D-128）。
- 权限 · 工具调用：按插件分组，显示工具的类别（只读、写入、破坏性、未声明，D-157），策略为允许、询问、禁止，支持按类别批量设置。
- 权限 · 用户：一张表包含管理员与禁止（管理员是身份，禁止是状态），添加用户对话框可从聊天目录选择（D-130）。
- 会话：按插件列出会话摘要，可清除，不显示条目内容；有配置文件列与筛选（D-168，配置文件页由 goal-165 实现）。
- 系统设置：桌面后台样式（左边名称与说明，右边控件）；界面语言、外观（浅色、深色、跟随系统，使用 goal-166 的 useTheme）、管理令牌（更换令牌对话框，ADMIN-API §20）、网络代理、全局日志级别（含 CRITICAL）、日志保留；没有命令前缀（D-168）；右上角有「重启核心」与「登出」。
- secret 明文查看要求重新输入管理 token，只在当前视图显示。

不做：管理 API（T-09.14 等）；日志页（goal-94）；商店、统计、用户任务以外的插件页面内容、存储视图（goal-156）；模型与配置文件页（goal-146、goal-165）。
```

### goal-94 · T-09.15 日志页

状态：等待前置（T-09.8、T-09.10）。T-09.19（PR #213，`6f6856c`）、T-02.5（PR #209，`d75e43d`）已合并。2026-11-17 按 D-169 以 Figma v4 改写，纳入 D-156（诊断包）与 D-158（CRITICAL）；日志页不再设置级别（级别在系统设置与插件「高级」中，见 goal-44 与 goal-43）。目标、依据与要点由 PLAN T-09 步骤 7 与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-94（切片 T-09.15：日志页）。goal 目录 goal-94，分支 goal-94，worktree .worktrees/goal-94，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：实现日志页：终端视图实时显示日志，按级别、来源与日期筛选，搜索，清空已载入的记录，导出诊断包。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（日志、诊断包）；CONTRACTS §14.1、§14.3、§14.4、§14.7；ADMIN-API §8、§19；DECISIONS D-156、D-158、D-169；TEST-MATRIX V-62、V-63、V-127、V-129 的页面分项。画板：「v4 · 日志」「v4 · 日志（筛选）」「v4 · 日志（导出诊断包）」「v4 · 日志（深色）」「v4 · 手机 · 日志」。

要点：
- 只有终端视图：终端区域用 --color-console-* 与 --font-mono，每条记录一行，列定宽对齐：时间、来源、级别、源码位置（取 Record 的 source 中 file 的最后两段与 line，悬停显示 function 与完整路径，点击复制）、消息；消息过长时换行并对齐到消息列；方括号内不补空格。ERROR 整行红色，WARNING 整行黄色（--color-yellow，不用棕色或琥珀色），CRITICAL 整行紫色，DEBUG 用暗色；不用整行色块，不加柱状分布图、字号按钮或列表与终端切换。
- 工具栏：「筛选」按钮打开弹层（级别多选含 DEBUG 与 CRITICAL；来源分 Core、插件与账号多选，可搜索；日期），筛选生效时按钮带圆点；CRITICAL、ERROR 与 WARNING 计数，点击只看该级别；搜索内容或 trace；实时开关；「清空」只丢弃页面已载入的记录，不删除日志文件；「导出诊断包」打开对话框，选择时间范围与三组内容，缺省脱敏（ADMIN-API §19）。
- 今天的记录先读 recent，再订阅 SSE（会先重发缓冲中的记录），按时间与内容去重；过去日期查询 history。筛选与搜索只在浏览器中进行，条件保存在 localStorage。
- 手机上每条记录两行：第一行时间、级别、来源与源码位置（截断），第二行消息；实时、清空与导出收进 ⋯。

不做：日志 API、诊断包与 CLI（T-09.10、T-09.19）；级别设置（goal-44、goal-43）。
```

### goal-156 · T-09.23 官方 WebUI 的后台进程、统计、商店、用户与数据页

状态：等待前置（T-09.8、T-09.20、T-09.21）。T-09.22（PR #220，`b5e572e`）、T-09.24（PR #215，`b911f27`）已合并。2026-11-17 按 D-162～D-166 与 D-169 新增。目标、依据与要点由 PLAN T-09 步骤 7（WebUI 后台）与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-156（切片 T-09.23：官方 WebUI 的后台进程、统计、商店、用户与数据页）。goal 目录 goal-156，分支 goal-156，worktree .worktrees/goal-156，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：为官方 WebUI 组件加上 Go 入口（D-162）：订阅 `command.invoked` 与消息事件，按小时分桶统计收发量、命令触发与错误，提供统计路由；读取插件源 `index.json` 提供商店数据。前端接入总览的图表、插件商店与发布插件、插件详情的「存储」视图与 changelog 显示。

依据：PLAN T-09 步骤 7（WebUI 后台）；CONTRACTS §7.5、§9.1（WebUI 的 Go 入口）、§9.2（插件源）、§8（标准事件 command.invoked）；ADMIN-API §20～§22；DECISIONS D-162～D-166、D-169；TEST-MATRIX V-136，V-134、V-135、V-138 的页面分项。画板：「v4 · 总览」「v4 · 插件 · 商店」「v4 · 插件 · 商店（发布插件）」「v4 · 插件 · 天气 · 数据（存储）」「v4 · 插件 · Agent · 数据（会话记录）」及其深色与手机画板。

要点：
- 后台进程放在 web/admin 的 Go 入口包中，按插件的方式使用受管 SQLite 保存小时桶；路由要求管理会话，前端同源调用。统计只依赖 Core 的基础设施（事件与 command.invoked），Core 不提供图表 API。
- 总览：消息量（收到与发出）、各账号、热门插件（只算常规插件的命令触发次数）、错误与警告图表，时间范围 24 小时、7 天、30 天；故障标在曲线上。
- 商店：网格卡片，分类只按组件类型，作者标签显示在卡片上；只显示 GitHub 星数；官方组织（alonabot/*）的插件名旁加认证图标；可以切换插件源（缺省官方源）；分页用数字页码；「发布插件」入口说明如何把仓库加入 index.json，并链接插件模板仓库 AlonaBot/alonabot-plugin-template（D-169）。
- 存储视图：KV（键列表、查看与编辑、删除）、数据目录文件（浏览、上传、下载、新建目录、改名、删除）、整个插件数据的导出与清空（ADMIN-API §21）；会话记录视图列出 Core 托管的 Agent 会话，可查看与清除。
- 更新后在插件详情与「关于」中显示 changelog（ADMIN-API §22，Markdown 由前端渲染，不执行其中的 HTML）。
- 后台进程的 build 与 pack 按 D-162 照插件出平台包；官方 WebUI 从此按平台打包（T-14.3 同步）。

不做：Core 的事件、API 与插件进程的支持（T-09.20～T-09.22、T-09.24）；模型与配置文件页（goal-146、goal-165）。
```

## 第 20 波

### goal-146 · T-09.18 模型页

状态：等待前置（T-09.8、T-09.14、T-12.16）。2026-11-17 按 D-154 与 D-169 新增。目标、依据与要点由 PLAN T-09 步骤 7（模型页）与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-146（切片 T-09.18：模型页）。goal 目录 goal-146，分支 goal-146，worktree .worktrees/goal-146，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：侧栏一级入口「模型」：编辑 Agent 配置中的 providers 与 models，调用 Agent 的模型管理路由拉取模型列表、测试模型与读取当天用量。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（模型页）；CONTRACTS §13.5、§13.6；ADMIN-API §12；DECISIONS D-154、D-155、D-169；TEST-MATRIX V-126。画板：「v4 · 模型」「v4 · 模型（添加提供商）」及其深色画板。

要点：
- 按 `role: agent` 找到已安装的 Agent；没有时显示空状态与去商店的按钮，侧栏的「模型」置灰；Agent 没有模型管理路由时只提供配置编辑。
- 按类型筛选（对话、语音识别、语音合成、向量、重排、决策），按提供商分组列出模型，显示模型 ID、类型、当天用量与测试按钮；预留的五类只能添加、拉取与测试。
- 添加提供商对话框：预设、名称、Base URL、API Key（secret，只显示已设置），「拉取模型列表」后勾选要加入的模型，拉取失败时可以手动输入模型 ID。
- 修改经配置 API 的 Pending 与 reload（只提交 patch）；删除被 chat_model 或 fallback_chat_models 引用的模型时先提示引用位置。

不做：Agent 的模型管理路由（T-12.16）；配置文件中的模型选择（goal-165）。
```

## 第 22 波

### goal-165 · T-09.25 配置文件页

状态：等待前置（T-04.19、T-11.6、T-12.19、T-09.18）。2026-11-17 按 D-167、D-168 与 D-169 新增。目标、依据与要点由 PLAN T-09 步骤 7（配置文件页）与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-165（切片 T-09.25：配置文件页）。goal 目录 goal-165，分支 goal-165，worktree .worktrees/goal-165，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：侧栏「配置文件」打开即是当前配置文件的设置（只有一层），分区为 AI、平台与消息与插件；切换、新建、使用范围、重命名、复制与删除；插件配置页可切换配置文件。本切片不改 Go 代码。

依据：PLAN T-09 步骤 7（配置文件页）；CONTRACTS §5.6；ADMIN-API §23；DECISIONS D-167、D-168、D-169；TEST-MATRIX V-145。画板：「v4 · 配置文件」「v4 · 配置文件（平台与消息）」「v4 · 配置文件（切换配置文件）」「v4 · 配置文件（更多菜单）」「v4 · 配置文件（使用范围）」「v4 · 配置文件（AI · VIP 群）」「v4 · 配置文件（插件 · VIP 群）」「v4 · 配置文件（没有 Agent）」、y=5200 一排的配置文件深色画板与「v4 · 手机 · 配置文件」。

要点：
- 标题「● 名称 ⌄」打开切换菜单（各配置文件的名称与会话数、新建配置文件）；旁边的「N 个会话」打开使用范围对话框（各账号作为缺省的开关，旁边写当前缺省；单独指定的会话可添加与移出）；⋯ 菜单为重命名、复制与删除，「默认」的删除置灰。页面没有返回按钮。
- AI 分区：已安装 Agent 的 schema 中写了 `profile` 的字段（D-168）——对话模型、图像转述、人格、触发方式（多选：@ 或回复机器人、唤醒词、私聊消息）、选中唤醒词时显示的唤醒词列表、群成员各自独立对话——以及按配置文件覆盖的工具策略（只列覆盖了全局的工具，「+ 覆盖其他工具」）。没有 Agent 时不显示 AI 分区，在分区上方提示并给出去商店的按钮。
- 平台与消息分区：命令前缀（可增删的标签），回复时 @ 发送人、引用原消息、回复前缀与长回复合并转发（标注只在支持的平台生效）。
- 插件分区：每个插件一行，开关与「覆盖 N 项：字段名」或「跟随默认」，箭头进入该插件的配置页并预先选中这个配置文件；全局已停用的插件置灰。
- 「默认」的字段修改走插件配置 API（Pending 与 reload，保存时提示会重载的插件）；其他配置文件经 ADMIN-API §23 立即生效。插件配置页选中「默认」以外的配置文件时，只有写了 `profile` 的字段可以编辑，其余字段置灰。无效覆盖值显示原因并可一键清理。

不做：配置文件的 Core、SDK 与 Agent 实现（T-08.18、T-04.18、T-04.19、T-11.6、T-12.19）；会话页的配置文件列（goal-44）。
```

### goal-175 · S-06.3 休眠策略与运行状态的页面

状态：等待前置（T-09.8、T-09.9、S-06.1）。2026-11-17 维护者决定把插件休眠（PLAN §7）纳入本轮时新增：休眠控件从 goal-43、goal-44 移到本切片，使 WebUI 主线与发布链不依赖休眠切片（S 工作包不进入 v0.1 发布门槛）。目标、依据与要点由 PLAN §7 S-06 步骤 1、ADMIN-API §24 与设计稿确定；前置合并后，审查者按合并后的代码写出接口与测试段，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-175（切片 S-06.3：休眠策略与运行状态的页面）。goal 目录 goal-175，分支 goal-175，worktree .worktrees/goal-175，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」与 docs/goals/webui.md「设计稿与共同约定」。

目标：在 goal-43 的插件详情「高级」标签加入休眠策略（跟随全局、永不休眠或时长；manifest 要求常驻或其他原因不能休眠时锁定并显示原因与提示），在 goal-44 的系统设置加入全局运行模式与「插件空闲休眠」时长；插件列表与详情显示运行状态（含「已休眠」）与最近的唤醒/休眠原因。本切片不改 Go 代码。

依据：PLAN §7 S-06 步骤 1；CONTRACTS §4.4、§4.5、§15.1；ADMIN-API §24；DECISIONS D-153；TEST-MATRIX V-81 的页面分项。画板：「v4 · 插件 · 天气 · 高级」「v4 · 插件 · 天气 · 高级（不能改休眠）」「v4 · 系统设置」。

要点：
- 只调用 ADMIN-API §24；读取不唤醒插件。保存后唤醒失败时显示实际错误并重新读取当前设置。
- 原因与警告的文案按 §24 的稳定值在 i18n 中翻译；未声明 Core 过滤的消息事件、Raw、HTTP、Hook 或持续监听的插件被显式给出时长时，提示会被频繁唤醒。

不做：休眠的 Core 实现与管理 API（S-01.1～S-06.1）；固定机器测量（S-06.2）。
```
