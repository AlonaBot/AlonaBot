## [0.3.1] - 2026-11-23

### Fixed

- Core creates an editable configuration file when a component is first enabled. If the file is missing at reload, Core recreates it from the active configuration. A successful reload writes the active configuration back to disk. Reloading, updating or rolling back preserves any edits that have yet to be applied ([#321](https://github.com/AlonaBot/AlonaBot/pull/321)).
- Deleting an adapter instance removes its event cursor and its entries in the profile's session access list. A replacement with the same name starts with a fresh cursor. The access mode and entries for other instances stay in place ([#321](https://github.com/AlonaBot/AlonaBot/pull/321)).

The `ghcr.io/alonabot/alonabot:0.3.1` Compose image includes Agent and WebUI v0.3.0.

[0.3.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.1
