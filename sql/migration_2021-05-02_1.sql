CREATE TABLE ketchup.repository_new (
  id BIGINT NOT NULL DEFAULT nextval('ketchup.repository_seq'),
  kind ketchup.repository_kind NOT NULL,
  name TEXT NOT NULL,
  part TEXT,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);

INSERT INTO ketchup.repository_new (id, kind, name, part, creation_date) SELECT id, kind, name, '', creation_date from ketchup.repository where kind = 'github';
INSERT INTO ketchup.repository_new (id, kind, name, part, creation_date) SELECT id, kind, SPLIT_PART(name, '@', 2), SPLIT_PART(name, '@', 1), creation_date from ketchup.repository where kind = 'helm';

ALTER TABLE ketchup.repository_version DROP CONSTRAINT repository_version_repository_id_fkey;
ALTER TABLE ketchup.ketchup DROP CONSTRAINT ketchup_repository_id_fkey;

ALTER SEQUENCE ketchup.repository_seq OWNED BY ketchup.repository_new.id;

DROP TABLE ketchup.repository;
DROP INDEX IF EXISTS repository_id;
DROP INDEX IF EXISTS repository_repository;

ALTER TABLE ketchup.repository_new RENAME TO repository;

CREATE UNIQUE INDEX repository_id ON ketchup.repository(id);
CREATE UNIQUE INDEX repository_repository ON ketchup.repository(name, part);

ALTER TABLE ketchup.repository_version
  ADD CONSTRAINT repository_version_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES repository(id) ON DELETE CASCADE;

ALTER TABLE ketchup.ketchup
  ADD CONSTRAINT ketchup_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES repository(id);
