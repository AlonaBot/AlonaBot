-- 保存每次启动读到的配置，展示开发数据库跨启动保留数据。
CREATE TABLE IF NOT EXISTS startup (
    id INTEGER PRIMARY KEY,
    greeting TEXT NOT NULL
);
