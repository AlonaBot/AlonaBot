-- 第二步迁移的索引；故障注入时随该步事务回滚。
CREATE INDEX configdb_note_value ON configdb_note (value);
