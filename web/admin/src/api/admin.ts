/**
 * Admin API 客户端与数据结构定义。
 * 遵循 spec/ADMIN-API.md 规范与 CONTRACTS §7.3～§7.5、§9.3、§12.1。
 */

export class AdminAPIError extends Error {
  status: number;
  code?: string;
  affected?: string[];

  constructor(status: number, message: string, code?: string, affected?: string[]) {
    super(message);
    this.name = 'AdminAPIError';
    this.status = status;
    this.code = code;
    this.affected = affected;
  }
}

/** 判断错误是否为服务端返回的需要显式确认的 409 确认对象。 */
export function isConfirmationError(err: unknown): err is AdminAPIError & { affected: string[] } {
  return err instanceof AdminAPIError && err.status === 409 && Array.isArray(err.affected);
}

export type CoreInfo = {
  version: string;
  listen: string;
  admin_api_major: number;
};

export type Package = {
  ID: string;
  Role: string;
  AdapterID: string;
  Version: string;
  Source: string;
  PublisherKeyID: string;
  Pin: string;
  Enabled: boolean;
  Uninstalled: boolean;
  Signed: boolean;
  Active: string;
  State: string;
  Warnings?: string[];
  pages?: { path: string; title: string }[];
};

export type Generation = {
  id: string;
  version: string;
  created_at: string;
  active: boolean;
  has_db: boolean;
};

export type ConfigOption = { value: string; label: string; description?: string };
export type OptionsPage = { options: ConfigOption[]; more: boolean };

export type Correction = { path: string; reason: 'converted' | 'reset' | 'removed'; original?: unknown };
export type ConfigDiagnostics = { corrections: Correction[]; unknown: string[] };
export type PendingConfig = { base_generation: string; config: Record<string, unknown> };
export type ConfigSnapshot = ConfigDiagnostics & {
  active_generation: string;
  active: Record<string, unknown>;
  pending: PendingConfig | null;
  schema: unknown;
  i18n: Record<string, Record<string, unknown>>;
  messages: Record<string, Record<string, unknown>>;
};

export type CommandScope = { kind: 'plugin' } | { kind: 'group'; group: string[] } | { kind: 'route'; route_id: string };
export type CommandOverride = { scope: CommandScope; permissions: string[] | null; enabled: boolean | null };
export type InvalidCommandName = { route_id: string; name: string[] | null; reason: 'declared' | 'unregistered' };
export type CommandPolicy = {
  route_id: string;
  enabled: boolean;
  permissions: string[] | null;
  declared_names: string[][];
  effective_names: string[][];
  has_names_override: boolean;
  invalid_names: InvalidCommandName[];
};
export type CommandReference = { plugin_id: string; route_id: string };
export type PluginPolicies = {
  overrides: CommandOverride[];
  commands: CommandPolicy[];
  shadows: { path: string[]; winner: CommandReference; shadowed: CommandReference[] }[];
  invalid_names: InvalidCommandName[];
};

export type ToolPolicyValue = 'allow' | 'confirm' | 'deny' | 'inherit';
export type ToolPolicy = {
  plugin_id: string;
  service: string;
  default: ToolPolicyValue;
  override: ToolPolicyValue;
  effect: 'read_only' | 'write' | 'destructive' | 'unspecified';
};
export type ToolPolicyUpdate = { plugin_id: string; service: string; policy: ToolPolicyValue };

export type AllowedInstances = { all: boolean; instances: string[] };
export type PluginJob = {
  job_id: string;
  trigger: string;
  paused: boolean;
  last_run_at?: string;
  next_run_at?: string;
};
export type GlobalLogLevel = 'debug' | 'info' | 'warn' | 'error' | 'critical';

export type SourceLogLevel = 'debug' | 'info' | 'warn' | 'error' | 'critical' | 'off' | 'inherit';

export type LogLevels = {
  global: string;
  sources: Record<string, SourceLogLevel>;
};

// 内存中维护的 CSRF Token（不存入 localStorage/Cookie，防明文泄露）
let currentCSRFToken: string | null = null;

function setCSRFToken(token: string | null): void {
  currentCSRFToken = token;
}

type SessionExpiredHandler = () => void;
const sessionExpiredHandlers = new Set<SessionExpiredHandler>();

/** 订阅受保护请求收到的 401 认证过期事件，返回取消订阅函数。 */
export function onSessionExpired(handler: SessionExpiredHandler): () => void {
  sessionExpiredHandlers.add(handler);
  return () => {
    sessionExpiredHandlers.delete(handler);
  };
}

