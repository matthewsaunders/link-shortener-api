CREATE TABLE IF NOT EXISTS visits
(
  id              uuid DEFAULT uuid_generate_v4 (),
  link_id         uuid NOT NULL,
  created_at      TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  referrer        TEXT,
  remote_address  TEXT,
  PRIMARY KEY (id)
);
