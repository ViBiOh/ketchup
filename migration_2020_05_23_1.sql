CREATE SCHEMA auth;

ALTER TABLE login_profile SET SCHEMA auth;
ALTER TABLE profile SET SCHEMA auth;
ALTER TABLE login SET SCHEMA auth;

CREATE SCHEMA ketchup;

ALTER TABLE "user" SET SCHEMA ketchup;
ALTER TABLE ketchup SET SCHEMA ketchup;
ALTER TABLE repository SET SCHEMA ketchup;
