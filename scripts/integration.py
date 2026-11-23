"""运行独立组件整合测试，并把需要 Core 私有入口的测试临时 overlay 回原包。"""

import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--run', default='.')
    parser.add_argument('--race', action='store_true')
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    environment = os.environ.copy()
    # 托管 runner 的端到端包已超过 Go 缺省的 10 分钟总期限，包含组件冷构建。
    flags = ['-p', '1', '-timeout', '30m'] + (['-race'] if args.race else [])
    subprocess.run(['go', 'test', *flags, '-run', args.run, './...'], cwd=root / 'tests', env=environment, check=True)
    agent = subprocess.check_output(
        ['go', 'list', '-m', '-f', '{{.Path}}@{{.Version}}', 'github.com/AlonaBot/alonabot-agent'],
        cwd=root / 'tests', env=environment, text=True,
    ).strip()
    with tempfile.TemporaryDirectory(prefix='alonabot-integration-') as directory:
        directory = Path(directory)
        modfile = directory / 'integration.mod'
        shutil.copyfile(root / 'go.mod', modfile)
        shutil.copyfile(root / 'go.sum', modfile.with_suffix('.sum'))
        subprocess.run(['go', 'mod', 'edit', '-modfile', str(modfile), '-require', agent], cwd=root, check=True)
        overlay = directory / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(root / 'internal/app/agent_service_test.go'): str(root / 'tests/app/agent_service_test.go.tmpl')}}))
        # 临时 module 图只用于当前构建，独立 fixture 的子命令沿用各自 go.mod。
        subprocess.run(['go', 'test', '-modfile', str(modfile), '-mod=mod', '-overlay', str(overlay), *flags, '-run', args.run, './internal/app'], cwd=root, env=environment, check=True)


if __name__ == '__main__':
    main()
