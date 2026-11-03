# goal 语句：WebUI

本文件收录「WebUI」章节的 goal 语句，切片表在[WebUI 实施计划](../../spec/plan/webui.md)。用法、状态与所有 goal 的共同约束见 [README](README.md)；本文件只由审查者维护。

## 总览

| goal | 切片 | 标题 | 前置 | 状态 |
| --- | --- | --- | --- | --- |
| goal-34 | T-08.7 | web/admin 骨架、配置表单组件与界面文案 | T-08.1、T-08.10 | 等待前置 |
| goal-84 | T-09.12 | WebUI 组件与组件图标 | T-07.4、T-09.1、T-09.4、T-13.1 | 等待前置 |
| goal-43 | T-09.8 | Admin 管理页 | T-09.6、T-08.7、T-09.12 | 等待前置 |
| goal-44 | T-09.9 | Admin 配置、权限、工具策略与插件页 | T-09.8、T-11.1、T-11.3 | 等待前置 |

## 第 8 波

### goal-34 · T-08.7 web/admin 骨架、配置表单组件与界面文案

状态：等待前置（T-08.10）。T-08.1 已合并（PR #54，合并提交 `021e972`）。

已用 `sdk/generate/config_test.go` 中的期望 Schema 核对 T-08.1 的格式：字段按声明顺序出现在 `properties` 中；可空字段的 `type` 为 `["<类型>","null"]`；数组为 `type: array` 加 `items`；`map[string]T` 为 `type: object` 加 `additionalProperties`，没有 `properties`；`x-alonabot-secret`、`x-alonabot-hidden` 只在为 true 时出现；`default` 写在 Default 上。

核对时发现生成器的标签键与契约当时的 i18 写法冲突。维护者按 D-74 改为仿照 AstrBot，由 T-08.10（goal-90）修订生成器：
- label、description 改为标签中的英文源文案；`x-alonabot-form` 增加 labels、subtitles、step、language、when、obvious。
- 删除 `x-alonabot-i18n` 与 `option_icons`，选项不再带图标。
- 翻译文件的 `config` 按 JSON 字段名嵌套，覆盖 label、description、labels、subtitles。

本语句的 `Field`、`fieldText` 与表单行为已按此修订。goal-90 合并后，审查者用它生成 fixtures 并核对本语句，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-34（切片 T-08.7：web/admin 骨架、配置表单组件与界面文案）。goal 目录 goal-34，分支 goal-34，worktree .worktrees/goal-34，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：建立 web/admin 前端工程（Vite、TypeScript、React 19、vitest），它是官方 WebUI 组件（`role: webui`）的源码，构建产物 dist/ 是组件的静态根目录；提供由配置 Schema 驱动的表单组件：字段按声明顺序展示，hidden 字段不显示也不被删除，secret 字段显示「已设置」状态而不把掩码当成保存值，提交生成 RFC 7396 Merge Patch；字段标签按 i18 回退规则显示；WebUI 自身文案有 en-US 与 zh-CN 目录和语言切换组件。真实管理页面与 API 在 T-09 接入。

依据：PLAN T-08 步骤 9（前端）；CONTRACTS §5.1 第 4 条、§5.3、§5.4 最后一条、§7.5、§9.1（WebUI 字段）；DECISIONS D-43、D-62、D-63、D-74；docs/development.md（Node 版本与已批准的前端依赖）；TEST-MATRIX V-25 的表单分项、V-54 的界面组件分项。

范围（只新增或修改这些）：
- 新增 web/admin/（package.json、package-lock.json、tsconfig.json、vite.config.ts、index.html、src/、fixtures/、i18n/，以及 manifest plugin.yaml：`role: webui`、`assets: dist`、Admin API 主版本范围 1～1）
- .gitignore：追加 web/admin/node_modules/ 与 web/admin/dist/
- docs/development.md 不改；Node 使用 asdf 管理的 26.5.1，包管理用 npm
允许的依赖：只用 react、react-dom、typescript、vite、@vitejs/plugin-react、vitest（均属已批准的 Vite、TypeScript、React 19、vitest）。需要 DOM 测试环境（jsdom、happy-dom）或 Testing Library 时标阻塞，等待维护者批准。

