// 本文件校验维护文档的本地链接和仓库技能入口；不联网、不运行产品代码，也不判断注释语义。
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

// checkLinks 返回正文中不可移植或不存在的本地文件引用；代码块不作为 Markdown 链接解析。
export function checkLinks(text, file, root) {
  // 1. 仅提取 Markdown 正文中的内联链接，外部 URL 和页内锚点不访问网络。
  const body = text.replace(/^(`{3,}|~{3,})[^\n]*\n[\s\S]*?^\1\s*$/gm, '');
  const errors = [];
  for (const match of body.matchAll(/\[[^\]]*\]\(([^\s)]+)\)/g)) {
    const target = match[1].split('#')[0];
    if (!target || /^[a-z][a-z\d+.-]*:/i.test(target)) continue;

    // 2. 从引用文件解析相对路径，拒绝依赖维护者机器目录的契约入口。
    const resolved = path.resolve(path.dirname(file), decodeURIComponent(target));
    const relative = path.relative(root, resolved);
    if (path.isAbsolute(target) || relative === '..' || relative.startsWith(`..${path.sep}`)) {
      errors.push(`${path.relative(root, file)}: 本地引用必须位于仓库内：${target}`);
    } else if (!existsSync(resolved)) {
      errors.push(`${path.relative(root, file)}: 本地引用不存在：${target}`);
    }
  }
  return errors;
}

// checkSkill 检查项目技能的名称、描述及可选 UI 提示；完整 YAML 语法由技能维护时单独验证。
export function checkSkill(text, name, ui) {
  // 1. 核对技能入口的基础元数据，目录名就是可发现的调用名。
  const errors = [];
  const metadata = text.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/)?.[1];
  if (!metadata || !new RegExp(`^name: ${name}$`, 'm').test(metadata) || !/^description: .+$/m.test(metadata)) {
    errors.push(`${name}: 缺少与目录一致的 name 或非空单行 description`);
  }
  // 2. 提供 UI 提示时，确保它仍然引用当前技能。
  if (ui !== undefined && !ui.includes(`$${name}`)) errors.push(`${name}: UI 提示未引用本技能`);
  return errors;
}

// checkRepository 从项目维护入口检查直接引用；不扫描历史规格中的全部链接或判定锚点。
export function checkRepository(root) {
  // 1. 收集固定维护入口和可发现的项目技能，保持检查范围与 AGENTS 路由一致。
  const files = ['AGENTS.md', 'CLAUDE.md', 'CONTRIBUTING.md', 'ARCHITECTURE.md', 'docs/index.md', 'docs/development.md', 'docs/coding-standards.md'];
  const errors = [];
  const skills = path.join(root, '.agents/skills');
  for (const entry of readdirSync(skills, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const relative = `.agents/skills/${entry.name}/SKILL.md`;
    const skill = path.join(root, relative);
    files.push(relative);
    if (!existsSync(skill)) continue; // 缺少入口的错误由下面统一返回。
    const ui = path.join(path.dirname(skill), 'agents/openai.yaml');
    errors.push(...checkSkill(readFileSync(skill, 'utf8'), entry.name, existsSync(ui) ? readFileSync(ui, 'utf8') : undefined));
  }

  // 2. 检查每个入口及其直接链接，把所有错误一次返回给维护者。
  for (const relative of files) {
    const file = path.join(root, relative);
    if (!existsSync(file)) errors.push(`缺少维护入口：${relative}`);
    else errors.push(...checkLinks(readFileSync(file, 'utf8'), file, root));
  }
  return errors;
}

// 命令行入口仅报告检查结果；被测试导入时不执行仓库扫描。
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const errors = checkRepository(process.cwd());
  if (errors.length) {
    console.error(errors.join('\n'));
    process.exitCode = 1;
  } else {
    console.log('仓库维护入口、本地文件引用和技能元数据检查通过。');
  }
}
