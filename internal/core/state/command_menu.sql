-- 命令菜单按平台账号保存模式、聊天关系与最后成功写入的位置。
CREATE TABLE command_menu_account (
    id TEXT PRIMARY KEY NOT NULL,
    adapter_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    mode TEXT NOT NULL,
    ready INTEGER NOT NULL,
    last_error TEXT NOT NULL,
    clear_automatic INTEGER NOT NULL
);
CREATE TABLE command_menu_chat (
    id TEXT PRIMARY KEY NOT NULL,
    adapter_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    chat_id TEXT NOT NULL,
    active INTEGER NOT NULL
);
CREATE TABLE command_menu_position (
    id TEXT PRIMARY KEY NOT NULL,
    adapter_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    scope TEXT NOT NULL,
    chat_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    language_code TEXT NOT NULL,
    mode TEXT NOT NULL,
    owner_plugin_id TEXT NOT NULL
);