接口（模块与导出，照此实现）：
- src/schema/model.ts
  export type Field = { path: string[]; type: 'string'|'integer'|'number'|'boolean'|'object'|'array'|'map'; nullable: boolean; required: boolean; secret: boolean; hidden: boolean; obvious: boolean; multiple: boolean; widget?: string; options?: (string|number)[]; labels?: string[]; subtitles?: string[]; min?: number; max?: number; step?: number; language?: string; when?: Record<string, unknown>; defaultValue?: unknown; label?: string; description?: string; children?: Field[] } // label、description、labels、subtitles 是标签中的英文源文案；path 是 JSON 名的路径，数组元素与 map 值不含下标或键
  export function fieldsFromSchema(schema: unknown): Field[] // 按 properties 的声明顺序（Schema 的 JSON 键顺序）；hidden 字段保留在模型中，标记 hidden
- src/schema/patch.ts
  export type SecretState = { set: boolean } // 来自管理读取的 {"$secret_set": true}
  export function buildPatch(original: unknown, edited: unknown, fields: Field[]): Record<string, unknown> // RFC 7396：只包含改变的键；未触碰的 secret 与 hidden 字段不出现；清空可空字段生成 null 删除；数组整体替换
- src/i18n/catalog.ts
  export function pickLocale(available: string[], preferred: readonly string[]): string // 精确 → 主语言（字典序最小）→ en-US
  export function fieldText(locale: Record<string, unknown>, field: Field): { label: string; description?: string; labels?: string[]; subtitles?: string[] } // 按 field.path 在 locale.config 中逐层查找覆盖 → 源文案 → 字段的 JSON 名；locale 为 pickLocale 选出的插件语言文件
- src/components/ConfigForm.tsx：export function ConfigForm(props: { schema: unknown; value: unknown; secrets: Record<string, SecretState>; messages: Record<string, unknown>; onSubmit: (patch: Record<string, unknown>) => void })
- src/components/LanguageSwitch.tsx：export function LanguageSwitch(props: { value: string; onChange: (tag: string) => void })
- i18n/en-US.json、i18n/zh-CN.json：WebUI 文案（键集合一致）

行为：
- 表单：string 用文本框（widget=password 时为密码框，textarea 时为多行），integer/number 用数字框并按 min/max 限制，boolean 用开关，integer/number 的 widget=slider 用滑块（min、max、step），string 的 widget=code 用代码编辑区（language 只作标注，不引入编辑器依赖），select 用下拉（显示 labels，下方显示 subtitles，提交 options 中的存储值；multiple 为多选），description 作为提示，obvious 时醒目显示；when 的条件不满足时字段不显示，但已有值不被删除，也不进 patch；object 嵌套为分组，array 与 map 提供增删行。required 字段标记必填。表单只在提交时生成 patch，不自行补默认值（默认值由 SDK 在准备时补，T-08.3）。
- secret：显示「已设置/未设置」与「修改」按钮；未点「修改」时不出现在 patch 中；点了并输入后以新值出现；提供「清除」时（可空字段）生成 null。不把任何掩码字符串放进 patch。
- hidden：不渲染，但 patch 永远不包含它（不会因未显示而被删除）。
- 未知键：表单只渲染 Schema 中的字段；patch 只包含被编辑的已知字段，未知键不受影响。
- 字段文案：标签、提示与选项文案按 fieldText 取得：pickLocale 选出的插件语言文件中，`config` 按字段的 JSON 名逐层嵌套，覆盖键为 label、description、labels、subtitles（D-74，仿照 AstrBot）；缺失时用 Schema 中的源文案，标签再回落字段的 JSON 名，说明缺失时不显示；options 的存储值不翻译。
- 语言切换：初始语言按浏览器 navigator.languages 用 pickLocale 在 en-US 与 zh-CN 中选择；用户选择后保存在 localStorage，与机器人回复语言无关。
- fixtures/：至少三个 Schema fixture（嵌套与顺序、secret 与 hidden、可空与数组/map），由 goal-90 修订后的生成器为对应 Go 类型实际生成后放入，并附一个覆盖嵌套字段与选项文案的插件语言文件（审查者在派发前核对）。
- npm scripts：dev、build、test（vitest run）、typecheck（tsc --noEmit）。

测试（按编码规范「测试提交规则」）：
- 提交：web/admin/src/schema/schema.test.ts（vitest，不依赖 DOM）：用 fixtures 驱动 fieldsFromSchema 与 buildPatch，覆盖字段顺序、hidden 不进 patch、未修改的 secret 不进 patch、修改 secret 进 patch、清空可空字段生成 null、数组整体替换、未知键不受影响；pickLocale 与 fieldText 的回退（当前语言覆盖 → 源文案 → JSON 名）。这是 Schema 格式与表单之间的契约测试。
- 本地自测（不提交，PR 中写明结果）：npm run build、npm run typecheck；在浏览器中手动打开 dev 页面，用 fixture 渲染表单并切换语言（写明浏览器与版本，属于手动验证）。

