## [0.3.2] - 2026-11-25

### Fixed

- 平台完成发送或编辑后，`Send` 和 `Edit` 不再等待 `AfterSend` 钩子执行完毕，调用方直接收到平台回执及平台返回的错误。Core 在后台按既定顺序运行同一批钩子 ([#328](https://github.com/AlonaBot/AlonaBot/pull/328))。
- 事件在队列中等待的时间也计入处理期限。出队时已经过期的事件会跳过 `PreEvent` 和业务处理；其余事件按剩余时间继续处理 ([#329](https://github.com/AlonaBot/AlonaBot/pull/329))。
- 命令和事件处理器调用失败时，Core 会保留原始错误，避免把缺失的响应误报为解码失败。失败响应中有效的 `STOP` 仍会阻止后续处理 ([#330](https://github.com/AlonaBot/AlonaBot/pull/330))。
- 在同一会话内编辑早先发出的回复时，Core 仍能保留原附带的提及和前缀；这些关联不再随着发送量不断累积在内存里 ([#331](https://github.com/AlonaBot/AlonaBot/pull/331))。
- 插件和 Agent 省略 `sleep_capable` 时，默认允许 Core 在符合休眠条件时让组件休眠；显式设置为 `false` 时仍保持常驻。如果旧构建器把源码中省略的值写成了 `false`，请用修复后的构建器重新打包，让该组件采用默认行为 ([#332](https://github.com/AlonaBot/AlonaBot/pull/332))。
- Core 和 SDK 的日志正文会直接写明哪个组件执行了什么操作、状态如何变化，以及失败的原因。完整的结构化字段保存在 JSON 日志文件中，旧的文本历史仍可查询 ([#333](https://github.com/AlonaBot/AlonaBot/pull/333))。

[0.3.2]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.2
