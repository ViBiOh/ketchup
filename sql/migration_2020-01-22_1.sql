-- repository_version
CREATE TABLE ketchup.repository_version (
  repository_id BIGINT NOT NULL REFERENCES ketchup.repository(id),
  pattern TEXT NOT NULL DEFAULT 'stable',
  version TEXT NOT NULL
);

INSERT INTO ketchup.repository_version (repository_id, version)
SELECT id, version
FROM ketchup.repository;
