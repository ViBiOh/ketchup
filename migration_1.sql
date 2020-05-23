-- create schema
CREATE SCHEMA ketchup;

ALTER TABLE "user" SET SCHEMA ketchup;
ALTER TABLE ketchup SET SCHEMA ketchup;
ALTER TABLE repository SET SCHEMA ketchup;
