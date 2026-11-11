-- 笔记的初始结构，由第一步 versioned 迁移执行。
CREATE TABLE configdb_note (key TEXT PRIMARY KEY NOT NULL, value TEXT NOT NULL);
-- 启动记录单独存放，避免与用户的笔记键冲突。
CREATE TABLE configdb_startup (key TEXT PRIMARY KEY NOT NULL, value TEXT NOT NULL);
