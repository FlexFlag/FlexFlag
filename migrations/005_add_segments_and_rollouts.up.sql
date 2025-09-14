-- First, add project_id to flags table
ALTER TABLE flags ADD COLUMN IF NOT EXISTS project_id UUID;
ALTER TABLE flags ADD COLUMN IF NOT EXISTS rollout_config JSONB DEFAULT '{}';
ALTER TABLE flags ADD COLUMN IF NOT EXISTS experiment_config JSONB DEFAULT '{}';

-- Create default project if it doesn't exist
INSERT INTO projects (key, name, description) 
VALUES ('default', 'Default Project', 'Default project for existing flags')
ON CONFLICT (key) DO NOTHING;

-- Update existing flags to belong to default project
UPDATE flags 
SET project_id = (SELECT id FROM projects WHERE key = 'default' LIMIT 1)
WHERE project_id IS NULL;

-- Now add the foreign key constraint
ALTER TABLE flags ADD CONSTRAINT flags_project_id_fkey 
FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- Make project_id NOT NULL after updating existing data
ALTER TABLE flags ALTER COLUMN project_id SET NOT NULL;

-- Drop and recreate segments table with proper project reference
DROP TABLE IF EXISTS segments CASCADE;
CREATE TABLE segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rules JSONB NOT NULL DEFAULT '[]',
    environment VARCHAR(50) NOT NULL DEFAULT 'production',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, key, environment)
);

-- Add rollouts table for A/B testing and percentage rollouts
CREATE TABLE IF NOT EXISTS rollouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_id UUID NOT NULL REFERENCES flags(id) ON DELETE CASCADE,
    environment VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'percentage', -- percentage, experiment, segment
    name VARCHAR(255) NOT NULL,
    description TEXT,
    config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, paused, completed
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add audit_logs table for tracking all changes
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    user_id UUID REFERENCES users(id),
    resource_type VARCHAR(50) NOT NULL, -- flag, segment, rollout, user, project
    resource_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- create, update, delete, toggle, deploy
    old_values JSONB,
    new_values JSONB,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add sticky assignments table for consistent user experiences
CREATE TABLE IF NOT EXISTS sticky_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_id UUID REFERENCES flags(id) ON DELETE CASCADE,
    environment VARCHAR(100) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    user_key VARCHAR(255) NOT NULL,
    variation_id VARCHAR(100) NOT NULL,
    bucket_key VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(flag_id, environment, user_id, user_key)
);

-- Add indexes for performance
CREATE INDEX idx_segments_project_key ON segments(project_id, key);
CREATE INDEX idx_rollouts_flag_env ON rollouts(flag_id, environment);
CREATE INDEX idx_rollouts_status ON rollouts(status);
CREATE INDEX idx_flags_project_env ON flags(project_id, environment);
CREATE INDEX idx_audit_logs_project ON audit_logs(project_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
CREATE INDEX idx_sticky_assignments_flag_env_user ON sticky_assignments(flag_id, environment, user_id, user_key);