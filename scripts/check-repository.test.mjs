// 本文件验证维护入口检查器能拒绝失效路径和技能名漂移，并保留合法的相对引用。
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { checkLinks, checkSkill } from './check-repository.mjs';

// 验证路径检查的成功与失败边界，不读取开发者的真实文件。
test('维护链接保留仓库内相对路径，拒绝失效路径及宿主机路径', t => {
  // 1. 用隔离的文件树建立引用目标，结束时清理测试自己的数据。
  const root = mkdtempSync(path.join(os.tmpdir(), 'alonabot-docs-'));
  t.after(() => rmSync(root, { recursive: true }));
  mkdirSync(path.join(root, 'docs'));
  writeFileSync(path.join(root, 'AGENTS.md'), '# 工作地图');
  const file = path.join(root, 'docs/index.md');
  // 2. 分别检查有效引用、失效/越界引用与不应被解析的代码示例。
  assert.deepEqual(checkLinks('[工作地图](../AGENTS.md#入口) [网页](https://example.com)', file, root), []);
  assert.equal(checkLinks('[缺失](missing.md) [越界](../../outside.md)', file, root).length, 2);
  assert.equal(checkLinks(`[绝对路径](${path.join(root, 'AGENTS.md')})`, file, root).length, 1);
  assert.deepEqual(checkLinks('```md\n[仅为示例](missing.md)\n```', file, root), []);
});

// 验证技能名称和提示始终指向当前入口，不以存在一个 SKILL.md 冒充有效路由。
test('技能名称及 UI 提示必须与目录一致', () => {
  const text = '---\nname: alonabot-test\ndescription: 选择实际变更需要的测试。\n---\n# 测试';
  assert.deepEqual(checkSkill(text, 'alonabot-test', 'default_prompt: "使用 $alonabot-test"'), []);
  assert.ok(checkSkill(text, 'alonabot-review').length);
  assert.ok(checkSkill(text, 'alonabot-test', 'default_prompt: "使用 $wrong"').length);
});
