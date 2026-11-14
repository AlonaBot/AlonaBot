# 用 devbot 完成一次收发

在仓库根目录操作。插件与回环 Adapter 都是可导入的组件工程；CLI 为它们生成独立进程入口，`main.go` 将同一组 `New()` 装配到当前进程。以下命令使用 fish。

先构建 CLI，把初始设置复制到临时数据目录，并指定回复观察文件：

```fish
go build -o bin/alonabot ./cmd/alonabot
set -gx DEVBOT_OUTPUT (mktemp)
set -gx DEVBOT_DATA (mktemp -d)
cp examples/devbot/core_config.example.json $DEVBOT_DATA/core_config.json
printf 'DEVBOT_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_DATA $DEVBOT_OUTPUT
bin/alonabot run examples/devbot/plugin examples/devbot/adapter --data-dir $DEVBOT_DATA --reload
```

命令占用前台。另开终端，将 `DEVBOT_DATA`、`DEVBOT_OUTPUT` 设为上一步的实际路径。首次在终端启动时，Core 向 stderr 显示一次管理 token；将其填入 `set token '…'`。非终端启动会把 token 写到数据目录，可用 `show` 读取：

```fish
set token (bin/alonabot admin-token show --data-dir $DEVBOT_DATA)
```

取得 token 后注入命令：

```fish
curl --fail-with-body -X POST \
  -H "Authorization: Bearer $token" \
  'http://127.0.0.1:7650/plugins/alonabot_adapter_devbot/inject?instance=one&text=%2Fhello'
cat $DEVBOT_OUTPUT
```

注入返回 HTTP 204；等发送完成后，文件内容为 `Hello from devbot!`。这是异步消息分发，204 表示 Core 已接纳事件，稍后再读文件即可观察回复。普通文本例如 `text=hello` 原样回声。路由要求 `instance` 与非空 `text` 查询参数，使用 URL 编码。样例固定 Actor 为 `devloop:user`，私聊会话为 `devloop:chat`，实例为 `one`，消息经 `adapter.Core.Deliver` 进入标准命令／事件与发送流水线。

`/inject` 是私有路由，复用 Core 管理 Bearer token；Cookie 认证的 POST 还需要受控 Origin 与 `X-Alonabot-CSRF`。Host 必须匹配 `base_url`；样例的监听与 Base URL 均为 `127.0.0.1:7650`。停止已有 Core 后再复用此端口。

注入接口的状态码：

| 状态码 | 含义 |
| --- | --- |
| 204 | Core 已接纳事件，回复异步发送 |
| 400 | 缺少 `instance` 或非空 `text` |
| 401 | 缺少有效认证 |
| 503 | 实例不可用，或 Core 分发队列已满、事件被丢弃；响应正文给出英文原因 |
| 500 | `Deliver` 返回错误，响应正文保留错误说明 |

## 工程与配置

```text
examples/devbot/
├── main.go                       # 同进程入口：信号 ctx → Start → Wait
├── core_config.example.json      # 首次复制为 data_dir/core_config.json
├── plugin/
│   ├── plugin.go                 # New、hello 命令与普通消息回声
│   ├── plugin.yaml               # alonabot_plugin_devbot
│   ├── config_schema.json        # greetings；启动时读取首项 text
│   ├── startup.sql               # startup 表
│   ├── i18n/{en-US,zh-CN}.json
│   ├── .gitignore                # data/
│   └── alonabot_gen.go
└── adapter/
    ├── adapter.go                # New、私有注入与文件观察
    ├── plugin.yaml               # alonabot_adapter_devbot；adapter_id=devloop
    ├── adapter_config_schema.json
    └── alonabot_gen.go
```

样例中实例所在的设置如下，`${DEVBOT_OUTPUT}` 在读取设置时展开为环境变量的完整值。`output` 可填写绝对文件路径；空串表示只返回发送回执。文件保存最近一次发送的文本，权限为 0600。

```json
{
  "command_prefixes": ["/"],
  "adapters": {
    "devloop": {
      "package": "alonabot_adapter_devbot",
      "instances": {"one": {"output": "${DEVBOT_OUTPUT}"}}
    }
  }
}
```

