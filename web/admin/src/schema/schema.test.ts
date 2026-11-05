import { describe, expect, it } from 'vitest';
import nested from '../../fixtures/nested.config_schema.json';
import secrets from '../../fixtures/secrets.config_schema.json';
import collections from '../../fixtures/collections.config_schema.json';
import templates from '../../fixtures/templates.config_schema.json';
import chinese from '../../fixtures/i18n/zh-CN.json';
import { fieldsFromSchema } from './model';
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
