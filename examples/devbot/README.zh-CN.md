# 用 devbot 完成一次收发

[English](README.md)

在仓库根目录先构建 CLI：`go build -o bin/alonabot ./cmd/alonabot`。以下命令使用 fish。

1. 创建插件工程：`bin/alonabot init ./alonabot-plugin-hello`。
2. 运行：`bin/alonabot run ./alonabot-plugin-hello`。
3. 在同一个终端输入 `/hello`，看到 `Hello AlonaBot`。Ctrl-C 停止。

已安装 CLI 时，在工程目录直接执行 `alonabot run` 即可。终端输入且设置中没有任何 Adapter 实例时，自动启用终端对话；`--console` 强制启用（支持管道），`--console=false` 关闭。终端用户 `console:local` 在当前运行中具有 Core Admin 权限，实例与管理员保持在内存中。输入 `#1` 点击最近一条带按钮消息的第一个回调按钮，编辑回复显示为 `(edited)` 新段，媒体显示方括号占位。管道结束后 Core 继续运行到收到停止信号。

开发目录缺省开启重载；修改源码、翻译或配置后，下一次输入即可观察新回复。`--reload=false` 关闭。外部工程使用发布版 CLI 时，`init` 自动写入 SDK 版本并执行 `go mod tidy`；失败保留工程并显示恢复命令。开发版 CLI 的 SDK 补齐提示见[开发指南](../../docs/development.zh-CN.md#开发运行与热部署)。

## 不经终端：回环 Adapter 与 curl

下面使用已有 devbot 插件和回环 Adapter。两者都是可导入的组件工程；CLI 为它们生成独立进程入口，`main.go` 将同一组 `New()` 装配到当前进程。

已有改名前的数据目录时，先停止 Core 和 devbot，再按[开发指南](../../docs/development.zh-CN.md#开发运行与热部署)重建；本例的组件 ID 为 `alonabot-plugin-devbot` 和 `alonabot-adapter-devbot`，平台 ID 仍为 `devloop`。

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
  'http://127.0.0.1:7650/plugins/alonabot-adapter-devbot/inject?instance=one&text=%2Fhello'
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
│   ├── plugin.yaml               # alonabot-plugin-devbot
│   ├── config_schema.json        # greetings；启动时读取首项 text
│   ├── startup.sql               # startup 表
│   ├── i18n/{en-US,zh-CN}.json
│   └── .gitignore                # data/
└── adapter/
    ├── adapter.go                # New、私有注入与文件观察
    ├── plugin.yaml               # alonabot-adapter-devbot；adapter_id=devloop
    └── adapter_config_schema.json
```

样例中实例所在的设置如下，`${DEVBOT_OUTPUT}` 在读取设置时展开为环境变量的完整值。`output` 可填写绝对文件路径；空串表示只返回发送回执。文件保存最近一次发送的文本，权限为 0600。

```json
{
  "command_prefixes": ["/"],
  "adapters": {
    "devloop": {
      "package": "alonabot-adapter-devbot",
      "instances": {"one": {"output": "${DEVBOT_OUTPUT}"}}
    }
  }
}
```

数据目录缺省为当前工作目录的 `data/`。组件配置在 `<data_dir>/plugins/alonabot-plugin-devbot/config.json`，受管数据库在同目录的 `dev.sqlite`。配置缺失时使用 schema 缺省的 `greetings: [{"text":"Hello"}]`。每次插件启动向 `startup` 表追加首项问候文本，重启保留记录；原始配置文件保持原文。开发代次固定为 `dev:<组件 ID>`；管理入口对 dev 来源的 update、rollback、uninstall 返回 `FAILED_PRECONDITION`。

## 改源码、翻译和配置

保持 `run` 运行，修改 `plugin/i18n/en-US.json` 的 `messages.hello`，再次注入 `/hello`，观察新的译文。设置 `default_locale` 为 `zh-CN` 后，新建立的插件 Session 使用中文。修改 `plugin.go` 同样会重新构建并只重启插件，Core、回环实例及已安装组件保持运行。

在 `plugin/` 临时放一个有语法错误的 `.go` 文件，日志会保留编译器诊断，旧插件仍能回复；修正或删除该文件后自动恢复构建。

只改配置时可在另一个终端操作：

```fish
mkdir -p $DEVBOT_DATA/plugins/alonabot-plugin-devbot
printf '%s\n' '{"greetings":[{"text":"Updated"}]}' > $DEVBOT_DATA/plugins/alonabot-plugin-devbot/config.json
```

这会重启插件并追加 startup 记录，`<data_dir>/dev/bin/alonabot-plugin-devbot` 的字节和 mtime 保持。只改一个 Adapter 实例的设置时，只重启该实例。schema、i18n、源码、嵌入资源、manifest、go.mod/go.sum 的变化会触发重新构建；编译失败保留当前进程。轮询缺省为 500ms，可用 `--args DEV_RELOAD_POLL=1s` 调整。

## 同进程调试

先用 Ctrl-C 停止独立进程模式，再准备另一个数据目录，复用相同 Adapter 身份与实例 schema：

```fish
set -gx DEVBOT_LIBRARY_DATA (mktemp -d)
set -gx DEVBOT_OUTPUT (mktemp)
cp examples/devbot/core_config.example.json $DEVBOT_LIBRARY_DATA/core_config.json
printf 'DEVBOT_LIBRARY_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_LIBRARY_DATA $DEVBOT_OUTPUT
set -gx DEVBOT_BIN (mktemp -d)
go run ./cmd/alonabot build examples/devbot --out $DEVBOT_BIN
$DEVBOT_BIN/devbot -data-dir $DEVBOT_LIBRARY_DATA
```

使用上面的 token 与注入步骤，将数据目录替换为 `DEVBOT_LIBRARY_DATA`。库模式的资源由 `alonabot build` 在构建时编入，直接 `go run` 或 `go build` 的程序在启动组件前会报缺少资源。调试器可在 `plugin.New()` 注册的 Handler 与 Adapter 的 `Send` 中打断点。源码与资源变化后，重新运行 `alonabot build` 再启动程序。

两个模式收到 SIGINT/SIGTERM 后按同一停止预算排空、执行停止回调，缺省总计 `STOP_TOTAL=10s`。独立进程由 supervisor 隔离并执行有限崩溃恢复；库模式下未捕获的后台 panic 使整个进程非零退出，停止预算到期仍有组件运行时 `Wait` 返回错误。SDK Handler 边界捕获的 panic 记录 stack 并返回 `INTERNAL`，后续调用继续。

## 创建与检查工程

`alonabot init [目录] --id ID` 在空目录创建可导入插件包、作者 manifest、schema 与忽略 `data/`、`bin/` 的 `.gitignore`，导出只做注册的 `New()`，工程中没有生成文件，资源在构建与 `go test` 时自动取得。需要翻译时自行添加 i18n 目录；本示例另含启动计数 SQL、回声与回环 Adapter。库模式宿主沿用本示例根目录的 main.go。

```fish
go run ./cmd/alonabot build examples/devbot/plugin --out bin/devbot-plugin
go run ./cmd/alonabot build examples/devbot/adapter --out bin/devbot-adapter
go run ./cmd/alonabot check examples/devbot/plugin
go run ./cmd/alonabot check examples/devbot/adapter
go run ./cmd/alonabot test examples/devbot/plugin
```

创建外部工程时，目标目录已有 Go module 就复用；否则 module 缺省取目录名，也可用 `--module` 指定，SDK 版本来源及 `init → build → check → test → pack` 完整步骤见 [开发文档](../../docs/development.zh-CN.md)。管理界面的开发配置写回与重启由 T-09.14 承接。
