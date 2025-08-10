-- Revert to the old global unique constraint (rollback migration)

-- Drop the project-scoped constraint
ALTER TABLE flags DROP CONSTRAINT IF EXISTS flags_project_key_environment_key;

-- Drop the associated index
DROP INDEX IF EXISTS idx_flags_project_key_env;

-- Add back the old global constraint
ALTER TABLE flags ADD CONSTRAINT flags_key_environment_key UNIQUE(key, environment);