# 未发布

### Added

- 插件可按目标实例查询出站格式、扩展用途与例子及组合规则，并通过 `sdk.DescribeMessage`、`sdk.CheckMessage` 或对应 Context 方法执行只读查询和预检。旧 Adapter 的标准格式保留 `unknown`，已有文本限额继续可读。
- Adapter 可通过 `Descriptor.Formats`、`Descriptor.Combinations`、`Extension.Describe` 和可选 `MessageChecker` 公布现行能力与组合校验。预检保留来源类别并使用 Core 授权后的可信 MIME，无媒体取件或平台发送。
- `sdk/tool.ConfirmRequired` 可在操作策略为 `allow` 时要求可信交互用户确认；`deny` 优先拒绝，原 `Confirm` 和发送入口保持现行策略语义。
