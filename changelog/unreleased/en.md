# Unreleased

### Added

- Rich sends can require caller-input receipt indexes with `sdk.RequireOriginalSegmentIndexes()`. Core explicitly rejects any BeforeSend rewrite before platform side effects, for both interactive and scheduled sends.

- Plugins can query target-specific outbound formats, extension purposes/examples and combination rules, and run read-only checks through `sdk.DescribeMessage`, `sdk.CheckMessage` or their Context methods. Standard formats from older Adapters remain `unknown` while existing text limits stay available.
- Adapters can describe current behavior through `Descriptor.Formats`, `Descriptor.Combinations`, `Extension.Describe` and an optional `MessageChecker`. Checks preserve source kinds and use Core-authorized MIME metadata without fetching media or sending to the platform.
- `sdk/tool.ConfirmRequired` requires a trusted interactive user's confirmation even under an `allow` operation policy. `deny` takes precedence; existing `Confirm` and send entry points retain their policy semantics.

### Fixed

- Strict receipt checks compare Hook responses after completing reference targets, allowing equal SDK round trips while continuing to reject actual rewrites.

- Send receipts retain caller-input segment indexes after Core adds reply quotes, mentions, prefixes or a forward wrapper, preserving the correct parts for partial-delivery recovery.
- Outbound format queries use the instance's `message.reference` capability to update quote support.