function notifySessionExpired() {
  currentCSRFToken = null;
  for (const handler of sessionExpiredHandlers) {
    try {
      handler();
    } catch (err) {
      console.error('Session expired handler failed:', err);
    }
  }
}

async function parseErrorResponse(response: Response): Promise<AdminAPIError> {
  const text = await response.text();
  if (response.headers.get('content-type')?.includes('application/json')) {
    let data: unknown;
    try {
      data = JSON.parse(text);
    } catch {
      return new AdminAPIError(response.status, text);
    }
    if (data && typeof data === 'object') {
      const error = data as Record<string, unknown>;
      const affected = Array.isArray(error.affected) && error.affected.every(item => typeof item === 'string')
        ? error.affected as string[] : undefined;
      const code = typeof error.code === 'string' ? error.code : undefined;
      const message = typeof error.safe_message === 'string' ? error.safe_message : text;
      return new AdminAPIError(response.status, message, code, affected);
    }
  }
  return new AdminAPIError(response.status, text || `HTTP ${response.status}`);
}

interface RequestOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: string | FormData;
}

async function request<T = unknown>(path: string, options: RequestOptions = {}): Promise<T> {
  const method = options.method || 'GET';
  const headers: Record<string, string> = { ...(options.headers || {}) };
  const isMutating = method !== 'GET' && method !== 'HEAD';
  const isLogin = path === '/api/auth/login';

  // 受保护的变更请求携带内存中的 X-Alonabot-CSRF
  if (isMutating && !isLogin && currentCSRFToken) {
    headers['X-Alonabot-CSRF'] = currentCSRFToken;
  }

  // Origin 由浏览器按当前页面来源设置，Cookie 仅发往同源入口。
  const response = await fetch(path, {
    method,
    headers,
    body: options.body,
    credentials: 'same-origin',
  });

  if (!response.ok) {
    // 受保护请求收到 401 触发会话过期（login 失败留在登录页展示，不触发全局过期通知）
    if (response.status === 401 && !isLogin) {
      notifySessionExpired();
    }
    throw await parseErrorResponse(response);
  }

  if (response.status === 204) {
    return undefined as unknown as T;
  }

  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    return (await response.json()) as T;
  }

  return (await response.text()) as unknown as T;
}

// ----------------------------------------------------------------------------
// 认证与会话 API
// ----------------------------------------------------------------------------

/** 提交管理 token 登录，200 时取 csrf_token 存入内存。 */
export async function login(token: string): Promise<void> {
  const result = await request<{ csrf_token: string }>('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });
  setCSRFToken(result.csrf_token);
}

/** 页面加载或刷新后取回已有管理会话与固定过期时间。 */
export async function restoreSession(): Promise<{ expires_at_unix_ms: number }> {
  const result = await request<{ csrf_token: string; expires_at_unix_ms: number }>('/api/auth/session', {
    method: 'GET',
  });
  setCSRFToken(result.csrf_token);
  return { expires_at_unix_ms: result.expires_at_unix_ms };
}

/** 登出并撤销管理会话，清空本地 CSRF。 */
export async function logout(): Promise<void> {
  await request<void>('/api/auth/logout', { method: 'POST' });
  setCSRFToken(null);
}

// ----------------------------------------------------------------------------
// Core 与组件管理 API
// ----------------------------------------------------------------------------

/** 获取 Core 基本信息（版本、监听地址、Admin API 主版本号）。 */
export async function getCore(): Promise<CoreInfo> {
  return request<CoreInfo>('/api/v1/core');
}

let currentClientLocale = 'zh-CN';

export function setClientLocale(locale: string): void {
  currentClientLocale = locale;
}

export function getClientLocale(): string {
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem('alonabot.webui.locale');
    if (saved) return saved;
  }
  return currentClientLocale;
}

/** 获取全部组件包列表。 */
export async function listPackages(): Promise<Package[]> {
  const lang = getClientLocale();
  return request<Package[]>('/api/v1/packages', {
    headers: { 'Accept-Language': lang },
  });
}

/** 返回组件图标的绝对路径 URL。 */
export function packageIconURL(id: string): string {
  return `/api/v1/packages/${encodeURIComponent(id)}/icon`;
}

