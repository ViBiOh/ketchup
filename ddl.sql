-- clean
DROP TABLE IF EXISTS ketchup;
DROP TABLE IF EXISTS repository;
DROP TABLE IF EXISTS "user";

DROP INDEX IF EXISTS ketchup_user;
DROP INDEX IF EXISTS repository_id;
DROP INDEX IF EXISTS repository_repository;
DROP INDEX IF EXISTS user_id;

DROP SEQUENCE IF EXISTS repository_seq;
DROP SEQUENCE IF EXISTS user_seq;

-- user
CREATE SEQUENCE user_seq;
CREATE TABLE "user" (
  id BIGINT NOT NULL DEFAULT nextval('user_seq'),
  email TEXT NOT NULL,
  login_id BIGINT NOT NULL REFERENCES login(id) ON DELETE CASCADE,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE user_seq OWNED BY "user".id;

CREATE UNIQUE INDEX user_id ON "user"(id);
CREATE UNIQUE INDEX user_login_id ON "user"(login_id);
CREATE UNIQUE INDEX user_email ON "user"(email);

-- repository
CREATE SEQUENCE repository_seq;
CREATE TABLE repository (
  id BIGINT NOT NULL DEFAULT nextval('repository_seq'),
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE repository_seq OWNED BY repository.id;

CREATE UNIQUE INDEX repository_id ON repository(id);
CREATE UNIQUE INDEX repository_repository ON repository(name);

-- ketchup
CREATE TABLE ketchup (
  user_id BIGINT NOT NULL REFERENCES "user"(id),
  repository_id BIGINT NOT NULL REFERENCES repository(id),
  version TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNIQUE INDEX ketchup_user ON ketchup(user_id);
