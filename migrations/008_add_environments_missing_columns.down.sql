-- Remove the added columns
ALTER TABLE environments DROP COLUMN IF EXISTS is_active;
ALTER TABLE environments DROP COLUMN IF EXISTS settings;