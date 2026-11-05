import type { Field } from './model';

/** MissingSecretsError 列出整体替换数组时必须重新输入的 secret 值路径（JSON Pointer）。 */
export class MissingSecretsError extends Error {
  constructor(public readonly paths: string[]) {
    super(`missing secret values at ${paths.join(', ')}`);
    this.name = 'MissingSecretsError';
  }
}

function isObject(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/** isSecretSet 只接受恰好包含 $secret_set: true 的对象。 */
export function isSecretSet(value: unknown): boolean {
  return isObject(value) && Object.keys(value).length === 1 && Object.hasOwn(value, '$secret_set') && value.$secret_set === true;
}

function equal(left: unknown, right: unknown): boolean {
  if (left === right) return true;
  if (Array.isArray(left) && Array.isArray(right)) {
    return left.length === right.length && left.every((value, index) => equal(value, right[index]));
  }
  if (isObject(left) && isObject(right)) {
    const keys = Object.keys(left);
    return keys.length === Object.keys(right).length && keys.every(key => Object.hasOwn(right, key) && equal(left[key], right[key]));
  }
  return false;
}

function pointer(path: string[]): string {
  return '/' + path.map(part => part.replaceAll('~', '~0').replaceAll('/', '~1')).join('/');
}

function missingSecrets(value: unknown, path: string[], missing: string[]) {
  if (isSecretSet(value)) {
    missing.push(pointer(path));
  } else if (Array.isArray(value)) {
    value.forEach((entry, index) => missingSecrets(entry, [...path, String(index)], missing));
  } else if (isObject(value)) {
    for (const [key, entry] of Object.entries(value)) missingSecrets(entry, [...path, key], missing);
  }
}

// undefined 只表示此节点无需补丁，null 保留 RFC 7396 的删除含义。
function diff(original: unknown, edited: unknown, path: string[], missing: string[]): unknown {
  if (equal(original, edited) || isSecretSet(edited)) return undefined;
  if (Array.isArray(edited)) {
    missingSecrets(edited, path, missing);
    return edited;
  }
  if (isObject(edited)) {
    const before = isObject(original) ? original : {};
    const entries: [string, unknown][] = [];
    for (const key of new Set([...Object.keys(before), ...Object.keys(edited)])) {
      const change = Object.hasOwn(edited, key) ? diff(before[key], edited[key], [...path, key], missing) : null;
      if (change !== undefined) entries.push([key, change]);
    }
    const patch = Object.fromEntries(entries);
    // 新的空对象也要写入；已有对象无变化时省略整个节点。
    return entries.length || !isObject(original) ? patch : undefined;
  }
  return edited;
}

function fieldDiff(original: unknown, edited: unknown, field: Field, path: string[], missing: string[]): unknown {
  if (edited === null || isSecretSet(edited) || equal(original, edited)) {
    return diff(original, edited, path, missing);
  }
  if (field.secret) {
    if (typeof edited !== 'string') throw new Error(`${pointer(path)}: expected a secret string`);
    return edited;
  }
  if (field.type === 'object') {
    return fieldsDiff(isObject(original) ? original : {}, edited as Record<string, unknown>, field.children!, path, missing);
  }
  if (field.type === 'template') {
    const before = isObject(original) ? original : {};
    const after = edited as Record<string, unknown>;
    const entries: [string, unknown][] = [];
    for (const template of field.templates!) {
      const key = template.key;
      if (Object.hasOwn(after, key)) {
        const change = fieldsDiff(isObject(before[key]) ? before[key] : {}, after[key] as Record<string, unknown>, template.children, [...path, key], missing);
        if (Object.keys(change).length || !Object.hasOwn(before, key)) entries.push([key, change]);
      } else if (Object.hasOwn(before, key)) entries.push([key, null]);
    }
    return entries.length ? Object.fromEntries(entries) : undefined;
  }
  return diff(original, edited, path, missing);
}

function fieldsDiff(original: Record<string, unknown>, edited: Record<string, unknown>, fields: Field[], path: string[], missing: string[]): Record<string, unknown> {
  const entries: [string, unknown][] = [];
  for (const field of fields) {
    if (field.invisible || !Object.hasOwn(edited, field.key)) continue;
    if (field.condition && !Object.entries(field.condition).every(([key, expected]) => equal(edited[key], expected))) continue;
    const change = fieldDiff(original[field.key], edited[field.key], field, [...path, field.key], missing);
    if (change !== undefined && !(field.type === 'object' && isObject(change) && Object.keys(change).length === 0)) {
      entries.push([field.key, change]);
    }
  }
  return Object.fromEntries(entries);
}

/** buildPatch 生成已知可见字段的 RFC 7396 补丁，数组有残留 secret 状态时抛错。 */
export function buildPatch(original: unknown, edited: unknown, fields: Field[]): Record<string, unknown> {
  if (!isObject(original) || !isObject(edited)) throw new Error('configuration must be an object');
  const missing: string[] = [];
  const patch = fieldsDiff(original, edited, fields, [], missing);
  if (missing.length) throw new MissingSecretsError(missing);
  return patch;
}
