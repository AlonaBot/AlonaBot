-- 测试插件的单行持久计数器。
CREATE TABLE IF NOT EXISTS counter (id INTEGER PRIMARY KEY CHECK (id = 1), value INTEGER NOT NULL);
INSERT OR IGNORE INTO counter (id, value) VALUES (1, 0);
