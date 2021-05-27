DROP INDEX IF EXISTS ketchup_id;
CREATE UNIQUE INDEX ketchup_id ON ketchup.ketchup(user_id, repository_id, pattern);
