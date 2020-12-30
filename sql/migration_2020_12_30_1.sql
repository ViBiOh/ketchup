ALTER TYPE ketchup.repository_type RENAME TO ketchup.repository_kind
ALTER TABLE ketchup.repository RENAME COLUMN type TO kind;
