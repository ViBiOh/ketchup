ALTER TABLE ketchup.user
  DROP CONSTRAINT user_login_id_fkey;

ALTER TABLE auth.basic
  DROP CONSTRAINT basic_user_id_fkey;

ALTER TABLE auth."user"
ALTER COLUMN id TYPE TEXT USING id::TEXT;

ALTER TABLE ketchup.user
ALTER COLUMN login_id TYPE TEXT USING login_id::TEXT;

ALTER TABLE auth.basic
ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;

ALTER TABLE auth.basic
  ADD CONSTRAINT basic_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.user(id) ON DELETE CASCADE;

ALTER TABLE ketchup.user
  ADD CONSTRAINT user_login_id_fkey FOREIGN KEY (login_id) REFERENCES auth.user(id) ON DELETE CASCADE;