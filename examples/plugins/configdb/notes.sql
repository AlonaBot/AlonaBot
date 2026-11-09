-- 笔记的初始结构，由第一步 versioned 迁移执行。
CREATE TABLE configdb_note (key TEXT PRIMARY KEY NOT NULL, value TEXT NOT NULL);
