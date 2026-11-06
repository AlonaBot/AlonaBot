import { describe, expect, it } from 'vitest';
import nested from '../../fixtures/nested.config_schema.json';
import secrets from '../../fixtures/secrets.config_schema.json';
import collections from '../../fixtures/collections.config_schema.json';
import templates from '../../fixtures/templates.config_schema.json';
import chinese from '../../fixtures/i18n/zh-CN.json';
import { fieldsFromSchema } from './model';
import { buildPatch, isSecretSet, MissingSecretsError } from './patch';
import { fieldText, pickLocale, templateText } from '../i18n/catalog';

describe('config_schema.json form contract (V-25)', () => {
  it('preserves nested field order and control metadata', () => {
    const fields = fieldsFromSchema(nested);
    expect(fields.map(field => field.key)).toEqual(['welcome', 'reply', 'level', 'advanced', 'template']);
    expect(fields[0]).toMatchObject({ required: true, defaultValue: 'Hello', invisible: false, secret: false });
    expect(fields[1].children?.map(field => field.key)).toEqual(['prefix', 'mode']);
    expect(fields[1].children?.[1]).toMatchObject({ options: ['fast', 'safe'], labels: ['Fast', 'Safe'], schemaPath: ['reply', 'items', 'mode'] });
    expect(fields[2].slider).toEqual({ min: 1, max: 5, step: 1 });
    expect(fields[4]).toMatchObject({ condition: { advanced: true }, editorMode: true, editorLanguage: 'json', obviousHint: true });
  });

  it('retains invisible fields and list element fields', () => {
    expect(fieldsFromSchema(secrets).map(field => field.key)).toEqual(['token', 'backup', 'internal', 'endpoint']);
    expect(fieldsFromSchema(secrets)[2]).toMatchObject({ invisible: true, defaultValue: 'v1' });
    const fields = fieldsFromSchema(collections);
    expect(fields.map(field => field.key)).toEqual(['tags', 'keywords', 'accounts', 'headers', 'ratio']);
    expect(fields[2].children?.map(field => field.key)).toEqual(['name', 'api_key']);
    expect(fields[2].children?.[1]).toMatchObject({ secret: true, schemaPath: ['accounts', 'items', 'api_key'] });
    expect(fields[4].slider).toEqual({ min: 0, max: 1, step: 0.1 });
  });

  it('retains template order and display metadata with complete schema paths', () => {
    const fields = fieldsFromSchema(templates);
    expect(fields.map(field => field.key)).toEqual(['music', 'sources']);
    expect(fields[0].templates?.map(template => template.key)).toEqual(['netease', 'local']);
    expect(fields[1].templates?.map(template => template.key)).toEqual(['rss', 'webhook']);
    expect(fields[0].templates?.[0].children[0].schemaPath).toEqual(['music', 'templates', 'netease', 'items', 'cookie']);
    expect(fields[1].templates?.[0].displayItem).toBe('url');
    expect(fields[1].templates?.[1].hideHintInList).toBe(true);
  });

  it.each([
    [{ field: { type: 'file' } }, 'field.type'],
    [{ field: { type: 'string', _special: 'x' } }, 'field._special'],
    [{ field: { type: 'dict', template_schema: {} } }, 'field.template_schema'],
    [{ field: { type: 'string', editor_theme: 'dark' } }, 'field.editor_theme'],
    [{ field: { type: 'object', items: { child: { type: 'unknown' } } } }, 'field.items.child.type'],
    [{ field: { type: 'template', templates: { first: { items: {}, unknown: true } } } }, 'field.templates.first.unknown'],
    [{ field: { type: 'int', slider: { min: 0, max: 1, unknown: 2 } } }, 'field.slider.unknown'],
  ])('rejects unsupported schema at its path', (schema, path) => {
    expect(() => fieldsFromSchema(schema)).toThrow(path);
  });
});

