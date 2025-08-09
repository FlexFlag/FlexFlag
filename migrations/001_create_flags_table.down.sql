DROP TRIGGER IF EXISTS update_flags_updated_at ON flags;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS flags;
DROP TYPE IF EXISTS flag_type;