-- Cleaning
DROP TABLE IF EXISTS target;

DROP INDEX IF EXISTS target_id;

DROP SEQUENCE IF EXISTS target_id_seq;

-- target
CREATE SEQUENCE target_seq;
CREATE TABLE target (
  id BIGINT NOT NULL DEFAULT nextval('target_seq'),
  owner TEXT NOT NULL,
  repository TEXT NOT NULL,
  version TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE target_seq OWNED BY target.id;

CREATE UNIQUE INDEX target_id ON target (id);
