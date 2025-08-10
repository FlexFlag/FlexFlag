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
DROP TABLE IF EXISTS sticky_assignments;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS rollouts;
DROP TABLE IF EXISTS segments;

-- Remove columns from flags table
ALTER TABLE flags DROP COLUMN IF EXISTS experiment_config;
ALTER TABLE flags DROP COLUMN IF EXISTS rollout_config;
ALTER TABLE flags DROP COLUMN IF EXISTS project_id;