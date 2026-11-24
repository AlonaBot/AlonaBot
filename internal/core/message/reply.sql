CREATE TABLE reply_decoration (
  session TEXT NOT NULL,
  target_key TEXT NOT NULL,
  prefix TEXT NOT NULL,
  mention BLOB,
  PRIMARY KEY (session, target_key)
);
