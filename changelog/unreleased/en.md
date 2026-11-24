# Unreleased

### Fixed

- Return Send/Edit receipts independently; manage ordered AfterSend batches with bounded admission and Core draining (#325).
- Apply event deadlines to queueing and PreEvent; skip business processing for expired queued events (#324).
- Preserve original Handler failures and valid STOP outcomes consistently for commands and events (#326).
- Store reply decorations in Session-isolated runtime SQLite, preserving historical edits without a growing in-memory message map (#323).
