# 同时运行 Telegram、echo 与媒体示例

在两个测试 Bot 的会话里发送普通文字或图片，echo 会保留顺序回送。发送 `/media`，media 会上传内置 PNG，发送两张图片组成的相册和一条带按钮的文字；点击按钮后编辑那条文字。

先在仓库根创建数据目录，将下面内容保存为 `<data_dir>/core_config.json`：

```json
{
  "adapters": {
    "telegram": {
      "package": "alonabot_adapter_telegram",
      "$version": 1,
      "default_instance": "tg_a",
      "instances": {
        "tg_a": {
          "token": "${TELEGRAM_BOT_TOKEN_A}",
          "api_base_url": "https://api.telegram.org",
          "proxy_mode": "none"
        },
        "tg_b": {
          "token": "${TELEGRAM_BOT_TOKEN_B}",
          "api_base_url": "https://api.telegram.org",
          "proxy_mode": "none"
        }
      }
    }
  }
}
```

在当前 shell 导出 `TELEGRAM_BOT_TOKEN_A` 与 `TELEGRAM_BOT_TOKEN_B`，值取自两个不同的 Bot；Core 保留文件中的占位符，将运行副本交给 Telegram Adapter。然后在仓库根运行：

```fish
go run ./cmd/alonabot run adapters/telegram examples/plugins/echo examples/plugins/media --data-dir <data_dir>
```

追加 `--stdio` 可切换进程传输，追加 `--reload` 可监视源码与配置。回复语言跟随用户语言，缺失时使用 Core 默认语言。

echo 直接回送入站图片的 Incoming 来源；同实例发送会复用平台 file_id。media 的相册后接文本，按钮附在文本上。示例只在发送成功且尾随文本带可用消息 ID 时保存 `result.Refs()` 的末条句柄；部分失败保留平台已发送的内容并返回错误。

编辑按钮为一次性演示，待点击句柄保存在插件的受管 SQLite 中，成功编辑后删除。

## 结构体参数与子插件

发送 `/caption a  custom caption`，得到带这段原文说明的示例图片。`Caption.Text` 使用 `arg:"text,raw"` 保留内部空白，Handler 返回 `(message.Content, error)` 自动发送。它属于 `p.Sub("captions")`，默认要求 `sdk.User`；命令路径为 `/caption`，route_id 为 `cmd.caption`。

在仓库根目录检查和构建：

```fish
go run ./cmd/alonabot check --dir examples/plugins/media
go run ./cmd/alonabot build --dir examples/plugins/media --out /tmp/alonabot-media
```

`New()` 只做注册，所有注册都显式可见。原 `/media` 使用 `c.Send` 获取完整 `SendResult`，先保存可编辑文字的消息句柄，再接收按钮点击；`/caption` 展示结构体参数和子插件，不需要持有发送回执。英文和中文按钮、提示取自 i18n 资源。

真实 CLI/Core/官方 Adapter/示例进程与本地 Bot API 的主路径由 `TestSDKExamples` 和原有 `TestTelegramDevelopment` 验证；真实 Telegram 验收沿用所属切片的记录。
