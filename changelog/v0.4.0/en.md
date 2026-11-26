## [0.4.0] - 2026-11-26

### Added

- Plugins can ship an `assets/` directory at the project root. Packaging copies it into the component archive unchanged instead of compiling it into the executable. At runtime, call `sdk.PackageDir(ctx)` to locate the installed files, which suits large files such as fonts. In library mode it returns an empty string, so plugins need their own fallback ([#344](https://github.com/AlonaBot/AlonaBot/pull/344)).
- Plugins can use `sdk/packages` to list installed components or install one from GitHub Releases by `owner/repo`. Core only does so when the message that triggered the call came from a Core administrator. The install counts as the administrator confirming source trust and same-kind replacement, and it returns the affected items ([#344](https://github.com/AlonaBot/AlonaBot/pull/344)).
- Commands from `sdk/command.List` also include the owning plugin's name and description in the caller's language, plus the command parameters, so help plugins can group usage by plugin ([#344](https://github.com/AlonaBot/AlonaBot/pull/344)).
- When a plugin calls another plugin's Service, the callee can read the original message's chat type with `c.ChatType()`, alongside the existing `c.User()` and `c.Session()` ([#344](https://github.com/AlonaBot/AlonaBot/pull/344)).

### Changed

- Operations that components start with `tool.Confirm` now run directly when an administrator has not set a policy, instead of asking the user first. To require confirmation, set the operation to confirm or deny in the WebUI tool policies ([#344](https://github.com/AlonaBot/AlonaBot/pull/344)).
- The official source list replaces the help plugin with the builtin plugin (`alonabot-plugin-builtin-official`). An installed help plugin keeps running, but installs and updates ask to trust its source like any other ([#344](https://github.com/AlonaBot/AlonaBot/pull/344)).

[0.4.0]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.4.0
