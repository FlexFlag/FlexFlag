-- Remove columns that were added to projects table
ALTER TABLE projects DROP COLUMN IF EXISTS is_active;
ALTER TABLE projects DROP COLUMN IF EXISTS settings;