describe('form text and locale contract (V-54)', () => {
  it('resolves locale preferences with deterministic primary-language fallback', () => {
    const available = ['zh-TW', 'zh-CN', 'en-US'];
    expect(pickLocale(available, ['zh-TW'])).toBe('zh-TW');
    expect(pickLocale(available, ['zh-HK'])).toBe('zh-CN');
    expect(pickLocale(available, ['fr-FR', 'zh-HK'])).toBe('zh-CN');
    expect(pickLocale(available, ['ZH-cn'])).toBe('zh-CN');
    expect(pickLocale(available, ['fr-FR'])).toBe('en-US');
    expect(pickLocale(available, [])).toBe('en-US');
    expect(available).toEqual(['zh-TW', 'zh-CN', 'en-US']);
  });

  it('uses same-shaped field overrides, then source text, then keys', () => {
    const fields = fieldsFromSchema(nested);
    const mode = fields[1].children![1];
    expect(fieldText(chinese, mode)).toEqual({ description: '模式', labels: ['快速', '安全'] });
    expect(fieldText({}, mode)).toEqual({ description: 'Mode', labels: ['Fast', 'Safe'] });
    expect(fieldText(chinese, fields[1].children![0])).toEqual({ description: 'Prefix', hint: 'Text placed before every reply' });
    expect(fieldText({}, fieldsFromSchema(secrets)[2])).toEqual({ description: 'internal' });
    expect(fieldText(chinese, fieldsFromSchema(secrets)[3])).toEqual({ description: 'Endpoint', hint: '服务的基础地址' });
    expect(fieldText(chinese, fieldsFromSchema(collections)[2].children![1]).description).toBe('API 密钥');
  });

  it('uses template overrides and nested field paths without translating stored values', () => {
    const music = fieldsFromSchema(templates)[0];
    const netease = music.templates![0];
    expect(templateText(chinese, netease)).toEqual({ name: '网易云音乐', hint: 'Log in with a cookie' });
    expect(templateText({}, netease)).toEqual({ name: 'NetEase Cloud Music', hint: 'Log in with a cookie' });
    expect(fieldText(chinese, netease.children[0]).description).toBe('Cookie');
    const unnamed = fieldsFromSchema({ field: { type: 'template', templates: { simple: { items: {} } } } })[0].templates![0];
    expect(templateText({}, unnamed)).toEqual({ name: 'simple' });
    expect(fieldsFromSchema(nested)[1].children![1].options).toEqual(['fast', 'safe']);
  });
});

