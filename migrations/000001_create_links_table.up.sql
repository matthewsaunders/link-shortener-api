CREATE TABLE IF NOT EXISTS links
(
  id         uuid DEFAULT uuid_generate_v4 (),
  name       TEXT,
  source     TEXT,
  token      TEXT UNIQUE,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  version    INTEGER NOT NULL DEFAULT 1,
  PRIMARY KEY (id)
);
