-- clean
DROP TABLE IF EXISTS ketchup.ketchup;
DROP TABLE IF EXISTS ketchup.repository_version;
DROP TABLE IF EXISTS ketchup.repository;
DROP TABLE IF EXISTS ketchup.user;

DROP TYPE IF EXISTS ketchup.repository_kind;

DROP INDEX IF EXISTS ketchup_user;
DROP INDEX IF EXISTS repository_id;
DROP INDEX IF EXISTS repository_repository;
DROP INDEX IF EXISTS user_id;

DROP SEQUENCE IF EXISTS ketchup.repository_seq;
DROP SEQUENCE IF EXISTS ketchup.user_seq;

DROP SCHEMA IF EXISTS ketchup CASCADE;

-- schema
CREATE SCHEMA ketchup;

-- user
CREATE SEQUENCE ketchup.user_seq;
CREATE TABLE ketchup.user (
  id BIGINT NOT NULL DEFAULT nextval('ketchup.user_seq'),
  email TEXT NOT NULL,
  login_id BIGINT NOT NULL REFERENCES auth.login(id) ON DELETE CASCADE,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE ketchup.user_seq OWNED BY ketchup.user.id;

CREATE UNIQUE INDEX user_id ON ketchup.user(id);
CREATE UNIQUE INDEX user_login_id ON ketchup.user(login_id);
CREATE UNIQUE INDEX user_email ON ketchup.user(email);

-- repository_kind
CREATE TYPE ketchup.repository_kind AS ENUM ('github', 'helm');

-- repository
CREATE SEQUENCE ketchup.repository_seq;
CREATE TABLE ketchup.repository (
  id BIGINT NOT NULL DEFAULT nextval('ketchup.repository_seq'),
  name TEXT NOT NULL,
  kind ketchup.repository_kind NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE ketchup.repository_seq OWNED BY ketchup.repository.id;

CREATE UNIQUE INDEX repository_id ON ketchup.repository(id);
CREATE UNIQUE INDEX repository_repository ON ketchup.repository(name);

-- repository_version
CREATE TABLE ketchup.repository_version (
  repository_id BIGINT NOT NULL REFERENCES ketchup.repository(id) ON DELETE CASCADE,
  pattern TEXT NOT NULL DEFAULT 'stable',
  version TEXT NOT NULL
);

CREATE UNIQUE INDEX repository_version_id ON ketchup.repository_version(repository_id, pattern);

-- ketchup
CREATE TABLE ketchup.ketchup (
  user_id BIGINT NOT NULL REFERENCES ketchup.user(id) ON DELETE CASCADE,
  repository_id BIGINT NOT NULL REFERENCES ketchup.repository(id),
  pattern TEXT NOT NULL default 'stable',
  version TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNIQUE INDEX ketchup_id ON ketchup.ketchup(user_id, repository_id);

-- data
DO $$
  DECLARE login_id BIGINT;
  DECLARE profile_id BIGINT;
  BEGIN
    INSERT INTO auth.login (login, password) VALUES ('scheduler', 'service-account') RETURNING id INTO login_id;
    INSERT INTO auth.profile (name) VALUES ('admin') RETURNING id INTO profile_id;
    INSERT INTO auth.login_profile (login_id, profile_id) VALUES (login_id, profile_id);
END $$;
