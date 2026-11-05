import { expect, it } from 'vitest';
import nested from '../../fixtures/nested.config_schema.json';
import secrets from '../../fixtures/secrets.config_schema.json';
import collections from '../../fixtures/collections.config_schema.json';
import templates from '../../fixtures/templates.config_schema.json';

it('nested fixture retains field order', () => {
  expect(Object.keys(nested)).toEqual(["welcome", "reply", "level", "advanced", "template"]);
});

it('secrets fixture retains field order', () => {
  expect(Object.keys(secrets)).toEqual(["token", "backup", "internal", "endpoint"]);
});

it('collections fixture retains field order', () => {
  expect(Object.keys(collections)).toEqual(["tags", "keywords", "accounts", "headers", "ratio"]);
});

it('templates fixture retains field order', () => {
  expect(Object.keys(templates)).toEqual(["music", "sources"]);
});
