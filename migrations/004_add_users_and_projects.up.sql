-- Add users table for authentication and RBAC
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer', -- admin, editor, viewer
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add projects table for multi-project support
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add project members for project-specific permissions
CREATE TABLE IF NOT EXISTS project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer', -- admin, editor, viewer
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);

-- Add environments table for custom environments
CREATE TABLE IF NOT EXISTS environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, key)
);

-- Create default environments for existing projects
INSERT INTO environments (project_id, key, name, description, sort_order)
SELECT 
    p.id, 
    'production', 
    'Production', 
    'Production environment', 
    0
FROM projects p
WHERE NOT EXISTS (
    SELECT 1 FROM environments e 
    WHERE e.project_id = p.id AND e.key = 'production'
);

INSERT INTO environments (project_id, key, name, description, sort_order)
SELECT 
    p.id, 
    'staging', 
    'Staging', 
    'Staging environment', 
    1
FROM projects p
WHERE NOT EXISTS (
    SELECT 1 FROM environments e 
    WHERE e.project_id = p.id AND e.key = 'staging'
);

INSERT INTO environments (project_id, key, name, description, sort_order)
SELECT 
    p.id, 
    'development', 
    'Development', 
    'Development environment', 
    2
FROM projects p
WHERE NOT EXISTS (
    SELECT 1 FROM environments e 
    WHERE e.project_id = p.id AND e.key = 'development'
);

-- Add indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_projects_key ON projects(key);
CREATE INDEX idx_project_members_project_user ON project_members(project_id, user_id);
CREATE INDEX idx_environments_project_key ON environments(project_id, key);