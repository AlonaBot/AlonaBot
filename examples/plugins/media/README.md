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
