-- Ensure default project has required environments
-- This migration fixes missing environments for the default project

-- Create environments for default project if they don't exist
INSERT INTO environments (project_id, key, name, description, sort_order) 
SELECT 
    p.id, 
    'production', 
    'Production', 
    'Production environment', 
    0
FROM projects p
WHERE p.key = 'default' 
  AND NOT EXISTS (
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
WHERE p.key = 'default' 
  AND NOT EXISTS (
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
WHERE p.key = 'default' 
  AND NOT EXISTS (
    SELECT 1 FROM environments e 
    WHERE e.project_id = p.id AND e.key = 'development'
  );

-- Ensure all existing projects have at least production environment
INSERT INTO environments (project_id, key, name, description, sort_order, is_active)
SELECT 
    p.id,
    'production',
    'Production',
    'Production environment',
    0,
    true
FROM projects p
WHERE NOT EXISTS (
    SELECT 1 FROM environments e 
    WHERE e.project_id = p.id
);

-- Create default admin user if no users exist
INSERT INTO users (email, password_hash, full_name, role, is_active)
SELECT 
    'admin@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LYApJ5oQTyVa.dSgi', -- Password: SecureAdmin123!
    'Default Admin',
    'admin',
    true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE role = 'admin');