-- Create API keys table for environment-specific access control
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 hash of the API key
    key_prefix VARCHAR(50) NOT NULL, -- First few characters for display
    permissions TEXT[] NOT NULL DEFAULT '{"read"}', -- Array of permissions
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Add indexes for performance
CREATE INDEX idx_api_keys_project_env ON api_keys(project_id, environment_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_expires ON api_keys(expires_at);

-- Add constraint to ensure meaningful permissions
ALTER TABLE api_keys ADD CONSTRAINT check_permissions 
    CHECK (array_length(permissions, 1) > 0 AND permissions <@ ARRAY['read', 'write', 'admin']);