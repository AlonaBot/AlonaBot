-- 受管 SQLite 的 KV 保留表，独立于插件迁移账本。
CREATE TABLE IF NOT EXISTS _alonabot_kv (
    key TEXT PRIMARY KEY,
    value BLOB NOT NULL,
    updated_at INTEGER NOT NULL
);
