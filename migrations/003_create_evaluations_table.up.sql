CREATE TABLE IF NOT EXISTS evaluations (
    id UUID DEFAULT uuid_generate_v4(),
    flag_key VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    user_key VARCHAR(255),
    variation VARCHAR(255),
    value JSONB,
    reason VARCHAR(50),
    rule_id VARCHAR(255),
    environment VARCHAR(50) NOT NULL DEFAULT 'production',
    attributes JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_evaluations_flag_key ON evaluations(flag_key);
CREATE INDEX idx_evaluations_user_id ON evaluations(user_id);
CREATE INDEX idx_evaluations_timestamp ON evaluations(timestamp);
CREATE INDEX idx_evaluations_environment ON evaluations(environment);

-- Create partitions for recent months
CREATE TABLE evaluations_y2025m01 PARTITION OF evaluations
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE evaluations_y2025m02 PARTITION OF evaluations
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE evaluations_y2025m03 PARTITION OF evaluations
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');