完成条件：在 web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 全部成功；go build ./... 不受影响。

不做：真实管理 API、登录与页面路由（T-09.1、T-09.8）；CI 中的前端任务（另行由维护者决定是否加入）；DOM 渲染测试（需要新依赖）。
```


## 第 9 波

### goal-84 · T-09.12 WebUI 组件与组件图标

状态：等待前置（T-07.4、T-09.1、T-09.4、T-13.1）。语句只写了目标、依据和要点。依据 D-62、D-63 新增；前置合并后由审查者按实际的 webhost、adminauth、manage 与根包装配写出接口和测试段。

目标：Core 在 admin host 上以组件方式提供 WebUI 与组件图标：已启用 WebUI 组件的静态文件（保留 `/api/` 与内部路径，无扩展名的不存在路径回退 index.html，未安装时 503 与英文提示），Admin API 主版本常量与 WebUI 声明范围的检查（不兼容时拒绝安装与启用并提示），只提供 manifest 声明的图标并带 sandbox CSP 与 nosniff、要求管理员会话；写出 spec/ADMIN-API.md（主版本、全部接口、认证与错误格式）；根包增加 `alonabot.WebUI(fsys fs.FS)` 选项。

依据：PLAN T-09 步骤 9、§3.1 的 T-09.12 行；CONTRACTS §7.3、§7.5、§9.1、§15.4；DECISIONS D-62、D-63；TEST-MATRIX V-83。

要点：
- 文件只从不可变包目录读取，路径按 CONTRACTS §9.1 的相对路径规则校验，不跟随链接，不经组件进程。
- 图标的类型只按扩展名取 image/svg+xml、image/png、image/webp；未声明的包内文件即使存在也返回 404。
- spec/ADMIN-API.md 由本切片新建，之后新增接口的切片同步追加；同一主版本内只做向后兼容的增加。
- 提交的集成测试经真实 Core 与临时安装的 WebUI 测试包覆盖 V-83 的全部场景；含脚本的 SVG 在浏览器中不执行的检查没有浏览器时记为未运行。


## 第 16 波

### goal-43 · T-09.8 Admin 管理页

状态：等待前置（T-09.6、T-08.7、T-09.12）。语句中 goal-34、goal-36、goal-39～goal-41、goal-84 提供的符号（web/admin 的工程与组件、adminauth 的登录接口、manage 的 HTTP API、Admin API 参考与 WebUI 的提供）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-43（切片 T-09.8：Admin 管理页）。goal 目录 goal-43，分支 goal-43，worktree .worktrees/goal-43，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在官方 WebUI 组件 web/admin 中实现真实的登录、总览状态、各角色组件的安装、启停与替换（`--source`、`--replace` 的页面入口，组件图标用 `<img>` 显示）、Adapter 实例管理、更新与恢复（rollback）页面，全部只调用 spec/ADMIN-API.md 列出的 Admin API；静态文件由 T-09.12 以 WebUI 组件的方式提供。页面显示 Pending、候选失败、Degraded 与维护原因，只展示用户可操作的状态和具体错误。

依据：PLAN T-09 步骤 7（管理页）；CONTRACTS §7.3、§7.4、§9.3、§12.1；DECISIONS D-25；TEST-MATRIX V-40、V-47 的页面分项。

范围（只新增或修改这些）：
- web/admin/src/：新增 api/（管理 API 客户端，带 CSRF 头）、pages/（Login、Overview、Packages、Instances、Updates）、路由（基于 window.location 与 history 的简单路由，不引入路由库）
- web/admin/i18n/：补充页面文案（en-US 与 zh-CN 键集合一致）
- 新增 internal/core/adminui/（嵌入 web/admin/dist 的静态资源并在 admin host 提供；go:embed）
- cmd/alonabot 或根包 assembly：admin host 的非 /api/ 路径交给 adminui（只加调用点）
- 构建流程：web/admin 的 npm run build 产物 dist/ 由 internal/core/adminui 的 go:generate 复制进包目录；生成物提交到仓库（与 proto 生成物同样处理），并在 docs/development.md 的生成命令中登记（本切片获授权修改该处）
- 上述包的测试文件
允许的依赖：前端仍只用已批准的 React、Vite、TypeScript、vitest；不引入路由、状态管理或 UI 库。

行为：
- 登录：POST /api/auth/login，成功后保存 csrf_token（内存中，刷新页面时重新登录或经 /api/auth/session 取回；若 goal-36 没有该端点，本切片在 adminauth 中追加 GET /api/auth/session 返回当前会话的 csrf_token，只在已登录时可用）；登出。未登录访问任何页面跳转到登录页。
- Overview：Core 版本、监听地址、每个包的状态（未启用、运行、Pending、候选失败、Degraded 及原因、维护中）。
- Packages：上传安装包（archive、release.json、.sig 三个文件，multipart 由 API 接收，若 goal-39 的 API 只接受路径，本切片在 manage 中追加 multipart 端点并写入临时目录后调用 Install）；启用、停用、重启、卸载（含 purge）、pin；需要确认时展示受影响列表并要求再次确认（API 的 409）。
- Instances：列出 Adapter 实例（不显示凭证，只显示「已设置」）、创建（表单由该 Adapter 包的配置 Schema 经 goal-34 的 ConfigForm 生成）、启停、删除。
- Updates：update 上传新包、rollback 选择保留代次；显示操作结果与失败原因。
- 安全：所有修改请求带 X-Alonabot-CSRF；页面不读取或展示 token、secret 明文；静态资源响应带 CSP（default-src 'self'）、X-Frame-Options: DENY 与 nosniff。
- adminui：只在 admin host 提供；未知路径返回 index.html（单页应用）；/plugins/ 与 /api/ 不经过它。

测试（按编码规范「测试提交规则」）：
- 提交：
  - web/admin/src/api/api.test.ts（vitest，不依赖 DOM）：用 fetch 的测试替身验证 API 客户端的 CSRF 头、409 确认流程的数据结构、错误格式解析。
  - internal/core/supervisor/adminui_test.go（//go:build unix）：admin host 上 / 返回 index.html 且带 CSP；plugin host 上不提供；未登录时 API 返回 401；登录后经 API 完成一次安装与启用（复用 goal-39 的测试包）。
- 本地自测（不提交，PR 中写明结果）：npm run build 与 typecheck；在浏览器中手动走一遍：登录 → 安装 → 启用 → 创建实例 → 停用 → rollback（写明浏览器与版本，属于手动验证）。

完成条件：web/admin 下 npm ci、npm run typecheck、npm run test、npm run build 成功；go generate ./... 后无未提交差异；gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：配置、命令权限、发信许可、工具策略、会话与插件页（T-09.9）；日志页（T-09.10）；Core 设置页（T-09.9）。
```


