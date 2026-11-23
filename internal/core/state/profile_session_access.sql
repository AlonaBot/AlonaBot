CREATE TABLE profile_session_access (
  profile_id TEXT NOT NULL PRIMARY KEY REFERENCES profile(id) ON DELETE CASCADE,
  mode TEXT NOT NULL CHECK (mode IN ('all', 'allow', 'deny')),
  sessions TEXT NOT NULL
);
