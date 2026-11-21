"""使用发行 Compose 验收镜像的登录、默认组件、数据卷和退出。"""

import argparse
import json
from pathlib import Path
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--compose', required=True)
    parser.add_argument('--image', required=True)
    args = parser.parse_args()
    project = 'alonabot-release-' + uuid.uuid4().hex[:12]
    with tempfile.TemporaryDirectory(prefix=project) as temporary:
        compose = Path(temporary) / 'compose.json'
        config = json.loads(subprocess.check_output(['docker', 'compose', '-p', project, '-f', args.compose, 'config', '--format', 'json']))
        config['services']['alonabot']['image'] = args.image
        compose.write_text(json.dumps(config))
        command = ['docker', 'compose', '-p', project, '-f', str(compose)]

        def run(*arguments):
            return subprocess.check_output([*command, *arguments], text=True).strip()

        def call(path, token='', body=None):
            headers = {'Authorization': 'Bearer ' + token, 'Origin': 'http://127.0.0.1:7650'}
            if body is not None:
                headers['Content-Type'] = 'application/json'
            request = urllib.request.Request('http://127.0.0.1:7650' + path, data=body, headers=headers)
            with urllib.request.urlopen(request, timeout=5) as response:
                return response.read()

        def ready():
            # 启动验收最多等待 90 秒；超时保留容器日志并使发行失败。
            deadline = time.monotonic() + 90
            while True:
                try:
                    assert b'<html' in call('/')
                    return
                except (OSError, urllib.error.URLError):
                    if time.monotonic() >= deadline:
                        raise
                    time.sleep(0.5)

        def packages(token):
            actual = {item['ID']: item for item in json.loads(call('/api/v1/packages', token))}
            assert set(actual) == {'alonabot-webui-admin', 'alonabot-agent-official'}, actual
            for identity, enabled in [('alonabot-webui-admin', True), ('alonabot-agent-official', False)]:
                item = actual[identity]
                assert item['Enabled'] == enabled and item['Signed'], item
                assert item['State'] == ('Activated' if enabled else 'Installed'), item

        try:
            run('up', '-d', '--pull', 'never')
            ready()
            token = run('exec', '-T', 'alonabot', 'alonabot', 'admin-token', 'show', '--data-dir', '/data')
            assert token
            packages(token)
            call('/api/auth/login', token, json.dumps({'token': token}).encode())
            # 领取后凭证文件已删除；同一 token 跨进程重启有效，证明持久化状态仍在卷内。
            run('restart')
            ready()
            packages(token)
            run('stop', '-t', '30')
            container = run('ps', '-aq')
            state = json.loads(subprocess.check_output(['docker', 'inspect', container]))[0]['State']
            assert state['Status'] == 'exited' and state['ExitCode'] == 0, state
            logs = run('logs', '--no-color')
            assert token not in logs, 'admin token leaked into container logs'
            print('Compose login, components, persisted authentication and clean shutdown passed')
        except BaseException:
            subprocess.run([*command, 'logs', '--no-color'], check=False)
            raise
        finally:
            subprocess.run([*command, 'down', '-v'], check=True)


if __name__ == '__main__':
    main()
