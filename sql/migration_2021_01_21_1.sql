ALTER TABLE ketchup.ketchup
  ADD COLUMN kind TEXT NULL DEFAULT 'release'
  ADD COLUMN upstream TEXT NULL;

ALTER TABLE ketchup.ketchup RENAME COLUMN version TO current;

UPDATE
  ketchup.ketchup
SET
  k.upstream = r.version
FROM
  ketchup.ketchup k,
  ketchup.repository r
WHERE
  r.id = k.repository_id;
