# Run Telegram, echo and the media example together

[简体中文](README.zh-CN.md)

Send plain text or images in chats with two test Bots; echo returns them in order. Send `/media` to upload the bundled PNG, send an album of two images, and send text with a button. Clicking the button edits that text.

Create a new data directory at the repository root and save the following as `<data_dir>/core_config.json`. When updating an example environment from before the rename, stop Core and development components and rebuild using the [development guide](../../../docs/development.md#development-runs-and-hot-reload). The configuration key `telegram` is the platform ID; `package` is the component ID:

```json
{
  "adapters": {
    "telegram": {
      "package": "alonabot-adapter-telegram",
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

Export `TELEGRAM_BOT_TOKEN_A` and `TELEGRAM_BOT_TOKEN_B` in the current shell using two different Bots. Core preserves the placeholders in the file and passes an expanded runtime copy to the Telegram Adapter. Then run from the repository root:

```fish
git clone https://github.com/AlonaBot/alonabot-adapter-telegram.git ../alonabot-adapter-telegram
go run ./cmd/alonabot run ../alonabot-adapter-telegram examples/plugins/echo examples/plugins/media --data-dir <data_dir>
```

Append `--stdio` to switch process transport, or `--reload` to watch source and configuration. Replies follow the user's locale, falling back to the Core default.

echo returns the Incoming source of inbound images directly; sending within the same instance reuses the platform file_id. media sends text after the album and attaches the button to that text. The example saves the last handle from `result.Refs()` only after a successful send whose trailing text has a usable message ID. Partial failures preserve content already sent by the platform and return an error.

The edit button is a one-shot demonstration. Its pending handle is stored in the plugin's managed SQLite and deleted after a successful edit.

## Struct arguments and subplugins

Send `/caption a  custom caption` to receive the example image with that exact caption. `Caption.Text` uses `arg:"text,raw"` to preserve internal whitespace. Returning `(message.Content, error)` from the Handler sends it automatically. It belongs to `p.Sub("captions")`, requires `sdk.User` by default, and has command path `/caption` and route_id `cmd.caption`.

Check and build from the repository root:

```fish
go run ./cmd/alonabot check examples/plugins/media
go run ./cmd/alonabot build examples/plugins/media --out /tmp/alonabot-media
```

`New()` only registers behavior, with every registration explicit. The original `/media` uses `c.Send` for the full `SendResult`, saves the editable text handle, then handles button clicks. `/caption` demonstrates struct arguments and subplugins without retaining send receipts. English and Chinese button labels and prompts come from i18n resources.

`TestSDKExamples` and the existing `TestTelegramDevelopment` cover the main path through real CLI/Core/official Adapter/example processes and a local Bot API. Real Telegram acceptance follows the records for its original work slice.
