// 检查 Core 可执行文件与 SDK 根包在五个目标上链接的 Go 模块都在允许清单中（CONTRACTS §10、D-52）。
// 允许清单只在本文件维护，是 docs/development.md 已批准依赖的子集；新增模块须先按该文件的依赖流程取得批准。
import { execFileSync } from 'node:child_process';

// modernc.org/sqlite 及其传递依赖；go-isatty 与 go-strftime 只在部分平台出现。
const sqlite = [
  'modernc.org/sqlite',
  'modernc.org/libc',
  'modernc.org/mathutil',
  'modernc.org/memory',
  'github.com/dustin/go-humanize',
  'github.com/google/uuid',
  'github.com/remyoudompheng/bigfft',
  'github.com/mattn/go-isatty',
  'github.com/ncruces/go-strftime',
];

const allowed = {
  './cmd/alonabot': new Set([
    'google.golang.org/protobuf',
    'golang.org/x/sys',
    'go.yaml.in/yaml/v3',
    'github.com/google/jsonschema-go',
    'github.com/spf13/cobra',
    'github.com/spf13/pflag',
    'github.com/inconshreveable/mousetrap', // Cobra 只在 Windows 链接。
    ...sqlite,
  ]),
  './sdk': new Set(['google.golang.org/protobuf', 'golang.org/x/sys', 'github.com/google/jsonschema-go']),
};

const targets = ['linux/amd64', 'linux/arm64', 'windows/amd64', 'darwin/amd64', 'darwin/arm64'];

let failed = false;
for (const [pkg, ok] of Object.entries(allowed)) {
  for (const target of targets) {
    const [GOOS, GOARCH] = target.split('/');
    const out = execFileSync(
      'go',
      ['list', '-deps', '-f', '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}', pkg],
      { env: { ...process.env, GOOS, GOARCH, CGO_ENABLED: '0' }, encoding: 'utf8' },
    );
    const extra = [...new Set(out.split('\n').filter(Boolean))].filter((m) => !ok.has(m)).sort();
    if (extra.length > 0) {
      failed = true;
      console.error(`${pkg} (${target}) links modules outside the allowlist: ${extra.join(', ')}`);
    }
  }
}

if (failed) {
  process.exit(1);
}
console.log('Go dependency allowlist check passed');
