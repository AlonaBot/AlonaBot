# 未发布

### Added

- 富消息发送可通过 `sdk.RequireOriginalSegmentIndexes()` 要求原始片段回执；任一 BeforeSend 改链时，Core 在平台副作用前明确拒绝，交互和定时发送使用同一保证。

- 插件可按目标实例查询出站格式、扩展用途与例子及组合规则，并通过 `sdk.DescribeMessage`、`sdk.CheckMessage` 或对应 Context 方法执行只读查询和预检。旧 Adapter 的标准格式保留 `unknown`，已有文本限额继续可读。
- Adapter 可通过 `Descriptor.Formats`、`Descriptor.Combinations`、`Extension.Describe` 和可选 `MessageChecker` 公布现行能力与组合校验。预检保留来源类别并使用 Core 授权后的可信 MIME，无媒体取件或平台发送。
- `sdk/tool.ConfirmRequired` 可在操作策略为 `allow` 时要求可信交互用户确认；`deny` 优先拒绝，原 `Confirm` 和发送入口保持现行策略语义。

### Fixed

- Core 自动添加引用、提及、前缀或封装合并转发后，发送回执保留调用方原始片段下标，避免部分成功时误选补发片段。
- 出站格式查询按实例的 `message.reference` 能力更新 quote 支持状态。
