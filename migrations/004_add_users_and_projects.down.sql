-- Drop indexes
DROP INDEX IF EXISTS idx_environments_project_key;
DROP INDEX IF EXISTS idx_project_members_project_user;
DROP INDEX IF EXISTS idx_projects_key;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in reverse order
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;