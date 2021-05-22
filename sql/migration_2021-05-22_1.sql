ALTER TYPE ketchup.repository_kind ADD VALUE 'docker';

DROP INDEX repository_repository;
CREATE UNIQUE INDEX repository_repository ON ketchup.repository(kind, name, part);