/** 安装组件（远程 Source: "owner/repo" 或浏览器本地上传 FormData）。 */
export async function installPackage(
  input: { Source: string; Replace?: boolean; Confirm?: boolean } | FormData,
): Promise<Package> {
  if (input instanceof FormData) {
    return request<Package>('/api/v1/packages/install', {
      method: 'POST',
      body: input,
    });
  }

  return request<Package>('/api/v1/packages/install', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
}

/** 更新组件（JSON 沿用原来源或 FormData 上传更新包）。 */
export async function updatePackage(
  id: string,
  input: { confirm?: boolean } | FormData,
): Promise<{ warnings?: string[] }> {
  if (input instanceof FormData) {
    input.set('id', id);
    return request<{ warnings?: string[] }>('/api/v1/packages/update', {
      method: 'POST',
      body: input,
    });
  }

  return request<{ warnings?: string[] }>('/api/v1/packages/update', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id,
      confirm: input.confirm ?? false,
      install: {},
    }),
  });
}

/** 启停、重启或重载组件。 */
export async function changePackage(
  action: 'enable' | 'disable' | 'restart' | 'reload',
  id: string,
  confirm?: boolean,
): Promise<void> {
  const body: Record<string, unknown> = { id };
  if (action === 'disable') {
    body.confirm = confirm ?? false;
  }
  await request<void>(`/api/v1/packages/${action}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

/** 卸载组件。purge: 0 保留数据和配置，1 清除数据，2 清除数据和配置。 */
export async function uninstallPackage(id: string, purge: 0 | 1 | 2, confirm?: boolean): Promise<void> {
  await request<void>('/api/v1/packages/uninstall', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id,
      purge,
      confirm: confirm ?? false,
    }),
  });
}

/** 回退组件到指定代次。 */
export async function rollbackPackage(id: string, generation: string, confirm?: boolean): Promise<void> {
  await request<void>('/api/v1/packages/rollback', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id,
      generation,
      confirm: confirm ?? false,
    }),
  });
}

/** 获取全局及来源日志级别。 */
export async function getLogLevels(): Promise<LogLevels> {
  return request<LogLevels>('/api/v1/logs/levels');
}

/** 设置特定组件的日志级别。 */
export async function setPackageLogLevel(id: string, level: SourceLogLevel): Promise<LogLevels> {
  return request<LogLevels>('/api/v1/logs/levels', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      sources: { [id]: level },
    }),
  });
}

// ----------------------------------------------------------------------------
// Adapter 实例管理 API
// ----------------------------------------------------------------------------

/** 读取 Adapter 实例配置 Schema。 */
export async function getInstanceSchema(adapterID: string): Promise<unknown> {
  return request<unknown>(`/api/v1/adapters/${encodeURIComponent(adapterID)}/instance-schema`);
}

/** 读取指定 Adapter 实例的脱敏配置（凭证形如 {$secret_set:true}）。 */
export async function getInstanceConfig(
  adapterID: string,
  instanceID: string,
): Promise<ConfigDiagnostics & { config: Record<string, unknown> }> {
  return request<ConfigDiagnostics & { config: Record<string, unknown> }>(
    `/api/v1/adapters/${encodeURIComponent(adapterID)}/instances/${encodeURIComponent(instanceID)}`,
  );
}

/** 创建 Adapter 实例。 */
export async function createInstance(input: {
  package_id: string;
  instance_id: string;
  config: Record<string, unknown>;
  confirm?: boolean;
}): Promise<void> {
  await request<void>('/api/v1/adapter/instances/create', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      package_id: input.package_id,
      instance_id: input.instance_id,
      config: input.config,
      confirm: input.confirm ?? false,
    }),
  });
}

/** 启停或删除指定 Adapter 实例。 */
export async function changeInstance(action: 'start' | 'stop' | 'delete', instanceID: string): Promise<void> {
  await request<void>(`/api/v1/adapter/instances/${action}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      instance_id: instanceID,
    }),
  });
}

/** listGenerations 读取按提交时间降序排列的保留代次。 */
export async function listGenerations(id: string): Promise<Generation[]> {
  const result = await request<{ generations: Generation[] }>(
    `/api/v1/plugins/${encodeURIComponent(id)}/generations`,
  );
  return result.generations;
}

export type AdapterInstance = {
  instance_id: string;
  adapter_id: string;
  package_id: string;
  state: 'running' | 'stopped' | 'failed';
  safe_error: string;
  default: boolean;
};