describe('form Merge Patch contract (V-25)', () => {
  const setSecret = { $secret_set: true };

  it('omits invisible and unknown fields and unchanged secrets', () => {
    const original = { token: setSecret, backup: '', internal: 'v1', endpoint: 'https://example.com', unknown: { keep: 1 } };
    const edited = { ...original, internal: 'v2', unknown: { keep: 2 }, endpoint: 'https://new.example' };
    expect(buildPatch(original, edited, fieldsFromSchema(secrets))).toEqual({ endpoint: 'https://new.example' });
    expect(buildPatch(original, { ...edited, token: 'new token' }, fieldsFromSchema(secrets))).toEqual({ token: 'new token', endpoint: 'https://new.example' });
    expect(original.token).toEqual({ $secret_set: true });
  });

  it('uses null for explicit default restoration and secret clearing', () => {
    expect(buildPatch({ token: setSecret, endpoint: 'https://custom.example' }, { token: null, endpoint: null }, fieldsFromSchema(secrets)))
      .toEqual({ token: null, endpoint: null });
    expect(buildPatch({ level: 4 }, { level: null }, fieldsFromSchema(nested))).toEqual({ level: null });
  });

  it('replaces arrays and recursively deletes dict keys', () => {
    const original = { tags: ['news'], keywords: ['one'], headers: { A: 'one', B: 'two', nested: { keep: true, remove: 1 } } };
    const edited = { tags: ['sports', 'tech'], keywords: [], headers: { A: 'one', nested: { keep: true } } };
    expect(buildPatch(original, edited, fieldsFromSchema(collections))).toEqual({
      tags: ['sports', 'tech'], keywords: [], headers: { B: null, nested: { remove: null } },
    });
  });

  it.each(['{"__proto__":{}}', '{"nested":{"__proto__":{}}}'])('preserves new empty dict objects with own data keys: %s', json => {
    const original = { headers: {} };
    const edited = { headers: JSON.parse(json) };
    const snapshot = JSON.stringify({ original, edited });
    expect(JSON.stringify(buildPatch(original, edited, fieldsFromSchema(collections)))).toBe(JSON.stringify(edited));
    expect(JSON.stringify({ original, edited })).toBe(snapshot);
  });

  it('preserves and changes existing own dict data keys', () => {
    const original = JSON.parse('{"headers":{"__proto__":{"before":1},"constructor":"keep"}}');
    const edited = JSON.parse('{"headers":{"__proto__":{},"constructor":"keep"}}');
    const snapshot = JSON.stringify({ original, edited });
    const fields = fieldsFromSchema(collections);
    expect(buildPatch(original, structuredClone(original), fields)).toEqual({});
    expect(JSON.stringify(buildPatch(original, edited, fields)))
      .toBe('{"headers":{"__proto__":{"before":null}}}');
    expect(JSON.stringify({ original, edited })).toBe(snapshot);
  });

  it('treats missing Schema data keys as absent when creating dict values', () => {
    const fields = fieldsFromSchema({ ['__proto__']: collections.headers });
    const edited = JSON.parse('{"__proto__":{}}');
    expect(JSON.stringify(buildPatch({}, edited, fields))).toBe('{"__proto__":{}}');
    expect(JSON.stringify(edited)).toBe('{"__proto__":{}}');
  });

  it('omits restored Schema fields in replacement rows while retaining dict JSON nulls', () => {
    const original = { sources: [{ rss: { url: 'https://feed.example', every_minutes: 15 } }] };
    const edited = { sources: [{ rss: { url: 'https://feed.example', every_minutes: null } }] };
    expect(buildPatch(original, edited, fieldsFromSchema(templates)))
      .toEqual({ sources: [{ rss: { url: 'https://feed.example' } }] });
    expect(edited.sources[0].rss.every_minutes).toBeNull();

    const schema = fieldsFromSchema({ rows: { type: 'list', items: {
      ...collections, ...nested, choice: templates.music, sources: templates.sources,
    } } });
    const row = { ratio: null, reply: { prefix: null, mode: 'safe' },
      accounts: [{ name: null, api_key: 'new' }], headers: { arbitrary: { keep: null } },
      choice: { local: { dir: null } }, sources: [{ rss: { url: 'https://feed.example', every_minutes: null } }],
      unknown: { keep: null } };
    expect(buildPatch({ rows: [] }, { rows: [row] }, schema)).toEqual({ rows: [{
      reply: { mode: 'safe' }, accounts: [{ api_key: 'new' }], headers: { arbitrary: { keep: null } },
      choice: { local: {} }, sources: [{ rss: { url: 'https://feed.example' } }], unknown: { keep: null },
    }] });
    expect(row.ratio).toBeNull();
  });

  it('omits conditionally hidden edits while preserving explicit scalar values', () => {
    const original = { advanced: true, template: 'old', welcome: 'Hello', reply: { prefix: '> ', mode: 'fast', unknown: true } };
    expect(buildPatch(original, { ...original, advanced: false, template: 'edited', welcome: '', reply: { ...original.reply, prefix: '', unknown: false } }, fieldsFromSchema(nested)))
      .toEqual({ advanced: false, welcome: '', reply: { prefix: '' } });
    expect(buildPatch({ ratio: 0.5 }, { ratio: 0 }, fieldsFromSchema(collections))).toEqual({ ratio: 0 });
  });

  it('omits unchanged masked arrays but reports every missing account secret after an edit', () => {
    const original = { accounts: [{ name: 'one', api_key: setSecret }, { name: 'two', api_key: setSecret }] };
    expect(buildPatch(original, structuredClone(original), fieldsFromSchema(collections))).toEqual({});
    const edited = structuredClone(original);
    edited.accounts[0].name = 'changed';
    try {
      buildPatch(original, edited, fieldsFromSchema(collections));
      throw new Error('expected MissingSecretsError');
    } catch (error) {
      expect(error).toBeInstanceOf(MissingSecretsError);
      expect((error as MissingSecretsError).paths).toEqual(['/accounts/0/api_key', '/accounts/1/api_key']);
    }
    expect(buildPatch(original, { accounts: [{ name: 'changed', api_key: 'first' }, { name: 'two', api_key: 'second' }] }, fieldsFromSchema(collections)))
      .toEqual({ accounts: [{ name: 'changed', api_key: 'first' }, { name: 'two', api_key: 'second' }] });
  });

  it('requires secrets for edited template lists and permits deleting masked rows', () => {
    const original = { sources: [{ webhook: { token: setSecret } }] };
    expect(() => buildPatch(original, { sources: [...original.sources, { rss: { url: 'https://feed.example', every_minutes: 30 } }] }, fieldsFromSchema(templates)))
      .toThrow(MissingSecretsError);
    expect(buildPatch(original, { sources: [] }, fieldsFromSchema(templates))).toEqual({ sources: [] });
    expect(buildPatch(original, { sources: [{ webhook: { token: 'new value' } }] }, fieldsFromSchema(templates)))
      .toEqual({ sources: [{ webhook: { token: 'new value' } }] });
  });

  it('removes the old template key when selecting a new template', () => {
    const original = { music: { netease: { cookie: setSecret } } };
    expect(buildPatch(original, { music: { local: { dir: 'music' } } }, fieldsFromSchema(templates)))
      .toEqual({ music: { netease: null, local: { dir: 'music' } } });
    expect(buildPatch({}, { music: { local: {} } }, fieldsFromSchema(templates))).toEqual({ music: { local: {} } });
  });

  it('merges secret-bearing objects by key without resubmitting masked siblings', () => {
    const schema = fieldsFromSchema({ object: { type: 'object', items: secrets } });
    const original = { object: { token: setSecret, internal: 'v1', endpoint: 'before', unknown: 'keep' } };
    expect(buildPatch(original, { object: { token: setSecret, internal: 'changed', endpoint: 'after', unknown: 'changed' } }, schema))
      .toEqual({ object: { endpoint: 'after' } });
    const dict = { headers: { auth: setSecret, nested: { token: setSecret, plain: 'before' } } };
    expect(buildPatch(dict, { headers: { auth: setSecret, nested: { token: setSecret, plain: 'after' } } }, fieldsFromSchema(collections)))
      .toEqual({ headers: { nested: { plain: 'after' } } });
  });

  it.each([
    { before: { auth: setSecret }, after: { authorization: setSecret }, path: '/headers/authorization',
      resolved: { authorization: 'new value' }, patch: { auth: null, authorization: 'new value' } },
    { before: {}, after: { nested: { 'a/b~c': setSecret } }, path: '/headers/nested/a~1b~0c',
      resolved: { nested: { 'a/b~c': 'new value' } }, patch: { nested: { 'a/b~c': 'new value' } } },
  ])('requires a new secret value at a new dict path: $path', ({ before, after, path, resolved, patch }) => {
    const original = { headers: before };
    const edited = { headers: after };
    const snapshot = structuredClone({ original, edited });
    const fields = fieldsFromSchema(collections);
    try {
      buildPatch(original, edited, fields);
      throw new Error('expected MissingSecretsError');
    } catch (error) {
      expect(error).toBeInstanceOf(MissingSecretsError);
      expect((error as MissingSecretsError).paths).toEqual([path]);
    }
    expect({ original, edited }).toEqual(snapshot);
    expect(buildPatch(original, { headers: resolved }, fields)).toEqual({ headers: patch });
  });

  it('recognizes only the exact secret marker and escapes missing value paths', () => {
    expect(isSecretSet(setSecret)).toBe(true);
    for (const value of [null, true, '$secret_set', { $secret_set: false }, { $secret_set: true, other: 1 }]) {
      expect(isSecretSet(value)).toBe(false);
    }
    const original = { headers: { rows: [] } };
    expect(() => buildPatch(original, { headers: { rows: [{ 'a/b~c': setSecret }] } }, fieldsFromSchema(collections)))
      .toThrow('missing secret values at /headers/rows/0/a~1b~0c');
  });

  it.each([
    ['value', 'value'], [false, false], [0, 0], ['', ''],
    [[null, { keep: null }], [null, { keep: null }]],
    [{ token: setSecret, keep: null }, {}],
  ])('disambiguates a marker-shaped dict difference with an unchanged key', (keep, preserved) => {
    const original = { headers: { nested: { keep } } };
    const edited = { headers: { nested: { keep, $secret_set: true } } };
    expect(buildPatch(original, edited, fieldsFromSchema(collections)))
      .toEqual({ headers: { nested: { $secret_set: true, keep: preserved } } });
    expect(original).toEqual({ headers: { nested: { keep } } });
    expect(edited).toEqual({ headers: { nested: { keep, $secret_set: true } } });
  });

  it('rejects marker-shaped differences when every unchanged key requires a masked value or deletion', () => {
    for (const keep of [setSecret, [setSecret], null]) {
      const original = { headers: { nested: { keep } } };
      const edited = { headers: { nested: { keep, $secret_set: true } } };
      expect(() => buildPatch(original, edited, fieldsFromSchema(collections)))
        .toThrow('/headers/nested: patch would contain a reserved secret marker');
      expect(original).toEqual({ headers: { nested: { keep } } });
      expect(edited).toEqual({ headers: { nested: { keep, $secret_set: true } } });
    }
  });
});
