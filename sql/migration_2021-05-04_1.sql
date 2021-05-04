CREATE TYPE ketchup.ketchup_frequency AS ENUM ('none', 'daily', 'weekly');

ALTER TABLE ketchup.ketchup
  ADD COLUMN frequency ketchup_frequency NOT NULL DEFAULT 'daily';