数据目录缺省为当前工作目录的 `data/`。组件配置在 `<data_dir>/plugins/alonabot_plugin_devbot/config.json`，受管数据库在同目录的 `dev.sqlite`。配置缺失时使用 schema 缺省的 `greetings: [{"text":"Hello"}]`。每次插件启动向 `startup` 表追加首项问候文本，重启保留记录；原始配置文件保持原文。开发代次固定为 `dev:<组件 ID>`；管理入口对 dev 来源的 update、rollback、uninstall 返回 `FAILED_PRECONDITION`。

## 改源码、翻译和配置

保持 `run --reload` 运行，修改 `plugin/i18n/en-US.json` 的 `messages.hello`，再次注入 `/hello`，观察新的译文。设置 `default_locale` 为 `zh-CN` 后，新建立的插件 Session 使用中文。修改 `plugin.go` 同样会重新构建并只重启插件，Core、回环实例及已安装组件保持运行。

在 `plugin/` 临时放一个有语法错误的 `.go` 文件，日志会保留编译器诊断，旧插件仍能回复；修正或删除该文件后自动恢复构建。

只改配置时可在另一个终端操作：

```fish
mkdir -p $DEVBOT_DATA/plugins/alonabot_plugin_devbot
printf '%s\n' '{"greetings":[{"text":"Updated"}]}' > $DEVBOT_DATA/plugins/alonabot_plugin_devbot/config.json
```

这会重启插件并追加 startup 记录，`<data_dir>/dev/bin/alonabot_plugin_devbot` 的字节和 mtime 保持。只改一个 Adapter 实例的设置时，只重启该实例。schema、i18n、源码、嵌入资源、manifest、go.mod/go.sum 的变化会触发重新构建；编译失败保留当前进程。轮询缺省为 500ms，可用 `--args DEV_RELOAD_POLL=1s` 调整。

## 同进程调试

先用 Ctrl-C 停止独立进程模式，再准备另一个数据目录，复用相同 Adapter 身份与实例 schema：

```fish
set -gx DEVBOT_LIBRARY_DATA (mktemp -d)
set -gx DEVBOT_OUTPUT (mktemp)
cp examples/devbot/core_config.example.json $DEVBOT_LIBRARY_DATA/core_config.json
printf 'DEVBOT_LIBRARY_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_LIBRARY_DATA $DEVBOT_OUTPUT
go generate ./examples/devbot/...
go run ./examples/devbot -data-dir $DEVBOT_LIBRARY_DATA
```

使用上面的 token 与注入步骤，将数据目录替换为 `DEVBOT_LIBRARY_DATA`。调试器可直接在 `plugin.New()` 注册的 Handler 与 Adapter 的 `Send` 中打断点。源码与资源变化后，重新生成并重新启动 Go 程序。

两个模式收到 SIGINT/SIGTERM 后按同一停止预算排空、执行停止回调，缺省总计 `STOP_TOTAL=10s`。独立进程由 supervisor 隔离并执行有限崩溃恢复；库模式下未捕获的后台 panic 使整个进程非零退出，停止预算到期仍有组件运行时 `Wait` 返回错误。SDK Handler 边界捕获的 panic 记录 stack 并返回 `INTERNAL`，后续调用继续。

## 创建与检查工程

`alonabot init [目录] --id ID` 在空目录创建可导入插件包、作者 manifest、schema、i18n 与忽略 `data/`、`bin/` 的 `.gitignore`，导出 `New()` 并用 `sdk.UseGenerated` 绑定资源。模板与本示例的 plugin 具有相同工程形状；本示例另含启动计数 SQL、回声与回环 Adapter。库模式宿主沿用本示例根目录的 main.go。

```fish
go run ./cmd/alonabot build --dir examples/devbot/plugin --out bin/devbot-plugin
go run ./cmd/alonabot build --dir examples/devbot/adapter --out bin/devbot-adapter
go run ./cmd/alonabot check --dir examples/devbot/plugin
go run ./cmd/alonabot check --dir examples/devbot/adapter
go run ./cmd/alonabot test --dir examples/devbot/plugin
```

创建外部工程时，目标目录已有 Go module 就复用；否则须提供 `--module`，SDK 版本来源及 `init → build → check → test → pack` 完整步骤见 [开发文档](../../docs/development.md)。管理界面的开发配置写回与重启由 T-09.14 承接。
