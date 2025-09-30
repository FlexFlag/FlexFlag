-- Remove the admin user fix
-- This rollback removes the updated admin user

DELETE FROM users WHERE email = 'admin@example.com';