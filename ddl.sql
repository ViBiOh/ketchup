-- Cleaning
DROP TABLE IF EXISTS target;

DROP INDEX IF EXISTS target_id;

DROP SEQUENCE IF EXISTS target_id_seq;

-- target
CREATE SEQUENCE target_seq;
CREATE TABLE target (
  id BIGINT NOT NULL DEFAULT nextval('target_seq'),
  repository TEXT NOT NULL,
  current_version TEXT NOT NULL,
  latest_version TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE target_seq OWNED BY target.id;

CREATE UNIQUE INDEX target_id ON target (id);
CREATE UNIQUE INDEX target_repository ON target (repository);
