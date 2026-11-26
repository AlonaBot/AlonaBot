# Send and receive a message with devbot

[简体中文](README.zh-CN.md)

Build the CLI from the repository root first: `go build -o bin/alonabot ./cmd/alonabot`. The commands below use fish.

1. Create a plugin project: `bin/alonabot init ./alonabot-plugin-hello`.
2. Run it: `bin/alonabot run ./alonabot-plugin-hello`.
3. Enter `/hello` in the same terminal to see `Hello AlonaBot`. Press Ctrl-C to stop.

With an installed CLI, run `alonabot run` from the project directory. Console chat starts automatically when stdin is a terminal and no Adapter instances are configured; `--console` forces it on (including pipes), and `--console=false` disables it. The console user `console:local` has Core Admin permission for this run; the instance and administrator remain in memory. Enter `#1` to click the first callback button on the most recent message with buttons. Edits appear as new `(edited)` segments; media appear as bracketed placeholders. After a pipe ends, Core keeps running until it receives a stop signal.

Development directories enable reload by default. After changing source, translations or configuration, send another message to observe the new reply. Use `--reload=false` to disable it. For external projects with a release CLI, `init` writes the SDK version and runs `go mod tidy`; on failure it keeps the project and prints recovery commands. See the [development guide](https://alonabot.github.io/en/docs/reference/cli#alonabot-init) for SDK setup with a development CLI.

## Without a terminal: loopback Adapter and curl

The following uses the existing devbot plugin and loopback Adapter. Both are importable component projects. The CLI generates separate process entry points, while `main.go` assembles the same `New()` functions in the current process.

If you have a data directory from before the rename, stop Core and devbot first and rebuild as described in the [development guide](../../docs/development.md#development-runs-and-hot-reload). The component IDs are `alonabot-plugin-devbot` and `alonabot-adapter-devbot`; the platform ID remains `devloop`.

Build the CLI, copy the initial settings into a temporary data directory, and choose a file in which to observe replies:

```fish
go build -o bin/alonabot ./cmd/alonabot
set -gx DEVBOT_OUTPUT (mktemp)
set -gx DEVBOT_DATA (mktemp -d)
cp examples/devbot/core_config.example.json $DEVBOT_DATA/core_config.json
printf 'DEVBOT_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_DATA $DEVBOT_OUTPUT
bin/alonabot run examples/devbot/plugin examples/devbot/adapter --data-dir $DEVBOT_DATA --reload
```

The command runs in the foreground. Open another terminal and set `DEVBOT_DATA` and `DEVBOT_OUTPUT` to the actual paths from the previous step. On first startup, Core persists the management token in `$DEVBOT_DATA/admin_token` with mode 0600 and prints only its location. Management CLI commands read it automatically using `--data-dir`; for manual HTTP requests, explicitly read it with `admin-token show`:

```fish
set token (bin/alonabot admin-token show --data-dir $DEVBOT_DATA)
```

After obtaining the token, inject a command:

```fish
curl --fail-with-body -X POST \
  -H "Authorization: Bearer $token" \
  'http://127.0.0.1:7650/plugins/alonabot-adapter-devbot/inject?instance=one&text=%2Fhello'
cat $DEVBOT_OUTPUT
```

Injection returns HTTP 204. Once the asynchronous send finishes, the file contains `Hello from devbot!`. A 204 means Core accepted the event; read the file again shortly afterward to observe the reply. Plain text such as `text=hello` is echoed unchanged. The route requires `instance` and a nonempty `text` query parameter; URL-encode them. The example fixes Actor to `devloop:user`, the private chat to `devloop:chat`, and the instance to `one`. Messages enter the normal command/event and send pipeline through `adapter.Core.Deliver`.

`/inject` is a private route using the Core management Bearer token. Cookie-authenticated POST requests also require a controlled Origin and `X-Alonabot-CSRF`. Host must match `base_url`; both the example listener and Base URL are `127.0.0.1:7650`. Stop any existing Core before reusing this port.

Injection status codes:

| Status | Meaning |
| --- | --- |
| 204 | Core accepted the event; the reply is sent asynchronously |
| 400 | Missing `instance` or nonempty `text` |
| 401 | Missing valid authentication |
| 503 | Instance unavailable, or Core's dispatch queue was full and the event was dropped; the body gives the reason in English |
| 500 | `Deliver` returned an error; the body retains its description |

## Project and configuration

```text
examples/devbot/
├── main.go                       # 同进程入口：信号 ctx → Start → Wait
├── core_config.example.json      # 首次复制为 data_dir/core_config.json
├── plugin/
│   ├── plugin.go                 # New、hello 命令与普通消息回声
│   ├── plugin.yaml               # alonabot-plugin-devbot
│   ├── plugin_config_schema.json        # greetings；启动时读取首项 text
│   ├── startup.sql               # startup 表
│   ├── i18n/{en-US,zh-CN}.json
│   └── .gitignore                # data/
└── adapter/
    ├── adapter.go                # New、私有注入与文件观察
    ├── plugin.yaml               # alonabot-adapter-devbot；adapter_id=devloop
    └── adapter_config_schema.json
```

The example's instance settings are shown below. `${DEVBOT_OUTPUT}` expands to the complete environment variable value when settings are read. `output` accepts an absolute file path; an empty string returns only the send receipt. The file stores the most recently sent text with mode 0600.

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

The data directory defaults to `data/` in the current working directory. Component configuration is at `<data_dir>/plugins/alonabot-plugin-devbot/config.json`, with the managed database `dev.sqlite` in the same directory. Missing configuration uses the schema default `greetings: [{"text":"Hello"}]`. Each plugin startup appends the first greeting to the `startup` table, retaining rows across restarts; the original configuration file is preserved. The development generation is fixed to `dev:<component ID>`. Management update, rollback and uninstall return `FAILED_PRECONDITION` for dev sources.

## Change source, translations and configuration

Keep `run` running, edit `messages.hello` in `plugin/i18n/en-US.json`, then inject `/hello` again to see the new translation. After setting `default_locale` to `zh-CN`, new plugin Sessions use Chinese. Changing `plugin.go` also rebuilds and restarts only that plugin; Core, the loopback instance and installed components keep running.

Temporarily add a `.go` file with a syntax error under `plugin/`. Logs retain compiler diagnostics and the old plugin can still reply. Fix or remove the file to resume builds automatically.

For a configuration-only change, use another terminal:

```fish
mkdir -p $DEVBOT_DATA/plugins/alonabot-plugin-devbot
printf '%s\n' '{"greetings":[{"text":"Updated"}]}' > $DEVBOT_DATA/plugins/alonabot-plugin-devbot/config.json
```

This restarts the plugin and appends a startup row while preserving the bytes and mtime of `<data_dir>/dev/bin/alonabot-plugin-devbot`. Changing only one Adapter instance's settings restarts only that instance. Changes to schemas, i18n, source, embedded resources, the manifest or go.mod/go.sum trigger a rebuild; a compilation failure keeps the current process. Polling defaults to 500ms; adjust it with `--args DEV_RELOAD_POLL=1s`.

## Debug in the same process

Stop the separate-process run with Ctrl-C, then prepare another data directory using the same Adapter identity and instance schema:

```fish
set -gx DEVBOT_LIBRARY_DATA (mktemp -d)
set -gx DEVBOT_OUTPUT (mktemp)
cp examples/devbot/core_config.example.json $DEVBOT_LIBRARY_DATA/core_config.json
printf 'DEVBOT_LIBRARY_DATA=%s\nDEVBOT_OUTPUT=%s\n' $DEVBOT_LIBRARY_DATA $DEVBOT_OUTPUT
set -gx DEVBOT_BIN (mktemp -d)
go run ./cmd/alonabot build examples/devbot --out $DEVBOT_BIN
$DEVBOT_BIN/devbot -data-dir $DEVBOT_LIBRARY_DATA
```

Follow the token and injection steps above, replacing the data directory with `DEVBOT_LIBRARY_DATA`. Library-mode resources are embedded by `alonabot build`; a plain `go run` or `go build` executable reports missing resources before starting components. Set debugger breakpoints in Handlers registered by `plugin.New()` and in the Adapter's `Send`. After changing source or resources, run `alonabot build` again and restart the program.

Both modes drain work and run stop callbacks on SIGINT/SIGTERM with the same stop budget, default total `STOP_TOTAL=10s`. The supervisor isolates separate processes and performs bounded crash recovery. In library mode, an unhandled background panic exits the entire process with a nonzero status; `Wait` returns an error if components remain after the stop budget expires. Panics caught at SDK Handler boundaries log a stack and return `INTERNAL`; later calls continue.

## Create and check a project

`alonabot init [directory] --id ID` creates an importable plugin package, author manifest, schema and `.gitignore` excluding `data/` and `bin/`. Its exported `New()` only registers behavior. No generated files are kept in the project; resources are collected automatically during builds and `go test`. The template and this example's plugin share the same project shape; the example additionally includes startup-count SQL, echo behavior, i18n and the loopback Adapter. Library-mode hosts follow the root main.go in this example.

```fish
go run ./cmd/alonabot build examples/devbot/plugin --out bin/devbot-plugin
go run ./cmd/alonabot build examples/devbot/adapter --out bin/devbot-adapter
go run ./cmd/alonabot check examples/devbot/plugin
go run ./cmd/alonabot check examples/devbot/adapter
go run ./cmd/alonabot test examples/devbot/plugin
```

An external project reuses an existing Go module in the target directory; otherwise the module defaults to the directory name, or can be set with `--module`. See the [development guide](https://alonabot.github.io/en/docs/reference/cli#alonabot-init) for SDK version selection and the full `init → build → check → test → pack` flow. T-09.14 covers development configuration writeback and restarts through the management UI.
