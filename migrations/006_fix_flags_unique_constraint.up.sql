-- Update unique constraint to be project-scoped instead of global
-- This allows different projects to have flags with the same key

-- Drop the old constraint
ALTER TABLE flags DROP CONSTRAINT IF EXISTS flags_key_environment_key;

-- Add the new project-scoped constraint
ALTER TABLE flags ADD CONSTRAINT flags_project_key_environment_key UNIQUE(project_id, key, environment);

-- Add index for better performance on the new constraint
CREATE INDEX IF NOT EXISTS idx_flags_project_key_env ON flags(project_id, key, environment);