/** listInstances 读取 Core 持久配置中的账号及当前运行状态。 */
export async function listInstances(): Promise<AdapterInstance[]> {
  const result = await request<{ instances: AdapterInstance[] }>('/api/v1/adapter/instances');
  return result.instances;
}

export type Chat = {
  session: string;
  title: string;
  chat_type: 'private' | 'group' | 'channel';
  instance_id: string;
  last_active_at: string;
};

/** listChats 读取完整聊天目录，再按实例归组，避免把单页或 Adapter 总数当成账号会话数。 */
export async function listChats(): Promise<Chat[]> {
  const chats: Chat[] = [];
  let total: number;
  do {
    const page = await request<{ chats: Chat[]; total: number }>(`/api/v1/chats?offset=${chats.length}&limit=50`);
    chats.push(...page.chats);
    total = page.total;
  } while (chats.length < total);
  return chats;
}

export type Profile = { id: string; name: string; color: string; sessions: number };
export type ProfileAssignments = {
  instances: Record<string, string>;
  sessions: { instance: string; session: string; profile: string }[];
};

/** listProfiles 读取配置文件选择项。 */
export function listProfiles(): Promise<Profile[]> {
  return request<Profile[]>('/api/v1/profiles');
}

/** getProfileAssignments 读取实例缺省与会话单独指定的配置文件。 */
export function getProfileAssignments(): Promise<ProfileAssignments> {
  return request<ProfileAssignments>('/api/v1/profiles/assignments');
}

/** assignInstanceProfile 保存已创建实例的配置文件指定。 */
export async function assignInstanceProfile(instanceID: string, profile: string): Promise<void> {
  await request('/api/v1/profiles/assignments/instances/' + encodeURIComponent(instanceID), {
    method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profile }),
  });
}

// ----------------------------------------------------------------------------
// 插件配置与诊断 API
// ----------------------------------------------------------------------------

export async function getPluginConfig(id: string): Promise<ConfigSnapshot> {
  return request<ConfigSnapshot>(`/api/v1/plugins/${encodeURIComponent(id)}/config`);
}

export async function patchPluginConfig(
  id: string,
  baseGeneration: string,
  patch: Record<string, unknown>,
): Promise<PendingConfig> {
  return request<PendingConfig>(`/api/v1/plugins/${encodeURIComponent(id)}/config/pending`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ base_generation: baseGeneration, patch }),
  });
}

export async function discardPluginConfig(id: string): Promise<void> {
  await request<void>(`/api/v1/plugins/${encodeURIComponent(id)}/config/pending`, {
    method: 'DELETE',
  });
}

export async function prunePluginConfig(id: string): Promise<PendingConfig> {
  return request<PendingConfig>(`/api/v1/plugins/${encodeURIComponent(id)}/config/prune`, {
    method: 'POST',
  });
}

export async function getCoreConfigDiagnostics(): Promise<ConfigDiagnostics> {
  return request<ConfigDiagnostics>('/api/v1/core/settings/corrections');
}

export async function pruneCoreConfig(): Promise<void> {
  await request<void>('/api/v1/core/settings/prune', {
    method: 'POST',
  });
}

export async function pruneInstanceConfig(adapterID: string, instanceID: string): Promise<void> {
  await request<void>(
    `/api/v1/adapters/${encodeURIComponent(adapterID)}/instances/${encodeURIComponent(instanceID)}/prune`,
    {
      method: 'POST',
    },
  );
}

/** 将实际值路径逐段将 ~ 转成 ~0、. 转成 ~1 后用点连接。数组下标占一段（§12.2）。 */
export function encodeSecretPath(segments: (string | number)[]): string {
  return segments.map(s => String(s).replace(/~/g, '~0').replace(/\./g, '~1')).join('.');
}

export async function revealPluginSecret(id: string, path: string, token: string): Promise<{ value: string }> {
  return request<{ value: string }>(`/api/v1/secrets/${encodeURIComponent(id)}/${encodeURIComponent(path)}/reveal`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token }),
  });
}

// ----------------------------------------------------------------------------
// 选项查询与示例 API
// ----------------------------------------------------------------------------

export async function listConfigOptions(
  id: string,
  path: string,
  q: string,
  locale: string,
): Promise<OptionsPage> {
  const query = new URLSearchParams({ path, q });
  return request<OptionsPage>(`/api/v1/plugins/${encodeURIComponent(id)}/config/options?${query.toString()}`, {
    headers: { 'Accept-Language': locale },
  });
}

