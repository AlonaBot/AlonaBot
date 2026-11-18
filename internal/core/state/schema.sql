CREATE TABLE schema_version (
  version INTEGER NOT NULL
);

INSERT INTO schema_version (version) VALUES (0);

CREATE TABLE install (
  plugin_id    TEXT PRIMARY KEY,
  install_seq  INTEGER NOT NULL UNIQUE,
  source       TEXT NOT NULL,
  entrypoint   TEXT NOT NULL,
  build_digest BLOB NOT NULL,
  created_at   INTEGER NOT NULL,
  role         TEXT NOT NULL,
  adapter_id   TEXT NOT NULL,
  version      TEXT NOT NULL,
  publisher_key_id TEXT NOT NULL,
  pin          TEXT NOT NULL,
  uninstalled  INTEGER NOT NULL
);

CREATE TABLE generation (
  generation_id   TEXT PRIMARY KEY,
  plugin_id       TEXT NOT NULL REFERENCES install(plugin_id),
  config_snapshot TEXT NOT NULL,
  config_corrections TEXT NOT NULL,
  created_at      INTEGER NOT NULL,
  committed_at    INTEGER
);

CREATE TABLE route_snapshot (
  generation_id TEXT PRIMARY KEY REFERENCES generation(generation_id) ON DELETE CASCADE,
  plugin_id     TEXT NOT NULL REFERENCES install(plugin_id),
  routes        BLOB NOT NULL,
  route_hash    BLOB NOT NULL,
  build_digest  BLOB NOT NULL
);

