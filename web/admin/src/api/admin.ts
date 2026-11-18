/**
 * Admin API 客户端与数据结构定义。
 * 遵循 spec/ADMIN-API.md 规范与 CONTRACTS §7.3～§7.5、§9.3、§12.1。
 */

export class AdminAPIError extends Error {
  status: number;
  code?: string;
  affected?: string[];
  responseBody?: string;

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
  pages?: unknown[];
};

export type Generation = {
  id: string;
  version: string;
  created_at: string;
  active: boolean;
  has_db: boolean;
};

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
      const result = new AdminAPIError(response.status, message, code, affected);
      result.responseBody = text;
      return result;
    }
  }
  return new AdminAPIError(response.status, text || `HTTP ${response.status}`);
}

interface RequestOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: string | FormData | Blob;
  response?: 'blob';
}

async function request<T = unknown>(path: string, options: RequestOptions = {}): Promise<T> {
  if (!path.startsWith('/') || path.startsWith('//')) throw new Error('Expected a same-origin path');
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

  if (options.response === 'blob') return await response.blob() as T;

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

/** 获取全部组件包列表。 */
export async function listPackages(): Promise<Package[]> {
  return request<Package[]>('/api/v1/packages');
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
): Promise<{ config: Record<string, unknown> }> {
  return request<{ config: Record<string, unknown> }>(
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

/** adminRequest 共享同源请求的管理 Cookie、CSRF 与认证过期处理。 */
export function adminRequest<T>(path: string, options?: { method?: string; headers?: Record<string, string>; body?: string | FormData | Blob }): Promise<T> {
  return request<T>(path, options);
}

const pluginPath = (id: string) => `/api/v1/plugins/${encodeURIComponent(id)}`;
const jsonBody = (body: unknown, method = 'POST'): RequestOptions => ({ method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });

export type SQLiteValue = string | number | boolean | null | { $base64: string };
export type DataColumn = { name: string; type: string; pk: boolean; notnull: boolean };
export type DataTable = { name: string; rows: number; writable: boolean; columns: DataColumn[] };
export type DataRow = { rowid: number | null; values: Record<string, SQLiteValue> };
export function listPluginTables(id: string, generation = ''): Promise<{ tables: DataTable[] }> {
  return request(`${pluginPath(id)}/data/tables?${new URLSearchParams({ generation })}`);
}
export function listPluginRows(id: string, table: string, input: { offset: number; limit: number; generation?: string; order?: string; desc?: boolean; where?: Record<string, string> }): Promise<{ rows: DataRow[]; total: number }> {
  const query = new URLSearchParams({ offset: String(input.offset), limit: String(input.limit), generation: input.generation ?? '' });
  if (input.order) query.set('order', input.order);
  if (input.desc) query.set('desc', '1');
  for (const [column, value] of Object.entries(input.where ?? {})) query.set(`where.${column}`, value);
  return request(`${pluginPath(id)}/data/tables/${encodeURIComponent(table)}/rows?${query}`);
}
export function insertPluginRow(id: string, table: string, values: Record<string, SQLiteValue>): Promise<{ rowid: number }> {
  return request(`${pluginPath(id)}/data/tables/${encodeURIComponent(table)}/rows`, jsonBody({ values }));
}
export function updatePluginRow(id: string, table: string, rowid: number, values: Record<string, SQLiteValue>): Promise<void> {
  return request(`${pluginPath(id)}/data/tables/${encodeURIComponent(table)}/rows/${rowid}`, jsonBody({ values }, 'PATCH'));
}
export function deletePluginRow(id: string, table: string, rowid: number): Promise<void> {
  return request(`${pluginPath(id)}/data/tables/${encodeURIComponent(table)}/rows/${rowid}`, { method: 'DELETE' });
}

export function listPluginKV(id: string, input: { prefix: string; offset: number; limit: number }): Promise<{ items: { key: string; size: number; updated_at: string }[]; total: number }> {
  return request(`${pluginPath(id)}/data/kv?${new URLSearchParams({ prefix: input.prefix, offset: String(input.offset), limit: String(input.limit) })}`);
}
export function getPluginKV(id: string, key: string): Promise<{ value_base64: string; text?: string }> { return request(`${pluginPath(id)}/data/kv/${encodeURIComponent(key)}`); }
export function setPluginKV(id: string, key: string, value: { text: string } | { value_base64: string }): Promise<void> { return request(`${pluginPath(id)}/data/kv/${encodeURIComponent(key)}`, jsonBody(value, 'PUT')); }
export function deletePluginKV(id: string, key: string): Promise<void> { return request(`${pluginPath(id)}/data/kv/${encodeURIComponent(key)}`, { method: 'DELETE' }); }
export function listPluginFiles(id: string, path: string): Promise<{ entries: { name: string; dir: boolean; size: number; modified_at: string }[] }> { return request(`${pluginPath(id)}/data/files?${new URLSearchParams({ path })}`); }
export function downloadPluginFile(id: string, path: string): Promise<Blob> { return request(`${pluginPath(id)}/data/files/content?${new URLSearchParams({ path })}`, { response: 'blob' }); }
export function uploadPluginFile(id: string, path: string, file: File, overwrite: boolean): Promise<void> { return request(`${pluginPath(id)}/data/files/content?${new URLSearchParams({ path, ...(overwrite ? { overwrite: '1' } : {}) })}`, { method: 'PUT', body: file }); }
export function makePluginDirectory(id: string, path: string): Promise<void> { return request(`${pluginPath(id)}/data/files/mkdir`, jsonBody({ path })); }
export function renamePluginFile(id: string, from: string, to: string): Promise<void> { return request(`${pluginPath(id)}/data/files/rename`, jsonBody({ from, to })); }
export function deletePluginFile(id: string, path: string, recursive: boolean): Promise<void> { return request(`${pluginPath(id)}/data/files/delete`, jsonBody({ path, recursive })); }
export function exportPluginData(id: string): Promise<Blob> { return request(`${pluginPath(id)}/data/export`, { response: 'blob' }); }
export function clearPluginData(id: string, confirm: true): Promise<void> { return request(`${pluginPath(id)}/data/clear`, jsonBody({ confirm })); }
export function getCoreAdmins(): Promise<string[]> { return request('/api/v1/admins'); }
export function changeCoreAdmin(action: 'add' | 'remove', user: string): Promise<void> { return request(`/api/v1/admins/${action}`, jsonBody({ user })); }
export function getCoreDenies(): Promise<{ users: string[]; chats: string[] }> { return request('/api/v1/core/denies'); }
export function setCoreDenies(value: { users: string[]; chats: string[] }): Promise<void> { return request('/api/v1/core/denies', jsonBody(value, 'PUT')); }
export function listPluginConversations(id: string): Promise<{ key: string; entries: number; updated_at: string }[]> { return request(`${pluginPath(id)}/conversations`); }
export function clearPluginConversation(id: string, key: string): Promise<void> { return request(`${pluginPath(id)}/conversations/${encodeURIComponent(key)}`, { method: 'DELETE' }); }
export type Changelog = { locale: string; entries: { version: string; date?: string; markdown: string }[] };
export function getPackageChangelog(id: string, locale: string): Promise<Changelog> { return request(`/api/v1/packages/${encodeURIComponent(id)}/changelog`, { headers: { 'Accept-Language': locale } }); }
export function getCoreChangelog(locale: string): Promise<Changelog> { return request('/api/v1/core/changelog', { headers: { 'Accept-Language': locale } }); }
