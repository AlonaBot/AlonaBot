## [0.3.1] - 2026-11-23

### Fixed

- Core creates editable configuration snapshots after activation and restores missing snapshots from Active configuration before reload. Successful reloads save normalized configuration while preserving unapplied disk edits during reload, update and rollback.
- Deleting an Adapter instance clears its persisted event cursor, so a replacement using the same instance name starts with a fresh cursor.
- Deleting instances also removes their Profile session access-list entries, preserving the access mode and entries for other instances.

### Upgrade and compatibility

- Upgrade Core to v0.3.1; existing v0.3.0 official components remain supported. Compose uses `ghcr.io/alonabot/alonabot:0.3.1` with Agent and WebUI v0.3.0.
- Existing configuration values, Secret references and managed data are retained. The state schema remains v3. Back up the complete data directory before upgrading.

[0.3.1]: https://github.com/AlonaBot/AlonaBot/releases/tag/v0.3.1
