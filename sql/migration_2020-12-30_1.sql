ALTER TYPE ketchup.repository_type RENAME TO repository_kind;
ALTER TABLE ketchup.repository RENAME COLUMN type TO kind;
