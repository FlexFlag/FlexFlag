-- Note: PostgreSQL doesn't support removing enum values directly
-- This would require recreating the enum type and updating all references
-- For now, we'll leave a comment explaining the limitation
-- To truly rollback, you would need to:
-- 1. Create a new enum without 'variant'
-- 2. Update all tables to use the new enum
-- 3. Drop the old enum
-- 4. Rename the new enum

-- SELECT 'Cannot automatically rollback enum value addition' AS notice;