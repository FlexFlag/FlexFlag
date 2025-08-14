-- Add missing columns to environments table
ALTER TABLE environments ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE environments ADD COLUMN IF NOT EXISTS settings JSONB;