export async function resolveConfigOptions(
  id: string,
  path: string,
  values: string[],
  locale: string,
): Promise<OptionsPage> {
  return request<OptionsPage>(`/api/v1/plugins/${encodeURIComponent(id)}/config/options/resolve`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept-Language': locale,
    },
    body: JSON.stringify({ path, values }),
  });
}

export async function listInstanceOptions(
  adapterID: string,
  instanceID: string,
  path: string,
  q: string,
  locale: string,
): Promise<OptionsPage> {
  const query = new URLSearchParams({ path, q });
  return request<OptionsPage>(
    `/api/v1/adapters/${encodeURIComponent(adapterID)}/instances/${encodeURIComponent(instanceID)}/options?${query.toString()}`,
    {
      headers: { 'Accept-Language': locale },
    },
  );
}

export async function resolveInstanceOptions(
  adapterID: string,
  instanceID: string,
  path: string,
  values: string[],
  locale: string,
): Promise<OptionsPage> {
  return request<OptionsPage>(
    `/api/v1/adapters/${encodeURIComponent(adapterID)}/instances/${encodeURIComponent(instanceID)}/options/resolve`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept-Language': locale,
      },
      body: JSON.stringify({ path, values }),
    },
  );
}

function buildFilterParams(params: URLSearchParams, filter: Record<string, string>): void {
  for (const [k, v] of Object.entries(filter)) {
    params.set(`filter[${k}]`, v);
  }
}

export async function listCoreOptions(
  source: 'core.chats' | 'core.adapter_instances',
  q: string,
  filter: Record<string, string>,
  locale: string,
): Promise<OptionsPage> {
  const params = new URLSearchParams({ q });
  buildFilterParams(params, filter);
  return request<OptionsPage>(`/api/v1/core/options/${encodeURIComponent(source)}?${params.toString()}`, {
    headers: { 'Accept-Language': locale },
  });
}

export async function resolveCoreOptions(
  source: 'core.chats' | 'core.adapter_instances',
  values: string[],
  filter: Record<string, string>,
  locale: string,
): Promise<OptionsPage> {
  const params = new URLSearchParams();
  buildFilterParams(params, filter);
  const qs = params.toString();
  const url = `/api/v1/core/options/${encodeURIComponent(source)}/resolve${qs ? `?${qs}` : ''}`;
  return request<OptionsPage>(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept-Language': locale,
    },
    body: JSON.stringify({ values }),
  });
}

export function pluginExampleURL(id: string, image: string): string {
  const stripped = image.startsWith('examples/') ? image.slice('examples/'.length) : image;
  const encodedSegments = stripped
    .split('/')
    .filter(Boolean)
    .map(seg => encodeURIComponent(seg))
    .join('/');
  return `/api/v1/packages/${encodeURIComponent(id)}/examples/${encodedSegments}`;
}

// ----------------------------------------------------------------------------
// 命令权限与名称 API
// ----------------------------------------------------------------------------

export async function getPluginPolicies(id: string): Promise<PluginPolicies> {
  return request<PluginPolicies>(`/api/v1/plugins/${encodeURIComponent(id)}/policies`);
}

export async function setPluginPolicy(id: string, override: CommandOverride): Promise<void> {
  await request<void>(`/api/v1/plugins/${encodeURIComponent(id)}/policies`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(override),
  });
}

export async function setCommandNames(id: string, routeID: string, names: string[][]): Promise<void> {
  await request<void>(
    `/api/v1/plugins/${encodeURIComponent(id)}/commands/${encodeURIComponent(routeID)}/names`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ names }),
    },
  );
}

export async function resetCommandNames(id: string, routeID: string): Promise<void> {
  await request<void>(
    `/api/v1/plugins/${encodeURIComponent(id)}/commands/${encodeURIComponent(routeID)}/names`,
    {
      method: 'DELETE',
    },
  );
}

// ----------------------------------------------------------------------------
// 工具、发信许可、会话与静态任务 API
// ----------------------------------------------------------------------------

export async function getToolPolicies(): Promise<{ policies: ToolPolicy[] }> {
  return request<{ policies: ToolPolicy[] }>('/api/v1/tools/policies');
}

export async function setToolPolicy(item: ToolPolicyUpdate): Promise<void> {
  await request<void>('/api/v1/tools/policies', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(item),
  });
}

export async function setToolPolicies(items: ToolPolicyUpdate[]): Promise<void> {
  await request<void>('/api/v1/tools/policies', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ items }),
  });
}

