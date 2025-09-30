-- Fix default admin user with proper password
-- This migration ensures the admin user has a working password for development setup

-- Update existing admin user or create if not exists
INSERT INTO users (email, password_hash, full_name, role, is_active)
VALUES (
    'admin@example.com',
    '$2a$10$wU9ZFK.YCfVaF53OF7AIuu2JVZ8ByRiU/vfKs0The8aO1ydKrZeTG', -- Password: admin123
    'Admin User',
    'admin',
    true
)
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;