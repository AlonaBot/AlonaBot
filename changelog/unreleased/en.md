# Unreleased

### Added

- Plugins can query target-specific outbound formats, extension purposes/examples and combination rules, and run read-only checks through `sdk.DescribeMessage`, `sdk.CheckMessage` or their Context methods. Standard formats from older Adapters remain `unknown` while existing text limits stay available.
- Adapters can describe current behavior through `Descriptor.Formats`, `Descriptor.Combinations`, `Extension.Describe` and an optional `MessageChecker`. Checks preserve source kinds and use Core-authorized MIME metadata without fetching media or sending to the platform.
- `sdk/tool.ConfirmRequired` requires a trusted interactive user's confirmation even under an `allow` operation policy. `deny` takes precedence; existing `Confirm` and send entry points retain their policy semantics.