export async function getAllowedInstances(id: string): Promise<AllowedInstances> {
  return request<AllowedInstances>(`/api/v1/plugins/${encodeURIComponent(id)}/allowed-instances`);
}

export async function setAllowedInstances(id: string, value: AllowedInstances): Promise<void> {
  await request<void>(`/api/v1/plugins/${encodeURIComponent(id)}/allowed-instances`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(value),
  });
}

export async function listPluginConversations(
  id: string,
): Promise<{ key: string; entries: number; updated_at: string }[]> {
  return request<{ key: string; entries: number; updated_at: string }[]>(
    `/api/v1/plugins/${encodeURIComponent(id)}/conversations`,
  );
}

export async function clearPluginConversation(id: string, key: string): Promise<void> {
  await request<void>(
    `/api/v1/plugins/${encodeURIComponent(id)}/conversations/${encodeURIComponent(key)}`,
    {
      method: 'DELETE',
    },
  );
}

export async function listPluginJobs(id: string): Promise<{ jobs: PluginJob[] }> {
  return request<{ jobs: PluginJob[] }>(`/api/v1/plugins/${encodeURIComponent(id)}/jobs`);
}

export async function changePluginJob(
  id: string,
  jobID: string,
  action: 'pause' | 'resume',
): Promise<void> {
  await request<void>(
    `/api/v1/plugins/${encodeURIComponent(id)}/jobs/${encodeURIComponent(jobID)}/${action}`,
    {
      method: 'POST',
    },
  );
}

// ----------------------------------------------------------------------------
// Core 设置与管理认证 API
// ----------------------------------------------------------------------------

export async function getCoreAdmins(): Promise<string[]> {
  return request<string[]>('/api/v1/admins');
}

export async function changeCoreAdmin(action: 'add' | 'remove', user: string): Promise<void> {
  await request<void>(`/api/v1/admins/${action}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user }),
  });
}

export async function getCoreDenies(): Promise<{ users: string[]; chats: string[] }> {
  return request<{ users: string[]; chats: string[] }>('/api/v1/core/denies');
}

export async function setCoreDenies(value: { users: string[]; chats: string[] }): Promise<void> {
  await request<void>('/api/v1/core/denies', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(value),
  });
}

export async function getDefaultLocale(): Promise<{ default_locale: string }> {
  return request<{ default_locale: string }>('/api/v1/core/default-locale');
}

export async function setDefaultLocale(locale: string): Promise<void> {
  await request<void>('/api/v1/core/default-locale', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ default_locale: locale }),
  });
}

export async function getDefaultPlatform(): Promise<{
  command_prefixes: string[];
  reply: { mention_sender: boolean; quote: boolean; prefix: string; forward_threshold: number };
}> {
  return request<{
    command_prefixes: string[];
    reply: { mention_sender: boolean; quote: boolean; prefix: string; forward_threshold: number };
  }>('/api/v1/profiles/default/platform');
}

export async function setGlobalLogLevel(level: GlobalLogLevel): Promise<LogLevels> {
  return request<LogLevels>('/api/v1/logs/levels', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ global: level }),
  });
}

export async function rotateAdminToken(token: string): Promise<{ token: string }> {
  const result = await request<{ token: string }>('/api/auth/token/rotate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });
  setCSRFToken(null);
  return result;
}

// ----------------------------------------------------------------------------
// 插件页面代理 API
// ----------------------------------------------------------------------------

export async function requestPluginPage(
  id: string,
  input: { method: string; path: string; headers?: Record<string, string>; body_base64?: string },
): Promise<{ status: number; headers: Record<string, string>; body_base64: string }> {
  const headers = new Headers(input.headers);
  if (input.method !== 'GET' && input.method !== 'HEAD' && currentCSRFToken) headers.set('X-Alonabot-CSRF', currentCSRFToken);
  const body = input.body_base64 === undefined ? undefined : Uint8Array.from(atob(input.body_base64), char => char.charCodeAt(0));
  const response = await fetch(`/plugins/${encodeURIComponent(id)}/${input.path}`, {
    method: input.method, headers, body, credentials: 'same-origin', redirect: 'error',
  });
  if (response.status === 401) notifySessionExpired();
  const bytes = new Uint8Array(await response.arrayBuffer());
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return { status: response.status, headers: Object.fromEntries(response.headers), body_base64: btoa(binary) };
}