## 第 17 波

### goal-44 · T-09.9 Admin 配置、权限、工具策略与插件页

状态：等待前置（T-09.8、T-11.1、T-11.3）。语句中 goal-13、goal-29、goal-31、goal-34、goal-39、goal-43、goal-53～goal-56、goal-26 提供的符号（policy 的覆盖与 Admin 方法、config.Service、secret.Redact、ConfigForm、manage API、web/admin 页面框架、i18n 的默认 locale、tool.Catalog、conversation.Service、webhost 的票据）只写名字；合并后由审查者核对，再改为待派发。

```text
/goal 使用 $alonabot-execute 推进 goal-44（切片 T-09.9：Admin 配置、权限、工具策略与插件页）。goal 目录 goal-44，分支 goal-44，worktree .worktrees/goal-44，从最新 origin/main 创建。开始前完整读取 docs/goals/README.md「所有 goal 的共同约束」。

目标：在管理 API 与 web/admin 中补齐：插件配置（表单与完整 JSON 编辑，使用 raw JSON、Merge Patch、base_generation 与 secret 语义，显示 Pending 并经 reload 激活）、命令权限与启用、发信许可（allowed_adapter_instances）、Core 设置（Core Admin 列表、默认 locale、全局禁止表）、工具调用策略、Agent 会话的列出与清除、插件页面入口（经兑换票据进入 plugin host），以及界面语言切换。

依据：PLAN T-09 步骤 7（配置与插件页）；CONTRACTS §5、§2.3、§8（allowed_adapter_instances）、§13、§7.3、§5.4；DECISIONS D-45；TEST-MATRIX V-12、V-22、V-23、V-25、V-43、V-54、V-55、V-56 的页面分项。

范围（只新增或修改这些）：
- internal/core/manage/：新增 config_api.go、policy_api.go、settings_api.go、tool_api.go、conversation_api.go、pluginpage_api.go
- web/admin/src/pages/：新增 Config、Permissions、CoreSettings、Tools、Conversations；Packages 页增加「打开插件页面」入口
- web/admin/i18n/：补充文案
- internal/core/adminui：重新生成嵌入的构建产物
- 上述包的测试文件
允许的依赖：与 goal-43 相同，不新增前端依赖。

接口（HTTP，全部在 /api/v1 下、经 adminauth.Require；JSON；错误格式与 goal-39 一致）：
- GET /plugins/{id}/config：{active_generation, active: Redact 后的快照, pending: {base_generation, config: Redact 后}|null, schema}
- PUT /plugins/{id}/config/pending：{base_generation, config}（完整替换）；PATCH /plugins/{id}/config/pending：{base_generation, patch}（RFC 7396）；DELETE 丢弃 Pending；POST /plugins/{id}/config/import 导入 config.json；secret 字段的新值在写入 Pending 前经 secret.Extract 存为引用，省略或未修改的 secret 保持原引用；base 不是当前 Active 时 409 并返回 FAILED_PRECONDITION。
- POST /plugins/{id}/reload：经 goal-40/41 激活 Pending。
- GET/PUT /plugins/{id}/policies：命令覆盖（作用域、permission、enabled），以及 Router 的遮蔽诊断。
- GET/PUT /plugins/{id}/allowed-instances：发信许可集合。
- GET/PUT/DELETE /core/admins、/core/denies，GET/PUT /core/default-locale。
- GET/PUT /tools/policies：工具策略（缺省与覆盖）。
- GET /plugins/{id}/conversations、DELETE /plugins/{id}/conversations/{key}。
- POST /plugins/{id}/page-ticket：返回 webhost.IssueTicket 的 URL（只给有 HTTP 路由的插件）。
- GET /secrets/{plugin}/{path}/reveal：要求请求体中的当前 token 再认证（adminauth.Reauthenticate），通过后返回明文并以 INFO 审计记录（不含明文）。

行为：
- 配置页：用 ConfigForm 编辑（只提交 patch）或切换到完整 JSON 编辑（textarea，提交完整替换）；显示 Pending 与 Active 的差异（按键列出，secret 只显示是否已设置）；「激活」调用 reload；候选失败与 Degraded 显示具体原因。hidden 字段只在完整 JSON 编辑中可见。
- 权限页：按命令树显示有效权限与启用状态（经 policy.CheckCommand 的同一规则计算），可设置插件、命令组、路由三层覆盖；显示遮蔽关系。
- 发信许可：列出全部 Adapter 实例，勾选授予。
- Core 设置：Core Admin 列表（与 alonabot admin add|remove|list 同一服务）、默认 locale、命令前缀 `command_prefixes` 与按 Adapter 实例的覆盖（保存在状态库，经 goal-87 的 `router.New` prefixes 参数生效）、全局禁止表。
- 工具页：列出全部工具及缺省与覆盖策略。
- 会话页：按插件列出会话摘要（键、条目数、最后时间），可清除；不显示条目内容。
- 语言切换：沿用 goal-34 的组件，保存在 localStorage。
- 插件页面：点击后请求 page-ticket，在新窗口打开 plugin host 上的兑换 URL；未配置第二主机名时按钮禁用并提示。
- 所有修改请求带 CSRF 头；修改后的策略在下一次派发时生效（页面不需要重启插件）。

测试（按编码规范「测试提交规则」）：
- 提交：
  - internal/core/supervisor/admin_api_test.go（//go:build unix）：经管理 API 完成：PATCH 配置（未修改的 secret 保持原引用、新 secret 生成新引用、hidden 不被删除）、陈旧 base 409、reload 后 Active 改变；设置命令覆盖后下一条消息按新权限处理；授予发信许可后插件后台 Send 成功；设置默认 locale 后新 Session 使用；工具策略改为 deny 后 ListTools 不含它；清除会话；page-ticket 只给有 HTTP 路由的插件；reveal 需要再认证。
  - web/admin 的 vitest：API 客户端对新端点的请求格式与 409 处理。
- 本地自测（不提交，PR 中写明结果）：npm run build；浏览器中手动走一遍各页面并切换语言（写明浏览器与版本）。

完成条件：web/admin 下 npm run typecheck、npm run test、npm run build 成功；go generate ./... 后无未提交差异；gofmt -l 无输出；go vet ./...；go test -race ./internal/... ./cmd/...；go build ./...。

不做：日志页（T-09.10）；故障注入与验收汇总（T-09.11）。
```
