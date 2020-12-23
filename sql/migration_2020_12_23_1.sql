CREATE TYPE ketchup.repository_type AS ENUM ('github', 'helm');

ALTER TABLE ketchup.repository
  ADD COLUMN type ketchup.repository_type NULL;

UPDATE ketchup.repository SET type = 'github';

ALTER TABLE ketchup.repository
  ALTER COLUMN type SET NOT NULL;
