-- Add missing columns to projects table
-- The Go code expects is_active and settings columns that are currently missing

ALTER TABLE projects ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}';