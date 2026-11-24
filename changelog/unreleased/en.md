# Unreleased

### Fixed

- Return Send/Edit receipts independently; manage ordered AfterSend batches with bounded admission and Core draining (#325).
- Apply event deadlines to queueing and PreEvent; skip business processing for expired queued events (#324).
- Preserve original Handler failures and valid STOP outcomes consistently for commands and events (#326).
- Store reply decorations in Session-isolated runtime SQLite, preserving historical edits without a growing in-memory message map (#323).
- Default omitted plugin/agent `sleep_capable` to true, preserve omission when building, and document rebuilding old packages (#322).
- Show self-contained natural-language messages at every log level, including DEBUG state transitions and failure reasons, while preserving complete JSON files, searchable details, legacy history and Agent model/tool diagnostics (#327).
