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
  adapter_id   TEXT NOT NULL
);

CREATE TABLE generation (
  generation_id   TEXT PRIMARY KEY,
  plugin_id       TEXT NOT NULL REFERENCES install(plugin_id),
  config_snapshot TEXT NOT NULL,
  config_version  INTEGER NOT NULL,
  created_at      INTEGER NOT NULL,
  committed_at    INTEGER
);

CREATE TABLE plugin_state (
  plugin_id         TEXT PRIMARY KEY REFERENCES install(plugin_id),
  active_generation TEXT REFERENCES generation(generation_id),
  enabled           INTEGER NOT NULL DEFAULT 0
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
