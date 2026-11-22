# Unreleased

## Added

- Add chat and account scoped command menus with typed Set/Get/Delete/Status/Sync, ready events, persisted ownership and lifecycle synchronization.
- Preserve trusted MIME metadata when Core prepares media sources for adapters, including GIF identification.

## Changed

- Command registration now only registers handlers by default; add `.Menu()` to root commands that should appear automatically. Scoped Telegram menus require the SDK/Core with CommandMenuController; existing adapters retain SyncCommands.

- Provide English and Simplified Chinese user and developer guides, with shared bilingual pages for short navigation and example notes. Keep internal maintenance and agent instructions in their original single-language form; consolidate contribution and coding rules and share runnable example procedures. Replace the retired `spec/` directory with protocol and management API help in the existing guides; protocol fields remain defined in `.proto` sources.
