# echo 示例

发送普通文字或图片，插件会原样回送文字和入站图片。识别到命令、文本匹配或 Raw 后，事件回声通过 `sdk.NotCommand()` 读取 Core 的消费状态，避免重复回复。

在仓库根目录检查并构建：

```fish
go run ./cmd/alonabot check --dir examples/plugins/echo
go run ./cmd/alonabot build --dir examples/plugins/echo --out /tmp/alonabot-echo
```

按[开发指南](../../../docs/development.md)配置临时 `data_dir` 中的 Telegram 实例，再运行：

```fish
go run ./cmd/alonabot run adapters/telegram examples/plugins/echo examples/plugins/media --data-dir /tmp/alonabot-examples
```

| 输入 | 行为与对应写法 |
| --- | --- |
| `/echo hello world` | `p.Command(...).Args(...).Optional(...).RawTail().Handle(echoText)` 显式绑定包级函数，使用 `c.Reply`；省略文字时回复 `echo` |
| `/note add first  note`、`/n add second note` | 子插件 `notebook` 中的 `Group("note")`；位置参数 `RawTail` 保留内部空白，方法值绑定显式创建的 `notebook` 对象 |
| `/note list`、`/n list` | 两条命令分别声明完整的 `/n` 子命令别名；输出同一个对象中的笔记 |
| `/calc 4`、`/calc 4 3` | 结构体 `Calculation` 的 `arg`/`default` 标签，分别回复 `5`、`7`；返回 `(message.Content, error)` 自动回复 |
| `roll 2d6` | Regex 通过 `c.Match()` 读取两个捕获，回复 `Rolling 2d6.`，展示捕获参数 |
| `ECHO EXAMPLE KEYWORD` | 专属关键词加 `IgnoreCase`，返回一次自动回复 |
| `echo example raw` | Raw 的 Go Rule 仅接受这个专属输入，其他普通文本和图片继续原有回声 |
| `/ask` | 可信用户命令通过 `.Continue()` 进入延续调用；回复提问并 `c.Wait(time.Minute)`，回送同用户、同会话的下一条文字 |
| `/note publish`、`/note published` | 发布带类型的 `NotesSnapshot`，订阅回调只保存业务载荷；后一条命令读取已收到的快照 |

`New()` 每次创建独立插件和笔记对象。笔记是进程内、所有使用者共享的演示列表，重启后清空。子插件 ID 为 `notebook`，命令路径为 `/note ...`；route_id 仍为 `cmd.note.add` 等原规则生成的标识。

`notesService = service.JSON[NotesQuery, NotesSnapshot]("echo/notes")` 与 `service.Serve(p, notesService)` 提供公开只读 Service。调用方在有效 Handler 内用相同类型声明的 `notesService.Call(c, NotesQuery{})` 读取快照；跨插件工程应共享声明与具名类型的源码。示例本地联验在临时工程中复用 `commands.go`，构建真实调用方进程。服务公开意味着其他已激活插件可以读取这份演示列表。

`notesPublished = bus.Define[NotesSnapshot]("echo/notes_published")` 同时供 Publish 与 Subscribe 使用。发布成功只确认 Core 接纳；订阅处理异步进行，`/note published` 显示当时收到的快照。订阅 Context 没有用户 Actor 或默认目标，Handler 内保存列表即可。

新增声明和文案后不需要额外步骤：`alonabot run`、`build` 与 `go test` 每次都会重新扫描资源，工程中没有生成文件。
