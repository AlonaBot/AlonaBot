# Unreleased

## Added

- Add chat and account scoped command menus with typed Set/Get/Delete/Status/Sync, ready events, persisted ownership and lifecycle synchronization.
- Preserve trusted MIME metadata when Core prepares media sources for adapters, including GIF identification.

- Provide component and Core changelog management APIs.

## Changed

- Command registration now only registers handlers by default; add `.Menu()` to root commands that should appear automatically. Scoped Telegram menus require the SDK/Core with CommandMenuController; existing adapters retain SyncCommands.

- Provide English and Simplified Chinese READMEs with language links, documentation homepages and a shared Views badge; move brand images to `docs/assets`.
- Generate plugin and agent projects without an i18n directory; the example replies with `Hello AlonaBot` directly. Generated Go files and `plugin.yaml` contain no comments. Add translation files when needed.

## Fixed

- Return published Core changelogs in English or Simplified Chinese according to `Accept-Language`, with English fallback. Exclude unreleased changes from the API.
