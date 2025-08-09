CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE flag_type AS ENUM ('boolean', 'string', 'number', 'json');

CREATE TABLE IF NOT EXISTS flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type flag_type NOT NULL,
    enabled BOOLEAN DEFAULT false,
    default_value JSONB,
    variations JSONB,
    targeting JSONB,
    environment VARCHAR(50) NOT NULL DEFAULT 'production',
    tags TEXT[],
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(key, environment)
);

CREATE INDEX idx_flags_key ON flags(key);
CREATE INDEX idx_flags_environment ON flags(environment);
CREATE INDEX idx_flags_enabled ON flags(enabled);
CREATE INDEX idx_flags_tags ON flags USING GIN(tags);
CREATE INDEX idx_flags_created_at ON flags(created_at);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_flags_updated_at BEFORE UPDATE ON flags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();