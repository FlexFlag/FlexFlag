-- Drop indexes
DROP INDEX IF EXISTS idx_sticky_assignments_flag_env_user;
DROP INDEX IF EXISTS idx_audit_logs_created;
DROP INDEX IF EXISTS idx_audit_logs_user;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_project;
DROP INDEX IF EXISTS idx_flags_project_env;
DROP INDEX IF EXISTS idx_rollouts_status;
DROP INDEX IF EXISTS idx_rollouts_flag_env;
DROP INDEX IF EXISTS idx_segments_project_key;

-- Drop tables
DROP TABLE IF EXISTS sticky_assignments CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS rollouts CASCADE;
DROP TABLE IF EXISTS segments CASCADE;

-- Remove foreign key constraint first
ALTER TABLE flags DROP CONSTRAINT IF EXISTS flags_project_id_fkey;

-- Remove columns from flags table
ALTER TABLE flags DROP COLUMN IF EXISTS experiment_config;
ALTER TABLE flags DROP COLUMN IF EXISTS rollout_config;
ALTER TABLE flags DROP COLUMN IF EXISTS project_id;

-- Recreate the simple segments table from migration 002
CREATE TABLE IF NOT EXISTS segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rules JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);