CREATE TABLE plugin_state (
  plugin_id         TEXT PRIMARY KEY REFERENCES install(plugin_id),
  active_generation TEXT REFERENCES generation(generation_id),
  enabled           INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE plugin_runtime (
  plugin_id        TEXT PRIMARY KEY REFERENCES install(plugin_id),
  state            TEXT NOT NULL CHECK (state IN ('STOPPED', 'WAKING', 'ACTIVE', 'QUIESCING', 'SUSPENDED', 'FAILED')),
  reason           TEXT NOT NULL,
  changed_at       DATETIME NOT NULL,
  last_wake_reason TEXT NOT NULL,
  last_wake_at     DATETIME NOT NULL,
  last_suspend_at  DATETIME NOT NULL
);

CREATE TABLE pending_config (
  plugin_id       TEXT PRIMARY KEY REFERENCES install(plugin_id),
  base_generation TEXT NOT NULL,
  config_json     TEXT NOT NULL,
  updated_at      INTEGER NOT NULL
);

CREATE TABLE adapter_cursor (
  instance_id TEXT PRIMARY KEY,
  plugin_id   TEXT NOT NULL REFERENCES install(plugin_id),
  cursor      BLOB NOT NULL,
  updated_at  INTEGER NOT NULL
);

CREATE TABLE chat (
  adapter_id  TEXT NOT NULL,
  chat_id     TEXT NOT NULL,
  chat_type   TEXT NOT NULL,
  title       TEXT NOT NULL,
  instance_id TEXT NOT NULL,
  first_seen  INTEGER NOT NULL,
  last_seen   INTEGER NOT NULL,
  PRIMARY KEY (adapter_id, chat_id)
);

CREATE INDEX chat_last_seen ON chat(last_seen DESC, adapter_id, chat_id);

CREATE TABLE plugin_adapter_permission (
  plugin_id           TEXT NOT NULL REFERENCES install(plugin_id),
  adapter_instance_id TEXT NOT NULL,
  PRIMARY KEY (plugin_id, adapter_instance_id)
);

CREATE TABLE plugin_adapter_restriction (
  plugin_id TEXT PRIMARY KEY REFERENCES install(plugin_id)
);

CREATE TABLE conversation_entry (
  plugin_id        TEXT NOT NULL REFERENCES install(plugin_id),
  conversation_key TEXT NOT NULL,
  seq              INTEGER NOT NULL,
  entry_json       TEXT NOT NULL,
  created_at       INTEGER NOT NULL,
  PRIMARY KEY (plugin_id, conversation_key, seq)
);

CREATE TABLE secret (
  ref_id     TEXT PRIMARY KEY,
  plugin_id  TEXT NOT NULL REFERENCES install(plugin_id),
  value      BLOB NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE generation_secret_path (
  generation_id TEXT NOT NULL REFERENCES generation(generation_id),
  path          TEXT NOT NULL,
  PRIMARY KEY (generation_id, path)
);

CREATE TABLE admin_credential (
  id              INTEGER PRIMARY KEY CHECK (id = 1),
  token_sha256    BLOB NOT NULL,
  auth_generation INTEGER NOT NULL,
  updated_at      INTEGER NOT NULL
);

CREATE TABLE admin_session (
  session_id      TEXT PRIMARY KEY,
  csrf_token      TEXT NOT NULL,
  auth_generation INTEGER NOT NULL,
  expires_at      INTEGER NOT NULL
);

CREATE TABLE command_policy (
  plugin_id   TEXT NOT NULL REFERENCES install(plugin_id),
  scope       TEXT NOT NULL,
  permissions TEXT,
  enabled     INTEGER,
  names       TEXT,
  PRIMARY KEY (plugin_id, scope)
);

CREATE TABLE global_deny (
  kind       TEXT NOT NULL,
  adapter_id TEXT NOT NULL,
  subject_id TEXT NOT NULL,
  PRIMARY KEY (kind, adapter_id, subject_id)
);

CREATE TABLE tool_policy (
  plugin_id    TEXT NOT NULL REFERENCES install(plugin_id),
  service_name TEXT NOT NULL,
  policy       INTEGER NOT NULL,
  PRIMARY KEY (plugin_id, service_name)
);

CREATE TABLE button_token (
  token                 TEXT PRIMARY KEY,
  purpose               TEXT NOT NULL CHECK (purpose IN ('plugin', 'confirm')),
  plugin_id             TEXT NOT NULL REFERENCES install(plugin_id),
  route_id              TEXT NOT NULL,
  action_id             TEXT NOT NULL,
  action_schema_version INTEGER NOT NULL,
  generation_id         TEXT NOT NULL REFERENCES generation(generation_id),
  adapter_id            TEXT NOT NULL,
  instance_id           TEXT NOT NULL,
  account_id            TEXT NOT NULL,
  chat_id               TEXT NOT NULL,
  thread_id             TEXT NOT NULL,
  value                 TEXT NOT NULL,
  expires_at            INTEGER NOT NULL
);

CREATE INDEX button_token_plugin ON button_token(plugin_id);

CREATE TABLE dynamic_task (
  plugin_id TEXT NOT NULL REFERENCES install(plugin_id),
  name TEXT NOT NULL,
  key TEXT NOT NULL,
  at INTEGER NOT NULL,
  interval INTEGER NOT NULL,
  cron TEXT NOT NULL,
  timezone TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  anchor INTEGER NOT NULL,
  created_at INTEGER NOT NULL,
  PRIMARY KEY (plugin_id, name, key)
);

CREATE TABLE scheduled_job (
  plugin_id TEXT NOT NULL REFERENCES install(plugin_id),
  job_id TEXT NOT NULL,
  anchor INTEGER NOT NULL,
  interval INTEGER NOT NULL,
  paused INTEGER NOT NULL,
  PRIMARY KEY (plugin_id, job_id)
);

CREATE TABLE journal (
  op_id TEXT PRIMARY KEY,
  plugin_id TEXT NOT NULL REFERENCES install(plugin_id),
  kind TEXT NOT NULL,
  expected_old TEXT NOT NULL,
  new_generation TEXT NOT NULL,
  phase TEXT NOT NULL,
  package_dir TEXT NOT NULL,
  database_path TEXT NOT NULL,
  entrypoint TEXT NOT NULL,
  version TEXT NOT NULL,
  publisher_key_id TEXT NOT NULL,
  build_digest BLOB NOT NULL,
  session_id BLOB NOT NULL,
  owns_package INTEGER NOT NULL,
  was_running INTEGER NOT NULL,
  was_enabled INTEGER NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE TABLE profile (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  color TEXT NOT NULL,
  profile_seq INTEGER NOT NULL UNIQUE
);

INSERT INTO profile (id, name, color, profile_seq) VALUES ('default', '默认', 'teal', 0);

CREATE TABLE profile_assignment (
  instance TEXT NOT NULL,
  session TEXT NOT NULL,
  profile_id TEXT NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  PRIMARY KEY (instance, session)
);

CREATE TABLE profile_override (
  profile_id TEXT NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  plugin_id TEXT NOT NULL REFERENCES install(plugin_id) ON DELETE CASCADE,
  path TEXT NOT NULL,
  value TEXT NOT NULL,
  PRIMARY KEY (profile_id, plugin_id, path)
);

CREATE TABLE profile_plugin (
  profile_id TEXT NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  plugin_id TEXT NOT NULL REFERENCES install(plugin_id) ON DELETE CASCADE,
  enabled INTEGER NOT NULL,
  PRIMARY KEY (profile_id, plugin_id)
);

CREATE TABLE profile_tool (
  profile_id TEXT NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  plugin_id TEXT NOT NULL REFERENCES install(plugin_id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  operation INTEGER NOT NULL,
  policy TEXT NOT NULL,
  PRIMARY KEY (profile_id, plugin_id, name, operation)
);

CREATE TABLE profile_platform (
  profile_id TEXT PRIMARY KEY REFERENCES profile(id) ON DELETE CASCADE,
  settings TEXT NOT NULL
);
